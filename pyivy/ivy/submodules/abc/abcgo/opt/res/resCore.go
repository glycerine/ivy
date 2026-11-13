package res

import "github.com/glycerine/abcgo/base/abc"
import "github.com/glycerine/abcgo/sat/bsat"
import "github.com/glycerine/abcgo/misc/util"
import "github.com/glycerine/abcgo/misc/extra"
import "github.com/glycerine/abcgo/aig/hop"
import "github.com/glycerine/abcgo/bool/kit"

import (
	"fmt"
	"os"
)

/**CFile****************************************************************

  FileName    [resCore.c]

  SystemName  [ABC: Logic synthesis and verification system.]

  PackageName [Resynthesis package.]

  Synopsis    [Top-level resynthesis procedure.]

  Author      [Alan Mishchenko]

  Affiliation [UC Berkeley]

  Date        [Ver. 1.0. Started - January 15, 2007.]

  Revision    [$Id: resCore.c,v 1.00 2007/01/15 00:00:00 alanmi Exp $]

***********************************************************************/

//#include "base/abc/abc.h"
//#include "resInt.h"
//#include "bool/kit/kit.h"
//#include "sat/bsat/satStore.h"

//ABC_NAMESPACE_IMPL_START

////////////////////////////////////////////////////////////////////////
///                        DECLARATIONS                              ///
////////////////////////////////////////////////////////////////////////

type Res_Man_t struct {
	// general parameters
	pPars *Res_Par_t
	// specialized manager
	pWin     *Res_Win_t // windowing manager
	pAig     *Abc_Ntk_t // the strashed window
	pSim     *Res_Sim_t // simulation manager
	pCnf     *Sto_Man_t // the CNF of the SAT problem
	pMan     *Int_Man_t // interpolation manager;
	vMem     *Vec_Int_t // memory for intermediate SOPs
	vResubs  *Vec_Vec_t // resubstitution candidates of the AIG
	vResubsW *Vec_Vec_t // resubstitution candidates of the window
	vLevels  *Vec_Vec_t // levelized structure for updating
	// statistics
	nWins        int32 // the number of windows tried
	nWinNodes    int32 // the total number of window nodes
	nDivNodes    int32 // the total number of divisors
	nWinsTriv    int32 // the total number of trivial windows
	nWinsUsed    int32 // the total number of useful windows (with at least one candidate)
	nConstsUsed  int32 // the total number of constant nodes under ODC
	nCandSets    int32 // the total number of candidates
	nProvedSets  int32 // the total number of proved groups
	nSimEmpty    int32 // the empty simulation info
	nTotalNets   int32 // the total number of nets
	nTotalNodes  int32 // the total number of nodess
	nTotalNets2  int32 // the total number of nets
	nTotalNodes2 int32 // the total number of nodess
	// runtime
	timeWin      abctime // windowing
	timeDiv      abctime // divisors
	timeAig      abctime // strashing
	timeSim      abctime // simulation
	timeCand     abctime // resubstitution candidates
	timeSatTotal abctime // SAT solving total
	timeSatSat   abctime // SAT solving (sat calls)
	timeSatUnsat abctime // SAT solving (unsat calls)
	timeSatSim   abctime // SAT solving (simulation)
	timeInt      abctime // interpolation
	timeUpd      abctime // updating
	timeTotal    abctime // total runtime
}

// extern Hop_Obj_t * Kit_GraphToHop( Hop_Man_t * pMan, Kit_Graph_t * pGraph );

var s_ResynTime abctime

////////////////////////////////////////////////////////////////////////
///                     FUNCTION DEFINITIONS                         ///
////////////////////////////////////////////////////////////////////////

/*
*Function*************************************************************

	Synopsis    [Allocate resynthesis manager.]

	Description []

	SideEffects []

	SeeAlso     []

**********************************************************************
*/
func Res_ManAlloc(pPars *Res_Par_t) *Res_Man_t {
	var p *Res_Man_t
	p = new(Res_Man_t)
	*p = Res_Man_t{}
	assert(pPars.nWindow > 0 && pPars.nWindow < 100, "pPars.nWindow > 0 && pPars.nWindow < 100")
	assert(pPars.nCands > 0 && pPars.nCands < 100, "pPars.nCands > 0 && pPars.nCands < 100")
	p.pPars = pPars
	p.pWin = Res_WinAlloc()
	p.pSim = Res_SimAlloc(pPars.nSimWords)
	p.pMan = bsat.Int_ManAlloc()
	p.vMem = Vec_IntAlloc(0)
	p.vResubs = Vec_VecStart(pPars.nCands)
	p.vResubsW = Vec_VecStart(pPars.nCands)
	p.vLevels = Vec_VecStart(32)
	return p
}

