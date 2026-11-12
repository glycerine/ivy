package saig

import (
	"fmt"
	"time"
)

/**CFile****************************************************************

  FileName    [saigRetFwd.c]

  SystemName  [ABC: Logic synthesis and verification system.]

  PackageName [Sequential AIG package.]

  Synopsis    [Most-forward retiming.]

  Author      [Alan Mishchenko]
  
  Affiliation [UC Berkeley]

  Date        [Ver. 1.0. Started - June 20, 2005.]

  Revision    [$Id: saigRetFwd.c,v 1.00 2005/06/20 00:00:00 alanmi Exp $]

***********************************************************************/

//#include "saig.h"

//ABC_NAMESPACE_IMPL_START


////////////////////////////////////////////////////////////////////////
///                        DECLARATIONS                              ///
////////////////////////////////////////////////////////////////////////

func Aig_ObjFanoutStatic(pObj *Aig_Obj_t, i int32) *Aig_Obj_t { return ((Aig_Obj_t **)pObj.pData)[i];             }
func Aig_ObjSetFanoutStatic(pObj *Aig_Obj_t, pFan *Aig_Obj_t) { ((Aig_Obj_t **)pObj.pData)[pObj.nRefs++] = pFan; }

// Original macro:
// #define Aig_ObjForEachFanoutStatic( pObj, pFan, i )                        \
//     for ( i = 0; (i < (int)(pObj)->nRefs) && ((pFan) = Aig_ObjFanoutStatic(pObj, i)); i++ )


////////////////////////////////////////////////////////////////////////
///                     FUNCTION DEFINITIONS                         ///
////////////////////////////////////////////////////////////////////////

