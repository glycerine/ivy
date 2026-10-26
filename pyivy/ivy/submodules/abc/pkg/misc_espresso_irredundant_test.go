package abc

// Go-only assignment models for src/misc/espresso/irred.c.

import (
	"errors"
	"slices"
	"sort"
	"testing"
)

func espressoIrredModel(layout *EspressoCubeLayout, onMasks, dcMasks []uint64) (essential, total, partial []int, rows [][]int) {
	dcTruth := uint64(0)
	for _, mask := range dcMasks {
		dcTruth |= espressoCubeTruth(layout, mask)
	}
	truths := make([]uint64, len(onMasks))
	for index, mask := range onMasks {
		truths[index] = espressoCubeTruth(layout, mask)
	}
	for index, truth := range truths {
		other := dcTruth
		for current, candidate := range truths {
			if current != index {
				other |= candidate
			}
		}
		if truth&^other == 0 {
			partial = append(partial, index)
		} else {
			essential = append(essential, index)
		}
	}
	fixed := dcTruth
	for _, index := range essential {
		fixed |= truths[index]
	}
	redundant := slices.Clone(partial)
	partial = partial[:0]
	for _, index := range redundant {
		if truths[index]&^fixed == 0 {
			total = append(total, index)
		} else {
			partial = append(partial, index)
		}
	}
	target := uint64(0)
	for _, index := range partial {
		target |= truths[index] &^ fixed
	}
	seen := make(map[string]bool)
	for assignment := 0; assignment < 64; assignment++ {
		if target>>uint(assignment)&1 == 0 {
			continue
		}
		row := make([]int, 0)
		for _, index := range partial {
			if truths[index]>>uint(assignment)&1 != 0 {
				row = append(row, index)
			}
		}
		key := ""
		for _, index := range row {
			key += string(rune(index + 1))
		}
		if !seen[key] {
			seen[key] = true
			rows = append(rows, row)
		}
	}
	return
}

func espressoSortRows(rows [][]int) {
	sort.Slice(rows, func(i, j int) bool {
		for index := 0; index < min(len(rows[i]), len(rows[j])); index++ {
			if rows[i][index] != rows[j][index] {
				return rows[i][index] < rows[j][index]
			}
		}
		return len(rows[i]) < len(rows[j])
	})
}

func TestEspressoIrredundantSplitTableMarkAndOwnership(t *testing.T) {
	layout, _ := NewEspressoCubeLayout([]int{2, 2, 2}, 3)
	onMasks := []uint64{0b01 | 0b11<<2 | 0b11<<4, 0b11 | 0b01<<2 | 0b11<<4, 0b01 | 0b01<<2 | 0b11<<4, 0b10 | 0b10<<2 | 0b01<<4}
	dcMasks := []uint64{0b10 | 0b10<<2 | 0b10<<4}
	on := espressoMapFamily(t, 6, onMasks...)
	dc := espressoMapFamily(t, 6, dcMasks...)
	wantE, wantT, wantP, wantRows := espressoIrredModel(layout, onMasks, dcMasks)
	result, err := layout.Irredundant(on, dc, EspressoIrredundantOptions{})
	gotRows := slices.Clone(result.Table.Rows)
	espressoSortRows(gotRows)
	espressoSortRows(wantRows)
	if err != nil || !slices.Equal(result.Split.EssentialIndices, wantE) || !slices.Equal(result.Split.TotallyRedundantIndices, wantT) || !slices.Equal(result.Split.PartiallyRedundantIndices, wantP) || !slices.EqualFunc(gotRows, wantRows, slices.Equal) {
		t.Fatal(result.Split, gotRows, wantE, wantT, wantP, wantRows, err)
	}
	if espressoFamilyTruth(t, layout, result.Cover)|espressoFamilyTruth(t, layout, dc) != espressoFamilyTruth(t, layout, on)|espressoFamilyTruth(t, layout, dc) {
		t.Fatal("function changed")
	}
	for _, index := range result.Split.EssentialIndices {
		if result.Marked.sets[index].Flags&(EspressoSetActive|EspressoSetRelEssential) != EspressoSetActive|EspressoSetRelEssential {
			t.Fatal("essential flags", index)
		}
	}
	result.Marked.sets[0].words[0] = 0
	if !slices.Equal(espressoFamilyMasks(t, on), onMasks) {
		t.Fatal("source mutated")
	}
}

