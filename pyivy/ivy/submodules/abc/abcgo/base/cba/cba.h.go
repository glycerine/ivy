package cba

import (
	"fmt"
	"strconv"
	"strings"
	"unsafe"
)

/**CFile****************************************************************

  FileName    [cba.h]

  SystemName  [ABC: Logic synthesis and verification system.]

  PackageName [Hierarchical word-level netlist.]

  Synopsis    [External declarations.]

  Author      [Alan Mishchenko]

  Affiliation [UC Berkeley]

  Date        [Ver. 1.0. Started - July 21, 2015.]

  Revision    [$Id: cba.h,v 1.00 2014/11/29 00:00:00 alanmi Exp $]

***********************************************************************/

////////////////////////////////////////////////////////////////////////
///                          INCLUDES                                ///
////////////////////////////////////////////////////////////////////////

//#include "aig/gia/gia.h"
//#include "misc/extra/extra.h"
//#include "misc/util/utilNam.h"
//#include "misc/vec/vecHash.h"
//#include "cbaTypes.h"

////////////////////////////////////////////////////////////////////////
///                         PARAMETERS                               ///
////////////////////////////////////////////////////////////////////////

//ABC_NAMESPACE_HEADER_START

////////////////////////////////////////////////////////////////////////
///                         BASIC TYPES                              ///
////////////////////////////////////////////////////////////////////////

type Cba_Ntk_t = Cba_Ntk_t_
type Cba_Man_t = Cba_Man_t_

// network
type Cba_Ntk_t_ struct {
	pDesign *Cba_Man_t // design
	Id      int32      // network ID
	NameId  int32      // name ID
	iCopy   int32      // copy module
	Mark    uint32     /* C bitfield: 1 bits; review packing and truncation */ // visit mark
	fComb   uint32     /* C bitfield: 1 bits; review packing and truncation */ // the network is combinational
	fSeq    uint32     /* C bitfield: 1 bits; review packing and truncation */ // the network is sequential
	// interface
	vInputs  Vec_Int_t // inputs
	vOutputs Vec_Int_t // outputs
	vOrder   Vec_Int_t // order
	vSeq     Vec_Int_t // sequential objects
	// stucture
	vObjType Vec_Str_t // type
	vObjFin0 Vec_Int_t // fanins
	vObjFon0 Vec_Int_t // outputs
	vFinFon  Vec_Int_t // fons
	vFonObj  Vec_Int_t // object
	// optional
	vObjCopy  Vec_Int_t // copy
	vObjFunc  Vec_Int_t // function
	vObjName  Vec_Int_t // name
	vObjAttr  Vec_Int_t // attribute offset
	vAttrSto  Vec_Int_t // attribute storage
	vFonCopy  Vec_Int_t // copy
	vFonName  Vec_Int_t // name
	vFonRange Vec_Int_t // range
	vFonPrev  Vec_Int_t // fanout: prev fon
	vFonNext  Vec_Int_t // fanout: next fon
	vFinFon0  Vec_Int_t // fanout: first fon
	vFinObj   Vec_Int_t // object
	vNtkObjs  Vec_Int_t // instances
	vFonBits  Vec_Int_t // fon mapping into AIG nodes
	// other
	vOther  *Vec_Ptr_t // various data
	vArray0 Vec_Int_t
	vArray1 Vec_Int_t
}

// design
type Cba_Man_t_ struct {
	// design names
	pName      string         // design name
	pSpec      string         // spec file name
	pStrs      *Abc_Nam_t     // string manager
	pFuns      *Abc_Nam_t     // constant manager
	pMods      *Abc_Nam_t     // module name manager
	vHash      *Hash_IntMan_t // variable ranges
	vNameMap   Vec_Int_t      // mapping names
	vNameMap2  Vec_Int_t      // mapping names
	vUsed      Vec_Int_t      // used map entries
	vUsed2     Vec_Int_t      // used map entries
	pTypeNames [CBA_BOX_LAST]string
	nObjs      [CBA_BOX_LAST]int32 // counter of objects of each type
	nAnds      [CBA_BOX_LAST]int32 // counter of AND gates after blasting
	// internal data
	iRoot int32     // root network
	vNtks Vec_Ptr_t // networks
	// user data
	nOpens  int32
	vOut    Vec_Str_t
	vOut2   Vec_Str_t
	pMioLib any
}

////////////////////////////////////////////////////////////////////////
///                      MACRO DEFINITIONS                           ///
////////////////////////////////////////////////////////////////////////

func Cba_ManName(p *Cba_Man_t) string            { return p.pName }
func Cba_ManSpec(p *Cba_Man_t) string            { return p.pSpec }
func Cba_ManNtkNum(p *Cba_Man_t) int32           { return Vec_PtrSize(&p.vNtks) - 1 }
func Cba_ManNtkIsOk(p *Cba_Man_t, i int32) int32 { return boolToInt(i > 0 && i <= Cba_ManNtkNum(p)) }
func Cba_ManNtk(p *Cba_Man_t, i int32) *Cba_Ntk_t {
	if Cba_ManNtkIsOk(p, i) != 0 {
		value := Vec_PtrEntry(&p.vNtks, i)
		if value == nil {
			return nil
		}
		return value.(*Cba_Ntk_t)
	} else {
		return nil
	}
}
func Cba_ManNtkFindId(p *Cba_Man_t, pName string) int32 { return Abc_NamStrFind(p.pMods, pName) }
func Cba_ManNtkFind(p *Cba_Man_t, pName string) *Cba_Ntk_t {
	return Cba_ManNtk(p, Cba_ManNtkFindId(p, pName))
}
func Cba_ManRoot(p *Cba_Man_t) *Cba_Ntk_t            { return Cba_ManNtk(p, p.iRoot) }
func Cba_ManStr(p *Cba_Man_t, i int32) string        { return Abc_NamStr(p.pStrs, i) }
func Cba_ManStrId(p *Cba_Man_t, pStr string) int32   { return Abc_NamStrFind(p.pStrs, pStr) }
func Cba_ManNameIdMax(p *Cba_Man_t) int32            { return Abc_NamObjNumMax(p.pStrs) + 1 }
func Cba_ManConst(p *Cba_Man_t, i int32) string      { return Abc_NamStr(p.pFuns, i) }
func Cba_ManConstId(p *Cba_Man_t, pStr string) int32 { return Abc_NamStrFind(p.pFuns, pStr) }
func Cba_ManConstIdMax(p *Cba_Man_t) int32           { return Abc_NamObjNumMax(p.pFuns) + 1 }

func Cba_NtkMan(p *Cba_Ntk_t) *Cba_Man_t              { return p.pDesign }
func Cba_NtkNtk(p *Cba_Ntk_t, i int32) *Cba_Ntk_t     { return Cba_ManNtk(p.pDesign, i) }
func Cba_NtkId(p *Cba_Ntk_t) int32                    { return p.Id }
func Cba_NtkPi(p *Cba_Ntk_t, i int32) int32           { return Vec_IntEntry(&p.vInputs, i) }
func Cba_NtkPo(p *Cba_Ntk_t, i int32) int32           { return Vec_IntEntry(&p.vOutputs, i) }
func Cba_NtkPioOrder(p *Cba_Ntk_t, i int32) int32     { return Vec_IntEntry(&p.vOrder, i) }
func Cba_NtkBoxSeq(p *Cba_Ntk_t, i int32) int32       { return Vec_IntEntry(&p.vSeq, i) }
func Cba_NtkNam(p *Cba_Ntk_t) *Abc_Nam_t              { return p.pDesign.pStrs }
func Cba_NtkStr(p *Cba_Ntk_t, i int32) string         { return Cba_ManStr(p.pDesign, i) }
func Cba_NtkStrId(p *Cba_Ntk_t, pName string) int32   { return Cba_ManStrId(p.pDesign, pName) }
func Cba_NtkConst(p *Cba_Ntk_t, i int32) string       { return Cba_ManConst(p.pDesign, i) }
func Cba_NtkSop(p *Cba_Ntk_t, i int32) string         { return Cba_ManConst(p.pDesign, i) }
func Cba_NtkNameId(p *Cba_Ntk_t) int32                { return p.NameId }
func Cba_NtkName(p *Cba_Ntk_t) string                 { return Cba_NtkStr(p, p.NameId) }
func Cba_NtkTypeName(p *Cba_Ntk_t, Type int32) string { return p.pDesign.pTypeNames[Type] }
func Cba_NtkCopy(p *Cba_Ntk_t) int32                  { return p.iCopy }
func Cba_NtkCopyNtk(pNew *Cba_Man_t, p *Cba_Ntk_t) *Cba_Ntk_t {
	return Cba_ManNtk(pNew, Cba_NtkCopy(p))
}
func Cba_NtkSetCopy(p *Cba_Ntk_t, i int32) { assert(p.iCopy == 0, "p.iCopy == 0"); p.iCopy = i }
func Cba_NtkHashRange(p *Cba_Ntk_t, l int32, r int32) int32 {
	return Hash_Int2ManInsert(p.pDesign.vHash, l, r, 0)
}
func Cba_NtkRangeLeft(p *Cba_Ntk_t, h int32) int32 {
	if h != 0 {
		return Hash_IntObjData0(p.pDesign.vHash, h)
	} else {
		return 0
	}
}
func Cba_NtkRangeRight(p *Cba_Ntk_t, h int32) int32 {
	if h != 0 {
		return Hash_IntObjData1(p.pDesign.vHash, h)
	} else {
		return 0
	}
}
func Cba_NtkRangeSize(p *Cba_Ntk_t, h int32) int32 {
	l, r := Cba_NtkRangeLeft(p, h), Cba_NtkRangeRight(p, h)
	if l > r {
		return 1 + l - r
	}
	return 1 + r - l
}

