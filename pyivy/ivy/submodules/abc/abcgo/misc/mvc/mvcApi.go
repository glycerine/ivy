package mvc

/**CFile****************************************************************

  FileName    [mvcApi.c]

  PackageName [MVSIS 2.0: Multi-valued logic synthesis system.]

  Synopsis    []

  Author      [MVSIS Group]
  
  Affiliation [UC Berkeley]

  Date        [Ver. 1.0. Started - February 1, 2003.]

  Revision    [$Id: mvcApi.c,v 1.4 2003/04/03 06:31:48 alanmi Exp $]

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
func Mvc_CoverReadWordNum(pCover *Mvc_Cover_t) int32 { return pCover.nWords;         }
func Mvc_CoverReadBitNum(pCover *Mvc_Cover_t) int32 { return pCover.nBits;          }
func Mvc_CoverReadCubeNum(pCover *Mvc_Cover_t) int32 { return pCover.lCubes.nItems; }
func Mvc_CoverReadCubeHead(pCover *Mvc_Cover_t) *Mvc_Cube_t { return pCover.lCubes.pHead;  }
func Mvc_CoverReadCubeTail(pCover *Mvc_Cover_t) *Mvc_Cube_t { return pCover.lCubes.pTail;  }
func Mvc_CoverReadCubeList(pCover *Mvc_Cover_t) *Mvc_List_t { return &pCover.lCubes;        }


/**Function*************************************************************

  Synopsis    []

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Mvc_ListReadCubeNum(pList *Mvc_List_t) int32 { return pList.nItems; }
func Mvc_ListReadCubeHead(pList *Mvc_List_t) *Mvc_Cube_t { return pList.pHead;  }
func Mvc_ListReadCubeTail(pList *Mvc_List_t) *Mvc_Cube_t { return pList.pTail;  }

/**Function*************************************************************

  Synopsis    []

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Mvc_CoverSetCubeNum(pCover *Mvc_Cover_t, nItems int32) { pCover.lCubes.nItems = nItems; }
func Mvc_CoverSetCubeHead(pCover *Mvc_Cover_t, pCube *Mvc_Cube_t) { pCover.lCubes.pHead = pCube;   }
func Mvc_CoverSetCubeTail(pCover *Mvc_Cover_t, pCube *Mvc_Cube_t) { pCover.lCubes.pTail = pCube;   }
func Mvc_CoverSetCubeList(pCover *Mvc_Cover_t, pList *Mvc_List_t) { pCover.lCubes = *pList;        }

/**Function*************************************************************

  Synopsis    []

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Mvc_CoverIsEmpty(pCover *Mvc_Cover_t) int32 {
    return Mvc_CoverReadCubeNum(pCover) == 0;
}

/**Function*************************************************************

  Synopsis    []

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Mvc_CoverIsTautology(pCover *Mvc_Cover_t) int32 {
    var pCube *Mvc_Cube_t;
    var iBit, Value int32;

    if Mvc_CoverReadCubeNum(pCover) != 1 {
        return 0; }

    pCube = Mvc_CoverReadCubeHead( pCover );
    // inlining of macro Mvc_CubeForEachBit
    // inlining of macro Mvc_CubeBitValue
    // inlining of macro Mvc_CubeWhichWord
    // inlining of macro Mvc_CubeWhichBit
    for  iBit = 0;          iBit < pCover.nBits && ((Value = (((pCube).pData[((iBit) >> BITS_PER_WORD_LOG)] &   (((Mvc_CubeWord_t)1)<<(((iBit) &  BITS_PER_WORD_MINUS)))) > 0)), 1);          iBit++  {
        if Value == 0 {
            return 0; }
    }
    return 1;
}

/**Function*************************************************************

  Synopsis    [Returns 1 if the cover is a binary buffer.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Mvc_CoverIsBinaryBuffer(pCover *Mvc_Cover_t) int32 {
    var pCube *Mvc_Cube_t;
    if pCover.nBits != 2 {
        return 0; }
    if Mvc_CoverReadCubeNum(pCover) != 1 {
        return 0; }
    pCube = pCover.lCubes.pHead;
    // inlining of macro Mvc_CubeBitValue
    // inlining of macro Mvc_CubeWhichWord
    // inlining of macro Mvc_CubeWhichBit
    if  (((pCube).pData[((0) >> BITS_PER_WORD_LOG)] &   (((Mvc_CubeWord_t)1)<<(((0) &  BITS_PER_WORD_MINUS)))) > 0) == 0 && (((pCube).pData[((1) >> BITS_PER_WORD_LOG)] &   (((Mvc_CubeWord_t)1)<<(((1) &  BITS_PER_WORD_MINUS)))) > 0) == 1 {
        return 1; }
    return 0;
}


/**Function*************************************************************

  Synopsis    []

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Mvc_CoverMakeEmpty(pCover *Mvc_Cover_t) {
    var pCube, pCube2 *Mvc_Cube_t;
    // inlining of macro Mvc_CoverForEachCubeSafe
    for  pCube = (pCover).lCubes.pHead, pCube2 = (pCube? pCube.pNext: nil);          (pCube != nil);          pCube = pCube2, pCube2 = (pCube? pCube.pNext: nil)  {
        Mvc_CubeFree( pCover, pCube );
    }
    pCover.lCubes.nItems = 0;
    pCover.lCubes.pHead = nil;
    pCover.lCubes.pTail = nil;
}

/**Function*************************************************************

  Synopsis    []

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Mvc_CoverMakeTautology(pCover *Mvc_Cover_t) {
    var pCubeNew *Mvc_Cube_t;
    Mvc_CoverMakeEmpty( pCover );
    pCubeNew = Mvc_CubeAlloc( pCover );
    // inlining of macro Mvc_CubeBitFill
    // inlining of macro Mvc_Cube1Words
    // inlining of macro Mvc_Cube1BitFill
    // inlining of macro Mvc_Cube2Words
    // inlining of macro Mvc_Cube2BitFill
    if  (pCubeNew).iLast == 0      { (pCubeNew).pData[0] = (BITS_FULL >> (pCubeNew).nUnused);; }    else if  (pCubeNew).iLast == 1 { ((pCubeNew).pData[0] =  BITS_FULL);
// inlining of macro Mvc_CubeNBitFill
((pCubeNew).pData[1] = (BITS_FULL >> (pCubeNew).nUnused)); ; }    else                             { {    var _i_ int32;    (pCubeNew).pData[(pCubeNew).iLast] = (BITS_FULL >> (pCubeNew).nUnused);    for  _i_ = (pCubeNew).iLast - 1; _i_ >= 0; _i_--  {        (pCubeNew).pData[_i_] =  BITS_FULL; }}; };
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
func Mvc_CoverCreateEmpty(pCover *Mvc_Cover_t) *Mvc_Cover_t {
    var pCoverNew *Mvc_Cover_t;
    pCoverNew = Mvc_CoverAlloc( pCover.pMem, pCover.nBits );
    return pCoverNew;
}

/**Function*************************************************************

  Synopsis    []

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Mvc_CoverCreateTautology(pCover *Mvc_Cover_t) *Mvc_Cover_t {
    var pCubeNew *Mvc_Cube_t;
    var pCoverNew *Mvc_Cover_t;
    pCoverNew = Mvc_CoverAlloc( pCover.pMem, pCover.nBits );
    pCubeNew = Mvc_CubeAlloc( pCoverNew );
    // inlining of macro Mvc_CubeBitFill
    // inlining of macro Mvc_Cube1Words
    // inlining of macro Mvc_Cube1BitFill
    // inlining of macro Mvc_Cube2Words
    // inlining of macro Mvc_Cube2BitFill
    if  (pCubeNew).iLast == 0      { (pCubeNew).pData[0] = (BITS_FULL >> (pCubeNew).nUnused);; }    else if  (pCubeNew).iLast == 1 { ((pCubeNew).pData[0] =  BITS_FULL);
// inlining of macro Mvc_CubeNBitFill
((pCubeNew).pData[1] = (BITS_FULL >> (pCubeNew).nUnused)); ; }    else                             { {    var _i_ int32;    (pCubeNew).pData[(pCubeNew).iLast] = (BITS_FULL >> (pCubeNew).nUnused);    for  _i_ = (pCubeNew).iLast - 1; _i_ >= 0; _i_--  {        (pCubeNew).pData[_i_] =  BITS_FULL; }}; };
    // inlining of macro Mvc_CoverAddCubeTail
    // inlining of macro Mvc_ListAddCubeTail
    // inlining of macro Mvc_CubeSetNext
    {    var pList *Mvc_List_t  = &pCoverNew.lCubes;    {        if pList.pHead == nil {            pList.pHead = pCubeNew; }        else {            ((pList.pTail).pNext = (pCubeNew)); }
        // inlining of macro Mvc_CubeSetNext
        pList.pTail    = pCubeNew;        ((pCubeNew).pNext = (nil));        pList.nItems++;    };};
    return pCoverNew;
}


////////////////////////////////////////////////////////////////////////
///                       END OF FILE                                ///
////////////////////////////////////////////////////////////////////////


//ABC_NAMESPACE_IMPL_END

