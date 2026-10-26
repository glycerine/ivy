package abc

// Tests for src/misc/mvc/mvcOpAlg.c and Mvc_CoverSupport in mvcUtils.c.

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"slices"
	"testing"
)

type mvcAlgebraicOracle struct {
	Cases []struct {
		Bits            int
		Disjoint        int
		OverlapDisjoint int `json:"overlap_disjoint"`
		ProductCount    int `json:"product_count"`
		Product         string
		DifferenceCount int `json:"difference_count"`
		Difference      string
		Forward         int
		Reverse         int
		Duplicates      int
	} `json:"cases"`
}

func mvcAlgebraicPatternCover(manager *MvcManager, bitCount, count int, odd bool) (*MvcCover, []*MvcCube, error) {
	cover, err := NewMvcCover(manager, bitCount)
	if err != nil {
		return nil, nil, err
	}
	cubes := make([]*MvcCube, count)
	for cubeIndex := range cubes {
		cubes[cubeIndex], err = cover.NewCube()
		if err != nil {
			return nil, nil, err
		}
		for bit := 0; bit < bitCount; bit++ {
			include := bit%2 == 0 && (bit/2+cubeIndex)%3 != 0
			if odd {
				include = bit%2 == 1 && (bit/2+cubeIndex)%2 == 0
			}
			if include {
				_ = cubes[cubeIndex].SetBit(bit, true)
			}
		}
		if err = cover.AddTail(cubes[cubeIndex]); err != nil {
			return nil, nil, err
		}
	}
	return cover, cubes, nil
}

func mvcCoverDigest(cover *MvcCover) (string, error) {
	hash := uint64(1469598103934665603)
	hash ^= uint64(cover.BitCount())
	hash *= 1099511628211
	hash ^= uint64(cover.CubeCount())
	hash *= 1099511628211
	cubes, err := cover.Cubes()
	if err != nil {
		return "", err
	}
	for _, cube := range cubes {
		words, wordErr := cube.Words()
		if wordErr != nil {
			return "", wordErr
		}
		for _, word := range words {
			hash ^= uint64(word)
			hash *= 1099511628211
		}
		hash ^= uint64(0xffffffff)
		hash *= 1099511628211
	}
	return fmt.Sprintf("%016x", hash), nil
}

func TestMvcAlgebraicNative(t *testing.T) {
	data, err := os.ReadFile("test_vectors/oracle/mvc_algebraic.json")
	if err != nil {
		t.Fatal(err)
	}
	var oracle mvcAlgebraicOracle
	if err = json.Unmarshal(data, &oracle); err != nil || len(oracle.Cases) != 9 {
		t.Fatal(len(oracle.Cases), err)
	}
	manager := NewMvcManager()
	for test, record := range oracle.Cases {
		a, aCubes, err := mvcAlgebraicPatternCover(manager, record.Bits, 4, false)
		if err != nil {
			t.Fatal(test, err)
		}
		b, bCubes, err := mvcAlgebraicPatternCover(manager, record.Bits, 3, true)
		if err != nil {
			t.Fatal(test, err)
		}
		overlap, _ := a.Clone()
		for _, cube := range bCubes {
			if _, err = overlap.AddCloneTail(cube); err != nil {
				t.Fatal(test, err)
			}
		}
		_, _ = overlap.AddCloneTail(aCubes[0])
		remove, _ := a.Clone()
		_, _ = remove.AddCloneTail(aCubes[1])
		_, _ = remove.AddCloneTail(aCubes[3])
		superset, _ := a.Duplicate()
		extra, _ := superset.NewCube()
		for bit := 0; bit < record.Bits; bit++ {
			_ = extra.SetBit(bit, true)
		}
		_ = superset.AddTail(extra)
		duplicates, _ := a.Clone()
		_, _ = duplicates.AddCloneTail(aCubes[0])
		_, _ = duplicates.AddCloneTail(aCubes[0])
		singleton, _ := a.Clone()
		_, _ = singleton.AddCloneTail(aCubes[0])

		disjoint, disjointErr := a.SupportsDisjoint(b)
		overlapDisjoint, overlapErr := a.SupportsDisjoint(overlap)
		if disjointErr != nil || overlapErr != nil || disjoint != (record.Disjoint != 0) || overlapDisjoint != (record.OverlapDisjoint != 0) {
			t.Fatal(test, disjoint, overlapDisjoint, disjointErr, overlapErr)
		}
		product, err := a.AlgebraicMultiply(b)
		if err != nil {
			t.Fatal(test, err)
		}
		productDigest, err := mvcCoverDigest(product)
		if err != nil || product.CubeCount() != record.ProductCount || productDigest != record.Product {
			t.Fatal(test, product.CubeCount(), productDigest, record.Product, err)
		}
		difference, err := a.AlgebraicSubtract(remove)
		if err != nil {
			t.Fatal(test, err)
		}
		differenceDigest, err := mvcCoverDigest(difference)
		if err != nil || difference.CubeCount() != record.DifferenceCount || differenceDigest != record.Difference {
			t.Fatal(test, difference.CubeCount(), differenceDigest, record.Difference, err)
		}
		forward, forwardErr := a.AlgebraicEqual(superset)
		reverse, reverseErr := superset.AlgebraicEqual(a)
		duplicateEqual, duplicateErr := duplicates.AlgebraicEqual(singleton)
		if forwardErr != nil || reverseErr != nil || duplicateErr != nil || forward != (record.Forward != 0) || reverse != (record.Reverse != 0) || duplicateEqual != (record.Duplicates != 0) {
			t.Fatal(test, forward, reverse, duplicateEqual, forwardErr, reverseErr, duplicateErr)
		}
		support, err := a.Support()
		if err != nil {
			t.Fatal(test, err)
		}
		for bit := 1; bit < record.Bits; bit += 2 {
			value, _ := support.Bit(bit)
			if value {
				t.Fatal(test, "unexpected odd support", bit)
			}
		}
		if err = support.Close(); err != nil {
			t.Fatal(test, err)
		}
		for _, cover := range []*MvcCover{difference, product, singleton, duplicates, superset, remove, overlap, b, a} {
			if err = cover.Close(); err != nil {
				t.Fatal(test, err)
			}
		}
	}
	if err := manager.Close(); err != nil {
		t.Fatal(err)
	}
}

