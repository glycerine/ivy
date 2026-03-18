package art

import (
	"testing"

	"github.com/glycerine/goivy/actions"
	"github.com/glycerine/goivy/ast"
	"github.com/glycerine/goivy/clauseops"
	lg "github.com/glycerine/goivy/logic"
	"github.com/glycerine/goivy/module"
	"github.com/glycerine/goivy/transrel"
)

// --- helpers ---

func testModule() *module.Module {
	return module.New()
}

func testClauses() *clauseops.Clauses {
	return clauseops.NewClauses([]lg.Expr{lg.True}, nil, nil)
}

func falseClauses() *clauseops.Clauses {
	return clauseops.NewClauses([]lg.Expr{lg.False}, nil, nil)
}

func testState(mod *module.Module) *State {
	return NewState(mod, testClauses())
}

func testGraph() *AnalysisGraph {
	return NewAnalysisGraph(testModule())
}

func addStates(ag *AnalysisGraph, n int) []*State {
	states := make([]*State, n)
	for i := 0; i < n; i++ {
		s := testState(ag.Domain)
		ag.Add(s, nil)
		states[i] = s
	}
	return states
}

// --- State tests ---

func TestNewState(t *testing.T) {
	mod := testModule()
	s := NewState(mod, testClauses())
	if s.ID != -1 {
		t.Errorf("expected ID -1, got %d", s.ID)
	}
	if s.Domain != mod {
		t.Error("domain mismatch")
	}
	if s.Clauses == nil {
		t.Error("clauses should not be nil")
	}
}

func TestStateIsBottom(t *testing.T) {
	mod := testModule()
	s := NewState(mod, falseClauses())
	if !s.IsBottom() {
		t.Error("expected state to be bottom")
	}
	s2 := NewState(mod, testClauses())
	if s2.IsBottom() {
		t.Error("expected state to not be bottom")
	}
}

func TestStateCopy(t *testing.T) {
	mod := testModule()
	s := NewState(mod, testClauses())
	s.ID = 42
	s.Label = "test"
	cp := s.Copy()
	if cp.ID != 42 || cp.Label != "test" {
		t.Error("copy did not preserve fields")
	}
	cp.ID = 99
	if s.ID == 99 {
		t.Error("copy should be independent")
	}
}

func TestStateString(t *testing.T) {
	s := &State{ID: 7}
	str := s.String()
	if str != "State(7)" {
		t.Errorf("expected State(7), got %s", str)
	}
}

func TestStateNilClauses(t *testing.T) {
	s := &State{ID: 0}
	if s.IsBottom() {
		t.Error("nil clauses should not be bottom")
	}
}

// --- Expression tests ---

func TestActionAppExpr(t *testing.T) {
	s := &State{ID: 0}
	expr := NewActionApp("myaction", s)
	if !IsActionApp(expr) {
		t.Error("expected ActionApp")
	}
	if IsStateJoin(expr) {
		t.Error("should not be StateJoin")
	}
	aa := expr
	if aa.Rep != "myaction" {
		t.Error("wrong rep")
	}
	if len(aa.Args) != 1 || aa.Args[0] != s {
		t.Error("wrong args")
	}
}

func TestStateJoinExpr(t *testing.T) {
	s1 := &State{ID: 0}
	s2 := &State{ID: 1}
	expr := NewStateJoin(s1, s2)
	if !IsStateJoin(expr) {
		t.Error("expected StateJoin")
	}
	if IsActionApp(expr) {
		t.Error("should not be ActionApp")
	}
	if len(expr.Args) != 2 {
		t.Error("expected 2 args")
	}
}

func TestIsActionAppNil(t *testing.T) {
	if IsActionApp(nil) {
		t.Error("nil should not be ActionApp")
	}
}

func TestIsStateJoinNil(t *testing.T) {
	if IsStateJoin(nil) {
		t.Error("nil should not be StateJoin")
	}
}

// --- Counterexample tests ---

func TestCounterexample(t *testing.T) {
	cex := &Counterexample{
		Clauses: testClauses(),
		State:   &State{ID: 3},
		Msg:     "assertion failure",
	}
	if cex.IsFailed() {
		t.Error("IsFailed should return false")
	}
	str := cex.String()
	if str != "Counterexample: assertion failure" {
		t.Errorf("unexpected string: %s", str)
	}
}

