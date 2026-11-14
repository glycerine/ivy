package gia

import (
	"fmt"
	"github.com/glycerine/abcgo/misc/util"
	"github.com/glycerine/abcgo/misc/vec"
	"unsafe"
)

func Mf_ManPrintStats(p *Mf_Man_t, title string) {
	if p.pPars.FVerbose == 0 {
		return
	}
	q := p.pPars
	fmt.Printf("%s :  Level =%6d   Area =%9d   Edge =%9d   ", title, q.Delay, q.Area, q.Edge)
	if q.FGenCnf != 0 {
		fmt.Printf("CNF =%9d   ", q.Clause)
	}
	if q.FGenLit != 0 {
		fmt.Printf("FFL =%9d   ", q.Clause)
	}
	util.Abc_PrintTime(1, "Time", util.Abc_Clock()-p.clkStart)
}
func Mf_ManPrintInit(p *Mf_Man_t) {
	q := p.pPars
	if q.FVerbose == 0 {
		return
	}
	fmt.Printf("LutSize = %d  CutNum = %d  Iter = %d  Edge = %d  CutMin = %d  Coarse = %d  CNF = %d  FFL = %d\nComputing cuts...\r", q.NLutSize, q.NCutNum, q.NRounds+q.NRoundsEla, q.FOptEdge, q.FCutMin, q.FCoarsen, q.FGenCnf, q.FGenLit)
}
func Mf_ManPrintQuit(p *Mf_Man_t, result *Gia_Man_t) {
	if p.CutCount[0] == 0 {
		p.CutCount[0] = 1
	}
	if p.pPars.FVerbose == 0 {
		return
	}
	fmt.Printf("CutPair = %.0f  Merge = %.0f (%.2f %%)  Eval = %.0f (%.2f %%)  Cut = %.0f (%.2f %%)\n", p.CutCount[0], p.CutCount[1], 100*p.CutCount[1]/p.CutCount[0], p.CutCount[2], 100*p.CutCount[2]/p.CutCount[0], p.CutCount[3], 100*p.CutCount[3]/p.CutCount[0])
	gia := Gia_ManMemory(p.pGia) / (1 << 20)
	man := float64(unsafe.Sizeof(Mf_Obj_t{})) * float64(len(p.pLfObjs)) / (1 << 20)
	cuts := float64(4*mfPageSize*len(p.vPages)) / (1 << 20)
	var truth float64
	if p.vTtMem != nil {
		truth = vec.Vec_MemMemory(p.vTtMem) / (1 << 20)
	}
	mapping := vec.Vec_IntMemory(result.VMapping) / (1 << 20)
	fmt.Printf("Gia = %.2f MB  Man = %.2f MB  Cut = %.2f MB   Map = %.2f MB  TT = %.2f MB  Total = %.2f MB\n", gia, man, cuts, mapping, truth, gia+man+cuts+mapping+truth)
	for i := int32(0); i <= p.pPars.NLutSize; i++ {
		fmt.Printf("%d = %d  ", i, p.nCutCounts[i])
	}
	if p.vTtMem != nil {
		fmt.Printf("TT = %d (%.2f %%)  ", vec.Vec_MemEntryNum(p.vTtMem), 100*float64(vec.Vec_MemEntryNum(p.vTtMem))/p.CutCount[2])
	}
	util.Abc_PrintTime(1, "Time", util.Abc_Clock()-p.clkStart)
}
func mfRoundTitle(p *Mf_Man_t) string {
	if p.fUseEla != 0 {
		return "Ela  "
	}
	if p.Iter != 0 {
		return "Area "
	}
	return "Delay"
}
func Mf_ManComputeCuts(p *Mf_Man_t) {
	for _, obj := range p.pGia.pObjs {
		if Gia_ObjIsAnd(obj) != 0 {
			Mf_ObjMergeOrder(p, obj.id)
		}
	}
	Mf_ManSetMapRefs(p)
	Mf_ManPrintStats(p, mfRoundTitle(p))
}
func Mf_ManComputeMapping(p *Mf_Man_t) {
	for _, obj := range p.pGia.pObjs {
		if Gia_ObjIsAnd(obj) != 0 {
			Mf_ObjComputeBestCut(p, obj.id)
		}
	}
	Mf_ManSetMapRefs(p)
	Mf_ManPrintStats(p, mfRoundTitle(p))
}

func Mf_ManPerformMapping(g *Gia_Man_t, pars *Jf_Par_t) *Gia_Man_t {
	if pars.FGenCnf != 0 || pars.FGenLit != 0 {
		pars.FCutMin = 1
	}
	if Gia_ManHasChoices(g) != 0 {
		pars.FCutMin = 1
		pars.FCoarsen = 0
	}
	coarse := g
	if pars.FCoarsen != 0 {
		coarse = Gia_ManDupMuxes(g, pars.NCoarseLimit)
	}
	p := Mf_ManAlloc(coarse, pars)
	p.pGia0 = g
	if pars.FVerbose != 0 && pars.FCoarsen != 0 {
		fmt.Print("Initial ")
		Gia_ManPrintMuxStats(g)
		fmt.Print("\nDerived ")
		Gia_ManPrintMuxStats(coarse)
		fmt.Println()
	}
	Mf_ManPrintInit(p)
	Mf_ManComputeCuts(p)
	for p.Iter = 1; p.Iter < pars.NRounds; p.Iter++ {
		Mf_ManComputeMapping(p)
	}
	p.fUseEla = 1
	for ; p.Iter < pars.NRounds+pars.NRoundsEla; p.Iter++ {
		Mf_ManComputeMapping(p)
	}
	if pars.FVeryVerbose != 0 && pars.FCutMin != 0 {
		vec.Vec_MemDumpTruthTables(p.vTtMem, p.pGia.pName, int(pars.NLutSize))
	}
	var result *Gia_Man_t
	if pars.FCutMin != 0 {
		result = Mf_ManDeriveMappingGia(p)
	} else if pars.FCoarsen != 0 {
		result = Mf_ManDeriveMappingCoarse(p)
	} else {
		result = Mf_ManDeriveMapping(p)
	}
	if pars.FGenCnf != 0 {
		g.pData = Mf_ManDeriveCnf(p, pars.fCnfObjIds, pars.fAddOrCla)
	}
	Gia_ManMappingVerify(result)
	Mf_ManPrintQuit(p, result)
	Mf_ManFree(p)
	if coarse != g {
		Gia_ManStop(coarse)
	}
	return result
}
