package bmc

import (
	"github.com/glycerine/abcgo/aig/gia"
	"github.com/glycerine/abcgo/misc/util"
	"github.com/glycerine/abcgo/misc/vec"
)

// Bmc_GiaGenerateGiaOne creates the cube of justified register values at end
// and returns the counterexample register state at begin.
func Bmc_GiaGenerateGiaOne(p *gia.Gia_Man_t, cex *util.Abc_Cex_t, initial **vec.Vec_Bit_t, begin, end int32) *gia.Gia_Man_t {
	if p == nil || cex == nil || initial == nil || begin < 0 || end < begin || end > cex.IFrame {
		panic("bmc: invalid target frame range")
	}
	var values, just *vec.Vec_Bit_t
	Bmc_GiaGenerateJust(p, cex, &values, &just)
	n, regs := gia.Gia_ManObjNum(p), gia.Gia_ManRegNum(p)
	*initial = vec.Vec_BitStart(int(regs))
	for i := int32(0); i < regs; i++ {
		id := gia.Gia_ObjId(p, gia.Gia_ManRo(p, i))
		if vec.Vec_BitEntry(values, int(n*begin+id)) != 0 {
			vec.Vec_BitWriteEntry(*initial, int(i), 1)
		}
	}
	q := gia.Gia_ManStart(2*regs + 2)
	cube := int32(1)
	for i := int32(0); i < regs; i++ {
		lit := gia.Gia_ManAppendCi(q)
		id := gia.Gia_ObjId(p, gia.Gia_ManRo(p, i))
		if vec.Vec_BitEntry(just, int(n*end+id)) == 0 {
			continue
		}
		if vec.Vec_BitEntry(values, int(n*end+id)) == 0 {
			lit ^= 1
		}
		cube = gia.Gia_ManAppendAnd(q, cube, lit)
	}
	gia.Gia_ManAppendCo(q, cube)
	return q
}
