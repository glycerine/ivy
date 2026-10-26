package abc

// Tests for the remaining active transforms in src/misc/mvc/mvcUtils.c.

import (
	"encoding/hex"
	"encoding/json"
	"errors"
	"os"
	"slices"
	"testing"
)

type mvcUtilsTransformOracle struct {
	Cases []struct {
		Bits       int
		Source     string
		Diffs      string
		Remap      string
		Inverse    string
		Removed    string
		Cofactor   string
		Flipped    string
		Quantified string
		Transposed string
		Unused     int
	} `json:"cases"`
}

func mvcUtilsTransformCover(manager *MvcManager, bitCount int) (*MvcCover, error) {
	cover, err := NewMvcCover(manager, bitCount)
	if err != nil {
		return nil, err
	}
	for cubeIndex := 0; cubeIndex < 5; cubeIndex++ {
		cube, cubeErr := cover.NewCube()
		if cubeErr != nil {
			return nil, cubeErr
		}
		for variable := 0; variable < bitCount/2; variable++ {
			value := (variable+cubeIndex)%3 + 1
			_ = cube.SetBit(variable*2, value&1 != 0)
			_ = cube.SetBit(variable*2+1, value&2 != 0)
		}
		if cubeErr = cover.AddTail(cube); cubeErr != nil {
			return nil, cubeErr
		}
	}
	return cover, nil
}

func TestMvcUtilsTransformNative(t *testing.T) {
	data, err := os.ReadFile("test_vectors/oracle/mvc_utils_transform.json")
	if err != nil {
		t.Fatal(err)
	}
	var oracle mvcUtilsTransformOracle
	if err = json.Unmarshal(data, &oracle); err != nil || len(oracle.Cases) != 7 {
		t.Fatal(len(oracle.Cases), err)
	}
	manager := NewMvcManager()
	for test, record := range oracle.Cases {
		cover, makeErr := mvcUtilsTransformCover(manager, record.Bits)
		if makeErr != nil {
			t.Fatal(test, makeErr)
		}
		sourceDigest, _ := mvcCoverDigest(cover)
		differences, differenceErr := cover.CubePairDifferences()
		mapping := make([]int, record.Bits+3)
		for index := range mapping {
			mapping[index] = (index*7 + 3) % record.Bits
			if index%5 == 0 {
				mapping[index] = -1
			}
		}
		remapped, remapErr := cover.Remap(mapping)
		inverse, _ := cover.Duplicate()
		invertErr := inverse.Invert()
		removed, removeErr := cover.RemoveDontCareLiterals()
		cofactor, cofactorErr := cover.Cofactor(0, 1)
		flipped, flipErr := cover.FlipVariable(0, 1)
		quantified, quantifyErr := cover.UniversalQuantify(0, 1, 2, 3)
		transposed, transposeErr := cover.Transpose()
		unused, unusedErr := cover.UnusedBitsZero()
		if differenceErr != nil || remapErr != nil || invertErr != nil || removeErr != nil || cofactorErr != nil || flipErr != nil || quantifyErr != nil || transposeErr != nil || unusedErr != nil {
			t.Fatal(test, differenceErr, remapErr, invertErr, removeErr, cofactorErr, flipErr, quantifyErr, transposeErr, unusedErr)
		}
		digests := make([]string, 0, 7)
		for _, candidate := range []*MvcCover{remapped, inverse, removed, cofactor, flipped, quantified, transposed} {
			digest, digestErr := mvcCoverDigest(candidate)
			if digestErr != nil {
				t.Fatal(test, digestErr)
			}
			digests = append(digests, digest)
		}
		wantDigests := []string{record.Remap, record.Inverse, record.Removed, record.Cofactor, record.Flipped, record.Quantified, record.Transposed}
		if sourceDigest != record.Source || hex.EncodeToString(differences) != record.Diffs || !slices.Equal(digests, wantDigests) || unused != (record.Unused != 0) {
			t.Fatal(test, sourceDigest, hex.EncodeToString(differences), digests, unused, record)
		}
		for _, candidate := range []*MvcCover{transposed, quantified, flipped, cofactor, removed, inverse, remapped, cover} {
			if err = candidate.Close(); err != nil {
				t.Fatal(test, err)
			}
		}
	}
	if err = manager.Close(); err != nil {
		t.Fatal(err)
	}
}

func TestMvcUtilsTransformErrors(t *testing.T) {
	manager := NewMvcManager()
	cover, _ := NewMvcCover(manager, 4)
	_ = mvcAddBitCube(cover, 0, 2)
	if result, err := cover.Remap([]int{4}); result != nil || !errors.Is(err, ErrDimension) {
		t.Fatal(result, err)
	}
	if result, err := cover.Cofactor(-1, 0); result != nil || !errors.Is(err, ErrDimension) {
		t.Fatal(result, err)
	}
	if result, err := cover.FlipVariable(0, 2); result != nil || !errors.Is(err, ErrDimension) {
		t.Fatal(result, err)
	}
	if result, err := cover.UniversalQuantify(0, 1, 2, 4); result != nil || !errors.Is(err, ErrDimension) {
		t.Fatal(result, err)
	}
	invalid, _ := NewMvcCover(manager, 4)
	_ = mvcAddBitCube(invalid, 2)
	if result, err := invalid.FlipVariable(0, 1); result != nil || !errors.Is(err, ErrInvariant) {
		t.Fatal(result, err)
	}
	if result, err := invalid.UniversalQuantify(0, 1, 2, 3); result != nil || !errors.Is(err, ErrInvariant) {
		t.Fatal(result, err)
	}
	for _, candidate := range []*MvcCover{invalid, cover} {
		if err := candidate.Close(); err != nil {
			t.Fatal(err)
		}
	}
	if err := manager.Close(); err != nil {
		t.Fatal(err)
	}
}

