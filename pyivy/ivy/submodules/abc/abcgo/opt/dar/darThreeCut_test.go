package dar

import (
	"github.com/glycerine/abcgo/aig/aig"
	"testing"
)

// darCore.c appends constant-one to each three-leaf cut before matching the
// four-input library; a three-input AND must therefore reach that path.
func TestRewriteEvaluatesThreeInputCuts(t *testing.T) {
	g := aig.NewGraph()
	a, b, c := g.NewInput(), g.NewInput(), g.NewInput()
	root := g.AND(g.AND(a, b), c)
	g.AddOutput(root)
	m, err := Dar_ManStart(g, nil)
	if err != nil {
		t.Fatal(err)
	}
	defer Dar_ManStop(m)
	if err := m.Rewrite(); err != nil {
		t.Fatal(err)
	}
	if m.Stats.Cuts == 0 {
		t.Fatal("three-input cuts never reached the rewrite library")
	}
	for mask := 0; mask < 8; mask++ {
		input := []bool{mask&1 != 0, mask&2 != 0, mask&4 != 0}
		out, err := g.Evaluate(input)
		if err != nil {
			t.Fatal(err)
		}
		want := input[0] && input[1] && input[2]
		if out[0] != want {
			t.Fatalf("input %v got %v want %t", input, out, want)
		}
	}
}