func Cba_NtkPiNum(p *Cba_Ntk_t) int32       { return Vec_IntSize(&p.vInputs) }
func Cba_NtkPoNum(p *Cba_Ntk_t) int32       { return Vec_IntSize(&p.vOutputs) }
func Cba_NtkPioNum(p *Cba_Ntk_t) int32      { return Cba_NtkPiNum(p) + Cba_NtkPoNum(p) }
func Cba_NtkPiNumAlloc(p *Cba_Ntk_t) int32  { return Vec_IntCap(&p.vInputs) }
func Cba_NtkPoNumAlloc(p *Cba_Ntk_t) int32  { return Vec_IntCap(&p.vOutputs) }
func Cba_NtkPioOrderNum(p *Cba_Ntk_t) int32 { return Vec_IntSize(&p.vOrder) }
func Cba_NtkObjNum(p *Cba_Ntk_t) int32      { return Vec_StrSize(&p.vObjType) - 1 }
func Cba_NtkObjNumAlloc(p *Cba_Ntk_t) int32 { return Vec_StrCap(&p.vObjType) - 1 }
func Cba_NtkFinNum(p *Cba_Ntk_t) int32      { return Vec_IntSize(&p.vFinFon) - 1 }
func Cba_NtkFinNumAlloc(p *Cba_Ntk_t) int32 { return Vec_IntCap(&p.vFinFon) - 1 }
func Cba_NtkFonNum(p *Cba_Ntk_t) int32      { return Vec_IntSize(&p.vFonObj) - 1 }
func Cba_NtkFonNumAlloc(p *Cba_Ntk_t) int32 { return Vec_IntCap(&p.vFonObj) - 1 }
func Cba_NtkTypeNum(p *Cba_Ntk_t, Type int32) int32 {
	return Vec_StrCountEntry(&p.vObjType, byte(Type))
}
func Cba_NtkBoxNum(p *Cba_Ntk_t) int32     { return Cba_NtkObjNum(p) - Cba_NtkPioNum(p) }
func Cba_NtkBoxUserNum(p *Cba_Ntk_t) int32 { return Cba_NtkTypeNum(p, int32(CBA_OBJ_BOX)) }
func Cba_NtkBoxPrimNum(p *Cba_Ntk_t) int32 { return Vec_StrCountLarger(&p.vObjType, byte(CBA_OBJ_BOX)) }
func Cba_NtkBoxSeqNum(p *Cba_Ntk_t) int32  { return Vec_IntSize(&p.vSeq) }

func Cba_NtkCleanObjCopies(p *Cba_Ntk_t) { Vec_IntFill(&p.vObjCopy, Vec_StrCap(&p.vObjType), -1) }
func Cba_NtkCleanObjFuncs(p *Cba_Ntk_t)  { Vec_IntFill(&p.vObjFunc, Vec_StrCap(&p.vObjType), 0) }
func Cba_NtkCleanObjNames(p *Cba_Ntk_t)  { Vec_IntFill(&p.vObjName, Vec_StrCap(&p.vObjType), 0) }
func Cba_NtkCleanObjAttrs(p *Cba_Ntk_t) {
	Vec_IntFill(&p.vObjAttr, Vec_StrCap(&p.vObjType), 0)
	Vec_IntFill(&p.vAttrSto, 1, -1)
}
func Cba_NtkCleanFonCopies(p *Cba_Ntk_t) { Vec_IntFill(&p.vFonCopy, Vec_IntCap(&p.vFonObj), 0) }
func Cba_NtkCleanFonNames(p *Cba_Ntk_t)  { Vec_IntFill(&p.vFonName, Vec_IntCap(&p.vFonObj), 0) }
func Cba_NtkCleanFonRanges(p *Cba_Ntk_t) { Vec_IntFill(&p.vFonRange, Vec_IntCap(&p.vFonObj), 0) }
func Cba_NtkCleanFonPrevs(p *Cba_Ntk_t)  { Vec_IntFill(&p.vFonPrev, Vec_IntCap(&p.vFonObj), 0) }
func Cba_NtkCleanFonNexts(p *Cba_Ntk_t)  { Vec_IntFill(&p.vFonNext, Vec_IntCap(&p.vFonObj), 0) }
func Cba_NtkCleanFinFon0(p *Cba_Ntk_t)   { Vec_IntFill(&p.vFinFon0, Vec_IntCap(&p.vFinFon), 0) }
func Cba_NtkCleanFinObjs(p *Cba_Ntk_t)   { Vec_IntFill(&p.vFinObj, Vec_IntCap(&p.vFinFon), 0) }

func Cba_NtkHasObjCopies(p *Cba_Ntk_t) int32 { return boolToInt(Vec_IntSize(&p.vObjCopy) > 0) }
func Cba_NtkHasObjFuncs(p *Cba_Ntk_t) int32  { return boolToInt(Vec_IntSize(&p.vObjFunc) > 0) }
func Cba_NtkHasObjNames(p *Cba_Ntk_t) int32  { return boolToInt(Vec_IntSize(&p.vObjName) > 0) }
func Cba_NtkHasObjAttrs(p *Cba_Ntk_t) int32  { return boolToInt(Vec_IntSize(&p.vObjAttr) > 0) }
func Cba_NtkHasFonCopies(p *Cba_Ntk_t) int32 { return boolToInt(Vec_IntSize(&p.vFonCopy) > 0) }
func Cba_NtkHasFonNames(p *Cba_Ntk_t) int32  { return boolToInt(Vec_IntSize(&p.vFonName) > 0) }
func Cba_NtkHasFonRanges(p *Cba_Ntk_t) int32 { return boolToInt(Vec_IntSize(&p.vFonRange) > 0) }
func Cba_NtkHasFonPrevs(p *Cba_Ntk_t) int32  { return boolToInt(Vec_IntSize(&p.vFonPrev) > 0) }
func Cba_NtkHasFonNexts(p *Cba_Ntk_t) int32  { return boolToInt(Vec_IntSize(&p.vFonNext) > 0) }
func Cba_NtkHasFinFon0(p *Cba_Ntk_t) int32   { return boolToInt(Vec_IntSize(&p.vFinFon0) > 0) }
func Cba_NtkHasFinObjs(p *Cba_Ntk_t) int32   { return boolToInt(Vec_IntSize(&p.vFinObj) > 0) }

func Cba_NtkFreeObjCopies(p *Cba_Ntk_t) { Vec_IntErase(&p.vObjCopy) }
func Cba_NtkFreeObjFuncs(p *Cba_Ntk_t)  { Vec_IntErase(&p.vObjFunc) }
func Cba_NtkFreeObjNames(p *Cba_Ntk_t)  { Vec_IntErase(&p.vObjName) }
func Cba_NtkFreeObjAttrs(p *Cba_Ntk_t)  { Vec_IntErase(&p.vObjAttr) }
func Cba_NtkFreeFonCopies(p *Cba_Ntk_t) { Vec_IntErase(&p.vFonCopy) }
func Cba_NtkFreeFonNames(p *Cba_Ntk_t)  { Vec_IntErase(&p.vFonName) }
func Cba_NtkFreeFonRanges(p *Cba_Ntk_t) { Vec_IntErase(&p.vFonRange) }
func Cba_NtkFreeFonPrevs(p *Cba_Ntk_t)  { Vec_IntErase(&p.vFonPrev) }
func Cba_NtkFreeFonNexts(p *Cba_Ntk_t)  { Vec_IntErase(&p.vFonNext) }
func Cba_NtkFreeFinFon0(p *Cba_Ntk_t)   { Vec_IntErase(&p.vFinFon0) }
func Cba_NtkFreeFinObjs(p *Cba_Ntk_t)   { Vec_IntErase(&p.vFinObj) }

func Cba_ObjType(p *Cba_Ntk_t, i int32) Cba_ObjType_t {
	assert(i > 0, "i>0")
	return Cba_ObjType_t(uint8(Vec_StrEntry(&p.vObjType, i)))
}
func Cba_ObjCleanType(p *Cba_Ntk_t, i int32) {
	assert(i > 0, "i>0")
	Vec_StrWriteEntry(&p.vObjType, i, byte(CBA_OBJ_NONE))
}

func Cba_ObjIsPi(p *Cba_Ntk_t, i int32) int32 { return boolToInt(Cba_ObjType(p, i) == CBA_OBJ_PI) }
func Cba_ObjIsPo(p *Cba_Ntk_t, i int32) int32 { return boolToInt(Cba_ObjType(p, i) == CBA_OBJ_PO) }
func Cba_ObjIsPio(p *Cba_Ntk_t, i int32) int32 {
	return boolToInt(Cba_ObjIsPi(p, i) != 0 || Cba_ObjIsPo(p, i) != 0)
}
func Cba_ObjIsBox(p *Cba_Ntk_t, i int32) int32 { return Cba_TypeIsBox(Cba_ObjType(p, i)) }
func Cba_ObjIsBoxUser(p *Cba_Ntk_t, i int32) int32 {
	return boolToInt(Cba_ObjType(p, i) == CBA_OBJ_BOX)
}
func Cba_ObjIsBoxPrim(p *Cba_Ntk_t, i int32) int32 {
	return boolToInt(Cba_ObjType(p, i) > CBA_OBJ_BOX && Cba_ObjType(p, i) < CBA_BOX_LAST)
}
func Cba_ObjIsGate(p *Cba_Ntk_t, i int32) int32 { return boolToInt(Cba_ObjType(p, i) == CBA_BOX_GATE) }
func Cba_ObjIsSlice(p *Cba_Ntk_t, i int32) int32 {
	return boolToInt(Cba_ObjType(p, i) == CBA_BOX_SLICE)
}
func Cba_ObjIsConcat(p *Cba_Ntk_t, i int32) int32 {
	return boolToInt(Cba_ObjType(p, i) == CBA_BOX_CONCAT)
}
func Cba_ObjIsUnary(p *Cba_Ntk_t, i int32) int32 { return Cba_TypeIsUnary(Cba_ObjType(p, i)) }

func Cba_ObjFin0(p *Cba_Ntk_t, i int32) int32 {
	assert(i > 0, "i>0")
	return Vec_IntEntry(&p.vObjFin0, i)
}
func Cba_ObjFon0(p *Cba_Ntk_t, i int32) int32 {
	assert(i > 0, "i>0")
	return Vec_IntEntry(&p.vObjFon0, i)
}
func Cba_ObjFin(p *Cba_Ntk_t, i int32, k int32) int32 {
	assert(i > 0, "i>0")
	return Cba_ObjFin0(p, i) + k
}
func Cba_ObjFon(p *Cba_Ntk_t, i int32, k int32) int32 {
	assert(i > 0, "i>0")
	return Cba_ObjFon0(p, i) + k
}
func Cba_ObjFinNum(p *Cba_Ntk_t, i int32) int32 {
	assert(i > 0, "i>0")
	return Cba_ObjFin0(p, i+1) - Cba_ObjFin0(p, i)
}
func Cba_ObjFonNum(p *Cba_Ntk_t, i int32) int32 {
	assert(i > 0, "i>0")
	return Cba_ObjFon0(p, i+1) - Cba_ObjFon0(p, i)
}

