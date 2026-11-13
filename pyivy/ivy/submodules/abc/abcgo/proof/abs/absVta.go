package abs

import (
	"fmt"
	"os"
	"sort"
	"time"
	"unsafe"

	"github.com/glycerine/abcgo/aig/gia"
	abcmain "github.com/glycerine/abcgo/base/main"
	"github.com/glycerine/abcgo/misc/extra"
	"github.com/glycerine/abcgo/misc/util"
	"github.com/glycerine/abcgo/misc/vec"
	"github.com/glycerine/abcgo/sat/bsat"
)

/**CFile****************************************************************

  FileName    [absVta.c]

  SystemName  [ABC: Logic synthesis and verification system.]

  PackageName [Abstraction package.]

  Synopsis    [Variable time-frame abstraction.]

  Author      [Alan Mishchenko]

  Affiliation [UC Berkeley]

  Date        [Ver. 1.0. Started - June 20, 2005.]

  Revision    [$Id: absVta.c,v 1.00 2005/06/20 00:00:00 alanmi Exp $]

***********************************************************************/

//#include "sat/bsat/satSolver2.h"
//#include "base/main/main.h"
//#include "abs.h"

//ABC_NAMESPACE_IMPL_START

const VTA_LARGE = 0xFFFFFFF

////////////////////////////////////////////////////////////////////////
///                        DECLARATIONS                              ///
////////////////////////////////////////////////////////////////////////

// manager
type Vta_Man_t struct {
	// user data
	pGia  *gia.Gia_Man_t // AIG manager
	pPars *Abs_Par_t     // parameters
	// internal data
	nObjs      int32           // the number of objects
	nObjsAlloc int32           // the number of objects allocated
	nBins      int32           // number of hash table entries
	objects    *vtaObjectTable // stable objects and hash table
	vOrder     *vec.Vec_Int_t  // objects in DPS order
	// abstraction
	nObjBits  int32          // the number of bits to represent objects
	nObjMask  uint32         // object mask
	vFrames   *vec.Vec_Ptr_t // start abstraction for each frame
	nWords    int32          // the number of words in the record
	nCexes    int32          // the number of CEXes
	nObjAdded int32          // objects added to the abstraction
	vSeens    *vec.Vec_Int_t // seen objects
	vSeenGla  *vec.Vec_Bit_t // seen objects in all frames
	nSeenGla  int32          // seen objects in all frames
	nSeenAll  int32          // seen objects in all frames
	// other data
	vCores    *vec.Vec_Ptr_t // unsat core for each frame
	deadline  time.Time
	pSat      *bsat.LabeledSolver2 // incremental SAT solver
	vAddedNew *vec.Vec_Int_t       // the IDs of variables added to the solver
	// statistics
	timeSat   int64
	timeUnsat int64
	timeCex   int64
	timeOther int64
}

// ternary simulation

const VTA_VAR0 = 1
const VTA_VAR1 = 2
const VTA_VARX = 3

func Vta_ValIs0(pThis *Vta_Obj_t, fCompl int32) int32 {
	if pThis.Value == VTA_VAR1 && (fCompl != 0) {
		return 1
	}
	if pThis.Value == VTA_VAR0 && !(fCompl != 0) {
		return 1
	}
	return 0
}
func Vta_ValIs1(pThis *Vta_Obj_t, fCompl int32) int32 {
	if pThis.Value == VTA_VAR0 && (fCompl != 0) {
		return 1
	}
	if pThis.Value == VTA_VAR1 && !(fCompl != 0) {
		return 1
	}
	return 0
}

func Vta_ManObj(p *Vta_Man_t, i int32) *Vta_Obj_t { return p.objects.object(i) }
func Vta_ObjId(p *Vta_Man_t, object *Vta_Obj_t) int32 {
	if object == nil || object.id <= 0 || object.id >= p.nObjs || p.objects.object(object.id) != object {
		panic("invalid VTA object")
	}
	return object.id
}

// Original macro:
// #define Vta_ManForEachObj( p, pObj, i )                                 \
//     for ( i = 1; (i < p->nObjs) && ((pObj) = Vta_ManObj(p, i)); i++ )

// Original macro:
// #define Vta_ManForEachObjObj( p, pObjVta, pObjGia, i )                  \
//     for ( i = 1; (i < p->nObjs) && ((pObjVta) = Vta_ManObj(p, i)) && ((pObjGia) = gia.Gia_ManObj(p->pGia, pObjVta->iObj)); i++ )

// Original macro:
// #define Vta_ManForEachObjObjReverse( p, pObjVta, pObjGia, i )           \
//     for ( i = vec.Vec_IntSize(vVec) - 1; (i >= 1) && ((pObjVta) = Vta_ManObj(p, i)) && ((pObjGia) = gia.Gia_ManObj(p->pGia, pObjVta->iObj)); i++ )

// Original macro:
// #define Vta_ManForEachObjVec( vVec, p, pObj, i )                        \
//     for ( i = 0; (i < vec.Vec_IntSize(vVec)) && ((pObj) = Vta_ManObj(p, vec.Vec_IntEntry(vVec,i))); i++ )

// Original macro:
// #define Vta_ManForEachObjVecReverse( vVec, p, pObj, i )                 \
//     for ( i = vec.Vec_IntSize(vVec) - 1; (i >= 0) && ((pObj) = Vta_ManObj(p, vec.Vec_IntEntry(vVec,i))); i-- )

// Original macro:
// #define Vta_ManForEachObjObjVec( vVec, p, pObj, pObjG, i )              \
//     for ( i = 0; (i < vec.Vec_IntSize(vVec)) && ((pObj) = Vta_ManObj(p, vec.Vec_IntEntry(vVec,i))) && ((pObjG) = gia.Gia_ManObj(p->pGia, pObj->iObj)); i++ )

// Original macro:
// #define Vta_ManForEachObjObjVecReverse( vVec, p, pObj, pObjG, i )       \
//     for ( i = vec.Vec_IntSize(vVec) - 1; (i >= 0) && ((pObj) = Vta_ManObj(p, vec.Vec_IntEntry(vVec,i))) && ((pObjG) = gia.Gia_ManObj(p->pGia, pObj->iObj)); i-- )

// abstraction is given as an array of integers:
// - the first entry is the number of timeframes (F)
// - the next (F+1) entries give the beginning position of each timeframe
// - the following entries give the object IDs
// invariant:  assert( vec[vec[0]+1] == size(vec) );

// extern void Vga_ManAddClausesOne( Vta_Man_t * p, int iObj, int iFrame );

////////////////////////////////////////////////////////////////////////
///                     FUNCTION DEFINITIONS                         ///
////////////////////////////////////////////////////////////////////////

/**Function*************************************************************

  Synopsis    [Converting from one array to per-frame arrays.]

  Description []

  SideEffects []

  SeeAlso     []

***********************************************************************/
/**Function*************************************************************

  Synopsis    [Detects how many frames are completed.]

  Description []

  SideEffects []

  SeeAlso     []

***********************************************************************/
/**Function*************************************************************

  Synopsis    []

  Description []

  SideEffects []

  SeeAlso     []

***********************************************************************/
func Vga_ManHash(object, frame, bins int32) int32              { return int32(vtaHash(object, frame, int(bins))) }
func Vga_ManLookup(p *Vta_Man_t, object, frame int32) *int32   { return p.objects.lookup(object, frame) }
func Vga_ManFind(p *Vta_Man_t, object, frame int32) *Vta_Obj_t { return p.objects.find(object, frame) }
func Vga_ManFindOrAdd(p *Vta_Man_t, object, frame int32) *Vta_Obj_t {
	result := p.objects.findOrAdd(object, frame)
	p.nObjs = int32(len(p.objects.objects))
	p.nObjsAlloc = int32(cap(p.objects.objects))
	p.nBins = int32(len(p.objects.bins))
	return result
}
func Vga_ManDelete(p *Vta_Man_t, object, frame int32) { p.objects.delete(object, frame) }

/*
*Function*************************************************************

	Synopsis    [Derives counter-example using current assignments.]

	Description []

	SideEffects []

	SeeAlso     []

**********************************************************************
*/
func Vga_ManDeriveCex(p *Vta_Man_t) *util.Abc_Cex_t {
	var pCex *util.Abc_Cex_t
	var pThis *Vta_Obj_t
	var pObj *gia.Gia_Obj_t
	var i int32
	pCex = util.Abc_CexAlloc(gia.Gia_ManRegNum(p.pGia), gia.Gia_ManPiNum(p.pGia), p.pPars.iFrame+1)
	pCex.IPo = 0
	pCex.IFrame = p.pPars.iFrame
	// inlining of macro Vta_ManForEachObjObj
	for i = 1; i < p.nObjs; i++ {
		pThis = Vta_ManObj(p, i)
		pObj = gia.Gia_ManObj(p.pGia, pThis.iObj)
		if (gia.Gia_ObjIsPi(p.pGia, pObj) != 0) && vtaSolverValue(p.pSat, Vta_ObjId(p, pThis)) {
			util.Abc_InfoSetBit(pCex.PData, pCex.NRegs+pThis.iFrame*pCex.NPis+gia.Gia_ObjCioId(pObj))
		}
	}
	return pCex
}

/*
*Function*************************************************************

	Synopsis    [Remaps core into frame/node pairs.]

	Description []

	SideEffects []

	SeeAlso     []

**********************************************************************
*/
func Vta_ManUnsatCoreRemap(p *Vta_Man_t, vCore *vec.Vec_Int_t) {
	for i, entry := range vec.Vec_IntArray(vCore) {
		object := Vta_ManObj(p, entry)
		if object == nil {
			panic("invalid VTA core object")
		}
		vec.Vec_IntWriteEntry(vCore, i, (object.iFrame<<p.nObjBits)|object.iObj)
	}
}

/*
*Function*************************************************************

	Synopsis    [Compares two objects by their distance.]

	Description []

	SideEffects []

	SeeAlso     []

**********************************************************************
*/
func Vta_ManComputeDepthIncrease(pp1 **Vta_Obj_t, pp2 **Vta_Obj_t) int32 {
	a, b := *pp1, *pp2
	if a.Prio < b.Prio {
		return -1
	}
	if a.Prio > b.Prio {
		return 1
	}
	if a.id < b.id {
		return -1
	}
	if a.id > b.id {
		return 1
	}
	return 0
}

