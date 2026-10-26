package abc

// Tests for the common-cube/support slice of src/misc/mvc/mvcUtils.c.

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"slices"
	"strings"
	"testing"
)

type mvcUtilsCommonOracle struct {
	Cases []struct {
		Bits         int
		Common       string
		Intersection string
		Size         int
		Belongs      string
		FreeBefore   int    `json:"free_before"`
		FreeAfter    int    `json:"free_after"`
		CommonCover  string `json:"common_cover"`
		Normalized   string
		Contains     int
		Reverse      int
	} `json:"cases"`
}

func mvcUtilsPatternCover(manager *MvcManager, bits int) (*MvcCover, error) {
	cover, err := NewMvcCover(manager, bits)
	if err != nil {
		return nil, err
	}
	for cubeIndex := 0; cubeIndex < 4; cubeIndex++ {
		cube, cubeErr := cover.NewCube()
		if cubeErr != nil {
			return nil, cubeErr
		}
		for variable := 0; variable < bits/2; variable++ {
			value := 0
			switch variable & 3 {
			case 0:
				value = 3
			case 1:
				value = 1
				if cubeIndex == 2 {
					value = 3
				}
			case 2:
				value = 2
				if cubeIndex == 1 {
					value = 3
				}
			default:
				value = 1
				if cubeIndex == 1 {
					value = 2
				} else if cubeIndex == 2 {
					value = 3
				}
			}
			_ = cube.SetBit(variable*2, value&1 != 0)
			_ = cube.SetBit(variable*2+1, value&2 != 0)
		}
		if cubeErr = cover.AddTail(cube); cubeErr != nil {
			return nil, cubeErr
		}
	}
	return cover, nil
}

func mvcUtilsOtherCover(manager *MvcManager, bits int) (*MvcCover, error) {
	cover, err := NewMvcCover(manager, bits)
	if err != nil {
		return nil, err
	}
	cube, err := cover.NewCube()
	if err != nil {
		return nil, err
	}
	if bits > 0 {
		_ = cube.SetBit(0, true)
	}
	if err = cover.AddTail(cube); err != nil {
		return nil, err
	}
	return cover, nil
}

func mvcCubeDigest(cube *MvcCube) (string, error) {
	if err := cube.check(); err != nil {
		return "", err
	}
	hash := uint64(1469598103934665603)
	hash ^= uint64(cube.cover.bits)
	hash *= 1099511628211
	for _, word := range cube.words {
		hash ^= uint64(word)
		hash *= 1099511628211
	}
	return fmt.Sprintf("%016x", hash), nil
}