func Cba_ObjCopy(p *Cba_Ntk_t, i int32) int32 {
	assert(i > 0, "i>0")
	assert((Cba_NtkHasObjCopies(p) != 0), "Cba_NtkHasObjCopies(p)")
	return Vec_IntGetEntryFull(&p.vObjCopy, i)
}
func Cba_ObjFunc(p *Cba_Ntk_t, i int32) int32 {
	assert(i > 0, "i>0")
	assert((Cba_NtkHasObjFuncs(p) != 0), "Cba_NtkHasObjFuncs(p)")
	return Vec_IntGetEntry(&p.vObjFunc, i)
}
func Cba_ObjName(p *Cba_Ntk_t, i int32) int32 {
	assert(i > 0, "i>0")
	assert((Cba_NtkHasObjNames(p) != 0), "Cba_NtkHasObjNames(p)")
	return Vec_IntGetEntry(&p.vObjName, i)
}
func Cba_ObjNameStr(p *Cba_Ntk_t, i int32) string {
	assert(i > 0, "i>0")
	return Cba_NtkStr(p, Cba_ObjName(p, i))
}
func Cba_ObjAttr(p *Cba_Ntk_t, i int32) int32 {
	assert(i >= 0, "i>=0")
	if Cba_NtkHasObjAttrs(p) != 0 {
		return Vec_IntGetEntry(&p.vObjAttr, i)
	} else {
		return 0
	}
}
func Cba_ObjAttrSize(p *Cba_Ntk_t, i int32) int32 {
	assert(i >= 0, "i>=0")
	if Cba_ObjAttr(p, i) != 0 {
		return Vec_IntEntry(&p.vAttrSto, Cba_ObjAttr(p, i))
	} else {
		return 0
	}
}
func Cba_ObjAttrArray(p *Cba_Ntk_t, i int32) []int32 {
	assert(i >= 0, "i>=0")
	if Cba_ObjAttr(p, i) != 0 {
		return Vec_IntArray(&p.vAttrSto)[Cba_ObjAttr(p, i)+1:]
	} else {
		return nil
	}
}
func Cba_ObjAttrValue(p *Cba_Ntk_t, i int32, x int32) int32 {
	var k int32
	var s int32 = Cba_ObjAttrSize(p, i)
	a := Cba_ObjAttrArray(p, i)
	for k = 0; k < s; k += 2 {
		if a[k] == x {
			return a[k+1]
		}
	}
	return 0
}

func Cba_ObjSetCopy(p *Cba_Ntk_t, i int32, x int32) {
	assert(Cba_ObjCopy(p, i) == -1, "Cba_ObjCopy(p, i) == -1")
	Vec_IntSetEntry(&p.vObjCopy, i, x)
}
func Cba_ObjSetFunc(p *Cba_Ntk_t, i int32, x int32) {
	assert(Cba_ObjFunc(p, i) == 0, "Cba_ObjFunc(p, i) ==  0")
	Vec_IntSetEntry(&p.vObjFunc, i, x)
}
func Cba_ObjSetName(p *Cba_Ntk_t, i int32, x int32) {
	assert(Cba_ObjName(p, i) == 0, "Cba_ObjName(p, i) ==  0")
	Vec_IntSetEntry(&p.vObjName, i, x)
}
func Cba_ObjSetAttrs(p *Cba_Ntk_t, i int32, a []int32, s int32) {
	assert(Cba_ObjAttr(p, i) == 0, "Cba_ObjAttr(p, i) == 0")
	if !(a != nil) {
		return
	}
	Vec_IntSetEntry(&p.vObjAttr, i, Vec_IntSize(&p.vAttrSto))
	Vec_IntPush(&p.vAttrSto, s)
	Vec_IntPushArray(&p.vAttrSto, a[:s], s)
}

func Cba_FinFon(p *Cba_Ntk_t, f int32) int32 {
	assert(f > 0, "f>0")
	return Vec_IntEntry(&p.vFinFon, f)
}
func Cba_PatchFinFon(p *Cba_Ntk_t, f int32, x int32) {
	assert(f > 0, "f>0")
	Vec_IntWriteEntry(&p.vFinFon, f, x)
}
func Cba_ObjFinFon(p *Cba_Ntk_t, i int32, k int32) int32 {
	assert(i > 0, "i>0")
	return Cba_FinFon(p, Cba_ObjFin(p, i, k))
}
func Cba_ObjFinFons(p *Cba_Ntk_t, i int32) *int32 {
	assert(i > 0, "i>0")
	return Vec_IntEntryP(&p.vFinFon, Cba_ObjFin0(p, i))
}

func Cba_ObjSetFinFon(p *Cba_Ntk_t, i int32, k int32, x int32) {
	assert(i > 0, "i>0")
	assert(Cba_ObjFinFon(p, i, k) == 0, "Cba_ObjFinFon(p, i, k)== 0")
	Vec_IntWriteEntry(&p.vFinFon, Cba_ObjFin(p, i, k), x)
}
func Cba_ObjCleanFinFon(p *Cba_Ntk_t, i int32, k int32) {
	assert(i > 0, "i>0")
	assert(Cba_ObjFinFon(p, i, k) != 0, "Cba_ObjFinFon(p, i, k)!= 0")
	Vec_IntWriteEntry(&p.vFinFon, Cba_ObjFin(p, i, k), 0)
}
func Cba_ObjPatchFinFon(p *Cba_Ntk_t, i int32, k int32, x int32) {
	assert(i > 0, "i>0")
	Cba_ObjCleanFinFon(p, i, k)
	Cba_ObjSetFinFon(p, i, k, x)
}

func Cba_ObjNtkId(p *Cba_Ntk_t, i int32) int32 {
	assert(i > 0, "i>0")
	if (Cba_ObjIsBoxUser(p, i) != 0) && (Cba_NtkHasObjFuncs(p) != 0) {
		return Cba_ObjFunc(p, i)
	} else {
		return 0
	}
}
func Cba_ObjNtk(p *Cba_Ntk_t, i int32) *Cba_Ntk_t {
	assert(i > 0, "i>0")
	return Cba_NtkNtk(p, Cba_ObjNtkId(p, i))
}
func Cba_ObjSetNtkId(p *Cba_Ntk_t, i int32, x int32) {
	assert(i > 0, "i>0")
	assert((Cba_ObjIsBoxUser(p, i) != 0), "Cba_ObjIsBoxUser(p, i)")
	Cba_ObjSetFunc(p, i, x)
}
func Cba_ObjIsSeq(p *Cba_Ntk_t, i int32) int32 {
	assert(i > 0, "i>0")
	if Cba_ObjIsBoxUser(p, i) != 0 {
		return int32(Cba_ObjNtk(p, i).fSeq)
	} else {
		return Cba_TypeIsSeq(Cba_ObjType(p, i))
	}
}

func Cba_FonConstRangeSize(p *Cba_Ntk_t, f int32) int32 {
	assert((Cba_FonIsConst(f) != 0), "Cba_FonIsConst(f)")
	value, _ := strconv.Atoi(Cba_NtkConst(p, Cba_FonConst(f)))
	return int32(value)
}
func Cba_FonConstSigned(p *Cba_Ntk_t, f int32) int32 {
	assert((Cba_FonIsConst(f) != 0), "Cba_FonIsConst(f)")
	return boolToInt(strings.Contains(Cba_NtkConst(p, Cba_FonConst(f)), "s"))
}