/*
*Function*************************************************************

	Synopsis    [Returns 1 if the object is already used.]

	Description []

	SideEffects []

	SeeAlso     []

**********************************************************************
*/
func Vta_ManObjIsUsed(p *Vta_Man_t, iObj int32) int32 {
	start := int(p.nWords * iObj)
	words := vec.Vec_IntArray(p.vSeens)
	for _, word := range words[start : start+int(p.nWords)] {
		if word != 0 {
			return 1
		}
	}
	return 0
}

/*
*Function*************************************************************

	Synopsis    [Finds predecessors of the node.]

	Description []

	SideEffects []

	SeeAlso     []

**********************************************************************
*/
func Vta_ObjPreds(p *Vta_Man_t, pThis *Vta_Obj_t, pObj *gia.Gia_Obj_t, ppThis0 **Vta_Obj_t, ppThis1 **Vta_Obj_t) {
	*ppThis0 = nil
	*ppThis1 = nil
	//    if ( !pThis->fAdded )
	//        return;
	assert((gia.Gia_ObjIsPi(p.pGia, pObj) == 0), "(gia.Gia_ObjIsPi(p.pGia, pObj) == 0)")
	if (gia.Gia_ObjIsConst0(pObj) != 0) || ((gia.Gia_ObjIsCi(pObj) != 0) && pThis.iFrame == 0) {
		return
	}
	if gia.Gia_ObjIsAnd(pObj) != 0 {
		*ppThis0 = Vga_ManFind(p, gia.Gia_ObjFaninId0p(p.pGia, pObj), pThis.iFrame)
		*ppThis1 = Vga_ManFind(p, gia.Gia_ObjFaninId1p(p.pGia, pObj), pThis.iFrame)
		//        assert( *ppThis0 && *ppThis1 );
		return
	}
	assert((gia.Gia_ObjIsRo(p.pGia, pObj) != 0) && pThis.iFrame > 0, "(gia.Gia_ObjIsRo(p.pGia, pObj) != 0) && pThis.iFrame > 0")
	pObj = gia.Gia_ObjRoToRi(p.pGia, pObj)
	*ppThis0 = Vga_ManFind(p, gia.Gia_ObjFaninId0p(p.pGia, pObj), pThis.iFrame-1)
	// assert( *ppThis0 );
}

/*
*Function*************************************************************

	Synopsis    [Collect const/PI/RO/AND in a topological order.]

	Description []

	SideEffects []

	SeeAlso     []

**********************************************************************
*/
func Vta_ManCollectNodes_rec(p *Vta_Man_t, pThis *Vta_Obj_t, vOrder *vec.Vec_Int_t) {
	var pObj *gia.Gia_Obj_t
	var pThis0, pThis1 *Vta_Obj_t
	if pThis.fVisit != 0 {
		return
	}
	pThis.fVisit = 1
	pObj = gia.Gia_ManObj(p.pGia, pThis.iObj)
	if pThis.fAdded != 0 {
		Vta_ObjPreds(p, pThis, pObj, &pThis0, &pThis1)
		if pThis0 != nil {
			Vta_ManCollectNodes_rec(p, pThis0, vOrder)
		}
		if pThis1 != nil {
			Vta_ManCollectNodes_rec(p, pThis1, vOrder)
		}
	}
	vec.Vec_IntPush(vOrder, Vta_ObjId(p, pThis))
}
func Vta_ManCollectNodes(p *Vta_Man_t, f int32) *vec.Vec_Int_t {
	var pThis *Vta_Obj_t
	var pObj *gia.Gia_Obj_t
	vec.Vec_IntClear(p.vOrder)
	pObj = gia.Gia_ManPo(p.pGia, 0)
	pThis = Vga_ManFind(p, gia.Gia_ObjFaninId0p(p.pGia, pObj), f)
	assert(pThis != nil, "pThis != nil")
	assert((pThis.fVisit == 0), "(pThis.fVisit == 0)")
	Vta_ManCollectNodes_rec(p, pThis, p.vOrder)
	assert((pThis.fVisit != 0), "(pThis.fVisit != 0)")
	return p.vOrder
}

/*
*Function*************************************************************

	Synopsis    [Refines abstraction.]

	Description []

	SideEffects []

	SeeAlso     []

**********************************************************************
*/
func Vta_ManSatVerify(p *Vta_Man_t) {
	var pThis, pThis0, pThis1 *Vta_Obj_t
	var pObj *gia.Gia_Obj_t
	var i int32
	// inlining of macro Vta_ManForEachObj
	for i = 1; i < p.nObjs; i++ {
		pThis = Vta_ManObj(p, i)
		if vtaSolverValue(p.pSat, i) {
			pThis.Value = VTA_VAR1
		} else {
			pThis.Value = VTA_VAR0
		}
	}
	// inlining of macro Vta_ManForEachObjObj
	for i = 1; i < p.nObjs; i++ {
		pThis = Vta_ManObj(p, i)
		pObj = gia.Gia_ManObj(p.pGia, pThis.iObj)
		if pThis.fAdded == 0 {
			continue
		}
		Vta_ObjPreds(p, pThis, pObj, &pThis0, &pThis1)
		if gia.Gia_ObjIsAnd(pObj) != 0 {
			if pThis.Value == VTA_VAR1 {
				assert((Vta_ValIs1(pThis0, gia.Gia_ObjFaninC0(pObj)) != 0) && (Vta_ValIs1(pThis1, gia.Gia_ObjFaninC1(pObj)) != 0), "Vta_ValIs1(pThis0, gia.Gia_ObjFaninC0(pObj)) && Vta_ValIs1(pThis1, gia.Gia_ObjFaninC1(pObj))")
			} else if pThis.Value == VTA_VAR0 {
				assert((Vta_ValIs0(pThis0, gia.Gia_ObjFaninC0(pObj)) != 0) || (Vta_ValIs0(pThis1, gia.Gia_ObjFaninC1(pObj)) != 0), "Vta_ValIs0(pThis0, gia.Gia_ObjFaninC0(pObj)) || Vta_ValIs0(pThis1, gia.Gia_ObjFaninC1(pObj))")
			} else {
				assert(false, "0")
			}

		} else if gia.Gia_ObjIsRo(p.pGia, pObj) != 0 {
			pObj = gia.Gia_ObjRoToRi(p.pGia, pObj)
			if pThis.iFrame == 0 {
				assert(pThis.Value == VTA_VAR0, "pThis.Value == VTA_VAR0")
			} else if pThis.Value == VTA_VAR0 {
				assert((Vta_ValIs0(pThis0, gia.Gia_ObjFaninC0(pObj)) != 0), "Vta_ValIs0(pThis0, gia.Gia_ObjFaninC0(pObj))")
			} else if pThis.Value == VTA_VAR1 {
				assert((Vta_ValIs1(pThis0, gia.Gia_ObjFaninC0(pObj)) != 0), "Vta_ValIs1(pThis0, gia.Gia_ObjFaninC0(pObj))")
			} else {
				assert(false, "0")
			}
		}
	}
}

/*
*Function*************************************************************

	Synopsis    [Refines abstraction.]

	Description []

	SideEffects []

	SeeAlso     []

**********************************************************************
*/
func Vta_ManProfileAddition(p *Vta_Man_t, vTermsToAdd *vec.Vec_Int_t) {
	var pThis *Vta_Obj_t
	// profile the added ones
	// inlining of macro ABC_CALLOC
	var i int32
	pCounters := make([]int32, p.pPars.iFrame+1)
	// inlining of macro Vta_ManForEachObjObjVec
	for i = 0; i < int32(vec.Vec_IntSize(vTermsToAdd)); i++ {
		pThis = Vta_ManObj(p, vec.Vec_IntEntry(vTermsToAdd, int(i)))
		pCounters[pThis.iFrame]++
	}
	for i = 0; i <= p.pPars.iFrame; i++ {
		util.Abc_Print(1, "%2d", pCounters[i])
	}
	util.Abc_Print(1, "***\n")
}

