package abc

// Tests for the active surface of src/misc/mvc/mvcOpBool.c.

import (
	"encoding/json"
	"errors"
	"os"
	"slices"
	"testing"
)

type mvcBooleanOracle struct {
	Cases []struct {
		Bits  int
		Count int
		Union string
		Left  string
		Right string
	} `json:"cases"`
}

func TestMvcBooleanOrNative(t *testing.T) {
	data, err := os.ReadFile("test_vectors/oracle/mvc_boolean.json")
	if err != nil {
		t.Fatal(err)
	}
	var oracle mvcBooleanOracle
	if err = json.Unmarshal(data, &oracle); err != nil || len(oracle.Cases) != 9 {
		t.Fatal(len(oracle.Cases), err)
	}
	manager := NewMvcManager()
	for test, record := range oracle.Cases {
		left, leftCubes, makeErr := mvcAlgebraicPatternCover(manager, record.Bits, 4, false)
		if makeErr != nil {
			t.Fatal(test, makeErr)
		}
		right, rightCubes, makeErr := mvcAlgebraicPatternCover(manager, record.Bits, 3, true)
		if makeErr != nil {
			t.Fatal(test, makeErr)
		}
		union, unionErr := left.BooleanOr(right)
		if unionErr != nil {
			t.Fatal(test, unionErr)
		}
		unionDigest, _ := mvcCoverDigest(union)
		leftDigest, _ := mvcCoverDigest(left)
		rightDigest, _ := mvcCoverDigest(right)
		if union.CubeCount() != record.Count || unionDigest != record.Union || leftDigest != record.Left || rightDigest != record.Right {
			t.Fatal(test, union.CubeCount(), unionDigest, leftDigest, rightDigest, record)
		}
		if len(leftCubes) > 0 && record.Bits > 0 {
			before, _ := union.cubes[0].Bit(0)
			_ = leftCubes[0].SetBit(0, !before)
			after, _ := union.cubes[0].Bit(0)
			if after != before {
				t.Fatal(test, "result aliases left source")
			}
		}
		if len(rightCubes) > 0 && record.Bits > 0 {
			index := len(leftCubes)
			before, _ := rightCubes[0].Bit(0)
			resultBefore, _ := union.cubes[index].Bit(0)
			_ = union.cubes[index].SetBit(0, !resultBefore)
			after, _ := rightCubes[0].Bit(0)
			if after != before {
				t.Fatal(test, "right source aliases result")
			}
		}
		for _, cover := range []*MvcCover{union, right, left} {
			if err = cover.Close(); err != nil {
				t.Fatal(test, err)
			}
		}
	}
	if err = manager.Close(); err != nil {
		t.Fatal(err)
	}
}

func TestMvcBooleanOrErrorsAndEmpty(t *testing.T) {
	manager := NewMvcManager()
	empty, _ := NewMvcCover(manager, 0)
	other, _ := NewMvcCover(manager, 0)
	result, err := empty.BooleanOr(other)
	if err != nil || result.CubeCount() != 0 {
		t.Fatal(result, err)
	}
	wide, _ := NewMvcCover(manager, 1)
	if got, gotErr := empty.BooleanOr(wide); got != nil || !errors.Is(gotErr, ErrDimension) {
		t.Fatal(got, gotErr)
	}
	if err = other.Close(); err != nil {
		t.Fatal(err)
	}
	if got, gotErr := empty.BooleanOr(other); got != nil || !errors.Is(gotErr, ErrInvariant) {
		t.Fatal(got, gotErr)
	}
	for _, cover := range []*MvcCover{wide, result, empty} {
		if err = cover.Close(); err != nil {
			t.Fatal(err)
		}
	}
	if err = manager.Close(); err != nil {
		t.Fatal(err)
	}
}

func FuzzMvcBooleanOr(f *testing.F) {
	f.Add(uint16(65), uint8(4), uint8(3), []byte{1, 2, 3}, []byte{4, 5, 6})
	f.Add(uint16(0), uint8(0), uint8(2), []byte{}, []byte{})
	f.Fuzz(func(t *testing.T, rawBits uint16, rawLeft, rawRight uint8, leftData, rightData []byte) {
		bits := int(rawBits) % 258
		leftCount := int(rawLeft) % 12
		rightCount := int(rawRight) % 12
		words := (bits + 31) / 32
		unused := words*32 - bits
		leftValues := make([][]uint32, leftCount)
		for index := range leftValues {
			leftValues[index] = mvcWordsFromData(words, unused, index, leftData, 0x31)
		}
		rightValues := make([][]uint32, rightCount)
		for index := range rightValues {
			rightValues[index] = mvcWordsFromData(words, unused, index, rightData, 0xa7)
		}
		manager := NewMvcManager()
		left, err := mvcCoverFromWords(manager, bits, leftValues)
		if err != nil {
			t.Fatal(err)
		}
		right, err := mvcCoverFromWords(manager, bits, rightValues)
		if err != nil {
			t.Fatal(err)
		}
		result, err := left.BooleanOr(right)
		if err != nil || result.CubeCount() != leftCount+rightCount {
			t.Fatal(result, err)
		}
		want := append(slices.Clone(leftValues), rightValues...)
		for index, cube := range result.cubes {
			got, wordErr := cube.Words()
			if wordErr != nil || !slices.Equal(got, want[index]) {
				t.Fatal(index, got, want[index], wordErr)
			}
		}
		if len(result.cubes) > 0 && bits > 0 {
			before, _ := result.cubes[0].Bit(0)
			if len(left.cubes) > 0 {
				_ = left.cubes[0].SetBit(0, !before)
			} else {
				_ = right.cubes[0].SetBit(0, !before)
			}
			after, _ := result.cubes[0].Bit(0)
			if after != before {
				t.Fatal("result aliases source")
			}
		}
		for _, cover := range []*MvcCover{result, right, left} {
			if err = cover.Close(); err != nil {
				t.Fatal(err)
			}
		}
		if err = manager.Close(); err != nil {
			t.Fatal(err)
		}
	})
}
