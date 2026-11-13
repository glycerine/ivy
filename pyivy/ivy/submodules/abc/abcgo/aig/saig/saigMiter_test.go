package saig

import (
	"github.com/glycerine/abcgo/aig/aig"
	"testing"
)

func miterFixture(change bool) *aig.Graph {
	g := aig.NewGraph()
	g.Name = "machine"
	a, b, q := g.NewInput(), g.NewInput(), g.NewInput()
	e := g.AND(a, q.Not())
	if change {
		e = g.OR(a, q)
	}
	g.AddOutput(e)
	g.AddOutput(g.XOR(b, q))
	aig.Aig_ManSetRegNum(g, 1)
	return g
}
func miterEval(t *testing.T, g *aig.Graph, in []bool) []bool {
	t.Helper()
	if err := g.Check(); err != nil {
		t.Fatal(err)
	}
	out, err := g.Evaluate(in)
	if err != nil {
		t.Fatal(err)
	}
	return out
}
func TestMiterSequentialAndCombinational(t *testing.T) {
	a, b := miterFixture(false), miterFixture(true)
	for oper := int32(0); oper < 2; oper++ {
		g := Saig_ManCreateMiter(a, b, oper)
		if Saig_ManPiNum(g) != 2 || Saig_ManPoNum(g) != 1 || Saig_ManRegNum(g) != 2 {
			t.Fatal("sequential interface")
		}
		for x := 0; x < 16; x++ {
			in := []bool{x&1 != 0, x&2 != 0, x&4 != 0, x&8 != 0}
			left := miterEval(t, a, in[:3])
			right := miterEval(t, b, []bool{in[0], in[1], in[3]})
			out := miterEval(t, g, in)
			want := left[0] != right[0]
			if oper == 1 {
				want = left[0] && !right[0]
			}
			if out[0] != want || out[1] != left[1] || out[2] != right[1] {
				t.Fatalf("oper%d assignment%d", oper, x)
			}
		}
		comb := Saig_ManCreateMiterComb(a, b, oper)
		if Saig_ManRegNum(comb) != 0 {
			t.Fatal("comb has registers")
		}
		for x := 0; x < 8; x++ {
			in := []bool{x&1 != 0, x&2 != 0, x&4 != 0}
			left, right := miterEval(t, a, in), miterEval(t, b, in)
			out := miterEval(t, comb, in)
			for i := range out {
				want := left[i] != right[i]
				if oper == 1 {
					want = left[i] && !right[i]
				}
				if out[i] != want {
					t.Fatal("combinational mismatch")
				}
			}
		}
	}
}
func TestMiterUnrollTwo(t *testing.T) {
	bot, top := miterFixture(true), miterFixture(false)
	for frames := int32(1); frames <= 3; frames++ {
		g := Saig_ManUnrollTwo(bot, top, frames)
		if g.InputCount() != 1+int(frames)*2 || len(g.Outputs()) != int(frames)+1 || Saig_ManRegNum(g) != 0 {
			t.Fatal("unrolled interface")
		}
		for x := 0; x < 1<<g.InputCount(); x++ {
			in := make([]bool, g.InputCount())
			for i := range in {
				in[i] = x&(1<<i) != 0
			}
			out := miterEval(t, g, in)
			state := in[0]
			for f := 0; f < int(frames); f++ {
				src := top
				if f == 0 {
					src = bot
				}
				step := miterEval(t, src, []bool{in[1+2*f], in[2+2*f], state})
				if out[f] != step[0] {
					t.Fatal("frame output")
				}
				state = step[1]
			}
			if out[len(out)-1] != state {
				t.Fatal("final state")
			}
		}
	}
}
func TestDemiterRoundTrip(t *testing.T) {
	a, b := miterFixture(false), miterFixture(true)
	m := Saig_ManCreateMiter(a, b, 0)
	var left, right *aig.Graph
	if Saig_ManDemiterSimpleDiff(m, &left, &right) != 1 {
		t.Fatal("failed demiter")
	}
	for x := 0; x < 16; x++ {
		in := []bool{x&1 != 0, x&2 != 0, x&4 != 0, x&8 != 0}
		l := miterEval(t, left, in[:3])
		r := miterEval(t, right, []bool{in[0], in[1], in[3]})
		want := miterEval(t, m, in)
		if (l[0] != r[0]) != want[0] || l[1] != want[1] || r[1] != want[2] {
			t.Fatalf("demiter assignment%d", x)
		}
	}
	for _, n := range m.Nodes() {
		if n.MarkA || n.MarkB {
			t.Fatal("marks leaked")
		}
	}
}
func TestDualRailAllInputCodes(t *testing.T) {
	g := aig.NewGraph()
	a, b := g.NewInput(), g.NewInput()
	g.AddOutput(g.AND(a, b.Not()))
	d := Saig_ManDualRail(g, 0)
	for x := 0; x < 16; x++ {
		in := []bool{x&1 != 0, x&2 != 0, x&4 != 0, x&8 != 0}
		out := miterEval(t, d, in)
		a0, a1, b0, b1 := in[0], in[1], in[3], in[2]
		want0 := a0 && !a1 || b0 && !b1
		want1 := !a0 && a1 && !b0 && b1
		if out[0] != want0 || out[1] != want1 {
			t.Fatalf("dual rails assignment%d", x)
		}
	}
}
func TestSpecialProofAndResimulation(t *testing.T) {
	a, b := miterFixture(false), miterFixture(false)
	if Ssw_SecSpecial(a, b, 2, 0) != 1 {
		t.Fatal("equal machines differ")
	}
	if Ssw_SecSpecial(a, miterFixture(true), 2, 0) != 0 {
		t.Fatal("different machines equal")
	}
	g := aig.NewGraph()
	x := g.NewInput()
	g.AddOutput(x)
	g.AddOutput(x.Not())
	g.AddOutput(g.Const1())
	aig.Aig_ManSetRegNum(g, 0)
	var count int32
	if first := Ssw_SecCexResimulate(g, []int32{0}, &count); first != 1 || count != 2 {
		t.Fatalf("first%d count%d", first, count)
	}
	s := Sec_MiterStatus(g)
	if s.nSat != 3 || s.iOut != 0 {
		t.Fatal("quick status")
	}
}

