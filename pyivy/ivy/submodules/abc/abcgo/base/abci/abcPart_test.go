package abci

import (
	"testing"

	"github.com/glycerine/abcgo/base/abc"
	"github.com/glycerine/abcgo/misc/vec"
)

func TestSupportManagerMergesSortedSetsAndReusesEntries(t *testing.T) {
	m := Supp_ManStart(128, 8)
	a := Supp_ManFetchEntry(m, 3, 2)
	a.pOuts = append(a.pOuts, 0, 2, 5)
	a.nOuts = 3
	b := Supp_ManFetchEntry(m, 3, 2)
	b.pOuts = append(b.pOuts, 1, 2, 6)
	b.nOuts = 3
	merged := Supp_ManMergeEntry(m, a, b, 1)
	want := []int32{0, 1, 2, 5, 6}
	if merged.nOuts != int32(len(want)) {
		t.Fatalf("merged length = %d, want %d", merged.nOuts, len(want))
	}
	for i, value := range want {
		if merged.pOuts[i] != value {
			t.Fatalf("merged[%d] = %d, want %d", i, merged.pOuts[i], value)
		}
	}
	Supp_ManRecycleEntry(m, a)
	Supp_ManRecycleEntry(m, b)
	reused := Supp_ManFetchEntry(m, 3, 1)
	if reused.nOuts != 0 || reused.nRefs != 1 || len(reused.pOuts) != 0 {
		t.Fatalf("recycled entry was not reset: %+v", reused)
	}
	Supp_ManStop(m)
}

func TestSupportManagerFetchAndRecycleMemory(t *testing.T) {
	m := Supp_ManStart(16, 8)
	first := Supp_ManFetch(m, 9)
	if len(first) != 9 {
		t.Fatalf("fetched memory size = %d, want 9", len(first))
	}
	first[0] = 17
	Supp_ManRecycle(m, first, 9)
	second := Supp_ManFetch(m, 9)
	if second[0] != 17 {
		t.Fatal("recycled support memory did not preserve contents")
	}
	Supp_ManStop(m)
}

func TestSmartAndNaiveSupportsIgnoreAIGPolarity(t *testing.T) {
	n, _, _ := fraigTestNetwork(t)
	defer abc.Abc_NtkDelete(n)
	for name, compute := range map[string]func(*abc.Network) *vec.Vec_Ptr_t{
		"smart": Abc_NtkComputeSupportsSmart,
		"naive": Abc_NtkComputeSupportsNaive,
	} {
		got := compute(n)
		if vec.Vec_PtrSize(got) != 1 {
			t.Fatalf("%s support count = %d, want 1", name, vec.Vec_PtrSize(got))
		}
		support := vec.Vec_PtrEntry(got, 0).(*vec.Vec_Int_t)
		if vec.Vec_IntSize(support) != 3 || vec.Vec_IntEntry(support, 0) != 0 || vec.Vec_IntEntry(support, 1) != 1 || vec.Vec_IntEntry(support, 2) != 0 {
			t.Fatalf("%s support = %v, want [0 1 0] (two CIs and output index)", name, vec.Vec_IntArray(support))
		}
		freeIntVectorList(got)
	}
}

func TestPartitionSmartFindPartUsesSharedSupportAndLimit(t *testing.T) {
	parts := vec.Vec_PtrAlloc(2)
	supports := vec.Vec_PtrAlloc(2)
	chars := vec.Vec_PtrAlloc(2)
	for _, values := range [][]int32{{0, 1, 2}, {3, 4}} {
		v := vec.Vec_IntAlloc(len(values))
		for _, x := range values {
			vec.Vec_IntPush(v, x)
		}
		vec.Vec_PtrPush(supports, v)
		vec.Vec_PtrPush(parts, vec.Vec_IntAlloc(1))
		bits := make([]uint32, 1)
		for _, x := range values {
			bits[x/32] |= 1 << uint(x%32)
		}
		vec.Vec_PtrPush(chars, bits)
	}
	query := vec.Vec_IntAlloc(2)
	vec.Vec_IntPush(query, 1)
	vec.Vec_IntPush(query, 3)
	if got := Abc_NtkPartitionSmartFindPart(supports, parts, chars, 0, query); got != 0 {
		t.Fatalf("best matching partition = %d, want 0", got)
	}
	if got := Abc_NtkPartitionSmartFindPart(supports, parts, chars, 1, query); got != -1 {
		t.Fatalf("size-limited partition = %d, want -1", got)
	}
	freeIntVectorList(parts)
	freeIntVectorList(supports)
	vec.Vec_PtrFree(chars)
	vec.Vec_IntFree(query)
}
