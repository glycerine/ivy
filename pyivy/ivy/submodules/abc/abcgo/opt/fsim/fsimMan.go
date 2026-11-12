package fsim

/**CFile****************************************************************

  FileName    [fsimMan.c]

  SystemName  [ABC: Logic synthesis and verification system.]

  PackageName [Fast sequential AIG simulator.]

  Synopsis    [Simulation manager.]

  Author      [Alan Mishchenko]
  
  Affiliation [UC Berkeley]

  Date        [Ver. 1.0. Started - June 20, 2005.]

  Revision    [$Id: fsimMan.c,v 1.00 2005/06/20 00:00:00 alanmi Exp $]

***********************************************************************/

//#include "fsimInt.h"

//ABC_NAMESPACE_IMPL_START


////////////////////////////////////////////////////////////////////////
///                        DECLARATIONS                              ///
////////////////////////////////////////////////////////////////////////

////////////////////////////////////////////////////////////////////////
///                     FUNCTION DEFINITIONS                         ///
////////////////////////////////////////////////////////////////////////

/**Function*************************************************************

  Synopsis    [Creates fast simulation manager.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Fsim_ManCreate_rec(p *Fsim_Man_t, pObj *Aig_Obj_t) int32 {
    var iFan0, iFan1, iTemp int32;
    assert(!Aig_IsComplement(pObj), "!Aig_IsComplement(pObj)");
    if pObj.iData {
        return pObj.iData; }
    assert(!Aig_ObjIsConst1(pObj), "!Aig_ObjIsConst1(pObj)");
    if Aig_ObjIsNode(pObj) {
        iFan0 = Fsim_ManCreate_rec( p, Aig_ObjFanin0(pObj) );
        iFan1 = Fsim_ManCreate_rec( p, Aig_ObjFanin1(pObj) );
        assert(iFan0 != iFan1, "iFan0 != iFan1");
        if --p.pRefs[iFan0] == 0 {
            p.nCrossCut--; }
        iFan0 = Fsim_Var2Lit( iFan0, Aig_ObjFaninC0(pObj) );
        if --p.pRefs[iFan1] == 0 {
            p.nCrossCut--; }
        iFan1 = Fsim_Var2Lit( iFan1, Aig_ObjFaninC1(pObj) );
        if p.pAig.pEquivs {
            Fsim_ManCreate_rec( p, Aig_ObjEquiv(p.pAig, pObj) ); }
    
    } else if Aig_ObjIsPo(pObj) {
        assert(Aig_ObjRefs(pObj) == 0, "Aig_ObjRefs(pObj) == 0");
        iFan0 = Fsim_ManCreate_rec( p, Aig_ObjFanin0(pObj) );
        if --p.pRefs[iFan0] == 0 {
            p.nCrossCut--; }
        iFan0 = Fsim_Var2Lit( iFan0, Aig_ObjFaninC0(pObj) );
        iFan1 = 0;
    
    } else {
        iFan0 = iFan1 = 0;
        Vec_IntPush( p.vCis2Ids, Aig_ObjPioNum(pObj) );
    }
    if iFan0 < iFan1 {
        iTemp = iFan0;
iFan0 = iFan1;
iFan1 = iTemp; }
    p.pFans0[p.nObjs] = iFan0;
    p.pFans1[p.nObjs] = iFan1;
    p.pRefs[p.nObjs]  = Aig_ObjRefs(pObj);
    if p.pRefs[p.nObjs] {
        if p.nCrossCutMax < ++p.nCrossCut { 
            p.nCrossCutMax = p.nCrossCut; } }
    return pObj.iData = p.nObjs++;
}

/**Function*************************************************************

  Synopsis    [Creates fast simulation manager.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Fsim_ManCreate(pAig *Aig_Man_t) *Fsim_Man_t {
    var p *Fsim_Man_t;
    var pObj *Aig_Obj_t;
    var i, nObjs int32;
    Aig_ManCleanData( pAig );
    // inlining of macro ABC_ALLOC
    p = (Fsim_Man_t *)((Fsim_Man_t *) malloc(sizeof(Fsim_Man_t) * (size_t)(1)));
    memset( p, 0, sizeof(Fsim_Man_t) );
    p.pAig = pAig;
    p.nPis = Saig_ManPiNum(pAig);
    p.nPos = Saig_ManPoNum(pAig);
    p.nCis = Aig_ManPiNum(pAig);
    p.nCos = Aig_ManPoNum(pAig);
    p.nNodes = Aig_ManNodeNum(pAig);
    nObjs = p.nCis + p.nCos + p.nNodes + 2;
    // inlining of macro ABC_ALLOC
    p.pFans0 = ((int *) malloc(sizeof(int) * (size_t)(nObjs)));
    // inlining of macro ABC_ALLOC
    p.pFans1 = ((int *) malloc(sizeof(int) * (size_t)(nObjs)));
    // inlining of macro ABC_ALLOC
    p.pRefs  = ((int *) malloc(sizeof(int) * (size_t)(nObjs)));
    p.vCis2Ids = Vec_IntAlloc( Aig_ManPiNum(pAig) );
    // add objects (0=unused; 1=const1)
    p.pFans0[0] = p.pFans1[0] = 0;
    p.pFans0[1] = p.pFans1[1] = 0;
    p.pRefs[0] = 0;
    p.nObjs = 2;
    pObj = Aig_ManConst1( pAig );
    pObj.iData = 1;
    p.pRefs[1] = Aig_ObjRefs(pObj);
    if p.pRefs[1] {
        p.nCrossCut = 1; }
    Aig_ManForEachPi( pAig, pObj, i )
        if ( Aig_ObjRefs(pObj) == 0 )
            Fsim_ManCreate_rec( p, pObj );
    Aig_ManForEachPo( pAig, pObj, i )
        Fsim_ManCreate_rec( p, pObj );
    assert(Vec_IntSize(p.vCis2Ids) == Aig_ManPiNum(pAig), "Vec_IntSize(p.vCis2Ids) == Aig_ManPiNum(pAig)");
    assert(p.nObjs == nObjs, "p.nObjs == nObjs");
    // check references
    assert(p.nCrossCut == 0, "p.nCrossCut == 0");
    // inlining of macro Aig_ManForEachObj
    // inlining of macro Vec_PtrForEachEntry
    for  i = 0; (i < Vec_PtrSize(pAig.vObjs)) && (((pObj) = (Aig_Obj_t *)Vec_PtrEntry(pAig.vObjs, i)), 1); i++  { if (pObj) == nil {} else {
        assert(p.pRefs[pObj.iData] == 0, "p.pRefs[pObj.iData] == 0");
        p.pRefs[pObj.iData] = Aig_ObjRefs(pObj);
    }
    }
    // collect flop outputs
    p.vLos = Vec_IntAlloc( Aig_ManRegNum(pAig) );
    // inlining of macro Saig_ManForEachLo
    for  i = 0; (i < Saig_ManRegNum(pAig)) && (((pObj) = (Aig_Obj_t *)Vec_PtrEntry(pAig.vCis, i+Saig_ManPiNum(pAig))), 1); i++  {
        Vec_IntPush( p.vLos, pObj.iData );
    }
    // collect flop inputs
    p.vLis = Vec_IntAlloc( Aig_ManRegNum(pAig) );
    // inlining of macro Saig_ManForEachLi
    for  i = 0; (i < Saig_ManRegNum(pAig)) && (((pObj) = (Aig_Obj_t *)Vec_PtrEntry(pAig.vCos, i+Saig_ManPoNum(pAig))), 1); i++  {
        Vec_IntPush( p.vLis, pObj.iData );
    }
    // determine the frontier size
    p.nFront = 1 + int32((1.1 * p.nCrossCutMax)); 
    return p;
}

/**Function*************************************************************

  Synopsis    [Deletes fast simulation manager.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Fsim_ManDelete(p *Fsim_Man_t) {
    Vec_IntFree( p.vCis2Ids );
    Vec_IntFree( p.vLos );
    Vec_IntFree( p.vLis );
    // inlining of macro ABC_FREE
    ((p.pDataAig2) ? (free((char *) (p.pDataAig2)), (p.pDataAig2) = 0) : 0);
    // inlining of macro ABC_FREE
    ((p.pDataAig) ? (free((char *) (p.pDataAig)), (p.pDataAig) = 0) : 0);
    // inlining of macro ABC_FREE
    ((p.pFans0) ? (free((char *) (p.pFans0)), (p.pFans0) = 0) : 0);
    // inlining of macro ABC_FREE
    ((p.pFans1) ? (free((char *) (p.pFans1)), (p.pFans1) = 0) : 0);
    // inlining of macro ABC_FREE
    ((p.pRefs) ? (free((char *) (p.pRefs)), (p.pRefs) = 0) : 0);
    // inlining of macro ABC_FREE
    ((p.pDataSim) ? (free((char *) (p.pDataSim)), (p.pDataSim) = 0) : 0);
    // inlining of macro ABC_FREE
    ((p.pDataSimCis) ? (free((char *) (p.pDataSimCis)), (p.pDataSimCis) = 0) : 0);
    // inlining of macro ABC_FREE
    ((p.pDataSimCos) ? (free((char *) (p.pDataSimCos)), (p.pDataSimCos) = 0) : 0);
    // inlining of macro ABC_FREE
    ((p.pData1) ? (free((char *) (p.pData1)), (p.pData1) = 0) : 0);
    // inlining of macro ABC_FREE
    ((p.pData2) ? (free((char *) (p.pData2)), (p.pData2) = 0) : 0);
    // inlining of macro ABC_FREE
    ((p) ? (free((char *) (p)), (p) = 0) : 0);
}

/**Function*************************************************************

  Synopsis    [Testing procedure.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Fsim_ManTest(pAig *Aig_Man_t) {
    var p *Fsim_Man_t;
    p = Fsim_ManCreate( pAig );
    Fsim_ManFront( p, 0 );
    Fsim_ManDelete( p );
}

////////////////////////////////////////////////////////////////////////
///                       END OF FILE                                ///
////////////////////////////////////////////////////////////////////////


//ABC_NAMESPACE_IMPL_END

