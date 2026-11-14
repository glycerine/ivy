package if_

import "github.com/glycerine/abcgo/aig/gia"

// Gia_ManSetIfParsDefault initializes the GIA-facing IF mapping defaults.
func Gia_ManSetIfParsDefault(p *If_Par_t) {
	If_ManSetDefaultPars(p)
	p.fUseCoAttrs = 1
}

// Gia_ManToIf copies a GIA network into the IF mapper. This adapter lives in
// IF because IF already depends on GIA; the inverse dependency would cycle.
func Gia_ManToIf(src *gia.Gia_Man_t, pars *If_Par_t) *If_Man_t {
	if src == nil || pars == nil {
		panic("if: missing GIA or mapping parameters")
	}
	gia.Gia_ManChoiceLevel(src)
	if gia.Gia_ManHasChoices(src) != 0 {
		gia.Gia_ManMarkFanoutDrivers(src)
		defer gia.Gia_ManCleanMark0(src)
	}
	result := If_ManStart(pars)
	copies := make([]*If_Obj_t, gia.Gia_ManObjNum(src))
	copies[0] = If_Not(If_ManConst1(result))
	gia.Gia_ManFillValue(src)
	gia.Gia_ObjSetValue(gia.Gia_ManConst0(src), If_ObjId(If_ManConst1(result)))
	copyEdge := func(lit int32) *If_Obj_t {
		if lit < 0 || int(lit>>1) >= len(copies) || copies[lit>>1] == nil {
			panic("if: invalid GIA fanin")
		}
		return If_NotCond(copies[lit>>1], lit&1)
	}
	for id := int32(1); id < gia.Gia_ManObjNum(src); id++ {
		obj := gia.Gia_ManObj(src, id)
		a, b, c, kind := src.Gate(id)
		var mapped *If_Obj_t
		switch kind {
		case 1:
			mapped = If_ManCreateCi(result)
			If_ObjSetLevel(mapped, gia.Gia_ObjLevel(src, obj))
			if result.nLevelMax < int32(mapped.Level) {
				result.nLevelMax = int32(mapped.Level)
			}
		case 2:
			mapped = If_ManCreateAnd(result, copyEdge(a), copyEdge(b))
		case 3:
			mapped = If_ManCreateXor(result, copyEdge(a), copyEdge(b))
		case 4:
			mapped = If_ManCreateMux(result, copyEdge(c), copyEdge(b), copyEdge(a))
		case 5:
			mapped = copyEdge(a)
		case 6:
			mapped = If_ManCreateCo(result, copyEdge(a))
		default:
			panic("if: unsupported GIA node")
		}
		copies[id] = mapped
		gia.Gia_ObjSetValue(obj, If_ObjId(If_Regular(mapped)))
		if gia.Gia_ObjIsAnd(obj) != 0 && gia.Gia_ObjMark0(obj) != 0 && gia.Gia_ObjSibl(src, id) != 0 {
			prev := If_Regular(mapped)
			for sibling := gia.Gia_ObjSibl(src, id); sibling != 0; sibling = gia.Gia_ObjSibl(src, sibling) {
				next := If_Regular(copies[sibling])
				if next == nil {
					panic("if: choice sibling must precede representative")
				}
				If_ObjSetChoice(prev, next)
				prev = next
			}
			If_ManCreateChoice(result, If_Regular(mapped))
			pars.FExpRed = 0
		}
	}
	return result
}
