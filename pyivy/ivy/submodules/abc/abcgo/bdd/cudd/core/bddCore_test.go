package core

import (
	"math"
	"reflect"
	"testing"
)

func bddTruth(t *testing.T, dd *DdManager, n *DdNode, bits int) bool {
	t.Helper()
	a := make([]bool, len(dd.Vars))
	for i := range a {
		a[i] = bits&(1<<i) != 0
	}
	v, err := dd.Evaluate(n, a)
	if err != nil {
		t.Fatal(err)
	}
	return v
}
func TestBDDCanonicalLogic(t *testing.T) {
	dd := NewBDDManager(3)
	x, y, z := dd.Vars[0], dd.Vars[1], dd.Vars[2]
	if Cudd_Not(Cudd_Not(x)) != x || Cudd_Regular(Cudd_Not(x)) != x || !Cudd_IsComplement(Cudd_Not(x)) {
		t.Fatal("complement handles")
	}
	if Cudd_bddAnd(dd, x, Cudd_Not(x)) != dd.LogicZero || Cudd_bddOr(dd, x, Cudd_Not(x)) != dd.One {
		t.Fatal("complement laws")
	}
	lhs := Cudd_bddAnd(dd, x, Cudd_bddOr(dd, y, z))
	rhs := Cudd_bddOr(dd, Cudd_bddAnd(dd, x, y), Cudd_bddAnd(dd, x, z))
	if lhs != rhs {
		t.Fatal("noncanonical distributivity")
	}
	for bits := 0; bits < 8; bits++ {
		a, b, c := bits&1 != 0, bits&2 != 0, bits&4 != 0
		for _, tc := range []struct {
			n    *DdNode
			want bool
		}{{lhs, a && (b || c)}, {Cudd_bddIte(dd, x, y, z), a && b || !a && c}, {Cudd_bddXor(dd, x, y), a != b}, {Cudd_bddXnor(dd, x, y), a == b}, {Cudd_bddNand(dd, x, y), !(a && b)}} {
			if got := bddTruth(t, dd, tc.n, bits); got != tc.want {
				t.Fatalf("bits %d got %v want %v", bits, got, tc.want)
			}
		}
	}
	for _, n := range dd.unique {
		if Cudd_IsComplement(n.T) || n.T == n.E || dd.level(n.T) <= dd.level(n) || dd.level(n.E) <= dd.level(n) {
			t.Fatal("ROBDD invariant")
		}
	}
}
func TestBDDPermuteComposeAndCofactor(t *testing.T) {
	dd := NewBDDManager(3)
	x, y, z := dd.Vars[0], dd.Vars[1], dd.Vars[2]
	f := Cudd_bddAnd(dd, x, Cudd_Not(y))
	p := Cudd_bddPermute(dd, f, []int{2, 0, 1})
	if p != Cudd_bddAnd(dd, z, Cudd_Not(x)) {
		t.Fatal("simultaneous permutation")
	}
	if Cudd_bddCompose(dd, f, z, 0) != Cudd_bddAnd(dd, z, Cudd_Not(y)) {
		t.Fatal("composition")
	}
	if Cudd_Cofactor(dd, f, x) != Cudd_Not(y) || Cudd_Cofactor(dd, f, Cudd_Not(x)) != dd.LogicZero {
		t.Fatal("cofactor polarity")
	}
	if Cudd_Cofactor(dd, f, Cudd_bddOr(dd, x, y)) != nil {
		t.Fatal("accepted noncube")
	}
	if Cudd_bddPermute(dd, f, []int{0}) != nil {
		t.Fatal("accepted incomplete permutation")
	}
}
func TestBDDQuantificationSupportAndCount(t *testing.T) {
	dd := NewBDDManager(3)
	x, y, z := dd.Vars[0], dd.Vars[1], dd.Vars[2]
	f := Cudd_bddOr(dd, Cudd_bddAnd(dd, x, y), Cudd_bddAnd(dd, Cudd_Not(x), z))
	if Cudd_bddExistAbstract(dd, f, x) != Cudd_bddOr(dd, y, z) {
		t.Fatal("existential")
	}
	if Cudd_bddUnivAbstract(dd, f, x) != Cudd_bddAnd(dd, y, z) {
		t.Fatal("universal")
	}
	if Cudd_bddAndAbstract(dd, f, x, x) != y {
		t.Fatal("and abstraction")
	}
	if !reflect.DeepEqual(dd.Support(f), []int{0, 1, 2}) || Cudd_SupportSize(dd, f) != 3 {
		t.Fatal("support")
	}
	if Cudd_CountMinterm(dd, f, 3) != 4 || Cudd_CountMinterm(dd, Cudd_Not(f), 3) != 4 {
		t.Fatal("minterms")
	}
	if Cudd_DagSize(x) != 2 || Cudd_DagSize(Cudd_Not(x)) != 2 {
		t.Fatal("DAG count regularization")
	}
	if !Cudd_bddLeq(dd, Cudd_bddAnd(dd, x, y), x) || Cudd_bddLeq(dd, x, y) {
		t.Fatal("implication")
	}
}
func TestBDDManagerSafetyAndReferences(t *testing.T) {
	dd := Cudd_Init(1, 0, 256, 256, 0)
	x := dd.Vars[0]
	Cudd_Ref(Cudd_Not(x))
	if x.Ref != 1 {
		t.Fatal("reference regularization")
	}
	Cudd_RecursiveDeref(dd, x)
	Cudd_Deref(x)
	if x.Ref != 0 {
		t.Fatal("underflow")
	}
	if Cudd_bddAnd(dd, x, NewBDDManager(1).Vars[0]) != nil {
		t.Fatal("foreign manager accepted")
	}
	if _, err := dd.Evaluate(x, nil); err == nil {
		t.Fatal("missing assignment accepted")
	}
	if Cudd_bddIthVar(dd, 4) == nil || Cudd_ReadSize(dd) != 5 || Cudd_ReadInvPerm(dd, 4) != 4 {
		t.Fatal("variable growth")
	}
	Cudd_Quit(dd)
	if Cudd_bddAnd(dd, x, x) != nil {
		t.Fatal("closed manager accepted")
	}
}

