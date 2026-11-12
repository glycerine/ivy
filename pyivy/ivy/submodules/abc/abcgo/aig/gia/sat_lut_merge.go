package gia

import "github.com/glycerine/abcgo/misc/vec"

// sblMergeCuts enumerates the minimal feasible unions of two cut lists.
// Each list uses four aligned words per cut: two input and two node masks.
func sblMergeCuts(left, right [4][]word, size int32) [4]*Vec_Wrd_t {
	for i := 1; i < 4; i++ {
		if len(left[i]) != len(left[0]) || len(right[i]) != len(right[0]) {
			panic("gia: inconsistent SAT LUT cut lists")
		}
	}
	out := [4]*Vec_Wrd_t{vec.Vec_WrdAlloc(32), vec.Vec_WrdAlloc(32), vec.Vec_WrdAlloc(32), vec.Vec_WrdAlloc(32)}
	for i := range left[0] {
		for j := range right[0] {
			a, b, c, d := left[0][i]|right[0][j], left[1][i]|right[1][j], left[2][i]|right[2][j], left[3][i]|right[3][j]
			if Sbl_CutIsFeasible(a, b, c, d, size) != 0 {
				Sbl_CutPushUncontained(out[0], out[1], out[2], out[3], a, b, c, d)
			}
		}
	}
	return out
}

func Sbl_ManFindAndPrintCut(p *Sbl_Man_t, index int32) int32 {
	return Sbl_ManPrintCut(vec.Vec_WrdEntry(p.vCutsI1, int(index)), vec.Vec_WrdEntry(p.vCutsI2, int(index)), vec.Vec_WrdEntry(p.vCutsN1, int(index)), vec.Vec_WrdEntry(p.vCutsN2, int(index)))
}
func Sbl_ManComputeCutsOne(p *Sbl_Man_t, a, b, obj int32) {
	if obj < 0 || obj >= 128 {
		panic("gia: invalid SAT window node")
	}
	vectors := [4]*Vec_Wrd_t{p.vCutsI1, p.vCutsI2, p.vCutsN1, p.vCutsN2}
	startA, nA := vec.Vec_IntEntry(p.vCutsStart, int(a)), vec.Vec_IntEntry(p.vCutsNum, int(a))
	startB, nB := vec.Vec_IntEntry(p.vCutsStart, int(b)), vec.Vec_IntEntry(p.vCutsNum, int(b))
	var left, right [4][]word
	for i, v := range vectors {
		data := vec.Vec_WrdArray(v)
		left[i] = data[startA : startA+nA]
		right[i] = data[startB : startB+nB]
	}
	merged := sblMergeCuts(left, right, p.LutSize)
	n := vec.Vec_WrdSize(merged[0])
	vec.Vec_IntPush(p.vCutsStart, int32(vec.Vec_WrdSize(p.vCutsI1)))
	vec.Vec_IntPush(p.vCutsNum, int32(n+1))
	for i, v := range vectors {
		for _, x := range vec.Vec_WrdArray(merged[i]) {
			vec.Vec_WrdPush(v, x)
		}
		var unit word
		if i == 2 && obj < 64 {
			unit = word(1) << uint(obj)
		}
		if i == 3 && obj >= 64 {
			unit = word(1) << uint(obj-64)
		}
		vec.Vec_WrdPush(v, unit)
	}
	for i := 0; i <= n; i++ {
		vec.Vec_IntPush(p.vCutsObj, obj)
	}
}
func Sbl_ManFindCut(p *Sbl_Man_t, obj int32, a, b, c, d word) int32 {
	start := vec.Vec_IntEntry(p.vCutsStart, int(obj))
	end := start + vec.Vec_IntEntry(p.vCutsNum, int(obj))
	for i := start; i < end; i++ {
		if vec.Vec_WrdEntry(p.vCutsI1, int(i)) == a && vec.Vec_WrdEntry(p.vCutsI2, int(i)) == b && vec.Vec_WrdEntry(p.vCutsN1, int(i)) == c && vec.Vec_WrdEntry(p.vCutsN2, int(i)) == d {
			return i
		}
	}
	return -1
}
