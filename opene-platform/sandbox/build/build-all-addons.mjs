// Usage: node build-all-addons.mjs <source> <webapp> <JDK> <servlet.jar> <output>
// Every project is compiled in isolation; failed classes are never packaged.
import fs from 'node:fs';
import path from 'node:path';
import {spawnSync} from 'node:child_process';
import {inflateRawSync} from 'node:zlib';
import {sanitizeWsdl} from './sanitize-wsdl.mjs';
const [src,app,jdk,servlet,out]=process.argv.slice(2).map(x=>path.resolve(x));
const walk=d=>fs.readdirSync(d,{withFileTypes:true}).flatMap(e=>e.isDirectory()?walk(path.join(d,e.name)):[path.join(d,e.name)]);
fs.mkdirSync(out,{recursive:true}); fs.mkdirSync(path.join(out,'jars'),{recursive:true});
const projects=fs.readdirSync(src).filter(p=>fs.existsSync(path.join(src,p,'src'))).sort();
// Mechanical rename of the hash enum; no authorization changes.
for(const f of walk(path.join(src,'MultiSigningHandler2','src')).filter(f=>f.endsWith('.java'))){let b=fs.readFileSync(f,'latin1');b=b.replaceAll('hash.HashAlgorithms;','hash.HashAlgorithm;').replaceAll('HashAlgorithms.SHA1','HashAlgorithm.SHA1');fs.writeFileSync(f,b,'latin1');}
for(const f of walk(path.join(src,'DocumentSignerModule','src')).filter(f=>f.endsWith('.java'))){let b=fs.readFileSync(f,'latin1');b=b.replaceAll('hash.HashAlgorithms;','hash.HashAlgorithm;').replaceAll('HashAlgorithms.SHA256','HashAlgorithm.SHA256.getValue()');fs.writeFileSync(f,b,'latin1');}
// Disable hard-coded Lantmateriet service/token endpoints in the sandbox.
// These modules are not configured until a local contract is implemented.
for(const [relative,names] of [
 ['LMSearchProvider/src/com/nordicpeak/lmsearch/LMConnection.java',['BASICAUTH_LMSERVICE_URL','OAUTH_LMSERVICE_URL']],
 ['LMSearchProvider/src/com/nordicpeak/lmsearch/utils/LMTokenHandler.java',['SERVICE_URL']],
 ['BiznodeVehicleAlternativeListProvider/src/com/nordicpeak/flowengine/biznodevehiclealternativelist/BiznodeVehicleAlternativeListProviderModule.java',['BISNODE_TOKEN_OAUTH2','BISNODE_VEHICLE_API']],
 ['BoverketIntegrationProvider/src/com/nordicpeak/flowengine/integration/boverket/BoverketIntegrationProvider.java',['apiURL']]
]){
 const file=path.join(src,...relative.split('/'));let source=fs.readFileSync(file,'latin1');
 for(const name of names){const re=new RegExp('(\\b'+name+'\\s*=\\s*")[^"]+("\\s*;)');if(!re.test(source))throw Error(`Expected ${name} endpoint constant`);source=source.replace(re,'$1http://127.0.0.1:1/disabled$2');}
 fs.writeFileSync(file,source,'latin1');
}
const built=new Set(); const results={};
// Explicit incremental retry for this development session; release builds use a clean output.
if(process.env.ADDON_RESUME==='1'&&fs.existsSync(path.join(out,'addon-report.json'))){for(const r of JSON.parse(fs.readFileSync(path.join(out,'addon-report.json')))){results[r.project]=r;if(r.status.startsWith('built')&&fs.existsSync(path.join(out,'jars',r.project+'.jar')))built.add(r.project);}}
for(let round=1;round<=projects.length;round++){
 if(process.env.ADDON_DEDUPE_ONLY==='1')break;
 let progress=0;
 for(const p of projects.filter(p=>!built.has(p))){
  const root=path.join(src,p,'src'), classes=path.join(out,'classes',p);
  fs.rmSync(classes,{recursive:true,force:true});fs.mkdirSync(classes,{recursive:true});
  const files=walk(root), excluded=p==='WeightEvaluationProviders'?['weightapi','setweightfromapievaluator','settextfieldweightfromapievaluator']:[];
  const java=files.filter(f=>f.endsWith('.java')&&!excluded.some(x=>f.split(path.sep).includes(x)));
  if(!java.length){results[p]={project:p,status:'no-java-sources',reason:'No Java source files'};continue;}
  const list=path.join(out,p+'.args');fs.writeFileSync(list,java.map(f=>'"'+path.relative(root,f).replaceAll('\\','/')+'"').join('\n'));
  const cp=[path.join(app,'WEB-INF','lib','*'),servlet,path.join(out,'jars','*'),path.join(src,p,'lib','*')].join(path.delimiter);
  const r=spawnSync(path.join(jdk,'bin','javac.exe'),['--release','8','-encoding','ISO-8859-1','-proc:none','-Xlint:-options','-cp',cp,'-d',classes,'@'+list],{cwd:root,encoding:'utf8',maxBuffer:16*1024*1024});
  fs.writeFileSync(path.join(out,p+'.log'),(r.stdout||'')+(r.stderr||'')+(r.error||''));
  if(r.status!==0){results[p]={project:p,status:'not-built',reason:(r.stderr||String(r.error)).split('\n').filter(l=>l.includes('error:')).slice(0,8).join(' | ')};continue;}
  for(const f of files.filter(f=>!f.endsWith('.java'))){const dest=path.join(classes,path.relative(root,f));fs.mkdirSync(path.dirname(dest),{recursive:true});if(/\.(wsdl|xsd)$/i.test(f)){const data=fs.readFileSync(f).toString('latin1');const safe=sanitizeWsdl(data);fs.writeFileSync(dest,safe.result,'latin1');if(safe.changed)console.log(`SANITIZED ${p}: ${safe.changed} contract endpoint/import addresses`);}else fs.copyFileSync(f,dest);}
  const jar=spawnSync(path.join(jdk,'bin','jar.exe'),['--create','--file',path.join(out,'jars',p+'.jar'),'--no-manifest','-C',classes,'.'],{encoding:'utf8'});
  if(jar.status!==0)throw Error(jar.stderr);
  built.add(p); progress++;results[p]={project:p,status:excluded.length?'built-partial-not-enabled':'built-not-enabled',excluded,reason:excluded.length?'Optional API endpoints excluded because upstream APISourceProvider changed; authorization is not bypassed':'Activation and end-to-end verification tracked separately'};
  const localLib=path.join(src,p,'lib');if(fs.existsSync(localLib))for(const f of fs.readdirSync(localLib).filter(f=>f.endsWith('.jar')))fs.copyFileSync(path.join(localLib,f),path.join(app,'WEB-INF','lib',f));
  console.log(`BUILT ${p}`);
 }
 fs.writeFileSync(path.join(out,'addon-report.json'),JSON.stringify(projects.map(p=>results[p]),null,2)+'\n');
 console.log(`ROUND ${round}: ${built.size}/${projects.length} built`);
 if(!progress)break;
}
// The archive contains projects that also exist in the pinned OpenE core.
// Do not put two versions of a class on Tomcat's unordered classpath.
function classes(file){
 const b=fs.readFileSync(file);let e=-1;
 for(let i=b.length-22;i>=Math.max(0,b.length-65557);i--){if(b.readUInt32LE(i)===0x06054b50){e=i;break;}}
 if(e<0)throw Error(`Invalid jar: ${file}`);
 let off=b.readUInt32LE(e+16), count=b.readUInt16LE(e+10), result=[];
 if(off===0xffffffff||count===0xffff)throw Error(`ZIP64 jar unsupported: ${file}`);
 for(let i=0;i<count;i++){
  if(b.readUInt32LE(off)!==0x02014b50)throw Error(`Invalid central directory: ${file}`);
  const nameLength=b.readUInt16LE(off+28),extraLength=b.readUInt16LE(off+30),commentLength=b.readUInt16LE(off+32);
  const name=b.subarray(off+46,off+46+nameLength).toString('utf8');
  if(name.endsWith('.class')&&!name.startsWith('META-INF/')&&!name.endsWith('/module-info.class')&&name!=='module-info.class')result.push(name);
  off+=46+nameLength+extraLength+commentLength;
 }
 return result;
}
const owners=new Map(),lib=path.join(app,'WEB-INF','lib');
for(const f of fs.readdirSync(lib).filter(f=>f.endsWith('.jar')).sort())for(const c of classes(path.join(lib,f)))if(!owners.has(c))owners.set(c,f);
for(const p of [...built].sort()){
 const jar=path.join(out,'jars',p+'.jar'),names=classes(jar),conflicts=names.filter(c=>owners.has(c));
 if(conflicts.length){results[p].status='built-excluded-duplicate-classes';results[p].reason=`${conflicts.length} classes duplicate existing jars: `+conflicts.slice(0,4).map(c=>`${c} (${owners.get(c)})`).join(', ');continue;}
 fs.copyFileSync(jar,path.join(lib,p+'.jar'));for(const c of names)owners.set(c,p+'.jar');
}
// Fail the build if a packaged jar still contains a remote WSDL endpoint or
// import, including WSDL files supplied by a project's bundled dependencies.
// Inspect the central directory rather than executing or loading jar code.
for(const name of fs.readdirSync(lib).filter(f=>f.endsWith('.jar'))){
 const file=path.join(lib,name),b=fs.readFileSync(file);let end=-1;
 for(let i=b.length-22;i>=Math.max(0,b.length-65557);i--){if(b.readUInt32LE(i)===0x06054b50){end=i;break;}}
 if(end<0)throw Error(`Invalid jar: ${file}`);
 let offset=b.readUInt32LE(end+16), count=b.readUInt16LE(end+10);
 if(offset===0xffffffff||count===0xffff)throw Error(`ZIP64 jar unsupported: ${file}`);
 for(let i=0;i<count;i++){
  if(b.readUInt32LE(offset)!==0x02014b50)throw Error(`Invalid central directory: ${file}`);
  const method=b.readUInt16LE(offset+10),size=b.readUInt32LE(offset+20),unpacked=b.readUInt32LE(offset+24),n=b.readUInt16LE(offset+28),x=b.readUInt16LE(offset+30),c=b.readUInt16LE(offset+32),local=b.readUInt32LE(offset+42);
  const entry=b.subarray(offset+46,offset+46+n).toString('utf8');
  if(entry.toLowerCase().endsWith('.wsdl')){
   if(unpacked>8*1024*1024 || local+30>b.length || b.readUInt32LE(local)!==0x04034b50)throw Error(`Invalid WSDL entry: ${name}/${entry}`);
   const start=local+30+b.readUInt16LE(local+26)+b.readUInt16LE(local+28);
   const compressed=b.subarray(start,start+size);
   const data=method===0?compressed:method===8?inflateRawSync(compressed,{maxOutputLength:8*1024*1024}):null;
   if(!data||data.length!==unpacked)throw Error(`Unsupported WSDL compression: ${name}/${entry}`);
   if(sanitizeWsdl(data.toString('latin1')).changed)throw Error(`Remote WSDL target remains in ${name}/${entry}`);
  }
  offset+=46+n+x+c;
 }
}
fs.writeFileSync(path.join(out,'addon-report.json'),JSON.stringify(projects.map(p=>results[p]),null,2)+'\n');
