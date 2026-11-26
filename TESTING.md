# Verifiering av den kombinerade skillen

Kör validering och paketering enligt [README](README.md). GitHub Actions kör samma kontroller och sparar skillpaketet som artifact.

Sammanföringen den 26 november 2026 omfattar följande grenar och revisioner.

| Gren | Revision |
|---|---|
| main | 93d73b9 |
| cesam-kunskapsbank | 8b09173 |
| claude/opene-academy | 789ed9a |
| claude/opene-sandbox | 134b714 |
| codex/portable-opene-sandbox | 18f7464 |

Git-historiken bevarar grenarnas commits. Konflikten i SKILL.md löstes genom att behålla Cesam-inriktningen, Academy-referenserna och båda sandboxalternativen. Valet av miljö utgår från den miljö användaren har pekat ut och den tillgängliga värdplattformen.

| Kontroll | Utfall lokalt |
|---|---|
| Skill-frontmatter, konfliktmarkörer, relativa länkar, XML-fixtures | Godkänt |
| Python-tester av paketets innehåll, determinism och negativa fall | 5 godkända |
| Node-tester för sanering av WSDL | 2 godkända |
| Go-tester av arkivhantering och simulatorer | 14 huvudtester godkända, inklusive betalningens pay/cancel |
| Windows x64-testbinär | Korskompilerad med syntetisk testpayload |
| Shell- och JavaScript-syntax | Godkänt |
| Paketering och uppackad skill | Godkänt |

Go-testerna kördes på Linux med Go 1.25.0. Windows processhantering är avgränsad till Windows och har spärrade testfunktioner på andra system. Korskompileringen kontrollerar även Windows-koden. Det genererade testprogrammet innehåller en liten syntetisk ZIP och ska endast användas för tester.

Full uppstart har därefter provats från det uppackade skillpaketet på Windows Server 2025 och Ubuntu 26.04 med Docker. Windows klarar 12 röktester, 6 webbläsartester, omstart och reset. Docker klarar källbygge, första start, 12 röktester och 6 webbläsartester samt samma tester efter omstart. Körresultat: [full uppstart på Windows och Docker](https://github.com/frebergfabian/OpenE-AI-skill/actions/runs/45895222814). Detaljer, avgränsningar och rättade fel finns i [uppstartsprotokollet](opene-platform/references/startup-tests.md).

Windowsprovet använder den publicerade v0.5.0-dev-exe:n. Ett nytt fullbygge av exe-filen ingår inte. Tidigare integrationsprov ersätts inte av dessa uppstartstester.
