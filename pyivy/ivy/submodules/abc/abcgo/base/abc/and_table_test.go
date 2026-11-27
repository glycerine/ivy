package abc

import (
	"math/rand/v2"
	"testing"
)

// TestAndTableMatchesMap drives andTable and a Go map with the same random
// sets and deletes on a small key space and checks they always agree.
func TestAndTableMatchesMap(t *testing.T) {
	var tab andTable
	ref := map[[2]int]*Object{}
	objs := make([]*Object, 8)
	for i := range objs {
		objs[i] = &Object{ID: i}
	}
	r := rand.New(rand.NewPCG(1, 2))
	for step := 0; step < 200000; step++ {
		k := [2]int{r.IntN(40) - 20, r.IntN(40) - 20}
		switch r.IntN(3) {
		case 0, 1:
			o := objs[r.IntN(len(objs))]
			tab.set(k, o)
			ref[k] = o
		case 2:
			tab.del(k)
			delete(ref, k)
		}
		if step%997 == 0 {
			if tab.len() != len(ref) {
				t.Fatalf("step %d: len %d vs %d", step, tab.len(), len(ref))
			}
			for a := -20; a < 20; a++ {
				for b := -20; b < 20; b++ {
					if tab.get([2]int{a, b}) != ref[[2]int{a, b}] {
						t.Fatalf("step %d: key %d,%d differs", step, a, b)
					}
				}
			}
		}
		if step == 100000 {
			tab.clear()
			clear(ref)
		}
	}
}
