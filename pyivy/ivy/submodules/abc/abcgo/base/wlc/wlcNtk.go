package wlc

import (
	"fmt"
	"math"
	"os"
	"unsafe"
)

/**CFile****************************************************************

  FileName    [wlcNtk.c]

  SystemName  [ABC: Logic synthesis and verification system.]

  PackageName [Verilog parser.]

  Synopsis    [Network data-structure.]

  Author      [Alan Mishchenko]

  Affiliation [UC Berkeley]

  Date        [Ver. 1.0. Started - August 22, 2014.]

  Revision    [$Id: wlcNtk.c,v 1.00 2014/09/12 00:00:00 alanmi Exp $]

***********************************************************************/

//#include <math.h>
//#include "wlc.h"
//#include "misc/vec/vecWec.h"

//ABC_NAMESPACE_IMPL_START

////////////////////////////////////////////////////////////////////////
///                        DECLARATIONS                              ///
////////////////////////////////////////////////////////////////////////

// object types
var Wlc_Names = [...]string{
	"",       // 00: unknown
	"pi",     // 01: primary input
	"po",     // 02: primary output (unused)
	"ff",     // 03: flop output
	"bi",     // 04: flop input (unused)
	"ff",     // 05: flop (unused)
	"const",  // 06: constant
	"buf",    // 07: buffer
	"mux",    // 08: multiplexer
	">>",     // 09: shift right
	">>>",    // 10: shift right (arithmetic)
	"<<",     // 11: shift left
	"<<<",    // 12: shift left (arithmetic)
	"rotR",   // 13: rotate right
	"rotL",   // 14: rotate left
	"~",      // 15: bitwise NOT
	"&",      // 16: bitwise AND
	"|",      // 17: bitwise OR
	"^",      // 18: bitwise XOR
	"~&",     // 19: bitwise NAND
	"~|",     // 20: bitwise NOR
	"~^",     // 21: bitwise NXOR
	"[:]",    // 22: bit selection
	"{,}",    // 23: bit concatenation
	"zPad",   // 24: zero padding
	"sExt",   // 25: sign extension
	"!",      // 26: logic NOT
	"=>",     // 27: logic implication
	"&&",     // 28: logic AND
	"||",     // 29: logic OR
	"^^",     // 30: logic XOR
	"==",     // 31: compare equal
	"!=",     // 32: compare not equal
	"<",      // 33: compare less
	">",      // 34: compare more
	"<=",     // 35: compare less or equal
	">=",     // 36: compare more or equal
	"&",      // 37: reduction AND
	"|",      // 38: reduction OR
	"^",      // 39: reduction XOR
	"~&",     // 40: reduction NAND
	"~|",     // 41: reduction NOR
	"~^",     // 42: reduction NXOR
	"+",      // 43: arithmetic addition
	"-",      // 44: arithmetic subtraction
	"*",      // 45: arithmetic multiplier
	"/",      // 46: arithmetic division
	"%",      // 47: arithmetic reminder
	"mod",    // 48: arithmetic modulus
	"**",     // 49: arithmetic power
	"-",      // 50: arithmetic minus
	"sqrt",   // 51: integer square root
	"squar",  // 52: integer square
	"table",  // 53: bit table
	"READ",   // 54: mem read port
	"WRITE",  // 55: mem write port
	"addsub", // 56: adder/subtractor
	"sel",    // 57: selector
	"dec",    // 58: decoder
	"LUT",    // 59: lookup table
	"",       // 58: unused
}

func Wlc_ObjTypeName(p *Wlc_Obj_t) string {
	if p != nil {
		if p.Type < uint32(WLC_OBJ_NUMBER) {
			return Wlc_Names[p.Type]
		} else {
			return "out_of_bound"
		}
	} else {
		return "no_obj"
	}
}

////////////////////////////////////////////////////////////////////////
///                     FUNCTION DEFINITIONS                         ///
////////////////////////////////////////////////////////////////////////

/*
*Function*************************************************************

	Synopsis    []

	Description []

	SideEffects []

	SeeAlso     []

**********************************************************************
*/
func Wlc_ManSetDefaultParams(pPars *Wlc_Par_t) {
	*pPars = Wlc_Par_t{}
	pPars.nBitsAdd = ABC_INFINITY  // adder bit-width
	pPars.nBitsMul = ABC_INFINITY  // multiplier bit-widht
	pPars.nBitsMux = ABC_INFINITY  // MUX bit-width
	pPars.nBitsFlop = ABC_INFINITY // flop bit-width
	pPars.nIterMax = 1000          // the max number of iterations
	pPars.nLimit = ABC_INFINITY    // the max number of signals
	pPars.fXorOutput = 1           // XOR outputs of word-level miter
	pPars.fCheckClauses = 1        // Check clauses in the reloaded trace
	pPars.fPushClauses = 0         // Push clauses in the reloaded trace
	pPars.fMFFC = 1                // Refine the entire MFFC of a PPI
	pPars.fPdra = 0                // Use pdr -nct
	pPars.fLoadTrace = 1           // Load previous PDR traces
	pPars.fProofRefine = 0         // Use proof-based refinement
	pPars.fHybrid = 1              // Use a hybrid of CBR and PBR
	pPars.fCheckCombUnsat = 0      // Check if ABS becomes comb. unsat
	pPars.fAbs2 = 0                // Use UFAR style createAbs
	pPars.fProofUsePPI = 0         // Use PPI values in PBR
	pPars.fUseBmc3 = 0             // Run BMC3 in parallel
	pPars.fShrinkAbs = 0           // Shrink Abs with BMC
	pPars.fShrinkScratch = 0       // Restart pdr from scratch after shrinking
	pPars.fVerbose = 0             // verbose output`
	pPars.fPdrVerbose = 0          // show verbose PDR output
}

/*
*Function*************************************************************

	Synopsis    [Working with models.]

	Description []

	SideEffects []

	SeeAlso     []

**********************************************************************
*/
func Wlc_NtkAlloc(pName string, nObjsAlloc int32) *Wlc_Ntk_t {
	var p *Wlc_Ntk_t
	p = new(Wlc_Ntk_t)
	if pName != "" {
		p.pName = Extra_FileNameGeneric(pName)
	} else {
		p.pName = ""
	}
	Vec_IntGrow(&p.vPis, 111)
	Vec_IntGrow(&p.vPos, 111)
	Vec_IntGrow(&p.vCis, 111)
	Vec_IntGrow(&p.vCos, 111)
	Vec_IntGrow(&p.vFfs, 111)
	p.pMemFanin = Mem_FlexStart()
	p.nObjsAlloc = nObjsAlloc
	p.objStore = make([]Wlc_Obj_t, p.nObjsAlloc)
	p.pObjs = &p.objStore[0]
	p.iObj = 1
	return p
}
func Wlc_ObjSetCi(p *Wlc_Ntk_t, pObj *Wlc_Obj_t) {
	assert(Wlc_ObjIsCi(pObj) != 0, "Wlc_ObjIsCi(pObj)")
	assert(Wlc_ObjFaninNum(pObj) == 0, "Wlc_ObjFaninNum(pObj) == 0")
	if Wlc_NtkPiNum(p) == Wlc_NtkCiNum(p) || pObj.Type != uint32(WLC_OBJ_PI) {
		pObj.Fanins[1] = Vec_IntSize(&p.vCis)
		Vec_IntPush(&p.vCis, Wlc_ObjId(p, pObj))

	} else { // insert in the array of CI at the end of PIs

		var pTemp *Wlc_Obj_t
		var i int32
		Vec_IntInsert(&p.vCis, Wlc_NtkPiNum(p), Wlc_ObjId(p, pObj))
		// other CI IDs are invalidated... naive fix!
		// inlining of macro Wlc_NtkForEachCi
		for i = 0; i < Wlc_NtkCiNum(p); i++ {
			pTemp = Wlc_NtkCi(p, i)
			pTemp.Fanins[1] = i
		}
	}
	if pObj.Type == uint32(WLC_OBJ_PI) {
		Vec_IntPush(&p.vPis, Wlc_ObjId(p, pObj))
	}
}
func Wlc_ObjSetCo(p *Wlc_Ntk_t, pObj *Wlc_Obj_t, fFlopInput int32) {
	//    pObj->Fanins[1] = Vec_IntSize(&p->vCos);
	Vec_IntPush(&p.vCos, Wlc_ObjId(p, pObj))
	if !(fFlopInput != 0) {
		Vec_IntPush(&p.vPos, Wlc_ObjId(p, pObj))
	}
	if fFlopInput != 0 {
		pObj.fIsFi = 1
	} else {
		pObj.fIsPo = 1
	}
}
func Wlc_ObjAlloc(p *Wlc_Ntk_t, Type int32, Signed int32, End int32, Beg int32) int32 {
	var pObj *Wlc_Obj_t
	assert(Type != int32(WLC_OBJ_PO) && Type != int32(WLC_OBJ_FI), "Type != WLC_OBJ_PO && Type != WLC_OBJ_FI")
	if p.iObj == p.nObjsAlloc {
		p.objStore = append(p.objStore, make([]Wlc_Obj_t, p.nObjsAlloc)...)
		p.pObjs = &p.objStore[0]
		p.nObjsAlloc *= 2
	}
	pObj = Wlc_NtkObj(p, p.iObj)
	pObj.Type = uint32(Type)
	pObj.Signed = uint32(Signed)
	pObj.End = End
	pObj.Beg = Beg
	if Wlc_ObjIsCi(pObj) != 0 {
		Wlc_ObjSetCi(p, pObj)
	}
	p.nObjs[Type]++
	id := p.iObj
	p.iObj++
	return id
}
func Wlc_ObjCreate(p *Wlc_Ntk_t, Type int32, Signed int32, End int32, Beg int32, vFanins *Vec_Int_t) int32 {
	var iFaninNew int32 = Wlc_ObjAlloc(p, Type, Signed, End, Beg)
	Wlc_ObjAddFanins(p, Wlc_NtkObj(p, iFaninNew), vFanins)
	return iFaninNew
}

var spStatic_Wlc_ObjName_Buffer string // was [100]byte in C

