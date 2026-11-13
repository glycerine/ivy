package abs

import (
	"github.com/glycerine/abcgo/aig/gia"
	"github.com/glycerine/abcgo/misc/vec"
)

// Gia_ManClasses2Flops converts the nonzero entries of a flop map to IDs.
func Gia_ManClasses2Flops(classes *vec.Vec_Int_t) *vec.Vec_Int_t {
	flops := vec.Vec_IntAlloc(vec.Vec_IntSize(classes))
	for id, class := range vec.Vec_IntArray(classes) {
		if class != 0 {
			vec.Vec_IntPush(flops, int32(id))
		}
	}
	return flops
}

// Gia_ManFlops2Classes converts a set of flop IDs to a Boolean map.
func Gia_ManFlops2Classes(graph *gia.Gia_Man_t, flops *vec.Vec_Int_t) *vec.Vec_Int_t {
	classes := vec.Vec_IntStart(int(gia.Gia_ManRegNum(graph)))
	for _, id := range vec.Vec_IntArray(flops) {
		vec.Vec_IntWriteEntry(classes, int(id), 1)
	}
	return classes
}
