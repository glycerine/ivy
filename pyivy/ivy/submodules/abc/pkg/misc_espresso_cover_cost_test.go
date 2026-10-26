package abc

// Go-only model and fuzz tests for src/misc/espresso/cvrmisc.c.

import (
	"errors"
	"slices"
	"testing"
)

func espressoCoverCostModel(layout *EspressoCubeLayout, masks []uint64, sparse []bool, primes []bool) EspressoCoverCost {
	zeros := make([]int, layout.VariableCount())
	for _, mask := range masks {
		for variable, size := range layout.partSizes {
			for part := 0; part < size; part++ {
				if mask&(uint64(1)<<uint(layout.firstParts[variable]+part)) == 0 {
					zeros[variable]++
				}
			}
		}
	}
	cost := EspressoCoverCost{Cubes: len(masks)}
	for variable := 0; variable < layout.binaryVariables; variable++ {
		cost.Input += zeros[variable]
	}
	for variable := layout.binaryVariables; variable < layout.VariableCount()-1; variable++ {
		if sparse[variable] {
			cost.MultipleValued += len(masks)*layout.partSizes[variable] - zeros[variable]
		} else {
			cost.MultipleValued += zeros[variable]
		}
	}
	if layout.binaryVariables != layout.VariableCount() {
		output := layout.VariableCount() - 1
		cost.Output = len(masks)*layout.partSizes[output] - zeros[output]
	}
	for _, prime := range primes {
		if prime {
			cost.Primes++
		}
	}
	cost.Total = cost.Input + cost.Output + cost.MultipleValued
	return cost
}

func TestEspressoCoverCostFormatsTotalsAndOwnership(t *testing.T) {
	layout, err := NewEspressoCubeLayout([]int{2, 2, 3, 2}, 2)
	if err != nil {
		t.Fatal(err)
	}
	masks := []uint64{0b11 | 0b01<<2 | 0b101<<4 | 0b01<<7, 0b01 | 0b11<<2 | 0b010<<4 | 0b11<<7}
	family := espressoMapFamily(t, 9, masks...)
	family.sets[0].Flags |= EspressoSetPrime
	sparse := []bool{false, false, true, true}
	cost, err := layout.CoverCost(family, EspressoCoverCostOptions{SparseVariables: sparse})
	want := EspressoCoverCost{Cubes: 2, Input: 2, Output: 3, MultipleValued: 3, Total: 8, Primes: 1}
	if err != nil || cost != want || cost.Clone() != cost {
		t.Fatal(cost, want, err)
	}
	formatted, err := layout.FormatCoverCost(cost)
	size, sizeErr := layout.FormatCoverSize("REDUCE", family, EspressoCoverCostOptions{SparseVariables: sparse})
	trace, traceErr := layout.FormatCoverTrace("EXPAND", 1234, family, EspressoCoverCostOptions{SparseVariables: sparse})
	if err != nil || sizeErr != nil || traceErr != nil || formatted != "c=2(1) in=2 mv=3 out=3" || size != "# REDUCE\tCost is c=2(1) in=2 mv=3 out=3\n" || trace != "# EXPAND\tTime was 1.23 sec, cost is c=2(1) in=2 mv=3 out=3\n" {
		t.Fatal(formatted, size, trace, err, sizeErr, traceErr)
	}
	totals := NewEspressoCoverTotals([]string{"READ", "COMPL"})
	recorded, line, err := totals.Record(1, 2345, layout, family, EspressoCoverCostOptions{SparseVariables: sparse}, true)
	if err != nil || recorded != cost || line != "# COMPL\tTime was 2.34 sec, cost is c=2(1) in=2 mv=3 out=3\n" || !slices.Equal(totals.Milliseconds, []int64{0, 2345}) || !slices.Equal(totals.Calls, []int{0, 1}) {
		t.Fatal(recorded, line, totals, err)
	}
	if !slices.Equal(espressoFamilyMasks(t, family), masks) || !slices.Equal(sparse, []bool{false, false, true, true}) {
		t.Fatal("inputs mutated")
	}

	binaryLayout, _ := NewEspressoCubeLayout([]int{2, 2, 2}, 2)
	binaryFamily := espressoMapFamily(t, 6, 0b010101)
	binaryCost, _ := binaryLayout.CoverCost(binaryFamily, EspressoCoverCostOptions{})
	binaryText, _ := binaryLayout.FormatCoverCost(binaryCost)
	if binaryText != "c=1(1) in=2 out=1 tot=3" {
		t.Fatal(binaryText, binaryCost)
	}
}

