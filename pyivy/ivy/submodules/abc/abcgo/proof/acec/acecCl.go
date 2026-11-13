package acec

import "fmt"

/**CFile****************************************************************

  FileName    [acecCl.c]

  SystemName  [ABC: Logic synthesis and verification system.]

  PackageName [CEC for arithmetic circuits.]

  Synopsis    [Core procedures.]

  Author      [Alan Mishchenko]
  
  Affiliation [UC Berkeley]

  Date        [Ver. 1.0. Started - June 20, 2005.]

  Revision    [$Id: acecCl.c,v 1.00 2005/06/20 00:00:00 alanmi Exp $]

***********************************************************************/

//#include "acecInt.h"
//#include "misc/vec/vecWec.h"
//#include "misc/extra/extra.h"

//ABC_NAMESPACE_IMPL_START

////////////////////////////////////////////////////////////////////////
///                        DECLARATIONS                              ///
////////////////////////////////////////////////////////////////////////



////////////////////////////////////////////////////////////////////////
///                     FUNCTION DEFINITIONS                         ///
////////////////////////////////////////////////////////////////////////

/**Function*************************************************************

  Synopsis    []

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Acec_RewriteTop(p *Gia_Man_t, pBox *Acec_Box_t) *Vec_Int_t {
    var vRes *Vec_Int_t  = Vec_IntAlloc( Gia_ManCoNum(p) + 1 );
    var vLevel *Vec_Int_t;
    var i, k, iStart, iLit, Driver, Count int32;
    // determine how much to shift
    Driver = Gia_ObjFaninId0p( p, Gia_ManCo(p, 0) );
    // inlining of macro Vec_WecForEachLevel
    for  iStart = 0; (iStart < Vec_WecSize(pBox.vRootLits)) && (((vLevel) = Vec_WecEntry(pBox.vRootLits, iStart)), 1); iStart++  {
        if Abc_Lit2Var(Vec_IntEntry(vLevel,0)) == Driver {
            break; }
    }
    assert(iStart < Gia_ManCoNum(p), "iStart < Gia_ManCoNum(p)");
    //Vec_WecPrintLits( pBox->vRootLits );
    // inlining of macro Vec_WecForEachLevelStart
    for  i = iStart; (i < Vec_WecSize(pBox.vRootLits)) && (((vLevel) = Vec_WecEntry(pBox.vRootLits, i)), 1); i++  {
        var In [3]int32 ;
        var Out [2]int32;
        assert(Vec_IntSize(vLevel) > 0, "Vec_IntSize(vLevel) > 0");
        assert(Vec_IntSize(vLevel) <= 3, "Vec_IntSize(vLevel) <= 3");
        if Vec_IntSize(vLevel) == 1 {
            Vec_IntPush( vRes, Vec_IntEntry(vLevel, 0) );
            continue;
        }
        // inlining of macro Vec_IntForEachEntry
        for  k = 0; (k < Vec_IntSize(vLevel)) && (((iLit) = Vec_IntEntry(vLevel, k)), 1); k++  {
            In[k] = iLit;
        }
        Acec_InsertFadd( p, In, Out );
        Vec_IntPush( vRes, Out[0] );
        if i+1 < Vec_WecSize(pBox.vRootLits) {
            Vec_IntPush( Vec_WecEntry(pBox.vRootLits, i+1), Out[1] );
        } else {
            Vec_IntPush( Vec_WecPushLevel(pBox.vRootLits), Out[1] ); }
        Count++;
    }
    assert(Vec_IntSize(vRes) >= Gia_ManCoNum(p), "Vec_IntSize(vRes) >= Gia_ManCoNum(p)");
    Vec_IntShrink( vRes, Gia_ManCoNum(p) );
    fmt.Printf("Added %d adders for replace CLAs.  ", Count);
    return vRes;
}
func Acec_RewriteReplace(p *Gia_Man_t, vRes *Vec_Int_t) *Gia_Man_t {
    var pNew, pTemp *Gia_Man_t;
    var pObj *Gia_Obj_t;  var i int32;
    assert(Gia_ManCoNum(p) == Vec_IntSize(vRes), "Gia_ManCoNum(p) == Vec_IntSize(vRes)");
    // create new manager
    pNew = Gia_ManStart( Gia_ManObjNum(p) );
    pNew.pName = Abc_UtilStrsav( p.pName );
    pNew.pSpec = Abc_UtilStrsav( p.pSpec );
    Gia_ManFillValue( p );
    Gia_ManConst0(p).Value = 0;
    // inlining of macro Gia_ManForEachCi
    for  i = 0; (i < Vec_IntSize(p.vCis)) && ((pObj) = Gia_ManCi(p, i)); i++  {
        pObj.Value = Gia_ManAppendCi(pNew);
    }
    // inlining of macro Gia_ManForEachAnd
    for  i = 0; (i < p.nObjs) && ((pObj) = Gia_ManObj(p, i)); i++  {      if !Gia_ObjIsAnd(pObj) {} else {
        pObj.Value = Gia_ManAppendAnd( pNew, Gia_ObjFanin0Copy(pObj), Gia_ObjFanin1Copy(pObj) ); }
    }
    // inlining of macro Gia_ManForEachCo
    for  i = 0; (i < Vec_IntSize(p.vCos)) && ((pObj) = Gia_ManCo(p, i)); i++  {
        var iLit int32  = Vec_IntEntry( vRes, i );
        var pRepr *Gia_Obj_t  = Gia_ManObj( p, Abc_Lit2Var(iLit) );
        pObj.Value = Gia_ManAppendCo( pNew, pRepr.Value );
    }
    // set correct phase
    Gia_ManSetPhase( p );
    Gia_ManSetPhase( pNew );
    // inlining of macro Gia_ManForEachCo
    for  i = 0; (i < Vec_IntSize(pNew.vCos)) && ((pObj) = Gia_ManCo(pNew, i)); i++  {
        if Gia_ObjPhase(pObj) != Gia_ObjPhase(Gia_ManCo(p, i)) {
            Gia_ObjFlipFaninC0( pObj ); }
    }
    // remove dangling nodes
    pNew = Gia_ManCleanup( pTemp = pNew );
    Gia_ManStop( pTemp );
    return pNew;
}
func Acec_ManDecla(pGia *Gia_Man_t, fBooth int32, fVerbose int32) *Gia_Man_t {
    var clk abctime  = Abc_Clock();
    var pNew *Gia_Man_t  = nil;
    var vIgnore *Vec_Bit_t;
    if fBooth != 0 {
        vIgnore = Acec_BoothFindPPG(pGia);
    } else {
        vIgnore = nil;
    }
    var pBox *Acec_Box_t  = Acec_DeriveBox( pGia, vIgnore, 0, 0, fVerbose );
    var vResult *Vec_Int_t;
    Vec_BitFreeP( &vIgnore );
    if pBox == nil { // cannot match
    
        fmt.Printf("Cannot find arithmetic boxes.\n");
        return Gia_ManDup( pGia );
    }
    vResult = Acec_RewriteTop( pGia, pBox );
    Acec_BoxFreeP( &pBox );
    pNew = Acec_RewriteReplace( pGia, vResult );
    Vec_IntFree( vResult );
    Abc_PrintTime( 1, "Time", Abc_Clock() - clk );
    return pNew;
}

////////////////////////////////////////////////////////////////////////
///                       END OF FILE                                ///
////////////////////////////////////////////////////////////////////////


//ABC_NAMESPACE_IMPL_END

