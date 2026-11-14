package abci

import (
	"github.com/glycerine/abcgo/base/abc"
	"github.com/glycerine/abcgo/map/mio"
	"testing"
)

func TestPhaseMapPreservesFunctionAndImprovesArea(t *testing.T) {
	library, err := mio.ReadLibrary("phase", `GATE zero 0 O=CONST0;
GATE one 0 O=CONST1;
GATE buf 1 O=a; PIN * NONINV 1 999 1 0 1 0
GATE inv 1 O=!a; PIN * INV 1 999 1 0 1 0
GATE and2 3 O=a*b; PIN * NONINV 1 999 1 0 1 0
GATE nand2 1 O=!(a*b); PIN * INV 1 999 1 0 1 0
GATE or2 3 O=a+b; PIN * NONINV 1 999 1 0 1 0
GATE nor2 1 O=!(a+b); PIN * INV 1 999 1 0 1 0
`, nil)
	if err != nil {
		t.Fatal(err)
	}
	n := abc.Abc_NtkAlloc(abc.ABC_NTK_LOGIC, abc.ABC_FUNC_MAP, 1)
	n.FunctionManager = library
	a, b := abc.Abc_NtkCreatePi(n), abc.Abc_NtkCreatePi(n)
	node := abc.Abc_NtkCreateNode(n)
	node.Data = mio.Mio_LibraryReadGateByName(library, "and2", "")
	abc.Abc_ObjAddFanin(node, a)
	abc.Abc_ObjAddFanin(node, b)
	abc.Abc_ObjAddFanin(abc.Abc_NtkCreatePo(n), node)
	abc.Abc_NtkAddDummyPiNames(n)
	abc.Abc_NtkAddDummyPoNames(n)
	if err := abcPhaseMap(n, 0); err != nil {
		t.Fatal(err)
	}
	var area float64
	for _, node := range n.ObjectsOfKind(abc.ObjectNode) {
		gate, ok := node.Data.(*mio.Mio_Gate_t)
		if !ok {
			t.Fatalf("lost mapped gate %T", node.Data)
		}
		area += mio.Mio_GateReadArea(gate)
	}
	if area >= 3 {
		t.Fatalf("phase mapping failed to reduce area: %g", area)
	}
	for bits := 0; bits < 4; bits++ {
		if got := evaluateNetwork(t, n, bits)[0]; got != (bits == 3) {
			t.Fatalf("bits=%d got=%v", bits, got)
		}
	}
	abc.Abc_NtkDelete(n)
	plain := truthNetwork(t, 1, 2)
	defer abc.Abc_NtkDelete(plain)
	if abcPhaseMap(plain, 0) == nil {
		t.Fatal("missing library accepted")
	}
}
