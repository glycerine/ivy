package hop

import (
	"fmt"
	"time"
)

/**CFile****************************************************************

  FileName    [hopMan.c]

  SystemName  [ABC: Logic synthesis and verification system.]

  PackageName [Minimalistic And-Inverter Graph package.]

  Synopsis    [AIG manager.]

  Author      [Alan Mishchenko]
  
  Affiliation [UC Berkeley]

  Date        [Ver. 1.0. Started - May 11, 2006.]

  Revision    [$Id: hopMan.c,v 1.00 2006/05/11 00:00:00 alanmi Exp $]

***********************************************************************/

//#include "hop.h"

//ABC_NAMESPACE_IMPL_START


////////////////////////////////////////////////////////////////////////
///                        DECLARATIONS                              ///
////////////////////////////////////////////////////////////////////////

////////////////////////////////////////////////////////////////////////
///                     FUNCTION DEFINITIONS                         ///
////////////////////////////////////////////////////////////////////////

/**Function*************************************************************

  Synopsis    [Starts the AIG manager.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Hop_ManStart() *Hop_Man_t {
    var p *Hop_Man_t;
    // start the manager
    // inlining of macro ABC_ALLOC
    p = ((Hop_Man_t *) malloc(sizeof(Hop_Man_t) * (size_t)(1)));
    memset( p, 0, sizeof(Hop_Man_t) );
    // perform initializations
    p.nTravIds = 1;
    p.fRefCount = 1;
    p.fCatchExor = 0;
    // allocate arrays for nodes
    p.vPis = Vec_PtrAlloc( 100 );
    p.vPos = Vec_PtrAlloc( 100 );
    // prepare the internal memory manager
    Hop_ManStartMemory( p );
    // create the constant node
    p.pConst1 = Hop_ManFetchMemory( p );
    p.pConst1.Type = AIG_CONST1;
    p.pConst1.fPhase = 1;
    p.nCreated = 1;
    // start the table
//    p->nTableSize = 107;
    p.nTableSize = 10007;
    // inlining of macro ABC_ALLOC
    p.pTable = ((Hop_Obj_t * *) malloc(sizeof(Hop_Obj_t *) * (size_t)(p.nTableSize)));
    memset( p.pTable, 0, sizeof(Hop_Obj_t *) * p.nTableSize );
    return p;
}

/**Function*************************************************************

  Synopsis    [Stops the AIG manager.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Hop_ManStop(p *Hop_Man_t) {
    var pObj *Hop_Obj_t;
    var i int32;
    // make sure the nodes have clean marks
    pObj = Hop_ManConst1(p);
    assert(!pObj.fMarkA && !pObj.fMarkB, "!pObj.fMarkA && !pObj.fMarkB");
    // inlining of macro Hop_ManForEachPi
    // inlining of macro Vec_PtrForEachEntry
    for  i = 0; (i < Vec_PtrSize(p.vPis)) && (((pObj) = (Hop_Obj_t *)Vec_PtrEntry(p.vPis, i)), 1); i++  {
        assert(!pObj.fMarkA && !pObj.fMarkB, "!pObj.fMarkA && !pObj.fMarkB");
    }
    // inlining of macro Hop_ManForEachPo
    // inlining of macro Vec_PtrForEachEntry
    for  i = 0; (i < Vec_PtrSize(p.vPos)) && (((pObj) = (Hop_Obj_t *)Vec_PtrEntry(p.vPos, i)), 1); i++  {
        assert(!pObj.fMarkA && !pObj.fMarkB, "!pObj.fMarkA && !pObj.fMarkB");
    }
    // inlining of macro Hop_ManForEachNode
    for  i = 0; i < p.nTableSize; i++  {                                               if ((pObj) = p.pTable[i]) == nil {} else {
        assert(!pObj.fMarkA && !pObj.fMarkB, "!pObj.fMarkA && !pObj.fMarkB"); }
    }
    // print time
    // inlining of macro ABC_PRT
    if p.time1 { Abc_Print(1, "%s =", ("time1"));
Abc_Print(1, "%9.2f sec\n", 1.0*(float64((p.time1)))/(float64(int64(time.Second))));; }
    // inlining of macro ABC_PRT
    if p.time2 { Abc_Print(1, "%s =", ("time2"));
Abc_Print(1, "%9.2f sec\n", 1.0*(float64((p.time2)))/(float64(int64(time.Second))));; }
//    Hop_TableProfile( p );
    if p.vChunks {  Hop_ManStopMemory( p ); }
    if p.vPis {     Vec_PtrFree( p.vPis ); }
    if p.vPos {     Vec_PtrFree( p.vPos ); }
    if p.vObjs {    Vec_PtrFree( p.vObjs ); }
    // inlining of macro ABC_FREE
    ((p.pTable) ? (free((char *) (p.pTable)), (p.pTable) = 0) : 0);
    // inlining of macro ABC_FREE
    ((p) ? (free((char *) (p)), (p) = 0) : 0);
}

/**Function*************************************************************

  Synopsis    [Returns the number of dangling nodes removed.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Hop_ManCleanup(p *Hop_Man_t) int32 {
    var vObjs *Vec_Ptr_t;
    var pNode *Hop_Obj_t;
    var i, nNodesOld int32;
    assert(p.fRefCount, "p.fRefCount");
    nNodesOld = Hop_ManNodeNum(p);
    // collect roots of dangling nodes
    vObjs = Vec_PtrAlloc( 100 );
    // inlining of macro Hop_ManForEachNode
    for  i = 0; i < p.nTableSize; i++  {                                               if ((pNode) = p.pTable[i]) == nil {} else
        if Hop_ObjRefs(pNode) == 0 {
            Vec_PtrPush( vObjs, pNode ); }
    }
    // recursively remove dangling nodes
    // inlining of macro Vec_PtrForEachEntry
    for  i = 0; (i < Vec_PtrSize(vObjs)) && (((pNode) = (Hop_Obj_t *)Vec_PtrEntry(vObjs, i)), 1); i++  {
        Hop_ObjDelete_rec( p, pNode );
    }
    Vec_PtrFree( vObjs );
    return nNodesOld - Hop_ManNodeNum(p);
}

/**Function*************************************************************

  Synopsis    [Stops the AIG manager.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Hop_ManPrintStats(p *Hop_Man_t) {
    fmt.Printf("PI/PO = %d/%d. ", Hop_ManPiNum(p), Hop_ManPoNum(p));
    fmt.Printf("A = %7d. ", Hop_ManAndNum(p));
    fmt.Printf("X = %5d. ", Hop_ManExorNum(p));
    fmt.Printf("Cre = %7d. ", p.nCreated);
    fmt.Printf("Del = %7d. ", p.nDeleted);
    fmt.Printf("Lev = %3d. ", Hop_ManCountLevels(p));
    fmt.Printf("\n");
}



////////////////////////////////////////////////////////////////////////
///                       END OF FILE                                ///
////////////////////////////////////////////////////////////////////////


//ABC_NAMESPACE_IMPL_END

