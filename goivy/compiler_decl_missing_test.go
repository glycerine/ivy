package goivy

import (
	"strings"
	"testing"
)

// --- RED tests for missing DomainSetup declaration handlers ---
// These tests document expected behavior from Python's IvyDomainSetup.
// They should all FAIL until the corresponding handlers are implemented.

// TestDomainSetupParameter checks that parameter declarations populate
// mod.Params and mod.ParamDefaults.
// Python: IvyDomainSetup.parameter (ivy_compiler.py:1108)
func TestDomainSetupParameter(t *testing.T) {
	cfg := NewAstConfig()
	c := newTestCompiler()
	d := NewDomainSetup(c)

	// Add sort "nat" to signature
	natSort := &UninterpretedSort{Name: "nat"}
	c.Sig.Sorts.Set("nat", natSort)

	// parameter p : nat
	atom := cfg.NewAtom("p")
	atom.ASort = cfg.NewSymbol("nat", nil)
	decl := cfg.NewParameterDecl(atom)

	err := d.ProcessDecl(decl)
	if err != nil {
		t.Fatalf("ProcessDecl(ParameterDecl): %v", err)
	}

	if len(c.Module.Params) != 1 {
		t.Fatalf("expected 1 param, got %d", len(c.Module.Params))
	}
	if c.Module.Params[0].Name != "p" {
		t.Errorf("expected param name 'p', got %q", c.Module.Params[0].Name)
	}
	if len(c.Module.ParamDefaults) != 1 {
		t.Fatalf("expected 1 param default entry, got %d", len(c.Module.ParamDefaults))
	}
	// No default value: should be nil
	if c.Module.ParamDefaults[0] != nil {
		t.Errorf("expected nil default, got %v", c.Module.ParamDefaults[0])
	}
}

// TestDomainSetupParameterWithDefault checks parameter with a default value.
// Python: when v is a Definition, lhs is param, rhs is default.
func TestDomainSetupParameterWithDefault(t *testing.T) {
	cfg := NewAstConfig()
	c := newTestCompiler()
	d := NewDomainSetup(c)

	natSort := &UninterpretedSort{Name: "nat"}
	c.Sig.Sorts.Set("nat", natSort)
	c.Sig.AddSymbol("zero", natSort)

	// parameter p : nat = zero
	paramAtom := cfg.NewAtom("p")
	paramAtom.ASort = cfg.NewSymbol("nat", nil)
	defaultAtom := cfg.NewAtom("zero")
	def := cfg.NewDefinition(paramAtom, defaultAtom)
	decl := cfg.NewParameterDecl(def)

	err := d.ProcessDecl(decl)
	if err != nil {
		t.Fatalf("ProcessDecl(ParameterDecl with default): %v", err)
	}

	if len(c.Module.Params) != 1 {
		t.Fatalf("expected 1 param, got %d", len(c.Module.Params))
	}
	if c.Module.Params[0].Name != "p" {
		t.Errorf("expected param name 'p', got %q", c.Module.Params[0].Name)
	}
	if len(c.Module.ParamDefaults) != 1 {
		t.Fatalf("expected 1 param default entry, got %d", len(c.Module.ParamDefaults))
	}
	// With default value: should be non-nil (raw AST node)
	if c.Module.ParamDefaults[0] == nil {
		t.Errorf("expected non-nil default for parameter with default value")
	}
}

