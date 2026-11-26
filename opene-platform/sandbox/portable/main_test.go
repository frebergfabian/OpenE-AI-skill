package main

import (
	"archive/zip"
	"bufio"
	"bytes"
	"context"
	"fmt"
	"net"
	"net/http"
	"net/http/httptest"
	"net/url"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"testing"
)

func TestArchiveRejectsTraversal(t *testing.T) {
	for _, name := range []string{"../escape", "/absolute", "C:/absolute", "..\\escape"} {
		var b bytes.Buffer
		z := zip.NewWriter(&b)
		w, _ := z.Create(name)
		w.Write([]byte("no"))
		z.Close()
		if extractZip(b.Bytes(), t.TempDir()) == nil {
			t.Fatal("accepted", name)
		}
	}
}

func TestArchiveExtractsRelativeFile(t *testing.T) {
	var buffer bytes.Buffer
	archive := zip.NewWriter(&buffer)
	file, err := archive.Create("webapp/test.txt")
	if err != nil {
		t.Fatal(err)
	}
	if _, err = file.Write([]byte("synthetic")); err != nil {
		t.Fatal(err)
	}
	if err = archive.Close(); err != nil {
		t.Fatal(err)
	}
	destination := t.TempDir()
	if err = extractZip(buffer.Bytes(), destination); err != nil {
		t.Fatal(err)
	}
	content, err := os.ReadFile(filepath.Join(destination, "webapp", "test.txt"))
	if err != nil || string(content) != "synthetic" {
		t.Fatalf("extracted content %q, error %v", content, err)
	}
}

func TestSMTPRejectsRealRecipient(t *testing.T) {
	s := newSimulator(t.TempDir(), 12345)
	client, server := net.Pipe()
	defer client.Close()
	go s.smtpConnection(server)
	r := bufio.NewReader(client)
	read := func(prefix string) {
		t.Helper()
		line, err := r.ReadString('\n')
		if err != nil || !strings.HasPrefix(line, prefix) {
			t.Fatalf("wanted %s, got %q: %v", prefix, line, err)
		}
	}
	read("220")
	fmt.Fprint(client, "EHLO example\r\n")
	read("250")
	fmt.Fprint(client, "MAIL FROM:<sender@sandbox.invalid>\r\n")
	read("250")
	fmt.Fprint(client, "RCPT TO:<person@real.example>\r\n")
	read("550")
	fmt.Fprint(client, "RCPT TO:<test@exempelkoping.example.com>\r\n")
	read("250")
	fmt.Fprint(client, "QUIT\r\n")
	read("221")
}
func TestRemoveRejectsParent(t *testing.T) {
	d := t.TempDir()
	p := filepath.Join(d, "parent")
	os.Mkdir(p, 0700)
	if safeRemove(p, d) == nil {
		t.Fatal("removed parent")
	}
	if safeRemove(p, p) == nil {
		t.Fatal("removed root")
	}
}
func TestUnknownIntegrationNeverSucceeds(t *testing.T) {
	s := newSimulator(t.TempDir(), 12345)
	r := httptest.NewRequest("POST", "http://127.0.0.1:12345/not-implemented", nil)
	w := httptest.NewRecorder()
	s.handle(w, r)
	if w.Code != 501 {
		t.Fatal(w.Code)
	}
}
func TestSimulatorBlocksCrossOriginAndRebinding(t *testing.T) {
	s := newSimulator(t.TempDir(), 12345)
	for _, host := range []string{"evil.invalid:12345", "127.0.0.1:12345"} {
		r := httptest.NewRequest("POST", "http://"+host+"/scenarios", strings.NewReader("module=http&scenario=success"))
		r.Header.Set("Origin", "https://evil.invalid")
		w := httptest.NewRecorder()
		s.handle(w, r)
		if w.Code != 403 {
			t.Fatal(w.Code)
		}
	}
}
func TestPUDContract(t *testing.T) {
	s := newSimulator(t.TempDir(), 12345)
	r := httptest.NewRequest("GET", "http://127.0.0.1:12345/search/registerenheter?q=Testfastighet", nil)
	w := httptest.NewRecorder()
	s.handle(w, r)
	if w.Code != 200 || !strings.Contains(w.Body.String(), `"objid":"00000000-0000-4000-8000-000000000001"`) {
		t.Fatal(w.Code, w.Body)
	}
}

