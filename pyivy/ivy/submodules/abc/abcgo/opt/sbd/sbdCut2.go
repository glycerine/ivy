package sbd

import "fmt"

/**CFile****************************************************************

  FileName    [sbdCut2.c]

  SystemName  [ABC: Logic synthesis and verification system.]

  PackageName [SAT-based optimization using internal don't-cares.]

  Synopsis    [Cut computation.]

  Author      [Alan Mishchenko]
  
  Affiliation [UC Berkeley]

  Date        [Ver. 1.0. Started - June 20, 2005.]

  Revision    [$Id: sbdCut2.c,v 1.00 2005/06/20 00:00:00 alanmi Exp $]

***********************************************************************/

//#include "sbdInt.h"

//ABC_NAMESPACE_IMPL_START


////////////////////////////////////////////////////////////////////////
///                        DECLARATIONS                              ///
////////////////////////////////////////////////////////////////////////

const SBD_MAX_CUTSIZE =   10
const SBD_MAX_CUTNUM =    501

const SBD_CUT_NO_LEAF =   0xF

    
type Sbd_Cut_t struct {
    Sign            word;                      // signature
    iFunc             int32;                     // functionality
    Cost             int32;                      // cut cost
    CostLev             int32;                   // cut cost
    nTreeLeaves uint32; /* C bitfield: 9 bits; review packing and truncation */         // tree leaves
    nSlowLeaves uint32; /* C bitfield: 9 bits; review packing and truncation */         // slow leaves
    nTopLeaves uint32; /* C bitfield: 10 bits; review packing and truncation */         // top leaves
    nLeaves uint32; /* C bitfield: 4 bits; review packing and truncation */         // leaf count
    pLeaves [SBD_MAX_CUTSIZE]int32;  // leaves
}

type Sbd_Srv_t_ struct {
    nLutSize             int32;
    nCutSize             int32;
    nCutNum             int32;
    fVerbose             int32;
    pGia      *Gia_Man_t;                      // user's AIG manager (will be modified by adding nodes)
    vMirrors      *Vec_Int_t;                  // mirrors for each node
    vLutLevs      *Vec_Int_t;                  // delays for each node
    vLevs      *Vec_Int_t;                     // levels for each node
    vRefs      *Vec_Int_t;                     // refs for each node
    pCuts [SBD_MAX_CUTNUM]Sbd_Cut_t;     // temporary cuts
    ppCuts [SBD_MAX_CUTNUM]*Sbd_Cut_t;    // temporary cut pointers
    clkStart         abctime;                  // starting time
    vCut0      *Vec_Int_t;                     // current cut
    vCut      *Vec_Int_t;                      // current cut
    vCutTop      *Vec_Int_t;                   // current cut
    vCutBot      *Vec_Int_t;                   // current cut
}

////////////////////////////////////////////////////////////////////////
///                     FUNCTION DEFINITIONS                         ///
////////////////////////////////////////////////////////////////////////

