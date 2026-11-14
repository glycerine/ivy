package aig

import (
	"slices"
	"testing"
)

func TestReverseDfsUsesSourceFanoutOrderAndMarks(t *testing.T) {
	g := NewGraph()
	a, b, c, d := g.NewInput(), g.NewInput(), g.NewInput(), g.NewInput()
	root := g.AND(a, b)
	left, right := g.AND(root, c), g.AND(root, d)
	g.AddOutput(left)
	g.AddOutput(right)
	dangling := g.AND(a, c)
	Aig_ManFanoutStart(g)
	defer Aig_ManFanoutStop(g)
	Aig_ObjRemoveFanoutPin(g, root, left, 1)
	Aig_ObjAddFanoutPin(g, root, left, 1)
	old := g.travID
	want := []Edge{right, left, root, dangling}
	if got := Aig_ManDfsReverse(g); !slices.Equal(got, want) {
		t.Errorf("reverse DFS = %v, want %v", got, want)
	}
	if g.travID != old+1 {
		t.Error("reverse DFS did not start a traversal")
	}
	for _, e := range append(want, g.outputs...) {
		if g.node(e).TravID != g.travID {
			t.Errorf("object %d lacks source traversal mark", e)
		}
	}
	if g.node(a).TravID == g.travID || g.node(g.Const1()).TravID == g.travID {
		t.Error("reverse DFS marked a CI or constant")
	}
	Aig_ManIncrementTravId(g)
	for _, e := range append([]Edge{right}, g.outputs...) {
		g.node(e).TravID = g.travID
	}
	var got []Edge
	Aig_ManDfsReverse_rec(g, root, &got)
	Aig_ManDfsReverse_rec(g, root, &got)
	if !slices.Equal(got, []Edge{left, root}) {
		t.Errorf("seeded reverse DFS = %v, want %v", got, []Edge{left, root})
	}
}

func TestReverseDfsRequiresSourceFanoutAllocation(t *testing.T) {
	g := NewGraph()
	a, b := g.NewInput(), g.NewInput()
	g.AddOutput(g.AND(a, b))
	defer func() {
		if recover() == nil {
			t.Fatal("reverse DFS accepted a missing source fanout allocation")
		}
	}()
	Aig_ManDfsReverse(g)
}

func TestStartReverseLevelsRequiresSourceSetup(t *testing.T) {
	for _, marked := range []bool{false, true} {
		func() {
			g := NewGraph()
			a, b := g.NewInput(), g.NewInput()
			root := g.AND(a, b)
			g.AddOutput(root)
			if marked {
				Aig_ManFanoutStart(g)
				defer Aig_ManFanoutStop(g)
				g.node(root).MarkA = true
			}
			defer func() {
				if recover() == nil {
					t.Errorf("marked=%v: reverse timing ignored source setup precondition", marked)
				}
			}()
			Aig_ManStartReverseLevels(g, 0)
		}()
	}
}

func TestReverseTimingUsesOutputDriverLevels(t *testing.T) {
	g := NewGraph()
	a, b := g.NewInput(), g.NewInput()
	root := g.AND(a, b)
	out := g.AddOutput(root)
	Aig_ObjSetLevel(g, out, 23)
	Aig_ManFanoutStart(g)
	defer Aig_ManFanoutStop(g)
	Aig_ManStartReverseLevels(g, 2)
	defer Aig_ManStopReverseLevels(g)
	if g.maxRequiredLevel != 3 {
		t.Fatalf("level bound = %d, want driver level 1 plus allowance 2", g.maxRequiredLevel)
	}
}

func TestLevelNumUsesSourceOutputDriverLevels(t *testing.T) {
	g := NewGraph()
	a, b := g.NewInput(), g.NewInput()
	root := g.AND(a, b)
	out := g.AddOutput(root)
	Aig_ObjSetLevel(g, out, 23)
	if got := Aig_ManLevelNum(g); got != 1 {
		t.Fatalf("manager level = %d, want output driver level 1", got)
	}
}
