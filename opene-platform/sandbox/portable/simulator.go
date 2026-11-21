package main

import (
	"bufio"
	"bytes"
	"encoding/json"
	"encoding/xml"
	"fmt"
	"html/template"
	"io"
	"net"
	"net/http"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"sync"
	"time"
)

type Call struct {
	Time     string `json:"time"`
	Method   string `json:"method"`
	Path     string `json:"path"`
	Scenario string `json:"scenario"`
	Body     string `json:"body"`
}

var syntheticRecipient = regexp.MustCompile(`(?i)^<[^<>@\s]+@(sandbox\.invalid|exempelkoping\.example\.com)>$`)

type Simulator struct {
	appPort  int
	dir      string
	port     int
	mu       sync.Mutex
	scenario map[string]string
}

func newSimulator(dir string, port int) *Simulator {
	s := &Simulator{dir: dir, port: port, scenario: map[string]string{}}
	_ = readJSON(filepath.Join(dir, "scenarios.json"), &s.scenario)
	return s
}
func (s *Simulator) record(method, path, scenario, body string) {
	s.mu.Lock()
	defer s.mu.Unlock()
	f, e := os.OpenFile(filepath.Join(s.dir, "calls.jsonl"), os.O_CREATE|os.O_APPEND|os.O_WRONLY, 0600)
	if e != nil {
		return
	}
	defer f.Close()
	_ = json.NewEncoder(f).Encode(Call{time.Now().UTC().Format(time.RFC3339Nano), method, path, scenario, body})
}
func jsonResponse(w http.ResponseWriter, v any) {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	json.NewEncoder(w).Encode(v)
}
func (s *Simulator) serve(l net.Listener) {
	server := &http.Server{Handler: http.HandlerFunc(s.handle), ReadHeaderTimeout: 5 * time.Second}
	_ = server.Serve(l)
}
func (s *Simulator) handle(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("X-Content-Type-Options", "nosniff")
	w.Header().Set("Content-Security-Policy", "default-src 'none'; style-src 'unsafe-inline'; form-action 'self'; frame-ancestors 'none'")
	if r.Host != fmt.Sprintf("127.0.0.1:%d", s.port) {
		http.Error(w, "Unexpected Host", 403)
		return
	}
	if origin := r.Header.Get("Origin"); origin != "" && origin != localURL(s.port) {
		http.Error(w, "Cross-origin request blocked", 403)
		return
	}
	if r.Header.Get("Sec-Fetch-Site") == "cross-site" {
		http.Error(w, "Cross-site request blocked", 403)
		return
	}
	if strings.HasPrefix(r.URL.Path, "/psp/paymentorders") || strings.HasPrefix(r.URL.Path, "/checkout/") {
		s.payment(w, r)
		return
	}
	switch r.URL.Path {
	case "/health":
		jsonResponse(w, map[string]any{"healthy": true, "syntheticOnly": true})
		return
	case "/api/calls":
		b, _ := os.ReadFile(filepath.Join(s.dir, "calls.jsonl"))
		w.Header().Set("Content-Type", "application/x-ndjson")
		w.Write(b)
		return
	case "/api/mail":
		files, _ := filepath.Glob(filepath.Join(s.dir, "mail", "*.eml"))
		rows := []map[string]string{}
		for _, p := range files {
			b, _ := os.ReadFile(p)
			rows = append(rows, map[string]string{"id": filepath.Base(p), "message": string(b)})
		}
		jsonResponse(w, rows)
		return
	case "/scenarios":
		if r.Method != "POST" {
			http.Error(w, "POST required", 405)
			return
		}
		r.Body = http.MaxBytesReader(w, r.Body, 4096)
		if e := r.ParseForm(); e != nil {
			http.Error(w, "Invalid form", 400)
			return
		}
		module := r.FormValue("module")
		scenario := r.FormValue("scenario")
		if (module != "pud" && module != "http" && module != "payment" && module != "sms" && module != "ssbt" && module != "artvise" && module != "easit") || (scenario != "success" && scenario != "denied" && scenario != "timeout" && scenario != "error") {
			http.Error(w, "Unknown module or scenario", 400)
			return
		}
		s.mu.Lock()
		s.scenario[module] = scenario
		e := writeJSON(filepath.Join(s.dir, "scenarios.json"), s.scenario)
		s.mu.Unlock()
		if e != nil {
			http.Error(w, "Could not persist scenario", 500)
			return
		}
		http.Redirect(w, r, "/", 303)
		return
	case "/", "/mail":
		s.page(w, r)
		return
	}
	module := ""
	if r.URL.Path == "/search/registerenheter" || r.URL.Path == "/search/addresses" {
		module = "pud"
	}
	if r.URL.Path == "/integration/http" {
		module = "http"
	}
	if r.URL.Path == "/sms" {
		module = "sms"
	}
	if r.URL.Path == "/ssbt/en" || r.URL.Path == "/ssbt/gu" || r.URL.Path == "/ssbt/ro" {
		module = "ssbt"
	}
	if r.URL.Path == "/integration/artvise" {
		module = "artvise"
	}
	if r.URL.Path == "/integration/easit" {
		module = "easit"
	}
	if module == "" {
		http.Error(w, "No verified simulator contract for this route", 501)
		return
	}
	s.mu.Lock()
	scenario := s.scenario[module]
	s.mu.Unlock()
	if q := r.Header.Get("X-Sandbox-Scenario"); q != "" {
		scenario = q
	}
	if scenario == "" {
		scenario = "success"
	}
	r.Body = http.MaxBytesReader(w, r.Body, 8*1024*1024)
	b, e := io.ReadAll(r.Body)
	if e != nil {
		http.Error(w, "Body too large", 413)
		return
	}
	if module == "sms" {
		var m struct {
			Recipients []string `json:"recipients"`
		}
		if r.Method != "POST" || json.Unmarshal(b, &m) != nil || len(m.Recipients) == 0 {
			http.Error(w, "Invalid SMS", 400)
			return
		}
		for _, number := range m.Recipients {
			if number != "0700000000" && number != "+46700000000" {
				http.Error(w, "Only the synthetic test number is accepted", 403)
				return
			}
		}
	}
	ssbtID := ""
	ssbtDetails := map[string]bool{}
	if module == "ssbt" {
		if r.Method != "POST" || r.URL.Path == "/ssbt/ro" {
			http.Error(w, "Unsupported synthetic SSBT operation", 501)
			return
		}
		var envelope struct {
			XMLName xml.Name `xml:"http://schemas.xmlsoap.org/soap/envelope/ Envelope"`
			Body    struct {
				EN *struct{} `xml:"http://schema.bolagsverket.se/ssbten/engagemang EngagemangBegaran"`
				GU *struct{} `xml:"http://schema.bolagsverket.se/ssbtgu/v2/grundlaggandeuppgifter GrundlaggandeUppgifterBegaran"`
			} `xml:"http://schemas.xmlsoap.org/soap/envelope/ Body"`
		}
		if xml.Unmarshal(b, &envelope) != nil || (r.URL.Path == "/ssbt/en" && envelope.Body.EN == nil) || (r.URL.Path == "/ssbt/gu" && envelope.Body.GU == nil) {
			http.Error(w, "Invalid SSBT SOAP operation", 400)
			return
		}
		allowed := map[string]bool{"190001010008": true, "190001010016": true, "190001010024": true, "190001010032": true, "190001010040": true}
		decoder := xml.NewDecoder(bytes.NewReader(b))
		for {
			tok, e := decoder.Token()
			if e == io.EOF {
				break
			}
			if e != nil {
				http.Error(w, "Invalid XML", 400)
				return
			}
			start, ok := tok.(xml.StartElement)
			if !ok {
				continue
			}
			switch start.Name.Local {
			case "Personnummer", "Organisationsnummer", "ForetagsformKod", "GrundlaggandeUppgiftId":
				var value string
				if decoder.DecodeElement(&value, &start) != nil {
					http.Error(w, "Invalid XML value", 400)
					return
				}
				switch start.Name.Local {
				case "Personnummer":
					value = strings.ReplaceAll(strings.TrimSpace(value), "-", "")
					if !allowed[value] {
						http.Error(w, "Only sandbox identities are supported", 403)
						return
					}
					ssbtID = value
				case "Organisationsnummer":
					if strings.TrimSpace(value) != "0000000000" {
						http.Error(w, "Only sandbox organizations are supported", 403)
						return
					}
				case "ForetagsformKod":
					if strings.TrimSpace(value) != "AB" {
						http.Error(w, "Company type not simulated", 501)
						return
					}
				case "GrundlaggandeUppgiftId":
					value = strings.TrimSpace(value)
					if value != "UD0001" && value != "UD0003" && value != "UD0004" && value != "UD0006" {
						http.Error(w, "Requested detail not simulated", 501)
						return
					}
					ssbtDetails[value] = true
				}
			}
		}
		if ssbtID == "" {
			http.Error(w, "Sandbox identity required", 403)
			return
		}
		if r.URL.Path == "/ssbt/gu" && !ssbtDetails["UD0001"] {
			http.Error(w, "Required company name not requested", 501)
			return
		}
	}
	var easitUIDs []string
	if module == "easit" {
		var envelope struct {
			XMLName xml.Name `xml:"http://schemas.xmlsoap.org/soap/envelope/ Envelope"`
			Body    struct {
				Import *struct {
					Items []struct {
						UID string `xml:"UID,attr"`
					} `xml:"http://www.easit.com/bps/schemas ItemToImport"`
				} `xml:"http://www.easit.com/bps/schemas ImportItemsRequest"`
			} `xml:"http://schemas.xmlsoap.org/soap/envelope/ Body"`
		}
		if r.Method != "POST" || xml.Unmarshal(b, &envelope) != nil || envelope.Body.Import == nil || len(envelope.Body.Import.Items) == 0 || len(envelope.Body.Import.Items) > 100 {
			http.Error(w, "Only synthetic ImportItems SOAP requests are supported", 400)
			return
		}
		for _, item := range envelope.Body.Import.Items {
			if !regexp.MustCompile(`^[a-zA-Z0-9-]{1,64}$`).MatchString(item.UID) {
				http.Error(w, "Invalid item identifier", 400)
				return
			}
			easitUIDs = append(easitUIDs, item.UID)
		}
	}
	if module == "artvise" {
		var envelope struct {
			XMLName xml.Name `xml:"http://schemas.xmlsoap.org/soap/envelope/ Envelope"`
			Body    struct {
				Create *struct{} `xml:"http://www.artvise.se/MercuryAPI/ CreateCase"`
			} `xml:"http://schemas.xmlsoap.org/soap/envelope/ Body"`
		}
		if r.Method != "POST" || xml.Unmarshal(b, &envelope) != nil || envelope.Body.Create == nil {
			http.Error(w, "Only the synthetic CreateCase SOAP operation is supported", 400)
			return
		}
	}
	s.record(r.Method, r.URL.RequestURI(), scenario, string(b))
	switch scenario {
	case "timeout":
		select {
		case <-r.Context().Done():
		case <-time.After(40 * time.Second):
		}
		return
	case "error":
		if module == "artvise" || module == "easit" || module == "ssbt" {
			w.Header().Set("Content-Type", "text/xml; charset=utf-8")
			w.WriteHeader(500)
			io.WriteString(w, `<s:Envelope xmlns:s="http://schemas.xmlsoap.org/soap/envelope/"><s:Body><s:Fault><faultcode>s:Server</faultcode><faultstring>Simulated backend error</faultstring></s:Fault></s:Body></s:Envelope>`)
			return
		}
		http.Error(w, "Simulated backend failure", 500)
		return
	case "denied":
		if module == "ssbt" {
			w.Header().Set("Content-Type", "text/xml; charset=utf-8")
			w.WriteHeader(500)
			io.WriteString(w, `<s:Envelope xmlns:s="http://schemas.xmlsoap.org/soap/envelope/"><s:Body><s:Fault><faultcode>s:Client</faultcode><faultstring>Simulated denial</faultstring></s:Fault></s:Body></s:Envelope>`)
			return
		}
		if module == "easit" {
			w.Header().Set("Content-Type", "text/xml; charset=utf-8")
			io.WriteString(w, `<s:Envelope xmlns:s="http://schemas.xmlsoap.org/soap/envelope/"><s:Body><ImportItemsResponse xmlns="http://www.easit.com/bps/schemas">`)
			for _, uid := range easitUIDs {
				fmt.Fprintf(w, `<ImportItemResult UID="%s" result="Skipped"><SkippedOrExceptionMessage>Simulated denial</SkippedOrExceptionMessage></ImportItemResult>`, uid)
			}
			io.WriteString(w, `</ImportItemsResponse></s:Body></s:Envelope>`)
			return
		}
		if module == "artvise" {
			w.Header().Set("Content-Type", "text/xml; charset=utf-8")
			io.WriteString(w, `<s:Envelope xmlns:s="http://schemas.xmlsoap.org/soap/envelope/"><s:Body><CreateCaseResponse xmlns="http://www.artvise.se/MercuryAPI/"><CreateCaseResult xmlns:a="http://schemas.datacontract.org/2004/07/Artvise.Mercury.Base.DataContracts.API"><a:Message>Simulated denial</a:Message><a:Ok>false</a:Ok></CreateCaseResult></CreateCaseResponse></s:Body></s:Envelope>`)
			return
		}
		http.Error(w, "Simulated denial", 403)
		return
	case "success":
	default:
		http.Error(w, "Unknown scenario", 400)
		return
	}
	if module == "ssbt" {
		body := ssbtENResponse
		if r.URL.Path == "/ssbt/gu" {
			body = ssbtGUResponse
			extra := ""
			if ssbtDetails["UD0003"] {
				extra += `<ns2:UD0003><ns3:Postadress><ns3:Utdelningsadress1>Testgatan 1</ns3:Utdelningsadress1><ns3:Postnummer>00000</ns3:Postnummer><ns3:Postort>Exempelkoping</ns3:Postort></ns3:Postadress></ns2:UD0003>`
			}
			if ssbtDetails["UD0004"] {
				extra += `<ns2:UD0004><ns3:EPostAdress>referens@exempelkoping.example.com</ns3:EPostAdress></ns2:UD0004>`
			}
			if ssbtDetails["UD0006"] {
				extra += `<ns2:UD0006><ns3:Telefonnummer>0700000000</ns3:Telefonnummer></ns2:UD0006>`
			}
			body = strings.Replace(body, "</ns2:GrundlaggandeUppgifterSvarDetaljer>", extra+"</ns2:GrundlaggandeUppgifterSvarDetaljer>", 1)
		}
		body = strings.ReplaceAll(body, "190001010040", ssbtID)
		w.Header().Set("Content-Type", "text/xml; charset=utf-8")
		io.WriteString(w, `<s:Envelope xmlns:s="http://schemas.xmlsoap.org/soap/envelope/"><s:Body>`+body+`</s:Body></s:Envelope>`)
		return
	}
	if module == "artvise" {
		id := token(16)
		objectID := time.Now().UnixNano()%2000000000 + 1
		w.Header().Set("Content-Type", "text/xml; charset=utf-8")
		fmt.Fprintf(w, `<s:Envelope xmlns:s="http://schemas.xmlsoap.org/soap/envelope/"><s:Body><CreateCaseResponse xmlns="http://www.artvise.se/MercuryAPI/"><CreateCaseResult xmlns:a="http://schemas.datacontract.org/2004/07/Artvise.Mercury.Base.DataContracts.API"><a:Message>Simulated</a:Message><a:ObjectInformation><a:Identifier>SIM-%s</a:Identifier><a:ObjectID>%d</a:ObjectID></a:ObjectInformation><a:Ok>true</a:Ok></CreateCaseResult></CreateCaseResponse></s:Body></s:Envelope>`, id, objectID)
		return
	}
	if module == "easit" {
		w.Header().Set("Content-Type", "text/xml; charset=utf-8")
		io.WriteString(w, `<s:Envelope xmlns:s="http://schemas.xmlsoap.org/soap/envelope/"><s:Body><ImportItemsResponse xmlns="http://www.easit.com/bps/schemas">`)
		for _, uid := range easitUIDs {
			fmt.Fprintf(w, `<ImportItemResult UID="%s" result="OK"><ReturnValues><ReturnValue name="id">EASIT-%s</ReturnValue><ReturnValue name="item">%s</ReturnValue></ReturnValues></ImportItemResult>`, uid, token(8), uid)
		}
		io.WriteString(w, `</ImportItemsResponse></s:Body></s:Envelope>`)
		return
	}
	if module == "http" {
		if r.Method != "POST" {
			http.Error(w, "POST required", 405)
			return
		}
		w.Header().Set("Content-Type", "application/xml")
		w.Write([]byte(`<response><status>accepted</status><externalID>SYNTHETIC-001</externalID></response>`))
		return
	}
	if module == "sms" {
		jsonResponse(w, map[string]any{"accepted": true, "synthetic": true})
		return
	}
	if r.Method != "GET" {
		http.Error(w, "GET required", 405)
		return
	}
	name := "Exempelköping Testfastighet 1:1"
	if strings.HasSuffix(r.URL.Path, "addresses") {
		name = "Testgatan 1, Exempelköping"
	}
	id := "00000000-0000-4000-8000-000000000001"
	q := r.URL.Query().Get("q")
	if q == "" {
		q = r.URL.Query().Get("fnr")
	}
	rows := []any{}
	if q == id || (strings.TrimSpace(q) != "" && strings.Contains(strings.ToLower(name), strings.ToLower(q))) {
		rows = append(rows, map[string]any{"properties": map[string]string{"name": name, "objid": id}})
	}
	jsonResponse(w, rows)
}

