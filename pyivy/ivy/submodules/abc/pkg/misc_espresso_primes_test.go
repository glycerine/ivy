package abc

// Independent exhaustive prime-implicant tests for src/misc/espresso/primes.c.

import (
	"errors"
	"slices"
	"testing"
)

func espressoAllProductCubes(layout *EspressoCubeLayout) []uint64 {
	result := []uint64{0}
	for variable, size := range layout.partSizes {
		start := layout.firstParts[variable]
		next := make([]uint64, 0, len(result)*((1<<uint(size))-1))
		for _, prefix := range result {
			for subset := 1; subset < 1<<uint(size); subset++ {
				next = append(next, prefix|uint64(subset)<<uint(start))
			}
		}
		result = next
	}
	return result
}

func espressoPrimeModel(layout *EspressoCubeLayout, function uint64) []uint64 {
	all := espressoAllProductCubes(layout)
	implicants := make([]uint64, 0, len(all))
	for _, cube := range all {
		truth := espressoCubeTruth(layout, cube)
		if truth != 0 && truth&^function == 0 {
			implicants = append(implicants, cube)
		}
	}
	primes := make([]uint64, 0, len(implicants))
	for _, cube := range implicants {
		maximal := true
		for _, other := range implicants {
			if cube != other && cube&^other == 0 {
				maximal = false
				break
			}
		}
		if maximal {
			primes = append(primes, cube)
		}
	}
	slices.Sort(primes)
	return primes
}

func espressoPrimeMasks(t testing.TB, family *EspressoSetFamily) []uint64 {
	t.Helper()
	result := espressoFamilyMasks(t, family)
	slices.Sort(result)
	return result
}

func TestEspressoPrimeConsensusSpecialCasesAndMixedLayout(t *testing.T) {
	layout, _ := NewEspressoCubeLayout([]int{2, 3}, 1)
	cases := [][]uint64{
		nil,
		{0b01 | 0b001<<2},
		{0b11 | 0b111<<2},
		{0b01 | 0b011<<2, 0b10 | 0b110<<2},
		{0b01 | 0b001<<2, 0b01 | 0b010<<2, 0b10 | 0b100<<2},
	}
	for index, masks := range cases {
		family := espressoMapFamily(t, layout.BitCount(), masks...)
		list, err := NewEspressoCubeList(layout, family)
		if err != nil {
			t.Fatal(err)
		}
		got, err := layout.PrimeConsensus(list, EspressoPrimeOptions{})
		if err != nil {
			t.Fatalf("case %d: %v", index, err)
		}
		function := espressoFamilyTruth(t, layout, family)
		want := espressoPrimeModel(layout, function)
		if gotMasks := espressoPrimeMasks(t, got); !slices.Equal(gotMasks, want) {
			t.Fatalf("case %d: got %b want %b function %b", index, gotMasks, want, function)
		}
		if !slices.Equal(espressoFamilyMasks(t, family), masks) {
			t.Fatal("input family mutated")
		}
	}
}

func TestEspressoPrimeConsensusRecursiveBinary(t *testing.T) {
	layout, _ := NewEspressoCubeLayout([]int{2, 2, 2}, 3)
	// Three-input parity forces binate recursion and cross-branch consensus.
	masks := []uint64{
		0b10 | 0b01<<2 | 0b01<<4,
		0b01 | 0b10<<2 | 0b01<<4,
		0b01 | 0b01<<2 | 0b10<<4,
		0b10 | 0b10<<2 | 0b10<<4,
	}
	family := espressoMapFamily(t, 6, masks...)
	list, _ := NewEspressoCubeList(layout, family)
	got, err := layout.PrimeConsensus(list, EspressoPrimeOptions{})
	want := espressoPrimeModel(layout, espressoFamilyTruth(t, layout, family))
	if err != nil || !slices.Equal(espressoPrimeMasks(t, got), want) {
		t.Fatal(espressoPrimeMasks(t, got), want, err)
	}
}

func TestEspressoPrimeConsensusErrorsLimitsAndCancellation(t *testing.T) {
	layout, _ := NewEspressoCubeLayout([]int{2, 2, 2}, 3)
	family := espressoMapFamily(t, 6, 0b010101, 0b101010, 0b100110, 0b011001)
	list, _ := NewEspressoCubeList(layout, family)
	if result, err := layout.PrimeConsensus(list, EspressoPrimeOptions{MaxCubes: 1}); result != nil || !errors.Is(err, ErrLimit) {
		t.Fatal(result, err)
	}
	if result, err := layout.PrimeConsensus(list, EspressoPrimeOptions{MaxOperations: 1}); result != nil || !errors.Is(err, ErrLimit) {
		t.Fatal(result, err)
	}
	if result, err := layout.PrimeConsensus(list, EspressoPrimeOptions{MaxDepth: 1}); result != nil || !errors.Is(err, ErrLimit) {
		t.Fatal(result, err)
	}
	if result, err := layout.PrimeConsensus(list, EspressoPrimeOptions{Stop: func() bool { return true }}); result != nil || !errors.Is(err, ErrCancelled) {
		t.Fatal(result, err)
	}
	if _, err := layout.PrimeConsensus(list, EspressoPrimeOptions{MaxDepth: -1}); !errors.Is(err, ErrDimension) {
		t.Fatal(err)
	}
	wrongLayout, _ := NewEspressoCubeLayout([]int{2, 4}, 1)
	if result, err := wrongLayout.PrimeConsensus(list, EspressoPrimeOptions{}); result != nil || !errors.Is(err, ErrDimension) {
		t.Fatal(result, err)
	}
}

func FuzzEspressoPrimeConsensus(f *testing.F) {
	f.Add([]byte{2, 3, 0x55, 0xaa})
	f.Add([]byte{4, 5, 0xff, 0x01, 0x7e})
	f.Fuzz(func(t *testing.T, data []byte) {
		if len(data) < 3 {
			return
		}
		variables := int(data[0]%4) + 1
		layout, err := NewEspressoCubeLayout(slices.Repeat([]int{2}, variables), variables)
		if err != nil {
			t.Fatal(err)
		}
		cubeCount := int(data[1] % 8)
		masks := make([]uint64, cubeCount)
		for cube := range masks {
			for variable := 0; variable < variables; variable++ {
				choice := data[(2+cube*variables+variable)%len(data)] % 3
				masks[cube] |= uint64(choice+1) << uint(2*variable)
			}
		}
		family := espressoMapFamily(t, layout.BitCount(), masks...)
		list, err := NewEspressoCubeList(layout, family)
		if err != nil {
			t.Fatal(err)
		}
		got, err := layout.PrimeConsensus(list, EspressoPrimeOptions{})
		if err != nil {
			t.Fatal(err)
		}
		function := espressoFamilyTruth(t, layout, family)
		want := espressoPrimeModel(layout, function)
		if gotMasks := espressoPrimeMasks(t, got); !slices.Equal(gotMasks, want) || !slices.Equal(espressoFamilyMasks(t, family), masks) {
			t.Fatal(gotMasks, want, function)
		}
	})
}
