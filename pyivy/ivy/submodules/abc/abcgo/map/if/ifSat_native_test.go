package if_

import (
	"testing"

	"github.com/glycerine/abcgo/opt/dau"
)

func TestIfSatCheckXYDerivesExactLutTruthTables(t *testing.T) {
	truth := dau.Dau_DsdToTruth("((ab)c)", 3)
	set := uint32(1<<0 | 1<<2)
	var bound, free word
	if If_ManSatCheckXY(If_ManSatBuildXY(2), 2, &truth[0], 3, set, &bound, &free, nil) == 0 {
		t.Fatal("expected factorization for nested AND")
	}
	var bs, ss, fs []int32
	for v := int32(0); v < 3; v++ {
		switch set >> (2 * v) & 3 {
		case 0:
			fs = append(fs, v)
		case 1:
			bs = append(bs, v)
		case 3:
			ss = append(ss, v)
		}
	}
	got := If_ManSat6Truth(bound, free, bs, ss, fs)
	if got != truth[0] {
		t.Fatalf("reconstructed truth %#x, want %#x", got, truth[0])
	}
}

func TestIfSatCheckXYRejectsThreeFreeCofactors(t *testing.T) {
	var truth word
	for assignment := 0; assignment < 8; assignment++ {
		bound, free := assignment&3, assignment>>2&1
		value := bound == 3 || bound == 1 && free == 1 || bound == 2 && free == 0
		if value {
			truth |= 1 << uint(assignment)
		}
	}
	set := uint32(1<<0 | 1<<2)
	if If_ManSatCheckXY(If_ManSatBuildXY(2), 2, &truth, 3, set, nil, nil, nil) != 0 {
		t.Fatal("expected prime function to reject one-bit intermediate")
	}
}

func TestIfSatCheckXYAllReturnsAValidPartition(t *testing.T) {
	truth := dau.Dau_DsdToTruth("((ab)c)", 3)
	set := If_ManSatCheckXYall(If_ManSatBuildXY(2), 2, &truth[0], 3, nil)
	if set == 0 {
		t.Fatal("expected a valid LUT decomposition")
	}
	if If_ManSatCheckXY(nil, 2, &truth[0], 3, set, nil, nil, nil) == 0 {
		t.Fatalf("returned invalid partition %#x", set)
	}
}
