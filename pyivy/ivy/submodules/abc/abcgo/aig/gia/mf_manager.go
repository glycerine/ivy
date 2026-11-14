package gia

import (
	"github.com/glycerine/abcgo/misc/util"
	"github.com/glycerine/abcgo/misc/vec"
)

func Mf_ManSetDefaultPars(p *Jf_Par_t) {
	*p = Jf_Par_t{NLutSize: 6, NCutNum: 8, NRounds: 2, NRoundsEla: 1, NCoarseLimit: 3, NAreaTuner: 1, NVerbLimit: 5, DelayTarget: -1, FOptEdge: 1, FCoarsen: 1, NLutSizeMax: MF_LEAF_MAX, NCutNumMax: MF_CUT_MAX}
}
func Mf_ManAlloc(g *Gia_Man_t, pars *Jf_Par_t) *Mf_Man_t {
	if pars.NCutNum <= 1 || pars.NCutNum > MF_CUT_MAX || pars.NLutSize <= 1 || pars.NLutSize > MF_LEAF_MAX {
		panic("gia: invalid mapper parameters")
	}
	g.Refs = nil
	g.vMapping = nil
	if Gia_ManHasChoices(g) != 0 {
		Gia_ManSetPhase(g)
	}
	p := &Mf_Man_t{pGia: g, pPars: pars, pLfObjs: make([]Mf_Obj_t, g.nObjs), iCur: 2, clkStart: util.Abc_Clock()}
	if pars.FCutMin != 0 {
		p.vTtMem = vec.Vec_MemAllocForTT(int(pars.NLutSize), 0)
	}
	if pars.FGenCnf != 0 || pars.FGenLit != 0 {
		vec.Vec_IntPush(&p.vCnfSizes, 1)
		vec.Vec_IntPush(&p.vCnfSizes, 2)
	}
	refs := vec.Vec_IntAlloc(0)
	Mf_ManSetFlowRefs(g, refs)
	for i, n := range vec.Vec_IntArray(refs) {
		p.pLfObjs[i].nFlowRefs = float32(n)
	}
	return p
}
func Mf_ManFree(p *Mf_Man_t) {
	if p == nil {
		return
	}
	if p.pPars.FGenCnf != 0 && p.pPars.FGenLit != 0 && vec.Vec_IntSize(&p.vCnfSizes) != vec.Vec_MemEntryNum(p.vTtMem) {
		panic("gia: inconsistent truth and CNF counts")
	}
	if p.vTtMem != nil {
		vec.Vec_MemHashFree(p.vTtMem)
		vec.Vec_MemFree(p.vTtMem)
		p.vTtMem = nil
	}
	p.vPages = nil
	p.pLfObjs = nil
	p.vCnfSizes = Vec_Int_t{}
	p.vCnfMem = Vec_Int_t{}
	p.vTemp = Vec_Int_t{}
}
