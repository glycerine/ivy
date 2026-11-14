package mpm

import (
	"github.com/glycerine/abcgo/misc/util"
	. "github.com/glycerine/abcgo/misc/vec"
)

/**CFile****************************************************************

  FileName    [mpmMig.h]

  SystemName  [ABC: Logic synthesis and verification system.]

  PackageName [Configurable technology mapper.]

  Synopsis    [Internal declarations.]

  Author      [Alan Mishchenko]

  Affiliation [UC Berkeley]

  Date        [Ver. 1.0. Started - June 1, 2013.]

  Revision    [$Id: mpmMig.h,v 1.00 2013/06/01 00:00:00 alanmi Exp $]

***********************************************************************/

////////////////////////////////////////////////////////////////////////
///                          INCLUDES                                ///
////////////////////////////////////////////////////////////////////////

//#include "misc/vec/vec.h"

//ABC_NAMESPACE_HEADER_START

////////////////////////////////////////////////////////////////////////
///                         PARAMETERS                               ///
////////////////////////////////////////////////////////////////////////

const MIG_NONE = 0x7FFFFFFF

// #define MIG_MASK 0x0000FFFF
// #define MIG_BASE 16
const MIG_MASK = 0x0000FFF
const MIG_BASE = 12

////////////////////////////////////////////////////////////////////////
///                         BASIC TYPES                              ///
////////////////////////////////////////////////////////////////////////

type Mig_Fan_t struct {
	fCompl uint32 /* C bitfield: 1 bits; review packing and truncation */  // the complemented attribute
	Id     uint32 /* C bitfield: 31 bits; review packing and truncation */ // fanin ID
}

type Mig_Obj_t struct {
	pFans [4]Mig_Fan_t // fanins
	owner *Mig_Man_t   // replaces the C page-prefix manager pointer
}

type Mig_Man_t struct {
	pName    string    // name
	nObjs    int32     // number of objects
	nRegs    int32     // number of flops
	nChoices int32     // number of choices
	vPages   Vec_Ptr_t // memory pages
	vCis     Vec_Int_t // CI IDs
	vCos     Vec_Int_t // CO IDs
	// object iterator
	pPage *Mig_Obj_t // current page
	iPage int32      // current page index
	// attributes
	nTravIds int32     // traversal ID counter
	vTravIds Vec_Int_t // traversal IDs
	vLevels  Vec_Int_t // levels
	vSibls   Vec_Int_t // choice nodes
	vRefs    Vec_Int_t // ref counters
	vCopies  Vec_Int_t // copies
	pMan     any       // mapping manager
}

/*
   Usage of fanin atrributes
   --------------------------------------------------------------------------------------------------------------
      Const0  Terminal    CI      CO     Buf     Node    Node2   Node3   And2    XOR2    MUX     MAJ    Sentinel
   --------------------------------------------------------------------------------------------------------------
   0    -     -/fanin0    -     fanin0  fanin0  fanin0  fanin0  fanin0  fanin0  fanin1  fanin0  fanin1     -
   1    -        -        -       -       -     fanin1  fanin1  fanin1  fanin1  fanin0  fanin1  fanin0     -
   2    -      CIO ID   CIO ID  CIO ID    -    -/fanin2    -    fanin2    -       -     fanin2  fanin2     -
   3    0        ID       ID      ID      ID      ID      ID      ID      ID      ID      ID      ID       -
   --------------------------------------------------------------------------------------------------------------

   One memory page contain 2^MIG_BASE+2 16-byte objects.
   - the first object contains the pointer to the manager (8 bytes)
   - the next 2^MIG_BASE are potentially used as objects
   - the last object is a sentinel to signal the end of the page
*/

func Mig_IdPage(v int32) int32 { return v >> MIG_BASE }
func Mig_IdCell(v int32) int32 { return v & MIG_MASK }

