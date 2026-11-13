package abs

import (
	"time"

	"github.com/glycerine/abcgo/aig/gia"
	"github.com/glycerine/abcgo/misc/vec"
	"github.com/glycerine/abcgo/sat/bsat"
	"github.com/glycerine/abcgo/sat/cnf"
)

type Rfn_Obj_t struct{ Value, fVisit, fPPi, Prio, Sign uint32 }
type Gla_Obj_t struct {
	id                                                              int32
	iGiaObj                                                         int32
	fAbs, fCompl0, fConst, fPi, fPo, fRo, fRi, fAnd, fMark, nFanins uint32
	Fanins                                                          [4]int32
	vFrames                                                         vec.Vec_Int_t
}
type Gla_Man_t struct {
	pGia0, pGia                                          *gia.Gia_Man_t
	pPars                                                *Abs_Par_t
	vAbs                                                 *vec.Vec_Int_t
	pObjRoot                                             *Gla_Obj_t
	pObjs                                                []Gla_Obj_t
	pObj2Obj                                             []uint32
	nObjs, nAbsOld, nCexes, nObjAdded, nSatVars          int32
	pCnf                                                 *cnf.Cnf_Dat_t
	pSat                                                 *bsat.LabeledSolver2
	vTemp, vAddedNew, vObjCounts, vCoreCounts, vProofIds *vec.Vec_Int_t
	nProofIds                                            int32
	pvRefis                                              [][]Rfn_Obj_t
	pGia2                                                *gia.Gia_Man_t
	pRnm                                                 *Rnm_Man_t
	timeInit, timeSat, timeUnsat, timeCex, timeOther     time.Duration
}

func Gla_ObjId(p *Gla_Man_t, obj *Gla_Obj_t) int32 {
	if obj == nil || obj.id <= 0 || obj.id >= p.nObjs || &p.pObjs[obj.id] != obj {
		panic("invalid old GLA object")
	}
	return obj.id
}
func Gla_ManObj(p *Gla_Man_t, id int32) *Gla_Obj_t {
	if id < 0 || id >= p.nObjs {
		panic("invalid old GLA object index")
	}
	if id == 0 {
		return nil
	}
	return &p.pObjs[id]
}
func Gla_ManGiaObj(p *Gla_Man_t, obj *Gla_Obj_t) *gia.Gia_Obj_t {
	return gia.Gia_ManObj(p.pGia, obj.iGiaObj)
}
func Gla_ObjSatValue(p *Gla_Man_t, id, frame int32) int32 {
	obj := int32(p.pObj2Obj[id])
	if Gla_ManCheckVar(p, obj, frame) == 0 {
		return 0
	}
	value, _ := p.pSat.Value(Gla_ManGetVar(p, obj, frame))
	if value {
		return 1
	}
	return 0
}
func Gla_ObjRef(p *Gla_Man_t, obj *gia.Gia_Obj_t, frame int32) *Rfn_Obj_t {
	id := gia.Gia_ObjId(p.pGia, obj)
	if frame < 0 {
		panic("negative refinement frame")
	}
	if int(frame) >= len(p.pvRefis[id]) {
		p.pvRefis[id] = append(p.pvRefis[id], make([]Rfn_Obj_t, int(frame)+1-len(p.pvRefis[id]))...)
	}
	return &p.pvRefis[id][frame]
}
func Gla_ObjClearRef(obj *Rfn_Obj_t) { *obj = Rfn_Obj_t{} }
func Gla_ManCheckVar(p *Gla_Man_t, id, frame int32) int32 {
	obj := Gla_ManObj(p, id)
	if obj.fPo != 0 || obj.fRi != 0 {
		panic("output has no independent SAT variable")
	}
	if vec.Vec_IntGetEntry(&obj.vFrames, int(frame)) > 0 {
		return 1
	}
	return 0
}
func Gla_ManGetVar(p *Gla_Man_t, id, frame int32) int32 {
	obj := Gla_ManObj(p, id)
	if obj.fPo != 0 || obj.fRi != 0 {
		panic("output has no independent SAT variable")
	}
	v := vec.Vec_IntGetEntry(&obj.vFrames, int(frame))
	if v == 0 {
		v = p.nSatVars
		p.nSatVars++
		vec.Vec_IntSetEntry(&obj.vFrames, int(frame), v)
		vec.Vec_IntPush(p.vAddedNew, id)
		vec.Vec_IntPush(p.vAddedNew, frame)
	}
	return v
}
