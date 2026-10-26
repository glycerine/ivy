package abc

// Go-only model and fuzz tests for src/misc/espresso/dominate.c.

import (
	"errors"
	"math"
	"math/bits"
	"slices"
	"testing"
)

func mvcDominanceMasks(t testing.TB, cover *MvcCover) []uint64 {
	t.Helper()
	result := make([]uint64, len(cover.cubes))
	for index, cube := range cover.cubes {
		result[index] = mvcMinimumMask(t, cube)
	}
	return result
}

func mvcActiveMask(active []bool) uint64 {
	var result uint64
	for column, yes := range active {
		if yes {
			result |= uint64(1) << uint(column)
		}
	}
	return result
}

func mvcBruteMinimumCost(columns int, rows []uint64, weights []int, allowed uint64) int {
	best := math.MaxInt
	for selected := uint64(0); selected < uint64(1)<<uint(columns); selected++ {
		if selected&^allowed != 0 {
			continue
		}
		valid := true
		for _, row := range rows {
			if selected&row == 0 {
				valid = false
				break
			}
		}
		if !valid {
			continue
		}
		cost := 0
		for column := 0; column < columns; column++ {
			if selected>>uint(column)&1 != 0 {
				if weights == nil {
					cost++
				} else {
					cost += weights[column]
				}
			}
		}
		best = min(best, cost)
	}
	return best
}

func TestMvcRowDominanceExactOrderAndOwnership(t *testing.T) {
	manager, cover := mvcMinimumCoverFixture(t, 4, []uint64{0b0001, 0b0011, 0b0001, 0b1100})
	before := mvcDominanceMasks(t, cover)
	result, err := cover.ReduceDominatedRows(MvcDominanceOptions{})
	if err != nil {
		t.Fatal(err)
	}
	if result.RemovedRows != 2 || result.RemovedColumns != 0 || !slices.Equal(mvcDominanceMasks(t, result.Cover), []uint64{0b0001, 0b1100}) || mvcActiveMask(result.ActiveColumns) != 0b1101 {
		t.Fatal(result.RemovedRows, result.RemovedColumns, mvcDominanceMasks(t, result.Cover), result.ActiveColumns)
	}
	if !slices.Equal(before, mvcDominanceMasks(t, cover)) {
		t.Fatal("input cover mutated")
	}
	result.Cover.cubes[0].SetBit(1, true)
	if !slices.Equal(before, mvcDominanceMasks(t, cover)) {
		t.Fatal("result aliases input")
	}
	if err = result.Cover.Close(); err != nil {
		t.Fatal(err)
	}
	mvcCloseMinimumFixture(t, manager, cover, nil)
}

func TestMvcColumnDominanceUnitWeightedAndTieOrder(t *testing.T) {
	rows := []uint64{0b011, 0b111, 0b110}
	manager, cover := mvcMinimumCoverFixture(t, 3, rows)
	unit, err := cover.ReduceDominatedColumns(nil, MvcDominanceOptions{})
	if err != nil || unit.RemovedColumns != 2 || mvcActiveMask(unit.ActiveColumns) != 0b010 || !slices.Equal(mvcDominanceMasks(t, unit.Cover), []uint64{0b010, 0b010, 0b010}) {
		t.Fatal(unit, err)
	}
	if err = unit.Cover.Close(); err != nil {
		t.Fatal(err)
	}
	weighted, err := cover.ReduceDominatedColumns([]int{1, 5, 1}, MvcDominanceOptions{})
	if err != nil || weighted.RemovedColumns != 0 || mvcActiveMask(weighted.ActiveColumns) != 0b111 {
		t.Fatal(weighted, err)
	}
	if err = weighted.Cover.Close(); err != nil {
		t.Fatal(err)
	}
	mvcCloseMinimumFixture(t, manager, cover, nil)

	manager, cover = mvcMinimumCoverFixture(t, 2, []uint64{0b11, 0b11})
	tied, err := cover.ReduceDominatedColumns(nil, MvcDominanceOptions{})
	if err != nil || mvcActiveMask(tied.ActiveColumns) != 0b10 {
		t.Fatal(tied, err)
	}
	if err = tied.Cover.Close(); err != nil {
		t.Fatal(err)
	}
	mvcCloseMinimumFixture(t, manager, cover, nil)
}

