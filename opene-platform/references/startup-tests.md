# Verifierad uppstart av skill och sandbox

Verifieringen den 26 november 2026 körs av `.github/workflows/startup.yml`. Varje jobb bygger skillpaketet, packar upp det i en tom katalog och kör startskripten därifrån. Testdata, lösenord, databas och filarkiv ligger i en separat engångsinstans. Ingen gemensam sandbox eller produktionsintegration används.

Testkällor: [full uppstart, körning 8](https://github.com/frebergfabian/OpenE-AI-skill/actions/runs/45895222814) och [Docker efter Playwright-uppdatering](https://github.com/frebergfabian/OpenE-AI-skill/actions/runs/45894801201/job/140165949657). Windowsinstallationen i körning 8 körs med Windows PowerShell 5.1 utan GitHub-inloggning; efterföljande CLI-steg körs från PowerShell 7.

## Miljöer och omfång

| Miljö | Distribution | Verifiering |
|---|---|---|
| Windows Server 2025 x64, GitHub Actions | Låst Windows-exe v0.5.0-dev; Java och MariaDB ingår | Nedladdning, SHA-256, första start, konton, 12 röktester, 6 webbläsartester, omstart, reset och hälsokontroll |
| Ubuntu 26.04 x64, Docker Engine/Compose | Grundplattform byggd från låsta källor; Tomcat 9.0.123, Java 21, MariaDB 11.8, Mailpit och nginx | Källbygge, tom databas, testkonton, 12 röktester och 6 webbläsartester; samma tester efter stopp och omstart |

Webbläsartesterna kör Chromium i skrivbords- och mobilstorlek. De visar katalogen, loggar in som administratör, öppnar flödesadministration, statusgrupper, användare och handläggning samt kontrollerar att utloggade användare inte når administrationen. På Windows kontrolleras även att samtliga sex automatiska syntetiska tjänster syns i administrationen. Hälsokontrollen efter Windowsomstart och reset kontrollerar bland annat administratörsinloggning och applikationslogg.

Windowsinstallationen använder den redan publicerade exe-filen. Detta är inget nytt fullbygge av Windowskörmiljön från PR:ns Go-källor. Go-källorna har separata enhetstester och korskompilering. PowerShellhämtaren och Dockerbygget körs från det aktuella skillpaketet.

## Rättade installationsproblem

- Dockerstart väntar på inloggningssidan och databasens markering för färdiga testkonton. Tidigare räckte katalogens HTTP 200.
- Kontoseedningen avbryter vid saknat lösenord eller misslyckad SQL och sätter inte färdigmarkeringen efter ett sådant fel.
- Status för en stoppad Dockerinstans ger giltig JSON med HTTP-kod 0 och misslyckad hälsostatus. Tidigare kunde status innehålla ogiltiga tal som `000000`.
- Tomcats konfigurationskatalog skapas med rätt ägare innan containern byter till användaren `opene`. Ett observerat `SEVERE: Unable to create directory for deployment` rättades.
- Dockerbygget använder den publicerade SVN-källspegeln över HTTPS och kontrollerar dess låsta SHA-256. Revisionerna är fortfarande OpenHierarchy r3190, Utils r1649 och PureCaptcha r15. Detta kräver inte tillgång till SVN port 3690 från Dockerbygget.
- `sandboxctl test` installerar de låsta npm-beroendena och Chromium innan webbläsartesterna. Playwright har uppdaterats från 1.55.0 till 1.65.1 för Ubuntu 26.04. Värden behöver Node 22+ och Chromiums systembibliotek. Se [Playwrights systemkrav](https://playwright.dev/docs/intro#system-requirements).
- Röktestet verifierar funktionernas administrationslänkar i stället för ett särskilt temas knapptexter. Textkontrollen konsumerar hela svaret för att undvika SIGPIPE från `grep -q` med `pipefail`.
- Windowshämtaren använder publik HTTPS om GitHub CLI saknar inloggning. Den behöver inget GitHub-konto. Grundläggande HTTP-parsning och avstängd nedladdningsprogress undviker äldre PowerShells IE-beroende och onödigt långsam filhämtning.

- Skillpaketeringen hoppar över npm:s uteslutna beroendekatalog innan symlänkkontrollen körs. Ett regressionstest verifierar både att npm-länkar utelämnas och att symlänkar i själva skillmaterialet fortfarande avvisas.

## Användning och begränsningar

Installera det kombinerade skillpaketet från PR:ns paketeringsartifact. Skillpaketet i den äldre v0.5.0-dev-releasen är inte samma dokumentationsversion. Skillen behöver lokal terminalåtkomst för att starta sandboxen och ett webbläsarverktyg för att arbeta i gränssnittet. Windowsportarna väljs dynamiskt; läs `info --name <namn>` igen efter omstart.

Verifieringen gäller uppstart och den beskrivna navigationen. Den verifierar inte alla Cesam-exporter, alla frågetyper eller alla leverantörskontrakt. Windows innehåller sex automatiska tjänster och lokala integrationssimulatorer; Docker innehåller grundplattformen med Mailpit och är inte samma tilläggsdistribution. Tidigare integrationsprov finns i [portabel status](../sandbox/PORTABLE-STATUS.md).

Windows 11, en vanlig användares personliga Windowsprofil, Docker Desktop/WSL2 och macOS har inte provats i denna körning. Mobiltestet är webbläsaremulering. Inga verkliga e-legitimationer, betalningar eller verksamhetssystem används. Windows Java-proxy utgör inte en fullständig spärr för råa nätverkssocketar.
