package abs

import (
	"fmt"
	"time"
)

/**CFile****************************************************************

  FileName    [absRef2.c]

  SystemName  [ABC: Logic synthesis and verification system.]

  PackageName [Abstraction package.]

  Synopsis    [Refinement manager to compute all justifying subsets.]

  Author      [Alan Mishchenko]
  
  Affiliation [UC Berkeley]

  Date        [Ver. 1.0. Started - June 20, 2005.]

  Revision    [$Id: absRef2.c,v 1.00 2005/06/20 00:00:00 alanmi Exp $]

***********************************************************************/

//#include "abs.h"
//#include "absRef2.h"

//ABC_NAMESPACE_IMPL_START

/*
    Description of the refinement manager

    This refinement manager should be 
    * started by calling Rf2_ManStart()
       this procedure takes one argument, the user's seq miter as a GIA manager
       - the manager should have only one property output
       - this manager should not change while the refinement manager is alive
       - it cannot be used by external applications for any purpose
       - when the refinement manager stop, GIA manager is the same as at the beginning
       - in the meantime, it will have some data-structures attached to its nodes...
    * stopped by calling Rf2_ManStop()
    * between starting and stopping, refinements are obtained by calling Rf2_ManRefine()

    Procedure Rf2_ManRefine() takes the following arguments:
    * the refinement manager previously started by Rf2_ManStart()
    * counter-example (CEX) obtained by abstracting some logic of GIA
    * mapping (vMap) of inputs of the CEX into the object IDs of the GIA manager
       - only PI, flop outputs, and internal AND nodes can be used in vMap
       - the ordering of objects in vMap is not important
       - however, the index of a non-PI object in vMap is used as its priority
         (the smaller the index, the more likely this non-PI object apears in a refinement)
       - only the logic between PO and the objects listed in vMap is traversed by the manager
         (as a result, GIA can be arbitrarily large, but only objects used in the abstraction
         and the pseudo-PI, that is, objects in the cut, will be visited by the manager)
    * flag fPropFanout defines whether value propagation is done through the fanout
       - it this flag is enabled, theoretically refinement should be better (the result smaller)
    * flag fVerbose may print some statistics

    The refinement manager returns a minimal-size array of integer IDs of GIA objects
    which should be added to the abstraction to possibly prevent the given counter-example
       - only flop output and internal AND nodes from vMap may appear in the resulting array
       - if the resulting array is empty, the CEX is a true CEX 
         (in other words, non-PI objects are not needed to set the PO value to 1)

    Verification of the selected refinement is performed by 
       - initializing all PI objects in vMap to value 0 or 1 they have in the CEX
       - initializing all remaining objects in vMap to value X
       - initializing objects used in the refiment to value 0 or 1 they have in the CEX
       - simulating through as many timeframes as required by the CEX
       - if the PO value in the last frame is 1, the refinement is correct
         (however, the minimality of the refinement is not currently checked)
        
*/

////////////////////////////////////////////////////////////////////////
///                        DECLARATIONS                              ///
////////////////////////////////////////////////////////////////////////

////////////////////////////////////////////////////////////////////////
///                     FUNCTION DEFINITIONS                         ///
////////////////////////////////////////////////////////////////////////
 
