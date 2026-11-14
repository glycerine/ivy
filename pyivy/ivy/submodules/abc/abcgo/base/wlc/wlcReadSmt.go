package wlc

import (
	"fmt"
	"math/big"
	"strconv"
	"strings"
)

import "os"

/**CFile****************************************************************

  FileName    [wlcParse.c]

  SystemName  [ABC: Logic synthesis and verification system.]

  PackageName [Verilog parser.]

  Synopsis    [Parses several flavors of word-level Verilog.]

  Author      [Alan Mishchenko]

  Affiliation [UC Berkeley]

  Date        [Ver. 1.0. Started - August 22, 2014.]

  Revision    [$Id: wlcParse.c,v 1.00 2014/09/12 00:00:00 alanmi Exp $]

***********************************************************************/

//#include "wlc.h"
//#include "misc/vec/vecWec.h"

//ABC_NAMESPACE_IMPL_START

////////////////////////////////////////////////////////////////////////
///                        DECLARATIONS                              ///
////////////////////////////////////////////////////////////////////////

// parser

type Smt_Prs_t struct {
	// input data
	pName   string     // file name
	pBuffer string     // file contents
	pLimit  int        // byte offset of end of file
	pCur    int        // current byte offset
	pStrs   *Abc_Nam_t // string manager
	// network structure
	vStack Vec_Int_t // current node on each level
	//Vec_Wec_t    vDepth;      // objects on each level
	vObjs     Vec_Wec_t // objects
	NameCount int32
	nDigits   int32
	vTempFans Vec_Int_t
	// error handling
	ErrorStr [1000]byte
}

// #define SMT_GLO_SUFFIX "_glb"
const SMT_GLO_SUFFIX = ""

// parser name types
type Smt_LineType_t int32

const (
	SMT_PRS_NONE Smt_LineType_t = 0

	SMT_PRS_SET_OPTION Smt_LineType_t = SMT_PRS_NONE + 1

	SMT_PRS_SET_LOGIC Smt_LineType_t = SMT_PRS_SET_OPTION + 1

	SMT_PRS_SET_INFO Smt_LineType_t = SMT_PRS_SET_LOGIC + 1

	SMT_PRS_DEFINE_FUN Smt_LineType_t = SMT_PRS_SET_INFO + 1

	SMT_PRS_DECLARE_FUN Smt_LineType_t = SMT_PRS_DEFINE_FUN + 1

	SMT_PRS_ASSERT Smt_LineType_t = SMT_PRS_DECLARE_FUN + 1

	SMT_PRS_LET Smt_LineType_t = SMT_PRS_ASSERT + 1

	SMT_PRS_CHECK_SAT Smt_LineType_t = SMT_PRS_LET + 1

	SMT_PRS_GET_VALUE Smt_LineType_t = SMT_PRS_CHECK_SAT + 1

	SMT_PRS_EXIT Smt_LineType_t = SMT_PRS_GET_VALUE + 1

	SMT_PRS_END Smt_LineType_t = SMT_PRS_EXIT + 1
)

type Smt_Pair_t struct {
	Type  Smt_LineType_t
	pName string
}

var s_Types [SMT_PRS_END]Smt_Pair_t = [SMT_PRS_END]Smt_Pair_t{
	{SMT_PRS_NONE, ""},
	{SMT_PRS_SET_OPTION, "set-option"},
	{SMT_PRS_SET_LOGIC, "set-logic"},
	{SMT_PRS_SET_INFO, "set-info"},
	{SMT_PRS_DEFINE_FUN, "define-fun"},
	{SMT_PRS_DECLARE_FUN, "declare-fun"},
	{SMT_PRS_ASSERT, "assert"},
	{SMT_PRS_LET, "let"},
	{SMT_PRS_CHECK_SAT, "check-sat"},
	{SMT_PRS_GET_VALUE, "get-value"},
	{SMT_PRS_EXIT, "exit"},
}

func Smt_GetTypeName(Type Smt_LineType_t) string {
	var i int32
	for i = 1; i < int32(SMT_PRS_END); i++ {
		if s_Types[i].Type == Type {
			return s_Types[i].pName
		}
	}
	return ""
}
func Smt_AddTypes(p *Abc_Nam_t) {
	var Type int32
	for Type = 1; Type < int32(SMT_PRS_END); Type++ {
		Abc_NamStrFindOrAdd(p, Smt_GetTypeName(Smt_LineType_t(Type)), nil)
	}
	assert(Abc_NamObjNumMax(p) == int32(SMT_PRS_END), "Abc_NamObjNumMax(p) == SMT_PRS_END")
}

func Smt_EntryIsName(Fan int32) int32 { return Abc_LitIsCompl(Fan) }
func Smt_EntryIsType(Fan int32, Type Smt_LineType_t) int32 {
	assert((Smt_EntryIsName(Fan) != 0), "Smt_EntryIsName(Fan)")
	return boolToInt32(Abc_Lit2Var(Fan) == int32(Type))
}
func Smt_EntryName(p *Smt_Prs_t, Fan int32) string {
	assert((Smt_EntryIsName(Fan) != 0), "Smt_EntryIsName(Fan)")
	return Abc_NamStr(p.pStrs, Abc_Lit2Var(Fan))
}
func Smt_EntryNode(p *Smt_Prs_t, Fan int32) *Vec_Int_t {
	assert(!(Smt_EntryIsName(Fan) != 0), "!Smt_EntryIsName(Fan)")
	return Vec_WecEntry(&p.vObjs, Abc_Lit2Var(Fan))
}

func Smt_VecEntryIsType(vFans *Vec_Int_t, i int32, Type Smt_LineType_t) int32 {
	return boolToInt32(i < Vec_IntSize(vFans) && Smt_EntryIsName(Vec_IntEntry(vFans, i)) != 0 && Smt_EntryIsType(Vec_IntEntry(vFans, i), Type) != 0)
}
func Smt_VecEntryName(p *Smt_Prs_t, vFans *Vec_Int_t, i int32) string {
	if Smt_EntryIsName(Vec_IntEntry(vFans, i)) != 0 {
		return Smt_EntryName(p, Vec_IntEntry(vFans, i))
	} else {
		return ""
	}
}
func Smt_VecEntryNode(p *Smt_Prs_t, vFans *Vec_Int_t, i int32) *Vec_Int_t {
	if Smt_EntryIsName(Vec_IntEntry(vFans, i)) != 0 {
		return nil
	} else {
		return Smt_EntryNode(p, Vec_IntEntry(vFans, i))
	}
}