func TestMvcAlgebraicErrorsAndEmpty(t *testing.T) {
	manager := NewMvcManager()
	empty, _ := NewMvcCover(manager, 3)
	one, _ := NewMvcCover(manager, 3)
	cube, _ := one.NewCube()
	_ = cube.SetBit(0, true)
	_ = one.AddTail(cube)
	product, err := empty.AlgebraicMultiply(one)
	if err != nil || product.CubeCount() != 0 {
		t.Fatal(product, err)
	}
	forward, err := empty.AlgebraicEqual(one)
	reverse, reverseErr := one.AlgebraicEqual(empty)
	if err != nil || reverseErr != nil || !forward || reverse {
		t.Fatal(forward, reverse, err, reverseErr)
	}
	wide, _ := NewMvcCover(manager, 4)
	if result, err := one.AlgebraicMultiply(wide); result != nil || !errors.Is(err, ErrDimension) {
		t.Fatal(result, err)
	}
	if result, err := one.AlgebraicSubtract(wide); result != nil || !errors.Is(err, ErrDimension) {
		t.Fatal(result, err)
	}
	if _, err := one.AlgebraicEqual(wide); !errors.Is(err, ErrDimension) {
		t.Fatal(err)
	}
	if err := one.Close(); err != nil {
		t.Fatal(err)
	}
	if _, err := one.Support(); !errors.Is(err, ErrInvariant) {
		t.Fatal(err)
	}
	for _, cover := range []*MvcCover{wide, product, empty} {
		if err := cover.Close(); err != nil {
			t.Fatal(err)
		}
	}
	if err := manager.Close(); err != nil {
		t.Fatal(err)
	}
}

func mvcWordsFromData(wordCount, unused, cubeIndex int, data []byte, salt byte) []uint32 {
	words := make([]uint32, wordCount)
	for word := range words {
		for offset := 0; offset < 4; offset++ {
			value := salt + byte(cubeIndex*29+word*11+offset*5)
			if len(data) > 0 {
				value ^= data[(cubeIndex+word*4+offset)%len(data)]
			}
			words[word] |= uint32(value) << uint(offset*8)
		}
	}
	if len(words) > 0 && unused > 0 {
		words[len(words)-1] &= ^uint32(0) >> uint(unused)
	}
	return words
}

