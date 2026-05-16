//go:build web

package webui

import (
	"fmt"
	goivy "github.com/glycerine/ivy/goivy"
	"strings"
	"testing"
)

// ---------------------------------------------------------------------------
// Helpers
// ---------------------------------------------------------------------------

func mkSort(name string) *goivy.UninterpretedSort {
	return &goivy.UninterpretedSort{Name: name}
}

func mkVar(name string, s goivy.Sort) *goivy.LogicVariable {
	v, err := goivy.NewVariable(name, s)
	if err != nil {
		panic(err)
	}
	return v
}

func mkConst(name string, s goivy.Sort) *goivy.Const {
	return goivy.NewConst(name, s)
}

func mkFuncSort(sorts ...goivy.Sort) *goivy.LogicFunctionSort {
	fs, err := goivy.NewFunctionSort(sorts...)
	if err != nil {
		panic(err)
	}
	return fs
}

func mkApply(fn goivy.Expr, args ...goivy.Expr) goivy.Expr {
	a, err := goivy.NewApply(fn, args...)
	if err != nil {
		panic(err)
	}
	return a
}

func mkEq(t1, t2 goivy.Expr) goivy.Expr {
	e, err := goivy.NewEq(t1, t2)
	if err != nil {
		panic(err)
	}
	return e
}

func mkNot(body goivy.Expr) goivy.Expr {
	n, _ := goivy.NewNot(body)
	return n
}

func mkAnd(terms ...goivy.Expr) goivy.Expr {
	a, _ := goivy.NewAnd(terms...)
	return a
}

func mkOr(terms ...goivy.Expr) goivy.Expr {
	o, _ := goivy.NewOr(terms...)
	return o
}

func mkForAll(vars []*goivy.LogicVariable, body goivy.Expr) goivy.Expr {
	f, _ := goivy.NewForAll(vars, body)
	return f
}

func mkExists(vars []*goivy.LogicVariable, body goivy.Expr) goivy.Expr {
	e, _ := goivy.NewExists(vars, body)
	return e
}

func mkImplies(t1, t2 goivy.Expr) goivy.Expr {
	i, _ := goivy.NewImplies(t1, t2)
	return i
}

// sortedConcept domain for tests.
func testDomainSetup() (*CDConceptDomain, *goivy.UninterpretedSort) {
	S := mkSort("S")
	X := mkVar("X", S)
	Y := mkVar("Y", S)

	unaryRel := mkFuncSort(S, goivy.Boolean)
	binaryRel := mkFuncSort(S, S, goivy.Boolean)

	p := mkConst("p", unaryRel)
	q := mkConst("q", unaryRel)
	r := mkConst("r", binaryRel)

	cBoth := MustCDConcept("both", []*goivy.LogicVariable{X}, mkAnd(mkApply(p, X), mkApply(q, X)))
	cOnlyP := MustCDConcept("onlyp", []*goivy.LogicVariable{X}, mkAnd(mkApply(p, X), mkNot(mkApply(q, X))))
	cOnlyQ := MustCDConcept("onlyq", []*goivy.LogicVariable{X}, mkAnd(mkNot(mkApply(p, X)), mkApply(q, X)))
	cNone := MustCDConcept("none", []*goivy.LogicVariable{X}, mkAnd(mkNot(mkApply(p, X)), mkNot(mkApply(q, X))))
	cR := MustCDConcept("r", []*goivy.LogicVariable{X, Y}, mkApply(r, X, Y))

	concepts := NewCDConceptDict()
	concepts.SetConcept("both", cBoth)
	concepts.SetConcept("none", cNone)
	concepts.SetConcept("onlyp", cOnlyP)
	concepts.SetConcept("onlyq", cOnlyQ)
	concepts.SetConcept("r", cR)
	concepts.SetList("nodes", []string{"both", "none", "onlyp", "onlyq"})
	concepts.SetList("edges", []string{"r"})
	concepts.SetList("node_labels", []string{})

	cd := NewCDConceptDomain(concepts, GetStandardCombiners(), GetStandardCombinations())
	return cd, S
}

// ---------------------------------------------------------------------------
// CDConcept Tests
// ---------------------------------------------------------------------------

func TestCDConceptCreation(t *testing.T) {
	S := mkSort("S")
	X := mkVar("X", S)
	p := mkConst("p", mkFuncSort(S, goivy.Boolean))

	c, err := NewCDConcept("test", []*goivy.LogicVariable{X}, mkApply(p, X))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if c.Name != "test" {
		t.Errorf("expected name 'test', got %q", c.Name)
	}
	if c.Arity() != 1 {
		t.Errorf("expected arity 1, got %d", c.Arity())
	}
}

func TestCDConceptCreationEmpty(t *testing.T) {
	_, err := NewCDConcept("", nil, nil)
	if err == nil {
		t.Error("expected error for empty name")
	}
}

func TestCDConceptCreationHigherOrder(t *testing.T) {
	S := mkSort("S")
	fs := mkFuncSort(S, goivy.Boolean)
	V := mkVar("V", fs)
	_, err := NewCDConcept("bad", []*goivy.LogicVariable{V}, V)
	if err == nil {
		t.Error("expected error for higher-order variable")
	}
}

func TestCDConceptArity(t *testing.T) {
	S := mkSort("S")
	X := mkVar("X", S)
	Y := mkVar("Y", S)
	eq, _ := goivy.NewEq(X, Y)
	c := MustCDConcept("eq", []*goivy.LogicVariable{X, Y}, eq)
	if c.Arity() != 2 {
		t.Errorf("expected arity 2, got %d", c.Arity())
	}
}

func TestCDConceptSorts(t *testing.T) {
	S := mkSort("S")
	T := mkSort("T")
	X := mkVar("X", S)
	Y := mkVar("Y", T)
	c := MustCDConcept("mixed", []*goivy.LogicVariable{X, Y}, goivy.True)
	sorts := c.Sorts()
	if len(sorts) != 2 {
		t.Fatalf("expected 2 sorts, got %d", len(sorts))
	}
	if sorts[0].String() != "S" || sorts[1].String() != "T" {
		t.Errorf("unexpected sorts: %v", sorts)
	}
}

func TestCDConceptSort(t *testing.T) {
	S := mkSort("S")
	X := mkVar("X", S)
	eq := webuiMustEq(X, X)
	c := MustCDConcept("self", []*goivy.LogicVariable{X}, eq)
	if c.Sort().String() != "S" {
		t.Errorf("expected sort S, got %s", c.Sort())
	}
}

func TestCDConceptCall(t *testing.T) {
	S := mkSort("S")
	X := mkVar("X", S)
	p := mkConst("p", mkFuncSort(S, goivy.Boolean))
	c := MustCDConcept("test", []*goivy.LogicVariable{X}, mkApply(p, X))

	a := mkConst("a", S)
	result, err := c.Call(a)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result == nil {
		t.Fatal("expected non-nil result")
	}
	// Result should be p(a).
	if !strings.Contains(result.String(), "p") || !strings.Contains(result.String(), "a") {
		t.Errorf("unexpected result: %s", result)
	}
}

