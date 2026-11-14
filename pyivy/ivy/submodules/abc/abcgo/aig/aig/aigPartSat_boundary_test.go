package aig

import (
	"fmt"
	"strings"
	"testing"
)

func TestSATSplitCollapsedBufferKeepsBothBoundaryIDs(t *testing.T) {
	g := NewGraph()
	a, b := g.NewInput(), g.NewInput()
	gate := g.AND(a, b)
	buffer := Aig_ObjCreate(g, &Node{Kind: Buffer, Fanins: [2]Edge{gate, InvalidEdge}})
	root := g.AND(gate, buffer)
	g.AddOutput(gate)
	g.AddOutput(root)
	part, pio := Aig_ManPartSplitOne(g, []Edge{gate, buffer, root})
	if len(part.outputs) != 2 || len(pio) != 4 || pio[2] != int32(gate.NodeID()) || pio[3] != int32(root.NodeID()) {
		t.Fatalf("collapsed aliases lost original boundary IDs: inputs=%d outputs=%d pio=%v", len(part.inputs), len(part.outputs), pio)
	}
}

func TestSATSplitBoundaryMapsMatchOriginalC(t *testing.T) {
	var want strings.Builder
	for mask := 1; mask < 8; mask++ {
		for _, reverse := range []bool{false, true} {
			g := NewGraph()
			a, b, c := g.NewInput(), g.NewInput(), g.NewInput()
			x, y := g.AND(a, b), g.AND(b.Not(), c)
			z := g.AND(x, y)
			g.AddOutput(x.Not())
			g.AddOutput(z)
			g.node(b).MarkA = true
			g.node(c).MarkB = true
			all := []Edge{x, y, z}
			var selected []Edge
			for i := 0; i < 3; i++ {
				j := i
				if reverse {
					j = 2 - i
				}
				if mask&(1<<j) != 0 {
					selected = append(selected, all[j])
				}
			}
			part, pio := Aig_ManPartSplitOne(g, selected)
			fmt.Fprintf(&want, "mask %d reverse %t pio:", mask, reverse)
			for _, id := range pio {
				fmt.Fprintf(&want, " %d", id)
			}
			fmt.Fprintln(&want)
			for _, e := range part.LiveEdges() {
				n := part.Node(e)
				fmt.Fprintf(&want, "%d %d %t %t", e.NodeID(), n.Refs, n.MarkA, n.MarkB)
				switch n.Kind {
				case And:
					fmt.Fprintf(&want, " and %d %d", n.Fanins[0], n.Fanins[1])
				case Output:
					fmt.Fprintf(&want, " out %d", n.Fanins[0])
				}
				fmt.Fprintln(&want)
			}
			if err := part.Check(); err != nil {
				t.Fatal(err)
			}
		}
	}
	got := runAigSourceFixture(t, "sat_partition_boundary_compare.c", "aigMan.c", "aigObj.c", "aigOper.c", "aigTable.c", "aigMem.c", "aigFanout.c", "aigUtil.c", "aigPartSat.c", "../../misc/tim/timMan.c", "../../misc/mem/mem.c")
	if got != want.String() {
		t.Fatalf("C boundaries:\n%sGo boundaries:\n%s", got, want.String())
	}
}
