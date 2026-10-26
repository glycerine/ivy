package abc

// Tests for the arbitrary-precision arithmetic port from src/bdd/cudd/cuddApa.c.

import (
	"errors"
	"math/big"
	"math/rand"
	"slices"
	"strconv"
	"strings"
	"testing"
)

func cuddAPATestBig(number CuddAPANumber) *big.Int {
	result := new(big.Int)
	for _, digit := range number {
		result.Lsh(result, 32)
		result.Or(result, new(big.Int).SetUint64(uint64(digit)))
	}
	return result
}

func cuddAPATestFromBig(value *big.Int, digits int) CuddAPANumber {
	result := make(CuddAPANumber, digits)
	work := new(big.Int).Set(value)
	mask := new(big.Int).SetUint64(0xffffffff)
	for index := digits - 1; index >= 0; index-- {
		result[index] = uint32(new(big.Int).And(work, mask).Uint64())
		work.Rsh(work, 32)
	}
	return result
}

func TestCuddAPABasicsAndFormatting(t *testing.T) {
	for bitsCount, want := range map[int]int{1: 1, 32: 1, 33: 2, 64: 2, 65: 3} {
		got, err := CuddAPANumberOfDigits(bitsCount)
		if err != nil || got != want {
			t.Fatalf("digits %d: %d %v", bitsCount, got, err)
		}
	}
	if _, err := CuddAPANumberOfDigits(0); !errors.Is(err, ErrDimension) {
		t.Fatal(err)
	}
	if _, err := NewCuddAPANumber(0); !errors.Is(err, ErrDimension) {
		t.Fatal(err)
	}
	number := CuddAPANumber{1, 2}
	clone := make(CuddAPANumber, 2)
	if err := CuddAPACopy(number, clone); err != nil || !slices.Equal(number, clone) {
		t.Fatal(clone, err)
	}
	clone[0] = 9
	if number[0] != 1 {
		t.Fatal("copy aliases source")
	}
	if err := CuddAPASetToLiteral(clone, 17); err != nil || !slices.Equal(clone, CuddAPANumber{0, 17}) {
		t.Fatal(clone, err)
	}
	if err := CuddAPAPowerOfTwo(clone, 63); err != nil || !slices.Equal(clone, CuddAPANumber{0x80000000, 0}) {
		t.Fatal(clone, err)
	}
	if err := CuddAPAPowerOfTwo(clone, 64); err != nil || !slices.Equal(clone, CuddAPANumber{0, 0}) {
		t.Fatal(clone, err)
	}
	if err := CuddAPAPowerOfTwo(clone, -1); !errors.Is(err, ErrDimension) {
		t.Fatal(err)
	}
	hex, err := CuddAPAHex(number)
	if err != nil || hex != "0000000100000002" {
		t.Fatal(hex, err)
	}
	decimal, err := CuddAPADecimal(number)
	if err != nil || decimal != "4294967298" {
		t.Fatal(decimal, err)
	}
	exponential, err := CuddAPAExponential(CuddAPANumber{123456}, 4)
	if err != nil || exponential != "1.234e+5" {
		t.Fatal(exponential, err)
	}
	exponential, err = CuddAPAExponential(CuddAPANumber{0}, 8)
	if err != nil || exponential != "0e+0" {
		t.Fatal(exponential, err)
	}
	if _, err = CuddAPAExponential(number, 0); !errors.Is(err, ErrDimension) {
		t.Fatal(err)
	}
	if _, err = CuddAPADecimalWithOptions(number, CuddAPAOptions{MaxOperations: 1}); !errors.Is(err, ErrLimit) {
		t.Fatal(err)
	}
	if _, err = CuddAPAExponentialWithOptions(number, 2, CuddAPAOptions{Stop: func() bool { return true }}); !errors.Is(err, ErrCancelled) {
		t.Fatal(err)
	}
}

