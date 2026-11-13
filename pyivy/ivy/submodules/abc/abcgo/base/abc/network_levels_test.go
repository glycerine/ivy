package abc

import (
	"testing"

	"github.com/glycerine/abcgo/misc/vec"
)

func TestReverseLevelsAndRequiredLevels(t *testing.T) {
	n := NewNetwork("levels")
	pi := Abc_NtkCreatePi(n)
	first := Abc_NtkCreateNode(n)
	second := Abc_NtkCreateNode(n)
	po := Abc_NtkCreatePo(n)
	Abc_ObjAddFanin(first, pi)
	Abc_ObjAddFanin(second, first)
	Abc_ObjAddFanin(po, second)

	Abc_NtkStartReverseLevels(n, 3)
	if got := Abc_ObjReverseLevel(second); got != 1 {
		t.Fatalf("second reverse level = %d, want 1", got)
	}
	if got := Abc_ObjReverseLevel(first); got != 2 {
		t.Fatalf("first reverse level = %d, want 2", got)
	}
	if got := Abc_ObjRequiredLevel(first); got != 4 {
		t.Fatalf("first required level = %d, want 4", got)
	}
	if got := Abc_ObjRequiredLevel(second); got != 5 {
		t.Fatalf("second required level = %d, want 5", got)
	}
	Abc_NtkStopReverseLevels(n)
	func() {
		defer func() {
			if recover() == nil {
				t.Fatal("reverse-level read after stop did not panic")
			}
		}()
		Abc_ObjReverseLevel(first)
	}()
}

func TestNetworkUpdateMaintainsForwardAndReverseLevels(t *testing.T) {
	n := NewNetwork("update-levels")
	pi0 := Abc_NtkCreatePi(n)
	pi1 := Abc_NtkCreatePi(n)
	pi2 := Abc_NtkCreatePi(n)
	first := Abc_NtkCreateNode(n)
	old := Abc_NtkCreateNode(n)
	downstream := Abc_NtkCreateNode(n)
	po := Abc_NtkCreatePo(n)
	Abc_ObjAddFanin(first, pi0)
	Abc_ObjAddFanin(old, pi1)
	Abc_ObjAddFanin(downstream, old)
	Abc_ObjAddFanin(downstream, pi2)
	Abc_ObjAddFanin(po, downstream)
	replacement := Abc_NtkCreateNode(n)
	Abc_ObjAddFanin(replacement, first)
	Abc_ObjAddFanin(replacement, pi2)

	Abc_NtkStartReverseLevels(n, 0)
	levels := vec.Vec_VecStart(4)
	Abc_NtkUpdate(old, replacement, levels)
	if got := Abc_ObjLevel(replacement); got != 2 {
		t.Fatalf("replacement level = %d, want 2", got)
	}
	if got := Abc_ObjLevel(downstream); got != 3 {
		t.Fatalf("downstream level = %d, want 3", got)
	}
	if got := Abc_ObjReverseLevel(replacement); got != 2 {
		t.Fatalf("replacement reverse level = %d, want 2", got)
	}
	if got := Abc_ObjReverseLevel(first); got != 3 {
		t.Fatalf("first reverse level = %d, want 3", got)
	}
	Abc_NtkStopReverseLevels(n)
	vec.Vec_VecFree(levels)
}
