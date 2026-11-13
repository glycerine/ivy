package abs

import (
	"github.com/glycerine/abcgo/aig/gia"
	"github.com/glycerine/abcgo/misc/vec"
)

// Gia_ManTransformFlops changes state coordinates for the selected initial bits.
func Gia_ManTransformFlops(p *gia.Gia_Man_t, flops, initial *vec.Vec_Int_t) *gia.Gia_Man_t {
	if vec.Vec_IntSize(flops) != vec.Vec_IntSize(initial) {
		panic("abs: initial state size mismatch")
	}
	bits := make([]int32, (gia.Gia_ManRegNum(p)+31)/32)
	for i := 0; i < vec.Vec_IntSize(flops); i++ {
		obj := gia.Gia_ManObj(p, vec.Vec_IntEntry(flops, i))
		if gia.Gia_ObjIsRo(p, obj) == 0 {
			panic("abs: expected register output")
		}
		if vec.Vec_IntEntry(initial, i) == 0 {
			continue
		}
		index := gia.Gia_ObjCioId(obj) - gia.Gia_ManPiNum(p)
		if index < 0 || index >= gia.Gia_ManRegNum(p) {
			panic("abs: invalid register index")
		}
		bits[index>>5] |= int32(uint32(1) << uint(index&31))
	}
	return gia.Gia_ManDupFlip(p, bits)
}
