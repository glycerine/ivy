package webui

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/glycerine/ivy/goivy/module"
)

func TestArgStepInClientServerDiagnosticEdge(t *testing.T) {
	path := filepath.Join("..", "..", "ivy-lang-examples", "doc", "examples", "client_server_example.ivy")
	content, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("read example: %v", err)
	}

	s := NewSession(module.NewConfig(), "test-step-in")
	if err := s.LoadFileContent("client_server_example.ivy", content); err != nil {
		t.Fatalf("LoadFileContent: %v", err)
	}
	cr := s.RunCheck("induction")
	if cr.Result != "fail" {
		t.Fatalf("RunCheck induction result = %q, want fail; message: %s", cr.Result, cr.Message)
	}
	if cr.FailedConjecture == "" {
		t.Fatalf("RunCheck induction failed without reporting failed conjecture; message: %s", cr.Message)
	}
	if !strings.Contains(cr.FailedConjecture, "link") {
		t.Fatalf("failed conjecture %q does not mention expected relation link", cr.FailedConjecture)
	}
	if s.AG == nil || len(s.AG.Transitions) == 0 {
		t.Fatalf("induction failure did not populate ARG transitions")
	}
	if got := s.AG.Transitions[0].Label; got != "call ext" {
		t.Fatalf("ARG transition label = %q, want %q", got, "call ext")
	}

	result, err := s.ArgNodeAction("state_0", "decompose", map[string]interface{}{"target": "state_1"})
	if err != nil {
		t.Fatalf("ArgNodeAction decompose: %v", err)
	}
	if result["decomposed"] != true {
		t.Fatalf("decomposed = %v, want true; result=%#v", result["decomposed"], result)
	}
	subARG, ok := result["sub_arg"].(map[string]interface{})
	if !ok {
		t.Fatalf("sub_arg missing or wrong type: %#v", result["sub_arg"])
	}
	elements, ok := subARG["elements"].([]WebUICyElement)
	if !ok {
		t.Fatalf("sub_arg elements missing or wrong type: %#v", subARG["elements"])
	}
	if len(elements) == 0 {
		t.Fatal("sub_arg elements empty")
	}
}
