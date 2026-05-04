package webui

import (
	"os"
	"path/filepath"
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
	elements, ok := subARG["elements"].([]CyElement)
	if !ok {
		t.Fatalf("sub_arg elements missing or wrong type: %#v", subARG["elements"])
	}
	if len(elements) == 0 {
		t.Fatal("sub_arg elements empty")
	}
}
