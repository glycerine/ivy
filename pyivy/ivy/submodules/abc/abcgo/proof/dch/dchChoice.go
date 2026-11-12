package dch

import "fmt"

/**CFile****************************************************************

  FileName    [dchChoice.c]

  SystemName  [ABC: Logic synthesis and verification system.]

  PackageName [Choice computation for tech-mapping.]

  Synopsis    [Contrustion of choices.]

  Author      [Alan Mishchenko]
  
  Affiliation [UC Berkeley]

  Date        [Ver. 1.0. Started - June 29, 2008.]

  Revision    [$Id: dchChoice.c,v 1.00 2008/07/29 00:00:00 alanmi Exp $]

***********************************************************************/

//#include "dchInt.h"

//ABC_NAMESPACE_IMPL_START


////////////////////////////////////////////////////////////////////////
///                        DECLARATIONS                              ///
////////////////////////////////////////////////////////////////////////

////////////////////////////////////////////////////////////////////////
///                     FUNCTION DEFINITIONS                         ///
////////////////////////////////////////////////////////////////////////

/**Function*************************************************************

  Synopsis    [Counts support nodes.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Dch_ObjCountSupp_rec(p *Aig_Man_t, pObj *Aig_Obj_t) int32 {
    var Count int32;
    if Aig_ObjIsTravIdCurrent( p, pObj ) {
        return 0; }
    Aig_ObjSetTravIdCurrent( p, pObj );
    if Aig_ObjIsCi(pObj) {
        return 1; }
    assert(Aig_ObjIsNode(pObj), "Aig_ObjIsNode(pObj)");
    Count  = Dch_ObjCountSupp_rec( p, Aig_ObjFanin0(pObj) );
    Count += Dch_ObjCountSupp_rec( p, Aig_ObjFanin1(pObj) );
    return Count;
}
func Dch_ObjCountSupp(p *Aig_Man_t, pObj *Aig_Obj_t) int32 {
    Aig_ManIncrementTravId( p );
    return Dch_ObjCountSupp_rec( p, pObj );
}

/**Function*************************************************************

  Synopsis    [Counts the number of representatives.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Dch_DeriveChoiceCountReprs(pAig *Aig_Man_t) int32 {
    var pObj, pRepr *Aig_Obj_t;
    var i, nReprs int32;
    // inlining of macro Aig_ManForEachObj
    // inlining of macro Vec_PtrForEachEntry
    for  i = 0; (i < Vec_PtrSize(pAig.vObjs)) && (((pObj) = (Aig_Obj_t *)Vec_PtrEntry(pAig.vObjs, i)), 1); i++  { if (pObj) == nil {} else {
        pRepr = Aig_ObjRepr( pAig, pObj );
        if pRepr == nil {
            continue; }
        assert(pRepr.Id < pObj.Id, "pRepr.Id < pObj.Id");
        nReprs++;
    }
    }
    return nReprs;
}
func Dch_DeriveChoiceCountEquivs(pAig *Aig_Man_t) int32 {
    var pObj, pEquiv *Aig_Obj_t;
    var i, nEquivs int32;
    // inlining of macro Aig_ManForEachObj
    // inlining of macro Vec_PtrForEachEntry
    for  i = 0; (i < Vec_PtrSize(pAig.vObjs)) && (((pObj) = (Aig_Obj_t *)Vec_PtrEntry(pAig.vObjs, i)), 1); i++  { if (pObj) == nil {} else {
        pEquiv = Aig_ObjEquiv( pAig, pObj );
        if pEquiv == nil {
            continue; }
        assert(pEquiv.Id < pObj.Id, "pEquiv.Id < pObj.Id");
        nEquivs++;
    }
    }
    return nEquivs;
}

/**Function*************************************************************

  Synopsis    [Marks the TFI of the node.]

  Description [Returns 1 if there is a CI not marked with previous ID.]
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Dch_ObjMarkTfi_rec(p *Aig_Man_t, pObj *Aig_Obj_t) int32 {
    var RetValue int32;
    if pObj == nil {
        return 0; }
    if Aig_ObjIsTravIdCurrent( p, pObj ) {
        return 0; }
    if Aig_ObjIsCi(pObj) {
        RetValue = !Aig_ObjIsTravIdPrevious( p, pObj );
        Aig_ObjSetTravIdCurrent( p, pObj );
        return RetValue;
    }
    assert(Aig_ObjIsNode(pObj), "Aig_ObjIsNode(pObj)");
    Aig_ObjSetTravIdCurrent( p, pObj );
    RetValue  = Dch_ObjMarkTfi_rec( p, Aig_ObjFanin0(pObj) );
    RetValue += Dch_ObjMarkTfi_rec( p, Aig_ObjFanin1(pObj) );
//    RetValue += Dch_ObjMarkTfi_rec( p, Aig_ObjEquiv(p, pObj) );
    return (RetValue > 0);
}
func Dch_ObjCheckSuppRed(p *Aig_Man_t, pObj *Aig_Obj_t, pRepr *Aig_Obj_t) int32 {
    // mark support of the representative node (pRepr)
    Aig_ManIncrementTravId( p );
    Dch_ObjMarkTfi_rec( p, pRepr );
    // detect if the new node (pObj) depends on additional variables
    Aig_ManIncrementTravId( p );
    if Dch_ObjMarkTfi_rec( p, pObj ) != 0 {
        return 1; }//, printf( "1" );
    // detect if the representative node (pRepr) depends on additional variables
    Aig_ManIncrementTravId( p );
    if Dch_ObjMarkTfi_rec( p, pRepr ) != 0 {
        return 1; }//, printf( "2" );
    // skip the choice if this is what is happening
    return 0;
}

/**Function*************************************************************

  Synopsis    [Make sure reprsentative nodes do not have representatives.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Aig_ManCheckReprs(p *Aig_Man_t) {
    var fPrintConst int32;
    var pObj *Aig_Obj_t;
    var i, fProb int32;
    var Class0, ClassCi int32;
    // inlining of macro Aig_ManForEachObj
    // inlining of macro Vec_PtrForEachEntry
    for  i = 0; (i < Vec_PtrSize(p.vObjs)) && (((pObj) = (Aig_Obj_t *)Vec_PtrEntry(p.vObjs, i)), 1); i++  { if (pObj) == nil {} else {
        if Aig_ObjRepr(p, pObj) == nil {
            continue; }
        if !Aig_ObjIsNode(pObj) {
            fmt.Printf("Obj %d is not an AND but it has a repr %d.\n", i, Aig_ObjId(Aig_ObjRepr(p, pObj)));
fProb = 1;
        } else if Aig_ObjRepr(p, Aig_ObjRepr(p, pObj)) {
            fmt.Printf("Obj %d has repr %d with a repr %d.\n", i, Aig_ObjId(Aig_ObjRepr(p, pObj)), Aig_ObjId(Aig_ObjRepr(p, Aig_ObjRepr(p, pObj))));
fProb = 1; }
    }
    }
    if !(fProb != 0) {
        fmt.Printf("Representive verification successful.\n");
    } else {
        fmt.Printf("Representive verification FAILED.\n"); }
    if !(fPrintConst != 0) {
        return; }
    // count how many representatives are const0 or a CI
    // inlining of macro Aig_ManForEachObj
    // inlining of macro Vec_PtrForEachEntry
    for  i = 0; (i < Vec_PtrSize(p.vObjs)) && (((pObj) = (Aig_Obj_t *)Vec_PtrEntry(p.vObjs, i)), 1); i++  { if (pObj) == nil {} else {
        if Aig_ObjRepr(p, pObj) == Aig_ManConst1(p) {
            Class0++; }
        if Aig_ObjRepr(p, pObj) && Aig_ObjIsCi(Aig_ObjRepr(p, pObj)) {
            ClassCi++; }
    }
    }
    fmt.Printf("Const0 nodes = %d.  ConstCI nodes = %d.\n", Class0, ClassCi);
}

/**Function*************************************************************

  Synopsis    [Verify correctness of choices.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Dch_CheckChoices(p *Aig_Man_t, fSkipRedSupps int32) {
    var pObj *Aig_Obj_t;
    var i, fProb int32;
    Aig_ManCleanMarkA( p );
    // inlining of macro Aig_ManForEachNode
    // inlining of macro Vec_PtrForEachEntry
    for  i = 0; (i < Vec_PtrSize(p.vObjs)) && (((pObj) = (Aig_Obj_t *)Vec_PtrEntry(p.vObjs, i)), 1); i++  { if (pObj) == nil || !Aig_ObjIsNode(pObj) {} else {
        if p.pEquivs[i] != nil {
            if pObj.fMarkA == 1 {
                fmt.Printf("node %d participates in more than one choice class\n", i);
fProb = 1; }
            pObj.fMarkA = 1;
            // check redundancy
            if (fSkipRedSupps != 0) && (Dch_ObjCheckSuppRed( p, pObj, p.pEquivs[i]) != 0) {
                fmt.Printf("node %d and repr %d have diff supports\n", pObj.Id, p.pEquivs[i].Id); }
            // consider the next one
            pObj = p.pEquivs[i];
            if p.pEquivs[Aig_ObjId(pObj)] == nil {
                if pObj.fMarkA == 1 {
                    fmt.Printf("repr %d has final node %d participates in more than one choice class\n", i, pObj.Id);
fProb = 1; }
                pObj.fMarkA = 1;
            }
            // consider the non-head ones
            if pObj.nRefs > 0 {
                fmt.Printf("node %d belonging to choice has fanout %d\n", pObj.Id, pObj.nRefs); }
        }
        if p.pReprs && p.pReprs[i] != nil {
            if pObj.nRefs > 0 {
                fmt.Printf("node %d has representative %d and fanout count %d\n", pObj.Id, p.pReprs[i].Id, pObj.nRefs);
fProb = 1; }
        }
    }
    }
    Aig_ManCleanMarkA( p );
    if !(fProb != 0) {
        fmt.Printf("Verification of choice AIG succeeded.\n");
    } else {
        fmt.Printf("Verification of choice AIG FAILED.\n"); }
}

/**Function*************************************************************

  Synopsis    [Checks for combinational loops in the AIG.]

  Description [Returns 1 if combinational loop is detected.]
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Aig_ManCheckAcyclic_rec(p *Aig_Man_t, pNode *Aig_Obj_t, fVerbose int32) int32 {
    var pFanin *Aig_Obj_t;
    var fAcyclic int32;
    if Aig_ObjIsCi(pNode) || Aig_ObjIsConst1(pNode) {
        return 1; }
    assert(Aig_ObjIsNode(pNode), "Aig_ObjIsNode(pNode)");
    // make sure the node is not visited
    assert(!Aig_ObjIsTravIdPrevious(p, pNode), "!Aig_ObjIsTravIdPrevious(p, pNode)");
    // check if the node is part of the combinational loop
    if Aig_ObjIsTravIdCurrent(p, pNode) {
        if fVerbose != 0 {
            Abc_Print( 1, "Network \"%s\" contains combinational loop!\n", p.pSpec? p.pSpec : nil ); }
        if fVerbose != 0 {
        Abc_Print( 1, "Node \"%d\" is encountered twice on the following path to the COs:\n", Aig_ObjId(pNode) ); }
        return 0;
    }
    // mark this node as a node on the current path
    Aig_ObjSetTravIdCurrent( p, pNode );

    // visit the transitive fanin
    pFanin = Aig_ObjFanin0(pNode);
    // check if the fanin is visited
    if !Aig_ObjIsTravIdPrevious(p, pFanin) {
        // traverse the fanin's cone searching for the loop
        if !(fAcyclic = Aig_ManCheckAcyclic_rec(p, pFanin, fVerbose)) {
            // return as soon as the loop is detected
            if fVerbose != 0 {
            Abc_Print( 1, " %d ->", Aig_ObjId(pFanin) ); }
            return 0;
        }
    }

    // visit the transitive fanin
    pFanin = Aig_ObjFanin1(pNode);
    // check if the fanin is visited
    if !Aig_ObjIsTravIdPrevious(p, pFanin) {
        // traverse the fanin's cone searching for the loop
        if !(fAcyclic = Aig_ManCheckAcyclic_rec(p, pFanin, fVerbose)) {
            // return as soon as the loop is detected
            if fVerbose != 0 {
            Abc_Print( 1, " %d ->", Aig_ObjId(pFanin) ); }
            return 0;
        }
    }

    // visit choices
    if Aig_ObjRepr(p, pNode) == nil && Aig_ObjEquiv(p, pNode) != nil {
        for  pFanin = Aig_ObjEquiv(p, pNode); (pFanin != nil); pFanin = Aig_ObjEquiv(p, pFanin)  {
            // check if the fanin is visited
            if Aig_ObjIsTravIdPrevious(p, pFanin) { 
                continue; }
            // traverse the fanin's cone searching for the loop
            if fAcyclic = Aig_ManCheckAcyclic_rec(p, pFanin, fVerbose) {
                continue; }
            // return as soon as the loop is detected
            if fVerbose != 0 {
            Abc_Print( 1, " %d", Aig_ObjId(pFanin) ); }
            if fVerbose != 0 {
            Abc_Print( 1, " (choice of %d) -> ", Aig_ObjId(pNode) ); }
            return 0;
        }
    }
    // mark this node as a visited node
    Aig_ObjSetTravIdPrevious( p, pNode );
    return 1;
}
func Aig_ManCheckAcyclic(p *Aig_Man_t, fVerbose int32) int32 {
    var pNode *Aig_Obj_t;
    var fAcyclic int32;
    var i int32;
    // set the traversal ID for this DFS ordering
    Aig_ManIncrementTravId( p );   
    Aig_ManIncrementTravId( p );   
    // pNode->TravId == pNet->nTravIds      means "pNode is on the path"
    // pNode->TravId == pNet->nTravIds - 1  means "pNode is visited but is not on the path"
    // pNode->TravId <  pNet->nTravIds - 1  means "pNode is not visited"
    // traverse the network to detect cycles
    fAcyclic = 1;
    // inlining of macro Aig_ManForEachCo
    // inlining of macro Vec_PtrForEachEntry
    for  i = 0; (i < Vec_PtrSize(p.vCos)) && (((pNode) = (Aig_Obj_t *)Vec_PtrEntry(p.vCos, i)), 1); i++  {
        pNode = Aig_ObjFanin0(pNode);
        if Aig_ObjIsTravIdPrevious(p, pNode) {
            continue; }
        // traverse the output logic cone
        if fAcyclic = Aig_ManCheckAcyclic_rec(p, pNode, fVerbose) {
            continue; }
        // stop as soon as the first loop is detected
        if fVerbose != 0 {
        Abc_Print( 1, " CO %d\n", i ); }
        break;
    }
    return fAcyclic;
}


/**Function*************************************************************

  Synopsis    [Returns 1 if the choice node of pRepr is in the TFI of pObj.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Dch_ObjCheckTfi_rec(p *Aig_Man_t, pObj *Aig_Obj_t) int32 {
    // check the trivial cases
    if pObj == nil {
        return 0; }
    if Aig_ObjIsCi(pObj) {
        return 0; }
    if pObj.fMarkA {
        return 1; }
    // skip the visited node
    if Aig_ObjIsTravIdCurrent( p, pObj ) {
        return 0; }
    Aig_ObjSetTravIdCurrent( p, pObj );
    // check the children
    if Dch_ObjCheckTfi_rec( p, Aig_ObjFanin0(pObj) ) != 0 {
        return 1; }
    if Dch_ObjCheckTfi_rec( p, Aig_ObjFanin1(pObj) ) != 0 {
        return 1; }
    // check equivalent nodes
    return Dch_ObjCheckTfi_rec( p, Aig_ObjEquiv(p, pObj) );
}
func Dch_ObjCheckTfi(p *Aig_Man_t, pObj *Aig_Obj_t, pRepr *Aig_Obj_t) int32 {
    var pTemp *Aig_Obj_t;
    var RetValue int32;
    assert(!Aig_IsComplement(pObj), "!Aig_IsComplement(pObj)");
    assert(!Aig_IsComplement(pRepr), "!Aig_IsComplement(pRepr)");
    // mark nodes of the choice node
    for  pTemp = pRepr; pTemp; pTemp = Aig_ObjEquiv(p, pTemp)  {
        pTemp.fMarkA = 1;
    }
    // traverse the new node
    Aig_ManIncrementTravId( p );
    RetValue = Dch_ObjCheckTfi_rec( p, pObj );
    // unmark nodes of the choice node
    for  pTemp = pRepr; pTemp; pTemp = Aig_ObjEquiv(p, pTemp)  {
        pTemp.fMarkA = 0;
    }
    return RetValue;
}

/**Function*************************************************************

  Synopsis    [Returns representatives of fanin in approapriate polarity.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Aig_ObjGetRepr(p *Aig_Man_t, pObj *Aig_Obj_t) *Aig_Obj_t {
    var pRepr *Aig_Obj_t;
    if pRepr = Aig_ObjRepr(p, Aig_Regular(pObj)) {
        return Aig_NotCond( pRepr, Aig_Regular(pObj).fPhase ^ pRepr.fPhase ^ Aig_IsComplement(pObj) ); }
    return pObj;
}
func Aig_ObjChild0CopyRepr(p *Aig_Man_t, pObj *Aig_Obj_t) *Aig_Obj_t { return Aig_ObjGetRepr( p, Aig_ObjChild0Copy(pObj) ); }
func Aig_ObjChild1CopyRepr(p *Aig_Man_t, pObj *Aig_Obj_t) *Aig_Obj_t { return Aig_ObjGetRepr( p, Aig_ObjChild1Copy(pObj) ); }

/**Function*************************************************************

  Synopsis    [Derives the AIG with choices from representatives.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Dch_DeriveChoiceAigNode(pAigNew *Aig_Man_t, pAigOld *Aig_Man_t, pObj *Aig_Obj_t, fSkipRedSupps int32) {
    var pRepr, pObjNew, pReprNew *Aig_Obj_t;
    assert(!Aig_IsComplement(pObj), "!Aig_IsComplement(pObj)");
    // get the representative
    pRepr = Aig_ObjRepr( pAigOld, pObj );
    if pRepr != nil && (Aig_ObjIsConst1(pRepr) || Aig_ObjIsCi(pRepr)) {
        assert(pRepr.pData != nil, "pRepr.pData != nil");
        pObj.pData = Aig_NotCond( (Aig_Obj_t *)pRepr.pData, pObj.fPhase ^ pRepr.fPhase );
        return;
    }
    // get the new node
    pObjNew = Aig_And( pAigNew, 
        Aig_ObjChild0CopyRepr(pAigNew, pObj), 
        Aig_ObjChild1CopyRepr(pAigNew, pObj) );
//    pObjNew = Aig_ObjGetRepr( pAigNew, pObjNew );
    for {
        var pObjNew2 *Aig_Obj_t  = pObjNew;
        pObjNew = Aig_ObjGetRepr( pAigNew, pObjNew2 );
        if pObjNew == pObjNew2 {
            break; }
    }
//    assert( Aig_ObjRepr( pAigNew, pObjNew ) == NULL );
    // assign the copy
    assert(pObj.pData == nil, "pObj.pData == nil");
    pObj.pData = pObjNew;
    // skip those without reprs
    if pRepr == nil {
        return; }
    assert(pRepr.Id < pObj.Id, "pRepr.Id < pObj.Id");
    assert(Aig_ObjIsNode(pRepr), "Aig_ObjIsNode(pRepr)");
    // get the corresponding new nodes
    pObjNew  = Aig_Regular((Aig_Obj_t *)pObj.pData);
    pReprNew = Aig_Regular((Aig_Obj_t *)pRepr.pData);
    // skip earlier nodes
    if pReprNew.Id >= pObjNew.Id {
        return; }
    assert(pReprNew.Id < pObjNew.Id, "pReprNew.Id < pObjNew.Id");
    // set the representatives
    Aig_ObjSetRepr( pAigNew, pObjNew, pReprNew );
    // skip used nodes
    if pObjNew.nRefs > 0 {
        return; }
    assert(pObjNew.nRefs == 0, "pObjNew.nRefs == 0");
    // skip choices that can lead to combo loops
    if Dch_ObjCheckTfi( pAigNew, pObjNew, pReprNew ) != 0 {
        return; }
    // don't add choice if structural support of pObjNew and pReprNew differ
    if (fSkipRedSupps != 0) && (Dch_ObjCheckSuppRed(pAigNew, pObjNew, pReprNew) != 0) {
        return; }
    // add choice to the end of the list
    for  pAigNew.pEquivs[pReprNew.Id] != nil  {
        pReprNew = pAigNew.pEquivs[pReprNew.Id];
    }
    assert(pAigNew.pEquivs[pReprNew.Id] == nil, "pAigNew.pEquivs[pReprNew.Id] == nil");
    pAigNew.pEquivs[pReprNew.Id] = pObjNew;
}
func Dch_DeriveChoiceAigInt(pAig *Aig_Man_t, fSkipRedSupps int32) *Aig_Man_t {
    var pChoices *Aig_Man_t;
    var pObj *Aig_Obj_t;
    var i int32;
    // start recording equivalences
    pChoices = Aig_ManStart( Aig_ManObjNumMax(pAig) );
    // inlining of macro ABC_CALLOC
    pChoices.pEquivs = ((Aig_Obj_t * *) calloc((size_t)(Aig_ManObjNumMax(pAig)), sizeof(Aig_Obj_t *)));
    // inlining of macro ABC_CALLOC
    pChoices.pReprs  = ((Aig_Obj_t * *) calloc((size_t)(Aig_ManObjNumMax(pAig)), sizeof(Aig_Obj_t *)));
    // map constants and PIs
    Aig_ManCleanData( pAig );
    Aig_ManConst1(pAig).pData = Aig_ManConst1(pChoices);
    // inlining of macro Aig_ManForEachCi
    // inlining of macro Vec_PtrForEachEntry
    for  i = 0; (i < Vec_PtrSize(pAig.vCis)) && (((pObj) = (Aig_Obj_t *)Vec_PtrEntry(pAig.vCis, i)), 1); i++  {
        pObj.pData = Aig_ObjCreateCi( pChoices );
    }
    // construct choices for the internal nodes
    assert(pAig.pReprs != nil, "pAig.pReprs != nil");
    // inlining of macro Aig_ManForEachNode
    // inlining of macro Vec_PtrForEachEntry
    for  i = 0; (i < Vec_PtrSize(pAig.vObjs)) && (((pObj) = (Aig_Obj_t *)Vec_PtrEntry(pAig.vObjs, i)), 1); i++  { if (pObj) == nil || !Aig_ObjIsNode(pObj) {} else { 
        Dch_DeriveChoiceAigNode( pChoices, pAig, pObj, fSkipRedSupps ); }
    }
    // inlining of macro Aig_ManForEachCo
    // inlining of macro Vec_PtrForEachEntry
    for  i = 0; (i < Vec_PtrSize(pAig.vCos)) && (((pObj) = (Aig_Obj_t *)Vec_PtrEntry(pAig.vCos, i)), 1); i++  {
        Aig_ObjCreateCo( pChoices, Aig_ObjChild0CopyRepr(pChoices, pObj) );
    }
    Aig_ManSetRegNum( pChoices, Aig_ManRegNum(pAig) );
    return pChoices;
}
func Dch_DeriveChoiceAig(pAig *Aig_Man_t, fSkipRedSupps int32) *Aig_Man_t {
    var fCheck int32;
    var pChoices, pTemp *Aig_Man_t;
    // verify
    if fCheck != 0 {
        Aig_ManCheckReprs( pAig ); } 
    // compute choices
    pChoices = Dch_DeriveChoiceAigInt( pAig, fSkipRedSupps );
    pChoices.pReprs = nil;
    // verify
    if fCheck != 0 {
        Dch_CheckChoices( pChoices, fSkipRedSupps ); } 
    // find correct topo order with choices
    pChoices = Aig_ManDupDfs( pTemp = pChoices );
    Aig_ManStop( pTemp );
    // verify
    if fCheck != 0 {
        Dch_CheckChoices( pChoices, fSkipRedSupps ); } 
    return pChoices;
}

////////////////////////////////////////////////////////////////////////
///                       END OF FILE                                ///
////////////////////////////////////////////////////////////////////////


//ABC_NAMESPACE_IMPL_END

