package abc

// Go-only model and fuzz tests for src/misc/espresso/map.c.

import (
	"errors"
	"slices"
	"testing"
)

func espressoMapFamily(t testing.TB, bitCount int, masks ...uint64) *EspressoSetFamily {
	t.Helper()
	family, err := NewEspressoSetFamily(len(masks), bitCount)
	if err != nil {
		t.Fatal(err)
	}
	for _, mask := range masks {
		if err = family.Add(espressoSetFromMask(t, bitCount, mask)); err != nil {
			t.Fatal(err)
		}
	}
	return family
}

func espressoMintermModel(layout *EspressoCubeLayout, masks []uint64) []int {
	total := 1
	for _, size := range layout.partSizes {
		total *= size
	}
	result := make([]int, 0)
	for assignment := 0; assignment < total; assignment++ {
		matched := false
		for _, mask := range masks {
			remaining, valid := assignment, true
			for variable, size := range layout.partSizes {
				part := remaining % size
				remaining /= size
				if mask&(uint64(1)<<uint(layout.firstParts[variable]+part)) == 0 {
					valid = false
					break
				}
			}
			if valid {
				matched = true
				break
			}
		}
		if matched {
			result = append(result, assignment)
		}
	}
	return result
}

func TestEspressoMintermsMixedRadixAndOwnership(t *testing.T) {
	layout, err := NewEspressoCubeLayout([]int{2, 2, 3}, 2)
	if err != nil {
		t.Fatal(err)
	}
	masks := []uint64{1<<0 | 1<<1 | 1<<3 | 1<<4 | 1<<6, 1<<1 | 1<<2 | 1<<5}
	family := espressoMapFamily(t, layout.BitCount(), masks...)
	minterms, err := layout.Minterms(family, EspressoMapOptions{})
	if err != nil {
		t.Fatal(err)
	}
	elements, err := minterms.Elements()
	if err != nil || !slices.Equal(elements, []int{2, 3, 5, 10, 11}) || minterms.BitCount() != 12 {
		t.Fatal(elements, minterms.BitCount(), err)
	}
	if !slices.Equal(espressoFamilyMasks(t, family), masks) {
		t.Fatal("input family mutated")
	}
}

func TestEspressoKarnaughIndexAndFormat(t *testing.T) {
	wantIndex := [...]uint8{
		0, 1, 3, 2, 16, 17, 19, 18, 80, 81, 83, 82, 64, 65, 67, 66,
		4, 5, 7, 6, 20, 21, 23, 22, 84, 85, 87, 86, 68, 69, 71, 70,
		12, 13, 15, 14, 28, 29, 31, 30, 92, 93, 95, 94, 76, 77, 79, 78,
		8, 9, 11, 10, 24, 25, 27, 26, 88, 89, 91, 90, 72, 73, 75, 74,
		32, 33, 35, 34, 48, 49, 51, 50, 112, 113, 115, 114, 96, 97, 99, 98,
		36, 37, 39, 38, 52, 53, 55, 54, 116, 117, 119, 118, 100, 101, 103, 102,
		44, 45, 47, 46, 60, 61, 63, 62, 124, 125, 127, 126, 108, 109, 111, 110,
		40, 41, 43, 42, 56, 57, 59, 58, 120, 121, 123, 122, 104, 105, 107, 106,
		160, 161, 163, 162, 176, 177, 179, 178, 240, 241, 243, 242, 224, 225, 227, 226,
		164, 165, 167, 166, 180, 181, 183, 182, 244, 245, 247, 246, 228, 229, 231, 230,
		172, 173, 175, 174, 188, 189, 191, 190, 252, 253, 255, 254, 236, 237, 239, 238,
		168, 169, 171, 170, 184, 185, 187, 186, 248, 249, 251, 250, 232, 233, 235, 234,
		128, 129, 131, 130, 144, 145, 147, 146, 208, 209, 211, 210, 192, 193, 195, 194,
		132, 133, 135, 134, 148, 149, 151, 150, 212, 213, 215, 214, 196, 197, 199, 198,
		140, 141, 143, 142, 156, 157, 159, 158, 220, 221, 223, 222, 204, 205, 207, 206,
		136, 137, 139, 138, 152, 153, 155, 154, 216, 217, 219, 218, 200, 201, 203, 202,
	}
	for coordinate, want := range wantIndex {
		if got := espressoKarnaughIndex(coordinate/16, coordinate%16); got != int(want) {
			t.Fatalf("index[%d,%d]=%d want %d", coordinate/16, coordinate%16, got, want)
		}
	}
	layout, err := NewEspressoCubeLayout([]int{2, 2, 2}, 2)
	if err != nil {
		t.Fatal(err)
	}
	// Output zero is true at inputs zero and three; output one at one and two.
	family := espressoMapFamily(t, 6,
		1<<0|1<<2|1<<4,
		1<<1|1<<3|1<<4,
		1<<1|1<<2|1<<5,
		1<<0|1<<3|1<<5,
	)
	formatted, err := layout.FormatKarnaughMap(family, EspressoMapOptions{})
	if err != nil {
		t.Fatal(err)
	}
	want := "\n\nOutput space # 0\n1.1.        \n                        \n\nOutput space # 1\n.1.1        \n                        "
	if formatted != want {
		t.Fatalf("format mismatch\ngot  %q\nwant %q", formatted, want)
	}
}