/*
*Function*************************************************************

	Synopsis    [Deallocate resynthesis manager.]

	Description []

	SideEffects []

	SeeAlso     []

**********************************************************************
*/
func Res_ManFree(p *Res_Man_t) {
	if p.pPars.fVerbose != 0 {
		fmt.Printf("Reduction in nodes = %5d. (%.2f %%) ", p.nTotalNodes-p.nTotalNodes2, 100*float64(p.nTotalNodes-p.nTotalNodes2)/float64(p.nTotalNodes))
		fmt.Printf("Reduction in edges = %5d. (%.2f %%) ", p.nTotalNets-p.nTotalNets2, 100*float64(p.nTotalNets-p.nTotalNets2)/float64(p.nTotalNets))
		fmt.Printf("\n")

		fmt.Printf("Winds = %d. ", p.nWins)
		fmt.Printf("Nodes = %d. (Ave = %5.1f)  ", p.nWinNodes, float64(p.nWinNodes)/float64(p.nWins))
		fmt.Printf("Divs = %d. (Ave = %5.1f)  ", p.nDivNodes, float64(p.nDivNodes)/float64(p.nWins))
		fmt.Printf("\n")
		fmt.Printf("WinsTriv = %d. ", p.nWinsTriv)
		fmt.Printf("SimsEmpt = %d. ", p.nSimEmpty)
		fmt.Printf("Const = %d. ", p.nConstsUsed)
		fmt.Printf("WindUsed = %d. ", p.nWinsUsed)
		fmt.Printf("Cands = %d. ", p.nCandSets)
		fmt.Printf("Proved = %d.", p.nProvedSets)
		fmt.Printf("\n")

		util.Abc_PrintTimeP(1, "Windowing  ", p.timeWin, p.timeTotal)
		util.Abc_PrintTimeP(1, "Divisors   ", p.timeDiv, p.timeTotal)
		util.Abc_PrintTimeP(1, "Strashing  ", p.timeAig, p.timeTotal)
		util.Abc_PrintTimeP(1, "Simulation ", p.timeSim, p.timeTotal)
		util.Abc_PrintTimeP(1, "Candidates ", p.timeCand, p.timeTotal)
		util.Abc_PrintTimeP(1, "SAT solver ", p.timeSatTotal, p.timeTotal)
		util.Abc_PrintTimeP(1, "    sat    ", p.timeSatSat, p.timeTotal)
		util.Abc_PrintTimeP(1, "    unsat  ", p.timeSatUnsat, p.timeTotal)
		util.Abc_PrintTimeP(1, "    simul  ", p.timeSatSim, p.timeTotal)
		util.Abc_PrintTimeP(1, "Interpol   ", p.timeInt, p.timeTotal)
		util.Abc_PrintTimeP(1, "Undating   ", p.timeUpd, p.timeTotal)
		util.Abc_PrintTimeP(1, "TOTAL      ", p.timeTotal, p.timeTotal)
	}
	Res_WinFree(p.pWin)
	if p.pAig != nil {
		abc.Abc_NtkDelete(p.pAig)
	}
	Res_SimFree(p.pSim)
	if p.pCnf != nil {
		bsat.Sto_ManFree(p.pCnf)
	}
	bsat.Int_ManFree(p.pMan)
	Vec_IntFree(p.vMem)
	Vec_VecFree(p.vResubs)
	Vec_VecFree(p.vResubsW)
	Vec_VecFree(p.vLevels)
}

/*
*Function*************************************************************

	Synopsis    [Incrementally updates level of the nodes.]

	Description []

	SideEffects []

	SeeAlso     []

**********************************************************************
*/
func Res_UpdateNetwork(pObj *Abc_Obj_t, vFanins *Vec_Ptr_t, pFunc *Hop_Obj_t, vLevels *Vec_Vec_t) {
	var pObjNew, pFanin *Abc_Obj_t
	var k int32

	// create the new node
	pObjNew = abc.Abc_NtkCreateNode(pObj.Network())
	pObjNew.Data = pFunc
	for k = 0; k < Vec_PtrSize(vFanins); k++ {
		pFanin = Vec_PtrEntry(vFanins, k)
		abc.Abc_ObjAddFanin(pObjNew, pFanin)
	}
	// replace the old node by the new node
	//printf( "Replacing node " ); Abc_ObjPrint( stdout, pObj );
	//printf( "Inserting node " ); Abc_ObjPrint( stdout, pObjNew );
	// update the level of the node
	Abc_NtkUpdate(pObj, pObjNew, vLevels)
}

