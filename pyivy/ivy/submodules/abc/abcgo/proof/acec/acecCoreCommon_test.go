package acec

import (
	"github.com/glycerine/abcgo/aig/gia"
	"testing"
)

func TestCommonGraphHashesEquivalentCircuits(t *testing.T) {
	makeGraph := func() (*Gia_Man_t, int32) {
		p := gia.Gia_ManStart(16)
		a, b := gia.Gia_ManAppendCi(p), gia.Gia_ManAppendCi(p)
		n := gia.Gia_ManAppendAnd(p, a, b)
		gia.Gia_ManAppendCo(p, n)
		return p, n >> 1
	}
	a, aid := makeGraph()
	b, bid := makeGraph()
	common := Acec_CommonStart(nil, a)
	before := gia.Gia_ManObjNum(common)
	Acec_CommonStart(common, b)
	if gia.Gia_ManObjNum(common) != before || gia.Gia_ManObj(a, aid).Value != gia.Gia_ManObj(b, bid).Value {
		t.Fatal("identical ANDs did not share a node")
	}
	Acec_CommonFinish(common)
	if gia.Gia_ManCoNum(common) != 1 {
		t.Fatalf("common outputs %d", gia.Gia_ManCoNum(common))
	}
	for x := 0; x < 4; x++ {
		in := []bool{x&1 != 0, x&2 != 0}
		got, err := common.Evaluate(in)
		if err != nil || got[0] != (in[0] && in[1]) {
			t.Fatalf("input %v output %v error %v", in, got, err)
		}
	}
}
func TestDefaultCecParameters(t *testing.T) {
	p := Acec_ParCec_t{FVerbose: 1}
	Acec_ManCecSetDefaultParams(&p)
	if p.NBTLimit != 1000 || p.iOutFail != -1 || p.FVerbose != 0 {
		t.Fatalf("parameters %+v", p)
	}
}
