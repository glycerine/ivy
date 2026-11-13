package abs

import (
	"github.com/glycerine/abcgo/aig/gia"
	"github.com/glycerine/abcgo/misc/vec"
	"testing"
)

func TestSpecialTruthOrdersLiteralsAndFoldsConstants(t *testing.T) {
	p := gia.Gia_ManStart(8)
	a := gia.Gia_ManAppendCi(p)
	b := gia.Gia_ManAppendCi(p)
	root := gia.Gia_ManAppendAnd(p, a, b)
	leaves := vec.Vec_IntAlloc(2)
	for _, lit := range []int32{a, b} {
		vec.Vec_IntPush(leaves, lit>>1)
		gia.Gia_ObjSetPhaseValue(gia.Gia_ManObj(p, lit>>1), 1)
	}
	lits := vec.Vec_IntAlloc(2)
	vec.Vec_IntPush(lits, 7)
	vec.Vec_IntPush(lits, 4)
	truth := Ga2_ObjComputeTruthSpecial(p, gia.Gia_ManObj(p, root>>1), leaves, lits)
	if truth != 0x22222222 || vec.Vec_IntEntry(lits, 0) != 4 || vec.Vec_IntEntry(lits, 1) != 6 {
		t.Fatalf("truth=%08x lits=%d,%d", truth, vec.Vec_IntEntry(lits, 0), vec.Vec_IntEntry(lits, 1))
	}
	vec.Vec_IntClear(lits)
	vec.Vec_IntPush(lits, 0)
	vec.Vec_IntPush(lits, 4)
	if got := Ga2_ObjComputeTruthSpecial(p, gia.Gia_ManObj(p, root>>1), leaves, lits); got != 0 {
		t.Fatalf("constant-folded truth=%08x", got)
	}
	if gia.Gia_ManObj(p, a>>1).Value != 0 || gia.Gia_ManObj(p, b>>1).Value != 0 {
		t.Fatal("leaf scratch values not cleared")
	}
}