func FuzzMvcUtilsTransform(f *testing.F) {
	f.Add(uint16(66), uint8(6), []byte{1, 2, 3, 4})
	f.Add(uint16(4), uint8(0), []byte{})
	f.Fuzz(func(t *testing.T, rawBits uint16, rawCount uint8, data []byte) {
		bitCount := (int(rawBits)%64 + 2) * 2
		count := int(rawCount) % 10
		wordCount := (bitCount + 31) / 32
		unused := wordCount*32 - bitCount
		values := make([][]uint32, count)
		for index := range values {
			values[index] = mvcWordsFromData(wordCount, unused, index, data, 0x47)
			for _, pair := range []int{0, 2} {
				if values[index][pair>>5]&(uint32(3)<<uint(pair&31)) == 0 {
					values[index][pair>>5] |= uint32(1) << uint(pair&31)
				}
			}
		}
		manager := NewMvcManager()
		cover, err := mvcCoverFromWords(manager, bitCount, values)
		if err != nil {
			t.Fatal(err)
		}
		differences, err := cover.CubePairDifferences()
		if err != nil {
			t.Fatal(err)
		}
		pair := 0
		for first := 0; first < count; first++ {
			for second := first + 1; second < count; second++ {
				population := 0
				for word := range values[first] {
					population += bitsOnes32(values[first][word] ^ values[second][word])
				}
				if differences[pair] != uint8(population) {
					t.Fatal(pair, differences[pair], population)
				}
				pair++
			}
		}
		mapping := []int{-1, bitCount - 1, 0, bitCount / 2, -1}
		remapped, err := cover.Remap(mapping)
		if err != nil {
			t.Fatal(err)
		}
		for cubeIndex, cube := range remapped.cubes {
			for destination, source := range mapping {
				got, _ := cube.Bit(destination)
				want := true
				if source >= 0 {
					want = values[cubeIndex][source>>5]&(uint32(1)<<uint(source&31)) != 0
				}
				if got != want {
					t.Fatal(cubeIndex, destination, got, want)
				}
			}
		}
		inverse, _ := cover.Duplicate()
		if err = inverse.Invert(); err != nil {
			t.Fatal(err)
		}
		removed, err := cover.RemoveDontCareLiterals()
		if err != nil {
			t.Fatal(err)
		}
		for cubeIndex := range values {
			inverseWords, _ := inverse.cubes[cubeIndex].Words()
			removedWords, _ := removed.cubes[cubeIndex].Words()
			for bit := 0; bit < bitCount; bit++ {
				original := values[cubeIndex][bit>>5]&(uint32(1)<<uint(bit&31)) != 0
				inverted := inverseWords[bit>>5]&(uint32(1)<<uint(bit&31)) != 0
				if inverted == original {
					t.Fatal("inverse", cubeIndex, bit)
				}
				partner := bit ^ 1
				partnerSet := values[cubeIndex][partner>>5]&(uint32(1)<<uint(partner&31)) != 0
				removedSet := removedWords[bit>>5]&(uint32(1)<<uint(bit&31)) != 0
				if removedSet != original && !partnerSet {
					t.Fatal("removed", cubeIndex, bit)
				}
				if original && partnerSet && removedSet {
					t.Fatal("dc retained", cubeIndex, bit)
				}
			}
		}
		flipped, err := cover.FlipVariable(0, 1)
		quantified, quantifyErr := cover.UniversalQuantify(0, 1, 2, 3)
		transposed, transposeErr := cover.Transpose()
		if err != nil || quantifyErr != nil || transposeErr != nil {
			t.Fatal(err, quantifyErr, transposeErr)
		}
		for cubeIndex, cube := range flipped.cubes {
			first, _ := cube.Bit(0)
			second, _ := cube.Bit(1)
			wantFirst := values[cubeIndex][0]&2 != 0
			wantSecond := values[cubeIndex][0]&1 != 0
			if first != wantFirst || second != wantSecond {
				t.Fatal("flip", cubeIndex, first, second, wantFirst, wantSecond)
			}
		}
		if transposed.BitCount() != count || transposed.CubeCount() != bitCount {
			t.Fatal(transposed.BitCount(), transposed.CubeCount(), count, bitCount)
		}
		for bit, cube := range transposed.cubes {
			for cubeIndex := 0; cubeIndex < count; cubeIndex++ {
				got, _ := cube.Bit(cubeIndex)
				want := values[cubeIndex][bit>>5]&(uint32(1)<<uint(bit&31)) != 0
				if got != want {
					t.Fatal("transpose", bit, cubeIndex, got, want)
				}
			}
		}
		unusedZero, err := cover.UnusedBitsZero()
		if err != nil || !unusedZero {
			t.Fatal(unusedZero, err)
		}
		for _, candidate := range []*MvcCover{transposed, quantified, flipped, removed, inverse, remapped, cover} {
			if err = candidate.Close(); err != nil {
				t.Fatal(err)
			}
		}
		if err = manager.Close(); err != nil {
			t.Fatal(err)
		}
	})
}

func bitsOnes32(value uint32) int {
	value -= (value >> 1) & 0x55555555
	value = (value & 0x33333333) + ((value >> 2) & 0x33333333)
	value = (value + (value >> 4)) & 0x0f0f0f0f
	return int((value * 0x01010101) >> 24)
}