func TestMvcDominancePreservesExhaustiveCoverOptima(t *testing.T) {
	for columns := 1; columns <= 3; columns++ {
		rowKinds := (1 << uint(columns)) - 1
		for family := 0; family < 1<<uint(rowKinds); family++ {
			rows := make([]uint64, 0, rowKinds)
			for row := 1; row <= rowKinds; row++ {
				if family>>uint(row-1)&1 != 0 {
					rows = append(rows, uint64(row))
				}
			}
			manager, cover := mvcMinimumCoverFixture(t, columns, rows)
			rowResult, err := cover.ReduceDominatedRows(MvcDominanceOptions{})
			if err != nil {
				t.Fatal(columns, family, err)
			}
			all := uint64(1)<<uint(columns) - 1
			if mvcBruteMinimumCost(columns, rows, nil, all) != mvcBruteMinimumCost(columns, mvcDominanceMasks(t, rowResult.Cover), nil, mvcActiveMask(rowResult.ActiveColumns)) {
				t.Fatal("row dominance changed optimum", columns, family)
			}
			rowResult.Cover.Close()
			weights := make([]int, columns)
			for index := range weights {
				weights[index] = 1 + (index+family)%3
			}
			columnResult, columnErr := cover.ReduceDominatedColumns(weights, MvcDominanceOptions{})
			if columnErr != nil {
				t.Fatal(columns, family, columnErr)
			}
			if mvcBruteMinimumCost(columns, rows, weights, all) != mvcBruteMinimumCost(columns, mvcDominanceMasks(t, columnResult.Cover), weights, mvcActiveMask(columnResult.ActiveColumns)) {
				t.Fatal("column dominance changed optimum", columns, family, weights, columnResult.ActiveColumns)
			}
			columnResult.Cover.Close()
			mvcCloseMinimumFixture(t, manager, cover, nil)
		}
	}
}

func TestMvcDominanceErrorsLimitsAndCancellation(t *testing.T) {
	if result, err := (*MvcCover)(nil).ReduceDominatedRows(MvcDominanceOptions{}); result.Cover != nil || err == nil {
		t.Fatal(result, err)
	}
	manager, empty := mvcMinimumCoverFixture(t, 2, []uint64{0})
	if result, err := empty.ReduceDominatedRows(MvcDominanceOptions{}); result.Cover != nil || !errors.Is(err, ErrInvariant) {
		t.Fatal(result, err)
	}
	mvcCloseMinimumFixture(t, manager, empty, nil)

	manager, cover := mvcMinimumCoverFixture(t, 3, []uint64{0b011, 0b110})
	for _, options := range []MvcDominanceOptions{{MaxRows: -1}, {MaxColumns: -1}, {MaxChecks: -1}, {MaxRows: 1}, {MaxColumns: 2}, {MaxChecks: 1}, {Stop: func() bool { return true }}} {
		result, err := cover.ReduceDominatedRows(options)
		if result.Cover != nil || !(errors.Is(err, ErrDimension) || errors.Is(err, ErrLimit) || errors.Is(err, ErrCancelled)) {
			t.Fatal(options, result, err)
		}
	}
	if result, err := cover.ReduceDominatedColumns([]int{1}, MvcDominanceOptions{}); result.Cover != nil || !errors.Is(err, ErrDimension) {
		t.Fatal(result, err)
	}
	mvcCloseMinimumFixture(t, manager, cover, nil)
}

func FuzzMvcDominance(f *testing.F) {
	f.Add(uint8(5), []byte{3, 7, 12, 17, 24, 9}, []byte{1, 3, 2, 4, 2})
	f.Fuzz(func(t *testing.T, widthByte uint8, data, weightData []byte) {
		columns := 1 + int(widthByte%8)
		if len(data) > 12 {
			data = data[:12]
		}
		limit := byte((1 << uint(columns)) - 1)
		rows := make([]uint64, len(data))
		for index, value := range data {
			value &= limit
			if value == 0 {
				value = 1 << uint(index%columns)
			}
			rows[index] = uint64(value)
		}
		weights := make([]int, columns)
		for column := range weights {
			if len(weightData) == 0 {
				weights[column] = 1
			} else {
				weights[column] = 1 + int(weightData[column%len(weightData)]%7)
			}
		}
		manager, cover := mvcMinimumCoverFixture(t, columns, rows)
		all := uint64(1)<<uint(columns) - 1
		wantUnit := mvcBruteMinimumCost(columns, rows, nil, all)
		wantWeighted := mvcBruteMinimumCost(columns, rows, weights, all)
		rowResult, err := cover.ReduceDominatedRows(MvcDominanceOptions{MaxChecks: 10000})
		if err != nil {
			t.Fatal(err)
		}
		if got := mvcBruteMinimumCost(columns, mvcDominanceMasks(t, rowResult.Cover), nil, mvcActiveMask(rowResult.ActiveColumns)); got != wantUnit {
			t.Fatalf("row reduction changed optimum: got %d want %d rows=%b", got, wantUnit, rows)
		}
		rowResult.Cover.Close()
		columnResult, err := cover.ReduceDominatedColumns(weights, MvcDominanceOptions{MaxChecks: 10000})
		if err != nil {
			t.Fatal(err)
		}
		if got := mvcBruteMinimumCost(columns, mvcDominanceMasks(t, columnResult.Cover), weights, mvcActiveMask(columnResult.ActiveColumns)); got != wantWeighted {
			t.Fatalf("column reduction changed optimum: got %d want %d rows=%b weights=%v active=%08b", got, wantWeighted, rows, weights, mvcActiveMask(columnResult.ActiveColumns))
		}
		columnResult.Cover.Close()
		mvcCloseMinimumFixture(t, manager, cover, nil)
	})
}

func TestMvcDominanceIndependentPopulationSanity(t *testing.T) {
	if bits.OnesCount64(mvcActiveMask([]bool{true, false, true})) != 2 {
		t.Fatal("active-column test model is broken")
	}
}
