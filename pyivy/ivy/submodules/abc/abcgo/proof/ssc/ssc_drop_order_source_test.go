package ssc

import (
	"reflect"
	"testing"

	"github.com/glycerine/abcgo/aig/gia"
)

// In sscUtil.c each removed output is replaced by the constant-true SAT
// literal. Identical outputs are therefore removed from left to right, with
// the final implementation retained rather than an arbitrary representative.
func TestDropContainedKeepsLastEquivalentOutput(t *testing.T) {
	g := gia.Gia_ManStart(20)
	a, b, c := gia.Gia_ManAppendCi(g), gia.Gia_ManAppendCi(g), gia.Gia_ManAppendCi(g)
	ab := gia.Gia_ManAppendAnd(g, a, b)
	first := gia.Gia_ManAppendAnd(g, ab, c)
	bc := gia.Gia_ManAppendAnd(g, b, c)
	last := gia.Gia_ManAppendAnd(g, a, bc)
	gia.Gia_ManAppendCo(g, first)
	gia.Gia_ManAppendCo(g, 0)
	gia.Gia_ManAppendCo(g, last)
	result := Gia_ManDropContained(g)
	if result == nil {
		t.Fatal("containment failed")
	}
	expected := gia.Gia_ManDupCones(g, []int32{2}, 1, 0)
	firstOnly := gia.Gia_ManDupCones(g, []int32{0}, 1, 0)
	rootLast := expected.OutputLiterals()[0] >> 1
	rootFirst := firstOnly.OutputLiterals()[0] >> 1
	lastA, lastB, _, _ := expected.Gate(rootLast)
	firstA, firstB, _, _ := firstOnly.Gate(rootFirst)
	if lastA == firstA && lastB == firstB {
		t.Fatal("fixture does not distinguish the two implementations")
	}
	if !reflect.DeepEqual(result.OutputLiterals(), expected.OutputLiterals()) || result.ObjectCount() != expected.ObjectCount() {
		t.Fatal("did not retain last output cone")
	}
	for id := int32(0); int(id) < result.ObjectCount(); id++ {
		a0, b0, c0, k0 := result.Gate(id)
		a1, b1, c1, k1 := expected.Gate(id)
		if a0 != a1 || b0 != b1 || c0 != c1 || k0 != k1 {
			t.Fatalf("object %d differs from last output cone", id)
		}
	}
}
