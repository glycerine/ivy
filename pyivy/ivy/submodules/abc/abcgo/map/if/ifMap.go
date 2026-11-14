package if_

import (
	"os"

	"github.com/glycerine/abcgo/misc/extra"
	"github.com/glycerine/abcgo/misc/tim"
	abcutil "github.com/glycerine/abcgo/misc/util"
)

/**CFile****************************************************************

  FileName    [ifMap.c]

  SystemName  [ABC: Logic synthesis and verification system.]

  PackageName [FPGA mapping based on priority cuts.]

  Synopsis    [Mapping procedures.]

  Author      [Alan Mishchenko]

  Affiliation [UC Berkeley]

  Date        [Ver. 1.0. Started - November 21, 2006.]

  Revision    [$Id: ifMap.c,v 1.00 2006/11/21 00:00:00 alanmi Exp $]

***********************************************************************/

//#include "if.h"
//#include "misc/extra/extra.h"

//ABC_NAMESPACE_IMPL_START

////////////////////////////////////////////////////////////////////////
///                        DECLARATIONS                              ///
////////////////////////////////////////////////////////////////////////

// extern char * Dau_DsdMerge( char * pDsd0i, int * pPerm0, char * pDsd1i, int * pPerm1, int fCompl0, int fCompl1, int nVars );

// extern int    If_CutDelayRecCost3( If_Man_t* p, If_Cut_t* pCut, If_Obj_t * pObj );

// extern int    Abc_ExactDelayCost( word * pTruth, int nVars, int * pArrTimeProfile, char * pPerm, int * Cost, int AigLevel );

////////////////////////////////////////////////////////////////////////
///                     FUNCTION DEFINITIONS                         ///
////////////////////////////////////////////////////////////////////////

/*
*Function*************************************************************

	Synopsis    [Compute delay of the cut's output in terms of logic levels.]

	Description [Uses the best arrival time of the fanins of the cut
	to compute the arrival times of the output of the cut.]

	SideEffects []

	SeeAlso     []

**********************************************************************
*/
func If_ManCutAigDelay_rec(p *If_Man_t, pObj *If_Obj_t, vVisited *Vec_Ptr_t) int32 {
	var Delay0, Delay1 int32
	if pObj.fVisit != 0 {
		return pObj.iCopy
	}
	if If_ObjIsCi(pObj) != 0 || If_ObjIsConst1(pObj) != 0 {
		return -1
	}
	// store the node in the structure by level
	assert(If_ObjIsAnd(pObj) != 0, "If_ObjIsAnd(pObj)")
	pObj.fVisit = 1
	Vec_PtrPush(vVisited, pObj)
	Delay0 = If_ManCutAigDelay_rec(p, pObj.pFanin0, vVisited)
	Delay1 = If_ManCutAigDelay_rec(p, pObj.pFanin1, vVisited)
	if Delay0 >= 0 && Delay1 >= 0 {
		pObj.iCopy = 1 + Abc_MaxInt(Delay0, Delay1)
	} else {
		pObj.iCopy = -1
	}
	return pObj.iCopy
}
func If_ManCutAigDelay(p *If_Man_t, pObj *If_Obj_t, pCut *If_Cut_t) int32 {
	var Delay int32
	Vec_PtrClear(p.vVisited)
	// inlining of macro If_CutForEachLeaf
	for i := 0; i < int(pCut.nLeaves); i++ {
		pLeaf := If_ManObj(p, pCut.pLeaves[i])
		assert(pLeaf.fVisit == 0, "pLeaf.fVisit == 0")
		pLeaf.fVisit = 1
		Vec_PtrPush(p.vVisited, pLeaf)
		pLeaf.iCopy = int32(If_ObjCutBest(pLeaf).Delay)
	}
	Delay = If_ManCutAigDelay_rec(p, pObj, p.vVisited)
	// inlining of macro Vec_PtrForEachEntry
	for _, entry := range Vec_PtrArray(p.vVisited) {
		entry.(*If_Obj_t).fVisit = 0
	}
	//    assert( Delay <= (int)pObj->Level );
	return Delay
}

/*
*Function*************************************************************

	Synopsis    [Counts the number of 1s in the signature.]

	Description []

	SideEffects []

	SeeAlso     []

**********************************************************************
*/
var spStatic_If_CutDelaySpecial_Pin2Pin [2][3]float32 = [2][3]float32{{1.0, 1.0, 1.0}, {1.0, 1.0, 0.0}}

