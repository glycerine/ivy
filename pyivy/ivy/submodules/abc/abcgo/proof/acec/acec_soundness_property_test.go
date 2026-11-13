package acec

import (
	"github.com/glycerine/abcgo/aig/gia"
	"github.com/glycerine/abcgo/misc/util"
	"testing"
)

func acecSignedChain(flags int) *Gia_Man_t {
	p := gia.Gia_ManStart(128)
	var in [5]int32
	for i := range in {
		in[i] = gia.Gia_ManAppendCi(p) ^ int32(flags>>i&1)
	}
	var lo, hi [2]int32
	Acec_InsertFadd(p, in[:3], lo[:])
	Acec_InsertFadd(p, []int32{lo[1], in[3], in[4]}, hi[:])
	for i, out := range []int32{lo[0], hi[0], hi[1]} {
		gia.Gia_ManAppendCo(p, out^int32(flags>>(5+i)&1))
	}
	return p
}
func TestMatchedAdderCecAgainstExhaustiveOracle(t *testing.T) {
	for flags := 0; flags < 64; flags++ {
		a, b := acecSignedChain(0), acecSignedChain(flags)
		differs := false
		for mask := 0; mask < 32; mask++ {
			in := []bool{mask&1 != 0, mask&2 != 0, mask&4 != 0, mask&8 != 0, mask&16 != 0}
			x, e1 := a.Evaluate(in)
			y, e2 := b.Evaluate(in)
			if e1 != nil || e2 != nil {
				t.Fatal(e1, e2)
			}
			for i := range x {
				differs = differs || x[i] != y[i]
			}
		}
		var pars Acec_ParCec_t
		Acec_ManCecSetDefaultParams(&pars)
		got := Acec_Solve(a, b, &pars)
		want := int32(1)
		if differs {
			want = 0
		}
		if got != want {
			t.Fatalf("flags %d: status %d want %d", flags, got, want)
		}
		if differs {
			cex := a.CombinationalCounterexample()
			if cex == nil {
				t.Fatalf("flags %d: missing witness", flags)
			}
			in := make([]bool, 5)
			for i := range in {
				in[i] = util.Abc_InfoHasBit(cex.PData, int32(i))
			}
			x, _ := a.Evaluate(in)
			y, _ := b.Evaluate(in)
			mismatch := false
			for i := range x {
				mismatch = mismatch || x[i] != y[i]
			}
			if !mismatch {
				t.Fatalf("flags %d: invalid witness %v", flags, in)
			}
		}
	}
}
