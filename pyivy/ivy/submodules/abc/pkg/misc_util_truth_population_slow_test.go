package abc

import (
	"errors"
	"slices"
	"testing"
)

func TestTruthPopulationSlowPackedCrossCheck(t *testing.T) {
	if result, err := (*TruthTable)(nil).CofactorPopulationsSlow(); result.Zero != nil || !errors.Is(err, ErrDimension) {
		t.Fatal(result, err)
	}
	for variables := 0; variables <= 12; variables++ {
		count, err := TruthTableWordCount(variables)
		if err != nil {
			t.Fatal(err)
		}
		words := make([]uint64, count)
		for index := range words {
			words[index] = uint64(index+1)*0x9e3779b97f4a7c15 ^ uint64(variables)*0xd6e8feb86659fd93
		}
		if variables < 6 {
			period := 1 << uint(variables)
			mask := uint64(1)<<uint(period) - 1
			base := words[0] & mask
			words[0] = 0
			for shift := 0; shift < 64; shift += period {
				words[0] |= base << uint(shift)
			}
		}
		source, err := TruthTableFromWords(variables, words, 0)
		if err != nil {
			t.Fatal(variables, err)
		}
		actual, err := source.CofactorPopulationsSlow()
		expected, expectedErr := source.CofactorPopulations()
		if err != nil || expectedErr != nil || actual.Ones != expected.Ones || !slices.Equal(actual.Zero, expected.Zero) || !slices.Equal(actual.One, expected.One) {
			t.Fatal(variables, actual, expected, err, expectedErr)
		}
		before := source.Words()
		if len(actual.Zero) != 0 {
			actual.Zero[0] = -1
			actual.One[0] = -1
			fresh, freshErr := source.CofactorPopulationsSlow()
			if freshErr != nil || fresh.Zero[0] < 0 || fresh.One[0] < 0 || !slices.Equal(source.Words(), before) {
				t.Fatal("ownership", variables, freshErr)
			}
		}
	}
}

func TestNormalizeSmallTruthWord(t *testing.T) {
	cases := []struct {
		variables  int
		word, want uint64
	}{
		{0, 0, 0},
		{0, 3, ^uint64(0)},
		{1, 2, 0xaaaaaaaaaaaaaaaa},
		{2, 0x1d, 0xdddddddddddddddd},
		{3, 0x1a5, 0xa5a5a5a5a5a5a5a5},
		{4, 0x12345, 0x2345234523452345},
		{5, 0x123456789, 0x2345678923456789},
		{6, 0x123456789abcdef0, 0x123456789abcdef0},
		{12, 0xfedcba9876543210, 0xfedcba9876543210},
	}
	for _, test := range cases {
		got, err := NormalizeSmallTruthWord(test.word, test.variables)
		if err != nil || got != test.want {
			t.Fatal(test, got, err)
		}
		again, err := NormalizeSmallTruthWord(got, test.variables)
		if err != nil || again != got {
			t.Fatal("idempotence", test, again, err)
		}
	}
	if _, err := NormalizeSmallTruthWord(0, -1); !errors.Is(err, ErrDimension) {
		t.Fatal(err)
	}
	if _, err := NormalizeSmallTruthWord(0, 37); !errors.Is(err, ErrLimit) {
		t.Fatal(err)
	}
}

func FuzzTruthPopulationSlow(f *testing.F) {
	f.Add(uint8(0), []byte{0})
	f.Add(uint8(6), []byte{0x55, 0xaa})
	f.Add(uint8(12), []byte{1, 2, 4, 8, 16, 32, 64, 128})
	f.Fuzz(func(t *testing.T, width uint8, data []byte) {
		if len(data) == 0 {
			return
		}
		variables := int(width % 13)
		count, _ := TruthTableWordCount(variables)
		words := make([]uint64, count)
		logicalBits := 1 << uint(variables)
		for bit := 0; bit < count*64; bit++ {
			logical := bit % logicalBits
			if data[(logical/8)%len(data)]&(1<<uint(logical%8)) != 0 {
				words[bit/64] |= uint64(1) << uint(bit%64)
			}
		}
		source, err := TruthTableFromWords(variables, words, 0)
		if err != nil {
			t.Fatal(err)
		}
		actual, err := source.CofactorPopulationsSlow()
		expected, expectedErr := source.CofactorPopulations()
		if err != nil || expectedErr != nil || actual.Ones != expected.Ones || !slices.Equal(actual.Zero, expected.Zero) || !slices.Equal(actual.One, expected.One) {
			t.Fatal(variables, actual, expected, err, expectedErr)
		}
	})
}