func Cba_FonObj(p *Cba_Ntk_t, f int32) int32 {
	if Cba_FonIsReal(f) != 0 {
		return Vec_IntEntry(&p.vFonObj, f)
	} else {
		return 0
	}
}
func Cba_FonRangeId(p *Cba_Ntk_t, f int32) int32 {
	assert((Cba_FonIsReal(f) != 0), "Cba_FonIsReal(f)")
	if Cba_NtkHasFonRanges(p) != 0 {
		return Vec_IntGetEntry(&p.vFonRange, f)
	} else {
		return 0
	}
}
func Cba_FonRange(p *Cba_Ntk_t, f int32) int32 {
	assert((Cba_FonIsReal(f) != 0), "Cba_FonIsReal(f)")
	return Abc_Lit2Var(Cba_FonRangeId(p, f))
}
func Cba_FonLeft(p *Cba_Ntk_t, f int32) int32  { return Cba_NtkRangeLeft(p, Cba_FonRange(p, f)) }
func Cba_FonRight(p *Cba_Ntk_t, f int32) int32 { return Cba_NtkRangeRight(p, Cba_FonRange(p, f)) }
func Cba_FonSigned(p *Cba_Ntk_t, f int32) int32 {
	if Cba_FonIsConst(f) != 0 {
		return Cba_FonConstSigned(p, f)
	} else {
		return Abc_LitIsCompl(Cba_FonRangeId(p, f))
	}
}
func Cba_FonRangeSize(p *Cba_Ntk_t, f int32) int32 {
	if Cba_FonIsConst(f) != 0 {
		return Cba_FonConstRangeSize(p, f)
	} else {
		return Cba_NtkRangeSize(p, Cba_FonRange(p, f))
	}
}
func Cba_FonSetRangeSign(p *Cba_Ntk_t, f int32, x int32) {
	assert((Cba_NtkHasFonRanges(p) != 0), "Cba_NtkHasFonRanges(p)")
	Vec_IntSetEntry(&p.vFonRange, f, x)
}
func Cba_FonSetRange(p *Cba_Ntk_t, f int32, x int32) {
	assert((Cba_NtkHasFonRanges(p) != 0), "Cba_NtkHasFonRanges(p)")
	Vec_IntSetEntry(&p.vFonRange, f, Abc_Var2Lit(x, 0))
}
func Cba_FonHashRange(p *Cba_Ntk_t, f int32, l int32, r int32) {
	Cba_FonSetRange(p, f, Cba_NtkHashRange(p, l, r))
}
func Cba_FonCopy(p *Cba_Ntk_t, f int32) int32 {
	if Cba_FonIsReal(f) != 0 {
		return Vec_IntEntry(&p.vFonCopy, f)
	} else {
		return f
	}
}
func Cba_FonSetCopy(p *Cba_Ntk_t, f int32, x int32) {
	assert((Cba_FonIsReal(f) != 0), "Cba_FonIsReal(f)")
	assert(Cba_FonCopy(p, f) == 0, "Cba_FonCopy(p, f) == 0")
	Vec_IntWriteEntry(&p.vFonCopy, f, x)
}
func Cba_FonName(p *Cba_Ntk_t, f int32) int32 {
	assert((Cba_NtkHasFonNames(p) != 0), "Cba_NtkHasFonNames(p)")
	assert((Cba_FonIsReal(f) != 0), "Cba_FonIsReal(f)")
	return Vec_IntGetEntry(&p.vFonName, f)
}
func Cba_FonNameStr(p *Cba_Ntk_t, f int32) string {
	assert((Cba_NtkHasFonNames(p) != 0), "Cba_NtkHasFonNames(p)")
	assert((Cba_FonIsReal(f) != 0), "Cba_FonIsReal(f)")
	return Cba_NtkStr(p, Cba_FonName(p, f))
}
func Cba_FonSetName(p *Cba_Ntk_t, f int32, x int32) {
	assert((Cba_NtkHasFonNames(p) != 0), "Cba_NtkHasFonNames(p)")
	assert((Cba_FonIsReal(f) != 0), "Cba_FonIsReal(f)")
	assert(Cba_FonName(p, f) == 0, "Cba_FonName(p, f) == 0")
	Vec_IntSetEntry(&p.vFonName, f, x)
}
func Cba_FonCleanName(p *Cba_Ntk_t, f int32) {
	assert((Cba_NtkHasFonNames(p) != 0), "Cba_NtkHasFonNames(p)")
	assert((Cba_FonIsReal(f) != 0), "Cba_FonIsReal(f)")
	assert(Cba_FonName(p, f) != 0, "Cba_FonName(p, f) != 0")
	Vec_IntSetEntry(&p.vFonName, f, 0)
}
func Cba_FonPatchName(p *Cba_Ntk_t, f int32, x int32) {
	assert((Cba_NtkHasFonNames(p) != 0), "Cba_NtkHasFonNames(p)")
	assert((Cba_FonIsReal(f) != 0), "Cba_FonIsReal(f)")
	Cba_FonCleanName(p, f)
	Cba_FonSetName(p, f, x)
}
func Cba_FonIndex(p *Cba_Ntk_t, f int32) int32 {
	assert((Cba_FonIsReal(f) != 0), "Cba_FonIsReal(f)")
	return f - Cba_ObjFon0(p, Cba_FonObj(p, f))
}
func Cba_FonNtkId(p *Cba_Ntk_t, f int32) int32 {
	assert((Cba_FonIsReal(f) != 0), "Cba_FonIsReal(f)")
	return Cba_ObjNtkId(p, Cba_FonObj(p, f))
}
func Cba_FonNtk(p *Cba_Ntk_t, f int32) *Cba_Ntk_t {
	assert((Cba_FonIsReal(f) != 0), "Cba_FonIsReal(f)")
	return Cba_ObjNtk(p, Cba_FonObj(p, f))
}

func Cba_ObjFanin(p *Cba_Ntk_t, i int32, k int32) int32 {
	assert(i > 0, "i>0")
	return Cba_FonObj(p, Cba_ObjFinFon(p, i, k))
}
func Cba_ObjLeft(p *Cba_Ntk_t, i int32) int32      { return Cba_FonLeft(p, Cba_ObjFon0(p, i)) }
func Cba_ObjRight(p *Cba_Ntk_t, i int32) int32     { return Cba_FonRight(p, Cba_ObjFon0(p, i)) }
func Cba_ObjRangeSize(p *Cba_Ntk_t, i int32) int32 { return Cba_FonRangeSize(p, Cba_ObjFon0(p, i)) }
func Cba_ObjSigned(p *Cba_Ntk_t, i int32) int32    { return Cba_FonSigned(p, Cba_ObjFon0(p, i)) }
func Cba_ObjSign(p *Cba_Ntk_t, i int32) int32 {
	return Abc_Var2Lit(Cba_FonRangeSize(p, i), Cba_ObjSigned(p, i))
}

////////////////////////////////////////////////////////////////////////
///                          ITERATORS                               ///
////////////////////////////////////////////////////////////////////////

// Original macro:
// #define Cba_ManForEachNtk( p, pNtk, i )                                   \
//     for ( i = 1; (i <= Cba_ManNtkNum(p)) && (((pNtk) = Cba_ManNtk(p, i)), 1); i++ )

// Original macro:
// #define Cba_NtkForEachPi( p, iObj, i )                                    \
//     for ( i = 0; (i < Cba_NtkPiNum(p))  && (((iObj) = Cba_NtkPi(p, i)), 1); i++ )

// Original macro:
// #define Cba_NtkForEachPo( p, iObj, i )                                    \
//     for ( i = 0; (i < Cba_NtkPoNum(p))  && (((iObj) = Cba_NtkPo(p, i)), 1); i++ )

// Original macro:
// #define Cba_NtkForEachBoxSeq( p, iObj, i )                                \
//     for ( i = 0; (i < Cba_NtkBoxSeqNum(p))  && (((iObj) = Cba_NtkBoxSeq(p, i)), 1); i++ )

// Original macro:
// #define Cba_NtkForEachPioOrder( p, iObj, i )                              \
//     for ( i = 0; (i < Cba_NtkPioOrderNum(p))  && (((iObj) = Cba_NtkPioOrder(p, i)), 1); i++ )

// Original macro:
// #define Cba_NtkForEachPiFon( p, iObj, iFon, i )                           \
//     for ( i = 0; (i < Cba_NtkPiNum(p))  && (((iObj) = Cba_NtkPi(p, i)), 1)  && (((iFon) = Cba_ObjFon0(p, iObj)), 1); i++ )

// Original macro:
// #define Cba_NtkForEachPoDriverFon( p, iObj, iFon, i )                     \
//     for ( i = 0; (i < Cba_NtkPoNum(p))  && (((iObj) = Cba_NtkPo(p, i)), 1)  && (((iFon) = Cba_ObjFinFon(p, iObj, 0)), 1); i++ )

// Original macro:
// #define Cba_NtkForEachPoDriver( p, iObj, i )                              \
//     for ( i = 0; (i < Cba_NtkPoNum(p))  && (((iObj) = Cba_ObjFanin(p, Cba_NtkPo(p, i), 0)), 1); i++ )

// Original macro:
// #define Cba_NtkForEachObj( p, i )                                         \
//     for ( i = 1; i < Vec_StrSize(&p->vObjType); i++ ) if ( !Cba_ObjType(p, i) ) {} else

// Original macro:
// #define Cba_NtkForEachObjType( p, Type, i )                               \
//     for ( i = 1; i < Vec_StrSize(&p->vObjType)  && (((Type) = Cba_ObjType(p, i)), 1); i++ ) if ( !Type ) {} else

// Original macro:
// #define Cba_NtkForEachBox( p, i )                                         \
//     for ( i = 1; i < Vec_StrSize(&p->vObjType); i++ ) if ( !Cba_ObjIsBox(p, i) ) {} else

// Original macro:
// #define Cba_NtkForEachBoxComb( p, i )                                     \
//     for ( i = 1; i < Vec_StrSize(&p->vObjType); i++ ) if ( !Cba_ObjIsBox(p, i) || Cba_ObjIsSeq(p, i) ) {} else

// Original macro:
// #define Cba_NtkForEachBoxUser( p, i )                                     \
//     for ( i = 1; i < Vec_StrSize(&p->vObjType); i++ ) if ( !Cba_ObjIsBoxUser(p, i) ) {} else

// Original macro:
// #define Cba_NtkForEachBoxPrim( p, i )                                     \
//     for ( i = 1; i < Vec_StrSize(&p->vObjType); i++ ) if ( !Cba_ObjIsBoxPrim(p, i) ) {} else

// Original macro:
// #define Cba_NtkForEachFon( p, i )                                         \
//     for ( i = 1; i < Vec_IntSize(&p->vFonObj); i++ )

// Original macro:
// #define Cba_NtkForEachFinFon( p, iFon, iFin )                             \
//     for ( iFin = 1; iFin < Vec_IntSize(&p->vFinFon) && (((iFon) = Vec_IntEntry(&p->vFinFon, iFin)), 1); iFin++ ) if ( !iFon ) {} else

// Original macro:
// #define Cba_NtkForEachFonName( p, NameId, iFon )                          \
//     for ( iFon = 1; iFon < Vec_IntSize(&p->vFonName) && (((NameId) = Vec_IntEntry(&p->vFonName, iFon)), 1); iFon++ ) if ( !NameId ) {} else

// Original macro:
// #define Cba_ObjForEachFin( p, iObj, iFin, k )                             \
//     for ( k = 0, iFin = Cba_ObjFin0(p, iObj); iFin < Cba_ObjFin0(p, iObj+1); iFin++, k++ )

// Original macro:
// #define Cba_ObjForEachFon( p, iObj, iFon, k )                             \
//     for ( k = 0, iFon = Cba_ObjFon0(p, iObj); iFon < Cba_ObjFon0(p, iObj+1); iFon++, k++ )

// Original macro:
// #define Cba_ObjForEachFinFon( p, iObj, iFin, iFon, k )                    \
//     for ( k = 0, iFin = Cba_ObjFin0(p, iObj); iFin < Cba_ObjFin0(p, iObj+1) && ((iFon = Cba_FinFon(p, iFin)), 1); iFin++, k++ )

// Original macro:
// #define Cba_ObjForEachFinFanin( p, iObj, iFin, iFanin, k )                \
//     for ( k = 0, iFin = Cba_ObjFin0(p, iObj); iFin < Cba_ObjFin0(p, iObj+1) && ((iFanin = Cba_FonObj(p, Cba_FinFon(p, iFin))), 1); iFin++, k++ )

