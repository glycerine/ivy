package fraig

import "fmt"

import "os"

/**CFile****************************************************************

  FileName    [fraigUtil.c]

  PackageName [FRAIG: Functionally reduced AND-INV graphs.]

  Synopsis    [Various utilities.]

  Author      [Alan Mishchenko <alanmi@eecs.berkeley.edu>]
  
  Affiliation [UC Berkeley]

  Date        [Ver. 2.0. Started - October 1, 2004]

  Revision    [$Id: fraigUtil.c,v 1.15 2005/07/08 01:01:34 alanmi Exp $]

***********************************************************************/

//#include "fraigInt.h"
//#include <limits.h>

//ABC_NAMESPACE_IMPL_START


////////////////////////////////////////////////////////////////////////
///                        DECLARATIONS                              ///
////////////////////////////////////////////////////////////////////////

var bit_count [256]int32 = [256]int32{
  0,1,1,2,1,2,2,3,1,2,2,3,2,3,3,4,1,2,2,3,2,3,3,4,2,3,3,4,3,4,4,5,
  1,2,2,3,2,3,3,4,2,3,3,4,3,4,4,5,2,3,3,4,3,4,4,5,3,4,4,5,4,5,5,6,
  1,2,2,3,2,3,3,4,2,3,3,4,3,4,4,5,2,3,3,4,3,4,4,5,3,4,4,5,4,5,5,6,
  2,3,3,4,3,4,4,5,3,4,4,5,4,5,5,6,3,4,4,5,4,5,5,6,4,5,5,6,5,6,6,7,
  1,2,2,3,2,3,3,4,2,3,3,4,3,4,4,5,2,3,3,4,3,4,4,5,3,4,4,5,4,5,5,6,
  2,3,3,4,3,4,4,5,3,4,4,5,4,5,5,6,3,4,4,5,4,5,5,6,4,5,5,6,5,6,6,7,
  2,3,3,4,3,4,4,5,3,4,4,5,4,5,5,6,3,4,4,5,4,5,5,6,4,5,5,6,5,6,6,7,
  3,4,4,5,4,5,5,6,4,5,5,6,5,6,6,7,4,5,5,6,5,6,6,7,5,6,6,7,6,7,7,8,
};

// static void Fraig_Dfs_rec( Fraig_Man_t * pMan, Fraig_Node_t * pNode, Fraig_NodeVec_t * vNodes, int fEquiv );

// static int  Fraig_CheckTfi_rec( Fraig_Man_t * pMan, Fraig_Node_t * pNode, Fraig_Node_t * pOld );


////////////////////////////////////////////////////////////////////////
///                     FUNCTION DEFINITIONS                         ///
////////////////////////////////////////////////////////////////////////

/**Function*************************************************************

  Synopsis    [Computes the DFS ordering of the nodes.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Fraig_Dfs(pMan *Fraig_Man_t, fEquiv int32) *Fraig_NodeVec_t {
    var vNodes *Fraig_NodeVec_t;
    var i int32;
    pMan.nTravIds++;
    vNodes = Fraig_NodeVecAlloc( 100 );
    for  i = 0; i < pMan.vOutputs.nSize; i++  {
        // inlining of macro Fraig_Regular
        Fraig_Dfs_rec( pMan, ((Fraig_Node_t *)((ABC_PTRUINT_T)(pMan.vOutputs.pArray[i]) & ~01)) , vNodes, fEquiv );
    }
    return vNodes;
}

/**Function*************************************************************

  Synopsis    [Computes the DFS ordering of the nodes.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Fraig_DfsOne(pMan *Fraig_Man_t, pNode *Fraig_Node_t, fEquiv int32) *Fraig_NodeVec_t {
    var vNodes *Fraig_NodeVec_t;
    pMan.nTravIds++;
    vNodes = Fraig_NodeVecAlloc( 100 );
    // inlining of macro Fraig_Regular
    Fraig_Dfs_rec( pMan, ((Fraig_Node_t *)((ABC_PTRUINT_T)(pNode) & ~01)) , vNodes, fEquiv );
    return vNodes;
}

/**Function*************************************************************

  Synopsis    [Computes the DFS ordering of the nodes.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Fraig_DfsNodes(pMan *Fraig_Man_t, ppNodes **Fraig_Node_t, nNodes int32, fEquiv int32) *Fraig_NodeVec_t {
    var vNodes *Fraig_NodeVec_t;
    var i int32;
    pMan.nTravIds++;
    vNodes = Fraig_NodeVecAlloc( 100 );
    for  i = 0; i < nNodes; i++  {
        // inlining of macro Fraig_Regular
        Fraig_Dfs_rec( pMan, ((Fraig_Node_t *)((ABC_PTRUINT_T)(ppNodes[i]) & ~01)) , vNodes, fEquiv );
    }
    return vNodes;
}

/**Function*************************************************************

  Synopsis    [Recursively computes the DFS ordering of the nodes.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Fraig_Dfs_rec(pMan *Fraig_Man_t, pNode *Fraig_Node_t, vNodes *Fraig_NodeVec_t, fEquiv int32) {
    // inlining of macro Fraig_IsComplement
    assert(!(((int32(((ABC_PTRUINT_T) (pNode) & 01)) != 0))), "!/*__starting_point_inline_macro:Fraig_IsComplement__*/((int32(((ABC_PTRUINT_T) (pNode) & 01))))");
    // skip the visited node
    if pNode.TravId == pMan.nTravIds {
        return; }
    pNode.TravId = pMan.nTravIds;
    // visit the transitive fanin
    if Fraig_NodeIsAnd(pNode) {
        // inlining of macro Fraig_Regular
        Fraig_Dfs_rec( pMan, ((Fraig_Node_t *)((ABC_PTRUINT_T)(pNode.p1) & ~01)) , vNodes, fEquiv );
        // inlining of macro Fraig_Regular
        Fraig_Dfs_rec( pMan, ((Fraig_Node_t *)((ABC_PTRUINT_T)(pNode.p2) & ~01)) , vNodes, fEquiv );
    }
    if (fEquiv != 0) && pNode.pNextE {
        Fraig_Dfs_rec( pMan, pNode.pNextE, vNodes, fEquiv ); }
    // save the node
    Fraig_NodeVecPush( vNodes, pNode );
}

