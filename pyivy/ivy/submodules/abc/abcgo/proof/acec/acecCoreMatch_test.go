package acec

import (
	"github.com/glycerine/abcgo/misc/vec"
	"testing"
)

func TestMatchingCopiesLiteralPolarity(t *testing.T) {
	levels := vec.Vec_WecStart(1)
	vec.Vec_WecPush(levels, 0, 2)
	vec.Vec_WecPush(levels, 0, 5)
	mapping := vec.Vec_IntStart(3)
	vec.Vec_IntWriteEntry(mapping, 1, 10)
	vec.Vec_IntWriteEntry(mapping, 2, 12)
	got := vec.Vec_IntArray(vec.Vec_WecEntry(Acec_MatchCopy(levels, mapping), 0))
	if len(got) != 2 || got[0] != 10 || got[1] != 13 {
		t.Fatalf("mapped %v", got)
	}
}
func TestMoveDuplicateClasses(t *testing.T) {
	lits, classes := vec.Vec_WecStart(2), vec.Vec_WecStart(2)
	for _, x := range []int32{20, 22, 24, 26} {
		vec.Vec_WecPush(lits, 0, x)
	}
	for _, x := range []int32{6, 6, 4, 4} {
		vec.Vec_WecPush(classes, 0, x)
	}
	Acec_MoveDuplicates(lits, classes)
	if vec.Vec_IntSize(vec.Vec_WecEntry(lits, 0)) != 0 {
		t.Fatal("duplicate inputs remain")
	}
	got := vec.Vec_IntArray(vec.Vec_WecEntry(classes, 1))
	if len(got) != 2 || got[0] != 6 || got[1] != 4 {
		t.Fatalf("moved classes %v", got)
	}
	gl := vec.Vec_IntArray(vec.Vec_WecEntry(lits, 1))
	if gl[0] != 22 || gl[1] != 26 {
		t.Fatalf("moved literals %v", gl)
	}
}
func TestMatchShiftCounts(t *testing.T) {
	a, b := vec.Vec_WecStart(1), vec.Vec_WecStart(2)
	for _, x := range []int32{8, 6, 4} {
		vec.Vec_WecPush(a, 0, x)
		vec.Vec_WecPush(b, 1, x)
	}
	if got := Acec_MatchCountCommon(a, b, 1); got != 3 {
		t.Fatalf("common %d", got)
	}
	if got := Acec_MatchCountCommon(a, b, 0); got != 0 {
		t.Fatalf("unshifted common %d", got)
	}
}
