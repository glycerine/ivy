package cov

/**CFile****************************************************************

  FileName    [covMan.c]

  SystemName  [ABC: Logic synthesis and verification system.]

  PackageName [Mapping into network of SOPs/ESOPs.]

  Synopsis    [Decomposition manager.]

  Author      [Alan Mishchenko]
  
  Affiliation [UC Berkeley]

  Date        [Ver. 1.0. Started - June 20, 2005.]

  Revision    [$Id: covMan.c,v 1.00 2005/06/20 00:00:00 alanmi Exp $]

***********************************************************************/

//#include "cov.h"

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
func Cov_ManAlloc(pNtk *Abc_Ntk_t, nFaninMax int32, nCubesMax int32) *Cov_Man_t {
    var pMan *Cov_Man_t;
    var pMem *Cov_Obj_t;
    var pObj *Abc_Obj_t;
    var i int32;
    assert(pNtk.pManCut == nil, "pNtk.pManCut == nil");

    // start the manager
    // inlining of macro ABC_ALLOC
    pMan = ((Cov_Man_t *) malloc(sizeof(Cov_Man_t) * (size_t)(1)));
    memset( pMan, 0, sizeof(Cov_Man_t) );
    pMan.nFaninMax = nFaninMax;
    pMan.nCubesMax = nCubesMax;
    pMan.nWords    = Abc_BitWordNum( nFaninMax * 2 );

    // get the cubes
    pMan.vComTo0 = Vec_IntAlloc( 2*nFaninMax );
    pMan.vComTo1 = Vec_IntAlloc( 2*nFaninMax );
    pMan.vPairs0 = Vec_IntAlloc( nFaninMax );
    pMan.vPairs1 = Vec_IntAlloc( nFaninMax );
    pMan.vTriv0  = Vec_IntAlloc( 1 );  Vec_IntPush( pMan.vTriv0, -1 ); 
    pMan.vTriv1  = Vec_IntAlloc( 1 );  Vec_IntPush( pMan.vTriv1, -1 ); 

    // allocate memory for object structures
    // inlining of macro ABC_ALLOC
    pMan.pMemory = pMem = ((Cov_Obj_t *) malloc(sizeof(Cov_Obj_t) * (size_t)(sizeof(Cov_Obj_t) * Abc_NtkObjNumMax(pNtk))));
    memset( pMem, 0, sizeof(Cov_Obj_t) * Abc_NtkObjNumMax(pNtk) );
    // allocate storage for the pointers to the memory
    pMan.vObjStrs = Vec_PtrAlloc( Abc_NtkObjNumMax(pNtk) );
    Vec_PtrFill( pMan.vObjStrs, Abc_NtkObjNumMax(pNtk), nil );
    // inlining of macro Abc_NtkForEachObj
    for  i = 0; (i < Vec_PtrSize((pNtk).vObjs)) && (((pObj) = Abc_NtkObj(pNtk, i)), 1); i++  {            if (pObj) == nil {} else {
        Vec_PtrWriteEntry( pMan.vObjStrs, i, pMem + i ); }
    }
    // create the cube manager
    pMan.pManMin = Min_ManAlloc( nFaninMax );
    return pMan;
}

/**Function*************************************************************

  Synopsis    []

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Cov_ManFree(p *Cov_Man_t) {
    var vSupp *Vec_Int_t;
    var i int32;
    for  i = 0; i < p.vObjStrs.nSize; i++  {
        vSupp = ((Cov_Obj_t *)p.vObjStrs.pArray[i]).vSupp;
        if vSupp != nil { Vec_IntFree( vSupp ); }
    }

    Min_ManFree( p.pManMin );
    Vec_PtrFree( p.vObjStrs );
    Vec_IntFree( p.vFanCounts );
    Vec_IntFree( p.vTriv0 );
    Vec_IntFree( p.vTriv1 );
    Vec_IntFree( p.vComTo0 );
    Vec_IntFree( p.vComTo1 );
    Vec_IntFree( p.vPairs0 );
    Vec_IntFree( p.vPairs1 );
    // inlining of macro ABC_FREE
    ((p.pMemory) ? (free((char *) (p.pMemory)), (p.pMemory) = 0) : 0);
    // inlining of macro ABC_FREE
    ((p) ? (free((char *) (p)), (p) = 0) : 0);
}

/**Function*************************************************************

  Synopsis    [Drop the covers at the node.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Abc_NodeCovDropData(p *Cov_Man_t, pObj *Abc_Obj_t) {
    var nFanouts int32;
    assert(p.vFanCounts, "p.vFanCounts");
    nFanouts = Vec_IntEntry( p.vFanCounts, pObj.Id );
    assert(nFanouts > 0, "nFanouts > 0");
    if nFanouts--; nFanouts == 0 {
        Vec_IntFree( Abc_ObjGetSupp(pObj) );
        Abc_ObjSetSupp( pObj, nil );
        Min_CoverRecycle( p.pManMin, Abc_ObjGetCover2(pObj) );
        Abc_ObjSetCover2( pObj, nil );
        p.nSupps--;
    }
    Vec_IntWriteEntry( p.vFanCounts, pObj.Id, nFanouts );
}

////////////////////////////////////////////////////////////////////////
///                       END OF FILE                                ///
////////////////////////////////////////////////////////////////////////


//ABC_NAMESPACE_IMPL_END

