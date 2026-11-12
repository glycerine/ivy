package nwk

import (
	"fmt"
	"time"
)

/**CFile****************************************************************

  FileName    [nwkBidec.c]

  SystemName  [ABC: Logic synthesis and verification system.]

  PackageName [Logic network representation.]

  Synopsis    [Bi-decomposition of local functions.]

  Author      [Alan Mishchenko]
  
  Affiliation [UC Berkeley]

  Date        [Ver. 1.0. Started - June 20, 2005.]

  Revision    [$Id: nwkBidec.c,v 1.00 2005/06/20 00:00:00 alanmi Exp $]

***********************************************************************/

//#include "nwk.h"

//ABC_NAMESPACE_IMPL_START


////////////////////////////////////////////////////////////////////////
///                        DECLARATIONS                              ///
////////////////////////////////////////////////////////////////////////

func Extra_TruthWordNum(nVars int32) int32 { if nVars <= 5 {
    return 1;
} else {
    return (1 << (nVars - 5));
} }
func Extra_TruthNot(pOut *uint32, pIn *uint32, nVars int32) {
    var w int32;
    for  w = Extra_TruthWordNum(nVars)-1; w >= 0; w--  {
        pOut[w] = ~pIn[w];
    }
}
func Extra_TruthOr(pOut *uint32, pIn0 *uint32, pIn1 *uint32, nVars int32) {
    var w int32;
    for  w = Extra_TruthWordNum(nVars)-1; w >= 0; w--  {
        pOut[w] = pIn0[w] | pIn1[w];
    }
}
func Extra_TruthSharp(pOut *uint32, pIn0 *uint32, pIn1 *uint32, nVars int32) {
    var w int32;
    for  w = Extra_TruthWordNum(nVars)-1; w >= 0; w--  {
        pOut[w] = pIn0[w] & ~pIn1[w];
    }
}

func Bdc_FunCopyHop(pObj *Bdc_Fun_t) *Hop_Obj_t { return Hop_NotCond( (Hop_Obj_t *)Bdc_FuncCopy(Bdc_Regular(pObj)), Bdc_IsComplement(pObj) );  }

////////////////////////////////////////////////////////////////////////
///                     FUNCTION DEFINITIONS                         ///
////////////////////////////////////////////////////////////////////////

/**Function*************************************************************

  Synopsis    [Resynthesizes nodes using bi-decomposition.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Nwk_NodeIfNodeResyn(p *Bdc_Man_t, pHop *Hop_Man_t, pRoot *Hop_Obj_t, nVars int32, vTruth *Vec_Int_t, puCare *uint32, dProb float32) *Hop_Obj_t {
    var pTruth *uint32;
    var pFunc *Bdc_Fun_t;
    var nNodes, i int32;
    assert(nVars <= 16, "nVars <= 16");
    // derive truth table
    pTruth = Hop_ManConvertAigToTruth( pHop, Hop_Regular(pRoot), nVars, vTruth, 0 );
    if Hop_IsComplement(pRoot) {
        for  i = Abc_TruthWordNum(nVars)-1; i >= 0; i--  {
            pTruth[i] = ~pTruth[i];
        } }
    // perform power-aware decomposition
    if dProb >= 0.0 {
        var Prob float32  = float32(2.0) * dProb * (1.0 - dProb);
        assert(Prob >= 0.0 && Prob <= 0.5, "Prob >= 0.0 && Prob <= 0.5");
        if Prob >= 0.4 {
            Extra_TruthNot( puCare, puCare, nVars );
            if dProb > 0.5 { // more 1s than 0s
                Extra_TruthOr( pTruth, pTruth, puCare, nVars );
            } else {
                Extra_TruthSharp( pTruth, pTruth, puCare, nVars ); }
            Extra_TruthNot( puCare, puCare, nVars );
            // decompose truth table
            Bdc_ManDecompose( p, pTruth, nil, nVars, nil, 1000 );
        
        } else {
            // decompose truth table
            Bdc_ManDecompose( p, pTruth, puCare, nVars, nil, 1000 );
        }
    
    } else {
        // decompose truth table
        Bdc_ManDecompose( p, pTruth, puCare, nVars, nil, 1000 );
    }
    // convert back into HOP
    Bdc_FuncSetCopy( Bdc_ManFunc( p, 0 ), Hop_ManConst1( pHop ) );
    for  i = 0; i < nVars; i++  {
        Bdc_FuncSetCopy( Bdc_ManFunc( p, i+1 ), Hop_ManPi( pHop, i ) );
    }
    nNodes = Bdc_ManNodeNum(p);
    for  i = nVars + 1; i < nNodes; i++  {
        pFunc = Bdc_ManFunc( p, i );
        Bdc_FuncSetCopy( pFunc, Hop_And( pHop, Bdc_FunCopyHop(Bdc_FuncFanin0(pFunc)), Bdc_FunCopyHop(Bdc_FuncFanin1(pFunc)) ) );
    }
    return Bdc_FunCopyHop( Bdc_ManRoot(p) );
}

/**Function*************************************************************

  Synopsis    [Resynthesizes nodes using bi-decomposition.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Nwk_ManBidecResyn(pNtk *Nwk_Man_t, fVerbose int32) {
    var Pars Bdc_Par_t ;
    var pPars *Bdc_Par_t = &Pars;
    var p *Bdc_Man_t;
    var pObj *Nwk_Obj_t;
    var vTruth *Vec_Int_t;
    var i, nGainTotal, nNodes1, nNodes2 int32;
    var clk abctime  = Abc_Clock();
    pPars.nVarsMax = Nwk_ManGetFaninMax( pNtk );
    pPars.fVerbose = fVerbose;
    if pPars.nVarsMax < 2 {
        fmt.Printf("Resynthesis is not performed for networks whose nodes are less than 2 inputs.\n");
        return;
    }
    if pPars.nVarsMax > 15 {
        if fVerbose != 0 {
        fmt.Printf("Resynthesis is not performed for nodes with more than 15 inputs.\n"); }
        pPars.nVarsMax = 15;
    }
    vTruth = Vec_IntAlloc( 0 );
    p = Bdc_ManAlloc( pPars );
    // inlining of macro Nwk_ManForEachNode
    for  i = 0; (i < Vec_PtrSize(pNtk.vObjs)) && (((pObj) = (Nwk_Obj_t *)Vec_PtrEntry(pNtk.vObjs, i)), 1); i++  {         if (pObj) == nil || !Nwk_ObjIsNode(pObj) {} else {
        if Nwk_ObjFaninNum(pObj) > 15 {
            continue; }
        nNodes1 = Hop_DagSize(pObj.pFunc);
        pObj.pFunc = Nwk_NodeIfNodeResyn( p, pNtk.pManHop, pObj.pFunc, Nwk_ObjFaninNum(pObj), vTruth, nil, -1.0 );
        nNodes2 = Hop_DagSize(pObj.pFunc);
        nGainTotal += nNodes1 - nNodes2;
    }
    }
    Bdc_ManFree( p );
    Vec_IntFree( vTruth );
    if fVerbose != 0 {
    fmt.Printf("Total gain in AIG nodes = %d.  ", nGainTotal);
    // inlining of macro ABC_PRT
    Abc_Print(1, "%s =", ("Total runtime"));
Abc_Print(1, "%9.2f sec\n", 1.0*(float64((Abc_Clock() - clk)))/(float64(int64(time.Second))));;
    }
}


////////////////////////////////////////////////////////////////////////
///                       END OF FILE                                ///
////////////////////////////////////////////////////////////////////////


//ABC_NAMESPACE_IMPL_END

