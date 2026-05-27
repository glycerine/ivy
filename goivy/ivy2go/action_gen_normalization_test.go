package ivy2go

import (
	"strings"
	"testing"

	"github.com/glycerine/ivy/goivy"
)

// These tests lock down the Python ivy_to_cpp.py action-gen normalization
// pipeline:
//
//   trim_clauses -> expand_field_references -> extract_input_fields ->
//   extract_defined_parameters -> relevant_definitions -> variant_axioms
//
// They deliberately test both the pure clause helpers and the generated Go
// shape, because the Go backend must not confuse the solver's synthetic inputs
// with the actual action arguments passed to execute().

func actionGenBody(t *testing.T, actions, structName, method string) string {
	t.Helper()
	marker := "func (g *" + structName + ") " + method
	idx := strings.Index(actions, marker)
	if idx < 0 {
		t.Fatalf("%s not emitted:\n%s", marker, actions)
	}
	next := strings.Index(actions[idx+len(marker):], "\nfunc ")
	if next < 0 {
		return actions[idx:]
	}
	return actions[idx : idx+len(marker)+next]
}

func TestActionGenNormalizationExtractsDestructorFieldInputs(t *testing.T) {
	mod := compileIvySource(t, `
type color = {red, green}
type cell
destructor shade(C:cell) : color
individual saved : color
action set(c:cell) = {
	require shade(c) = red;
	saved := shade(c)
}
export set
`)
	out, err := Generate(mod, Config{Target: "test", PackageName: "p"})
	if err != nil {
		t.Fatalf("Generate: %v", err)
	}
	actions := out.Files["actions.go"]
	decl := actionGenBody(t, actions, "actionGen_Set", "generate")
	for _, want := range []string{
		`__in0 := goivy.NewConst("__fml:c__shade", &goivy.LogicEnumeratedSort{Name: "color", Extension: []string{"red", "green"}})`,
		`preFmla := &goivy.LogicAnd{Terms: []goivy.Expr{}}`,
		`g.In_C.Shade = Red`,
	} {
		if !strings.Contains(decl, want) {
			t.Fatalf("field-extracted generate body missing %q:\n%s", want, decl)
		}
	}
	if strings.Contains(decl, `mustApply(goivy.NewConst("shade"`) {
		t.Fatalf("precondition should use extracted field symbol, not destructor Apply:\n%s", decl)
	}
	if strings.Contains(actions, "In_C__shade Color") || strings.Contains(actions, "In___FmlC__shade Color") {
		t.Fatalf("synthetic field input must not become an action argument field:\n%s", actions)
	}
	execBody := actionGenBody(t, actions, "actionGen_Set", "execute")
	if !strings.Contains(execBody, "state.Set(g.In_C)") {
		t.Fatalf("execute should call action with the reconstructed root argument:\n%s", execBody)
	}
	if strings.Contains(execBody, "In_C__shade") || strings.Contains(execBody, "In___FmlC__shade") {
		t.Fatalf("execute must not pass synthetic field symbols to the action:\n%s", execBody)
	}
}

func TestActionGenNormalizationDefinedParameterComputedAfterSolve(t *testing.T) {
	mod := compileIvySource(t, `
type idx = {0..7}
individual stored : idx
action set(x:idx) = {
	require x = 3;
	stored := x
}
export set
`)
	out, err := Generate(mod, Config{Target: "test", PackageName: "p"})
	if err != nil {
		t.Fatalf("Generate: %v", err)
	}
	body := actionGenBody(t, out.Files["actions.go"], "actionGen_Set", "generate")
	if !strings.Contains(body, "g.In_X = 3") {
		t.Fatalf("defined parameter should be computed directly after SAT:\n%s", body)
	}
	if strings.Contains(body, "g.In_X = Idx(pickUintOrChoose") ||
		strings.Contains(body, "g.In_X = pickUintOrChoose") {
		t.Fatalf("defined parameter should not be read back from the solver model:\n%s", body)
	}
}

func TestActionGenNormalizationAddsRelevantDefinitions(t *testing.T) {
	mod := compileIvySource(t, `
type color = {red, green}
individual saved : color
relation is_green(C:color)
definition is_green(C:color) = C = green
action set(c:color) = {
	require is_green(c);
	saved := c
}
export set
`)
	out, err := Generate(mod, Config{Target: "test", PackageName: "p"})
	if err != nil {
		t.Fatalf("Generate: %v", err)
	}
	body := actionGenBody(t, out.Files["actions.go"], "actionGen_Set", "generate")
	for _, want := range []string{
		`goivy.NewConst("is_green", mustNewFunctionSort`,
		`&goivy.LogicIff{T1: mustApply(goivy.NewConst("is_green"`,
		`T2: &goivy.Eq{T1: mustNewVariable("C"`,
		`T2: goivy.NewConst("green"`,
	} {
		if !strings.Contains(body, want) {
			t.Fatalf("relevant definition constraint missing %q:\n%s", want, body)
		}
	}
}

