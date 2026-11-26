# OpenE sandbox

**Windows x64:** använd den [portabla förhandsversionen](../references/portable-sandbox.md), som inkluderar Java och MariaDB och inte behöver Docker. `get-sandbox.ps1` hämtar den låsta testreleasen `v0.5.0-dev`. Instruktionerna nedan gäller den separata Docker-varianten.

En lokal, isolerad Open ePlatform-miljö som byggs från **låsta källor** och styrs med `sandboxctl`. Den är fristående från Claude Code. Skillen, Codex och andra agenter är bara klienter som anropar CLI:t.

```bash
./sandboxctl create          # skapar instansen "default" (persistent), bygger imagen första gången
./sandboxctl status          # verklig hälsa: HTTP, inloggningssida, seedade konton
./sandboxctl test            # röktest (inloggning, adminsidor) + Playwright om Node finns
./sandboxctl info --secrets  # URL:er och testlösenord
./sandboxctl stop            # stoppar, data behålls
./sandboxctl start
./sandboxctl reset --yes     # raderar data, ny ren bas
./sandboxctl destroy --yes   # tar bort allt för instansen
./sandboxctl create --ephemeral   # engångsinstans med eget namn, egna portar och volymer
./sandboxctl remote https://host:8443   # läs-endast kontroll av befintlig miljö
```

## Beroenden på värddatorn
| Krav | Varför |
|---|---|
| Docker Engine 24+ med Compose v2 (Windows: Docker Desktop + **WSL2**) | Kör db, OpenE, Mailpit och proxy |
| bash (Linux/macOS, Windows: Git Bash) + curl | `sandboxctl` och röktest |
| Utgående nät vid **bygge**: github.com (inklusive releasefiler), Maven Central, Docker Hub; HTTPS | Källor och beroenden |
| Valfritt: Node 22+ med npm och operativsystemets Chromium-bibliotek | Playwright-tester (`tests/playwright`); `sandboxctl test` kör npm ci och hämtar Chromium. På en minimal Linuxvärd: kör `npx playwright install --with-deps chromium` i testkatalogen med lämpliga installationsrättigheter. |

Ingen Java, Maven eller SVN behövs på värden: allt byggs i Dockerfilens byggsteg.

## Vad som körs
```
127.0.0.1:<http>  ─┐
127.0.0.1:<mail>  ─┤ proxy (nginx) ── nät "edge"
                   └─────────────── nät "backend" (internal: ingen internetåtkomst)
                                       ├─ opene  (Tomcat 9.0.123, JRE 21, egen OpenE-distribution)
                                       ├─ db     (MariaDB 11.8, latin1, egen volym)
                                       └─ mail   (Mailpit: all e-post fångas, inget skickas vidare)
```
- Portar publiceras bara på 127.0.0.1. Varje instans får egna portar (8480, 8490 …), eget Compose-projekt `opene-<namn>`, egna nät och volymer.
- `opene`, `db` och `mail` kan inte nå internet eller produktionssystem. `opene` kör som icke-root, alla containrar har `no-new-privileges` och `cap_drop: ALL`.
- Hemligheter slumpas per instans till `~/.opene-sandbox/instances/<namn>/.env` (rättigheter 600) och finns aldrig i Git eller i imagen. `config.xml` skapas vid start.

## Egen OpenE-distribution (`build/`)
Dockerbygget hämtar den publicerade SVN-källspegeln från v0.5.0-dev och verifierar SHA-256 före användning. Den innehåller samma låsta revisioner, inga ersättningsberoenden. Ett fristående källbygge utan `OPENE_SVN_MIRROR` behöver fortfarande SVN port 3690.

`build/build-opene.sh` bygger en komplett webbrot från versionerna i `build/versions.env`:
1. Hämtar Open ePlatform (Git, låst commit), OpenHierarchy (svn r3190), Nordic Peak Utils (svn r1649) och PureCaptcha (svn r15).
2. Kompilerar Utils, cron4j-forken, PureCaptcha, OpenHierarchy, SiteProfile, MinimalUser, PaymentProvider och PersonProvider med `javac --release 8 -encoding ISO-8859-1`.
3. Bygger Open ePlatform med repots Gradle-wrapper. SAML-adaptern och integrationscallbacken hoppas över.
4. Sätter ihop OpenHierarchys WebContent, demodesignen, alla JAR:er (dubblettversioner rensas) och MariaDB Connector/J.
5. Skriver `BUILDINFO` med alla versioner. Imagen märks `opene-sandbox:local`.

Skriptet fungerar även utanför Docker (Linux, eller Git Bash med JDK 17+ och svn) för felsökning.

## Data och konton
- Första start: upstreams demodatabas (`designs/demo.oeplatform.org/.../Example DB.sql`) och `docker/sql/2*.sql` (SMTP → Mailpit, demokontot `admin` inaktiveras). Därefter migrerar modulerna databasen själva.
- `seed-users.sh` skapar `sandbox-admin`, `handlaggare`, `sokande` och `sokande2` med PBKDF2-lösenord från `.env`, när migreringarna är klara.
- Persistent data: volymerna `db-data` och `opene-files`. `stop`/`start` behåller dem, `reset`/`destroy` tar bort dem (kräver `--yes` för persistenta instanser).

## Uppdatera versioner
Ändra `build/versions.env`, kör `./sandboxctl build`, skapa en ephemeral instans, kör `./sandboxctl test <namn>` och uppdatera status i [genomförbarhetsanalysen](../references/sandbox.md).

## Felsökning
| Symptom | Åtgärd |
|---|---|
| `error during connect … dockerDesktopLinuxEngine` | Docker Desktop körs inte eller WSL2 saknas (`wsl --install` som admin, starta om) |
| Bygget hänger vid "Hämtar låsta källor" | Port 3690 (svn://) blockeras. Testa `svn info svn://svn.openhierarchy.org/openhierarchy` |
| `status` visar http 000 länge | `./sandboxctl logs <namn> opene`. Första start migrerar databasen (flera minuter) |
| Inloggning misslyckas | `status`: "konton" måste vara seedat. Se `logs … opene` efter `[opene-seed]` |
| Å/ä/ö fel i skript | Sidorna är ISO-8859-1. Läs med `TextDecoder('windows-1252')` (se skillens arbetsregel 7) |
