package cec

import (
	"github.com/glycerine/abcgo/aig/gia"
	"github.com/glycerine/abcgo/misc/util"
	"testing"
)

func TestSatSolveStoresCombinationalModelsForAllCOs(t *testing.T) {
	g := gia.Gia_ManStart(16)
	pi := gia.Gia_ManAppendCi(g)
	state := gia.Gia_ManAppendCi(g)
	gia.Gia_ManAppendCo(g, pi)
	gia.Gia_ManAppendCo(g, state)
	gia.Gia_ManSetRegNum(g, 1)
	Cec_ManSatSolve(nil, g, &Cec_ParSat_t{nBTLimit: 1000, fSaveCexes: 1}, nil, nil, nil, 0)
	models := g.CombinationalCounterexamples()
	if len(models) != 2 {
		t.Fatalf("stored %d models, want 2 COs", len(models))
	}
	for i, m := range models {
		if m == nil || m == util.AbcCexSentinel || m.NRegs != 0 || m.NPis != 2 || m.IPo != int32(i) {
			t.Fatalf("CO %d model %#v", i, m)
		}
	}
}
