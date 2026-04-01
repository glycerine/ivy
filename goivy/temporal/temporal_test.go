package temporal

import (
	"strings"
	"testing"

	"github.com/glycerine/ivy/goivy/actions"
	"github.com/glycerine/ivy/goivy/ast"
	lg "github.com/glycerine/ivy/goivy/logic"
	"github.com/glycerine/ivy/goivy/module"
)

// helpers for tests
func boolConst(name string) *lg.Const {
	return lg.NewConst(name, lg.Boolean)
}

func makeActionTerm(inputs, outputs []*lg.Const, labels []string, stmt actions.Action) *ActionTerm {
	return &ActionTerm{
		Inputs:  inputs,
		Outputs: outputs,
		Labels:  labels,
		Stmt:    stmt,
	}
}

// --- ActionTerm tests ---

func TestActionTermString_NoParams(t *testing.T) {
	stmt := actions.NewSequence()
	at := makeActionTerm(nil, nil, nil, stmt)
	s := at.String()
	if !strings.HasPrefix(s, "action") {
		t.Errorf("expected ActionTerm.String() to start with 'action', got %q", s)
	}
	if !strings.Contains(s, "{") {
		t.Errorf("expected ActionTerm.String() to contain '{', got %q", s)
	}
}

func TestActionTermString_WithInputs(t *testing.T) {
	x := boolConst("x")
	stmt := actions.NewSequence()
	at := makeActionTerm([]*lg.Const{x}, nil, nil, stmt)
	s := at.String()
	if !strings.Contains(s, "x") {
		t.Errorf("expected ActionTerm.String() to mention input 'x', got %q", s)
	}
}

func TestActionTermString_WithOutputs(t *testing.T) {
	y := boolConst("y")
	stmt := actions.NewSequence()
	at := makeActionTerm(nil, []*lg.Const{y}, nil, stmt)
	s := at.String()
	if !strings.Contains(s, "returns") {
		t.Errorf("expected ActionTerm.String() to contain 'returns', got %q", s)
	}
}

func TestActionTermClone(t *testing.T) {
	stmt1 := actions.NewSequence()
	stmt2 := actions.NewAssumeAction(lg.True)
	at := makeActionTerm(nil, nil, []string{"env1"}, stmt1)
	cloned := at.Clone(stmt2)
	if cloned.Stmt != stmt2 {
		t.Error("Clone should replace the statement")
	}
	if len(cloned.Labels) != 1 || cloned.Labels[0] != "env1" {
		t.Error("Clone should preserve labels")
	}
}

// --- ActionTermBinding tests ---

func TestActionTermBindingString(t *testing.T) {
	stmt := actions.NewSequence()
	at := makeActionTerm(nil, nil, nil, stmt)
	b := &ActionTermBinding{Name: "myaction", Action: at}
	s := b.String()
	if !strings.Contains(s, "myaction") {
		t.Errorf("expected binding string to contain name, got %q", s)
	}
	if !strings.Contains(s, "=") {
		t.Errorf("expected binding string to contain '=', got %q", s)
	}
}

func TestActionTermBindingClone(t *testing.T) {
	stmt := actions.NewSequence()
	at1 := makeActionTerm(nil, nil, nil, stmt)
	at2 := makeActionTerm(nil, nil, []string{"x"}, stmt)
	b := &ActionTermBinding{Name: "test", Action: at1}
	cloned := b.Clone(at2)
	if cloned.Name != "test" {
		t.Error("Clone should preserve name")
	}
	if cloned.Action != at2 {
		t.Error("Clone should replace action")
	}
}

// --- NormalProgram tests ---

func TestNormalProgramString(t *testing.T) {
	stmt := actions.NewSequence()
	at := makeActionTerm(nil, nil, nil, stmt)
	binding := &ActionTermBinding{Name: "act1", Action: at}
	np := &NormalProgram{
		Bindings: []*ActionTermBinding{binding},
		Init:     actions.NewSequence(),
		Invars:   nil,
		Asms:     nil,
		Calls:    []string{"act1", "act2"},
	}
	s := np.String()
	if !strings.Contains(s, "let") {
		t.Error("expected NormalProgram string to contain 'let'")
	}
	if !strings.Contains(s, "while *") {
		t.Error("expected NormalProgram string to contain 'while *'")
	}
	if !strings.Contains(s, "diverge") {
		t.Error("expected NormalProgram string to contain 'diverge'")
	}
	if !strings.Contains(s, "act1") {
		t.Error("expected NormalProgram string to contain 'act1'")
	}
}

