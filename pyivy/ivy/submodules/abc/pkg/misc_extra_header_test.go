package abc

// Go-only tests for the inline definitions in src/misc/extra/extra.h.

import (
	"errors"
	"slices"
	"strconv"
	"testing"
)

func extraHeaderPopulationModel(word uint32) int {
	count := 0
	for word != 0 {
		count += int(word & 1)
		word >>= 1
	}
	return count
}

func TestExtraHeaderWordCountsAndPopulation(t *testing.T) {
	// Fixed vectors pin the source reduction's boundary cases independently
	// of the exhaustive comparison below.
	for word, want := range map[uint32]int{
		0: 0, 1: 1, 0x0000000f: 4, 0x0000ffff: 16,
		0x55555555: 16, 0xffffffff: 32,
	} {
		if got := ExtraWordCountOnes32(word); got != want {
			t.Fatalf("population(%08x) = %d want %d", word, got, want)
		}
	}
	for bitCount := 0; bitCount <= 4096; bitCount++ {
		got, err := ExtraBitWordCount32(bitCount, 0)
		if err != nil || got != (bitCount+31)/32 {
			t.Fatal(bitCount, got, err)
		}
	}
	for variables := 0; variables <= 20; variables++ {
		got, err := ExtraTruthWordCount32(variables, 1<<20)
		want := 1
		if variables > 5 {
			want = 1 << uint(variables-5)
		}
		if err != nil || got != want {
			t.Fatal(variables, got, want, err)
		}
	}
	for value := 0; value < 1<<16; value++ {
		word := uint32(value) | uint32(value^0x5a5a)<<16
		if got := ExtraWordCountOnes32(word); got != extraHeaderPopulationModel(word) {
			t.Fatal(word, got)
		}
	}
	if _, err := ExtraBitWordCount32(-1, 0); !errors.Is(err, ErrDimension) {
		t.Fatal(err)
	}
	if _, err := ExtraTruthWordCount32(64, 8); !errors.Is(err, ErrLimit) {
		t.Fatal(err)
	}
	if _, err := ExtraTruthWordCount32(strconv.IntSize+4, int(^uint(0)>>1)); !errors.Is(err, ErrLimit) {
		t.Fatal("sign-bit word count", err)
	}
	if _, err := ExtraBitWordCount32(33, 1); !errors.Is(err, ErrLimit) {
		t.Fatal(err)
	}
	if _, err := ExtraTruthWordCount32(-1, 0); !errors.Is(err, ErrDimension) {
		t.Fatal(err)
	}
}

func TestExtraHeaderBitsOwnershipAndLifecycle(t *testing.T) {
	input := []uint32{1, 2}
	truth, err := NewExtraPackedTruth32(6, input, ExtraPackedTruth32Limits{})
	if err != nil {
		t.Fatal(err)
	}
	input[0] = 99
	if got, _ := truth.Words(); !slices.Equal(got, []uint32{1, 2}) {
		t.Fatal("constructor retained caller storage", got)
	}
	if err = truth.SetBit(63); err != nil {
		t.Fatal(err)
	}
	if got, _ := truth.HasBit(63); !got {
		t.Fatal("set bit")
	}
	if err = truth.XorBit(63); err != nil {
		t.Fatal(err)
	}
	if got, _ := truth.HasBit(63); got {
		t.Fatal("xor bit")
	}
	words, _ := truth.Words()
	words[0] = 99
	again, _ := truth.Words()
	if slices.Equal(words, again) || !slices.Equal(again, []uint32{1, 2}) {
		t.Fatal(words, again)
	}
	clone, err := truth.Clone()
	if err != nil {
		t.Fatal(err)
	}
	if err = clone.SetBit(31); err != nil {
		t.Fatal(err)
	}
	equal, err := truth.Equal(clone)
	if err != nil || equal {
		t.Fatal(equal, err)
	}
	if err = truth.SetBit(-1); !errors.Is(err, ErrDimension) {
		t.Fatal(err)
	}
	if err = truth.SetBit(64); !errors.Is(err, ErrDimension) {
		t.Fatal(err)
	}
	truth.Close()
	truth.Close()
	if _, err = truth.Words(); !errors.Is(err, ErrInvariant) {
		t.Fatal(err)
	}
	if _, err = (*ExtraPackedTruth32)(nil).Words(); !errors.Is(err, ErrInvariant) {
		t.Fatal(err)
	}
	if _, err = NewExtraPackedTruth32(6, []uint32{1}, ExtraPackedTruth32Limits{}); !errors.Is(err, ErrDimension) {
		t.Fatal(err)
	}
	if _, err = NewExtraPackedTruth32(6, []uint32{1, 2}, ExtraPackedTruth32Limits{MaxWords: -1}); !errors.Is(err, ErrDimension) {
		t.Fatal(err)
	}
	if _, err = NewExtraPackedTruth32(6, []uint32{1, 2}, ExtraPackedTruth32Limits{MaxWords: 1}); !errors.Is(err, ErrLimit) {
		t.Fatal(err)
	}
	if _, err = NewExtraPackedTruth32(5, []uint32{0}, ExtraPackedTruth32Limits{Stop: func() bool { return true }}); !errors.Is(err, ErrCancelled) {
		t.Fatal(err)
	}
}

