package cba

import (
	"fmt"
	"strings"
	"unsafe"
)

import "os"

/**CFile****************************************************************

  FileName    [cbaPrs.h]

  SystemName  [ABC: Logic synthesis and verification system.]

  PackageName [Hierarchical word-level netlist.]

  Synopsis    [Parser declarations.]

  Author      [Alan Mishchenko]

  Affiliation [UC Berkeley]

  Date        [Ver. 1.0. Started - November 29, 2014.]

  Revision    [$Id: cbaPrs.h,v 1.00 2014/11/29 00:00:00 alanmi Exp $]

***********************************************************************/

////////////////////////////////////////////////////////////////////////
///                          INCLUDES                                ///
////////////////////////////////////////////////////////////////////////

//#include "aig/gia/gia.h"
//#include "misc/util/utilNam.h"

////////////////////////////////////////////////////////////////////////
///                         PARAMETERS                               ///
////////////////////////////////////////////////////////////////////////

//ABC_NAMESPACE_HEADER_START

// Verilog keywords
type Cba_VerType_t = int32

const (
	PRS_VER_NONE Cba_VerType_t = 0
	// 0:  unused
	PRS_VER_INPUT Cba_VerType_t = PRS_VER_NONE + 1
	// 1:  input
	PRS_VER_OUTPUT Cba_VerType_t = PRS_VER_INPUT + 1
	// 2:  output
	PRS_VER_INOUT Cba_VerType_t = PRS_VER_OUTPUT + 1
	// 3:  inout
	PRS_VER_WIRE Cba_VerType_t = PRS_VER_INOUT + 1
	// 4:  wire
	PRS_VER_REG Cba_VerType_t = PRS_VER_WIRE + 1
	// 5:  reg
	PRS_VER_MODULE Cba_VerType_t = PRS_VER_REG + 1
	// 6:  module
	PRS_VER_ASSIGN Cba_VerType_t = PRS_VER_MODULE + 1
	// 7:  assign
	PRS_VER_ALWAYS Cba_VerType_t = PRS_VER_ASSIGN + 1
	// 8:  always
	PRS_VER_FUNCTION Cba_VerType_t = PRS_VER_ALWAYS + 1
	// 9:  function
	PRS_VER_DEFPARAM Cba_VerType_t = PRS_VER_FUNCTION + 1
	// 10: defparam
	PRS_VER_BEGIN Cba_VerType_t = PRS_VER_DEFPARAM + 1
	// 11: begin
	PRS_VER_END Cba_VerType_t = PRS_VER_BEGIN + 1
	// 12: end
	PRS_VER_CASE Cba_VerType_t = PRS_VER_END + 1
	// 13: case
	PRS_VER_ENDCASE Cba_VerType_t = PRS_VER_CASE + 1
	// 14: endcase
	PRS_VER_SIGNED Cba_VerType_t = PRS_VER_ENDCASE + 1
	// 15: signed
	PRS_VER_ENDMODULE Cba_VerType_t = PRS_VER_SIGNED + 1
	// 16: endmodule
	PRS_VER_UNKNOWN Cba_VerType_t = PRS_VER_ENDMODULE + 1 // 17: unknown

)

// parser name types
type Prs_ManType_t = int32

const (
	CBA_PRS_NAME Prs_ManType_t = 0
	// 0:  name/variable
	CBA_PRS_SLICE Prs_ManType_t = CBA_PRS_NAME + 1
	// 1:  slice
	CBA_PRS_CONST Prs_ManType_t = CBA_PRS_SLICE + 1
	// 2:  constant
	CBA_PRS_CONCAT Prs_ManType_t = CBA_PRS_CONST + 1
	// 3:  concatentation

)

////////////////////////////////////////////////////////////////////////
///                         BASIC TYPES                              ///
////////////////////////////////////////////////////////////////////////

// network

type Prs_Ntk_t struct {
	// general info
	iModuleName int32
	fMapped     int32 /* C bitfield: 1 bits; review packing and truncation */
	fSlices     int32 /* C bitfield: 1 bits; review packing and truncation */
	fHasC0s     int32 /* C bitfield: 1 bits; review packing and truncation */
	fHasC1s     int32 /* C bitfield: 1 bits; review packing and truncation */
	fHasCXs     int32 /* C bitfield: 1 bits; review packing and truncation */
	fHasCZs     int32 /* C bitfield: 1 bits; review packing and truncation */
	pStrs       *Abc_Nam_t
	pFuns       *Abc_Nam_t
	vHash       *Hash_IntMan_t
	// interface
	vOrder Vec_Int_t // order of signals
	// signal names
	vInouts  Vec_Int_t // inouts
	vInputs  Vec_Int_t // inputs
	vOutputs Vec_Int_t // outputs
	vWires   Vec_Int_t // wires
	// signal ranges
	vInoutsR  Vec_Int_t // inouts
	vInputsR  Vec_Int_t // inputs
	vOutputsR Vec_Int_t // outputs
	vWiresR   Vec_Int_t // wires
	// slices/concatenations/objects
	vSlices  Vec_Int_t // NameId + RangeId
	vConcats Vec_Int_t // array of NameId/SliceId/ConstId
	vBoxes   Vec_Int_t // ModuleId + InstId + array of pairs {FormNameId, ActSignalId(NameId/SliceId/ConstId/ConcatId)}
	vObjs    Vec_Int_t // box handles
}

