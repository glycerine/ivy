package io

import (
	"testing"

	"github.com/glycerine/abcgo/base/abc"
)

// ABC's Io_ReadAiger builds its result with Abc_NtkAlloc(ABC_NTK_STRASH, ...),
// which always attaches an AIG manager. Commands such as rewrite rely on it.
func TestReadAigerAttachesAigManager(t *testing.T) {
	data := []byte("aag 3 2 0 1 1\n2\n4\n6\n6 2 4\n")
	n, err := ReadAigerFromMemory(data, true)
	if err != nil {
		t.Fatal(err)
	}
	m, ok := n.FunctionManager.(*abc.Abc_Aig_t)
	if !ok || m == nil {
		t.Fatalf("read AIG has function manager %T, want *abc.Abc_Aig_t", n.FunctionManager)
	}
	if abc.Abc_NtkNodeNum(n) != 1 {
		t.Fatalf("got %d ANDs, want 1", abc.Abc_NtkNodeNum(n))
	}
}

// Io_ReadAiger creates the constant, PIs, POs, then each latch as
// latch/BI/BO, then the ANDs. Object IDs follow from that order and steer
// later traversals, so they must match.
func TestReadAigerObjectIDOrder(t *testing.T) {
	// Two inputs, one latch whose next state is the AND, one output.
	data := []byte("aag 4 2 1 1 1\n2\n4\n6 8\n8\n8 2 4\n")
	n, err := ReadAigerFromMemory(data, true)
	if err != nil {
		t.Fatal(err)
	}
	want := []abc.ObjectKind{abc.ObjectConst1, abc.ObjectPI, abc.ObjectPI, abc.ObjectPO, abc.ObjectLatch, abc.ObjectBI, abc.ObjectBO, abc.ObjectNode}
	if got := abc.Abc_NtkObjNumMax(n); got != int32(len(want)) {
		t.Fatalf("got %d objects, want %d", got, len(want))
	}
	for id, kind := range want {
		if o := n.Object(id); o == nil || o.Kind != kind {
			t.Fatalf("object %d is %v, want kind %v", id, o, kind)
		}
	}
	and := n.Object(7)
	if po := n.Object(3); abc.Abc_ObjFanin0(po) != and {
		t.Fatal("PO does not read the AND")
	}
	if bi := n.Object(5); abc.Abc_ObjFanin0(bi) != and || abc.Abc_ObjFanin0(n.Object(4)) != bi || abc.Abc_ObjFanin0(n.Object(6)) != n.Object(4) {
		t.Fatal("latch boundary is not wired BI -> latch -> BO")
	}
}
