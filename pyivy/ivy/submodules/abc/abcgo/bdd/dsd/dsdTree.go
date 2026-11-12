package dsd

import "fmt"

import "os"

/**CFile****************************************************************

  FileName    [dsdTree.c]

  PackageName [DSD: Disjoint-support decomposition package.]

  Synopsis    [Managing the decomposition tree.]

  Author      [Alan Mishchenko]
  
  Affiliation [UC Berkeley]

  Date        [Ver. 8.0. Started - September 22, 2003.]

  Revision    [$Id: dsdTree.c,v 1.0 2002/22/09 00:00:00 alanmi Exp $]

***********************************************************************/

//#include "dsdInt.h"
//#include "misc/util/utilTruth.h"
//#include "opt/dau/dau.h"

//ABC_NAMESPACE_IMPL_START


////////////////////////////////////////////////////////////////////////
///                    FUNCTION DECLARATIONS                         ///
////////////////////////////////////////////////////////////////////////

// static void Dsd_TreeUnmark_rec( Dsd_Node_t * pNode );

// static void Dsd_TreeGetInfo_rec( Dsd_Node_t * pNode, int RankCur );

// static int  Dsd_TreeCountNonTerminalNodes_rec( Dsd_Node_t * pNode );

// static int  Dsd_TreeCountPrimeNodes_rec( Dsd_Node_t * pNode );

// static int  Dsd_TreeCollectDecomposableVars_rec( DdManager * dd, Dsd_Node_t * pNode, int * pVars, int * nVars );

// static void Dsd_TreeCollectNodesDfs_rec( Dsd_Node_t * pNode, Dsd_Node_t * ppNodes[], int * pnNodes );

// static void Dsd_TreePrint_rec( FILE * pFile, Dsd_Node_t * pNode, int fCcmp, char * pInputNames[], char * pOutputName, int nOffset, int * pSigCounter, int fShortNames );

// static void Dsd_NodePrint_rec( FILE * pFile, Dsd_Node_t * pNode, int fComp, char * pOutputName, int nOffset, int * pSigCounter );


////////////////////////////////////////////////////////////////////////
///                      STATIC VARIABLES                            ///
////////////////////////////////////////////////////////////////////////

var s_DepthMax int32;
var s_GateSizeMax int32;

////////////////////////////////////////////////////////////////////////
///                     FUNCTION DEFINITIONS                         ///
////////////////////////////////////////////////////////////////////////

/**Function*************************************************************

  Synopsis    [Create the DSD node.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Dsd_TreeNodeCreate(Type int32, nDecs int32, BlockNum int32) *Dsd_Node_t {
    // allocate memory for this node 
    var p *Dsd_Node_t  = (Dsd_Node_t *) ABC_ALLOC( char, sizeof(Dsd_Node_t) );
    memset( p, 0, sizeof(Dsd_Node_t) );
    p.Type       = (Dsd_Type_t)Type;       // the type of this block
    p.nDecs      = nDecs;                  // the number of decompositions
    if p.nDecs {
        p.pDecs      = (Dsd_Node_t **) ABC_ALLOC( char, p.nDecs * sizeof(Dsd_Node_t *) );
        p.pDecs[0]   = nil;
    }
    return p;
}

/**Function*************************************************************

  Synopsis    [Frees the DSD node.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Dsd_TreeNodeDelete(dd *DdManager, pNode *Dsd_Node_t) {
    if pNode.G {  Cudd_RecursiveDeref( dd, pNode.G ); }
    if pNode.S {  Cudd_RecursiveDeref( dd, pNode.S ); }
    ABC_FREE( pNode.pDecs );
    ABC_FREE( pNode );
}

/**Function*************************************************************

  Synopsis    [Unmarks the decomposition tree.]

  Description [This function assumes that originally pNode->nVisits are 
  set to zero!]
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Dsd_TreeUnmark(pDsdMan *Dsd_Manager_t) {
    var i int32;
    for  i = 0; i < pDsdMan.nRoots; i++  {
        // inlining of macro Dsd_Regular
        Dsd_TreeUnmark_rec( ((Dsd_Node_t *)((ABC_PTRUINT_T)(pDsdMan.pRoots[i]) & ~01))  );
    }
}


/**Function*************************************************************

  Synopsis    [Recursive unmarking.]

  Description [This function should be called with a non-complemented 
  pointer.]
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Dsd_TreeUnmark_rec(pNode *Dsd_Node_t) {
    var i int32;

    assert((pNode != nil), "pNode");
    // inlining of macro Dsd_IsComplement
    assert(!(((int32(((ABC_PTRUINT_T) (pNode) & 01)) != 0))), "!/*__starting_point_inline_macro:Dsd_IsComplement__*/((int32(((ABC_PTRUINT_T) (pNode) & 01))))");
    assert(pNode.nVisits > 0, "pNode.nVisits > 0");

    if --pNode.nVisits { // if this is not the last visit, return
        return; }

    // upon the last visit, go through the list of successors and call recursively 
    if pNode.Type != DSD_NODE_BUF && pNode.Type != DSD_NODE_CONST1 {
    for  i = 0; i < pNode.nDecs; i++  {
        // inlining of macro Dsd_Regular
        Dsd_TreeUnmark_rec( ((Dsd_Node_t *)((ABC_PTRUINT_T)(pNode.pDecs[i]) & ~01))  );
    } }
}

