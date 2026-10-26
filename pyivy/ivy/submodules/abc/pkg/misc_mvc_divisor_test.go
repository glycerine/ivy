package abc

// Tests for src/misc/mvc/mvcDivisor.c.

import (
	"encoding/json"
	"errors"
	"os"
	"slices"
	"testing"
)

type mvcDivisorOracle struct {
	Cases []struct {
		Bits   int
		Source string
		After  string
		Count  int
		Result string
	} `json:"cases"`
	Trivial []int `json:"trivial"`
}

func mvcDivisorProductCover(manager *MvcManager, bits int) (*MvcCover, error) {
	cover, err := NewMvcCover(manager, bits)
	if err != nil {
		return nil, err
	}
	for _, cube := range [][]int{{0, 1, 4}, {0, 1, 5}, {0, 2, 4}, {0, 2, 5}, {3, 1, 4}, {3, 1, 5}, {3, 2, 4}, {3, 2, 5}} {
		if err = mvcAddBitCube(cover, cube...); err != nil {
			return nil, err
		}
	}
	return cover, nil
}

func TestMvcDivisorNative(t *testing.T) {
	data, err := os.ReadFile("test_vectors/oracle/mvc_divisor.json")
	if err != nil {
		t.Fatal(err)
	}
	var oracle mvcDivisorOracle
	if err = json.Unmarshal(data, &oracle); err != nil || len(oracle.Cases) != 7 || !slices.Equal(oracle.Trivial, []int{1, 1}) {
		t.Fatal(len(oracle.Cases), oracle.Trivial, err)
	}
	manager := NewMvcManager()
	for test, record := range oracle.Cases {
		cover, makeErr := mvcDivisorProductCover(manager, record.Bits)
		if makeErr != nil {
			t.Fatal(test, makeErr)
		}
		before, _ := mvcCoverDigest(cover)
		result, divisorErr := cover.Divisor()
		after, _ := mvcCoverDigest(cover)
		if divisorErr != nil || result == nil {
			t.Fatal(test, result, divisorErr)
		}
		resultDigest, _ := mvcCoverDigest(result)
		if before != record.Source || after != record.After || result.CubeCount() != record.Count || resultDigest != record.Result {
			t.Fatal(test, before, after, result.CubeCount(), resultDigest, record)
		}
		if err = result.Close(); err != nil {
			t.Fatal(test, err)
		}
		if err = cover.Close(); err != nil {
			t.Fatal(test, err)
		}
	}
	if err = manager.Close(); err != nil {
		t.Fatal(err)
	}
}

func mvcDivisorModel(bitCount int, input [][]uint32) [][]uint32 {
	if len(input) <= 1 {
		return nil
	}
	values := make([][]uint32, len(input))
	for index := range input {
		values[index] = slices.Clone(input[index])
	}
	hasRepeated := false
	for bit := 0; bit < bitCount; bit++ {
		count := 0
		for _, cube := range values {
			if cube[bit>>5]&(uint32(1)<<uint(bit&31)) != 0 {
				count++
			}
		}
		hasRepeated = hasRepeated || count > 1
	}
	if !hasRepeated {
		return nil
	}
	for {
		literal, worst := -1, int(^uint(0)>>1)
		for bit := 0; bit < bitCount; bit++ {
			count := 0
			for _, cube := range values {
				if cube[bit>>5]&(uint32(1)<<uint(bit&31)) != 0 {
					count++
				}
			}
			if count >= 2 && count < worst {
				literal, worst = bit, count
			}
		}
		if literal < 0 {
			return values
		}
		filtered := make([][]uint32, 0, len(values))
		mask := uint32(1) << uint(literal&31)
		for _, cube := range values {
			if cube[literal>>5]&mask != 0 {
				copyCube := slices.Clone(cube)
				copyCube[literal>>5] &^= mask
				filtered = append(filtered, copyCube)
			}
		}
		values = filtered
		common := mvcIntersectionModel(bitCount, values)
		for _, cube := range values {
			for word := range cube {
				cube[word] &^= common[word]
			}
		}
	}
}

func TestMvcDivisorErrorsAndTrivial(t *testing.T) {
	manager := NewMvcManager()
	empty, _ := NewMvcCover(manager, 4)
	if result, err := empty.Divisor(); result != nil || err != nil {
		t.Fatal(result, err)
	}
	_ = mvcAddBitCube(empty, 0)
	if result, err := empty.Divisor(); result != nil || err != nil {
		t.Fatal(result, err)
	}
	if err := empty.Close(); err != nil {
		t.Fatal(err)
	}
	if result, err := empty.Divisor(); result != nil || !errors.Is(err, ErrInvariant) {
		t.Fatal(result, err)
	}
	if err := manager.Close(); err != nil {
		t.Fatal(err)
	}
}

func FuzzMvcDivisor(f *testing.F) {
	f.Add(uint16(66), uint8(7), []byte{1, 2, 3, 4})
	f.Add(uint16(0), uint8(0), []byte{})
	f.Fuzz(func(t *testing.T, rawBits uint16, rawCount uint8, data []byte) {
		bitCount := int(rawBits) % 130
		count := int(rawCount) % 12
		wordCount := (bitCount + 31) / 32
		unused := wordCount*32 - bitCount
		values := make([][]uint32, count)
		for index := range values {
			values[index] = mvcWordsFromData(wordCount, unused, index, data, 0x9b)
		}
		model := mvcDivisorModel(bitCount, values)
		manager := NewMvcManager()
		cover, err := mvcCoverFromWords(manager, bitCount, values)
		if err != nil {
			t.Fatal(err)
		}
		before, _ := mvcCoverDigest(cover)
		result, err := cover.Divisor()
		if err != nil {
			t.Fatal(err)
		}
		after, _ := mvcCoverDigest(cover)
		if before != after {
			t.Fatal("source mutated", before, after)
		}
		if model == nil {
			if result != nil {
				t.Fatal("unexpected divisor", result.CubeCount())
			}
		} else {
			if result == nil || result.CubeCount() != len(model) {
				t.Fatal(result, len(model))
			}
			for index, cube := range result.cubes {
				got, wordErr := cube.Words()
				if wordErr != nil || !slices.Equal(got, model[index]) {
					t.Fatal(index, got, model[index], wordErr)
				}
			}
			if err = result.Close(); err != nil {
				t.Fatal(err)
			}
		}
		if err = cover.Close(); err != nil {
			t.Fatal(err)
		}
		if err = manager.Close(); err != nil {
			t.Fatal(err)
		}
	})
}