func If_CutDelaySpecial(p *If_Man_t, pCut *If_Cut_t, fCarry int32) float32 {
	var DelayCur float32
	var Delay float32 = -IF_FLOAT_LARGE
	assert(pCut.nLeaves <= 3, "pCut.nLeaves <= 3")
	// inlining of macro If_CutForEachLeaf
	for i := 0; i < int(pCut.nLeaves); i++ {
		pLeaf := If_ManObj(p, pCut.pLeaves[i])
		DelayCur = If_ObjCutBest(pLeaf).Delay
		// inlining of macro IF_MAX

		pinDelay := spStatic_If_CutDelaySpecial_Pin2Pin[int(fCarry)][i]
		if Delay < pinDelay+DelayCur {
			Delay = pinDelay + DelayCur
		}
	}
	return Delay
}

/*
*Function*************************************************************

	Synopsis    [Returns arrival time profile of the cut.]

	Description [The procedure returns static storage, which should not be
	deallocated and is only valid until before the procedure is called again.]

	SideEffects []

	SeeAlso     []

**********************************************************************
*/
func If_CutArrTimeProfile(p *If_Man_t, pCut *If_Cut_t) *int32 {
	var i int32
	for i = 0; i < If_CutLeaveNum(pCut); i++ {
		p.pArrTimeProfile[i] = int32(If_ObjCutBest(If_CutLeaf(p, pCut, i)).Delay)
	}
	return &p.pArrTimeProfile[0]
}

func ifMapFlag(value bool) int32 {
	if value {
		return 1
	}
	return 0
}

func ifMapTruth32(p *If_Man_t, cut *If_Cut_t) []uint32 {
	If_CutTruth(p, cut)
	return p.pTruth32[:2*int(p.nTruth6Words[cut.nLeaves])]
}

func ifMapTruth64(p *If_Man_t, cut *If_Cut_t) []uint64 {
	If_CutTruthW(p, cut)
	return p.puTempW[:p.nTruth6Words[cut.nLeaves]]
}

const ifMapExactInfinity = int32(1_000_000_000)

