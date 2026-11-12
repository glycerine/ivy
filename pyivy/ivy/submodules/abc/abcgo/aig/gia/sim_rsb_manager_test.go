package gia

import (
	"github.com/glycerine/abcgo/misc/vec"
	"testing"
)

func TestResubstitutionPartitionUpdates(t *testing.T) {
	sims, transposed := vec.Vec_WrdStart(64), vec.Vec_WrdStart(64)
	vec.Vec_WrdWriteEntry(sims, 0, 0xc)
	vec.Vec_WrdWriteEntry(sims, 1, 0xa)
	cands := vec.Vec_IntAlloc(2)
	vec.Vec_IntPush(cands, 0)
	vec.Vec_IntPush(cands, 1)
	p := Gia_RsbAlloc(nil, []uint64{3}, []uint64{12}, sims, 1, transposed, 1, cands)
	defer Gia_RsbFree(p)
	if Gia_RsbCost(p) != 4 {
		t.Fatal("incorrect initial cost")
	}
	Gia_RsbUpdateAdd(p, 0)
	if Gia_RsbCost(p) != 0 || Gia_RsbRemovalCost(p, 0) != 4 {
		t.Fatal("incorrect partition or removal cost")
	}
	Gia_RsbUpdateAdd(p, 1)
	var cost int32
	if Gia_RsbFindNodeToRemove(p, &cost) != 1 || cost != 0 {
		t.Fatalf("wrong removable candidate cost %d", cost)
	}
	Gia_RsbUpdateRemove(p, 1)
	if Gia_RsbCost(p) != 0 {
		t.Fatal("remove changed cost")
	}
	Gia_RsbUpdateRemove(p, 0)
	if Gia_RsbCost(p) != 4 {
		t.Fatal("original partition not restored")
	}
}
