package acb

import (
	"fmt"
	"unsafe"
)

/**CFile****************************************************************

  FileName    [acb.h]

  SystemName  [ABC: Logic synthesis and verification system.]

  PackageName [Hierarchical word-level netlist.]

  Synopsis    [External declarations.]

  Author      [Alan Mishchenko]

  Affiliation [UC Berkeley]

  Date        [Ver. 1.0. Started - July 21, 2015.]

  Revision    [$Id: acb.h,v 1.00 2014/11/29 00:00:00 alanmi Exp $]

***********************************************************************/

////////////////////////////////////////////////////////////////////////
///                          INCLUDES                                ///
////////////////////////////////////////////////////////////////////////

//#include "aig/gia/gia.h"
//#include "misc/extra/extra.h"
//#include "misc/util/utilNam.h"
//#include "misc/vec/vecHash.h"
//#include "aig/miniaig/abcOper.h"
//#include "misc/vec/vecQue.h"

////////////////////////////////////////////////////////////////////////
///                         PARAMETERS                               ///
////////////////////////////////////////////////////////////////////////

//ABC_NAMESPACE_HEADER_START

////////////////////////////////////////////////////////////////////////
///                         BASIC TYPES                              ///
////////////////////////////////////////////////////////////////////////

type Acb_Aig_t = Acb_Aig_t_
type Acb_Ntk_t = Acb_Ntk_t_
type Acb_Man_t = Acb_Man_t_

// network
type Acb_Ntk_t_ struct {
	pDesign *Acb_Man_t // design
	Id      int32      // network ID
	NameId  int32      // name ID
	iCopy   int32      // copy module
	Mark    uint32     /* C bitfield: 1 bits; review packing and truncation */ // visit mark
	fComb   uint32     /* C bitfield: 1 bits; review packing and truncation */ // the network is combinational
	fSeq    uint32     /* C bitfield: 1 bits; review packing and truncation */ // the network is sequential
	// interface
	vCis      Vec_Int_t // inputs
	vCos      Vec_Int_t // outputs
	vOrder    Vec_Int_t // order
	vSeq      Vec_Int_t // sequential objects
	nRegs     int32     // flop count
	nFaninMax int32     // default fanin count
	nObjTravs int32     // trav ID
	LevelMax  int32     // max level
	nPaths    int32     // the number of paths
	// stucture
	vObjType Vec_Str_t // type
	vObjFans Vec_Int_t // fanin offsets
	vFanSto  Vec_Int_t // fanin storage
	// optional
	vObjCopy   Vec_Int_t   // copy
	vObjFunc   Vec_Int_t   // function
	vObjWeight Vec_Int_t   // weight
	vObjTruth  Vec_Wrd_t   // function
	vObjName   Vec_Int_t   // name
	vObjRange  Vec_Int_t   // range
	vObjTrav   Vec_Int_t   // trav ID
	vObjBits   Vec_Int_t   // obj mapping into AIG nodes
	vObjAttr   Vec_Int_t   // attribute offset
	vAttrSto   Vec_Int_t   // attribute storage
	vNtkObjs   Vec_Int_t   // instances
	vTargets   Vec_Int_t   // targets
	vLevelD    Vec_Int_t   // level
	vLevelR    Vec_Int_t   // level
	vPathD     Vec_Int_t   // path
	vPathR     Vec_Int_t   // path
	vCounts    Vec_Flt_t   // priority counts
	vFanouts   Vec_Wec_t   // fanouts
	vCnfs      []Vec_Str_t // CNF, one byte vector per object
	vCnf       Vec_Str_t   // CNF
	vSuppOld   Vec_Int_t   // previous support
	// other
	vQue    *Vec_Que_t // temporary
	vCover  Vec_Int_t  // temporary
	vArray0 Vec_Int_t  // temporary
	vArray1 Vec_Int_t  // temporary
	vArray2 Vec_Int_t  // temporary
}

