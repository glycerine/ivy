package goivy

// Tests for compilation of *ast.RequiresAction, *ast.EnsuresAction,
// *ast.SubgoalAction through CompileActionBody and CompileNode.

import (
	"testing"
)

// ---------------------------------------------------------------------------
// RequiresAction compilation
// ---------------------------------------------------------------------------

func TestCompileRequiresAction_NoProof(t *testing.T) {
	cfg := NewAstConfig()
	c := newTestCompiler()

	formula := cfg.NewAtom("true")
	reqAction := cfg.NewRequiresAction(formula)

	act, err := c.CompileActionBody(reqAction)
	if err != nil {
		t.Fatalf("CompileActionBody failed: %v", err)
	}

	ra, ok := act.(*LogicRequiresAction)
	if !ok {
		t.Fatalf("expected *actions.RequiresAction, got %T", act)
	}
	if ra.Formula == nil {
		t.Fatal("expected Formula to be set")
	}
	if ra.Proof != nil {
		t.Errorf("expected Proof to be nil, got %v", ra.Proof)
	}
}

func TestCompileRequiresAction_WithProof(t *testing.T) {
	cfg := NewAstConfig()
	c := newTestCompiler()

	formula := cfg.NewAtom("true")
	proof := cfg.NewComposeTactics(nil)
	reqAction := cfg.NewRequiresAction(formula, proof)

	act, err := c.CompileActionBody(reqAction)
	if err != nil {
		t.Fatalf("CompileActionBody failed: %v", err)
	}

	ra, ok := act.(*LogicRequiresAction)
	if !ok {
		t.Fatalf("expected *actions.RequiresAction, got %T", act)
	}
	if ra.Proof == nil {
		t.Fatal("expected Proof to be set, got nil")
	}
}

func TestCompileRequiresAction_WithLabeledFormula(t *testing.T) {
	cfg := NewAstConfig()
	c := newTestCompiler()

	innerFormula := cfg.NewAtom("true")
	label := cfg.NewAtom("my_require")
	lf := cfg.NewLabeledFormula(label, innerFormula)
	lf.Unprovable = true

	reqAction := cfg.NewRequiresAction(lf)

	act, err := c.CompileActionBody(reqAction)
	if err != nil {
		t.Fatalf("CompileActionBody failed: %v", err)
	}

	ra, ok := act.(*LogicRequiresAction)
	if !ok {
		t.Fatalf("expected *actions.RequiresAction, got %T", act)
	}
	if !ra.Unprovable {
		t.Error("expected Unprovable to be true")
	}
	if ra.LF == nil {
		t.Error("expected LF (compiled LabeledFormula) to be set")
	}
}

// ---------------------------------------------------------------------------
// EnsuresAction compilation
// ---------------------------------------------------------------------------

func TestCompileEnsuresAction_NoProof(t *testing.T) {
	cfg := NewAstConfig()
	c := newTestCompiler()

	formula := cfg.NewAtom("true")
	ensAction := cfg.NewEnsuresAction(formula)

	act, err := c.CompileActionBody(ensAction)
	if err != nil {
		t.Fatalf("CompileActionBody failed: %v", err)
	}

	ea, ok := act.(*LogicEnsuresAction)
	if !ok {
		t.Fatalf("expected *actions.EnsuresAction, got %T", act)
	}
	if ea.Formula == nil {
		t.Fatal("expected Formula to be set")
	}
	if ea.Proof != nil {
		t.Errorf("expected Proof to be nil, got %v", ea.Proof)
	}
}

func TestCompileEnsuresAction_WithProof(t *testing.T) {
	cfg := NewAstConfig()
	c := newTestCompiler()

	formula := cfg.NewAtom("true")
	proof := cfg.NewComposeTactics(nil)
	ensAction := cfg.NewEnsuresAction(formula, proof)

	act, err := c.CompileActionBody(ensAction)
	if err != nil {
		t.Fatalf("CompileActionBody failed: %v", err)
	}

	ea, ok := act.(*LogicEnsuresAction)
	if !ok {
		t.Fatalf("expected *actions.EnsuresAction, got %T", act)
	}
	if ea.Proof == nil {
		t.Fatal("expected Proof to be set, got nil")
	}
}

