package abs

import (
	"math/bits"

	"github.com/glycerine/abcgo/misc/vec"
)

// Rf2_ManProcessVector retains the inclusion-minimal justification masks with
// no more than limit set bits. The refinement algorithm uses at most 30 bits.
func Rf2_ManProcessVector(p *vec.Vec_Int_t, limit int32) {
	write := 0
	for _, entry := range vec.Vec_IntArray(p) {
		if int32(bits.OnesCount32(uint32(entry))) <= limit {
			vec.Vec_IntWriteEntry(p, write, entry)
			write++
		}
	}
	vec.Vec_IntShrink(p, write)
	vec.Vec_IntSort(p, 0)
	write = 0
	for _, entry := range vec.Vec_IntArray(p) {
		contained := false
		for _, prior := range vec.Vec_IntArray(p)[:write] {
			if prior&entry == prior {
				contained = true
				break
			}
		}
		if !contained {
			vec.Vec_IntWriteEntry(p, write, entry)
			write++
		}
	}
	vec.Vec_IntShrink(p, write)
}
