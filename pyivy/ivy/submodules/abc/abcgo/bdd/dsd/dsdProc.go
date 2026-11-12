package dsd

import (
	"fmt"
	"os"
	"time"
)

/**CFile****************************************************************

  FileName    [dsdProc.c]

  PackageName [DSD: Disjoint-support decomposition package.]

  Synopsis    [The core procedures of the package.]

  Author      [Alan Mishchenko]
  
  Affiliation [UC Berkeley]

  Date        [Ver. 8.0. Started - September 22, 2003.]

  Revision    [$Id: dsdProc.c,v 1.0 2002/22/09 00:00:00 alanmi Exp $]

***********************************************************************/

//#include "dsdInt.h"

//ABC_NAMESPACE_IMPL_START



////////////////////////////////////////////////////////////////////////
///                    FUNCTION DECLARATIONS                         ///
////////////////////////////////////////////////////////////////////////

// the most important procedures
// void dsdKernelDecompose( Dsd_Manager_t * pDsdMan, DdNode ** pbFuncs, int nFuncs );

// static Dsd_Node_t * dsdKernelDecompose_rec( Dsd_Manager_t * pDsdMan, DdNode * F );


// additional procedures
// static Dsd_Node_t * dsdKernelFindContainingComponent( Dsd_Manager_t * pDsdMan, Dsd_Node_t * pWhere, DdNode * Var, int * fPolarity );

// static int dsdKernelFindCommonComponents( Dsd_Manager_t * pDsdMan, Dsd_Node_t * pL, Dsd_Node_t * pH, Dsd_Node_t *** pCommon, Dsd_Node_t ** pLastDiffL, Dsd_Node_t ** pLastDiffH );

// static void dsdKernelComputeSumOfComponents( Dsd_Manager_t * pDsdMan, Dsd_Node_t ** pCommon, int nCommon, DdNode ** pCompF, DdNode ** pCompS, int fExor );

// static int dsdKernelCheckContainment( Dsd_Manager_t * pDsdMan, Dsd_Node_t * pL, Dsd_Node_t * pH, Dsd_Node_t ** pLarge, Dsd_Node_t ** pSmall );


// list copying
// static void dsdKernelCopyListPlusOne( Dsd_Node_t * p, Dsd_Node_t * First, Dsd_Node_t ** ppList, int nListSize );

// static void dsdKernelCopyListPlusOneMinusOne( Dsd_Node_t * p, Dsd_Node_t * First, Dsd_Node_t ** ppList, int nListSize, int Skipped );


// debugging procedures
// static int dsdKernelVerifyDecomposition( Dsd_Manager_t * pDsdMan, Dsd_Node_t * pDE );


////////////////////////////////////////////////////////////////////////
///                       STATIC VARIABLES                           ///
////////////////////////////////////////////////////////////////////////

// the counter of marks
var s_Mark int32;

// debugging flag
//static int s_Show = 0;
// temporary var used for debugging
var Depth int32;

var s_Loops1 int32;
var s_Loops2 int32;
var s_Loops3 int32;
var s_Common int32;
var s_CommonNo int32;

var s_Case4Calls int32;
var s_Case4CallsSpecial int32;

//static int s_Case5;
//static int s_Loops2Useless;

// statistical variables
var s_nDecBlocks int32;
var s_nLiterals int32;
var s_nExorGates int32; 
var s_nReusedBlocks int32;
var s_nCascades int32;
var s_nPrimeBlocks int32;

var HashSuccess int32;
var HashFailure int32;

var s_CacheEntries int32;


////////////////////////////////////////////////////////////////////////
///                     DECOMPOSITION FUNCTIONS                      ///
////////////////////////////////////////////////////////////////////////

/**Function*************************************************************

  Synopsis    [Performs DSD for the array of functions represented by BDDs.]

  Description [This function takes the DSD manager, which should be
  previously allocated by the call to Dsd_ManagerStart(). The resulting
  DSD tree is stored in the DSD manager (pDsdMan->pRoots, pDsdMan->nRoots).
  Access to the tree is through the APIs of the manager. The resulting
  tree is a shared DSD DAG for the functions given in the array. For one
  function the resulting DAG is always a tree. The root node pointers can 
  be complemented, as discussed in the literature referred to in "dsd.h".
  This procedure can be called repeatedly for different functions. There is
  no need to remove the decomposition tree after it is returned, because
  the next call to the DSD manager will "recycle" the tree. The user should
  not modify or dereference any data associated with the nodes of the 
  DSD trees (the user can only change the contents of a temporary
  mark associated with each node by the calling to Dsd_NodeSetMark()).
  All the decomposition trees and intermediate nodes will be removed when
  the DSD manager is deallocated at the end by calling Dsd_ManagerStop().]
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Dsd_Decompose(pDsdMan *Dsd_Manager_t, pbFuncs **DdNode, nFuncs int32) {
    var dd *DdManager  = pDsdMan.dd;
    var i int32;
    var clk abctime;
    var pTemp *Dsd_Node_t;
    var SumMaxGateSize int32;
    var nDecOutputs int32;
    var nCBFOutputs int32;
/*
s_Loops1 = 0;
s_Loops2 = 0;
s_Loops3 = 0;
s_Case4Calls = 0;
s_Case4CallsSpecial = 0;
s_Case5 = 0;
s_Loops2Useless = 0;
*/
    // resize the number of roots in the manager
    if pDsdMan.nRootsAlloc < nFuncs {
        if pDsdMan.nRootsAlloc > 0 {
            ABC_FREE( pDsdMan.pRoots ); }
        pDsdMan.nRootsAlloc = nFuncs;
        pDsdMan.pRoots = (Dsd_Node_t **) ABC_ALLOC( char, pDsdMan.nRootsAlloc * sizeof(Dsd_Node_t *) );
    }

    if pDsdMan.fVerbose {
        fmt.Printf("\nDecomposability statistics for individual outputs:\n"); }

    // set the counter of decomposition nodes
    s_nDecBlocks = 0;

    // perform decomposition for all outputs
    clk = Abc_Clock();
    pDsdMan.nRoots = 0;
    s_nCascades = 0;
    for  i = 0; i < nFuncs; i++  {
        var nLiteralsPrev int32;
        var nDecBlocksPrev int32;
        var nExorGatesPrev int32;
        var nReusedBlocksPres int32;
        var nCascades int32;
        var MaxBlock int32;
        var nPrimeBlocks int32;
        var clk abctime;

        clk = Abc_Clock();
        nLiteralsPrev     = s_nLiterals;
        nDecBlocksPrev    = s_nDecBlocks;
        nExorGatesPrev    = s_nExorGates;
        nReusedBlocksPres = s_nReusedBlocks;
        nPrimeBlocks      = s_nPrimeBlocks;

        pDsdMan.pRoots[ pDsdMan.nRoots++ ] = dsdKernelDecompose_rec( pDsdMan, pbFuncs[i] );

        Dsd_TreeNodeGetInfoOne( pDsdMan.pRoots[i], &nCascades, &MaxBlock );
        // inlining of macro ddMax
        
        if (nCascades) > (s_nCascades) {
            s_nCascades = (nCascades);
        } else {
            s_nCascades = (s_nCascades);
        }
        // inlining of macro Dsd_Regular
        pTemp = ((Dsd_Node_t *)((ABC_PTRUINT_T)(pDsdMan.pRoots[i]) & ~01)) ;
        if pTemp.Type != DSD_NODE_PRIME || pTemp.nDecs != Extra_bddSuppSize(dd,pTemp.S) {
            nDecOutputs++; }
        if MaxBlock < 3 {
            nCBFOutputs++; }
        SumMaxGateSize += MaxBlock;

        if pDsdMan.fVerbose {
            fmt.Printf("#%02d: ", i);                              
            fmt.Printf("Ins=%2d. ", Cudd_SupportSize(dd,pbFuncs[i]));                  
            fmt.Printf("Gts=%3d. ", Dsd_TreeCountNonTerminalNodesOne( pDsdMan.pRoots[i] )); 
            fmt.Printf("Pri=%3d. ", Dsd_TreeCountPrimeNodesOne( pDsdMan.pRoots[i] )); 
            fmt.Printf("Max=%3d. ", MaxBlock); 
            fmt.Printf("Reuse=%2d. ", s_nReusedBlocks-nReusedBlocksPres); 
            fmt.Printf("Csc=%2d. ", nCascades); 
            fmt.Printf("T= %.2f s. ", float32((Abc_Clock()-clk))/float32((int64(time.Second)))) ;
            fmt.Printf("Bdd=%2d. ", Cudd_DagSize(pbFuncs[i])); 
            fmt.Printf("\n");
            fflush( os.Stdout );
        }
    }
    assert(pDsdMan.nRoots == nFuncs, "pDsdMan.nRoots == nFuncs");

    if pDsdMan.fVerbose {
        fmt.Printf("\n");
        fmt.Printf("The cumulative decomposability statistics:\n");
        fmt.Printf("  Total outputs                             = %5d\n", nFuncs);
        fmt.Printf("  Decomposable outputs                      = %5d\n", nDecOutputs);
        fmt.Printf("  Completely decomposable outputs           = %5d\n", nCBFOutputs);
        fmt.Printf("  The sum of max gate sizes                 = %5d\n", SumMaxGateSize);
        fmt.Printf("  Shared BDD size                           = %5d\n", Cudd_SharingSize( pbFuncs, nFuncs ));
        fmt.Printf("  Decomposition entries                     = %5d\n", st__count( pDsdMan.Table ));
        fmt.Printf("  Pure decomposition time                   =  %.2f sec\n", float32((Abc_Clock() - clk))/float32((int64(time.Second))));
    }
