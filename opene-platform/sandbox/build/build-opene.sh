#!/usr/bin/env bash
# Bygger en egen, reproducerbar Open ePlatform-distribution (webbrot) för sandbox och test.
#
#   build-opene.sh <arbetskatalog> <utkatalog>
#
# Kräver: bash, svn, git, JDK 17+ (javac/jar), curl/unzip. Körs i Docker-byggsteget men
# fungerar även i Git Bash på Windows (för felsökning).
#
# Resultat: <utkatalog>/webapp  (en komplett Tomcat-webbrot, utan config.xml med hemligheter)
#           <utkatalog>/sql     (demodatabas + labbjusteringar)
#           <utkatalog>/BUILDINFO
set -euo pipefail

WORK=${1:?arbetskatalog}; OUT=${2:?utkatalog}
HERE=$(cd "$(dirname "$0")" && pwd)
# shellcheck disable=SC1091
source "$HERE/versions.env"

mkdir -p "$WORK" "$OUT"
WORK=$(cd "$WORK" && pwd); OUT=$(cd "$OUT" && pwd)
SRC=$WORK/src; JARS=$WORK/jars; LIBS=$WORK/libs
mkdir -p "$SRC" "$JARS" "$LIBS"

case "$(uname -s)" in MINGW*|MSYS*|CYGWIN*) SEP=';'; native(){ cygpath -w "$1"; } ;; *) SEP=':'; native(){ echo "$1"; } ;; esac
log(){ printf '\n== %s\n' "$*"; }

svn_export(){ # url rev dir
  [ -d "$SRC/$3" ] && return 0
  svn export -q --non-interactive -r "$2" "$1" "$SRC/$3"
}

# ---------------------------------------------------------------- källor
log "Hämtar låsta källor"
svn_export "$OH_SVN/framework/trunk"            "$OH_REV"    OpenHierarchy
svn_export "$OH_SVN/modules/siteprofile/trunk"  "$OH_REV"    SiteProfile
svn_export "$OH_SVN/modules/minimaluser/trunk"  "$OH_REV"    MinimalUser
svn_export "$OH_SVN/interfaces/PaymentProvider/trunk" "$OH_REV" PaymentProvider
svn_export "$OH_SVN/interfaces/PersonProvider/trunk"  "$OH_REV" PersonProvider
for p in standardutils:StandardUtils emailutils:EmailUtils fileuploadutils:FileUploadUtils \
         log4jutils:Log4jUtils webutils:WebUtils cron4jutils:Cron4JUtils; do
  svn_export "$UTILS_SVN/${p%%:*}/trunk" "$UTILS_REV" "${p##*:}"
done
for l in cron4j javamail activation log4j commonsfileupload Commons-IO svnkit ServletAPI; do
  svn_export "$UTILS_SVN/libraries/$l/trunk" "$UTILS_REV" "lib-$l"
done
svn_export "$PURECAPTCHA_SVN" "$PURECAPTCHA_REV" PureCaptcha

if [ ! -d "$SRC/Open-ePlatform/.git" ]; then
  git -c core.longpaths=true clone -q --filter=blob:none "$OPENE_GIT_URL" "$SRC/Open-ePlatform"
fi
git -C "$SRC/Open-ePlatform" -c core.longpaths=true -c advice.detachedHead=false checkout -q "$OPENE_GIT_REF"

