package bsat

import "fmt"

/**CFile****************************************************************

  FileName    [satMem.h]

  SystemName  [ABC: Logic synthesis and verification system.]

  PackageName [SAT solver.]

  Synopsis    [Memory management.]

  Author      [Alan Mishchenko <alanmi@eecs.berkeley.edu>]
  
  Affiliation [UC Berkeley]

  Date        [Ver. 1.0. Started - January 1, 2004.]

  Revision    [$Id: satMem.h,v 1.0 2004/01/01 1:00:00 alanmi Exp $]

***********************************************************************/
 



////////////////////////////////////////////////////////////////////////
///                          INCLUDES                                ///
////////////////////////////////////////////////////////////////////////

//#include "misc/util/abc_global.h"

//ABC_NAMESPACE_HEADER_START

////////////////////////////////////////////////////////////////////////
///                         PARAMETERS                               ///
////////////////////////////////////////////////////////////////////////

//#define LEARNT_MAX_START_DEFAULT  0
const LEARNT_MAX_START_DEFAULT =  10000
const LEARNT_MAX_INCRE_DEFAULT =   1000
const LEARNT_MAX_RATIO_DEFAULT =     50

////////////////////////////////////////////////////////////////////////
///                    STRUCTURE DEFINITIONS                         ///
////////////////////////////////////////////////////////////////////////
 
//=================================================================================================
// Clause datatype + minor functions:

   
type clause struct {
    lrn uint32; /* C bitfield: 1 bits; review packing and truncation */
    mark uint32; /* C bitfield: 1 bits; review packing and truncation */
    partA uint32; /* C bitfield: 1 bits; review packing and truncation */
    lbd uint32; /* C bitfield: 8 bits; review packing and truncation */
    size uint32; /* C bitfield: 21 bits; review packing and truncation */
    lit        lits[0];
}

// learned clauses have "hidden" literal (c->lits[c->size]) to store clause ID

// data-structure for logging entries
// memory is allocated in 2^nPageSize word-sized pages
// the first 'word' of each page are stores the word limit

// although clause memory pieces are aligned to 64-bit words
// the integer clause handles are in terms of 32-bit unsigneds
// allowing for the first bit to be used for labeling 2-lit clauses


   
type Sat_Mem_t struct {
    nEntries [2]int32;  // entry count
    BookMarkH [2]int32; // bookmarks for handles
    BookMarkE [2]int32; // bookmarks for entries
    iPage [2]int32;     // current memory page
    nPageSize                 int32;    // page log size in terms of ints
    uPageMask            uint32;    // page mask
    uLearnedMask            uint32; // learned mask
    nPagesAlloc                 int32;  // page count allocated
    pPages               **int32;       // page pointers
} 

func Sat_MemLimit(p *int32) int32 { return p[0];                                 }
func Sat_MemIncLimit(p *int32, nInts int32) int32 { return p[0] += nInts;                        }
func Sat_MemWriteLimit(p *int32, nInts int32) { p[0] = nInts;                                }

func Sat_MemHandPage(p *Sat_Mem_t, h cla) int32 { return h >> p.nPageSize;                    }
func Sat_MemHandShift(p *Sat_Mem_t, h cla) int32 { return h & p.uPageMask;                     }

//static inline int       Sat_MemIntSize( int size, int lrn )          { return (size + 2 + lrn) & ~01;               }
func Sat_MemIntSize(size int32, lrn int32) int32 { return 2*((size + 2 + lrn)/2);               }
func Sat_MemClauseSize(p *clause) int32 { return Sat_MemIntSize(p.size, p.lrn);      }
func Sat_MemClauseSize2(p *clause) int32 { return Sat_MemIntSize(p.size, 1);           }

//static inline clause *  Sat_MemClause( Sat_Mem_t * p, int i, int k ) { assert(i <= p->iPage[i&1] && k <= Sat_MemLimit(p->pPages[i]));  return (clause *)(p->pPages[i] + k ); }
func Sat_MemClause(p *Sat_Mem_t, i int32, k int32) *clause { assert((k != 0), "k"); return (clause *)(p.pPages[i] + k);         }
//static inline clause *  Sat_MemClauseHand( Sat_Mem_t * p, cla h )    { assert(Sat_MemHandPage(p, h) <= p->iPage[(h & p->uLearnedMask) > 0]); assert(Sat_MemHandShift(p, h) >= 2 && Sat_MemHandShift(p, h) < (int)p->uLearnedMask); return Sat_MemClause( p, Sat_MemHandPage(p, h), Sat_MemHandShift(p, h) );     }
func Sat_MemClauseHand(p *Sat_Mem_t, h cla) *clause { if h != 0 {
    return Sat_MemClause( p, Sat_MemHandPage(p, h), Sat_MemHandShift(p, h) );
} else {
    return nil;
}                  }
func Sat_MemEntryNum(p *Sat_Mem_t, lrn int32) int32 { return p.nEntries[lrn];                                                                              }