// Original macro:
// #define Cba_ObjForEachFinFaninReal( p, iObj, iFin, iFanin, k )            \
//     for ( k = 0, iFin = Cba_ObjFin0(p, iObj); iFin < Cba_ObjFin0(p, iObj+1) && ((iFanin = Cba_FonObj(p, Cba_FinFon(p, iFin))), 1); iFin++, k++ ) if ( !iFanin ) {} else

////////////////////////////////////////////////////////////////////////
///                    FUNCTION DECLARATIONS                         ///
////////////////////////////////////////////////////////////////////////

/*
*Function*************************************************************

	Synopsis    [Object APIs.]

	Description []

	SideEffects []

	SeeAlso     []

**********************************************************************
*/
func Cba_ObjAlloc(p *Cba_Ntk_t, Type Cba_ObjType_t, nFins int32, nFons int32) int32 {
	var i int32
	var iObj int32 = Vec_StrSize(&p.vObjType)
	if Type == CBA_OBJ_PI {
		Vec_IntPush(&p.vInputs, iObj)
	} else if Type == CBA_OBJ_PO {
		Vec_IntPush(&p.vOutputs, iObj)
	}
	Vec_StrPush(&p.vObjType, byte(Type))
	// add fins
	for i = 0; i < nFins; i++ {
		Vec_IntPush(&p.vFinFon, 0)
	}
	if Vec_IntSize(&p.vObjFin0) != 0 {
		Vec_IntPush(&p.vObjFin0, Vec_IntSize(&p.vFinFon))
	}
	// add fons
	for i = 0; i < nFons; i++ {
		Vec_IntPush(&p.vFonObj, iObj)
	}
	if Vec_IntSize(&p.vObjFon0) != 0 {
		Vec_IntPush(&p.vObjFon0, Vec_IntSize(&p.vFonObj))
	}
	return iObj
}
func Cba_ObjDup(pNew *Cba_Ntk_t, p *Cba_Ntk_t, i int32) int32 {
	var iObj int32 = Cba_ObjAlloc(pNew, Cba_ObjType(p, i), Cba_ObjFinNum(p, i), Cba_ObjFonNum(p, i))
	Cba_ObjSetCopy(p, i, iObj)
	return iObj
}
func Cba_ObjDelete(p *Cba_Ntk_t, i int32) {
	var k, iFin, iFon int32
	Cba_ObjCleanType(p, i)
	// inlining of macro Cba_ObjForEachFin
	for k, iFin = 0, Cba_ObjFin0(p, i); iFin < Cba_ObjFin0(p, i+1); iFin, k = iFin+1, k+1 {
		Vec_IntWriteEntry(&p.vFinFon, iFin, 0)
	}
	// inlining of macro Cba_ObjForEachFon
	for k, iFon = 0, Cba_ObjFon0(p, i); iFon < Cba_ObjFon0(p, i+1); iFon, k = iFon+1, k+1 {
		Vec_IntWriteEntry(&p.vFonObj, iFon, 0)
	}
}

/*
*Function*************************************************************

	Synopsis    [Network APIs.]

	Description []

	SideEffects []

	SeeAlso     []

**********************************************************************
*/
func Cba_NtkAlloc(p *Cba_Man_t, NameId int32, nIns int32, nOuts int32, nObjs int32, nFins int32, nFons int32) *Cba_Ntk_t {
	// inlining of macro ABC_CALLOC
	var pNew *Cba_Ntk_t = &Cba_Ntk_t{}
	assert(nIns >= 0 && nOuts >= 0 && nObjs >= 0 && nFins >= 0 && nFons >= 0, "nIns >= 0 && nOuts >= 0 && nObjs >= 0 && nFins >= 0 && nFons >= 0")
	pNew.Id = Vec_PtrSize(&p.vNtks)
	Vec_PtrPush(&p.vNtks, pNew)
	pNew.NameId = NameId
	pNew.pDesign = p
	Vec_IntGrow(&pNew.vInputs, nIns)
	Vec_IntGrow(&pNew.vOutputs, nOuts)
	Vec_StrGrow(&pNew.vObjType, nObjs+1)
	Vec_StrPush(&pNew.vObjType, byte(CBA_OBJ_NONE))
	Vec_IntGrow(&pNew.vObjFin0, nObjs+2)
	Vec_IntPush(&pNew.vObjFin0, 0)
	Vec_IntPush(&pNew.vObjFin0, 1)
	Vec_IntGrow(&pNew.vObjFon0, nObjs+2)
	Vec_IntPush(&pNew.vObjFon0, 0)
	Vec_IntPush(&pNew.vObjFon0, 1)
	Vec_IntGrow(&pNew.vFinFon, nFins+1)
	Vec_IntPush(&pNew.vFinFon, 0)
	Vec_IntGrow(&pNew.vFonObj, nFons+1)
	Vec_IntPush(&pNew.vFonObj, 0)
	return pNew
}
func Cba_NtkFree(p *Cba_Ntk_t) {
	// interface
	Vec_IntErase(&p.vInputs)
	Vec_IntErase(&p.vOutputs)
	Vec_IntErase(&p.vOrder)
	Vec_IntErase(&p.vSeq)
	// stucture
	Vec_StrErase(&p.vObjType)
	Vec_IntErase(&p.vObjFin0)
	Vec_IntErase(&p.vObjFon0)
	Vec_IntErase(&p.vFinFon)
	Vec_IntErase(&p.vFonObj)
	// optional
	Vec_IntErase(&p.vObjCopy)
	Vec_IntErase(&p.vObjFunc)
	Vec_IntErase(&p.vObjName)
	Vec_IntErase(&p.vObjAttr)
	Vec_IntErase(&p.vAttrSto)
	Vec_IntErase(&p.vFonCopy)
	Vec_IntErase(&p.vFonName)
	Vec_IntErase(&p.vFonRange)
	Vec_IntErase(&p.vFonPrev)
	Vec_IntErase(&p.vFonNext)
	Vec_IntErase(&p.vFinFon0)
	Vec_IntErase(&p.vFinObj)
	Vec_IntErase(&p.vNtkObjs)
	Vec_IntErase(&p.vFonBits)
	// other
	Vec_IntErase(&p.vArray0)
	Vec_IntErase(&p.vArray1)
	if p.vOther != nil {
		Vec_VecFree(p.vOther)
	}
	p = nil
}
func Cba_NtkNewStrId(pNtk *Cba_Ntk_t, format string, args ...any) int32 {
	return Abc_NamStrFindOrAdd(Cba_NtkNam(pNtk), fmt.Sprintf(format, args...), nil)
}
func Cba_ManNewConstId(p *Cba_Ntk_t, vBits *Vec_Str_t) int32 {
	value := fmt.Sprintf("%d'b%s", Vec_StrSize(vBits)-1, string(Vec_StrArray(vBits)))
	return Abc_NamStrFindOrAdd(p.pDesign.pFuns, value, nil)
}
func Cba_ManNewConstZero(p *Cba_Ntk_t, nBits int32) int32 {
	value := fmt.Sprintf("%d'b", nBits)
	return Abc_NamStrFindOrAdd(p.pDesign.pFuns, value, nil)
}

