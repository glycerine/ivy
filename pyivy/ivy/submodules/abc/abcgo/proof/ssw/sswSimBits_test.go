package ssw

import (
	"github.com/glycerine/abcgo/aig/aig"
	"testing"
)

func TestBitComparisonsAndConstantPatterns(t *testing.T) {
	g := aig.NewGraph()
	pi, lo := g.NewInput(), g.NewInput()
	g.AddOutput(lo)
	aig.Aig_ManSetRegNum(g, 1)
	if Ssw_SmlObjIsConstBit(g, pi) != 1 || Ssw_SmlObjsAreEqualBit(g, pi, lo) != 1 {
		t.Fatal("zero marks should match")
	}
	aig.Aig_ObjSetMarkB(g, lo)
	if Ssw_SmlObjIsConstBit(g, lo) != 0 || Ssw_SmlObjsAreEqualBit(g, pi, lo) != 0 {
		t.Fatal("different marks should differ")
	}
	p := &Ssw_Man_t{pAig: g, nFrames: 2, pPatWords: make([]uint32, 1)}
	Ssw_SmlSavePattern1(p, 1)
	if p.pPatWords[0] != (^uint32(0) ^ (1 << 2)) {
		t.Fatal("initial latch bit not cleared")
	}
	Ssw_SmlSavePattern0(p, 0)
	if p.pPatWords[0] != 0 {
		t.Fatal("pattern not cleared")
	}
}
