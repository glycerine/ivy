package aig

import (
	"github.com/glycerine/abcgo/misc/tim"
	"testing"
)

func TestVerifyTopoOrderRejectsForwardReferencesInValidDAG(t *testing.T) {
	g := NewGraph()
	a, b, c := g.NewInput(), g.NewInput(), g.NewInput()
	early := Aig_ObjCreate(g, &Node{Kind: Buffer, Fanins: [2]Edge{a, InvalidEdge}})
	later := g.AND(b, c)
	Aig_ObjConnect(g, early, later, InvalidEdge)
	g.AddOutput(early)
	if err := g.Check(); err != nil {
		t.Fatal(err)
	}
	if Aig_ManVerifyTopoOrder(g) != 0 {
		t.Fatal("accepted a valid DAG whose buffer precedes its fanin in object order")
	}
}

func TestVerifyTopoOrderChecksTimingBoxDependencies(t *testing.T) {
	for _, ordered := range []bool{false, true} {
		g := NewGraph()
		a := g.NewInput()
		var box Edge
		if ordered {
			g.AddOutput(a)
			box = g.NewInput()
		} else {
			box = g.NewInput()
			g.AddOutput(a)
		}
		g.AddOutput(g.AND(a, box))
		tm := tim.Tim_ManStart(2, 2)
		tim.Tim_ManCreateBox(tm, 0, 1, 1, 1, -1, 0)
		g.Timing = tm
		want := int32(0)
		if ordered {
			want = 1
		}
		if got := Aig_ManVerifyTopoOrder(g); got != want {
			t.Fatalf("ordered=%v got=%d want=%d", ordered, got, want)
		}
	}
}