/*
    printf( "s_Loops1 = %d.\n", s_Loops1 );
    printf( "s_Loops2 = %d.\n", s_Loops2 );
    printf( "s_Loops3 = %d.\n", s_Loops3 );
    printf( "s_Case4Calls = %d.\n", s_Case4Calls );
    printf( "s_Case4CallsSpecial = %d.\n", s_Case4CallsSpecial );
    printf( "s_Case5 = %d.\n", s_Case5 );
    printf( "s_Loops2Useless = %d.\n", s_Loops2Useless );
*/
}

/**Function*************************************************************

  Synopsis    [Performs decomposition for one function.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Dsd_DecomposeOne(pDsdMan *Dsd_Manager_t, bFunc *DdNode) *Dsd_Node_t {
    return dsdKernelDecompose_rec( pDsdMan, bFunc );
}

/**Function*************************************************************

  Synopsis    [The main function of this module. Recursive implementation of DSD.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
var spStatic_dsdKernelDecompose_rec_pNonOverlap [MAXINPUTS]*Dsd_Node_t;
var spStatic_dsdKernelDecompose_rec_pMarkedLeft [MAXINPUTS]*Dsd_Node_t;
var spStatic_dsdKernelDecompose_rec_pMarkedPols [MAXINPUTS]byte;
func dsdKernelDecompose_rec(pDsdMan *Dsd_Manager_t, bFunc0 *DdNode) *Dsd_Node_t {
    var dd *DdManager  = pDsdMan.dd;
    var bLow *DdNode;
    var bLowR *DdNode;
    var bHigh *DdNode;

    var VarInt int32;
    var bVarCur *DdNode;
    var pVarCurDE *Dsd_Node_t; 
    // works only if var indices start from 0!!!
    var bSuppNew *DdNode = nil;
    var bTemp *DdNode;

    var fContained int32;
    var nSuppLH int32;
    var nSuppL int32;
    var nSuppH int32;



    // various decomposition nodes
    var pThis, pL, pH, pLR, pHR *Dsd_Node_t;

    var pSmallR, pLargeR *Dsd_Node_t;
    var pTableEntry *Dsd_Node_t;


    // treat the complemented case
    // inlining of macro Cudd_Regular
    var bF *DdNode  = ((DdNode *)((ptruint)(bFunc0) & ~01));
    var fCompF int32  = int32((bF != bFunc0));

    // check cache
    if st__lookup( pDsdMan.Table, (char*)bF, (char**)&pTableEntry ) { // the entry is present 
        HashSuccess++;
        // inlining of macro Dsd_NotCond
        return ((Dsd_Node_t *)((ABC_PTRUINT_T)(pTableEntry) ^ (fCompF)));
    }
    HashFailure++;
    Depth++;

    // proceed to consider "four cases"
    //////////////////////////////////////////////////////////////////////
    // TERMINAL CASES - CASES 1 and 2
    //////////////////////////////////////////////////////////////////////
    // inlining of macro cuddE
    bLow    = ((bF).type.kids.E);
    // inlining of macro Cudd_Regular
    bLowR   = ((DdNode *)((ptruint)(bLow) & ~01));
    // inlining of macro cuddT
    bHigh   = ((bF).type.kids.T);
    VarInt    = bF.index;
    bVarCur   = dd.vars[VarInt];
    pVarCurDE = pDsdMan.pInputs[VarInt]; 
    // works only if var indices start from 0!!!
    bSuppNew = nil;

    if bLowR.index == CUDD_CONST_INDEX || bHigh.index == CUDD_CONST_INDEX { // one of the cofactors in the constant
        if bHigh == b1 {  // bHigh cannot be equal to b0, because then it will be complemented
          if bLow == b0 { // bLow cannot be equal to b1, because then the node will have bLow == bHigh
          /////////////////////////////////////////////////////////////////
          // bLow == 0, bHigh == 1, F = x'&0 + x&1 = x
          /////////////////////////////////////////////////////////////////
           // create the elementary variable node
            assert(false, "0"); // should be already in the hash table
            pThis = Dsd_TreeNodeCreate( DSD_NODE_BUF, 1, s_nDecBlocks++ );
            pThis.pDecs[0] = nil;
          
          } else { // if ( bLow != constant )
          /////////////////////////////////////////////////////////////////
          // bLow != const, bHigh == 1, F = x'&bLow + x&1 = bLow + x  --- DSD_NODE_OR(x,bLow)
          /////////////////////////////////////////////////////////////////
          
            pL  = dsdKernelDecompose_rec( pDsdMan, bLow );
            // inlining of macro Dsd_Regular
            pLR = ((Dsd_Node_t *)((ABC_PTRUINT_T)(pL) & ~01)) ;
            bSuppNew = Cudd_bddAnd( dd, bVarCur, pLR.S ); Cudd_Ref(bSuppNew);
            if pLR.Type == DSD_NODE_OR && pL == pLR { // OR and no complement
             // add to the components
                pThis = Dsd_TreeNodeCreate( DSD_NODE_OR, pL.nDecs+1, s_nDecBlocks++ );
                dsdKernelCopyListPlusOne( pThis, pVarCurDE, pL.pDecs, pL.nDecs );
            
            } else { // all other cases
             // create a new 2-input OR-gate
                pThis = Dsd_TreeNodeCreate( DSD_NODE_OR, 2, s_nDecBlocks++ );
                dsdKernelCopyListPlusOne( pThis, pVarCurDE, &pL, 1 );
            }
          }
        } else { // if ( bHigh != const ) // meaning that bLow should be a constant
        
          pH = dsdKernelDecompose_rec( pDsdMan, bHigh );
          // inlining of macro Dsd_Regular
          pHR = ((Dsd_Node_t *)((ABC_PTRUINT_T)(pH) & ~01)) ;
          bSuppNew = Cudd_bddAnd( dd, bVarCur, pHR.S ); Cudd_Ref(bSuppNew);
          if bLow == b0 {
          /////////////////////////////////////////////////////////////////
          // Low == 0, High != 1, F = x'&0+x&High = (x'+High')'--- NOR(x',High')
          /////////////////////////////////////////////////////////////////
            if pHR.Type == DSD_NODE_OR && pH != pHR { // DSD_NODE_OR and complement
             // add to the components
              pThis = Dsd_TreeNodeCreate( DSD_NODE_OR, pHR.nDecs+1, s_nDecBlocks++ );
              // inlining of macro Dsd_Not
              dsdKernelCopyListPlusOne( pThis, ((Dsd_Node_t *)((ABC_PTRUINT_T)(pVarCurDE) ^ 01)) , pHR.pDecs, pHR.nDecs );
              // inlining of macro Dsd_Not
              pThis = ((Dsd_Node_t *)((ABC_PTRUINT_T)(pThis) ^ 01)) ;
            
            } else { // all other cases
             // create a new 2-input NOR gate
              pThis = Dsd_TreeNodeCreate( DSD_NODE_OR, 2, s_nDecBlocks++ );
              // inlining of macro Dsd_Not
              pH = ((Dsd_Node_t *)((ABC_PTRUINT_T)(pH) ^ 01)) ;
              // inlining of macro Dsd_Not
              dsdKernelCopyListPlusOne( pThis, ((Dsd_Node_t *)((ABC_PTRUINT_T)(pVarCurDE) ^ 01)) , &pH, 1 );
              // inlining of macro Dsd_Not
              pThis = ((Dsd_Node_t *)((ABC_PTRUINT_T)(pThis) ^ 01)) ;
            }
          } else // if ( bLow == b1 )
          /////////////////////////////////////////////////////////////////
          // Low == 1, High != 1, F = x'&1 + x&High = x' + High --- DSD_NODE_OR(x',High)
          /////////////////////////////////////////////////////////////////
            if pHR.Type == DSD_NODE_OR && pH == pHR { // OR and no complement
             // add to the components
                pThis = Dsd_TreeNodeCreate( DSD_NODE_OR, pH.nDecs+1, s_nDecBlocks++ );
                // inlining of macro Dsd_Not
                dsdKernelCopyListPlusOne( pThis, ((Dsd_Node_t *)((ABC_PTRUINT_T)(pVarCurDE) ^ 01)) , pH.pDecs, pH.nDecs );
            
            } else { // all other cases
             // create a new 2-input OR-gate
                pThis = Dsd_TreeNodeCreate( DSD_NODE_OR, 2, s_nDecBlocks++ );
                // inlining of macro Dsd_Not
                dsdKernelCopyListPlusOne( pThis, ((Dsd_Node_t *)((ABC_PTRUINT_T)(pVarCurDE) ^ 01)) , &pH, 1 );
            }
        }
        goto EXIT;
    }
    // else if ( bLow != const && bHigh != const )

    // the case of equal cofactors (up to complementation)
    if bLowR == bHigh {
    /////////////////////////////////////////////////////////////////
    // Low == G, High == G', F = x'&G + x&G' = (x(+)G) --- EXOR(x,Low)
    /////////////////////////////////////////////////////////////////
    
        pL  = dsdKernelDecompose_rec( pDsdMan, bLow );
        // inlining of macro Dsd_Regular
        pLR = ((Dsd_Node_t *)((ABC_PTRUINT_T)(pL) & ~01)) ;
        bSuppNew = Cudd_bddAnd( dd, bVarCur, pLR.S ); Cudd_Ref(bSuppNew);
        if pLR.Type == DSD_NODE_EXOR { // complemented or not - does not matter!
         // add to the components
            pThis = Dsd_TreeNodeCreate( DSD_NODE_EXOR, pLR.nDecs+1, s_nDecBlocks++ );
            dsdKernelCopyListPlusOne( pThis, pVarCurDE, pLR.pDecs, pLR.nDecs );
            if pL != pLR {
                // inlining of macro Dsd_Not
                pThis = ((Dsd_Node_t *)((ABC_PTRUINT_T)(pThis) ^ 01)) ; }
        
        } else { // all other cases
         // create a new 2-input EXOR-gate
            pThis = Dsd_TreeNodeCreate( DSD_NODE_EXOR, 2, s_nDecBlocks++ );
            if pL != pLR { // complemented
            
                dsdKernelCopyListPlusOne( pThis, pVarCurDE, &pLR, 1 );
                // inlining of macro Dsd_Not
                pThis = ((Dsd_Node_t *)((ABC_PTRUINT_T)(pThis) ^ 01)) ;
            
            } else { // non-complemented
                dsdKernelCopyListPlusOne( pThis, pVarCurDE, &pL, 1 ); }
        }
        goto EXIT;
    }

    //////////////////////////////////////////////////////////////////////
    // solve subproblems
    //////////////////////////////////////////////////////////////////////
    pL   = dsdKernelDecompose_rec( pDsdMan, bLow );
    pH   = dsdKernelDecompose_rec( pDsdMan, bHigh );
    // inlining of macro Dsd_Regular
    pLR  = ((Dsd_Node_t *)((ABC_PTRUINT_T)(pL) & ~01)) ;
    // inlining of macro Dsd_Regular
    pHR  = ((Dsd_Node_t *)((ABC_PTRUINT_T)(pH) & ~01)) ;

    assert(pLR.Type == DSD_NODE_BUF || pLR.Type == DSD_NODE_OR || pLR.Type == DSD_NODE_EXOR || pLR.Type == DSD_NODE_PRIME, "pLR.Type == DSD_NODE_BUF || pLR.Type == DSD_NODE_OR || pLR.Type == DSD_NODE_EXOR || pLR.Type == DSD_NODE_PRIME");
    assert(pHR.Type == DSD_NODE_BUF || pHR.Type == DSD_NODE_OR || pHR.Type == DSD_NODE_EXOR || pHR.Type == DSD_NODE_PRIME, "pHR.Type == DSD_NODE_BUF || pHR.Type == DSD_NODE_OR || pHR.Type == DSD_NODE_EXOR || pHR.Type == DSD_NODE_PRIME");

/*
if ( Depth == 1 )
{
//  PRK(bLow,pDecTreeTotal->nInputs);
//  PRK(bHigh,pDecTreeTotal->nInputs);
if ( s_Show )
{
    PRD( pL );
    PRD( pH );
}
}
*/
    // compute the new support
    bTemp    = Cudd_bddAnd( dd, pLR.S, pHR.S );   Cudd_Ref( bTemp );
    nSuppL   = Extra_bddSuppSize( dd, pLR.S );
    nSuppH   = Extra_bddSuppSize( dd, pHR.S );
    nSuppLH  = Extra_bddSuppSize( dd, bTemp );
    bSuppNew = Cudd_bddAnd( dd, bTemp, bVarCur );   Cudd_Ref( bSuppNew );
    Cudd_RecursiveDeref( dd, bTemp );


    // several possibilities are possible
    // (1) support of one component contains another
    // (2) none of the supports is contained in another
    fContained = dsdKernelCheckContainment( pDsdMan, pLR, pHR, &pLargeR, &pSmallR );

    //////////////////////////////////////////////////////////////////////
    // CASE 3.b One of the cofactors in a constant (OR and EXOR)
    //////////////////////////////////////////////////////////////////////
    // the support of the larger component should contain the support of the smaller
    // it is possible to have PRIME function in this role
    // for example: F = ITE( a+b, c(+)d, e+f ), F0 = ITE( b, c(+)d, e+f ), F1 = c(+)d
    if fContained != 0 {
        var pSmall, pLarge *Dsd_Node_t;
        var c int32;
        var iCompLarge int32 = -1; // the number of the component is Large is equal to the whole of Small; suppress "might be used uninitialized"
        var fLowIsLarge int32;

        var bFTemp *DdNode;     // the changed input function
        var pDETemp, pDENew *Dsd_Node_t;

        var pComp *Dsd_Node_t  = nil;
        var nComp int32  = -1; // Suppress "might be used uninitialized"

        if pSmallR == pLR { // Low is Small => High is Large
            pSmall = pL;
            pLarge = pH;
            fLowIsLarge = 0;
        
        } else { // vice versa
            pSmall = pH;
            pLarge = pL;
            fLowIsLarge = 1;
        }

        // treat the situation when the larger is PRIME
        if pLargeR.Type == DSD_NODE_PRIME { //&& pLargeR->nDecs != pSmallR->nDecs )
        
            // QUESTION: Is it possible for pLargeR->nDecs > 3 
            // and pSmall contained as one of input in pLarge?
            // Yes, for example F = a'c + a & MUX(b,c',d) = a'c + abc' + ab'd is non-decomposable
            // Consider the function H(a->xy) = F( xy, b, c, d )
            // H0 = H(x=0) = F(0,b,c,d) = c
            // H1 = F(x=1) = F(y,b,c,d) - non-decomposable
            //
            // QUESTION: Is it possible that pLarge is PRIME(3) and pSmall is OR(2),
            // which is not contained in PRIME as one input?
            // Yes, for example F = abcd + b'c'd' + a'c'd' = PRIME(ab, c, d)
            // F(a=0) = c'd' = NOT(OR(a,d))  F(a=1) = bcd + b'c'd' = PRIME(b,c,d)
            // To find decomposition, we have to prove that F(a=1)|b=0 = F(a=0)

            // Is it possible that (pLargeR->nDecs == pSmallR->nDecs) and yet this case holds?
            // Yes, consider the function such that F(a=0) = PRIME(a,b+c,d,e) and F(a=1) = OR(b,c,d,e)
            // They have the same number of inputs and it is possible that they will be the cofactors
            // as discribed in the previous example.

            // find the component, which when substituted for 0 or 1, produces the desired result
            var g int32;
            var fFoundComp int32 = -1; // {0,1} depending on whether setting cofactor to 0 or 1 worked out; suppress "might be used uninitialized"

            var bLarge, bSmall *DdNode;
            if fLowIsLarge != 0 {
                bLarge = bLow;
                bSmall = bHigh;
            
            } else {
                bLarge = bHigh;
                bSmall = bLow;
            }

            for  g = 0; g < pLargeR.nDecs; g++  {
//          if ( g != c )
            
                pDETemp = pLargeR.pDecs[g]; // cannot be complemented
                if Dsd_CheckRootFunctionIdentity( dd, bLarge, bSmall, pDETemp.G, b1 ) {
                    fFoundComp = 1;
                    break;
                }

                s_Loops1++;

                // inlining of macro Cudd_Not
                if Dsd_CheckRootFunctionIdentity( dd, bLarge, bSmall, ((DdNode *)((ptrint)(pDETemp.G) ^ 01)), b1 ) {
                    fFoundComp = 0;
                    break;
                }

                s_Loops1++;
            }

            if g != pLargeR.nDecs { // decomposition is found
                if fFoundComp != 0 {
                    if fLowIsLarge != 0 {
                        bFTemp = Cudd_bddOr( dd, bVarCur, pLargeR.pDecs[g].G );
                    } else {
                        // inlining of macro Cudd_Not
                        bFTemp = Cudd_bddOr( dd, ((DdNode *)((ptrint)(bVarCur) ^ 01)), pLargeR.pDecs[g].G ); }
                } else
                    if fLowIsLarge != 0 {
                        // inlining of macro Cudd_Not
                        bFTemp = Cudd_bddAnd( dd, ((DdNode *)((ptrint)(bVarCur) ^ 01)), pLargeR.pDecs[g].G );
                    } else {
                        bFTemp = Cudd_bddAnd( dd, bVarCur, pLargeR.pDecs[g].G ); }
                Cudd_Ref( bFTemp );

                pDENew = dsdKernelDecompose_rec( pDsdMan, bFTemp );
                // inlining of macro Dsd_Regular
                pDENew = ((Dsd_Node_t *)((ABC_PTRUINT_T)(pDENew) & ~01)) ;
                Cudd_RecursiveDeref( dd, bFTemp );

                // get the new gate
                pThis = Dsd_TreeNodeCreate( DSD_NODE_PRIME, pLargeR.nDecs, s_nDecBlocks++ );
                dsdKernelCopyListPlusOneMinusOne( pThis, pDENew, pLargeR.pDecs, pLargeR.nDecs, g );
                goto EXIT;
            }
        }

        // try to find one component in the pLarger that is equal to the whole of pSmaller
        for  c = 0; c < pLargeR.nDecs; c++  {
            // inlining of macro Dsd_Not
            if pLargeR.pDecs[c] == pSmall || pLargeR.pDecs[c] == ((Dsd_Node_t *)((ABC_PTRUINT_T)(pSmall) ^ 01)) {
                iCompLarge = c;
                break;
            }
        }

        // assign the equal component
        if c != pLargeR.nDecs {  // the decomposition is possible!
         
            pComp  = pLargeR.pDecs[iCompLarge];
            nComp  = 1;
        
        } else { // the decomposition is still possible
         // for example F = OR(ab,c,d), F(a=0) = OR(c,d), F(a=1) = OR(b,c,d)
            // supp(F0) is contained in supp(F1), Polarity(F(a=0)) == Polarity(F(a=1))

            // try to find a group of common components
            if pLargeR.Type == pSmallR.Type &&
                (pLargeR.Type == DSD_NODE_EXOR || (pSmallR.Type == DSD_NODE_OR && ((pLarge==pLargeR) == (pSmall==pSmallR)))) {
                var pCommon **Dsd_Node_t;
                var pLastDiffL *Dsd_Node_t = nil;
                var pLastDiffH *Dsd_Node_t = nil; 
                var nCommon int32  = dsdKernelFindCommonComponents( pDsdMan, pLargeR, pSmallR, &pCommon, &pLastDiffL, &pLastDiffH );
                // if all the components of pSmall are contained in pLarge,
                // then the decomposition exists
                if nCommon == pSmallR.nDecs {
                    pComp = pSmallR;
                    nComp = pSmallR.nDecs;
                }
            }
        }

        if pComp != nil { // the decomposition is possible!
        
//          Dsd_Node_t * pComp  = pLargeR->pDecs[iCompLarge];
            // inlining of macro Dsd_Regular
            var pCompR *Dsd_Node_t  = ((Dsd_Node_t *)((ABC_PTRUINT_T)(pComp) & ~01)) ;
            var fComp1 int32  = int32(( pLarge != pLargeR ));
            var fComp2 int32  = int32(( pComp  != pCompR ));
            var fComp3 int32  = int32(( pSmall != pSmallR ));

            var bFuncComp *DdNode;  // the function of the given component
            var bFuncNew *DdNode;   // the function of the input component

            if pLargeR.Type == DSD_NODE_OR { // Figure 4 of Matsunaga's paper
             
                // the decomposition exists only if the polarity assignment 
                // along the paths is the same
                if (fComp1 ^ fComp2) == fComp3 { // decomposition exists = consider 4 cases
                    // consideration of cases leads to the following conclusion
                    // fComp1 gives the polarity of the resulting DSD_NODE_OR gate
                    // fComp2 gives the polarity of the common component feeding into the DSD_NODE_OR gate
                    //
                    //                  |  fComp1              pL/  |pS
                    //                  <> .........<=>....... <>   |
                    //                  |                     /     |
                    //                [OR]                  [OR]    | fComp3
                    //                /  \  fComp2          / | \   |
                    //              <>    <> .......<=>... /..|..<> | 
                    //             /        \             /   |    \|
                    //          [OR]        [C]          S1   S2    C 
                    //          /  \      .
                    //        <>    \     .
                    //       /       \    .
                    //     [OR]      [x]  .
                    //     /  \           .
                    //    S1   S2         .
                    //


                    // at this point we have the function F (bFTemp) and the common component C (bFuncComp)
                    // to get the remainder, R, in the relationship F = R + C, supp(R) & supp(C) = 0
                    // we compute the following R = Exist( F - C, supp(C) )
                    if fComp1 != 0 {
                        // inlining of macro Cudd_Not
                        bFTemp = ((DdNode *)((ptrint)(bF) ^ 01));
                    } else {
                        bFTemp = bF;
                    }
                    if fComp2 != 0 {
                        // inlining of macro Cudd_Not
                        bFuncComp = ((DdNode *)((ptrint)(pCompR.G) ^ 01));
                    } else {
                        bFuncComp = pCompR.G;
                    }
                    // inlining of macro Cudd_Not
                    bFuncNew  = Cudd_bddAndAbstract( dd, bFTemp, ((DdNode *)((ptrint)(bFuncComp) ^ 01)), pCompR.S ); Cudd_Ref( bFuncNew );

                    // there is no need to copy the dec entry list first, because pComp is a component
                    // which will not be destroyed by the recursive call to decomposition
                    pDENew = dsdKernelDecompose_rec( pDsdMan, bFuncNew );
                    // inlining of macro Dsd_IsComplement
                    assert((((int32(((ABC_PTRUINT_T) (pDENew) & 01)) != 0))), "/*__starting_point_inline_macro:Dsd_IsComplement__*/((int32(((ABC_PTRUINT_T) (pDENew) & 01))))"); // follows from the consideration of cases
                    Cudd_RecursiveDeref( dd, bFuncNew );

                    // get the new gate
                    if nComp == 1 {
                        pThis = Dsd_TreeNodeCreate( DSD_NODE_OR, 2, s_nDecBlocks++ );
                        pThis.pDecs[0] = pDENew;
                        pThis.pDecs[1] = pComp; // takes the complement
                    
                    } else {  // pComp is not complemented
                        pThis = Dsd_TreeNodeCreate( DSD_NODE_OR, nComp+1, s_nDecBlocks++ );
                        dsdKernelCopyListPlusOne( pThis, pDENew, pComp.pDecs, nComp );
                    }
                    
                    if fComp1 != 0 {
                        // inlining of macro Dsd_Not
                        pThis = ((Dsd_Node_t *)((ABC_PTRUINT_T)(pThis) ^ 01)) ; }
                    goto EXIT;
                }
            
            } else if pLargeR.Type == DSD_NODE_EXOR { // Figure 5 of Matsunaga's paper (with correction)
             // decomposition always exists = consider 4 cases

                // consideration of cases leads to the following conclusion
                // fComp3 gives the COMPLEMENT of the polarity of the resulting EXOR gate
                // (if fComp3 is 0, the EXOR gate is complemented, and vice versa)
                //
                //                  |  fComp1              pL/  |pS
                //                  <> .........<=>....... /....|  fComp3
                //                  |                     /     |
                //                [XOR]                [XOR]    |
                //                /  \  fComp2==0       / | \   |
                //              /     \                /  |  \  | 
                //             /        \             /   |    \|
                //          [OR]        [C]          S1   S2    C 
                //          /  \     .
                //        <>    \    .
                //       /       \   .
                //    [XOR]      [x] .
                //     /  \          .
                //    S1   S2        .
                //

                assert(fComp2 == 0, "fComp2 == 0");
                // find the functionality of the lower gates
                if fComp3 != 0 {
                    bFTemp = bF;
                } else {
                    // inlining of macro Cudd_Not
                    bFTemp = ((DdNode *)((ptrint)(bF) ^ 01));
                }
                bFuncNew = Cudd_bddXor( dd, bFTemp, pComp.G );   Cudd_Ref( bFuncNew );

                pDENew = dsdKernelDecompose_rec( pDsdMan, bFuncNew );
                // inlining of macro Dsd_IsComplement
                assert(!(((int32(((ABC_PTRUINT_T) (pDENew) & 01)) != 0))), "!/*__starting_point_inline_macro:Dsd_IsComplement__*/((int32(((ABC_PTRUINT_T) (pDENew) & 01))))"); // follows from the consideration of cases
                Cudd_RecursiveDeref( dd, bFuncNew ); 

                // get the new gate
                if nComp == 1 {
                    pThis = Dsd_TreeNodeCreate( DSD_NODE_EXOR, 2, s_nDecBlocks++ );
                    pThis.pDecs[0] = pDENew;
                    pThis.pDecs[1] = pComp; 
                
                } else {  // pComp is not complemented
                    pThis = Dsd_TreeNodeCreate( DSD_NODE_EXOR, nComp+1, s_nDecBlocks++ );
                    dsdKernelCopyListPlusOne( pThis, pDENew, pComp.pDecs, nComp );
                }

                if !(fComp3 != 0) {
                    // inlining of macro Dsd_Not
                    pThis = ((Dsd_Node_t *)((ABC_PTRUINT_T)(pThis) ^ 01)) ; }
                goto EXIT;
            }
        }
    }

    // this case was added to fix the trivial bug found November 4, 2002 in Japan
    // by running the example provided by T. Sasao
    if nSuppLH == nSuppL + nSuppH { // the supports of the components are disjoint
    
        // create a new component of the type ITE( a, pH, pL )
        pThis = Dsd_TreeNodeCreate( DSD_NODE_PRIME, 3, s_nDecBlocks++ );
        if dd.perm[pLR.S.index] < dd.perm[pHR.S.index] { // pLR is higher in the varible order
        
            pThis.pDecs[1] = pLR;
            pThis.pDecs[2] = pHR;
        
        } else {  // pHR is higher in the varible order
        
            pThis.pDecs[1] = pHR;
            pThis.pDecs[2] = pLR;
        }
        // add the first component
        pThis.pDecs[0] = pVarCurDE;
        goto EXIT;
    }


    //////////////////////////////////////////////////////////////////////
    // CASE 3.a Neither of the cofactors is a constant (OR, EXOR, PRIME)
    //////////////////////////////////////////////////////////////////////
    // the component types are identical 
    // and if they are OR, they are either both complemented or both not complemented
    // and if they are PRIME, their dec numbers should be the same
    if pLR.Type == pHR.Type && 
         pLR.Type != DSD_NODE_BUF &&           
        (pLR.Type != DSD_NODE_OR    || ( (pL == pLR && pH == pHR) || (pL != pLR && pH != pHR) ) ) &&
        (pLR.Type != DSD_NODE_PRIME || pLR.nDecs == pHR.nDecs) {
        // array to store common comps in pL and pH
        var pCommon **Dsd_Node_t;
        var pLastDiffL *Dsd_Node_t = nil;
        var pLastDiffH *Dsd_Node_t = nil; 
        var nCommon int32  = dsdKernelFindCommonComponents( pDsdMan, pLR, pHR, &pCommon, &pLastDiffL, &pLastDiffH );
        if nCommon != 0 {
            if pLR.Type == DSD_NODE_OR { // Figure 2 of Matsunaga's paper
             // at this point we have the function F and the group of common components C
                // to get the remainder, R, in the relationship F = R + C, supp(R) & supp(C) = 0
                // we compute the following R = Exist( F - C, supp(C) )

                // compute the sum total of the common components and the union of their supports
                var bCommF, bCommS, bFTemp, bFuncNew *DdNode;
                var pDENew *Dsd_Node_t;

                dsdKernelComputeSumOfComponents( pDsdMan, pCommon, nCommon, &bCommF, &bCommS, 0 );
                Cudd_Ref( bCommF );
                Cudd_Ref( bCommS );
                if pL != pLR {
                    // inlining of macro Cudd_Not
                    bFTemp = ((DdNode *)((ptrint)(bF) ^ 01));
                } else {
                    bFTemp = bF;
                }

                // inlining of macro Cudd_Not
                bFuncNew = Cudd_bddAndAbstract( dd, bFTemp, ((DdNode *)((ptrint)(bCommF) ^ 01)), bCommS ); Cudd_Ref( bFuncNew );
                Cudd_RecursiveDeref( dd, bCommF );
                Cudd_RecursiveDeref( dd, bCommS );

                // get the new gate

                // copy the components first, then call the decomposition
                // because decomposition will distroy the list used for copying
                pThis = Dsd_TreeNodeCreate( DSD_NODE_OR, nCommon + 1, s_nDecBlocks++ );
                dsdKernelCopyListPlusOne( pThis, nil, pCommon, nCommon );

                // call the decomposition recursively
                pDENew = dsdKernelDecompose_rec( pDsdMan, bFuncNew );
//              assert( !Dsd_IsComplement(pDENew) ); // follows from the consideration of cases
                Cudd_RecursiveDeref( dd, bFuncNew );

                // add the first component
                pThis.pDecs[0] = pDENew;
                
                if pL != pLR {
                    // inlining of macro Dsd_Not
                    pThis = ((Dsd_Node_t *)((ABC_PTRUINT_T)(pThis) ^ 01)) ; }
                goto EXIT;
            
            } else
            if pLR.Type == DSD_NODE_EXOR { // Figure 3 of Matsunaga's paper
            
                // compute the sum total of the common components and the union of their supports
                var bCommF, bFuncNew *DdNode;
                var pDENew *Dsd_Node_t;
                var fCompExor int32;

                dsdKernelComputeSumOfComponents( pDsdMan, pCommon, nCommon, &bCommF, nil, 1 );
                Cudd_Ref( bCommF );

                bFuncNew = Cudd_bddXor( dd, bF, bCommF ); Cudd_Ref( bFuncNew );
                Cudd_RecursiveDeref( dd, bCommF );

                // get the new gate

                // copy the components first, then call the decomposition
                // because decomposition will distroy the list used for copying
                pThis = Dsd_TreeNodeCreate( DSD_NODE_EXOR, nCommon + 1, s_nDecBlocks++ );
                dsdKernelCopyListPlusOne( pThis, nil, pCommon, nCommon );

                // call the decomposition recursively
                pDENew = dsdKernelDecompose_rec( pDsdMan, bFuncNew );
                Cudd_RecursiveDeref( dd, bFuncNew );

                // remember the fact that it was complemented
                // inlining of macro Dsd_IsComplement
                fCompExor = ((int32(((ABC_PTRUINT_T) (pDENew) & 01))));
                // inlining of macro Dsd_Regular
                pDENew = ((Dsd_Node_t *)((ABC_PTRUINT_T)(pDENew) & ~01)) ;

                // add the first component
                pThis.pDecs[0] = pDENew;

    
                if fCompExor != 0 {
                    // inlining of macro Dsd_Not
                    pThis = ((Dsd_Node_t *)((ABC_PTRUINT_T)(pThis) ^ 01)) ; }
                goto EXIT;
            
            } else 
            if pLR.Type == DSD_NODE_PRIME && (nCommon == pLR.nDecs-1 || nCommon == pLR.nDecs) {
                // for example the function F(a,b,c,d) = ITE(b,c,a(+)d) produces
                // two cofactors F(a=0) = PRIME(b,c,d) and F(a=1) = PRIME(b,c,d)
                // with exactly the same list of common components

                var pDENew *Dsd_Node_t;
                var bFuncNew *DdNode;
                var fCompComp int32;  // this flag can be {0,1,2}
                // if it is 0 there is no identity
                // if it is 1/2, the cofactored functions are equal in the direct/complemented polarity

                if nCommon == pLR.nDecs {   // all the components are the same
                    // find the formal input, in which pLow and pHigh differ (if such input exists)
                    var m int32;
                    var pTempL, pTempH *Dsd_Node_t;

                    s_Common++;
                    for  m = 0; m < pLR.nDecs; m++  {
                        pTempL = pLR.pDecs[m]; // cannot be complemented
                        pTempH = pHR.pDecs[m]; // cannot be complemented

                        // inlining of macro Cudd_Not
                        if Dsd_CheckRootFunctionIdentity( dd, bLow, bHigh,          pTempL.G, ((DdNode *)((ptrint)(pTempH.G) ^ 01)) ) &&
                             // inlining of macro Cudd_Not
                             Dsd_CheckRootFunctionIdentity( dd, bLow, bHigh, ((DdNode *)((ptrint)(pTempL.G) ^ 01)),         pTempH.G ) {
                             pLastDiffL = pTempL;
                             pLastDiffH = pTempH;
                             assert(pLastDiffL == pLastDiffH, "pLastDiffL == pLastDiffH");
                             fCompComp = 2;
                             break;
                        }

                        s_Loops2++;
                        s_Loops2++;
/* 
                        if ( s_Loops2 % 10000  == 0 )
                        {
                            int i;
                            for ( i = 0; i < pLR->nDecs; i++ )
                                printf( " %d(s=%d)", pLR->pDecs[i]->Type,
                                    Extra_bddSuppSize(dd, pLR->pDecs[i]->S) );
                            printf( "\n" );
                        }
*/

                    }
//                    if ( pLR->nDecs == Extra_bddSuppSize(dd, pLR->S) )
//                        s_Loops2Useless += pLR->nDecs * 2;

                    if fCompComp != 0 { // put the equal components into pCommon, so that they could be copied into the new dec entry
                        nCommon = 0;
                        for  m = 0; m < pLR.nDecs; m++  {
                            if pLR.pDecs[m] != pLastDiffL {
                                 pCommon[nCommon++] = pLR.pDecs[m]; }
                        }
                        assert(nCommon == pLR.nDecs-1, "nCommon == pLR.nDecs-1");
                    }
                
                } else {  // the differing components are known - check that they have compatible PRIME function

                    s_CommonNo++;

                    // find the numbers of different components
                    assert((pLastDiffL != nil), "pLastDiffL");
                    assert((pLastDiffH != nil), "pLastDiffH");
                    // also, they cannot be complemented, because the decomposition type is PRIME

                    // inlining of macro Cudd_Not
                    if Dsd_CheckRootFunctionIdentity( dd, bLow, bHigh, ((DdNode *)((ptrint)(pLastDiffL.G) ^ 01)), ((DdNode *)((ptrint)(pLastDiffH.G) ^ 01)) ) &&
                         Dsd_CheckRootFunctionIdentity( dd, bLow, bHigh,          pLastDiffL.G,           pLastDiffH.G ) {
                        fCompComp = 1;
                    // inlining of macro Cudd_Not
                    } else if Dsd_CheckRootFunctionIdentity( dd, bLow, bHigh,          pLastDiffL.G, ((DdNode *)((ptrint)(pLastDiffH.G) ^ 01)) ) &&
                              // inlining of macro Cudd_Not
                              Dsd_CheckRootFunctionIdentity( dd, bLow, bHigh, ((DdNode *)((ptrint)(pLastDiffL.G) ^ 01)),         pLastDiffH.G ) {
                        fCompComp = 2; }

                    s_Loops3 += 4;
                }

                if fCompComp != 0 {
                    if fCompComp == 1 { // it is true that bLow(G=0) == bHigh(H=0) && bLow(G=1) == bHigh(H=1)
                        bFuncNew = Cudd_bddIte( dd, bVarCur, pLastDiffH.G, pLastDiffL.G ); 
                    } else { // it is true that bLow(G=0) == bHigh(H=1) && bLow(G=1) == bHigh(H=0)
                        // inlining of macro Cudd_Not
                        bFuncNew = Cudd_bddIte( dd, bVarCur, ((DdNode *)((ptrint)(pLastDiffH.G) ^ 01)), pLastDiffL.G ); } 
                    Cudd_Ref( bFuncNew );

                    // get the new gate

                    // copy the components first, then call the decomposition
                    // because decomposition will distroy the list used for copying
                    pThis = Dsd_TreeNodeCreate( DSD_NODE_PRIME, pLR.nDecs, s_nDecBlocks++ );
                    dsdKernelCopyListPlusOne( pThis, nil, pCommon, nCommon );

                    // create a new component
                    pDENew = dsdKernelDecompose_rec( pDsdMan, bFuncNew );
                    Cudd_RecursiveDeref( dd, bFuncNew );
                    // the BDD of the argument function in PRIME decomposition, should be regular
                    // inlining of macro Dsd_Regular
                    pDENew = ((Dsd_Node_t *)((ABC_PTRUINT_T)(pDENew) & ~01)) ;

                    // add the first component
                    pThis.pDecs[0] = pDENew;
                    goto EXIT;
                }
            } // end of PRIME type
        } // end of existing common components
    } // end of CASE 3.a

