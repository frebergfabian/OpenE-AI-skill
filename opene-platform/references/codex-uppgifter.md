# Frågor och uppgifter till Codex (systemsidan)

Från Claudes session 2026-11-12. Ingenting här har ändrats av Claude på servernivå.

| # | Observation | Förslag | Prioritet |
|---|---|---|---|
| 1 | Självsignerat TLS-certifikat med CN = IP. Webbläsare måste lita på det manuellt | Lokal CA eller ett `.test`-värdnamn (t.ex. `exempelkoping.test`) med certifikat för namnet. Det gör också att agenter får logga in med testkonton | Hög |
| 2 | Kortadresser visas som `https://<sandbox-ip>/<alias>` **utan port 8443** | Rätta portalens bas-URL eller SiteProfile | Medel |
| 3 | Statusgruppen **Standard** är tom, och "Lägg till standardstatusar" ger då inga statusar | Fyll gruppen (Sparat, Inskickat, Under behandling, Arkiverat + komplettering) eller döp om den | Medel |
| 4 | `/lab-lists` visar "Inga listor hittades" medan byggaren listar 8 interna listor (`InternalAlternativeListProviderModule`) | Kontrollera vilken modulinstans och sektion listorna hör till | Låg |
| 5 | JS-fel på varje portalsida: `init-modernizr.js:13` läser `$.browser.msie` (borttaget i jQuery ≥ 1.9) | Ta bort eller skydda i temats JS | Låg |
| 6 | Testwidget "test 123 Om / Mest använda" syns i sidofältet på alla adminsidor (`/administration/widgetadmin`, id 12) | Ta bort eller begränsa om den inte är avsiktlig | Låg |
| 7 | Hjälpikonen (?) överlappar frågerubriken vid 375 px | CSS i temat | Låg |
| 8 | Två knappar utan tillgängligt namn i formulärsteg. Obligatorisk textarea saknar `aria-required` | Tema eller XSL | Låg |
| 9 | SMTP-fångst: ärende 794 skickades in med tre sökandenotiser påslagna och adminkontots riktiga adress som mottagare innan notiserna stängdes av | Bekräfta att fångsten avvisade adressen och att inget lämnade servern. Gör fångsten till enda transport | **Hög, kontrollera** |
| 10 | Importen godtar regler med döda fråge- eller alternativreferenser utan varning | Upstream-ärende eller valideringsskript i Runnern | Medel |
| 11 | "Kräv meddelande vid statusbyte" (`addExternalMessage`) tvingades inte fram via `/flowinstanceadmin/status/<id>` | Verifiera i upstream om det är avsett | Låg |
| 12 | Separata rollkonton kräver mänsklig inloggning | Playwright-storageState per roll, skapad lokalt och lagrad utanför repot | Medel |

## Testdata som skapats av Claude (sandbox)
- E-tjänster: **Claude-labbet A** (familj 56: v1 flowID 20 publicerad, v2 flowID 22 inaktiv), **Claude-labbet C** (familj 57, flowID 21 publicerad och synlig i katalogen), importtester flowID 23–26 (inaktiva, inga ärenden).
- Ärenden: 794 (A, arkiverat), 795 (C, inskickat, tilldelat Sam Handläggning).
- Konto `claude`: tillagt i E-tjänsteutvecklare och E-tjänstepublicerare. Handläggare på familj 56 och 57.
- Ändrade notiser: sökandenotiser avstängda på flowID 20, 21 och 26.
