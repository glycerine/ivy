package fra

import (
	"math/rand"
	"testing"

	"github.com/glycerine/abcgo/aig/aig"
)

// fraCec.c checks each output partition and stops on a failing partition.
// Enumerating every input provides an oracle independent of partitioning or SAT.
func TestPartitionedCECMatchesTruthTables(t *testing.T) {
	rng := rand.New(rand.NewSource(26389))
	for trial := 0; trial < 32; trial++ {
		build := func() *aig.Graph {
			g := aig.NewGraph()
			edges := []aig.Edge{g.NewInput(), g.NewInput(), g.NewInput()}
			for i := 0; i < 8; i++ {
				a, b := edges[rng.Intn(len(edges))], edges[rng.Intn(len(edges))]
				if rng.Intn(2) == 0 {
					a = a.Not()
				}
				if rng.Intn(2) == 0 {
					b = b.Not()
				}
				edges = append(edges, g.AND(a, b))
			}
			for i := 0; i < 4; i++ {
				e := edges[rng.Intn(len(edges))]
				if rng.Intn(2) == 0 {
					e = e.Not()
				}
				g.AddOutput(e)
			}
			return g
		}
		left := build()
		right := aig.Aig_ManDupOrdered(left)
		if trial%2 != 0 {
			right = build()
		}
		equivalent := true
		for mask := 0; mask < 8; mask++ {
			inputs := []bool{mask&1 != 0, mask&2 != 0, mask&4 != 0}
			a, err := left.Evaluate(inputs)
			if err != nil {
				t.Fatal(err)
			}
			b, err := right.Evaluate(inputs)
			if err != nil {
				t.Fatal(err)
			}
			for i := range a {
				if a[i] != b[i] {
					equivalent = false
				}
			}
		}
		want := int32(0)
		if equivalent {
			want = 1
		}
		for _, size := range []int32{0, 1, 2, 4} {
			for _, smart := range []int32{0, 1} {
				if got := Fra_FraigCecTop(left, right, 0, size, smart, 0); got != want {
					t.Fatalf("trial=%d size=%d smart=%d got=%d want=%d", trial, size, smart, got, want)
				}
			}
		}
	}
}

// fraCec.c compares each count to itself; an unused extra input on the
// second graph does not prevent source partitioned equivalence checking.
func TestCecTopSourceAcceptsUnusedExtraInput(t *testing.T) {
	left, right := aig.NewGraph(), aig.NewGraph()
	a, b := left.NewInput(), right.NewInput()
	right.NewInput()
	left.AddOutput(a)
	right.AddOutput(b)
	if status := Fra_FraigCecTop(left, right, 0, 0, 0, 0); status != 1 {
		t.Fatalf("status %d", status)
	}
}
