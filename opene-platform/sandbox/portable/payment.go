package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"html/template"
	"io"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"time"
)

type PaymentOrder struct {
	ID          string         `json:"id"`
	Status      string         `json:"status"`
	Operation   string         `json:"operation"`
	Currency    string         `json:"currency"`
	Amount      int64          `json:"amount"`
	VAT         int64          `json:"vatAmount"`
	Description string         `json:"description"`
	URLs        map[string]any `json:"urls"`
	Payee       map[string]any `json:"payeeInfo"`
}

var checkoutPage = template.Must(template.New("checkout").Parse(`<!doctype html><html lang="sv"><meta charset="utf-8"><title>Simulerad betalning</title><style>body{font:20px system-ui;max-width:640px;margin:70px auto;padding:24px}button{font:inherit;padding:16px;margin:12px}</style><h1>Simulerad betalning</h1><p>Endast syntetiskt test. Inga pengar debiteras.</p><p>{{.Amount}} öre · {{.Currency}}</p><p>Status: {{.Status}}</p><form method="post"><button name="action" value="pay">Betala</button><button name="action" value="cancel">Avbryt</button></form></html>`))

func (s *Simulator) payment(w http.ResponseWriter, r *http.Request) {
	r.Body = http.MaxBytesReader(w, r.Body, 1024*1024)
	base := "/psp/paymentorders"
	id := strings.TrimPrefix(r.URL.Path, base+"/")
	if strings.HasPrefix(r.URL.Path, "/checkout/") {
		id = strings.TrimPrefix(r.URL.Path, "/checkout/")
	}
	if r.URL.Path == base && r.Method == "POST" {
		var request struct {
			Order PaymentOrder `json:"paymentOrder"`
		}
		b, e := io.ReadAll(r.Body)
		if e != nil || json.Unmarshal(b, &request) != nil {
			http.Error(w, "Invalid JSON", 400)
			return
		}
		o := request.Order
		if o.Operation != "Purchase" || o.Currency != "SEK" || o.Amount < 0 {
			http.Error(w, "Unsupported payment order", 400)
			return
		}
		for _, k := range []string{"completeUrl", "cancelUrl", "callbackUrl"} {
			v, _ := o.URLs[k].(string)
			u, e := url.Parse(v)
			if e != nil || u.Scheme != "http" || u.Host != fmt.Sprintf("127.0.0.1:%d", s.appPort) || u.User != nil {
				http.Error(w, "Only this instance's OpenE URLs are allowed", 400)
				return
			}
		}
		s.mu.Lock()
		scenario := s.scenario["payment"]
		s.mu.Unlock()
		if h := r.Header.Get("X-Sandbox-Scenario"); h != "" {
			scenario = h
		}
		if scenario == "" {
			scenario = "success"
		}
		s.record(r.Method, r.URL.Path, scenario, string(b))
		switch scenario {
		case "timeout":
			select {
			case <-r.Context().Done():
			case <-time.After(40 * time.Second):
			}
			return
		case "error":
			http.Error(w, "Simulated payment backend failure", 500)
			return
		case "denied":
			http.Error(w, "Simulated payment denial", 403)
			return
		case "success":
		default:
			http.Error(w, "Invalid scenario", 400)
			return
		}
		id = token(16)
		o.ID = base + "/" + id
		o.Status = "Initialized"
		if e = os.MkdirAll(filepath.Join(s.dir, "payments"), 0700); e != nil {
			http.Error(w, "Storage error", 500)
			return
		}
		if e = writeJSON(filepath.Join(s.dir, "payments", id+".json"), o); e != nil {
			http.Error(w, "Storage error", 500)
			return
		}
		w.Header().Set("Location", o.ID)
		w.WriteHeader(201)
		s.paymentResponse(w, o, id)
		return
	}
	if !regexp.MustCompile(`^[a-f0-9]{32}$`).MatchString(id) {
		http.NotFound(w, r)
		return
	}
	file := filepath.Join(s.dir, "payments", id+".json")
	s.mu.Lock()
	var order PaymentOrder
	e := readJSON(file, &order)
	s.mu.Unlock()
	if e != nil {
		http.NotFound(w, r)
		return
	}
	if strings.HasPrefix(r.URL.Path, "/checkout/") {
		if r.Method == "GET" {
			w.Header().Set("Content-Type", "text/html; charset=utf-8")
			checkoutPage.Execute(w, order)
			return
		}
		if r.Method != "POST" {
			http.Error(w, "POST required", 405)
			return
		}
		if e = r.ParseForm(); e != nil {
			http.Error(w, "Invalid form", 400)
			return
		}
		action := r.FormValue("action")
		if action != "pay" && action != "cancel" {
			http.Error(w, "Invalid action", 400)
			return
		}
		s.mu.Lock()
		readJSON(file, &order)
		if order.Status == "Initialized" {
			if action == "pay" {
				order.Status = "Paid"
			} else {
				order.Status = "Aborted"
			}
			e = writeJSON(file, order)
		}
		s.mu.Unlock()
		if e != nil {
			http.Error(w, "Storage error", 500)
			return
		}
		s.record("POST", r.URL.Path, order.Status, "action="+action)
		// Use a fixed local target and disallow redirects; never call arbitrary callback hosts.
		callback, _ := order.URLs["callbackUrl"].(string)
		u, e := url.Parse(callback)
		if e != nil || u.Scheme != "http" || u.Host != fmt.Sprintf("127.0.0.1:%d", s.appPort) {
			http.Error(w, "Callback target changed", 409)
			return
		}
		b, _ := json.Marshal(map[string]any{"paymentOrder": map[string]string{"id": order.ID}, "timestamp": time.Now().UTC().Format(time.RFC3339Nano)})
		client := &http.Client{Timeout: 15 * time.Second, Transport: &http.Transport{Proxy: nil}, CheckRedirect: func(*http.Request, []*http.Request) error { return http.ErrUseLastResponse }}
		response, e := client.Post(callback, "application/json", bytes.NewReader(b))
		if e != nil {
			s.record("CALLBACK", order.ID, "error", e.Error())
			http.Error(w, "Callback failed; retry payment action", 502)
			return
		}
		io.Copy(io.Discard, response.Body)
		response.Body.Close()
		s.record("CALLBACK", order.ID, fmt.Sprint(response.StatusCode), string(b))
		if response.StatusCode != 200 {
			http.Error(w, "Callback failed; retry payment action", 502)
			return
		}
		target, _ := order.URLs["completeUrl"].(string)
		if order.Status == "Aborted" {
			target, _ = order.URLs["cancelUrl"].(string)
		}
		http.Redirect(w, r, target, 303)
		return
	}
	if r.Method != "GET" {
		http.Error(w, "GET required", 405)
		return
	}
	s.record("GET", r.URL.Path, "success", "")
	s.paymentResponse(w, order, id)
}
func (s *Simulator) paymentResponse(w http.ResponseWriter, o PaymentOrder, id string) {
	jsonResponse(w, map[string]any{"paymentOrder": o, "operations": []map[string]string{{"method": "GET", "href": localURL(s.port) + "/checkout/" + id, "rel": "redirect-checkout", "contentType": "text/html"}}})
}