func Mig_ManName(p *Mig_Man_t) string  { return p.pName }
func Mig_ManCiNum(p *Mig_Man_t) int32  { return int32(Vec_IntSize(&p.vCis)) }
func Mig_ManCoNum(p *Mig_Man_t) int32  { return int32(Vec_IntSize(&p.vCos)) }
func Mig_ManPiNum(p *Mig_Man_t) int32  { return int32(Vec_IntSize(&p.vCis)) - p.nRegs }
func Mig_ManPoNum(p *Mig_Man_t) int32  { return int32(Vec_IntSize(&p.vCos)) - p.nRegs }
func Mig_ManRegNum(p *Mig_Man_t) int32 { return p.nRegs }
func Mig_ManObjNum(p *Mig_Man_t) int32 { return p.nObjs }
func Mig_ManNodeNum(p *Mig_Man_t) int32 {
	return p.nObjs - int32(Vec_IntSize(&p.vCis)) - int32(Vec_IntSize(&p.vCos)) - 1
}
func Mig_ManCandNum(p *Mig_Man_t) int32      { return Mig_ManCiNum(p) + Mig_ManNodeNum(p) }
func Mig_ManChoiceNum(p *Mig_Man_t) int32    { return p.nChoices }
func Mig_ManSetRegNum(p *Mig_Man_t, v int32) { p.nRegs = v }

func Mig_ManPage(p *Mig_Man_t, v int32) []Mig_Obj_t {
	return Vec_PtrEntry(&p.vPages, int(Mig_IdPage(v))).([]Mig_Obj_t)
}
func Mig_ManObj(p *Mig_Man_t, v int32) *Mig_Obj_t {
	assert(v >= 0 && v < p.nObjs, "v >= 0 && v < p.nObjs")
	return &Mig_ManPage(p, v)[Mig_IdCell(v)]
}
func Mig_ManCi(p *Mig_Man_t, v int32) *Mig_Obj_t { return Mig_ManObj(p, Vec_IntEntry(&p.vCis, int(v))) }
func Mig_ManCo(p *Mig_Man_t, v int32) *Mig_Obj_t { return Mig_ManObj(p, Vec_IntEntry(&p.vCos, int(v))) }
func Mig_ManPi(p *Mig_Man_t, v int32) *Mig_Obj_t {
	assert(v < Mig_ManPiNum(p), "v < Mig_ManPiNum(p)")
	return Mig_ManCi(p, v)
}
func Mig_ManPo(p *Mig_Man_t, v int32) *Mig_Obj_t {
	assert(v < Mig_ManPoNum(p), "v < Mig_ManPoNum(p)")
	return Mig_ManCo(p, v)
}
func Mig_ManRo(p *Mig_Man_t, v int32) *Mig_Obj_t {
	assert(v < Mig_ManRegNum(p), "v < Mig_ManRegNum(p)")
	return Mig_ManCi(p, Mig_ManPiNum(p)+v)
}
func Mig_ManRi(p *Mig_Man_t, v int32) *Mig_Obj_t {
	assert(v < Mig_ManRegNum(p), "v < Mig_ManRegNum(p)")
	return Mig_ManCo(p, Mig_ManPoNum(p)+v)
}
func Mig_ManConst0(p *Mig_Man_t) *Mig_Obj_t { return Mig_ManObj(p, 0) }

