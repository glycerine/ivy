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
	defer func() {
		if recover() == nil {
			t.Fatal("nil clock-domain vector must be rejected")
		}
	}()
	Cec_SequentialSynthesisPart(g, &pars)
}
