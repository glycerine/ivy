package gia

import "github.com/glycerine/abcgo/misc/tim"

// Gia_ManChoiceLevel_rec computes levels through fanins, choice siblings and
// the inputs of the timing box that drives a combinational input.
func Gia_ManChoiceLevel_rec(p *Gia_Man_t, obj *Gia_Obj_t) {
	if Gia_ObjIsTravIdCurrent(p, obj) != 0 {
		return
	}
	Gia_ObjSetTravIdCurrent(p, obj)
	var level int32
	switch {
	case Gia_ObjIsCi(obj) != 0:
		if p.PManTime != nil {
			timing, ok := p.PManTime.(*tim.Tim_Man_t)
			if !ok {
				panic("gia: invalid timing manager")
			}
			box := tim.Tim_ManBoxForCi(timing, int(Gia_ObjCioId(obj)))
			if box >= 0 {
				first, count := tim.Tim_ManBoxInputFirst(timing, box), tim.Tim_ManBoxInputNum(timing, box)
				for i := 0; i < count; i++ {
					input := Gia_ManCo(p, int32(first+i))
					Gia_ManChoiceLevel_rec(p, input)
					level = max(level, Gia_ObjLevel(p, input))
				}
				level++
			}
		}
	case Gia_ObjIsCo(obj) != 0:
		fanin := Gia_ObjFanin0(obj)
		Gia_ManChoiceLevel_rec(p, fanin)
		level = Gia_ObjLevel(p, fanin)
	case Gia_ObjIsAnd(obj) != 0:
		a, b := Gia_ObjFanin0(obj), Gia_ObjFanin1(obj)
		Gia_ManChoiceLevel_rec(p, a)
		Gia_ManChoiceLevel_rec(p, b)
		level = 1 + max(Gia_ObjLevel(p, a), Gia_ObjLevel(p, b))
		if sibling := Gia_ObjSiblObj(p, obj.id); sibling != nil {
			Gia_ManChoiceLevel_rec(p, sibling)
			level = max(level, Gia_ObjLevel(p, sibling))
		}
	case Gia_ObjIsConst0(obj) == 0:
		panic("gia: invalid choice-level node")
	}
	Gia_ObjSetLevel(p, obj, level)
}
func Gia_ManChoiceLevel(p *Gia_Man_t) int32 {
	Gia_ManCleanLevels(p, p.nObjs)
	Gia_ManIncrementTravId(p)
	var maximum int32
	for i := int32(0); i < Gia_ManCoNum(p); i++ {
		obj := Gia_ManCo(p, i)
		Gia_ManChoiceLevel_rec(p, obj)
		maximum = max(maximum, Gia_ObjLevel(p, obj))
	}
	for i := int32(0); i < Gia_ManCiNum(p); i++ {
		obj := Gia_ManCi(p, i)
		Gia_ManChoiceLevel_rec(p, obj)
		maximum = max(maximum, Gia_ObjLevel(p, obj))
	}
	return maximum
}
