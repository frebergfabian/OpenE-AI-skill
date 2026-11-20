package main

import (
	"archive/zip"
	"bytes"
	"crypto/pbkdf2"
	"crypto/rand"
	"crypto/sha256"
	_ "embed"
	"encoding/hex"
	"encoding/json"
	"errors"
	"flag"
	"fmt"
	"io"
	"net"
	"net/http"
	"net/http/cookiejar"
	"net/url"
	"os"
	"os/exec"
	"os/user"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"
	"sync"
	"time"
)

//go:embed payload.zip
var payload []byte
var version = "0.1.0-dev"
var root string

type State struct {
	Name      string `json:"name"`
	Version   string `json:"version"`
	Ephemeral bool   `json:"ephemeral"`
	HTTP      int    `json:"httpPort"`
	DB        int    `json:"dbPort"`
	Simulator int    `json:"simulatorPort"`
	SMTP      int    `json:"smtpPort"`
	Control   int    `json:"controlPort"`
	Phase     string `json:"phase"`
	Error     string `json:"error,omitempty"`
}
type Secrets struct {
	DB       string            `json:"database"`
	Control  string            `json:"control"`
	Accounts map[string]string `json:"accounts"`
}
type Instance struct {
	Dir     string
	Program string
	State   State
	Secrets Secrets
}