/**Function*************************************************************

  Synopsis    [Computes the DFS ordering of the nodes.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Fraig_CountNodes(pMan *Fraig_Man_t, fEquiv int32) int32 {
    var vNodes *Fraig_NodeVec_t;
    var RetValue int32;
    vNodes = Fraig_Dfs( pMan, fEquiv );
    RetValue = vNodes.nSize;
    Fraig_NodeVecFree( vNodes );
    return RetValue;
}

/**Function*************************************************************

  Synopsis    [Returns 1 if pOld is in the TFI of pNew.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Fraig_CheckTfi(pMan *Fraig_Man_t, pOld *Fraig_Node_t, pNew *Fraig_Node_t) int32 {
    // inlining of macro Fraig_IsComplement
    assert(!(((int32(((ABC_PTRUINT_T) (pOld) & 01)) != 0))), "!/*__starting_point_inline_macro:Fraig_IsComplement__*/((int32(((ABC_PTRUINT_T) (pOld) & 01))))");
    // inlining of macro Fraig_IsComplement
    assert(!(((int32(((ABC_PTRUINT_T) (pNew) & 01)) != 0))), "!/*__starting_point_inline_macro:Fraig_IsComplement__*/((int32(((ABC_PTRUINT_T) (pNew) & 01))))");
    pMan.nTravIds++;
    return Fraig_CheckTfi_rec( pMan, pNew, pOld );
}

