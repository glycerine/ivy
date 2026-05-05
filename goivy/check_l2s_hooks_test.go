package goivy

import (
	"strings"
	"testing"
)

// --- TemporalAndL2S ---

func TestTemporalAndL2S(t *testing.T) {
	tests := []struct {
		name string
		want bool
	}{
		{"l2s_waiting", true},
		{"l2s_frozen", true},
		{"l2s_saved", true},
		{"l2s_g", false},     // starts with l2s_g → excluded
		{"l2s_g_foo", false}, // starts with l2s_g → excluded
		{"_old_l2s", true},   // starts with _old_l2s
		{"_old_l2s_g", true}, // _old_l2s prefix matches first
		{"foo", false},
		{"", false},
		{"l2s", true},     // "l2s" prefix, not "l2s_g"
		{"l2", false},     // too short
		{"l2s_", true},    // "l2s_" but not "l2s_g"
		{"l2s_ga", false}, // starts with "l2s_g"
		{"l2s_h", true},   // starts with "l2s_" but not "l2s_g"
	}
	for _, tc := range tests {
		got := TemporalAndL2S(tc.name)
		if got != tc.want {
			t.Errorf("TemporalAndL2S(%q) = %v, want %v", tc.name, got, tc.want)
		}
	}
}

// --- applyRenamingToHandler ---

func TestApplyRenaming_NilHandler(t *testing.T) {
	// Must not panic.
	applyRenamingToHandler(nil, map[string]string{"a": "b"})
}

func TestApplyRenaming_EmptySubs(t *testing.T) {
	h := &MatchHandler{Lines: []string{"hello"}}
	applyRenamingToHandler(h, nil)
	if h.Lines[0] != "hello" {
		t.Errorf("expected unchanged, got %q", h.Lines[0])
	}
}

func TestApplyRenaming_Replace(t *testing.T) {
	h := &MatchHandler{Lines: []string{"val=_c42"}}
	applyRenamingToHandler(h, map[string]string{"_c42": "l2s_w(X)"})
	if h.Lines[0] != "val=l2s_w(X)" {
		t.Errorf("expected val=l2s_w(X), got %q", h.Lines[0])
	}
}

func TestApplyRenaming_MultiLine(t *testing.T) {
	h := &MatchHandler{Lines: []string{"_c1 and _c2", "_c2 only"}}
	applyRenamingToHandler(h, map[string]string{"_c1": "a", "_c2": "b"})
	if h.Lines[0] != "a and b" {
		t.Errorf("line 0: expected 'a and b', got %q", h.Lines[0])
	}
	if h.Lines[1] != "b only" {
		t.Errorf("line 1: expected 'b only', got %q", h.Lines[1])
	}
}

// --- helpers ---

func hooksVar(name string, s Sort) *Variable {
	v, _ := NewVariable(name, s)
	return v
}

// --- predLHSArgs ---

func TestPredLHSArgs_Apply(t *testing.T) {
	s := &UninterpretedSort{Name: "S"}
	x := hooksVar("X", s)
	y := hooksVar("Y", s)
	fs, _ := NewFunctionSort(s, s, Boolean)
	f := NewConst("f", fs)
	app := MustApply(f, x, y)
	eq := &Eq{T1: app, T2: True}

	args := predLHSArgs(eq)
	if len(args) != 2 {
		t.Fatalf("expected 2 args, got %d", len(args))
	}
	if args[0] != x || args[1] != y {
		t.Error("args don't match expected variables")
	}
}

func TestPredLHSArgs_Nil(t *testing.T) {
	args := predLHSArgs(nil)
	if args != nil {
		t.Error("expected nil for nil pred")
	}
}

func TestPredLHSArgs_ConstLHS(t *testing.T) {
	c := NewConst("c", Boolean)
	eq := &Eq{T1: c, T2: True}
	args := predLHSArgs(eq)
	if args != nil {
		t.Errorf("expected nil for Const LHS, got %v", args)
	}
}

// --- predLHSRep ---

func TestPredLHSRep_Apply(t *testing.T) {
	s := &UninterpretedSort{Name: "S"}
	fs, _ := NewFunctionSort(s, Boolean)
	f := NewConst("f", fs)
	x := hooksVar("X", s)
	app := MustApply(f, x)
	eq := &Eq{T1: app, T2: True}

	rep := predLHSRep(eq)
	if rep == nil || rep.Name != "f" {
		t.Errorf("expected f, got %v", rep)
	}
}

