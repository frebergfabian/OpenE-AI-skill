# Frågetyper i e-tjänstebyggaren

Källa: e-tjänsttypen **Kommunens e-tjänster** (`/flowadmin/flowtype/4`) och formuläret **Lägg till fråga** (`/flowadmin/addquery/<flowID>?step=<stepID>`) i sandboxen 2026-11-12. Beskrivningarna nedan är plattformens egna hjälptexter i sammandrag. Klassnamnet är det `queryTypeID` som står i `.oeflow`. Vilka typer som syns styrs av e-tjänsttypens lista över tillåtna frågetyper.

**Lägg till fråga** har bara fem fält: steg, frågetyp, namn, kommentar och standardläge (`VISIBLE` = Valfri, `VISIBLE_REQUIRED` = Obligatorisk, `HIDDEN` = Dold). Typspecifika inställningar görs efter att frågan har skapats.

| Frågetyp | `queryTypeID` | Kräver/kommentar |
|---|---|---|
| Ange adress eller fastighetsbeteckning | `com.nordicpeak.flowengine.queries.pudquery.PUDQueryProviderModule` | Lantmäteriet (sandbox: lokal HTTP-fixtur) |
| Bokningsfråga | `com.nordicpeak.flowengine.booking.queries.BookingQueryProviderModule` | Aktiv kalender. Bara en synlig bokningsfråga per kalender per tjänst, annars risk för överbokning |
| Datum/Tid-fråga | `com.nordicpeak.flowengine.queries.DateTimeQueryProviderModule` | Kan begränsa datum och intervall |
| Filuppladdningsfråga | `com.nordicpeak.flowengine.queries.fileuploadquery.FileUploadQueryProviderModule` | En eller flera filer |
| Företagsuppgiftsfråga (SSBTEN/SSBTGU) | `com.nordicpeak.flowengine.queries.ssbt.CompanyDetailsQueryProviderModule` | Bolagsverket och servicenamn (sandbox: fixtur) |
| Information om barn | `com.nordicpeak.flowengine.queries.childinfoquery.ChildInfoQueryProviderModule` | Navet (sandbox: dummyprovider) |
| Informationsfråga | `com.nordicpeak.flowengine.queries.infoquery.InfoQueryProviderModule` | Visar text och har inget svar |
| Kontaktuppgiftsfråga (interna användare) | `com.nordicpeak.flowengine.queries.internalcontactdetails.InternalContactDetailQueryProviderModule` | Personalkatalog |
| Kontaktuppgiftsfråga (organisation) | `com.nordicpeak.flowengine.queries.organizationdetailquery.OrganizationDetailQueryProviderModule` | Kan sparas under Mina organisationer |
| Kontaktuppgiftsfråga (privatperson) | `com.nordicpeak.flowengine.queries.contactdetailquery.ContactDetailQueryProviderModule` | Fält väljs i inställningar. Förifylls från Navet om kopplat |
| Kryssrutefråga | `com.nordicpeak.flowengine.queries.checkboxquery.CheckboxQueryProviderModule` | Egna alternativ |
| Kryssrutefråga med alternativ från lista | `com.nordicpeak.flowengine.queries.alternativelistcheckboxquery.AlternativeListCheckboxQueryProviderModule` | Central intern eller extern lista |
| Köfråga | `com.nordicpeak.flowengine.queue.query.QueueQueryProviderModule` | Aktiv kö |
| Personuppgiftshantering | `com.nordicpeak.flowengine.persondatasavinginformer.query.PersonDataInformerQueryProviderModule` | **Kräver tillägget Personuppgiftshantering på tjänsten**, annars fungerar inte e-tjänsten |
| Radioknappfråga | `com.nordicpeak.flowengine.queries.radiobuttonquery.RadioButtonQueryProviderModule` | Ett val |
| Radioknappfråga med alternativ från lista | `com.nordicpeak.flowengine.queries.alternativelistradiobuttonquery.AlternativeListRadioButtonQueryProviderModule` | Central lista |
| Rullistafråga | `com.nordicpeak.flowengine.queries.dropdownquery.DropDownQueryProviderModule` | Ett val |
| Rullistafråga med alternativ från lista | `com.nordicpeak.flowengine.queries.alternativelistdropdownquery.AlternativeListDropDownQueryProviderModule` | Central lista |
| Stoppfråga | `com.nordicpeak.flowengine.queries.stopquery.StopQueryProviderModule` | Blockerar nästa steg och inskick när den är synlig |
| Tabellfråga (dynamisk) | `com.nordicpeak.flowengine.queries.dynamictablequery.DynamicTableQueryProviderModule` | Sökanden lägger till rader. Kolumner, obligatoriskhet, summering och kolumntyp |
| Tabellfråga (statisk) | `com.nordicpeak.flowengine.queries.fixedtablequery.FixedTableQueryProviderModule` | Fast antal rader och kolumner |
| Textareafråga | `com.nordicpeak.flowengine.queries.textareaquery.TextAreaQueryProviderModule` | Flera rader |
| Textfältsfråga | `com.nordicpeak.flowengine.queries.textfieldquery.TextFieldQueryProviderModule` | Ett eller flera fält, som skapas under frågans inställningar |
| Val av barn | `com.nordicpeak.flowengine.queries.childquery.ChildQueryProviderModule` | Navet. Krav på val när frågan är obligatorisk |

