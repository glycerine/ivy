package abc

// Tests for src/misc/mvc/mvcCompare.c.

import (
	"encoding/json"
	"errors"
	"math/bits"
	"os"
	"slices"
	"testing"
)

type mvcCompareOracle struct {
	Cases []struct {
		Bits                                 int
		Int, Size, Under, Outside, Partition string
		ASizes                               []int `json:"a_sizes"`
		BSizes                               []int `json:"b_sizes"`
	} `json:"cases"`
}

func mvcComparisonCase(bitCount, which int) (a, b, mask []uint32) {
	wordCount := (bitCount + 31) / 32
	a = make([]uint32, wordCount)
	b = make([]uint32, wordCount)
	mask = make([]uint32, wordCount)
	for word := 0; word < wordCount; word++ {
		a[word] = 0x9e3779b9*uint32(word+1) + 0x01020304*uint32(which) + 0x00010001*uint32(bitCount)
		b[word] = 0x7f4a7c15*uint32(word+3) + 0x00112233*uint32(which) + 0x00010101*uint32(bitCount)
		switch which {
		case 0:
			mask[word] = 0
		case 1:
			mask[word] = ^uint32(0)
		default:
			mask[word] = 0xf0f0a5a5 ^ 0x11111111*uint32(which) ^ 0x01020408*uint32(word)
		}
	}
	if wordCount > 0 && bitCount&31 != 0 {
		tail := ^uint32(0) >> uint(32-bitCount&31)
		a[wordCount-1] &= tail
		b[wordCount-1] &= tail
		mask[wordCount-1] &= tail
	}
	if bitCount > 0 && slices.Equal(a, b) {
		bit := (which*17 + bitCount) % bitCount
		b[bit>>5] ^= uint32(1) << uint(bit&31)
	}
	return a, b, mask
}

func mvcCountWords(words []uint32) int {
	count := 0
	for _, word := range words {
		count += bits.OnesCount32(word)
	}
	return count
}

func mvcComparisonChar(result int) byte {
	if result < 0 {
		return '-'
	}
	if result > 0 {
		return '+'
	}
	return '0'
}

func TestMvcCompareNative(t *testing.T) {
	data, err := os.ReadFile("test_vectors/oracle/mvc_compare.json")
	if err != nil {
		t.Fatal(err)
	}
	var oracle mvcCompareOracle
	if err = json.Unmarshal(data, &oracle); err != nil || len(oracle.Cases) != 12 {
		t.Fatal(len(oracle.Cases), err)
	}
	manager := NewMvcManager()
	for test, record := range oracle.Cases {
		if len(record.Int) != 8 || len(record.Size) != 8 || len(record.Under) != 8 || len(record.Outside) != 8 || len(record.Partition) != 8 || len(record.ASizes) != 8 || len(record.BSizes) != 8 {
			t.Fatal(test, record)
		}
		cover, err := NewMvcCover(manager, record.Bits)
		if err != nil {
			t.Fatal(test, err)
		}
		for which := 0; which < 8; which++ {
			aWords, bWords, maskWords := mvcComparisonCase(record.Bits, which)
			a, _ := cover.NewCube()
			b, _ := cover.NewCube()
			mask, _ := cover.NewCube()
			if err = a.SetWords(aWords); err != nil {
				t.Fatal(test, which, err)
			}
			if err = b.SetWords(bWords); err != nil {
				t.Fatal(test, which, err)
			}
			if err = mask.SetWords(maskWords); err != nil {
				t.Fatal(test, which, err)
			}
			aSize, sizeErr := a.OneCount()
			bSize, otherSizeErr := b.OneCount()
			if sizeErr != nil || otherSizeErr != nil || aSize != record.ASizes[which] || bSize != record.BSizes[which] {
				t.Fatal(test, which, aSize, bSize, sizeErr, otherSizeErr)
			}
			integer, intErr := a.CompareInt(b)
			size, sizeCompareErr := a.CompareSizeAndInt(b)
			under, underErr := a.CompareIntUnderMask(b, mask)
			outside, outsideErr := a.CompareIntOutsideMask(b, mask)
			partition, partitionErr := a.CompareIntOutsideAndUnderMask(b, mask)
			if intErr != nil || sizeCompareErr != nil || underErr != nil || outsideErr != nil || partitionErr != nil {
				t.Fatal(test, which, intErr, sizeCompareErr, underErr, outsideErr, partitionErr)
			}
			got := []byte{mvcComparisonChar(integer), mvcComparisonChar(size), mvcComparisonChar(under), mvcComparisonChar(outside), mvcComparisonChar(partition)}
			want := []byte{record.Int[which], record.Size[which], record.Under[which], record.Outside[which], record.Partition[which]}
			if !slices.Equal(got, want) {
				t.Fatal(test, which, string(got), string(want))
			}
			for _, cube := range []*MvcCube{mask, b, a} {
				if err = cube.Close(); err != nil {
					t.Fatal(test, which, err)
				}
			}
		}
		if err = cover.Close(); err != nil {
			t.Fatal(test, err)
		}
	}
	if err := manager.Close(); err != nil {
		t.Fatal(err)
	}
}

