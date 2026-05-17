package webui

import (
	"bytes"
	"encoding/json"
	goivy "github.com/glycerine/ivy/goivy"
	"io"
	"mime/multipart"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

const hostedContractClientServer = `#lang ivy1.7

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

func TestHostedWebUIContract(t *testing.T) {
	cfg := goivy.NewConfig()
	server := httptest.NewServer(NewServer(cfg, ":0"))
	defer server.Close()

	session := hostedPostJSON[map[string]string](t, server.URL+"/api/session/new", nil)
	sessionID := session["session_id"]
	if sessionID == "" {
		t.Fatalf("missing session_id in %#v", session)
	}

	load := hostedPostMultipart(t, server.URL+"/api/session/"+sessionID+"/load", "client_server_example.ivy", hostedContractClientServer)
	if load["status"] != "ok" {
		t.Fatalf("load status = %#v, want ok; body=%#v", load["status"], load)
	}
	if load["isolate"] != NoIsolatesFoundChoice {
		t.Fatalf("load isolate = %#v, want %q; body=%#v", load["isolate"], NoIsolatesFoundChoice, load)
	}

	check := hostedPostJSON[map[string]any](t, server.URL+"/api/session/"+sessionID+"/check", strings.NewReader(`{"mode":"induction"}`))
	if check["result"] != "fail" {
		t.Fatalf("check result = %#v, want fail; body=%#v", check["result"], check)
	}
	if check["message"] == "" {
		t.Fatalf("check message missing: %#v", check)
	}
	if _, ok := check["z3_contacted"].(bool); !ok {
		t.Fatalf("z3_contacted missing or not bool: %#v", check["z3_contacted"])
	}

	arg := hostedGetJSON[map[string]any](t, server.URL+"/api/session/"+sessionID+"/arg")
	if len(hostedElements(t, arg)) == 0 {
		t.Fatalf("arg elements missing/empty: %#v", arg["elements"])
	}

	concept := hostedGetJSON[map[string]any](t, server.URL+"/api/session/"+sessionID+"/concept")
	if len(hostedStringSlice(t, concept["relations"])) == 0 {
		t.Fatalf("concept relations missing/empty: %#v", concept["relations"])
	}
	if _, ok := concept["toggles"].(map[string]any); !ok {
		t.Fatalf("concept toggles missing or wrong type: %#v", concept["toggles"])
	}
}

func hostedPostJSON[T any](t *testing.T, url string, body io.Reader) T {
	t.Helper()
	req, err := http.NewRequest(http.MethodPost, url, body)
	if err != nil {
		t.Fatalf("new request: %v", err)
	}
	if body != nil {
		req.Header.Set("Content-Type", "application/json")
	}
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		t.Fatalf("post %s: %v", url, err)
	}
	defer resp.Body.Close()
	return hostedDecodeJSON[T](t, resp)
}

func hostedPostMultipart(t *testing.T, url, filename, content string) map[string]any {
	t.Helper()
	var body bytes.Buffer
	writer := multipart.NewWriter(&body)
	part, err := writer.CreateFormFile("file", filename)
	if err != nil {
		t.Fatalf("create form file: %v", err)
	}
	if _, err := io.Copy(part, strings.NewReader(content)); err != nil {
		t.Fatalf("write form file: %v", err)
	}
	if err := writer.Close(); err != nil {
		t.Fatalf("close multipart writer: %v", err)
	}
	req, err := http.NewRequest(http.MethodPost, url, &body)
	if err != nil {
		t.Fatalf("new request: %v", err)
	}
	req.Header.Set("Content-Type", writer.FormDataContentType())
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		t.Fatalf("post multipart %s: %v", url, err)
	}
	defer resp.Body.Close()
	return hostedDecodeJSON[map[string]any](t, resp)
}

func hostedGetJSON[T any](t *testing.T, url string) T {
	t.Helper()
	resp, err := http.Get(url)
	if err != nil {
		t.Fatalf("get %s: %v", url, err)
	}
	defer resp.Body.Close()
	return hostedDecodeJSON[T](t, resp)
}

func hostedDecodeJSON[T any](t *testing.T, resp *http.Response) T {
	t.Helper()
	data, err := io.ReadAll(resp.Body)
	if err != nil {
		t.Fatalf("read response: %v", err)
	}
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		t.Fatalf("%s returned %d: %s", resp.Request.URL, resp.StatusCode, data)
	}
	var out T
	if err := json.Unmarshal(data, &out); err != nil {
		t.Fatalf("decode json: %v\nbody: %s", err, data)
	}
	return out
}

func hostedElements(t *testing.T, payload map[string]any) []any {
	t.Helper()
	elements, ok := payload["elements"].([]any)
	if !ok {
		t.Fatalf("elements has type %T, want []any", payload["elements"])
	}
	return elements
}

func hostedStringSlice(t *testing.T, value any) []string {
	t.Helper()
	raw, ok := value.([]any)
	if !ok {
		t.Fatalf("value has type %T, want []any", value)
	}
	out := make([]string, 0, len(raw))
	for _, item := range raw {
		text, ok := item.(string)
		if !ok {
			t.Fatalf("slice item has type %T, want string", item)
		}
		out = append(out, text)
	}
	return out
}
