package wln

import (
	"github.com/glycerine/abcgo/misc/vec"
	"testing"
	"unsafe"
)

// wlnObj.c stores CI/CO indices in the second inline slot, then grows the
// object/fanin arrays together when the original capacity is exhausted.
func TestObjectAllocationSourceIndicesAndGrowth(t *testing.T) {
	p := Wln_NtkAlloc("", 1)
	a := Wln_ObjAlloc(p, ABC_OPER_CI, 0, 7, 0)
	b := Wln_ObjAlloc(p, ABC_OPER_CI, 1, 3, 0)
	co := Wln_ObjAlloc(p, ABC_OPER_CO, 0, 7, 0)
	Wln_ObjAddFanin(p, co, a)
	if a != 1 || b != 2 || co != 3 || Wln_ObjCioId(p, a) != 0 || Wln_ObjCioId(p, b) != 1 || Wln_ObjCioId(p, co) != 0 {
		t.Fatalf("objects %d %d %d", a, b, co)
	}
	if Wln_ObjFanin0(p, co) != a || Wln_ObjRange(p, a) != 8 || Wln_ObjIsSigned(p, b) != 1 {
		t.Fatal("object metadata changed")
	}
	if Wln_NtkCiNum(p) != 2 || Wln_NtkCoNum(p) != 1 {
		t.Fatal("interface counters")
	}
}

func TestStaticFanoutSliceWritesBackingVector(t *testing.T) {
	p := &Wln_Ntk_t{}
	vec.Vec_IntFill(&p.vFanout, 5, 0)
	vec.Vec_IntWriteEntry(&p.vFanout, 1, 3)
	Wln_ObjSetFanout(p, 1, 1, 17)
	if Wln_ObjFanout(p, 1, 1) != 17 || vec.Vec_IntEntry(&p.vFanout, 4) != 17 {
		t.Fatal("fanout write did not update storage")
	}
}

func TestNetworkSourceFanoutAndAcyclic(t *testing.T) {
	p := Wln_NtkAlloc("", 8)
	a := Wln_ObjAlloc(p, ABC_OPER_CI, 0, 0, 0)
	node := Wln_ObjAlloc(p, ABC_OPER_NONE, 0, 0, 0)
	Wln_ObjAddFanin(p, node, a)
	co := Wln_ObjAlloc(p, ABC_OPER_CO, 0, 0, 0)
	Wln_ObjAddFanin(p, co, node)
	if Wln_NtkIsAcyclic(p) != 1 {
		t.Fatal("acyclic network rejected")
	}
	Wln_NtkStaticFanoutStart(p)
	if Wln_ObjFanoutNum(p, a) != 1 || Wln_ObjFanout(p, a, 0) != node || Wln_ObjFanoutNum(p, node) != 1 || Wln_ObjFanout(p, node, 0) != co {
		t.Fatal("wrong static fanouts")
	}
	Wln_NtkStaticFanoutStop(p)
	Wln_ObjSetFanin(p, node, 0, node)
	if Wln_NtkIsAcyclic(p) != 0 {
		t.Fatal("self-cycle not found")
	}
}

func TestNetworkSourceDuplicateDFS(t *testing.T) {
	p := Wln_NtkAlloc("", 8)
	a := Wln_ObjAlloc(p, ABC_OPER_CI, 0, 7, 0)
	n := Wln_ObjAlloc(p, ABC_OPER_NONE, 0, 7, 0)
	Wln_ObjAddFanin(p, n, a)
	co := Wln_ObjAlloc(p, ABC_OPER_CO, 0, 7, 0)
	Wln_ObjAddFanin(p, co, n)
	dup := Wln_NtkDupDfs(p)
	if Wln_NtkCiNum(dup) != 1 || Wln_NtkCoNum(dup) != 1 || Wln_NtkObjNum(dup) != 4 {
		t.Fatal("duplicate interface")
	}
	if Wln_ObjFanin0(dup, Wln_NtkCo(dup, 0)) != n || Wln_ObjRange(dup, n) != 8 {
		t.Fatal("duplicate cone")
	}
}

func TestNetworkSourceMemoryEstimate64Bit(t *testing.T) {
	if unsafe.Sizeof(uintptr(0)) != 8 {
		t.Skip("ABC fixture uses 64-bit ABI")
	}
	p := Wln_NtkAlloc("", 8)
	Wln_ObjAlloc(p, ABC_OPER_CI, 0, 7, 0)
	Wln_ObjAlloc(p, ABC_OPER_NONE, 0, 7, 0)
	Wln_ObjAlloc(p, ABC_OPER_CO, 0, 7, 0)
	if got := unsafe.Sizeof(cWlnNtk{}); got != 1112 {
		t.Fatalf("ABI layout %d", got)
	}
	if got := Wln_NtkMemUsage(p); got != 5049 {
		t.Fatalf("source memory estimate %d, want 5049", got)
	}
}
