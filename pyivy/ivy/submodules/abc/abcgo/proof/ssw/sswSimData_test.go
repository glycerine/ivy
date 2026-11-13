package ssw

import (
	"testing"

	"github.com/glycerine/abcgo/aig/aig"
	"github.com/glycerine/abcgo/misc/vec"
)

func TestSimulationDataLayoutAndPrefixComparisons(t *testing.T) {
	g := aig.NewGraph()
	a, b := g.NewInput(), g.NewInput()
	p := Ssw_SmlStart(g, 1, 2, 2)
	if Ssw_SmlNumFrames(p) != 3 || Ssw_SmlNumWordsTotal(p) != 6 {
		t.Fatalf("bad dimensions: %+v", p)
	}
	x, y := Ssw_SmlSimInfo(p, a), Ssw_SmlSimInfo(p, b)
	x[0], x[1] = 3, 5
	if Ssw_SmlObjIsConstWord(p, a) != 1 || Ssw_SmlObjsAreEqualWord(p, a, b) != 1 {
		t.Fatal("prefix affected comparisons")
	}
	x[2], y[2] = 7, 7
	x[5], y[5] = 11, 11
	wantHash := uint32(7*1093 ^ 11*1249)
	if Ssw_SmlObjHashWord(p, a) != wantHash || Ssw_SmlObjIsConstWord(p, a) != 0 || Ssw_SmlObjsAreEqualWord(p, a, b) != 1 {
		t.Fatal("incorrect hash or comparison")
	}
	y[3] = 1
	if Ssw_SmlObjsAreEqualWord(p, a, b) != 0 {
		t.Fatal("different suffixes compared equal")
	}
	data := Ssw_SmlSimDataPointers(p)
	view := vec.Vec_PtrEntry(data, a.NodeID()).([]uint32)
	view[3] = 17
	if x[3] != 17 {
		t.Fatal("simulation vector should expose the live object slice")
	}
	Ssw_SmlClean(p)
	for _, value := range p.pData {
		if value != 0 {
			t.Fatal("clean retained simulation words")
		}
	}
	Ssw_SmlStop(p)
	if p.pData != nil || p.pAig != nil {
		t.Fatal("stop retained owned storage")
	}
}

func TestSimulationDimensionsAndComplementValidation(t *testing.T) {
	g := aig.NewGraph()
	x := g.NewInput()
	for _, start := range []func(){
		func() { Ssw_SmlStart(g, -1, 1, 1) },
		func() { Ssw_SmlStart(g, 2147483647, 1, 1) },
		func() { Ssw_SmlSimInfo(Ssw_SmlStart(g, 0, 1, 1), x.Not()) },
	} {
		func() {
			defer func() {
				if recover() == nil {
					t.Error("invalid simulation input accepted")
				}
			}()
			start()
		}()
	}
}