/*
*Function*************************************************************

	Synopsis    [Finds the best cut for the given node.]

	Description [Mapping modes: delay (0), area flow (1), area (2).]

	SideEffects []

	SeeAlso     []

**********************************************************************
*/
func If_ObjPerformMappingAnd(p *If_Man_t, pObj *If_Obj_t, Mode int32, fPreprocess int32, fFirst int32) {
	var pCutSet *If_Set_t
	var pCut0, pCut1, pCut *If_Cut_t
	var pCut0R, pCut1R *If_Cut_t
	var fFunc0R, fFunc1R int32
	var i, k, v, iCutDsd, fChange int32
	fSave0 := ifMapFlag(p.pPars.FDelayOpt != 0 || p.pPars.FDelayOptLut != 0 || p.pPars.FDsdBalance != 0 || p.pPars.FUserRecLib != 0 || p.pPars.FUserSesLib != 0 ||
		p.pPars.FUseDsdTune != 0 || p.pPars.FUseCofVars != 0 || p.pPars.FUseAndVars != 0 || p.pPars.FUse34Spec != 0 || p.pPars.PLutStruct != "" || p.pPars.pFuncCell2 != nil)
	fUseAndCut := p.pPars.NAndDelay > 0 || p.pPars.NAndArea > 0
	assert(If_ObjIsAnd(pObj.pFanin0) == 0 || pObj.pFanin0.pCutSet.nCuts > 0, "fanin 0 has cuts when it is an AND")
	assert(If_ObjIsAnd(pObj.pFanin1) == 0 || pObj.pFanin1.pCutSet.nCuts > 0, "fanin 1 has cuts when it is an AND")

	// prepare
	if Mode == 0 {
		pObj.EstRefs = float32(pObj.nRefs)
	} else if Mode == 1 {
		pObj.EstRefs = (2*pObj.EstRefs + float32(pObj.nRefs)) / 3
	}
	// deref the selected cut
	if Mode != 0 && pObj.nRefs > 0 {
		If_CutAreaDeref(p, If_ObjCutBest(pObj))
	}

	// prepare the cutset
	pCutSet = If_ManSetupNodeCutSet(p, pObj)

	// get the current assigned best cut
	pCut = If_ObjCutBest(pObj)
	if !(fFirst != 0) {
		// recompute the parameters of the best cut
		if p.pPars.FDelayOpt != 0 {
			pCut.Delay = float32(If_CutSopBalanceEval(p, pCut, nil))
		} else if p.pPars.FDsdBalance != 0 {
			pCut.Delay = float32(If_CutDsdBalanceEval(p, pCut, nil))
		} else if p.pPars.FUserRecLib != 0 {
			panic("if: recursive library delay mode is not ported")
		} else if p.pPars.FUserSesLib != 0 {
			panic("if: exact SAT delay mode is not ported")

		} else if p.pPars.FDelayOptLut != 0 {
			pCut.Delay = float32(If_CutLutBalanceEval(p, pCut))
		} else if p.pPars.NGateSize > 0 {
			pCut.Delay = float32(If_CutDelaySop(p, pCut))
		} else {
			pCut.Delay = If_CutDelay(p, pObj, pCut)
		}
		assert(pCut.Delay != -1, "pCut.Delay != -1")
		//        assert( pCut->Delay <= pObj->Required + p->fEpsilon );
		if pCut.Delay > pObj.Required+2*p.fEpsilon {
			Abc_Print(1, "If_ObjPerformMappingAnd(): Warning! Node with ID %d has delay (%f) exceeding the required times (%f).\n",
				pObj.Id, pCut.Delay, pObj.Required+p.fEpsilon)
		}
		if Mode == 2 {
			pCut.Area = If_CutAreaDerefed(p, pCut)
		} else {
			pCut.Area = If_CutAreaFlow(p, pCut)
		}
		if p.pPars.FEdge != 0 {
			if Mode == 2 {
				pCut.Edge = If_CutEdgeDerefed(p, pCut)
			} else {
				pCut.Edge = If_CutEdgeFlow(p, pCut)
			}
		}
		if p.pPars.FPower != 0 {
			if Mode == 2 {
				pCut.Power = If_CutPowerDerefed(p, pCut, pObj)
			} else {
				pCut.Power = If_CutPowerFlow(p, pCut, pObj)
			}
		}
		// save the best cut from the previous iteration
		if !(fPreprocess != 0) {
			If_CutCopy(p, pCutSet.ppCuts[pCutSet.nCuts], pCut)
			pCutSet.nCuts++
		}
	}

	// generate cuts
	// inlining of macro If_ObjForEachCut
	for i = 0; i < int32(pObj.pFanin0.pCutSet.nCuts); i++ {
		pCut0 = pObj.pFanin0.pCutSet.ppCuts[i]
		// inlining of macro If_ObjForEachCut
		for k = 0; k < int32(pObj.pFanin1.pCutSet.nCuts); k++ {
			pCut1 = pObj.pFanin1.pCutSet.ppCuts[k]
			// get the next free cut
			assert(pCutSet.nCuts <= pCutSet.nCutsMax, "pCutSet.nCuts <= pCutSet.nCutsMax")
			pCut = pCutSet.ppCuts[pCutSet.nCuts]
			// make sure K-feasible cut exists
			if If_WordCountOnes(pCut0.uSign|pCut1.uSign) > p.pPars.NLutSize {
				continue
			}

			pCut0R = pCut0
			pCut1R = pCut1
			fFunc0R = pCut0.iCutFunc ^ int32(pCut0.fCompl) ^ int32(pObj.fCompl0)
			fFunc1R = pCut1.iCutFunc ^ int32(pCut1.fCompl) ^ int32(pObj.fCompl1)
			if p.pPars.FUseTtPerm == 0 || pCut0.nLeaves > pCut1.nLeaves || (pCut0.nLeaves == pCut1.nLeaves && fFunc0R > fFunc1R) {

			} else {
				// inlining of macro ABC_SWAP
				{
					var t *If_Cut_t = pCut0R
					pCut0R = pCut1R
					pCut1R = t
				}
				// inlining of macro ABC_SWAP
				{
					var t int32 = fFunc0R
					fFunc0R = fFunc1R
					fFunc1R = t
				}
			}

			// merge the cuts
			if p.pPars.FUseTtPerm != 0 {
				if If_CutMerge(p, pCut0R, pCut1R, pCut) == 0 {
					continue
				}

			} else {
				if If_CutMergeOrdered(p, pCut0, pCut1, pCut) == 0 {
					continue
				}
			}
			if pObj.fSpec != 0 && pCut.nLeaves == uint32(p.pPars.NLutSize) {
				continue
			}
			p.nCutsMerged++
			p.nCutsTotal++
			// check if this cut is contained in any of the available cuts
			if p.pPars.fSkipCutFilter == 0 && If_CutFilter(pCutSet, pCut, fSave0) != 0 {
				continue
			}
			// check if the cut is a special AND-gate cut
			pCut.fAndCut = uint32(ifMapFlag(fUseAndCut && pCut.nLeaves == 2 && pCut.pLeaves[0] == pObj.pFanin0.Id && pCut.pLeaves[1] == pObj.pFanin1.Id))
			//assert( pCut->nLeaves != 2 || pCut->pLeaves[0] < pCut->pLeaves[1] );
			//assert( pCut->nLeaves != 2 || pObj->pFanin0->Id < pObj->pFanin1->Id );
			// compute the truth table
			pCut.iCutFunc = -1
			pCut.fCompl = 0
			if p.pPars.FTruth != 0 {
				//            int nShared = pCut0->nLeaves + pCut1->nLeaves - pCut->nLeaves;
				var clk abctime
				if p.pPars.FVerbose != 0 {
					clk = Abc_Clock()
				}
				if p.pPars.FUseTtPerm != 0 {
					fChange = If_CutComputeTruthPerm(p, pCut, pCut0R, pCut1R, fFunc0R, fFunc1R)
				} else {
					fChange = If_CutComputeTruth(p, pCut, pCut0, pCut1, int32(pObj.fCompl0), int32(pObj.fCompl1))
				}
				if p.pPars.FVerbose != 0 {
					p.timeCache[4] += Abc_Clock() - clk
				}
				if p.pPars.fSkipCutFilter == 0 && fChange != 0 && If_CutFilter(pCutSet, pCut, fSave0) != 0 {
					continue
				}
				if p.pPars.FLut6Filter != 0 && pCut.nLeaves == 6 && If_CutCheckTruth6(p, pCut) == 0 {
					continue
				}
				if p.pPars.FUseDsd != 0 {
					var truthId int32 = Abc_Lit2Var(pCut.iCutFunc)
					if truthId >= int32(Vec_IntSize(p.vTtDsds[pCut.nLeaves])) || Vec_IntEntry(p.vTtDsds[pCut.nLeaves], int(truthId)) == -1 {
						for truthId >= int32(Vec_IntSize(p.vTtDsds[pCut.nLeaves])) {
							Vec_IntPush(p.vTtDsds[pCut.nLeaves], -1)
							for v = 0; v < Abc_MaxInt(6, int32(pCut.nLeaves)); v++ {
								Vec_StrPush(p.vTtPerms[pCut.nLeaves], IF_BIG_CHAR)
							}
						}
						iCutDsd = If_DsdManCompute(p.pIfDsdMan, ifMapTruth64(p, pCut), int32(pCut.nLeaves), If_CutDsdPerm(p, pCut), p.pPars.PLutStruct)
						Vec_IntWriteEntry(p.vTtDsds[pCut.nLeaves], int(truthId), iCutDsd)
					}
					assert(If_DsdManSuppSize(p.pIfDsdMan, If_CutDsdLit(p, pCut)) == int32(pCut.nLeaves), "If_DsdManSuppSize(p.pIfDsdMan, If_CutDsdLit(p, pCut)) == int32(pCut.nLeaves)")
					//If_ManCacheRecord( p, If_CutDsdLit(p, pCut0), If_CutDsdLit(p, pCut1), nShared, If_CutDsdLit(p, pCut) );
				}
				// run user functions
				pCut.fUseless = 0
				if p.pPars.PFuncCell != nil || p.pPars.pFuncCell2 != nil {
					assert(p.pPars.FUseTtPerm == 0, "p.pPars.fUseTtPerm == 0")
					assert(pCut.nLimit >= 4 && pCut.nLimit <= 16, "pCut.nLimit >= 4 && pCut.nLimit <= 16")
					if p.pPars.FUseDsd != 0 {
						pCut.fUseless = uint32(If_DsdManCheckDec(p.pIfDsdMan, If_CutDsdLit(p, pCut)))
					} else if p.pPars.pFuncCell2 != nil {
						pCut.fUseless = uint32(ifMapFlag(p.pPars.pFuncCell2(p, ifMapTruth64(p, pCut), int32(pCut.nLeaves), nil, nil) == 0))
					} else {
						pCut.fUseless = uint32(ifMapFlag(p.pPars.PFuncCell(p, ifMapTruth32(p, pCut), Abc_MaxInt(6, int32(pCut.nLeaves)), int32(pCut.nLeaves), p.pPars.PLutStruct) == 0))
					}
					p.nCutsUselessAll += ifMapFlag(pCut.fUseless != 0)
					p.nCutsUseless[pCut.nLeaves] += ifMapFlag(pCut.fUseless != 0)
					p.nCutsCountAll++
					p.nCutsCount[pCut.nLeaves]++
					// skip 5-input cuts, which cannot be decomposed
					if (p.pPars.FEnableCheck75 != 0 || p.pPars.FEnableCheck75u != 0) && pCut.nLeaves == 5 && pCut.nLimit == 5 {
						var TruthU uint32 = *If_CutTruth(p, pCut)
						var Truth word = (word(TruthU) << 32) | word(TruthU)
						p.nCuts5++
						if If_CluCheckDecInAny(Truth, 5) != 0 {
							p.nCuts5a++
						} else {
							continue
						}

					} else if p.pPars.FVerbose != 0 && pCut.nLeaves == 5 {
						var TruthU uint32 = *If_CutTruth(p, pCut)
						var Truth word = (word(TruthU) << 32) | word(TruthU)
						p.nCuts5++
						if If_CluCheckDecInAny(Truth, 5) != 0 || If_CluCheckDecOut(Truth, 5) != 0 {
							p.nCuts5a++
						}
					}

				} else if p.pPars.FUseDsdTune != 0 {
					pCut.fUseless = uint32(If_DsdManReadMark(p.pIfDsdMan, If_CutDsdLit(p, pCut)))
					p.nCutsUselessAll += int32(pCut.fUseless)
					p.nCutsUseless[pCut.nLeaves] += int32(pCut.fUseless)
					p.nCutsCountAll++
					p.nCutsCount[pCut.nLeaves]++

				} else if p.pPars.FUse34Spec != 0 {
					assert(pCut.nLeaves <= 4, "pCut.nLeaves <= 4")
					if pCut.nLeaves == 4 && Abc_Tt4Check(int32(0xFFFF&*If_CutTruth(p, pCut))) == 0 {
						pCut.fUseless = 1
					}

				} else {
					if p.pPars.FUseAndVars != 0 {
						var iDecMask int32 = -1
						var truthId int32 = Abc_Lit2Var(pCut.iCutFunc)
						assert(p.pPars.NLutSize <= 13, "p.pPars.nLutSize <= 13")
						if truthId >= int32(Vec_IntSize(p.vTtDecs[pCut.nLeaves])) || Vec_IntEntry(p.vTtDecs[pCut.nLeaves], int(truthId)) == -1 {
							for truthId >= int32(Vec_IntSize(p.vTtDecs[pCut.nLeaves])) {
								Vec_IntPush(p.vTtDecs[pCut.nLeaves], -1)
							}
							if int32(pCut.nLeaves) > p.pPars.NLutSize/2 && int32(pCut.nLeaves) <= 2*(p.pPars.NLutSize/2) {
								iDecMask = abcutil.Abc_TtProcessBiDec(ifMapTruth64(p, pCut), int32(pCut.nLeaves), p.pPars.NLutSize/2)
							} else {
								iDecMask = 0
							}
							Vec_IntWriteEntry(p.vTtDecs[pCut.nLeaves], int(truthId), iDecMask)
						}
						iDecMask = Vec_IntEntry(p.vTtDecs[pCut.nLeaves], int(truthId))
						assert(iDecMask >= 0, "iDecMask >= 0")
						pCut.fUseless = uint32(ifMapFlag(iDecMask == 0 && int32(pCut.nLeaves) > p.pPars.NLutSize/2))
						p.nCutsUselessAll += int32(pCut.fUseless)
						p.nCutsUseless[pCut.nLeaves] += int32(pCut.fUseless)
						p.nCutsCountAll++
						p.nCutsCount[pCut.nLeaves]++
					}
					if p.pPars.FUseCofVars != 0 && (p.pPars.FUseAndVars == 0 || pCut.fUseless != 0) {
						var iCofVar int32 = -1
						var truthId int32 = Abc_Lit2Var(pCut.iCutFunc)
						if truthId >= int32(Vec_StrSize(p.vTtVars[pCut.nLeaves])) || Vec_StrEntry(p.vTtVars[pCut.nLeaves], int(truthId)) == 0xff {
							for truthId >= int32(Vec_StrSize(p.vTtVars[pCut.nLeaves])) {
								Vec_StrPush(p.vTtVars[pCut.nLeaves], 0xff)
							}
							iCofVar = abcutil.Abc_TtCheckCondDep(ifMapTruth64(p, pCut), int32(pCut.nLeaves), p.pPars.NLutSize/2)
							Vec_StrWriteEntry(p.vTtVars[pCut.nLeaves], int(truthId), byte(iCofVar))
						}
						iCofVar = int32(Vec_StrEntry(p.vTtVars[pCut.nLeaves], int(truthId)))
						assert(iCofVar >= 0 && iCofVar <= int32(pCut.nLeaves), "iCofVar >= 0 && iCofVar <= int32(pCut.nLeaves)")
						pCut.fUseless = uint32(ifMapFlag(iCofVar == int32(pCut.nLeaves) && pCut.nLeaves > 0))
						p.nCutsUselessAll += int32(pCut.fUseless)
						p.nCutsUseless[pCut.nLeaves] += int32(pCut.fUseless)
						p.nCutsCountAll++
						p.nCutsCount[pCut.nLeaves]++
					}
				}
			}

			// compute the application-specific cost and depth
			pCut.fUser = uint32(ifMapFlag(p.pPars.pFuncCost != nil))
			if p.pPars.pFuncCost != nil {
				pCut.Cost = uint32(p.pPars.pFuncCost(p, pCut))
			} else {
				pCut.Cost = 0
			}
			if pCut.Cost == IF_COST_MAX {
				continue
			}
			// check if the cut satisfies the required times
			if p.pPars.FDelayOpt != 0 {
				pCut.Delay = float32(If_CutSopBalanceEval(p, pCut, nil))
			} else if p.pPars.FDsdBalance != 0 {
				pCut.Delay = If_CutDsdBalanceEval(p, pCut, nil)
			} else if p.pPars.FUserRecLib != 0 {
				panic("if: recursive library delay cost is outside the ported IF subsystem")
			} else if p.pPars.FUserSesLib != 0 {
				// The C implementation delegates this optional mode to its external
				// exact-delay/SAT subsystem, which is not included in this Go package.
				panic("if: exact SAT delay cost requires the unported ABC exact-delay subsystem")
			} else if p.pPars.FDelayOptLut != 0 {
				pCut.Delay = float32(If_CutLutBalanceEval(p, pCut))
			} else if p.pPars.NGateSize > 0 {
				pCut.Delay = float32(If_CutDelaySop(p, pCut))
			} else {
				pCut.Delay = If_CutDelay(p, pObj, pCut)
			}
			if pCut.Delay == -1 {
				continue
			}
			if (Mode != 0) && pCut.Delay > pObj.Required+p.fEpsilon {
				continue
			}
			// compute area of the cut (this area may depend on the application specific cost)
			if Mode == 2 {
				pCut.Area = If_CutAreaDerefed(p, pCut)
			} else {
				pCut.Area = If_CutAreaFlow(p, pCut)
			}
			if p.pPars.FEdge != 0 {
				if Mode == 2 {
					pCut.Edge = If_CutEdgeDerefed(p, pCut)
				} else {
					pCut.Edge = If_CutEdgeFlow(p, pCut)
				}
			}
			if p.pPars.FPower != 0 {
				if Mode == 2 {
					pCut.Power = If_CutPowerDerefed(p, pCut, pObj)
				} else {
					pCut.Power = If_CutPowerFlow(p, pCut, pObj)
				}
			}
			//        pCut->AveRefs = (Mode == 0)? (float)0.0 : If_CutAverageRefs( p, pCut );
			// insert the cut into storage
			If_CutSort(p, pCutSet, pCut)
			//        If_CutTraverse( p, pObj, pCut );
		}
	}
	assert(pCutSet.nCuts > 0, "pCutSet.nCuts > 0")
	//    If_CutVerifyCuts( pCutSet, !p->pPars->fUseTtPerm );

	// update the best cut
	if !(fPreprocess != 0) || pCutSet.ppCuts[0].Delay <= pObj.Required+p.fEpsilon {
		If_CutCopy(p, If_ObjCutBest(pObj), pCutSet.ppCuts[0])
		if p.pPars.FUserRecLib != 0 || p.pPars.FUserSesLib != 0 {
			assert(If_ObjCutBest(pObj).Cost < IF_COST_MAX && If_ObjCutBest(pObj).Delay < IF_INFINITY, "If_ObjCutBest(pObj).Cost < IF_COST_MAX && If_ObjCutBest(pObj).Delay < IF_INFINITY")
		}
	}
	// add the trivial cut to the set
	if pObj.fSkipCut == 0 && If_ObjCutBest(pObj).nLeaves > 1 {
		If_ManSetupCutTriv(p, pCutSet.ppCuts[pCutSet.nCuts], pObj.Id)
		pCutSet.nCuts++
		assert(pCutSet.nCuts <= pCutSet.nCutsMax+1, "pCutSet.nCuts <= pCutSet.nCutsMax+1")
	}
	//    if ( If_ObjCutBest(pObj)->nLeaves == 0 )
	//        p->nBestCutSmall[0]++;
	//    else if ( If_ObjCutBest(pObj)->nLeaves == 1 )
	//        p->nBestCutSmall[1]++;

	// ref the selected cut
	if (Mode != 0) && pObj.nRefs > 0 {
		If_CutAreaRef(p, If_ObjCutBest(pObj))
	}
	if If_ObjCutBest(pObj).fUseless != 0 {
		Abc_Print(1, "The best cut is useless.\n")
	}
	// call the user specified function for each cut
	if p.pPars.pFuncUser != nil {
		// inlining of macro If_ObjForEachCut
		for i := 0; i < int(pObj.pCutSet.nCuts); i++ {
			pCut = pObj.pCutSet.ppCuts[i]
			p.pPars.pFuncUser(p, pObj, pCut)
		}
	}
	// free the cuts
	If_ManDerefNodeCutSet(p, pObj)
}