var simPage = template.Must(template.New("page").Parse(`<!doctype html><html lang="sv"><meta charset="utf-8"><meta name="viewport" content="width=device-width"><title>OpenE Simulator</title><style>body{font:17px system-ui;max-width:1000px;margin:40px auto;padding:0 24px;background:#f3f6f8;color:#132c40}a{color:#056184}section{background:white;padding:24px;border-radius:12px;margin:20px 0}pre{white-space:pre-wrap;overflow-wrap:anywhere;font-size:13px}select,button{font:inherit;padding:8px}nav{display:flex;gap:24px}</style><h1>OpenE Simulator</h1><p>Exempelköpings kommun · Endast syntetiska testuppgifter</p><nav><a href="/">Anrop och scenarier</a><a href="/mail">E-postfångst</a><a href="/api/calls">JSON-logg</a></nav><section><h2>Scenarier</h2><form method="post" action="/scenarios"><select name="module"><option value="pud">Fastighet och adress</option><option value="http">HTTPIntegrationProvider</option><option value="artvise">Artvise ärende</option><option value="easit">Easit ärende</option><option value="payment">Betalning</option><option value="sms">SMS</option><option value="ssbt">Företagsuppslag</option></select> <select name="scenario"><option value="success">Lyckat</option><option value="denied">Nekat</option><option value="timeout">Timeout</option><option value="error">Felsvar</option></select> <button>Spara</button></form><p>Valet gäller efterföljande anrop. Ej implementerade integrationer svarar 501.</p><pre>{{.Scenarios}}</pre></section><section><h2>{{.Title}}</h2><pre>{{.Content}}</pre></section></html>`))

