package mvc

/**CFile****************************************************************

  FileName    [mvcList.c]

  PackageName [MVSIS 2.0: Multi-valued logic synthesis system.]

  Synopsis    [Manipulating list of cubes in the cover.]

  Author      [MVSIS Group]
  
  Affiliation [UC Berkeley]

  Date        [Ver. 1.0. Started - February 1, 2003.]

  Revision    [$Id: mvcList.c,v 1.4 2003/04/03 06:31:50 alanmi Exp $]

***********************************************************************/

//#include "mvc.h"

//ABC_NAMESPACE_IMPL_START


////////////////////////////////////////////////////////////////////////
///                        DECLARATIONS                              ///
////////////////////////////////////////////////////////////////////////

////////////////////////////////////////////////////////////////////////
///                     FUNCTION DEFINITIONS                         ///
////////////////////////////////////////////////////////////////////////

/**Function*************************************************************

  Synopsis    []

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Mvc_ListAddCubeHead_(pList *Mvc_List_t, pCube *Mvc_Cube_t) {
    if pList.pHead == nil {
        // inlining of macro Mvc_CubeSetNext
        ((pCube).pNext = (nil));
        pList.pHead = pCube;
        pList.pTail = pCube;
    
    } else {
        // inlining of macro Mvc_CubeSetNext
        ((pCube).pNext = (pList.pHead));
        pList.pHead = pCube;
    }
    pList.nItems++;
}


/**Function*************************************************************

  Synopsis    []

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Mvc_ListAddCubeTail_(pList *Mvc_List_t, pCube *Mvc_Cube_t) {
    if pList.pHead == nil {
        pList.pHead = pCube;
    } else {
        // inlining of macro Mvc_CubeSetNext
        ((pList.pTail).pNext = (pCube)); }
    pList.pTail    = pCube;
    // inlining of macro Mvc_CubeSetNext
    ((pCube).pNext = (nil));
    pList.nItems++;
}


/**Function*************************************************************

  Synopsis    []

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Mvc_ListDeleteCube_(pList *Mvc_List_t, pPrev *Mvc_Cube_t, pCube *Mvc_Cube_t) {
    if pPrev == nil { // deleting the head cube
        // inlining of macro Mvc_CubeReadNext
        pList.pHead = ((pCube).pNext);
    } else {
        pPrev.pNext = pCube.pNext; }
    if pList.pTail == pCube { // deleting the tail cube
    
        // inlining of macro Mvc_CubeReadNext
        assert(((pCube).pNext) == nil, "/*__starting_point_inline_macro:Mvc_CubeReadNext__*/((pCube).pNext) == nil");
        pList.pTail = pPrev;
    }
    pList.nItems--;
}



/**Function*************************************************************

  Synopsis    []

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Mvc_CoverAddCubeHead_(pCover *Mvc_Cover_t, pCube *Mvc_Cube_t) {
    var pList *Mvc_List_t  = &pCover.lCubes;
    if pList.pHead == nil {
        // inlining of macro Mvc_CubeSetNext
        ((pCube).pNext = (nil));
        pList.pHead = pCube;
        pList.pTail = pCube;
    
    } else {
        // inlining of macro Mvc_CubeSetNext
        ((pCube).pNext = (pList.pHead));
        pList.pHead = pCube;
    }
    pList.nItems++;
}

/**Function*************************************************************

  Synopsis    []

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Mvc_CoverAddCubeTail_(pCover *Mvc_Cover_t, pCube *Mvc_Cube_t) {
    var pList *Mvc_List_t  = &pCover.lCubes;

    if pList.pHead == nil {
        pList.pHead = pCube;
    } else {
        // inlining of macro Mvc_CubeSetNext
        ((pList.pTail).pNext = (pCube)); }
    pList.pTail    = pCube;
    // inlining of macro Mvc_CubeSetNext
    ((pCube).pNext = (nil));
    pList.nItems++;
}

/**Function*************************************************************

  Synopsis    []

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Mvc_CoverDeleteCube_(pCover *Mvc_Cover_t, pPrev *Mvc_Cube_t, pCube *Mvc_Cube_t) {
    var pList *Mvc_List_t  = &pCover.lCubes;
 
    if pPrev == nil { // deleting the head cube
        // inlining of macro Mvc_CubeReadNext
        pList.pHead = ((pCube).pNext);
    } else {
        pPrev.pNext = pCube.pNext; }
    if pList.pTail == pCube { // deleting the tail cube
    
        // inlining of macro Mvc_CubeReadNext
        assert(((pCube).pNext) == nil, "/*__starting_point_inline_macro:Mvc_CubeReadNext__*/((pCube).pNext) == nil");
        pList.pTail = pPrev;
    }
    pList.nItems--;
}



