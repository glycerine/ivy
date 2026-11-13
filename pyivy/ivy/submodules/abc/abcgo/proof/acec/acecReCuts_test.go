package acec

import (
	"github.com/glycerine/abcgo/misc/vec"
	"testing"
)

func TestThreeLeafCutMergeAndLookup(t *testing.T) {
	a, b := []int32{2, 1, 3, 0}, []int32{2, 2, 3, 0}
	out := make([]int32, 5)
	if Ree_ManCutMergeOne(a, b, out) != 1 || out[0] != 3 || out[1] != 1 || out[2] != 2 || out[3] != 3 {
		t.Fatalf("merge %v", out)
	}
	if Ree_ManCutMergeOne([]int32{2, 4, 5, 0}, b, out) != 0 {
		t.Fatal("accepted four distinct leaves")
	}
	cuts := vec.Vec_IntAlloc(6)
	for _, x := range []int32{1, 3, 1, 2, 3, 0} {
		vec.Vec_IntPush(cuts, x)
	}
	if Ree_ManCutCheckEqual(cuts, []int32{3, 1, 2, 3, 0}) != 1 || Ree_ManCutCheckEqual(cuts, []int32{2, 1, 2, 0}) != 0 {
		t.Fatal("bad cut equality")
	}
}
func TestCutTruthVariableExpansion(t *testing.T) {
	target := []int32{3, 3, 5, 7, 0}
	for i, id := range []int32{3, 5, 7} {
		base := []int32{1, id, 0x55}
		want := []int32{0x55, 0x33, 0x0f}[i]
		if got := Ree_ManCutTruthOne(base, target); got != want {
			t.Fatalf("var %d = %x", i, got)
		}
		base[2] = 0xaa
		if got := Ree_ManCutTruthOne(base, target); got != (^want & 0xff) {
			t.Fatalf("inverted var %d = %x", i, got)
		}
	}
}