/**Function*************************************************************

  Synopsis    [Returns 1 if pOld is in the TFI of pNew.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Fraig_CheckTfi_rec(pMan *Fraig_Man_t, pNode *Fraig_Node_t, pOld *Fraig_Node_t) int32 {
    // check the trivial cases
    if pNode == nil {
        return 0; }
    if pNode.Num < pOld.Num && !pMan.fChoicing {
        return 0; }
    if pNode == pOld {
        return 1; }
    // skip the visited node
    if pNode.TravId == pMan.nTravIds {
        return 0; }
    pNode.TravId = pMan.nTravIds;
    // check the children
    // inlining of macro Fraig_Regular
    if Fraig_CheckTfi_rec( pMan, ((Fraig_Node_t *)((ABC_PTRUINT_T)(pNode.p1) & ~01)) , pOld ) {
        return 1; }
    // inlining of macro Fraig_Regular
    if Fraig_CheckTfi_rec( pMan, ((Fraig_Node_t *)((ABC_PTRUINT_T)(pNode.p2) & ~01)) , pOld ) {
        return 1; }
    // check equivalent nodes
    return Fraig_CheckTfi_rec( pMan, pNode.pNextE, pOld );
}


/**Function*************************************************************

  Synopsis    [Returns 1 if pOld is in the TFI of pNew.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Fraig_CheckTfi2(pMan *Fraig_Man_t, pOld *Fraig_Node_t, pNew *Fraig_Node_t) int32 {
    var vNodes *Fraig_NodeVec_t;
    var RetValue int32;
    vNodes = Fraig_DfsOne( pMan, pNew, 1 );
    RetValue = (pOld.TravId == pMan.nTravIds);
    Fraig_NodeVecFree( vNodes );
    return RetValue;
}

/**Function*************************************************************

  Synopsis    [Sets the number of fanouts (none, one, or many).]

  Description [This procedure collects the nodes reachable from
  the POs of the AIG and sets the type of fanout counter (none, one,
  or many) for each node. This procedure is useful to determine
  fanout-free cones of AND-nodes, which is helpful for rebalancing
  the AIG (see procedure Fraig_ManRebalance, or something like that).]
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Fraig_ManMarkRealFanouts(p *Fraig_Man_t) {
    var vNodes *Fraig_NodeVec_t;
    var pNodeR *Fraig_Node_t;
    var i int32;
    // collect the nodes reachable
    vNodes = Fraig_Dfs( p, 0 );
    // clean the fanouts field
    for  i = 0; i < vNodes.nSize; i++  {
        vNodes.pArray[i].nFanouts = 0;
        vNodes.pArray[i].pData0 = nil;
    }
    // mark reachable nodes by setting the two-bit counter pNode->nFans
    for  i = 0; i < vNodes.nSize; i++  {
        // inlining of macro Fraig_Regular
        pNodeR = ((Fraig_Node_t *)((ABC_PTRUINT_T)(vNodes.pArray[i].p1) & ~01)) ;
        if pNodeR && ++pNodeR.nFanouts == 3 {
            pNodeR.nFanouts = 2; }
        // inlining of macro Fraig_Regular
        pNodeR = ((Fraig_Node_t *)((ABC_PTRUINT_T)(vNodes.pArray[i].p2) & ~01)) ;
        if pNodeR && ++pNodeR.nFanouts == 3 {
            pNodeR.nFanouts = 2; }
    }
    Fraig_NodeVecFree( vNodes );
}

/**Function*************************************************************

  Synopsis    [Creates the constant 1 node.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Fraig_BitStringCountOnes(pString *uint32, nWords int32) int32 {
    unsigned char * pSuppBytes = (unsigned char *)pString;
    var i int32;
    var nOnes int32;
    var nBytes int32 = sizeof(unsigned) * nWords;
    // count the number of ones in the simulation vector
    for  i = nOnes = 0; i < nBytes; i++  {
        nOnes += bit_count[pSuppBytes[i]];
    }
    return nOnes;
}

/**Function*************************************************************

  Synopsis    [Verify one useful property.]

  Description [This procedure verifies one useful property. After 
  the FRAIG construction with choice nodes is over, each primary node 
  should have fanins that are primary nodes. The primary nodes is the 
  one that does not have pNode->pRepr set to point to another node.]
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Fraig_ManCheckConsistency(p *Fraig_Man_t) int32 {
    var pNode *Fraig_Node_t;
    var pVec *Fraig_NodeVec_t;
    var i int32;
    pVec = Fraig_Dfs( p, 0 );
    for  i = 0; i < pVec.nSize; i++  {
        pNode = pVec.pArray[i];
        if Fraig_NodeIsVar(pNode) {
            if pNode.pRepr {
                fmt.Printf("Primary input %d is a secondary node.\n", pNode.Num); }
        
        } else if Fraig_NodeIsConst(pNode) {
            if pNode.pRepr {
                fmt.Printf("Constant 1 %d is a secondary node.\n", pNode.Num); }
        
        } else {
            if pNode.pRepr {
                fmt.Printf("Internal node %d is a secondary node.\n", pNode.Num); }
            // inlining of macro Fraig_Regular
            if  ((Fraig_Node_t *)((ABC_PTRUINT_T)(pNode.p1) & ~01)) .pRepr {
                // inlining of macro Fraig_Regular
                fmt.Printf("Internal node %d has first fanin %d that is a secondary node.\n", pNode.Num, ((Fraig_Node_t *)((ABC_PTRUINT_T)(pNode.p1) & ~01)) .Num); }
            // inlining of macro Fraig_Regular
            if  ((Fraig_Node_t *)((ABC_PTRUINT_T)(pNode.p2) & ~01)) .pRepr {
                // inlining of macro Fraig_Regular
                fmt.Printf("Internal node %d has second fanin %d that is a secondary node.\n", pNode.Num, ((Fraig_Node_t *)((ABC_PTRUINT_T)(pNode.p2) & ~01)) .Num); }
        }
    }
    Fraig_NodeVecFree( pVec );
    return 1;
}

/**Function*************************************************************

  Synopsis    [Prints the node.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Fraig_PrintNode(p *Fraig_Man_t, pNode *Fraig_Node_t) {
    var vNodes *Fraig_NodeVec_t;
    var pTemp *Fraig_Node_t;
    var fCompl1, fCompl2, i int32;

    vNodes = Fraig_DfsOne( p, pNode, 0 );
    for  i = 0; i < vNodes.nSize; i++  {
        pTemp = vNodes.pArray[i];
        if Fraig_NodeIsVar(pTemp) {
            fmt.Printf("%3d : PI          ", pTemp.Num);
            Fraig_PrintBinary( os.Stdout, (unsigned *)&pTemp.puSimR, 20 );
            fmt.Printf("   ");
            Fraig_PrintBinary( os.Stdout, (unsigned *)&pTemp.puSimD, 20 );
            fmt.Printf("  %d\n", pTemp.fInv);
            continue;
        }

        // inlining of macro Fraig_IsComplement
        fCompl1 = ((int32(((ABC_PTRUINT_T) (pTemp.p1) & 01))));
        // inlining of macro Fraig_IsComplement
        fCompl2 = ((int32(((ABC_PTRUINT_T) (pTemp.p2) & 01))));
        // inlining of macro Fraig_Regular
        fmt.Printf("%3d : %c%3d %c%3d   ", pTemp.Num, (fCompl1? '-':'+'), ((Fraig_Node_t *)((ABC_PTRUINT_T)(pTemp.p1) & ~01)) .Num, (fCompl2? '-':'+'), ((Fraig_Node_t *)((ABC_PTRUINT_T)(pTemp.p2) & ~01)) .Num);
        Fraig_PrintBinary( os.Stdout, (unsigned *)&pTemp.puSimR, 20 );
        fmt.Printf("   ");
        Fraig_PrintBinary( os.Stdout, (unsigned *)&pTemp.puSimD, 20 );
        fmt.Printf("  %d\n", pTemp.fInv);
    }
    Fraig_NodeVecFree( vNodes );
}

/**Function*************************************************************

  Synopsis    [Prints the bit string.]

  Description []

  SideEffects []

  SeeAlso     []

***********************************************************************/
func Fraig_PrintBinary(pFile *os.File, pSign *uint32, nBits int32) {
    var Remainder, nWords int32;
    var w, i int32;

    Remainder = (nBits%(sizeof(unsigned)*8));
    nWords    = (nBits/(sizeof(unsigned)*8)) + (Remainder>0);

    for  w = nWords-1; w >= 0; w--  {
        for  i = ((w == nWords-1 && Remainder)? Remainder-1: 31); i >= 0; i--  {
            fmt.Fprintf(pFile, "%c", '0' + int32(((pSign[w] & (1<<i)) > 0)));
        }
    }

//  fprintf( pFile, "\n" );
}