func TestCounterexampleNilState(t *testing.T) {
	cex := &Counterexample{Msg: "test"}
	if cex.IsFailed() {
		t.Error("IsFailed should return false regardless")
	}
}

// --- AC tests ---

func TestACCreation(t *testing.T) {
	ag := testGraph()
	ac := NewAC(ag, false)
	if ac.NoAdd {
		t.Error("NoAdd should be false")
	}
	if ac.Domain != ag.Domain {
		t.Error("domain mismatch")
	}
}

func TestACGet(t *testing.T) {
	ag := testGraph()
	ag.Actions["foo"] = actions.NewAssumeAction(lg.True)
	ac := NewAC(ag, false)
	if ac.Get("foo") == nil {
		t.Error("expected to find action foo")
	}
	if ac.Get("bar") != nil {
		t.Error("expected nil for unknown action")
	}
}

func TestACNewState(t *testing.T) {
	ag := testGraph()
	ac := NewAC(ag, false)
	s := ac.NewState(testClauses(), false, nil)
	if s == nil {
		t.Fatal("state should not be nil")
	}
	if len(ag.States) != 1 {
		t.Error("state should have been added to graph")
	}
}

func TestACNewStateNoAdd(t *testing.T) {
	ag := testGraph()
	ac := NewAC(ag, true)
	s := ac.NewState(testClauses(), false, nil)
	if s == nil {
		t.Fatal("state should not be nil")
	}
	if len(ag.States) != 0 {
		t.Error("state should NOT have been added to graph")
	}
}

func TestACNewStateExact(t *testing.T) {
	ag := testGraph()
	ac := NewAC(ag, true)
	s := ac.NewState(testClauses(), true, nil)
	if len(s.Unders) != 1 {
		t.Error("exact state should have 1 under-approximation")
	}
}

// --- AnalysisGraph tests ---

func TestNewAnalysisGraph(t *testing.T) {
	ag := testGraph()
	if ag.Domain == nil {
		t.Error("domain should not be nil")
	}
	if ag.StateCount() != 0 {
		t.Error("should start with 0 states")
	}
}

func TestNewAnalysisGraphNilModule(t *testing.T) {
	ag := NewAnalysisGraph(nil)
	if ag.Domain == nil {
		t.Error("nil module should result in a fresh module")
	}
}

func TestAnalysisGraphAdd(t *testing.T) {
	ag := testGraph()
	s := testState(ag.Domain)
	ag.Add(s, nil)
	if ag.StateCount() != 1 {
		t.Error("expected 1 state")
	}
	if s.ID != 0 {
		t.Errorf("expected ID 0, got %d", s.ID)
	}
}

func TestAnalysisGraphAddMultiple(t *testing.T) {
	ag := testGraph()
	states := addStates(ag, 5)
	for i, s := range states {
		if s.ID != i {
			t.Errorf("state %d has wrong ID %d", i, s.ID)
		}
	}
	if ag.StateCount() != 5 {
		t.Error("expected 5 states")
	}
}

func TestAnalysisGraphLastState(t *testing.T) {
	ag := testGraph()
	if ag.LastState() != nil {
		t.Error("empty graph should return nil")
	}
	states := addStates(ag, 3)
	if ag.LastState() != states[2] {
		t.Error("LastState should return the last state")
	}
}

func TestAnalysisGraphAddWithActionApp(t *testing.T) {
	ag := testGraph()
	pre := testState(ag.Domain)
	ag.Add(pre, nil)

	post := testState(ag.Domain)
	ag.Add(post, NewActionApp("act", pre))

	if len(ag.Transitions) != 1 {
		t.Fatalf("expected 1 transition, got %d", len(ag.Transitions))
	}
	tr := ag.Transitions[0]
	if tr.Pre != pre || tr.Post != post || tr.Label != "act" {
		t.Error("transition fields mismatch")
	}
}

func TestAnalysisGraphAddWithJoin(t *testing.T) {
	ag := testGraph()
	s1 := testState(ag.Domain)
	ag.Add(s1, nil)
	s2 := testState(ag.Domain)
	ag.Add(s2, nil)

	s3 := testState(ag.Domain)
	ag.Add(s3, NewStateJoin(s1, s2))

	if len(ag.Transitions) != 2 {
		t.Fatalf("expected 2 transitions (one per join arg), got %d", len(ag.Transitions))
	}
	for _, tr := range ag.Transitions {
		if tr.Label != "join" {
			t.Error("join transitions should have label 'join'")
		}
	}
}

