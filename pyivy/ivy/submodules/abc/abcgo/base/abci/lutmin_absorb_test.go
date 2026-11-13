package abci

import (
	"reflect"
	"testing"

	"github.com/glycerine/abcgo/base/abc"
	"github.com/glycerine/abcgo/misc/vec"
)

func TestObjCheckAbsorbSharedFaninAndLimit(t *testing.T) {
	n := abc.NewNetwork("absorb")
	a, b, c := abc.Abc_NtkCreatePi(n), abc.Abc_NtkCreatePi(n), abc.Abc_NtkCreatePi(n)
	pivot, obj := abc.Abc_NtkCreateNode(n), abc.Abc_NtkCreateNode(n)
	abc.Abc_ObjAddFanin(pivot, a)
	abc.Abc_ObjAddFanin(pivot, b)
	abc.Abc_ObjAddFanin(obj, pivot)
	abc.Abc_ObjAddFanin(obj, b)
	abc.Abc_ObjAddFanin(obj, c)
	fanins := vec.Vec_PtrAlloc(4)
	vec.Vec_PtrPush(fanins, pivot)
	if got := Abc_ObjCheckAbsorb(obj, pivot, 3, fanins); got != 1 {
		t.Fatalf("absorb returned %d, want 1", got)
	}
	if got, want := vec.Vec_PtrArray(fanins), []any{b, c, a}; !reflect.DeepEqual(got, want) {
		t.Fatalf("fanins = %v, want %v", got, want)
	}
	if got := Abc_ObjCheckAbsorb(obj, pivot, 2, fanins); got != 0 {
		t.Fatalf("limited absorb returned %d, want 0", got)
	}
	if got, want := vec.Vec_PtrArray(fanins), []any{b, c, a}; !reflect.DeepEqual(got, want) {
		t.Fatalf("partial fanins = %v, want %v", got, want)
	}
}

func TestObjCheckAbsorbEmptyPivot(t *testing.T) {
	n := abc.NewNetwork("empty")
	pivot, obj := abc.Abc_NtkCreateNode(n), abc.Abc_NtkCreateNode(n)
	a := abc.Abc_NtkCreatePi(n)
	abc.Abc_ObjAddFanin(obj, pivot)
	abc.Abc_ObjAddFanin(obj, a)
	fanins := vec.Vec_PtrAlloc(1)
	// C checks the bound inside the pivot-fanin loop, which is empty here.
	if got := Abc_ObjCheckAbsorb(obj, pivot, 0, fanins); got != 1 {
		t.Fatalf("empty pivot returned %d, want 1", got)
	}
	if got, want := vec.Vec_PtrArray(fanins), []any{a}; !reflect.DeepEqual(got, want) {
		t.Fatalf("fanins = %v, want %v", got, want)
	}
}

func TestNetworkAbsorbCountsAllFanoutsAndDanglingNodes(t *testing.T) {
	n := abc.NewNetwork("counts")
	a, b := abc.Abc_NtkCreatePi(n), abc.Abc_NtkCreatePi(n)
	pivot := abc.Abc_NtkCreateNode(n)
	abc.Abc_ObjAddFanin(pivot, a)
	abc.Abc_ObjAddFanin(pivot, b)
	x, y := abc.Abc_NtkCreateNode(n), abc.Abc_NtkCreateNode(n)
	abc.Abc_ObjAddFanin(x, pivot)
	abc.Abc_ObjAddFanin(y, pivot)
	out := abc.Abc_NtkCreatePo(n)
	abc.Abc_ObjAddFanin(out, x)
	for _, tc := range []struct{ size, absorbed, fully int }{{2, 2, 2}, {1, 0, 1}} {
		absorbed, fully := lutAbsorbCounts(n, tc.size)
		if absorbed != tc.absorbed || fully != tc.fully {
			t.Fatalf("limit=%d: absorbed=%d fully=%d, want %d,%d", tc.size, absorbed, fully, tc.absorbed, tc.fully)
		}
	}
}
