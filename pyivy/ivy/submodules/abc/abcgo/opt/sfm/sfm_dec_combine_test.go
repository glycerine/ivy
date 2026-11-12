package sfm

import (
	"slices"
	"testing"
)

func TestCombineDecomposition(t *testing.T) {
	a := []uint64{0xaaaaaaaaaaaaaaaa}
	b := []uint64{0xaaaaaaaaaaaaaaaa}
	truth, support := sfmCombineDecomposition(a, b, []int32{1}, []int32{3}, 2, 6)
	if !slices.Equal(support, []int32{1, 2, 3}) {
		t.Fatalf("support %v", support)
	}
	for x := 0; x < 64; x++ {
		want := x&1 != 0
		if x&2 != 0 {
			want = x&4 != 0
		}
		if (truth[0]>>uint(x)&1 != 0) != want {
			t.Fatalf("assignment %d", x)
		}
	}
	same, sameSupport := sfmCombineDecomposition(a, a, []int32{1}, []int32{1}, 2, 6)
	if !slices.Equal(same, a) || !slices.Equal(sameSupport, []int32{1}) {
		t.Fatal("identical cofactors")
	}
	if truth, _ := sfmCombineDecomposition(a, b, []int32{1}, []int32{3}, 2, 2); truth != nil {
		t.Fatal("accepted excess support")
	}
}