func TestArtviseCreateCaseSOAP(t *testing.T) {
	s := newSimulator(t.TempDir(), 12345)
	request := `<s:Envelope xmlns:s="http://schemas.xmlsoap.org/soap/envelope/"><s:Body><CreateCase xmlns="http://www.artvise.se/MercuryAPI/"><apiKey>sandbox</apiKey><caseDto/></CreateCase></s:Body></s:Envelope>`
	for _, tc := range []struct {
		scenario string
		code     int
		want     string
	}{
		{"success", 200, "<a:Ok>true</a:Ok>"}, {"denied", 200, "<a:Ok>false</a:Ok>"}, {"error", 500, "<s:Fault>"},
	} {
		r := httptest.NewRequest("POST", "http://127.0.0.1:12345/integration/artvise", strings.NewReader(request))
		r.Header.Set("X-Sandbox-Scenario", tc.scenario)
		w := httptest.NewRecorder()
		s.handle(w, r)
		if w.Code != tc.code || !strings.Contains(w.Body.String(), tc.want) {
			t.Fatalf("%s: HTTP %d, body %s", tc.scenario, w.Code, w.Body.String())
		}
	}
	r := httptest.NewRequest("POST", "http://127.0.0.1:12345/integration/artvise", strings.NewReader(`<not-soap/>`))
	w := httptest.NewRecorder()
	s.handle(w, r)
	if w.Code != 400 {
		t.Fatalf("unexpected operation: %d", w.Code)
	}
}

func TestEasitImportItemsSOAP(t *testing.T) {
	s := newSimulator(t.TempDir(), 12345)
	request := `<s:Envelope xmlns:s="http://schemas.xmlsoap.org/soap/envelope/"><s:Body><ImportItemsRequest xmlns="http://www.easit.com/bps/schemas"><ItemToImport UID="test-123"/></ImportItemsRequest></s:Body></s:Envelope>`
	for _, tc := range []struct {
		scenario string
		code     int
		want     string
	}{
		{"success", 200, `UID="test-123" result="OK"`}, {"denied", 200, `result="Skipped"`}, {"error", 500, "<s:Fault>"},
	} {
		r := httptest.NewRequest("POST", "http://127.0.0.1:12345/integration/easit", strings.NewReader(request))
		r.Header.Set("X-Sandbox-Scenario", tc.scenario)
		w := httptest.NewRecorder()
		s.handle(w, r)
		if w.Code != tc.code || !strings.Contains(w.Body.String(), tc.want) {
			t.Fatalf("%s: HTTP %d, body %s", tc.scenario, w.Code, w.Body.String())
		}
	}
	r := httptest.NewRequest("POST", "http://127.0.0.1:12345/integration/easit", strings.NewReader(`<not-soap/>`))
	w := httptest.NewRecorder()
	s.handle(w, r)
	if w.Code != 400 {
		t.Fatalf("unexpected operation: %d", w.Code)
	}
}

func TestSSBTSoapRequiresSyntheticData(t *testing.T) {
	s := newSimulator(t.TempDir(), 12345)
	for _, tc := range []struct{ path, namespace, operation string }{
		{"/ssbt/en", "http://schema.bolagsverket.se/ssbten/engagemang", "EngagemangBegaran"},
		{"/ssbt/gu", "http://schema.bolagsverket.se/ssbtgu/v2/grundlaggandeuppgifter", "GrundlaggandeUppgifterBegaran"},
	} {
		for _, person := range []string{"190001010032", "199901011234"} {
			body := fmt.Sprintf(`<s:Envelope xmlns:s="http://schemas.xmlsoap.org/soap/envelope/"><s:Body><q:%s xmlns:q="%s"><Personnummer>%s</Personnummer><Organisationsnummer>0000000000</Organisationsnummer><GrundlaggandeUppgiftId>UD0001</GrundlaggandeUppgiftId></q:%s></s:Body></s:Envelope>`, tc.operation, tc.namespace, person, tc.operation)
			r := httptest.NewRequest("POST", "http://127.0.0.1:12345"+tc.path, strings.NewReader(body))
			w := httptest.NewRecorder()
			s.handle(w, r)
			if person == "190001010032" && (w.Code != 200 || !strings.Contains(w.Body.String(), person)) {
				t.Fatalf("%s: HTTP %d %s", tc.path, w.Code, w.Body.String())
			}
			if person != "190001010032" && w.Code != 403 {
				t.Fatalf("%s accepted non-synthetic identity: %d", tc.path, w.Code)
			}
		}
	}
}