func TestAnalysisGraphCover(t *testing.T) {
	ag := testGraph()
	states := addStates(ag, 2)
	ok := ag.Cover(states[0], states[1])
	if !ok {
		t.Error("cover should succeed (stubbed)")
	}
	if len(ag.Covering) != 1 {
		t.Error("expected 1 covering pair")
	}
}

func TestAnalysisGraphIsCovered(t *testing.T) {
	ag := testGraph()
	states := addStates(ag, 3)
	if ag.IsCovered(states[0]) {
		t.Error("should not be covered initially")
	}
	ag.Cover(states[0], states[1])
	if !ag.IsCovered(states[0]) {
		t.Error("should be covered after covering")
	}
	if ag.IsCovered(states[2]) {
		t.Error("state 2 should not be covered")
	}
}

func TestAnalysisGraphUnreachable(t *testing.T) {
	ag := testGraph()
	states := addStates(ag, 1)
	if ag.Unreachable(states[0]) {
		t.Error("stub should return false")
	}
}

func TestAnalysisGraphTransitionTo(t *testing.T) {
	ag := testGraph()
	pre := testState(ag.Domain)
	ag.Add(pre, nil)
	post := testState(ag.Domain)
	ag.Add(post, NewActionApp("act", pre))

	tr := ag.TransitionTo(post)
	if tr == nil {
		t.Fatal("should find transition")
	}
	if tr.Pre != pre {
		t.Error("wrong prestate")
	}

	if ag.TransitionTo(pre) != nil {
		t.Error("pre has no incoming transition")
	}
}

func TestAnalysisGraphJoin(t *testing.T) {
	ag := testGraph()
	states := addStates(ag, 2)
	joined := ag.Join(states[0], states[1], nil)
	if joined == nil {
		t.Fatal("join result should not be nil")
	}
	if ag.StateCount() != 3 {
		t.Error("joined state should be added to graph")
	}
}

func TestAnalysisGraphJoinDefaultState2(t *testing.T) {
	ag := testGraph()
	states := addStates(ag, 2)
	joined := ag.Join(states[0], nil, nil)
	if joined == nil {
		t.Fatal("join result should not be nil")
	}
	// state2 should default to last state = states[1]
}

func TestAnalysisGraphRemoveMarkedStates(t *testing.T) {
	ag := testGraph()
	states := addStates(ag, 4)
	states[1].ID = -1
	states[3].ID = -1
	ag.RemoveMarkedStates()
	if ag.StateCount() != 2 {
		t.Errorf("expected 2 states, got %d", ag.StateCount())
	}
	for i, s := range ag.States {
		if s.ID != i {
			t.Errorf("state %d has wrong ID %d after renumber", i, s.ID)
		}
	}
}

func TestAnalysisGraphReplaceState(t *testing.T) {
	ag := testGraph()
	states := addStates(ag, 2)
	states[0].Label = "original"

	replacement := testState(ag.Domain)
	replacement.Label = "replaced"
	ag.ReplaceState(states[0], replacement)
	if states[0].Label != "original" {
		t.Error("label should be preserved")
	}
	if states[0].ID != 0 {
		t.Error("ID should be preserved")
	}
}

func TestAnalysisGraphUncoveredStates(t *testing.T) {
	ag := testGraph()
	states := addStates(ag, 3)
	uncov := ag.UncoveredStates()
	if len(uncov) != 3 {
		t.Errorf("expected 3 uncovered, got %d", len(uncov))
	}
	ag.Cover(states[0], states[1])
	uncov = ag.UncoveredStates()
	if len(uncov) != 2 {
		t.Errorf("expected 2 uncovered, got %d", len(uncov))
	}
}

func TestAnalysisGraphUncoveredWithJoin(t *testing.T) {
	ag := testGraph()
	states := addStates(ag, 2)
	s3 := testState(ag.Domain)
	ag.Add(s3, NewStateJoin(states[0], states[1]))
	// states[0] and states[1] are joined into s3, so they should be
	// marked as "joined" and excluded from uncovered.
	uncov := ag.UncoveredStates()
	if len(uncov) != 1 {
		t.Errorf("expected 1 uncovered (the joined state), got %d", len(uncov))
	}
	if uncov[0] != s3 {
		t.Error("the only uncovered state should be s3")
	}
}

