package cec

import (
	"github.com/glycerine/abcgo/aig/gia"
	"testing"
)

func TestCec4EntryReplaysNontrivialMiterWitness(t *testing.T) {
	for mode := int32(0); mode < 3; mode++ {
		g := gia.Gia_ManStart(10)
		a, b := gia.Gia_ManAppendCi(g), gia.Gia_ManAppendCi(g)
		mismatch := gia.Gia_ManAppendAnd(g, a, b^1)
		gia.Gia_ManAppendCo(g, 0)
		gia.Gia_ManAppendCo(g, mismatch)
		p := Cec_ParFra_t{jType: mode, nWords: 1, nRounds: 1, nItersMax: 10, nBTLimit: 100, fCheckMiter: 1}
		if got := Cec4_ManPerformSweeping(g, &p, nil, 0); got != 0 {
			t.Fatalf("mode %d status %d", mode, got)
		}
		cex := g.SequentialCounterexample()
		if cex == nil || cex.IPo != 1 {
			t.Fatalf("mode %d witness %+v", mode, cex)
		}
		inputs := []bool{cex.PData[0]&1 != 0, cex.PData[0]&2 != 0}
		outputs, err := g.Evaluate(inputs)
		if err != nil {
			t.Fatal(err)
		}
		if !outputs[1] {
			t.Fatalf("mode %d witness %v did not fail", mode, inputs)
		}
	}
}