// if ( Depth != 1) 
// {

//CASE4:
    //////////////////////////////////////////////////////////////////////
    // CASE 4
    //////////////////////////////////////////////////////////////////////
    {
    // estimate the number of entries in the list
    var nEntriesMax int32  = pDsdMan.nInputs - dd.perm[VarInt];

    // create the new decomposition entry
    var nEntries int32;

    var SuppL, SuppH, SuppL_init, SuppH_init *DdNode;
    var pHigher *Dsd_Node_t  = nil; // Suppress "might be used uninitialized"
        var pLower, pTemp, pDENew *Dsd_Node_t;


    var levTopSuppL int32;
    var levTopSuppH int32;
    var levTop int32;

    pThis = Dsd_TreeNodeCreate( DSD_NODE_PRIME, nEntriesMax, s_nDecBlocks++ );
    pThis.pDecs[ nEntries++ ] = pVarCurDE;
    // other entries will be added to this list one-by-one during analysis

    // count how many times does it happen that the decomposition entries are
    s_Case4Calls++;
 
    // consider the simplest case: when the supports are equal 
    // and at least one of the components
    // is the PRIME without decompositions, or 
    // when both of them are without decomposition
    if (((pLR.Type == DSD_NODE_PRIME && nSuppL == pLR.nDecs) || (pHR.Type == DSD_NODE_PRIME && nSuppH == pHR.nDecs)) && pLR.S == pHR.S)  ||
          ((pLR.Type == DSD_NODE_PRIME && nSuppL == pLR.nDecs) && (pHR.Type == DSD_NODE_PRIME && nSuppH == pHR.nDecs)) {

         s_Case4CallsSpecial++;
         // walk through both supports and create the decomposition list composed of simple entries
         SuppL = pLR.S;
         SuppH = pHR.S;
         for
         {
             // determine levels
             // inlining of macro cuddI
             
             if (SuppL.index)==CUDD_CONST_INDEX {
                 levTopSuppL = int32((SuppL.index));
             } else {
                 levTopSuppL = (dd).perm[(SuppL.index)];
             }
             // inlining of macro cuddI
             
             if (SuppH.index)==CUDD_CONST_INDEX {
                 levTopSuppH = int32((SuppH.index));
             } else {
                 levTopSuppH = (dd).perm[(SuppH.index)];
             }

             // skip the topmost variable in both supports
             if levTopSuppL <= levTopSuppH {
                 levTop = levTopSuppL;
                 // inlining of macro cuddT
                 SuppL  = ((SuppL).type.kids.T);
             
             } else {
                 levTop = levTopSuppH; }

             if levTopSuppH <= levTopSuppL {
                 // inlining of macro cuddT
                 SuppH = ((SuppH).type.kids.T); }

             // set the new decomposition entry
             pThis.pDecs[ nEntries++ ] = pDsdMan.pInputs[ dd.invperm[levTop] ];
         
if !( SuppL != b1 || SuppH != b1 ) { break }
}
         
    
    } else {

        // compare two different decomposition lists
        SuppL_init = pLR.S;
        SuppH_init = pHR.S;
        // start references (because these supports will change)
        SuppL = pLR.S;  Cudd_Ref( SuppL );
        SuppH = pHR.S;  Cudd_Ref( SuppH );
        for  SuppL != b1 || SuppH != b1  {
            // determine the top level in cofactors and
            // whether they have the same top level
            var TopLevL int32;
// inlining of macro cuddI

            if (SuppL.index)==CUDD_CONST_INDEX {
                TopLevL = int32((SuppL.index));
            } else {
                TopLevL = (dd).perm[(SuppL.index)];
            }
            var TopLevH int32;
// inlining of macro cuddI

            if (SuppH.index)==CUDD_CONST_INDEX {
                TopLevH = int32((SuppH.index));
            } else {
                TopLevH = (dd).perm[(SuppH.index)];
            }
            var TopLevel int32  = TopLevH;
            var fEqualLevel int32;

            var bVarTop *DdNode;
            var bSuppSubract *DdNode;


            if TopLevL < TopLevH {
                pHigher = pLR;
                pLower  = pHR;
                TopLevel = TopLevL;
            
            } else if TopLevL > TopLevH {
                pHigher = pHR;
                pLower  = pLR;
            
            } else {
                fEqualLevel = 1; }
            assert(TopLevel != CUDD_CONST_INDEX, "TopLevel != CUDD_CONST_INDEX");


            // find the currently top variable in the decomposition lists
            bVarTop = dd.vars[dd.invperm[TopLevel]];

            if !(fEqualLevel != 0) {
                // find the lower support
                var bSuppLower *DdNode;
                if TopLevL < TopLevH {
                    bSuppLower = SuppH_init;
                } else {
                    bSuppLower = SuppL_init;
                } 

                // find the first component in pHigher 
                // whose support does not overlap with supp(Lower) 
                // and remember the previous component
                var fPolarity int32;          
                var pPrev *Dsd_Node_t  = nil;       // the pointer to the component proceeding pCur
                var pCur *Dsd_Node_t   = pHigher;    // the first component not contained in supp(Lower)
                for  Extra_bddSuppOverlapping( dd, pCur.S, bSuppLower )  {   // get the next component
                    pPrev = pCur;
                    pCur  = dsdKernelFindContainingComponent( pDsdMan, pCur, bVarTop, &fPolarity );
                };

                // look for the possibility to subtract more than one component
                if pPrev == nil || pPrev.Type == DSD_NODE_PRIME { // if there is no previous component, or if the previous component is PRIME
                  // there is no way to subtract more than one component

                    // add the new decomposition entry (it is already regular)
                    pThis.pDecs[ nEntries++ ] = pCur;
                    // assign the support to be subtracted from both components
                    bSuppSubract = pCur.S;
                
                } else { // all other types
                
                    // go through the decomposition list of pPrev and find components 
                    // whose support does not overlap with supp(Lower) 

                    
                    var i, nNonOverlap int32;
                    for  i = 0; i < pPrev.nDecs; i++  {
                        // inlining of macro Dsd_Regular
                        pTemp = ((Dsd_Node_t *)((ABC_PTRUINT_T)(pPrev.pDecs[i]) & ~01)) ;
                        if !Extra_bddSuppOverlapping( dd, pTemp.S, bSuppLower ) {
                            spStatic_dsdKernelDecompose_rec_pNonOverlap[ nNonOverlap++ ] = pPrev.pDecs[i]; }
                    }
                    assert(nNonOverlap > 0, "nNonOverlap > 0");

                    if nNonOverlap == 1 { // one one component was found, which is the original one
                        // inlining of macro Dsd_Regular
                        assert(((Dsd_Node_t *)((ABC_PTRUINT_T)(spStatic_dsdKernelDecompose_rec_pNonOverlap[0]) & ~01))  == pCur, "/*__starting_point_inline_macro:Dsd_Regular__*/((Dsd_Node_t *)((ABC_PTRUINT_T)(spStatic_dsdKernelDecompose_rec_pNonOverlap[0]) & ~01))  == pCur");
                        // add the new decomposition entry
                        pThis.pDecs[ nEntries++ ] = pCur;
                        // assign the support to be subtracted from both components
                        bSuppSubract = pCur.S;
                    
                    } else { // more than one components was found
                    
                        // find the OR (EXOR) of the non-overlapping components
                        var bCommF *DdNode;
                        dsdKernelComputeSumOfComponents( pDsdMan, spStatic_dsdKernelDecompose_rec_pNonOverlap, nNonOverlap, &bCommF, nil, int32((pPrev.Type==DSD_NODE_EXOR)) );
                        Cudd_Ref( bCommF );

                        // create a new gated 
                        pDENew = dsdKernelDecompose_rec( pDsdMan, bCommF );
                        Cudd_RecursiveDeref(dd, bCommF);
                        // make it regular... it must be regular already
                        // inlining of macro Dsd_IsComplement
                        assert(!(((int32(((ABC_PTRUINT_T) (pDENew) & 01)) != 0))), "!/*__starting_point_inline_macro:Dsd_IsComplement__*/((int32(((ABC_PTRUINT_T) (pDENew) & 01))))");

                        // add the new decomposition entry
                        pThis.pDecs[ nEntries++ ] = pDENew;
                        // assign the support to be subtracted from both components
                        bSuppSubract = pDENew.S;
                    }
                }
                
                // subtract its support from the support of upper component
                if TopLevL < TopLevH {
                    SuppL = Cudd_bddExistAbstract( dd, bTemp = SuppL, bSuppSubract ); Cudd_Ref( SuppL );
                    Cudd_RecursiveDeref(dd, bTemp);
                
                } else {
                    SuppH = Cudd_bddExistAbstract( dd, bTemp = SuppH, bSuppSubract ); Cudd_Ref( SuppH );
                    Cudd_RecursiveDeref(dd, bTemp);
                }
             // end of if ( !fEqualLevel )
            } else { // if ( fEqualLevel ) -- they have the same top level var
            
                 // the pointers to the marked blocks
                 // polarities of the marked blocks
                var nMarkedLeft int32;

                var fPolarity int32;
                var pTempL *Dsd_Node_t  = pLR;

                var fPolarityCurH int32;
                var pPrevH *Dsd_Node_t = nil;
                var pCurH *Dsd_Node_t = pHR;

                var fPolarityCurL int32;
                var pPrevL *Dsd_Node_t = nil;
                var pCurL *Dsd_Node_t = pLR; // = pMarkedLeft[0];
                var index int32  = 1;

                // set the new mark
                s_Mark++;

                // go over the dec list of pL, mark all components that contain the given variable
                assert(Extra_bddSuppContainVar( dd, pLR.S, bVarTop ), "Extra_bddSuppContainVar( dd, pLR.S, bVarTop )");
                assert(Extra_bddSuppContainVar( dd, pHR.S, bVarTop ), "Extra_bddSuppContainVar( dd, pHR.S, bVarTop )");
                for {
                    pTempL.Mark = s_Mark;
                    spStatic_dsdKernelDecompose_rec_pMarkedLeft[ nMarkedLeft ] = pTempL;
                    spStatic_dsdKernelDecompose_rec_pMarkedPols[ nMarkedLeft ] = fPolarity;
                    nMarkedLeft++;
                
if !( (pTempL = dsdKernelFindContainingComponent( pDsdMan, pTempL, bVarTop, &fPolarity )) ) { break }
} 

                // go over the dec list of pH, and find the component that is marked and the previos one
                // (such component always exists, because they have common variables)
                for  pCurH.Mark != s_Mark  {
                    pPrevH = pCurH;
                    pCurH  = dsdKernelFindContainingComponent( pDsdMan, pCurH, bVarTop, &fPolarityCurH );
                    assert((pCurH != nil), "pCurH");
                }

                // go through the first list once again and find 
                // the component proceeding the one marked found in the second list
                for  pCurL != pCurH  {
                    pPrevL = pCurL;
                    pCurL  = spStatic_dsdKernelDecompose_rec_pMarkedLeft[index];
                    fPolarityCurL = spStatic_dsdKernelDecompose_rec_pMarkedPols[index];
                    index++;
                }

                // look for the possibility to subtract more than one component
                if !(pPrevL != nil) || !(pPrevH != nil) || pPrevL.Type != pPrevH.Type || pPrevL.Type == DSD_NODE_PRIME || fPolarityCurL != fPolarityCurH { // there is no way to extract more than one
                    pThis.pDecs[ nEntries++ ] = pCurH;
                    // assign the support to be subtracted from both components
                    bSuppSubract = pCurH.S;
                
                } else {
                    // find the equal components in two decomposition lists
                    var pCommon **Dsd_Node_t;
                    var pLastDiffL *Dsd_Node_t = nil;
                    var pLastDiffH *Dsd_Node_t = nil; 
                    var nCommon int32  = dsdKernelFindCommonComponents( pDsdMan, pPrevL, pPrevH, &pCommon, &pLastDiffL, &pLastDiffH );
        
                    if nCommon == 0 || nCommon == 1 { // one one component was found, which is the original one
    //                  assert( Dsd_Regular(pCommon[0]) == pCurL);
                        // add the new decomposition entry
                        pThis.pDecs[ nEntries++ ] = pCurL;
                        // assign the support to be subtracted from both components
                        bSuppSubract = pCurL.S;
                    
                    } else { // more than one components was found
                    
                        // find the OR (EXOR) of the non-overlapping components
                        var bCommF *DdNode;
                        dsdKernelComputeSumOfComponents( pDsdMan, pCommon, nCommon, &bCommF, nil, int32((pPrevL.Type==DSD_NODE_EXOR)) );
                        Cudd_Ref( bCommF );

                        pDENew = dsdKernelDecompose_rec( pDsdMan, bCommF );
                        // inlining of macro Dsd_IsComplement
                        assert(!(((int32(((ABC_PTRUINT_T) (pDENew) & 01)) != 0))), "!/*__starting_point_inline_macro:Dsd_IsComplement__*/((int32(((ABC_PTRUINT_T) (pDENew) & 01))))"); // cannot be complemented because of construction
                        Cudd_RecursiveDeref( dd, bCommF );

                        // add the new decomposition entry
                        pThis.pDecs[ nEntries++ ] = pDENew;

                        // assign the support to be subtracted from both components
                        bSuppSubract = pDENew.S;
                    }
                }

                SuppL = Cudd_bddExistAbstract( dd, bTemp = SuppL, bSuppSubract );
Cudd_Ref( SuppL );
                Cudd_RecursiveDeref(dd, bTemp);

                SuppH = Cudd_bddExistAbstract( dd, bTemp = SuppH, bSuppSubract );
Cudd_Ref( SuppH );
                Cudd_RecursiveDeref(dd, bTemp);

            } // end of if ( fEqualLevel ) 

        } // end of decomposition list comparison
        Cudd_RecursiveDeref( dd, SuppL );
        Cudd_RecursiveDeref( dd, SuppH );

    }

    // check that the estimation of the number of entries was okay
    assert(nEntries <= nEntriesMax, "nEntries <= nEntriesMax");

//    if ( nEntries != Extra_bddSuppSize(dd, bSuppNew) )
//        s_Case5++;

    // update the number of entries in the new decomposition list
    pThis.nDecs = nEntries;
    }