// parser

type Prs_Man_t struct {
	// input data
	pName   string         // file name
	pBuffer string         // file contents
	pLimit  int            // end offset in input buffer
	pCur    int            // current byte offset
	pStrs   *Abc_Nam_t     // string manager
	pFuns   *Abc_Nam_t     // cover manager
	vHash   *Hash_IntMan_t // variable ranges
	pNtk    *Prs_Ntk_t     // current network
	vNtks   *Vec_Ptr_t     // input networks
	// temporary data
	vCover Vec_Str_t // one SOP cover
	vTemp  Vec_Int_t // array of tokens
	vTemp2 Vec_Int_t // array of tokens
	vTemp3 Vec_Int_t // array of tokens
	vTemp4 Vec_Int_t // array of tokens
	// statistics
	vKnown     Vec_Int_t
	vFailed    Vec_Int_t
	vSucceeded Vec_Int_t
	// error handling
	nOpens      int32      // open port counter
	fUsingTemp2 int32      // vTemp2 is in use
	FuncNameId  int32      // temp value
	FuncRangeId int32      // temp value
	ErrorStr    [1000]byte // error
}

func Prs_ManNtk(vPrs *Vec_Ptr_t, i int32) *Prs_Ntk_t {
	if i >= 0 && i < Vec_PtrSize(vPrs) {
		p, _ := Vec_PtrEntry(vPrs, i).(*Prs_Ntk_t)
		return p
	} else {
		return nil
	}
}
func Prs_ManRoot(vPrs *Vec_Ptr_t) *Prs_Ntk_t    { return Prs_ManNtk(vPrs, 0) }
func Prs_ManNameMan(vPrs *Vec_Ptr_t) *Abc_Nam_t { return Prs_ManRoot(vPrs).pStrs }
func Prs_ManFuncMan(vPrs *Vec_Ptr_t) *Abc_Nam_t { return Prs_ManRoot(vPrs).pFuns }

func Prs_NtkId(p *Prs_Ntk_t) int32     { return p.iModuleName }
func Prs_NtkPioNum(p *Prs_Ntk_t) int32 { return Vec_IntSize(&p.vInouts) }
func Prs_NtkPiNum(p *Prs_Ntk_t) int32  { return Vec_IntSize(&p.vInputs) }
func Prs_NtkPoNum(p *Prs_Ntk_t) int32  { return Vec_IntSize(&p.vOutputs) }
func Prs_NtkBoxNum(p *Prs_Ntk_t) int32 { return Vec_IntSize(&p.vObjs) }
func Prs_NtkObjNum(p *Prs_Ntk_t) int32 {
	return Prs_NtkPioNum(p) + Prs_NtkPiNum(p) + Prs_NtkPoNum(p) + Prs_NtkBoxNum(p)
}
func Prs_NtkStr(p *Prs_Ntk_t, h int32) string   { return Abc_NamStr(p.pStrs, h) }
func Prs_NtkSop(p *Prs_Ntk_t, h int32) string   { return Abc_NamStr(p.pFuns, h) }
func Prs_NtkConst(p *Prs_Ntk_t, h int32) string { return Abc_NamStr(p.pFuns, h) }
func Prs_NtkName(p *Prs_Ntk_t) string           { return Prs_NtkStr(p, Prs_NtkId(p)) }
func Prs_NtkSigName(p *Prs_Ntk_t, i int32) int32 {
	if p.fSlices == 0 {
		return i
	}
	assert(Abc_Lit2Att2(i) == int32(CBA_PRS_NAME), "Abc_Lit2Att2(i) == int32(CBA_PRS_NAME)")
	return Abc_Lit2Var2(i)
}
func Ptr_NtkRangeSize(p *Prs_Ntk_t, h int32) int32 {
	var l int32 = Hash_IntObjData0(p.vHash, h)
	var r int32 = Hash_IntObjData1(p.vHash, h)
	if l > r {
		return 1 + l - r
	}
	return 1 + r - l
}

