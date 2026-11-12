package extra

import "fmt"

/**CFile****************************************************************

  FileName    [extraZddTrunc.c]

  PackageName [extra]

  Synopsis    [Procedure to truncate a ZDD using variable probabilities.]

  Author      [Alan Mishchenko]

  Affiliation [UC Berkeley]

  Date        [Ver. 2.0. Started - September 1, 2003.]

  Revision    [$Id: extraZddTrunc.c,v 1.0 2003/05/21 18:03:50 alanmi Exp $]

***********************************************************************/

//#include <stdio.h>
//#include <stdlib.h>
//#include <string.h>
//#include <assert.h>

//#include "misc/st/st.h"
//#include "bdd/cudd/cuddInt.h"

#ifdef _WIN32
#define inline __inline // compatible with MS VS 6.0
#endif

//ABC_NAMESPACE_IMPL_START


const TEST_VAR_MAX = 10
const TEST_SET_MAX = 10

/*---------------------------------------------------------------------------*/
/* Constant declarations                                                     */
/*---------------------------------------------------------------------------*/

/*---------------------------------------------------------------------------*/
/* Stucture declarations                                                     */
/*---------------------------------------------------------------------------*/

/*---------------------------------------------------------------------------*/
/* Type declarations                                                         */
/*---------------------------------------------------------------------------*/

/*---------------------------------------------------------------------------*/
/* Variable declarations                                                     */
/*---------------------------------------------------------------------------*/

/*---------------------------------------------------------------------------*/
/* Macro declarations                                                        */
/*---------------------------------------------------------------------------*/


/**AutomaticStart*************************************************************/

/*---------------------------------------------------------------------------*/
/* Static function prototypes                                                */
/*---------------------------------------------------------------------------*/


// dynamic vector of intergers
         
type Vec_Int_t struct {
    nCap              int32;
    nSize              int32;
    pArray             *int32;
}
func Vec_IntAlloc(nCap int32) *Vec_Int_t {
    var p *Vec_Int_t;
    p = new(Vec_Int_t);
    if nCap > 0 && nCap < 16 {
        nCap = 16; }
    p.nSize  = 0;
    p.nCap   = nCap;
    if p.nCap {
        p.pArray = ABC_ALLOC( int, p.nCap );
    } else {
        p.pArray = nil;
    }
    return p;
}
func Vec_IntFree(p *Vec_Int_t) {
    ABC_FREE( p.pArray );
    ABC_FREE( p );
}
func Vec_IntReleaseArray(p *Vec_Int_t) *int32 {
    var pArray *int32  = p.pArray;
    p.nCap = 0;
    p.nSize = 0;
    p.pArray = nil;
    return pArray;
}
func Vec_IntAddToEntry(p *Vec_Int_t, i int32, Addition int32) int32 {
    assert(i >= 0 && i < p.nSize, "i >= 0 && i < p.nSize");
    return p.pArray[i] += Addition;
}
func Vec_IntGrow(p *Vec_Int_t, nCapMin int32) {
    if p.nCap >= nCapMin {
        return; }
    p.pArray = ABC_REALLOC( int, p.pArray, nCapMin ); 
    assert(p.pArray, "p.pArray");
    p.nCap   = nCapMin;
}
func Vec_IntPop(p *Vec_Int_t) int32 {
    assert(p.nSize > 0, "p.nSize > 0");
    return p.pArray[--p.nSize];
}
func Vec_IntPush(p *Vec_Int_t, Entry int32) {
    if p.nSize == p.nCap {
        if p.nCap < 16 {
            Vec_IntGrow( p, 16 );
        } else {
            Vec_IntGrow( p, 2 * p.nCap ); }
    }
    p.pArray[p.nSize++] = Entry;
}
func Vec_IntAppend(vVec1 *Vec_Int_t, vVec2 *Vec_Int_t) {
    var i int32;
    for  i = 0; i < vVec2.nSize; i++  {
        Vec_IntPush( vVec1, vVec2.pArray[i] );
    }
}


/**AutomaticEnd***************************************************************/


/*---------------------------------------------------------------------------*/
/* Definition of exported functions                                          */
/*---------------------------------------------------------------------------*/

