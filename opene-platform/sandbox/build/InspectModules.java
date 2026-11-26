import java.io.*;
import java.lang.reflect.*;
import java.util.*;
import java.util.jar.*;
import com.github.openjson.*;
import se.unlogic.hierarchy.core.interfaces.modules.ForegroundModule;
import se.unlogic.hierarchy.core.interfaces.modules.BackgroundModule;
public class InspectModules {
 public static void main(String[] args)throws Exception{
  JSONArray rows=new JSONArray();
  for(String file:args){try(JarFile jar=new JarFile(file)){Enumeration<JarEntry> entries=jar.entries();while(entries.hasMoreElements()){
   String n=entries.nextElement().getName();if(!n.endsWith(".class")||n.contains("$"))continue;
   String name=n.substring(0,n.length()-6).replace('/','.');
   try{Class<?> c=Class.forName(name,false,InspectModules.class.getClassLoader());if(c.isInterface()||Modifier.isAbstract(c.getModifiers())||(!ForegroundModule.class.isAssignableFrom(c)&&!BackgroundModule.class.isAssignableFrom(c)))continue;
    JSONArray bases=new JSONArray();for(Class<?> b=c;b!=null;b=b.getSuperclass())bases.put(b.getName());
    String folder=n.substring(0,n.lastIndexOf('/')+1);JSONArray xsls=new JSONArray();Enumeration<JarEntry> all=jar.entries();while(all.hasMoreElements()){String x=all.nextElement().getName();if(x.startsWith(folder)&&x.endsWith(".sv.xsl")&&x.substring(folder.length()).indexOf('/')<0)xsls.put(x.substring(folder.length()));}
    rows.put(new JSONObject().put("project",new File(file).getName().replace(".jar","")).put("class",name).put("foreground",ForegroundModule.class.isAssignableFrom(c)).put("bases",bases).put("xsl",xsls).put("staticContent",jar.getEntry(folder+"staticcontent/")!=null));
   }catch(LinkageError e){rows.put(new JSONObject().put("class",name).put("loadError",e.toString()));}
  }}}
  System.out.println(rows.toString(2));
 }
}
