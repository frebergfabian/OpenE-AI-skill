# Vanliga fel och verifierade lösningar

| Symptom | Orsak | Lösning | Status |
|---|---|---|---|
| **Lägg till e-tjänst** har tom kategorilista, **Aktivera** utgråad | Kontot saknar e-tjänsttypens administrationsgrupper (admin-flaggan räcker inte) | Lägg kontot i typens grupper (`/flowadmin/flowtype/<id>` visar vilka) | Verifierad |
| Ny tjänst saknar statusar | Statusgruppen **Standard** är tom | **Ersätt statusar med statusgrupp** → Enkel eller en grupp med komplettering | Verifierad |
| "E-tjänsten går inte att publicera då dess statusar inte innehåller samtliga obligatoriska statusmappningar" | Saknad save- eller submit-mappning (t.ex. efter import) | Bocka åtgärden på rätt status (`/flowadmin/updatestatus/<id>`) | Verifierad |
| Tjänsten aktiverad men "Ej publicerad" | Publiceringsdatum saknas | Sätt publiceringsdatum i grundinformationen | Verifierad |
| Fråga/regel ger 403 eller "Det går inte att redigera frågorna…" | Versionen är publicerad eller har ärenden | Skapa en ny version (kopiera eller importera) | Verifierad |
| "Du måste fylla i fältet valideringsmeddelande!" | Validerare vald utan meddelande (gränssnittet förifyller via JS) | Fyll i `invalidFormatMessage` | Verifierad |
| "Du måste fylla i fältet XML-elementnamn!" | **Exportera frågan** ikryssad | Ange XML-elementnamn | Verifierad |
| Regel gör ingenting | Ny regel är **inaktiverad** som standard, eller tomma mål/alternativ efter import | Kryssa **Aktivera regeln**, kontrollera mål och alternativ | Verifierad |
| Viktregel ("…baserat på vikt") påverkar fel fråga | Den styr frågan den sitter på (har inga mål) | Skapa den på målfrågan | Verifierad |
| Ta bort-länk ger 403 | Länken kräver POST (`confirmHyperlinkPost`) | Klicka i gränssnittet, eller POST i automatisering | Verifierad |
| Dynamisk tabell sparar tomma eller ofullständiga rader | Kolumnkrav kontrolleras bara när frågan är obligatorisk | Gör tabellfrågan obligatorisk, eller acceptera att den är valfri | Verifierad |
| Personnummer avvisas trots rätt format | Luhn-kontroll | Använd syntetiska nummer med korrekt kontrollsiffra, t.ex. 19000101-0008 | Verifierad |
| Admin ser inga ärenden i Handläggning | Kontot är inte handläggare på familjen | **Välj handläggare** på tjänsten | Verifierad |
| Admintestet visar kvittens men inget ärende finns | **Testa e-tjänsten** sparar aldrig (ärendenummer 12345) | Publicera i sandbox och skicka in på riktigt | Verifierad |
| Text i meddelande eller notering blir `Ã¤` | Multipart skickat som UTF-8. Servern tolkar multipart som latin1 | Använd det riktiga formuläret, eller koda texten som latin1 | Verifierad (automatiseringsfel) |
| `fetch()` av sida visar `�` för å/ä/ö | Sidorna är ISO-8859-1 men `Response.text()` avkodar som UTF-8 | `new TextDecoder('windows-1252').decode(await r.arrayBuffer())` | Verifierad (mätfel, inte plattformsfel) |
| Sidan visar gammal status efter POST | Fliken visar cachad version | Läs om utan cache | Observerad |
| Fil "försvinner" efter uppladdning i automatisering | Widgeten skapar ett nytt `input[type=file]` per val. Att skriva över det gamla tar bort den väntande filen | Sätt filen på det **sista** tomma inputfältet | Verifierad |
| Typografiskt tankstreck ger SQL-fel i namn (Codex) | latin1-kolumner | Använd bindestreck i namnfält. I HTML-fält via CKEditor blir `–` till `&ndash;` | Codex-verifierad + delvis bekräftad |
| Sökning i användarens gruppfält hittar inget på "E-tjänste" | [H] Sökningen hanterar inte bindestreck eller å/ä/ö | Sök på ett ASCII-delord ("utveckl") | Observerad |