func Sat_MemHand(p *Sat_Mem_t, i int32, k int32) cla { return (i << p.nPageSize) | k;                                                                       }
func Sat_MemHandCurrent(p *Sat_Mem_t, lrn int32) cla { return (p.iPage[lrn] << p.nPageSize) | Sat_MemLimit( p.pPages[p.iPage[lrn]] );                    }

func Sat_MemClauseUsed(p *Sat_Mem_t, h cla) int32 { return h < p.BookMarkH[(h & p.uLearnedMask) > 0];                                                   }

func Sat_MemMemoryHand(p *Sat_Mem_t, h cla) float64 { return 1.0 * ((Sat_MemHandPage(p, h) + 2)/2 * (1 << (p.nPageSize+2)) + Sat_MemHandShift(p, h) * 4);  }
func Sat_MemMemoryUsed(p *Sat_Mem_t, lrn int32) float64 { return Sat_MemMemoryHand( p, Sat_MemHandCurrent(p, lrn) );                                            }
func Sat_MemMemoryAllUsed(p *Sat_Mem_t) float64 { return Sat_MemMemoryUsed( p, 0 ) + Sat_MemMemoryUsed( p, 1 );                                         }
func Sat_MemMemoryAll(p *Sat_Mem_t) float64 { return 1.0 * (p.iPage[0] + p.iPage[1] + 2) * (1 << (p.nPageSize+2));                               }

// p is memory storage
// c is clause pointer
// i is page number
// k is page offset

// print problem clauses NOT in proof mode
// Original macro:
// #define Sat_MemForEachClause( p, c, i, k )      \
//     for ( i = 0; i <= p->iPage[0]; i += 2 )     \
//         for ( k = 2; k < Sat_MemLimit(p->pPages[i]) && ((c) = Sat_MemClause( p, i, k )); k += Sat_MemClauseSize(c) ) if ( i == 0 && k == 2 ) {} else


// print problem clauses in proof mode
// Original macro:
// #define Sat_MemForEachClause2( p, c, i, k )      \
//     for ( i = 0; i <= p->iPage[0]; i += 2 )     \
//         for ( k = 2; k < Sat_MemLimit(p->pPages[i]) && ((c) = Sat_MemClause( p, i, k )); k += Sat_MemClauseSize2(c) ) if ( i == 0 && k == 2 ) {} else


// Original macro:
// #define Sat_MemForEachLearned( p, c, i, k )     \
//     for ( i = 1; i <= p->iPage[1]; i += 2 )     \
//         for ( k = 2; k < Sat_MemLimit(p->pPages[i]) && ((c) = Sat_MemClause( p, i, k )); k += Sat_MemClauseSize(c) )


////////////////////////////////////////////////////////////////////////
///                       GLOBAL VARIABLES                           ///
////////////////////////////////////////////////////////////////////////

////////////////////////////////////////////////////////////////////////
///                       MACRO DEFINITIONS                          ///
////////////////////////////////////////////////////////////////////////

func clause_from_lit(l lit) int32 { return l + l + 1;                            }
func clause_is_lit(h cla) int32 { return (h & 1);                              }
func clause_read_lit(h cla) lit { return (lit)(h >> 1);                        }

func clause_learnt_h(p *Sat_Mem_t, h cla) int32 { return (h & p.uLearnedMask) > 0;            }
func clause_learnt(c *clause) int32 { return c.lrn;                               }
func clause_id(c *clause) int32 { return c.lits[c.size];                     }
func clause_set_id(c *clause, id int32) { c.lits[c.size] = id;                       }
func clause_size(c *clause) int32 { return c.size;                              }
func clause_begin(c *clause) *lit { return c.lits;                              }
func clause_end(c *clause) *lit { return c.lits + c.size;                    }
func clause_print_(c *clause) { 
    var i int32;
    fmt.Printf("{ ");
    for  i = 0; i < clause_size(c); i++  {
        fmt.Printf("%d ", (clause_begin(c)[i] & 1)? -(clause_begin(c)[i] >> 1) : clause_begin(c)[i] >> 1);
    }
    fmt.Printf("}\n");
}

