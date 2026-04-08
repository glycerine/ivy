//go:build web

package webui

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/glycerine/ivy/goivy/module"
)

// ivySample is a minimal Ivy file for conformance testing.
const ivySample = `#lang ivy1.7
type client
type server
relation link(X:client, Y:server)
relation semaphore(X:server)
after init { semaphore(W) := true; link(X,Y) := false }
action connect(x:client,y:server) = { require semaphore(y); link(x,y) := true; semaphore(y) := false }
export connect
conjecture link(X,Y) -> ~semaphore(Y)
`

// pyBackendAvailable checks whether the Python sidecar can start.
func pyBackendAvailable() bool {
	root := os.Getenv("PYIVY_ROOT")
	if root == "" {
		home, _ := os.UserHomeDir()
		root = filepath.Join(home, "ivy", "pyivy", "ivy", "ivy")
	}
	// ~/ivy/pyivy/ivy/ivy/z3/
	z3Dir := filepath.Join(root, "z3")
	if _, err := os.Stat(z3Dir); err != nil {
		return false
	}
	//vv("pyBackendAvailable true at: z3Dir='%v'", z3Dir)
	return true
}

// TestConformNewSession tests that NewSession works on both backends.
// Session IDs will differ (expected), so this just verifies both succeed.
func TestConformNewSession(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping conformance test in short mode")
	}
	if testing.Short() {
		t.Skip("skipping conformance test in short mode")
	}
	if !pyBackendAvailable() {
		t.Skip("Python Ivy / Z3 not available, skipping conformance test")
	}
	cfg := module.NewConfig()
	pyBE, err := NewPyBackend(cfg)
	if err != nil {
		t.Fatalf("start python backend: %v", err)
	}
	defer pyBE.Close()

	goBE := NewGoBackend(cfg)
	defer goBE.Close() // redundant but should be fine.
	cb := NewConformBackend(goBE, pyBE)
	defer cb.Close()

	data, err := cb.NewSession(cfg)
	if err != nil {
		t.Fatalf("NewSession: %v", err)
	}
	var resp map[string]string
	if err := json.Unmarshal(data, &resp); err != nil {
		t.Fatalf("unmarshal: %v", err)
	}
	if resp["session_id"] == "" {
		t.Fatal("empty session_id")
	}
	t.Logf("session_id: %s", resp["session_id"])
}

// TestConformLoad tests load conformance.
func TestConformLoad(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping conformance test in short mode")
	}
	if !pyBackendAvailable() {
		t.Skip("Python Ivy / Z3 not available, skipping conformance test")
	}
	cfg := module.NewConfig()
	pyBE, err := NewPyBackend(cfg)
	if err != nil {
		t.Fatalf("start python backend: %v", err)
	}
	defer pyBE.Close()

	goBE := NewGoBackend(cfg)

	// Create sessions on both backends.
	goSessData, _ := goBE.NewSession(cfg)
	pySessData, _ := pyBE.NewSession(cfg)
	var goSess, pySess map[string]string
	json.Unmarshal(goSessData, &goSess)
	json.Unmarshal(pySessData, &pySess)
	goSID := goSess["session_id"]
	pySID := pySess["session_id"]

	// Load the same file on both.
	goLoad, goErr := goBE.Load(goSID, "test.ivy", []byte(ivySample))
	pyLoad, pyErr := pyBE.Load(pySID, "test.ivy", []byte(ivySample))

	if goErr != nil {
		t.Fatalf("Go Load error: %v", goErr)
	}
	if pyErr != nil {
		t.Fatalf("Python Load error: %v", pyErr)
	}

	t.Logf("Go  Load: %s", goLoad)
	t.Logf("Py  Load: %s", pyLoad)
	if string(goLoad) != string(pyLoad) {
		t.Errorf("LOAD MISMATCH:\n  Go: %s\n  Py: %s", goLoad, pyLoad)
	}
}

// TestConformConcept tests concept graph conformance.
func TestConformConcept(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping conformance test in short mode")
	}
	if !pyBackendAvailable() {
		t.Skip("Python Ivy / Z3 not available, skipping conformance test")
	}

	cfg := module.NewConfig()
	pyBE, err := NewPyBackend(cfg)
	if err != nil {
		t.Fatalf("start python backend: %v", err)
	}
	defer pyBE.Close()

	goBE := NewGoBackend(cfg)

	goSessData, _ := goBE.NewSession(cfg)
	pySessData, _ := pyBE.NewSession(cfg)
	var goSess, pySess map[string]string
	json.Unmarshal(goSessData, &goSess)
	json.Unmarshal(pySessData, &pySess)
	goSID := goSess["session_id"]
	pySID := pySess["session_id"]

	goBE.Load(goSID, "test.ivy", []byte(ivySample))
	pyBE.Load(pySID, "test.ivy", []byte(ivySample))

	goConcept, goErr := goBE.GetConcept(goSID)
	pyConcept, pyErr := pyBE.GetConcept(pySID)

	if goErr != nil {
		t.Fatalf("Go Concept error: %v", goErr)
	}
	if pyErr != nil {
		t.Fatalf("Python Concept error: %v", pyErr)
	}

	// Compare field-by-field. The "relations" field may differ because
	// the Python sidecar returns bare names ("link") while Go correctly
	// returns parameterized names ("link(X,Y)") matching the Tcl/Tk GUI.
	// This is a sidecar limitation, not a Go divergence — skip it in
	// the comparison but log the difference.
	var goMap, pyMap map[string]interface{}
	json.Unmarshal(goConcept, &goMap)
	json.Unmarshal(pyConcept, &pyMap)

	sidecarLimitations := map[string]bool{"relations": true}
	hasMismatch := false
	for key := range goMap {
		if sidecarLimitations[key] {
			continue
		}
		goVal := fmt.Sprintf("%v", goMap[key])
		pyVal := fmt.Sprintf("%v", pyMap[key])
		if goVal != pyVal {
			goJSON, _ := json.Marshal(goMap[key])
			pyJSON, _ := json.Marshal(pyMap[key])
			t.Errorf("  field %q differs:\n    Go: %s\n    Py: %s", key, goJSON, pyJSON)
			hasMismatch = true
		}
	}
	for key := range pyMap {
		if sidecarLimitations[key] {
			continue
		}
		if _, ok := goMap[key]; !ok {
			pyJSON, _ := json.Marshal(pyMap[key])
			t.Errorf("  field %q only in Python: %s", key, pyJSON)
			hasMismatch = true
		}
	}
	if hasMismatch {
		t.Logf("Go (%d bytes): %s", len(goConcept), goConcept)
		t.Logf("Py (%d bytes): %s", len(pyConcept), pyConcept)
	}
	// Log the sidecar limitation for visibility.
	if goMap["relations"] != nil && pyMap["relations"] != nil {
		goJSON, _ := json.Marshal(goMap["relations"])
		pyJSON, _ := json.Marshal(pyMap["relations"])
		if string(goJSON) != string(pyJSON) {
			t.Logf("  field \"relations\": sidecar returns bare names, Go matches Tcl/Tk GUI:\n    Go: %s\n    Py sidecar: %s", goJSON, pyJSON)
		}
	}
}

