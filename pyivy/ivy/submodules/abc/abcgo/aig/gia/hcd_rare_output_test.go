package gia

import (
	"fmt"
	"testing"

	"github.com/glycerine/abcgo/aig/aig"
)

func TestHcdChoicesPreservesRareHighestLevelOutput(t *testing.T) {
	for _, synthesis := range []int32{0, 1} {
		for _, tc := range []struct{ mini, limit int32 }{{0, 1000}, {1, 1000}, {0, -1}} {
			t.Run(fmt.Sprintf("synthesis%d_backend%d_limit%d", synthesis, tc.mini, tc.limit), func(t *testing.T) {
				g := aig.NewGraph()
				pool := make([]aig.Edge, 32)
				for i := range pool {
					pool[i] = g.NewInput()
				}
				for len(pool) > 1 {
					var next []aig.Edge
					for i := 0; i < len(pool); i += 2 {
						next = append(next, g.AND(pool[i], pool[i+1]))
					}
					pool = next
				}
				g.AddOutput(pool[0])
				g.AddOutput(pool[0].Not())
				q := Hcd_ComputeChoices(g, tc.limit, synthesis, tc.mini, 0)
				if err := q.Check(); err != nil {
					t.Fatal(err)
				}
				for zero := -1; zero < 32; zero++ {
					inputs := make([]bool, 32)
					for i := range inputs {
						inputs[i] = i != zero
					}
					got, err := q.Evaluate(inputs)
					want := zero == -1
					if err != nil || len(got) != 2 || got[0] != want || got[1] == want {
						t.Fatalf("zero input %d: %v, %v; want [%t %t]", zero, got, err, want, !want)
					}
				}
			})
		}
	}
}