/*
*Function*************************************************************

	Synopsis    [Refines abstraction.]

	Description []

	SideEffects []

	SeeAlso     []

**********************************************************************
*/
func Vta_ManRefineAbstraction(p *Vta_Man_t, f int32) *util.Abc_Cex_t {
	var fVerify int32
	var pCex *util.Abc_Cex_t = nil
	var vOrder, vTermsToAdd *vec.Vec_Int_t
	var vTermsUsed, vTermsUnused *vec.Vec_Ptr_t
	var pThis, pThis0, pThis1, pTop *Vta_Obj_t
	var pObj *gia.Gia_Obj_t
	var i, Counter int32

	if fVerify != 0 {
		Vta_ManSatVerify(p)
	}

	// collect nodes in a topological order
	vOrder = Vta_ManCollectNodes(p, f)
	// inlining of macro Vta_ManForEachObjObjVec
	for i = 0; i < int32(vec.Vec_IntSize(vOrder)); i++ {
		pThis = Vta_ManObj(p, vec.Vec_IntEntry(vOrder, int(i)))
		pObj = gia.Gia_ManObj(p.pGia, pThis.iObj)
		pThis.Prio = VTA_LARGE
		if vtaSolverValue(p.pSat, Vta_ObjId(p, pThis)) {
			pThis.Value = VTA_VAR1
		} else {
			pThis.Value = VTA_VAR0
		}
		pThis.fVisit = 0
	}

	// verify
	if fVerify != 0 {
		// inlining of macro Vta_ManForEachObjObjVec
		for i = 0; i < int32(vec.Vec_IntSize(vOrder)); i++ {
			pThis = Vta_ManObj(p, vec.Vec_IntEntry(vOrder, int(i)))
			pObj = gia.Gia_ManObj(p.pGia, pThis.iObj)
			if pThis.fAdded == 0 {
				continue
			}
			Vta_ObjPreds(p, pThis, pObj, &pThis0, &pThis1)
			if gia.Gia_ObjIsAnd(pObj) != 0 {
				if pThis.Value == VTA_VAR1 {
					assert((Vta_ValIs1(pThis0, gia.Gia_ObjFaninC0(pObj)) != 0) && (Vta_ValIs1(pThis1, gia.Gia_ObjFaninC1(pObj)) != 0), "Vta_ValIs1(pThis0, gia.Gia_ObjFaninC0(pObj)) && Vta_ValIs1(pThis1, gia.Gia_ObjFaninC1(pObj))")
				} else if pThis.Value == VTA_VAR0 {
					assert((Vta_ValIs0(pThis0, gia.Gia_ObjFaninC0(pObj)) != 0) || (Vta_ValIs0(pThis1, gia.Gia_ObjFaninC1(pObj)) != 0), "Vta_ValIs0(pThis0, gia.Gia_ObjFaninC0(pObj)) || Vta_ValIs0(pThis1, gia.Gia_ObjFaninC1(pObj))")
				} else {
					assert(false, "0")
				}

			} else if gia.Gia_ObjIsRo(p.pGia, pObj) != 0 {
				pObj = gia.Gia_ObjRoToRi(p.pGia, pObj)
				if pThis.iFrame == 0 {
					assert(pThis.Value == VTA_VAR0, "pThis.Value == VTA_VAR0")
				} else if pThis.Value == VTA_VAR0 {
					assert((Vta_ValIs0(pThis0, gia.Gia_ObjFaninC0(pObj)) != 0), "Vta_ValIs0(pThis0, gia.Gia_ObjFaninC0(pObj))")
				} else if pThis.Value == VTA_VAR1 {
					assert((Vta_ValIs1(pThis0, gia.Gia_ObjFaninC0(pObj)) != 0), "Vta_ValIs1(pThis0, gia.Gia_ObjFaninC0(pObj))")
				} else {
					assert(false, "0")
				}
			}
		}
	}

	// compute distance in reverse order
	pThis = Vta_ManObj(p, vec.Vec_IntEntryLast(vOrder))
	pThis.Prio = 1
	// collect used and unused terms
	vTermsUsed = vec.Vec_PtrAlloc(1015)
	vTermsUnused = vec.Vec_PtrAlloc(1016)
	// inlining of macro Vta_ManForEachObjObjVecReverse
	for i = int32(vec.Vec_IntSize(vOrder)) - 1; i >= 0; i-- {
		pThis = Vta_ManObj(p, vec.Vec_IntEntry(vOrder, int(i)))
		pObj = gia.Gia_ManObj(p.pGia, pThis.iObj)
		// there is no unreachable states
		assert(pThis.Prio < VTA_LARGE, "pThis.Prio < VTA_LARGE")
		// skip constants and PIs
		if (gia.Gia_ObjIsConst0(pObj) != 0) || (gia.Gia_ObjIsPi(p.pGia, pObj) != 0) {
			pThis.Prio = 0 // set highest priority
			continue
		}
		// collect terminals
		assert((gia.Gia_ObjIsAnd(pObj) != 0) || (gia.Gia_ObjIsRo(p.pGia, pObj) != 0), "(gia.Gia_ObjIsAnd(pObj) != 0) || (gia.Gia_ObjIsRo(p.pGia, pObj) != 0)")
		if pThis.fAdded == 0 {
			assert(pThis.Prio > 0, "pThis.Prio > 0")
			if Vta_ManObjIsUsed(p, pThis.iObj) != 0 {
				vec.Vec_PtrPush(vTermsUsed, pThis)
			} else {
				vec.Vec_PtrPush(vTermsUnused, pThis)
			}
			continue
		}
		// propagate
		Vta_ObjPreds(p, pThis, pObj, &pThis0, &pThis1)
		if pThis0 != nil {
			pThis0.Prio = min(pThis0.Prio, pThis.Prio+1)
		}
		if pThis1 != nil {
			pThis1.Prio = min(pThis1.Prio, pThis.Prio+1)
		}
	}

	/*
	   Vta_ManForEachObjObjVecReverse( vOrder, p, pThis, pObj, i )
	       if ( pThis->Prio > 0 )
	           pThis->Prio = 10;
	*/
	/*
	   // update priorities according to reconvergence counters
	   Vec_PtrForEachEntry( Vta_Obj_t *, vTermsUsed, pThis, i )
	   {
	       Vta_Obj_t * pThis0, * pThis1;
	       gia.Gia_Obj_t * pObj = gia.Gia_ManObj( p->pGia, pThis->iObj );
	       Vta_ObjPreds( p, pThis, pObj, &pThis0, &pThis1 );
	       pThis->Prio += 10000000;
	       if ( pThis0 )
	           pThis->Prio -= 1000000 * pThis0->fAdded;
	       if ( pThis1 )
	           pThis->Prio -= 1000000 * pThis1->fAdded;
	   }
	   Vec_PtrForEachEntry( Vta_Obj_t *, vTermsUnused, pThis, i )
	   {
	       Vta_Obj_t * pThis0, * pThis1;
	       gia.Gia_Obj_t * pObj = gia.Gia_ManObj( p->pGia, pThis->iObj );
	       Vta_ObjPreds( p, pThis, pObj, &pThis0, &pThis1 );
	       pThis->Prio += 10000000;
	       if ( pThis0 )
	           pThis->Prio -= 1000000 * pThis0->fAdded;
	       if ( pThis1 )
	           pThis->Prio -= 1000000 * pThis1->fAdded;
	   }
	*/

	// update priorities according to reconvergence counters
	// inlining of macro Vec_PtrForEachEntry
	for i = 0; i < int32(vec.Vec_PtrSize(vTermsUsed)); i++ {
		pThis = vec.Vec_PtrEntry(vTermsUsed, int(i)).(*Vta_Obj_t)
		pThis.Prio = uint32(pThis.iObj)
	}
	// inlining of macro Vec_PtrForEachEntry
	for i = 0; i < int32(vec.Vec_PtrSize(vTermsUnused)); i++ {
		pThis = vec.Vec_PtrEntry(vTermsUnused, int(i)).(*Vta_Obj_t)
		pThis.Prio = uint32(pThis.iObj)
	}

	// objects with equal distance should receive priority based on number
	// those objects whose prototypes have been added in other timeframes
	// should have higher priority than the current object
	for _, terms := range []*vec.Vec_Ptr_t{vTermsUsed, vTermsUnused} {
		items := vec.Vec_PtrArray(terms)
		sort.Slice(items, func(i, j int) bool {
			a, b := items[i].(*Vta_Obj_t), items[j].(*Vta_Obj_t)
			return Vta_ManComputeDepthIncrease(&a, &b) < 0
		})
	}
	if vec.Vec_PtrSize(vTermsUsed) > 1 {
		pThis0 = vec.Vec_PtrEntry(vTermsUsed, 0).(*Vta_Obj_t)
		pThis1 = vec.Vec_PtrEntryLast(vTermsUsed).(*Vta_Obj_t)
		assert(pThis0.Prio <= pThis1.Prio, "pThis0.Prio <= pThis1.Prio")
	}
	// assign the priority based on these orders
	Counter = 1
	// inlining of macro Vec_PtrForEachEntry
	for i = 0; i < int32(vec.Vec_PtrSize(vTermsUsed)); i++ {
		pThis = vec.Vec_PtrEntry(vTermsUsed, int(i)).(*Vta_Obj_t)
		pThis.Prio = uint32(Counter)
		Counter++
	}
	// inlining of macro Vec_PtrForEachEntry
	for i = 0; i < int32(vec.Vec_PtrSize(vTermsUnused)); i++ {
		pThis = vec.Vec_PtrEntry(vTermsUnused, int(i)).(*Vta_Obj_t)
		pThis.Prio = uint32(Counter)
		Counter++
	}
	//    util.Abc_Print( 1, "Used %d  Unused %d\n", vec.Vec_PtrSize(vTermsUsed), vec.Vec_PtrSize(vTermsUnused) );

	// propagate in the direct order
	// inlining of macro Vta_ManForEachObjObjVec
	for i = 0; i < int32(vec.Vec_IntSize(vOrder)); i++ {
		pThis = Vta_ManObj(p, vec.Vec_IntEntry(vOrder, int(i)))
		pObj = gia.Gia_ManObj(p.pGia, pThis.iObj)
		assert(pThis.fVisit == 0, "pThis.fVisit == 0")
		assert(pThis.Prio < VTA_LARGE, "pThis.Prio < VTA_LARGE")
		// skip terminal objects
		if pThis.fAdded == 0 {
			continue
		}
		// assumes that values are assigned!!!
		assert(pThis.Value != 0, "pThis.Value != 0")
		// propagate
		if gia.Gia_ObjIsAnd(pObj) != 0 {
			pThis0 = Vga_ManFind(p, gia.Gia_ObjFaninId0p(p.pGia, pObj), pThis.iFrame)
			pThis1 = Vga_ManFind(p, gia.Gia_ObjFaninId1p(p.pGia, pObj), pThis.iFrame)
			assert(pThis0 != nil && pThis1 != nil, "pThis0 && pThis1")
			if pThis.Value == VTA_VAR1 {
				assert((Vta_ValIs1(pThis0, gia.Gia_ObjFaninC0(pObj)) != 0) && (Vta_ValIs1(pThis1, gia.Gia_ObjFaninC1(pObj)) != 0), "Vta_ValIs1(pThis0, gia.Gia_ObjFaninC0(pObj)) && Vta_ValIs1(pThis1, gia.Gia_ObjFaninC1(pObj))")
				pThis.Prio = max(pThis0.Prio, pThis1.Prio)

			} else if pThis.Value == VTA_VAR0 {
				if (Vta_ValIs0(pThis0, gia.Gia_ObjFaninC0(pObj)) != 0) && (Vta_ValIs0(pThis1, gia.Gia_ObjFaninC1(pObj)) != 0) {
					pThis.Prio = min(pThis0.Prio, pThis1.Prio) // choice!!!
				} else if Vta_ValIs0(pThis0, gia.Gia_ObjFaninC0(pObj)) != 0 {
					pThis.Prio = pThis0.Prio
				} else if Vta_ValIs0(pThis1, gia.Gia_ObjFaninC1(pObj)) != 0 {
					pThis.Prio = pThis1.Prio
				} else {
					assert(false, "0")
				}

			} else {
				assert(false, "0")
			}

		} else if gia.Gia_ObjIsRo(p.pGia, pObj) != 0 {
			if pThis.iFrame > 0 {
				pObj = gia.Gia_ObjRoToRi(p.pGia, pObj)
				pThis0 = Vga_ManFind(p, gia.Gia_ObjFaninId0p(p.pGia, pObj), pThis.iFrame-1)
				assert(pThis0 != nil, "pThis0")
				pThis.Prio = pThis0.Prio

			} else {
				pThis.Prio = 0
			}

		} else if gia.Gia_ObjIsConst0(pObj) != 0 {
			pThis.Prio = 0
		} else {
			assert(false, "0")
		}
	}

	// select important values
	pTop = Vta_ManObj(p, vec.Vec_IntEntryLast(vOrder))
	pTop.fVisit = 1
	vTermsToAdd = vec.Vec_IntAlloc(100)
	// inlining of macro Vta_ManForEachObjObjVecReverse
	for i = int32(vec.Vec_IntSize(vOrder)) - 1; i >= 0; i-- {
		pThis = Vta_ManObj(p, vec.Vec_IntEntry(vOrder, int(i)))
		pObj = gia.Gia_ManObj(p.pGia, pThis.iObj)
		if pThis.fVisit == 0 {
			continue
		}
		pThis.fVisit = 0
		assert(pThis.Prio >= 0 && pThis.Prio <= pTop.Prio, "pThis.Prio >= 0 && pThis.Prio <= pTop.Prio")
		// skip terminal objects
		if pThis.fAdded == 0 {
			assert((gia.Gia_ObjIsAnd(pObj) != 0) || (gia.Gia_ObjIsRo(p.pGia, pObj) != 0) || (gia.Gia_ObjIsConst0(pObj) != 0) || (gia.Gia_ObjIsPi(p.pGia, pObj) != 0), "(gia.Gia_ObjIsAnd(pObj) != 0) || (gia.Gia_ObjIsRo(p.pGia, pObj) != 0) || (gia.Gia_ObjIsConst0(pObj) != 0) || (gia.Gia_ObjIsPi(p.pGia, pObj) != 0)")
			vec.Vec_IntPush(vTermsToAdd, Vta_ObjId(p, pThis))
			continue
		}
		// assumes that values are assigned!!!
		assert(pThis.Value != 0, "pThis.Value != 0")
		// propagate
		if gia.Gia_ObjIsAnd(pObj) != 0 {
			pThis0 = Vga_ManFind(p, gia.Gia_ObjFaninId0p(p.pGia, pObj), pThis.iFrame)
			pThis1 = Vga_ManFind(p, gia.Gia_ObjFaninId1p(p.pGia, pObj), pThis.iFrame)
			assert(pThis0 != nil && pThis1 != nil, "pThis0 && pThis1")
			if pThis.Value == VTA_VAR1 {
				assert((Vta_ValIs1(pThis0, gia.Gia_ObjFaninC0(pObj)) != 0) && (Vta_ValIs1(pThis1, gia.Gia_ObjFaninC1(pObj)) != 0), "Vta_ValIs1(pThis0, gia.Gia_ObjFaninC0(pObj)) && Vta_ValIs1(pThis1, gia.Gia_ObjFaninC1(pObj))")
				assert(pThis0.Prio <= pThis.Prio, "pThis0.Prio <= pThis.Prio")
				assert(pThis1.Prio <= pThis.Prio, "pThis1.Prio <= pThis.Prio")
				pThis0.fVisit = 1
				pThis1.fVisit = 1

			} else if pThis.Value == VTA_VAR0 {
				if (Vta_ValIs0(pThis0, gia.Gia_ObjFaninC0(pObj)) != 0) && (Vta_ValIs0(pThis1, gia.Gia_ObjFaninC1(pObj)) != 0) {
					if pThis0.fVisit != 0 {

					} else if pThis1.fVisit != 0 {

					} else if pThis0.Prio <= pThis1.Prio { // choice!!(pThis0.fVisit == 0) = 1
						assert(pThis0.Prio == pThis.Prio, "pThis0.Prio == pThis.Prio")

					} else {
						pThis1.fVisit = 1
						assert(pThis1.Prio == pThis.Prio, "pThis1.Prio == pThis.Prio")
					}

				} else if Vta_ValIs0(pThis0, gia.Gia_ObjFaninC0(pObj)) != 0 {
					pThis0.fVisit = 1
					assert(pThis0.Prio == pThis.Prio, "pThis0.Prio == pThis.Prio")

				} else if Vta_ValIs0(pThis1, gia.Gia_ObjFaninC1(pObj)) != 0 {
					pThis1.fVisit = 1
					assert(pThis1.Prio == pThis.Prio, "pThis1.Prio == pThis.Prio")

				} else {
					assert(false, "0")
				}

			} else {
				assert(false, "0")
			}

		} else if gia.Gia_ObjIsRo(p.pGia, pObj) != 0 {
			if pThis.iFrame > 0 {
				pObj = gia.Gia_ObjRoToRi(p.pGia, pObj)
				pThis0 = Vga_ManFind(p, gia.Gia_ObjFaninId0p(p.pGia, pObj), pThis.iFrame-1)
				assert(pThis0 != nil, "pThis0")
				pThis0.fVisit = 1
				assert(pThis0.Prio == pThis.Prio, "pThis0.Prio == pThis.Prio")
			}

		} else if gia.Gia_ObjIsConst0(pObj) == 0 {
			assert(false, "0")
		}
	}

	if p.pPars.fAddLayer != 0 {
		// mark those currently included
		// inlining of macro Vta_ManForEachObjVec
		for i = 0; i < int32(vec.Vec_IntSize(vTermsToAdd)); i++ {
			pThis = Vta_ManObj(p, vec.Vec_IntEntry(vTermsToAdd, int(i)))
			assert(pThis.fVisit == 0, "pThis.fVisit == 0")
			pThis.fVisit = 1
		}
		// add used terms, which have close relationship
		Counter = int32(vec.Vec_IntSize(vTermsToAdd))
		// inlining of macro Vec_PtrForEachEntry
		for i = 0; i < int32(vec.Vec_PtrSize(vTermsUsed)); i++ {
			pThis = vec.Vec_PtrEntry(vTermsUsed, int(i)).(*Vta_Obj_t)
			if pThis.fVisit != 0 {
				continue
			}
			//        Vta_ObjPreds( p, pThis, gia.Gia_ManObj(p->pGia, pThis->iObj), &pThis0, &pThis1 );
			//        if ( (pThis0 && (pThis0->fAdded || pThis0->fVisit)) || (pThis1 && (pThis1->fAdded || pThis1->fVisit)) )
			vec.Vec_IntPush(vTermsToAdd, Vta_ObjId(p, pThis))
		}
		// remove those currenty included
		// inlining of macro Vta_ManForEachObjVec
		for i = 0; i < int32(vec.Vec_IntSize(vTermsToAdd)); i++ {
			pThis = Vta_ManObj(p, vec.Vec_IntEntry(vTermsToAdd, int(i)))
			pThis.fVisit = 0
		}
	}
	//    printf( "\n%d -> %d\n", Counter, vec.Vec_IntSize(vTermsToAdd) );
	//vec.Vec_IntReverseOrder( vTermsToAdd );
	//vec.Vec_IntSort( vTermsToAdd, 1 );

	// cleanup
	vec.Vec_PtrFree(vTermsUsed)
	vec.Vec_PtrFree(vTermsUnused)

	if fVerify != 0 {
		// verify
		// inlining of macro Vta_ManForEachObjVec
		for i = 0; i < int32(vec.Vec_IntSize(vOrder)); i++ {
			pThis = Vta_ManObj(p, vec.Vec_IntEntry(vOrder, int(i)))
			pThis.Value = VTA_VARX
		}
		// inlining of macro Vta_ManForEachObjVec
		for i = 0; i < int32(vec.Vec_IntSize(vTermsToAdd)); i++ {
			pThis = Vta_ManObj(p, vec.Vec_IntEntry(vTermsToAdd, int(i)))
			assert((pThis.fAdded == 0), "(pThis.fAdded == 0)")
			if vtaSolverValue(p.pSat, Vta_ObjId(p, pThis)) {
				pThis.Value = VTA_VAR1
			} else {
				pThis.Value = VTA_VAR0
			}
		}
		// simulate
		// inlining of macro Vta_ManForEachObjObjVec
		for i = 0; i < int32(vec.Vec_IntSize(vOrder)); i++ {
			pThis = Vta_ManObj(p, vec.Vec_IntEntry(vOrder, int(i)))
			pObj = gia.Gia_ManObj(p.pGia, pThis.iObj)
			assert(pThis.fVisit == 0, "pThis.fVisit == 0")
			if pThis.fAdded == 0 {
				continue
			}
			if gia.Gia_ObjIsAnd(pObj) != 0 {
				pThis0 = Vga_ManFind(p, gia.Gia_ObjFaninId0p(p.pGia, pObj), pThis.iFrame)
				pThis1 = Vga_ManFind(p, gia.Gia_ObjFaninId1p(p.pGia, pObj), pThis.iFrame)
				assert(pThis0 != nil && pThis1 != nil, "pThis0 && pThis1")
				if (Vta_ValIs1(pThis0, gia.Gia_ObjFaninC0(pObj)) != 0) && (Vta_ValIs1(pThis1, gia.Gia_ObjFaninC1(pObj)) != 0) {
					pThis.Value = VTA_VAR1
				} else if (Vta_ValIs0(pThis0, gia.Gia_ObjFaninC0(pObj)) != 0) || (Vta_ValIs0(pThis1, gia.Gia_ObjFaninC1(pObj)) != 0) {
					pThis.Value = VTA_VAR0
				} else {
					pThis.Value = VTA_VARX
				}

			} else if gia.Gia_ObjIsRo(p.pGia, pObj) != 0 {
				if pThis.iFrame > 0 {
					pObj = gia.Gia_ObjRoToRi(p.pGia, pObj)
					pThis0 = Vga_ManFind(p, gia.Gia_ObjFaninId0p(p.pGia, pObj), pThis.iFrame-1)
					assert(pThis0 != nil, "pThis0")
					if Vta_ValIs0(pThis0, gia.Gia_ObjFaninC0(pObj)) != 0 {
						pThis.Value = VTA_VAR0
					} else if Vta_ValIs1(pThis0, gia.Gia_ObjFaninC0(pObj)) != 0 {
						pThis.Value = VTA_VAR1
					} else {
						pThis.Value = VTA_VARX
					}

				} else {
					pThis.Value = VTA_VAR0
				}

			} else if gia.Gia_ObjIsConst0(pObj) != 0 {
				pThis.Value = VTA_VAR0

			} else {
				assert(false, "0")
			}
			// double check the solver
			assert(pThis.Value == VTA_VARX || (pThis.Value == VTA_VAR1) == vtaSolverValue(p.pSat, Vta_ObjId(p, pThis)), "ternary value agrees with SAT assignment")
		}

		// check the output
		if !(Vta_ValIs1(pTop, gia.Gia_ObjFaninC0(gia.Gia_ManPo(p.pGia, 0))) != 0) {
			util.Abc_Print(1, "Vta_ManRefineAbstraction(): Terminary simulation verification failed!\n")
		}
		//    else
		//        util.Abc_Print( 1, "Verification OK.\n" );
	}

	// produce true counter-example
	if pTop.Prio == 0 {
		pCex = Vga_ManDeriveCex(p)
	} else {
		//       Vta_ManProfileAddition( p, vTermsToAdd );

		// inlining of macro Vta_ManForEachObjObjVec
		for i = 0; i < int32(vec.Vec_IntSize(vTermsToAdd)); i++ {
			pThis = Vta_ManObj(p, vec.Vec_IntEntry(vTermsToAdd, int(i)))
			pObj = gia.Gia_ManObj(p.pGia, pThis.iObj)
			if gia.Gia_ObjIsPi(p.pGia, pObj) == 0 {
				Vga_ManAddClausesOne(p, pThis.iObj, pThis.iFrame)
			}
		}
		p.pSat.Simplify()
	}
	p.nObjAdded += int32(vec.Vec_IntSize(vTermsToAdd))
	vec.Vec_IntFree(vTermsToAdd)
	return pCex
}

