package core

import (
	"errors"
	"math/big"
	"math/bits"
	"strings"
	"testing"
)

func TestAPAArithmeticAliasingAndBorrow(t *testing.T) {
	a := DdApaNumber{0xffffffff, 0xffffffff}
	b := DdApaNumber{0, 1}
	sum := make(DdApaNumber, 2)
	if Cudd_ApaAdd(2, a, b, sum) != 1 || apaInteger(2, sum).Sign() != 0 {
		t.Fatal("carry")
	}
	if Cudd_ApaSubtract(2, sum, b, sum) != 0xffffffff || Cudd_ApaCompare(2, sum, 2, a) != 0 {
		t.Fatal("borrow/alias")
	}
	quotient := append(DdApaNumber(nil), a...)
	remainder := Cudd_ApaIntDivision(2, quotient, 7, quotient)
	recovered := apaInteger(2, quotient)
	recovered.Mul(recovered, big.NewInt(7))
	recovered.Add(recovered, new(big.Int).SetUint64(uint64(remainder)))
	if recovered.Cmp(apaInteger(2, a)) != 0 {
		t.Fatal("division")
	}
	shifted := DdApaNumber{1, 0}
	Cudd_ApaShiftRight(2, 1, shifted, shifted)
	if shifted[0] != 0x80000000 || shifted[1] != 0x80000000 {
		t.Fatal("in-place shift", shifted)
	}
	n := make(DdApaNumber, 3)
	Cudd_ApaPowerOfTwo(3, n, 64)
	if n[0] != 1 || n[1] != 0 || n[2] != 0 {
		t.Fatal(n)
	}
	Cudd_ApaPowerOfTwo(3, n, 96)
	if apaInteger(3, n).Sign() != 0 {
		t.Fatal("overflow power")
	}
	Cudd_ApaSetToLiteral(3, n, 19)
	if apaInteger(3, n).Int64() != 19 {
		t.Fatal(n)
	}
	if Cudd_ApaNumberOfDigits(32) != 1 || Cudd_ApaNumberOfDigits(33) != 2 {
		t.Fatal("digit rounding")
	}
}
func TestAPARatiosAndPrinting(t *testing.T) {
	if Cudd_ApaCompareRatios(1, DdApaNumber{7}, 3, 1, DdApaNumber{14}, 6) != 0 || Cudd_ApaCompareRatios(1, DdApaNumber{7}, 3, 1, DdApaNumber{2}, 1) != 1 {
		t.Fatal("ratios")
	}
	var out strings.Builder
	Cudd_ApaPrintHex(&out, 2, DdApaNumber{1, 2})
	if out.String() != "0000000100000002" {
		t.Fatal(out.String())
	}
	out.Reset()
	Cudd_ApaPrintDecimal(&out, 2, DdApaNumber{1, 2})
	if out.String() != "4294967298" {
		t.Fatal(out.String())
	}
	out.Reset()
	Cudd_ApaPrintExponential(&out, 2, DdApaNumber{1, 2}, 4)
	if out.String() != "4.294e+9" {
		t.Fatal(out.String())
	}
	if Cudd_ApaPrintDecimal(apaFailWriter{}, 1, DdApaNumber{1}) != 0 || Cudd_ApaPrintHex(apaFailWriter{}, 1, DdApaNumber{1}) != 0 {
		t.Fatal("writer error ignored")
	}
}

type apaFailWriter struct{}

func (apaFailWriter) Write([]byte) (int, error) { return 0, errors.New("write failed") }
func TestAPAExactMintermsBeyondFloatPrecision(t *testing.T) {
	dd := NewBDDManager(2)
	f := Cudd_bddXor(dd, dd.Vars[0], dd.Vars[1])
	var digits int
	count := Cudd_ApaCountMinterm(dd, f, 200, &digits)
	want := new(big.Int).Lsh(big.NewInt(1), 199)
	if count == nil || apaInteger(digits, count).Cmp(want) != 0 {
		t.Fatal("exact minterms", digits, count)
	}
	complemented := Cudd_ApaCountMinterm(dd, Cudd_Not(f), 200, &digits)
	if apaInteger(digits, complemented).Cmp(want) != 0 {
		t.Fatal("complement count")
	}
	var out strings.Builder
	if Cudd_ApaPrintMinterm(&out, dd, dd.One, 4) != 1 || out.String() != "16\n" {
		t.Fatal(out.String())
	}
	out.Reset()
	if Cudd_ApaPrintDensity(&out, dd, dd.Vars[0], 4) != 1 || out.String() != "4.0\n" {
		t.Fatal(out.String())
	}
	dd.Background = dd.One
	if apaInteger(digits, Cudd_ApaCountMinterm(dd, dd.One, 200, &digits)).Sign() != 0 {
		t.Fatal("background not honored")
	}
}

func TestAPACountUsesConfiguredADDBackground(t *testing.T) {
	dd := NewBDDManager(2)
	defer Cudd_Quit(dd)
	Cudd_SetBackground(dd, dd.PlusInfinity)
	f := Cudd_addIte(dd, Cudd_addIthVar(dd, 1), dd.Zero, dd.PlusInfinity)
	for _, tc := range []struct {
		node *DdNode
		want uint32
	}{{dd.Zero, 16}, {dd.PlusInfinity, 0}, {f, 8}} {
		var digits int
		got := Cudd_ApaCountMinterm(dd, tc.node, 4, &digits)
		if digits != 1 || len(got) != 1 || got[0] != tc.want {
			t.Errorf("node %d count %v, want %d", tc.node.ID, got, tc.want)
		}
	}
}

func TestAPACompareIgnoresLeadingDigitsWithoutAllocating(t *testing.T) {
	a := DdApaNumber{0, 0, 1, 7}
	b := DdApaNumber{1, 7, 99}
	if Cudd_ApaCompare(4, a, 2, b) != 0 || Cudd_ApaCompare(3, a, 2, b) >= 0 || Cudd_ApaCompare(2, b, 3, a) <= 0 || Cudd_ApaCompare(0, nil, 2, a) != 0 {
		t.Fatal("comparison with unequal widths or leading zeros")
	}
	if n := testing.AllocsPerRun(20, func() { Cudd_ApaCompare(4, a, 2, b) }); n != 0 {
		t.Fatalf("comparing digit slices allocated %g objects", n)
	}
}

func TestAPAScaledCountsAcrossDigitBoundaries(t *testing.T) {
	dd := NewBDDManager(3)
	defer Cudd_Quit(dd)
	for table := 0; table < 256; table++ {
		root := bddFromTruth(dd, uint8(table), 3)
		for _, width := range []int{3, 31, 32, 33, 63, 64, 65, 97, 511, 512, 513} {
			var digits int
			got := Cudd_ApaCountMinterm(dd, root, width, &digits)
			want := new(big.Int).Lsh(big.NewInt(int64(bits.OnesCount8(uint8(table)))), uint(width-3))
			if digits != (width+32)/32 || apaInteger(digits, got).Cmp(want) != 0 {
				t.Fatalf("table %02x at width %d got %v, want %x", table, width, got, want)
			}
		}
	}
}