////////////////////////////////////////////////////////////////////////
///                     FUNCTION DECLARATIONS                        ///
////////////////////////////////////////////////////////////////////////

/**Function*************************************************************

  Synopsis    [Allocating vector.]

  Description []

  SideEffects []

  SeeAlso     []

***********************************************************************/
func Sat_MemCountL(p *Sat_Mem_t) int32 {
    var c *clause;
    var i, k, Count int32;
    // inlining of macro Sat_MemForEachLearned
    for  i = 1; i <= p.iPage[1]; i += 2  {             for  k = 2; k < Sat_MemLimit(p.pPages[i]) && ((c) = Sat_MemClause( p, i, k )); k += Sat_MemClauseSize(c)  {
        Count++;
    }
    }
    return Count;
}

/**Function*************************************************************

  Synopsis    [Allocating vector.]

  Description []

  SideEffects []

  SeeAlso     []

***********************************************************************/
func Sat_MemAlloc_(p *Sat_Mem_t, nPageSize int32) {
    assert(nPageSize > 8 && nPageSize < 32, "nPageSize > 8 && nPageSize < 32");
    memset( p, 0, sizeof(Sat_Mem_t) );
    p.nPageSize    = nPageSize;
    p.uLearnedMask = uint32((1 << nPageSize));
    p.uPageMask    = uint32(((1 << nPageSize) - 1));
    p.nPagesAlloc  = 256;
    // inlining of macro ABC_CALLOC
    p.pPages       = ((int * *) calloc((size_t)(p.nPagesAlloc), sizeof(int *)));
    // inlining of macro ABC_ALLOC
    p.pPages[0]    = ((int *) malloc(sizeof(int) * (size_t)(int32((((word)1) << p.nPageSize)))));
    // inlining of macro ABC_ALLOC
    p.pPages[1]    = ((int *) malloc(sizeof(int) * (size_t)(int32((((word)1) << p.nPageSize)))));
    p.iPage[0]     = 0;
    p.iPage[1]     = 1;
    Sat_MemWriteLimit( p.pPages[0], 2 );
    Sat_MemWriteLimit( p.pPages[1], 2 );
}
func Sat_MemAlloc(nPageSize int32) *Sat_Mem_t {
    
    // inlining of macro ABC_CALLOC
    p := &Sat_Mem_t{};
    Sat_MemAlloc_( p, nPageSize );
    return p;
}

/**Function*************************************************************

  Synopsis    [Resetting vector.]

  Description []

  SideEffects []

  SeeAlso     []

***********************************************************************/
func Sat_MemRestart(p *Sat_Mem_t) {
    p.nEntries[0]  = 0;
    p.nEntries[1]  = 0;
    p.iPage[0]     = 0;
    p.iPage[1]     = 1;
    Sat_MemWriteLimit( p.pPages[0], 2 );
    Sat_MemWriteLimit( p.pPages[1], 2 );
}

/**Function*************************************************************

  Synopsis    [Sets the bookmark.]

  Description []

  SideEffects []

  SeeAlso     []

***********************************************************************/
func Sat_MemBookMark(p *Sat_Mem_t) {
    p.BookMarkE[0] = p.nEntries[0];
    p.BookMarkE[1] = p.nEntries[1];
    p.BookMarkH[0] = Sat_MemHandCurrent( p, 0 );
    p.BookMarkH[1] = Sat_MemHandCurrent( p, 1 );
}
func Sat_MemRollBack(p *Sat_Mem_t) {
    p.nEntries[0]  = p.BookMarkE[0];
    p.nEntries[1]  = p.BookMarkE[1];
    p.iPage[0]     = Sat_MemHandPage( p, p.BookMarkH[0] );
    p.iPage[1]     = Sat_MemHandPage( p, p.BookMarkH[1] );
    Sat_MemWriteLimit( p.pPages[p.iPage[0]], Sat_MemHandShift( p, p.BookMarkH[0] ) );
    Sat_MemWriteLimit( p.pPages[p.iPage[1]], Sat_MemHandShift( p, p.BookMarkH[1] ) );
}