func migBool(v bool) int32 {
	if v {
		return 1
	}
	return 0
}
func Mig_FanCompl(p *Mig_Obj_t, i int32) int32  { return int32(p.pFans[i].fCompl) }
func Mig_FanId(p *Mig_Obj_t, i int32) int32     { return int32(p.pFans[i].Id) }
func Mig_FanIsNone(p *Mig_Obj_t, i int32) int32 { return migBool(p.pFans[i].Id == MIG_NONE) }
func Mig_FanSetCompl(p *Mig_Obj_t, i, v int32) int32 {
	assert(v >= 0 && v <= 1, "one-bit complement")
	p.pFans[i].fCompl = uint32(v)
	return v
}
func Mig_FanSetId(p *Mig_Obj_t, i, v int32) int32 {
	assert(v >= 0 && v < MIG_NONE, "valid fanin id")
	p.pFans[i].Id = uint32(v)
	return v
}
func Mig_ObjIsNone(p *Mig_Obj_t) int32   { return Mig_FanIsNone(p, 3) }
func Mig_ObjIsConst0(p *Mig_Obj_t) int32 { return migBool(Mig_FanId(p, 3) == 0) }
func Mig_ObjIsTerm(p *Mig_Obj_t) int32 {
	return migBool(Mig_FanIsNone(p, 1) != 0 && Mig_FanIsNone(p, 2) == 0)
}
func Mig_ObjIsCi(p *Mig_Obj_t) int32 {
	return migBool(Mig_ObjIsTerm(p) != 0 && Mig_FanIsNone(p, 0) != 0)
}
func Mig_ObjIsCo(p *Mig_Obj_t) int32 {
	return migBool(Mig_ObjIsTerm(p) != 0 && Mig_FanIsNone(p, 0) == 0)
}
func Mig_ObjIsBuf(p *Mig_Obj_t) int32 {
	return migBool(Mig_FanIsNone(p, 1) != 0 && Mig_FanIsNone(p, 2) != 0 && Mig_FanIsNone(p, 0) == 0)
}
func Mig_ObjIsNode(p *Mig_Obj_t) int32 { return migBool(Mig_FanIsNone(p, 1) == 0) }
func Mig_ObjIsNode2(p *Mig_Obj_t) int32 {
	return migBool(Mig_ObjIsNode(p) != 0 && Mig_FanIsNone(p, 2) != 0)
}
func Mig_ObjIsNode3(p *Mig_Obj_t) int32 {
	return migBool(Mig_ObjIsNode(p) != 0 && Mig_FanIsNone(p, 2) == 0)
}
func Mig_ObjIsAnd(p *Mig_Obj_t) int32 {
	return migBool(Mig_ObjIsNode2(p) != 0 && Mig_FanId(p, 0) < Mig_FanId(p, 1))
}
func Mig_ObjIsXor(p *Mig_Obj_t) int32 {
	return migBool(Mig_ObjIsNode2(p) != 0 && Mig_FanId(p, 0) > Mig_FanId(p, 1))
}
func Mig_ObjIsMux(p *Mig_Obj_t) int32  { return Mig_ObjIsNode3(p) }
func Mig_ObjIsCand(p *Mig_Obj_t) int32 { return migBool(Mig_ObjIsNode(p) != 0 || Mig_ObjIsCi(p) != 0) }
func Mig_ObjNodeType(p *Mig_Obj_t) int32 {
	if Mig_ObjIsAnd(p) != 0 {
		return 1
	}
	if Mig_ObjIsXor(p) != 0 {
		return 2
	}
	return 3
}

func Mig_ObjId(p *Mig_Obj_t) int32       { return Mig_FanId(p, 3) }
func Mig_ObjSetId(p *Mig_Obj_t, v int32) { Mig_FanSetId(p, 3, v) }
func Mig_ObjCioId(p *Mig_Obj_t) int32 {
	assert((Mig_ObjIsTerm(p) != 0), "Mig_ObjIsTerm(p)")
	return Mig_FanId(p, 2)
}
func Mig_ObjSetCioId(p *Mig_Obj_t, v int32) {
	assert((Mig_FanIsNone(p, 1) != 0), "Mig_FanIsNone(p, 1)")
	Mig_FanSetId(p, 2, v)
}
func Mig_ObjPhase(p *Mig_Obj_t) int32       { return Mig_FanCompl(p, 3) }
func Mig_ObjSetPhase(p *Mig_Obj_t, v int32) { Mig_FanSetCompl(p, 3, v) }

func Mig_ObjMan(p *Mig_Obj_t) *Mig_Man_t { return p.owner }

// static inline Mig_Obj_t ** Mig_ObjPageP( Mig_Obj_t * p )                       { return *((Mig_Obj_t***)(p - Mig_IdCell(Mig_ObjId(p))) - 1);}
func Mig_ObjObj(p *Mig_Obj_t, i int32) *Mig_Obj_t { return Mig_ManObj(Mig_ObjMan(p), i) }

