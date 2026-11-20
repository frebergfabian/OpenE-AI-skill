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

func TestCompanySandboxIsLocalAndSynthetic(t *testing.T) {
	s := newSimulator(t.TempDir(), 12345)
	for _, tc := range []struct {
		id     string
		status int
	}{{"190001010032", 200}, {"199001011234", 403}} {
		w := httptest.NewRecorder()
		r := httptest.NewRequest("GET", "http://127.0.0.1:12345/ssbt/companies?citizenIdentifier="+tc.id, nil)
		s.handle(w, r)
		if w.Code != tc.status {
			t.Fatalf("%s: %d", tc.id, w.Code)
		}
		if tc.status == 200 && !strings.Contains(w.Body.String(), "0000000000") {
			t.Fatal(w.Body)
		}
	}
	w := httptest.NewRecorder()
	r := httptest.NewRequest("GET", "http://127.0.0.1:12345/ssbt/companies?citizenIdentifier=190001010032", nil)
	r.Header.Set("X-Sandbox-Scenario", "error")
	s.handle(w, r)
	if w.Code != 500 {
		t.Fatal(w.Code)
	}
}
