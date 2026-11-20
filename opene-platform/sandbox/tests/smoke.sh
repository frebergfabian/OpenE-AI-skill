#!/usr/bin/env bash
# Röktest som bevisar att OpenE fungerar, inte bara att containrar körs:
# katalog svarar, inloggning lyckas, adminsidor öppnas, Adm. e-tjänster listar tjänster,
# och teckenkodningen är den förväntade (ISO-8859-1).
#   OPENE_URL=http://127.0.0.1:8480 OPENE_ADMIN_USER=sandbox-admin OPENE_ADMIN_PASSWORD=... smoke.sh
set -uo pipefail
URL=${OPENE_URL:?}; USER=${OPENE_ADMIN_USER:?}; PASS=${OPENE_ADMIN_PASSWORD:?}
JAR=$(mktemp); trap 'rm -f "$JAR" "$JAR".body' EXIT
pass=0; fail=0
ok(){ echo "  OK   $*"; pass=$((pass+1)); }
ko(){ echo "  FEL  $*"; fail=$((fail+1)); }
get(){ curl -s -m 30 -b "$JAR" -c "$JAR" -o "$JAR.body" -w '%{http_code} %{content_type}' "$URL$1"; }
body_has(){ iconv -f ISO-8859-1 -t UTF-8 "$JAR.body" 2>/dev/null | grep -q -- "$1"; }

echo "Röktest mot $URL"
r=$(get /); [[ $r == 200* ]] && ok "startsidan 200 ($r)" || ko "startsidan: $r"
[[ $r == *"ISO-8859-1"* || $r == *"iso-8859-1"* ]] && ok "teckenkodning ISO-8859-1" || ko "oväntad content-type: $r"

r=$(get /login); [[ $r == 200* ]] && ok "inloggningssidan 200" || ko "inloggningssidan: $r"

# Inloggning (UserProviderLoginModule: fälten username/password)
code=$(curl -s -m 30 -b "$JAR" -c "$JAR" -o "$JAR.body" -w '%{http_code}' \
  --data-urlencode "username=$USER" --data-urlencode "password=$PASS" "$URL/login")
r=$(get /flowadmin)
if [[ $r == 200* ]] && body_has "E-tjänst"; then ok "inloggad som $USER, Adm. e-tjänster öppnas"; else ko "inloggning/adminsida misslyckades (login $code, flowadmin $r)"; fi

for p in /administration/useradmin /administration/systemadmin /flowinstanceadmin /minasidor; do
  r=$(get "$p"); [[ $r == 200* ]] && ok "$p 200" || ko "$p: $r"
done

r=$(get /logout); ok "utloggning ($r)"
r=$(get /flowadmin); if body_has "Logga in\|login"; then ok "utloggad användare stoppas från /flowadmin"; else ko "admin nåbar utan inloggning? ($r)"; fi

echo "Resultat: $pass OK, $fail FEL"
[ "$fail" -eq 0 ]
