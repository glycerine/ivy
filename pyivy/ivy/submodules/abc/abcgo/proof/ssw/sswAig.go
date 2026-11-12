package ssw

/**CFile****************************************************************

  FileName    [sswAig.c]

  SystemName  [ABC: Logic synthesis and verification system.]

  PackageName [Inductive prover with constraints.]

  Synopsis    [AIG manipulation.]

  Author      [Alan Mishchenko]
  
  Affiliation [UC Berkeley]

  Date        [Ver. 1.0. Started - September 1, 2008.]

  Revision    [$Id: sswAig.c,v 1.00 2008/09/01 00:00:00 alanmi Exp $]

***********************************************************************/

//#include "sswInt.h"

//ABC_NAMESPACE_IMPL_START


////////////////////////////////////////////////////////////////////////
///                        DECLARATIONS                              ///
////////////////////////////////////////////////////////////////////////

////////////////////////////////////////////////////////////////////////
///                     FUNCTION DEFINITIONS                         ///
////////////////////////////////////////////////////////////////////////

/**Function*************************************************************

  Synopsis    [Starts the SAT manager.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Ssw_FrmStart(pAig *Aig_Man_t) *Ssw_Frm_t {
    var p *Ssw_Frm_t;
    // inlining of macro ABC_ALLOC
    p = ((Ssw_Frm_t *) malloc(sizeof(Ssw_Frm_t) * (size_t)(1)));
    memset( p, 0, sizeof(Ssw_Frm_t) );
    p.pAig          = pAig;
    p.nObjs         = Aig_ManObjNumMax( pAig );
    p.nFrames       = 0;
    p.pFrames       = nil;
    p.vAig2Frm      = Vec_PtrAlloc( 0 );
    Vec_PtrFill( p.vAig2Frm, 2 * p.nObjs, nil );
    return p;
}

/**Function*************************************************************

  Synopsis    [Starts the SAT manager.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Ssw_FrmStop(p *Ssw_Frm_t) {
    if p.pFrames {
        Aig_ManStop( p.pFrames ); }
    Vec_PtrFree( p.vAig2Frm );
    // inlining of macro ABC_FREE
    ((p) ? (free((char *) (p)), (p) = 0) : 0);
}

/**Function*************************************************************

  Synopsis    [Performs speculative reduction for one node.]

  Description []

  SideEffects []

  SeeAlso     []

***********************************************************************/
func Ssw_FramesConstrainNode(p *Ssw_Man_t, pFrames *Aig_Man_t, pAig *Aig_Man_t, pObj *Aig_Obj_t, iFrame int32, fTwoPos int32) {
    var pObjNew, pObjNew2, pObjRepr, pObjReprNew, pMiter *Aig_Obj_t;
    // skip nodes without representative
    pObjRepr = Aig_ObjRepr(pAig, pObj);
    if pObjRepr == nil {
        return; }
    p.nConstrTotal++;
    assert(pObjRepr.Id < pObj.Id, "pObjRepr.Id < pObj.Id");
    // get the new node
    pObjNew = Ssw_ObjFrame( p, pObj, iFrame );
    // get the new node of the representative
    pObjReprNew = Ssw_ObjFrame( p, pObjRepr, iFrame );
    // if this is the same node, no need to add constraints
    if pObj.fPhase == pObjRepr.fPhase {
        assert(pObjNew != Aig_Not(pObjReprNew), "pObjNew != Aig_Not(pObjReprNew)");
        if pObjNew == pObjReprNew {
            return; }
    
    } else {
        assert(pObjNew != pObjReprNew, "pObjNew != pObjReprNew");
        if pObjNew == Aig_Not(pObjReprNew) {
            return; }
    }
    p.nConstrReduced++;
    // these are different nodes - perform speculative reduction
    pObjNew2 = Aig_NotCond( pObjReprNew, pObj.fPhase ^ pObjRepr.fPhase );
    // set the new node
    Ssw_ObjSetFrame( p, pObj, iFrame, pObjNew2 );
    // add the constraint
    if fTwoPos != 0 {
        Aig_ObjCreateCo( pFrames, pObjNew2 );
        Aig_ObjCreateCo( pFrames, pObjNew );
    
    } else {
        pMiter = Aig_Exor( pFrames, pObjNew, pObjNew2 );
        Aig_ObjCreateCo( pFrames, Aig_NotCond(pMiter, Aig_ObjPhaseReal(pMiter)) );
    }
}

