// Copyright Satoko contributors. Distributed under the BSD 2-Clause License.
package utils

import "math"

type sdbl_t = uint64

const SDBL_CONST1 uint64 = 0x0000800000000000
const SDBL_MAX uint64 = ^uint64(0)

func sdbl_exp(a uint64) uint64 { return a >> 48 }
func sdbl_mnt(a uint64) uint64 { return a & 0xffffffffffff }
func sdbl2double(a uint64) float64 {
	if sdbl_exp(a) >= 1023 {
		panic("sdbl exponent exceeds double")
	}
	return math.Float64frombits(((sdbl_exp(a) + 1023) << 52) | ((a & 0x7fffffffffff) << 5))
}
func double2sdbl(v float64) uint64 {
	if v < 1 || math.IsNaN(v) || math.IsInf(v, 0) {
		panic("sdbl requires finite value >=1")
	}
	bits := math.Float64bits(v)
	return (((bits >> 52) - 1023) << 48) | SDBL_CONST1 | ((bits << 12) >> 17)
}
func sdbl_add(a, b uint64) uint64 {
	if a < b {
		a, b = b, a
	}
	e := sdbl_exp(a)
	m := sdbl_mnt(a) + (sdbl_mnt(b) >> (sdbl_exp(a) - sdbl_exp(b)))
	if m>>48 != 0 {
		e++
		m >>= 1
	}
	if e>>16 != 0 {
		return SDBL_MAX
	}
	return e<<48 | m
}
func sdbl_mult(a, b uint64) uint64 {
	am, bm := sdbl_mnt(a), sdbl_mnt(b)
	ah, al, bh, bl := am>>32, am&0xffffffff, bm>>32, bm&0xffffffff
	m := (ah*bh)<<17 + (al*bl)>>47 + (al*bh)>>15 + (ah*bl)>>15
	e := sdbl_exp(a) + sdbl_exp(b)
	if m>>48 != 0 {
		e++
		m >>= 1
	}
	if e>>16 != 0 {
		return SDBL_MAX
	}
	return e<<48 | m
}
func sdbl_div(a uint64, n uint32) uint64 {
	if sdbl_exp(a) < uint64(n) {
		return SDBL_CONST1
	}
	return (sdbl_exp(a)-uint64(n))<<48 | sdbl_mnt(a)
}
func ActivityFromFloat(v float64) uint64        { return double2sdbl(v) }
func ActivityFloat(v uint64) float64            { return sdbl2double(v) }
func ActivityAdd(a, b uint64) uint64            { return sdbl_add(a, b) }
func ActivityMultiply(a, b uint64) uint64       { return sdbl_mult(a, b) }
func ActivityRescale(a uint64, n uint32) uint64 { return sdbl_div(a, n) }
