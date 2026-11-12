package ivy

import (
	"fmt"
	"slices"
	"testing"
)

func TestTraversalCutsMatchExpansionAndClearScratch(t *testing.T) {
	p := Ivy_ManStart()
	a, b, c := Ivy_ObjCreatePi(p), Ivy_ObjCreatePi(p), Ivy_ObjCreatePi(p)
	ab := Ivy_And(p, a, b)
	root := Ivy_And(p, ab, c)
	Ivy_ObjCreatePo(p, root)
	var nodes, front []*Ivy_Obj_t
	var arena []uint32
	var rows [][][]uint32
	cuts := Ivy_NodeFindCutsTravAll(p, root, 3, 60, &nodes, &front, &arena, &rows)
	expected := Ivy_NodeFindCutsAll(p, root, 3)
	want := map[string]bool{}
	for _, cut := range expected.pCuts[:expected.nCuts] {
		want[fmt.Sprint(cut.pArray[:cut.nSize])] = true
	}
	if cuts.nCuts != expected.nCuts {
		t.Fatalf("traversal count %d, expansion count %d", cuts.nCuts, expected.nCuts)
	}
	for _, cut := range cuts.pCuts[:cuts.nCuts] {
		if !want[fmt.Sprint(cut.pArray[:cut.nSize])] {
			t.Fatal("unexpected traversal cut")
		}
	}
	for _, n := range nodes {
		if n.TravId != 0 || n.fMarkA != 0 {
			t.Fatal("scratch indexes/marks leaked")
		}
	}
	first := *cuts
	Ivy_NodeFindCutsTravAll(p, ab, 2, 60, &nodes, &front, &arena, &rows)
	if *cuts != first {
		t.Fatal("subsequent traversal mutated owned result")
	}
}
func TestCutVolumeDeduplicatesSharedCone(t *testing.T) {
	p := Ivy_ManStart()
	inputs := make([]*Ivy_Obj_t, 5)
	for i := range inputs {
		inputs[i] = Ivy_ObjCreatePi(p)
	}
	shared := Ivy_And(p, inputs[0], inputs[1])
	left := Ivy_And(p, shared, inputs[2])
	right := Ivy_And(p, shared, inputs[3])
	root := Ivy_And(p, left, right)
	var nodes, front []*Ivy_Obj_t
	Ivy_NodeComputeVolume(root, 60, &nodes, &front)
	seen := map[*Ivy_Obj_t]bool{}
	for _, n := range nodes {
		if seen[n] {
			t.Fatal("shared node duplicated")
		}
		seen[n] = true
		if n.fMarkA != 0 {
			t.Fatal("volume marks leaked")
		}
	}
	if len(nodes) != 8 || len(front) != 4 {
		t.Fatalf("nodes=%d frontier=%d", len(nodes), len(front))
	}
	if !slices.Contains(nodes, shared) {
		t.Fatal("shared cone not collected")
	}
	Ivy_NodeComputeVolume2(root, 4, &nodes, &front)
	if len(nodes) < 4 {
		t.Fatal("size-driven volume stopped too early")
	}
	for i, n := range nodes {
		if n.fMarkA != 0 || n.TravId != int32(i) {
			t.Fatal("volume2 scratch inconsistency")
		}
	}
}
func TestTraversalCycleRejectedWithoutMarks(t *testing.T) {
	p := Ivy_ManStart()
	a, b := Ivy_ObjCreatePi(p), Ivy_ObjCreatePi(p)
	root := Ivy_And(p, a, b)
	root.pFanin0 = root
	root.TravId = 73
	var nodes, front []*Ivy_Obj_t
	var arena []uint32
	var rows [][][]uint32
	func() {
		defer func() {
			if recover() == nil {
				t.Error("cyclic traversal accepted")
			}
		}()
		Ivy_NodeFindCutsTravAll(p, root, 3, 60, &nodes, &front, &arena, &rows)
	}()
	if root.fMarkA != 0 || a.fMarkA != 0 || b.fMarkA != 0 || root.TravId != 73 {
		t.Fatal("cycle rejection changed marks or IDs")
	}
}
func TestBitCutWordsAndDominanceRejectionPreservesPriorCuts(t *testing.T) {
	var arena []uint32
	first := Ivy_NodeCutElementary(&arena, 3, 64)
	for i := 0; i < 1000; i++ {
		Ivy_NodeCutElementary(&arena, 3, int32(i%96))
	}
	if first[2] != 1 {
		t.Fatal("arena growth invalidated elementary cut")
	}
	// The candidate {1,2} would dominate the first cut, but is rejected by {1}.
	// Rejection must leave both original records untouched.
	cuts := [][]uint32{{14}, {2}}
	before := [][]uint32{slices.Clone(cuts[0]), slices.Clone(cuts[1])}
	Ivy_NodeFindCutsMerge([][]uint32{{2}}, [][]uint32{{4}}, &cuts, 3, 1, &arena)
	if len(cuts) != 2 || !slices.Equal(cuts[0], before[0]) || !slices.Equal(cuts[1], before[1]) {
		t.Fatal("rejected candidate corrupted existing cuts")
	}
}
func TestTraversalBufferCut(t *testing.T) {
	p := Ivy_ManStart()
	a, b := Ivy_ObjCreatePi(p), Ivy_ObjCreatePi(p)
	buffer := Ivy_ObjCreate(p, Ivy_ObjCreateGhost(p, a, nil, IVY_BUF, IVY_INIT_NONE))
	root := Ivy_And(p, buffer, b)
	var nodes, front []*Ivy_Obj_t
	var arena []uint32
	var rows [][][]uint32
	cuts := Ivy_NodeFindCutsTravAll(p, root, 2, 60, &nodes, &front, &arena, &rows)
	if cuts.nCuts < 2 {
		t.Fatal("buffered traversal cut missing")
	}
}

func TestTraversalCutsCrossWordBoundary(t *testing.T) {
	p := Ivy_ManStart()
	layer := make([]*Ivy_Obj_t, 32)
	for i := range layer {
		layer[i] = Ivy_ObjCreatePi(p)
	}
	for len(layer) > 1 {
		var next []*Ivy_Obj_t
		for i := 0; i < len(layer); i += 2 {
			next = append(next, Ivy_And(p, layer[i], layer[i+1]))
		}
		layer = next
	}
	root := layer[0]
	var nodes, front []*Ivy_Obj_t
	var arena []uint32
	var rows [][][]uint32
	store := Ivy_NodeFindCutsTravAll(p, root, 3, 60, &nodes, &front, &arena, &rows)
	if len(nodes) != 63 || store.nCuts != 4 {
		t.Fatalf("nodes=%d cuts=%d", len(nodes), store.nCuts)
	}
	for _, cut := range store.pCuts[:store.nCuts] {
		boundary := map[int32]bool{}
		for _, id := range cut.pArray[:cut.nSize] {
			boundary[id] = true
		}
		var covered func(*Ivy_Obj_t) bool
		covered = func(n *Ivy_Obj_t) bool {
			if boundary[n.Id] {
				return true
			}
			if Ivy_ObjIsCi(n) != 0 {
				return false
			}
			return covered(Ivy_ObjFanin0(n)) && covered(Ivy_ObjFanin1(n))
		}
		if !covered(root) {
			t.Fatalf("cut does not separate inputs: %v", cut.pArray[:cut.nSize])
		}
	}
}