func TestAnalysisGraphContext(t *testing.T) {
	ag := testGraph()
	ctx := ag.Context()
	if ctx == nil {
		t.Fatal("context should not be nil")
	}
	if ctx.Domain != ag.Domain {
		t.Error("domain mismatch")
	}
}

func TestAnalysisGraphExecuteAction(t *testing.T) {
	ag := testGraph()
	act := actions.NewAssumeAction(lg.True)
	ag.Actions["test_action"] = act
	pre := testState(ag.Domain)
	ag.Add(pre, nil)

	post := ag.ExecuteAction("test_action", pre, nil)
	if post == nil {
		t.Fatal("post state should not be nil")
	}
	if ag.StateCount() != 2 {
		t.Error("should have 2 states after execute")
	}
}

func TestAnalysisGraphExecuteActionNotFound(t *testing.T) {
	ag := testGraph()
	pre := testState(ag.Domain)
	ag.Add(pre, nil)
	post := ag.ExecuteAction("nonexistent", pre, nil)
	if post != nil {
		t.Error("should return nil for unknown action")
	}
}

func TestAnalysisGraphPostState(t *testing.T) {
	ag := testGraph()
	act := actions.NewAssumeAction(lg.True)
	pre := testState(ag.Domain)
	post := ag.PostState(act, pre, nil)
	if post == nil {
		t.Fatal("post state should not be nil")
	}
	if post.Action != act {
		t.Error("action should be set")
	}
	if post.Pred != pre {
		t.Error("predecessor should be set")
	}
}

func TestAnalysisGraphPostStateWithAbstractor(t *testing.T) {
	ag := testGraph()
	act := actions.NewAssumeAction(lg.True)
	pre := testState(ag.Domain)
	called := false
	abs := AbstractorFunc(func(s *State) { called = true })
	ag.PostState(act, pre, abs)
	if !called {
		t.Error("abstractor should have been called")
	}
}

func TestAnalysisGraphCopyPath(t *testing.T) {
	ag := testGraph()
	s0 := testState(ag.Domain)
	ag.Add(s0, nil)
	s1 := testState(ag.Domain)
	s1.Pred = s0
	s1.Expr = NewActionApp("act", s0)
	ag.Add(s1, s1.Expr)

	other := NewAnalysisGraph(ag.Domain)
	result := ag.CopyPath(s1, other, nil)
	if result == nil {
		t.Fatal("copy path result should not be nil")
	}
	if other.StateCount() < 2 {
		t.Errorf("other graph should have at least 2 states, got %d", other.StateCount())
	}
}

func TestAnalysisGraphCopyPathBounded(t *testing.T) {
	ag := testGraph()
	s0 := testState(ag.Domain)
	ag.Add(s0, nil)
	s1 := testState(ag.Domain)
	s1.Pred = s0
	s1.Expr = NewActionApp("act", s0)
	ag.Add(s1, s1.Expr)

	other := NewAnalysisGraph(ag.Domain)
	bound := 0
	result := ag.CopyPath(s1, other, &bound)
	if result == nil {
		t.Fatal("copy path result should not be nil")
	}
	// With bound=0, should only copy the leaf state
	if other.StateCount() != 1 {
		t.Errorf("expected 1 state with bound=0, got %d", other.StateCount())
	}
}

func TestAnalysisGraphBMC(t *testing.T) {
	ag := testGraph()
	addStates(ag, 1)

	// BMC with error condition True on a satisfiable state should find
	// a counterexample (True AND True is SAT).
	result := ag.BMC(ag.States[0], lg.True, nil, nil)
	if result == nil {
		t.Error("BMC with True error on SAT state should find counterexample")
	}

	// BMC with error condition False should not find a counterexample
	// (True AND False is UNSAT).
	result2 := ag.BMC(ag.States[0], lg.False, nil, nil)
	if result2 != nil {
		t.Error("BMC with False error should return nil (no counterexample)")
	}
}

func TestAnalysisGraphCheckSafety(t *testing.T) {
	ag := testGraph()
	addStates(ag, 1)
	res := ag.CheckSafety(ag.States[0])
	if !res.Safe {
		t.Error("stub should return safe")
	}
}

func TestAnalysisGraphCheckBoundedSafety(t *testing.T) {
	ag := testGraph()
	addStates(ag, 1)
	res := ag.CheckBoundedSafety(ag.States[0], nil)
	if !res.Safe {
		t.Error("stub should return safe")
	}
}