/**Function*************************************************************

  Synopsis    []

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Sbd_ManCutServerStart(pGia *Gia_Man_t, vMirrors *Vec_Int_t, vLutLevs *Vec_Int_t, vLevs *Vec_Int_t, vRefs *Vec_Int_t, nLutSize int32, nCutSize int32, nCutNum int32, fVerbose int32) *Sbd_Srv_t {
    var p *Sbd_Srv_t;
    assert(nLutSize <= nCutSize, "nLutSize <= nCutSize");
    assert(nCutSize < SBD_CUT_NO_LEAF, "nCutSize < SBD_CUT_NO_LEAF");
    assert(nCutSize > 1 && nCutSize <= SBD_MAX_CUTSIZE, "nCutSize > 1 && nCutSize <= SBD_MAX_CUTSIZE");
    assert(nCutNum > 1 && nCutNum < SBD_MAX_CUTNUM, "nCutNum > 1 && nCutNum < SBD_MAX_CUTNUM");
    p = new(Sbd_Srv_t);
    p.clkStart = Abc_Clock();
    p.nLutSize = nLutSize;
    p.nCutSize = nCutSize;
    p.nCutNum  = nCutNum;
    p.fVerbose = fVerbose;
    p.pGia     = pGia;
    p.vMirrors = vMirrors;
    p.vLutLevs = vLutLevs;
    p.vLevs    = vLevs;
    p.vRefs    = vRefs;
    p.vCut0    = Vec_IntAlloc( 100 );
    p.vCut     = Vec_IntAlloc( 100 );
    p.vCutTop  = Vec_IntAlloc( 100 );
    p.vCutBot  = Vec_IntAlloc( 100 );
    return p;
}
func Sbd_ManCutServerStop(p *Sbd_Srv_t) {
    Vec_IntFree( p.vCut0 );
    Vec_IntFree( p.vCut );
    Vec_IntFree( p.vCutTop );
    Vec_IntFree( p.vCutBot );
    ABC_FREE( p );
}

/**Function*************************************************************

  Synopsis    []

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Sbd_ManCutIsTopo_rec(p *Gia_Man_t, vMirrors *Vec_Int_t, iObj int32) int32 {
    var pObj *Gia_Obj_t; 
    var Ret0, Ret1 int32;
    if Vec_IntEntry(vMirrors, iObj) >= 0 {
        iObj = Abc_Lit2Var(Vec_IntEntry(vMirrors, iObj)); }
    if !(iObj != 0) || Gia_ObjIsTravIdCurrentId(p, iObj) {
        return 1; }
    Gia_ObjSetTravIdCurrentId(p, iObj);
    pObj = Gia_ManObj( p, iObj );
    if Gia_ObjIsCi(pObj) {
        return 0; }
    assert(Gia_ObjIsAnd(pObj), "Gia_ObjIsAnd(pObj)");
    Ret0 = Sbd_ManCutIsTopo_rec( p, vMirrors, Gia_ObjFaninId0(pObj, iObj) );
    Ret1 = Sbd_ManCutIsTopo_rec( p, vMirrors, Gia_ObjFaninId1(pObj, iObj) );
    return Ret0 && Ret1;
}
func Sbd_ManCutIsTopo(p *Gia_Man_t, vMirrors *Vec_Int_t, vCut *Vec_Int_t, iObj int32) int32 {
    var i, Entry, RetValue int32;
    Gia_ManIncrementTravId( p );
    Vec_IntForEachEntry( vCut, Entry, i )
        Gia_ObjSetTravIdCurrentId( p, Entry );        
    RetValue = Sbd_ManCutIsTopo_rec( p, vMirrors, iObj );
    if RetValue == 0 {
        fmt.Printf("Cut of node %d is not tological\n", iObj); }
    assert((RetValue != 0), "RetValue");
    return RetValue;
}

/**Function*************************************************************

  Synopsis    []

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Sbd_ManCutExpandOne(p *Gia_Man_t, vMirrors *Vec_Int_t, vLutLevs *Vec_Int_t, vCut *Vec_Int_t, iThis int32, iObj int32) int32 {
    var Lit0m, Lit1m, Fan0, Fan1, iPlace0, iPlace1 int32;
    var LutLev int32  = Vec_IntEntry( vLutLevs, iObj );
    var pObj *Gia_Obj_t  = Gia_ManObj(p, iObj);
    if Gia_ObjIsCi(pObj) {
        return 0; }
    assert(Gia_ObjIsAnd(pObj), "Gia_ObjIsAnd(pObj)");
    Lit0m = Vec_IntEntry( vMirrors, Gia_ObjFaninId0(pObj, iObj) );
    Lit1m = Vec_IntEntry( vMirrors, Gia_ObjFaninId1(pObj, iObj) );
    if Lit0m >= 0 {
        Fan0 = Abc_Lit2Var(Lit0m);
    } else {
        Fan0 = Gia_ObjFaninId0(pObj, iObj);
    }
    if Lit1m >= 0 {
        Fan1 = Abc_Lit2Var(Lit1m);
    } else {
        Fan1 = Gia_ObjFaninId1(pObj, iObj);
    }
    iPlace0 = Vec_IntFind( vCut, Fan0 );
    iPlace1 = Vec_IntFind( vCut, Fan1 );
    if iPlace0 == -1 && iPlace1 == -1 {
        return 0; }
    if Vec_IntEntry(vLutLevs, Fan0) > LutLev || Vec_IntEntry(vLutLevs, Fan1) > LutLev {
        return 0; }
    Vec_IntDrop( vCut, iThis );
    if iPlace0 == -1 && Fan0 {
        Vec_IntPushOrder( vCut, Fan0 ); }
    if iPlace1 == -1 && (Fan1 != 0) {
        Vec_IntPushOrder( vCut, Fan1 ); }
    return 1;
}
func Vec_IntIsOrdered(vCut *Vec_Int_t) {
    var i, Prev, Entry int32;
    Prev = Vec_IntEntry( vCut, 0 );
    Vec_IntForEachEntryStart( vCut, Entry, i, 1 )
    {
        assert(Prev < Entry, "Prev < Entry");
        Prev = Entry;
    }
}
func Sbd_ManCutExpand(p *Gia_Man_t, vMirrors *Vec_Int_t, vLutLevs *Vec_Int_t, vCut *Vec_Int_t) {
    var i, Entry int32;
    for
    {
        Vec_IntForEachEntry( vCut, Entry, i )
            if ( Sbd_ManCutExpandOne( p, vMirrors, vLutLevs, vCut, i, Entry ) )
                break;
    
if !( i < Vec_IntSize(vCut) ) { break }
} 
    
}
func Sbd_ManCutReload(vMirrors *Vec_Int_t, vLutLevs *Vec_Int_t, LevStop int32, vCut *Vec_Int_t, vCutTop *Vec_Int_t, vCutBot *Vec_Int_t) {
    var i, Entry int32;
    Vec_IntClear( vCutTop );
    Vec_IntClear( vCutBot );
    Vec_IntForEachEntry( vCut, Entry, i )
    {
        assert((Entry != 0), "Entry");
        assert(Vec_IntEntry(vMirrors, Entry) == -1, "Vec_IntEntry(vMirrors, Entry) == -1");
        assert(Vec_IntEntry(vLutLevs, Entry) <= LevStop, "Vec_IntEntry(vLutLevs, Entry) <= LevStop");
        if Vec_IntEntry(vLutLevs, Entry) == LevStop {
            Vec_IntPush( vCutTop, Entry );
        } else {
            Vec_IntPush( vCutBot, Entry ); }
    }
    Vec_IntIsOrdered( vCut );
}
func Sbd_ManCutCollect_rec(p *Gia_Man_t, vMirrors *Vec_Int_t, iObj int32, LevStop int32, vLutLevs *Vec_Int_t, vCut *Vec_Int_t) int32 {
    var pObj *Gia_Obj_t; 
    var Ret0, Ret1 int32;
    if Vec_IntEntry(vMirrors, iObj) >= 0 {
        iObj = Abc_Lit2Var(Vec_IntEntry(vMirrors, iObj)); }
    if !(iObj != 0) || Gia_ObjIsTravIdCurrentId(p, iObj) {
        return 1; }
    Gia_ObjSetTravIdCurrentId(p, iObj);
    pObj = Gia_ManObj( p, iObj );
    if Gia_ObjIsCi(pObj) || Vec_IntEntry(vLutLevs, iObj) <= LevStop {
        Vec_IntPush( vCut, iObj );
        return Vec_IntEntry(vLutLevs, iObj) <= LevStop;
    }
    assert(Gia_ObjIsAnd(pObj), "Gia_ObjIsAnd(pObj)");
    Ret0 = Sbd_ManCutCollect_rec( p, vMirrors, Gia_ObjFaninId0(pObj, iObj), LevStop, vLutLevs, vCut );
    Ret1 = Sbd_ManCutCollect_rec( p, vMirrors, Gia_ObjFaninId1(pObj, iObj), LevStop, vLutLevs, vCut );
    return Ret0 && Ret1;
}

/**Function*************************************************************

  Synopsis    []

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Sbd_ManCutReduceTop(p *Gia_Man_t, vMirrors *Vec_Int_t, iObj int32, vLutLevs *Vec_Int_t, vCut *Vec_Int_t, vCutTop *Vec_Int_t, nCutSize int32) int32 {
    var i, Entry, Lit0m, Lit1m, Fan0, Fan1 int32;
    var LevStop int32  = Vec_IntEntry(vLutLevs, iObj) - 2;
    Vec_IntIsOrdered( vCut );
    Vec_IntForEachEntryReverse( vCutTop, Entry, i )
    {
        var pObj *Gia_Obj_t  = Gia_ManObj( p, Entry );
        if Gia_ObjIsCi(pObj) {
            continue; }
        assert(Gia_ObjIsAnd(pObj), "Gia_ObjIsAnd(pObj)");
        assert(Vec_IntEntry(vLutLevs, Entry) == LevStop, "Vec_IntEntry(vLutLevs, Entry) == LevStop");
        Lit0m = Vec_IntEntry( vMirrors, Gia_ObjFaninId0(pObj, Entry) );
        Lit1m = Vec_IntEntry( vMirrors, Gia_ObjFaninId1(pObj, Entry) );
        if Lit0m >= 0 {
            Fan0 = Abc_Lit2Var(Lit0m);
        } else {
            Fan0 = Gia_ObjFaninId0(pObj, Entry);
        }
        if Lit1m >= 0 {
            Fan1 = Abc_Lit2Var(Lit1m);
        } else {
            Fan1 = Gia_ObjFaninId1(pObj, Entry);
        }
        if Vec_IntEntry(vLutLevs, Fan0) > LevStop || Vec_IntEntry(vLutLevs, Fan1) > LevStop {
            continue; }
        assert(Vec_IntEntry(vLutLevs, Fan0) <= LevStop, "Vec_IntEntry(vLutLevs, Fan0) <= LevStop");
        assert(Vec_IntEntry(vLutLevs, Fan1) <= LevStop, "Vec_IntEntry(vLutLevs, Fan1) <= LevStop");
        if Vec_IntEntry(vLutLevs, Fan0) == LevStop && Vec_IntEntry(vLutLevs, Fan1) == LevStop {
            continue; }
        Vec_IntRemove( vCut, Entry );
        if Fan0 { Vec_IntPushUniqueOrder( vCut, Fan0 ); }
        if Fan1 != 0 { Vec_IntPushUniqueOrder( vCut, Fan1 ); }
        //Sbd_ManCutIsTopo( p, vMirrors, vCut, iObj );
        return 1;
    }
    return 0;
}

/**Function*************************************************************

  Synopsis    []

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Sbd_ManCutServerFirst(p *Sbd_Srv_t, iObj int32, pLeaves *int32) int32 {
    var RetValue int32;
    var LevStop int32 = Vec_IntEntry(p.vLutLevs, iObj) - 2;

    Vec_IntClear( p.vCut );
    Gia_ManIncrementTravId( p.pGia );
    RetValue = Sbd_ManCutCollect_rec( p.pGia, p.vMirrors, iObj, LevStop, p.vLutLevs, p.vCut );
    if RetValue == 0 { // cannot build delay-improving cut
        return -1; }
    // check if the current cut is good
    Vec_IntSort( p.vCut, 0 );
/*
    Sbd_ManCutReload( p->vMirrors, p->vLutLevs, LevStop, p->vCut, p->vCutTop, p->vCutBot );
    if ( Vec_IntSize(p->vCut) <= p->nCutSize && Vec_IntSize(p->vCutTop) <= p->nLutSize-1 )
    {
        //printf( "%d ", Vec_IntSize(p->vCut) );
        memcpy( pLeaves, Vec_IntArray(p->vCut), sizeof(int) * Vec_IntSize(p->vCut) );
        return Vec_IntSize(p->vCut);
    }
*/
    // try to expand the cut
    Sbd_ManCutExpand( p.pGia, p.vMirrors, p.vLutLevs, p.vCut );
    Sbd_ManCutReload( p.vMirrors, p.vLutLevs, LevStop, p.vCut, p.vCutTop, p.vCutBot );
    if Vec_IntSize(p.vCut) <= p.nCutSize && Vec_IntSize(p.vCutTop) <= p.nLutSize-1 {
        //printf( "1=(%d,%d) ", Vec_IntSize(p->vCutTop), Vec_IntSize(p->vCutBot) );
        //printf( "%d ", Vec_IntSize(p->vCut) );
        memcpy( pLeaves, Vec_IntArray(p.vCut), sizeof(int) * Vec_IntSize(p.vCut) );
        return Vec_IntSize(p.vCut);
    }

    // try to reduce the topmost
    Vec_IntClear( p.vCut0 );
    Vec_IntAppend( p.vCut0, p.vCut );
    if Vec_IntSize(p.vCut) < p.nCutSize && (Sbd_ManCutReduceTop( p.pGia, p.vMirrors, iObj, p.vLutLevs, p.vCut, p.vCutTop, p.nCutSize ) != 0) {
        Sbd_ManCutExpand( p.pGia, p.vMirrors, p.vLutLevs, p.vCut );
        Sbd_ManCutReload( p.vMirrors, p.vLutLevs, LevStop, p.vCut, p.vCutTop, p.vCutBot );
        assert(Vec_IntSize(p.vCut) <= p.nCutSize, "Vec_IntSize(p.vCut) <= p.nCutSize");
        if Vec_IntSize(p.vCutTop) <= p.nLutSize-1 {
            //printf( "%d -> %d (%d + %d)\n", Vec_IntSize(p->vCut0), Vec_IntSize(p->vCut), Vec_IntSize(p->vCutTop), Vec_IntSize(p->vCutBot) );
            memcpy( pLeaves, Vec_IntArray(p.vCut), sizeof(int) * Vec_IntSize(p.vCut) );
            return Vec_IntSize(p.vCut);
        }
        // try again
        if Vec_IntSize(p.vCut) < p.nCutSize && (Sbd_ManCutReduceTop( p.pGia, p.vMirrors, iObj, p.vLutLevs, p.vCut, p.vCutTop, p.nCutSize ) != 0) {
            Sbd_ManCutExpand( p.pGia, p.vMirrors, p.vLutLevs, p.vCut );
            Sbd_ManCutReload( p.vMirrors, p.vLutLevs, LevStop, p.vCut, p.vCutTop, p.vCutBot );
            assert(Vec_IntSize(p.vCut) <= p.nCutSize, "Vec_IntSize(p.vCut) <= p.nCutSize");
            if Vec_IntSize(p.vCutTop) <= p.nLutSize-1 {
                //printf( "* %d -> %d (%d + %d)\n", Vec_IntSize(p->vCut0), Vec_IntSize(p->vCut), Vec_IntSize(p->vCutTop), Vec_IntSize(p->vCutBot) );
                memcpy( pLeaves, Vec_IntArray(p.vCut), sizeof(int) * Vec_IntSize(p.vCut) );
                return Vec_IntSize(p.vCut);
            }
            // try again
            if Vec_IntSize(p.vCut) < p.nCutSize && (Sbd_ManCutReduceTop( p.pGia, p.vMirrors, iObj, p.vLutLevs, p.vCut, p.vCutTop, p.nCutSize ) != 0) {
                Sbd_ManCutExpand( p.pGia, p.vMirrors, p.vLutLevs, p.vCut );
                Sbd_ManCutReload( p.vMirrors, p.vLutLevs, LevStop, p.vCut, p.vCutTop, p.vCutBot );
                assert(Vec_IntSize(p.vCut) <= p.nCutSize, "Vec_IntSize(p.vCut) <= p.nCutSize");
                if Vec_IntSize(p.vCutTop) <= p.nLutSize-1 {
                    //printf( "** %d -> %d (%d + %d)\n", Vec_IntSize(p->vCut0), Vec_IntSize(p->vCut), Vec_IntSize(p->vCutTop), Vec_IntSize(p->vCutBot) );
                    memcpy( pLeaves, Vec_IntArray(p.vCut), sizeof(int) * Vec_IntSize(p.vCut) );
                    return Vec_IntSize(p.vCut);
                }
                // try again
                if Vec_IntSize(p.vCut) < p.nCutSize && (Sbd_ManCutReduceTop( p.pGia, p.vMirrors, iObj, p.vLutLevs, p.vCut, p.vCutTop, p.nCutSize ) != 0) {
                    Sbd_ManCutExpand( p.pGia, p.vMirrors, p.vLutLevs, p.vCut );
                    Sbd_ManCutReload( p.vMirrors, p.vLutLevs, LevStop, p.vCut, p.vCutTop, p.vCutBot );
                    assert(Vec_IntSize(p.vCut) <= p.nCutSize, "Vec_IntSize(p.vCut) <= p.nCutSize");
                    if Vec_IntSize(p.vCutTop) <= p.nLutSize-1 {
                        //printf( "*** %d -> %d (%d + %d)\n", Vec_IntSize(p->vCut0), Vec_IntSize(p->vCut), Vec_IntSize(p->vCutTop), Vec_IntSize(p->vCutBot) );
                        memcpy( pLeaves, Vec_IntArray(p.vCut), sizeof(int) * Vec_IntSize(p.vCut) );
                        return Vec_IntSize(p.vCut);
                    }
                }
            }
        }
    }

    // recompute the cut
    Vec_IntClear( p.vCut );
    Gia_ManIncrementTravId( p.pGia );
    RetValue = Sbd_ManCutCollect_rec( p.pGia, p.vMirrors, iObj, LevStop-1, p.vLutLevs, p.vCut );
    if RetValue == 0 { // cannot build delay-improving cut
        return -1; }
    // check if the current cut is good
    Vec_IntSort( p.vCut, 0 );