func TestMvcUtilsCommonNative(t *testing.T) {
	data, err := os.ReadFile("test_vectors/oracle/mvc_utils_common.json")
	if err != nil {
		t.Fatal(err)
	}
	var oracle mvcUtilsCommonOracle
	if err = json.Unmarshal(data, &oracle); err != nil || len(oracle.Cases) != 8 {
		t.Fatal(len(oracle.Cases), err)
	}
	manager := NewMvcManager()
	for test, record := range oracle.Cases {
		cover, makeErr := mvcUtilsPatternCover(manager, record.Bits)
		if makeErr != nil {
			t.Fatal(test, makeErr)
		}
		other, makeErr := mvcUtilsOtherCover(manager, record.Bits)
		if makeErr != nil {
			t.Fatal(test, makeErr)
		}
		normalized, _ := cover.Duplicate()
		common, _ := cover.CommonCube()
		intersection, _ := cover.SupportIntersection()
		commonDigest, _ := mvcCubeDigest(common)
		intersectionDigest, _ := mvcCubeDigest(intersection)
		size, _ := cover.BinarySupportSize()
		var belongs strings.Builder
		for variable := 0; variable < record.Bits/2; variable++ {
			has, hasErr := cover.BinarySupportHasVariable(variable)
			if hasErr != nil {
				t.Fatal(test, variable, hasErr)
			}
			if has {
				belongs.WriteByte('1')
			} else {
				belongs.WriteByte('0')
			}
		}
		freeBefore, _ := cover.IsCubeFree()
		if err = normalized.MakeCubeFree(); err != nil {
			t.Fatal(test, err)
		}
		freeAfter, _ := normalized.IsCubeFree()
		commonCover, _ := cover.CommonCubeCover()
		commonCoverDigest, _ := mvcCoverDigest(commonCover)
		normalizedDigest, _ := mvcCoverDigest(normalized)
		contains, _ := cover.SupportContains(other)
		reverse, _ := other.SupportContains(cover)
		if commonDigest != record.Common || intersectionDigest != record.Intersection || size != record.Size || belongs.String() != record.Belongs || freeBefore != (record.FreeBefore != 0) || freeAfter != (record.FreeAfter != 0) || commonCoverDigest != record.CommonCover || normalizedDigest != record.Normalized || contains != (record.Contains != 0) || reverse != (record.Reverse != 0) {
			t.Fatal(test, commonDigest, intersectionDigest, size, belongs.String(), freeBefore, freeAfter, commonCoverDigest, normalizedDigest, contains, reverse, record)
		}
		for _, cube := range []*MvcCube{intersection, common} {
			if err = cube.Close(); err != nil {
				t.Fatal(test, err)
			}
		}
		for _, candidate := range []*MvcCover{commonCover, normalized, other, cover} {
			if err = candidate.Close(); err != nil {
				t.Fatal(test, err)
			}
		}
	}
	if err = manager.Close(); err != nil {
		t.Fatal(err)
	}
}

func mvcIntersectionModel(bitCount int, all [][]uint32) []uint32 {
	words := (bitCount + 31) / 32
	result := make([]uint32, words)
	for index := range result {
		result[index] = ^uint32(0)
	}
	if words > 0 && words*32 != bitCount {
		result[words-1] &= ^uint32(0) >> uint(words*32-bitCount)
	}
	for _, cube := range all {
		for index := range result {
			result[index] &= cube[index]
		}
	}
	return result
}

func mvcUnionModel(bitCount int, all [][]uint32) []uint32 {
	result := make([]uint32, (bitCount+31)/32)
	for _, cube := range all {
		for index := range result {
			result[index] |= cube[index]
		}
	}
	return result
}

func TestMvcUtilsCommonEmptyAndErrors(t *testing.T) {
	manager := NewMvcManager()
	empty, _ := NewMvcCover(manager, 6)
	common, err := empty.CommonCube()
	if err != nil {
		t.Fatal(err)
	}
	words, _ := common.Words()
	if !slices.Equal(words, []uint32{0x3f}) {
		t.Fatal(words)
	}
	free, err := empty.IsCubeFree()
	size, sizeErr := empty.BinarySupportSize()
	has, hasErr := empty.BinarySupportHasVariable(1)
	if err != nil || sizeErr != nil || hasErr != nil || free || size != 0 || has {
		t.Fatal(free, size, has, err, sizeErr, hasErr)
	}
	if _, err = empty.BinarySupportHasVariable(-1); !errors.Is(err, ErrDimension) {
		t.Fatal(err)
	}
	if _, err = empty.BinarySupportHasVariable(3); !errors.Is(err, ErrDimension) {
		t.Fatal(err)
	}
	wide, _ := NewMvcCover(manager, 8)
	if _, err = empty.SupportContains(wide); !errors.Is(err, ErrDimension) {
		t.Fatal(err)
	}
	if err = common.Close(); err != nil {
		t.Fatal(err)
	}
	if err = empty.Close(); err != nil {
		t.Fatal(err)
	}
	if _, err = empty.CommonCube(); !errors.Is(err, ErrInvariant) {
		t.Fatal(err)
	}
	if err = wide.Close(); err != nil {
		t.Fatal(err)
	}
	if err = manager.Close(); err != nil {
		t.Fatal(err)
	}
}

