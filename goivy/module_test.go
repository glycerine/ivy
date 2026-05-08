package goivy

import (
	"testing"
)

// dummyAction is a minimal Action implementation for tests.
type dummyAction struct {
	ActionBase
	Tag string
}

func (d *dummyAction) String() string                 { return d.Tag }
func (d *dummyAction) ActionClone(args []Expr) Action { return d }
func (d *dummyAction) ActionArgs() []Expr             { return nil }
func (d *dummyAction) IterCalls() []string            { return nil }
func (d *dummyAction) IterSubactions() []Action       { return []Action{d} }
func (d *dummyAction) Name() string                   { return "dummy" }
func (d *dummyAction) Decompose() [][]Action          { return nil }
func (d *dummyAction) Args() []Node                   { return nil }
func (d *dummyAction) Clone(args []Node) Node         { return d }
func (d *dummyAction) Children() []Expr               { return nil }
func (d *dummyAction) NodeSort() Sort                 { return ActionS }
func (d *dummyAction) Equal(other Expr) bool          { return false }
func (d *dummyAction) GetAstConfig() *AstConfig       { return nil }
func (d *dummyAction) Sexp() NodeKey                  { return "(dummyAction)" }
func (d *dummyAction) Canon() Canonical               { return Canonical(d.Sexp()) }

func TestModuleNew(t *testing.T) {
	m := NewModule()
	if m == nil {
		t.Fatal("New returned nil")
	}
	if m.Sig == nil {
		t.Error("Module.Sig should not be nil")
	}
	if m.Actions == nil {
		t.Error("Actions map should be initialized")
	}
}

func TestModuleClear(t *testing.T) {
	m := NewModule()
	m.Actions.Set("test", &dummyAction{Tag: "dummy"})
	m.LabeledAxioms = append(m.LabeledAxioms, m.Cfg.AstCfg.NewLabeledFormula(nil, nil))
	m.Clear()
	if m.Actions.Len() != 0 {
		t.Error("Clear should empty Actions")
	}
	if len(m.LabeledAxioms) != 0 {
		t.Error("Clear should empty LabeledAxioms")
	}
}

func TestModuleCopy(t *testing.T) {
	m := NewModule()
	sort := &UninterpretedSort{Name: "node"}
	m.Sig.AddSort(sort)
	m.Actions.Set("act1", &dummyAction{Tag: "dummy"})
	m.LabeledAxioms = append(m.LabeledAxioms, m.Cfg.AstCfg.NewLabeledFormula(nil, &LogicAnd{}))
	m.GhostSorts["ghost"] = true

	c := m.Copy()

	// Verify copy has same data
	if _, ok := c.Sig.Sorts.Get2("node"); !ok {
		t.Error("copy should have node sort")
	}
	if _, ok := c.Actions.Get2("act1"); !ok {
		t.Error("copy should have act1 action")
	}
	if len(c.LabeledAxioms) != 1 {
		t.Error("copy should have 1 axiom")
	}
	if !c.GhostSorts["ghost"] {
		t.Error("copy should have ghost sort")
	}

	// Modify copy, verify original unchanged
	c.Actions.Set("act2", &dummyAction{Tag: "new"})
	if _, ok := m.Actions.Get2("act2"); ok {
		t.Error("modifying copy should not affect original")
	}
	c.Sig.AddSort(&UninterpretedSort{Name: "extra"})
	if _, ok := m.Sig.Sorts.Get2("extra"); ok {
		t.Error("modifying copy's sig should not affect original")
	}
}

func TestModuleAddToHierarchy(t *testing.T) {
	m := NewModule()
	m.AddToHierarchy("protocol")
	thisMap, ok := m.Hierarchy.Get2("this")
	if !ok || !thisMap.Get("protocol") {
		t.Error("protocol should be under 'this'")
	}
}

func TestModuleAddToHierarchyDotted(t *testing.T) {
	m := NewModule()
	m.AddToHierarchy("net.protocol")
	thisMap2, ok2 := m.Hierarchy.Get2("this")
	if !ok2 || !thisMap2.Get("net") {
		t.Error("net should be under 'this'")
	}
	netMap, ok3 := m.Hierarchy.Get2("net")
	if !ok3 || !netMap.Get("protocol") {
		t.Error("protocol should be under 'net'")
	}
}

func TestModuleAddObject(t *testing.T) {
	m := NewModule()
	m.AddObject("myobj")
	if _, ok := m.Hierarchy.Get2("myobj"); !ok {
		t.Error("AddObject should create hierarchy entry")
	}
}

func TestModuleFindAction(t *testing.T) {
	m := NewModule()
	m.Actions.Set("send", &dummyAction{Tag: "action_impl"})
	a, ok := m.FindAction("send")
	if !ok {
		t.Fatal("expected to find send")
	}
	if a.(*dummyAction).Tag != "action_impl" {
		t.Error("unexpected action value")
	}

	_, ok = m.FindAction("nonexistent")
	if ok {
		t.Error("expected false for nonexistent action")
	}
}