/**Function*************************************************************

  Synopsis    [Prepares the inductive case with speculative reduction.]

  Description []

  SideEffects []

  SeeAlso     []

***********************************************************************/
func Ssw_FramesWithClasses(p *Ssw_Man_t) *Aig_Man_t {
    var pFrames *Aig_Man_t;
    var pObj, pObjLi, pObjLo, pObjNew *Aig_Obj_t;
    var i, f, iLits int32;
    assert(p.pFrames == nil, "p.pFrames == nil");
    assert(Aig_ManRegNum(p.pAig) > 0, "Aig_ManRegNum(p.pAig) > 0");
    assert(Aig_ManRegNum(p.pAig) < Aig_ManCiNum(p.pAig), "Aig_ManRegNum(p.pAig) < Aig_ManCiNum(p.pAig)");
    p.nConstrTotal = p.nConstrReduced = 0;

    // start the fraig package
    pFrames = Aig_ManStart( Aig_ManObjNumMax(p.pAig) * p.nFrames );
    // create latches for the first frame
    // inlining of macro Saig_ManForEachLo
    for  i = 0; (i < Saig_ManRegNum(p.pAig)) && (((pObj) = (Aig_Obj_t *)Vec_PtrEntry(p.pAig.vCis, i+Saig_ManPiNum(p.pAig))), 1); i++  {
        Ssw_ObjSetFrame( p, pObj, 0, Aig_ObjCreateCi(pFrames) );
    }
    // add timeframes
    iLits = 0;
    for  f = 0; f < p.pPars.nFramesK; f++  {
        // map constants and PIs
        Ssw_ObjSetFrame( p, Aig_ManConst1(p.pAig), f, Aig_ManConst1(pFrames) );
        // inlining of macro Saig_ManForEachPi
        // inlining of macro Vec_PtrForEachEntryStop
        for  i = 0; (i < Saig_ManPiNum(p.pAig)) && (((pObj) = (Aig_Obj_t *)Vec_PtrEntry(p.pAig.vCis, i)), 1); i++  {
            pObjNew = Aig_ObjCreateCi(pFrames);
            pObjNew.fPhase = (p.vInits != nil) && Vec_IntEntry(p.vInits, iLits++);
            Ssw_ObjSetFrame( p, pObj, f, pObjNew );
        }
        // set the constraints on the latch outputs
        // inlining of macro Saig_ManForEachLo
        for  i = 0; (i < Saig_ManRegNum(p.pAig)) && (((pObj) = (Aig_Obj_t *)Vec_PtrEntry(p.pAig.vCis, i+Saig_ManPiNum(p.pAig))), 1); i++  {
            Ssw_FramesConstrainNode( p, pFrames, p.pAig, pObj, f, 1 );
        }
        // add internal nodes of this frame
        // inlining of macro Aig_ManForEachNode
        // inlining of macro Vec_PtrForEachEntry
        for  i = 0; (i < Vec_PtrSize(p.pAig.vObjs)) && (((pObj) = (Aig_Obj_t *)Vec_PtrEntry(p.pAig.vObjs, i)), 1); i++  { if (pObj) == nil || !Aig_ObjIsNode(pObj) {} else {
            pObjNew = Aig_And( pFrames, Ssw_ObjChild0Fra(p, pObj, f), Ssw_ObjChild1Fra(p, pObj, f) );
            Ssw_ObjSetFrame( p, pObj, f, pObjNew );
            Ssw_FramesConstrainNode( p, pFrames, p.pAig, pObj, f, 1 );
        }
        }
        // transfer to the primary outputs
        // inlining of macro Aig_ManForEachCo
        // inlining of macro Vec_PtrForEachEntry
        for  i = 0; (i < Vec_PtrSize(p.pAig.vCos)) && (((pObj) = (Aig_Obj_t *)Vec_PtrEntry(p.pAig.vCos, i)), 1); i++  {
            Ssw_ObjSetFrame( p, pObj, f, Ssw_ObjChild0Fra(p, pObj,f) );
        }
        // transfer latch input to the latch outputs
        // inlining of macro Saig_ManForEachLiLo
        for  i = 0; (i < Saig_ManRegNum(p.pAig)) && (((pObjLi) = Saig_ManLi(p.pAig, i)), 1)            && (((pObjLo)=Saig_ManLo(p.pAig, i)), 1); i++  {
            Ssw_ObjSetFrame( p, pObjLo, f+1, Ssw_ObjFrame(p, pObjLi,f) );
        }
    }
    assert(p.vInits == nil || Vec_IntSize(p.vInits) == iLits + Saig_ManPiNum(p.pAig), "p.vInits == nil || Vec_IntSize(p.vInits) == iLits + Saig_ManPiNum(p.pAig)");
    // add the POs for the latch outputs of the last frame
    // inlining of macro Saig_ManForEachLo
    for  i = 0; (i < Saig_ManRegNum(p.pAig)) && (((pObj) = (Aig_Obj_t *)Vec_PtrEntry(p.pAig.vCis, i+Saig_ManPiNum(p.pAig))), 1); i++  {
        Aig_ObjCreateCo( pFrames, Ssw_ObjFrame( p, pObj, p.pPars.nFramesK ) );
    }

    // remove dangling nodes
    Aig_ManCleanup( pFrames );
    // make sure the satisfying assignment is node assigned
    assert(pFrames.pData == nil, "pFrames.pData == nil");
//Aig_ManShow( pFrames, 0, NULL );
    return pFrames;
}