// Original macro:
// #define Smt_ManForEachDir( p, Type, vVec, i )                                                       \
//     for ( i = 0; (i < Vec_WecSize(&p->vObjs)) && (((vVec) = Vec_WecEntry(&p->vObjs, i)), 1); i++ )  \
//         if ( !Smt_VecEntryIsType(vVec, 0, Type) ) {} else

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
func Smt_StrToType(pName string, pfSigned *int32) int32 {
	var Type int32
	*pfSigned = 0
	if !(int32(strings.Compare(pName, "ite")) != 0) {
		Type = int32(WLC_OBJ_MUX) // 08: multiplexer
	} else if !(int32(strings.Compare(pName, "bvlshr")) != 0) {
		Type = int32(WLC_OBJ_SHIFT_R) // 09: shift right
	} else if !(int32(strings.Compare(pName, "bvashr")) != 0) {
		Type = int32(WLC_OBJ_SHIFT_RA)
		*pfSigned = 1 // 10: shift right (arithmetic)
	} else if !(int32(strings.Compare(pName, "bvshl")) != 0) {
		Type = int32(WLC_OBJ_SHIFT_L) // 11: shift left
		//    else if ( !strcmp(pName, "") )
		//        Type = WLC_OBJ_SHIFT_LA;    // 12: shift left (arithmetic)
	} else if !(int32(strings.Compare(pName, "rotate_right")) != 0) {
		Type = int32(WLC_OBJ_ROTATE_R) // 13: rotate right
	} else if !(int32(strings.Compare(pName, "rotate_left")) != 0) {
		Type = int32(WLC_OBJ_ROTATE_L) // 14: rotate left
	} else if !(int32(strings.Compare(pName, "bvnot")) != 0) {
		Type = int32(WLC_OBJ_BIT_NOT) // 15: bitwise NOT
	} else if !(int32(strings.Compare(pName, "bvand")) != 0) {
		Type = int32(WLC_OBJ_BIT_AND) // 16: bitwise AND
	} else if !(int32(strings.Compare(pName, "bvor")) != 0) {
		Type = int32(WLC_OBJ_BIT_OR) // 17: bitwise OR
	} else if !(int32(strings.Compare(pName, "bvxor")) != 0) {
		Type = int32(WLC_OBJ_BIT_XOR) // 18: bitwise XOR
	} else if !(int32(strings.Compare(pName, "bvnand")) != 0) {
		Type = int32(WLC_OBJ_BIT_NAND) // 16: bitwise NAND
	} else if !(int32(strings.Compare(pName, "bvnor")) != 0) {
		Type = int32(WLC_OBJ_BIT_NOR) // 17: bitwise NOR
	} else if !(int32(strings.Compare(pName, "bvxnor")) != 0) {
		Type = int32(WLC_OBJ_BIT_NXOR) // 18: bitwise NXOR
	} else if !(int32(strings.Compare(pName, "extract")) != 0) {
		Type = int32(WLC_OBJ_BIT_SELECT) // 19: bit selection
	} else if !(int32(strings.Compare(pName, "concat")) != 0) {
		Type = int32(WLC_OBJ_BIT_CONCAT) // 20: bit concatenation
	} else if !(int32(strings.Compare(pName, "zero_extend")) != 0) {
		Type = int32(WLC_OBJ_BIT_ZEROPAD) // 21: zero padding
	} else if !(int32(strings.Compare(pName, "sign_extend")) != 0) {
		Type = int32(WLC_OBJ_BIT_SIGNEXT) // 22: sign extension
	} else if !(int32(strings.Compare(pName, "not")) != 0) {
		Type = int32(WLC_OBJ_LOGIC_NOT) // 23: logic NOT
	} else if !(int32(strings.Compare(pName, "=>")) != 0) {
		Type = int32(WLC_OBJ_LOGIC_IMPL) // 24: logic AND
	} else if !(int32(strings.Compare(pName, "and")) != 0) {
		Type = int32(WLC_OBJ_LOGIC_AND) // 24: logic AND
	} else if !(int32(strings.Compare(pName, "or")) != 0) {
		Type = int32(WLC_OBJ_LOGIC_OR) // 25: logic OR
	} else if !(int32(strings.Compare(pName, "xor")) != 0) {
		Type = int32(WLC_OBJ_LOGIC_XOR) // 26: logic OR
	} else if !(int32(strings.Compare(pName, "bvcomp")) != 0) || !(int32(strings.Compare(pName, "=")) != 0) {
		Type = int32(WLC_OBJ_COMP_EQU) // 27: compare equal
	} else if !(int32(strings.Compare(pName, "distinct")) != 0) {
		Type = int32(WLC_OBJ_COMP_NOTEQU) // 28: compare not equal
	} else if !(int32(strings.Compare(pName, "bvult")) != 0) {
		Type = int32(WLC_OBJ_COMP_LESS) // 29: compare less
	} else if !(int32(strings.Compare(pName, "bvugt")) != 0) {
		Type = int32(WLC_OBJ_COMP_MORE) // 30: compare more
	} else if !(int32(strings.Compare(pName, "bvule")) != 0) {
		Type = int32(WLC_OBJ_COMP_LESSEQU) // 31: compare less or equal
	} else if !(int32(strings.Compare(pName, "bvuge")) != 0) {
		Type = int32(WLC_OBJ_COMP_MOREEQU) // 32: compare more or equal
	} else if !(int32(strings.Compare(pName, "bvslt")) != 0) {
		Type = int32(WLC_OBJ_COMP_LESS)
		*pfSigned = 1 // 29: compare less
	} else if !(int32(strings.Compare(pName, "bvsgt")) != 0) {
		Type = int32(WLC_OBJ_COMP_MORE)
		*pfSigned = 1 // 30: compare more
	} else if !(int32(strings.Compare(pName, "bvsle")) != 0) {
		Type = int32(WLC_OBJ_COMP_LESSEQU)
		*pfSigned = 1 // 31: compare less or equal
	} else if !(int32(strings.Compare(pName, "bvsge")) != 0) {
		Type = int32(WLC_OBJ_COMP_MOREEQU)
		*pfSigned = 1 // 32: compare more or equal
	} else if !(int32(strings.Compare(pName, "bvredand")) != 0) {
		Type = int32(WLC_OBJ_REDUCT_AND) // 33: reduction AND
	} else if !(int32(strings.Compare(pName, "bvredor")) != 0) {
		Type = int32(WLC_OBJ_REDUCT_OR) // 34: reduction OR
	} else if !(int32(strings.Compare(pName, "bvredxor")) != 0) {
		Type = int32(WLC_OBJ_REDUCT_XOR) // 35: reduction XOR
	} else if !(int32(strings.Compare(pName, "bvadd")) != 0) {
		Type = int32(WLC_OBJ_ARI_ADD) // 36: arithmetic addition
	} else if !(int32(strings.Compare(pName, "bvsub")) != 0) {
		Type = int32(WLC_OBJ_ARI_SUB) // 37: arithmetic subtraction
	} else if !(int32(strings.Compare(pName, "bvmul")) != 0) {
		Type = int32(WLC_OBJ_ARI_MULTI) // 38: arithmetic multiplier
	} else if !(int32(strings.Compare(pName, "bvudiv")) != 0) {
		Type = int32(WLC_OBJ_ARI_DIVIDE) // 39: arithmetic division
	} else if !(int32(strings.Compare(pName, "bvurem")) != 0) {
		Type = int32(WLC_OBJ_ARI_REM) // 40: arithmetic remainder
	} else if !(int32(strings.Compare(pName, "bvsdiv")) != 0) {
		Type = int32(WLC_OBJ_ARI_DIVIDE)
		*pfSigned = 1 // 39: arithmetic division
	} else if !(int32(strings.Compare(pName, "bvsrem")) != 0) {
		Type = int32(WLC_OBJ_ARI_REM)
		*pfSigned = 1 // 40: arithmetic remainder
	} else if !(int32(strings.Compare(pName, "bvsmod")) != 0) {
		Type = int32(WLC_OBJ_ARI_MODULUS)
		*pfSigned = 1 // 40: arithmetic modulus
	} else if !(int32(strings.Compare(pName, "=")) != 0) {
		Type = int32(WLC_OBJ_COMP_EQU) // 40: arithmetic modulus
		//    else if ( !strcmp(pName, "") )
		//        Type = WLC_OBJ_ARI_POWER;     // 41: arithmetic power
	} else if !(int32(strings.Compare(pName, "bvneg")) != 0) {
		Type = int32(WLC_OBJ_ARI_MINUS) // 42: arithmetic minus
		//    else if ( !strcmp(pName, "") )
		//        Type = WLC_OBJ_TABLE;         // 43: bit table
	} else {
		fmt.Printf("The following operations is currently not supported (%s)\n", pName)
		_ = os.Stdout.Sync()
	}
	return Type
}
func Smt_PrsReadType(p *Smt_Prs_t, iSig int32, pfSigned *int32, Value1 *int32, Value2 *int32) int32 {
	if Smt_EntryIsName(iSig) != 0 {
		return Smt_StrToType(Smt_EntryName(p, iSig), pfSigned)
	} else {
		var vFans *Vec_Int_t = Smt_EntryNode(p, iSig)
		var pStr string = Smt_VecEntryName(p, vFans, 0)
		var Type int32
		assert(Vec_IntSize(vFans) >= 3, "Vec_IntSize(vFans) >= 3")
		assert(!(int32(strings.Compare(pStr, "_")) != 0), "!strcmp(pStr, \"_\")") // special op
		*Value1, *Value2 = -1, -1
		assert(pStr[0] != 'b' || pStr[1] != 'v', "pStr[0] != 'b' || pStr[1] != 'v'")
		// read type
		Type = Smt_StrToType(Smt_VecEntryName(p, vFans, 1), pfSigned)
		if Type == 0 {
			return 0
		}
		*Value1 = atoi(Smt_VecEntryName(p, vFans, 2))
		if Vec_IntSize(vFans) > 3 {
			*Value2 = atoi(Smt_VecEntryName(p, vFans, 3))
		}
		return Type
	}
}

func Smt_StrType(str string) int32 { return Smt_StrToType(str, nil) }

