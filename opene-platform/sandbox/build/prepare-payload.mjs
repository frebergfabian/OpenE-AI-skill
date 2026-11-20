import fs from 'node:fs';
import path from 'node:path';
import {spawnSync} from 'node:child_process';
// Explicit inputs; never copy a running instance or its database/configuration.
const [dist,runtime,payload,portable]=process.argv.slice(2).map(x=>path.resolve(x));
// When portable sources are staged outside the checkout, pass the sandbox root.
const sb=process.argv[6]?path.resolve(process.argv[6]):path.dirname(portable);
for(const [from,to] of [[path.join(dist,'webapp'),'webapp'],[path.join(runtime,'mariadb-11.8.10-winx64'),'mariadb'],[path.join(runtime,'apache-tomcat-9.0.123'),'tomcat']]){
 fs.cpSync(from,path.join(payload,to),{recursive:true,filter:p=>!/(?:^|[\\/])(?:logs|work|temp|webapps)(?:[\\/]|$)/.test(path.relative(from,p))&&!/(?:config\.xml|systemlog\.log|.*\.log)$/.test(p)});
}
fs.mkdirSync(path.join(payload,'sql'),{recursive:true});
// The demo theme otherwise loads versioned copies of jQuery repeatedly. This
// destroys plugins registered by earlier scripts, including the calendar.
const theme=path.join(payload,'webapp','WEB-INF','xsl','design.xsl');
let themeBytes=fs.readFileSync(theme).toString('latin1');
const jquery='<script type="text/javascript" src="{/document/requestinfo/contextpath}/static/global/jquery/jquery.js"/>';
if(!themeBytes.includes(jquery))throw Error('Unexpected demo theme');
themeBytes=themeBytes.replace(jquery,jquery+'\n<script type="text/javascript" src="{/document/requestinfo/contextpath}/static/global/jquery/jquery-migrate.js"/>');
themeBytes=themeBytes.replace("scripts/script[src != '/static/global/jquery/jquery.js']","scripts/script[not(starts-with(src, '/static/global/jquery/jquery.js')) and not(starts-with(src, '/static/global/jquery/jquery-migrate.js'))]");
fs.writeFileSync(theme,themeBytes,'latin1');
// Modern jQuery no longer provides $.browser; keep the legacy theme's IE
// detection harmless in modern browsers without changing other feature tests.
const modernizrInit=path.join(payload,'webapp','js','init-modernizr.js');
let modernizrSource=fs.readFileSync(modernizrInit,'latin1');
if(!modernizrSource.includes('$.browser.msie'))throw Error('Unexpected demo browser detection');
modernizrSource=modernizrSource.replace('$.browser.msie','($.browser && $.browser.msie)');
fs.writeFileSync(modernizrInit,modernizrSource,'latin1');
let dump=fs.readFileSync(path.join(dist,'sql','10-demo.sql'),'utf8');
// Do not distribute upstream demo credentials or account data in the executable.
dump=dump.replace(/^INSERT INTO `simple_users`[^\r\n]*\r?\n/gm,'').replace(/^INSERT INTO `simple_user_groups`[^\r\n]*\r?\n/gm,'');
if(dump.includes('6e2b592b196249838b312cfa026ebe2146dfa6dc')||/INSERT INTO `simple_users`/.test(dump))throw Error('Demo account scrub failed');
fs.writeFileSync(path.join(payload,'sql','10-demo.sql'),dump);
fs.copyFileSync(path.join(sb,'docker','sql','20-sandbox-pre.sql'),path.join(payload,'sql','20-sandbox-pre.sql'));
if(process.env.OPENE_EXPERIMENTAL_ADDONS==='1')fs.copyFileSync(path.join(sb,'experimental','25-addons.sql'),path.join(payload,'sql','25-addons.sql'));
else fs.rmSync(path.join(payload,'sql','25-addons.sql'),{force:true});
if(process.env.OPENE_EXPERIMENTAL_ADDONS==='1'){
 fs.copyFileSync(path.join(sb,'experimental','30-fixtures.sql'),path.join(payload,'sql','30-fixtures.sql'));
 fs.mkdirSync(path.join(payload,'fixtures'),{recursive:true});
 fs.copyFileSync(path.join(sb,'..','assets','funktionslabbet.oeflow'),path.join(payload,'fixtures','funktionslabbet.oeflow'));
 fs.copyFileSync(path.join(sb,'..','assets','sandbox-fastighet.oeflow'),path.join(payload,'fixtures','sandbox-fastighet.oeflow'));
 fs.copyFileSync(path.join(sb,'..','assets','sandbox-barn.oeflow'),path.join(payload,'fixtures','sandbox-barn.oeflow'));
 fs.copyFileSync(path.join(sb,'..','assets','sandbox-betalning.oeflow'),path.join(payload,'fixtures','sandbox-betalning.oeflow'));
 fs.copyFileSync(path.join(sb,'..','assets','sandbox-foretag.oeflow'),path.join(payload,'fixtures','sandbox-foretag.oeflow'));
 fs.copyFileSync(path.join(sb,'..','assets','sandbox-flerpart.oeflow'),path.join(payload,'fixtures','sandbox-flerpart.oeflow'));
}else{
 fs.rmSync(path.join(payload,'sql','30-fixtures.sql'),{force:true});
 fs.rmSync(path.join(payload,'fixtures'),{force:true,recursive:true});
}
fs.copyFileSync(path.join(sb,'docker','config.xml.tmpl'),path.join(payload,'config.xml.tmpl'));
fs.copyFileSync(path.join(dist,'BUILDINFO'),path.join(payload,'BUILDINFO'));
fs.rmSync(path.join(portable,'payload.zip'),{force:true});
const r=spawnSync(process.env.OPENE_7ZIP||'C:/Program Files/7-Zip/7z.exe',['a','-tzip','-mx=1',path.join(portable,'payload.zip'),'.'],{cwd:payload,encoding:'utf8'});
if(r.status!==0)throw Error(r.stderr);console.log(r.stdout);
