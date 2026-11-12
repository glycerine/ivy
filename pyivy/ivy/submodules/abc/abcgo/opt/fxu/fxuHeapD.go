package fxu

import "fmt"

import "os"

/**CFile****************************************************************

  FileName    [fxuHeapD.c]

  PackageName [MVSIS 2.0: Multi-valued logic synthesis system.]

  Synopsis    [The priority queue for double cube divisors.]

  Author      [MVSIS Group]
  
  Affiliation [UC Berkeley]

  Date        [Ver. 1.0. Started - February 1, 2003.]

  Revision    [$Id: fxuHeapD.c,v 1.0 2003/02/01 00:00:00 alanmi Exp $]

***********************************************************************/

//#include "fxuInt.h"

//ABC_NAMESPACE_IMPL_START


////////////////////////////////////////////////////////////////////////
///                        DECLARATIONS                              ///
////////////////////////////////////////////////////////////////////////

// Original macro:
// #define FXU_HEAP_DOUBLE_WEIGHT(pDiv)           ((pDiv)->Weight)

// Original macro:
// #define FXU_HEAP_DOUBLE_CURRENT(p, pDiv)       ((p)->pTree[(pDiv)->HNum])

// Original macro:
// #define FXU_HEAP_DOUBLE_PARENT_EXISTS(p, pDiv) ((pDiv)->HNum > 1)

// Original macro:
// #define FXU_HEAP_DOUBLE_CHILD1_EXISTS(p, pDiv) (((pDiv)->HNum << 1) <= p->nItems)

// Original macro:
// #define FXU_HEAP_DOUBLE_CHILD2_EXISTS(p, pDiv) ((((pDiv)->HNum << 1)+1) <= p->nItems)

// Original macro:
// #define FXU_HEAP_DOUBLE_PARENT(p, pDiv)        ((p)->pTree[(pDiv)->HNum >> 1])

// Original macro:
// #define FXU_HEAP_DOUBLE_CHILD1(p, pDiv)        ((p)->pTree[(pDiv)->HNum << 1])

// Original macro:
// #define FXU_HEAP_DOUBLE_CHILD2(p, pDiv)        ((p)->pTree[((pDiv)->HNum << 1)+1])

// Original macro:
// #define FXU_HEAP_DOUBLE_ASSERT(p, pDiv)        assert( (pDiv)->HNum >= 1 && (pDiv)->HNum <= p->nItemsAlloc )


// static void Fxu_HeapDoubleResize( Fxu_HeapDouble * p );
                  
// static void Fxu_HeapDoubleSwap( Fxu_Double ** pDiv1, Fxu_Double ** pDiv2 );
  
// static void Fxu_HeapDoubleMoveUp( Fxu_HeapDouble * p, Fxu_Double * pDiv );
  
// static void Fxu_HeapDoubleMoveDn( Fxu_HeapDouble * p, Fxu_Double * pDiv );
  

////////////////////////////////////////////////////////////////////////
///                     FUNCTION DEFINITIONS                         ///
////////////////////////////////////////////////////////////////////////

