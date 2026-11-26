# Open ePlatform-skill för Claude

En svensk arbetsguide för att bygga, importera och testa e-tjänster i Open ePlatform. Skillen ligger i `opene-platform/` och kan paketeras som en Claude-skill med `SKILL.md` i skillmappens rot.

Dokumentationen utgår från ett lokalt labb med syntetiska data och använder miljövariabler för lokala anslutningar.

## Testa i Claude Code på Windows

Hämta [förhandsreleasen v0.5.0-dev](https://github.com/frebergfabian/OpenE-AI-skill/releases/tag/v0.5.0-dev). Packa upp `opene-platform.skill` (zip) till projektets `.claude/skills/` så att `.claude/skills/opene-platform/SKILL.md` finns. Be Claude Code använda skillen och starta sandboxen. Se [startinstruktioner och begränsningar](opene-platform/references/portable-sandbox.md).

Förhandsversionen är till för praktiska tester och uppfyller ännu inte samtliga ursprungliga acceptanskriterier.

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

Valideringen kontrollerar frontmatter, mergekonflikter, interna länkar och att exempel är inaktiverad XML utan DTD eller entiteter. Paketeringen utelämnar `.env`, installerade Node-beroenden, testutdata och inbäddade körmiljöer. Tester av simulatorer använder en tillfällig syntetisk payload. Windows processhantering och full appserveruppstart behöver testas i respektive körmiljö. GitHub Actions bygger skillpaketet som en nedladdningsbar artifact.
