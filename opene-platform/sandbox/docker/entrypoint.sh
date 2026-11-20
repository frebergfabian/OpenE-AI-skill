#!/usr/bin/env bash
# Startar OpenE i sandboxen: konfiguration från miljövariabler, databasinitiering vid
# första start, Tomcat, och därefter seedning av testkonton när migreringarna har körts.
set -euo pipefail

: "${DB_HOST:=db}" "${DB_NAME:=opene}" "${DB_USER:=opene}"
: "${DB_PASSWORD:?DB_PASSWORD saknas}"
APP=/usr/local/tomcat/webapps/ROOT
MYSQL=(mariadb --skip-ssl -h "$DB_HOST" -u "$DB_USER" -p"$DB_PASSWORD" "$DB_NAME")
log(){ echo "[opene-entrypoint] $*"; }

# 1. config.xml från mall (aldrig incheckad med lösenord)
esc(){ printf '%s' "$1" | sed -e 's/[&/\]/\\&/g'; }
sed -e "s/@DB_HOST@/$(esc "$DB_HOST")/" -e "s/@DB_NAME@/$(esc "$DB_NAME")/" \
    -e "s/@DB_USER@/$(esc "$DB_USER")/" -e "s/@DB_PASSWORD@/$(esc "$DB_PASSWORD")/" \
    /opt/opene/config.xml.tmpl > "$APP/WEB-INF/config.xml"
chmod 600 "$APP/WEB-INF/config.xml"

# 2. vänta på databasen
for i in $(seq 1 60); do "${MYSQL[@]}" -e 'SELECT 1' >/dev/null 2>&1 && break; sleep 2; done
"${MYSQL[@]}" -e 'SELECT 1' >/dev/null

# 3. första start: ladda demodatabas + sandboxjusteringar (före migreringar)
if [ -z "$("${MYSQL[@]}" -Nse "SHOW TABLES LIKE 'sandbox_meta'")" ]; then
  log "Tom databas: laddar demodata"
  for f in /opt/opene/sql/1*.sql /opt/opene/sql/2*.sql; do
    [ -f "$f" ] || continue
    log "  $f"; "${MYSQL[@]}" --default-character-set=latin1 < "$f"
  done
  "${MYSQL[@]}" -e "CREATE TABLE sandbox_meta (k VARCHAR(64) PRIMARY KEY, v VARCHAR(255));
                    INSERT INTO sandbox_meta VALUES ('initialized', NOW()), ('build', '$(tr '\n' ' ' < /opt/opene/BUILDINFO)');"
fi

# 4. seedning av testkonton efter att Tomcat startat och migrerat
( /opt/opene/seed-users.sh "${MYSQL[@]}" || log "VARNING: seedning av testkonton misslyckades" ) &

log "Startar Tomcat ($(tr '\n' ' ' < /opt/opene/BUILDINFO))"
exec "$@"
