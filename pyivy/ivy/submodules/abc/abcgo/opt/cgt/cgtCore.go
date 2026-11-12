package cgt

import (
	"fmt"
	"os"
	"time"
)

/**CFile****************************************************************

  FileName    [cgtCore.c]

  SystemName  [ABC: Logic synthesis and verification system.]

  PackageName [Clock gating package.]

  Synopsis    []

  Author      [Alan Mishchenko]
  
  Affiliation [UC Berkeley]

  Date        [Ver. 1.0. Started - June 20, 2005.]

  Revision    [$Id: cgtCore.c,v 1.00 2005/06/20 00:00:00 alanmi Exp $]

***********************************************************************/

//#include "cgtInt.h"
//#include "misc/bar/bar.h"

//ABC_NAMESPACE_IMPL_START


////////////////////////////////////////////////////////////////////////
///                        DECLARATIONS                              ///
////////////////////////////////////////////////////////////////////////

////////////////////////////////////////////////////////////////////////
///                     FUNCTION DEFINITIONS                         ///
////////////////////////////////////////////////////////////////////////

/**Function*************************************************************

  Synopsis    [This procedure sets default parameters.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Cgt_SetDefaultParams(p *Cgt_Par_t) {
    memset( p, 0, sizeof(Cgt_Par_t) );
    p.nLevelMax  =    25;   // the max number of levels to look for clock-gates
    p.nCandMax   =  1000;   // the max number of candidates at each node
    p.nOdcMax    =     0;   // the max number of ODC levels to consider
    p.nConfMax   =    10;   // the max number of conflicts at a node
    p.nVarsMin   =  1000;   // the min number of vars to recycle the SAT solver
    p.nFlopsMin  =    10;   // the min number of flops to recycle the SAT solver
    p.fAreaOnly  =     0;   // derive clock-gating to minimize area
    p.fVerbose   =     0;   // verbosity flag
}

/**Function*************************************************************

  Synopsis    [Returns 1 if simulation does not filter out this candidate.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Cgt_SimulationFilter(p *Cgt_Man_t, pCandPart *Aig_Obj_t, pMiterPart *Aig_Obj_t) int32 {
    var pInfoCand, pInfoMiter *uint32;
    var w int32;
    var nWords int32 = Abc_BitWordNum( p.nPatts );  
    pInfoCand  = (unsigned *)Vec_PtrEntry( p.vPatts, Aig_ObjId(Aig_Regular(pCandPart)) );
    pInfoMiter = (unsigned *)Vec_PtrEntry( p.vPatts, Aig_ObjId(pMiterPart) );
    // C => !M -- true   is the same as    C & M -- false
    if !Aig_IsComplement(pCandPart) {
        for  w = 0; w < nWords; w++  {
            if pInfoCand[w] & pInfoMiter[w] {
                return 0; }
        }
    
    } else {
        for  w = 0; w < nWords; w++  {
            if ~pInfoCand[w] & pInfoMiter[w] {
                return 0; }
        }
    }
    return 1;
}

/**Function*************************************************************

  Synopsis    [Saves one simulation pattern.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Cgt_SimulationRecord(p *Cgt_Man_t) {
    var pObj *Aig_Obj_t;
    var i int32;
    // inlining of macro Aig_ManForEachObj
    // inlining of macro Vec_PtrForEachEntry
    for  i = 0; (i < Vec_PtrSize(p.pPart.vObjs)) && (((pObj) = (Aig_Obj_t *)Vec_PtrEntry(p.pPart.vObjs, i)), 1); i++  { if (pObj) == nil {} else
        if sat_solver_var_value( p.pSat, p.pCnf.pVarNums[i] ) {
            Abc_InfoSetBit( (unsigned *)Vec_PtrEntry(p.vPatts, i), p.nPatts ); }
    }
    p.nPatts++;
    if p.nPatts == 32 * p.nPattWords {
        Vec_PtrReallocSimInfo( p.vPatts );
        Vec_PtrCleanSimInfo( p.vPatts, p.nPattWords, 2 * p.nPattWords );
        p.nPattWords *= 2;
    }
}

/**Function*************************************************************

  Synopsis    [Performs clock-gating for the AIG.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Cgt_ClockGatingRangeCheck(p *Cgt_Man_t, iStart int32, nOutputs int32) {
    var vNodes *Vec_Ptr_t  = p.vFanout;
    var pMiter, pCand, pMiterFrame, pCandFrame, pMiterPart, pCandPart *Aig_Obj_t;
    var i, k, RetValue, nCalls int32;
    assert(Vec_VecSize(p.vGatesAll) == Aig_ManCoNum(p.pFrame), "Vec_VecSize(p.vGatesAll) == Aig_ManCoNum(p.pFrame)");
    // go through all the registers inputs of this range
    for  i = iStart; i < iStart + nOutputs; i++  {
        nCalls = p.nCalls;
        pMiter = Saig_ManLi( p.pAig, i );
        Cgt_ManDetectCandidates( p.pAig, p.vUseful, Aig_ObjFanin0(pMiter), p.pPars.nLevelMax, vNodes );
        // go through the candidates of this PO
        // inlining of macro Vec_PtrForEachEntry
        for  k = 0; (k < Vec_PtrSize(vNodes)) && (((pCand) = (Aig_Obj_t *)Vec_PtrEntry(vNodes, k)), 1); k++  {
            // get the corresponding nodes from the frames
            pCandFrame  = (Aig_Obj_t *)pCand.pData;
            pMiterFrame = (Aig_Obj_t *)pMiter.pData;
            // get the corresponding nodes from the part
            pCandPart   = (Aig_Obj_t *)pCandFrame.pData;
            pMiterPart  = (Aig_Obj_t *)pMiterFrame.pData;
            // try direct polarity
            if Cgt_SimulationFilter( p, pCandPart, pMiterPart ) != 0 {
                RetValue = Cgt_CheckImplication( p, pCandPart, pMiterPart );
                if RetValue == 1 {
                    Vec_VecPush( p.vGatesAll, i, pCand );
                    continue;
                }
                if RetValue == 0 {
                    Cgt_SimulationRecord( p ); }
            
            } else {
                p.nCallsFiltered++; }
            // try reverse polarity
            if Cgt_SimulationFilter( p, Aig_Not(pCandPart), pMiterPart ) != 0 {
                RetValue = Cgt_CheckImplication( p, Aig_Not(pCandPart), pMiterPart );
                if RetValue == 1 {
                    Vec_VecPush( p.vGatesAll, i, Aig_Not(pCand) );
                    continue;
                }
                if RetValue == 0 {
                    Cgt_SimulationRecord( p ); }
            
            } else {
                p.nCallsFiltered++; }
        }

        if p.pPars.fVerbose {
//            printf( "Flop %3d : Cand = %4d. Gate = %4d. SAT calls = %3d.\n", 
//                i, Vec_PtrSize(vNodes), Vec_PtrSize(Vec_VecEntry(p->vGatesAll, i)), p->nCalls-nCalls );
        }

    }
} 

/**Function*************************************************************

  Synopsis    [Performs clock-gating for the AIG.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Cgt_ClockGatingRange(p *Cgt_Man_t, iStart int32) int32 {
    var nOutputs, iStop int32;
    var clk abctime;
    var clkTotal abctime = Abc_Clock();
    var nCallsUnsat int32     = p.nCallsUnsat;
    var nCallsSat int32       = p.nCallsSat;
    var nCallsUndec int32     = p.nCallsUndec;
    var nCallsFiltered int32  = p.nCallsFiltered;
clk = Abc_Clock();
    p.pPart = Cgt_ManDupPartition( p.pFrame, p.pPars.nVarsMin, p.pPars.nFlopsMin, iStart, p.pCare, p.vSuppsInv, &nOutputs );
    p.pCnf  = Cnf_DeriveSimple( p.pPart, nOutputs );
    p.pSat  = (sat_solver *)Cnf_DataWriteIntoSolver( p.pCnf, 1, 0 );
    sat_solver_compress( p.pSat );
    p.vPatts = Vec_PtrAllocSimInfo( Aig_ManObjNumMax(p.pPart), p.nPattWords );
    Vec_PtrCleanSimInfo( p.vPatts, 0, p.nPattWords );
p.timePrepare += Abc_Clock() - clk;
    Cgt_ClockGatingRangeCheck( p, iStart, nOutputs );
    iStop = iStart + nOutputs;
    if p.pPars.fVeryVerbose {
        fmt.Printf("%5d : D =%4d. C =%5d. Var =%6d. Pr =%5d. Cex =%5d. F =%4d. Saved =%6d. ", iStart, iStop-iStart, Aig_ManCoNum(p.pPart)-nOutputs, p.pSat.size, p.nCallsUnsat-nCallsUnsat, p.nCallsSat  -nCallsSat, p.nCallsUndec-nCallsUndec, p.nCallsFiltered-nCallsFiltered);
        // inlining of macro ABC_PRT
        Abc_Print(1, "%s =", ("Time"));
Abc_Print(1, "%9.2f sec\n", 1.0*(float64((Abc_Clock() - clkTotal)))/(float64(int64(time.Second))));;
    }
    Cgt_ManClean( p );
    p.nRecycles++;
    return iStop;
}

/**Function*************************************************************

  Synopsis    [Performs clock-gating for the AIG.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Cgt_ClockGatingCandidates(pAig *Aig_Man_t, pCare *Aig_Man_t, pPars *Cgt_Par_t, vUseful *Vec_Int_t) *Vec_Vec_t {
    var pProgress *Bar_Progress_t  = nil;
    var Pars Cgt_Par_t; 
    var p *Cgt_Man_t;
    var vGatesAll *Vec_Vec_t;
    var iStart int32;
    var clk abctime = Abc_Clock();
    var clkTotal abctime = Abc_Clock();
    // reset random numbers
    Aig_ManRandom( 1 );
    if pPars == nil {
        Cgt_SetDefaultParams( pPars = &Pars ); }    
    p = Cgt_ManCreate( pAig, pCare, pPars );
    p.vUseful = vUseful;
    p.pFrame = Cgt_ManDeriveAigForGating( p );
p.timeAig += Abc_Clock() - clk;
    assert(Aig_ManCoNum(p.pFrame) == Saig_ManRegNum(p.pAig), "Aig_ManCoNum(p.pFrame) == Saig_ManRegNum(p.pAig)");
    pProgress = Bar_ProgressStart( os.Stdout, Aig_ManCoNum(p.pFrame) );
    for  iStart = 0; iStart < Aig_ManCoNum(p.pFrame);  {
        Bar_ProgressUpdate( pProgress, iStart, nil );
        iStart = Cgt_ClockGatingRange( p, iStart );
    }
    Bar_ProgressStop( pProgress );
    vGatesAll = p.vGatesAll;
    p.vGatesAll = nil;
p.timeTotal = Abc_Clock() - clkTotal;
    Cgt_ManStop( p );
    return vGatesAll;
}

/**Function*************************************************************

  Synopsis    [Performs clock-gating for the AIG.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Cgt_ClockGatingInt(pAig *Aig_Man_t, pCare *Aig_Man_t, pPars *Cgt_Par_t, vUseful *Vec_Int_t) *Vec_Vec_t {
    var vGatesAll, vGates *Vec_Vec_t;
    vGatesAll = Cgt_ClockGatingCandidates( pAig, pCare, pPars, vUseful );
    if pPars.fAreaOnly {
        vGates = Cgt_ManDecideArea( pAig, vGatesAll, pPars.nOdcMax, pPars.fVerbose );
    } else {
        vGates = Cgt_ManDecideSimple( pAig, vGatesAll, pPars.nOdcMax, pPars.fVerbose ); }
    Vec_VecFree( vGatesAll );
    return vGates;
}
func Cgt_ClockGating(pAig *Aig_Man_t, pCare *Aig_Man_t, pPars *Cgt_Par_t) *Aig_Man_t {
    var pGated *Aig_Man_t;
    var vGates *Vec_Vec_t  = Cgt_ClockGatingInt( pAig, pCare, pPars, nil );
    var nNodesUsed int32;
    if pPars.fVerbose {
//        printf( "Before CG: " );
//        Aig_ManPrintStats( pAig );
    }
    pGated = Cgt_ManDeriveGatedAig( pAig, vGates, pPars.fAreaOnly, &nNodesUsed );
    if pPars.fVerbose {
//        printf( "After  CG: " );
//        Aig_ManPrintStats( pGated );
        fmt.Printf("Nodes: Before CG = %6d. After CG = %6d. (%6.2f %%).  Total after CG = %6d.\n", Aig_ManNodeNum(pAig), nNodesUsed, 100.0*nNodesUsed/Aig_ManNodeNum(pAig), Aig_ManNodeNum(pGated));
    }
    Vec_VecFree( vGates );
    return pGated;
}

////////////////////////////////////////////////////////////////////////
///                       END OF FILE                                ///
////////////////////////////////////////////////////////////////////////


//ABC_NAMESPACE_IMPL_END

