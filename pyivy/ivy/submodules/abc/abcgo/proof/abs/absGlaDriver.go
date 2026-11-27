package abs

import (
	"github.com/glycerine/abcgo/aig/gia"
	abcmain "github.com/glycerine/abcgo/base/main"
	"github.com/glycerine/abcgo/misc/util"
	"github.com/glycerine/abcgo/misc/vec"
	"github.com/glycerine/abcgo/sat/bsat"
)

// Gia_ManPerformGla performs gate-level abstraction (absGla.c). It returns
// 0 for a counter-example (stored in the GIA), 1 when the abstraction was
// proved by a concurrent prover, and -1 otherwise; the abstraction is left
// in the GIA's gate classes.
func Gia_ManPerformGla(pAig *gia.Gia_Man_t, pPars *Abs_Par_t) int32 {
	fUseSecondCore := true
	var p *Ga2_Man_t
	var vCore, vPPis *vec.Vec_Int_t
	var clk2 int64
	// Go-only: XOR/MUX nodes are lowered to ANDs first (absGla.c's cut
	// markup treats them as ANDs and computes wrong functions)
	if glaOldHasSpecialGates(pAig) {
		return glaPerformLowered(pAig, pPars, false, func(lowered *gia.Gia_Man_t) int32 {
			return Gia_ManPerformGla(lowered, pPars)
		})
	}
	clk := util.Abc_Clock()
	Status := int32(bsat.Unknown)
	RetValue := int32(-1)
	iFrameTryToProve := int32(-1)
	fOneIsSent := false
	var i, c, f, Lit int32
	if pPars.NTimeOut != 0 {
		defer absPinClockThread()()
	}
	pPars.IFrame = -1
	// check trivial case
	assert(gia.Gia_ManPoNum(pAig) == 1, "Gia_ManPoNum(pAig) == 1")
	pAig.ClearSequentialCounterexample()
	if gia.Gia_ObjIsConst0(gia.Gia_ObjFanin0(gia.Gia_ManPo(pAig, 0))) != 0 {
		if gia.Gia_ObjFaninC0(gia.Gia_ManPo(pAig, 0)) == 0 {
			util.Abc_Print(1, "Sequential miter is trivially UNSAT.\n")
			return 1
		}
		if err := pAig.SetSequentialCounterexample(util.Abc_CexMakeTriv(gia.Gia_ManRegNum(pAig), gia.Gia_ManPiNum(pAig), 1, 0)); err != nil {
			panic(err)
		}
		util.Abc_Print(1, "Sequential miter is trivially SAT.\n")
		return 0
	}
	// an output driven by a primary input has no abstraction ID
	// (Ga2_ObjFindLit would read vIds[-1]; CBUGS_FIXED.md #33)
	if gia.Gia_ObjIsPi(pAig, gia.Gia_ObjFanin0(gia.Gia_ManPo(pAig, 0))) != 0 {
		pCex := util.Abc_CexAlloc(gia.Gia_ManRegNum(pAig), gia.Gia_ManPiNum(pAig), 1)
		if gia.Gia_ObjFaninC0(gia.Gia_ManPo(pAig, 0)) == 0 {
			util.Abc_InfoSetBit(pCex.PData, pCex.NRegs+gia.Gia_ObjCioId(gia.Gia_ObjFanin0(gia.Gia_ManPo(pAig, 0))))
		}
		if err := pAig.SetSequentialCounterexample(pCex); err != nil {
			panic(err)
		}
		util.Abc_Print(1, "Sequential miter is trivially SAT.\n")
		return 0
	}
	// create gate classes if not given
	if gia.Gia_ManGateClasses(pAig) == nil {
		classes := vec.Vec_IntStart(int(gia.Gia_ManObjNum(pAig)))
		vec.Vec_IntWriteEntry(classes, 0, 1)
		vec.Vec_IntWriteEntry(classes, int(gia.Gia_ObjFaninId0p(pAig, gia.Gia_ManPo(pAig, 0))), 1)
		gia.Gia_ManSetGateClasses(pAig, classes)
	}
	// start the manager
	p = Ga2_ManStart(pAig, pPars)
	p.timeInit = util.Abc_Clock() - clk
	// perform initial abstraction
	if p.pPars.FVerbose != 0 {
		util.Abc_Print(1, "Running gate-level abstraction (GLA) with the following parameters:\n")
		util.Abc_Print(1, "FrameMax = %d  ConfMax = %d  Timeout = %d  Limit = %d %%  Limit2 = %d %%  RatioMax = %d %%\n",
			pPars.NFramesMax, pPars.NConfLimit, pPars.NTimeOut, pPars.NRatioMin, pPars.NRatioMin2, pPars.NRatioMax)
		util.Abc_Print(1, "LrnStart = %d  LrnDelta = %d  LrnRatio = %d %%  Skip = %d  SimpleCNF = %d  Dump = %d\n",
			pPars.NLearnedStart, pPars.NLearnedDelta, pPars.NLearnedPerce, pPars.FUseSkip, pPars.FUseSimple, pPars.FDumpVabs|pPars.FDumpMabs)
		if pPars.FDumpVabs != 0 || pPars.FDumpMabs != 0 {
			what := "Miter with abstraction map"
			if pPars.FDumpVabs != 0 {
				what = "Abstracted model"
			}
			util.Abc_Print(1, "%s will be continously dumped into file \"%s\".\n", what, Ga2_GlaGetFileName(p, pPars.FDumpVabs))
		}
		if pPars.FDumpMabs != 0 {
			Ga2_ManWriteStatus(p, -1)
			// create trivial abstraction map
			Ga2_ManDumpInitialMap(p)
		}
		util.Abc_Print(1, " Frame   %%   Abs  PPI   FF   LUT   Confl  Cex   Vars   Clas   Lrns     Time        Mem\n")
	}
	// iterate unrolling
	for i, f = 0, 0; pPars.NFramesMax == 0 || f < pPars.NFramesMax; i++ {
		var nAbsOld int32
		// remember the timeframe
		p.pPars.IFrame = -1
		// create new SAT solver
		Ga2_ManRestart(p)
		// remember abstraction size after the last restart
		nAbsOld = int32(vec.Vec_IntSize(p.vAbs))
		// unroll the circuit
		for f = 0; pPars.NFramesMax == 0 || f < pPars.NFramesMax; f++ {
			// remember current limits
			nConflsBeg := bsat.Sat_solver2_nconflicts(p.pSat)
			nAbs := int32(vec.Vec_IntSize(p.vAbs))
			nValues := int32(vec.Vec_IntSize(p.vValues))
			var nVarsOld int32
			// remember the timeframe
			p.pPars.IFrame = f
			// extend and clear storage
			if int(f) == len(p.vId2Lit) {
				p.vId2Lit = append(p.vId2Lit, vec.Vec_IntAlloc(0))
			}
			vec.Vec_IntFillExtra(Ga2_MapFrameMap(p, f), vec.Vec_IntSize(p.vValues), -1)
			// add static clauses to this timeframe
			Ga2_ManAddAbsClauses(p, f)
			// skip checking if skipcheck is enabled (&gla -s)
			if p.pPars.FUseSkip != 0 && f <= p.pPars.iFrameProved {
				continue
			}
			// skip checking if we need to skip several starting frames (&gla -S <num>)
			if p.pPars.NFramesStart != 0 && f <= p.pPars.NFramesStart {
				continue
			}
			// get the output literal
			Lit = Ga2_ObjFindLit(p, gia.Gia_ObjFanin0(gia.Gia_ManPo(pAig, 0)), f)
			assert(Lit >= 0, "Lit >= 0")
			Lit = util.Abc_LitNotCond(Lit, gia.Gia_ObjFaninC0(gia.Gia_ManPo(pAig, 0)))
			if Lit == 0 {
				continue
			}
			assert(Lit > 1, "Lit > 1")
			// check for counter-examples
			if p.nSatVars > bsat.Sat_solver2_nvars(p.pSat) {
				bsat.Sat_solver2_setnvars(p.pSat, p.nSatVars)
			}
			nVarsOld = p.nSatVars
			for c = 0; ; c++ {
				// consider the special case when the target literal is implied false
				// by implications which happened as a result of previous refinements
				// note that incremental UNSAT core cannot be computed because there is no learned clauses
				// in this case, we will assume that UNSAT core cannot reduce the problem
				if bsat.Var_is_assigned(p.pSat, util.Abc_Lit2Var(Lit)) != 0 {
					bsat.Sat_solver2_StopPrf2(p.pSat)
					break
				}
				// perform SAT solving
				clk2 = util.Abc_Clock()
				Status = bsat.Sat_solver2_solve(p.pSat, []int32{Lit}, int64(pPars.NConfLimit), 0, 0, 0)
				if Status == int32(bsat.Satisfiable) { // perform refinement
					p.nCexes++
					p.timeSat += util.Abc_Clock() - clk2

					// cancel old one if it was sent
					if abcmain.Abc_FrameIsBridgeMode() != 0 && fOneIsSent {
						Gia_Ga2SendCancel(p, pPars.FVerbose)
						fOneIsSent = false
					}
					if iFrameTryToProve >= 0 {
						Gia_GlaProveCancel(pPars.FVerbose)
						iFrameTryToProve = -1
					}

					// perform refinement
					clk2 = util.Abc_Clock()
					Rnm_ManSetRefId(p.pRnm, c)
					vPPis = Ga2_ManRefine(p)
					p.timeCex += util.Abc_Clock() - clk2
					if vPPis == nil {
						if pPars.FVerbose != 0 {
							Ga2_ManAbsPrintFrame(p, f, bsat.Sat_solver2_nconflicts(p.pSat)-nConflsBeg, c, util.Abc_Clock()-clk, 1)
						}
						goto finish
					}

					if c == 0 {
						// create bookmark to be used for rollback
						assert(nVarsOld == bsat.Sat_solver2_Size(p.pSat), "nVarsOld == p->pSat->size")
						bsat.Sat_solver2_bookmark(p.pSat)
						// start incremental proof manager
						assert(bsat.Sat_solver2_Prf2(p.pSat) == nil, "p->pSat->pPrf2 == NULL")
						bsat.Sat_solver2_SetPrf2(p.pSat, bsat.Prf_ManAlloc())
						if bsat.Sat_solver2_Prf2(p.pSat) != nil {
							p.nProofIds = 0
							ga2FillProofIds(p)
							bsat.Prf_ManRestart(bsat.Sat_solver2_Prf2(p.pSat), &p.vProofIds, bsat.Sat_solver2_nlearnts(p.pSat), int32(vec.Vec_IntSize(vPPis)))
						}
					} else {
						// resize the proof logger
						if bsat.Sat_solver2_Prf2(p.pSat) != nil {
							bsat.Prf_ManGrow(bsat.Sat_solver2_Prf2(p.pSat), p.nProofIds+int32(vec.Vec_IntSize(vPPis)))
						}
					}

					Ga2_ManAddToAbs(p, vPPis)
					if pPars.FVerbose != 0 {
						Ga2_ManAbsPrintFrame(p, f, bsat.Sat_solver2_nconflicts(p.pSat)-nConflsBeg, c+1, util.Abc_Clock()-clk, 0)
					}
					// check if the number of objects is below limit
					if pPars.NRatioMin2 != 0 && int32(vec.Vec_IntSize(p.vAbs)) >= p.nMarked*pPars.NRatioMin2/100 {
						Status = int32(bsat.Unknown)
						goto finish
					}
					continue
				}
				p.timeUnsat += util.Abc_Clock() - clk2
				if Status == int32(bsat.Unknown) { // ran out of resources
					goto finish
				}
				if bsat.Sat_solver2_RuntimeLimit(p.pSat) != 0 && absCPUClock() > bsat.Sat_solver2_RuntimeLimit(p.pSat) { // timeout
					goto finish
				}
				if c == 0 {
					if f > p.pPars.iFrameProved {
						p.pPars.nFramesNoChange++
					}
					break
				}
				if f > p.pPars.iFrameProved {
					p.pPars.nFramesNoChange = 0
				}

				// derive the core
				assert(bsat.Sat_solver2_Prf2(p.pSat) != nil, "p->pSat->pPrf2 != NULL")
				vCore = ga2ProofCore(p)
				bsat.Sat_solver2_StopPrf2(p.pSat)
				// update the SAT solver
				bsat.Sat_solver2_rollback(p.pSat)
				// reduce abstraction
				Ga2_ManShrinkAbs(p, nAbs, nValues, nVarsOld)

				// purify UNSAT core
				if fUseSecondCore {
					// create bookmark to be used for rollback
					assert(nVarsOld == bsat.Sat_solver2_Size(p.pSat), "nVarsOld == p->pSat->size")
					bsat.Sat_solver2_bookmark(p.pSat)
					// start incremental proof manager
					assert(bsat.Sat_solver2_Prf2(p.pSat) == nil, "p->pSat->pPrf2 == NULL")
					bsat.Sat_solver2_SetPrf2(p.pSat, bsat.Prf_ManAlloc())
					if bsat.Sat_solver2_Prf2(p.pSat) != nil {
						p.nProofIds = 0
						ga2FillProofIds(p)
						bsat.Prf_ManRestart(bsat.Sat_solver2_Prf2(p.pSat), &p.vProofIds, bsat.Sat_solver2_nlearnts(p.pSat), int32(vec.Vec_IntSize(vCore)))

						Ga2_ManAddToAbs(p, vCore)
					}
					// run SAT solver
					clk2 = util.Abc_Clock()
					Status = bsat.Sat_solver2_solve(p.pSat, []int32{Lit}, int64(pPars.NConfLimit), 0, 0, 0)
					if Status == int32(bsat.Unknown) {
						goto finish
					}
					assert(Status == int32(bsat.Unsatisfiable), "Status == l_False")
					p.timeUnsat += util.Abc_Clock() - clk2

					// derive the core
					assert(bsat.Sat_solver2_Prf2(p.pSat) != nil, "p->pSat->pPrf2 != NULL")
					vCore = ga2ProofCore(p)
					bsat.Sat_solver2_StopPrf2(p.pSat)
					// update the SAT solver
					bsat.Sat_solver2_rollback(p.pSat)
					// reduce abstraction
					Ga2_ManShrinkAbs(p, nAbs, nValues, nVarsOld)
				}

				Ga2_ManAddToAbs(p, vCore)
				break
			}
			// remember the last proved frame
			if p.pPars.iFrameProved < f {
				p.pPars.iFrameProved = f
			}
			// print statistics
			if pPars.FVerbose != 0 {
				Ga2_ManAbsPrintFrame(p, f, bsat.Sat_solver2_nconflicts(p.pSat)-nConflsBeg, c, util.Abc_Clock()-clk, 1)
			}
			// check if abstraction was proved
			if Gia_GlaProveCheck(pPars.FVerbose) != 0 {
				RetValue = 1
				goto finish
			}
			if c > 0 {
				if p.pPars.FVeryVerbose != 0 {
					util.Abc_Print(1, "\n")
				}
				// recompute the abstraction
				gia.Gia_ManSetGateClasses(pAig, Ga2_ManAbsTranslate(p))
				// check if the number of objects is below limit
				if pPars.NRatioMin != 0 && int32(vec.Vec_IntSize(p.vAbs)) >= p.nMarked*pPars.NRatioMin/100 {
					Status = int32(bsat.Unknown)
					goto finish
				}
			}
			// check the number of stable frames
			if p.pPars.nFramesNoChange == p.pPars.NFramesNoChangeLim {
				// dump the model into file
				if p.pPars.FDumpVabs != 0 || p.pPars.FDumpMabs != 0 {
					Ga2_ManWriteStatus(p, f)
					Ga2_GlaDumpAbsracted(p, pPars.FVerbose)
				}
				// call the prover
				if p.pPars.FCallProver != 0 {
					// cancel old one if it is proving
					if iFrameTryToProve >= 0 {
						Gia_GlaProveCancel(pPars.FVerbose)
					}
					// prove new one
					Gia_GlaProveAbsracted(pAig, pPars.FSimpProver, pPars.FVeryVerbose)
					iFrameTryToProve = f
					p.nPdrCalls++
				}
				// speak to the bridge
				if abcmain.Abc_FrameIsBridgeMode() != 0 {
					// cancel old one if it was sent
					if fOneIsSent {
						Gia_Ga2SendCancel(p, pPars.FVerbose)
					}
					// send new one
					Gia_Ga2SendAbsracted(p, pPars.FVerbose)
					fOneIsSent = true
				}
			}
			// if abstraction grew more than a certain percentage, force a restart
			if pPars.NRatioMax == 0 {
				continue
			}
			if c > 0 && (f > 20 || vec.Vec_IntSize(p.vAbs) > 100) && int32(vec.Vec_IntSize(p.vAbs))-nAbsOld >= nAbsOld*pPars.NRatioMax/100 {
				if p.pPars.FVerbose != 0 {
					util.Abc_Print(1, "Forcing restart because abstraction grew from %d to %d (more than %d %%).\n",
						nAbsOld, vec.Vec_IntSize(p.vAbs), pPars.NRatioMax)
				}
				break
			}
		}
	}
finish:
	bsat.Sat_solver2_StopPrf2(p.pSat)
	// cancel old one if it is proving
	if iFrameTryToProve >= 0 {
		Gia_GlaProveCancel(pPars.FVerbose)
	}
	// analize the results
	if p.fUseNewLine == 0 {
		util.Abc_Print(1, "\n")
	}
	if RetValue == 1 {
		util.Abc_Print(1, "GLA completed %d frames and proved abstraction derived in frame %d  ", p.pPars.iFrameProved+1, iFrameTryToProve)
	} else if pAig.SequentialCounterexample() == nil {
		gia.Gia_ManSetGateClasses(pAig, Ga2_ManAbsTranslate(p))
		if p.pPars.NTimeOut != 0 && absCPUClock() >= bsat.Sat_solver2_RuntimeLimit(p.pSat) {
			util.Abc_Print(1, "GLA reached timeout %d sec in frame %d with a %d-stable abstraction.    ", p.pPars.NTimeOut, p.pPars.iFrameProved+1, p.pPars.nFramesNoChange)
		} else if pPars.NConfLimit != 0 && bsat.Sat_solver2_nconflicts(p.pSat) >= pPars.NConfLimit {
			util.Abc_Print(1, "GLA exceeded %d conflicts in frame %d with a %d-stable abstraction.  ", pPars.NConfLimit, p.pPars.iFrameProved+1, p.pPars.nFramesNoChange)
		} else if pPars.NRatioMin2 != 0 && int32(vec.Vec_IntSize(p.vAbs)) >= p.nMarked*pPars.NRatioMin2/100 {
			util.Abc_Print(1, "GLA found that the size of abstraction exceeds %d %% in frame %d during refinement.  ", pPars.NRatioMin2, p.pPars.iFrameProved+1)
		} else if pPars.NRatioMin != 0 && int32(vec.Vec_IntSize(p.vAbs)) >= p.nMarked*pPars.NRatioMin/100 {
			util.Abc_Print(1, "GLA found that the size of abstraction exceeds %d %% in frame %d.  ", pPars.NRatioMin, p.pPars.iFrameProved+1)
		} else {
			util.Abc_Print(1, "GLA finished %d frames and produced a %d-stable abstraction.  ", p.pPars.iFrameProved+1, p.pPars.nFramesNoChange)
		}
		p.pPars.IFrame = p.pPars.iFrameProved
	} else {
		if p.pPars.FVerbose != 0 {
			util.Abc_Print(1, "\n")
		}
		if gia.Gia_ManVerifyCex(pAig, pAig.SequentialCounterexample(), 0) == 0 {
			util.Abc_Print(1, "    Gia_ManPerformGlaOld(): CEX verification has failed!\n")
		}
		util.Abc_Print(1, "True counter-example detected in frame %d.  ", f)
		p.pPars.IFrame = f - 1
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
		Ga2_ManReportMemory(p)
	}
	Ga2_ManStop(p)
	return RetValue
}

// ga2FillProofIds is Vec_IntFill( p->vProofIds, Gia_ManObjNum(p->pGia), -1 ).
func ga2FillProofIds(p *Ga2_Man_t) {
	n := int(gia.Gia_ManObjNum(p.pGia))
	if cap(p.vProofIds) < n {
		p.vProofIds = make([]int32, n)
	}
	p.vProofIds = p.vProofIds[:n]
	for k := range p.vProofIds {
		p.vProofIds[k] = -1
	}
}

// ga2ProofCore is (Vec_Int_t *)Sat_ProofCore( p->pSat ).
func ga2ProofCore(p *Ga2_Man_t) *vec.Vec_Int_t {
	core := bsat.Sat_ProofCore(p.pSat)
	return vec.Vec_IntAllocArrayCopy(core, len(core))
}