func TestEspressoMapErrorsLimitsAndCancellation(t *testing.T) {
	layout, _ := NewEspressoCubeLayout([]int{2, 2}, 1)
	family := espressoMapFamily(t, 4, 0b1111)
	if result, err := layout.Minterms(family, EspressoMapOptions{MaxCubes: 0, MaxMinterms: 3}); result != nil || !errors.Is(err, ErrLimit) {
		t.Fatal(result, err)
	}
	if result, err := layout.Minterms(family, EspressoMapOptions{MaxVisits: 1}); result != nil || !errors.Is(err, ErrLimit) {
		t.Fatal(result, err)
	}
	twoCubes := espressoMapFamily(t, 4, 0b1111, 0b0101)
	if result, err := layout.Minterms(twoCubes, EspressoMapOptions{MaxCubes: 1}); result != nil || !errors.Is(err, ErrLimit) {
		t.Fatal(result, err)
	}
	if result, err := layout.Minterms(family, EspressoMapOptions{Stop: func() bool { return true }}); result != nil || !errors.Is(err, ErrCancelled) {
		t.Fatal(result, err)
	}
	wrong := espressoMapFamily(t, 3, 0b111)
	if result, err := layout.Minterms(wrong, EspressoMapOptions{}); result != nil || !errors.Is(err, ErrDimension) {
		t.Fatal(result, err)
	}
	nonMap, _ := NewEspressoCubeLayout([]int{2, 3, 2}, 1)
	if text, err := nonMap.FormatKarnaughMap(espressoMapFamily(t, 7, 0), EspressoMapOptions{}); text != "" || !errors.Is(err, ErrDimension) {
		t.Fatal(text, err)
	}
	if _, err := layout.Minterms(family, EspressoMapOptions{MaxVisits: -1}); !errors.Is(err, ErrDimension) {
		t.Fatal(err)
	}
}

func FuzzEspressoMinterms(f *testing.F) {
	f.Add([]byte{2, 2, 2, 0x15, 0x2a})
	f.Add([]byte{4, 3, 1, 0xff, 0x00, 0x81})
	f.Fuzz(func(t *testing.T, data []byte) {
		if len(data) < 3 {
			return
		}
		binary := int(data[0]%4) + 1
		outputs := int(data[1]%3) + 1
		layout, err := NewEspressoCubeLayout(append(slices.Repeat([]int{2}, binary), outputs), binary)
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
		got, err := layout.Minterms(family, EspressoMapOptions{})
		if err != nil {
			t.Fatal(err)
		}
		elements, err := got.Elements()
		want := espressoMintermModel(layout, masks)
		if err != nil || !slices.Equal(elements, want) || !slices.Equal(espressoFamilyMasks(t, family), masks) {
			t.Fatal(elements, want, err)
		}
	})
}