/*
*Function*************************************************************

	Synopsis    []

	Description []

	SideEffects []

	SeeAlso     []

**********************************************************************
*/
func Vga_ManStart(pGia *gia.Gia_Man_t, pPars *Abs_Par_t) *Vta_Man_t {

	// inlining of macro ABC_CALLOC
	p := &Vta_Man_t{}
	p.pGia = pGia
	p.pPars = pPars
	// internal data
	p.nObjsAlloc = (1 << 18)
	p.objects = newVtaObjectTable(int(p.nObjsAlloc))
	p.nObjs = 1
	p.nBins = int32(len(p.objects.bins))
	p.vOrder = vec.Vec_IntAlloc(1013)
	// abstraction
	p.nObjBits = util.Abc_Base2Log(uint32(gia.Gia_ManObjNum(pGia)))
	p.nObjMask = (1 << p.nObjBits) - 1
	assert(gia.Gia_ManObjNum(pGia) <= int32(p.nObjMask), "gia.Gia_ManObjNum(pGia) <= int32(p.nObjMask)")
	p.nWords = 1
	p.vSeens = vec.Vec_IntStart(int(gia.Gia_ManObjNum(pGia) * p.nWords))
	p.vSeenGla = vec.Vec_BitStart(int(gia.Gia_ManObjNum(pGia)))
	p.nSeenGla = 1
	p.nSeenAll = 1
	// other data
	p.vCores = vec.Vec_PtrAlloc(100)
	p.pSat = bsat.NewLabeledSolver2()
	//    p->pSat->fVerbose = p->pPars->fVerbose;
	//    sat_solver2_set_learntmax( p->pSat, pPars->nLearnedMax );
	p.pSat.SetLearnedPolicy(p.pPars.nLearnedStart, p.pPars.nLearnedDelta, p.pPars.nLearnedPerce)
	// start the abstraction
	assert(gia.Gia_ManObjClasses(pGia) != nil, "gia.Gia_ManObjClasses(pGia) != nil")
	p.vFrames = Gia_VtaAbsToFrames(gia.Gia_ManObjClasses(pGia))
	p.vAddedNew = vec.Vec_IntAlloc(1000)
	return p
}

