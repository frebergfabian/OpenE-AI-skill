# Lokal OpenE-sandbox: genomförbarhet, arkitektur och status

Undersökning och PoC påbörjad 2026-11-20 av Claude. Implementationen finns i [`../sandbox/`](../sandbox/README.md).
Märkning: **[Byggt]** kod finns, **[Testat]** körd och verifierad, **[Teori]** ej körd.

## 1. Inventering av byggkedjan (verifierad)

| Komponent | Källa | Låst version | Byggs med |
|---|---|---|---|
| Open ePlatform | github.com/Open-ePlatform/Open-ePlatform | `a34ab2868be5…` (2026-10-13, 2.0.0-SNAPSHOT) | Gradle-wrapper 9.7.0, JDK 17+ (`--release 8`) |
| OpenHierarchy (ramverk) | `svn://svn.openhierarchy.org/openhierarchy/framework/trunk` | r3190 (HEAD är r3191) | egen javac (Eclipse-projekt utan byggfil) |
| SiteProfile, MinimalUser | `…/modules/{siteprofile,minimaluser}/trunk` | r3190 | javac |
| PaymentProvider, PersonProvider | `…/interfaces/*/trunk` | r3190 | javac |
| StandardUtils, EmailUtils, FileUploadUtils, Log4jUtils, WebUtils, Cron4JUtils | `svn://svn.nordicpeak.com/utils/*/trunk` | r1649 (HEAD r1650) | javac |
| cron4j (fork med `Scheduler(String)`) | `…/utils/libraries/cron4j/trunk` (källkod) | r1649 | javac |
| javamail 1.6.2, activation, reload4j 1.2.19, commons-fileupload 1.3.3, commons-io 2.6, svnkit 1.8.6 | `…/utils/libraries/*` (färdiga jar) | r1649 | – |
| PureCaptcha | `svn://svn.nordicpeak.com/purecaptcha/trunk` | r15 | javac |
| Övriga tredjepartsbibliotek | Maven Central via Gradle | enligt upstream `build.gradle` | Gradle |
| Demodatabas | upstream `designs/demo.oeplatform.org/trunk/db/Open ePlatform - Example DB.sql` | samma commit | – |
| Runtime | Tomcat 9.0.123 (javax.servlet), JRE 21, MariaDB 11.8, Connector/J 2.7.6 | – | Docker-images |

**Kontroll mot dokumentationen.** Commit, Gradle 9.7.0, JDK 21, Tomcat 9.0.123, MariaDB 11.8 och Connector/J 2.7.6 stämmer med `installation.md`. "OpenHierarchy 1.4.5" kunde inte hittas i SVN (senaste tagg är 1.2.8). Versionen bör anges som **SVN-revision r3190**.

**[Testat]** `build/build-opene.sh` byggde hela kedjan i Git Bash med Temurin 21.0.13 och SlikSVN 1.14.2. Resultatet blev en webbrot med 188 JAR:er (113 MB) efter dubblettrensning. Byggtid är några minuter med varm Gradle-cache.

## 2. Hinder för reproducerbarhet

| Hinder | Risk | Hantering |
|---|---|---|
| OpenHierarchy och Utils finns bara i Subversion över `svn://` (port 3690, ingen HTTP) | Brandväggar blockerar ofta 3690. Servrarna kan försvinna | Låsta revisioner + **byggd image som artefakt** (bygg en gång, återanvänd). Rekommendation: spegla källorna (svn export → eget Git-repo eller tarball med checksumma) |
| Eclipse-projekt utan byggfiler | Beroenden måste härledas från `.classpath` | Kodat i `build-opene.sh` (ordning och beroenden) |
| Delprojekt drar in olika transitiva versioner (BouncyCastle 1.64/1.67, HttpClient 4.5.12/13, juniversalchardet 1.0.3/2.3.2) | Klassladdningskonflikter | Dubblettrensning: högsta version per artefakt behålls |
| SAML-adaptern och integrationscallbacken kräver OpenSAML respektive JAX-WS (WSModule) | Stora beroenden | Hoppas över i sandboxdistributionen (behövs inte för formulär och handläggning) |
| Demodatabasen är gammal (FlowEngine-schema v52, migreras till ca v169 vid start) | Codex mötte migreringsfel: saknade åtgärds-ID före migrering 137, radstorlek i StandardFlowNotificationHandler v42, SiteProfile 6→8 | Hanteras i `docker/sql/2*.sql` när de reproducerats (se status) |
| Demokontot `admin` har en känd SHA1-hash | Säkerhet | Inaktiveras. Egna konton med PBKDF2 från `.env` |
| Nordic Peaks tilläggsmoduler (bokning, kö, signering …) ingår inte i Open-ePlatform-repot | Funktionsglapp mot den gemensamma sandboxen | Steg 2: bygg valda moduler från Nordic Peaks AGPL-arkiv på samma sätt |

