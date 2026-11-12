package gia

import "testing"

func jfTestCut(time int32, flow float32, leaves ...int32) *Jf_Cut_t {
	c := &Jf_Cut_t{Time: time, Flow: flow}
	c.pCut[0] = int32(len(leaves))
	copy(c.pCut[1:], leaves)
	c.Sign = Jf_CutGetSign(c.pCut[:])
	return c
}
func TestJfCutComparisons(t *testing.T) {
	a, b := jfTestCut(2, 3, 2), jfTestCut(3, 1, 2, 4)
	if Jf_CutCompareDelay(a, b) >= 0 || Jf_CutCompareArea(a, b) <= 0 {
		t.Fatal("comparison priority")
	}
	b.Time = 2
	b.pCut[0] = 1
	b.Flow = 3 + JF_EPSILON/2
	if Jf_CutCompareDelay(a, b) != 0 || Jf_CutCompareArea(a, b) != 0 {
		t.Fatal("flow tolerance")
	}
}
func TestJfCutStoreFilteringAndOrdering(t *testing.T) {
	for _, min := range []bool{false, true} {
		a, b, c := jfTestCut(1, 1, 2, 4), jfTestCut(3, 1, 6, 8), jfTestCut(2, 1, 2, 4, 6)
		store := []*Jf_Cut_t{a, b, c, nil}
		if jfObjCutFilter(store, 2, min) {
			t.Fatal("contained candidate accepted")
		}
		if got := jfObjAddCutToStore(store, 2, 3, min, Jf_CutCompareDelay); got != 2 || store[0] != a || store[1] != b {
			t.Fatal("contained candidate inserted")
		}
		c = jfTestCut(0, 1, 6)
		store[2] = c
		if got := jfObjAddCutToStore(store, 2, 3, min, Jf_CutCompareDelay); got != 2 || store[0] != c || store[1] != a {
			t.Fatalf("dominance/order %d %v", got, store)
		}
	}
	a, b, c := jfTestCut(1, 1, 2, 4), jfTestCut(1, 1, 6, 8), jfTestCut(1, 1, 2)
	store := []*Jf_Cut_t{a, b, c}
	if n := jfObjCutFilterBoth(store, 2); n != 1 || store[0] != b || store[1] != c {
		t.Fatalf("bidirectional filter %d", n)
	}
}
