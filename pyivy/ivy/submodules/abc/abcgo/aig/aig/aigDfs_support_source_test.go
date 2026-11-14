package aig

import (
	"slices"
	"testing"
)

func TestSupportNodesSourceClearsResultAndMarksCone(t *testing.T) {
	g := NewGraph()
	a, b, unused := g.NewInput(), g.NewInput(), g.NewInput()
	root := g.AND(a, b)
	out := g.AddOutput(root.Not())
	old := g.travID
	got := []Edge{unused}
	Aig_SupportNodes(g, []Edge{out, a, out, unused}, 3, &got)
	if !slices.Equal(got, []Edge{a, b}) {
		t.Errorf("support = %v, want %v", got, []Edge{a, b})
	}
	if g.travID != old+1 {
		t.Error("support collection did not start a traversal")
	}
	for _, e := range []Edge{g.Const1(), a, b, root} {
		if g.node(e).TravID != g.travID {
			t.Errorf("cone object %d not marked", e)
		}
	}
	if g.node(unused).TravID == g.travID || g.node(out).TravID == g.travID {
		t.Error("support traversal marked an unused CI or CO")
	}
	Aig_SupportNodes(g, nil, 0, &got)
	if len(got) != 0 {
		t.Errorf("empty roots retained previous support: %v", got)
	}
}

func TestSupportRecursiveHelpersRespectSeededCones(t *testing.T) {
	for _, countOnly := range []bool{false, true} {
		g := NewGraph()
		a, b, c := g.NewInput(), g.NewInput(), g.NewInput()
		barrier := g.AND(a, b)
		root := g.AND(barrier, c)
		Aig_ManIncrementTravId(g)
		g.node(barrier).TravID = g.travID
		if countOnly {
			count := int32(7)
			Aig_SupportSize_rec(g, root, &count)
			Aig_SupportSize_rec(g, root, &count)
			if count != 8 {
				t.Errorf("seeded support count = %d, want 8", count)
			}
		} else {
			got := []Edge{a}
			Aig_Support_rec(g, root, &got)
			Aig_Support_rec(g, root, &got)
			if !slices.Equal(got, []Edge{a, c}) {
				t.Errorf("seeded recursive support = %v, want %v", got, []Edge{a, c})
			}
		}
		if g.node(root).TravID != g.travID || g.node(c).TravID != g.travID || g.node(a).TravID == g.travID || g.node(b).TravID == g.travID {
			t.Error("recursive traversal ignored the seeded gate boundary")
		}
	}
}

func TestSupportSizeTestCountsSmallSupportGates(t *testing.T) {
	g := NewGraph()
	inputs := make([]Edge, 17)
	for i := range inputs {
		inputs[i] = g.NewInput()
	}
	root := inputs[0]
	for _, e := range inputs[1:] {
		root = g.AND(root, e)
	}
	g.AddOutput(root)
	g.AddOutput(root)
	// Include a dangling two-input gate, as C iterates all live logic nodes.
	g.AND(inputs[0], inputs[16])
	if got := Aig_SupportSizeTest(g); got != 16 {
		t.Fatalf("small-support nodes = %d, want 16", got)
	}
}

func TestSupportWrappersRequireRegularNonOutputRoots(t *testing.T) {
	for _, sizeOnly := range []bool{false, true} {
		for _, output := range []bool{false, true} {
			func() {
				g := NewGraph()
				a := g.NewInput()
				root := a.Not()
				if output {
					root = g.AddOutput(a)
				}
				defer func() {
					if recover() == nil {
						t.Errorf("sizeOnly=%v output=%v: accepted root forbidden by C", sizeOnly, output)
					}
				}()
				if sizeOnly {
					Aig_SupportSize(g, root)
				} else {
					Aig_Support(g, root)
				}
			}()
		}
	}
}

func TestSupportWrappersHandleBuffersXorsAndConstants(t *testing.T) {
	g := NewGraph()
	a, b, c, unused := g.NewInput(), g.NewInput(), g.NewInput(), g.NewInput()
	gate := g.NewXor(g.AND(a, b), c)
	buffer := Aig_ObjCreate(g, &Node{Kind: Buffer, Fanins: [2]Edge{gate.Not(), InvalidEdge}})
	for run := 0; run < 2; run++ {
		old := g.travID
		// Structural hashing orders the lower-ID CI c before the AND gate.
		if got := Aig_Support(g, buffer); !slices.Equal(got, []Edge{c, a, b}) {
			t.Fatalf("buffer support = %v, want %v", got, []Edge{c, a, b})
		}
		if g.travID != old+1 || g.node(unused).TravID == g.travID || g.node(gate).TravID != g.travID {
			t.Fatal("support wrapper did not preserve source traversal scope")
		}
		if got := Aig_SupportSize(g, buffer); got != 3 {
			t.Fatalf("buffer support size = %d, want 3", got)
		}
	}
	if got := Aig_Support(g, g.Const1()); len(got) != 0 || g.node(g.Const1()).TravID != g.travID {
		t.Fatal("constant support should be empty with the constant marked")
	}
}

func TestSupportRecUsesMarksRatherThanResultMembership(t *testing.T) {
	g := NewGraph()
	a, b := g.NewInput(), g.NewInput()
	root := g.AND(a, b)
	Aig_ManIncrementTravId(g)
	got := []Edge{b}
	Aig_Support_rec(g, root, &got)
	if !slices.Equal(got, []Edge{b, a, b}) {
		t.Fatalf("recursive result = %v, want source append semantics", got)
	}
}