func TestCDConceptCallWrongArity(t *testing.T) {
	S := mkSort("S")
	X := mkVar("X", S)
	eq := webuiMustEq(X, X)
	c := MustCDConcept("test", []*goivy.LogicVariable{X}, eq)

	a := mkConst("a", S)
	b := mkConst("b", S)
	_, err := c.Call(a, b)
	if err == nil {
		t.Error("expected arity error")
	}
}

func TestCDConceptCallNoArgs(t *testing.T) {
	S := mkSort("S")
	X := mkVar("X", S)
	eq := webuiMustEq(X, X)
	c := MustCDConcept("test", []*goivy.LogicVariable{X}, eq)

	_, err := c.Call()
	if err == nil {
		t.Error("expected arity error for 0 args on arity-1 concept")
	}
}

func TestCDConceptString(t *testing.T) {
	S := mkSort("S")
	X := mkVar("X", S)
	eq := webuiMustEq(X, X)
	c := MustCDConcept("test", []*goivy.LogicVariable{X}, eq)
	s := c.String()
	if !strings.Contains(s, "Concept") {
		t.Errorf("expected 'Concept' in string, got %q", s)
	}
}

func TestCDConceptFormulaStr(t *testing.T) {
	S := mkSort("S")
	X := mkVar("X", S)
	eq := webuiMustEq(X, X)
	c := MustCDConcept("test", []*goivy.LogicVariable{X}, eq)
	s := c.FormulaStr()
	if s == "" {
		t.Error("expected non-empty formula string")
	}
}

// ---------------------------------------------------------------------------
// CDConceptCombiner Tests
// ---------------------------------------------------------------------------

func TestCDConceptCombinerCall(t *testing.T) {
	S := mkSort("S")
	X := mkVar("X", S)

	unaryRel := mkFuncSort(S, goivy.Boolean)
	p := mkConst("p", unaryRel)

	// Concept: p(X)
	c := MustCDConcept("test", []*goivy.LogicVariable{X}, mkApply(p, X))

	// Combiner: "none" = ~Exists X. U(X)
	combiners := GetStandardCombiners()
	noneCombiner := combiners.GetCombiner("none")
	if noneCombiner == nil {
		t.Fatal("expected 'none' combiner")
	}

	result, err := noneCombiner.Call(c)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result == nil {
		t.Fatal("expected non-nil result")
	}
}

func TestCDConceptCombinerArity(t *testing.T) {
	combiners := GetStandardCombiners()
	none := combiners.GetCombiner("none")
	if none.Arity() != 1 {
		t.Errorf("expected arity 1 for 'none', got %d", none.Arity())
	}
	ata := combiners.GetCombiner("all_to_all")
	if ata.Arity() != 3 {
		t.Errorf("expected arity 3 for 'all_to_all', got %d", ata.Arity())
	}
}

func TestCDConceptCombinerArities(t *testing.T) {
	combiners := GetStandardCombiners()
	ata := combiners.GetCombiner("all_to_all")
	arities := ata.Arities()
	// B:2, U1:1, U2:1
	if len(arities) != 3 {
		t.Fatalf("expected 3 arities, got %d", len(arities))
	}
	if arities[0] != 2 || arities[1] != 1 || arities[2] != 1 {
		t.Errorf("expected arities [2,1,1], got %v", arities)
	}
}

func TestCDConceptCombinerCallWrongArity(t *testing.T) {
	S := mkSort("S")
	X := mkVar("X", S)
	p := mkConst("p", mkFuncSort(S, goivy.Boolean))
	c := MustCDConcept("test", []*goivy.LogicVariable{X}, mkApply(p, X))

	combiners := GetStandardCombiners()
	ata := combiners.GetCombiner("all_to_all")
	// all_to_all expects 3 concepts (B, U1, U2), give it 1
	_, err := ata.Call(c)
	if err == nil {
		t.Error("expected arity error")
	}
}

func TestCDConceptCombinerString(t *testing.T) {
	combiners := GetStandardCombiners()
	none := combiners.GetCombiner("none")
	s := none.String()
	if !strings.Contains(s, "ConceptCombiner") {
		t.Errorf("expected 'ConceptCombiner' in string, got %q", s)
	}
}

func TestCDConceptCombinerCallConceptArityMismatch(t *testing.T) {
	S := mkSort("S")
	X := mkVar("X", S)
	Y := mkVar("Y", S)
	p := mkConst("p", mkFuncSort(S, goivy.Boolean))
	// Binary concept when combiner expects unary
	cBin := MustCDConcept("bin", []*goivy.LogicVariable{X, Y}, mkApply(p, X))
	combiners := GetStandardCombiners()
	none := combiners.GetCombiner("none")
	_, err := none.Call(cBin)
	if err == nil {
		t.Error("expected arity mismatch error")
	}
}

// ---------------------------------------------------------------------------
// CDConceptDict Tests
// ---------------------------------------------------------------------------

func TestCDConceptDictBasic(t *testing.T) {
	d := NewCDConceptDict()
	if d.Len() != 0 {
		t.Error("expected empty dict")
	}

	S := mkSort("S")
	X := mkVar("X", S)
	c := MustCDConcept("test", []*goivy.LogicVariable{X}, webuiMustEq(X, X))
	d.SetConcept("test", c)

	if d.Len() != 1 {
		t.Errorf("expected len 1, got %d", d.Len())
	}
	if !d.Has("test") {
		t.Error("expected to have 'test'")
	}
	if d.GetConcept("test") != c {
		t.Error("expected to get back same concept")
	}
}

func TestCDConceptDictList(t *testing.T) {
	d := NewCDConceptDict()
	d.SetList("nodes", []string{"a", "b", "c"})
	l := d.GetList("nodes")
	if len(l) != 3 {
		t.Fatalf("expected 3 items, got %d", len(l))
	}
	if l[0] != "a" || l[1] != "b" || l[2] != "c" {
		t.Errorf("unexpected list: %v", l)
	}
}

func TestCDConceptDictAppendToList(t *testing.T) {
	d := NewCDConceptDict()
	d.SetList("nodes", []string{"a"})
	d.AppendToList("nodes", "b")
	l := d.GetList("nodes")
	if len(l) != 2 || l[1] != "b" {
		t.Errorf("expected [a b], got %v", l)
	}
}

func TestCDConceptDictDelete(t *testing.T) {
	d := NewCDConceptDict()
	d.SetList("a", []string{"1"})
	d.SetList("b", []string{"2"})
	d.SetList("c", []string{"3"})
	d.Delete("b")
	if d.Len() != 2 {
		t.Errorf("expected len 2, got %d", d.Len())
	}
	if d.Has("b") {
		t.Error("expected 'b' to be deleted")
	}
	keys := d.Keys()
	if keys[0] != "a" || keys[1] != "c" {
		t.Errorf("expected [a c], got %v", keys)
	}
}

