package fraig

import "testing"

func TestNodeVectorOwnershipAndOrdering(t *testing.T) {
	a := &Fraig_Node_t{Num: 3, Level: 1, nRefs: 2}
	b := &Fraig_Node_t{Num: 1, Level: 4, nRefs: 1}
	c := &Fraig_Node_t{Num: 2, Level: 2, nRefs: 2}
	v := Fraig_NodeVecAlloc(0)
	for _, node := range []*Fraig_Node_t{a, Fraig_Not(b), c, b} {
		if Fraig_NodeVecPushUniqueOrder(v, node) != 0 {
			t.Fatal("new node reported duplicate")
		}
	}
	if Fraig_NodeVecPushUniqueOrder(v, b) != 1 {
		t.Fatal("duplicate not detected")
	}
	want := []*Fraig_Node_t{b, Fraig_Not(b), c, a}
	for i, node := range want {
		if v.pArray[i] != node {
			t.Fatalf("ordering = %v", v.pArray)
		}
	}
	copy := Fraig_NodeVecDup(v)
	Fraig_NodeVecRemove(copy, c)
	if len(v.pArray) != 4 || len(copy.pArray) != 3 {
		t.Fatal("duplicate shares vector storage")
	}
	Fraig_NodeVecSortByLevel(v, 0)
	if Fraig_Regular(v.pArray[0]).Level != 4 || Fraig_Regular(v.pArray[3]).Level != 1 {
		t.Fatal("level order")
	}
	Fraig_NodeVecSortByRefCount(v)
	if Fraig_Regular(v.pArray[0]).nRefs != 1 || Fraig_Regular(v.pArray[3]).Level != 2 {
		t.Fatal("reference order")
	}
	Fraig_NodeVecClear(v)
	if len(v.pArray) != 0 {
		t.Fatal("clear retained entries")
	}
}

func TestNodeVectorGrowthAndRemoval(t *testing.T) {
	v := Fraig_NodeVecAlloc(1)
	for i := int32(0); i < 100; i++ {
		Fraig_NodeVecPush(v, &Fraig_Node_t{Num: i})
	}
	Fraig_NodeVecGrow(v, 300)
	if Fraig_NodeVecReadSize(v) != 100 || cap(v.pArray) < 300 {
		t.Fatal("growth lost entries")
	}
	removed := Fraig_NodeVecReadEntry(v, 25)
	Fraig_NodeVecRemove(v, removed)
	if Fraig_NodeVecReadEntry(v, 25).Num != 26 || Fraig_NodeVecPop(v).Num != 99 {
		t.Fatal("removal changed ordering")
	}
	Fraig_NodeVecShrink(v, 2)
	if Fraig_NodeVecReadSize(v) != 2 {
		t.Fatal("shrink failed")
	}
}