func Mig_ManIdToCioId(p *Mig_Man_t, Id int32) int32  { return Mig_ObjCioId(Mig_ManObj(p, Id)) }
func Mig_ManCiIdToId(p *Mig_Man_t, CiId int32) int32 { return Mig_ObjId(Mig_ManCi(p, CiId)) }
func Mig_ManCoIdToId(p *Mig_Man_t, CoId int32) int32 { return Mig_ObjId(Mig_ManCo(p, CoId)) }

func Mig_ObjIsPi(p *Mig_Obj_t) int32 {
	return migBool(Mig_ObjIsCi(p) != 0 && Mig_ObjCioId(p) < Mig_ManPiNum(Mig_ObjMan(p)))
}
func Mig_ObjIsPo(p *Mig_Obj_t) int32 {
	return migBool(Mig_ObjIsCo(p) != 0 && Mig_ObjCioId(p) < Mig_ManPoNum(Mig_ObjMan(p)))
}
func Mig_ObjIsRo(p *Mig_Obj_t) int32 {
	return migBool(Mig_ObjIsCi(p) != 0 && Mig_ObjCioId(p) >= Mig_ManPiNum(Mig_ObjMan(p)))
}
func Mig_ObjIsRi(p *Mig_Obj_t) int32 {
	return migBool(Mig_ObjIsCo(p) != 0 && Mig_ObjCioId(p) >= Mig_ManPoNum(Mig_ObjMan(p)))
}

func Mig_ObjRoToRi(p *Mig_Obj_t) *Mig_Obj_t {
	var pMan *Mig_Man_t = Mig_ObjMan(p)
	assert((Mig_ObjIsRo(p) != 0), "Mig_ObjIsRo(p)")
	return Mig_ManCo(pMan, Mig_ManCoNum(pMan)-Mig_ManCiNum(pMan)+Mig_ObjCioId(p))
}
func Mig_ObjRiToRo(p *Mig_Obj_t) *Mig_Obj_t {
	var pMan *Mig_Man_t = Mig_ObjMan(p)
	assert((Mig_ObjIsRi(p) != 0), "Mig_ObjIsRi(p)")
	return Mig_ManCi(pMan, Mig_ManCiNum(pMan)-Mig_ManCoNum(pMan)+Mig_ObjCioId(p))
}

func Mig_ObjHasFanin(p *Mig_Obj_t, i int32) int32 {
	return migBool(i < 3 && Mig_FanId(p, i) != MIG_NONE)
}
func Mig_ObjFaninId(p *Mig_Obj_t, i int32) int32 {
	assert(i < 3 && Mig_FanId(p, i) < Mig_ObjId(p), "i < 3 && Mig_FanId(p, i) < Mig_ObjId(p)")
	return Mig_FanId(p, i)
}
func Mig_ObjFaninId0(p *Mig_Obj_t) int32 { return Mig_FanId(p, 0) }
func Mig_ObjFaninId1(p *Mig_Obj_t) int32 { return Mig_FanId(p, 1) }
func Mig_ObjFaninId2(p *Mig_Obj_t) int32 { return Mig_FanId(p, 2) }
func Mig_ObjFanin(p *Mig_Obj_t, i int32) *Mig_Obj_t {
	return Mig_ManObj(Mig_ObjMan(p), Mig_ObjFaninId(p, i))
}