func TestCDConceptDictCopy(t *testing.T) {
	d := NewCDConceptDict()
	S := mkSort("S")
	X := mkVar("X", S)
	c := MustCDConcept("test", []*goivy.LogicVariable{X}, webuiMustEq(X, X))
	d.SetConcept("test", c)
	d.SetList("nodes", []string{"test"})

	cp := d.Copy()
	if cp.Len() != d.Len() {
		t.Error("copy should have same length")
	}
	cp.Delete("test")
	if d.Len() != 2 { // original unchanged
		t.Error("original should be unchanged after copy modification")
	}
}

func TestCDConceptDictOrder(t *testing.T) {
	d := NewCDConceptDict()
	d.SetList("z", nil)
	d.SetList("a", nil)
	d.SetList("m", nil)
	keys := d.Keys()
	if keys[0] != "z" || keys[1] != "a" || keys[2] != "m" {
		t.Errorf("expected insertion order [z a m], got %v", keys)
	}
}

func TestCDConceptDictReorder(t *testing.T) {
	d := NewCDConceptDict()
	d.SetList("c", []string{"3"})
	d.SetList("a", []string{"1"})
	d.SetList("b", []string{"2"})

	reordered := d.Reorder([]string{"a", "b", "c"})
	keys := reordered.Keys()
	if keys[0] != "a" || keys[1] != "b" || keys[2] != "c" {
		t.Errorf("expected [a b c], got %v", keys)
	}
}

func TestCDConceptDictForEachConcept(t *testing.T) {
	d := NewCDConceptDict()
	S := mkSort("S")
	X := mkVar("X", S)
	c := MustCDConcept("test", []*goivy.LogicVariable{X}, webuiMustEq(X, X))
	d.SetConcept("test", c)
	d.SetList("nodes", []string{"test"})

	count := 0
	d.ForEachConcept(func(name string, concept *CDConcept) {
		count++
		if name != "test" {
			t.Errorf("expected name 'test', got %q", name)
		}
	})
	if count != 1 {
		t.Errorf("expected 1 concept, got %d", count)
	}
}

// ---------------------------------------------------------------------------
// CDConceptDomain Tests
// ---------------------------------------------------------------------------

func TestCDConceptDomainCopy(t *testing.T) {
	cd, _ := testDomainSetup()
	cp := cd.Copy()
	if cp.Concepts.Len() != cd.Concepts.Len() {
		t.Error("copy should have same number of concepts")
	}
	if len(cp.Combinations) != len(cd.Combinations) {
		t.Error("copy should have same number of combinations")
	}
}

func TestCDConceptDomainConceptsByArity(t *testing.T) {
	cd, _ := testDomainSetup()
	unary := cd.ConceptsByArity(1)
	if len(unary) != 4 { // both, none, onlyp, onlyq
		t.Errorf("expected 4 unary concepts, got %d: %v", len(unary), unary)
	}
	binary := cd.ConceptsByArity(2)
	if len(binary) != 1 { // r
		t.Errorf("expected 1 binary concept, got %d", len(binary))
	}
}

func TestCDConceptDomainPossibleNodeLabels(t *testing.T) {
	cd, _ := testDomainSetup()
	labels := cd.PossibleNodeLabels()
	// All unary concepts not in "nodes" list -- none since both,none,onlyp,onlyq are all in nodes
	if len(labels) != 0 {
		t.Errorf("expected 0 possible node labels, got %d: %v", len(labels), labels)
	}
}

func TestCDConceptDomainSplit(t *testing.T) {
	cd, S := testDomainSetup()

	// Create a unary concept to split by.
	X := mkVar("X", S)
	p := mkConst("p", mkFuncSort(S, goivy.Boolean))
	splitter := MustCDConcept("splitter", []*goivy.LogicVariable{X}, mkApply(p, X))
	cd.Concepts.SetConcept("splitter", splitter)

	origLen := cd.Concepts.Len()
	if err := cd.Split("both", "splitter"); err != nil {
		t.Fatal(err)
	}

	// Should have created (both+splitter) and (both-splitter), removed both.
	if cd.Concepts.Has("both") {
		t.Error("expected 'both' to be removed after split")
	}
	if !cd.Concepts.Has("(both+splitter)") {
		t.Error("expected '(both+splitter)' to exist")
	}
	if !cd.Concepts.Has("(both-splitter)") {
		t.Error("expected '(both-splitter)' to exist")
	}
	// Net: removed 1 concept ("both"), added 2 new ones = +1
	if cd.Concepts.Len() != origLen+1 {
		t.Errorf("expected %d concepts after split, got %d", origLen+1, cd.Concepts.Len())
	}
}

func TestCDConceptDomainSplitReturnsConstructorErrors(t *testing.T) {
	cd, S := testDomainSetup()
	X := mkVar("X", S)
	bad := MustCDConcept("bad", []*goivy.LogicVariable{X}, X)
	cd.Concepts.SetConcept("bad", bad)

	if err := cd.Split("both", "bad"); err == nil {
		t.Fatal("expected split to return constructor error for non-boolean split formula")
	}
}

func TestCDConceptDomainReplaceConcept(t *testing.T) {
	cd, _ := testDomainSetup()
	cd.ReplaceConcept("both", []string{"pos", "neg"})

	nodes := cd.Concepts.GetList("nodes")
	found := false
	for _, n := range nodes {
		if n == "pos" || n == "neg" {
			found = true
		}
		if n == "both" {
			t.Error("expected 'both' to be replaced in nodes list")
		}
	}
	if !found {
		t.Error("expected replacement names in nodes list")
	}
}

func TestCDConceptDomainGetFacts(t *testing.T) {
	cd, _ := testDomainSetup()
	facts, err := cd.GetFacts(func(n, c string) bool { return true })
	if err != nil {
		t.Fatalf("GetFacts: %v", err)
	}
	if len(facts) == 0 {
		t.Error("expected some facts")
	}
	// Check that we have node_info facts.
	foundNodeInfo := false
	for _, f := range facts {
		if len(f.Tag) > 0 && f.Tag[0] == "node_info" {
			foundNodeInfo = true
			break
		}
	}
	if !foundNodeInfo {
		t.Error("expected node_info facts")
	}
}

func TestCDConceptDomainGetFactsEdgeInfo(t *testing.T) {
	cd, _ := testDomainSetup()
	facts, err := cd.GetFacts(func(n, c string) bool { return true })
	if err != nil {
		t.Fatalf("GetFacts: %v", err)
	}
	foundEdgeInfo := false
	for _, f := range facts {
		if len(f.Tag) > 0 && f.Tag[0] == "edge_info" {
			foundEdgeInfo = true
			break
		}
	}
	if !foundEdgeInfo {
		t.Error("expected edge_info facts")
	}
}

