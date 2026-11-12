# Utvecklingsprinciper

**Observerad Cesam-praxis.** 18 av 24 flöden hade två steg. Tre hade ett steg, två hade tre och ett hade fem. Alla 24 hade lång beskrivning, ikon och taggar, och 22 hade sökord i `Checks`. X001 visar tvåstegsformen och X027 visar längre uppdelning. Belägg i `/Flow/Steps/Step/name`, `/Flow/longDescription`, `/Flow/Tags` och `/Flow/Checks`.

**Byggrecept.** Börja med ett kort startsyfte, en handlingsinriktad tjänstetitel, konkreta steg- och frågerubriker och hjälp vid första svåra beslutet. Ge tjänsten verksamhetsord som sökande faktiskt använder. Gruppera kontaktuppgifter, verksamhetsval och bilagor efter vad sökanden behöver göra. Dela upp stegen när villkor eller användarroller skiljer sig. Exempel på nyskriven text är **Ansök om en testplats** och **Välj vilken period ansökan gäller**.

**Observerad variation.** Inloggning fanns i 20 av 24, signering i 12, förhandsgranskning i 17 och enkät i 18. Spara-knappen var dold i 10. Konfigurera dessa val utifrån process, målgrupp och behov. Förhandsgranska hela flödet före publicering. Kontrollera anonym åtkomst, synlighet, ägarskap och eventuella tjänstefamiljeberoenden på den lokala versionen. Belägg i `/Flow/requireAuthentication`, `requireSigning`, `usePreview`, `showSubmitSurvey` och `hideSaveButton`. X001 och X007 visar olika kombinationer.

**Hypotes att utreda.** Ett tvåstegsflöde kan vara en praktisk startmall för enkla tjänster. Underlaget beskriver formens frekvens. Det mäter varken slutförandegrad eller nytta för sökanden.