// TestDomainSetupDestructor checks that destructor declarations populate
// mod.DestructorSorts and mod.SortDestructors.
// Python: IvyDomainSetup.destructor (ivy_compiler.py:1118)
func TestDomainSetupDestructor(t *testing.T) {
	cfg := NewAstConfig()
	c := newTestCompiler()
	d := NewDomainSetup(c)

	pairSort := &UninterpretedSort{Name: "pair"}
	natSort := &UninterpretedSort{Name: "nat"}
	c.Sig.Sorts.Set("pair", pairSort)
	c.Sig.Sorts.Set("nat", natSort)

	// destructor val(X:pair) : nat
	x := cfg.NewVariable("X", "pair")
	atom := cfg.NewAtom("val", x)
	atom.ASort = cfg.NewSymbol("nat", nil)
	decl := cfg.NewDestructorDecl(atom)

	err := d.ProcessDecl(decl)
	if err != nil {
		t.Fatalf("ProcessDecl(DestructorDecl): %v", err)
	}

	// Check DestructorSorts["val"] == pair sort
	ds, ok := c.Module.DestructorSorts["val"]
	if !ok {
		t.Fatal("expected DestructorSorts to have entry 'val'")
	}
	if ds.String() != "pair" {
		t.Errorf("expected destructor sort 'pair', got %q", ds.String())
	}

	// Check SortDestructors["pair"] contains the val symbol
	sd, ok := c.Module.SortDestructors["pair"]
	if !ok {
		t.Fatal("expected SortDestructors to have entry 'pair'")
	}
	found := false
	for _, sym := range sd {
		if sym.Name == "val" {
			found = true
			break
		}
	}
	if !found {
		t.Error("expected SortDestructors['pair'] to contain 'val' symbol")
	}
}

// TestDomainSetupDestructorNoDomain checks that a 0-arity destructor raises an error.
// Python: raises IvyError "A destructor must have at least one parameter"
func TestDomainSetupDestructorNoDomain(t *testing.T) {
	cfg := NewAstConfig()
	c := newTestCompiler()
	d := NewDomainSetup(c)

	natSort := &UninterpretedSort{Name: "nat"}
	c.Sig.Sorts.Set("nat", natSort)

	// destructor val : nat  (0-arity — no parameters)
	atom := cfg.NewAtom("val")
	atom.ASort = cfg.NewSymbol("nat", nil)
	decl := cfg.NewDestructorDecl(atom)

	err := d.ProcessDecl(decl)
	if err == nil {
		t.Fatal("expected error for 0-arity destructor, got nil")
	}
	if !strings.Contains(err.Error(), "at least one parameter") {
		t.Errorf("expected error about 'at least one parameter', got: %v", err)
	}
}

// TestDomainSetupConstructor checks that constructor declarations populate
// mod.ConstructorSorts and mod.SortConstructors.
// Python: IvyDomainSetup.constructor (ivy_compiler.py:1125)
func TestDomainSetupConstructor(t *testing.T) {
	cfg := NewAstConfig()
	c := newTestCompiler()
	d := NewDomainSetup(c)

	natSort := &UninterpretedSort{Name: "nat"}
	pairSort := &UninterpretedSort{Name: "pair"}
	c.Sig.Sorts.Set("nat", natSort)
	c.Sig.Sorts.Set("pair", pairSort)

	// constructor mk_pair(X:nat, Y:nat) : pair
	x := cfg.NewVariable("X", "nat")
	y := cfg.NewVariable("Y", "nat")
	atom := cfg.NewAtom("mk_pair", x, y)
	atom.ASort = cfg.NewSymbol("pair", nil)
	decl := cfg.NewConstructorDecl(atom)

	err := d.ProcessDecl(decl)
	if err != nil {
		t.Fatalf("ProcessDecl(ConstructorDecl): %v", err)
	}

	// Check ConstructorSorts["mk_pair"] == pair sort
	cs, ok := c.Module.ConstructorSorts["mk_pair"]
	if !ok {
		t.Fatal("expected ConstructorSorts to have entry 'mk_pair'")
	}
	if cs.String() != "pair" {
		t.Errorf("expected constructor sort 'pair', got %q", cs.String())
	}

	// Check SortConstructors["pair"] contains mk_pair symbol
	sc, ok := c.Module.SortConstructors["pair"]
	if !ok {
		t.Fatal("expected SortConstructors to have entry 'pair'")
	}
	found := false
	for _, sym := range sc {
		if sym.Name == "mk_pair" {
			found = true
			break
		}
	}
	if !found {
		t.Error("expected SortConstructors['pair'] to contain 'mk_pair' symbol")
	}
}

