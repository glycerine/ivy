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