func TestCDConceptDomainGetFactsNoProjection(t *testing.T) {
	cd, _ := testDomainSetup()
	facts, err := cd.GetFacts(nil)
	if err != nil {
		t.Fatalf("GetFacts: %v", err)
	}
	if len(facts) == 0 {
		t.Error("expected some facts with nil projection")
	}
}

func TestCDConceptDomainOutput(t *testing.T) {
	cd, _ := testDomainSetup()
	// Just make sure it doesn't panic.
	cd.Output()
}

// ---------------------------------------------------------------------------
// Standard Combiners Tests
// ---------------------------------------------------------------------------

func TestGetStandardCombiners(t *testing.T) {
	combiners := GetStandardCombiners()
	expectedNames := []string{
		"none", "at_least_one", "at_most_one",
		"node_necessarily", "node_necessarily_not",
		"mutually_exclusive",
		"all_to_all", "none_to_none", "total", "functional", "surjective", "injective",
	}
	for _, name := range expectedNames {
		if combiners.GetCombiner(name) == nil {
			t.Errorf("expected combiner %q to exist", name)
		}
	}
	// Check groups.
	nodeInfo := combiners.GetList("node_info")
	if len(nodeInfo) != 3 {
		t.Errorf("expected 3 node_info combiners, got %d", len(nodeInfo))
	}
	edgeInfo := combiners.GetList("edge_info")
	if len(edgeInfo) != 2 { // all_to_all, none_to_none
		t.Errorf("expected 2 edge_info combiners, got %d", len(edgeInfo))
	}
}

func TestGetStandardCombinations(t *testing.T) {
	combos := GetStandardCombinations()
	if len(combos) != 3 {
		t.Errorf("expected 3 standard combinations, got %d", len(combos))
	}
	names := []string{"node_info", "edge_info", "node_label"}
	for i, c := range combos {
		if c.Name() != names[i] {
			t.Errorf("expected combination %d name %q, got %q", i, names[i], c.Name())
		}
	}
}

// ---------------------------------------------------------------------------
// GetInitialConceptDomain Tests
// ---------------------------------------------------------------------------

func TestGetInitialConceptDomain(t *testing.T) {
	sorts := map[string]goivy.Sort{
		"node": mkSort("node"),
	}
	unaryRel := mkFuncSort(mkSort("node"), goivy.Boolean)
	binaryRel := mkFuncSort(mkSort("node"), mkSort("node"), goivy.Boolean)
	symbols := map[string]*goivy.Const{
		"link": mkConst("link", binaryRel),
		"flag": mkConst("flag", unaryRel),
	}
	cd := GetInitialConceptDomain(sorts, symbols)
	if cd == nil {
		t.Fatal("expected non-nil concept domain")
	}
	// Should have "node" sort concept + "=" + "link" + "flag" + "=link"? No.
	if !cd.Concepts.Has("node") {
		t.Error("expected 'node' concept")
	}
	if !cd.Concepts.Has("=") {
		t.Error("expected '=' concept")
	}
	if !cd.Concepts.Has("link") {
		t.Error("expected 'link' concept")
	}
	if !cd.Concepts.Has("flag") {
		t.Error("expected 'flag' concept")
	}
}

func TestGetInitialConceptDomainBinaryFunctionIsNotBooleanEdge(t *testing.T) {
	mem := mkSort("mem_type")
	clock := mkSort("tar_clock")
	value := mkSort("value")
	sorts := map[string]goivy.Sort{
		"mem_type":  mem,
		"tar_clock": clock,
		"value":     value,
	}
	predqSort := mkFuncSort(mem, clock, value)
	symbols := map[string]*goivy.Const{
		"rfn.abs.predq": mkConst("rfn.abs.predq", predqSort),
	}

	cd, err := GetInitialConceptDomainE(sorts, symbols)
	if err != nil {
		t.Fatal(err)
	}
	if stringSliceContains(cd.Concepts.GetList("edges"), "rfn.abs.predq") {
		t.Fatal("non-boolean binary function was incorrectly classified as an edge relation")
	}
	c := cd.Concepts.GetConcept("rfn.abs.predq")
	if c == nil {
		t.Fatal("expected non-boolean binary function concept")
	}
	if c.Arity() != 3 {
		t.Fatalf("expected binary function graph concept arity 3, got %d", c.Arity())
	}
	if _, err := cd.GetFacts(nil); err != nil {
		t.Fatalf("initial domain facts should not treat non-boolean binary functions as edge predicates: %v", err)
	}
}

func TestGetInitialConceptDomainEmpty(t *testing.T) {
	cd := GetInitialConceptDomain(nil, nil)
	if cd == nil {
		t.Fatal("expected non-nil concept domain even with nil inputs")
	}
	if !cd.Concepts.Has("=") {
		t.Error("expected '=' concept")
	}
}

// ---------------------------------------------------------------------------
// ConceptInteractiveSession Tests
// ---------------------------------------------------------------------------

func TestCISCreation(t *testing.T) {
	cd, _ := testDomainSetup()
	sess := NewConceptInteractiveSession(
		cd, goivy.True, goivy.True, nil, nil, nil, nil, nil, false,
	)
	if sess == nil {
		t.Fatal("expected non-nil session")
	}
	if sess.Domain == nil {
		t.Error("expected non-nil domain")
	}
}

func TestCISPushPop(t *testing.T) {
	cd, _ := testDomainSetup()
	sess := NewConceptInteractiveSession(
		cd, goivy.True, goivy.True, nil, nil, nil, nil, nil, false,
	)
	origLen := sess.Domain.Concepts.Len()
	sess.Push()
	// Modify domain.
	sess.Domain.Concepts.Delete("both")
	if sess.Domain.Concepts.Len() == origLen {
		t.Error("expected domain to be modified")
	}
	// Pop should restore.
	err := sess.Pop()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if sess.Domain.Concepts.Len() != origLen {
		t.Errorf("expected domain length %d after pop, got %d", origLen, sess.Domain.Concepts.Len())
	}
}

func TestCISPopEmpty(t *testing.T) {
	cd, _ := testDomainSetup()
	sess := NewConceptInteractiveSession(
		cd, goivy.True, goivy.True, nil, nil, nil, nil, nil, false,
	)
	err := sess.Pop()
	if err == nil {
		t.Error("expected error for pop on empty stack")
	}
}

func TestCISUndo(t *testing.T) {
	cd, _ := testDomainSetup()
	sess := NewConceptInteractiveSession(
		cd, goivy.True, goivy.True, nil, nil, nil, nil, nil, false,
	)
	origLen := sess.Domain.Concepts.Len()
	sess.Push()
	sess.Domain.Concepts.Delete("both")
	err := sess.Undo()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if sess.Domain.Concepts.Len() != origLen {
		t.Errorf("expected domain length %d after undo, got %d", origLen, sess.Domain.Concepts.Len())
	}
}