func TestBDDConstantQueryWitnessAndLimit(t *testing.T) {
	dd := NewBDDManager(3)
	x, y, z := dd.Vars[0], dd.Vars[1], dd.Vars[2]
	before := len(dd.unique)
	if Cudd_bddIteConstant(dd, x, x, dd.One) != dd.One {
		t.Fatal("constant true query")
	}
	if Cudd_bddIteConstant(dd, x, y, Cudd_Not(y)) != DD_NON_CONSTANT {
		t.Fatal("nonconstant query")
	}
	if len(dd.unique) != before {
		t.Fatal("constant query allocated nodes")
	}
	if Cudd_bddAndLimit(dd, x, y, 0) != nil || len(dd.unique) != before {
		t.Fatal("limit did not rollback")
	}
	conjunction := Cudd_bddAndLimit(dd, x, y, 1)
	if conjunction == nil {
		t.Fatal("one-node limit failed")
	}
	if Cudd_bddAndLimit(dd, x, y, 0) != conjunction {
		t.Fatal("cached result should not allocate")
	}
	f, g := Cudd_bddOr(dd, x, y), Cudd_bddOr(dd, Cudd_Not(x), z)
	witness := Cudd_bddIntersect(dd, f, g)
	if witness == dd.LogicZero || !Cudd_bddLeq(dd, witness, Cudd_bddAnd(dd, f, g)) {
		t.Fatal("invalid intersection witness")
	}
}

