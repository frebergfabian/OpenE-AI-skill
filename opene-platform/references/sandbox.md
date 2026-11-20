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

## 7. Status och kvarstående
Se tabellen "Verifieringsstatus" längst ned (uppdateras efter varje provkörning).
