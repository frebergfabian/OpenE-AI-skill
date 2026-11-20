-- Experimental registrations for the three self-contained query providers.
-- Booking and queue require further modules and seeded calendars and are not enabled here.
CREATE TEMPORARY TABLE sandbox_query_modules (classname VARCHAR(255), alias VARCHAR(80), prefix VARCHAR(80));
INSERT INTO sandbox_query_modules VALUES
 ('com.nordicpeak.flowengine.queries.infoquery.InfoQueryProviderModule','sandbox-info','InfoQuery'),
 ('com.nordicpeak.flowengine.queries.DateTimeQueryProviderModule','sandbox-datetime','DateTimeQuery'),
 ('com.nordicpeak.flowengine.queries.dynamictablequery.DynamicTableQueryProviderModule','sandbox-dynamictable','DynamicTableQuery'),
 ('com.nordicpeak.flowengine.booking.queries.BookingQueryProviderModule','sandbox-booking-query','BookingQuery'),
 ('com.nordicpeak.flowengine.queue.query.QueueQueryProviderModule','sandbox-queue-query','QueueQuery');
INSERT INTO openhierarchy_foreground_modules
 (classname,name,alias,description,xslPath,xslPathType,anonymousAccess,userAccess,adminAccess,enabled,sectionID,visibleInMenu,staticContentPackage)
SELECT q.classname,q.prefix,q.alias,'Sandbox query provider',CONCAT(q.prefix,'Admin.sv.xsl'),'Classpath',1,1,1,1,1,0,
       CASE WHEN q.prefix IN ('DateTimeQuery','DynamicTableQuery','BookingQuery','QueueQuery') THEN 'staticcontent' ELSE NULL END
FROM sandbox_query_modules q WHERE NOT EXISTS (SELECT 1 FROM openhierarchy_foreground_modules m WHERE m.classname=q.classname);
INSERT INTO openhierarchy_foreground_module_settings(moduleID,id,value)
SELECT m.moduleID,'queryStyleSheet',CONCAT(q.prefix,'.sv.xsl') FROM sandbox_query_modules q JOIN openhierarchy_foreground_modules m ON m.classname=q.classname
WHERE NOT EXISTS (SELECT 1 FROM openhierarchy_foreground_module_settings s WHERE s.moduleID=m.moduleID AND s.id='queryStyleSheet');
INSERT INTO openhierarchy_foreground_module_settings(moduleID,id,value)
SELECT m.moduleID,'pdfStyleSheet',CONCAT(q.prefix,'PDF.sv.xsl') FROM sandbox_query_modules q JOIN openhierarchy_foreground_modules m ON m.classname=q.classname
WHERE NOT EXISTS (SELECT 1 FROM openhierarchy_foreground_module_settings s WHERE s.moduleID=m.moduleID AND s.id='pdfStyleSheet');
INSERT IGNORE INTO flowengine_flow_type_allowed_queries(flowTypeID,queryTypeID)
SELECT t.flowTypeID,q.classname FROM flowengine_flow_types t CROSS JOIN sandbox_query_modules q;
DROP TEMPORARY TABLE sandbox_query_modules;

CREATE TEMPORARY TABLE sandbox_support_modules (classname VARCHAR(255), alias VARCHAR(80), xsl VARCHAR(100));
INSERT INTO sandbox_support_modules VALUES
 ('com.nordicpeak.flowengine.booking.CalendarAdminModule','sandbox-calendar','CalendarAdminModule.sv.xsl'),
 ('com.nordicpeak.flowengine.booking.BookingSlotAdminModule','sandbox-bookings','BookingSlotAdminModule.sv.xsl'),
 ('com.nordicpeak.flowengine.booking.UserBookingModule','sandbox-mybookings','UserBookingModule.sv.xsl'),
 ('com.nordicpeak.flowengine.booking.notificationhandler.BookingNotificationHandler','sandbox-booking-notices','BookingNotificationHandler.sv.xsl'),
 ('com.nordicpeak.flowengine.queue.QueueAdminModule','sandbox-queues','QueueAdminModule.sv.xsl'),
 ('com.nordicpeak.flowengine.queue.QueueSlotAdminModule','sandbox-queue-slots','QueueSlotAdminModule.sv.xsl'),
 ('com.nordicpeak.flowengine.queue.UserQueueSlotModule','sandbox-myqueue','UserQueueSlotModule.sv.xsl');
INSERT INTO openhierarchy_foreground_modules
 (classname,name,alias,description,xslPath,xslPathType,anonymousAccess,userAccess,adminAccess,enabled,sectionID,visibleInMenu,staticContentPackage)
SELECT q.classname,q.alias,q.alias,'Experimental sandbox module',q.xsl,'Classpath',0,0,1,1,1,0,'staticcontent'
FROM sandbox_support_modules q WHERE NOT EXISTS (SELECT 1 FROM openhierarchy_foreground_modules m WHERE m.classname=q.classname);
INSERT INTO openhierarchy_foreground_module_settings(moduleID,id,value)
SELECT m.moduleID,'supportedActionIDs','com.nordicpeak.flowengine.FlowBrowserModule.submit'
FROM sandbox_support_modules q JOIN openhierarchy_foreground_modules m ON m.classname=q.classname
WHERE NOT EXISTS (SELECT 1 FROM openhierarchy_foreground_module_settings s WHERE s.moduleID=m.moduleID AND s.id='supportedActionIDs');
DROP TEMPORARY TABLE sandbox_support_modules;
