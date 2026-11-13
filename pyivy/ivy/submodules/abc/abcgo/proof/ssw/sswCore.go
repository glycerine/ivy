package ssw

// Signal and latch correspondence drivers, ported from sswCore.c.
import (
	"fmt"
	"github.com/glycerine/abcgo/aig/aig"
	"github.com/glycerine/abcgo/aig/saig"
	"github.com/glycerine/abcgo/proof/cec"
	"time"
)

func sswInvokeCallback(pars *Ssw_Pars_t) {
	if pars.pFunc == nil {
		return
	}
	switch callback := pars.pFunc.(type) {
	case func(any) int32:
		callback(pars.pData)
	case func(any) int:
		callback(pars.pData)
	case func(any):
		callback(pars.pData)
	default:
		panic("ssw: unsupported callback signature")
	}
}
func Ssw_SignalCorrespondenceRefine(p *Ssw_Man_t) *aig.Graph {
	started := time.Now()
	p.nLitsBeg = Ssw_ClassesLitNum(p.ppClasses)
	p.nNodesBeg = aig.Aig_ManNodeNum(p.pAig)
	p.nRegsBeg = aig.Aig_ManRegNum(p.pAig)
	if p.pPars.fVerbose != 0 {
		fmt.Print("Before BMC: ")
		Ssw_ClassesPrint(p.ppClasses, 0)
	}
	if p.pPars.fLatchCorr == 0 || p.pPars.nFramesK > 1 {
		p.pMSat = Ssw_SatStart(0)
		if p.pPars.fConstrs != 0 {
			Ssw_ManSweepBmcConstr(p)
		} else {
			Ssw_ManSweepBmc(p)
		}
		Ssw_SatStop(p.pMSat)
		p.pMSat = nil
		Ssw_ManCleanup(p)
	}
	if p.pPars.fVerbose != 0 {
		fmt.Print("After  BMC: ")
		Ssw_ClassesPrint(p.ppClasses, 0)
	}
	if p.pPars.pFunc != nil {
		sswInvokeCallback(p.pPars)
		sswInvokeCallback(p.pPars)
	}
	iteration := int32(-1)
	var previousProof, previousSat, previousRecycles, previousFailures int32
	if p.pPars.nStepsMax != 0 {
		for iteration = 0; ; iteration++ {
			if p.pPars.nStepsMax == iteration {
				fmt.Printf("Stopped signal correspondence after %d refinement iterations.\n", iteration)
				break
			}
			if p.pPars.nItersStop >= 0 && p.pPars.nItersStop == iteration {
				speculative := Ssw_SpeculativeReduction(p)
				if err := aig.Aig_ManDumpBlif(speculative, "srm.blif", nil, nil); err != nil {
					fmt.Printf("Cannot write speculative miter: %v\n", err)
				}
				aig.Aig_ManStop(speculative)
				fmt.Printf("Iterative refinement is stopped before iteration %d.\nThe network is reduced using candidate equivalences.\nSpeculatively reduced miter is saved in file %q.\nIf the miter is SAT, the reduced result is incorrect.\n", iteration, "srm.blif")
				break
			}
			start := time.Now()
			p.pMSat = Ssw_SatStart(0)
			var changed int32
			if p.pPars.fLatchCorrOpt != 0 {
				changed = Ssw_ManSweepLatch(p)
			} else if p.pPars.fConstrs != 0 {
				changed = Ssw_ManSweepConstr(p)
			} else if p.pPars.fDynamic != 0 {
				changed = Ssw_ManSweepDyn(p)
			} else {
				changed = Ssw_ManSweep(p)
			}
			if p.pPars.fLatchCorrOpt == 0 {
				p.pPars.nConflicts += int32(p.pMSat.pSat.Statistics().Conflicts)
			}
			if p.pPars.fVerbose != 0 {
				fmt.Print(sswSweepStats(p, iteration, previousProof, previousSat, previousRecycles, previousFailures))
				fmt.Printf("T = %.2f sec\n", time.Since(start).Seconds())
			}
			if p.pPars.fLatchCorrOpt == 0 && p.pPars.fStopWhenGone != 0 && saig.Saig_ManPoNum(p.pAig) == 1 && Ssw_ObjIsConst1Cand(p.pAig, aig.Aig_ObjFanin0(p.pAig, aig.Aig_ManCo(p.pAig, 0))) == 0 {
				fmt.Printf("Iterative refinement is stopped after iteration %d because the property output is no longer a candidate constant.\n", iteration)
				p.nLitsEnd = p.nLitsBeg
				p.nNodesEnd = p.nNodesBeg
				p.nRegsEnd = p.nRegsBeg
				Ssw_SatStop(p.pMSat)
				p.pMSat = nil
				Ssw_ManCleanup(p)
				aig.Aig_ManSetPhase(p.pAig)
				aig.Aig_ManCleanMarkB(p.pAig)
				return aig.Aig_ManDupSimple(p.pAig)
			}
			previousProof, previousSat = p.nSatProof, p.nSatCallsSat
			previousRecycles, previousFailures = p.nRecycles, p.nSatFailsReal
			if p.pMSat.nSatVars > p.nVarsMax {
				p.nVarsMax = p.pMSat.nSatVars
			}
			if p.pMSat.nSolverCalls > p.nCallsMax {
				p.nCallsMax = p.pMSat.nSolverCalls
			}
			Ssw_SatStop(p.pMSat)
			p.pMSat = nil
			Ssw_ManCleanup(p)
			if changed == 0 {
				break
			}
			sswInvokeCallback(p.pPars)
		}
	} else {
		fmt.Println("Stopped signal correspondence after BMC.")
	}
	p.pPars.nIters = iteration + 1
	p.timeTotal = time.Since(started)
	Ssw_ManUpdateEquivs(p, p.pAig, p.pPars.fVerbose)
	result := aig.Aig_ManDupRepr(p.pAig, 0)
	aig.Aig_ManSeqCleanup(result)
	p.nLitsEnd = Ssw_ClassesLitNum(p.ppClasses)
	p.nNodesEnd = aig.Aig_ManNodeNum(result)
	p.nRegsEnd = aig.Aig_ManRegNum(result)
	aig.Aig_ManSetPhase(p.pAig)
	aig.Aig_ManCleanMarkB(p.pAig)
	return result
}
func Ssw_SignalCorrespondence(graph *aig.Graph, pars *Ssw_Pars_t) *aig.Graph {
	if aig.Aig_ManRegNum(graph) <= 0 {
		panic("ssw: signal correspondence requires registers")
	}
	aig.Aig_ManRandom(1)
	if pars == nil {
		pars = new(Ssw_Pars_t)
		Ssw_ManSetDefaultParams(pars)
	}
	// ABC's correspondence entrypoint does not consult TimeLimit.
	if aig.Aig_ManNodeNum(graph) == 0 {
		pars.nIters = 0
		aig.Aig_ManReprStart(graph, aig.Aig_ManObjNumMax(graph))
		return aig.Aig_ManDupOrdered(graph)
	}
	if pars.fLatchCorrOpt != 0 {
		pars.fLatchCorr = 1
		pars.nFramesAddSim = 0
		if len(graph.ClockDomains) > 0 {
			return Ssw_SignalCorrespondencePart(graph, pars)
		}
	} else {
		if pars.nFramesK <= 0 {
			panic("ssw: induction depth must be positive")
		}
		if pars.nPartSize > 0 && pars.nPartSize < aig.Aig_ManRegNum(graph) || len(graph.ClockDomains) > 0 {
			return Ssw_SignalCorrespondencePart(graph, pars)
		}
	}
	if pars.fScorrGia != 0 {
		if pars.fLatchCorrOpt != 0 {
			return cec.Cec_LatchCorrespondence(graph, pars.nBTLimit, pars.fUseCSat)
		}
		return cec.Cec_SignalCorrespondence(graph, pars.nBTLimit, pars.fUseCSat)
	}
	p := Ssw_ManCreate(graph, pars)
	defer Ssw_ManStop(p)
	if pars.fConstrs != 0 {
		p.ppClasses = Ssw_ClassesPrepareSimple(graph, pars.fLatchCorr, pars.nMaxLevs)
		Ssw_ClassesSetData(p.ppClasses, graph, nil, Ssw_SmlObjIsConstBit, Ssw_SmlObjsAreEqualBit)
		status := Ssw_ManSetConstrPhases(graph, pars.nFramesK+1, &p.vInits)
		if status != 0 {
			fmt.Println("Ssw_SignalCorrespondence(): The init state does not satisfy the constraints!")
			pars.fVerbose = 0
			return nil
		}
		Ssw_ManRefineByConstrSim(p)
	} else {
		p.ppClasses = Ssw_ClassesPrepare(graph, pars.nFramesK, pars.fLatchCorr, pars.fConstCorr, pars.fOutputCorr, pars.nMaxLevs, pars.fVerbose)
		frames := int32(1) + pars.nFramesAddSim
		if pars.fLatchCorrOpt != 0 {
			frames = 2
		} else if pars.fDynamic != 0 {
			frames = p.nFrames + pars.nFramesAddSim
		}
		p.pSml = Ssw_SmlStart(graph, 0, frames, 1)
		Ssw_ClassesSetData(p.ppClasses, p.pSml, func(data any, obj aig.Edge) uint32 { return Ssw_SmlObjHashWord(data.(*Ssw_Sml_t), obj) }, func(data any, obj aig.Edge) int32 { return Ssw_SmlObjIsConstWord(data.(*Ssw_Sml_t), obj) }, func(data any, a, b aig.Edge) int32 { return Ssw_SmlObjsAreEqualWord(data.(*Ssw_Sml_t), a, b) })
	}
	if pars.fLocalSim != 0 && p.pSml != nil {
		p.pVisited = make([]int32, Ssw_SmlNumFrames(p.pSml)*aig.Aig_ManObjNumMax(graph))
	}
	result := Ssw_SignalCorrespondenceRefine(p)
	if result != nil && pars.fConstrs != 0 && pars.fVerbose != 0 {
		Ssw_ReportConeReductions(p, graph, result)
	}
	return result
}

