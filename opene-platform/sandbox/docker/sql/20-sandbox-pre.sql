-- Sandboxjusteringar som körs EN gång efter demodatabasen och före första start.
-- Bara labbvärden. Inga lösenord här (de sätts av seed-users.sh från miljövariabler).

-- FlowEngine-migrering 137 lägger standardstatusar med åtgärderna pay och multisigning.
-- Demodumpen saknar dem och modulerna registrerar dem först efter migreringen, så de
-- måste finnas innan första start. Annars avbryts migreringen halvvägs (DDL kan inte
-- rullas tillbaka) och alla senare starter fastnar på "Duplicate column" i v53.
INSERT IGNORE INTO flowengine_flow_actions (actionID, name, required) VALUES
  ('com.nordicpeak.flowengine.FlowBrowserModule.pay', 'Användare går vidare till betalning', 0),
  ('com.nordicpeak.flowengine.FlowBrowserModule.multisigning', 'Användare går vidare till flerpartsignering', 0);

-- All e-post till den lokala SMTP-fångsten (Mailpit). Ingenting lämnar sandboxen.
DELETE s FROM openhierarchy_foreground_module_settings s
  JOIN openhierarchy_foreground_modules m ON m.moduleID = s.moduleID
 WHERE m.classname = 'se.unlogic.hierarchy.foregroundmodules.mailsenders.direct.DirectMailSender'
   AND s.id IN ('host', 'port', 'useAuth', 'connectionTimeout', 'priority');
INSERT INTO openhierarchy_foreground_module_settings (moduleID, id, value)
SELECT m.moduleID, v.id, v.value
  FROM openhierarchy_foreground_modules m
  JOIN (SELECT 'host' id, 'mail' value UNION ALL SELECT 'port', '1025'
        UNION ALL SELECT 'useAuth', 'false' UNION ALL SELECT 'connectionTimeout', '5000'
        UNION ALL SELECT 'priority', '100') v
 WHERE m.classname = 'se.unlogic.hierarchy.foregroundmodules.mailsenders.direct.DirectMailSender';

-- Filkataloger och tempkataloger (volymen opene-files i containern).
CREATE TEMPORARY TABLE sandbox_settings (classname VARCHAR(255), id VARCHAR(255), value VARCHAR(255));
INSERT INTO sandbox_settings VALUES
  ('com.nordicpeak.flowengine.FlowAdminModule',            'tempDir',           '/var/lib/opene/tmp'),
  ('com.nordicpeak.flowengine.FlowAdminModule',            'pdfFormFilestore',  '/var/lib/opene/files/flowforms'),
  ('com.nordicpeak.flowengine.FlowBrowserModule',          'tempDir',           '/var/lib/opene/tmp'),
  ('com.nordicpeak.flowengine.UserFlowInstanceModule',     'tempDir',           '/var/lib/opene/tmp'),
  ('com.nordicpeak.flowengine.FlowInstanceAdminModule',    'tempDir',           '/var/lib/opene/tmp'),
  ('com.nordicpeak.flowengine.pdf.PDFGeneratorModule',     'pdfDir',            '/var/lib/opene/files/pdf'),
  ('com.nordicpeak.flowengine.pdf.PDFGeneratorModule',     'tempDir',           '/var/lib/opene/tmp'),
  ('com.nordicpeak.flowengine.XMLProviderModule',          'xmlDir',            '/var/lib/opene/files/xml'),
  ('com.nordicpeak.flowengine.queries.fileuploadquery.FileUploadQueryProviderModule', 'fileStore',     '/var/lib/opene/files/uploads'),
  ('com.nordicpeak.flowengine.queries.fileuploadquery.FileUploadQueryProviderModule', 'tempFileStore', '/var/lib/opene/tmp'),
  ('se.unlogic.openhierarchy.foregroundmodules.siteprofile.SiteProfilesAdminModule',  'filestore',     '/var/lib/opene/files/siteprofile'),
  ('se.unlogic.openhierarchy.foregroundmodules.siteprofile.SiteProfilesAdminModule',  'tempFileStore', '/var/lib/opene/tmp');