/*
*Function*************************************************************

	Synopsis    []

	Description []

	SideEffects []

	SeeAlso     []

**********************************************************************
*/
func Smt_PrsCreateNodeOld(pNtk *Wlc_Ntk_t, Type int32, fSigned int32, Range int32, vFanins *Vec_Int_t, pName string) int32 {
	var Buffer string // was [100]byte in C

	var NameId, fFound int32
	var iObj int32 = Wlc_ObjAlloc(pNtk, Type, fSigned, Range-1, 0)
	assert(Type > 0, "Type > 0")
	assert(Range >= 0, "Range >= 0")
	assert(fSigned >= 0, "fSigned >= 0")

	// add node's name
	if pName == "" {
		Buffer = fmt.Sprintf("_n%d_", iObj)
		pName = Buffer
	}
	NameId = Abc_NamStrFindOrAdd(pNtk.pManName, pName, &fFound)
	assert(!(fFound != 0), "!fFound")
	assert(iObj == NameId, "iObj == NameId")

	Wlc_ObjAddFanins(pNtk, Wlc_NtkObj(pNtk, iObj), vFanins)
	if fSigned != 0 {
		Wlc_NtkObj(pNtk, iObj).Signed = uint32(fSigned)
		//        if ( Vec_IntSize(vFanins) > 0 )
		//            Wlc_NtkObj(pNtk, Vec_IntEntry(vFanins, 0))->Signed = fSigned;
		//        if ( Vec_IntSize(vFanins) > 1 )
		//            Wlc_NtkObj(pNtk, Vec_IntEntry(vFanins, 1))->Signed = fSigned;
	}

	return iObj
}
func Smt_PrsCreateNode(pNtk *Wlc_Ntk_t, Type int32, fSigned int32, Range int32, vFanins *Vec_Int_t, pName string) int32 {
	var Buffer string // was [100]byte in C

	var pNameFanin string
	var NameId, fFound, old_size, new_size int32
	var iObj, iFanin0, iFanin1 int32
	var v2Fanins *Vec_Int_t = Vec_IntStartFull(2)

	assert(Type > 0, "Type > 0")
	assert(Range >= 0, "Range >= 0")
	assert(fSigned >= 0, "fSigned >= 0")

	//if (Vec_IntSize(vFanins)<=2 || Type == int32(WLC_OBJ_BIT_CONCAT) || Type == int32(WLC_OBJ_MUX) )
	// explicitely secify allowed multi operators
	if Vec_IntSize(vFanins) <= 2 ||
		!(Type == int32(WLC_OBJ_BIT_AND) || // 16:`` bitwise AND
			Type == int32(WLC_OBJ_BIT_OR) || // 17: bitwise OR
			Type == int32(WLC_OBJ_BIT_XOR) || // 18: bitwise XOR
			Type == int32(WLC_OBJ_BIT_NAND) || // 19: bitwise AND
			Type == int32(WLC_OBJ_BIT_NOR) || // 20: bitwise OR
			Type == int32(WLC_OBJ_BIT_NXOR) || // 21: bitwise NXOR

			Type == int32(WLC_OBJ_LOGIC_IMPL) || // 27: logic implication
			Type == int32(WLC_OBJ_LOGIC_AND) || // 28: logic AND
			Type == int32(WLC_OBJ_LOGIC_OR) || // 29: logic OR
			Type == int32(WLC_OBJ_LOGIC_XOR) || // 30: logic XOR
			Type == int32(WLC_OBJ_COMP_EQU) || // 31: compare equal
			//          Type == int32(WLC_OBJ_COMP_NOTEQU) ||   // 32: compare not equal -- bug fix
			Type == int32(WLC_OBJ_COMP_LESS) || // 33: compare less
			Type == int32(WLC_OBJ_COMP_MORE) || // 34: compare more
			Type == int32(WLC_OBJ_COMP_LESSEQU) || // 35: compare less or equal
			Type == int32(WLC_OBJ_COMP_MOREEQU) || // 36: compare more or equal

			Type == int32(WLC_OBJ_ARI_ADD) || // 43: arithmetic addition
			Type == int32(WLC_OBJ_ARI_SUB) || // 44: arithmetic subtraction
			Type == int32(WLC_OBJ_ARI_MULTI) || // 45: arithmetic multiplier
			Type == int32(WLC_OBJ_ARI_DIVIDE)) { // 46: arithmetic division
		goto FINISHED_WITH_FANINS
	}

	// we will be creating nodes backwards. this way we can pop from vFanins easier
	for Vec_IntSize(vFanins) > 2 {
		// getting 2 fanins at a time
		old_size = Vec_IntSize(vFanins)
		iFanin0 = Vec_IntPop(vFanins)
		iFanin1 = Vec_IntPop(vFanins)

		Vec_IntWriteEntry(v2Fanins, 0, iFanin0)
		Vec_IntWriteEntry(v2Fanins, 1, iFanin1)

		iObj = Wlc_ObjAlloc(pNtk, Type, fSigned, Range-1, 0)
		Buffer = fmt.Sprintf("_n%d_", iObj)
		pNameFanin = Buffer
		NameId = Abc_NamStrFindOrAdd(pNtk.pManName, pNameFanin, &fFound)

		assert(!(fFound != 0), "!fFound")
		assert(iObj == NameId, "iObj == NameId")

		Wlc_ObjAddFanins(pNtk, Wlc_NtkObj(pNtk, iObj), v2Fanins)

		// pushing the new node
		Vec_IntPush(vFanins, iObj)
		new_size = Vec_IntSize(vFanins)
		assert(new_size < old_size, "new_size<old_size")
	}

FINISHED_WITH_FANINS:

	Vec_IntFree(v2Fanins)

	//added to deal with long shifts create extra bit select (ROTATE as well ??)
	// this is a temporary hack
	// basically we keep only 32 bits.
	// bit[31] will be the copy of original MSB (sign bit, just in case) UPDATE: assume it is unsigned first????
	// bit[31] will be the reduction or of any bits from [31] to Range
	if Type == int32(WLC_OBJ_SHIFT_R) || Type == int32(WLC_OBJ_SHIFT_RA) || Type == int32(WLC_OBJ_SHIFT_L) {
		var range1, iObj1, iObj2, iObj3 int32
		assert(Vec_IntSize(vFanins) >= 2, "Vec_IntSize(vFanins)>=2")
		iFanin1 = Vec_IntEntry(vFanins, 1)
		range1 = Wlc_ObjRange(Wlc_NtkObj(pNtk, iFanin1))
		if range1 > 32 {
			var newFanins *Vec_Int_t = Vec_IntAlloc(10)
			Vec_IntPush(newFanins, iFanin1)
			Vec_IntPushTwo(newFanins, 30, 0)

			iObj1 = Wlc_ObjAlloc(pNtk, int32(WLC_OBJ_BIT_SELECT), 0, 30, 0)
			Buffer = fmt.Sprintf("_n%d_", iObj1)
			pNameFanin = Buffer
			NameId = Abc_NamStrFindOrAdd(pNtk.pManName, pNameFanin, &fFound)

			assert(!(fFound != 0), "!fFound")
			assert(iObj1 == NameId, "iObj1 == NameId")

			Wlc_ObjAddFanins(pNtk, Wlc_NtkObj(pNtk, iObj1), newFanins)

			//printf("obj1: %d\n",iObj1);

			// bit select of larger bits
			Vec_IntPop(newFanins)
			Vec_IntPop(newFanins)
			Vec_IntPushTwo(newFanins, range1-1, 31)
			iObj2 = Wlc_ObjAlloc(pNtk, int32(WLC_OBJ_BIT_SELECT), 0, range1-1, 31)
			Buffer = fmt.Sprintf("_n%d_", iObj2)
			pNameFanin = Buffer
			NameId = Abc_NamStrFindOrAdd(pNtk.pManName, pNameFanin, &fFound)

			assert(!(fFound != 0), "!fFound")
			assert(iObj2 == NameId, "iObj2 == NameId")

			Wlc_ObjAddFanins(pNtk, Wlc_NtkObj(pNtk, iObj2), newFanins)
			//printf("obj2: %d\n",iObj2);

			// reduction or
			Vec_IntPop(newFanins)
			Vec_IntPop(newFanins)
			Vec_IntWriteEntry(newFanins, 0, iObj2)
			iObj3 = Wlc_ObjAlloc(pNtk, int32(WLC_OBJ_REDUCT_OR), 0, 0, 0)
			Buffer = fmt.Sprintf("_n%d_", iObj3)
			pNameFanin = Buffer
			NameId = Abc_NamStrFindOrAdd(pNtk.pManName, pNameFanin, &fFound)

			assert(!(fFound != 0), "!fFound")
			assert(iObj3 == NameId, "iObj3 == NameId")

			Wlc_ObjAddFanins(pNtk, Wlc_NtkObj(pNtk, iObj3), newFanins)
			//printf("obj3: %d\n",iObj3);

			// concat all together
			Vec_IntWriteEntry(newFanins, 0, iObj3)
			Vec_IntPush(newFanins, iObj1)
			iObj = Wlc_ObjAlloc(pNtk, int32(WLC_OBJ_BIT_CONCAT), 0, 31, 0)
			Buffer = fmt.Sprintf("_n%d_", iObj)
			pNameFanin = Buffer
			NameId = Abc_NamStrFindOrAdd(pNtk.pManName, pNameFanin, &fFound)

			assert(!(fFound != 0), "!fFound")
			assert(iObj == NameId, "iObj == NameId")

			Wlc_ObjAddFanins(pNtk, Wlc_NtkObj(pNtk, iObj), newFanins)
			//printf("obj: %d\n",iObj);

			// pushing the new node
			Vec_IntWriteEntry(vFanins, 1, iObj)
			Vec_IntFree(newFanins)
		}
	}

	iObj = Wlc_ObjAlloc(pNtk, Type, fSigned, Range-1, 0)

	// add node's name
	if pName == "" {
		Buffer = fmt.Sprintf("_n%d_", iObj)
		pName = Buffer
	}
	NameId = Abc_NamStrFindOrAdd(pNtk.pManName, pName, &fFound)
	assert(!(fFound != 0), "!fFound")
	assert(iObj == NameId, "iObj == NameId")

	Wlc_ObjAddFanins(pNtk, Wlc_NtkObj(pNtk, iObj), vFanins)
	if fSigned != 0 {
		Wlc_NtkObj(pNtk, iObj).Signed = uint32(fSigned)
		//        if ( Vec_IntSize(vFanins) > 0 )
		//            Wlc_NtkObj(pNtk, Vec_IntEntry(vFanins, 0))->Signed = fSigned;
		//        if ( Vec_IntSize(vFanins) > 1 )
		//            Wlc_NtkObj(pNtk, Vec_IntEntry(vFanins, 1))->Signed = fSigned;
	}

	return iObj
}

func Smt_GetHexFromDecimalString(pStr string) string {
	value, ok := new(big.Int).SetString(pStr, 10)
	if !ok || value.Sign() < 0 {
		return ""
	}
	return value.Text(16)
}

// smtReadHexWords mirrors the parser's least-significant-word-first constant layout.
func smtReadHexWords(dst []int32, hexText string) int32 {
	value, ok := new(big.Int).SetString(hexText, 16)
	if !ok {
		return 0
	}
	for i := range dst {
		dst[i] = int32(new(big.Int).And(value, big.NewInt(0xffffffff)).Uint64())
		value.Rsh(value, 32)
	}
	return int32(len(hexText))
}