func Cba_NtkAdd(p *Cba_Man_t, pNtk *Cba_Ntk_t) {
	var fFound int32
	var NtkId int32 = Abc_NamStrFindOrAdd(p.pMods, Cba_NtkName(pNtk), &fFound)
	if fFound != 0 {
		fmt.Printf("Network with name \"%s\" already exists.\n", Cba_NtkName(pNtk))
	} else {
		assert(NtkId == pNtk.Id, "NtkId == pNtk.Id")
	}
}
func Cba_NtkUpdate(p *Cba_Man_t, pNtk *Cba_Ntk_t) {
	var fFound int32
	var NtkId int32 = Abc_NamStrFindOrAdd(p.pMods, Cba_NtkName(pNtk), &fFound)
	if !(fFound != 0) {
		fmt.Printf("Network with name \"%s\" does not exist.\n", Cba_NtkName(pNtk))
	} else {
		Cba_NtkFree(Cba_ManNtk(p, NtkId))
		Vec_PtrWriteEntry(&p.vNtks, NtkId, pNtk)
	}
}
func Cba_NtkCollect(p *Cba_Ntk_t) *Vec_Int_t {
	var iObj int32
	var vObjs *Vec_Int_t = Vec_IntAlloc(Cba_NtkObjNum(p))
	// inlining of macro Cba_NtkForEachObj
	for iObj = 1; iObj < Vec_StrSize(&p.vObjType); iObj++ {
		if Cba_ObjType(p, iObj) == CBA_OBJ_NONE {
		} else {
			Vec_IntPush(vObjs, iObj)
		}
	}
	return vObjs
}
func Cba_NtkIsSeq(p *Cba_Ntk_t) int32 {
	var iObj int32
	if p.fSeq != 0 {
		return 1
	}
	if p.fComb != 0 {
		return 0
	}
	assert(p.fSeq == 0 && p.fComb == 0, "!p.fSeq && !p.fComb")
	// inlining of macro Cba_NtkForEachBox
	for iObj = 1; iObj < Vec_StrSize(&p.vObjType); iObj++ {
		if !(Cba_ObjIsBox(p, iObj) != 0) {
		} else if Cba_ObjIsBoxUser(p, iObj) != 0 {
			if Cba_NtkIsSeq(Cba_ObjNtk(p, iObj)) != 0 {
				p.fSeq = 1
				return 1
			}

		} else if Cba_ObjIsSeq(p, iObj) != 0 {
			p.fSeq = 1
			return 1
		}
	}
	p.fComb = 1
	return 0
}
func Cba_NtkPrepareSeq(p *Cba_Ntk_t) {
	var iObj int32
	assert(Cba_NtkBoxSeqNum(p) == 0, "Cba_NtkBoxSeqNum(p) == 0")
	if !(Cba_NtkIsSeq(p) != 0) {
		return
	}
	// inlining of macro Cba_NtkForEachBox
	for iObj = 1; iObj < Vec_StrSize(&p.vObjType); iObj++ {
		if !(Cba_ObjIsBox(p, iObj) != 0) {
		} else if Cba_ObjIsSeq(p, iObj) != 0 {
			Vec_IntPush(&p.vSeq, iObj)
		}
	}
	// Cba_NtkObjOrder( p, &p->vSeq, NULL );
}
func Cba_NtkCountParams(p *Cba_Ntk_t, vObjs *Vec_Int_t, nPis *int32, nPos *int32, nFins *int32, nFons *int32) {
	var i, iObj int32
	*nPis, *nPos, *nFins, *nFons = 0, 0, 0, 0
	// inlining of macro Vec_IntForEachEntry
	for i = 0; i < Vec_IntSize(vObjs); i++ {
		iObj = Vec_IntEntry(vObjs, i)
		*nPis += Cba_ObjIsPi(p, iObj)
		*nPos += Cba_ObjIsPo(p, iObj)
		*nFins += Cba_ObjFinNum(p, iObj)
		*nFons += Cba_ObjFonNum(p, iObj)
	}
}
func Cba_NtkDup(pMan *Cba_Man_t, p *Cba_Ntk_t, vObjs *Vec_Int_t) *Cba_Ntk_t {
	var pNew *Cba_Ntk_t
	var i, k, iObj, iObjNew, iFin, iFon int32
	var nPis, nPos, nFins, nFons int32
	Cba_NtkCountParams(p, vObjs, &nPis, &nPos, &nFins, &nFons)
	pNew = Cba_NtkAlloc(pMan, Cba_NtkNameId(p), nPis, nPos, Vec_IntSize(vObjs), nFins, nFons)
	Cba_NtkCleanObjCopies(p)
	Cba_NtkCleanFonCopies(p)
	// inlining of macro Vec_IntForEachEntry
	for i = 0; i < Vec_IntSize(vObjs); i++ {
		iObj = Vec_IntEntry(vObjs, i)
		iObjNew = Cba_ObjDup(pNew, p, iObj)
		// inlining of macro Cba_ObjForEachFon
		for k, iFon = 0, Cba_ObjFon0(p, iObj); iFon < Cba_ObjFon0(p, iObj+1); iFon, k = iFon+1, k+1 {
			Cba_FonSetCopy(p, iFon, Cba_ObjFon(pNew, iObjNew, k))
		}
	}
	// inlining of macro Vec_IntForEachEntry
	for i = 0; i < Vec_IntSize(vObjs); i++ {
		iObj = Vec_IntEntry(vObjs, i)
		iObjNew = Cba_ObjCopy(p, iObj)
		// inlining of macro Cba_ObjForEachFinFon
		for k, iFin = 0, Cba_ObjFin0(p, iObj); iFin < Cba_ObjFin0(p, iObj+1); iFin, k = iFin+1, k+1 {
			iFon = Cba_FinFon(p, iFin)
			Cba_ObjSetFinFon(pNew, iObjNew, k, Cba_FonCopy(p, iFon))
		}
	}
	//Cba_NtkFreeObjCopies( p );
	//Cba_NtkFreeFonCopies( p );
	assert(Cba_NtkObjNum(pNew) == Cba_NtkObjNumAlloc(pNew), "Cba_NtkObjNum(pNew) == Cba_NtkObjNumAlloc(pNew)")
	Cba_NtkSetCopy(p, Cba_NtkId(pNew))
	return pNew
}
func Cba_NtkDupOrder(pMan *Cba_Man_t, p *Cba_Ntk_t, pFuncOrder func(*Cba_Ntk_t) *Vec_Int_t) *Cba_Ntk_t {
	var pNew *Cba_Ntk_t
	var vObjs *Vec_Int_t
	if pFuncOrder != nil {
		vObjs = pFuncOrder(p)
	} else {
		vObjs = Cba_NtkCollect(p)
	}
	if vObjs == nil {
		return nil
	}
	pNew = Cba_NtkDup(pMan, p, vObjs)
	Vec_IntFree(vObjs)
	//Cba_NtkPrepareSeq( pNew );
	return pNew
}
func Cba_NtkDupAttrs(pNew *Cba_Ntk_t, p *Cba_Ntk_t) {
	var i, iObj int32
	assert(Vec_IntSize(&pNew.vOrder) == 0, "Vec_IntSize(&pNew.vOrder) == 0")
	// inlining of macro Cba_NtkForEachPioOrder
	for i = 0; i < Cba_NtkPioOrderNum(p); i++ {
		iObj = Cba_NtkPioOrder(p, i)
		Vec_IntPush(&pNew.vOrder, Cba_ObjCopy(p, iObj))
	}
	//    Vec_IntRemapArray( &p->vObjCopy, &p->vOrder,    &pNew->vOrder,    Cba_NtkPioOrderNum(p) );
	//    Vec_IntRemapArray( &p->vObjCopy, &p->vSeq,      &pNew->vSeq,      Cba_NtkBoxSeqNum(p) );
	// transfer object attributes
	Vec_IntRemapArray(&p.vObjCopy, &p.vObjFunc, &pNew.vObjFunc, Cba_NtkObjNum(pNew)+1)
	Vec_IntRemapArray(&p.vObjCopy, &p.vObjName, &pNew.vObjName, Cba_NtkObjNum(pNew)+1)
	Vec_IntRemapArray(&p.vObjCopy, &p.vObjAttr, &pNew.vObjAttr, Cba_NtkObjNum(pNew)+1)
	// transfer fon attributes
	Vec_IntRemapArray(&p.vFonCopy, &p.vFonName, &pNew.vFonName, Cba_NtkFonNum(pNew)+1)
	Vec_IntRemapArray(&p.vFonCopy, &p.vFonRange, &pNew.vFonRange, Cba_NtkFonNum(pNew)+1)
	// duplicate attributes
	Vec_IntAppend(&pNew.vAttrSto, &p.vAttrSto)
	if p.vOther != nil {
		pNew.vOther = Vec_VecDup(p.vOther)
	} else {
		pNew.vOther = nil
	}
}
func Cba_NtkMemory(p *Cba_Ntk_t) int32 {
	var nMem int32 = int32(unsafe.Sizeof(Cba_Ntk_t{}))
	// interface
	nMem += int32(Vec_IntMemory(&p.vInputs))
	nMem += int32(Vec_IntMemory(&p.vOutputs))
	nMem += int32(Vec_IntMemory(&p.vOrder))
	nMem += int32(Vec_IntMemory(&p.vSeq))
	// stucture
	nMem += int32(Vec_StrMemory(&p.vObjType))
	nMem += int32(Vec_IntMemory(&p.vObjFin0))
	nMem += int32(Vec_IntMemory(&p.vObjFon0))
	nMem += int32(Vec_IntMemory(&p.vFinFon))
	nMem += int32(Vec_IntMemory(&p.vFonObj))
	// optional
	nMem += int32(Vec_IntMemory(&p.vObjCopy))
	nMem += int32(Vec_IntMemory(&p.vObjFunc))
	nMem += int32(Vec_IntMemory(&p.vObjName))
	nMem += int32(Vec_IntMemory(&p.vObjAttr))
	nMem += int32(Vec_IntMemory(&p.vAttrSto))
	nMem += int32(Vec_IntMemory(&p.vFonCopy))
	nMem += int32(Vec_IntMemory(&p.vFonName))
	nMem += int32(Vec_IntMemory(&p.vFonRange))
	nMem += int32(Vec_IntMemory(&p.vFonPrev))
	nMem += int32(Vec_IntMemory(&p.vFonNext))
	nMem += int32(Vec_IntMemory(&p.vFinFon0))
	nMem += int32(Vec_IntMemory(&p.vFinObj))
	nMem += int32(Vec_IntMemory(&p.vNtkObjs))
	nMem += int32(Vec_IntMemory(&p.vFonBits))
	// other
	nMem += int32(Vec_IntMemory(&p.vArray1))
	nMem += int32(Vec_IntMemory(&p.vArray1))
	return nMem
}
func Cba_NtkPrintStats(p *Cba_Ntk_t) {
	fmt.Printf("pi =%5d  ", Cba_NtkPiNum(p))
	fmt.Printf("po =%5d  ", Cba_NtkPoNum(p))
	fmt.Printf("seq =%5d  ", Cba_NtkBoxSeqNum(p))
	fmt.Printf("box =%5d  ", Cba_NtkBoxNum(p))
	fmt.Printf("prim =%5d  ", Cba_NtkBoxPrimNum(p))
	fmt.Printf("user =%5d  ", Cba_NtkBoxUserNum(p))
	//printf( "topo =%4s  ", Cba_NtkIsTopoOrder(p) ? "yes" : "no" );
	fmt.Printf("  %s ", Cba_NtkName(p))
	if Vec_IntSize(&p.vNtkObjs) != 0 {
		fmt.Printf("-> %s", Cba_NtkName(Cba_NtkNtk(p, Vec_IntEntry(&p.vNtkObjs, 0))))
	}
	fmt.Printf("\n")
	// Vec_StrIntPrint( &p->vObjType );
}
func Cba_NtkPrint(p *Cba_Ntk_t) {
	var i, Type int32
	fmt.Printf("Interface (%d):\n", Cba_NtkPioNum(p))
	fmt.Printf("Objects (%d):\n", Cba_NtkObjNum(p))
	// inlining of macro Cba_NtkForEachObjType
	for i = 1; i < Vec_StrSize(&p.vObjType); i++ {
		Type = int32(Cba_ObjType(p, i))
		if Type != 0 {
			fmt.Printf("%6d : ", i)
			fmt.Printf("Type =%3d  ", Type)
			fmt.Printf("Fins = %d  ", Cba_ObjFinNum(p, i))
			fmt.Printf("Fons = %d  ", Cba_ObjFonNum(p, i))
			if (Cba_NtkHasObjNames(p) != 0) && (Cba_ObjName(p, i) != 0) {
				fmt.Printf("%s", Cba_ObjNameStr(p, i))
			}
			fmt.Printf("\n")
		}
	}
}

