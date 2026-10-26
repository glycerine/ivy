package abc

// Go-only reconstruction, brute-force, and fuzz tests for src/misc/espresso/gimpel.c.

import (
	"errors"
	"math/bits"
	"slices"
	"testing"
)

func mvcCloseGimpel(t testing.TB, result MvcGimpelReduction) {
	t.Helper()
	if result.Solution != nil {
		if err := result.Solution.Close(); err != nil {
			t.Fatal(err)
		}
	}
	if result.Reduced != nil {
		if err := result.Reduced.Close(); err != nil {
			t.Fatal(err)
		}
	}
}

func mvcSolutionMask(t testing.TB, solution *MvcCoverSolution) uint64 {
	t.Helper()
	columns, err := solution.Columns()
	if err != nil {
		t.Fatal(err)
	}
	var result uint64
	for _, column := range columns {
		result |= uint64(1) << uint(column)
	}
	return result
}

func TestMvcGimpelReductionBothReconstructionChoices(t *testing.T) {
	for _, test := range []struct {
		rows            []uint64
		wantSolution    uint64
		wantReduced     []uint64
		wantReconstruct int
	}{
		{[]uint64{0b000011, 0b000101, 0b001010, 0b010010, 0b100000}, 0b100110, []uint64{0b001100, 0b010100, 0b100000}, 1},
		{[]uint64{0b00011, 0b00101, 0b01010, 0b10010, 0b01000, 0b10000}, 0b11001, []uint64{0b01100, 0b10100, 0b01000, 0b10000}, 0},
	} {
		columns := 6
		if test.wantSolution < 1<<5 {
			columns = 5
		}
		manager, cover := mvcMinimumCoverFixture(t, columns, test.rows)
		before := mvcDominanceMasks(t, cover)
		result, err := cover.GimpelReduce(nil, MvcGimpelOptions{})
		if err != nil || !result.Applied || !result.WithinBound || result.PrimaryRow != 0 || result.SecondaryRow != 1 || result.FirstColumn != 0 || result.SecondColumn != 1 {
			t.Fatal(result, err)
		}
		if got := mvcSolutionMask(t, result.Solution); got != test.wantSolution || !slices.Equal(mvcDominanceMasks(t, result.Reduced), test.wantReduced) {
			t.Fatalf("solution=%06b reduced=%06b", got, mvcDominanceMasks(t, result.Reduced))
		}
		if test.wantSolution>>uint(test.wantReconstruct)&1 == 0 {
			t.Fatal("test does not select expected reconstruction")
		}
		if !slices.Equal(before, mvcDominanceMasks(t, cover)) {
			t.Fatal("input cover mutated")
		}
		mvcCloseGimpel(t, result)
		mvcCloseMinimumFixture(t, manager, cover, nil)
	}
}

func TestMvcGimpelNoPatternBoundAndBaseOwnership(t *testing.T) {
	for _, rows := range [][]uint64{
		{0b0001, 0b0010, 0b1100},
		{0b110111, 0b110001, 0b100010},
		{0b000011, 0b000001, 0b000010},
	} {
		manager, cover := mvcMinimumCoverFixture(t, 6, rows)
		result, err := cover.GimpelReduce(nil, MvcGimpelOptions{})
		if err != nil || result.Applied || result.Reduced != nil || result.Solution != nil {
			t.Fatal(rows, result, err)
		}
		mvcCloseMinimumFixture(t, manager, cover, nil)
	}

	rows := []uint64{0b000011, 0b000101, 0b001010, 0b010010, 0b100000}
	manager, cover := mvcMinimumCoverFixture(t, 7, rows)
	var result MvcGimpelReduction
	var err error
	base, _ := NewMvcCoverSolution(7)
	base.Add(6, nil)
	result, err = cover.GimpelReduce(base, MvcGimpelOptions{Bound: 4})
	if err != nil || !result.Applied || result.WithinBound || result.Solution != nil || result.Reduced == nil {
		t.Fatal(result, err)
	}
	baseColumns, _ := base.Columns()
	if !slices.Equal(baseColumns, []int{6}) {
		t.Fatal("base mutated", baseColumns)
	}
	mvcCloseGimpel(t, result)
	base.Close()
	mvcCloseMinimumFixture(t, manager, cover, nil)
}

func TestMvcGimpelErrorsLimitsAndCancellation(t *testing.T) {
	if result, err := (*MvcCover)(nil).GimpelReduce(nil, MvcGimpelOptions{}); result.Reduced != nil || err == nil {
		t.Fatal(result, err)
	}
	manager, empty := mvcMinimumCoverFixture(t, 3, []uint64{0})
	if result, err := empty.GimpelReduce(nil, MvcGimpelOptions{}); result.Reduced != nil || !errors.Is(err, ErrInvariant) {
		t.Fatal(result, err)
	}
	mvcCloseMinimumFixture(t, manager, empty, nil)

	rows := []uint64{0b000011, 0b000101, 0b001010, 0b010010, 0b100000}
	manager, cover := mvcMinimumCoverFixture(t, 6, rows)
	for _, options := range []MvcGimpelOptions{
		{MaxRows: -1}, {MaxColumns: -1}, {MaxEntries: -1}, {Bound: -1},
		{MaxRows: 4}, {MaxColumns: 5}, {MaxEntries: 2},
		{Stop: func() bool { return true }},
		{Minimum: MvcMinimumCoverOptions{MaxSearchNodes: -1}},
	} {
		result, err := cover.GimpelReduce(nil, options)
		if result.Reduced != nil || result.Solution != nil || !(errors.Is(err, ErrDimension) || errors.Is(err, ErrLimit) || errors.Is(err, ErrCancelled)) {
			t.Fatal(options, result, err)
		}
	}
	wrongBase, _ := NewMvcCoverSolution(5)
	if result, err := cover.GimpelReduce(wrongBase, MvcGimpelOptions{}); result.Reduced != nil || !errors.Is(err, ErrDimension) {
		t.Fatal(result, err)
	}
	wrongBase.Close()
	mvcCloseMinimumFixture(t, manager, cover, nil)
}

func FuzzMvcGimpelReduction(f *testing.F) {
	f.Add(uint8(6), []byte{3, 5, 10, 18, 32})
	f.Fuzz(func(t *testing.T, widthByte uint8, data []byte) {
		columns := 2 + int(widthByte%7)
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
		manager, cover := mvcMinimumCoverFixture(t, columns, rows)
		result, err := cover.GimpelReduce(nil, MvcGimpelOptions{MaxRows: 100, MaxColumns: 100, MaxEntries: 1000, Minimum: MvcMinimumCoverOptions{MaxSearchNodes: 100000}})
		if err != nil {
			t.Fatal(err)
		}
		if result.Applied {
			if !result.WithinBound || result.Solution == nil || result.Reduced == nil {
				t.Fatal(result)
			}
			mask := mvcSolutionMask(t, result.Solution)
			for _, row := range rows {
				if mask&row == 0 {
					t.Fatalf("solution %08b misses row %08b in %08b", mask, row, rows)
				}
			}
			if bits.OnesCount64(mask) != mvcBruteMinimumCount(columns, rows) {
				t.Fatalf("nonminimum solution %08b for %08b", mask, rows)
			}
		}
		mvcCloseGimpel(t, result)
		mvcCloseMinimumFixture(t, manager, cover, nil)
	})
}