/**Function*************************************************************

  Synopsis    [Sets up the mask.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Fraig_GetMaxLevel(pMan *Fraig_Man_t) int32 {
    var nLevelMax, i int32;
    nLevelMax = 0;
    for  i = 0; i < pMan.vOutputs.nSize; i++  {
        // inlining of macro Fraig_Regular
        if nLevelMax > ((Fraig_Node_t *)((ABC_PTRUINT_T)(pMan.vOutputs.pArray[i]) & ~01)) .Level {
            nLevelMax = nLevelMax;
        } else {
            // inlining of macro Fraig_Regular
            nLevelMax = ((Fraig_Node_t *)((ABC_PTRUINT_T)(pMan.vOutputs.pArray[i]) & ~01)) .Level;
        }
    }
    return nLevelMax;
}

/**Function*************************************************************

  Synopsis    [Analyses choice nodes.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Fraig_MappingUpdateLevel_rec(pMan *Fraig_Man_t, pNode *Fraig_Node_t, fMaximum int32) int32 {
    var pTemp *Fraig_Node_t;
    var Level1, Level2, LevelE int32;
    // inlining of macro Fraig_IsComplement
    assert(!(((int32(((ABC_PTRUINT_T) (pNode) & 01)) != 0))), "!/*__starting_point_inline_macro:Fraig_IsComplement__*/((int32(((ABC_PTRUINT_T) (pNode) & 01))))");
    if !Fraig_NodeIsAnd(pNode) {
        return pNode.Level; }
    // skip the visited node
    if pNode.TravId == pMan.nTravIds {
        return pNode.Level; }
    pNode.TravId = pMan.nTravIds;
    // compute levels of the children nodes
    // inlining of macro Fraig_Regular
    Level1 = Fraig_MappingUpdateLevel_rec( pMan, ((Fraig_Node_t *)((ABC_PTRUINT_T)(pNode.p1) & ~01)) , fMaximum );
    // inlining of macro Fraig_Regular
    Level2 = Fraig_MappingUpdateLevel_rec( pMan, ((Fraig_Node_t *)((ABC_PTRUINT_T)(pNode.p2) & ~01)) , fMaximum );
    pNode.Level = 1 + Abc_MaxInt( Level1, Level2 );
    if pNode.pNextE {
        LevelE = Fraig_MappingUpdateLevel_rec( pMan, pNode.pNextE, fMaximum );
        if fMaximum != 0 {
            if pNode.Level < LevelE {
                pNode.Level = LevelE; }
        
        } else {
            if pNode.Level > LevelE {
                pNode.Level = LevelE; }
        }
        // set the level of all equivalent nodes to be the same minimum
        if pNode.pRepr == nil { // the primary node
            for  pTemp = pNode.pNextE; (pTemp != nil); pTemp = pTemp.pNextE  {
                pTemp.Level = pNode.Level;
            } }
    }
    return pNode.Level;
}