/**Function*************************************************************

  Synopsis    [Getting information about the node.]

  Description [This function computes the max depth and the max gate size 
  of the tree rooted at the node.]
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Dsd_TreeNodeGetInfo(pDsdMan *Dsd_Manager_t, DepthMax *int32, GateSizeMax *int32) {
    var i int32;
    s_DepthMax    = 0;
    s_GateSizeMax = 0;

    for  i = 0; i < pDsdMan.nRoots; i++  {
        // inlining of macro Dsd_Regular
        Dsd_TreeGetInfo_rec( ((Dsd_Node_t *)((ABC_PTRUINT_T)(pDsdMan.pRoots[i]) & ~01)) , 0 );
    }

    if DepthMax != nil { 
        *DepthMax     = s_DepthMax; }
    if GateSizeMax != nil { 
        *GateSizeMax  = s_GateSizeMax; }
}

/**Function*************************************************************

  Synopsis    [Getting information about the node.]

  Description [This function computes the max depth and the max gate size 
  of the tree rooted at the node.]
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Dsd_TreeNodeGetInfoOne(pNode *Dsd_Node_t, DepthMax *int32, GateSizeMax *int32) {
    s_DepthMax    = 0;
    s_GateSizeMax = 0;

    // inlining of macro Dsd_Regular
    Dsd_TreeGetInfo_rec( ((Dsd_Node_t *)((ABC_PTRUINT_T)(pNode) & ~01)) , 0 );

    if DepthMax != nil { 
        *DepthMax     = s_DepthMax; }
    if GateSizeMax != nil { 
        *GateSizeMax  = s_GateSizeMax; }
}


/**Function*************************************************************

  Synopsis    [Performs the recursive step of Dsd_TreeNodeGetInfo().]

  Description [pNode is the node, for the tree rooted in which we are 
  determining info. RankCur is the current rank to assign to the node.
  fSetRank is the flag saying whether the rank will be written in the 
  node. s_DepthMax is the maximum depths of the tree. s_GateSizeMax is 
  the maximum gate size.]
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Dsd_TreeGetInfo_rec(pNode *Dsd_Node_t, RankCur int32) {
    var i int32;
    var GateSize int32;

    assert((pNode != nil), "pNode");
    // inlining of macro Dsd_IsComplement
    assert(!(((int32(((ABC_PTRUINT_T) (pNode) & 01)) != 0))), "!/*__starting_point_inline_macro:Dsd_IsComplement__*/((int32(((ABC_PTRUINT_T) (pNode) & 01))))");
    assert(pNode.nVisits >= 0, "pNode.nVisits >= 0");

    // we don't want the two-input gates to count for non-decomposable blocks
    if pNode.Type == DSD_NODE_OR  || 
         pNode.Type == DSD_NODE_EXOR {
        GateSize = 2;
    } else {
        GateSize = pNode.nDecs; }

    // update the max size of the node
    if s_GateSizeMax < GateSize {
         s_GateSizeMax = GateSize; }

    if pNode.nDecs < 2 {
        return; }

    // update the max rank
    if s_DepthMax < RankCur+1 {
         s_DepthMax = RankCur+1; }

    // call recursively
    for  i = 0; i < pNode.nDecs; i++  {
        // inlining of macro Dsd_Regular
        Dsd_TreeGetInfo_rec( ((Dsd_Node_t *)((ABC_PTRUINT_T)(pNode.pDecs[i]) & ~01)) , RankCur+1 );
    }
}

/**Function*************************************************************

  Synopsis    [Counts AIG nodes needed to implement this node.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Dsd_TreeGetAigCost_rec(pNode *Dsd_Node_t) int32 {
    var i, Counter int32;

    assert((pNode != nil), "pNode");
    // inlining of macro Dsd_IsComplement
    assert(!(((int32(((ABC_PTRUINT_T) (pNode) & 01)) != 0))), "!/*__starting_point_inline_macro:Dsd_IsComplement__*/((int32(((ABC_PTRUINT_T) (pNode) & 01))))");
    assert(pNode.nVisits >= 0, "pNode.nVisits >= 0");

    if pNode.nDecs < 2 {
        return 0; }

    // we don't want the two-input gates to count for non-decomposable blocks
    if pNode.Type == DSD_NODE_OR {
        Counter += pNode.nDecs - 1;
    } else if pNode.Type == DSD_NODE_EXOR {
        Counter += 3*(pNode.nDecs - 1);
    } else if pNode.Type == DSD_NODE_PRIME && pNode.nDecs == 3 {
        Counter += 3; }

    // call recursively
    for  i = 0; i < pNode.nDecs; i++  {
        // inlining of macro Dsd_Regular
        Counter += Dsd_TreeGetAigCost_rec( ((Dsd_Node_t *)((ABC_PTRUINT_T)(pNode.pDecs[i]) & ~01))  );
    }
    return Counter;
}

/**Function*************************************************************

  Synopsis    [Counts AIG nodes needed to implement this node.]

  Description [Assumes that the only primes of the DSD tree are MUXes.]
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Dsd_TreeGetAigCost(pNode *Dsd_Node_t) int32 {
    // inlining of macro Dsd_Regular
    return Dsd_TreeGetAigCost_rec( ((Dsd_Node_t *)((ABC_PTRUINT_T)(pNode) & ~01))  );
}

/**Function*************************************************************

  Synopsis    [Counts non-terminal nodes of the DSD tree.]

  Description [Nonterminal nodes include all the nodes with the
  support more than 1. These are OR, EXOR, and PRIME nodes. They
  do not include the elementary variable nodes and the constant 1
  node.]
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Dsd_TreeCountNonTerminalNodes(pDsdMan *Dsd_Manager_t) int32 {
    var Counter, i int32;
    Counter = 0;
    for  i = 0; i < pDsdMan.nRoots; i++  {
        // inlining of macro Dsd_Regular
        Counter += Dsd_TreeCountNonTerminalNodes_rec( ((Dsd_Node_t *)((ABC_PTRUINT_T)(pDsdMan.pRoots[i]) & ~01))  );
    }
    Dsd_TreeUnmark( pDsdMan );
    return Counter;
}

/**Function*************************************************************

  Synopsis    []

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Dsd_TreeCountNonTerminalNodesOne(pRoot *Dsd_Node_t) int32 {
    var Counter int32;

    // go through the list of successors and call recursively 
    // inlining of macro Dsd_Regular
    Counter = Dsd_TreeCountNonTerminalNodes_rec( ((Dsd_Node_t *)((ABC_PTRUINT_T)(pRoot) & ~01))  );

    // inlining of macro Dsd_Regular
    Dsd_TreeUnmark_rec( ((Dsd_Node_t *)((ABC_PTRUINT_T)(pRoot) & ~01))  );
    return Counter;
}


/**Function*************************************************************

  Synopsis    [Counts non-terminal nodes for one root.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Dsd_TreeCountNonTerminalNodes_rec(pNode *Dsd_Node_t) int32 {
    var i int32;
    var Counter int32;

    assert((pNode != nil), "pNode");
    // inlining of macro Dsd_IsComplement
    assert(!(((int32(((ABC_PTRUINT_T) (pNode) & 01)) != 0))), "!/*__starting_point_inline_macro:Dsd_IsComplement__*/((int32(((ABC_PTRUINT_T) (pNode) & 01))))");
    assert(pNode.nVisits >= 0, "pNode.nVisits >= 0");

    if pNode.nVisits++ { // if this is not the first visit, return zero
        return 0; }

    if pNode.nDecs <= 1 {
        return 0; }

    // upon the first visit, go through the list of successors and call recursively 
    for  i = 0; i < pNode.nDecs; i++  {
        // inlining of macro Dsd_Regular
        Counter += Dsd_TreeCountNonTerminalNodes_rec( ((Dsd_Node_t *)((ABC_PTRUINT_T)(pNode.pDecs[i]) & ~01))  );
    }

    return Counter + 1;
}


