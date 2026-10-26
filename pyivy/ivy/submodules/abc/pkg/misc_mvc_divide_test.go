package abc

// Tests for src/misc/mvc/mvcDivide.c.

import (
	"encoding/json"
	"errors"
	"os"
	"slices"
	"testing"
)

type mvcDivideOracle struct {
	Cases []struct {
		Bits             int
		Cover            string
		Divisor          string
		Quotient         string
		Remainder        string
		LiteralQuotient  string `json:"literal_quotient"`
		LiteralRemainder string `json:"literal_remainder"`
		CubeQuotient     string `json:"cube_quotient"`
		CubeRemainder    string `json:"cube_remainder"`
		InPlace          string `json:"in_place"`
		CountFail        int    `json:"count_fail"`
		SupportFail      int    `json:"support_fail"`
	} `json:"cases"`
}

func mvcAddBitCube(cover *MvcCover, bits ...int) error {
	cube, err := cover.NewCube()
	if err != nil {
		return err
	}
	for _, bit := range bits {
		if bit >= 0 {
			if err = cube.SetBit(bit, true); err != nil {
				return err
			}
		}
	}
	return cover.AddTail(cube)
}

func mvcNativeGeneralDivisor(manager *MvcManager, bits int) (*MvcCover, error) {
	cover, err := NewMvcCover(manager, bits)
	if err != nil {
		return nil, err
	}
	for _, bit := range []int{2, 0, 1} {
		if err = mvcAddBitCube(cover, bit); err != nil {
			return nil, err
		}
	}
	return cover, nil
}

func mvcNativeGeneralCover(manager *MvcManager, bits int) (*MvcCover, error) {
	cover, err := NewMvcCover(manager, bits)
	if err != nil {
		return nil, err
	}
	q0, q1 := bits-2, bits-1
	for _, cube := range [][]int{{q0, 1}, {q1, 2}, {q0, q1, 0}, {q0, 0}, {q1, 0}, {q0, q1, 2}, {q0, 2}, {q0, q1, 1}} {
		if err = mvcAddBitCube(cover, cube...); err != nil {
			return nil, err
		}
	}
	return cover, nil
}

func mvcNativeLiteralSource(manager *MvcManager, bits, literal int) (*MvcCover, error) {
	cover, err := NewMvcCover(manager, bits)
	if err != nil {
		return nil, err
	}
	for _, cube := range [][]int{{literal, 0}, {1}, {literal, 2}, {3}, {literal, 0, 3}} {
		if err = mvcAddBitCube(cover, cube...); err != nil {
			return nil, err
		}
	}
	return cover, nil
}

func mvcNativeCubeSource(manager *MvcManager, bits int) (*MvcCover, error) {
	cover, err := NewMvcCover(manager, bits)
	if err != nil {
		return nil, err
	}
	for _, cube := range [][]int{{0, 33, bits - 1}, {0, bits - 1}, {33, bits - 2}, {0, 33}, {0, 33, bits - 2}} {
		if err = mvcAddBitCube(cover, cube...); err != nil {
			return nil, err
		}
	}
	return cover, nil
}

func TestMvcDivideNative(t *testing.T) {
	data, err := os.ReadFile("test_vectors/oracle/mvc_divide.json")
	if err != nil {
		t.Fatal(err)
	}
	var oracle mvcDivideOracle
	if err = json.Unmarshal(data, &oracle); err != nil || len(oracle.Cases) != 4 {
		t.Fatal(len(oracle.Cases), err)
	}
	manager := NewMvcManager()
	for test, record := range oracle.Cases {
		cover, _ := mvcNativeGeneralCover(manager, record.Bits)
		divisor, _ := mvcNativeGeneralDivisor(manager, record.Bits)
		quotient, remainder, divideErr := cover.Divide(divisor)
		if divideErr != nil || quotient == nil || remainder == nil {
			t.Fatal(test, quotient, remainder, divideErr)
		}
		literal := record.Bits - 1
		literalSource, _ := mvcNativeLiteralSource(manager, record.Bits, literal)
		literalDivisor, _ := NewMvcCover(manager, record.Bits)
		_ = mvcAddBitCube(literalDivisor, literal)
		literalQuotient, literalRemainder, divideErr := literalSource.DivideByLiteral(literalDivisor)
		if divideErr != nil {
			t.Fatal(test, divideErr)
		}
		cubeSource, _ := mvcNativeCubeSource(manager, record.Bits)
		cubeDivisor, _ := NewMvcCover(manager, record.Bits)
		_ = mvcAddBitCube(cubeDivisor, 0, 33)
		cubeQuotient, cubeRemainder, divideErr := cubeSource.DivideByCube(cubeDivisor)
		if divideErr != nil {
			t.Fatal(test, divideErr)
		}
		inPlace, _ := mvcNativeLiteralSource(manager, record.Bits, literal)
		if err = inPlace.DivideByLiteralQuotient(literal); err != nil {
			t.Fatal(test, err)
		}
		tooSmall, _ := NewMvcCover(manager, record.Bits)
		_ = mvcAddBitCube(tooSmall, 0)
		countQ, countR, countErr := tooSmall.Divide(divisor)
		supportSource, _ := NewMvcCover(manager, record.Bits)
		_ = mvcAddBitCube(supportSource, 0)
		supportDivisor, _ := NewMvcCover(manager, record.Bits)
		_ = mvcAddBitCube(supportDivisor, 1)
		supportQ, supportR, supportErr := supportSource.Divide(supportDivisor)
		digests := make([]string, 0, 9)
		for _, candidate := range []*MvcCover{cover, divisor, quotient, remainder, literalQuotient, literalRemainder, cubeQuotient, cubeRemainder, inPlace} {
			digest, digestErr := mvcCoverDigest(candidate)
			if digestErr != nil {
				t.Fatal(test, digestErr)
			}
			digests = append(digests, digest)
		}
		want := []string{record.Cover, record.Divisor, record.Quotient, record.Remainder, record.LiteralQuotient, record.LiteralRemainder, record.CubeQuotient, record.CubeRemainder, record.InPlace}
		if !slices.Equal(digests, want) || countErr != nil || supportErr != nil || countQ != nil || countR != nil || supportQ != nil || supportR != nil || record.CountFail != 1 || record.SupportFail != 1 {
			t.Fatal(test, digests, want, countQ, countR, countErr, supportQ, supportR, supportErr)
		}
		for _, candidate := range []*MvcCover{supportDivisor, supportSource, tooSmall, inPlace, cubeRemainder, cubeQuotient, cubeDivisor, cubeSource, literalRemainder, literalQuotient, literalDivisor, literalSource, remainder, quotient, divisor, cover} {
			if err = candidate.Close(); err != nil {
				t.Fatal(test, err)
			}
		}
	}
	if err = manager.Close(); err != nil {
		t.Fatal(err)
	}
}

