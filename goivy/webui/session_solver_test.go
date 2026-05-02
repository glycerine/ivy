//go:build web

package webui

import (
	"testing"

	"github.com/glycerine/ivy/goivy/module"
)

// loadTestSession creates a session loaded with ivySample (from backend_conform_test.go).
func loadTestSession(t *testing.T) *Session {
	t.Helper()
	cfg := module.NewConfig()
	s := NewSession(cfg, "test-solver")
	if err := s.LoadFileContent("test.ivy", []byte(ivySample)); err != nil {
		t.Fatalf("LoadFileContent: %v", err)
	}
	return s
}

// drainEvents reads and discards all pending events from a session.
func drainEvents(s *Session) {
	for {
		select {
		case <-s.Events:
		default:
			return
		}
	}
}

// --- Step 0: AG initialization ---

func TestSolverARGInitialized(t *testing.T) {
	s := loadTestSession(t)
	if s.AG == nil {
		t.Fatal("AG should be non-nil after LoadFileContent")
	}
	if len(s.AG.States) == 0 {
		t.Fatal("AG should have at least 1 state (initial)")
	}
	if s.Graph == nil {
		t.Fatal("Graph should be non-nil")
	}
	if len(s.Graph.States) != len(s.AG.States) {
		t.Errorf("Graph.States=%d, AG.States=%d — should match",
			len(s.Graph.States), len(s.AG.States))
	}
}

// --- Step 1: PDR Step ---

func TestSolverPDRStep(t *testing.T) {
	s := loadTestSession(t)
	drainEvents(s)
	result, err := s.ExecuteAction("pdr_step", nil)
	if err != nil {
		t.Fatalf("pdr_step error: %v", err)
	}
	if _, ok := result["valid"]; !ok {
		t.Error("result should contain 'valid' key")
	}
	if _, ok := result["stats"]; !ok {
		t.Error("result should contain 'stats' key")
	}
}

func TestSolverPDRStepNoModule(t *testing.T) {
	cfg := module.NewConfig()
	s := NewSession(cfg, "test-no-mod")
	drainEvents(s)
	_, err := s.ExecuteAction("pdr_step", nil)
	if err == nil {
		t.Error("pdr_step should fail without compiled module")
	}
}

// --- Step 2: Concrete ---

func TestSolverConcrete(t *testing.T) {
	s := loadTestSession(t)
	drainEvents(s)
	result, err := s.ExecuteAction("concrete", nil)
	if err != nil {
		t.Fatalf("concrete error: %v", err)
	}
	if _, ok := result["sat"]; !ok {
		t.Error("result should contain 'sat' key")
	}
}

func TestSolverConcreteNoModule(t *testing.T) {
	cfg := module.NewConfig()
	s := NewSession(cfg, "test-no-mod")
	drainEvents(s)
	_, err := s.ExecuteAction("concrete", nil)
	if err == nil {
		t.Error("concrete should fail without compiled module")
	}
}

// --- Step 3: Reverse ---

func TestSolverReverse(t *testing.T) {
	s := loadTestSession(t)
	drainEvents(s)
	// Execute "connect" action first to create a post-state with Update.
	_, err := s.ExecuteAction("connect", nil)
	if err != nil {
		t.Fatalf("connect action error: %v", err)
	}
	drainEvents(s)
	result, err := s.ExecuteAction("reverse", nil)
	if err != nil {
		t.Fatalf("reverse error: %v", err)
	}
	if ps, ok := result["pre_state"]; !ok || ps == "" {
		t.Error("result should contain non-empty 'pre_state'")
	}
}

func TestSolverReverseNoModule(t *testing.T) {
	cfg := module.NewConfig()
	s := NewSession(cfg, "test-no-mod")
	drainEvents(s)
	_, err := s.ExecuteAction("reverse", nil)
	if err == nil {
		t.Error("reverse should fail without compiled module")
	}
}

// --- Step 4: Reach ---