## 3. Alternativ och val

| | A. Bygga från källkod vid varje installation | **B. Egen versionerad image byggd från källkod** | C. Befintlig officiell/community-image |
|---|---|---|---|
| Reproducerbarhet | Beror på att SVN och Maven är nåbara varje gång | Låsta källor, en byggd artefakt per version (`BUILDINFO`) | Okänt ursprung och version |
| Säkerhet | Ok | Ok: ingen hemlighet i imagen, icke-root, internt nät | Måste granskas, oklart innehåll |
| Underhåll | Långsam start, kräver JDK och SVN i bygget | En `versions.env`, bygg och testa vid uppgradering | Beroende av tredje part |
| Enkel start för agenter | Minuter till tiotals minuter | Sekunder när imagen finns | Sekunder |
| Finns? | – | – | Ingen officiell image hittad (Nordic Peak publicerar ingen) |

**Val: B.** Dockerfilen bygger distributionen i ett byggsteg (alternativ A, en gång) och producerar en image som återanvänds för alla instanser. Imagen kan senare publiceras till ett privat register (t.ex. `ghcr.io/<org>/opene-sandbox:<opene-commit>-oh<rev>`) så att agenter slipper bygga.

## 4. Sandboxlägen
| Läge | Implementation | Status |
|---|---|---|
| **Persistent** | `sandboxctl create` → instansen `default`, namngivna volymer, `stop/start` behåller data | [Byggt] |
| **Ephemeral** | `sandboxctl create --ephemeral` → unikt namn, egna portar, nät och volymer, `destroy` utan extra bekräftelse | [Byggt] |
| **Remote** | `sandboxctl remote <url>` gör bara läs-endast HTTP-kontroller. Den befintliga gemensamma sandboxen ändras inte | [Byggt] (inte kört mot den gemensamma miljön) |

## 5. Säkerhet och isolering
- Ingen produktionsdatabas: egen MariaDB per instans på internt Docker-nät.
- Ingen e-post ut: DirectMailSender pekas mot Mailpit, och Mailpit vidarebefordrar inget.
- Ingen väg ut: `opene`, `db` och `mail` sitter på ett `internal: true`-nät utan utgående trafik. Endast nginx-proxyn har ett andra nät och publicerar på 127.0.0.1.
- Minsta privilegier: icke-root i OpenE-containern, `no-new-privileges`, `cap_drop: ALL` (db får bara CHOWN, SETUID, SETGID och DAC_OVERRIDE).
- Hemligheter: slumpas per instans, lagras utanför Git (`~/.opene-sandbox`, 600). `.env`, rapporter och resultat är git-ignorerade.
- `reset`/`destroy` kräver `--yes` för persistenta instanser.
- Syntetiska konton och domänen `sandbox.invalid`.

## 6. Agentintegration
Sandboxmotorn är `sandboxctl` + `compose.yaml` + imagen, utan Claude-specifik kod. En agent behöver bara:
1. `sandboxctl status --json` (exit 0 = frisk, 3 = saknas).
2. Vid behov `sandboxctl create` eller `start`.
3. Läsa URL och lösenord (`sandboxctl info --secrets`) och köra webbläsar- eller HTTP-arbete.
4. Importera, testa och exportera `.oeflow` enligt skillens arbetsflöden.
5. `sandboxctl test` som regressionskontroll. `reset --yes` eller `destroy` när miljön ska rensas.

## 7. Upptäckta och lösta startproblem (verifierat 2026-11-20)
Provkörning av den byggda webbroten på Windows (Tomcat 9.0.123 zip, MariaDB 11.8.10 zip, JDK 21.0.13, bara localhost) med samma SQL och skript som containern.

| # | Problem vid första start | Orsak | Lösning i `docker/sql/20-sandbox-pre.sql` | Resultat |
|---|---|---|---|---|
| 1 | FlowEngine "upgrade failed at version 137", sedan "Duplicate column 'firstSubmitted'" vid v53 vid varje omstart | Migrering 137 kräver åtgärderna `FlowBrowserModule.pay`/`.multisigning`, som demodumpen saknar. DDL kan inte rullas tillbaka, så halvkörd migrering låser databasen | Lägg in åtgärderna innan första start. Vid fel: **återställ från ren bas** (reset), laga inte för hand | FlowEngine v52 → **v169** |
| 2 | Temp- och filkataloger saknas (FlowAdmin, FlowBrowser, Mina sidor, Ärenden, PDF, XML, filuppladdning, SiteProfile) | Demodumpen saknar sökvägarna | Inställningar per modulklass → `/var/lib/opene/...` (volym). Obs: FlowAdmins id är `pdfFormFilestore` | Inga fel |
| 3 | "No format validation handler found" | `FormatValidationHandlerModule` är inte registrerad | Registreras som förgrundsmodul | Validerare E-post, Personnummer, Positivt heltal finns |
| 4 | Notishanteraren migrerar inte (v3) | Modulen är avstängd i dumpen | Aktiveras. Med `innodb-strict-mode=OFF` går v3 → **v43** utan radstorleksfelet som Codex mötte | Inga fel |
| 5 | ClassNotFound för TextTagAdmin, favoriter, NewEvents, e-posttest | Klasser som inte finns i Open-ePlatform-repot | Modulerna stängs av | Inga fel |
| 6 | Alla sidor omdirigeras till `https://<host>/` utan port | Rotsektionen "Hem" har `requiredProtocol = HTTPS` | `NULL` på sektioner och moduler | 200 |
| 7 | Demokontot `admin` med känd SHA1-hash | – | Inaktiveras. Egna PBKDF2-konton seedas | Inloggning ok |

