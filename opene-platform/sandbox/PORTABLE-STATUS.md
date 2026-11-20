# Portabel sandbox: utvecklingsstatus

Detta är en experimentell implementation, inte en godkänd release. Den av användaren valda releasedestinationen är det offentliga repot OpenE-AI-skill. Ingen release är publicerad.

## Byggt och verifierat

- OpenE har byggts från låsta källor och en lokal SVN-spegel med SHA-256 `744d32447e3d39b91db333110074c004bcde6981401ea42c233caeed866c63ad`. En reproducerbar byggväg finns i `build/build-experimental.ps1`, men den behöver slutverifieras från en ren commit och ger inte nödvändigtvis bitidentiska filer.
- Den ursprungliga grund-exe:n klarade röktest 12/12, Playwright 6/6 och två samtidiga instanser. Dessa körningar omfattade inte hela tilläggspaketet.
- Separata utvecklingsinstanser har verifierat bestående bokning och kö, PUD, barnrelation, enkel- och flerpartssignering, HTTP, lokal e-post, SMS samt lyckad och avbruten simulerad betalning med sparade ärenden. Förutsättningar och begränsningar finns i `tests/integration-cases-result.txt`.
- En SSBT-fråga har sparat syntetiska företagsuppgifter via en särskild lokal JSON-adapter. Det är inte en verifiering av Bolagsverkets SOAP-kontrakt.
- Första-start-koden importerar fem syntetiska tjänster och konfigurerar testdata. Publicering och omladdning av webbservern är ny kod som ännu inte har verifierats med en ren installation.
- Tilläggens byggresultat och eventuella klasskonflikter dokumenteras av byggskriptet. Byggt betyder inte automatiskt aktiverat eller funktionellt testat.

## Kvar före release

Verifiera den senaste fullständiga exe:n på en ren Windows-användare, inklusive tidsgränser, uppstart utan ERROR, alla automatiska fixtures, röktest och Playwright, omstart, reset och samtidiga instanser. Aktivera och granska alla byggbara tillägg, verifiera samtliga nödvändiga integrationers exakta kontrakt och scenarier, inklusive fler verksamhetssystem och DocumentSigner, samt granska betalningscallbackens återhämtning och idempotens. Därefter behövs en ren byggning med publicerad källspegel, BUILDINFO, checksummor och en GitHub Release.

Java-proxyinställningarna spärrar inte råa socketar. Produkten får inte beskrivas som helt isolerad från all utgående trafik utan ytterligare skydd. Den befintliga gemensamma sandboxen har inte använts eller ändrats.
