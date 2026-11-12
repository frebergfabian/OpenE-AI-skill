# Regler (evaluators) och attribut

## Var regler skapas
Regeln skapas på **källfrågan**: ikonen "Lägg till regel" bredvid frågan, eller `/flowadmin/addevaluator/<queryID>`. Formuläret tar regeltyp, namn och kommentar. Därefter öppnas regeltypens egen konfiguration. **Nya regler är inaktiverade.** Kryssa i **Aktivera regeln**.

## Regeltyper (`evaluatorTypeID`)

| Namn i gränssnittet | Klass (`com.nordicpeak.flowengine.evaluators.`…) | Finns för |
|---|---|---|
| Sätt vikt om frågan är besvarad | `setqueryweightevaluator.SetQueryWeightEvaluationProviderModule` | alla frågor |
| Vikta alternativ | `setweightevaluator.SetWeightEvaluationProviderModule` | radio, kryssruta, rullista |
| Ändra standardläge | `querystateevaluator.QueryStateEvaluationProviderModule` | radio, kryssruta, rullista |
| … baserat på antal barn | `childcountquerystateevaluator.ChildCountQueryStateEvaluationProviderModule` | alla |
| … baserat på beräknad vikt | `weightcalculatedquerystateevaluator.WeightCalculatedQueryStateEvaluationProviderModule` | alla |
| … baserat på eIDAS-inloggning | `eidas.EIDASQueryStateEvaluationProviderModule` | alla |
| … baserat på kommunkod | `municipalitycode.MunicipalityCodeQueryStateEvaluationProviderModule` | alla |
| … baserat på om användaren är inloggad | `loggedin.LoggedInQueryStateEvaluationProviderModule` | alla |
| … baserat på sökandes ålder | `userage.UserAgeQueryStateEvaluationProviderModule` | alla |
| … baserat på vikt | `weightquerystateevaluator.WeightQueryStateEvaluationProviderModule` | alla |

Den äldre klassen `calculatedweightquerystateevaluator…` finns inte. Import ger "Regeltypen för regeln … hittades inte".

## Ändra standardläge (`/querystateevaluationprovider/config/<id>`)
- Aktiveringsläge: `ANY` (något markerat alternativ) eller `ALL` (samtliga).
- Alternativ: källfrågans alternativ (numeriska ID:n per frågetyp).
- Frågeläge: `VISIBLE`, `VISIBLE_REQUIRED`, `HIDDEN`.
- Målfrågor: valfria frågor i **alla steg** (`queryID`).
- "Återställ INTE frågornas standardläge automatiskt" behövs när flera regler påverkar samma fråga.
- **Verifierat:** Ja → Telefonnummer obligatoriskt, Nej → tillbaka till Dold. Rullista "Utanför kommunen" → Stoppfråga synlig, vilket tar bort knappen Nästa steg och stoppar steget.

### Hur utvärderingen sker
Varje ändring i en källfråga skickar ett POST till flödets URL (`/oversikt/flow/<flowID>[/<instanceID>]` eller `/flowadmin/testflow/<flowID>`). Servern svarar med JSON:
```json
{"QueryModifications":[{"queryID":"4140","queryType":"TextFieldQueryInstance","action":"HIDE"}]}
```
`SHOW` skickar med HTML för frågan. Ett tomt svar `{}` betyder att inget ändras. Det gör att regler kan testas utan rendering genom att läsa svaret.

## Vikter
- **Vikta alternativ** (`/lab-setweight/config/<id>`): namngiven **Typ av vikt** (t.ex. `kontaktvikt`), en vikt per alternativ och valet "använd bara högsta markerade".
- **Ändra standardläge baserat på vikt** (`/lab-weightquerystate/config/<id>`): typ av vikt, intervall (från och med / till och med) och frågeläge. **Har inga målfrågor**, utan styr frågan regeln sitter på. Skapa den alltså på frågan som ska visas eller döljas.
- [?] En viktregel på en fråga i samma steg som viktkällan, kombinerad med en "Ändra standardläge"-regel på samma mål, slog inte till (live-svaret var `{}`). Ej utrett: rangordning mellan regler kontra utvärdering först vid stegbyte.

## Attribut
- Textfält, radio, kryssruta, rullista, datum, textarea och filuppladdning har **Spara fältets värde som attribut** + **Attributnamn**. Radio/rullista/kryssruta kan ha "Värde för attribut" per alternativ (annars alternativets namn).
- Attribut används av **automatisk tilldelning av handläggare** (attributnamn + värden, ett per rad, inverterbart, "inkludera osatt attribut"), av statusregler och av notismallar (`$flowInstance…`).
- **Verifierat:** rullistan `stadsdel` → attribut → regel `stadsdel = Norra Exempelköping` → handläggaren Sam tilldelades vid inskick.
- Attribut på datumfrågan sätts för startdatum (`startdatum`).

## Import och regler
Regler refererar till fråge-ID och alternativ-ID. Vid import mappas de om konsekvent. **Ogiltiga referenser tas bort tyst** (regeln blir kvar utan mål eller alternativ). Kontrollera reglerna efter import. Se [import och export](import-export.md).