func TestAnalysisGraphCallAction(t *testing.T) {
	ag := testGraph()
	pre := testState(ag.Domain)
	ag.Add(pre, nil)

	post := ag.CallAction("test", func(sub *AnalysisGraph) {
		s := testState(sub.Domain)
		sub.Add(s, nil)
	}, pre)
	if post == nil {
		t.Fatal("should get a post state")
	}
}

func TestAnalysisGraphDecomposeState(t *testing.T) {
	ag := testGraph()
	addStates(ag, 1)
	sub := ag.DecomposeState(ag.States[0])
	if sub != nil {
		t.Error("stub should return nil")
	}
}

func TestAnalysisGraphFixedpointCandidate(t *testing.T) {
	ag := testGraph()
	states := addStates(ag, 3)
	states[0].Label = "A"
	states[1].Label = "A"
	states[2].Label = "B"
	fpc := ag.FixedpointCandidate()
	if len(fpc["A"]) != 2 {
		t.Errorf("expected 2 states for label A, got %d", len(fpc["A"]))
	}
	if len(fpc["B"]) != 1 {
		t.Errorf("expected 1 state for label B, got %d", len(fpc["B"]))
	}
}

func TestAnalysisGraphGetHistory(t *testing.T) {
	ag := testGraph()
	addStates(ag, 1)
	h := ag.GetHistory(ag.States[0], nil)
	if h == nil {
		t.Fatal("history should not be nil")
	}
}

func TestAnalysisGraphRecalculate(t *testing.T) {
	ag := testGraph()
	act := actions.NewAssumeAction(lg.True)
	ag.Actions["myact"] = act
	pre := testState(ag.Domain)
	ag.Add(pre, nil)
	post := testState(ag.Domain)
	ag.Add(post, NewActionApp("myact", pre))

	tr := ag.Transitions[0]
	result := ag.Recalculate(tr, nil)
	if result == nil {
		t.Fatal("recalculate should return a state")
	}
}

// --- AnalysisSubgraph tests ---

func TestAnalysisSubgraph(t *testing.T) {
	ag := testGraph()
	sub := NewAnalysisSubgraph("myop", ag)
	if sub.Op != "myop" {
		t.Error("wrong op")
	}
	if sub.Graph != ag {
		t.Error("wrong graph")
	}
	if sub.String() != "myop" {
		t.Errorf("expected 'myop', got '%s'", sub.String())
	}
}

// --- LabelFromAction tests ---

func TestLabelFromAction(t *testing.T) {
	act := actions.NewAssumeAction(lg.True)
	label := LabelFromAction(act)
	if label == "" {
		t.Error("label should not be empty")
	}
}

// --- SafetyResult tests ---

func TestSafetyResult(t *testing.T) {
	r := &SafetyResult{Safe: true}
	if !r.Safe {
		t.Error("should be safe")
	}
	r2 := &SafetyResult{Safe: false, Cex: &Counterexample{Msg: "fail"}}
	if r2.Safe {
		t.Error("should not be safe")
	}
	if r2.Cex.Msg != "fail" {
		t.Error("wrong msg")
	}
}

// --- Transition tests ---

func TestTransition(t *testing.T) {
	pre := &State{ID: 0}
	post := &State{ID: 1}
	act := actions.NewAssumeAction(lg.True)
	tr := Transition{Pre: pre, Op: act, Label: "act", Post: post}
	if tr.Pre != pre || tr.Post != post {
		t.Error("wrong states")
	}
	if tr.Label != "act" {
		t.Error("wrong label")
	}
}

// --- CoveringPair tests ---

func TestCoveringPair(t *testing.T) {
	s1 := &State{ID: 0}
	s2 := &State{ID: 1}
	cp := CoveringPair{Covered: s1, Covering: s2}
	if cp.Covered != s1 || cp.Covering != s2 {
		t.Error("wrong states")
	}
}

// --- Integration-style tests ---

func TestGraphBuildAndCover(t *testing.T) {
	ag := testGraph()
	// Build a small graph: s0 -> s1 -> s2, cover s2 by s0
	s0 := testState(ag.Domain)
	ag.Add(s0, nil)

	s1 := testState(ag.Domain)
	ag.Add(s1, NewActionApp("step1", s0))

	s2 := testState(ag.Domain)
	ag.Add(s2, NewActionApp("step2", s1))

	ag.Cover(s2, s0)

	if !ag.IsCovered(s2) {
		t.Error("s2 should be covered")
	}
	if ag.IsCovered(s0) {
		t.Error("s0 should not be covered")
	}

	uncov := ag.UncoveredStates()
	// s0 is uncovered, s1 is uncovered (s2's arg), s2 is covered
	for _, u := range uncov {
		if u.ID == s2.ID {
			t.Error("s2 should not be in uncovered states")
		}
	}
}