func TestEspressoIrredundantTautologyCoverageExactAndErrors(t *testing.T) {
	layout, _ := NewEspressoCubeLayout([]int{2, 2}, 2)
	full := espressoMapFamily(t, 4, 0b0111, 0b1011)
	notFull := espressoMapFamily(t, 4, 0b0101)
	tautology, err := layout.CoverIsTautology(full, EspressoIrredundantOptions{})
	other, otherErr := layout.CoverIsTautology(notFull, EspressoIrredundantOptions{})
	cube := espressoSetFromMask(t, 4, 0b0101)
	covered, coveredErr := layout.CubeCoveredByCover(full, cube, EspressoIrredundantOptions{})
	if err != nil || otherErr != nil || coveredErr != nil || !tautology || other || !covered {
		t.Fatal(tautology, other, covered, err, otherErr, coveredErr)
	}
	dc := espressoMapFamily(t, 4)
	exact, err := layout.Irredundant(full, dc, EspressoIrredundantOptions{Exact: true})
	if err != nil || espressoFamilyTruth(t, layout, exact.Cover) != espressoFamilyTruth(t, layout, full) {
		t.Fatal(exact, err)
	}
	for _, options := range []EspressoIrredundantOptions{{MaxCubes: -1}, {MaxCubes: 1}, {MaxOperations: 1}, {MaxRows: -1}, {Stop: func() bool { return true }}} {
		result, err := layout.Irredundant(full, dc, options)
		if err == nil || result.Cover != nil || !(errors.Is(err, ErrDimension) || errors.Is(err, ErrLimit) || errors.Is(err, ErrCancelled)) {
			t.Fatal(options, result, err)
		}
	}
	split, _ := layout.SplitIrredundant(full, dc, EspressoIrredundantOptions{})
	split.PartiallyRedundantIndices = append(split.PartiallyRedundantIndices, 99)
	if _, err := layout.DeriveIrredundantTable(dc, split, EspressoIrredundantOptions{}); !errors.Is(err, ErrInvariant) {
		t.Fatal(err)
	}
}

func FuzzEspressoIrredundant(f *testing.F) {
	f.Add([]byte{3, 4, 1, 0x55, 0xaa, 0x33})
	f.Add([]byte{1, 1, 0, 3})
	f.Fuzz(func(t *testing.T, data []byte) {
		if len(data) < 3 {
			return
		}
		variables := int(data[0]%3) + 1
		layout, _ := NewEspressoCubeLayout(slices.Repeat([]int{2}, variables), variables)
		onCount, dcCount := int(data[1]%7), int(data[2]%4)
		makeMasks := func(count, offset int) []uint64 {
			masks := make([]uint64, count)
			for cube := range masks {
				for variable := 0; variable < variables; variable++ {
					masks[cube] |= uint64(data[(offset+cube*variables+variable)%len(data)]%3+1) << uint(2*variable)
				}
			}
			return masks
		}
		onMasks := makeMasks(onCount, 3)
		dcMasks := makeMasks(dcCount, 3+onCount*variables)
		on := espressoMapFamily(t, layout.BitCount(), onMasks...)
		dc := espressoMapFamily(t, layout.BitCount(), dcMasks...)
		wantE, wantT, wantP, wantRows := espressoIrredModel(layout, onMasks, dcMasks)
		for _, exact := range []bool{false, true} {
			result, err := layout.Irredundant(on, dc, EspressoIrredundantOptions{Exact: exact, MaxRows: 1 << 14})
			if err != nil {
				t.Fatal(err)
			}
			gotRows := slices.Clone(result.Table.Rows)
			espressoSortRows(gotRows)
			expectedRows := make([][]int, len(wantRows))
			for index := range wantRows {
				expectedRows[index] = slices.Clone(wantRows[index])
			}
			espressoSortRows(expectedRows)
			if !slices.Equal(result.Split.EssentialIndices, wantE) || !slices.Equal(result.Split.TotallyRedundantIndices, wantT) || !slices.Equal(result.Split.PartiallyRedundantIndices, wantP) || !slices.EqualFunc(gotRows, expectedRows, slices.Equal) {
				t.Fatal(onMasks, dcMasks, result.Split, gotRows, wantE, wantT, wantP, expectedRows)
			}
			oldTruth := espressoFamilyTruth(t, layout, on) | espressoFamilyTruth(t, layout, dc)
			newTruth := espressoFamilyTruth(t, layout, result.Cover) | espressoFamilyTruth(t, layout, dc)
			if oldTruth != newTruth {
				t.Fatal(oldTruth, newTruth, result.Selection)
			}
		}
	})
}
