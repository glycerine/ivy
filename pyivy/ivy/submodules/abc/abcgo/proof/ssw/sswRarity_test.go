package ssw

import (
	"github.com/glycerine/abcgo/aig/aig"
	"github.com/glycerine/abcgo/aig/saig"
	"github.com/glycerine/abcgo/misc/util"
	"github.com/glycerine/abcgo/misc/vec"
	"testing"
)

func TestRarityFilterReturnsInvalidWitnessFromNonzeroPrefix(t *testing.T) {
	g := aig.NewGraph()
	x := g.NewInput()
	r := g.NewInput()
	g.AddOutput(g.AND(x, r))
	g.AddOutput(g.Const1()) // next(r) = 1
	aig.Aig_ManSetRegNum(g, 1)

	prefix := util.Abc_CexAlloc(1, 1, 2)
	prefix.IFrame = 1
	util.Abc_InfoSetBit(prefix.PData, 2) // x = 1 in frame 1
	if saig.Saig_ManVerifyCex(g, prefix) == 0 {
		t.Fatal("expected a valid prefix reaching r = 1")
	}
	pars := &Ssw_RarPars_t{NWords: 1, NFrames: 1, NBinSize: 1, NRounds: 1, FMiter: 1, PCex: prefix}
	status := Ssw_RarSignalFilter(g, pars)
	witness := g.SequentialCounterexample()
	if status != 0 || witness == nil || saig.Saig_ManVerifyCex(g, witness) != 0 {
		t.Fatal("expected ABC's failure status and invalid zero-initialized witness")
	}
}

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
	pars := &Ssw_RarPars_t{NWords: 2, NFrames: 1, NBinSize: 8}
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
	p := Ssw_RarManStart(g, &Ssw_RarPars_t{NWords: 1, NFrames: 1, NBinSize: 1})
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

func TestSswRarityFindsSequentialCounterexample(t *testing.T) {
	g := aig.NewGraph()
	pi, lo := g.NewInput(), g.NewInput()
	g.AddOutput(lo)
	g.AddOutput(pi)
	aig.Aig_ManSetRegNum(g, 1)
	pars := &Ssw_RarPars_t{NWords: 1, NFrames: 2, NBinSize: 8, NRounds: 1}
	if got := Ssw_RarSimulate(g, pars); got != 0 {
		t.Fatalf("result=%d", got)
	}
	cex := g.SequentialCounterexample()
	if cex == nil || cex.IFrame != 1 {
		t.Fatalf("counterexample: %+v", cex)
	}
}
func TestSswRarityTransferPreservesRegisterPatterns(t *testing.T) {
	g := aig.NewGraph()
	g.NewInput()
	for i := 0; i < 8; i++ {
		g.NewInput()
	}
	g.AddOutput(g.Const0())
	for i := 0; i < 8; i++ {
		g.AddOutput(g.Const0())
	}
	aig.Aig_ManSetRegNum(g, 8)
	p := Ssw_RarManStart(g, &Ssw_RarPars_t{NWords: 1, NFrames: 1, NBinSize: 8})
	defer Ssw_RarManStop(p)
	for i, li := range g.Outputs()[1:] {
		Ssw_RarObjSim(p, int32(li.NodeID()))[0] = uint64(i + 1)
	}
	init := vec.Vec_IntAlloc(8)
	Ssw_RarTransferPatterns(p, init)
	if vec.Vec_IntSize(init) != 8 || vec.Vec_IntSize(p.vPatBests) != 1 {
		t.Fatal("incorrect selected pattern size")
	}
	idx := vec.Vec_IntEntry(p.vPatBests, 0)
	for i, li := range g.Outputs()[1:] {
		want := int32(Ssw_RarObjSim(p, int32(li.NodeID()))[0] >> uint(idx) & 1)
		if vec.Vec_IntEntry(init, i) != want {
			t.Fatalf("reg=%d selected bit mismatch", i)
		}
	}
}

func TestRarityHashRetainsSourceRawWordBehavior(t *testing.T) {
	g := aig.NewGraph()
	a, b := g.NewInput(), g.NewInput()
	g.AddOutput(a)
	g.AddOutput(b)
	aig.Aig_ManSetRegNum(g, 1)
	aig.Aig_ObjSetPhase(g, b, true)
	p := Ssw_RarManStart(g, &Ssw_RarPars_t{NWords: 1, NFrames: 1, NBinSize: 1})
	defer Ssw_RarManStop(p)
	Ssw_RarObjSim(p, int32(a.NodeID()))[0] = 0x0123456789abcdef
	Ssw_RarObjSim(p, int32(b.NodeID()))[0] = ^uint64(0x0123456789abcdef)
	if Ssw_RarManObjsAreEqual(p, a, b) != 1 {
		t.Fatal("test patterns should be phase-normalized equal")
	}
	if x, y := Ssw_RarManObjHashWord(p, a), Ssw_RarManObjHashWord(p, b); x == y {
		t.Fatalf("source hashes raw words, unexpectedly equal: %x", x)
	}
}

func TestRarityRetainsSourceInvalidCounterexampleBehavior(t *testing.T) {
	g := aig.NewGraph()
	pi := g.NewInput()
	g.NewInput()
	g.AddOutput(g.Const0())
	g.AddOutput(pi)
	aig.Aig_ManSetRegNum(g, 1)
	p := Ssw_RarManStart(g, &Ssw_RarPars_t{NWords: 1, NFrames: 1, NBinSize: 1})
	defer Ssw_RarManStop(p)
	if cex := Ssw_RarDeriveCex(p, 0, 0, 0, 0); cex == nil || saig.Saig_ManVerifyCex(g, cex) != 0 {
		t.Fatal("source returns the invalid counterexample after warning")
	}
}
