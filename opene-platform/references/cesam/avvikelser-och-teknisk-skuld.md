# Avvikelser och teknisk skuld

**Andra varianter.** De 24 högkonfidensmatchningarna utanför entydigt Cesam-upphov inkluderar tio med Norrköping som upphov och andra med enskilt eller blandat kommunupphov. 22 av 24 i jämförelsegruppen hade personuppgiftsfråga, jämfört med 15 av 24 i Cesam-urvalet. Nio av 24 dolde externa meddelanden, jämfört med tre av 24 i Cesam-urvalet. Grupperna är urval av exporter. Variationerna kan bero på tjänstebehov, moduler och versioner. Underlaget etablerar ingen kommunstandard.

**Speciallösningar.** Gruppstyrd frågesynlighet förekom i fem exporter i hela materialet och i noll i Cesam-urvalet, exempel X008 och X058. Vissa flöden kombinerar poäng, tabeller eller styckeinformation. Ett fåtal har meddelanderegler som läser mottagaren från ett fält. Behandla dem som separata beslut med test av rätt behörighet och rätt mottagare.

**Misstänkt teknisk skuld.** Signeringsflaggan i exporten och katalogfältet hade olika värde i nio av de 44 högkonfidensparen med jämförbara ja/nej-värden. Inloggningsflaggan stämde i 47 jämförbara par. Skillnaden kan bero på att katalogfältet även avser andra former av signering eller att versionerna skiljer sig. Bevara båda källvärdena och verifiera faktisk konfiguration i rätt version.

**Kompatibilitet.** Vissa äldre klassnamn behöver bekräftad mappning, exempelvis viktutvärderingens paketnamn. Kontrollera också katalogens tillåtna typer, lokala list-ID:n, förifyllning och statusmappning. I den kontrollerade XML-sökvägen för frågereglers `TargetQueryIDs` hittades inga olösta mål-ID:n. Det säger inget om externa referenser eller alternativen i en mottagande installation. Se [importguiden](../webblasarfloden.md).
