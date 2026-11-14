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

func TestSequentialSynthesisTransfersClassesAcrossDomains(t *testing.T) {
	g := gia.Gia_ManStart(16)
	pi0, pi1 := gia.Gia_ManAppendCi(g), gia.Gia_ManAppendCi(g)
	ros := make([]int32, 4)
	for i := range ros {
		ros[i] = gia.Gia_ManAppendCi(g)
	}
	for _, lit := range []int32{pi0, pi0, pi1, pi1} {
		gia.Gia_ManAppendCo(g, lit)
	}
	gia.Gia_ManSetRegNum(g, int32(len(ros)))
	g.SetClockDomains([][]int32{{0, 1}, {2, 3}})
	var pars Cec_ParSeq_t
	Cec_SeqSynthesisSetDefaultParams(&pars)
	if status := Cec_SequentialSynthesisPart(g, &pars); status != 1 {
		t.Fatalf("status = %d", status)
	}
	for _, pair := range [][2]int{{0, 1}, {2, 3}} {
		repr := gia.Gia_ObjRepr(g, ros[pair[1]]>>1)
		if repr != ros[pair[0]]>>1 {
			t.Fatalf("domain %v representative = %d, want %d", pair, repr, ros[pair[0]]>>1)
		}
	}
}

func TestSequentialSynthesisNormalizesClassesAcrossOverlappingDomains(t *testing.T) {
	g := gia.Gia_ManStart(16)
	input := gia.Gia_ManAppendCi(g)
	ros := make([]int32, 3)
	for i := range ros {
		ros[i] = gia.Gia_ManAppendCi(g)
	}
	for range ros {
		gia.Gia_ManAppendCo(g, input)
	}
	gia.Gia_ManSetRegNum(g, int32(len(ros)))
	// The second partition shares register 1 with the first. Its new class
	// should normalize through register 1 to register 0.
	g.SetClockDomains([][]int32{{0, 1}, {1, 2}})
	var pars Cec_ParSeq_t
	Cec_SeqSynthesisSetDefaultParams(&pars)
	if status := Cec_SequentialSynthesisPart(g, &pars); status != 1 {
		t.Fatalf("status = %d", status)
	}
	if got := gia.Gia_ObjRepr(g, ros[1]>>1); got != ros[0]>>1 {
		t.Fatalf("shared register representative = %d, want %d", got, ros[0]>>1)
	}
	if got := gia.Gia_ObjRepr(g, ros[2]>>1); got != ros[0]>>1 {
		t.Fatalf("later-domain class did not normalize through shared register: got %d want %d", got, ros[0]>>1)
	}
}

func TestSequentialSynthesisDoesNotMergeComplementaryClassesAcrossDomains(t *testing.T) {
	g := gia.Gia_ManStart(20)
	x, y := gia.Gia_ManAppendCi(g), gia.Gia_ManAppendCi(g)
	ros := make([]int32, 3)
	for i := range ros {
		ros[i] = gia.Gia_ManAppendCi(g)
	}
	f := gia.Gia_ManAppendAnd(g, x, y)
	for _, lit := range []int32{f, f ^ 1, f} {
		gia.Gia_ManAppendCo(g, lit)
	}
	gia.Gia_ManSetRegNum(g, int32(len(ros)))
	g.SetClockDomains([][]int32{{0, 1}, {1, 2}})
	var pars Cec_ParSeq_t
	Cec_SeqSynthesisSetDefaultParams(&pars)
	if status := Cec_SequentialSynthesisPart(g, &pars); status != 1 {
		t.Fatalf("status = %d", status)
	}
	for i, ro := range ros {
		if repr := gia.Gia_ObjRepr(g, ro>>1); repr != gia.GIA_VOID {
			t.Fatalf("register %d in phase-separated chain got representative %d", i, repr)
		}
	}
}
