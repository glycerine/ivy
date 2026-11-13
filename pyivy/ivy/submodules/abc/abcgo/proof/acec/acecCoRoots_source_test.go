package acec

import (
	"slices"
	"testing"

	"github.com/glycerine/abcgo/misc/vec"
)

// The source acecCo.c root collector descends through carry outputs only if
// the matching sum is not marked; zero denotes an absent adder input.
func TestSourceAdderRootCollectionMarkedSum(t *testing.T) {
	adds := vec.Vec_IntAllocArrayCopy([]int32{
		1, 2, 3, 4, 5, 0,
		5, 6, 0, 8, 9, 0,
	}, 12)
	mapping := Gia_PolynComputeMap(adds, 10)
	marks := vec.Vec_BitStart(10)
	roots := vec.Vec_IntAlloc(4)
	Gia_PolyCollectRoots(adds, mapping, marks, 1, roots)
	if got := vec.Vec_IntArray(roots); !slices.Equal(got, []int32{8, 9, 4}) {
		t.Fatalf("unmarked roots %v", got)
	}
	vec.Vec_BitWriteEntry(marks, 4, 1)
	Gia_PolyCollectRoots(adds, mapping, marks, 1, roots)
	if got := vec.Vec_IntArray(roots); !slices.Equal(got, []int32{8, 9}) {
		t.Fatalf("marked roots %v", got)
	}
}
