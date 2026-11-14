package abc

import "testing"

// The fixture in test_vectors/sweep_oracle.c produces
// before=3 removed=2 after=1 fanins=2 first=1 second=2 with ABC_USE_CUDD.
func TestSourceSweepInverterAndDangling(t *testing.T) {
	n := Abc_NtkAlloc(ABC_NTK_LOGIC, ABC_FUNC_SOP, 1)
	a, b := Abc_NtkCreatePi(n), Abc_NtkCreatePi(n)
	inv := Abc_NtkCreateNodeInv(n, a)
	root := Abc_NtkCreateNodeAnd(n, []*Object{inv, b})
	po := Abc_NtkCreatePo(n)
	Abc_ObjAddFanin(po, root)
	Abc_NtkCreateNodeBuf(n, a)
	before := Abc_NtkNodeNum(n)
	removed := Abc_NtkSweep(n, 0)
	if before != 3 || removed != 2 || Abc_NtkNodeNum(n) != 1 {
		t.Fatalf("before=%d removed=%d after=%d", before, removed, Abc_NtkNodeNum(n))
	}
	root = Abc_ObjFanin0(po)
	if Abc_ObjFaninNum(root) != 2 || Abc_ObjFanin0(root) != a || Abc_ObjFanin1(root) != b {
		t.Fatalf("wrong sweep fanins")
	}
	fn := localFunction(root)
	for bits := 0; bits < 4; bits++ {
		x, y := bits&1 != 0, bits&2 != 0
		if got := fn([]bool{x, y}); got != (!x && y) {
			t.Fatalf("a=%v b=%v got=%v", x, y, got)
		}
	}
}

func TestSourceSweepPushesInverterIntoSingleFanoutDriver(t *testing.T) {
	n := Abc_NtkAlloc(ABC_NTK_LOGIC, ABC_FUNC_SOP, 1)
	a, b := Abc_NtkCreatePi(n), Abc_NtkCreatePi(n)
	and := Abc_NtkCreateNodeAnd(n, []*Object{a, b})
	inv := Abc_NtkCreateNodeInv(n, and)
	po := Abc_NtkCreatePo(n)
	Abc_ObjAddFanin(po, inv)
	if removed := Abc_NtkSweep(n, 0); removed != 1 {
		t.Fatalf("removed %d want 1", removed)
	}
	if root := Abc_ObjFanin0(po); root != and {
		t.Fatal("inverter was not pushed into driver")
	}
	fn := localFunction(and)
	for bits := 0; bits < 4; bits++ {
		x, y := bits&1 != 0, bits&2 != 0
		if got := fn([]bool{x, y}); got != !(x && y) {
			t.Fatalf("a=%v b=%v got=%v", x, y, got)
		}
	}
}

func TestSourceSweepConstantInput(t *testing.T) {
	n := Abc_NtkAlloc(ABC_NTK_LOGIC, ABC_FUNC_SOP, 1)
	a := Abc_NtkCreatePi(n)
	zero := Abc_NtkCreateNodeConst0(n)
	root := Abc_NtkCreateNodeAnd(n, []*Object{zero, a})
	po := Abc_NtkCreatePo(n)
	Abc_ObjAddFanin(po, root)
	if removed := Abc_NtkSweep(n, 0); removed != 1 {
		t.Fatalf("removed %d want 1", removed)
	}
	out := Abc_ObjFanin0(po)
	if Abc_ObjFaninNum(out) != 0 || Abc_NodeIsConst0(out) == 0 {
		t.Fatal("zero input was not propagated")
	}
}
