# Integrationsstatus för den experimentella sandboxen

Testbevis och exakta förutsättningar finns i `tests/integration-cases-result.txt`. Denna lista skiljer på funktionella lokala tester och fullständig kontraktskompatibilitet.

| Integration | Status |
| --- | --- |
| Betalning | Egen adapter och lokal Checkout-simulator; lyckad och avbruten betalning med sparade ärenden och transaktionslogg är provade. Endast en delmängd av [Swedbank Pays kontrakt](https://developer.swedbankpay.com/checkout-v3/get-started/payment-request/) är implementerad. Återhämtning vid avbrott mitt i callback är inte verifierad. |
| Enkel-, flerparts- och dokumentsignering | DummySigningProvider och respektive OpenE-modul har provats med bestående data. En separat flerpartstjänst importeras och publiceras automatiskt. |
| PUD | Lokal söktjänst, syntetisk fastighet och sparat ärende är provade. LMProxy, LMSearch och WMS är inte verifierade end-to-end. |
| SSBT | De genererade SOAP-klienterna för företagsengagemang och grunduppgift UD0001 har använts mot lokala mottagare med ett sparat företagsärende. Ytterligare uppgifter och rollintyg är inte simulerade. |
| Barn och Navet | DummyChildRelationProvider och ChildInfo har provats med ett sparat ärende och syntetiskt barn. Ett separat Navet-kontrakt och PersonQueryProvider är inte verifierade. |
| SMS | Lokal SMS-mottagare och avsändarmodul har provats med sparat ärende. Mottagaren tillåter endast syntetiska testnummer. Övriga SMS-evaluatorer är inte individuellt verifierade. |
| HTTP-verksamhetssystem | HTTPIntegrationProvider har skickat ärende-XML till den lokala mottagaren och loggat svaret. |
| Artvise | Den genererade SOAP-klientens CreateCase-operation har provats mot en lokal mottagare med ett sparat ärende och syntetiskt externt ID. Övriga SOAP-operationer är inte implementerade. |
| Easit | Den genererade SOAP-klientens ImportItems-operation har provats mot en lokal mottagare med ett sparat barnärende och syntetiskt externt ID. Övriga SOAP-operationer är inte implementerade. |
| Evolution, Treserva, Castor, Cityworks, Boverket, Efecte, Sokigo, RemoteXML med flera | Leverantörsspecifika kontrakt och fullständiga ärendeflöden saknas eller är inte verifierade. Ett kompilerat tillägg är inte ett integrationsprov. |
| E-post | SMTP-fångsten har provats med sparade ärenden och tillåter endast angivna syntetiska domäner. |
| Bokning och kö | Syntetiska kalendrar och separata köer skapas automatiskt; bokning och köplats har verifierats på sparade ärenden. |

De generiska simulatorernas fel-, timeout- och nekningslägen har enhetstester, men de är inte alla verifierade genom varje OpenE-modul. Inga anrop ska riktas mot produktionssystem. Proxyinställningarna spärrar inte all rå nätverkstrafik.
