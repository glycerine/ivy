package aig

import (
	"fmt"
	"math/rand/v2"
	"reflect"
	"strings"
	"testing"
)

func TestSATLevelPartitionsIncludeTerminalPositions(t *testing.T) {
	g := NewGraph()
	a, b := g.NewInput(), g.NewInput()
	g.AddOutput(g.AND(a, b))
	// Reverse level order includes the CO, then the AND, then both CIs and
	// the constant. Terminals consume partition positions just as gates do.
	if got, want := Aig_ManPartitionLevelized(g, 1), []int32{4, 3, 2, 1, 0}; !reflect.DeepEqual(got, want) {
		t.Fatalf("partition map %v, want %v", got, want)
	}
}

func TestSATPartitionLabelsMatchOriginalC(t *testing.T) {
	rng := rand.New(rand.NewPCG(270, 182))
	var input, want strings.Builder
	fmt.Fprintln(&input, 32)
	for trial := 0; trial < 32; trial++ {
		g := NewGraph()
		pool := []Edge{g.Const1(), g.NewInput(), g.NewInput(), g.NewInput()}
		for step := 0; step < 12; step++ {
			a, b := rng.IntN(len(pool)*2), rng.IntN(len(pool)*2)
			fmt.Fprintln(&input, a, b)
			pool = append(pool, g.AND(dupPhase(pool[a/2], a&1 != 0), dupPhase(pool[b/2], b&1 != 0)))
		}
		for step := 0; step < 4; step++ {
			a := rng.IntN(len(pool) * 2)
			if step == 0 {
				// Levelize requires all live objects to fit the output level range.
				for i, e := range pool {
					if g.Node(e).Level > g.Node(pool[a/2]).Level {
						a = i * 2
					}
				}
			}
			fmt.Fprintln(&input, a)
			g.AddOutput(dupPhase(pool[a/2], a&1 != 0))
		}
		Aig_ManCleanup(g)
		for _, size := range []int32{1, 2, 3, 7} {
			for mode := 0; mode < 4; mode++ {
				var labels []int32
				switch mode {
				case 0:
					labels = Aig_ManPartitionMonolithic(g)
				case 1:
					labels = Aig_ManPartitionLevelized(g, size)
				case 2:
					labels = Aig_ManPartitionDfs(g, size, 0)
				case 3:
					labels = Aig_ManPartitionDfs(g, size, 1)
				}
				fmt.Fprintf(&want, "case %d size %d mode %d:", trial, size, mode)
				for _, label := range labels {
					fmt.Fprintf(&want, " %d", label)
				}
				fmt.Fprintln(&want)
			}
		}
	}
	got := runAigSourceFixtureInput(t, "sat_partition_labels_compare.c", input.String(), "aigMan.c", "aigObj.c", "aigOper.c", "aigTable.c", "aigMem.c", "aigFanout.c", "aigUtil.c", "aigPartSat.c", "aigDfs.c", "../../misc/tim/timMan.c", "../../misc/mem/mem.c")
	if got != want.String() {
		c, goLines := strings.Split(got, "\n"), strings.Split(want.String(), "\n")
		for i := range c {
			if i >= len(goLines) {
				t.Fatalf("missing Go result row %d: C %s", i, c[i])
			}
			if c[i] != goLines[i] {
				t.Fatalf("first source mismatch at row %d:\nC  %s\nGo %s", i, c[i], goLines[i])
			}
		}
		t.Fatal("partition label output lengths differ")
	}
}

func TestSATDfsPartitionsVisitChoiceFirstAndStartTraversal(t *testing.T) {
	for _, preorder := range []int32{0, 1} {
		g := NewGraph()
		a, b, c := g.NewInput(), g.NewInput(), g.NewInput()
		alternative := g.AND(a, g.AND(b, c))
		root := g.AND(g.AND(a, b), c)
		g.AddOutput(root)
		g.equivs = map[int]Edge{root.NodeID(): alternative}
		var order []Edge
		if preorder != 0 {
			order = Aig_ManDfsPreorder(g, 1)
		} else {
			order = Aig_ManDfs(g, 1)
		}
		want := make([]int32, len(g.nodes))
		for i, e := range order {
			if preorder == 0 {
				want[e.NodeID()] = int32(len(order) - 1 - i)
			} else {
				want[e.NodeID()] = int32(i)
			}
		}
		before := g.travID
		got := Aig_ManPartitionDfs(g, 1, preorder)
		if !reflect.DeepEqual(got, want) {
			t.Errorf("preorder %d choice labels %v want %v", preorder, got, want)
		}
		if g.travID != before+1 {
			t.Errorf("preorder %d traversal generation did not advance", preorder)
		}
	}
}

func TestSATPartitionSplitUsesBoundedTraversalGeneration(t *testing.T) {
	g := NewGraph()
	a, b := g.NewInput(), g.NewInput()
	root := g.AND(a, b)
	g.AddOutput(root)
	g.travID = (1 << 30) - 1
	g.node(a).TravID = g.travID
	g.node(b).TravID = 17
	part, pio := Aig_ManPartSplitOne(g, []Edge{root})
	if g.travID != 2 {
		t.Fatalf("split traversal escaped bounded generations: %d", g.travID)
	}
	if len(pio) != 3 || len(part.inputs) != 2 || len(part.outputs) != 1 {
		t.Fatal("split boundary changed", pio)
	}
	for _, e := range []Edge{a, b, root} {
		if g.node(e).TravID != g.travID {
			t.Fatal("split did not mark copied node")
		}
	}
}
