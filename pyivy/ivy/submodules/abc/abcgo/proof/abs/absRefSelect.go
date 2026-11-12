package abs

import "fmt"

/**CFile****************************************************************

  FileName    [absRefSelect.c]

  SystemName  [ABC: Logic synthesis and verification system.]

  PackageName [Abstraction package.]

  Synopsis    [Post-processes the set of selected refinement objects.]

  Author      [Alan Mishchenko]
  
  Affiliation [UC Berkeley]

  Date        [Ver. 1.0. Started - June 20, 2005.]

  Revision    [$Id: absRefSelect.c,v 1.00 2005/06/20 00:00:00 alanmi Exp $]

***********************************************************************/
 
//#include "abs.h"
//#include "absRef.h"

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
func Rnm_ManPrintSelected(p *Rnm_Man_t, vNewPPis *Vec_Int_t) {
    var pObj *Gia_Obj_t;
    var i, Counter int32;
    // inlining of macro Gia_ManForEachObjVec
    for  i = 0; (i < Vec_IntSize(p.vMap)) && ((pObj) = Gia_ManObj(p.pGia, Vec_IntEntry(p.vMap,i))); i++  {
        if Gia_ObjIsPi(p.pGia, pObj) { 
            fmt.Printf("-");
        } else if Vec_IntFind(vNewPPis, Gia_ObjId(p.pGia, pObj)) >= 0 {// this is PPI
            fmt.Printf("1");
Counter++;
        } else {
            fmt.Printf("0"); }
    }
    fmt.Printf(" %3d\n", Counter);
}

