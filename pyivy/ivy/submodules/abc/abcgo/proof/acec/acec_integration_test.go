package acec

import (
	"github.com/glycerine/abcgo/aig/gia"
	"github.com/glycerine/abcgo/misc/util"
	"testing"
)

func integrationAdderTree(swapped bool, invert int) *gia.Gia_Man_t {
	g := gia.Gia_ManStart(128)
	var in [5]int32
	for i := range in {
		in[i] = gia.Gia_ManAppendCi(g)
	}
	first := []int32{in[0], in[1], in[2]}
	if swapped {
		first[0], first[1] = first[1], first[0]
	}
	var lo, hi [2]int32
	Acec_InsertFadd(g, first, lo[:])
	Acec_InsertFadd(g, []int32{lo[1], in[3], in[4]}, hi[:])
	if invert == 0 {
		lo[0] ^= 1
	}
	if invert == 1 {
		hi[0] ^= 1
	}
	if invert == 2 {
		hi[1] ^= 1
	}
	gia.Gia_ManAppendCo(g, lo[0])
	gia.Gia_ManAppendCo(g, hi[0])
	gia.Gia_ManAppendCo(g, hi[1])
	return g
}
func TestAcecSolveAdderTreesAndReplay(t *testing.T) {
	for _, changedOutput := range []int{-1, 0, 1, 2} {
		changed := changedOutput >= 0
		left, right := integrationAdderTree(false, -1), integrationAdderTree(true, changedOutput)
		for mask := 0; mask < 32; mask++ {
			inputs := make([]bool, 5)
			for i := range inputs {
				inputs[i] = mask&(1<<i) != 0
			}
			a, ea := left.Evaluate(inputs)
			b, eb := right.Evaluate(inputs)
			if ea != nil || eb != nil {
				t.Fatal(ea, eb)
			}
			diff := false
			for i := range a {
				diff = diff || a[i] != b[i]
			}
			if diff != changed {
				t.Fatalf("fixture mask %d diff=%v", mask, diff)
			}
		}
		var opts Acec_ParCec_t
		Acec_ManCecSetDefaultParams(&opts)
		got := Acec_Solve(left, right, &opts)
		want := int32(1)
		if changed {
			want = 0
		}
		if got != want {
			t.Fatalf("changed=%v result=%d want=%d", changed, got, want)
		}
		if changed {
			cex := left.CombinationalCounterexample()
			if cex == nil {
				t.Fatal("missing counterexample")
			}
			inputs := make([]bool, 5)
			for i := range inputs {
				inputs[i] = util.Abc_InfoHasBit(cex.PData, int32(i))
			}
			a, _ := left.Evaluate(inputs)
			b, _ := right.Evaluate(inputs)
			diff := false
			for i := range a {
				diff = diff || a[i] != b[i]
			}
			if !diff {
				t.Fatalf("witness %v does not distinguish %v and %v", inputs, a, b)
			}
		}
	}
}
