package vec

/**CFile****************************************************************

  FileName    [vecSet.h]

  SystemName  [ABC: Logic synthesis and verification system.]

  PackageName [SAT solvers.]

  Synopsis    [Multi-page dynamic array.]

  Author      [Alan Mishchenko]

  Affiliation [UC Berkeley]

  Date        [Ver. 1.0. Started - June 20, 2005.]

  Revision    [$Id: vecSet.h,v 1.00 2005/06/20 00:00:00 alanmi Exp $]

***********************************************************************/





////////////////////////////////////////////////////////////////////////
///                          INCLUDES                                ///
////////////////////////////////////////////////////////////////////////

//#include <stdio.h>

//ABC_NAMESPACE_HEADER_START


////////////////////////////////////////////////////////////////////////
///                         PARAMETERS                               ///
////////////////////////////////////////////////////////////////////////

////////////////////////////////////////////////////////////////////////
///                         BASIC TYPES                              ///
////////////////////////////////////////////////////////////////////////

// data-structure for logging entries
// memory is allocated in 2^nPageSize word-sized pages
// the first two 'words' of each page are used for bookkeeping
// the first 'word' of bookkeeping data stores the word limit
// the second 'word' of bookkeeping data stores the shadow word limit
// (the shadow word limit is only used during garbage collection)

   
type Vec_Set_t struct {
    nPageSize               int32;    // page size
    uPageMask          uint32;    // page mask
    nEntries               int32;     // entry count
    iPage               int32;        // current page
    iPageS               int32;       // shadow page
    nPagesAlloc               int32;  // page count allocated
    pPages            **word;       // page pointers
} 

////////////////////////////////////////////////////////////////////////
///                      MACRO DEFINITIONS                           ///
////////////////////////////////////////////////////////////////////////

func Vec_SetHandPage(p *Vec_Set_t, h int32) int32 { return h >> p.nPageSize;      }
func Vec_SetHandShift(p *Vec_Set_t, h int32) int32 { return h & p.uPageMask;       }
func Vec_SetWordNum(nSize int32) int32 { return (nSize + 1) >> 1;       }

//static inline word *  Vec_SetEntry( Vec_Set_t * p, int h )       { assert(Vec_SetHandPage(p, h) >= 0 && Vec_SetHandPage(p, h) <= p->iPage); assert(Vec_SetHandShift(p, h) >= 2 && Vec_SetHandShift(p, h) < (1 << p->nPageSize)); return p->pPages[Vec_SetHandPage(p, h)] + Vec_SetHandShift(p, h);     }
func Vec_SetEntry(p *Vec_Set_t, h int32) *word { return p.pPages[Vec_SetHandPage(p, h)] + Vec_SetHandShift(p, h);                     }
func Vec_SetEntryNum(p *Vec_Set_t) int32 { return p.nEntries;            }
func Vec_SetWriteEntryNum(p *Vec_Set_t, i int32) { p.nEntries = i;               }

func Vec_SetLimit(p *word) int32 { return p[0];                   }
func Vec_SetLimitS(p *word) int32 { return p[1];                   }

func Vec_SetIncLimit(p *word, nWords int32) int32 { return p[0] += nWords;         }
func Vec_SetIncLimitS(p *word, nWords int32) int32 { return p[1] += nWords;         }

func Vec_SetWriteLimit(p *word, nWords int32) { p[0] = nWords;                 }
func Vec_SetWriteLimitS(p *word, nWords int32) { p[1] = nWords;                 }

func Vec_SetHandCurrent(p *Vec_Set_t) int32 { return (p.iPage << p.nPageSize)  + Vec_SetLimit(p.pPages[p.iPage]);               }
func Vec_SetHandCurrentS(p *Vec_Set_t) int32 { return (p.iPageS << p.nPageSize) + Vec_SetLimitS(p.pPages[p.iPageS]);             }