func TestCISSplit(t *testing.T) {
	cd, S := testDomainSetup()
	X := mkVar("X", S)
	p := mkConst("p", mkFuncSort(S, goivy.Boolean))
	splitter := MustCDConcept("splitter", []*goivy.LogicVariable{X}, mkApply(p, X))
	cd.Concepts.SetConcept("splitter", splitter)

	sess := NewConceptInteractiveSession(
		cd, goivy.True, goivy.True, nil, nil, nil, nil, nil, false,
	)
	if err := sess.Split("both", "splitter"); err != nil {
		t.Fatal(err)
	}
	if sess.Domain.Concepts.Has("both") {
		t.Error("expected 'both' to be removed after split")
	}
	if !sess.Domain.Concepts.Has("(both+splitter)") {
		t.Error("expected '(both+splitter)' to exist")
	}
	// Should be able to undo.
	err := sess.Undo()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !sess.Domain.Concepts.Has("both") {
		t.Error("expected 'both' to be restored after undo")
	}
}

func TestCISSupposeEmpty(t *testing.T) {
	cd, _ := testDomainSetup()
	sess := NewConceptInteractiveSession(
		cd, goivy.True, goivy.True, nil, nil, nil, nil, nil, false,
	)
	origConstraints := len(sess.SupposeConstraints)
	sess.SupposeEmpty("both")
	if len(sess.SupposeConstraints) <= origConstraints {
		t.Error("expected suppose constraint to be added")
	}
}

func TestCISRemoveConcepts(t *testing.T) {
	cd, _ := testDomainSetup()
	sess := NewConceptInteractiveSession(
		cd, goivy.True, goivy.True, nil, nil, nil, nil, nil, false,
	)
	sess.RemoveConcepts("both", "none")
	if sess.Domain.Concepts.Has("both") || sess.Domain.Concepts.Has("none") {
		t.Error("expected concepts to be removed")
	}
}

func TestCISToFormula(t *testing.T) {
	cd, _ := testDomainSetup()
	sess := NewConceptInteractiveSession(
		cd, goivy.True, goivy.True, nil, nil, nil, nil, nil, false,
	)
	f := sess.ToFormula()
	if f == nil {
		t.Error("expected non-nil formula")
	}
}

func TestCISFreshConstName(t *testing.T) {
	cd, _ := testDomainSetup()
	sess := NewConceptInteractiveSession(
		cd, goivy.True, goivy.True, nil, nil, nil, nil, nil, false,
	)
	name := sess.FreshConstName(nil)
	if name == "" {
		t.Error("expected non-empty fresh name")
	}
	// Should start with __c
	if !strings.HasPrefix(name, "__c") {
		t.Errorf("expected fresh name to start with '__c', got %q", name)
	}
}

func TestCISClone(t *testing.T) {
	cd, _ := testDomainSetup()
	sess := NewConceptInteractiveSession(
		cd, goivy.True, goivy.True, nil, nil, nil, nil, nil, false,
	)
	clone := sess.Clone(false)
	if clone == nil {
		t.Fatal("expected non-nil clone")
	}
	if clone.Domain == sess.Domain {
		t.Error("expected clone to have different domain pointer")
	}
}

func TestCISSaveDomain(t *testing.T) {
	cd, _ := testDomainSetup()
	sess := NewConceptInteractiveSession(
		cd, goivy.True, goivy.True, nil, nil, nil, nil, nil, false,
	)
	sess.SaveDomain("test")
	if _, ok := sess.AnalysisSession["test"]; !ok {
		t.Error("expected domain to be saved")
	}
}

func TestCISLoadDomain(t *testing.T) {
	cd, _ := testDomainSetup()
	sess := NewConceptInteractiveSession(
		cd, goivy.True, goivy.True, nil, nil, nil, nil, nil, false,
	)
	sess.SaveDomain("test")
	sess.Domain.Concepts.Delete("both")
	err := sess.LoadDomain("test")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !sess.Domain.Concepts.Has("both") {
		t.Error("expected 'both' to be restored")
	}
}

func TestCISLoadDomainNotFound(t *testing.T) {
	cd, _ := testDomainSetup()
	sess := NewConceptInteractiveSession(
		cd, goivy.True, goivy.True, nil, nil, nil, nil, nil, false,
	)
	err := sess.LoadDomain("nonexistent")
	if err == nil {
		t.Error("expected error for nonexistent domain")
	}
}

func TestCISReplaceDomain(t *testing.T) {
	cd, _ := testDomainSetup()
	sess := NewConceptInteractiveSession(
		cd, goivy.True, goivy.True, nil, nil, nil, nil, nil, false,
	)
	newDomain := cd.Copy()
	newDomain.Concepts.Delete("both")
	if err := sess.ReplaceDomain(newDomain, nil); err != nil {
		t.Fatalf("ReplaceDomain: %v", err)
	}
	if sess.Domain.Concepts.Has("both") {
		t.Error("expected 'both' to be gone after replace")
	}
}

func TestCISMaterializeNode(t *testing.T) {
	cd, _ := testDomainSetup()
	sess := NewConceptInteractiveSession(
		cd, goivy.True, goivy.True, nil, nil, nil, nil, nil, false,
	)
	// Materialize should push and recompute without error.
	if err := sess.MaterializeNode("both"); err != nil {
		t.Fatalf("MaterializeNode: %v", err)
	}
	if len(sess.UndoStack) != 1 {
		t.Errorf("expected 1 undo entry, got %d", len(sess.UndoStack))
	}
}

func TestCISMaterializeEdge(t *testing.T) {
	cd, _ := testDomainSetup()
	sess := NewConceptInteractiveSession(
		cd, goivy.True, goivy.True, nil, nil, nil, nil, nil, false,
	)
	if err := sess.MaterializeEdge("r", "both", "none", true); err != nil {
		t.Fatalf("MaterializeEdge: %v", err)
	}
	if len(sess.UndoStack) != 1 {
		t.Errorf("expected 1 undo entry, got %d", len(sess.UndoStack))
	}
}

func TestCISSuppose(t *testing.T) {
	cd, _ := testDomainSetup()
	sess := NewConceptInteractiveSession(
		cd, goivy.True, goivy.True, nil, nil, nil, nil, nil, false,
	)
	S := mkSort("S")
	a := mkConst("a", S)
	b := mkConst("b", S)
	eq := mkEq(a, b)
	sess.Suppose(eq)
	if len(sess.SupposeConstraints) != 1 {
		t.Errorf("expected 1 suppose constraint, got %d", len(sess.SupposeConstraints))
	}
}