/**Function*************************************************************

  Synopsis    []

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Fxu_HeapDoubleStart() *Fxu_HeapDouble {
    var p *Fxu_HeapDouble;
    // inlining of macro ABC_ALLOC
    p = ((Fxu_HeapDouble *) malloc(sizeof(Fxu_HeapDouble) * (size_t)(1)));
    memset( p, 0, sizeof(Fxu_HeapDouble) );
    p.nItems      = 0;
    p.nItemsAlloc = 10000;
    // inlining of macro ABC_ALLOC
    p.pTree       = ((Fxu_Double * *) malloc(sizeof(Fxu_Double *) * (size_t)(p.nItemsAlloc + 1)));
    p.pTree[0]    = nil;
    return p;
}


/**Function*************************************************************

  Synopsis    []

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Fxu_HeapDoubleResize(p *Fxu_HeapDouble) {
    p.nItemsAlloc *= 2;
    // inlining of macro ABC_REALLOC
    
    if p.pTree {
        p.pTree = ((Fxu_Double * *) realloc((char *)(p.pTree), sizeof(Fxu_Double *) * (size_t)(p.nItemsAlloc + 1)));
    } else {
        p.pTree = ((Fxu_Double * *) malloc(sizeof(Fxu_Double *) * (size_t)(p.nItemsAlloc + 1)));
    }
}

/**Function*************************************************************

  Synopsis    []

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Fxu_HeapDoubleStop(p *Fxu_HeapDouble) {
    // inlining of macro ABC_FREE
    ((p.pTree) ? (free((char *) (p.pTree)), (p.pTree) = 0) : 0);
    // inlining of macro ABC_FREE
    ((p) ? (free((char *) (p)), (p) = 0) : 0);
}


/**Function*************************************************************

  Synopsis    []

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Fxu_HeapDoublePrint(pFile *os.File, p *Fxu_HeapDouble) {
    var pDiv *Fxu_Double;
    var Counter int32  = 1;
    var Degree int32   = 1;

    Fxu_HeapDoubleCheck( p );
    fmt.Fprintf(pFile, "The contents of the heap:\n");
    fmt.Fprintf(pFile, "Level %d:  ", Degree);
    // inlining of macro Fxu_HeapDoubleForEachItem
    for  p.i = 1;          p.i <= p.nItems && (pDiv = p.pTree[p.i]);          p.i++  {
        assert(Counter == p.pTree[Counter].HNum, "Counter == p.pTree[Counter].HNum");
        // inlining of macro FXU_HEAP_DOUBLE_WEIGHT
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
func Fxu_HeapDoubleCheck(p *Fxu_HeapDouble) {
    var pDiv *Fxu_Double;
    // inlining of macro Fxu_HeapDoubleForEachItem
    for  p.i = 1;          p.i <= p.nItems && (pDiv = p.pTree[p.i]);          p.i++  {
        assert(pDiv.HNum == p.i, "pDiv.HNum == p.i");
        Fxu_HeapDoubleCheckOne( p, pDiv );
    }
}

/**Function*************************************************************

  Synopsis    []

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Fxu_HeapDoubleCheckOne(p *Fxu_HeapDouble, pDiv *Fxu_Double) {
    var Weight1, Weight2 int32;
    // inlining of macro FXU_HEAP_DOUBLE_CHILD1_EXISTS
    if  ((pDiv).HNum << 1) <= p.nItems {
        // inlining of macro FXU_HEAP_DOUBLE_WEIGHT
        Weight1 = ((pDiv).Weight);
        // inlining of macro FXU_HEAP_DOUBLE_WEIGHT
        // inlining of macro FXU_HEAP_DOUBLE_CHILD1
        Weight2 = ((((p).pTree[(pDiv).HNum << 1])).Weight);
        assert(Weight1 >= Weight2, "Weight1 >= Weight2");
    }
    // inlining of macro FXU_HEAP_DOUBLE_CHILD2_EXISTS
    if  (((pDiv).HNum << 1)+1) <= p.nItems {
        // inlining of macro FXU_HEAP_DOUBLE_WEIGHT
        Weight1 = ((pDiv).Weight);
        // inlining of macro FXU_HEAP_DOUBLE_WEIGHT
        // inlining of macro FXU_HEAP_DOUBLE_CHILD2
        Weight2 = ((((p).pTree[((pDiv).HNum << 1)+1])).Weight);
        assert(Weight1 >= Weight2, "Weight1 >= Weight2");
    }
}

/**Function*************************************************************

  Synopsis    []

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Fxu_HeapDoubleInsert(p *Fxu_HeapDouble, pDiv *Fxu_Double) {
    if p.nItems == p.nItemsAlloc {
        Fxu_HeapDoubleResize( p ); }
    // put the last entry to the last place and move up
    p.pTree[++p.nItems] = pDiv;
    pDiv.HNum = p.nItems;
    // move the last entry up if necessary
    Fxu_HeapDoubleMoveUp( p, pDiv );
}

/**Function*************************************************************

  Synopsis    []

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Fxu_HeapDoubleUpdate(p *Fxu_HeapDouble, pDiv *Fxu_Double) {
//printf( "Updating divisor %d.\n", pDiv->Num );

    // inlining of macro FXU_HEAP_DOUBLE_ASSERT
    assert((pDiv).HNum >= 1 && (pDiv).HNum <= p.nItemsAlloc, "(pDiv).HNum >= 1 && (pDiv).HNum <= p.nItemsAlloc");
    // inlining of macro FXU_HEAP_DOUBLE_PARENT_EXISTS
    if  ((pDiv).HNum > 1) && 
         // inlining of macro FXU_HEAP_DOUBLE_WEIGHT
         // inlining of macro FXU_HEAP_DOUBLE_PARENT
         ((pDiv).Weight) > ((((p).pTree[(pDiv).HNum >> 1])).Weight) {
        Fxu_HeapDoubleMoveUp( p, pDiv );
    // inlining of macro FXU_HEAP_DOUBLE_CHILD1_EXISTS
    } else if  (((pDiv).HNum << 1) <= p.nItems) && 
        // inlining of macro FXU_HEAP_DOUBLE_WEIGHT
        // inlining of macro FXU_HEAP_DOUBLE_CHILD1
        ((pDiv).Weight) < ((((p).pTree[(pDiv).HNum << 1])).Weight) {
        Fxu_HeapDoubleMoveDn( p, pDiv );
    // inlining of macro FXU_HEAP_DOUBLE_CHILD2_EXISTS
    } else if  ((((pDiv).HNum << 1)+1) <= p.nItems) && 
        // inlining of macro FXU_HEAP_DOUBLE_WEIGHT
        // inlining of macro FXU_HEAP_DOUBLE_CHILD2
        ((pDiv).Weight) < ((((p).pTree[((pDiv).HNum << 1)+1])).Weight) {
        Fxu_HeapDoubleMoveDn( p, pDiv ); }
}

/**Function*************************************************************

  Synopsis    []

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Fxu_HeapDoubleDelete(p *Fxu_HeapDouble, pDiv *Fxu_Double) {
    // inlining of macro FXU_HEAP_DOUBLE_ASSERT
    assert((pDiv).HNum >= 1 && (pDiv).HNum <= p.nItemsAlloc, "(pDiv).HNum >= 1 && (pDiv).HNum <= p.nItemsAlloc");
    // put the last entry to the deleted place
    // decrement the number of entries
    p.pTree[pDiv.HNum] = p.pTree[p.nItems--];
    p.pTree[pDiv.HNum].HNum = pDiv.HNum;
    // move the top entry down if necessary
    Fxu_HeapDoubleUpdate( p, p.pTree[pDiv.HNum] );
    pDiv.HNum = 0;
}

/**Function*************************************************************

  Synopsis    []

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Fxu_HeapDoubleReadMax(p *Fxu_HeapDouble) *Fxu_Double {
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
func Fxu_HeapDoubleGetMax(p *Fxu_HeapDouble) *Fxu_Double {
    var pDiv *Fxu_Double;
    if p.nItems == 0 {
        return nil; }
    // prepare the return value
    pDiv = p.pTree[1];
    pDiv.HNum = 0;
    // put the last entry on top
    // decrement the number of entries
    p.pTree[1] = p.pTree[p.nItems--];
    p.pTree[1].HNum = 1;
    // move the top entry down if necessary
    Fxu_HeapDoubleMoveDn( p, p.pTree[1] );
    return pDiv;     
}

/**Function*************************************************************

  Synopsis    []

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Fxu_HeapDoubleReadMaxWeight(p *Fxu_HeapDouble) int32 {
    if p.nItems == 0 {
        return -1;
    } else {
        // inlining of macro FXU_HEAP_DOUBLE_WEIGHT
        return ((p.pTree[1]).Weight); }
}



/**Function*************************************************************

  Synopsis    []

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Fxu_HeapDoubleSwap(pDiv1 **Fxu_Double, pDiv2 **Fxu_Double) {
    var pDiv *Fxu_Double;
    var Temp int32;
    pDiv   = *pDiv1;
    *pDiv1 = *pDiv2;
    *pDiv2 = pDiv;
    Temp          = (*pDiv1).HNum;
    (*pDiv1).HNum = (*pDiv2).HNum;
    (*pDiv2).HNum = Temp;
}

/**Function*************************************************************

  Synopsis    []

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Fxu_HeapDoubleMoveUp(p *Fxu_HeapDouble, pDiv *Fxu_Double) {
    var ppDiv, ppPar **Fxu_Double;
    // inlining of macro FXU_HEAP_DOUBLE_CURRENT
    ppDiv = &((p).pTree[(pDiv).HNum]);
    // inlining of macro FXU_HEAP_DOUBLE_PARENT_EXISTS
    for  (*ppDiv).HNum > 1  {
        // inlining of macro FXU_HEAP_DOUBLE_PARENT
        ppPar = &((p).pTree[(*ppDiv).HNum >> 1]);
        // inlining of macro FXU_HEAP_DOUBLE_WEIGHT
        if  ((*ppDiv).Weight) > ((*ppPar).Weight) {
            Fxu_HeapDoubleSwap( ppDiv, ppPar );
            ppDiv = ppPar;
        
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
func Fxu_HeapDoubleMoveDn(p *Fxu_HeapDouble, pDiv *Fxu_Double) {
    var ppChild1, ppChild2, ppDiv **Fxu_Double;
    // inlining of macro FXU_HEAP_DOUBLE_CURRENT
    ppDiv = &((p).pTree[(pDiv).HNum]);
    // inlining of macro FXU_HEAP_DOUBLE_CHILD1_EXISTS
    for  ((*ppDiv).HNum << 1) <= p.nItems  { // if Child1 does not exist, Child2 also does not exists

        // get the children
        // inlining of macro FXU_HEAP_DOUBLE_CHILD1
        ppChild1 = &((p).pTree[(*ppDiv).HNum << 1]);
        // inlining of macro FXU_HEAP_DOUBLE_CHILD2_EXISTS
        if  (((*ppDiv).HNum << 1)+1) <= p.nItems {
            // inlining of macro FXU_HEAP_DOUBLE_CHILD2
            ppChild2 = &((p).pTree[((*ppDiv).HNum << 1)+1]);

            // consider two cases
            // inlining of macro FXU_HEAP_DOUBLE_WEIGHT
            if  ((*ppDiv).Weight) >= ((*ppChild1).Weight) &&
                 // inlining of macro FXU_HEAP_DOUBLE_WEIGHT
                 ((*ppDiv).Weight) >= ((*ppChild2).Weight) { // Div is larger than both, skip
                break;
            
            } else { // Div is smaller than one of them, then swap it with larger 
                // inlining of macro FXU_HEAP_DOUBLE_WEIGHT
                if  ((*ppChild1).Weight) >= ((*ppChild2).Weight) {
                    Fxu_HeapDoubleSwap( ppDiv, ppChild1 );
                    // update the pointer
                    ppDiv = ppChild1;
                
                } else {
                    Fxu_HeapDoubleSwap( ppDiv, ppChild2 );
                    // update the pointer
                    ppDiv = ppChild2;
                }
            }
        
        } else { // Child2 does not exist
        
            // consider two cases
            // inlining of macro FXU_HEAP_DOUBLE_WEIGHT
            if  ((*ppDiv).Weight) >= ((*ppChild1).Weight) { // Div is larger than Child1, skip
                break;
            
            } else { // Div is smaller than Child1, then swap them
                Fxu_HeapDoubleSwap( ppDiv, ppChild1 );
                // update the pointer
                ppDiv = ppChild1;
            }
        }
    }
}


////////////////////////////////////////////////////////////////////////
///                       END OF FILE                                ///
////////////////////////////////////////////////////////////////////////


//ABC_NAMESPACE_IMPL_END