DELETE s FROM openhierarchy_foreground_module_settings s
  JOIN openhierarchy_foreground_modules m ON m.moduleID = s.moduleID
  JOIN sandbox_settings v ON v.classname = m.classname AND v.id = s.id;
INSERT INTO openhierarchy_foreground_module_settings (moduleID, id, value)
SELECT m.moduleID, v.id, v.value FROM openhierarchy_foreground_modules m JOIN sandbox_settings v ON v.classname = m.classname;
DELETE s FROM openhierarchy_background_module_settings s
  JOIN openhierarchy_background_modules m ON m.moduleID = s.moduleID
  JOIN sandbox_settings v ON v.classname = m.classname AND v.id = s.id;
INSERT INTO openhierarchy_background_module_settings (moduleID, id, value)
SELECT m.moduleID, v.id, v.value FROM openhierarchy_background_modules m JOIN sandbox_settings v ON v.classname = m.classname;
DROP TEMPORARY TABLE sandbox_settings;

-- Formatvaliderare (e-post, personnummer, positivt heltal) finns i FlowEngine men är inte
-- registrerade i demodumpen. Utan modulen saknar textfälten validerare.
INSERT INTO openhierarchy_foreground_modules
  (classname, name, alias, description, xslPath, xslPathType, anonymousAccess, userAccess, adminAccess, enabled, visibleInMenu, sectionID, dataSourceID, staticContentPackage, requiredProtocol)
SELECT 'com.nordicpeak.flowengine.formatvalidation.FormatValidationHandlerModule', 'Formatvaliderare', 'formatvalidation',
       'Registrerad av sandboxen', NULL, NULL, 0, 0, 1, 1, 0, 1, NULL, NULL, NULL
 WHERE NOT EXISTS (SELECT 1 FROM openhierarchy_foreground_modules WHERE classname = 'com.nordicpeak.flowengine.formatvalidation.FormatValidationHandlerModule');

-- Notishanteraren är avstängd i demodumpen. Den behövs för e-postnotiser (fångas av Mailpit).
UPDATE openhierarchy_foreground_modules SET enabled = 1
 WHERE classname = 'com.nordicpeak.flowengine.notifications.StandardFlowNotificationHandler';

-- Moduler i demodumpen vars klasser inte finns i Open ePlatform-repot (äldre eller
-- separata Nordic Peak-moduler). De stängs av så att uppstarten blir felfri.
UPDATE openhierarchy_foreground_modules SET enabled = 0 WHERE classname IN (
  'com.nordicpeak.flowengine.TextTagAdminModule',
  'com.nordicpeak.flowengine.UserFavouriteForegroundModule',
  'se.unlogic.hierarchy.foregroundmodules.test.EmailMailTestModule');
UPDATE openhierarchy_background_modules SET enabled = 0 WHERE classname IN (
  'com.nordicpeak.flowengine.NewEventsBackgroundModule',
  'com.nordicpeak.flowengine.UserFavouriteBackgroundModule');

-- Demodumpen kräver HTTPS för rotsektionen och omdirigerar till https://<host>/ utan port.
-- Sandboxen nås via HTTP på 127.0.0.1 bakom proxyn, så kravet tas bort.
UPDATE openhierarchy_sections SET requiredProtocol = NULL;
UPDATE openhierarchy_foreground_modules SET requiredProtocol = NULL;

-- The background user menu serves CSS and JS from its classpath package.
UPDATE openhierarchy_background_modules SET staticContentPackage = 'staticcontent'
 WHERE classname = 'com.nordicpeak.flowengine.UserFlowInstanceMenuModule';

-- Demokontot 'admin' från upstream har ett känt standardlösenord: inaktivera det.
-- Sandboxens administratör skapas av seed-users.sh med lösenord från .env.
UPDATE simple_users SET enabled = 0 WHERE username = 'admin';