func Prs_SliceName(p *Prs_Ntk_t, h int32) int32  { return Vec_IntEntry(&p.vSlices, h) }
func Prs_SliceRange(p *Prs_Ntk_t, h int32) int32 { return Vec_IntEntry(&p.vSlices, h+1) }

func Prs_CatSize(p *Prs_Ntk_t, h int32) int32   { return Vec_IntEntry(&p.vConcats, h) }
func Prs_CatArray(p *Prs_Ntk_t, h int32) *int32 { return Vec_IntEntryP(&p.vConcats, h+1) }

var spStatic_Prs_CatSignals_V Vec_Int_t

func Prs_CatSignals(p *Prs_Ntk_t, h int32) *Vec_Int_t {
	Vec_IntClear(&spStatic_Prs_CatSignals_V)
	start := h + 1
	for i := int32(0); i < Prs_CatSize(p, h); i++ {
		Vec_IntPush(&spStatic_Prs_CatSignals_V, Vec_IntEntry(&p.vConcats, start+i))
	}
	return &spStatic_Prs_CatSignals_V
}

func Prs_BoxHand(p *Prs_Ntk_t, i int32) int32  { return Vec_IntEntry(&p.vObjs, i) }
func Prs_BoxSize(p *Prs_Ntk_t, i int32) int32  { return Vec_IntEntry(&p.vBoxes, Prs_BoxHand(p, i)) - 2 }
func Prs_BoxIONum(p *Prs_Ntk_t, i int32) int32 { return Prs_BoxSize(p, i) / 2 }
func Prs_BoxNtk(p *Prs_Ntk_t, i int32) int32   { return Vec_IntEntry(&p.vBoxes, Prs_BoxHand(p, i)+1) }
func Prs_BoxSetNtk(p *Prs_Ntk_t, i int32, m int32) {
	Vec_IntWriteEntry(&p.vBoxes, Prs_BoxHand(p, i)+1, m)
}
func Prs_BoxName(p *Prs_Ntk_t, i int32) int32 { return Vec_IntEntry(&p.vBoxes, Prs_BoxHand(p, i)+2) }
func Prs_BoxIsNode(p *Prs_Ntk_t, i int32) int32 {
	return boolToInt(Vec_IntEntry(&p.vBoxes, Prs_BoxHand(p, i)+3) == 0)
}                                               // no formal names
func Prs_BoxArray(p *Prs_Ntk_t, i int32) *int32 { return Vec_IntEntryP(&p.vBoxes, Prs_BoxHand(p, i)+3) }

var spStatic_Prs_BoxSignals_V Vec_Int_t

func Prs_BoxSignals(p *Prs_Ntk_t, i int32) *Vec_Int_t {
	Vec_IntClear(&spStatic_Prs_BoxSignals_V)
	start := Prs_BoxHand(p, i) + 3
	for k := int32(0); k < Prs_BoxSize(p, i); k++ {
		Vec_IntPush(&spStatic_Prs_BoxSignals_V, Vec_IntEntry(&p.vBoxes, start+k))
	}
	return &spStatic_Prs_BoxSignals_V
}

// Original macro:
// #define Prs_ManForEachNameVec( vVec, p, pName, i )   \
//     for ( i = 0; (i < Vec_IntSize(vVec)) && ((pName) = Abc_NamStr(p->pStrs, Vec_IntEntry(vVec,i))); i++ )

// Original macro:
// #define Prs_NtkForEachPio( p, NameId, i )            \
//     for ( i = 0; i < Prs_NtkPioNum(p) && ((NameId) = Vec_IntEntry(&p->vInouts, i)); i++ )

// Original macro:
// #define Prs_NtkForEachPi( p, NameId, i )             \
//     for ( i = 0; i < Prs_NtkPiNum(p) && ((NameId) = Vec_IntEntry(&p->vInputs, i)); i++ )

// Original macro:
// #define Prs_NtkForEachPo( p, NameId, i )             \
//     for ( i = 0; i < Prs_NtkPoNum(p) && ((NameId) = Vec_IntEntry(&p->vOutputs, i)); i++ )

// Original macro:
// #define Prs_NtkForEachBox( p, vVec, i )              \
//     for ( i = 0; i < Prs_NtkBoxNum(p) && ((vVec) = Prs_BoxSignals(p, i)); i++ )

////////////////////////////////////////////////////////////////////////
///                      MACRO DEFINITIONS                           ///
////////////////////////////////////////////////////////////////////////