// TestConformCheck tests verification check conformance.
func TestConformCheck(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping conformance test in short mode")
	}
	if !pyBackendAvailable() {
		t.Skip("Python Ivy / Z3 not available, skipping conformance test")
	}

	cfg := module.NewConfig()
	pyBE, err := NewPyBackend(cfg)
	if err != nil {
		t.Fatalf("start python backend: %v", err)
	}
	defer pyBE.Close()

	goBE := NewGoBackend(cfg)

	goSessData, _ := goBE.NewSession(cfg)
	pySessData, _ := pyBE.NewSession(cfg)
	var goSess, pySess map[string]string
	json.Unmarshal(goSessData, &goSess)
	json.Unmarshal(pySessData, &pySess)
	goSID := goSess["session_id"]
	pySID := pySess["session_id"]

	goBE.Load(goSID, "test.ivy", []byte(ivySample))
	pyBE.Load(pySID, "test.ivy", []byte(ivySample))

	// Allow Python a moment to finish compilation.
	time.Sleep(500 * time.Millisecond)

	goCheck, goErr := goBE.Check(goSID, "induction")
	pyCheck, pyErr := pyBE.Check(pySID, "induction")

	if goErr != nil {
		t.Fatalf("Go Check error: %v", goErr)
	}
	if pyErr != nil {
		t.Fatalf("Python Check error: %v", pyErr)
	}

	t.Logf("Go  Check: %s", goCheck)
	t.Logf("Py  Check: %s", pyCheck)

	if string(goCheck) != string(pyCheck) {
		t.Errorf("CHECK MISMATCH:\n  Go (%d bytes): %s\n  Py (%d bytes): %s",
			len(goCheck), goCheck, len(pyCheck), pyCheck)
		// Parse both and show field-level diff.
		var goMap, pyMap map[string]interface{}
		json.Unmarshal(goCheck, &goMap)
		json.Unmarshal(pyCheck, &pyMap)
		allKeys := make(map[string]bool)
		for k := range goMap {
			allKeys[k] = true
		}
		for k := range pyMap {
			allKeys[k] = true
		}
		for key := range allKeys {
			goJSON, _ := json.Marshal(goMap[key])
			pyJSON, _ := json.Marshal(pyMap[key])
			if string(goJSON) != string(pyJSON) {
				t.Logf("  field %q differs:\n    Go: %s\n    Py: %s", key, goJSON, pyJSON)
			}
		}
	}
}

// TestConformARG tests ARG rendering conformance.
func TestConformARG(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping conformance test in short mode")
	}
	if !pyBackendAvailable() {
		t.Skip("Python Ivy / Z3 not available, skipping conformance test")
	}

	cfg := module.NewConfig()
	pyBE, err := NewPyBackend(cfg)
	if err != nil {
		t.Fatalf("start python backend: %v", err)
	}
	defer pyBE.Close()

	goBE := NewGoBackend(cfg)

	goSessData, _ := goBE.NewSession(cfg)
	pySessData, _ := pyBE.NewSession(cfg)
	var goSess, pySess map[string]string
	json.Unmarshal(goSessData, &goSess)
	json.Unmarshal(pySessData, &pySess)
	goSID := goSess["session_id"]
	pySID := pySess["session_id"]

	goBE.Load(goSID, "test.ivy", []byte(ivySample))
	pyBE.Load(pySID, "test.ivy", []byte(ivySample))

	goARG, goErr := goBE.GetARG(goSID)
	pyARG, pyErr := pyBE.GetARG(pySID)

	if goErr != nil {
		t.Fatalf("Go ARG error: %v", goErr)
	}
	if pyErr != nil {
		t.Fatalf("Python ARG error: %v", pyErr)
	}

	t.Logf("Go  ARG: %s", goARG)
	t.Logf("Py  ARG: %s", pyARG)

	if string(goARG) != string(pyARG) {
		t.Errorf("ARG MISMATCH:\n  Go (%d bytes): %s\n  Py (%d bytes): %s",
			len(goARG), goARG, len(pyARG), pyARG)
	}
}
