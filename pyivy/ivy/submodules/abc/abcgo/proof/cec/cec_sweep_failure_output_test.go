package cec

import (
	"io"
	"os"
	"strings"
	"testing"

	"github.com/glycerine/abcgo/aig/gia"
)

// cecCore.c reports the failing miter output unless fSilent is set.
func TestSourceSweepFailureDiagnostic(t *testing.T) {
	for _, silent := range []int32{0, 1} {
		g := gia.Gia_ManStart(8)
		a, b := gia.Gia_ManAppendCi(g), gia.Gia_ManAppendCi(g)
		gia.Gia_ManAppendCo(g, a)
		gia.Gia_ManAppendCo(g, b)
		gia.Gia_ManCecSimAlloc(g, 1)
		gia.Gia_ManCecSaveCIWord(g, 2)
		gia.Gia_ManCecSaveCIWord(g, 0)
		p := Cec_ParFra_t{nWords: 1, nRounds: 1, nItersMax: 2, nBTLimit: 100, fCheckMiter: 1, fDualOut: 1, iOutFail: -1}
		r, w, err := os.Pipe()
		if err != nil {
			t.Fatal(err)
		}
		old := os.Stdout
		os.Stdout = w
		out := Cec_ManSatSweeping(g, &p, silent)
		w.Close()
		os.Stdout = old
		printed, err := io.ReadAll(r)
		r.Close()
		if err != nil {
			t.Fatal(err)
		}
		if out != nil || p.Error != nil || p.iOutFail != 0 || g.CombinationalCounterexample() == nil {
			t.Fatalf("silent=%d: result=%v error=%v output=%d", silent, out, p.Error, p.iOutFail)
		}
		want := ""
		if silent == 0 {
			want = "Disproved at least one output of the miter (zero-based number 0).\n"
		}
		if string(printed) != want {
			t.Fatalf("silent=%d: printed %q; want %q", silent, strings.TrimSpace(string(printed)), strings.TrimSpace(want))
		}
	}
}
