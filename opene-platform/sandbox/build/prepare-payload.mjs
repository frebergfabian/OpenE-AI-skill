import fs from 'node:fs';
import path from 'node:path';
import {spawnSync} from 'node:child_process';
// Explicit inputs; never copy a running instance or its database/configuration.
const [dist,runtime,payload,portable]=process.argv.slice(2).map(x=>path.resolve(x));
const sb=path.dirname(portable);
for(const [from,to] of [[path.join(dist,'webapp'),'webapp'],[path.join(runtime,'mariadb-11.8.10-winx64'),'mariadb'],[path.join(runtime,'apache-tomcat-9.0.123'),'tomcat']]){
 fs.cpSync(from,path.join(payload,to),{recursive:true,filter:p=>!/(?:^|[\\/])(?:logs|work|temp|webapps)(?:[\\/]|$)/.test(path.relative(from,p))&&!/(?:config\.xml|systemlog\.log|.*\.log)$/.test(p)});
}
fs.mkdirSync(path.join(payload,'sql'),{recursive:true});
let dump=fs.readFileSync(path.join(dist,'sql','10-demo.sql'),'utf8');
// Do not distribute upstream demo credentials or account data in the executable.
dump=dump.replace(/^INSERT INTO `simple_users`[^\r\n]*\r?\n/gm,'').replace(/^INSERT INTO `simple_user_groups`[^\r\n]*\r?\n/gm,'');
if(dump.includes('6e2b592b196249838b312cfa026ebe2146dfa6dc')||/INSERT INTO `simple_users`/.test(dump))throw Error('Demo account scrub failed');
fs.writeFileSync(path.join(payload,'sql','10-demo.sql'),dump);
fs.copyFileSync(path.join(sb,'docker','sql','20-sandbox-pre.sql'),path.join(payload,'sql','20-sandbox-pre.sql'));
fs.copyFileSync(path.join(sb,'docker','config.xml.tmpl'),path.join(payload,'config.xml.tmpl'));
fs.copyFileSync(path.join(dist,'BUILDINFO'),path.join(payload,'BUILDINFO'));
const r=spawnSync('C:/Program Files/7-Zip/7z.exe',['a','-tzip','-mx=1',path.join(portable,'payload.zip'),'.'],{cwd:payload,encoding:'utf8'});
if(r.status!==0)throw Error(r.stderr);console.log(r.stdout);
