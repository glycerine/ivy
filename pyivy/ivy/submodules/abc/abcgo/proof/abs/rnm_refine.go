package abs

import (
	"fmt"
	"time"

	"github.com/glycerine/abcgo/aig/gia"
	"github.com/glycerine/abcgo/misc/util"
	"github.com/glycerine/abcgo/misc/vec"
)

// Rnm_ManVerifyUsingTerSim checks the selected pseudo-inputs by ternary simulation.
func Rnm_ManVerifyUsingTerSim(p *gia.Gia_Man_t, cex *util.Abc_Cex_t, mapping, objects, selected *vec.Vec_Int_t) {
	bit := cex.NRegs
	gia.Gia_ObjTerSimSet0(gia.Gia_ManConst0(p))
	for frame := int32(0); frame <= cex.IFrame; frame++ {
		for i, id := range vec.Vec_IntArray(mapping) {
			obj := gia.Gia_ManObj(p, id)
			obj.Value = 0
			if util.Abc_InfoHasBit(cex.PData, bit+int32(i)) {
				obj.Value = 1
			}
			if gia.Gia_ObjIsPi(p, obj) == 0 {
				gia.Gia_ObjTerSimSetX(obj)
			} else if obj.Value != 0 {
				gia.Gia_ObjTerSimSet1(obj)
			} else {
				gia.Gia_ObjTerSimSet0(obj)
			}
		}
		for _, id := range vec.Vec_IntArray(selected) {
			obj := gia.Gia_ManObj(p, id)
			if obj.Value != 0 {
				gia.Gia_ObjTerSimSet1(obj)
			} else {
				gia.Gia_ObjTerSimSet0(obj)
			}
		}
		for _, id := range vec.Vec_IntArray(objects) {
			obj := gia.Gia_ManObj(p, id)
			switch {
			case gia.Gia_ObjIsCo(obj) != 0:
				gia.Gia_ObjTerSimCo(obj)
			case gia.Gia_ObjIsAnd(obj) != 0:
				gia.Gia_ObjTerSimAnd(obj)
			case frame == 0:
				gia.Gia_ObjTerSimSet0(obj)
			default:
				gia.Gia_ObjTerSimRo(p, obj)
			}
		}
		bit += cex.NPis
	}
	for _, id := range vec.Vec_IntArray(mapping) {
		gia.Gia_ManObj(p, id).Value = 0
	}
	if gia.Gia_ObjTerSimGet1(gia.Gia_ManPo(p, 0)) == 0 {
		fmt.Println("\nRefinement verification has failed!!!")
	}
}

// Rnm_ManRefine computes a set of pseudo-inputs to refine a counterexample.
func Rnm_ManRefine(p *Rnm_Man_t, cex *util.Abc_Cex_t, mapping *vec.Vec_Int_t, propFanout, newRefinement, verbose int32) *vec.Vec_Int_t {
	start := time.Now()
	p.nCalls++
	p.pCex, p.vMap, p.fPropFanout, p.fVerbose = cex, mapping, propFanout, verbose
	Rnm_ManCollect(p)
	p.nObjsFrame = int32(1 + vec.Vec_IntSize(mapping) + vec.Vec_IntSize(p.vObjs))
	p.nObjs = p.nObjsFrame * (cex.IFrame + 1)
	if p.nObjs > p.nObjsAlloc {
		p.nObjsAlloc = p.nObjs + 10000
		p.pObjs = make([]Rnm_Obj_t, p.nObjsAlloc)
	}
	clear(p.pObjs[:p.nObjs])
	clock := time.Now()
	good := vec.Vec_IntAlloc(100)
	if Rnm_ManSensitize(p) != 0 {
		p.timeFwd += time.Since(clock)
		clock = time.Now()
		p.nVisited = 0
		Rnm_ManJustify_rec(p, gia.Gia_ObjFanin0(gia.Gia_ManPo(p.pGia, 0)), cex.IFrame, good)
		vec.Vec_IntUniqify(good)
		p.timeBwd += time.Since(clock)
	}
	clock = time.Now()
	Rnm_ManVerifyUsingTerSim(p.pGia, cex, mapping, p.vObjs, good)
	p.timeVer += time.Since(clock)
	if vec.Vec_IntSize(good) > 0 {
		var refined *vec.Vec_Int_t
		if newRefinement == 0 {
			refined = Rnm_ManFilterSelected(p, good)
		} else {
			refined = Rnm_ManFilterSelectedNew(p, good)
		}
		if vec.Vec_IntSize(refined) > 0 {
			good = refined
		}
	}
	Rnm_ManCleanValues(p)
	p.timeTotal += time.Since(start)
	p.nRefines += int32(vec.Vec_IntSize(good))
	return good
}