/*
*Function*************************************************************

	Synopsis    [Delete manager.]

	Description []

	SideEffects []

	SeeAlso     []

**********************************************************************
*/
func Vga_ManStop(p *Vta_Man_t) {
	if p.pPars.fVerbose != 0 {
		util.Abc_Print(1, "SAT solver:  Var = %d  Cla = %d  Conf = %d  Lrn = %d  Reduce = %d  Cex = %d  Objs+ = %d\n",
			p.pSat.NumVars(), int32(p.pSat.Statistics().Clauses), int32(p.pSat.Statistics().Conflicts),
			int32(p.pSat.Statistics().Learned), p.pSat.Reductions(), p.nCexes, p.nObjAdded)
	}
	vec.Vec_VecFreeP(&p.vCores)
	vec.Vec_VecFreeP(&p.vFrames)
	vec.Vec_BitFreeP(&p.vSeenGla)
	vec.Vec_IntFreeP(&p.vSeens)
	vec.Vec_IntFreeP(&p.vOrder)
	vec.Vec_IntFreeP(&p.vAddedNew)
	p.pSat = nil
	p.objects = nil
	p = nil
}

/*
*Function*************************************************************

	Synopsis    [Returns the output literal.]

	Description []

	SideEffects []

	SeeAlso     []

**********************************************************************
*/
func Vga_ManGetOutLit(p *Vta_Man_t, f int32) int32 {
	var pObj *gia.Gia_Obj_t = gia.Gia_ManPo(p.pGia, 0)
	var pThis *Vta_Obj_t = Vga_ManFind(p, gia.Gia_ObjFaninId0p(p.pGia, pObj), f)
	assert(pThis != nil && (pThis.fAdded != 0), "pThis != nil && (pThis.fAdded != 0)")
	if f == 0 && gia.Gia_ObjIsRo(p.pGia, gia.Gia_ObjFanin0(pObj)) != 0 && gia.Gia_ObjFaninC0(pObj) == 0 {
		return -Vta_ObjId(p, pThis)
	}
	return util.Abc_Var2Lit(Vta_ObjId(p, pThis), gia.Gia_ObjFaninC0(pObj))
}

/*
*Function*************************************************************

	Synopsis    [Finds the set of clauses involved in the UNSAT core.]

	Description []

	SideEffects []

	SeeAlso     []

**********************************************************************
*/
func Vta_ManUnsatCore(iLit int32, pSat *bsat.LabeledSolver2, nConfMax int32, fVerbose int32, piRetValue *int32, pnConfls *int32) *vec.Vec_Int_t {
	if piRetValue != nil {
		*piRetValue = 1
	}
	if iLit < 0 {
		core := vec.Vec_IntAlloc(1)
		vec.Vec_IntPush(core, -iLit)
		return core
	}
	before := pSat.Statistics().Conflicts
	pSat.SetLimits(int64(nConfMax), 0)
	status := pSat.Solve(bsat.Literal(iLit))
	if pnConfls != nil {
		*pnConfls = int32(pSat.Statistics().Conflicts - before)
	}
	if status == bsat.Unknown {
		if piRetValue != nil {
			*piRetValue = -1
		}
		return nil
	}
	if status == bsat.Satisfiable {
		if piRetValue != nil {
			*piRetValue = 0
		}
		return nil
	}
	core := pSat.ProofCore()
	return vec.Vec_IntAllocArrayCopy(core, len(core))
}

