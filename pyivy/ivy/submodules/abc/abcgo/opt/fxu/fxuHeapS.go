package fxu

import "fmt"

import "os"

/**CFile****************************************************************

  FileName    [fxuHeapS.c]

  PackageName [MVSIS 2.0: Multi-valued logic synthesis system.]

  Synopsis    [The priority queue for variables.]

  Author      [MVSIS Group]
  
  Affiliation [UC Berkeley]

  Date        [Ver. 1.0. Started - February 1, 2003.]

  Revision    [$Id: fxuHeapS.c,v 1.0 2003/02/01 00:00:00 alanmi Exp $]

***********************************************************************/

//#include "fxuInt.h"

//ABC_NAMESPACE_IMPL_START


////////////////////////////////////////////////////////////////////////
///                        DECLARATIONS                              ///
////////////////////////////////////////////////////////////////////////

// Original macro:
// #define FXU_HEAP_SINGLE_WEIGHT(pSingle)           ((pSingle)->Weight)

// Original macro:
// #define FXU_HEAP_SINGLE_CURRENT(p, pSingle)       ((p)->pTree[(pSingle)->HNum])

// Original macro:
// #define FXU_HEAP_SINGLE_PARENT_EXISTS(p, pSingle) ((pSingle)->HNum > 1)

// Original macro:
// #define FXU_HEAP_SINGLE_CHILD1_EXISTS(p, pSingle) (((pSingle)->HNum << 1) <= p->nItems)

// Original macro:
// #define FXU_HEAP_SINGLE_CHILD2_EXISTS(p, pSingle) ((((pSingle)->HNum << 1)+1) <= p->nItems)

// Original macro:
// #define FXU_HEAP_SINGLE_PARENT(p, pSingle)        ((p)->pTree[(pSingle)->HNum >> 1])

// Original macro:
// #define FXU_HEAP_SINGLE_CHILD1(p, pSingle)        ((p)->pTree[(pSingle)->HNum << 1])

// Original macro:
// #define FXU_HEAP_SINGLE_CHILD2(p, pSingle)        ((p)->pTree[((pSingle)->HNum << 1)+1])

// Original macro:
// #define FXU_HEAP_SINGLE_ASSERT(p, pSingle)        assert( (pSingle)->HNum >= 1 && (pSingle)->HNum <= p->nItemsAlloc )


// static void Fxu_HeapSingleResize( Fxu_HeapSingle * p );
                  
// static void Fxu_HeapSingleSwap( Fxu_Single ** pSingle1, Fxu_Single ** pSingle2 );
  
// static void Fxu_HeapSingleMoveUp( Fxu_HeapSingle * p, Fxu_Single * pSingle );
  
// static void Fxu_HeapSingleMoveDn( Fxu_HeapSingle * p, Fxu_Single * pSingle );
  

////////////////////////////////////////////////////////////////////////
///                     FUNCTION DEFINITIONS                         ///
////////////////////////////////////////////////////////////////////////