/*
*Function*************************************************************

	Synopsis    [Finds the best cut for the choice node.]

	Description []

	SideEffects []

	SeeAlso     []

**********************************************************************
*/
func If_ObjPerformMappingChoice(p *If_Man_t, pObj *If_Obj_t, Mode int32, fPreprocess int32) {
	var pCutSet *If_Set_t
	var pTemp *If_Obj_t
	var pCutTemp, pCut *If_Cut_t
	var i int32
	fSave0 := ifMapFlag(p.pPars.FDelayOpt != 0 || p.pPars.FDelayOptLut != 0 || p.pPars.FDsdBalance != 0 || p.pPars.FUserRecLib != 0 || p.pPars.FUserSesLib != 0 || p.pPars.FUse34Spec != 0)
	assert(pObj.pEquiv != nil, "pObj.pEquiv != nil")

	// prepare
	if (Mode != 0) && pObj.nRefs > 0 {
		If_CutAreaDeref(p, If_ObjCutBest(pObj))
	}

	// remove elementary cuts
	for pTemp = pObj; pTemp != nil; pTemp = pTemp.pEquiv {
		if pTemp != pObj || pTemp.pCutSet.nCuts > 1 {
			pTemp.pCutSet.nCuts--
		}
	}

	// update the cutset of the node
	pCutSet = pObj.pCutSet

	// generate cuts
	for pTemp = pObj.pEquiv; pTemp != nil; pTemp = pTemp.pEquiv {
		if pTemp.pCutSet.nCuts == 0 {
			continue
		}
		// go through the cuts of this node
		// inlining of macro If_ObjForEachCut
		for i = 0; i < int32(pTemp.pCutSet.nCuts); i++ {
			pCutTemp = pTemp.pCutSet.ppCuts[i]
			if pCutTemp.fUseless != 0 {
				continue
			}
			// get the next free cut
			assert(pCutSet.nCuts <= pCutSet.nCutsMax, "pCutSet.nCuts <= pCutSet.nCutsMax")
			pCut = pCutSet.ppCuts[pCutSet.nCuts]
			// copy the cut into storage
			If_CutCopy(p, pCut, pCutTemp)
			// check if this cut is contained in any of the available cuts
			if If_CutFilter(pCutSet, pCut, fSave0) != 0 {
				continue
			}
			// check if the cut satisfies the required times
			//            assert( pCut->Delay == If_CutDelay( p, pTemp, pCut ) );
			if Mode != 0 && pCut.Delay > pObj.Required+p.fEpsilon {
				continue
			}
			// set the phase attribute
			pCut.fCompl = pObj.fPhase ^ pTemp.fPhase
			// compute area of the cut (this area may depend on the application specific cost)
			if Mode == 2 {
				pCut.Area = If_CutAreaDerefed(p, pCut)
			} else {
				pCut.Area = If_CutAreaFlow(p, pCut)
			}
			if p.pPars.FEdge != 0 {
				if Mode == 2 {
					pCut.Edge = If_CutEdgeDerefed(p, pCut)
				} else {
					pCut.Edge = If_CutEdgeFlow(p, pCut)
				}
			}
			if p.pPars.FPower != 0 {
				if Mode == 2 {
					pCut.Power = If_CutPowerDerefed(p, pCut, pObj)
				} else {
					pCut.Power = If_CutPowerFlow(p, pCut, pObj)
				}
			}
			//            pCut->AveRefs = (Mode == 0)? (float)0.0 : If_CutAverageRefs( p, pCut );
			// insert the cut into storage
			If_CutSort(p, pCutSet, pCut)
		}
	}
	assert(pCutSet.nCuts > 0, "pCutSet.nCuts > 0")

	// update the best cut
	if !(fPreprocess != 0) || pCutSet.ppCuts[0].Delay <= pObj.Required+p.fEpsilon {
		If_CutCopy(p, If_ObjCutBest(pObj), pCutSet.ppCuts[0])
	}
	// add the trivial cut to the set
	if pObj.fSkipCut == 0 && If_ObjCutBest(pObj).nLeaves > 1 {
		If_ManSetupCutTriv(p, pCutSet.ppCuts[pCutSet.nCuts], pObj.Id)
		pCutSet.nCuts++
		assert(pCutSet.nCuts <= pCutSet.nCutsMax+1, "pCutSet.nCuts <= pCutSet.nCutsMax+1")
	}

	// ref the selected cut
	if Mode != 0 && pObj.nRefs > 0 {
		If_CutAreaRef(p, If_ObjCutBest(pObj))
	}
	// free the cuts
	If_ManDerefChoiceCutSet(p, pObj)
}