func Wlc_ObjName(p *Wlc_Ntk_t, iObj int32) string {

	if Wlc_NtkHasNameId(p) != 0 && Wlc_ObjNameId(p, iObj) != 0 {
		return Abc_NamStr(p.pManName, Wlc_ObjNameId(p, iObj))
	}
	spStatic_Wlc_ObjName_Buffer = fmt.Sprintf("n%d", iObj)
	return spStatic_Wlc_ObjName_Buffer
}
func Wlc_ObjUpdateType(p *Wlc_Ntk_t, pObj *Wlc_Obj_t, Type int32) {
	assert(pObj.Type == uint32(WLC_OBJ_NONE), "pObj.Type == uint32(WLC_OBJ_NONE)")
	p.nObjs[pObj.Type]--
	pObj.Type = uint32(Type)
	p.nObjs[pObj.Type]++
}
func Wlc_ObjAddFanins(p *Wlc_Ntk_t, pObj *Wlc_Obj_t, vFanins *Vec_Int_t) {
	assert(pObj.nFanins == 0, "pObj.nFanins == 0")
	pObj.nFanins = uint32(Vec_IntSize(vFanins))
	// special treatment of CONST, SELECT and TABLE
	if pObj.Type == uint32(WLC_OBJ_CONST) {
		pObj.nFanins = 0
	} else if pObj.Type == uint32(WLC_OBJ_BIT_SELECT) || pObj.Type == uint32(WLC_OBJ_TABLE) {
		pObj.nFanins = 1
	}
	// C stores the full constant bit vector and bit-select operands in the
	// object's flexible fanin allocation even when nFanins is reset to zero
	// (CONST) or one (BIT_SELECT). Keep the backing data on the Go object so
	// Wlc_ObjFanins and Wlc_ObjConstValue see the same words after parsing.
	if Wlc_ObjHasArray(pObj) != 0 {
		pObj.pFanins[0] = append(pObj.pFanins[0][:0], Vec_IntArray(vFanins)...)
	}
	copy(Wlc_ObjFanins(pObj), Vec_IntArray(vFanins))
}
func Wlc_NtkFree(p *Wlc_Ntk_t) {
	if p.pManName != nil {
		Abc_NamStop(p.pManName)
	}
	if p.pMemFanin != nil {
		Mem_FlexStop(p.pMemFanin, 0)
	}
	if p.pMemTable != nil {
		Mem_FlexStop(p.pMemTable, 0)
	}
	Vec_IntClear(&p.vPoPairs)
	Vec_PtrFreeP(&p.vTables)
	Vec_WrdFreeP(&p.vLutTruths)
	Vec_IntClear(&p.vPis)
	Vec_IntClear(&p.vPos)
	Vec_IntClear(&p.vCis)
	Vec_IntClear(&p.vCos)
	Vec_IntClear(&p.vFfs)
	Vec_IntClear(&p.vFfs2)
	Vec_IntFreeP(&p.vInits)
	Vec_IntFreeP(&p.vArsts)
	Vec_IntClear(&p.vTravIds)
	Vec_IntClear(&p.vNameIds)
	Vec_IntClear(&p.vValues)
	Vec_IntClear(&p.vCopies)
	Vec_IntClear(&p.vBits)
	Vec_IntClear(&p.vLevels)
	Vec_IntClear(&p.vRefs)
	p.pInits = ""
	p.pObjs = nil
	p.objStore = nil
	p.pName = ""
	p.pSpec = ""
}
func Wlc_NtkMemUsage(p *Wlc_Ntk_t) int32 {
	memBytes := int32(unsafe.Sizeof(Wlc_Ntk_t{}))
	memBytes += 4 * int32(Vec_IntCap(&p.vPis)+Vec_IntCap(&p.vPos)+Vec_IntCap(&p.vCis)+Vec_IntCap(&p.vCos)+Vec_IntCap(&p.vFfs)+Vec_IntCap(&p.vFfs2))
	memBytes += int32(unsafe.Sizeof(Wlc_Obj_t{})) * p.nObjsAlloc
	memBytes += Abc_NamMemUsed(p.pManName)
	memBytes += Mem_FlexReadMemUsage(p.pMemFanin)
	return memBytes
}

/*
*Function*************************************************************

	Synopsis    [Assigns object levels.]

	Description []

	SideEffects []

	SeeAlso     []

**********************************************************************
*/
func Wlc_NtkCreateLevels_(p *Wlc_Ntk_t) int32 {
	var pObj *Wlc_Obj_t
	var i, k, iFanin, Level, LevelMax int32
	Vec_IntFill(&p.vLevels, Wlc_NtkObjNumMax(p), 0)
	// inlining of macro Wlc_NtkForEachObj
	for i = 1; i < Wlc_NtkObjNumMax(p); i++ {
		pObj = Wlc_NtkObj(p, i)
		Level = 0
		// inlining of macro Wlc_ObjForEachFanin
		for k = 0; k < Wlc_ObjFaninNum(pObj); k++ {
			iFanin = Wlc_ObjFaninId(pObj, k)
			Level = Abc_MaxInt(Level, Wlc_ObjLevelId(p, iFanin)+1)
		}
		Vec_IntWriteEntry(&p.vLevels, i, Level)
		LevelMax = Abc_MaxInt(LevelMax, Level)
	}
	return LevelMax
}
func Wlc_NtkCreateLevelsRev(p *Wlc_Ntk_t) int32 {
	var pObj *Wlc_Obj_t
	var i, k, iFanin, Level, LevelMax int32
	Vec_IntFill(&p.vLevels, Wlc_NtkObjNumMax(p), 0)
	// inlining of macro Wlc_NtkForEachObjReverse
	for i = Wlc_NtkObjNumMax(p) - 1; i > 0; i-- {
		pObj = Wlc_NtkObj(p, i)
		if Wlc_ObjIsCi(pObj) != 0 {
			continue
		}
		Level = Wlc_ObjLevel(p, pObj) + 1
		// inlining of macro Wlc_ObjForEachFanin
		for k = 0; k < Wlc_ObjFaninNum(pObj); k++ {
			iFanin = Wlc_ObjFaninId(pObj, k)
			Vec_IntUpdateEntry(&p.vLevels, iFanin, Level)
		}
		LevelMax = Abc_MaxInt(LevelMax, Level)
	}
	// reverse the values
	// inlining of macro Wlc_NtkForEachObj
	for i = 1; i < Wlc_NtkObjNumMax(p); i++ {
		pObj = Wlc_NtkObj(p, i)
		Vec_IntWriteEntry(&p.vLevels, i, LevelMax-Wlc_ObjLevelId(p, i))
	}
	// inlining of macro Wlc_NtkForEachCi
	for i = 0; i < Wlc_NtkCiNum(p); i++ {
		pObj = Wlc_NtkCi(p, i)
		Vec_IntWriteEntry(&p.vLevels, Wlc_ObjId(p, pObj), 0)
	}
	//Wlc_NtkForEachObj( p, pObj, i )
	//    printf( "%d -> %d\n", i, Wlc_ObjLevelId(p, i) );
	return LevelMax
}

/*
*Function*************************************************************

	Synopsis    [Assigns object levels.]

	Description []

	SideEffects []

	SeeAlso     []

**********************************************************************
*/
func Wlc_NtkCreateLevels_rec(p *Wlc_Ntk_t, pObj *Wlc_Obj_t) {
	var k int32
	var iFanin int32
	var Level int32
	var iObj int32 = Wlc_ObjId(p, pObj)
	if Wlc_ObjIsCi(pObj) != 0 || Wlc_ObjIsFf(p, iObj) != 0 || Wlc_ObjFaninNum(pObj) == 0 || Wlc_ObjLevel(p, pObj) > 0 {
		return
	}
	// inlining of macro Wlc_ObjForEachFanin
	for k = 0; k < Wlc_ObjFaninNum(pObj); k++ {
		iFanin = Wlc_ObjFaninId(pObj, k)
		if iFanin != 0 {
			Wlc_NtkCreateLevels_rec(p, Wlc_NtkObj(p, iFanin))
		}
	}
	// inlining of macro Wlc_ObjForEachFanin
	for k = 0; k < Wlc_ObjFaninNum(pObj); k++ {
		iFanin = Wlc_ObjFaninId(pObj, k)
		if iFanin != 0 {
			Level = Abc_MaxInt(Level, Wlc_ObjLevelId(p, iFanin))
		}
	}
	Vec_IntWriteEntry(&p.vLevels, Wlc_ObjId(p, pObj), Level+1)
}
func Wlc_NtkCreateLevels(p *Wlc_Ntk_t) int32 {
	var pObj *Wlc_Obj_t
	var i, LeveMax int32
	Vec_IntFill(&p.vLevels, Wlc_NtkObjNumMax(p), 0)
	// inlining of macro Wlc_NtkForEachObj
	for i = 1; i < Wlc_NtkObjNumMax(p); i++ {
		pObj = Wlc_NtkObj(p, i)
		Wlc_NtkCreateLevels_rec(p, pObj)
	}
	// inlining of macro Wlc_NtkForEachObj
	for i = 1; i < Wlc_NtkObjNumMax(p); i++ {
		pObj = Wlc_NtkObj(p, i)
		if Wlc_ObjIsCi(pObj) == 0 && Wlc_ObjFaninNum(pObj) != 0 {
			Vec_IntAddToEntry(&p.vLevels, i, 1)
		}
	}
	LeveMax = Vec_IntFindMax(&p.vLevels)
	// inlining of macro Wlc_NtkForEachFf2
	for i = 0; i < Vec_IntSize(&p.vFfs2); i++ {
		pObj = Wlc_NtkFf2(p, i)
		Vec_IntWriteEntry(&p.vLevels, Wlc_ObjId(p, pObj), LeveMax+1)
	}
	//Wlc_NtkPrintObjects( p );
	return LeveMax + 1
}
func Wlc_NtkRemapLevels(p *Wlc_Ntk_t, vObjs *Vec_Int_t, nLevels int32) int32 {
	var i int32
	var k int32
	var iFanin int32
	var iObj int32
	var Entry int32
	var Level int32
	var Res int32 = nLevels
	var vMap *Vec_Int_t = Vec_IntStart(nLevels + 1)
	var vUsed *Vec_Int_t = Vec_IntStart(nLevels + 1)
	// mark used levels
	Vec_IntWriteEntry(vUsed, nLevels, 1)
	for i = 0; i < Vec_IntSize(vObjs); i++ {
		iObj = Vec_IntEntry(vObjs, i)
		Vec_IntWriteEntry(vUsed, Wlc_ObjLevelId(p, iObj), 1)
		// inlining of macro Wlc_ObjForEachFanin
		for k = 0; k < Wlc_ObjFaninNum(Wlc_NtkObj(p, iObj)); k++ {
			iFanin = Wlc_ObjFaninId(Wlc_NtkObj(p, iObj), k)
			if iFanin != 0 {
				Vec_IntWriteEntry(vUsed, Wlc_ObjLevelId(p, iFanin), 1)
			}
		}
	}
	// create level map
	for i = 0; i < Vec_IntSize(vUsed); i++ {
		Entry = Vec_IntEntry(vUsed, i)
		if Entry != 0 {
			Vec_IntWriteEntry(vMap, i, Level)
			Level++
		}
	}
	//printf( "Total used levels %d -> %d\n", nLevels, Level );
	// remap levels
	for i = 0; i < Vec_IntSize(&p.vLevels); i++ {
		Level = Vec_IntEntry(&p.vLevels, i)
		if Vec_IntEntry(vUsed, Level) != 0 {
			Vec_IntWriteEntry(&p.vLevels, i, Vec_IntEntry(vMap, Level))
		} else {
			Vec_IntWriteEntry(&p.vLevels, i, -1)
		}
	}
	Res = Vec_IntEntry(vMap, nLevels)
	Vec_IntFree(vUsed)
	Vec_IntFree(vMap)
	return Res
}

