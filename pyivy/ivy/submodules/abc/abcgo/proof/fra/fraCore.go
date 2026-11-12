package fra

import "fmt"

/**CFile****************************************************************

  FileName    [fraCore.c]

  SystemName  [ABC: Logic synthesis and verification system.]

  PackageName [New FRAIG package.]

  Synopsis    []

  Author      [Alan Mishchenko]
  
  Affiliation [UC Berkeley]

  Date        [Ver. 1.0. Started - June 30, 2007.]

  Revision    [$Id: fraCore.c,v 1.00 2007/06/30 00:00:00 alanmi Exp $]

***********************************************************************/

//#include "fra.h"

//ABC_NAMESPACE_IMPL_START


////////////////////////////////////////////////////////////////////////
///                        DECLARATIONS                              ///
////////////////////////////////////////////////////////////////////////

/* 
    Speculating reduction in the sequential case leads to an interesting 
    situation when a counter-ex may not refine any classes. This happens
    for non-constant equivalence classes. In such cases the representative
    of the class (proved by simulation to be non-constant) may be reduced 
    to a constant during the speculative reduction. The fraig-representative 
    of this representative node is a constant node, even though this is a 
    non-constant class. Experiments have shown that this situation happens 
    very often at the beginning of the refinement iteration when there are 
    many spurious candidate equivalence classes (especially if heavy-duty 
    simulatation of BMC was node used at the beginning). As a result, the 
    SAT solver run may return a counter-ex that  distinguishes the given 
    representative node from the constant-1 node but this counter-ex
    does not distinguish the nodes in the non-costant class... This is why 
    there is no check of refinement after a counter-ex in the sequential case.
*/

////////////////////////////////////////////////////////////////////////
///                     FUNCTION DEFINITIONS                         ///
////////////////////////////////////////////////////////////////////////

/**Function*************************************************************

  Synopsis    [Reports the status of the miter.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Fra_FraigMiterStatus(p *Aig_Man_t) int32 {
    var pObj, pChild *Aig_Obj_t;
    var i, CountConst0, CountNonConst0, CountUndecided int32;
    if p.pData {
        return 0; }
    // inlining of macro Aig_ManForEachPoSeq
    // inlining of macro Vec_PtrForEachEntryStop
    for  i = 0; (i < Aig_ManCoNum(p)-Aig_ManRegNum(p)) && (((pObj) = (Aig_Obj_t *)Vec_PtrEntry(p.vCos, i)), 1); i++  {
        pChild = Aig_ObjChild0(pObj);
        // check if the output is constant 0
        if pChild == Aig_ManConst0(p) {
            CountConst0++;
            continue;
        }
        // check if the output is constant 1
        if pChild == Aig_ManConst1(p) {
            CountNonConst0++;
            continue;
        }
        // check if the output is a primary input
        if Aig_ObjIsCi(Aig_Regular(pChild)) && Aig_ObjCioId(Aig_Regular(pChild)) < p.nTruePis {
            CountNonConst0++;
            continue;
        }
        // check if the output can be not constant 0
        if Aig_Regular(pChild).fPhase != uint32(Aig_IsComplement(pChild)) {
            CountNonConst0++;
            continue;
        }
        CountUndecided++;
    }
/*
    if ( p->pParams->fVerbose )
    {
        printf( "Miter has %d outputs. ", Aig_ManCoNum(p->pManAig) );
        printf( "Const0 = %d.  ", CountConst0 );
        printf( "NonConst0 = %d.  ", CountNonConst0 );
        printf( "Undecided = %d.  ", CountUndecided );
        printf( "\n" );
    }
*/
    if CountNonConst0 {
        return 0; }
    if CountUndecided != 0 {
        return -1; }
    return 1;
}

