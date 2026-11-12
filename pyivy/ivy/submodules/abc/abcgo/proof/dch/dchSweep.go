package dch

import "os"

/**CFile****************************************************************

  FileName    [dchSweep.c]

  SystemName  [ABC: Logic synthesis and verification system.]

  PackageName [Choice computation for tech-mapping.]

  Synopsis    [One round of SAT sweeping.]

  Author      [Alan Mishchenko]
  
  Affiliation [UC Berkeley]

  Date        [Ver. 1.0. Started - June 29, 2008.]

  Revision    [$Id: dchSweep.c,v 1.00 2008/07/29 00:00:00 alanmi Exp $]

***********************************************************************/

//#include "dchInt.h"
//#include "misc/bar/bar.h"

//ABC_NAMESPACE_IMPL_START


////////////////////////////////////////////////////////////////////////
///                        DECLARATIONS                              ///
////////////////////////////////////////////////////////////////////////

func Dch_ObjChild0Fra(pObj *Aig_Obj_t) *Aig_Obj_t { assert(!Aig_IsComplement(pObj), "!Aig_IsComplement(pObj)"); if Aig_ObjFanin0(pObj) {
    return Aig_NotCond(Dch_ObjFraig(Aig_ObjFanin0(pObj)), Aig_ObjFaninC0(pObj));
} else {
    return nil;
}  }
func Dch_ObjChild1Fra(pObj *Aig_Obj_t) *Aig_Obj_t { assert(!Aig_IsComplement(pObj), "!Aig_IsComplement(pObj)"); if Aig_ObjFanin1(pObj) {
    return Aig_NotCond(Dch_ObjFraig(Aig_ObjFanin1(pObj)), Aig_ObjFaninC1(pObj));
} else {
    return nil;
}  }

////////////////////////////////////////////////////////////////////////
///                     FUNCTION DEFINITIONS                         ///
////////////////////////////////////////////////////////////////////////

/**Function*************************************************************

  Synopsis    [Performs fraiging for one node.]

  Description [Returns the fraiged node.]
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Dch_ManSweepNode(p *Dch_Man_t, pObj *Aig_Obj_t) { 
    var pObjRepr, pObjFraig, pObjFraig2, pObjReprFraig *Aig_Obj_t;
    var RetValue int32;
    // get representative of this class
    pObjRepr = Aig_ObjRepr( p.pAigTotal, pObj );
    if pObjRepr == nil {
        return; }
    // get the fraiged node
    pObjFraig = Dch_ObjFraig( pObj );
    if pObjFraig == nil {
        return; }
    // get the fraiged representative
    pObjReprFraig = Dch_ObjFraig( pObjRepr );
    if pObjReprFraig == nil {
        return; }
    // if the fraiged nodes are the same, return
    if Aig_Regular(pObjFraig) == Aig_Regular(pObjReprFraig) {
        // remember the proved equivalence
        p.pReprsProved[ pObj.Id ] = pObjRepr;
        return;
    }
    assert(Aig_Regular(pObjFraig) != Aig_ManConst1(p.pAigFraig), "Aig_Regular(pObjFraig) != Aig_ManConst1(p.pAigFraig)");
    RetValue = Dch_NodesAreEquiv( p, Aig_Regular(pObjReprFraig), Aig_Regular(pObjFraig) );
    if RetValue == -1 { // timed out
    
        Dch_ObjSetFraig( pObj, nil );
        return;
    }
    if RetValue == 1 {  // proved equivalent
    
        pObjFraig2 = Aig_NotCond( pObjReprFraig, pObj.fPhase ^ pObjRepr.fPhase );
        Dch_ObjSetFraig( pObj, pObjFraig2 );
        // remember the proved equivalence
        p.pReprsProved[ pObj.Id ] = pObjRepr;
        return;
    }
    // disproved the equivalence
    if p.pPars.fSimulateTfo {
        Dch_ManResimulateCex( p, pObj, pObjRepr );
    } else {
        Dch_ManResimulateCex2( p, pObj, pObjRepr ); }
    assert(Aig_ObjRepr( p.pAigTotal, pObj ) != pObjRepr, "Aig_ObjRepr( p.pAigTotal, pObj ) != pObjRepr");
}

/**Function*************************************************************

  Synopsis    [Performs fraiging for the internal nodes.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Dch_ManSweep(p *Dch_Man_t) {
    var pProgress *Bar_Progress_t  = nil;
    var pObj, pObjNew *Aig_Obj_t;
    var i int32;
    // map constants and PIs
    p.pAigFraig = Aig_ManStart( Aig_ManObjNumMax(p.pAigTotal) );
    Aig_ManCleanData( p.pAigTotal );
    Aig_ManConst1(p.pAigTotal).pData = Aig_ManConst1(p.pAigFraig);
    // inlining of macro Aig_ManForEachCi
    // inlining of macro Vec_PtrForEachEntry
    for  i = 0; (i < Vec_PtrSize(p.pAigTotal.vCis)) && (((pObj) = (Aig_Obj_t *)Vec_PtrEntry(p.pAigTotal.vCis, i)), 1); i++  {
        pObj.pData = Aig_ObjCreateCi( p.pAigFraig );
    }
    // sweep internal nodes
    pProgress = Bar_ProgressStart( os.Stdout, Aig_ManObjNumMax(p.pAigTotal) );
    // inlining of macro Aig_ManForEachNode
    // inlining of macro Vec_PtrForEachEntry
    for  i = 0; (i < Vec_PtrSize(p.pAigTotal.vObjs)) && (((pObj) = (Aig_Obj_t *)Vec_PtrEntry(p.pAigTotal.vObjs, i)), 1); i++  { if (pObj) == nil || !Aig_ObjIsNode(pObj) {} else {
        Bar_ProgressUpdate( pProgress, i, nil );
        if Dch_ObjFraig(Aig_ObjFanin0(pObj)) == nil || 
             Dch_ObjFraig(Aig_ObjFanin1(pObj)) == nil {
            continue; }
        pObjNew = Aig_And( p.pAigFraig, Dch_ObjChild0Fra(pObj), Dch_ObjChild1Fra(pObj) );
        if pObjNew == nil {
            continue; }
        Dch_ObjSetFraig( pObj, pObjNew );
        Dch_ManSweepNode( p, pObj );
    }
    }
    Bar_ProgressStop( pProgress );
    // update the representatives of the nodes (makes classes invalid)
    p.pAigTotal.pReprs = nil;
    p.pAigTotal.pReprs = p.pReprsProved;
    p.pReprsProved = nil;
    // clean the mark
    Aig_ManCleanMarkB( p.pAigTotal );
}

////////////////////////////////////////////////////////////////////////
///                       END OF FILE                                ///
////////////////////////////////////////////////////////////////////////


//ABC_NAMESPACE_IMPL_END