/**Function*************************************************************

  Synopsis    [Counts prime nodes of the DSD tree.]

  Description [Prime nodes are nodes with the support more than 2,
  that is not an OR or EXOR gate.]
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Dsd_TreeCountPrimeNodes(pDsdMan *Dsd_Manager_t) int32 {
    var Counter, i int32;
    Counter = 0;
    for  i = 0; i < pDsdMan.nRoots; i++  {
        // inlining of macro Dsd_Regular
        Counter += Dsd_TreeCountPrimeNodes_rec( ((Dsd_Node_t *)((ABC_PTRUINT_T)(pDsdMan.pRoots[i]) & ~01))  );
    }
    Dsd_TreeUnmark( pDsdMan );
    return Counter;
}

/**Function*************************************************************

  Synopsis    [Counts prime nodes for one root.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Dsd_TreeCountPrimeNodesOne(pRoot *Dsd_Node_t) int32 {
    var Counter int32;

    // go through the list of successors and call recursively 
    // inlining of macro Dsd_Regular
    Counter = Dsd_TreeCountPrimeNodes_rec( ((Dsd_Node_t *)((ABC_PTRUINT_T)(pRoot) & ~01))  );

    // inlining of macro Dsd_Regular
    Dsd_TreeUnmark_rec( ((Dsd_Node_t *)((ABC_PTRUINT_T)(pRoot) & ~01))  );
    return Counter;
}


/**Function*************************************************************

  Synopsis    []

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Dsd_TreeCountPrimeNodes_rec(pNode *Dsd_Node_t) int32 {
    var i int32;
    var Counter int32;

    assert((pNode != nil), "pNode");
    // inlining of macro Dsd_IsComplement
    assert(!(((int32(((ABC_PTRUINT_T) (pNode) & 01)) != 0))), "!/*__starting_point_inline_macro:Dsd_IsComplement__*/((int32(((ABC_PTRUINT_T) (pNode) & 01))))");
    assert(pNode.nVisits >= 0, "pNode.nVisits >= 0");

    if pNode.nVisits++ { // if this is not the first visit, return zero
        return 0; }

    if pNode.nDecs <= 1 {
        return 0; }

    // upon the first visit, go through the list of successors and call recursively 
    for  i = 0; i < pNode.nDecs; i++  {
        // inlining of macro Dsd_Regular
        Counter += Dsd_TreeCountPrimeNodes_rec( ((Dsd_Node_t *)((ABC_PTRUINT_T)(pNode.pDecs[i]) & ~01))  );
    }

    if pNode.Type == DSD_NODE_PRIME {
        Counter++; }

    return Counter;
}


/**Function*************************************************************

  Synopsis    [Collects the decomposable vars on the PI side.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Dsd_TreeCollectDecomposableVars(pDsdMan *Dsd_Manager_t, pVars *int32) int32 {
    var nVars int32;

    // set the vars collected to 0
    nVars = 0;
    // inlining of macro Dsd_Regular
    Dsd_TreeCollectDecomposableVars_rec( pDsdMan.dd, ((Dsd_Node_t *)((ABC_PTRUINT_T)(pDsdMan.pRoots[0]) & ~01)) , pVars, &nVars );
    // return the number of collected vars
    return nVars;
}

/**Function*************************************************************

  Synopsis    [Implements the recursive part of Dsd_TreeCollectDecomposableVars().]

  Description [Adds decomposable variables as they are found to pVars and increments 
  nVars. Returns 1 if a non-dec node with more than 4 inputs was encountered 
  in the processed subtree. Returns 0, otherwise. ]
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Dsd_TreeCollectDecomposableVars_rec(dd *DdManager, pNode *Dsd_Node_t, pVars *int32, nVars *int32) int32 {
    var fSkipThisNode, i int32;
    var pTemp *Dsd_Node_t;
    var fVerbose int32;

    assert((pNode != nil), "pNode");
    // inlining of macro Dsd_IsComplement
    assert(!(((int32(((ABC_PTRUINT_T) (pNode) & 01)) != 0))), "!/*__starting_point_inline_macro:Dsd_IsComplement__*/((int32(((ABC_PTRUINT_T) (pNode) & 01))))");

    if pNode.nDecs <= 1 {
        return 0; }

    // go through the list of successors and call recursively 
    fSkipThisNode = 0;
    for  i = 0; i < pNode.nDecs; i++  {
        // inlining of macro Dsd_Regular
        if Dsd_TreeCollectDecomposableVars_rec(dd, ((Dsd_Node_t *)((ABC_PTRUINT_T)(pNode.pDecs[i]) & ~01)) , pVars, nVars) {
            fSkipThisNode = 1; }
    }

    if !fSkipThisNode && (pNode.Type == DSD_NODE_OR || pNode.Type == DSD_NODE_EXOR || pNode.nDecs <= 4) {
if fVerbose != 0 {
fmt.Printf("Node of type <%d> (OR=6,EXOR=8,RAND=1): ", pNode.Type); }

        for  i = 0; i < pNode.nDecs; i++  {
            // inlining of macro Dsd_Regular
            pTemp = ((Dsd_Node_t *)((ABC_PTRUINT_T)(pNode.pDecs[i]) & ~01)) ;
            if pTemp.Type == DSD_NODE_BUF {
                if pVars != nil {
                    pVars[ (*nVars)++ ] = pTemp.S.index;
                } else {
                    (*nVars)++; }
                    
if fVerbose != 0 {
fmt.Printf("%d ", pTemp.S.index); }
            }
        }
if fVerbose != 0 {
fmt.Printf("\n"); }
    
    } else {
        fSkipThisNode = 1; }


    return fSkipThisNode;
}


