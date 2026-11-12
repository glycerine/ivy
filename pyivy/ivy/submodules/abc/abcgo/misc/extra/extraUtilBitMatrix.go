package extra

import "fmt"

/**CFile****************************************************************

  FileName    [extraUtilBitMatrix.c]

  PackageName [extra]

  Synopsis    [Various reusable software utilities.]

  Author      [Alan Mishchenko]
  
  Affiliation [UC Berkeley]

  Date        [Ver. 1.0. Started - September 1, 2003.]

  Revision    [$Id: extraUtilBitMatrix.c,v 1.0 2003/09/01 00:00:00 alanmi Exp $]

***********************************************************************/

//#include "extra.h"

//ABC_NAMESPACE_IMPL_START


/*---------------------------------------------------------------------------*/
/* Constant declarations                                                     */
/*---------------------------------------------------------------------------*/

/*---------------------------------------------------------------------------*/
/* Stucture declarations                                                     */
/*---------------------------------------------------------------------------*/

type Extra_BitMat_t_ struct {
    ppData  **uint32;      // bit data
    nSize         int32;       // the number of bits in one dimension
    nWords         int32;      // the number of words in one dimension
    nBitShift         int32;   // the number of bits to shift to get words
    uMask    uint32;       // the mask to get the number of bits in the word
    nLookups         int32;    // the number of lookups  
    nInserts         int32;    // the number of inserts
    nDeletes         int32;    // the number of deletions
}

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

/**AutomaticEnd***************************************************************/


/*---------------------------------------------------------------------------*/
/* Definition of exported functions                                          */
/*---------------------------------------------------------------------------*/

