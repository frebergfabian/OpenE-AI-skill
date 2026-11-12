# Sanitets- och riskrapport

Originalen lästes i en privat katalog utanför Git. XML-parsern avvisade DTD, entitetsdeklarationer och extern upplösning. Ingen HTML, blankett eller länk från originalen öppnades som aktivt innehåll. Metadata var redan delvis maskerad enligt leveransens integritetsrapport. Interna uppgifter finns kvar även där.

| Signal i exporter | Filer av 114 |
|---|---:|
| E-postliknande värde | 105 |
| Telefonliknande värde | 86 |
| URL eller protokollreferens | 83 |
| Gatuadressliknande text | 19 |
| HTML-signal som kan ladda eller köra innehåll | 5 |
| Inbäddad blankett | 49 |

69 blanketter ligger inbäddade totalt. Skanningen letade även efter personnummer med avskiljare, privat IP och nyckelord för hemligheter. Den fann inga utslag med just de mönstren. Sökningen registrerade konfigurationsfält för bland annat namn, användar-ID, kontaktvägar, notifiering, integration och lista för lokal granskning. Telefon- och adressmönster kan ge falska träffar. Personnamn, hemligheter i fria texter, PDF-innehåll och bildmetadata kräver manuell granskning.

**Före portering av en verklig export.** Arbeta i en kopia. Ta bort ikoner och bilagor eller bygg egna syntetiska varianter. Ersätt person- och kontaktvärden, maskera organisation, adresser, ID, mottagare, domäner och externa referenser. Granska rich text och borttagna aktiva element. Inventera notifieringar, API- och registeranrop, integrationsregler och listor och koppla bort dem före import. Kontrollera både filens innehåll och runtimekonfiguration. En lista över hittade fält är ett underlag för granskning.

**Gräns för PR:n.** Delbart innehåll omfattar aggregat, strukturella XML-hänvisningar och nyskrivna exempel. Originalnamn, titlar, hashvärden, interna katalog-ID, matchningsnycklar, lokala skärmbilder och råkonfiguration förvaras lokalt. Labbet använde syntetiska mallar.
