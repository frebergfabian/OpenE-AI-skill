-- Experimental registrations. The alternative-list handler must load before
-- providers that declare it as a required instance-manager dependency.
INSERT INTO openhierarchy_foreground_modules
 (classname,name,alias,description,anonymousAccess,userAccess,adminAccess,enabled,sectionID,visibleInMenu)
SELECT 'com.nordicpeak.flowengine.alternativelist.AlternativeListHandlerModule','Sandbox list handler','sandbox-list-handler','Local list handler',0,0,1,1,1,0
WHERE NOT EXISTS (SELECT 1 FROM openhierarchy_foreground_modules WHERE classname='com.nordicpeak.flowengine.alternativelist.AlternativeListHandlerModule');
CREATE TEMPORARY TABLE sandbox_query_modules (classname VARCHAR(255), alias VARCHAR(80), prefix VARCHAR(80));
INSERT INTO sandbox_query_modules VALUES
 ('com.nordicpeak.flowengine.queries.infoquery.InfoQueryProviderModule','sandbox-info','InfoQuery'),
 ('com.nordicpeak.flowengine.queries.DateTimeQueryProviderModule','sandbox-datetime','DateTimeQuery'),
 ('com.nordicpeak.flowengine.queries.dynamictablequery.DynamicTableQueryProviderModule','sandbox-dynamictable','DynamicTableQuery'),
 ('com.nordicpeak.flowengine.booking.queries.BookingQueryProviderModule','sandbox-booking-query','BookingQuery'),
 ('com.nordicpeak.flowengine.queue.query.QueueQueryProviderModule','sandbox-queue-query','QueueQuery'),
 ('com.nordicpeak.flowengine.queries.pudquery.PUDQueryProviderModule','sandbox-pud','PUDQuery'),
 ('com.nordicpeak.flowengine.queries.childinfoquery.ChildInfoQueryProviderModule','sandbox-child','ChildInfoQuery'),
 ('com.nordicpeak.flowengine.queries.ssbt.CompanyDetailsQueryProviderModule','sandbox-company','CompanyDetailsQuery'),
 ('com.nordicpeak.flowengine.queries.unsecuremultisignquery.UnsecureMultiSignQueryProviderModule','sandbox-multisign-query','UnsecureMultiSignQuery'),
 ('com.nordicpeak.flowengine.queries.manualmultisignquery.ManualMultiSignQueryProviderModule','sandbox-manual-multisign','ManualMultiSignQuery'),
 ('com.nordicpeak.flowengine.queries.colorpickerquery.ColorPickerQueryProviderModule','sandbox-color','ColorPickerQuery'),
 ('com.nordicpeak.flowengine.queries.fixedtablequery.FixedTableQueryProviderModule','sandbox-fixed-table','FixedTableQuery'),
 ('com.nordicpeak.flowengine.queries.manualcheckboxquery.ManualCheckboxQueryProviderModule','sandbox-manual-checkbox','ManualCheckboxQuery'),
 ('com.nordicpeak.flowengine.queries.manualsignquery.ManualSignQueryProviderModule','sandbox-manual-sign','ManualSignQuery'),
 ('com.nordicpeak.flowengine.queries.manualtextareaquery.ManualTextAreaQueryProviderModule','sandbox-manual-textarea','ManualTextAreaQuery'),
 ('com.nordicpeak.flowengine.queries.paragraphinfoquery.ParagraphInfoQueryProviderModule','sandbox-paragraph','ParagraphInfoQuery'),
 ('com.nordicpeak.flowengine.queries.pricetablequery.PriceTableQueryProviderModule','sandbox-price-table','PriceTableQuery'),
 ('com.nordicpeak.flowengine.queries.priorityquery.PriorityQueryProviderModule','sandbox-priority','PriorityQuery'),
 ('com.nordicpeak.flowengine.queries.ratingscalequery.RatingScaleQueryProviderModule','sandbox-rating','RatingScaleQuery'),
 ('com.nordicpeak.flowengine.queries.userlistquery.UserListQueryProviderModule','sandbox-user-list','UserListQuery'),
 ('com.nordicpeak.flowengine.queries.usertreequery.UserTreeQueryProviderModule','sandbox-user-tree','UserTreeQuery'),
 ('com.nordicpeak.flowengine.queries.internalcontactdetails.InternalContactDetailQueryProviderModule','sandbox-internal-contact','InternalContactDetailQuery'),
 ('com.nordicpeak.flowengine.queries.internaluserquery.InternalUserQueryProviderModule','sandbox-internal-user','InternalUserQuery'),
 ('com.nordicpeak.flowengine.queries.alternativelistcheckboxquery.AlternativeListCheckboxQueryProviderModule','sandbox-list-checkbox','AlternativeListCheckboxQuery'),
 ('com.nordicpeak.flowengine.queries.alternativelistdropdownquery.AlternativeListDropDownQueryProviderModule','sandbox-list-dropdown','AlternativeListDropDownQuery'),
 ('com.nordicpeak.flowengine.queries.alternativelistpriorityquery.AlternativeListPriorityQueryProviderModule','sandbox-list-priority','AlternativeListPriorityQuery'),
 ('com.nordicpeak.flowengine.queries.alternativelistradiobuttonquery.AlternativeListRadioButtonQueryProviderModule','sandbox-list-radio','AlternativeListRadioButtonQuery'),
 ('com.nordicpeak.flowengine.queries.listsearchqueryprovider.ListSearchQueryProviderModule','sandbox-list-search','ListSearchQuery'),
 ('com.nordicpeak.flowengine.queries.predefineduserquery.PredefinedUserQueryProviderModule','sandbox-predefined-user','PredefinedUserQuery');
