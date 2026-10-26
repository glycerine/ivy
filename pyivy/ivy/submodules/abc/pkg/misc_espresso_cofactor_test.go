package abc

// Go-only model and fuzz tests for src/misc/espresso/cofactor.c.

import (
	"errors"
	"math/bits"
	"slices"
	"testing"
)

func espressoCubeListMasks(t testing.TB, list *EspressoCubeList) ([]uint64, uint64) {
	t.Helper()
	count, err := list.Count()
	if err != nil {
		t.Fatal(err)
	}
	result := make([]uint64, count)
	for index := range result {
		cube, cubeErr := list.Cube(index)
		if cubeErr != nil {
			t.Fatal(cubeErr)
		}
		result[index] = espressoSetMask(t, cube)
	}
	cofactor, err := list.CofactorMask()
	if err != nil {
		t.Fatal(err)
	}
	return result, espressoSetMask(t, cofactor)
}

func espressoDistanceZeroMask(layout *EspressoCubeLayout, first, second uint64) bool {
	for variable, size := range layout.partSizes {
		mask := (uint64(1)<<uint(size) - 1) << uint(layout.firstParts[variable])
		if first&second&mask == 0 {
			return false
		}
	}
	return true
}

func TestEspressoCubeListConstructCofactorUnlistAndOwnership(t *testing.T) {
	layout, _ := NewEspressoCubeLayout([]int{2, 2, 3}, 2)
	firstMasks := []uint64{0b01 | 0b11<<2 | 0b101<<4, 0b10 | 0b01<<2 | 0b110<<4}
	secondMasks := []uint64{0b11 | 0b10<<2 | 0b011<<4}
	first := espressoMapFamily(t, 7, firstMasks...)
	second := espressoMapFamily(t, 7, secondMasks...)
	list, err := NewEspressoCubeList(layout, first, second)
	if err != nil {
		t.Fatal(err)
	}
	masks, cofactor := espressoCubeListMasks(t, list)
	if !slices.Equal(masks, append(slices.Clone(firstMasks), secondMasks...)) || cofactor != 0 {
		t.Fatal(masks, cofactor)
	}
	cube := espressoSetFromMask(t, 7, 0b11|0b01<<2|0b111<<4)
	cofactored, err := list.Cofactor(cube, -1, EspressoCofactorOptions{})
	if err != nil {
		t.Fatal(err)
	}
	got, gotMask := espressoCubeListMasks(t, cofactored)
	if !slices.Equal(got, []uint64{firstMasks[0], firstMasks[1]}) || gotMask != 0b10<<2 {
		t.Fatal(got, gotMask)
	}
	unlisted, err := cofactored.Unlist(EspressoCofactorOptions{})
	if err != nil || !slices.Equal(espressoFamilyMasks(t, unlisted), []uint64{firstMasks[0] | gotMask, firstMasks[1] | gotMask}) {
		t.Fatal(espressoFamilyMasks(t, unlisted), err)
	}
	excluded, err := list.CofactorAt(0, EspressoCofactorOptions{})
	if err != nil {
		t.Fatal(err)
	}
	excludedMasks, _ := espressoCubeListMasks(t, excluded)
	if slices.Contains(excludedMasks, firstMasks[0]) {
		t.Fatal("source cube was not excluded", excludedMasks)
	}
	if !slices.Equal(espressoFamilyMasks(t, first), firstMasks) || !slices.Equal(espressoFamilyMasks(t, second), secondMasks) {
		t.Fatal("source family mutated")
	}
}

