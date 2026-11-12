# Underlag för OpenE Runner (automatiserad testning)

Ingen ny testmotor byggdes. Det här är verifierade mekanismer och föreslagna testfall från sessionen 2026-11-12. Webbläsaren var Claude-appens inbyggda Chromium-panel (klick, tangentbord, JS, nätverkslogg, viewport-emulering). Playwright fanns inte.

## Mekanismer att bygga på
- **Inga CSRF-tokens.** Byggarens formulär kan köras som `application/x-www-form-urlencoded;charset=UTF-8`-POST. Lyckat = redirect till config/showflow. Misslyckat = samma formulär med felruta (`#lab-main [class*=error]`).
- **Multipart** (import, meddelanden, filer) tolkas som latin1. Skicka text som latin1-bytes.
- **Läs sidor som windows-1252.**
- **Regelutvärdering via JSON:** ändra källfrågan med POST till flödes-URL:en och läs `QueryModifications` (SHOW/HIDE per `queryID`). Det är snabbt och kräver ingen rendering.
- **Navigering i flöden:** knappen `a.btn.next` (Nästa steg, Förhandsgranska, Skicka in), `a.btn` "Föregående steg", Spara `a.btn.force-submit`. Stegtitel = `document.title` före " - ". Fel: `[class*=error]`, `aria-invalid="true"` och `aria-describedby="<fält>_error"`.
- **Fältnamn:** `q<queryID>_field<fieldID>` (textfält), `q<queryID>_value` (textarea), `q<queryID>_startDate`/`_endDate`, `q<queryID>cell_<radID>_<kolumnID>` (tabell, första raden `newRow`), `q<queryID>_newfile` (fil, ett nytt input per val).
- **Dialoger:** filtypsfel och vissa bekräftelser använder `alert()`/`confirm()`. Runnern måste fånga dialoger.
- **Ärendeverifiering:** kvittensens URL `/oversikt/submitted/<instanceID>`, sedan Mina sidor, `/flowinstanceadmin/overview/<id>`, PDF `/…/pdf/<id>/<eventID>` (kontrollera `%PDF-`) och XML.
- **Rollbyte** kräver separata webbläsarkontexter eller sessioner per testkonto. Lösenord hanteras av en människa eller en hemlighetshanterare, aldrig i repot.

## Föreslagna automatiska tester
1. **Byggarkontroll** per tjänst: alla frågor har konfig-URL, obligatoriska statusmappningar finns, ingen tom statusgrupp, alla regler aktiverade och med minst ett mål och ett alternativ.
2. **Importrundtur:** exportera, importera som ny, exportera igen och jämför normaliserad struktur (antal steg, frågor, regler, mappningar, regelreferenser).
3. **Felaktiga importer:** trunkerad, DTD, okänd frågetyp eller regeltyp ska avvisas. Regler med döda referenser ska flaggas av Runnern, eftersom plattformen inte gör det.
4. **Regelmatris:** för varje regel, sätt källvärden och kontrollera `QueryModifications`. Testa både på- och avslag.
5. **Validering:** tomt obligatoriskt fält, min/max-längd, varje validerare (inklusive Luhn), datumregler (veckodag, intervall, framtid), filtyp och filstorlek, tabellkrav för obligatorisk och valfri tabell.
6. **Stoppfråga:** synlig stopp ska ta bort Nästa och stoppa steget.
7. **Livscykel:** spara utkast → Mina sidor (Ej inskickade) → återuppta → skicka in → kvittens med riktigt nummer → PDF och XML → handläggare: tilldela, status, meddelande, notering → sökande: ser status och meddelande men inte noteringen → komplettering → automatisk Kompletterat → arkivering → Avslutade.
8. **Behörighet:** sökande 2 får avslag på sökande 1:s ärende, filer och PDF. Handläggare utanför familjen ser inget. Admin utan handläggarroll ser inget.
9. **Automatisk tilldelning:** attributvärde → förväntad handläggare. Värde utan träff → reservtilldelning.
10. **Notiser:** med syntetisk mottagare, kontrollera EML i SMTP-fångsten (kräver Codex-åtkomst).
11. **Visuellt:** skärmbild per steg i 375, 768 och 1280 px. Ingen horisontell scroll. Inga överlappande element (hjälpikonen vid 375 px).
12. **Tillgänglighet:** axe-core per steg. Tab-ordning, synligt fokus, `aria-invalid`/`aria-describedby` på fel, namn på alla knappar.

## Det som behövs för att bygga Runnern
- Playwright (Chromium) med `ignoreHTTPSErrors` endast mot sandboxen, eller betrott CA.
- En storageState per testroll, skapad av en människa eller via lokal testinloggning.
- Testdata: syntetiska personnummer med giltig Luhn-siffra, PDF-fixtur, syntetiska e-postdomäner.
- Rensning: testtjänster märkta "Claude-labbet" eller "Runner-" och inaktiva efter körning.