/*
*Function*************************************************************

	Synopsis    [Manager APIs.]

	Description []

	SideEffects []

	SeeAlso     []

**********************************************************************
*/
func Cba_ManAlloc(pFileName string, nNtks int32, pStrs *Abc_Nam_t, pFuns *Abc_Nam_t, pMods *Abc_Nam_t, vHash *Hash_IntMan_t) *Cba_Man_t {
	// inlining of macro ABC_CALLOC
	var pNew *Cba_Man_t = &Cba_Man_t{}
	pNew.pName = Extra_FileDesignName(pFileName)
	pNew.pSpec = Abc_UtilStrsav(pFileName)
	if pStrs != nil {
		pNew.pStrs = pStrs
	} else {
		pNew.pStrs = Abc_NamStart(1000, 24)
	}
	if pFuns != nil {
		pNew.pFuns = pFuns
	} else {
		pNew.pFuns = Abc_NamStart(100, 24)
	}
	if pMods != nil {
		pNew.pMods = pMods
	} else {
		pNew.pMods = Abc_NamStart(100, 24)
	}
	if vHash != nil {
		pNew.vHash = vHash
	} else {
		pNew.vHash = Hash_IntManStart(1000)
	}
	if pFuns == nil {
		Abc_NamStrFindOrAdd(pNew.pFuns, "1'b0", nil)
		Abc_NamStrFindOrAdd(pNew.pFuns, "1'b1", nil)
		Abc_NamStrFindOrAdd(pNew.pFuns, "1'bx", nil)
		Abc_NamStrFindOrAdd(pNew.pFuns, "1'bz", nil)
	}
	//    if ( vHash == NULL )
	//        Hash_Int2ManInsert( pNew->vHash, 0, 0, 0 );
	Vec_PtrGrow(&pNew.vNtks, nNtks+1)
	Vec_PtrPush(&pNew.vNtks, nil)
	// set default root module
	pNew.iRoot = 1
	return pNew
}
func Cba_ManDupTypeNames(pNew *Cba_Man_t, p *Cba_Man_t) {
	pNew.pTypeNames = p.pTypeNames
}
func Cba_ManDup(p *Cba_Man_t, pFuncOrder func(*Cba_Ntk_t) *Vec_Int_t) *Cba_Man_t {
	var pNtk, pNtkNew *Cba_Ntk_t
	var i int32
	var pNew *Cba_Man_t = Cba_ManAlloc(p.pSpec, Cba_ManNtkNum(p), Abc_NamRef(p.pStrs), Abc_NamRef(p.pFuns), Abc_NamStart(100, 24), Hash_IntManRef(p.vHash))
	Cba_ManDupTypeNames(pNew, p)
	// inlining of macro Cba_ManForEachNtk
	for i = 1; i <= Cba_ManNtkNum(p); i++ {
		pNtk = Cba_ManNtk(p, i)
		pNtkNew = Cba_NtkDupOrder(pNew, pNtk, pFuncOrder)
		Cba_NtkAdd(pNew, pNtkNew)
		Cba_NtkDupAttrs(pNtkNew, pNtk)
	}
	//    Cba_ManForEachNtk( p, pNtk, i )
	//        if ( (pHost = Cba_NtkHostNtk(pNtk)) )
	//            Cba_NtkSetHost( Cba_NtkCopyNtk(pNew, pNtk), Cba_NtkCopy(pHost), Cba_ObjCopy(pHost, Cba_NtkHostObj(pNtk)) );
	pNew.iRoot = Cba_ManNtkNum(pNew)
	return pNew
}
func Cba_ManPrepareSeq(p *Cba_Man_t) {
	var pNtk *Cba_Ntk_t
	var i int32
	// inlining of macro Cba_ManForEachNtk
	for i = 1; i <= Cba_ManNtkNum(p); i++ {
		pNtk = Cba_ManNtk(p, i)
		Cba_NtkPrepareSeq(pNtk)
	}
}
func Cba_ManFree(p *Cba_Man_t) {
	var pNtk *Cba_Ntk_t
	var i int32
	// inlining of macro Cba_ManForEachNtk
	for i = 1; i <= Cba_ManNtkNum(p); i++ {
		pNtk = Cba_ManNtk(p, i)
		Cba_NtkFree(pNtk)
	}
	Vec_PtrErase(&p.vNtks)
	Abc_NamDeref(p.pStrs)
	Abc_NamDeref(p.pFuns)
	Abc_NamDeref(p.pMods)
	Hash_IntManDeref(p.vHash)
	Vec_IntErase(&p.vNameMap)
	Vec_IntErase(&p.vUsed)
	Vec_IntErase(&p.vNameMap2)
	Vec_IntErase(&p.vUsed2)
	Vec_StrErase(&p.vOut)
	Vec_StrErase(&p.vOut2)
	p.pName = ""
	p.pSpec = ""
	p = nil
}
func Cba_ManGetMap(p *Cba_Man_t, i int32) int32 {
	return Vec_IntGetEntry(&p.vNameMap, i)
}
func Cba_ManSetMap(p *Cba_Man_t, i int32, x int32) {
	assert(Vec_IntGetEntry(&p.vNameMap, i) == 0, "Vec_IntGetEntry(&p.vNameMap, i) == 0")
	Vec_IntSetEntry(&p.vNameMap, i, x)
	Vec_IntPush(&p.vUsed, i)
}
func Cba_ManUnsetMap(p *Cba_Man_t, i int32) {
	Vec_IntSetEntry(&p.vNameMap, i, 0)
}
func Cba_ManCleanMap(p *Cba_Man_t) {
	var i, Entry int32
	// inlining of macro Vec_IntForEachEntry
	for i = 0; i < Vec_IntSize(&p.vUsed); i++ {
		Entry = Vec_IntEntry(&p.vUsed, i)
		Vec_IntWriteEntry(&p.vNameMap, Entry, 0)
	}
	Vec_IntClear(&p.vUsed)
}
func Cba_NtkGetMap(p *Cba_Ntk_t, i int32) int32    { return Cba_ManGetMap(p.pDesign, i) }
func Cba_NtkSetMap(p *Cba_Ntk_t, i int32, x int32) { Cba_ManSetMap(p.pDesign, i, x) }
func Cba_NtkUnsetMap(p *Cba_Ntk_t, i int32)        { Cba_ManUnsetMap(p.pDesign, i) }
func Cba_NtkCleanMap(p *Cba_Ntk_t)                 { Cba_ManCleanMap(p.pDesign) }

func Cba_ManGetMap2(p *Cba_Man_t, i int32) int32 {
	return Vec_IntGetEntry(&p.vNameMap2, i)
}
func Cba_ManSetMap2(p *Cba_Man_t, i int32, x int32) {
	assert(Vec_IntGetEntry(&p.vNameMap2, i) == 0, "Vec_IntGetEntry(&p.vNameMap2, i) == 0")
	Vec_IntSetEntry(&p.vNameMap2, i, x)
	Vec_IntPush(&p.vUsed2, i)
}
func Cba_ManUnsetMap2(p *Cba_Man_t, i int32) {
	Vec_IntSetEntry(&p.vNameMap2, i, 0)
}
func Cba_ManCleanMap2(p *Cba_Man_t) {
	var i, Entry int32
	// inlining of macro Vec_IntForEachEntry
	for i = 0; i < Vec_IntSize(&p.vUsed2); i++ {
		Entry = Vec_IntEntry(&p.vUsed2, i)
		Vec_IntWriteEntry(&p.vNameMap2, Entry, 0)
	}
	Vec_IntClear(&p.vUsed2)
}
func Cba_NtkGetMap2(p *Cba_Ntk_t, i int32) int32    { return Cba_ManGetMap2(p.pDesign, i) }
func Cba_NtkSetMap2(p *Cba_Ntk_t, i int32, x int32) { Cba_ManSetMap2(p.pDesign, i, x) }
func Cba_NtkUnsetMap2(p *Cba_Ntk_t, i int32)        { Cba_ManUnsetMap2(p.pDesign, i) }
func Cba_NtkCleanMap2(p *Cba_Ntk_t)                 { Cba_ManCleanMap2(p.pDesign) }

func Cba_ManMemory(p *Cba_Man_t) int32 {
	var pNtk *Cba_Ntk_t
	var i int32
	var nMem int32 = int32(unsafe.Sizeof(Cba_Man_t{}))
	nMem += int32(len(p.pName))
	nMem += int32(len(p.pSpec))
	nMem += Abc_NamMemUsed(p.pStrs)
	nMem += Abc_NamMemUsed(p.pFuns)
	nMem += Abc_NamMemUsed(p.pMods)
	nMem += int32(Vec_IntMemory(&p.vNameMap))
	nMem += int32(Vec_IntMemory(&p.vUsed))
	nMem += int32(Vec_StrMemory(&p.vOut))
	nMem += int32(Vec_StrMemory(&p.vOut2))
	nMem += int32(Vec_PtrMemory(&p.vNtks))
	// inlining of macro Cba_ManForEachNtk
	for i = 1; i <= Cba_ManNtkNum(p); i++ {
		pNtk = Cba_ManNtk(p, i)
		nMem += Cba_NtkMemory(pNtk)
	}
	return nMem
}
func Cba_ManObjNum(p *Cba_Man_t) int32 {
	var pNtk *Cba_Ntk_t
	var i, Count int32
	// inlining of macro Cba_ManForEachNtk
	for i = 1; i <= Cba_ManNtkNum(p); i++ {
		pNtk = Cba_ManNtk(p, i)
		Count += Cba_NtkObjNum(pNtk)
	}
	return Count
}
func Cba_ManBoxNum(p *Cba_Man_t) int32 {
	var pNtk *Cba_Ntk_t
	var i, Count int32
	// inlining of macro Cba_ManForEachNtk
	for i = 1; i <= Cba_ManNtkNum(p); i++ {
		pNtk = Cba_ManNtk(p, i)
		Count += Cba_NtkBoxNum(pNtk)
	}
	return Count
}
func Cba_ManBoxNumRec_rec(p *Cba_Ntk_t, pCountP []int32, pCountU []int32) {
	var iObj int32
	var Id int32 = Cba_NtkId(p)
	if pCountP[Id] >= 0 {
		return
	}
	pCountP[Id], pCountU[Id] = 0, 0
	// inlining of macro Cba_NtkForEachObj
	for iObj = 1; iObj < Vec_StrSize(&p.vObjType); iObj++ {
		if Cba_ObjType(p, iObj) == CBA_OBJ_NONE {
		} else {
			if Cba_ObjIsBoxUser(p, iObj) != 0 {
				Cba_ManBoxNumRec_rec(Cba_ObjNtk(p, iObj), pCountP, pCountU)
				pCountP[Id] += pCountP[Cba_ObjNtkId(p, iObj)]
				pCountU[Id] += pCountU[Cba_ObjNtkId(p, iObj)] + 1

			} else {
				pCountP[Id] += 1
			}
		}
	}
}
func Cba_ManBoxNumRec(p *Cba_Man_t, pnPrims *int32, pnUsers *int32) {
	var pNtk *Cba_Ntk_t = Cba_ManRoot(p)
	pCountP := make([]int32, Cba_ManNtkNum(p)+1)
	pCountU := make([]int32, Cba_ManNtkNum(p)+1)
	for i := range pCountP {
		pCountP[i], pCountU[i] = -1, -1
	}
	Cba_ManBoxNumRec_rec(pNtk, pCountP, pCountU)
	*pnPrims = pCountP[Cba_NtkId(pNtk)]
	*pnUsers = pCountU[Cba_NtkId(pNtk)]
}
func Cba_ManPrintStats(p *Cba_Man_t, nModules int32, fVerbose int32) {
	var pNtk *Cba_Ntk_t
	var i, nPrims, nUsers int32
	var pRoot *Cba_Ntk_t = Cba_ManRoot(p)
	Cba_ManBoxNumRec(p, &nPrims, &nUsers)
	fmt.Printf("%-12s : ", Cba_ManName(p))
	fmt.Printf("pi =%5d  ", Cba_NtkPiNum(pRoot))
	fmt.Printf("po =%5d  ", Cba_NtkPoNum(pRoot))
	fmt.Printf("mod =%5d  ", Cba_ManNtkNum(p))
	fmt.Printf("box =%5d  ", nPrims+nUsers)
	fmt.Printf("prim =%5d  ", nPrims)
	fmt.Printf("user =%5d  ", nUsers)
	fmt.Printf("mem =%6.3f MB", float64(Cba_ManMemory(p))/float64(1<<20))
	fmt.Printf("\n")
	// inlining of macro Cba_ManForEachNtk
	for i = 1; i <= Cba_ManNtkNum(p); i++ {
		pNtk = Cba_ManNtk(p, i)
		if i == nModules+1 {
			break
		}
		fmt.Printf("Module %5d : ", i)
		Cba_NtkPrintStats(pNtk)
	}
}