/**Function*************************************************************

  Synopsis    []

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Fxu_HeapSingleStart() *Fxu_HeapSingle {
    var p *Fxu_HeapSingle;
    // inlining of macro ABC_ALLOC
    p = ((Fxu_HeapSingle *) malloc(sizeof(Fxu_HeapSingle) * (size_t)(1)));
    memset( p, 0, sizeof(Fxu_HeapSingle) );
    p.nItems      = 0;
    p.nItemsAlloc = 2000;
    // inlining of macro ABC_ALLOC
    p.pTree       = ((Fxu_Single * *) malloc(sizeof(Fxu_Single *) * (size_t)(p.nItemsAlloc + 10)));
    p.pTree[0]    = nil;
    return p;
}


/**Function*************************************************************

  Synopsis    []

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Fxu_HeapSingleResize(p *Fxu_HeapSingle) {
    p.nItemsAlloc *= 2;
    // inlining of macro ABC_REALLOC
    
    if p.pTree {
        p.pTree = ((Fxu_Single * *) realloc((char *)(p.pTree), sizeof(Fxu_Single *) * (size_t)(p.nItemsAlloc + 10)));
    } else {
        p.pTree = ((Fxu_Single * *) malloc(sizeof(Fxu_Single *) * (size_t)(p.nItemsAlloc + 10)));
    }
}

/**Function*************************************************************

  Synopsis    []

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Fxu_HeapSingleStop(p *Fxu_HeapSingle) {
    var i int32;
    i = 0;
    p.pTree = nil;
    i = 1;
    p = nil;
}


/**Function*************************************************************

  Synopsis    []

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Fxu_HeapSinglePrint(pFile *os.File, p *Fxu_HeapSingle) {
    var pSingle *Fxu_Single;
    var Counter int32  = 1;
    var Degree int32   = 1;

    Fxu_HeapSingleCheck( p );
    fmt.Fprintf(pFile, "The contents of the heap:\n");
    fmt.Fprintf(pFile, "Level %d:  ", Degree);
    // inlining of macro Fxu_HeapSingleForEachItem
    for  p.i = 1;          p.i <= p.nItems && (pSingle = p.pTree[p.i]);          p.i++  {
        assert(Counter == p.pTree[Counter].HNum, "Counter == p.pTree[Counter].HNum");
        // inlining of macro FXU_HEAP_SINGLE_WEIGHT
        fmt.Fprintf(pFile, "%2d=%3d  ", Counter, ((p.pTree[Counter]).Weight));
        if Counter++; Counter == (1 << Degree) {
            fmt.Fprintf(pFile, "\n");
            Degree++;
            fmt.Fprintf(pFile, "Level %d:  ", Degree);
        }
    }
    fmt.Fprintf(pFile, "\n");
    fmt.Fprintf(pFile, "End of the heap printout.\n");
}

/**Function*************************************************************

  Synopsis    []

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Fxu_HeapSingleCheck(p *Fxu_HeapSingle) {
    var pSingle *Fxu_Single;
    // inlining of macro Fxu_HeapSingleForEachItem
    for  p.i = 1;          p.i <= p.nItems && (pSingle = p.pTree[p.i]);          p.i++  {
        assert(pSingle.HNum == p.i, "pSingle.HNum == p.i");
        Fxu_HeapSingleCheckOne( p, pSingle );
    }
}

/**Function*************************************************************

  Synopsis    []

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Fxu_HeapSingleCheckOne(p *Fxu_HeapSingle, pSingle *Fxu_Single) {
    var Weight1, Weight2 int32;
    // inlining of macro FXU_HEAP_SINGLE_CHILD1_EXISTS
    if  ((pSingle).HNum << 1) <= p.nItems {
        // inlining of macro FXU_HEAP_SINGLE_WEIGHT
        Weight1 = ((pSingle).Weight);
        // inlining of macro FXU_HEAP_SINGLE_WEIGHT
        // inlining of macro FXU_HEAP_SINGLE_CHILD1
        Weight2 = ((((p).pTree[(pSingle).HNum << 1])).Weight);
        assert(Weight1 >= Weight2, "Weight1 >= Weight2");
    }
    // inlining of macro FXU_HEAP_SINGLE_CHILD2_EXISTS
    if  (((pSingle).HNum << 1)+1) <= p.nItems {
        // inlining of macro FXU_HEAP_SINGLE_WEIGHT
        Weight1 = ((pSingle).Weight);
        // inlining of macro FXU_HEAP_SINGLE_WEIGHT
        // inlining of macro FXU_HEAP_SINGLE_CHILD2
        Weight2 = ((((p).pTree[((pSingle).HNum << 1)+1])).Weight);
        assert(Weight1 >= Weight2, "Weight1 >= Weight2");
    }
}

/**Function*************************************************************

  Synopsis    []

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Fxu_HeapSingleInsert(p *Fxu_HeapSingle, pSingle *Fxu_Single) {
    if p.nItems == p.nItemsAlloc {
        Fxu_HeapSingleResize( p ); }
    // put the last entry to the last place and move up
    p.pTree[++p.nItems] = pSingle;
    pSingle.HNum = p.nItems;
    // move the last entry up if necessary
    Fxu_HeapSingleMoveUp( p, pSingle );
}

/**Function*************************************************************

  Synopsis    []

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Fxu_HeapSingleUpdate(p *Fxu_HeapSingle, pSingle *Fxu_Single) {
    // inlining of macro FXU_HEAP_SINGLE_ASSERT
    assert((pSingle).HNum >= 1 && (pSingle).HNum <= p.nItemsAlloc, "(pSingle).HNum >= 1 && (pSingle).HNum <= p.nItemsAlloc");
    // inlining of macro FXU_HEAP_SINGLE_PARENT_EXISTS
    if  ((pSingle).HNum > 1) && 
         // inlining of macro FXU_HEAP_SINGLE_WEIGHT
         // inlining of macro FXU_HEAP_SINGLE_PARENT
         ((pSingle).Weight) > ((((p).pTree[(pSingle).HNum >> 1])).Weight) {
        Fxu_HeapSingleMoveUp( p, pSingle );
    // inlining of macro FXU_HEAP_SINGLE_CHILD1_EXISTS
    } else if  (((pSingle).HNum << 1) <= p.nItems) && 
        // inlining of macro FXU_HEAP_SINGLE_WEIGHT
        // inlining of macro FXU_HEAP_SINGLE_CHILD1
        ((pSingle).Weight) < ((((p).pTree[(pSingle).HNum << 1])).Weight) {
        Fxu_HeapSingleMoveDn( p, pSingle );
    // inlining of macro FXU_HEAP_SINGLE_CHILD2_EXISTS
    } else if  ((((pSingle).HNum << 1)+1) <= p.nItems) && 
        // inlining of macro FXU_HEAP_SINGLE_WEIGHT
        // inlining of macro FXU_HEAP_SINGLE_CHILD2
        ((pSingle).Weight) < ((((p).pTree[((pSingle).HNum << 1)+1])).Weight) {
        Fxu_HeapSingleMoveDn( p, pSingle ); }
}

/**Function*************************************************************

  Synopsis    []

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Fxu_HeapSingleDelete(p *Fxu_HeapSingle, pSingle *Fxu_Single) {
    var Place int32  = pSingle.HNum;
    // inlining of macro FXU_HEAP_SINGLE_ASSERT
    assert((pSingle).HNum >= 1 && (pSingle).HNum <= p.nItemsAlloc, "(pSingle).HNum >= 1 && (pSingle).HNum <= p.nItemsAlloc");
    // put the last entry to the deleted place
    // decrement the number of entries
    p.pTree[Place] = p.pTree[p.nItems--];
    p.pTree[Place].HNum = Place;
    // move the top entry down if necessary
    Fxu_HeapSingleUpdate( p, p.pTree[Place] );
    pSingle.HNum = 0;
}

/**Function*************************************************************

  Synopsis    []

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Fxu_HeapSingleReadMax(p *Fxu_HeapSingle) *Fxu_Single {
    if p.nItems == 0 {
        return nil; }
    return p.pTree[1];
}

/**Function*************************************************************

  Synopsis    []

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Fxu_HeapSingleGetMax(p *Fxu_HeapSingle) *Fxu_Single {
    var pSingle *Fxu_Single;
    if p.nItems == 0 {
        return nil; }
    // prepare the return value
    pSingle = p.pTree[1];
    pSingle.HNum = 0;
    // put the last entry on top
    // decrement the number of entries
    p.pTree[1] = p.pTree[p.nItems--];
    p.pTree[1].HNum = 1;
    // move the top entry down if necessary
    Fxu_HeapSingleMoveDn( p, p.pTree[1] );
    return pSingle;     
}

/**Function*************************************************************

  Synopsis    []

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Fxu_HeapSingleReadMaxWeight(p *Fxu_HeapSingle) int32 {
    if p.nItems == 0 {
        return -1; }
    // inlining of macro FXU_HEAP_SINGLE_WEIGHT
    return ((p.pTree[1]).Weight);
}



/**Function*************************************************************

  Synopsis    []

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Fxu_HeapSingleSwap(pSingle1 **Fxu_Single, pSingle2 **Fxu_Single) {
    var pSingle *Fxu_Single;
    var Temp int32;
    pSingle   = *pSingle1;
    *pSingle1 = *pSingle2;
    *pSingle2 = pSingle;
    Temp          = (*pSingle1).HNum;
    (*pSingle1).HNum = (*pSingle2).HNum;
    (*pSingle2).HNum = Temp;
}

/**Function*************************************************************

  Synopsis    []

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Fxu_HeapSingleMoveUp(p *Fxu_HeapSingle, pSingle *Fxu_Single) {
    var ppSingle, ppPar **Fxu_Single;
    // inlining of macro FXU_HEAP_SINGLE_CURRENT
    ppSingle = &((p).pTree[(pSingle).HNum]);
    // inlining of macro FXU_HEAP_SINGLE_PARENT_EXISTS
    for  (*ppSingle).HNum > 1  {
        // inlining of macro FXU_HEAP_SINGLE_PARENT
        ppPar = &((p).pTree[(*ppSingle).HNum >> 1]);
        // inlining of macro FXU_HEAP_SINGLE_WEIGHT
        if  ((*ppSingle).Weight) > ((*ppPar).Weight) {
            Fxu_HeapSingleSwap( ppSingle, ppPar );
            ppSingle = ppPar;
        
        } else {
            break; }
    }
}
 
/**Function*************************************************************

  Synopsis    []

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Fxu_HeapSingleMoveDn(p *Fxu_HeapSingle, pSingle *Fxu_Single) {
    var ppChild1, ppChild2, ppSingle **Fxu_Single;
    // inlining of macro FXU_HEAP_SINGLE_CURRENT
    ppSingle = &((p).pTree[(pSingle).HNum]);
    // inlining of macro FXU_HEAP_SINGLE_CHILD1_EXISTS
    for  ((*ppSingle).HNum << 1) <= p.nItems  { // if Child1 does not exist, Child2 also does not exists

        // get the children
        // inlining of macro FXU_HEAP_SINGLE_CHILD1
        ppChild1 = &((p).pTree[(*ppSingle).HNum << 1]);
        // inlining of macro FXU_HEAP_SINGLE_CHILD2_EXISTS
        if  (((*ppSingle).HNum << 1)+1) <= p.nItems {
            // inlining of macro FXU_HEAP_SINGLE_CHILD2
            ppChild2 = &((p).pTree[((*ppSingle).HNum << 1)+1]);

            // consider two cases
            // inlining of macro FXU_HEAP_SINGLE_WEIGHT
            if  ((*ppSingle).Weight) >= ((*ppChild1).Weight) &&
                 // inlining of macro FXU_HEAP_SINGLE_WEIGHT
                 ((*ppSingle).Weight) >= ((*ppChild2).Weight) { // Var is larger than both, skip
                break;
            
            } else { // Var is smaller than one of them, then swap it with larger 
                // inlining of macro FXU_HEAP_SINGLE_WEIGHT
                if  ((*ppChild1).Weight) >= ((*ppChild2).Weight) {
                    Fxu_HeapSingleSwap( ppSingle, ppChild1 );
                    // update the pointer
                    ppSingle = ppChild1;
                
                } else {
                    Fxu_HeapSingleSwap( ppSingle, ppChild2 );
                    // update the pointer
                    ppSingle = ppChild2;
                }
            }
        
        } else { // Child2 does not exist
        
            // consider two cases
            // inlining of macro FXU_HEAP_SINGLE_WEIGHT
            if  ((*ppSingle).Weight) >= ((*ppChild1).Weight) { // Var is larger than Child1, skip
                break;
            
            } else { // Var is smaller than Child1, then swap them
                Fxu_HeapSingleSwap( ppSingle, ppChild1 );
                // update the pointer
                ppSingle = ppChild1;
            }
        }
    }
}


////////////////////////////////////////////////////////////////////////
///                       END OF FILE                                ///
////////////////////////////////////////////////////////////////////////
//ABC_NAMESPACE_IMPL_END

