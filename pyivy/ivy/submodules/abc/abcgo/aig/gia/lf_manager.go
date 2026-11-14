package gia

import (
	"fmt"
	"github.com/glycerine/abcgo/misc/tim"
	"github.com/glycerine/abcgo/misc/util"
	"github.com/glycerine/abcgo/misc/vec"
	"time"
)

// LfSwitchingProvider supplies the source mapper's switching simulation until
// the GIA switching simulator is connected to the native mapper.
var LfSwitchingProvider func(*Gia_Man_t, int32, int32, int32) *Vec_Flt_t

func Lf_ManComputeSwitching(g *Gia_Man_t, result *Vec_Flt_t) {
	if LfSwitchingProvider == nil {
		panic("LF switching simulator is not connected")
	}
	switching := LfSwitchingProvider(g, 48, 16, 0)
	if switching == nil || vec.Vec_FltSize(switching) != int(g.nObjs) {
		panic("LF invalid switching data")
	}
	*result = *switching
}
func Lf_ManAlloc(g *Gia_Man_t, pars *Jf_Par_t) *Lf_Man_t {
	if pars.NCutNum <= 1 || pars.NCutNum > LF_CUT_MAX || pars.NLutSize <= 1 || pars.NLutSize > LF_LEAF_MAX {
		panic("LF invalid mapping limits")
	}
	g.Refs = nil
	g.vMapping = nil
	Gia_ManCleanValue(g)
	if Gia_ManHasChoices(g) != 0 {
		Gia_ManSetPhase(g)
	}
	p := &Lf_Man_t{pGia: g, pPars: pars, clkStart: time.Now()}
	Lf_ManAnalyzeCoDrivers(g, &p.nCoDrivers, &p.nInverters)
	if pars.FPower != 0 {
		Lf_ManComputeSwitching(g, &p.vSwitches)
	}
	// Original packed header occupies six 32-bit words with variable leaves.
	p.nCutWords = (6 + pars.NLutSize + 1) >> 1
	p.nSetWords = p.nCutWords * pars.NCutNum
	if pars.FCutMin != 0 {
		p.vTtMem = vec.Vec_MemAllocForTT(int(pars.NLutSize), 0)
		if pars.FUseMux7 != 0 {
			vec.Vec_MemAddMuxTT(p.vTtMem, int(pars.NLutSize))
		}
	}
	n, nAnd := int(g.nObjs), int(Gia_ManAndNotBufNum(g))
	p.pObjBests = make([]Lf_Bst_t, nAnd)
	Lf_MemAlloc(&p.vStoreOld, 16, &p.vFreePages, p.nCutWords)
	Lf_MemAlloc(&p.vStoreNew, 16, &p.vFreePages, p.nCutWords)
	vec.Vec_IntFill(&p.vOffsets, n, -1)
	vec.Vec_IntFill(&p.vRequired, n, util.ABC_INFINITY)
	vec.Vec_IntFill(&p.vCutSets, nAnd, -1)
	vec.Vec_FltFill(&p.vFlowRefs, nAnd, 0)
	vec.Vec_IntFill(&p.vMapRefs, nAnd, 0)
	vec.Vec_IntFill(&p.vCiArrivals, int(Gia_ManCiNum(g)), 0)
	var index int32
	for _, obj := range g.pObjs {
		if Gia_ObjIsAndNotBuf(obj) != 0 {
			vec.Vec_IntWriteEntry(&p.vOffsets, int(obj.id), index)
			index++
		}
	}
	if int(index) != nAnd {
		panic("LF inconsistent object offsets")
	}
	Lf_ManSetFlowRefs(g, &p.vFlowRefs, &p.vOffsets)
	if pars.pTimesArr != nil {
		for i := int32(0); i < Gia_ManPiNum(g); i++ {
			vec.Vec_IntWriteEntry(&p.vCiArrivals, int(i), int32(pars.pTimesArr[i]))
		}
	}
	return p
}
func Lf_ManFree(p *Lf_Man_t) {
	if p == nil {
		return
	}
	p.pPars.pTimesArr = nil
	p.pPars.pTimesReq = nil
	if p.vTtMem != nil {
		vec.Vec_MemHashFree(p.vTtMem)
		vec.Vec_MemFree(p.vTtMem)
	}
	p.vTtMem = nil
	p.vMemSets = Vec_Ptr_t{}
	p.vFreePages = Vec_Ptr_t{}
	p.vStoreOld = Lf_Mem_t{}
	p.vStoreNew = Lf_Mem_t{}
	p.vFreeSets = Vec_Int_t{}
	p.vOffsets = Vec_Int_t{}
	p.vRequired = Vec_Int_t{}
	p.vCutSets = Vec_Int_t{}
	p.vFlowRefs = Vec_Flt_t{}
	p.vMapRefs = Vec_Int_t{}
	p.vSwitches = Vec_Flt_t{}
	p.vCiArrivals = Vec_Int_t{}
	p.pObjBests = nil
}
func Lf_ManSetDefaultPars(p *Jf_Par_t) {
	*p = Jf_Par_t{NLutSize: 6, NCutNum: 8, NRounds: 4, NRoundsEla: 1, NCoarseLimit: 3, NAreaTuner: 1, NVerbLimit: 5, DelayTarget: -1, FOptEdge: 1, FCoarsen: 1, NLutSizeMax: LF_LEAF_MAX, NCutNumMax: LF_CUT_MAX}
}
func Lf_ManPrintStats(p *Lf_Man_t, title string) {
	if p.pPars.FVerbose == 0 {
		return
	}
	fmt.Printf("%s : Level =%6d   Area =%9d   Edge =%9d   LUT =%9d  ", title, p.pPars.Delay, p.pPars.Area, p.pPars.Edge, p.pPars.Area+word(p.nInverters))
	if vec.Vec_FltSize(&p.vSwitches) > 0 {
		fmt.Printf("Swt =%8.1f  ", p.Switches)
	}
	if p.pPars.FUseMux7 != 0 {
		fmt.Printf("Mux7 =%7d  ", p.pPars.Mux7)
	}
	fmt.Printf("Time = %s\n", time.Since(p.clkStart))
}
func Lf_ManPrintInit(p *Lf_Man_t) {
	if p.pPars.FVerbose != 0 {
		fmt.Printf("LutSize = %d  CutNum = %d  Iter = %d  Edge = %d  CutMin = %d  Coarse = %d  Cut/Set = %d/%d Bytes\n", p.pPars.NLutSize, p.pPars.NCutNum, p.pPars.NRounds+p.pPars.NRoundsEla, p.pPars.FOptEdge, p.pPars.FCutMin, p.pPars.FCoarsen, 8*p.nCutWords, 8*p.nSetWords)
	}
}
func Lf_ManComputeMapping(p *Lf_Man_t) {
	if p.vStoreNew.iCur != 0 {
		panic("LF new cut store is not empty")
	}
	Lf_ManSetCutRefs(p)
	if p.pGia.pManTime != nil {
		if Gia_ManBufNum(p.pGia) != 0 {
			panic("LF timing with buffers")
		}
		tm, ok := p.pGia.pManTime.(*tim.Tim_Man_t)
		if !ok {
			panic("LF invalid timing manager")
		}
		tim.Tim_ManIncrementTravId(tm)
		for _, obj := range p.pGia.pObjs[1:] {
			if Gia_ObjIsBuf(obj) != 0 {
				continue
			}
			if Gia_ObjIsAnd(obj) != 0 {
				Lf_ObjMergeOrder(p, obj.id)
			} else if Gia_ObjIsCi(obj) != 0 {
				Lf_ObjSetCiArrival(p, Gia_ObjCioId(obj), int32(tim.Tim_ManGetCiArrival(tm, int(Gia_ObjCioId(obj)))))
			} else if Gia_ObjIsCo(obj) != 0 {
				tim.Tim_ManSetCoArrival(tm, int(Gia_ObjCioId(obj)), float32(Lf_ObjCoArrival(p, Gia_ObjCioId(obj))))
			} else {
				panic("LF unexpected timing object")
			}
		}
	} else {
		for _, obj := range p.pGia.pObjs {
			if Gia_ObjIsAndNotBuf(obj) != 0 {
				Lf_ObjMergeOrder(p, obj.id)
			}
		}
	}
	Lf_MemRecycle(&p.vStoreOld)
	p.vStoreOld, p.vStoreNew = p.vStoreNew, p.vStoreOld
	if p.fUseEla != 0 {
		Lf_ManCountMapRefs(p)
		Lf_ManPrintStats(p, "Ela  ")
	} else {
		Lf_ManSetMapRefs(p)
		if p.Iter != 0 {
			Lf_ManPrintStats(p, "Area ")
		} else {
			Lf_ManPrintStats(p, "Delay")
		}
	}
}
