# Funktionslabbet

Det här scenariot kombinerar formulär, bokning, kö, signering och handläggning i en syntetisk ansökan. Använd [det portabla utkastet](../assets/funktionslabbet.oeflow) om frågetyperna är installerade, eller bygg det i gränssnittet. [Provansökan](../assets/provansokan.oeflow) har bara en enkel fråga och passar ett första importprov.

## Byggrecept

| Steg | Frågor och regler |
|---|---|
| Ansökan | Sökbart obligatoriskt textfält med längdgräns, informationstext, obligatorisk textarea med teckenräknare, kontaktfråga med syntetisk e-post |
| Bilagor och val | Kryssrutor, radioknappar, rullista, datum och PDF-uppladdning med filstorleksgräns. En regel på radioknappens Ja gör bilagefrågan synlig och obligatorisk. Nej återgår till dolt standardläge. |
| Bokning och kö | Dynamisk deltagartabell med textfält och kryssruta, en bokningsfråga och en separat köfråga |

Konfigurera lokal inloggning, lokal signering, signeringsunderlag i PDF, inskickskvittens, läskvitton, återkopplingsenkät, notifieringar och handläggargrupper. Lägg till statusgruppen **Enkel** via **Ersätt statusar med statusgrupp** om byggaren saknar statusar. Kontrollera obligatoriska åtgärdsmappningar innan publicering.

Den portabla exporten är ett inaktiverat utkast. Koppla bokningsfrågan och köfrågan till en lokalt skapad kalender respektive kö. Underliggande numeriska referenser, publiceringsdatum och kontaktadress har tagits bort. Lägg upp kommande kostnadsfria tider och tilldela testhandläggare till kalendern och kön. Ange en mottagare inom SMTP-fångstens syntetiska domän. Testkontona behöver varsitt icke-verkligt `citizenIdentifier`-attribut. Det ska aldrig användas för verklig identifiering.

## Verifiering

1. Importera Provansökan och det portabla utkastet med webbläsarens filväljare. Kontrollera ny familj, inaktiverad version, antal steg och frågor samt bevarad regel. Använd en lokal kopia om du även provar en verklig referenstjänst.
2. Publicera den färdigkonfigurerade syntetiska tjänsten. Som sökande, prova att gå vidare med ett tomt obligatoriskt fält. Fyll i, spara, öppna **Mina sidor** och fortsätt samma ärende. Välj Ja, ladda upp en syntetisk PDF och verifiera att Nej döljer frågan igen.
3. Lägg till en tabellrad, boka en ledig tid, anslut till kön och slutför testsigneringen. Öppna kvittensen, hämta PDF och svara på enkäten direkt i kvittensvyn. Kontrollera att bokningen och köplatsen visas för ägaren. Kontrollera också att respektive rad är kopplad till det inskickade ärendet.
4. Som handläggare eller chef, öppna ärendet, byt status och skicka ett meddelande. Läs status och meddelandet som sökande. Kontrollera att ett annat sökandekonto får avslag. Läs de lokalt fångade e-postmeddelandena och kontrollera XML-exporten.
5. Exportera via versionsraden. Chromium ska spara en giltig `.oeflow` med svenska tecken och rätt ändelse. Äldre kod med enbart `filename` i Content-Disposition kunde tappa delar av namnet. Open ePlatforms export kan använda den befintliga hjälparen `FileUtils.toContentDispositionFilename(...)` för UTF-8-parametern `filename*`.

## Fallgropar från provkörningen

- På kalender- och kömodulerna måste `supportedActionIDs` uttryckligen innehålla `com.nordicpeak.flowengine.FlowBrowserModule.submit` i den testade uppsättningen. En lyckad kvittens kunde annars visas trots att bokningen och köplatsen aldrig skapades. Kontrollera både databasen och invånarvyerna.
- Sätt `staticContentPackage=staticcontent` för modulklasser som faktiskt har den resurskatalogen på klassvägen. Använd modulens befintliga `StaticContentLinks.properties` för delade resurser. Ett HTTP 200 för modulsidan kan annars dölja 404 för dess JS och CSS.
- Vanligt bindestreck fungerade i namn där ett typografiskt tankstreck gav fel mot en äldre latin1-kolumn.
- Tillgång till en kalender via modulens huvudmeny räcker inte för att lägga till bokningstillfällen. Ett konto eller en grupp behöver även vara kalenderhandläggare. Samma typ av avgränsning finns för köer.

Lokalt verifierades import i Chromium, skapande från grunden, 3 steg och 12 frågor, obligatorisk validering, sparat utkast och återupptagning, båda grenarna i bilageregeln, filuppladdning, testsignering, kvittens, PDF, enkät, bokning, kö, status, meddelande, ägaravgränsning och lokala notifieringar. Tio sanerade referenstjänster verifierades fram till formulärets första steg efter publicering. De behöver fortfarande egna kompletta scenariotester.

Fristående dokumentsignering verifierades separat. En handläggarchef lade upp en syntetisk PDF, valde ett internt testkonto och skickade uppdraget. Sökanden hittade det i **Mina dokument för signering**, signerade med dummyprovidern och fick en signerad post i invånarvyn. Systemet registrerade dokumentet som signerat.
