package gia

import "fmt"

/**CFile****************************************************************

  FileName    [giaFanout.c]

  SystemName  [ABC: Logic synthesis and verification system.]

  PackageName [Scalable AIG package.]

  Synopsis    []

  Author      [Alan Mishchenko]
  
  Affiliation [UC Berkeley]

  Date        [Ver. 1.0. Started - June 20, 2005.]

  Revision    [$Id: giaFanout.c,v 1.00 2005/06/20 00:00:00 alanmi Exp $]

***********************************************************************/

//#include "gia.h"

//ABC_NAMESPACE_IMPL_START


////////////////////////////////////////////////////////////////////////
///                        DECLARATIONS                              ///
////////////////////////////////////////////////////////////////////////

// 0: first iFan
// 1: prev iFan0
// 2: prev iFan1
// 3: next iFan0
// 4: next iFan1

 




// these two procedures are only here for the use inside the iterator



// iterator over the fanouts
// Original macro:
// #define Gia_ObjForEachFanout( p, pObj, pFanout, iFan, i )                       \
//     for ( assert(p->pFanData), i = 0; (i < (int)(pObj)->nRefs) &&               \
//           (((iFan) = i? Gia_ObjFanoutNext(p, iFan) : Gia_ObjFanout0Int(p, Gia_ObjId(p, pObj))), 1) && \
//           (((pFanout) = Gia_ManObj(p, iFan>>1)), 1); i++ )


////////////////////////////////////////////////////////////////////////
///                     FUNCTION DEFINITIONS                         ///
////////////////////////////////////////////////////////////////////////

/**Function*************************************************************

  Synopsis    [Create fanout for all objects in the manager.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/


/**Function*************************************************************

  Synopsis    [Deletes fanout for all objects in the manager.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/


/**Function*************************************************************

  Synopsis    [Adds fanout (pFanout) of node (pObj).]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/


/**Function*************************************************************

  Synopsis    [Removes fanout (pFanout) of node (pObj).]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/





/**Function*************************************************************

  Synopsis    [Compute the map of all edges.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/


/**Function*************************************************************

  Synopsis    [Allocates static fanout.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/



/**Function*************************************************************

  Synopsis    [Compute the map of all edges.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Gia_ManStartMappingFanoutMap(p *Gia_Man_t, vFanoutNums *Vec_Int_t) *Vec_Int_t {
    var pObj *Gia_Obj_t;
    var i int32;
    var iOffset int32 = Gia_ManObjNum(p);
    var vEdgeMap *Vec_Int_t  = Vec_IntAlloc( 2 * iOffset );
    Vec_IntFill( vEdgeMap, iOffset, 0 );
    // inlining of macro Gia_ManForEachObj
    for  i = 0; (i < p.nObjs) && ((pObj) = Gia_ManObj(p, i)); i++  {
        if Vec_IntEntry(vFanoutNums, i) == 0 {
            continue; }
        Vec_IntWriteEntry( vEdgeMap, i, iOffset );
        iOffset += Vec_IntEntry( vFanoutNums, i );
        Vec_IntFillExtra( vEdgeMap, iOffset, 0 );
    }
    //printf( "Fanout map is %.2fx larger than AIG manager.\n", 1.0*Vec_IntSize(vEdgeMap)/Gia_ManObjNum(p) );
    return vEdgeMap;
}

/**Function*************************************************************

  Synopsis    [Allocates static fanout.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Gia_ManStaticMappingFanoutStart(p *Gia_Man_t) {
    var vCounts *Vec_Int_t;
    var pRefsOld *int32;
    var pObj, pFanin *Gia_Obj_t;
    var i, k, iFan, iFanout int32;
    assert(p.vFanoutNums == nil, "p.vFanoutNums == nil");
    assert(p.vFanout == nil, "p.vFanout == nil");
    // recompute reference counters
    pRefsOld = p.pLutRefs; p.pLutRefs = nil;
    Gia_ManSetLutRefs(p);
    p.vFanoutNums = Vec_IntAllocArray( p.pLutRefs, Gia_ManObjNum(p) );
    p.pLutRefs = pRefsOld;
    // start the fanout maps
    p.vFanout = Gia_ManStartMappingFanoutMap( p, p.vFanoutNums );
    // incrementally add fanouts
    vCounts = Vec_IntStart( Gia_ManObjNum(p) );
    // inlining of macro Gia_ManForEachLut
    for  i = 1; i < Gia_ManObjNum(p); i++  { if !Gia_ObjIsLut(p, i) {} else {
        pObj = Gia_ManObj( p, i );
        // inlining of macro Gia_LutForEachFanin
        for  k = 0; k < Gia_ObjLutSize(p,i) && ((iFan = Gia_ObjLutFanins(p,i)[k]),1); k++  {
            pFanin = Gia_ManObj( p, iFan );
            iFanout = Vec_IntEntry( vCounts, iFan );
            Gia_ObjSetFanout( p, pFanin, iFanout, pObj );
            Vec_IntAddToEntry( vCounts, iFan, 1 );
        }
    }
    }
    // inlining of macro Gia_ManForEachCo
    for  i = 0; (i < Vec_IntSize(p.vCos)) && ((pObj) = Gia_ManCo(p, i)); i++  {
        iFan = Gia_ObjFaninId0p(p, pObj);
        pFanin = Gia_ManObj( p, iFan );
        iFanout = Vec_IntEntry( vCounts, iFan );
        Gia_ObjSetFanout( p, pFanin, iFanout, pObj );
        Vec_IntAddToEntry( vCounts, iFan, 1 );
    }
    // double-check the current number of fanouts added
    // inlining of macro Gia_ManForEachObj
    for  i = 0; (i < p.nObjs) && ((pObj) = Gia_ManObj(p, i)); i++  {
        assert(Vec_IntEntry(vCounts, i) == Gia_ObjFanoutNum(p, pObj), "Vec_IntEntry(vCounts, i) == Gia_ObjFanoutNum(p, pObj)");
    }
    Vec_IntFree( vCounts );
}

/**Function*************************************************************

  Synopsis    [Deallocates static fanout.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/



/**Function*************************************************************

  Synopsis    [Tests static fanout.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Gia_ManStaticFanoutTest(p *Gia_Man_t) {
    var pObj, pFanout *Gia_Obj_t;
    var i, k int32;
    Gia_ManStaticFanoutStart( p );
    // inlining of macro Gia_ManForEachObj
    for  i = 0; (i < p.nObjs) && ((pObj) = Gia_ManObj(p, i)); i++  {
        Gia_ObjPrint( p, pObj );
        fmt.Printf("   Fanouts : ");
        // inlining of macro Gia_ObjForEachFanoutStatic
        for  k = 0; (k < Gia_ObjFanoutNum(p, pObj))   && (((pFanout) = Gia_ObjFanout(p, pObj, k)), 1); k++  {
            fmt.Printf("%5d ", Gia_ObjId(p, pFanout));
        }
        fmt.Printf("\n");
    }
    Gia_ManStaticFanoutStop( p );
}

////////////////////////////////////////////////////////////////////////
///                       END OF FILE                                ///
////////////////////////////////////////////////////////////////////////


//ABC_NAMESPACE_IMPL_END

