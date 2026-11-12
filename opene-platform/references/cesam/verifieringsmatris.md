# Verifieringsmatris

Chromium användes för de lokala körningarna. Endast syntetiska testuppgifter och mallar användes. Testutkasten ligger kvar inaktiverade. Skärmbilder och körloggar finns i den privata leveransen.

| Mönster | Prov | Utfall |
|---|---|---|
| Enkel ansökan | Import i filväljaren, export från byggaren och jämförelse av struktur. Ett steg och en fråga. Inskick i administrativt testläge. | Godkänt. Kvittens visades och utkastet fick inget persisterat ärende. |
| Villkor och dynamisk tabell | Import och återexport av tre steg, tolv frågor och en regel. Förhandsgranskning av utkast. Prov i den syntetiska testtjänsten med tom obligatorisk fråga, båda grenarna i bilageregeln och tillagd tabellrad. | Godkänt för import, renderad fråga, validering, synlighet och radtillägg. Persistering av alla tabellvärden, filgränser och summeringar behöver ett riktat test. |
| Inskick och handläggning | Sökande skickade ett nytt syntetiskt ärende. Kontroll i databas och Mina sidor. Ett annat konto fick avslag. En handläggarchef bytte status och sökanden såg den. | Godkänt för persisterat ärende, ägaravgränsning och statusbyte. Kompletteringskedjan provades ej. |
| Bokning och kö | Kontroll i invånarvyn och i tidigare sparade syntetiska ärendens boknings- och köposter. | Tidigare persistering återkontrollerad. Ny bokning, avbokning och dubbelbokningsgränser behöver egna körningar. |
| Flerpartssignering | Syntetisk import och återexport med manuell flerpartsfråga. Sökanden och den angivna medsigneraren signerade i separata sessioner med lokal DummySigningProvider. | Godkänt. Ärendet väntade efter första signaturen. Efter andra signaturen blev ärendet inskickat. Båda kunde öppna ärendet och hämta signerad PDF. |
| Externa integrationer | Register, identitetsleverantör, SMS, betalning och externa mottagare. | Ingen extern provkörning. Kräver lokala fixtures och separata kontraktstester. |

## Manuell flerpartssignering i sandlådan

Den första importen rapporterade saknad frågetyp. I den testade FlowEngine-versionen fanns `ManualMultiSignQueryProviderModule` redan i JAR-filen. Lösningen var att registrera modulen med administrations-, fråge- och PDF-stilmall och lägga till den bland e-tjänsttypens tillåtna frågor. Flerpartshanteraren och den lokala signeringsprovidern var redan installerade. Kontrollera både modulregistret och frågans tillåtelse när samma fel uppstår.

Det syntetiska flödet använde statusgruppen **Flerpartsignering**, sekventiell signering och två olika testkonton med giltiga testpersonnummer. Frågan nekade att sökanden angav sig själv som andra part. Efter sökandens signatur hade ärendet status `WAITING_FOR_MULTISIGN`. Ett konto som inte var part fick meddelandet **Åtgärden ej möjlig** på signerings- och PDF-vyerna. Ett simulerat signeringsfel för den andra parten lämnade sessionen oavslutad. När andra parten försökte igen och signerade registrerades två `SIGNED` med olika aktörer och därefter en `SUBMITTED`. Båda kunde öppna ärendet via Mina sidor och hämta signerad PDF. Den lokala SMTP-fångsten tog emot en signeringsinbjudan och två inskicksnotiser.

Testet omfattar lokal dummy-signering med en sökande och en medsignerare. BankID, fler än två parter och avbruten signeringssession har separata verifieringsbehov.

## Prioriterade testfall för OpenE Runner

| Prioritet | Test | Godkännandekriterium |
|---|---|---|
| 1 | Import som ny familj och ny version | Jämför frågetyp, reglers mål, åtgärdsmappningar och status. Markera saknade typer och katalogberoenden. |
| 1 | Flerpartssignering | Automatisera det verifierade tvåpartsfallet med väntestatus, fel, omförsök, åtkomst och sparade händelser. Lägg till avbrott och ordningsprov med flera medsignerare. |
| 1 | Regel- och attributkontrakt | Verifiera A → B → A, överlappande regler och viktgränser. Kontrollera sparade värden och XML för varje gren. |
| 1 | Komplettering och notifiering | Sökande, chef och handläggare har olika rättigheter. Skicka lokal testnotis, lämna bilaga, återinskicka och kontrollera att mottagare och status stämmer. |
| 2 | Tabell och fil | Testa flera rader, summering, återupptagning, ändring, filgränser och att ett annat konto nekas åtkomst till bilaga. |
| 2 | Bokning och kö | Skapa isolerade tider och köer per körning. Testa sista platsen, dublett, avbokning, ägande och registrerad händelse. |
| 2 | Lista, register och eIDAS | Provdata för tom lista, ändrat alternativ-ID, uteblivet svar och olika identitetsattribut. Kontrollera lokal reservväg. |
