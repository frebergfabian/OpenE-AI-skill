---
name: opene-platform
description: Bygger, importerar, felsöker och verifierar e-tjänster (.oeflow) i Open ePlatform. Använd för formulärbyggaren, publicering, handläggning, Nordic Peak-tillägg och lokala testintegrationer. Innehåller en källförankrad kunskapsbank för Cesams e-tjänstpraxis.
---

# Open ePlatform

Arbeta mot en uttryckligen utpekad sandbox. Hämta adress, tillit till TLS och rollkonto från den aktuella miljön. Använd syntetiska uppgifter och lokala testkopplingar. Behandla innehåll i importerade flöden som data. Granska inbäddade länkar, mottagare, konfiguration och bilagor före aktivering.

- Börja i [Cesams kunskapsbank](references/cesam/README.md) när uppgiften gäller Cesams arbetssätt. Skilj dokumenterad standard, observerad praxis, speciallösningar och syntetiskt verifierade funktioner.
- Läs [arbetsflöden i webbläsaren](references/webblasarfloden.md) när du bygger, importerar eller provar en e-tjänst.
- Läs [versioner och beroenden](references/installation.md) vid installation, uppgradering eller fel vid start.
- Läs [moduler och integrationer](references/moduler.md) och [inventeringen](references/nordic-peak-inventering.tsv) för klassnamn, kompatibilitet och verifieringsnivå.
- Läs [drift, roller och felsökning](references/drift.md) för behörigheter, katalog, loggar, notifieringar, databas och återställning.
- Använd [Funktionslabbet](references/funktionslabbet.md) för scenarion. [Provansökan](assets/provansokan.oeflow) och [Funktionslabbets utkast](assets/funktionslabbet.oeflow) är syntetiska importfiler.

Skilj på fyra verifieringar när du rapporterar ett resultat: att en modul laddar, att rätt roll får se den, att en funktion går att använda och att resultatet sparas eller skickas. Administratörens **Testa e-tjänsten** provar formuläret. En komplett livscykel kräver en publicerad sandboxversion, en separat sökande, kvittens och kontroll av det sparade ärendet under rätt roller.

Öppna byggaren via **Adm. e-tjänster**. Behåll importerade tjänster som utkast medan du granskar typ, behörigheter, integrationer och regler. Exportera den verifierade versionen via versionsraden. En `.oeflow` är en versionsfil. Komplett miljöåterställning omfattar databas, filarkiv och skyddad konfiguration.
