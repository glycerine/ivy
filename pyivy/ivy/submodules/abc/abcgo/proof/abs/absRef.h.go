package abs

import (
	"time"

	"github.com/glycerine/abcgo/aig/gia"
	"github.com/glycerine/abcgo/misc/util"
	"github.com/glycerine/abcgo/misc/vec"
)

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
	Value   uint32 /* C bitfield: 1 bits; review packing and truncation */  // binary value
	fVisit  uint32 /* C bitfield: 1 bits; review packing and truncation */  // visited object
	fVisitJ uint32 /* C bitfield: 1 bits; review packing and truncation */  // justified visited object
	fPPi    uint32 /* C bitfield: 1 bits; review packing and truncation */  // PPI object
	Prio    uint32 /* C bitfield: 24 bits; review packing and truncation */ // priority (0 - highest)
}

// refinement manager
type Rnm_Man_t struct {
	// user data
	pGia        *gia.Gia_Man_t  // working AIG manager (it is completely owned by this package)
	pCex        *util.Abc_Cex_t // counter-example
	vMap        *vec.Vec_Int_t  // mapping of CEX inputs into objects (PI + PPI, in any order)
	fPropFanout int32           // propagate fanouts
	fVerbose    int32           // verbose flag
	nRefId      int32           // refinement ID
	// traversing data
	vObjs *vec.Vec_Int_t // internal objects used in value propagation
	// filtering of selected objects
	vCounts *vec.Vec_Str_t // fanin counters
	vFanins *vec.Vec_Int_t // fanins
	/*
	   // SAT solver
	   sat_solver2 *   pSat;            // incremental SAT solver
	   Vec_Int_t *     vSatVars;        // SAT variables
	   Vec_Int_t *     vSat2Ids;        // mapping of SAT variables into object IDs
	   Vec_Int_t *     vIsopMem;        // memory for ISOP computation
	*/
	// internal data
	pObjs      []Rnm_Obj_t // refinement objects
	nObjs      int32       // the number of used objects
	nObjsAlloc int32       // the number of allocated objects
	nObjsFrame int32       // the number of used objects in each frame
	nCalls     int32       // total number of calls
	nRefines   int32       // total refined objects
	nVisited   int32       // visited during justification
	// statistics
	timeFwd   time.Duration // forward propagation
	timeBwd   time.Duration // backward propagation
	timeVer   time.Duration // ternary simulation
	timeTotal time.Duration // other time
}

// accessing the refinement object
func Rnm_ManObj(p *Rnm_Man_t, pObj *gia.Gia_Obj_t, f int32) *Rnm_Obj_t {
	assert(gia.Gia_ObjIsConst0(pObj) != 0 || pObj.Value != 0, "Gia_ObjIsConst0(pObj) || pObj.Value")
	assert(int32(pObj.Value) < p.nObjsFrame, "int32(pObj.Value) < p.nObjsFrame")
	assert(f >= 0 && f <= p.pCex.IFrame, "f >= 0 && f <= p.pCex.iFrame")
	return &p.pObjs[int(f*p.nObjsFrame)+int(pObj.Value)]
}
func Rnm_ManSetRefId(p *Rnm_Man_t, RefId int32) { p.nRefId = RefId }

func Rnm_ObjCount(p *Rnm_Man_t, pObj *gia.Gia_Obj_t) int32 {
	return int32(vec.Vec_StrEntry(p.vCounts, int(gia.Gia_ObjId(p.pGia, pObj))))
}
func Rnm_ObjSetCount(p *Rnm_Man_t, pObj *gia.Gia_Obj_t, c int32) {
	vec.Vec_StrWriteEntry(p.vCounts, int(gia.Gia_ObjId(p.pGia, pObj)), byte(c))
}
func Rnm_ObjAddToCount(p *Rnm_Man_t, pObj *gia.Gia_Obj_t) int32 {
	var c int32 = Rnm_ObjCount(p, pObj)
	if c < 16 {
		Rnm_ObjSetCount(p, pObj, c+1)
	}
	return c
}

func Rnm_ObjIsJust(p *Rnm_Man_t, pObj *gia.Gia_Obj_t) int32 {
	if gia.Gia_ObjIsConst0(pObj) != 0 || (pObj.Value != 0 && Rnm_ManObj(p, pObj, 0).fVisitJ != 0) {
		return 1
	}
	return 0
}

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