func TestBDDRandomizedCanonicalTruthTables(t *testing.T) {
	dd := NewBDDManager(4)
	type entry struct {
		node  *DdNode
		table uint16
	}
	entries := []entry{{dd.One, 0xffff}, {dd.LogicZero, 0}}
	for i, v := range dd.Vars {
		var table uint16
		for bits := 0; bits < 16; bits++ {
			if bits&(1<<i) != 0 {
				table |= 1 << bits
			}
		}
		entries = append(entries, entry{v, table})
	}
	canonical := map[uint16]*DdNode{}
	for _, e := range entries {
		canonical[e.table] = e.node
	}
	// A deterministic PRNG keeps failures reproducible without seed state.
	seed := uint32(0x12ab341f)
	next := func() uint32 { seed = seed*1664525 + 1013904223; return seed }
	for i := 0; i < 1500; i++ {
		a, b, c := entries[int(next())%len(entries)], entries[int(next())%len(entries)], entries[int(next())%len(entries)]
		var n *DdNode
		var table uint16
		switch next() % 5 {
		case 0:
			n = Cudd_bddAnd(dd, a.node, b.node)
			table = a.table & b.table
		case 1:
			n = Cudd_bddOr(dd, a.node, b.node)
			table = a.table | b.table
		case 2:
			n = Cudd_bddXor(dd, a.node, b.node)
			table = a.table ^ b.table
		case 3:
			n = Cudd_Not(a.node)
			table = ^a.table
		case 4:
			n = Cudd_bddIte(dd, a.node, b.node, c.node)
			table = (a.table & b.table) | (^a.table & c.table)
		}
		for bits := 0; bits < 16; bits++ {
			if bddTruth(t, dd, n, bits) != (table&(1<<bits) != 0) {
				t.Fatalf("step %d assignment %d", i, bits)
			}
		}
		if old := canonical[table]; old != nil && old != n {
			t.Fatalf("noncanonical truth table %04x", table)
		}
		canonical[table] = n
		entries = append(entries, entry{n, table})
	}
}
func TestZDDUniverseInitialization(t *testing.T) {
	dd := Cudd_Init(2, 4, 256, 256, 0)
	if dd.ZddSize != 4 || len(dd.Univ) != 4 || dd.Zero == dd.LogicZero {
		t.Fatal("BDD/ZDD constant distinction")
	}
	var count func(*DdNode) int
	count = func(n *DdNode) int {
		if n == dd.Zero {
			return 0
		}
		if n == dd.One {
			return 1
		}
		return count(n.T) + count(n.E)
	}
	for i, n := range dd.Univ {
		if n.Index != uint32(i) || count(n) != 1<<(4-i) {
			t.Fatal("wrong universe", i, count(n))
		}
	}
	if dd.zddNode(0, dd.Zero, dd.One) != dd.One {
		t.Fatal("zero suppression")
	}
	if Cudd_bddAnd(dd, dd.Zero, dd.One) != nil || Cudd_bddAnd(dd, dd.Univ[0], dd.One) != nil {
		t.Fatal("non-BDD edges accepted by BDD APIs")
	}
}
func TestBDDAllocationLimitPartialRollback(t *testing.T) {
	dd := NewBDDManager(4)
	f := Cudd_bddXor(dd, dd.Vars[0], dd.Vars[2])
	g := Cudd_bddXor(dd, dd.Vars[1], dd.Vars[3])
	before, id := len(dd.unique), dd.nextID
	if Cudd_bddAndLimit(dd, f, g, 1) != nil || len(dd.unique) != before || dd.nextID != id {
		t.Fatal("partial allocation rollback")
	}
	n := Cudd_bddAnd(dd, f, g)
	if n == nil {
		t.Fatal("failed after rollback")
	}
	for bits := 0; bits < 16; bits++ {
		if bddTruth(t, dd, n, bits) != (bddTruth(t, dd, f, bits) && bddTruth(t, dd, g, bits)) {
			t.Fatal(bits)
		}
	}
}

func TestBDDBooleanDifferenceDependencyAndCubeValidation(t *testing.T) {
	dd := NewBDDManager(3)
	x, y, z := dd.Vars[0], dd.Vars[1], dd.Vars[2]
	f := Cudd_bddXnor(dd, x, Cudd_bddOr(dd, y, z))
	if Cudd_bddBooleanDiff(dd, f, 0) != dd.One {
		t.Fatal("boolean difference")
	}
	if Cudd_bddBooleanDiff(dd, f, 9) != dd.LogicZero {
		t.Fatal("absent variable difference")
	}
	independent := Cudd_bddOr(dd, x, y)
	before := len(dd.unique)
	if !Cudd_bddVarIsDependent(dd, f, x) || Cudd_bddVarIsDependent(dd, independent, x) {
		t.Fatal("variable dependency")
	}
	if len(dd.unique) != before {
		t.Fatal("dependency query allocated")
	}
	if Cudd_bddExistAbstract(dd, f, Cudd_Not(x)) != nil || Cudd_bddExistAbstract(dd, f, Cudd_bddOr(dd, x, y)) != nil {
		t.Fatal("nonpositive cube accepted")
	}
	if Cudd_bddXorExistAbstract(dd, x, y, x) != dd.One {
		t.Fatal("xor existential abstraction")
	}
}

func TestReviewPermutationValidationAndConstantReferences(t *testing.T) {
	dd := NewBDDManager(2)
	f := Cudd_bddAnd(dd, dd.Vars[0], dd.Vars[1])
	before := len(dd.Vars)
	if Cudd_bddPermute(dd, f, []int{8, -1}) != nil || len(dd.Vars) != before {
		t.Fatal("invalid permutation mutated manager")
	}
	if Cudd_bddPermute(dd, f, []int{0, int(CUDD_CONST_INDEX)}) != nil {
		t.Fatal("constant index accepted as variable")
	}
	Cudd_Ref(dd.Zero)
	Cudd_RecursiveDeref(dd, dd.Zero)
	if dd.Zero.Ref != 0 {
		t.Fatal("ADD constant dereference ignored")
	}
	if dd.constant(math.Copysign(0, -1)) != dd.Zero {
		t.Fatal("negative zero not canonical")
	}
	if Cudd_ReadNodeCount(dd) != len(dd.unique)+len(dd.zddUnique)+len(dd.constants) {
		t.Fatal("node statistics exclude constants")
	}
}
