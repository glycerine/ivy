package framgr

import (
	"github.com/glycerine/abcgo/aig/aig"
	"github.com/glycerine/abcgo/misc/vec"
	"math/bits"
	"testing"
)

func TestClassAdaptersChooseSupportAndDeriveFrames(t *testing.T) {
	g := aig.NewGraph()
	x, y, z, q := g.NewInput(), g.NewInput(), g.NewInput(), g.NewInput()
	wide := g.AND(g.AND(x, y), z)
	narrow := g.AND(x, y.Not())
	g.AddOutput(x)
	aig.Aig_ManSetRegNum(g, 1)
	p := Fra_ClassesStart(g)
	Fra_ClassesTest(p, int32(wide.NodeID()), int32(narrow.NodeID()))
	Fra_ClassesSelectRepr(p)
	if p.Reprs[wide.NodeID()] != narrow.NodeID() || p.Reprs[narrow.NodeID()] != -1 {
		t.Fatalf("representatives = %v", p.Reprs)
	}
	p.Reprs[wide.NodeID()] = -1
	p.Reprs[narrow.NodeID()] = -1
	p.Reprs[q.NodeID()] = 0
	result := Fra_ClassesDeriveAig(p, 1)
	if result.InputCount() != 7 || result.OutputCount() != 2 || result.Assertions != 1 {
		t.Fatalf("derived interface = %d/%d assertions %d", result.InputCount(), result.OutputCount(), result.Assertions)
	}
	for mask := 0; mask < 128; mask++ {
		v := make([]bool, 7)
		for i := range v {
			v[i] = mask&(1<<i) != 0
		}
		out, err := result.Evaluate(v)
		if err != nil || out[0] != !v[0] || out[1] != !v[1] {
			t.Fatalf("inputs %v outputs %v err %v", v, out, err)
		}
	}
}

func TestBMCUnrollAdapterInitializedOutputs(t *testing.T) {
	g := aig.NewGraph()
	x, q := g.NewInput(), g.NewInput()
	g.AddOutput(q)
	g.AddOutput(x)
	aig.Aig_ManSetRegNum(g, 1)
	out, mapping := UnrollInitialized(g, 3, true)
	if mapping[q.NodeID()*3] != out.Const0() {
		t.Fatal("latch is not initialized to zero")
	}
	for mask := 0; mask < 8; mask++ {
		v := []bool{mask&1 != 0, mask&2 != 0, mask&4 != 0}
		got, err := out.Evaluate(v)
		if err != nil || len(got) != 3 || got[0] || got[1] != v[0] || got[2] != v[1] {
			t.Fatalf("inputs %v outputs %v err %v", v, got, err)
		}
	}
}

func TestBMCSweepAdapterFiltersImplications(t *testing.T) {
	g := aig.NewGraph()
	x, q := g.NewInput(), g.NewInput()
	gate := g.AND(x, q)
	g.AddOutput(gate)
	g.AddOutput(x)
	aig.Aig_ManSetRegNum(g, 1)
	frames, mapping := UnrollInitialized(g, 2, false)
	imps := vec.Vec_IntAlloc(2)
	yes := Fra_ImpCreate(int32(gate.NodeID()), int32(gate.NodeID()))
	no := Fra_ImpCreate(int32(gate.NodeID()), 0)
	vec.Vec_IntPush(imps, yes)
	vec.Vec_IntPush(imps, no)
	out, mapped := SweepBMCImplications(frames, g, 0, 2, mapping, imps, 1000)
	if out == nil || len(mapped) == 0 {
		t.Fatal("sweep did not return graph and mapping")
	}
	if vec.Vec_IntSize(imps) != 1 || vec.Vec_IntEntry(imps, 0) != yes {
		t.Fatalf("remaining implications = %v", vec.Vec_IntArray(imps))
	}
}

// Source fraClass.c weights each candidate's 32-word simulation distance,
// then discards candidates below half the largest distance.
func TestClassPostprocessUsesGlobalSimulationWeight(t *testing.T) {
	g := aig.NewGraph()
	a, b, c := g.NewInput(), g.NewInput(), g.NewInput()
	u := g.AND(g.AND(a, b), c)
	v := g.AND(a, g.AND(b, c))
	p := Fra_ClassesStart(g)
	p.Constants = []int{a.NodeID()}
	p.Classes = [][]int{{u.NodeID(), v.NodeID()}}
	p.Reprs[a.NodeID()] = 0
	p.Reprs[v.NodeID()] = u.NodeID()
	aig.Aig_ManRandom(1)
	sim := Fra_SmlSimulateComb(g, 32, 0)
	wa, wc := 0, 0
	for i := 0; i < 32; i++ {
		wa += bits.OnesCount32(sim.NodeWords(a.NodeID())[i])
		wc += bits.OnesCount32(sim.NodeWords(u.NodeID())[i] ^ sim.NodeWords(v.NodeID())[i])
	}
	maxw := wa
	if wc > maxw {
		maxw = wc
	}
	threshold := maxw / 2
	if threshold == 0 || wc != 0 {
		t.Fatalf("fixture does not exercise pruning: threshold=%d class weight=%d", threshold, wc)
	}
	aig.Aig_ManRandom(1)
	Fra_ClassesPostprocess(p)
	if got := len(p.Constants) > 0; got != (wa >= threshold) {
		t.Errorf("constant retained=%v weight %d threshold %d", got, wa, threshold)
	}
	if got := len(p.Classes) > 0; got != (wc >= threshold) {
		t.Errorf("class retained=%v weight %d threshold %d", got, wc, threshold)
	}
}