func Smt_PrsBuildConstant(pNtk *Wlc_Ntk_t, pStr string, nBits int32, pName string) int32 {
	var i, nDigits, iObj int32
	var vFanins *Vec_Int_t = Vec_IntAlloc(10)
	if pStr[0] != '#' { // decimal

		if pStr[0] >= '0' && pStr[0] <= '9' {
			// added: sanity check for large constants
			/*
			   Vec_Int_t * temp = Vec_IntAlloc(10);
			   int fullBits = -1;
			   Smt_GetBinaryFromDecimalString(pStr,temp,&fullBits);
			   Vec_IntFree(temp);
			   assert(fullBits < 32);*/

			var pHex string = Smt_GetHexFromDecimalString(pStr)

			if nBits == -1 {
				nBits = int32(len(pHex) * 4)
			}

			//printf("nbits: %d\n",nBits);

			Vec_IntFill(vFanins, (nBits+31)/32, 0)
			nDigits = smtReadHexWords(Vec_IntArray(vFanins), pHex)
			pHex = ""

			/*
			   int w, nWords, Number = atoi( pStr );
			   if ( nBits == -1 )
			   {
			       nBits = Abc_Base2Log( Number+1 );
			       assert( nBits < 32 );
			   }
			   nWords = Abc_BitWordNum( nBits );
			   for ( w = 0; w < nWords; w++ )
			       Vec_IntPush( vFanins, w ? 0 : Number );
			*/

		} else {
			var fFound int32
			var iObj int32 = Abc_NamStrFindOrAdd(pNtk.pManName, pStr, &fFound)
			assert((fFound != 0), "fFound")
			Vec_IntFree(vFanins)
			return iObj
		}

	} else if pStr[1] == 'b' { // binary

		if nBits == -1 {
			nBits = int32(len(pStr[2:]))
		}
		Vec_IntFill(vFanins, (nBits+31)/32, 0)
		for i = 0; i < nBits; i++ {
			if pStr[2+i] == '1' {
				*Vec_IntEntryP(vFanins, (nBits-1-i)/32) |= int32(1) << uint((nBits-1-i)%32)
			} else if pStr[2+i] != '0' {
				Vec_IntFree(vFanins)
				return 0
			}
		}

	} else if pStr[1] == 'x' { // hexadecimal

		if nBits == -1 {
			nBits = int32(len(pStr[2:])) * 4
		}
		Vec_IntFill(vFanins, (nBits+31)/32, 0)
		nDigits = smtReadHexWords(Vec_IntArray(vFanins), pStr[2:])
		if nDigits != (nBits+3)/4 {
			Vec_IntFree(vFanins)
			return 0
		}

	} else {
		Vec_IntFree(vFanins)
		return 0
	}
	// create constant node
	iObj = Smt_PrsCreateNode(pNtk, int32(WLC_OBJ_CONST), 0, nBits, vFanins, pName)
	Vec_IntFree(vFanins)
	return iObj
}
func Smt_PrsBuildNode(pNtk *Wlc_Ntk_t, p *Smt_Prs_t, iNode int32, RangeOut int32, pName string) int32 {
	if Smt_EntryIsName(iNode) != 0 { // name or constant

		var pStr string = Abc_NamStr(p.pStrs, Abc_Lit2Var(iNode))
		if (pStr[0] >= '0' && pStr[0] <= '9') || pStr[0] == '#' {
			// (_ BitVec 8) #x19
			return Smt_PrsBuildConstant(pNtk, pStr, RangeOut, pName)

		} else {
			// s3087
			var fFound int32
			var iObj int32 = Abc_NamStrFindOrAdd(pNtk.pManName, pStr, &fFound)
			assert((fFound != 0), "fFound")
			return iObj
		}

	} else { // node

		var vFans *Vec_Int_t = Smt_EntryNode(p, iNode)
		var pStr0 string = Smt_VecEntryName(p, vFans, 0)
		var pStr1 string = Smt_VecEntryName(p, vFans, 1)
		if pStr0 != "" && pStr1 != "" && pStr0[0] == '_' && pStr1[0] == 'b' && pStr1[1] == 'v' {
			// (_ bv1 32)
			var pStr2 string = Smt_VecEntryName(p, vFans, 2)
			assert(Vec_IntSize(vFans) == 3, "Vec_IntSize(vFans) == 3")
			return Smt_PrsBuildConstant(pNtk, pStr1[2:], atoi(pStr2), pName)

		} else if pStr0 != "" && pStr0[0] == '=' {
			assert(Vec_IntSize(vFans) == 3, "Vec_IntSize(vFans) == 3")
			iNode = Vec_IntEntry(vFans, 2)
			assert((Smt_EntryIsName(iNode) != 0), "Smt_EntryIsName(iNode)")
			pStr0 = Smt_EntryName(p, iNode)
			// check the last one is "#b1"
			if !(int32(strings.Compare("#b1", pStr0)) != 0) {
				iNode = Vec_IntEntry(vFans, 1)
				return Smt_PrsBuildNode(pNtk, p, iNode, -1, pName)

			} else {
				var vFanins *Vec_Int_t = Vec_IntAlloc(2)
				// get the constant
				var iObj int32
				var iOper int32
				var iConst int32 = Smt_PrsBuildConstant(pNtk, pStr0, -1, "")
				// check the middle one is an operator
				iNode = Vec_IntEntry(vFans, 1)
				iOper = Smt_PrsBuildNode(pNtk, p, iNode, -1, pName)
				// build comparator
				Vec_IntPushTwo(vFanins, iOper, iConst)
				iObj = Smt_PrsCreateNode(pNtk, int32(WLC_OBJ_COMP_EQU), 0, 1, vFanins, pName)
				Vec_IntFree(vFanins)
				return iObj
			}

		} else {
			var i int32
			var Fan int32
			var NameId int32
			var iFanin int32
			var fSigned int32
			var Range int32
			var Value1 int32 = -1
			var Value2 int32 = -1
			var Type int32 = Smt_PrsReadType(p, Vec_IntEntry(vFans, 0), &fSigned, &Value1, &Value2)
			// collect fanins
			var vFanins *Vec_Int_t = Vec_IntAlloc(100)
			for i = 1; i < Vec_IntSize(vFans); i++ {
				Fan = Vec_IntEntry(vFans, i)
				iFanin = Smt_PrsBuildNode(pNtk, p, Fan, -1, "")
				if iFanin == 0 {
					Vec_IntFree(vFanins)
					return 0
				}
				Vec_IntPush(vFanins, iFanin)
			}
			// update specialized nodes
			assert(Type != int32(WLC_OBJ_BIT_SIGNEXT) && Type != int32(WLC_OBJ_BIT_ZEROPAD), "Type != int32(WLC_OBJ_BIT_SIGNEXT) && Type != int32(WLC_OBJ_BIT_ZEROPAD)")
			if Type == int32(WLC_OBJ_BIT_SELECT) {
				assert(Value1 >= 0 && Value2 >= 0 && Value1 >= Value2, "Value1 >= 0 && Value2 >= 0 && Value1 >= Value2")
				Vec_IntPushTwo(vFanins, Value1, Value2)

			} else if Type == int32(WLC_OBJ_ROTATE_R) || Type == int32(WLC_OBJ_ROTATE_L) {
				var Buffer string // was [100]byte in C

				assert(Value1 >= 0, "Value1 >= 0")
				Buffer = fmt.Sprintf("%d", Value1)
				NameId = Smt_PrsBuildConstant(pNtk, Buffer, -1, "")
				Vec_IntPush(vFanins, NameId)
			}
			// find range
			Range = 0
			if Type >= int32(WLC_OBJ_LOGIC_NOT) && Type <= int32(WLC_OBJ_REDUCT_XOR) {
				Range = 1
			} else if Type == int32(WLC_OBJ_BIT_SELECT) {
				Range = Value1 - Value2 + 1
			} else if Type == int32(WLC_OBJ_BIT_CONCAT) {
				for i = 0; i < Vec_IntSize(vFanins); i++ {
					NameId = Vec_IntEntry(vFanins, i)
					Range += Wlc_ObjRange(Wlc_NtkObj(pNtk, NameId))
				}

			} else if Type == int32(WLC_OBJ_MUX) {
				var pArray []int32 = Vec_IntArray(vFanins)
				assert(Vec_IntSize(vFanins) == 3, "Vec_IntSize(vFanins) == 3")
				pArray[1], pArray[2] = pArray[2], pArray[1]
				NameId = Vec_IntEntry(vFanins, 1)
				Range = Wlc_ObjRange(Wlc_NtkObj(pNtk, NameId))

			} else { // to determine range, look at the first argument

				NameId = Vec_IntEntry(vFanins, 0)
				Range = Wlc_ObjRange(Wlc_NtkObj(pNtk, NameId))
			}
			// create node
			assert(Range > 0, "Range > 0")
			NameId = Smt_PrsCreateNode(pNtk, Type, fSigned, Range, vFanins, pName)
			Vec_IntFree(vFanins)
			return NameId
		}
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
func Smt_PrsBuild(p *Smt_Prs_t) *Wlc_Ntk_t {
	var pNtk *Wlc_Ntk_t
	var vFans, vFans2, vFans3 *Vec_Int_t
	var vAsserts *Vec_Int_t = Vec_IntAlloc(100)
	var pName, pRange, pValue string
	var i, k, Fan, Fan3, iObj, Status, Range, NameId, nBits int32
	// issue warnings about unknown dirs
	vFans = Vec_WecEntry(&p.vObjs, 0)
	for i = 0; i < Vec_IntSize(vFans); i++ {
		Fan = Vec_IntEntry(vFans, i)
		assert(!(Smt_EntryIsName(Fan) != 0), "!Smt_EntryIsName(Fan)")
		vFans2 = Smt_VecEntryNode(p, vFans, i)
		Fan = Vec_IntEntry(vFans2, 0)
		assert((Smt_EntryIsName(Fan) != 0), "Smt_EntryIsName(Fan)")
		if Abc_Lit2Var(Fan) >= int32(SMT_PRS_END) {
			fmt.Printf("Ignoring directive \"%s\".\n", Smt_EntryName(p, Fan))
		}
	}
	// start network and create primary inputs
	pNtk = Wlc_NtkAlloc(p.pName, 1000)
	pNtk.pManName = Abc_NamStart(1000, 24)
	pNtk.fSmtLib = 1
	// inlining of macro Smt_ManForEachDir
	for i = 0; i < Vec_WecSize(&p.vObjs); i++ {
		vFans = Vec_WecEntry(&p.vObjs, i)
		if !(Smt_VecEntryIsType(vFans, 0, SMT_PRS_DECLARE_FUN) != 0) {
		} else {
			assert(Vec_IntSize(vFans) == 4, "Vec_IntSize(vFans) == 4")
			assert((Smt_VecEntryIsType(vFans, 0, SMT_PRS_DECLARE_FUN) != 0), "Smt_VecEntryIsType(vFans, 0, SMT_PRS_DECLARE_FUN)")
			// get name
			Fan = Vec_IntEntry(vFans, 1)
			assert((Smt_EntryIsName(Fan) != 0), "Smt_EntryIsName(Fan)")
			pName = Smt_EntryName(p, Fan)
			// skip ()
			Fan = Vec_IntEntry(vFans, 2)
			assert(!(Smt_EntryIsName(Fan) != 0), "!Smt_EntryIsName(Fan)")
			// check type (Bool or BitVec)
			Fan = Vec_IntEntry(vFans, 3)
			if Smt_EntryIsName(Fan) != 0 {
				//(declare-fun s1 () Bool)
				assert(!(int32(strings.Compare("Bool", Smt_VecEntryName(p, vFans, 3))) != 0), "!strcmp(\"Bool\", Smt_VecEntryName(p, vFans, 3))")
				Range = 1

			} else {
				// (declare-fun s1 () (_ BitVec 64))
				// get range
				Fan = Vec_IntEntry(vFans, 3)
				assert(!(Smt_EntryIsName(Fan) != 0), "!Smt_EntryIsName(Fan)")
				vFans2 = Smt_EntryNode(p, Fan)
				assert(Vec_IntSize(vFans2) == 3, "Vec_IntSize(vFans2) == 3")
				assert(!(int32(strings.Compare("_", Smt_VecEntryName(p, vFans2, 0))) != 0), "!strcmp(\"_\", Smt_VecEntryName(p, vFans2, 0))")
				assert(!(int32(strings.Compare("BitVec", Smt_VecEntryName(p, vFans2, 1))) != 0), "!strcmp(\"BitVec\", Smt_VecEntryName(p, vFans2, 1))")
				Fan = Vec_IntEntry(vFans2, 2)
				assert((Smt_EntryIsName(Fan) != 0), "Smt_EntryIsName(Fan)")
				pRange = Smt_EntryName(p, Fan)
				Range = atoi(pRange)
			}
			// create node
			iObj = Wlc_ObjAlloc(pNtk, int32(WLC_OBJ_PI), 0, Range-1, 0)
			NameId = Abc_NamStrFindOrAdd(pNtk.pManName, pName, nil)
			assert(iObj == NameId, "iObj == NameId")
			// save values
			Vec_IntPush(&pNtk.vValues, NameId)
			Vec_IntPush(&pNtk.vValues, nBits)
			Vec_IntPush(&pNtk.vValues, Range)
			nBits += Range
		}
	}
	// create constants
	// inlining of macro Smt_ManForEachDir
	for i = 0; i < Vec_WecSize(&p.vObjs); i++ {
		vFans = Vec_WecEntry(&p.vObjs, i)
		if !(Smt_VecEntryIsType(vFans, 0, SMT_PRS_DEFINE_FUN) != 0) {
		} else {
			assert(Vec_IntSize(vFans) == 5, "Vec_IntSize(vFans) == 5")
			assert((Smt_VecEntryIsType(vFans, 0, SMT_PRS_DEFINE_FUN) != 0), "Smt_VecEntryIsType(vFans, 0, SMT_PRS_DEFINE_FUN)")
			// get name
			Fan = Vec_IntEntry(vFans, 1)
			assert((Smt_EntryIsName(Fan) != 0), "Smt_EntryIsName(Fan)")
			pName = Smt_EntryName(p, Fan)
			// skip ()
			Fan = Vec_IntEntry(vFans, 2)
			assert(!(Smt_EntryIsName(Fan) != 0), "!Smt_EntryIsName(Fan)")
			// check type (Bool or BitVec)
			Fan = Vec_IntEntry(vFans, 3)
			if Smt_EntryIsName(Fan) != 0 {
				// (define-fun s_2 () Bool false)
				assert(!(int32(strings.Compare("Bool", Smt_VecEntryName(p, vFans, 3))) != 0), "!strcmp(\"Bool\", Smt_VecEntryName(p, vFans, 3))")
				Range = 1
				pValue = Smt_VecEntryName(p, vFans, 4)
				if !(int32(strings.Compare("false", pValue)) != 0) {
					pValue = "#b0"
				} else if !(int32(strings.Compare("true", pValue)) != 0) {
					pValue = "#b1"
				} else {
					assert(false, "0")
				}
				Status = Smt_PrsBuildConstant(pNtk, pValue, Range, pName)

			} else {
				// (define-fun s702 () (_ BitVec 4) #xe)
				// (define-fun s1 () (_ BitVec 8) (bvneg #x7f))
				// get range
				Fan = Vec_IntEntry(vFans, 3)
				assert(!(Smt_EntryIsName(Fan) != 0), "!Smt_EntryIsName(Fan)")
				vFans2 = Smt_VecEntryNode(p, vFans, 3)
				assert(Vec_IntSize(vFans2) == 3, "Vec_IntSize(vFans2) == 3")
				assert(!(int32(strings.Compare("_", Smt_VecEntryName(p, vFans2, 0))) != 0), "!strcmp(\"_\", Smt_VecEntryName(p, vFans2, 0))")
				assert(!(int32(strings.Compare("BitVec", Smt_VecEntryName(p, vFans2, 1))) != 0), "!strcmp(\"BitVec\", Smt_VecEntryName(p, vFans2, 1))")
				// get range
				Fan = Vec_IntEntry(vFans2, 2)
				assert((Smt_EntryIsName(Fan) != 0), "Smt_EntryIsName(Fan)")
				pRange = Smt_EntryName(p, Fan)
				Range = atoi(pRange)
				// get constant
				Fan = Vec_IntEntry(vFans, 4)
				Status = Smt_PrsBuildNode(pNtk, p, Fan, Range, pName)
			}
			if Status == 0 {
				Wlc_NtkFree(pNtk)
				pNtk = nil
				goto finish
			}
		}
	}
	// process let-statements
	// inlining of macro Smt_ManForEachDir
	for i = 0; i < Vec_WecSize(&p.vObjs); i++ {
		vFans = Vec_WecEntry(&p.vObjs, i)
		if !(Smt_VecEntryIsType(vFans, 0, SMT_PRS_LET) != 0) {
		} else {
			// let ((s35550 (bvor s48 s35549)))
			assert(Vec_IntSize(vFans) == 3, "Vec_IntSize(vFans) == 3")
			assert((Smt_VecEntryIsType(vFans, 0, SMT_PRS_LET) != 0), "Smt_VecEntryIsType(vFans, 0, SMT_PRS_LET)")
			// get parts
			Fan = Vec_IntEntry(vFans, 1)
			assert(!(Smt_EntryIsName(Fan) != 0), "!Smt_EntryIsName(Fan)")
			vFans2 = Smt_EntryNode(p, Fan)
			if Smt_VecEntryIsType(vFans2, 0, SMT_PRS_LET) != 0 {
				continue
			}
			// iterate through the parts
			for k = 0; k < Vec_IntSize(vFans2); k++ {
				Fan = Vec_IntEntry(vFans2, k)
				// s35550 (bvor s48 s35549)
				Fan = Vec_IntEntry(vFans2, 0)
				assert(!(Smt_EntryIsName(Fan) != 0), "!Smt_EntryIsName(Fan)")
				vFans3 = Smt_EntryNode(p, Fan)
				// get the name
				Fan3 = Vec_IntEntry(vFans3, 0)
				assert((Smt_EntryIsName(Fan3) != 0), "Smt_EntryIsName(Fan3)")
				pName = Smt_EntryName(p, Fan3)
				// get function
				Fan3 = Vec_IntEntry(vFans3, 1)
				assert(!(Smt_EntryIsName(Fan3) != 0), "!Smt_EntryIsName(Fan3)")
				Status = Smt_PrsBuildNode(pNtk, p, Fan3, -1, pName)
				if Status == 0 {
					Wlc_NtkFree(pNtk)
					pNtk = nil
					goto finish
				}
			}
		}
	}
	// process assert-statements
	Vec_IntClear(vAsserts)
	// inlining of macro Smt_ManForEachDir
	for i = 0; i < Vec_WecSize(&p.vObjs); i++ {
		vFans = Vec_WecEntry(&p.vObjs, i)
		if !(Smt_VecEntryIsType(vFans, 0, SMT_PRS_ASSERT) != 0) {
		} else {
			//(assert ; no quantifiers
			//   (let ((s2 (bvsge s0 s1)))
			//   (not s2)))
			//(assert (not (= s0 #x00)))
			assert(Vec_IntSize(vFans) == 2, "Vec_IntSize(vFans) == 2")
			assert((Smt_VecEntryIsType(vFans, 0, SMT_PRS_ASSERT) != 0), "Smt_VecEntryIsType(vFans, 0, SMT_PRS_ASSERT)")
			// get second directive
			Fan = Vec_IntEntry(vFans, 1)
			if !(Smt_EntryIsName(Fan) != 0) {
				// find the final let-statement
				vFans2 = Smt_VecEntryNode(p, vFans, 1)
				for Smt_VecEntryIsType(vFans2, 0, SMT_PRS_LET) != 0 {
					Fan = Vec_IntEntry(vFans2, 2)
					if Smt_EntryIsName(Fan) != 0 {
						break
					}
					vFans2 = Smt_VecEntryNode(p, vFans2, 2)
				}
			}
			// find name or create node
			iObj = Smt_PrsBuildNode(pNtk, p, Fan, -1, "")
			if !(iObj != 0) {
				Wlc_NtkFree(pNtk)
				pNtk = nil
				goto finish
			}
			Vec_IntPush(vAsserts, iObj)
		}
	}
	// build AND of asserts
	if Vec_IntSize(vAsserts) == 1 {
		iObj = Smt_PrsCreateNode(pNtk, int32(WLC_OBJ_BUF), 0, 1, vAsserts, "miter_output")
	} else {
		iObj = Smt_PrsCreateNode(pNtk, int32(WLC_OBJ_BIT_CONCAT), 0, Vec_IntSize(vAsserts), vAsserts, "")
		Vec_IntFill(vAsserts, 1, iObj)
		iObj = Smt_PrsCreateNode(pNtk, int32(WLC_OBJ_REDUCT_AND), 0, 1, vAsserts, "miter_output")
	}
	Wlc_ObjSetCo(pNtk, Wlc_NtkObj(pNtk, iObj), 0)
	// create nameIDs
	vFans = Vec_IntStartNatural(Wlc_NtkObjNumMax(pNtk))
	Vec_IntAppend(&pNtk.vNameIds, vFans)
	Vec_IntFree(vFans)
	//Wlc_NtkReport( pNtk, NULL );
finish:
	// cleanup
	Vec_IntFree(vAsserts)
	return pNtk
}

/*
*Function*************************************************************

	Synopsis    []

	Description []

	SideEffects []

	SeeAlso     []

**********************************************************************
*/
var spStatic_Smt_PrsGenName_Buffer string // was [16]byte in C

func atoi(s string) int32 { n, _ := strconv.Atoi(s); return int32(n) }

func smtNameOrGenerated(p *Smt_Prs_t, name string) string {
	if name != "" {
		return name
	}
	return Smt_PrsGenName(p)
}

func Smt_PrsGenName(p *Smt_Prs_t) string {

	p.NameCount++
	spStatic_Smt_PrsGenName_Buffer = fmt.Sprintf("_%0*X_", p.nDigits, p.NameCount)
	return spStatic_Smt_PrsGenName_Buffer
}
func Smt_PrsBuild2_rec(pNtk *Wlc_Ntk_t, p *Smt_Prs_t, iNode int32, iObjPrev int32, pName string) int32 {
	var suffix string // was [100]byte in C

	suffix = fmt.Sprintf("_as%d", pNtk.nAssert)

	//char * prepStr = Abc_NamStr(p->pStrs, Abc_Lit2Var(iNode));
	//printf("prestr: %s\n",prepStr);

	//printf("inode: %d %d\n",iNode,Smt_EntryIsName(iNode));

	if Smt_EntryIsName(iNode) != 0 {
		var pStr string = Abc_NamStr(p.pStrs, Abc_Lit2Var(iNode))
		// handle built-in constants
		if !(int32(strings.Compare(pStr, "false")) != 0) {
			pStr = "#b0"
		} else if !(int32(strings.Compare(pStr, "true")) != 0) {
			pStr = "#b1"
		}
		if pStr[0] == '#' {
			return Smt_PrsBuildConstant(pNtk, pStr, -1, smtNameOrGenerated(p, pName))
		} else {
			var fFound, iObj int32
			// look either for global DECLARE-FUN variable or local LET
			pStr_glb := pStr + SMT_GLO_SUFFIX
			pStr_loc := pStr + suffix

			fFound = Abc_NamStrFind(pNtk.pManName, pStr_glb)

			if fFound != 0 {
				pStr = pStr_glb
			} else {
				assert(Abc_NamStrFind(pNtk.pManName, pStr_loc) != 0, "Abc_NamStrFind( pNtk.pManName, pStr_loc )")
				pStr = pStr_loc
			}
			// FIXME: delete memory of pStr

			iObj = Abc_NamStrFindOrAdd(pNtk.pManName, pStr, &fFound)
			assert((fFound != 0), "fFound")
			// create buffer if the name of the fanin has different name
			if pName != "" && (int32(strings.Compare(Wlc_ObjName(pNtk, iObj), pName)) != 0) {
				Vec_IntFill(&p.vTempFans, 1, iObj)
				iObj = Smt_PrsCreateNode(pNtk, int32(WLC_OBJ_BUF), 0, Wlc_ObjRange(Wlc_NtkObj(pNtk, iObj)), &p.vTempFans, pName)
			}
			pStr_glb = ""
			pStr_loc = ""
			return iObj
		}

	} else {
		var vRoots, vRoots1, vFans3 *Vec_Int_t
		var iObj, iRoot0, iRoot1, iRoot2, Fan, Fan3, k int32
		assert(!(Smt_EntryIsName(iNode) != 0), "!Smt_EntryIsName(iNode)")
		vRoots = Smt_EntryNode(p, iNode)
		iRoot0 = Vec_IntEntry(vRoots, 0)
		if Smt_EntryIsName(iRoot0) != 0 {
			var pName2 string
			var pStr0 string = Abc_NamStr(p.pStrs, Abc_Lit2Var(iRoot0))
			if Abc_Lit2Var(iRoot0) == int32(SMT_PRS_LET) || Abc_Lit2Var(iRoot0) == int32(SMT_PRS_DEFINE_FUN) { //added define-fun is similar to let

				// let ((s35550 (bvor s48 s35549)))
				assert(Vec_IntSize(vRoots) == 3, "Vec_IntSize(vRoots) == 3")
				assert((Smt_VecEntryIsType(vRoots, 0, SMT_PRS_LET) != 0), "Smt_VecEntryIsType(vRoots, 0, SMT_PRS_LET)")
				// get parts
				iRoot1 = Vec_IntEntry(vRoots, 1)
				assert(!(Smt_EntryIsName(iRoot1) != 0), "!Smt_EntryIsName(iRoot1)")
				vRoots1 = Smt_EntryNode(p, iRoot1)
				// iterate through the parts
				for k = 0; k < Vec_IntSize(vRoots1); k++ {
					Fan = Vec_IntEntry(vRoots1, k)
					var temp string
					// s35550 (bvor s48 s35549)
					assert(!(Smt_EntryIsName(Fan) != 0), "!Smt_EntryIsName(Fan)")
					vFans3 = Smt_EntryNode(p, Fan)
					assert(Vec_IntSize(vFans3) == 2, "Vec_IntSize(vFans3) == 2")
					// get the name
					Fan3 = Vec_IntEntry(vFans3, 0)
					assert((Smt_EntryIsName(Fan3) != 0), "Smt_EntryIsName(Fan3)")
					pName2 = Smt_EntryName(p, Fan3)
					// create a local name with suffix
					if Abc_Lit2Var(iRoot0) == int32(SMT_PRS_LET) {
						temp = pName2 + suffix

					} else {
						temp = pName2 + SMT_GLO_SUFFIX
					}
					// FIXME: need to delete memory of pName2
					pName2 = temp
					// get function
					Fan3 = Vec_IntEntry(vFans3, 1)
					//assert( !Smt_EntryIsName(Fan3) );
					// solve the problem
					iObj = Smt_PrsBuild2_rec(pNtk, p, Fan3, -1, pName2) // NULL ); //pName2 );
					temp = ""
					if iObj == 0 {
						return 0
					}
					// create buffer
					//Vec_IntFill( &p->vTempFans, 1, iObj );
					//Smt_PrsCreateNode( pNtk, int32(WLC_OBJ_BUF), 0, Wlc_ObjRange(Wlc_NtkObj(pNtk, iObj)), &p->vTempFans, pName2 );
				}
				// process the final part of let
				iRoot2 = Vec_IntEntry(vRoots, 2)
				return Smt_PrsBuild2_rec(pNtk, p, iRoot2, -1, pName)

			} else if pStr0[0] == '_' {
				var pStr1 string = Smt_VecEntryName(p, vRoots, 1)
				if pStr1[0] == 'b' && pStr1[1] == 'v' {
					// (_ bv1 32)
					var pStr2 string = Smt_VecEntryName(p, vRoots, 2)
					assert(Vec_IntSize(vRoots) == 3, "Vec_IntSize(vRoots) == 3")
					return Smt_PrsBuildConstant(pNtk, pStr1[2:], atoi(pStr2), smtNameOrGenerated(p, pName))

				} else {
					var Type1 int32
					var fSigned int32
					var Range int32 = -1
					assert(iObjPrev != -1, "iObjPrev != -1")
					Type1 = Smt_StrToType(pStr1, &fSigned)
					if Type1 == 0 {
						return 0
					}
					// find out this branch
					Vec_IntFill(&p.vTempFans, 1, iObjPrev)
					if Type1 == int32(WLC_OBJ_BIT_SIGNEXT) || Type1 == int32(WLC_OBJ_BIT_ZEROPAD) || Type1 == int32(WLC_OBJ_ROTATE_R) || Type1 == int32(WLC_OBJ_ROTATE_L) {
						var iRoot2 int32 = Vec_IntEntry(vRoots, 2)
						var pStr2 string = Abc_NamStr(p.pStrs, Abc_Lit2Var(iRoot2))
						var Num int32 = atoi(pStr2)
						//fSigned = (Type1 == int32(WLC_OBJ_BIT_SIGNEXT));
						if Type1 == int32(WLC_OBJ_ROTATE_R) || Type1 == int32(WLC_OBJ_ROTATE_L) {
							var iConst int32 = Smt_PrsBuildConstant(pNtk, pStr2, -1, Smt_PrsGenName(p))
							Vec_IntClear(&p.vTempFans)
							Vec_IntPushTwo(&p.vTempFans, iObjPrev, iConst)
							Range = Wlc_ObjRange(Wlc_NtkObj(pNtk, iObjPrev))

						} else {
							Range = Num + Wlc_ObjRange(Wlc_NtkObj(pNtk, iObjPrev))
						}

					} else if Type1 == int32(WLC_OBJ_BIT_SELECT) {
						var iRoot2 int32 = Vec_IntEntry(vRoots, 2)
						var iRoot3 int32 = Vec_IntEntry(vRoots, 3)
						var pStr2 string = Abc_NamStr(p.pStrs, Abc_Lit2Var(iRoot2))
						var pStr3 string = Abc_NamStr(p.pStrs, Abc_Lit2Var(iRoot3))
						var Num1 int32 = atoi(pStr2)
						var Num2 int32 = atoi(pStr3)
						assert(Num1 >= Num2, "Num1 >= Num2")
						Range = Num1 - Num2 + 1
						Vec_IntPushTwo(&p.vTempFans, Num1, Num2)

					} else {
						assert(false, "0")
					}
					iObj = Smt_PrsCreateNode(pNtk, Type1, fSigned, Range, &p.vTempFans, smtNameOrGenerated(p, pName))
					return iObj
				}

			} else {
				var vFanins *Vec_Int_t
				var i, Fan, fSigned, Range, Type0 int32
				var iObj int32 = Abc_NamStrFind(pNtk.pManName, pStr0)
				if iObj != 0 {
					return iObj
				}
				Type0 = Smt_StrToType(pStr0, &fSigned)
				if Type0 == 0 {
					return 0
				}
				assert(Type0 != int32(WLC_OBJ_BIT_SIGNEXT) && Type0 != int32(WLC_OBJ_BIT_ZEROPAD) && Type0 != int32(WLC_OBJ_BIT_SELECT) && Type0 != int32(WLC_OBJ_ROTATE_R) && Type0 != int32(WLC_OBJ_ROTATE_L), "Type0 != int32(WLC_OBJ_BIT_SIGNEXT) && Type0 != int32(WLC_OBJ_BIT_ZEROPAD) && Type0 != int32(WLC_OBJ_BIT_SELECT) && Type0 != int32(WLC_OBJ_ROTATE_R) && Type0 != int32(WLC_OBJ_ROTATE_L)")

				// collect fanins
				vFanins = Vec_IntAlloc(100)
				for i = 1; i < Vec_IntSize(vRoots); i++ {
					Fan = Vec_IntEntry(vRoots, i)
					iObj = Smt_PrsBuild2_rec(pNtk, p, Fan, -1, "")
					if iObj == 0 {
						Vec_IntFree(vFanins)
						return 0
					}
					Vec_IntPush(vFanins, iObj)
				}
				// find range
				Range = 0
				if Type0 >= int32(WLC_OBJ_LOGIC_NOT) && Type0 <= int32(WLC_OBJ_REDUCT_XOR) {
					Range = 1
				} else if Type0 == int32(WLC_OBJ_BIT_CONCAT) {
					for i = 0; i < Vec_IntSize(vFanins); i++ {
						Fan = Vec_IntEntry(vFanins, i)
						Range += Wlc_ObjRange(Wlc_NtkObj(pNtk, Fan))
					}

				} else if Type0 == int32(WLC_OBJ_MUX) {
					var pArray []int32 = Vec_IntArray(vFanins)
					assert(Vec_IntSize(vFanins) == 3, "Vec_IntSize(vFanins) == 3")
					pArray[1], pArray[2] = pArray[2], pArray[1]
					iObj = Vec_IntEntry(vFanins, 1)
					Range = Wlc_ObjRange(Wlc_NtkObj(pNtk, iObj))

				} else { // to determine range, look at the first argument

					iObj = Vec_IntEntry(vFanins, 0)
					Range = Wlc_ObjRange(Wlc_NtkObj(pNtk, iObj))
				}
				// create node
				assert(Range > 0, "Range > 0")
				iObj = Smt_PrsCreateNode(pNtk, Type0, fSigned, Range, vFanins, smtNameOrGenerated(p, pName))
				Vec_IntFree(vFanins)
				return iObj
			}

		} else if Vec_IntSize(vRoots) == 2 { // could be   ((_ extract 48 16) (bvmul ?v_5 ?v_12))

			var iObjPrev int32 = Smt_PrsBuild2_rec(pNtk, p, Vec_IntEntry(vRoots, 1), -1, "")
			if iObjPrev == 0 {
				return 0
			}
			return Smt_PrsBuild2_rec(pNtk, p, Vec_IntEntry(vRoots, 0), iObjPrev, pName)

		} else {
			assert(false, "0")
		} // return Smt_PrsBuild2_rec( pNtk, p, iRoot0 );
		return 0
	}
}
func Smt_PrsBuild2(p *Smt_Prs_t) *Wlc_Ntk_t {
	var pNtk *Wlc_Ntk_t
	var vFansRoot, vFans, vFans2 *Vec_Int_t
	var vAsserts *Vec_Int_t = Vec_IntAlloc(100)
	var i, Root, Fan, iObj, NameId, Range, nBits int32
	var pName, pRange string
	// start network and create primary inputs
	pNtk = Wlc_NtkAlloc(p.pName, 1000)
	pNtk.pManName = Abc_NamStart(1000, 24)
	pNtk.fSmtLib = 1
	// collect top-level asserts
	vFansRoot = Vec_WecEntry(&p.vObjs, 0)
	for i = 0; i < Vec_IntSize(vFansRoot); i++ {
		Root = Vec_IntEntry(vFansRoot, i)
		assert(!(Smt_EntryIsName(Root) != 0), "!Smt_EntryIsName(Root)")
		vFans = Smt_VecEntryNode(p, vFansRoot, i)
		Fan = Vec_IntEntry(vFans, 0)
		assert((Smt_EntryIsName(Fan) != 0), "Smt_EntryIsName(Fan)")
		// create variables
		if Abc_Lit2Var(Fan) == int32(SMT_PRS_DECLARE_FUN) {
			var pName_glb string
			assert(Vec_IntSize(vFans) == 4, "Vec_IntSize(vFans) == 4")
			assert((Smt_VecEntryIsType(vFans, 0, SMT_PRS_DECLARE_FUN) != 0), "Smt_VecEntryIsType(vFans, 0, SMT_PRS_DECLARE_FUN)")
			// get name
			Fan = Vec_IntEntry(vFans, 1)
			assert((Smt_EntryIsName(Fan) != 0), "Smt_EntryIsName(Fan)")
			pName = Smt_EntryName(p, Fan)
			// added: giving a global suffix
			pName_glb = pName + SMT_GLO_SUFFIX
			// FIXME: delete memory of pName
			pName = pName_glb
			// skip ()
			Fan = Vec_IntEntry(vFans, 2)
			assert(!(Smt_EntryIsName(Fan) != 0), "!Smt_EntryIsName(Fan)")
			// check type (Bool or BitVec)
			Fan = Vec_IntEntry(vFans, 3)
			if Smt_EntryIsName(Fan) != 0 { //(declare-fun s1 () Bool)
				assert(!(int32(strings.Compare("Bool", Smt_VecEntryName(p, vFans, 3))) != 0), "!strcmp(\"Bool\", Smt_VecEntryName(p, vFans, 3))")
				Range = 1

			} else { // (declare-fun s1 () (_ BitVec 64))
				Fan = Vec_IntEntry(vFans, 3)
				assert(!(Smt_EntryIsName(Fan) != 0), "!Smt_EntryIsName(Fan)")
				vFans2 = Smt_EntryNode(p, Fan)
				assert(Vec_IntSize(vFans2) == 3, "Vec_IntSize(vFans2) == 3")
				assert(!(int32(strings.Compare("_", Smt_VecEntryName(p, vFans2, 0))) != 0), "!strcmp(\"_\", Smt_VecEntryName(p, vFans2, 0))")
				assert(!(int32(strings.Compare("BitVec", Smt_VecEntryName(p, vFans2, 1))) != 0), "!strcmp(\"BitVec\", Smt_VecEntryName(p, vFans2, 1))")
				Fan = Vec_IntEntry(vFans2, 2)
				assert((Smt_EntryIsName(Fan) != 0), "Smt_EntryIsName(Fan)")
				pRange = Smt_EntryName(p, Fan)
				Range = atoi(pRange)
			}
			// create node
			iObj = Wlc_ObjAlloc(pNtk, int32(WLC_OBJ_PI), 0, Range-1, 0)
			NameId = Abc_NamStrFindOrAdd(pNtk.pManName, pName, nil)
			assert(iObj == NameId, "iObj == NameId")
			// save values
			Vec_IntPush(&pNtk.vValues, NameId)
			Vec_IntPush(&pNtk.vValues, nBits)
			Vec_IntPush(&pNtk.vValues, Range)
			nBits += Range
			pName_glb = ""

			// create constants
			/*
			   else if ( Abc_Lit2Var(Fan) == int32(SMT_PRS_DEFINE_FUN) ) // added: we parse DEFINE_FUN in LET
			   {
			       assert( Vec_IntSize(vFans) == 5 );
			       assert( Smt_VecEntryIsType(vFans, 0, SMT_PRS_DEFINE_FUN) );
			       // get name
			       Fan = Vec_IntEntry(vFans, 1);
			       assert( Smt_EntryIsName(Fan) );
			       pName = Smt_EntryName(p, Fan);

			       // added: giving a global suffix
			       char * pName_glb = pName + SMT_GLO_SUFFIX;
			       // FIXME: delete memory of pName
			       pName = pName_glb;

			       // skip ()
			       Fan = Vec_IntEntry(vFans, 2);
			       assert( !Smt_EntryIsName(Fan) );
			       // check type (Bool or BitVec)
			       Fan = Vec_IntEntry(vFans, 3);
			       if ( Smt_EntryIsName(Fan) )
			       {
			           // (define-fun s_2 () Bool false)
			           assert( !strcmp("Bool", Smt_VecEntryName(p, vFans, 3)) );
			           Range = 1;
			           pValue = Smt_VecEntryName(p, vFans, 4);

			           //printf("value: %s\n",pValue);

			           if ( !strcmp("false", pValue) )
			               pValue = "#b0";
			           else if ( !strcmp("true", pValue) )
			               pValue = "#b1";
			           else assert( 0 );
			           Status = Smt_PrsBuildConstant( pNtk, pValue, Range, pName );
			       }
			       else
			       {
			           // (define-fun s702 () (_ BitVec 4) #xe)
			           // (define-fun s1 () (_ BitVec 8) (bvneg #x7f))
			           // get range
			           Fan = Vec_IntEntry(vFans, 3);

			           assert( !Smt_EntryIsName(Fan) );
			           vFans2 = Smt_VecEntryNode(p, vFans, 3);
			           assert( Vec_IntSize(vFans2) == 3 );
			           assert( !strcmp("_", Smt_VecEntryName(p, vFans2, 0)) );
			           assert( !strcmp("BitVec", Smt_VecEntryName(p, vFans2, 1)) );
			           // get range
			           Fan = Vec_IntEntry(vFans2, 2);

			           assert( Smt_EntryIsName(Fan) );
			           pRange = Smt_EntryName(p, Fan);
			           Range = atoi(pRange);

			           // added: can parse functions too
			           Vec_Int_t * vFans3 = Smt_VecEntryNode(p, vFans, 4);
			           Fan = Vec_IntEntry(vFans3, 0);

			           // get constant
			           //Fan = Vec_IntEntry(vFans, 4);

			           //printf("fan3: %s\n",Fan);
			           //printf("fan0: %s\n",Smt_VecEntryName(p, vFans3, 0));
			           //printf("fan1: %s\n",Smt_VecEntryName(p, vFans3, 1));
			           //printf("fan2: %s\n",Smt_VecEntryName(p, vFans3, 2));
			           //printf("fan3: %s\n",Smt_VecEntryName(p, vFans3, 3));

			           Status = Smt_PrsBuildNode( pNtk, p, Fan, Range, pName );
			       }
			       if ( !Status )
			       {
			           Wlc_NtkFree( pNtk ); pNtk = NULL;
			           goto finish;
			       }
			   }
			*/
			// added: new way to parse define-fun
			// create constants
		} else if Abc_Lit2Var(Fan) == int32(SMT_PRS_DEFINE_FUN) {
			var pName_glb string
			// (define-fun def_16001 () Bool (or def_15999 def_16000))
			// (define-fun def_15990 () (_ BitVec 24) (concat def_15988 def_15989))
			assert((Smt_VecEntryIsType(vFans, 0, SMT_PRS_DEFINE_FUN) != 0), "Smt_VecEntryIsType(vFans, 0, SMT_PRS_DEFINE_FUN)")
			assert(Vec_IntSize(vFans) == 5, "Vec_IntSize(vFans) == 5") // const or definition

			// get name
			Fan = Vec_IntEntry(vFans, 1)
			assert((Smt_EntryIsName(Fan) != 0), "Smt_EntryIsName(Fan)")
			pName = Smt_EntryName(p, Fan)
			// added: giving a global suffix
			pName_glb = pName + SMT_GLO_SUFFIX
			// FIXME: delete memory of pName
			pName = pName_glb

			//get range
			Fan = Vec_IntEntry(vFans, 3)
			if Smt_EntryIsName(Fan) != 0 {
				// (define-fun s_2 () Bool false)
				assert(!(int32(strings.Compare("Bool", Smt_VecEntryName(p, vFans, 3))) != 0), "!strcmp(\"Bool\", Smt_VecEntryName(p, vFans, 3))")
				Range = 1

			} else {
				// (define-fun s702 () (_ BitVec 4) #xe)
				// (define-fun s1 () (_ BitVec 8) (bvneg #x7f))
				assert(!(Smt_EntryIsName(Fan) != 0), "!Smt_EntryIsName(Fan)")
				vFans2 = Smt_VecEntryNode(p, vFans, 3)
				assert(Vec_IntSize(vFans2) == 3, "Vec_IntSize(vFans2) == 3")
				assert(!(int32(strings.Compare("_", Smt_VecEntryName(p, vFans2, 0))) != 0), "!strcmp(\"_\", Smt_VecEntryName(p, vFans2, 0))")
				assert(!(int32(strings.Compare("BitVec", Smt_VecEntryName(p, vFans2, 1))) != 0), "!strcmp(\"BitVec\", Smt_VecEntryName(p, vFans2, 1))")
				// get range
				Fan = Vec_IntEntry(vFans2, 2)
				assert((Smt_EntryIsName(Fan) != 0), "Smt_EntryIsName(Fan)")
				pRange = Smt_EntryName(p, Fan)
				Range = atoi(pRange)
			}

			iObj = Smt_PrsBuild2_rec(pNtk, p, Vec_IntEntry(vFans, 4), Range, pName)
			assert((iObj != 0), "iObj")
			pName_glb = ""

			// collect assertion outputs
		} else if Abc_Lit2Var(Fan) == int32(SMT_PRS_ASSERT) {
			//(assert ; no quantifiers
			//   (let ((s2 (bvsge s0 s1)))
			//   (not s2)))
			//(assert (not (= s0 #x00)))
			assert(Vec_IntSize(vFans) == 2, "Vec_IntSize(vFans) == 2")
			assert((Smt_VecEntryIsType(vFans, 0, SMT_PRS_ASSERT) != 0), "Smt_VecEntryIsType(vFans, 0, SMT_PRS_ASSERT)")
			pNtk.nAssert++ // added
			iObj = Smt_PrsBuild2_rec(pNtk, p, Vec_IntEntry(vFans, 1), -1, "")
			if iObj == 0 {
				Wlc_NtkFree(pNtk)
				pNtk = nil
				goto finish
			}
			Vec_IntPush(vAsserts, iObj)

			// issue warnings about unknown dirs
		} else if Abc_Lit2Var(Fan) >= int32(SMT_PRS_END) {
			fmt.Printf("Ignoring directive \"%s\".\n", Smt_EntryName(p, Fan))
		}
	}
	// build AND of asserts
	if Vec_IntSize(vAsserts) == 1 {
		iObj = Smt_PrsCreateNode(pNtk, int32(WLC_OBJ_BUF), 0, 1, vAsserts, "miter")
		// added: 0 asserts
	} else if Vec_IntSize(vAsserts) == 0 {
		iObj = Smt_PrsBuildConstant(pNtk, "#b1", 1, "miter")
	} else {
		iObj = Smt_PrsCreateNode(pNtk, int32(WLC_OBJ_BIT_CONCAT), 0, Vec_IntSize(vAsserts), vAsserts, "")
		Vec_IntFill(vAsserts, 1, iObj)
		iObj = Smt_PrsCreateNode(pNtk, int32(WLC_OBJ_REDUCT_AND), 0, 1, vAsserts, "miter")
	}
	Wlc_ObjSetCo(pNtk, Wlc_NtkObj(pNtk, iObj), 0)
	// create nameIDs
	vFans = Vec_IntStartNatural(Wlc_NtkObjNumMax(pNtk))
	Vec_IntAppend(&pNtk.vNameIds, vFans)
	Vec_IntFree(vFans)
	//Wlc_NtkReport( pNtk, NULL );
finish:
	// cleanup
	Vec_IntFree(vAsserts)
	return pNtk
}

/**Function*************************************************************

  Synopsis    []

  Description []

  SideEffects []

  SeeAlso     []

***********************************************************************/
// create error message
func Smt_PrsErrorSet(p *Smt_Prs_t, pError string, Value int32) int32 {
	if p.ErrorStr[0] == 0 {
		copy(p.ErrorStr[:], pError)
	}
	return Value
}

func Smt_PrsErrorPrint(p *Smt_Prs_t) int32 {
	if p.ErrorStr[0] == 0 {
		return 1
	}
	line := int32(strings.Count(p.pBuffer[:min(p.pCur, len(p.pBuffer))], "\n"))
	fmt.Printf("Line %d: %s\n", line, strings.TrimRight(string(p.ErrorStr[:]), "\x00"))
	return 0
}

func Smt_PrsLoadFile(pFileName string, ppLimit *int) string {
	data, err := os.ReadFile(pFileName)
	if err != nil {
		fmt.Printf("Cannot open input file.\n")
		return ""
	}
	buffer := "\n" + string(data) + "\n"
	*ppLimit = len(buffer)
	return buffer
}

// Smt_PrsRemoveComments counts expression parentheses while ignoring SMT-LIB
// comments, quoted strings, and quoted symbols. The buffer remains immutable;
// Smt_PrsReadLines skips those regions while tokenizing.
func Smt_PrsRemoveComments(pBuffer string, pLimit int) int32 {
	if pLimit < 0 || pLimit > len(pBuffer) {
		pLimit = len(pBuffer)
	}
	var opens, closes int32
	inBar, inString, inComment, escaped := false, false, false, false
	for i := 0; i < pLimit; i++ {
		c := pBuffer[i]
		if inComment {
			if c == '\n' {
				inComment = false
			}
			continue
		}
		if inString {
			if escaped {
				escaped = false
			} else if c == '\\' {
				escaped = true
			} else if c == '"' {
				inString = false
			}
			continue
		}
		if c == ';' && !inBar {
			inComment = true
			continue
		}
		if c == '"' && !inBar {
			inString = true
			continue
		}
		if c == '|' {
			inBar = !inBar
			continue
		}
		if !inBar {
			if c == '(' {
				opens++
			} else if c == ')' {
				closes++
			}
		}
	}
	if opens != closes {
		fmt.Printf("The input SMTLIB file has different number of opening and closing parentheses (%d and %d).\n", opens, closes)
		return 0
	}
	if opens == 0 {
		fmt.Printf("The input SMTLIB file has no opening or closing parentheses.\n")
	}
	return opens
}

func Smt_PrsAlloc(pFileName string, pBuffer string, pLimit int, nObjs int32) *Smt_Prs_t {
	if nObjs == 0 {
		return nil
	}
	if pLimit < 0 || pLimit > len(pBuffer) {
		pLimit = len(pBuffer)
	}
	p := new(Smt_Prs_t)
	p.pName, p.pBuffer, p.pLimit = pFileName, pBuffer, pLimit
	p.pStrs = Abc_NamStart(1000, 24)
	Smt_AddTypes(p.pStrs)
	Vec_IntGrow(&p.vStack, 100)
	Vec_WecGrow(&p.vObjs, nObjs+1)
	return p
}

func Smt_PrsFree(p *Smt_Prs_t) {
	if p == nil {
		return
	}
	if p.pStrs != nil {
		Abc_NamStop(p.pStrs)
	}
	Vec_IntErase(&p.vStack)
	Vec_IntErase(&p.vTempFans)
	Vec_WecErase(&p.vObjs)
}

func Smt_PrsIsSpace(c byte) int32 {
	return boolToInt32(c == ' ' || c == '\t' || c == '\r' || c == '\n')
}

func Smt_PrsSkipSpaces(p *Smt_Prs_t) {
	for p.pCur < p.pLimit && Smt_PrsIsSpace(p.pBuffer[p.pCur]) != 0 {
		p.pCur++
	}
}

func Smt_PrsSkipNonSpaces(p *Smt_Prs_t) {
	for p.pCur < p.pLimit && Smt_PrsIsSpace(p.pBuffer[p.pCur]) == 0 && p.pBuffer[p.pCur] != '(' && p.pBuffer[p.pCur] != ')' && p.pBuffer[p.pCur] != ';' && p.pBuffer[p.pCur] != '"' {
		p.pCur++
	}
}

func Smt_PrsReadLines(p *Smt_Prs_t) {
	assert(Vec_IntSize(&p.vStack) == 0, "empty parser stack")
	assert(Vec_WecSize(&p.vObjs) == 0, "empty parser objects")
	Vec_IntPush(&p.vStack, Vec_WecSize(&p.vObjs))
	Vec_WecPushLevel(&p.vObjs)
	p.pCur = 0
	firstToken := true
	for p.pCur < p.pLimit {
		Smt_PrsSkipSpaces(p)
		if p.pCur >= p.pLimit {
			break
		}
		c := p.pBuffer[p.pCur]
		if c == ';' { // SMT-LIB line comment
			for p.pCur < p.pLimit && p.pBuffer[p.pCur] != '\n' {
				p.pCur++
			}
			continue
		}
		if c == '"' { // metadata string literals are discarded by the C parser
			p.pCur++
			for p.pCur < p.pLimit {
				if p.pBuffer[p.pCur] == '\\' && p.pCur+1 < p.pLimit {
					p.pCur += 2
					continue
				}
				if p.pBuffer[p.pCur] == '"' {
					p.pCur++
					break
				}
				p.pCur++
			}
			continue
		}
		if firstToken && c == '|' { // leading source metadata block
			firstToken = false
			p.pCur++
			for p.pCur < p.pLimit && p.pBuffer[p.pCur] != '|' {
				p.pCur++
			}
			if p.pCur < p.pLimit {
				p.pCur++
			}
			continue
		}
		if c == '(' {
			Vec_IntPush(Vec_WecEntry(&p.vObjs, Vec_IntEntryLast(&p.vStack)), Abc_Var2Lit(Vec_WecSize(&p.vObjs), 0))
			Vec_IntPush(&p.vStack, Vec_WecSize(&p.vObjs))
			Vec_WecPushLevel(&p.vObjs)
			p.pCur++
			continue
		}
		if c == ')' {
			if Vec_IntSize(&p.vStack) > 1 {
				Vec_IntPop(&p.vStack)
			}
			p.pCur++
			continue
		}
		start := p.pCur
		Smt_PrsSkipNonSpaces(p)
		if p.pCur == start {
			p.pCur++
			continue
		}
		if firstToken {
			firstToken = false
		}
		token := p.pBuffer[start:p.pCur]
		id := Abc_NamStrFindOrAdd(p.pStrs, token, nil)
		Vec_IntPush(Vec_WecEntry(&p.vObjs, Vec_IntEntryLast(&p.vStack)), Abc_Var2Lit(id, 1))
	}
	assert(Vec_IntSize(&p.vStack) == 1, "parser stack has top level")
	assert(Vec_WecSize(&p.vObjs) == Vec_WecCap(&p.vObjs), "object count matches capacity")
	p.nDigits = Abc_Base16Log(uint32(Vec_WecSize(&p.vObjs)))
}

func Smt_PrsPrintParser_rec(p *Smt_Prs_t, iObj int32, Depth int32) {
	var vFans *Vec_Int_t
	var i, Fan int32
	fmt.Printf("%*s(\n", Depth, "")
	vFans = Vec_WecEntry(&p.vObjs, iObj)
	for i = 0; i < Vec_IntSize(vFans); i++ {
		Fan = Vec_IntEntry(vFans, i)
		if Abc_LitIsCompl(Fan) != 0 {
			fmt.Printf("%*s", Depth+2, "")
			fmt.Printf("%s\n", Abc_NamStr(p.pStrs, Abc_Lit2Var(Fan)))

		} else {
			Smt_PrsPrintParser_rec(p, Abc_Lit2Var(Fan), Depth+4)
		}
	}
	fmt.Printf("%*s)\n", Depth, "")
}
func Smt_PrsPrintParser(p *Smt_Prs_t) {
	//    Vec_WecPrint( &p->vDepth, 0 );
	Smt_PrsPrintParser_rec(p, 0, 0)
}

/*
*Function*************************************************************

	Synopsis    []

	Description []

	SideEffects []

	SeeAlso     []

**********************************************************************
*/
func Wlc_ReadSmtBuffer(pFileName string, pBuffer string, pLimit int, fOldParser int32, fPrintTree int32) *Wlc_Ntk_t {
	var pNtk *Wlc_Ntk_t = nil
	var nCount int32 = Smt_PrsRemoveComments(pBuffer, pLimit)
	var p *Smt_Prs_t = Smt_PrsAlloc(pFileName, pBuffer, pLimit, nCount)
	if p == nil {
		return nil
	}
	Smt_PrsReadLines(p)
	if fPrintTree != 0 {
		Smt_PrsPrintParser(p)
	}
	if Smt_PrsErrorPrint(p) != 0 {
		if fOldParser != 0 {
			pNtk = Smt_PrsBuild(p)
		} else {
			pNtk = Smt_PrsBuild2(p)
		}
	}
	Smt_PrsFree(p)
	return pNtk
}
func Wlc_ReadSmt(pFileName string, fOldParser int32, fPrintTree int32) *Wlc_Ntk_t {
	var pNtk *Wlc_Ntk_t = nil
	var pBuffer string
	var pLimit int
	pBuffer = Smt_PrsLoadFile(pFileName, &pLimit)
	if pBuffer == "" {
		return nil
	}
	pNtk = Wlc_ReadSmtBuffer(pFileName, pBuffer, pLimit, fOldParser, fPrintTree)
	return pNtk
}

////////////////////////////////////////////////////////////////////////
///                       END OF FILE                                ///
////////////////////////////////////////////////////////////////////////

//ABC_NAMESPACE_IMPL_END