func (s *Simulator) page(w http.ResponseWriter, r *http.Request) {
	s.mu.Lock()
	sc, _ := json.MarshalIndent(s.scenario, "", "  ")
	s.mu.Unlock()
	content := ""
	title := "Anropslogg"
	if r.URL.Path == "/mail" {
		title = "Fångad e-post"
		files, _ := filepath.Glob(filepath.Join(s.dir, "mail", "*.eml"))
		for _, p := range files {
			b, _ := os.ReadFile(p)
			content += filepath.Base(p) + "\n" + string(b) + "\n\n"
		}
	} else {
		b, _ := os.ReadFile(filepath.Join(s.dir, "calls.jsonl"))
		content = string(b)
	}
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	_ = simPage.Execute(w, map[string]string{"Scenarios": string(sc), "Title": title, "Content": content})
}
func (s *Simulator) smtp(l net.Listener) {
	for {
		c, e := l.Accept()
		if e != nil {
			return
		}
		go s.smtpConnection(c)
	}
}
func (s *Simulator) smtpConnection(c net.Conn) {
	defer c.Close()
	_ = c.SetDeadline(time.Now().Add(time.Minute))
	fmt.Fprint(c, "220 sandbox.invalid SMTP capture\r\n")
	rd := bufio.NewReaderSize(c, 64*1024)
	sender := ""
	recipients := []string{}
	for {
		line, e := rd.ReadString('\n')
		if e != nil || len(line) > 8192 {
			return
		}
		line = strings.TrimSpace(line)
		upper := strings.ToUpper(line)
		switch {
		case strings.HasPrefix(upper, "EHLO"), strings.HasPrefix(upper, "HELO"):
			fmt.Fprint(c, "250 sandbox.invalid\r\n")
		case strings.HasPrefix(upper, "MAIL FROM:"):
			sender = line[10:]
			recipients = nil
			fmt.Fprint(c, "250 OK\r\n")
		case strings.HasPrefix(upper, "RCPT TO:"):
			address := strings.TrimSpace(line[8:])
			if !syntheticRecipient.MatchString(address) {
				fmt.Fprint(c, "550 Synthetic recipient required\r\n")
				continue
			}
			recipients = append(recipients, address)
			fmt.Fprint(c, "250 OK\r\n")
		case upper == "DATA":
			if sender == "" || len(recipients) == 0 {
				fmt.Fprint(c, "503 Envelope required\r\n")
				continue
			}
			fmt.Fprint(c, "354 End with dot\r\n")
			var b strings.Builder
			for {
				line, e = rd.ReadString('\n')
				if e != nil {
					return
				}
				if line == ".\r\n" || line == ".\n" {
					break
				}
				if strings.HasPrefix(line, "..") {
					line = line[1:]
				}
				if b.Len()+len(line) > 8*1024*1024 {
					fmt.Fprint(c, "552 Message too large\r\n")
					return
				}
				b.WriteString(line)
			}
			d := filepath.Join(s.dir, "mail")
			if e = os.MkdirAll(d, 0700); e != nil {
				fmt.Fprint(c, "451 Storage unavailable\r\n")
				continue
			}
			p := filepath.Join(d, time.Now().UTC().Format("20060102T150405.000000000")+"-"+token(4)+".eml")
			if e = os.WriteFile(p, []byte(b.String()), 0600); e != nil {
				fmt.Fprint(c, "451 Storage unavailable\r\n")
				continue
			}
			s.record("SMTP", "/mail", "success", "Envelope: "+sender+" -> "+strings.Join(recipients, ", "))
			fmt.Fprint(c, "250 Captured locally\r\n")
			sender = ""
			recipients = nil
		case upper == "QUIT":
			fmt.Fprint(c, "221 Bye\r\n")
			return
		case upper == "RSET":
			sender = ""
			recipients = nil
			fmt.Fprint(c, "250 OK\r\n")
		case upper == "NOOP":
			fmt.Fprint(c, "250 OK\r\n")
		default:
			fmt.Fprint(c, "502 Unsupported command\r\n")
		}
	}
}
