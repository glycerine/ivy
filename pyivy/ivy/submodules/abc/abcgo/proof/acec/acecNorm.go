package acec

/**CFile****************************************************************

  FileName    [acecNorm.c]

  SystemName  [ABC: Logic synthesis and verification system.]

  PackageName [CEC for arithmetic circuits.]

  Synopsis    [Adder tree normalization.]

  Author      [Alan Mishchenko]
  
  Affiliation [UC Berkeley]

  Date        [Ver. 1.0. Started - June 20, 2005.]

  Revision    [$Id: acecNorm.c,v 1.00 2005/06/20 00:00:00 alanmi Exp $]

***********************************************************************/

//#include "acecInt.h"

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
func Acec_InsertHadd(pNew *Gia_Man_t, In []int32, Out []int32) {
    var And, Or int32;
    Out[1] = Gia_ManAppendAnd2( pNew, In[0], In[1] );
    And    = Gia_ManAppendAnd2( pNew, Abc_LitNot(In[0]), Abc_LitNot(In[1]) );
    Or     = Gia_ManAppendOr2( pNew, Out[1], And );
    Out[0] = Abc_LitNot( Or );
}
func Acec_InsertFadd(pNew *Gia_Man_t, In []int32, Out []int32) {
    var In2, Out1, Out2 [2]int32;
    Acec_InsertHadd( pNew, In, Out1 );
    In2[0] = Out1[0];
    In2[1] = In[2];
    Acec_InsertHadd( pNew, In2, Out2 );
    Out[0] = Out2[0];
    Out[1] = Gia_ManAppendOr2( pNew, Out1[1], Out2[1] );
}
func Acec_InsertTree(pNew *Gia_Man_t, vLeafMap *Vec_Wec_t) *Vec_Int_t {
    var vRootRanks *Vec_Int_t  = Vec_IntAlloc( Vec_WecSize(vLeafMap) + 5 );
    var vLevel *Vec_Int_t;
    var i int32;
    var In [3]int32;
    var Out [2]int32;
    // inlining of macro Vec_WecForEachLevel
    for  i = 0; (i < Vec_WecSize(vLeafMap)) && (((vLevel) = Vec_WecEntry(vLeafMap, i)), 1); i++  {
        if Vec_IntSize(vLevel) == 0 {
            Vec_IntPush( vRootRanks, 0 );
            continue;
        }
        for  Vec_IntSize(vLevel) > 1  {
            if Vec_IntSize(vLevel) == 2 {
                Vec_IntPush( vLevel, 0 ); }
            //In[2] = Vec_IntPop( vLevel );
            //In[1] = Vec_IntPop( vLevel );
            //In[0] = Vec_IntPop( vLevel );

            In[0] = Vec_IntEntry( vLevel, 0 );
            Vec_IntDrop( vLevel, 0 );

            In[1] = Vec_IntEntry( vLevel, 0 );
            Vec_IntDrop( vLevel, 0 );

            In[2] = Vec_IntEntry( vLevel, 0 );
            Vec_IntDrop( vLevel, 0 );

            Acec_InsertFadd( pNew, In, Out );
            Vec_IntPush( vLevel, Out[0] );
            if i+1 < Vec_WecSize(vLeafMap) {
                vLevel = Vec_WecEntry(vLeafMap, i+1);
            } else {
                vLevel = Vec_WecPushLevel(vLeafMap); }
            Vec_IntPush( vLevel, Out[1] );
            vLevel = Vec_WecEntry(vLeafMap, i);
        }
        assert(Vec_IntSize(vLevel) == 1, "Vec_IntSize(vLevel) == 1");
        Vec_IntPush( vRootRanks, Vec_IntEntry(vLevel, 0) );
    }
    return vRootRanks;
}

