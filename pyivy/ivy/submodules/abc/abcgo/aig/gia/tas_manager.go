package gia

import (
	"time"

	"github.com/glycerine/abcgo/misc/vec"
)

type Tas_Par_t struct {
	nBTLimit, nJustLimit, nBTThis, nBTThisNc, nJustThis, nBTTotal, nJustTotal int32
	VarDecay                                                                  float32
	VarInc                                                                    int32
	fUseActive, fUseHighest, fUseLowest, fUseMaxFF, fVerbose                  int32
}

type Tas_Man_t struct {
	Pars                                              Tas_Par_t
	pAig                                              *Gia_Man_t
	pProp, pJust, pClauses                            Tas_Que_t
	pIter                                             int32
	vLevReas, vModel, vWatchLits, vActiveVars         *vec.Vec_Int_t
	vTemp                                             *vec.Vec_Ptr_t
	pStore                                            Tas_Sto_t
	pWatches                                          []int32
	nClauses                                          int32
	pActivity                                         []float32
	nSatUnsat, nSatSat, nSatUndec, nSatTotal          int32
	nConfUnsat, nConfSat, nConfUndec                  int32
	timeSatUnsat, timeSatSat, timeSatUndec, timeTotal time.Duration
}

func Tas_SetDefaultParams(p *Tas_Par_t) {
	*p = Tas_Par_t{nBTLimit: 2000, nJustLimit: 2000, fUseHighest: 1, fVerbose: 1, VarDecay: 0.95, VarInc: 1}
}
func Tas_ManAlloc(aig *Gia_Man_t, limit int32) *Tas_Man_t {
	if aig == nil {
		panic("gia: SAT manager requires an AIG")
	}
	p := &Tas_Man_t{pAig: aig, vLevReas: vec.Vec_IntAlloc(1000), vModel: vec.Vec_IntAlloc(1000), vTemp: vec.Vec_PtrAlloc(1000), vWatchLits: vec.Vec_IntAlloc(100), vActiveVars: vec.Vec_IntAlloc(100), pWatches: make([]int32, 2*int(aig.nObjs)), pActivity: make([]float32, aig.nObjs)}
	Tas_SetDefaultParams(&p.Pars)
	p.Pars.nBTLimit = limit
	p.pClauses.pData = []*Gia_Obj_t{nil}
	p.pClauses.iHead, p.pClauses.iTail = 1, 1
	p.pStore.reset()
	return p
}
func Tas_ManStop(p *Tas_Man_t) {
	if p != nil {
		*p = Tas_Man_t{}
	}
}
func Tas_ReadModel(p *Tas_Man_t) *vec.Vec_Int_t { return p.vModel }
func Tas_ManCheckLimits(p *Tas_Man_t) int32 {
	return giaBool(p.Pars.nJustThis > p.Pars.nJustLimit || p.Pars.nBTThis > p.Pars.nBTLimit)
}

func Tas_VarIsAssigned(node *Gia_Obj_t) int32 { return int32(node.fMark0) }
func Tas_VarAssign(node *Gia_Obj_t) {
	if node.fMark0 != 0 {
		panic("gia: SAT variable already assigned")
	}
	node.fMark0 = 1
}
func Tas_VarUnassign(node *Gia_Obj_t) {
	if node.fMark0 == 0 {
		panic("gia: SAT variable is unassigned")
	}
	node.fMark0, node.fMark1, node.Value = 0, 0, ^uint32(0)
}
func Tas_VarValue(node *Gia_Obj_t) int32 {
	if node.fMark0 == 0 {
		panic("gia: SAT variable is unassigned")
	}
	return int32(node.fMark1)
}
func Tas_VarSetValue(node *Gia_Obj_t, value int32) {
	if node.fMark0 == 0 || value < 0 || value > 1 {
		panic("gia: invalid SAT variable value")
	}
	node.fMark1 = uint32(value)
}
func Tas_VarIsJust(node *Gia_Obj_t) int32 {
	return giaBool(Gia_ObjIsAnd(node) != 0 && Tas_VarIsAssigned(Gia_ObjFanin0(node)) == 0 && Tas_VarIsAssigned(Gia_ObjFanin1(node)) == 0)
}
func tasFaninValue(node *Gia_Obj_t, fanin *Gia_Obj_t, complement int32) int32 {
	if Tas_VarIsAssigned(fanin) == 0 {
		return 2
	}
	return Tas_VarValue(fanin) ^ complement
}
func Tas_VarFanin0Value(node *Gia_Obj_t) int32 {
	return tasFaninValue(node, Gia_ObjFanin0(node), Gia_ObjFaninC0(node))
}
func Tas_VarFanin1Value(node *Gia_Obj_t) int32 {
	return tasFaninValue(node, Gia_ObjFanin1(node), Gia_ObjFaninC1(node))
}
func Tas_VarToLit(p *Tas_Man_t, node *Gia_Obj_t) int32 {
	return Gia_ObjId(p.pAig, node)<<1 | (1 ^ Tas_VarValue(node))
}
func Tas_LitIsTrue(node *Gia_Obj_t, lit int32) int32 { return giaBool(Tas_VarValue(node) != lit&1) }
func Tas_ClsHandle(p *Tas_Man_t, clause *Tas_Cls_t) int32 {
	if clause == nil || clause.store != &p.pStore {
		panic("gia: foreign SAT clause")
	}
	return clause.handle
}
func Tas_ClsFromHandle(p *Tas_Man_t, handle int32) *Tas_Cls_t { return p.pStore.clause(handle) }
func tasReason(p *Tas_Man_t, node *Gia_Obj_t, index int) int32 {
	if node.Value == ^uint32(0) {
		panic("gia: missing SAT trail index")
	}
	return vec.Vec_IntEntry(p.vLevReas, 3*int(node.Value)+index)
}
func Tas_VarDecLevel(p *Tas_Man_t, node *Gia_Obj_t) int32 { return tasReason(p, node, 0) }
func Tas_VarReason0(p *Tas_Man_t, node *Gia_Obj_t) *Gia_Obj_t {
	return Gia_ManObj(p.pAig, node.id+tasReason(p, node, 1))
}
func Tas_VarReason1(p *Tas_Man_t, node *Gia_Obj_t) *Gia_Obj_t {
	return Gia_ManObj(p.pAig, node.id+tasReason(p, node, 2))
}
func Tas_ClauseDecLevel(p *Tas_Man_t, clause int32) int32 {
	return Tas_VarDecLevel(p, p.pClauses.pData[clause])
}
func Tas_VarHasReasonCls(p *Tas_Man_t, node *Gia_Obj_t) int32 {
	return giaBool(tasReason(p, node, 1) == 0 && tasReason(p, node, 2) != 0)
}
func Tas_VarReasonCls(p *Tas_Man_t, node *Gia_Obj_t) *Tas_Cls_t {
	return Tas_ClsFromHandle(p, tasReason(p, node, 2))
}
func Tas_ManSaveModel(p *Tas_Man_t, model *vec.Vec_Int_t) {
	vec.Vec_IntClear(model)
	p.pProp.iHead = 0
	for i := p.pProp.iHead; i < p.pProp.iTail && p.pProp.pData[i] != nil; i++ {
		node := p.pProp.pData[i]
		if Gia_ObjIsCi(node) != 0 {
			vec.Vec_IntPush(model, Gia_ObjCioId(node)<<1|(1^Tas_VarValue(node)))
		}
	}
}
