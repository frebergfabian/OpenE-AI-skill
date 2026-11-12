# Återanvändbara mönster

Exemplen är syntetiska byggrecept. Börja med [Provansökan](../../assets/provansokan.oeflow) för ett litet importprov och [Funktionslabbet](../../assets/funktionslabbet.oeflow) för villkor, tabell och lokala modulberoenden.

| Byggrecept | Koppla ihop | Evidens och gräns |
|---|---|---|
| Enkel ansökan | Inledning, en obligatorisk uppgift, kvittens och statusgrupp | X001 visar grundstruktur. Syntetiskt inskick, persistering och behörighet verifierade. |
| Villkorad kompletterande bilaga | Radio Ja/Nej, dold bilaga som blir obligatorisk på Ja, test-PDF | X007 visar frågetillstånd och X001 filfråga. Kombinationen är syntetisk. Båda grenarna provade. |
| Upprepade rader | Dynamisk tabell med text och kryssruta, validering per kolumn | X027 och X036 visar dynamisk tabell. Tillägg av rad provat. Testa summering separat när sådan används. |
| Flerpart | Syntetisk sökande och andra part, signeringsstatus, avslag och återförsök | X027 och X059 visar manuellt signeringsval. Kräver separat test av två parter. |
| Boka eller köa | Kalender med framtida kostnadsfria tider, alternativt en lokal kö, inskick och Mina sidor | X011 visar bokning, X064 kö som speciallösning i hela materialet. Tidigare syntetiska bokningar och köplatser återkontrollerades. |

För en ny tjänst, skriv ner vilken del av mönstret som har stöd i Cesam-underlaget, vilket beslut verksamheten behöver fatta och vad som måste visas i test. Återanvänd form och kontrollpunkter. Skapa egna lokala mottagare, roller, listor och testvärden.