// sswSweepStats reports per-iteration changes, as in Ssw_SignalCorrespondenceRefine.
func sswSweepStats(p *Ssw_Man_t, iteration, previousProof, previousSat, previousRecycles, previousFailures int32) string {
	result := fmt.Sprintf("%3d : C =%7d. Cl =%7d. ", iteration, Ssw_ClassesCand1Num(p.ppClasses), Ssw_ClassesClassNum(p.ppClasses))
	if p.pPars.fLatchCorrOpt != 0 {
		return result + fmt.Sprintf("Pr =%6d. Cex =%5d. R =%4d. F =%4d. ", p.nSatProof-previousProof, p.nSatCallsSat-previousSat, p.nRecycles-previousRecycles, p.nSatFailsReal-previousFailures)
	}
	result += fmt.Sprintf("LR =%6d. NR =%6d. ", p.nConstrReduced, aig.Aig_ManNodeNum(p.pFrames))
	if p.pPars.fDynamic != 0 {
		result += fmt.Sprintf("Cex =%5d. R =%4d. ", p.nSatCallsSat-previousSat, p.nRecycles-previousRecycles)
	}
	property := "-"
	if saig.Saig_ManPoNum(p.pAig) == 1 && Ssw_ObjIsConst1Cand(p.pAig, aig.Aig_ObjFanin0(p.pAig, aig.Aig_ManCo(p.pAig, 0))) != 0 {
		property = "+"
	}
	return result + fmt.Sprintf("F =%5d. %s ", p.nSatFailsReal-previousFailures, property)
}

func Ssw_LatchCorrespondence(graph *aig.Graph, pars *Ssw_Pars_t) *aig.Graph {
	if pars == nil {
		pars = new(Ssw_Pars_t)
		Ssw_ManSetDefaultParamsLcorr(pars)
	}
	return Ssw_SignalCorrespondence(graph, pars)
}