func TestCISSupposeTautology(t *testing.T) {
	cd, _ := testDomainSetup()
	sess := NewConceptInteractiveSession(
		cd, goivy.True, goivy.True, nil, nil, nil, nil, nil, false,
	)
	S := mkSort("S")
	a := mkConst("a", S)
	eq := mkEq(a, a) // tautology
	sess.Suppose(eq)
	if len(sess.SupposeConstraints) != 0 {
		t.Errorf("expected 0 suppose constraints for tautology, got %d", len(sess.SupposeConstraints))
	}
}

func TestCISAddEdge(t *testing.T) {
	cd, S := testDomainSetup()
	sess := NewConceptInteractiveSession(
		cd, goivy.True, goivy.True, nil, nil, nil, nil, nil, false,
	)
	X := mkVar("X", S)
	Y := mkVar("Y", S)
	eq := mkEq(X, Y)
	c := MustCDConcept("newEdge", []*goivy.LogicVariable{X, Y}, eq)
	sess.AddEdge("newEdge", c)
	if !sess.Domain.Concepts.Has("newEdge") {
		t.Error("expected 'newEdge' to exist")
	}
	edges := sess.Domain.Concepts.GetList("edges")
	found := false
	for _, e := range edges {
		if e == "newEdge" {
			found = true
		}
	}
	if !found {
		t.Error("expected 'newEdge' in edges list")
	}
}

func TestCISAddCustomEdge(t *testing.T) {
	cd, _ := testDomainSetup()
	sess := NewConceptInteractiveSession(
		cd, goivy.True, goivy.True, nil, nil, nil, nil, nil, false,
	)
	origCombinations := len(sess.Domain.Combinations)
	sess.AddCustomEdge("r", "both", "none")
	if len(sess.Domain.Combinations) != origCombinations+1 {
		t.Errorf("expected %d combinations, got %d", origCombinations+1, len(sess.Domain.Combinations))
	}
}

func TestCISAddCustomNodeLabel(t *testing.T) {
	cd, _ := testDomainSetup()
	sess := NewConceptInteractiveSession(
		cd, goivy.True, goivy.True, nil, nil, nil, nil, nil, false,
	)
	origCombinations := len(sess.Domain.Combinations)
	sess.AddCustomNodeLabel("both", "someLabel")
	if len(sess.Domain.Combinations) != origCombinations+1 {
		t.Errorf("expected %d combinations, got %d", origCombinations+1, len(sess.Domain.Combinations))
	}
}

// ---------------------------------------------------------------------------
// WebUIAlpha Tests
// ---------------------------------------------------------------------------

func TestAlphaNoSolver(t *testing.T) {
	cd, _ := testDomainSetup()
	cache := map[string]bool{
		"node_info|none|both": true,
	}
	result, err := WebUIAlphaNoSolver(cd, cache, func(string, string) bool { return true })
	if err != nil {
		t.Fatalf("WebUIAlphaNoSolver: %v", err)
	}
	if len(result) == 0 {
		t.Error("expected some results")
	}
	// Check that cached value is used.
	found := false
	for _, tv := range result {
		if TagString(tv.Tag) == "node_info|none|both" && tv.Value {
			found = true
		}
	}
	if !found {
		t.Error("expected cached value for node_info|none|both to be true")
	}
}

func TestAlphaNoSolverEmptyCache(t *testing.T) {
	cd, _ := testDomainSetup()
	result, err := WebUIAlphaNoSolver(cd, nil, func(string, string) bool { return true })
	if err != nil {
		t.Fatalf("WebUIAlphaNoSolver: %v", err)
	}
	if len(result) == 0 {
		t.Error("expected some results")
	}
	// All should be false with empty cache.
	for _, tv := range result {
		if tv.Value {
			t.Errorf("expected all values false with empty cache, got true for %v", tv.Tag)
		}
	}
}

// ---------------------------------------------------------------------------
// Tag Tests
// ---------------------------------------------------------------------------

func TestTagString(t *testing.T) {
	tag := Tag{"node_info", "none", "both"}
	s := TagString(tag)
	if s != "node_info|none|both" {
		t.Errorf("expected 'node_info|none|both', got %q", s)
	}
}

func TestTagFromString(t *testing.T) {
	tag := TagFromString("edge_info|all_to_all|r|both|none")
	if len(tag) != 5 {
		t.Fatalf("expected 5 elements, got %d", len(tag))
	}
	if tag[0] != "edge_info" || tag[4] != "none" {
		t.Errorf("unexpected tag: %v", tag)
	}
}

// ---------------------------------------------------------------------------
// Combination Tests
// ---------------------------------------------------------------------------

func TestCombination(t *testing.T) {
	c := NewCombination("node_info", "node_info", "nodes")
	if c.Name() != "node_info" {
		t.Errorf("expected name 'node_info', got %q", c.Name())
	}
	if c.CombinerClass() != "node_info" {
		t.Errorf("expected combiner class 'node_info', got %q", c.CombinerClass())
	}
	sets := c.ConceptSets()
	if len(sets) != 1 || sets[0] != "nodes" {
		t.Errorf("expected concept sets [nodes], got %v", sets)
	}
}

// ---------------------------------------------------------------------------
// Cartesian Product Tests
// ---------------------------------------------------------------------------

func TestCartesianProduct(t *testing.T) {
	result := cartesianProduct([][]string{{"a", "b"}, {"1", "2"}})
	if len(result) != 4 {
		t.Errorf("expected 4 products, got %d", len(result))
	}
}

func TestCartesianProductEmpty(t *testing.T) {
	result := cartesianProduct(nil)
	if len(result) != 1 || len(result[0]) != 0 {
		t.Errorf("expected [[]], got %v", result)
	}
}

func TestCartesianProductSingle(t *testing.T) {
	result := cartesianProduct([][]string{{"a", "b", "c"}})
	if len(result) != 3 {
		t.Errorf("expected 3, got %d", len(result))
	}
}

// ---------------------------------------------------------------------------
// Union Lists Tests
// ---------------------------------------------------------------------------

func TestUnionLists(t *testing.T) {
	result := cdUnionLists([][]string{{"a", "b"}, {"b", "c"}, {"a", "d"}})
	if len(result) != 4 {
		t.Errorf("expected 4 unique items, got %d: %v", len(result), result)
	}
	if result[0] != "a" || result[1] != "b" || result[2] != "c" || result[3] != "d" {
		t.Errorf("unexpected order: %v", result)
	}
}

// ---------------------------------------------------------------------------
// Concept Space Tests
// ---------------------------------------------------------------------------

func TestToConceptSpaceAtom(t *testing.T) {
	node, err := ToConceptSpace("foo(X, Y)")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if node == nil {
		t.Fatal("expected non-nil node")
	}
	ns, ok := node.(*NamedSpace)
	if !ok {
		t.Fatalf("expected NamedSpace, got %T", node)
	}
	if ns.Lit.Atom.RelName != "foo" {
		t.Errorf("expected relname 'foo', got %q", ns.Lit.Atom.RelName)
	}
	if len(ns.Lit.Atom.Args) != 2 {
		t.Errorf("expected 2 args, got %d", len(ns.Lit.Atom.Args))
	}
}

