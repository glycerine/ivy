package acec

import (
	"github.com/glycerine/abcgo/aig/gia"
	"testing"
)

func TestSolveMismatchedArithmeticInterfacesReturnsUnknown(t *testing.T) {
	for _, extraInput := range []bool{true, false} {
		a, b := reviewAdderChain(-1), reviewAdderChain(-1)
		if extraInput {
			gia.Gia_ManAppendCi(b)
		} else {
			gia.Gia_ManAppendCo(b, 0)
		}
		var opts Acec_ParCec_t
		Acec_ManCecSetDefaultParams(&opts)
		if got := Acec_Solve(a, b, &opts); got != -1 {
			t.Fatalf("extra input %v: got %d, want unknown", extraInput, got)
		}
	}
}
