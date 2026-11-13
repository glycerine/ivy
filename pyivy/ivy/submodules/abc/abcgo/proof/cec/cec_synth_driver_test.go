package cec

import (
	"github.com/glycerine/abcgo/aig/gia"
	"testing"
)

func TestSequentialSynthesisWithNoClockDomains(t *testing.T) {
	g := gia.Gia_ManStart(8)
	in := gia.Gia_ManAppendCi(g)
	gia.Gia_ManAppendCo(g, in)
	var pars Cec_ParSeq_t
	Cec_SeqSynthesisSetDefaultParams(&pars)
	if Cec_SequentialSynthesisPart(g, &pars) != 1 {
		t.Fatal("synthesis did not complete")
	}
	if !g.HasRepresentatives() {
		t.Fatal("missing empty representative table")
	}
	for id := int32(0); id < gia.Gia_ManObjNum(g); id++ {
		if gia.Gia_ObjRepr(g, id) != gia.GIA_VOID {
			t.Fatalf("unexpected representative of %d", id)
		}
	}
}
