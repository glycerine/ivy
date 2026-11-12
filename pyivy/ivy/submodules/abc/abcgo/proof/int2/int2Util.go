package int2

/**CFile****************************************************************

  FileName    [int2Util.c]

  SystemName  [ABC: Logic synthesis and verification system.]

  PackageName [Interpolation engine.]

  Synopsis    [Various utilities.]

  Author      [Alan Mishchenko]
  
  Affiliation [UC Berkeley]

  Date        [Ver. 1.0. Started - Dec 1, 2013.]

  Revision    [$Id: int2Util.c,v 1.00 2013/12/01 00:00:00 alanmi Exp $]

***********************************************************************/

//#include "int2Int.h"

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
func Int2_ManComputeCoPres(vSop *Vec_Int_t, nRegs int32) *Vec_Int_t {
    var vCoPres, vMap *Vec_Int_t;
    vCoPres = Vec_IntAlloc( 100 );
    if vSop == nil {
        Vec_IntPush( vCoPres, 0 );
    } else {
        var i, k, Limit int32;
        vMap = Vec_IntStart( nRegs );
        // inlining of macro Vec_IntForEachEntryStart
        for  i = 1; (i < Vec_IntSize(vSop)) && (((Limit) = Vec_IntEntry(vSop, i)), 1); i++  {
            for  k = 0; k < Limit; k++  {
                i++;
                assert(Vec_IntEntry(vSop, i + k) < 2 * nRegs, "Vec_IntEntry(vSop, i + k) < 2 * nRegs");
                Vec_IntWriteEntry( vMap, Abc_Lit2Var(Vec_IntEntry(vSop, i + k)), 1 );
            }
        }
        // inlining of macro Vec_IntForEachEntry
        for  i = 0; (i < Vec_IntSize(vMap)) && (((Limit) = Vec_IntEntry(vMap, i)), 1); i++  {
            if Limit != 0 {
                Vec_IntPush( vCoPres, i+1 ); }
        }
        Vec_IntFree( vMap );
    }
    return vCoPres;
}

func Int2_ManCollectInternal_rec(p *Gia_Man_t, pObj *Gia_Obj_t, vNodes *Vec_Int_t) {
    if Gia_ObjIsTravIdCurrent(p, pObj) {
        return; }
    Gia_ObjSetTravIdCurrent(p, pObj);
    if Gia_ObjIsCi(pObj) {
        return; }
    assert(Gia_ObjIsAnd(pObj), "Gia_ObjIsAnd(pObj)");
    Int2_ManCollectInternal_rec( p, Gia_ObjFanin0(pObj), vNodes );
    Int2_ManCollectInternal_rec( p, Gia_ObjFanin1(pObj), vNodes );
    Vec_IntPush( vNodes, Gia_ObjId(p, pObj) );
}
func Int2_ManCollectInternal(p *Gia_Man_t, vCoPres *Vec_Int_t) *Vec_Int_t {
    var vNodes *Vec_Int_t;
    var pObj *Gia_Obj_t;
    var i, Entry int32;
    Gia_ManIncrementTravId( p );
    Gia_ObjSetTravIdCurrent(p, Gia_ManConst0(p));
    // inlining of macro Gia_ManForEachCi
    for  i = 0; (i < Vec_IntSize(p.vCis)) && ((pObj) = Gia_ManCi(p, i)); i++  {
        Gia_ObjSetTravIdCurrent(p, pObj);
    }
    vNodes = Vec_IntAlloc( 1000 );
    // inlining of macro Vec_IntForEachEntry
    for  i = 0; (i < Vec_IntSize(vCoPres)) && (((Entry) = Vec_IntEntry(vCoPres, i)), 1); i++  {
        Int2_ManCollectInternal_rec( p, Gia_ObjFanin0(Gia_ManCo(p, Entry)), vNodes );
    }
    return vNodes;
}
func Int2_ManProbToGia(p *Gia_Man_t, vSop *Vec_Int_t) *Gia_Man_t {
    var vCoPres, vNodes *Vec_Int_t;
    var pNew, pTemp *Gia_Man_t;
    var pObj *Gia_Obj_t;
    var i, k, Entry, Limit int32;
    var Lit, Cube, Sop int32;
    assert(Gia_ManPoNum(p) == 1, "Gia_ManPoNum(p) == 1");
    // collect COs and ANDs
    vCoPres = Int2_ManComputeCoPres( vSop, Gia_ManRegNum(p) );
    vNodes = Int2_ManCollectInternal( p, vCoPres );
    // create new manager
    pNew = Gia_ManStart( Gia_ManObjNum(p) );
    pNew.pName = Abc_UtilStrsav( p.pName );
    pNew.pSpec = Abc_UtilStrsav( p.pSpec );
    Gia_ManConst0(p).Value = 0;
    // inlining of macro Gia_ManForEachCi
    for  i = 0; (i < Vec_IntSize(p.vCis)) && ((pObj) = Gia_ManCi(p, i)); i++  {
        pObj.Value = Gia_ManAppendCi(pNew);
    }
    Gia_ManHashAlloc( pNew );
    // inlining of macro Gia_ManForEachObjVec
    for  i = 0; (i < Vec_IntSize(vNodes)) && ((pObj) = Gia_ManObj(p, Vec_IntEntry(vNodes,i))); i++  {
        pObj.Value = Gia_ManHashAnd( pNew, Gia_ObjFanin0Copy(pObj), Gia_ObjFanin1Copy(pObj) );
    }
    // inlining of macro Vec_IntForEachEntry
    for  i = 0; (i < Vec_IntSize(vCoPres)) && (((Entry) = Vec_IntEntry(vCoPres, i)), 1); i++  {
        pObj = Gia_ManCo(p, Entry);
        pObj.Value = Gia_ObjFanin0Copy( pObj );
    }
    // create additional cubes
    Sop = 0;
    // inlining of macro Vec_IntForEachEntryStart
    for  i = 1; (i < Vec_IntSize(vSop)) && (((Limit) = Vec_IntEntry(vSop, i)), 1); i++  {
        Cube = 1;
        for  k = 0; k < Limit; k++  {
            i++;
            Lit = Vec_IntEntry( vSop, i + k );
            pObj = Gia_ManRi( p, Abc_Lit2Var(Lit) );
            Cube = Gia_ManHashAnd( pNew, Cube, Abc_LitNotCond(pObj.Value, Abc_LitIsCompl(Lit)) );
        }
        Sop = Gia_ManHashOr( pNew, Sop, Cube );
    }
    Gia_ManAppendCo( pNew, Sop );
    Gia_ManHashStop( pNew );
    // cleanup
    pNew = Gia_ManCleanup( pTemp = pNew );
    Gia_ManStop( pTemp );
    return pNew;
}

////////////////////////////////////////////////////////////////////////
///                       END OF FILE                                ///
////////////////////////////////////////////////////////////////////////


//ABC_NAMESPACE_IMPL_END

