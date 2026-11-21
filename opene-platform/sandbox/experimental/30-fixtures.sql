-- Run after module migrations and after the reviewed fixture has been imported.
-- Names and data are synthetic; the transaction is safe to retry.
START TRANSACTION;
-- The upstream migration skips version 7; version 8 adds the ordered domain
-- column that the current SiteProfile DAO requires. The local profile is used
-- by DocumentSigner to associate a document with its originating site.
ALTER TABLE site_profile_domains ADD COLUMN IF NOT EXISTS sortIndex INTEGER UNSIGNED NOT NULL AFTER domain;
UPDATE table_versions SET version=8 WHERE tableGroupName='se.unlogic.openhierarchy.foregroundmodules.siteprofile.SiteProfilesAdminModule' AND version=6;
INSERT INTO site_profiles(name,design) SELECT 'Sandbox lokal',NULL WHERE NOT EXISTS (SELECT 1 FROM site_profiles WHERE name='Sandbox lokal');
SET @profile=(SELECT profileID FROM site_profiles WHERE name='Sandbox lokal' ORDER BY profileID LIMIT 1);
INSERT IGNORE INTO site_profile_domains(profileID,domain,sortIndex) VALUES(@profile,'127.0.0.1',0),(@profile,'localhost',1);
-- Publish only our named synthetic fixtures, using a relative date so a fresh
-- installation works regardless of when the portable build was produced.
UPDATE flowengine_flows SET enabled=1,publishDate=CURRENT_DATE,unPublishDate=NULL
WHERE name IN ('Funktionslabbet - portabelt utkast','Sandbox fastighetstest','Sandbox barntest','Sandbox betalningstest','Sandbox foretagstest','Sandbox flerpartssignering');
UPDATE flowengine_flows SET requireSigning=1,appendSigningSignatureToPDF=1,showSubmitSurvey=0
WHERE name IN ('Funktionslabbet - portabelt utkast','Sandbox betalningstest','Sandbox flerpartssignering');
INSERT INTO booking_calendars(name,enabled,price,hoursBeforeBookingLocked,disableLeaving,disableRebooking,restrictRebookingToPreviousManager,allowDuplicateBookings)
SELECT 'Sandbox testkalender',1,0,0,0,0,0,0 WHERE NOT EXISTS (SELECT 1 FROM booking_calendars WHERE name='Sandbox testkalender');
SET @calendar=(SELECT calendarID FROM booking_calendars WHERE name='Sandbox testkalender' ORDER BY calendarID LIMIT 1);
INSERT IGNORE INTO booking_calendar_manager_groups(calendarID,groupID) VALUES(@calendar,11);
INSERT INTO booking_slots(calendarID,start,end,added,poster,size)
SELECT @calendar,DATE_ADD(CURRENT_DATE,INTERVAL 3 DAY)+INTERVAL 10 HOUR,DATE_ADD(CURRENT_DATE,INTERVAL 3 DAY)+INTERVAL 10 HOUR + INTERVAL 30 MINUTE,NOW(),u.userID,4 FROM simple_users u
WHERE u.username='sandbox-admin' AND NOT EXISTS (SELECT 1 FROM booking_slots WHERE calendarID=@calendar);
INSERT INTO queues(name,enabled,showQueuePosition,price,allowDuplicateQueueSlots)
SELECT 'Sandbox testko',1,1,0,0 WHERE NOT EXISTS (SELECT 1 FROM queues WHERE name='Sandbox testko');
SET @queue=(SELECT queueID FROM queues WHERE name='Sandbox testko' ORDER BY queueID LIMIT 1);
INSERT IGNORE INTO queue_manager_groups(queueID,groupID) VALUES(@queue,11);
UPDATE booking_queries q JOIN flowengine_query_descriptors d ON d.queryID=q.queryID JOIN flowengine_steps s ON s.stepID=d.stepID JOIN flowengine_flows f ON f.flowID=s.flowID
SET q.calendarID=@calendar WHERE f.name='Funktionslabbet - portabelt utkast';
UPDATE queue_queries q JOIN flowengine_query_descriptors d ON d.queryID=q.queryID JOIN flowengine_steps s ON s.stepID=d.stepID JOIN flowengine_flows f ON f.flowID=s.flowID
SET q.queueID=@queue WHERE f.name='Funktionslabbet - portabelt utkast';
UPDATE contact_detail_queries q JOIN flowengine_query_descriptors d ON d.queryID=q.queryID JOIN flowengine_steps s ON s.stepID=d.stepID JOIN flowengine_flows f ON f.flowID=s.flowID
SET q.fieldMobilePhone='VISIBLE',q.allowSMS=1 WHERE f.name='Funktionslabbet - portabelt utkast';
INSERT INTO flowengine_flow_family_manager_groups(flowFamilyID,groupID,restricted,allowUpdatingManagers,hideGroupInAssignment,notifyGroupMembersPersonally)
SELECT f.flowFamilyID,11,0,0,0,0 FROM flowengine_flows f WHERE f.name='Funktionslabbet - portabelt utkast'
AND NOT EXISTS (SELECT 1 FROM flowengine_flow_family_manager_groups g WHERE g.flowFamilyID=f.flowFamilyID AND g.groupID=11);
INSERT INTO http_integration_settings(flowFamilyID,enabled,url,triggerOnStatusChanges)
SELECT f.flowFamilyID,1,'@SIMULATOR@/integration/http',0 FROM flowengine_flows f WHERE f.name='Funktionslabbet - portabelt utkast'
AND NOT EXISTS (SELECT 1 FROM http_integration_settings h WHERE h.flowFamilyID=f.flowFamilyID);
INSERT INTO artvise_integration_endpoints(name,address)
SELECT 'Sandbox Artvise','@SIMULATOR@/integration/artvise' WHERE NOT EXISTS (SELECT 1 FROM artvise_integration_endpoints WHERE name='Sandbox Artvise');
SET @artvise=(SELECT endpointID FROM artvise_integration_endpoints WHERE name='Sandbox Artvise' ORDER BY endpointID LIMIT 1);
UPDATE artvise_integration_endpoints SET address='@SIMULATOR@/integration/artvise' WHERE endpointID=@artvise;
INSERT INTO artvise_integration_settings(flowFamilyID,enabled,endpointID,apiKey,caseInboxGuid,caseHTMLDescriptionObjectTypeGuid,caseHTMLDescriptionName,caseHTMLDescriptionContent,addAttachmentsSeparately,sendErrorEmail,requiresAnswer)
SELECT f.flowFamilyID,1,@artvise,'sandbox-key','00000000-0000-4000-8000-000000000001','00000000-0000-4000-8000-000000000002','Syntetiskt arende','Endast syntetiskt test',0,0,0
FROM flowengine_flows f WHERE f.name='Funktionslabbet - portabelt utkast'
AND NOT EXISTS (SELECT 1 FROM artvise_integration_settings a WHERE a.flowFamilyID=f.flowFamilyID);
INSERT INTO easit_integration_endpoints(name,webServiceAddress,apiKey)
SELECT 'Sandbox Easit','@SIMULATOR@/integration/easit','sandbox-key' WHERE NOT EXISTS (SELECT 1 FROM easit_integration_endpoints WHERE name='Sandbox Easit');
SET @easit=(SELECT endpointID FROM easit_integration_endpoints WHERE name='Sandbox Easit' ORDER BY endpointID LIMIT 1);
UPDATE easit_integration_endpoints SET webServiceAddress='@SIMULATOR@/integration/easit' WHERE endpointID=@easit;
INSERT INTO easit_integration_settings(flowFamilyID,enabled,endpointID)
SELECT f.flowFamilyID,1,@easit FROM flowengine_flows f WHERE f.name='Sandbox barntest'
AND NOT EXISTS (SELECT 1 FROM easit_integration_settings e WHERE e.flowFamilyID=f.flowFamilyID);
INSERT INTO easit_integration_parametergroups(flowFamilyID,name,importHandlerIdentifier,enabled,importWithPrevious,importParametersSeparately,sortIndex)
SELECT f.flowFamilyID,'Sandbox arende','sandbox-import',1,0,0,1 FROM flowengine_flows f WHERE f.name='Sandbox barntest'
AND NOT EXISTS (SELECT 1 FROM easit_integration_parametergroups g WHERE g.flowFamilyID=f.flowFamilyID AND g.name='Sandbox arende');
-- Other notification switches remain off. Contacts still choose their channels.
INSERT INTO flow_familiy_notification_settings(flowFamilyID,sendFlowInstanceSubmittedUserEmail,flowInstanceSubmittedUserEmailSubject,flowInstanceSubmittedUserEmailMessage,sendFlowInstanceSubmittedUserSMS,flowInstanceSubmittedUserSMS)
SELECT f.flowFamilyID,1,'Sandbox - fiktivt arende mottaget','<p>Det syntetiska testarendet har skickats in.</p>',1,'Sandbox: det syntetiska testarendet har skickats in.'
FROM flowengine_flows f WHERE f.name='Funktionslabbet - portabelt utkast'
AND NOT EXISTS (SELECT 1 FROM flow_familiy_notification_settings n WHERE n.flowFamilyID=f.flowFamilyID);
INSERT INTO booking_calendars(name,enabled,price,hoursBeforeBookingLocked,disableLeaving,disableRebooking,restrictRebookingToPreviousManager,allowDuplicateBookings)
SELECT 'Sandbox betalningskalender',1,100,0,0,0,0,0 WHERE NOT EXISTS (SELECT 1 FROM booking_calendars WHERE name='Sandbox betalningskalender');
SET @paycalendar=(SELECT calendarID FROM booking_calendars WHERE name='Sandbox betalningskalender' ORDER BY calendarID LIMIT 1);
INSERT IGNORE INTO booking_calendar_manager_groups(calendarID,groupID) VALUES(@paycalendar,11);
INSERT INTO booking_slots(calendarID,start,end,added,poster,size)
SELECT @paycalendar,DATE_ADD(CURRENT_DATE,INTERVAL 3 DAY)+INTERVAL 11 HOUR,DATE_ADD(CURRENT_DATE,INTERVAL 3 DAY)+INTERVAL 11 HOUR+INTERVAL 30 MINUTE,NOW(),u.userID,4 FROM simple_users u
WHERE u.username='sandbox-admin' AND NOT EXISTS (SELECT 1 FROM booking_slots WHERE calendarID=@paycalendar);
UPDATE booking_queries q JOIN flowengine_query_descriptors d ON d.queryID=q.queryID JOIN flowengine_steps s ON s.stepID=d.stepID JOIN flowengine_flows f ON f.flowID=s.flowID
SET q.calendarID=@paycalendar,q.invoiceLineText='Syntetisk bokning' WHERE f.name='Sandbox betalningstest';
INSERT INTO queues(name,enabled,showQueuePosition,price,allowDuplicateQueueSlots)
SELECT 'Sandbox betalningsko',1,1,0,0 WHERE NOT EXISTS (SELECT 1 FROM queues WHERE name='Sandbox betalningsko');
SET @payqueue=(SELECT queueID FROM queues WHERE name='Sandbox betalningsko' ORDER BY queueID LIMIT 1);
INSERT IGNORE INTO queue_manager_groups(queueID,groupID) VALUES(@payqueue,11);
UPDATE queue_queries q JOIN flowengine_query_descriptors d ON d.queryID=q.queryID JOIN flowengine_steps s ON s.stepID=d.stepID JOIN flowengine_flows f ON f.flowID=s.flowID
SET q.queueID=@payqueue WHERE f.name='Sandbox betalningstest';
INSERT INTO flow_generic_payment_settings(flowFamilyID,invoicePaymentEventDescription)
SELECT f.flowFamilyID,'Simulerad faktura' FROM flowengine_flows f WHERE f.name='Sandbox betalningstest'
AND NOT EXISTS (SELECT 1 FROM flow_generic_payment_settings p WHERE p.flowFamilyID=f.flowFamilyID);
INSERT INTO payment_chart_of_accounts(name,archived) SELECT 'Sandbox kontoplan',0 WHERE NOT EXISTS (SELECT 1 FROM payment_chart_of_accounts WHERE name='Sandbox kontoplan');
SET @chart=(SELECT chartID FROM payment_chart_of_accounts WHERE name='Sandbox kontoplan' ORDER BY chartID LIMIT 1);
INSERT INTO payment_accounts(name,accountType,chartID,archived) SELECT 'Sandbox intakt','INCOME',@chart,0
WHERE NOT EXISTS (SELECT 1 FROM payment_accounts WHERE name='Sandbox intakt' AND chartID=@chart);
SET @account=(SELECT accountID FROM payment_accounts WHERE name='Sandbox intakt' AND chartID=@chart ORDER BY accountID LIMIT 1);
INSERT INTO payment_flow_family_settings(flowFamilyID,accountID,paymentProductSettingsEnabled)
SELECT f.flowFamilyID,@account,0 FROM flowengine_flows f WHERE f.name='Sandbox betalningstest'
AND NOT EXISTS (SELECT 1 FROM payment_flow_family_settings p WHERE p.flowFamilyID=f.flowFamilyID);
COMMIT;
