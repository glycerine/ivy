package acec

// Adder and XOR root collection from src/proof/acec/acecPool.c.
import (
	"github.com/glycerine/abcgo/aig/gia"
	"github.com/glycerine/abcgo/misc/vec"
)

func Acec_ManCreateCarryMap(p *Gia_Man_t, adds *Vec_Int_t) *Vec_Int_t {
	m := vec.Vec_IntStartFull(int(gia.Gia_ManObjNum(p)))
	for i := 0; 6*i < vec.Vec_IntSize(adds); i++ {
		vec.Vec_IntWriteEntry(m, int(vec.Vec_IntEntry(adds, 6*i+4)), int32(i))
	}
	return m
}
func Acec_ManCheckCarryMap(p *Gia_Man_t, carry int32, adds, m *Vec_Int_t) int32 {
	box := vec.Vec_IntEntry(m, int(carry))
	if box < 0 {
		panic("acec: missing carry map")
	}
	for k := 0; k < 3; k++ {
		if vec.Vec_IntEntry(m, int(vec.Vec_IntEntry(adds, int(box)*6+k))) >= 0 {
			return 1
		}
	}
	return 0
}
func Acec_ManCollectCarryRoots(p *Gia_Man_t, adds *Vec_Int_t) *Vec_Int_t {
	roots := vec.Vec_IntAlloc(100)
	ins := vec.Vec_BitStart(int(gia.Gia_ManObjNum(p)))
	for i := 0; 6*i < vec.Vec_IntSize(adds); i++ {
		for k := 0; k < 3; k++ {
			vec.Vec_BitWriteEntry(ins, int(vec.Vec_IntEntry(adds, 6*i+k)), 1)
		}
	}
	for i := 0; 6*i < vec.Vec_IntSize(adds); i++ {
		id := vec.Vec_IntEntry(adds, 6*i+4)
		if vec.Vec_BitEntry(ins, int(id)) == 0 {
			vec.Vec_IntPush(roots, id)
		}
	}
	return roots
}
func Acec_ManCollectXorRoots(p *Gia_Man_t, xors *Vec_Int_t) *Vec_Int_t {
	roots := vec.Vec_IntAlloc(100)
	ins := vec.Vec_BitStart(int(gia.Gia_ManObjNum(p)))
	for i := 0; 4*i < vec.Vec_IntSize(xors); i++ {
		for k := 1; k < 4; k++ {
			vec.Vec_BitWriteEntry(ins, int(vec.Vec_IntEntry(xors, 4*i+k)), 1)
		}
	}
	for i := 0; 4*i < vec.Vec_IntSize(xors); i++ {
		id := vec.Vec_IntEntry(xors, 4*i)
		if vec.Vec_BitEntry(ins, int(id)) == 0 {
			vec.Vec_IntPush(roots, id)
		}
	}
	return roots
}
func Acec_ManPoolGetPointed(p *Gia_Man_t, adds *Vec_Int_t) *Vec_Bit_t {
	m := vec.Vec_BitStart(int(gia.Gia_ManObjNum(p)))
	for i := 0; 6*i < vec.Vec_IntSize(adds); i++ {
		for k := 0; k < 3; k++ {
			vec.Vec_BitWriteEntry(m, int(vec.Vec_IntEntry(adds, 6*i+k)), 1)
		}
	}
	return m
}
func Acec_ManPoolTopMost(p *Gia_Man_t, adds *Vec_Int_t) *Vec_Int_t {
	tops := vec.Vec_IntAlloc(1000)
	m := Acec_ManPoolGetPointed(p, adds)
	for i := 0; 6*i < vec.Vec_IntSize(adds); i++ {
		if vec.Vec_BitEntry(m, int(vec.Vec_IntEntry(adds, 6*i+3))) == 0 && vec.Vec_BitEntry(m, int(vec.Vec_IntEntry(adds, 6*i+4))) == 0 {
			vec.Vec_IntPush(tops, int32(i))
		}
	}
	return tops
}
