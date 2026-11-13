package ssw

import (
	"reflect"
	"testing"

	"github.com/glycerine/abcgo/aig/aig"
	"github.com/glycerine/abcgo/misc/vec"
)

// sswPairs.c regularizes each source object's pData, skips identical copies,
// and orders each remaining pair by destination object ID.
func TestTransferSignalPairsReadsSourceCopyMetadata(t *testing.T) {
	first, second, miter := aig.NewGraph(), aig.NewGraph(), aig.NewGraph()
	a, b, c := first.NewInput(), first.NewInput(), first.NewInput()
	x, y, z := second.NewInput(), second.NewInput(), second.NewInput()
	low, same, high := miter.NewInput(), miter.NewInput(), miter.NewInput()
	aig.Aig_ObjSetCopy(first, a, high.Not())
	aig.Aig_ObjSetCopy(second, x, low)
	aig.Aig_ObjSetCopy(first, b, same)
	aig.Aig_ObjSetCopy(second, y, same.Not())
	aig.Aig_ObjSetCopy(first, c, low)
	aig.Aig_ObjSetCopy(second, z, high.Not())
	left, right := vec.Vec_IntAlloc(3), vec.Vec_IntAlloc(3)
	for _, e := range []aig.Edge{a, b, c} {
		vec.Vec_IntPush(left, int32(e.NodeID()))
	}
	for _, e := range []aig.Edge{x, y, z} {
		vec.Vec_IntPush(right, int32(e.NodeID()))
	}
	got := Ssw_TransferSignalPairs(miter, first, second, left, right)
	want := []int32{int32(low.NodeID()), int32(high.NodeID()), int32(low.NodeID()), int32(high.NodeID())}
	if !reflect.DeepEqual(vec.Vec_IntArray(got), want) {
		t.Fatalf("transferred pairs = %v, want %v", vec.Vec_IntArray(got), want)
	}
}

func TestPairClassesMatchSourceBoundedCases(t *testing.T) {
	cases := []struct{ pairs, want []int32 }{
		{[]int32{1, 1}, []int32{1, 1}},
		{[]int32{3, 5, 1, 5}, []int32{1, 3, 5}},
		{[]int32{1, 2, 3, 4, 2, 4}, []int32{1, 2, 3, 4}},
		{[]int32{3, 5, 1, 2, 2, 5}, []int32{1, 2, 3, 5}},
	}
	for _, tc := range cases {
		pairs := vec.Vec_IntAlloc(len(tc.pairs))
		for _, v := range tc.pairs {
			vec.Vec_IntPush(pairs, v)
		}
		classes := Ssw_TransformPairsIntoTempClasses(pairs, 12)
		if classes[1] == nil || !reflect.DeepEqual(vec.Vec_IntArray(classes[1]), tc.want) {
			t.Errorf("pairs %v: class %v, want %v", tc.pairs, classes[1], tc.want)
		}
	}
}

func TestPairClassesSourceIndexReuseCanMisalignPairs(t *testing.T) {
	pairs := vec.Vec_IntAlloc(24)
	for _, x := range []int32{6, 8, 3, 8, 5, 8, 2, 6, 1, 5, 1, 5, 3, 8, 2, 7, 4, 9, 1, 5, 1, 8, 2, 6} {
		vec.Vec_IntPush(pairs, x)
	}
	defer func() {
		if recover() == nil {
			t.Error("expected source assertion on misaligned pair")
		}
	}()
	Ssw_TransformPairsIntoTempClasses(pairs, 12)
}
