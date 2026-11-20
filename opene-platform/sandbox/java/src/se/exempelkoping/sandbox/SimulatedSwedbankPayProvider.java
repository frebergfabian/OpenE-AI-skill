package se.exempelkoping.sandbox;

import java.io.*;
import java.net.*;
import java.math.BigDecimal;
import java.nio.charset.StandardCharsets;
import java.security.SecureRandom;
import java.sql.*;
import java.util.*;
import javax.servlet.http.*;
import javax.sql.DataSource;
import com.github.openjson.*;
import com.nordicpeak.openhierarchy.payment.interfaces.*;
import com.nordicpeak.openhierarchy.payment.beans.*;
import com.nordicpeak.openhierarchy.payment.enums.PaymentType;
import com.nordicpeak.openhierarchy.payment.events.*;
import se.unlogic.hierarchy.foregroundmodules.AnnotatedForegroundModule;
import se.unlogic.hierarchy.core.annotations.WebPublic;
import se.unlogic.hierarchy.core.beans.*;
import se.unlogic.hierarchy.core.enums.EventTarget;
import se.unlogic.hierarchy.core.interfaces.*;
import se.unlogic.hierarchy.core.interfaces.modules.descriptors.ForegroundModuleDescriptor;
import se.unlogic.standardutils.validation.*;
import se.unlogic.webutils.http.URIParser;

