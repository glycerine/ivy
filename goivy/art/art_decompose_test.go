package art

import (
	"testing"

	"github.com/glycerine/ivy/goivy/ast"
	"github.com/glycerine/ivy/goivy/interp"
	lg "github.com/glycerine/ivy/goivy/logic"
	"github.com/glycerine/ivy/goivy/module"
)

// ---------------------------------------------------------------------------
// helpers for building interp.InterpState chains in tests
// ---------------------------------------------------------------------------

func interpState(mod *module.Module) *interp.InterpState {
	sv := interp.NewStateValue(nil, module.TrueClauses(nil), module.FalseClauses(nil))
	return interp.NewInterpState(mod, sv, nil, "")
}

// interpChain builds a chain of n interp.States where each non-root state
// has .Expr set to an ActionApp pointing at its predecessor, and .CachedPred
// set explicitly. Returns the deepest (last) state.
func interpChain(mod *module.Module, n int) *interp.InterpState {
	if n <= 0 {
		return nil
	}
	cfg := ast.NewAstConfig()
	if mod != nil && mod.Cfg != nil && mod.Cfg.AstCfg != nil {
		cfg = mod.Cfg.AstCfg
	}
	root := interpState(mod)
	root.Label = "root"
	prev := root
	for i := 1; i < n; i++ {
		s := interpState(mod)
		s.Label = ""
		s.Expr = interp.InterpActionApp(cfg, "step", interp.WrapState(prev))
		s.SetPred(prev)
		prev = s
	}
	return prev
}

// collectArtChain walks an art.State chain from deepest to root via .Pred,
// returning them in root-first order.
func collectArtChain(s *State) []*State {
	var chain []*State
	for cur := s; cur != nil; cur = cur.Pred {
		chain = append(chain, cur)
	}
	for i, j := 0, len(chain)-1; i < j; i, j = i+1, j-1 {
		chain[i], chain[j] = chain[j], chain[i]
	}
	return chain
}

// ---------------------------------------------------------------------------
// InterpToArtState: Prov preservation
// ---------------------------------------------------------------------------

// TestInterpToArtStatePreservesProv verifies that interp.InterpState.Expr (ast.Node)
// is converted to art.State.Prov (art.Provenance) by InterpToArtState.
func TestInterpToArtStatePreservesProv(t *testing.T) {
	mod := testModule()
	cfg := ast.NewAstConfig()

	root := interpState(mod)
	child := interpState(mod)
	child.Expr = interp.InterpActionApp(cfg, "myact", interp.WrapState(root))
	child.SetPred(root)

	artChild := InterpToArtState(child)
	if artChild.Prov == nil {
		t.Fatal("art.State.Prov should not be nil when interp.InterpState.Expr was set")
	}
	aa, ok := artChild.Prov.(*ActionApp)
	if !ok {
		t.Fatalf("expected *ActionApp, got %T", artChild.Prov)
	}
	if aa.Rep != "myact" {
		t.Errorf("expected Rep='myact', got %v", aa.Rep)
	}
	if len(aa.Args) != 1 {
		t.Fatalf("expected 1 arg, got %d", len(aa.Args))
	}
}

// TestInterpToArtStateActionAppRep verifies the ActionApp Rep field carries
// the action name from the interp expression.
func TestInterpToArtStateActionAppRep(t *testing.T) {
	mod := testModule()
	cfg := ast.NewAstConfig()

	root := interpState(mod)
	child := interpState(mod)
	child.Expr = interp.InterpActionApp(cfg, "fire_missile", interp.WrapState(root))
	child.SetPred(root)

	artChild := InterpToArtState(child)
	aa := artChild.Prov.(*ActionApp)
	if aa.Rep != "fire_missile" {
		t.Errorf("expected Rep='fire_missile', got %v", aa.Rep)
	}
}

