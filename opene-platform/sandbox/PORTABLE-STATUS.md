# Portabel sandbox: utvecklingsstatus

Detta är en experimentell implementation. Förhandsversionen `v0.5.0-dev` publiceras i det offentliga repot OpenE-AI-skill för praktiska tester med Claude Code på Windows. Den är inte godkänd mot samtliga ursprungliga acceptanskriterier. Se releasens testprotokoll för verifieringen av den distribuerade exe-filen.

## Byggt och verifierat

- OpenE har byggts från låsta källor och en lokal SVN-spegel med SHA-256 `744d32447e3d39b91db333110074c004bcde6981401ea42c233caeed866c63ad`. `build/build-experimental.ps1` kördes färdigt från ren commit `6cf8e0c` och skapade exe, BUILDINFO och SHA-256. Det är ett byggprov av en tidigare utvecklingsversion; den slutliga revisionen behöver byggas och verifieras separat. Bygget är inte verifierat som bitidentiskt.
- Den ursprungliga grund-exe:n klarade röktest 12/12, Playwright 6/6 och två samtidiga instanser. Dessa körningar omfattade inte hela tilläggspaketet.
- Separata utvecklingsinstanser har verifierat bestående bokning och kö, PUD, barnrelation, enkel- och flerpartssignering, HTTP, lokal e-post, SMS samt lyckad och avbruten simulerad betalning med sparade ärenden. Förutsättningar och begränsningar finns i `tests/integration-cases-result.txt`.
- En SSBT-fråga har sparat syntetiska företagsuppgifter efter lokala SOAP-anrop med de genererade klienterna för företagsengagemang och grunduppgift UD0001. Övriga SSBT-uppgifter och rollintyg är inte simulerade.
- Betalningssimulatorn implementerar en begränsad del av [Swedbank Pays payment order](https://developer.swedbankpay.com/checkout-v3/get-started/payment-request/), [redirect-operation](https://developer.swedbankpay.com/checkout-v3/get-started/display-ui/) och [callback](https://developer.swedbankpay.com/checkout-v3/features/payment-operations/callback/). Den är inte en fullständig implementation av leverantörens API.
- Första-start-koden importerar och publicerar sex syntetiska tjänster och konfigurerar testdata. Detta har provats i rena utvecklingsinstanser, inklusive flerpartssignering. Dokumentsignering med en automatiskt konfigurerad lokal webbplatsprofil har också verifierats med ett sparat, signerat PDF-dokument.
- Tilläggens byggresultat och eventuella klasskonflikter dokumenteras av byggskriptet. Byggt betyder inte automatiskt aktiverat eller funktionellt testat.
- En granskning hittade skarpa adresskonstanter i vissa leverantörers kod och WSDL-resurser i de äldre utvecklingsbyggena. Byggkoden har därefter ändrats för att neutralisera kända konstanter och WSDL/XSD-mål samt neka kvarvarande fjärrmål i WSDL. Det sanerade fullbyggets tolv WSDL-filer har kontrollerats; alla tänkbara hårdkodade mål behöver fortfarande granskas. Äldre provbyggen ska inte distribueras.

## Kvar före full acceptans

Verifiera den senaste fullständiga exe:n på en ren Windows-användare, inklusive tidsgränser, uppstart utan ERROR, alla automatiska fixtures, röktest och Playwright, omstart, reset och samtidiga instanser. Aktivera och granska alla byggbara tillägg, verifiera samtliga nödvändiga integrationers exakta kontrakt och scenarier, inklusive fler verksamhetssystem, samt granska betalningscallbackens återhämtning och idempotens. Därefter behövs en ren byggning med publicerad källspegel, BUILDINFO, checksummor och en GitHub Release.

Java-proxyinställningarna spärrar inte råa socketar. Produkten får inte beskrivas som helt isolerad från all utgående trafik utan ytterligare skydd. Den befintliga gemensamma sandboxen har inte använts eller ändrats.