/**Function*************************************************************

  Synopsis    [Resets the levels of the nodes in the choice graph.]

  Description [Makes the level of the choice nodes to be equal to the
  maximum of the level of the nodes in the equivalence class. This way
  sorting by level leads to the reverse topological order, which is
  needed for the required time computation.]
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Fraig_MappingSetChoiceLevels(pMan *Fraig_Man_t, fMaximum int32) {
    var i int32;
    pMan.nTravIds++;
    for  i = 0; i < pMan.vOutputs.nSize; i++  {
        // inlining of macro Fraig_Regular
        Fraig_MappingUpdateLevel_rec( pMan, ((Fraig_Node_t *)((ABC_PTRUINT_T)(pMan.vOutputs.pArray[i]) & ~01)) , fMaximum );
    }
}

/**Function*************************************************************

  Synopsis    [Reports statistics on choice nodes.]

  Description [The number of choice nodes is the number of primary nodes,
  which has pNextE set to a pointer. The number of choices is the number
  of entries in the equivalent-node lists of the primary nodes.]
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Fraig_ManReportChoices(pMan *Fraig_Man_t) {
    var pNode, pTemp *Fraig_Node_t;
    var nChoiceNodes, nChoices int32;
    var i, LevelMax1, LevelMax2 int32;

    // report the number of levels
    LevelMax1 = Fraig_GetMaxLevel( pMan );
    Fraig_MappingSetChoiceLevels( pMan, 0 );
    LevelMax2 = Fraig_GetMaxLevel( pMan );

    // report statistics about choices
    nChoiceNodes = nChoices = 0;
    for  i = 0; i < pMan.vNodes.nSize; i++  {
        pNode = pMan.vNodes.pArray[i];
        if pNode.pRepr == nil && pNode.pNextE != nil { // this is a choice node = the primary node that has equivalent nodes
            nChoiceNodes++;
            for  pTemp = pNode; (pTemp != nil); pTemp = pTemp.pNextE  {
                nChoices++;
            }
        }
    }
    fmt.Printf("Maximum level: Original = %d. Reduced due to choices = %d.\n", LevelMax1, LevelMax2);
    fmt.Printf("Choice stats:  Choice nodes = %d. Total choices = %d.\n", nChoiceNodes, nChoices);
}

/**Function*************************************************************

  Synopsis    [Returns 1 if the node is the root of EXOR/NEXOR gate.]

  Description [The node can be complemented.]
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Fraig_NodeIsExorType(pNode *Fraig_Node_t) int32 {
    var pNode1, pNode2 *Fraig_Node_t;
    // make the node regular (it does not matter for EXOR/NEXOR)
    // inlining of macro Fraig_Regular
    pNode = ((Fraig_Node_t *)((ABC_PTRUINT_T)(pNode) & ~01)) ;
    // if the node or its children are not ANDs or not compl, this cannot be EXOR type
    if !Fraig_NodeIsAnd(pNode) {
        return 0; }
    // inlining of macro Fraig_IsComplement
    if !Fraig_NodeIsAnd(pNode.p1) || !(((int32(((ABC_PTRUINT_T) (pNode.p1) & 01)) != 0))) {
        return 0; }
    // inlining of macro Fraig_IsComplement
    if !Fraig_NodeIsAnd(pNode.p2) || !(((int32(((ABC_PTRUINT_T) (pNode.p2) & 01)) != 0))) {
        return 0; }

    // get children
    // inlining of macro Fraig_Regular
    pNode1 = ((Fraig_Node_t *)((ABC_PTRUINT_T)(pNode.p1) & ~01)) ;
    // inlining of macro Fraig_Regular
    pNode2 = ((Fraig_Node_t *)((ABC_PTRUINT_T)(pNode.p2) & ~01)) ;
    assert(pNode1.Num < pNode2.Num, "pNode1.Num < pNode2.Num");

    // compare grandchildren
    // inlining of macro Fraig_Not
    return pNode1.p1 == ((Fraig_Node_t *)((ABC_PTRUINT_T)(pNode2.p1) ^ 01))  && pNode1.p2 == ((Fraig_Node_t *)((ABC_PTRUINT_T)(pNode2.p2) ^ 01)) ;
}

/**Function*************************************************************

  Synopsis    [Returns 1 if the node is the root of MUX or EXOR/NEXOR.]

  Description [The node can be complemented.]
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Fraig_NodeIsMuxType(pNode *Fraig_Node_t) int32 {
    var pNode1, pNode2 *Fraig_Node_t;

    // make the node regular (it does not matter for EXOR/NEXOR)
    // inlining of macro Fraig_Regular
    pNode = ((Fraig_Node_t *)((ABC_PTRUINT_T)(pNode) & ~01)) ;
    // if the node or its children are not ANDs or not compl, this cannot be EXOR type
    if !Fraig_NodeIsAnd(pNode) {
        return 0; }
    // inlining of macro Fraig_IsComplement
    if !Fraig_NodeIsAnd(pNode.p1) || !(((int32(((ABC_PTRUINT_T) (pNode.p1) & 01)) != 0))) {
        return 0; }
    // inlining of macro Fraig_IsComplement
    if !Fraig_NodeIsAnd(pNode.p2) || !(((int32(((ABC_PTRUINT_T) (pNode.p2) & 01)) != 0))) {
        return 0; }

    // get children
    // inlining of macro Fraig_Regular
    pNode1 = ((Fraig_Node_t *)((ABC_PTRUINT_T)(pNode.p1) & ~01)) ;
    // inlining of macro Fraig_Regular
    pNode2 = ((Fraig_Node_t *)((ABC_PTRUINT_T)(pNode.p2) & ~01)) ;
    assert(pNode1.Num < pNode2.Num, "pNode1.Num < pNode2.Num");

    // compare grandchildren
    // node is an EXOR/NEXOR
    // inlining of macro Fraig_Not
    if pNode1.p1 == ((Fraig_Node_t *)((ABC_PTRUINT_T)(pNode2.p1) ^ 01))  && pNode1.p2 == ((Fraig_Node_t *)((ABC_PTRUINT_T)(pNode2.p2) ^ 01)) {
        return 1; } 

    // otherwise the node is MUX iff it has a pair of equal grandchildren
    // inlining of macro Fraig_Not
    return pNode1.p1 == ((Fraig_Node_t *)((ABC_PTRUINT_T)(pNode2.p1) ^ 01))  || 
           // inlining of macro Fraig_Not
           pNode1.p1 == ((Fraig_Node_t *)((ABC_PTRUINT_T)(pNode2.p2) ^ 01))  ||
           // inlining of macro Fraig_Not
           pNode1.p2 == ((Fraig_Node_t *)((ABC_PTRUINT_T)(pNode2.p1) ^ 01))  ||
           // inlining of macro Fraig_Not
           pNode1.p2 == ((Fraig_Node_t *)((ABC_PTRUINT_T)(pNode2.p2) ^ 01)) ;
}

/**Function*************************************************************

  Synopsis    [Returns 1 if the node is EXOR, 0 if it is NEXOR.]

  Description [The node should be EXOR type and not complemented.]
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Fraig_NodeIsExor(pNode *Fraig_Node_t) int32 {
    var pNode1 *Fraig_Node_t;
    // inlining of macro Fraig_IsComplement
    assert(!(((int32(((ABC_PTRUINT_T) (pNode) & 01)) != 0))), "!/*__starting_point_inline_macro:Fraig_IsComplement__*/((int32(((ABC_PTRUINT_T) (pNode) & 01))))");
    assert((Fraig_NodeIsExorType(pNode) != 0), "Fraig_NodeIsExorType(pNode)");
    // inlining of macro Fraig_IsComplement
    assert((((int32(((ABC_PTRUINT_T) (pNode.p1) & 01)) != 0))), "/*__starting_point_inline_macro:Fraig_IsComplement__*/((int32(((ABC_PTRUINT_T) (pNode.p1) & 01))))");
    // get children
    // inlining of macro Fraig_Regular
    pNode1 = ((Fraig_Node_t *)((ABC_PTRUINT_T)(pNode.p1) & ~01)) ;
    // inlining of macro Fraig_IsComplement
    return ((int32(((ABC_PTRUINT_T) (pNode1.p1) & 01)))) == ((int32(((ABC_PTRUINT_T) (pNode1.p2) & 01))));
}