func token(n int) string {
	b := make([]byte, n)
	if _, e := rand.Read(b); e != nil {
		panic(e)
	}
	return hex.EncodeToString(b)
}
func must(e error) {
	if e != nil {
		panic(e)
	}
}
func readJSON(p string, v any) error {
	b, e := os.ReadFile(p)
	if e != nil {
		return e
	}
	return json.Unmarshal(b, v)
}
func writeJSON(p string, v any) error {
	b, e := json.MarshalIndent(v, "", "  ")
	if e != nil {
		return e
	}
	tmp := p + ".tmp"
	if e = os.WriteFile(tmp, b, 0600); e != nil {
		return e
	}
	return os.Rename(tmp, p)
}
func localURL(p int) string     { return fmt.Sprintf("http://127.0.0.1:%d", p) }
func (i *Instance) save() error { return writeJSON(filepath.Join(i.Dir, "state.json"), i.State) }
func load(name string) (*Instance, error) {
	if !regexp.MustCompile(`^[a-zA-Z0-9][a-zA-Z0-9_-]{0,47}$`).MatchString(name) {
		return nil, errors.New("invalid instance name")
	}
	i := &Instance{Dir: filepath.Join(root, "instances", name), Program: filepath.Join(root, version), State: State{Name: name, Version: version}}
	e := readJSON(filepath.Join(i.Dir, "state.json"), &i.State)
	if e != nil {
		return i, e
	}
	if i.State.Name != name {
		return i, errors.New("instance name mismatch")
	}
	if i.State.Version != version {
		return i, fmt.Errorf("instance uses %s; use that executable before upgrading", i.State.Version)
	}
	return i, readJSON(filepath.Join(i.Dir, "secrets.json"), &i.Secrets)
}
func main() {
	defer func() {
		if r := recover(); r != nil {
			fmt.Fprintln(os.Stderr, "OpenE-Sandbox:", r)
			os.Exit(1)
		}
	}()
	root = os.Getenv("OPENE_SANDBOX_HOME")
	if root == "" {
		root = filepath.Join(os.Getenv("LOCALAPPDATA"), "OpenE-Sandbox")
	}
	var e error
	root, e = filepath.Abs(root)
	must(e)
	args := os.Args[1:]
	cmd := "start"
	if len(args) > 0 {
		cmd = args[0]
		args = args[1:]
	}
	if cmd == "version" {
		fmt.Println(version)
		return
	}
	flags := flag.NewFlagSet(cmd, flag.ExitOnError)
	name := flags.String("name", "default", "instance name")
	ephemeral := flags.Bool("ephemeral", false, "create a disposable instance")
	yes := flags.Bool("yes", false, "confirm reset/destroy")
	secrets := flags.Bool("secrets", false, "include local credentials")
	asJSON := flags.Bool("json", false, "JSON output")
	noBrowser := flags.Bool("no-browser", false, "do not open browser")
	must(flags.Parse(args))
	if flags.NArg() != 0 {
		panic("unexpected positional arguments")
	}
	if *ephemeral && cmd == "start" && *name == "default" {
		*name = "ephemeral-" + token(5)
	}
	i, loadErr := load(*name)
	if i == nil {
		must(loadErr)
	}
	if cmd == "_serve" {
		must(loadErr)
		must(i.serve())
		return
	}
	if cmd == "status" {
		result := map[string]any{"healthy": false, "name": *name, "exists": loadErr == nil}
		if loadErr == nil {
			result = i.health()
		}
		if *asJSON {
			printJSON(result)
		} else {
			fmt.Printf("%v\n", result)
		}
		if result["healthy"] != true {
			os.Exit(3)
		}
		return
	}
	if cmd == "info" {
		must(loadErr)
		r := map[string]any{"name": *name, "url": localURL(i.State.HTTP), "simulatorUrl": localURL(i.State.Simulator), "mailUrl": localURL(i.State.Simulator) + "/mail", "dataDirectory": i.Dir, "credentialsFile": filepath.Join(i.Dir, "secrets.json"), "version": version}
		if *secrets {
			r["accounts"] = i.Secrets.Accounts
		}
		printJSON(r)
		return
	}
	if cmd == "logs" {
		must(loadErr)
		for _, p := range []string{"supervisor.log", "tomcat.log", "mariadb.log", "webapps/ROOT/WEB-INF/systemlog.log"} {
			b, e := os.ReadFile(filepath.Join(i.Dir, p))
			if e == nil {
				fmt.Printf("\n--- %s ---\n%s", p, b)
			}
		}
		return
	}
	unlock, e := lock(i.Dir + "-operation")
	must(e)
	defer unlock()
	switch cmd {
	case "start":
		if loadErr != nil && !os.IsNotExist(loadErr) {
			must(loadErr)
		}
		if os.IsNotExist(loadErr) {
			must(i.create(*ephemeral))
		}
		must(i.start())
		printJSON(i.health())
		fmt.Println("OpenE:", localURL(i.State.HTTP), "\nSimulator:", localURL(i.State.Simulator), "\nE-post:", localURL(i.State.Simulator)+"/mail", "\nLösenord:", filepath.Join(i.Dir, "secrets.json"))
		if !*noBrowser {
			c := exec.Command("rundll32.exe", "url.dll,FileProtocolHandler", localURL(i.State.HTTP))
			hidden(c, false)
			_ = c.Start()
		}
	case "stop":
		must(loadErr)
		must(i.stop())
	case "reset", "destroy":
		must(loadErr)
		if !*yes && !i.State.Ephemeral {
			panic("persistent instances require --yes")
		}
		must(i.stop())
		must(safeRemove(filepath.Join(root, "instances"), i.Dir))
		if cmd == "reset" {
			must(i.create(i.State.Ephemeral))
			must(i.start())
		}
	default:
		panic("commands: start, stop, status --json, info [--secrets], reset --yes, destroy --yes, logs, version")
	}
}
func printJSON(v any) { b, e := json.MarshalIndent(v, "", "  "); must(e); fmt.Println(string(b)) }
func safeRemove(parent, target string) error {
	p, e := filepath.Abs(parent)
	if e != nil {
		return e
	}
	t, e := filepath.Abs(target)
	if e != nil {
		return e
	}
	rel, e := filepath.Rel(p, t)
	if e != nil || rel == "." || strings.HasPrefix(rel, "..") || filepath.IsAbs(rel) {
		return errors.New("unsafe removal path")
	}
	// Refuse junction/symlink ancestors before recursive removal.
	for q := t; q != p; q = filepath.Dir(q) {
		s, e := os.Lstat(q)
		if e != nil {
			return e
		}
		if s.Mode()&os.ModeSymlink != 0 {
			return errors.New("symlink removal refused")
		}
	}
	return os.RemoveAll(t)
}
func (i *Instance) create(ephemeral bool) error {
	if _, e := os.Stat(i.Dir); e == nil {
		return errors.New("incomplete instance directory exists; inspect it before removal")
	}
	if e := os.MkdirAll(i.Dir, 0700); e != nil {
		return e
	}
	u, e := user.Current()
	if e != nil {
		return e
	}
	c := exec.Command("icacls.exe", i.Dir, "/inheritance:r", "/grant:r", u.Username+":(OI)(CI)F", "SYSTEM:(OI)(CI)F")
	hidden(c, false)
	if b, e := c.CombinedOutput(); e != nil {
		return fmt.Errorf("protect data directory: %s: %w", b, e)
	}
	i.Secrets = Secrets{DB: token(24), Control: token(32), Accounts: map[string]string{}}
	for _, u := range []string{"sandbox-admin", "handlaggare", "handlaggarchef", "sokande", "sokande2"} {
		i.Secrets.Accounts[u] = token(18)
	}
	i.State.Ephemeral = ephemeral
	i.State.Phase = "created"
	if e := writeJSON(filepath.Join(i.Dir, "secrets.json"), i.Secrets); e != nil {
		return e
	}
	return i.save()
}
func extractZip(data []byte, dest string) error {
	z, e := zip.NewReader(bytes.NewReader(data), int64(len(data)))
	if e != nil {
		return e
	}
	for _, f := range z.File {
		n := filepath.FromSlash(f.Name)
		if !filepath.IsLocal(n) || f.Mode()&os.ModeSymlink != 0 {
			return fmt.Errorf("unsafe archive member %q", f.Name)
		}
		p := filepath.Join(dest, n)
		if f.FileInfo().IsDir() {
			if e = os.MkdirAll(p, 0700); e != nil {
				return e
			}
			continue
		}
		if e = os.MkdirAll(filepath.Dir(p), 0700); e != nil {
			return e
		}
		r, e := f.Open()
		if e != nil {
			return e
		}
		w, e := os.OpenFile(p, os.O_CREATE|os.O_TRUNC|os.O_WRONLY, 0600)
		if e != nil {
			r.Close()
			return e
		}
		_, e = io.Copy(w, r)
		ce := w.Close()
		r.Close()
		if e != nil {
			return e
		}
		if ce != nil {
			return ce
		}
	}
	return nil
}
func (i *Instance) unpack() error {
	unlock, e := lock(i.Program + "-extract")
	if e != nil {
		return e
	}
	defer unlock()
	sum := sha256.Sum256(payload)
	expected := hex.EncodeToString(sum[:])
	marker := filepath.Join(i.Program, "payload.sha256")
	if b, e := os.ReadFile(marker); e == nil && string(b) == expected {
		return nil
	}
	if _, e := os.Stat(i.Program); e == nil {
		return errors.New("program directory exists with a different/incomplete payload; use a new version")
	}
	temp := i.Program + "-extract-" + token(6)
	if e = os.MkdirAll(temp, 0700); e != nil {
		return e
	}
	if e = extractZip(payload, temp); e != nil {
		_ = safeRemove(root, temp)
		return e
	}
	if e = os.WriteFile(filepath.Join(temp, "payload.sha256"), []byte(expected), 0600); e != nil {
		return e
	}
	// Windows scanners can briefly hold a newly extracted file or directory.
	// Keep the complete staging tree and retry instead of reporting a broken install.
	for n := 0; n < 40; n++ {
		if e = os.Rename(temp, i.Program); e == nil { return nil }
		time.Sleep(250 * time.Millisecond)
	}
	return e
}
func (i *Instance) control(method, p string) ([]byte, error) {
	req, e := http.NewRequest(method, localURL(i.State.Control)+p, nil)
	if e != nil {
		return nil, e
	}
	req.Header.Set("Authorization", "Bearer "+i.Secrets.Control)
	c := &http.Client{Timeout: 5 * time.Second, Transport: &http.Transport{Proxy: nil}}
	r, e := c.Do(req)
	if e != nil {
		return nil, e
	}
	defer r.Body.Close()
	b, e := io.ReadAll(io.LimitReader(r.Body, 1024*1024))
	if r.StatusCode != 200 {
		return nil, fmt.Errorf("control status %d", r.StatusCode)
	}
	return b, e
}
func (i *Instance) start() error {
	if _, e := i.control("GET", "/alive"); e == nil {
		if i.health()["healthy"] == true {
			return nil
		}
		return errors.New("instance is running but unhealthy; inspect logs")
	}
	if i.State.Phase == "failed" {
		return errors.New("previous startup failed; inspect logs and reset --yes from a clean database")
	}
	if e := i.unpack(); e != nil {
		return e
	}
	i.State.Phase = "starting"
	i.State.Error = ""
	if e := i.save(); e != nil {
		return e
	}
	exe, e := os.Executable()
	if e != nil {
		return e
	}
	f, e := os.OpenFile(filepath.Join(i.Dir, "supervisor.log"), os.O_CREATE|os.O_APPEND|os.O_WRONLY, 0600)
	if e != nil {
		return e
	}
	defer f.Close()
	c := exec.Command(exe, "_serve", "--name", i.State.Name)
	hidden(c, true)
	c.Stdout = f
	c.Stderr = f
	if e = c.Start(); e != nil {
		return e
	}
	_ = c.Process.Release()
	deadline := time.Now().Add(3 * time.Minute)
	for time.Now().Before(deadline) {
		time.Sleep(time.Second)
		var s State
		if readJSON(filepath.Join(i.Dir, "state.json"), &s) != nil {
			continue
		}
		i.State = s
		if s.Phase == "failed" {
			return errors.New(s.Error)
		}
		if s.Phase == "ready" && i.health()["healthy"] == true {
			return nil
		}
	}
	_ = i.stop()
	return errors.New("startup exceeded three minutes; see supervisor.log")
}
func (i *Instance) stop() error {
	_, e := i.control("POST", "/stop")
	if e != nil {
		unlock, le := lock(i.Dir + "-running")
		if le != nil {
			return fmt.Errorf("supervisor is running but control channel is unavailable: %w", e)
		}
		unlock()
		return nil
	}
	for n := 0; n < 80; n++ {
		time.Sleep(250 * time.Millisecond)
		unlock, e := lock(i.Dir + "-running")
		if e == nil {
			unlock()
			return nil
		}
	}
	return errors.New("timed out waiting for clean shutdown")
}
func (i *Instance) health() map[string]any {
	r := map[string]any{"healthy": false, "name": i.State.Name, "phase": i.State.Phase, "url": localURL(i.State.HTTP), "simulatorUrl": localURL(i.State.Simulator)}
	b, e := i.control("GET", "/health")
	if e != nil {
		r["error"] = "supervisor unavailable"
		return r
	}
	if json.Unmarshal(b, &r) != nil {
		r["healthy"] = false
	}
	return r
}
func listen() (net.Listener, error) { return net.Listen("tcp4", "127.0.0.1:0") }
func port(l net.Listener) int       { return l.Addr().(*net.TCPAddr).Port }
func freePort() (int, error) {
	l, e := listen()
	if e != nil {
		return 0, e
	}
	defer l.Close()
	return port(l), nil
}
func copyTree(src, dst string) error {
	return filepath.WalkDir(src, func(p string, d os.DirEntry, e error) error {
		if e != nil {
			return e
		}
		rel, e := filepath.Rel(src, p)
		if e != nil {
			return e
		}
		q := filepath.Join(dst, rel)
		if d.IsDir() {
			return os.MkdirAll(q, 0700)
		}
		if d.Type()&os.ModeSymlink != 0 {
			return errors.New("symlinks not allowed in payload")
		}
		b, e := os.ReadFile(p)
		if e != nil {
			return e
		}
		return os.WriteFile(q, b, 0600)
	})
}
func (i *Instance) sql(query string, db bool) (string, error) {
	args := []string{"--defaults-file=" + filepath.Join(i.Dir, "client.ini"), "--default-character-set=utf8mb4", "--batch", "--skip-column-names"}
	if db {
		args = append(args, "opene")
	}
	c := exec.Command(filepath.Join(i.Program, "mariadb", "bin", "mariadb.exe"), args...)
	hidden(c, false)
	c.Stdin = strings.NewReader(query)
	b, e := c.CombinedOutput()
	if e != nil {
		return "", fmt.Errorf("database command failed (SQL and credentials omitted): %w; %s", e, b)
	}
	return strings.TrimSpace(string(b)), nil
}
func (i *Instance) seed() error {
	q := "START TRANSACTION;\n"
	groups := map[string][]int{"sandbox-admin": {6, 10, 11}, "handlaggare": {11}, "handlaggarchef": {11}, "sokande": {9}, "sokande2": {9}}
	identifiers := map[string]string{"sandbox-admin": "19000101-0008", "handlaggare": "19000101-0016", "handlaggarchef": "19000101-0024", "sokande": "19000101-0032", "sokande2": "19000101-0040"}
	for _, u := range []string{"sandbox-admin", "handlaggare", "handlaggarchef", "sokande", "sokande2"} {
		salt := make([]byte, 128)
		if _, e := rand.Read(salt); e != nil {
			return e
		}
		h, e := pbkdf2.Key(sha256.New, i.Secrets.Accounts[u], salt, 310000, 64)
		if e != nil {
			return e
		}
		hash := "310000:" + hex.EncodeToString(salt) + ":" + hex.EncodeToString(h)
		admin := 0
		if u == "sandbox-admin" {
			admin = 1
		}
		q += fmt.Sprintf("INSERT INTO simple_users(username,password,hashAlgoritm,firstname,lastname,email,admin,enabled,added) VALUES('%s','%s','PBKDF2WithHmacSHA256','Sandbox','%s','%s@exempelkoping.example.com',%d,1,NOW());\n", u, hash, u, u, admin)
		q += fmt.Sprintf("INSERT INTO simple_user_attributes(userID,name,value) SELECT userID,'citizenIdentifier','%s' FROM simple_users WHERE username='%s';\n", identifiers[u], u)
		for _, g := range groups[u] {
			q += fmt.Sprintf("INSERT INTO simple_user_groups(userID,groupID) SELECT userID,%d FROM simple_users WHERE username='%s';\n", g, u)
		}
	}
	q += "REPLACE INTO sandbox_meta VALUES('users_seeded','1'); COMMIT;"
	_, e := i.sql(q, true)
	return e
}
func (i *Instance) serve() (err error) {
	unlock, e := lock(i.Dir + "-running")
	if e != nil {
		return e
	}
	defer unlock()
	var children []*exec.Cmd
	var files []*os.File
	defer func() {
		for j := len(children) - 1; j >= 0; j-- {
			_ = children[j].Process.Kill()
			_ = children[j].Wait()
		}
		for _, f := range files {
			f.Close()
		}
		if err != nil {
			i.State.Phase = "failed"
			i.State.Error = err.Error()
		} else {
			i.State.Phase = "stopped"
		}
		_ = i.save()
	}()
	control, e := listen()
	if e != nil {
		return e
	}
	defer control.Close()
	sim, e := listen()
	if e != nil {
		return e
	}
	defer sim.Close()
	smtp, e := listen()
	if e != nil {
		return e
	}
	defer smtp.Close()
	deny, e := listen()
	if e != nil {
		return e
	}
	defer deny.Close()
	i.State.Control = port(control)
	i.State.Simulator = port(sim)
	i.State.SMTP = port(smtp)
	i.State.HTTP, e = freePort()
	if e != nil {
		return e
	}
	i.State.DB, e = freePort()
	if e != nil {
		return e
	}
	if e = i.save(); e != nil {
		return e
	}
	stop := make(chan struct{})
	var stopOnce sync.Once
	var ready bool
	var mu sync.RWMutex
	mux := http.NewServeMux()
	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		if r.Host != fmt.Sprintf("127.0.0.1:%d", i.State.Control) || r.Header.Get("Authorization") != "Bearer "+i.Secrets.Control {
			http.Error(w, "Forbidden", 403)
			return
		}
		switch r.URL.Path {
		case "/alive":
			w.Write([]byte("ok"))
		case "/stop":
			if r.Method != "POST" {
				http.Error(w, "POST required", 405)
				return
			}
			w.Write([]byte("stopping"))
			stopOnce.Do(func() { close(stop) })
		case "/health":
			mu.RLock()
			ok := ready
			mu.RUnlock()
			checks := map[string]bool{"ready": ok, "http": false, "loginForm": false, "accounts": false, "simulator": false, "adminLogin": false}
			if ok {
				checks = i.checks()
			}
			healthy := true
			for _, v := range checks {
				healthy = healthy && v
			}
			w.Header().Set("Content-Type", "application/json")
			json.NewEncoder(w).Encode(map[string]any{"healthy": healthy, "name": i.State.Name, "url": localURL(i.State.HTTP), "simulatorUrl": localURL(i.State.Simulator), "checks": checks})
		default:
			http.NotFound(w, r)
		}
	})
	server := &http.Server{Handler: mux, ReadHeaderTimeout: 5 * time.Second}
	defer server.Close()
	go server.Serve(control)
	go func() {
		for {
			c, e := deny.Accept()
			if e != nil {
				return
			}
			c.Close()
		}
	}()
	simulator := newSimulator(i.Dir, i.State.Simulator)
	simulator.appPort = i.State.HTTP
	go simulator.serve(sim)
	go simulator.smtp(smtp)
	start := func(exe string, args []string, log string, env []string) error {
		f, e := os.OpenFile(filepath.Join(i.Dir, log), os.O_CREATE|os.O_APPEND|os.O_WRONLY, 0600)
		if e != nil {
			return e
		}
		files = append(files, f)
		c := exec.Command(exe, args...)
		hidden(c, false)
		c.Dir = i.Dir
		c.Env = append(os.Environ(), env...)
		c.Stdout = f
		c.Stderr = f
		if e = c.Start(); e != nil {
			return e
		}
		children = append(children, c)
		return nil
	}
	dbData := filepath.Join(i.Dir, "database")
	if _, e = os.Stat(dbData); os.IsNotExist(e) {
		c := exec.Command(filepath.Join(i.Program, "mariadb", "bin", "mariadb-install-db.exe"), "--datadir="+dbData, "--password="+i.Secrets.DB, "--port="+strconv.Itoa(i.State.DB))
		hidden(c, false)
		b, e := c.CombinedOutput()
		if e != nil {
			return fmt.Errorf("database initialization failed: %w: %s", e, b)
		}
	}
	ini := fmt.Sprintf("[client]\nhost=127.0.0.1\nport=%d\nuser=root\npassword=%s\nprotocol=tcp\n", i.State.DB, i.Secrets.DB)
	if e = os.WriteFile(filepath.Join(i.Dir, "client.ini"), []byte(ini), 0600); e != nil {
		return e
	}
	dbArgs := []string{"--no-defaults", "--basedir=" + filepath.Join(i.Program, "mariadb"), "--datadir=" + dbData, "--port=" + strconv.Itoa(i.State.DB), "--bind-address=127.0.0.1", "--character-set-server=latin1", "--collation-server=latin1_swedish_ci", "--sql-mode=NO_ENGINE_SUBSTITUTION", "--innodb-strict-mode=OFF", "--max-allowed-packet=64M", "--console"}
	if e = start(filepath.Join(i.Program, "mariadb", "bin", "mariadbd.exe"), dbArgs, "mariadb.log", nil); e != nil {
		return e
	}
	if e = poll(30*time.Second, func() bool { _, e := i.sql("SELECT 1;", false); return e == nil }); e != nil {
		return e
	}
	if _, e = i.sql("CREATE DATABASE IF NOT EXISTS opene; CREATE USER IF NOT EXISTS 'opene'@'127.0.0.1' IDENTIFIED BY '"+i.Secrets.DB+"'; GRANT ALL ON opene.* TO 'opene'@'127.0.0.1';", false); e != nil {
		return e
	}
	init, _ := i.sql("SELECT v FROM sandbox_meta WHERE k='initialized';", true)
	first := init != "1"
	if first {
		for _, d := range []string{"files/flowforms", "files/pdf", "files/xml", "files/uploads", "files/siteprofile", "tmp"} {
			if e = os.MkdirAll(filepath.Join(i.Dir, d), 0700); e != nil {
				return e
			}
		}
		for _, p := range []string{"10-demo.sql", "20-sandbox-pre.sql", "25-addons.sql"} {
			b, e := os.ReadFile(filepath.Join(i.Program, "sql", p))
			if os.IsNotExist(e) && p == "25-addons.sql" {
				continue
			}
			if e != nil {
				return e
			}
			s := strings.ReplaceAll(string(b), "/var/lib/opene", filepath.ToSlash(i.Dir))
			s = strings.ReplaceAll(s, "@SIMULATOR@", localURL(i.State.Simulator))
			s = strings.ReplaceAll(s, "'mail'", "'127.0.0.1'")
			s = strings.ReplaceAll(s, "'1025'", "'"+strconv.Itoa(i.State.SMTP)+"'")
			if _, e = i.sql(s, true); e != nil {
				return fmt.Errorf("load %s: %w", p, e)
			}
		}
		if _, e = i.sql("CREATE TABLE sandbox_meta(k VARCHAR(64) PRIMARY KEY,v VARCHAR(255)); INSERT INTO sandbox_meta VALUES('initialized','1');", true); e != nil {
			return e
		}
	}
	if e = i.configure(); e != nil {
		return e
	}
	// Dynamic settings must follow newly allocated ports on every restart.
	q := fmt.Sprintf("UPDATE openhierarchy_foreground_module_settings s JOIN openhierarchy_foreground_modules m USING(moduleID) SET s.value='%d' WHERE m.classname='se.unlogic.hierarchy.foregroundmodules.mailsenders.direct.DirectMailSender' AND s.id='port';", i.State.SMTP)
	q += fmt.Sprintf("UPDATE openhierarchy_foreground_module_settings s JOIN openhierarchy_foreground_modules m USING(moduleID) SET s.value='%s/search' WHERE m.classname='com.nordicpeak.flowengine.queries.pudquery.PUDQueryProviderModule' AND s.id='searchServiceURL';", localURL(i.State.Simulator))
	q += "DELETE s FROM openhierarchy_foreground_module_settings s JOIN openhierarchy_foreground_modules m USING(moduleID) WHERE m.classname='com.nordicpeak.flowengine.notifications.StandardFlowNotificationHandler' AND s.id IN ('installationBaseURL','emailSenderName','emailSenderAddress','smsSenderName','userFlowInstanceModuleAlias','flowInstanceAdminModuleAlias','flowAdminModuleAlias');"
	q += fmt.Sprintf("INSERT INTO openhierarchy_foreground_module_settings(moduleID,id,value) SELECT m.moduleID,v.id,v.value FROM openhierarchy_foreground_modules m JOIN (SELECT 'installationBaseURL' id,'%s' value UNION ALL SELECT 'emailSenderName','OpenE Sandbox' UNION ALL SELECT 'emailSenderAddress','opene@exempelkoping.example.com' UNION ALL SELECT 'smsSenderName','OpenESandbox' UNION ALL SELECT 'userFlowInstanceModuleAlias','/minasidor' UNION ALL SELECT 'flowInstanceAdminModuleAlias','/flowinstanceadmin' UNION ALL SELECT 'flowAdminModuleAlias','/flowadmin') v WHERE m.classname='com.nordicpeak.flowengine.notifications.StandardFlowNotificationHandler';", localURL(i.State.HTTP))
	if _, e = i.sql(q, true); e != nil {
		return e
	}
	javaArgs := []string{"-Dfile.encoding=ISO-8859-1", "-Djava.awt.headless=true", "-Dhttp.proxyHost=127.0.0.1", "-Dhttp.proxyPort=" + strconv.Itoa(port(deny)), "-Dhttps.proxyHost=127.0.0.1", "-Dhttps.proxyPort=" + strconv.Itoa(port(deny)), "-Dhttp.nonProxyHosts=localhost|127.0.0.1", "-Dcatalina.home=" + filepath.Join(i.Program, "tomcat"), "-Dcatalina.base=" + i.Dir, "-Djava.io.tmpdir=" + filepath.Join(i.Dir, "tmp"), "-Dopene.sandbox.simulator=" + localURL(i.State.Simulator), "-cp", filepath.Join(i.Program, "tomcat", "bin", "bootstrap.jar") + ";" + filepath.Join(i.Program, "tomcat", "bin", "tomcat-juli.jar"), "org.apache.catalina.startup.Bootstrap", "start"}
	if e = start(filepath.Join(i.Program, "jre", "bin", "java.exe"), javaArgs, "tomcat.log", nil); e != nil {
		return e
	}
	if e = poll(90*time.Second, func() bool {
		c, e := i.sql("SELECT COUNT(*) FROM information_schema.COLUMNS WHERE TABLE_SCHEMA='opene' AND TABLE_NAME='simple_users' AND COLUMN_NAME='hashAlgoritm';", true)
		return e == nil && c == "1"
	}); e != nil {
		return e
	}
	seeded, _ := i.sql("SELECT v FROM sandbox_meta WHERE k='users_seeded';", true)
	if seeded != "1" {
		if e = i.seed(); e != nil {
			return e
		}
	}
	if e = poll(90*time.Second, func() bool { c := i.checks(); return c["http"] && c["loginForm"] && c["adminLogin"] }); e != nil {
		return e
	}
	if e = i.seedFixtures(); e != nil {
		return e
	}
	mu.Lock()
	ready = true
	mu.Unlock()
	i.State.Phase = "ready"
	if e = i.save(); e != nil {
		return e
	}
	fmt.Println("ready", localURL(i.State.HTTP))
	<-stop
	mu.Lock()
	ready = false
	mu.Unlock()
	// Gracefully stop Tomcat via Bootstrap, then shut down MariaDB.
	shutdown := append([]string{}, javaArgs...)
	shutdown[len(shutdown)-1] = "stop"
	c := exec.Command(filepath.Join(i.Program, "jre", "bin", "java.exe"), shutdown...)
	hidden(c, false)
	_ = c.Run()
	c = exec.Command(filepath.Join(i.Program, "mariadb", "bin", "mariadb-admin.exe"), "--defaults-file="+filepath.Join(i.Dir, "client.ini"), "shutdown")
	hidden(c, false)
	_ = c.Run()
	return nil
}
func poll(d time.Duration, f func() bool) error {
	end := time.Now().Add(d)
	for time.Now().Before(end) {
		if f() {
			return nil
		}
		time.Sleep(500 * time.Millisecond)
	}
	return errors.New("readiness timed out; inspect component logs")
}
func (i *Instance) configure() error {
	if _, e := os.Stat(filepath.Join(i.Dir, "webapps", "ROOT")); os.IsNotExist(e) {
		if e = copyTree(filepath.Join(i.Program, "webapp"), filepath.Join(i.Dir, "webapps", "ROOT")); e != nil {
			return e
		}
	}
	if e := copyTree(filepath.Join(i.Program, "tomcat", "conf"), filepath.Join(i.Dir, "conf")); e != nil {
		return e
	}
	for _, d := range []string{"logs", "temp", "work"} {
		if e := os.MkdirAll(filepath.Join(i.Dir, d), 0700); e != nil {
			return e
		}
	}
	shutdown, e := freePort()
	if e != nil {
		return e
	}
	server := fmt.Sprintf(`<Server port="%d" address="127.0.0.1" shutdown="%s"><Service name="Catalina"><Connector address="127.0.0.1" port="%d" protocol="HTTP/1.1" connectionTimeout="20000" URIEncoding="UTF-8"/><Engine name="Catalina" defaultHost="localhost"><Host name="localhost" appBase="webapps" unpackWARs="false" autoDeploy="false"/></Engine></Service></Server>`, shutdown, i.Secrets.Control, i.State.HTTP)
	if e = os.WriteFile(filepath.Join(i.Dir, "conf", "server.xml"), []byte(server), 0600); e != nil {
		return e
	}
	b, e := os.ReadFile(filepath.Join(i.Program, "config.xml.tmpl"))
	if e != nil {
		return e
	}
	s := string(b)
	for k, v := range map[string]string{"@DB_HOST@": "127.0.0.1", "@DB_NAME@": "opene", "@DB_USER@": "opene", "@DB_PASSWORD@": i.Secrets.DB, ":3306/": ":" + strconv.Itoa(i.State.DB) + "/"} {
		s = strings.ReplaceAll(s, k, v)
	}
	return os.WriteFile(filepath.Join(i.Dir, "webapps", "ROOT", "WEB-INF", "config.xml"), []byte(s), 0600)
}
func (i *Instance) checks() map[string]bool {
	r := map[string]bool{"ready": true, "http": false, "loginForm": false, "accounts": false, "simulator": false, "adminLogin": false, "systemLog": false}
	if log, err := os.ReadFile(filepath.Join(i.Dir, "webapps", "ROOT", "WEB-INF", "systemlog.log")); err == nil {
		r["systemLog"] = !bytes.Contains(log, []byte("[ERROR]"))
	}
	jar, _ := cookiejar.New(nil)
	c := &http.Client{Jar: jar, Timeout: 4 * time.Second, Transport: &http.Transport{Proxy: nil}, CheckRedirect: func(req *http.Request, via []*http.Request) error {
		if req.URL.Host != fmt.Sprintf("127.0.0.1:%d", i.State.HTTP) {
			return errors.New("external redirect blocked")
		}
		if len(via) > 5 {
			return errors.New("redirect loop")
		}
		return nil
	}}
	get := func(p string) (int, string) {
		res, e := c.Get(p)
		if e != nil {
			return 0, ""
		}
		defer res.Body.Close()
		b, _ := io.ReadAll(io.LimitReader(res.Body, 2*1024*1024))
		return res.StatusCode, string(b)
	}
	code, _ := get(localURL(i.State.HTTP) + "/")
	r["http"] = code == 200
	code, body := get(localURL(i.State.HTTP) + "/login")
	r["loginForm"] = code == 200 && strings.Contains(body, `name="password"`)
	count, e := i.sql("SELECT COUNT(*) FROM simple_users WHERE username IN ('sandbox-admin','handlaggare','handlaggarchef','sokande','sokande2') AND enabled=1 AND hashAlgoritm='PBKDF2WithHmacSHA256';", true)
	r["accounts"] = e == nil && count == "5"
	res, e := c.PostForm(localURL(i.State.HTTP)+"/login", url.Values{"username": {"sandbox-admin"}, "password": {i.Secrets.Accounts["sandbox-admin"]}})
	if e == nil {
		io.Copy(io.Discard, res.Body)
		res.Body.Close()
		code, body = get(localURL(i.State.HTTP) + "/flowadmin")
		r["adminLogin"] = code == 200 && strings.Contains(body, "importflow") && !strings.Contains(body, `name="password"`)
		get(localURL(i.State.HTTP) + "/logout")
	}
	sc := &http.Client{Timeout: 2 * time.Second, Transport: &http.Transport{Proxy: nil}}
	res, e = sc.Get(localURL(i.State.Simulator) + "/health")
	if e == nil {
		r["simulator"] = res.StatusCode == 200
		res.Body.Close()
	}
	return r
}
