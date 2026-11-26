package goivy

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestMypyvyExportsMutableRelationsAndNewState(t *testing.T) {
	const source = `#lang ivy1.8
type item
individual exemplar : item
relation ready(X:item)
after init { ready(X) := false }
action inner(x:item) = { ready(x) := true }
export action step(x:item) = { call inner(x) }
invariant [safe] ready(exemplar) | ~ready(exemplar)
`
	mod := New()
	if err := SourceString("mypyvy.ivy", source, mod, mod.Sig, map[string]interface{}{"create_isolate": false}); err != nil {
		t.Fatal(err)
	}
	if err := CreateIsolate("this", mod); err != nil {
		t.Fatal(err)
	}
	mod.Name = filepath.Join(t.TempDir(), "output")
	if err := MypyvyExport(mod); err != nil {
		t.Fatal(err)
	}
	data, err := os.ReadFile(mod.Name + ".pyv")
	if err != nil {
		t.Fatal(err)
	}
	for _, expected := range []string{
		"sort item", "mutable relation ready(item)", "immutable constant exemplar: item",
		"init (forall", "transition ext_c_step(", "modifies ready", "new(ready(", "invariant [safe]",
	} {
		if !strings.Contains(string(data), expected) {
			t.Errorf("mypyvy output is missing %q:\n%s", expected, data)
		}
	}
}

func TestMypyvyExportsEnumConstructorsAndCoverage(t *testing.T) {
	const source = `#lang ivy1.8
type color = {red, green}
individual saved : color
after init { saved := red }
export action change = { saved := green }
invariant [known_color] saved = red | saved = green
`
	mod := New()
	if err := SourceString("mypyvy_enum.ivy", source, mod, mod.Sig, map[string]interface{}{"create_isolate": false}); err != nil {
		t.Fatal(err)
	}
	if err := CreateIsolate("this", mod); err != nil {
		t.Fatal(err)
	}
	mod.Name = filepath.Join(t.TempDir(), "enum")
	if err := MypyvyExport(mod); err != nil {
		t.Fatal(err)
	}
	data, err := os.ReadFile(mod.Name + ".pyv")
	if err != nil {
		t.Fatal(err)
	}
	for _, expected := range []string{"sort color", "immutable constant red: color", "immutable constant green: color",
		"axiom !((red = green))", "forall IVYENUM:color", "(IVYENUM = red)", "(IVYENUM = green)"} {
		if !strings.Contains(string(data), expected) {
			t.Errorf("mypyvy enum output is missing %q:\n%s", expected, data)
		}
	}
}

func TestMypyvyExportOptionsUseSolverAndIntermediateDefinitions(t *testing.T) {
	mod := New()
	if err := ApplyIvyCheckParams(mod.Cfg, map[string]string{"unfold_macros": "true", "simplify": "true"}); err != nil {
		t.Fatal(err)
	}
	a := NewConst("a", Boolean)
	b := NewConst("b", Boolean)
	intermediate := NewConst("__m_intermediate", Boolean)
	next := NewConst(ActionNewName(b.Name), Boolean)
	defs := []*IvyDefinition{NewDefinition(intermediate, a), NewDefinition(next, intermediate)}
	redundant := &LogicOr{Terms: []Expr{a, b}}
	update := NewClauses([]Expr{a, redundant}, defs, nil)
	failure := NewClauses([]Expr{b}, nil, nil)
	result := mypyvyTransitionFormula(mod, update, failure)
	if UsedSymbolsAST(result).Get(Key(intermediate)) != nil || strings.Contains(result.String(), redundant.String()) {
		t.Fatalf("intermediate or redundant constraint survived: %s", result)
	}
	solver := NewSolver(mod, mod.Cfg.SolverOpts)
	defer solver.Close()
	want := &LogicAnd{Terms: []Expr{a, &LogicNot{Body: b}, next}}
	bad := &LogicNot{Body: &LogicIff{T1: result, T2: want}}
	if model, err := solver.GetModelClauses(NewClauses([]Expr{bad}, nil, nil)); err != nil || model != nil {
		t.Fatalf("export options changed transition semantics: %v", err)
	}
}