func TestSolverReach(t *testing.T) {
	s := loadTestSession(t)
	drainEvents(s)
	// Execute "connect" to create a post-state with predecessor.
	_, err := s.ExecuteAction("connect", nil)
	if err != nil {
		t.Fatalf("connect action error: %v", err)
	}
	drainEvents(s)
	result, err := s.ExecuteAction("reach", nil)
	if err != nil {
		t.Fatalf("reach error: %v", err)
	}
	if _, ok := result["reachable"]; !ok {
		t.Error("result should contain 'reachable' key")
	}
}

// --- Step 5: Weaken ---

func TestSolverWeaken(t *testing.T) {
	s := loadTestSession(t)
	drainEvents(s)
	nBefore := len(s.CompiledModule.LabeledConjs)
	if nBefore == 0 {
		t.Skip("no conjectures in sample — cannot test weaken")
	}
	args := map[string]interface{}{
		"indices": []interface{}{float64(0)},
	}
	result, err := s.ExecuteAction("weaken", args)
	if err != nil {
		t.Fatalf("weaken error: %v", err)
	}
	if rc, ok := result["removed_count"].(int); !ok || rc != 1 {
		t.Errorf("removed_count = %v, want 1", result["removed_count"])
	}
	if len(s.CompiledModule.LabeledConjs) != nBefore-1 {
		t.Errorf("LabeledConjs len = %d, want %d",
			len(s.CompiledModule.LabeledConjs), nBefore-1)
	}
}

func TestSolverWeakenNoIndices(t *testing.T) {
	s := loadTestSession(t)
	drainEvents(s)
	_, err := s.ExecuteAction("weaken", nil)
	if err == nil {
		t.Error("weaken should fail without indices")
	}
}

func TestSolverWeakenOutOfRange(t *testing.T) {
	s := loadTestSession(t)
	drainEvents(s)
	nBefore := len(s.CompiledModule.LabeledConjs)
	args := map[string]interface{}{
		"indices": []interface{}{float64(99)},
	}
	result, err := s.ExecuteAction("weaken", args)
	if err != nil {
		t.Fatalf("weaken error: %v", err)
	}
	// Index 99 is out of range — nothing removed.
	if rc, ok := result["removed_count"].(int); !ok || rc != 0 {
		t.Errorf("removed_count = %v, want 0", result["removed_count"])
	}
	if len(s.CompiledModule.LabeledConjs) != nBefore {
		t.Errorf("LabeledConjs should be unchanged: got %d, want %d",
			len(s.CompiledModule.LabeledConjs), nBefore)
	}
}

// --- Step 6: Save Abstraction ---

func TestSolverSaveAbstraction(t *testing.T) {
	s := loadTestSession(t)
	drainEvents(s)
	result, err := s.ExecuteAction("save_abstraction", nil)
	if err != nil {
		t.Fatalf("save_abstraction error: %v", err)
	}
	content, ok := result["content"].(string)
	if !ok || content == "" {
		t.Error("result should contain non-empty 'content'")
	}
	if len(s.CompiledModule.LabeledConjs) > 0 {
		// The sample has a conjecture — content should mention "invariant".
		if !contains(content, "invariant") {
			t.Errorf("content should contain 'invariant', got: %s", content)
		}
	}
}

func TestSolverSaveAbstractionNoModule(t *testing.T) {
	cfg := module.NewConfig()
	s := NewSession(cfg, "test-no-mod")
	drainEvents(s)
	_, err := s.ExecuteAction("save_abstraction", nil)
	if err == nil {
		t.Error("save_abstraction should fail without compiled module")
	}
}

func contains(s, sub string) bool {
	return len(s) >= len(sub) && (s == sub || len(s) > 0 && containsSubstr(s, sub))
}

func containsSubstr(s, sub string) bool {
	for i := 0; i+len(sub) <= len(s); i++ {
		if s[i:i+len(sub)] == sub {
			return true
		}
	}
	return false
}

// --- Step 7: Add Relation ---

