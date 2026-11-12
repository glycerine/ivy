package extra

import (
	"github.com/glycerine/abcgo/bdd/cudd/core"
	"math"
	"slices"
	"testing"
)

func unpackZddSubsets(t *testing.T, packed []int32) []uint32 {
	t.Helper()
	result := make([]uint32, 0, packed[0])
	at := 1
	for i := int32(0); i < packed[0]; i++ {
		var mask uint32
		for packed[at] != -1 {
			mask |= 1 << packed[at]
			at++
		}
		at++
		result = append(result, mask)
	}
	if at != len(packed) {
		t.Fatal("packed subset length")
	}
	return result
}
func TestZddTruncateAllThreeVariableFamilies(t *testing.T) {
	dd := core.Cudd_Init(0, 3, 256, 256, 0)
	defer core.Cudd_Quit(dd)
	probabilities := []float64{0.25, 0.5, 1}
	for family := uint32(0); family < 256; family++ {
		var subsets [][]int32
		for mask := uint32(0); mask < 8; mask++ {
			if family>>mask&1 == 0 {
				continue
			}
			var row []int32
			for v := int32(0); v < 3; v++ {
				if mask>>v&1 != 0 {
					row = append(row, v)
				}
			}
			row = append(row, -1)
			subsets = append(subsets, row)
		}
		function := Extra_zddCreateSubsets(dd, subsets, int32(len(subsets)))
		for _, limit := range []float64{0.1, 0.125, 0.25, 0.5, 1} {
			got := unpackZddSubsets(t, Extra_zddTruncate(dd, function, probabilities, limit))
			var want []uint32
			for _, mask := range []uint32{0, 4, 2, 6, 1, 5, 3, 7} {
				if family>>mask&1 == 0 {
					continue
				}
				product := 1.0
				for v, p := range probabilities {
					if mask>>v&1 != 0 {
						product *= p
					}
				}
				if product >= limit {
					want = append(want, mask)
				}
			}
			if !slices.Equal(got, want) {
				t.Fatalf("family%x limit%g got%v want%v", family, limit, got, want)
			}
		}
	}
}
func TestZddTruncateOriginalExample(t *testing.T) {
	dd := core.Cudd_Init(0, 10, 256, 256, 0)
	defer core.Cudd_Quit(dd)
	function := Extra_zddCreateSubsets(dd, [][]int32{{0, 3, 5, -1}, {1, 2, 3, 6, 9, -1}, {1, 5, 7, 8, -1}, {2, 4, -1}, {0, 5, 6, 9, -1}}, 5)
	p := make([]float64, 10)
	for i := range p {
		p[i] = 0.1
	}
	got := unpackZddSubsets(t, Extra_zddTruncate(dd, function, p, 0.001))
	slices.Sort(got)
	if !slices.Equal(got, []uint32{20, 41}) {
		t.Fatal(got)
	}
}
func TestZddTruncateInvalidProbabilities(t *testing.T) {
	dd := core.Cudd_Init(0, 1, 256, 256, 0)
	defer core.Cudd_Quit(dd)
	for _, p := range []float64{0, -1, 1.1, math.NaN(), math.Inf(1)} {
		func() {
			defer func() {
				if recover() == nil {
					t.Errorf("accepted probability %g", p)
				}
			}()
			Extra_zddTruncate(dd, core.Cudd_ReadOne(dd), []float64{p}, 0.1)
		}()
	}
}
