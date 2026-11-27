package abs

import (
	"fmt"

	"github.com/glycerine/abcgo/aig/gia"
	abcmain "github.com/glycerine/abcgo/base/main"
	"github.com/glycerine/abcgo/misc/extra"
	"github.com/glycerine/abcgo/misc/util"
	"github.com/glycerine/abcgo/misc/vec"
	"github.com/glycerine/abcgo/sat/bsat"
)

// Gia_ManPerformGlaOld performs gate-level abstraction with the earlier
// Gla_Man_t engine (absGlaOld.c; &gla -n).
func Gia_ManPerformGlaOld(pAig *gia.Gia_Man_t, pPars *Abs_Par_t, fStartVta int32) int32 {
	// Go-only: XOR/MUX nodes are lowered to ANDs first (Gia_ManToAigSimple
	// in Gla_ManStart treats them as ANDs)
	if glaOldHasSpecialGates(pAig) {
		return glaOldPerformLowered(pAig, pPars, fStartVta)
	}
	var p *Gla_Man_t
	var vPPis, vCore *vec.Vec_Int_t
	var pCex *util.Abc_Cex_t
	var f, i, iPrev, nConfls, Status, nVarsOld, nCoreSize, RetValue int32
	fOneIsSent := false
	RetValue = -1
	var clk2 int64
	clk := util.Abc_Clock()
	if pPars.NTimeOut != 0 {
		defer absPinClockThread()()
	}
	// preconditions
	assert(gia.Gia_ManPoNum(pAig) == 1, "Gia_ManPoNum(pAig) == 1")
	assert(pPars.NFramesMax == 0 || pPars.NFramesStart <= pPars.NFramesMax, "pPars->nFramesMax == 0 || pPars->nFramesStart <= pPars->nFramesMax")
	if gia.Gia_ObjIsConst0(gia.Gia_ObjFanin0(gia.Gia_ManPo(pAig, 0))) != 0 {
		if gia.Gia_ObjFaninC0(gia.Gia_ManPo(pAig, 0)) == 0 {
			fmt.Printf("Sequential miter is trivially UNSAT.\n")
			return 1
		}
		pAig.ClearSequentialCounterexample()
		if err := pAig.SetSequentialCounterexample(util.Abc_CexMakeTriv(gia.Gia_ManRegNum(pAig), gia.Gia_ManPiNum(pAig), 1, 0)); err != nil {
			panic(err)
		}
		fmt.Printf("Sequential miter is trivially SAT.\n")
		return 0
	}
	// an output driven by a primary input has no SAT variable
	// (Gla_ManGetOutLit would read an empty vFrames; CBUGS_FIXED.md #33)
	if gia.Gia_ObjIsPi(pAig, gia.Gia_ObjFanin0(gia.Gia_ManPo(pAig, 0))) != 0 {
		pAig.ClearSequentialCounterexample()
		pCex := util.Abc_CexAlloc(gia.Gia_ManRegNum(pAig), gia.Gia_ManPiNum(pAig), 1)
		if gia.Gia_ObjFaninC0(gia.Gia_ManPo(pAig, 0)) == 0 {
			util.Abc_InfoSetBit(pCex.PData, pCex.NRegs+gia.Gia_ObjCioId(gia.Gia_ObjFanin0(gia.Gia_ManPo(pAig, 0))))
		}
		if err := pAig.SetSequentialCounterexample(pCex); err != nil {
			panic(err)
		}
		fmt.Printf("Sequential miter is trivially SAT.\n")
		return 0
	}

	// compute intial abstraction
	if gia.Gia_ManGateClasses(pAig) == nil {
		if fStartVta != 0 {
			nFramesMaxOld := pPars.NFramesMax
			nFramesStartOld := pPars.NFramesStart
			nTimeOutOld := pPars.NTimeOut
			nDumpOld := pPars.FDumpVabs
			pPars.NFramesMax = pPars.NFramesStart
			pPars.NFramesStart = min(pPars.NFramesStart/2+1, 3)
			pPars.NTimeOut = 20
			pPars.FDumpVabs = 0
			RetValue = Gia_VtaPerformInt(pAig, pPars)
			pPars.NFramesMax = nFramesMaxOld
			pPars.NFramesStart = nFramesStartOld
			pPars.NTimeOut = nTimeOutOld
			pPars.FDumpVabs = nDumpOld
			// create gate classes
			gia.Gia_ManSetGateClasses(pAig, nil)
			if classes := gia.Gia_ManObjClasses(pAig); classes != nil {
				gia.Gia_ManSetGateClasses(pAig, Gia_VtaConvertToGla(pAig, classes))
			}
			gia.Gia_ManSetObjClasses(pAig, nil)
			// return if VTA solve the problem if could not start
			if RetValue == 0 || gia.Gia_ManGateClasses(pAig) == nil {
				return RetValue
			}
		} else {
			classes := vec.Vec_IntStart(int(gia.Gia_ManObjNum(pAig)))
			vec.Vec_IntWriteEntry(classes, 0, 1)
			vec.Vec_IntWriteEntry(classes, int(gia.Gia_ObjFaninId0p(pAig, gia.Gia_ManPo(pAig, 0))), 1)
			gia.Gia_ManSetGateClasses(pAig, classes)
		}
	}
	// start the manager
	p = Gla_ManStart(pAig, pPars)
	p.timeInit = util.Abc_Clock() - clk
	// set runtime limit
	if p.pPars.NTimeOut != 0 {
		bsat.Sat_solver2_set_runtime_limit(p.pSat, absCPUClock, int64(p.pPars.NTimeOut)*util.CLOCKS_PER_SEC+absCPUClock())
	}
	// perform initial abstraction
	if p.pPars.FVerbose != 0 {
		util.Abc_Print(1, "Running gate-level abstraction (GLA) with the following parameters:\n")
		util.Abc_Print(1, "FrameMax = %d  ConfMax = %d  Timeout = %d  RatioMin = %d %%.\n",
			pPars.NFramesMax, pPars.NConfLimit, pPars.NTimeOut, pPars.NRatioMin)
		util.Abc_Print(1, "LearnStart = %d  LearnDelta = %d  LearnRatio = %d %%.\n",
			pPars.NLearnedStart, pPars.NLearnedDelta, pPars.NLearnedPerce)
		util.Abc_Print(1, " Frame   %%   Abs  PPI   FF   LUT   Confl  Cex   Vars   Clas   Lrns     Time        Mem\n")
	}
	for f, i, iPrev = 0, 0, 0; p.pPars.NFramesMax == 0 || f < p.pPars.NFramesMax; f, iPrev = f+1, i {
		nConflsBeg := bsat.Sat_solver2_nconflicts(p.pSat)
		p.pPars.IFrame = f

		// load timeframe
		Gia_GlaAddTimeFrame(p, f)

		// iterate as long as there are counter-examples
		for i = 0; ; i++ {
			clk2 = util.Abc_Clock()
			vCore = Gla_ManUnsatCore(p, f, p.pSat, pPars.NConfLimit, pPars.FVerbose, &Status, &nConfls)
			if Status == -1 || (bsat.Sat_solver2_RuntimeLimit(p.pSat) != 0 && absCPUClock() > bsat.Sat_solver2_RuntimeLimit(p.pSat)) { // resource limit is reached
				bsat.Sat_solver2_StopPrf2(p.pSat)
				goto finish
			}
			if Status == 1 {
				bsat.Sat_solver2_StopPrf2(p.pSat)
				p.timeUnsat += util.Abc_Clock() - clk2
				break
			}
			p.timeSat += util.Abc_Clock() - clk2
			assert(Status == 0, "Status == 0")
			p.nCexes++

			// cancel old one if it was sent
			if abcmain.Abc_FrameIsBridgeMode() != 0 && fOneIsSent {
				Gia_GlaSendCancel(p, pPars.FVerbose)
				fOneIsSent = false
			}

			// perform the refinement
			clk2 = util.Abc_Clock()
			vPPis = nil
			if pPars.FAddLayer != 0 {
				vPPis = Gla_ManCollectPPis(p, nil)
				// if all pseudo-PIs are PIs, the abstraction is complete: check
				// the CEX below (CBUGS_FIXED.md #35)
				if vec.Vec_IntSize(vPPis) == 0 {
					vPPis = nil
				}
			}
			if vPPis == nil {
				vPPis = Gla_ManRefinement(p)
				if vPPis == nil {
					bsat.Sat_solver2_StopPrf2(p.pSat)
					pCex = p.pGia.TakeSequentialCounterexample()
					break
				}
			}
			assert(pCex == nil, "pCex == NULL")

			// start proof logging
			if i == 0 {
				// create bookmark to be used for rollback
				bsat.Sat_solver2_bookmark(p.pSat)
				vec.Vec_IntClear(p.vAddedNew)
				p.nAbsOld = int32(vec.Vec_IntSize(p.vAbs))
				nVarsOld = p.nSatVars

				// start incremental proof manager
				assert(bsat.Sat_solver2_Prf2(p.pSat) == nil, "p->pSat->pPrf2 == NULL")
				if !bsat.Sat_solver2_HasPrf1(p.pSat) {
					bsat.Sat_solver2_SetPrf2(p.pSat, bsat.Prf_ManAlloc())
				}
				if bsat.Sat_solver2_Prf2(p.pSat) != nil {
					p.nProofIds = 0
					n := int(gia.Gia_ManObjNum(p.pGia))
					p.vProofIds = make([]int32, n)
					for k := range p.vProofIds {
						p.vProofIds[k] = -1
					}
					bsat.Prf_ManRestart(bsat.Sat_solver2_Prf2(p.pSat), &p.vProofIds, bsat.Sat_solver2_nlearnts(p.pSat), int32(vec.Vec_IntSize(vPPis)))
				}
			} else {
				// resize the proof logger
				if bsat.Sat_solver2_Prf2(p.pSat) != nil {
					bsat.Prf_ManGrow(bsat.Sat_solver2_Prf2(p.pSat), p.nProofIds+int32(vec.Vec_IntSize(vPPis)))
				}
			}

			Gia_GlaAddToAbs(p, vPPis, 1)
			Gia_GlaAddOneSlice(p, f, vPPis)

			// print the result (do not count it towards change)
			if p.pPars.FVerbose != 0 {
				Gla_ManAbsPrintFrame(p, -1, f+1, bsat.Sat_solver2_nconflicts(p.pSat)-nConflsBeg, i, util.Abc_Clock()-clk)
			}
		}
		if pCex != nil {
			break
		}
		assert(Status == 1, "Status == 1")

		// valid core is obtained
		nCoreSize = 1
		if vCore != nil {
			nCoreSize += int32(vec.Vec_IntSize(vCore))
			Gia_GlaAddToCounters(p, vCore)
		}
		if i == 0 {
			p.pPars.nFramesNoChange++
			vCore = nil
		} else {
			p.pPars.nFramesNoChange = 0
			// update the SAT solver
			bsat.Sat_solver2_rollback(p.pSat)
			// update storage
			Gla_ManRollBack(p)
			p.nSatVars = nVarsOld
			// load this timeframe
			Gia_GlaAddToAbs(p, vCore, 0)
			Gia_GlaAddOneSlice(p, f, vCore)
			// run SAT solver
			clk2 = util.Abc_Clock()
			vCore = Gla_ManUnsatCore(p, f, p.pSat, pPars.NConfLimit, p.pPars.FVerbose, &Status, &nConfls)
			p.timeUnsat += util.Abc_Clock() - clk2
			vCore = nil
			if Status == -1 { // resource limit is reached
				break
			}
			if Status == 0 {
				break
			}
		}
		// print the result
		if p.pPars.FVerbose != 0 {
			Gla_ManAbsPrintFrame(p, nCoreSize, f+1, bsat.Sat_solver2_nconflicts(p.pSat)-nConflsBeg, i, util.Abc_Clock()-clk)
		}

		if f > 2 && iPrev > 0 && i == 0 { // change has happened
			if abcmain.Abc_FrameIsBridgeMode() != 0 {
				// cancel old one if it was sent
				if fOneIsSent {
					Gia_GlaSendCancel(p, pPars.FVerbose)
				}
				// send new one
				Gia_GlaSendAbsracted(p, pPars.FVerbose)
				fOneIsSent = true
			}

			// dump the model into file
			if p.pPars.FDumpVabs != 0 {
				abcmain.Abc_FrameSetStatus(-1)
				abcmain.Abc_FrameSetCex(nil)
				abcmain.Abc_FrameSetNFrames(f + 1)
				name := p.pPars.PFileVabs
				if name == "" {
					name = "glabs.aig"
				}
				abcmain.Cmd_CommandExecute(abcmain.Abc_FrameGetGlobalFrame(), fmt.Sprintf("write_status %s", extra.Extra_FileNameGenericAppend(name, ".status")))
				Gia_GlaDumpAbsracted(p, pPars.FVerbose)
			}
		}

		// check if the number of objects is below limit
		if Gia_GlaAbsCount(p, 0, 0) >= (p.nObjs-1)*(100-pPars.NRatioMin)/100 {
			Status = -1
			break
		}
	}
finish:
	// analize the results
	if pCex == nil {
		if p.pPars.FVerbose != 0 && Status == -1 {
			fmt.Printf("\n")
		}
		gia.Gia_ManSetGateClasses(pAig, Gla_ManTranslate(p))
		if Status == -1 {
			if p.pPars.NTimeOut != 0 && absCPUClock() >= bsat.Sat_solver2_RuntimeLimit(p.pSat) {
				util.Abc_Print(1, "Timeout %d sec in frame %d with a %d-stable abstraction.    ", p.pPars.NTimeOut, f, p.pPars.nFramesNoChange)
			} else if pPars.NConfLimit != 0 && bsat.Sat_solver2_nconflicts(p.pSat) >= pPars.NConfLimit {
				util.Abc_Print(1, "Exceeded %d conflicts in frame %d with a %d-stable abstraction.  ", pPars.NConfLimit, f, p.pPars.nFramesNoChange)
			} else if Gia_GlaAbsCount(p, 0, 0) >= (p.nObjs-1)*(100-pPars.NRatioMin)/100 {
				util.Abc_Print(1, "The ratio of abstracted objects is less than %d %% in frame %d.  ", pPars.NRatioMin, f)
			} else {
				util.Abc_Print(1, "Abstraction stopped for unknown reason in frame %d.  ", f)
			}
		} else {
			p.pPars.IFrame++
			util.Abc_Print(1, "GLA completed %d frames with a %d-stable abstraction.  ", f, p.pPars.nFramesNoChange)
		}
	} else {
		if p.pPars.FVerbose != 0 {
			fmt.Printf("\n")
		}
		pAig.ClearSequentialCounterexample()
		if err := pAig.SetSequentialCounterexample(pCex); err != nil {
			panic(err)
		}
		if gia.Gia_ManVerifyCex(pAig, pCex, 0) == 0 {
			util.Abc_Print(1, "    Gia_ManPerformGlaOld(): CEX verification has failed!\n")
		}
		util.Abc_Print(1, "Counter-example detected in frame %d.  ", f)
		p.pPars.IFrame = pCex.IFrame - 1
		gia.Gia_ManSetGateClasses(pAig, nil)
		RetValue = 0
	}
	util.Abc_PrintTime(1, "Time", util.Abc_Clock()-clk)
	if p.pPars.FVerbose != 0 {
		p.timeOther = (util.Abc_Clock() - clk) - p.timeUnsat - p.timeSat - p.timeCex - p.timeInit
		util.Abc_PrintTimeP(1, "Runtime: Initializing", p.timeInit, util.Abc_Clock()-clk)
		util.Abc_PrintTimeP(1, "Runtime: Solver UNSAT", p.timeUnsat, util.Abc_Clock()-clk)
		util.Abc_PrintTimeP(1, "Runtime: Solver SAT  ", p.timeSat, util.Abc_Clock()-clk)
		util.Abc_PrintTimeP(1, "Runtime: Refinement  ", p.timeCex, util.Abc_Clock()-clk)
		util.Abc_PrintTimeP(1, "Runtime: Other       ", p.timeOther, util.Abc_Clock()-clk)
		util.Abc_PrintTimeP(1, "Runtime: TOTAL       ", util.Abc_Clock()-clk, util.Abc_Clock()-clk)
		Gla_ManReportMemory(p)
	}
	Gla_ManStop(p)
	return RetValue
}