/**Function*************************************************************

  Synopsis    [Starts the bit matrix.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Extra_BitMatrixStart(nSize int32) *Extra_BitMat_t {
    var p *Extra_BitMat_t;
    var i int32;
    // inlining of macro ABC_ALLOC
    p = ((Extra_BitMat_t *) malloc(sizeof(Extra_BitMat_t) * (size_t)(1)));
    memset( p, 0, sizeof(Extra_BitMat_t) );
    p.nSize     = nSize;
    if sizeof(unsigned) == 4 {
        p.nBitShift = 5;
    } else {
        p.nBitShift = 6;
    }
    if sizeof(unsigned) == 4 {
        p.uMask = 31;
    } else {
        p.uMask = 63;
    }
    p.nWords    = nSize / (8 * sizeof(unsigned)) + ((nSize % (8 * sizeof(unsigned))) > 0);
    // inlining of macro ABC_ALLOC
    p.ppData    = ((unsigned * *) malloc(sizeof(unsigned *) * (size_t)(nSize)));
    // inlining of macro ABC_ALLOC
    p.ppData[0] = ((unsigned *) malloc(sizeof(unsigned) * (size_t)(nSize * p.nWords)));
    memset( p.ppData[0], 0, sizeof(unsigned) * nSize * p.nWords );
    for  i = 1; i < nSize; i++  {
        p.ppData[i] = p.ppData[i-1] + p.nWords;
    }
    return p;
}

/**Function*************************************************************

  Synopsis    [Stops the bit matrix.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Extra_BitMatrixClean(p *Extra_BitMat_t) {
    memset( p.ppData[0], 0, sizeof(unsigned) * p.nSize * p.nWords );
}

/**Function*************************************************************

  Synopsis    [Stops the bit matrix.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Extra_BitMatrixStop(p *Extra_BitMat_t) {
    // inlining of macro ABC_FREE
    ((p.ppData[0]) ? (free((char *) (p.ppData[0])), (p.ppData[0]) = 0) : 0);
    // inlining of macro ABC_FREE
    ((p.ppData) ? (free((char *) (p.ppData)), (p.ppData) = 0) : 0);
    // inlining of macro ABC_FREE
    ((p) ? (free((char *) (p)), (p) = 0) : 0);
}

/**Function*************************************************************

  Synopsis    [Prints the bit-matrix.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Extra_BitMatrixPrint(pMat *Extra_BitMat_t) {
    var i, k, nVars int32;
    fmt.Printf("\n");
    nVars = Extra_BitMatrixReadSize( pMat );
    for  i = 0; i < nVars; i++  {
        for  k = 0; k <= i; k++  {
            fmt.Printf(" ");
        }
        for  k = i+1; k < nVars; k++  {
            if Extra_BitMatrixLookup1( pMat, i, k ) != 0 {
                fmt.Printf("1");
            } else {
                fmt.Printf("."); }
        }
        fmt.Printf("\n");
    }
}


/**Function*************************************************************

  Synopsis    [Reads the matrix size.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Extra_BitMatrixReadSize(p *Extra_BitMat_t) int32 {
    return p.nSize;
}

/**Function*************************************************************

  Synopsis    [Inserts the element into the upper part.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Extra_BitMatrixInsert1(p *Extra_BitMat_t, i int32, k int32) {
    p.nInserts++;
    if i < k {
        p.ppData[i][k>>p.nBitShift] |= (1<<(k & p.uMask));
    } else {
        p.ppData[k][i>>p.nBitShift] |= (1<<(i & p.uMask)); }
}

/**Function*************************************************************

  Synopsis    [Inserts the element into the upper part.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Extra_BitMatrixLookup1(p *Extra_BitMat_t, i int32, k int32) int32 {
    p.nLookups++;
    if i < k {
        return ((p.ppData[i][k>>p.nBitShift] & (1<<(k & p.uMask))) > 0);
    } else {
        return ((p.ppData[k][i>>p.nBitShift] & (1<<(i & p.uMask))) > 0); }
}

/**Function*************************************************************

  Synopsis    [Inserts the element into the upper part.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Extra_BitMatrixDelete1(p *Extra_BitMat_t, i int32, k int32) {
    p.nDeletes++;
    if i < k {
        p.ppData[i][k>>p.nBitShift] &= ~(1<<(k & p.uMask));
    } else {
        p.ppData[k][i>>p.nBitShift] &= ~(1<<(i & p.uMask)); }
}



/**Function*************************************************************

  Synopsis    [Inserts the element into the upper part.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Extra_BitMatrixInsert2(p *Extra_BitMat_t, i int32, k int32) {
    p.nInserts++;
    if i > k {
        p.ppData[i][k>>p.nBitShift] |= (1<<(k & p.uMask));
    } else {
        p.ppData[k][i>>p.nBitShift] |= (1<<(i & p.uMask)); }
}

/**Function*************************************************************

  Synopsis    [Inserts the element into the upper part.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Extra_BitMatrixLookup2(p *Extra_BitMat_t, i int32, k int32) int32 {
    p.nLookups++;
    if i > k {
        return ((p.ppData[i][k>>p.nBitShift] & (1<<(k & p.uMask))) > 0);
    } else {
        return ((p.ppData[k][i>>p.nBitShift] & (1<<(i & p.uMask))) > 0); }
}

/**Function*************************************************************

  Synopsis    [Inserts the element into the upper part.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Extra_BitMatrixDelete2(p *Extra_BitMat_t, i int32, k int32) {
    p.nDeletes++;
    if i > k {
        p.ppData[i][k>>p.nBitShift] &= ~(1<<(k & p.uMask));
    } else {
        p.ppData[k][i>>p.nBitShift] &= ~(1<<(i & p.uMask)); }
}


/**Function*************************************************************

  Synopsis    [Inserts the element into the upper part.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Extra_BitMatrixOr(p *Extra_BitMat_t, i int32, pInfo *uint32) {
    var w int32;
    for  w = 0; w < p.nWords; w++  {
        p.ppData[i][w] |= pInfo[w];
    }
}

/**Function*************************************************************

  Synopsis    [Inserts the element into the upper part.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Extra_BitMatrixOrTwo(p *Extra_BitMat_t, i int32, j int32) {
    var w int32;
    for  w = 0; w < p.nWords; w++  {
        p.ppData[i][w] = p.ppData[j][w] = (p.ppData[i][w] | p.ppData[j][w]);
    }
}

/**Function*************************************************************

  Synopsis    [Counts the number of 1's in the upper rectangle.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Extra_BitMatrixCountOnesUpper(p *Extra_BitMat_t) int32 {
    var i, k, nTotal int32;
    for  i = 0; i < p.nSize; i++  {
        for  k = i + 1; k < p.nSize; k++  {
            nTotal += ( (p.ppData[i][k>>5] & (1 << (k&31))) > 0 );
        }
    }
    return nTotal;
}

/**Function*************************************************************

  Synopsis    [Returns 1 if the matrices have no entries in common.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Extra_BitMatrixIsDisjoint(p1 *Extra_BitMat_t, p2 *Extra_BitMat_t) int32 {
    var i, w int32;
    assert(p1.nSize == p2.nSize, "p1.nSize == p2.nSize");
    for  i = 0; i < p1.nSize; i++  {
        for  w = 0; w < p1.nWords; w++  {
            if p1.ppData[i][w] & p2.ppData[i][w] {
                return 0; }
        }
    }
    return 1;
}

/**Function*************************************************************

  Synopsis    [Returns 1 if the matrix is a set of cliques.]

  Description [For example pairwise symmetry info should satisfy this property.]
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Extra_BitMatrixIsClique(pMat *Extra_BitMat_t) int32 {
    var v, u, i int32;
    for  v = 0; v < pMat.nSize; v++  {
    for  u = v+1; u < pMat.nSize; u++  {
        if !(Extra_BitMatrixLookup1( pMat, v, u ) != 0) {
            continue; }
        // v and u are symmetric
        for  i = 0; i < pMat.nSize; i++  {
            if i == v || i == u {
                continue; }
            // i is neither v nor u
            // the symmetry status of i is the same w.r.t. to v and u
            if Extra_BitMatrixLookup1( pMat, i, v ) != Extra_BitMatrixLookup1( pMat, i, u ) {
                return 0; }
        }
    }
    }
    return 1;
}


////////////////////////////////////////////////////////////////////////
///                       END OF FILE                                ///
////////////////////////////////////////////////////////////////////////


//ABC_NAMESPACE_IMPL_END

