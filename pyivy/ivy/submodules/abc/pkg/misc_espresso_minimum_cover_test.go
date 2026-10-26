package abc

// Go-only exhaustive and fuzz tests for src/misc/espresso/sminterf.c.

import (
	"errors"
	"math/bits"
	"slices"
	"testing"
)

func mvcMinimumCoverFixture(t testing.TB, columns int, rows []uint64) (*MvcManager, *MvcCover) {
	t.Helper()
	manager := NewMvcManager()
	cover, err := NewMvcCover(manager, columns)
	if err != nil {
		t.Fatal(err)
	}
	for _, row := range rows {
		cube, cubeErr := cover.NewCube()
		if cubeErr != nil {
			t.Fatal(cubeErr)
		}
		for column := 0; column < columns; column++ {
			if row>>uint(column)&1 != 0 {
				if cubeErr = cube.SetBit(column, true); cubeErr != nil {
					t.Fatal(cubeErr)
				}
			}
		}
		if cubeErr = cover.AddTail(cube); cubeErr != nil {
			t.Fatal(cubeErr)
		}
	}
	return manager, cover
}

func mvcCloseMinimumFixture(t testing.TB, manager *MvcManager, cover *MvcCover, result *MvcCube) {
	t.Helper()
	if result != nil {
		if err := result.Close(); err != nil {
			t.Fatal(err)
		}
	}
	if err := cover.Close(); err != nil {
		t.Fatal(err)
	}
	if err := manager.Close(); err != nil {
		t.Fatal(err)
	}
}

func mvcMinimumMask(t testing.TB, cube *MvcCube) uint64 {
	t.Helper()
	if cube.cover.bits > 64 {
		t.Fatal("test mask exceeds 64 bits")
	}
	words, err := cube.Words()
	if err != nil {
		t.Fatal(err)
	}
	var result uint64
	for index, word := range words {
		result |= uint64(word) << uint(32*index)
	}
	return result
}

func mvcBruteMinimumCount(columns int, rows []uint64) int {
	best := columns + 1
	for mask := uint64(0); mask < uint64(1)<<uint(columns); mask++ {
		count := bits.OnesCount64(mask)
		if count >= best {
			continue
		}
		valid := true
		for _, row := range rows {
			if mask&row == 0 {
				valid = false
				break
			}
		}
		if valid {
			best = count
		}
	}
	return best
}

func TestMvcMinimumColumnCoverExactDeterministicAndOwned(t *testing.T) {
	rows := []uint64{0b0011, 0b0110, 0b1100}
	manager, cover := mvcMinimumCoverFixture(t, 4, rows)
	before := make([][]uint32, len(cover.cubes))
	for index, cube := range cover.cubes {
		before[index], _ = cube.Words()
	}
	first, err := cover.MinimumColumnCover(MvcMinimumCoverOptions{})
	if err != nil {
		t.Fatal(err)
	}
	second, err := cover.MinimumColumnCover(MvcMinimumCoverOptions{})
	if err != nil {
		t.Fatal(err)
	}
	if got := mvcMinimumMask(t, first); got != 0b0110 || mvcMinimumMask(t, second) != got {
		t.Fatalf("got %04b and %04b", got, mvcMinimumMask(t, second))
	}
	if err = first.SetBit(0, true); err != nil {
		t.Fatal(err)
	}
	if mvcMinimumMask(t, second) != 0b0110 {
		t.Fatal("results alias")
	}
	for index, cube := range cover.cubes {
		after, wordErr := cube.Words()
		if wordErr != nil || !slices.Equal(before[index], after) {
			t.Fatal("input cover mutated", index, wordErr)
		}
	}
	if err = first.Close(); err != nil {
		t.Fatal(err)
	}
	mvcCloseMinimumFixture(t, manager, cover, second)
}

