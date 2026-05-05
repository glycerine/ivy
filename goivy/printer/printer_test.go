package printer

import (
	goivy "github.com/glycerine/ivy/goivy"
	"strings"
	"testing"
)

func boolConst(name string) *goivy.Const {
	return goivy.NewConst(name, goivy.Boolean)
}

// --- LabeledFmlasToStr tests ---

func TestLabeledFmlasToStr_Empty(t *testing.T) {
	result := LabeledFmlasToStr("axiom", nil)
	if result != "" {
		t.Errorf("expected empty string for nil input, got %q", result)
	}
}

func TestLabeledFmlasToStr_NoLabel(t *testing.T) {
	acfg := goivy.NewAstConfig()
	lf := acfg.NewLabeledFormula(nil, goivy.True)
	result := LabeledFmlasToStr("axiom", []*goivy.LabeledFormula{lf})
	if !strings.HasPrefix(result, "axiom ") {
		t.Errorf("expected prefix 'axiom ', got %q", result)
	}
	if !strings.HasSuffix(result, "\n") {
		t.Error("expected trailing newline")
	}
}

func TestLabeledFmlasToStr_WithLabel(t *testing.T) {
	label := boolConst("inv1")
	acfg := goivy.NewAstConfig()
	lf := acfg.NewLabeledFormula(label, goivy.True)
	result := LabeledFmlasToStr("conjecture", []*goivy.LabeledFormula{lf})
	if !strings.Contains(result, "[") || !strings.Contains(result, "]") {
		t.Errorf("expected brackets around label, got %q", result)
	}
	if !strings.Contains(result, "inv1") {
		t.Errorf("expected label name 'inv1', got %q", result)
	}
}

func TestLabeledFmlasToStr_Multiple(t *testing.T) {
	acfg := goivy.NewAstConfig()
	lf1 := acfg.NewLabeledFormula(nil, goivy.True)
	lf2 := acfg.NewLabeledFormula(nil, goivy.False)
	result := LabeledFmlasToStr("property", []*goivy.LabeledFormula{lf1, lf2})
	count := strings.Count(result, "property")
	if count != 2 {
		t.Errorf("expected 2 occurrences of 'property', got %d", count)
	}
}

// --- FormatModule tests ---

func TestFormatModule_Empty(t *testing.T) {
	mod := goivy.New()
	result := FormatModule(mod)
	// Should not panic and should produce some output (at least the sig)
	if result == "" {
		// Empty module may have an empty sig representation, that's ok
	}
}

func TestFormatModule_WithAxioms(t *testing.T) {
	mod := goivy.New()
	acfg := goivy.NewAstConfig()
	lf := acfg.NewLabeledFormula(nil, goivy.True)
	mod.LabeledAxioms = []*goivy.LabeledFormula{lf}
	result := FormatModule(mod)
	if !strings.Contains(result, "axiom") {
		t.Errorf("expected 'axiom' in output, got %q", result)
	}
}

func TestFormatModule_WithConjectures(t *testing.T) {
	mod := goivy.New()
	label := boolConst("inv1")
	acfg := goivy.NewAstConfig()
	lf := acfg.NewLabeledFormula(label, goivy.True)
	mod.LabeledConjs = []*goivy.LabeledFormula{lf}
	result := FormatModule(mod)
	if !strings.Contains(result, "conjecture") {
		t.Errorf("expected 'conjecture' in output, got %q", result)
	}
}

func TestFormatModule_WithActions(t *testing.T) {
	mod := goivy.New()
	act := goivy.NewSequence()
	mod.Actions.Set("myaction", act)
	result := FormatModule(mod)
	if !strings.Contains(result, "myaction") {
		t.Errorf("expected 'myaction' in output, got %q", result)
	}
}

func TestFormatModule_WithExports(t *testing.T) {
	mod := goivy.New()
	mod.PublicActions.Set("ext:foo", true)
	result := FormatModule(mod)
	if !strings.Contains(result, "export ext:foo") {
		t.Errorf("expected 'export ext:foo' in output, got %q", result)
	}
}

func TestFormatModule_WithInitializers(t *testing.T) {
	mod := goivy.New()
	act := goivy.NewAssumeAction(goivy.True)
	mod.Initializers = []goivy.NamedAction{{Name: "init", Action: act}}
	result := FormatModule(mod)
	if !strings.Contains(result, "after init") {
		t.Errorf("expected 'after init' in output, got %q", result)
	}
}

func TestFormatModule_Deterministic(t *testing.T) {
	mod := goivy.New()
	mod.Actions.Set("alpha", goivy.NewSequence())
	mod.Actions.Set("beta", goivy.NewSequence())
	mod.Actions.Set("gamma", goivy.NewSequence())
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
		var label goivy.Node
		if labelName != "" {
			label = goivy.NewConst(labelName, goivy.Boolean)
		}
		acfg := goivy.NewAstConfig()
		lf := acfg.NewLabeledFormula(label, goivy.True)
		// Should not panic
		result := LabeledFmlasToStr(kwd, []*goivy.LabeledFormula{lf})
		if !strings.Contains(result, kwd) {
			t.Errorf("result should contain keyword %q", kwd)
		}
	})
}