/*
*Function*************************************************************

	Synopsis    [Entrace into the resynthesis package.]

	Description []

	SideEffects []

	SeeAlso     []

**********************************************************************
*/
func Abc_NtkResynthesize(pNtk *Abc_Ntk_t, pPars *Res_Par_t) int32 {
	var pProgress *extra.ProgressBar
	var p *Res_Man_t
	var pObj *Abc_Obj_t
	var pFunc *Hop_Obj_t
	var pGraph *Kit_Graph_t
	var vFanins *Vec_Ptr_t
	var puTruth []uint32
	var kitMemory []uint32
	var i, k, RetValue, nNodesOld, nFanins, nFaninsMax int32
	var clk abctime
	var clkTotal abctime = util.Abc_Clock()

	// start the manager
	p = Res_ManAlloc(pPars)
	p.nTotalNets = abc.Abc_NtkGetTotalFanins(pNtk)
	p.nTotalNodes = abc.Abc_NtkNodeNum(pNtk)
	nFaninsMax = abc.Abc_NtkGetFaninMax(pNtk)
	if nFaninsMax > 8 {
		nFaninsMax = 8
	}

	// perform the network sweep
	Abc_NtkSweep(pNtk, 0)

	// convert into the AIG
	if abc.Abc_NtkToAig(pNtk) == 0 {
		fmt.Fprintf(os.Stdout, "Converting to BDD has failed.\n")
		Res_ManFree(p)
		return 0
	}
	assert(abc.Abc_NtkHasAig(pNtk) != 0, "abc.Abc_NtkHasAig(pNtk)")

	// set the number of levels
	abc.Abc_NtkLevel(pNtk)
	Abc_NtkStartReverseLevels(pNtk, pPars.nGrowthLevel)

	// try resynthesizing nodes in the topological order
	nNodesOld = abc.Abc_NtkObjNumMax(pNtk)
	pProgress = extra.Extra_ProgressBarStart(os.Stdout, nNodesOld)
	for i = 0; i < abc.Abc_NtkObjNumMax(pNtk); i++ {
		pObj = abc.Abc_NtkObj(pNtk, i)
		extra.Extra_ProgressBarUpdate(pProgress, i, "")
		if abc.Abc_ObjIsNode(pObj) == 0 {
			continue
		}
		if abc.Abc_ObjFaninNum(pObj) > 8 {
			continue
		}
		if pObj.ID > int(nNodesOld) {
			break
		}

		// create the window for this node
		clk = util.Abc_Clock()
		RetValue = Res_WinCompute(pObj, p.pPars.nWindow/10, p.pPars.nWindow%10, p.pWin)
		p.timeWin += util.Abc_Clock() - clk
		if RetValue == 0 {
			continue
		}
		p.nWinsTriv += Res_WinIsTrivial(p.pWin)

		if p.pPars.fVeryVerbose != 0 {
			fmt.Printf("%5d (lev=%2d) : ", pObj.ID, pObj.Level)
			fmt.Printf("Win = %3d/%3d/%4d/%3d   ", Vec_PtrSize(p.pWin.vLeaves), Vec_PtrSize(p.pWin.vBranches), Vec_PtrSize(p.pWin.vNodes), Vec_PtrSize(p.pWin.vRoots))
		}

		// collect the divisors
		clk = util.Abc_Clock()
		Res_WinDivisors(p.pWin, Abc_ObjRequiredLevel(pObj)-1)
		p.timeDiv += util.Abc_Clock() - clk

		p.nWins++
		p.nWinNodes += Vec_PtrSize(p.pWin.vNodes)
		p.nDivNodes += Vec_PtrSize(p.pWin.vDivs)

		if p.pPars.fVeryVerbose != 0 {
			fmt.Printf("D = %3d ", Vec_PtrSize(p.pWin.vDivs))
			fmt.Printf("D+ = %3d ", p.pWin.nDivsPlus)
		}

		// create the AIG for the window
		clk = util.Abc_Clock()
		if p.pAig != nil {
			abc.Abc_NtkDelete(p.pAig)
		}
		p.pAig = Res_WndStrash(p.pWin)
		p.timeAig += util.Abc_Clock() - clk

		if p.pPars.fVeryVerbose != 0 {
			fmt.Printf("AIG = %4d ", abc.Abc_NtkNodeNum(p.pAig))
			fmt.Printf("\n")
		}

		// prepare simulation info
		clk = util.Abc_Clock()
		RetValue = Res_SimPrepare(p.pSim, p.pAig, Vec_PtrSize(p.pWin.vLeaves), 0) //p->pPars->fVerbose );
		p.timeSim += util.Abc_Clock() - clk
		if RetValue == 0 {
			p.nSimEmpty++
			continue
		}

		// consider the case of constant node
		if p.pSim.fConst0 != 0 || p.pSim.fConst1 != 0 {
			p.nConstsUsed++

			if p.pSim.fConst1 != 0 {
				pFunc = hop.Hop_ManConst1(pNtk.FunctionManager.(*Hop_Man_t))
			} else {
				pFunc = hop.Hop_ManConst0(pNtk.FunctionManager.(*Hop_Man_t))
			}
			vFanins = Vec_VecEntry(p.vResubsW, 0)
			Vec_PtrClear(vFanins)
			Res_UpdateNetwork(pObj, vFanins, pFunc, p.vLevels)
			continue
		}

		//        printf( " " );

		// find resub candidates for the node
		clk = util.Abc_Clock()
		if p.pPars.fArea != 0 {
			RetValue = Res_FilterCandidates(p.pWin, p.pAig, p.pSim, p.vResubs, p.vResubsW, nFaninsMax, 1)
		} else {
			RetValue = Res_FilterCandidates(p.pWin, p.pAig, p.pSim, p.vResubs, p.vResubsW, nFaninsMax, 0)
		}
		p.timeCand += util.Abc_Clock() - clk
		p.nCandSets += RetValue
		if RetValue == 0 {
			continue
		}

		//        printf( "%d(%d) ", Vec_PtrSize(p->pWin->vDivs), RetValue );

		p.nWinsUsed++

		// iterate through candidate resubstitutions
		for k = 0; k < Vec_VecSize(p.vResubs); k++ {
			vFanins = Vec_VecEntry(p.vResubs, k)
			if Vec_PtrSize(vFanins) == 0 {
				break
			}

			// solve the SAT problem and get clauses
			clk = util.Abc_Clock()
			if p.pCnf != nil {
				bsat.Sto_ManFree(p.pCnf)
			}
			if proof := Res_SatProveUnsat(p.pAig, vFanins); proof != nil {
				p.pCnf, _ = proof.(*Sto_Man_t)
			} else {
				p.pCnf = nil
			}
			if p.pCnf == nil {
				p.timeSatSat += util.Abc_Clock() - clk
				//                printf( " Sat\n" );
				//                printf( "-" );
				continue
			}
			p.timeSatUnsat += util.Abc_Clock() - clk
			//            printf( "+" );

			p.nProvedSets++
			//            printf( " Unsat\n" );
			//            continue;
			//            printf( "Proved %d.\n", k );

			// write it into a file
			//            Sto_ManDumpClauses( p->pCnf, "trace.cnf" );

			// interpolate the problem if it was UNSAT
			clk = util.Abc_Clock()
			nFanins = bsat.Int_ManInterpolate(p.pMan, p.pCnf, 0, &puTruth)
			p.timeInt += util.Abc_Clock() - clk
			if nFanins != Vec_PtrSize(vFanins)-2 {
				continue
			}
			assert((puTruth != nil), "puTruth")
			//            Extra_PrintBinary( stdout, puTruth, 1 << nFanins );  printf( "\n" );

			// transform interpolant into the AIG
			pGraph = kit.Kit_TruthToGraph(puTruth, nFanins, &kitMemory)

			// derive the AIG for the decomposition tree
			pFunc = kit.Kit_GraphToHop(pNtk.FunctionManager.(*Hop_Man_t), pGraph)
			kit.Kit_GraphFree(pGraph)

			// update the network
			clk = util.Abc_Clock()
			Res_UpdateNetwork(pObj, Vec_VecEntry(p.vResubsW, k), pFunc, p.vLevels)
			p.timeUpd += util.Abc_Clock() - clk
			break
		}
		//        printf( "\n" );
	}
	extra.Extra_ProgressBarStop(pProgress)
	Abc_NtkStopReverseLevels(pNtk)

	p.timeSatSim += p.pSim.timeSat
	p.timeSatTotal = p.timeSatSat + p.timeSatUnsat + p.timeSatSim

	p.nTotalNets2 = abc.Abc_NtkGetTotalFanins(pNtk)
	p.nTotalNodes2 = abc.Abc_NtkNodeNum(pNtk)

	// quit resubstitution manager
	p.timeTotal = util.Abc_Clock() - clkTotal
	Res_ManFree(p)

	s_ResynTime += util.Abc_Clock() - clkTotal
	// check the resulting network
	if abc.Abc_NtkCheck(pNtk) == 0 {
		fmt.Fprintf(os.Stdout, "Abc_NtkResynthesize(): Network check has failed.\n")
		return 0
	}
	return 1
}

////////////////////////////////////////////////////////////////////////
///                       END OF FILE                                ///
////////////////////////////////////////////////////////////////////////

//ABC_NAMESPACE_IMPL_END