func TestCuddAPAArithmeticAgainstBigInt(t *testing.T) {
	random := rand.New(rand.NewSource(0x415041))
	modulus := new(big.Int).Lsh(big.NewInt(1), 32*6)
	for iteration := 0; iteration < 5000; iteration++ {
		digits := 1 + random.Intn(6)
		first := make(CuddAPANumber, digits)
		second := make(CuddAPANumber, digits)
		for index := range first {
			first[index] = random.Uint32()
			second[index] = random.Uint32()
		}
		firstBig, secondBig := cuddAPATestBig(first), cuddAPATestBig(second)
		widthModulus := new(big.Int).Rsh(new(big.Int).Set(modulus), uint(32*(6-digits)))

		sum := make(CuddAPANumber, digits)
		carry, err := CuddAPAAdd(first, second, sum)
		if err != nil {
			t.Fatal(err)
		}
		modelSum := new(big.Int).Add(firstBig, secondBig)
		modelCarry := uint32(0)
		if modelSum.Cmp(widthModulus) >= 0 {
			modelCarry = 1
			modelSum.Sub(modelSum, widthModulus)
		}
		if carry != modelCarry || cuddAPATestBig(sum).Cmp(modelSum) != 0 {
			t.Fatalf("add mismatch iteration %d", iteration)
		}

		difference := make(CuddAPANumber, digits)
		borrow, err := CuddAPASubtract(first, second, difference)
		if err != nil {
			t.Fatal(err)
		}
		modelDifference := new(big.Int).Sub(firstBig, secondBig)
		modelBorrow := uint32(0)
		if modelDifference.Sign() < 0 {
			modelBorrow = 0xffffffff
			modelDifference.Add(modelDifference, widthModulus)
		}
		if borrow != modelBorrow || cuddAPATestBig(difference).Cmp(modelDifference) != 0 {
			t.Fatalf("subtract mismatch iteration %d", iteration)
		}

		divisor := random.Uint32() | 1
		quotient := make(CuddAPANumber, digits)
		remainder, err := CuddAPAIntDivision(first, divisor, quotient)
		if err != nil {
			t.Fatal(err)
		}
		modelQuotient, modelRemainder := new(big.Int), new(big.Int)
		modelQuotient.QuoRem(firstBig, new(big.Int).SetUint64(uint64(divisor)), modelRemainder)
		if cuddAPATestBig(quotient).Cmp(modelQuotient) != 0 || remainder != uint32(modelRemainder.Uint64()) {
			t.Fatalf("division mismatch iteration %d", iteration)
		}

		shifted := append(CuddAPANumber(nil), first...)
		in := random.Uint32()
		if err = CuddAPAShiftRight(in, shifted, shifted); err != nil {
			t.Fatal(err)
		}
		modelShift := new(big.Int).Rsh(firstBig, 1)
		if in&1 != 0 {
			modelShift.SetBit(modelShift, digits*32-1, 1)
		}
		if cuddAPATestBig(shifted).Cmp(modelShift) != 0 {
			t.Fatalf("shift mismatch iteration %d", iteration)
		}

		comparison, err := CuddAPACompare(first, second)
		if err != nil || comparison != firstBig.Cmp(secondBig) {
			t.Fatalf("compare mismatch iteration %d: %d %v", iteration, comparison, err)
		}
		decimal, err := CuddAPADecimal(first)
		if err != nil || decimal != firstBig.String() {
			t.Fatalf("decimal mismatch iteration %d: %s %s %v", iteration, decimal, firstBig, err)
		}
	}
}

func TestCuddAPAAliasingRatiosAndErrors(t *testing.T) {
	first := CuddAPANumber{0xffffffff, 0xffffffff}
	second := CuddAPANumber{0, 1}
	want := new(big.Int).Add(cuddAPATestBig(first), cuddAPATestBig(second))
	want.And(want, new(big.Int).Sub(new(big.Int).Lsh(big.NewInt(1), 64), big.NewInt(1)))
	if carry, err := CuddAPAAdd(first, second, first); err != nil || carry != 1 || cuddAPATestBig(first).Cmp(want) != 0 {
		t.Fatal(first, carry, err)
	}
	value := CuddAPANumber{0x80000001, 3}
	model, modelRemainder := new(big.Int), new(big.Int)
	model.QuoRem(cuddAPATestBig(value), big.NewInt(7), modelRemainder)
	if remainder, err := CuddAPAShortDivision(value, 7, value); err != nil || remainder != uint32(modelRemainder.Uint64()) || cuddAPATestBig(value).Cmp(model) != 0 {
		t.Fatal(value, remainder, err)
	}
	left := CuddAPANumber{10}
	right := CuddAPANumber{15}
	comparison, err := CuddAPACompareRatios(left, 3, right, 5)
	if err != nil || comparison != 1 {
		t.Fatal(comparison, err)
	}
	comparison, err = CuddAPACompareRatios(CuddAPANumber{2}, 3, CuddAPANumber{4}, 6)
	if err != nil || comparison != 0 {
		t.Fatal(comparison, err)
	}
	if _, err = CuddAPACompareRatios(left, 0, right, 1); !errors.Is(err, ErrDimension) {
		t.Fatal(err)
	}
	if _, err = CuddAPAShortDivision(left, 0, make(CuddAPANumber, 1)); !errors.Is(err, ErrDimension) {
		t.Fatal(err)
	}
	if _, err = CuddAPAAdd(left, CuddAPANumber{1, 2}, make(CuddAPANumber, 1)); !errors.Is(err, ErrDimension) {
		t.Fatal(err)
	}
}