/*
*Function*************************************************************

	Synopsis    [Performs one mapping pass over all nodes.]

	Description []

	SideEffects []

	SeeAlso     []

**********************************************************************
*/
func If_ManPerformMappingRound(p *If_Man_t, nCutsUsed, mode, preprocess, first int32, label string) int32 {
	assert(mode >= 0 && mode <= 2, "mapping mode is delay, flow, or area")
	started := Abc_Clock()
	p.nBestCutSmall = [2]int32{}
	p.SortMode = 0
	if mode != 0 || p.pPars.FArea != 0 {
		p.SortMode = 1
	} else if p.pPars.FFancy != 0 {
		p.SortMode = 2
	}
	p.nCutsUsed, p.nCutsMerged = nCutsUsed, 0
	for _, entry := range Vec_PtrArray(p.vObjs) {
		obj := entry.(*If_Obj_t)
		if obj.Type == IF_AND {
			assert(obj.nVisits == obj.nVisitsCopy, "AND visit counts are initialized")
		}
	}

	var progress *extra.ProgressBar
	if p.pManTim != nil {
		tim.Tim_ManIncrementTravId(p.pManTim)
	} else {
		progress = extra.Extra_ProgressBarStart(os.Stdout, int32(Vec_PtrSize(p.vObjs)))
	}
	defer extra.Extra_ProgressBarStop(progress)
	for i, entry := range Vec_PtrArray(p.vObjs) {
		obj := entry.(*If_Obj_t)
		switch obj.Type {
		case IF_AND:
			if progress != nil {
				extra.Extra_ProgressBarUpdate(progress, int32(i), label)
			}
			If_ObjPerformMappingAnd(p, obj, mode, preprocess, first)
			if obj.fRepr != 0 {
				If_ObjPerformMappingChoice(p, obj, mode, preprocess)
			}
		case IF_CI:
			if p.pManTim != nil {
				If_ObjSetArrTime(obj, tim.Tim_ManGetCiArrival(p.pManTim, int(obj.IdPio)))
			}
		case IF_CO:
			if p.pManTim != nil {
				tim.Tim_ManSetCoArrival(p.pManTim, int(obj.IdPio), If_ObjArrTime(obj.pFanin0))
			}
		case IF_CONST1:
			If_ObjSetArrTime(obj, -IF_INFINITY)
		default:
			panic("if: invalid object type in mapping pass")
		}
	}
	for _, entry := range Vec_PtrArray(p.vObjs) {
		obj := entry.(*If_Obj_t)
		if obj.Type == IF_AND {
			assert(obj.nVisits == 0, "AND visit counts return to zero")
		}
	}
	If_ManComputeRequired(p)
	if p.pPars.FVerbose != 0 {
		symbol := byte('A')
		if preprocess != 0 {
			symbol = 'P'
		} else if mode == 0 {
			symbol = 'D'
		} else if mode == 1 {
			symbol = 'F'
		}
		Abc_Print(1, "%c: Del = %7.2f. Ar = %9.1f. Edge = %8d. Switch = %7.2f. Cuts = %8d. ", symbol, p.RequiredGlo, p.AreaGlo, p.nNets, p.dPower, p.nCutsMerged)
		Abc_PrintTime(1, "T", Abc_Clock()-started)
	}
	return 1
}

////////////////////////////////////////////////////////////////////////
///                       END OF FILE                                ///
////////////////////////////////////////////////////////////////////////

//ABC_NAMESPACE_IMPL_END
