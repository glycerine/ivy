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

func TestSequentialSynthesisClockDomainOptions(t *testing.T) {
	for _, mode := range []string{"cleanup", "scorr", "lcorr"} {
		t.Run(mode, func(t *testing.T) {
			g := gia.Gia_ManStart(12)
			a := gia.Gia_ManAppendCi(g)
			gia.Gia_ManAppendCi(g)
			gia.Gia_ManAppendCi(g)
			gia.Gia_ManAppendCo(g, a)
			gia.Gia_ManAppendCo(g, a)
			gia.Gia_ManAppendCo(g, a)
			gia.Gia_ManSetRegNum(g, 2)
			g.SetClockDomains([][]int32{{0, 1}})
			var p Cec_ParSeq_t
			Cec_SeqSynthesisSetDefaultParams(&p)
			if mode == "scorr" {
				p.fUseScorr = 1
			}
			if mode == "lcorr" {
				p.fUseLcorr = 1
			}
			if got := Cec_SequentialSynthesisPart(g, &p); got != 1 || !g.HasRepresentatives() {
				t.Fatalf("status=%d classes=%v", got, g.HasRepresentatives())
			}
			r0 := gia.Gia_ObjId(g, gia.Gia_ManRo(g, 0))
			r1 := gia.Gia_ObjId(g, gia.Gia_ManRo(g, 1))
			if got := gia.Gia_ObjRepr(g, r1); got != r0 {
				t.Fatalf("register representative=%d want %d", got, r0)
			}
		})
	}
}

func TestSequentialSynthesisSkipsDomainWithoutPrimaryInputs(t *testing.T) {
	g := gia.Gia_ManStart(8)
	r := gia.Gia_ManAppendCi(g)
	gia.Gia_ManAppendCo(g, r)
	gia.Gia_ManSetRegNum(g, 1)
	g.SetClockDomains([][]int32{{0}})
	var p Cec_ParSeq_t
	Cec_SeqSynthesisSetDefaultParams(&p)
	if Cec_SequentialSynthesisPart(g, &p) != 1 || !g.HasRepresentatives() {
		t.Fatal("self-only domain failed")
	}
	if gia.Gia_ObjRepr(g, r>>1) != gia.GIA_VOID {
		t.Fatal("source skipped domain, but Go changed its representative")
	}
}

func TestSequentialSynthesisEmptyClockDomainVector(t *testing.T) {
	g := gia.Gia_ManStart(8)
	a := gia.Gia_ManAppendCi(g)
	gia.Gia_ManAppendCo(g, a)
	g.SetClockDomains([][]int32{})
	var p Cec_ParSeq_t
	Cec_SeqSynthesisSetDefaultParams(&p)
	if Cec_SequentialSynthesisPart(g, &p) != 1 || !g.HasRepresentatives() {
		t.Fatal("empty vector should complete with empty classes")
	}
}