INSERT INTO openhierarchy_foreground_modules
 (classname,name,alias,description,xslPath,xslPathType,anonymousAccess,userAccess,adminAccess,enabled,sectionID,visibleInMenu,staticContentPackage)
SELECT q.classname,q.prefix,q.alias,'Sandbox query provider',CONCAT(q.prefix,'Admin.sv.xsl'),'Classpath',1,1,1,1,1,0,
       CASE WHEN q.prefix <> 'InfoQuery' THEN 'staticcontent' ELSE NULL END
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
INSERT INTO openhierarchy_foreground_module_settings(moduleID,id,value)
SELECT m.moduleID,'supportedAttributes','email:E-post' FROM openhierarchy_foreground_modules m
WHERE m.classname='com.nordicpeak.flowengine.queries.internalcontactdetails.InternalContactDetailQueryProviderModule'
AND NOT EXISTS (SELECT 1 FROM openhierarchy_foreground_module_settings s WHERE s.moduleID=m.moduleID AND s.id='supportedAttributes');
INSERT INTO openhierarchy_foreground_module_settings(moduleID,id,value)
SELECT m.moduleID,'maxHits','20' FROM openhierarchy_foreground_modules m
WHERE m.classname='com.nordicpeak.flowengine.queries.listsearchqueryprovider.ListSearchQueryProviderModule'
AND NOT EXISTS (SELECT 1 FROM openhierarchy_foreground_module_settings s WHERE s.moduleID=m.moduleID AND s.id='maxHits');
INSERT INTO openhierarchy_foreground_module_settings(moduleID,id,value)
SELECT m.moduleID,'searchServiceURL','@SIMULATOR@/search' FROM openhierarchy_foreground_modules m
WHERE m.classname='com.nordicpeak.flowengine.queries.pudquery.PUDQueryProviderModule'
AND NOT EXISTS (SELECT 1 FROM openhierarchy_foreground_module_settings s WHERE s.moduleID=m.moduleID AND s.id='searchServiceURL');

