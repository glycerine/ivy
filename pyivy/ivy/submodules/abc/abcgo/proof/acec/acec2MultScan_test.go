package acec

import (
	"github.com/glycerine/abcgo/misc/vec"
	"testing"
)

func TestCutDifferenceAndContainment(t *testing.T) {
	cuts := vec.Vec_WecStart(2)
	for _, v := range []int32{1, 2, 3, 4, 5} {
		vec.Vec_WecPush(cuts, 0, v)
	}
	for _, v := range []int32{1, 2, 3, 4, 6} {
		vec.Vec_WecPush(cuts, 1, v)
	}
	if got := Sdb_StoDiffExactlyOne(cuts, 2, []int32{4, 1, 2, 3, 7}); got != 7 {
		t.Fatalf("difference %d", got)
	}
	if got := Sdb_StoDiffExactlyOne(cuts, 2, []int32{4, 1, 2, 3, 4}); got != -1 {
		t.Fatalf("contained result %d", got)
	}
	var count int32
	if got := Sdb_StoDiffExactlyOne3(cuts, 2, []int32{4, 1, 2, 3, 7}, &count); got != 7 || count != 2 {
		t.Fatalf("difference %d count %d", got, count)
	}
	if got := Sdb_StoFindAll(cuts); vec.Vec_IntSize(got) != 6 {
		t.Fatalf("unique inputs %v", vec.Vec_IntArray(got))
	}
}
func TestCutPatternScan(t *testing.T) {
	p := &Sdb_Sto_t{vCuts: vec.Vec_WecStart(1), vTtMem: vec.Vec_MemAllocForTT(6, 0)}
	id := vec.Vec_MemHashInsert(p.vTtMem, []uint64{uint64(s_FuncTruths4[0])})
	v := vec.Vec_WecEntry(p.vCuts, 0)
	for _, x := range []int32{1, 4, 1, 2, 3, 4, int32(id * 2)} {
		vec.Vec_IntPush(v, x)
	}
	var cut []int32
	if got := Sdb_StoIterCutsOne(p, 0, 4, &cut); got != 2 || len(cut) != 6 || cut[4] != 4 {
		t.Fatalf("pattern=%d cut=%v", got, cut)
	}
	if got := Sdb_StoIterCutsOne(p, 0, 5, &cut); got != 0 {
		t.Fatalf("unexpected size match %d", got)
	}
}
