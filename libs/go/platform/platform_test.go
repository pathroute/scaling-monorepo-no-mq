package platform

import (
	"bytes"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func env(m map[string]string) func(string) string { return func(k string) string { return m[k] } }

func TestLoadConfig(t *testing.T) {
	c, err := LoadConfig("todo", env(map[string]string{"PORT": "9000", "DEBUG": "true"}))
	if err != nil || c.Addr != ":9000" || !c.Debug || c.Service != "todo" {
		t.Fatalf("got %+v, %v", c, err)
	}
	c, err = LoadConfig("todo", env(nil))
	if err != nil || c.Addr != ":8080" || c.Debug {
		t.Fatalf("defaults: got %+v, %v", c, err)
	}
}

func TestLoadConfigRejectsBadPort(t *testing.T) {
	for _, p := range []string{"abc", "0", "70000"} {
		if _, err := LoadConfig("x", env(map[string]string{"PORT": p})); err == nil {
			t.Errorf("PORT=%s: want error", p)
		}
	}
}

func TestHealthzAndLogging(t *testing.T) {
	var buf bytes.Buffer
	h := WithLogging(NewLogger(&buf, Config{Service: "svc"}), NewMux())
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, httptest.NewRequest(http.MethodGet, "/healthz", nil))
	if rec.Code != http.StatusOK || !strings.Contains(rec.Body.String(), `"ok"`) {
		t.Fatalf("healthz: %d %s", rec.Code, rec.Body)
	}
	if !strings.Contains(buf.String(), `"service":"svc"`) || !strings.Contains(buf.String(), `"status":200`) {
		t.Fatalf("log line: %s", buf.String())
	}
}

func TestReadJSONRejectsUnknownFields(t *testing.T) {
	var v struct{ Name string }
	r := httptest.NewRequest(http.MethodPost, "/", strings.NewReader(`{"Name":"a","Extra":1}`))
	if err := ReadJSON(r, &v); err == nil {
		t.Fatal("want error for unknown field")
	}
}
