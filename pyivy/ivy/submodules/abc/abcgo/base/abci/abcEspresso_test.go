package abci

import (
	"testing"

	"github.com/glycerine/abcgo/base/abc"
)

func TestEspressoSOPCoverRoundTrip(t *testing.T) {
	for _, sop := range []string{" 0\n", " 1\n", "01 1\n10 1\n", "1- 1\n"} {
		cover := Abc_SopToEspresso(sop)
		got := Abc_SopFromEspresso(cover)
		want := sop
		if abc.Abc_SopIsConst0(sop) != 0 {
			want = abc.Abc_SopCreateConst0(nil)
		}
		if got != want {
			t.Errorf("round trip %q -> %q, want %q", sop, got, want)
		}
	}
}
