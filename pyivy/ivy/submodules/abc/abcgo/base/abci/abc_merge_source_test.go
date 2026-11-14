package abci

import (
	"testing"

	"github.com/glycerine/abcgo/base/abc"
	"github.com/glycerine/abcgo/misc/vec"
)

func TestAbcNtkCountTotalFaninsCountsOnlyNewCandidateInputs(t *testing.T) {
	n := abc.NewNetwork("merge-union")
	a, b, c := n.NewObject(abc.ObjectPI), n.NewObject(abc.ObjectPI), n.NewObject(abc.ObjectPI)
	shared := n.NewObject(abc.ObjectNode)
	left := n.NewObject(abc.ObjectNode)
	right := n.NewObject(abc.ObjectNode)
	for _, edge := range []struct{ to, from *abc.Object }{{shared, a}, {left, shared}, {left, b}, {right, shared}, {right, c}} {
		if err := n.AddFanin(edge.to, edge.from.Signal()); err != nil {
			t.Fatal(err)
		}
	}
	shared.MarkC = true
	if got := Abc_NtkCountTotalFanins(left, right); got != 3 {
		t.Fatalf("merged support size = %d, want 3", got)
	}
	if !shared.MarkC {
		t.Fatal("source routine must leave candidate fanin marks for its caller to clean")
	}
}

func TestAbcNtkCollectOverlapCandsFiltersByMergedSupport(t *testing.T) {
	n := abc.NewNetwork("merge-candidates")
	n.Type = uint8(abc.ABC_NTK_LOGIC)
	a, b, c := n.NewObject(abc.ObjectPI), n.NewObject(abc.ObjectPI), n.NewObject(abc.ObjectPI)
	shared := n.NewObject(abc.ObjectNode)
	extra := n.NewObject(abc.ObjectNode)
	lut := n.NewObject(abc.ObjectNode)
	candidate := n.NewObject(abc.ObjectNode)
	for _, edge := range []struct{ to, from *abc.Object }{{shared, a}, {extra, c}, {lut, shared}, {lut, b}, {candidate, shared}, {candidate, extra}} {
		if err := n.AddFanin(edge.to, edge.from.Signal()); err != nil {
			t.Fatal(err)
		}
	}
	pars := &Nwk_LMPars_t{nMaxFanout: 10, nMaxLevelDiff: 10, nMaxSuppSize: 3}
	cands := vec.Vec_PtrAlloc(2)
	Abc_NtkCollectOverlapCands(lut, cands, pars)
	if vec.Vec_PtrSize(cands) != 1 || vec.Vec_PtrEntry(cands, 0) != candidate {
		t.Fatalf("overlap candidates = %#v, want candidate %d", cands, candidate.ID)
	}
	if shared.MarkC || extra.MarkC {
		t.Fatal("overlap collection must clear temporary fanin marks")
	}
}

func TestAbcNtkLutMergeReturnsEligibleOverlapPair(t *testing.T) {
	n := abc.NewNetwork("merge-pair")
	n.Type = uint8(abc.ABC_NTK_LOGIC)
	a, b, c := n.NewObject(abc.ObjectPI), n.NewObject(abc.ObjectPI), n.NewObject(abc.ObjectPI)
	shared := n.NewObject(abc.ObjectNode)
	left, right := n.NewObject(abc.ObjectNode), n.NewObject(abc.ObjectNode)
	for _, edge := range []struct{ to, from *abc.Object }{{shared, a}, {left, shared}, {left, b}, {right, shared}, {right, c}} {
		if err := n.AddFanin(edge.to, edge.from.Signal()); err != nil {
			t.Fatal(err)
		}
	}
	pars := &Nwk_LMPars_t{nMaxLutSize: 4, nMaxSuppSize: 4, nMaxFanout: 10, nMaxLevelDiff: 10}
	pairs := Abc_NtkLutMerge(n, pars)
	if got := vec.Vec_IntSize(pairs); got != 2 {
		t.Fatalf("merge result has %d endpoints, want one pair", got)
	}
	aID, bID := vec.Vec_IntEntry(pairs, 0), vec.Vec_IntEntry(pairs, 1)
	if !((aID == int32(left.ID) && bID == int32(right.ID)) || (aID == int32(right.ID) && bID == int32(left.ID))) {
		t.Fatalf("merge pair = (%d,%d), want (%d,%d)", aID, bID, left.ID, right.ID)
	}
}

func TestAbcNtkCollectNonOverlapCandsHonorsDistanceAndSupport(t *testing.T) {
	n := abc.NewNetwork("merge-distance")
	n.Type = uint8(abc.ABC_NTK_LOGIC)
	a, b, c := n.NewObject(abc.ObjectPI), n.NewObject(abc.ObjectPI), n.NewObject(abc.ObjectPI)
	root := n.NewObject(abc.ObjectNode)
	other := n.NewObject(abc.ObjectNode)
	lut := n.NewObject(abc.ObjectNode)
	near := n.NewObject(abc.ObjectNode)
	for _, edge := range []struct{ to, from *abc.Object }{{root, a}, {other, c}, {lut, root}, {lut, b}, {near, lut}, {near, other}} {
		if err := n.AddFanin(edge.to, edge.from.Signal()); err != nil {
			t.Fatal(err)
		}
	}
	pars := &Nwk_LMPars_t{nMaxSuppSize: 4, nMaxDistance: 1, nMaxLevelDiff: 10, nMaxFanout: 10, fUseTfiTfo: 1}
	start, next, cands := vec.Vec_PtrAlloc(4), vec.Vec_PtrAlloc(4), vec.Vec_PtrAlloc(4)
	Abc_NtkCollectNonOverlapCands(lut, start, next, cands, pars)
	found := false
	for i := 0; i < vec.Vec_PtrSize(cands); i++ {
		if vec.Vec_PtrEntry(cands, i) == near {
			found = true
		}
	}
	if !found {
		t.Fatalf("distance-one candidates omitted adjacent node %d", near.ID)
	}
}