// TestDomainSetupConcept checks that concept declarations populate
// mod.ConceptSpaces.
// Python: IvyDomainSetup.concept (ivy_compiler.py:1208)
func TestDomainSetupConcept(t *testing.T) {
	cfg := NewAstConfig()
	c := newTestCompiler()
	d := NewDomainSetup(c)

	nodeSort := &UninterpretedSort{Name: "node"}
	c.Sig.Sorts.Set("node", nodeSort)

	// concept rel(X:node) = true
	x := cfg.NewVariable("X", "node")
	rel := cfg.NewAtom("crel", x)
	body := cfg.NewAtom("true")
	lf := cfg.NewLabeledFormula(rel, body)
	decl := cfg.NewConceptDecl(lf)

	err := d.ProcessDecl(decl)
	if err != nil {
		t.Fatalf("ProcessDecl(ConceptDecl): %v", err)
	}

	if len(c.Module.ConceptSpaces) == 0 {
		t.Fatal("expected ConceptSpaces to have at least 1 entry, got 0")
	}
}

// TestDomainSetupRely checks that rely declarations populate mod.Rely.
// Python: IvyDomainSetup.rely (ivy_compiler.py:1203)
func TestDomainSetupRely(t *testing.T) {
	cfg := NewAstConfig()
	c := newTestCompiler()
	d := NewDomainSetup(c)

	// rely true
	formula := cfg.NewAtom("true")
	lf := cfg.NewLabeledFormula(nil, formula)
	decl := cfg.NewRelyDecl(lf)

	err := d.ProcessDecl(decl)
	if err != nil {
		t.Fatalf("ProcessDecl(RelyDecl): %v", err)
	}

	if len(c.Module.Rely) == 0 {
		t.Fatal("expected Rely to have at least 1 entry, got 0")
	}
}

// TestDomainSetupMixord checks that mixord declarations populate mod.MixOrd.
// Python: IvyDomainSetup.mixord (ivy_compiler.py:1206)
func TestDomainSetupMixord(t *testing.T) {
	cfg := NewAstConfig()
	c := newTestCompiler()
	d := NewDomainSetup(c)

	// mixord with an ordering atom
	atom := cfg.NewAtom("some_ordering")
	decl := cfg.NewMixOrdDecl(atom)

	err := d.ProcessDecl(decl)
	if err != nil {
		t.Fatalf("ProcessDecl(MixOrdDecl): %v", err)
	}

	if len(c.Module.MixOrd) == 0 {
		t.Fatal("expected MixOrd to have at least 1 entry, got 0")
	}
}

// TestDomainSetupUpdate checks that update declarations populate mod.Updates.
// Python: IvyDomainSetup.update (ivy_compiler.py:1214)
func TestDomainSetupUpdate(t *testing.T) {
	cfg := NewAstConfig()
	c := newTestCompiler()
	d := NewDomainSetup(c)

	// update wrapping an atom (simplified)
	atom := cfg.NewAtom("some_update")
	decl := cfg.NewUpdateDecl(atom)

	err := d.ProcessDecl(decl)
	if err != nil {
		t.Fatalf("ProcessDecl(UpdateDecl): %v", err)
	}

	if len(c.Module.Updates) == 0 {
		t.Fatal("expected Updates to have at least 1 entry, got 0")
	}
}

