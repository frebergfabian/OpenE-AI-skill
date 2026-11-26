// Configure the original generated SSBT SOAP clients against the local
// simulator. Authorization and query behavior remain in the vendor module.
import fs from 'node:fs';
import path from 'node:path';
const file=path.join(path.resolve(process.argv[2]),'SSBTQueries','src','com','nordicpeak','flowengine','queries','ssbt','CompanyDetailsQueryProviderModule.java');
let source=fs.readFileSync(file,'latin1');
const setup='super.moduleConfigured();';
if(source.split(setup).length!==2)throw Error('Unexpected SSBT module setup');
source=source.replace(setup,setup+`
		// SSBT_SANDBOX_SOAP: no certificates or external targets in sandbox.
		if (System.getProperty("opene.sandbox.simulator") != null) {
			java.net.URI base = new java.net.URI(System.getProperty("opene.sandbox.simulator"));
			if (!"http".equals(base.getScheme()) || !"127.0.0.1".equals(base.getHost()) || base.getPort() < 1 || base.getUserInfo() != null) throw new IllegalStateException("Local simulator required");
			SsbtenService en = new SsbtenServicePorts().getSsbtenService();
			SsbtguService gu = new SsbtguServicePorts().getSsbtguService();
			SsbtroService ro = new SsbtroServicePorts().getSsbtroService();
			WSUtils.setEndpoint((BindingProvider) en, base.toString() + "/ssbt/en");
			WSUtils.setEndpoint((BindingProvider) gu, base.toString() + "/ssbt/gu");
			WSUtils.setEndpoint((BindingProvider) ro, base.toString() + "/ssbt/ro");
			WSUtils.setTimeouts((BindingProvider) en, connectionTimeout, requestTimeout);
			WSUtils.setTimeouts((BindingProvider) gu, connectionTimeout, requestTimeout);
			WSUtils.setTimeouts((BindingProvider) ro, connectionTimeout, requestTimeout);
			this.enServices = new HashMap<TargetEnvironment, SsbtenService>();
			this.guServices = new HashMap<TargetEnvironment, SsbtguService>();
			this.roServices = new HashMap<TargetEnvironment, SsbtroService>();
			for (TargetEnvironment env : TargetEnvironment.values()) { enServices.put(env,en); guServices.put(env,gu); roServices.put(env,ro); }
			configured = true;
			return;
		}`);
const endpointPattern=/(WSUtils\.setEndpoint\(\(BindingProvider\)\s+\w+,\s*")[^"]+("\);)/g;
const endpoints=[...source.matchAll(endpointPattern)];
if(endpoints.length!==9)throw Error(`Unexpected SSBT endpoint count: ${endpoints.length}`);
source=source.replace(endpointPattern,'$1http://127.0.0.1:1/disabled$2');
fs.writeFileSync(file,source,'latin1');
