package abs

/**CFile****************************************************************

  FileName    [absRef.h]

  SystemName  [ABC: Logic synthesis and verification system.]

  PackageName [Abstraction package.]

  Synopsis    [Refinement manager.]

  Author      [Alan Mishchenko]
  
  Affiliation [UC Berkeley]

  Date        [Ver. 1.0. Started - June 20, 2005.]

  Revision    [$Id: absRef.h,v 1.00 2005/06/20 00:00:00 alanmi Exp $]

***********************************************************************/
 




////////////////////////////////////////////////////////////////////////
///                          INCLUDES                                ///
////////////////////////////////////////////////////////////////////////

////////////////////////////////////////////////////////////////////////
///                         PARAMETERS                               ///
////////////////////////////////////////////////////////////////////////

//ABC_NAMESPACE_HEADER_START


////////////////////////////////////////////////////////////////////////
///                         BASIC TYPES                              ///
////////////////////////////////////////////////////////////////////////


////////////////////////////////////////////////////////////////////////
///                      MACRO DEFINITIONS                           ///
////////////////////////////////////////////////////////////////////////

    // refinement object
type Rnm_Obj_t struct {
    Value uint32; /* C bitfield: 1 bits; review packing and truncation */  // binary value
    fVisit uint32; /* C bitfield: 1 bits; review packing and truncation */  // visited object
    fVisitJ uint32; /* C bitfield: 1 bits; review packing and truncation */  // justified visited object
    fPPi uint32; /* C bitfield: 1 bits; review packing and truncation */  // PPI object
    Prio uint32; /* C bitfield: 24 bits; review packing and truncation */  // priority (0 - highest)
}

    // refinement manager
type Rnm_Man_t struct {
    // user data
    pGia      *Gia_Man_t;            // working AIG manager (it is completely owned by this package)
    pCex      *Abc_Cex_t;            // counter-example
    vMap      *Vec_Int_t;            // mapping of CEX inputs into objects (PI + PPI, in any order)
    fPropFanout             int32;     // propagate fanouts
    fVerbose             int32;        // verbose flag
    nRefId             int32;          // refinement ID
    // traversing data
    vObjs      *Vec_Int_t;           // internal objects used in value propagation
    // filtering of selected objects
    vCounts      *Vec_Str_t;         // fanin counters
    vFanins      *Vec_Int_t;         // fanins
/*
    // SAT solver
    sat_solver2 *   pSat;            // incremental SAT solver
    Vec_Int_t *     vSatVars;        // SAT variables
    Vec_Int_t *     vSat2Ids;        // mapping of SAT variables into object IDs
    Vec_Int_t *     vIsopMem;        // memory for ISOP computation
*/
    // internal data
    pObjs      *Rnm_Obj_t;           // refinement objects
    nObjs             int32;           // the number of used objects
    nObjsAlloc             int32;      // the number of allocated objects
    nObjsFrame             int32;      // the number of used objects in each frame
    nCalls             int32;          // total number of calls
    nRefines             int32;        // total refined objects
    nVisited             int32;        // visited during justification
    // statistics  
    timeFwd         abctime;         // forward propagation
    timeBwd         abctime;         // backward propagation
    timeVer         abctime;         // ternary simulation
    timeTotal         abctime;       // other time
}

// accessing the refinement object
func Rnm_ManObj(p *Rnm_Man_t, pObj *Gia_Obj_t, f int32) *Rnm_Obj_t { 
    assert(Gia_ObjIsConst0(pObj) || pObj.Value, "Gia_ObjIsConst0(pObj) || pObj.Value");
    assert(int32(pObj.Value) < p.nObjsFrame, "int32(pObj.Value) < p.nObjsFrame");
    assert(f >= 0 && f <= p.pCex.iFrame, "f >= 0 && f <= p.pCex.iFrame"); 
    return p.pObjs + f * p.nObjsFrame + pObj.Value;  
}
func Rnm_ManSetRefId(p *Rnm_Man_t, RefId int32) { p.nRefId = RefId; }

func Rnm_ObjCount(p *Rnm_Man_t, pObj *Gia_Obj_t) int32 { return Vec_StrEntry( p.vCounts, Gia_ObjId(p.pGia, pObj) );                           }
func Rnm_ObjSetCount(p *Rnm_Man_t, pObj *Gia_Obj_t, c int32) { Vec_StrWriteEntry( p.vCounts, Gia_ObjId(p.pGia, pObj), byte(c) );                    }
func Rnm_ObjAddToCount(p *Rnm_Man_t, pObj *Gia_Obj_t) int32 { var c int32  = Rnm_ObjCount(p, pObj); if c < 16 {  Rnm_ObjSetCount(p, pObj, c+1); }
 return c; }

func Rnm_ObjIsJust(p *Rnm_Man_t, pObj *Gia_Obj_t) int32 { return Gia_ObjIsConst0(pObj) || (pObj.Value && Rnm_ManObj(p, pObj, 0).fVisitJ);      }

////////////////////////////////////////////////////////////////////////
///                    FUNCTION DECLARATIONS                         ///
////////////////////////////////////////////////////////////////////////

/*=== absRef.c ===========================================================*/
// extern Rnm_Man_t *  Rnm_ManStart( Gia_Man_t * pGia );

// extern void         Rnm_ManStop( Rnm_Man_t * p, int fProfile );

// extern double       Rnm_ManMemoryUsage( Rnm_Man_t * p );

// extern Vec_Int_t *  Rnm_ManRefine( Rnm_Man_t * p, Abc_Cex_t * pCex, Vec_Int_t * vMap, int fPropFanout, int fNewRefinement, int fVerbose );

/*=== absRefSelected.c ===========================================================*/
// extern Vec_Int_t *  Rnm_ManFilterSelected( Rnm_Man_t * p, Vec_Int_t * vOldPPis );

// extern Vec_Int_t *  Rnm_ManFilterSelectedNew( Rnm_Man_t * p, Vec_Int_t * vOldPPis );



//ABC_NAMESPACE_HEADER_END



////////////////////////////////////////////////////////////////////////
///                       END OF FILE                                ///
////////////////////////////////////////////////////////////////////////

