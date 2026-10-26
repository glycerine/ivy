package abc

// Go-only exhaustive, model, and fuzz tests for src/misc/espresso/mincov.c.

import (
	"errors"
	"math"
	"slices"
	"testing"
)

func mvcSearchResultMask(result MvcMinimumCoverSearchResult) uint64 {
	var mask uint64
	for _, column := range result.Columns {
		mask |= uint64(1) << uint(column)
	}
	return mask
}

func mvcSearchResultCost(result MvcMinimumCoverSearchResult, weights []int) int {
	cost := 0
	for _, column := range result.Columns {
		if weights == nil {
			cost++
		} else {
			cost += weights[column]
		}
	}
	return cost
}

func mvcSearchResultCovers(rows []uint64, result MvcMinimumCoverSearchResult) bool {
	mask := mvcSearchResultMask(result)
	for _, row := range rows {
		if row&mask == 0 {
			return false
		}
	}
	return true
}

func TestMvcMinimumWeightedColumnCoverExactHeuristicAndOwned(t *testing.T) {
	rows := []uint64{0b00011, 0b00101, 0b01010, 0b10100}
	manager, cover := mvcMinimumCoverFixture(t, 5, rows)
	before := mvcDominanceMasks(t, cover)
	heuristic, err := cover.MinimumWeightedColumnCover(nil, MvcMinimumCoverSearchOptions{Heuristic: true})
	if err != nil || heuristic.Cost != 3 || !slices.Equal(heuristic.Columns, []int{0, 1, 2}) || !heuristic.Stats.Heuristic || !mvcSearchResultCovers(rows, heuristic) {
		t.Fatal(heuristic, err)
	}
	exact, err := cover.MinimumWeightedColumnCover(nil, MvcMinimumCoverSearchOptions{})
	if err != nil || exact.Cost != 2 || !slices.Equal(exact.Columns, []int{1, 2}) || exact.Stats.Heuristic || exact.Stats.LowerBound != 2 || exact.Stats.Nodes < 2 {
		t.Fatal(exact, err)
	}
	weighted, err := cover.MinimumWeightedColumnCover([]int{9, 4, 3, 8, 7}, MvcMinimumCoverSearchOptions{})
	wantCost := mvcBruteMinimumCost(5, rows, []int{9, 4, 3, 8, 7}, 0b11111)
	if err != nil || weighted.Cost != wantCost || weighted.Cost != mvcSearchResultCost(weighted, []int{9, 4, 3, 8, 7}) || !mvcSearchResultCovers(rows, weighted) {
		t.Fatal(weighted, wantCost, err)
	}
	exact.Columns[0] = 99
	again, err := cover.MinimumWeightedColumnCover(nil, MvcMinimumCoverSearchOptions{})
	if err != nil || !slices.Equal(again.Columns, []int{1, 2}) || !slices.Equal(before, mvcDominanceMasks(t, cover)) {
		t.Fatal(again, err)
	}
	mvcCloseMinimumFixture(t, manager, cover, nil)
}

func TestMvcMinimumWeightedColumnCoverExhaustiveSmall(t *testing.T) {
	for columns := 1; columns <= 3; columns++ {
		rowKinds := (1 << uint(columns)) - 1
		weightCases := 1
		for range columns {
			weightCases *= 3
		}
		for family := 0; family < 1<<uint(rowKinds); family++ {
			rows := make([]uint64, 0, rowKinds)
			for row := 1; row <= rowKinds; row++ {
				if family>>uint(row-1)&1 != 0 {
					rows = append(rows, uint64(row))
				}
			}
			manager, cover := mvcMinimumCoverFixture(t, columns, rows)
			for encoded := 0; encoded < weightCases; encoded++ {
				weights, value := make([]int, columns), encoded
				for column := range weights {
					weights[column], value = value%3, value/3
				}
				got, err := cover.MinimumWeightedColumnCover(weights, MvcMinimumCoverSearchOptions{})
				want := mvcBruteMinimumCost(columns, rows, weights, uint64(1)<<uint(columns)-1)
				if err != nil || got.Cost != want || got.Cost != mvcSearchResultCost(got, weights) || !mvcSearchResultCovers(rows, got) || got.Stats.LowerBound > got.Cost {
					t.Fatalf("columns=%d family=%b weights=%v got=%+v want=%d err=%v", columns, family, weights, got, want, err)
				}
			}
			mvcCloseMinimumFixture(t, manager, cover, nil)
		}
	}
}