func FuzzMvcUtilsCommon(f *testing.F) {
	f.Add(uint16(66), uint8(4), uint8(3), []byte{1, 2, 3}, []byte{4, 5, 6})
	f.Add(uint16(0), uint8(0), uint8(0), []byte{}, []byte{})
	f.Fuzz(func(t *testing.T, rawBits uint16, rawFirst, rawSecond uint8, firstData, secondData []byte) {
		bitCount := int(rawBits) % 258
		firstCount := int(rawFirst) % 10
		secondCount := int(rawSecond) % 10
		wordCount := (bitCount + 31) / 32
		unused := wordCount*32 - bitCount
		firstValues := make([][]uint32, firstCount)
		for index := range firstValues {
			firstValues[index] = mvcWordsFromData(wordCount, unused, index, firstData, 0x25)
		}
		secondValues := make([][]uint32, secondCount)
		for index := range secondValues {
			secondValues[index] = mvcWordsFromData(wordCount, unused, index, secondData, 0xc3)
		}
		manager := NewMvcManager()
		first, err := mvcCoverFromWords(manager, bitCount, firstValues)
		if err != nil {
			t.Fatal(err)
		}
		second, err := mvcCoverFromWords(manager, bitCount, secondValues)
		if err != nil {
			t.Fatal(err)
		}
		modelCommon := mvcIntersectionModel(bitCount, firstValues)
		common, err := first.CommonCube()
		intersection, intersectionErr := first.SupportIntersection()
		if err != nil || intersectionErr != nil {
			t.Fatal(err, intersectionErr)
		}
		commonWords, _ := common.Words()
		intersectionWords, _ := intersection.Words()
		if !slices.Equal(commonWords, modelCommon) || !slices.Equal(intersectionWords, modelCommon) {
			t.Fatal(commonWords, intersectionWords, modelCommon)
		}
		modelSize := 0
		for variable := 0; variable < bitCount/2; variable++ {
			bit := variable * 2
			want := (modelCommon[bit>>5]>>uint(bit&31))&3 != 3
			if want {
				modelSize++
			}
			got, hasErr := first.BinarySupportHasVariable(variable)
			if hasErr != nil || got != want {
				t.Fatal(variable, got, want, hasErr)
			}
		}
		size, sizeErr := first.BinarySupportSize()
		free, freeErr := first.IsCubeFree()
		modelFree := true
		for _, word := range modelCommon {
			modelFree = modelFree && word == 0
		}
		if sizeErr != nil || freeErr != nil || size != modelSize || free != modelFree {
			t.Fatal(size, modelSize, free, modelFree, sizeErr, freeErr)
		}
		commonCover, err := first.CommonCubeCover()
		if err != nil || commonCover.CubeCount() != 1 {
			t.Fatal(commonCover, err)
		}
		commonCoverWords, _ := commonCover.cubes[0].Words()
		if !slices.Equal(commonCoverWords, modelCommon) {
			t.Fatal(commonCoverWords, modelCommon)
		}
		firstUnion := mvcUnionModel(bitCount, firstValues)
		secondUnion := mvcUnionModel(bitCount, secondValues)
		wantContains := true
		for index := range firstUnion {
			wantContains = wantContains && secondUnion[index]&^firstUnion[index] == 0
		}
		contains, containsErr := first.SupportContains(second)
		if containsErr != nil || contains != wantContains {
			t.Fatal(contains, wantContains, containsErr)
		}
		if err = first.MakeCubeFree(); err != nil {
			t.Fatal(err)
		}
		for index, cube := range first.cubes {
			got, wordErr := cube.Words()
			want := slices.Clone(firstValues[index])
			for word := range want {
				want[word] &^= modelCommon[word]
			}
			if wordErr != nil || !slices.Equal(got, want) {
				t.Fatal(index, got, want, wordErr)
			}
		}
		for _, cube := range []*MvcCube{intersection, common} {
			if err = cube.Close(); err != nil {
				t.Fatal(err)
			}
		}
		for _, cover := range []*MvcCover{commonCover, second, first} {
			if err = cover.Close(); err != nil {
				t.Fatal(err)
			}
		}
		if err = manager.Close(); err != nil {
			t.Fatal(err)
		}
	})
}