// TestInterpToArtStateNilExprGivesNilProv verifies root states with Expr==nil
// produce Prov==nil.
func TestInterpToArtStateNilExprGivesNilProv(t *testing.T) {
	mod := testModule()
	root := interpState(mod)
	// root.Expr is nil by default

	artRoot := InterpToArtState(root)
	if artRoot.Prov != nil {
		t.Errorf("root state should have Prov==nil, got %T", artRoot.Prov)
	}
}

// TestInterpToArtStateChainAllNonRootHaveProv verifies that converting
// a multi-step interp.InterpState chain populates Prov on every non-root state.
func TestInterpToArtStateChainAllNonRootHaveProv(t *testing.T) {
	mod := testModule()
	deepest := interpChain(mod, 5)

	artDeepest := InterpToArtState(deepest)
	chain := collectArtChain(artDeepest)

	if len(chain) != 5 {
		t.Fatalf("expected 5 states in chain, got %d", len(chain))
	}
	// Root (index 0) should have nil Prov
	if chain[0].Prov != nil {
		t.Errorf("root state should have nil Prov, got %T", chain[0].Prov)
	}
	// All non-root should have non-nil Prov
	for i := 1; i < len(chain); i++ {
		if chain[i].Prov == nil {
			t.Errorf("chain[%d] should have non-nil Prov", i)
		}
		if !IsActionApp(chain[i].Prov) {
			t.Errorf("chain[%d].Prov should be *ActionApp, got %T", i, chain[i].Prov)
		}
	}
}

// TestInterpToArtStateMemoIdentity verifies the same interp.InterpState pointer
// always maps to the same art.State pointer.
func TestInterpToArtStateMemoIdentity(t *testing.T) {
	mod := testModule()
	cfg := ast.NewAstConfig()

	shared := interpState(mod)
	child1 := interpState(mod)
	child1.Expr = interp.InterpActionApp(cfg, "a1", interp.WrapState(shared))
	child1.SetPred(shared)
	child2 := interpState(mod)
	child2.Expr = interp.InterpActionApp(cfg, "a2", interp.WrapState(shared))
	child2.SetPred(shared)

	memo := make(map[*interp.InterpState]*State)
	art1 := interpToArtMemo(child1, memo)
	art2 := interpToArtMemo(child2, memo)

	// The shared predecessor should be the same pointer in both
	pred1 := art1.Prov.(*ActionApp).Args[0]
	pred2 := art2.Prov.(*ActionApp).Args[0]
	if pred1 != pred2 {
		t.Error("memo should ensure shared interp.InterpState maps to same art.State")
	}
}

// TestInterpToArtStateActionName verifies ActionName is copied.
func TestInterpToArtStateActionName(t *testing.T) {
	mod := testModule()
	is := interpState(mod)
	is.ActionName = "test_action"

	s := InterpToArtState(is)
	if s.ActionName != "test_action" {
		t.Errorf("expected ActionName='test_action', got %q", s.ActionName)
	}
}

// ---------------------------------------------------------------------------
// InterpToArtState: StateJoin provenance
// ---------------------------------------------------------------------------

// TestInterpToArtStateJoinProv verifies an interp Or expression becomes
// an *art.StateJoin provenance.
func TestInterpToArtStateJoinProv(t *testing.T) {
	mod := testModule()
	cfg := ast.NewAstConfig()
	if mod != nil && mod.Cfg != nil && mod.Cfg.AstCfg != nil {
		cfg = mod.Cfg.AstCfg
	}

	s1 := interpState(mod)
	s2 := interpState(mod)
	joined := interpState(mod)
	joined.Expr = cfg.NewOr(interp.WrapState(s1), interp.WrapState(s2))
	joined.JoinOf = []*interp.InterpState{s1, s2}

	artJoined := InterpToArtState(joined)
	if artJoined.Prov == nil {
		t.Fatal("joined state should have non-nil Prov")
	}
	sj, ok := artJoined.Prov.(*StateJoin)
	if !ok {
		t.Fatalf("expected *StateJoin, got %T", artJoined.Prov)
	}
	if len(sj.Args) != 2 {
		t.Errorf("expected 2 join args, got %d", len(sj.Args))
	}
}