func TestCompileEnsuresAction_Unprovable(t *testing.T) {
	cfg := NewAstConfig()
	c := newTestCompiler()

	innerFormula := cfg.NewAtom("true")
	label := cfg.NewAtom("my_ensure")
	lf := cfg.NewLabeledFormula(label, innerFormula)
	lf.Unprovable = true

	ensAction := cfg.NewEnsuresAction(lf)

	act, err := c.CompileActionBody(ensAction)
	if err != nil {
		t.Fatalf("CompileActionBody failed: %v", err)
	}

	ea, ok := act.(*LogicEnsuresAction)
	if !ok {
		t.Fatalf("expected *actions.EnsuresAction, got %T", act)
	}
	if !ea.Unprovable {
		t.Error("expected Unprovable to be true")
	}
}

// ---------------------------------------------------------------------------
// SubgoalAction compilation
// ---------------------------------------------------------------------------

func TestCompileSubgoalAction_NoProof(t *testing.T) {
	cfg := NewAstConfig()
	c := newTestCompiler()

	formula := cfg.NewAtom("true")
	sgAction := cfg.NewSubgoalAction(formula)

	act, err := c.CompileActionBody(sgAction)
	if err != nil {
		t.Fatalf("CompileActionBody failed: %v", err)
	}

	sa, ok := act.(*LogicSubgoalAction)
	if !ok {
		t.Fatalf("expected *actions.SubgoalAction, got %T", act)
	}
	if sa.Formula == nil {
		t.Fatal("expected Formula to be set")
	}
	if sa.Proof != nil {
		t.Errorf("expected Proof to be nil, got %v", sa.Proof)
	}
}

func TestCompileSubgoalAction_WithProof(t *testing.T) {
	cfg := NewAstConfig()
	c := newTestCompiler()

	formula := cfg.NewAtom("true")
	proof := cfg.NewComposeTactics(nil)
	sgAction := cfg.NewSubgoalAction(formula, proof)

	act, err := c.CompileActionBody(sgAction)
	if err != nil {
		t.Fatalf("CompileActionBody failed: %v", err)
	}

	sa, ok := act.(*LogicSubgoalAction)
	if !ok {
		t.Fatalf("expected *actions.SubgoalAction, got %T", act)
	}
	if sa.Proof == nil {
		t.Fatal("expected Proof to be set, got nil")
	}
}

// ---------------------------------------------------------------------------
// CompileNode dispatch — verify these types route through CompileActionBody
// and do NOT fall to the default/OtherThing case.
// ---------------------------------------------------------------------------

func TestCompileNode_RequiresAction(t *testing.T) {
	cfg := NewAstConfig()
	c := newTestCompiler()

	formula := cfg.NewAtom("true")
	reqAction := cfg.NewRequiresAction(formula)

	result, err := c.CompileNode(reqAction)
	if err != nil {
		t.Fatalf("CompileNode failed: %v", err)
	}

	if _, ok := result.(*LogicRequiresAction); !ok {
		t.Fatalf("CompileNode should dispatch RequiresAction to CompileActionBody, got %T", result)
	}
}

func TestCompileNode_EnsuresAction(t *testing.T) {
	cfg := NewAstConfig()
	c := newTestCompiler()

	formula := cfg.NewAtom("true")
	ensAction := cfg.NewEnsuresAction(formula)

	result, err := c.CompileNode(ensAction)
	if err != nil {
		t.Fatalf("CompileNode failed: %v", err)
	}

	if _, ok := result.(*LogicEnsuresAction); !ok {
		t.Fatalf("CompileNode should dispatch EnsuresAction to CompileActionBody, got %T", result)
	}
}

