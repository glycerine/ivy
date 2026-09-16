package goivy

import (
	"strings"
	"testing"
)

func makeNamedDomainSetup(t *testing.T) (*DomainSetup, *AstConfig, *LogicVariable) {
	t.Helper()
	cfg := NewAstConfig()
	c := newTestCompiler()
	d := NewDomainSetup(c)

	node := &UninterpretedSort{Name: "node"}
	c.Sig.AddSort(node)
	pSort, err := NewFunctionSort(node, node, Boolean)
	if err != nil {
		t.Fatalf("NewFunctionSort: %v", err)
	}
	p := NewConst("p", pSort)
	c.Sig.AddSymbol(p.Name, p.CSort)

	x, err := NewVariable("X", node)
	if err != nil {
		t.Fatalf("NewVariable X: %v", err)
	}
	y, err := NewVariable("Y", node)
	if err != nil {
		t.Fatalf("NewVariable Y: %v", err)
	}
	pyx, err := NewApply(p, y, x)
	if err != nil {
		t.Fatalf("NewApply(p, Y, X): %v", err)
	}
	d.LastFact = cfg.NewLabeledFormula(cfg.NewAtom("prop"), &LogicExists{
		Variables: []*LogicVariable{x},
		Body:      pyx,
	})
	return d, cfg, y
}

func TestDomainSetupNamedStoresAppliedSymbolForParametersLikePython(t *testing.T) {
	d, cfg, wantY := makeNamedDomainSetup(t)

	if err := d.Named(cfg.NewAtom("wit", cfg.NewVariable("Y", ""))); err != nil {
		t.Fatalf("Named: %v", err)
	}
	if len(d.Compiler.Module.Named) != 1 {
		t.Fatalf("Named entries = %d, want 1", len(d.Compiler.Module.Named))
	}
	got, ok := d.Compiler.Module.Named[0].Name.(*Apply)
	if !ok {
		t.Fatalf("Named entry term = %T, want *Apply wit(Y)", d.Compiler.Module.Named[0].Name)
	}
	if gotFunc, ok := got.Func.(*Const); !ok || gotFunc.Name != "wit" {
		t.Fatalf("Named entry function = %#v, want symbol wit", got.Func)
	}
	if len(got.Terms) != 1 || got.Terms[0] != wantY {
		t.Fatalf("Named entry args = %#v, want original free variable Y", got.Terms)
	}
}

func TestDomainSetupNamedRejectsMissingFreeParameterLikePython(t *testing.T) {
	d, cfg, _ := makeNamedDomainSetup(t)

	err := d.Named(cfg.NewAtom("wit"))
	if err == nil {
		t.Fatal("Named accepted missing free variable parameter, want Python IvyError")
	}
	if !strings.Contains(err.Error(), "Y must be a parameter of wit") {
		t.Fatalf("Named error = %v, want missing-parameter message", err)
	}
}
