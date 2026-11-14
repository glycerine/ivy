package aig

import (
	"fmt"
	"strings"
	"testing"
)

func TestDOTAgainstCompiledSource(t *testing.T) {
	var input, want strings.Builder
	for used := 0; used < 2; used++ {
		for highlighted := 0; highlighted < 2; highlighted++ {
			for history := 0; history < 2; history++ {
				g := NewGraph()
				a, b := g.NewInput(), g.NewInput()
				x, y := g.AND(a, b.Not()), g.AND(a.Not(), b)
				root := g.AND(x, y.Not())
				g.AddOutput(root)
				g.AddOutput(x.Not())
				if used != 0 {
					g.AddOutput(g.Const0())
				}
				var bold []Edge
				if highlighted != 0 {
					bold = []Edge{x.Not(), root}
				}
				fmt.Fprintf(&input, "%d %d %d\n", used, highlighted, history)
				data, err := aigDOT(g, bold)
				if err != nil {
					t.Fatal(err)
				}
				want.Write(data)
			}
		}
	}
	got := runAigSourceFixtureInput(t, "visualization_compare.c", input.String(), "aigUtil.c")
	if got != want.String() {
		gotLines, wantLines := strings.Split(got, "\n"), strings.Split(want.String(), "\n")
		for i := 0; i < len(gotLines) && i < len(wantLines); i++ {
			if gotLines[i] != wantLines[i] {
				t.Fatalf("line %d: C %q, Go %q", i, gotLines[i], wantLines[i])
			}
		}
		t.Fatalf("C has %d lines, Go has %d", len(gotLines), len(wantLines))
	}
}
