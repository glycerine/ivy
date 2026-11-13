package dch

import (
	"testing"

	aig "github.com/glycerine/abcgo/aig/aig"
)

// dchCore.c resets Aig_ManRandom, and dchSim.c consumes one word for
// each CI in each simulation round. The caller observes the advanced stream.
func TestComputeUpdatesSourceRandomStream(t *testing.T) {
	for _, choices := range []bool{false, true} {
		t.Run(map[bool]string{false: "equivalences", true: "choices"}[choices], func(t *testing.T) {
			graph := aig.NewGraph()
			input := graph.NewInput()
			graph.AddOutput(input)
			params := DefaultParams()
			aig.Aig_ManRandom(1)
			for i := 0; i < 11; i++ {
				aig.Aig_ManRandom(0)
			}
			if choices {
				if _, err := ComputeChoices(graph, &params); err != nil {
					t.Fatal(err)
				}
			} else if err := ComputeEquivalences(graph, &params); err != nil {
				t.Fatal(err)
			}
			got := aig.Aig_ManRandom(0)
			wantSource := aig.NewAigRandomSource()
			wantSource.Uint32() // reset draw
			for i := 0; i < 8*params.Words; i++ {
				wantSource.Uint32()
			}
			want := wantSource.Uint32()
			if got != want {
				t.Fatalf("next global random word = %08x, want %08x", got, want)
			}
		})
	}
}
