package fxu

import "fmt"

/**CFile****************************************************************

  FileName    [fxuList.c]

  PackageName [MVSIS 2.0: Multi-valued logic synthesis system.]

  Synopsis    [Operations on lists.]

  Author      [MVSIS Group]
  
  Affiliation [UC Berkeley]

  Date        [Ver. 1.0. Started - February 1, 2003.]

  Revision    [$Id: fxuList.c,v 1.0 2003/02/01 00:00:00 alanmi Exp $]

***********************************************************************/

//#include "fxuInt.h"

//ABC_NAMESPACE_IMPL_START


////////////////////////////////////////////////////////////////////////
///                        DECLARATIONS                              ///
////////////////////////////////////////////////////////////////////////

////////////////////////////////////////////////////////////////////////
///                     FUNCTION DEFINITIONS                         ///
////////////////////////////////////////////////////////////////////////

// matrix -> var

/**Function*************************************************************

  Synopsis    []

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Fxu_ListMatrixAddVariable(p *Fxu_Matrix, pLink *Fxu_Var) {
    var pList *Fxu_ListVar  = &p.lVars;
    if pList.pHead == nil {
        pList.pHead = pLink;
        pList.pTail = pLink;
        pLink.pPrev = nil;
        pLink.pNext = nil;
    
    } else {
        pLink.pNext = nil;
        pList.pTail.pNext = pLink;
        pLink.pPrev = pList.pTail;
        pList.pTail = pLink;
    }
    pList.nItems++;
}

/**Function*************************************************************

  Synopsis    []

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Fxu_ListMatrixDelVariable(p *Fxu_Matrix, pLink *Fxu_Var) {
    var pList *Fxu_ListVar  = &p.lVars;
    if pList.pHead == pLink {
         pList.pHead = pLink.pNext; }
    if pList.pTail == pLink {
         pList.pTail = pLink.pPrev; }
    if pLink.pPrev {
         pLink.pPrev.pNext = pLink.pNext; }
    if pLink.pNext {
         pLink.pNext.pPrev = pLink.pPrev; }
    pList.nItems--;
}


// matrix -> cube

/**Function*************************************************************

  Synopsis    []

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Fxu_ListMatrixAddCube(p *Fxu_Matrix, pLink *Fxu_Cube) {
    var pList *Fxu_ListCube  = &p.lCubes;
    if pList.pHead == nil {
        pList.pHead = pLink;
        pList.pTail = pLink;
        pLink.pPrev = nil;
        pLink.pNext = nil;
    
    } else {
        pLink.pNext = nil;
        pList.pTail.pNext = pLink;
        pLink.pPrev = pList.pTail;
        pList.pTail = pLink;
    }
    pList.nItems++;
}

/**Function*************************************************************

  Synopsis    []

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Fxu_ListMatrixDelCube(p *Fxu_Matrix, pLink *Fxu_Cube) {
    var pList *Fxu_ListCube  = &p.lCubes;
    if pList.pHead == pLink {
         pList.pHead = pLink.pNext; }
    if pList.pTail == pLink {
         pList.pTail = pLink.pPrev; }
    if pLink.pPrev {
         pLink.pPrev.pNext = pLink.pNext; }
    if pLink.pNext {
         pLink.pNext.pPrev = pLink.pPrev; }
    pList.nItems--;
}


// matrix -> single

/**Function*************************************************************

  Synopsis    []

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Fxu_ListMatrixAddSingle(p *Fxu_Matrix, pLink *Fxu_Single) {
    var pList *Fxu_ListSingle  = &p.lSingles;
    if pList.pHead == nil {
        pList.pHead = pLink;
        pList.pTail = pLink;
        pLink.pPrev = nil;
        pLink.pNext = nil;
    
    } else {
        pLink.pNext = nil;
        pList.pTail.pNext = pLink;
        pLink.pPrev = pList.pTail;
        pList.pTail = pLink;
    }
    pList.nItems++;
}

/**Function*************************************************************

  Synopsis    []

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Fxu_ListMatrixDelSingle(p *Fxu_Matrix, pLink *Fxu_Single) {
    var pList *Fxu_ListSingle  = &p.lSingles;
    if pList.pHead == pLink {
         pList.pHead = pLink.pNext; }
    if pList.pTail == pLink {
         pList.pTail = pLink.pPrev; }
    if pLink.pPrev {
         pLink.pPrev.pNext = pLink.pNext; }
    if pLink.pNext {
         pLink.pNext.pPrev = pLink.pPrev; }
    pList.nItems--;
}


// table -> divisor

/**Function*************************************************************

  Synopsis    []

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Fxu_ListTableAddDivisor(p *Fxu_Matrix, pLink *Fxu_Double) {
    var pList *Fxu_ListDouble  = &(p.pTable[pLink.Key]);
    if pList.pHead == nil {
        pList.pHead = pLink;
        pList.pTail = pLink;
        pLink.pPrev = nil;
        pLink.pNext = nil;
    
    } else {
        pLink.pNext = nil;
        pList.pTail.pNext = pLink;
        pLink.pPrev = pList.pTail;
        pList.pTail = pLink;
    }
    pList.nItems++;
    p.nDivs++;
}

/**Function*************************************************************

  Synopsis    []

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Fxu_ListTableDelDivisor(p *Fxu_Matrix, pLink *Fxu_Double) {
    var pList *Fxu_ListDouble  = &(p.pTable[pLink.Key]);
    if pList.pHead == pLink {
         pList.pHead = pLink.pNext; }
    if pList.pTail == pLink {
         pList.pTail = pLink.pPrev; }
    if pLink.pPrev {
         pLink.pPrev.pNext = pLink.pNext; }
    if pLink.pNext {
         pLink.pNext.pPrev = pLink.pPrev; }
    pList.nItems--;
    p.nDivs--;
}


// cube -> literal 

/**Function*************************************************************

  Synopsis    []

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Fxu_ListCubeAddLiteral(pCube *Fxu_Cube, pLink *Fxu_Lit) {
    var pList *Fxu_ListLit  = &(pCube.lLits);
    if pList.pHead == nil {
        pList.pHead = pLink;
        pList.pTail = pLink;
        pLink.pHPrev = nil;
        pLink.pHNext = nil;
    
    } else {
        pLink.pHNext = nil;
        pList.pTail.pHNext = pLink;
        pLink.pHPrev = pList.pTail;
        pList.pTail = pLink;
    }
    pList.nItems++;
}

/**Function*************************************************************

  Synopsis    []

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Fxu_ListCubeDelLiteral(pCube *Fxu_Cube, pLink *Fxu_Lit) {
    var pList *Fxu_ListLit  = &(pCube.lLits);
    if pList.pHead == pLink {
         pList.pHead = pLink.pHNext; }
    if pList.pTail == pLink {
         pList.pTail = pLink.pHPrev; }
    if pLink.pHPrev {
         pLink.pHPrev.pHNext = pLink.pHNext; }
    if pLink.pHNext {
         pLink.pHNext.pHPrev = pLink.pHPrev; }
    pList.nItems--;
}


// var -> literal

/**Function*************************************************************

  Synopsis    []

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Fxu_ListVarAddLiteral(pVar *Fxu_Var, pLink *Fxu_Lit) {
    var pList *Fxu_ListLit  = &(pVar.lLits);
    if pList.pHead == nil {
        pList.pHead = pLink;
        pList.pTail = pLink;
        pLink.pVPrev = nil;
        pLink.pVNext = nil;
    
    } else {
        pLink.pVNext = nil;
        pList.pTail.pVNext = pLink;
        pLink.pVPrev = pList.pTail;
        pList.pTail = pLink;
    }
    pList.nItems++;
}

/**Function*************************************************************

  Synopsis    []

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Fxu_ListVarDelLiteral(pVar *Fxu_Var, pLink *Fxu_Lit) {
    var pList *Fxu_ListLit  = &(pVar.lLits);
    if pList.pHead == pLink {
         pList.pHead = pLink.pVNext; }
    if pList.pTail == pLink {
         pList.pTail = pLink.pVPrev; }
    if pLink.pVPrev {
         pLink.pVPrev.pVNext = pLink.pVNext; }
    if pLink.pVNext {
         pLink.pVNext.pVPrev = pLink.pVPrev; }
    pList.nItems--;
}



// divisor -> pair

/**Function*************************************************************

  Synopsis    []

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Fxu_ListDoubleAddPairLast(pDiv *Fxu_Double, pLink *Fxu_Pair) {
    var pList *Fxu_ListPair  = &pDiv.lPairs;
    if pList.pHead == nil {
        pList.pHead = pLink;
        pList.pTail = pLink;
        pLink.pDPrev = nil;
        pLink.pDNext = nil;
    
    } else {
        pLink.pDNext = nil;
        pList.pTail.pDNext = pLink;
        pLink.pDPrev = pList.pTail;
        pList.pTail = pLink;
    }
    pList.nItems++;
}

/**Function*************************************************************

  Synopsis    []

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Fxu_ListDoubleAddPairFirst(pDiv *Fxu_Double, pLink *Fxu_Pair) {
    var pList *Fxu_ListPair  = &pDiv.lPairs;
    if pList.pHead == nil {
        pList.pHead = pLink;
        pList.pTail = pLink;
        pLink.pDPrev = nil;
        pLink.pDNext = nil;
    
    } else {
        pLink.pDPrev = nil;
        pList.pHead.pDPrev = pLink;
        pLink.pDNext = pList.pHead;
        pList.pHead = pLink;
    }
    pList.nItems++;
}

/**Function*************************************************************

  Synopsis    [Adds the entry in the middle of the list after the spot.]

  Description [Assumes that spot points to the link, after which the given
  link should be added. Spot cannot be NULL or the tail of the list.
  Therefore, the head and the tail of the list are not changed.]
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Fxu_ListDoubleAddPairMiddle(pDiv *Fxu_Double, pSpot *Fxu_Pair, pLink *Fxu_Pair) {
    var pList *Fxu_ListPair  = &pDiv.lPairs;
    assert((pSpot != nil), "pSpot");
    assert(pSpot != pList.pTail, "pSpot != pList.pTail");
    pLink.pDPrev = pSpot;
    pLink.pDNext = pSpot.pDNext;
    pLink.pDPrev.pDNext = pLink;
    pLink.pDNext.pDPrev = pLink;
    pList.nItems++;
}

/**Function*************************************************************

  Synopsis    []

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Fxu_ListDoubleDelPair(pDiv *Fxu_Double, pLink *Fxu_Pair) {
    var pList *Fxu_ListPair  = &pDiv.lPairs;
    if pList.pHead == pLink {
         pList.pHead = pLink.pDNext; }
    if pList.pTail == pLink {
         pList.pTail = pLink.pDPrev; }
    if pLink.pDPrev {
         pLink.pDPrev.pDNext = pLink.pDNext; }
    if pLink.pDNext {
         pLink.pDNext.pDPrev = pLink.pDPrev; }
    pList.nItems--;
}

/**Function*************************************************************

  Synopsis    []

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Fxu_ListDoubleAddPairPlace(pDiv *Fxu_Double, pPair *Fxu_Pair, pPairSpot *Fxu_Pair) {
    fmt.Printf("Fxu_ListDoubleAddPairPlace() is called!\n");
}



////////////////////////////////////////////////////////////////////////
///                       END OF FILE                                ///
////////////////////////////////////////////////////////////////////////


//ABC_NAMESPACE_IMPL_END

