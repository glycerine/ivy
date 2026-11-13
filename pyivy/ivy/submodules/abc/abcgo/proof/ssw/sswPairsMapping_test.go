package ssw

import (
	"github.com/glycerine/abcgo/aig/aig"
	"github.com/glycerine/abcgo/misc/vec"
	"reflect"
	"testing"
)

func TestTransferSignalPairsMappedUsesRegularOrderedIDs(t *testing.T) {
	first, second := vec.Vec_IntAlloc(3), vec.Vec_IntAlloc(3)
	for _, v := range []int32{1, 2, 3} {
		vec.Vec_IntPush(first, v)
		vec.Vec_IntPush(second, v)
	}
	a := []aig.Edge{aig.InvalidEdge, aig.Edge(14), aig.Edge(5), aig.Edge(11)}
	b := []aig.Edge{aig.InvalidEdge, aig.Edge(9), aig.Edge(4), aig.Edge(16)}
	got := Ssw_TransferSignalPairsMapped(first, second, a, b)
	if !reflect.DeepEqual(vec.Vec_IntArray(got), []int32{4, 7, 5, 8}) {
		t.Fatalf("mapped pairs %v", vec.Vec_IntArray(got))
	}
}
