package bmc

import (
	"github.com/glycerine/abcgo/misc/vec"
	"testing"
)

func TestMaj3Value(t *testing.T) {
	for n := int32(1); n <= 9; n++ {
		for m := int32(0); m < 1<<n; m++ {
			count := int32(0)
			for i := int32(0); i < n; i++ {
				count += (m >> i) & 1
			}
			want := int32(0)
			if count > n/2 {
				want = 1
			}
			if Maj3_ManValue(m, n) != want || Zyx_ManValue(m, n) != want {
				t.Fatalf("n=%d m=%d", n, m)
			}
		}
	}
}
func TestMaj3Levels(t *testing.T) {
	v := vec.Vec_IntAllocArrayCopy([]int32{1, 2, 1}, 3)
	first := make([]int32, 4)
	levels := make([]int32, 7)
	Maj3_ManFirstAndLevel(v, first, levels, 3, 7)
	for i, w := range []int32{0, 3, 4, 6} {
		if first[i] != w {
			t.Fatal(first)
		}
	}
	for i, w := range []int32{0, 0, 0, 1, 2, 2, 3} {
		if levels[i] != w {
			t.Fatal(levels)
		}
	}
	if vec.Vec_IntEntry(v, 0) != 1 || vec.Vec_IntEntry(v, 1) != 2 || vec.Vec_IntEntry(v, 2) != 1 {
		t.Fatal("levels changed")
	}
}
func TestZyxSymmetry(t *testing.T) {
	// Three-input majority is symmetric under any exchange.
	truth := []uint64{0xe8e8e8e8e8e8e8e8}
	pairs := vec.Vec_IntArray(Zyx_ManCreateSymVarPairs(truth, 3))
	want := []int32{0, 1, 0, 2, 1, 2}
	if len(pairs) != len(want) {
		t.Fatalf("pairs %v", pairs)
	}
	for i := range want {
		if pairs[i] != want[i] {
			t.Fatalf("pairs %v", pairs)
		}
	}
}

func TestMaj3ExactThreeGateMajority(t *testing.T) {
	// Standard 5-input majority: two bottom gates, two second-level gates,
	// and a top gate. ABC's level vector is top-first.
	levels := vec.Vec_IntAllocArrayCopy([]int32{1, 3, 1}, 3)
	if Maj3_ManExactSynthesis(5, 5, 0, levels) != 1 {
		t.Fatal("failed to synthesize five-input majority")
	}
}

func TestMaj3SevenInputSynthesis(t *testing.T) {
	levels := vec.Vec_IntAllocArrayCopy([]int32{1, 2, 2, 2}, 4)
	if Maj3_ManExactSynthesis(7, 7, 0, levels) != 1 {
		t.Fatal("failed to synthesize seven-input majority")
	}
}