// create error message
func Prs_ManErrorSet(p *Prs_Man_t, pError string, Value int32) int32 {
	assert(p.ErrorStr[0] == 0, "p.ErrorStr[0] == 0")
	copy(p.ErrorStr[:], pError)
	return Value
}

// clear error message
func Prs_ManErrorClear(p *Prs_Man_t) {
	p.ErrorStr[0] = 0
}

// print error message
func Prs_ManErrorPrint(p *Prs_Man_t) int32 {
	var iLine int32
	if p.ErrorStr[0] == 0 {
		return 1
	}
	if p.pCur <= len(p.pBuffer) {
		iLine = int32(strings.Count(p.pBuffer[:p.pCur], "\n"))
	}
	fmt.Printf("Line %d: %s\n", iLine, p.ErrorStr)
	return 0
}

// parsing network
func Prs_ManInitializeNtk(p *Prs_Man_t, iName int32, fSlices int32) {
	assert(p.pNtk == nil, "p.pNtk == nil")
	// inlining of macro ABC_CALLOC
	p.pNtk = &Prs_Ntk_t{}
	p.pNtk.iModuleName = iName
	p.pNtk.fSlices = fSlices
	p.pNtk.pStrs = Abc_NamRef(p.pStrs)
	p.pNtk.pFuns = Abc_NamRef(p.pFuns)
	p.pNtk.vHash = Hash_IntManRef(p.vHash)
	Vec_PtrPush(p.vNtks, p.pNtk)
}
func Prs_ManFinalizeNtk(p *Prs_Man_t) {
	assert(p.pNtk != nil, "p.pNtk != nil")
	p.pNtk = nil
}

func Prs_ManNewStrId(p *Prs_Man_t, format string, args ...any) int32 {
	return Abc_NamStrFindOrAdd(p.pStrs, fmt.Sprintf(format, args...), nil)
}

// parsing slice/concatentation/box
func Prs_NtkAddSlice(p *Prs_Ntk_t, Name int32, Range int32) int32 {
	var Value int32 = Vec_IntSize(&p.vSlices)
	Vec_IntPushTwo(&p.vSlices, Name, Range)
	return Value
}
func Prs_NtkAddConcat(p *Prs_Ntk_t, vTemp *Vec_Int_t) int32 {
	var Value int32
	if Vec_IntSize(&p.vConcats)&1 == 0 {
		Vec_IntPush(&p.vConcats, -1)
	}
	Value = Vec_IntSize(&p.vConcats)
	assert(Value&1 != 0, "Value & 1")
	Vec_IntPush(&p.vConcats, Vec_IntSize(vTemp))
	Vec_IntAppend(&p.vConcats, vTemp)
	return Value
}
func Prs_NtkAddBox(p *Prs_Ntk_t, ModName int32, InstName int32, vTemp *Vec_Int_t) {
	var Value int32
	assert(Vec_IntSize(vTemp)%2 == 0, "Vec_IntSize(vTemp) % 2 == 0")
	if Vec_IntSize(&p.vBoxes)&1 == 0 {
		Vec_IntPush(&p.vBoxes, -1)
	}
	Value = Vec_IntSize(&p.vBoxes)
	assert(Value&1 != 0, "Value & 1")
	Vec_IntPush(&p.vObjs, Value)
	// create entry
	Vec_IntPush(&p.vBoxes, Vec_IntSize(vTemp)+2)
	Vec_IntPush(&p.vBoxes, ModName)
	Vec_IntPush(&p.vBoxes, InstName)
	Vec_IntAppend(&p.vBoxes, vTemp)
}
func Prs_ManLoadFile(pFileName string) (string, int) {
	data, err := os.ReadFile(pFileName)
	if err != nil {
		fmt.Printf("Cannot open input file.\n")
		return "", 0
	}
	buffer := "\n" + string(data) + "\n\x00"
	return buffer, len(buffer)
}
func Prs_ManAlloc(pFileName string) *Prs_Man_t {

	// inlining of macro ABC_CALLOC
	p := &Prs_Man_t{}
	if pFileName != "" {
		var pBuffer string
		var pLimit int
		pBuffer, pLimit = Prs_ManLoadFile(pFileName)
		if pBuffer == "" {
			return nil
		}
		p.pName = pFileName
		p.pBuffer = pBuffer
		p.pLimit = pLimit
		p.pCur = 0
	}
	p.pStrs = Abc_NamStart(1000, 24)
	p.pFuns = Abc_NamStart(100, 24)
	p.vHash = Hash_IntManStart(1000)
	p.vNtks = Vec_PtrAlloc(100)
	return p
}