/**Function*************************************************************

  Synopsis    [Perform structural analysis.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Ga2_StructAnalize(p *Gia_Man_t, vFront *Vec_Int_t, vInter *Vec_Int_t, vNewPPis *Vec_Int_t) {
    var vFanins *Vec_Int_t;
    var pObj, pFanin *Gia_Obj_t;
    var i, k int32;
    // clean labels
    // inlining of macro Gia_ManForEachObj
    for  i = 0; (i < p.nObjs) && ((pObj) = Gia_ManObj(p, i)); i++  {
        pObj.fMark0 = pObj.fMark1 = 0;
    }
    // label frontier
    // inlining of macro Gia_ManForEachObjVec
    for  i = 0; (i < Vec_IntSize(vFront)) && ((pObj) = Gia_ManObj(p, Vec_IntEntry(vFront,i))); i++  {
        pObj.fMark0 = 1, pObj.fMark1 = 0;
    }
    // label objects
    // inlining of macro Gia_ManForEachObjVec
    for  i = 0; (i < Vec_IntSize(vInter)) && ((pObj) = Gia_ManObj(p, Vec_IntEntry(vInter,i))); i++  {
        pObj.fMark1 = 0, pObj.fMark1 = 1;
    }
    // label selected
    // inlining of macro Gia_ManForEachObjVec
    for  i = 0; (i < Vec_IntSize(vNewPPis)) && ((pObj) = Gia_ManObj(p, Vec_IntEntry(vNewPPis,i))); i++  {
        pObj.fMark1 = 1, pObj.fMark1 = 1;
    }
    // explore selected
    // inlining of macro Gia_ManForEachObjVec
    for  i = 0; (i < Vec_IntSize(vNewPPis)) && ((pObj) = Gia_ManObj(p, Vec_IntEntry(vNewPPis,i))); i++  {
        fmt.Printf("Selected PPI %3d : ", i+1);
        fmt.Printf("%6d ", Gia_ObjId(p, pObj));
        fmt.Printf("\n");
        vFanins = Ga2_ObjLeaves( p, pObj );
        // inlining of macro Gia_ManForEachObjVec
        for  k = 0; (k < Vec_IntSize(vFanins)) && ((pFanin) = Gia_ManObj(p, Vec_IntEntry(vFanins,k))); k++  {
            fmt.Printf("    ");
            fmt.Printf("%6d ", Gia_ObjId(p, pFanin));
            if pFanin.fMark0 && pFanin.fMark1 {
                fmt.Printf("selected PPI");
            } else if pFanin.fMark0 && !pFanin.fMark1 {
                fmt.Printf("frontier (original PI or PPI)");
            } else if !pFanin.fMark0 &&  pFanin.fMark1 {
                fmt.Printf("abstracted node");
            } else if !pFanin.fMark0 && !pFanin.fMark1 {
                fmt.Printf("free variable"); }
            fmt.Printf("\n");
        }
    }
}

/**Function*************************************************************

  Synopsis    [Postprocessing the set of PPIs using structural analysis.]

  Description [The following sets are used:
  The set of all PI+PPI is in p->vMap.
  The set of all abstracted objects is in p->vObjs;
  The set of important PPIs is in vOldPPis.
  The new set of selected PPIs is in vNewPPis.]
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Rnm_ManFilterSelected(p *Rnm_Man_t, vOldPPis *Vec_Int_t) *Vec_Int_t {
    var fVerbose int32;
    var vNewPPis, vFanins *Vec_Int_t;
    var pObj, pFanin *Gia_Obj_t;
    var i int32;
    var k int32;
    var RetValue int32;
    var Counters [3]int32 ;

    // (0) make sure fanin counters are 0 at the beginning
//    Gia_ManForEachObj( p->pGia, pObj, i )
//        assert( Rnm_ObjCount(p, pObj) == 0 );

    // (1) increment PPI fanin counters
    Vec_IntClear( p.vFanins );
    // inlining of macro Gia_ManForEachObjVec
    for  i = 0; (i < Vec_IntSize(vOldPPis)) && ((pObj) = Gia_ManObj(p.pGia, Vec_IntEntry(vOldPPis,i))); i++  {
        vFanins = Ga2_ObjLeaves( p.pGia, pObj );
        // inlining of macro Gia_ManForEachObjVec
        for  k = 0; (k < Vec_IntSize(vFanins)) && ((pFanin) = Gia_ManObj(p.pGia, Vec_IntEntry(vFanins,k))); k++  {
            if Rnm_ObjAddToCount(p, pFanin) == 0 { // fanin counter is 0 -- save it
                Vec_IntPush( p.vFanins, Gia_ObjId(p.pGia, pFanin) ); }
        }
    }

    // (3) select objects with reconvergence, which create potential constraints
    // - flop objects
    // - objects whose fanin belongs to the justified area
    // - objects whose fanins overlap
    // (these do not guantee reconvergence, but may potentially have it)
    // (other objects cannot have reconvergence, even if they are added)
    vNewPPis = Vec_IntAlloc( 100 );
    // inlining of macro Gia_ManForEachObjVec
    for  i = 0; (i < Vec_IntSize(vOldPPis)) && ((pObj) = Gia_ManObj(p.pGia, Vec_IntEntry(vOldPPis,i))); i++  {
        if Gia_ObjIsRo(p.pGia, pObj) {
            if fVerbose != 0 {
                Counters[0]++; }
            Vec_IntPush( vNewPPis, Gia_ObjId(p.pGia, pObj) );
            continue;
        }
        vFanins = Ga2_ObjLeaves( p.pGia, pObj );
        // inlining of macro Gia_ManForEachObjVec
        for  k = 0; (k < Vec_IntSize(vFanins)) && ((pFanin) = Gia_ManObj(p.pGia, Vec_IntEntry(vFanins,k))); k++  {
            if Rnm_ObjIsJust(p, pFanin) || Rnm_ObjCount(p, pFanin) > 1 {
                if fVerbose != 0 {
                    Counters[1] += Rnm_ObjIsJust(p, pFanin); }
                if fVerbose != 0 {
                    Counters[2] += (Rnm_ObjCount(p, pFanin) > 1); }
                Vec_IntPush( vNewPPis, Gia_ObjId(p.pGia, pObj) );
                break;
            }
        }
    }
    RetValue = Vec_IntUniqify( vNewPPis );
    assert(RetValue == 0, "RetValue == 0");

    // (4) clear fanin counters
    // this is important for counters to be correctly set in the future iterations -- see step (0)
    // inlining of macro Gia_ManForEachObjVec
    for  i = 0; (i < Vec_IntSize(p.vFanins)) && ((pObj) = Gia_ManObj(p.pGia, Vec_IntEntry(p.vFanins,i))); i++  {
        Rnm_ObjSetCount( p, pObj, 0 );
    }

    // visualize
    if fVerbose != 0 {
        fmt.Printf("*** Refinement %3d : PI+PPI =%4d. Old =%4d. New =%4d.   FF =%4d. Just =%4d. Shared =%4d.\n", p.nRefId, Vec_IntSize(p.vMap), Vec_IntSize(vOldPPis), Vec_IntSize(vNewPPis), Counters[0], Counters[1], Counters[2]); }

//    Rnm_ManPrintSelected( p, vNewPPis );
//    Ga2_StructAnalize( p->pGia, p->vMap, p->vObjs, vNewPPis );
    return vNewPPis;
}


/**Function*************************************************************

  Synopsis    [Improved postprocessing the set of PPIs.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
var spStatic_Rnm_ManFilterSelectedNew_Counter int32;
func Rnm_ManFilterSelectedNew(p *Rnm_Man_t, vOldPPis *Vec_Int_t) *Vec_Int_t {
    
    var fVerbose int32;
    var vNewPPis, vFanins, vFanins2 *Vec_Int_t;
    var pObj, pFanin, pFanin2 *Gia_Obj_t;
    var i int32;
    var k int32;
    var k2 int32;
    var RetValue int32;
    var Counters [3]int32 ;

    // return full set of PPIs once in a while
    if spStatic_Rnm_ManFilterSelectedNew_Counter++; spStatic_Rnm_ManFilterSelectedNew_Counter % 9 == 0 {
        return Vec_IntDup( vOldPPis ); }
    return Rnm_ManFilterSelected( p, vOldPPis );

    // (0) make sure fanin counters are 0 at the beginning
//    Gia_ManForEachObj( p->pGia, pObj, i )
//        assert( Rnm_ObjCount(p, pObj) == 0 );

    // (1) increment two levels of PPI fanin counters
    Vec_IntClear( p.vFanins );
    // inlining of macro Gia_ManForEachObjVec
    for  i = 0; (i < Vec_IntSize(vOldPPis)) && ((pObj) = Gia_ManObj(p.pGia, Vec_IntEntry(vOldPPis,i))); i++  {
        // go through the fanins
        vFanins = Ga2_ObjLeaves( p.pGia, pObj );
        // inlining of macro Gia_ManForEachObjVec
        for  k = 0; (k < Vec_IntSize(vFanins)) && ((pFanin) = Gia_ManObj(p.pGia, Vec_IntEntry(vFanins,k))); k++  {
            Rnm_ObjAddToCount(p, pFanin);
            if Rnm_ObjIsJust(p, pFanin) {   // included in the abstraction
                Rnm_ObjAddToCount(p, pFanin); } // count it second time!
            Vec_IntPush( p.vFanins, Gia_ObjId(p.pGia, pFanin) );

            // go through the fanins of the fanins
            vFanins2 = Ga2_ObjLeaves( p.pGia, pFanin );
            // inlining of macro Gia_ManForEachObjVec
            for  k2 = 0; (k2 < Vec_IntSize(vFanins2)) && ((pFanin2) = Gia_ManObj(p.pGia, Vec_IntEntry(vFanins2,k2))); k2++  {
                Rnm_ObjAddToCount(p, pFanin2);
                if Rnm_ObjIsJust(p, pFanin2) {   // included in the abstraction
                    Rnm_ObjAddToCount(p, pFanin2); } // count it second time!
                Vec_IntPush( p.vFanins, Gia_ObjId(p.pGia, pFanin2) );
            }
        }
    }

    // (3) select objects with reconvergence, which create potential constraints
    // - flop objects - yes
    // - objects whose fanin (or fanins' fanin) belongs to the justified area - yes
    // - objects whose fanins (or fanins' fanin) overlap - yes
    // (these do not guantee reconvergence, but may potentially have it)
    // (other objects cannot have reconvergence, even if they are added)
    vNewPPis = Vec_IntAlloc( 100 );
    // inlining of macro Gia_ManForEachObjVec
    for  i = 0; (i < Vec_IntSize(vOldPPis)) && ((pObj) = Gia_ManObj(p.pGia, Vec_IntEntry(vOldPPis,i))); i++  {
        if Gia_ObjIsRo(p.pGia, pObj) {
            if fVerbose != 0 {
                Counters[0]++; }
            Vec_IntPush( vNewPPis, Gia_ObjId(p.pGia, pObj) );
            continue;
        } 
        // go through the first fanins
        vFanins = Ga2_ObjLeaves( p.pGia, pObj );
        // inlining of macro Gia_ManForEachObjVec
        for  k = 0; (k < Vec_IntSize(vFanins)) && ((pFanin) = Gia_ManObj(p.pGia, Vec_IntEntry(vFanins,k))); k++  {
            if Rnm_ObjCount(p, pFanin) > 1 {
                Vec_IntPush( vNewPPis, Gia_ObjId(p.pGia, pObj) ); }
            continue;

            // go through the fanins of the fanins
            vFanins2 = Ga2_ObjLeaves( p.pGia, pFanin );
            // inlining of macro Gia_ManForEachObjVec
            for  k2 = 0; (k2 < Vec_IntSize(vFanins2)) && ((pFanin2) = Gia_ManObj(p.pGia, Vec_IntEntry(vFanins2,k2))); k2++  {
                if Rnm_ObjCount(p, pFanin2) > 1 {
//                    Vec_IntPush( vNewPPis, Gia_ObjId(p->pGia, pFanin) );
                    Vec_IntPush( vNewPPis, Gia_ObjId(p.pGia, pObj) );
                }
            }
        }
    }
    RetValue = Vec_IntUniqify( vNewPPis );
//    assert( RetValue == 0 ); // we will have duplicated entries here!

    // (4) clear fanin counters
    // this is important for counters to be correctly set in the future iterations -- see step (0)
    // inlining of macro Gia_ManForEachObjVec
    for  i = 0; (i < Vec_IntSize(p.vFanins)) && ((pObj) = Gia_ManObj(p.pGia, Vec_IntEntry(p.vFanins,i))); i++  {
        Rnm_ObjSetCount( p, pObj, 0 );
    }

    // visualize
    if fVerbose != 0 {
        fmt.Printf("*** Refinement %3d : PI+PPI =%4d. Old =%4d. New =%4d.   FF =%4d. Just =%4d. Shared =%4d.\n", p.nRefId, Vec_IntSize(p.vMap), Vec_IntSize(vOldPPis), Vec_IntSize(vNewPPis), Counters[0], Counters[1], Counters[2]); }

//    Rnm_ManPrintSelected( p, vNewPPis );
//    Ga2_StructAnalize( p->pGia, p->vMap, p->vObjs, vNewPPis );
    return vNewPPis;
}

////////////////////////////////////////////////////////////////////////
///                       END OF FILE                                ///
////////////////////////////////////////////////////////////////////////


//ABC_NAMESPACE_IMPL_END

