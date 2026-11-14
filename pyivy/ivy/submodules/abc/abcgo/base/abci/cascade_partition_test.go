package abci

import (
	"reflect"
	"testing"
)

func TestSourceCascadePartitionLayouts(t *testing.T) {
	a, b := make([]uint32, 3), make([]uint32, 3)
	Abc_ResStartPart(8, a, 3)
	Abc_ResStartPart2(8, b, 3)
	if !reflect.DeepEqual(a, []uint32{7, 56, 192}) || !reflect.DeepEqual(b, []uint32{73, 146, 36}) {
		t.Fatalf("contiguous=%v striped=%v", a, b)
	}
}
func TestSourceCascadeNonStrictPatterns(t *testing.T) {
	p := []int8{0, 0, 1, 1}
	if got := Abc_ResCheckNonStrict(p, 2, 1); got != 1 {
		t.Fatalf("one-bit count=%d", got)
	}
	if got := Abc_ResCheckNonStrict(p, 2, 2); got != 2 {
		t.Fatalf("two-bit count=%d", got)
	}
	if Abc_ResCheckUnique([]int8{-1, 2}, 2, -1) != 0 || Abc_ResCheckUnique([]int8{-1, 2}, 2, 255) != 1 {
		t.Fatal("signed char comparisons differ")
	}
}
