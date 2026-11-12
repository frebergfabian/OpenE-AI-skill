# Nordic Peak och lokala integrationer

Den verifierade källan var Nordic Peaks officiella AGPL-arkiv med SHA-256 `46a0739855341300dc5478156e92beb8a991461929e39cc77b35313be2fbe157`. Arkivets revisionsfil anger `oep-dev` revision 2199. Se den [fullständiga inventeringen](nordic-peak-inventering.tsv) för skillnaden mellan projekt i arkivet och moduler som faktiskt laddats. [Nordic Peak](https://nordicpeak.com/open-eplatform)

## Bygg och aktivering

- Bygg mot plattformens staged runtime-JAR:er med JDK 21, `--release 8` och källkodning ISO-8859-1. Kopiera även modulernas egna XSL, JS och övriga resurser till JAR, utan att slå ihop projektens manifest.
- PersonProvider och PaymentProvider är officiella OpenHierarchy-gränssnitt från SVN r3190. De gör att lokala boknings- och köklasser kan länkas. En laddad betalningsklass eller gränssnitts-JAR utför ingen betalning på egen hand.
- För MultiSigningHandler2 och DocumentSignerAdminModule skiljer sig StandardUtils namn för hashalgoritmer. Mappa till det installerade `HashAlgorithm` med samma värde, exempelvis `HashAlgorithm.SHA256.toString()` för dokument. Bevara befintliga säkerhetskontroller.
- Registrera frågans administrations-XSL som klasspath-resurs och inställningarna `queryStyleSheet` och `pdfStyleSheet` för formuläret och PDF. Lägg till frågetypen i e-tjänsttypens tillåtna lista.
- Bokning behöver CalendarAdminModule, BookingSlotAdminModule, UserBookingModule, BookingQueryProviderModule och BookingNotificationHandler. Koppla notifieringsmodulens interna URL-alias till invånarsidan och handläggarsidan. Kö använder QueueAdminModule, QueueSlotAdminModule, UserQueueSlotModule och QueueQueryProviderModule.
- Dokumentsignering använder DocumentSignerAdminModule och DocumentSignerUserModule. Adminmodulen behöver skrivbar filkatalog, tempkatalog och en SiteProfile som matchar den aktiva portalens värdnamn utan port. Invånarmodulen behöver en GenericSigningProvider. Undvik att exponera filarkivet via webbservern.
- Registrera TargetGroupAdminModule separat. Den öppnar målgruppsadministration och ett tillägg i byggaren. En alternativ FlowBrowser kan konkurrera med den aktiva katalogmodulen. Sandboxens katalogfilter mappar livshändelser till kategorier i portalens JS och XSL.
- Vissa projekt kan kompileras utan att de är säkra att aktivera. Registrera aldrig öppna REST- och SOAP-endpoints bara för att JAR-filen finns. Kontrollera kontrakt, autentisering och avgränsning först.

FlowInstanceAdmin2 och FlowInstanceListModule byggde inte direkt mot kombinationen. De förutsätter andra status- eller meddelandesignaturer. FlowVoteModule använder en annan variant av ModuleUtils och notifierings-API:t. ReservationModule har en inkompatibel ReservationProvider-signatur. Byt till en matchande officiell version, eller porta och testa kontrakten separat. I synnerhet bör handläggnings- och statusbehörigheter aldrig ändras bara för att få kompilering att lyckas.

## Testkopplingar

| Funktion | Lokal koppling och gräns |
|---|---|
| Signering | Open ePlatforms DummySigningProvider för syntetisk identitet. Enkelsignering kan verifieras hela vägen till kvittens. Flerpartssignering är en separat testkedja. |
| E-post | DirectMailSender mot en lokal SMTP-fångst som avvisar mottagare utanför avsedd syntetisk domän. Fångsten sparar EML-filer och vidarebefordrar inget. |
| Fastighet och adress | PUD-frågan anropar lokal HTTP-fixtur. Matchande uppslag ger syntetiskt `objid` och `name` i `properties` via `/search/registerenheter` och `/search/addresses`. |
| Företag | Testgrenen `opene.lab.ssbtFixture=true` returnerar ett syntetiskt företag och skapar inga SSBT-proxyer. |
| Barn och listor | DummyChildRelationProvider och centrala interna alternativlistor med påhittade värden. Ursprungliga list-ID:n och data kan saknas i exportfilen. |
| Prenumerationer | Lokal användarsida kan laddas. Producenten förväntas leverera XML med `/Subscriptions/Subscription`. En aktiv leverantör kräver ett verifierat kontrakt. |
| Betalning, eIDAS, SAML, SMS, verksamhetssystem | Produktionsanslutningar kräver leverantör, behörighet och rätt konfiguration. Kontoplaner och transaktionsvyn kan startas utan att pengar överförs. |

De publika referensportalerna visar katalog, livshändelser, kategorier, sökning, kort/lista, Mina sidor och i vissa fall synliga bokningstjänster. De ger ingen verifierbar bild av intern handläggningskonfiguration eller externa integrationer. [Finspång](https://minasidor.finspang.se/) · [Norrköping](https://minasidor.norrkoping.se/) · [Västervik](https://minasidor.vastervik.se/) · [Itsam](https://minasidor.itsam.se/)