/*
*Function*************************************************************

	Synopsis    []

	Description []

	SideEffects []

	SeeAlso     []

**********************************************************************
*/
func Vta_ManAbsPrintFrame(p *Vta_Man_t, vCore *vec.Vec_Int_t, nFrames int32, nConfls int32, nCexes int32, Time int64, fVerbose int32) int32 {
	var pInfo []int32
	var pCountAll []int32
	var pCountUni []int32
	var i, iFrame, iObj, Entry, fChanges int32
	// print info about frames
	if vCore != nil {
		// inlining of macro ABC_CALLOC
		pCountAll = make([]int32, nFrames+1)
		// inlining of macro ABC_CALLOC
		pCountUni = make([]int32, nFrames+1)
		// inlining of macro Vec_IntForEachEntry
		for i = 0; i < int32(vec.Vec_IntSize(vCore)); i++ {
			Entry = vec.Vec_IntEntry(vCore, int(i))
			iObj = (int32(uint32(Entry) & p.nObjMask))
			iFrame = (Entry >> p.nObjBits)
			assert(iFrame < nFrames, "iFrame < nFrames")
			pInfo = vec.Vec_IntArray(p.vSeens)[p.nWords*iObj:]
			if uint32(pInfo[iFrame>>5])&(uint32(1)<<uint32(iFrame&31)) == 0 {
				pInfo[iFrame>>5] |= int32(uint32(1) << uint32(iFrame&31))
				pCountUni[iFrame+1]++
				pCountUni[0]++
				p.nSeenAll++
			}
			pCountAll[iFrame+1]++
			pCountAll[0]++
			if vec.Vec_BitEntry(p.vSeenGla, int(iObj)) == 0 {
				vec.Vec_BitWriteEntry(p.vSeenGla, int(iObj), 1)
				p.nSeenGla++
				fChanges = 1
			}
		}
	}
	if !(fVerbose != 0) {
		pCountAll = nil
		pCountUni = nil
		return fChanges
	}

	if abcmain.Abc_FrameIsBatchMode() != 0 && !(vCore != nil) {
		return fChanges
	}

	//    util.Abc_Print( 1, "%5d%5d", pCountAll[0], pCountUni[0] );
	util.Abc_Print(1, "%4d :", nFrames-1)
	util.Abc_Print(1, "%4d", util.Abc_MinInt(100, 100*p.nSeenGla/(gia.Gia_ManRegNum(p.pGia)+gia.Gia_ManAndNum(p.pGia)+1)))
	util.Abc_Print(1, "%6d", p.nSeenGla)
	util.Abc_Print(1, "%4d", util.Abc_MinInt(100, 100*p.nSeenAll/(p.nSeenGla*nFrames)))
	util.Abc_Print(1, "%8d", nConfls)
	if nCexes == 0 {
		util.Abc_Print(1, "%5c", '-')
	} else {
		util.Abc_Print(1, "%5d", nCexes)
	}
	//    util.Abc_Print( 1, " %9d", sat_solver2_nvars(p->pSat) );
	util.Abc_PrintInt(p.pSat.NumVars())
	util.Abc_PrintInt(int32(p.pSat.Statistics().Clauses))
	util.Abc_PrintInt(int32(p.pSat.Statistics().Learned))
	if vCore == nil {
		util.Abc_Print(1, "    ...")
		//        for ( k = 0; k < 7; k++ )
		//            util.Abc_Print( 1, "     " );
		util.Abc_Print(1, "%9.2f sec", float64(Time)/float64(time.Second))
		util.Abc_Print(1, "%5.1f GB", vtaSolverMemory(p.pSat)/(1<<30))
		util.Abc_Print(1, "\r")

	} else {
		util.Abc_PrintInt(pCountAll[0])
		/*
		   if ( nFrames > 7 )
		   {
		       for ( k = 0; k < 3; k++ )
		           util.Abc_Print( 1, "%5d", pCountAll[k+1] );
		       util.Abc_Print( 1, "  ..." );
		       for ( k = nFrames-3; k < nFrames; k++ )
		           util.Abc_Print( 1, "%5d", pCountAll[k+1] );
		   }
		   else
		   {
		       for ( k = 0; k < nFrames; k++ )
		           util.Abc_Print( 1, "%5d", pCountAll[k+1] );
		       for ( k = nFrames; k < 7; k++ )
		           util.Abc_Print( 1, "     " );
		   }
		*/
		util.Abc_Print(1, "%9.2f sec", float64(Time)/float64(time.Second))
		util.Abc_Print(1, "%5.1f GB", vtaSolverMemory(p.pSat)/(1<<30))
		util.Abc_Print(1, "\n")
	}
	os.Stdout.Sync()

	if vCore != nil {
		pCountAll = nil
		pCountUni = nil
	}
	return fChanges
}

/*
*Function*************************************************************

	Synopsis    [Adds clauses to the solver.]

	Description []

	SideEffects []

	SeeAlso     []

**********************************************************************
*/
func Vga_ManAddClausesOne(p *Vta_Man_t, iObj int32, iFrame int32) {
	var pThis0, pThis1 *Vta_Obj_t
	var pObj *gia.Gia_Obj_t = gia.Gia_ManObj(p.pGia, iObj)
	var pThis *Vta_Obj_t = Vga_ManFindOrAdd(p, iObj, iFrame)
	var iThis0 int32
	var iMainVar int32 = Vta_ObjId(p, pThis)
	assert(pThis.iObj == iObj && pThis.iFrame == iFrame, "pThis.iObj == iObj && pThis.iFrame == iFrame")
	if pThis.fAdded != 0 {
		return
	}
	pThis.fAdded = 1
	vec.Vec_IntPush(p.vAddedNew, iMainVar)
	if gia.Gia_ObjIsAnd(pObj) != 0 {
		pThis0 = Vga_ManFindOrAdd(p, gia.Gia_ObjFaninId0p(p.pGia, pObj), iFrame)
		iThis0 = Vta_ObjId(p, pThis0)
		pThis1 = Vga_ManFindOrAdd(p, gia.Gia_ObjFaninId1p(p.pGia, pObj), iFrame)
		vtaSolverAddAnd(p.pSat, iMainVar, iThis0, Vta_ObjId(p, pThis1),
			gia.Gia_ObjFaninC0(pObj), gia.Gia_ObjFaninC1(pObj), 0, iMainVar)

	} else if gia.Gia_ObjIsRo(p.pGia, pObj) != 0 {
		if iFrame == 0 {
			if p.pPars.fUseTermVars != 0 {
				pThis0 = Vga_ManFindOrAdd(p, iObj, -1)
				vtaSolverAddConstraint(p.pSat, iMainVar, Vta_ObjId(p, pThis0), 1, 0, iMainVar)

			} else {
				vtaSolverAddConst(p.pSat, iMainVar, 1, 0, iMainVar)
			}

		} else {
			pObj = gia.Gia_ObjRoToRi(p.pGia, pObj)
			pThis0 = Vga_ManFindOrAdd(p, gia.Gia_ObjFaninId0p(p.pGia, pObj), iFrame-1)
			vtaSolverAddBuffer(p.pSat, iMainVar, Vta_ObjId(p, pThis0), gia.Gia_ObjFaninC0(pObj), 0, iMainVar)
		}

	} else if gia.Gia_ObjIsConst0(pObj) != 0 {
		vtaSolverAddConst(p.pSat, iMainVar, 1, 0, iMainVar)

	} else { //if ( (gia.Gia_ObjIsPi(p->pGia, pObj) == 0) )
		assert(false, "0")
	}
}

/*
*Function*************************************************************

	Synopsis    []

	Description []

	SideEffects []

	SeeAlso     []

**********************************************************************
*/
func Vga_ManLoadSlice(p *Vta_Man_t, vOne *vec.Vec_Int_t, Lift int32) {
	var i, Entry int32
	// inlining of macro Vec_IntForEachEntry
	for i = 0; i < int32(vec.Vec_IntSize(vOne)); i++ {
		Entry = vec.Vec_IntEntry(vOne, int(i))
		Vga_ManAddClausesOne(p, int32(uint32(Entry)&p.nObjMask), (Entry>>p.nObjBits)+Lift)
	}
	p.pSat.Simplify()
}

/*
*Function*************************************************************

	Synopsis    []

	Description []

	SideEffects []

	SeeAlso     []

**********************************************************************
*/
func Vga_ManPrintCore(p *Vta_Man_t, vCore *vec.Vec_Int_t, Lift int32) {
	var i, Entry, iObj, iFrame int32
	// inlining of macro Vec_IntForEachEntry
	for i = 0; i < int32(vec.Vec_IntSize(vCore)); i++ {
		Entry = vec.Vec_IntEntry(vCore, int(i))
		iObj = (int32(uint32(Entry) & p.nObjMask))
		iFrame = (Entry >> p.nObjBits)
		util.Abc_Print(1, "%d*%d ", iObj, iFrame+Lift)
	}
	util.Abc_Print(1, "\n")
}

/*
*Function*************************************************************

	Synopsis    []

	Description []

	SideEffects []

	SeeAlso     []

**********************************************************************
*/
func Vga_ManRollBack(p *Vta_Man_t, nObjOld int32) {
	var pThis *Vta_Obj_t
	var i, Entry int32
	p.objects.rollback(nObjOld)
	p.nObjs = nObjOld
	// inlining of macro Vec_IntForEachEntry
	for i = 0; i < int32(vec.Vec_IntSize(p.vAddedNew)); i++ {
		Entry = vec.Vec_IntEntry(p.vAddedNew, int(i))
		if Entry < p.nObjs {
			pThis = Vta_ManObj(p, Entry)
			assert(pThis.fAdded == 1, "pThis.fAdded == 1")
			pThis.fAdded = 0
		}
	}
}

/*
*Function*************************************************************

	Synopsis    [Send abstracted model or send cancel.]

	Description [Counter-example will be sent automatically when &vta terminates.]

	SideEffects []

	SeeAlso     []

**********************************************************************
*/
func Gia_VtaSendAbsracted(p *Vta_Man_t, fVerbose int32) {
	var pAbs *gia.Gia_Man_t
	assert(abcmain.Abc_FrameIsBridgeMode() != 0, "abcmain.Abc_FrameIsBridgeMode()")
	//    if ( fVerbose )
	//        util.Abc_Print( 1, "Sending abstracted model...\n" );
	// create obj classes
	gia.Gia_ManSetObjClasses(p.pGia, nil)
	gia.Gia_ManSetObjClasses(p.pGia, Gia_VtaFramesToAbs(p.vCores))
	// create gate classes
	gia.Gia_ManSetGateClasses(p.pGia, nil)
	gia.Gia_ManSetGateClasses(p.pGia, Gia_VtaConvertToGla(p.pGia, gia.Gia_ManObjClasses(p.pGia)))
	gia.Gia_ManSetObjClasses(p.pGia, nil)
	// create abstrated model
	pAbs = Gia_ManDupAbsGates(p.pGia, gia.Gia_ManGateClasses(p.pGia))
	gia.Gia_ManSetGateClasses(p.pGia, nil)
	// send it out
	net, err := gia.Gia_ManToBridgeNetlist(pAbs)
	if err != nil {
		panic(err)
	}
	util.Gia_ManToBridgeAbsNetlist(os.Stdout, net, 107)
	gia.Gia_ManStop(pAbs)
}
func Gia_VtaSendCancel(p *Vta_Man_t, fVerbose int32) {
	// C forward declaration: Gia_ManToBridgeBadAbs(FILE *pFile).
	assert(abcmain.Abc_FrameIsBridgeMode() != 0, "abcmain.Abc_FrameIsBridgeMode()")
	//    if ( fVerbose )
	//        util.Abc_Print( 1, "Cancelling previously sent model...\n" );
	util.Gia_ManToBridgeBadAbs(os.Stdout)
}

