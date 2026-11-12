package gia

import (
	"github.com/glycerine/abcgo/misc/vec"
	"testing"
)

func TestSblCutFeasibleAndDominance(t *testing.T) {
	if Sbl_CutIsFeasible(15, 0, 0, 0, 4) != 1 || Sbl_CutIsFeasible(31, 0, 0, 0, 4) != 0 || Sbl_CutIsFeasible(63, 0, 0, 0, 6) != 1 || Sbl_CutIsFeasible(127, 0, 0, 0, 6) != 0 {
		t.Fatal("cut capacity")
	}
	a, b, c, d := vec.Vec_WrdAlloc(2), vec.Vec_WrdAlloc(2), vec.Vec_WrdAlloc(2), vec.Vec_WrdAlloc(2)
	if Sbl_CutPushUncontained(a, b, c, d, 7, 0, 0, 0) != 0 || Sbl_CutPushUncontained(a, b, c, d, 15, 0, 0, 0) != 1 || Sbl_CutPushUncontained(a, b, c, d, 3, 0, 0, 0) != 0 {
		t.Fatal("cut insertion")
	}
	if vec.Vec_WrdSize(a) != 1 || vec.Vec_WrdEntry(a, 0) != 3 {
		t.Fatal("dominated cut retained")
	}
}

func TestSblMergeCuts(t *testing.T) {
	left := [4][]word{{1, 2}, {0, 0}, {0, 0}, {0, 0}}
	right := [4][]word{{4, 6}, {0, 0}, {0, 0}, {0, 0}}
	result := sblMergeCuts(left, right, 4)
	got := vec.Vec_WrdArray(result[0])
	if len(got) != 2 || got[0] != 5 || got[1] != 6 {
		t.Fatalf("minimal merged cuts %v", got)
	}
	for i := 1; i < 4; i++ {
		if vec.Vec_WrdSize(result[i]) != 2 {
			t.Fatal("unaligned cut masks")
		}
	}
}