func TestCuddAPABDDMintermsAndDensity(t *testing.T) {
	manager, err := NewExtraBDDManager(3, 128)
	if err != nil {
		t.Fatal(err)
	}
	x0, _ := manager.Variable(0)
	x1, _ := manager.Variable(1)
	root, err := manager.And(x0, x1, ExtraBDDOptions{})
	if err != nil {
		t.Fatal(err)
	}
	for _, candidate := range []ExtraBDDLit{ExtraBDDFalse, ExtraBDDTrue, x0, root, root.Not()} {
		model := 0
		for assignment := 0; assignment < 8; assignment++ {
			values := []bool{assignment&1 != 0, assignment&2 != 0, assignment&4 != 0}
			value, evalErr := manager.Evaluate(candidate, values)
			if evalErr != nil {
				t.Fatal(evalErr)
			}
			if value {
				model++
			}
		}
		count, countErr := manager.CuddAPACountMinterms(candidate, 3, CuddAPAOptions{})
		if countErr != nil || cuddAPATestBig(count).Cmp(big.NewInt(int64(model))) != 0 {
			t.Fatalf("candidate %d: %v %v want %d", candidate, count, countErr, model)
		}
	}
	text, err := manager.CuddAPAMintermDecimal(root, 3, CuddAPAOptions{})
	if err != nil || text != "2" {
		t.Fatal(text, err)
	}
	exponential, err := manager.CuddAPAMintermExponential(root, 3, 3, CuddAPAOptions{})
	if err != nil || exponential != "2e+0" {
		t.Fatal(exponential, err)
	}
	density, err := manager.CuddAPADensity(x0, 3, CuddAPAOptions{})
	if err != nil || density != "2.0" {
		t.Fatal(density, err)
	}
	if _, err = manager.CuddAPAMintermDecimal(ExtraBDDTrue, 1, CuddAPAOptions{MaxOperations: 3}); !errors.Is(err, ErrLimit) {
		t.Fatalf("decimal wrapper ignored formatting budget: %v", err)
	}
	if count, countErr := manager.CuddAPACountMinterms(root, 1, CuddAPAOptions{}); count != nil || !errors.Is(countErr, ErrDimension) {
		t.Fatal(count, countErr)
	}
	if count, countErr := manager.CuddAPACountMinterms(root, 3, CuddAPAOptions{MaxDigits: 1, MaxOperations: 1}); count != nil || !errors.Is(countErr, ErrLimit) {
		t.Fatal(count, countErr)
	}
	if count, countErr := manager.CuddAPACountMinterms(root, 3, CuddAPAOptions{Stop: func() bool { return true }}); count != nil || !errors.Is(countErr, ErrCancelled) {
		t.Fatal(count, countErr)
	}
	if count, countErr := manager.CuddAPACountMinterms(ExtraBDDLit(10000), 3, CuddAPAOptions{}); count != nil || !errors.Is(countErr, ErrInvalidLiteral) {
		t.Fatal(count, countErr)
	}
}

func FuzzCuddAPAArithmetic(f *testing.F) {
	f.Add([]byte{0})
	f.Add([]byte{0xff, 0xff, 0xff, 0xff})
	f.Add([]byte("arbitrary precision"))
	f.Fuzz(func(t *testing.T, data []byte) {
		if len(data) > 128 {
			data = data[:128]
		}
		digits := (len(data) + 3) / 4
		if digits == 0 {
			digits = 1
		}
		number := make(CuddAPANumber, digits)
		for index, value := range data {
			shift := uint(24 - 8*(index&3))
			number[index/4] |= uint32(value) << shift
		}
		model := cuddAPATestBig(number)
		decimal, err := CuddAPADecimal(number)
		if err != nil || decimal != model.String() {
			t.Fatalf("decimal %q want %q: %v", decimal, model, err)
		}
		hex, err := CuddAPAHex(number)
		wantHex := strings.Repeat("0", digits*8-len(model.Text(16))) + model.Text(16)
		if err != nil || hex != wantHex {
			t.Fatalf("hex %q want %q: %v", hex, wantHex, err)
		}
		divisor := uint32(1)
		if len(data) >= 4 {
			parsed, parseErr := strconv.ParseUint(hex[len(hex)-8:], 16, 32)
			if parseErr != nil {
				t.Fatal(parseErr)
			}
			divisor = uint32(parsed) | 1
		}
		quotient := make(CuddAPANumber, digits)
		remainder, err := CuddAPAIntDivision(number, divisor, quotient)
		modelQuotient, modelRemainder := new(big.Int), new(big.Int)
		modelQuotient.QuoRem(model, new(big.Int).SetUint64(uint64(divisor)), modelRemainder)
		if err != nil || cuddAPATestBig(quotient).Cmp(modelQuotient) != 0 || remainder != uint32(modelRemainder.Uint64()) {
			t.Fatalf("division mismatch: %v", err)
		}
	})
}
