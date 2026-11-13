package bmc

import (
	"github.com/glycerine/abcgo/misc/vec"
	"testing"
)

func TestDivisorTableRehash(t *testing.T) {
	table := Tab_TabAlloc(4)
	for i := int32(0); i < 40; i++ {
		if Tab_TabHashAdd(table, []int32{i, i + 1, TAB_UNUSED}, 1, 1) != 0 {
			t.Fatal("duplicate")
		}
	}
	for i := int32(0); i < 40; i++ {
		if Tab_TabHashAdd(table, []int32{i, i + 1, TAB_UNUSED}, 1, i+1) != 1 {
			t.Fatal("missing entry", i)
		}
	}
	best := vec.Vec_IntArray(Tab_TabFindBest(table, 2))
	if len(best) != 8 || best[1] != 39 || best[5] != 38 {
		t.Fatalf("best %v", best)
	}
}
