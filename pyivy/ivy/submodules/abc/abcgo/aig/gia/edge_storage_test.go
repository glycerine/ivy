package gia

import (
	"github.com/glycerine/abcgo/misc/vec"
	"reflect"
	"testing"
)

func TestEdgeAssignmentRoundTripAndPacking(t *testing.T) {
	p := Gia_ManStart(8)
	for i := 0; i < 5; i++ {
		Gia_ManAppendCi(p)
	}
	edges := permutation(1, 3, 2, 3, 4, 5)
	Gia_ManEdgeFromArray(p, edges)
	if Gia_ObjEdgeCount(3, p.VEdge1, p.VEdge2) != 2 || Gia_ObjCheckEdge(p, 3, 1) != 1 {
		t.Fatal("edge adjacency")
	}
	if got := vec.Vec_IntArray(Gia_ManEdgeToArray(p)); !reflect.DeepEqual(got, []int32{1, 3, 2, 3, 4, 5}) {
		t.Fatalf("edge round trip %v", got)
	}
	Gia_ObjEdgeRemove(3, 1, p.VEdge1, p.VEdge2)
	if Gia_ObjEdgeCount(3, p.VEdge1, p.VEdge2) != 1 || Gia_ObjCheckEdge(p, 3, 2) != 1 {
		t.Fatal("edge removal")
	}
	Gia_ObjEdgeClean(3, p.VEdge1, p.VEdge2)
	if Gia_ObjEdgeCount(3, p.VEdge1, p.VEdge2) != 0 {
		t.Fatal("edge clean")
	}
	p.VPacking = permutation(2, 3, 1, 2, 3, 2, 4, 5)
	Gia_ManConvertPackingToEdges(p)
	if got := vec.Vec_IntArray(Gia_ManEdgeToArray(p)); !reflect.DeepEqual(got, []int32{1, 3, 2, 3, 4, 5}) {
		t.Fatalf("packing %v", got)
	}
}