/**Function*************************************************************

  Synopsis    [Freeing vector.]

  Description []

  SideEffects []

  SeeAlso     []

***********************************************************************/
func Sat_MemFree_(p *Sat_Mem_t) {
    var i int32;
    for  i = 0; i < p.nPagesAlloc; i++  {
        // inlining of macro ABC_FREE
        ((p.pPages[i]) ? (free((char *) (p.pPages[i])), (p.pPages[i]) = 0) : 0);
    }
    // inlining of macro ABC_FREE
    ((p.pPages) ? (free((char *) (p.pPages)), (p.pPages) = 0) : 0);
}
func Sat_MemFree(p *Sat_Mem_t) {
    Sat_MemFree_( p );
    // inlining of macro ABC_FREE
    ((p) ? (free((char *) (p)), (p) = 0) : 0);
}

/**Function*************************************************************

  Synopsis    [Creates new clause.]

  Description [The resulting clause is fully initialized.]

  SideEffects []

  SeeAlso     []

***********************************************************************/
func Sat_MemAppend(p *Sat_Mem_t, pArray *int32, nSize int32, lrn int32, fPlus1 int32) int32 {
    var c *clause;
    var pPage *int32  = p.pPages[p.iPage[lrn]];
    var nInts int32  = Sat_MemIntSize( nSize, lrn | fPlus1 );
    assert(nInts + 3 < (1 << p.nPageSize), "nInts + 3 < (1 << p.nPageSize)");
    // need two extra at the begining of the page and one extra in the end
    if Sat_MemLimit(pPage) + nInts + 2 >= (1 << p.nPageSize) { 
        p.iPage[lrn] += 2;
        if p.iPage[lrn] >= p.nPagesAlloc {
            // inlining of macro ABC_REALLOC
            
            if p.pPages {
                p.pPages = ((int * *) realloc((char *)(p.pPages), sizeof(int *) * (size_t)(p.nPagesAlloc * 2)));
            } else {
                p.pPages = ((int * *) malloc(sizeof(int *) * (size_t)(p.nPagesAlloc * 2)));
            }
            memset( p.pPages + p.nPagesAlloc, 0, sizeof(int *) * p.nPagesAlloc );
            p.nPagesAlloc *= 2;
        }
        if p.pPages[p.iPage[lrn]] == nil {
            // inlining of macro ABC_ALLOC
            p.pPages[p.iPage[lrn]] = ((int *) malloc(sizeof(int) * (size_t)(int32((((word)1) << p.nPageSize))))); }
        pPage = p.pPages[p.iPage[lrn]];
        Sat_MemWriteLimit( pPage, 2 );
    }
    pPage[Sat_MemLimit(pPage)] = 0;
    c = (clause *)(pPage + Sat_MemLimit(pPage));
    c.size = nSize;
    c.lrn = lrn;
    if pArray != nil {
        memcpy( c.lits, pArray, sizeof(int) * nSize ); }
    if lrn | fPlus1 {
        c.lits[c.size] = p.nEntries[lrn]; }
    p.nEntries[lrn]++;
    Sat_MemIncLimit( pPage, nInts );
    return Sat_MemHandCurrent(p, lrn) - nInts;
}

/**Function*************************************************************

  Synopsis    [Shrinking vector size.]

  Description []

  SideEffects [This procedure does not update the number of entries.]

  SeeAlso     []

***********************************************************************/
func Sat_MemShrink(p *Sat_Mem_t, h int32, lrn int32) {
    assert(clause_learnt_h(p, h) == lrn, "clause_learnt_h(p, h) == lrn");
    assert((h != 0) && h <= Sat_MemHandCurrent(p, lrn), "h && h <= Sat_MemHandCurrent(p, lrn)");
    p.iPage[lrn] = Sat_MemHandPage(p, h);
    Sat_MemWriteLimit( p.pPages[p.iPage[lrn]], Sat_MemHandShift(p, h) );
}


