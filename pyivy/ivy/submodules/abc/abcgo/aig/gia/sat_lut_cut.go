package gia

import (
	"fmt"
	"github.com/glycerine/abcgo/misc/vec"
	"math/bits"
)

func Sbl_ManPrintCut(i1, i2, n1, n2 word) int32 {
	fmt.Print("{ ")
	count := int32(0)
	for j, v := range []word{i1, i2, n1, n2} {
		if j == 2 {
			fmt.Print(" ")
		}
		for b := 0; b < 64; b++ {
			if v&(word(1)<<uint(b)) != 0 {
				prefix := 'i'
				if j >= 2 {
					prefix = 'n'
				}
				fmt.Printf("%c%d ", prefix, (j%2)*64+b)
				count++
			}
		}
	}
	fmt.Println("};")
	return count
}
func Sbl_CutIsFeasible(i1, i2, n1, n2 word, size int32) int32 {
	if size < 1 || size > 6 {
		panic("gia: invalid SAT LUT size")
	}
	// The original uses a four-input limit for sizes up to four, six otherwise.
	limit := 4
	if size > 4 {
		limit = 6
	}
	count := bits.OnesCount64(i1) + bits.OnesCount64(i2) + bits.OnesCount64(n1) + bits.OnesCount64(n2)
	if count <= limit {
		return 1
	}
	return 0
}
func Sbl_CutPushUncontained(i1, i2, n1, n2 *Vec_Wrd_t, a, b, c, d word) int32 {
	vectors := [4]*Vec_Wrd_t{i1, i2, n1, n2}
	candidate := [4]word{a, b, c, d}
	size := vec.Vec_WrdSize(i1)
	for _, v := range vectors {
		if vec.Vec_WrdSize(v) != size {
			panic("gia: inconsistent SAT LUT cut vectors")
		}
	}
	for i := 0; i < size; i++ {
		contained := true
		for j, v := range vectors {
			x := vec.Vec_WrdEntry(v, i)
			if x&candidate[j] != x {
				contained = false
				break
			}
		}
		if contained {
			return 1
		}
	}
	keep := 0
	for i := 0; i < size; i++ {
		superset := true
		for j, v := range vectors {
			if vec.Vec_WrdEntry(v, i)&candidate[j] != candidate[j] {
				superset = false
				break
			}
		}
		if !superset {
			for _, v := range vectors {
				vec.Vec_WrdWriteEntry(v, keep, vec.Vec_WrdEntry(v, i))
			}
			keep++
		}
	}
	for j, v := range vectors {
		vec.Vec_WrdShrink(v, keep)
		vec.Vec_WrdPush(v, candidate[j])
	}
	return 0
}