func TestSolverAddRelation(t *testing.T) {
	s := loadTestSession(t)
	drainEvents(s)
	args := map[string]interface{}{
		"formula": "link(X,Y)",
	}
	result, err := s.ExecuteAction("add_relation", args)
	if err != nil {
		t.Fatalf("add_relation error: %v", err)
	}
	name, ok := result["concept_name"].(string)
	if !ok || name == "" {
		t.Error("result should contain non-empty 'concept_name'")
	}
	// Verify concept was added to domain.
	if s.SimpleSess != nil {
		if _, exists := s.SimpleSess.Domain.Concepts[name]; !exists {
			t.Error("concept should be in SimpleSess domain")
		}
	}
}

func TestSolverAddRelationParseError(t *testing.T) {
	s := loadTestSession(t)
	drainEvents(s)
	args := map[string]interface{}{
		"formula": "(((",
	}
	_, err := s.ExecuteAction("add_relation", args)
	if err == nil {
		t.Error("add_relation should fail on unparseable formula")
	}
}

func TestSolverAddRelationEmpty(t *testing.T) {
	s := loadTestSession(t)
	drainEvents(s)
	_, err := s.ExecuteAction("add_relation", map[string]interface{}{
		"formula": "",
	})
	if err == nil {
		t.Error("add_relation should fail with empty formula")
	}
}

func TestSolverAddRelationNoArgs(t *testing.T) {
	s := loadTestSession(t)
	drainEvents(s)
	_, err := s.ExecuteAction("add_relation", nil)
	if err == nil {
		t.Error("add_relation should fail with nil args")
	}
}

// --- Step 8: Action Execution ---

func TestSolverActionExecute(t *testing.T) {
	s := loadTestSession(t)
	drainEvents(s)
	nStatesBefore := len(s.AG.States)
	result, err := s.ExecuteAction("connect", nil)
	if err != nil {
		t.Fatalf("connect action error: %v", err)
	}
	if _, ok := result["post_state_id"]; !ok {
		t.Error("result should contain 'post_state_id'")
	}
	if len(s.AG.States) <= nStatesBefore {
		t.Errorf("AG should have more states after action: before=%d, after=%d",
			nStatesBefore, len(s.AG.States))
	}
	// Graph should be synced.
	if len(s.Graph.States) != len(s.AG.States) {
		t.Errorf("Graph.States=%d != AG.States=%d",
			len(s.Graph.States), len(s.AG.States))
	}
}

func TestSolverActionExecuteNotFound(t *testing.T) {
	s := loadTestSession(t)
	drainEvents(s)
	// "nonexistent" is not in module.Actions — should fall through to default.
	_, err := s.ExecuteAction("nonexistent", nil)
	if err != nil {
		t.Errorf("unknown action should not error, got: %v", err)
	}
}

func TestSolverActionExecuteNoModule(t *testing.T) {
	cfg := module.NewConfig()
	s := NewSession(cfg, "test-no-mod")
	drainEvents(s)
	// Without a compiled module, action falls through to default.
	_, err := s.ExecuteAction("connect", nil)
	if err != nil {
		t.Errorf("action without module should not error, got: %v", err)
	}
}

// --- Multi-step workflow ---

func TestSolverMultiStepWorkflow(t *testing.T) {
	s := loadTestSession(t)
	drainEvents(s)

	// Step 1: Execute "connect" to create a post-state.
	result1, err := s.ExecuteAction("connect", nil)
	if err != nil {
		t.Fatalf("step 1 connect: %v", err)
	}
	if _, ok := result1["post_state_id"]; !ok {
		t.Fatal("step 1: no post_state_id")
	}
	drainEvents(s)

	// Step 2: Try reachability on the post-state.
	result2, err := s.ExecuteAction("reach", nil)
	if err != nil {
		t.Fatalf("step 2 reach: %v", err)
	}
	if _, ok := result2["reachable"]; !ok {
		t.Error("step 2: no reachable key")
	}
	drainEvents(s)

	// Step 3: Get a concrete model.
	result3, err := s.ExecuteAction("concrete", nil)
	if err != nil {
		t.Fatalf("step 3 concrete: %v", err)
	}
	if _, ok := result3["sat"]; !ok {
		t.Error("step 3: no sat key")
	}
}