/**Function*************************************************************

  Synopsis    []

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Mvc_CoverAddDupCubeHead(pCover *Mvc_Cover_t, pCube *Mvc_Cube_t) {
    var pCubeNew *Mvc_Cube_t;
    pCubeNew = Mvc_CubeAlloc( pCover );
    // inlining of macro Mvc_CubeBitCopy
    // inlining of macro Mvc_Cube1Words
    // inlining of macro Mvc_Cube1BitCopy
    // inlining of macro Mvc_Cube2Words
    // inlining of macro Mvc_Cube2BitCopy
    if  (pCubeNew).iLast == 0     { (((pCubeNew).pData[0]) = ((pCube).pData[0])); }    else if  (pCubeNew).iLast == 1{ (((pCubeNew).pData[0]) = ((pCube).pData[0]));
// inlining of macro Mvc_CubeNBitCopy
(((pCubeNew).pData[1])=  ((pCube).pData[1]));; }    else                             { {    var _i_ int32;    for _i_ = (pCubeNew).iLast; _i_ >= 0; _i_-- {        ((pCubeNew).pData[_i_]) = ((pCube).pData[_i_]); }}; };
    // inlining of macro Mvc_CoverAddCubeHead
    // inlining of macro Mvc_ListAddCubeHead
    // inlining of macro Mvc_CubeSetNext
    {    var pList *Mvc_List_t  = &pCover.lCubes;    {        if pList.pHead == nil        {            ((pCubeNew).pNext = (nil));            pList.pHead = pCubeNew;            pList.pTail = pCubeNew;        }        else        {            ((pCubeNew).pNext = (pList.pHead));            pList.pHead = pCubeNew;        }
        pList.nItems++;    };};
}

/**Function*************************************************************

  Synopsis    []

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Mvc_CoverAddDupCubeTail(pCover *Mvc_Cover_t, pCube *Mvc_Cube_t) {
    var pCubeNew *Mvc_Cube_t;
    // copy the cube as part of this cover
    pCubeNew = Mvc_CubeAlloc( pCover );
    // inlining of macro Mvc_CubeBitCopy
    // inlining of macro Mvc_Cube1Words
    // inlining of macro Mvc_Cube1BitCopy
    // inlining of macro Mvc_Cube2Words
    // inlining of macro Mvc_Cube2BitCopy
    if  (pCubeNew).iLast == 0     { (((pCubeNew).pData[0]) = ((pCube).pData[0])); }    else if  (pCubeNew).iLast == 1{ (((pCubeNew).pData[0]) = ((pCube).pData[0]));
// inlining of macro Mvc_CubeNBitCopy
(((pCubeNew).pData[1])=  ((pCube).pData[1]));; }    else                             { {    var _i_ int32;    for _i_ = (pCubeNew).iLast; _i_ >= 0; _i_-- {        ((pCubeNew).pData[_i_]) = ((pCube).pData[_i_]); }}; };
    // clean the last bits of the new cube
//    pCubeNew->pData[pCubeNew->iLast] &= (BITS_FULL >> pCubeNew->nUnused);
    // add the cube at the end
    // inlining of macro Mvc_CoverAddCubeTail
    // inlining of macro Mvc_ListAddCubeTail
    // inlining of macro Mvc_CubeSetNext
    {    var pList *Mvc_List_t  = &pCover.lCubes;    {        if pList.pHead == nil {            pList.pHead = pCubeNew; }        else {            ((pList.pTail).pNext = (pCubeNew)); }
        // inlining of macro Mvc_CubeSetNext
        pList.pTail    = pCubeNew;        ((pCubeNew).pNext = (nil));        pList.nItems++;    };};
}


/**Function*************************************************************

  Synopsis    []

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Mvc_CoverAddLiteralsOfCube(pCover *Mvc_Cover_t, pCube *Mvc_Cube_t) {
//    int iBit, Value;
//    assert( pCover->pLits );
//    Mvc_CubeForEachBit( pCover, pCube, iBit, Value )
//        if ( Value )
//            pCover->pLits[iBit] += Value;
}

/**Function*************************************************************

  Synopsis    []

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Mvc_CoverDeleteLiteralsOfCube(pCover *Mvc_Cover_t, pCube *Mvc_Cube_t) {
//    int iBit, Value;
//    assert( pCover->pLits );
//    Mvc_CubeForEachBit( pCover, pCube, iBit, Value )
//        if ( Value )
//            pCover->pLits[iBit] -= Value;
}


/**Function*************************************************************

  Synopsis    [Transfers the cubes from the list into the array.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Mvc_CoverList2Array(pCover *Mvc_Cover_t) {
    var pCube *Mvc_Cube_t;
    var Counter int32;
    // resize storage if necessary
    Mvc_CoverAllocateArrayCubes( pCover );
    // iterate through the cubes
    Counter = 0;
    // inlining of macro Mvc_CoverForEachCube
    for  pCube = (pCover).lCubes.pHead;          (pCube != nil);          pCube = pCube.pNext  {
        pCover.pCubes[ Counter++ ] = pCube;
    }
    assert(Counter == Mvc_CoverReadCubeNum(pCover), "Counter == Mvc_CoverReadCubeNum(pCover)");
}

/**Function*************************************************************

  Synopsis    [Transfers the cubes from the array into list.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Mvc_CoverArray2List(pCover *Mvc_Cover_t) {
    var pCube *Mvc_Cube_t;
    var nCubes, i int32;

    assert(pCover.pCubes, "pCover.pCubes");

    nCubes = Mvc_CoverReadCubeNum(pCover);
    if nCubes == 0 {
        return; }
    if nCubes == 1 {
        pCube    = pCover.pCubes[0];
        pCube.pNext = nil;
        pCover.lCubes.pHead = pCover.lCubes.pTail = pCube;
        return;
    }
    // set up the first cube
    pCube = pCover.pCubes[0];
    pCover.lCubes.pHead = pCube;
    // set up the last cube
    pCube = pCover.pCubes[nCubes-1];
    pCube.pNext = nil;
    pCover.lCubes.pTail = pCube;

    // link all cubes starting from the first one
    for  i = 0; i < nCubes - 1; i++  {
        pCover.pCubes[i].pNext = pCover.pCubes[i+1];
    }
}

/**Function*************************************************************

  Synopsis    [Returns the tail of the linked list given by the head.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Mvc_ListGetTailFromHead(pHead *Mvc_Cube_t) *Mvc_Cube_t {
    var pCube, pTail *Mvc_Cube_t;
    for  pTail = pCube = pHead; 
          (pCube != nil); 
          // inlining of macro Mvc_CubeReadNext
          pTail = pCube, pCube = ((pCube).pNext)  {}
    return pTail;
}


////////////////////////////////////////////////////////////////////////
///                       END OF FILE                                ///
////////////////////////////////////////////////////////////////////////


//ABC_NAMESPACE_IMPL_END

