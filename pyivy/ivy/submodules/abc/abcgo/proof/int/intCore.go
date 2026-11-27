package int

// Line-by-line port of src/proof/int/intCore.c: the top-level loop of the
// interpolation-based model checker.
import (
	"fmt"
	"time"

	"github.com/glycerine/abcgo/aig/aig"
	"github.com/glycerine/abcgo/aig/saig"
	"github.com/glycerine/abcgo/misc/util"
	"github.com/glycerine/abcgo/misc/vec"
	"github.com/glycerine/abcgo/opt/dar"
	"github.com/glycerine/abcgo/sat/bmc"
	"github.com/glycerine/abcgo/sat/cnf"
)

// interRwsat is Dar_ManRwsat( pAig, 1, 0 ); NULL (a timeout) is nil.
func interRwsat(pAig *Aig_Man_t) *Aig_Man_t {
	pNew, err := dar.Dar_ManRwsat(pAig, true, false)
	if err != nil {
		return nil
	}
	return pNew
}

/*
*Function*************************************************************

	Synopsis    [This procedure sets default values of interpolation parameters.]

**********************************************************************
*/
func Inter_ManSetDefaultParams(p *Inter_ManParams_t) {
	*p = Inter_ManParams_t{}
	p.NBTLimit = 0      // limit on the number of conflicts
	p.NFramesMax = 0    // the max number timeframes to unroll
	p.NSecLimit = 0     // time limit in seconds
	p.NFramesK = 1      // the number of timeframes to use in induction
	p.FRewrite = 0      // use additional rewriting to simplify timeframes
	p.FTransLoop = 0    // add transition into the init state under new PI var
	p.FUsePudlak = 0    // use Pudluk interpolation procedure
	p.FUseOther = 0     // use other undisclosed option
	p.FUseMiniSat = 0   // use MiniSat-1.14p instead of internal proof engine
	p.FCheckKstep = 1   // check using K-step induction
	p.FUseBias = 0      // bias decisions to global variables
	p.FUseBackward = 0  // perform backward interpolation
	p.FUseSeparate = 0  // solve each output separately
	p.FUseTwoFrames = 0 // create OR of two last timeframes
	p.FDropSatOuts = 0  // replace by 1 the solved outputs
	p.FVerbose = 0      // print verbose statistics
	p.IFrameMax = -1
}

