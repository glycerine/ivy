package acec

import (
	"github.com/glycerine/abcgo/aig/gia"
	"github.com/glycerine/abcgo/misc/util"
	"testing"
)

func integrationPerturbedAdder(invert int) *gia.Gia_Man_t {
	g := gia.Gia_ManStart(128)
	inputs := make([]int32, 5)
	for i := range inputs {
		inputs[i] = gia.Gia_ManAppendCi(g)
		if i == invert {
			inputs[i] ^= 1
		}
	}
	lo, hi := make([]int32, 2), make([]int32, 2)
	Acec_InsertFadd(g, inputs[:3], lo)
	Acec_InsertFadd(g, []int32{lo[1], inputs[3], inputs[4]}, hi)
	gia.Gia_ManAppendCo(g, lo[0])
	gia.Gia_ManAppendCo(g, hi[0])
	gia.Gia_ManAppendCo(g, hi[1])
	return g
}
func TestAcecInternalPolarityPerturbations(t *testing.T) {
	for invert := 0; invert < 5; invert++ {
		a, b := integrationPerturbedAdder(-1), integrationPerturbedAdder(invert)
		var opts Acec_ParCec_t
		Acec_ManCecSetDefaultParams(&opts)
		got := Acec_Solve(a, b, &opts)
		if got != 0 {
			t.Fatalf("input %d inverted: got status %d want failure", invert, got)
		}
		cex := a.CombinationalCounterexample()
		if cex == nil {
			t.Fatalf("input %d inverted: missing witness", invert)
		}
		inputs := make([]bool, 5)
		for i := range inputs {
			inputs[i] = util.Abc_InfoHasBit(cex.PData, int32(i))
		}
		av, ea := a.Evaluate(inputs)
		bv, eb := b.Evaluate(inputs)
		if ea != nil || eb != nil {
			t.Fatal(ea, eb)
		}
		different := false
		for i := range av {
			different = different || av[i] != bv[i]
		}
		if !different {
			t.Fatalf("input %d inverted: invalid witness %v", invert, inputs)
		}
	}
}