/**Function*************************************************************

  Synopsis    [Recognizes what nodes are control and data inputs of a MUX.]

  Description [If the node is a MUX, returns the control variable C.
  Assigns nodes T and E to be the then and else variables of the MUX. 
  Node C is never complemented. Nodes T and E can be complemented.
  This function also recognizes EXOR/NEXOR gates as MUXes.]
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Fraig_NodeRecognizeMux(pNode *Fraig_Node_t, ppNodeT **Fraig_Node_t, ppNodeE **Fraig_Node_t) *Fraig_Node_t {
    var pNode1, pNode2 *Fraig_Node_t;
    // inlining of macro Fraig_IsComplement
    assert(!(((int32(((ABC_PTRUINT_T) (pNode) & 01)) != 0))), "!/*__starting_point_inline_macro:Fraig_IsComplement__*/((int32(((ABC_PTRUINT_T) (pNode) & 01))))");
    assert((Fraig_NodeIsMuxType(pNode) != 0), "Fraig_NodeIsMuxType(pNode)");
    // get children
    // inlining of macro Fraig_Regular
    pNode1 = ((Fraig_Node_t *)((ABC_PTRUINT_T)(pNode.p1) & ~01)) ;
    // inlining of macro Fraig_Regular
    pNode2 = ((Fraig_Node_t *)((ABC_PTRUINT_T)(pNode.p2) & ~01)) ;
    // find the control variable
    // inlining of macro Fraig_Not
    if pNode1.p1 == ((Fraig_Node_t *)((ABC_PTRUINT_T)(pNode2.p1) ^ 01)) {
        // inlining of macro Fraig_IsComplement
        if  (int32(((ABC_PTRUINT_T) (pNode1.p1) & 01)) != 0) { // pNode2->p1 is positive phase of C
            // inlining of macro Fraig_Not
            *ppNodeT = ((Fraig_Node_t *)((ABC_PTRUINT_T)(pNode2.p2) ^ 01)) ;
            // inlining of macro Fraig_Not
            *ppNodeE = ((Fraig_Node_t *)((ABC_PTRUINT_T)(pNode1.p2) ^ 01)) ;
            return pNode2.p1;
        
        } else { // pNode1->p1 is positive phase of C
            // inlining of macro Fraig_Not
            *ppNodeT = ((Fraig_Node_t *)((ABC_PTRUINT_T)(pNode1.p2) ^ 01)) ;
            // inlining of macro Fraig_Not
            *ppNodeE = ((Fraig_Node_t *)((ABC_PTRUINT_T)(pNode2.p2) ^ 01)) ;
            return pNode1.p1;
        }
    
    // inlining of macro Fraig_Not
    } else if pNode1.p1 == ((Fraig_Node_t *)((ABC_PTRUINT_T)(pNode2.p2) ^ 01)) {
        // inlining of macro Fraig_IsComplement
        if  (int32(((ABC_PTRUINT_T) (pNode1.p1) & 01)) != 0) { // pNode2->p2 is positive phase of C
            // inlining of macro Fraig_Not
            *ppNodeT = ((Fraig_Node_t *)((ABC_PTRUINT_T)(pNode2.p1) ^ 01)) ;
            // inlining of macro Fraig_Not
            *ppNodeE = ((Fraig_Node_t *)((ABC_PTRUINT_T)(pNode1.p2) ^ 01)) ;
            return pNode2.p2;
        
        } else { // pNode1->p1 is positive phase of C
            // inlining of macro Fraig_Not
            *ppNodeT = ((Fraig_Node_t *)((ABC_PTRUINT_T)(pNode1.p2) ^ 01)) ;
            // inlining of macro Fraig_Not
            *ppNodeE = ((Fraig_Node_t *)((ABC_PTRUINT_T)(pNode2.p1) ^ 01)) ;
            return pNode1.p1;
        }
    
    // inlining of macro Fraig_Not
    } else if pNode1.p2 == ((Fraig_Node_t *)((ABC_PTRUINT_T)(pNode2.p1) ^ 01)) {
        // inlining of macro Fraig_IsComplement
        if  (int32(((ABC_PTRUINT_T) (pNode1.p2) & 01)) != 0) { // pNode2->p1 is positive phase of C
            // inlining of macro Fraig_Not
            *ppNodeT = ((Fraig_Node_t *)((ABC_PTRUINT_T)(pNode2.p2) ^ 01)) ;
            // inlining of macro Fraig_Not
            *ppNodeE = ((Fraig_Node_t *)((ABC_PTRUINT_T)(pNode1.p1) ^ 01)) ;
            return pNode2.p1;
        
        } else { // pNode1->p2 is positive phase of C
            // inlining of macro Fraig_Not
            *ppNodeT = ((Fraig_Node_t *)((ABC_PTRUINT_T)(pNode1.p1) ^ 01)) ;
            // inlining of macro Fraig_Not
            *ppNodeE = ((Fraig_Node_t *)((ABC_PTRUINT_T)(pNode2.p2) ^ 01)) ;
            return pNode1.p2;
        }
    
    // inlining of macro Fraig_Not
    } else if pNode1.p2 == ((Fraig_Node_t *)((ABC_PTRUINT_T)(pNode2.p2) ^ 01)) {
        // inlining of macro Fraig_IsComplement
        if  (int32(((ABC_PTRUINT_T) (pNode1.p2) & 01)) != 0) { // pNode2->p2 is positive phase of C
            // inlining of macro Fraig_Not
            *ppNodeT = ((Fraig_Node_t *)((ABC_PTRUINT_T)(pNode2.p1) ^ 01)) ;
            // inlining of macro Fraig_Not
            *ppNodeE = ((Fraig_Node_t *)((ABC_PTRUINT_T)(pNode1.p1) ^ 01)) ;
            return pNode2.p2;
        
        } else { // pNode1->p2 is positive phase of C
            // inlining of macro Fraig_Not
            *ppNodeT = ((Fraig_Node_t *)((ABC_PTRUINT_T)(pNode1.p1) ^ 01)) ;
            // inlining of macro Fraig_Not
            *ppNodeE = ((Fraig_Node_t *)((ABC_PTRUINT_T)(pNode2.p1) ^ 01)) ;
            return pNode1.p2;
        }
    }
    assert(false, "0"); // this is not MUX
    return nil;
}