func TestNormalProgramBindingMap(t *testing.T) {
	stmt := actions.NewAssumeAction(lg.True)
	at := makeActionTerm(nil, nil, nil, stmt)
	binding := &ActionTermBinding{Name: "myact", Action: at}
	np := &NormalProgram{
		Bindings: []*ActionTermBinding{binding},
		Init:     actions.NewSequence(),
		Calls:    nil,
	}
	bm := np.BindingMap()
	if _, ok := bm["myact"]; !ok {
		t.Error("BindingMap should contain 'myact'")
	}
}

func TestNormalProgramFormulas(t *testing.T) {
	stmt := actions.NewSequence()
	at := makeActionTerm(nil, nil, nil, stmt)
	binding := &ActionTermBinding{Name: "a", Action: at}
	acfg := ast.NewAstConfig()
	lf := acfg.NewLabeledFormula(nil, lg.True)
	np := &NormalProgram{
		Bindings: []*ActionTermBinding{binding},
		Init:     actions.NewSequence(),
		Invars:   []*ast.LabeledFormula{lf},
		Asms:     nil,
		Calls:    nil,
	}
	fmlas := np.Formulas()
	// Should have binding + init + invar = 3
	if len(fmlas) != 3 {
		t.Errorf("expected 3 formulas, got %d", len(fmlas))
	}
}

// --- OldActionToNew / NewActionToOld ---

func TestOldActionToNewRoundTrip(t *testing.T) {
	act := actions.NewAssumeAction(lg.True)
	act.SetFormalParams([]*lg.Const{boolConst("p")})
	act.SetFormalReturns([]*lg.Const{boolConst("r")})
	at := OldActionToNew(act)
	if len(at.Inputs) != 1 {
		t.Errorf("expected 1 input, got %d", len(at.Inputs))
	}
	if len(at.Outputs) != 1 {
		t.Errorf("expected 1 output, got %d", len(at.Outputs))
	}
	back := NewActionToOld(at)
	if back != act {
		t.Error("roundtrip should return the original action")
	}
}

// --- NormalProgramFromModule ---

func TestNormalProgramFromModule_Empty(t *testing.T) {
	mod := module.New()
	np := NormalProgramFromModule(mod)
	if len(np.Bindings) != 0 {
		t.Error("empty module should have no bindings")
	}
	if len(np.Calls) != 0 {
		t.Error("empty module should have no calls")
	}
}

func TestNormalProgramFromModule_WithActions(t *testing.T) {
	mod := module.New()
	act := actions.NewSequence()
	mod.Actions.Set("ext:myact", act)
	mod.PublicActions.Set("ext:myact", true)
	np := NormalProgramFromModule(mod)
	if len(np.Bindings) != 1 {
		t.Errorf("expected 1 binding, got %d", len(np.Bindings))
	}
	if len(np.Calls) != 1 {
		t.Errorf("expected 1 call, got %d", len(np.Calls))
	}
}

// --- EnvAction ---

func TestEnvAction_Basic(t *testing.T) {
	stmt := actions.NewSequence()
	at := makeActionTerm(nil, nil, nil, stmt)
	b1 := &ActionTermBinding{Name: "ext:foo", Action: at}
	b2 := &ActionTermBinding{Name: "bar", Action: at}
	env := EnvAction([]*ActionTermBinding{b1, b2})
	if env == nil {
		t.Fatal("EnvAction should not be nil")
	}
	if len(env.Branches) != 2 {
		t.Errorf("expected 2 branches, got %d", len(env.Branches))
	}
}

// --- PropEvent ---

func TestPropEvent_Eventually(t *testing.T) {
	body := boolConst("phi")
	notBody, _ := lg.NewNot(body)
	ev, _ := lg.NewEventually(nil, notBody)
	loc := ast.Location{Line: 42}
	event := PropEvent(ev, loc)
	if _, ok := event.(*actions.AssertAction); !ok {
		t.Error("PropEvent for Eventually should produce AssertAction")
	}
	if event.GetLineno().Line != 42 {
		t.Errorf("expected lineno 42, got %d", event.GetLineno().Line)
	}
}

func TestPropEvent_Globally(t *testing.T) {
	body := boolConst("psi")
	g, _ := lg.NewGlobally(nil, body)
	loc := ast.Location{Line: 10}
	event := PropEvent(g, loc)
	if _, ok := event.(*actions.AssumeAction); !ok {
		t.Error("PropEvent for Globally should produce AssumeAction")
	}
}

// --- IsGloballyFormula / IsEventuallyFormula / IsTemporalFormula ---

func TestIsGloballyFormula(t *testing.T) {
	body := boolConst("x")
	g, _ := lg.NewGlobally(nil, body)
	if !IsGloballyFormula(g) {
		t.Error("should be true for Globally")
	}
	if IsGloballyFormula(body) {
		t.Error("should be false for non-Globally")
	}
}