// TestDomainSetupScenario checks that scenario declarations create relation
// symbols for places and populate mod.Relations and mod.AllRelations.
// Python: IvyDomainSetup.scenario (ivy_compiler.py:1333)
func TestDomainSetupScenario(t *testing.T) {
	cfg := NewAstConfig()
	c := newTestCompiler()
	d := NewDomainSetup(c)

	// scenario with places: state_a, state_b
	placeA := cfg.NewAtom("state_a")
	placeB := cfg.NewAtom("state_b")
	places := cfg.NewPlaceList([]Node{placeA, placeB})
	scenDef := cfg.NewScenarioDef([]Node{places})
	decl := cfg.NewScenarioDecl(scenDef)

	err := d.ProcessDecl(decl)
	if err != nil {
		t.Fatalf("ProcessDecl(ScenarioDecl): %v", err)
	}

	// Each place should have a relation symbol in sig
	if _, ok := c.Sig.Symbols.Get2("state_a"); !ok {
		t.Error("expected 'state_a' in sig symbols")
	}
	if _, ok := c.Sig.Symbols.Get2("state_b"); !ok {
		t.Error("expected 'state_b' in sig symbols")
	}

	// Each place should have a relation entry
	if _, ok := c.Module.Relations.Get2("state_a"); !ok {
		t.Error("expected 'state_a' in module Relations")
	}
	if _, ok := c.Module.Relations.Get2("state_b"); !ok {
		t.Error("expected 'state_b' in module Relations")
	}

	// AllRelations should have 2 entries
	if len(c.Module.AllRelations) < 2 {
		t.Errorf("expected at least 2 AllRelations entries, got %d", len(c.Module.AllRelations))
	}
}

// TestDomainSetupImplementtype checks that implement type declarations
// populate mod.Interps.
// Python: IvyDomainSetup.implementtype (ivy_compiler.py:1254)
func TestDomainSetupImplementtype(t *testing.T) {
	cfg := NewAstConfig()
	c := newTestCompiler()
	d := NewDomainSetup(c)

	fooSort := &UninterpretedSort{Name: "foo"}
	barSort := &UninterpretedSort{Name: "bar"}
	c.Sig.Sorts.Set("foo", fooSort)
	c.Sig.Sorts.Set("bar", barSort)

	// implement type foo = bar
	lhs := cfg.NewSymbol("foo", nil)
	rhs := cfg.NewSymbol("bar", nil)
	def := cfg.NewDefinition(lhs, rhs)
	lf := cfg.NewLabeledFormula(nil, def)
	decl := cfg.NewImplementTypeDecl(lf)

	err := d.ProcessDecl(decl)
	if err != nil {
		t.Fatalf("ProcessDecl(ImplementTypeDecl): %v", err)
	}

	interps, ok := c.Module.Interps["foo"]
	if !ok || len(interps) == 0 {
		t.Fatal("expected Interps['foo'] to have at least 1 entry")
	}
}

// TestDomainSetupImplementtypeAlreadyInterpreted checks that implementing an
// already-interpreted type raises an error.
// Python: raises IvyError "{} is already interpreted"
func TestDomainSetupImplementtypeAlreadyInterpreted(t *testing.T) {
	cfg := NewAstConfig()
	c := newTestCompiler()
	d := NewDomainSetup(c)

	fooSort := &UninterpretedSort{Name: "foo"}
	barSort := &UninterpretedSort{Name: "bar"}
	c.Sig.Sorts.Set("foo", fooSort)
	c.Sig.Sorts.Set("bar", barSort)

	// Mark foo as already having a native type interpretation
	c.Module.NativeTypes["foo"] = cfg.NewNativeType(cfg.NewAtom("already_interp"))

	// implement type foo = bar  (should fail — already interpreted)
	lhs := cfg.NewSymbol("foo", nil)
	rhs := cfg.NewSymbol("bar", nil)
	def := cfg.NewDefinition(lhs, rhs)
	lf := cfg.NewLabeledFormula(nil, def)
	decl := cfg.NewImplementTypeDecl(lf)

	err := d.ProcessDecl(decl)
	if err == nil {
		t.Fatal("expected error for already-interpreted type, got nil")
	}
	if !strings.Contains(err.Error(), "already interpreted") {
		t.Errorf("expected error about 'already interpreted', got: %v", err)
	}
}