// static inline Mig_Obj_t *  Mig_ObjFanin( Mig_Obj_t * p, int i )                { return Mig_ObjPageP(p)[Mig_IdPage(Mig_ObjFaninId(p, i))] + Mig_IdCell(Mig_ObjFaninId(p, i));    }
func Mig_ObjFanin0(p *Mig_Obj_t) *Mig_Obj_t {
	if Mig_FanIsNone(p, 0) != 0 {
		return nil
	} else {
		return Mig_ObjFanin(p, 0)
	}
}
func Mig_ObjFanin1(p *Mig_Obj_t) *Mig_Obj_t {
	if Mig_FanIsNone(p, 1) != 0 {
		return nil
	} else {
		return Mig_ObjFanin(p, 1)
	}
}
func Mig_ObjFanin2(p *Mig_Obj_t) *Mig_Obj_t {
	if Mig_FanIsNone(p, 2) != 0 {
		return nil
	} else {
		return Mig_ObjFanin(p, 2)
	}
}
func Mig_ObjFaninC(p *Mig_Obj_t, i int32) int32 { assert(i < 3, "i < 3"); return Mig_FanCompl(p, i) }
func Mig_ObjFaninC0(p *Mig_Obj_t) int32         { return Mig_FanCompl(p, 0) }
func Mig_ObjFaninC1(p *Mig_Obj_t) int32         { return Mig_FanCompl(p, 1) }
func Mig_ObjFaninC2(p *Mig_Obj_t) int32         { return Mig_FanCompl(p, 2) }
func Mig_ObjFaninLit(p *Mig_Obj_t, i int32) int32 {
	return util.Abc_Var2Lit(Mig_FanId(p, i), Mig_FanCompl(p, i))
}
func Mig_ObjFlipFaninC(p *Mig_Obj_t, i int32) { Mig_FanSetCompl(p, i, Mig_FanCompl(p, i)^1) }
func Mig_ObjWhatFanin(p *Mig_Obj_t, i int32) int32 {
	if Mig_FanId(p, 0) == i {
		return 0
	}
	if Mig_FanId(p, 1) == i {
		return 1
	}
	if Mig_FanId(p, 2) == i {
		return 2
	}
	return -1
}
func Mig_ObjSetFaninLit(p *Mig_Obj_t, i int32, l int32) {
	assert(l >= 0 && (l>>1) < Mig_ObjId(p), "l >= 0 && (l >> 1) < Mig_ObjId(p)")
	Mig_FanSetId(p, i, util.Abc_Lit2Var(l))
	Mig_FanSetCompl(p, i, util.Abc_LitIsCompl(l))
}

func Mig_ObjSiblId(p *Mig_Obj_t) int32 {
	if Vec_IntSize(&Mig_ObjMan(p).vSibls) == 0 {
		return 0
	} else {
		return Vec_IntEntry(&Mig_ObjMan(p).vSibls, int(Mig_ObjId(p)))
	}
}
func Mig_ObjSetSiblId(p *Mig_Obj_t, s int32) {
	assert(s > 0 && Mig_ObjId(p) > s, "s > 0 && Mig_ObjId(p) > s")
	Vec_IntWriteEntry(&Mig_ObjMan(p).vSibls, int(Mig_ObjId(p)), s)
}
func Mig_ObjSibl(p *Mig_Obj_t) *Mig_Obj_t {
	if Mig_ObjSiblId(p) == 0 {
		return nil
	} else {
		return Mig_ObjObj(p, Mig_ObjSiblId(p))
	}
}
func Mig_ObjRefNum(p *Mig_Obj_t) int32 { return Vec_IntEntry(&Mig_ObjMan(p).vRefs, int(Mig_ObjId(p))) }

func Mig_ManCleanCopy(p *Mig_Man_t) {
	if Vec_IntArray(&p.vCopies) == nil {
		Vec_IntFill(&p.vCopies, int(Mig_ManObjNum(p)), -1)
	}
}
func Mig_ObjCopy(p *Mig_Obj_t) int32 {
	if Vec_IntSize(&Mig_ObjMan(p).vCopies) == 0 {
		return -1
	} else {
		return Vec_IntEntry(&Mig_ObjMan(p).vCopies, int(Mig_ObjId(p)))
	}
}
func Mig_ObjSetCopy(p *Mig_Obj_t, i int32) {
	assert(Vec_IntSize(&Mig_ObjMan(p).vCopies) != 0, "Vec_IntSize(&Mig_ObjMan(p).vCopies) != 0")
	Vec_IntWriteEntry(&Mig_ObjMan(p).vCopies, int(Mig_ObjId(p)), i)
}

