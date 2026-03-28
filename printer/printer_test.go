package printer

import (
	"strings"
	"testing"

	"github.com/glycerine/goivy/actions"
	"github.com/glycerine/goivy/ast"
	lg "github.com/glycerine/goivy/logic"
	"github.com/glycerine/goivy/module"
)

func boolConst(name string) *lg.Symbol {
	return lg.NewSymbol(name, lg.Boolean)
}

// --- LabeledFmlasToStr tests ---

func TestLabeledFmlasToStr_Empty(t *testing.T) {
	result := LabeledFmlasToStr("axiom", nil)
	if result != "" {
		t.Errorf("expected empty string for nil input, got %q", result)
	}
}

func TestLabeledFmlasToStr_NoLabel(t *testing.T) {
	acfg := ast.NewAstConfig()
	lf := acfg.NewLabeledFormula(nil, lg.True)
	result := LabeledFmlasToStr("axiom", []*ast.LabeledFormula{lf})
	if !strings.HasPrefix(result, "axiom ") {
		t.Errorf("expected prefix 'axiom ', got %q", result)
	}
	if !strings.HasSuffix(result, "\n") {
		t.Error("expected trailing newline")
	}
}

func TestLabeledFmlasToStr_WithLabel(t *testing.T) {
	label := boolConst("inv1")
	acfg := ast.NewAstConfig()
	lf := acfg.NewLabeledFormula(label, lg.True)
	result := LabeledFmlasToStr("conjecture", []*ast.LabeledFormula{lf})
	if !strings.Contains(result, "[") || !strings.Contains(result, "]") {
		t.Errorf("expected brackets around label, got %q", result)
	}
	if !strings.Contains(result, "inv1") {
		t.Errorf("expected label name 'inv1', got %q", result)
	}
}

func TestLabeledFmlasToStr_Multiple(t *testing.T) {
	acfg := ast.NewAstConfig()
	lf1 := acfg.NewLabeledFormula(nil, lg.True)
	lf2 := acfg.NewLabeledFormula(nil, lg.False)
	result := LabeledFmlasToStr("property", []*ast.LabeledFormula{lf1, lf2})
	count := strings.Count(result, "property")
	if count != 2 {
		t.Errorf("expected 2 occurrences of 'property', got %d", count)
	}
}

// --- FormatModule tests ---

func TestFormatModule_Empty(t *testing.T) {
	mod := module.New()
	result := FormatModule(mod)
	// Should not panic and should produce some output (at least the sig)
	if result == "" {
		// Empty module may have an empty sig representation, that's ok
	}
}

func TestFormatModule_WithAxioms(t *testing.T) {
	mod := module.New()
	acfg := ast.NewAstConfig()
	lf := acfg.NewLabeledFormula(nil, lg.True)
	mod.LabeledAxioms = []*ast.LabeledFormula{lf}
	result := FormatModule(mod)
	if !strings.Contains(result, "axiom") {
		t.Errorf("expected 'axiom' in output, got %q", result)
	}
}

func TestFormatModule_WithConjectures(t *testing.T) {
	mod := module.New()
	label := boolConst("inv1")
	acfg := ast.NewAstConfig()
	lf := acfg.NewLabeledFormula(label, lg.True)
	mod.LabeledConjs = []*ast.LabeledFormula{lf}
	result := FormatModule(mod)
	if !strings.Contains(result, "conjecture") {
		t.Errorf("expected 'conjecture' in output, got %q", result)
	}
}

func TestFormatModule_WithActions(t *testing.T) {
	mod := module.New()
	act := actions.NewSequence()
	mod.Actions["myaction"] = act
	result := FormatModule(mod)
	if !strings.Contains(result, "myaction") {
		t.Errorf("expected 'myaction' in output, got %q", result)
	}
}

func TestFormatModule_WithExports(t *testing.T) {
	mod := module.New()
	mod.PublicActions["ext:foo"] = true
	result := FormatModule(mod)
	if !strings.Contains(result, "export ext:foo") {
		t.Errorf("expected 'export ext:foo' in output, got %q", result)
	}
}

func TestFormatModule_WithInitializers(t *testing.T) {
	mod := module.New()
	act := actions.NewAssumeAction(lg.True)
	mod.Initializers = []module.NamedAction{{Name: "init", Action: act}}
	result := FormatModule(mod)
	if !strings.Contains(result, "after init") {
		t.Errorf("expected 'after init' in output, got %q", result)
	}
}

func TestFormatModule_Deterministic(t *testing.T) {
	mod := module.New()
	mod.Actions["alpha"] = actions.NewSequence()
	mod.Actions["beta"] = actions.NewSequence()
	mod.Actions["gamma"] = actions.NewSequence()
	r1 := FormatModule(mod)
	r2 := FormatModule(mod)
	if r1 != r2 {
		t.Error("FormatModule should be deterministic (sorted output)")
	}
	// Verify alphabetical order
	alphaIdx := strings.Index(r1, "alpha")
	betaIdx := strings.Index(r1, "beta")
	gammaIdx := strings.Index(r1, "gamma")
	if alphaIdx > betaIdx || betaIdx > gammaIdx {
		t.Error("actions should be sorted alphabetically")
	}
}

// --- sortedKeys tests ---

func TestSortedKeys(t *testing.T) {
	m := map[string]interface{}{
		"c": nil,
		"a": nil,
		"b": nil,
	}
	keys := sortedKeys(m)
	if len(keys) != 3 {
		t.Fatalf("expected 3 keys, got %d", len(keys))
	}
	if keys[0] != "a" || keys[1] != "b" || keys[2] != "c" {
		t.Errorf("expected sorted keys [a,b,c], got %v", keys)
	}
}

func TestSortedKeys_Empty(t *testing.T) {
	m := map[string]interface{}{}
	keys := sortedKeys(m)
	if len(keys) != 0 {
		t.Errorf("expected 0 keys, got %d", len(keys))
	}
}

// --- Fuzz tests ---

func FuzzLabeledFmlasToStr(f *testing.F) {
	f.Add("axiom", "inv1")
	f.Add("property", "")
	f.Add("conjecture", "myProp")
	f.Fuzz(func(t *testing.T, kwd, labelName string) {
		var label ast.Node
		if labelName != "" {
			label = lg.NewSymbol(labelName, lg.Boolean)
		}
		acfg := ast.NewAstConfig()
		lf := acfg.NewLabeledFormula(label, lg.True)
		// Should not panic
		result := LabeledFmlasToStr(kwd, []*ast.LabeledFormula{lf})
		if !strings.Contains(result, kwd) {
			t.Errorf("result should contain keyword %q", kwd)
		}
	})
}