func TestPredLHSRep_Const(t *testing.T) {
	c := NewConst("c", Boolean)
	eq := &Eq{T1: c, T2: True}
	rep := predLHSRep(eq)
	if rep == nil || rep.Name != "c" {
		t.Errorf("expected c, got %v", rep)
	}
}

func TestPredLHSRep_Nil(t *testing.T) {
	rep := predLHSRep(nil)
	if rep != nil {
		t.Error("expected nil for nil pred")
	}
}

// --- makeSkolems ---

func TestMakeSkolems_Vars(t *testing.T) {
	s := &UninterpretedSort{Name: "S"}
	v, _ := NewVariable("X", s)
	sks := makeSkolems([]Expr{v})
	if len(sks) != 1 {
		t.Fatalf("expected 1, got %d", len(sks))
	}
	if sks[0].Name != "@X" {
		t.Errorf("expected @X, got %s", sks[0].Name)
	}
	if !sks[0].CSort.Equal(s) {
		t.Errorf("expected sort S, got %s", sks[0].CSort)
	}
}

func TestMakeSkolems_Consts(t *testing.T) {
	s := &UninterpretedSort{Name: "S"}
	c := NewConst("c", s)
	sks := makeSkolems([]Expr{c})
	if len(sks) != 1 {
		t.Fatalf("expected 1, got %d", len(sks))
	}
	if sks[0].Name != "@c" {
		t.Errorf("expected @c, got %s", sks[0].Name)
	}
}

func TestMakeSkolems_Empty(t *testing.T) {
	sks := makeSkolems(nil)
	if len(sks) != 0 {
		t.Errorf("expected empty, got %d", len(sks))
	}
}

// --- exprNameSort ---

func TestExprNameSort_Variable(t *testing.T) {
	s := &UninterpretedSort{Name: "S"}
	v, _ := NewVariable("X", s)
	name, sort := exprNameSort(v)
	if name != "X" {
		t.Errorf("expected X, got %s", name)
	}
	if !sort.Equal(s) {
		t.Errorf("expected sort S, got %s", sort)
	}
}

func TestExprNameSort_Const(t *testing.T) {
	s := &UninterpretedSort{Name: "S"}
	c := NewConst("c", s)
	name, sort := exprNameSort(c)
	if name != "c" {
		t.Errorf("expected c, got %s", name)
	}
	if !sort.Equal(s) {
		t.Errorf("expected sort S, got %s", sort)
	}
}

func TestExprNameSort_Panic(t *testing.T) {
	defer func() {
		r := recover()
		if r == nil {
			t.Fatal("expected panic for unhandled type")
		}
		msg, ok := r.(string)
		if !ok || !strings.Contains(msg, "unhandled type") {
			t.Errorf("unexpected panic message: %v", r)
		}
	}()
	exprNameSort(&And{Terms: []Expr{True}})
}

// --- evalSkolemInHandler ---

func TestEvalSkolem_NilHandler(t *testing.T) {
	sk := NewConst("@X", Boolean)
	val := evalSkolemInHandler(nil, sk)
	if val != nil {
		t.Error("expected nil for nil handler")
	}
}

func TestEvalSkolem_Found(t *testing.T) {
	sk := NewConst("@X", Boolean)
	rhs := NewConst("val42", Boolean)
	h := &MatchHandler{
		Eqs: map[NodeKey][]Expr{
			Key(sk): {&Eq{T1: sk, T2: rhs}},
		},
	}
	val := evalSkolemInHandler(h, sk)
	if val == nil {
		t.Fatal("expected non-nil value")
	}
	if c, ok := val.(*Const); !ok || c.Name != "val42" {
		t.Errorf("expected val42, got %v", val)
	}
}

func TestEvalSkolem_NotFound(t *testing.T) {
	sk := NewConst("@X", Boolean)
	h := &MatchHandler{Eqs: make(map[NodeKey][]Expr)}
	val := evalSkolemInHandler(h, sk)
	if val != nil {
		t.Error("expected nil for missing key")
	}
}

