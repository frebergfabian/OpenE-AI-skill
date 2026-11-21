package main

import (
	"bytes"
	"fmt"
	"io"
	"mime/multipart"
	"net/http"
	"net/http/cookiejar"
	"net/url"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"time"
)

// The experiment is opt-in by inclusion of the fixture SQL in the payload.
// A pre-existing import is reused so an interrupted seed never duplicates it.
func (i *Instance) seedFixtures() error {
	sqlPath := filepath.Join(i.Program, "sql", "30-fixtures.sql")
	sqlData, err := os.ReadFile(sqlPath)
	if os.IsNotExist(err) {
		return nil
	}
	if err != nil {
		return err
	}
	done, err := i.sql("SELECT v FROM sandbox_meta WHERE k='fixtures_seeded';", true)
	if err != nil {
		return err
	}
	if done == "1" {
		_, err = i.sql("UPDATE http_integration_settings SET url='"+localURL(i.State.Simulator)+"/integration/http' WHERE flowFamilyID IN (SELECT flowFamilyID FROM flowengine_flows WHERE name='Funktionslabbet - portabelt utkast'); UPDATE artvise_integration_endpoints SET address='"+localURL(i.State.Simulator)+"/integration/artvise' WHERE name='Sandbox Artvise'; UPDATE easit_integration_endpoints SET webServiceAddress='"+localURL(i.State.Simulator)+"/integration/easit' WHERE name='Sandbox Easit';", true)
		return err
	}
	fixtures := []struct{ name, file string }{{"Funktionslabbet - portabelt utkast", "funktionslabbet.oeflow"}, {"Sandbox fastighetstest", "sandbox-fastighet.oeflow"}, {"Sandbox barntest", "sandbox-barn.oeflow"}, {"Sandbox betalningstest", "sandbox-betalning.oeflow"}, {"Sandbox foretagstest", "sandbox-foretag.oeflow"}, {"Sandbox flerpartssignering", "sandbox-flerpart.oeflow"}}
	for _, fixture := range fixtures {
		count, err := i.sql("SELECT COUNT(*) FROM flowengine_flows WHERE name='"+fixture.name+"';", true)
		if err != nil {
			return err
		}
		if count == "0" {
			data, err := os.ReadFile(filepath.Join(i.Program, "fixtures", fixture.file))
			if err != nil {
				return err
			}
			jar, _ := cookiejar.New(nil)
			client := &http.Client{Jar: jar, Timeout: 30 * time.Second, Transport: &http.Transport{Proxy: nil}, CheckRedirect: func(_ *http.Request, _ []*http.Request) error { return http.ErrUseLastResponse }}
			base := localURL(i.State.HTTP)
			r, err := client.Get(base + "/login")
			if err != nil {
				return err
			}
			io.Copy(io.Discard, r.Body)
			r.Body.Close()
			if r.StatusCode != 200 {
				return fmt.Errorf("fixture login form HTTP %d", r.StatusCode)
			}
			r, err = client.PostForm(base+"/login", url.Values{"username": {"sandbox-admin"}, "password": {i.Secrets.Accounts["sandbox-admin"]}})
			if err != nil {
				return err
			}
			io.Copy(io.Discard, r.Body)
			r.Body.Close()
			if r.StatusCode != 302 {
				return fmt.Errorf("fixture login HTTP %d", r.StatusCode)
			}
			var body bytes.Buffer
			writer := multipart.NewWriter(&body)
			if err = writer.WriteField("categoryID", "108"); err != nil {
				return err
			}
			part, err := writer.CreateFormFile("flow", fixture.file)
			if err != nil {
				return err
			}
			if _, err = part.Write(data); err != nil {
				return err
			}
			if err = writer.Close(); err != nil {
				return err
			}
			r, err = client.Post(base+"/flowadmin/importflow/4", writer.FormDataContentType(), &body)
			if err != nil {
				return err
			}
			b, readErr := io.ReadAll(io.LimitReader(r.Body, 1024*1024))
			r.Body.Close()
			if readErr != nil {
				return readErr
			}
			if r.StatusCode != 302 || !regexp.MustCompile(`^/flowadmin/showflow/[0-9]+$`).MatchString(r.Header.Get("Location")) {
				return fmt.Errorf("fixture import failed: HTTP %d, response length %d", r.StatusCode, len(b))
			}
		} else if count != "1" {
			return fmt.Errorf("fixture name must be unique; found %s", strings.TrimSpace(count))
		}
	}
	if _, err = i.sql(strings.ReplaceAll(string(sqlData), "@SIMULATOR@", localURL(i.State.Simulator)), true); err != nil {
		return fmt.Errorf("fixture configuration: %w", err)
	}
	result, err := i.sql("SELECT COUNT(*) FROM booking_queries b JOIN flowengine_query_descriptors d ON b.queryID=d.queryID JOIN flowengine_steps s ON s.stepID=d.stepID JOIN flowengine_flows f ON f.flowID=s.flowID WHERE f.name='Funktionslabbet - portabelt utkast' AND b.calendarID IS NOT NULL; SELECT COUNT(*) FROM queue_queries q JOIN flowengine_query_descriptors d ON q.queryID=d.queryID JOIN flowengine_steps s ON s.stepID=d.stepID JOIN flowengine_flows f ON f.flowID=s.flowID WHERE f.name='Funktionslabbet - portabelt utkast' AND q.queueID IS NOT NULL;", true)
	if err != nil {
		return err
	}
	if strings.ReplaceAll(strings.TrimSpace(result), "\r\n", "\n") != "1\n1" {
		return fmt.Errorf("fixture links not verified: %q", result)
	}
	_, err = i.sql("REPLACE INTO sandbox_meta VALUES('fixtures_seeded','1');", true)
	return err
}
