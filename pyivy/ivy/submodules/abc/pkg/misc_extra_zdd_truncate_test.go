package abc

// Go-only finite-family and exhaustive-model tests for extraZddTrunc.c.

import (
	"errors"
	"math"
	"slices"
	"testing"
)

func TestExtraZDDConstructionOrderingOwnershipAndVariable(t *testing.T) {
	input := [][]int{{3, 1, 1}, {}, {0, 2}, {1, 3}, {3, 1}}
	zdd, err := NewExtraZDD(4, input, ExtraZDDLimits{})
	if err != nil {
		t.Fatal(err)
	}
	want := [][]int{{}, {1, 3}, {0, 2}}
	if !slices.EqualFunc(zdd.Subsets(), want, slices.Equal) || zdd.Variables() != 4 {
		t.Fatal(zdd.Subsets(), zdd.Variables())
	}
	input[0][0] = 0
	snapshot := zdd.Subsets()
	snapshot[1][0] = 0
	if !slices.Equal(zdd.Subsets()[1], []int{1, 3}) {
		t.Fatal("ownership")
	}
	variable, err := NewExtraZDDVariable(4, 2, ExtraZDDLimits{})
	if err != nil || !slices.EqualFunc(variable.Subsets(), [][]int{{2}}, slices.Equal) {
		t.Fatal(variable, err)
	}
	for _, bad := range []struct {
		variables int
		subsets   [][]int
	}{{-1, nil}, {2, [][]int{{2}}}, {2, [][]int{{-1}}}} {
		if _, err = NewExtraZDD(bad.variables, bad.subsets, ExtraZDDLimits{}); !errors.Is(err, ErrDimension) {
			t.Fatal(bad, err)
		}
	}
	if _, err = NewExtraZDDVariable(2, 2, ExtraZDDLimits{}); !errors.Is(err, ErrDimension) {
		t.Fatal(err)
	}
}

func TestExtraZDDTruncateExampleFlatAndFormat(t *testing.T) {
	example, err := ExtraZDDTruncateExample()
	if err != nil {
		t.Fatal(err)
	}
	want := [][]int{{2, 4}, {0, 3, 5}}
	if !slices.EqualFunc(example.Result.Subsets, want, slices.Equal) {
		t.Fatal(example.Result.Subsets)
	}
	if !slices.Equal(example.Result.Flat, []int{2, 2, 4, -1, 0, 3, 5, -1}) {
		t.Fatal(example.Result.Flat)
	}
	wantText := "The set contains 2 subsets:\n" +
		"Subset   0 : { 2 4 }\n" +
		"Subset   1 : { 0 3 5 }\n"
	if example.Formatted != wantText || FormatExtraZDDSubsets(want) != wantText {
		t.Fatalf("%q", example.Formatted)
	}
	zdd, _ := NewExtraZDD(3, [][]int{{}, {0, 1}, {2}}, ExtraZDDLimits{})
	result, err := zdd.Truncate([]float64{0.5, 0.5, 0.5}, 0.25, ExtraZDDLimits{})
	if err != nil || !slices.EqualFunc(result.Subsets, [][]int{{}, {2}, {0, 1}}, slices.Equal) {
		t.Fatal(result, err)
	}
}