func TestExtraHeaderLogicAndAtomicCancellation(t *testing.T) {
	firstWords := []uint32{0x01234567, 0x89abcdef, 0xf0f0f0f0, 0xaaaaaaaa}
	secondWords := []uint32{0x76543210, 0xfedcba98, 0x0ff00ff0, 0x55555555}
	first, _ := NewExtraPackedTruth32(7, firstWords, ExtraPackedTruth32Limits{})
	second, _ := NewExtraPackedTruth32(7, secondWords, ExtraPackedTruth32Limits{})
	operations := []struct {
		call func() (*ExtraPackedTruth32, error)
		want func(uint32, uint32) uint32
	}{
		{func() (*ExtraPackedTruth32, error) { return first.And(second) }, func(a, b uint32) uint32 { return a & b }},
		{func() (*ExtraPackedTruth32, error) { return first.Or(second) }, func(a, b uint32) uint32 { return a | b }},
		{func() (*ExtraPackedTruth32, error) { return first.Sharp(second) }, func(a, b uint32) uint32 { return a &^ b }},
		{func() (*ExtraPackedTruth32, error) { return first.Nand(second) }, func(a, b uint32) uint32 { return ^(a & b) }},
	}
	for _, operation := range operations {
		result, err := operation.call()
		if err != nil {
			t.Fatal(err)
		}
		words, _ := result.Words()
		for index := range words {
			if words[index] != operation.want(firstWords[index], secondWords[index]) {
				t.Fatal(index, words)
			}
		}
	}
	for mask := 0; mask < 4; mask++ {
		result, err := first.AndPhase(second, mask&1 != 0, mask&2 != 0)
		if err != nil {
			t.Fatal(err)
		}
		words, _ := result.Words()
		for index, got := range words {
			a, b := firstWords[index], secondWords[index]
			if mask&1 != 0 {
				a = ^a
			}
			if mask&2 != 0 {
				b = ^b
			}
			if got != a&b {
				t.Fatal(mask, index, got, a&b)
			}
		}
	}
	if implied, err := first.Sharp(second); err != nil {
		t.Fatal(err)
	} else if ok, checkErr := implied.Implies(first); checkErr != nil || !ok {
		t.Fatal(ok, checkErr)
	}
	constant, _ := NewConstantExtraPackedTruth32(7, true, ExtraPackedTruth32Limits{})
	if full, _ := constant.IsConstant(true); !full {
		t.Fatal("constant one")
	}
	if err := constant.Clear(); err != nil {
		t.Fatal(err)
	}
	if zero, _ := constant.IsConstant(false); !zero {
		t.Fatal("clear")
	}
	if err := constant.Fill(); err != nil {
		t.Fatal(err)
	}
	if err := constant.CopyFrom(first); err != nil {
		t.Fatal(err)
	}
	cancel, calls := false, 0
	cancelled, _ := NewExtraPackedTruth32(15, make([]uint32, 1<<(15-5)), ExtraPackedTruth32Limits{Stop: func() bool {
		if !cancel {
			return false
		}
		calls++
		return calls >= 2
	}})
	before, _ := cancelled.Words()
	cancel = true
	if err := cancelled.Fill(); !errors.Is(err, ErrCancelled) {
		t.Fatal(err)
	}
	cancel = false
	after, _ := cancelled.Words()
	if !slices.Equal(before, after) {
		t.Fatal("cancelled fill mutated receiver")
	}
	mismatch, _ := NewConstantExtraPackedTruth32(6, false, ExtraPackedTruth32Limits{})
	if _, err := first.And(mismatch); !errors.Is(err, ErrDimension) {
		t.Fatal(err)
	}
	mismatch.Close()
	if _, err := first.And(mismatch); !errors.Is(err, ErrInvariant) {
		t.Fatal(err)
	}
}

