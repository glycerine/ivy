package aig

import (
	"fmt"
	"math/rand/v2"
	"strings"
	"testing"
)

func TestCleanupPreservesStandaloneBufferRoot(t *testing.T) {
	g := NewGraph()
	a, b := g.NewInput(), g.NewInput()
	gate := g.AND(a, b)
	buffer := Aig_ObjCreate(g, &Node{Kind: Buffer, Fanins: [2]Edge{gate, InvalidEdge}})
	if removed := Aig_ManCleanup(g); removed != 0 {
		t.Fatalf("removed %d gates, want zero for a buffer root", removed)
	}
	if !g.valid(buffer) || !g.valid(gate) || g.node(gate).Refs != 1 {
		t.Fatal("source retains the standalone buffer and its fanin")
	}
	if err := g.Check(); err != nil {
		t.Fatal(err)
	}
}

func TestCleanupAgainstCompiledSource(t *testing.T) {
	rng := rand.New(rand.NewPCG(13, 37))
	var input, want strings.Builder
	for trial := 0; trial < 256; trial++ {
		for fanouts := 0; fanouts < 2; fanouts++ {
			g := NewGraph()
			available := []Edge{g.Const1(), g.NewInput(), g.NewInput(), g.NewInput()}
			for step := 0; step < 24; step++ {
				a, b := available[rng.IntN(len(available))], available[rng.IntN(len(available))]
				if rng.IntN(2) == 0 {
					a = a.Not()
				}
				var root Edge
				switch rng.IntN(3) {
				case 0:
					root = g.AND(a, b)
				case 1:
					root = g.NewXor(a, b)
				default:
					root = Aig_ObjCreate(g, &Node{Kind: Buffer, Fanins: [2]Edge{a, InvalidEdge}})
				}
				available = append(available, root)
				if rng.IntN(6) == 0 {
					g.AddOutput(root)
				}
			}
			choice, alternative := -1, -1
			if trial == 0 {
				g = NewGraph()
				a, b, c := g.NewInput(), g.NewInput(), g.NewInput()
				root := g.AND(g.AND(a, b), c)
				alt := g.AND(a, g.AND(b, c))
				g.AddOutput(root)
				g.equivs = map[int]Edge{root.NodeID(): alt}
				choice, alternative = root.NodeID(), alt.NodeID()
				available = g.LiveEdges()
			}
			held := -1
			if trial%3 == 0 {
				for _, edge := range available {
					if g.node(edge).Kind == And || g.node(edge).Kind == ExclusiveOr {
						held = edge.NodeID()
						break
					}
				}
			}
			fmt.Fprintf(&input, "%d %d %d %d %d\n", len(g.nodes), fanouts, held, choice, alternative)
			for _, node := range g.nodes[1:] {
				a, b := int64(-1), int64(-1)
				pins := nodeFanins(node)
				if len(pins) > 0 {
					a = int64(pins[0])
				}
				if len(pins) > 1 {
					b = int64(pins[1])
				}
				fmt.Fprintf(&input, "%d %d %d\n", node.Kind, a, b)
			}
			if fanouts != 0 {
				Aig_ManFanoutStart(g)
			}
			if held >= 0 {
				g.nodes[held].Refs++
			}
			for pass := 0; pass < 2; pass++ {
				removed := Aig_ManCleanup(g)
				fmt.Fprintf(&want, "removed %d gates %d hash %d:", removed, Aig_ManNodeNum(g), Aig_TableCountEntries(g))
				for id, node := range g.nodes {
					if g.deleted[id] {
						want.WriteString(" dead")
						continue
					}
					a, b := int64(-1), int64(-1)
					pins := nodeFanins(node)
					if len(pins) > 0 {
						a = int64(pins[0])
					}
					if len(pins) > 1 {
						b = int64(pins[1])
					}
					fmt.Fprintf(&want, " %d/%d/%d", node.Refs, a, b)
					if fanouts != 0 {
						want.WriteByte('[')
						for _, pin := range g.fanouts.Handles(Edge(id) << 1) {
							fmt.Fprintf(&want, "%d,", pin)
						}
						want.WriteByte(']')
					}
				}
				want.WriteByte('\n')
				if pass == 0 && held >= 0 {
					g.nodes[held].Refs--
				}
				if err := g.Check(); err != nil {
					t.Fatalf("trial %d pass %d: %v", trial, pass, err)
				}
			}
		}
	}
	got := runAigSourceFixtureInput(t, "cleanup_compare.c", input.String(), "aigMan.c", "aigObj.c", "aigTable.c", "aigMem.c", "aigFanout.c", "aigUtil.c", "../../misc/tim/timMan.c", "../../misc/mem/mem.c")
	if got != want.String() {
		c, goLines := strings.Split(got, "\n"), strings.Split(want.String(), "\n")
		for i := 0; i < len(c) && i < len(goLines); i++ {
			if c[i] != goLines[i] {
				t.Fatalf("cleanup %d: C %q, Go %q", i, c[i], goLines[i])
			}
		}
		t.Fatalf("cleanup trace lengths C %d, Go %d", len(c), len(goLines))
	}
}

func TestCleanupDoesNotTreatChoiceLinksAsReferences(t *testing.T) {
	g := NewGraph()
	a, b, c := g.NewInput(), g.NewInput(), g.NewInput()
	root := g.AND(g.AND(a, b), c)
	alternative := g.AND(a, g.AND(b, c))
	g.AddOutput(root)
	g.equivs = map[int]Edge{root.NodeID(): alternative}
	if removed := Aig_ManCleanup(g); removed != 2 || g.valid(alternative) || !g.valid(root) {
		t.Fatalf("auxiliary choice links changed cleanup liveness: removed %d", removed)
	}
	// Like C, the primitive leaves caller-owned auxiliary tables alone. Such a
	// weak entry needs rebuilding before a later choice traversal.
	if g.equivs[root.NodeID()] != alternative {
		t.Fatal("cleanup rewrote the caller's auxiliary choice table")
	}
}

func TestCleanupCountsGatesWhenRecursingThroughBuffer(t *testing.T) {
	g := NewGraph()
	a, b, c := g.NewInput(), g.NewInput(), g.NewInput()
	inner := g.AND(a, b)
	buffer := Aig_ObjCreate(g, &Node{Kind: Buffer, Fanins: [2]Edge{inner, InvalidEdge}})
	root := g.AND(buffer, c)
	if removed := Aig_ManCleanup(g); removed != 2 {
		t.Fatalf("removed %d gates, want two (buffer storage is excluded)", removed)
	}
	for _, edge := range []Edge{inner, buffer, root} {
		if g.valid(edge) {
			t.Fatalf("dead cone object %d remains live", edge)
		}
	}
	if err := g.Check(); err != nil {
		t.Fatal(err)
	}
}