/**Function*************************************************************

  Synopsis    [Creates the DFS ordered array of DSD nodes in the tree.]

  Description [The collected nodes do not include the terminal nodes
  and the constant 1 node. The array of nodes is returned. The number
  of entries in the array is returned in the variale pnNodes.]
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Dsd_TreeCollectNodesDfs(pDsdMan *Dsd_Manager_t, pnNodes *int32) **Dsd_Node_t {
    var ppNodes **Dsd_Node_t;
    var nNodes, nNodesAlloc int32;
    var i int32;

    nNodesAlloc = Dsd_TreeCountNonTerminalNodes(pDsdMan);
    nNodes  = 0;
    ppNodes = ABC_ALLOC( Dsd_Node_t *, nNodesAlloc );
    for  i = 0; i < pDsdMan.nRoots; i++  {
        // inlining of macro Dsd_Regular
        Dsd_TreeCollectNodesDfs_rec( ((Dsd_Node_t *)((ABC_PTRUINT_T)(pDsdMan.pRoots[i]) & ~01)) , ppNodes, &nNodes );
    }
    Dsd_TreeUnmark( pDsdMan );
    assert(nNodesAlloc == nNodes, "nNodesAlloc == nNodes");
    *pnNodes = nNodes;
    return ppNodes;
}

/**Function*************************************************************

  Synopsis    [Creates the DFS ordered array of DSD nodes in the tree.]

  Description [The collected nodes do not include the terminal nodes
  and the constant 1 node. The array of nodes is returned. The number
  of entries in the array is returned in the variale pnNodes.]
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Dsd_TreeCollectNodesDfsOne(pDsdMan *Dsd_Manager_t, pNode *Dsd_Node_t, pnNodes *int32) **Dsd_Node_t {
    var ppNodes **Dsd_Node_t;
    var nNodes, nNodesAlloc int32;
    nNodesAlloc = Dsd_TreeCountNonTerminalNodesOne(pNode);
    nNodes  = 0;
    ppNodes = ABC_ALLOC( Dsd_Node_t *, nNodesAlloc );
    // inlining of macro Dsd_Regular
    Dsd_TreeCollectNodesDfs_rec( ((Dsd_Node_t *)((ABC_PTRUINT_T)(pNode) & ~01)) , ppNodes, &nNodes );
    // inlining of macro Dsd_Regular
    Dsd_TreeUnmark_rec(((Dsd_Node_t *)((ABC_PTRUINT_T)(pNode) & ~01)) );
    assert(nNodesAlloc == nNodes, "nNodesAlloc == nNodes");
    *pnNodes = nNodes;
    return ppNodes;
}


/**Function*************************************************************

  Synopsis    []

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Dsd_TreeCollectNodesDfs_rec(pNode *Dsd_Node_t, ppNodes []*Dsd_Node_t, pnNodes *int32) {
    var i int32;
    assert((pNode != nil), "pNode");
    // inlining of macro Dsd_IsComplement
    assert(!(((int32(((ABC_PTRUINT_T) (pNode) & 01)) != 0))), "!/*__starting_point_inline_macro:Dsd_IsComplement__*/((int32(((ABC_PTRUINT_T) (pNode) & 01))))");
    assert(pNode.nVisits >= 0, "pNode.nVisits >= 0");

    if pNode.nVisits++ { // if this is not the first visit, return zero
        return; }
    if pNode.nDecs <= 1 {
        return; }

    // upon the first visit, go through the list of successors and call recursively 
    for  i = 0; i < pNode.nDecs; i++  {
        // inlining of macro Dsd_Regular
        Dsd_TreeCollectNodesDfs_rec( ((Dsd_Node_t *)((ABC_PTRUINT_T)(pNode.pDecs[i]) & ~01)) , ppNodes, pnNodes );
    }

    ppNodes[ (*pnNodes)++ ] = pNode;
}

