param(
    [string]$Version = 'v0.5.0-dev',
    [string]$Name = 'default',
    [switch]$Ephemeral,
    [switch]$NoBrowser
)
$ErrorActionPreference = 'Stop'
$repo = 'frebergfabian/OpenE-AI-skill'
if ($Version -ne 'latest' -and $Version -notmatch '^v?[0-9]+\.[0-9]+\.[0-9]+(?:-[A-Za-z0-9.-]+)?$') { throw 'Ogiltig releaseversion' }
if ($Name -notmatch '^[A-Za-z0-9][A-Za-z0-9_-]{0,47}$') { throw 'Ogiltigt instansnamn' }
if ($Ephemeral -and $Name -eq 'default') { throw 'Ange -Name med -Ephemeral för att kunna kontrollera instansen' }
if ($Version -eq 'latest') {
    if (Get-Command gh -ErrorAction SilentlyContinue) {
        $Version = (& gh release view --repo $repo --json tagName --jq .tagName)
        if ($LASTEXITCODE -ne 0) { throw 'Kunde inte läsa senaste release via gh' }
    } else {
        $Version = (Invoke-RestMethod -Uri "https://api.github.com/repos/$repo/releases/latest" -Headers @{ 'User-Agent' = 'OpenE-Sandbox' }).tag_name
    }
}
if ($Version -notmatch '^v?[0-9]+\.[0-9]+\.[0-9]+(?:-[A-Za-z0-9.-]+)?$') { throw 'Ogiltig releasetagg' }
$asset = "OpenE-Sandbox-$Version-win-x64.exe"
$bin = Join-Path $env:LOCALAPPDATA 'OpenE-Sandbox\bin'
New-Item -ItemType Directory -Path $bin -Force | Out-Null
$tmp = Join-Path $bin ('.download-' + [Guid]::NewGuid().ToString('N'))
New-Item -ItemType Directory -Path $tmp | Out-Null
try {
    if (Get-Command gh -ErrorAction SilentlyContinue) {
        & gh release download $Version --repo $repo --pattern $asset --pattern "$asset.sha256" --dir $tmp
        if ($LASTEXITCODE -ne 0) { throw 'Kunde inte hämta release via gh' }
    } else {
        $base = "https://github.com/$repo/releases/download/$Version"
        Invoke-WebRequest -Uri "$base/$asset" -OutFile (Join-Path $tmp $asset)
        Invoke-WebRequest -Uri "$base/$asset.sha256" -OutFile (Join-Path $tmp "$asset.sha256")
    }
    $checksum = (Get-Content -LiteralPath (Join-Path $tmp "$asset.sha256") -Raw).Trim()
    if ($checksum -notmatch '^([a-fA-F0-9]{64})(?:\s+\*?([^\r\n]+))?$') { throw 'Ogiltigt checksummeformat' }
    $expected = $Matches[1].ToLowerInvariant()
    if ($Matches[2] -and $Matches[2] -ne $asset) { throw 'Checksumman avser en annan fil' }
    $download = Join-Path $tmp $asset
    if ((Get-FileHash -LiteralPath $download -Algorithm SHA256).Hash.ToLowerInvariant() -ne $expected) { throw 'SHA-256 stämmer inte' }
    $exe = Join-Path $bin $asset
    Move-Item -LiteralPath $download -Destination $exe -Force
    $arguments = @('start', '--name', $Name)
    if ($Ephemeral) { $arguments += '--ephemeral' }
    if ($NoBrowser) { $arguments += '--no-browser' }
    & $exe @arguments
    if ($LASTEXITCODE -ne 0) { throw 'Sandboxen startade inte' }
    & $exe status --name $Name --json
    if ($LASTEXITCODE -ne 0) { throw 'Sandboxens hälsokontroll misslyckades' }
} finally {
    Remove-Item -LiteralPath $tmp -Recurse -Force -ErrorAction SilentlyContinue
}
