package int2

import (
	"github.com/glycerine/abcgo/aig/gia"
	"github.com/glycerine/abcgo/misc/vec"
	"testing"
)

func TestRefineCubeSourceRejectsRegisterCI(t *testing.T) {
	g := gia.Gia_ManStart(8)
	state := gia.Gia_ManAppendCi(g)
	gia.Gia_ManAppendCo(g, state)
	gia.Gia_ManAppendCo(g, state)
	gia.Gia_ManSetRegNum(g, 1)
	assignment := vec.Vec_IntAllocArrayCopy([]int32{1}, 1)
	priority := vec.Vec_IntAllocArrayCopy([]int32{1}, 1)
	defer func() {
		if recover() == nil {
			t.Fatal("C source rejects the register CI as not an AND")
		}
	}()
	Int2_ManRefineCube(g, assignment, priority)
}

func TestPreimageReportsUnsupportedForPrimaryInput(t *testing.T) {
	g := gia.Gia_ManStart(8)
	x, s := gia.Gia_ManAppendCi(g), gia.Gia_ManAppendCi(g)
	gia.Gia_ManAppendCo(g, gia.Gia_ManAppendAnd(g, x, s))
	prefix, suffix := sat_solver_new(), sat_solver_new()
	ci := vec.Vec_IntAllocArrayCopy([]int32{0}, 1)
	co := vec.Vec_IntAllocArrayCopy([]int32{0, 1}, 2)
	prio := vec.Vec_IntAllocArrayCopy([]int32{1, 1}, 2)
	image, status := Int2_ManComputePreimage(g, prefix, suffix, ci, co, prio)
	if image != nil || status != INT2_TIME_OUT {
		t.Fatalf("image %v status %d", image, status)
	}
}

func TestSourceJustificationUsesPriorityAsBoolean(t *testing.T) {
	g := gia.Gia_ManStart(8)
	x, y, state := gia.Gia_ManAppendCi(g), gia.Gia_ManAppendCi(g), gia.Gia_ManAppendCi(g)
	gate := gia.Gia_ManAppendAnd(g, x, y)
	gia.Gia_ManAppendCo(g, gate)
	gia.Gia_ManAppendCo(g, state)
	gia.Gia_ManSetRegNum(g, 1)
	a, b, node := gia.Gia_ManObj(g, x>>1), gia.Gia_ManObj(g, y>>1), gia.Gia_ManObj(g, gate>>1)
	gia.Gia_ObjSetMark0(a, 0)
	gia.Gia_ObjSetMark0(b, 0)
	gia.Gia_ObjSetMark0(node, 0)
	gia.Gia_ObjSetValue(a, 2)
	gia.Gia_ObjSetValue(b, 3)
	gia.Gia_ObjSetValue(node, 1)
	selected := vec.Vec_IntAlloc(0)
	Int2_ManJustify_rec(g, node, selected)
	if gia.Gia_ObjMark1(a) != 1 || gia.Gia_ObjMark1(b) != 1 {
		t.Fatal("source Value==1 branch must visit both inputs despite false mark")
	}
}

func TestSourceJustificationChoosesByMarks(t *testing.T) {
	g := gia.Gia_ManStart(8)
	x, y, state := gia.Gia_ManAppendCi(g), gia.Gia_ManAppendCi(g), gia.Gia_ManAppendCi(g)
	gate := gia.Gia_ManAppendAnd(g, x, y^1)
	gia.Gia_ManAppendCo(g, gate)
	gia.Gia_ManAppendCo(g, state)
	gia.Gia_ManSetRegNum(g, 1)
	a, b, node := gia.Gia_ManObj(g, x>>1), gia.Gia_ManObj(g, y>>1), gia.Gia_ManObj(g, gate>>1)
	gia.Gia_ObjSetMark0(a, 0)
	gia.Gia_ObjSetMark0(b, 1)
	gia.Gia_ObjSetMark0(node, 0)
	gia.Gia_ObjSetValue(a, 9)
	gia.Gia_ObjSetValue(b, 2)
	gia.Gia_ObjSetValue(node, 9)
	Int2_ManJustify_rec(g, node, vec.Vec_IntAlloc(0))
	if gia.Gia_ObjMark1(a) != 1 || gia.Gia_ObjMark1(b) != 0 {
		t.Fatal("source must visit lower-mark input despite its higher priority")
	}
}