func TestGraphRemoveMarkedWithTransitions(t *testing.T) {
	ag := testGraph()
	s0 := testState(ag.Domain)
	ag.Add(s0, nil)
	s1 := testState(ag.Domain)
	ag.Add(s1, NewActionApp("act", s0))
	s2 := testState(ag.Domain)
	ag.Add(s2, NewActionApp("act2", s1))

	// Mark s1 for deletion
	s1.ID = -1
	ag.RemoveMarkedStates()

	if ag.StateCount() != 2 {
		t.Errorf("expected 2 states, got %d", ag.StateCount())
	}
	// Transition from s0->s1 should be gone (s1 was deleted)
	// Transition from s1->s2 should be gone (s1 was deleted)
	if len(ag.Transitions) != 0 {
		t.Errorf("expected 0 transitions, got %d", len(ag.Transitions))
	}
}

func TestGraphExecuteAndTraverse(t *testing.T) {
	ag := testGraph()
	act := actions.NewAssumeAction(lg.True)
	ag.Actions["step"] = act
	ag.PublicActions["step"] = true

	s0 := testState(ag.Domain)
	ag.Add(s0, nil)

	s1 := ag.ExecuteAction("step", s0, nil)
	if s1 == nil {
		t.Fatal("execute should produce a state")
	}

	tr := ag.TransitionTo(s1)
	if tr == nil {
		t.Fatal("should find transition to s1")
	}
	if tr.Pre != s0 {
		t.Error("transition pre should be s0")
	}
}

// --- Fuzz tests ---

func FuzzAnalysisGraphAddRemove(f *testing.F) {
	f.Add(uint8(3), uint8(1))
	f.Add(uint8(0), uint8(0))
	f.Add(uint8(10), uint8(5))

	f.Fuzz(func(t *testing.T, numStates uint8, removeIdx uint8) {
		n := int(numStates) % 20 // cap at 20
		ag := testGraph()
		for i := 0; i < n; i++ {
			s := testState(ag.Domain)
			ag.Add(s, nil)
		}
		if ag.StateCount() != n {
			t.Errorf("expected %d states, got %d", n, ag.StateCount())
		}
		if n > 0 {
			idx := int(removeIdx) % n
			ag.States[idx].ID = -1
			ag.RemoveMarkedStates()
			if ag.StateCount() != n-1 {
				t.Errorf("expected %d states after remove, got %d", n-1, ag.StateCount())
			}
			// Check IDs are contiguous
			for i, s := range ag.States {
				if s.ID != i {
					t.Errorf("state at index %d has ID %d", i, s.ID)
				}
			}
		}
	})
}

func FuzzAnalysisGraphCoverUncover(f *testing.F) {
	f.Add(uint8(5), uint8(0), uint8(1))
	f.Add(uint8(2), uint8(0), uint8(1))

	f.Fuzz(func(t *testing.T, numStates uint8, coverIdx uint8, coverByIdx uint8) {
		n := int(numStates)%20 + 2 // at least 2
		ag := testGraph()
		for i := 0; i < n; i++ {
			s := testState(ag.Domain)
			ag.Add(s, nil)
		}
		ci := int(coverIdx) % n
		cbi := int(coverByIdx) % n
		if ci == cbi {
			cbi = (cbi + 1) % n
		}
		ag.Cover(ag.States[ci], ag.States[cbi])
		if !ag.IsCovered(ag.States[ci]) {
			t.Error("covered state should be covered")
		}
		if ci != cbi && ag.IsCovered(ag.States[cbi]) {
			// cbi might also be covered if ci covers cbi in some other pair,
			// but we only added one pair, so it shouldn't be.
			// Only check if cbi wasn't also used as covered.
		}
		uncov := ag.UncoveredStates()
		for _, u := range uncov {
			if u.ID == ci {
				t.Error("covered state should not be in uncovered list")
			}
		}
	})
}

// --- Z3-backed method tests ---

