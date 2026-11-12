package msat

/**CFile****************************************************************

  FileName    [msatVec.c]

  PackageName [A C version of SAT solver MINISAT, originally developed 
  in C++ by Niklas Een and Niklas Sorensson, Chalmers University of 
  Technology, Sweden: http://www.cs.chalmers.se/~een/Satzoo.]

  Synopsis    [Integer vector borrowed from Extra.]

  Author      [Alan Mishchenko <alanmi@eecs.berkeley.edu>]
  
  Affiliation [UC Berkeley]

  Date        [Ver. 1.0. Started - January 1, 2004.]

  Revision    [$Id: msatVec.c,v 1.0 2004/01/01 1:00:00 alanmi Exp $]

***********************************************************************/

//#include "msatInt.h"

//ABC_NAMESPACE_IMPL_START


////////////////////////////////////////////////////////////////////////
///                        DECLARATIONS                              ///
////////////////////////////////////////////////////////////////////////

// static int Msat_IntVecSortCompare1( int * pp1, int * pp2 );

// static int Msat_IntVecSortCompare2( int * pp1, int * pp2 );


////////////////////////////////////////////////////////////////////////
///                     FUNCTION DEFINITIONS                         ///
////////////////////////////////////////////////////////////////////////

/**Function*************************************************************

  Synopsis    [Allocates a vector with the given capacity.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Msat_IntVecAlloc(nCap int32) *Msat_IntVec_t {
    var p *Msat_IntVec_t;
    // inlining of macro ABC_ALLOC
    p = ((Msat_IntVec_t *) malloc(sizeof(Msat_IntVec_t) * (size_t)(1)));
    if nCap > 0 && nCap < 16 {
        nCap = 16; }
    p.nSize  = 0;
    p.nCap   = nCap;
    if p.nCap {
        // inlining of macro ABC_ALLOC
        p.pArray = ((int *) malloc(sizeof(int) * (size_t)(p.nCap)));
    } else {
        p.pArray = nil;
    }
    return p;
}

/**Function*************************************************************

  Synopsis    [Creates the vector from an integer array of the given size.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Msat_IntVecAllocArray(pArray *int32, nSize int32) *Msat_IntVec_t {
    var p *Msat_IntVec_t;
    // inlining of macro ABC_ALLOC
    p = ((Msat_IntVec_t *) malloc(sizeof(Msat_IntVec_t) * (size_t)(1)));
    p.nSize  = nSize;
    p.nCap   = nSize;
    p.pArray = pArray;
    return p;
}

/**Function*************************************************************

  Synopsis    [Creates the vector from an integer array of the given size.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Msat_IntVecAllocArrayCopy(pArray *int32, nSize int32) *Msat_IntVec_t {
    var p *Msat_IntVec_t;
    // inlining of macro ABC_ALLOC
    p = ((Msat_IntVec_t *) malloc(sizeof(Msat_IntVec_t) * (size_t)(1)));
    p.nSize  = nSize;
    p.nCap   = nSize;
    // inlining of macro ABC_ALLOC
    p.pArray = ((int *) malloc(sizeof(int) * (size_t)(nSize)));
    memcpy( p.pArray, pArray, sizeof(int) * nSize );
    return p;
}

/**Function*************************************************************

  Synopsis    [Duplicates the integer array.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Msat_IntVecDup(pVec *Msat_IntVec_t) *Msat_IntVec_t {
    var p *Msat_IntVec_t;
    // inlining of macro ABC_ALLOC
    p = ((Msat_IntVec_t *) malloc(sizeof(Msat_IntVec_t) * (size_t)(1)));
    p.nSize  = pVec.nSize;
    p.nCap   = pVec.nCap;
    if p.nCap {
        // inlining of macro ABC_ALLOC
        p.pArray = ((int *) malloc(sizeof(int) * (size_t)(p.nCap)));
    } else {
        p.pArray = nil;
    }
    memcpy( p.pArray, pVec.pArray, sizeof(int) * pVec.nSize );
    return p;
}

/**Function*************************************************************

  Synopsis    [Transfers the array into another vector.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Msat_IntVecDupArray(pVec *Msat_IntVec_t) *Msat_IntVec_t {
    var p *Msat_IntVec_t;
    // inlining of macro ABC_ALLOC
    p = ((Msat_IntVec_t *) malloc(sizeof(Msat_IntVec_t) * (size_t)(1)));
    p.nSize  = pVec.nSize;
    p.nCap   = pVec.nCap;
    p.pArray = pVec.pArray;
    pVec.nSize  = 0;
    pVec.nCap   = 0;
    pVec.pArray = nil;
    return p;
}

/**Function*************************************************************

  Synopsis    []

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Msat_IntVecFree(p *Msat_IntVec_t) {
    // inlining of macro ABC_FREE
    ((p.pArray) ? (free((char *) (p.pArray)), (p.pArray) = 0) : 0);
    // inlining of macro ABC_FREE
    ((p) ? (free((char *) (p)), (p) = 0) : 0);
}

/**Function*************************************************************

  Synopsis    [Fills the vector with given number of entries.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Msat_IntVecFill(p *Msat_IntVec_t, nSize int32, Entry int32) {
    var i int32;
    Msat_IntVecGrow( p, nSize );
    p.nSize = nSize;
    for  i = 0; i < p.nSize; i++  {
        p.pArray[i] = Entry;
    }
}

/**Function*************************************************************

  Synopsis    []

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Msat_IntVecReleaseArray(p *Msat_IntVec_t) *int32 {
    var pArray *int32  = p.pArray;
    p.nCap = 0;
    p.nSize = 0;
    p.pArray = nil;
    return pArray;
}

/**Function*************************************************************

  Synopsis    []

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Msat_IntVecReadArray(p *Msat_IntVec_t) *int32 {
    return p.pArray;
}

/**Function*************************************************************

  Synopsis    []

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Msat_IntVecReadSize(p *Msat_IntVec_t) int32 {
    return p.nSize;
}

/**Function*************************************************************

  Synopsis    []

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Msat_IntVecReadEntry(p *Msat_IntVec_t, i int32) int32 {
    assert(i >= 0 && i < p.nSize, "i >= 0 && i < p.nSize");
    return p.pArray[i];
}

/**Function*************************************************************

  Synopsis    []

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Msat_IntVecWriteEntry(p *Msat_IntVec_t, i int32, Entry int32) {
    assert(i >= 0 && i < p.nSize, "i >= 0 && i < p.nSize");
    p.pArray[i] = Entry;
}

/**Function*************************************************************

  Synopsis    []

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Msat_IntVecReadEntryLast(p *Msat_IntVec_t) int32 {
    return p.pArray[p.nSize-1];
}

/**Function*************************************************************

  Synopsis    [Resizes the vector to the given capacity.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Msat_IntVecGrow(p *Msat_IntVec_t, nCapMin int32) {
    if p.nCap >= nCapMin {
        return; }
    // inlining of macro ABC_REALLOC
    
    if p.pArray {
        p.pArray = ((int *) realloc((char *)(p.pArray), sizeof(int) * (size_t)(nCapMin)));
    } else {
        p.pArray = ((int *) malloc(sizeof(int) * (size_t)(nCapMin)));
    } 
    p.nCap   = nCapMin;
}

/**Function*************************************************************

  Synopsis    []

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Msat_IntVecShrink(p *Msat_IntVec_t, nSizeNew int32) {
    assert(p.nSize >= nSizeNew, "p.nSize >= nSizeNew");
    p.nSize = nSizeNew;
}

/**Function*************************************************************

  Synopsis    []

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Msat_IntVecClear(p *Msat_IntVec_t) {
    p.nSize = 0;
}

/**Function*************************************************************

  Synopsis    []

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Msat_IntVecPush(p *Msat_IntVec_t, Entry int32) {
    if p.nSize == p.nCap {
        if p.nCap < 16 {
            Msat_IntVecGrow( p, 16 );
        } else {
            Msat_IntVecGrow( p, 2 * p.nCap ); }
    }
    p.pArray[p.nSize++] = Entry;
}

/**Function*************************************************************

  Synopsis    [Add the element while ensuring uniqueness.]

  Description [Returns 1 if the element was found, and 0 if it was new. ]
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Msat_IntVecPushUnique(p *Msat_IntVec_t, Entry int32) int32 {
    var i int32;
    for  i = 0; i < p.nSize; i++  {
        if p.pArray[i] == Entry {
            return 1; }
    }
    Msat_IntVecPush( p, Entry );
    return 0;
}

/**Function*************************************************************

  Synopsis    [Inserts the element while sorting in the increasing order.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Msat_IntVecPushUniqueOrder(p *Msat_IntVec_t, Entry int32, fIncrease int32) {
    var Entry1, Entry2 int32;
    var i int32;
    Msat_IntVecPushUnique( p, Entry );
    // find the p of the node
    for  i = p.nSize-1; i > 0; i--  {
        Entry1 = p.pArray[i  ];
        Entry2 = p.pArray[i-1];
        if ( (fIncrease != 0) && Entry1 >= Entry2) ||
            (!(fIncrease != 0) && Entry1 <= Entry2) {
            break; }
        p.pArray[i  ] = Entry2;
        p.pArray[i-1] = Entry1;
    }
}


/**Function*************************************************************

  Synopsis    [Returns the last entry and removes it from the list.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Msat_IntVecPop(p *Msat_IntVec_t) int32 {
    assert(p.nSize > 0, "p.nSize > 0");
    return p.pArray[--p.nSize];
}

/**Function*************************************************************

  Synopsis    [Sorting the entries by their integer value.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Msat_IntVecSort(p *Msat_IntVec_t, fReverse int32) {
    if fReverse != 0 { 
        qsort( (void *)p.pArray, (size_t)p.nSize, sizeof(int), 
                        (int (*)(void *, void *)) Msat_IntVecSortCompare2 );
    } else {
        qsort( (void *)p.pArray, (size_t)p.nSize, sizeof(int), 
                        (int (*)(void *, void *)) Msat_IntVecSortCompare1 ); }
}

/**Function*************************************************************

  Synopsis    [Comparison procedure for two clauses.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Msat_IntVecSortCompare1(pp1 *int32, pp2 *int32) int32 {
    // for some reason commenting out lines (as shown) led to crashing of the release version
    if *pp1 < *pp2 {
        return -1; }
    if *pp1 > *pp2 { //
        return 1; }
    return 0; //
}

/**Function*************************************************************

  Synopsis    [Comparison procedure for two clauses.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Msat_IntVecSortCompare2(pp1 *int32, pp2 *int32) int32 {
    // for some reason commenting out lines (as shown) led to crashing of the release version
    if *pp1 > *pp2 {
        return -1; }
    if *pp1 < *pp2 { //
        return 1; }
    return 0; //
}

////////////////////////////////////////////////////////////////////////
///                       END OF FILE                                ///
////////////////////////////////////////////////////////////////////////


//ABC_NAMESPACE_IMPL_END