//}
EXIT:

    {
    // if the component created is complemented, it represents a function without complement
    // therefore, as it is, without complement, it should recieve the complemented function
    // inlining of macro Dsd_Regular
    var pThisR *Dsd_Node_t  = ((Dsd_Node_t *)((ABC_PTRUINT_T)(pThis) & ~01)) ;
    assert(pThisR.G == nil, "pThisR.G == nil");
    assert(pThisR.S == nil, "pThisR.S == nil");

    if pThisR == pThis { // set regular function
        pThisR.G = bF; 
    } else { // set complemented function
        // inlining of macro Cudd_Not
        pThisR.G = ((DdNode *)((ptrint)(bF) ^ 01)); }    
    Cudd_Ref(bF);           // reference the function in the component

    assert((bSuppNew != nil), "bSuppNew");
    pThisR.S = bSuppNew;   // takes the reference from the new support
    if st__insert( pDsdMan.Table, (char*)bF, (char*)pThis ) {
        assert(false, "0");
    }
    s_CacheEntries++;


/*
    if ( dsdKernelVerifyDecomposition(dd, pThis) == 0 )
    {
        // write the function, for which verification does not work
        cout << endl << "Internal verification failed!"" );

        // create the variable mask
        static int s_pVarMask[MAXINPUTS];
        int nInputCounter = 0;

        Cudd_SupportArray( dd, bF, s_pVarMask );
        int k; 
        for ( k = 0; k < dd->size; k++ )
            if ( s_pVarMask[k] )
                nInputCounter++;

        cout << endl << "The problem function is "" );

        DdNode * zNewFunc = Cudd_zddIsopCover( dd, bF, bF ); Cudd_Ref( zNewFunc );
        cuddWriteFunctionSop( stdout, dd, zNewFunc, -1, dd->size, "1", s_pVarMask );
        Cudd_RecursiveDerefZdd( dd, zNewFunc );
    }
*/

    }

    Depth--;
    // inlining of macro Dsd_NotCond
    return ((Dsd_Node_t *)((ABC_PTRUINT_T)(pThis) ^ (fCompF)));
}