func TestExtraHeaderIndependentModelCampaign(t *testing.T) {
	state := uint64(0x9e3779b97f4a7c15)
	for execution := 0; execution < 5000; execution++ {
		state ^= state << 7
		state ^= state >> 9
		state ^= state << 8
		variables := 1 + int(state%10)
		count, _ := ExtraTruthWordCount32(variables, 0)
		firstWords, secondWords := make([]uint32, count), make([]uint32, count)
		for index := range firstWords {
			state = state*6364136223846793005 + 1442695040888963407
			firstWords[index] = uint32(state)
			secondWords[index] = uint32(state >> 32)
		}
		first, _ := NewExtraPackedTruth32(variables, firstWords, ExtraPackedTruth32Limits{})
		second, _ := NewExtraPackedTruth32(variables, secondWords, ExtraPackedTruth32Limits{})
		and, err := first.And(second)
		if err != nil {
			t.Fatal(execution, err)
		}
		or, _ := first.Or(second)
		not, _ := first.Not()
		andWords, _ := and.Words()
		orWords, _ := or.Words()
		notWords, _ := not.Words()
		wantCount := 0
		for index := range firstWords {
			if andWords[index] != firstWords[index]&secondWords[index] || orWords[index] != firstWords[index]|secondWords[index] || notWords[index] != ^firstWords[index] {
				t.Fatal(execution, index)
			}
			wantCount += extraHeaderPopulationModel(firstWords[index])
		}
		if got, countErr := first.CountOnes(); countErr != nil || got != wantCount {
			t.Fatal(execution, got, wantCount, countErr)
		}
	}
}

func FuzzExtraPackedTruth32(f *testing.F) {
	f.Add(uint8(6), []byte{1, 2, 3, 4, 5, 6, 7, 8})
	f.Add(uint8(3), []byte{0xff, 0, 0xaa, 0x55})
	f.Fuzz(func(t *testing.T, rawVariables uint8, data []byte) {
		variables := int(rawVariables % 13)
		count, _ := ExtraTruthWordCount32(variables, 0)
		if count > 256 {
			return
		}
		firstWords, secondWords := make([]uint32, count), make([]uint32, count)
		for index := range firstWords {
			for byteIndex := 0; byteIndex < 4; byteIndex++ {
				if len(data) != 0 {
					firstWords[index] |= uint32(data[(8*index+byteIndex)%len(data)]) << uint(8*byteIndex)
					secondWords[index] |= uint32(data[(8*index+4+byteIndex)%len(data)]) << uint(8*byteIndex)
				}
			}
		}
		first, err := NewExtraPackedTruth32(variables, firstWords, ExtraPackedTruth32Limits{})
		if err != nil {
			t.Fatal(err)
		}
		second, _ := NewExtraPackedTruth32(variables, secondWords, ExtraPackedTruth32Limits{})
		results := make([]*ExtraPackedTruth32, 0, 8)
		for _, call := range []func() (*ExtraPackedTruth32, error){
			func() (*ExtraPackedTruth32, error) { return first.And(second) },
			func() (*ExtraPackedTruth32, error) { return first.Or(second) },
			func() (*ExtraPackedTruth32, error) { return first.Sharp(second) },
			func() (*ExtraPackedTruth32, error) { return first.Nand(second) },
			first.Not,
		} {
			result, operationErr := call()
			if operationErr != nil {
				t.Fatal(operationErr)
			}
			results = append(results, result)
		}
		for index := range firstWords {
			got := make([]uint32, len(results))
			for resultIndex, result := range results {
				words, _ := result.Words()
				got[resultIndex] = words[index]
			}
			want := []uint32{
				firstWords[index] & secondWords[index],
				firstWords[index] | secondWords[index],
				firstWords[index] &^ secondWords[index],
				^(firstWords[index] & secondWords[index]),
				^firstWords[index],
			}
			if !slices.Equal(got, want) {
				t.Fatal(index, got, want)
			}
		}
		if countOnes, countErr := first.CountOnes(); countErr != nil {
			t.Fatal(countErr)
		} else {
			want := 0
			for _, word := range firstWords {
				want += extraHeaderPopulationModel(word)
			}
			if countOnes != want {
				t.Fatal(countOnes, want)
			}
		}
		if variables <= 5 && count != 1 {
			t.Fatal(variables, count)
		}
	})
}
