// Go port of manager lifecycle and concrete counterexample routines from
// pdrMan.c by Alan Mishchenko, UC Berkeley (2010).
package pdr

import (
	"fmt"
	"time"

	"github.com/glycerine/abcgo/aig/aig"
	"github.com/glycerine/abcgo/aig/gia"
	"github.com/glycerine/abcgo/aig/saig"
	"github.com/glycerine/abcgo/misc/util"
	"github.com/glycerine/abcgo/misc/vec"
	"github.com/glycerine/abcgo/sat/cnf"
)

func Pdr_ManStart(graph *aig.Graph, pars *Pdr_Par_t, initialPriority *vec.Vec_Int_t) *Pdr_Man_t {
	if graph == nil || pars == nil {
		panic("pdr: manager needs graph and parameters")
	}
	p := &Pdr_Man_t{pPars: pars, pAig: graph, vClauses: vec.Vec_VecAlloc(0)}
	if pars.FFlopPrio != 0 || pars.FNewXSim != 0 || pars.FUseAbs != 0 {
		p.pGia = gia.Gia_ManFromAigSimple(graph)
	}
	regs := aig.Aig_ManRegNum(graph)
	p.pOrder = make([]int32, regs)
	p.vActVars = vec.Vec_IntAlloc(256)
	if pars.FMonoCnf == 0 {
		p.vVLits = vec.Vec_WecStart(int(1 + max(int32(1), aig.Aig_ManLevels(graph))))
	}
	p.nPrioShift = util.Abc_Base2Log(uint32(regs))
	if initialPriority != nil {
		p.vPrio = initialPriority
	} else if pars.FFlopPrio != 0 {
		p.vPrio = Pdr_ManDeriveFlopPriorities2(p.pGia, 1)
	} else {
		p.vPrio = vec.Vec_IntStart(int(regs))
	}
	p.vLits, p.vCiObjs, p.vCoObjs = vec.Vec_IntAlloc(100), vec.Vec_IntAlloc(100), vec.Vec_IntAlloc(100)
	p.vCiVals, p.vCoVals, p.vNodes = vec.Vec_IntAlloc(100), vec.Vec_IntAlloc(100), vec.Vec_IntAlloc(100)
	p.vUndo, p.vVisits, p.vCi2Rem, p.vRes = vec.Vec_IntAlloc(100), vec.Vec_IntAlloc(100), vec.Vec_IntAlloc(100), vec.Vec_IntAlloc(100)
	p.pCnfMan = cnf.Cnf_ManStart()
	if pars.FNewXSim != 0 {
		p.pTxs3 = Txs3_ManStart(p, graph, p.vPrio)
	}
	if !aig.Aig_ManHasFanout(graph) {
		aig.Aig_ManFanoutStart(graph)
	}
	outputs := saig.Saig_ManPoNum(graph)
	if pars.NTimeOutOne > 0 {
		p.pTime4Outs = make([]abctime, outputs)
		for i := range p.pTime4Outs {
			p.pTime4Outs[i] = int64(pars.NTimeOutOne)*int64(time.Second)/1000 + pdrClockQuantum
		}
	}
	if pars.FSolveAll != 0 {
		p.vCexes = vec.Vec_PtrStart(int(outputs))
		pars.VOutMap = vec.Vec_IntStart(int(outputs))
		vec.Vec_IntFill(pars.VOutMap, int(outputs), -2)
	}
	return p
}

func Pdr_ManStop(p *Pdr_Man_t) {
	if p == nil {
		return
	}
	if p.pAig != nil {
		aig.Aig_ManCleanMarkAB(p.pAig)
	}
	if p.pPars != nil && p.pPars.FVerbose != 0 {
		Pdr_ManPrintStats(p)
	}
	if p.vClauses != nil {
		for i := 0; i < vec.Vec_VecSize(p.vClauses); i++ {
			for _, item := range vec.Vec_PtrArray(vec.Vec_VecEntry(p.vClauses, i)) {
				if set, ok := item.(*Pdr_Set_t); ok && set != nil {
					Pdr_SetDeref(set)
				}
			}
		}
	}
	Pdr_QueueStop(p)
	if p.pTxs3 != nil {
		Txs3_ManStop(p.pTxs3)
	}
	if p.pCnfMan != nil {
		cnf.Cnf_ManStop(p.pCnfMan)
	}
	if p.pAig != nil && aig.Aig_ManHasFanout(p.pAig) {
		aig.Aig_ManFanoutStop(p.pAig)
	}
	*p = Pdr_Man_t{}
}

func Pdr_ManDeriveCex(p *Pdr_Man_t) *util.Abc_Cex_t {
	var frames int32
	for entry := p.pQueue; entry != nil; entry = entry.pNext {
		frames++
	}
	cex := util.Abc_CexAlloc(aig.Aig_ManRegNum(p.pAig), saig.Saig_ManPiNum(p.pAig), frames)
	cex.IPo, cex.IFrame = p.iOutCur, frames-1
	var frame int32
	for entry := p.pQueue; entry != nil; entry = entry.pNext {
		for _, lit := range entry.pState.Lits[entry.pState.nLits:entry.pState.nTotal] {
			if lit&1 == 0 && lit>>1 < cex.NPis {
				util.Abc_InfoSetBit(cex.PData, cex.NRegs+frame*cex.NPis+(lit>>1))
			}
		}
		frame++
	}
	if saig.Saig_ManVerifyCex(p.pAig, cex) == 0 {
		fmt.Printf("CEX for output %d is not valid.\n", p.iOutCur)
	}
	return cex
}

// Pdr_ManPrintStats is the verbose report of Pdr_ManStop.
func Pdr_ManPrintStats(p *Pdr_Man_t) {
	satShare := "-nan" // C's 100.0 * 0 / 0
	if p.nCalls != 0 {
		satShare = fmt.Sprintf("%.1f", 100.0*float64(p.nCallsS)/float64(p.nCalls))
	}
	util.Abc_Print(1, "Block =%5d  Oblig =%6d  Clause =%6d  Call =%6d (sat=%s%%)  Cex =%4d  Start =%4d\n",
		p.nBlocks, p.nObligs, p.nCubes, p.nCalls, satShare, p.nCexesTotal, p.nStarts)
	util.Abc_PrintTimeP(1, "SAT solving", p.tSat, p.tTotal)
	util.Abc_PrintTimeP(1, "  unsat    ", p.tSatUnsat, p.tTotal)
	util.Abc_PrintTimeP(1, "  sat      ", p.tSatSat, p.tTotal)
	util.Abc_PrintTimeP(1, "Generalize ", p.tGeneral, p.tTotal)
	util.Abc_PrintTimeP(1, "Push clause", p.tPush, p.tTotal)
	util.Abc_PrintTimeP(1, "Ternary sim", p.tTsim, p.tTotal)
	util.Abc_PrintTimeP(1, "Containment", p.tContain, p.tTotal)
	util.Abc_PrintTimeP(1, "CNF compute", p.tCnf, p.tTotal)
	util.Abc_PrintTimeP(1, "Refinement ", p.tAbs, p.tTotal)
	util.Abc_PrintTimeP(1, "TOTAL      ", p.tTotal, p.tTotal)
}
