-- Sandboxjusteringar som körs EN gång efter demodatabasen och före första start.
-- Bara labbvärden. Inga lösenord här (de sätts av seed-users.sh från miljövariabler).

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

-- Demokontot 'admin' från upstream har ett känt standardlösenord: inaktivera det.
-- Sandboxens administratör skapas av seed-users.sh med lösenord från .env.
UPDATE simple_users SET enabled = 0 WHERE username = 'admin';