func mvcDivisionWords(bitCount, cubeIndex int, data []byte, salt byte) []uint32 {
	words := (bitCount + 31) / 32
	return mvcWordsFromData(words, words*32-bitCount, cubeIndex, data, salt)
}

func TestMvcDivideErrors(t *testing.T) {
	manager := NewMvcManager()
	cover, _ := NewMvcCover(manager, 8)
	empty, _ := NewMvcCover(manager, 8)
	if quotient, remainder, err := cover.Divide(empty); quotient != nil || remainder != nil || !errors.Is(err, ErrInvariant) {
		t.Fatal(quotient, remainder, err)
	}
	zero, _ := NewMvcCover(manager, 8)
	zeroCube, _ := zero.NewCube()
	_ = zero.AddTail(zeroCube)
	if quotient, remainder, err := cover.DivideByLiteral(zero); quotient != nil || remainder != nil || !errors.Is(err, ErrInvariant) {
		t.Fatal(quotient, remainder, err)
	}
	wide, _ := NewMvcCover(manager, 9)
	if quotient, remainder, err := cover.Divide(wide); quotient != nil || remainder != nil || !errors.Is(err, ErrDimension) {
		t.Fatal(quotient, remainder, err)
	}
	if err := cover.DivideByLiteralQuotient(-1); !errors.Is(err, ErrDimension) {
		t.Fatal(err)
	}
	if err := cover.DivideByLiteralQuotient(8); !errors.Is(err, ErrDimension) {
		t.Fatal(err)
	}
	for _, candidate := range []*MvcCover{wide, zero, empty, cover} {
		if err := candidate.Close(); err != nil {
			t.Fatal(err)
		}
	}
	if err := manager.Close(); err != nil {
		t.Fatal(err)
	}
}

