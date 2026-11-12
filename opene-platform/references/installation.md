# Installation och kompatibilitet

Den verifierade kombinationen var Open ePlatform `a34ab2868be5bfcde7ea1be5f4ffd0be4601fba1` (2.0.0-SNAPSHOT), OpenHierarchy 1.4.5, JDK 21, Gradle-wrapper 9.7.0, Tomcat 9.0.123 och MariaDB 11.8 med Connector/J 2.7.6. Källan kompilerades med `--release 8` och kördes på Java 21. Använd en servletcontainer med `javax.servlet` för den här källversionen. Versionsnumren är en verifierad kombination för reproduktion. [Open ePlatform](https://github.com/Open-ePlatform/Open-ePlatform/tree/a34ab2868be5bfcde7ea1be5f4ffd0be4601fba1)

1. Läs upstreams `BUILDING.md`, `libs/README.md`, `build.gradle` och `settings.gradle`. Kör Gradle med repots wrapper.
2. Checka ut `framework/trunk` samt projekt för SiteProfile, WSModule, SAMLLoginProvider och MinimalUser från OpenHierarchy SVN r3190. PersonProvider och PaymentProvider hämtas från `interfaces/<projekt>/trunk` på samma revision. Se `svn info` för att kontrollera ursprunget.
3. Hämta Utils-projekt och externa bibliotek från Nordic Peaks utils-SVN r1649, inklusive cron4j-forken. PureCaptcha kommer från r15. Bygg SVN-källan med ISO-8859-1. WSModule behöver på modern Java använda JAXB RI:s `com.sun.xml.bind.v2.ClassFactory` i stället för JDK 8:s interna motsvarighet. Bevara upstreams licenser och verifiera hämtade arkiv.
4. Bygg bibliotek och lägg de JAR-filer som Gradle söker i `libs/`. Kör `./gradlew --no-daemon build openHierarchyDependencyReport`. Bygg webbrot från demodesignens WebContent, modulernas runtime-JAR:er, JDBC-driver och eventuell egen XSL/CSS/JS som överlagring.
5. Ladda upstreams demodatabas i ett nytt schema. Ge JDBC-kontot rättigheter enbart i detta schema. Konfigurera datakällan i `WEB-INF/config.xml`, lokala skrivbara filarkiv och separat loggning. Inaktivera demointegrationer och ersätt demokontots lösenord före första start. Flytta lokal autentisering till PBKDF2 enligt installerad SimpleUserProvider.
6. Bind Tomcat till testmiljön, använd lokalt betrott TLS och begränsa åtkomsten till avsedda klienter. Kör appservern med ett separat tjänstekonto och skrivbehörighet i de kataloger appen behöver. Håll databas, SMTP-fångst och registerfixturer på loopback.

SVN-ursprung: `svn://svn.openhierarchy.org/openhierarchy`, `svn://svn.nordicpeak.com/utils` och `svn://svn.nordicpeak.com/purecaptcha/trunk`.

## Migrationer

Den gamla demodumpen saknade `FlowBrowserModule.pay` och `FlowBrowserModule.multisigning` som FlowEngine-migration 137 förutsätter. Registrera de två åtgärds-ID:na i den rena labbdumpen före uppstart. Verifierad FlowEngine-migrering gick från 52 till 169. SQL-DDL kan ha slagit igenom även om versionsmarkören står kvar. Återställ en ren bas vid partiell migrering.

StandardFlowNotificationHandler fastnade vid v42 på MariaDBs radstorleksgräns. I den verifierade databasen hade den tomma inställningstabellen redan v41-schema. Flytt av långa oindexerade fält till TEXT med CHECK för ursprungliga längdgränser gjorde att resterande upstream-DDL kunde köras till v43. Kontrollera tabellens faktiska schema före motsvarande åtgärd. OpenHierarchy core nådde 36 och SimpleUserProvider 10.

SiteProfile r3190:s migreringsfil går från 6 till 8. Den sekventiella migreraren stannar därför vid 6. När profildomäner sparas kräver Java-koden redan v8-kolumnen. Efter backup och kontroll av aktuell nivå kan du använda upstreams v8-DDL `ALTER TABLE site_profile_domains ADD COLUMN sortIndex INTEGER UNSIGNED NOT NULL AFTER domain` och därefter markera just SiteProfilesAdminModules tabellgrupp som version 8.