func TestCompileNode_SubgoalAction(t *testing.T) {
	cfg := NewAstConfig()
	c := newTestCompiler()

	formula := cfg.NewAtom("true")
	sgAction := cfg.NewSubgoalAction(formula)

	result, err := c.CompileNode(sgAction)
	if err != nil {
		t.Fatalf("CompileNode failed: %v", err)
	}

	if _, ok := result.(*LogicSubgoalAction); !ok {
		t.Fatalf("CompileNode should dispatch SubgoalAction to CompileActionBody, got %T", result)
	}
}

// ---------------------------------------------------------------------------
// LocalAction with assert-like body
// ---------------------------------------------------------------------------

func TestCompileLocal_WithRequiresBody(t *testing.T) {
	cfg := NewAstConfig()
	c := newTestCompiler()

	natSort := &UninterpretedSort{Name: "nat"}
	c.Sig.Sorts.Set("nat", natSort)
	c.Sig.AddSymbol("y", natSort)

	// var x:nat := y; require true
	lhsNode := cfg.NewAtom("loc:x")
	rhsNode := cfg.NewAtom("y")
	localDecls := []Node{cfg.NewAssignAction(lhsNode, rhsNode)}

	body := cfg.NewRequiresAction(cfg.NewAtom("true"))

	result, err := c.CompileLocal(localDecls, body)
	if err != nil {
		t.Fatalf("CompileLocal failed: %v", err)
	}

	la, ok := result.(*LogicLocalAction)
	if !ok {
		t.Fatalf("expected *actions.LocalAction, got %T", result)
	}
	if len(la.Locals) == 0 {
		t.Fatal("expected at least one local variable")
	}
}

func TestCompileLocal_WithEnsuresBody(t *testing.T) {
	cfg := NewAstConfig()
	c := newTestCompiler()

	natSort := &UninterpretedSort{Name: "nat"}
	c.Sig.Sorts.Set("nat", natSort)
	c.Sig.AddSymbol("y", natSort)

	lhsNode := cfg.NewAtom("loc:x")
	rhsNode := cfg.NewAtom("y")
	localDecls := []Node{cfg.NewAssignAction(lhsNode, rhsNode)}

	body := cfg.NewEnsuresAction(cfg.NewAtom("true"))

	result, err := c.CompileLocal(localDecls, body)
	if err != nil {
		t.Fatalf("CompileLocal failed: %v", err)
	}

	la, ok := result.(*LogicLocalAction)
	if !ok {
		t.Fatalf("expected *actions.LocalAction, got %T", result)
	}
	if len(la.Locals) == 0 {
		t.Fatal("expected at least one local variable")
	}
}

// ---------------------------------------------------------------------------
// Shared helper: verify each factory produces the correct distinct type.
// ---------------------------------------------------------------------------

func TestCompileAssertLikeFormula_FactoryPreservesType(t *testing.T) {
	cfg := NewAstConfig()
	c := newTestCompiler()

	formula := cfg.NewAtom("true")

	tests := []struct {
		name     string
		compile  func(Node) (ActionsAction, error)
		wantType string
	}{
		{"Assert", c.CompileAssertFormula, "*actions.AssertAction"},
		{"Requires", c.CompileRequiresFormula, "*actions.RequiresAction"},
		{"Ensures", c.CompileEnsuresFormula, "*actions.EnsuresAction"},
		{"Subgoal", c.CompileSubgoalFormula, "*actions.SubgoalAction"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			act, err := tt.compile(formula)
			if err != nil {
				t.Fatalf("compile failed: %v", err)
			}

			switch tt.name {
			case "Assert":
				if _, ok := act.(*LogicAssertAction); !ok {
					t.Fatalf("expected %s, got %T", tt.wantType, act)
				}
			case "Requires":
				if _, ok := act.(*LogicRequiresAction); !ok {
					t.Fatalf("expected %s, got %T", tt.wantType, act)
				}
			case "Ensures":
				if _, ok := act.(*LogicEnsuresAction); !ok {
					t.Fatalf("expected %s, got %T", tt.wantType, act)
				}
			case "Subgoal":
				if _, ok := act.(*LogicSubgoalAction); !ok {
					t.Fatalf("expected %s, got %T", tt.wantType, act)
				}
			}
		})
	}
}
