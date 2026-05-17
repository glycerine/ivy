package server

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

const engineRouteClientServer = `#lang ivy1.7

type client
type server

relation link(X:client, Y:server)
relation semaphore(X:server)

after init {
    semaphore(W) := true;
    link(X,Y) := false
}

action connect(x:client,y:server) = {
    require semaphore(y);
    link(x,y) := true;
    semaphore(y) := false
}

action disconnect(x:client,y:server) = {
    require link(x,y);
    link(x,y) := false;
    semaphore(y) := true
}

invariant ~(X ~= Z & link(X,Y) & link(Z,Y))

export connect
export disconnect
`

func TestEngineRoutesRequireAuthenticationAndCSRF(t *testing.T) {
	srv := newEngineTestServer(t)
	project := srv.projects.CreatePersonalProject("dev-user", "Demo", "demo")

	anonymous := httptest.NewRecorder()
	srv.ServeHTTP(anonymous, httptest.NewRequest(http.MethodPost, "/api/engine/session", strings.NewReader(`{"projectId":"`+project.ID+`"}`)))
	if anonymous.Code != http.StatusForbidden {
		t.Fatalf("anonymous/missing csrf status = %d, want csrf rejection", anonymous.Code)
	}

	missingCSRF := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodPost, "/api/engine/session", strings.NewReader(`{"projectId":"`+project.ID+`"}`))
	req.AddCookie(&http.Cookie{Name: sessionCookieName, Value: "user:dev-user"})
	srv.ServeHTTP(missingCSRF, req)
	if missingCSRF.Code != http.StatusForbidden {
		t.Fatalf("missing csrf status = %d, want forbidden", missingCSRF.Code)
	}
}

func TestEngineRoutesCreateLoadCheckAndSnapshot(t *testing.T) {
	srv := newEngineTestServer(t)
	defer srv.engine.Close()
	project := srv.projects.CreatePersonalProject("dev-user", "Demo", "demo")

	session := enginePost(t, srv, "/api/engine/session", `{"projectId":"`+project.ID+`"}`)
	sessionBody := decodeMap(t, session)
	sessionPayload := sessionBody["session"].(map[string]any)
	sessionID := sessionPayload["id"].(string)
	if sessionPayload["projectId"] != project.ID {
		t.Fatalf("projectId = %#v, want %s", sessionPayload["projectId"], project.ID)
	}

	load := enginePost(t, srv, "/api/engine/session/"+sessionID+"/load", `{"model":{"id":"model-1","projectId":"`+project.ID+`","filename":"client_server_example.ivy","text":`+quoteJSON(t, engineRouteClientServer)+`,"engineRevision":1}}`)
	if decodeMap(t, load)["job"].(map[string]any)["kind"] != "load" {
		t.Fatalf("load body = %s", load.Body.String())
	}

	check := enginePost(t, srv, "/api/engine/session/"+sessionID+"/command", `{"intent":{"id":"intent-1","commandId":"check.induction"}}`)
	checkBody := decodeMap(t, check)
	if checkBody["job"].(map[string]any)["kind"] != "check-induction" {
		t.Fatalf("check body = %s", check.Body.String())
	}
	if checkBody["result"].(map[string]any)["result"] != "fail" {
		t.Fatalf("check result body = %s", check.Body.String())
	}

	snapshot := engineGet(t, srv, "/api/engine/session/"+sessionID+"/snapshot")
	snapshotBody := decodeMap(t, snapshot)
	if snapshotBody["arg"] == nil || snapshotBody["concept"] == nil {
		t.Fatalf("snapshot body missing arg/concept: %s", snapshot.Body.String())
	}
}

func TestEngineRoutesEnforceProjectAccess(t *testing.T) {
	srv := newEngineTestServer(t)
	defer srv.engine.Close()
	project := srv.projects.CreatePersonalProject("owner", "Secret", "secret")

	req := httptest.NewRequest(http.MethodPost, "/api/engine/session", strings.NewReader(`{"projectId":"`+project.ID+`"}`))
	addEngineAuth(req)
	rec := httptest.NewRecorder()
	srv.ServeHTTP(rec, req)

	if rec.Code != http.StatusForbidden {
		t.Fatalf("status = %d, want forbidden", rec.Code)
	}
}

func newEngineTestServer(t *testing.T) *Server {
	t.Helper()
	srv, err := New(Config{DevMode: true})
	if err != nil {
		t.Fatal(err)
	}
	return srv
}

func enginePost(t *testing.T, srv *Server, path, body string) *httptest.ResponseRecorder {
	t.Helper()
	req := httptest.NewRequest(http.MethodPost, path, strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	addEngineAuth(req)
	rec := httptest.NewRecorder()
	srv.ServeHTTP(rec, req)
	if rec.Code != http.StatusOK {
		t.Fatalf("POST %s status = %d body=%s", path, rec.Code, rec.Body.String())
	}
	return rec
}

func engineGet(t *testing.T, srv *Server, path string) *httptest.ResponseRecorder {
	t.Helper()
	req := httptest.NewRequest(http.MethodGet, path, nil)
	req.AddCookie(&http.Cookie{Name: sessionCookieName, Value: "user:dev-user"})
	rec := httptest.NewRecorder()
	srv.ServeHTTP(rec, req)
	if rec.Code != http.StatusOK {
		t.Fatalf("GET %s status = %d body=%s", path, rec.Code, rec.Body.String())
	}
	return rec
}

func addEngineAuth(req *http.Request) {
	req.AddCookie(&http.Cookie{Name: sessionCookieName, Value: "user:dev-user"})
	req.AddCookie(&http.Cookie{Name: "ivysvk_csrf", Value: "csrf"})
	req.Header.Set("x-csrf-token", "csrf")
}

func decodeMap(t *testing.T, rec *httptest.ResponseRecorder) map[string]any {
	t.Helper()
	var body map[string]any
	if err := json.Unmarshal(rec.Body.Bytes(), &body); err != nil {
		t.Fatalf("decode response: %v\nbody=%s", err, rec.Body.String())
	}
	return body
}

func quoteJSON(t *testing.T, value string) string {
	t.Helper()
	encoded, err := json.Marshal(value)
	if err != nil {
		t.Fatal(err)
	}
	return string(encoded)
}