func Vec_SetHandMemory(p *Vec_Set_t, h int32) int32 { return Vec_SetHandPage(p, h) * (1 << (p.nPageSize+3)) + Vec_SetHandShift(p, h) * 8;  }
func Vec_SetMemory(p *Vec_Set_t) int32 { return Vec_SetHandMemory(p, Vec_SetHandCurrent(p));                                   }
func Vec_SetMemoryS(p *Vec_Set_t) int32 { return Vec_SetHandMemory(p, Vec_SetHandCurrentS(p));                                  }
func Vec_SetMemoryAll(p *Vec_Set_t) int32 { return (p.iPage+1) * (1 << (p.nPageSize+3));                                        }

// Type is the Set type
// pVec is vector of set
// nSize should be given by the user
// pSet is the pointer to the set
// p (page) and s (shift) are variables used here
// Original macro:
// #define Vec_SetForEachEntry( Type, pVec, nSize, pSet, p, s )   \
//     for ( p = 0; p <= pVec->iPage; p++ )                       \
//         for ( s = 2; s < Vec_SetLimit(pVec->pPages[p]) && ((pSet) = (Type)(pVec->pPages[p] + (s))); s += nSize )


////////////////////////////////////////////////////////////////////////
///                     FUNCTION DEFINITIONS                         ///
////////////////////////////////////////////////////////////////////////

/**Function*************************************************************

  Synopsis    [Allocating vector.]

  Description []

  SideEffects []

  SeeAlso     []

***********************************************************************/
func Vec_SetAlloc_(p *Vec_Set_t, nPageSize int32) {
    assert(nPageSize > 8, "nPageSize > 8");
    memset( p, 0, sizeof(Vec_Set_t) );
    p.nPageSize    = nPageSize;
    p.uPageMask    = uint32(((1 << nPageSize) - 1));
    p.nPagesAlloc  = 256;
    p.pPages       = ABC_CALLOC( word *, p.nPagesAlloc );
    p.pPages[0]    = ABC_ALLOC( word, int32((((word)1) << p.nPageSize)) );
    p.pPages[0][0] = ~0;
    p.pPages[0][1] = ~0;
    Vec_SetWriteLimit( p.pPages[0], 2 );
}
func Vec_SetAlloc(nPageSize int32) *Vec_Set_t {
    var p *Vec_Set_t;
    p = new(Vec_Set_t);
    Vec_SetAlloc_( p, nPageSize );
    return p;
}

/**Function*************************************************************

  Synopsis    [Resetting vector.]

  Description []

  SideEffects []

  SeeAlso     []

***********************************************************************/
func Vec_SetRestart(p *Vec_Set_t) {
    p.nEntries     = 0;
    p.iPage        = 0;
    p.iPageS       = 0;
    p.pPages[0][0] = ~0;
    p.pPages[0][1] = ~0;
    Vec_SetWriteLimit( p.pPages[0], 2 );
}

/**Function*************************************************************

  Synopsis    [Freeing vector.]

  Description []

  SideEffects []

  SeeAlso     []

***********************************************************************/
func Vec_SetFree_(p *Vec_Set_t) {
    var i int32;
    if p == nil { return; }
    for  i = 0; i < p.nPagesAlloc; i++  {
        ABC_FREE( p.pPages[i] );
    }
    ABC_FREE( p.pPages );
}
func Vec_SetFree(p *Vec_Set_t) {
    if p == nil { return; }
    Vec_SetFree_( p );
    ABC_FREE( p );
}

/**Function*************************************************************

  Synopsis    [Returns memory in bytes occupied by the vector.]

  Description []

  SideEffects []

  SeeAlso     []

***********************************************************************/
func Vec_ReportMemory(p *Vec_Set_t) float64 {
    var Mem float64  = sizeof(Vec_Set_t);
    Mem += p.nPagesAlloc * sizeof(void *);
    Mem += sizeof(word) * (size_t)(((word)1) << p.nPageSize) * (size_t)(1 + p.iPage);
    return Mem;
}