/*
    Sbd_ManCutReload( p->vMirrors, p->vLutLevs, LevStop, p->vCut, p->vCutTop, p->vCutBot );
    if ( Vec_IntSize(p->vCut) <= p->nCutSize && Vec_IntSize(p->vCutTop) <= p->nLutSize-1 )
    {
        //printf( "%d ", Vec_IntSize(p->vCut) );
        memcpy( pLeaves, Vec_IntArray(p->vCut), sizeof(int) * Vec_IntSize(p->vCut) );
        return Vec_IntSize(p->vCut);
    }
*/
    // try to expand the cut
    Sbd_ManCutExpand( p.pGia, p.vMirrors, p.vLutLevs, p.vCut );
    Sbd_ManCutReload( p.vMirrors, p.vLutLevs, LevStop, p.vCut, p.vCutTop, p.vCutBot );
    if Vec_IntSize(p.vCut) <= p.nCutSize && Vec_IntSize(p.vCutTop) <= p.nLutSize-1 {
        //printf( "2=(%d,%d) ", Vec_IntSize(p->vCutTop), Vec_IntSize(p->vCutBot) );
        //printf( "%d ", Vec_IntSize(p->vCut) );
        memcpy( pLeaves, Vec_IntArray(p.vCut), sizeof(int) * Vec_IntSize(p.vCut) );
        return Vec_IntSize(p.vCut);
    }

    return -1;
}

////////////////////////////////////////////////////////////////////////
///                       END OF FILE                                ///
////////////////////////////////////////////////////////////////////////


//ABC_NAMESPACE_IMPL_END

