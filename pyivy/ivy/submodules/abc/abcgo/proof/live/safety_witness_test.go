package live

import (
	"testing"

	"github.com/glycerine/abcgo/aig/aig"
	"github.com/glycerine/abcgo/misc/util"
)

func TestRejectMalformedPDRWitnesses(t *testing.T) {
	g := aig.NewGraph()
	input, state := g.NewInput(), g.NewInput()
	g.AddOutput(input)
	g.AddOutput(state)
	aig.Aig_ManSetRegNum(g, 1)
	valid := util.Abc_CexAlloc(1, 1, 1)
	util.Abc_InfoSetBit(valid.PData, 1)
	if got := decodeSafetyWitness(g, 0, valid, 1); got.Status != 0 {
		t.Fatalf("valid witness rejected: %+v", got)
	}
	truncated := *valid
	truncated.PData = nil
	badLength := *valid
	badLength.NBits++
	nonzero := *valid
	nonzero.PData = append([]uint32(nil), valid.PData...)
	util.Abc_InfoSetBit(nonzero.PData, 0)
	badValue := *valid
	badValue.PData = []uint32{0}
	for i, cex := range []*util.Abc_Cex_t{nil, &truncated, &badLength, &nonzero, &badValue} {
		got := decodeSafetyWitness(g, 0, cex, 1)
		if got.Status != -1 || got.Error == nil {
			t.Fatalf("case %d accepted: %+v", i, got)
		}
	}
}
