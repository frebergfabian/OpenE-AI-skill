# Bygga och prova i webbläsaren

## Import och export

1. Logga in med en roll som får bygga e-tjänster. Välj **Adm. e-tjänster → Importera e-tjänst**, välj e-tjänsttyp och underkategori, bifoga filen och tryck **Importera**. Det är en vanlig filuppladdning i webbläsaren.
2. Kontrollera att byggaren öppnar en ny tjänstfamilj som inaktiverad och att steg, frågor och regler finns kvar. **Importera ny version** hör till en befintlig familj.
3. I versionslistan finns exportåtgärden vid versionsraden. Filen har XML-roten `Flow` och servern kan svara med `text/oeflow`. Kontrollera innehåll, version och frågetyper. En rundtur ska bevara semantik och struktur. Interna nycklar kan ändras.
4. Kör **Förhandsgranska frågor** för alla steg och **Testa e-tjänsten** för ett administrativt formulärtest. För test av persistering publicerar du i sandbox, skickar in som en syntetisk invånare och följer upp på **Mina sidor** och **Handläggning**.

Granska källfilen lokalt före import. Leta efter externa URL:er, postmottagare, formgivning med aktivt innehåll, externa integrationer och inbäddade PDF-blanketter. Håll flödet avstängt under arbetet. Blockera XML-DTD och entitetsdeklarationer i verktyg för förbehandling. Använd en XML-parser och en HTML-sanerare för respektive format. Koppla bort externa länkar, stäng av betalning, ersätt mottagare med labbadresser och ta bort aktiva PDF-åtgärder och inbäddningar. En sanering kräver fortfarande granskning före vidare distribution.

Vid inkompatibel export, jämför `queryTypeID`, `evaluatorTypeID` och formatvaliderare med de laddade och tillåtna typerna för vald e-tjänsttyp. En äldre beräknad viktregel använder paketdelen `calculatedweightquerystateevaluator` där den testade modulen använder `weightcalculatedquerystateevaluator`. Äldre `ChildInfoQuery` utan `lookupMode` behöver det aktuella tilläggets standard `POSTER` när avsikten är uppslag för ärendets sökande. Namngivna alternativa listor och deras käll-ID:n kan behöva byggas upp i målmiljön. Granska status-ID:n, handläggargrupper och eventuella referenser till andra flöden. Bekräfta att den importerade familjens tilldelning, notifiering och tillägg är korrekt.

## Skapa tjänst

- **Lägg till e-tjänst** skapar en intern tjänst med e-tjänsttyp, kategori, namn, åtkomst, beskrivningar och statusgrupp. Behåll den inaktiverad medan du konfigurerar den. Behov av signering, meddelanden, läskvitton, enkät och publiceringsdatum anges på tjänstnivå.
- Under **Frågor och steg** skapar du namngivna steg och lägger till frågor. Standardläget kan vara synlig, obligatorisk eller dold. Textfältsfråga kräver dessutom ett fält. En flervalsfråga kräver egna alternativ eller en ansluten lista. En tabell kräver kolumner. Regler kopplas till källfrågan och väljer ett eller flera mål.
- Lägg till handläggargrupper i familjens handläggarvy, spara notifieringar och kontrollera statusarnas åtkomst. En bokningsfråga behöver en aktiv kalender med bokningsbara tillfällen. En köfråga behöver en aktiv kö. Tilldela kalenderns eller köns egna handläggare utöver modulbehörighet.
- Aktivera och publicera efter förhandsgranskning. Testa en utomstående roll, en sökande och en handläggare. Använd webbläsarens nätverkslogg och serverns logg när ett formulär ger ett allmänt fel.

Äldre XSL och databastabeller använder delvis ISO-8859-1 och `latin1_swedish_ci`. Ett typografiskt tankstreck i namnfältet gav SQL-fel i den verifierade installationen. Vanligt bindestreck och svenska bokstäver fungerade. En generell Unicode-migrering kräver samordning av formulär, servlet, JDBC och databaskolumner. Rich text är en separat väg och kan serialisera tecken som HTML-entiteter.
