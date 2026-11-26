# Open ePlatform-skill för Claude

En svensk arbetsguide för att bygga, importera och testa e-tjänster i Open ePlatform. Skillen ligger i `opene-platform/` och kan paketeras som en Claude-skill med `SKILL.md` i skillmappens rot.

Dokumentationen utgår från ett lokalt labb med syntetiska data och använder miljövariabler för lokala anslutningar.

## Installera och börja använda

Hämta artifacten **opene-platform-skill** från den senaste godkända [paketeringskörningen för PR #3](https://github.com/frebergfabian/OpenE-AI-skill/actions/workflows/skill.yml?query=branch%3Acodex%2Fcombined-testable-skill). Packa först upp artifactens ZIP och sedan `opene-platform.skill` (också ZIP) till projektets `.claude/skills/`. Kontrollera att `.claude/skills/opene-platform/SKILL.md` finns. Be Claude Code använda `opene-platform` och starta en lokal sandbox. Skillen behöver lokal verktygsåtkomst; en chatt utan lokal terminal kan inte starta servern.

På Windows x64 behövs PowerShell och en webbläsare. Java och MariaDB ingår i den separat hämtade körmiljön. Från projektets rot:

```powershell
powershell.exe -NoProfile -ExecutionPolicy Bypass -File .claude/skills/opene-platform/sandbox/get-sandbox.ps1 -Name demo -NoBrowser
$exe = Join-Path $env:LOCALAPPDATA 'OpenE-Sandbox/bin/OpenE-Sandbox-v0.5.0-dev-win-x64.exe'
& $exe info --name demo --secrets
```

Öppna URL:en från `info` och logga in som `sandbox-admin` med det lokalt genererade lösenordet. Kör `info` igen efter omstart: Windowsportarna väljs dynamiskt. [Windowsguiden](opene-platform/references/portable-sandbox.md) beskriver konton, simulatorer och begränsningar. Det äldre skillpaketet i v0.5.0-dev-releasen innehåller inte PR #3:s sammanslagna dokumentation; använd PR-artifacten ovan. Windowskörmiljön hämtas fortfarande från den låsta releasen.

På Linux med Docker Engine och Compose v2:

```bash
bash .claude/skills/opene-platform/sandbox/sandboxctl create demo
bash .claude/skills/opene-platform/sandbox/sandboxctl info demo --secrets
bash .claude/skills/opene-platform/sandbox/sandboxctl test demo
```

Första Dockerbygget hämtar och kompilerar beroenden. Bash och curl behövs på värden; Node/npm behövs för webbläsartesterna. Se [Dockerberoenden](opene-platform/sandbox/README.md) och [verifierade resultat](TESTING.md). Miljöerna har olika omfång: Windows inkluderar sex automatiska tjänster och integrationssimulatorer; Docker bygger grundplattformen med lokal e-post. Förhandsversionen uppfyller ännu inte samtliga integrations- och acceptanskriterier.

## Testa och installera den kombinerade skillen

Kör från repositoryts rot med Python 3.10 eller senare.

```bash
python3 -m unittest discover -s tests -v
python3 tools/skill.py --output dist/opene-platform.skill
```

Paketet innehåller en toppmapp `opene-platform` med `SKILL.md`, Cesams kunskapsbank, Academy-referenser, syntetiska exempel och sandboxens källkod och startskript. Packa upp det i projektets `.claude/skills/` för Claude Code. ZIP-formatet kan också användas av klienter som stöder `.skill`-import. Välj testmiljö enligt skillens instruktioner.

Med Node 18+ och Go 1.25+ kan du även köra simulatorernas tester.

```bash
python3 tools/skill.py --runtime-tests --output dist/opene-platform.skill
```

Valideringen kontrollerar frontmatter, mergekonflikter, interna länkar och att exempel är inaktiverad XML utan DTD eller entiteter. Paketeringen utelämnar `.env`, installerade Node-beroenden, testutdata och inbäddade körmiljöer. Tester av simulatorer använder en tillfällig syntetisk payload. Workflowen `Full sandbox startup` testar installation av det uppackade paketet på Windows och Ubuntu/Docker, inloggning, webbläsarnavigation och omstart. Se TESTING.md för körresultaten. GitHub Actions bygger skillpaketet som en nedladdningsbar artifact.
