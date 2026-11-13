package ssw

import (
	"github.com/glycerine/abcgo/aig/aig"
	"github.com/glycerine/abcgo/misc/vec"
	"testing"
)

func TestSswRarityTransposeAndInitialization(t *testing.T) {
	g := aig.NewGraph()
	g.NewInput()
	for i := 0; i < 70; i++ {
		g.NewInput()
	}
	g.AddOutput(g.Const0())
	for i := 0; i < 70; i++ {
		g.AddOutput(g.Const0())
	}
	aig.Aig_ManSetRegNum(g, 70)
	pars := &Ssw_RarPars_t{nWords: 2, nFrames: 1, nBinSize: 8}
	p := Ssw_RarManStart(g, pars)
	defer Ssw_RarManStop(p)
	init := vec.Vec_IntAlloc(140)
	for w := 0; w < 2; w++ {
		for i := 0; i < 70; i++ {
			value := int32((i + w) & 1)
			vec.Vec_IntPush(init, value)
		}
	}
	Ssw_RarManInitialize(p, init)
	for i := 0; i < 70; i++ {
		lo := g.Inputs()[i+1]
		s := Ssw_RarObjSim(p, int32(lo.NodeID()))
		for w := 0; w < 2; w++ {
			want := uint64(0)
			if (i+w)&1 != 0 {
				want = ^uint64(0)
			}
			if s[w] != want {
				t.Fatalf("LO %d word %d", i, w)
			}
		}
		li := g.Outputs()[i+1]
		copy(Ssw_RarObjSim(p, int32(li.NodeID())), s)
	}
	Ssw_RarTranspose(p)
	for pat := 0; pat < 128; pat++ {
		data := Ssw_RarPatSim(p, int32(pat))
		for reg := 0; reg < 70; reg++ {
			got := data[reg/64] >> uint(reg%64) & 1
			want := uint64((reg + pat/64) & 1)
			if got != want {
				t.Fatalf("pat=%d reg=%d got=%d want=%d", pat, reg, got, want)
			}
		}
	}
}
func TestSswRarityBinCountersAndObjectComparison(t *testing.T) {
	g := aig.NewGraph()
	a, b := g.NewInput(), g.NewInput()
	g.AddOutput(a)
	g.AddOutput(b)
	aig.Aig_ManSetRegNum(g, 1)
	p := Ssw_RarManStart(g, &Ssw_RarPars_t{nWords: 1, nFrames: 1, nBinSize: 1})
	defer Ssw_RarManStop(p)
	Ssw_RarAddToBinPat(p, 0, 1)
	Ssw_RarAddToBinPat(p, 0, 1)
	if Ssw_RarGetBinPat(p, 0, 1) != 2 {
		t.Fatal("counter")
	}
	Ssw_RarObjSim(p, int32(a.NodeID()))[0] = 4
	Ssw_RarObjSim(p, int32(b.NodeID()))[0] = 4
	if Ssw_RarManObjWhichOne(p, a) != 2 || Ssw_RarManObjsAreEqual(p, a, b) != 1 {
		t.Fatal("simulation comparison")
	}
}
