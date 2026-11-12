package amap

import (
	"testing"

	"github.com/glycerine/abcgo/bool/kit"
)

func testAndLibrary(t *testing.T) (*Amap_Lib_t, *Amap_Nod_t) {
	t.Helper()
	lib := &Amap_Lib_t{}
	Amap_LibCreateVar(lib)
	// Amap library nodes are functions over one generic variable leaf. The
	// rule builder composes this same leaf into distinct truth-table inputs.
	nodeID := Amap_LibCreateNode(lib, 0, 0, 0)
	return lib, Amap_LibNod(lib, nodeID)
}

func TestLibraryFaninCollectionAndDSDPermutation(t *testing.T) {
	lib, andNode := testAndLibrary(t)
	fanins := Amap_LibCollectFanins(lib, andNode)
	gotFanins := append([]int32(nil), Vec_IntArray(fanins)...)
	Vec_IntFree(fanins)
	if len(gotFanins) != 2 || gotFanins[0] != 0 || gotFanins[1] != 0 {
		t.Fatalf("flattened AND fanins = %v, want [0 0]", gotFanins)
	}

	// The 2-input AND truth table in Kit's packed variable order.
	ntk := kit.Kit_DsdDecompose([]uint32{0x88888888}, 2)
	defer kit.Kit_DsdNtkFree(ntk)
	var permutation [AMAP_MAXINS]byte
	if Amap_LibDeriveGatePerm(lib, &Amap_Gat_t{nPins: 2}, ntk, andNode, permutation[:]) == 0 {
		t.Fatal("equivalent DSD AND did not produce a pin permutation")
	}
	if (permutation[0] != 0 && permutation[0] != 2) || (permutation[1] != 0 && permutation[1] != 2) || permutation[0] == permutation[1] {
		t.Fatalf("AND permutation = %v, want a permutation of [0 2]", permutation[:2])
	}
}

func TestRuleConstructionDeduplicatesCommutativeAnd(t *testing.T) {
	lib, _ := testAndLibrary(t)
	a, b := Vec_IntAlloc(1), Vec_IntAlloc(1)
	defer Vec_IntFree(a)
	defer Vec_IntFree(b)
	Vec_IntPush(a, 0)
	Vec_IntPush(b, 0)
	result := Vec_IntAlloc(2)
	defer Vec_IntFree(result)
	Amap_CreateRulesTwo(lib, result, a, b, 0)
	Amap_CreateRulesTwo(lib, result, b, a, 0)
	if Vec_IntSize(result) != 1 {
		t.Fatalf("commutative AND produced %d unique nodes, want one", Vec_IntSize(result))
	}
	if Vec_IntEntry(result, 0) != Abc_Var2Lit(1, 0) {
		t.Fatalf("rule result = %d, want node 1", Vec_IntEntry(result, 0))
	}
}

func TestSymmetryPermutationRetainsDestinationPinPhase(t *testing.T) {
	set := &Amap_Set_t{}
	set.Ins[0] = byte(Abc_Var2Lit(4, 1))
	set.Ins[1] = byte(Abc_Var2Lit(2, 0))
	copy := amapSymmetryCopy(set, 0, 1)
	if got, want := int32(copy.Ins[0]), Abc_Var2Lit(2, 1); got != want {
		t.Fatalf("swapped pin 0 = %d, want %d (phase retained)", got, want)
	}
	if got, want := int32(copy.Ins[1]), Abc_Var2Lit(4, 0); got != want {
		t.Fatalf("swapped pin 1 = %d, want %d (phase retained)", got, want)
	}
}