func TestEspressoCubeListSpecializedCountsSplitAndSimplify(t *testing.T) {
	layout, _ := NewEspressoCubeLayout([]int{2, 2, 3}, 2)
	masks := []uint64{0b01 | 0b01<<2 | 0b001<<4, 0b10 | 0b01<<2 | 0b011<<4, 0b11 | 0b10<<2 | 0b100<<4, 0b01 | 0b01<<2 | 0b101<<4}
	family := espressoMapFamily(t, 7, masks...)
	family.sets[0].Rank, family.sets[0].Flags = 11, EspressoSetPrime
	family.sets[3].Rank = 22
	list, _ := NewEspressoCubeList(layout, family)
	active := espressoSetFromMask(t, 7, 0b11|0b01<<2|0b111<<4)
	general, err := list.Cofactor(active, -1, EspressoCofactorOptions{})
	specialized, specializedErr := list.SingleVariableCofactor(active, 1, -1, EspressoCofactorOptions{})
	generalMasks, generalCofactor := espressoCubeListMasks(t, general)
	specialMasks, specialCofactor := espressoCubeListMasks(t, specialized)
	if err != nil || specializedErr != nil || !slices.Equal(generalMasks, specialMasks) || generalCofactor != specialCofactor {
		t.Fatal(generalMasks, specialMasks, generalCofactor, specialCofactor, err, specializedErr)
	}
	counts, err := general.Analyze(EspressoCofactorOptions{})
	if err != nil || !slices.Equal(counts.PartZeros, []int{1, 2, 0, 0, 0, 2, 2}) || !slices.Equal(counts.VariableZeros, []int{3, 0, 4}) || !slices.Equal(counts.PartsActive, []int{2, 0, 2}) || !slices.Equal(counts.IsUnate, []bool{false, false, false}) || counts.VariablesActive != 2 || counts.VariablesUnate != 0 || counts.Best != 2 {
		t.Fatal(counts, err)
	}
	variable, left, right, err := general.BinateSplit(counts)
	if err != nil || variable != 2 || espressoSetMask(t, left) != 0b1111|0b001<<4 || espressoSetMask(t, right) != 0b1111|0b110<<4 {
		t.Fatal(variable, espressoSetMask(t, left), espressoSetMask(t, right), err)
	}
	// Mask output part six so the first and fourth cubes become equal.
	maskOutputSix := espressoSetFromMask(t, 7, 0b1111|0b011<<4)
	masked, _ := list.Cofactor(maskOutputSix, -1, EspressoCofactorOptions{})
	simplified, err := masked.Simplify(EspressoCofactorOptions{})
	simplifiedMasks, _ := espressoCubeListMasks(t, simplified)
	retained, retainedErr := simplified.Cube(1)
	if err != nil || retainedErr != nil || !slices.Equal(simplifiedMasks, []uint64{masks[1], masks[0]}) || retained.Rank != 11 || retained.Flags != EspressoSetPrime {
		t.Fatal(simplifiedMasks, err)
	}
}

func TestEspressoCubeListErrorsLimitsAndCancellation(t *testing.T) {
	layout, _ := NewEspressoCubeLayout([]int{2, 2}, 1)
	family := espressoMapFamily(t, 4, 0b1111, 0b0101)
	list, _ := NewEspressoCubeList(layout, family)
	cube := espressoSetFromMask(t, 4, 0b1111)
	if result, err := list.Cofactor(cube, 2, EspressoCofactorOptions{}); result != nil || !errors.Is(err, ErrDimension) {
		t.Fatal(result, err)
	}
	if result, err := list.Cofactor(cube, -1, EspressoCofactorOptions{MaxCubes: 1}); result != nil || !errors.Is(err, ErrLimit) {
		t.Fatal(result, err)
	}
	if result, err := list.Cofactor(cube, -1, EspressoCofactorOptions{MaxChecks: 1}); result != nil || !errors.Is(err, ErrLimit) {
		t.Fatal(result, err)
	}
	if result, err := list.Cofactor(cube, -1, EspressoCofactorOptions{Stop: func() bool { return true }}); result != nil || !errors.Is(err, ErrCancelled) {
		t.Fatal(result, err)
	}
	notSingle := espressoSetFromMask(t, 4, 0b0101)
	if result, err := list.SingleVariableCofactor(notSingle, 0, -1, EspressoCofactorOptions{}); result != nil || !errors.Is(err, ErrDimension) {
		t.Fatal(result, err)
	}
	if _, err := list.Analyze(EspressoCofactorOptions{MaxChecks: 1}); !errors.Is(err, ErrLimit) {
		t.Fatal(err)
	}
	if _, _, _, err := list.BinateSplit(EspressoCubeCounts{}); !errors.Is(err, ErrDimension) {
		t.Fatal(err)
	}
	if result, err := list.Simplify(EspressoCofactorOptions{MaxChecks: 1}); result != nil || !errors.Is(err, ErrLimit) {
		t.Fatal(result, err)
	}
	if _, err := NewEspressoCubeList(layout); !errors.Is(err, ErrDimension) {
		t.Fatal(err)
	}
	emptyFamily := espressoMapFamily(t, 4)
	emptyList, _ := NewEspressoCubeList(layout, emptyFamily)
	emptyCounts, err := emptyList.Analyze(EspressoCofactorOptions{})
	if err != nil || emptyCounts.Best != 0 || emptyCounts.VariablesActive != 0 {
		t.Fatal(emptyCounts, err)
	}
}

