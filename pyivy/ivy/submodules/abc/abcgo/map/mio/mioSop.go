// Subsumption-free sum-of-products operations from mioSop.c.
package mio

import (
	"github.com/glycerine/abcgo/misc/vec"
	"strings"
)

func Mio_CubeVar0(v int32) uint32             { return 1 << uint(v*2) }
func Mio_CubeVar1(v int32) uint32             { return 1 << uint(v*2+1) }
func Mio_CubeHasVar0(c uint32, v int32) int32 { return mioBool(c&Mio_CubeVar0(v) != 0) }
func Mio_CubeHasVar1(c uint32, v int32) int32 { return mioBool(c&Mio_CubeVar1(v) != 0) }
func Mio_CubeEmpty(c uint32) int32            { return mioBool(c&(c>>1)&0x55555555 != 0) }
func Mio_CubeAnd(a, b uint32) uint32          { return a | b }
func Mio_CubeContains(a, b uint32) int32      { return mioBool(a|b == b) }
func Mio_SopPushSCC(p *Vec_Int_t, c uint32) {
	entries := vec.Vec_IntArray(p)
	for _, e := range entries {
		if Mio_CubeContains(uint32(e), c) != 0 {
			return
		}
	}
	next := 0
	for _, e := range entries {
		if Mio_CubeContains(c, uint32(e)) == 0 {
			entries[next] = e
			next++
		}
	}
	vec.Vec_IntShrink(p, next)
	vec.Vec_IntPush(p, int32(c))
}
func Mio_SopCoverOr(a, b *Vec_Int_t) *Vec_Int_t {
	r := vec.Vec_IntAllocArrayCopy(vec.Vec_IntArray(a), vec.Vec_IntSize(a))
	for _, c := range vec.Vec_IntArray(b) {
		Mio_SopPushSCC(r, uint32(c))
	}
	return r
}
func Mio_SopCoverAnd(a, b *Vec_Int_t) *Vec_Int_t {
	r := vec.Vec_IntAlloc(0)
	for _, x := range vec.Vec_IntArray(a) {
		for _, y := range vec.Vec_IntArray(b) {
			c := uint32(x) | uint32(y)
			if Mio_CubeEmpty(c) == 0 {
				Mio_SopPushSCC(r, c)
			}
		}
	}
	return r
}
func Mio_SopVar0(i int32) *Vec_Int_t { return expVector(int32(Mio_CubeVar0(i))) }
func Mio_SopVar1(i int32) *Vec_Int_t { return expVector(int32(Mio_CubeVar1(i))) }
func Mio_SopConst0() *Vec_Int_t      { return vec.Vec_IntAlloc(0) }
func Mio_SopConst1() *Vec_Int_t      { return expVector(0) }
func mioStoreSop(s string, buffer *Vec_Str_t) string {
	if buffer != nil {
		vec.Vec_StrClear(buffer)
		for i := range s {
			vec.Vec_StrPush(buffer, s[i])
		}
		vec.Vec_StrPush(buffer, 0)
	}
	return s
}
func Mio_SopDeriveFromArray(cover *Vec_Int_t, n int32, buffer *Vec_Str_t, polarity int32) string {
	cubes := vec.Vec_IntArray(cover)
	if len(cubes) == 0 {
		return mioStoreSop(string([]byte{' ', byte('1' - polarity), '\n'}), buffer)
	}
	if len(cubes) == 1 && cubes[0] == 0 {
		return mioStoreSop(string([]byte{' ', byte('0' + polarity), '\n'}), buffer)
	}
	var s strings.Builder
	for _, cube := range cubes {
		for v := int32(0); v < n; v++ {
			switch {
			case Mio_CubeHasVar0(uint32(cube), v) != 0:
				s.WriteByte('0')
			case Mio_CubeHasVar1(uint32(cube), v) != 0:
				s.WriteByte('1')
			default:
				s.WriteByte('-')
			}
		}
		s.WriteByte(' ')
		s.WriteByte(byte('0' + polarity))
		s.WriteByte('\n')
	}
	return mioStoreSop(s.String(), buffer)
}
func Mio_LibDeriveSop(n int32, expression *Vec_Int_t, buffer *Vec_Str_t) string {
	if n < 0 || n > 16 {
		panic("MIO compact SOP requires 0..16 variables")
	}
	a := vec.Vec_IntArray(expression)
	zero, one := make([]*Vec_Int_t, n), make([]*Vec_Int_t, n)
	for i := int32(0); i < n; i++ {
		zero[i] = Mio_SopVar0(i)
		one[i] = Mio_SopVar1(i)
	}
	literal := func(lit int32, positive bool) *Vec_Int_t {
		if lit == EXP_CONST0 {
			if positive {
				return Mio_SopConst0()
			}
			return Mio_SopConst1()
		}
		if lit == EXP_CONST1 {
			if positive {
				return Mio_SopConst1()
			}
			return Mio_SopConst0()
		}
		if (lit&1 != 0) == positive {
			return zero[lit/2]
		}
		return one[lit/2]
	}
	for i := 0; i < len(a)/2; i++ {
		left, right := a[2*i], a[2*i+1]
		p := Mio_SopCoverAnd(literal(left, true), literal(right, true))
		q := Mio_SopCoverOr(literal(left, false), literal(right, false))
		one = append(one, p)
		zero = append(zero, q)
	}
	positive, negative := literal(a[len(a)-1], true), literal(a[len(a)-1], false)
	if vec.Vec_IntSize(negative) < vec.Vec_IntSize(positive) {
		return Mio_SopDeriveFromArray(negative, n, buffer, 0)
	}
	return Mio_SopDeriveFromArray(positive, n, buffer, 1)
}
