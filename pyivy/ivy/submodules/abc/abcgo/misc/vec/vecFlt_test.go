package vec

import (
	"math"
	"testing"
)

func TestFloatVector(t *testing.T) {
	p := Vec_FltAlloc(1)
	for i := 0; i < 100; i++ {
		Vec_FltPush(p, float32(i))
	}
	if Vec_FltSize(p) != 100 || Vec_FltCap(p) < 100 {
		t.Fatal("growth")
	}
	q := Vec_FltDup(p)
	Vec_FltWriteEntry(q, 0, 1000)
	if Vec_FltEntry(p, 0) != 0 {
		t.Fatal("copy alias")
	}
	Vec_FltSort(q, 1)
	if Vec_FltPop(q) != 1 {
		t.Fatal("descending sort")
	}
	if Vec_FltPushUnique(p, 50) != 1 || Vec_FltPushUnique(p, 101) != 0 {
		t.Fatal("unique")
	}
	if Vec_FltRemove(p, 50) != 1 || Vec_FltFind(p, 50) != -1 {
		t.Fatal("remove")
	}
	Vec_FltFill(p, 0, 0)
	Vec_FltPushOrder(p, 3)
	Vec_FltPushOrder(p, 1)
	Vec_FltPushOrder(p, 2)
	if Vec_FltEntry(p, 0) != 1 || Vec_FltEntryLast(p) != 3 {
		t.Fatal("ordered insert")
	}
	moved := Vec_FltDupArray(p)
	if Vec_FltSize(p) != 0 || Vec_FltSize(moved) != 3 {
		t.Fatal("ownership transfer")
	}
	full := Vec_FltStartFull(1)
	if math.Float32bits(Vec_FltEntry(full, 0)) != 0xffffffff {
		t.Fatal("full bits")
	}
	Vec_FltFillExtra(moved, 20, 9)
	if Vec_FltEntry(moved, 19) != 9 {
		t.Fatal("fill extra")
	}
}
