package module

import (
	"testing"

	"github.com/glycerine/goivy/ast"
	il "github.com/glycerine/goivy/ivylogic"
	iu "github.com/glycerine/goivy/ivyutils"
	lg "github.com/glycerine/goivy/logic"
)

// dummyAction is a minimal Action implementation for tests.
type dummyAction struct {
	ActionBase
	Tag string
}

func (d *dummyAction) String() string                       { return d.Tag }
func (d *dummyAction) ActionClone(args []lg.Expr) Action    { return d }
func (d *dummyAction) ActionArgs() []lg.Expr                { return nil }
func (d *dummyAction) IterCalls() []string                  { return nil }
func (d *dummyAction) IterSubactions() []Action             { return []Action{d} }
func (d *dummyAction) Name() string                         { return "dummy" }
func (d *dummyAction) Decompose() [][]Action                { return nil }
func (d *dummyAction) Args() []ast.Node                     { return nil }
func (d *dummyAction) Clone(args []ast.Node) ast.Node       { return d }
func (d *dummyAction) Children() []lg.Expr                  { return nil }
func (d *dummyAction) NodeSort() lg.Sort                    { return lg.ActionS }
func (d *dummyAction) Equal(other lg.Expr) bool             { return false }
func (d *dummyAction) GetAstConfig() *ast.AstConfig         { return nil }
func (d *dummyAction) Sexp() lg.NodeKey                     { return "(dummyAction)" }
func (d *dummyAction) Canon() iu.Canonical                  { return iu.Canonical(d.Sexp()) }

