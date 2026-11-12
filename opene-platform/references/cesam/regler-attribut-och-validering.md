# Regler, attribut och validering

**Observerad Cesam-praxis.** 22 av 24 har frågetillståndsregler, 20 viktning av alternativ, 19 frågetillstånd baserat på vikt, sju en beräknad viktregel och 14 eIDAS-relaterad villkorslogik. Alla 24 har minst en fråga dold i standardläget och 23 minst en obligatorisk fråga. Belägg i `QueryDescriptor/defaultQueryState` och `QueryDescriptor/EvaluatorDescriptors/EvaluatorDescriptor`. X007 visar tillstånd och eIDAS, X009 beräknad vikt, X025 en kombination av eIDAS och dynamisk tabell.

- För en villkorsfråga, välj startläge, källalternativ, mål och återställningsbeteende. Testa A → B → A, obligatoriskhet på rätt gren och vad som sparas när ett tidigare svar döljs.
- För en viktkedja, dokumentera källvärde, poängskala, beräkning och gränsintervall. Testa gränsvärden, saknat värde och flera samtidiga villkor. Ge regeln ett namn som beskriver effekten.
- Håll isär validering, ärendeattribut och XML. 19 av 24 har någon `formatValidator`, 20 har `attributeName` och åtta har `xsdElementName`. E-postvalidator fanns i tio och postnummervalidering i sju. Belägg i frågans typkonfiguration, exempelvis X001 för attribut och X007 för XML-namn.
- Välj stabila, verksamhetsbegripliga namn när ett attribut är kontrakt mot vy, notifiering eller integration. Verifiera fältets aktiveringsflagga, faktisk export och vilket värde som skapas. Närvaro av en XML-tag visar konfiguration. Ett mottagarsystem kan tolka samma tagg annorlunda.
- Ett lokalt testkonto med dummyinloggning verifierar frågans reservgren. För eIDAS-regler behövs en kontrollerad identitetsfixture och tester för respektive attributuppsättning.