func Prs_NtkFree(p *Prs_Ntk_t) {
	if p.pStrs != nil {
		Abc_NamDeref(p.pStrs)
	}
	if p.pFuns != nil {
		Abc_NamDeref(p.pFuns)
	}
	if p.vHash != nil {
		Hash_IntManDeref(p.vHash)
	}
	Vec_IntErase(&p.vOrder)
	Vec_IntErase(&p.vInouts)
	Vec_IntErase(&p.vInputs)
	Vec_IntErase(&p.vOutputs)
	Vec_IntErase(&p.vWires)
	Vec_IntErase(&p.vInoutsR)
	Vec_IntErase(&p.vInputsR)
	Vec_IntErase(&p.vOutputsR)
	Vec_IntErase(&p.vWiresR)
	Vec_IntErase(&p.vSlices)
	Vec_IntErase(&p.vConcats)
	Vec_IntErase(&p.vBoxes)
	Vec_IntErase(&p.vObjs)
	p = nil
}

func Prs_ManVecFree(vPrs *Vec_Ptr_t) {
	var pNtk *Prs_Ntk_t
	var i int32
	// inlining of macro Vec_PtrForEachEntry
	for i = 0; i < Vec_PtrSize(vPrs); i++ {
		pNtk, _ = Vec_PtrEntry(vPrs, i).(*Prs_Ntk_t)
		Prs_NtkFree(pNtk)
	}
	Vec_PtrFree(vPrs)
}

func Prs_ManFree(p *Prs_Man_t) {
	if p.pStrs != nil {
		Abc_NamDeref(p.pStrs)
	}
	if p.pFuns != nil {
		Abc_NamDeref(p.pFuns)
	}
	if p.vHash != nil {
		Hash_IntManDeref(p.vHash)
	}
	if p.vNtks != nil {
		Prs_ManVecFree(p.vNtks)
	}
	// temporary
	Vec_StrErase(&p.vCover)
	Vec_IntErase(&p.vTemp)
	Vec_IntErase(&p.vTemp2)
	Vec_IntErase(&p.vTemp3)
	Vec_IntErase(&p.vTemp4)
	Vec_IntErase(&p.vKnown)
	Vec_IntErase(&p.vFailed)
	Vec_IntErase(&p.vSucceeded)
	p.pBuffer = ""
	p = nil
}

func Prs_NtkMemory(p *Prs_Ntk_t) int32 {
	var nMem int32 = int32(unsafe.Sizeof(Prs_Ntk_t{}))
	nMem += int32(Vec_IntMemory(&p.vOrder))
	nMem += int32(Vec_IntMemory(&p.vInouts))
	nMem += int32(Vec_IntMemory(&p.vInputs))
	nMem += int32(Vec_IntMemory(&p.vOutputs))
	nMem += int32(Vec_IntMemory(&p.vWires))
	nMem += int32(Vec_IntMemory(&p.vInoutsR))
	nMem += int32(Vec_IntMemory(&p.vInputsR))
	nMem += int32(Vec_IntMemory(&p.vOutputsR))
	nMem += int32(Vec_IntMemory(&p.vWiresR))
	nMem += int32(Vec_IntMemory(&p.vSlices))
	nMem += int32(Vec_IntMemory(&p.vBoxes))
	nMem += int32(Vec_IntMemory(&p.vConcats))
	return nMem
}
func Prs_ManMemory(vPrs *Vec_Ptr_t) int32 {
	var pNtk *Prs_Ntk_t
	var i int32
	var nMem int32 = int32(Vec_PtrMemory(vPrs))
	// inlining of macro Vec_PtrForEachEntry
	for i = 0; i < Vec_PtrSize(vPrs); i++ {
		pNtk, _ = Vec_PtrEntry(vPrs, i).(*Prs_Ntk_t)
		nMem += Prs_NtkMemory(pNtk)
	}
	nMem += Abc_NamMemUsed(Prs_ManNameMan(vPrs))
	return nMem
}

/**Function*************************************************************

  Synopsis    [Other APIs.]

  Description []

  SideEffects []

  SeeAlso     []

***********************************************************************/

////////////////////////////////////////////////////////////////////////
///                             ITERATORS                            ///
////////////////////////////////////////////////////////////////////////

////////////////////////////////////////////////////////////////////////
///                    FUNCTION DECLARATIONS                         ///
////////////////////////////////////////////////////////////////////////

/*=== cbaReadVer.c ========================================================*/
// extern void Prs_NtkAddVerilogDirectives( Prs_Man_t * p );

//ABC_NAMESPACE_HEADER_END

////////////////////////////////////////////////////////////////////////
///                       END OF FILE                                ///
////////////////////////////////////////////////////////////////////////