/**Function********************************************************************

  Synopsis    [Compute the set of subsets whose probability is more than ProbLimit.]

  Description [The resulting array has the following form:  The first integer entry 
  is the number of resulting subsets.  The following integer entries in the array
  contain as many subsets. Each subset is an array of integers followed by -1. 
  See how subsets are printed in the included test procedure below.]

  SideEffects []

  SeeAlso     []

******************************************************************************/
void Extra_zddTruncate_rec(
  DdManager * dd, 
  DdNode * zFunc,       // zFunc is the ZDD to be truncated
  double * pVarProbs,    // pVarProbs is probabilities of each variable (should have at least dd->sizeZ entries)
  double ProbLimit,      // ProbLimit is the limit on the probabilities (only those more than this will be collected)
  double ProbThis,       // current path probability
  Vec_Int_t * vSubset,  // current subset under construction
  Vec_Int_t * vResult ) // resulting subsets to be returned to the user
{
    // quit if probability of the path is less then the limit
    if ProbThis < ProbLimit {
        return; }
    // quit if there is no subsets
    if zFunc == Cudd_ReadZero(dd) { 
        return; }
    // quit and save a new subset if there is one
    if zFunc == Cudd_ReadOne(dd) {
        Vec_IntAddToEntry( vResult, 0, 1 );
        Vec_IntAppend( vResult, vSubset );
        Vec_IntPush( vResult, -1 );
        return;
    }
    // call recursively for the set without the given variable
    // inlining of macro cuddE
    Extra_zddTruncate_rec( dd, ((zFunc).type.kids.E), pVarProbs, ProbLimit, ProbThis, vSubset, vResult );
    // call recursively for the set with the given variable
    Vec_IntPush( vSubset, Cudd_NodeReadIndex(zFunc) );
    // inlining of macro cuddT
    Extra_zddTruncate_rec( dd, ((zFunc).type.kids.T), pVarProbs, ProbLimit, ProbThis * pVarProbs[Cudd_NodeReadIndex(zFunc)], vSubset, vResult );
    Vec_IntPop( vSubset );
}
int * Extra_zddTruncate( 
  DdManager * dd, 
  DdNode * zFunc,       // zFunc is the ZDD to be truncated
  double * pVarProbs,    // pVarProbs is probabilities of each variable (should have at least dd->sizeZ entries)
  double ProbLimit )     // ProbLimit is the limit on the probabilities (only those more than this will be collected)
{
    var vSubset, vResult *Vec_Int_t;
    var i int32;
    var sizeZ int32 = Cudd_ReadZddSize(dd);
    var pResult *int32;
    // check that probabilities are reasonable
    assert(ProbLimit > 0 && ProbLimit <= 1, "ProbLimit > 0 && ProbLimit <= 1");
    for  i = 0; i < sizeZ; i++  {
        assert(pVarProbs[i] > 0 && pVarProbs[i] <= 1, "pVarProbs[i] > 0 && pVarProbs[i] <= 1");
    }
    // enumerate assignments satisfying the probability limit
    vSubset = Vec_IntAlloc( sizeZ );
    vResult = Vec_IntAlloc( 10 * sizeZ );
    Vec_IntPush( vResult, 0 );
    Extra_zddTruncate_rec( dd, zFunc, pVarProbs, ProbLimit, 1, vSubset, vResult );
    Vec_IntFree( vSubset );
    pResult = Vec_IntReleaseArray( vResult );
    Vec_IntFree( vResult );
    return pResult;
} // end of Extra_zddTruncate 

/**Function*************************************************************

  Synopsis    [Creates the combination composed of a single ZDD variable.]

  Description []

  SideEffects []

  SeeAlso     []

***********************************************************************/
func Extra_zddVariable(dd *DdManager, iVar int32) *DdNode {
    var zRes *DdNode;
    for {
        dd.reordered = 0;
        zRes = cuddZddGetNode( dd, iVar, Cudd_ReadOne(dd), Cudd_ReadZero(dd) ); 
    
if !(dd.reordered == 1) { break }
} 
    return zRes;
}

