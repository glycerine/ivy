package util

// Extended positive floating-point arithmetic from ABC utilDouble.h, by
// Alan Mishchenko and Bruno Schmitt. The packed word holds a 16-bit exponent
// above a normalized 48-bit mantissa. Multiplication deliberately preserves
// the original four-product truncation rather than substituting IEEE rounding.
import (
	"fmt"
	"math"
)

type xdbl = uint64

func Xdbl_Exp(a xdbl) uint64 { return a >> 48 }
func Xdbl_Mnt(a xdbl) uint64 { return a & ((1 << 48) - 1) }
func Xdbl_Create(exp, mnt uint64) xdbl {
	if exp>>16 != 0 || mnt>>47 != 1 {
		panic("Xdbl_Create: exponent or mantissa out of range")
	}
	return exp<<48 | mnt
}
func Xdbl_Const1() xdbl       { return Xdbl_Create(0, 1<<47) }
func Xdbl_Const2() xdbl       { return Xdbl_Create(1, 1<<47) }
func Xdbl_Const3() xdbl       { return Xdbl_Create(1, 3<<46) }
func Xdbl_Const12() xdbl      { return Xdbl_Create(3, 3<<46) }
func Xdbl_Const1point5() xdbl { return Xdbl_Create(0, 3<<46) }
func Xdbl_Const2point5() xdbl { return Xdbl_Create(1, 5<<45) }
func Xdbl_Maximum() xdbl      { return math.MaxUint64 }
func Xdbl_ToDouble(a xdbl) float64 {
	if Xdbl_Exp(a) >= 1023 {
		panic("Xdbl_ToDouble: exponent exceeds IEEE double range")
	}
	return math.Float64frombits((Xdbl_Exp(a)+1023)<<52 | (a&((1<<47)-1))<<5)
}
func Xdbl_FromDouble(a float64) xdbl {
	if !(a >= 1) {
		panic("Xdbl_FromDouble: input must be at least one")
	}
	bits := math.Float64bits(a)
	return Xdbl_Create((bits>>52)-1023, 1<<47|((bits<<12)>>17))
}
func Xdbl_Add(a, b xdbl) xdbl {
	if a < b {
		a, b = b, a
	}
	mnt := Xdbl_Mnt(a) + (Xdbl_Mnt(b) >> (Xdbl_Exp(a) - Xdbl_Exp(b)))
	return xdblNormalize(Xdbl_Exp(a), mnt)
}
func Xdbl_Mul(a, b xdbl) xdbl {
	ma, mb := Xdbl_Mnt(a), Xdbl_Mnt(b)
	ah, bh, al, bl := ma>>32, mb>>32, ma&0xffffffff, mb&0xffffffff
	mnt := ((ah * bh) << 17) + ((al * bl) >> 47) + ((al * bh) >> 15) + ((ah * bl) >> 15)
	return xdblNormalize(Xdbl_Exp(a)+Xdbl_Exp(b), mnt)
}
func xdblNormalize(exp, mnt uint64) xdbl {
	if mnt>>48 != 0 {
		exp++
		mnt >>= 1
	}
	if exp>>16 != 0 {
		return Xdbl_Maximum()
	}
	return Xdbl_Create(exp, mnt)
}
func Xdbl_Div(a xdbl, degree uint32) xdbl {
	if Xdbl_Exp(a) >= uint64(degree) {
		return Xdbl_Create(Xdbl_Exp(a)-uint64(degree), Xdbl_Mnt(a))
	}
	return Xdbl_Const1()
}
func Xdbl_Test() {
	for _, v := range []float64{1, 2, 3, 12, 1.5, 2.5} {
		fmt.Printf("%g = %f\n", v, Xdbl_ToDouble(Xdbl_FromDouble(v)))
	}
	fmt.Printf("1.5 + 2.5 = %f\n", Xdbl_ToDouble(Xdbl_Add(Xdbl_Const1point5(), Xdbl_Const2point5())))
	fmt.Printf("1.5 * 2.5 = %f\n", Xdbl_ToDouble(Xdbl_Mul(Xdbl_Const1point5(), Xdbl_Const2point5())))
}
