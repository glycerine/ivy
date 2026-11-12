package gia

import "github.com/glycerine/abcgo/misc/vec"

// Shr_ManComputeTruths simulates divisors in topological order over six variables.
func Shr_ManComputeTruths(p *Gia_Man_t, nVars int32, divs *Vec_Int_t, divTruths, truths *Vec_Wrd_t) {
	patterns := [8]word{0xaaaaaaaaaaaaaaaa, 0xcccccccccccccccc, 0xf0f0f0f0f0f0f0f0, 0xff00ff00ff00ff00, 0xffff0000ffff0000, 0xffffffff00000000, 0, 0xffffffffffffffff}
	data := vec.Vec_IntArray(divs)
	if nVars < 0 || nVars > 6 || int(nVars) > len(data) {
		panic("gia: invalid shrink divisor inputs")
	}
	vec.Vec_WrdClear(divTruths)
	for i, id := range data {
		if id < 0 || int(id) >= vec.Vec_WrdSize(truths) {
			panic("gia: invalid shrink divisor")
		}
		var value word
		if i < int(nVars) {
			value = patterns[i]
		} else {
			obj := Gia_ManObj(p, id)
			if Gia_ObjIsAnd(obj) == 0 {
				panic("gia: shrink divisor is not an AND")
			}
			a, b := vec.Vec_WrdEntry(truths, int(Gia_ObjFaninId0(obj, id))), vec.Vec_WrdEntry(truths, int(Gia_ObjFaninId1(obj, id)))
			if Gia_ObjFaninC0(obj) != 0 {
				a = ^a
			}
			if Gia_ObjFaninC1(obj) != 0 {
				b = ^b
			}
			value = a & b
		}
		vec.Vec_WrdWriteEntry(truths, int(id), value)
		vec.Vec_WrdPush(divTruths, value)
	}
}
