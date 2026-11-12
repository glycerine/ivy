package tim

import "fmt"

/**CFile****************************************************************

  FileName    [timMan.c]

  SystemName  [ABC: Logic synthesis and verification system.]

  PackageName [Hierarchy/timing manager.]

  Synopsis    [Manipulation of manager data-structure.]

  Author      [Alan Mishchenko]
  
  Affiliation [UC Berkeley]

  Date        [Ver. 1.0. Started - April 28, 2007.]

  Revision    [$Id: timMan.c,v 1.00 2007/04/28 00:00:00 alanmi Exp $]

***********************************************************************/

//#include "timInt.h"
//#include "map/if/if.h"

//ABC_NAMESPACE_IMPL_START

////////////////////////////////////////////////////////////////////////
///                        DECLARATIONS                              ///
////////////////////////////////////////////////////////////////////////

////////////////////////////////////////////////////////////////////////
///                     FUNCTION DEFINITIONS                         ///
////////////////////////////////////////////////////////////////////////

/**Function*************************************************************

  Synopsis    [Starts the timing manager.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Tim_ManStart(nCis int32, nCos int32) *Tim_Man_t {
    var p *Tim_Man_t;
    var pObj *Tim_Obj_t;
    var i int32;
    // inlining of macro ABC_ALLOC
    p = ((Tim_Man_t *) malloc(sizeof(Tim_Man_t) * (size_t)(1)));
    memset( p, 0, sizeof(Tim_Man_t) );
    p.pMemObj = Mem_FlexStart();
    p.nCis = nCis;
    p.nCos = nCos;
    // inlining of macro ABC_ALLOC
    p.pCis = ((Tim_Obj_t *) malloc(sizeof(Tim_Obj_t) * (size_t)(nCis)));
    memset( p.pCis, 0, sizeof(Tim_Obj_t) * nCis );
    // inlining of macro ABC_ALLOC
    p.pCos = ((Tim_Obj_t *) malloc(sizeof(Tim_Obj_t) * (size_t)(nCos)));
    memset( p.pCos, 0, sizeof(Tim_Obj_t) * nCos );
    // inlining of macro Tim_ManForEachCi
    for  i = 0; (i < (p).nCis) && ((pObj) = (p).pCis + i); i++  {
        pObj.Id = i;
        pObj.iObj2Box = pObj.iObj2Num = -1;
        pObj.timeReq = TIM_ETERNITY;
    }
    // inlining of macro Tim_ManForEachCo
    for  i = 0; (i < (p).nCos) && ((pObj) = (p).pCos + i); i++  {
        pObj.Id = i;
        pObj.iObj2Box = pObj.iObj2Num = -1;
        pObj.timeReq = TIM_ETERNITY;
    }
    p.fUseTravId = 1;
    return p;
}

/**Function*************************************************************

  Synopsis    [Duplicates the timing manager.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Tim_ManDup(p *Tim_Man_t, fUnitDelay int32) *Tim_Man_t {
    var pNew *Tim_Man_t;
    var pBox *Tim_Box_t;
    var pObj *Tim_Obj_t;
    var pDelayTable, pDelayTableNew *float32;
    var i, k, nInputs, nOutputs int32;
    // clear traversal IDs
    // inlining of macro Tim_ManForEachCi
    for  i = 0; (i < (p).nCis) && ((pObj) = (p).pCis + i); i++  { 
        pObj.TravId = 0;
    }          
    // inlining of macro Tim_ManForEachCo
    for  i = 0; (i < (p).nCos) && ((pObj) = (p).pCos + i); i++  { 
        pObj.TravId = 0;
    }          
    // create new manager
    pNew = Tim_ManStart( p.nCis, p.nCos );
    // copy box connectivity information
    memcpy( pNew.pCis, p.pCis, sizeof(Tim_Obj_t) * p.nCis ); 
    memcpy( pNew.pCos, p.pCos, sizeof(Tim_Obj_t) * p.nCos ); 
    if fUnitDelay != 0 {
        // discretize PI arrival times
//        Tim_ManForEachPi( pNew, pObj, k )
//          pObj->timeArr = (int)pObj->timeArr;
        // discretize PO required times
//        Tim_ManForEachPo( pNew, pObj, k )
//          pObj->timeReq = 1 + (int)pObj->timeReq;
        // clear PI arrival and PO required
        Tim_ManInitPiArrivalAll( p, 0.0 );
        Tim_ManInitPoRequiredAll( p, float32(TIM_ETERNITY) );
    }
    // duplicate delay tables
    if Tim_ManDelayTableNum(p) > 0 {
        pNew.vDelayTables = Vec_PtrStart( Vec_PtrSize(p.vDelayTables) );
        // inlining of macro Tim_ManForEachTable
        // inlining of macro Vec_PtrForEachEntry
        for  i = 0; (i < Vec_PtrSize(p.vDelayTables)) && (((pDelayTable) = (float *)Vec_PtrEntry(p.vDelayTables, i)), 1); i++  {
            if pDelayTable == nil {
                continue; }
            assert(i == int32(pDelayTable[0]), "i == int32(pDelayTable[0])");
            nInputs   = int32(pDelayTable[1]);
            nOutputs  = int32(pDelayTable[2]);
            // inlining of macro ABC_ALLOC
            pDelayTableNew = ((float *) malloc(sizeof(float) * (size_t)(3 + nInputs * nOutputs)));
            pDelayTableNew[0] = int32(pDelayTable[0]);
            pDelayTableNew[1] = int32(pDelayTable[1]);
            pDelayTableNew[2] = int32(pDelayTable[2]);
            for  k = 0; k < nInputs * nOutputs; k++  {
                if pDelayTable[3+k] == -ABC_INFINITY {
                    pDelayTableNew[3+k] = -ABC_INFINITY;
                } else {
                    pDelayTableNew[3+k] = fUnitDelay ? float32(fUnitDelay) : pDelayTable[3+k]; }
            }
//            assert( (int)pDelayTableNew[0] == Vec_PtrSize(pNew->vDelayTables) );
            assert(Vec_PtrEntry(pNew.vDelayTables, i) == nil, "Vec_PtrEntry(pNew.vDelayTables, i) == nil");
            Vec_PtrWriteEntry( pNew.vDelayTables, i, pDelayTableNew );
//printf( "Finished duplicating delay table %d.\n", i );
        }
    }
    // duplicate boxes
    if Tim_ManBoxNum(p) > 0 {
        pNew.vBoxes = Vec_PtrAlloc( Tim_ManBoxNum(p) );
        // inlining of macro Tim_ManForEachBox
        // inlining of macro Vec_PtrForEachEntry
        for  i = 0; (i < Vec_PtrSize(p.vBoxes)) && (((pBox) = (Tim_Box_t *)Vec_PtrEntry(p.vBoxes, i)), 1); i++  {
           Tim_ManCreateBox( pNew, pBox.Inouts[0], pBox.nInputs, 
                pBox.Inouts[pBox.nInputs], pBox.nOutputs, pBox.iDelayTable, pBox.fBlack );
           Tim_ManBoxSetCopy( pNew, i, pBox.iCopy );
        }
    }
    return pNew;
}

/**Function*************************************************************

  Synopsis    [Trims the timing manager.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Tim_ManTrim(p *Tim_Man_t, vBoxPres *Vec_Int_t) *Tim_Man_t {
    var pNew *Tim_Man_t;
    var pBox *Tim_Box_t;
    var pObj *Tim_Obj_t;
    var pDelayTable, pDelayTableNew *float32;
    var i, k, nNewCis, nNewCos, nInputs, nOutputs int32;
    assert(Vec_IntSize(vBoxPres) == Tim_ManBoxNum(p), "Vec_IntSize(vBoxPres) == Tim_ManBoxNum(p)");
    // count the number of CIs and COs in the trimmed manager
    nNewCis = Tim_ManPiNum(p);
    nNewCos = Tim_ManPoNum(p);
    if Tim_ManBoxNum(p) != 0 {
        // inlining of macro Tim_ManForEachBox
        // inlining of macro Vec_PtrForEachEntry
        for  i = 0; (i < Vec_PtrSize(p.vBoxes)) && (((pBox) = (Tim_Box_t *)Vec_PtrEntry(p.vBoxes, i)), 1); i++  {
            if Vec_IntEntry(vBoxPres, i) {
                nNewCis += pBox.nOutputs;
                nNewCos += pBox.nInputs;
            }
        } }
    if nNewCis == Tim_ManCiNum(p) && nNewCos == Tim_ManCoNum(p) {
        return Tim_ManDup( p, 0 ); }
    assert(nNewCis < Tim_ManCiNum(p), "nNewCis < Tim_ManCiNum(p)");
    assert(nNewCos < Tim_ManCoNum(p), "nNewCos < Tim_ManCoNum(p)");
    // clear traversal IDs
    // inlining of macro Tim_ManForEachCi
    for  i = 0; (i < (p).nCis) && ((pObj) = (p).pCis + i); i++  { 
        pObj.TravId = 0;
    }          
    // inlining of macro Tim_ManForEachCo
    for  i = 0; (i < (p).nCos) && ((pObj) = (p).pCos + i); i++  { 
        pObj.TravId = 0;
    }          
    // create new manager
    pNew = Tim_ManStart( nNewCis, nNewCos );
    // copy box connectivity information
    memcpy( pNew.pCis, p.pCis, sizeof(Tim_Obj_t) * Tim_ManPiNum(p) );
    memcpy( pNew.pCos + nNewCos - Tim_ManPoNum(p), 
                p.pCos + Tim_ManCoNum(p) - Tim_ManPoNum(p), 
                sizeof(Tim_Obj_t) * Tim_ManPoNum(p) );
    // duplicate delay tables
    if Tim_ManDelayTableNum(p) > 0 {
        pNew.vDelayTables = Vec_PtrStart( Vec_PtrSize(p.vDelayTables) );
        // inlining of macro Tim_ManForEachTable
        // inlining of macro Vec_PtrForEachEntry
        for  i = 0; (i < Vec_PtrSize(p.vDelayTables)) && (((pDelayTable) = (float *)Vec_PtrEntry(p.vDelayTables, i)), 1); i++  {
            if pDelayTable == nil {
                continue; }
            assert(i == int32(pDelayTable[0]), "i == int32(pDelayTable[0])");
            nInputs   = int32(pDelayTable[1]);
            nOutputs  = int32(pDelayTable[2]);
            // inlining of macro ABC_ALLOC
            pDelayTableNew = ((float *) malloc(sizeof(float) * (size_t)(3 + nInputs * nOutputs)));
            pDelayTableNew[0] = int32(pDelayTable[0]);
            pDelayTableNew[1] = int32(pDelayTable[1]);
            pDelayTableNew[2] = int32(pDelayTable[2]);
            for  k = 0; k < nInputs * nOutputs; k++  {
                pDelayTableNew[3+k] = pDelayTable[3+k];
            }
//            assert( (int)pDelayTableNew[0] == Vec_PtrSize(pNew->vDelayTables) );
            assert(Vec_PtrEntry(pNew.vDelayTables, i) == nil, "Vec_PtrEntry(pNew.vDelayTables, i) == nil");
            Vec_PtrWriteEntry( pNew.vDelayTables, i, pDelayTableNew );
        }
    }
    // duplicate boxes
    if Tim_ManBoxNum(p) > 0 {
        var curPi int32  = Tim_ManPiNum(p);
        var curPo int32;
        pNew.vBoxes = Vec_PtrAlloc( Tim_ManBoxNum(p) );
        // inlining of macro Tim_ManForEachBox
        // inlining of macro Vec_PtrForEachEntry
        for  i = 0; (i < Vec_PtrSize(p.vBoxes)) && (((pBox) = (Tim_Box_t *)Vec_PtrEntry(p.vBoxes, i)), 1); i++  {
            if Vec_IntEntry(vBoxPres, i) {
                Tim_ManCreateBox( pNew, curPo, pBox.nInputs, curPi, pBox.nOutputs, pBox.iDelayTable, pBox.fBlack );
                Tim_ManBoxSetCopy( pNew, Tim_ManBoxNum(pNew) - 1, Tim_ManBoxCopy(p, i) == -1 ? i : Tim_ManBoxCopy(p, i) );
                curPi += pBox.nOutputs;
                curPo += pBox.nInputs;
            }
        }
        curPo += Tim_ManPoNum(p);
        assert(curPi == Tim_ManCiNum(pNew), "curPi == Tim_ManCiNum(pNew)");
        assert(curPo == Tim_ManCoNum(pNew), "curPo == Tim_ManCoNum(pNew)");
    }
    return pNew;
}

/**Function*************************************************************

  Synopsis    [Reduces the timing manager.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Tim_ManReduce(p *Tim_Man_t, vBoxesLeft *Vec_Int_t, nTermsDiff int32) *Tim_Man_t {
    var pNew *Tim_Man_t;
    var pBox *Tim_Box_t;
    var pObj *Tim_Obj_t;
    var pDelayTable, pDelayTableNew *float32;
    var i, k, iBox, nNewCis, nNewCos, nInputs, nOutputs int32;
    var nNewPiNum int32  = Tim_ManPiNum(p) - nTermsDiff;
    var nNewPoNum int32  = Tim_ManPoNum(p) - nTermsDiff;
    assert(Vec_IntSize(vBoxesLeft) <= Tim_ManBoxNum(p), "Vec_IntSize(vBoxesLeft) <= Tim_ManBoxNum(p)");
    // count the number of CIs and COs in the trimmed manager
    nNewCis = nNewPiNum;
    nNewCos = nNewPoNum;
    // inlining of macro Vec_IntForEachEntry
    for  i = 0; (i < Vec_IntSize(vBoxesLeft)) && (((iBox) = Vec_IntEntry(vBoxesLeft, i)), 1); i++  {
        pBox = Tim_ManBox( p, iBox );
        nNewCis += pBox.nOutputs;
        nNewCos += pBox.nInputs;
    }
    assert(nNewCis <= Tim_ManCiNum(p) - nTermsDiff, "nNewCis <= Tim_ManCiNum(p) - nTermsDiff");
    assert(nNewCos <= Tim_ManCoNum(p) - nTermsDiff, "nNewCos <= Tim_ManCoNum(p) - nTermsDiff");
    // clear traversal IDs
    // inlining of macro Tim_ManForEachCi
    for  i = 0; (i < (p).nCis) && ((pObj) = (p).pCis + i); i++  { 
        pObj.TravId = 0;
    }          
    // inlining of macro Tim_ManForEachCo
    for  i = 0; (i < (p).nCos) && ((pObj) = (p).pCos + i); i++  { 
        pObj.TravId = 0;
    }          
    // create new manager
    pNew = Tim_ManStart( nNewCis, nNewCos );
    // copy box connectivity information
    memcpy( pNew.pCis, p.pCis, sizeof(Tim_Obj_t) * nNewPiNum );
    memcpy( pNew.pCos + nNewCos - nNewPoNum, 
                p.pCos + Tim_ManCoNum(p) - Tim_ManPoNum(p), 
                sizeof(Tim_Obj_t) * nNewPoNum );
    // duplicate delay tables
    if Tim_ManDelayTableNum(p) > 0 {
        var fWarning int32;
        pNew.vDelayTables = Vec_PtrStart( Vec_PtrSize(p.vDelayTables) );
        // inlining of macro Tim_ManForEachTable
        // inlining of macro Vec_PtrForEachEntry
        for  i = 0; (i < Vec_PtrSize(p.vDelayTables)) && (((pDelayTable) = (float *)Vec_PtrEntry(p.vDelayTables, i)), 1); i++  {
            if pDelayTable == nil {
                continue; }
            if i != int32(pDelayTable[0]) && fWarning == 0 {
                fmt.Printf("Warning: Mismatch in delay-table number between the manager and the box.\n");
                fWarning = 1;
            }
            //assert( i == (int)pDelayTable[0] );
            nInputs   = int32(pDelayTable[1]);
            nOutputs  = int32(pDelayTable[2]);
            // inlining of macro ABC_ALLOC
            pDelayTableNew = ((float *) malloc(sizeof(float) * (size_t)(3 + nInputs * nOutputs)));
            pDelayTableNew[0] = i;//(int)pDelayTable[0];
            pDelayTableNew[1] = int32(pDelayTable[1]);
            pDelayTableNew[2] = int32(pDelayTable[2]);
            for  k = 0; k < nInputs * nOutputs; k++  {
                pDelayTableNew[3+k] = pDelayTable[3+k];
            }
//            assert( (int)pDelayTableNew[0] == Vec_PtrSize(pNew->vDelayTables) );
            assert(Vec_PtrEntry(pNew.vDelayTables, i) == nil, "Vec_PtrEntry(pNew.vDelayTables, i) == nil");
            Vec_PtrWriteEntry( pNew.vDelayTables, i, pDelayTableNew );
        }
    }
    // duplicate boxes
    if Tim_ManBoxNum(p) > 0 {
        var curPi int32  = nNewPiNum;
        var curPo int32;
        pNew.vBoxes = Vec_PtrAlloc( Tim_ManBoxNum(p) );
        // inlining of macro Vec_IntForEachEntry
        for  i = 0; (i < Vec_IntSize(vBoxesLeft)) && (((iBox) = Vec_IntEntry(vBoxesLeft, i)), 1); i++  {
            pBox = Tim_ManBox( p, iBox );
            Tim_ManCreateBox( pNew, curPo, pBox.nInputs, curPi, pBox.nOutputs, pBox.iDelayTable, pBox.fBlack );
            Tim_ManBoxSetCopy( pNew, Tim_ManBoxNum(pNew) - 1, Tim_ManBoxCopy(p, iBox) == -1 ? iBox : Tim_ManBoxCopy(p, iBox) );
            curPi += pBox.nOutputs;
            curPo += pBox.nInputs;
        }
        curPo += nNewPoNum;
        assert(curPi == Tim_ManCiNum(pNew), "curPi == Tim_ManCiNum(pNew)");
        assert(curPo == Tim_ManCoNum(pNew), "curPo == Tim_ManCoNum(pNew)");
    }
    return pNew;
}

/**Function*************************************************************

  Synopsis    [Aligns two sets of boxes using the copy field.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Tim_ManAlignTwo(pSpec *Tim_Man_t, pImpl *Tim_Man_t) *Vec_Int_t {
    var vBoxPres *Vec_Int_t;
    var pBox *Tim_Box_t;
    var i int32;
    assert(Tim_ManBoxNum(pSpec) > Tim_ManBoxNum(pImpl), "Tim_ManBoxNum(pSpec) > Tim_ManBoxNum(pImpl)");
    // check if boxes of pImpl can be aligned
    // inlining of macro Tim_ManForEachBox
    // inlining of macro Vec_PtrForEachEntry
    for  i = 0; (i < Vec_PtrSize(pImpl.vBoxes)) && (((pBox) = (Tim_Box_t *)Vec_PtrEntry(pImpl.vBoxes, i)), 1); i++  {
        if pBox.iCopy < 0 || pBox.iCopy >= Tim_ManBoxNum(pSpec) { 
            return nil; }
    }
    // map dropped boxes into 1, others into 0
    vBoxPres = Vec_IntStart( Tim_ManBoxNum(pSpec) );
    // inlining of macro Tim_ManForEachBox
    // inlining of macro Vec_PtrForEachEntry
    for  i = 0; (i < Vec_PtrSize(pImpl.vBoxes)) && (((pBox) = (Tim_Box_t *)Vec_PtrEntry(pImpl.vBoxes, i)), 1); i++  {
        assert(!Vec_IntEntry(vBoxPres, pBox.iCopy), "!Vec_IntEntry(vBoxPres, pBox.iCopy)");
        Vec_IntWriteEntry( vBoxPres, pBox.iCopy, 1 );
    }
    return vBoxPres;
}

/**Function*************************************************************

  Synopsis    [Stops the timing manager.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Tim_ManStop(p *Tim_Man_t) {
    Vec_PtrFreeFree( p.vDelayTables );
    Vec_PtrFreeP( &p.vBoxes );
    Mem_FlexStop( p.pMemObj, 0 );
    // inlining of macro ABC_FREE
    ((p.pCis) ? (free((char *) (p.pCis)), (p.pCis) = 0) : 0);
    // inlining of macro ABC_FREE
    ((p.pCos) ? (free((char *) (p.pCos)), (p.pCos) = 0) : 0);
    // inlining of macro ABC_FREE
    ((p) ? (free((char *) (p)), (p) = 0) : 0);
}
func Tim_ManStopP(p **Tim_Man_t) {
    if *p == nil {
        return; }
    Tim_ManStop( *p );
    *p = nil;
}

/**Function*************************************************************

  Synopsis    [Creates manager using hierarchy / box library / delay info.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Tim_ManCreate(p *Tim_Man_t, pLib any, vInArrs *Vec_Flt_t, vOutReqs *Vec_Flt_t) {
    var pLibBox *If_LibBox_t  = (If_LibBox_t *)pLib;
    var pIfBox *If_Box_t;
    var pBox *Tim_Box_t;
    var pObj *Tim_Obj_t;
    var pTable *float32;
    var i, k int32;
    assert(p.vDelayTables == nil, "p.vDelayTables == nil");
    if pLibBox != nil {
        p.vDelayTables = Vec_PtrStart( Vec_PtrSize(pLibBox.vBoxes) );
    } else {
        p.vDelayTables = Vec_PtrAlloc( 100 );
    }
    if p.vBoxes {
    // inlining of macro Tim_ManForEachBox
    // inlining of macro Vec_PtrForEachEntry
    for  i = 0; (i < Vec_PtrSize(p.vBoxes)) && (((pBox) = (Tim_Box_t *)Vec_PtrEntry(p.vBoxes, i)), 1); i++  {
        if pBox.iDelayTable == -1 || pLibBox == nil {
            // create table with constants
            // inlining of macro ABC_ALLOC
            pTable = ((float *) malloc(sizeof(float) * (size_t)(3 + pBox.nInputs * pBox.nOutputs)));
            pTable[0] = pBox.iDelayTable;
            pTable[1] = pBox.nInputs;
            pTable[2] = pBox.nOutputs;
            for  k = 0; k < pBox.nInputs * pBox.nOutputs; k++  {
                pTable[3 + k] = 1.0;
            }
            // save table
            pBox.iDelayTable = Vec_PtrSize(p.vDelayTables);
            Vec_PtrPush( p.vDelayTables, pTable );
            continue;
        }
        assert(pBox.iDelayTable >= 0 && pBox.iDelayTable < Vec_PtrSize(pLibBox.vBoxes), "pBox.iDelayTable >= 0 && pBox.iDelayTable < Vec_PtrSize(pLibBox.vBoxes)");
        pIfBox = (If_Box_t *)Vec_PtrEntry( pLibBox.vBoxes, pBox.iDelayTable );
        assert(pIfBox != nil, "pIfBox != nil");
        assert(pIfBox.nPis == pBox.nInputs, "pIfBox.nPis == pBox.nInputs");
        assert(pIfBox.nPos == pBox.nOutputs, "pIfBox.nPos == pBox.nOutputs");
        pBox.fBlack = pIfBox.fBlack;
        if Vec_PtrEntry( p.vDelayTables, pBox.iDelayTable ) != nil {
            continue; }
        // create table of boxes
        // inlining of macro ABC_ALLOC
        pTable = ((float *) malloc(sizeof(float) * (size_t)(3 + pBox.nInputs * pBox.nOutputs)));
        pTable[0] = pBox.iDelayTable;
        pTable[1] = pBox.nInputs;
        pTable[2] = pBox.nOutputs;
        for  k = 0; k < pBox.nInputs * pBox.nOutputs; k++  {
            pTable[3 + k] = pIfBox.pDelays[k];
        }
        // save table
        Vec_PtrWriteEntry( p.vDelayTables, pBox.iDelayTable, pTable );
    } }
    // create arrival times
    if vInArrs != nil {
        assert(Vec_FltSize(vInArrs) == Tim_ManPiNum(p), "Vec_FltSize(vInArrs) == Tim_ManPiNum(p)");
        // inlining of macro Tim_ManForEachPi
        // inlining of macro Tim_ManForEachCi
        for  i = 0; (i < (p).nCis) && ((pObj) = (p).pCis + i); i++  { if pObj.iObj2Box >= 0 {} else { 
            pObj.timeArr = Vec_FltEntry(vInArrs, i); }
        }

    }
    // create required times
    if vOutReqs != nil {
        k = 0;
        assert(Vec_FltSize(vOutReqs) == Tim_ManPoNum(p), "Vec_FltSize(vOutReqs) == Tim_ManPoNum(p)");
        // inlining of macro Tim_ManForEachPo
        // inlining of macro Tim_ManForEachCo
        for  i = 0; (i < (p).nCos) && ((pObj) = (p).pCos + i); i++  { if pObj.iObj2Box >= 0 {} else { 
            pObj.timeReq = Vec_FltEntry(vOutReqs, k++); }
        }
        assert(k == Tim_ManPoNum(p), "k == Tim_ManPoNum(p)");
    }
}


/**Function*************************************************************

  Synopsis    [Get arrival and required times if they are non-trivial.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Tim_ManGetArrTimes(p *Tim_Man_t) *float32 {
    var pTimes *float32;
    var pObj *Tim_Obj_t;
    var i int32;
    // inlining of macro Tim_ManForEachPi
    // inlining of macro Tim_ManForEachCi
    for  i = 0; (i < (p).nCis) && ((pObj) = (p).pCis + i); i++  { if pObj.iObj2Box >= 0 {} else 
        if pObj.timeArr != 0.0 {
            break; }
    }
    if i == Tim_ManPiNum(p) {
        return nil; }
    // inlining of macro ABC_FALLOC
    pTimes  = ((float *) memset(malloc(sizeof(float) * (size_t)(Tim_ManCiNum(p))), 0xff, sizeof(float) * (size_t)(Tim_ManCiNum(p))));
    // inlining of macro Tim_ManForEachPi
    // inlining of macro Tim_ManForEachCi
    for  i = 0; (i < (p).nCis) && ((pObj) = (p).pCis + i); i++  { if pObj.iObj2Box >= 0 {} else { 
        pTimes[i] = pObj.timeArr; }
    }
    return pTimes;
}
func Tim_ManGetReqTimes(p *Tim_Man_t) *float32 {
    var pTimes *float32;
    var pObj *Tim_Obj_t;
    var i, k int32;
    // inlining of macro Tim_ManForEachPo
    // inlining of macro Tim_ManForEachCo
    for  i = 0; (i < (p).nCos) && ((pObj) = (p).pCos + i); i++  { if pObj.iObj2Box >= 0 {} else 
        if pObj.timeReq != TIM_ETERNITY {
            break; }
    }
    if i == Tim_ManPoNum(p) {
        return nil; }
    // inlining of macro ABC_FALLOC
    pTimes  = ((float *) memset(malloc(sizeof(float) * (size_t)(Tim_ManCoNum(p))), 0xff, sizeof(float) * (size_t)(Tim_ManCoNum(p))));
    // inlining of macro Tim_ManForEachPo
    // inlining of macro Tim_ManForEachCo
    for  i = 0; (i < (p).nCos) && ((pObj) = (p).pCos + i); i++  { if pObj.iObj2Box >= 0 {} else { 
        pTimes[k++] = pObj.timeArr; }
    }
    assert(k == Tim_ManPoNum(p), "k == Tim_ManPoNum(p)");
    return pTimes;
}


/**Function*************************************************************

  Synopsis    [Prints the timing manager.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Tim_ManPrint(p *Tim_Man_t) {
    var pBox *Tim_Box_t;
    var pObj, pPrev *Tim_Obj_t;
    var pTable *float32;
    var i, j, k, TableX, TableY int32;
    if p == nil {
        return; }
    fmt.Printf("TIMING MANAGER:\n");
    fmt.Printf("PI = %d. CI = %d. PO = %d. CO = %d. Box = %d.\n", Tim_ManPiNum(p), Tim_ManCiNum(p), Tim_ManPoNum(p), Tim_ManCoNum(p), Tim_ManBoxNum(p));

    // print CI info
    pPrev = p.pCis;
    // inlining of macro Tim_ManForEachPi
    // inlining of macro Tim_ManForEachCi
    for  i = 0; (i < (p).nCis) && ((pObj) = (p).pCis + i); i++  { if pObj.iObj2Box >= 0 {} else 
        if pPrev.timeArr != pObj.timeArr || pPrev.timeReq != pObj.timeReq {
            break; }
    }
    if i == Tim_ManCiNum(p) {
        fmt.Printf("All PIs     :  arrival = %5.3f  required = %5.3f\n", pPrev.timeArr, pPrev.timeReq);
    } else {
        // inlining of macro Tim_ManForEachPi
        // inlining of macro Tim_ManForEachCi
        for  i = 0; (i < (p).nCis) && ((pObj) = (p).pCis + i); i++  { if pObj.iObj2Box >= 0 {} else { 
            fmt.Printf("PI%5d     :  arrival = %5.3f  required = %5.3f\n", i, pObj.timeArr, pObj.timeReq); }
        } }

    // print CO info
    pPrev = p.pCos;
    // inlining of macro Tim_ManForEachPo
    // inlining of macro Tim_ManForEachCo
    for  i = 0; (i < (p).nCos) && ((pObj) = (p).pCos + i); i++  { if pObj.iObj2Box >= 0 {} else 
        if pPrev.timeArr != pObj.timeArr || pPrev.timeReq != pObj.timeReq {
            break; }
    }
    if i == Tim_ManCoNum(p) {
        fmt.Printf("All POs     :  arrival = %5.3f  required = %5.3f\n", pPrev.timeArr, pPrev.timeReq);
    } else {
        var k int32;
        // inlining of macro Tim_ManForEachPo
        // inlining of macro Tim_ManForEachCo
        for  i = 0; (i < (p).nCos) && ((pObj) = (p).pCos + i); i++  { if pObj.iObj2Box >= 0 {} else { 
            fmt.Printf("PO%5d     :  arrival = %5.3f  required = %5.3f\n", k++, pObj.timeArr, pObj.timeReq); }
        }
    }

    // print box info
    if Tim_ManBoxNum(p) > 0 {
    // inlining of macro Tim_ManForEachBox
    // inlining of macro Vec_PtrForEachEntry
    for  i = 0; (i < Vec_PtrSize(p.vBoxes)) && (((pBox) = (Tim_Box_t *)Vec_PtrEntry(p.vBoxes, i)), 1); i++  {
        fmt.Printf("*** Box %5d :  I =%4d. O =%4d. I1 =%6d. O1 =%6d. Table =%4d\n", i, pBox.nInputs, pBox.nOutputs, Tim_ManBoxInputFirst(p, i), Tim_ManBoxOutputFirst(p, i), pBox.iDelayTable);

        // print box inputs
        pPrev = Tim_ManBoxInput( p, pBox, 0 );
        // inlining of macro Tim_ManBoxForEachInput
        for  k = 0; (k < (pBox).nInputs) && ((pObj) = Tim_ManBoxInput(p, pBox, k)); k++  {
            if pPrev.timeArr != pObj.timeArr || pPrev.timeReq != pObj.timeReq {
                break; }
        }
        if k == Tim_ManBoxInputNum(p, pBox.iBox) {
            fmt.Printf("Box inputs  :  arrival = %5.3f  required = %5.3f\n", pPrev.timeArr, pPrev.timeReq);
        } else {
            // inlining of macro Tim_ManBoxForEachInput
            for  k = 0; (k < (pBox).nInputs) && ((pObj) = Tim_ManBoxInput(p, pBox, k)); k++  {
                fmt.Printf("box-in%4d :  arrival = %5.3f  required = %5.3f\n", k, pObj.timeArr, pObj.timeReq);
            } }

        // print box outputs
        pPrev = Tim_ManBoxOutput( p, pBox, 0 );
        // inlining of macro Tim_ManBoxForEachOutput
        for  k = 0; (k < (pBox).nOutputs) && ((pObj) = Tim_ManBoxOutput(p, pBox, k)); k++  {
            if pPrev.timeArr != pObj.timeArr || pPrev.timeReq != pObj.timeReq {
                break; }
        }
        if k == Tim_ManBoxOutputNum(p, pBox.iBox) {
            fmt.Printf("Box outputs :  arrival = %5.3f  required = %5.3f\n", pPrev.timeArr, pPrev.timeReq);
        } else {
            // inlining of macro Tim_ManBoxForEachOutput
            for  k = 0; (k < (pBox).nOutputs) && ((pObj) = Tim_ManBoxOutput(p, pBox, k)); k++  {
                fmt.Printf("box-out%3d :  arrival = %5.3f  required = %5.3f\n", k, pObj.timeArr, pObj.timeReq);
            } }

        if i > 2 {
            break; }
    } }

    // print delay tables
    if Tim_ManDelayTableNum(p) > 0 {
    // inlining of macro Tim_ManForEachTable
    // inlining of macro Vec_PtrForEachEntry
    for  i = 0; (i < Vec_PtrSize(p.vDelayTables)) && (((pTable) = (float *)Vec_PtrEntry(p.vDelayTables, i)), 1); i++  {
        if pTable == nil {
            continue; }
        fmt.Printf("Delay table %d:\n", i);
        assert(i == int32(pTable[0]), "i == int32(pTable[0])");
        TableX = int32(pTable[1]);
        TableY = int32(pTable[2]);
        for  j = 0; j < TableY; j++, printf( "\n" )  {
            for  k = 0; k < TableX; k++  {
                if pTable[3+j*TableX+k] == -ABC_INFINITY {
                    fmt.Printf("%5s", "-");
                } else {
                    fmt.Printf("%5.0f", pTable[3+j*TableX+k]); }
            }
        }
    } }
    fmt.Printf("\n");
}
func Tim_ManPrintBoxCopy(p *Tim_Man_t) {
    var pBox *Tim_Box_t;
    var i int32;
    if p == nil {
        return; }
    fmt.Printf("TIMING MANAGER:\n");
    fmt.Printf("PI = %d. CI = %d. PO = %d. CO = %d. Box = %d.\n", Tim_ManPiNum(p), Tim_ManCiNum(p), Tim_ManPoNum(p), Tim_ManCoNum(p), Tim_ManBoxNum(p));

    if Tim_ManBoxNum(p) > 0 {
    // inlining of macro Tim_ManForEachBox
    // inlining of macro Vec_PtrForEachEntry
    for  i = 0; (i < Vec_PtrSize(p.vBoxes)) && (((pBox) = (Tim_Box_t *)Vec_PtrEntry(p.vBoxes, i)), 1); i++  {
        fmt.Printf("%d ", pBox.iCopy);
    } }
    fmt.Printf("\n");
}

/**Function*************************************************************

  Synopsis    [Prints statistics of the timing manager.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Tim_ManPrintStats(p *Tim_Man_t, nAnd2Delay int32) {
    var pBox *Tim_Box_t;
    var vCounts *Vec_Int_t;
    var vBoxes *Vec_Ptr_t;
    var i, Count, IdMax int32;
    if p == nil {
        return; }
    Abc_Print( 1, "Hierarchy      :  " );
    fmt.Printf("PI/CI = %d/%d   PO/CO = %d/%d   Box = %d   ", Tim_ManPiNum(p), Tim_ManCiNum(p), Tim_ManPoNum(p), Tim_ManCoNum(p), Tim_ManBoxNum(p));
    if nAnd2Delay != 0 {
        fmt.Printf("delay(AND2) = %d", nAnd2Delay); }
    fmt.Printf("\n");
    if Tim_ManBoxNum(p) == 0 {
        return; }
    IdMax = 0;
    // inlining of macro Tim_ManForEachBox
    // inlining of macro Vec_PtrForEachEntry
    for  i = 0; (i < Vec_PtrSize(p.vBoxes)) && (((pBox) = (Tim_Box_t *)Vec_PtrEntry(p.vBoxes, i)), 1); i++  {
        IdMax = Abc_MaxInt( IdMax, pBox.iDelayTable );
    }
    vCounts = Vec_IntStart( IdMax+1 );
    vBoxes  = Vec_PtrStart( IdMax+1 );
    // inlining of macro Tim_ManForEachBox
    // inlining of macro Vec_PtrForEachEntry
    for  i = 0; (i < Vec_PtrSize(p.vBoxes)) && (((pBox) = (Tim_Box_t *)Vec_PtrEntry(p.vBoxes, i)), 1); i++  {
        Vec_IntAddToEntry( vCounts, pBox.iDelayTable, 1 );
        Vec_PtrWriteEntry( vBoxes, pBox.iDelayTable, pBox );
    }
    // print statistics about boxes
    // inlining of macro Vec_IntForEachEntry
    for  i = 0; (i < Vec_IntSize(vCounts)) && (((Count) = Vec_IntEntry(vCounts, i)), 1); i++  {
        if Count == 0 { continue; }
        pBox = (Tim_Box_t *)Vec_PtrEntry( vBoxes, i );
        fmt.Printf("    Box %4d      ", i);
        fmt.Printf("Num = %4d   ", Count);
        fmt.Printf("Ins = %4d   ", pBox.nInputs);
        fmt.Printf("Outs = %4d", pBox.nOutputs);
        fmt.Printf("\n");
    }
    Vec_IntFree( vCounts );
    Vec_PtrFree( vBoxes );
}

/**Function*************************************************************

  Synopsis    [Read parameters.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Tim_ManCiNum(p *Tim_Man_t) int32 {
    return p.nCis;
}
func Tim_ManCoNum(p *Tim_Man_t) int32 {
    return p.nCos;
}
func Tim_ManPiNum(p *Tim_Man_t) int32 {
    if Tim_ManBoxNum(p) == 0 {
        return Tim_ManCiNum(p); }
    return Tim_ManBoxOutputFirst(p, 0);
}
func Tim_ManPoNum(p *Tim_Man_t) int32 {
    var iLastBoxId int32;
    if Tim_ManBoxNum(p) == 0 {
        return Tim_ManCoNum(p); }
    iLastBoxId = Tim_ManBoxNum(p) - 1;
    return Tim_ManCoNum(p) - (Tim_ManBoxInputFirst(p, iLastBoxId) + Tim_ManBoxInputNum(p, iLastBoxId));
}
func Tim_ManBoxNum(p *Tim_Man_t) int32 {
    if p.vBoxes {
        return Vec_PtrSize(p.vBoxes);
    } else {
        return 0;
    }
}
func Tim_ManBlackBoxNum(p *Tim_Man_t) int32 {
    var pBox *Tim_Box_t;
    var i, Counter int32;
    if Tim_ManBoxNum(p) != 0 {
        // inlining of macro Tim_ManForEachBox
        // inlining of macro Vec_PtrForEachEntry
        for  i = 0; (i < Vec_PtrSize(p.vBoxes)) && (((pBox) = (Tim_Box_t *)Vec_PtrEntry(p.vBoxes, i)), 1); i++  {
            Counter += pBox.fBlack;
        } }
    return Counter;
}
func Tim_ManBlackBoxIoNum(p *Tim_Man_t, pnBbIns *int32, pnBbOuts *int32) {
    var pBox *Tim_Box_t;
    var i int32;
    *pnBbIns = *pnBbOuts = 0;
    if Tim_ManBoxNum(p) != 0 {
        // inlining of macro Tim_ManForEachBox
        // inlining of macro Vec_PtrForEachEntry
        for  i = 0; (i < Vec_PtrSize(p.vBoxes)) && (((pBox) = (Tim_Box_t *)Vec_PtrEntry(p.vBoxes, i)), 1); i++  {
            if !pBox.fBlack {//&& pBox->nInputs <= 6 )
                continue; }
            *pnBbIns  += Tim_ManBoxInputNum( p, i );
            *pnBbOuts += Tim_ManBoxOutputNum( p, i );
        } }
}
func Tim_ManDelayTableNum(p *Tim_Man_t) int32 {
    if p.vDelayTables {
        return Vec_PtrSize(p.vDelayTables);
    } else {
        return 0;
    }
}

/**Function*************************************************************

  Synopsis    [Sets the vector of timing tables associated with the manager.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Tim_ManSetDelayTables(p *Tim_Man_t, vDelayTables *Vec_Ptr_t) {
    assert(p.vDelayTables == nil, "p.vDelayTables == nil");
    p.vDelayTables = vDelayTables;
}

/**Function*************************************************************

  Synopsis    [Disables the use of the traversal ID.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Tim_ManTravIdDisable(p *Tim_Man_t) {
    p.fUseTravId = 0;
}

/**Function*************************************************************

  Synopsis    [Enables the use of the traversal ID.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Tim_ManTravIdEnable(p *Tim_Man_t) {
    p.fUseTravId = 1;
}

////////////////////////////////////////////////////////////////////////
///                       END OF FILE                                ///
////////////////////////////////////////////////////////////////////////


//ABC_NAMESPACE_IMPL_END

