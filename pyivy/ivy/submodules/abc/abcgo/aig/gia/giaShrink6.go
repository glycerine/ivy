package gia

import (
	"fmt"
	"time"
)

/**CFile****************************************************************

  FileName    [giaShrink6.c]

  SystemName  [ABC: Logic synthesis and verification system.]

  PackageName [Scalable AIG package.]

  Synopsis    [Implementation of DAG-aware unmapping for 6-input cuts.]

  Author      [Alan Mishchenko]
  
  Affiliation [UC Berkeley]

  Date        [Ver. 1.0. Started - June 20, 2005.]

  Revision    [$Id: giaShrink6.c,v 1.00 2005/06/20 00:00:00 alanmi Exp $]

***********************************************************************/

//#include "gia.h"
//#include "bool/bdc/bdc.h"
//#include "bool/rsb/rsb.h"
//#include "misc/util/utilTruth.h"

//ABC_NAMESPACE_IMPL_START

////////////////////////////////////////////////////////////////////////
///                        DECLARATIONS                              ///
////////////////////////////////////////////////////////////////////////

/**Function*************************************************************

  Synopsis    [Resynthesizes nodes using bi-decomposition.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Shr_ObjPerformBidec(p *Shr_Man_t, pManDec *Bdc_Man_t, pNew *Gia_Man_t, vLeafLits *Vec_Int_t, uTruth1 word, uTruthC word) int32 {
    var pFunc *Bdc_Fun_t;
    var pObj *Gia_Obj_t;
    var i, iVar, iLit, nNodes, iLast int32;
    var nVars int32  = Vec_IntSize(vLeafLits);
    assert(uTruth1 != 0 && uTruthC != 0, "uTruth1 != 0 && uTruthC != 0");
    Bdc_ManDecompose( pManDec, (unsigned *)&uTruth1, (unsigned *)&uTruthC, nVars, nil, 1000 );
    Bdc_FuncSetCopyInt( Bdc_ManFunc(pManDec, 0), 1 );
    // inlining of macro Vec_IntForEachEntry
    for  i = 0; (i < Vec_IntSize(vLeafLits)) && (((iVar) = Vec_IntEntry(vLeafLits, i)), 1); i++  {
        Bdc_FuncSetCopyInt( Bdc_ManFunc(pManDec, i+1), Abc_Var2Lit(iVar, 0) );
    }
    nNodes = Bdc_ManNodeNum( pManDec );
    for  i = nVars + 1; i < nNodes; i++  {
        pFunc = Bdc_ManFunc( pManDec, i );
        iLast = Gia_ManObjNum(pNew);
        iLit  = Gia_ManHashAnd( pNew, Bdc_FunFanin0Copy(pFunc), Bdc_FunFanin1Copy(pFunc) );
        Bdc_FuncSetCopyInt( pFunc, iLit );
        if iLast == Gia_ManObjNum(pNew) {
            continue; }
        assert(iLast + 1 == Gia_ManObjNum(pNew), "iLast + 1 == Gia_ManObjNum(pNew)");
        pObj = Gia_ManObj(pNew, Abc_Lit2Var(iLit));
        Gia_ObjSetAndLevel( pNew, pObj );
        Shr_ManAddFanout( p, Gia_ObjFaninId0p(pNew, pObj), Gia_ObjId(pNew, pObj) );
        Shr_ManAddFanout( p, Gia_ObjFaninId1p(pNew, pObj), Gia_ObjId(pNew, pObj) );
        assert(Gia_ManObjNum(pNew) < p.nNewSize, "Gia_ManObjNum(pNew) < p.nNewSize");
    }
    return Bdc_FunObjCopy( Bdc_ManRoot(pManDec) );
}


/**Function*************************************************************

  Synopsis    [Compute truth table.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Shr_ManComputeTruth6_rec(p *Gia_Man_t, iNode int32, vTruths *Vec_Wrd_t) word {
    var pObj *Gia_Obj_t;
    var Truth0, Truth1 word;
    if Gia_ObjIsTravIdCurrentId(p, iNode) {
        return Vec_WrdEntry(vTruths, iNode); }
    Gia_ObjSetTravIdCurrentId(p, iNode);
    pObj = Gia_ManObj( p, iNode );
    assert(Gia_ObjIsAnd(pObj), "Gia_ObjIsAnd(pObj)");
    Truth0 = Shr_ManComputeTruth6_rec( p, Gia_ObjFaninId0p(p, pObj), vTruths );
    Truth1 = Shr_ManComputeTruth6_rec( p, Gia_ObjFaninId1p(p, pObj), vTruths );
    if Gia_ObjFaninC0(pObj) {
        Truth0 = ~Truth0; }
    if Gia_ObjFaninC1(pObj) {
        Truth1 = ~Truth1; }
    Vec_WrdWriteEntry( vTruths, iNode, Truth0 & Truth1 );
    return Truth0 & Truth1;
}
func Shr_ManComputeTruth6(p *Gia_Man_t, pObj *Gia_Obj_t, vLeaves *Vec_Int_t, vTruths *Vec_Wrd_t) word {
    var i, iLeaf int32;
    assert(Gia_ObjIsAnd(pObj), "Gia_ObjIsAnd(pObj)");
    Gia_ManIncrementTravId( p );
    // inlining of macro Vec_IntForEachEntry
    for  i = 0; (i < Vec_IntSize(vLeaves)) && (((iLeaf) = Vec_IntEntry(vLeaves, i)), 1); i++  {
        Gia_ObjSetTravIdCurrentId( p, iLeaf );
        Vec_WrdWriteEntry( vTruths, iLeaf, Truth[i] );
    }
    return Shr_ManComputeTruth6_rec( p, Gia_ObjId(p, pObj), vTruths );
}

/**Function*************************************************************

  Synopsis    [Compute truth table.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
/**Function*************************************************************

  Synopsis    []

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Gia_ManMapShrink6(p *Gia_Man_t, nFanoutMax int32, fKeepLevel int32, fVerbose int32) *Gia_Man_t {
    var pMan *Shr_Man_t;
    var pObj, pFanin *Gia_Obj_t;
    var uTruth, uTruth0, uTruth1 word;
    var i, k, nDivs, iNode int32;
    var RetValue, Counter1, Counter2 int32;
    var clk2 abctime;
    var clk abctime = Abc_Clock();
    var timeFanout abctime;
    assert(Gia_ManHasMapping(p), "Gia_ManHasMapping(p)");
    pMan = Shr_ManAlloc( p );
    Gia_ManFillValue( p );
    Gia_ManConst0(p).Value = 0;
    // inlining of macro Gia_ManForEachObj1
    for  i = 1; (i < p.nObjs) && ((pObj) = Gia_ManObj(p, i)); i++  {
        if Gia_ObjIsCi(pObj) {
            pObj.Value = Gia_ManAppendCi( pMan.pNew );
            if p.vLevels {
                Gia_ObjSetLevel( pMan.pNew, Gia_ObjFromLit(pMan.pNew, Gia_ObjValue(pObj)), Gia_ObjLevel(p, pObj) ); }
        
        } else if Gia_ObjIsCo(pObj) {
            pObj.Value = Gia_ManAppendCo( pMan.pNew, Gia_ObjFanin0Copy(pObj) );
        
        } else if Gia_ObjIsLut(p, i) {
            // collect leaves of this gate  
            Vec_IntClear( pMan.vLeaves );
            // inlining of macro Gia_LutForEachFanin
            for  k = 0; k < Gia_ObjLutSize(p,i) && ((iNode = Gia_ObjLutFanins(p,i)[k]),1); k++  {
                Vec_IntPush( pMan.vLeaves, iNode );
            }
            assert(Vec_IntSize(pMan.vLeaves) <= 6, "Vec_IntSize(pMan.vLeaves) <= 6");
            // compute truth table 
            uTruth = Shr_ManComputeTruth6( pMan.pGia, pObj, pMan.vLeaves, pMan.vTruths );
            assert(pObj.Value == ~0, "pObj.Value == ~0");
            if uTruth == 0 || ~uTruth == 0 {
                pObj.Value = Abc_LitNotCond( 0, ~uTruth == 0 );
            } else {
                // inlining of macro Gia_ManForEachObjVec
                for  k = 0; (k < Vec_IntSize(pMan.vLeaves)) && ((pFanin) = Gia_ManObj(p, Vec_IntEntry(pMan.vLeaves,k))); k++  {
                    if uTruth == Truth[k] || ~uTruth == Truth[k] {
                        pObj.Value = Abc_LitNotCond( pFanin.Value, ~uTruth == Truth[k] ); }
                } }
            if pObj.Value != ~0 {
                continue; }
            // translate into new nodes
            // inlining of macro Gia_ManForEachObjVec
            for  k = 0; (k < Vec_IntSize(pMan.vLeaves)) && ((pFanin) = Gia_ManObj(p, Vec_IntEntry(pMan.vLeaves,k))); k++  {
                if Abc_LitIsCompl(pFanin.Value) {
                    uTruth = ((uTruth & Truth[k]) >> (1 << k)) | ((uTruth & ~Truth[k]) << (1 << k)); }
                Vec_IntWriteEntry( pMan.vLeaves, k, Abc_Lit2Var(pFanin.Value) );
            }
            // compute divisors
            clk2 = Abc_Clock();
            nDivs = Shr_ManCollectDivisors( pMan, pMan.vLeaves, pMan.nDivMax, nFanoutMax );
            assert(nDivs <= pMan.nDivMax, "nDivs <= pMan.nDivMax");
            timeFanout += Abc_Clock() - clk2;
            // compute truth tables
            Shr_ManComputeTruths( pMan.pNew, Vec_IntSize(pMan.vLeaves), pMan.vDivs, pMan.vDivTruths, pMan.vTruths );
            // perform resubstitution
            RetValue = Rsb_ManPerformResub6( pMan.pManRsb, Vec_IntSize(pMan.vLeaves), uTruth, pMan.vDivTruths, &uTruth0, &uTruth1, 0 );
            if RetValue { // resub exists
            
                var vResult *Vec_Int_t  = Rsb_ManGetFanins(pMan.pManRsb);
                Vec_IntClear( pMan.vDivResub );
                // inlining of macro Vec_IntForEachEntry
                for  k = 0; (k < Vec_IntSize(vResult)) && (((iNode) = Vec_IntEntry(vResult, k)), 1); k++  {
                    Vec_IntPush( pMan.vDivResub, Vec_IntEntry(pMan.vDivs, iNode) );
                }
                pObj.Value = Shr_ObjPerformBidec( pMan, pMan.pManDec, pMan.pNew, pMan.vDivResub, uTruth1, uTruth0 | uTruth1 );
                Counter1++;
            
            } else {
                pObj.Value = Shr_ObjPerformBidec( pMan, pMan.pManDec, pMan.pNew, pMan.vLeaves, uTruth, ~(word)0 );
                Counter2++;
            }
        }
    }
    if fVerbose != 0 {
        fmt.Printf("Performed %d resubs and %d decomps.  ", Counter1, Counter2);
        fmt.Printf("Gain in AIG nodes = %d.  ", Gia_ManObjNum(p)-Gia_ManObjNum(pMan.pNew));
        // inlining of macro ABC_PRT
        Abc_Print(1, "%s =", ("Runtime"));
Abc_Print(1, "%9.2f sec\n", 1.0*(float64((Abc_Clock() - clk)))/(float64(int64(time.Second))));;
//        ABC_PRT( "Divisors", timeFanout );        
    }
    return Shr_ManFree( pMan );
}

////////////////////////////////////////////////////////////////////////
///                       END OF FILE                                ///
////////////////////////////////////////////////////////////////////////


//ABC_NAMESPACE_IMPL_END

