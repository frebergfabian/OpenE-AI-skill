#!/usr/bin/env bash
# Röktest som bevisar att OpenE fungerar, inte bara att containrar körs:
# katalog svarar, inloggning lyckas, adminsidor öppnas med rätt innehåll, behörighet
# stoppar utloggade, och teckenkodningen är den förväntade (ISO-8859-1).
#   OPENE_URL=http://127.0.0.1:8480 OPENE_ADMIN_USER=sandbox-admin OPENE_ADMIN_PASSWORD=... smoke.sh
set -uo pipefail
URL=${OPENE_URL:?}; USER=${OPENE_ADMIN_USER:?}; PASS=${OPENE_ADMIN_PASSWORD:?}
JAR=$(mktemp); BODY=$(mktemp); trap 'rm -f "$JAR" "$BODY"' EXIT
pass=0; fail=0
ok(){ echo "  OK   $*"; pass=$((pass+1)); }
ko(){ echo "  FEL  $*"; fail=$((fail+1)); }
# get <sökväg> -> "kod|content-type|redirect"
get(){ curl -s -m 30 -b "$JAR" -c "$JAR" -o "$BODY" -w '%{http_code}|%{content_type}|%{redirect_url}' "$URL$1"; }
# OpenE skickar ISO-8859-1 och HTML-entiteter: normalisera innan sökning.
text(){ iconv -f ISO-8859-1 -t UTF-8 "$BODY" 2>/dev/null | sed 's/&auml;/ä/g; s/&aring;/å/g; s/&ouml;/ö/g; s/&Auml;/Ä/g; s/&Aring;/Å/g; s/&Ouml;/Ö/g'; }
has(){ text | grep -q -E -- "$1"; }

echo "Röktest mot $URL"
r=$(get /); [[ $r == 200* ]] && ok "katalogen 200" || ko "katalogen: $r"
[[ ${r,,} == *"iso-8859-1"* ]] && ok "teckenkodning ISO-8859-1" || ko "oväntad content-type: $r"
r=$(get /login); [[ $r == 200* ]] && has 'name="password"' && ok "inloggningsformulär" || ko "inloggningssidan: $r"

r=$(get /flowadmin); [[ $r == 302*login* ]] && ok "utloggad skickas till inloggning från /flowadmin" || ko "/flowadmin utan inloggning: $r"

curl -s -m 30 -b "$JAR" -c "$JAR" -o "$BODY" --data-urlencode "username=$USER" --data-urlencode "password=$PASS" "$URL/login"
r=$(get /flowadmin)
if [[ $r == 200* ]] && has '/flowadmin/addflow' && has '/flowadmin/importflow' && ! has 'name="password"'; then ok "inloggad som $USER: Adm. e-tjänster visas"; else ko "inloggning eller Adm. e-tjänster misslyckades ($r)"; fi

for p in /flowadmin/standardstatuses /administration/useradmin /administration/systemadmin /flowinstanceadmin /minasidor; do
  r=$(get "$p")
  if [[ $r == 200* ]] && ! has 'Ett oväntat fel|name="password"'; then ok "$p"; else ko "$p: $r"; fi
done
get /flowadmin/standardstatuses >/dev/null; has 'Enkel' && ok "statusgrupper migrerade (Enkel finns)" || ko "statusgruppen Enkel saknas"

get /logout >/dev/null
r=$(get /administration/useradmin); [[ $r == 302*login* || $r == 403* ]] && ok "utloggning tar bort åtkomst" || ko "adminsida nåbar efter utloggning: $r"

echo "Resultat: $pass OK, $fail FEL"
[ "$fail" -eq 0 ]
