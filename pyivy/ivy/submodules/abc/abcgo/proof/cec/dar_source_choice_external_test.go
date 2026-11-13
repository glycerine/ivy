package cec_test

import (
	"testing"

	"github.com/glycerine/abcgo/aig/aig"
	"github.com/glycerine/abcgo/proof/cec"
)

func TestDarSourceChoicePublicDefaults(t *testing.T) {
	g := aig.NewGraph()
	a, b := g.NewInput(), g.NewInput()
	g.AddOutput(g.AND(a, b))
	result, err := cec.Dar_ManChoiceNew(g, nil, nil)
	if err != nil {
		t.Fatal(err)
	}
	for _, input := range [][]bool{{false, false}, {false, true}, {true, false}, {true, true}} {
		got, err := result.Evaluate(input)
		if err != nil || got[0] != (input[0] && input[1]) {
			t.Fatalf("input %v: result %v error %v", input, got, err)
		}
	}
}
