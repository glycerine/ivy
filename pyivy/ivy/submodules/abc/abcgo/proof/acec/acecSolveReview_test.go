package acec

import (
	"github.com/glycerine/abcgo/aig/gia"
	"testing"
)

func reviewAdderChain(invert int) *Gia_Man_t {
	p := gia.Gia_ManStart(128)
	var in [5]int32
	for i := range in {
		in[i] = gia.Gia_ManAppendCi(p)
	}
	var lo, hi [2]int32
	Acec_InsertFadd(p, in[:3], lo[:])
	Acec_InsertFadd(p, []int32{lo[1], in[3], in[4]}, hi[:])
	outputs := [3]int32{lo[0], hi[0], hi[1]}
	for i, lit := range outputs {
		if i == invert {
			lit ^= 1
		}
		gia.Gia_ManAppendCo(p, lit)
	}
	return p
}

// C zeros the last two outputs after matching boxes, producing false equivalence.
func TestSolveArithmeticMatchesSourceOutputRemoval(t *testing.T) {
	for invert := -1; invert < 3; invert++ {
		params := &Acec_ParCec_t{}
		Acec_ManCecSetDefaultParams(params)
		a, b := reviewAdderChain(-1), reviewAdderChain(invert)
		status := Acec_Solve(a, b, params)
		want := int32(0)
		if invert != 0 {
			want = 1
		}
		if status != want {
			t.Fatalf("inverted output %d: status %d, want %d", invert, status, want)
		}
		if want == 0 && a.CombinationalCounterexample() == nil {
			t.Fatalf("inverted output %d: missing witness", invert)
		}
	}
}
