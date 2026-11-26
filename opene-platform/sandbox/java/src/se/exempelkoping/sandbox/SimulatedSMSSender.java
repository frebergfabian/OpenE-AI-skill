package se.exempelkoping.sandbox;

import java.io.OutputStream;
import java.net.HttpURLConnection;
import java.net.Proxy;
import java.net.URI;
import java.net.URL;
import java.nio.charset.StandardCharsets;
import javax.sql.DataSource;
import com.github.openjson.JSONArray;
import com.github.openjson.JSONObject;
import se.unlogic.hierarchy.core.interfaces.SectionInterface;
import se.unlogic.hierarchy.core.interfaces.modules.descriptors.ForegroundModuleDescriptor;
import se.unlogic.hierarchy.core.interfaces.sms.SMS;
import se.unlogic.hierarchy.core.interfaces.sms.SMSSender;
import se.unlogic.hierarchy.foregroundmodules.AnnotatedForegroundModule;

/** Local, synthetic SMS transport. Never forwards to a telephone network. */
public class SimulatedSMSSender extends AnnotatedForegroundModule implements SMSSender {
 private String base;
 @Override public void init(ForegroundModuleDescriptor desc, SectionInterface section, DataSource dataSource) throws Exception {
  base=System.getProperty("opene.sandbox.simulator"); URI u=new URI(base);
  if(!"http".equals(u.getScheme())||!"127.0.0.1".equals(u.getHost())||u.getPort()<1)throw new IllegalStateException("Local simulator required");
  super.init(desc,section,dataSource);
  if(!systemInterface.getInstanceHandler().addInstance(SMSSender.class,this))throw new IllegalStateException("Another SMS sender is registered");
 }
 @Override public void unload() throws Exception {systemInterface.getInstanceHandler().removeInstance(SMSSender.class,this);super.unload();}
 @Override public boolean send(SMS sms){
  try{
   JSONArray recipients=new JSONArray();for(String r:sms.getRecipients())recipients.put(r);
   byte[] body=new JSONObject().put("sender",sms.getSenderName()).put("recipients",recipients).put("message",sms.getMessage()).toString().getBytes(StandardCharsets.UTF_8);
   HttpURLConnection c=(HttpURLConnection)new URL(base+"/sms").openConnection(Proxy.NO_PROXY);
   c.setInstanceFollowRedirects(false);c.setConnectTimeout(3000);c.setReadTimeout(8000);c.setRequestMethod("POST");c.setDoOutput(true);c.setRequestProperty("Content-Type","application/json; charset=utf-8");
   try(OutputStream o=c.getOutputStream()){o.write(body);}int status=c.getResponseCode();c.disconnect();return status>=200&&status<300;
  }catch(Exception ex){log.warn("Sandbox SMS delivery failed",ex);return false;}
 }
}