/**Function*************************************************************

  Synopsis    [Creates a new manager.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
/**Function*************************************************************

  Synopsis    [Collect internal objects to be used in value propagation.]

  Description [Resulting array vObjs contains RO, AND, PO/RI in a topo order.]
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
/**Function*************************************************************

  Synopsis    [Performs sensitization analysis.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Rf2_ManSensitize(p *Rf2_Man_t) int32 {
    var pRnm, pRnm0, pRnm1 *Rf2_Obj_t;
    var pObj *Gia_Obj_t;
    var f int32;
    var i int32;
    var iBit int32 = p.pCex.nRegs;
    // const0 is initialized automatically in all timeframes
    for  f = 0; f <= p.pCex.iFrame; f++, iBit += p.pCex.nPis  {
        // inlining of macro Gia_ManForEachObjVec
        for  i = 0; (i < Vec_IntSize(p.vMap)) && ((pObj) = Gia_ManObj(p.pGia, Vec_IntEntry(p.vMap,i))); i++  {
            assert(Gia_ObjIsCi(pObj) || Gia_ObjIsAnd(pObj), "Gia_ObjIsCi(pObj) || Gia_ObjIsAnd(pObj)");
            pRnm = Rf2_ManObj( p, pObj, f );
            pRnm.Value = Abc_InfoHasBit( p.pCex.pData, iBit + i );
            if !Gia_ObjIsPi(p.pGia, pObj) { // this is PPI
            
                assert(pObj.Value > 0, "pObj.Value > 0");
                pRnm.Prio = pObj.Value;
                pRnm.fPPi = 1;
            }
        }
        // inlining of macro Gia_ManForEachObjVec
        for  i = 0; (i < Vec_IntSize(p.vObjs)) && ((pObj) = Gia_ManObj(p.pGia, Vec_IntEntry(p.vObjs,i))); i++  {
            assert(Gia_ObjIsRo(p.pGia, pObj) || Gia_ObjIsAnd(pObj) || Gia_ObjIsCo(pObj), "Gia_ObjIsRo(p.pGia, pObj) || Gia_ObjIsAnd(pObj) || Gia_ObjIsCo(pObj)");
            pRnm = Rf2_ManObj( p, pObj, f );
            assert(!pRnm.fPPi, "!pRnm.fPPi");
            if Gia_ObjIsRo(p.pGia, pObj) {
                if f == 0 {
                    continue; }
                pRnm0 = Rf2_ManObj( p, Gia_ObjRoToRi(p.pGia, pObj), f-1 );
                pRnm.Value = pRnm0.Value;
                pRnm.Prio  = pRnm0.Prio;
                continue;
            }
            if Gia_ObjIsCo(pObj) {
                pRnm0 = Rf2_ManObj( p, Gia_ObjFanin0(pObj), f );
                pRnm.Value = (pRnm0.Value ^ Gia_ObjFaninC0(pObj));
                pRnm.Prio  = pRnm0.Prio;
                continue;
            }
            assert(Gia_ObjIsAnd(pObj), "Gia_ObjIsAnd(pObj)");
            pRnm0 = Rf2_ManObj( p, Gia_ObjFanin0(pObj), f );
            pRnm1 = Rf2_ManObj( p, Gia_ObjFanin1(pObj), f );
            pRnm.Value = (pRnm0.Value ^ Gia_ObjFaninC0(pObj)) & (pRnm1.Value ^ Gia_ObjFaninC1(pObj));
            if pRnm.Value == 1 {
                pRnm.Prio  = Abc_MaxInt( pRnm0.Prio, pRnm1.Prio );
            } else if (pRnm0.Value ^ Gia_ObjFaninC0(pObj)) == 0 && (pRnm1.Value ^ Gia_ObjFaninC1(pObj)) == 0 {
                pRnm.Prio  = Abc_MinInt( pRnm0.Prio, pRnm1.Prio ); // choice
            } else if (pRnm0.Value ^ Gia_ObjFaninC0(pObj)) == 0 {
                pRnm.Prio  = pRnm0.Prio;
            } else { 
                pRnm.Prio  = pRnm1.Prio; }
        }
    }
    assert(iBit == p.pCex.nBits, "iBit == p.pCex.nBits");
    pRnm = Rf2_ManObj( p, Gia_ManPo(p.pGia, 0), p.pCex.iFrame );
    if pRnm.Value != 1 {
        fmt.Printf("Output value is incorrect.\n"); }
    return pRnm.Prio;
}



/**Function*************************************************************

  Synopsis    [Performs refinement.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Rf2_ManVerifyUsingTerSim(p *Gia_Man_t, pCex *Abc_Cex_t, vMap *Vec_Int_t, vObjs *Vec_Int_t, vRes *Vec_Int_t) {
    var pObj *Gia_Obj_t;
    var i int32;
    var f int32;
    var iBit int32 = pCex.nRegs;
    Gia_ObjTerSimSet0( Gia_ManConst0(p) );
    for  f = 0; f <= pCex.iFrame; f++, iBit += pCex.nPis  {
        // inlining of macro Gia_ManForEachObjVec
        for  i = 0; (i < Vec_IntSize(vMap)) && ((pObj) = Gia_ManObj(p, Vec_IntEntry(vMap,i))); i++  {
            pObj.Value = Abc_InfoHasBit( pCex.pData, iBit + i );
            if !Gia_ObjIsPi(p, pObj) {
                Gia_ObjTerSimSetX( pObj );
            } else if pObj.Value {
                Gia_ObjTerSimSet1( pObj );
            } else {
                Gia_ObjTerSimSet0( pObj ); }
        }
        // inlining of macro Gia_ManForEachObjVec
        for  i = 0; (i < Vec_IntSize(vRes)) && ((pObj) = Gia_ManObj(p, Vec_IntEntry(vRes,i))); i++  { // vRes is subset of vMap
        
            if pObj.Value {
                Gia_ObjTerSimSet1( pObj );
            } else {
                Gia_ObjTerSimSet0( pObj ); }
        }
        // inlining of macro Gia_ManForEachObjVec
        for  i = 0; (i < Vec_IntSize(vObjs)) && ((pObj) = Gia_ManObj(p, Vec_IntEntry(vObjs,i))); i++  {
            if Gia_ObjIsCo(pObj) {
                Gia_ObjTerSimCo( pObj );
            } else if Gia_ObjIsAnd(pObj) {
                Gia_ObjTerSimAnd( pObj );
            } else if f == 0 {
                Gia_ObjTerSimSet0( pObj );
            } else {
                Gia_ObjTerSimRo( p, pObj ); }
        }
    }
    // inlining of macro Gia_ManForEachObjVec
    for  i = 0; (i < Vec_IntSize(vMap)) && ((pObj) = Gia_ManObj(p, Vec_IntEntry(vMap,i))); i++  {
        pObj.Value = 0;
    }
    pObj = Gia_ManPo( p, 0 );
    if !Gia_ObjTerSimGet1(pObj) {
        Abc_Print( 1, "\nRefinement verification has failed!!!\n" ); }
}

/**Function*************************************************************

  Synopsis    [Computes the refinement for a given counter-example.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Rf2_ManGatherFanins_rec(p *Rf2_Man_t, pObj *Gia_Obj_t, vFanins *Vec_Int_t, Depth int32, RootId int32, fFirst int32) {
    if Gia_ObjIsTravIdCurrent(p.pGia, pObj) {
        return; }
    Gia_ObjSetTravIdCurrent(p.pGia, pObj);
    if pObj.fPhase && !(fFirst != 0) {
        var vVec *Vec_Int_t  = Rf2_ObjVec( p, pObj );
//        if ( Vec_IntEntry( vVec, 0 ) == 0 )
//            return;
        if Vec_IntSize(vVec) == 0 {
            Vec_IntPush( vFanins, Gia_ObjId(p.pGia, pObj) ); }
        Vec_IntPushUnique( vVec, RootId );
        if Depth == 0 {
            return; }
    }
    if Gia_ObjIsPi(p.pGia, pObj) || Gia_ObjIsConst0(pObj) {
        return; }
    if Gia_ObjIsRo(p.pGia, pObj) {
        assert(pObj.fPhase, "pObj.fPhase");
        pObj = Gia_ObjRoToRi(p.pGia, pObj);
        Rf2_ManGatherFanins_rec( p, Gia_ObjFanin0(pObj), vFanins, Depth - 1, RootId, 0 );
    
    } else if Gia_ObjIsAnd(pObj) {
        Rf2_ManGatherFanins_rec( p, Gia_ObjFanin0(pObj), vFanins, Depth - pObj.fPhase, RootId, 0 );
        Rf2_ManGatherFanins_rec( p, Gia_ObjFanin1(pObj), vFanins, Depth - pObj.fPhase, RootId, 0 );
    
    } else { assert(false, "0"); }
}
func Rf2_ManGatherFanins(p *Rf2_Man_t, Depth int32) {
    var vUsed *Vec_Int_t;
    var vVec *Vec_Int_t;
    var pObj *Gia_Obj_t;
    var i, k, Entry int32;
    // mark PPIs
    // inlining of macro Gia_ManForEachObjVec
    for  i = 0; (i < Vec_IntSize(p.vMap)) && ((pObj) = Gia_ManObj(p.pGia, Vec_IntEntry(p.vMap,i))); i++  {
        vVec = Rf2_ObjVec( p, pObj );
        assert(Vec_IntSize(vVec) == 0, "Vec_IntSize(vVec) == 0");
        Vec_IntPush( vVec, 0 );
    }
    // collect internal
    Vec_IntClear( p.vFanins );
    // inlining of macro Gia_ManForEachObjVec
    for  i = 0; (i < Vec_IntSize(p.vMap)) && ((pObj) = Gia_ManObj(p.pGia, Vec_IntEntry(p.vMap,i))); i++  {
        if Gia_ObjIsPi(p.pGia, pObj) {
            continue; }
        Gia_ManIncrementTravId( p.pGia );
        Rf2_ManGatherFanins_rec( p, pObj, p.vFanins, Depth, i, 1 );
    }

    vUsed = Vec_IntStart( Vec_IntSize(p.vMap) );

    // evaluate collected
    fmt.Printf("\nMap (%d): ", Vec_IntSize(p.vMap));
    // inlining of macro Gia_ManForEachObjVec
    for  i = 0; (i < Vec_IntSize(p.vMap)) && ((pObj) = Gia_ManObj(p.pGia, Vec_IntEntry(p.vMap,i))); i++  {
        vVec = Rf2_ObjVec( p, pObj );
        if Vec_IntSize(vVec) > 1 {
            fmt.Printf("%d=%d ", i, Vec_IntSize(vVec) - 1); }
        // inlining of macro Vec_IntForEachEntryStart
        for  k = 1; (k < Vec_IntSize(vVec)) && (((Entry) = Vec_IntEntry(vVec, k)), 1); k++  {
            Vec_IntAddToEntry( vUsed, Entry, 1 );
        }
        Vec_IntClear( vVec );
    }
    fmt.Printf("\n");
    // evaluate internal
    fmt.Printf("Int (%d): ", Vec_IntSize(p.vFanins));
    // inlining of macro Gia_ManForEachObjVec
    for  i = 0; (i < Vec_IntSize(p.vFanins)) && ((pObj) = Gia_ManObj(p.pGia, Vec_IntEntry(p.vFanins,i))); i++  {
        vVec = Rf2_ObjVec( p, pObj );
        if Vec_IntSize(vVec) > 1 {
            fmt.Printf("%d=%d ", i, Vec_IntSize(vVec)); }
        if Vec_IntSize(vVec) > 1 {
            // inlining of macro Vec_IntForEachEntry
            for  k = 0; (k < Vec_IntSize(vVec)) && (((Entry) = Vec_IntEntry(vVec, k)), 1); k++  {
                Vec_IntAddToEntry( vUsed, Entry, 1 );
            } }
        Vec_IntClear( vVec );
    }
    fmt.Printf("\n");
    // evaluate PPIs    
    // inlining of macro Vec_IntForEachEntry
    for  k = 0; (k < Vec_IntSize(vUsed)) && (((Entry) = Vec_IntEntry(vUsed, k)), 1); k++  {
        fmt.Printf("%d ", Entry);
    }
    fmt.Printf("\n");

    Vec_IntFree( vUsed );
}


/**Function*************************************************************

  Synopsis    [Sort, make dup- and containment-free, and filter.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Rf2_ManCountPpis(p *Rf2_Man_t) int32 {
    var pObj *Gia_Obj_t;
    var i, Counter int32;
    // inlining of macro Gia_ManForEachObjVec
    for  i = 0; (i < Vec_IntSize(p.vMap)) && ((pObj) = Gia_ManObj(p.pGia, Vec_IntEntry(p.vMap,i))); i++  {
        if !Gia_ObjIsPi(p.pGia, pObj) { // this is PPI
            Counter++; }
    }
    return Counter;
}

/**Function*************************************************************

  Synopsis    [Sort, make dup- and containment-free, and filter.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Rf2_ManPrintVector(p *Vec_Int_t, Num int32) {
    var i, k, Entry int32;
    // inlining of macro Vec_IntForEachEntry
    for  i = 0; (i < Vec_IntSize(p)) && (((Entry) = Vec_IntEntry(p, i)), 1); i++  {
        for  k = 0; k < Num; k++  {
            fmt.Printf("%c", '0' + ((Entry>>k) & 1));
        }
        fmt.Printf("\n");
    }
}

/**Function*************************************************************

  Synopsis    [Sort, make dup- and containment-free, and filter.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
/**Function*************************************************************

  Synopsis    [Assigns a unique justifification ID for each PPI.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Rf2_ManAssignJustIds(p *Rf2_Man_t) int32 {
    var pObj *Gia_Obj_t;
    var nPpis int32  = Rf2_ManCountPpis( p );
    var nGroupSize int32  = (nPpis / 30) + (nPpis % 30 > 0);
    var i, k int32;
    Vec_VecClear( p.vGrp2Ppi );
    // inlining of macro Gia_ManForEachObjVec
    for  i = 0; (i < Vec_IntSize(p.vMap)) && ((pObj) = Gia_ManObj(p.pGia, Vec_IntEntry(p.vMap,i))); i++  {
        if !Gia_ObjIsPi(p.pGia, pObj) { // this is PPI
            Vec_VecPushInt( p.vGrp2Ppi, (k++ / nGroupSize), i ); }
    }
    fmt.Printf("Considering %d PPIs combined into %d groups of size %d.\n", k, (k-1)/nGroupSize+1, nGroupSize);
    return (k-1)/nGroupSize+1;
}

/**Function*************************************************************

  Synopsis    [Sort, make dup- and containment-free, and filter.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Rf2_ManPrintVectorSpecial(p *Rf2_Man_t, vVec *Vec_Int_t) {
    var pObj *Gia_Obj_t;
    var nPpis int32  = Rf2_ManCountPpis( p );
    var nGroupSize int32  = (nPpis / 30) + (nPpis % 30 > 0);
    var s, i, k, Entry, Counter int32;

    // inlining of macro Vec_IntForEachEntry
    for  s = 0; (s < Vec_IntSize(vVec)) && (((Entry) = Vec_IntEntry(vVec, s)), 1); s++  {
        k = 0;
        Counter = 0;
        // inlining of macro Gia_ManForEachObjVec
        for  i = 0; (i < Vec_IntSize(p.vMap)) && ((pObj) = Gia_ManObj(p.pGia, Vec_IntEntry(p.vMap,i))); i++  {
            if !Gia_ObjIsPi(p.pGia, pObj) { // this is PPI
            
                if (Entry >> (k++ / nGroupSize)) & 1 {
                    fmt.Printf("1");
Counter++;
                } else {
                    fmt.Printf("0"); }
            
            } else {
                fmt.Printf("-"); }
        }
        fmt.Printf(" %3d \n", Counter);
    }
}

/**Function*************************************************************

  Synopsis    [Performs justification propagation.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Rf2_ManPropagate(p *Rf2_Man_t, Limit int32) *Vec_Int_t {
    var vVec, vVec0, vVec1 *Vec_Int_t;
    var pObj *Gia_Obj_t;
    var f int32;
    var i int32;
    var k int32;
    var j int32;
    var Entry int32;
    var Entry2 int32;
    var iBit int32 = p.pCex.nRegs;
    // init constant
    pObj = Gia_ManConst0(p.pGia);
    pObj.fMark0 = 0;
    Vec_IntFill( Rf2_ObjVec(p, pObj), 1, 0 );
    // iterate through the timeframes
    for  f = 0; f <= p.pCex.iFrame; f++, iBit += p.pCex.nPis  {
        // initialize frontier values and init justification sets
        // inlining of macro Gia_ManForEachObjVec
        for  i = 0; (i < Vec_IntSize(p.vMap)) && ((pObj) = Gia_ManObj(p.pGia, Vec_IntEntry(p.vMap,i))); i++  {
            assert(Gia_ObjIsCi(pObj) || Gia_ObjIsAnd(pObj), "Gia_ObjIsCi(pObj) || Gia_ObjIsAnd(pObj)");
            pObj.fMark0 = Abc_InfoHasBit( p.pCex.pData, iBit + i );
            Vec_IntFill( Rf2_ObjVec(p, pObj), 1, 0 );
        }
        // assign justification sets for PPis
        // inlining of macro Vec_VecForEachLevelInt
        for  i = 0; (i < Vec_VecSize(p.vGrp2Ppi)) && (((vVec) = Vec_VecEntryInt(p.vGrp2Ppi, i)), 1); i++  {
            // inlining of macro Vec_IntForEachEntry
            for  k = 0; (k < Vec_IntSize(vVec)) && (((Entry) = Vec_IntEntry(vVec, k)), 1); k++  {
                assert(i < 31, "i < 31");
                pObj = Gia_ManObj( p.pGia, Vec_IntEntry(p.vMap, Entry) );
                assert(Vec_IntSize(Rf2_ObjVec(p, pObj)) == 1, "Vec_IntSize(Rf2_ObjVec(p, pObj)) == 1");
                Vec_IntAddToEntry( Rf2_ObjVec(p, pObj), 0, (1 << i) );
            }
        }
        // propagate internal nodes
        // inlining of macro Gia_ManForEachObjVec
        for  i = 0; (i < Vec_IntSize(p.vObjs)) && ((pObj) = Gia_ManObj(p.pGia, Vec_IntEntry(p.vObjs,i))); i++  {
            pObj.fMark0 = 0;
            vVec = Rf2_ObjVec(p, pObj);
            Vec_IntClear( vVec );
            if Gia_ObjIsRo(p.pGia, pObj) {
                if f == 0 {
                    Vec_IntPush( vVec, 0 );
                    continue;
                }
                pObj.fMark0 = Gia_ObjRoToRi(p.pGia, pObj).fMark0;
                vVec0 = Rf2_ObjVec( p, Gia_ObjRoToRi(p.pGia, pObj) );
                Vec_IntAppend( vVec, vVec0 );
                continue;
            }
            if Gia_ObjIsCo(pObj) {
                pObj.fMark0 = (Gia_ObjFanin0(pObj).fMark0 ^ Gia_ObjFaninC0(pObj));
                vVec0 = Rf2_ObjVec( p, Gia_ObjFanin0(pObj) );
                Vec_IntAppend( vVec, vVec0 );
                continue;
            }
            assert(Gia_ObjIsAnd(pObj), "Gia_ObjIsAnd(pObj)");
            vVec0 = Rf2_ObjVec(p, Gia_ObjFanin0(pObj));
            vVec1 = Rf2_ObjVec(p, Gia_ObjFanin1(pObj));
            pObj.fMark0 = (Gia_ObjFanin0(pObj).fMark0 ^ Gia_ObjFaninC0(pObj)) & (Gia_ObjFanin1(pObj).fMark0 ^ Gia_ObjFaninC1(pObj));
            if pObj.fMark0 == 1 {
                // inlining of macro Vec_IntForEachEntry
                for  k = 0; (k < Vec_IntSize(vVec0)) && (((Entry) = Vec_IntEntry(vVec0, k)), 1); k++  {
                    // inlining of macro Vec_IntForEachEntry
                    for  j = 0; (j < Vec_IntSize(vVec1)) && (((Entry2) = Vec_IntEntry(vVec1, j)), 1); j++  {
                        Vec_IntPush( vVec, Entry | Entry2 );
                    }
                }
                Rf2_ManProcessVector( vVec, Limit );
            
            } else if (Gia_ObjFanin0(pObj).fMark0 ^ Gia_ObjFaninC0(pObj)) == 0 && (Gia_ObjFanin1(pObj).fMark0 ^ Gia_ObjFaninC1(pObj)) == 0 {
                Vec_IntAppend( vVec, vVec0 );
                Vec_IntAppend( vVec, vVec1 );
                Rf2_ManProcessVector( vVec, Limit );
            
            } else if (Gia_ObjFanin0(pObj).fMark0 ^ Gia_ObjFaninC0(pObj)) == 0 {
                Vec_IntAppend( vVec, vVec0 );
            } else { 
                Vec_IntAppend( vVec, vVec1 ); }
        }
    }
    assert(iBit == p.pCex.nBits, "iBit == p.pCex.nBits");
    if Gia_ManPo(p.pGia, 0).fMark0 != 1 {
        fmt.Printf("Output value is incorrect.\n"); }
    return Rf2_ObjVec(p, Gia_ManPo(p.pGia, 0));
}

/**Function*************************************************************

  Synopsis    [Performs justification propagation.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Rf2_ManBounds(p *Rf2_Man_t) {
    var pObj *Gia_Obj_t;
    var f int32;
    var i int32;
    var iBit int32 = p.pCex.nRegs;
    // init constant
    pObj = Gia_ManConst0(p.pGia);
    pObj.fMark0 = 0;
    Rf2_ObjStart( p, pObj, Vec_IntSize(p.vMap) + Vec_IntSize(p.vObjs) );
    // iterate through the timeframes
    for  f = 0; f <= p.pCex.iFrame; f++, iBit += p.pCex.nPis  {
        // initialize frontier values and init justification sets
        // inlining of macro Gia_ManForEachObjVec
        for  i = 0; (i < Vec_IntSize(p.vMap)) && ((pObj) = Gia_ManObj(p.pGia, Vec_IntEntry(p.vMap,i))); i++  {
            assert(Gia_ObjIsCi(pObj) || Gia_ObjIsAnd(pObj), "Gia_ObjIsCi(pObj) || Gia_ObjIsAnd(pObj)");
            pObj.fMark0 = Abc_InfoHasBit( p.pCex.pData, iBit + i );
            Rf2_ObjStart( p, pObj, i );
        }
        // propagate internal nodes
        // inlining of macro Gia_ManForEachObjVec
        for  i = 0; (i < Vec_IntSize(p.vObjs)) && ((pObj) = Gia_ManObj(p.pGia, Vec_IntEntry(p.vObjs,i))); i++  {
            pObj.fMark0 = 0;
            Rf2_ObjClear( p, pObj );
            if Gia_ObjIsRo(p.pGia, pObj) {
                if f == 0 {
                    Rf2_ObjStart( p, pObj, Vec_IntSize(p.vMap) + i );
                    continue;
                }
                pObj.fMark0 = Gia_ObjRoToRi(p.pGia, pObj).fMark0;
                Rf2_ObjCopy( p, pObj, Gia_ObjRoToRi(p.pGia, pObj) );
                continue;
            }
            if Gia_ObjIsCo(pObj) {
                pObj.fMark0 = (Gia_ObjFanin0(pObj).fMark0 ^ Gia_ObjFaninC0(pObj));
                Rf2_ObjCopy( p, pObj, Gia_ObjFanin0(pObj) );
                continue;
            }
            assert(Gia_ObjIsAnd(pObj), "Gia_ObjIsAnd(pObj)");
            pObj.fMark0 = (Gia_ObjFanin0(pObj).fMark0 ^ Gia_ObjFaninC0(pObj)) & (Gia_ObjFanin1(pObj).fMark0 ^ Gia_ObjFaninC1(pObj));
            if pObj.fMark0 == 1 {
                Rf2_ObjDeriveAnd( p, pObj, 1 );
            } else if (Gia_ObjFanin0(pObj).fMark0 ^ Gia_ObjFaninC0(pObj)) == 0 && (Gia_ObjFanin1(pObj).fMark0 ^ Gia_ObjFaninC1(pObj)) == 0 {
                Rf2_ObjDeriveAnd( p, pObj, 0 );
            } else if (Gia_ObjFanin0(pObj).fMark0 ^ Gia_ObjFaninC0(pObj)) == 0 {
                Rf2_ObjCopy( p, pObj, Gia_ObjFanin0(pObj) );
            } else { 
                Rf2_ObjCopy( p, pObj, Gia_ObjFanin1(pObj) ); }
        }
    }
    assert(iBit == p.pCex.nBits, "iBit == p.pCex.nBits");
    if Gia_ManPo(p.pGia, 0).fMark0 != 1 {
        fmt.Printf("Output value is incorrect.\n"); }

    fmt.Printf("Bounds: \n");
    Rf2_ObjPrint( p, Gia_ManPo(p.pGia, 0) );
}

/**Function*************************************************************

  Synopsis    [Computes the refinement for a given counter-example.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Rf2_ManRefine(p *Rf2_Man_t, pCex *Abc_Cex_t, vMap *Vec_Int_t, fPropFanout int32, fVerbose int32) *Vec_Int_t {
    var vJusts *Vec_Int_t;
//    Vec_Int_t * vSelected = Vec_IntAlloc( 100 );
    var vSelected *Vec_Int_t  = nil;
    var clk int64;
    var clk2 int64 = time.Now().UnixNano();
    var nGroups int32;
    p.nCalls++;
    // initialize
    p.pCex = pCex;
    p.vMap = vMap;
    p.fPropFanout = fPropFanout;
    p.fVerbose    = fVerbose;
    // collects used objects
    Rf2_ManCollect( p );
    // collect reconvergence points
//    Rf2_ManGatherFanins( p, 2 );
    // propagate justification IDs
    nGroups = Rf2_ManAssignJustIds( p );
    vJusts = Rf2_ManPropagate( p, 32 );

//    printf( "\n" );
//    Rf2_ManPrintVector( vJusts, nGroups );
    Rf2_ManPrintVectorSpecial( p, vJusts );
    if Vec_IntSize(vJusts) == 0 {
        fmt.Printf("Empty set of justifying subsets.\n");
        return nil;
    }

//    p->nMapWords = Abc_BitWordNum( Vec_IntSize(p->vMap) + Vec_IntSize(p->vObjs) + 1 ); // Map + Flops + Const
//    Rf2_ManBounds( p );

    // select the result
//    Abc_PrintTime( 1, "Time", clock() - clk2 );

    // verify (empty) refinement
    clk = time.Now().UnixNano();
//    Rf2_ManVerifyUsingTerSim( p->pGia, p->pCex, p->vMap, p->vObjs, vSelected );
//    Vec_IntUniqify( vSelected );
//    Vec_IntReverseOrder( vSelected );
    p.timeVer += time.Now().UnixNano() - clk;
    p.timeTotal += time.Now().UnixNano() - clk2;
//    p->nRefines += Vec_IntSize(vSelected);
    return vSelected;
}

////////////////////////////////////////////////////////////////////////
///                       END OF FILE                                ///
////////////////////////////////////////////////////////////////////////


//ABC_NAMESPACE_IMPL_END

