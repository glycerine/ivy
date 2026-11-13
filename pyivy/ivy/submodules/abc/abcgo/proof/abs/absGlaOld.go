package abs

import (
	"fmt"
	"os"
	"time"
)

/**CFile****************************************************************

  FileName    [absGla.c]

  SystemName  [ABC: Logic synthesis and verification system.]

  PackageName [Abstraction package.]

  Synopsis    [Gate-level abstraction.]

  Author      [Alan Mishchenko]
  
  Affiliation [UC Berkeley]

  Date        [Ver. 1.0. Started - June 20, 2005.]

  Revision    [$Id: absGla.c,v 1.00 2005/06/20 00:00:00 alanmi Exp $]

***********************************************************************/

//#include "sat/cnf/cnf.h"
//#include "sat/bsat/satSolver2.h"
//#include "base/main/main.h"
//#include "abs.h"
//#include "absRef.h"

//ABC_NAMESPACE_IMPL_START

////////////////////////////////////////////////////////////////////////
///                        DECLARATIONS                              ///
////////////////////////////////////////////////////////////////////////

    // refinement object
type Rfn_Obj_t struct {
    Value uint32; /* C bitfield: 1 bits; review packing and truncation */  // value
    fVisit uint32; /* C bitfield: 1 bits; review packing and truncation */  // visited
    fPPi uint32; /* C bitfield: 1 bits; review packing and truncation */  // PPI
    Prio uint32; /* C bitfield: 16 bits; review packing and truncation */  // priority
    Sign uint32; /* C bitfield: 12 bits; review packing and truncation */  // traversal signature
}

    // abstraction object
type Gla_Obj_t struct {
    iGiaObj            int32;         // corresponding GIA obj
    fAbs uint32; /* C bitfield: 1 bits; review packing and truncation */  // belongs to abstraction
    fCompl0 uint32; /* C bitfield: 1 bits; review packing and truncation */  // compl bit of the first fanin
    fConst uint32; /* C bitfield: 1 bits; review packing and truncation */  // object attribute
    fPi uint32; /* C bitfield: 1 bits; review packing and truncation */  // object attribute
    fPo uint32; /* C bitfield: 1 bits; review packing and truncation */  // object attribute
    fRo uint32; /* C bitfield: 1 bits; review packing and truncation */  // object attribute
    fRi uint32; /* C bitfield: 1 bits; review packing and truncation */  // object attribute
    fAnd uint32; /* C bitfield: 1 bits; review packing and truncation */  // object attribute
    fMark uint32; /* C bitfield: 1 bits; review packing and truncation */  // nearby object
    nFanins uint32; /* C bitfield: 23 bits; review packing and truncation */  // fanin count
    Fanins [4]int32;       // fanins
    vFrames      Vec_Int_t;         // variables in each timeframe
}

    // manager
type Gla_Man_t struct {
    // user data
    pGia0     *Gia_Man_t;           // starting AIG manager
    pGia     *Gia_Man_t;            // working AIG manager
    pPars     *Abs_Par_t;           // parameters
    // internal data
    vAbs     *Vec_Int_t;            // abstracted objects
    pObjRoot     *Gla_Obj_t;        // the primary output
    pObjs     *Gla_Obj_t;           // objects
    pObj2Obj      *uint32;        // mapping of GIA obj into GLA obj
    nObjs            int32;           // the number of objects
    nAbsOld            int32;         // previous abstraction
//    int            nAbsNew;         // previous abstraction
//    int            nLrnOld;         // the number of bytes
//    int            nLrnNew;         // the number of bytes
    // other data
    nCexes            int32;          // the number of counter-examples
    nObjAdded            int32;       // total number of objects added
    nSatVars            int32;        // the number of SAT variables
    pCnf     *Cnf_Dat_t;            // CNF derived for the nodes
    pSat   *sat_solver2;            // incremental SAT solver
    vTemp     *Vec_Int_t;           // temporary array
    vAddedNew     *Vec_Int_t;       // temporary array
    vObjCounts     *Vec_Int_t;      // object counters
    vCoreCounts     *Vec_Int_t;     // counts how many times each object appears in the core
    vProofIds     *Vec_Int_t;       // counts how many times each object appears in the core
    nProofIds            int32;       // proof ID counter
    // refinement
    pvRefis     *Vec_Int_t;         // vectors of each object
    // refinement manager
    pGia2     *Gia_Man_t;
    pRnm     *Rnm_Man_t;
    // statistics  
    timeInit        abctime;
    timeSat        abctime;
    timeUnsat        abctime;
    timeCex        abctime;
    timeOther        abctime;
}

// declarations
// static Vec_Int_t * Gla_ManCollectPPis( Gla_Man_t * p, Vec_Int_t * vPis );

// static int         Gla_ManCheckVar( Gla_Man_t * p, int iObj, int iFrame );

// static int         Gla_ManGetVar( Gla_Man_t * p, int iObj, int iFrame );


// object procedures
func Gla_ObjId(p *Gla_Man_t, pObj *Gla_Obj_t) int32 { assert(pObj > p.pObjs && pObj < p.pObjs + p.nObjs, "pObj > p.pObjs && pObj < p.pObjs + p.nObjs"); return pObj - p.pObjs;    }
func Gla_ManObj(p *Gla_Man_t, i int32) *Gla_Obj_t { assert(i >= 0 && i < p.nObjs, "i >= 0 && i < p.nObjs"); if i != 0 {
    return p.pObjs + i;
} else {
    return nil;
}                   }
func Gla_ManGiaObj(p *Gla_Man_t, pObj *Gla_Obj_t) *Gia_Obj_t { return Gia_ManObj( p.pGia, pObj.iGiaObj );                                        }
func Gla_ObjSatValue(p *Gla_Man_t, iGia int32, f int32) int32 { if Gla_ManCheckVar(p, p.pObj2Obj[iGia], f) != 0 {
    return sat_solver2_var_value( p.pSat, Gla_ManGetVar(p, p.pObj2Obj[iGia], f) );
} else {
    return 0;
}  }

func Gla_ObjRef(p *Gla_Man_t, pObj *Gia_Obj_t, f int32) *Rfn_Obj_t { return (Rfn_Obj_t *)Vec_IntGetEntryP( &(p.pvRefis[Gia_ObjId(p.pGia, pObj)]), f ); }
func Gla_ObjClearRef(p *Rfn_Obj_t) { *((int *)p) = 0;                                                                    }


// iterator through abstracted objects
// Original macro:
// #define Gla_ManForEachObj( p, pObj )                       \
//     for ( pObj = p->pObjs + 1; pObj < p->pObjs + p->nObjs; pObj++ ) 

// Original macro:
// #define Gla_ManForEachObjAbs( p, pObj, i )                 \
//     for ( i = 0; i < Vec_IntSize(p->vAbs) && ((pObj = Gla_ManObj(p, Vec_IntEntry(p->vAbs, i))),1); i++) 

// Original macro:
// #define Gla_ManForEachObjAbsVec( vVec, p, pObj, i )        \
//     for ( i = 0; i < Vec_IntSize(vVec) && ((pObj = Gla_ManObj(p, Vec_IntEntry(vVec, i))),1); i++) 


// iterator through fanins of an abstracted object
// Original macro:
// #define Gla_ObjForEachFanin( p, pObj, pFanin, i )          \
//     for ( i = 0; (i < (int)pObj->nFanins) && ((pFanin = Gla_ManObj(p, pObj->Fanins[i])),1); i++ )


// some lessons learned from debugging mismatches between GIA and mapped CNF
// - inputs/output of AND-node may be PPIs (have SAT vars), but the node is not included in the abstraction
// - some logic node can be a PPI of one LUT and an internal node of another LUT

////////////////////////////////////////////////////////////////////////
///                     FUNCTION DEFINITIONS                         ///
////////////////////////////////////////////////////////////////////////

