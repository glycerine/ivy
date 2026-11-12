package gia

import (
	"fmt"
	"github.com/glycerine/abcgo/misc/vec"
)

func Gia_ManDupWithConstrCollectAnd_rec(p *Gia_Man_t, obj *Gia_Obj_t, super *Vec_Int_t, first int32) {
	lit := Gia_ObjToLit(p, obj)
	regular := Gia_Regular(obj)
	if first == 0 && (lit&1 != 0 || Gia_ObjIsAnd(regular) == 0) {
		vec.Vec_IntPushUnique(super, lit)
		return
	}
	a, b, _, kind := p.Gate(regular.id)
	if kind != 2 {
		vec.Vec_IntPushUnique(super, lit)
		return
	}
	Gia_ManDupWithConstrCollectAnd_rec(p, Gia_ObjFromLit(p, a), super, 0)
	Gia_ManDupWithConstrCollectAnd_rec(p, Gia_ObjFromLit(p, b), super, 0)
}
func Gia_ManDupWithConstr(p *Gia_Man_t) *Gia_Man_t {
	if Gia_ManPoNum(p) != 1 || Gia_ManRegNum(p) != 0 {
		panic("gia: constraint extraction requires one combinational output")
	}
	output := Gia_ManPo(p, 0)
	driver := Gia_ObjFaninLit0(output, output.id)
	if driver&1 != 0 {
		fmt.Println("The miter's output is not AND-decomposable.")
		return nil
	}
	if driver == 0 {
		fmt.Println("The miter's output is a constant.")
		return nil
	}
	super := vec.Vec_IntAlloc(16)
	Gia_ManDupWithConstrCollectAnd_rec(p, Gia_ObjFromLit(p, driver), super, 1)
	if vec.Vec_IntSize(super) < 2 {
		return nil
	}
	Gia_ManLevelNum(p)
	best := int32(-1)
	level := int32(-1)
	for _, lit := range vec.Vec_IntArray(super) {
		if l := Gia_ObjLevelId(p, lit>>1); l > level {
			level = l
			best = lit
		}
	}
	out := giaQuantStart(p)
	giaComposeNodes(out, p, true)
	other := int32(1)
	for _, lit := range vec.Vec_IntArray(super) {
		if lit != best {
			other = Gia_ManHashAnd(out, other, Gia_ObjLitCopy(p, lit))
		}
	}
	Gia_ManAppendCo(out, Gia_ObjLitCopy(p, best))
	Gia_ManAppendCo(out, other^1)
	out.nConstrs = 1
	return giaCleanComposed(out)
}