/**Function*************************************************************

  Synopsis    [Compacts learned clauses by removing marked entries.]

  Description [Returns the number of remaining entries.]

  SideEffects []

  SeeAlso     []

***********************************************************************/
func Sat_MemCompactLearned(p *Sat_Mem_t, fDoMove int32) int32 {
    var c *clause;
    var cPivot *clause = nil;
    var i int32;
    var k int32;
    var iNew int32 = 1;
    var kNew int32 = 2;
    var nInts int32;
    var fStartLooking int32;
    var Counter int32;
    var hLimit int32  = Sat_MemHandCurrent(p, 1);
    if hLimit == Sat_MemHand(p, 1, 2) {
        return 0; }
    if (fDoMove != 0) && p.BookMarkH[1] {
        // move the pivot
        assert(p.BookMarkH[1] >= Sat_MemHand(p, 1, 2) && p.BookMarkH[1] <= hLimit, "p.BookMarkH[1] >= Sat_MemHand(p, 1, 2) && p.BookMarkH[1] <= hLimit");
        // get the pivot and remember it may be pointed offlimit
        cPivot = Sat_MemClauseHand( p, p.BookMarkH[1] );
        if p.BookMarkH[1] < hLimit && !cPivot.mark {
            p.BookMarkH[1] = cPivot.lits[cPivot.size];
            cPivot = nil;
        }
        // else find the next used clause after cPivot
    }
    // iterate through the learned clauses
    fStartLooking = 0;
    // inlining of macro Sat_MemForEachLearned
    for  i = 1; i <= p.iPage[1]; i += 2  {             for  k = 2; k < Sat_MemLimit(p.pPages[i]) && ((c) = Sat_MemClause( p, i, k )); k += Sat_MemClauseSize(c)  {
        assert(c.lrn, "c.lrn");
        // skip marked entry
        if c.mark {
            // if pivot is a marked clause, start looking for the next non-marked one
            if (cPivot != nil) && cPivot == c {
                fStartLooking = 1;
                cPivot = nil;
            }
            continue;
        }
        // if we started looking before, we found it!
        if fStartLooking != 0 {
            fStartLooking = 0;
            p.BookMarkH[1] = c.lits[c.size];
        }
        // compute entry size
        nInts = Sat_MemClauseSize(c);
        assert(!(nInts & 1), "!(nInts & 1)");
        // check if we need to scroll to the next page
        if kNew + nInts >= (1 << p.nPageSize) {
            // set the limit of the current page
            if fDoMove != 0 {
                Sat_MemWriteLimit( p.pPages[iNew], kNew ); }
            // move writing position to the new page
            iNew += 2;
            kNew = 2;
        }
        if fDoMove != 0 {
            // make sure the result is the same as previous dry run
            assert(c.lits[c.size] == Sat_MemHand(p, iNew, kNew), "c.lits[c.size] == Sat_MemHand(p, iNew, kNew)");
            // only copy the clause if it has changed
            if i != iNew || k != kNew {
                memmove( p.pPages[iNew] + kNew, c, sizeof(int) * nInts );
//                c = Sat_MemClause( p, iNew, kNew ); // assersions do not hold during dry run
                c = (clause *)(p.pPages[iNew] + kNew);
                assert(nInts == Sat_MemClauseSize(c), "nInts == Sat_MemClauseSize(c)");
            }
            // set the new ID value
            c.lits[c.size] = Counter;
        
        } else { // remember the address of the clause in the new location
            c.lits[c.size] = Sat_MemHand(p, iNew, kNew); }
        // update writing position
        kNew += nInts;
        assert(iNew <= i && kNew < (1 << p.nPageSize), "iNew <= i && kNew < (1 << p.nPageSize)");
        // update counter
        Counter++;
    }
    }
    if fDoMove != 0 {
        // update the counter
        p.nEntries[1] = Counter;
        // update the page count
        p.iPage[1] = iNew;
        // set the limit of the last page
        Sat_MemWriteLimit( p.pPages[iNew], kNew );
        // check if the pivot need to be updated
        if p.BookMarkH[1] {
            if cPivot != nil {
                p.BookMarkH[1] = Sat_MemHandCurrent(p, 1);
                p.BookMarkE[1] = p.nEntries[1];
            
            } else {
                p.BookMarkE[1] = clause_id(Sat_MemClauseHand( p, p.BookMarkH[1] )); }
        }

    }
    return Counter;
}


//ABC_NAMESPACE_HEADER_END



////////////////////////////////////////////////////////////////////////
///                       END OF FILE                                ///
////////////////////////////////////////////////////////////////////////

