package acec

import (
	"github.com/glycerine/abcgo/aig/gia"
	"github.com/glycerine/abcgo/misc/util"
	"testing"
)

func integrationMutatedTree(mutation int) *gia.Gia_Man_t {
	g := gia.Gia_ManStart(128)
	in := make([]int32, 5)
	for i := range in {
		in[i] = gia.Gia_ManAppendCi(g)
	}
	lo, hi := make([]int32, 2), make([]int32, 2)
	Acec_InsertFadd(g, in[:3], lo)
	carry := lo[1]
	if mutation == 0 {
		carry ^= 1
	}
	Acec_InsertFadd(g, []int32{carry, in[3], in[4]}, hi)
	out := []int32{lo[0], hi[0], hi[1]}
	switch mutation {
	case 1:
		out[0], out[1] = out[1], out[0]
	case 2:
		out[2] = lo[1]
	case 3:
		out[1] = lo[0]
	}
	for _, v := range out {
		gia.Gia_ManAppendCo(g, v)
	}
	return g
}
func TestAcecAdderStructuralMutations(t *testing.T) {
	for mutation := 0; mutation < 4; mutation++ {
		a, b := integrationMutatedTree(-1), integrationMutatedTree(mutation)
		differs := false
		for mask := 0; mask < 32; mask++ {
			inputs := make([]bool, 5)
			for i := range inputs {
				inputs[i] = mask&(1<<i) != 0
			}
			x, e1 := a.Evaluate(inputs)
			y, e2 := b.Evaluate(inputs)
			if e1 != nil || e2 != nil {
				t.Fatal(e1, e2)
			}
			for i := range x {
				differs = differs || x[i] != y[i]
			}
		}
		if !differs {
			t.Fatalf("mutation %d has no effect", mutation)
		}
		var opts Acec_ParCec_t
		Acec_ManCecSetDefaultParams(&opts)
		if got := Acec_Solve(a, b, &opts); got != 0 {
			t.Fatalf("mutation %d got %d, want failure", mutation, got)
		}
		cex := a.CombinationalCounterexample()
		if cex == nil {
			t.Fatalf("mutation %d missing witness", mutation)
		}
		inputs := make([]bool, 5)
		for i := range inputs {
			inputs[i] = util.Abc_InfoHasBit(cex.PData, int32(i))
		}
		x, _ := a.Evaluate(inputs)
		y, _ := b.Evaluate(inputs)
		different := false
		for i := range x {
			different = different || x[i] != y[i]
		}
		if !different {
			t.Fatalf("mutation %d invalid witness %v", mutation, inputs)
		}
	}
}