func TestModuleIsVariant(t *testing.T) {
	m := NewModule()
	lsort := &UninterpretedSort{Name: "msg"}
	rsort := &UninterpretedSort{Name: "req"}
	m.Variants["msg"] = []Sort{rsort}

	if !m.IsVariant(lsort, rsort) {
		t.Error("req should be variant of msg")
	}
	if m.IsVariant(rsort, lsort) {
		t.Error("msg should not be variant of req")
	}
}

func TestModuleVariantIndex(t *testing.T) {
	m := NewModule()
	s1 := &UninterpretedSort{Name: "a"}
	s2 := &UninterpretedSort{Name: "b"}
	lsort := &UninterpretedSort{Name: "msg"}
	m.Variants["msg"] = []Sort{s1, s2}

	if m.VariantIndex(lsort, s1) != 0 {
		t.Error("expected index 0")
	}
	if m.VariantIndex(lsort, s2) != 1 {
		t.Error("expected index 1")
	}
	if m.VariantIndex(lsort, lsort) != -1 {
		t.Error("expected -1 for non-variant")
	}
}

func TestModuleSortCard(t *testing.T) {
	m := NewModule()
	// Enumerated sort should return card
	es := &LogicEnumeratedSort{Name: "color", Extension: []string{"red", "green", "blue"}}
	if SortCardDefault(es) != 3 {
		t.Errorf("expected 3, got %d", SortCardDefault(es))
	}

	// Function sort
	fs, _ := NewFunctionSort(&UninterpretedSort{Name: "t"}, Boolean)
	if m.SortCard(fs) != -1 {
		t.Error("function sort should have unknown card")
	}

	// Uninterpreted sort
	us := &UninterpretedSort{Name: "t"}
	if m.SortCard(us) != -1 {
		t.Error("uninterpreted sort should have unknown card")
	}
}

func TestModuleSortDependencies(t *testing.T) {
	m := NewModule()
	tSort := &UninterpretedSort{Name: "t"}
	uSort := &UninterpretedSort{Name: "u"}
	dSort, _ := NewFunctionSort(tSort, uSort)
	destr := NewConst("d", dSort)
	m.SortDestructors["t"] = []*Const{destr}

	deps := m.SortDependencies("t", false)
	if len(deps) != 1 || deps[0] != "u" {
		t.Errorf("expected [u], got %v", deps)
	}
}

func TestModuleSortDependenciesVariants(t *testing.T) {
	m := NewModule()
	v1 := &UninterpretedSort{Name: "v1"}
	v2 := &UninterpretedSort{Name: "v2"}
	m.Variants["msg"] = []Sort{v1, v2}

	deps := m.SortDependencies("msg", true)
	if len(deps) != 2 {
		t.Errorf("expected 2 variant deps, got %d", len(deps))
	}
}

func TestModuleModuleString(t *testing.T) {
	m := NewModule()
	s := m.String()
	if len(s) == 0 {
		t.Error("Module.String() should not be empty")
	}
}

func TestModuleNewWithSig(t *testing.T) {
	sig := NewSig()
	sig.AddSort(&UninterpretedSort{Name: "custom"})
	m := NewWithSig(sig)
	if _, ok := m.Sig.Sorts.Get2("custom"); !ok {
		t.Error("module should use provided sig")
	}
}

func TestModuleLabeledFormula(t *testing.T) {
	acfg := NewAstConfig()
	lf := acfg.NewLabeledFormula(nil, &LogicAnd{})
	lf.Temporal = BoolPtr(true)
	lf.SetLineno(Location{Line: 42})
	if !lf.IsTemporal() {
		t.Error("should be temporal")
	}
	if lf.Lineno() != 42 {
		t.Error("wrong lineno")
	}
}

func TestModuleIsolateInfo(t *testing.T) {
	info := &IsolateInfo{}
	info.Implementations = append(info.Implementations, MixinTriple{
		Mixer: "a", Mixee: "b",
	})
	if len(info.Implementations) != 1 {
		t.Error("expected 1 implementation")
	}
}

// --- Fuzz tests ---

func FuzzModuleCopy(f *testing.F) {
	f.Add("sort1", "sym1")
	f.Add("", "")
	f.Add("a.b.c", "x.y")

	f.Fuzz(func(t *testing.T, sortName, symName string) {
		m := NewModule()
		if sortName != "" {
			m.Sig.AddSort(&UninterpretedSort{Name: sortName})
		}
		if symName != "" {
			m.Sig.AddSymbol(symName, TopS)
		}
		m.GhostSorts["g"] = true
		m.Actions.Set("a", &dummyAction{Tag: "v"})

		c := m.Copy()
		// Modify copy
		c.GhostSorts["g2"] = true
		c.Actions.Set("b", &dummyAction{Tag: "w"})

		// Original should be unchanged
		if m.GhostSorts["g2"] {
			t.Error("original modified via copy")
		}
		if _, ok := m.Actions.Get2("b"); ok {
			t.Error("original modified via copy")
		}
	})
}
