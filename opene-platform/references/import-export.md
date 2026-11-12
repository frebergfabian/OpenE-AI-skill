# Import och export av .oeflow

## Var
- Ny tjänst: **Adm. e-tjänster → Importera e-tjänst** → välj e-tjänsttyp → `/flowadmin/importflow/<flowTypeID>` (underkategori + fil `flow`, multipart).
- Ny version: tjänstens **Versioner → Importera en ny version** → `/flowadmin/importversion/<flowID>`.
- Export: versionsraden → `/flowadmin/exportflow/<flowID>` (`text/oeflow`, `filename*=UTF-8''…`). Enskild fråga: `/flowadmin/exportquery/<queryID>`. Frågor kan också importeras till ett steg (`/flowadmin/importqueries/<flowID>?step=<stepID>`).

## Filformat
- `<?xml … encoding="ISO-8859-1"?>` från servern. Repots handskrivna filer är UTF-8, och båda fungerar. **Läs serverexporter som latin1.**
- Rot `Flow` → flaggor, `Category`, `FlowFamily` (alias, kontakt, statistik), `Statuses/Status`, `defaultFlowStateMappings/DefaultStatusMapping (actionID + Status)`, `Steps/Step/QueryDescriptors/QueryDescriptor` med typspecifikt barn (`TextFieldQuery`, `RadioButtonQuery` …) och `EvaluatorDescriptors/EvaluatorDescriptor` (med `TargetQueryIDs`, typspecifikt barn, t.ex. `QueryStateEvaluator/RequiredAlternativeIDs`). Deskriptorer är nästlade dubbelt.
- HTML-fält (beskrivning, hjälptext, kort och lång beskrivning) är CDATA eller escapad HTML. **Använd HTML-entiteter för tecken utanför ASCII** (`&auml;`, `&ndash;`), som CKEditor gör.
- **Följer inte med:** handläggare, automatisk tilldelning, notisinställningar, målgrupper/livshändelser. `flowInstanceCount` läcker antal ärenden.

## Vad importen gör (verifierat)
| Fall | Resultat |
|---|---|
| Ny tjänst | Ny familj, **inaktiverad**, alias borttaget, alla ID:n (frågor, alternativ, regler, statusar) nya och konsekvent ommappade |
| Ny version | Ny inaktiv version i samma familj. Tidigare publicerad version påverkas inte. Handläggare (familjenivå) ärvs och notiser följer med |
| Trunkerad eller tom XML | "Den gick inte att tolka innehållet i filen …" |
| `<!DOCTYPE>` / entiteter | Avvisas som ovan |
| Fel filändelse | "Filen … är av en felaktig filtyp. Följande filtyper är tillåtna: oeflow" |
| Okänd `queryTypeID` | "Frågetypen för frågan \"…\" hittades inte." |
| Okänd `evaluatorTypeID` | "Regeltypen för regeln \"…\" hittades inte." (en per regel) |
| **Regel → okänd fråga** | **Importeras tyst**. Regeln får inga mål |
| **Regel → okänt alternativ** | **Importeras tyst**. Regeln utlöses aldrig |
| Saknad `submit`-mappning | Importeras tyst. Admintestet ger kvittens. **Publicering stoppas** |

## Checklista efter import
1. Kontrollera antal steg, frågor och regler på `showflow` mot källan.
2. Exportera igen och jämför antal `TargetQueryIDs/queryID` och `RequiredAlternativeIDs/alternativeID` per regel med källan. Tomma listor betyder en trasig referens.
3. Statusar och mappningar (save, submit, eventuellt submitcompletion). Kontrollera att statusgruppen inte är tom.
4. Lägg till handläggare, automatisk tilldelning och notiser. Stäng av sökandenotiser i labb.
5. Alias: sätt på nytt (importen tar bort det).
6. Granska externa länkar, mottagare och integrationer innan aktivering (se [arbetsflöden](webblasarfloden.md)).
7. Kör **Testa e-tjänsten**, publicera i sandbox och gör ett riktigt inskick. Admintestet sparar inget och avslöjar inte saknade mappningar.

## Programmatisk import (för OpenE Runner)
```js
const fd = new FormData();
fd.append('categoryID', '107');
fd.append('flow', new File([bytes], 'x.oeflow'));
await fetch('/flowadmin/importflow/4', {method: 'POST', body: fd});
// Lyckat: redirect till /flowadmin/showflow/<nyttFlowID>. Fel: samma sida med felruta.
```
Ändra serverexporter i latin1: avkoda med `TextDecoder('iso-8859-1')` och koda tillbaka byte för byte.
