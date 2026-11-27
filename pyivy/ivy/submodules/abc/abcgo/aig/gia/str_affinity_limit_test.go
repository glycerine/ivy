package gia

import (
	"strings"
	"testing"
	"time"

	"github.com/glycerine/abcgo/misc/vec"
)

// TestStrAffinityLimitStopsOnlyFaninLoop: in Str_ManVectorAffinity, reaching
// nLimit while collecting an AND's fanins breaks only the fanin loop; the
// later supergate entries still OR their rows into fanins already
// collected. The port stopped the whole collection, losing those rows.
func TestStrAffinityLimitStopsOnlyFaninLoop(t *testing.T) {
	p := Gia_ManStart(16)
	a, b, c, d := Gia_ManAppendCi(p), Gia_ManAppendCi(p), Gia_ManAppendCi(p), Gia_ManAppendCi(p)
	n1, n2, n3 := Gia_ManAppendAnd(p, a, b), Gia_ManAppendAnd(p, a, c), Gia_ManAppendAnd(p, c, d)
	Gia_ManAppendCo(p, n3)
	Gia_ManHashStart(p)
	vec.Vec_IntFill(&p.vCopies, int(Gia_ManObjNum(p)), -1)
	super := vec.Vec_IntAlloc(8)
	for _, lit := range []int32{n1, n2, n3} {
		vec.Vec_IntPush(super, lit)
	}
	delays := vec.Vec_IntStart(int(Gia_ManObjNum(p)))
	matrix := make([]uint64, 256)
	// limit 4: a is collected (row n1|n2), then b and c hit the limit
	if got := Str_ManVectorAffinity(p, super, delays, matrix, 4); got != 1 {
		t.Fatalf("affinity rows = %d, want 1 (fanin a shared by n1 and n2)", got)
	}
}

// TestStrNtkPsLikeNative: Str_NtkPs prints the counts and then the time with
// Abc_PrintTime ("  Time =     0.00 sec").
func TestStrNtkPsLikeNative(t *testing.T) {
	p := &Str_Ntk_t{}
	out := isoReviewCapture(t, func() { Str_NtkPs(p, time.Now()) })
	if !strings.HasPrefix(out, "Network contains 0 ands, 0 xors, 0 muxes (0 trees in 0 groups).  Time =") || !strings.HasSuffix(out, " sec\n") {
		t.Fatalf("got %q", out)
	}
}