func TestToConceptSpaceNegated(t *testing.T) {
	node, err := ToConceptSpace("~foo(X)")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	ns, ok := node.(*NamedSpace)
	if !ok {
		t.Fatalf("expected NamedSpace, got %T", node)
	}
	if ns.Lit.Polarity != 0 {
		t.Errorf("expected polarity 0 for negated, got %d", ns.Lit.Polarity)
	}
}

func TestToConceptSpaceProduct(t *testing.T) {
	node, err := ToConceptSpace("(foo(X) * bar(Y))")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	ps, ok := node.(*ProductSpace)
	if !ok {
		t.Fatalf("expected ProductSpace, got %T", node)
	}
	if len(ps.Spaces) != 2 {
		t.Errorf("expected 2 spaces in product, got %d", len(ps.Spaces))
	}
}

func TestToConceptSpaceSum(t *testing.T) {
	node, err := ToConceptSpace("(foo(X) + bar(Y))")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	ss, ok := node.(*SumSpace)
	if !ok {
		t.Fatalf("expected SumSpace, got %T", node)
	}
	if len(ss.Spaces) != 2 {
		t.Errorf("expected 2 spaces in sum, got %d", len(ss.Spaces))
	}
}

func TestToConceptSpaceNested(t *testing.T) {
	node, err := ToConceptSpace("((foo(X) * bar(Y)) + baz(Z))")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	ss, ok := node.(*SumSpace)
	if !ok {
		t.Fatalf("expected SumSpace, got %T", node)
	}
	if len(ss.Spaces) != 2 {
		t.Errorf("expected 2 spaces in sum, got %d", len(ss.Spaces))
	}
}

func TestToConceptSpaceSimpleSymbol(t *testing.T) {
	node, err := ToConceptSpace("foo")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	ns, ok := node.(*NamedSpace)
	if !ok {
		t.Fatalf("expected NamedSpace, got %T", node)
	}
	if ns.Lit.Atom.RelName != "foo" {
		t.Errorf("expected relname 'foo', got %q", ns.Lit.Atom.RelName)
	}
}

func TestNamedSpaceEnumerate(t *testing.T) {
	atom := &CSAtom{RelName: "foo", Args: []CSTerm{{Name: "X", IsVariable: true}}}
	lit := &CSLiteral{Polarity: 1, Atom: atom}
	ns := &NamedSpace{Lit: lit}

	result := ns.Enumerate(nil, func(lits []*CSLiteral) bool { return true })
	if len(result) != 1 {
		t.Errorf("expected 1 clause, got %d", len(result))
	}
}

func TestProductSpaceEnumerate(t *testing.T) {
	ns1 := &NamedSpace{Lit: &CSLiteral{Polarity: 1, Atom: &CSAtom{RelName: "a"}}}
	ns2 := &NamedSpace{Lit: &CSLiteral{Polarity: 1, Atom: &CSAtom{RelName: "b"}}}
	ps := &ProductSpace{Spaces: []CSNode{ns1, ns2}}

	result := ps.Enumerate(nil, func(lits []*CSLiteral) bool { return true })
	if len(result) != 1 { // single product
		t.Errorf("expected 1 clause, got %d", len(result))
	}
	if len(result[0]) != 2 {
		t.Errorf("expected clause of length 2, got %d", len(result[0]))
	}
}

func TestSumSpaceEnumerate(t *testing.T) {
	ns1 := &NamedSpace{Lit: &CSLiteral{Polarity: 1, Atom: &CSAtom{RelName: "a"}}}
	ns2 := &NamedSpace{Lit: &CSLiteral{Polarity: 1, Atom: &CSAtom{RelName: "b"}}}
	ss := &SumSpace{Spaces: []CSNode{ns1, ns2}}

	result := ss.Enumerate(nil, func(lits []*CSLiteral) bool { return true })
	if len(result) != 2 {
		t.Errorf("expected 2 clauses, got %d", len(result))
	}
}

func TestProductSpaceEnumerateEmpty(t *testing.T) {
	ps := &ProductSpace{Spaces: nil}
	result := ps.Enumerate(nil, func(lits []*CSLiteral) bool { return true })
	if len(result) != 1 || len(result[0]) != 0 {
		t.Errorf("expected [[]], got %v", result)
	}
}

func TestCSLiteralNegate(t *testing.T) {
	atom := &CSAtom{RelName: "foo"}
	lit := &CSLiteral{Polarity: 1, Atom: atom}
	neg := lit.Negate()
	if neg.Polarity != 0 {
		t.Error("expected negated polarity 0")
	}
	negNeg := neg.Negate()
	if negNeg.Polarity != 1 {
		t.Error("expected double-negated polarity 1")
	}
}

// ---------------------------------------------------------------------------
// GetStructureConceptDomain Tests
// ---------------------------------------------------------------------------

func TestGetStructureConceptDomain(t *testing.T) {
	S := mkSort("S")
	universe := map[string][]*goivy.Const{
		"S": {mkConst("s0", S), mkConst("s1", S)},
	}
	cd := GetStructureConceptDomain(goivy.True, universe, nil)
	if cd == nil {
		t.Fatal("expected non-nil concept domain")
	}
	nodes := cd.Concepts.GetList("nodes")
	if len(nodes) < 2 {
		t.Errorf("expected at least 2 nodes, got %d", len(nodes))
	}
}

func TestGetStructureConceptDomainListsConcreteLabelsAndRelations(t *testing.T) {
	S := mkSort("S")
	s0 := mkConst("0", S)
	rel := mkConst("r", mkFuncSort(S, S, goivy.Boolean))
	witness := mkConst("@X", S)
	universe := map[string][]*goivy.Const{"S": {s0}}
	state := mkAnd(mkEq(witness, s0))

	cd := GetStructureConceptDomain(state, universe, map[string]*goivy.Const{"r": rel})
	labels := cd.Concepts.GetList("node_labels")
	if !stringSliceContains(labels, "="+UniverseElementToConceptName(s0)) {
		t.Fatalf("node_labels = %v, missing numeric concrete label", labels)
	}
	if !stringSliceContains(labels, "=@X") {
		t.Fatalf("node_labels = %v, missing CTI witness constant", labels)
	}
	if !stringSliceContains(cd.Concepts.GetList("edges"), "r") {
		t.Fatalf("edges = %v, missing binary relation r", cd.Concepts.GetList("edges"))
	}

	av := GetStructureConceptAbstractValue(state, universe)
	nodeName := UniverseElementToConceptName(s0)
	if !av["node_label|node_necessarily|"+nodeName+"|=@X"] {
		t.Fatalf("abstract value did not attach @X to concrete node %q: %v", nodeName, av)
	}
}