/** Sandbox only. Checkout 3.1 subset; endpoint is forced to the instance's loopback simulator. */
public class SimulatedSwedbankPayProvider extends AnnotatedForegroundModule implements PaymentProvider {
 private DataSource db;
 private String endpoint;
 public void init(ForegroundModuleDescriptor descriptor, SectionInterface section, DataSource ds) throws Exception {
  db=ds;endpoint=System.getProperty("opene.sandbox.simulator");
  URI u=new URI(endpoint);if(!"http".equals(u.getScheme())||!"127.0.0.1".equals(u.getHost())||u.getPort()<1)throw new IllegalStateException("Local simulator required");
  super.init(descriptor,section,ds);
  try(Connection c=db.getConnection();Statement s=c.createStatement()){s.executeUpdate("CREATE TABLE IF NOT EXISTS sandbox_payment_orders(orderID VARCHAR(200) PRIMARY KEY,nonce VARCHAR(64) NOT NULL,sourceType VARCHAR(50),sourceID INTEGER,invoiceID VARCHAR(100),amount DECIMAL(18,2),userID INTEGER,delivered BOOLEAN NOT NULL DEFAULT 0)");}
  if(!systemInterface.getInstanceHandler().addInstance(PaymentProvider.class,this))throw new IllegalStateException("Payment provider already registered");
 }
 public void unload() throws Exception{systemInterface.getInstanceHandler().removeInstance(PaymentProvider.class,this);super.unload();}
 public String getProviderName(){return "Simulerad betalning";}
 public String getCurrencyCode(){return "SEK";}
 private static String nonce(){byte[] b=new byte[32];new SecureRandom().nextBytes(b);StringBuilder s=new StringBuilder();for(byte x:b)s.append(String.format("%02x",x));return s.toString();}
 private static String esc(String s){return s.replace("&","&amp;").replace("\"","&quot;").replace("<","&lt;").replace(">","&gt;");}
 private JSONObject request(String method,String path,JSONObject body)throws Exception{
  if(!path.startsWith("/psp/paymentorders")||path.contains(".."))throw new IOException("Invalid payment resource");
  HttpURLConnection c=(HttpURLConnection)new URL(endpoint+path).openConnection(Proxy.NO_PROXY);c.setInstanceFollowRedirects(false);c.setConnectTimeout(5000);c.setReadTimeout(10000);c.setRequestMethod(method);c.setRequestProperty("Accept","application/json;version=3.1");
  if(body!=null){c.setDoOutput(true);c.setRequestProperty("Content-Type","application/json;version=3.1");try(OutputStream o=c.getOutputStream()){o.write(body.toString().getBytes(StandardCharsets.UTF_8));}}
  try{if(c.getResponseCode()<200||c.getResponseCode()>299)throw new IOException("Simulator HTTP "+c.getResponseCode());ByteArrayOutputStream out=new ByteArrayOutputStream();try(InputStream in=c.getInputStream()){byte[] b=new byte[4096];int n;while((n=in.read(b))!=-1){if(out.size()+n>1048576)throw new IOException("Response too large");out.write(b,0,n);}}return new JSONObject(new String(out.toByteArray(),StandardCharsets.UTF_8));}finally{c.disconnect();}
 }
 public ViewFragment showPaymentForm(String config,Invoice invoice,HttpServletRequest req,User user,URIParser parser,PaymentCallback callback,List<ValidationError> errors)throws Exception{
  String key=nonce();String origin="http://127.0.0.1:"+req.getLocalPort();
  String returnURL=callback.getProcessPaymentURL();if(!returnURL.startsWith("http"))returnURL=origin+returnURL;
  URI ret=new URI(returnURL);if(!"127.0.0.1".equals(ret.getHost())||ret.getPort()!=req.getLocalPort())throw new IOException("External return URL refused");
  JSONObject urls=new JSONObject().put("completeUrl",returnURL).put("cancelUrl",returnURL).put("callbackUrl",origin+getFullAlias()+"/callback?key="+key).put("hostUrls",new JSONArray().put(origin));
  JSONObject order=new JSONObject().put("operation","Purchase").put("currency","SEK").put("amount",invoice.getTotalSum().movePointRight(2).longValueExact()).put("vatAmount",0).put("description","Syntetisk kommunal ansokan").put("language","sv-SE").put("urls",urls).put("payeeInfo",new JSONObject().put("payeeId","sandbox.invalid").put("payeeReference",invoice.getInvoiceID()));
  JSONObject response=request("POST","/psp/paymentorders",new JSONObject().put("paymentOrder",order));String id=response.getJSONObject("paymentOrder").getString("id");
  try(Connection c=db.getConnection();PreparedStatement p=c.prepareStatement("INSERT INTO sandbox_payment_orders(orderID,nonce,sourceType,sourceID,invoiceID,amount,userID) VALUES(?,?,?,?,?,?,?)")){p.setString(1,id);p.setString(2,key);p.setString(3,invoice.getSourceType());p.setInt(4,invoice.getSourceID());p.setString(5,invoice.getInvoiceID());p.setBigDecimal(6,invoice.getTotalSum());if(user==null)p.setNull(7,Types.INTEGER);else p.setInt(7,user.getUserID());p.executeUpdate();}
  JSONArray ops=response.getJSONArray("operations");String redirect=null;for(int n=0;n<ops.length();n++){JSONObject op=ops.getJSONObject(n);if("redirect-checkout".equals(op.getString("rel")))redirect=op.getString("href");}
  if(redirect==null||!redirect.startsWith(endpoint+"/checkout/"))throw new IOException("Invalid checkout URL");
  req.getSession().setAttribute("sandbox.order."+invoice.getInvoiceID(),id);
  return new SimpleViewFragment("<div class=\"sandbox-payment\"><h2>Simulerad betalning</h2><p>"+esc(invoice.getTotalSum().toPlainString())+" SEK. Endast syntetiskt test.</p><a href=\""+esc(redirect)+"\">Oppna simulerad betalsida</a></div>");
 }
 public PaymentResponse processPayment(String config,Invoice invoice,HttpServletRequest req,HttpServletResponse res,User user,URIParser parser,PaymentCallback callback)throws Exception{
  Object saved=req.getSession().getAttribute("sandbox.order."+invoice.getInvoiceID());if(!(saved instanceof String))return null;
  JSONObject order=request("GET",(String)saved,null).getJSONObject("paymentOrder");
  if(!"Paid".equals(order.getString("status")))return null;
  if(order.getLong("amount")!=invoice.getTotalSum().movePointRight(2).longValueExact()||!"SEK".equals(order.getString("currency")))throw new IOException("Payment amount/currency mismatch");
  // Async callback is authoritative. Redirect to persisted case after callback delivery.
  res.sendRedirect("/minasidor/overview/"+invoice.getSourceID());return null;
 }
 @WebPublic public synchronized ForegroundModuleResponse callback(HttpServletRequest req,HttpServletResponse res,User ignored,URIParser parser)throws Exception{
  if(!"POST".equals(req.getMethod())||!"127.0.0.1".equals(req.getRemoteAddr())){res.sendError(403);return null;}
  String key=req.getParameter("key");if(key==null||!key.matches("[a-f0-9]{64}")){res.sendError(403);return null;}
  try(Connection c=db.getConnection();PreparedStatement p=c.prepareStatement("SELECT * FROM sandbox_payment_orders WHERE nonce=?")){p.setString(1,key);try(ResultSet r=p.executeQuery()){
   if(!r.next()){res.sendError(404);return null;}if(r.getBoolean("delivered")){res.setStatus(200);return null;}
   String id=r.getString("orderID");JSONObject order=request("GET",id,null).getJSONObject("paymentOrder");if(!"Paid".equals(order.getString("status"))){res.setStatus(200);return null;}
   BigDecimal amount=r.getBigDecimal("amount");if(order.getLong("amount")!=amount.movePointRight(2).longValueExact()||!"SEK".equals(order.getString("currency"))){res.sendError(409);return null;}
   Invoice invoice=new SimpleInvoice(r.getString("sourceType"),r.getInt("sourceID"),r.getString("invoiceID"),"Simulerad betalning","Syntetiskt test",amount);
   int uid=r.getInt("userID");User user=r.wasNull()?null:systemInterface.getUserHandler().getUser(uid,true,true);
   PaymentResponse result=new SimplePaymentResponse(PaymentType.DIRECT,"SEK",amount,"Simulerad betalning",BigDecimal.ZERO,Collections.singletonMap("sandboxPaymentOrder",id));
   systemInterface.getEventHandler().sendEvent(PaymentProvider.class,new DelayedPaymentEvent(invoice,result,user,new Timestamp(System.currentTimeMillis())),EventTarget.ALL);
   try(PreparedStatement done=c.prepareStatement("UPDATE sandbox_payment_orders SET delivered=1 WHERE orderID=?")){done.setString(1,id);done.executeUpdate();}
  }}res.setStatus(200);return null;
 }
 public ViewFragment showConfiguration(String c,HttpServletRequest r,User u,URIParser p){return new SimpleViewFragment("<p>Simulerad betalning via lokal Checkout 3.1-simulator.</p>");}
 public ViewFragment showConfigurationForm(String c,HttpServletRequest r,User u,URIParser p,ValidationException v){return showConfiguration(c,r,u,p);}
 public void processPaymentConfiguration(String c,HttpServletRequest r,User u,URIParser p){}
 public void deletePaymentConfiguration(String c){}
}