CREATE TEMPORARY TABLE sandbox_support_modules (classname VARCHAR(255), alias VARCHAR(80), xsl VARCHAR(100));
INSERT INTO sandbox_support_modules VALUES
 ('com.nordicpeak.flowengine.booking.CalendarAdminModule','sandbox-calendar','CalendarAdminModule.sv.xsl'),
 ('com.nordicpeak.flowengine.booking.BookingSlotAdminModule','sandbox-bookings','BookingSlotAdminModule.sv.xsl'),
 ('com.nordicpeak.flowengine.booking.UserBookingModule','sandbox-mybookings','UserBookingModule.sv.xsl'),
 ('com.nordicpeak.flowengine.booking.notificationhandler.BookingNotificationHandler','sandbox-booking-notices','BookingNotificationHandler.sv.xsl'),
 ('com.nordicpeak.flowengine.queue.QueueAdminModule','sandbox-queues','QueueAdminModule.sv.xsl'),
 ('com.nordicpeak.flowengine.queue.QueueSlotAdminModule','sandbox-queue-slots','QueueSlotAdminModule.sv.xsl'),
 ('com.nordicpeak.flowengine.queue.UserQueueSlotModule','sandbox-myqueue','UserQueueSlotModule.sv.xsl'),
 ('com.nordicpeak.flowengine.integration.http.HTTPIntegrationProvider','sandbox-http','HTTPIntegrationProvider.sv.xsl'),
 ('se.exempelkoping.sandbox.SimulatedSMSSender','sandbox-sms',NULL),
 ('com.nordicpeak.childrelationprovider.DummyChildRelationProvider','sandbox-child-relations',NULL),
 ('se.exempelkoping.sandbox.SimulatedSwedbankPayProvider','sandbox-pay',NULL),
 ('com.nordicpeak.flowengine.paymentprovider.GenericFlowPaymentProvider','sandbox-flow-pay','GenericFlowPaymentProvider.sv.xsl'),
 ('com.nordicpeak.flowengine.paymentlog.AccountModule','sandbox-accounts','AccountModule.sv.xsl'),
 ('com.nordicpeak.flowengine.paymentlog.TransactionModule','sandbox-transactions','TransactionModule.sv.xsl'),
 ('com.nordicpeak.flowengine.paymentlog.TransactionFlowAdminExtensionModule','sandbox-transaction-settings','TransactionFlowAdminExtensionModule.sv.xsl'),
 ('com.nordicpeak.flowengine.multisigninghandlers.MultiSigningHandler2Module','sandbox-multisign','MultiSigningHandler2Module.sv.xsl'),
 ('com.nordicpeak.flowengine.documentsigner.DocumentSignerAdminModule','sandbox-documents','DocumentSignerAdminModule.sv.xsl'),
 ('com.nordicpeak.flowengine.documentsigner.DocumentSignerUserModule','sandbox-mydocuments','DocumentSignerUserModule.sv.xsl'),
 ('com.nordicpeak.flowengine.internalalternativelist.InternalAlternativeListProviderModule','sandbox-lists','InternalAlternativeListProviderModule.sv.xsl'),
 ('com.nordicpeak.flowengine.queries.userlistquery.FlowManagersListUserProvider','sandbox-list-managers',NULL),
 ('com.nordicpeak.flowengine.queries.usertreequery.UserHandlerListUserProvider','sandbox-list-users',NULL);
INSERT INTO openhierarchy_foreground_modules
 (classname,name,alias,description,xslPath,xslPathType,anonymousAccess,userAccess,adminAccess,enabled,sectionID,visibleInMenu,staticContentPackage)