/*
*Function*************************************************************

	Synopsis    [Name handling.]

	Description []

	SideEffects []

	SeeAlso     []

**********************************************************************
*/
func Cba_NtkAddMissingFonNames(p *Cba_Ntk_t, pPref string) {
	var iFon, NameId, Index int32
	// populate map
	Cba_ManCleanMap(p.pDesign)
	// inlining of macro Vec_IntForEachEntryStart
	for iFon = 1; iFon < Vec_IntSize(&p.vFonName); iFon++ {
		NameId = Vec_IntEntry(&p.vFonName, iFon)
		if NameId != 0 {
			Cba_ManSetMap(p.pDesign, NameId, iFon)
		}
	}
	// check remaining ones
	// inlining of macro Vec_IntForEachEntryStart
	for iFon = 1; iFon < Vec_IntSize(&p.vFonName); iFon++ {
		NameId = Vec_IntEntry(&p.vFonName, iFon)
		if NameId != 0 {
			continue
		}
		NameId = Cba_NtkNewStrId(p, "%s%d", pPref, iFon)
		for Index = 1; Cba_ManGetMap(p.pDesign, NameId) != 0; Index++ {
			NameId = Cba_NtkNewStrId(p, "%s%d_%d", pPref, iFon, Index)
		}
		Cba_FonSetName(p, iFon, NameId)
		Cba_ManSetMap(p.pDesign, NameId, iFon)
	}
}
func Cba_NtkCreateFonNames(p *Cba_Ntk_t, pPref string) {
	var i, iObj, iFon int32 //, NameId;
	Cba_NtkCleanFonNames(p)
	// inlining of macro Cba_NtkForEachPiFon
	for i = 0; i < Cba_NtkPiNum(p); i++ {
		iObj = Cba_NtkPi(p, i)
		iFon = Cba_ObjFon0(p, iObj)
		if !(Cba_FonName(p, iFon) != 0) {
			Cba_FonSetName(p, iFon, Cba_ObjName(p, iObj))
		}
	}
	// inlining of macro Cba_NtkForEachPoDriverFon
	for i = 0; i < Cba_NtkPoNum(p); i++ {
		iObj = Cba_NtkPo(p, i)
		iFon = Cba_ObjFinFon(p, iObj, 0)
		if (Cba_FonIsReal(iFon) != 0) && !(Cba_FonName(p, iFon) != 0) {
			Cba_FonSetName(p, iFon, Cba_ObjName(p, iObj))
		}
	}
	//    Vec_IntForEachEntryStart( &p->vFonName, NameId, iFon, 1 )
	//        if ( NameId == 0 )
	//            Vec_IntWriteEntry( &p->vFonName, iFon, Cba_NtkNewStrId(p, "%s%d", pPref, iFon) );
	Cba_NtkAddMissingFonNames(p, pPref)
}
func Cba_NtkMissingFonNames(p *Cba_Ntk_t, pPref string) {
	var i, iObj, iFon int32 //, NameId;
	// inlining of macro Cba_NtkForEachPiFon
	for i = 0; i < Cba_NtkPiNum(p); i++ {
		iObj = Cba_NtkPi(p, i)
		iFon = Cba_ObjFon0(p, iObj)
		if !(Cba_FonName(p, iFon) != 0) {
			Cba_FonSetName(p, iFon, Cba_ObjName(p, iObj))
		}
	}
	// inlining of macro Cba_NtkForEachPoDriverFon
	for i = 0; i < Cba_NtkPoNum(p); i++ {
		iObj = Cba_NtkPo(p, i)
		iFon = Cba_ObjFinFon(p, iObj, 0)
		if (Cba_FonIsReal(iFon) != 0) && !(Cba_FonName(p, iFon) != 0) {
			Cba_FonSetName(p, iFon, Cba_ObjName(p, iObj))
		}
	}
	//    Vec_IntForEachEntryStart( &p->vFonName, NameId, iFon, 1 )
	//        if ( NameId == 0 )
	//            Cba_FonSetName( p, iFon, Cba_NtkNewStrId(p, "%s%d", pPref, iFon) );
	Cba_NtkAddMissingFonNames(p, pPref)
}

/*=== cbaBlast.c =============================================================*/
// extern Gia_Man_t *   Cba_ManBlast( Cba_Man_t * p, int fBarBufs, int fSeq, int fVerbose );

// extern Cba_Man_t *   Cba_ManInsertGia( Cba_Man_t * p, Gia_Man_t * pGia );

// extern Cba_Man_t *   Cba_ManInsertAbc( Cba_Man_t * p, void * pAbc );

/*=== cbaCba.c ===============================================================*/
// extern Cba_Man_t *   Cba_ManReadCba( char * pFileName );

// extern void          Cba_ManWriteCba( char * pFileName, Cba_Man_t * p );

/*=== cbaCom.c ===============================================================*/
/*=== cbaNtk.c ===============================================================*/
// extern void          Cba_NtkPrintStatsFull( Cba_Ntk_t * p, int fDistrib, int fVerbose );

// extern void          Cba_NtkPrintNodes( Cba_Ntk_t * p, int Type );

// extern void          Cba_NtkPrintDistribOld( Cba_Ntk_t * p );

// extern void          Cba_ManPrintDistrib( Cba_Man_t * p );

//extern void           Cba_ManPrepareTypeNames( Cba_Man_t * p );
// extern void          Cba_NtkObjOrder( Cba_Ntk_t * p, Vec_Int_t * vObjs, Vec_Int_t * vNameIds );

// extern int           Cba_NtkCiFonNum( Cba_Ntk_t * p );

// extern int           Cba_NtkCoFinNum( Cba_Ntk_t * p );

// extern int           Cba_NtkCheckComboLoop( Cba_Ntk_t * p );

// extern int           Cba_ManIsTopoOrder( Cba_Man_t * p );

// extern Vec_Int_t *   Cba_NtkCollectDfs( Cba_Ntk_t * p );

// extern Cba_Man_t *   Cba_ManCollapse( Cba_Man_t * p );

// extern Cba_Man_t *   Cba_ManExtractGroup( Cba_Man_t * p, Vec_Int_t * vObjs );

// extern Cba_Man_t *   Cba_ManDeriveFromGia( Cba_Man_t * p, Gia_Man_t * pGia, int fUseXor );

// extern Cba_Man_t *   Cba_ManInsertGroup( Cba_Man_t * p, Vec_Int_t * vObjs, Cba_Ntk_t * pSyn );

/*=== cbaReadBlif.c ==========================================================*/
// extern Cba_Man_t *   Prs_ManBuildCbaBlif( char * pFileName, Vec_Ptr_t * vDes );

// extern void          Prs_ManReadBlifTest( char * pFileName );

// extern Cba_Man_t *   Cba_ManReadBlif( char * pFileName );

/*=== cbaReadVer.c ===========================================================*/
// extern Cba_Man_t *   Prs_ManBuildCbaVerilog( char * pFileName, Vec_Ptr_t * vDes );

// extern void          Prs_ManReadVerilogTest( char * pFileName );

// extern Cba_Man_t *   Cba_ManReadVerilog( char * pFileName );

/*=== cbaWriteBlif.c =========================================================*/
// extern void          Prs_ManWriteBlif( char * pFileName, Vec_Ptr_t * p );

// extern void          Cba_ManWriteBlif( char * pFileName, Cba_Man_t * p );

/*=== cbaWriteVer.c ==========================================================*/
// extern void          Cba_ManCreatePrimMap( char ** pMap );

// extern char *        Cba_ManGetSliceName( Cba_Ntk_t * p, int iFon, int RangeId );

// extern void          Prs_ManWriteVerilog( char * pFileName, Vec_Ptr_t * p );

// extern void          Cba_ManWriteVerilog( char * pFileName, Cba_Man_t * p, int fInlineConcat );

//ABC_NAMESPACE_HEADER_END

////////////////////////////////////////////////////////////////////////
///                       END OF FILE                                ///
////////////////////////////////////////////////////////////////////////
