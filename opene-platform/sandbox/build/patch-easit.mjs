// Restrict the Easit SOAP client to this sandbox instance's simulator.
import fs from 'node:fs';
import path from 'node:path';
const file=path.join(path.resolve(process.argv[2]),'EasitIntegrationProvider','src','com','nordicpeak','flowengine','integration','easit','EasitIntegrationProviderModule.java');
let source=fs.readFileSync(file,'latin1');
const old='((BindingProvider) proxy).getRequestContext().put(BindingProvider.ENDPOINT_ADDRESS_PROPERTY, integrationEndpoint.getWebServiceAddress());';
const replacement=`java.net.URI sandboxEndpoint = java.net.URI.create(System.getProperty("opene.sandbox.simulator"));
		if (!"http".equals(sandboxEndpoint.getScheme()) || !"127.0.0.1".equals(sandboxEndpoint.getHost()) || sandboxEndpoint.getPort() < 1 || sandboxEndpoint.getUserInfo() != null) throw new IllegalStateException("Local simulator required");
		((BindingProvider) proxy).getRequestContext().put(BindingProvider.ENDPOINT_ADDRESS_PROPERTY, sandboxEndpoint.toString() + "/integration/easit");`;
if(source.includes(old)){
 if(source.split(old).length!==2)throw Error('Unexpected Easit client setup');
 source=source.replace(old,replacement);
}else if(!source.includes('sandboxEndpoint.toString() + "/integration/easit"'))throw Error('Unexpected Easit client setup');
fs.writeFileSync(file,source,'latin1');
// Preserve the requested case-sensitive comparison while fixing an invalid
// BINARY plus general_ci declaration rejected by MariaDB 11.8.
const migration=path.join(path.dirname(file),'DB script.xml');
let sql=fs.readFileSync(migration,'latin1');
const invalid='VARCHAR(255) BINARY CHARACTER SET utf8 COLLATE utf8_general_ci NOT NULL';
const fixed='VARCHAR(255) CHARACTER SET utf8 COLLATE utf8_bin NOT NULL';
if(sql.includes(invalid))sql=sql.replace(invalid,fixed);
else if(!sql.includes(fixed))throw Error('Unexpected Easit migration');
fs.writeFileSync(migration,sql,'latin1');
