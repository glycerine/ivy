package if_

import (
	"math/rand"
	"slices"
	"testing"
)

// ifSortedSubset must answer If_CutCheckDominance for ordered leaves.
func TestSortedSubsetMatchesDominance(t *testing.T) {
	r := rand.New(rand.NewSource(1))
	set := func() []int32 {
		n := r.Intn(7)
		m := map[int32]bool{}
		for len(m) < n {
			m[int32(r.Intn(12))] = true
		}
		var s []int32
		for k := range m {
			s = append(s, k)
		}
		slices.Sort(s)
		return s
	}
	for i := 0; i < 200000; i++ {
		a, b := set(), set()
		if len(a) > len(b) {
			a, b = b, a
		}
		dom := &If_Cut_t{pLeaves: a, nLeaves: uint32(len(a))}
		cut := &If_Cut_t{pLeaves: b, nLeaves: uint32(len(b))}
		if got, want := ifSortedSubset(a, b), If_CutCheckDominance(dom, cut) != 0; got != want {
			t.Fatalf("%v in %v: got %v, want %v", a, b, got, want)
		}
	}
}
