package abc

// Go-only assignment models for src/misc/espresso/compl.c.

import (
	"errors"
	"slices"
	"testing"
)

func espressoTruthUniverse(layout *EspressoCubeLayout) uint64 {
	assignments := 1
	for _, size := range layout.partSizes {
		assignments *= size
	}
	return uint64(1)<<uint(assignments) - 1
}

func TestEspressoComplementEmptyFullCubeAndOwnership(t *testing.T) {
	layout, _ := NewEspressoCubeLayout([]int{2, 3}, 1)
	empty := espressoMapFamily(t, 5)
	fullFamily := espressoMapFamily(t, 5, 0b11111)
	emptyComplement, err := layout.ComplementCover(empty, EspressoSharpOptions{})
	fullComplement, fullErr := layout.ComplementCover(fullFamily, EspressoSharpOptions{})
	if err != nil || fullErr != nil || espressoFamilyTruth(t, layout, emptyComplement) != espressoTruthUniverse(layout) || espressoFamilyTruth(t, layout, fullComplement) != 0 {
		t.Fatal(espressoFamilyMasks(t, emptyComplement), espressoFamilyMasks(t, fullComplement), err, fullErr)
	}
	cube := espressoSetFromMask(t, 5, 0b01|0b101<<2)
	cube.Rank, cube.Flags = 7, EspressoSetPrime|EspressoSetActive
	source, _ := NewEspressoSetFamily(1, 5)
	_ = source.Add(cube)
	before := espressoFamilyMasks(t, source)
	complement, err := layout.ComplementCover(source, EspressoSharpOptions{})
	want := espressoTruthUniverse(layout) &^ espressoFamilyTruth(t, layout, source)
	if err != nil || espressoFamilyTruth(t, layout, complement) != want || !slices.Equal(espressoFamilyMasks(t, source), before) {
		t.Fatal(espressoFamilyMasks(t, complement), want, err)
	}
}

func TestEspressoSimplifyAndComplementSemanticPair(t *testing.T) {
	layout, _ := NewEspressoCubeLayout([]int{2, 2, 2}, 2)
	cover := espressoMapFamily(t, 6,
		0b01|0b11<<2|0b01<<4,
		0b10|0b01<<2|0b11<<4,
		0b01|0b01<<2|0b01<<4,
	)
	want := espressoFamilyTruth(t, layout, cover)
	simplified, complement, err := layout.SimplifyAndComplementCover(cover, EspressoSharpOptions{})
	if err != nil || espressoFamilyTruth(t, layout, simplified) != want || espressoFamilyTruth(t, layout, complement) != espressoTruthUniverse(layout)&^want {
		t.Fatal(espressoFamilyMasks(t, simplified), espressoFamilyMasks(t, complement), err)
	}
	standalone, err := layout.SimplifyCover(cover, EspressoSharpOptions{})
	if err != nil || espressoFamilyTruth(t, layout, standalone) != want {
		t.Fatal(espressoFamilyMasks(t, standalone), err)
	}
}

func TestEspressoComplementErrorsLimitsAndCancellation(t *testing.T) {
	layout, _ := NewEspressoCubeLayout([]int{2, 2}, 2)
	cover := espressoMapFamily(t, 4, 0b0101, 0b1010)
	if _, err := layout.ComplementCover(cover, EspressoSharpOptions{MaxOperations: 1}); !errors.Is(err, ErrLimit) {
		t.Fatal(err)
	}
	if _, err := layout.ComplementCover(cover, EspressoSharpOptions{MaxCubes: 1}); !errors.Is(err, ErrLimit) {
		t.Fatal(err)
	}
	if _, err := layout.ComplementCover(cover, EspressoSharpOptions{Stop: func() bool { return true }}); !errors.Is(err, ErrCancelled) {
		t.Fatal(err)
	}
	if _, _, err := layout.SimplifyAndComplementCover(cover, EspressoSharpOptions{MaxOperations: 1}); !errors.Is(err, ErrLimit) {
		t.Fatal(err)
	}
	wrong := espressoMapFamily(t, 3, 0b111)
	if _, err := layout.ComplementCover(wrong, EspressoSharpOptions{}); !errors.Is(err, ErrDimension) {
		t.Fatal(err)
	}
	if _, err := layout.ComplementCover(cover, EspressoSharpOptions{MaxCubes: -1}); !errors.Is(err, ErrDimension) {
		t.Fatal(err)
	}
}

func FuzzEspressoComplementAndSimplify(f *testing.F) {
	f.Add([]byte{3, 4, 0x55, 0xaa, 0x33})
	f.Add([]byte{1, 2, 0xff})
	f.Fuzz(func(t *testing.T, data []byte) {
		if len(data) < 2 {
			return
		}
		variables := int(data[0]%4) + 1
		layout, err := NewEspressoCubeLayout(slices.Repeat([]int{2}, variables), variables)
		if err != nil {
			t.Fatal(err)
		}
		count := int(data[1] % 7)
		masks := make([]uint64, count)
		for cube := range masks {
			for variable := 0; variable < variables; variable++ {
				masks[cube] |= uint64(data[(2+cube*variables+variable)%len(data)]%3+1) << uint(2*variable)
			}
		}
		cover := espressoMapFamily(t, layout.BitCount(), masks...)
		before := slices.Clone(masks)
		truth := espressoFamilyTruth(t, layout, cover)
		universe := espressoTruthUniverse(layout)
		simplified, complement, err := layout.SimplifyAndComplementCover(cover, EspressoSharpOptions{})
		if err != nil {
			t.Fatal(err)
		}
		if espressoFamilyTruth(t, layout, complement) != universe&^truth || espressoFamilyTruth(t, layout, simplified) != truth || !slices.Equal(espressoFamilyMasks(t, cover), before) {
			t.Fatal(truth, espressoFamilyMasks(t, simplified), espressoFamilyMasks(t, complement), masks)
		}
	})
}