/*
*Function*************************************************************

	Synopsis    [Collects statistics for each side of the miter.]

	Description []

	SideEffects []

	SeeAlso     []

**********************************************************************
*/
func Wlc_NtkCollectStats(p *Wlc_Ntk_t, nObjs [2][WLC_OBJ_NUMBER]int32) {
	var pObj *Wlc_Obj_t
	var n, i int32
	if Wlc_NtkPoNum(p) != 2 {
		return
	}
	for n = 0; n < 2; n++ {
		Wlc_NtkMarkCone(p, n, 1, 1, 0)
		// inlining of macro Wlc_NtkForEachObj
		for i = 1; i < Wlc_NtkObjNumMax(p); i++ {
			pObj = Wlc_NtkObj(p, i)
			if pObj.Mark != 0 {
				nObjs[n][pObj.Type]++
			}
		}
	}
	Wlc_NtkCleanMarks(p)
}
func Wlc_NtkCountRealPis(p *Wlc_Ntk_t) int32 {
	var pObj *Wlc_Obj_t
	var i, Count int32
	Wlc_NtkMarkCone(p, -1, -1, 1, 0)
	// inlining of macro Wlc_NtkForEachPi
	for i = 0; i < Wlc_NtkPiNum(p); i++ {
		pObj = Wlc_NtkPi(p, i)
		Count += int32(pObj.Mark)
	}
	Wlc_NtkCleanMarks(p)
	return Count
}

/*
*Function*************************************************************

	Synopsis    [Prints distribution of operator types.]

	Description []

	SideEffects []

	SeeAlso     []

**********************************************************************
*/
func Vec_WrdSelectSortCost2(pArray []uint64, nSize int, pCosts []uint64) {
	var i, j, best_i int
	for i = 0; i < nSize-1; i++ {
		best_i = i
		for j = i + 1; j < nSize; j++ {
			if pCosts[j] < pCosts[best_i] {
				best_i = j
			}
		}
		pArray[i], pArray[best_i] = pArray[best_i], pArray[i]
		pCosts[i], pCosts[best_i] = pCosts[best_i], pCosts[i]
	}
}
func Wlc_NtkPrintDistribMakeSign(s int32, s0 int32, s1 int32) uint64 {
	return (uint64(s1) << 42) | (uint64(s0) << 21) | uint64(s)
}
func Wlc_NtkPrintDistribFromSign(sss uint64, s *int32, s0 *int32, s1 *int32) {
	*s1 = int32((sss >> 42))
	*s0 = int32((sss >> 21)) & 0x1FFFFF
	*s = int32(sss) & 0x1FFFFF
}
func Wlc_NtkPrintDistribAddOne(vTypes *Vec_Ptr_t, vOccurs *Vec_Ptr_t, Type int32, Sign uint64) {
	var vType *Vec_Wrd_t = Vec_PtrEntry(vTypes, Type).(*Vec_Wrd_t)
	var vOccur *Vec_Wrd_t = Vec_PtrEntry(vOccurs, Type).(*Vec_Wrd_t)
	var Entry uint64
	var i int32
	for i = 0; i < int32(Vec_WrdSize(vType)); i++ {
		Entry = Vec_WrdEntry(vType, i)
		if Entry == Sign {
			Vec_WrdAddToEntry(vOccur, i, 1)
			return
		}
	}
	Vec_WrdPush(vType, Sign)
	Vec_WrdPush(vOccur, 1)
}
func Wlc_NtkPrintDistribSortOne(vTypes *Vec_Ptr_t, vOccurs *Vec_Ptr_t, Type int32) {
	vType := Vec_PtrEntry(vTypes, Type).(*Vec_Wrd_t)
	vOccur := Vec_PtrEntry(vOccurs, Type).(*Vec_Wrd_t)
	Vec_WrdSelectSortCost2(Vec_WrdArray(vType), int(Vec_WrdSize(vType)), Vec_WrdArray(vOccur))
	Vec_WrdReverseOrder(vType)
	Vec_WrdReverseOrder(vOccur)
}