////////////////////////////////////////////////////////////////////////
///                        OTHER FUNCTIONS                           ///
////////////////////////////////////////////////////////////////////////

/**Function*************************************************************

  Synopsis    [Finds the corresponding decomposition entry.]

  Description [This function returns the non-complemented pointer to the 
  DecEntry of that component which contains the given variable in its 
  support, or NULL if no such component exists]
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func dsdKernelFindContainingComponent(pDsdMan *Dsd_Manager_t, pWhere *Dsd_Node_t, Var *DdNode, fPolarity *int32) *Dsd_Node_t {
    var pTemp *Dsd_Node_t;
    var i int32;

//  assert( !Dsd_IsComplement( pWhere ) );
//  assert( Extra_bddSuppContainVar( pDsdMan->dd, pWhere->S, Var ) );

    if pWhere.nDecs == 1 {
        return nil; }

    for  i = 0; i < pWhere.nDecs; i++  {
        // inlining of macro Dsd_Regular
        pTemp = ((Dsd_Node_t *)((ABC_PTRUINT_T)(pWhere.pDecs[i]) & ~01)) ;
        if Extra_bddSuppContainVar( pDsdMan.dd, pTemp.S, Var ) {
            *fPolarity = int32(( pTemp != pWhere.pDecs[i] ));
            return pTemp;
        }
    }
    assert(false, "0");
    return nil;
}

/**Function*************************************************************

  Synopsis    [Find the common decomposition components.]

  Description [This function determines the common components. It counts 
  the number of common components in the decomposition lists of pL and pH
  and returns their number and the lists of common components. It assumes 
  that pL and pH are regular pointers. It retuns also the pointers to the 
  last different components encountered in pL and pH.]
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
var spStatic_dsdKernelFindCommonComponents_Common [MAXINPUTS]*Dsd_Node_t;
func dsdKernelFindCommonComponents(pDsdMan *Dsd_Manager_t, pL *Dsd_Node_t, pH *Dsd_Node_t, pCommon ***Dsd_Node_t, pLastDiffL **Dsd_Node_t, pLastDiffH **Dsd_Node_t) int32 {
    
    var nCommon int32;

    // pointers to the current decomposition entries
    var pLcur *Dsd_Node_t;
    var pHcur *Dsd_Node_t;

    // the pointers to their supports
    var bSLcur *DdNode;
    var bSHcur *DdNode;

    // the top variable in the supports
    var TopVar int32;

    // the indices running through the components
    var iCurL int32;
    var iCurH int32;
    for  iCurL < pL.nDecs && iCurH < pH.nDecs  { // both did not run out

        // inlining of macro Dsd_Regular
        pLcur = ((Dsd_Node_t *)((ABC_PTRUINT_T)(pL.pDecs[iCurL]) & ~01)) ;
        // inlining of macro Dsd_Regular
        pHcur = ((Dsd_Node_t *)((ABC_PTRUINT_T)(pH.pDecs[iCurH]) & ~01)) ;

        bSLcur = pLcur.S;
        bSHcur = pHcur.S;

        // find out what component is higher in the BDD
        if pDsdMan.dd.perm[bSLcur.index] < pDsdMan.dd.perm[bSHcur.index] {
            TopVar = bSLcur.index;
        } else {
            TopVar = bSHcur.index; }

        if TopVar == bSLcur.index && TopVar == bSHcur.index {
            // the components may be equal - should match exactly!
            if pL.pDecs[iCurL] == pH.pDecs[iCurH] {
                spStatic_dsdKernelFindCommonComponents_Common[nCommon++] = pL.pDecs[iCurL];
            } else {
                *pLastDiffL = pL.pDecs[iCurL];
                *pLastDiffH = pH.pDecs[iCurH];
            }

            // skip both
            iCurL++;
            iCurH++;
        
        } else if TopVar == bSLcur.index {  // the components cannot be equal
            // skip the top-most one
            *pLastDiffL = pL.pDecs[iCurL++];
        
        } else { // if ( TopVar == bSHcur->index )
          // the components cannot be equal
            // skip the top-most one
            *pLastDiffH = pH.pDecs[iCurH++];
        }
    }

    // if one of the lists still has components, write the first one down
    if iCurL < pL.nDecs {
        *pLastDiffL = pL.pDecs[iCurL]; }

    if iCurH < pH.nDecs {
        *pLastDiffH = pH.pDecs[iCurH]; }

    // return the pointer to the array
    *pCommon = spStatic_dsdKernelFindCommonComponents_Common;
    // return the number of common components
    return nCommon;         
}

/**Function*************************************************************

  Synopsis    [Computes the sum (OR or EXOR) of the functions of the components.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func dsdKernelComputeSumOfComponents(pDsdMan *Dsd_Manager_t, pCommon **Dsd_Node_t, nCommon int32, pCompF **DdNode, pCompS **DdNode, fExor int32) {
    var dd *DdManager  = pDsdMan.dd;
    var bF, bFadd, bTemp *DdNode;
        var bS *DdNode  = nil; // Suppress "might be used uninitialized"
    var pDE, pDER *Dsd_Node_t;
    var i int32;

    // start the function
    bF = b0; Cudd_Ref( bF );
    // start the support
    if pCompS {
        bS = b1;
Cudd_Ref( bS ); }

    assert(nCommon > 0, "nCommon > 0");
    for  i = 0; i < nCommon; i++  {
        pDE  = pCommon[i];
        // inlining of macro Dsd_Regular
        pDER = ((Dsd_Node_t *)((ABC_PTRUINT_T)(pDE) & ~01)) ;
        if pDE != pDER {
            // inlining of macro Cudd_Not
            bFadd = ((DdNode *)((ptrint)(pDER.G) ^ 01));
        } else {
            bFadd = pDER.G;
        }
        // add to the function
        if fExor != 0 {
            bF = Cudd_bddXor( dd, bTemp = bF, bFadd );
        } else {
            bF = Cudd_bddOr( dd, bTemp = bF, bFadd ); }
        Cudd_Ref( bF );
        Cudd_RecursiveDeref( dd, bTemp );
        if pCompS {
            // add to the support
            bS = Cudd_bddAnd( dd, bTemp = bS, pDER.S );  Cudd_Ref( bS );
            Cudd_RecursiveDeref( dd, bTemp );
        }
    }
    // return the function
    Cudd_Deref( bF );
    *pCompF = bF;

    // return the support
    if pCompS {
        Cudd_Deref( bS );
*pCompS = bS; }
}

/**Function*************************************************************

  Synopsis    [Checks support containment of the decomposition components.]

  Description [This function returns 1 if support of one component is contained 
  in that of another. In this case, pLarge (pSmall) is assigned to point to the 
  larger (smaller) support. If the supports are identical return 0, and does not 
  assign the components.]
]
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func dsdKernelCheckContainment(pDsdMan *Dsd_Manager_t, pL *Dsd_Node_t, pH *Dsd_Node_t, pLarge **Dsd_Node_t, pSmall **Dsd_Node_t) int32 {
    var dd *DdManager  = pDsdMan.dd;
    var bSuppLarge, bSuppSmall *DdNode;
    var RetValue int32;
    
    RetValue = Extra_bddSuppCheckContainment( dd, pL.S, pH.S, &bSuppLarge, &bSuppSmall );

    if RetValue == 0 { 
        return 0; }

    if pH.S == bSuppLarge {
        *pLarge = pH;
        *pSmall = pL;
    
    } else { // if ( pL->S == bSuppLarge )
    
        *pLarge = pL;
        *pSmall = pH;
    }
    return 1;
}

/**Function*************************************************************

  Synopsis    [Copies the list of components plus one.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func dsdKernelCopyListPlusOne(p *Dsd_Node_t, First *Dsd_Node_t, ppList **Dsd_Node_t, nListSize int32) {
    var i int32;
    assert(nListSize+1 == p.nDecs, "nListSize+1 == p.nDecs");
    p.pDecs[0] = First;
    for  i = 0; i < nListSize; i++  {
        p.pDecs[i+1] = ppList[i];
    }
}

/**Function*************************************************************

  Synopsis    [Copies the list of components plus one, and skips one.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func dsdKernelCopyListPlusOneMinusOne(p *Dsd_Node_t, First *Dsd_Node_t, ppList **Dsd_Node_t, nListSize int32, iSkipped int32) {
    var i, Counter int32;
    assert(nListSize == p.nDecs, "nListSize == p.nDecs");
    p.pDecs[0] = First;
    for  i, Counter = 0, 1; i < nListSize; i++  {
        if i != iSkipped {
            p.pDecs[Counter++] = ppList[i]; }
    }
}

/**Function*************************************************************

  Synopsis    [Debugging procedure to compute the functionality of the decomposed structure.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
var spStatic_dsdKernelVerifyDecomposition_bGVars [MAXINPUTS]*DdNode;
func dsdKernelVerifyDecomposition(pDsdMan *Dsd_Manager_t, pDE *Dsd_Node_t) int32 {
    var dd *DdManager  = pDsdMan.dd;
    // inlining of macro Dsd_Regular
    var pR *Dsd_Node_t     = ((Dsd_Node_t *)((ABC_PTRUINT_T)(pDE) & ~01)) ;
    var RetValue int32;

    var bRes *DdNode;
    if pR.Type == DSD_NODE_CONST1 {
        bRes = b1;
    } else if pR.Type == DSD_NODE_BUF {
        bRes = pR.G;
    } else if pR.Type == DSD_NODE_OR || pR.Type == DSD_NODE_EXOR {
        dsdKernelComputeSumOfComponents( pDsdMan, pR.pDecs, pR.nDecs, &bRes, nil, int32((pR.Type == DSD_NODE_EXOR)) );
    } else if pR.Type == DSD_NODE_PRIME {
        var i int32;
        
        // transform the function of this block, so that it depended on inputs
        // corresponding to the formal inputs
        var bNewFunc *DdNode  = Dsd_TreeGetPrimeFunctionOld( dd, pR, 1 );  Cudd_Ref( bNewFunc );

        // compose this function with the inputs
        // create the elementary permutation
        for  i = 0; i < dd.size; i++  {
            spStatic_dsdKernelVerifyDecomposition_bGVars[i] = dd.vars[i];
        }

        // assign functions to be composed
        for  i = 0; i < pR.nDecs; i++  {
            spStatic_dsdKernelVerifyDecomposition_bGVars[dd.invperm[i]] = pR.pDecs[i].G;
        }

        // perform the composition
        bRes = Cudd_bddVectorCompose( dd, bNewFunc, spStatic_dsdKernelVerifyDecomposition_bGVars );      Cudd_Ref( bRes );
        Cudd_RecursiveDeref( dd, bNewFunc );

        /////////////////////////////////////////////////////////
        RetValue = int32(( bRes == pR.G ));//|| bRes == Cudd_Not(pR->G) );
        /////////////////////////////////////////////////////////
        Cudd_Deref( bRes );
    
    } else {
        assert(false, "0");
    }

    Cudd_Ref( bRes );
    RetValue = int32(( bRes == pR.G ));//|| bRes == Cudd_Not(pR->G) );
    Cudd_RecursiveDeref( dd, bRes );
    return RetValue;
}

////////////////////////////////////////////////////////////////////////
///                           END OF FILE                            ///
////////////////////////////////////////////////////////////////////////
//ABC_NAMESPACE_IMPL_END

