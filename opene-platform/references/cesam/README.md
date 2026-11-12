# Cesams kunskapsbank

Använd kunskapsbanken när du bygger eller granskar en e-tjänst för Cesam. Börja med verksamhetens faktiska krav. Leta sedan efter ett jämförbart mönster och följ dess evidens och testgräns.

| Sök efter | Läs |
|---|---|
| Mall, namn, rubriker, publicering, granskning | [Utvecklingsprinciper](utvecklingsprinciper.md) |
| Kontaktfråga, flerval, tabell, bilaga, dataskydd | [Formulär och frågetyper](formular-och-fragetyper.md) |
| Villkor, vikt, eIDAS, attribut, validering, XML | [Regler och attribut](regler-attribut-och-validering.md) |
| Status, komplettering, signering, tilldelning, notifiering | [Status och handläggning](statusar-handlaggning-notiser.md) |
| Lista, uppslag, bokning, kö, mottagare | [Integrationer och listor](integrationer-och-listor.md) |
| Syntetiska byggrecept | [Återanvändbara mönster](ateranvandbara-monster.md) |
| Andra upphov, avvikelser, äldre konfiguration | [Avvikelser och teknisk skuld](avvikelser-och-teknisk-skuld.md) |
| Underlag, urval, osäkerhet | [Matchningsrapport](matchningsrapport.md) och [sanitets- och riskrapport](sanitets-och-riskrapport.md) |
| Labbutfall, prioriterad Runner-kö | [Verifieringsmatris](verifieringsmatris.md) |

## Läs evidensnivån

- **Dokumenterad standard** kräver ett normerande Cesam-underlag med ägare och giltighet. Sådant ingick inte i leveransen.
- **Observerad Cesam-praxis** avser återkommande struktur i de 24 högkonfidensmatchade exporter med entydigt Cesam-upphov enligt katalogen.
- **Speciallösning** avser en avgränsad teknisk lösning. Pröva verksamhetsbehov och beroenden innan återanvändning.
- **Annan variant** avser tydligt annat eller blandat upphov. Behåll det som jämförelse.
- **Misstänkt teknisk skuld** avser en skillnad eller kompatibilitetsrisk som bör utredas.
- **Hypotes** behöver mer evidens. **Verifierat i labb** avser endast det angivna syntetiska utfallet och kan gälla flera av kategorierna ovan.

`Xnnn` är ett anonymt beläggs-ID i just denna leverans. Det motsvarar exportens ordning vid sortering på filnamn i den lokala inventeringen. Den privata nyckeln innehåller filnamn, SHA-256 och eventuellt katalogpost-ID. XML-sökvägar i referenserna visar vilket skikt som stöder observationen. Exempeltexter är nyskrivna och underlaget återges genom aggregat och struktur.
