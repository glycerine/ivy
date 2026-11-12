package abci

import "fmt"

/**CFile****************************************************************

  FileName    [abcBidec.c]

  SystemName  [ABC: Logic synthesis and verification system.]

  PackageName [Network and node package.]

  Synopsis    [Interface to bi-decomposition.]

  Author      [Alan Mishchenko]
  
  Affiliation [UC Berkeley]

  Date        [Ver. 1.0. Started - June 20, 2005.]

  Revision    [$Id: abcBidec.c,v 1.00 2005/06/20 00:00:00 alanmi Exp $]

***********************************************************************/

//#include "base/abc/abc.h"
//#include "bool/bdc/bdc.h"
//#include "bool/kit/kit.h"

//ABC_NAMESPACE_IMPL_START


////////////////////////////////////////////////////////////////////////
///                        DECLARATIONS                              ///
////////////////////////////////////////////////////////////////////////

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
func Abc_NodeIfNodeResyn(p *Bdc_Man_t, pHop *Hop_Man_t, pRoot *Hop_Obj_t, nVars int32, vTruth *Vec_Int_t, puCare *uint32, dProb float32) *Hop_Obj_t {
    var pTruth *uint32;
    var pFunc *Bdc_Fun_t;
    var nNodes, i int32;
    assert(nVars <= 16, "nVars <= 16");
    // derive truth table
    pTruth = Hop_ManConvertAigToTruth( pHop, Hop_Regular(pRoot), nVars, vTruth, 0 );
    if Hop_IsComplement(pRoot) {
        Extra_TruthNot( pTruth, pTruth, nVars ); }
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
func Abc_NtkBidecResyn(pNtk *Abc_Ntk_t, fVerbose int32) {
    var Pars Bdc_Par_t ;
    var pPars *Bdc_Par_t = &Pars;
    var p *Bdc_Man_t;
    var pObj *Abc_Obj_t;
    var vTruth *Vec_Int_t;
    var i, nGainTotal, nNodes1, nNodes2 int32;
    var clk abctime  = Abc_Clock();
    assert(Abc_NtkIsLogic(pNtk), "Abc_NtkIsLogic(pNtk)");
    if !Abc_NtkToAig(pNtk) {
        return; }
    pPars.nVarsMax = Abc_NtkGetFaninMax( pNtk );
    pPars.fVerbose = fVerbose;
    if pPars.nVarsMax > 15 {
        if fVerbose != 0 {
        fmt.Printf("Resynthesis is not performed for nodes with more than 15 inputs.\n"); }
        pPars.nVarsMax = 15;
    }
    vTruth = Vec_IntAlloc( 0 );
    p = Bdc_ManAlloc( pPars );
    Abc_NtkForEachNode( pNtk, pObj, i )
    {
        if Abc_ObjFaninNum(pObj) > 15 {
            continue; }
        nNodes1 = Hop_DagSize((Hop_Obj_t *)pObj.pData);
        pObj.pData = Abc_NodeIfNodeResyn( p, (Hop_Man_t *)pNtk.pManFunc, (Hop_Obj_t *)pObj.pData, Abc_ObjFaninNum(pObj), vTruth, nil, -1.0 );
        nNodes2 = Hop_DagSize((Hop_Obj_t *)pObj.pData);
        nGainTotal += nNodes1 - nNodes2;
    }
    Bdc_ManFree( p );
    Vec_IntFree( vTruth );
    if fVerbose != 0 {
    fmt.Printf("Total gain in AIG nodes = %d.  ", nGainTotal);
    ABC_PRT( "Total runtime", Abc_Clock() - clk );
    }
}


////////////////////////////////////////////////////////////////////////
///                       END OF FILE                                ///
////////////////////////////////////////////////////////////////////////


//ABC_NAMESPACE_IMPL_END

