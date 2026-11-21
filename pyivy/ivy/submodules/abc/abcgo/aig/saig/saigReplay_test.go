package saig

import (
	"testing"

	"github.com/glycerine/abcgo/aig/aig"
	"github.com/glycerine/abcgo/misc/util"
	"github.com/glycerine/abcgo/misc/vec"
)

func TestFailingStateUsesRegisterCountWithExtraOutputs(t *testing.T) {
	g := aig.NewGraph()
	q := g.NewInput()
	g.AddOutput(q)
	g.AddOutput(q.Not())
	g.AddOutput(g.Const1())
	aig.Aig_ManSetRegNum(g, 1)
	g.SetTruePoCount(1)
	c := util.Abc_CexAlloc(1, 0, 1)
	c.PData[0] = 1
	got := Saig_ManReturnFailingState(g, c, 1)
	if got == nil || vec.Vec_IntSize(got) != 1 || vec.Vec_IntEntry(got, 0) != 0 {
		t.Fatalf("wanted the single next-state bit, got %v", got)
	}
}

func TestReplayRejectsWrappedBitCountBeforeFirstFrame(t *testing.T) {
	g := aig.NewGraph()
	g.NewInput()
	g.NewInput()
	g.AddOutput(g.Const1())
	aig.Aig_ManSetRegNum(g, 0)
	c := &util.Abc_Cex_t{IFrame: (1 << 30) - 1, NPis: 2, NBits: -(1 << 31), PData: []uint32{0}}
	defer func() {
		if recover() == nil {
			t.Fatal("accepted a wrapped bit count")
		}
	}()
	replayCexFrames(g, c, func(int32, []bool) { t.Error("evaluated a frame with a wrapped bit count"); panic("stop") })
}

func TestReplayRewrittenDanglingLogicAndSparseIDs(t *testing.T) {
	g := aig.NewGraph()
	a, q := g.NewInput(), g.NewInput()
	buf := aig.Aig_ObjCreate(g, &aig.Node{Kind: aig.Buffer, Fanins: [2]aig.Edge{q, aig.InvalidEdge}})
	x := g.NewXor(q, a)
	aig.Aig_ObjConnect(g, buf, x.Not(), aig.InvalidEdge)
	g.AddOutput(buf)
	g.AddOutput(x)
	hole := g.AND(q, a)
	aig.Aig_ManCleanup(g)
	dangling := g.AND(q.Not(), a)
	if g.ValidEdge(hole) || !g.ValidEdge(dangling) {
		t.Fatal("fixture")
	}
	aig.Aig_ManSetRegNum(g, 1)
	c := util.Abc_CexAlloc(1, 1, 3)
	c.IFrame = 2
	c.PData[0] = 11 // initial 1, inputs 1,0,1; final state 0
	if Saig_ManVerifyCexNoClear(g, c) != 0 || !g.Node(dangling).MarkB || g.Node(buf).MarkB || !g.Node(x).MarkB {
		t.Fatal("final-frame marks")
	}
	if Saig_ManVerifyCex(g, c) != 0 {
		t.Fatal("output")
	}
	for _, n := range g.Nodes() {
		if n.MarkB {
			t.Fatal("marks after replay")
		}
	}
}

func BenchmarkReplayEnabledCounter(b *testing.B) {
	const width = 6
	g := aig.NewGraph()
	var state, next [width]aig.Edge
	carry, bad := g.NewInput(), g.Const1()
	for i := range state {
		state[i] = g.NewInput()
		bad = g.AND(bad, state[i])
	}
	for i, q := range state {
		next[i] = g.XOR(q, carry)
		carry = g.AND(carry, q)
	}
	g.AddOutput(bad)
	for _, v := range next {
		g.AddOutput(v)
	}
	aig.Aig_ManSetRegNum(g, width)
	c := util.Abc_CexAlloc(width, 1, 64)
	c.IFrame = 63
	for i := width; i < int(c.NBits); i++ {
		c.PData[i/32] |= 1 << uint(i%32)
	}
	b.ReportAllocs()
	for b.Loop() {
		if Saig_ManVerifyCex(g, c) != 1 {
			b.Fatal("failed replay")
		}
	}
}

func TestReplayFrameBuffersAreReusable(t *testing.T) {
	g := dupFixture()
	order := g.TopologicalObjects()
	values := make([]bool, aig.Aig_ManObjNumMax(g))
	for i := range values {
		values[i] = true
	}
	for mask := 3; mask >= 0; mask-- {
		inputs := []bool{mask&1 != 0, mask&2 != 0}
		g.SimulateFrame(order, inputs, values)
		want, err := g.Evaluate(inputs)
		if err != nil {
			t.Fatal(err)
		}
		for i, e := range g.Outputs() {
			if values[e.NodeID()] != want[i] {
				t.Fatalf("inputs %02b output %d", mask, i)
			}
		}
	}
}
