package csw

import (
	"github.com/glycerine/abcgo/aig/aig"
	"math/rand"
	"testing"
)

func compareSweep(t *testing.T, g *aig.Graph, cuts, leaves int32) {
	t.Helper()
	out := Csw_Sweep(g, cuts, leaves, 0)
	if err := out.Check(); err != nil {
		t.Fatal(err)
	}
	n := len(g.Inputs())
	for assignment := 0; assignment < 1<<n; assignment++ {
		in := make([]bool, n)
		for i := range in {
			in[i] = assignment&(1<<i) != 0
		}
		want, err := g.Evaluate(in)
		if err != nil {
			t.Fatal(err)
		}
		got, err := out.Evaluate(in)
		if err != nil {
			t.Fatal(err)
		}
		if len(want) != len(got) {
			t.Fatal("output count")
		}
		for i := range want {
			if want[i] != got[i] {
				t.Fatalf("cuts=%d leaves=%d assignment=%d output=%d", cuts, leaves, assignment, i)
			}
		}
	}
}

func TestSweepRandomExhaustive(t *testing.T) {
	r := rand.New(rand.NewSource(726))
	for sample := 0; sample < 100; sample++ {
		g := aig.NewGraph()
		edges := []aig.Edge{g.Const1(), g.Const0()}
		for i := 0; i < 5; i++ {
			edges = append(edges, g.NewInput())
		}
		for i := 0; i < 50; i++ {
			a, b := edges[r.Intn(len(edges))], edges[r.Intn(len(edges))]
			if r.Intn(2) == 0 {
				a = a.Not()
			}
			if r.Intn(2) == 0 {
				b = b.Not()
			}
			var e aig.Edge
			if sample%3 == 0 && i%5 == 0 {
				e = g.NewXor(a, b)
			} else {
				e = g.AND(a, b)
			}
			edges = append(edges, e)
		}
		for i := 0; i < 8; i++ {
			g.AddOutput(edges[r.Intn(len(edges))])
		}
		for _, k := range []int32{2, 3, 5, 6} {
			compareSweep(t, g, 4, k)
		}
	}
}

func TestSweepReducesAssociativeDuplicates(t *testing.T) {
	g := aig.NewGraph()
	a, b, c := g.NewInput(), g.NewInput(), g.NewInput()
	x := g.AND(g.AND(a, b), c)
	y := g.AND(a, g.AND(b, c))
	g.AddOutput(x)
	g.AddOutput(y)
	out := Csw_Sweep(g, 8, 6, 0)
	compareSweep(t, g, 8, 6)
	if aig.Aig_ManNodeNum(out) >= aig.Aig_ManNodeNum(g) {
		t.Fatal("equivalent cuts not merged")
	}
}

func TestWideTruthAndBuffer(t *testing.T) {
	g := aig.NewGraph()
	var inputs []aig.Edge
	for i := 0; i < 16; i++ {
		inputs = append(inputs, g.NewInput())
	}
	a := inputs[0]
	for _, e := range inputs[1:] {
		a = g.AND(a, e)
	}
	b := inputs[15]
	for i := 14; i >= 0; i-- {
		b = g.AND(inputs[i], b)
	}
	buffer := aig.Aig_ObjCreate(g, &aig.Node{Kind: aig.Buffer, Fanins: [2]aig.Edge{a, aig.InvalidEdge}})
	g.AddOutput(a)
	g.AddOutput(b.Not())
	g.AddOutput(buffer)
	out := Csw_Sweep(g, 8, 16, 0)
	if err := out.Check(); err != nil {
		t.Fatal(err)
	}
	if aig.Aig_ManNodeNum(out) >= aig.Aig_ManNodeNum(g) {
		t.Fatal("wide equivalent cuts did not merge")
	}
	for assignment := 0; assignment < 1<<16; assignment++ {
		in := make([]bool, 16)
		for i := range in {
			in[i] = assignment&(1<<i) != 0
		}
		want, _ := g.Evaluate(in)
		got, _ := out.Evaluate(in)
		for i := range want {
			if got[i] != want[i] {
				t.Fatalf("assignment %d", assignment)
			}
		}
	}
}

