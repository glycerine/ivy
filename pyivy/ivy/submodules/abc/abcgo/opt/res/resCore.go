package res

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
    pPars    *Res_Par_t;
    // specialized manager
    pWin    *Res_Win_t;          // windowing manager
    pAig    *Abc_Ntk_t;          // the strashed window
    pSim    *Res_Sim_t;          // simulation manager
    pCnf    *Sto_Man_t;          // the CNF of the SAT problem
    pMan    *Int_Man_t;          // interpolation manager;
    vMem    *Vec_Int_t;          // memory for intermediate SOPs
    vResubs    *Vec_Vec_t;       // resubstitution candidates of the AIG
    vResubsW    *Vec_Vec_t;      // resubstitution candidates of the window
    vLevels    *Vec_Vec_t;       // levelized structure for updating
    // statistics
    nWins           int32;         // the number of windows tried
    nWinNodes           int32;     // the total number of window nodes
    nDivNodes           int32;     // the total number of divisors
    nWinsTriv           int32;     // the total number of trivial windows
    nWinsUsed           int32;     // the total number of useful windows (with at least one candidate)
    nConstsUsed           int32;   // the total number of constant nodes under ODC
    nCandSets           int32;     // the total number of candidates
    nProvedSets           int32;   // the total number of proved groups
    nSimEmpty           int32;     // the empty simulation info
    nTotalNets           int32;    // the total number of nets
    nTotalNodes           int32;   // the total number of nodess
    nTotalNets2           int32;   // the total number of nets
    nTotalNodes2           int32;  // the total number of nodess
    // runtime
    timeWin       abctime;       // windowing
    timeDiv       abctime;       // divisors
    timeAig       abctime;       // strashing
    timeSim       abctime;       // simulation
    timeCand       abctime;      // resubstitution candidates
    timeSatTotal       abctime;  // SAT solving total 
    timeSatSat       abctime;    // SAT solving (sat calls)
    timeSatUnsat       abctime;  // SAT solving (unsat calls)
    timeSatSim       abctime;    // SAT solving (simulation)
    timeInt       abctime;       // interpolation 
    timeUpd       abctime;       // updating  
    timeTotal       abctime;     // total runtime
}

// extern Hop_Obj_t * Kit_GraphToHop( Hop_Man_t * pMan, Kit_Graph_t * pGraph );


var s_ResynTime abctime;

////////////////////////////////////////////////////////////////////////
///                     FUNCTION DEFINITIONS                         ///
////////////////////////////////////////////////////////////////////////

