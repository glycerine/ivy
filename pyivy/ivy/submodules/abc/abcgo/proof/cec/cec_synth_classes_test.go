package cec

import (
	"github.com/glycerine/abcgo/aig/gia"
	"testing"
)

func TestTransferAndNormalizePartitionClasses(t *testing.T) {
	part := gia.Gia_ManStart(10)
	for i := 0; i < 4; i++ {
		gia.Gia_ManAppendCi(part)
	}
	part.StartRepresentatives()
	gia.Gia_ObjSetRepr(part, 2, 1)
	gia.Gia_ObjSetRepr(part, 4, 2)
	reprs := []int32{-1, -1, -1, -1, -1, -1, -1}
	// Partition object order differs from the original object order.
	if got := Gia_TransferMappedClasses(part, []int32{0, 2, 4, 1, 6}, reprs); got != 2 {
		t.Fatalf("transferred %d classes", got)
	}
	graph := gia.Gia_ManStart(10)
	for i := 0; i < 6; i++ {
		gia.Gia_ManAppendCi(graph)
	}
	Gia_ManNormalizeEquivalences(graph, reprs)
	for _, id := range []int32{4, 6} {
		if got := gia.Gia_ObjRepr(graph, id); got != 2 {
			t.Errorf("representative of %d = %d, want 2", id, got)
		}
	}
	if graph.Next(2) != 4 || graph.Next(4) != 6 || graph.Next(6) != 0 {
		t.Fatalf("bad linked class: %d %d %d", graph.Next(2), graph.Next(4), graph.Next(6))
	}
}

// cecSynth.c:212-225 skips mapped self-equivalences and transfers constants
// to representative zero before normalizing chains from different partitions.
func TestTransferPartitionConstantsAndDuplicateMapping(t *testing.T) {
	part := gia.Gia_ManStart(8)
	for i := 0; i < 4; i++ {
		gia.Gia_ManAppendCi(part)
	}
	part.StartRepresentatives()
	gia.Gia_ObjSetRepr(part, 1, 0)
	gia.Gia_ObjSetRepr(part, 2, 1)
	gia.Gia_ObjSetRepr(part, 4, 3)
	reprs := []int32{-1, -1, -1, -1, -1, -1}
	if n := Gia_TransferMappedClasses(part, []int32{0, 2, 2, 1, 5}, reprs); n != 2 {
		t.Fatalf("transferred %d", n)
	}
	if reprs[2] != 0 || reprs[5] != 1 {
		t.Fatalf("representatives %v", reprs)
	}
	graph := gia.Gia_ManStart(8)
	for i := 0; i < 5; i++ {
		gia.Gia_ManAppendCi(graph)
	}
	Gia_ManNormalizeEquivalences(graph, reprs)
	if gia.Gia_ObjRepr(graph, 2) != 0 || gia.Gia_ObjRepr(graph, 5) != 1 {
		t.Fatalf("constant %d pair %d", gia.Gia_ObjRepr(graph, 2), gia.Gia_ObjRepr(graph, 5))
	}
}
