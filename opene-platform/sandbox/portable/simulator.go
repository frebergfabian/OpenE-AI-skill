package main

import (
	"bufio"
	"encoding/json"
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
		if (module != "pud" && module != "http" && module != "payment" && module != "sms" && module != "ssbt") || (scenario != "success" && scenario != "denied" && scenario != "timeout" && scenario != "error") {
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
	if r.URL.Path == "/ssbt/companies" {
		module = "ssbt"
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
	if module == "ssbt" {
		if r.Method != "GET" {
			http.Error(w, "GET required", 405)
			return
		}
		id := strings.ReplaceAll(r.URL.Query().Get("citizenIdentifier"), "-", "")
		allowed := map[string]bool{"190001010008": true, "190001010016": true, "190001010024": true, "190001010032": true, "190001010040": true}
		if !allowed[id] {
			http.Error(w, "Only sandbox identities are supported", 403)
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
		http.Error(w, "Simulated backend failure", 500)
		return
	case "denied":
		http.Error(w, "Simulated denial", 403)
		return
	case "success":
	default:
		http.Error(w, "Unknown scenario", 400)
		return
	}
	if module == "ssbt" {
		jsonResponse(w, map[string]any{"companies": []map[string]string{{"organizationNumber": "0000000000", "name": "Exempelbolaget (syntetiskt test)", "address": "Testgatan 1", "zip": "00000", "city": "Exempelkoping", "email": "referens@exempelkoping.example.com"}}})
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

var simPage = template.Must(template.New("page").Parse(`<!doctype html><html lang="sv"><meta charset="utf-8"><meta name="viewport" content="width=device-width"><title>OpenE Simulator</title><style>body{font:17px system-ui;max-width:1000px;margin:40px auto;padding:0 24px;background:#f3f6f8;color:#132c40}a{color:#056184}section{background:white;padding:24px;border-radius:12px;margin:20px 0}pre{white-space:pre-wrap;overflow-wrap:anywhere;font-size:13px}select,button{font:inherit;padding:8px}nav{display:flex;gap:24px}</style><h1>OpenE Simulator</h1><p>Exempelköpings kommun · Endast syntetiska testuppgifter</p><nav><a href="/">Anrop och scenarier</a><a href="/mail">E-postfångst</a><a href="/api/calls">JSON-logg</a></nav><section><h2>Scenarier</h2><form method="post" action="/scenarios"><select name="module"><option value="pud">Fastighet och adress</option><option value="http">HTTPIntegrationProvider</option><option value="payment">Betalning</option><option value="sms">SMS</option><option value="ssbt">Företagsuppslag</option></select> <select name="scenario"><option value="success">Lyckat</option><option value="denied">Nekat</option><option value="timeout">Timeout</option><option value="error">Felsvar</option></select> <button>Spara</button></form><p>Valet gäller efterföljande anrop. Ej implementerade integrationer svarar 501.</p><pre>{{.Scenarios}}</pre></section><section><h2>{{.Title}}</h2><pre>{{.Content}}</pre></section></html>`))

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
