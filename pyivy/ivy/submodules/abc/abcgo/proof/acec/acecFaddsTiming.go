package acec

// Timing-box construction from src/proof/acec/acecFadds.c.
import (
	"github.com/glycerine/abcgo/aig/gia"
	"github.com/glycerine/abcgo/misc/tim"
	"github.com/glycerine/abcgo/misc/util"
)

func Gia_ManGenerateDelayTableFloat(ins, outs int32) []float32 {
	total := int(ins * outs)
	table := make([]float32, total+3)
	table[1], table[2] = float32(ins), float32(outs)
	for i := 3; i < len(table); i++ {
		table[i] = 1
	}
	table[total+3-int(ins)] = -float32(util.ABC_INFINITY)
	return table
}
func Gia_ManGenerateTim(pis, pos, boxes, ins, outs int32) *tim.Tim_Man_t {
	tables := [][]float32{Gia_ManGenerateDelayTableFloat(ins, outs)}
	p := tim.Tim_ManStart(int(pis+outs*boxes), int(pos+ins*boxes))
	tim.Tim_ManSetDelayTables(p, tables)
	ci, co := int(pis), 0
	for i := int32(0); i < boxes; i++ {
		tim.Tim_ManCreateBox(p, co, int(ins), ci, int(outs), 0, 0)
		ci += int(outs)
		co += int(ins)
	}
	co += int(pos)
	if ci != tim.Tim_ManCiNum(p) || co != tim.Tim_ManCoNum(p) {
		panic("acec: timing-box size mismatch")
	}
	return p
}
func Gia_ManGenerateExtraAig(boxes, ins, outs int32) *Gia_Man_t {
	if ins >= 16 || outs >= 16 {
		panic("acec: timing box too large")
	}
	capacity := boxes * 20
	if capacity < 1 {
		capacity = 1
	}
	p := gia.Gia_ManStart(capacity)
	var in, out [16]int32
	for i := int32(0); i < ins; i++ {
		in[i] = gia.Gia_ManAppendCi(p)
	}
	out[0] = gia.Gia_ManAppendXor(p, gia.Gia_ManAppendXor(p, in[0], in[1]), in[2])
	out[1] = gia.Gia_ManAppendMaj(p, in[0], in[1], in[2])
	for i := int32(0); i < boxes; i++ {
		for k := int32(0); k < outs; k++ {
			gia.Gia_ManAppendCo(p, out[k])
		}
	}
	return p
}