/**Function*************************************************************

  Synopsis    []

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Acec_InsertBox_rec(pNew *Gia_Man_t, p *Gia_Man_t, pObj *Gia_Obj_t) int32 {
    if ~pObj.Value {
        return pObj.Value; }
    assert(Gia_ObjIsAnd(pObj), "Gia_ObjIsAnd(pObj)");
    Acec_InsertBox_rec( pNew, p, Gia_ObjFanin0(pObj) );
    Acec_InsertBox_rec( pNew, p, Gia_ObjFanin1(pObj) );
    return (pObj.Value = Gia_ManAppendAnd2( pNew, Gia_ObjFanin0Copy(pObj), Gia_ObjFanin1Copy(pObj) ));
}
func Acec_BuildTree(pNew *Gia_Man_t, p *Gia_Man_t, vLeafLits *Vec_Wec_t, vRootLits *Vec_Int_t) *Vec_Int_t {
    var vLeafMap *Vec_Wec_t  = Vec_WecStart( Vec_WecSize(vLeafLits) );
    var vLevel, vRootRanks *Vec_Int_t;  
    var i, k, iLit, iLitNew int32;
    // add roo literals
    if vRootLits != nil {
        // inlining of macro Vec_IntForEachEntry
        for  i = 0; (i < Vec_IntSize(vRootLits)) && (((iLit) = Vec_IntEntry(vRootLits, i)), 1); i++  {
            if i < Vec_WecSize(vLeafMap) {
                vLevel = Vec_WecEntry(vLeafMap, i);
            } else {
                vLevel = Vec_WecPushLevel(vLeafMap); }
            Vec_IntPush( vLevel, iLit );
        } }
    // add other literals
    // inlining of macro Vec_WecForEachLevel
    for  i = 0; (i < Vec_WecSize(vLeafLits)) && (((vLevel) = Vec_WecEntry(vLeafLits, i)), 1); i++  {
        // inlining of macro Vec_IntForEachEntry
        for  k = 0; (k < Vec_IntSize(vLevel)) && (((iLit) = Vec_IntEntry(vLevel, k)), 1); k++  {
            var pObj *Gia_Obj_t  = Gia_ManObj( p, Abc_Lit2Var(iLit) );
            iLitNew = Acec_InsertBox_rec( pNew, p, pObj );
            iLitNew = Abc_LitNotCond( iLitNew, Abc_LitIsCompl(iLit) );
            Vec_WecPush( vLeafMap, i, iLitNew );
        }
    }
    // construct map of root literals
    vRootRanks = Acec_InsertTree( pNew, vLeafMap );
    Vec_WecFree( vLeafMap );
    return vRootRanks;
}
func Acec_InsertBox(pBox *Acec_Box_t, fAll int32) *Gia_Man_t {
    var p *Gia_Man_t  = pBox.pGia;
    var pNew *Gia_Man_t;
    var pObj *Gia_Obj_t;
    var vRootRanks, vLevel, vTemp *Vec_Int_t;
    var i, k, iLit, iLitNew int32;
    pNew = Gia_ManStart( Gia_ManObjNum(p) );
    pNew.pName = Abc_UtilStrsav( p.pName );
    pNew.pSpec = Abc_UtilStrsav( p.pSpec );
    Gia_ManFillValue(p);
    Gia_ManConst0(p).Value = 0;
    // inlining of macro Gia_ManForEachCi
    for  i = 0; (i < Vec_IntSize(p.vCis)) && ((pObj) = Gia_ManCi(p, i)); i++  {
        pObj.Value = Gia_ManAppendCi( pNew );
    }
    // implement tree
    if fAll != 0 {
        vRootRanks = Acec_BuildTree( pNew, p, pBox.vLeafLits, nil );
    } else {
        assert(pBox.vShared != nil, "pBox.vShared != nil");
        assert(pBox.vUnique != nil, "pBox.vUnique != nil");
        vRootRanks = Acec_BuildTree( pNew, p, pBox.vShared, nil );
        vRootRanks = Acec_BuildTree( pNew, p, pBox.vUnique, vTemp = vRootRanks );
        Vec_IntFree( vTemp );
    }
    // update polarity of literals
    // inlining of macro Vec_WecForEachLevel
    for  i = 0; (i < Vec_WecSize(pBox.vRootLits)) && (((vLevel) = Vec_WecEntry(pBox.vRootLits, i)), 1); i++  {
        // inlining of macro Vec_IntForEachEntry
        for  k = 0; (k < Vec_IntSize(vLevel)) && (((iLit) = Vec_IntEntry(vLevel, k)), 1); k++  {
            pObj = Gia_ManObj( p, Abc_Lit2Var(iLit) );
            if k {
                iLitNew = 0;
            } else {
                iLitNew = Vec_IntEntry( vRootRanks, i );
            }
            pObj.Value = Abc_LitNotCond( iLitNew, Abc_LitIsCompl(iLit) );
        }
    }
    Vec_IntFree( vRootRanks );
    // construct the outputs
    // inlining of macro Gia_ManForEachCo
    for  i = 0; (i < Vec_IntSize(p.vCos)) && ((pObj) = Gia_ManCo(p, i)); i++  {
        Acec_InsertBox_rec( pNew, p, Gia_ObjFanin0(pObj) );
    }
    // inlining of macro Gia_ManForEachCo
    for  i = 0; (i < Vec_IntSize(p.vCos)) && ((pObj) = Gia_ManCo(p, i)); i++  {
        pObj.Value = Gia_ManAppendCo( pNew, Gia_ObjFanin0Copy(pObj) );
    }
    Gia_ManSetRegNum( pNew, Gia_ManRegNum(p) );
    return pNew;
}

/**Function*************************************************************

  Synopsis    []

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Acec_Normalize(pGia *Gia_Man_t, fBooth int32, fVerbose int32) *Gia_Man_t {
    var vIgnore *Vec_Bit_t;
    if fBooth != 0 {
        vIgnore = Acec_BoothFindPPG( pGia );
    } else {
        vIgnore = nil;
    }
    var pBox *Acec_Box_t  = Acec_DeriveBox( pGia, vIgnore, 0, 0, fVerbose );
    var pNew *Gia_Man_t   = Acec_InsertBox( pBox, 1 );
    Acec_BoxFreeP( &pBox );
    Vec_BitFreeP( &vIgnore );
    return pNew;
}

////////////////////////////////////////////////////////////////////////
///                       END OF FILE                                ///
////////////////////////////////////////////////////////////////////////


//ABC_NAMESPACE_IMPL_END