func Mig_ManIncrementTravId(p *Mig_Man_t) {
	if Vec_IntArray(&p.vTravIds) == nil {
		Vec_IntFill(&p.vTravIds, int(Mig_ManObjNum(p))+500, 0)
	}
	p.nTravIds++
}
func Mig_ObjIncrementTravId(p *Mig_Obj_t) {
	if Vec_IntArray(&Mig_ObjMan(p).vTravIds) == nil {
		Vec_IntFill(&Mig_ObjMan(p).vTravIds, int(Mig_ManObjNum(Mig_ObjMan(p)))+500, 0)
	}
	Mig_ObjMan(p).nTravIds++
}
func Mig_ObjSetTravIdCurrent(p *Mig_Obj_t) {
	Vec_IntSetEntry(&Mig_ObjMan(p).vTravIds, int(Mig_ObjId(p)), Mig_ObjMan(p).nTravIds)
}
func Mig_ObjSetTravIdPrevious(p *Mig_Obj_t) {
	Vec_IntSetEntry(&Mig_ObjMan(p).vTravIds, int(Mig_ObjId(p)), Mig_ObjMan(p).nTravIds-1)
}
func Mig_ObjIsTravIdCurrent(p *Mig_Obj_t) int32 {
	return migBool(Vec_IntGetEntry(&Mig_ObjMan(p).vTravIds, int(Mig_ObjId(p))) == Mig_ObjMan(p).nTravIds)
}
func Mig_ObjIsTravIdPrevious(p *Mig_Obj_t) int32 {
	return migBool(Vec_IntGetEntry(&Mig_ObjMan(p).vTravIds, int(Mig_ObjId(p))) == Mig_ObjMan(p).nTravIds-1)
}
func Mig_ObjSetTravIdCurrentId(p *Mig_Man_t, Id int32) {
	Vec_IntSetEntry(&p.vTravIds, int(Id), p.nTravIds)
}
func Mig_ObjIsTravIdCurrentId(p *Mig_Man_t, Id int32) int32 {
	return migBool(Vec_IntGetEntry(&p.vTravIds, int(Id)) == p.nTravIds)
}

