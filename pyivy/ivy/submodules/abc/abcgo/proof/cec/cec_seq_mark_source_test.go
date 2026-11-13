package cec

import (
	"github.com/glycerine/abcgo/aig/gia"
	"testing"
)

// cecSeq.c marks all PO drivers, considers other existing marks, and clears
// every PO driver's mark even if it was set by the caller.
func TestNonTrivialCandidatesUseSourceScratchMarks(t *testing.T) {
	g := gia.Gia_ManStart(8)
	a := gia.Gia_ManAppendCi(g)
	b := gia.Gia_ManAppendCi(g)
	gia.Gia_ManAppendCo(g, a)
	g.StartRepresentatives()
	gia.Gia_ObjSetRepr(g, b>>1, 0)
	objA, objB := gia.Gia_ManObj(g, a>>1), gia.Gia_ManObj(g, b>>1)
	gia.Gia_ObjSetMark0(objA, 1)
	gia.Gia_ObjSetMark0(objB, 1)
	if got := Cec_ManCheckNonTrivialCands(g); got != 0 {
		t.Fatalf("marked candidate result %d", got)
	}
	if gia.Gia_ObjMark0(objA) != 0 || gia.Gia_ObjMark0(objB) != 1 {
		t.Fatal("source mark cleanup differs")
	}
	gia.Gia_ObjSetMark0(objB, 0)
	if got := Cec_ManCheckNonTrivialCands(g); got != 1 {
		t.Fatalf("unmarked candidate result %d", got)
	}
}
