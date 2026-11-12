# Integrationer och listor

| Mönster | Observerad Cesam-praxis | Kontroll i ny miljö |
|---|---|---|
| Alternativlista | `listName` i 11 av 24. X001 och X007. | Bind lista på nytt, jämför ordning och stabila alternativvärden. Kontrollera villkor som läser alternativ-ID. |
| Företagsuppslag | 5 av 24 med företagsfråga. X025 och X036. | Syntetiska företag, ingen träff, timeout och manuell reservväg. |
| Fastighetsuppslag | 4 av 24 med PUD-fråga. X025 och X045. | Använd lokal fixture. Testa giltig träff, ingen träff och spärrad extern åtkomst. |
| Kalender | 4 av 24 med bokningsfråga och kalenderreferens. X011 och X062. | Lokal kalender, rätt handläggare, ledig och full tid, avbokning och persistering. |
| Kö | Två exporter i hela materialet. X064 och X084. | Lokal kö, dublettbeteende, köposition och ägargräns. Underlaget räcker för en speciallösning. |

Belägg finns i `QueryDescriptor/queryTypeID` och typkonfigurationens `listName`, `Calendar` eller `Queue`. Metadatafälten för listor, leveranser och notiser används som ledtråd. De visar ingen fungerande anslutning på egen hand. Behörigheter, listinnehåll och endpoints kan ligga utanför exporten.

Vid portering, klassificera varje beroende som lokal konfiguration, syntetisk fixture eller extern koppling. Välj avstängda utskick eller lokal fångst. Återanvänd endast ett externt kontrakt när ägare, behörighet och testgräns är kända. Import kan lyckas trots saknad anslutning eller trots att bokningens submit-händelse är obunden. Kontrollera effekten i både ärende och invånarvy.
