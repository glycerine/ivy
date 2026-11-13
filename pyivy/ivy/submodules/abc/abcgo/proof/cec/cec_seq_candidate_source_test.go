package cec

import (
	"github.com/glycerine/abcgo/aig/gia"
	"testing"
)

// cecSeq.c marks PO drivers during the candidate scan, then clears those
// marks even when they were set before the call. The early no-classes return
// precedes all marking.
func TestSeqCandidateScanSourceMarkSideEffects(t *testing.T) {
	g := gia.Gia_ManStart(12)
	a, b := gia.Gia_ManAppendCi(g), gia.Gia_ManAppendCi(g)
	x := gia.Gia_ManAppendAnd(g, a, b)
	y := gia.Gia_ManAppendAnd(g, a, b^1)
	gia.Gia_ManAppendCo(g, x)
	driver := gia.Gia_ManObj(g, x>>1)
	gia.Gia_ObjSetMark0(driver, 1)
	if got := Cec_ManCheckNonTrivialCands(g); got != 0 || gia.Gia_ObjMark0(driver) != 1 {
		t.Fatal("no-class early return changed mark")
	}
	g.StartRepresentatives()
	gia.Gia_ObjSetRepr(g, x>>1, a>>1)
	if got := Cec_ManCheckNonTrivialCands(g); got != 0 || gia.Gia_ObjMark0(driver) != 0 {
		t.Fatalf("PO-only candidate: %d mark %d", got, gia.Gia_ObjMark0(driver))
	}
	gia.Gia_ObjSetRepr(g, y>>1, a>>1)
	gia.Gia_ObjSetMark0(driver, 1)
	if got := Cec_ManCheckNonTrivialCands(g); got != 1 || gia.Gia_ObjMark0(driver) != 0 {
		t.Fatalf("internal candidate: %d mark %d", got, gia.Gia_ObjMark0(driver))
	}
}
