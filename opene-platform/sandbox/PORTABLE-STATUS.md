# Portabel sandbox – utvecklingsstatus 2026-11-20

Detta är ett pågående arbete, inte en release eller ett godkänt paket. `portable` innehåller en Go-starthanterare och en lokal simulator. Utvecklings-exe:n som provades byggdes med Claudes grundwebbrot, utan Nordic Peak-tillägg och utan betalningsadaptern. En senare exe innehåller utvecklingskod för betalningssimulatorn, men även den saknar betalningsadaptern i den inbäddade webbroten.

## Byggt och testat
- Låst OpenE-källbygge kördes lokalt med JDK 21: 188 jar, se ursprungliga BUILDINFO i byggutdata. SVN används fortfarande direkt och är inte speglat.
- Första grund-exe:n: första start lyckades, status rapporterade HTTP, inloggningsformulär, fem konton, administratörsinloggning och simulator. Röktest 12/12 och 0 ERROR i systemlogg. Stopp och omstart med samma data lyckades. Starttider har inte mätts systematiskt.
- Go-enhetstester: arkivets sökvägar, borttagningens gräns, skydd mot cross-origin, okänd integration och PUD-exempelsvar.
- Swedbank-adaptern kompilerar separat. Checkout-implementationen är inte aktiverad, inte end-to-end-testad och har inte säkerhetsgranskats.
- Tilläggsrapporten `addon-build-status.tsv` är en ögonblicksbild av en kompilering; byggd betyder inte aktiverad eller funktionellt verifierad. WeightEvaluationProviders byggdes delvis med API-delar undantagna. Arkivet innehöll 144 projektkataloger i denna verifierade nedladdning.

## Återstår före release
Automatisk modulregistrering och all konfiguration, boknings- och ködata, fixtures, komplett betalningskedja och transaction log, signering, SSBT, barn, SMS och verksamhetssystem med persistenta ärenden, PUD-integrationskonfiguration, verifiering av e-post, fler samtidiga instanser, reset, Playwright, rena Windows-konton, verifierad spärr av utgående trafik, spegling och checksummor av SVN-källor, komplett reproducerbart byggskript, skillens hämtningsskript, samt GitHub Release. Proxyinställningar täcker inte råa socketar. Distributionen får inte beskrivas som isolerad från all utgående trafik. Säkerhetsgranska även simulerade betalningars callback och idempotens innan användning.

Playwright försöktes men startade inte: Chromium headless shell saknades. Ingen release har publicerats. Befintlig gemensam sandbox har inte rörts.

## Fortsatt verifiering
- Playwright 1.55 kördes mot grund-exe:n med installerad Microsoft Edge (`OPENE_BROWSER_PATH`): 6/6 tester gröna, desktop och mobil. Detta är inte ett test av tillägg eller integrationsflöden.
- Två grundinstanser startades samtidigt med egna portar och kontrollerades med `status --json`: båda rapporterade friska; därefter förstördes de två testinstanserna.
- Källspegel `svn-r3190-r1649-r15.tar.gz` har skapats lokalt från exporterna; SHA-256 `744d32447e3d39b91db333110074c004bcde6981401ea42c233caeed866c63ad`. Byggskriptet accepterar spegeln med uttrycklig checksumma. Spegeln är inte publicerad.
- Kod för syntetiska personidentifierare i kontoseedningen är tillagd men ännu inte provad i en ny exe.
- Spegelbygget kördes till slut med `OPENE_SVN_MIRROR` och SHA-256 ovan: OpenE-webbrot med 188 jar. Detta verifierar byggbarhet från spegeln, inte bitidentiska binärer.
- Båda namngivna GitHub-repona är offentliga; destinationen för en intern release måste väljas. Ingen release har publicerats.

## Fortsatt arbete: frågetyper och hämtningsskript
- `get-sandbox.ps1` och `get-sandbox.sh` är tillagda; PowerShell-syntax och avvisning av ogiltigt versionsnamn har provats. Nedladdning kan inte end-to-end-testas innan en release finns.
- En särskild experimentell exe (`0.1.2-dev`) byggdes med Information, Datum/Tid, Dynamisk tabell och deras beroenden. Alla tre moduler migrerade och laddades med 0 ERROR; alla tre tilläts i e-tjänsttypen. Fem syntetiska citizenIdentifier-attribut seedades. Hälsokontrollen har utökats med systemloggen.
- Import av `funktionslabbet.oeflow` med administratörskonto ger fortfarande två fel: frågetyperna Bokning och Kö hittades inte. Importen lyckades inte och inget komplett ärende har testats.
- Den första experimentkörningen saknade AlternativeListHandler och genererade ett ERROR trots att dåvarande hälsokontrollen rapporterade frisk. Beroendet lades till och ny ren instans fick 0 ERROR. Detta bevisar att äldre exe-varianter inte har tillräcklig hälsokontroll.
- Användaren har valt OpenE-AI-skill som releasedestination trots att repot och releasen är offentliga. Ingen release har publicerats.

## Bokning och kö: experiment 2026-11-20
- Den experimentella SQL-registreringen utökades med boknings- och kömoduler. En ren `0.1.3-dev` med de byggda BookingModules och QueueModules startade med 0 ERROR och frisk status.
- `funktionslabbet.oeflow` importerades med HTTP multipart: `POST /flowadmin/importflow/4` gav `302 /flowadmin/showflow/1`. Detta bevisar att frågetyperna hittades och importen accepterades, men inte att en publicerad ansökan fungerar.
- I testinstansen skapades manuellt en aktiverad syntetisk kalender, en framtida tid, en aktiverad kö och handläggargrupp; de importerade frågorna kopplades till dessa. Detta är inte automatiserad seedning och är inte provat genom inskick eller databasverifierad bokning/köplats.
- Modulerna saknar fortfarande verifierad användarbehörighet, notisinställningar, fullständiga fixtures och integrationsflöden. Betalning och andra integrationskategorier är inte end-to-end-testade. Ingen release har publicerats.
