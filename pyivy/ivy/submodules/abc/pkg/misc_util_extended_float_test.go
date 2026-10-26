package abc

// Independent integer-model and fuzz tests for utilFloat.h and utilDouble.h.

import (
	"errors"
	"math"
	"math/big"
	"testing"
)

func TestExtendedFloatConstantsConversionsAndArithmetic(t *testing.T) {
	tests32 := []struct {
		number ExtendedFloat32
		value  float32
	}{{ExtendedFloat32One, 1}, {ExtendedFloat32Two, 2}, {ExtendedFloat32Three, 3}, {ExtendedFloat32Twelve, 12}, {ExtendedFloat32OneHalf, 1.5}, {ExtendedFloat32TwoHalf, 2.5}}
	for _, test := range tests32 {
		value, err := test.number.Float32()
		if err != nil || value != test.value {
			t.Fatal(test, value, err)
		}
		number, err := ExtendedFloat32FromFloat(test.value)
		if err != nil || number != test.number {
			t.Fatal(test, number, err)
		}
	}
	tests64 := []struct {
		number ExtendedFloat64
		value  float64
	}{{ExtendedFloat64One, 1}, {ExtendedFloat64Two, 2}, {ExtendedFloat64Three, 3}, {ExtendedFloat64Twelve, 12}, {ExtendedFloat64OneHalf, 1.5}, {ExtendedFloat64TwoHalf, 2.5}}
	for _, test := range tests64 {
		value, err := test.number.Float64()
		if err != nil || value != test.value {
			t.Fatal(test, value, err)
		}
		number, err := ExtendedFloat64FromFloat(test.value)
		if err != nil || number != test.number {
			t.Fatal(test, number, err)
		}
	}
	if number, err := ExtendedFloat64FromFloat(1e100); err != nil || number.Bits() != 0x014c924d692ca61b {
		t.Fatalf("1e100: %016x %v", number.Bits(), err)
	}
	if sum, err := ExtendedFloat32One.Add(ExtendedFloat32OneHalf); err != nil || sum != ExtendedFloat32TwoHalf {
		t.Fatal(sum, err)
	}
	if product, err := ExtendedFloat32Two.Multiply(ExtendedFloat32OneHalf); err != nil || product != ExtendedFloat32Three {
		t.Fatal(product, err)
	}
	if sum, err := ExtendedFloat64One.Add(ExtendedFloat64OneHalf); err != nil || sum != ExtendedFloat64TwoHalf {
		t.Fatal(sum, err)
	}
	if product, err := ExtendedFloat64Two.Multiply(ExtendedFloat64OneHalf); err != nil || product != ExtendedFloat64Three {
		t.Fatal(product, err)
	}
	if quotient, err := ExtendedFloat32Twelve.DividePowerOfTwo(2); err != nil || quotient != ExtendedFloat32Three {
		t.Fatal(quotient, err)
	}
	if quotient, err := ExtendedFloat64Twelve.DividePowerOfTwo(2); err != nil || quotient != ExtendedFloat64Three {
		t.Fatal(quotient, err)
	}
	if less, err := ExtendedFloat32One.Less(ExtendedFloat32Two); err != nil || !less {
		t.Fatal(less, err)
	}
	if less, err := ExtendedFloat64Two.Less(ExtendedFloat64One); err != nil || less {
		t.Fatal(less, err)
	}
}

func TestExtendedFloatBoundariesAndErrors(t *testing.T) {
	if sum, err := ExtendedFloat32Maximum.Add(ExtendedFloat32Maximum); err != nil || sum != ExtendedFloat32Maximum {
		t.Fatal(sum, err)
	}
	if product, err := ExtendedFloat64Maximum.Multiply(ExtendedFloat64Maximum); err != nil || product != ExtendedFloat64Maximum {
		t.Fatal(product, err)
	}
	if value, err := ExtendedFloat32One.DividePowerOfTwo(1); err != nil || value != ExtendedFloat32One {
		t.Fatal(value, err)
	}
	if value, err := ExtendedFloat64One.DividePowerOfTwo(math.MaxUint32); err != nil || value != ExtendedFloat64One {
		t.Fatal(value, err)
	}
	if _, err := ExtendedFloat32One.DividePowerOfTwo(0xffff); !errors.Is(err, ErrDimension) {
		t.Fatal(err)
	}
	for _, input := range []float32{0, -1, float32(math.Inf(1)), float32(math.NaN())} {
		if _, err := ExtendedFloat32FromFloat(input); !errors.Is(err, ErrDimension) {
			t.Fatal(input, err)
		}
	}
	for _, input := range []float64{0, -1, math.Inf(1), math.NaN()} {
		if _, err := ExtendedFloat64FromFloat(input); !errors.Is(err, ErrDimension) {
			t.Fatal(input, err)
		}
	}
	if _, err := NewExtendedFloat32(0x10000, 0x8000); !errors.Is(err, ErrDimension) {
		t.Fatal(err)
	}
	if _, err := NewExtendedFloat32(0, 0x7fff); !errors.Is(err, ErrDimension) {
		t.Fatal(err)
	}
	if _, err := NewExtendedFloat64(0, 1<<47-1); !errors.Is(err, ErrDimension) {
		t.Fatal(err)
	}
	if _, err := ExtendedFloat32(0).Add(ExtendedFloat32One); !errors.Is(err, ErrDimension) {
		t.Fatal(err)
	}
	if _, err := ExtendedFloat64(0).Float64(); !errors.Is(err, ErrDimension) {
		t.Fatal(err)
	}
	large32, _ := NewExtendedFloat32(127, 0x8000)
	if _, err := large32.Float32(); !errors.Is(err, ErrDimension) {
		t.Fatal(err)
	}
	large64, _ := NewExtendedFloat64(1023, 1<<47)
	if _, err := large64.Float64(); !errors.Is(err, ErrDimension) {
		t.Fatal(err)
	}
}

