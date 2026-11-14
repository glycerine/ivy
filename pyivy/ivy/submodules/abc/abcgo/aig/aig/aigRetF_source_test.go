package aig

import (
	"fmt"
	"strings"
	"testing"
)

func TestRetimingFrontierMarkingAgainstCompiledSource(t *testing.T) {
	var want strings.Builder
	for mode := 0; mode < 3; mode++ {
		g := NewGraph()
		a, q, r, s := g.NewInput(), g.NewInput(), g.NewInput(), g.NewInput()
		root := g.AND(q, r)
		g.AND(r, s) // Dangling logic is outside the CO-based marking traversal.
		g.AddOutput(root)
		if mode == 0 {
			g.AddOutput(a)
		} else {
			g.AddOutput(r)
		}
		g.AddOutput(s)
		if mode == 0 {
			g.AddOutput(r.Not())
		} else {
			g.AddOutput(a)
		}
		g.nRegs = 3
		if mode == 2 {
			g.nRegs = 4
		}
		// The Go representation keeps terminal indices separate from C's pNext
		// scratch union. Marking must also leave unrelated payloads alone.
		g.objectData = map[int]any{q.NodeID(): "latch", g.outputs[1].NodeID(): "input"}
		for pass := 0; pass < 2; pass++ {
			Aig_ManRetimeMark(g)
			fmt.Fprintf(&want, "mode %d pass %d generation %d:", mode, pass, g.travID)
			for id, n := range g.nodes {
				mark := 0
				if n.MarkB {
					mark = 1
				}
				fmt.Fprintf(&want, " %d/%d", mark, n.TravID)
				g.nodes[id].MarkB = false
			}
			want.WriteByte('\n')
			if g.objectData[q.NodeID()] != "latch" || g.objectData[g.outputs[1].NodeID()] != "input" {
				t.Fatal("marking replaced unrelated object payloads")
			}
		}
	}
	got := runAigSourceFixture(t, "retiming_mark_compare.c", "aigUtil.c")
	if got != want.String() {
		t.Fatalf("Go:\n%sC:\n%s", want.String(), got)
	}
}

func TestRetimingFrontierSimplifiedNextState(t *testing.T) {
	for _, opposite := range []bool{false, true} {
		for _, firstComplement := range []bool{false, true} {
			for _, secondComplement := range []bool{false, true} {
				g := NewGraph()
				a, q, r := g.NewInput(), g.NewInput(), g.NewInput()
				first, second, next := q, r, a
				if firstComplement {
					first = first.Not()
				}
				if secondComplement {
					second = second.Not()
				}
				if opposite {
					next = next.Not()
				}
				g.AddOutput(g.AND(first, second))
				g.AddOutput(a)
				g.AddOutput(next)
				g.nRegs = 2
				before := Aig_ManDupOrdered(g)
				out := Aig_ManRetimeFrontier(g, 1)
				if !out.RetimingChanged || out.nRegs != 1 {
					t.Fatal("eligible gate was not retimed")
				}
				checkRetFTraces(t, before, out)
				for id, n := range g.nodes {
					if !g.deleted[id] && (n.MarkA || n.MarkB) {
						t.Fatalf("simplified result leaked scratch flags at %d", id)
					}
				}
			}
		}
	}
}