/**Function*************************************************************

  Synopsis    [Prepares the inductive case with speculative reduction.]

  Description []

  SideEffects []

  SeeAlso     []

***********************************************************************/
func Ssw_SpeculativeReduction(p *Ssw_Man_t) *Aig_Man_t {
    var pFrames *Aig_Man_t;
    var pObj, pObjNew *Aig_Obj_t;
    var i int32;
    assert(p.pFrames == nil, "p.pFrames == nil");
    assert(Aig_ManRegNum(p.pAig) > 0, "Aig_ManRegNum(p.pAig) > 0");
    assert(Aig_ManRegNum(p.pAig) < Aig_ManCiNum(p.pAig), "Aig_ManRegNum(p.pAig) < Aig_ManCiNum(p.pAig)");
    p.nConstrTotal = p.nConstrReduced = 0;

    // start the fraig package
    pFrames = Aig_ManStart( Aig_ManObjNumMax(p.pAig) * p.nFrames );
    pFrames.pName = Abc_UtilStrsav( p.pAig.pName );
    // map constants and PIs
    Ssw_ObjSetFrame( p, Aig_ManConst1(p.pAig), 0, Aig_ManConst1(pFrames) );
    // inlining of macro Saig_ManForEachPi
    // inlining of macro Vec_PtrForEachEntryStop
    for  i = 0; (i < Saig_ManPiNum(p.pAig)) && (((pObj) = (Aig_Obj_t *)Vec_PtrEntry(p.pAig.vCis, i)), 1); i++  {
        Ssw_ObjSetFrame( p, pObj, 0, Aig_ObjCreateCi(pFrames) );
    }
    // create latches for the first frame
    // inlining of macro Saig_ManForEachLo
    for  i = 0; (i < Saig_ManRegNum(p.pAig)) && (((pObj) = (Aig_Obj_t *)Vec_PtrEntry(p.pAig.vCis, i+Saig_ManPiNum(p.pAig))), 1); i++  {
        Ssw_ObjSetFrame( p, pObj, 0, Aig_ObjCreateCi(pFrames) );
    }
    // set the constraints on the latch outputs
    // inlining of macro Saig_ManForEachLo
    for  i = 0; (i < Saig_ManRegNum(p.pAig)) && (((pObj) = (Aig_Obj_t *)Vec_PtrEntry(p.pAig.vCis, i+Saig_ManPiNum(p.pAig))), 1); i++  {
        Ssw_FramesConstrainNode( p, pFrames, p.pAig, pObj, 0, 0 );
    }
    // add internal nodes of this frame
    // inlining of macro Aig_ManForEachNode
    // inlining of macro Vec_PtrForEachEntry
    for  i = 0; (i < Vec_PtrSize(p.pAig.vObjs)) && (((pObj) = (Aig_Obj_t *)Vec_PtrEntry(p.pAig.vObjs, i)), 1); i++  { if (pObj) == nil || !Aig_ObjIsNode(pObj) {} else {
        pObjNew = Aig_And( pFrames, Ssw_ObjChild0Fra(p, pObj, 0), Ssw_ObjChild1Fra(p, pObj, 0) );
        Ssw_ObjSetFrame( p, pObj, 0, pObjNew );
        Ssw_FramesConstrainNode( p, pFrames, p.pAig, pObj, 0, 0 );
    }
    }
    // add the POs for the latch outputs of the last frame
    // inlining of macro Saig_ManForEachLi
    for  i = 0; (i < Saig_ManRegNum(p.pAig)) && (((pObj) = (Aig_Obj_t *)Vec_PtrEntry(p.pAig.vCos, i+Saig_ManPoNum(p.pAig))), 1); i++  {
        Aig_ObjCreateCo( pFrames, Ssw_ObjChild0Fra(p, pObj,0) );
    }
    // remove dangling nodes
    Aig_ManCleanup( pFrames );
    Aig_ManSetRegNum( pFrames, Aig_ManRegNum(p.pAig) );
//    Abc_Print( 1, "SpecRed: Total constraints = %d. Reduced constraints = %d.\n", 
//        p->nConstrTotal, p->nConstrReduced );
    return pFrames;
}

////////////////////////////////////////////////////////////////////////
///                       END OF FILE                                ///
////////////////////////////////////////////////////////////////////////


//ABC_NAMESPACE_IMPL_END
