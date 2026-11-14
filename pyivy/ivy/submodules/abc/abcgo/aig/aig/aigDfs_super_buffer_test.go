package aig

import (
	"reflect"
	"testing"
)

func TestCollectSuperBypassesBufferChainsAndRetainsPhase(t *testing.T) {
	for _, complemented := range []bool{false, true} {
		t.Run(map[bool]string{false: "positive", true: "negative"}[complemented], func(t *testing.T) {
			g := NewGraph()
			a, b, c := g.NewInput(), g.NewInput(), g.NewInput()
			ab := g.AND(a, b)
			first := Aig_ObjCreate(g, &Node{Kind: Buffer, Fanins: [2]Edge{ab.Not(), InvalidEdge}})
			driver := first.Not()
			if complemented {
				driver = first
			}
			second := Aig_ObjCreate(g, &Node{Kind: Buffer, Fanins: [2]Edge{driver, InvalidEdge}})
			root := g.AND(second, c)
			g.AddOutput(root)
			var leaves []Edge
			if result := Aig_ObjCollectSuper(g, root, &leaves); result != 0 {
				t.Fatalf("unexpected duplicate result %d", result)
			}
			want := []Edge{a, b, c}
			if complemented {
				want = []Edge{ab.Not(), c}
			}
			// Hash ordering can place c before the buffer; compare leaf membership.
			gotSet, wantSet := map[Edge]bool{}, map[Edge]bool{}
			for _, e := range leaves {
				gotSet[e] = true
				g.AddOutput(e)
			}
			for _, e := range want {
				wantSet[e] = true
			}
			if !reflect.DeepEqual(gotSet, wantSet) || len(leaves) != len(want) {
				t.Fatalf("supergate leaves=%v want=%v", leaves, want)
			}
			for mask := 0; mask < 8; mask++ {
				values, err := g.Evaluate([]bool{mask&1 != 0, mask&2 != 0, mask&4 != 0})
				if err != nil {
					t.Fatal(err)
				}
				combined := true
				for _, v := range values[1:] {
					combined = combined && v
				}
				if combined != values[0] {
					t.Fatalf("mask=%d leaves do not reproduce root", mask)
				}
			}
		})
	}
}
