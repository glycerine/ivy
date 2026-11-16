package gia

import (
	"slices"
	"testing"

	"github.com/glycerine/abcgo/misc/vec"
)

func TestIsoSortTieOrderFeedsSourceCanonicalPermutation(t *testing.T) {
	graph := iso2ReviewInputs(11)
	p := Gia_IsoManStart(graph)
	defer Gia_IsoManStop(p)
	Gia_IsoPrepare(p)
	Gia_IsoAssignUnique(p)
	for i := int32(0); i < 11; i++ {
		Gia_ManCi(graph, i).Value = uint32(2 - i%2)
	}
	if Gia_IsoSort(p) != 1 {
		t.Fatal("the two signature groups were not refined")
	}
	// Hand-traced source QuickSort3->QuickSort2 partition: the five key1
	// items precede the six key2 items. This order determines tie labels.
	want := []int32{10, 2, 4, 8, 6, 9, 11, 5, 3, 1, 7}
	if !slices.Equal(vec.Vec_IntArray(p.vClasses), []int32{1, 5, 6, 6}) {
		t.Fatalf("class boundaries%v", vec.Vec_IntArray(p.vClasses))
	}
	for i, id := range want {
		if got := Gia_IsoGetItem(p, int32(i+1)); got != uint32(id) {
			t.Fatalf("source tie item%d got%d want%d", i, got, id)
		}
	}
	Gia_IsoAssignOneClass(p, 0)
	Gia_IsoManTransferUnique(p)
	cis, ands, cos := vec.Vec_IntAlloc(11), vec.Vec_IntAlloc(0), vec.Vec_IntAlloc(1)
	var permutation *Vec_Int_t
	Gia_ManFindCaninicalOrder(graph, cis, ands, cos, &permutation)
	wantPermutation := make([]int32, len(want))
	for i, id := range want {
		wantPermutation[i] = id - 1
	}
	if !slices.Equal(vec.Vec_IntArray(cis), want) || !slices.Equal(vec.Vec_IntArray(permutation), wantPermutation) {
		t.Fatalf("source canonical inputs%v permutation%v", vec.Vec_IntArray(cis), vec.Vec_IntArray(permutation))
	}
}
