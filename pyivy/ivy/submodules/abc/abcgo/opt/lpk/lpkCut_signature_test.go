package lpk

import (
	"testing"

	"github.com/glycerine/abcgo/base/abc"
	"github.com/glycerine/abcgo/misc/vec"
)

func TestCutSignatureAndDuplicateFiltering(t *testing.T) {
	cut := Lpk_Cut_t{nLeaves: 3}
	cut.pLeaves[0], cut.pLeaves[1], cut.pLeaves[2] = 1, 33, 65
	Lpk_NodeCutSignature(&cut)
	if cut.uSign != [2]uint32{2, 2} {
		t.Fatalf("signature = %v", cut.uSign)
	}
	cuts := []Lpk_Cut_t{cut}
	if Lpk_NodeCutsOneFilter(cuts, 1, &cut) != 1 {
		t.Fatal("identical cut was not found")
	}
	other := cut
	other.pLeaves[2] = 97 // same signature, different leaf IDs
	if Lpk_NodeCutsOneFilter(cuts, 1, &other) != 0 {
		t.Fatal("signature collision was mistaken for an identical cut")
	}
}

func TestCutDominanceFiltersSupersets(t *testing.T) {
	makeCut := func(leaves ...int32) Lpk_Cut_t {
		cut := Lpk_Cut_t{nLeaves: uint32(len(leaves))}
		copy(cut.pLeaves[:], leaves)
		Lpk_NodeCutSignature(&cut)
		return cut
	}
	large := makeCut(3, 9, 20)
	small := makeCut(3, 20)
	cuts := []Lpk_Cut_t{large}
	if Lpk_NodeCutsOneFilter(cuts, 1, &small) != 0 {
		t.Fatal("new smaller cut should be kept")
	}
	if cuts[0].nLeaves != 0 {
		t.Fatal("larger existing cut was not discarded by its subset")
	}
	cuts = []Lpk_Cut_t{small}
	if Lpk_NodeCutsOneFilter(cuts, 1, &large) != 1 {
		t.Fatal("existing smaller cut should dominate the new superset")
	}
}

func TestLpkPackTruthWordsPreservesWideAndOddTables(t *testing.T) {
	got := lpkPackTruthWords([]uint32{0x01234567, 0x89abcdef, 0xfedcba98})
	want := []uint64{0x89abcdef01234567, 0x00000000fedcba98}
	if len(got) != len(want) {
		t.Fatalf("packed word count = %d, want %d", len(got), len(want))
	}
	for i := range want {
		if got[i] != want[i] {
			t.Fatalf("packed word %d = %#016x, want %#016x", i, got[i], want[i])
		}
	}
}

func TestAbcNtkDecFromTruthRequiresAbcAPI(t *testing.T) {
	truth := []uint32{0, 0, ^uint32(0), ^uint32(0)}
	if network := Abc_NtkDecFromTruth(truth, 7, 4, nil); network != nil {
		abc.Abc_NtkDelete(network)
		t.Fatal("truth decomposition without abci integration unexpectedly returned a network")
	}
	if network := Abc_NtkDecFromTruth([]uint32{1}, 7, 4, nil); network != nil {
		abc.Abc_NtkDelete(network)
		t.Fatal("short truth table unexpectedly decomposed")
	}
}

func TestNodeImpactUsesRecordedObjectIDs(t *testing.T) {
	n := abc.Abc_NtkAlloc(abc.ABC_NTK_LOGIC, abc.ABC_FUNC_SOP, 1)
	defer abc.Abc_NtkDelete(n)
	input := abc.Abc_NtkCreatePi(n)
	inputID := input.ID
	p := &Lpk_Man_t{pNtk: n, vVisited: vec.Vec_VecStart(inputID + 1)}
	record := vec.Vec_VecEntry(p.vVisited, inputID)
	vec.Vec_PtrPush(record, inputID)
	vec.Vec_PtrPush(record, int32(0))
	if Lpk_NodeHasChanged(p, int32(inputID)) != 0 {
		t.Fatal("existing recorded object was treated as changed")
	}
	abc.Abc_NtkDeleteObj(input)
	if Lpk_NodeHasChanged(p, int32(inputID)) != 1 {
		t.Fatal("deleted recorded object was not detected")
	}
	if Lpk_NodeHasChanged(p, -1) != 1 {
		t.Fatal("negative object ID was not treated as changed")
	}
}

func TestCountSuppUsesAndClearsSourcePersistMarks(t *testing.T) {
	n := abc.Abc_NtkAlloc(abc.ABC_NTK_LOGIC, abc.ABC_FUNC_SOP, 1)
	defer abc.Abc_NtkDelete(n)
	a, b := abc.Abc_NtkCreatePi(n), abc.Abc_NtkCreatePi(n)
	node := abc.Abc_NtkCreateNode(n)
	abc.Abc_ObjAddFanin(node, a)
	abc.Abc_ObjAddFanin(node, b)
	abc.Abc_NtkIncrementTravId(n)
	a.Persist = true
	nodes := vec.Vec_PtrAlloc(1)
	vec.Vec_PtrPush(nodes, node)
	if got := Lpk_CountSupp(n, nodes); got != 1 {
		t.Fatalf("support count = %d, want 1", got)
	}
	if a.Persist || b.Persist {
		t.Fatal("source clears the persist marks after counting")
	}
}