func TestExtraZDDLimitsCancellationAndDimensions(t *testing.T) {
	if _, err := NewExtraZDD(3, [][]int{{0}}, ExtraZDDLimits{MaxVariables: 2}); !errors.Is(err, ErrLimit) {
		t.Fatal(err)
	}
	if _, err := NewExtraZDD(3, [][]int{{0}, {1}}, ExtraZDDLimits{MaxSubsets: 1}); !errors.Is(err, ErrLimit) {
		t.Fatal(err)
	}
	if _, err := NewExtraZDD(3, [][]int{{0, 1}}, ExtraZDDLimits{MaxElements: 1}); !errors.Is(err, ErrLimit) {
		t.Fatal(err)
	}
	if _, err := NewExtraZDD(3, [][]int{{0}}, ExtraZDDLimits{Stop: func() bool { return true }}); !errors.Is(err, ErrCancelled) {
		t.Fatal(err)
	}
	if _, err := NewExtraZDD(1, nil, ExtraZDDLimits{MaxSubsets: -1}); !errors.Is(err, ErrDimension) {
		t.Fatal(err)
	}
	zdd, _ := NewExtraZDD(3, [][]int{{}, {0}, {1, 2}}, ExtraZDDLimits{})
	for _, row := range []struct {
		probabilities []float64
		limit         float64
	}{{[]float64{1, 1}, 0.5}, {[]float64{1, 1, 1}, 0}, {[]float64{1, 1, 1}, 1.1}, {[]float64{1, math.NaN(), 1}, 0.5}, {[]float64{1, 0, 1}, 0.5}} {
		if _, err := zdd.Truncate(row.probabilities, row.limit, ExtraZDDLimits{}); !errors.Is(err, ErrDimension) {
			t.Fatal(row, err)
		}
	}
	if _, err := zdd.Truncate([]float64{1, 1, 1}, 1, ExtraZDDLimits{MaxOutputElements: 3}); !errors.Is(err, ErrLimit) {
		t.Fatal(err)
	}
	if _, err := zdd.Truncate([]float64{1, 1, 1}, 1, ExtraZDDLimits{MaxElements: 1}); !errors.Is(err, ErrLimit) {
		t.Fatal(err)
	}
	if _, err := zdd.Truncate([]float64{1, 1, 1}, 1, ExtraZDDLimits{Stop: func() bool { return true }}); !errors.Is(err, ErrCancelled) {
		t.Fatal(err)
	}
	var nilZDD *ExtraZDD
	if _, err := nilZDD.Truncate(nil, 1, ExtraZDDLimits{}); !errors.Is(err, ErrDimension) {
		t.Fatal(err)
	}
}

func extraReverseBits(value, width int) int {
	result := 0
	for bit := 0; bit < width; bit++ {
		result = result<<1 | value>>uint(bit)&1
	}
	return result
}

func extraSubsetFromMask(mask, variables int) []int {
	result := []int{}
	for variable := 0; variable < variables; variable++ {
		if mask&(1<<uint(variable)) != 0 {
			result = append(result, variable)
		}
	}
	return result
}

func FuzzExtraZDDTruncate(f *testing.F) {
	f.Add(uint8(5), []byte{1, 3, 7, 16, 31})
	f.Add(uint8(0), []byte{0, 0})
	f.Fuzz(func(t *testing.T, flags uint8, data []byte) {
		if len(data) == 0 {
			return
		}
		variables := int(flags % 9)
		universe := 1 << uint(variables)
		count := 1 + len(data)%32
		input := make([][]int, count)
		present := make([]bool, universe)
		for i := range input {
			mask := int(data[i%len(data)]) % universe
			present[mask] = true
			input[i] = extraSubsetFromMask(mask, variables)
			if i&1 != 0 {
				slices.Reverse(input[i])
				if len(input[i]) > 0 {
					input[i] = append(input[i], input[i][0])
				}
			}
		}
		zdd, err := NewExtraZDD(variables, input, ExtraZDDLimits{})
		if err != nil {
			t.Fatal(err)
		}
		choices := [...]float64{0.25, 0.5, 0.75, 1}
		probabilities := make([]float64, variables)
		for variable := range probabilities {
			probabilities[variable] = choices[(int(flags)+variable)%len(choices)]
		}
		limits := [...]float64{0.0625, 0.125, 0.25, 0.5, 1}
		limit := limits[int(flags>>4)%len(limits)]
		wantAll, wantKept := [][]int{}, [][]int{}
		for order := 0; order < universe; order++ {
			mask := extraReverseBits(order, variables)
			if !present[mask] {
				continue
			}
			subset := extraSubsetFromMask(mask, variables)
			wantAll = append(wantAll, subset)
			probability := 1.0
			for _, variable := range subset {
				probability *= probabilities[variable]
			}
			if probability >= limit {
				wantKept = append(wantKept, subset)
			}
		}
		if !slices.EqualFunc(zdd.Subsets(), wantAll, slices.Equal) {
			t.Fatal("order", variables, zdd.Subsets(), wantAll)
		}
		result, err := zdd.Truncate(probabilities, limit, ExtraZDDLimits{})
		if err != nil || !slices.EqualFunc(result.Subsets, wantKept, slices.Equal) {
			t.Fatal("truncate", variables, result.Subsets, wantKept, err)
		}
		flat, err := flattenExtraZDDSubsets(wantKept, 1<<22)
		if err != nil || !slices.Equal(result.Flat, flat) {
			t.Fatal("flat", result.Flat, flat, err)
		}
	})
}