// TestARGSetupScenario checks that scenario declarations in ARGSetup (pass 3)
// create init actions, register mixins, and create mixer actions.
// Python: IvyARGSetup.scenario (ivy_compiler.py:1462-1530)
func TestARGSetupScenario(t *testing.T) {
	cfg := NewAstConfig()
	c := newTestCompiler()

	// First run DomainSetup to create place symbols (pass 1)
	relSort := RelationSort([]Sort{})
	for _, name := range []string{"s0", "s1"} {
		_, err := c.AddSymbol(name, relSort, c.Sig)
		if err != nil {
			t.Fatalf("AddSymbol(%s): %v", name, err)
		}
		c.Module.Relations.Set(name, relSort)
	}

	// Also need "init" and "a" as known actions
	c.Module.Actions.Set("init", nil)
	c.Module.Actions.Set("a", nil)

	// Build a ScenarioDef like scen1.ivy:
	// scenario { -> s0; s0 -> s1 : before a { q := true }  s1 -> s0 : before a { q := false } }
	initPlaces := cfg.NewPlaceList([]Node{cfg.NewAtom("s0")})

	// Transition 0: s0 -> s1 : before a { ... }
	actionAtom0 := cfg.NewAtom("a")
	body0 := cfg.NewAnd() // placeholder body
	adef0 := cfg.NewActionDef(actionAtom0, body0, nil, nil)
	mixer0 := cfg.NewAtom("a[before]")
	mixin0 := cfg.NewScenarioBeforeMixin(mixer0, adef0)
	tr0 := cfg.NewScenarioTransition(
		cfg.NewPlaceList([]Node{cfg.NewAtom("s0")}),
		cfg.NewPlaceList([]Node{cfg.NewAtom("s1")}),
		mixin0,
	)

	// Transition 1: s1 -> s0 : before a { ... }
	actionAtom1 := cfg.NewAtom("a")
	body1 := cfg.NewAnd()
	adef1 := cfg.NewActionDef(actionAtom1, body1, nil, nil)
	mixer1 := cfg.NewAtom("a[before]")
	mixin1 := cfg.NewScenarioBeforeMixin(mixer1, adef1)
	tr1 := cfg.NewScenarioTransition(
		cfg.NewPlaceList([]Node{cfg.NewAtom("s1")}),
		cfg.NewPlaceList([]Node{cfg.NewAtom("s0")}),
		mixin1,
	)

	scenDef := cfg.NewScenarioDef([]Node{initPlaces, tr0, tr1})
	scenDecl := cfg.NewScenarioDecl(scenDef)

	// Run ARGSetup
	as := NewARGSetup(c)
	err := as.ProcessDecls([]Node{scenDecl})
	if err != nil {
		t.Fatalf("ARGSetup.ProcessDecls: %v", err)
	}

	// Check init actions exist
	if _, ok := c.Module.Actions.Get2("s0[init]"); !ok {
		t.Error("expected 's0[init]' in mod.Actions")
	}
	if _, ok := c.Module.Actions.Get2("s1[init]"); !ok {
		t.Error("expected 's1[init]' in mod.Actions")
	}

	// Check mixins for "init" have 2 entries (s0[init] and s1[init])
	initMixins := c.Module.Mixins.Get("init")
	if len(initMixins) < 2 {
		t.Errorf("expected at least 2 mixins for 'init', got %d", len(initMixins))
	}

	// Check that mixer action for "a[before]" exists
	if _, ok := c.Module.Actions.Get2("a[before]"); !ok {
		t.Error("expected 'a[before]' in mod.Actions")
	}

	// Check mixins for "a" has a MixinBeforeDef
	aMixins := c.Module.Mixins.Get("a")
	if len(aMixins) < 1 {
		t.Errorf("expected at least 1 mixin for 'a', got %d", len(aMixins))
	}
}

// Ensure imports are used.
var _ = RelationSort
