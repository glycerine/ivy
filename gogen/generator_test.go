package gogen

import (
	"strings"
	"testing"

	"github.com/glycerine/goivy/actions"
	il "github.com/glycerine/goivy/ivylogic"
	lg "github.com/glycerine/goivy/logic"
	"github.com/glycerine/goivy/module"
)

// newTestModule builds a simple module for testing.
func newTestModule() *module.Module {
	mod := module.New()
	mod.Sig = il.NewSig()

	// Add an enumerated sort: Color = {red, green, blue}
	colorSort := &lg.EnumeratedSort{
		Name:      "color",
		Extension: []string{"red", "green", "blue"},
	}
	mod.Sig.Sorts["color"] = colorSort
	mod.SortOrder = append(mod.SortOrder, "color")

	// Add a boolean relation: link(int, int) -> bool
	linkSort, _ := lg.NewFunctionSort(lg.Boolean, lg.Boolean, lg.Boolean)
	mod.Relations["link"] = linkSort

	// Add a function: data -> color (using bool as domain placeholder)
	dataSort, _ := lg.NewFunctionSort(lg.Boolean, colorSort)
	mod.Functions["data"] = dataSort

	return mod
}

func TestGenerator_PackageDecl(t *testing.T) {
	mod := newTestModule()
	gen := NewGenerator(mod, "main")
	out, err := gen.Generate()
	if err != nil {
		t.Fatal(err)
	}
	if !strings.HasPrefix(out, "package main\n") {
		t.Errorf("expected package main at start, got prefix: %q", out[:min(50, len(out))])
	}
}

func TestGenerator_CustomPackage(t *testing.T) {
	mod := newTestModule()
	gen := NewGenerator(mod, "mylib")
	out, err := gen.Generate()
	if err != nil {
		t.Fatal(err)
	}
	if !strings.HasPrefix(out, "package mylib\n") {
		t.Errorf("expected package mylib, got prefix: %q", out[:min(50, len(out))])
	}
	// Non-main package should not have a main() function.
	if strings.Contains(out, "func main()") {
		t.Errorf("non-main package should not have main()")
	}
}

func TestGenerator_Imports(t *testing.T) {
	mod := newTestModule()
	gen := NewGenerator(mod, "main")
	out, err := gen.Generate()
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(out, `"fmt"`) {
		t.Errorf("expected fmt import")
	}
	if !strings.Contains(out, `"math/rand"`) {
		t.Errorf("expected math/rand import")
	}
}

func TestGenerator_EnumDecl(t *testing.T) {
	mod := newTestModule()
	gen := NewGenerator(mod, "main")
	out, err := gen.Generate()
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(out, "type Color int") {
		t.Errorf("expected Color type declaration, got: %s", out)
	}
	if !strings.Contains(out, "Red Color = iota") {
		t.Errorf("expected iota enum, got: %s", out)
	}
	if !strings.Contains(out, "Green") {
		t.Errorf("expected Green enum value")
	}
	if !strings.Contains(out, "Blue") {
		t.Errorf("expected Blue enum value")
	}
}

func TestGenerator_QuantifierHelpers(t *testing.T) {
	mod := newTestModule()
	gen := NewGenerator(mod, "main")
	out, err := gen.Generate()
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(out, "func forAll_Color(f func(Color) bool) bool") {
		t.Errorf("expected forAll_Color helper")
	}
	if !strings.Contains(out, "func exists_Color(f func(Color) bool) bool") {
		t.Errorf("expected exists_Color helper")
	}
}

func TestGenerator_StateStruct(t *testing.T) {
	mod := newTestModule()
	gen := NewGenerator(mod, "main")
	out, err := gen.Generate()
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(out, "type State struct {") {
		t.Errorf("expected State struct")
	}
}

func TestGenerator_NewState(t *testing.T) {
	mod := newTestModule()
	gen := NewGenerator(mod, "main")
	out, err := gen.Generate()
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(out, "func NewState() *State {") {
		t.Errorf("expected NewState constructor")
	}
	if !strings.Contains(out, "make(") {
		t.Errorf("expected make() calls for maps")
	}
}

func TestGenerator_Init(t *testing.T) {
	mod := newTestModule()
	gen := NewGenerator(mod, "main")
	out, err := gen.Generate()
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(out, "func (s *State) Init() {") {
		t.Errorf("expected Init method")
	}
}

func TestGenerator_MainFunc(t *testing.T) {
	mod := newTestModule()
	gen := NewGenerator(mod, "main")
	out, err := gen.Generate()
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(out, "func main() {") {
		t.Errorf("expected main function")
	}
	if !strings.Contains(out, "NewState()") {
		t.Errorf("expected NewState call in main")
	}
	if !strings.Contains(out, "s.Init()") {
		t.Errorf("expected Init call in main")
	}
}

func TestGenerator_WithAction(t *testing.T) {
	mod := newTestModule()

	// Add a simple action.
	assignAct := actions.NewAssignAction(
		testConst("x", lg.Boolean),
		testConst("y", lg.Boolean),
	)
	mod.Actions.Set("send", assignAct)

	gen := NewGenerator(mod, "main")
	out, err := gen.Generate()
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(out, "func (s *State) Send()") {
		t.Errorf("expected Send method, got: %s", out)
	}
	if !strings.Contains(out, "x = y") {
		t.Errorf("expected assignment in method body")
	}
}