func TestCutMergeDominanceAndPhase(t *testing.T) {
	p := &Csw_Man_t{nLeafMax: 6}
	for a := 1; a < 64; a++ {
		for b := 1; b < 64; b++ {
			x, y := &Csw_Cut_t{}, &Csw_Cut_t{}
			for i := 0; i < 6; i++ {
				if a&(1<<i) != 0 {
					x.pFanins = append(x.pFanins, int32(i))
					x.uSign |= 1 << i
				}
				if b&(1<<i) != 0 {
					y.pFanins = append(y.pFanins, int32(i))
					y.uSign |= 1 << i
				}
			}
			c := &Csw_Cut_t{}
			if Csw_CutMerge(p, x, y, c) != 1 {
				t.Fatal("valid union rejected")
			}
			var mask int
			for _, id := range c.pFanins {
				mask |= 1 << id
			}
			if mask != a|b {
				t.Fatal("union differs")
			}
			if (Csw_CutCheckDominance(x, y) != 0) != (a&b == a) {
				t.Fatal("dominance differs")
			}
			phase := Cut_TruthPhase(c, x)
			var want uint32
			for i, id := range c.pFanins {
				if a&(1<<id) != 0 {
					want |= 1 << i
				}
			}
			if phase != want {
				t.Fatal("truth placement differs")
			}
		}
	}
	p.nLeafMax = 2
	c := &Csw_Cut_t{}
	if Csw_CutMerge(p, &Csw_Cut_t{pFanins: []int32{1, 3}}, &Csw_Cut_t{pFanins: []int32{2, 3}}, c) != 0 || len(c.pFanins) != 0 {
		t.Fatal("overfull cut retained")
	}
}

func TestHashCollisionAndOppositePhase(t *testing.T) {
	p := &Csw_Man_t{pTable: make([]*Csw_Cut_t, 1)}
	a := &Csw_Cut_t{iNode: 9, pFanins: []int32{1, 2, 3}, uSign: 14, truth: []uint32{0x96969696}}
	Csw_TableCutInsert(p, a)
	other := &Csw_Cut_t{iNode: 10, pFanins: []int32{1, 2, 35}, uSign: 14, truth: []uint32{0x96969696}}
	if Csw_TableCutLookup(p, other) != aig.InvalidEdge {
		t.Fatal("signature collision mistaken for same support")
	}
	same := &Csw_Cut_t{pFanins: []int32{1, 2, 3}, uSign: 14, truth: []uint32{0x69696969}}
	if Csw_TableCutLookup(p, same) != aig.Edge(19) {
		t.Fatal("opposite truth phase lost")
	}
	same.truth[0] = 0x80808080
	if Csw_TableCutLookup(p, same) != aig.InvalidEdge {
		t.Fatal("different truth matched")
	}
	if Csw_TableCountCuts(p) != 1 {
		t.Fatal("bucket count")
	}
}

func TestTwoVariableResubstitution(t *testing.T) {
	for _, truth := range []uint32{1, 2, 4, 8, 14, 13, 11, 7, 6, 9} {
		g := aig.NewGraph()
		a, b := g.NewInput(), g.NewInput()
		p := Csw_ManStart(g, 4, 3, 0)
		repeated := truth * 0x11111111
		c := &Csw_Cut_t{pFanins: []int32{int32(a.NodeID()), int32(b.NodeID())}, truth: []uint32{repeated}}
		e := Csw_ObjTwoVarCut(p, c)
		if truth == 6 || truth == 9 {
			if e != aig.InvalidEdge {
				t.Fatal("XOR incorrectly resubstituted as AND")
			}
			continue
		}
		p.pManRes.AddOutput(e)
		for q := 0; q < 4; q++ {
			got, err := p.pManRes.Evaluate([]bool{q&1 != 0, q&2 != 0})
			if err != nil {
				t.Fatal(err)
			}
			if got[0] != (truth&(1<<q) != 0) {
				t.Fatalf("truth %x assignment %d", truth, q)
			}
		}
	}
}