/**Function*************************************************************

  Synopsis    [Counts the number of EXOR type nodes.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Fraig_ManCountExors(pMan *Fraig_Man_t) int32 {
    var i, nExors int32;
    nExors = 0;
    for  i = 0; i < pMan.vNodes.nSize; i++  {
        nExors += Fraig_NodeIsExorType( pMan.vNodes.pArray[i] );
    }
    return nExors;

}

/**Function*************************************************************

  Synopsis    [Counts the number of EXOR type nodes.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Fraig_ManCountMuxes(pMan *Fraig_Man_t) int32 {
    var i, nMuxes int32;
    nMuxes = 0;
    for  i = 0; i < pMan.vNodes.nSize; i++  {
        nMuxes += Fraig_NodeIsMuxType( pMan.vNodes.pArray[i] );
    }
    return nMuxes;

}

/**Function*************************************************************

  Synopsis    [Returns 1 if siminfo of Node1 is contained in siminfo of Node2.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Fraig_NodeSimsContained(pMan *Fraig_Man_t, pNode1 *Fraig_Node_t, pNode2 *Fraig_Node_t) int32 {
    var pUnsigned1, pUnsigned2 *uint32;
    var i int32;

    // compare random siminfo
    pUnsigned1 = pNode1.puSimR;
    pUnsigned2 = pNode2.puSimR;
    for  i = 0; i < pMan.nWordsRand; i++  {
        if pUnsigned1[i] & ~pUnsigned2[i] {
            return 0; }
    }

    // compare systematic siminfo
    pUnsigned1 = pNode1.puSimD;
    pUnsigned2 = pNode2.puSimD;
    for  i = 0; i < pMan.iWordStart; i++  {
        if pUnsigned1[i] & ~pUnsigned2[i] {
            return 0; }
    }

    return 1;
}

/**Function*************************************************************

  Synopsis    [Count the number of PI variables.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Fraig_CountPis(p *Fraig_Man_t, vVarNums *Msat_IntVec_t) int32 {
    var pVars *int32;
    var nVars int32;
    var i int32;
    var Counter int32;

    nVars = Msat_IntVecReadSize(vVarNums);
    pVars = Msat_IntVecReadArray(vVarNums);
    Counter = 0;
    for  i = 0; i < nVars; i++  {
        Counter += Fraig_NodeIsVar( p.vNodes.pArray[pVars[i]] );
    }
    return Counter;
}



/**Function*************************************************************

  Synopsis    [Counts the number of EXOR type nodes.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Fraig_ManPrintRefs(pMan *Fraig_Man_t) int32 {
    var vPivots *Fraig_NodeVec_t;
    var pNode, pNode2 *Fraig_Node_t;
    var i, k, Counter, nProved int32;
    var clk abctime;

    vPivots = Fraig_NodeVecAlloc( 1000 );
    for  i = 0; i < pMan.vNodes.nSize; i++  {
        pNode = pMan.vNodes.pArray[i];

        if pNode.nOnes == 0 || pNode.nOnes == uint32(pMan.nWordsRand) * 32 {
            continue; }

        if pNode.nRefs > 5 {
            Fraig_NodeVecPush( vPivots, pNode );
//            printf( "Node %6d : nRefs = %2d   Level = %3d.\n", pNode->Num, pNode->nRefs, pNode->Level );
        }
    }
    fmt.Printf("Total nodes = %d. Referenced nodes = %d.\n", pMan.vNodes.nSize, vPivots.nSize);

clk = Abc_Clock();
    // count implications
    Counter = nProved = 0;
    for  i = 0; i < vPivots.nSize; i++  {
        for  k = i+1; k < vPivots.nSize; k++  {
            pNode  = vPivots.pArray[i];
            pNode2 = vPivots.pArray[k];
            if Fraig_NodeSimsContained( pMan, pNode, pNode2 ) != 0 {
                if Fraig_NodeIsImplication( pMan, pNode, pNode2, -1 ) {
                    nProved++; }
                Counter++;
            
            } else if Fraig_NodeSimsContained( pMan, pNode2, pNode ) != 0 {
                if Fraig_NodeIsImplication( pMan, pNode2, pNode, -1 ) {
                    nProved++; }
                Counter++;
            }
        }
    }
    fmt.Printf("Number of candidate pairs = %d.  Proved = %d.\n", Counter, nProved);
//ABC_PRT( "Time", Abc_Clock() - clk );
    return 0;
}


/**Function*************************************************************

  Synopsis    [Checks if pNew exists among the implication fanins of pOld.]

  Description [If pNew is an implication fanin of pOld, returns 1. 
  If Fraig_Not(pNew) is an implication fanin of pOld, return -1.
  Otherwise returns 0.]
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Fraig_NodeIsInSupergate(pOld *Fraig_Node_t, pNew *Fraig_Node_t) int32 {
    var RetValue1, RetValue2 int32;
    // inlining of macro Fraig_Regular
    if  ((Fraig_Node_t *)((ABC_PTRUINT_T)(pOld) & ~01))  == ((Fraig_Node_t *)((ABC_PTRUINT_T)(pNew) & ~01)) {
        if pOld == pNew {
            return 1;
        } else {
            return -1;
        } }
    // inlining of macro Fraig_IsComplement
    if  (((int32(((ABC_PTRUINT_T) (pOld) & 01)) != 0))) || Fraig_NodeIsVar(pOld) {
        return 0; }
    RetValue1 = Fraig_NodeIsInSupergate( pOld.p1, pNew );
    RetValue2 = Fraig_NodeIsInSupergate( pOld.p2, pNew );
    if RetValue1 == -1 || RetValue2 == -1 {
        return -1; }
    if RetValue1 ==  1 || RetValue2 ==  1 {
        return 1; }
    return 0;
}


/**Function*************************************************************

  Synopsis    [Returns the array of nodes to be combined into one multi-input AND-gate.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Fraig_CollectSupergate_rec(pNode *Fraig_Node_t, vSuper *Fraig_NodeVec_t, fFirst int32, fStopAtMux int32) {
    // if the new node is complemented or a PI, another gate begins
//    if ( Fraig_IsComplement(pNode) || Fraig_NodeIsVar(pNode) || Fraig_NodeIsMuxType(pNode) )
    // inlining of macro Fraig_Regular
    if (!fFirst && ((Fraig_Node_t *)((ABC_PTRUINT_T)(pNode) & ~01)) .nRefs > 1) || 
          // inlining of macro Fraig_IsComplement
          ((int32(((ABC_PTRUINT_T) (pNode) & 01)))) || Fraig_NodeIsVar(pNode) || 
          (fStopAtMux && Fraig_NodeIsMuxType(pNode)) {
        Fraig_NodeVecPushUnique( vSuper, pNode );
        return;
    }
    // go through the branches
    Fraig_CollectSupergate_rec( pNode.p1, vSuper, 0, fStopAtMux );
    Fraig_CollectSupergate_rec( pNode.p2, vSuper, 0, fStopAtMux );
}

/**Function*************************************************************

  Synopsis    [Returns the array of nodes to be combined into one multi-input AND-gate.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Fraig_CollectSupergate(pNode *Fraig_Node_t, fStopAtMux int32) *Fraig_NodeVec_t {
    var vSuper *Fraig_NodeVec_t;
    vSuper = Fraig_NodeVecAlloc( 8 );
    Fraig_CollectSupergate_rec( pNode, vSuper, 1, fStopAtMux );
    return vSuper;
}


/**Function*************************************************************

  Synopsis    []

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Fraig_ManIncrementTravId(pMan *Fraig_Man_t) {
    pMan.nTravIds2++;
}

/**Function*************************************************************

  Synopsis    []

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Fraig_NodeSetTravIdCurrent(pMan *Fraig_Man_t, pNode *Fraig_Node_t) {
    pNode.TravId2 = pMan.nTravIds2;
}

/**Function*************************************************************

  Synopsis    []

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Fraig_NodeIsTravIdCurrent(pMan *Fraig_Man_t, pNode *Fraig_Node_t) int32 {
    return pNode.TravId2 == pMan.nTravIds2;
}

/**Function*************************************************************

  Synopsis    []

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Fraig_NodeIsTravIdPrevious(pMan *Fraig_Man_t, pNode *Fraig_Node_t) int32 {
    return pNode.TravId2 == pMan.nTravIds2 - 1;
}

////////////////////////////////////////////////////////////////////////
///                       END OF FILE                                ///
////////////////////////////////////////////////////////////////////////


//ABC_NAMESPACE_IMPL_END