func mvcCoverFromWords(manager *MvcManager, bitCount int, all [][]uint32) (*MvcCover, error) {
	cover, err := NewMvcCover(manager, bitCount)
	if err != nil {
		return nil, err
	}
	for _, words := range all {
		cube, cubeErr := cover.NewCube()
		if cubeErr != nil {
			return nil, cubeErr
		}
		if cubeErr = cube.SetWords(words); cubeErr != nil {
			return nil, cubeErr
		}
		if cubeErr = cover.AddTail(cube); cubeErr != nil {
			return nil, cubeErr
		}
	}
	return cover, nil
}

func FuzzMvcAlgebraic(f *testing.F) {
	f.Add(uint16(65), uint8(4), uint8(3), []byte{1, 2, 3}, []byte{4, 5, 6})
	f.Add(uint16(0), uint8(0), uint8(2), []byte{}, []byte{})
	f.Fuzz(func(t *testing.T, rawBits uint16, rawLeft, rawRight uint8, leftData, rightData []byte) {
		bitCount := int(rawBits) % 130
		leftCount := int(rawLeft) % 9
		rightCount := int(rawRight) % 9
		wordCount := (bitCount + 31) / 32
		unused := wordCount*32 - bitCount
		leftWords := make([][]uint32, leftCount)
		rightWords := make([][]uint32, rightCount)
		for index := range leftWords {
			leftWords[index] = mvcWordsFromData(wordCount, unused, index, leftData, 0x19)
		}
		for index := range rightWords {
			rightWords[index] = mvcWordsFromData(wordCount, unused, index, rightData, 0xa3)
		}
		if leftCount > 0 && rightCount > 0 && len(rightData) > 0 && rightData[0]&1 != 0 {
			rightWords[0] = slices.Clone(leftWords[int(rightData[0])%leftCount])
		}
		manager := NewMvcManager()
		left, err := mvcCoverFromWords(manager, bitCount, leftWords)
		if err != nil {
			t.Fatal(err)
		}
		right, err := mvcCoverFromWords(manager, bitCount, rightWords)
		if err != nil {
			t.Fatal(err)
		}
		product, err := left.AlgebraicMultiply(right)
		if err != nil || product.CubeCount() != leftCount*rightCount {
			t.Fatal(product.CubeCount(), leftCount*rightCount, err)
		}
		productCubes, _ := product.Cubes()
		position := 0
		for _, a := range leftWords {
			for _, b := range rightWords {
				got, _ := productCubes[position].Words()
				want := make([]uint32, wordCount)
				for word := range want {
					want[word] = a[word] | b[word]
				}
				if !slices.Equal(got, want) {
					t.Fatal(position)
				}
				position++
			}
		}
		difference, err := left.AlgebraicSubtract(right)
		if err != nil {
			t.Fatal(err)
		}
		differenceCubes, _ := difference.Cubes()
		wantDifference := make([][]uint32, 0, leftCount)
		for _, a := range leftWords {
			found := false
			for _, b := range rightWords {
				found = found || slices.Equal(a, b)
			}
			if !found {
				wantDifference = append(wantDifference, a)
			}
		}
		if len(differenceCubes) != len(wantDifference) {
			t.Fatal(len(differenceCubes), len(wantDifference))
		}
		for index, cube := range differenceCubes {
			got, _ := cube.Words()
			if !slices.Equal(got, wantDifference[index]) {
				t.Fatal(index)
			}
		}
		included, err := left.AlgebraicEqual(right)
		wantIncluded := true
		for _, a := range leftWords {
			found := false
			for _, b := range rightWords {
				found = found || slices.Equal(a, b)
			}
			wantIncluded = wantIncluded && found
		}
		if err != nil || included != wantIncluded {
			t.Fatal(included, wantIncluded, err)
		}
		leftSupport, _ := left.Support()
		rightSupport, _ := right.Support()
		wantDisjoint := true
		for word := range leftSupport.words {
			if leftSupport.words[word]&rightSupport.words[word] != 0 {
				wantDisjoint = false
			}
		}
		disjoint, err := left.SupportsDisjoint(right)
		if err != nil || disjoint != wantDisjoint {
			t.Fatal(disjoint, wantDisjoint, err)
		}
		_ = rightSupport.Close()
		_ = leftSupport.Close()
		for _, cover := range []*MvcCover{difference, product, right, left} {
			if err = cover.Close(); err != nil {
				t.Fatal(err)
			}
		}
		if err = manager.Close(); err != nil {
			t.Fatal(err)
		}
	})
}