/**Function*************************************************************

  Synopsis    [Allocate resynthesis manager.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Res_ManAlloc(pPars *Res_Par_t) *Res_Man_t {
    var p *Res_Man_t;
    p = new(Res_Man_t);
    *p = Res_Man_t{};
    assert(pPars.nWindow > 0 && pPars.nWindow < 100, "pPars.nWindow > 0 && pPars.nWindow < 100");
    assert(pPars.nCands > 0 && pPars.nCands < 100, "pPars.nCands > 0 && pPars.nCands < 100");
    p.pPars = pPars;
    p.pWin = Res_WinAlloc();
    p.pSim = Res_SimAlloc( pPars.nSimWords );
    p.pMan = Int_ManAlloc();
    p.vMem = Vec_IntAlloc( 0 );
    p.vResubs  = Vec_VecStart( pPars.nCands );
    p.vResubsW = Vec_VecStart( pPars.nCands );
    p.vLevels  = Vec_VecStart( 32 );
    return p;
}

/**Function*************************************************************

  Synopsis    [Deallocate resynthesis manager.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Res_ManFree(p *Res_Man_t) {
    if p.pPars.fVerbose {
        fmt.Printf("Reduction in nodes = %5d. (%.2f %%) ", p.nTotalNodes-p.nTotalNodes2, 100.0*(p.nTotalNodes-p.nTotalNodes2)/p.nTotalNodes);
        fmt.Printf("Reduction in edges = %5d. (%.2f %%) ", p.nTotalNets-p.nTotalNets2, 100.0*(p.nTotalNets-p.nTotalNets2)/p.nTotalNets);
        fmt.Printf("\n");

        fmt.Printf("Winds = %d. ", p.nWins);
        fmt.Printf("Nodes = %d. (Ave = %5.1f)  ", p.nWinNodes, 1.0*p.nWinNodes/p.nWins);
        fmt.Printf("Divs = %d. (Ave = %5.1f)  ", p.nDivNodes, 1.0*p.nDivNodes/p.nWins);
        fmt.Printf("\n");
        fmt.Printf("WinsTriv = %d. ", p.nWinsTriv);
        fmt.Printf("SimsEmpt = %d. ", p.nSimEmpty);
        fmt.Printf("Const = %d. ", p.nConstsUsed);
        fmt.Printf("WindUsed = %d. ", p.nWinsUsed);
        fmt.Printf("Cands = %d. ", p.nCandSets);
        fmt.Printf("Proved = %d.", p.nProvedSets);
        fmt.Printf("\n");

        ABC_PRTP( "Windowing  ", p.timeWin,      p.timeTotal );
        ABC_PRTP( "Divisors   ", p.timeDiv,      p.timeTotal );
        ABC_PRTP( "Strashing  ", p.timeAig,      p.timeTotal );
        ABC_PRTP( "Simulation ", p.timeSim,      p.timeTotal );
        ABC_PRTP( "Candidates ", p.timeCand,     p.timeTotal );
        ABC_PRTP( "SAT solver ", p.timeSatTotal, p.timeTotal );
        ABC_PRTP( "    sat    ", p.timeSatSat,   p.timeTotal );
        ABC_PRTP( "    unsat  ", p.timeSatUnsat, p.timeTotal );
        ABC_PRTP( "    simul  ", p.timeSatSim,   p.timeTotal );
        ABC_PRTP( "Interpol   ", p.timeInt,      p.timeTotal );
        ABC_PRTP( "Undating   ", p.timeUpd,      p.timeTotal );
        ABC_PRTP( "TOTAL      ", p.timeTotal,    p.timeTotal );
    }
    Res_WinFree( p.pWin );
    if p.pAig { Abc_NtkDelete( p.pAig ); }
    Res_SimFree( p.pSim );
    if p.pCnf { Sto_ManFree( p.pCnf ); }
    Int_ManFree( p.pMan );
    Vec_IntFree( p.vMem );
    Vec_VecFree( p.vResubs );
    Vec_VecFree( p.vResubsW );
    Vec_VecFree( p.vLevels );
    ABC_FREE( p );
}

/**Function*************************************************************

  Synopsis    [Incrementally updates level of the nodes.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Res_UpdateNetwork(pObj *Abc_Obj_t, vFanins *Vec_Ptr_t, pFunc *Hop_Obj_t, vLevels *Vec_Vec_t) {
    var pObjNew, pFanin *Abc_Obj_t;
    var k int32;

    // create the new node
    pObjNew = Abc_NtkCreateNode( pObj.pNtk );
    pObjNew.pData = pFunc;
    Vec_PtrForEachEntry( Abc_Obj_t *, vFanins, pFanin, k )
        Abc_ObjAddFanin( pObjNew, pFanin );
    // replace the old node by the new node
//printf( "Replacing node " ); Abc_ObjPrint( stdout, pObj );
//printf( "Inserting node " ); Abc_ObjPrint( stdout, pObjNew );
    // update the level of the node
    Abc_NtkUpdate( pObj, pObjNew, vLevels );
}

/**Function*************************************************************

  Synopsis    [Entrace into the resynthesis package.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Abc_NtkResynthesize(pNtk *Abc_Ntk_t, pPars *Res_Par_t) int32 {
    var pProgress *ProgressBar;
    var p *Res_Man_t;
    var pObj *Abc_Obj_t;
    var pFunc *Hop_Obj_t;
    var pGraph *Kit_Graph_t;
    var vFanins *Vec_Ptr_t;
    var puTruth *uint32;
    var i, k, RetValue, nNodesOld, nFanins, nFaninsMax int32;
    var clk abctime;
    var clkTotal abctime = Abc_Clock();

    // start the manager
    p = Res_ManAlloc( pPars );
    p.nTotalNets = Abc_NtkGetTotalFanins(pNtk);
    p.nTotalNodes = Abc_NtkNodeNum(pNtk);
    nFaninsMax = Abc_NtkGetFaninMax(pNtk);
    if nFaninsMax > 8 {
        nFaninsMax = 8; }

    // perform the network sweep
    Abc_NtkSweep( pNtk, 0 );

    // convert into the AIG
    if !Abc_NtkToAig(pNtk) {
        fmt.Fprintf(os.Stdout, "Converting to BDD has failed.\n");
        Res_ManFree( p );
        return 0;
    }
    assert(Abc_NtkHasAig(pNtk), "Abc_NtkHasAig(pNtk)");

    // set the number of levels
    Abc_NtkLevel( pNtk );
    Abc_NtkStartReverseLevels( pNtk, pPars.nGrowthLevel );

    // try resynthesizing nodes in the topological order
    nNodesOld = Abc_NtkObjNumMax(pNtk);
    pProgress = Extra_ProgressBarStart( os.Stdout, nNodesOld );
    Abc_NtkForEachObj( pNtk, pObj, i )
    {
        Extra_ProgressBarUpdate( pProgress, i, nil );
        if !Abc_ObjIsNode(pObj) {
            continue; }
        if Abc_ObjFaninNum(pObj) > 8 {
            continue; }
        if pObj.Id > nNodesOld {
            break; }

        // create the window for this node
clk = Abc_Clock();
        RetValue = Res_WinCompute( pObj, p.pPars.nWindow/10, p.pPars.nWindow%10, p.pWin );
p.timeWin += Abc_Clock() - clk;
        if !RetValue {
            continue; }
        p.nWinsTriv += Res_WinIsTrivial( p.pWin );

        if p.pPars.fVeryVerbose {
            fmt.Printf("%5d (lev=%2d) : ", pObj.Id, pObj.Level);
            fmt.Printf("Win = %3d/%3d/%4d/%3d   ", Vec_PtrSize(p.pWin.vLeaves), Vec_PtrSize(p.pWin.vBranches), Vec_PtrSize(p.pWin.vNodes), Vec_PtrSize(p.pWin.vRoots));
        }

        // collect the divisors
clk = Abc_Clock();
        Res_WinDivisors( p.pWin, Abc_ObjRequiredLevel(pObj) - 1 );
p.timeDiv += Abc_Clock() - clk;

        p.nWins++;
        p.nWinNodes += Vec_PtrSize(p.pWin.vNodes);
        p.nDivNodes += Vec_PtrSize( p.pWin.vDivs);

        if p.pPars.fVeryVerbose {
            fmt.Printf("D = %3d ", Vec_PtrSize(p.pWin.vDivs));
            fmt.Printf("D+ = %3d ", p.pWin.nDivsPlus);
        }

        // create the AIG for the window
clk = Abc_Clock();
        if p.pAig { Abc_NtkDelete( p.pAig ); }
        p.pAig = Res_WndStrash( p.pWin );
p.timeAig += Abc_Clock() - clk;

        if p.pPars.fVeryVerbose {
            fmt.Printf("AIG = %4d ", Abc_NtkNodeNum(p.pAig));
            fmt.Printf("\n");
        }
 
        // prepare simulation info
clk = Abc_Clock();
        RetValue = Res_SimPrepare( p.pSim, p.pAig, Vec_PtrSize(p.pWin.vLeaves), 0 ); //p->pPars->fVerbose );
p.timeSim += Abc_Clock() - clk;
        if !RetValue {
            p.nSimEmpty++;
            continue;
        }

        // consider the case of constant node
        if p.pSim.fConst0 || p.pSim.fConst1 {
            p.nConstsUsed++;

            if p.pSim.fConst1 {
                pFunc = Hop_ManConst1((Hop_Man_t *)pNtk.pManFunc);
            } else {
                pFunc = Hop_ManConst0((Hop_Man_t *)pNtk.pManFunc);
            }
            vFanins = Vec_VecEntry( p.vResubsW, 0 );
            Vec_PtrClear( vFanins );
            Res_UpdateNetwork( pObj, vFanins, pFunc, p.vLevels );
            continue;
        }

//        printf( " " );

        // find resub candidates for the node
clk = Abc_Clock();
        if p.pPars.fArea {
            RetValue = Res_FilterCandidates( p.pWin, p.pAig, p.pSim, p.vResubs, p.vResubsW, nFaninsMax, 1 );
        } else {
            RetValue = Res_FilterCandidates( p.pWin, p.pAig, p.pSim, p.vResubs, p.vResubsW, nFaninsMax, 0 ); }
p.timeCand += Abc_Clock() - clk;
        p.nCandSets += RetValue;
        if RetValue == 0 {
            continue; }

//        printf( "%d(%d) ", Vec_PtrSize(p->pWin->vDivs), RetValue );

        p.nWinsUsed++;

        // iterate through candidate resubstitutions
        Vec_VecForEachLevel( p.vResubs, vFanins, k )
        {
            if Vec_PtrSize(vFanins) == 0 {
                break; }

            // solve the SAT problem and get clauses
clk = Abc_Clock();
            if p.pCnf { Sto_ManFree( p.pCnf ); }
            p.pCnf = (Sto_Man_t *)Res_SatProveUnsat( p.pAig, vFanins );
            if p.pCnf == nil {
p.timeSatSat += Abc_Clock() - clk;
//                printf( " Sat\n" );
//                printf( "-" );
                continue;
            }
p.timeSatUnsat += Abc_Clock() - clk;
//            printf( "+" );

            p.nProvedSets++;
//            printf( " Unsat\n" );
//            continue;
//            printf( "Proved %d.\n", k );

            // write it into a file
//            Sto_ManDumpClauses( p->pCnf, "trace.cnf" );

            // interpolate the problem if it was UNSAT
clk = Abc_Clock();
            nFanins = Int_ManInterpolate( p.pMan, p.pCnf, 0, &puTruth );
p.timeInt += Abc_Clock() - clk;
            if nFanins != Vec_PtrSize(vFanins) - 2 {
                continue; }
            assert((puTruth != nil), "puTruth");
//            Extra_PrintBinary( stdout, puTruth, 1 << nFanins );  printf( "\n" );

            // transform interpolant into the AIG
            pGraph = Kit_TruthToGraph( puTruth, nFanins, p.vMem );

            // derive the AIG for the decomposition tree
            pFunc = Kit_GraphToHop( (Hop_Man_t *)pNtk.pManFunc, pGraph );
            Kit_GraphFree( pGraph );

            // update the network
clk = Abc_Clock();
            Res_UpdateNetwork( pObj, Vec_VecEntry(p.vResubsW, k), pFunc, p.vLevels );
p.timeUpd += Abc_Clock() - clk;
            break;
        }
//        printf( "\n" );
    }
    Extra_ProgressBarStop( pProgress );
    Abc_NtkStopReverseLevels( pNtk );

p.timeSatSim += p.pSim.timeSat;
p.timeSatTotal = p.timeSatSat + p.timeSatUnsat + p.timeSatSim;

    p.nTotalNets2 = Abc_NtkGetTotalFanins(pNtk);
    p.nTotalNodes2 = Abc_NtkNodeNum(pNtk);

    // quit resubstitution manager
p.timeTotal = Abc_Clock() - clkTotal;
    Res_ManFree( p );

s_ResynTime += Abc_Clock() - clkTotal;
    // check the resulting network
    if !Abc_NtkCheck( pNtk ) {
        fmt.Fprintf(os.Stdout, "Abc_NtkResynthesize(): Network check has failed.\n");
        return 0;
    }
    return 1;
}

////////////////////////////////////////////////////////////////////////
///                       END OF FILE                                ///
////////////////////////////////////////////////////////////////////////


//ABC_NAMESPACE_IMPL_END

