package ssw

import (
	"github.com/glycerine/abcgo/aig/aig"
	"github.com/glycerine/abcgo/misc/util"
	"github.com/glycerine/abcgo/sat/bsat"
	"testing"
)

func TestSswCnfTruthTables(t *testing.T) {
	for _, flip := range []int32{0, 1} {
		for _, kind := range []string{"and", "mux", "xor"} {
			g := aig.NewGraph()
			a, b, c := g.NewInput(), g.NewInput(), g.NewInput()
			aig.Aig_ObjSetPhase(g, a, true)
			var root aig.Edge
			switch kind {
			case "and":
				root = g.AND(a, g.AND(b, c.Not()))
			case "mux":
				root = g.Mux(a, b, c.Not())
			case "xor":
				root = g.NewXor(a, b)
			}
			g.AddOutput(root)
			p := Ssw_SatStart(flip)
			p.pAig = g
			for _, e := range []aig.Edge{a, b, c} {
				Ssw_CnfNodeAddToSolver(p, e)
			}
			Ssw_CnfNodeAddToSolver(p, aig.Aig_Regular(root))
			for bits := 0; bits < 16; bits++ {
				input := []bool{bits&1 != 0, bits&2 != 0, bits&4 != 0}
				want, err := g.Evaluate(input)
				if err != nil {
					t.Fatal(err)
				}
				assume := []int32{}
				for i, e := range []aig.Edge{a, b, c, root} {
					value := bits&(1<<i) != 0
					assume = append(assume, sswLit(p, e, !value))
				}
				status := p.pSat.Solve(assume...)
				if (status == bsat.Satisfiable) != (want[0] == (bits&8 != 0)) {
					t.Fatalf("flip=%d %s bits=%04b status=%v want output=%v", flip, kind, bits, status, want[0])
				}
			}
		}
	}
}
func TestSswBmcDynamicFindsCounterexample(t *testing.T) {
	g := aig.NewGraph()
	pi, lo := g.NewInput(), g.NewInput()
	g.AddOutput(lo)
	g.AddOutput(pi)
	aig.Aig_ManSetRegNum(g, 1)
	last := int32(-1)
	if got := Ssw_BmcDynamic(g, 3, 0, 0, &last); got != 0 || last != 1 {
		t.Fatalf("result=%d frame=%d", got, last)
	}
	c := g.SequentialCounterexample()
	if c == nil || c.IFrame != 1 || !util.Abc_InfoHasBit(c.PData, 1) {
		t.Fatalf("wrong counterexample: %+v", c)
	}
}
func TestSswBmcDynamicProvesBoundedSafe(t *testing.T) {
	g := aig.NewGraph()
	pi := g.NewInput()
	g.NewInput()
	g.AddOutput(g.Const0())
	g.AddOutput(pi)
	aig.Aig_ManSetRegNum(g, 1)
	if got := Ssw_BmcDynamic(g, 3, 0, 0, nil); got != 1 {
		t.Fatalf("result=%d", got)
	}
}

func TestSswCnfBufferIsConstrained(t *testing.T) {
	g := aig.NewGraph()
	a := g.NewInput()
	b := aig.Aig_ObjCreate(g, &aig.Node{Kind: aig.Buffer, Fanins: [2]aig.Edge{a.Not(), aig.InvalidEdge}})
	p := Ssw_SatStart(0)
	p.pAig = g
	Ssw_CnfNodeAddToSolver(p, b)
	for _, v := range []bool{false, true} {
		for _, out := range []bool{false, true} {
			status := p.pSat.Solve(sswLit(p, a, !v), sswLit(p, b, !out))
			if (status == bsat.Satisfiable) != (out != v) {
				t.Fatalf("input=%v output=%v status=%v", v, out, status)
			}
		}
	}
}
func TestSswBmcUnrollBuffer(t *testing.T) {
	g := aig.NewGraph()
	a := g.NewInput()
	g.NewInput()
	b := aig.Aig_ObjCreate(g, &aig.Node{Kind: aig.Buffer, Fanins: [2]aig.Edge{a.Not(), aig.InvalidEdge}})
	o := g.AddOutput(b)
	g.AddOutput(a)
	aig.Aig_ManSetRegNum(g, 1)
	p := Ssw_FrmStart(g)
	defer Ssw_FrmStop(p)
	p.pFrames = aig.NewGraph()
	result := Ssw_BmcUnroll_rec(p, o, 0)
	p.pFrames.AddOutput(result)
	got, err := p.pFrames.Evaluate([]bool{true})
	if err != nil || got[0] {
		t.Fatalf("output=%v error=%v", got, err)
	}
}
