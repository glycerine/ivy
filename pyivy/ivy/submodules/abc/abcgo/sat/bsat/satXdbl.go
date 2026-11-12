package bsat

import "math"

// ABC's positive extended-range activity format: a 16-bit exponent and a
// 48-bit significand (including its leading one). Integer order is value order.
type xdbl = uint64

func Xdbl_Exp(a xdbl) uint64 { return a >> 48 }
func Xdbl_Mnt(a xdbl) uint64 { return a & ((1 << 48) - 1) }
func Xdbl_Create(exp, mantissa uint64) xdbl {
	if exp>>16 != 0 || mantissa>>47 != 1 {
		panic("invalid extended activity")
	}
	return exp<<48 | mantissa
}
func Xdbl_Const1() xdbl  { return Xdbl_Create(0, 1<<47) }
func Xdbl_Maximum() xdbl { return ^uint64(0) }
func Xdbl_FromDouble(a float64) xdbl {
	if math.IsInf(a, 0) || math.IsNaN(a) || a < 1 {
		panic("extended activity requires finite value >= 1")
	}
	bits := math.Float64bits(a)
	return Xdbl_Create((bits>>52)-1023, (1<<47)|((bits<<12)>>17))
}
func Xdbl_ToDouble(a xdbl) float64 {
	if Xdbl_Exp(a) >= 1023 {
		panic("extended activity exceeds float64 range")
	}
	return math.Float64frombits((Xdbl_Exp(a)+1023)<<52 | ((a<<17)>>17)<<5)
}
func Xdbl_Add(a, b xdbl) xdbl {
	if a < b {
		a, b = b, a
	}
	exp := Xdbl_Exp(a)
	mantissa := Xdbl_Mnt(a) + (Xdbl_Mnt(b) >> (exp - Xdbl_Exp(b)))
	if mantissa>>48 != 0 {
		exp++
		mantissa >>= 1
	}
	if exp>>16 != 0 {
		return Xdbl_Maximum()
	}
	return Xdbl_Create(exp, mantissa)
}
func Xdbl_Mul(a, b xdbl) xdbl {
	if a < b {
		a, b = b, a
	}
	ma, mb := Xdbl_Mnt(a), Xdbl_Mnt(b)
	ah, bh := ma>>32, mb>>32
	al, bl := ma&0xffffffff, mb&0xffffffff
	mantissa := ((ah * bh) << 17) + ((al * bl) >> 47) + ((al * bh) >> 15) + ((ah * bl) >> 15)
	exp := Xdbl_Exp(a) + Xdbl_Exp(b)
	if mantissa>>48 != 0 {
		exp++
		mantissa >>= 1
	}
	if exp>>16 != 0 {
		return Xdbl_Maximum()
	}
	return Xdbl_Create(exp, mantissa)
}
func Xdbl_Div(a xdbl, power uint32) xdbl {
	if Xdbl_Exp(a) >= uint64(power) {
		return Xdbl_Create(Xdbl_Exp(a)-uint64(power), Xdbl_Mnt(a))
	}
	return Xdbl_Const1()
}