/**Function*************************************************************

  Synopsis    [Prints the decompostion tree into file.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Dsd_TreePrint(pFile *os.File, pDsdMan *Dsd_Manager_t, pInputNames []string, pOutputNames []string, fShortNames int32, Output int32) {
    var pNode *Dsd_Node_t;
    var SigCounter int32;
    var i int32;
    SigCounter = 1;

    if Output == -1 {
        for  i = 0; i < pDsdMan.nRoots; i++  {
            // inlining of macro Dsd_Regular
            pNode = ((Dsd_Node_t *)((ABC_PTRUINT_T)(pDsdMan.pRoots[i]) & ~01)) ;
            Dsd_TreePrint_rec( pFile, pNode, (pNode != pDsdMan.pRoots[i]), pInputNames, pOutputNames[i], 0, &SigCounter, fShortNames );
        }
    
    } else {
        assert(Output >= 0 && Output < pDsdMan.nRoots, "Output >= 0 && Output < pDsdMan.nRoots");
        // inlining of macro Dsd_Regular
        pNode = ((Dsd_Node_t *)((ABC_PTRUINT_T)(pDsdMan.pRoots[Output]) & ~01)) ;
        Dsd_TreePrint_rec( pFile, pNode, (pNode != pDsdMan.pRoots[Output]), pInputNames, pOutputNames[Output], 0, &SigCounter, fShortNames );
    }
}

/**Function*************************************************************

  Synopsis    [Prints the decompostion tree into file.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Dsd_TreePrint_rec(pFile *os.File, pNode *Dsd_Node_t, fComp int32, pInputNames []string, pOutputName string, nOffset int32, pSigCounter *int32, fShortNames int32) {
    var Buffer string // was [100]byte in C

    var pInput *Dsd_Node_t;
    var pInputNums *int32;
    var fCompNew, i int32;

    assert(pNode.Type == DSD_NODE_BUF || pNode.Type == DSD_NODE_CONST1 || 
            pNode.Type == DSD_NODE_PRIME || pNode.Type == DSD_NODE_OR || pNode.Type == DSD_NODE_EXOR, "pNode.Type == DSD_NODE_BUF || pNode.Type == DSD_NODE_CONST1 || \n        pNode.Type == DSD_NODE_PRIME || pNode.Type == DSD_NODE_OR || pNode.Type == DSD_NODE_EXOR"); 

    Extra_PrintSymbols( pFile, ' ', nOffset, 0 );
    if !(fComp != 0) {
        fmt.Fprintf(pFile, "%s = ", pOutputName);
    } else {
        fmt.Fprintf(pFile, "NOT(%s) = ", pOutputName); }
    pInputNums = ABC_ALLOC( int, pNode.nDecs );
    if pNode.Type == DSD_NODE_CONST1 {
        fmt.Fprintf(pFile, " Constant 1.\n");
    
    } else if pNode.Type == DSD_NODE_BUF {
        if fShortNames != 0 {
            fmt.Fprintf(pFile, "%d", 'a' + pNode.S.index);
        } else {
            fmt.Fprintf(pFile, "%s", pInputNames[pNode.S.index]); }
        fmt.Fprintf(pFile, "\n");
    
    } else if pNode.Type == DSD_NODE_PRIME {
        // print the line
        fmt.Fprintf(pFile, "PRIME(");
        for  i = 0; i < pNode.nDecs; i++  {
            // inlining of macro Dsd_Regular
            pInput   = ((Dsd_Node_t *)((ABC_PTRUINT_T)(pNode.pDecs[i]) & ~01)) ;
            fCompNew = int32(( pInput != pNode.pDecs[i] ));
            if i != 0 {
                fmt.Fprintf(pFile, ","); }
            if fCompNew {
                fmt.Fprintf(pFile, " NOT(");
            } else {
                fmt.Fprintf(pFile, " "); }
            if pInput.Type == DSD_NODE_BUF {
                pInputNums[i] = 0;
                if fShortNames != 0 {
                    fmt.Fprintf(pFile, "%d", pInput.S.index);
                } else {
                    fmt.Fprintf(pFile, "%s", pInputNames[pInput.S.index]); }
            
            } else {
                pInputNums[i] = (*pSigCounter)++;
                fmt.Fprintf(pFile, "<%d>", pInputNums[i]);
            }
            if fCompNew {
                fmt.Fprintf(pFile, ")"); }
        }
        fmt.Fprintf(pFile, " )\n");
        // call recursively for the following blocks
        for  i = 0; i < pNode.nDecs; i++  {
            if pInputNums[i] {
                // inlining of macro Dsd_Regular
                pInput   = ((Dsd_Node_t *)((ABC_PTRUINT_T)(pNode.pDecs[i]) & ~01)) ;
                Buffer = fmt.Sprintf("<%d>", pInputNums[i]);
                // inlining of macro Dsd_Regular
                Dsd_TreePrint_rec( pFile, ((Dsd_Node_t *)((ABC_PTRUINT_T)(pNode.pDecs[i]) & ~01)) , 0, pInputNames, Buffer, nOffset + 6, pSigCounter, fShortNames );
            }
        }
    
    } else if pNode.Type == DSD_NODE_OR {
        // print the line
        fmt.Fprintf(pFile, "OR(");
        for  i = 0; i < pNode.nDecs; i++  {
            // inlining of macro Dsd_Regular
            pInput = ((Dsd_Node_t *)((ABC_PTRUINT_T)(pNode.pDecs[i]) & ~01)) ;
            fCompNew  = int32(( pInput != pNode.pDecs[i] ));
            if i != 0 {
                fmt.Fprintf(pFile, ","); }
            if fCompNew {
                fmt.Fprintf(pFile, " NOT(");
            } else {
                fmt.Fprintf(pFile, " "); }
            if pInput.Type == DSD_NODE_BUF {
                pInputNums[i] = 0;
                if fShortNames != 0 {
                    fmt.Fprintf(pFile, "%c", 'a' + pInput.S.index);
                } else {
                    fmt.Fprintf(pFile, "%s", pInputNames[pInput.S.index]); }
            
            } else {
                pInputNums[i] = (*pSigCounter)++;
                fmt.Fprintf(pFile, "<%d>", pInputNums[i]);
            }
            if fCompNew {
                fmt.Fprintf(pFile, ")"); }
        }
        fmt.Fprintf(pFile, " )\n");
        // call recursively for the following blocks
        for  i = 0; i < pNode.nDecs; i++  {
            if pInputNums[i] {
                // inlining of macro Dsd_Regular
                pInput = ((Dsd_Node_t *)((ABC_PTRUINT_T)(pNode.pDecs[i]) & ~01)) ;
                Buffer = fmt.Sprintf("<%d>", pInputNums[i]);
                // inlining of macro Dsd_Regular
                Dsd_TreePrint_rec( pFile, ((Dsd_Node_t *)((ABC_PTRUINT_T)(pNode.pDecs[i]) & ~01)) , 0, pInputNames, Buffer, nOffset + 6, pSigCounter, fShortNames );
            }
        }
    
    } else if pNode.Type == DSD_NODE_EXOR {
        // print the line
        fmt.Fprintf(pFile, "EXOR(");
        for  i = 0; i < pNode.nDecs; i++  {
            // inlining of macro Dsd_Regular
            pInput = ((Dsd_Node_t *)((ABC_PTRUINT_T)(pNode.pDecs[i]) & ~01)) ;
            fCompNew  = int32(( pInput != pNode.pDecs[i] ));
            if i != 0 {
                fmt.Fprintf(pFile, ","); }
            if fCompNew {
                fmt.Fprintf(pFile, " NOT(");
            } else {
                fmt.Fprintf(pFile, " "); }
            if pInput.Type == DSD_NODE_BUF {
                pInputNums[i] = 0;
                if fShortNames != 0 {
                    fmt.Fprintf(pFile, "%c", 'a' + pInput.S.index);
                } else {
                    fmt.Fprintf(pFile, "%s", pInputNames[pInput.S.index]); }
            
            } else {
                pInputNums[i] = (*pSigCounter)++;
                fmt.Fprintf(pFile, "<%d>", pInputNums[i]);
            }
            if fCompNew {
                fmt.Fprintf(pFile, ")"); }
        }
        fmt.Fprintf(pFile, " )\n");
        // call recursively for the following blocks
        for  i = 0; i < pNode.nDecs; i++  {
            if pInputNums[i] {
                // inlining of macro Dsd_Regular
                pInput = ((Dsd_Node_t *)((ABC_PTRUINT_T)(pNode.pDecs[i]) & ~01)) ;
                Buffer = fmt.Sprintf("<%d>", pInputNums[i]);
                // inlining of macro Dsd_Regular
                Dsd_TreePrint_rec( pFile, ((Dsd_Node_t *)((ABC_PTRUINT_T)(pNode.pDecs[i]) & ~01)) , 0, pInputNames, Buffer, nOffset + 6, pSigCounter, fShortNames );
            }
        }
    }
    ABC_FREE( pInputNums );
}

/**Function*************************************************************

  Synopsis    [Prints the decompostion tree into file.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Dsd_TreeFunc2Truth_rec(dd *DdManager, bFunc *DdNode) word {
    var Cof0, Cof1 word;
    var Level int32;
    if bFunc == b0 {
        return 0; }
    if bFunc == b1 {
        return ~(word)0; }
    // inlining of macro Cudd_IsComplement
    if  int32( ((ptrint) (bFunc) & 01)) != 0 {
        // inlining of macro Cudd_Not
        return ~Dsd_TreeFunc2Truth_rec( dd, ((DdNode *)((ptrint)(bFunc) ^ 01)) ); }
    Level = dd.perm[bFunc.index];
    assert(Level >= 0 && Level < 6, "Level >= 0 && Level < 6");
    // inlining of macro cuddE
    Cof0 = Dsd_TreeFunc2Truth_rec( dd, ((bFunc).type.kids.E) );
    // inlining of macro cuddT
    Cof1 = Dsd_TreeFunc2Truth_rec( dd, ((bFunc).type.kids.T) );
    return (s_Truths6[Level] & Cof1) | (~s_Truths6[Level] & Cof0);
}
func Dsd_TreePrint2_rec(pFile *os.File, dd *DdManager, pNode *Dsd_Node_t, fComp int32, pInputNames []string) {
    var i int32;
    if pNode.Type == DSD_NODE_CONST1 {
        fmt.Fprintf(pFile, "Const%d", !fComp);
        return;
    }
    assert(pNode.Type == DSD_NODE_BUF || pNode.Type == DSD_NODE_PRIME || pNode.Type == DSD_NODE_OR || pNode.Type == DSD_NODE_EXOR, "pNode.Type == DSD_NODE_BUF || pNode.Type == DSD_NODE_PRIME || pNode.Type == DSD_NODE_OR || pNode.Type == DSD_NODE_EXOR"); 
//    fprintf( pFile, "%s", (fComp ^ (pNode->Type == DSD_NODE_OR))? "!" : "" );
    if pNode.Type == DSD_NODE_BUF {
        fmt.Fprintf(pFile, "%s", fComp? "!" : "");
        fmt.Fprintf(pFile, "%s", pInputNames[pNode.S.index]);
    
    } else if pNode.Type == DSD_NODE_PRIME {
        fmt.Fprintf(pFile, " ");
        if pNode.nDecs <= 6 {
            var pCanonPerm [6]byte; var uCanonPhase int32;
            // compute truth table
            var bFunc *DdNode  = Dsd_TreeGetPrimeFunction( dd, pNode );  
            var uTruth word  = Dsd_TreeFunc2Truth_rec( dd, bFunc );
            Cudd_Ref( bFunc );
            Cudd_RecursiveDeref( dd, bFunc );
            // canonicize truth table
            uCanonPhase = Abc_TtCanonicize( &uTruth, pNode.nDecs, pCanonPerm );
            fmt.Fprintf(pFile, "%s", (fComp ^ ((uCanonPhase >> pNode.nDecs) & 1)) ? "!" : "");
            Abc_TtPrintHexRev( pFile, &uTruth, pNode.nDecs );
            fmt.Fprintf(pFile, "{");
            for  i = 0; i < pNode.nDecs; i++  {
                var pInput *Dsd_Node_t  = pNode.pDecs[int32(pCanonPerm[i])];
                // inlining of macro Dsd_Regular
                // inlining of macro Dsd_IsComplement
                Dsd_TreePrint2_rec( pFile, dd, ((Dsd_Node_t *)((ABC_PTRUINT_T)(pInput) & ~01)) , ((int32(((ABC_PTRUINT_T) (pInput) & 01)))) ^ ((uCanonPhase>>i)&1), pInputNames );
            }
            fmt.Fprintf(pFile, "} ");
        
        } else {
            fmt.Fprintf(pFile, "|%d|", pNode.nDecs);
            fmt.Fprintf(pFile, "{");
            for  i = 0; i < pNode.nDecs; i++  {
                // inlining of macro Dsd_Regular
                // inlining of macro Dsd_IsComplement
                Dsd_TreePrint2_rec( pFile, dd, ((Dsd_Node_t *)((ABC_PTRUINT_T)(pNode.pDecs[i]) & ~01)) , ((int32(((ABC_PTRUINT_T) (pNode.pDecs[i]) & 01)))), pInputNames );
            }
            fmt.Fprintf(pFile, "} ");
        }
    
    } else if pNode.Type == DSD_NODE_OR {
        fmt.Fprintf(pFile, "%s", !fComp? "!" : "");
        fmt.Fprintf(pFile, "(");
        for  i = 0; i < pNode.nDecs; i++  {
            // inlining of macro Dsd_Regular
            // inlining of macro Dsd_IsComplement
            Dsd_TreePrint2_rec( pFile, dd, ((Dsd_Node_t *)((ABC_PTRUINT_T)(pNode.pDecs[i]) & ~01)) , !((int32(((ABC_PTRUINT_T) (pNode.pDecs[i]) & 01)))), pInputNames );
        }
        fmt.Fprintf(pFile, ")");
    
    } else if pNode.Type == DSD_NODE_EXOR {
        fmt.Fprintf(pFile, "%s", fComp? "!" : "");
        fmt.Fprintf(pFile, "[");
        for  i = 0; i < pNode.nDecs; i++  {
            // inlining of macro Dsd_Regular
            // inlining of macro Dsd_IsComplement
            Dsd_TreePrint2_rec( pFile, dd, ((Dsd_Node_t *)((ABC_PTRUINT_T)(pNode.pDecs[i]) & ~01)) , ((int32(((ABC_PTRUINT_T) (pNode.pDecs[i]) & 01)))), pInputNames );
        }
        fmt.Fprintf(pFile, "]");
    }
}
func Dsd_TreePrint2(pFile *os.File, pDsdMan *Dsd_Manager_t, pInputNames []string, pOutputNames []string, Output int32) {
    if Output == -1 {
        var i int32;
        for  i = 0; i < pDsdMan.nRoots; i++  {
            fmt.Fprintf(pFile, "%8s = ", pOutputNames[i]);
            // inlining of macro Dsd_Regular
            // inlining of macro Dsd_IsComplement
            Dsd_TreePrint2_rec( pFile, pDsdMan.dd, ((Dsd_Node_t *)((ABC_PTRUINT_T)(pDsdMan.pRoots[i]) & ~01)) , ((int32(((ABC_PTRUINT_T) (pDsdMan.pRoots[i]) & 01)))), pInputNames );
            fmt.Fprintf(pFile, "\n");
        }
    
    } else {
        assert(Output >= 0 && Output < pDsdMan.nRoots, "Output >= 0 && Output < pDsdMan.nRoots");
        fmt.Fprintf(pFile, "%8s = ", pOutputNames[Output]);
        // inlining of macro Dsd_Regular
        // inlining of macro Dsd_IsComplement
        Dsd_TreePrint2_rec( pFile, pDsdMan.dd, ((Dsd_Node_t *)((ABC_PTRUINT_T)(pDsdMan.pRoots[Output]) & ~01)) , ((int32(((ABC_PTRUINT_T) (pDsdMan.pRoots[Output]) & 01)))), pInputNames );
        fmt.Fprintf(pFile, "\n");
    }
}

/**Function*************************************************************

  Synopsis    [Prints the decompostion tree into file.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Dsd_NodePrint(pFile *os.File, pNode *Dsd_Node_t) {
    var pNodeR *Dsd_Node_t;
    var SigCounter int32  = 1;
    // inlining of macro Dsd_Regular
    pNodeR = ((Dsd_Node_t *)((ABC_PTRUINT_T)(pNode) & ~01)) ;
    Dsd_NodePrint_rec( pFile, pNodeR, pNodeR != pNode, "F", 0, &SigCounter );
}

/**Function*************************************************************

  Synopsis    [Prints one node of the decomposition tree.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Dsd_NodePrint_rec(pFile *os.File, pNode *Dsd_Node_t, fComp int32, pOutputName string, nOffset int32, pSigCounter *int32) {
    var Buffer string // was [100]byte in C

    var pInput *Dsd_Node_t;
    var pInputNums *int32;
    var fCompNew, i int32;

    assert(pNode.Type == DSD_NODE_BUF || pNode.Type == DSD_NODE_CONST1 || 
            pNode.Type == DSD_NODE_PRIME || pNode.Type == DSD_NODE_OR || pNode.Type == DSD_NODE_EXOR, "pNode.Type == DSD_NODE_BUF || pNode.Type == DSD_NODE_CONST1 || \n        pNode.Type == DSD_NODE_PRIME || pNode.Type == DSD_NODE_OR || pNode.Type == DSD_NODE_EXOR"); 

    Extra_PrintSymbols( pFile, ' ', nOffset, 0 );
    if !(fComp != 0) {
        fmt.Fprintf(pFile, "%s = ", pOutputName);
    } else {
        fmt.Fprintf(pFile, "NOT(%s) = ", pOutputName); }
    pInputNums = ABC_ALLOC( int, pNode.nDecs );
    if pNode.Type == DSD_NODE_CONST1 {
        fmt.Fprintf(pFile, " Constant 1.\n");
    
    } else if pNode.Type == DSD_NODE_BUF {
        fmt.Fprintf(pFile, " ");
        fmt.Fprintf(pFile, "%c", 'a' + pNode.S.index);
        fmt.Fprintf(pFile, "\n");
    
    } else if pNode.Type == DSD_NODE_PRIME {
        // print the line
        fmt.Fprintf(pFile, "PRIME(");
        for  i = 0; i < pNode.nDecs; i++  {
            // inlining of macro Dsd_Regular
            pInput   = ((Dsd_Node_t *)((ABC_PTRUINT_T)(pNode.pDecs[i]) & ~01)) ;
            fCompNew = int32(( pInput != pNode.pDecs[i] ));
            assert(fCompNew == 0, "fCompNew == 0");
            if i != 0 {
                fmt.Fprintf(pFile, ","); }
            if pInput.Type == DSD_NODE_BUF {
                pInputNums[i] = 0;
                fmt.Fprintf(pFile, " %c", 'a' + pInput.S.index);
            
            } else {
                pInputNums[i] = (*pSigCounter)++;
                fmt.Fprintf(pFile, " <%d>", pInputNums[i]);
            }
            if fCompNew {
                fmt.Fprintf(pFile, "\'"); }
        }
        fmt.Fprintf(pFile, " )\n");
/*
        fprintf( pFile, " )  " );  
        {
            DdNode * bLocal;
            bLocal = Dsd_TreeGetPrimeFunction( dd, pNodeDsd );  Cudd_Ref( bLocal );
            Extra_bddPrint( dd, bLocal );
            Cudd_RecursiveDeref( dd, bLocal );
        }
*/
        // call recursively for the following blocks
        for  i = 0; i < pNode.nDecs; i++  {
            if pInputNums[i] {
                // inlining of macro Dsd_Regular
                pInput   = ((Dsd_Node_t *)((ABC_PTRUINT_T)(pNode.pDecs[i]) & ~01)) ;
                Buffer = fmt.Sprintf("<%d>", pInputNums[i]);
                // inlining of macro Dsd_Regular
                Dsd_NodePrint_rec( pFile, ((Dsd_Node_t *)((ABC_PTRUINT_T)(pNode.pDecs[i]) & ~01)) , 0, Buffer, nOffset + 6, pSigCounter );
            }
        }
    
    } else if pNode.Type == DSD_NODE_OR {
        // print the line
        fmt.Fprintf(pFile, "OR(");
        for  i = 0; i < pNode.nDecs; i++  {
            // inlining of macro Dsd_Regular
            pInput = ((Dsd_Node_t *)((ABC_PTRUINT_T)(pNode.pDecs[i]) & ~01)) ;
            fCompNew  = int32(( pInput != pNode.pDecs[i] ));
            if i != 0 {
                fmt.Fprintf(pFile, ","); }
            if pInput.Type == DSD_NODE_BUF {
                pInputNums[i] = 0;
                fmt.Fprintf(pFile, " %c", 'a' + pInput.S.index);
            
            } else {
                pInputNums[i] = (*pSigCounter)++;
                fmt.Fprintf(pFile, " <%d>", pInputNums[i]);
            }
            if fCompNew {
                fmt.Fprintf(pFile, "\'"); }
        }
        fmt.Fprintf(pFile, " )\n");
        // call recursively for the following blocks
        for  i = 0; i < pNode.nDecs; i++  {
            if pInputNums[i] {
                // inlining of macro Dsd_Regular
                pInput = ((Dsd_Node_t *)((ABC_PTRUINT_T)(pNode.pDecs[i]) & ~01)) ;
                Buffer = fmt.Sprintf("<%d>", pInputNums[i]);
                // inlining of macro Dsd_Regular
                Dsd_NodePrint_rec( pFile, ((Dsd_Node_t *)((ABC_PTRUINT_T)(pNode.pDecs[i]) & ~01)) , 0, Buffer, nOffset + 6, pSigCounter );
            }
        }
    
    } else if pNode.Type == DSD_NODE_EXOR {
        // print the line
        fmt.Fprintf(pFile, "EXOR(");
        for  i = 0; i < pNode.nDecs; i++  {
            // inlining of macro Dsd_Regular
            pInput = ((Dsd_Node_t *)((ABC_PTRUINT_T)(pNode.pDecs[i]) & ~01)) ;
            fCompNew  = int32(( pInput != pNode.pDecs[i] ));
            assert(fCompNew == 0, "fCompNew == 0");
            if i != 0 {
                fmt.Fprintf(pFile, ","); }
            if pInput.Type == DSD_NODE_BUF {
                pInputNums[i] = 0;
                fmt.Fprintf(pFile, " %c", 'a' + pInput.S.index);
            
            } else {
                pInputNums[i] = (*pSigCounter)++;
                fmt.Fprintf(pFile, " <%d>", pInputNums[i]);
            }
            if fCompNew {
                fmt.Fprintf(pFile, "\'"); }
        }
        fmt.Fprintf(pFile, " )\n");
        // call recursively for the following blocks
        for  i = 0; i < pNode.nDecs; i++  {
            if pInputNums[i] {
                // inlining of macro Dsd_Regular
                pInput = ((Dsd_Node_t *)((ABC_PTRUINT_T)(pNode.pDecs[i]) & ~01)) ;
                Buffer = fmt.Sprintf("<%d>", pInputNums[i]);
                // inlining of macro Dsd_Regular
                Dsd_NodePrint_rec( pFile, ((Dsd_Node_t *)((ABC_PTRUINT_T)(pNode.pDecs[i]) & ~01)) , 0, Buffer, nOffset + 6, pSigCounter );
            }
        }
    }
    ABC_FREE( pInputNums );
}


