package abc

// Go-only model, exhaustive, and fuzz tests for src/misc/espresso/indep.c.

import (
	"errors"
	"math"
	"math/bits"
	"slices"
	"sort"
	"testing"
)

func mvcIndependentModel(rows []uint64, weights []int) MvcIndependentRows {
	active := make([]bool, len(rows))
	for row := range active {
		active[row] = true
	}
	result := MvcIndependentRows{}
	remaining := len(rows)
	for remaining != 0 {
		best, bestDegree := -1, 0
		for row, yes := range active {
			if !yes {
				continue
			}
			degree := 0
			for other, otherActive := range active {
				if otherActive && rows[row]&rows[other] != 0 {
					degree++
				}
			}
			if best < 0 || degree < bestDegree {
				best, bestDegree = row, degree
			}
		}
		minimum := 1
		if weights != nil {
			minimum = math.MaxInt
			for columns := rows[best]; columns != 0; columns &= columns - 1 {
				column := bits.TrailingZeros64(columns)
				minimum = min(minimum, weights[column])
			}
		}
		result.Cost += minimum
		result.SelectionOrder = append(result.SelectionOrder, best)
		for row, yes := range active {
			if yes && rows[best]&rows[row] != 0 {
				active[row] = false
				remaining--
			}
		}
	}
	result.Rows = slices.Clone(result.SelectionOrder)
	sort.Ints(result.Rows)
	return result
}

func mvcCheckIndependentResult(t testing.TB, rows []uint64, result MvcIndependentRows) {
	t.Helper()
	selected := make([]bool, len(rows))
	for _, row := range result.Rows {
		if row < 0 || row >= len(rows) || selected[row] {
			t.Fatalf("invalid selected rows %v", result.Rows)
		}
		selected[row] = true
	}
	for first, row := range result.Rows {
		for _, other := range result.Rows[first+1:] {
			if rows[row]&rows[other] != 0 {
				t.Fatalf("selected rows %d and %d intersect", row, other)
			}
		}
	}
	for row, yes := range selected {
		if yes {
			continue
		}
		maximal := false
		for _, chosen := range result.Rows {
			if rows[row]&rows[chosen] != 0 {
				maximal = true
				break
			}
		}
		if !maximal {
			t.Fatalf("unselected row %d could extend independent set %v", row, result.Rows)
		}
	}
}

func TestMvcMaximalIndependentRowsOrderWeightsAndOwnership(t *testing.T) {
	rows := []uint64{0b000011, 0b000110, 0b001000, 0b011000, 0b100000}
	manager, cover := mvcMinimumCoverFixture(t, 6, rows)
	before := mvcDominanceMasks(t, cover)
	unit, err := cover.MaximalIndependentRows(nil, MvcIndependentRowsOptions{})
	if err != nil || !slices.Equal(unit.SelectionOrder, []int{4, 0, 2}) || !slices.Equal(unit.Rows, []int{0, 2, 4}) || unit.Cost != 3 {
		t.Fatal(unit, err)
	}
	weighted, err := cover.MaximalIndependentRows([]int{5, 2, 7, 4, 1, 9}, MvcIndependentRowsOptions{})
	if err != nil || !slices.Equal(weighted.SelectionOrder, []int{4, 0, 2}) || !slices.Equal(weighted.Rows, []int{0, 2, 4}) || weighted.Cost != 15 {
		t.Fatal(weighted, err)
	}
	mvcCheckIndependentResult(t, rows, weighted)
	weighted.Rows[0] = 99
	weighted.SelectionOrder[0] = 99
	again, err := cover.MaximalIndependentRows([]int{5, 2, 7, 4, 1, 9}, MvcIndependentRowsOptions{})
	if err != nil || !slices.Equal(again.Rows, []int{0, 2, 4}) || !slices.Equal(again.SelectionOrder, []int{4, 0, 2}) {
		t.Fatal("result storage was not independent", again, err)
	}
	if !slices.Equal(before, mvcDominanceMasks(t, cover)) {
		t.Fatal("input cover mutated")
	}
	mvcCloseMinimumFixture(t, manager, cover, nil)
}

func TestMvcMaximalIndependentRowsExhaustiveSmall(t *testing.T) {
	for columns := 1; columns <= 3; columns++ {
		rowKinds := (1 << uint(columns)) - 1
		weights := []int{3, 1, 4}[:columns]
		for family := 0; family < 1<<uint(rowKinds); family++ {
			rows := make([]uint64, 0, rowKinds)
			for row := 1; row <= rowKinds; row++ {
				if family>>uint(row-1)&1 != 0 {
					rows = append(rows, uint64(row))
				}
			}
			manager, cover := mvcMinimumCoverFixture(t, columns, rows)
			got, err := cover.MaximalIndependentRows(weights, MvcIndependentRowsOptions{})
			if err != nil {
				t.Fatal(columns, family, err)
			}
			want := mvcIndependentModel(rows, weights)
			if !slices.Equal(got.Rows, want.Rows) || !slices.Equal(got.SelectionOrder, want.SelectionOrder) || got.Cost != want.Cost {
				t.Fatalf("columns=%d family=%b got=%+v want=%+v", columns, family, got, want)
			}
			mvcCheckIndependentResult(t, rows, got)
			if len(rows) != 0 && got.Cost > mvcBruteMinimumCost(columns, rows, weights, uint64(1)<<uint(columns)-1) {
				t.Fatalf("independent cost is not a lower bound: columns=%d family=%b result=%+v", columns, family, got)
			}
			mvcCloseMinimumFixture(t, manager, cover, nil)
		}
	}
}