func TestCoverWithZ3Implication(t *testing.T) {
	ag := testGraph()

	// State with True clauses should be covered by another True-clauses state
	// because True => True.
	s1 := testState(ag.Domain)
	ag.Add(s1, nil)
	s2 := testState(ag.Domain)
	ag.Add(s2, nil)
	if !ag.Cover(s1, s2) {
		t.Error("True should imply True")
	}
}

func TestCoverFalseImpliesAnything(t *testing.T) {
	ag := testGraph()

	// False => anything should be true.
	s1 := NewState(ag.Domain, falseClauses())
	ag.Add(s1, nil)
	s2 := testState(ag.Domain)
	ag.Add(s2, nil)
	if !ag.Cover(s1, s2) {
		t.Error("False should imply anything (bottom covers everything)")
	}
}

func TestCoverNilClausesFails(t *testing.T) {
	ag := testGraph()
	s1 := &State{ID: -1, Domain: ag.Domain}
	ag.Add(s1, nil)
	s2 := testState(ag.Domain)
	ag.Add(s2, nil)
	if ag.Cover(s1, s2) {
		t.Error("nil clauses should not cover")
	}
}

func TestUnreachableFalseState(t *testing.T) {
	ag := testGraph()
	s := NewState(ag.Domain, falseClauses())
	ag.Add(s, nil)
	if !ag.Unreachable(s) {
		t.Error("false state should be unreachable")
	}
}

func TestUnreachableTrueState(t *testing.T) {
	ag := testGraph()
	s := testState(ag.Domain)
	ag.Add(s, nil)
	if ag.Unreachable(s) {
		t.Error("true state should not be unreachable")
	}
}

func TestUnreachableNilClauses(t *testing.T) {
	ag := testGraph()
	s := &State{ID: -1, Domain: ag.Domain}
	ag.Add(s, nil)
	if ag.Unreachable(s) {
		t.Error("nil clauses should not be unreachable")
	}
}

func TestJoinStatesDisjunction(t *testing.T) {
	ag := testGraph()
	s1 := testState(ag.Domain)
	ag.Add(s1, nil)
	s2 := NewState(ag.Domain, falseClauses())
	ag.Add(s2, nil)

	// Join of True and False should be satisfiable (True OR False = True).
	joined := ag.JoinStates(s1, s2, nil)
	if joined == nil {
		t.Fatal("joined state should not be nil")
	}
	if joined.JoinOf == nil || len(joined.JoinOf) != 2 {
		t.Error("JoinOf should have 2 entries")
	}
	if joined.Clauses == nil {
		t.Fatal("joined clauses should not be nil")
	}
}

func TestCheckSafetyNoAssertions(t *testing.T) {
	ag := testGraph()
	s := testState(ag.Domain)
	ag.Add(s, nil)
	res := ag.CheckSafety(s)
	if !res.Safe {
		t.Error("no assertions should mean safe")
	}
}

func TestCheckSafetySatisfiedAssertion(t *testing.T) {
	ag := testGraph()
	// Add an assertion that is True.
	ag.Assertions = append(ag.Assertions, &ast.LabeledFormula{
		Formula: lg.True,
	})
	s := testState(ag.Domain)
	ag.Add(s, nil)
	res := ag.CheckSafety(s)
	if !res.Safe {
		t.Error("True assertion on True state should be safe")
	}
}

func TestGetHistoryWithPredecessor(t *testing.T) {
	ag := testGraph()
	s0 := testState(ag.Domain)
	ag.Add(s0, nil)
	s1 := testState(ag.Domain)
	s1.Pred = s0
	ag.Add(s1, NewActionApp("act", s0))

	h := ag.GetHistory(s1, nil)
	if h == nil {
		t.Fatal("history should not be nil")
	}
}

func TestGetHistoryBounded(t *testing.T) {
	ag := testGraph()
	s0 := testState(ag.Domain)
	ag.Add(s0, nil)
	s1 := testState(ag.Domain)
	s1.Pred = s0
	ag.Add(s1, NewActionApp("act", s0))

	bound := 0
	h := ag.GetHistory(s1, &bound)
	if h == nil {
		t.Fatal("history should not be nil")
	}
	// With bound 0, should stop at s1 (no predecessor expansion).
	if len(h.Actions) != 0 {
		t.Error("bound=0 should produce no action steps")
	}
}

// Verify that unused imports are used.
var _ = transrel.NullUpdate