func TestMvcCompareErrorsAndEquality(t *testing.T) {
	manager := NewMvcManager()
	empty, _ := NewMvcCover(manager, 0)
	emptyA, _ := empty.NewCube()
	emptyB, _ := empty.NewCube()
	emptyMask, _ := empty.NewCube()
	if result, err := emptyA.CompareInt(emptyB); err != nil || result != 0 {
		t.Fatal(result, err)
	}
	if result, err := emptyA.CompareSizeAndInt(emptyB); err != nil || result != 0 {
		t.Fatal(result, err)
	}
	if result, err := emptyA.CompareIntOutsideAndUnderMask(emptyB, emptyMask); err != nil || result != 0 {
		t.Fatal(result, err)
	}
	narrow, _ := NewMvcCover(manager, 32)
	wide, _ := NewMvcCover(manager, 33)
	a, _ := narrow.NewCube()
	b, _ := narrow.NewCube()
	mask, _ := narrow.NewCube()
	foreign, _ := wide.NewCube()
	if result, err := a.CompareInt(a); err != nil || result != 0 {
		t.Fatal(result, err)
	}
	if result, err := a.CompareIntOutsideAndUnderMask(a, mask); err != nil || result != 0 {
		t.Fatal(result, err)
	}
	if _, err := a.CompareInt(foreign); !errors.Is(err, ErrDimension) {
		t.Fatal(err)
	}
	if _, err := a.CompareIntUnderMask(b, foreign); !errors.Is(err, ErrDimension) {
		t.Fatal(err)
	}
	if _, err := (*MvcCube)(nil).CompareInt(b); !errors.Is(err, ErrDimension) {
		t.Fatal(err)
	}
	if err := b.Close(); err != nil {
		t.Fatal(err)
	}
	if _, err := a.CompareSizeAndInt(b); !errors.Is(err, ErrInvariant) {
		t.Fatal(err)
	}
	for _, cube := range []*MvcCube{foreign, mask, a, emptyMask, emptyB, emptyA} {
		if err := cube.Close(); err != nil {
			t.Fatal(err)
		}
	}
	for _, cover := range []*MvcCover{wide, narrow, empty} {
		if err := cover.Close(); err != nil {
			t.Fatal(err)
		}
	}
	if err := manager.Close(); err != nil {
		t.Fatal(err)
	}
}

func mvcModelCompare(first, second, mask []uint32, mode int) int {
	for index := len(first) - 1; index >= 0; index-- {
		left, right := first[index], second[index]
		switch mode {
		case 1:
			left &= mask[index]
			right &= mask[index]
		case 2:
			left |= mask[index]
			right |= mask[index]
		case 3:
			left &^= mask[index]
			right &^= mask[index]
		}
		if left < right {
			return -1
		}
		if left > right {
			return 1
		}
	}
	return 0
}

func FuzzMvcCompare(f *testing.F) {
	f.Add(uint16(65), []byte{1, 2, 3}, []byte{4, 5}, []byte{0xf0, 0x0f})
	f.Add(uint16(0), []byte{}, []byte{}, []byte{})
	f.Fuzz(func(t *testing.T, rawBits uint16, aData, bData, maskData []byte) {
		bitCount := int(rawBits) % 513
		manager := NewMvcManager()
		cover, err := NewMvcCover(manager, bitCount)
		if err != nil {
			t.Fatal(err)
		}
		makeWords := func(data []byte, salt byte) []uint32 {
			words := make([]uint32, cover.WordCount())
			for word := range words {
				for offset := 0; offset < 4; offset++ {
					value := salt + byte(word*29+offset*7)
					if len(data) > 0 {
						value ^= data[(word*4+offset)%len(data)]
					}
					words[word] |= uint32(value) << uint(offset*8)
				}
			}
			if len(words) > 0 && cover.UnusedBits() > 0 {
				words[len(words)-1] &= ^uint32(0) >> uint(cover.UnusedBits())
			}
			return words
		}
		aWords := makeWords(aData, 0x19)
		bWords := makeWords(bData, 0xa7)
		maskWords := makeWords(maskData, 0x5c)
		a, _ := cover.NewCube()
		b, _ := cover.NewCube()
		mask, _ := cover.NewCube()
		_ = a.SetWords(aWords)
		_ = b.SetWords(bWords)
		_ = mask.SetWords(maskWords)

		integer, err := a.CompareInt(b)
		if err != nil || integer != mvcModelCompare(aWords, bWords, maskWords, 0) {
			t.Fatal(integer, err)
		}
		under, err := a.CompareIntUnderMask(b, mask)
		if err != nil || under != mvcModelCompare(aWords, bWords, maskWords, 1) {
			t.Fatal(under, err)
		}
		outside, err := a.CompareIntOutsideMask(b, mask)
		if err != nil || outside != mvcModelCompare(aWords, bWords, maskWords, 2) {
			t.Fatal(outside, err)
		}
		partition, err := a.CompareIntOutsideAndUnderMask(b, mask)
		wantPartition := mvcModelCompare(aWords, bWords, maskWords, 3)
		if wantPartition == 0 {
			wantPartition = mvcModelCompare(aWords, bWords, maskWords, 1)
		}
		if err != nil || partition != wantPartition {
			t.Fatal(partition, wantPartition, err)
		}
		size, err := a.CompareSizeAndInt(b)
		wantSize := 0
		aCount, bCount := mvcCountWords(aWords), mvcCountWords(bWords)
		if aCount < bCount {
			wantSize = 1
		} else if aCount > bCount {
			wantSize = -1
		} else {
			wantSize = mvcModelCompare(aWords, bWords, maskWords, 0)
		}
		if err != nil || size != wantSize {
			t.Fatal(size, wantSize, err)
		}
		reverse, err := b.CompareInt(a)
		if err != nil || reverse != -integer {
			t.Fatal(reverse, integer, err)
		}
		for _, cube := range []*MvcCube{mask, b, a} {
			if err = cube.Close(); err != nil {
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