/**Function*************************************************************

  Synopsis    [Prepares CEX and vMap for refinement.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Gia_GlaPrepareCexAndMap(p *Gla_Man_t, ppCex **Abc_Cex_t, pvMap **Vec_Int_t) {
    var pCex *Abc_Cex_t;
    var vMap *Vec_Int_t;
    var pObj, pFanin *Gla_Obj_t;
    var pGiaObj *Gia_Obj_t;
    var f, i, k int32;
    // find PIs and PPIs
    vMap = Vec_IntAlloc( 1000 );
    // inlining of macro Gla_ManForEachObjAbs
    for  i = 0; i < Vec_IntSize(p.vAbs) && ((pObj = Gla_ManObj(p, Vec_IntEntry(p.vAbs, i))),1); i++ {
        assert(pObj.fConst || pObj.fRo || pObj.fAnd, "pObj.fConst || pObj.fRo || pObj.fAnd");
        // inlining of macro Gla_ObjForEachFanin
        for  k = 0; (k < int32(pObj.nFanins)) && ((pFanin = Gla_ManObj(p, pObj.Fanins[k])),1); k++  {
            if !pFanin.fAbs {
                Vec_IntPush( vMap, pFanin.iGiaObj ); }
        }
    }
    Vec_IntUniqify( vMap );
    // derive counter-example
    pCex = Abc_CexAlloc( 0, Vec_IntSize(vMap), p.pPars.iFrame+1 );
    pCex.iFrame = p.pPars.iFrame;
    for  f = 0; f <= p.pPars.iFrame; f++  {
        // inlining of macro Gia_ManForEachObjVec
        for  k = 0; (k < Vec_IntSize(vMap)) && ((pGiaObj) = Gia_ManObj(p.pGia, Vec_IntEntry(vMap,k))); k++  {
            if Gla_ObjSatValue( p, Gia_ObjId(p.pGia, pGiaObj), f ) != 0 {
                Abc_InfoSetBit( pCex.pData, f * Vec_IntSize(vMap) + k ); }
        }
    }
    *pvMap = vMap;
    *ppCex = pCex;
}

/**Function*************************************************************

  Synopsis    [Derives counter-example using current assignments.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Gla_ManDeriveCex(p *Gla_Man_t, vPis *Vec_Int_t) *Abc_Cex_t {
    var pCex *Abc_Cex_t;
    var pObj *Gia_Obj_t;
    var i, f int32;
    pCex = Abc_CexAlloc( Gia_ManRegNum(p.pGia), Gia_ManPiNum(p.pGia), p.pPars.iFrame+1 );
    pCex.iPo = 0;
    pCex.iFrame = p.pPars.iFrame;
    // inlining of macro Gia_ManForEachObjVec
    for  i = 0; (i < Vec_IntSize(vPis)) && ((pObj) = Gia_ManObj(p.pGia, Vec_IntEntry(vPis,i))); i++  {
        if !Gia_ObjIsPi(p.pGia, pObj) {
            continue; }
        assert(Gia_ObjIsPi(p.pGia, pObj), "Gia_ObjIsPi(p.pGia, pObj)");
        for  f = 0; f <= pCex.iFrame; f++  {
            if Gla_ObjSatValue( p, Gia_ObjId(p.pGia, pObj), f ) != 0 {
                Abc_InfoSetBit( pCex.pData, pCex.nRegs + f * pCex.nPis + Gia_ObjCioId(pObj) ); }
        }
    }
    return pCex;
}

/**Function*************************************************************

  Synopsis    [Collects GIA abstraction objects.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Gla_ManCollectInternal_rec(p *Gia_Man_t, pGiaObj *Gia_Obj_t, vRoAnds *Vec_Int_t) {
    if Gia_ObjIsTravIdCurrent(p, pGiaObj) {
        return; }
    Gia_ObjSetTravIdCurrent(p, pGiaObj);
    assert(Gia_ObjIsAnd(pGiaObj), "Gia_ObjIsAnd(pGiaObj)");
    Gla_ManCollectInternal_rec( p, Gia_ObjFanin0(pGiaObj), vRoAnds );
    Gla_ManCollectInternal_rec( p, Gia_ObjFanin1(pGiaObj), vRoAnds );
    Vec_IntPush( vRoAnds, Gia_ObjId(p, pGiaObj) );
}
func Gla_ManCollect(p *Gla_Man_t, vPis *Vec_Int_t, vPPis *Vec_Int_t, vCos *Vec_Int_t, vRoAnds *Vec_Int_t) {
    var pObj, pFanin *Gla_Obj_t; 
    var pGiaObj *Gia_Obj_t;
    var i, k int32;

    // collect COs
    Vec_IntPush( vCos, Gia_ObjId(p.pGia, Gia_ManPo(p.pGia, 0)) );
    // collect fanins of abstracted objects
    // inlining of macro Gla_ManForEachObjAbs
    for  i = 0; i < Vec_IntSize(p.vAbs) && ((pObj = Gla_ManObj(p, Vec_IntEntry(p.vAbs, i))),1); i++ {
        assert(pObj.fConst || pObj.fRo || pObj.fAnd, "pObj.fConst || pObj.fRo || pObj.fAnd");
        if pObj.fRo {
            pGiaObj = Gia_ObjRoToRi( p.pGia, Gia_ManObj(p.pGia, pObj.iGiaObj) );
            Vec_IntPush( vCos, Gia_ObjId(p.pGia, pGiaObj) );
        }
        // inlining of macro Gla_ObjForEachFanin
        for  k = 0; (k < int32(pObj.nFanins)) && ((pFanin = Gla_ManObj(p, pObj.Fanins[k])),1); k++  {
            if !pFanin.fAbs {
                Vec_IntPush( (pFanin.fPi ? vPis : vPPis), pFanin.iGiaObj ); }
        }
    }
    Vec_IntUniqify( vPis );
    Vec_IntUniqify( vPPis );
    Vec_IntSort( vCos, 0 );
    // sorting PIs/PPIs/COs leads to refinements that are more "well-aligned"...

    // mark const/PIs/PPIs
    Gia_ManIncrementTravId( p.pGia );
    Gia_ObjSetTravIdCurrent( p.pGia, Gia_ManConst0(p.pGia) );
    // inlining of macro Gia_ManForEachObjVec
    for  i = 0; (i < Vec_IntSize(vPis)) && ((pGiaObj) = Gia_ManObj(p.pGia, Vec_IntEntry(vPis,i))); i++  {
        Gia_ObjSetTravIdCurrent( p.pGia, pGiaObj );
    }
    // inlining of macro Gia_ManForEachObjVec
    for  i = 0; (i < Vec_IntSize(vPPis)) && ((pGiaObj) = Gia_ManObj(p.pGia, Vec_IntEntry(vPPis,i))); i++  {
        Gia_ObjSetTravIdCurrent( p.pGia, pGiaObj );
    }
    // mark and add ROs first
    // inlining of macro Gia_ManForEachObjVec
    for  i = 0; (i < Vec_IntSize(vCos)) && ((pGiaObj) = Gia_ManObj(p.pGia, Vec_IntEntry(vCos,i))); i++  {
        if i == 0 { continue; }
        pGiaObj = Gia_ObjRiToRo( p.pGia, pGiaObj );
        Gia_ObjSetTravIdCurrent( p.pGia, pGiaObj );
        Vec_IntPush( vRoAnds, Gia_ObjId(p.pGia, pGiaObj) );
    }
    // collect nodes between PIs/PPIs/ROs and COs
    // inlining of macro Gia_ManForEachObjVec
    for  i = 0; (i < Vec_IntSize(vCos)) && ((pGiaObj) = Gia_ManObj(p.pGia, Vec_IntEntry(vCos,i))); i++  {
        Gla_ManCollectInternal_rec( p.pGia, Gia_ObjFanin0(pGiaObj), vRoAnds );
    }
}

/**Function*************************************************************

  Synopsis    [Drive implications of the given node towards primary outputs.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Gia_ManRefSetAndPropFanout_rec(p *Gla_Man_t, pObj *Gia_Obj_t, f int32, vSelect *Vec_Int_t, Sign int32) {
    var i int32;//, Id = Gia_ObjId(p->pGia, pObj);
    var pRef0 *Rfn_Obj_t;
    var pRef1 *Rfn_Obj_t;
    var pRef *Rfn_Obj_t = Gla_ObjRef( p, pObj, f );
    var pFanout *Gia_Obj_t;
    var k int32;
    if int32(pRef.Sign) != Sign {
        return; }
    assert(pRef.fVisit == 0, "pRef.fVisit == 0");
    pRef.fVisit = 1;
    if pRef.fPPi {
        assert(int32(pRef.Prio) > 0, "int32(pRef.Prio) > 0");
        for  i = p.pPars.iFrame; i >= 0; i--  {
            if !Gla_ObjRef(p, pObj, i).fVisit {
                Gia_ManRefSetAndPropFanout_rec( p, pObj, i, vSelect, Sign ); }
        }
        Vec_IntPush( vSelect, Gia_ObjId(p.pGia, pObj) );
        return;
    }
    if (Gia_ObjIsCo(pObj) && f == p.pPars.iFrame) || Gia_ObjIsPo(p.pGia, pObj) {
        return; }
    if Gia_ObjIsRi(p.pGia, pObj) {
        pFanout = Gia_ObjRiToRo(p.pGia, pObj);
        if !Gla_ObjRef(p, pFanout, f+1).fVisit {
            Gia_ManRefSetAndPropFanout_rec( p, pFanout, f+1, vSelect, Sign ); }
        return;
    }
    assert(Gia_ObjIsRo(p.pGia, pObj) || Gia_ObjIsAnd(pObj), "Gia_ObjIsRo(p.pGia, pObj) || Gia_ObjIsAnd(pObj)");
    // inlining of macro Gia_ObjForEachFanoutStatic
    for  k = 0; (k < Gia_ObjFanoutNum(p.pGia, pObj))   && (((pFanout) = Gia_ObjFanout(p.pGia, pObj, k)), 1); k++  {
//        Rfn_Obj_t * pRefF = Gla_ObjRef(p, pFanout, f);
        if Gla_ObjRef(p, pFanout, f).fVisit {
            continue; }
        if Gia_ObjIsCo(pFanout) {
            Gia_ManRefSetAndPropFanout_rec( p, pFanout, f, vSelect, Sign );
            continue;
        } 
        assert(Gia_ObjIsAnd(pFanout), "Gia_ObjIsAnd(pFanout)");
        pRef0 = Gla_ObjRef( p, Gia_ObjFanin0(pFanout), f );
        pRef1 = Gla_ObjRef( p, Gia_ObjFanin1(pFanout), f );
        if ((pRef0.Value ^ Gia_ObjFaninC0(pFanout)) == 0 && pRef0.fVisit) ||
             ((pRef1.Value ^ Gia_ObjFaninC1(pFanout)) == 0 && pRef1.fVisit) || 
           ( ((pRef0.Value ^ Gia_ObjFaninC0(pFanout)) == 1 && pRef0.fVisit) && 
             ((pRef1.Value ^ Gia_ObjFaninC1(pFanout)) == 1 && pRef1.fVisit) ) {
           Gia_ManRefSetAndPropFanout_rec( p, pFanout, f, vSelect, Sign ); }
    }
}
 
/**Function*************************************************************

  Synopsis    [Selects assignments to be refined.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Gla_ManRefSelect_rec(p *Gla_Man_t, pObj *Gia_Obj_t, f int32, vSelect *Vec_Int_t, Sign int32) { 
    var i int32;//, Id = Gia_ObjId(p->pGia, pObj);
    var pRef *Rfn_Obj_t  = Gla_ObjRef( p, pObj, f );
//    assert( (int)pRef->Sign == Sign );
    if pRef.fVisit {
        return; }
    if p.pPars.fPropFanout {
        Gia_ManRefSetAndPropFanout_rec( p, pObj, f, vSelect, Sign );
    } else {
        pRef.fVisit = 1; }
    if pRef.fPPi {
        assert(int32(pRef.Prio) > 0, "int32(pRef.Prio) > 0");
        if p.pPars.fPropFanout {
            for  i = p.pPars.iFrame; i >= 0; i--  {
                if !Gla_ObjRef(p, pObj, i).fVisit {
                    Gia_ManRefSetAndPropFanout_rec( p, pObj, i, vSelect, Sign ); }
            }
        
        } else {
            Vec_IntPush( vSelect, Gia_ObjId(p.pGia, pObj) );
            Vec_IntAddToEntry( p.vObjCounts, f, 1 );
        }
        return;
    }
    if Gia_ObjIsPi(p.pGia, pObj) || Gia_ObjIsConst0(pObj) {
        return; }
    if Gia_ObjIsRo(p.pGia, pObj) {
        if f > 0 {
            Gla_ManRefSelect_rec( p, Gia_ObjFanin0(Gia_ObjRoToRi(p.pGia, pObj)), f-1, vSelect, Sign ); }
        return;
    }
    if Gia_ObjIsAnd(pObj) {
        var pRef0 *Rfn_Obj_t  = Gla_ObjRef( p, Gia_ObjFanin0(pObj), f );
        var pRef1 *Rfn_Obj_t  = Gla_ObjRef( p, Gia_ObjFanin1(pObj), f );
        if pRef.Value == 1 {
            if pRef0.Prio > 0 {
                Gla_ManRefSelect_rec( p, Gia_ObjFanin0(pObj), f, vSelect, Sign ); }
            if pRef1.Prio > 0 {
                Gla_ManRefSelect_rec( p, Gia_ObjFanin1(pObj), f, vSelect, Sign ); }
        
        } else { // select one value
        
            if (pRef0.Value ^ Gia_ObjFaninC0(pObj)) == 0 && (pRef1.Value ^ Gia_ObjFaninC1(pObj)) == 0 {
                if pRef0.Prio <= pRef1.Prio { // choice
                
                    if pRef0.Prio > 0 {
                        Gla_ManRefSelect_rec( p, Gia_ObjFanin0(pObj), f, vSelect, Sign ); }
                
                } else {
                    if pRef1.Prio > 0 {
                        Gla_ManRefSelect_rec( p, Gia_ObjFanin1(pObj), f, vSelect, Sign ); }
                }
            
            } else if (pRef0.Value ^ Gia_ObjFaninC0(pObj)) == 0 {
                if pRef0.Prio > 0 {
                    Gla_ManRefSelect_rec( p, Gia_ObjFanin0(pObj), f, vSelect, Sign ); }
            
            } else if (pRef1.Value ^ Gia_ObjFaninC1(pObj)) == 0 {
                if pRef1.Prio > 0 {
                    Gla_ManRefSelect_rec( p, Gia_ObjFanin1(pObj), f, vSelect, Sign ); }
            
            } else { assert(false, "0"); }
        }
    
    } else { assert(false, "0"); }
}


/**Function*************************************************************

  Synopsis    [Performs refinement.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Gla_ManVerifyUsingTerSim(p *Gla_Man_t, vPis *Vec_Int_t, vPPis *Vec_Int_t, vRoAnds *Vec_Int_t, vCos *Vec_Int_t, vRes *Vec_Int_t) {
    var pObj *Gia_Obj_t;
    var i, f int32;
//    Gia_ManForEachObj( p->pGia, pObj, i )
//        assert( Gia_ObjTerSimGetC(pObj) );
    for  f = 0; f <= p.pPars.iFrame; f++  {
        Gia_ObjTerSimSet0( Gia_ManConst0(p.pGia) );
        // inlining of macro Gia_ManForEachObjVec
        for  i = 0; (i < Vec_IntSize(vPis)) && ((pObj) = Gia_ManObj(p.pGia, Vec_IntEntry(vPis,i))); i++  {
            if Gla_ObjSatValue( p, Gia_ObjId(p.pGia, pObj), f ) != 0 {
                Gia_ObjTerSimSet1( pObj );
            } else {
                Gia_ObjTerSimSet0( pObj ); }
        }
        // inlining of macro Gia_ManForEachObjVec
        for  i = 0; (i < Vec_IntSize(vPPis)) && ((pObj) = Gia_ManObj(p.pGia, Vec_IntEntry(vPPis,i))); i++  {
            Gia_ObjTerSimSetX( pObj );
        }
        // inlining of macro Gia_ManForEachObjVec
        for  i = 0; (i < Vec_IntSize(vRes)) && ((pObj) = Gia_ManObj(p.pGia, Vec_IntEntry(vRes,i))); i++  {
            if Gla_ObjSatValue( p, Gia_ObjId(p.pGia, pObj), f ) != 0 {
                Gia_ObjTerSimSet1( pObj );
            } else {
                Gia_ObjTerSimSet0( pObj ); }
        }

        // inlining of macro Gia_ManForEachObjVec
        for  i = 0; (i < Vec_IntSize(vRoAnds)) && ((pObj) = Gia_ManObj(p.pGia, Vec_IntEntry(vRoAnds,i))); i++  {
            if Gia_ObjIsAnd(pObj) {
                Gia_ObjTerSimAnd( pObj );
            } else if f == 0 {
                Gia_ObjTerSimSet0( pObj );
            } else {
                Gia_ObjTerSimRo( p.pGia, pObj ); }
        }
        // inlining of macro Gia_ManForEachObjVec
        for  i = 0; (i < Vec_IntSize(vCos)) && ((pObj) = Gia_ManObj(p.pGia, Vec_IntEntry(vCos,i))); i++  {
            Gia_ObjTerSimCo( pObj );
        }
    }
    pObj = Gia_ManPo( p.pGia, 0 );
    if !Gia_ObjTerSimGet1(pObj) {
        Abc_Print( 1, "\nRefinement verification has failed!!!\n" ); }
    // clear
    Gia_ObjTerSimSetC( Gia_ManConst0(p.pGia) );
    // inlining of macro Gia_ManForEachObjVec
    for  i = 0; (i < Vec_IntSize(vPis)) && ((pObj) = Gia_ManObj(p.pGia, Vec_IntEntry(vPis,i))); i++  {
        Gia_ObjTerSimSetC( pObj );
    }
    // inlining of macro Gia_ManForEachObjVec
    for  i = 0; (i < Vec_IntSize(vPPis)) && ((pObj) = Gia_ManObj(p.pGia, Vec_IntEntry(vPPis,i))); i++  {
        Gia_ObjTerSimSetC( pObj );
    }
    // inlining of macro Gia_ManForEachObjVec
    for  i = 0; (i < Vec_IntSize(vRoAnds)) && ((pObj) = Gia_ManObj(p.pGia, Vec_IntEntry(vRoAnds,i))); i++  {
        Gia_ObjTerSimSetC( pObj );
    }
    // inlining of macro Gia_ManForEachObjVec
    for  i = 0; (i < Vec_IntSize(vCos)) && ((pObj) = Gia_ManObj(p.pGia, Vec_IntEntry(vCos,i))); i++  {
        Gia_ObjTerSimSetC( pObj );
    }
}


/**Function*************************************************************

  Synopsis    [Performs refinement.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Gla_ManRefinement(p *Gla_Man_t) *Vec_Int_t {
    var pCex *Abc_Cex_t;
    var vMap, vVec *Vec_Int_t;
    var pObj *Gia_Obj_t;
    var i int32;
    Gia_GlaPrepareCexAndMap( p, &pCex, &vMap );
    vVec = Rnm_ManRefine( p.pRnm, pCex, vMap, p.pPars.fPropFanout, p.pPars.fNewRefine, 1 );
    Abc_CexFree( pCex );
    if Vec_IntSize(vVec) == 0 {
        Vec_IntFree( vVec );
        Abc_CexFreeP( &p.pGia.pCexSeq );
        p.pGia.pCexSeq = Gla_ManDeriveCex( p, vMap );
        Vec_IntFree( vMap );
        return nil;
    }
    Vec_IntFree( vMap );
    // remap them into GLA objects
    // inlining of macro Gia_ManForEachObjVec
    for  i = 0; (i < Vec_IntSize(vVec)) && ((pObj) = Gia_ManObj(p.pGia, Vec_IntEntry(vVec,i))); i++  {
        Vec_IntWriteEntry( vVec, i, p.pObj2Obj[Gia_ObjId(p.pGia, pObj)] );
    }
    p.nObjAdded += Vec_IntSize(vVec);
    return vVec;
}

/**Function*************************************************************

  Synopsis    [Performs refinement.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
var spStatic_Gla_ManRefinement2_Sign int32;
func Gla_ManRefinement2(p *Gla_Man_t) *Vec_Int_t {
    var fVerify int32  = 1;
    
    var vPis *Vec_Int_t;
    var vPPis *Vec_Int_t;
    var vCos *Vec_Int_t;
    var vRoAnds *Vec_Int_t;
    var vSelect *Vec_Int_t = nil;
    var pRef, pRef0, pRef1 *Rfn_Obj_t;
    var pObj *Gia_Obj_t;
    var i, f int32;
    spStatic_Gla_ManRefinement2_Sign++;

    // compute PIs and pseudo-PIs
    vCos = Vec_IntAlloc( 1000 );
    vPis = Vec_IntAlloc( 1000 );  
    vPPis = Vec_IntAlloc( 1000 );
    vRoAnds = Vec_IntAlloc( 1000 );  
    Gla_ManCollect( p, vPis, vPPis, vCos, vRoAnds );

/*
    // check how many pseudo PIs have variables
    Gla_ManForEachObjAbsVec( vPis, p, pGla, i )
    {
        Abc_Print( 1, "  %5d : ", Gla_ObjId(p, pGla) );
        for ( f = 0; f <= p->pPars->iFrame; f++ )
            Abc_Print( 1, "%d", Gla_ManCheckVar(p, Gla_ObjId(p, pGla), f) );
        Abc_Print( 1, "\n" );
    }    

    // check how many pseudo PIs have variables
    Gla_ManForEachObjAbsVec( vPPis, p, pGla, i )
    {
        Abc_Print( 1, "%5d : ", Gla_ObjId(p, pGla) );
        for ( f = 0; f <= p->pPars->iFrame; f++ )
            Abc_Print( 1, "%d", Gla_ManCheckVar(p, Gla_ObjId(p, pGla), f) );
        Abc_Print( 1, "\n" );
    }    
*/
    // propagate values
    for  f = 0; f <= p.pPars.iFrame; f++  {
        // constant
        pRef = Gla_ObjRef( p, Gia_ManConst0(p.pGia), f );  Gla_ObjClearRef( pRef );
        pRef.Value  = 0;
        pRef.Prio   = 0;
        pRef.Sign   = spStatic_Gla_ManRefinement2_Sign;
        // primary input
        // inlining of macro Gia_ManForEachObjVec
        for  i = 0; (i < Vec_IntSize(vPis)) && ((pObj) = Gia_ManObj(p.pGia, Vec_IntEntry(vPis,i))); i++  {
//            assert( f == p->pPars->iFrame || Gla_ManCheckVar(p, p->pObj2Obj[Gia_ObjId(p->pGia, pObj)], f) );
            pRef = Gla_ObjRef( p, pObj, f );   Gla_ObjClearRef( pRef );
            pRef.Value = Gla_ObjSatValue( p, Gia_ObjId(p.pGia, pObj), f );
            pRef.Prio  = 0;
            pRef.Sign  = spStatic_Gla_ManRefinement2_Sign;
            assert(pRef.fVisit == 0, "pRef.fVisit == 0");
        }
        // primary input
        // inlining of macro Gia_ManForEachObjVec
        for  i = 0; (i < Vec_IntSize(vPPis)) && ((pObj) = Gia_ManObj(p.pGia, Vec_IntEntry(vPPis,i))); i++  {
//            assert( f == p->pPars->iFrame || Gla_ManCheckVar(p, p->pObj2Obj[Gia_ObjId(p->pGia, pObj)], f) );
            assert(Gia_ObjIsAnd(pObj) || Gia_ObjIsRo(p.pGia, pObj), "Gia_ObjIsAnd(pObj) || Gia_ObjIsRo(p.pGia, pObj)");
            pRef = Gla_ObjRef( p, pObj, f );   Gla_ObjClearRef( pRef );
            pRef.Value = Gla_ObjSatValue( p, Gia_ObjId(p.pGia, pObj), f );
            pRef.Prio  = i+1;
            pRef.fPPi  = 1;
            pRef.Sign  = spStatic_Gla_ManRefinement2_Sign;
            assert(pRef.fVisit == 0, "pRef.fVisit == 0");
        }
        // internal nodes
        // inlining of macro Gia_ManForEachObjVec
        for  i = 0; (i < Vec_IntSize(vRoAnds)) && ((pObj) = Gia_ManObj(p.pGia, Vec_IntEntry(vRoAnds,i))); i++  {
            assert(Gia_ObjIsAnd(pObj) || Gia_ObjIsRo(p.pGia, pObj), "Gia_ObjIsAnd(pObj) || Gia_ObjIsRo(p.pGia, pObj)");
            pRef = Gla_ObjRef( p, pObj, f );   Gla_ObjClearRef( pRef );
            if Gia_ObjIsRo(p.pGia, pObj) {
                if f == 0 {
                    pRef.Value = 0;
                    pRef.Prio  = 0;
                    pRef.Sign  = spStatic_Gla_ManRefinement2_Sign;
                
                } else {
                    pRef0 = Gla_ObjRef( p, Gia_ObjRoToRi(p.pGia, pObj), f-1 );
                    pRef.Value = pRef0.Value;
                    pRef.Prio  = pRef0.Prio;
                    pRef.Sign  = spStatic_Gla_ManRefinement2_Sign;
                }
                continue;
            }
            assert(Gia_ObjIsAnd(pObj), "Gia_ObjIsAnd(pObj)");
            pRef0 = Gla_ObjRef( p, Gia_ObjFanin0(pObj), f );
            pRef1 = Gla_ObjRef( p, Gia_ObjFanin1(pObj), f );
            pRef.Value = (pRef0.Value ^ Gia_ObjFaninC0(pObj)) & (pRef1.Value ^ Gia_ObjFaninC1(pObj));

            if p.pObj2Obj[Gia_ObjId(p.pGia, pObj)] != ~0 && 
                 (Gla_ManCheckVar(p, p.pObj2Obj[Gia_ObjId(p.pGia, pObj)], f) != 0) &&
                 int32(pRef.Value) != Gla_ObjSatValue(p, Gia_ObjId(p.pGia, pObj), f) {
                    Abc_Print( 1, "Object has value mismatch    " );
                    Gia_ObjPrint( p.pGia, pObj );
            }

            if pRef.Value == 1 {
                pRef.Prio  = Abc_MaxInt( pRef0.Prio, pRef1.Prio );
            } else if (pRef0.Value ^ Gia_ObjFaninC0(pObj)) == 0 && (pRef1.Value ^ Gia_ObjFaninC1(pObj)) == 0 {
                pRef.Prio  = Abc_MinInt( pRef0.Prio, pRef1.Prio ); // choice
            } else if (pRef0.Value ^ Gia_ObjFaninC0(pObj)) == 0 {
                pRef.Prio  = pRef0.Prio;
            } else { 
                pRef.Prio  = pRef1.Prio; }
            assert(pRef.fVisit == 0, "pRef.fVisit == 0");
            pRef.Sign  = spStatic_Gla_ManRefinement2_Sign;
        }
        // output nodes
        // inlining of macro Gia_ManForEachObjVec
        for  i = 0; (i < Vec_IntSize(vCos)) && ((pObj) = Gia_ManObj(p.pGia, Vec_IntEntry(vCos,i))); i++  {
            pRef = Gla_ObjRef( p, pObj, f );    Gla_ObjClearRef( pRef );
            pRef0 = Gla_ObjRef( p, Gia_ObjFanin0(pObj), f ); 
            pRef.Value = (pRef0.Value ^ Gia_ObjFaninC0(pObj));
            pRef.Prio  = pRef0.Prio;
            assert(pRef.fVisit == 0, "pRef.fVisit == 0");
            pRef.Sign  = spStatic_Gla_ManRefinement2_Sign;
        }
    } 

    // make sure the output value is 1
    pObj = Gia_ManPo( p.pGia, 0 );
    pRef = Gla_ObjRef( p, pObj, p.pPars.iFrame );
    if pRef.Value != 1 {
        Abc_Print( 1, "\nCounter-example verification has failed!!!\n" ); }

    // check the CEX
    if pRef.Prio == 0 {
        p.pGia.pCexSeq = Gla_ManDeriveCex( p, vPis );
        Vec_IntFree( vPis );
        Vec_IntFree( vPPis );
        Vec_IntFree( vRoAnds );
        Vec_IntFree( vCos );
        return nil;
    }

    // select objects
    vSelect = Vec_IntAlloc( 100 );
    Vec_IntFill( p.vObjCounts, p.pPars.iFrame+1, 0 );
    Gla_ManRefSelect_rec( p, Gia_ObjFanin0(Gia_ManPo(p.pGia, 0)), p.pPars.iFrame, vSelect, spStatic_Gla_ManRefinement2_Sign );
    Vec_IntUniqify( vSelect );