/*
*Function*************************************************************

	Synopsis    [Interplates while the number of conflicts is not exceeded.]

	Description [Returns 1 if proven. 0 if failed. -1 if undecided.]

	SideEffects [Does not check the property in 0-th frame.]

**********************************************************************
*/
func Inter_ManPerformInterpolation(pAig *Aig_Man_t, pPars *Inter_ManParams_t, piFrame *int32) int32 {
	var p *Inter_Man_t
	var pCheck *Inter_Check_t
	var pAigTemp *Aig_Man_t
	var s, i, RetValue, Status int32
	var clk, clk2, timeTemp abctime
	clkTotal := util.Abc_Clock()
	var nTimeNewOut abctime
	if pPars.NSecLimit != 0 {
		nTimeNewOut = int64(pPars.NSecLimit)*util.CLOCKS_PER_SEC + util.Abc_Clock()
	}

	// enable ORing of the interpolants, if containment check is performed inductively with K > 1
	if pPars.NFramesK > 1 {
		pPars.FTransLoop = 1
	}
	// backward interpolation works only with the transition loop: without it,
	// the initial interpolant (Inter_ManStartOneOutput) has one more CI than
	// the transition relation (Inter_ManStartDuplicated) and the connector
	// clauses are wrong (safe properties were reported as failing);
	// CBUGS_FIXED.md #42
	if pPars.FUseBackward != 0 {
		pPars.FTransLoop = 1
	}

	// sanity checks
	assert(saig.Saig_ManRegNum(pAig) > 0, "Saig_ManRegNum(pAig) > 0")
	assert(saig.Saig_ManPiNum(pAig) > 0, "Saig_ManPiNum(pAig) > 0")
	assert(saig.Saig_ManPoNum(pAig)-saig.Saig_ManConstrNum(pAig) == 1, "Saig_ManPoNum(pAig)-Saig_ManConstrNum(pAig) == 1")
	if pPars.FVerbose != 0 && saig.Saig_ManConstrNum(pAig) != 0 {
		fmt.Printf("Performing interpolation with %d constraints...\n", saig.Saig_ManConstrNum(pAig))
	}

	if Inter_ManCheckInitialState(pAig) != 0 {
		*piFrame = -1
		fmt.Printf("Property trivially fails in the initial state.\n")
		return 0
	}
	/*
	   if ( Inter_ManCheckAllStates(pAig) )
	   {
	       printf( "Property trivially holds in all states.\n" );
	       return 1;
	   }
	*/
	// create interpolation manager
	// can perform SAT sweeping and/or rewriting of this AIG...
	p = Inter_ManCreate(pAig, pPars)
	if pPars.FTransLoop != 0 {
		p.pAigTrans = Inter_ManStartOneOutput(pAig, 0)
	} else {
		p.pAigTrans = Inter_ManStartDuplicated(pAig)
	}
	// derive CNF for the transformed AIG
	clk = util.Abc_Clock()
	p.pCnfAig = cnf.Cnf_Derive(p.pAigTrans, aig.Aig_ManRegNum(p.pAigTrans))
	p.timeCnf += util.Abc_Clock() - clk
	if pPars.FVerbose != 0 {
		fmt.Printf("AIG: PI/PO/Reg = %d/%d/%d. And = %d. Lev = %d.  CNF: Var/Cla = %d/%d.\n",
			saig.Saig_ManPiNum(pAig), saig.Saig_ManPoNum(pAig), saig.Saig_ManRegNum(pAig),
			aig.Aig_ManAndNum(pAig), aig.Aig_ManLevelNum(pAig),
			p.pCnfAig.Variables, len(p.pCnfAig.Clauses))
	}

	// derive interpolant
	*piFrame = -1
	p.nFrames = 1
	for s = 0; ; s++ {
		var pCnfInter2 *Cnf_Dat_t

		clk2 = util.Abc_Clock()
		// initial state
		if pPars.FUseBackward != 0 {
			p.pInter = Inter_ManStartOneOutput(pAig, 1)
		} else {
			p.pInter = Inter_ManStartInitState(aig.Aig_ManRegNum(pAig))
		}
		assert(aig.Aig_ManCoNum(p.pInter) == 1, "Aig_ManCoNum(p->pInter) == 1")
		clk = util.Abc_Clock()
		p.pCnfInter = cnf.Cnf_Derive(p.pInter, 0)
		p.timeCnf += util.Abc_Clock() - clk
		// timeframes
		p.pFrames = Inter_ManFramesInter(pAig, p.nFrames, pPars.FUseBackward, pPars.FUseTwoFrames)
		clk = util.Abc_Clock()
		if pPars.FRewrite != 0 {
			pAigTemp = p.pFrames
			p.pFrames = interRwsat(pAigTemp)
			aig.Aig_ManStop(pAigTemp)
			//        p->pFrames = Fra_FraigEquivence( pAigTemp = p->pFrames, 100, 0 );
			//        Aig_ManStop( pAigTemp );
		}
		p.timeRwr += util.Abc_Clock() - clk
		// can also do SAT sweeping on the timeframes...
		clk = util.Abc_Clock()
		if pPars.FUseBackward != 0 {
			p.pCnfFrames = cnf.Cnf_Derive(p.pFrames, aig.Aig_ManCoNum(p.pFrames))
		} else {
			//            p->pCnfFrames = Cnf_Derive( p->pFrames, 0 );
			p.pCnfFrames = cnf.Cnf_DeriveSimple(p.pFrames, 0)
		}
		p.timeCnf += util.Abc_Clock() - clk
		// report statistics
		if pPars.FVerbose != 0 {
			fmt.Printf("Step = %2d. Frames = 1 + %d. And = %5d. Lev = %5d.  ",
				s+1, p.nFrames, aig.Aig_ManNodeNum(p.pFrames), aig.Aig_ManLevelNum(p.pFrames))
			util.Abc_PrintTime(1, "Time", util.Abc_Clock()-clk2)
		}

		//////////////////////////////////////////
		// start containment checking
		if !(pPars.FTransLoop != 0 || pPars.FUseBackward != 0 || pPars.NFramesK > 1) {
			pCheck = Inter_CheckStart(p.pAigTrans, pPars.NFramesK)
			// try new containment check for the initial state
			clk = util.Abc_Clock()
			pCnfInter2 = cnf.Cnf_Derive(p.pInter, 1)
			p.timeCnf += util.Abc_Clock() - clk
			clk = util.Abc_Clock()
			RetValue = Inter_CheckPerform(pCheck, pCnfInter2, nTimeNewOut)
			p.timeEqu += util.Abc_Clock() - clk
			//            assert( RetValue == 0 );
			cnf.Cnf_DataFree(pCnfInter2)
			if p.vInters != nil {
				vec.Vec_PtrPush(p.vInters, aig.Aig_ManDupSimple(p.pInter))
			}
		}
		//////////////////////////////////////////

		// iterate the interpolation procedure
		for i = 0; ; i++ {
			if pPars.NFramesMax != 0 && p.nFrames+i >= pPars.NFramesMax {
				if pPars.FVerbose != 0 {
					fmt.Printf("Reached limit (%d) on the number of timeframes.\n", pPars.NFramesMax)
				}
				p.timeTotal = util.Abc_Clock() - clkTotal
				Inter_ManStop(p, 0)
				Inter_CheckStop(pCheck)
				return -1
			}

			// perform interpolation
			clk = util.Abc_Clock()
			// #ifdef ABC_USE_LIBRARIES
			//             if ( pPars->fUseMiniSat )
			//             {
			//                 assert( !pPars->fUseBackward );
			//                 RetValue = Inter_ManPerformOneStepM114p( p, pPars->fUsePudlak, pPars->fUseOther );
			//             }
			//             else
			// #endif
			// ABC is built without ABC_USE_LIBRARIES, so -m (fUseMiniSat) has no
			// effect; the MiniSat-1.14p port stays available in intM114p.go.
			RetValue = Inter_ManPerformOneStep(p, pPars.FUseBias, pPars.FUseBackward, nTimeNewOut)

			if pPars.FVerbose != 0 {
				fmt.Printf("   I = %2d. Bmc =%3d. IntAnd =%6d. IntLev =%5d. Conf =%6d.  ",
					i+1, i+1+p.nFrames, aig.Aig_ManNodeNum(p.pInter), aig.Aig_ManLevelNum(p.pInter), p.nConfCur)
				util.Abc_PrintTime(1, "Time", util.Abc_Clock()-clk)
			}
			// remember the number of timeframes completed
			pPars.IFrameMax = i - 1 + p.nFrames
			if RetValue == 0 { // found a (spurious?) counter-example
				if i == 0 { // real counterexample
					if pPars.FVerbose != 0 {
						fmt.Printf("Found a real counterexample in frame %d.\n", p.nFrames)
					}
					p.timeTotal = util.Abc_Clock() - clkTotal
					*piFrame = p.nFrames
					//                    pAig->pSeqModel = (Abc_Cex_t *)Inter_ManGetCounterExample( pAig, p->nFrames+1, pPars->fVerbose );
					{
						var RetValue int32
						var ParsBmc bmc.Saig_ParBmc_t
						pParsBmc := &ParsBmc
						bmc.Saig_ParBmcSetDefaultParams(pParsBmc)
						pParsBmc.NConfLimit = 100000000
						pParsBmc.NStart = p.nFrames
						pParsBmc.FVerbose = pPars.FVerbose
						RetValue = bmc.Saig_ManBmcScalable(pAig, pParsBmc)
						if RetValue == 1 {
							fmt.Printf("Error: The problem should be SAT but it is UNSAT.\n")
						} else if RetValue == -1 {
							fmt.Printf("Error: The problem timed out.\n")
						}
					}
					Inter_ManStop(p, 0)
					Inter_CheckStop(pCheck)
					return 0
				}
				// likely spurious counter-example
				p.nFrames += i
				Inter_ManClean(p)
				break
			} else if RetValue == -1 {
				if pPars.NSecLimit != 0 && util.Abc_Clock() > nTimeNewOut { // timed out
					if pPars.FVerbose != 0 {
						fmt.Printf("Reached timeout (%d seconds).\n", pPars.NSecLimit)
					}
				} else {
					assert(p.nConfCur >= p.nConfLimit, "p->nConfCur >= p->nConfLimit")
					if pPars.FVerbose != 0 {
						fmt.Printf("Reached limit (%d) on the number of conflicts.\n", p.nConfLimit)
					}
				}
				p.timeTotal = util.Abc_Clock() - clkTotal
				Inter_ManStop(p, 0)
				Inter_CheckStop(pCheck)
				return -1
			}
			assert(RetValue == 1, "RetValue == 1") // found new interpolant
			// compress the interpolant
			clk = util.Abc_Clock()
			if p.pInterNew != nil {
				// save the timeout value
				if nTimeNewOut != 0 {
					p.pInterNew.Time2Quit = time.Now().Add(time.Duration(nTimeNewOut - util.Abc_Clock()))
				}
				//                Ioa_WriteAiger( p->pInterNew, "interpol.aig", 0, 0 );
				pAigTemp = p.pInterNew
				p.pInterNew = interRwsat(pAigTemp)
				//                p->pInterNew = Dar_ManRwsat( pAigTemp = p->pInterNew, 0, 0 );
				aig.Aig_ManStop(pAigTemp)
				if p.pInterNew == nil {
					fmt.Printf("Reached timeout (%d seconds) during rewriting.\n", pPars.NSecLimit)
					p.timeTotal = util.Abc_Clock() - clkTotal
					Inter_ManStop(p, 1)
					Inter_CheckStop(pCheck)
					return -1
				}
			}
			p.timeRwr += util.Abc_Clock() - clk

			// check if interpolant is trivial
			if p.pInterNew == nil || aig.Aig_ObjChild0(p.pInterNew, aig.Aig_ManCo(p.pInterNew, 0)) == aig.Aig_ManConst0(p.pInterNew) {
				//                printf( "interpolant is constant 0\n" );
				if pPars.FVerbose != 0 {
					fmt.Printf("The problem is trivially true for all states.\n")
				}
				p.timeTotal = util.Abc_Clock() - clkTotal
				Inter_ManStop(p, 1)
				Inter_CheckStop(pCheck)
				return 1
			}

			// check containment of interpolants
			clk = util.Abc_Clock()
			if pPars.FCheckKstep != 0 { // k-step unique-state induction
				if aig.Aig_ManCiNum(p.pInterNew) == aig.Aig_ManCiNum(p.pInter) {
					if pPars.FTransLoop != 0 || pPars.FUseBackward != 0 || pPars.NFramesK > 1 {
						clk2 = util.Abc_Clock()
						Status = Inter_ManCheckInductiveContainment(p.pAigTrans, p.pInterNew, util.Abc_MinInt(i+1, pPars.NFramesK), pPars.FUseBackward)
						timeTemp = util.Abc_Clock() - clk2
					} else { // new containment check
						clk2 = util.Abc_Clock()
						pCnfInter2 = cnf.Cnf_Derive(p.pInterNew, 1)
						p.timeCnf += util.Abc_Clock() - clk2
						timeTemp = util.Abc_Clock() - clk2

						Status = Inter_CheckPerform(pCheck, pCnfInter2, nTimeNewOut)
						cnf.Cnf_DataFree(pCnfInter2)
						if p.vInters != nil {
							vec.Vec_PtrPush(p.vInters, aig.Aig_ManDupSimple(p.pInterNew))
						}
					}
				} else {
					Status = 0
				}
			} else { // combinational containment
				if aig.Aig_ManCiNum(p.pInterNew) == aig.Aig_ManCiNum(p.pInter) {
					Status = Inter_ManCheckContainment(p.pInterNew, p.pInter)
				} else {
					Status = 0
				}
			}
			p.timeEqu += util.Abc_Clock() - clk - timeTemp
			if Status != 0 { // contained
				if pPars.FVerbose != 0 {
					fmt.Printf("Proved containment of interpolants.\n")
				}
				p.timeTotal = util.Abc_Clock() - clkTotal
				Inter_ManStop(p, 1)
				Inter_CheckStop(pCheck)
				return 1
			}
			if pPars.NSecLimit != 0 && util.Abc_Clock() > nTimeNewOut {
				fmt.Printf("Reached timeout (%d seconds).\n", pPars.NSecLimit)
				p.timeTotal = util.Abc_Clock() - clkTotal
				Inter_ManStop(p, 1)
				Inter_CheckStop(pCheck)
				return -1
			}
			// save interpolant and convert it into CNF
			if pPars.FTransLoop != 0 {
				aig.Aig_ManStop(p.pInter)
				p.pInter = p.pInterNew
			} else {
				if pPars.FUseBackward != 0 {
					pAigTemp = p.pInter
					p.pInter = aig.Aig_ManCreateMiter(pAigTemp, p.pInterNew, 2)
					aig.Aig_ManStop(pAigTemp)
					aig.Aig_ManStop(p.pInterNew)
					// compress the interpolant
					clk = util.Abc_Clock()
					pAigTemp = p.pInter
					p.pInter = interRwsat(pAigTemp)
					aig.Aig_ManStop(pAigTemp)
					p.timeRwr += util.Abc_Clock() - clk
				} else { // forward with the new containment checking (using only the frontier)
					aig.Aig_ManStop(p.pInter)
					p.pInter = p.pInterNew
				}
			}
			p.pInterNew = nil
			cnf.Cnf_DataFree(p.pCnfInter)
			clk = util.Abc_Clock()
			p.pCnfInter = cnf.Cnf_Derive(p.pInter, 0)
			p.timeCnf += util.Abc_Clock() - clk
		}

		// start containment checking
		Inter_CheckStop(pCheck)
	}
}