func TestDemiterDualAndRejection(t *testing.T) {
	g := aig.NewGraph()
	a, b, q, r := g.NewInput(), g.NewInput(), g.NewInput(), g.NewInput()
	g.AddOutput(g.AND(a, q))
	g.AddOutput(g.AND(b, r))
	g.AddOutput(q.Not())
	g.AddOutput(r.Not())
	aig.Aig_ManSetRegNum(g, 2)
	var left, right *aig.Graph
	if Saig_ManDemiterDual(g, &left, &right) != 1 {
		t.Fatal("dual split")
	}
	if Saig_ManRegNum(left) != 1 || Saig_ManRegNum(right) != 1 || Saig_ManPoNum(left) != 1 || Saig_ManPoNum(right) != 1 {
		t.Fatal("sequential cleanup failed")
	}
	for x := 0; x < 8; x++ {
		in := []bool{x&1 != 0, x&2 != 0, x&4 != 0}
		l, r := miterEval(t, left, in), miterEval(t, right, in)
		if l[0] != (in[0] && in[2]) || r[0] != (in[1] && in[2]) || l[1] != !in[2] || r[1] != !in[2] {
			t.Fatal("split behavior")
		}
	}
	if Saig_ManDemiterSimpleDiff(g, &left, &right) != 0 {
		t.Fatal("non-XOR outputs accepted")
	}
	for _, n := range g.Nodes() {
		if n.MarkA || n.MarkB {
			t.Fatal("failure leaked marks")
		}
	}
	c := aig.NewGraph()
	c.NewInput()
	c.AddOutput(c.Const0())
	if Saig_ManDemiterSimpleDiff(c, &left, &right) != 0 {
		t.Fatal("combinational graph accepted")
	}
}

func TestGeneralDemiterPreservesOriginalFailure(t *testing.T) {
	a, b := miterFixture(false), miterFixture(true)
	g := Saig_ManCreateMiter(a, b, 0)
	if Saig_ManDemiter(g, nil, nil) != 1 {
		t.Fatal("diagnostic labels")
	}
	defer func() {
		if recover() == nil {
			t.Fatal("unfinished original graph construction should fail explicitly")
		}
	}()
	var result *aig.Graph
	Saig_ManDemiter(g, &result, nil)
}
