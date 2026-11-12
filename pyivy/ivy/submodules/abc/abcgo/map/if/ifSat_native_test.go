package if_

import (
	"fmt"
	"testing"

	"github.com/glycerine/abcgo/opt/dau"
)

func TestIfSatCheckXYDerivesExactLutTruthTables(t *testing.T) {
	truth := dau.Dau_DsdToTruth("((ab)c)", 3)
	set := uint32(1<<0 | 1<<2)
	var bound, free word
	if If_ManSatCheckXY(If_ManSatBuildXY(2), 2, truth, 3, set, &bound, &free, nil) == 0 {
		t.Fatal("expected factorization for nested AND")
	}
	for i := 0; i < 64; i++ {
		if (bound>>uint(i))&1 != (bound>>uint(i&3))&1 || (free>>uint(i))&1 != (free>>uint(i&3))&1 {
			t.Fatal("LUT truth tables are not stretched to 64 bits")
		}
	}
	var bs, ss, fs []int32
	for v := int32(0); v < 3; v++ {
		switch set >> (2 * v) & 3 {
		case 0:
			fs = append(fs, v)
		case 1:
			bs = append(bs, v)
		case 3:
			ss = append(ss, v)
		}
	}
	got := If_ManSat6Truth(bound, free, bs, ss, fs)
	if got != truth[0] {
		t.Fatalf("reconstructed truth %#x, want %#x", got, truth[0])
	}
}

func TestIfSatCheckXYRejectsThreeFreeCofactors(t *testing.T) {
	var truth word
	for assignment := 0; assignment < 8; assignment++ {
		bound, free := assignment&3, assignment>>2&1
		value := bound == 3 || bound == 1 && free == 1 || bound == 2 && free == 0
		if value {
			truth |= 1 << uint(assignment)
		}
	}
	set := uint32(1<<0 | 1<<2)
	if If_ManSatCheckXY(If_ManSatBuildXY(2), 2, []word{truth}, 3, set, nil, nil, nil) != 0 {
		t.Fatal("expected prime function to reject one-bit intermediate")
	}
}

func TestIfSatCheckXYAllReturnsAValidPartition(t *testing.T) {
	truth := dau.Dau_DsdToTruth("((ab)c)", 3)
	set := If_ManSatCheckXYall(If_ManSatBuildXY(2), 2, truth, 3, nil)
	if set == 0 {
		t.Fatal("expected a valid LUT decomposition")
	}
	if If_ManSatCheckXY(nil, 2, truth, 3, set, nil, nil, nil) == 0 {
		t.Fatalf("returned invalid partition %#x", set)
	}
}

func TestIfSatCheckXYMultiwordCuts(t *testing.T) {
	for n := 7; n <= 11; n++ {
		t.Run(fmt.Sprintf("%d inputs", n), func(t *testing.T) {
			boundPins := n - 5
			truth := make([]word, (1<<uint(n)+63)/64)
			var set uint32
			for i := 0; i < boundPins; i++ {
				set |= 1 << uint(2*i)
			}
			for x := 0; x < 1<<uint(n); x++ {
				bound := x&((1<<uint(boundPins))-1) == (1<<uint(boundPins))-1
				free := (x>>uint(boundPins))&31 == 19
				if bound != free {
					truth[x>>6] |= 1 << uint(x&63)
				}
			}
			found := If_ManSatCheckXYall(nil, 6, truth, int32(n), nil)
			if found == 0 || If_ManSatCheckXY(nil, 6, truth, int32(n), found, nil, nil, nil) == 0 {
				t.Fatalf("search failed for %d inputs: %#x", n, found)
			}
			var b, f word
			if If_ManSatCheckXY(nil, 6, truth, int32(n), set, &b, &f, nil) == 0 {
				t.Fatal("expected two-LUT factorization")
			}
			for x := 0; x < 1<<uint(n); x++ {
				bIndex := x & ((1 << uint(boundPins)) - 1)
				fIndex := int(b>>uint(bIndex)&1) | ((x >> uint(boundPins)) << 1)
				got := f >> uint(fIndex) & 1
				want := truth[x>>6] >> uint(x&63) & 1
				if got != want {
					t.Fatalf("assignment %d: got %d want %d", x, got, want)
				}
			}
		})
	}
}
