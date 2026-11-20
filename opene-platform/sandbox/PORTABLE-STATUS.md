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
