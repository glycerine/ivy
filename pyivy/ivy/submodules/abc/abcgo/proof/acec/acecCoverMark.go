package acec

// Adder-cone marking from src/proof/acec/acecCover.c.
import (
	"github.com/glycerine/abcgo/aig/gia"
	"github.com/glycerine/abcgo/misc/vec"
)

func Gia_AcecMark_rec(p *Gia_Man_t, id, first int32) {
	o := gia.Gia_ManObj(p, id)
	if gia.Gia_ObjMark0(o) != 0 && first == 0 {
		return
	}
	if gia.Gia_ObjIsAnd(o) == 0 {
		panic("acec: unmarked adder leaf")
	}
	gia.Gia_ObjSetMark1(o, 1)
	Gia_AcecMark_rec(p, gia.Gia_ObjFaninId0(o, id), 0)
	Gia_AcecMark_rec(p, gia.Gia_ObjFaninId1(o, id), 0)
}
func Gia_AcecMarkFadd(p *Gia_Man_t, sigs []int32) {
	for _, id := range sigs[:3] {
		gia.Gia_ObjSetMark0(gia.Gia_ManObj(p, id), 1)
	}
	Gia_AcecMark_rec(p, sigs[3], 1)
	Gia_AcecMark_rec(p, sigs[4], 1)
}
func Gia_AcecMarkHadd(p *Gia_Man_t, sigs []int32) {
	o := gia.Gia_ManObj(p, sigs[0])
	gia.Gia_ObjSetMark0(gia.Gia_ManObj(p, gia.Gia_ObjFaninId0(o, sigs[0])), 1)
	gia.Gia_ObjSetMark0(gia.Gia_ManObj(p, gia.Gia_ObjFaninId1(o, sigs[0])), 1)
	Gia_AcecMark_rec(p, sigs[0], 1)
	Gia_AcecMark_rec(p, sigs[1], 1)
}
func Gia_AcecCollectXors_rec(p *Gia_Man_t, o *Gia_Obj_t, m *Vec_Bit_t, xors *Vec_Int_t) {
	if gia.Gia_ObjIsXor(o) == 0 {
		return
	}
	vec.Vec_IntPush(xors, gia.Gia_ObjId(p, o))
	Gia_AcecCollectXors_rec(p, gia.Gia_ObjFanin0(o), m, xors)
	Gia_AcecCollectXors_rec(p, gia.Gia_ObjFanin1(o), m, xors)
}
func Gia_AcecCollectXors(p *Gia_Man_t, m *Vec_Bit_t) *Vec_Int_t {
	xors := vec.Vec_IntAlloc(100)
	Gia_AcecCollectXors_rec(p, gia.Gia_ObjFanin0(gia.Gia_ManCo(p, gia.Gia_ManCoNum(p)-1)), m, xors)
	return xors
}
