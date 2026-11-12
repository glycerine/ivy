package tim

/**CFile****************************************************************

  FileName    [timBox.c]

  SystemName  [ABC: Logic synthesis and verification system.]

  PackageName [Hierarchy/timing manager.]

  Synopsis    [Manipulation of timing boxes.]

  Author      [Alan Mishchenko]
  
  Affiliation [UC Berkeley]

  Date        [Ver. 1.0. Started - April 28, 2007.]

  Revision    [$Id: timBox.c,v 1.00 2007/04/28 00:00:00 alanmi Exp $]

***********************************************************************/

//#include "timInt.h"

//ABC_NAMESPACE_IMPL_START

////////////////////////////////////////////////////////////////////////
///                        DECLARATIONS                              ///
////////////////////////////////////////////////////////////////////////

////////////////////////////////////////////////////////////////////////
///                     FUNCTION DEFINITIONS                         ///
////////////////////////////////////////////////////////////////////////

/**Function*************************************************************

  Synopsis    [Creates the new timing box.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Tim_ManCreateBox(p *Tim_Man_t, firstIn int32, nIns int32, firstOut int32, nOuts int32, iDelayTable int32, fBlack int32) {
    var pBox *Tim_Box_t;
    var i int32;
    if p.vBoxes == nil {
        p.vBoxes = Vec_PtrAlloc( 100 ); }
    pBox = (Tim_Box_t *)Mem_FlexEntryFetch( p.pMemObj, sizeof(Tim_Box_t) + sizeof(int) * (nIns+nOuts) );
    memset( pBox, 0, sizeof(Tim_Box_t) );
    pBox.iBox = Vec_PtrSize( p.vBoxes );
    Vec_PtrPush( p.vBoxes, pBox );
    pBox.iDelayTable = iDelayTable;
    pBox.nInputs  = nIns;
    pBox.nOutputs = nOuts;
    pBox.fBlack = fBlack;
    for  i = 0; i < nIns; i++  {
        assert(firstIn+i < p.nCos, "firstIn+i < p.nCos");
        pBox.Inouts[i] = firstIn+i;
        p.pCos[firstIn+i].iObj2Box = pBox.iBox;
        p.pCos[firstIn+i].iObj2Num = i;
    }
    for  i = 0; i < nOuts; i++  {
        assert(firstOut+i < p.nCis, "firstOut+i < p.nCis");
        pBox.Inouts[nIns+i] = firstOut+i;
        p.pCis[firstOut+i].iObj2Box = pBox.iBox;
        p.pCis[firstOut+i].iObj2Num = i;
    }
//    if ( pBox->iBox < 20 )
//        printf( "%4d  %4d  %4d  %4d  \n", firstIn, nIns, firstOut, nOuts );
}

/**Function*************************************************************

  Synopsis    [Returns the box number for the given input.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Tim_ManBoxForCi(p *Tim_Man_t, iCi int32) int32 {
    if iCi >= p.nCis {
        return -1; }
    return p.pCis[iCi].iObj2Box;
}

/**Function*************************************************************

  Synopsis    [Returns the box number for the given output.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Tim_ManBoxForCo(p *Tim_Man_t, iCo int32) int32 {
    if iCo >= p.nCos {
        return -1; }
    return p.pCos[iCo].iObj2Box;
}

/**Function*************************************************************

  Synopsis    [Returns the first input of the box.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Tim_ManBoxInputFirst(p *Tim_Man_t, iBox int32) int32 {
    return Tim_ManBox(p, iBox).Inouts[0];
}

/**Function*************************************************************

  Synopsis    [Returns the last input of the box.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Tim_ManBoxInputLast(p *Tim_Man_t, iBox int32) int32 {
    return Tim_ManBox(p, iBox).Inouts[0] + Tim_ManBoxInputNum(p, iBox) - 1;
}

/**Function*************************************************************

  Synopsis    [Returns the first output of the box.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Tim_ManBoxOutputFirst(p *Tim_Man_t, iBox int32) int32 {
    return Tim_ManBox(p, iBox).Inouts[Tim_ManBox(p, iBox).nInputs];
}

/**Function*************************************************************

  Synopsis    [Returns the last output of the box.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Tim_ManBoxOutputLast(p *Tim_Man_t, iBox int32) int32 {
    return Tim_ManBox(p, iBox).Inouts[Tim_ManBox(p, iBox).nInputs] + Tim_ManBoxOutputNum(p, iBox) - 1;
}

/**Function*************************************************************

  Synopsis    [Returns the number of box inputs.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Tim_ManBoxInputNum(p *Tim_Man_t, iBox int32) int32 {
    return Tim_ManBox(p, iBox).nInputs;
}

/**Function*************************************************************

  Synopsis    [Returns the number of box outputs.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Tim_ManBoxOutputNum(p *Tim_Man_t, iBox int32) int32 {
    return Tim_ManBox(p, iBox).nOutputs;
}

/**Function*************************************************************

  Synopsis    [Return the delay table id.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Tim_ManBoxDelayTableId(p *Tim_Man_t, iBox int32) int32 {
    return Tim_ManBox(p, iBox).iDelayTable;
}

/**Function*************************************************************

  Synopsis    [Return the delay table.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Tim_ManBoxDelayTable(p *Tim_Man_t, iBox int32) *float32 {
    var pTable *float32;
    var pBox *Tim_Box_t  = Tim_ManBox(p, iBox);
    if pBox.iDelayTable < 0 {
        return nil; }
    pTable = (float *)Vec_PtrEntry( p.vDelayTables, pBox.iDelayTable );
    assert(int32(pTable[1]) == pBox.nInputs, "int32(pTable[1]) == pBox.nInputs");
    assert(int32(pTable[2]) == pBox.nOutputs, "int32(pTable[2]) == pBox.nOutputs");
    return pTable;
}

/**Function*************************************************************

  Synopsis    [Return 1 if the box is black.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Tim_ManBoxIsBlack(p *Tim_Man_t, iBox int32) int32 {
    return Tim_ManBox(p, iBox).fBlack;
}


/**Function*************************************************************

  Synopsis    [Returns the copy of the box.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Tim_ManBoxCopy(p *Tim_Man_t, iBox int32) int32 {
    return Tim_ManBox(p, iBox).iCopy;
}

/**Function*************************************************************

  Synopsis    [Sets the copy of the box.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Tim_ManBoxSetCopy(p *Tim_Man_t, iBox int32, iCopy int32) {
    Tim_ManBox(p, iBox).iCopy = iCopy;
}

/**Function*************************************************************

  Synopsis    []

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Tim_ManBoxFindFromCiNum(p *Tim_Man_t, iCiNum int32) int32 {
    var pBox *Tim_Box_t;
    var i int32;
    assert(iCiNum >= 0 && iCiNum < Tim_ManCiNum(p), "iCiNum >= 0 && iCiNum < Tim_ManCiNum(p)");
    if iCiNum < Tim_ManPiNum(p) {
        return -1; }
    // inlining of macro Tim_ManForEachBox
    // inlining of macro Vec_PtrForEachEntry
    for  i = 0; (i < Vec_PtrSize(p.vBoxes)) && (((pBox) = (Tim_Box_t *)Vec_PtrEntry(p.vBoxes, i)), 1); i++  {
        if iCiNum < Tim_ManBoxOutputFirst(p, i) {
            return i - 1; }
    }
    return -2;
}

////////////////////////////////////////////////////////////////////////
///                       END OF FILE                                ///
////////////////////////////////////////////////////////////////////////


//ABC_NAMESPACE_IMPL_END

