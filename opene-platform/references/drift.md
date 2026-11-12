# Drift, roller och felsökning

## Roller och ärenden

| Roll | Normal uppgift |
|---|---|
| Sökande | Katalog, egna utkast, inskick och egna ärenden, bokningar, köplatser och signeringsuppdrag |
| Handläggare | Tilldelade ärenden, ärendehändelser, meddelanden och de kalendrar eller köer där rollen har åtkomst |
| Handläggarchef | Fördelning, status och arkivering inom tilldelade familjer |
| E-tjänsteutvecklare | Formulärbyggare, regler, import, export och testläge |
| E-tjänsteadministratör | Behörighet att bygga och publicera i tillåtna kategorier |
| Redaktör | Sidor, driftinformation och redaktionellt innehåll enligt sektionens ACL |
| Systemadministratör | Moduler, sektioner, användare, grupper, profiler, datakällor och drift |

Testa åtkomst med minst två invånarkonton. Det andra kontot ska sakna åtkomst till den förstas ärenden och filer. Modultillträde, e-tjänsttyp, familjens handläggarregler, statusbehörighet och ärendets tilldelning kontrolleras separat. Logga in på nytt efter ändrat gruppmedlemskap. Ett HTTP 200 kan vara en tom lista eller en inloggningssida efter omdirigering.

## Konfiguration

OpenHierarchy laddar moduler från databasen. `openhierarchy_foreground_modules` innehåller klass, sektion, alias och flaggor. `openhierarchy_foreground_module_settings` har en rad per inställningsvärde, även i flervalsfält. `openhierarchy_foreground_module_groups` ger modulbehörighet. Bakgrunds- och filtermoduler har egna tabeller. Registrera först beroenden och kontrollera att klasser, XSL och resurser finns i webbappen. Verifiera appens faktiska logg efter omstart. Tomcat som startat innebär inte att OpenHierarchy laddat alla moduler.

I byggd webbrot ligger JDBC-inställningen i `WEB-INF/config.xml` och loggkonfigurationen i `WEB-INF/log4j.xml`. Håll dem och TLS-nycklar skyddade. Bind databasen och testadaptrarna till loopback. Vid egen XSL-överlagring använder katalogen den inbyggda FlowBrowser-listan och sökindexet. Testa omstartsberoende cache efter manuella SQL-ändringar och utför löpande ändringar i webbgränssnittet. En kataloglänk ersätter aldrig serverns behörighetskontroll.

| Tabellfamilj | Innehåll |
|---|---|
| `simple_users`, `simple_groups`, `simple_user_groups` | Lokala konton och roller |
| `flowengine_flow_types`, `flowengine_flow_categories` | E-tjänsttyp och kategori |
| `flowengine_flow_families`, `flowengine_flows`, `flowengine_steps`, `flowengine_query_descriptors` | Tjänstfamilj, version, steg och frågor |
| `flowengine_flow_type_allowed_queries` | Tillåtna frågetyper per e-tjänsttyp |
| `flowengine_flow_instances`, `flowengine_flow_instance_events`, `flowengine_external_messages` | Ärenden, händelser och meddelanden |
| `flowengine_flow_family_manager_groups`, `flowengine_flow_instance_managers` | Åtkomst till familjens ärenden och handläggartilldelning |
| `flowengine_standard_status_groups`, `flowengine_flow_statuses` | Statusmallar och tjänstens egna statusar |
| `booking_calendars`, `booking_slots`, `bookings`, `queues`, `queue_slots` | Bokning och kö |
| `table_versions` | Databasmigreringar per modul |

## Kontroll vid fel

- Importen anger saknad eller otillåten frågetyp eller regeltyp. Jämför fullständigt klassnamn och e-tjänsttypens tillåtna typer, inklusive validerare.
- Ett allmänt byggarfel kan vara ett teckenkodnings- eller databasfel. Läs appens stacktrace och det första underliggande SQL-felet. Prova om med syntetiska värden och dokumentera vilken encoding som fallerar.
- Vid utebliven notis, kontrollera händelsen, tjänstens notisinställningar, kontaktfrågan, mottagarens syntetiska adress, SMTP-fångstens domänregel och den faktiska sparade EML-filen.
- För PDF verifierar du typ och innehåll som behörig roll. För XML kontrollerar du XMLProviderModule, skrivbar filkatalog och att inskicket gjordes efter aktivering. Statistikens standardperiod kan vara avslutad och därför sakna nya testärenden.
- CKEditor 4.22.1 visar sin säkerhetsvarning i den verifierade uppsättningen. En åtgärd kräver licensierad CKEditor 4 LTS eller en kompatibilitetsmigrering till en annan editor. [CKEditor 4 till 5](https://ckeditor.com/docs/ckeditor5/latest/updating/ckeditor4/migration-from-ckeditor-4.html)

## Backup

För komplett miljökopia, ta en konsistent dump av sandboxtabellerna och arkivera runtimekonfiguration och filarkiv. Skydda dump, katalog, nycklar, ärenden och lokala lösenord. Lagra kontrollsummor och återläs endast i avsedd testmiljö. Vid återställning, stoppa appen, återläs databas och filer, kontrollera ägare och rättigheter, starta appen och upprepa inloggning, katalog, ärendeåtkomst och en syntetisk import. Publicera bara syntetiska `.oeflow`-exporter som har granskats.