/**Function*************************************************************

  Synopsis    [Retuns the function of one node of the decomposition tree.]

  Description [This is the old procedure. It is now superceded by the
  procedure Dsd_TreeGetPrimeFunction() found in "dsdLocal.c".]
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
var spStatic_Dsd_TreeGetPrimeFunctionOld_Permute [MAXINPUTS]int32;
func Dsd_TreeGetPrimeFunctionOld(dd *DdManager, pNode *Dsd_Node_t, fRemap int32) *DdNode {
    var bCof0, bCof1, bCube0, bCube1, bNewFunc, bTemp *DdNode;
    var i int32;
    

    assert((pNode != nil), "pNode");
    // inlining of macro Dsd_IsComplement
    assert(!(((int32(((ABC_PTRUINT_T) (pNode) & 01)) != 0))), "!/*__starting_point_inline_macro:Dsd_IsComplement__*/((int32(((ABC_PTRUINT_T) (pNode) & 01))))");
    assert(pNode.Type == DSD_NODE_PRIME, "pNode.Type == DSD_NODE_PRIME");

    // transform the function of this block to depend on inputs
    // corresponding to the formal inputs

    // first, substitute those inputs that have some blocks associated with them
    // second, remap the inputs to the top of the manager (then, it is easy to output them)

    // start the function
    bNewFunc = pNode.G;  Cudd_Ref( bNewFunc );
    // go over all primary inputs
    for  i = 0; i < pNode.nDecs; i++  {
    if pNode.pDecs[i].Type != DSD_NODE_BUF { // remap only if it is not the buffer
    
        // inlining of macro Cudd_Not
        bCube0 = Extra_bddFindOneCube( dd, ((DdNode *)((ptrint)(pNode.pDecs[i].G) ^ 01)) );  Cudd_Ref( bCube0 );
        bCof0 = Cudd_Cofactor( dd, bNewFunc, bCube0 );                     Cudd_Ref( bCof0 );
        Cudd_RecursiveDeref( dd, bCube0 );

        bCube1 = Extra_bddFindOneCube( dd,          pNode.pDecs[i].G  );  Cudd_Ref( bCube1 );
        bCof1 = Cudd_Cofactor( dd, bNewFunc, bCube1 );                     Cudd_Ref( bCof1 );
        Cudd_RecursiveDeref( dd, bCube1 );

        Cudd_RecursiveDeref( dd, bNewFunc );

        // use the variable in the i-th level of the manager
//      bNewFunc = Cudd_bddIte( dd, dd->vars[dd->invperm[i]],bCof1,bCof0 );     Cudd_Ref( bNewFunc );
        // use the first variale in the support of the component
        bNewFunc = Cudd_bddIte( dd, dd.vars[pNode.pDecs[i].S.index],bCof1,bCof0 );     Cudd_Ref( bNewFunc );
        Cudd_RecursiveDeref( dd, bCof0 );
        Cudd_RecursiveDeref( dd, bCof1 );
    }
    }

    if fRemap != 0 {
        // remap the function to the top of the manager
        // remap the function to the first variables of the manager
        for  i = 0; i < pNode.nDecs; i++  {
    //      Permute[ pNode->pDecs[i]->S->index ] = dd->invperm[i];
            spStatic_Dsd_TreeGetPrimeFunctionOld_Permute[ pNode.pDecs[i].S.index ] = i;
        }

        bNewFunc = Cudd_bddPermute( dd, bTemp = bNewFunc, spStatic_Dsd_TreeGetPrimeFunctionOld_Permute );   Cudd_Ref( bNewFunc );
        Cudd_RecursiveDeref( dd, bTemp );
    }

    Cudd_Deref( bNewFunc );
    return bNewFunc;
}


////////////////////////////////////////////////////////////////////////
///                           END OF FILE                            ///
////////////////////////////////////////////////////////////////////////
//ABC_NAMESPACE_IMPL_END

