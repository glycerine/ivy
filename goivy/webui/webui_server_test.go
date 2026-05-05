//go:build web

package webui

import (
	"encoding/json"
	goivy "github.com/glycerine/ivy/goivy"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

// --- helpers ---

func doReq(t *testing.T, srv *Server, method, path string, body string) *httptest.ResponseRecorder {
	t.Helper()
	var r *http.Request
	if body != "" {
		r = httptest.NewRequest(method, path, strings.NewReader(body))
	} else {
		r = httptest.NewRequest(method, path, nil)
	}
	w := httptest.NewRecorder()
	srv.ServeHTTP(w, r)
	return w
}

func jsonBody(t *testing.T, w *httptest.ResponseRecorder) map[string]interface{} {
	t.Helper()
	var m map[string]interface{}
	if err := json.Unmarshal(w.Body.Bytes(), &m); err != nil {
		t.Fatalf("bad json: %v\nbody: %s", err, w.Body.String())
	}
	return m
}

func createSession(t *testing.T, srv *Server) string {
	t.Helper()
	w := doReq(t, srv, "POST", "/api/session/new", "")
	if w.Code != 200 {
		t.Fatalf("new session: status %d", w.Code)
	}
	m := jsonBody(t, w)
	id, ok := m["session_id"].(string)
	if !ok || id == "" {
		t.Fatal("no session_id in response")
	}
	return id
}

// --- Server tests ---

func TestNewServer(t *testing.T) {
	cfg := goivy.NewConfig()
	srv := NewServer(cfg, ":0")
	if srv == nil {
		t.Fatal("nil server")
	}
	if srv.addr != ":0" {
		t.Errorf("addr = %q, want :0", srv.addr)
	}
}

func TestHandleIndex(t *testing.T) {
	cfg := goivy.NewConfig()
	srv := NewServer(cfg, ":0")
	w := doReq(t, srv, "GET", "/", "")
	if w.Code != 200 {
		t.Errorf("status = %d, want 200", w.Code)
	}
	ct := w.Header().Get("Content-Type")
	if !strings.Contains(ct, "text/html") {
		t.Errorf("content-type = %q", ct)
	}
	if !strings.Contains(w.Body.String(), "cytoscape") {
		t.Error("index page missing cytoscape reference")
	}
}

func TestHandleIndex404(t *testing.T) {
	cfg := goivy.NewConfig()
	srv := NewServer(cfg, ":0")
	w := doReq(t, srv, "GET", "/nonexistent", "")
	if w.Code != 404 {
		t.Errorf("status = %d, want 404", w.Code)
	}
}

func TestStaticNotFound(t *testing.T) {
	cfg := goivy.NewConfig()
	srv := NewServer(cfg, ":0")
	w := doReq(t, srv, "GET", "/static/foo.js", "")
	if w.Code != 404 {
		t.Errorf("status = %d, want 404", w.Code)
	}
}

func TestAPINewSession(t *testing.T) {
	cfg := goivy.NewConfig()
	srv := NewServer(cfg, ":0")
	id := createSession(t, srv)
	if id == "" {
		t.Error("empty session id")
	}
}

func TestAPINewSessionGETFails(t *testing.T) {
	cfg := goivy.NewConfig()
	srv := NewServer(cfg, ":0")
	w := doReq(t, srv, "GET", "/api/session/new", "")
	if w.Code != http.StatusMethodNotAllowed {
		t.Errorf("status = %d, want 405", w.Code)
	}
}

func TestAPILoadFile(t *testing.T) {
	cfg := goivy.NewConfig()
	srv := NewServer(cfg, ":0")
	id := createSession(t, srv)
	w := doReq(t, srv, "POST", "/api/session/"+id+"/load", `{"path":"test.ivy"}`)
	if w.Code != 200 {
		t.Errorf("status = %d, want 200\nbody: %s", w.Code, w.Body.String())
	}
}

func TestAPILoadFileEmpty(t *testing.T) {
	cfg := goivy.NewConfig()
	srv := NewServer(cfg, ":0")
	id := createSession(t, srv)
	w := doReq(t, srv, "POST", "/api/session/"+id+"/load", `{"path":""}`)
	if w.Code == 200 {
		t.Error("expected error for empty path")
	}
}

func TestAPILoadFileBadJSON(t *testing.T) {
	cfg := goivy.NewConfig()
	srv := NewServer(cfg, ":0")
	id := createSession(t, srv)
	w := doReq(t, srv, "POST", "/api/session/"+id+"/load", `not json`)
	if w.Code != 400 {
		t.Errorf("status = %d, want 400", w.Code)
	}
}

func TestAPIAction(t *testing.T) {
	cfg := goivy.NewConfig()
	srv := NewServer(cfg, ":0")
	id := createSession(t, srv)
	w := doReq(t, srv, "POST", "/api/session/"+id+"/action", `{"action":"check_conjectures"}`)
	if w.Code != 200 {
		t.Errorf("status = %d\nbody: %s", w.Code, w.Body.String())
	}
}

func TestAPIActionEmpty(t *testing.T) {
	cfg := goivy.NewConfig()
	srv := NewServer(cfg, ":0")
	id := createSession(t, srv)
	w := doReq(t, srv, "POST", "/api/session/"+id+"/action", `{"action":""}`)
	if w.Code == 200 {
		t.Error("expected error for empty action")
	}
}

func TestAPIARG(t *testing.T) {
	cfg := goivy.NewConfig()
	srv := NewServer(cfg, ":0")
	id := createSession(t, srv)
	w := doReq(t, srv, "GET", "/api/session/"+id+"/arg", "")
	if w.Code != 200 {
		t.Errorf("status = %d", w.Code)
	}
	m := jsonBody(t, w)
	if _, ok := m["elements"]; !ok {
		t.Error("missing elements key in ARG response")
	}
}

func TestAPIConcept(t *testing.T) {
	cfg := goivy.NewConfig()
	srv := NewServer(cfg, ":0")
	id := createSession(t, srv)
	w := doReq(t, srv, "GET", "/api/session/"+id+"/concept", "")
	if w.Code != 200 {
		t.Errorf("status = %d", w.Code)
	}
}

func TestAPIConceptSplitNotFound(t *testing.T) {
	cfg := goivy.NewConfig()
	srv := NewServer(cfg, ":0")
	id := createSession(t, srv)
	w := doReq(t, srv, "POST", "/api/session/"+id+"/concept/split", `{"concept":"x","split_by":"y"}`)
	if w.Code == 200 {
		t.Error("expected error for missing concept")
	}
}

func TestAPIConceptEmptyNotFound(t *testing.T) {
	cfg := goivy.NewConfig()
	srv := NewServer(cfg, ":0")
	id := createSession(t, srv)
	w := doReq(t, srv, "POST", "/api/session/"+id+"/concept/empty", `{"concept":"x"}`)
	if w.Code == 200 {
		t.Error("expected error for missing concept")
	}
}

func TestAPIConceptRemoveNotFound(t *testing.T) {
	cfg := goivy.NewConfig()
	srv := NewServer(cfg, ":0")
	id := createSession(t, srv)
	w := doReq(t, srv, "POST", "/api/session/"+id+"/concept/remove", `{"concept":"x"}`)
	if w.Code == 200 {
		t.Error("expected error for missing concept")
	}
}

func TestAPIConceptUndoEmpty(t *testing.T) {
	cfg := goivy.NewConfig()
	srv := NewServer(cfg, ":0")
	id := createSession(t, srv)
	w := doReq(t, srv, "POST", "/api/session/"+id+"/concept/undo", "")
	if w.Code == 200 {
		t.Error("expected error for undo with empty stack")
	}
}

func TestAPIConceptMaterializeNotFound(t *testing.T) {
	cfg := goivy.NewConfig()
	srv := NewServer(cfg, ":0")
	id := createSession(t, srv)
	w := doReq(t, srv, "POST", "/api/session/"+id+"/concept/materialize", `{"concept":"x"}`)
	if w.Code == 200 {
		t.Error("expected error for missing concept")
	}
}

func TestAPICheck(t *testing.T) {
	cfg := goivy.NewConfig()
	srv := NewServer(cfg, ":0")
	id := createSession(t, srv)
	w := doReq(t, srv, "POST", "/api/session/"+id+"/check", "")
	if w.Code != 200 {
		t.Errorf("status = %d", w.Code)
	}
	m := jsonBody(t, w)
	if m["result"] == nil {
		t.Errorf("result should not be nil, got: %v", m)
	}
	// Result should be a string indicating the check status
	if _, ok := m["result"].(string); !ok {
		t.Errorf("result should be a string, got: %T", m["result"])
	}
}

func TestAPISessionNotFound(t *testing.T) {
	cfg := goivy.NewConfig()
	srv := NewServer(cfg, ":0")
	w := doReq(t, srv, "GET", "/api/session/bogus/arg", "")
	if w.Code != 404 {
		t.Errorf("status = %d, want 404", w.Code)
	}
}

func TestAPIUnknownEndpoint(t *testing.T) {
	cfg := goivy.NewConfig()
	srv := NewServer(cfg, ":0")
	id := createSession(t, srv)
	w := doReq(t, srv, "GET", "/api/session/"+id+"/bogus", "")
	if w.Code != 404 {
		t.Errorf("status = %d, want 404", w.Code)
	}
}

func TestAPIEventsSSE(t *testing.T) {
	cfg := goivy.NewConfig()
	srv := NewServer(cfg, ":0")
	id := createSession(t, srv)

	// Push an event before connecting — access the GoBackend's session directly.
	goBE := srv.backend.(*GoBackend)
	goBE.mu.RLock()
	sess := goBE.sessions[id]
	goBE.mu.RUnlock()
	sess.emit(Event{Type: "test", Data: "hello"})

	// Use a real HTTP test server for SSE since httptest.ResponseRecorder
	// does not implement http.Flusher in all Go versions.
	ts := httptest.NewServer(srv)
	defer ts.Close()

	resp, err := http.Get(ts.URL + "/api/session/" + id + "/events")
	if err != nil {
		t.Fatal(err)
	}
	defer resp.Body.Close()

	if ct := resp.Header.Get("Content-Type"); !strings.Contains(ct, "text/event-stream") {
		t.Errorf("content-type = %q", ct)
	}

	// Read the first event.
	buf := make([]byte, 4096)
	n, err := resp.Body.Read(buf)
	if err != nil && err != io.EOF {
		t.Fatal(err)
	}
	got := string(buf[:n])
	if !strings.Contains(got, `"type":"test"`) {
		t.Errorf("SSE data = %q, want test event", got)
	}
}

func TestMultipleSessions(t *testing.T) {
	cfg := goivy.NewConfig()
	srv := NewServer(cfg, ":0")
	id1 := createSession(t, srv)
	id2 := createSession(t, srv)
	if id1 == id2 {
		t.Error("session ids should differ")
	}
	// Load different files into each.
	w1 := doReq(t, srv, "POST", "/api/session/"+id1+"/load", `{"path":"a.ivy"}`)
	w2 := doReq(t, srv, "POST", "/api/session/"+id2+"/load", `{"path":"b.ivy"}`)
	if w1.Code != 200 || w2.Code != 200 {
		t.Error("load failed")
	}
}

func TestAPIMethodNotAllowed(t *testing.T) {
	cfg := goivy.NewConfig()
	srv := NewServer(cfg, ":0")
	id := createSession(t, srv)
	// ARG is GET only.
	w := doReq(t, srv, "POST", "/api/session/"+id+"/arg", "")
	if w.Code != http.StatusMethodNotAllowed {
		t.Errorf("status = %d, want 405", w.Code)
	}
}
