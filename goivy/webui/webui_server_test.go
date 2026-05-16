//go:build web

package webui

import (
	"bytes"
	"context"
	"encoding/json"
	goivy "github.com/glycerine/ivy/goivy"
	"io"
	"mime/multipart"
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

type loadIsolateBackend struct {
	Backend
	filename string
	content  string
	isolate  string
}

func (b *loadIsolateBackend) NewSession(cfg *goivy.Config) ([]byte, error) {
	return canonicalJSON(map[string]string{"session_id": "s1"})
}

func (b *loadIsolateBackend) Load(sessionID, filename string, content []byte, isolate string) ([]byte, error) {
	b.filename = filename
	b.content = string(content)
	b.isolate = isolate
	return canonicalJSON(map[string]interface{}{
		"status":   "ok",
		"filename": filename,
		"isolate":  isolate,
		"isolates": []string{isolate},
	})
}

func TestAPILoadMultipartPassesIsolate(t *testing.T) {
	cfg := goivy.NewConfig()
	be := &loadIsolateBackend{}
	srv := NewServer(cfg, ":0", be)
	id := createSession(t, srv)

	var body bytes.Buffer
	writer := multipart.NewWriter(&body)
	if err := writer.WriteField("isolate", "cf_live"); err != nil {
		t.Fatal(err)
	}
	part, err := writer.CreateFormFile("file", "ord_live.ivy")
	if err != nil {
		t.Fatal(err)
	}
	if _, err := part.Write([]byte("#lang ivy1.7\n")); err != nil {
		t.Fatal(err)
	}
	if err := writer.Close(); err != nil {
		t.Fatal(err)
	}

	req := httptest.NewRequest("POST", "/api/session/"+id+"/load", &body)
	req.Header.Set("Content-Type", writer.FormDataContentType())
	w := httptest.NewRecorder()
	srv.ServeHTTP(w, req)
	if w.Code != 200 {
		t.Fatalf("status = %d, body: %s", w.Code, w.Body.String())
	}
	if be.filename != "ord_live.ivy" || be.isolate != "cf_live" || be.content != "#lang ivy1.7\n" {
		t.Fatalf("load payload = filename %q isolate %q content %q", be.filename, be.isolate, be.content)
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
	w := doReq(t, srv, "POST", "/api/session/"+id+"/action", `{"action":"get_conjectures"}`)
	if w.Code != 200 {
		t.Errorf("status = %d\nbody: %s", w.Code, w.Body.String())
	}
}

func TestAPIActionUnknownFails(t *testing.T) {
	cfg := goivy.NewConfig()
	srv := NewServer(cfg, ":0")
	id := createSession(t, srv)
	w := doReq(t, srv, "POST", "/api/session/"+id+"/action", `{"action":"definitely_not_a_real_action"}`)
	if w.Code == 200 {
		t.Fatalf("expected unknown action to fail, body: %s", w.Body.String())
	}
	if !strings.Contains(w.Body.String(), "unknown action") {
		t.Fatalf("unknown action response = %q, want message containing unknown action", w.Body.String())
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

type conceptNodeBackend struct {
	Backend
	nodeID  string
	sheetID string
}

func (b *conceptNodeBackend) NewSession(cfg *goivy.Config) ([]byte, error) {
	return canonicalJSON(map[string]string{"session_id": "s1"})
}

func (b *conceptNodeBackend) GetConcept(sessionID, sheetID, nodeID string) ([]byte, error) {
	b.nodeID = nodeID
	b.sheetID = sheetID
	return canonicalJSON(map[string]string{"node": nodeID, "sheet": sheetID})
}

func TestAPIConceptPassesSelectedARGNode(t *testing.T) {
	cfg := goivy.NewConfig()
	be := &conceptNodeBackend{}
	srv := NewServer(cfg, ":0", be)
	id := createSession(t, srv)
	w := doReq(t, srv, "GET", "/api/session/"+id+"/concept?node=state_1", "")
	if w.Code != 200 {
		t.Fatalf("status = %d, body: %s", w.Code, w.Body.String())
	}
	m := jsonBody(t, w)
	if got := m["node"]; got != "state_1" {
		t.Fatalf("concept node route = %v, want state_1", got)
	}
	if be.nodeID != "state_1" {
		t.Fatalf("backend saw nodeID %q, want state_1", be.nodeID)
	}
}

func TestAPIConceptPassesSheetID(t *testing.T) {
	cfg := goivy.NewConfig()
	be := &conceptNodeBackend{}
	srv := NewServer(cfg, ":0", be)
	id := createSession(t, srv)
	w := doReq(t, srv, "GET", "/api/session/"+id+"/concept?sheet=sheet-2&node=state_1", "")
	if w.Code != 200 {
		t.Fatalf("status = %d, body: %s", w.Code, w.Body.String())
	}
	m := jsonBody(t, w)
	if got := m["sheet"]; got != "sheet-2" {
		t.Fatalf("concept sheet route = %v, want sheet-2", got)
	}
	if be.sheetID != "sheet-2" {
		t.Fatalf("backend saw sheetID %q, want sheet-2", be.sheetID)
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

type materializePayloadBackend struct {
	Backend
	req ConceptMaterializeRequest
}

func (b *materializePayloadBackend) NewSession(cfg *goivy.Config) ([]byte, error) {
	return canonicalJSON(map[string]string{"session_id": "s1"})
}

func (b *materializePayloadBackend) ConceptMaterialize(sessionID string, req ConceptMaterializeRequest) ([]byte, error) {
	b.req = req
	return okJSON, nil
}

func TestAPIConceptMaterializeEdgePayload(t *testing.T) {
	cfg := goivy.NewConfig()
	be := &materializePayloadBackend{}
	srv := NewServer(cfg, ":0", be)
	id := createSession(t, srv)
	body := `{"type":"edge","relation":"link","source":"client","target":"server","positive":false}`
	w := doReq(t, srv, "POST", "/api/session/"+id+"/concept/materialize", body)
	if w.Code != 200 {
		t.Fatalf("status = %d, body: %s", w.Code, w.Body.String())
	}
	if be.req.Type != "edge" || be.req.Relation != "link" || be.req.Source != "client" || be.req.Target != "server" || be.req.Positive {
		t.Fatalf("materialize request = %#v, want negative link(client,server)", be.req)
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

type checkDefaultModeBackend struct {
	Backend
	mode string
}

func (b *checkDefaultModeBackend) NewSession(cfg *goivy.Config) ([]byte, error) {
	return canonicalJSON(map[string]string{"session_id": "s1"})
}

func (b *checkDefaultModeBackend) Check(sessionID, mode string, options CheckOptions) ([]byte, error) {
	b.mode = mode
	return canonicalJSON(map[string]string{"result": mode})
}

type checkContextBackend struct {
	Backend
	cancelled bool
}

func (b *checkContextBackend) NewSession(cfg *goivy.Config) ([]byte, error) {
	return canonicalJSON(map[string]string{"session_id": "s1"})
}

func (b *checkContextBackend) Check(sessionID, mode string, options CheckOptions) ([]byte, error) {
	b.cancelled = options.Context != nil && options.Context.Err() != nil
	return canonicalJSON(map[string]bool{"cancelled": b.cancelled})
}

func TestAPICheckDefaultsToPDR(t *testing.T) {
	cfg := goivy.NewConfig()
	be := &checkDefaultModeBackend{}
	srv := NewServer(cfg, ":0", be)
	id := createSession(t, srv)
	w := doReq(t, srv, "POST", "/api/session/"+id+"/check", "")
	if w.Code != 200 {
		t.Fatalf("status = %d, body: %s", w.Code, w.Body.String())
	}
	if be.mode != "pdr" {
		t.Fatalf("default check mode = %q, want pdr", be.mode)
	}
}

func TestAPICheckPassesRequestContext(t *testing.T) {
	cfg := goivy.NewConfig()
	be := &checkContextBackend{}
	srv := NewServer(cfg, ":0", be)
	id := createSession(t, srv)
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	req := httptest.NewRequest("POST", "/api/session/"+id+"/check", strings.NewReader(`{"mode":"induction"}`)).WithContext(ctx)
	w := httptest.NewRecorder()

	srv.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("status = %d, body: %s", w.Code, w.Body.String())
	}
	if !be.cancelled {
		t.Fatalf("backend did not receive the cancelled request context")
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
