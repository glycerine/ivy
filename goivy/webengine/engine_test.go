package webengine

import (
	"context"
	"strings"
	"testing"

	goivy "github.com/glycerine/ivy/goivy"
	"github.com/glycerine/ivy/goivy/webui"
)

const clientServerModel = `#lang ivy1.7

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

func TestEngineWrapsWebuiBackendWithTypedMethods(t *testing.T) {
	ctx := context.Background()
	engine := New(goivy.NewConfig())
	defer engine.Close()

	session, err := engine.NewSession(ctx)
	if err != nil {
		t.Fatalf("NewSession: %v", err)
	}
	if session.ID == "" {
		t.Fatal("NewSession returned empty id")
	}

	load, err := engine.LoadModel(ctx, session.ID, "client_server_example.ivy", []byte(clientServerModel))
	if err != nil {
		t.Fatalf("LoadModel: %v", err)
	}
	if load.Status != "ok" {
		t.Fatalf("load status = %q, want ok", load.Status)
	}

	check, err := engine.Check(ctx, session.ID, CheckRequest{Mode: "induction"})
	if err != nil {
		t.Fatalf("Check: %v", err)
	}
	if check.Result != "fail" {
		t.Fatalf("check result = %q, want fail", check.Result)
	}
	if check.Message == "" {
		t.Fatal("check message is empty")
	}

	arg, err := engine.ARG(ctx, session.ID)
	if err != nil {
		t.Fatalf("ARG: %v", err)
	}
	if len(elements(t, arg)) == 0 {
		t.Fatalf("ARG returned no elements: %#v", arg)
	}

	concept, err := engine.Concept(ctx, session.ID, ConceptRequest{})
	if err != nil {
		t.Fatalf("Concept: %v", err)
	}
	if _, ok := concept["toggles"].(map[string]any); !ok {
		t.Fatalf("concept toggles missing or wrong type: %#v", concept["toggles"])
	}
}

func TestEngineHonorsCancelledContextBeforeBackendWork(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	engine := New(goivy.NewConfig())
	defer engine.Close()

	if _, err := engine.NewSession(ctx); err == nil {
		t.Fatal("NewSession with cancelled context succeeded")
	}
}

func TestEngineMirrorsCheckProgressToEventSink(t *testing.T) {
	ctx := context.Background()
	engine := New(goivy.NewConfig())
	defer engine.Close()

	session, err := engine.NewSession(ctx)
	if err != nil {
		t.Fatalf("NewSession: %v", err)
	}
	if _, err := engine.LoadModel(ctx, session.ID, "client_server_example.ivy", []byte(clientServerModel)); err != nil {
		t.Fatalf("LoadModel: %v", err)
	}

	var progress []webui.Event
	if err := engine.SetEventSink(ctx, session.ID, func(event webui.Event) {
		if event.Type == "check_progress" {
			progress = append(progress, event)
		}
	}); err != nil {
		t.Fatalf("SetEventSink: %v", err)
	}

	if _, err := engine.Check(ctx, session.ID, CheckRequest{Mode: "induction"}); err != nil {
		t.Fatalf("Check: %v", err)
	}

	if len(progress) == 0 {
		t.Fatal("expected induction check_progress events mirrored to event sink")
	}
	foundConjectureProgress := false
	for _, event := range progress {
		data, _ := event.Data.(map[string]interface{})
		message, _ := data["message"].(string)
		if strings.Contains(message, "Checking conjecture 1 of 1:") {
			foundConjectureProgress = true
		}
	}
	if !foundConjectureProgress {
		t.Fatalf("missing conjecture progress event; got %#v", progress)
	}
}

func TestEngineArgActionSurfacesBackendErrors(t *testing.T) {
	ctx := context.Background()
	engine := New(goivy.NewConfig())
	defer engine.Close()

	session, err := engine.NewSession(ctx)
	if err != nil {
		t.Fatalf("NewSession: %v", err)
	}
	if _, err := engine.LoadModel(ctx, session.ID, "client_server_example.ivy", []byte(clientServerModel)); err != nil {
		t.Fatalf("LoadModel: %v", err)
	}

	_, err = engine.ArgAction(ctx, session.ID, ArgActionRequest{
		Node:   "state_0",
		Action: "execute_action",
		Args:   map[string]any{"sheet_id": "sheet-1"},
	})
	if err == nil {
		t.Fatal("ArgAction without action_name succeeded")
	}
	if got := err.Error(); got == "webengine: decode backend json: unexpected end of JSON input" {
		t.Fatalf("ArgAction returned JSON decode error instead of backend error: %v", err)
	}
	if got := err.Error(); got != "execute_action: missing action_name" {
		t.Fatalf("ArgAction error = %q, want execute_action: missing action_name", got)
	}
}

func elements(t *testing.T, payload Payload) []any {
	t.Helper()
	raw, ok := payload["elements"].([]any)
	if !ok {
		t.Fatalf("elements has type %T, want []any", payload["elements"])
	}
	return raw
}