/*
*Function*************************************************************

	Synopsis    []

	Description []

	SideEffects []

	SeeAlso     []

**********************************************************************
*/
func Mig_ManAppendObj(p *Mig_Man_t) *Mig_Obj_t {
	var pObj *Mig_Obj_t
	assert(p.nObjs < MIG_NONE, "p.nObjs < MIG_NONE")
	if int(p.nObjs) >= (Vec_PtrSize(&p.vPages) << MIG_BASE) {
		assert(int(p.nObjs) == (Vec_PtrSize(&p.vPages)<<MIG_BASE), "object page boundary")
		// The extra sentinel and unused slots have the all-ones C bitfields.
		page := make([]Mig_Obj_t, MIG_MASK+2)
		for i := range page {
			page[i].owner = p
			for j := range page[i].pFans {
				page[i].pFans[j] = Mig_Fan_t{fCompl: 1, Id: MIG_NONE}
			}
		}
		Vec_PtrPush(&p.vPages, page)
	}
	p.nObjs++
	pObj = Mig_ManObj(p, p.nObjs-1)
	assert((Mig_ObjIsNone(pObj) != 0), "Mig_ObjIsNone(pObj)")
	Mig_ObjSetId(pObj, p.nObjs-1)
	return pObj
}
func Mig_ManAppendCi(p *Mig_Man_t) int32 {
	var pObj *Mig_Obj_t = Mig_ManAppendObj(p)
	Mig_ObjSetCioId(pObj, int32(Vec_IntSize(&p.vCis)))
	Vec_IntPush(&p.vCis, Mig_ObjId(pObj))
	return Mig_ObjId(pObj) << 1
}
func Mig_ManAppendCo(p *Mig_Man_t, iLit0 int32) int32 {
	var pObj *Mig_Obj_t
	assert(!(Mig_ObjIsCo(Mig_ManObj(p, util.Abc_Lit2Var(iLit0))) != 0), "!Mig_ObjIsCo(Mig_ManObj(p, util.Abc_Lit2Var(iLit0)))")
	pObj = Mig_ManAppendObj(p)
	Mig_ObjSetFaninLit(pObj, 0, iLit0)
	Mig_ObjSetCioId(pObj, int32(Vec_IntSize(&p.vCos)))
	Vec_IntPush(&p.vCos, Mig_ObjId(pObj))
	return Mig_ObjId(pObj) << 1
}
func Mig_ManAppendBuf(p *Mig_Man_t, iLit0 int32) int32 {
	var pObj *Mig_Obj_t
	pObj = Mig_ManAppendObj(p)
	Mig_ObjSetFaninLit(pObj, 0, iLit0)
	return Mig_ObjId(pObj) << 1
}
func Mig_ManAppendAnd(p *Mig_Man_t, iLit0 int32, iLit1 int32) int32 {
	var pObj *Mig_Obj_t = Mig_ManAppendObj(p)
	assert(iLit0 != iLit1, "iLit0 != iLit1")
	Mig_ObjSetFaninLit(pObj, 0, min(iLit0, iLit1))
	Mig_ObjSetFaninLit(pObj, 1, max(iLit0, iLit1))
	return Mig_ObjId(pObj) << 1
}
func Mig_ManAppendXor(p *Mig_Man_t, iLit0 int32, iLit1 int32) int32 {
	var pObj *Mig_Obj_t = Mig_ManAppendObj(p)
	assert(iLit0 != iLit1, "iLit0 != iLit1")
	assert(util.Abc_LitIsCompl(iLit0) == 0 && util.Abc_LitIsCompl(iLit1) == 0, "util.Abc_LitIsCompl(iLit0) == 0 && util.Abc_LitIsCompl(iLit1) == 0")
	Mig_ObjSetFaninLit(pObj, 0, max(iLit0, iLit1))
	Mig_ObjSetFaninLit(pObj, 1, min(iLit0, iLit1))
	return Mig_ObjId(pObj) << 1
}
func Mig_ManAppendMux(p *Mig_Man_t, iLit0 int32, iLit1 int32, iCtrl int32) int32 {
	var pObj *Mig_Obj_t = Mig_ManAppendObj(p)
	assert(iLit0 != iLit1 && iLit0 != iCtrl && iLit1 != iCtrl, "iLit0 != iLit1 && iLit0 != iCtrl && iLit1 != iCtrl")
	assert(util.Abc_LitIsCompl(iLit0) == 0 || util.Abc_LitIsCompl(iLit1) == 0, "util.Abc_LitIsCompl(iLit0) == 0 || util.Abc_LitIsCompl(iLit1) == 0")
	Mig_ObjSetFaninLit(pObj, 0, min(iLit0, iLit1))
	Mig_ObjSetFaninLit(pObj, 1, max(iLit0, iLit1))
	if iLit0 >= iLit1 {
		iCtrl = util.Abc_LitNot(iCtrl)
	}
	Mig_ObjSetFaninLit(pObj, 2, iCtrl)
	return Mig_ObjId(pObj) << 1
}
func Mig_ManAppendMaj(p *Mig_Man_t, iLit0 int32, iLit1 int32, iLit2 int32) int32 {
	var pObj *Mig_Obj_t = Mig_ManAppendObj(p)
	assert(iLit0 != iLit1 && iLit0 != iLit2 && iLit1 != iLit2, "iLit0 != iLit1 && iLit0 != iLit2 && iLit1 != iLit2")
	Mig_ObjSetFaninLit(pObj, 0, max(iLit0, iLit1))
	Mig_ObjSetFaninLit(pObj, 1, min(iLit0, iLit1))
	Mig_ObjSetFaninLit(pObj, 2, iLit2)
	return Mig_ObjId(pObj) << 1
}

////////////////////////////////////////////////////////////////////////
///                      MACRO DEFINITIONS                           ///
////////////////////////////////////////////////////////////////////////