func FuzzMvcDivideSingleCube(f *testing.F) {
	f.Add(uint16(66), uint8(5), uint8(17), []byte{1, 2, 3, 4})
	f.Add(uint16(2), uint8(0), uint8(0), []byte{})
	f.Fuzz(func(t *testing.T, rawBits uint16, rawCount, rawLiteral uint8, data []byte) {
		bitCount := int(rawBits)%129 + 2
		count := int(rawCount) % 12
		literal := int(rawLiteral) % bitCount
		values := make([][]uint32, count)
		for index := range values {
			values[index] = mvcDivisionWords(bitCount, index, data, 0x69)
		}
		manager := NewMvcManager()
		cover, err := mvcCoverFromWords(manager, bitCount, values)
		if err != nil {
			t.Fatal(err)
		}
		literalDivisor, _ := NewMvcCover(manager, bitCount)
		literalCube, _ := literalDivisor.NewCube()
		_ = literalCube.SetBit(literal, true)
		_ = literalDivisor.AddTail(literalCube)
		quotient, remainder, err := cover.DivideByLiteral(literalDivisor)
		if err != nil {
			t.Fatal(err)
		}
		wantQuotient, wantRemainder := make([][]uint32, 0), make([][]uint32, 0)
		for _, words := range values {
			copyWords := slices.Clone(words)
			if copyWords[literal>>5]&(uint32(1)<<uint(literal&31)) != 0 {
				copyWords[literal>>5] &^= uint32(1) << uint(literal&31)
				wantQuotient = append(wantQuotient, copyWords)
			} else {
				wantRemainder = append(wantRemainder, copyWords)
			}
		}
		for index, cube := range quotient.cubes {
			got, _ := cube.Words()
			if !slices.Equal(got, wantQuotient[index]) {
				t.Fatal("literal quotient", index, got, wantQuotient[index])
			}
		}
		for index, cube := range remainder.cubes {
			got, _ := cube.Words()
			if !slices.Equal(got, wantRemainder[index]) {
				t.Fatal("literal remainder", index, got, wantRemainder[index])
			}
		}
		inPlace, _ := mvcCoverFromWords(manager, bitCount, values)
		if err = inPlace.DivideByLiteralQuotient(literal); err != nil {
			t.Fatal(err)
		}
		if len(inPlace.cubes) != len(wantQuotient) {
			t.Fatal(len(inPlace.cubes), len(wantQuotient))
		}
		for index, cube := range inPlace.cubes {
			got, _ := cube.Words()
			if !slices.Equal(got, wantQuotient[index]) {
				t.Fatal("in-place", index, got, wantQuotient[index])
			}
		}

		cubeDivisor, _ := NewMvcCover(manager, bitCount)
		divisorCube, _ := cubeDivisor.NewCube()
		_ = divisorCube.SetBit(literal, true)
		secondLiteral := (literal + 1) % bitCount
		_ = divisorCube.SetBit(secondLiteral, true)
		_ = cubeDivisor.AddTail(divisorCube)
		cubeQuotient, cubeRemainder, err := cover.DivideByCube(cubeDivisor)
		if err != nil {
			t.Fatal(err)
		}
		qIndex, rIndex := 0, 0
		for _, words := range values {
			contains := words[literal>>5]&(uint32(1)<<uint(literal&31)) != 0 && words[secondLiteral>>5]&(uint32(1)<<uint(secondLiteral&31)) != 0
			if contains {
				got, _ := cubeQuotient.cubes[qIndex].Words()
				want := slices.Clone(words)
				want[literal>>5] &^= uint32(1) << uint(literal&31)
				want[secondLiteral>>5] &^= uint32(1) << uint(secondLiteral&31)
				if !slices.Equal(got, want) {
					t.Fatal("cube quotient", qIndex, got, want)
				}
				qIndex++
			} else {
				got, _ := cubeRemainder.cubes[rIndex].Words()
				if !slices.Equal(got, words) {
					t.Fatal("cube remainder", rIndex, got, words)
				}
				rIndex++
			}
		}
		for _, candidate := range []*MvcCover{cubeRemainder, cubeQuotient, cubeDivisor, inPlace, remainder, quotient, literalDivisor, cover} {
			if err = candidate.Close(); err != nil {
				t.Fatal(err)
			}
		}
		if err = manager.Close(); err != nil {
			t.Fatal(err)
		}
	})
}

func FuzzMvcDivideGeneralIdentity(f *testing.F) {
	f.Add(uint16(66), uint8(5), uint8(2))
	f.Add(uint16(8), uint8(2), uint8(1))
	f.Fuzz(func(t *testing.T, rawBits uint16, rawGroups, rawFailed uint8) {
		bitCount := int(rawBits)%123 + 8
		groups := int(rawGroups)%8 + 2
		failedGroup := int(rawFailed) % groups
		manager := NewMvcManager()
		divisor, _ := NewMvcCover(manager, bitCount)
		for literal := 0; literal < 2; literal++ {
			cube, _ := divisor.NewCube()
			_ = cube.SetBit(literal, true)
			_ = divisor.AddTail(cube)
		}
		cover, _ := NewMvcCover(manager, bitCount)
		for group := 0; group < groups; group++ {
			outside := 2 + group%(bitCount-2)
			limit := 2
			if group == failedGroup {
				limit = 1
			}
			for literal := 0; literal < limit; literal++ {
				cube, _ := cover.NewCube()
				_ = cube.SetBit(outside, true)
				_ = cube.SetBit(literal, true)
				_ = cover.AddTail(cube)
			}
		}
		original, _ := cover.Duplicate()
		quotient, remainder, err := cover.Divide(divisor)
		if err != nil || quotient == nil || remainder == nil {
			t.Fatal(quotient, remainder, err)
		}
		product, err := divisor.AlgebraicMultiply(quotient)
		if err != nil {
			t.Fatal(err)
		}
		reconstructed, err := product.BooleanOr(remainder)
		if err != nil {
			t.Fatal(err)
		}
		forward, err := original.AlgebraicEqual(reconstructed)
		reverse, reverseErr := reconstructed.AlgebraicEqual(original)
		if err != nil || reverseErr != nil || !forward || !reverse {
			t.Fatal(forward, reverse, err, reverseErr)
		}
		for _, candidate := range []*MvcCover{reconstructed, product, remainder, quotient, original, cover, divisor} {
			if err = candidate.Close(); err != nil {
				t.Fatal(err)
			}
		}
		if err = manager.Close(); err != nil {
			t.Fatal(err)
		}
	})
}
