package gia

import (
	"github.com/glycerine/abcgo/misc/vec"
	"testing"
)

func TestSleCutMergeAndContainment(t *testing.T) {
	a := []int32{Sle_CutSetSizeSign(2, (1<<1)|(1<<3)), 1, 3}
	b := []int32{Sle_CutSetSizeSign(2, (1<<3)|(1<<5)), 3, 5}
	out := make([]int32, 5)
	if Sle_CutMergeOrder(a, b, out, 3) != 1 || Sle_CutSize(out) != 3 || out[1] != 1 || out[2] != 3 || out[3] != 5 {
		t.Fatalf("merge %v", out)
	}
	if Sle_CutMergeOrder(a, b, out, 2) != 0 {
		t.Fatal("oversized cut accepted")
	}
	if Sle_SetCutIsContainedOrder(out, a) != 1 || Sle_SetCutIsContainedOrder(a, b) != 0 {
		t.Fatal("bad cut containment")
	}
	set := vec.Vec_IntStart(1)
	Sle_SetAddCut(set, out)
	if Sle_SetLastCutIsContained(set, a) != 0 {
		t.Fatal("superset contains candidate")
	}
	Sle_SetAddCut(set, a)
	if Sle_CutIsUsed(vec.Vec_IntArray(set)[1:]) != 0 || Sle_SetLastCutIsContained(set, out) != 1 {
		t.Fatal("dominance filtering failed")
	}
}