# Förbyggda tredjepartsjar från Utils/libraries och OpenHierarchy
find "$SRC"/lib-{javamail,activation,log4j,commonsfileupload,Commons-IO,svnkit} -name '*.jar' -exec cp {} "$LIBS/" \;
cp "$SRC"/OpenHierarchy/WebContent/WEB-INF/lib/*.jar "$LIBS/"
cp "$SRC"/PureCaptcha/lib/*.jar "$LIBS/"
SERVLET=$(find "$SRC/lib-ServletAPI" -name '*.jar' | head -1)

# ---------------------------------------------------------------- kompilering
cp_of(){ # bygger classpath av namngivna jar i $JARS + alla $LIBS + servlet
  local cp="" j
  for j in "$@"; do cp="$cp${SEP}$(native "$JARS/$j.jar")"; done
  for j in "$LIBS"/*.jar; do cp="$cp${SEP}$(native "$j")"; done
  echo "${cp#$SEP}${SEP}$(native "$SERVLET")"
}

build_jar(){ # namn källkatalog [beroenden...]
  local name=$1 dir=$2; shift 2
  [ -f "$JARS/$name.jar" ] && return 0
  log "Kompilerar $name"
  local cls="$WORK/classes/$name"; rm -rf "$cls"; mkdir -p "$cls"
  local cp; cp=$(cp_of "$@")
  (cd "$dir" && find . -name '*.java' | sed 's|^\./||' > "$WORK/classes/$name.files" \
     && javac -nowarn -proc:none --release 8 -encoding ISO-8859-1 -Xlint:-options \
          -cp "$cp" -d "$(native "$cls")" @"$(native "$WORK/classes/$name.files")" 2>&1 \
     ) || { echo "Kompileringen av $name misslyckades" >&2; exit 1; }
  [ "$(find "$cls" -name '*.class' | head -1)" ] || { echo "Kompileringen av $name misslyckades" >&2; exit 1; }
  # resurser (XSL, SQL, properties, ...) ligger blandat i src
  (cd "$dir" && find . -type f ! -name '*.java' -print0 | xargs -0 -I{} cp --parents {} "$cls/")
  (cd "$cls" && jar --create --file "$(native "$JARS/$name.jar")" --no-manifest .)
}

build_jar cron4j          "$SRC/lib-cron4j/src"
build_jar StandardUtils   "$SRC/StandardUtils/src"
build_jar EmailUtils      "$SRC/EmailUtils/src"      StandardUtils
build_jar FileUploadUtils "$SRC/FileUploadUtils/src" StandardUtils
build_jar Log4jUtils      "$SRC/Log4jUtils/src"      StandardUtils
build_jar WebUtils        "$SRC/WebUtils/src"        StandardUtils
build_jar Cron4JUtils     "$SRC/Cron4JUtils/src"     cron4j StandardUtils
build_jar PureCaptcha     "$SRC/PureCaptcha/src"
build_jar OpenHierarchy   "$SRC/OpenHierarchy/src"   StandardUtils EmailUtils FileUploadUtils Log4jUtils WebUtils PureCaptcha
OHDEPS=(OpenHierarchy StandardUtils EmailUtils FileUploadUtils Log4jUtils WebUtils)
build_jar SiteProfile     "$SRC/SiteProfile/src"     "${OHDEPS[@]}"
build_jar MinimalUser     "$SRC/MinimalUser/src"     "${OHDEPS[@]}"
build_jar PaymentProvider "$SRC/PaymentProvider/src" "${OHDEPS[@]}"
build_jar PersonProvider  "$SRC/PersonProvider/src"  "${OHDEPS[@]}"

# ---------------------------------------------------------------- Open ePlatform (Gradle)
log "Bygger Open ePlatform med Gradle"
OPE=$SRC/Open-ePlatform
rm -f "$OPE"/libs/*.jar
for j in OpenHierarchy Cron4JUtils SiteProfile cron4j StandardUtils EmailUtils FileUploadUtils Log4jUtils WebUtils PureCaptcha MinimalUser PaymentProvider PersonProvider; do
  cp "$JARS/$j.jar" "$OPE/libs/"
done
# Projekt som kräver SAML eller JAX-WS hoppas över i sandboxdistributionen.
SKIP=(MinimalUserSAMLAdapter FlowEngineIntegrationCallback)
GRADLE_ARGS=(--no-daemon -q jar)
for s in "${SKIP[@]}"; do GRADLE_ARGS+=(-x ":$s:jar" -x ":$s:compileJava"); done
(cd "$OPE" && ./gradlew "${GRADLE_ARGS[@]}")

# ---------------------------------------------------------------- webbrot
log "Sätter ihop webbroten"
APP=$OUT/webapp; rm -rf "$APP"; mkdir -p "$APP"
cp -r "$SRC/OpenHierarchy/WebContent/." "$APP/"
cp -r "$OPE/openeplatform/designs/demo.oeplatform.org/trunk/WebContent/." "$APP/"
mkdir -p "$APP/WEB-INF/lib"
cp "$LIBS"/*.jar "$APP/WEB-INF/lib/"
cp "$JARS"/*.jar "$APP/WEB-INF/lib/"
find "$OPE/openeplatform" -path '*/build/libs/*.jar' -exec cp {} "$APP/WEB-INF/lib/" \;
# Tredjepartsberoenden som Gradle löste från Maven Central
DEPS_ARGS=(--no-daemon -q -I "$(native "$HERE/copy-runtime-deps.gradle")" copyRuntimeDeps -PdepsDir="$(native "$APP/WEB-INF/lib")")
for s in "${SKIP[@]}"; do DEPS_ARGS+=(-x ":$s:copyRuntimeDeps"); done
(cd "$OPE" && ./gradlew "${DEPS_ARGS[@]}")
rm -f "$APP"/WEB-INF/lib/javax.servlet-api-*.jar "$APP/WEB-INF/config.xml"
# Behåll bara högsta versionen per artefakt (delprojekt kan dra in olika transitiva versioner).
(cd "$APP/WEB-INF/lib" && ls | grep -E -- '-[0-9][0-9A-Za-z._-]*\.jar$' \
   | sed -E 's/^(.*)-([0-9][0-9A-Za-z._-]*)\.jar$/\1 \2/' | sort -k1,1 -k2,2V \
   | awk '{ if ($1==prev) print prevfile; prev=$1; prevfile=$1"-"$2".jar" }' | xargs -r rm -f --)
curl -fsSL -o "$APP/WEB-INF/lib/mariadb-java-client-$MARIADB_CONNECTOR_VERSION.jar" \
  "https://repo1.maven.org/maven2/org/mariadb/jdbc/mariadb-java-client/$MARIADB_CONNECTOR_VERSION/mariadb-java-client-$MARIADB_CONNECTOR_VERSION.jar"

mkdir -p "$OUT/sql"
cp "$OPE/openeplatform/designs/demo.oeplatform.org/trunk/db/Open ePlatform - Example DB.sql" "$OUT/sql/10-demo.sql"

{
  echo "built=$(date -u +%Y-%m-%dT%H:%M:%SZ)"
  echo "opene=$OPENE_GIT_REF"; echo "openhierarchy=r$OH_REV"; echo "utils=r$UTILS_REV"; echo "purecaptcha=r$PURECAPTCHA_REV"
  echo "javac=$(javac -version 2>&1)"
  echo "jars=$(ls "$APP/WEB-INF/lib" | wc -l)"
} > "$OUT/BUILDINFO"
log "Klart: $(cat "$OUT/BUILDINFO" | tr '\n' ' ')"
