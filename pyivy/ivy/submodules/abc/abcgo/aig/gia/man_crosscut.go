package gia

import "github.com/glycerine/abcgo/misc/vec"

func Gia_ManDfsForCrossCut_rec(p *Gia_Man_t, obj *Gia_Obj_t, nodes *Vec_Int_t) {
	if Gia_ObjIsTravIdCurrent(p, obj) != 0 {
		return
	}
	Gia_ObjSetTravIdCurrent(p, obj)
	if Gia_ObjIsCi(obj) != 0 {
		vec.Vec_IntPush(nodes, obj.id)
		return
	}
	if Gia_ObjIsCo(obj) != 0 {
		fanin := Gia_ObjFanin0(obj)
		fanin.Value++
		Gia_ManDfsForCrossCut_rec(p, fanin, nodes)
		vec.Vec_IntPush(nodes, obj.id)
		return
	}
	if Gia_ObjIsAnd(obj) == 0 {
		panic("gia: crosscut expected AND")
	}
	a, b := Gia_ObjFanin0(obj), Gia_ObjFanin1(obj)
	a.Value++
	b.Value++
	Gia_ManDfsForCrossCut_rec(p, a, nodes)
	Gia_ManDfsForCrossCut_rec(p, b, nodes)
	vec.Vec_IntPush(nodes, obj.id)
}
func Gia_ManDfsForCrossCut(p *Gia_Man_t, reverse int32) *Vec_Int_t {
	Gia_ManCleanValue(p)
	nodes := vec.Vec_IntAlloc(int(p.nObjs))
	Gia_ManIncrementTravId(p)
	cos := vec.Vec_IntArray(p.vCos)
	if reverse != 0 {
		for i := len(cos) - 1; i >= 0; i-- {
			obj := Gia_ManObj(p, cos[i])
			if Gia_ObjIsConst0(Gia_ObjFanin0(obj)) == 0 {
				Gia_ManDfsForCrossCut_rec(p, obj, nodes)
			}
		}
	} else {
		for _, id := range cos {
			obj := Gia_ManObj(p, id)
			if Gia_ObjIsConst0(Gia_ObjFanin0(obj)) == 0 {
				Gia_ManDfsForCrossCut_rec(p, obj, nodes)
			}
		}
	}
	return nodes
}
func Gia_ManCrossCut(p *Gia_Man_t, reverse int32) int32 {
	nodes := Gia_ManDfsForCrossCut(p, reverse)
	var current, maximum int32
	for _, id := range vec.Vec_IntArray(nodes) {
		obj := Gia_ManObj(p, id)
		if obj.Value != 0 {
			current++
		}
		maximum = max(maximum, current)
		if Gia_ObjIsAnd(obj) != 0 {
			for _, fanin := range []*Gia_Obj_t{Gia_ObjFanin0(obj), Gia_ObjFanin1(obj)} {
				fanin.Value--
				if fanin.Value == 0 {
					current--
				}
			}
		} else if Gia_ObjIsCo(obj) != 0 {
			fanin := Gia_ObjFanin0(obj)
			fanin.Value--
			if fanin.Value == 0 {
				current--
			}
		}
	}
	for _, obj := range p.pObjs {
		if obj.Value != 0 {
			panic("gia: crosscut references remain")
		}
	}
	return maximum
}