/**Function*************************************************************

  Synopsis    [Allocate static fanout for all nodes in the AIG manager.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Aig_ManStaticFanoutStart(p *Aig_Man_t) **Aig_Obj_t {
    var ppFanouts **Aig_Obj_t;
    var pObj *Aig_Obj_t;
    var i, nFanouts, nFanoutsAlloc int32;
    // allocate fanouts
    nFanoutsAlloc = 2 * Aig_ManObjNumMax(p) - Aig_ManCiNum(p) - Aig_ManCoNum(p);
    // inlining of macro ABC_ALLOC
    ppFanouts = ((Aig_Obj_t * *) malloc(sizeof(Aig_Obj_t *) * (size_t)(nFanoutsAlloc)));
    // mark up storage
    nFanouts = 0;
    // inlining of macro Aig_ManForEachObj
    // inlining of macro Vec_PtrForEachEntry
    for  i = 0; (i < Vec_PtrSize(p.vObjs)) && (((pObj) = (Aig_Obj_t *)Vec_PtrEntry(p.vObjs, i)), 1); i++  { if (pObj) == nil {} else {
        pObj.pData = ppFanouts + nFanouts;
        nFanouts += pObj.nRefs;  
        pObj.nRefs = 0;
    }
    }
    assert(nFanouts < nFanoutsAlloc, "nFanouts < nFanoutsAlloc"); 
    // add fanouts
    // inlining of macro Aig_ManForEachObj
    // inlining of macro Vec_PtrForEachEntry
    for  i = 0; (i < Vec_PtrSize(p.vObjs)) && (((pObj) = (Aig_Obj_t *)Vec_PtrEntry(p.vObjs, i)), 1); i++  { if (pObj) == nil {} else {
        if Aig_ObjChild0(pObj) {
            Aig_ObjSetFanoutStatic( Aig_ObjFanin0(pObj), pObj ); }
        if Aig_ObjChild1(pObj) {
            Aig_ObjSetFanoutStatic( Aig_ObjFanin1(pObj), pObj ); }
    }
    }
    return ppFanouts;
}

/**Function*************************************************************

  Synopsis    [Marks the objects reachable from the given object.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Aig_ManMarkAutonomous_rec(p *Aig_Man_t, pObj *Aig_Obj_t) {
    var pFanout *Aig_Obj_t;
    var i int32;
    if Aig_ObjIsTravIdCurrent(p, pObj) {
        return; }
    Aig_ObjSetTravIdCurrent(p, pObj);
    // inlining of macro Aig_ObjForEachFanoutStatic
    for  i = 0; (i < int32((pObj).nRefs)) && ((pFanout) = Aig_ObjFanoutStatic(pObj, i)); i++  {
        Aig_ManMarkAutonomous_rec( p, pFanout );
    }
}

/**Function*************************************************************

  Synopsis    [Marks with current trav ID nodes reachable from Const1 and PIs.]

  Description [Returns the number of unreachable registers.]
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Saig_ManMarkAutonomous(p *Aig_Man_t) {
    var ppFanouts **Aig_Obj_t;
    var pObj, pObjLi, pObjLo *Aig_Obj_t;
    var i int32;
    // temporarily connect register outputs to register inputs
    // inlining of macro Saig_ManForEachLiLo
    for  i = 0; (i < Saig_ManRegNum(p)) && (((pObjLi) = Saig_ManLi(p, i)), 1)            && (((pObjLo)=Saig_ManLo(p, i)), 1); i++  {
        pObjLo.pFanin0 = pObjLi;
        pObjLi.nRefs = 1;
    }
    // mark nodes reachable from Const1 and PIs
    Aig_ManIncrementTravId( p );
    ppFanouts = Aig_ManStaticFanoutStart( p );
    Aig_ManMarkAutonomous_rec( p, Aig_ManConst1(p) );
    // inlining of macro Saig_ManForEachPi
    // inlining of macro Vec_PtrForEachEntryStop
    for  i = 0; (i < Saig_ManPiNum(p)) && (((pObj) = (Aig_Obj_t *)Vec_PtrEntry(p.vCis, i)), 1); i++  {
        Aig_ManMarkAutonomous_rec( p, pObj );
    }
    // inlining of macro ABC_FREE
    ((ppFanouts) ? (free((char *) (ppFanouts)), (ppFanouts) = 0) : 0);
    // disconnect LIs/LOs and label unreachable registers
    // inlining of macro Saig_ManForEachLiLo
    for  i = 0; (i < Saig_ManRegNum(p)) && (((pObjLi) = Saig_ManLi(p, i)), 1)            && (((pObjLo)=Saig_ManLo(p, i)), 1); i++  {
        assert(pObjLo.pFanin0 && pObjLi.nRefs == 1, "pObjLo.pFanin0 && pObjLi.nRefs == 1");
        pObjLo.pFanin0 = nil;
        pObjLi.nRefs = 0;
    }
}

/**Function*************************************************************

  Synopsis    [Derives the cut for forward retiming.]

  Description [Assumes topological ordering of the nodes.]
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Saig_ManRetimeForwardOne(p *Aig_Man_t, pnRegFixed *int32, pnRegMoves *int32) *Aig_Man_t {
    var pNew *Aig_Man_t;
    var vCut *Vec_Ptr_t;
    var pObj, pFanin *Aig_Obj_t;
    var i int32;
    // mark the retimable nodes
    Saig_ManMarkAutonomous( p );
    // mark the retimable registers with the fresh trav ID
    Aig_ManIncrementTravId( p );
    *pnRegFixed = 0;
    // inlining of macro Saig_ManForEachLo
    for  i = 0; (i < Saig_ManRegNum(p)) && (((pObj) = (Aig_Obj_t *)Vec_PtrEntry(p.vCis, i+Saig_ManPiNum(p))), 1); i++  {
        if Aig_ObjIsTravIdPrevious(p, pObj) {
            Aig_ObjSetTravIdCurrent(p, pObj);
        } else {
            (*pnRegFixed)++; }
    }
    // mark all the nodes that can be retimed forward
    *pnRegMoves = 0;
    // inlining of macro Aig_ManForEachNode
    // inlining of macro Vec_PtrForEachEntry
    for  i = 0; (i < Vec_PtrSize(p.vObjs)) && (((pObj) = (Aig_Obj_t *)Vec_PtrEntry(p.vObjs, i)), 1); i++  { if (pObj) == nil || !Aig_ObjIsNode(pObj) {} else
        if Aig_ObjIsTravIdCurrent(p, Aig_ObjFanin0(pObj)) && Aig_ObjIsTravIdCurrent(p, Aig_ObjFanin1(pObj)) {
            Aig_ObjSetTravIdCurrent(p, pObj);
            (*pnRegMoves)++;
        }
    }
    // mark the remaining registers
    // inlining of macro Saig_ManForEachLo
    for  i = 0; (i < Saig_ManRegNum(p)) && (((pObj) = (Aig_Obj_t *)Vec_PtrEntry(p.vCis, i+Saig_ManPiNum(p))), 1); i++  {
        Aig_ObjSetTravIdCurrent(p, pObj);
    }
    // find the cut (all such marked objects that fanout into unmarked nodes)
    vCut = Vec_PtrAlloc( 1000 );
    Aig_ManIncrementTravId( p );
    // inlining of macro Aig_ManForEachObj
    // inlining of macro Vec_PtrForEachEntry
    for  i = 0; (i < Vec_PtrSize(p.vObjs)) && (((pObj) = (Aig_Obj_t *)Vec_PtrEntry(p.vObjs, i)), 1); i++  { if (pObj) == nil {} else {
        if Aig_ObjIsTravIdPrevious(p, pObj) {
            continue; }
        pFanin = Aig_ObjFanin0(pObj);
        if (pFanin != nil) && Aig_ObjIsTravIdPrevious(p, pFanin) {
            Vec_PtrPush( vCut, pFanin );
            Aig_ObjSetTravIdCurrent( p, pFanin );
        }
        pFanin = Aig_ObjFanin1(pObj);
        if (pFanin != nil) && Aig_ObjIsTravIdPrevious(p, pFanin) {
            Vec_PtrPush( vCut, pFanin );
            Aig_ObjSetTravIdCurrent( p, pFanin );
        }
    }
    }
    // finally derive the new manager
    pNew = Saig_ManRetimeDupForward( p, vCut );
    Vec_PtrFree( vCut );
    return pNew;
}

/**Function*************************************************************

  Synopsis    [Derives the cut for forward retiming.]

  Description [Assumes topological ordering of the nodes.]
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Saig_ManRetimeForward(p *Aig_Man_t, nMaxIters int32, fVerbose int32) *Aig_Man_t {
    var pNew, pTemp *Aig_Man_t;
    var i int32;
    var nRegFixed int32;
    var nRegMoves int32 = 1;
    var clk abctime;
    pNew = p;
    for  i = 0; i < nMaxIters && nRegMoves > 0; i++  {
        clk = Abc_Clock();
        pNew = Saig_ManRetimeForwardOne( pTemp = pNew, &nRegFixed, &nRegMoves );
        if fVerbose != 0 {
            fmt.Printf("%2d : And = %6d. Reg = %5d. Unret = %5d. Move = %6d. ", i + 1, Aig_ManNodeNum(pTemp), Aig_ManRegNum(pTemp), nRegFixed, nRegMoves);
            // inlining of macro ABC_PRT
            Abc_Print(1, "%s =", ("Time"));
Abc_Print(1, "%9.2f sec\n", 1.0*(float64((Abc_Clock() - clk)))/(float64(int64(time.Second))));;
        }
        if pTemp != p {
            Aig_ManStop( pTemp ); }
    }
    clk = Abc_Clock();
    pNew = Aig_ManReduceLaches( pNew, fVerbose );
    if fVerbose != 0 {
        // inlining of macro ABC_PRT
        Abc_Print(1, "%s =", ("Register sharing time"));
Abc_Print(1, "%9.2f sec\n", 1.0*(float64((Abc_Clock() - clk)))/(float64(int64(time.Second))));;
    }
    return pNew;
}


////////////////////////////////////////////////////////////////////////
///                       END OF FILE                                ///
////////////////////////////////////////////////////////////////////////


//ABC_NAMESPACE_IMPL_END