<!-- Inställningar per frågetyp fylls på nedan när de har verifierats. -->

## Gemensamma inställningar (alla frågor)
Namn, kommentar (bara byggaren), beskrivning och hjälptext (CKEditor, hjälptexten öppnas via ?-ikon), standardläge, **Gruppera med föregående fråga**, **Exportera frågan** + **XML-elementnamn** (obligatoriskt när export är ikryssat), dölj rubrik, lås vid överlåtelse, lås vid ändring av handläggare. Många typer har **Spara som attribut**.

## Verifierade inställningar per typ

| Typ | Konfig-alias | Viktiga fält (formulärnamn) | Verifierat beteende |
|---|---|---|---|
| Textfält | `/textfieldprovider/config/<id>`, fält via `/addtextfield/<id>` | Per fält: `label`, `required`, `disabled`, `searchable`, `containsPrice`, `defaultValue`, `maskFieldContent`, `inactivateAutocomplete`, `formatValidator`, `invalidFormatMessage` (**krävs med validerare**), `placeholderText`, `minContentLength`, `maxContentLength`, `setAsAttribute`/`attributeName`. Fråga: `layout` = `FLOAT` (två kolumner), `NEW_LINE`, `NEW_LINE_FULL_WIDTH` | Min-längd: "Innehållet i det här fältet är 1 tecken men måste vara minst 2 tecken!". Personnummer kontrollerar **Luhn**. Positivt heltal avvisar 0. Maskering renderas som `type=password` och döljer placeholder |
| Textarea | `/textarea/config/<id>` | `maxLength`, `rows`, `showLetterCount`, validator, `searchable`, attribut, `keepalive` | Räknare "0/500" |
| Radio | `/radiobuttonquery/config/<id>` | Alternativ (UUID-rader, XML-värde, attributvärde, kostnad), fritextalternativ, validator, `columns` ONE/TWO/THREE, attribut, `lockForOwnerUpdate` | Regelkälla |
| Kryssruta | `/checkboxquery/config/<id>` | Som radio + `minChecked`, `maxChecked`, `showCheckAllBoxes` | Texten "Du måste välja minst 1 och får max välja 2" visas. Valfri fråga med min 1 och inget valt **godkänns** |
| Rullista | `/dropdown/config/<id>` | Kort beskrivning (förvald text), alternativ, fritext, `searchable`, attribut | Regelkälla. Attribut används för tilldelning |
| Information | `/lab-info/config/<id>` | Beskrivning (HTML), dölj bakgrund, dölj i förhandsgranskning/PDF, betalstöd, beskrivning som attribut | Inte fokuserbar, korrekt |
| Stopp | `/lab-stop/config/<id>` | Beskrivning, `displayValidationError` (röd ram) | Synlig stopp tar bort "Nästa steg" och stoppar steget |
| Datum/Tid | `/lab-datetime/config/<id>` | `useDateFields`, `useMultiDate` (från-till), min/max datum, `startDateRestrictionType` NONE/RESTRICT_FUTURE_DATES/RESTRICT_PAST_DATES + `restrictedStartDays`, `allowedDayOfWeek` MONDAY…SUNDAY, klockslags- och månadsspärrar, `restrictDateInterval` + MIN/MAX + DAYS/MONTHS, tidsfält, attribut | Native `type=date`. Fel: "Du måste välja ett från-datum som ligger på någon av följande veckodagar…", "Datumintervallet är för stort! Det får som mest vara 30 dagar…" |
| Filuppladdning | `/fileuploadquery/config/<id>` | `allowedFileExtensions` (en per rad), `maxFileCount`, `maxFileSize` (MB, standard 50), `maxQuerySize`, `maxFileNameLength`, PDF-inbäddning, filnamnsprefix, knapptext, attribut med filnamn | Fel filtyp avvisas i klienten med `alert("Filen du valde har inte ett tillåtet filformat")`. Fil hämtas via `/…/iquery/<flowID>/<instanceID>/q/<queryID>?file=<id>` |
| Dynamisk tabell | `/lab-dynamictable/config/<id>`, `/addcolumn/<id>` | `maxRows`, `decimalsToShow`. Kolumn: `label`, `required`, `summarise`, `width`, `columnType` (STRING, CHECKBOX, DROPDOWN, DROPDOWN_ALTERNATIVE_LIST, TEXTAREA, ROWSUM_ADD, ROWSUM_MULTIPLY, LIST_SEARCH), min/max, validator, `listName` | Summering live. Startar med en tom rad. **Kolumnkrav kontrolleras bara om frågan är obligatorisk.** En valfri tabell sparar tomma och halvfyllda rader. Mobil: kort per rad |
| Statisk tabell | `/lab-fixedtable/config/<id>`, `/addcolumn`, `/addrow` | Kolumner som ovan, rader med etikett | Summering. Tomma rader visas som "-" |

## Interna listor
Kolumntypen DROPDOWN_ALTERNATIVE_LIST och "…med alternativ från lista"-frågorna listar `InternalAlternativeListProviderModule`-listor (Landguiden, EU-länder, Familjehem, Förskoleklasser, Kommunala förskolor, Högstadieskolor, RTÖG, Sveriges vanligaste språk). [O] Menyn **Interna listor** (`/lab-lists`) visade ändå "Inga listor hittades". [H] Annan modulinstans. Fråga till Codex.