func FuzzEspressoCubeListCofactorAndCounts(f *testing.F) {
	f.Add([]byte{2, 2, 3, 0x55, 0x2a, 0x7f})
	f.Add([]byte{3, 1, 1, 0xff})
	f.Add([]byte("100171"))
	f.Fuzz(func(t *testing.T, data []byte) {
		if len(data) < 4 {
			return
		}
		binary := int(data[0]%3) + 1
		mvSize := int(data[1]%4) + 1
		parts := append(slices.Repeat([]int{2}, binary), mvSize)
		layout, err := NewEspressoCubeLayout(parts, binary)
		if err != nil {
			t.Fatal(err)
		}
		cubeCount := int(data[2] % 9)
		masks := make([]uint64, cubeCount)
		for cube := range masks {
			for part := 0; part < layout.BitCount(); part++ {
				position := 3 + cube*layout.BitCount() + part
				if position < len(data) && data[position]&1 != 0 {
					masks[cube] |= uint64(1) << uint(part)
				}
			}
		}
		family := espressoMapFamily(t, layout.BitCount(), masks...)
		list, err := NewEspressoCubeList(layout, family)
		if err != nil {
			t.Fatal(err)
		}
		fullMask := uint64(1)<<uint(layout.BitCount()) - 1
		var cofactorCubeMask uint64
		for part := 0; part < layout.BitCount(); part++ {
			if data[(3+part)%len(data)]&2 != 0 {
				cofactorCubeMask |= uint64(1) << uint(part)
			}
		}
		cofactorCubeMask &= fullMask
		cofactorCube := espressoSetFromMask(t, layout.BitCount(), cofactorCubeMask)
		got, err := list.Cofactor(cofactorCube, -1, EspressoCofactorOptions{})
		if err != nil {
			t.Fatal(err)
		}
		gotMasks, gotCofactor := espressoCubeListMasks(t, got)
		wantMasks := make([]uint64, 0, len(masks))
		for _, mask := range masks {
			if espressoDistanceZeroMask(layout, mask, cofactorCubeMask) {
				wantMasks = append(wantMasks, mask)
			}
		}
		wantCofactor := fullMask &^ cofactorCubeMask
		if !slices.Equal(gotMasks, wantMasks) || gotCofactor != wantCofactor {
			t.Fatal(gotMasks, wantMasks, gotCofactor, wantCofactor)
		}
		counts, err := got.Analyze(EspressoCofactorOptions{})
		if err != nil {
			t.Fatal(err)
		}
		wantPartZeros := make([]int, layout.BitCount())
		for _, mask := range wantMasks {
			for part := range wantPartZeros {
				if mask&(uint64(1)<<uint(part)) == 0 && wantCofactor&(uint64(1)<<uint(part)) == 0 {
					wantPartZeros[part]++
				}
			}
		}
		if !slices.Equal(counts.PartZeros, wantPartZeros) {
			t.Fatal(counts.PartZeros, wantPartZeros)
		}
		wantVariableZeros := make([]int, layout.VariableCount())
		wantPartsActive := make([]int, layout.VariableCount())
		wantUnate := make([]bool, layout.VariableCount())
		wantBest, variablesActive, variablesUnate := -1, 0, 0
		mostActive, mostZero, mostBalanced := 0, 0, 32000
		for current, size := range layout.partSizes {
			start := layout.firstParts[current]
			maxActive := 0
			for part := 0; part < size; part++ {
				zeros := wantPartZeros[start+part]
				wantVariableZeros[current] += zeros
				if zeros > 0 {
					wantPartsActive[current]++
				}
				if current < binary && zeros > maxActive {
					maxActive = zeros
				}
			}
			if current >= binary {
				maxActive = wantPartsActive[current]
			}
			active, zero := wantPartsActive[current], wantVariableZeros[current]
			if active > mostActive || active == mostActive && (zero > mostZero || zero == mostZero && maxActive < mostBalanced) {
				wantBest, mostActive, mostZero, mostBalanced = current, active, zero, maxActive
			}
			wantUnate[current] = active == 1
			if active > 0 {
				variablesActive++
			}
			if active == 1 {
				variablesUnate++
			}
		}
		if !slices.Equal(counts.VariableZeros, wantVariableZeros) || !slices.Equal(counts.PartsActive, wantPartsActive) || !slices.Equal(counts.IsUnate, wantUnate) || counts.Best != wantBest || counts.VariablesActive != variablesActive || counts.VariablesUnate != variablesUnate {
			t.Fatal(counts, wantVariableZeros, wantPartsActive, wantUnate, wantBest, variablesActive, variablesUnate)
		}
		variable := int(data[3] % byte(layout.VariableCount()))
		singleMask := fullMask
		start, size := layout.firstParts[variable], layout.partSizes[variable]
		for part := 0; part < size; part++ {
			if data[(4+part)%len(data)]&4 == 0 {
				singleMask &^= uint64(1) << uint(start+part)
			}
		}
		single := espressoSetFromMask(t, layout.BitCount(), singleMask)
		general, generalErr := list.Cofactor(single, -1, EspressoCofactorOptions{})
		specialized, specializedErr := list.SingleVariableCofactor(single, variable, -1, EspressoCofactorOptions{})
		if generalErr != nil || specializedErr != nil {
			t.Fatal(generalErr, specializedErr)
		}
		generalMasks, generalCofactor := espressoCubeListMasks(t, general)
		specializedMasks, specializedCofactor := espressoCubeListMasks(t, specialized)
		if !slices.Equal(generalMasks, specializedMasks) || generalCofactor != specializedCofactor {
			t.Fatal(generalMasks, specializedMasks, generalCofactor, specializedCofactor)
		}
		unlisted, err := got.Unlist(EspressoCofactorOptions{})
		if err != nil {
			t.Fatal(err)
		}
		wantUnlisted := slices.Clone(wantMasks)
		for index := range wantUnlisted {
			wantUnlisted[index] |= wantCofactor
		}
		if !slices.Equal(espressoFamilyMasks(t, unlisted), wantUnlisted) || bits.OnesCount64(gotCofactor) != bits.OnesCount64(wantCofactor) {
			t.Fatal(espressoFamilyMasks(t, unlisted), wantUnlisted)
		}
		type maskedItem struct {
			mask uint64
			key  uint64
		}
		items := make([]maskedItem, len(wantMasks))
		for index, mask := range wantMasks {
			items[index] = maskedItem{mask: mask, key: mask | wantCofactor}
		}
		slices.SortStableFunc(items, func(first, second maskedItem) int {
			if first.key > second.key {
				return -1
			}
			if first.key < second.key {
				return 1
			}
			return 0
		})
		wantSimplified := make([]uint64, 0, len(items))
		var previous uint64
		for index, item := range items {
			if index == 0 || item.key != previous {
				wantSimplified = append(wantSimplified, item.mask)
				previous = item.key
			}
		}
		simplified, err := got.Simplify(EspressoCofactorOptions{})
		if err != nil {
			t.Fatal(err)
		}
		simplifiedMasks, simplifiedCofactor := espressoCubeListMasks(t, simplified)
		if !slices.Equal(simplifiedMasks, wantSimplified) || simplifiedCofactor != wantCofactor {
			t.Fatal(simplifiedMasks, wantSimplified, simplifiedCofactor, wantCofactor)
		}
	})
}
