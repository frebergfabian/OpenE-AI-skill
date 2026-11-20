# Builds the current experimental Windows prototype. This is not a release gate:
# run the acceptance suite and review module activation and licenses separately.
param(
    [Parameter(Mandatory)][string]$BuildRoot,
    [Parameter(Mandatory)][string]$SourceMirror,
    [Parameter(Mandatory)][string]$AddonArchive,
    [Parameter(Mandatory)][string]$MariaDBZip,
    [Parameter(Mandatory)][string]$TomcatZip,
    [Parameter(Mandatory)][string]$JdkHome,
    [Parameter(Mandatory)][string]$GoExe,
    [switch]$AllowDirty,
    [string]$Version = '0.2.2-dev',
    [string]$SevenZip = 'C:\Program Files\7-Zip\7z.exe',
    [string]$Bash = 'C:\Program Files\Git\bin\bash.exe'
)
$ErrorActionPreference = 'Stop'
if ($Version -notmatch '^[0-9]+\.[0-9]+\.[0-9]+(?:-[A-Za-z0-9.-]+)?$') { throw 'Invalid version' }
function Check-Hash([string]$Path,[string]$Algorithm,[string]$Expected) {
    if ((Get-FileHash -LiteralPath $Path -Algorithm $Algorithm).Hash.ToLowerInvariant() -ne $Expected) { throw "Checksum mismatch: $Path" }
}
function Check-Exit([string]$Step) { if ($LASTEXITCODE -ne 0) { throw "$Step failed with exit code $LASTEXITCODE" } }
Check-Hash $SourceMirror SHA256 '744d32447e3d39b91db333110074c004bcde6981401ea42c233caeed866c63ad'
Check-Hash $AddonArchive SHA256 '46a0739855341300dc5478156e92beb8a991461929e39cc77b35313be2fbe157'
Check-Hash $MariaDBZip SHA256 '05d04b46763ea3dc9111bfc82dc061e0a158c682a73d31ba34b13f2653c88d17'
Check-Hash $TomcatZip SHA512 'a060f226f1cb61732d286c478982ac4362355e0f5672bb5d3743ae4da3c7b1e6590c81555a003fddd5b6bc190d6049e7c85d16cacfba56181bfd4610380d4fac'
if ((& (Join-Path $JdkHome 'bin\javac.exe') -version 2>&1 | Out-String) -notmatch 'javac 21\.0\.13') { throw 'Expected Temurin JDK 21.0.13' }
if ((& $GoExe version | Out-String) -notmatch '^go version go1\.27\.4 windows/amd64') { throw 'Expected Go 1.27.4 windows/amd64' }
$GoExe = (Resolve-Path -LiteralPath $GoExe).Path
if (Test-Path -LiteralPath $BuildRoot) { throw 'BuildRoot must not already exist' }
New-Item -ItemType Directory -Path $BuildRoot | Out-Null
$root = (Resolve-Path -LiteralPath $BuildRoot).Path
$here = $PSScriptRoot
$sandbox = Split-Path $here -Parent
$repo = Split-Path (Split-Path $sandbox -Parent) -Parent
$sourceRevision = (& git -C $repo rev-parse HEAD).Trim()
Check-Exit 'Reading sandbox source revision'
$dirty = @(& git -C $repo status --porcelain --untracked-files=normal)
Check-Exit 'Checking sandbox source changes'
if ($dirty.Count -gt 0 -and -not $AllowDirty) { throw 'Sandbox checkout has uncommitted changes; commit them before building or explicitly pass -AllowDirty for development' }
$env:JAVA_HOME = (Resolve-Path -LiteralPath $JdkHome).Path
$env:PATH = (Join-Path $env:JAVA_HOME 'bin') + ';' + $env:PATH
$env:OPENE_SVN_MIRROR = (Resolve-Path -LiteralPath $SourceMirror).Path
$env:OPENE_SVN_MIRROR_SHA256 = '744d32447e3d39b91db333110074c004bcde6981401ea42c233caeed866c63ad'
$posixRoot = (& $Bash -c 'cygpath -u "$1"' -- $root).Trim()
Check-Exit 'Resolving build path'
& $Bash (Join-Path $here 'build-opene.sh') "$posixRoot/source-build" "$posixRoot/dist"
Check-Exit 'OpenE build'
$addonDir = Join-Path $root 'addons'
New-Item -ItemType Directory -Path $addonDir | Out-Null
& tar -xzf (Resolve-Path -LiteralPath $AddonArchive).Path -C $addonDir
Check-Exit 'Extracting add-ons'
$runtime = Join-Path $root 'runtime'
New-Item -ItemType Directory -Path $runtime | Out-Null
& $SevenZip x (Resolve-Path -LiteralPath $MariaDBZip).Path "-o$runtime" -y | Out-Null
Check-Exit 'Extracting MariaDB'
& $SevenZip x (Resolve-Path -LiteralPath $TomcatZip).Path "-o$runtime" -y | Out-Null
Check-Exit 'Extracting Tomcat'
$webapp = Join-Path $root 'dist\webapp'
$servlet = Join-Path $runtime 'apache-tomcat-9.0.123\lib\servlet-api.jar'
$jaxws = Join-Path $webapp 'WEB-INF\lib\jaxws-api-2.3.1.jar'
$jws = Join-Path $webapp 'WEB-INF\lib\javax.jws-api-1.1.jar'
Invoke-WebRequest 'https://repo.maven.apache.org/maven2/javax/xml/ws/jaxws-api/2.3.1/jaxws-api-2.3.1.jar' -OutFile $jaxws
Check-Hash $jaxws SHA256 'a447f84f95658ea68b347acffe156f7700c62a37ede15d81e5298fb8e5fe6dcf'
Invoke-WebRequest 'https://repo.maven.apache.org/maven2/javax/jws/javax.jws-api/1.1/javax.jws-api-1.1.jar' -OutFile $jws
Check-Hash $jws SHA256 '9f20ab1fea3f9571ed52a9d98e3c651cc7c04c8a709addf238312b60987c6f2c'
& node (Join-Path $here 'patch-ssbt.mjs') $addonDir
Check-Exit 'SSBT sandbox adapter patch'
& node (Join-Path $here 'build-all-addons.mjs') $addonDir $webapp $env:JAVA_HOME $servlet (Join-Path $root 'addon-build')
Check-Exit 'Add-on build'
$buildInfo = Join-Path $root 'dist\BUILDINFO'
@(
    'source_mirror_sha256=744d32447e3d39b91db333110074c004bcde6981401ea42c233caeed866c63ad'
    'nordic_peak_addons_sha256=46a0739855341300dc5478156e92beb8a991461929e39cc77b35313be2fbe157'
    'mariadb=11.8.10;sha256=05d04b46763ea3dc9111bfc82dc061e0a158c682a73d31ba34b13f2653c88d17'
    'tomcat=9.0.123;sha512=a060f226f1cb61732d286c478982ac4362355e0f5672bb5d3743ae4da3c7b1e6590c81555a003fddd5b6bc190d6049e7c85d16cacfba56181bfd4610380d4fac'
    'mariadb_connector=2.7.6;sha256=d3e448b1e8799ccc06c7ad2745aa03e72f19e13124b676b9656a54a833af7a93'
    "launcher=$Version;go=1.27.4;target=windows-amd64"
    "sandbox_source_commit=$sourceRevision;dirty=$($dirty.Count -gt 0)"
) | Add-Content -LiteralPath $buildInfo -Encoding utf8
$javaClasses = Join-Path $root 'sandbox-classes'
New-Item -ItemType Directory -Path $javaClasses | Out-Null
$cp = (Join-Path $webapp 'WEB-INF\lib\*') + ';' + $servlet
& (Join-Path $env:JAVA_HOME 'bin\javac.exe') --release 8 -encoding UTF-8 -cp $cp -d $javaClasses (Join-Path $sandbox 'java\src\se\exempelkoping\sandbox\SimulatedSMSSender.java') (Join-Path $sandbox 'java\src\se\exempelkoping\sandbox\SimulatedSwedbankPayProvider.java')
Check-Exit 'Sandbox Java build'
& (Join-Path $env:JAVA_HOME 'bin\jar.exe') --create --file (Join-Path $webapp 'WEB-INF\lib\SandboxIntegrations.jar') -C $javaClasses .
Check-Exit 'Sandbox Java packaging'
$payload = Join-Path $root 'payload'
New-Item -ItemType Directory -Path $payload | Out-Null
& (Join-Path $env:JAVA_HOME 'bin\jlink.exe') --add-modules ALL-MODULE-PATH --strip-debug --no-header-files --no-man-pages --output (Join-Path $payload 'jre')
Check-Exit 'jlink'
# Use a separate source staging directory so go:embed never modifies the checkout.
$portable = Join-Path $root 'portable'
New-Item -ItemType Directory -Path $portable | Out-Null
Get-ChildItem -LiteralPath (Join-Path $sandbox 'portable') -File | Where-Object { $_.Name -ne 'payload.zip' } | Copy-Item -Destination $portable
$env:OPENE_EXPERIMENTAL_ADDONS = '1'
$env:OPENE_7ZIP = (Resolve-Path -LiteralPath $SevenZip).Path
& node (Join-Path $here 'prepare-payload.mjs') (Join-Path $root 'dist') $runtime $payload $portable $sandbox
Check-Exit 'Payload packaging'
$asset = "OpenE-Sandbox-$Version-win-x64.exe"
Push-Location $portable
try {
    & $GoExe test ./...
    Check-Exit 'Go tests'
    & $GoExe build -trimpath -ldflags "-s -w -X main.version=$Version" -o (Join-Path $root $asset) .
    Check-Exit 'Executable build'
} finally { Pop-Location }
if ((& git -C $repo rev-parse HEAD).Trim() -ne $sourceRevision) { throw 'Sandbox source revision changed during build' }
if (-not $AllowDirty -and @(& git -C $repo status --porcelain --untracked-files=normal).Count -gt 0) { throw 'Sandbox source changed during build' }
$hash = (Get-FileHash -LiteralPath (Join-Path $root $asset) -Algorithm SHA256).Hash.ToLowerInvariant()
[IO.File]::WriteAllText((Join-Path $root "$asset.sha256"), "$hash  $asset`n", [Text.Encoding]::ASCII)
Copy-Item -LiteralPath (Join-Path $root 'dist\BUILDINFO') -Destination (Join-Path $root 'BUILDINFO')
Write-Host "Experimental executable: $(Join-Path $root $asset)"
Write-Host 'Not a release: acceptance and integration tests remain required.'
