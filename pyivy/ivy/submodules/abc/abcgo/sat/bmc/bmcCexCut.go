package bmc

import "fmt"

/**CFile****************************************************************

  FileName    [bmcCexCut.c]

  SystemName  [ABC: Logic synthesis and verification system.]

  PackageName [SAT-based bounded model checking.]

  Synopsis    [Derives characterization of bad states.]

  Author      [Alan Mishchenko]
  
  Affiliation [UC Berkeley]

  Date        [Ver. 1.0. Started - June 20, 2005.]

  Revision    [$Id: bmcCexCut.c,v 1.00 2005/06/20 00:00:00 alanmi Exp $]

***********************************************************************/

//#include "bmc.h"
//#include "aig/gia/giaAig.h"

//ABC_NAMESPACE_IMPL_START


////////////////////////////////////////////////////////////////////////
///                        DECLARATIONS                              ///
////////////////////////////////////////////////////////////////////////

////////////////////////////////////////////////////////////////////////
///                     FUNCTION DEFINITIONS                         ///
////////////////////////////////////////////////////////////////////////

/**Function*************************************************************

  Synopsis    [Generate justifying assignments.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Bmc_GiaGenerateJust_rec(p *Gia_Man_t, iFrame int32, iObj int32, vValues *Vec_Bit_t, vJustis *Vec_Bit_t) int32 {
    var pObj *Gia_Obj_t;
    var Shift int32  = Gia_ManObjNum(p) * iFrame;
    if iFrame < 0 {
        return 0; }
    assert(iFrame >= 0, "iFrame >= 0");
    if Vec_BitEntry( vJustis, Shift + iObj ) {
        return 0; }
    Vec_BitWriteEntry( vJustis, Shift + iObj, 1 );
    pObj = Gia_ManObj( p, iObj );
    if Gia_ObjIsCo(pObj) {
        return Bmc_GiaGenerateJust_rec( p, iFrame, Gia_ObjFaninId0(pObj, iObj), vValues, vJustis ); }
    if Gia_ObjIsCi(pObj) {
        return Bmc_GiaGenerateJust_rec( p, iFrame-1, Gia_ObjId(p, Gia_ObjRoToRi(p, pObj)), vValues, vJustis ); }
    assert(Gia_ObjIsAnd(pObj), "Gia_ObjIsAnd(pObj)");
    if Vec_BitEntry( vValues, Shift + iObj ) {
        Bmc_GiaGenerateJust_rec( p, iFrame, Gia_ObjFaninId0(pObj, iObj), vValues, vJustis );
        Bmc_GiaGenerateJust_rec( p, iFrame, Gia_ObjFaninId1(pObj, iObj), vValues, vJustis );
    
    } else if Vec_BitEntry( vValues, Shift + Gia_ObjFaninId0(pObj, iObj) ) == Gia_ObjFaninC0(pObj) {
        Bmc_GiaGenerateJust_rec( p, iFrame, Gia_ObjFaninId0(pObj, iObj), vValues, vJustis );
    } else if Vec_BitEntry( vValues, Shift + Gia_ObjFaninId1(pObj, iObj) ) == Gia_ObjFaninC1(pObj) {
        Bmc_GiaGenerateJust_rec( p, iFrame, Gia_ObjFaninId1(pObj, iObj), vValues, vJustis );
    } else { assert(false, "0"); }
    return 0;
}
func Bmc_GiaGenerateJustNonRec(p *Gia_Man_t, iFrame int32, vValues *Vec_Bit_t, vJustis *Vec_Bit_t) {
    var pObj *Gia_Obj_t;
    var i int32;
    var k int32;
    var Shift int32 = Gia_ManObjNum(p) * iFrame;
    for  i = iFrame; i >= 0; i--, Shift -= Gia_ManObjNum(p)  {
        // inlining of macro Gia_ManForEachObjReverse
        for  k = p.nObjs - 1; (k >= 0) && ((pObj) = Gia_ManObj(p, k)); k--  {
            if k == 0 || Gia_ObjIsPi(p, pObj) {
                continue; }
            if !Vec_BitEntry( vJustis, Shift + k ) {
                continue; }
            if Gia_ObjIsAnd(pObj) {
                if Vec_BitEntry( vValues, Shift + k ) {
                    Vec_BitWriteEntry( vJustis, Shift + Gia_ObjFaninId0(pObj, k), 1 );
                    Vec_BitWriteEntry( vJustis, Shift + Gia_ObjFaninId1(pObj, k), 1 );
                
                } else if Vec_BitEntry( vValues, Shift + Gia_ObjFaninId0(pObj, k) ) == Gia_ObjFaninC0(pObj) {
                    Vec_BitWriteEntry( vJustis, Shift + Gia_ObjFaninId0(pObj, k), 1 );
                } else if Vec_BitEntry( vValues, Shift + Gia_ObjFaninId1(pObj, k) ) == Gia_ObjFaninC1(pObj) {
                    Vec_BitWriteEntry( vJustis, Shift + Gia_ObjFaninId1(pObj, k), 1 );
                } else { assert(false, "0"); }
            
            } else if Gia_ObjIsCo(pObj) {
                Vec_BitWriteEntry( vJustis, Shift + Gia_ObjFaninId0(pObj, k), 1 );
            } else if Gia_ObjIsCi(pObj) && (i != 0) {
                Vec_BitWriteEntry( vJustis, Shift - Gia_ManObjNum(p) + Gia_ObjId(p, Gia_ObjRoToRi(p, pObj)), 1 ); }
        }
    }
}
func Bmc_GiaGenerateJust(p *Gia_Man_t, pCex *Abc_Cex_t, pvValues **Vec_Bit_t, pvJustis **Vec_Bit_t) {
    var vValues *Vec_Bit_t  = Vec_BitStart( Gia_ManObjNum(p) * (pCex.iFrame + 1) );
    var vJustis *Vec_Bit_t  = Vec_BitStart( Gia_ManObjNum(p) * (pCex.iFrame + 1) );
    var pObj *Gia_Obj_t;
    var i, k, iBit, fCompl0, fCompl1, fJusti0, fJusti1, Shift int32;

    Gia_ManCleanMark0(p);
    Gia_ManCleanMark1(p);
    // inlining of macro Gia_ManForEachRi
    for  k = 0; (k < Gia_ManRegNum(p)) && ((pObj) = Gia_ManCo(p, Gia_ManPoNum(p)+k)); k++  {
        pObj.fMark0 = Abc_InfoHasBit(pCex.pData, iBit++);
    }
    for  Shift = i = 0; i <= pCex.iFrame; i++, Shift += Gia_ManObjNum(p)  {
        // inlining of macro Gia_ManForEachObj
        for  k = 0; (k < p.nObjs) && ((pObj) = Gia_ManObj(p, k)); k++  {
            if Gia_ObjIsAnd(pObj) {
                fCompl0 = Gia_ObjFanin0(pObj).fMark0 ^ Gia_ObjFaninC0(pObj);
                fCompl1 = Gia_ObjFanin1(pObj).fMark0 ^ Gia_ObjFaninC1(pObj);
                fJusti0 = Gia_ObjFanin0(pObj).fMark1;
                fJusti1 = Gia_ObjFanin1(pObj).fMark1;
                pObj.fMark0 = fCompl0 & fCompl1;
                if pObj.fMark0 {
                    pObj.fMark1 = fJusti0 & fJusti1;
                } else if !fCompl0 && !(fCompl1 != 0) {
                    pObj.fMark1 = fJusti0 | fJusti1;
                } else if !fCompl0 {
                    pObj.fMark1 = fJusti0;
                } else if !(fCompl1 != 0) {
                    pObj.fMark1 = fJusti1;
                } else { assert(false, "0"); }
            
            } else if Gia_ObjIsCi(pObj) {
                if Gia_ObjIsPi(p, pObj) {
                    pObj.fMark0 = Abc_InfoHasBit(pCex.pData, iBit++);
                    pObj.fMark1 = 1;
                
                } else {
                    pObj.fMark0 = Gia_ObjRoToRi(p, pObj).fMark0;
                    pObj.fMark1 = Gia_ObjRoToRi(p, pObj).fMark1;
                }
            
            } else if Gia_ObjIsCo(pObj) {
                pObj.fMark0 = Gia_ObjFanin0(pObj).fMark0 ^ Gia_ObjFaninC0(pObj);
                pObj.fMark1 = Gia_ObjFanin0(pObj).fMark1;
            
            } else if Gia_ObjIsConst0(pObj) {
                pObj.fMark1 = 1;
            } else { assert(false, "0"); }
            if pObj.fMark0 {
                Vec_BitWriteEntry( vValues, Shift + k, 1 ); }
            if pObj.fMark1 {
                Vec_BitWriteEntry( vJustis, Shift + k, 1 ); }
        }
    }
    assert(iBit == pCex.nBits, "iBit == pCex.nBits");
    Gia_ManCleanMark0(p);
    Gia_ManCleanMark1(p);
    // perform backward traversal to mark just nodes
    pObj = Gia_ManPo( p, pCex.iPo );
    assert(Vec_BitEntry(vJustis, Gia_ManObjNum(p) * pCex.iFrame + Gia_ObjId(p, pObj)) == 0, "Vec_BitEntry(vJustis, Gia_ManObjNum(p) * pCex.iFrame + Gia_ObjId(p, pObj)) == 0");
//    Bmc_GiaGenerateJust_rec( p, pCex->iFrame, Gia_ObjId(p, pObj), vValues, vJustis );
    Vec_BitWriteEntry(vJustis, Gia_ManObjNum(p) * pCex.iFrame + Gia_ObjId(p, pObj), 1);
    Bmc_GiaGenerateJustNonRec( p, pCex.iFrame, vValues, vJustis );
    assert(Vec_BitEntry(vJustis, Gia_ManObjNum(p) * pCex.iFrame + Gia_ObjId(p, pObj)) == 1, "Vec_BitEntry(vJustis, Gia_ManObjNum(p) * pCex.iFrame + Gia_ObjId(p, pObj)) == 1");

    // return the result
    *pvValues = vValues;
    *pvJustis = vJustis;
}

/**Function*************************************************************

  Synopsis    []

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Bmc_GiaGenerateGiaOne(p *Gia_Man_t, pCex *Abc_Cex_t, pvInits **Vec_Bit_t, iFrBeg int32, iFrEnd int32) *Gia_Man_t {
    var vValues *Vec_Bit_t;
    var vJustis *Vec_Bit_t;
    var pNew *Gia_Man_t;
    var pObj *Gia_Obj_t;
    var k int32;
    var Cube int32 = 1;
    var Counter int32;
    Bmc_GiaGenerateJust( p, pCex, &vValues, &vJustis );
    // collect flop values in frame iFrBeg
    *pvInits = Vec_BitStart( Gia_ManRegNum(p) );
    // inlining of macro Gia_ManForEachRo
    for  k = 0; (k < Gia_ManRegNum(p)) && ((pObj) = Gia_ManCi(p, Gia_ManPiNum(p)+k)); k++  {
        if Vec_BitEntry(vValues, Gia_ManObjNum(p) * iFrBeg + Gia_ObjId(p, pObj)) {
            Vec_BitWriteEntry( *pvInits, k, 1 ); }
    }
    // create GIA with justified values in iFrEnd
    pNew = Gia_ManStart( 2 * Gia_ManRegNum(p) + 2 );
    pNew.pName = Abc_UtilStrsav( p.pName );
    // inlining of macro Gia_ManForEachRo
    for  k = 0; (k < Gia_ManRegNum(p)) && ((pObj) = Gia_ManCi(p, Gia_ManPiNum(p)+k)); k++  {
        var Literal int32  = Gia_ManAppendCi(pNew);
        if !Vec_BitEntry(vJustis, Gia_ManObjNum(p) * iFrEnd + Gia_ObjId(p, pObj)) {
            continue; }
        if Vec_BitEntry(vValues, Gia_ManObjNum(p) * iFrEnd + Gia_ObjId(p, pObj)) {
            Cube = Gia_ManAppendAnd( pNew, Cube, Literal );
        } else {
            Cube = Gia_ManAppendAnd( pNew, Cube, Abc_LitNot(Literal) ); }
        Counter++;
    }
//    printf( "Only %d flops (out of %d) belong to the care set.\n", Counter, Gia_ManRegNum(p) );
    Gia_ManAppendCo( pNew, Cube );
    Vec_BitFree( vValues );
    Vec_BitFree( vJustis );
    return pNew;
}

/**Function*************************************************************

  Synopsis    [Generates all frames from G to the last one.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Bmc_GiaGenerateGiaAllFrames(p *Gia_Man_t, pCex *Abc_Cex_t, pvInits **Vec_Bit_t, iFrBeg int32, iFrEnd int32) *Gia_Man_t {
    var vInitEnd *Vec_Bit_t;
    var pNew, pTemp *Gia_Man_t;
    var pObj, pObjRo, pObjRi *Gia_Obj_t;
    var f, i, k, iBitOld, iBit, fCompl0, fCompl1 int32;

    // skip trough the first iFrEnd frames
    Gia_ManCleanMark0(p);
    // inlining of macro Gia_ManForEachRo
    for  k = 0; (k < Gia_ManRegNum(p)) && ((pObj) = Gia_ManCi(p, Gia_ManPiNum(p)+k)); k++  {
        pObj.fMark0 = Abc_InfoHasBit(pCex.pData, iBit++);
    }
    *pvInits = Vec_BitStart( Gia_ManRegNum(p) );
    for  i = 0; i < iFrEnd; i++  {
        // remember values in frame iFrBeg
        if i == iFrBeg {
            // inlining of macro Gia_ManForEachRo
            for  k = 0; (k < Gia_ManRegNum(p)) && ((pObjRo) = Gia_ManCi(p, Gia_ManPiNum(p)+k)); k++  {
                if pObjRo.fMark0 {
                    Vec_BitWriteEntry( *pvInits, k, 1 ); }
            } }
        // simulate other values
        // inlining of macro Gia_ManForEachPi
        for  k = 0; (k < Gia_ManPiNum(p)) && ((pObj) = Gia_ManCi(p, k)); k++  {
            pObj.fMark0 = Abc_InfoHasBit(pCex.pData, iBit++);
        }
        // inlining of macro Gia_ManForEachAnd
        for  k = 0; (k < p.nObjs) && ((pObj) = Gia_ManObj(p, k)); k++  {      if !Gia_ObjIsAnd(pObj) {} else {
            pObj.fMark0 = (Gia_ObjFanin0(pObj).fMark0 ^ Gia_ObjFaninC0(pObj)) & 
                           (Gia_ObjFanin1(pObj).fMark0 ^ Gia_ObjFaninC1(pObj)); }
        }
        // inlining of macro Gia_ManForEachCo
        for  k = 0; (k < Vec_IntSize(p.vCos)) && ((pObj) = Gia_ManCo(p, k)); k++  {
            pObj.fMark0 = Gia_ObjFanin0(pObj).fMark0 ^ Gia_ObjFaninC0(pObj);
        }
        // inlining of macro Gia_ManForEachRiRo
        for  k = 0; (k < Gia_ManRegNum(p)) && ((pObjRi) = Gia_ManCo(p, Gia_ManPoNum(p)+k)) && ((pObjRo) = Gia_ManCi(p, Gia_ManPiNum(p)+k)); k++  {
            pObjRo.fMark0 = pObjRi.fMark0;
        }
    }
    assert(i == iFrEnd, "i == iFrEnd");
    vInitEnd = Vec_BitStart( Gia_ManRegNum(p) );
    // inlining of macro Gia_ManForEachRo
    for  k = 0; (k < Gia_ManRegNum(p)) && ((pObjRo) = Gia_ManCi(p, Gia_ManPiNum(p)+k)); k++  {
        if pObjRo.fMark0 {
            Vec_BitWriteEntry( vInitEnd, k, 1 ); }
    }

    // create new AIG manager
    pNew = Gia_ManStart( 10000 );
    pNew.pName = Abc_UtilStrsav( p.pName );
    // inlining of macro Gia_ManForEachRo
    for  k = 0; (k < Gia_ManRegNum(p)) && ((pObjRo) = Gia_ManCi(p, Gia_ManPiNum(p)+k)); k++  {
        Gia_ManAppendCi(pNew);
    }
    Gia_ManHashStart( pNew );

    Gia_ManConst0(p).Value = 1;
    // inlining of macro Gia_ManForEachPi
    for  k = 0; (k < Gia_ManPiNum(p)) && ((pObj) = Gia_ManCi(p, k)); k++  {
        pObj.Value = 1;
    }

    iBitOld = iBit;
    for  f = iFrEnd; f <= pCex.iFrame; f++  {
        // set up correct init state
        // inlining of macro Gia_ManForEachRo
        for  k = 0; (k < Gia_ManRegNum(p)) && ((pObjRo) = Gia_ManCi(p, Gia_ManPiNum(p)+k)); k++  {
            pObjRo.fMark0 = Vec_BitEntry( vInitEnd, k );
        }
        // simulate it for a few frames
        iBit = iBitOld;
        for  i = iFrEnd; i < f; i++  {
            // inlining of macro Gia_ManForEachPi
            for  k = 0; (k < Gia_ManPiNum(p)) && ((pObj) = Gia_ManCi(p, k)); k++  {
                pObj.fMark0 = Abc_InfoHasBit(pCex.pData, iBit++);
            }
            // inlining of macro Gia_ManForEachAnd
            for  k = 0; (k < p.nObjs) && ((pObj) = Gia_ManObj(p, k)); k++  {      if !Gia_ObjIsAnd(pObj) {} else {
                pObj.fMark0 = (Gia_ObjFanin0(pObj).fMark0 ^ Gia_ObjFaninC0(pObj)) & 
                               (Gia_ObjFanin1(pObj).fMark0 ^ Gia_ObjFaninC1(pObj)); }
            }
            // inlining of macro Gia_ManForEachCo
            for  k = 0; (k < Vec_IntSize(p.vCos)) && ((pObj) = Gia_ManCo(p, k)); k++  {
                pObj.fMark0 = Gia_ObjFanin0(pObj).fMark0 ^ Gia_ObjFaninC0(pObj);
            }
            // inlining of macro Gia_ManForEachRiRo
            for  k = 0; (k < Gia_ManRegNum(p)) && ((pObjRi) = Gia_ManCo(p, Gia_ManPoNum(p)+k)) && ((pObjRo) = Gia_ManCi(p, Gia_ManPiNum(p)+k)); k++  {
                pObjRo.fMark0 = pObjRi.fMark0;
            }
        }
        // start creating values
        // inlining of macro Gia_ManForEachRo
        for  k = 0; (k < Gia_ManRegNum(p)) && ((pObjRo) = Gia_ManCi(p, Gia_ManPiNum(p)+k)); k++  {
            pObjRo.Value = Abc_LitNotCond( Gia_Obj2Lit(pNew, Gia_ManPi(pNew, k)), !pObjRo.fMark0 );
        }
        for  i = f; i <= pCex.iFrame; i++  {
            // inlining of macro Gia_ManForEachPi
            for  k = 0; (k < Gia_ManPiNum(p)) && ((pObj) = Gia_ManCi(p, k)); k++  {
                pObj.fMark0 = Abc_InfoHasBit(pCex.pData, iBit++);
            }
            // inlining of macro Gia_ManForEachAnd
            for  k = 0; (k < p.nObjs) && ((pObj) = Gia_ManObj(p, k)); k++  {      if !Gia_ObjIsAnd(pObj) {} else {
                fCompl0 = Gia_ObjFanin0(pObj).fMark0 ^ Gia_ObjFaninC0(pObj);
                fCompl1 = Gia_ObjFanin1(pObj).fMark0 ^ Gia_ObjFaninC1(pObj);
                pObj.fMark0 = fCompl0 & fCompl1;
                if pObj.fMark0 {
                    pObj.Value = Gia_ManHashAnd( pNew, Gia_ObjFanin0(pObj).Value, Gia_ObjFanin1(pObj).Value );
                } else if !fCompl0 && !(fCompl1 != 0) {
                    pObj.Value = Gia_ManHashOr( pNew, Gia_ObjFanin0(pObj).Value, Gia_ObjFanin1(pObj).Value );
                } else if !fCompl0 {
                    pObj.Value = Gia_ObjFanin0(pObj).Value;
                } else if !(fCompl1 != 0) {
                    pObj.Value = Gia_ObjFanin1(pObj).Value;
                } else { assert(false, "0"); }
                assert(pObj.Value > 0, "pObj.Value > 0");
            }
            }
            // inlining of macro Gia_ManForEachCo
            for  k = 0; (k < Vec_IntSize(p.vCos)) && ((pObj) = Gia_ManCo(p, k)); k++  {
                pObj.fMark0 = Gia_ObjFanin0(pObj).fMark0 ^ Gia_ObjFaninC0(pObj);
                pObj.Value = Gia_ObjFanin0(pObj).Value;
                assert(pObj.Value > 0, "pObj.Value > 0");
            } 
            if i == pCex.iFrame {
                break; }
            // inlining of macro Gia_ManForEachRiRo
            for  k = 0; (k < Gia_ManRegNum(p)) && ((pObjRi) = Gia_ManCo(p, Gia_ManPoNum(p)+k)) && ((pObjRo) = Gia_ManCi(p, Gia_ManPiNum(p)+k)); k++  {
                pObjRo.fMark0 = pObjRi.fMark0;
                pObjRo.Value = pObjRi.Value;
            }
        }
        assert(iBit == pCex.nBits, "iBit == pCex.nBits");
        // create PO
        Gia_ManAppendCo( pNew, Gia_ManPo(p, pCex.iPo).Value );
    }
    Gia_ManHashStop( pNew );
    Vec_BitFree( vInitEnd );

    // cleanup
    pNew = Gia_ManCleanup( pTemp = pNew );
    Gia_ManStop( pTemp );
    return pNew;
}

/**Function*************************************************************

  Synopsis    [Generates one frame.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Bmc_GiaGenerateGiaAllOne(p *Gia_Man_t, pCex *Abc_Cex_t, pvInits **Vec_Bit_t, iFrBeg int32, iFrEnd int32) *Gia_Man_t {
    var pNew, pTemp *Gia_Man_t;
    var pObj, pObjRo, pObjRi *Gia_Obj_t;
    var i, k, iBit, fCompl0, fCompl1 int32;

    // skip trough the first iFrEnd frames
    Gia_ManCleanMark0(p);
    // inlining of macro Gia_ManForEachRo
    for  k = 0; (k < Gia_ManRegNum(p)) && ((pObj) = Gia_ManCi(p, Gia_ManPiNum(p)+k)); k++  {
        pObj.fMark0 = Abc_InfoHasBit(pCex.pData, iBit++);
    }
    *pvInits = Vec_BitStart( Gia_ManRegNum(p) );
    for  i = 0; i < iFrEnd; i++  {
        // remember values in frame iFrBeg
        if i == iFrBeg {
            // inlining of macro Gia_ManForEachRo
            for  k = 0; (k < Gia_ManRegNum(p)) && ((pObjRo) = Gia_ManCi(p, Gia_ManPiNum(p)+k)); k++  {
                if pObjRo.fMark0 {
                    Vec_BitWriteEntry( *pvInits, k, 1 ); }
            } }
        // simulate other values
        // inlining of macro Gia_ManForEachPi
        for  k = 0; (k < Gia_ManPiNum(p)) && ((pObj) = Gia_ManCi(p, k)); k++  {
            pObj.fMark0 = Abc_InfoHasBit(pCex.pData, iBit++);
        }
        // inlining of macro Gia_ManForEachAnd
        for  k = 0; (k < p.nObjs) && ((pObj) = Gia_ManObj(p, k)); k++  {      if !Gia_ObjIsAnd(pObj) {} else {
            pObj.fMark0 = (Gia_ObjFanin0(pObj).fMark0 ^ Gia_ObjFaninC0(pObj)) & 
                           (Gia_ObjFanin1(pObj).fMark0 ^ Gia_ObjFaninC1(pObj)); }
        }
        // inlining of macro Gia_ManForEachCo
        for  k = 0; (k < Vec_IntSize(p.vCos)) && ((pObj) = Gia_ManCo(p, k)); k++  {
            pObj.fMark0 = Gia_ObjFanin0(pObj).fMark0 ^ Gia_ObjFaninC0(pObj);
        }
        // inlining of macro Gia_ManForEachRiRo
        for  k = 0; (k < Gia_ManRegNum(p)) && ((pObjRi) = Gia_ManCo(p, Gia_ManPoNum(p)+k)) && ((pObjRo) = Gia_ManCi(p, Gia_ManPiNum(p)+k)); k++  {
            pObjRo.fMark0 = pObjRi.fMark0;
        }
    }
    assert(i == iFrEnd, "i == iFrEnd");

    // create new AIG manager
    pNew = Gia_ManStart( 10000 );
    pNew.pName = Abc_UtilStrsav( p.pName );
    Gia_ManConst0(p).Value = 1;
    // inlining of macro Gia_ManForEachPi
    for  k = 0; (k < Gia_ManPiNum(p)) && ((pObj) = Gia_ManCi(p, k)); k++  {
        pObj.Value = 1;
    }
    // inlining of macro Gia_ManForEachRo
    for  k = 0; (k < Gia_ManRegNum(p)) && ((pObjRo) = Gia_ManCi(p, Gia_ManPiNum(p)+k)); k++  {
        pObjRo.Value = Abc_LitNotCond( Gia_ManAppendCi(pNew), !pObjRo.fMark0 );
    }
    Gia_ManHashStart( pNew );
    for  i = iFrEnd; i <= pCex.iFrame; i++  {
        // inlining of macro Gia_ManForEachPi
        for  k = 0; (k < Gia_ManPiNum(p)) && ((pObj) = Gia_ManCi(p, k)); k++  {
            pObj.fMark0 = Abc_InfoHasBit(pCex.pData, iBit++);
        }
        // inlining of macro Gia_ManForEachAnd
        for  k = 0; (k < p.nObjs) && ((pObj) = Gia_ManObj(p, k)); k++  {      if !Gia_ObjIsAnd(pObj) {} else {
            fCompl0 = Gia_ObjFanin0(pObj).fMark0 ^ Gia_ObjFaninC0(pObj);
            fCompl1 = Gia_ObjFanin1(pObj).fMark0 ^ Gia_ObjFaninC1(pObj);
            pObj.fMark0 = fCompl0 & fCompl1;
            if pObj.fMark0 {
                pObj.Value = Gia_ManHashAnd( pNew, Gia_ObjFanin0(pObj).Value, Gia_ObjFanin1(pObj).Value );
            } else if !fCompl0 && !(fCompl1 != 0) {
                pObj.Value = Gia_ManHashOr( pNew, Gia_ObjFanin0(pObj).Value, Gia_ObjFanin1(pObj).Value );
            } else if !fCompl0 {
                pObj.Value = Gia_ObjFanin0(pObj).Value;
            } else if !(fCompl1 != 0) {
                pObj.Value = Gia_ObjFanin1(pObj).Value;
            } else { assert(false, "0"); }
            assert(pObj.Value > 0, "pObj.Value > 0");
        }
        }
        // inlining of macro Gia_ManForEachCo
        for  k = 0; (k < Vec_IntSize(p.vCos)) && ((pObj) = Gia_ManCo(p, k)); k++  {
            pObj.fMark0 = Gia_ObjFanin0(pObj).fMark0 ^ Gia_ObjFaninC0(pObj);
            pObj.Value = Gia_ObjFanin0(pObj).Value;
            assert(pObj.Value > 0, "pObj.Value > 0");
        } 
        if i == pCex.iFrame {
            break; }
        // inlining of macro Gia_ManForEachRiRo
        for  k = 0; (k < Gia_ManRegNum(p)) && ((pObjRi) = Gia_ManCo(p, Gia_ManPoNum(p)+k)) && ((pObjRo) = Gia_ManCi(p, Gia_ManPiNum(p)+k)); k++  {
            pObjRo.fMark0 = pObjRi.fMark0;
            pObjRo.Value = pObjRi.Value;
        }
    }
    Gia_ManHashStop( pNew );
    assert(iBit == pCex.nBits, "iBit == pCex.nBits");

    // create PO
    Gia_ManAppendCo( pNew, Gia_ManPo(p, pCex.iPo).Value );
    // cleanup
    pNew = Gia_ManCleanup( pTemp = pNew );
    Gia_ManStop( pTemp );
    return pNew;
}

/**Function*************************************************************

  Synopsis    [Generate GIA for target bad states.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Bmc_GiaTargetStates(p *Gia_Man_t, pCex *Abc_Cex_t, iFrBeg int32, iFrEnd int32, fCombOnly int32, fUseOne int32, fAllFrames int32, fVerbose int32) *Gia_Man_t {
    var pNew, pTemp *Gia_Man_t;
    var vInitNew *Vec_Bit_t;

    if iFrBeg < 0 { fmt.Printf("Starting frame is less than 0.\n"); return nil; }
    if iFrEnd < 0 { fmt.Printf("Stopping frame is less than 0.\n"); return nil; }
    if iFrBeg > pCex.iFrame { fmt.Printf("Starting frame is more than the last frame of CEX (%d).\n", pCex.iFrame); return nil; }
    if iFrEnd > pCex.iFrame { fmt.Printf("Stopping frame is more than the last frame of CEX (%d).\n", pCex.iFrame); return nil; }
    if iFrBeg > iFrEnd { fmt.Printf("Starting frame (%d) should be less than stopping frame (%d).\n", iFrBeg, iFrEnd); return nil; }
    assert(iFrBeg >= 0 && iFrBeg <= pCex.iFrame, "iFrBeg >= 0 && iFrBeg <= pCex.iFrame");
    assert(iFrEnd >= 0 && iFrEnd <= pCex.iFrame, "iFrEnd >= 0 && iFrEnd <= pCex.iFrame");
    assert(iFrBeg < iFrEnd, "iFrBeg < iFrEnd");

    if fUseOne != 0 {
        pNew = Bmc_GiaGenerateGiaOne( p, pCex, &vInitNew, iFrBeg, iFrEnd );
    } else if fAllFrames != 0 {
        pNew = Bmc_GiaGenerateGiaAllFrames( p, pCex, &vInitNew, iFrBeg, iFrEnd );
    } else { 
        pNew = Bmc_GiaGenerateGiaAllOne( p, pCex, &vInitNew, iFrBeg, iFrEnd ); }

    if !(fCombOnly != 0) {
        // create new GIA
        pNew = Gia_ManDupWithNewPo( p, pTemp = pNew ); 
        Gia_ManStop( pTemp );

        // create new initial state
        pNew = Gia_ManDupFlip( pTemp = pNew, Vec_BitArray(vInitNew) );
        Gia_ManStop( pTemp );
    }

    Vec_BitFree( vInitNew );
    return pNew;
}

/**Function*************************************************************

  Synopsis    [Generate AIG for target bad states.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Bmc_AigTargetStates(p *Aig_Man_t, pCex *Abc_Cex_t, iFrBeg int32, iFrEnd int32, fCombOnly int32, fUseOne int32, fAllFrames int32, fVerbose int32) *Aig_Man_t {
    var pAig *Aig_Man_t;
    var pGia, pRes *Gia_Man_t;
    pGia = Gia_ManFromAigSimple( p );
    if !Gia_ManVerifyCex( pGia, pCex, 0 ) {
        Abc_Print( 1, "Current CEX does not fail AIG \"%s\".\n", p.pName ); 
        Gia_ManStop( pGia );
        return nil;
    }
    pRes = Bmc_GiaTargetStates( pGia, pCex, iFrBeg, iFrEnd, fCombOnly, fUseOne, fAllFrames, fVerbose );
    Gia_ManStop( pGia );
    pAig = Gia_ManToAigSimple( pRes );
    Gia_ManStop( pRes );
    return pAig;
}

////////////////////////////////////////////////////////////////////////
///                       END OF FILE                                ///
////////////////////////////////////////////////////////////////////////


//ABC_NAMESPACE_IMPL_END