// ---------------------------------------------------------------------------
// ArtToInterpState: Expr preservation (reverse direction)
// ---------------------------------------------------------------------------

// TestArtToInterpStatePreservesExpr verifies art.State.Prov is converted
// to interp.InterpState.Expr.
func TestArtToInterpStatePreservesExpr(t *testing.T) {
	ag := testGraph()
	pre := testState(ag.Domain)
	ag.Add(pre, nil)
	post := testState(ag.Domain)
	post.Prov = NewActionApp("myact", pre)
	post.Pred = pre

	interpPost := ArtToInterpState(post)
	if interpPost.Expr == nil {
		t.Fatal("interp.InterpState.Expr should not be nil when art.State.Prov was set")
	}
	if !interp.IsInterpActionApp(interpPost.Expr) {
		t.Errorf("expected ActionApp ast.Node, got %T", interpPost.Expr)
	}
}

// TestArtToInterpStateNilProv verifies nil Prov produces nil Expr.
func TestArtToInterpStateNilProv(t *testing.T) {
	s := testState(testModule())
	is := ArtToInterpState(s)
	if is.Expr != nil {
		t.Errorf("nil Prov should produce nil Expr, got %T", is.Expr)
	}
}

// ---------------------------------------------------------------------------
// Round-trip: art → interp → art
// ---------------------------------------------------------------------------

// TestRoundTripPreservesProvStructure verifies art→interp→art preserves
// the Provenance type (ActionApp stays ActionApp, rep and arg count match).
func TestRoundTripPreservesProvStructure(t *testing.T) {
	ag := testGraph()
	pre := testState(ag.Domain)
	ag.Add(pre, nil)

	post := testState(ag.Domain)
	post.Prov = NewActionApp("round_trip_act", pre)
	post.Pred = pre

	// art → interp → art
	interpPost := ArtToInterpState(post)
	artPost2 := InterpToArtState(interpPost)

	if artPost2.Prov == nil {
		t.Fatal("round-trip should preserve Prov")
	}
	aa, ok := artPost2.Prov.(*ActionApp)
	if !ok {
		t.Fatalf("expected *ActionApp after round-trip, got %T", artPost2.Prov)
	}
	if aa.Rep != "round_trip_act" {
		t.Errorf("expected Rep='round_trip_act', got %v", aa.Rep)
	}
	if len(aa.Args) != 1 {
		t.Errorf("expected 1 arg, got %d", len(aa.Args))
	}
}

// TestRoundTripChain verifies a multi-step chain survives art→interp→art.
func TestRoundTripChain(t *testing.T) {
	ag := testGraph()
	s0 := testState(ag.Domain)
	ag.Add(s0, nil)

	s1 := testState(ag.Domain)
	s1.Prov = NewActionApp("step1", s0)
	s1.Pred = s0

	s2 := testState(ag.Domain)
	s2.Prov = NewActionApp("step2", s1)
	s2.Pred = s1

	// Round-trip
	interpS2 := ArtToInterpState(s2)
	artS2 := InterpToArtState(interpS2)

	chain := collectArtChain(artS2)
	if len(chain) != 3 {
		t.Fatalf("expected 3 states, got %d", len(chain))
	}
	if chain[0].Prov != nil {
		t.Error("root should have nil Prov")
	}
	for i := 1; i < len(chain); i++ {
		if chain[i].Prov == nil {
			t.Errorf("chain[%d] should have non-nil Prov after round-trip", i)
		}
	}
}

// ---------------------------------------------------------------------------
// ConstructTransitionsFromExpressions with Prov
// ---------------------------------------------------------------------------