func extendedFloat32AddModel(a, b ExtendedFloat32) ExtendedFloat32 {
	if a < b {
		a, b = b, a
	}
	difference := a.Exponent() - b.Exponent()
	aligned := uint64(0)
	if difference < 64 {
		aligned = uint64(b.Mantissa()) >> difference
	}
	mantissa := uint64(a.Mantissa()) + aligned
	exponent := uint64(a.Exponent())
	if mantissa >= 1<<16 {
		exponent++
		mantissa >>= 1
	}
	if exponent > 0xffff {
		return ExtendedFloat32Maximum
	}
	return ExtendedFloat32(uint32(exponent)<<16 | uint32(mantissa))
}

func extendedFloat32MultiplyModel(a, b ExtendedFloat32) ExtendedFloat32 {
	mantissa := uint64(a.Mantissa()) * uint64(b.Mantissa()) >> 15
	exponent := uint64(a.Exponent()) + uint64(b.Exponent())
	if mantissa >= 1<<16 {
		exponent++
		mantissa >>= 1
	}
	if exponent > 0xffff {
		return ExtendedFloat32Maximum
	}
	return ExtendedFloat32(uint32(exponent)<<16 | uint32(mantissa))
}

func extendedFloat64MultiplyModel(a, b ExtendedFloat64) ExtendedFloat64 {
	ah, bh := new(big.Int).SetUint64(a.Mantissa()>>32), new(big.Int).SetUint64(b.Mantissa()>>32)
	al, bl := new(big.Int).SetUint64(uint64(uint32(a.Mantissa()))), new(big.Int).SetUint64(uint64(uint32(b.Mantissa())))
	term := new(big.Int).Lsh(new(big.Int).Mul(ah, bh), 17)
	term.Add(term, new(big.Int).Rsh(new(big.Int).Mul(al, bl), 47))
	term.Add(term, new(big.Int).Rsh(new(big.Int).Mul(al, bh), 15))
	term.Add(term, new(big.Int).Rsh(new(big.Int).Mul(ah, bl), 15))
	mantissa := term.Uint64()
	exponent := a.Exponent() + b.Exponent()
	if mantissa >= 1<<48 {
		exponent++
		mantissa >>= 1
	}
	if exponent > 0xffff {
		return ExtendedFloat64Maximum
	}
	return ExtendedFloat64(exponent<<48 | mantissa)
}

func FuzzExtendedFloatArithmetic(f *testing.F) {
	for _, seed := range [][]byte{{}, {0}, {1, 2, 3, 4}, {255, 255, 0, 127, 19}, {17, 31, 63, 127, 255, 1}} {
		f.Add(seed)
	}
	f.Fuzz(func(t *testing.T, data []byte) {
		byteAt := func(index int) byte {
			if len(data) == 0 {
				return byte(index * 37)
			}
			return data[index%len(data)]
		}
		exponentA := uint32(byteAt(0))<<8 | uint32(byteAt(1))
		exponentB := uint32(byteAt(2))<<8 | uint32(byteAt(3))
		mantissaA := uint32(0x8000) | uint32(byteAt(4))<<7 | uint32(byteAt(5))>>1
		mantissaB := uint32(0x8000) | uint32(byteAt(6))<<7 | uint32(byteAt(7))>>1
		a, _ := NewExtendedFloat32(exponentA, mantissaA)
		b, _ := NewExtendedFloat32(exponentB, mantissaB)
		if got, err := a.Add(b); err != nil || got != extendedFloat32AddModel(a, b) {
			t.Fatal(got, err)
		}
		if got, err := a.Multiply(b); err != nil || got != extendedFloat32MultiplyModel(a, b) {
			t.Fatal(got, err)
		}
		mantissa64A := uint64(1)<<47 | uint64(byteAt(8))<<39 | uint64(byteAt(9))<<31 | uint64(byteAt(10))<<23 | uint64(byteAt(11))<<15 | uint64(byteAt(12))<<7 | uint64(byteAt(13))>>1
		mantissa64B := uint64(1)<<47 | uint64(byteAt(14))<<39 | uint64(byteAt(15))<<31 | uint64(byteAt(16))<<23 | uint64(byteAt(17))<<15 | uint64(byteAt(18))<<7 | uint64(byteAt(19))>>1
		c, _ := NewExtendedFloat64(uint64(exponentA), mantissa64A)
		d, _ := NewExtendedFloat64(uint64(exponentB), mantissa64B)
		if got, err := c.Multiply(d); err != nil || got != extendedFloat64MultiplyModel(c, d) {
			t.Fatal(got, err)
		}
		gotAdd, err := c.Add(d)
		if err != nil {
			t.Fatal(err)
		}
		larger, smaller := c, d
		if larger < smaller {
			larger, smaller = smaller, larger
		}
		difference := larger.Exponent() - smaller.Exponent()
		aligned := uint64(0)
		if difference < 64 {
			aligned = smaller.Mantissa() >> difference
		}
		mantissa := larger.Mantissa() + aligned
		exponent := larger.Exponent()
		if mantissa >= 1<<48 {
			exponent++
			mantissa >>= 1
		}
		wantAdd := ExtendedFloat64Maximum
		if exponent <= 0xffff {
			wantAdd = ExtendedFloat64(exponent<<48 | mantissa)
		}
		if gotAdd != wantAdd {
			t.Fatal(gotAdd, wantAdd)
		}
	})
}
