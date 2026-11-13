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

func TestMatchCheckShiftDirectionsAndNoShift(t *testing.T) {
	newLevels := func(count int, rank int, lit int32) *Vec_Wec_t {
		levels := vec.Vec_WecStart(count)
		vec.Vec_WecPush(levels, rank, lit)
		return levels
	}
	newMap := func() *Vec_Int_t {
		mapping := vec.Vec_IntStart(4)
		for i := int32(0); i < 4; i++ {
			vec.Vec_IntWriteEntry(mapping, int(i), i<<1)
		}
		return mapping
	}

	t.Run("shift_up", func(t *testing.T) {
		left, right := newLevels(1, 0, 2), newLevels(2, 1, 2)
		leftRoots, rightRoots := vec.Vec_WecStart(1), vec.Vec_WecStart(2)
		mapLeft, mapRight := newMap(), newMap()
		defer vec.Vec_WecFree(left)
		defer vec.Vec_WecFree(right)
		defer vec.Vec_WecFree(leftRoots)
		defer vec.Vec_WecFree(rightRoots)
		defer vec.Vec_IntFree(mapLeft)
		defer vec.Vec_IntFree(mapRight)
		Acec_MatchCheckShift(nil, nil, left, right, mapLeft, mapRight, leftRoots, rightRoots)
		if vec.Vec_WecSize(left) != 2 || vec.Vec_IntSize(vec.Vec_WecEntry(left, 0)) != 0 || vec.Vec_WecSize(leftRoots) != 2 {
			t.Fatal("positive rank offset did not add an empty level on the left")
		}
	})
	t.Run("shift_down", func(t *testing.T) {
		left, right := newLevels(2, 1, 2), newLevels(1, 0, 2)
		leftRoots, rightRoots := vec.Vec_WecStart(2), vec.Vec_WecStart(1)
		mapLeft, mapRight := newMap(), newMap()
		defer vec.Vec_WecFree(left)
		defer vec.Vec_WecFree(right)
		defer vec.Vec_WecFree(leftRoots)
		defer vec.Vec_WecFree(rightRoots)
		defer vec.Vec_IntFree(mapLeft)
		defer vec.Vec_IntFree(mapRight)
		Acec_MatchCheckShift(nil, nil, left, right, mapLeft, mapRight, leftRoots, rightRoots)
		if vec.Vec_WecSize(right) != 2 || vec.Vec_IntSize(vec.Vec_WecEntry(right, 0)) != 0 || vec.Vec_WecSize(rightRoots) != 2 {
			t.Fatal("negative rank offset did not add an empty level on the right")
		}
	})
	t.Run("already_aligned", func(t *testing.T) {
		left, right := newLevels(1, 0, 2), newLevels(1, 0, 2)
		leftRoots, rightRoots := vec.Vec_WecStart(1), vec.Vec_WecStart(1)
		mapLeft, mapRight := newMap(), newMap()
		defer vec.Vec_WecFree(left)
		defer vec.Vec_WecFree(right)
		defer vec.Vec_WecFree(leftRoots)
		defer vec.Vec_WecFree(rightRoots)
		defer vec.Vec_IntFree(mapLeft)
		defer vec.Vec_IntFree(mapRight)
		Acec_MatchCheckShift(nil, nil, left, right, mapLeft, mapRight, leftRoots, rightRoots)
		if vec.Vec_WecSize(left) != 1 || vec.Vec_WecSize(right) != 1 || vec.Vec_WecSize(leftRoots) != 1 || vec.Vec_WecSize(rightRoots) != 1 {
			t.Fatal("aligned roots were shifted")
		}
	})
}