func TestIsEventuallyFormula(t *testing.T) {
	body := boolConst("x")
	e, _ := lg.NewEventually(nil, body)
	if !IsEventuallyFormula(e) {
		t.Error("should be true for Eventually")
	}
}

func TestIsTemporalFormula(t *testing.T) {
	body := boolConst("x")
	g, _ := lg.NewGlobally(nil, body)
	e, _ := lg.NewEventually(nil, body)
	if !IsTemporalFormula(g) {
		t.Error("Globally should be temporal")
	}
	if !IsTemporalFormula(e) {
		t.Error("Eventually should be temporal")
	}
	if IsTemporalFormula(body) {
		t.Error("Const should not be temporal")
	}
}

func TestHasTemporalOperator(t *testing.T) {
	body := boolConst("x")
	g, _ := lg.NewGlobally(nil, body)
	if !HasTemporalOperator(g) {
		t.Error("Globally should have temporal operator")
	}
	if HasTemporalOperator(body) {
		t.Error("plain const should not have temporal operator")
	}
}

func TestIsGprop(t *testing.T) {
	body := boolConst("x")
	g, _ := lg.NewGlobally(nil, body)
	if !IsGprop(g) {
		t.Error("G(non-temporal) should be Gprop")
	}

	inner, _ := lg.NewGlobally(nil, body)
	outer, _ := lg.NewGlobally(nil, inner)
	if IsGprop(outer) {
		t.Error("G(G(x)) should not be Gprop (inner is temporal)")
	}
}

// --- GetEnviron / EnvironStr ---

func TestGetEnviron(t *testing.T) {
	body := boolConst("x")
	g, _ := lg.NewGlobally(nil, body)
	if GetEnviron(g) != nil {
		t.Error("nil environ should return nil")
	}

	env := "myenv"
	g2, _ := lg.NewGlobally(&env, body)
	got := GetEnviron(g2)
	if got == nil || *got != "myenv" {
		t.Error("should return the environment")
	}
}

func TestEnvironStr(t *testing.T) {
	body := boolConst("x")
	g, _ := lg.NewGlobally(nil, body)
	if EnvironStr(g) != "" {
		t.Error("nil environ should give empty string")
	}
	env := "e1"
	g2, _ := lg.NewGlobally(&env, body)
	if EnvironStr(g2) != "e1" {
		t.Errorf("expected 'e1', got %q", EnvironStr(g2))
	}
}

// --- NormalProgramClone ---

func TestNormalProgramClone(t *testing.T) {
	stmt := actions.NewSequence()
	at := makeActionTerm(nil, nil, nil, stmt)
	binding := &ActionTermBinding{Name: "a", Action: at}
	np := &NormalProgram{
		Bindings: []*ActionTermBinding{binding},
		Init:     actions.NewSequence(),
		Invars:   nil,
		Asms:     nil,
		Calls:    []string{"a"},
	}
	cloned := NormalProgramClone(np)
	// Modify the clone and verify original is unaffected
	cloned.Calls = append(cloned.Calls, "b")
	if len(np.Calls) != 1 {
		t.Error("modifying clone should not affect original")
	}
}

// --- PrefixActionTerm ---

func TestPrefixActionTerm(t *testing.T) {
	body := actions.NewSequence()
	at := makeActionTerm(nil, nil, nil, body)
	prefix := actions.NewAssumeAction(lg.True)
	result := PrefixActionTerm(at, []actions.Action{prefix})
	if result.Stmt == at.Stmt {
		t.Error("PrefixActionTerm should create a new statement")
	}
}

// --- Fuzz tests ---

func FuzzActionTermString(f *testing.F) {
	f.Add("act1", "x", "y")
	f.Fuzz(func(t *testing.T, name, inp, out string) {
		x := lg.NewConst(inp, lg.Boolean)
		y := lg.NewConst(out, lg.Boolean)
		stmt := actions.NewSequence()
		at := makeActionTerm([]*lg.Const{x}, []*lg.Const{y}, nil, stmt)
		b := &ActionTermBinding{Name: name, Action: at}
		// Just verify no panic
		_ = b.String()
		_ = at.String()
	})
}

func FuzzNormalProgramString(f *testing.F) {
	f.Add("act1", "act2")
	f.Fuzz(func(t *testing.T, call1, call2 string) {
		stmt := actions.NewSequence()
		at := makeActionTerm(nil, nil, nil, stmt)
		binding := &ActionTermBinding{Name: call1, Action: at}
		np := &NormalProgram{
			Bindings: []*ActionTermBinding{binding},
			Init:     actions.NewSequence(),
			Calls:    []string{call1, call2},
		}
		// Should not panic
		_ = np.String()
	})
}
