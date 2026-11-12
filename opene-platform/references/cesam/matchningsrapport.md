# Matchningsrapport

Underlaget omfattar 114 XML-exporter med 114 olika filhashar samt 256 katalograder med 34 originalfält. Alla exporter hade roten `Flow`. JSONL och CSV innehöll 256 rader vardera. Fältkatalogen beskriver originalfälten. XLSX är en presentation av samma leverans. Inventeringen omfattar 239 steg, 3 506 frågor, 3 499 regler och 686 statusar. Frågor, regler och statusar räknas från respektive ägande nod eftersom exporter även kan innehålla inbäddade kopior.

| Koppling | Exporter | Bedömning |
|---|---:|---|
| Hög | 48 | 33 unikt normaliserat namn och stöd i kortadress, 12 unikt namn utan träff på kortadress, 3 entydig korsning av titel- och adresskandidater |
| Osäker | 56 | 51 med kortadress utan titelstöd och 5 med liknande titel utan adressstöd |
| Omatchad | 10 | Ingen kandidat med använda nycklar och tröskel |

Titlar jämfördes efter Unicode-normalisering, gemener och uppdelning på ord. Kortadress jämfördes via normaliserat sista sökvägssegment från katalogfältet och `Flow/FlowFamily/Aliases/Alias`. Exporterna har 112 familjer med kortadress och katalogen 234 ifyllda adresser. Fuzzy titelmatchning tog högst tre titelnycklar med teckenlikhet minst 0,78 som kandidater. En ensam kortadress kan vara omdöpt eller återanvänd. Den gav därför ingen hög konfidens utan titelstöd. Av de 51 adresskandidaterna hade 48 en ensam katalograd. För 31 av de 48 förekom den ena normaliserade titeln i den andra. Version eller familjeidentitet behöver fortfarande bekräftas.

De 48 högkonfidensmatchningarna pekar på 48 separata katalogposter. 208 katalogposter saknar högkonfidensmatch i leveransen, inklusive poster som kan finnas bland osäkra kandidater. Tre normaliserade titelnycklar och 13 kortadressnycklar förekommer på flera katalograder. Raderna bevarades separat. 96 av 256 har entydigt Cesam-upphov enligt källans normaliserade fält. Av 48 högkonfidensmatchningar har 24 den klassningen. Övriga 24 omfattar andra och blandade upphov. Urvalet på 24 styr Cesam-siffrorna och kan vara skevt jämfört med hela katalogen.

Lokalt finns separat exportinventering, kandidatlista, katalogposter utan säker koppling, risksignaler och strukturbelägg. Katalogens genererade post-ID har annan betydelse än OpenE:s flow-ID. Hög konfidens avser den sannolika katalogkopplingen. Versionslikhet och nuvarande produktionsläge kräver ytterligare underlag.