func TestActionGenNormalizationAddsVariantAxioms(t *testing.T) {
	mod := compileIvySource(t, `
type msg
variant req of msg
variant ack of msg
action use(m:msg, r:req) = {
	require m *> r
}
export use
`)
	out, err := Generate(mod, Config{Target: "test", PackageName: "p"})
	if err != nil {
		t.Fatalf("Generate: %v", err)
	}
	body := actionGenBody(t, out.Files["actions.go"], "actionGen_Use", "generate")
	for _, want := range []string{
		`goivy.NewConst("*>", mustNewFunctionSort`,
		`&goivy.ForAll{Variables:`,
		`&goivy.LogicImplies{T1:`,
	} {
		if !strings.Contains(body, want) {
			t.Fatalf("variant axiom constraint missing %q:\n%s", want, body)
		}
	}
}

func TestActionGenClauseHelperExtractDefinedParameters(t *testing.T) {
	p := goivy.NewConst("p", goivy.Boolean)
	q := goivy.NewConst("q", goivy.Boolean)
	r := goivy.NewConst("r", goivy.Boolean)
	eq, _ := goivy.NewEq(p, q)
	other, _ := goivy.NewEq(q, r)
	pre := goivy.NewClauses([]goivy.Expr{eq, other}, nil, nil)
	newPre, defs := extractDefinedParameters(pre, []*goivy.Const{p})
	if len(defs) != 1 {
		t.Fatalf("expected one extracted definition, got %d", len(defs))
	}
	if len(newPre.Fmlas) != 1 || !newPre.Fmlas[0].Equal(other) {
		t.Fatalf("expected only unrelated formula to remain, got %#v", newPre.Fmlas)
	}
}

func TestActionGenClauseHelperKeepsRecursiveDefinedParameter(t *testing.T) {
	p := goivy.NewConst("p", goivy.Boolean)
	q := goivy.NewConst("q", goivy.Boolean)
	eq, _ := goivy.NewEq(p, q)
	other, _ := goivy.NewEq(p, q)
	pre := goivy.NewClauses([]goivy.Expr{eq, other}, nil, nil)
	_, defs := extractDefinedParameters(pre, []*goivy.Const{p})
	if len(defs) != 0 {
		t.Fatalf("recursive input use must not be extracted; got %d defs", len(defs))
	}
}

func TestActionGenClauseHelperExtractInputFields(t *testing.T) {
	mod := compileIvySource(t, `
type color = {red, green}
type cell
destructor shade(C:cell) : color
destructor live(C:cell) : bool
`)
	cell, _ := mod.Sig.Sorts.Get2("cell")
	color, _ := mod.Sig.Sorts.Get2("color")
	shadeEntry, _ := mod.Sig.Symbols.Get2("shade")
	liveEntry, _ := mod.Sig.Symbols.Get2("live")
	c := goivy.NewConst("__fml:c", cell)
	shade := goivy.NewConst("shade", shadeEntry.Sort)
	live := goivy.NewConst("live", liveEntry.Sort)
	red := goivy.NewConst("red", color)
	shadeC := goivy.MustApply(shade, c)
	liveC := goivy.MustApply(live, c)
	eq, _ := goivy.NewEq(shadeC, red)
	pre := goivy.NewClauses([]goivy.Expr{eq}, nil, nil)
	newPre, inputs, fsyms := extractInputFields(pre, []*goivy.Const{c}, mod)
	if len(inputs) != 2 {
		t.Fatalf("expected sibling fields shade/live as solver inputs, got %d: %#v", len(inputs), inputs)
	}
	var sawShade, sawLive bool
	for _, in := range inputs {
		switch in.Name {
		case "__fml:c__shade":
			sawShade = true
			if mapped := fsyms[goivy.Key(in)]; mapped == nil || !mapped.Equal(shadeC) {
				t.Fatalf("shade synthetic input mapped to %#v, want %s", mapped, shadeC)
			}
		case "__fml:c__live":
			sawLive = true
			if mapped := fsyms[goivy.Key(in)]; mapped == nil || !mapped.Equal(liveC) {
				t.Fatalf("live synthetic input mapped to %#v, want %s", mapped, liveC)
			}
		}
	}
	if !sawShade || !sawLive {
		t.Fatalf("missing sibling field inputs shade=%v live=%v; inputs=%#v", sawShade, sawLive, inputs)
	}
	for _, sym := range goivy.UsedSymbolsAst(newPre.ToFormula()).All() {
		if c, ok := sym.(*goivy.Const); ok && c.Name == "shade" {
			t.Fatalf("field reference should be rewritten out of precondition: %s", newPre.ToFormula())
		}
	}
}
