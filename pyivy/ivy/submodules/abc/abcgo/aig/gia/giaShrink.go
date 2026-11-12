package gia

import (
	"fmt"
	"time"
)

/**CFile****************************************************************

  FileName    [giaShrink.c]

  SystemName  [ABC: Logic synthesis and verification system.]

  PackageName [Scalable AIG package.]

  Synopsis    [Implementation of mapShrink based on ideas of Niklas Een.]

  Author      [Alan Mishchenko]
  
  Affiliation [UC Berkeley]

  Date        [Ver. 1.0. Started - June 20, 2005.]

  Revision    [$Id: giaShrink.c,v 1.00 2005/06/20 00:00:00 alanmi Exp $]

***********************************************************************/

//#include "gia.h"
//#include "aig/aig/aig.h"
//#include "opt/dar/dar.h"

//ABC_NAMESPACE_IMPL_START


////////////////////////////////////////////////////////////////////////
///                        DECLARATIONS                              ///
////////////////////////////////////////////////////////////////////////

// extern int  Dar_LibEvalBuild( Gia_Man_t * p, Vec_Int_t * vCut, unsigned uTruth, int fKeepLevel, Vec_Int_t * vLeavesBest );


////////////////////////////////////////////////////////////////////////
///                     FUNCTION DEFINITIONS                         ///
////////////////////////////////////////////////////////////////////////

/**Function*************************************************************

  Synopsis    [Performs AIG shrinking using the current mapping.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Gia_ManMapShrink4(p *Gia_Man_t, fKeepLevel int32, fVerbose int32) *Gia_Man_t {
    var vLeaves, vTruth, vVisited, vLeavesBest *Vec_Int_t;
    var pNew, pTemp *Gia_Man_t;
    var pObj, pFanin *Gia_Obj_t;
    var pTruth *uint32;
    var i, k, iFan int32;
    var clk abctime  = Abc_Clock();
//    int ClassCounts[222] = {0};
    var pLutClass *int32;
    var Counter int32;
    assert(Gia_ManHasMapping(p), "Gia_ManHasMapping(p)");
    if Gia_ManLutSizeMax( p ) > 4 {
        fmt.Printf("Resynthesis is not performed when nodes have more than 4 inputs.\n");
        return nil;
    }
    // inlining of macro ABC_CALLOC
    pLutClass   = ((int *) calloc((size_t)(Gia_ManObjNum(p)), sizeof(int)));
    vLeaves     = Vec_IntAlloc( 0 );
    vTruth      = Vec_IntAlloc( (1<<16) );
    vVisited    = Vec_IntAlloc( 0 );
    vLeavesBest = Vec_IntAlloc( 4 );
    // prepare the library
    Dar_LibPrepare( 5 ); 
    // clean the old manager
    Gia_ManCleanTruth( p );
    Gia_ManSetPhase( p );
    Gia_ManFillValue( p );
    Gia_ManConst0(p).Value = 0;
    // start the new manager
    pNew = Gia_ManStart( Gia_ManObjNum(p) );
    pNew.pName = Abc_UtilStrsav( p.pName );
    pNew.pSpec = Abc_UtilStrsav( p.pSpec );
    Gia_ManHashAlloc( pNew );
    Gia_ManCleanLevels( pNew, Gia_ManObjNum(p) );
    // inlining of macro Gia_ManForEachObj1
    for  i = 1; (i < p.nObjs) && ((pObj) = Gia_ManObj(p, i)); i++  {
        if Gia_ObjIsCi(pObj) {
            pObj.Value = Gia_ManAppendCi( pNew );
            if p.vLevels {
                Gia_ObjSetLevel( pNew, Gia_ObjFromLit(pNew, Gia_ObjValue(pObj)), Gia_ObjLevel(p, pObj) ); }
        
        } else if Gia_ObjIsCo(pObj) {
            pObj.Value = Gia_ManAppendCo( pNew, Gia_ObjFanin0Copy(pObj) );
        
        } else if Gia_ObjIsLut(p, i) {
            Counter++;
            // collect leaves of this gate
            Vec_IntClear( vLeaves );
            // inlining of macro Gia_LutForEachFanin
            for  k = 0; k < Gia_ObjLutSize(p,i) && ((iFan = Gia_ObjLutFanins(p,i)[k]),1); k++  {
                Vec_IntPush( vLeaves, iFan );
            }
            for  ; k < 4; k++  {
                Vec_IntPush( vLeaves, 0 );
            }
            //.compute the truth table 
            pTruth = Gia_ManConvertAigToTruth( p, pObj, vLeaves, vTruth, vVisited );
            // change from node IDs to their literals
            // inlining of macro Gia_ManForEachObjVec
            for  k = 0; (k < Vec_IntSize(vLeaves)) && ((pFanin) = Gia_ManObj(p, Vec_IntEntry(vLeaves,k))); k++  {
//                assert( Gia_ObjValue(pFanin) != ~0 ); 
                Vec_IntWriteEntry( vLeaves, k, Gia_ObjValue(pFanin) != ~0 ? Gia_ObjValue(pFanin) : 0 );
            }
            // derive new structre
            if Gia_ManTruthIsConst0(pTruth, Vec_IntSize(vLeaves)) {
                pObj.Value = 0;
            } else if Gia_ManTruthIsConst1(pTruth, Vec_IntSize(vLeaves)) {
                pObj.Value = 1;
            } else {
                pObj.Value = Dar_LibEvalBuild( pNew, vLeaves, 0xffff & *pTruth, fKeepLevel, vLeavesBest );
                pObj.Value = Abc_LitNotCond( pObj.Value, Gia_ObjPhaseRealLit(pNew, pObj.Value) ^ pObj.fPhase );
            }
        }
    }
    // cleanup the AIG
    Gia_ManHashStop( pNew );
    // check the presence of dangling nodes
    if Gia_ManHasDangling(pNew) {
        pNew = Gia_ManCleanup( pTemp = pNew );
        if (fVerbose != 0) && Gia_ManAndNum(pNew) != Gia_ManAndNum(pTemp) {
            fmt.Printf("Gia_ManMapShrink4() node reduction after sweep %6d -> %6d.\n", Gia_ManAndNum(pTemp), Gia_ManAndNum(pNew)); }
        Gia_ManStop( pTemp );
    }
    Gia_ManSetRegNum( pNew, Gia_ManRegNum(p) );
    Vec_IntFree( vLeaves );
    Vec_IntFree( vTruth );
    Vec_IntFree( vVisited );
    Vec_IntFree( vLeavesBest );
    if fVerbose != 0 {
        fmt.Printf("Total gain in AIG nodes = %d.  ", Gia_ManObjNum(p)-Gia_ManObjNum(pNew));
        // inlining of macro ABC_PRT
        Abc_Print(1, "%s =", ("Total runtime"));
Abc_Print(1, "%9.2f sec\n", 1.0*(float64((Abc_Clock() - clk)))/(float64(int64(time.Second))));;
    }
    // inlining of macro ABC_FREE
    ((pLutClass) ? (free((char *) (pLutClass)), (pLutClass) = 0) : 0);
    return pNew;
}

////////////////////////////////////////////////////////////////////////
///                       END OF FILE                                ///
////////////////////////////////////////////////////////////////////////


//ABC_NAMESPACE_IMPL_END