func TestGetStructureConceptAbstractValue(t *testing.T) {
	S := mkSort("S")
	s0 := mkConst("s0", S)
	s1 := mkConst("s1", S)
	universe := map[string][]*goivy.Const{
		"S": {s0, s1},
	}
	state := mkAnd(mkEq(s0, s0), mkEq(s1, s1))
	av := GetStructureConceptAbstractValue(state, universe)
	if len(av) == 0 {
		t.Error("expected non-empty abstract value")
	}
}

func TestUniverseElementToConceptName(t *testing.T) {
	S := mkSort("S")
	c := mkConst("s0", S)
	name := UniverseElementToConceptName(c)
	if !strings.Contains(name, "s0") {
		t.Errorf("expected name to contain 's0', got %q", name)
	}
	if !strings.Contains(name, "S") {
		t.Errorf("expected name to contain 'S', got %q", name)
	}
}

func TestUniverseElementToConceptNameAlreadyContainsSort(t *testing.T) {
	S := mkSort("S")
	c := mkConst("s0:S", S)
	name := UniverseElementToConceptName(c)
	// Should not duplicate the sort
	if strings.Count(name, "S") > 1 {
		t.Errorf("sort should not be duplicated: %q", name)
	}
}

// ---------------------------------------------------------------------------
// GetDiagramConceptDomain Tests
// ---------------------------------------------------------------------------

func TestGetDiagramConceptDomain(t *testing.T) {
	S := mkSort("S")
	c := mkConst("foo", mkFuncSort(S, goivy.Boolean))
	cd := GetDiagramConceptDomain(nil, []*goivy.Const{c}, nil)
	if cd == nil {
		t.Fatal("expected non-nil concept domain")
	}
	if !cd.Concepts.Has("foo") {
		t.Error("expected 'foo' concept")
	}
}

// ---------------------------------------------------------------------------
// GetStructureRenaming Tests
// ---------------------------------------------------------------------------

func TestGetStructureRenaming(t *testing.T) {
	S := mkSort("S")
	s0 := mkConst("s0", S)
	s1 := mkConst("s1", S)
	universe := map[string][]*goivy.Const{
		"S": {s0, s1},
	}
	state := mkAnd(mkEq(s0, s0))
	result := GetStructureRenaming(state, universe, nil)
	if len(result) == 0 {
		t.Error("expected non-empty renaming")
	}
}

// ---------------------------------------------------------------------------
// CDCombinerDict Tests
// ---------------------------------------------------------------------------

func TestCDCombinerDictCopy(t *testing.T) {
	d := GetStandardCombiners()
	cp := d.Copy()
	if cp.GetCombiner("none") == nil {
		t.Error("expected 'none' combiner in copy")
	}
}

func TestCDCombinerDictHas(t *testing.T) {
	d := GetStandardCombiners()
	if !d.Has("none") {
		t.Error("expected to have 'none'")
	}
	if d.Has("nonexistent") {
		t.Error("expected not to have 'nonexistent'")
	}
}

// ---------------------------------------------------------------------------
// Fuzz Tests
// ---------------------------------------------------------------------------

func FuzzToConceptSpace(f *testing.F) {
	f.Add("foo(X)")
	f.Add("~bar(Y)")
	f.Add("(a(X) + b(Y))")
	f.Add("(a(X) * b(Y))")
	f.Add("((a(X) * b(Y)) + c(Z))")
	f.Add("")
	f.Add("foo")
	f.Add("~")
	f.Add("()")
	f.Add("(a + b + c)")

	f.Fuzz(func(t *testing.T, s string) {
		// Just ensure no panic.
		_, _ = ToConceptSpace(s)
	})
}

func FuzzCDConceptCall(f *testing.F) {
	f.Add("test", 1)
	f.Add("eq", 2)
	f.Add("x", 0)
	f.Add("longname", 3)

	f.Fuzz(func(t *testing.T, name string, arity int) {
		if arity < 0 || arity > 10 || name == "" {
			return
		}
		// Skip names that would fail NewVariable (needs uppercase first char).
		S := mkSort("S")
		varNames := []string{"X", "Y", "Z", "W", "V", "U", "A", "B", "C", "D"}
		if arity > len(varNames) {
			return
		}
		vars := make([]*goivy.LogicVariable, arity)
		for i := 0; i < arity; i++ {
			vars[i] = mkVar(varNames[i], S)
		}
		var formula goivy.Expr
		if arity == 0 {
			formula = goivy.True
		} else {
			formula = webuiMustEq(vars[0], vars[0])
		}
		c, err := NewCDConcept(name, vars, formula)
		if err != nil {
			return
		}
		// Try calling with right number of args.
		args := make([]goivy.Expr, arity)
		for i := 0; i < arity; i++ {
			args[i] = mkConst(fmt.Sprintf("c%d", i), S)
		}
		result, err := c.Call(args...)
		if err != nil {
			t.Errorf("unexpected error calling concept %q with %d args: %v", name, arity, err)
		}
		_ = result
	})
}

// ---------------------------------------------------------------------------
// Additional edge case tests
// ---------------------------------------------------------------------------

func TestCDConceptCallBinary(t *testing.T) {
	S := mkSort("S")
	X := mkVar("X", S)
	Y := mkVar("Y", S)
	r := mkConst("r", mkFuncSort(S, S, goivy.Boolean))
	c := MustCDConcept("rel", []*goivy.LogicVariable{X, Y}, mkApply(r, X, Y))

	a := mkConst("a", S)
	b := mkConst("b", S)
	result, err := c.Call(a, b)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result == nil {
		t.Fatal("expected non-nil result")
	}
	// Should be r(a, b)
	s := result.String()
	if !strings.Contains(s, "r") {
		t.Errorf("expected 'r' in result, got %q", s)
	}
}

func TestCDConceptDomainGetFactsProjectionFilter(t *testing.T) {
	cd, _ := testDomainSetup()
	// Only allow "both" concept.
	facts, err := cd.GetFacts(func(n, c string) bool { return n == "both" })
	if err != nil {
		t.Fatalf("GetFacts: %v", err)
	}
	for _, f := range facts {
		if len(f.Tag) >= 3 && f.Tag[0] == "node_info" && f.Tag[2] != "both" {
			t.Errorf("expected only 'both' in node_info facts, got %v", f.Tag)
		}
	}
}

func TestCDConceptDomainGetCombFacts(t *testing.T) {
	cd, _ := testDomainSetup()
	var facts []Fact
	if err := cd.GetCombFacts("test", "node_info", [][]string{{"both"}}, &facts); err != nil {
		t.Fatalf("GetCombFacts: %v", err)
	}
	if len(facts) == 0 {
		t.Error("expected some comb facts")
	}
}