SELECT q.classname,q.alias,q.alias,'Experimental sandbox module',q.xsl,IF(q.xsl IS NULL,NULL,'Classpath'),IF(q.alias='sandbox-pay',1,0),IF(q.alias IN ('sandbox-mybookings','sandbox-myqueue','sandbox-flow-pay','sandbox-multisign','sandbox-mydocuments'),1,0),1,1,1,0,IF(q.xsl IS NULL,NULL,'staticcontent')
FROM sandbox_support_modules q WHERE NOT EXISTS (SELECT 1 FROM openhierarchy_foreground_modules m WHERE m.classname=q.classname);
INSERT INTO openhierarchy_foreground_module_settings(moduleID,id,value)
SELECT m.moduleID,'supportedActionIDs','com.nordicpeak.flowengine.FlowBrowserModule.submit'
FROM sandbox_support_modules q JOIN openhierarchy_foreground_modules m ON m.classname=q.classname
WHERE NOT EXISTS (SELECT 1 FROM openhierarchy_foreground_module_settings s WHERE s.moduleID=m.moduleID AND s.id='supportedActionIDs');
DROP TEMPORARY TABLE sandbox_support_modules;
-- Local rule providers that do not require a remote service configuration.
CREATE TEMPORARY TABLE sandbox_rule_modules (classname VARCHAR(255), alias VARCHAR(80), xsl VARCHAR(160));
INSERT INTO sandbox_rule_modules VALUES
 ('com.nordicpeak.flowengine.evaluators.agequerystateevaluator.AgeQueryStateEvaluationProviderModule','sandbox-rule-age','AgeQueryStateEvaluationProviderModule.sv.xsl'),
 ('com.nordicpeak.flowengine.evaluators.childcountquerystateevaluator.ChildCountQueryStateEvaluationProviderModule','sandbox-rule-childcount','ChildCountQueryStateEvaluationProviderModule.sv.xsl'),
 ('com.nordicpeak.flowengine.evaluators.citizenidgroupmembershipquerystateevaluator.CitizenIDGroupMembershipQueryStateEvaluationProviderModule','sandbox-rule-id-group','CitizenIDGroupMembershipQueryStateEvaluationProviderModule.sv.xsl'),
 ('com.nordicpeak.flowengine.evaluators.citizenidentifier.CitizenIdentifierQueryStateEvaluationProviderModule','sandbox-rule-id','CitizenIdentifierQueryStateEvaluationProviderModule.sv.xsl'),
 ('com.nordicpeak.flowengine.evaluators.groupquerystateevaluator.GroupQueryStateEvaluationProviderModule','sandbox-rule-group','GroupQueryStateEvaluationProviderModule.sv.xsl'),
 ('com.nordicpeak.flowengine.evaluators.loggedin.LoggedInQueryStateEvaluationProviderModule','sandbox-rule-login','LoggedInQueryStateEvaluationProviderModule.sv.xsl'),
 ('com.nordicpeak.flowengine.evaluators.manager.ManagerQueryStateEvaluationProviderModule','sandbox-rule-manager','ManagerQueryStateEvaluationProviderModule.sv.xsl'),
 ('com.nordicpeak.flowengine.evaluators.maritalstatusquerystateevaluator.MaritalStatusQueryStateEvaluationProviderModule','sandbox-rule-marital','MaritalStatusQueryStateEvaluationProviderModule.sv.xsl'),
 ('com.nordicpeak.flowengine.evaluators.municipalitycode.MunicipalityCodeQueryStateEvaluationProviderModule','sandbox-rule-municipality','MunicipalityCodeQueryStateEvaluationProviderModule.sv.xsl'),
 ('com.nordicpeak.flowengine.evaluators.sendmailevaluator.SendMailEvaluationProviderModule','sandbox-rule-mail','SendMailEvaluationProviderModule.sv.xsl'),
 ('com.nordicpeak.flowengine.evaluators.sendtextfieldsmsevaluator.SendTextFieldSMSEvaluationProviderModule','sandbox-rule-sms','SendTextFieldSMSEvaluationProviderModule.sv.xsl'),
 ('com.nordicpeak.flowengine.evaluators.sendtextfieldmailevaluator.SendTextFieldMailEvaluationProviderModule','sandbox-rule-text-mail','SendTextFieldMailEvaluationProviderModule.sv.xsl'),
 ('com.nordicpeak.flowengine.evaluators.sendcontactquerymailevaluator.SendContactQueryMailEvaluationProviderModule','sandbox-rule-contact-mail','SendContactQueryMailEvaluationProviderModule.sv.xsl'),
 ('com.nordicpeak.flowengine.evaluators.userquerymailevaluator.UserQueryMailEvaluationProviderModule','sandbox-rule-user-mail','UserQueryMailEvaluationProviderModule.sv.xsl'),
 ('com.nordicpeak.flowengine.evaluators.internalaccount.MatchingInternalAccountQueryStateEvaluationProviderModule','sandbox-rule-internal-account','MatchingInternalAccountQueryStateEvaluationProviderModule.sv.xsl'),
 ('com.nordicpeak.flowengine.evaluators.officialaddressquerystateevaluator.OfficialAddressQueryStateEvaluationProviderModule','sandbox-rule-address','OfficialAddressQueryStateEvaluationProviderModule.sv.xsl'),
 ('com.nordicpeak.flowengine.evaluators.statusquerystate.StatusQueryStateEvaluationProviderModule','sandbox-rule-status','StatusQueryStateEvaluationProviderModule.sv.xsl'),
 ('com.nordicpeak.flowengine.evaluators.addclosestmanagerasflowinstancemanagerevaluator.AddClosestManagerAsFlowInstanceManagerEvaluationProviderModule','sandbox-rule-closest-manager','AddClosestManagerAsFlowInstanceManagerEvaluationProviderModule.sv.xsl'),
 ('com.nordicpeak.flowengine.evaluators.addflowinstancemanagersevaluator.AddFlowInstanceManagersEvaluationProviderModule','sandbox-rule-add-managers','AddFlowInstanceManagersEvaluationProviderModule.sv.xsl'),
 ('com.nordicpeak.flowengine.evaluators.anonymousposterevaluator.AnonymousPosterEvaluationProviderModule','sandbox-rule-anonymous','AnonymousPosterEvaluationProviderModule.sv.xsl'),
 ('com.nordicpeak.flowengine.evaluators.changeposterevaluator.ChangePosterEvaluationProviderModule','sandbox-rule-change-poster','ChangePosterEvaluationProviderModule.sv.xsl'),
 ('com.nordicpeak.flowengine.evaluators.multipartsigningmailevaluator.MultipartSigningMailEvaluatorEvaluationProviderModule','sandbox-rule-multipart-mail','MultipartSigningMailEvaluatorEvaluationProviderModule.sv.xsl'),
 ('com.nordicpeak.flowengine.evaluators.setattributefromalternativeid.SetAttributeFromAlternativeIDEvaluationProviderModule','sandbox-rule-list-attribute','SetAttributeFromAlternativeIDEvaluationProviderModule.sv.xsl'),
 ('com.nordicpeak.flowengine.evaluators.setcontactdetailsevaluator.SetContactDetailsEvaluationProviderModule','sandbox-rule-contact','SetContactDetailsEvaluationProviderModule.sv.xsl'),
 ('com.nordicpeak.flowengine.evaluators.settextfieldvalueattribute.SetTextFieldValueAttributeEvaluationProviderModule','sandbox-rule-text-attribute','SetTextFieldValueAttributeEvaluationProviderModule.sv.xsl'),
 ('com.nordicpeak.flowengine.evaluators.statusoverrideevaluator.StatusOverrideEvaluationProviderModule','sandbox-rule-status-override','StatusOverrideEvaluationProviderModule.sv.xsl'),
 ('com.nordicpeak.flowengine.evaluators.tableattributeevaluator.TableAttributeEvaluationProviderModule','sandbox-rule-table-attribute','TableAttributeEvaluationProviderModule.sv.xsl'),
 ('com.nordicpeak.flowengine.evaluators.tableattributesummaryevaluator.TableAttributeSummaryEvaluationProviderModule','sandbox-rule-table-summary','TableAttributeSummaryEvaluationProviderModule.sv.xsl'),
 ('com.nordicpeak.flowengine.evaluators.userage.UserAgeQueryStateEvaluationProviderModule','sandbox-rule-user-age','UserAgeQueryStateEvaluationProviderModule.sv.xsl');
