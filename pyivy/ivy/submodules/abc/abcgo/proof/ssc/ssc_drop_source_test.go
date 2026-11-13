package ssc

import (
	"testing"

	"github.com/glycerine/abcgo/aig/gia"
)

// Gia_ManDropContained in sscUtil.c derives CNF directly and never starts
// SSC or modifies the caller's representative classes.
func TestDropContainedKeepsExistingClasses(t *testing.T) {
	g := gia.Gia_ManStart(12)
	a, b := gia.Gia_ManAppendCi(g), gia.Gia_ManAppendCi(g)
	n := gia.Gia_ManAppendAnd(g, a, b)
	gia.Gia_ManAppendCo(g, a)
	gia.Gia_ManAppendCo(g, n)
	g.StartRepresentatives()
	gia.Gia_ObjSetRepr(g, n>>1, a>>1)
	gia.Gia_ObjSetProved(g, n>>1)
	result := Gia_ManDropContained(g)
	if result == nil {
		t.Fatal("containment failed")
	}
	if got := gia.Gia_ObjRepr(g, n>>1); got != a>>1 {
		t.Fatalf("representative changed to %d", got)
	}
	if gia.Gia_ObjProved(g, n>>1) != 1 {
		t.Fatal("proved flag changed")
	}
	if got := gia.Gia_ManPoNum(result); got != 1 {
		t.Fatalf("output count = %d, want 1", got)
	}
	for bits := 0; bits < 4; bits++ {
		input := []bool{bits&1 != 0, bits&2 != 0}
		out, err := result.Evaluate(input)
		if err != nil || len(out) != 1 || out[0] != input[0] {
			t.Fatalf("input %v: %v, %v", input, out, err)
		}
	}
}
