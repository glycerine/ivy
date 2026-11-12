package xsat

import (
	"fmt"
	"os"
)

/**CFile****************************************************************

  FileName    [xsatWatchList.h]

  SystemName  [ABC: Logic synthesis and verification system.]

  PackageName [xSAT - A SAT solver written in C.
               Read the license file for more info.]

  Synopsis    [Watch list and its related structures implementation]

  Author      [Bruno Schmitt <boschmitt@inf.ufrgs.br>]

  Affiliation [UC Berkeley / UFRGS]

  Date        [Ver. 1.0. Started - November 10, 2016.]

  Revision    []

***********************************************************************/



////////////////////////////////////////////////////////////////////////
///                          INCLUDES                                ///
////////////////////////////////////////////////////////////////////////
//#include "misc/util/abc_global.h"

//ABC_NAMESPACE_HEADER_START

////////////////////////////////////////////////////////////////////////
///                    STRUCTURE DEFINITIONS                         ///
////////////////////////////////////////////////////////////////////////
   
type xSAT_Watcher_t struct {
    CRef uint32;
    Blocker int32;
}

   
type xSAT_WatchList_t struct {
    nCap int32;
    nSize int32;
    pArray  *xSAT_Watcher_t;
}

   
type xSAT_VecWatchList_t struct {
    nCap int32;
    nSize int32;
    pArray  *xSAT_WatchList_t;
}

/**Function*************************************************************

  Synopsis    []

  Description []

  SideEffects []

  SeeAlso     []

***********************************************************************/
func xSAT_WatchListFree(v *xSAT_WatchList_t) {
    if v.pArray {
        // inlining of macro ABC_FREE
        ((v.pArray) ? (free((char *) (v.pArray)), (v.pArray) = 0) : 0); }
}

/**Function*************************************************************

  Synopsis    []

  Description []

  SideEffects []

  SeeAlso     []

***********************************************************************/
func xSAT_WatchListSize(v *xSAT_WatchList_t) int32 {
    return v.nSize;
}

/**Function*************************************************************

  Synopsis    []

  Description []

  SideEffects []

  SeeAlso     []

***********************************************************************/
func xSAT_WatchListShrink(v *xSAT_WatchList_t, k int32) {
    assert(k <= v.nSize, "k <= v.nSize");
    v.nSize = k;
}

/**Function*************************************************************

  Synopsis    []

  Description []

  SideEffects []

  SeeAlso     []

***********************************************************************/
func xSAT_WatchListPush(v *xSAT_WatchList_t, e xSAT_Watcher_t) {
    assert(v, "v");
    if v.nSize == v.nCap {
        var newsize int32;
        if v.nCap < 4 {
            newsize = 4;
        } else {
            newsize = ( v.nCap / 2 ) * 3;
        }

        // inlining of macro ABC_REALLOC
        
        if v.pArray {
            v.pArray = ((xSAT_Watcher_t *) realloc((char *)(v.pArray), sizeof(xSAT_Watcher_t) * (size_t)(newsize)));
        } else {
            v.pArray = ((xSAT_Watcher_t *) malloc(sizeof(xSAT_Watcher_t) * (size_t)(newsize)));
        }
        if v.pArray == nil {
            fmt.Printf("Failed to realloc memory from %.1f MB to %.1f MB.\n", 1.0 * v.nCap / (1<<20), 1.0 * newsize / (1<<20));
            fflush( os.Stdout );
        }
        v.nCap = newsize;
    }

    v.pArray[v.nSize++] = e;
}

/**Function*************************************************************

  Synopsis    []

  Description []

  SideEffects []

  SeeAlso     []

***********************************************************************/
func xSAT_WatchListArray(v *xSAT_WatchList_t) *xSAT_Watcher_t {
    return v.pArray;
}

/**Function*************************************************************

  Synopsis    []

  Description []

  SideEffects []

  SeeAlso     []

***********************************************************************/
func xSAT_WatchListRemove(v *xSAT_WatchList_t, CRef uint32) {
    var ws *xSAT_Watcher_t  = xSAT_WatchListArray(v);
    var j int32;

    for  ; ws[j].CRef != CRef; j++  {}
    assert(j < xSAT_WatchListSize( v ), "j < xSAT_WatchListSize( v )");
    memmove( v.pArray + j, v.pArray + j + 1, ( v.nSize - j - 1 ) * sizeof( xSAT_Watcher_t ) );
    v.nSize -= 1;
}

/**Function*************************************************************

  Synopsis    []

  Description []

  SideEffects []

  SeeAlso     []

***********************************************************************/
func xSAT_VecWatchListAlloc(nCap int32) *xSAT_VecWatchList_t {
    // inlining of macro ABC_ALLOC
    var v *xSAT_VecWatchList_t  = ((xSAT_VecWatchList_t *) malloc(sizeof(xSAT_VecWatchList_t) * (size_t)(1)));

    v.nCap   = 4;
    v.nSize  = 0;
    // inlining of macro ABC_CALLOC
    v.pArray = ( xSAT_WatchList_t * ) ((xSAT_WatchList_t *) calloc((size_t)(sizeof( xSAT_WatchList_t ) * v.nCap), sizeof(xSAT_WatchList_t)));
    return v;
}

/**Function*************************************************************

  Synopsis    []

  Description []

  SideEffects []

  SeeAlso     []

***********************************************************************/
func xSAT_VecWatchListFree(v *xSAT_VecWatchList_t) {
    var i int32;
    for  i = 0; i < v.nSize; i++  {
        xSAT_WatchListFree( v.pArray + i );
    }

    // inlining of macro ABC_FREE
    ((v.pArray) ? (free((char *) (v.pArray)), (v.pArray) = 0) : 0);
    // inlining of macro ABC_FREE
    ((v) ? (free((char *) (v)), (v) = 0) : 0);
}

/**Function*************************************************************

  Synopsis    []

  Description []

  SideEffects []

  SeeAlso     []

***********************************************************************/
func xSAT_VecWatchListPush(v *xSAT_VecWatchList_t) {
    if v.nSize == v.nCap {
        var newsize int32;
        if v.nCap < 4 {
            newsize = v.nCap * 2;
        } else {
            newsize = (v.nCap / 2) * 3;
        }

        // inlining of macro ABC_REALLOC
        
        if v.pArray {
            v.pArray = ((xSAT_WatchList_t *) realloc((char *)(v.pArray), sizeof(xSAT_WatchList_t) * (size_t)(newsize)));
        } else {
            v.pArray = ((xSAT_WatchList_t *) malloc(sizeof(xSAT_WatchList_t) * (size_t)(newsize)));
        }
        memset( v.pArray + v.nCap, 0, sizeof( xSAT_WatchList_t ) * ( newsize - v.nCap ) );
        if v.pArray == nil {
            fmt.Printf("Failed to realloc memory from %.1f MB to %.1f MB.\n", 1.0 * v.nCap / (1<<20), 1.0 * newsize / (1<<20));
            fflush( os.Stdout );
        }
        v.nCap = newsize;
    }

    v.nSize++;
}

/**Function*************************************************************

  Synopsis    []

  Description []

  SideEffects []

  SeeAlso     []

***********************************************************************/
func xSAT_VecWatchListEntry(v *xSAT_VecWatchList_t, iEntry int32) *xSAT_WatchList_t {
    assert(iEntry < v.nCap, "iEntry < v.nCap");
    assert(iEntry < v.nSize, "iEntry < v.nSize");
    return v.pArray + iEntry;
}

//ABC_NAMESPACE_HEADER_END


