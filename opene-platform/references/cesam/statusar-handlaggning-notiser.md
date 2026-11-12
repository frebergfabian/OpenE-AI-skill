# Statusar, handläggning och notifieringar

**Observerad Cesam-praxis.** I 24 matchade Cesam-exporter fanns `NEW` i 24, `SUBMITTED` och `ARCHIVED` i 23 vardera, `IN_PROGRESS` i 18 och `WAITING_FOR_COMPLETION` i 17. Sju hade `WAITING_FOR_MULTISIGN`. Alla hade åtgärdsmappningar. X001 visar kompletteringsstatus och X007 och X027 visar väntan på flerpartssignering. Belägg i `/Flow/Statuses/Status/contentType` och `/Flow/defaultFlowStateMappings`. Statusarnas typ och åtgärdsmappning är stabilare för jämförelse än de lokala status-ID:na.

- Utgå från processens övergångar. Kontrollera standardstatus för sparat och inskickat, tilldelning, komplettering, status efter återinskick och avslut. Testa att sökande och handläggare har rätt åtkomst vid varje övergång.
- För komplettering, verifiera vem som får ändra vad, hur begäran visas, hur bilaga lämnas och hur ärendet återgår till handläggning. Säkerställ att ett andra konto nekas även för bilagor och PDF.
- För flerpartssignering, definiera vilka som initierar och signerar, eventuell ordning, avbruten signering och återförsök. `requireSigning`, `skipPosterSigning` och väntestatus ska läsas tillsammans med faktiska signerare och vald provider.
- Katalogfält för notiser och ärendegrupper kan beskriva lokala inställningar utanför exporten. Separera händelse, kanal, mottagare, mall och utskicksbevis. En regel som skickar till en fältadress är en speciallösning, observerad i X070. Prova endast med lokal fångst eller avstängda utskick.
