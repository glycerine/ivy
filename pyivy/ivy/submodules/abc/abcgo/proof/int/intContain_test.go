package int

import (
	"testing"

	"github.com/glycerine/abcgo/aig/aig"
)

func unaryGraph(invert bool) *aig.Graph {
	g := aig.NewGraph()
	x := g.NewInput()
	if invert {
		x = x.Not()
	}
	g.AddOutput(x)
	return g
}

func TestContainmentAndEquivalence(t *testing.T) {
	x, notX := unaryGraph(false), unaryGraph(true)
	zero := aig.NewGraph()
	zero.NewInput()
	zero.AddOutput(zero.Const0())
	one := aig.NewGraph()
	one.NewInput()
	one.AddOutput(one.Const1())
	for _, test := range []struct {
		name                    string
		a, b                    *aig.Graph
		containment, equivalent int32
	}{
		{"same", x, x, 1, 1},
		{"complements", x, notX, 0, 0},
		{"empty subset", zero, x, 1, 0},
		{"nonempty not subset", one, x, 0, 0},
	} {
		t.Run(test.name, func(t *testing.T) {
			if got := Inter_ManCheckContainment(test.a, test.b); got != test.containment {
				t.Fatalf("containment %d want %d", got, test.containment)
			}
			if got := Inter_ManCheckEquivalence(test.a, test.b); got != test.equivalent {
				t.Fatalf("equivalence %d want %d", got, test.equivalent)
			}
		})
	}
}

func TestLatchFramesAndInductiveContainment(t *testing.T) {
	for _, toggle := range []bool{false, true} {
		g := aig.NewGraph()
		x, lo := g.NewInput(), g.NewInput()
		g.AddOutput(x)
		next := lo
		if toggle {
			next = next.Not()
		}
		g.AddOutput(next)
		aig.Aig_ManSetRegNum(g, 1)
		var mapping []aig.Edge
		frames := Inter_ManFramesLatches(g, 2, &mapping)
		if len(mapping) != 3 || frames.InputCount() != 3 {
			t.Fatalf("frames inputs=%d mapping=%v", frames.InputCount(), mapping)
		}
		for _, e := range mapping {
			frames.AddOutput(e)
		}
		for mask := 0; mask < 8; mask++ {
			initial := mask&1 != 0
			got, err := frames.Evaluate([]bool{initial, mask&2 != 0, mask&4 != 0})
			want1 := initial != toggle
			if err != nil || len(got) != 3 || got[0] != initial || got[1] != want1 || got[2] != initial {
				t.Fatalf("toggle=%v mask=%d got=%v err=%v", toggle, mask, got, err)
			}
		}
		want := int32(1)
		if toggle {
			want = 0
		}
		for _, backwards := range []int32{0, 1} {
			if got := Inter_ManCheckInductiveContainment(g, unaryGraph(false), 1, backwards); got != want {
				t.Fatalf("toggle=%v backwards=%d got=%d want=%d", toggle, backwards, got, want)
			}
		}
	}
}

func TestOriginalUniquenessLimit(t *testing.T) {
	if Inter_ManCheckUniqueness(nil, nil, nil, 1) != 1 {
		t.Fatal("single frame")
	}
	defer func() {
		if recover() == nil {
			t.Error("expected original multi-frame assertion")
		}
	}()
	Inter_ManCheckUniqueness(nil, nil, nil, 2)
}