/*
    for ( f = 0; f < p->pPars->iFrame; f++ )
        printf( "%2d", Vec_IntEntry(p->vObjCounts, f) );
    printf( "\n" );
*/
    if fVerify != 0 {
        Gla_ManVerifyUsingTerSim( p, vPis, vPPis, vRoAnds, vCos, vSelect ); }

    // remap them into GLA objects
    // inlining of macro Gia_ManForEachObjVec
    for  i = 0; (i < Vec_IntSize(vSelect)) && ((pObj) = Gia_ManObj(p.pGia, Vec_IntEntry(vSelect,i))); i++  {
        Vec_IntWriteEntry( vSelect, i, p.pObj2Obj[Gia_ObjId(p.pGia, pObj)] );
    }

    Vec_IntFree( vPis );
    Vec_IntFree( vPPis );
    Vec_IntFree( vRoAnds );
    Vec_IntFree( vCos );

    p.nObjAdded += Vec_IntSize(vSelect);
    return vSelect;
}


/**Function*************************************************************

  Synopsis    [Adds clauses for the given obj in the given frame.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Gla_ManCollectFanins(p *Gla_Man_t, pGla *Gla_Obj_t, iObj int32, vFanins *Vec_Int_t) {
    var i int32;
    var nClauses int32;
    var iFirstClause int32;
    var pLit *int32;
    nClauses = p.pCnf.pObj2Count[pGla.iGiaObj];
    iFirstClause = p.pCnf.pObj2Clause[pGla.iGiaObj];
    Vec_IntClear( vFanins );
    for  i = iFirstClause; i < iFirstClause + nClauses; i++  {
        for  pLit = p.pCnf.pClauses[i]; pLit < p.pCnf.pClauses[i+1]; pLit++  {
            if lit_var(*pLit) != iObj {
                Vec_IntPushUnique( vFanins, lit_var(*pLit) ); }
        }
    }
    assert(Vec_IntSize( vFanins ) <= 4, "Vec_IntSize( vFanins ) <= 4");
    Vec_IntSort( vFanins, 0 );
}


/**Function*************************************************************

  Synopsis    [Duplicates AIG while decoupling nodes duplicated in the mapping.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
/**Function*************************************************************

  Synopsis    [Creates a new manager.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Gla_ManStart(pGia0 *Gia_Man_t, pPars *Abs_Par_t) *Gla_Man_t {
    
    var pAig *Aig_Man_t;
    var pObj *Gia_Obj_t;
    var pGla *Gla_Obj_t;
    var vMappingNew *Vec_Int_t;
    var i int32;
    var k int32;
    var Offset int32;
    var pMapping *int32;
    var pLits *int32;
    var pObj2Count *int32;
    var pObj2Clause *int32;

    // start
    // inlining of macro ABC_CALLOC
    p := &Gla_Man_t{};
    p.pGia0 = pGia0;
    p.pPars = pPars;
    p.vAbs  = Vec_IntAlloc( 100 );
    p.vTemp = Vec_IntAlloc( 100 );
    p.vAddedNew = Vec_IntAlloc( 100 );
    p.vObjCounts = Vec_IntAlloc( 100 );

    // internal data
    pAig = Gia_ManToAigSimple( pGia0 );
    p.pCnf = Cnf_DeriveOther( pAig, 1 );
    Aig_ManStop( pAig );
    // create working GIA
    p.pGia = Gia_ManDupMapped( pGia0, p.pCnf.vMapping );
    if pPars.fPropFanout {
        Gia_ManStaticFanoutStart( p.pGia ); }

    // derive new gate map
    assert(pGia0.vGateClasses != 0, "pGia0.vGateClasses != 0");
    p.pGia.vGateClasses = Vec_IntStart( Gia_ManObjNum(p.pGia) );
    p.vCoreCounts = Vec_IntStart( Gia_ManObjNum(p.pGia) );
    p.vProofIds = Vec_IntAlloc(0);
    // update p->pCnf->vMapping, p->pCnf->pObj2Count, p->pCnf->pObj2Clause 
    // (here are not updating p->pCnf->pVarNums because it is not needed)
    vMappingNew = Vec_IntStart( Gia_ManObjNum(p.pGia) );
    // inlining of macro ABC_FALLOC
    pObj2Count  = ((int *) memset(malloc(sizeof(int) * (size_t)(Gia_ManObjNum(p.pGia))), 0xff, sizeof(int) * (size_t)(Gia_ManObjNum(p.pGia))));
    // inlining of macro ABC_FALLOC
    pObj2Clause = ((int *) memset(malloc(sizeof(int) * (size_t)(Gia_ManObjNum(p.pGia))), 0xff, sizeof(int) * (size_t)(Gia_ManObjNum(p.pGia))));
    // inlining of macro Gia_ManForEachObj
    for  i = 0; (i < pGia0.nObjs) && ((pObj) = Gia_ManObj(pGia0, i)); i++  {
        // skip internal nodes not used in the mapping
        if !~pObj.Value {
            continue; }
        // replace positive literal by variable
        assert(!Abc_LitIsCompl(pObj.Value), "!Abc_LitIsCompl(pObj.Value)");
        pObj.Value = Abc_Lit2Var(pObj.Value);
        assert(int32(pObj.Value) < Gia_ManObjNum(p.pGia), "int32(pObj.Value) < Gia_ManObjNum(p.pGia)");
        // update arrays
        pObj2Count[pObj.Value] = p.pCnf.pObj2Count[i];
        pObj2Clause[pObj.Value] = p.pCnf.pObj2Clause[i];
        if Vec_IntEntry(pGia0.vGateClasses, i) {
            Vec_IntWriteEntry( p.pGia.vGateClasses, pObj.Value, 1 ); }
        // update mappings
        Offset = Vec_IntEntry(p.pCnf.vMapping, i);
        Vec_IntWriteEntry( vMappingNew, pObj.Value, Vec_IntSize(vMappingNew) );
        pMapping = Vec_IntEntryP(p.pCnf.vMapping, Offset);
        Vec_IntPush( vMappingNew, pMapping[0] );
        for  k = 1; k <= 4; k++  {
            if pMapping[k] == -1 {
                Vec_IntPush( vMappingNew, -1 );
            } else {
                assert(~Gia_ManObj(pGia0, pMapping[k]).Value, "~Gia_ManObj(pGia0, pMapping[k]).Value");
                Vec_IntPush( vMappingNew, Gia_ManObj(pGia0, pMapping[k]).Value );
            }
        }
    }
    // update mapping after the offset (currently not being done because it is not used)
    Vec_IntFree( p.pCnf.vMapping );  p.pCnf.vMapping    = vMappingNew;
    p.pCnf.pObj2Count = nil;   p.pCnf.pObj2Count  = pObj2Count;
    p.pCnf.pObj2Clause = nil;  p.pCnf.pObj2Clause = pObj2Clause;


    // count the number of variables
    p.nObjs = 1;
    // inlining of macro Gia_ManForEachObj
    for  i = 0; (i < p.pGia.nObjs) && ((pObj) = Gia_ManObj(p.pGia, i)); i++  {
        if p.pCnf.pObj2Count[i] >= 0 {
            pObj.Value = p.nObjs++;
        } else {
            pObj.Value = ~0; }
    }

    // re-express CNF using new variable IDs
    pLits = p.pCnf.pClauses[0];
    for  i = 0; i < p.pCnf.nLiterals; i++  {
        // find the original AIG object
        pObj = Gia_ManObj( pGia0, lit_var(pLits[i]) );
        assert(~pObj.Value, "~pObj.Value");
        // find the working AIG object
        pObj = Gia_ManObj( p.pGia, pObj.Value );
        assert(~pObj.Value, "~pObj.Value");
        // express literal in terms of LUT variables
        pLits[i] = toLitCond( pObj.Value, lit_sign(pLits[i]) );
    }

    // create objects 
    // inlining of macro ABC_CALLOC
    p.pObjs    = ((Gla_Obj_t *) calloc((size_t)(p.nObjs), sizeof(Gla_Obj_t)));
    // inlining of macro ABC_FALLOC
    p.pObj2Obj = ((unsigned *) memset(malloc(sizeof(unsigned) * (size_t)(Gia_ManObjNum(p.pGia))), 0xff, sizeof(unsigned) * (size_t)(Gia_ManObjNum(p.pGia))));
//    p->pvRefis  = ABC_CALLOC( Vec_Int_t, Gia_ManObjNum(p->pGia) );
    // inlining of macro Gia_ManForEachObj
    for  i = 0; (i < p.pGia.nObjs) && ((pObj) = Gia_ManObj(p.pGia, i)); i++  {
        p.pObj2Obj[i] = pObj.Value;
        if !~pObj.Value {
            continue; }
        pGla = Gla_ManObj( p, pObj.Value );
        pGla.iGiaObj = i;
        pGla.fCompl0 = Gia_ObjFaninC0(pObj);
        pGla.fConst  = Gia_ObjIsConst0(pObj);
        pGla.fPi     = Gia_ObjIsPi(p.pGia, pObj);
        pGla.fPo     = Gia_ObjIsPo(p.pGia, pObj);
        pGla.fRi     = Gia_ObjIsRi(p.pGia, pObj);
        pGla.fRo     = Gia_ObjIsRo(p.pGia, pObj);
        pGla.fAnd    = Gia_ObjIsAnd(pObj);
        if Gia_ObjIsConst0(pObj) || Gia_ObjIsPi(p.pGia, pObj) {
            continue; }
        if Gia_ObjIsCo(pObj) {
            pGla.nFanins = 1;
            pGla.Fanins[0] = Gia_ObjFanin0(pObj).Value;
            continue;
        }
        if Gia_ObjIsAnd(pObj) {
//            Gla_ManCollectFanins( p, pGla, pObj->Value, p->vTemp );
//            pGla->nFanins = Vec_IntSize( p->vTemp );
//            memcpy( pGla->Fanins, Vec_IntArray(p->vTemp), sizeof(int) * Vec_IntSize(p->vTemp) );
            Offset = Vec_IntEntry( p.pCnf.vMapping, i );
            pMapping = Vec_IntEntryP( p.pCnf.vMapping, Offset );
            pGla.nFanins = 0;
            for  k = 1; k <= 4; k++  {
                if pMapping[k] != -1 {
                    pGla.Fanins[ pGla.nFanins++ ] = Gia_ManObj(p.pGia, pMapping[k]).Value; }
            }
            continue;
        }
        assert(Gia_ObjIsRo(p.pGia, pObj), "Gia_ObjIsRo(p.pGia, pObj)");
        pGla.nFanins   = 1;
        pGla.Fanins[0] = Gia_ObjFanin0( Gia_ObjRoToRi(p.pGia, pObj) ).Value;
        pGla.fCompl0   = Gia_ObjFaninC0( Gia_ObjRoToRi(p.pGia, pObj) );
    }
    p.pObjRoot = Gla_ManObj( p, Gia_ManPo(p.pGia, 0).Value );
    // abstraction 
    assert(p.pGia.vGateClasses != nil, "p.pGia.vGateClasses != nil");
    // inlining of macro Gla_ManForEachObj
    for  pGla = p.pObjs + 1; pGla < p.pObjs + p.nObjs; pGla++  {
        if Vec_IntEntry( p.pGia.vGateClasses, pGla.iGiaObj ) == 0 {
            continue; }
        pGla.fAbs = 1;
        Vec_IntPush( p.vAbs, Gla_ObjId(p, pGla) );
    }
    // other 
    p.pSat        = sat_solver2_new();
    if pPars.fUseFullProof {
        p.pSat.pPrf1 = Vec_SetAlloc( 20 ); }
//    p->pSat->fVerbose = p->pPars->fVerbose;
//    sat_solver2_set_learntmax( p->pSat, pPars->nLearnedMax );
    p.pSat.nLearntStart = p.pPars.nLearnedStart;
    p.pSat.nLearntDelta = p.pPars.nLearnedDelta;
    p.pSat.nLearntRatio = p.pPars.nLearnedPerce;
    p.pSat.nLearntMax   = p.pSat.nLearntStart;
    p.nSatVars  = 1;
    // start the refinement manager
//    p->pGia2 = Gia_ManDup( p->pGia );
    p.pRnm = Rnm_ManStart( p.pGia );
    return p;
}

/**Function*************************************************************

  Synopsis    [Creates a new manager.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Gla_ManStart2(pGia *Gia_Man_t, pPars *Abs_Par_t) *Gla_Man_t {
    
    var pAig *Aig_Man_t;
    var pObj *Gia_Obj_t;
    var pGla *Gla_Obj_t;
    var i int32;
    var pLits *int32;
    // start
    // inlining of macro ABC_CALLOC
    p := &Gla_Man_t{};
    p.pGia  = pGia;
    p.pPars = pPars;
    p.vAbs = Vec_IntAlloc( 100 );
    p.vTemp = Vec_IntAlloc( 100 );
    p.vAddedNew = Vec_IntAlloc( 100 );
    // internal data
    pAig = Gia_ManToAigSimple( p.pGia );
    p.pCnf  = Cnf_DeriveOther( pAig, 1 );
    Aig_ManStop( pAig );
    // count the number of variables
    p.nObjs = 1;
    // inlining of macro Gia_ManForEachObj
    for  i = 0; (i < p.pGia.nObjs) && ((pObj) = Gia_ManObj(p.pGia, i)); i++  {
        if p.pCnf.pObj2Count[i] >= 0 {
            pObj.Value = p.nObjs++;
        } else {
            pObj.Value = ~0; }
    }
    // re-express CNF using new variable IDs
    pLits = p.pCnf.pClauses[0];
    for  i = 0; i < p.pCnf.nLiterals; i++  {
        pObj = Gia_ManObj( p.pGia, lit_var(pLits[i]) );
        assert(~pObj.Value, "~pObj.Value");
        pLits[i] = toLitCond( pObj.Value, lit_sign(pLits[i]) );
    }
    // create objects 
    // inlining of macro ABC_CALLOC
    p.pObjs    = ((Gla_Obj_t *) calloc((size_t)(p.nObjs), sizeof(Gla_Obj_t)));
    // inlining of macro ABC_FALLOC
    p.pObj2Obj = ((unsigned *) memset(malloc(sizeof(unsigned) * (size_t)(Gia_ManObjNum(p.pGia))), 0xff, sizeof(unsigned) * (size_t)(Gia_ManObjNum(p.pGia))));
//    p->pvRefis  = ABC_CALLOC( Vec_Int_t, Gia_ManObjNum(p->pGia) );
    // inlining of macro Gia_ManForEachObj
    for  i = 0; (i < p.pGia.nObjs) && ((pObj) = Gia_ManObj(p.pGia, i)); i++  {
        p.pObj2Obj[i] = pObj.Value;
        if !~pObj.Value {
            continue; }
        pGla = Gla_ManObj( p, pObj.Value );
        pGla.iGiaObj = i;
        pGla.fCompl0 = Gia_ObjFaninC0(pObj);
        pGla.fConst  = Gia_ObjIsConst0(pObj);
        pGla.fPi     = Gia_ObjIsPi(p.pGia, pObj);
        pGla.fPo     = Gia_ObjIsPo(p.pGia, pObj);
        pGla.fRi     = Gia_ObjIsRi(p.pGia, pObj);
        pGla.fRo     = Gia_ObjIsRo(p.pGia, pObj);
        pGla.fAnd    = Gia_ObjIsAnd(pObj);
        if Gia_ObjIsConst0(pObj) || Gia_ObjIsPi(p.pGia, pObj) {
            continue; }
        if Gia_ObjIsAnd(pObj) || Gia_ObjIsCo(pObj) {
            Gla_ManCollectFanins( p, pGla, pObj.Value, p.vTemp );
            pGla.nFanins = Vec_IntSize( p.vTemp );
            memcpy( pGla.Fanins, Vec_IntArray(p.vTemp), sizeof(int) * Vec_IntSize(p.vTemp) );
            continue;
        }
        assert(Gia_ObjIsRo(p.pGia, pObj), "Gia_ObjIsRo(p.pGia, pObj)");
        pGla.nFanins   = 1;
        pGla.Fanins[0] = Gia_ObjFanin0( Gia_ObjRoToRi(p.pGia, pObj) ).Value;
        pGla.fCompl0   = Gia_ObjFaninC0( Gia_ObjRoToRi(p.pGia, pObj) );
    }
    p.pObjRoot = Gla_ManObj( p, Gia_ManPo(p.pGia, 0).Value );
    // abstraction 
    assert(pGia.vGateClasses != nil, "pGia.vGateClasses != nil");
    // inlining of macro Gla_ManForEachObj
    for  pGla = p.pObjs + 1; pGla < p.pObjs + p.nObjs; pGla++  {
        if Vec_IntEntry( pGia.vGateClasses, pGla.iGiaObj ) == 0 {
            continue; }
        pGla.fAbs = 1;
        Vec_IntPush( p.vAbs, Gla_ObjId(p, pGla) );
    }
    // other 
    p.pSat      = sat_solver2_new();
    p.nSatVars  = 1;
    return p;
}

/**Function*************************************************************

  Synopsis    [Creates a new manager.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Gla_ManStop(p *Gla_Man_t) {
    var pGla *Gla_Obj_t;
    var i int32;

    if p.pPars.fVerbose {
        Abc_Print( 1, "SAT solver:  Var = %d  Cla = %d  Conf = %d  Lrn = %d  Reduce = %d  Cex = %d  Objs+ = %d\n", 
            sat_solver2_nvars(p.pSat), sat_solver2_nclauses(p.pSat), sat_solver2_nconflicts(p.pSat), 
            sat_solver2_nlearnts(p.pSat), p.pSat.nDBreduces, p.nCexes, p.nObjAdded ); }

    // stop the refinement manager
//    Gia_ManStopP( &p->pGia2 );
    Rnm_ManStop( p.pRnm, 0 );

    if p.pvRefis {
    for  i = 0; i < Gia_ManObjNum(p.pGia); i++  {
        p.pvRefis[i].pArray = nil;
    } }
    // inlining of macro Gla_ManForEachObj
    for  pGla = p.pObjs + 1; pGla < p.pObjs + p.nObjs; pGla++  { 
        pGla.vFrames.pArray = nil;
    }
    Cnf_DataFree( p.pCnf );
    if p.pGia0 != nil {
        Gia_ManStop( p.pGia ); }
//    Gia_ManStaticFanoutStart( p->pGia0 );
    sat_solver2_delete( p.pSat );
    Vec_IntFreeP( &p.vObjCounts );
    Vec_IntFreeP( &p.vAddedNew );
    Vec_IntFreeP( &p.vCoreCounts );
    Vec_IntFreeP( &p.vProofIds );
    Vec_IntFreeP( &p.vTemp );
    Vec_IntFreeP( &p.vAbs );
    p.pvRefis = nil;
    p.pObj2Obj = nil;
    p.pObjs = nil;
    p = nil;
}

/**Function*************************************************************

  Synopsis    [Creates a new manager.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Gia_GlaAbsCount(p *Gla_Man_t, fRo int32, fAnd int32) int32 {
    var pObj *Gla_Obj_t;
    var i, Counter int32;
    if fRo {
        // inlining of macro Gla_ManForEachObjAbs
        for  i = 0; i < Vec_IntSize(p.vAbs) && ((pObj = Gla_ManObj(p, Vec_IntEntry(p.vAbs, i))),1); i++ { 
            Counter += (pObj.fRo && pObj.fAbs);
        }
    } else if fAnd {
        // inlining of macro Gla_ManForEachObjAbs
        for  i = 0; i < Vec_IntSize(p.vAbs) && ((pObj = Gla_ManObj(p, Vec_IntEntry(p.vAbs, i))),1); i++ { 
            Counter += (pObj.fAnd && pObj.fAbs);
        }
    } else {
        // inlining of macro Gla_ManForEachObjAbs
        for  i = 0; i < Vec_IntSize(p.vAbs) && ((pObj = Gla_ManObj(p, Vec_IntEntry(p.vAbs, i))),1); i++ { 
            Counter += (pObj.fAbs);
        } }
    return Counter;
}


/**Function*************************************************************

  Synopsis    [Derives new abstraction map.]

  Description [Returns 1 if node contains abstracted leaf on the path.]
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Gla_ManTranslate_rec(p *Gia_Man_t, pObj *Gia_Obj_t, vGla *Vec_Int_t, nUsageCount int32) int32 {
    var Value0, Value1 int32;
    if Gia_ObjIsTravIdCurrent(p, pObj) {
        return 1; }
    Gia_ObjSetTravIdCurrent(p, pObj);
    if Gia_ObjIsCi(pObj) {
        return 0; }
    assert(Gia_ObjIsAnd(pObj), "Gia_ObjIsAnd(pObj)");
    Value0 = Gla_ManTranslate_rec( p, Gia_ObjFanin0(pObj), vGla, nUsageCount );
    Value1 = Gla_ManTranslate_rec( p, Gia_ObjFanin1(pObj), vGla, nUsageCount );
    if Value0 || (Value1 != 0) {
        Vec_IntAddToEntry( vGla, Gia_ObjId(p, pObj), nUsageCount ); }
    return Value0 || Value1;
}
func Gla_ManTranslate(p *Gla_Man_t) *Vec_Int_t {
    var vGla, vGla2 *Vec_Int_t;
    var pObj, pFanin *Gla_Obj_t;
    var pGiaObj *Gia_Obj_t;
    var i, k, nUsageCount int32;
    vGla = Vec_IntStart( Gia_ManObjNum(p.pGia) );
    // inlining of macro Gla_ManForEachObjAbs
    for  i = 0; i < Vec_IntSize(p.vAbs) && ((pObj = Gla_ManObj(p, Vec_IntEntry(p.vAbs, i))),1); i++ {
        nUsageCount = Vec_IntEntry(p.vCoreCounts, pObj.iGiaObj);
        assert(nUsageCount >= 0, "nUsageCount >= 0");
        if nUsageCount == 0 {
            nUsageCount++; }
        pGiaObj = Gla_ManGiaObj( p, pObj );
        if Gia_ObjIsConst0(pGiaObj) || Gia_ObjIsRo(p.pGia, pGiaObj) {
            Vec_IntWriteEntry( vGla, pObj.iGiaObj, nUsageCount );
            continue;
        }
        assert(Gia_ObjIsAnd(pGiaObj), "Gia_ObjIsAnd(pGiaObj)");
        Gia_ManIncrementTravId( p.pGia );
        // inlining of macro Gla_ObjForEachFanin
        for  k = 0; (k < int32(pObj.nFanins)) && ((pFanin = Gla_ManObj(p, pObj.Fanins[k])),1); k++  {
            Gia_ObjSetTravIdCurrent( p.pGia, Gla_ManGiaObj(p, pFanin) );
        }
        Gla_ManTranslate_rec( p.pGia, pGiaObj, vGla, nUsageCount );
    }
    Vec_IntWriteEntry( vGla, 0, p.pPars.iFrame+1 );
    if p.pGia.vLutConfigs { // use mapping from new to old
    
        vGla2 = Vec_IntStart( Gia_ManObjNum(p.pGia0) );
        for  i = 0; i < Gia_ManObjNum(p.pGia); i++  {
            if Vec_IntEntry(vGla, i) {
                Vec_IntWriteEntry( vGla2, Vec_IntEntry(p.pGia.vLutConfigs, i), Vec_IntEntry(vGla, i) ); }
        }
        Vec_IntFree( vGla );
        return vGla2;
    }
    return vGla;
}


/**Function*************************************************************

  Synopsis    [Collect pseudo-PIs.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Gla_ManCollectPPis(p *Gla_Man_t, vPis *Vec_Int_t) *Vec_Int_t {
    var vPPis *Vec_Int_t;
    var pObj, pFanin *Gla_Obj_t;
    var i, k int32;
    vPPis = Vec_IntAlloc( 1000 );
    if vPis != nil {
        Vec_IntClear( vPis ); }
    // inlining of macro Gla_ManForEachObjAbs
    for  i = 0; i < Vec_IntSize(p.vAbs) && ((pObj = Gla_ManObj(p, Vec_IntEntry(p.vAbs, i))),1); i++ {
        assert(pObj.fConst || pObj.fRo || pObj.fAnd, "pObj.fConst || pObj.fRo || pObj.fAnd");
        // inlining of macro Gla_ObjForEachFanin
        for  k = 0; (k < int32(pObj.nFanins)) && ((pFanin = Gla_ManObj(p, pObj.Fanins[k])),1); k++  {
            if !pFanin.fPi && !pFanin.fAbs {
                Vec_IntPush( vPPis, pObj.Fanins[k] );
            } else if (vPis != nil) && pFanin.fPi && !pFanin.fAbs {
                Vec_IntPush( vPis, pObj.Fanins[k] ); }
        }
    }
    Vec_IntUniqify( vPPis );
    Vec_IntReverseOrder( vPPis );
    if vPis != nil {
        Vec_IntUniqify( vPis ); }
    return vPPis;
}
func Gla_ManCountPPis(p *Gla_Man_t) int32 {
    var vPPis *Vec_Int_t  = Gla_ManCollectPPis( p, nil );
    var RetValue int32  = Vec_IntSize( vPPis );
    Vec_IntFree( vPPis );
    return RetValue;
}
var spStatic_Gla_ManExplorePPis_Round int32;
func Gla_ManExplorePPis(p *Gla_Man_t, vPPis *Vec_Int_t) {
    
    var pObj, pFanin *Gla_Obj_t;
    var i, j, k, Count int32;
    if (spStatic_Gla_ManExplorePPis_Round++ % 5) == 0 {
        return; }
    j = 0;
    // inlining of macro Gla_ManForEachObjAbsVec
    for  i = 0; i < Vec_IntSize(vPPis) && ((pObj = Gla_ManObj(p, Vec_IntEntry(vPPis, i))),1); i++ {
        assert(pObj.fAbs == 0, "pObj.fAbs == 0");
        Count = 0;
        // inlining of macro Gla_ObjForEachFanin
        for  k = 0; (k < int32(pObj.nFanins)) && ((pFanin = Gla_ManObj(p, pObj.Fanins[k])),1); k++  {
            Count += pFanin.fAbs;
        }
        if Count == 0 || ((spStatic_Gla_ManExplorePPis_Round & 1) && Count == 1) {
            continue; }
        Vec_IntWriteEntry( vPPis, j++, Gla_ObjId(p, pObj) );
    }
//    printf( "\n%d -> %d\n", Vec_IntSize(vPPis), j );
    Vec_IntShrink( vPPis, j );
}


/**Function*************************************************************

  Synopsis    [Adds CNF for the given timeframe.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Gla_ManCheckVar(p *Gla_Man_t, iObj int32, iFrame int32) int32 {
    var pGla *Gla_Obj_t  = Gla_ManObj( p, iObj );
    var iVar int32  = Vec_IntGetEntry( &pGla.vFrames, iFrame );
    assert(!pGla.fPo && !pGla.fRi, "!pGla.fPo && !pGla.fRi");
    return (iVar > 0);
}
func Gla_ManGetVar(p *Gla_Man_t, iObj int32, iFrame int32) int32 {
    var pGla *Gla_Obj_t  = Gla_ManObj( p, iObj );
    var iVar int32  = Vec_IntGetEntry( &pGla.vFrames, iFrame );
    assert(!pGla.fPo && !pGla.fRi, "!pGla.fPo && !pGla.fRi");
    if iVar == 0 {
        Vec_IntSetEntry( &pGla.vFrames, iFrame, (iVar = p.nSatVars++) );
        // remember the change
        Vec_IntPush( p.vAddedNew, iObj );
        Vec_IntPush( p.vAddedNew, iFrame );
    }
    return iVar;
}
func Gla_ManAddClauses(p *Gla_Man_t, iObj int32, iFrame int32, vLits *Vec_Int_t) {
    var pGlaObj *Gla_Obj_t  = Gla_ManObj( p, iObj );
    var iVar, iVar1, iVar2 int32;
    if pGlaObj.fConst {
        iVar = Gla_ManGetVar( p, iObj, iFrame );
        sat_solver2_add_const( p.pSat, iVar, 1, 0, iObj );
    
    } else if pGlaObj.fRo {
        assert(pGlaObj.nFanins == 1, "pGlaObj.nFanins == 1");
        if iFrame == 0 {
            iVar = Gla_ManGetVar( p, iObj, iFrame );
            sat_solver2_add_const( p.pSat, iVar, 1, 0, iObj );
        
        } else {
            iVar1 = Gla_ManGetVar( p, iObj, iFrame );
            iVar2 = Gla_ManGetVar( p, pGlaObj.Fanins[0], iFrame-1 );
            sat_solver2_add_buffer( p.pSat, iVar1, iVar2, pGlaObj.fCompl0, 0, iObj );
        }
    
    } else if pGlaObj.fAnd {
        var i int32;
        var RetValue int32;
        var nClauses int32;
        var iFirstClause int32;
        var pLit *int32;
        nClauses = p.pCnf.pObj2Count[pGlaObj.iGiaObj];
        iFirstClause = p.pCnf.pObj2Clause[pGlaObj.iGiaObj];
        for  i = iFirstClause; i < iFirstClause + nClauses; i++  {
            Vec_IntClear( vLits );
            for  pLit = p.pCnf.pClauses[i]; pLit < p.pCnf.pClauses[i+1]; pLit++  {
                iVar = Gla_ManGetVar( p, lit_var(*pLit), iFrame );
                Vec_IntPush( vLits, toLitCond( iVar, lit_sign(*pLit) ) );
            }
            RetValue = sat_solver2_addclause( p.pSat, Vec_IntArray(vLits), Vec_IntArray(vLits)+Vec_IntSize(vLits), iObj );
        }
    
    } else { assert(false, "0"); }
}
func Gia_GlaAddToCounters(p *Gla_Man_t, vCore *Vec_Int_t) {
    var pGla *Gla_Obj_t;
    var i int32;
    // inlining of macro Gla_ManForEachObjAbsVec
    for  i = 0; i < Vec_IntSize(vCore) && ((pGla = Gla_ManObj(p, Vec_IntEntry(vCore, i))),1); i++ { 
        Vec_IntAddToEntry( p.vCoreCounts, pGla.iGiaObj, 1 );
    }
}
func Gia_GlaAddToAbs(p *Gla_Man_t, vAbsAdd *Vec_Int_t, fCheck int32) {
    var pGla *Gla_Obj_t;
    var i, k int32;
    // inlining of macro Gla_ManForEachObjAbsVec
    for  i = 0; i < Vec_IntSize(vAbsAdd) && ((pGla = Gla_ManObj(p, Vec_IntEntry(vAbsAdd, i))),1); i++ {
        if fCheck != 0 {
            assert(pGla.fAbs == 0, "pGla.fAbs == 0");
            if p.pSat.pPrf2 {
                Vec_IntWriteEntry( p.vProofIds, Gla_ObjId(p, pGla), p.nProofIds++ ); }
        }
        if pGla.fAbs {
            continue; }
        pGla.fAbs = 1;
        Vec_IntPush( p.vAbs, Gla_ObjId(p, pGla) );
        // filter clauses to remove those contained in the abstraction
        Vec_IntWriteEntry( vAbsAdd, k++, Gla_ObjId(p, pGla) );
    }
    Vec_IntShrink( vAbsAdd, k );
}
func Gia_GlaAddTimeFrame(p *Gla_Man_t, f int32) {
    var pObj *Gla_Obj_t;
    var i int32;
    // inlining of macro Gla_ManForEachObjAbs
    for  i = 0; i < Vec_IntSize(p.vAbs) && ((pObj = Gla_ManObj(p, Vec_IntEntry(p.vAbs, i))),1); i++ { 
        Gla_ManAddClauses( p, Gla_ObjId(p, pObj), f, p.vTemp );
    }
    sat_solver2_simplify( p.pSat );
}
func Gia_GlaAddOneSlice(p *Gla_Man_t, fCur int32, vCore *Vec_Int_t) {
    var f, i, iGlaObj int32;
    for  f = fCur; f >= 0; f--  {
        // inlining of macro Vec_IntForEachEntry
        for  i = 0; (i < Vec_IntSize(vCore)) && (((iGlaObj) = Vec_IntEntry(vCore, i)), 1); i++  {
            Gla_ManAddClauses( p, iGlaObj, f, p.vTemp );
        }
    }
    sat_solver2_simplify( p.pSat );
}
func Gla_ManRollBack(p *Gla_Man_t) {
    var i, iObj, iFrame int32;
    // inlining of macro Vec_IntForEachEntryDouble
    for  i = 0; (i+1 < Vec_IntSize(p.vAddedNew)) && (((iObj) = Vec_IntEntry(p.vAddedNew, i)), 1) && (((iFrame) = Vec_IntEntry(p.vAddedNew, i+1)), 1); i += 2  {
        assert(Vec_IntEntry( &Gla_ManObj(p, iObj).vFrames, iFrame ) > 0, "Vec_IntEntry( &Gla_ManObj(p, iObj).vFrames, iFrame ) > 0");
        Vec_IntWriteEntry( &Gla_ManObj(p, iObj).vFrames, iFrame, 0 );
    }
    // inlining of macro Vec_IntForEachEntryStart
    for  i = p.nAbsOld; (i < Vec_IntSize(p.vAbs)) && (((iObj) = Vec_IntEntry(p.vAbs, i)), 1); i++  {
        assert(Gla_ManObj( p, iObj ).fAbs == 1, "Gla_ManObj( p, iObj ).fAbs == 1");
        Gla_ManObj( p, iObj ).fAbs = 0;
    }
    Vec_IntShrink( p.vAbs, p.nAbsOld );
}


            


/**Function*************************************************************

  Synopsis    [Finds the set of clauses involved in the UNSAT core.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Gla_ManGetOutLit(p *Gla_Man_t, f int32) int32 {
    var pFanin *Gla_Obj_t  = Gla_ManObj( p, p.pObjRoot.Fanins[0] );
    var iSat int32  = Vec_IntEntry( &pFanin.vFrames, f );
    assert(iSat > 0, "iSat > 0");
    if f == 0 && pFanin.fRo && !p.pObjRoot.fCompl0 {
        return -1; }
    return Abc_Var2Lit( iSat, p.pObjRoot.fCompl0 );
}
func Gla_ManUnsatCore(p *Gla_Man_t, f int32, pSat *sat_solver2, nConfMax int32, fVerbose int32, piRetValue *int32, pnConfls *int32) *Vec_Int_t {
    var vCore *Vec_Int_t  = nil;
    var nConfPrev int32  = pSat.stats.conflicts;
    var RetValue int32;
    var iLit int32 = Gla_ManGetOutLit( p, f );
    var clk abctime  = Abc_Clock();
    if piRetValue != nil {
        *piRetValue = 1; }
    // consider special case when PO points to the flop
    // this leads to immediate conflict in the first timeframe
    if iLit == -1 {
        vCore = Vec_IntAlloc( 1 );
        Vec_IntPush( vCore, p.pObjRoot.Fanins[0] );
        return vCore;
    }
    // solve the problem
    RetValue = sat_solver2_solve( pSat, &iLit, &iLit+1, (ABC_INT64_T)nConfMax, (ABC_INT64_T)0, (ABC_INT64_T)0, (ABC_INT64_T)0 );
    if pnConfls != nil {
        *pnConfls = int32(pSat.stats.conflicts) - nConfPrev; }
    if RetValue == l_Undef {
        if piRetValue != nil {
            *piRetValue = -1; }
        return nil;
    }
    if RetValue == l_True {
        if piRetValue != nil {
            *piRetValue = 0; }
        return nil;
    }
    if fVerbose != 0 {
//        Abc_Print( 1, "%6d", (int)pSat->stats.conflicts - nConfPrev );
//        Abc_Print( 1, "UNSAT after %7d conflicts.      ", pSat->stats.conflicts );
//        Abc_PrintTime( 1, "Time", Abc_Clock() - clk );
    }
    assert(RetValue == l_False, "RetValue == l_False");
    // derive the UNSAT core
    clk = Abc_Clock();
    vCore = (Vec_Int_t *)Sat_ProofCore( pSat );
    if vCore != nil {
        Vec_IntSort( vCore, 1 ); }
    if fVerbose != 0 {
//        Abc_Print( 1, "Core is %8d vars    (out of %8d).   ", Vec_IntSize(vCore), sat_solver2_nvars(pSat) );
//        Abc_PrintTime( 1, "Time", Abc_Clock() - clk );
    }
    return vCore;
}


/**Function*************************************************************

  Synopsis    []

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Gla_ManAbsPrintFrame(p *Gla_Man_t, nCoreSize int32, nFrames int32, nConfls int32, nCexes int32, Time abctime) {
    if Abc_FrameIsBatchMode() && nCoreSize <= 0 {
        return; }
    Abc_Print( 1, "%4d :", nFrames-1 );
    Abc_Print( 1, "%4d", Abc_MinInt(100, 100 * Gia_GlaAbsCount(p, 0, 0) / (p.nObjs - Gia_ManPoNum(p.pGia) + Gia_ManCoNum(p.pGia) + 1)) ); 
    Abc_Print( 1, "%6d", Gia_GlaAbsCount(p, 0, 0) );
    Abc_Print( 1, "%5d", Gla_ManCountPPis(p) );
    Abc_Print( 1, "%5d", Gia_GlaAbsCount(p, 1, 0) );
    Abc_Print( 1, "%6d", Gia_GlaAbsCount(p, 0, 1) );
    Abc_Print( 1, "%8d", nConfls );
    if nCexes == 0 {
        Abc_Print( 1, "%5c", '-' ); 
    } else {
        Abc_Print( 1, "%5d", nCexes ); } 
//    Abc_Print( 1, " %9d", sat_solver2_nvars(p->pSat) ); 
    Abc_PrintInt( sat_solver2_nvars(p.pSat) );
    Abc_PrintInt( sat_solver2_nclauses(p.pSat) );
    Abc_PrintInt( sat_solver2_nlearnts(p.pSat) );
//    Abc_Print( 1, " %6d", nCoreSize > 0 ? nCoreSize : 0 ); 
    Abc_Print( 1, "%9.2f sec", 1.0*Time/int64(time.Second) );
    Abc_Print( 1, "%5.0f MB", (sat_solver2_memory_proof(p.pSat) + sat_solver2_memory(p.pSat, 0)) / (1<<20) );
//    Abc_PrintInt( p->nAbsNew );
//    Abc_PrintInt( p->nLrnNew );
//    Abc_Print( 1, "%4.1f MB", 4.0 * p->nLrnNew * Abc_BitWordNum(p->nAbsNew) / (1<<20) );
    Abc_Print( 1, "%s", (nCoreSize > 0 && nCexes > 0) ? "\n" : "\r" );
    fflush( os.Stdout );
}
func Gla_ManReportMemory(p *Gla_Man_t) {
    var pGla *Gla_Obj_t;
    var memTot float64;
    var memAig float64  = Gia_ManObjNum(p.pGia) * sizeof(Gia_Obj_t);
    var memSat float64  = sat_solver2_memory( p.pSat, 1 );
    var memPro float64  = sat_solver2_memory_proof( p.pSat );
    var memMap float64  = p.nObjs * sizeof(Gla_Obj_t) + Gia_ManObjNum(p.pGia) * sizeof(int);
    var memRef float64  = Rnm_ManMemoryUsage( p.pRnm );
    var memOth float64  = sizeof(Gla_Man_t);
    for  pGla = p.pObjs; pGla < p.pObjs + p.nObjs; pGla++  {
        memMap += Vec_IntCap(&pGla.vFrames) * sizeof(int);
    }
    memOth += Vec_IntCap(p.vAddedNew) * sizeof(int);
    memOth += Vec_IntCap(p.vTemp) * sizeof(int);
    memOth += Vec_IntCap(p.vAbs) * sizeof(int);
    memTot = memAig + memSat + memPro + memMap + memRef + memOth;
    // inlining of macro ABC_PRMP
    Abc_Print(1, "%s =", ("Memory: AIG      "));
Abc_Print(1, "%10.3f MB (%6.2f %%)\n",  (1.0*(float64((memAig)))/(1<<20)), ((float64((memTot)))? 100.0*(float64((memAig)))/(float64((memTot))) : 0.0) );;
    // inlining of macro ABC_PRMP
    Abc_Print(1, "%s =", ("Memory: SAT      "));
Abc_Print(1, "%10.3f MB (%6.2f %%)\n",  (1.0*(float64((memSat)))/(1<<20)), ((float64((memTot)))? 100.0*(float64((memSat)))/(float64((memTot))) : 0.0) );;
    // inlining of macro ABC_PRMP
    Abc_Print(1, "%s =", ("Memory: Proof    "));
Abc_Print(1, "%10.3f MB (%6.2f %%)\n",  (1.0*(float64((memPro)))/(1<<20)), ((float64((memTot)))? 100.0*(float64((memPro)))/(float64((memTot))) : 0.0) );;
    // inlining of macro ABC_PRMP
    Abc_Print(1, "%s =", ("Memory: Map      "));
Abc_Print(1, "%10.3f MB (%6.2f %%)\n",  (1.0*(float64((memMap)))/(1<<20)), ((float64((memTot)))? 100.0*(float64((memMap)))/(float64((memTot))) : 0.0) );;
    // inlining of macro ABC_PRMP
    Abc_Print(1, "%s =", ("Memory: Refine   "));
Abc_Print(1, "%10.3f MB (%6.2f %%)\n",  (1.0*(float64((memRef)))/(1<<20)), ((float64((memTot)))? 100.0*(float64((memRef)))/(float64((memTot))) : 0.0) );;
    // inlining of macro ABC_PRMP
    Abc_Print(1, "%s =", ("Memory: Other    "));
Abc_Print(1, "%10.3f MB (%6.2f %%)\n",  (1.0*(float64((memOth)))/(1<<20)), ((float64((memTot)))? 100.0*(float64((memOth)))/(float64((memTot))) : 0.0) );;
    // inlining of macro ABC_PRMP
    Abc_Print(1, "%s =", ("Memory: TOTAL    "));
Abc_Print(1, "%10.3f MB (%6.2f %%)\n",  (1.0*(float64((memTot)))/(1<<20)), ((float64((memTot)))? 100.0*(float64((memTot)))/(float64((memTot))) : 0.0) );;
}


/**Function*************************************************************

  Synopsis    [Send abstracted model or send cancel.]

  Description [Counter-example will be sent automatically when &vta terminates.]
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Gia_GlaSendAbsracted(p *Gla_Man_t, fVerbose int32) {
    var pAbs *Gia_Man_t;
    var vGateClasses *Vec_Int_t;
    assert(Abc_FrameIsBridgeMode(), "Abc_FrameIsBridgeMode()");
//    if ( fVerbose )
//        Abc_Print( 1, "Sending abstracted model...\n" );
    // create abstraction (value of p->pGia is not used here)
    vGateClasses = Gla_ManTranslate( p );
    pAbs = Gia_ManDupAbsGates( p.pGia0, vGateClasses );
    Vec_IntFreeP( &vGateClasses );
    // send it out
    Gia_ManToBridgeAbsNetlist( os.Stdout, pAbs, BRIDGE_ABS_NETLIST );
    Gia_ManStop( pAbs );
}
func Gia_GlaSendCancel(p *Gla_Man_t, fVerbose int32) {
    extern int Gia_ManToBridgeBadAbs( FILE * pFile );
    assert(Abc_FrameIsBridgeMode(), "Abc_FrameIsBridgeMode()");
//    if ( fVerbose )
//        Abc_Print( 1, "Cancelling previously sent model...\n" );
    Gia_ManToBridgeBadAbs( os.Stdout );
}

/**Function*************************************************************

  Synopsis    [Send abstracted model or send cancel.]

  Description [Counter-example will be sent automatically when &vta terminates.]
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Gia_GlaDumpAbsracted(p *Gla_Man_t, fVerbose int32) {
    var pFileNameDef string  = "glabs.aig";
    var pFileName string;
    if p.pPars.pFileVabs {
        pFileName = p.pPars.pFileVabs;
    } else {
        pFileName = pFileNameDef;
    }
    var pAbs *Gia_Man_t;
    var vGateClasses *Vec_Int_t;
    if fVerbose != 0 {
        Abc_Print( 1, "Dumping abstracted model into file \"%s\"...\n", pFileName ); }
    // create abstraction
    vGateClasses = Gla_ManTranslate( p );
    pAbs = Gia_ManDupAbsGates( p.pGia0, vGateClasses );
    Vec_IntFreeP( &vGateClasses );
    // write into file
    Gia_AigerWrite( pAbs, pFileName, 0, 0, 0 );
    Gia_ManStop( pAbs );
}


/**Function*************************************************************

  Synopsis    [Performs gate-level abstraction]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Gia_ManPerformGlaOld(pAig *Gia_Man_t, pPars *Abs_Par_t, fStartVta int32) int32 {
    extern int Gia_VtaPerformInt( Gia_Man_t * pAig, Abs_Par_t * pPars );
    extern void Ga2_ManDumpStats( Gia_Man_t * pGia, Abs_Par_t * pPars, sat_solver2 * pSat, int iFrame, int fUseN );
    var p *Gla_Man_t;
    var vPPis, vCore *Vec_Int_t;//, * vCore2 = NULL;
    var pCex *Abc_Cex_t  = nil;
    var f int32;
    var i int32;
    var iPrev int32;
    var nConfls int32;
    var Status int32;
    var nVarsOld int32;
    var nCoreSize int32;
    var fOneIsSent int32;
    var RetValue int32 = -1;
    var clk2 abctime;
    var clk abctime = Abc_Clock();
    // preconditions
    assert(Gia_ManPoNum(pAig) == 1, "Gia_ManPoNum(pAig) == 1");
    assert(pPars.nFramesMax == 0 || pPars.nFramesStart <= pPars.nFramesMax, "pPars.nFramesMax == 0 || pPars.nFramesStart <= pPars.nFramesMax");
    if Gia_ObjIsConst0(Gia_ObjFanin0(Gia_ManPo(pAig,0))) {
        if !Gia_ObjFaninC0(Gia_ManPo(pAig,0)) {
            fmt.Printf("Sequential miter is trivially UNSAT.\n");
            return 1;
        }
        pAig.pCexSeq = nil;
        pAig.pCexSeq = Abc_CexMakeTriv( Gia_ManRegNum(pAig), Gia_ManPiNum(pAig), 1, 0 );
        fmt.Printf("Sequential miter is trivially SAT.\n");
        return 0;
    }

    // compute intial abstraction
    if pAig.vGateClasses == nil {
        if fStartVta != 0 {
            var nFramesMaxOld int32    = pPars.nFramesMax;
            var nFramesStartOld int32  = pPars.nFramesStart;
            var nTimeOutOld int32      = pPars.nTimeOut;
            var nDumpOld int32         = pPars.fDumpVabs;
            pPars.nFramesMax   = pPars.nFramesStart;
            pPars.nFramesStart = Abc_MinInt( pPars.nFramesStart/2 + 1, 3 );
            pPars.nTimeOut     = 20;
            pPars.fDumpVabs    = 0;
            RetValue = Gia_VtaPerformInt( pAig, pPars );
            pPars.nFramesMax   = nFramesMaxOld;
            pPars.nFramesStart = nFramesStartOld;
            pPars.nTimeOut     = nTimeOutOld;
            pPars.fDumpVabs    = nDumpOld;
            // create gate classes
            Vec_IntFreeP( &pAig.vGateClasses );
            if pAig.vObjClasses {
                pAig.vGateClasses = Gia_VtaConvertToGla( pAig, pAig.vObjClasses ); }
            Vec_IntFreeP( &pAig.vObjClasses );
            // return if VTA solve the problem if could not start
            if RetValue == 0 || pAig.vGateClasses == nil {
                return RetValue; }
        
        } else {
            pAig.vGateClasses = Vec_IntStart( Gia_ManObjNum(pAig) );
            Vec_IntWriteEntry( pAig.vGateClasses, 0, 1 );
            Vec_IntWriteEntry( pAig.vGateClasses, Gia_ObjFaninId0p(pAig, Gia_ManPo(pAig, 0)), 1 );
        }
    }
    // start the manager
    p = Gla_ManStart( pAig, pPars );
    p.timeInit = Abc_Clock() - clk;
    // set runtime limit
    if p.pPars.nTimeOut {
        sat_solver2_set_runtime_limit( p.pSat, p.pPars.nTimeOut * int64(time.Second) + Abc_Clock() ); }
    // perform initial abstraction
    if p.pPars.fVerbose {
        Abc_Print( 1, "Running gate-level abstraction (GLA) with the following parameters:\n" );
        Abc_Print( 1, "FrameMax = %d  ConfMax = %d  Timeout = %d  RatioMin = %d %%.\n", 
            pPars.nFramesMax, pPars.nConfLimit, pPars.nTimeOut, pPars.nRatioMin );
        Abc_Print( 1, "LearnStart = %d  LearnDelta = %d  LearnRatio = %d %%.\n", 
            pPars.nLearnedStart, pPars.nLearnedDelta, pPars.nLearnedPerce );
        Abc_Print( 1, " Frame   %%   Abs  PPI   FF   LUT   Confl  Cex   Vars   Clas   Lrns     Time        Mem\n" );
    }
    for  f = i = iPrev = 0; !p.pPars.nFramesMax || f < p.pPars.nFramesMax; f, iPrev = f+1, i  {
        var nConflsBeg int32  = sat_solver2_nconflicts(p.pSat);
        p.pPars.iFrame = f;

        // load timeframe
        Gia_GlaAddTimeFrame( p, f );

        // iterate as long as there are counter-examples
        for  i = 0; ; i++  { 
            clk2 = Abc_Clock();
            vCore = Gla_ManUnsatCore( p, f, p.pSat, pPars.nConfLimit, pPars.fVerbose, &Status, &nConfls );
//            assert( (vCore != NULL) == (Status == 1) );
            if Status == -1 || (p.pSat.nRuntimeLimit && Abc_Clock() > p.pSat.nRuntimeLimit) { // resource limit is reached
            
                Prf_ManStopP( &p.pSat.pPrf2 );
//                if ( Gia_ManRegNum(p->pGia) > 1 ) // for comb cases, return the abstraction
//                    Vec_IntShrink( p->vAbs, p->nAbsOld );
                goto finish;
            }
            if Status == 1 {
                Prf_ManStopP( &p.pSat.pPrf2 );
                p.timeUnsat += Abc_Clock() - clk2;
                break;
            } 
            p.timeSat += Abc_Clock() - clk2;
            assert(Status == 0, "Status == 0");
            p.nCexes++;

            // cancel old one if it was sent
            if Abc_FrameIsBridgeMode() && (fOneIsSent != 0) {
                Gia_GlaSendCancel( p, pPars.fVerbose );
                fOneIsSent = 0;
            }

            // perform the refinement
            clk2 = Abc_Clock();
            if pPars.fAddLayer {
                vPPis = Gla_ManCollectPPis( p, nil );
//                Gla_ManExplorePPis( p, vPPis );
            
            } else {
                vPPis = Gla_ManRefinement( p );
                if vPPis == nil {
                    Prf_ManStopP( &p.pSat.pPrf2 );
                    pCex = p.pGia.pCexSeq; p.pGia.pCexSeq = nil;
                    break;
                }
            } 
            assert(pCex == nil, "pCex == nil");

            // start proof logging
            if i == 0 {
                // create bookmark to be used for rollback
                sat_solver2_bookmark( p.pSat );
                Vec_IntClear( p.vAddedNew );
                p.nAbsOld = Vec_IntSize( p.vAbs );
                nVarsOld = p.nSatVars;
//                p->nLrnOld = sat_solver2_nlearnts( p->pSat );
//                p->nAbsNew = 0;
//                p->nLrnNew = 0;

                // start incremental proof manager
                assert(p.pSat.pPrf2 == nil, "p.pSat.pPrf2 == nil");
                if p.pSat.pPrf1 == nil {
                    p.pSat.pPrf2 = Prf_ManAlloc(); }
                if p.pSat.pPrf2 {
                    p.nProofIds = 0;
                    Vec_IntFill( p.vProofIds, Gia_ManObjNum(p.pGia), -1 );
                    Prf_ManRestart( p.pSat.pPrf2, p.vProofIds, sat_solver2_nlearnts(p.pSat), Vec_IntSize(vPPis) );
                }
            
            } else {
                // resize the proof logger
                if p.pSat.pPrf2 {
                    Prf_ManGrow( p.pSat.pPrf2, p.nProofIds + Vec_IntSize(vPPis) ); }
            }

            Gia_GlaAddToAbs( p, vPPis, 1 );
            Gia_GlaAddOneSlice( p, f, vPPis );
            Vec_IntFree( vPPis );

            // print the result (do not count it towards change)
            if p.pPars.fVerbose {
            Gla_ManAbsPrintFrame( p, -1, f+1, sat_solver2_nconflicts(p.pSat)-nConflsBeg, i, Abc_Clock() - clk ); }
        }
        if pCex != nil {
            break; }
        assert(Status == 1, "Status == 1");

        // valid core is obtained
        nCoreSize = 1;
        if vCore != nil {
            nCoreSize += Vec_IntSize( vCore );
            Gia_GlaAddToCounters( p, vCore );
        }
        if i == 0 {
            p.pPars.nFramesNoChange++;
            Vec_IntFreeP( &vCore );
        
        } else {
            p.pPars.nFramesNoChange = 0;
//            p->nAbsNew = Vec_IntSize( p->vAbs ) - p->nAbsOld;
//            p->nLrnNew = Abc_AbsInt( sat_solver2_nlearnts( p->pSat ) - p->nLrnOld );
            // update the SAT solver
            sat_solver2_rollback( p.pSat );
            // update storage
            Gla_ManRollBack( p );
            p.nSatVars = nVarsOld;
            // load this timeframe
            Gia_GlaAddToAbs( p, vCore, 0 );
            Gia_GlaAddOneSlice( p, f, vCore );
            Vec_IntFree( vCore );
            // run SAT solver
            clk2 = Abc_Clock();
            vCore = Gla_ManUnsatCore( p, f, p.pSat, pPars.nConfLimit, p.pPars.fVerbose, &Status, &nConfls );
            p.timeUnsat += Abc_Clock() - clk2;
//            assert( (vCore != NULL) == (Status == 1) );
            Vec_IntFreeP( &vCore );
            if Status == -1 { // resource limit is reached
                break; }
            if Status == 0 {
                assert(false, "0");
    //            Vta_ManSatVerify( p );
                // make sure, there was no initial abstraction (otherwise, it was invalid)
                assert(pAig.vObjClasses == nil && f < p.pPars.nFramesStart, "pAig.vObjClasses == nil && f < p.pPars.nFramesStart");
    //            pCex = Vga_ManDeriveCex( p );
                break;
            }
        }
        // print the result
        if p.pPars.fVerbose {
        Gla_ManAbsPrintFrame( p, nCoreSize, f+1, sat_solver2_nconflicts(p.pSat)-nConflsBeg, i, Abc_Clock() - clk ); }

        if f > 2 && iPrev > 0 && i == 0 { // change has happened
        
            if Abc_FrameIsBridgeMode() {
                // cancel old one if it was sent
                if fOneIsSent != 0 {
                    Gia_GlaSendCancel( p, pPars.fVerbose ); }
                // send new one 
                Gia_GlaSendAbsracted( p, pPars.fVerbose );
                fOneIsSent = 1;
            }

            // dump the model into file
            if p.pPars.fDumpVabs {
                var Command string // was [1000]byte in C

                Abc_FrameSetStatus( -1 );
                Abc_FrameSetCex( nil );
                Abc_FrameSetNFrames( f+1 );
                Command = fmt.Sprintf("write_status %s", Extra_FileNameGenericAppend((char *)(p.pPars.pFileVabs ? p.pPars.pFileVabs : "glabs.aig"), ".status"));
                Cmd_CommandExecute( Abc_FrameGetGlobalFrame(), Command );
                Gia_GlaDumpAbsracted( p, pPars.fVerbose );
            }
        }

        // check if the number of objects is below limit
        if Gia_GlaAbsCount(p,0,0) >= (p.nObjs - 1) * (100 - pPars.nRatioMin) / 100 {
            Status = -1;
            break;
        }
    }
finish:
    // analize the results
    if pCex == nil {
        if p.pPars.fVerbose && Status == -1 {
            fmt.Printf("\n"); }
//        if ( pAig->vGateClasses != NULL )
//            Abc_Print( 1, "Replacing the old abstraction by a new one.\n" );
        Vec_IntFreeP( &pAig.vGateClasses );
        pAig.vGateClasses = Gla_ManTranslate( p );
        if Status == -1 {
            if p.pPars.nTimeOut && Abc_Clock() >= p.pSat.nRuntimeLimit { 
                Abc_Print( 1, "Timeout %d sec in frame %d with a %d-stable abstraction.    ", p.pPars.nTimeOut, f, p.pPars.nFramesNoChange );
            } else if pPars.nConfLimit && sat_solver2_nconflicts(p.pSat) >= pPars.nConfLimit {
                Abc_Print( 1, "Exceeded %d conflicts in frame %d with a %d-stable abstraction.  ", pPars.nConfLimit, f, p.pPars.nFramesNoChange );
            } else if Gia_GlaAbsCount(p,0,0) >= (p.nObjs - 1) * (100 - pPars.nRatioMin) / 100 {
                Abc_Print( 1, "The ratio of abstracted objects is less than %d %% in frame %d.  ", pPars.nRatioMin, f );
            } else {
                Abc_Print( 1, "Abstraction stopped for unknown reason in frame %d.  ", f ); }
        
        } else {
            p.pPars.iFrame++;
            Abc_Print( 1, "GLA completed %d frames with a %d-stable abstraction.  ", f, p.pPars.nFramesNoChange );
        }
    
    } else {
        if p.pPars.fVerbose {
            fmt.Printf("\n"); }
        pAig.pCexSeq = nil;
        pAig.pCexSeq = pCex;
        if !Gia_ManVerifyCex( pAig, pCex, 0 ) {
            Abc_Print( 1, "    Gia_ManPerformGlaOld(): CEX verification has failed!\n" ); }
        Abc_Print( 1, "Counter-example detected in frame %d.  ", f );
        p.pPars.iFrame = pCex.iFrame - 1;
        Vec_IntFreeP( &pAig.vGateClasses );
        RetValue = 0;
    }
    Abc_PrintTime( 1, "Time", Abc_Clock() - clk );
    if p.pPars.fVerbose {
        p.timeOther = (Abc_Clock() - clk) - p.timeUnsat - p.timeSat - p.timeCex - p.timeInit;
        // inlining of macro ABC_PRTP
        Abc_Print(1, "%s =", ("Runtime: Initializing"));
Abc_Print(1, "%9.2f sec (%6.2f %%)\n", 1.0*(float64((p.timeInit)))/(float64(int64(time.Second))), (float64((Abc_Clock() - clk)))? 100.0*(float64((p.timeInit)))/(float64((Abc_Clock() - clk))) : 0.0);;
        // inlining of macro ABC_PRTP
        Abc_Print(1, "%s =", ("Runtime: Solver UNSAT"));
Abc_Print(1, "%9.2f sec (%6.2f %%)\n", 1.0*(float64((p.timeUnsat)))/(float64(int64(time.Second))), (float64((Abc_Clock() - clk)))? 100.0*(float64((p.timeUnsat)))/(float64((Abc_Clock() - clk))) : 0.0);;
        // inlining of macro ABC_PRTP
        Abc_Print(1, "%s =", ("Runtime: Solver SAT  "));
Abc_Print(1, "%9.2f sec (%6.2f %%)\n", 1.0*(float64((p.timeSat)))/(float64(int64(time.Second))), (float64((Abc_Clock() - clk)))? 100.0*(float64((p.timeSat)))/(float64((Abc_Clock() - clk))) : 0.0);;
        // inlining of macro ABC_PRTP
        Abc_Print(1, "%s =", ("Runtime: Refinement  "));
Abc_Print(1, "%9.2f sec (%6.2f %%)\n", 1.0*(float64((p.timeCex)))/(float64(int64(time.Second))), (float64((Abc_Clock() - clk)))? 100.0*(float64((p.timeCex)))/(float64((Abc_Clock() - clk))) : 0.0);;
        // inlining of macro ABC_PRTP
        Abc_Print(1, "%s =", ("Runtime: Other       "));
Abc_Print(1, "%9.2f sec (%6.2f %%)\n", 1.0*(float64((p.timeOther)))/(float64(int64(time.Second))), (float64((Abc_Clock() - clk)))? 100.0*(float64((p.timeOther)))/(float64((Abc_Clock() - clk))) : 0.0);;
        // inlining of macro ABC_PRTP
        Abc_Print(1, "%s =", ("Runtime: TOTAL       "));
Abc_Print(1, "%9.2f sec (%6.2f %%)\n", 1.0*(float64((Abc_Clock() - clk)))/(float64(int64(time.Second))), (float64((Abc_Clock() - clk)))? 100.0*(float64((Abc_Clock() - clk)))/(float64((Abc_Clock() - clk))) : 0.0);;
        Gla_ManReportMemory( p );
    }
//    Ga2_ManDumpStats( pAig, p->pPars, p->pSat, p->pPars->iFrame, 1 );
    Gla_ManStop( p );
    fflush( os.Stdout );
    return RetValue;
}

////////////////////////////////////////////////////////////////////////
///                       END OF FILE                                ///
////////////////////////////////////////////////////////////////////////


//ABC_NAMESPACE_IMPL_END