/**Function********************************************************************

  Synopsis    [Creates ZDD representing a given set of subsets.]

  Description []

  SideEffects []

  SeeAlso     []

******************************************************************************/
func Extra_zddCreateSubsets(dd *DdManager, pSubsets [][TEST_VAR_MAX+1]int32, nSubsets int32) *DdNode {
    var i, k int32;
    var zOne, zVar, zRes, zTemp *DdNode;
    zRes = Cudd_ReadZero(dd); Cudd_Ref( zRes );
    for  i = 0; i < nSubsets; i++  {
        zOne = Cudd_ReadOne(dd); Cudd_Ref( zOne );
        for  k = 0; pSubsets[i][k] != -1; k++  {
            assert(pSubsets[i][k] < TEST_VAR_MAX, "pSubsets[i][k] < TEST_VAR_MAX");
            zVar = Extra_zddVariable( dd, pSubsets[i][k] );
            zOne = Cudd_zddUnateProduct( dd, zTemp = zOne, zVar ); Cudd_Ref( zOne );
            Cudd_RecursiveDerefZdd( dd, zTemp );
        }
        zRes = Cudd_zddUnion( dd, zRes, zOne ); Cudd_Ref( zRes );
        Cudd_RecursiveDerefZdd( dd, zOne );
    }
    Cudd_Deref( zRes );
    return zRes;
}

/**Function********************************************************************

  Synopsis    [Prints a set of subsets represented using as an array.]

  Description []

  SideEffects []

  SeeAlso     []

******************************************************************************/
func Extra_zddPrintSubsets(pSubsets *int32) {
    var i, k, Counter int32;
    fmt.Printf("The set contains %d subsets:\n", pSubsets[0]);
    for  i = k = 0; i < pSubsets[0]; i++  {
        fmt.Printf("Subset %3d : {", Counter);
        for  k++; pSubsets[k] != -1; k++  {
            fmt.Printf(" %d", pSubsets[k]);
        }
        fmt.Printf(" }\n");
        Counter++;
    }
}

/**Function********************************************************************

  Synopsis    [Testbench for the above truncation procedure.]

  Description []

  SideEffects []

  SeeAlso     []

******************************************************************************/
func Extra_zddTruncateTest() {
    // input data
    var nSubsets int32  = 5;
    var pSubsets [TEST_SET_MAX][TEST_VAR_MAX+1]int32 = [TEST_SET_MAX][TEST_VAR_MAX+1]int32{ {0, 3, 5, -1}, {1, 2, 3, 6, 9, -1}, {1, 5, 7, 8, -1}, {2, 4, -1}, {0, 5, 6, 9, -1} };
    // varible probabilities
    var pVarProbs [TEST_VAR_MAX]float64 = [TEST_VAR_MAX]float64{ 0.1, 0.1, 0.1, 0.1, 0.1, 0.1, 0.1, 0.1, 0.1, 0.1 };
    var ProbLimit float64  = 0.001;
    // output data
    var pOutput *int32;
    // start the manager and create ZDD representing the input subsets
    var dd *DdManager  = Cudd_Init( 0, TEST_VAR_MAX, CUDD_UNIQUE_SLOTS, CUDD_CACHE_SLOTS , 0 );
    var zFunc *DdNode  = Extra_zddCreateSubsets( dd, pSubsets, nSubsets );  Cudd_Ref( zFunc );
    assert(nSubsets <= TEST_SET_MAX, "nSubsets <= TEST_SET_MAX");
    // print the input ZDD
    fmt.Printf("The initial ZDD representing %d subsets:\n", nSubsets);
    Cudd_zddPrintMinterm( dd, zFunc );
    // compute the result of truncation
    pOutput = Extra_zddTruncate( dd, zFunc, pVarProbs, ProbLimit );
    // print the resulting ZDD
    fmt.Printf("The resulting ZDD representing %d subsets:\n", pOutput[0]);
    // print the resulting subsets
    Extra_zddPrintSubsets( pOutput );
    // cleanup
    ABC_FREE( pOutput );
    Cudd_RecursiveDerefZdd( dd, zFunc );
    Cudd_Quit( dd );
} 




////////////////////////////////////////////////////////////////////////
///                           END OF FILE                            ///
////////////////////////////////////////////////////////////////////////
//ABC_NAMESPACE_IMPL_END

