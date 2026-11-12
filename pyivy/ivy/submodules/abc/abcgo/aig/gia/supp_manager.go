package gia

import "github.com/glycerine/abcgo/sat/satoko"

// Gia_Man2Min_t maintains the SAT and simulation state used when checking
// whether the conjunction of two AIG literals is constant.
type Gia_Man2Min_t struct {
	pGia                                      *Gia_Man_t
	iLits                                     [2]int32
	vCis, vObjs                               [2]*Vec_Int_t
	pSat                                      *satoko.Solver
	vSims                                     *Vec_Wrd_t
	vFrontier, vFanins                        *Vec_Ptr_t
	vSatVars                                  *Vec_Int_t
	nCisOld, iPattern                         int32
	nSatSat, nSatUnsat, nCalls, nSims, nSupps int32
}

func Gia_Man2SuppStart(pGia *Gia_Man_t) *Gia_Man2Min_t {
	p := &Gia_Man2Min_t{
		pGia:      pGia,
		pSat:      satoko.New(),
		vSims:     Vec_WrdAlloc(1000),
		vFrontier: Vec_PtrAlloc(1000),
		vFanins:   Vec_PtrAlloc(100),
		vSatVars:  Vec_IntAlloc(100),
		iPattern:  1,
	}
	for i := range p.vCis {
		p.vCis[i] = Vec_IntAlloc(512)
		p.vObjs[i] = Vec_IntAlloc(512)
	}
	p.pSat.Options().LearntRatio = 0
	return p
}

func Gia_Man2SuppStop(p *Gia_Man2Min_t) {
	Gia_ManCleanMark01(p.pGia)
	p.pSat = nil
	p.vCis, p.vObjs = [2]*Vec_Int_t{}, [2]*Vec_Int_t{}
	p.vSims, p.vFrontier, p.vFanins, p.vSatVars = nil, nil, nil, nil
}

func Gia_Min2ObjSatId(p *Gia_Man_t, obj *Gia_Obj_t) int32 {
	return Gia_ObjCopyArray(p, Gia_ObjId(p, obj))
}
func Gia_Min2ObjSetSatId(p *Gia_Man_t, obj *Gia_Obj_t, id int32) int32 {
	if Gia_Min2ObjSatId(p, obj) != -1 {
		panic("gia: SAT ID already assigned")
	}
	Gia_ObjSetCopyArray(p, Gia_ObjId(p, obj), id)
	return id
}
func Gia_Min2ObjCleanSatId(p *Gia_Man_t, obj *Gia_Obj_t) {
	if Gia_Min2ObjSatId(p, obj) == -1 {
		panic("gia: SAT ID is not assigned")
	}
	Gia_ObjSetCopyArray(p, Gia_ObjId(p, obj), -1)
}

func Gia_Min2ManGatherSupp(p *Gia_Man2Min_t) int32 {
	p.nSupps++
	Gia_ManIncrementTravId(p.pGia)
	var overlap int32
	for i := range p.iLits {
		Vec_IntClear(p.vCis[i])
		Vec_IntClear(p.vObjs[i])
		Gia_ManIncrementTravId(p.pGia)
		overlap = Gia_Min2ManGatherSupp_rec(p.pGia, p.iLits[i]>>1, p.vCis[i], p.vObjs[i])
		if i == 0 && overlap != 0 {
			panic("gia: first support cone overlaps previous traversal")
		}
	}
	return overlap
}
