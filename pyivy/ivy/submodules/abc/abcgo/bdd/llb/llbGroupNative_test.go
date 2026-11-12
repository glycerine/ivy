package llb

import (
	"testing"

	"github.com/glycerine/abcgo/aig/aig"
	"github.com/glycerine/abcgo/misc/vec"
)

func llbTestManager(graph *aig.Graph) *Llb_Man_t {
	return &Llb_Man_t_{pAig: graph, vGroups: vec.Vec_PtrAlloc(8)}
}

func llbTestEdges(v *Vec_Ptr_t) []aig.Edge {
	result := make([]aig.Edge, vec.Vec_PtrSize(v))
	for i := range result {
		result[i] = llbGroupEdge(v, i)
	}
	return result
}

func TestGroupCreationCollectsSupportAndCone(t *testing.T) {
	g := aig.NewGraph()
	x, y := g.NewInput(), g.NewInput()
	root := g.AND(x, y)
	aig.Aig_ObjSetMarkA(g, root)
	manager := llbTestManager(g)
	group := Llb_ManGroupCreate(manager, root)
	if got := llbTestEdges(group.vIns); len(got) != 2 || got[0] != x || got[1] != y {
		t.Fatalf("group support = %v, want [%v %v]", got, x, y)
	}
	if got := llbTestEdges(group.vOuts); len(got) != 1 || got[0] != root {
		t.Fatalf("group outputs = %v, want [%v]", got, root)
	}
	if got := llbTestEdges(group.vNodes); len(got) != 1 || got[0] != root {
		t.Fatalf("group internal nodes = %v, want [%v]", got, root)
	}
}

func TestGroupsCombineDeduplicatesBoundaries(t *testing.T) {
	g := aig.NewGraph()
	x, y, z := g.NewInput(), g.NewInput(), g.NewInput()
	a, b := g.AND(x, y), g.AND(y, z)
	manager := llbTestManager(g)
	first, second := Llb_ManGroupAlloc(manager), Llb_ManGroupAlloc(manager)
	llbGroupPush(first.vIns, x)
	llbGroupPush(first.vIns, y)
	llbGroupPush(first.vOuts, a)
	llbGroupPush(second.vIns, y)
	llbGroupPush(second.vIns, z)
	llbGroupPush(second.vOuts, b)
	combined := Llb_ManGroupsCombine(first, second)
	if got := llbTestEdges(combined.vIns); len(got) != 3 || got[0] != x || got[1] != y || got[2] != z {
		t.Fatalf("combined inputs = %v", got)
	}
	if got := llbTestEdges(combined.vOuts); len(got) != 2 || got[0] != a || got[1] != b {
		t.Fatalf("combined outputs = %v", got)
	}
}

func TestGroupCreateFromCutsUsesReachableCutBoundary(t *testing.T) {
	g := aig.NewGraph()
	x, y, z, unused := g.NewInput(), g.NewInput(), g.NewInput(), g.NewInput()
	left := g.AND(x, y)
	root := g.AND(left, z)
	manager := llbTestManager(g)
	cut1 := vec.Vec_IntAllocArrayCopy([]int32{int32(unused.NodeID()), int32(x.NodeID()), int32(y.NodeID()), int32(z.NodeID())}, 4)
	cut2 := vec.Vec_IntAllocArrayCopy([]int32{int32(root.NodeID())}, 1)
	group := Llb_ManGroupCreateFromCuts(manager, cut1, cut2)
	if got := llbTestEdges(group.vIns); len(got) != 3 || got[0] != x || got[1] != y || got[2] != z {
		t.Fatalf("group inputs = %v, want reachable cut inputs [%v %v %v]", got, x, y, z)
	}
	if got := llbTestEdges(group.vOuts); len(got) != 1 || got[0] != root {
		t.Fatalf("group outputs = %v, want [%v]", got, root)
	}
	if got := llbTestEdges(group.vNodes); len(got) != 2 || got[0] != left || got[1] != root {
		t.Fatalf("group internal nodes = %v, want [%v %v]", got, left, root)
	}
}