func TestEvalSkolem_EmptyEqsList(t *testing.T) {
	sk := NewConst("@X", Boolean)
	h := &MatchHandler{
		Eqs: map[NodeKey][]Expr{
			Key(sk): {},
		},
	}
	val := evalSkolemInHandler(h, sk)
	if val != nil {
		t.Error("expected nil for empty eqs list")
	}
}

// --- evalSkolems ---

func TestEvalSkolems_AllFound(t *testing.T) {
	sk1 := NewConst("@X", Boolean)
	sk2 := NewConst("@Y", Boolean)
	v1 := NewConst("a", Boolean)
	v2 := NewConst("b", Boolean)

	h := &MatchHandler{
		Eqs: map[NodeKey][]Expr{
			Key(sk1): {&Eq{T1: sk1, T2: v1}},
			Key(sk2): {&Eq{T1: sk2, T2: v2}},
		},
	}
	vals := evalSkolems(h, []*Const{sk1, sk2})
	if vals == nil {
		t.Fatal("expected non-nil values")
	}
	if len(vals) != 2 {
		t.Fatalf("expected 2 values, got %d", len(vals))
	}
}

func TestEvalSkolems_MissingOne(t *testing.T) {
	sk1 := NewConst("@X", Boolean)
	sk2 := NewConst("@Y", Boolean)
	v1 := NewConst("a", Boolean)

	h := &MatchHandler{
		Eqs: map[NodeKey][]Expr{
			Key(sk1): {&Eq{T1: sk1, T2: v1}},
			// sk2 missing
		},
	}
	vals := evalSkolems(h, []*Const{sk1, sk2})
	if vals != nil {
		t.Error("expected nil when one Skolem is missing")
	}
}

// --- applyPredToVals ---

func TestApplyPredToVals_NoVals(t *testing.T) {
	rep := NewConst("f", Boolean)
	got := applyPredToVals(rep, nil)
	if got != "f" {
		t.Errorf("expected 'f', got %q", got)
	}
}

func TestApplyPredToVals_WithVals(t *testing.T) {
	rep := NewConst("f", Boolean)
	c1 := NewConst("c1", Boolean)
	c2 := NewConst("c2", Boolean)
	got := applyPredToVals(rep, []Expr{c1, c2})
	if !strings.Contains(got, "f(") || !strings.Contains(got, "c1") || !strings.Contains(got, "c2") {
		t.Errorf("unexpected output: %q", got)
	}
}

func TestApplyPredToVals_NilRep(t *testing.T) {
	got := applyPredToVals(nil, nil)
	if got != "<nil>" {
		t.Errorf("expected '<nil>', got %q", got)
	}
}

// --- extractJusticePredMap ---

func TestExtractJusticePredMap_NilMaps(t *testing.T) {
	result := extractJusticePredMap(nil, nil, nil)
	if len(result) != 0 {
		t.Errorf("expected empty map, got %d entries", len(result))
	}
}

func TestExtractJusticePredMap_NoProgressInvar(t *testing.T) {
	rsubs := make(map[string]*NamedBinder)
	fullSubs := make(map[string]Expr)
	result := extractJusticePredMap(nil, rsubs, fullSubs)
	if len(result) != 0 {
		t.Errorf("expected empty map, got %d entries", len(result))
	}
}

// --- lfName ---

var hooksTestAstCfg = NewAstConfig()

func TestLfName_Nil(t *testing.T) {
	if got := lfName(nil); got != "" {
		t.Errorf("expected empty, got %q", got)
	}
}

func TestLfName_ConstLabel(t *testing.T) {
	label := NewConst("myname", Boolean)
	lf := hooksTestAstCfg.NewLabeledFormula(label, True)
	if got := lfName(lf); got != "myname" {
		t.Errorf("expected myname, got %q", got)
	}
}

func TestLfName_NilLabel(t *testing.T) {
	lf := hooksTestAstCfg.NewLabeledFormula(nil, True)
	if got := lfName(lf); got != "" {
		t.Errorf("expected empty, got %q", got)
	}
}

// --- Fuzz ---

func FuzzTemporalAndL2S(f *testing.F) {
	f.Add("l2s_waiting")
	f.Add("l2s_g")
	f.Add("_old_l2s")
	f.Add("")
	f.Add("foo")

	f.Fuzz(func(t *testing.T, name string) {
		_ = TemporalAndL2S(name) // must not panic
	})
}
