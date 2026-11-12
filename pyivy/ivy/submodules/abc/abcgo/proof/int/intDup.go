package int

/**CFile****************************************************************

  FileName    [intDup.c]

  SystemName  [ABC: Logic synthesis and verification system.]

  PackageName [Interpolation engine.]

  Synopsis    [Specialized AIG duplication procedures.]

  Author      [Alan Mishchenko]
  
  Affiliation [UC Berkeley]

  Date        [Ver. 1.0. Started - June 24, 2008.]

  Revision    [$Id: intDup.c,v 1.00 2005/06/20 00:00:00 alanmi Exp $]

***********************************************************************/

//#include "intInt.h"

//ABC_NAMESPACE_IMPL_START


////////////////////////////////////////////////////////////////////////
///                        DECLARATIONS                              ///
////////////////////////////////////////////////////////////////////////

////////////////////////////////////////////////////////////////////////
///                     FUNCTION DEFINITIONS                         ///
////////////////////////////////////////////////////////////////////////

/**Function*************************************************************

  Synopsis    [Create trivial AIG manager for the init state.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Inter_ManStartInitState(nRegs int32) *Aig_Man_t {
    var p *Aig_Man_t;
    var pRes *Aig_Obj_t;
    var ppInputs **Aig_Obj_t;
    var i int32;
    assert(nRegs > 0, "nRegs > 0");
    // inlining of macro ABC_ALLOC
    ppInputs = ((Aig_Obj_t * *) malloc(sizeof(Aig_Obj_t *) * (size_t)(nRegs)));
    p = Aig_ManStart( nRegs );
    for  i = 0; i < nRegs; i++  {
        ppInputs[i] = Aig_Not( Aig_ObjCreateCi(p) );
    }
    pRes = Aig_Multi( p, ppInputs, nRegs, AIG_OBJ_AND );
    Aig_ObjCreateCo( p, pRes );
    // inlining of macro ABC_FREE
    ((ppInputs) ? (free((char *) (ppInputs)), (ppInputs) = 0) : 0);
    return p;
}

/**Function*************************************************************

  Synopsis    [Duplicate the AIG w/o POs.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Inter_ManStartDuplicated(p *Aig_Man_t) *Aig_Man_t {
    var pNew *Aig_Man_t;
    var pObj *Aig_Obj_t;
    var i int32;
    assert(Aig_ManRegNum(p) > 0, "Aig_ManRegNum(p) > 0");
    // create the new manager
    pNew = Aig_ManStart( Aig_ManObjNumMax(p) );
    pNew.pName = Abc_UtilStrsav( p.pName );
    pNew.pSpec = Abc_UtilStrsav( p.pSpec );
    // create the PIs
    Aig_ManCleanData( p );
    Aig_ManConst1(p).pData = Aig_ManConst1(pNew);
    // inlining of macro Aig_ManForEachCi
    // inlining of macro Vec_PtrForEachEntry
    for  i = 0; (i < Vec_PtrSize(p.vCis)) && (((pObj) = (Aig_Obj_t *)Vec_PtrEntry(p.vCis, i)), 1); i++  {
        pObj.pData = Aig_ObjCreateCi( pNew );
    }
    // set registers
    pNew.nTruePis = p.nTruePis;
    pNew.nTruePos = Saig_ManConstrNum(p);
    pNew.nRegs    = p.nRegs;
    // duplicate internal nodes
    // inlining of macro Aig_ManForEachNode
    // inlining of macro Vec_PtrForEachEntry
    for  i = 0; (i < Vec_PtrSize(p.vObjs)) && (((pObj) = (Aig_Obj_t *)Vec_PtrEntry(p.vObjs, i)), 1); i++  { if (pObj) == nil || !Aig_ObjIsNode(pObj) {} else {
        pObj.pData = Aig_And( pNew, Aig_ObjChild0Copy(pObj), Aig_ObjChild1Copy(pObj) ); }
    }

    // create constraint outputs
    // inlining of macro Saig_ManForEachPo
    // inlining of macro Vec_PtrForEachEntryStop
    for  i = 0; (i < Saig_ManPoNum(p)) && (((pObj) = (Aig_Obj_t *)Vec_PtrEntry(p.vCos, i)), 1); i++  {
        if i < Saig_ManPoNum(p)-Saig_ManConstrNum(p) {
            continue; }
        Aig_ObjCreateCo( pNew, Aig_Not( Aig_ObjChild0Copy(pObj) ) );
    }

    // create register inputs with MUXes
    // inlining of macro Saig_ManForEachLi
    for  i = 0; (i < Saig_ManRegNum(p)) && (((pObj) = (Aig_Obj_t *)Vec_PtrEntry(p.vCos, i+Saig_ManPoNum(p))), 1); i++  {
        Aig_ObjCreateCo( pNew, Aig_ObjChild0Copy(pObj) );
    }
    Aig_ManCleanup( pNew );
    return pNew;
}

/**Function*************************************************************

  Synopsis    [Duplicate the AIG w/o POs and transforms to transit into init state.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Inter_ManStartOneOutput(p *Aig_Man_t, fAddFirstPo int32) *Aig_Man_t {
    var pNew *Aig_Man_t;
    var pObj, pObjLi, pObjLo *Aig_Obj_t;
    var pCtrl *Aig_Obj_t  = nil; // Suppress "might be used uninitialized"
    var i int32;
    assert(Aig_ManRegNum(p) > 0, "Aig_ManRegNum(p) > 0");
    // create the new manager
    pNew = Aig_ManStart( Aig_ManObjNumMax(p) );
    pNew.pName = Abc_UtilStrsav( p.pName );
    pNew.pSpec = Abc_UtilStrsav( p.pSpec );
    // create the PIs
    Aig_ManCleanData( p );
    Aig_ManConst1(p).pData = Aig_ManConst1(pNew);
    // inlining of macro Aig_ManForEachCi
    // inlining of macro Vec_PtrForEachEntry
    for  i = 0; (i < Vec_PtrSize(p.vCis)) && (((pObj) = (Aig_Obj_t *)Vec_PtrEntry(p.vCis, i)), 1); i++  {
        if i == Saig_ManPiNum(p) {
            pCtrl = Aig_ObjCreateCi( pNew ); }
        pObj.pData = Aig_ObjCreateCi( pNew );
    }
    // set registers
    if fAddFirstPo != 0 {
        pNew.nRegs = 0;
    } else {
        pNew.nRegs = p.nRegs;
    }
    if fAddFirstPo != 0 {
        pNew.nTruePis = Aig_ManCiNum(p) + 1;
    } else {
        pNew.nTruePis = p.nTruePis + 1;
    }
    pNew.nTruePos = fAddFirstPo + Saig_ManConstrNum(p);
    // duplicate internal nodes
    // inlining of macro Aig_ManForEachNode
    // inlining of macro Vec_PtrForEachEntry
    for  i = 0; (i < Vec_PtrSize(p.vObjs)) && (((pObj) = (Aig_Obj_t *)Vec_PtrEntry(p.vObjs, i)), 1); i++  { if (pObj) == nil || !Aig_ObjIsNode(pObj) {} else {
        pObj.pData = Aig_And( pNew, Aig_ObjChild0Copy(pObj), Aig_ObjChild1Copy(pObj) ); }
    }

    // create constraint outputs
    // inlining of macro Saig_ManForEachPo
    // inlining of macro Vec_PtrForEachEntryStop
    for  i = 0; (i < Saig_ManPoNum(p)) && (((pObj) = (Aig_Obj_t *)Vec_PtrEntry(p.vCos, i)), 1); i++  {
        if i < Saig_ManPoNum(p)-Saig_ManConstrNum(p) {
            continue; }
        Aig_ObjCreateCo( pNew, Aig_Not( Aig_ObjChild0Copy(pObj) ) );
    }

    // add the PO
    if fAddFirstPo != 0 {
        pObj = Aig_ManCo( p, 0 );
        Aig_ObjCreateCo( pNew, Aig_ObjChild0Copy(pObj) );
    
    } else {
        // create register inputs with MUXes
        // inlining of macro Saig_ManForEachLiLo
        for  i = 0; (i < Saig_ManRegNum(p)) && (((pObjLi) = Saig_ManLi(p, i)), 1)            && (((pObjLo)=Saig_ManLo(p, i)), 1); i++  {
            pObj = Aig_Mux( pNew, pCtrl, (Aig_Obj_t *)pObjLo.pData, Aig_ObjChild0Copy(pObjLi) );
    //        pObj = Aig_Mux( pNew, pCtrl, Aig_ManConst0(pNew), Aig_ObjChild0Copy(pObjLi) );
            Aig_ObjCreateCo( pNew, pObj );
        }
    }
    Aig_ManCleanup( pNew );
    return pNew;
}

////////////////////////////////////////////////////////////////////////
///                       END OF FILE                                ///
////////////////////////////////////////////////////////////////////////


//ABC_NAMESPACE_IMPL_END

