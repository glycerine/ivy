package util

// Floating-point arithmetic from ABC utilFloat.h (Alan Mishchenko and Bruno
// Schmitt). Values have a 16-bit exponent and normalized 16-bit mantissa;
// their numeric value is 2^Exp * Mnt/32768. Overflow saturates, and division
// underflow returns the minimum representable value (one).
import (
	"fmt"
	"math"
)

type xFloat_t struct{ Mnt, Exp uint16 }

func xSat_Float2Uint(f xFloat_t) uint32         { return uint32(f.Exp)<<16 | uint32(f.Mnt) }
func xSat_Uint2Float(u uint32) xFloat_t         { return xFloat_t{uint16(u), uint16(u >> 16)} }
func xSat_LessThan(a, b xFloat_t) bool          { return xSat_Float2Uint(a) < xSat_Float2Uint(b) }
func xSat_Equal(a, b xFloat_t) bool             { return a == b }
func xSat_FloatCreate(exp, mnt uint32) xFloat_t { return xFloat_t{uint16(mnt), uint16(exp)} }
func xSat_FloatCreateConst1() xFloat_t          { return xSat_FloatCreate(0, 1<<15) }
func xSat_FloatCreateConst2() xFloat_t          { return xSat_FloatCreate(1, 1<<15) }
func xSat_FloatCreateConst3() xFloat_t          { return xSat_FloatCreate(1, 3<<14) }
func xSat_FloatCreateConst12() xFloat_t         { return xSat_FloatCreate(3, 3<<14) }
func xSat_FloatCreateConst1point5() xFloat_t    { return xSat_FloatCreate(0, 3<<14) }
func xSat_FloatCreateConst2point5() xFloat_t    { return xSat_FloatCreate(1, 5<<13) }
func xSat_FloatCreateMaximum() xFloat_t         { return xSat_Uint2Float(math.MaxUint32) }
func xSat_Float2Float(a xFloat_t) float32 {
	if a.Exp >= 127 {
		panic("xSat_Float2Float: exponent exceeds IEEE single range")
	}
	return math.Float32frombits((uint32(a.Exp)+127)<<23 | (uint32(a.Mnt)&0x7fff)<<8)
}
func xSat_FloatFromFloat(a float32) xFloat_t {
	if !(a >= 1) {
		panic("xSat_FloatFromFloat: input must be at least one")
	}
	bits := math.Float32bits(a)
	return xSat_FloatCreate((bits>>23)-127, 0x8000|((bits>>8)&0x7fff))
}
func xSat_FloatAdd(a, b xFloat_t) xFloat_t {
	if a.Exp < b.Exp {
		a, b = b, a
	}
	mnt := uint32(a.Mnt) + (uint32(b.Mnt) >> uint32(a.Exp-b.Exp))
	return xSat_FloatNormalize(uint32(a.Exp), mnt)
}
func xSat_FloatMul(a, b xFloat_t) xFloat_t {
	mnt := (uint32(a.Mnt) * uint32(b.Mnt)) >> 15
	return xSat_FloatNormalize(uint32(a.Exp)+uint32(b.Exp), mnt)
}
func xSat_FloatNormalize(exp, mnt uint32) xFloat_t {
	if mnt&0xffff0000 != 0 {
		exp++
		mnt >>= 1
	}
	if exp&0xffff0000 != 0 {
		return xSat_FloatCreateMaximum()
	}
	if mnt&0x8000 == 0 {
		panic("xSat float: unnormalized mantissa")
	}
	return xSat_FloatCreate(exp, mnt)
}
func xSat_FloatDiv(a xFloat_t, degree uint32) xFloat_t {
	if degree >= 0xffff {
		panic("xSat_FloatDiv: degree exceeds range")
	}
	if uint32(a.Exp) >= degree {
		return xSat_FloatCreate(uint32(a.Exp)-degree, uint32(a.Mnt))
	}
	return xSat_FloatCreateConst1()
}
func xSat_FloatTest() {
	for _, x := range []float32{1, 2, 3, 12, 1.5, 2.5} {
		fmt.Printf("%g = %f\n", x, xSat_Float2Float(xSat_FloatFromFloat(x)))
	}
	a, b := xSat_FloatFromFloat(1.5), xSat_FloatFromFloat(2.5)
	fmt.Printf("1.5 + 2.5 = %f\n", xSat_Float2Float(xSat_FloatAdd(a, b)))
	fmt.Printf("1.5 * 2.5 = %f\n", xSat_Float2Float(xSat_FloatMul(a, b)))
}
