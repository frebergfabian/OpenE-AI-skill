# Statusar, handläggning och ärendelivscykel

Verifierat 2026-11-12 på ärende 794 (Claude-labbet A) och 795 (Claude-labbet C). Samma konto var sökande och handläggare. Separata roller återstår, se [experimentloggen](experiment/logg-2026-11-12.md).

## Statusmodellen
En status har:
- **Innehållstyp** (`contentType`): `NEW` (sparat ej inskickat), `WAITING_FOR_MULTISIGN`, `WAITING_FOR_PAYMENT`, `SUBMITTED`, `IN_PROGRESS`, `WAITING_FOR_COMPLETION`, `ARCHIVED`. Styr var ärendet syns på Mina sidor (Ej inskickade / Pågående / Avslutade).
- **Åtgärdsmappning** (`actionID`): systemet byter automatiskt till statusen vid
  - `com.nordicpeak.flowengine.FlowBrowserModule.save` (obligatorisk)
  - `…FlowBrowserModule.submit` (obligatorisk)
  - `…FlowBrowserModule.pay`, `…FlowBrowserModule.multisigning`
  - `com.nordicpeak.flowengine.UserFlowInstanceModule.submitcompletion`
- **Rättigheter:** `isUserMutable` (sökanden får ändra, vilket ger komplettering), `isUserDeletable`, `isAdminMutable`, `isAdminDeletable`, `isRestrictedAdminDeletable`, `requireSigning` (handläggarsignering).
- **Krav vid byte:** `addExternalMessage`, `addInternalMessage`, förvalda meddelandemallar. [O] `addExternalMessage` tvingades inte fram vid byte via `/flowinstanceadmin/status/<id>`.
- **Meddelanden:** `newExternalMessagesDisallowed`, `newExternalMessagesAllowedDays`.
- **Åtkomst:** per användare/grupp och per föregående status. Beskrivning för sökande respektive handläggare, handläggningstid i dagar.

### Statusgrupper (`/flowadmin/standardstatuses`)
| Grupp | Statusar |
|---|---|
| **Standard** | **tom (0 statusar)** |
| Enkel | Sparat, Inskickat, Under behandling, Arkiverat |
| Betalning | + Väntar på betalning |
| Betalning och komplettering | + Väntar på betalning, Väntar på komplettering, Kompletterat |
| Flerpartsignering (… och betalning/komplettering) | + Väntar på flerpartsignering … |

**Fallgrop:** **Lägg till e-tjänst** med statusgrupp Standard och "Lägg till standardstatusar" ger en tjänst **utan statusar**. Välj Enkel eller en grupp med komplettering, eller kör **Ersätt statusar med statusgrupp** efteråt.

Publicering kräver att obligatoriska mappningar (save, submit) finns: "E-tjänsten går inte att publicera då dess statusar inte innehåller samtliga obligatoriska statusmappningar". **Aktivera** utan publiceringsdatum ger "Ej publicerad" och stoppas inte av kontrollen.

## Komplettering (verifierat)
1. Status A: "Väntar på komplettering", `WAITING_FOR_COMPLETION`, `isUserMutable=true`, beskrivning till sökanden.
2. Status B: "Kompletterat", `IN_PROGRESS`, `actionID=…UserFlowInstanceModule.submitcompletion`.
3. Handläggaren byter till A. Sökanden ser statusbeskrivningen, **Ändra ärende** och en **Ändra**-länk per fråga (`/minasidor/flowinstance/<flowID>/<instanceID>?step=…`).
4. Sökanden ändrar och väljer **Skicka in**, och ärendet går **automatiskt** till B. Historiken visar "Inskickat … Kompletterat".

## Publicering
- **Aktiverad + Publiceringsdatum** krävs för "Publicerad (datum)". Enbart aktiverad ger "Ej publicerad".
- Publicering låser frågor, steg och regler. Bygg och testa färdigt först.
- "Dölj e-tjänsten på förstasidan" (`hideFromOverview`) är användbart för tester.

## Handläggarens vy `/flowinstanceadmin/overview/<instanceID>`
| Åtgärd | Sökväg / formulär | Verifierat |
|---|---|---|
| Visa ärende | `/flowinstanceadmin/preview/<id>` | |
| Ändra ärende (handläggare) | `/flowinstanceadmin/flowinstance/<flowID>/<id>` | |
| Ändra status | `/flowinstanceadmin/status/<id>` (`statusID`) | ✔ |
| Ändra handläggare | `/flowinstanceadmin/managers/<id>` (`userID`, flera möjliga) | ✔ |
| Meddelande till sökande | multipart `?externalmessage` (`externalmessage`, `attachments`) | ✔ sökanden ser Meddelanden (1) |
| Intern notering | multipart `?internalmessage` | ✔ syns inte för sökanden |
| PDF / XML | `/flowinstanceadmin/pdf/<id>/<eventID>`, `/xml/<id>/<eventID>` | ✔ |
| Flagga, ta bort | finns | |

Översiktens räknare: akut, mina, flaggade, aktiva (andras), obehandlade (ingen handläggare). Ett nytt ärende utan automatisk tilldelning hamnar i **obehandlade**.

## Automatisk tilldelning `/flowadmin/updateautomanagerassignment/<familyID>/<flowID>`
- **Vid nytt ärende:** regler på attributvärden (användare och/eller grupper med begränsad behörighet), tilldelning när ingen regel matchar, handläggare som alltid tilldelas.
- **Vid statusbyte:** status, lägg till eller ersätt handläggare, skicka notis, villkor på attribut.
- Verifierat: attributregel → rätt handläggare vid inskick.

## Notiser `/flowadmin/updatenotifications/<flowID>`
- Till sökande: e-post/SMS vid statusbyte, nytt meddelande, inskickat, arkiverat. Till handläggare: nytt meddelande, intern notering, tilldelning, komplettering, statusbyte, förfallet, läskvitto. Globala adresser per händelse. Mallvariabler: `$flow.name`, `$flowInstance.flowInstanceID`, `$contact.firstname`, `$manager.firstname`.
- En ny tjänst får sandboxens globala standardmallar, med tre sökandenotiser påslagna. Vid import som ny familj återställs notiserna till standard. Vid ny version följer de med.
- **Säkerhet i labb:** utan kontaktfråga går e-posten till kontots adress. Stäng av sökandenotiser eller använd testkonton med syntetisk e-post innan något skickas in.

## Mina sidor
- Ej inskickade (status `NEW`): Fortsätt / Ta bort. Pågående: status och "Till ärende". Avslutade (`ARCHIVED`).
- Sökandens översikt `/minasidor/overview/<flowID>/<id>`: status + statusbeskrivning, historik, meddelanden, PDF, XML, bilagor (`/minasidor/iquery/…?file=`).
- Arkiverat ärende kan inte ändras, men sökanden kan fortfarande skicka meddelanden om slutstatusen inte har `newExternalMessagesDisallowed`.

## XML-export av ärende
UTF-8, namespace `http://www.oeplatform.org/version/2.0/schemas/flowinstance`. `Header` (Flow, FlowInstanceID, Status, Poster, Owner, Posted, Editor, FirstSubmitted …) + `Values`, där varje exporterad fråga har sitt XML-elementnamn och fälten som underelement. Ögonblicksbild från inskicksögonblicket (event-ID i URL:en). Schemat hämtas via `/flowadmin/generatexsd/<flowID>`.
