# Formulär och frågetyper

**Observerad Cesam-praxis.** Av 24 högkonfidensmatchade Cesam-exporter innehöll 24 informationstyp och radioalternativ, 22 textfält, 21 kryssrutor, 21 privat kontaktfråga och 16 textarea. 15 hade en fråga för information om personuppgiftshantering och 15 hade stoppfråga. 13 hade filuppladdning, nio dynamisk tabell och åtta barnval. 12 hade ett `helpText`-fält. X001 och X007 visar kärntyperna, X027 och X036 tabeller och X007 barnval. Belägg i `/Flow/Steps/Step/QueryDescriptors/QueryDescriptor/queryTypeID` och frågans specifika konfiguration.

- Använd radio för ett val och kryssrutor för oberoende val. Förklara konsekvensen i frågan när valet leder till en annan gren. Lägg en stoppfråga där sökanden behöver en begriplig förklaring och väg vidare.
- Skilj informationsfrågan från själva uppgiftsinsamlingen. För dataskyddsfrågan, verifiera att tjänstens personuppgiftstillägg är konfigurerat. Kontrollera vilka kontaktfält som kan förifyllas och vilka som måste vara redigerbara.
- Använd dynamisk tabell för upprepbara rader och fast tabell när raderna är förutbestämda. Testa tom tabell, en och flera rader, borttagning, radsummor om de används och återupptagning.
- Sätt filtyp, antal och storlek utifrån verksamhetskrav. Testa både giltig och ogiltig fil, utkast och nedladdning med rätt roll. 49 av samtliga 114 exporter hade inbäddad blankett. En blankett kan innehålla aktivt innehåll och produktionsuppgifter.

**Labbstöd.** Två valgrenar, kravvalidering och tillägg av dynamisk tabellrad ingår i [verifieringsmatrisen](verifieringsmatris.md). Koppling mellan en tabellcell och produktionssystem behöver egen verifiering.
