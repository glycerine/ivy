package msat

/**CFile****************************************************************

  FileName    [msatClauseVec.c]

  PackageName [A C version of SAT solver MINISAT, originally developed 
  in C++ by Niklas Een and Niklas Sorensson, Chalmers University of 
  Technology, Sweden: http://www.cs.chalmers.se/~een/Satzoo.]

  Synopsis    [Procedures working with arrays of SAT clauses.]

  Author      [Alan Mishchenko <alanmi@eecs.berkeley.edu>]
  
  Affiliation [UC Berkeley]

  Date        [Ver. 1.0. Started - January 1, 2004.]

  Revision    [$Id: msatClauseVec.c,v 1.0 2004/01/01 1:00:00 alanmi Exp $]

***********************************************************************/

//#include "msatInt.h"

//ABC_NAMESPACE_IMPL_START


////////////////////////////////////////////////////////////////////////
///                        DECLARATIONS                              ///
////////////////////////////////////////////////////////////////////////

////////////////////////////////////////////////////////////////////////
///                     FUNCTION DEFINITIONS                         ///
////////////////////////////////////////////////////////////////////////

/**Function*************************************************************

  Synopsis    [Allocates a vector with the given capacity.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Msat_ClauseVecAlloc(nCap int32) *Msat_ClauseVec_t {
    var p *Msat_ClauseVec_t;
    // inlining of macro ABC_ALLOC
    p = ((Msat_ClauseVec_t *) malloc(sizeof(Msat_ClauseVec_t) * (size_t)(1)));
    if nCap > 0 && nCap < 16 {
        nCap = 16; }
    p.nSize  = 0;
    p.nCap   = nCap;
    if p.nCap {
        // inlining of macro ABC_ALLOC
        p.pArray = ((Msat_Clause_t * *) malloc(sizeof(Msat_Clause_t *) * (size_t)(p.nCap)));
    } else {
        p.pArray = nil;
    }
    return p;
}

/**Function*************************************************************

  Synopsis    []

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Msat_ClauseVecFree(p *Msat_ClauseVec_t) {
    // inlining of macro ABC_FREE
    ((p.pArray) ? (free((char *) (p.pArray)), (p.pArray) = 0) : 0);
    // inlining of macro ABC_FREE
    ((p) ? (free((char *) (p)), (p) = 0) : 0);
}

/**Function*************************************************************

  Synopsis    []

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Msat_ClauseVecReadArray(p *Msat_ClauseVec_t) **Msat_Clause_t {
    return p.pArray;
}

/**Function*************************************************************

  Synopsis    []

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Msat_ClauseVecReadSize(p *Msat_ClauseVec_t) int32 {
    return p.nSize;
}

/**Function*************************************************************

  Synopsis    [Resizes the vector to the given capacity.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Msat_ClauseVecGrow(p *Msat_ClauseVec_t, nCapMin int32) {
    if p.nCap >= nCapMin {
        return; }
    // inlining of macro ABC_REALLOC
    
    if p.pArray {
        p.pArray = ((Msat_Clause_t * *) realloc((char *)(p.pArray), sizeof(Msat_Clause_t *) * (size_t)(nCapMin)));
    } else {
        p.pArray = ((Msat_Clause_t * *) malloc(sizeof(Msat_Clause_t *) * (size_t)(nCapMin)));
    } 
    p.nCap   = nCapMin;
}

/**Function*************************************************************

  Synopsis    []

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Msat_ClauseVecShrink(p *Msat_ClauseVec_t, nSizeNew int32) {
    assert(p.nSize >= nSizeNew, "p.nSize >= nSizeNew");
    p.nSize = nSizeNew;
}

/**Function*************************************************************

  Synopsis    []

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Msat_ClauseVecClear(p *Msat_ClauseVec_t) {
    p.nSize = 0;
}

/**Function*************************************************************

  Synopsis    []

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Msat_ClauseVecPush(p *Msat_ClauseVec_t, Entry *Msat_Clause_t) {
    if p.nSize == p.nCap {
        if p.nCap < 16 {
            Msat_ClauseVecGrow( p, 16 );
        } else {
            Msat_ClauseVecGrow( p, 2 * p.nCap ); }
    }
    p.pArray[p.nSize++] = Entry;
}

/**Function*************************************************************

  Synopsis    []

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Msat_ClauseVecPop(p *Msat_ClauseVec_t) *Msat_Clause_t {
    return p.pArray[--p.nSize];
}

/**Function*************************************************************

  Synopsis    []

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Msat_ClauseVecWriteEntry(p *Msat_ClauseVec_t, i int32, Entry *Msat_Clause_t) {
    assert(i >= 0 && i < p.nSize, "i >= 0 && i < p.nSize");
    p.pArray[i] = Entry;
}

/**Function*************************************************************

  Synopsis    []

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Msat_ClauseVecReadEntry(p *Msat_ClauseVec_t, i int32) *Msat_Clause_t {
    assert(i >= 0 && i < p.nSize, "i >= 0 && i < p.nSize");
    return p.pArray[i];
}

////////////////////////////////////////////////////////////////////////
///                       END OF FILE                                ///
////////////////////////////////////////////////////////////////////////


//ABC_NAMESPACE_IMPL_END