/*
*Function*************************************************************

	Synopsis    [Send abstracted model or send cancel.]

	Description [Counter-example will be sent automatically when &vta terminates.]

	SideEffects []

	SeeAlso     []

**********************************************************************
*/
func Gia_VtaDumpAbsracted(p *Vta_Man_t, fVerbose int32) {
	var pFileNameDef string = "vabs.aig"
	var pFileName string
	if p.pPars.pFileVabs != "" {
		pFileName = p.pPars.pFileVabs
	} else {
		pFileName = pFileNameDef
	}
	var pAbs *gia.Gia_Man_t
	if fVerbose != 0 {
		util.Abc_Print(1, "Dumping abstracted model into file \"%s\"...\n", pFileName)
	}
	// create obj classes
	gia.Gia_ManSetObjClasses(p.pGia, nil)
	gia.Gia_ManSetObjClasses(p.pGia, Gia_VtaFramesToAbs(p.vCores))
	// create gate classes
	gia.Gia_ManSetGateClasses(p.pGia, nil)
	gia.Gia_ManSetGateClasses(p.pGia, Gia_VtaConvertToGla(p.pGia, gia.Gia_ManObjClasses(p.pGia)))
	gia.Gia_ManSetObjClasses(p.pGia, nil)
	// create abstrated model
	pAbs = Gia_ManDupAbsGates(p.pGia, gia.Gia_ManGateClasses(p.pGia))
	gia.Gia_ManSetGateClasses(p.pGia, nil)
	// send it out
	gia.Gia_AigerWrite(pAbs, pFileName, 0, 0, 0)
	gia.Gia_ManStop(pAbs)
}

/*
*Function*************************************************************

	Synopsis    [Print memory report.]

	Description []

	SideEffects []

	SeeAlso     []

**********************************************************************
*/
func Gia_VtaPrintMemory(p *Vta_Man_t) {
	var memTot float64
	var memAig = float64(gia.Gia_ManObjNum(p.pGia)) * float64(unsafe.Sizeof(gia.Gia_Obj_t{}))
	memSat, memPro := p.pSat.Memory(true)
	var memMap = float64(p.nObjsAlloc)*float64(unsafe.Sizeof(Vta_Obj_t{})) + float64(p.nBins)*float64(unsafe.Sizeof(int32(0)))
	var memOth = float64(unsafe.Sizeof(Vta_Man_t{}))
	memOth += float64(vec.Vec_IntCap(p.vOrder) * 4)
	memOth += vec.Vec_VecMemoryInt(p.vFrames)
	memOth += float64(vec.Vec_BitCap(p.vSeenGla) / 8)
	memOth += vec.Vec_VecMemoryInt(p.vCores)
	memOth += float64(vec.Vec_IntCap(p.vAddedNew) * 4)
	memTot = memAig + memSat + memPro + memMap + memOth
	// inlining of macro ABC_PRMP
	util.Abc_Print(1, "%s =", ("Memory: AIG     "))
	util.Abc_Print(1, "%10.3f MB (%6.2f %%)\n", (1.0 * (float64((memAig))) / (1 << 20)), vtaPercent(memAig, memTot))
	// inlining of macro ABC_PRMP
	util.Abc_Print(1, "%s =", ("Memory: SAT     "))
	util.Abc_Print(1, "%10.3f MB (%6.2f %%)\n", (1.0 * (float64((memSat))) / (1 << 20)), vtaPercent(memSat, memTot))
	// inlining of macro ABC_PRMP
	util.Abc_Print(1, "%s =", ("Memory: Proof   "))
	util.Abc_Print(1, "%10.3f MB (%6.2f %%)\n", (1.0 * (float64((memPro))) / (1 << 20)), vtaPercent(memPro, memTot))
	// inlining of macro ABC_PRMP
	util.Abc_Print(1, "%s =", ("Memory: Map     "))
	util.Abc_Print(1, "%10.3f MB (%6.2f %%)\n", (1.0 * (float64((memMap))) / (1 << 20)), vtaPercent(memMap, memTot))
	// inlining of macro ABC_PRMP
	util.Abc_Print(1, "%s =", ("Memory: Other   "))
	util.Abc_Print(1, "%10.3f MB (%6.2f %%)\n", (1.0 * (float64((memOth))) / (1 << 20)), vtaPercent(memOth, memTot))
	// inlining of macro ABC_PRMP
	util.Abc_Print(1, "%s =", ("Memory: TOTAL   "))
	util.Abc_Print(1, "%10.3f MB (%6.2f %%)\n", (1.0 * (float64((memTot))) / (1 << 20)), vtaPercent(memTot, memTot))
}