INSERT INTO openhierarchy_foreground_modules
 (classname,name,alias,description,xslPath,xslPathType,anonymousAccess,userAccess,adminAccess,enabled,sectionID,visibleInMenu)
SELECT q.classname,q.alias,q.alias,'Sandbox rule provider',q.xsl,'Classpath',0,0,1,1,1,0
FROM sandbox_rule_modules q WHERE NOT EXISTS (SELECT 1 FROM openhierarchy_foreground_modules m WHERE m.classname=q.classname);
DROP TEMPORARY TABLE sandbox_rule_modules;
INSERT INTO openhierarchy_foreground_module_settings(moduleID,id,value)
SELECT m.moduleID,'pdfFilename','$flow.name - $flowInstance.flowInstanceID'
FROM openhierarchy_foreground_modules m
WHERE m.classname IN ('com.nordicpeak.flowengine.evaluators.addclosestmanagerasflowinstancemanagerevaluator.AddClosestManagerAsFlowInstanceManagerEvaluationProviderModule',
                      'com.nordicpeak.flowengine.evaluators.addflowinstancemanagersevaluator.AddFlowInstanceManagersEvaluationProviderModule')
AND NOT EXISTS (SELECT 1 FROM openhierarchy_foreground_module_settings s WHERE s.moduleID=m.moduleID AND s.id='pdfFilename');
INSERT INTO openhierarchy_foreground_module_settings(moduleID,id,value)
SELECT m.moduleID,v.id,v.value FROM openhierarchy_foreground_modules m
JOIN (SELECT 'child1' id,'true' value UNION ALL SELECT 'childID1','201001010006' UNION ALL SELECT 'ensam1','true') v
WHERE m.classname='com.nordicpeak.childrelationprovider.DummyChildRelationProvider';
INSERT INTO openhierarchy_foreground_module_settings(moduleID,id,value)
SELECT m.moduleID,v.id,v.value FROM openhierarchy_foreground_modules m JOIN
(SELECT 'com.nordicpeak.flowengine.documentsigner.DocumentSignerAdminModule' classname,'filestore' id,'/var/lib/opene/files/documents' value
 UNION ALL SELECT 'com.nordicpeak.flowengine.documentsigner.DocumentSignerAdminModule','tempDir','/var/lib/opene/tmp'
 UNION ALL SELECT 'com.nordicpeak.flowengine.documentsigner.DocumentSignerUserModule','disclaimerText','Endast syntetisk testsignering.'
 UNION ALL SELECT 'com.nordicpeak.flowengine.internalalternativelist.InternalAlternativeListProviderModule','tempDir','/var/lib/opene/tmp'
 UNION ALL SELECT 'com.nordicpeak.flowengine.internalalternativelist.InternalAlternativeListProviderModule','maxRequestSize','10'
 UNION ALL SELECT 'com.nordicpeak.flowengine.internalalternativelist.InternalAlternativeListProviderModule','ramThreshold','500'
 UNION ALL SELECT 'com.nordicpeak.flowengine.queries.userlistquery.FlowManagersListUserProvider','providerID','sandbox-managers'
 UNION ALL SELECT 'com.nordicpeak.flowengine.queries.usertreequery.UserHandlerListUserProvider','providerID','sandbox-users'
 UNION ALL SELECT 'com.nordicpeak.flowengine.multisigninghandlers.MultiSigningHandler2Module','serverURL','@OPENE@') v ON v.classname=m.classname;
