// Applies an explicit sandbox-only transport adapter, without changing user checks.
// This is a local JSON test contract, not an implementation of Bolagsverket SOAP.
import fs from 'node:fs';
import path from 'node:path';
const root=path.resolve(process.argv[2]);
const file=path.join(root,'SSBTQueries','src','com','nordicpeak','flowengine','queries','ssbt','CompanyDetailsQueryProviderModule.java');
let text=fs.readFileSync(file,'latin1');
if(!text.includes('SSBT_SANDBOX_TRANSPORT')){
 const setup='super.moduleConfigured();';
 if(text.split(setup).length!==2)throw Error('Unexpected SSBT module setup');
 text=text.replace(setup,setup+`
		// SSBT_SANDBOX_TRANSPORT: production SOAP endpoints are never initialized in sandbox.
		if (System.getProperty("opene.sandbox.simulator") != null) {
			java.net.URI u = new java.net.URI(System.getProperty("opene.sandbox.simulator"));
			if (!"http".equals(u.getScheme()) || !"127.0.0.1".equals(u.getHost()) || u.getPort() < 1) throw new IllegalStateException("Local simulator required");
			configured = true;
			return;
		}`);
 const marker='log.info("Getting companies for user " + user);';
 if(text.split(marker).length!==2)throw Error('Unexpected SSBT lookup method');
 text=text.replace(marker,`
		if (System.getProperty("opene.sandbox.simulator") != null) {
			try {
				java.net.URI base = new java.net.URI(System.getProperty("opene.sandbox.simulator"));
				if (!"http".equals(base.getScheme()) || !"127.0.0.1".equals(base.getHost()) || base.getPort() < 1) throw new java.io.IOException("Local simulator required");
				String path = "/ssbt/companies?citizenIdentifier=" + java.net.URLEncoder.encode(citizienIdentifier, "UTF-8");
				java.net.HttpURLConnection c = (java.net.HttpURLConnection) new java.net.URL(base.toString() + path).openConnection(java.net.Proxy.NO_PROXY);
				c.setConnectTimeout(3000); c.setReadTimeout(8000); c.setInstanceFollowRedirects(false);
				if (c.getResponseCode() != 200) throw new java.io.IOException("Simulator HTTP " + c.getResponseCode());
				java.io.ByteArrayOutputStream b = new java.io.ByteArrayOutputStream();
				try (java.io.InputStream in = c.getInputStream()) { byte[] block = new byte[2048]; int n; while ((n = in.read(block)) != -1) { if (b.size() + n > 65536) throw new java.io.IOException("Response too large"); b.write(block, 0, n); } } finally { c.disconnect(); }
				com.github.openjson.JSONArray items = new com.github.openjson.JSONObject(new String(b.toByteArray(), java.nio.charset.StandardCharsets.UTF_8)).getJSONArray("companies");
				if (queryInstance.getQuery().getCompanyTypes() == null || queryInstance.getQuery().getCompanyTypes().isEmpty()) throw new java.io.IOException("No permitted company type");
				java.util.List<Company> companies = new java.util.ArrayList<Company>();
				for (int j = 0; j < items.length(); j++) {
					com.github.openjson.JSONObject item = items.getJSONObject(j); Company company = new Company();
					company.setOrganizationNumber(item.getString("organizationNumber")); company.setCompanyName(item.getString("name"));
					company.setCompanyType(queryInstance.getQuery().getCompanyTypes().get(0)); CompanyDetails details = new CompanyDetails();
					details.setCompanyName(company.getCompanyName()); details.setLegalEntityAddress(item.getString("address")); details.setLegalEntityZip(item.getString("zip"));
					details.setLegalEntityPostalAddress(item.getString("city")); details.setEmail(item.getString("email")); company.setCompanyDetails(details); companies.add(company);
				}
				queryInstance.setError(null); queryInstance.setCompanies(companies);
			} catch (Exception e) { log.warn("Sandbox company lookup failed", e); queryInstance.setError(ErrorType.SERVICE_ERROR); queryInstance.setCompanies(Collections.<Company>emptyList()); }
			return;
		}
		`+marker);
}
// Keep the original access checks but remove all packaged remote endpoint
// defaults, including the unreachable production/test initialization branch.
const endpointPattern=/(WSUtils\.setEndpoint\(\(BindingProvider\)\s+\w+,\s*")[^"]+("\);)/g;
const endpoints=[...text.matchAll(endpointPattern)];
if(endpoints.length!==9)throw Error(`Unexpected SSBT endpoint count: ${endpoints.length}`);
text=text.replace(endpointPattern,'$1http://127.0.0.1:1/disabled$2');
fs.writeFileSync(file,text,'latin1');