// TestConstructTransitionsFromProvenance verifies that
// ConstructTransitionsFromExpressions uses Prov to build transitions.
func TestConstructTransitionsFromProvenance(t *testing.T) {
	mod := testModule()
	// Build an interp chain and convert to art — Prov should be set
	deepest := interpChain(mod, 3)
	artDeepest := InterpToArtState(deepest)
	chain := collectArtChain(artDeepest)

	ag := NewAnalysisGraph(mod)
	registerAction(ag, "step")
	for _, s := range chain {
		ag.Add(s, nil) // Add without transition (we'll reconstruct)
	}
	ag.Transitions = nil

	ag.ConstructTransitionsFromExpressions()
	// chain has 3 states, 2 non-root → expect 2 transitions
	if len(ag.Transitions) != 2 {
		t.Errorf("expected 2 transitions, got %d", len(ag.Transitions))
	}
	for _, tr := range ag.Transitions {
		if tr.Label != "step" {
			t.Errorf("expected label 'step', got %q", tr.Label)
		}
	}
}

// ---------------------------------------------------------------------------
// DecomposeState
// ---------------------------------------------------------------------------

// TestDecomposeStateNilReturnsNil verifies nil/empty cases.
func TestDecomposeStateNilReturnsNil(t *testing.T) {
	ag := testGraph()
	if ag.DecomposeState(nil) != nil {
		t.Error("nil state should return nil")
	}
	s := testState(ag.Domain)
	if ag.DecomposeState(s) != nil {
		t.Error("state with nil Prov should return nil")
	}
}

// TestDecomposeStateCachesSubgraph verifies second call returns cache.
func TestDecomposeStateCachesSubgraph(t *testing.T) {
	ag := testGraph()
	s := testState(ag.Domain)
	s.Prov = NewActionApp("act", testState(ag.Domain))

	// First call may return nil (no decomposition available for simple action),
	// but if it returns a subgraph, the second call should return the same one.
	sub1 := ag.DecomposeState(s)
	sub2 := ag.DecomposeState(s)
	if sub1 != nil && sub1 != sub2 {
		t.Error("second call should return cached subgraph")
	}
}

// ---------------------------------------------------------------------------
// Fuzz tests
// ---------------------------------------------------------------------------

// FuzzInterpToArtChainProv builds interp.InterpState chains of varying length
// and asserts every non-root converted art.State has non-nil Prov.
func FuzzInterpToArtChainProv(f *testing.F) {
	f.Add(uint8(1))
	f.Add(uint8(2))
	f.Add(uint8(5))
	f.Add(uint8(10))

	f.Fuzz(func(t *testing.T, chainLen uint8) {
		n := int(chainLen)%15 + 1
		mod := testModule()
		deepest := interpChain(mod, n)
		artDeepest := InterpToArtState(deepest)
		chain := collectArtChain(artDeepest)

		if len(chain) != n {
			t.Fatalf("expected %d states, got %d", n, len(chain))
		}
		if chain[0].Prov != nil {
			t.Error("root should have nil Prov")
		}
		for i := 1; i < len(chain); i++ {
			if chain[i].Prov == nil {
				t.Errorf("chain[%d] should have non-nil Prov", i)
			}
			if !IsActionApp(chain[i].Prov) {
				t.Errorf("chain[%d].Prov should be *ActionApp", i)
			}
		}
	})
}

// FuzzRoundTripChainProv builds art.State chains, converts art→interp→art,
// and asserts Prov structure is preserved at every step.
func FuzzRoundTripChainProv(f *testing.F) {
	f.Add(uint8(2))
	f.Add(uint8(5))

	f.Fuzz(func(t *testing.T, chainLen uint8) {
		n := int(chainLen)%10 + 2
		mod := testModule()

		// Build art chain
		root := testState(mod)
		prev := root
		for i := 1; i < n; i++ {
			s := testState(mod)
			s.Prov = NewActionApp("step", prev)
			s.Pred = prev
			prev = s
		}

		// Round-trip
		interpDeep := ArtToInterpState(prev)
		artDeep := InterpToArtState(interpDeep)
		chain := collectArtChain(artDeep)

		if len(chain) != n {
			t.Fatalf("expected %d states, got %d", n, len(chain))
		}
		for i := 1; i < len(chain); i++ {
			if chain[i].Prov == nil {
				t.Errorf("chain[%d] lost Prov in round-trip", i)
			}
		}
	})
}

// Ensure imports are used.
var _ = lg.True
