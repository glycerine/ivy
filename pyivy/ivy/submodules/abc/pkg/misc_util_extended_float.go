package abc

// Packed positive floating-point arithmetic from misc/util/utilFloat.h and utilDouble.h.

import "math"

type ExtendedFloat32 uint32
type ExtendedFloat64 uint64

const (
	ExtendedFloat32One     ExtendedFloat32 = 0x00008000
	ExtendedFloat32Two     ExtendedFloat32 = 0x00018000
	ExtendedFloat32Three   ExtendedFloat32 = 0x0001c000
	ExtendedFloat32Twelve  ExtendedFloat32 = 0x0003c000
	ExtendedFloat32OneHalf ExtendedFloat32 = 0x0000c000
	ExtendedFloat32TwoHalf ExtendedFloat32 = 0x0001a000
	ExtendedFloat32Maximum ExtendedFloat32 = 0xffffffff
	ExtendedFloat64One     ExtendedFloat64 = 0x0000800000000000
	ExtendedFloat64Two     ExtendedFloat64 = 0x0001800000000000
	ExtendedFloat64Three   ExtendedFloat64 = 0x0001c00000000000
	ExtendedFloat64Twelve  ExtendedFloat64 = 0x0003c00000000000
	ExtendedFloat64OneHalf ExtendedFloat64 = 0x0000c00000000000
	ExtendedFloat64TwoHalf ExtendedFloat64 = 0x0001a00000000000
	ExtendedFloat64Maximum ExtendedFloat64 = 0xffffffffffffffff
)

func NewExtendedFloat32(exponent, mantissa uint32) (ExtendedFloat32, error) {
	if exponent > 0xffff || mantissa < 0x8000 || mantissa > 0xffff {
		return 0, ErrDimension
	}
	return ExtendedFloat32(exponent<<16 | mantissa), nil
}

func ExtendedFloat32FromBits(bits uint32) (ExtendedFloat32, error) {
	return NewExtendedFloat32(bits>>16, bits&0xffff)
}

func (number ExtendedFloat32) Bits() uint32     { return uint32(number) }
func (number ExtendedFloat32) Exponent() uint32 { return uint32(number) >> 16 }
func (number ExtendedFloat32) Mantissa() uint32 { return uint32(number) & 0xffff }

func (number ExtendedFloat32) valid() bool { return number.Mantissa() >= 0x8000 }

func (number ExtendedFloat32) Less(other ExtendedFloat32) (bool, error) {
	if !number.valid() || !other.valid() {
		return false, ErrDimension
	}
	return number < other, nil
}

func (number ExtendedFloat32) Float32() (float32, error) {
	if !number.valid() || number.Exponent() >= 127 {
		return 0, ErrDimension
	}
	bits := (number.Exponent()+127)<<23 | (number.Mantissa()&0x7fff)<<8
	return math.Float32frombits(bits), nil
}

func ExtendedFloat32FromFloat(value float32) (ExtendedFloat32, error) {
	bits := math.Float32bits(value)
	exponent := (bits >> 23) & 0xff
	if value < 1 || exponent == 0xff {
		return 0, ErrDimension
	}
	return NewExtendedFloat32(exponent-127, 0x8000|(bits>>8)&0x7fff)
}

func (number ExtendedFloat32) Add(other ExtendedFloat32) (ExtendedFloat32, error) {
	if !number.valid() || !other.valid() {
		return 0, ErrDimension
	}
	if number < other {
		return other.Add(number)
	}
	difference := number.Exponent() - other.Exponent()
	aligned := uint32(0)
	if difference < 32 {
		aligned = other.Mantissa() >> difference
	}
	mantissa := number.Mantissa() + aligned
	exponent := number.Exponent()
	if mantissa&0xffff0000 != 0 {
		exponent++
		mantissa >>= 1
	}
	if exponent > 0xffff {
		return ExtendedFloat32Maximum, nil
	}
	return NewExtendedFloat32(exponent, mantissa)
}

func (number ExtendedFloat32) Multiply(other ExtendedFloat32) (ExtendedFloat32, error) {
	if !number.valid() || !other.valid() {
		return 0, ErrDimension
	}
	if number < other {
		return other.Multiply(number)
	}
	mantissa := number.Mantissa() * other.Mantissa() >> 15
	exponent := number.Exponent() + other.Exponent()
	if mantissa&0xffff0000 != 0 {
		exponent++
		mantissa >>= 1
	}
	if exponent > 0xffff {
		return ExtendedFloat32Maximum, nil
	}
	return NewExtendedFloat32(exponent, mantissa)
}