// iterators over objects
// Original macro:
// #define Mig_ManForEachObj( p, pObj )                                    \
//     for ( p->iPage = 0; p->iPage < Vec_PtrSize(&p->vPages) &&           \
//         ((p->pPage) = (Mig_Obj_t *)Vec_PtrEntry(&p->vPages, p->iPage)); p->iPage++ ) \
//         for ( pObj = p->pPage; !Mig_ObjIsNone(pObj); pObj++ )

// Original macro:
// #define Mig_ManForEachObj1( p, pObj )                                   \
//     for ( p->iPage = 0; p->iPage < Vec_PtrSize(&p->vPages) &&           \
//         ((p->pPage) = (Mig_Obj_t *)Vec_PtrEntry(&p->vPages, p->iPage)); p->iPage++ ) \
//         for ( pObj = p->pPage + (p->iPage == 0); !Mig_ObjIsNone(pObj); pObj++ )

// Original macro:
// #define Mig_ManForEachObjReverse( p, pObj )                             \
//     for ( p->iPage = Vec_PtrSize(&p->vPages) - 1; p->iPage >= 0 &&      \
//         ((p->pPage) = (Mig_Obj_t *)Vec_PtrEntry(&p->vPages, p->iPage)); p->iPage-- ) \
//         for ( pObj = (p->iPage == Vec_PtrSize(&p->vPages) - 1) ?        \
//             Mig_ManObj(p, Mig_ManObjNum(p)-1) :  p->pPage + MIG_MASK;   \
//                 pObj - p->pPage >= 0; pObj-- )

// Original macro:
// #define Mig_ManForEachObjVec( vVec, p, pObj, i )                        \
//     for ( i = 0; (i < Vec_IntSize(vVec)) && ((pObj) = Mig_ManObj(p, Vec_IntEntry(vVec,i))); i++ )

// Original macro:
// #define Mig_ManForEachNode( p, pObj )                                   \
//     Mig_ManForEachObj( p, pObj ) if ( !Mig_ObjIsNode(pObj) ) {} else

// Original macro:
// #define Mig_ManForEachCand( p, pObj )                                   \
//     Mig_ManForEachObj( p, pObj ) if ( !Mig_ObjIsCand(pObj) ) {} else

// Original macro:
// #define Mig_ManForEachCi( p, pObj, i )                                  \
//     for ( i = 0; (i < Vec_IntSize(&p->vCis)) && ((pObj) = Mig_ManCi(p, i)); i++ )

// Original macro:
// #define Mig_ManForEachCo( p, pObj, i )                                  \
//     for ( i = 0; (i < Vec_IntSize(&p->vCos)) && ((pObj) = Mig_ManCo(p, i)); i++ )

// iterators over fanins
// Original macro:
// #define Mig_ObjForEachFaninId( p, iFanin, i )                           \
//     for ( i = 0; Mig_ObjHasFanin(p, i) && ((iFanin) = Mig_ObjFaninId(p, i)); i++ )

// Original macro:
// #define Mig_ObjForEachFanin( p, pFanin, i )                             \
//     for ( i = 0; Mig_ObjHasFanin(p, i) && ((pFanin) = Mig_ObjFanin(p, i)); i++ )

////////////////////////////////////////////////////////////////////////
///                    FUNCTION DECLARATIONS                         ///
////////////////////////////////////////////////////////////////////////

/*=== mpmMig.c ===========================================================*/
// extern Mig_Man_t *           Mig_ManStart();

// extern void                  Mig_ManStop( Mig_Man_t * p );

// extern void                  Mig_ManSetRefs( Mig_Man_t * p );

// extern int                   Mig_ManAndNum( Mig_Man_t * p );

// extern int                   Mig_ManXorNum( Mig_Man_t * p );

// extern int                   Mig_ManMuxNum( Mig_Man_t * p );

//ABC_NAMESPACE_HEADER_END

////////////////////////////////////////////////////////////////////////
///                       END OF FILE                                ///
////////////////////////////////////////////////////////////////////////
