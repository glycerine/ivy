package ssw

import (
	"github.com/glycerine/abcgo/aig/aig"
	"github.com/glycerine/abcgo/misc/vec"
	"testing"
)

func TestClassesFromRepresentatives(t *testing.T) {
	g := aig.NewGraph()
	a, b, c, d := g.NewInput(), g.NewInput(), g.NewInput(), g.NewInput()
	aig.Aig_ManReprStart(g, 5)
	aig.Aig_ObjSetRepr(g, b, a)
	aig.Aig_ObjSetRepr(g, c, a)
	aig.Aig_ObjSetRepr(g, d, g.Const1())
	p := Ssw_ClassesPrepareFromReprs(g)
	var size int32
	class := Ssw_ClassesReadClass(p, a, &size)
	if size != 3 || class[0] != a || class[1] != b || class[2] != c || p.nCands1 != 1 || p.nLits != 2 {
		t.Fatal("wrong reconstructed classes")
	}
}
func TestPrepareLatchCandidates(t *testing.T) {
	g := aig.NewGraph()
	pi, lo := g.NewInput(), g.NewInput()
	gate := g.AND(pi, lo)
	g.AddOutput(gate)
	aig.Aig_ManSetRegNum(g, 1)
	p := Ssw_ClassesPrepareSimple(g, 1, 0)
	if p.nCands1 != 1 || aig.Aig_ObjRepr(g, lo) != g.Const1() || aig.Aig_ObjRepr(g, gate) != aig.InvalidEdge {
		t.Fatal("wrong latch candidates")
	}
}

func TestPrepareSimpleRespectsSourceLevelAndLatchOptions(t *testing.T) {
	for _, tc := range []struct {
		name      string
		latchOnly int32
		wantGate  bool
	}{
		{"all candidates limited by level", 0, true},
		{"latch only ignores level", 1, false},
	} {
		t.Run(tc.name, func(t *testing.T) {
			g := aig.NewGraph()
			x, y, latch := g.NewInput(), g.NewInput(), g.NewInput()
			level1 := g.AND(x, latch)
			level2 := g.AND(level1, y)
			g.AddOutput(level2)
			g.AddOutput(x)
			aig.Aig_ManSetRegNum(g, 1)
			p := Ssw_ClassesPrepareSimple(g, tc.latchOnly, 1)
			defer Ssw_ClassesStop(p)
			if aig.Aig_ObjRepr(g, latch) != g.Const1() {
				t.Fatal("latch was not selected")
			}
			if (aig.Aig_ObjRepr(g, level1) == g.Const1()) != tc.wantGate {
				t.Fatal("level-one gate selection disagrees with C")
			}
			if aig.Aig_ObjRepr(g, level2) != aig.InvalidEdge || aig.Aig_ObjRepr(g, x) != aig.InvalidEdge {
				t.Fatal("deep gate or primary input was selected")
			}
			want := int32(1)
			if tc.wantGate {
				want++
			}
			if p.nCands1 != want {
				t.Fatalf("candidate count = %d, want %d", p.nCands1, want)
			}
		})
	}
}

func TestPrepareClassesFromPairs(t *testing.T) {
	g := aig.NewGraph()
	a, b, c := g.NewInput(), g.NewInput(), g.NewInput()
	constants := vec.Vec_IntAlloc(2)
	vec.Vec_IntPush(constants, 0)
	vec.Vec_IntPush(constants, int32(c.NodeID()))
	group := vec.Vec_IntAlloc(2)
	vec.Vec_IntPush(group, int32(a.NodeID()))
	vec.Vec_IntPush(group, int32(b.NodeID()))
	classes := make([]*vec.Vec_Int_t, len(g.Nodes()))
	classes[0] = constants
	classes[a.NodeID()] = group
	p := Ssw_ClassesPreparePairs(g, classes)
	if p.nCands1 != 1 || p.nClasses != 1 || aig.Aig_ObjRepr(g, b) != a || aig.Aig_ObjRepr(g, c) != g.Const1() {
		t.Fatal("wrong classes")
	}
}

func TestPrepareSequentialSimulationClasses(t *testing.T) {
	g := aig.NewGraph()
	pi, lo := g.NewInput(), g.NewInput()
	gate := g.AND(pi, lo)
	g.AddOutput(gate)
	g.AddOutput(pi)
	aig.Aig_ManSetRegNum(g, 1)
	p := Ssw_ClassesPrepare(g, 1, 1, 0, 0, 0, 0)
	Ssw_ClassesCheck(p)
	if p.pAig != g {
		t.Fatal("graph not retained")
	}
}

func TestOutputOnlyClassPreparationExcludesOtherNodes(t *testing.T) {
	g := aig.NewGraph()
	pi, lo := g.NewInput(), g.NewInput()
	gate := g.AND(pi, lo)
	g.AddOutput(gate)
	g.AddOutput(g.Const0())
	aig.Aig_ManSetRegNum(g, 1)
	classes := Ssw_ClassesPrepare(g, 1, 0, 0, 1, 0, 0)
	defer Ssw_ClassesStop(classes)
	if aig.Aig_ObjRepr(g, gate) != g.Const1() {
		t.Fatal("constant output driver not selected")
	}
	if aig.Aig_ObjRepr(g, lo) != aig.InvalidEdge || aig.Aig_ObjRepr(g, pi) != aig.InvalidEdge {
		t.Fatal("non-output input selected")
	}
}
