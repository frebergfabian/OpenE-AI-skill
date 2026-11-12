# Gränssnittskarta och behörigheter

Verifierad i sandboxen 2026-11-12 med ett konto som är Systemadministratör + Portaladministratör + E-tjänsteutvecklare + E-tjänstepublicerare. Sökvägarna är relativa till portalens rot.

## Toppmeny

| Meny | Innehåll | Sökväg |
|---|---|---|
| Tjänster | Katalog med livshändelser, kategorier, sökning, kort- eller listvy | `/`, `/oversikt`, `/oversikt/overview/<familyID>` |
| Mina sidor | Ej inskickade, pågående och avslutade ärenden. Mina uppgifter, organisationer, köplatser, bokningar, abonnemang | `/minasidor`, `/minasidor/overview/<flowID>/<instanceID>` |
| Handläggning | Ärendeöversikt (akut, mina, flaggade, aktiva, obehandlade), sökning, bokningar, köplatser, dokumentsignering | `/flowinstanceadmin`, `/bokningar`, `/koplatser`, `/dokumentsignering` |
| Redaktörer | Sidor, driftmeddelanden, widgetar, menyer | `/administration/pageadmin`, `/operatingmessages`, `/widgetadmin`, `/menuadmin` |
| Adm. e-tjänster | Lista, lägg till, importera, statusgrupper, kategorier, livshändelser, katalog, kalendrar, köer, interna listor, kontoplaner, transaktioner | `/flowadmin`, `/flowadmin/addflow`, `/flowadmin/importflow`, `/flowadmin/standardstatuses`, `/flowadmin/flowtype/4`, `/livshandelser`, `/administration/kalendrar`, `/administration/koer`, `/lab-lists` |
| Systemadm. | Användare, grupper, profiler, moduler och sektioner, taggar, datakällor, systeminfo, trådinfo, inloggade användare | `/administration/useradmin`, `/groupadmin`, `/profiler`, `/systemadmin`, … |

## E-tjänstens adminsida `/flowadmin/showflow/<flowID>`

Flikar (ankare): Grundinformation · Förvaltningsinformation · Frågor och steg · Förhandsgranska frågor (`/flowadmin/testflowallsteps/<id>`) · Testa e-tjänsten (`/flowadmin/overview/<id>` → `/flowadmin/testflow/<id>`) · Blankett (PDF-mall) · Statusar · Handläggare · Notifieringar · Användarundersökningar · Meddelandemallar · Versioner · Tillägg och extrafunktioner · Livshändelser · Ändringslogg.

| Åtgärd | Sökväg | Anmärkning |
|---|---|---|
| Grundinformation | `/flowadmin/updateflow/<id>` | Aktivering, publiceringsdatum, inloggning, signering, meddelanden, dölj-flaggor, alias, kontakt |
| Steg | `/flowadmin/addstep/<id>`, `/updatestep/<stepID>`, `/deletestep/<stepID>` | |
| Fråga | `/flowadmin/addquery/<id>?step=<stepID>` → `/<modulalias>/config/<queryID>` | Se [frågetyper](fragetyper.md) |
| Regel | `/flowadmin/addevaluator/<queryID>` → `/<regelalias>/config/<evaluatorID>` | Se [regler och attribut](regler-och-attribut.md) |
| Status | `/flowadmin/addstatus/<id>`, `/updatestatus/<statusID>`, `/replaceflowstatuses/<id>`, `/sortstatuses/<id>` | |
| Handläggare | `/flowadmin/updatemanagers/<familyID>/<id>` | Familjenivå |
| Automatisk tilldelning | `/flowadmin/updateautomanagerassignment/<familyID>/<id>` | Familjenivå |
| Notiser | `/flowadmin/updatenotifications/<id>` | |
| Export/import | `/flowadmin/exportflow/<id>`, `/flowadmin/importflow/<flowTypeID>`, `/flowadmin/importversion/<id>`, `/flowadmin/exportquery/<queryID>`, `/flowadmin/importqueries/<id>?step=` | Se [import och export](import-export.md) |
| Ny version / kopia | `POST /flowadmin/copyflow` (radio `flowID`, knapp "Lägg till en ny version" eller `new_family`) | |
| Avpublicera familj | `/flowadmin/unpublishflowfamily/<familyID>/<id>` | POST via bekräftelse |
| Ta bort regel/fråga | `/flowadmin/deleteevaluator/<id>`, `/deletequery/<id>` | **Kräver POST** (GET ger 403) |

## Behörighetsmodell (verifierad)

- **Admin-flaggan räcker inte för att bygga.** Kategorin (e-tjänsttypen) har egna administrationsgrupper. Utan medlemskap blir **Lägg till e-tjänst** utan kategorier och **Aktivera** är utgråad. Lägg kontot i e-tjänsttypens grupper (här E-tjänsteutvecklare och E-tjänstepublicerare). Ändringen slog igenom utan ny inloggning.
- **Admin-flaggan ger inte ärendeåtkomst.** Handläggning visar bara ärenden för familjer där kontot eller dess grupp är handläggare.
- **Tjänstens handläggare** (grupper eller användare) styr vilka som kan tilldelas och se ärenden. Endast de visas i `/flowinstanceadmin/getmanagers/<flowID>/<instanceID>`.
- Statusar kan begränsas per användare/grupp (`useAccessCheckByUser`) och per föregående status (`useAccessCheckByStatus`).
- Sökanden ser bara egna ärenden. Interna noteringar syns aldrig för sökanden (verifierat).

## Låsningar

| Läge | Vad är låst |
|---|---|
| Versionen är **publicerad** | Frågor, steg och regler (redigering ger 403). Grundinformation, statusar, handläggare och notiser går att ändra |
| Versionen har **ärenden** | Frågor och steg. Versionen kan inte tas bort |
| Lösning | Ny version via kopiering eller import, bygg, testa och publicera |

## Testkonton i sandboxen (synteter)

`lab-medborgare` (Medborgare, personnummer-attribut), `lab-medborgare-2`, `lab-handlaggare` (Handläggare, Kommunservice), `lab-handlaggarchef`, `lab-utvecklare`, `lab-tjansteadm`, `lab-redaktor`, `lab-portaladmin`, `sandbox-admin`. Lösenord finns inte i repot. Rollbyte kräver att en människa loggar in i webbläsaren.