func TestMvcMinimumColumnCoverDominanceEmptyAndWide(t *testing.T) {
	manager, cover := mvcMinimumCoverFixture(t, 3, []uint64{0b001, 0b011, 0b001, 0b111})
	result, err := cover.MinimumColumnCover(MvcMinimumCoverOptions{})
	if err != nil || mvcMinimumMask(t, result) != 1 {
		t.Fatal(mvcMinimumMask(t, result), err)
	}
	mvcCloseMinimumFixture(t, manager, cover, result)

	manager, cover = mvcMinimumCoverFixture(t, 0, nil)
	result, err = cover.MinimumColumnCover(MvcMinimumCoverOptions{})
	if err != nil {
		t.Fatal(err)
	}
	count, err := result.OneCount()
	if err != nil || count != 0 {
		t.Fatal(count, err)
	}
	mvcCloseMinimumFixture(t, manager, cover, result)

	manager = NewMvcManager()
	cover, err = NewMvcCover(manager, 128)
	if err != nil {
		t.Fatal(err)
	}
	for column := 0; column < 128; column++ {
		cube, _ := cover.NewCube()
		cube.SetBit(column, true)
		cover.AddTail(cube)
	}
	result, err = cover.MinimumColumnCover(MvcMinimumCoverOptions{})
	if err != nil {
		t.Fatal(err)
	}
	count, _ = result.OneCount()
	if count != 128 {
		t.Fatal(count)
	}
	mvcCloseMinimumFixture(t, manager, cover, result)
}

func TestMvcMinimumColumnCoverExhaustiveSmallMatrices(t *testing.T) {
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
			result, err := cover.MinimumColumnCover(MvcMinimumCoverOptions{})
			if err != nil {
				t.Fatal(columns, family, err)
			}
			mask := mvcMinimumMask(t, result)
			if bits.OnesCount64(mask) != mvcBruteMinimumCount(columns, rows) {
				t.Fatalf("columns=%d family=%b rows=%b got=%b", columns, family, rows, mask)
			}
			for _, row := range rows {
				if mask&row == 0 {
					t.Fatalf("uncovered row: columns=%d family=%b row=%b mask=%b", columns, family, row, mask)
				}
			}
			mvcCloseMinimumFixture(t, manager, cover, result)
		}
	}
}

func TestMvcMinimumColumnCoverErrorsLimitsAndCancellation(t *testing.T) {
	if result, err := (*MvcCover)(nil).MinimumColumnCover(MvcMinimumCoverOptions{}); result != nil || err == nil {
		t.Fatal(result, err)
	}
	manager, emptyRow := mvcMinimumCoverFixture(t, 3, []uint64{0})
	if result, err := emptyRow.MinimumColumnCover(MvcMinimumCoverOptions{}); result != nil || !errors.Is(err, ErrInvariant) {
		t.Fatal(result, err)
	}
	mvcCloseMinimumFixture(t, manager, emptyRow, nil)

	rows := []uint64{0b00011, 0b00011, 0b00101, 0b00101, 0b01010, 0b10100}
	manager, cover := mvcMinimumCoverFixture(t, 5, rows)
	for _, options := range []MvcMinimumCoverOptions{
		{MaxRows: -1},
		{MaxColumns: -1},
		{MaxPairChecks: -1},
		{MaxSearchNodes: -1},
		{MaxRows: 5},
		{MaxColumns: 4},
		{MaxPairChecks: 1},
		{MaxSearchNodes: 1},
		{Stop: func() bool { return true }},
	} {
		result, err := cover.MinimumColumnCover(options)
		if result != nil || !(errors.Is(err, ErrDimension) || errors.Is(err, ErrLimit) || errors.Is(err, ErrCancelled)) {
			t.Fatal(options, result, err)
		}
	}
	mvcCloseMinimumFixture(t, manager, cover, nil)
}

func FuzzMvcMinimumColumnCover(f *testing.F) {
	f.Add(uint8(5), []byte{3, 5, 9, 17, 6, 12})
	f.Fuzz(func(t *testing.T, widthByte uint8, data []byte) {
		columns := 1 + int(widthByte%8)
		if len(data) > 12 {
			data = data[:12]
		}
		mask := byte((1 << uint(columns)) - 1)
		rows := make([]uint64, len(data))
		for index, value := range data {
			value &= mask
			if value == 0 {
				value = 1 << uint(index%columns)
			}
			rows[index] = uint64(value)
		}
		manager, cover := mvcMinimumCoverFixture(t, columns, rows)
		result, err := cover.MinimumColumnCover(MvcMinimumCoverOptions{MaxSearchNodes: 100000})
		if err != nil {
			t.Fatal(err)
		}
		got := mvcMinimumMask(t, result)
		if bits.OnesCount64(got) != mvcBruteMinimumCount(columns, rows) {
			t.Fatalf("columns=%d rows=%b got=%b", columns, rows, got)
		}
		for _, row := range rows {
			if got&row == 0 {
				t.Fatalf("row %b not covered by %b", row, got)
			}
		}
		mvcCloseMinimumFixture(t, manager, cover, result)
	})
}