func TestEspressoCoverCostErrorsLimitsAndAtomicTotals(t *testing.T) {
	layout, _ := NewEspressoCubeLayout([]int{2, 2}, 1)
	family := espressoMapFamily(t, 4, 0b1111, 0b0101)
	cases := []EspressoCoverCostOptions{
		{SparseVariables: []bool{true}},
		{MaxCubes: 1},
		{MaxChecks: 1},
		{MaxChecks: -1},
		{Stop: func() bool { return true }},
	}
	for index, options := range cases {
		if _, err := layout.CoverCost(family, options); err == nil {
			t.Fatalf("case %d", index)
		}
	}
	wrong := espressoMapFamily(t, 3, 0b111)
	if _, err := layout.CoverCost(wrong, EspressoCoverCostOptions{}); !errors.Is(err, ErrDimension) {
		t.Fatal(err)
	}
	if _, err := layout.FormatCoverCost(EspressoCoverCost{Cubes: 1, Total: 2}); !errors.Is(err, ErrInvariant) {
		t.Fatal(err)
	}
	if _, err := formatEspressoMilliseconds(-1); !errors.Is(err, ErrDimension) {
		t.Fatal(err)
	}
	totals := NewEspressoCoverTotals([]string{"X"})
	beforeMilliseconds, beforeCalls := slices.Clone(totals.Milliseconds), slices.Clone(totals.Calls)
	if _, _, err := totals.Record(0, 2, layout, family, EspressoCoverCostOptions{MaxChecks: 1}, true); !errors.Is(err, ErrLimit) || !slices.Equal(totals.Milliseconds, beforeMilliseconds) || !slices.Equal(totals.Calls, beforeCalls) {
		t.Fatal(err, totals)
	}
}

func FuzzEspressoCoverCost(f *testing.F) {
	f.Add([]byte{2, 3, 2, 0x55, 0xaa, 1})
	f.Add([]byte{1, 2, 1, 0xff})
	f.Fuzz(func(t *testing.T, data []byte) {
		if len(data) < 3 {
			return
		}
		binary := int(data[0]%3) + 1
		multiple := int(data[1]%3) + 1
		parts := slices.Repeat([]int{2}, binary)
		for variable := 0; variable < multiple; variable++ {
			parts = append(parts, int(data[(2+variable)%len(data)]%4)+1)
		}
		layout, err := NewEspressoCubeLayout(parts, binary)
		if err != nil {
			t.Fatal(err)
		}
		cubeCount := int(data[2] % 8)
		masks, primes := make([]uint64, cubeCount), make([]bool, cubeCount)
		family, _ := NewEspressoSetFamily(cubeCount, layout.BitCount())
		for cube := range masks {
			for part := 0; part < layout.BitCount(); part++ {
				position := 3 + cube*(layout.BitCount()+1) + part
				if position < len(data) && data[position]&1 != 0 {
					masks[cube] |= uint64(1) << uint(part)
				}
			}
			primePosition := 3 + cube*(layout.BitCount()+1) + layout.BitCount()
			primes[cube] = primePosition < len(data) && data[primePosition]&1 != 0
			set := espressoSetFromMask(t, layout.BitCount(), masks[cube])
			if primes[cube] {
				set.Flags = EspressoSetPrime
			}
			if err = family.Add(set); err != nil {
				t.Fatal(err)
			}
		}
		sparse := make([]bool, layout.VariableCount())
		for variable := binary; variable < layout.VariableCount(); variable++ {
			sparse[variable] = data[(3+variable)%len(data)]&2 != 0
		}
		got, err := layout.CoverCost(family, EspressoCoverCostOptions{SparseVariables: sparse})
		want := espressoCoverCostModel(layout, masks, sparse, primes)
		if err != nil || got != want || !slices.Equal(espressoFamilyMasks(t, family), masks) {
			t.Fatal(got, want, err)
		}
	})
}
