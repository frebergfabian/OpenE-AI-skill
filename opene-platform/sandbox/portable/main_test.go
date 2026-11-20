package main

import (
	"archive/zip"
	"bufio"
	"bytes"
	"fmt"
	"net"
	"net/http/httptest"
	"os"
	"path/filepath"
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
