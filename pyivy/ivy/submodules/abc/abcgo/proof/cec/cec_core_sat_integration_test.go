package cec

import (
	"github.com/glycerine/abcgo/aig/gia"
	"github.com/glycerine/abcgo/misc/util"
	"testing"
)

func TestSatSolvingTransfersAllCOModelsAndProvedSentinels(t *testing.T) {
	g := gia.Gia_ManStart(16)
	pi, ro := gia.Gia_ManAppendCi(g), gia.Gia_ManAppendCi(g)
	x := gia.Gia_ManAppendAnd(g, pi, ro)
	nx := gia.Gia_ManAppendAnd(g, pi^1, ro)
	zero := gia.Gia_ManAppendAnd(g, x, nx)
	gia.Gia_ManAppendCo(g, zero)
	gia.Gia_ManAppendCo(g, pi)
	gia.Gia_ManAppendCo(g, ro)
	gia.Gia_ManSetRegNum(g, 1)
	p := Cec_ParSat_t{SolverType: -1, NBTLimit: 1000, FSaveCexes: 1}
	out := Cec_ManSatSolving(g, &p, 1)
	if out == nil || gia.Gia_ManRegNum(out) != 1 {
		t.Fatal("missing sequential graph")
	}
	if g.CombinationalCounterexamples() != nil {
		t.Fatal("source retained transferred models")
	}
	models := out.CombinationalCounterexamples()
	if len(models) != 3 || models[0] != util.AbcCexSentinel {
		t.Fatalf("invalid proved sentinel: %+v", models)
	}
	if out.OutputLiterals()[0] != 0 {
		t.Fatal("proved output not patched to zero")
	}
	for i := 1; i < 3; i++ {
		cex := models[i]
		if cex == nil || cex == util.AbcCexSentinel || cex.IPo != int32(i) || cex.NRegs != 0 || cex.NPis != 2 {
			t.Fatalf("output %d witness %+v", i, cex)
		}
		inputs := []bool{cex.PData[0]&1 != 0, cex.PData[0]&2 != 0}
		vals, err := out.Evaluate(inputs)
		if err != nil {
			t.Fatal(err)
		}
		if !vals[i] {
			t.Fatalf("output %d witness %v does not replay", i, inputs)
		}
	}
}
