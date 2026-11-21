// Force the generated SOAP client to the current instance's simulator. User
// and manager authorization in the original module remains unchanged.
import fs from 'node:fs';
import path from 'node:path';
const file=path.join(path.resolve(process.argv[2]),'ArtviseIntegrationProvider','src','com','nordicpeak','flowengine','integration','artvise','ArtviseIntegrationProviderModule.java');
let source=fs.readFileSync(file,'latin1');
const old='WSUtils.setEndpoint((BindingProvider) service, settings.getEndpoint().getAddress());';
const replacement=`java.net.URI sandboxEndpoint = new java.net.URI(System.getProperty("opene.sandbox.simulator"));
			if (!"http".equals(sandboxEndpoint.getScheme()) || !"127.0.0.1".equals(sandboxEndpoint.getHost()) || sandboxEndpoint.getPort() < 1 || sandboxEndpoint.getUserInfo() != null) throw new IllegalStateException("Local simulator required");
			WSUtils.setEndpoint((BindingProvider) service, sandboxEndpoint.toString() + "/integration/artvise");`;
if(source.split(old).length!==2)throw Error('Unexpected Artvise client setup');
source=source.replace(old,replacement);
fs.writeFileSync(file,source,'latin1');