/**Function*************************************************************

  Synopsis    [Reports the status of the miter.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Fra_FraigMiterAssertedOutput(p *Aig_Man_t) int32 {
    var pObj, pChild *Aig_Obj_t;
    var i int32;
    // inlining of macro Aig_ManForEachPoSeq
    // inlining of macro Vec_PtrForEachEntryStop
    for  i = 0; (i < Aig_ManCoNum(p)-Aig_ManRegNum(p)) && (((pObj) = (Aig_Obj_t *)Vec_PtrEntry(p.vCos, i)), 1); i++  {
        pChild = Aig_ObjChild0(pObj);
        // check if the output is constant 0
        if pChild == Aig_ManConst0(p) {
            continue; }
        // check if the output is constant 1
        if pChild == Aig_ManConst1(p) {
            return i; }
        // check if the output can be not constant 0
        if Aig_Regular(pChild).fPhase != uint32(Aig_IsComplement(pChild)) {
            return i; }
    }
    return -1;
}

/**Function*************************************************************

  Synopsis    [Write speculative miter for one node.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
var spStatic_Fra_FraigNodeSpeculate_Counter int32;
func Fra_FraigNodeSpeculate(p *Fra_Man_t, pObj *Aig_Obj_t, pObjFraig *Aig_Obj_t, pObjReprFraig *Aig_Obj_t) { 
    
    var FileName string // was [20]byte in C

    var pTemp *Aig_Man_t;
    var pNode *Aig_Obj_t;
    var i int32;
    // create manager with the logic for these two nodes
    pTemp = Aig_ManExtractMiter( p.pManFraig, pObjFraig, pObjReprFraig );
    // dump the logic into a file
    FileName = fmt.Sprintf("aig\\%03d.blif", ++spStatic_Fra_FraigNodeSpeculate_Counter);
    Aig_ManDumpBlif( pTemp, FileName, nil, nil );
    printf( "Speculation cone with %d nodes was written into file \"%s\".\n", Aig_ManNodeNum(pTemp), FileName );
    // clean up
    Aig_ManStop( pTemp );
    // inlining of macro Aig_ManForEachObj
    // inlining of macro Vec_PtrForEachEntry
    for  i = 0; (i < Vec_PtrSize(p.pManFraig.vObjs)) && (((pNode) = (Aig_Obj_t *)Vec_PtrEntry(p.pManFraig.vObjs, i)), 1); i++  { if (pNode) == nil {} else {
        pNode.pData = p; }
    }
}

/**Function*************************************************************

  Synopsis    [Verifies the generated counter-ex.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Fra_FraigVerifyCounterEx(p *Fra_Man_t, vCex *Vec_Int_t) {
    var pObj *Aig_Obj_t;
    var ppClass **Aig_Obj_t;
    var i, c int32;
    assert(Aig_ManCiNum(p.pManAig) == Vec_IntSize(vCex), "Aig_ManCiNum(p.pManAig) == Vec_IntSize(vCex)");
    // make sure the input pattern is not used
    // inlining of macro Aig_ManForEachObj
    // inlining of macro Vec_PtrForEachEntry
    for  i = 0; (i < Vec_PtrSize(p.pManAig.vObjs)) && (((pObj) = (Aig_Obj_t *)Vec_PtrEntry(p.pManAig.vObjs, i)), 1); i++  { if (pObj) == nil {} else {
        assert(!pObj.fMarkB, "!pObj.fMarkB"); }
    }
    // simulate the cex through the AIG
    Aig_ManConst1(p.pManAig).fMarkB = 1;
    // inlining of macro Aig_ManForEachCi
    // inlining of macro Vec_PtrForEachEntry
    for  i = 0; (i < Vec_PtrSize(p.pManAig.vCis)) && (((pObj) = (Aig_Obj_t *)Vec_PtrEntry(p.pManAig.vCis, i)), 1); i++  {
        pObj.fMarkB = Vec_IntEntry(vCex, i);
    }
    // inlining of macro Aig_ManForEachNode
    // inlining of macro Vec_PtrForEachEntry
    for  i = 0; (i < Vec_PtrSize(p.pManAig.vObjs)) && (((pObj) = (Aig_Obj_t *)Vec_PtrEntry(p.pManAig.vObjs, i)), 1); i++  { if (pObj) == nil || !Aig_ObjIsNode(pObj) {} else {
        pObj.fMarkB = (Aig_ObjFanin0(pObj).fMarkB ^ Aig_ObjFaninC0(pObj)) & 
                       (Aig_ObjFanin1(pObj).fMarkB ^ Aig_ObjFaninC1(pObj)); }
    }
    // inlining of macro Aig_ManForEachCo
    // inlining of macro Vec_PtrForEachEntry
    for  i = 0; (i < Vec_PtrSize(p.pManAig.vCos)) && (((pObj) = (Aig_Obj_t *)Vec_PtrEntry(p.pManAig.vCos, i)), 1); i++  {
        pObj.fMarkB = Aig_ObjFanin0(pObj).fMarkB ^ Aig_ObjFaninC0(pObj);
    }
    // check if the classes hold
    // inlining of macro Vec_PtrForEachEntry
    for  i = 0; (i < Vec_PtrSize(p.pCla.vClasses1)) && (((pObj) = (Aig_Obj_t *)Vec_PtrEntry(p.pCla.vClasses1, i)), 1); i++  {
        if pObj.fPhase != pObj.fMarkB {
            fmt.Printf("The node %d is not constant under cex!\n", pObj.Id); }
    }
    // inlining of macro Vec_PtrForEachEntry
    for  i = 0; (i < Vec_PtrSize(p.pCla.vClasses)) && (((ppClass) = (Aig_Obj_t **)Vec_PtrEntry(p.pCla.vClasses, i)), 1); i++  {
        for  c = 1; ppClass[c]; c++  {
            if (ppClass[0].fPhase ^ ppClass[c].fPhase) != (ppClass[0].fMarkB ^ ppClass[c].fMarkB) {
                fmt.Printf("The nodes %d and %d are not equal under cex!\n", ppClass[0].Id, ppClass[c].Id); }
        }
//        for ( c = 0; ppClass[c]; c++ )
//            if ( Fra_ObjFraig(ppClass[c],p->pPars->nFramesK) == Aig_ManConst1(p->pManFraig) )
//                printf( "A member of non-constant class has a constant repr!\n" );
    }
    // clean the simulation pattern
    // inlining of macro Aig_ManForEachObj
    // inlining of macro Vec_PtrForEachEntry
    for  i = 0; (i < Vec_PtrSize(p.pManAig.vObjs)) && (((pObj) = (Aig_Obj_t *)Vec_PtrEntry(p.pManAig.vObjs, i)), 1); i++  { if (pObj) == nil {} else {
        pObj.fMarkB = 0; }
    }
}

/**Function*************************************************************

  Synopsis    [Performs fraiging for one node.]

  Description [Returns the fraiged node.]
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Fra_FraigNode(p *Fra_Man_t, pObj *Aig_Obj_t) { 
    var pObjRepr, pObjFraig, pObjFraig2, pObjReprFraig *Aig_Obj_t;
    var RetValue int32;
    assert(!Aig_IsComplement(pObj), "!Aig_IsComplement(pObj)");
    // get representative of this class
    pObjRepr = Fra_ClassObjRepr( pObj );
    if pObjRepr == nil || // this is a unique node
       (!p.pPars.fDoSparse && pObjRepr == Aig_ManConst1(p.pManAig)) { // this is a sparse node
        return; }
    // get the fraiged node
    pObjFraig = Fra_ObjFraig( pObj, p.pPars.nFramesK );
    // get the fraiged representative
    pObjReprFraig = Fra_ObjFraig( pObjRepr, p.pPars.nFramesK );
    // if the fraiged nodes are the same, return
    if Aig_Regular(pObjFraig) == Aig_Regular(pObjReprFraig) {
        p.nSatCallsSkipped++;
        return;
    }
    assert(p.pPars.nFramesK || Aig_Regular(pObjFraig) != Aig_ManConst1(p.pManFraig), "p.pPars.nFramesK || Aig_Regular(pObjFraig) != Aig_ManConst1(p.pManFraig)");
    // if they are proved different, the c-ex will be in p->pPatWords
    RetValue = Fra_NodesAreEquiv( p, Aig_Regular(pObjReprFraig), Aig_Regular(pObjFraig) );
    if RetValue == 1 {  // proved equivalent
    
//        if ( p->pPars->fChoicing )
//            Aig_ObjCreateRepr( p->pManFraig, Aig_Regular(pObjReprFraig), Aig_Regular(pObjFraig) );
        // the nodes proved equal
        pObjFraig2 = Aig_NotCond( pObjReprFraig, pObj.fPhase ^ pObjRepr.fPhase );
        Fra_ObjSetFraig( pObj, p.pPars.nFramesK, pObjFraig2 );
        return;
    }
    if RetValue == -1 { // failed
    
        if p.vTimeouts == nil {
            p.vTimeouts = Vec_PtrAlloc( 100 ); }
        Vec_PtrPush( p.vTimeouts, pObj );
        if !p.pPars.fSpeculate {
            return; }
        assert(false, "0");
        // speculate
        p.nSpeculs++;
        pObjFraig2 = Aig_NotCond( pObjReprFraig, pObj.fPhase ^ pObjRepr.fPhase );
        Fra_ObjSetFraig( pObj, p.pPars.nFramesK, pObjFraig2 );
        Fra_FraigNodeSpeculate( p, pObj, Aig_Regular(pObjFraig), Aig_Regular(pObjReprFraig) );
        return;
    }
    // disprove the nodes
    p.pCla.fRefinement = 1;
    // if we do not include the node into those disproved, we may end up 
    // merging this node with another representative, for which proof has timed out
    if p.vTimeouts {
        Vec_PtrPush( p.vTimeouts, pObj ); }
    // verify that the counter-example satisfies all the constraints
//    if ( p->vCex )
//        Fra_FraigVerifyCounterEx( p, p->vCex );
    // simulate the counter-example and return the Fraig node
    Fra_SmlResimulate( p );
    if p.pManFraig.pData {
        return; }
    if !p.pPars.nFramesK && Fra_ClassObjRepr(pObj) == pObjRepr {
        fmt.Printf("Fra_FraigNode(): Error in class refinement!\n"); }
    assert(p.pPars.nFramesK || Fra_ClassObjRepr(pObj) != pObjRepr, "p.pPars.nFramesK || Fra_ClassObjRepr(pObj) != pObjRepr");
}

/**Function*************************************************************

  Synopsis    [Performs fraiging for the internal nodes.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Fra_FraigSweep(p *Fra_Man_t) {
//    Bar_Progress_t * pProgress = NULL;
    var pObj, pObjNew *Aig_Obj_t;
    var i, Pos int32;
    var nBTracksOld int32;
    // fraig latch outputs
    // inlining of macro Aig_ManForEachLoSeq
    // inlining of macro Vec_PtrForEachEntryStart
    for  i = Aig_ManCiNum(p.pManAig)-Aig_ManRegNum(p.pManAig); (i < Vec_PtrSize(p.pManAig.vCis)) && (((pObj) = (Aig_Obj_t *)Vec_PtrEntry(p.pManAig.vCis, i)), 1); i++  {
        Fra_FraigNode( p, pObj );
        if p.pPars.fUseImps {
            Pos = Fra_ImpCheckForNode( p, p.pCla.vImps, pObj, Pos ); }
    }
    if p.pPars.fLatchCorr {
        return; }
    // fraig internal nodes
//    if ( !p->pPars->fDontShowBar )
//        pProgress = Bar_ProgressStart( stdout, Aig_ManObjNumMax(p->pManAig) );
    nBTracksOld = p.pPars.nBTLimitNode;
    // inlining of macro Aig_ManForEachNode
    // inlining of macro Vec_PtrForEachEntry
    for  i = 0; (i < Vec_PtrSize(p.pManAig.vObjs)) && (((pObj) = (Aig_Obj_t *)Vec_PtrEntry(p.pManAig.vObjs, i)), 1); i++  { if (pObj) == nil || !Aig_ObjIsNode(pObj) {} else {
//        if ( pProgress )
//            Bar_ProgressUpdate( pProgress, i, NULL );
        // derive and remember the new fraig node
        pObjNew = Aig_And( p.pManFraig, Fra_ObjChild0Fra(pObj,p.pPars.nFramesK), Fra_ObjChild1Fra(pObj,p.pPars.nFramesK) );
        Fra_ObjSetFraig( pObj, p.pPars.nFramesK, pObjNew );
        Aig_Regular(pObjNew).pData = p;
        // quit if simulation detected a counter-example for a PO
        if p.pManFraig.pData {
            continue; }
//        if ( Aig_SupportSize(p->pManAig,pObj) > 16 )
//            continue;
        // perform fraiging
        if p.pPars.nLevelMax && int32(pObj.Level) > p.pPars.nLevelMax {
            p.pPars.nBTLimitNode = 5; }
        Fra_FraigNode( p, pObj );
        if p.pPars.nLevelMax && int32(pObj.Level) > p.pPars.nLevelMax {
            p.pPars.nBTLimitNode = nBTracksOld; }
        // check implications
        if p.pPars.fUseImps {
            Pos = Fra_ImpCheckForNode( p, p.pCla.vImps, pObj, Pos ); }
    }
    }
//    if ( pProgress )
//        Bar_ProgressStop( pProgress );
    // try to prove the outputs of the miter
    p.nNodesMiter = Aig_ManNodeNum(p.pManFraig);
//    Fra_MiterStatus( p->pManFraig );
//    if ( p->pPars->fProve && p->pManFraig->pData == NULL )
//        Fra_MiterProve( p );
    // compress implications after processing all of them
    if p.pPars.fUseImps {
        Fra_ImpCompactArray( p.pCla.vImps ); }
}

/**Function*************************************************************

  Synopsis    [Performs fraiging of the AIG.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Fra_FraigPerform(pManAig *Aig_Man_t, pPars *Fra_Par_t) *Aig_Man_t {
    var p *Fra_Man_t;
    var pManAigNew *Aig_Man_t;
    var clk abctime;
    if Aig_ManNodeNum(pManAig) == 0 {
        return Aig_ManDupOrdered(pManAig); }
clk = Abc_Clock();
    p = Fra_ManStart( pManAig, pPars );
    p.pManFraig = Fra_ManPrepareComb( p );
    p.pSml = Fra_SmlStart( pManAig, 0, 1, pPars.nSimWords );
    Fra_SmlSimulate( p, 0 );
//    if ( p->pPars->fChoicing )
//        Aig_ManReprStart( p->pManFraig, Aig_ManObjNumMax(p->pManAig) );
    // collect initial states
    p.nLitsBeg  = Fra_ClassesCountLits( p.pCla );
    p.nNodesBeg = Aig_ManNodeNum(pManAig);
    p.nRegsBeg  = Aig_ManRegNum(pManAig);
    // perform fraig sweep
if p.pPars.fVerbose {
Fra_ClassesPrint( p.pCla, 1 ); }
    Fra_FraigSweep( p );
    // call back the procedure to check implications
    if pManAig.pImpFunc {
        pManAig.pImpFunc( p, pManAig.pImpData ); }
    // no need to filter one-hot clauses because they satisfy base case by construction
    // finalize the fraiged manager
    Fra_ManFinalizeComb( p );
    if p.pPars.fChoicing { 
var clk2 abctime  = Abc_Clock();
        Fra_ClassesCopyReprs( p.pCla, p.vTimeouts );
        pManAigNew = Aig_ManDupRepr( p.pManAig, 1 );
        Aig_ManReprStart( pManAigNew, Aig_ManObjNumMax(pManAigNew) );
        Aig_ManTransferRepr( pManAigNew, p.pManAig );
        Aig_ManMarkValidChoices( pManAigNew );
        Aig_ManStop( p.pManFraig );
        p.pManFraig = nil;
p.timeTrav += Abc_Clock() - clk2;
    
    } else {
        Fra_ClassesCopyReprs( p.pCla, p.vTimeouts );
        Aig_ManCleanup( p.pManFraig );
        pManAigNew = p.pManFraig;
        p.pManFraig = nil;
    }
p.timeTotal = Abc_Clock() - clk;
    // collect final stats
    p.nLitsEnd  = Fra_ClassesCountLits( p.pCla );
    p.nNodesEnd = Aig_ManNodeNum(pManAigNew);
    p.nRegsEnd  = Aig_ManRegNum(pManAigNew);
    Fra_ManStop( p );
    return pManAigNew;
}

/**Function*************************************************************

  Synopsis    [Performs choicing of the AIG.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Fra_FraigChoice(pManAig *Aig_Man_t, nConfMax int32, nLevelMax int32) *Aig_Man_t {
    var Pars Fra_Par_t;
    var pPars *Fra_Par_t = &Pars; 
    Fra_ParamsDefault( pPars );
    pPars.nBTLimitNode = nConfMax;
    pPars.fChoicing    = 1;
    pPars.fDoSparse    = 1;
    pPars.fSpeculate   = 0;
    pPars.fProve       = 0;
    pPars.fVerbose     = 0;
    pPars.fDontShowBar = 1;
    pPars.nLevelMax    = nLevelMax;
    return Fra_FraigPerform( pManAig, pPars );
}
 
/**Function*************************************************************

  Synopsis    []

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Fra_FraigEquivence(pManAig *Aig_Man_t, nConfMax int32, fProve int32) *Aig_Man_t {
    var pFraig *Aig_Man_t;
    var Pars Fra_Par_t;
    var pPars *Fra_Par_t = &Pars; 
    Fra_ParamsDefault( pPars );
    pPars.nBTLimitNode = nConfMax;
    pPars.fChoicing    = 0;
    pPars.fDoSparse    = 1;
    pPars.fSpeculate   = 0;
    pPars.fProve       = fProve;
    pPars.fVerbose     = 0;
    pPars.fDontShowBar = 1;
    pFraig = Fra_FraigPerform( pManAig, pPars );
    return pFraig;
} 

////////////////////////////////////////////////////////////////////////
///                       END OF FILE                                ///
////////////////////////////////////////////////////////////////////////


//ABC_NAMESPACE_IMPL_END