func TestSSBTDetailsAreLimitedToRequestedCodes(t *testing.T) {
	s := newSimulator(t.TempDir(), 12345)
	body := `<s:Envelope xmlns:s="http://schemas.xmlsoap.org/soap/envelope/"><s:Body><q:GrundlaggandeUppgifterBegaran xmlns:q="http://schema.bolagsverket.se/ssbtgu/v2/grundlaggandeuppgifter"><Personnummer>190001010032</Personnummer><Organisationsnummer>0000000000</Organisationsnummer><GrundlaggandeUppgiftId>UD0001</GrundlaggandeUppgiftId><GrundlaggandeUppgiftId>UD0003</GrundlaggandeUppgiftId><GrundlaggandeUppgiftId>UD0004</GrundlaggandeUppgiftId></q:GrundlaggandeUppgifterBegaran></s:Body></s:Envelope>`
	for _, tc := range []struct {
		body string
		code int
	}{{body, 200}, {strings.Replace(body, "UD0004", "UD0048", 1), 501}} {
		r := httptest.NewRequest("POST", "http://127.0.0.1:12345/ssbt/gu", strings.NewReader(tc.body))
		w := httptest.NewRecorder()
		s.handle(w, r)
		if w.Code != tc.code {
			t.Fatalf("HTTP %d: %s", w.Code, w.Body.String())
		}
		if tc.code == 200 && (!strings.Contains(w.Body.String(), "<ns2:UD0003>") || !strings.Contains(w.Body.String(), "<ns2:UD0004>") || strings.Contains(w.Body.String(), "<ns2:UD0006>")) {
			t.Fatal(w.Body.String())
		}
	}
}

func TestSMSOnlyAcceptsSyntheticNumber(t *testing.T) {
	s := newSimulator(t.TempDir(), 12345)
	for _, tc := range []struct {
		body   string
		status int
	}{
		{`{"recipients":["0700000000"],"message":"test"}`, 200},
		{`{"recipients":["0701234567"],"message":"test"}`, 403},
	} {
		r := httptest.NewRequest("POST", "http://127.0.0.1:12345/sms", strings.NewReader(tc.body))
		w := httptest.NewRecorder()
		s.handle(w, r)
		if w.Code != tc.status {
			t.Fatalf("wanted %d, got %d", tc.status, w.Code)
		}
	}
}

func TestCheckoutSuccessAndCancel(t *testing.T) {
	for _, action := range []string{"pay", "cancel"} {
		t.Run(action, func(t *testing.T) {
			callbacks := 0
			app := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) { callbacks++; w.WriteHeader(200) }))
			defer app.Close()
			u, _ := url.Parse(app.URL)
			_, portString, _ := net.SplitHostPort(u.Host)
			port, _ := strconv.Atoi(portString)
			s := newSimulator(t.TempDir(), 12345)
			s.appPort = port
			body := fmt.Sprintf(`{"paymentOrder":{"operation":"Purchase","currency":"SEK","amount":100,"urls":{"completeUrl":%q,"cancelUrl":%q,"callbackUrl":%q}}}`, app.URL+"/complete", app.URL+"/cancel", app.URL+"/callback")
			w := httptest.NewRecorder()
			s.handle(w, httptest.NewRequest("POST", "http://127.0.0.1:12345/psp/paymentorders", strings.NewReader(body)))
			if w.Code != 201 {
				t.Fatal(w.Code, w.Body)
			}
			id := strings.TrimPrefix(w.Header().Get("Location"), "/psp/paymentorders/")
			r := httptest.NewRequest("POST", "http://127.0.0.1:12345/checkout/"+id, strings.NewReader("action="+action))
			r.Header.Set("Content-Type", "application/x-www-form-urlencoded")
			w = httptest.NewRecorder()
			s.handle(w, r)
			if w.Code != 303 || callbacks != 1 {
				t.Fatal(w.Code, callbacks, w.Body)
			}
			w = httptest.NewRecorder()
			s.handle(w, httptest.NewRequest("GET", "http://127.0.0.1:12345/psp/paymentorders/"+id, nil))
			status := "Paid"
			if action == "cancel" {
				status = "Aborted"
			}
			if w.Code != 200 || !strings.Contains(w.Body.String(), `"status":"`+status+`"`) {
				t.Fatal(w.Code, w.Body)
			}
		})
	}
}

func TestSimulatorFailureScenarios(t *testing.T) {
	for _, tc := range []struct {
		scenario string
		status   int
	}{{"denied", 403}, {"error", 500}, {"bogus", 400}} {
		s := newSimulator(t.TempDir(), 12345)
		r := httptest.NewRequest("POST", "http://127.0.0.1:12345/integration/http", strings.NewReader("<synthetic/>"))
		r.Header.Set("X-Sandbox-Scenario", tc.scenario)
		w := httptest.NewRecorder()
		s.handle(w, r)
		if w.Code != tc.status {
			t.Fatalf("%s: got %d", tc.scenario, w.Code)
		}
	}
	s := newSimulator(t.TempDir(), 12345)
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	r := httptest.NewRequest("POST", "http://127.0.0.1:12345/integration/http", strings.NewReader("<synthetic/>")).WithContext(ctx)
	r.Header.Set("X-Sandbox-Scenario", "timeout")
	w := httptest.NewRecorder()
	s.handle(w, r)
	if w.Body.Len() != 0 {
		t.Fatal("cancelled timeout returned a body")
	}
}