func (number ExtendedFloat32) DividePowerOfTwo(degree uint32) (ExtendedFloat32, error) {
	if !number.valid() || degree >= 0xffff {
		return 0, ErrDimension
	}
	if number.Exponent() < degree {
		return ExtendedFloat32One, nil
	}
	return NewExtendedFloat32(number.Exponent()-degree, number.Mantissa())
}

func NewExtendedFloat64(exponent, mantissa uint64) (ExtendedFloat64, error) {
	if exponent > 0xffff || mantissa < 1<<47 || mantissa > (1<<48)-1 {
		return 0, ErrDimension
	}
	return ExtendedFloat64(exponent<<48 | mantissa), nil
}

func ExtendedFloat64FromBits(bits uint64) (ExtendedFloat64, error) {
	return NewExtendedFloat64(bits>>48, bits&((uint64(1)<<48)-1))
}

func (number ExtendedFloat64) Bits() uint64     { return uint64(number) }
func (number ExtendedFloat64) Exponent() uint64 { return uint64(number) >> 48 }
func (number ExtendedFloat64) Mantissa() uint64 { return uint64(number) & ((uint64(1) << 48) - 1) }

func (number ExtendedFloat64) valid() bool { return number.Mantissa() >= 1<<47 }

func (number ExtendedFloat64) Less(other ExtendedFloat64) (bool, error) {
	if !number.valid() || !other.valid() {
		return false, ErrDimension
	}
	return number < other, nil
}

func (number ExtendedFloat64) Float64() (float64, error) {
	if !number.valid() || number.Exponent() >= 1023 {
		return 0, ErrDimension
	}
	bits := (number.Exponent()+1023)<<52 | (number.Mantissa()&((uint64(1)<<47)-1))<<5
	return math.Float64frombits(bits), nil
}

func ExtendedFloat64FromFloat(value float64) (ExtendedFloat64, error) {
	bits := math.Float64bits(value)
	exponent := (bits >> 52) & 0x7ff
	if value < 1 || exponent == 0x7ff {
		return 0, ErrDimension
	}
	return NewExtendedFloat64(exponent-1023, 1<<47|(bits>>5)&((uint64(1)<<47)-1))
}

func (number ExtendedFloat64) Add(other ExtendedFloat64) (ExtendedFloat64, error) {
	if !number.valid() || !other.valid() {
		return 0, ErrDimension
	}
	if number < other {
		return other.Add(number)
	}
	difference := number.Exponent() - other.Exponent()
	aligned := uint64(0)
	if difference < 64 {
		aligned = other.Mantissa() >> difference
	}
	mantissa := number.Mantissa() + aligned
	exponent := number.Exponent()
	if mantissa>>48 != 0 {
		exponent++
		mantissa >>= 1
	}
	if exponent > 0xffff {
		return ExtendedFloat64Maximum, nil
	}
	return NewExtendedFloat64(exponent, mantissa)
}

func (number ExtendedFloat64) Multiply(other ExtendedFloat64) (ExtendedFloat64, error) {
	if !number.valid() || !other.valid() {
		return 0, ErrDimension
	}
	if number < other {
		return other.Multiply(number)
	}
	a, b := number.Mantissa(), other.Mantissa()
	ah, bh, al, bl := a>>32, b>>32, uint32(a), uint32(b)
	mantissa := (ah*bh)<<17 + (uint64(al) * uint64(bl) >> 47) + (uint64(al) * bh >> 15) + (ah * uint64(bl) >> 15)
	exponent := number.Exponent() + other.Exponent()
	if mantissa>>48 != 0 {
		exponent++
		mantissa >>= 1
	}
	if exponent > 0xffff {
		return ExtendedFloat64Maximum, nil
	}
	return NewExtendedFloat64(exponent, mantissa)
}

func (number ExtendedFloat64) DividePowerOfTwo(degree uint32) (ExtendedFloat64, error) {
	if !number.valid() {
		return 0, ErrDimension
	}
	if number.Exponent() < uint64(degree) {
		return ExtendedFloat64One, nil
	}
	return NewExtendedFloat64(number.Exponent()-uint64(degree), number.Mantissa())
}