func TestGenerator_WithInitializer(t *testing.T) {
	mod := newTestModule()

	initAct := actions.NewAssignAction(
		testConst("x", lg.Boolean),
		testConst("false", lg.Boolean),
	)
	mod.Initializers = append(mod.Initializers, module.NamedAction{
		Name:   "init_x",
		Action: initAct,
	})

	gen := NewGenerator(mod, "main")
	out, err := gen.Generate()
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(out, "// initializer: init_x") {
		t.Errorf("expected initializer comment")
	}
}

func TestGenerator_EmptyModule(t *testing.T) {
	mod := module.New()
	gen := NewGenerator(mod, "main")
	out, err := gen.Generate()
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(out, "package main") {
		t.Errorf("expected package declaration")
	}
	if !strings.Contains(out, "type State struct {") {
		t.Errorf("expected empty State struct")
	}
}

func TestGenerator_NilModule(t *testing.T) {
	gen := NewGenerator(nil, "main")
	out, err := gen.Generate()
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(out, "package main") {
		t.Errorf("expected package main")
	}
}

func TestGenerator_MultipleActions(t *testing.T) {
	mod := newTestModule()
	a1 := actions.NewAssignAction(testConst("x", lg.Boolean), testConst("y", lg.Boolean))
	a2 := actions.NewAssignAction(testConst("a", lg.Boolean), testConst("b", lg.Boolean))
	mod.Actions.Set("alpha", a1)
	mod.Actions.Set("beta", a2)

	gen := NewGenerator(mod, "main")
	out, err := gen.Generate()
	if err != nil {
		t.Fatal(err)
	}
	// Actions should be sorted alphabetically.
	alphaIdx := strings.Index(out, "func (s *State) Alpha()")
	betaIdx := strings.Index(out, "func (s *State) Beta()")
	if alphaIdx == -1 || betaIdx == -1 {
		t.Fatalf("expected both Alpha and Beta methods")
	}
	if alphaIdx > betaIdx {
		t.Errorf("expected Alpha before Beta (sorted)")
	}
}

func TestGenerator_ActionWithParams(t *testing.T) {
	mod := newTestModule()
	act := actions.NewAssignAction(testConst("x", lg.Boolean), testConst("y", lg.Boolean))
	act.FormalParams = []*lg.Symbol{
		lg.NewSymbol("src", lg.Boolean),
		lg.NewSymbol("dst", lg.Boolean),
	}
	mod.Actions.Set("send", act)

	gen := NewGenerator(mod, "main")
	out, err := gen.Generate()
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(out, "src bool") {
		t.Errorf("expected src parameter, got: %s", out)
	}
	if !strings.Contains(out, "dst bool") {
		t.Errorf("expected dst parameter, got: %s", out)
	}
}

func TestGenerator_SequenceInAction(t *testing.T) {
	mod := newTestModule()
	a1 := actions.NewAssignAction(testConst("x", lg.Boolean), testConst("y", lg.Boolean))
	a2 := actions.NewAssertAction(testConst("valid", lg.Boolean))
	seq := actions.NewSequence(a1, a2)
	mod.Actions.Set("do_stuff", seq)

	gen := NewGenerator(mod, "main")
	out, err := gen.Generate()
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(out, "x = y") {
		t.Errorf("expected assignment in sequence")
	}
	if !strings.Contains(out, "assertion failed") {
		t.Errorf("expected assertion in sequence")
	}
}

func TestGenerator_DefaultPackage(t *testing.T) {
	mod := newTestModule()
	gen := NewGenerator(mod, "")
	out, err := gen.Generate()
	if err != nil {
		t.Fatal(err)
	}
	if !strings.HasPrefix(out, "package main\n") {
		t.Errorf("empty package name should default to main")
	}
}

// --- formatFormalParams tests ---

func TestFormatFormalParams_Empty(t *testing.T) {
	got := formatFormalParams(nil)
	if got != "" {
		t.Errorf("expected empty, got: %s", got)
	}
}

func TestFormatFormalParams_One(t *testing.T) {
	params := []*lg.Symbol{lg.NewSymbol("x", lg.Boolean)}
	got := formatFormalParams(params)
	if got != "x bool" {
		t.Errorf("expected 'x bool', got: %s", got)
	}
}

func TestFormatFormalParams_Multiple(t *testing.T) {
	params := []*lg.Symbol{
		lg.NewSymbol("x", lg.Boolean),
		lg.NewSymbol("y", lg.Boolean),
	}
	got := formatFormalParams(params)
	if got != "x bool, y bool" {
		t.Errorf("expected 'x bool, y bool', got: %s", got)
	}
}

func TestFormatFormalReturns_Empty(t *testing.T) {
	got := formatFormalReturns(nil)
	if got != "" {
		t.Errorf("expected empty, got: %s", got)
	}
}

func TestFormatFormalReturns_Single(t *testing.T) {
	params := []*lg.Symbol{lg.NewSymbol("r", lg.Boolean)}
	got := formatFormalReturns(params)
	if got != "bool" {
		t.Errorf("expected 'bool', got: %s", got)
	}
}

// --- StateFieldType ---

func TestStateFieldType_BoolRelation(t *testing.T) {
	got := StateFieldType("active", lg.Boolean)
	if got != "bool" {
		t.Errorf("expected bool, got: %s", got)
	}
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}
