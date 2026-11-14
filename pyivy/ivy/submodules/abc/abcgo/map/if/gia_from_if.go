package if_

import (
	"github.com/glycerine/abcgo/aig/gia"
	"github.com/glycerine/abcgo/misc/vec"
)

// SetRecordedGiaBuilder connects the recorded-library implementation without
// making IF depend on the higher-level ABC package that owns that library.
func (p *If_Par_t_) SetRecordedGiaBuilder(build func(*Gia_Man_t, *If_Man_t, *If_Cut_t, *Vec_Int_t, int32) int32) {
	p.pFuncRecGia = build
}

// Gia_ManFromIfAigRec rebuilds an unmapped IF cone using structural hashing.
// Input and constant objects must already have their iCopy literals assigned.
func Gia_ManFromIfAigRec(dst *Gia_Man_t, obj *If_Obj_t) int32 {
	if obj.iCopy != 0 {
		return obj.iCopy
	}
	if If_ObjIsAnd(obj) == 0 {
		panic("if: unmapped terminal in reconstruction")
	}
	a := Gia_ManFromIfAigRec(dst, obj.pFanin0) ^ int32(obj.fCompl0)
	b := Gia_ManFromIfAigRec(dst, obj.pFanin1) ^ int32(obj.fCompl1)
	obj.iCopy = Gia_ManHashAnd(dst, a, b)
	return obj.iCopy
}

// Gia_ManFromIfAig rebuilds the mapped IF network as a structurally hashed GIA.
func Gia_ManFromIfAig(p *If_Man_t) *Gia_Man_t {
	pars := p.pPars
	assert(pars.PLutStruct == "", "LUT structure is not set")
	assert(pars.FDelayOpt != 0 || pars.FDsdBalance != 0 || pars.FUserRecLib != 0 || pars.FUserSesLib != 0, "balancing or recorded library is enabled")
	dst := Gia_ManStart(If_ManObjNum(p))
	Gia_ManHashAlloc(dst)
	mini, leaves := Vec_IntAlloc(1<<16), Vec_IntAlloc(16)
	defer Vec_IntFree(mini)
	defer Vec_IntFree(leaves)
	for i := int32(0); i < If_ManObjNum(p); i++ {
		obj := If_ManObj(p, i)
		if obj.nRefs == 0 && If_ObjIsTerm(obj) == 0 {
			continue
		}
		switch obj.Type {
		case IF_AND:
			cut := If_ObjCutBest(obj)
			vec.Vec_IntClear(leaves)
			for k := int32(0); k < int32(cut.nLeaves); k++ {
				Vec_IntPush(leaves, If_CutLeaf(p, cut, k).iCopy)
			}
			switch {
			case pars.FDelayOpt != 0:
				obj.iCopy = Gia_ManBuildFromMini(dst, p, cut, leaves, mini, 0, 0)
			case pars.FDsdBalance != 0:
				obj.iCopy = Gia_ManBuildFromMini(dst, p, cut, leaves, mini, 0, 1)
			case pars.FUserRecLib != 0:
				if pars.pFuncRecGia == nil {
					panic("if: recorded-library GIA builder is not configured")
				}
				obj.iCopy = pars.pFuncRecGia(dst, p, cut, leaves, 0)
			default:
				panic("if: unsupported recorded library")
			}
			if obj.iCopy < 0 {
				panic("if: cannot reconstruct mapped cut")
			}
		case IF_CI:
			obj.iCopy = Gia_ManAppendCi(dst)
		case IF_CO:
			obj.iCopy = Gia_ManAppendCo(dst, obj.pFanin0.iCopy^int32(obj.fCompl0))
		case IF_CONST1:
			obj.iCopy = 1
		default:
			panic("if: unknown object type")
		}
	}
	result := gia.Gia_ManRehash(dst, 0)
	Gia_ManStop(dst)
	return result
}
