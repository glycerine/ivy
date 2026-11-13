package fraig

import (
	"reflect"
	"testing"
)

// The C vector orders by the regular node's level, but uniqueness compares
// tagged edges, so both polarities of the same node may be present.
func TestNodeVectorLevelOrderAndTaggedUniqueness(t *testing.T) {
	a, b, c := &Fraig_Node_t{Level: 1, Num: 3, nRefs: 2}, &Fraig_Node_t{Level: 3, Num: 1, nRefs: 1}, &Fraig_Node_t{Level: 3, Num: 2, nRefs: 2}
	v := Fraig_NodeVecAlloc(1)
	defer Fraig_NodeVecFree(v)
	if Fraig_NodeVecPushUniqueOrderByLevel(v, a) != 0 || Fraig_NodeVecPushUniqueOrderByLevel(v, b) != 0 || Fraig_NodeVecPushUniqueOrderByLevel(v, c) != 0 || Fraig_NodeVecPushUniqueOrderByLevel(v, Fraig_Not(b)) != 0 || Fraig_NodeVecPushUniqueOrderByLevel(v, b) != 1 {
		t.Fatal("tagged uniqueness")
	}
	if want := []*Fraig_Node_t{b, c, Fraig_Not(b), a}; !reflect.DeepEqual(v.pArray, want) {
		t.Fatalf("ordered edges %v", v.pArray)
	}
	Fraig_NodeVecWriteEntry(v, 3, Fraig_Not(a))
	if Fraig_NodeVecReadEntry(v, 3) != Fraig_Not(a) {
		t.Fatal("write entry")
	}
	Fraig_NodeVecSortByNumber(v)
	for i, want := range []int32{1, 1, 2, 3} {
		if got := Fraig_Regular(v.pArray[i]).Num; got != want {
			t.Fatalf("number order %v", v.pArray)
		}
	}
	if Fraig_NodeVecCompareLevelsIncreasing(&a, &b) != -1 || Fraig_NodeVecCompareLevelsDecreasing(&a, &b) != 1 || Fraig_NodeVecCompareNumbers(&a, &b) != 1 || Fraig_NodeVecCompareRefCounts(&a, &c) != -1 {
		t.Fatal("source comparator results")
	}
}
