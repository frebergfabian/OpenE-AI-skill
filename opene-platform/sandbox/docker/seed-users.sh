#!/usr/bin/env bash
# Sätter lösenord (PBKDF2WithHmacSHA256, samma format som OpenHierarchys PBKDF2PasswordHandler)
# för sandboxens testkonton. Lösenorden kommer från miljövariabler och lagras aldrig i Git.
#   SANDBOX_USERS="admin:ADMIN_PASSWORD sokande:CITIZEN_PASSWORD handlaggare:MANAGER_PASSWORD"
set -uo pipefail
MYSQL=("$@")
log(){ echo "[opene-seed] $*"; }

# Vänta tills SimpleUserProvider har migrerat (kolumnen hashAlgoritm finns)
for i in $(seq 1 120); do
  c=$("${MYSQL[@]}" -Nse "SELECT COUNT(*) FROM information_schema.COLUMNS WHERE TABLE_SCHEMA=DATABASE() AND TABLE_NAME='simple_users' AND COLUMN_NAME='hashAlgoritm'" 2>/dev/null || echo 0)
  [ "$c" = "1" ] && break; sleep 5
done
[ "${c:-0}" = "1" ] || { log "simple_users saknar hashAlgoritm efter väntan"; exit 1; }

pbkdf2(){ # lösenord -> iterationer:salt:hash
  local salt hash
  salt=$(openssl rand -hex 128)
  hash=$(openssl kdf -keylen 64 -kdfopt digest:SHA256 -kdfopt "pass:$1" -kdfopt "hexsalt:$salt" \
                     -kdfopt iter:310000 PBKDF2 | tr -d ':' | tr 'A-F' 'a-f')
  echo "310000:$salt:$hash"
}

sql_str(){ printf "%s" "$1" | sed "s/'/''/g"; }

# username:VARIABEL_MED_LÖSENORD:förnamn:efternamn:e-post:grupp1,grupp2:admin(0/1)
# Gruppnamn som innehåller mellanslag skrivs med _ (ersätts nedan).
: "${SANDBOX_ACCOUNTS:=sandbox-admin:ADMIN_PASSWORD:Sandbox:Admin:sandbox-admin@sandbox.invalid:Systemadministratörer,E-tjänst_administratörer,Handläggare:1 sokande:CITIZEN_PASSWORD:Alex:Testsökande:sokande@sandbox.invalid:Medborgare:0 sokande2:CITIZEN_PASSWORD:Charlie:Testsökande:sokande2@sandbox.invalid:Medborgare:0 handlaggare:MANAGER_PASSWORD:Sam:Testhandläggare:handlaggare@sandbox.invalid:Handläggare:0}"

for spec in $SANDBOX_ACCOUNTS; do
  IFS=: read -r user var first last mail groups isadmin <<<"$spec"
  pw=${!var:-}
  [ -n "$pw" ] || { log "hoppar över $user ($var är inte satt)"; continue; }
  h=$(pbkdf2 "$pw")
  "${MYSQL[@]}" --default-character-set=utf8mb4 -e "
    INSERT INTO simple_users (username, password, firstname, lastname, email, admin, enabled, added, hashAlgoritm)
      VALUES ('$(sql_str "$user")', '$h', '$(sql_str "$first")', '$(sql_str "$last")', '$(sql_str "$mail")', $isadmin, 1, NOW(), 'PBKDF2WithHmacSHA256')
    ON DUPLICATE KEY UPDATE password=VALUES(password), hashAlgoritm=VALUES(hashAlgoritm), enabled=1, admin=VALUES(admin);" \
  && log "konto $user klart" || log "kunde inte skapa $user"
  for g in ${groups//,/ }; do
    g=${g//_/ }
    "${MYSQL[@]}" --default-character-set=utf8mb4 -e "INSERT IGNORE INTO simple_user_groups (userID, groupID)
       SELECT u.userID, g.groupID FROM simple_users u, simple_groups g WHERE u.username='$(sql_str "$user")' AND g.name='$(sql_str "$g")';"
  done
done
"${MYSQL[@]}" -e "REPLACE INTO sandbox_meta VALUES ('users_seeded', NOW());"