/*
*Function*************************************************************

	Synopsis    [Collect nodes/flops involved in different timeframes.]

	Description []

	SideEffects []

	SeeAlso     []

**********************************************************************
*/
func Gia_VtaPerformInt(pAig *gia.Gia_Man_t, pPars *Abs_Par_t) int32 {
	var p *Vta_Man_t
	var vCore *vec.Vec_Int_t
	var pCex *util.Abc_Cex_t = nil
	var i int32
	var f int32
	var nConfls int32
	var Status int32
	var nObjOld int32
	var RetValue int32 = -1
	var nCountNoChange int32
	var fOneIsSent int32
	var clk int64 = util.Abc_Clock()
	var clk2 int64
	// preconditions
	assert(gia.Gia_ManPoNum(pAig) == 1, "gia.Gia_ManPoNum(pAig) == 1")
	assert(pPars.nFramesMax == 0 || pPars.nFramesStart <= pPars.nFramesMax, "pPars.nFramesMax == 0 || pPars.nFramesStart <= pPars.nFramesMax")
	if gia.Gia_ObjIsConst0(gia.Gia_ObjFanin0(gia.Gia_ManPo(pAig, 0))) != 0 {
		if gia.Gia_ObjFaninC0(gia.Gia_ManPo(pAig, 0)) == 0 {
			fmt.Printf("Sequential miter is trivially UNSAT.\n")
			return 1
		}
		if err := pAig.SetSequentialCounterexample(util.Abc_CexMakeTriv(gia.Gia_ManRegNum(pAig), gia.Gia_ManPiNum(pAig), 1, 0)); err != nil {
			panic(err)
		}
		fmt.Printf("Sequential miter is trivially SAT.\n")
		return 0
	}

	// compute intial abstraction
	if gia.Gia_ManObjClasses(pAig) == nil {
		gia.Gia_ManSetObjClasses(pAig, vec.Vec_IntAlloc(5))
		vec.Vec_IntPush(gia.Gia_ManObjClasses(pAig), 1)
		vec.Vec_IntPush(gia.Gia_ManObjClasses(pAig), 3)
		vec.Vec_IntPush(gia.Gia_ManObjClasses(pAig), 4)
		vec.Vec_IntPush(gia.Gia_ManObjClasses(pAig), gia.Gia_ObjFaninId0p(pAig, gia.Gia_ManPo(pAig, 0)))
	}
	// start the manager
	p = Vga_ManStart(pAig, pPars)
	// set runtime limit
	if p.pPars.nTimeOut != 0 {
		p.deadline = time.Now().Add(time.Duration(p.pPars.nTimeOut) * time.Second)
		p.pSat.SetDeadline(p.deadline)
	}
	// perform initial abstraction
	if p.pPars.fVerbose != 0 {
		util.Abc_Print(1, "Running variable-timeframe abstraction (VTA) with the following parameters:\n")
		util.Abc_Print(1, "FramePast = %d  FrameMax = %d  ConfMax = %d  Timeout = %d  RatioMin = %d %%\n",
			pPars.nFramesPast, pPars.nFramesMax, pPars.nConfLimit, pPars.nTimeOut, pPars.nRatioMin)
		util.Abc_Print(1, "LearnStart = %d  LearnDelta = %d  LearnRatio = %d %%.\n",
			pPars.nLearnedStart, pPars.nLearnedDelta, pPars.nLearnedPerce)
		//        util.Abc_Print( 1, "Frame   %%   Abs   %%   Confl  Cex    SatVar   Core   F0   F1   F2  ...\n" );
		util.Abc_Print(1, " Frame   %%   Abs   %%   Confl  Cex   Vars   Clas   Lrns   Core     Time      Mem\n")
	}
	assert(vec.Vec_PtrSize(p.vFrames) > 0, "vec.Vec_PtrSize(p.vFrames) > 0")
	i = 0
	for f = 0; p.pPars.nFramesMax == 0 || f < p.pPars.nFramesMax; f++ {
		var nConflsBeg int32 = int32(p.pSat.Statistics().Conflicts)
		p.pPars.iFrame = f
		// realloc storage for abstraction marks
		if f == p.nWords*32 {
			p.nWords = Vec_IntDoubleWidth(p.vSeens, p.nWords)
		}

		// create bookmark to be used for rollback
		nObjOld = p.nObjs
		p.pSat.Bookmark()
		vec.Vec_IntClear(p.vAddedNew)

		// load new timeframe
		Vga_ManAddClausesOne(p, 0, f)
		if f < int32(vec.Vec_PtrSize(p.vFrames)) {
			Vga_ManLoadSlice(p, vec.Vec_PtrEntry(p.vFrames, int(f)).(*vec.Vec_Int_t), 0)
		} else {
			for i = 1; i <= util.Abc_MinInt(p.pPars.nFramesPast, f); i++ {
				Vga_ManLoadSlice(p, vec.Vec_PtrEntry(p.vCores, int(f-i)).(*vec.Vec_Int_t), i)
			}
		}

		// iterate as long as there are counter-examples
		for i = 0; ; i++ {
			clk2 = util.Abc_Clock()
			vCore = Vta_ManUnsatCore(Vga_ManGetOutLit(p, f), p.pSat, pPars.nConfLimit, pPars.fVerbose, &Status, &nConfls)
			assert((vCore != nil) == (Status == 1), "(vCore != nil) == (Status == 1)")
			if Status == -1 { // resource limit is reached

				Vga_ManRollBack(p, nObjOld)
				goto finish
			}
			// check timeout
			if !p.deadline.IsZero() && time.Now().After(p.deadline) {
				Vga_ManRollBack(p, nObjOld)
				goto finish
			}
			if vCore != nil {
				p.timeUnsat += util.Abc_Clock() - clk2
				break
			}
			p.timeSat += util.Abc_Clock() - clk2
			assert(Status == 0, "Status == 0")
			p.nCexes++
			// perform the refinement
			clk2 = util.Abc_Clock()
			pCex = Vta_ManRefineAbstraction(p, f)
			p.timeCex += util.Abc_Clock() - clk2
			if pCex != nil {
				goto finish
			}
			// print the result (do not count it towards change)
			Vta_ManAbsPrintFrame(p, nil, f+1, int32(p.pSat.Statistics().Conflicts)-nConflsBeg, i, util.Abc_Clock()-clk, p.pPars.fVerbose)
		}
		assert(Status == 1, "Status == 1")
		// valid core is obtained
		Vta_ManUnsatCoreRemap(p, vCore)
		vec.Vec_IntSort(vCore, 1)
		// update the SAT solver
		p.pSat.Rollback()
		// update storage
		Vga_ManRollBack(p, nObjOld)
		// load this timeframe
		Vga_ManLoadSlice(p, vCore, 0)
		vec.Vec_IntFree(vCore)

		// run SAT solver
		clk2 = util.Abc_Clock()
		vCore = Vta_ManUnsatCore(Vga_ManGetOutLit(p, f), p.pSat, pPars.nConfLimit, p.pPars.fVerbose, &Status, &nConfls)
		p.timeUnsat += util.Abc_Clock() - clk2
		assert((vCore != nil) == (Status == 1), "(vCore != nil) == (Status == 1)")
		if Status == -1 { // resource limit is reached
			break
		}
		if Status == 0 {
			Vta_ManSatVerify(p)
			// make sure, there was no initial abstraction (otherwise, it was invalid)
			assert(gia.Gia_ManObjClasses(pAig) == nil && f < p.pPars.nFramesStart, "gia.Gia_ManObjClasses(pAig) == nil && f < p.pPars.nFramesStart")
			pCex = Vga_ManDeriveCex(p)
			break
		}
		// add the core
		Vta_ManUnsatCoreRemap(p, vCore)
		// add in direct topological order
		vec.Vec_IntSort(vCore, 1)
		vec.Vec_PtrPush(p.vCores, vCore)
		// print the result
		if Vta_ManAbsPrintFrame(p, vCore, f+1, int32(p.pSat.Statistics().Conflicts)-nConflsBeg, i, util.Abc_Clock()-clk, p.pPars.fVerbose) != 0 {
			// reset the counter of frames without change
			nCountNoChange = 1
			p.pPars.nFramesNoChange = 0

		} else if nCountNoChange++; nCountNoChange == 2 { // time to send

			p.pPars.nFramesNoChange++
			if abcmain.Abc_FrameIsBridgeMode() != 0 {
				// cancel old one if it was sent
				if fOneIsSent != 0 {
					Gia_VtaSendCancel(p, pPars.fVerbose)
				}
				// send new one
				Gia_VtaSendAbsracted(p, pPars.fVerbose)
				fOneIsSent = 1
			}
		}
		// dump the model
		if (p.pPars.fDumpVabs != 0) && (f&1) != 0 {
			var Command string // was [1000]byte in C

			abcmain.Abc_FrameSetStatus(-1)
			abcmain.Abc_FrameSetCex(nil)
			abcmain.Abc_FrameSetNFrames(f + 1)
			fileName := p.pPars.pFileVabs
			if fileName == "" {
				fileName = "vtabs.aig"
			}
			Command = fmt.Sprintf("write_status %s", extra.Extra_FileNameGenericAppend(fileName, ".status"))
			abcmain.Cmd_CommandExecute(abcmain.Abc_FrameGetGlobalFrame(), Command)
			Gia_VtaDumpAbsracted(p, pPars.fVerbose)
		}
		// check if the number of objects is below limit
		if p.nSeenGla >= gia.Gia_ManCandNum(pAig)*(100-pPars.nRatioMin)/100 {
			Status = -1
			break
		}
	}
finish:
	// analize the results
	if pCex == nil {
		if (p.pPars.fVerbose != 0) && Status == -1 {
			fmt.Printf("\n")
		}
		if vec.Vec_PtrSize(p.vCores) == 0 {
			util.Abc_Print(1, "Abstraction is not produced because first frame is not solved.  ")
		} else {
			assert(vec.Vec_PtrSize(p.vCores) > 0, "vec.Vec_PtrSize(p.vCores) > 0")
			//            if ( pAig->vObjClasses != NULL )
			//                util.Abc_Print( 1, "Replacing the old abstraction by a new one.\n" );
			gia.Gia_ManSetObjClasses(pAig, nil)
			gia.Gia_ManSetObjClasses(pAig, Gia_VtaFramesToAbs(p.vCores))
			if Status == -1 {
				if !p.deadline.IsZero() && !time.Now().Before(p.deadline) {
					util.Abc_Print(1, "Timeout %d sec in frame %d with a %d-stable abstraction.    ", p.pPars.nTimeOut, f, p.pPars.nFramesNoChange)
				} else if pPars.nConfLimit != 0 && int32(p.pSat.Statistics().Conflicts) >= pPars.nConfLimit {
					util.Abc_Print(1, "Exceeded %d conflicts in frame %d with a %d-stable abstraction.  ", pPars.nConfLimit, f, p.pPars.nFramesNoChange)
				} else if p.nSeenGla >= gia.Gia_ManCandNum(pAig)*(100-pPars.nRatioMin)/100 {
					util.Abc_Print(1, "The ratio of abstracted objects is less than %d %% in frame %d.  ", pPars.nRatioMin, f)
				} else {
					util.Abc_Print(1, "Abstraction stopped for unknown reason in frame %d.  ", f)
				}

			} else {
				p.pPars.iFrame++
				util.Abc_Print(1, "VTA completed %d frames with a %d-stable abstraction.  ", f, p.pPars.nFramesNoChange)
			}
		}

	} else {
		if p.pPars.fVerbose != 0 {
			fmt.Printf("\n")
		}
		// ABC's C implementation reports SAT even when this replay fails. In
		// particular, optional terminal variables can yield a spurious initial
		// state. Never publish that witness as a property violation.
		if gia.Gia_ManVerifyCex(p.pGia, pCex, 0) == 0 {
			p.pGia.ClearSequentialCounterexample()
			util.Abc_Print(1, "    Gia_VtaPerform(): CEX verification has failed!\n")
		} else {
			if err := p.pGia.SetSequentialCounterexample(pCex); err != nil {
				panic(err)
			}
			util.Abc_Print(1, "Counter-example detected in frame %d.  ", f)
			p.pPars.iFrame = pCex.IFrame - 1
			gia.Gia_ManSetObjClasses(pAig, nil)
			RetValue = 0
		}
	}
	util.Abc_PrintTime(1, "Time", util.Abc_Clock()-clk)

	if p.pPars.fVerbose != 0 {
		p.timeOther = (util.Abc_Clock() - clk) - p.timeUnsat - p.timeSat - p.timeCex
		// inlining of macro ABC_PRTP
		util.Abc_Print(1, "%s =", ("Runtime: Solver UNSAT"))
		util.Abc_Print(1, "%9.2f sec (%6.2f %%)\n", 1.0*(float64((p.timeUnsat)))/(float64(int64(time.Second))), vtaPercent(float64(p.timeUnsat), float64(util.Abc_Clock()-clk)))
		// inlining of macro ABC_PRTP
		util.Abc_Print(1, "%s =", ("Runtime: Solver SAT  "))
		util.Abc_Print(1, "%9.2f sec (%6.2f %%)\n", 1.0*(float64((p.timeSat)))/(float64(int64(time.Second))), vtaPercent(float64(p.timeSat), float64(util.Abc_Clock()-clk)))
		// inlining of macro ABC_PRTP
		util.Abc_Print(1, "%s =", ("Runtime: Refinement  "))
		util.Abc_Print(1, "%9.2f sec (%6.2f %%)\n", 1.0*(float64((p.timeCex)))/(float64(int64(time.Second))), vtaPercent(float64(p.timeCex), float64(util.Abc_Clock()-clk)))
		// inlining of macro ABC_PRTP
		util.Abc_Print(1, "%s =", ("Runtime: Other       "))
		util.Abc_Print(1, "%9.2f sec (%6.2f %%)\n", 1.0*(float64((p.timeOther)))/(float64(int64(time.Second))), vtaPercent(float64(p.timeOther), float64(util.Abc_Clock()-clk)))
		// inlining of macro ABC_PRTP
		util.Abc_Print(1, "%s =", ("Runtime: TOTAL       "))
		util.Abc_Print(1, "%9.2f sec (%6.2f %%)\n", 1.0*(float64((util.Abc_Clock() - clk)))/(float64(int64(time.Second))), vtaPercent(float64(util.Abc_Clock()-clk), float64(util.Abc_Clock()-clk)))
		Gia_VtaPrintMemory(p)
	}

	Vga_ManStop(p)
	os.Stdout.Sync()
	return RetValue
}

/*
*Function*************************************************************

	Synopsis    [Collect nodes/flops involved in different timeframes.]

	Description []

	SideEffects []

	SeeAlso     []

**********************************************************************
*/
func Gia_VtaPerform(pAig *gia.Gia_Man_t, pPars *Abs_Par_t) int32 {
	var RetValue int32 = -1
	if gia.Gia_ManObjClasses(pAig) == nil && (pPars.fUseRollback != 0) {
		var nFramesMaxOld int32 = pPars.nFramesMax
		pPars.nFramesMax = pPars.nFramesStart
		RetValue = Gia_VtaPerformInt(pAig, pPars)
		pPars.nFramesMax = nFramesMaxOld
	}
	if RetValue == 0 {
		return RetValue
	}
	return Gia_VtaPerformInt(pAig, pPars)
}

////////////////////////////////////////////////////////////////////////
///                       END OF FILE                                ///
////////////////////////////////////////////////////////////////////////

//ABC_NAMESPACE_IMPL_END

// vtaPercent matches ABC_PRMP and ABC_PRTP for zero totals.
func vtaPercent(value, total float64) float64 {
	if total == 0 {
		return 0
	}
	return 100 * value / total
}