func TestMvcMinimumWeightedColumnCoverErrorsLimitsAndOverflow(t *testing.T) {
	var nilCover *MvcCover
	if _, err := nilCover.MinimumWeightedColumnCover(nil, MvcMinimumCoverSearchOptions{}); !errors.Is(err, ErrInvariant) {
		t.Fatal(err)
	}
	manager, cover := mvcMinimumCoverFixture(t, 5, []uint64{0b00011, 0b00101, 0b01010, 0b10100})
	checks := []struct {
		weights []int
		options MvcMinimumCoverSearchOptions
		want    error
	}{
		{nil, MvcMinimumCoverSearchOptions{MaxRows: -1}, ErrDimension},
		{[]int{1}, MvcMinimumCoverSearchOptions{}, ErrDimension},
		{[]int{1, 1, -1, 1, 1}, MvcMinimumCoverSearchOptions{}, ErrDimension},
		{nil, MvcMinimumCoverSearchOptions{MaxRows: 1}, ErrLimit},
		{nil, MvcMinimumCoverSearchOptions{MaxColumns: 1}, ErrLimit},
		{nil, MvcMinimumCoverSearchOptions{MaxPairChecks: 1}, ErrLimit},
		{nil, MvcMinimumCoverSearchOptions{MaxSearchNodes: 1}, ErrLimit},
		{nil, MvcMinimumCoverSearchOptions{Stop: func() bool { return true }}, ErrCancelled},
	}
	for index, check := range checks {
		got, err := cover.MinimumWeightedColumnCover(check.weights, check.options)
		if !errors.Is(err, check.want) || got.Columns != nil || got.Cost != 0 || got.Stats != (MvcMinimumCoverSearchStats{}) {
			t.Fatal(index, got, err)
		}
	}
	mvcCloseMinimumFixture(t, manager, cover, nil)

	manager, emptyRow := mvcMinimumCoverFixture(t, 1, []uint64{0})
	if _, err := emptyRow.MinimumWeightedColumnCover(nil, MvcMinimumCoverSearchOptions{}); !errors.Is(err, ErrInvariant) {
		t.Fatal(err)
	}
	mvcCloseMinimumFixture(t, manager, emptyRow, nil)

	manager, overflow := mvcMinimumCoverFixture(t, 2, []uint64{1, 2})
	if _, err := overflow.MinimumWeightedColumnCover([]int{math.MaxInt, 1}, MvcMinimumCoverSearchOptions{}); !errors.Is(err, ErrLimit) {
		t.Fatal(err)
	}
	mvcCloseMinimumFixture(t, manager, overflow, nil)
}

func FuzzMvcMinimumWeightedColumnCover(f *testing.F) {
	f.Add([]byte{5, 4, 3, 5, 10, 20, 9, 4, 3, 8, 7})
	f.Add([]byte{3, 3, 1, 2, 4, 0, 1, 2})
	f.Fuzz(func(t *testing.T, data []byte) {
		if len(data) < 2 {
			return
		}
		columns := int(data[0]%8) + 1
		rowCount := int(data[1] % 13)
		rows := make([]uint64, rowCount)
		for row := range rows {
			value := byte(row + 1)
			if 2+row < len(data) {
				value = data[2+row]
			}
			rows[row] = uint64(value)&(uint64(1)<<uint(columns)-1) | uint64(1)<<uint(row%columns)
		}
		weights := make([]int, columns)
		for column := range weights {
			value := byte(column)
			if 2+rowCount+column < len(data) {
				value = data[2+rowCount+column]
			}
			weights[column] = int(value % 10)
		}
		manager, cover := mvcMinimumCoverFixture(t, columns, rows)
		exact, err := cover.MinimumWeightedColumnCover(weights, MvcMinimumCoverSearchOptions{})
		if err != nil {
			t.Fatal(err)
		}
		want := mvcBruteMinimumCost(columns, rows, weights, uint64(1)<<uint(columns)-1)
		if exact.Cost != want || exact.Cost != mvcSearchResultCost(exact, weights) || !mvcSearchResultCovers(rows, exact) || exact.Stats.LowerBound > exact.Cost {
			t.Fatal(rows, weights, exact, want)
		}
		heuristic, err := cover.MinimumWeightedColumnCover(weights, MvcMinimumCoverSearchOptions{Heuristic: true})
		if err != nil || !mvcSearchResultCovers(rows, heuristic) || heuristic.Cost != mvcSearchResultCost(heuristic, weights) || heuristic.Cost < exact.Cost {
			t.Fatal(rows, weights, exact, heuristic, err)
		}
		mvcCloseMinimumFixture(t, manager, cover, nil)
	})
}
