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
