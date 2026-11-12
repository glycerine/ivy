package mapper

import "testing"

func TestNodeVecMutations(t *testing.T) {
	v := Map_NodeVecAlloc(0)
	defer Map_NodeVecFree(v)
	a, b := &Map_NodeStruct_t_{Num: 3, Level: 2}, &Map_NodeStruct_t_{Num: 1, Level: 1}
	Map_NodeVecPush(v, a)
	if got := Map_NodeVecPushUnique(v, a); got != 1 { t.Fatalf("duplicate push = %d, want 1", got) }
	if got := Map_NodeVecPushUnique(v, b); got != 0 { t.Fatalf("new push = %d, want 0", got) }
	Map_NodeVecSortByLevel(v)
	if got := Map_NodeVecReadEntry(v, 0); got != b { t.Fatalf("first node = %p, want %p", got, b) }
	copy := Map_NodeVecDup(v)
	if copy.nSize != 2 || copy.pArray[0] != b || copy.pArray[1] != a { t.Fatalf("duplicate vector = %#v", copy) }
	Map_NodeVecRemove(copy, b)
	if copy.nSize != 1 || copy.pArray[0] != a { t.Fatalf("after remove = %#v", copy) }
	Map_NodeVecClear(copy)
	if copy.nSize != 0 || len(copy.pArray) != 0 { t.Fatalf("clear left size=%d len=%d", copy.nSize, len(copy.pArray)) }
}

func TestComplementReferencesAreExplicitAndStable(t *testing.T) {
	node := &Map_NodeStruct_t_{Num: 7}
	neg := Map_Not(node)
	if !Map_IsComplement(neg) || Map_Regular(neg) != node { t.Fatal("complement wrapper lost its regular node or phase") }
	if Map_Not(neg) != node { t.Fatal("double complement did not return the original node") }
	if Map_NotCond(node, false) != node || Map_NotCond(node, true) != neg { t.Fatal("conditional complement mismatch") }
}