func Wlc_NtkPrintDistrib(p *Wlc_Ntk_t, fTwoSides int32, fVerbose int32) {
	var nObjs [2][WLC_OBJ_NUMBER]int32 = [2][WLC_OBJ_NUMBER]int32{{0}} // counter of objects of each type
	var pObj *Wlc_Obj_t
	var pObjRange *Wlc_Obj_t = nil
	var nCountRange int32
	var vTypes, vOccurs *Vec_Ptr_t
	var vAnds *Vec_Int_t = Vec_IntStart(int32(WLC_OBJ_NUMBER))
	var Sign uint64
	var i, k, s, s0, s1 int32
	if Wlc_NtkPoNum(p) != 2 {
		fTwoSides = 0
	}
	if fTwoSides != 0 {
		Wlc_NtkCollectStats(p, nObjs)
	}
	// allocate statistics arrays
	vTypes = Vec_PtrStart(int32(WLC_OBJ_NUMBER))
	vOccurs = Vec_PtrStart(int32(WLC_OBJ_NUMBER))
	for i = 0; i < int32(WLC_OBJ_NUMBER); i++ {
		Vec_PtrWriteEntry(vTypes, i, Vec_WrdAlloc(16))
	}
	for i = 0; i < int32(WLC_OBJ_NUMBER); i++ {
		Vec_PtrWriteEntry(vOccurs, i, Vec_WrdAlloc(16))
	}
	// add nodes
	// inlining of macro Wlc_NtkForEachObj
	for i = 1; i < Wlc_NtkObjNumMax(p); i++ {
		pObj = Wlc_NtkObj(p, i)
		//        char * pName = Wlc_ObjName(p, i);
		if Wlc_ObjSign(pObj) > 0x1FFFFF {
			fmt.Printf("Object %6d has range %d, which is reduced to %d in the statistics.\n", i, Wlc_ObjRange(pObj), Wlc_ObjRange(pObj)&0xFFFFF)
		}
		if pObj.Beg != 0 {
			if pObjRange == nil {
				pObjRange = pObj
			}
			nCountRange++
		}
		// 0-input types
		if Wlc_ObjIsCi(pObj) != 0 || pObj.Type == uint32(WLC_OBJ_CONST) || pObj.Type == uint32(WLC_OBJ_BIT_CONCAT) {
			Sign = Wlc_NtkPrintDistribMakeSign(Wlc_ObjSign(pObj), 0, 0)
			// 1-input types
		} else if pObj.Type == uint32(WLC_OBJ_BUF) || pObj.Type == uint32(WLC_OBJ_BIT_SELECT) || pObj.Type == uint32(WLC_OBJ_TABLE) ||
			pObj.Type == uint32(WLC_OBJ_BIT_ZEROPAD) || pObj.Type == uint32(WLC_OBJ_BIT_SIGNEXT) ||
			pObj.Type == uint32(WLC_OBJ_BIT_NOT) || pObj.Type == uint32(WLC_OBJ_LOGIC_NOT) || pObj.Type == uint32(WLC_OBJ_ARI_MINUS) {
			Sign = Wlc_NtkPrintDistribMakeSign(Wlc_ObjSign(pObj), Wlc_ObjSign(Wlc_ObjFanin0(p, pObj)), 0)
			// 2-input types (including MUX)
		} else if Wlc_ObjFaninNum(pObj) == 0 {
			fmt.Printf("Object %d with name \"%s\" has type 0. Looks like it was declared by not defined...\n", i, Wlc_ObjName(p, i))
		} else if Wlc_ObjFaninNum(pObj) == 1 {
			Sign = Wlc_NtkPrintDistribMakeSign(Wlc_ObjSign(pObj), Wlc_ObjSign(Wlc_ObjFanin0(p, pObj)), 0)
		} else {
			assert(Wlc_ObjFaninNum(pObj) >= 2, "Wlc_ObjFaninNum(pObj) >= 2")
			s0, s1 := int32(0), int32(0)
			if Wlc_ObjFaninId(pObj, 0) != 0 {
				s0 = Wlc_ObjSign(Wlc_ObjFanin0(p, pObj))
			}
			if Wlc_ObjFaninId(pObj, 1) != 0 {
				s1 = Wlc_ObjSign(Wlc_ObjFanin1(p, pObj))
			}
			Sign = Wlc_NtkPrintDistribMakeSign(Wlc_ObjSign(pObj), s0, s1)
		}
		// add to storage
		Wlc_NtkPrintDistribAddOne(vTypes, vOccurs, int32(pObj.Type), Sign)
		// count the number of AIG nodes
		if pObj.Type == uint32(WLC_OBJ_MUX) {
			Vec_IntAddToEntry(vAnds, int32(WLC_OBJ_MUX), 3*Wlc_ObjRange(pObj)*(Wlc_ObjFaninNum(pObj)-2))
		} else if pObj.Type == uint32(WLC_OBJ_SHIFT_R) {
			Vec_IntAddToEntry(vAnds, int32(WLC_OBJ_SHIFT_R), Abc_MinInt(Wlc_ObjRange(Wlc_ObjFanin0(p, pObj)), Abc_Base2Log(Wlc_ObjRange(pObj)))*3)
		} else if pObj.Type == uint32(WLC_OBJ_SHIFT_RA) {
			Vec_IntAddToEntry(vAnds, int32(WLC_OBJ_SHIFT_RA), Wlc_ObjRange(pObj)*Abc_MinInt(Wlc_ObjRange(Wlc_ObjFanin0(p, pObj)), Abc_Base2Log(Wlc_ObjRange(pObj)))*3)
		} else if pObj.Type == uint32(WLC_OBJ_SHIFT_L) {
			Vec_IntAddToEntry(vAnds, int32(WLC_OBJ_SHIFT_L), Wlc_ObjRange(pObj)*Abc_MinInt(Wlc_ObjRange(Wlc_ObjFanin0(p, pObj)), Abc_Base2Log(Wlc_ObjRange(pObj)))*3)
		} else if pObj.Type == uint32(WLC_OBJ_SHIFT_LA) {
			Vec_IntAddToEntry(vAnds, int32(WLC_OBJ_SHIFT_LA), Wlc_ObjRange(pObj)*Abc_MinInt(Wlc_ObjRange(Wlc_ObjFanin0(p, pObj)), Abc_Base2Log(Wlc_ObjRange(pObj)))*3)
		} else if pObj.Type == uint32(WLC_OBJ_ROTATE_R) {
			Vec_IntAddToEntry(vAnds, int32(WLC_OBJ_ROTATE_R), Wlc_ObjRange(pObj)*Abc_MinInt(Wlc_ObjRange(Wlc_ObjFanin0(p, pObj)), Abc_Base2Log(Wlc_ObjRange(pObj)))*3)
		} else if pObj.Type == uint32(WLC_OBJ_ROTATE_L) {
			Vec_IntAddToEntry(vAnds, int32(WLC_OBJ_ROTATE_L), Wlc_ObjRange(pObj)*Abc_MinInt(Wlc_ObjRange(Wlc_ObjFanin0(p, pObj)), Abc_Base2Log(Wlc_ObjRange(pObj)))*3)
		} else if pObj.Type == uint32(WLC_OBJ_BIT_NOT) {
			Vec_IntAddToEntry(vAnds, int32(WLC_OBJ_BIT_NOT), 0)
		} else if pObj.Type == uint32(WLC_OBJ_BIT_AND) {
			Vec_IntAddToEntry(vAnds, int32(WLC_OBJ_BIT_AND), Wlc_ObjRange(Wlc_ObjFanin0(p, pObj)))
		} else if pObj.Type == uint32(WLC_OBJ_BIT_OR) {
			Vec_IntAddToEntry(vAnds, int32(WLC_OBJ_BIT_OR), Wlc_ObjRange(Wlc_ObjFanin0(p, pObj)))
		} else if pObj.Type == uint32(WLC_OBJ_BIT_XOR) {
			Vec_IntAddToEntry(vAnds, int32(WLC_OBJ_BIT_XOR), 3*Wlc_ObjRange(Wlc_ObjFanin0(p, pObj)))
		} else if pObj.Type == uint32(WLC_OBJ_BIT_NAND) {
			Vec_IntAddToEntry(vAnds, int32(WLC_OBJ_BIT_NAND), Wlc_ObjRange(Wlc_ObjFanin0(p, pObj)))
		} else if pObj.Type == uint32(WLC_OBJ_BIT_NOR) {
			Vec_IntAddToEntry(vAnds, int32(WLC_OBJ_BIT_NOR), Wlc_ObjRange(Wlc_ObjFanin0(p, pObj)))
		} else if pObj.Type == uint32(WLC_OBJ_BIT_NXOR) {
			Vec_IntAddToEntry(vAnds, int32(WLC_OBJ_BIT_NXOR), 3*Wlc_ObjRange(Wlc_ObjFanin0(p, pObj)))
		} else if pObj.Type == uint32(WLC_OBJ_BIT_SELECT) {
			Vec_IntAddToEntry(vAnds, int32(WLC_OBJ_BIT_SELECT), 0)
		} else if pObj.Type == uint32(WLC_OBJ_BIT_CONCAT) {
			Vec_IntAddToEntry(vAnds, int32(WLC_OBJ_BIT_CONCAT), 0)
		} else if pObj.Type == uint32(WLC_OBJ_BIT_ZEROPAD) {
			Vec_IntAddToEntry(vAnds, int32(WLC_OBJ_BIT_ZEROPAD), 0)
		} else if pObj.Type == uint32(WLC_OBJ_BIT_SIGNEXT) {
			Vec_IntAddToEntry(vAnds, int32(WLC_OBJ_BIT_SIGNEXT), 0)
		} else if pObj.Type == uint32(WLC_OBJ_LOGIC_NOT) {
			Vec_IntAddToEntry(vAnds, int32(WLC_OBJ_LOGIC_NOT), Wlc_ObjRange(Wlc_ObjFanin0(p, pObj))-1)
		} else if pObj.Type == uint32(WLC_OBJ_LOGIC_IMPL) {
			Vec_IntAddToEntry(vAnds, int32(WLC_OBJ_LOGIC_IMPL), Wlc_ObjRange(Wlc_ObjFanin0(p, pObj))+Wlc_ObjRange(Wlc_ObjFanin1(p, pObj))-1)
		} else if pObj.Type == uint32(WLC_OBJ_LOGIC_AND) {
			Vec_IntAddToEntry(vAnds, int32(WLC_OBJ_LOGIC_AND), Wlc_ObjRange(Wlc_ObjFanin0(p, pObj))+Wlc_ObjRange(Wlc_ObjFanin1(p, pObj))-1)
		} else if pObj.Type == uint32(WLC_OBJ_LOGIC_OR) {
			Vec_IntAddToEntry(vAnds, int32(WLC_OBJ_LOGIC_OR), Wlc_ObjRange(Wlc_ObjFanin0(p, pObj))+Wlc_ObjRange(Wlc_ObjFanin1(p, pObj))-1)
		} else if pObj.Type == uint32(WLC_OBJ_LOGIC_XOR) {
			Vec_IntAddToEntry(vAnds, int32(WLC_OBJ_LOGIC_XOR), Wlc_ObjRange(Wlc_ObjFanin0(p, pObj))+Wlc_ObjRange(Wlc_ObjFanin1(p, pObj))+1)
		} else if pObj.Type == uint32(WLC_OBJ_COMP_EQU) {
			Vec_IntAddToEntry(vAnds, int32(WLC_OBJ_COMP_EQU), 4*Wlc_ObjRange(Wlc_ObjFanin0(p, pObj))-1)
		} else if pObj.Type == uint32(WLC_OBJ_COMP_NOTEQU) {
			Vec_IntAddToEntry(vAnds, int32(WLC_OBJ_COMP_NOTEQU), 4*Wlc_ObjRange(Wlc_ObjFanin0(p, pObj))-1)
		} else if pObj.Type == uint32(WLC_OBJ_COMP_LESS) {
			Vec_IntAddToEntry(vAnds, int32(WLC_OBJ_COMP_LESS), 6*Wlc_ObjRange(Wlc_ObjFanin0(p, pObj))-6)
		} else if pObj.Type == uint32(WLC_OBJ_COMP_MORE) {
			Vec_IntAddToEntry(vAnds, int32(WLC_OBJ_COMP_MORE), 6*Wlc_ObjRange(Wlc_ObjFanin0(p, pObj))-6)
		} else if pObj.Type == uint32(WLC_OBJ_COMP_LESSEQU) {
			Vec_IntAddToEntry(vAnds, int32(WLC_OBJ_COMP_LESSEQU), 6*Wlc_ObjRange(Wlc_ObjFanin0(p, pObj))-6)
		} else if pObj.Type == uint32(WLC_OBJ_COMP_MOREEQU) {
			Vec_IntAddToEntry(vAnds, int32(WLC_OBJ_COMP_MOREEQU), 6*Wlc_ObjRange(Wlc_ObjFanin0(p, pObj))-6)
		} else if pObj.Type == uint32(WLC_OBJ_REDUCT_AND) {
			Vec_IntAddToEntry(vAnds, int32(WLC_OBJ_REDUCT_AND), Wlc_ObjRange(Wlc_ObjFanin0(p, pObj))-1)
		} else if pObj.Type == uint32(WLC_OBJ_REDUCT_OR) {
			Vec_IntAddToEntry(vAnds, int32(WLC_OBJ_REDUCT_OR), Wlc_ObjRange(Wlc_ObjFanin0(p, pObj))-1)
		} else if pObj.Type == uint32(WLC_OBJ_REDUCT_XOR) {
			Vec_IntAddToEntry(vAnds, int32(WLC_OBJ_REDUCT_XOR), 3*Wlc_ObjRange(Wlc_ObjFanin0(p, pObj))-3)
		} else if pObj.Type == uint32(WLC_OBJ_REDUCT_NAND) {
			Vec_IntAddToEntry(vAnds, int32(WLC_OBJ_REDUCT_NAND), Wlc_ObjRange(Wlc_ObjFanin0(p, pObj))-1)
		} else if pObj.Type == uint32(WLC_OBJ_REDUCT_NOR) {
			Vec_IntAddToEntry(vAnds, int32(WLC_OBJ_REDUCT_NOR), Wlc_ObjRange(Wlc_ObjFanin0(p, pObj))-1)
		} else if pObj.Type == uint32(WLC_OBJ_REDUCT_NXOR) {
			Vec_IntAddToEntry(vAnds, int32(WLC_OBJ_REDUCT_NXOR), 3*Wlc_ObjRange(Wlc_ObjFanin0(p, pObj))-3)
		} else if pObj.Type == uint32(WLC_OBJ_ARI_ADD) {
			Vec_IntAddToEntry(vAnds, int32(WLC_OBJ_ARI_ADD), 9*Wlc_ObjRange(Wlc_ObjFanin0(p, pObj)))
		} else if pObj.Type == uint32(WLC_OBJ_ARI_SUB) {
			Vec_IntAddToEntry(vAnds, int32(WLC_OBJ_ARI_SUB), 9*Wlc_ObjRange(Wlc_ObjFanin0(p, pObj)))
		} else if pObj.Type == uint32(WLC_OBJ_ARI_MULTI) {
			Vec_IntAddToEntry(vAnds, int32(WLC_OBJ_ARI_MULTI), 9*Wlc_ObjRange(Wlc_ObjFanin0(p, pObj))*Wlc_ObjRange(Wlc_ObjFanin1(p, pObj)))
		} else if pObj.Type == uint32(WLC_OBJ_ARI_DIVIDE) {
			Vec_IntAddToEntry(vAnds, int32(WLC_OBJ_ARI_DIVIDE), 13*Wlc_ObjRange(Wlc_ObjFanin0(p, pObj))*Wlc_ObjRange(Wlc_ObjFanin0(p, pObj))-19*Wlc_ObjRange(Wlc_ObjFanin0(p, pObj))+10)
		} else if pObj.Type == uint32(WLC_OBJ_ARI_REM) {
			Vec_IntAddToEntry(vAnds, int32(WLC_OBJ_ARI_REM), 13*Wlc_ObjRange(Wlc_ObjFanin0(p, pObj))*Wlc_ObjRange(Wlc_ObjFanin0(p, pObj))-7*Wlc_ObjRange(Wlc_ObjFanin0(p, pObj))-2)
		} else if pObj.Type == uint32(WLC_OBJ_ARI_MODULUS) {
			Vec_IntAddToEntry(vAnds, int32(WLC_OBJ_ARI_MODULUS), 13*Wlc_ObjRange(Wlc_ObjFanin0(p, pObj))*Wlc_ObjRange(Wlc_ObjFanin0(p, pObj))-7*Wlc_ObjRange(Wlc_ObjFanin0(p, pObj))-2)
		} else if pObj.Type == uint32(WLC_OBJ_ARI_POWER) {
			Vec_IntAddToEntry(vAnds, int32(WLC_OBJ_ARI_POWER), 10*int32(math.Pow(float64(float64(Wlc_ObjRange(Wlc_ObjFanin0(p, pObj)))), float64(float64(Wlc_ObjRange(Wlc_ObjFanin0(p, pObj)))))))
		} else if pObj.Type == uint32(WLC_OBJ_ARI_MINUS) {
			Vec_IntAddToEntry(vAnds, int32(WLC_OBJ_ARI_MINUS), 4*Wlc_ObjRange(Wlc_ObjFanin0(p, pObj)))
		} else if pObj.Type == uint32(WLC_OBJ_ARI_SQRT) {
			Vec_IntAddToEntry(vAnds, int32(WLC_OBJ_ARI_SQRT), 11*Wlc_ObjRange(Wlc_ObjFanin0(p, pObj))*Wlc_ObjRange(Wlc_ObjFanin0(p, pObj))/8+5*Wlc_ObjRange(Wlc_ObjFanin0(p, pObj))/2-5)
		} else if pObj.Type == uint32(WLC_OBJ_ARI_SQUARE) {
			Vec_IntAddToEntry(vAnds, int32(WLC_OBJ_ARI_SQUARE), 5*Wlc_ObjRange(Wlc_ObjFanin0(p, pObj))*Wlc_ObjRange(Wlc_ObjFanin1(p, pObj)))
		}
	}
	if (nCountRange != 0) && Vec_IntSize(&p.vNameIds) > 0 {
		fmt.Printf("Warning: %d objects of the design have non-zero-based ranges.\n", nCountRange)
		fmt.Printf("In particular, object %6d with name \"%s\" has range %d=[%d:%d]\n", Wlc_ObjId(p, pObjRange), Abc_NamStr(p.pManName, Wlc_ObjNameId(p, Wlc_ObjId(p, pObjRange))), Wlc_ObjRange(pObjRange), pObjRange.End, pObjRange.Beg)
	}
	// print by occurrence
	sideLabel := ""
	if fTwoSides != 0 {
		sideLabel = "     Left Share Right"
	}
	fmt.Printf("ID  :  name  occurrence%s    and2 (occurrence)<output_range>=<input_range>.<input_range> ...\n", sideLabel)
	for i = 0; i < int32(WLC_OBJ_NUMBER); i++ {
		var vType *Vec_Wrd_t = Vec_PtrEntry(vTypes, i).(*Vec_Wrd_t)
		var vOccur *Vec_Wrd_t = Vec_PtrEntry(vOccurs, i).(*Vec_Wrd_t)
		if p.nObjs[i] == 0 {
			continue
		}
		fmt.Printf("%2d  :  %-8s  %6d", i, Wlc_Names[i], p.nObjs[i])
		if fTwoSides != 0 {
			var nTotal int32
			if i == int32(WLC_OBJ_PI) {
				nTotal = Wlc_NtkCountRealPis(p)
			} else {
				nTotal = p.nObjs[i]
			}
			fmt.Printf("   ")
			fmt.Printf("%6d", nObjs[0][i])
			fmt.Printf("%6d", nObjs[0][i]+nObjs[1][i]-nTotal)
			fmt.Printf("%6d", nObjs[1][i])
		}
		fmt.Printf("%8d ", Vec_IntEntry(vAnds, i))
		// sort by occurence
		Wlc_NtkPrintDistribSortOne(vTypes, vOccurs, i)
		for k = 0; k < int32(Vec_WrdSize(vType)); k++ {
			Sign = Vec_WrdEntry(vType, k)
			Wlc_NtkPrintDistribFromSign(Sign, &s, &s0, &s1)
			if ((k%6) == 5 && s1 != 0) || ((k%8) == 7 && s1 == 0) {
				fmt.Printf("\n                                ")
				if fTwoSides != 0 {
					fmt.Printf("                     ")
				}
			}
			fmt.Printf("(%d)", int32(Vec_WrdEntry(vOccur, k)))
			prefix := ""
			if Abc_LitIsCompl(s) != 0 {
				prefix = "-"
			}
			fmt.Printf("%s%d", prefix, Abc_Lit2Var(s))
			if s0 != 0 {
				prefix = ""
				if Abc_LitIsCompl(s0) != 0 {
					prefix = "-"
				}
				fmt.Printf("=%s%d", prefix, Abc_Lit2Var(s0))
			}
			if s1 != 0 {
				prefix = ""
				if Abc_LitIsCompl(s1) != 0 {
					prefix = "-"
				}
				fmt.Printf(".%s%d", prefix, Abc_Lit2Var(s1))
			}
			fmt.Printf(" ")
		}
		fmt.Printf("\n")
	}
	Vec_VecFree((*Vec_Vec_t)(unsafe.Pointer(vTypes)))
	Vec_VecFree((*Vec_Vec_t)(unsafe.Pointer(vOccurs)))
	Vec_IntFree(vAnds)
}
func Wlc_NtkPrintNode(p *Wlc_Ntk_t, pObj *Wlc_Obj_t) {
	fmt.Printf("%8d  :  ", Wlc_ObjId(p, pObj))
	if Vec_IntSize(&p.vLevels) != 0 {
		fmt.Printf("Lev = %2d  ", Vec_IntEntry(&p.vLevels, Wlc_ObjId(p, pObj)))
	}
	signed := " "
	if Wlc_ObjIsSigned(pObj) != 0 {
		signed = "s"
	}
	fmt.Printf("%6d%s = ", Wlc_ObjRange(pObj), signed)
	if pObj.Type == uint32(WLC_OBJ_PI) {
		fmt.Printf("            PI                   :    %-12s\n", Wlc_ObjName(p, Wlc_ObjId(p, pObj)))
		return
	}
	if pObj.Type == uint32(WLC_OBJ_FO) {
		fmt.Printf("            FO                   :    %-12s = %-12s\n", Wlc_ObjName(p, Wlc_ObjId(p, pObj)), Wlc_ObjName(p, Wlc_ObjId(p, Wlc_ObjFo2Fi(p, pObj))))
		return
	}
	if pObj.Type != uint32(WLC_OBJ_CONST) && Wlc_ObjFaninNum(pObj) == 0 {
		fmt.Printf("Unknown object without fanins    :    %-12s\n", Wlc_ObjName(p, Wlc_ObjId(p, pObj)))
		return
	}
	if pObj.Type != uint32(WLC_OBJ_CONST) {
		signed = " "
		if Wlc_ObjIsSigned(Wlc_ObjFanin0(p, pObj)) != 0 {
			signed = "s"
		}
		fmt.Printf("%6d%s  %5s  ", Wlc_ObjRange(Wlc_ObjFanin0(p, pObj)), signed, Wlc_Names[int32(pObj.Type)])
		if Wlc_ObjFaninNum(pObj) > 1 {
			signed = " "
			if Wlc_ObjIsSigned(Wlc_ObjFanin1(p, pObj)) != 0 {
				signed = "s"
			}
			fmt.Printf("%6d%s ", Wlc_ObjRange(Wlc_ObjFanin1(p, pObj)), signed)
		} else {
			fmt.Printf("        ")
		}
		if Wlc_ObjFaninNum(pObj) > 2 {
			signed = " "
			if Wlc_ObjIsSigned(Wlc_ObjFanin2(p, pObj)) != 0 {
				signed = "s"
			}
			fmt.Printf("%6d%s ", Wlc_ObjRange(Wlc_ObjFanin2(p, pObj)), signed)
		} else {
			fmt.Printf("        ")
		}

	} else {
		fmt.Printf("                                ")
	}
	fmt.Printf(" :    ")
	fmt.Printf("%-12s", Wlc_ObjName(p, Wlc_ObjId(p, pObj)))
	if pObj.Type == uint32(WLC_OBJ_CONST) {
		signChar := ""
		if Wlc_ObjIsSigned(pObj) != 0 {
			signChar = "s"
		}
		fmt.Printf(" = %d'%sh", Wlc_ObjRange(pObj), signChar)
		if pObj.fXConst != 0 {
			var k int32
			for k = 0; k < (Wlc_ObjRange(pObj)+3)/4; k++ {
				fmt.Printf("x")
			}

		} else {
			Abc_TtPrintHexArrayRev(os.Stdout, Wlc_ObjConstValue(pObj), (Wlc_ObjRange(pObj)+3)/4)
		}

	} else {
		fmt.Printf(" =  %-12s  %5s  ", Wlc_ObjName(p, Wlc_ObjFaninId0(pObj)), Wlc_Names[int32(pObj.Type)])
		if Wlc_ObjFaninNum(pObj) > 1 {
			fmt.Printf("%-12s ", Wlc_ObjName(p, Wlc_ObjFaninId1(pObj)))
		} else {
			fmt.Printf("             ")
		}
		if Wlc_ObjFaninNum(pObj) > 2 {
			fmt.Printf("%-12s ", Wlc_ObjName(p, Wlc_ObjFaninId2(pObj)))
		}
	}
	fmt.Printf("\n")
}
func Wlc_NtkPrintNodeArray(p *Wlc_Ntk_t, vArray *Vec_Int_t) {
	var pObj *Wlc_Obj_t
	var i int32
	// inlining of macro Wlc_NtkForEachObjVec
	for i = 0; i < Vec_IntSize(vArray); i++ {
		pObj = Wlc_NtkObj(p, Vec_IntEntry(vArray, i))
		Wlc_NtkPrintNode(p, pObj)
	}
}
func Wlc_NtkPrintNodes(p *Wlc_Ntk_t, Type int32) {
	var pObj *Wlc_Obj_t
	var i, Counter int32
	fmt.Printf("Operation %s\n", Wlc_Names[Type])
	// inlining of macro Wlc_NtkForEachObj
	for i = 1; i < Wlc_NtkObjNumMax(p); i++ {
		pObj = Wlc_NtkObj(p, i)
		if int32(pObj.Type) != Type {
			continue
		}
		fmt.Printf("%8d  :", Counter)
		Counter++
		Wlc_NtkPrintNode(p, pObj)
	}
}
func Wlc_NtkPrintStats(p *Wlc_Ntk_t, fDistrib int32, fTwoSides int32, fVerbose int32) {
	var i int32
	fmt.Printf("%-20s : ", p.pName)
	fmt.Printf("PI = %4d  ", Wlc_NtkCountRealPis(p)) //Wlc_NtkPiNum(p) );
	fmt.Printf("PO = %4d  ", Wlc_NtkPoNum(p))
	fmt.Printf("FF = %4d  ", Wlc_NtkFfNum(p))
	fmt.Printf("Obj = %6d  ", Wlc_NtkObjNum(p)-Wlc_NtkPiNum(p)-Wlc_NtkPoNum(p)-Wlc_NtkFfNum(p))
	fmt.Printf("Mem = %.3f MB", float64(Wlc_NtkMemUsage(p))/(1<<20))
	fmt.Printf("\n")
	if fDistrib != 0 {
		Wlc_NtkPrintDistrib(p, fTwoSides, fVerbose)
		return
	}
	if !(fVerbose != 0) {
		return
	}
	fmt.Printf("Node type statistics:\n")
	for i = 1; i < int32(WLC_OBJ_NUMBER); i++ {
		if p.nObjs[i] == 0 {
			continue
		}
		if p.nAnds[0] != 0 && p.nAnds[i] != 0 {
			fmt.Printf("%2d  :  %-8s  %6d  %7.2f %%\n", i, Wlc_Names[i], p.nObjs[i], 100.0*float64(p.nAnds[i])/float64(p.nAnds[0]))
		} else {
			fmt.Printf("%2d  :  %-8s  %6d\n", i, Wlc_Names[i], p.nObjs[i])
		}
	}
}
func Wlc_NtkPrintObjects(p *Wlc_Ntk_t) {
	var pObj *Wlc_Obj_t
	var i int32
	// inlining of macro Wlc_NtkForEachObj
	for i = 1; i < Wlc_NtkObjNumMax(p); i++ {
		pObj = Wlc_NtkObj(p, i)
		Wlc_NtkPrintNode(p, pObj)
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
func Wlc_NtkTransferNames(pNew *Wlc_Ntk_t, p *Wlc_Ntk_t) {
	var i int32
	assert(Wlc_NtkHasCopy(pNew) == 0 && Wlc_NtkHasCopy(p) != 0, "!Wlc_NtkHasCopy(pNew)   && Wlc_NtkHasCopy(p)")
	assert(Wlc_NtkHasNameId(pNew) == 0 && Wlc_NtkHasNameId(p) != 0, "!Wlc_NtkHasNameId(pNew) && Wlc_NtkHasNameId(p)")
	assert(pNew.pManName == nil && p.pManName != nil, "pNew.pManName  == nil && p.pManName != nil")
	Wlc_NtkCleanNameId(pNew)
	for i = 0; i < p.nObjsAlloc; i++ {
		if Wlc_ObjCopy(p, i) > 0 && i < Vec_IntSize(&p.vNameIds) && Wlc_ObjNameId(p, i) != 0 {
			Wlc_ObjSetNameId(pNew, Wlc_ObjCopy(p, i), Wlc_ObjNameId(p, i))
		}
	}
	pNew.pManName = p.pManName
	p.pManName = nil
	Vec_IntErase(&p.vNameIds)
	// transfer table
	pNew.pMemTable = p.pMemTable
	p.pMemTable = nil
	pNew.vTables = p.vTables
	p.vTables = nil
}

var spStatic_Wlc_NtkNewName_pBuffer string // was [1000]byte in C

func Wlc_NtkNewName(p *Wlc_Ntk_t, iCoId int32, fSeq int32) string {

	mode := "comb"
	if fSeq != 0 {
		mode = "seq"
	}
	spStatic_Wlc_NtkNewName_pBuffer = fmt.Sprintf("%s_o%d_%s", p.pName, iCoId, mode)
	return spStatic_Wlc_NtkNewName_pBuffer
}

/*
*Function*************************************************************

	Synopsis    [Reduce init vector.]

	Description []

	SideEffects []

	SeeAlso     []

**********************************************************************
*/
func Wlc_ReduceMarkedInitVec(p *Wlc_Ntk_t, vInit *Vec_Int_t) *Vec_Int_t {
	var vInitNew *Vec_Int_t = Vec_IntDup(vInit)
	var pObj *Wlc_Obj_t
	var i, k int32
	assert(Vec_IntSize(vInit) == Wlc_NtkCiNum(p)-Wlc_NtkPiNum(p), "Vec_IntSize(vInit) == Wlc_NtkCiNum(p) - Wlc_NtkPiNum(p)")
	// inlining of macro Wlc_NtkForEachCi
	for i = 0; i < Wlc_NtkCiNum(p); i++ {
		pObj = Wlc_NtkCi(p, i)
		if Wlc_ObjIsPi(pObj) == 0 && pObj.Mark != 0 {
			Vec_IntWriteEntry(vInitNew, k, Vec_IntEntry(vInit, i-Wlc_NtkPiNum(p)))
			k++
		}
	}
	Vec_IntShrink(vInitNew, k)
	return vInitNew
}
func Wlc_ReduceMarkedInitStr(p *Wlc_Ntk_t, pInit string) string {
	pInitNew := []byte(Abc_UtilStrsav(pInit))
	var pObj *Wlc_Obj_t
	var i, b, nBits, k int32
	// inlining of macro Wlc_NtkForEachCi
	for i = 0; i < Wlc_NtkCiNum(p); i++ {
		pObj = Wlc_NtkCi(p, i)
		if Wlc_ObjIsPi(pObj) == 0 && pObj.Mark != 0 {
			for b = 0; b < Wlc_ObjRange(pObj); b++ {
				pInitNew[k] = pInitNew[nBits+b]
				k++
			}
		}
		if Wlc_ObjIsPi(pObj) == 0 {
			nBits += Wlc_ObjRange(pObj)
		}
	}
	pInitNew[k] = 0
	assert(nBits == int32(len(pInit)), "nBits == int32(strlen(pInit))")
	return string(pInitNew[:k])
}

/*
*Function*************************************************************

	Synopsis    [Duplicates the network in a topological order.]

	Description []

	SideEffects []

	SeeAlso     []

**********************************************************************
*/
func Wlc_ObjCollectCopyFanins(p *Wlc_Ntk_t, iObj int32, vFanins *Vec_Int_t) {
	var i, iFanin int32
	var pObj *Wlc_Obj_t = Wlc_NtkObj(p, iObj)
	Vec_IntClear(vFanins)
	// inlining of macro Wlc_ObjForEachFanin
	for i = 0; i < Wlc_ObjFaninNum(pObj); i++ {
		iFanin = Wlc_ObjFaninId(pObj, i)
		Vec_IntPush(vFanins, Wlc_ObjCopy(p, iFanin))
	}
	// special treatment of CONST and SELECT
	if pObj.Type == uint32(WLC_OBJ_CONST) {
		var pInts *int32 = Wlc_ObjConstValue(pObj)
		var nInts int32 = (Wlc_ObjRange(pObj) + 31) / 32
		for i = 0; i < nInts; i++ {
			Vec_IntPush(vFanins, unsafe.Slice(pInts, int(nInts))[i])
		}

	} else if pObj.Type == uint32(WLC_OBJ_BIT_SELECT) {
		assert(Vec_IntSize(vFanins) == 1, "Vec_IntSize(vFanins) == 1")
		Vec_IntPushTwo(vFanins, Wlc_ObjRangeEnd(pObj), Wlc_ObjRangeBeg(pObj))

	} else if pObj.Type == uint32(WLC_OBJ_TABLE) {
		assert(Vec_IntSize(vFanins) == 1, "Vec_IntSize(vFanins) == 1")
		Vec_IntPush(vFanins, pObj.Fanins[1])
	}
}
func Wlc_ObjDup(pNew *Wlc_Ntk_t, p *Wlc_Ntk_t, iObj int32, vFanins *Vec_Int_t) int32 {
	var pObj *Wlc_Obj_t = Wlc_NtkObj(p, iObj)
	var iFaninNew int32 = Wlc_ObjAlloc(pNew, int32(pObj.Type), Wlc_ObjIsSigned(pObj), pObj.End, pObj.Beg)
	var pObjNew *Wlc_Obj_t = Wlc_NtkObj(pNew, iFaninNew)
	Wlc_ObjCollectCopyFanins(p, iObj, vFanins)
	Wlc_ObjAddFanins(pNew, pObjNew, vFanins)
	Wlc_ObjSetCopy(p, iObj, iFaninNew)
	pObjNew.fXConst = pObj.fXConst
	return iFaninNew
}
func Wlc_NtkDupDfs_rec(pNew *Wlc_Ntk_t, p *Wlc_Ntk_t, iObj int32, vFanins *Vec_Int_t) {
	var pObj *Wlc_Obj_t
	var i, iFanin int32
	if iObj == 0 {
		return
	}
	if Wlc_ObjCopy(p, iObj) != 0 {
		return
	}
	pObj = Wlc_NtkObj(p, iObj)
	//printf( "Visiting node %16s (ID %6d) of type %5s (type ID %2d)\n", Wlc_ObjName(p, iObj), iObj, Wlc_ObjTypeName(pObj), Wlc_NtkObj(p, iObj)->Type );
	assert(pObj.Type != uint32(WLC_OBJ_FF), "pObj.Type != uint32(WLC_OBJ_FF)")
	// inlining of macro Wlc_ObjForEachFanin
	for i = 0; i < Wlc_ObjFaninNum(pObj); i++ {
		iFanin = Wlc_ObjFaninId(pObj, i)
		Wlc_NtkDupDfs_rec(pNew, p, iFanin, vFanins)
	}
	Wlc_ObjDup(pNew, p, iObj, vFanins)
}

func Wlc_NtkDupDfsSimple(p *Wlc_Ntk_t) *Wlc_Ntk_t {
	var pNew *Wlc_Ntk_t
	var pObj *Wlc_Obj_t
	var vFanins *Vec_Int_t
	var i int32
	Wlc_NtkCleanCopy(p)
	vFanins = Vec_IntAlloc(100)
	pNew = Wlc_NtkAlloc(p.pName, p.nObjsAlloc)
	pNew.fSmtLib = p.fSmtLib
	pNew.fAsyncRst = p.fAsyncRst
	pNew.fMemPorts = p.fMemPorts
	pNew.fEasyFfs = p.fEasyFfs
	// inlining of macro Wlc_NtkForEachCi
	for i = 0; i < Wlc_NtkCiNum(p); i++ {
		pObj = Wlc_NtkCi(p, i)
		Wlc_ObjDup(pNew, p, Wlc_ObjId(p, pObj), vFanins)
	}
	// inlining of macro Wlc_NtkForEachCo
	for i = 0; i < Wlc_NtkCoNum(p); i++ {
		pObj = Wlc_NtkCo(p, i)
		Wlc_NtkDupDfs_rec(pNew, p, Wlc_ObjId(p, pObj), vFanins)
	}
	// inlining of macro Wlc_NtkForEachCo
	for i = 0; i < Wlc_NtkCoNum(p); i++ {
		pObj = Wlc_NtkCo(p, i)
		Wlc_ObjSetCo(pNew, Wlc_ObjCopyObj(pNew, p, pObj), int32(pObj.fIsFi))
	}
	if p.vInits != nil {
		pNew.vInits = Vec_IntDup(p.vInits)
	}
	if p.pInits != "" {
		pNew.pInits = Abc_UtilStrsav(p.pInits)
	}
	Vec_IntFree(vFanins)
	if p.pSpec != "" {
		pNew.pSpec = Abc_UtilStrsav(p.pSpec)
	}
	return pNew
}

func Wlc_NtkDupDfs(p *Wlc_Ntk_t, fMarked int32, fSeq int32) *Wlc_Ntk_t {
	var pNew *Wlc_Ntk_t
	var pObj, pObjNew *Wlc_Obj_t
	var vFanins *Vec_Int_t
	var i, k, iObj, iFanin int32
	vFanins = Vec_IntAlloc(100)
	Wlc_NtkCleanCopy(p)
	pNew = Wlc_NtkAlloc(p.pName, p.nObjsAlloc)
	pNew.fSmtLib = p.fSmtLib
	pNew.fAsyncRst = p.fAsyncRst
	pNew.fMemPorts = p.fMemPorts
	pNew.fEasyFfs = p.fEasyFfs
	// inlining of macro Wlc_NtkForEachCi
	for i = 0; i < Wlc_NtkCiNum(p); i++ {
		pObj = Wlc_NtkCi(p, i)
		if fMarked == 0 || pObj.Mark != 0 {
			var Type uint32 = pObj.Type
			if !(fSeq != 0) {
				pObj.Type = uint32(WLC_OBJ_PI)
			}
			Wlc_ObjDup(pNew, p, Wlc_ObjId(p, pObj), vFanins)
			pObj.Type = Type
		}
	}
	// inlining of macro Wlc_NtkForEachFf2
	for i = 0; i < Vec_IntSize(&p.vFfs2); i++ {
		pObj = Wlc_NtkFf2(p, i)
		var iObjNew int32 = Wlc_ObjAlloc(pNew, int32(pObj.Type), Wlc_ObjIsSigned(pObj), pObj.End, pObj.Beg)
		Wlc_ObjSetCopy(p, Wlc_ObjId(p, pObj), iObjNew)
		Vec_IntPush(&pNew.vFfs2, iObjNew)
	}
	// inlining of macro Wlc_NtkForEachCo
	for i = 0; i < Wlc_NtkCoNum(p); i++ {
		pObj = Wlc_NtkCo(p, i)
		if fMarked == 0 || pObj.Mark != 0 {
			Wlc_NtkDupDfs_rec(pNew, p, Wlc_ObjId(p, pObj), vFanins)
		}
	}
	// inlining of macro Wlc_NtkForEachCo
	for i = 0; i < Wlc_NtkCoNum(p); i++ {
		pObj = Wlc_NtkCo(p, i)
		if fMarked == 0 || pObj.Mark != 0 {
			fi := int32(0)
			if fSeq != 0 && pObj.fIsFi != 0 {
				fi = 1
			}
			Wlc_ObjSetCo(pNew, Wlc_ObjCopyObj(pNew, p, pObj), fi)
		}
	}
	// inlining of macro Wlc_NtkForEachFf2
	for i = 0; i < Vec_IntSize(&p.vFfs2); i++ {
		pObj = Wlc_NtkFf2(p, i)
		iObj = Wlc_ObjId(p, pObj)
		// inlining of macro Wlc_ObjForEachFanin
		for k = 0; k < Wlc_ObjFaninNum(pObj); k++ {
			iFanin = Wlc_ObjFaninId(pObj, k)
			Wlc_NtkDupDfs_rec(pNew, p, iFanin, vFanins)
		}
		Wlc_ObjCollectCopyFanins(p, iObj, vFanins)
		pObjNew = Wlc_NtkObj(pNew, Wlc_ObjCopy(p, iObj))
		Wlc_ObjAddFanins(pNew, pObjNew, vFanins)
		pObjNew.fXConst = pObj.fXConst
	}
	Vec_IntFree(vFanins)
	if fSeq != 0 && p.vInits != nil {
		if fMarked != 0 {
			if p.vInits != nil {
				pNew.vInits = Wlc_ReduceMarkedInitVec(p, p.vInits)
			}
			if p.pInits != "" {
				pNew.pInits = Wlc_ReduceMarkedInitStr(p, p.pInits)
			}

		} else {
			if p.vInits != nil {
				pNew.vInits = Vec_IntDup(p.vInits)
			}
			if p.pInits != "" {
				pNew.pInits = Abc_UtilStrsav(p.pInits)
			}
		}
	}
	if p.pSpec != "" {
		pNew.pSpec = Abc_UtilStrsav(p.pSpec)
	}
	if Wlc_NtkHasNameId(p) != 0 {
		Wlc_NtkTransferNames(pNew, p)
	}
	if Vec_IntSize(&p.vPoPairs) != 0 {
		Vec_IntAppend(&pNew.vPoPairs, &p.vPoPairs)
	}
	return pNew
}
func Wlc_NtkDupDfsAbs(p *Wlc_Ntk_t, vPisOld *Vec_Int_t, vPisNew *Vec_Int_t, vFlops *Vec_Int_t) *Wlc_Ntk_t {
	var pNew *Wlc_Ntk_t
	var pObj *Wlc_Obj_t
	var vFanins *Vec_Int_t
	var i int32
	Wlc_NtkCleanCopy(p)
	pNew = Wlc_NtkAlloc(p.pName, p.nObjsAlloc)
	pNew.fSmtLib = p.fSmtLib
	pNew.fAsyncRst = p.fAsyncRst
	pNew.fMemPorts = p.fMemPorts
	pNew.fEasyFfs = p.fEasyFfs

	// duplicate marked PIs
	vFanins = Vec_IntAlloc(100)
	// inlining of macro Wlc_NtkForEachObjVec
	for i = 0; i < Vec_IntSize(vPisOld); i++ {
		pObj = Wlc_NtkObj(p, Vec_IntEntry(vPisOld, i))
		assert(Wlc_ObjIsPi(pObj) != 0, "Wlc_ObjIsPi(pObj)")
		Wlc_ObjDup(pNew, p, Wlc_ObjId(p, pObj), vFanins)
	}
	// duplicate additional PIs
	// inlining of macro Wlc_NtkForEachObjVec
	for i = 0; i < Vec_IntSize(vPisNew); i++ {
		pObj = Wlc_NtkObj(p, Vec_IntEntry(vPisNew, i))
		var Type uint32 = pObj.Type
		var nFanins int32 = Wlc_ObjFaninNum(pObj)
		assert(Wlc_ObjIsPi(pObj) == 0, "!Wlc_ObjIsPi(pObj)")
		pObj.Type = uint32(WLC_OBJ_PI)
		pObj.nFanins = 0
		Wlc_ObjDup(pNew, p, Wlc_ObjId(p, pObj), vFanins)
		pObj.Type = Type
		pObj.nFanins = uint32(nFanins)
	}
	// duplicate flop outputs
	// inlining of macro Wlc_NtkForEachObjVec
	for i = 0; i < Vec_IntSize(vFlops); i++ {
		pObj = Wlc_NtkObj(p, Vec_IntEntry(vFlops, i))
		assert(Wlc_ObjIsPi(pObj) == 0 && Wlc_ObjIsCi(pObj) != 0, "!Wlc_ObjIsPi(pObj) && Wlc_ObjIsCi(pObj)")
		Wlc_ObjDup(pNew, p, Wlc_ObjId(p, pObj), vFanins)
	}

	// duplicate logic cones of primary outputs
	// inlining of macro Wlc_NtkForEachPo
	for i = 0; i < Wlc_NtkPoNum(p); i++ {
		pObj = Wlc_NtkPo(p, i)
		Wlc_NtkDupDfs_rec(pNew, p, Wlc_ObjId(p, pObj), vFanins)
	}
	// duplidate logic cone of flop inputs
	// inlining of macro Wlc_NtkForEachObjVec
	for i = 0; i < Vec_IntSize(vFlops); i++ {
		pObj = Wlc_NtkObj(p, Vec_IntEntry(vFlops, i))
		Wlc_NtkDupDfs_rec(pNew, p, Wlc_ObjId(p, Wlc_ObjFo2Fi(p, pObj)), vFanins)
	}

	// duplicate POs
	// inlining of macro Wlc_NtkForEachPo
	for i = 0; i < Wlc_NtkPoNum(p); i++ {
		pObj = Wlc_NtkPo(p, i)
		Wlc_ObjSetCo(pNew, Wlc_ObjCopyObj(pNew, p, pObj), 0)
	}
	// duplicate flop inputs
	// inlining of macro Wlc_NtkForEachObjVec
	for i = 0; i < Vec_IntSize(vFlops); i++ {
		pObj = Wlc_NtkObj(p, Vec_IntEntry(vFlops, i))
		Wlc_ObjSetCo(pNew, Wlc_ObjCopyObj(pNew, p, Wlc_ObjFo2Fi(p, pObj)), 1)
	}
	Vec_IntFree(vFanins)

	// mark flop outputs
	// inlining of macro Wlc_NtkForEachObjVec
	for i = 0; i < Vec_IntSize(vFlops); i++ {
		pObj = Wlc_NtkObj(p, Vec_IntEntry(vFlops, i))
		pObj.Mark = 1
	}
	if p.vInits != nil {
		pNew.vInits = Wlc_ReduceMarkedInitVec(p, p.vInits)
	}
	if p.pInits != "" {
		pNew.pInits = Wlc_ReduceMarkedInitStr(p, p.pInits)
	}
	Wlc_NtkCleanMarks(p)

	if p.pSpec != "" {
		pNew.pSpec = Abc_UtilStrsav(p.pSpec)
	}
	//Wlc_NtkTransferNames( pNew, p );
	return pNew
}

/*
*Function*************************************************************

	Synopsis    [Select the cone of the given output.]

	Description []

	SideEffects []

	SeeAlso     []

**********************************************************************
*/
func Wlc_NtkCleanMarks(p *Wlc_Ntk_t) {
	var pObj *Wlc_Obj_t
	var i int32
	// inlining of macro Wlc_NtkForEachObj
	for i = 1; i < Wlc_NtkObjNumMax(p); i++ {
		pObj = Wlc_NtkObj(p, i)
		pObj.Mark = 0
	}
}
func Wlc_NtkCountMarked(p *Wlc_Ntk_t, pnPis *int32, pnFos *int32, pnAdders *int32, pnMults *int32) int32 {
	var pObj *Wlc_Obj_t
	var i, nNodes int32
	*pnPis, *pnFos, *pnAdders, *pnMults = 0, 0, 0, 0
	// inlining of macro Wlc_NtkForEachObj
	for i = 1; i < Wlc_NtkObjNumMax(p); i++ {
		pObj = Wlc_NtkObj(p, i)
		if pObj.Mark == 0 {
			continue
		}
		if Wlc_ObjIsPi(pObj) != 0 {
			(*pnPis)++
		} else if Wlc_ObjIsCi(pObj) != 0 {
			(*pnFos)++
		} else if pObj.Mark != 0 {
			nNodes++
			if pObj.Type == uint32(WLC_OBJ_ARI_ADD) || pObj.Type == uint32(WLC_OBJ_ARI_SUB) {
				(*pnAdders)++
			} else if pObj.Type == uint32(WLC_OBJ_ARI_MULTI) {
				(*pnMults)++
			}
		}
	}
	return nNodes
}
func Wlc_NtkMarkCone_rec(p *Wlc_Ntk_t, pObj *Wlc_Obj_t, vFlops *Vec_Int_t) {
	var i, iFanin int32
	if pObj.Mark != 0 {
		return
	}
	pObj.Mark = 1
	if Wlc_ObjIsCi(pObj) != 0 {
		if Wlc_ObjIsPi(pObj) == 0 {
			Vec_IntPush(vFlops, Wlc_ObjCiId(pObj))
		}
		return
	}
	// inlining of macro Wlc_ObjForEachFanin
	for i = 0; i < Wlc_ObjFaninNum(pObj); i++ {
		iFanin = Wlc_ObjFaninId(pObj, i)
		if iFanin != 0 {
			Wlc_NtkMarkCone_rec(p, Wlc_NtkObj(p, iFanin), vFlops)
		}
	}
}
func Wlc_NtkMarkCone(p *Wlc_Ntk_t, iCoId int32, Range int32, fSeq int32, fAllPis int32) {
	var vFlops *Vec_Int_t
	var pObj *Wlc_Obj_t
	var i, CiId, CoId int32
	Wlc_NtkCleanMarks(p)
	if fAllPis != 0 {
		// inlining of macro Wlc_NtkForEachPi
		for i = 0; i < Wlc_NtkPiNum(p); i++ {
			pObj = Wlc_NtkPi(p, i)
			pObj.Mark = 1
		}
	}
	vFlops = Vec_IntAlloc(100)
	// inlining of macro Wlc_NtkForEachCo
	for i = 0; i < Wlc_NtkCoNum(p); i++ {
		pObj = Wlc_NtkCo(p, i)
		if iCoId == -1 || (i >= iCoId && i < iCoId+Range) {
			Wlc_NtkMarkCone_rec(p, pObj, vFlops)
		}
	}
	if fSeq != 0 {
		for i = 0; i < Vec_IntSize(vFlops); i++ {
			CiId = Vec_IntEntry(vFlops, i)
			CoId = Wlc_NtkPoNum(p) + CiId - Wlc_NtkPiNum(p)
			Wlc_NtkMarkCone_rec(p, Wlc_NtkCo(p, CoId), vFlops)
		}
	}
	Vec_IntFree(vFlops)
}
func Wlc_NtkProfileCones(p *Wlc_Ntk_t) {
	var i, nPis, nFos, nNodes, nAdders, nMults int32
	// inlining of macro Wlc_NtkForEachCo
	for i = 0; i < Wlc_NtkCoNum(p); i++ {
		Wlc_NtkMarkCone(p, i, 1, 0, 0)
		nNodes = Wlc_NtkCountMarked(p, &nPis, &nFos, &nAdders, &nMults)
		fmt.Printf("Cone %5d : ", i)
		fmt.Printf("PI = %4d  ", nPis)
		fmt.Printf("FO = %4d  ", nFos)
		fmt.Printf("Node = %6d  ", nNodes)
		fmt.Printf("Add/Sub = %4d  ", nAdders)
		fmt.Printf("Mult = %4d  ", nMults)
		fmt.Printf("\n")
	}
	Wlc_NtkCleanMarks(p)
}

/*
*Function*************************************************************

	Synopsis    [Duplicates the network by copying each node.]

	Description []

	SideEffects []

	SeeAlso     []

**********************************************************************
*/
func Wlc_NtkDupSingleNodes(p *Wlc_Ntk_t) *Wlc_Ntk_t {
	var pNew *Wlc_Ntk_t
	var vFanins *Vec_Int_t
	var pObj, pObjNew *Wlc_Obj_t
	var pFanin *Wlc_Obj_t
	var i, k, iFanin, iFaninNew, iObjNew, Count int32
	// count objects
	// inlining of macro Wlc_NtkForEachObj
	for i = 1; i < Wlc_NtkObjNumMax(p); i++ {
		pObj = Wlc_NtkObj(p, i)
		if Wlc_ObjIsCi(pObj) == 0 {
			Count += 1 + Wlc_ObjFaninNum(pObj)
		}
	}
	// copy objects
	Wlc_NtkCleanCopy(p)
	vFanins = Vec_IntAlloc(100)
	pNew = Wlc_NtkAlloc(p.pName, p.nObjsAlloc)
	pNew.fSmtLib = p.fSmtLib
	pNew.fAsyncRst = p.fAsyncRst
	pNew.fMemPorts = p.fMemPorts
	pNew.fEasyFfs = p.fEasyFfs
	// inlining of macro Wlc_NtkForEachObj
	for i = 1; i < Wlc_NtkObjNumMax(p); i++ {
		pObj = Wlc_NtkObj(p, i)
		if Wlc_ObjIsCi(pObj) != 0 {
			continue
		}
		if pObj.Type == uint32(WLC_OBJ_ARI_MULTI) {
			continue
		}
		if pObj.Type == uint32(WLC_OBJ_MUX) && Wlc_ObjFaninNum(pObj) > 3 {
			continue
		}
		// create CIs for the fanins
		// inlining of macro Wlc_ObjForEachFanin
		for k = 0; k < Wlc_ObjFaninNum(pObj); k++ {
			iFanin = Wlc_ObjFaninId(pObj, k)
			pFanin = Wlc_NtkObj(p, iFanin)
			iFaninNew = Wlc_ObjAlloc(pNew, int32(WLC_OBJ_PI), int32(pFanin.Signed), pFanin.End, pFanin.Beg)
			Wlc_ObjSetCopy(p, iFanin, iFaninNew)
			//Wlc_ObjSetCi( pNew, pFaninNew );
		}
		// create object
		iObjNew = Wlc_ObjDup(pNew, p, i, vFanins)
		pObjNew = Wlc_NtkObj(pNew, iObjNew)
		pObjNew.fIsPo = 1
		Vec_IntPush(&pNew.vPos, iObjNew)
	}
	Vec_IntFree(vFanins)
	Wlc_NtkTransferNames(pNew, p)
	if p.pSpec != "" {
		pNew.pSpec = Abc_UtilStrsav(p.pSpec)
	}
	return pNew
}

/*
*Function*************************************************************

	Synopsis    [Creates short names for all objects.]

	Description []

	SideEffects []

	SeeAlso     []

**********************************************************************
*/
func Wlc_NtkShortNames(p *Wlc_Ntk_t) {
	var pObj *Wlc_Obj_t
	var pBuffer string // was [100]byte in C

	var nDigits, NameId, fFound, i int32
	var nFlops int32 = Wlc_NtkCoNum(p) - Wlc_NtkPoNum(p)
	nDigits = Abc_Base10Log(uint32(nFlops))
	// inlining of macro Wlc_NtkForEachCo
	for i = 0; i < Wlc_NtkCoNum(p); i++ {
		pObj = Wlc_NtkCo(p, i)
		if Wlc_ObjIsPo(pObj) != 0 {
			continue
		}
		pBuffer = fmt.Sprintf("%s%0*d", "fi", nDigits, i-Wlc_NtkPoNum(p))
		NameId = Abc_NamStrFindOrAdd(p.pManName, pBuffer, &fFound)
		Wlc_ObjSetNameId(p, Wlc_ObjId(p, pObj), NameId)
	}
	// inlining of macro Wlc_NtkForEachCi
	for i = 0; i < Wlc_NtkCiNum(p); i++ {
		pObj = Wlc_NtkCi(p, i)
		if Wlc_ObjIsPi(pObj) != 0 {
			continue
		}
		pBuffer = fmt.Sprintf("%s%0*d", "fo", nDigits, i-Wlc_NtkPiNum(p))
		NameId = Abc_NamStrFindOrAdd(p.pManName, pBuffer, &fFound)
		Wlc_ObjSetNameId(p, Wlc_ObjId(p, pObj), NameId)
	}
	nDigits = Abc_Base10Log(uint32(Wlc_NtkPoNum(p)))
	// inlining of macro Wlc_NtkForEachPo
	for i = 0; i < Wlc_NtkPoNum(p); i++ {
		pObj = Wlc_NtkPo(p, i)
		pBuffer = fmt.Sprintf("%s%0*d", "po", nDigits, i)
		NameId = Abc_NamStrFindOrAdd(p.pManName, pBuffer, &fFound)
		Wlc_ObjSetNameId(p, Wlc_ObjId(p, pObj), NameId)
	}
	nDigits = Abc_Base10Log(uint32(Wlc_NtkPiNum(p)))
	// inlining of macro Wlc_NtkForEachPi
	for i = 0; i < Wlc_NtkPiNum(p); i++ {
		pObj = Wlc_NtkPi(p, i)
		pBuffer = fmt.Sprintf("%s%0*d", "pi", nDigits, i)
		NameId = Abc_NamStrFindOrAdd(p.pManName, pBuffer, &fFound)
		Wlc_ObjSetNameId(p, Wlc_ObjId(p, pObj), NameId)
	}
	nDigits = Abc_Base10Log(uint32(Wlc_NtkObjNum(p)))
	// inlining of macro Wlc_NtkForEachObj
	for i = 1; i < Wlc_NtkObjNumMax(p); i++ {
		pObj = Wlc_NtkObj(p, i)
		if Wlc_ObjIsCi(pObj) != 0 || Wlc_ObjIsCo(pObj) != 0 {
			continue
		}
		pBuffer = fmt.Sprintf("%s%0*d", "n", nDigits, i)
		NameId = Abc_NamStrFindOrAdd(p.pManName, pBuffer, &fFound)
		Wlc_ObjSetNameId(p, Wlc_ObjId(p, pObj), NameId)
	}
}

/*
*Function*************************************************************

	Synopsis    [Count the number of flops initialized to DC value.]

	Description []

	SideEffects []

	SeeAlso     []

**********************************************************************
*/
func Wlc_NtkDcFlopNum(p *Wlc_Ntk_t) int32 {
	var i, nFlops, Count int32
	if p.pInits == "" {
		return 0
	}
	nFlops = int32(len(p.pInits))
	for i = 0; i < nFlops; i++ {
		if p.pInits[i] == 'x' || p.pInits[i] == 'X' {
			Count++
		}
	}
	return Count
}

/*
*Function*************************************************************

	Synopsis    [Create references.]

	Description []

	SideEffects []

	SeeAlso     []

**********************************************************************
*/
func Wlc_NtkSetRefs(p *Wlc_Ntk_t) {
	var pObj *Wlc_Obj_t
	var i, k, Fanin int32
	Vec_IntFill(&p.vRefs, Wlc_NtkObjNumMax(p), 0)
	// inlining of macro Wlc_NtkForEachObj
	for i = 1; i < Wlc_NtkObjNumMax(p); i++ {
		pObj = Wlc_NtkObj(p, i)
		// inlining of macro Wlc_ObjForEachFanin
		for k = 0; k < Wlc_ObjFaninNum(pObj); k++ {
			Fanin = Wlc_ObjFaninId(pObj, k)
			Vec_IntAddToEntry(&p.vRefs, Fanin, 1)
		}
	}
	// inlining of macro Wlc_NtkForEachCo
	for i = 0; i < Wlc_NtkCoNum(p); i++ {
		pObj = Wlc_NtkCo(p, i)
		Vec_IntAddToEntry(&p.vRefs, Wlc_ObjId(p, pObj), 1)
	}
}

/*
*Function*************************************************************

	Synopsis    [This procedure simply count the number of PPI bits.]

	Description []

	SideEffects []

	SeeAlso     []

**********************************************************************
*/
func Wlc_NtkCountObjBits(p *Wlc_Ntk_t, vPisNew *Vec_Int_t) int32 {
	var pObj *Wlc_Obj_t
	var i, Count int32
	// inlining of macro Wlc_NtkForEachObjVec
	for i = 0; i < Vec_IntSize(vPisNew); i++ {
		pObj = Wlc_NtkObj(p, Vec_IntEntry(vPisNew, i))
		Count += Wlc_ObjRange(pObj)
	}
	return Count
}

////////////////////////////////////////////////////////////////////////
///                       END OF FILE                                ///
////////////////////////////////////////////////////////////////////////

//ABC_NAMESPACE_IMPL_END