func TestNew(t *testing.T) {
	m := New()
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

func TestClear(t *testing.T) {
	m := New()
	m.Actions["test"] = &dummyAction{Tag: "dummy"}
	m.LabeledAxioms = append(m.LabeledAxioms, m.Cfg.AstCfg.NewLabeledFormula(nil, nil))
	m.Clear()
	if len(m.Actions) != 0 {
		t.Error("Clear should empty Actions")
	}
	if len(m.LabeledAxioms) != 0 {
		t.Error("Clear should empty LabeledAxioms")
	}
}

func TestCopy(t *testing.T) {
	m := New()
	sort := &lg.UninterpretedSort{Name: "node"}
	m.Sig.AddSort(sort)
	m.Actions["act1"] = &dummyAction{Tag: "dummy"}
	m.LabeledAxioms = append(m.LabeledAxioms, m.Cfg.AstCfg.NewLabeledFormula(nil, &lg.And{}))
	m.GhostSorts["ghost"] = true

	c := m.Copy()

	// Verify copy has same data
	if _, ok := c.Sig.Sorts["node"]; !ok {
		t.Error("copy should have node sort")
	}
	if _, ok := c.Actions["act1"]; !ok {
		t.Error("copy should have act1 action")
	}
	if len(c.LabeledAxioms) != 1 {
		t.Error("copy should have 1 axiom")
	}
	if !c.GhostSorts["ghost"] {
		t.Error("copy should have ghost sort")
	}

	// Modify copy, verify original unchanged
	c.Actions["act2"] = &dummyAction{Tag: "new"}
	if _, ok := m.Actions["act2"]; ok {
		t.Error("modifying copy should not affect original")
	}
	c.Sig.AddSort(&lg.UninterpretedSort{Name: "extra"})
	if _, ok := m.Sig.Sorts["extra"]; ok {
		t.Error("modifying copy's sig should not affect original")
	}
}

func TestAddToHierarchy(t *testing.T) {
	m := New()
	m.AddToHierarchy("protocol")
	if m.Hierarchy["this"] == nil || !m.Hierarchy["this"]["protocol"] {
		t.Error("protocol should be under 'this'")
	}
}

func TestAddToHierarchyDotted(t *testing.T) {
	m := New()
	m.AddToHierarchy("net.protocol")
	if m.Hierarchy["this"] == nil || !m.Hierarchy["this"]["net"] {
		t.Error("net should be under 'this'")
	}
	if m.Hierarchy["net"] == nil || !m.Hierarchy["net"]["protocol"] {
		t.Error("protocol should be under 'net'")
	}
}

func TestAddObject(t *testing.T) {
	m := New()
	m.AddObject("myobj")
	if m.Hierarchy["myobj"] == nil {
		t.Error("AddObject should create hierarchy entry")
	}
}

func TestFindAction(t *testing.T) {
	m := New()
	m.Actions["send"] = &dummyAction{Tag: "action_impl"}
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

func TestIsVariant(t *testing.T) {
	m := New()
	lsort := &lg.UninterpretedSort{Name: "msg"}
	rsort := &lg.UninterpretedSort{Name: "req"}
	m.Variants["msg"] = []lg.Sort{rsort}

	if !m.IsVariant(lsort, rsort) {
		t.Error("req should be variant of msg")
	}
	if m.IsVariant(rsort, lsort) {
		t.Error("msg should not be variant of req")
	}
}

func TestVariantIndex(t *testing.T) {
	m := New()
	s1 := &lg.UninterpretedSort{Name: "a"}
	s2 := &lg.UninterpretedSort{Name: "b"}
	lsort := &lg.UninterpretedSort{Name: "msg"}
	m.Variants["msg"] = []lg.Sort{s1, s2}

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

func TestSortCard(t *testing.T) {
	m := New()
	// Enumerated sort should return card
	es := &lg.EnumeratedSort{Name: "color", Extension: []string{"red", "green", "blue"}}
	if SortCardDefault(es) != 3 {
		t.Errorf("expected 3, got %d", SortCardDefault(es))
	}

	// Function sort
	fs, _ := lg.NewFunctionSort(&lg.UninterpretedSort{Name: "t"}, lg.Boolean)
	if m.SortCard(fs) != -1 {
		t.Error("function sort should have unknown card")
	}

	// Uninterpreted sort
	us := &lg.UninterpretedSort{Name: "t"}
	if m.SortCard(us) != -1 {
		t.Error("uninterpreted sort should have unknown card")
	}
}

func TestSortDependencies(t *testing.T) {
	m := New()
	tSort := &lg.UninterpretedSort{Name: "t"}
	uSort := &lg.UninterpretedSort{Name: "u"}
	dSort, _ := lg.NewFunctionSort(tSort, uSort)
	destr := lg.NewSymbol("d", dSort)
	m.SortDestructors["t"] = []*lg.Symbol{destr}

	deps := m.SortDependencies("t", false)
	if len(deps) != 1 || deps[0] != "u" {
		t.Errorf("expected [u], got %v", deps)
	}
}

func TestSortDependenciesVariants(t *testing.T) {
	m := New()
	v1 := &lg.UninterpretedSort{Name: "v1"}
	v2 := &lg.UninterpretedSort{Name: "v2"}
	m.Variants["msg"] = []lg.Sort{v1, v2}

	deps := m.SortDependencies("msg", true)
	if len(deps) != 2 {
		t.Errorf("expected 2 variant deps, got %d", len(deps))
	}
}

func TestModuleString(t *testing.T) {
	m := New()
	s := m.String()
	if len(s) == 0 {
		t.Error("Module.String() should not be empty")
	}
}

func TestNewWithSig(t *testing.T) {
	sig := il.NewSig()
	sig.AddSort(&lg.UninterpretedSort{Name: "custom"})
	m := NewWithSig(sig)
	if _, ok := m.Sig.Sorts["custom"]; !ok {
		t.Error("module should use provided sig")
	}
}

func TestLabeledFormula(t *testing.T) {
	lf := &ast.LabeledFormula{
		Formula:  &lg.And{},
		Temporal: ast.BoolPtr(true),
		Lineno:   42,
	}
	if !lf.IsTemporal() {
		t.Error("should be temporal")
	}
	if lf.Lineno != 42 {
		t.Error("wrong lineno")
	}
}

func TestIsolateInfo(t *testing.T) {
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
		m := New()
		if sortName != "" {
			m.Sig.AddSort(&lg.UninterpretedSort{Name: sortName})
		}
		if symName != "" {
			m.Sig.AddSymbol(symName, lg.TopS)
		}
		m.GhostSorts["g"] = true
		m.Actions["a"] = &dummyAction{Tag: "v"}

		c := m.Copy()
		// Modify copy
		c.GhostSorts["g2"] = true
		c.Actions["b"] = &dummyAction{Tag: "w"}

		// Original should be unchanged
		if m.GhostSorts["g2"] {
			t.Error("original modified via copy")
		}
		if _, ok := m.Actions["b"]; ok {
			t.Error("original modified via copy")
		}
	})
}
