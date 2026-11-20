-- Run after module migrations and after the reviewed fixture has been imported.
-- Names and data are synthetic; the transaction is safe to retry.
START TRANSACTION;
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
COMMIT;