Efter åtgärderna: **0 ERROR** i OpenHierarchys systemlogg, uppstart på 13–17 s.

## 8. Verifieringsstatus

| Acceptanskriterium | Status | Bevis |
|---|---|---|
| OpenE byggs från dokumenterad, reproducerbar version | **[Testat]** (Git Bash + JDK 21) / [Byggt] i Dockerfile | `BUILDINFO`: opene a34ab28, OH r3190, utils r1649, purecaptcha r15. 188 jar |
| MariaDB initieras med ren testdatabas | **[Testat]** lokalt | Demodump + sandboxjusteringar laddade. Migreringar till FlowEngine v169, notiser v43, användare v10 |
| Tomcat startar OpenE | **[Testat]** lokalt | "Server startup in 16785 ms", 0 ERROR i systemloggen |
| Applikationen svarar på HTTP | **[Testat]** | `/` 200 `text/html;charset=ISO-8859-1` |
| Testanvändare kan logga in | **[Testat]** | `sandbox-admin` (PBKDF2 från seed-skriptet). Negativt test med fel lösenord fallerar som väntat |
| Administrationsgränssnittet kan öppnas | **[Testat]** | `/flowadmin` visar "Lägg till e-tjänst". useradmin, systemadmin, flowinstanceadmin, minasidor och statusgrupper ger 200 |
| En syntetisk testtjänst kan visas | **[Testat]** | `provansokan.oeflow` importerad med HTTP-multipart → `showflow/1` (1 steg, 1 fråga). Export 11 kB `text/oeflow` |
| Stopp/start utan dataförlust | **[Testat]** lokalt (Tomcat-omstart, samma databas) | Tjänst 1 och inloggning kvar efter omstart, röktest 12/12 |
| Återställning till definierat utgångsläge | **[Testat]** lokalt (drop + omladdning) / [Byggt] `sandboxctl reset` | Tre rena körningar i följd gav samma resultat |
| Docker: `compose up`, internt nät, proxy, Mailpit, volymer | **[Byggt]**, ej kört | `docker compose config` validerar. **Docker-motorn saknas på värden (WSL2 ej installerat)** |
| Röktest | **[Testat]** lokalt | `tests/smoke.sh`: 12 OK, 0 FEL |
| Playwright-navigering | se nedan | |

## 9. Kvarstående
1. **Kör hela kedjan i Docker** när WSL2 och Docker Desktop fungerar: `sandboxctl create`, `test`, `stop/start`, `reset`, `--ephemeral`, två instanser samtidigt.
2. **Frågetyper.** Distributionen har 8 aktiva frågetyper (text, textarea, kryssruta, radio, rullista, fil, kontakt privatperson/organisation). Information, Datum/Tid, Dynamisk/Statisk tabell, Bokning, Kö och de alternativlistbaserade frågorna finns **inte i Open-ePlatform-repot** utan i Nordic Peaks separata AGPL-arkiv (se `moduler.md`). Import av Funktionslabbet ger "Frågetypen … hittades inte" för 5 frågor. Nästa steg: bygg dem från arkivet med samma `build_jar`-mönster.
3. Registrera redan byggda men oregistrerade frågetyper (Stopp, Filinfo, PUD, Barn, Träd, Karta, Personuppgiftshantering) med modulrad, XSL-inställningar och e-tjänsttypens tillåtna lista.
4. SiteProfile står på v6. Codex beskrev en manuell v8-DDL för domänsortering. Behövs först när profildomäner sparas.
5. Spegla SVN-källorna (r3190, r1649, r15) till en arkiverad tarball med checksumma, så att bygget inte beror på `svn://` port 3690.
6. Publicera imagen i ett privat register så att agenter inte behöver bygga.
7. Fler testroller och syntetiska e-tjänster som fixtures (`fixtures/`), plus Playwright-flöden för import, inskick och handläggning.

## Portabel Windows-version, pågående
Se [utvecklingsstatus](../sandbox/PORTABLE-STATUS.md). Grund-exe:n är provad; integrationer och leverans är ännu inte färdiga.