func TestMvcMaximalIndependentRowsLimitsErrorsAndWide(t *testing.T) {
	var nilCover *MvcCover
	if _, err := nilCover.MaximalIndependentRows(nil, MvcIndependentRowsOptions{}); !errors.Is(err, ErrInvariant) {
		t.Fatal(err)
	}
	manager, cover := mvcMinimumCoverFixture(t, 2, []uint64{1, 2})
	checks := []struct {
		weights []int
		options MvcIndependentRowsOptions
		want    error
	}{
		{nil, MvcIndependentRowsOptions{MaxRows: -1}, ErrDimension},
		{[]int{1}, MvcIndependentRowsOptions{}, ErrDimension},
		{[]int{-1, 1}, MvcIndependentRowsOptions{}, ErrDimension},
		{nil, MvcIndependentRowsOptions{MaxRows: 1}, ErrLimit},
		{nil, MvcIndependentRowsOptions{MaxColumns: 1}, ErrLimit},
		{nil, MvcIndependentRowsOptions{MaxPairChecks: 1, MaxVisits: 1}, ErrLimit},
		{nil, MvcIndependentRowsOptions{MaxAdjacency: 1}, ErrLimit},
		{nil, MvcIndependentRowsOptions{Stop: func() bool { return true }}, ErrCancelled},
	}
	for index, check := range checks {
		got, err := cover.MaximalIndependentRows(check.weights, check.options)
		if !errors.Is(err, check.want) || got.Rows != nil || got.SelectionOrder != nil || got.Cost != 0 {
			t.Fatal(index, got, err)
		}
	}
	mvcCloseMinimumFixture(t, manager, cover, nil)

	manager, emptyRow := mvcMinimumCoverFixture(t, 1, []uint64{0})
	if _, err := emptyRow.MaximalIndependentRows(nil, MvcIndependentRowsOptions{}); !errors.Is(err, ErrInvariant) {
		t.Fatal(err)
	}
	mvcCloseMinimumFixture(t, manager, emptyRow, nil)

	manager, overflow := mvcMinimumCoverFixture(t, 2, []uint64{1, 2})
	if _, err := overflow.MaximalIndependentRows([]int{math.MaxInt, 1}, MvcIndependentRowsOptions{}); !errors.Is(err, ErrLimit) {
		t.Fatal(err)
	}
	mvcCloseMinimumFixture(t, manager, overflow, nil)

	wideManager := NewMvcManager()
	wide, err := NewMvcCover(wideManager, 129)
	if err != nil {
		t.Fatal(err)
	}
	for column := 0; column < 129; column++ {
		cube, cubeErr := wide.NewCube()
		if cubeErr == nil {
			cubeErr = cube.SetBit(column, true)
		}
		if cubeErr == nil {
			cubeErr = wide.AddTail(cube)
		}
		if cubeErr != nil {
			t.Fatal(cubeErr)
		}
	}
	wideResult, err := wide.MaximalIndependentRows(nil, MvcIndependentRowsOptions{MaxPairChecks: 10_000, MaxAdjacency: 1_000, MaxVisits: 20_000})
	if err != nil || len(wideResult.Rows) != 129 || wideResult.Cost != 129 || wideResult.Rows[128] != 128 {
		t.Fatal(len(wideResult.Rows), wideResult.Cost, err)
	}
	mvcCloseMinimumFixture(t, wideManager, wide, nil)
}

func FuzzMvcMaximalIndependentRows(f *testing.F) {
	f.Add([]byte{6, 5, 3, 7, 12, 24, 32, 9, 2, 7, 4, 1, 9})
	f.Add([]byte{3, 4, 1, 2, 4, 7, 0, 5, 2})
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
		got, err := cover.MaximalIndependentRows(weights, MvcIndependentRowsOptions{})
		if err != nil {
			t.Fatal(err)
		}
		want := mvcIndependentModel(rows, weights)
		if !slices.Equal(got.Rows, want.Rows) || !slices.Equal(got.SelectionOrder, want.SelectionOrder) || got.Cost != want.Cost {
			t.Fatalf("rows=%v weights=%v got=%+v want=%+v", rows, weights, got, want)
		}
		mvcCheckIndependentResult(t, rows, got)
		if len(rows) != 0 && got.Cost > mvcBruteMinimumCost(columns, rows, weights, uint64(1)<<uint(columns)-1) {
			t.Fatal("lower-bound violation", rows, weights, got)
		}
		mvcCloseMinimumFixture(t, manager, cover, nil)
	})
}