// design
type Acb_Man_t_ struct {
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
	pTypeNames [ABC_OPER_LAST]string
	nObjs      [ABC_OPER_LAST]int32 // counter of objects of each type
	nAnds      [ABC_OPER_LAST]int32 // counter of AND gates after blasting
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

func Acb_ManName(p *Acb_Man_t) string  { return p.pName }
func Acb_ManSpec(p *Acb_Man_t) string  { return p.pSpec }
func Acb_ManNtkNum(p *Acb_Man_t) int32 { return Vec_PtrSize(&p.vNtks) - 1 }
func Acb_ManNtkIsOk(p *Acb_Man_t, i int32) int32 {
	return boolInt(i > 0 && i <= Acb_ManNtkNum(p), 1, 0)
}
func Acb_ManNtk(p *Acb_Man_t, i int32) *Acb_Ntk_t {
	if Acb_ManNtkIsOk(p, i) != 0 {
		return Vec_PtrEntry(&p.vNtks, i).(*Acb_Ntk_t)
	} else {
		return nil
	}
}
func Acb_ManNtkFindId(p *Acb_Man_t, pName string) int32 { return Abc_NamStrFind(p.pMods, pName) }
func Acb_ManNtkFind(p *Acb_Man_t, pName string) *Acb_Ntk_t {
	return Acb_ManNtk(p, Acb_ManNtkFindId(p, pName))
}
func Acb_ManRoot(p *Acb_Man_t) *Acb_Ntk_t            { return Acb_ManNtk(p, p.iRoot) }
func Acb_ManStr(p *Acb_Man_t, i int32) string        { return Abc_NamStr(p.pStrs, i) }
func Acb_ManStrId(p *Acb_Man_t, pStr string) int32   { return Abc_NamStrFind(p.pStrs, pStr) }
func Acb_ManNameIdMax(p *Acb_Man_t) int32            { return Abc_NamObjNumMax(p.pStrs) + 1 }
func Acb_ManConst(p *Acb_Man_t, i int32) string      { return Abc_NamStr(p.pFuns, i) }
func Acb_ManConstId(p *Acb_Man_t, pStr string) int32 { return Abc_NamStrFind(p.pFuns, pStr) }
func Acb_ManConstIdMax(p *Acb_Man_t) int32           { return Abc_NamObjNumMax(p.pFuns) + 1 }

func Acb_NtkMan(p *Acb_Ntk_t) *Acb_Man_t              { return p.pDesign }
func Acb_NtkNtk(p *Acb_Ntk_t, i int32) *Acb_Ntk_t     { return Acb_ManNtk(p.pDesign, i) }
func Acb_NtkId(p *Acb_Ntk_t) int32                    { return p.Id }
func Acb_NtkCi(p *Acb_Ntk_t, i int32) int32           { return Vec_IntEntry(&p.vCis, i) }
func Acb_NtkCo(p *Acb_Ntk_t, i int32) int32           { return Vec_IntEntry(&p.vCos, i) }
func Acb_NtkCioOrder(p *Acb_Ntk_t, i int32) int32     { return Vec_IntEntry(&p.vOrder, i) }
func Acb_NtkBoxSeq(p *Acb_Ntk_t, i int32) int32       { return Vec_IntEntry(&p.vSeq, i) }
func Acb_NtkNam(p *Acb_Ntk_t) *Abc_Nam_t              { return p.pDesign.pStrs }
func Acb_NtkStr(p *Acb_Ntk_t, i int32) string         { return Acb_ManStr(p.pDesign, i) }
func Acb_NtkStrId(p *Acb_Ntk_t, pName string) int32   { return Acb_ManStrId(p.pDesign, pName) }
func Acb_NtkConst(p *Acb_Ntk_t, i int32) string       { return Acb_ManConst(p.pDesign, i) }
func Acb_NtkSop(p *Acb_Ntk_t, i int32) string         { return Acb_ManConst(p.pDesign, i) }
func Acb_NtkNameId(p *Acb_Ntk_t) int32                { return p.NameId }
func Acb_NtkName(p *Acb_Ntk_t) string                 { return Acb_NtkStr(p, p.NameId) }
func Acb_NtkTypeName(p *Acb_Ntk_t, Type int32) string { return p.pDesign.pTypeNames[Type] }
func Acb_NtkCopy(p *Acb_Ntk_t) int32                  { return p.iCopy }
func Acb_NtkCopyNtk(pNew *Acb_Man_t, p *Acb_Ntk_t) *Acb_Ntk_t {
	return Acb_ManNtk(pNew, Acb_NtkCopy(p))
}
func Acb_NtkSetCopy(p *Acb_Ntk_t, i int32) { assert(p.iCopy == 0, "p.iCopy == 0"); p.iCopy = i }
func Acb_NtkHashRange(p *Acb_Ntk_t, l int32, r int32) int32 {
	return Hash_Int2ManInsert(p.pDesign.vHash, l, r, 0)
}
func Acb_NtkRangeLeft(p *Acb_Ntk_t, h int32) int32 {
	if h != 0 {
		return Hash_IntObjData0(p.pDesign.vHash, h)
	} else {
		return 0
	}
}
func Acb_NtkRangeRight(p *Acb_Ntk_t, h int32) int32 {
	if h != 0 {
		return Hash_IntObjData1(p.pDesign.vHash, h)
	} else {
		return 0
	}
}
func Acb_NtkRangeSize(p *Acb_Ntk_t, h int32) int32 {
	var l int32 = Acb_NtkRangeLeft(p, h)
	var r int32 = Acb_NtkRangeRight(p, h)
	if l > r {
		return 1 + l - r
	}
	return 1 + r - l
}

func Acb_NtkCiNum(p *Acb_Ntk_t) int32            { return Vec_IntSize(&p.vCis) }
func Acb_NtkCoNum(p *Acb_Ntk_t) int32            { return Vec_IntSize(&p.vCos) }
func Acb_NtkCioNum(p *Acb_Ntk_t) int32           { return Acb_NtkCiNum(p) + Acb_NtkCoNum(p) }
func Acb_NtkCiNumAlloc(p *Acb_Ntk_t) int32       { return Vec_IntCap(&p.vCis) }
func Acb_NtkCoNumAlloc(p *Acb_Ntk_t) int32       { return Vec_IntCap(&p.vCos) }
func Acb_NtkRegNum(p *Acb_Ntk_t) int32           { return p.nRegs }
func Acb_NtkSetRegNum(p *Acb_Ntk_t, nRegs int32) { p.nRegs = nRegs }
func Acb_NtkPiNum(p *Acb_Ntk_t) int32            { return Acb_NtkCiNum(p) - Acb_NtkRegNum(p) }
func Acb_NtkPoNum(p *Acb_Ntk_t) int32            { return Acb_NtkCoNum(p) - Acb_NtkRegNum(p) }
func Acb_NtkCioOrderNum(p *Acb_Ntk_t) int32      { return Vec_IntSize(&p.vOrder) }
func Acb_NtkObjNumAlloc(p *Acb_Ntk_t) int32      { return Vec_StrCap(&p.vObjType) - 1 }
func Acb_NtkObjNum(p *Acb_Ntk_t) int32           { return Vec_StrSize(&p.vObjType) - 1 }
func Acb_NtkObjNumMax(p *Acb_Ntk_t) int32        { return Vec_StrSize(&p.vObjType) }
func Acb_NtkTypeNum(p *Acb_Ntk_t, Type int32) int32 {
	return Vec_StrCountEntry(&p.vObjType, byte(Type))
}
func Acb_NtkBoxNum(p *Acb_Ntk_t) int32  { return Acb_NtkTypeNum(p, ABC_OPER_BOX) }
func Acb_NtkNodeNum(p *Acb_Ntk_t) int32 { return Vec_StrCountLarger(&p.vObjType, byte(ABC_OPER_BOX)) }
func Acb_NtkSeqNum(p *Acb_Ntk_t) int32  { return Vec_IntSize(&p.vSeq) }

// Per-object slices track object count; spare allocation capacity is not an object.
func Acb_NtkCleanObjCopies(p *Acb_Ntk_t)  { Vec_IntFill(&p.vObjCopy, Vec_StrSize(&p.vObjType), -1) }
func Acb_NtkCleanObjFuncs(p *Acb_Ntk_t)   { Vec_IntFill(&p.vObjFunc, Vec_StrSize(&p.vObjType), -1) }
func Acb_NtkCleanObjWeights(p *Acb_Ntk_t) { Vec_IntFill(&p.vObjWeight, Vec_StrSize(&p.vObjType), 0) }
func Acb_NtkCleanObjTruths(p *Acb_Ntk_t)  { Vec_WrdFill(&p.vObjTruth, Vec_StrSize(&p.vObjType), 0) }
func Acb_NtkCleanObjNames(p *Acb_Ntk_t)   { Vec_IntFill(&p.vObjName, Vec_StrSize(&p.vObjType), 0) }
func Acb_NtkCleanObjRanges(p *Acb_Ntk_t)  { Vec_IntFill(&p.vObjRange, Vec_StrSize(&p.vObjType), 0) }
func Acb_NtkCleanObjTravs(p *Acb_Ntk_t)   { Vec_IntFill(&p.vObjTrav, Vec_StrSize(&p.vObjType), 0) }
func Acb_NtkCleanObjAttrs(p *Acb_Ntk_t) {
	Vec_IntFill(&p.vObjAttr, Vec_StrSize(&p.vObjType), 0)
	Vec_IntFill(&p.vAttrSto, 1, -1)
}
func Acb_NtkCleanObjLevelD(p *Acb_Ntk_t) { Vec_IntFill(&p.vLevelD, Vec_StrSize(&p.vObjType), 0) }
func Acb_NtkCleanObjLevelR(p *Acb_Ntk_t) { Vec_IntFill(&p.vLevelR, Vec_StrSize(&p.vObjType), 0) }
func Acb_NtkCleanObjPathD(p *Acb_Ntk_t)  { Vec_IntFill(&p.vPathD, Vec_StrSize(&p.vObjType), 0) }
func Acb_NtkCleanObjPathR(p *Acb_Ntk_t)  { Vec_IntFill(&p.vPathR, Vec_StrSize(&p.vObjType), 0) }
func Acb_NtkCleanObjCounts(p *Acb_Ntk_t) { Vec_FltFill(&p.vCounts, Vec_StrSize(&p.vObjType), 0) }
func Acb_NtkCleanObjFanout(p *Acb_Ntk_t) { Vec_WecInit(&p.vFanouts, Vec_StrSize(&p.vObjType)) }
func Acb_NtkCleanObjCnfs(p *Acb_Ntk_t)   { acbResizeCnfs(p, int(Vec_StrSize(&p.vObjType))) }

func Acb_NtkHasObjCopies(p *Acb_Ntk_t) int32  { return boolInt(Vec_IntSize(&p.vObjCopy) > 0, 1, 0) }
func Acb_NtkHasObjFuncs(p *Acb_Ntk_t) int32   { return boolInt(Vec_IntSize(&p.vObjFunc) > 0, 1, 0) }
func Acb_NtkHasObjWeights(p *Acb_Ntk_t) int32 { return boolInt(Vec_IntSize(&p.vObjWeight) > 0, 1, 0) }
func Acb_NtkHasObjTruths(p *Acb_Ntk_t) int32  { return boolInt(Vec_WrdSize(&p.vObjTruth) > 0, 1, 0) }
func Acb_NtkHasObjNames(p *Acb_Ntk_t) int32   { return boolInt(Vec_IntSize(&p.vObjName) > 0, 1, 0) }
func Acb_NtkHasObjRanges(p *Acb_Ntk_t) int32  { return boolInt(Vec_IntSize(&p.vObjRange) > 0, 1, 0) }
func Acb_NtkHasObjTravs(p *Acb_Ntk_t) int32   { return boolInt(Vec_IntSize(&p.vObjTrav) > 0, 1, 0) }
func Acb_NtkHasObjAttrs(p *Acb_Ntk_t) int32   { return boolInt(Vec_IntSize(&p.vObjAttr) > 0, 1, 0) }
func Acb_NtkHasObjLevelD(p *Acb_Ntk_t) int32  { return boolInt(Vec_IntSize(&p.vLevelD) > 0, 1, 0) }
func Acb_NtkHasObjLevelR(p *Acb_Ntk_t) int32  { return boolInt(Vec_IntSize(&p.vLevelR) > 0, 1, 0) }
func Acb_NtkHasObjPathD(p *Acb_Ntk_t) int32   { return boolInt(Vec_IntSize(&p.vPathD) > 0, 1, 0) }
func Acb_NtkHasObjPathR(p *Acb_Ntk_t) int32   { return boolInt(Vec_IntSize(&p.vPathR) > 0, 1, 0) }
func Acb_NtkHasObjCounts(p *Acb_Ntk_t) int32  { return boolInt(Vec_FltSize(&p.vCounts) > 0, 1, 0) }
func Acb_NtkHasObjFanout(p *Acb_Ntk_t) int32  { return boolInt(Vec_WecSize(&p.vFanouts) > 0, 1, 0) }
func Acb_NtkHasObjCnfs(p *Acb_Ntk_t) int32    { return boolInt(len(p.vCnfs) > 0, 1, 0) }

func Acb_NtkFreeObjCopies(p *Acb_Ntk_t)  { Vec_IntErase(&p.vObjCopy) }
func Acb_NtkFreeObjFuncs(p *Acb_Ntk_t)   { Vec_IntErase(&p.vObjFunc) }
func Acb_NtkFreeObjWeights(p *Acb_Ntk_t) { Vec_IntErase(&p.vObjWeight) }
func Acb_NtkFreeObjTruths(p *Acb_Ntk_t)  { Vec_WrdErase(&p.vObjTruth) }
func Acb_NtkFreeObjNames(p *Acb_Ntk_t)   { Vec_IntErase(&p.vObjName) }
func Acb_NtkFreeObjRanges(p *Acb_Ntk_t)  { Vec_IntErase(&p.vObjRange) }
func Acb_NtkFreeObjTravs(p *Acb_Ntk_t)   { Vec_IntErase(&p.vObjTrav) }
func Acb_NtkFreeObjAttrs(p *Acb_Ntk_t)   { Vec_IntErase(&p.vObjAttr) }
func Acb_NtkFreeObjLevelD(p *Acb_Ntk_t)  { Vec_IntErase(&p.vLevelD) }
func Acb_NtkFreeObjLevelR(p *Acb_Ntk_t)  { Vec_IntErase(&p.vLevelR) }
func Acb_NtkFreeObjPathD(p *Acb_Ntk_t)   { Vec_IntErase(&p.vPathD) }
func Acb_NtkFreeObjPathR(p *Acb_Ntk_t)   { Vec_IntErase(&p.vPathR) }
func Acb_NtkFreeObjCounts(p *Acb_Ntk_t)  { Vec_FltErase(&p.vCounts) }
func Acb_NtkFreeObjFanout(p *Acb_Ntk_t)  { Vec_WecErase(&p.vFanouts) }
func Acb_NtkFreeObjCnfs(p *Acb_Ntk_t)    { p.vCnfs = nil }

func Acb_ObjType(p *Acb_Ntk_t, i int32) Acb_ObjType_t {
	assert(i > 0, "i>0")
	return Acb_ObjType_t(Vec_StrEntry(&p.vObjType, i))
}
func Acb_ObjCleanType(p *Acb_Ntk_t, i int32) {
	assert(i > 0, "i>0")
	Vec_StrWriteEntry(&p.vObjType, i, byte(ABC_OPER_NONE))
}
func Acb_TypeIsSeq(Type Acb_ObjType_t) int32 {
	return boolInt(Type >= ABC_OPER_RAM && Type <= ABC_OPER_DFFLAST, 1, 0)
}
func Acb_TypeIsUnary(Type Acb_ObjType_t) int32 {
	return boolInt(Type == ABC_OPER_BIT_BUF || Type == ABC_OPER_BIT_INV || Type == ABC_OPER_LOGIC_NOT || Type == ABC_OPER_ARI_MIN || Type == ABC_OPER_ARI_SQRT || Type == ABC_OPER_ARI_ABS || (Type >= ABC_OPER_RED_AND && Type <= ABC_OPER_RED_NXOR), 1, 0)
}
func Acb_TypeIsMux(Type Acb_ObjType_t) int32 {
	return boolInt(Type == ABC_OPER_BIT_MUX || Type == ABC_OPER_SEL_NMUX || Type == ABC_OPER_SEL_SEL || Type == ABC_OPER_SEL_PSEL, 1, 0)
}

func Acb_ObjIsCi(p *Acb_Ntk_t, i int32) int32 { return boolInt(Acb_ObjType(p, i) == ABC_OPER_CI, 1, 0) }
func Acb_ObjIsCo(p *Acb_Ntk_t, i int32) int32 { return boolInt(Acb_ObjType(p, i) == ABC_OPER_CO, 1, 0) }
func Acb_ObjIsCio(p *Acb_Ntk_t, i int32) int32 {
	return boolInt((Acb_ObjIsCi(p, i) != 0 || Acb_ObjIsCo(p, i) != 0), 1, 0)
}
func Acb_ObjIsFon(p *Acb_Ntk_t, i int32) int32 {
	return boolInt(Acb_ObjType(p, i) == ABC_OPER_FON, 1, 0)
}
func Acb_ObjIsBox(p *Acb_Ntk_t, i int32) int32 {
	return boolInt(Acb_ObjType(p, i) == ABC_OPER_BOX, 1, 0)
}
func Acb_ObjIsGate(p *Acb_Ntk_t, i int32) int32 {
	return boolInt(Acb_ObjType(p, i) == ABC_OPER_GATE, 1, 0)
}
func Acb_ObjIsSlice(p *Acb_Ntk_t, i int32) int32 {
	return boolInt(Acb_ObjType(p, i) == ABC_OPER_SLICE, 1, 0)
}
func Acb_ObjIsConcat(p *Acb_Ntk_t, i int32) int32 {
	return boolInt(Acb_ObjType(p, i) == ABC_OPER_CONCAT, 1, 0)
}
func Acb_ObjIsUnary(p *Acb_Ntk_t, i int32) int32 { return Acb_TypeIsUnary(Acb_ObjType(p, i)) }

func Acb_ObjFanOffset(p *Acb_Ntk_t, i int32) int32 {
	assert(i > 0, "i>0")
	return Vec_IntEntry(&p.vObjFans, i)
}
func Acb_ObjFanins(p *Acb_Ntk_t, i int32) []int32 {
	return Vec_IntArray(&p.vFanSto)[Acb_ObjFanOffset(p, i):]
}
func Acb_ObjFanin(p *Acb_Ntk_t, i int32, k int32) int32 { return Acb_ObjFanins(p, i)[k+1] }
func Acb_ObjFaninNum(p *Acb_Ntk_t, i int32) int32       { return Acb_ObjFanins(p, i)[0] }
func Acb_ObjFanoutNum(p *Acb_Ntk_t, i int32) int32      { return Vec_IntSize(Vec_WecEntry(&p.vFanouts, i)) }
func Acb_ObjFanoutVec(p *Acb_Ntk_t, i int32) *Vec_Int_t {
	assert(i > 0, "i>0")
	return Vec_WecEntry(&p.vFanouts, i)
}
func Acb_ObjFanout(p *Acb_Ntk_t, i int32, k int32) int32 {
	return Vec_IntEntry(Acb_ObjFanoutVec(p, i), k)
}
func Acb_ObjFanin0(p *Acb_Ntk_t, i int32) int32 { return Acb_ObjFanins(p, i)[1] }
func Acb_ObjCioId(p *Acb_Ntk_t, i int32) int32  { return Acb_ObjFanins(p, i)[2] }

func Acb_ObjCopy(p *Acb_Ntk_t, i int32) int32 {
	assert(i > 0, "i>0")
	assert((Acb_NtkHasObjCopies(p) != 0), "Acb_NtkHasObjCopies(p)")
	return Vec_IntEntry(&p.vObjCopy, i)
}
func Acb_ObjFunc(p *Acb_Ntk_t, i int32) int32 {
	assert(i > 0, "i>0")
	assert((Acb_NtkHasObjFuncs(p) != 0), "Acb_NtkHasObjFuncs(p)")
	return Vec_IntEntry(&p.vObjFunc, i)
}
func Acb_ObjWeight(p *Acb_Ntk_t, i int32) int32 {
	assert(i > 0, "i>0")
	assert((Acb_NtkHasObjWeights(p) != 0), "Acb_NtkHasObjWeights(p)")
	return Vec_IntEntry(&p.vObjWeight, i)
}
func Acb_ObjTruth(p *Acb_Ntk_t, i int32) word {
	assert(i > 0, "i>0")
	assert((Acb_NtkHasObjTruths(p) != 0), "Acb_NtkHasObjTruths(p)")
	return Vec_WrdEntry(&p.vObjTruth, i)
}
func Acb_ObjTruthP(p *Acb_Ntk_t, i int32) *word {
	assert(i > 0, "i>0")
	assert((Acb_NtkHasObjTruths(p) != 0), "Acb_NtkHasObjTruths(p)")
	return Vec_WrdEntryP(&p.vObjTruth, i)
}
func Acb_ObjName(p *Acb_Ntk_t, i int32) int32 {
	assert(i > 0, "i>0")
	assert((Acb_NtkHasObjNames(p) != 0), "Acb_NtkHasObjNames(p)")
	return Vec_IntEntry(&p.vObjName, i)
}
func Acb_ObjNameStr(p *Acb_Ntk_t, i int32) string {
	assert(i > 0, "i>0")
	return Acb_NtkStr(p, Acb_ObjName(p, i))
}
func Acb_ObjAttr(p *Acb_Ntk_t, i int32) int32 {
	assert(i >= 0, "i>=0")
	if Acb_NtkHasObjAttrs(p) != 0 {
		return Vec_IntEntry(&p.vObjAttr, i)
	} else {
		return 0
	}
}
func Acb_ObjAttrSize(p *Acb_Ntk_t, i int32) int32 {
	assert(i >= 0, "i>=0")
	if Acb_ObjAttr(p, i) != 0 {
		return Vec_IntEntry(&p.vAttrSto, Acb_ObjAttr(p, i))
	} else {
		return 0
	}
}
func Acb_ObjAttrArray(p *Acb_Ntk_t, i int32) []int32 {
	assert(i >= 0, "i>=0")
	if Acb_ObjAttr(p, i) != 0 {
		return Vec_IntArray(&p.vAttrSto)[Acb_ObjAttr(p, i)+1:]
	} else {
		return nil
	}
}
func Acb_ObjAttrValue(p *Acb_Ntk_t, i int32, x int32) int32 {
	var k int32
	var s int32 = Acb_ObjAttrSize(p, i)
	var a []int32 = Acb_ObjAttrArray(p, i)
	for k = 0; k < s; k += 2 {
		if a[k] == x {
			return a[k+1]
		}
	}
	return 0
}
func Acb_ObjLevelD(p *Acb_Ntk_t, i int32) int32 {
	assert(i > 0, "i>0")
	return Vec_IntEntry(&p.vLevelD, i)
}
func Acb_ObjLevelR(p *Acb_Ntk_t, i int32) int32 {
	assert(i > 0, "i>0")
	return Vec_IntEntry(&p.vLevelR, i)
}
func Acb_ObjPathD(p *Acb_Ntk_t, i int32) int32 {
	assert(i > 0, "i>0")
	return Vec_IntEntry(&p.vPathD, i)
}
func Acb_ObjPathR(p *Acb_Ntk_t, i int32) int32 {
	assert(i > 0, "i>0")
	return Vec_IntEntry(&p.vPathR, i)
}
func Acb_ObjCounts(p *Acb_Ntk_t, i int32) float32 {
	assert(i > 0, "i>0")
	return Vec_FltEntry(&p.vCounts, i)
}

// acbResizeCnfs mirrors Vec_WecInit/Vec_WecPushLevel while keeping the
// byte-vector entries separately typed rather than relying on C struct casts.
func acbResizeCnfs(p *Acb_Ntk_t, n int) {
	if n <= cap(p.vCnfs) {
		p.vCnfs = p.vCnfs[:n]
		return
	}
	old := p.vCnfs[:cap(p.vCnfs)]
	p.vCnfs = append(old, make([]Vec_Str_t, n-len(old))...)
}
func Acb_ObjCnfs(p *Acb_Ntk_t, i int32) *Vec_Str_t { assert(i > 0, "i>0"); return &p.vCnfs[i] }

func Acb_ObjSetCopy(p *Acb_Ntk_t, i int32, x int32) {
	assert(Acb_ObjCopy(p, i) == -1, "Acb_ObjCopy(p, i) == -1")
	Vec_IntWriteEntry(&p.vObjCopy, i, x)
}
func Acb_ObjSetFunc(p *Acb_Ntk_t, i int32, x int32) {
	assert(Acb_ObjFunc(p, i) == -1, "Acb_ObjFunc(p, i) == -1")
	Vec_IntWriteEntry(&p.vObjFunc, i, x)
}
func Acb_ObjSetWeight(p *Acb_Ntk_t, i int32, x int32) {
	assert(Acb_ObjWeight(p, i) == 0, "Acb_ObjWeight(p, i)== 0")
	Vec_IntWriteEntry(&p.vObjWeight, i, x)
}
func Acb_ObjSetTruth(p *Acb_Ntk_t, i int32, x word) {
	assert(Acb_ObjTruth(p, i) == 0, "Acb_ObjTruth(p, i) == 0")
	Vec_WrdWriteEntry(&p.vObjTruth, i, x)
}
func Acb_ObjSetName(p *Acb_Ntk_t, i int32, x int32) {
	assert(Acb_ObjName(p, i) == 0, "Acb_ObjName(p, i) ==  0")
	Vec_IntWriteEntry(&p.vObjName, i, x)
}
func Acb_ObjSetAttrs(p *Acb_Ntk_t, i int32, a []int32, s int32) {
	assert(Acb_ObjAttr(p, i) == 0, "Acb_ObjAttr(p, i) == 0")
	if a == nil {
		return
	}
	Vec_IntWriteEntry(&p.vObjAttr, i, Vec_IntSize(&p.vAttrSto))
	Vec_IntPush(&p.vAttrSto, s)
	for k := int32(0); k < s; k++ {
		Vec_IntPush(&p.vAttrSto, a[k])
	}
}
func Acb_ObjSetLevelD(p *Acb_Ntk_t, i int32, x int32) int32 {
	Vec_IntWriteEntry(&p.vLevelD, i, x)
	return x
}
func Acb_ObjSetLevelR(p *Acb_Ntk_t, i int32, x int32) int32 {
	Vec_IntWriteEntry(&p.vLevelR, i, x)
	return x
}
func Acb_ObjSetPathD(p *Acb_Ntk_t, i int32, x int32) int32 {
	Vec_IntWriteEntry(&p.vPathD, i, x)
	return x
}
func Acb_ObjSetPathR(p *Acb_Ntk_t, i int32, x int32) int32 {
	Vec_IntWriteEntry(&p.vPathR, i, x)
	return x
}
func Acb_ObjSetCounts(p *Acb_Ntk_t, i int32, x float32) float32 {
	Vec_FltWriteEntry(&p.vCounts, i, x)
	return x
}
func Acb_ObjUpdateLevelD(p *Acb_Ntk_t, i int32, x int32) int32 {
	Vec_IntUpdateEntry(&p.vLevelD, i, x)
	return x
}
func Acb_ObjUpdateLevelR(p *Acb_Ntk_t, i int32, x int32) int32 {
	Vec_IntUpdateEntry(&p.vLevelR, i, x)
	return x
}
func Acb_ObjAddToPathD(p *Acb_Ntk_t, i int32, x int32) int32 {
	Vec_IntAddToEntry(&p.vPathD, i, x)
	return x
}
func Acb_ObjAddToPathR(p *Acb_Ntk_t, i int32, x int32) int32 {
	Vec_IntAddToEntry(&p.vPathR, i, x)
	return x
}

func Acb_ObjNtkId(p *Acb_Ntk_t, i int32) int32 {
	assert(i > 0, "i>0")
	if Acb_ObjIsBox(p, i) != 0 {
		return Acb_ObjFanin(p, i, Acb_ObjFaninNum(p, i))
	} else {
		return 0
	}
}
func Acb_ObjNtk(p *Acb_Ntk_t, i int32) *Acb_Ntk_t {
	assert(i > 0, "i>0")
	return Acb_NtkNtk(p, Acb_ObjNtkId(p, i))
}
func Acb_ObjIsSeq(p *Acb_Ntk_t, i int32) int32 {
	assert(i > 0, "i>0")
	if Acb_ObjIsBox(p, i) != 0 {
		return int32(Acb_ObjNtk(p, i).fSeq)
	} else {
		return Acb_TypeIsSeq(Acb_ObjType(p, i))
	}
}

func Acb_ObjRangeId(p *Acb_Ntk_t, i int32) int32 {
	if Acb_NtkHasObjRanges(p) != 0 {
		return Vec_IntEntry(&p.vObjRange, i)
	} else {
		return 0
	}
}
func Acb_ObjRange(p *Acb_Ntk_t, i int32) int32     { return Abc_Lit2Var(Acb_ObjRangeId(p, i)) }
func Acb_ObjLeft(p *Acb_Ntk_t, i int32) int32      { return Acb_NtkRangeLeft(p, Acb_ObjRange(p, i)) }
func Acb_ObjRight(p *Acb_Ntk_t, i int32) int32     { return Acb_NtkRangeRight(p, Acb_ObjRange(p, i)) }
func Acb_ObjSigned(p *Acb_Ntk_t, i int32) int32    { return Abc_LitIsCompl(Acb_ObjRangeId(p, i)) }
func Acb_ObjRangeSize(p *Acb_Ntk_t, i int32) int32 { return Acb_NtkRangeSize(p, Acb_ObjRange(p, i)) }
func Acb_ObjSetRangeSign(p *Acb_Ntk_t, i int32, x int32) {
	assert((Acb_NtkHasObjRanges(p) != 0), "Acb_NtkHasObjRanges(p)")
	Vec_IntWriteEntry(&p.vObjRange, i, x)
}
func Acb_ObjSetRange(p *Acb_Ntk_t, i int32, x int32) {
	assert((Acb_NtkHasObjRanges(p) != 0), "Acb_NtkHasObjRanges(p)")
	Vec_IntWriteEntry(&p.vObjRange, i, Abc_Var2Lit(x, 0))
}
func Acb_ObjHashRange(p *Acb_Ntk_t, i int32, l int32, r int32) {
	Acb_ObjSetRange(p, i, Acb_NtkHashRange(p, l, r))
}
func Acb_ObjRangeSign(p *Acb_Ntk_t, i int32) int32 {
	return Abc_Var2Lit(Acb_ObjRangeSize(p, i), Acb_ObjSigned(p, i))
}

func Acb_ObjTravId(p *Acb_Ntk_t, i int32) int32 { return Vec_IntEntry(&p.vObjTrav, i) }
func Acb_ObjTravIdDiff(p *Acb_Ntk_t, i int32) int32 {
	return p.nObjTravs - Vec_IntEntry(&p.vObjTrav, i)
}
func Acb_ObjIsTravIdCur(p *Acb_Ntk_t, i int32) int32 {
	return boolInt(Acb_ObjTravId(p, i) == p.nObjTravs, 1, 0)
}
func Acb_ObjIsTravIdPrev(p *Acb_Ntk_t, i int32) int32 {
	return boolInt(Acb_ObjTravId(p, i) == p.nObjTravs-1, 1, 0)
}
func Acb_ObjIsTravIdDiff(p *Acb_Ntk_t, i int32, d int32) int32 {
	return boolInt(Acb_ObjTravId(p, i) == p.nObjTravs-d, 1, 0)
}
func Acb_ObjSetTravIdCur(p *Acb_Ntk_t, i int32) int32 {
	var r int32 = Acb_ObjIsTravIdCur(p, i)
	Vec_IntWriteEntry(&p.vObjTrav, i, p.nObjTravs)
	return r
}
func Acb_ObjSetTravIdPrev(p *Acb_Ntk_t, i int32) int32 {
	var r int32 = Acb_ObjIsTravIdPrev(p, i)
	Vec_IntWriteEntry(&p.vObjTrav, i, p.nObjTravs-1)
	return r
}
func Acb_ObjSetTravIdDiff(p *Acb_Ntk_t, i int32, d int32) int32 {
	var r int32 = Acb_ObjTravIdDiff(p, i)
	Vec_IntWriteEntry(&p.vObjTrav, i, p.nObjTravs-d)
	return r
}
func Acb_NtkTravId(p *Acb_Ntk_t) int32 { return p.nObjTravs }
func Acb_NtkIncTravId(p *Acb_Ntk_t) {
	if !(Acb_NtkHasObjTravs(p) != 0) {
		Acb_NtkCleanObjTravs(p)
	}
	p.nObjTravs++
}

////////////////////////////////////////////////////////////////////////
///                          ITERATORS                               ///
////////////////////////////////////////////////////////////////////////

// Original macro:
// #define Acb_ManForEachNtk( p, pNtk, i )                                   \
//     for ( i = 1; (i <= Acb_ManNtkNum(p)) && (((pNtk) = Acb_ManNtk(p, i)), 1); i++ )

// Original macro:
// #define Acb_NtkForEachPi( p, iObj, i )                                    \
//     for ( i = 0; (i < Acb_NtkPiNum(p))  && (((iObj) = Acb_NtkCi(p, i)), 1); i++ )

// Original macro:
// #define Acb_NtkForEachPo( p, iObj, i )                                    \
//     for ( i = 0; (i < Acb_NtkPoNum(p))  && (((iObj) = Acb_NtkCo(p, i)), 1); i++ )

// Original macro:
// #define Acb_NtkForEachCi( p, iObj, i )                                    \
//     for ( i = 0; (i < Acb_NtkCiNum(p))  && (((iObj) = Acb_NtkCi(p, i)), 1); i++ )

// Original macro:
// #define Acb_NtkForEachCo( p, iObj, i )                                    \
//     for ( i = 0; (i < Acb_NtkCoNum(p))  && (((iObj) = Acb_NtkCo(p, i)), 1); i++ )

// Original macro:
// #define Acb_NtkForEachCoDriver( p, iObj, i )                              \
//     for ( i = 0; (i < Acb_NtkCoNum(p))  && (((iObj) = Acb_ObjFanin(p, Acb_NtkCo(p, i), 0)), 1); i++ )

// Original macro:
// #define Acb_NtkForEachCoAndDriver( p, iObj, iDriver, i )                  \
//     for ( i = 0; (i < Acb_NtkCoNum(p))  && (((iObj) = Acb_NtkCo(p, i)), 1)  && (((iDriver) = Acb_ObjFanin(p, iObj, 0)), 1); i++ )

// Original macro:
// #define Acb_NtkForEachCiVec( vVec, p, iObj, i )                           \
//     for ( i = 0; (i < Vec_IntSize(vVec))  && (((iObj) = Acb_NtkCi(p, Vec_IntEntry(vVec,i))), 1); i++ )

// Original macro:
// #define Acb_NtkForEachCoVec( vVec, p, iObj, i )                           \
//     for ( i = 0; (i < Vec_IntSize(vVec))  && (((iObj) = Acb_NtkCo(p, Vec_IntEntry(vVec,i))), 1); i++ )

// Original macro:
// #define Acb_NtkForEachCoDriverVec( vVec, p, iObj, i )                     \
//     for ( i = 0; (i < Vec_IntSize(vVec))  && (((iObj) = Acb_ObjFanin(p, Acb_NtkCo(p, Vec_IntEntry(vVec,i)), 0)), 1); i++ )

// Original macro:
// #define Acb_NtkForEachBoxSeq( p, iObj, i )                                \
//     for ( i = 0; (i < Acb_NtkSeqNum(p))  && (((iObj) = Acb_NtkBoxSeq(p, i)), 1); i++ )

// Original macro:
// #define Acb_NtkForEachCioOrder( p, iObj, i )                              \
//     for ( i = 0; (i < Acb_NtkCioOrderNum(p))  && (((iObj) = Acb_NtkCioOrder(p, i)), 1); i++ )

// Original macro:
// #define Acb_NtkForEachObj( p, i )                                         \
//     for ( i = 1; i < Vec_StrSize(&p->vObjType); i++ ) if ( Acb_ObjType(p, i) == 0 ) {} else

// Original macro:
// #define Acb_NtkForEachObjReverse( p, i )                                  \
//     for ( i = Vec_StrSize(&p->vObjType)-1; i > 0; i-- ) if ( Acb_ObjType(p, i) == 0 ) {} else

// Original macro:
// #define Acb_NtkForEachNode( p, i )                                        \
//     for ( i = 1; i < Vec_StrSize(&p->vObjType); i++ ) if ( Acb_ObjType(p, i) == 0 || Acb_ObjIsCio(p, i) ) {} else

// Original macro:
// #define Acb_NtkForEachNodeSupp( p, i, nSuppSize )                         \
//     for ( i = 1; i < Vec_StrSize(&p->vObjType); i++ ) if ( Acb_ObjType(p, i) == 0 || Acb_ObjIsCio(p, i) != 0 || Acb_ObjFaninNum(p, i) != nSuppSize ) {} else

// Original macro:
// #define Acb_NtkForEachNodeReverse( p, i )                                 \
//     for ( i = Vec_StrSize(&p->vObjType)-1; i > 0; i-- ) if ( Acb_ObjType(p, i) == 0 || Acb_ObjIsCio(p, i) ) {} else

// Original macro:
// #define Acb_NtkForEachObjType( p, Type, i )                               \
//     for ( i = 1; i < Vec_StrSize(&p->vObjType) && (((Type) = Acb_ObjType(p, i)), 1); i++ ) if ( !Type ) {} else

// Original macro:
// #define Acb_NtkForEachBox( p, i )                                         \
//     for ( i = 1; i < Vec_StrSize(&p->vObjType); i++ ) if ( !Acb_ObjIsBox(p, i) ) {} else

// Original macro:
// #define Acb_ObjForEachFanin( p, iObj, iFanin, k )                         \
//     for ( k = 0; k < Acb_ObjFaninNum(p, iObj) && ((iFanin = Acb_ObjFanin(p, iObj, k)), 1); k++ )

// Original macro:
// #define Acb_ObjForEachFaninFast( p, iObj, pFanins, iFanin, k )            \
//     for ( k = 0, pFanins = Acb_ObjFanins(p, iObj); k < pFanins[0] && ((iFanin = pFanins[k+1]), 1); k++ )

// Original macro:
// #define Acb_ObjForEachFanout( p, iObj, iFanout, k )                       \
//     Vec_IntForEachEntry( Vec_WecEntry(&p->vFanouts, iObj), iFanout, k ) if ( Acb_ObjType(p, iFanout) == 0 ) {} else

// Original macro:
// #define Acb_ObjForEachFon( p, iObj, iFon )                                \
//     for ( assert(Acb_ObjIsBox(p, iObj)), iFon = iObj + 1; iFon < Acb_NtkObjNum(p) && Acb_ObjIsFon(p, iFon); iFon++ )

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
func Acb_ObjFonNum(p *Acb_Ntk_t, iObj int32) int32 {
	var i, Count int32
	if !(Acb_ObjIsBox(p, iObj) != 0) {
		return 0
	}
	// inlining of macro Acb_ObjForEachFon
	assert(Acb_ObjIsBox(p, iObj) != 0, "Acb_ObjIsBox(p, iObj)")
	for i = iObj + 1; i < Acb_NtkObjNumMax(p) && (Acb_ObjIsFon(p, i) != 0); i++ {
		Count++
	}
	return Count
}
func Acb_ObjWhatFanin(p *Acb_Ntk_t, iObj int32, iFaninGiven int32) int32 {
	var k int32
	var iFanin int32
	var pFanins []int32
	// inlining of macro Acb_ObjForEachFaninFast
	pFanins = Acb_ObjFanins(p, iObj)
	for k = 0; k < pFanins[0]; k++ {
		iFanin = pFanins[k+1]
		if iFanin == iFaninGiven {
			return k
		}
	}
	return -1
}
func Acb_ObjAddFanin(p *Acb_Ntk_t, iObj int32, iFanin int32) {
	var pFanins []int32 = Acb_ObjFanins(p, iObj)
	assert(iFanin > 0, "iFanin > 0")
	assert(pFanins[1+pFanins[0]] == -1, "pFanins[ 1 + pFanins[0] ] == -1")
	pFanins[1+pFanins[0]] = iFanin
	pFanins[0]++
}
func Acb_ObjDeleteFaninIndex(p *Acb_Ntk_t, iObj int32, iFaninIndex int32) {
	var i int32
	var pFanins []int32 = Acb_ObjFanins(p, iObj)
	pFanins[0]--
	for i = iFaninIndex; i < pFanins[0]; i++ {
		pFanins[1+i] = pFanins[2+i]
	}
	pFanins[1+pFanins[0]] = -1
}
func Acb_ObjDeleteFanin(p *Acb_Ntk_t, iObj int32, iFanin int32) {
	var pFanins []int32 = Acb_ObjFanins(p, iObj)
	var iFaninIndex int32 = Acb_ObjWhatFanin(p, iObj, iFanin)
	assert(pFanins[1+iFaninIndex] == iFanin, "pFanins[ 1 + iFaninIndex ] == iFanin")
	Acb_ObjDeleteFaninIndex(p, iObj, iFaninIndex)
}
func Acb_ObjAddFanins(p *Acb_Ntk_t, iObj int32, vFanins *Vec_Int_t) {
	var i, iFanin int32
	if vFanins != nil {
		// inlining of macro Vec_IntForEachEntry
		for i = 0; i < Vec_IntSize(vFanins); i++ {
			iFanin = Vec_IntEntry(vFanins, i)
			Acb_ObjAddFanin(p, iObj, iFanin)
		}
	}
}

// Acb_ObjSetNtkId should be called after fanins are added.
func Acb_ObjSetNtkId(p *Acb_Ntk_t, iObj, iNtk int32) {
	var pFanins []int32 = Acb_ObjFanins(p, iObj)
	assert(pFanins[1+pFanins[0]] == -1, "pFanins[ 1 + pFanins[0] ] == -1")
	pFanins[1+pFanins[0]] = iNtk
}
func Acb_ObjRemoveFanins(p *Acb_Ntk_t, iObj int32) {
	var i int32
	var pFanins []int32 = Acb_ObjFanins(p, iObj)
	for i = 1; i <= pFanins[0]; i++ {
		assert(pFanins[i] > 0, "pFanins[i] > 0")
		pFanins[i] = -1
	}
	pFanins[0] = 0
}
func Acb_ObjAlloc(p *Acb_Ntk_t, Type Acb_ObjType_t, nFans int32, nFons int32) int32 {
	var i int32
	var nFansReal int32
	var CioId int32 = -1
	var iObj int32 = Vec_StrSize(&p.vObjType)
	Vec_StrPush(&p.vObjType, byte(Type))
	if Type == ABC_OPER_CI {
		assert(nFans == 0, "nFans == 0")
		CioId = Vec_IntSize(&p.vCis)
		Vec_IntPush(&p.vCis, iObj)
		nFansReal = 2

	} else if Type == ABC_OPER_CO {
		assert(nFans == 1, "nFans == 1")
		CioId = Vec_IntSize(&p.vCos)
		Vec_IntPush(&p.vCos, iObj)
		nFansReal = 2

	} else {
		nFansReal = Abc_MaxInt(nFans+Acb_ObjIsBox(p, iObj), p.nFaninMax)
	}
	// add fins
	Vec_IntPush(&p.vObjFans, Vec_IntSize(&p.vFanSto))
	Vec_IntPush(&p.vFanSto, 0)
	for i = 0; i < nFansReal; i++ {
		if CioId >= 0 && i == 1 {
			Vec_IntPush(&p.vFanSto, CioId)
		} else {
			Vec_IntPush(&p.vFanSto, -1)
		}
	}
	// add fons
	assert(!(Acb_ObjIsBox(p, iObj) != 0) || nFons > 0, "!Acb_ObjIsBox(p, iObj) || nFons > 0")
	for i = 0; i < nFons; i++ {
		Acb_ObjAddFanin(p, Acb_ObjAlloc(p, ABC_OPER_FON, 1, 0), iObj)
	}
	// extend storage
	if Acb_NtkHasObjCopies(p) != 0 {
		Vec_IntPush(&p.vObjCopy, -1)
	}
	if Acb_NtkHasObjFuncs(p) != 0 {
		Vec_IntPush(&p.vObjFunc, -1)
	}
	if Acb_NtkHasObjWeights(p) != 0 {
		Vec_IntPush(&p.vObjWeight, 0)
	}
	if Acb_NtkHasObjTruths(p) != 0 {
		Vec_WrdPush(&p.vObjTruth, 0)
	}
	if Acb_NtkHasObjNames(p) != 0 {
		Vec_IntPush(&p.vObjName, 0)
	}
	if Acb_NtkHasObjRanges(p) != 0 {
		Vec_IntPush(&p.vObjRange, 0)
	}
	if Acb_NtkHasObjTravs(p) != 0 {
		Vec_IntPush(&p.vObjTrav, 0)
	}
	if Acb_NtkHasObjAttrs(p) != 0 {
		Vec_IntPush(&p.vObjAttr, 0)
	}
	if Acb_NtkHasObjLevelD(p) != 0 {
		Vec_IntPush(&p.vLevelD, 0)
	}
	if Acb_NtkHasObjLevelR(p) != 0 {
		Vec_IntPush(&p.vLevelR, 0)
	}
	if Acb_NtkHasObjPathD(p) != 0 {
		Vec_IntPush(&p.vPathD, 0)
	}
	if Acb_NtkHasObjPathR(p) != 0 {
		Vec_IntPush(&p.vPathR, 0)
	}
	if Acb_NtkHasObjCounts(p) != 0 {
		Vec_FltPush(&p.vCounts, 0)
	}
	if Acb_NtkHasObjFanout(p) != 0 {
		Vec_WecPushLevel(&p.vFanouts)
	}
	if Acb_NtkHasObjCnfs(p) != 0 {
		acbResizeCnfs(p, len(p.vCnfs)+1)
	}
	if p.vQue != nil {
		Vec_QueSetPriority(p.vQue, Vec_FltArrayP(&p.vCounts))
	}
	return iObj
}
func Acb_ObjDup(pNew *Acb_Ntk_t, p *Acb_Ntk_t, i int32) int32 {
	var iObj int32 = Acb_ObjAlloc(pNew, Acb_ObjType(p, i), Acb_ObjFaninNum(p, i), Acb_ObjFonNum(p, i))
	Acb_ObjSetCopy(p, i, iObj)
	return iObj
}
func Acb_ObjDelete(p *Acb_Ntk_t, iObj int32) {
	var i int32
	Acb_ObjCleanType(p, iObj)
	if !(Acb_ObjIsBox(p, iObj) != 0) {
		return
	}
	// inlining of macro Acb_ObjForEachFon
	assert(Acb_ObjIsBox(p, iObj) != 0, "Acb_ObjIsBox(p, iObj)")
	for i = iObj + 1; i < Acb_NtkObjNumMax(p) && (Acb_ObjIsFon(p, i) != 0); i++ {
		Acb_ObjCleanType(p, i)
	}
}
func Acb_ObjAddFaninFanoutOne(p *Acb_Ntk_t, iObj int32, iFanin int32) {
	Vec_IntPush(Vec_WecEntry(&p.vFanouts, iFanin), iObj)
	Acb_ObjAddFanin(p, iObj, iFanin)
}
func Acb_ObjAddFaninFanout(p *Acb_Ntk_t, iObj int32) {
	var k int32
	var iFanin int32
	var pFanins []int32
	// inlining of macro Acb_ObjForEachFaninFast
	pFanins = Acb_ObjFanins(p, iObj)
	for k = 0; k < pFanins[0]; k++ {
		iFanin = pFanins[k+1]
		Vec_IntPush(Vec_WecEntry(&p.vFanouts, iFanin), iObj)
	}
}
func Acb_ObjRemoveFaninFanoutOne(p *Acb_Ntk_t, iObj int32, iFanin int32) {
	var RetValue int32 = Vec_IntRemove(Vec_WecEntry(&p.vFanouts, iFanin), iObj)
	assert((RetValue != 0), "RetValue")
	Acb_ObjDeleteFanin(p, iObj, iFanin)
}
func Acb_ObjRemoveFaninFanout(p *Acb_Ntk_t, iObj int32) {
	var k int32
	var iFanin int32
	var pFanins []int32
	// inlining of macro Acb_ObjForEachFaninFast
	pFanins = Acb_ObjFanins(p, iObj)
	for k = 0; k < pFanins[0]; k++ {
		iFanin = pFanins[k+1]
		var RetValue int32 = Vec_IntRemove(Vec_WecEntry(&p.vFanouts, iFanin), iObj)
		assert((RetValue != 0), "RetValue")
	}
}
func Acb_ObjPatchFanin(p *Acb_Ntk_t, iObj int32, iFanin int32, iFaninNew int32) {
	var i int32
	var RetValue int32
	var pFanins []int32 = Acb_ObjFanins(p, iObj)
	assert(iFanin != iFaninNew, "iFanin != iFaninNew")
	for i = 0; i < pFanins[0]; i++ {
		if pFanins[1+i] == iFanin {
			pFanins[1+i] = iFaninNew
		}
	}
	if !(Acb_NtkHasObjFanout(p) != 0) {
		return
	}
	RetValue = Vec_IntRemove(Vec_WecEntry(&p.vFanouts, iFanin), iObj)
	assert((RetValue != 0), "RetValue")
	Vec_IntPush(Vec_WecEntry(&p.vFanouts, iFaninNew), iObj)
}
func Acb_NtkCreateFanout(p *Acb_Ntk_t) {
	var iObj int32
	Acb_NtkCleanObjFanout(p)
	// inlining of macro Acb_NtkForEachObj
	for iObj = 1; iObj < Vec_StrSize(&p.vObjType); iObj++ {
		if Acb_ObjType(p, iObj) == ABC_OPER_NONE {
		} else {
			Acb_ObjAddFaninFanout(p, iObj)
		}
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
func Acb_NtkAlloc(p *Acb_Man_t, NameId int32, nCis int32, nCos int32, nObjs int32) *Acb_Ntk_t {
	// inlining of macro ABC_CALLOC
	var pNew *Acb_Ntk_t = &Acb_Ntk_t{}
	assert(nCis >= 0 && nCos >= 0 && nObjs >= 0, "nCis >= 0 && nCos >= 0 && nObjs >= 0")
	pNew.Id = Vec_PtrSize(&p.vNtks)
	Vec_PtrPush(&p.vNtks, pNew)
	pNew.NameId = NameId
	pNew.pDesign = p
	Vec_IntGrow(&pNew.vCis, nCis)
	Vec_IntGrow(&pNew.vCos, nCos)
	Vec_StrGrow(&pNew.vObjType, nObjs+1)
	Vec_StrPush(&pNew.vObjType, byte(ABC_OPER_NONE))
	Vec_IntGrow(&pNew.vObjFans, nObjs+1)
	Vec_IntPush(&pNew.vObjFans, -1)
	Vec_IntGrow(&pNew.vFanSto, 3*nObjs)
	return pNew
}
func Acb_NtkFree(p *Acb_Ntk_t) {
	// interface
	Vec_IntErase(&p.vCis)
	Vec_IntErase(&p.vCos)
	Vec_IntErase(&p.vOrder)
	Vec_IntErase(&p.vSeq)
	// stucture
	Vec_StrErase(&p.vObjType)
	Vec_IntErase(&p.vObjFans)
	Vec_IntErase(&p.vFanSto)
	// optional
	Vec_IntErase(&p.vObjCopy)
	Vec_IntErase(&p.vObjFunc)
	Vec_IntErase(&p.vObjWeight)
	Vec_WrdErase(&p.vObjTruth)
	Vec_IntErase(&p.vObjName)
	Vec_IntErase(&p.vObjRange)
	Vec_IntErase(&p.vObjTrav)
	Vec_IntErase(&p.vObjBits)
	Vec_IntErase(&p.vObjAttr)
	Vec_IntErase(&p.vAttrSto)
	Vec_IntErase(&p.vNtkObjs)
	Vec_IntErase(&p.vTargets)
	Vec_IntErase(&p.vLevelD)
	Vec_IntErase(&p.vLevelR)
	Vec_IntErase(&p.vPathD)
	Vec_IntErase(&p.vPathR)
	Vec_FltErase(&p.vCounts)
	Vec_WecErase(&p.vFanouts)
	p.vCnfs = nil
	Vec_StrErase(&p.vCnf)
	Vec_IntErase(&p.vSuppOld)
	// other
	Vec_QueFreeP(&p.vQue)
	Vec_IntErase(&p.vCover)
	Vec_IntErase(&p.vArray0)
	Vec_IntErase(&p.vArray1)
	Vec_IntErase(&p.vArray2)
	p = nil
}
func Acb_NtkNewStrId(pNtk *Acb_Ntk_t, format string, args ...any) int32 {
	return Abc_NamStrFindOrAdd(Acb_NtkNam(pNtk), fmt.Sprintf(format, args...), nil)
}
func Acb_ManNewConstId(p *Acb_Ntk_t, vBits *Vec_Str_t) int32 {
	text := fmt.Sprintf("%d'b%s", Vec_StrSize(vBits)-1, string(Vec_StrArray(vBits)))
	return Abc_NamStrFindOrAdd(p.pDesign.pFuns, text, nil)
}
func Acb_ManNewConstZero(p *Acb_Ntk_t, nBits int32) int32 {
	text := fmt.Sprintf("%d'b", nBits)
	return Abc_NamStrFindOrAdd(p.pDesign.pFuns, text, nil)
}
func Acb_NtkAdd(p *Acb_Man_t, pNtk *Acb_Ntk_t) {
	var fFound int32
	var NtkId int32 = Abc_NamStrFindOrAdd(p.pMods, Acb_NtkName(pNtk), &fFound)
	if fFound != 0 {
		fmt.Printf("Network with name \"%s\" already exists.\n", Acb_NtkName(pNtk))
	} else {
		assert(NtkId == pNtk.Id, "NtkId == pNtk.Id")
	}
}
func Acb_NtkUpdate(p *Acb_Man_t, pNtk *Acb_Ntk_t) {
	var fFound int32
	var NtkId int32 = Abc_NamStrFindOrAdd(p.pMods, Acb_NtkName(pNtk), &fFound)
	if !(fFound != 0) {
		fmt.Printf("Network with name \"%s\" does not exist.\n", Acb_NtkName(pNtk))
	} else {
		Acb_NtkFree(Acb_ManNtk(p, NtkId))
		Vec_PtrWriteEntry(&p.vNtks, NtkId, pNtk)
	}
}
func Acb_NtkCollect(p *Acb_Ntk_t) *Vec_Int_t {
	var iObj int32
	var vObjs *Vec_Int_t = Vec_IntAlloc(Acb_NtkObjNum(p))
	// inlining of macro Acb_NtkForEachObj
	for iObj = 1; iObj < Vec_StrSize(&p.vObjType); iObj++ {
		if Acb_ObjType(p, iObj) == ABC_OPER_NONE {
		} else {
			Vec_IntPush(vObjs, iObj)
		}
	}
	return vObjs
}
func Acb_NtkDup(pMan *Acb_Man_t, p *Acb_Ntk_t, vObjs *Vec_Int_t) *Acb_Ntk_t {
	var pNew *Acb_Ntk_t
	var i, k, iObj, iObjNew, iFanin int32
	pNew = Acb_NtkAlloc(pMan, Acb_NtkNameId(p), Acb_NtkCiNum(p), Acb_NtkCoNum(p), Vec_IntSize(vObjs))
	Acb_NtkCleanObjCopies(p)
	// inlining of macro Vec_IntForEachEntry
	for i = 0; i < Vec_IntSize(vObjs); i++ {
		iObj = Vec_IntEntry(vObjs, i)
		iObjNew = Acb_ObjDup(pNew, p, iObj)
	}
	// inlining of macro Vec_IntForEachEntry
	for i = 0; i < Vec_IntSize(vObjs); i++ {
		iObj = Vec_IntEntry(vObjs, i)
		iObjNew = Acb_ObjCopy(p, iObj)
		// inlining of macro Acb_ObjForEachFanin
		for k = 0; k < Acb_ObjFaninNum(p, iObj); k++ {
			iFanin = Acb_ObjFanin(p, iObj, k)
			Acb_ObjAddFanin(pNew, iObjNew, Acb_ObjCopy(p, iFanin))
		}
	}
	//Acb_NtkFreeObjCopies( p );
	assert(Acb_NtkObjNum(pNew) == Acb_NtkObjNumAlloc(pNew), "Acb_NtkObjNum(pNew) == Acb_NtkObjNumAlloc(pNew)")
	Acb_NtkSetCopy(p, Acb_NtkId(pNew))
	Acb_NtkSetRegNum(pNew, Acb_NtkRegNum(p))
	return pNew
}
func Acb_NtkDupOrder(pMan *Acb_Man_t, p *Acb_Ntk_t, pFuncOrder func(*Acb_Ntk_t) *Vec_Int_t) *Acb_Ntk_t {
	var pNew *Acb_Ntk_t
	var vObjs *Vec_Int_t
	if pFuncOrder != nil {
		vObjs = pFuncOrder(p)
	} else {
		vObjs = Acb_NtkCollect(p)
	}
	if vObjs == nil {
		return nil
	}
	pNew = Acb_NtkDup(pMan, p, vObjs)
	Vec_IntFree(vObjs)
	//Acb_NtkPrepareSeq( pNew );
	return pNew
}
func Acb_NtkDupAttrs(pNew *Acb_Ntk_t, p *Acb_Ntk_t) {
	var i, iObj int32
	assert(Vec_IntSize(&pNew.vOrder) == 0, "Vec_IntSize(&pNew.vOrder) == 0")
	// inlining of macro Acb_NtkForEachCioOrder
	for i = 0; i < Acb_NtkCioOrderNum(p); i++ {
		iObj = Acb_NtkCioOrder(p, i)
		Vec_IntPush(&pNew.vOrder, Acb_ObjCopy(p, iObj))
	}
	//    Vec_IntRemapArray( &p->vObjCopy, &p->vOrder,    &pNew->vOrder,    Acb_NtkCioOrderNum(p) );
	//    Vec_IntRemapArray( &p->vObjCopy, &p->vSeq,      &pNew->vSeq,      Acb_NtkSeqNum(p) );
	// transfer object attributes
	Vec_IntRemapArray(&p.vObjCopy, &p.vObjFunc, &pNew.vObjFunc, Acb_NtkObjNum(pNew)+1)
	Vec_IntRemapArray(&p.vObjCopy, &p.vObjWeight, &pNew.vObjWeight, Acb_NtkObjNum(pNew)+1)
	//    Vec_WrdRemapArray( &p->vObjCopy, &p->vObjTruth,  &pNew->vObjTruth,  Acb_NtkObjNum(pNew) + 1 );
	Vec_IntRemapArray(&p.vObjCopy, &p.vObjName, &pNew.vObjName, Acb_NtkObjNum(pNew)+1)
	Vec_IntRemapArray(&p.vObjCopy, &p.vObjRange, &pNew.vObjRange, Acb_NtkObjNum(pNew)+1)
	Vec_IntRemapArray(&p.vObjCopy, &p.vObjAttr, &pNew.vObjAttr, Acb_NtkObjNum(pNew)+1)
	// duplicate attributes
	Vec_IntAppend(&pNew.vAttrSto, &p.vAttrSto)
}
func Acb_NtkMemory(p *Acb_Ntk_t) int32 {
	var nMem int32 = int32(unsafe.Sizeof(Acb_Ntk_t{}))
	// interface
	nMem += int32(Vec_IntMemory(&p.vCis))
	nMem += int32(Vec_IntMemory(&p.vCos))
	nMem += int32(Vec_IntMemory(&p.vOrder))
	nMem += int32(Vec_IntMemory(&p.vSeq))
	// stucture
	nMem += int32(Vec_StrMemory(&p.vObjType))
	nMem += int32(Vec_IntMemory(&p.vObjFans))
	nMem += int32(Vec_IntMemory(&p.vFanSto))
	nMem += int32(Vec_WecMemory(&p.vFanouts))
	// optional
	nMem += int32(Vec_IntMemory(&p.vObjCopy))
	nMem += int32(Vec_IntMemory(&p.vObjFunc))
	nMem += int32(Vec_IntMemory(&p.vObjWeight))
	nMem += int32(Vec_WrdMemory(&p.vObjTruth))
	nMem += int32(Vec_IntMemory(&p.vObjName))
	nMem += int32(Vec_IntMemory(&p.vObjRange))
	nMem += int32(Vec_IntMemory(&p.vObjTrav))
	nMem += int32(Vec_IntMemory(&p.vObjBits))
	nMem += int32(Vec_IntMemory(&p.vObjAttr))
	nMem += int32(Vec_IntMemory(&p.vAttrSto))
	nMem += int32(Vec_IntMemory(&p.vNtkObjs))
	nMem += int32(Vec_IntMemory(&p.vTargets))
	// other
	nMem += int32(Vec_IntMemory(&p.vCover))
	nMem += int32(Vec_IntMemory(&p.vArray0))
	nMem += int32(Vec_IntMemory(&p.vArray1))
	nMem += int32(Vec_IntMemory(&p.vArray2))
	return nMem
}
func Acb_NtkPrintStats(p *Acb_Ntk_t) {
	fmt.Printf("pi =%5d  ", Acb_NtkPiNum(p))
	fmt.Printf("po =%5d  ", Acb_NtkPoNum(p))
	fmt.Printf("ff =%5d  ", Acb_NtkRegNum(p))
	fmt.Printf("node =%5d  ", Acb_NtkNodeNum(p))
	fmt.Printf("box =%5d  ", Acb_NtkBoxNum(p))
	//printf( "topo =%4s  ", Acb_NtkIsTopoOrder(p) ? "yes" : "no" );
	fmt.Printf("  %s ", Acb_NtkName(p))
	if Vec_IntSize(&p.vNtkObjs) != 0 {
		fmt.Printf("-> %s", Acb_NtkName(Acb_NtkNtk(p, Vec_IntEntry(&p.vNtkObjs, 0))))
	}
	fmt.Printf("\n")
	// Vec_StrIntPrint( &p->vObjType );
}
func Acb_NtkPrint(p *Acb_Ntk_t) {
	var i, Type int32
	fmt.Printf("Interface (%d):\n", Acb_NtkCioNum(p))
	fmt.Printf("Objects (%d):\n", Acb_NtkObjNum(p))
	// inlining of macro Acb_NtkForEachObjType
	for i = 1; i < Vec_StrSize(&p.vObjType); i++ {
		Type = int32(Acb_ObjType(p, i))
		if Type != 0 {
			fmt.Printf("%6d : ", i)
			fmt.Printf("Type =%3d  ", Type)
			fmt.Printf("Fanins = %d  ", Acb_ObjFaninNum(p, i))
			if (Acb_NtkHasObjNames(p) != 0) && (Acb_ObjName(p, i) != 0) {
				fmt.Printf("%s", Acb_ObjNameStr(p, i))
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
func Acb_ManAlloc(pFileName string, nNtks int32, pStrs *Abc_Nam_t, pFuns *Abc_Nam_t, pMods *Abc_Nam_t, vHash *Hash_IntMan_t) *Acb_Man_t {
	// inlining of macro ABC_CALLOC
	var pNew *Acb_Man_t = &Acb_Man_t{}
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
func Acb_ManDupTypeNames(pNew *Acb_Man_t, p *Acb_Man_t) {
	pNew.pTypeNames = p.pTypeNames
}
func Acb_ManDup(p *Acb_Man_t, pFuncOrder func(*Acb_Ntk_t) *Vec_Int_t) *Acb_Man_t {
	var pNtk, pNtkNew *Acb_Ntk_t
	var i int32
	var pNew *Acb_Man_t = Acb_ManAlloc(p.pSpec, Acb_ManNtkNum(p), Abc_NamRef(p.pStrs), Abc_NamRef(p.pFuns), Abc_NamStart(100, 24), Hash_IntManRef(p.vHash))
	Acb_ManDupTypeNames(pNew, p)
	// inlining of macro Acb_ManForEachNtk
	for i = 1; i <= Acb_ManNtkNum(p); i++ {
		pNtk = Acb_ManNtk(p, i)
		pNtkNew = Acb_NtkDupOrder(pNew, pNtk, pFuncOrder)
		Acb_NtkAdd(pNew, pNtkNew)
		Acb_NtkDupAttrs(pNtkNew, pNtk)
	}
	//    Acb_ManForEachNtk( p, pNtk, i )
	//        if ( (pHost = Acb_NtkHostNtk(pNtk)) )
	//            Acb_NtkSetHost( Acb_NtkCopyNtk(pNew, pNtk), Acb_NtkCopy(pHost), Acb_ObjCopy(pHost, Acb_NtkHostObj(pNtk)) );
	pNew.iRoot = Acb_ManNtkNum(pNew)
	return pNew
}
func Acb_ManFree(p *Acb_Man_t) {
	var pNtk *Acb_Ntk_t
	var i int32
	// inlining of macro Acb_ManForEachNtk
	for i = 1; i <= Acb_ManNtkNum(p); i++ {
		pNtk = Acb_ManNtk(p, i)
		Acb_NtkFree(pNtk)
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
func Acb_ManGetMap(p *Acb_Man_t, i int32) int32 {
	return Vec_IntGetEntry(&p.vNameMap, i)
}
func Acb_ManSetMap(p *Acb_Man_t, i int32, x int32) {
	assert(Vec_IntGetEntry(&p.vNameMap, i) == 0, "Vec_IntGetEntry(&p.vNameMap, i) == 0")
	Vec_IntSetEntry(&p.vNameMap, i, x)
	Vec_IntPush(&p.vUsed, i)
}
func Acb_ManUnsetMap(p *Acb_Man_t, i int32) {
	Vec_IntSetEntry(&p.vNameMap, i, 0)
}
func Acb_ManCleanMap(p *Acb_Man_t) {
	var i, Entry int32
	// inlining of macro Vec_IntForEachEntry
	for i = 0; i < Vec_IntSize(&p.vUsed); i++ {
		Entry = Vec_IntEntry(&p.vUsed, i)
		Vec_IntWriteEntry(&p.vNameMap, Entry, 0)
	}
	Vec_IntClear(&p.vUsed)
}
func Acb_NtkGetMap(p *Acb_Ntk_t, i int32) int32    { return Acb_ManGetMap(p.pDesign, i) }
func Acb_NtkSetMap(p *Acb_Ntk_t, i int32, x int32) { Acb_ManSetMap(p.pDesign, i, x) }
func Acb_NtkUnsetMap(p *Acb_Ntk_t, i int32)        { Acb_ManUnsetMap(p.pDesign, i) }
func Acb_NtkCleanMap(p *Acb_Ntk_t)                 { Acb_ManCleanMap(p.pDesign) }

func Acb_ManGetMap2(p *Acb_Man_t, i int32) int32 {
	return Vec_IntGetEntry(&p.vNameMap2, i)
}
func Acb_ManSetMap2(p *Acb_Man_t, i int32, x int32) {
	assert(Vec_IntGetEntry(&p.vNameMap2, i) == 0, "Vec_IntGetEntry(&p.vNameMap2, i) == 0")
	Vec_IntSetEntry(&p.vNameMap2, i, x)
	Vec_IntPush(&p.vUsed2, i)
}
func Acb_ManUnsetMap2(p *Acb_Man_t, i int32) {
	Vec_IntSetEntry(&p.vNameMap2, i, 0)
}
func Acb_ManCleanMap2(p *Acb_Man_t) {
	var i, Entry int32
	// inlining of macro Vec_IntForEachEntry
	for i = 0; i < Vec_IntSize(&p.vUsed2); i++ {
		Entry = Vec_IntEntry(&p.vUsed2, i)
		Vec_IntWriteEntry(&p.vNameMap2, Entry, 0)
	}
	Vec_IntClear(&p.vUsed2)
}
func Acb_NtkGetMap2(p *Acb_Ntk_t, i int32) int32    { return Acb_ManGetMap2(p.pDesign, i) }
func Acb_NtkSetMap2(p *Acb_Ntk_t, i int32, x int32) { Acb_ManSetMap2(p.pDesign, i, x) }
func Acb_NtkUnsetMap2(p *Acb_Ntk_t, i int32)        { Acb_ManUnsetMap2(p.pDesign, i) }
func Acb_NtkCleanMap2(p *Acb_Ntk_t)                 { Acb_ManCleanMap2(p.pDesign) }

func Acb_ManMemory(p *Acb_Man_t) int32 {
	var pNtk *Acb_Ntk_t
	var i int32
	var nMem int32 = int32(unsafe.Sizeof(Acb_Man_t{}))
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
	// inlining of macro Acb_ManForEachNtk
	for i = 1; i <= Acb_ManNtkNum(p); i++ {
		pNtk = Acb_ManNtk(p, i)
		nMem += Acb_NtkMemory(pNtk)
	}
	return nMem
}
func Acb_ManObjNum(p *Acb_Man_t) int32 {
	var pNtk *Acb_Ntk_t
	var i, Count int32
	// inlining of macro Acb_ManForEachNtk
	for i = 1; i <= Acb_ManNtkNum(p); i++ {
		pNtk = Acb_ManNtk(p, i)
		Count += Acb_NtkObjNum(pNtk)
	}
	return Count
}
func Acb_ManBoxNum(p *Acb_Man_t) int32 {
	var pNtk *Acb_Ntk_t
	var i, Count int32
	// inlining of macro Acb_ManForEachNtk
	for i = 1; i <= Acb_ManNtkNum(p); i++ {
		pNtk = Acb_ManNtk(p, i)
		Count += Acb_NtkBoxNum(pNtk)
	}
	return Count
}
func Acb_ManBoxNumRec_rec(p *Acb_Ntk_t, pCountP []int32, pCountU []int32) {
	var iObj int32
	var Id int32 = Acb_NtkId(p)
	if pCountP[Id] >= 0 {
		return
	}
	pCountP[Id], pCountU[Id] = 0, 0
	// inlining of macro Acb_NtkForEachObj
	for iObj = 1; iObj < Vec_StrSize(&p.vObjType); iObj++ {
		if Acb_ObjType(p, iObj) == ABC_OPER_NONE {
		} else {
			if Acb_ObjIsBox(p, iObj) != 0 {
				Acb_ManBoxNumRec_rec(Acb_ObjNtk(p, iObj), pCountP, pCountU)
				pCountP[Id] += pCountP[Acb_ObjNtkId(p, iObj)]
				pCountU[Id] += pCountU[Acb_ObjNtkId(p, iObj)] + 1

			} else {
				pCountP[Id] += 1
			}
		}
	}
}
func Acb_ManBoxNumRec(p *Acb_Man_t, pnPrims *int32, pnUsers *int32) {
	var pNtk *Acb_Ntk_t = Acb_ManRoot(p)
	// inlining of macro ABC_FALLOC
	pCountP := make([]int32, Acb_ManNtkNum(p)+1)
	// inlining of macro ABC_FALLOC
	pCountU := make([]int32, Acb_ManNtkNum(p)+1)
	for i := range pCountP {
		pCountP[i], pCountU[i] = -1, -1
	}
	Acb_ManBoxNumRec_rec(pNtk, pCountP, pCountU)
	*pnPrims = pCountP[Acb_NtkId(pNtk)]
	*pnUsers = pCountU[Acb_NtkId(pNtk)]
	pCountP = nil
	pCountU = nil
}
func Acb_ManPrintStats(p *Acb_Man_t, nModules int32, fVerbose int32) {
	var pNtk *Acb_Ntk_t
	var i, nPrims, nUsers int32
	var pRoot *Acb_Ntk_t = Acb_ManRoot(p)
	Acb_ManBoxNumRec(p, &nPrims, &nUsers)
	fmt.Printf("%-12s : ", Acb_ManName(p))
	fmt.Printf("pi =%5d  ", Acb_NtkCiNum(pRoot))
	fmt.Printf("po =%5d  ", Acb_NtkCoNum(pRoot))
	fmt.Printf("mod =%5d  ", Acb_ManNtkNum(p))
	fmt.Printf("box =%5d  ", nPrims+nUsers)
	fmt.Printf("prim =%5d  ", nPrims)
	fmt.Printf("user =%5d  ", nUsers)
	fmt.Printf("mem =%6.3f MB", float64(Acb_ManMemory(p))/(1<<20))
	fmt.Printf("\n")
	// inlining of macro Acb_ManForEachNtk
	for i = 1; i <= Acb_ManNtkNum(p); i++ {
		pNtk = Acb_ManNtk(p, i)
		if i == nModules+1 {
			break
		}
		fmt.Printf("Module %5d : ", i)
		Acb_NtkPrintStats(pNtk)
	}
}

/*=== acbUtil.c =============================================================*/
// extern Vec_Int_t * Acb_ObjCollectTfi( Acb_Ntk_t * p, int iObj, int fTerm );

// extern Vec_Int_t * Acb_ObjCollectTfo( Acb_Ntk_t * p, int iObj, int fTerm );

// extern Vec_Int_t * Acb_ObjCollectTfiVec( Acb_Ntk_t * p, Vec_Int_t * vObjs );

// extern Vec_Int_t * Acb_ObjCollectTfoVec( Acb_Ntk_t * p, Vec_Int_t * vObjs );

// extern int         Acb_NtkCountPiBuffers( Acb_Ntk_t * p, Vec_Int_t * vObjs );

// extern int         Acb_NtkCountPoDrivers( Acb_Ntk_t * p, Vec_Int_t * vObjs );

// extern int         Acb_NtkFindMffcSize( Acb_Ntk_t * p, Vec_Int_t * vObjsRefed, Vec_Int_t * vObjsDerefed, int nGates[5] );

// extern int         Acb_ObjComputeLevelD( Acb_Ntk_t * p, int iObj );

// extern int         Acb_NtkComputeLevelD( Acb_Ntk_t * p, Vec_Int_t * vTfo );

// extern void        Acb_NtkUpdateLevelD( Acb_Ntk_t * p, int iObj );

// extern void        Acb_NtkUpdateTiming( Acb_Ntk_t * p, int iObj );

// extern void        Acb_NtkPrintNode( Acb_Ntk_t * p, int iObj );

// extern int         Acb_NtkCreateNode( Acb_Ntk_t * p, word uTruth, Vec_Int_t * vSupp );

// extern void        Acb_NtkUpdateNode( Acb_Ntk_t * p, int Pivot, word uTruth, Vec_Int_t * vSupp );

// extern Acb_Ntk_t * Acb_VerilogSimpleRead( char * pFileName, char * pFileNameW );

//ABC_NAMESPACE_HEADER_END

////////////////////////////////////////////////////////////////////////
///                       END OF FILE                                ///
////////////////////////////////////////////////////////////////////////