/**Function*************************************************************

  Synopsis    [Appending entries to vector.]

  Description []

  SideEffects []

  SeeAlso     []

***********************************************************************/
func Vec_SetAppend(p *Vec_Set_t, pArray *int32, nSize int32) int32 {
    var nWords int32  = Vec_SetWordNum( nSize );
    assert(nWords < (1 << p.nPageSize), "nWords < (1 << p.nPageSize)");
    p.nEntries++;
    if Vec_SetLimit( p.pPages[p.iPage] ) + nWords >= (1 << p.nPageSize) {
        if ++p.iPage == p.nPagesAlloc {
            p.pPages = ABC_REALLOC( word *, p.pPages, p.nPagesAlloc * 2 );
            memset( p.pPages + p.nPagesAlloc, 0, sizeof(word *) * (size_t)p.nPagesAlloc );
            p.nPagesAlloc *= 2;
        }
        if p.pPages[p.iPage] == nil {
            p.pPages[p.iPage] = ABC_ALLOC( word, int32((((word)1) << p.nPageSize)) ); }
        Vec_SetWriteLimit( p.pPages[p.iPage], 2 );
        p.pPages[p.iPage][1] = ~0;
    }
    if pArray != nil {
        memcpy( p.pPages[p.iPage] + Vec_SetLimit(p.pPages[p.iPage]), pArray, sizeof(int) * (size_t)nSize ); }
    Vec_SetIncLimit( p.pPages[p.iPage], nWords );
    return Vec_SetHandCurrent(p) - nWords;
}
func Vec_SetAppendS(p *Vec_Set_t, nSize int32) int32 {
    var nWords int32  = Vec_SetWordNum( nSize );
    assert(nWords < (1 << p.nPageSize), "nWords < (1 << p.nPageSize)");
    if Vec_SetLimitS( p.pPages[p.iPageS] ) + nWords >= (1 << p.nPageSize) {
        Vec_SetWriteLimitS( p.pPages[++p.iPageS], 2 ); }
    Vec_SetIncLimitS( p.pPages[p.iPageS], nWords );
    return Vec_SetHandCurrentS(p) - nWords;
}
func Vec_SetFetchH(p *Vec_Set_t, nBytes int32) int32 {
    return Vec_SetAppend(p, nil, (nBytes + 3) >> 2);
}
func Vec_SetFetch(p *Vec_Set_t, nBytes int32) any {
    return (void *)Vec_SetEntry( p, Vec_SetFetchH(p, nBytes) );
}
func Vec_SetStrsav(p *Vec_Set_t, pName string) string {
    var pStr string  = (char *)Vec_SetFetch( p, int32(len(pName)) + 1 );
    pStr = pName;
    return pStr;
}

/**Function*************************************************************

  Synopsis    [Shrinking vector size.]

  Description []

  SideEffects [This procedure does not update the number of entries.]

  SeeAlso     []

***********************************************************************/
func Vec_SetShrink(p *Vec_Set_t, h int32) {
    assert(h <= Vec_SetHandCurrent(p), "h <= Vec_SetHandCurrent(p)");
    p.iPage = Vec_SetHandPage(p, h);
    Vec_SetWriteLimit( p.pPages[p.iPage], Vec_SetHandShift(p, h) );
}
func Vec_SetShrinkS(p *Vec_Set_t, h int32) {
    assert(h <= Vec_SetHandCurrent(p), "h <= Vec_SetHandCurrent(p)");
    p.iPageS = Vec_SetHandPage(p, h);
    Vec_SetWriteLimitS( p.pPages[p.iPageS], Vec_SetHandShift(p, h) );
}

func Vec_SetShrinkLimits(p *Vec_Set_t) {
    var i int32;
    for  i = 0; i <= p.iPage; i++  {
        Vec_SetWriteLimit( p.pPages[i], Vec_SetLimitS(p.pPages[i]) );
    }
}


//ABC_NAMESPACE_HEADER_END



////////////////////////////////////////////////////////////////////////
///                       END OF FILE                                ///
////////////////////////////////////////////////////////////////////////

