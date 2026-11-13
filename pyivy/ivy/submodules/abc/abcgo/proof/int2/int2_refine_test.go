package int2

import (
	"github.com/glycerine/abcgo/aig/gia"
	"github.com/glycerine/abcgo/misc/util"
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

func TestSourceJustificationSingleControllingFanin(t *testing.T) {
	for _, tc := range []struct {
		name    string
		v0, v1  int32
		skipped bool
	}{
		{"first", 0, 1, false},
		{"second", 1, 0, false},
		{"infinite priority", 0, 1, true},
	} {
		t.Run(tc.name, func(t *testing.T) {
			g := gia.Gia_ManStart(8)
			a, b, state := gia.Gia_ManAppendCi(g), gia.Gia_ManAppendCi(g), gia.Gia_ManAppendCi(g)
			gate := gia.Gia_ManAppendAnd(g, a^1, b)
			gia.Gia_ManAppendCo(g, gate)
			gia.Gia_ManAppendCo(g, state)
			gia.Gia_ManSetRegNum(g, 1)
			obj := gia.Gia_ManObj(g, gate>>1)
			f0, f1 := gia.Gia_ObjFanin0(obj), gia.Gia_ObjFanin1(obj)
			gia.Gia_ObjSetMark0(f0, tc.v0^gia.Gia_ObjFaninC0(obj))
			gia.Gia_ObjSetMark0(f1, tc.v1^gia.Gia_ObjFaninC1(obj))
			gia.Gia_ObjSetValue(f0, 2)
			gia.Gia_ObjSetValue(f1, 3)
			gia.Gia_ObjSetValue(obj, 4)
			if tc.skipped {
				gia.Gia_ObjSetValue(f0, util.ABC_INFINITY)
			}
			Int2_ManJustify_rec(g, obj, vec.Vec_IntAlloc(0))
			want0, want1 := tc.v0 == 0 && !tc.skipped, tc.v1 == 0
			if (gia.Gia_ObjMark1(f0) != 0) != want0 || (gia.Gia_ObjMark1(f1) != 0) != want1 {
				t.Fatalf("marks %d,%d want %v,%v", gia.Gia_ObjMark1(f0), gia.Gia_ObjMark1(f1), want0, want1)
			}
		})
	}
}

func TestSourceJustificationOutputMarkPreventsDuplicateSelection(t *testing.T) {
	g := gia.Gia_ManStart(8)
	a, state := gia.Gia_ManAppendCi(g), gia.Gia_ManAppendCi(g)
	gia.Gia_ManAppendCo(g, a)
	gia.Gia_ManAppendCo(g, state)
	gia.Gia_ManSetRegNum(g, 1)
	co := gia.Gia_ManCo(g, 1)
	selected := vec.Vec_IntAlloc(0)
	Int2_ManJustify_rec(g, co, selected)
	Int2_ManJustify_rec(g, co, selected)
	if vec.Vec_IntSize(selected) != 1 || vec.Vec_IntEntry(selected, 0) != gia.Gia_ObjCioId(co) {
		t.Fatalf("selected %v", vec.Vec_IntArray(selected))
	}
}
