package bac

import (
	"fmt"
	"strings"
	"unsafe"

	"github.com/glycerine/abcgo/bool/deco"
	"github.com/glycerine/abcgo/misc/util"
	"github.com/glycerine/abcgo/misc/vec"
)

/**CFile****************************************************************

  FileName    [bac.h]

  SystemName  [ABC: Logic synthesis and verification system.]

  PackageName [Hierarchical word-level netlist.]

  Synopsis    [External declarations.]

  Author      [Alan Mishchenko]

  Affiliation [UC Berkeley]

  Date        [Ver. 1.0. Started - November 29, 2014.]

  Revision    [$Id: bac.h,v 1.00 2014/11/29 00:00:00 alanmi Exp $]

***********************************************************************/

////////////////////////////////////////////////////////////////////////
///                          INCLUDES                                ///
////////////////////////////////////////////////////////////////////////

//#include "aig/gia/gia.h"
//#include "misc/extra/extra.h"
//#include "misc/util/utilNam.h"

////////////////////////////////////////////////////////////////////////
///                         PARAMETERS                               ///
////////////////////////////////////////////////////////////////////////

//ABC_NAMESPACE_HEADER_START

////////////////////////////////////////////////////////////////////////
///                         BASIC TYPES                              ///
////////////////////////////////////////////////////////////////////////

// network objects
type Bac_ObjType_t int32

const (
	BAC_OBJ_NONE Bac_ObjType_t = 0
	// 0:  unused
	BAC_OBJ_PI Bac_ObjType_t = BAC_OBJ_NONE + 1
	// 1:  input
	BAC_OBJ_PO Bac_ObjType_t = BAC_OBJ_PI + 1
	// 2:  output
	BAC_OBJ_BI Bac_ObjType_t = BAC_OBJ_PO + 1
	// 3:  box input
	BAC_OBJ_BO Bac_ObjType_t = BAC_OBJ_BI + 1
	// 4:  box output
	BAC_OBJ_BOX Bac_ObjType_t = BAC_OBJ_BO + 1
	// 5:  box

	BAC_BOX_CF Bac_ObjType_t = BAC_OBJ_BOX + 1

	BAC_BOX_CT Bac_ObjType_t = BAC_BOX_CF + 1

	BAC_BOX_CX Bac_ObjType_t = BAC_BOX_CT + 1

	BAC_BOX_CZ Bac_ObjType_t = BAC_BOX_CX + 1

	BAC_BOX_BUF Bac_ObjType_t = BAC_BOX_CZ + 1

	BAC_BOX_INV Bac_ObjType_t = BAC_BOX_BUF + 1

	BAC_BOX_AND Bac_ObjType_t = BAC_BOX_INV + 1

	BAC_BOX_NAND Bac_ObjType_t = BAC_BOX_AND + 1

	BAC_BOX_OR Bac_ObjType_t = BAC_BOX_NAND + 1

	BAC_BOX_NOR Bac_ObjType_t = BAC_BOX_OR + 1

	BAC_BOX_XOR Bac_ObjType_t = BAC_BOX_NOR + 1

	BAC_BOX_XNOR Bac_ObjType_t = BAC_BOX_XOR + 1

	BAC_BOX_SHARP Bac_ObjType_t = BAC_BOX_XNOR + 1

	BAC_BOX_SHARPL Bac_ObjType_t = BAC_BOX_SHARP + 1

	BAC_BOX_MUX Bac_ObjType_t = BAC_BOX_SHARPL + 1

	BAC_BOX_MAJ Bac_ObjType_t = BAC_BOX_MUX + 1

	BAC_BOX_RAND Bac_ObjType_t = BAC_BOX_MAJ + 1

	BAC_BOX_RNAND Bac_ObjType_t = BAC_BOX_RAND + 1

	BAC_BOX_ROR Bac_ObjType_t = BAC_BOX_RNAND + 1

	BAC_BOX_RNOR Bac_ObjType_t = BAC_BOX_ROR + 1

	BAC_BOX_RXOR Bac_ObjType_t = BAC_BOX_RNOR + 1

	BAC_BOX_RXNOR Bac_ObjType_t = BAC_BOX_RXOR + 1

	BAC_BOX_LAND Bac_ObjType_t = BAC_BOX_RXNOR + 1

	BAC_BOX_LNAND Bac_ObjType_t = BAC_BOX_LAND + 1

	BAC_BOX_LOR Bac_ObjType_t = BAC_BOX_LNAND + 1

	BAC_BOX_LNOR Bac_ObjType_t = BAC_BOX_LOR + 1

	BAC_BOX_LXOR Bac_ObjType_t = BAC_BOX_LNOR + 1

	BAC_BOX_LXNOR Bac_ObjType_t = BAC_BOX_LXOR + 1

	BAC_BOX_NMUX Bac_ObjType_t = BAC_BOX_LXNOR + 1

	BAC_BOX_SEL Bac_ObjType_t = BAC_BOX_NMUX + 1

	BAC_BOX_PSEL Bac_ObjType_t = BAC_BOX_SEL + 1

	BAC_BOX_ENC Bac_ObjType_t = BAC_BOX_PSEL + 1

	BAC_BOX_PENC Bac_ObjType_t = BAC_BOX_ENC + 1

	BAC_BOX_DEC Bac_ObjType_t = BAC_BOX_PENC + 1

	BAC_BOX_EDEC Bac_ObjType_t = BAC_BOX_DEC + 1

	BAC_BOX_ADD Bac_ObjType_t = BAC_BOX_EDEC + 1

	BAC_BOX_SUB Bac_ObjType_t = BAC_BOX_ADD + 1

	BAC_BOX_MUL Bac_ObjType_t = BAC_BOX_SUB + 1

	BAC_BOX_DIV Bac_ObjType_t = BAC_BOX_MUL + 1

	BAC_BOX_MOD Bac_ObjType_t = BAC_BOX_DIV + 1

	BAC_BOX_REM Bac_ObjType_t = BAC_BOX_MOD + 1

	BAC_BOX_POW Bac_ObjType_t = BAC_BOX_REM + 1

	BAC_BOX_MIN Bac_ObjType_t = BAC_BOX_POW + 1

	BAC_BOX_ABS Bac_ObjType_t = BAC_BOX_MIN + 1

	BAC_BOX_LTHAN Bac_ObjType_t = BAC_BOX_ABS + 1

	BAC_BOX_LETHAN Bac_ObjType_t = BAC_BOX_LTHAN + 1

	BAC_BOX_METHAN Bac_ObjType_t = BAC_BOX_LETHAN + 1

	BAC_BOX_MTHAN Bac_ObjType_t = BAC_BOX_METHAN + 1

	BAC_BOX_EQU Bac_ObjType_t = BAC_BOX_MTHAN + 1

	BAC_BOX_NEQU Bac_ObjType_t = BAC_BOX_EQU + 1

	BAC_BOX_SHIL Bac_ObjType_t = BAC_BOX_NEQU + 1

	BAC_BOX_SHIR Bac_ObjType_t = BAC_BOX_SHIL + 1

	BAC_BOX_ROTL Bac_ObjType_t = BAC_BOX_SHIR + 1

	BAC_BOX_ROTR Bac_ObjType_t = BAC_BOX_ROTL + 1

	BAC_BOX_GATE Bac_ObjType_t = BAC_BOX_ROTR + 1

	BAC_BOX_LUT Bac_ObjType_t = BAC_BOX_GATE + 1

	BAC_BOX_ASSIGN Bac_ObjType_t = BAC_BOX_LUT + 1

	BAC_BOX_TRI Bac_ObjType_t = BAC_BOX_ASSIGN + 1

	BAC_BOX_RAM Bac_ObjType_t = BAC_BOX_TRI + 1

	BAC_BOX_RAMR Bac_ObjType_t = BAC_BOX_RAM + 1

	BAC_BOX_RAMW Bac_ObjType_t = BAC_BOX_RAMR + 1

	BAC_BOX_RAMWC Bac_ObjType_t = BAC_BOX_RAMW + 1

	BAC_BOX_RAMBOX Bac_ObjType_t = BAC_BOX_RAMWC + 1

	BAC_BOX_LATCH Bac_ObjType_t = BAC_BOX_RAMBOX + 1

	BAC_BOX_LATCHRS Bac_ObjType_t = BAC_BOX_LATCH + 1

	BAC_BOX_DFF Bac_ObjType_t = BAC_BOX_LATCHRS + 1

	BAC_BOX_DFFRS Bac_ObjType_t = BAC_BOX_DFF + 1

	BAC_BOX_UNKNOWN Bac_ObjType_t = BAC_BOX_DFFRS + 1 // 67

)

// name types
type Bac_NameType_t int32

const (
	BAC_NAME_BIN Bac_NameType_t = 0
	// 0:  binary variable
	BAC_NAME_WORD Bac_NameType_t = BAC_NAME_BIN + 1
	// 1:  first bit of word-level variable
	BAC_NAME_INFO Bac_NameType_t = BAC_NAME_WORD + 1
	// 2:  first bit of special variable
	BAC_NAME_INDEX Bac_NameType_t = BAC_NAME_INFO + 1
	// 3:  index of word-level variable

)

type Bac_Ntk_t = Bac_Ntk_t_
type Bac_Man_t = Bac_Man_t_

// network
type Bac_Ntk_t_ struct {
	pDesign *Bac_Man_t // design
	ntkID   int32      // index in the design network array
	NameId  int32      // name ID
	iCopy   int32      // copy module
	iBoxNtk int32      // instance network ID
	iBoxObj int32      // instance object ID
	Count   int32      // object counter
	Mark    int32      // visit mark
	// interface
	vInputs  vec.Vec_Int_t // inputs
	vOutputs vec.Vec_Int_t // outputs
	vInfo    vec.Vec_Int_t // input/output/wire info
	// object attributes
	vType   vec.Vec_Str_t // types
	vFanin  vec.Vec_Int_t // fanin
	vIndex  vec.Vec_Int_t // index
	vName   vec.Vec_Int_t // original NameId or InstId
	vFanout vec.Vec_Int_t // fanout
	vCopy   vec.Vec_Int_t // copy
	// other
	vArray  vec.Vec_Int_t
	vArray2 vec.Vec_Int_t
}

// design
type Bac_Man_t_ struct {
	// design names
	pName string     // design name
	pSpec string     // spec file name
	pStrs *Abc_Nam_t // string manager
	pMods *Abc_Nam_t // module name manager
	// internal data
	iRoot int32       // root network
	nNtks int32       // number of current networks
	pNtks []Bac_Ntk_t // networks
	// user data
	vOut         *vec.Vec_Str_t
	vOut2        *vec.Vec_Str_t
	vBuf2RootNtk vec.Vec_Int_t
	vBuf2RootObj vec.Vec_Int_t
	vBuf2LeafNtk vec.Vec_Int_t
	vBuf2LeafObj vec.Vec_Int_t
	pMioLib      any
	ppGraphs     []*deco.Dec_Graph_t
	ElemGates    [4]int32
	pPrimNames   [BAC_BOX_UNKNOWN]string
	pPrimSymbs   [BAC_BOX_UNKNOWN]string
}

func Bac_ManName(p *Bac_Man_t) string   { return p.pName }
func Bac_ManSpec(p *Bac_Man_t) string   { return p.pSpec }
func Bac_ManNtkNum(p *Bac_Man_t) int32  { return p.nNtks }
func Bac_ManPrimNum(p *Bac_Man_t) int32 { return util.Abc_NamObjNumMax(p.pMods) - Bac_ManNtkNum(p) }
func Bac_ManNtkIsOk(p *Bac_Man_t, i int32) int32 {
	if i > 0 && i <= Bac_ManNtkNum(p) {
		return 1
	}
	return 0
}
func Bac_ManNtk(p *Bac_Man_t, i int32) *Bac_Ntk_t {
	if Bac_ManNtkIsOk(p, i) != 0 {
		return &p.pNtks[i]
	} else {
		return nil
	}
}
func Bac_ManNtkFindId(p *Bac_Man_t, pName string) int32 { return util.Abc_NamStrFind(p.pMods, pName) }
func Bac_ManNtkFind(p *Bac_Man_t, pName string) *Bac_Ntk_t {
	return Bac_ManNtk(p, Bac_ManNtkFindId(p, pName))
}
func Bac_ManRoot(p *Bac_Man_t) *Bac_Ntk_t                     { return Bac_ManNtk(p, p.iRoot) }
func Bac_ManStr(p *Bac_Man_t, i int32) string                 { return util.Abc_NamStr(p.pStrs, i) }
func Bac_ManStrId(p *Bac_Man_t, pStr string) int32            { return util.Abc_NamStrFind(p.pStrs, pStr) }
func Bac_ManPrimName(p *Bac_Man_t, Type Bac_ObjType_t) string { return p.pPrimNames[Type] }
func Bac_ManPrimSymb(p *Bac_Man_t, Type Bac_ObjType_t) string { return p.pPrimSymbs[Type] }

func Bac_NtkId(p *Bac_Ntk_t) int32 {
	i := p.ntkID
	assert(Bac_ManNtkIsOk(p.pDesign, i) != 0 && &p.pDesign.pNtks[i] == p, "network belongs to design")
	return i
}
func Bac_NtkMan(p *Bac_Ntk_t) *Bac_Man_t { return p.pDesign }
func Bac_NtkNameId(p *Bac_Ntk_t) int32   { return p.NameId }
func Bac_NtkName(p *Bac_Ntk_t) string    { return Bac_ManStr(p.pDesign, Bac_NtkNameId(p)) }
func Bac_NtkCopy(p *Bac_Ntk_t) int32     { return p.iCopy }
func Bac_NtkCopyNtk(pNew *Bac_Man_t, p *Bac_Ntk_t) *Bac_Ntk_t {
	return Bac_ManNtk(pNew, Bac_NtkCopy(p))
}
func Bac_NtkSetCopy(p *Bac_Ntk_t, i int32) { assert(p.iCopy == -1, "p.iCopy == -1"); p.iCopy = i }

func Bac_NtkObjNum(p *Bac_Ntk_t) int32      { return int32(vec.Vec_StrSize(&p.vType)) }
func Bac_NtkObjNumAlloc(p *Bac_Ntk_t) int32 { return int32(vec.Vec_StrCap(&p.vType)) }
func Bac_NtkPiNum(p *Bac_Ntk_t) int32       { return int32(vec.Vec_IntSize(&p.vInputs)) }
func Bac_NtkPoNum(p *Bac_Ntk_t) int32       { return int32(vec.Vec_IntSize(&p.vOutputs)) }
func Bac_NtkPioNum(p *Bac_Ntk_t) int32      { return Bac_NtkPiNum(p) + Bac_NtkPoNum(p) }
func Bac_NtkPiNumAlloc(p *Bac_Ntk_t) int32  { return int32(vec.Vec_IntCap(&p.vInputs)) }
func Bac_NtkPoNumAlloc(p *Bac_Ntk_t) int32  { return int32(vec.Vec_IntCap(&p.vOutputs)) }
func Bac_NtkBiNum(p *Bac_Ntk_t) int32 {
	return int32(vec.Vec_StrCountEntryLit(&p.vType, byte(BAC_OBJ_BI)))
}
func Bac_NtkBoNum(p *Bac_Ntk_t) int32 {
	return int32(vec.Vec_StrCountEntryLit(&p.vType, byte(BAC_OBJ_BO)))
}
func Bac_NtkCiNum(p *Bac_Ntk_t) int32 { return Bac_NtkPiNum(p) + Bac_NtkBoNum(p) }
func Bac_NtkCoNum(p *Bac_Ntk_t) int32 { return Bac_NtkPoNum(p) + Bac_NtkBiNum(p) }
func Bac_NtkBoxNum(p *Bac_Ntk_t) int32 {
	return Bac_NtkObjNum(p) - int32(vec.Vec_StrCountSmallerLit(&p.vType, byte(BAC_OBJ_BOX)))
}
func Bac_NtkPrimNum(p *Bac_Ntk_t) int32 {
	return int32(vec.Vec_StrCountLargerLit(&p.vType, byte(BAC_OBJ_BOX)))
}
func Bac_NtkUserNum(p *Bac_Ntk_t) int32 {
	return int32(vec.Vec_StrCountEntryLit(&p.vType, byte(BAC_OBJ_BOX)))
}

func Bac_NtkPi(p *Bac_Ntk_t, i int32) int32   { return vec.Vec_IntEntry(&p.vInputs, int(i)) }
func Bac_NtkPo(p *Bac_Ntk_t, i int32) int32   { return vec.Vec_IntEntry(&p.vOutputs, int(i)) }
func Bac_NtkStr(p *Bac_Ntk_t, i int32) string { return Bac_ManStr(p.pDesign, i) }
func Bac_NtkHostNtk(p *Bac_Ntk_t) *Bac_Ntk_t {
	if p.iBoxNtk > 0 {
		return Bac_ManNtk(p.pDesign, p.iBoxNtk)
	} else {
		return nil
	}
}
func Bac_NtkHostObj(p *Bac_Ntk_t) int32 { return p.iBoxObj }
func Bac_NtkSetHost(p *Bac_Ntk_t, n int32, i int32) {
	assert(p.iBoxNtk == -1, "p.iBoxNtk == -1")
	p.iBoxNtk = n
	p.iBoxObj = i
}

func Bac_InfoRange(Beg int32, End int32) int32 {
	if End > Beg {
		return End - Beg + 1
	} else {
		return Beg - End + 1
	}
}
func Bac_NtkInfoNum(p *Bac_Ntk_t) int32      { return int32(vec.Vec_IntSize(&p.vInfo)) / 3 }
func Bac_NtkInfoNumAlloc(p *Bac_Ntk_t) int32 { return int32(vec.Vec_IntCap(&p.vInfo)) / 3 }
func Bac_NtkInfoType(p *Bac_Ntk_t, i int32) int32 {
	return util.Abc_Lit2Att2(vec.Vec_IntEntry(&p.vInfo, int(3*i)))
}
func Bac_NtkInfoName(p *Bac_Ntk_t, i int32) int32 {
	return util.Abc_Lit2Var2(vec.Vec_IntEntry(&p.vInfo, int(3*i)))
}
func Bac_NtkInfoBeg(p *Bac_Ntk_t, i int32) int32 { return vec.Vec_IntEntry(&p.vInfo, int(3*i+1)) }
func Bac_NtkInfoEnd(p *Bac_Ntk_t, i int32) int32 { return vec.Vec_IntEntry(&p.vInfo, int(3*i+2)) }
func Bac_NtkInfoRange(p *Bac_Ntk_t, i int32) int32 {
	a := vec.Vec_IntArray(&p.vInfo)[int(3*i):]
	if a[1] >= 0 {
		return Bac_InfoRange(a[1], a[2])
	} else {
		return 1
	}
}
func Bac_NtkInfoIndex(p *Bac_Ntk_t, i int32, j int32) int32 {
	a := vec.Vec_IntArray(&p.vInfo)[int(3*i):]
	assert(a[1] >= 0, "a[1]>=0")
	if a[1] < a[2] {
		return a[1] + j
	} else {
		return a[1] - j
	}
}
func Bac_NtkAddInfo(p *Bac_Ntk_t, i int32, b int32, e int32) {
	Vec_IntPush(&p.vInfo, i)
	Vec_IntPushTwo(&p.vInfo, b, e)
}
func Bac_NtkSetInfoName(p *Bac_Ntk_t, i int32, n int32) { Vec_IntWriteEntry(&p.vInfo, 3*i, n) }

func Bac_NtkStartNames(p *Bac_Ntk_t) {
	assert((Bac_NtkObjNumAlloc(p) != 0), "Bac_NtkObjNumAlloc(p)")
	Vec_IntFill(&p.vName, Bac_NtkObjNumAlloc(p), 0)
}
func Bac_NtkStartFanouts(p *Bac_Ntk_t) {
	assert((Bac_NtkObjNumAlloc(p) != 0), "Bac_NtkObjNumAlloc(p)")
	Vec_IntFill(&p.vFanout, Bac_NtkObjNumAlloc(p), 0)
}
func Bac_NtkStartCopies(p *Bac_Ntk_t) {
	assert((Bac_NtkObjNumAlloc(p) != 0), "Bac_NtkObjNumAlloc(p)")
	Vec_IntFill(&p.vCopy, Bac_NtkObjNumAlloc(p), -1)
}
func Bac_NtkFreeNames(p *Bac_Ntk_t)   { Vec_IntErase(&p.vName) }
func Bac_NtkFreeFanouts(p *Bac_Ntk_t) { Vec_IntErase(&p.vFanout) }
func Bac_NtkFreeCopies(p *Bac_Ntk_t)  { Vec_IntErase(&p.vCopy) }
func Bac_NtkHasNames(p *Bac_Ntk_t) int32 {
	if vec.Vec_IntCap(&p.vName) != 0 {
		return 1
	}
	return 0
}
func Bac_NtkHasFanouts(p *Bac_Ntk_t) int32 {
	if vec.Vec_IntCap(&p.vFanout) != 0 {
		return 1
	}
	return 0
}
func Bac_NtkHasCopies(p *Bac_Ntk_t) int32 {
	if vec.Vec_IntCap(&p.vCopy) != 0 {
		return 1
	}
	return 0
}

func boolToInt32(v bool) int32 {
	if v {
		return 1
	}
	return 0
}
func Bac_TypeIsBox(Type Bac_ObjType_t) int32 {
	return boolToInt32(Type >= BAC_OBJ_BOX && Type < BAC_BOX_UNKNOWN)
}
func Bac_NameType(n int32) Bac_NameType_t {
	assert((n != 0), "n")
	return Bac_NameType_t(Abc_Lit2Att2(n))
}
func Bac_CharIsDigit(c byte) int32 { return boolToInt32(c >= '0' && c <= '9') }

func Bac_ObjType(p *Bac_Ntk_t, i int32) Bac_ObjType_t {
	return Bac_ObjType_t(Abc_Lit2Var(int32(byte(Vec_StrEntry(&p.vType, i)))))
}
func Bac_ObjIsPi(p *Bac_Ntk_t, i int32) int32 { return boolToInt32(Bac_ObjType(p, i) == BAC_OBJ_PI) }
func Bac_ObjIsPo(p *Bac_Ntk_t, i int32) int32 { return boolToInt32(Bac_ObjType(p, i) == BAC_OBJ_PO) }
func Bac_ObjIsPio(p *Bac_Ntk_t, i int32) int32 {
	return boolToInt32(Bac_ObjIsPi(p, i) != 0 || Bac_ObjIsPo(p, i) != 0)
}
func Bac_ObjIsBi(p *Bac_Ntk_t, i int32) int32 { return boolToInt32(Bac_ObjType(p, i) == BAC_OBJ_BI) }
func Bac_ObjIsBo(p *Bac_Ntk_t, i int32) int32 { return boolToInt32(Bac_ObjType(p, i) == BAC_OBJ_BO) }
func Bac_ObjIsBio(p *Bac_Ntk_t, i int32) int32 {
	return boolToInt32(Bac_ObjIsBi(p, i) != 0 || Bac_ObjIsBo(p, i) != 0)
}
func Bac_ObjIsBox(p *Bac_Ntk_t, i int32) int32 { return Bac_TypeIsBox(Bac_ObjType(p, i)) }
func Bac_ObjIsBoxUser(p *Bac_Ntk_t, i int32) int32 {
	return boolToInt32(Bac_ObjType(p, i) == BAC_OBJ_BOX)
}
func Bac_ObjIsBoxPrim(p *Bac_Ntk_t, i int32) int32 {
	return boolToInt32(Bac_ObjIsBox(p, i) != 0 && Bac_ObjIsBoxUser(p, i) == 0)
}
func Bac_ObjIsGate(p *Bac_Ntk_t, i int32) int32 {
	return boolToInt32(Bac_ObjType(p, i) == BAC_BOX_GATE)
}
func Bac_ObjIsCi(p *Bac_Ntk_t, i int32) int32 {
	return boolToInt32(Bac_ObjIsPi(p, i) != 0 || Bac_ObjIsBo(p, i) != 0)
}
func Bac_ObjIsCo(p *Bac_Ntk_t, i int32) int32 {
	return boolToInt32(Bac_ObjIsPo(p, i) != 0 || Bac_ObjIsBi(p, i) != 0)
}
func Bac_ObjIsCio(p *Bac_Ntk_t, i int32) int32 { return boolToInt32(Bac_ObjType(p, i) < BAC_OBJ_BOX) }
func Bac_ObjIsConst(p *Bac_Ntk_t, i int32) int32 {
	return boolToInt32(Bac_ObjType(p, i) >= BAC_BOX_CF && Bac_ObjType(p, i) <= BAC_BOX_CZ)
}
func Bac_ObjIsConstBin(p *Bac_Ntk_t, i int32) int32 {
	return boolToInt32(Bac_ObjType(p, i) == BAC_BOX_CF || Bac_ObjType(p, i) == BAC_BOX_CT)
}

func Bac_ObjBit(p *Bac_Ntk_t, i int32) int32 {
	assert(!(Bac_ObjIsBox(p, i) != 0), "!Bac_ObjIsBox(p, i)")
	return Abc_LitIsCompl(int32(Vec_StrEntry(&p.vType, i)))
}
func Bac_ObjSetBit(p *Bac_Ntk_t, i int32) {
	q := Vec_StrArray(&p.vType)
	assert(!(Bac_ObjIsBox(p, i) != 0), "!Bac_ObjIsBox(p, i)")
	q[int(i)] = byte(Abc_LitNot(int32(q[int(i)])))
}
func Bac_ObjFanin(p *Bac_Ntk_t, i int32) int32 {
	assert((Bac_ObjIsCo(p, i) != 0), "Bac_ObjIsCo(p, i)")
	return Vec_IntEntry(&p.vFanin, i)
}
func Bac_ObjIndex(p *Bac_Ntk_t, i int32) int32 {
	assert((Bac_ObjIsCio(p, i) != 0), "Bac_ObjIsCio(p, i)")
	return Vec_IntEntry(&p.vIndex, i)
}
func Bac_ObjNameInt(p *Bac_Ntk_t, i int32) int32 {
	assert(!(Bac_ObjIsCo(p, i) != 0), "!Bac_ObjIsCo(p, i)")
	return Vec_IntEntry(&p.vName, i)
}
func Bac_ObjName(p *Bac_Ntk_t, i int32) int32 {
	if Bac_ObjIsCo(p, i) != 0 {
		return Bac_ObjNameInt(p, Bac_ObjFanin(p, i))
	} else {
		return Bac_ObjNameInt(p, i)
	}
}
func Bac_ObjNameType(p *Bac_Ntk_t, i int32) Bac_NameType_t { return Bac_NameType(Bac_ObjName(p, i)) }
func Bac_ObjNameId(p *Bac_Ntk_t, i int32) int32            { return Abc_Lit2Var2(Bac_ObjName(p, i)) }
func Bac_ObjNameStr(p *Bac_Ntk_t, i int32) string {
	assert(Bac_ObjNameType(p, i) <= BAC_NAME_WORD, "Bac_ObjNameType(p, i) <= BAC_NAME_WORD")
	return Bac_NtkStr(p, Bac_ObjNameId(p, i))
}
func Bac_ObjCopy(p *Bac_Ntk_t, i int32) int32 { return Vec_IntEntry(&p.vCopy, i) }
func Bac_ObjFanout(p *Bac_Ntk_t, i int32) int32 {
	assert((Bac_ObjIsCi(p, i) != 0), "Bac_ObjIsCi(p, i)")
	return Vec_IntEntry(&p.vFanout, i)
}
func Bac_ObjNextFanout(p *Bac_Ntk_t, i int32) int32 {
	assert((Bac_ObjIsCo(p, i) != 0), "Bac_ObjIsCo(p, i)")
	return Vec_IntEntry(&p.vFanout, i)
}
func Bac_ObjSetFanout(p *Bac_Ntk_t, i int32, x int32) {
	assert((Bac_ObjIsCi(p, i) != 0), "Bac_ObjIsCi(p, i)")
	Vec_IntSetEntry(&p.vFanout, i, x)
}
func Bac_ObjSetNextFanout(p *Bac_Ntk_t, i int32, x int32) {
	assert((Bac_ObjIsCo(p, i) != 0), "Bac_ObjIsCo(p, i)")
	Vec_IntSetEntry(&p.vFanout, i, x)
}
func Bac_ObjCleanFanin(p *Bac_Ntk_t, i int32) {
	assert(Bac_ObjFanin(p, i) >= 0 && (Bac_ObjIsCo(p, i) != 0), "Bac_ObjFanin(p, i) >= 0 && Bac_ObjIsCo(p, i)")
	Vec_IntSetEntry(&p.vFanin, i, -1)
}
func Bac_ObjSetFanin(p *Bac_Ntk_t, i int32, x int32) {
	assert(Bac_ObjFanin(p, i) == -1 && (Bac_ObjIsCo(p, i) != 0), "Bac_ObjFanin(p, i) == -1 && Bac_ObjIsCo(p, i)")
	Vec_IntSetEntry(&p.vFanin, i, x)
}
func Bac_ObjSetIndex(p *Bac_Ntk_t, i int32, x int32) {
	assert(Bac_ObjIndex(p, i) == -1, "Bac_ObjIndex(p, i) == -1")
	Vec_IntSetEntry(&p.vIndex, i, x)
}
func Bac_ObjSetName(p *Bac_Ntk_t, i int32, x int32) {
	assert(Bac_ObjName(p, i) == 0 && !(Bac_ObjIsCo(p, i) != 0), "Bac_ObjName(p, i) == 0 && !Bac_ObjIsCo(p, i)")
	Vec_IntSetEntry(&p.vName, i, x)
}
func Bac_ObjSetCopy(p *Bac_Ntk_t, i int32, x int32) {
	assert(Bac_ObjCopy(p, i) == -1, "Bac_ObjCopy(p, i) == -1")
	Vec_IntSetEntry(&p.vCopy, i, x)
}
func Bac_ObjGetConst(p *Bac_Ntk_t, i int32) int32 {
	assert((Bac_ObjIsCi(p, i) != 0), "Bac_ObjIsCi(p, i)")
	if (Bac_ObjIsBo(p, i) != 0) && (Bac_ObjIsConst(p, i-1) != 0) {
		return int32(Bac_ObjType(p, i-1))
	} else {
		return 0
	}
}

func Bac_BoxBiNum(p *Bac_Ntk_t, i int32) int32 {
	var s int32 = i - 1
	assert((Bac_ObjIsBox(p, i) != 0), "Bac_ObjIsBox(p, i)")
	for i--; i >= 0 && Bac_ObjIsBi(p, i) != 0; i-- {
	}
	return s - i
}
func Bac_BoxBoNum(p *Bac_Ntk_t, i int32) int32 {
	var s int32 = i + 1
	assert((Bac_ObjIsBox(p, i) != 0), "Bac_ObjIsBox(p, i)")
	for i++; i < Bac_NtkObjNum(p) && Bac_ObjIsBo(p, i) != 0; i++ {
	}
	return i - s
}
func Bac_BoxSize(p *Bac_Ntk_t, i int32) int32 { return 1 + Bac_BoxBiNum(p, i) + Bac_BoxBoNum(p, i) }
func Bac_BoxBi(p *Bac_Ntk_t, b int32, i int32) int32 {
	assert((Bac_ObjIsBox(p, b) != 0), "Bac_ObjIsBox(p, b)")
	return b - 1 - i
}
func Bac_BoxBo(p *Bac_Ntk_t, b int32, i int32) int32 {
	assert((Bac_ObjIsBox(p, b) != 0), "Bac_ObjIsBox(p, b)")
	return b + 1 + i
}

// static inline int            Bac_BoxBiBox( Bac_Ntk_t * p, int i )            { assert(Bac_ObjIsBi(p, i)); return i + 1 + Bac_ObjIndex(p, i);                               }
func Bac_BoxBoBox(p *Bac_Ntk_t, i int32) int32 {
	assert((Bac_ObjIsBo(p, i) != 0), "Bac_ObjIsBo(p, i)")
	return i - 1 - Bac_ObjIndex(p, i)
}
func Bac_BoxFanin(p *Bac_Ntk_t, b int32, i int32) int32 { return Bac_ObjFanin(p, Bac_BoxBi(p, b, i)) }
func Bac_BoxFaninBox(p *Bac_Ntk_t, b int32, i int32) int32 {
	return Bac_BoxBoBox(p, Bac_BoxFanin(p, b, i))
}
func Bac_BoxBiRange(p *Bac_Ntk_t, i int32) int32 {
	var s int32 = i
	assert((Bac_ObjIsBi(p, i) != 0) && !(Bac_ObjBit(p, i) != 0), "Bac_ObjIsBi(p, i) && !Bac_ObjBit(p, i)")
	for i--; i >= 0 && Bac_ObjIsBi(p, i) != 0 && Bac_ObjBit(p, i) != 0; i-- {
	}
	return s - i
}
func Bac_BoxBoRange(p *Bac_Ntk_t, i int32) int32 {
	var s int32 = i
	assert((Bac_ObjIsBo(p, i) != 0) && !(Bac_ObjBit(p, i) != 0), "Bac_ObjIsBo(p, i) && !Bac_ObjBit(p, i)")
	for i++; i < Bac_NtkObjNum(p) && Bac_ObjIsBo(p, i) != 0 && Bac_ObjBit(p, i) != 0; i++ {
	}
	return i - s
}
func Bac_ObjPiRange(p *Bac_Ntk_t, i int32) int32 {
	var s int32 = i
	assert((Bac_ObjIsPi(p, i) != 0) && !(Bac_ObjBit(p, i) != 0), "Bac_ObjIsPi(p, i) && !Bac_ObjBit(p, i)")
	for i++; i < Bac_NtkObjNum(p) && Bac_ObjIsPi(p, i) != 0 && Bac_ObjBit(p, i) != 0; i++ {
	}
	return i - s
}

func Bac_BoxNtkId(p *Bac_Ntk_t, i int32) int32 {
	assert((Bac_ObjIsBox(p, i) != 0), "Bac_ObjIsBox(p, i)")
	return Vec_IntEntry(&p.vFanin, i)
}
func Bac_BoxSetNtkId(p *Bac_Ntk_t, i int32, x int32) {
	assert((Bac_ObjIsBox(p, i) != 0) && (Bac_ManNtkIsOk(p.pDesign, x) != 0), "Bac_ObjIsBox(p, i)&&Bac_ManNtkIsOk(p.pDesign, x)")
	Vec_IntSetEntry(&p.vFanin, i, x)
}

// static inline int            Bac_BoxBiNtkId( Bac_Ntk_t * p, int i )          { assert(Bac_ObjIsBi(p, i)); return Bac_BoxNtkId(p, Bac_BoxBiBox(p, i));                      }
func Bac_BoxBoNtkId(p *Bac_Ntk_t, i int32) int32 {
	assert((Bac_ObjIsBo(p, i) != 0), "Bac_ObjIsBo(p, i)")
	return Bac_BoxNtkId(p, Bac_BoxBoBox(p, i))
}
func Bac_BoxNtk(p *Bac_Ntk_t, i int32) *Bac_Ntk_t { return Bac_ManNtk(p.pDesign, Bac_BoxNtkId(p, i)) }

// static inline Bac_Ntk_t *    Bac_BoxBiNtk( Bac_Ntk_t * p, int i )            { return Bac_ManNtk( p->pDesign, Bac_BoxBiNtkId(p, i) );                                      }
func Bac_BoxBoNtk(p *Bac_Ntk_t, i int32) *Bac_Ntk_t {
	return Bac_ManNtk(p.pDesign, Bac_BoxBoNtkId(p, i))
}
func Bac_BoxNtkName(p *Bac_Ntk_t, i int32) string {
	return util.Abc_NamStr(p.pDesign.pMods, Bac_BoxNtkId(p, i))
}

////////////////////////////////////////////////////////////////////////
///                      MACRO DEFINITIONS                           ///
////////////////////////////////////////////////////////////////////////

////////////////////////////////////////////////////////////////////////
///                             ITERATORS                            ///
////////////////////////////////////////////////////////////////////////

// Original macro:
// #define Bac_ManForEachNtk( p, pNtk, i )                                   \
//     for ( i = 1; (i <= Bac_ManNtkNum(p)) && (((pNtk) = Bac_ManNtk(p, i)), 1); i++ )

// Original macro:
// #define Bac_NtkForEachPi( p, iObj, i )                                    \
//     for ( i = 0; (i < Bac_NtkPiNum(p))  && (((iObj) = Bac_NtkPi(p, i)), 1); i++ )

// Original macro:
// #define Bac_NtkForEachPo( p, iObj, i )                                    \
//     for ( i = 0; (i < Bac_NtkPoNum(p))  && (((iObj) = Bac_NtkPo(p, i)), 1); i++ )

// Original macro:
// #define Bac_NtkForEachPoDriver( p, iObj, i )                              \
//     for ( i = 0; (i < Bac_NtkPoNum(p))  && (((iObj) = Bac_ObjFanin(p, Bac_NtkPo(p, i))), 1); i++ )

// Original macro:
// #define Bac_NtkForEachPiMain( p, iObj, i )                                \
//     for ( i = 0; (i < Bac_NtkPiNum(p))  && (((iObj) = Bac_NtkPi(p, i)), 1); i++ )  if ( Bac_ObjBit(p, iObj) ) {} else

// Original macro:
// #define Bac_NtkForEachPoMain( p, iObj, i )                                \
//     for ( i = 0; (i < Bac_NtkPoNum(p))  && (((iObj) = Bac_NtkPo(p, i)), 1); i++ )  if ( Bac_ObjBit(p, iObj) ) {} else

// Original macro:
// #define Bac_NtkForEachObj( p, i )  if ( !Bac_ObjType(p, i) ) {} else      \
//     for ( i = 0; (i < Bac_NtkObjNum(p)); i++ )

// Original macro:
// #define Bac_NtkForEachObjType( p, Type, i )                               \
//     for ( i = 0; (i < Bac_NtkObjNum(p))  && (((Type) = Bac_ObjType(p, i)), 1); i++ ) if ( !Type ) {} else

// Original macro:
// #define Bac_NtkForEachBox( p, i )                                         \
//     for ( i = 0; (i < Bac_NtkObjNum(p)); i++ ) if ( !Bac_ObjIsBox(p, i) ) {} else

// Original macro:
// #define Bac_NtkForEachBoxUser( p, i )                                     \
//     for ( i = 0; (i < Bac_NtkObjNum(p)); i++ ) if ( !Bac_ObjIsBoxUser(p, i) ) {} else

// Original macro:
// #define Bac_NtkForEachBoxPrim( p, i )                                     \
//     for ( i = 0; (i < Bac_NtkObjNum(p)); i++ ) if ( !Bac_ObjIsBoxPrim(p, i) ) {} else

// Original macro:
// #define Bac_NtkForEachCi( p, i )                                          \
//     for ( i = 0; (i < Bac_NtkObjNum(p)); i++ ) if ( !Bac_ObjIsCi(p, i) ) {} else

// Original macro:
// #define Bac_NtkForEachCo( p, i )                                          \
//     for ( i = 0; (i < Bac_NtkObjNum(p)); i++ ) if ( !Bac_ObjIsCo(p, i) ) {} else

// Original macro:
// #define Bac_NtkForEachCio( p, i )                                         \
//     for ( i = 0; (i < Bac_NtkObjNum(p)); i++ ) if ( !Bac_ObjIsCio(p, i) ){} else

// Original macro:
// #define Bac_NtkForEachBi( p, i )                                          \
//     for ( i = 0; (i < Bac_NtkObjNum(p)); i++ ) if ( !Bac_ObjIsBi(p, i) ){} else

// Original macro:
// #define Bac_NtkForEachBo( p, i )                                          \
//     for ( i = 0; (i < Bac_NtkObjNum(p)); i++ ) if ( !Bac_ObjIsBo(p, i) ){} else

// Original macro:
// #define Bac_NtkForEachBio( p, i )                                         \
//     for ( i = 0; (i < Bac_NtkObjNum(p)); i++ ) if ( !Bac_ObjIsBio(p, i) ){} else

// Original macro:
// #define Bac_BoxForEachBi( p, iBox, iTerm, i )                             \
//     for ( iTerm = iBox - 1, i = 0; iTerm >= 0 && Bac_ObjIsBi(p, iTerm); iTerm--, i++ )

// Original macro:
// #define Bac_BoxForEachBo( p, iBox, iTerm, i )                             \
//     for ( iTerm = iBox + 1, i = 0; iTerm < Bac_NtkObjNum(p) && Bac_ObjIsBo(p, iTerm); iTerm++, i++ )

// Original macro:
// #define Bac_BoxForEachBiReverse( p, iBox, iTerm, i )                      \
//     for ( i = Bac_BoxBiNum(p, iBox), iTerm = iBox - i--; Bac_ObjIsBi(p, iTerm); iTerm++, i-- )

// Original macro:
// #define Bac_BoxForEachBiMain( p, iBox, iTerm, i )                         \
//     for ( iTerm = iBox - 1, i = 0; iTerm >= 0 && Bac_ObjIsBi(p, iTerm); iTerm--, i++ )                if ( Bac_ObjBit(p, iTerm) ) {} else

// Original macro:
// #define Bac_BoxForEachBoMain( p, iBox, iTerm, i )                         \
//     for ( iTerm = iBox + 1, i = 0; iTerm < Bac_NtkObjNum(p) && Bac_ObjIsBo(p, iTerm); iTerm++, i++ )  if ( Bac_ObjBit(p, iTerm) ) {} else

// Original macro:
// #define Bac_BoxForEachFanin( p, iBox, iFanin, i )                         \
//     for ( i = 0; iBox - 1 - i >= 0 && Bac_ObjIsBi(p, iBox - 1 - i) && (((iFanin) = Bac_BoxFanin(p, iBox, i)), 1); i++ )

// Original macro:
// #define Bac_BoxForEachFaninBox( p, iBox, iFanin, i )                      \
//     for ( i = 0; iBox - 1 - i >= 0 && Bac_ObjIsBi(p, iBox - 1 - i) && (((iFanin) = Bac_BoxFaninBox(p, iBox, i)), 1); i++ )

// Original macro:
// #define Bac_ObjForEachFanout( p, iCi, iCo )                               \
//     for ( iCo = Bac_ObjFanout(p, iCi); iCo; iCo = Bac_ObjNextFanout(p, iCo) )

// Original macro:
// #define Bac_BoxForEachFanoutBox( p, iBox, iCo, iFanBox )                  \
//     for ( assert(Bac_BoxBoNum(p, iBox) == 1), iCo = Bac_ObjFanout(p, Bac_BoxBo(p, iBox, 0)); iCo && ((iFanBox = Bac_BoxBiBox(p, iCo)), 1); iCo = Bac_ObjNextFanout(p, iCo) )

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
func Bac_ObjAlloc(p *Bac_Ntk_t, Type Bac_ObjType_t, Fanin int32) int32 {
	var iObj int32 = Bac_NtkObjNum(p)
	assert(iObj == int32(vec.Vec_IntSize(&p.vFanin)), "iObj == vec.Vec_IntSize(&p.vFanin)")
	if Type == BAC_OBJ_PI {
		Vec_IntPush(&p.vInputs, iObj)
	} else if Type == BAC_OBJ_PO {
		Vec_IntPush(&p.vOutputs, iObj)
	}
	Vec_StrPush(&p.vType, byte(Abc_Var2Lit(int32(Type), 0)))
	Vec_IntPush(&p.vFanin, Fanin)
	return iObj
}
func Bac_ObjDup(pNew *Bac_Ntk_t, p *Bac_Ntk_t, i int32) int32 {
	fanin := int32(-1)
	if Bac_ObjIsBox(p, i) != 0 {
		fanin = Bac_BoxNtkId(p, i)
	}
	iObj := Bac_ObjAlloc(pNew, Bac_ObjType(p, i), fanin)
	if (Bac_NtkHasNames(p) != 0) && (Bac_NtkHasNames(pNew) != 0) && !(Bac_ObjIsCo(p, i) != 0) {
		Bac_ObjSetName(pNew, iObj, Bac_ObjName(p, i))
	}
	Bac_ObjSetCopy(p, i, iObj)
	return iObj
}
func Bac_BoxAlloc(p *Bac_Ntk_t, Type Bac_ObjType_t, nIns int32, nOuts int32, iNtk int32) int32 {
	var i, iObj int32
	for i = nIns - 1; i >= 0; i-- {
		Bac_ObjAlloc(p, BAC_OBJ_BI, -1)
	}
	iObj = Bac_ObjAlloc(p, Type, iNtk)
	for i = 0; i < nOuts; i++ {
		Bac_ObjAlloc(p, BAC_OBJ_BO, -1)
	}
	return iObj
}
func Bac_BoxDup(pNew *Bac_Ntk_t, p *Bac_Ntk_t, iBox int32) int32 {
	var i, iTerm, iBoxNew int32
	// inlining of macro Bac_BoxForEachBiReverse
	i = Bac_BoxBiNum(p, iBox)
	iTerm = iBox - i
	i--
	for Bac_ObjIsBi(p, iTerm) != 0 {
		Bac_ObjDup(pNew, p, iTerm)
		iTerm++
		i--
	}
	iBoxNew = Bac_ObjDup(pNew, p, iBox)
	if (Bac_NtkHasNames(p) != 0) && (Bac_NtkHasNames(pNew) != 0) && (Bac_ObjName(p, iBox) != 0) {
		Bac_ObjSetName(pNew, iBoxNew, Bac_ObjName(p, iBox))
	}
	if Bac_BoxNtk(p, iBox) != nil {
		Bac_BoxSetNtkId(pNew, iBoxNew, Bac_NtkCopy(Bac_BoxNtk(p, iBox)))
	}
	// inlining of macro Bac_BoxForEachBo
	for iTerm, i = iBox+1, 0; iTerm < Bac_NtkObjNum(p) && (Bac_ObjIsBo(p, iTerm) != 0); iTerm, i = iTerm+1, i+1 {
		Bac_ObjDup(pNew, p, iTerm)
	}
	return iBoxNew
}
func Bac_BoxDelete(p *Bac_Ntk_t, iBox int32) {
	var iStart int32 = iBox - Bac_BoxBiNum(p, iBox)
	var i int32
	var iStop int32 = iBox + Bac_BoxBoNum(p, iBox)
	for i = iStart; i <= iStop; i++ {
		Vec_StrWriteEntry(&p.vType, i, byte(0))
		Vec_IntWriteEntry(&p.vFanin, i, -1)
		if Bac_NtkHasNames(p) != 0 {
			Vec_IntWriteEntry(&p.vName, i, 0)
		}
		if Bac_NtkHasFanouts(p) != 0 {
			Vec_IntWriteEntry(&p.vFanout, i, 0)
		}
	}
}
func Bac_BoxReplace(p *Bac_Ntk_t, iBox int32, pArray []int32, nSize int32) {
	// Bac_NtkUpdateFanout is defined in bacNtk.go.
	var i int32
	var Limit int32 = Bac_BoxBoNum(p, iBox)
	assert(Limit == nSize, "Limit == nSize")
	for i = 0; i < Limit; i++ {
		Bac_NtkUpdateFanout(p, Bac_BoxBo(p, iBox, i), pArray[i])
	}
}

var spStatic_Bac_BoxCollectRanges_Bits vec.Vec_Int_t
var spStatic_Bac_BoxCollectRanges_vBits *vec.Vec_Int_t = &spStatic_Bac_BoxCollectRanges_Bits
var spStatic_Bac_BoxCollectRanges_pArray [10]int32

func Bac_BoxCollectRanges(p *Bac_Ntk_t, iBox int32) *vec.Vec_Int_t {

	var i, iTerm int32
	assert(!(Bac_ObjIsBoxUser(p, iBox) != 0), "!Bac_ObjIsBoxUser(p, iBox)")
	// initialize array
	*spStatic_Bac_BoxCollectRanges_vBits = *vec.Vec_IntAllocArray(spStatic_Bac_BoxCollectRanges_pArray[:], 0)
	// iterate through inputs
	// inlining of macro Bac_BoxForEachBiMain
	for iTerm, i = iBox-1, 0; iTerm >= 0 && (Bac_ObjIsBi(p, iTerm) != 0); iTerm, i = iTerm-1, i+1 {
		if Bac_ObjBit(p, iTerm) != 0 {
		} else {
			Vec_IntPush(spStatic_Bac_BoxCollectRanges_vBits, Bac_BoxBiRange(p, iTerm))
		}
	}
	// iterate through outputs
	// inlining of macro Bac_BoxForEachBoMain
	for iTerm, i = iBox+1, 0; iTerm < Bac_NtkObjNum(p) && (Bac_ObjIsBo(p, iTerm) != 0); iTerm, i = iTerm+1, i+1 {
		if Bac_ObjBit(p, iTerm) != 0 {
		} else {
			Vec_IntPush(spStatic_Bac_BoxCollectRanges_vBits, Bac_BoxBoRange(p, iTerm))
		}
	}
	assert(vec.Vec_IntSize(spStatic_Bac_BoxCollectRanges_vBits) < 10, "vec.Vec_IntSize(spStatic_Bac_BoxCollectRanges_vBits) < 10")
	//Vec_IntPrint( vBits );
	return spStatic_Bac_BoxCollectRanges_vBits
}

/*
*Function*************************************************************

	Synopsis    [Prints vector.]

	Description []

	SideEffects []

	SeeAlso     []

**********************************************************************
*/
func Vec_StrPrint(p *vec.Vec_Str_t, fInt int32) {
	var i int32
	for i = 0; i < int32(vec.Vec_StrSize(p)); i++ {
		if fInt != 0 {
			fmt.Printf("%d ", int32(vec.Vec_StrArray(p)[i]))
		} else {
			fmt.Printf("%c ", vec.Vec_StrArray(p)[i])
		}
	}
	fmt.Printf("\n")
}

/*
*Function*************************************************************

	Synopsis    [Network APIs.]

	Description []

	SideEffects []

	SeeAlso     []

**********************************************************************
*/
func Bac_NtkAlloc(pNew *Bac_Ntk_t, NameId int32, nIns int32, nOuts int32, nObjs int32) {
	var NtkId, fFound int32
	assert(pNew.pDesign != nil, "pNew.pDesign != nil")
	assert(Bac_NtkPiNum(pNew) == 0, "Bac_NtkPiNum(pNew) == 0")
	assert(Bac_NtkPoNum(pNew) == 0, "Bac_NtkPoNum(pNew) == 0")
	pNew.NameId = NameId
	pNew.iCopy = -1
	pNew.iBoxNtk = -1
	pNew.iBoxObj = -1
	Vec_IntGrow(&pNew.vInputs, nIns)
	Vec_IntGrow(&pNew.vOutputs, nOuts)
	Vec_StrGrow(&pNew.vType, nObjs)
	Vec_IntGrow(&pNew.vFanin, nObjs)
	// check if the network is unique
	NtkId = Abc_NamStrFindOrAdd(pNew.pDesign.pMods, Bac_NtkStr(pNew, NameId), &fFound)
	if fFound != 0 {
		fmt.Printf("Network with name %s already exists.\n", Bac_NtkStr(pNew, NameId))
	} else {
		assert(NtkId == Bac_NtkId(pNew), "NtkId == Bac_NtkId(pNew)")
	}
}
func Bac_NtkDup(pNew *Bac_Ntk_t, p *Bac_Ntk_t) {
	var i, iObj int32
	assert(pNew != p, "pNew != p")
	Bac_NtkAlloc(pNew, Bac_NtkNameId(p), Bac_NtkPiNum(p), Bac_NtkPoNum(p), Bac_NtkObjNum(p))
	if vec.Vec_IntSize(&p.vInfo) != 0 {
		Vec_IntAppend(&pNew.vInfo, &p.vInfo)
	}
	Bac_NtkStartCopies(p)
	if Bac_NtkHasNames(p) != 0 {
		Bac_NtkStartNames(pNew)
	}
	// inlining of macro Bac_NtkForEachPi
	for i = 0; i < Bac_NtkPiNum(p); i++ {
		iObj = Bac_NtkPi(p, i)
		Bac_ObjDup(pNew, p, iObj)
	}
	// inlining of macro Bac_NtkForEachBox
	for iObj = 0; iObj < Bac_NtkObjNum(p); iObj++ {
		if !(Bac_ObjIsBox(p, iObj) != 0) {
		} else {
			Bac_BoxDup(pNew, p, iObj)
		}
	}
	// inlining of macro Bac_NtkForEachPo
	for i = 0; i < Bac_NtkPoNum(p); i++ {
		iObj = Bac_NtkPo(p, i)
		Bac_ObjDup(pNew, p, iObj)
	}
	// inlining of macro Bac_NtkForEachCo
	for iObj = 0; iObj < Bac_NtkObjNum(p); iObj++ {
		if !(Bac_ObjIsCo(p, iObj) != 0) {
		} else {
			Bac_ObjSetFanin(pNew, Bac_ObjCopy(p, iObj), Bac_ObjCopy(p, Bac_ObjFanin(p, iObj)))
		}
	}
	//Bac_NtkFreeCopies( p ); // needed for name transfer and host ntk
	assert(Bac_NtkObjNum(pNew) == Bac_NtkObjNumAlloc(pNew), "Bac_NtkObjNum(pNew) == Bac_NtkObjNumAlloc(pNew)")
}
func Bac_NtkDupUserBoxes(pNew *Bac_Ntk_t, p *Bac_Ntk_t) {
	var i, iObj int32
	assert(pNew != p, "pNew != p")
	Bac_NtkAlloc(pNew, Bac_NtkNameId(p), Bac_NtkPiNum(p), Bac_NtkPoNum(p), Bac_NtkObjNum(p)+3*Bac_NtkCoNum(p))
	if vec.Vec_IntSize(&p.vInfo) != 0 {
		Vec_IntAppend(&pNew.vInfo, &p.vInfo)
	}
	Bac_NtkStartCopies(p)
	// inlining of macro Bac_NtkForEachPi
	for i = 0; i < Bac_NtkPiNum(p); i++ {
		iObj = Bac_NtkPi(p, i)
		Bac_ObjDup(pNew, p, iObj)
	}
	// inlining of macro Bac_NtkForEachPo
	for i = 0; i < Bac_NtkPoNum(p); i++ {
		iObj = Bac_NtkPo(p, i)
		Bac_ObjDup(pNew, p, iObj)
	}
	// inlining of macro Bac_NtkForEachBoxUser
	for iObj = 0; iObj < Bac_NtkObjNum(p); iObj++ {
		if !(Bac_ObjIsBoxUser(p, iObj) != 0) {
		} else {
			Bac_BoxDup(pNew, p, iObj)
		}
	}
	// connect feed-throughs
	// inlining of macro Bac_NtkForEachCo
	for iObj = 0; iObj < Bac_NtkObjNum(p); iObj++ {
		if !(Bac_ObjIsCo(p, iObj) != 0) {
		} else if Bac_ObjCopy(p, iObj) >= 0 && Bac_ObjCopy(p, Bac_ObjFanin(p, iObj)) >= 0 {
			Bac_ObjSetFanin(pNew, Bac_ObjCopy(p, iObj), Bac_ObjCopy(p, Bac_ObjFanin(p, iObj)))
		}
	}
}
func Bac_NtkMoveNames(pNew *Bac_Ntk_t, p *Bac_Ntk_t) {
	var i, iBox, iObj int32
	assert((Bac_NtkHasNames(p) != 0), "Bac_NtkHasNames(p)")
	assert(!(Bac_NtkHasNames(pNew) != 0), "!Bac_NtkHasNames(pNew)")
	Bac_NtkStartNames(pNew)
	// inlining of macro Bac_NtkForEachPi
	for i = 0; i < Bac_NtkPiNum(p); i++ {
		iObj = Bac_NtkPi(p, i)
		Bac_ObjSetName(pNew, Bac_ObjCopy(p, iObj), Bac_ObjName(p, iObj))
	}
	// inlining of macro Bac_NtkForEachBoxUser
	for iBox = 0; iBox < Bac_NtkObjNum(p); iBox++ {
		if !(Bac_ObjIsBoxUser(p, iBox) != 0) {
		} else {
			Bac_ObjSetName(pNew, Bac_ObjCopy(p, iBox), Bac_ObjName(p, iBox))
			// inlining of macro Bac_BoxForEachBo
			for iObj, i = iBox+1, 0; iObj < Bac_NtkObjNum(p) && (Bac_ObjIsBo(p, iObj) != 0); iObj, i = iObj+1, i+1 {
				Bac_ObjSetName(pNew, Bac_ObjCopy(p, iObj), Bac_ObjName(p, iObj))
			}
		}
	}
	// inlining of macro Bac_NtkForEachBoxUser
	for iBox = 0; iBox < Bac_NtkObjNum(p); iBox++ {
		if !(Bac_ObjIsBoxUser(p, iBox) != 0) {
		} else {
			// inlining of macro Bac_BoxForEachBi
			for iObj, i = iBox-1, 0; iObj >= 0 && (Bac_ObjIsBi(p, iObj) != 0); iObj, i = iObj-1, i+1 {
				if !(Bac_ObjName(pNew, Bac_ObjFanin(pNew, Bac_ObjCopy(p, iObj))) != 0) {
					Bac_ObjSetName(pNew, Bac_ObjFanin(pNew, Bac_ObjCopy(p, iObj)), Bac_ObjName(p, iObj))
				}
			}
		}
	}
	// inlining of macro Bac_NtkForEachPo
	for i = 0; i < Bac_NtkPoNum(p); i++ {
		iObj = Bac_NtkPo(p, i)
		if !(Bac_ObjName(pNew, Bac_ObjFanin(pNew, Bac_ObjCopy(p, iObj))) != 0) {
			Bac_ObjSetName(pNew, Bac_ObjFanin(pNew, Bac_ObjCopy(p, iObj)), Bac_ObjName(p, iObj))
		}
	}
}

func Bac_NtkFree(p *Bac_Ntk_t) {
	Vec_IntErase(&p.vInputs)
	Vec_IntErase(&p.vOutputs)
	Vec_IntErase(&p.vInfo)
	Vec_StrErase(&p.vType)
	Vec_IntErase(&p.vFanin)
	Vec_IntErase(&p.vIndex)
	Vec_IntErase(&p.vName)
	Vec_IntErase(&p.vFanout)
	Vec_IntErase(&p.vCopy)
	Vec_IntErase(&p.vArray)
	Vec_IntErase(&p.vArray2)
}
func Bac_NtkMemory(p *Bac_Ntk_t) int32 {
	var nMem int32 = int32(unsafe.Sizeof(Bac_Ntk_t{}))
	nMem += int32(Vec_IntMemory(&p.vInputs))
	nMem += int32(Vec_IntMemory(&p.vOutputs))
	nMem += int32(Vec_IntMemory(&p.vInfo))
	nMem += int32(Vec_StrMemory(&p.vType))
	nMem += int32(Vec_IntMemory(&p.vFanin))
	nMem += int32(Vec_IntMemory(&p.vIndex))
	nMem += int32(Vec_IntMemory(&p.vName))
	nMem += int32(Vec_IntMemory(&p.vFanout))
	nMem += int32(Vec_IntMemory(&p.vCopy))
	return nMem
}
func Bac_NtkPrintStats(p *Bac_Ntk_t) {
	fmt.Printf("pi =%5d  ", Bac_NtkPiNum(p))
	fmt.Printf("pi =%5d  ", Bac_NtkPoNum(p))
	fmt.Printf("box =%6d  ", Bac_NtkBoxNum(p))
	fmt.Printf("clp =%7d  ", p.Count)
	fmt.Printf("obj =%7d  ", Bac_NtkObjNum(p))
	fmt.Printf("%s ", Bac_NtkName(p))
	if Bac_NtkHostNtk(p) != nil {
		fmt.Printf("-> %s", Bac_NtkName(Bac_NtkHostNtk(p)))
	}
	fmt.Printf("\n")
}
func Bac_NtkDeriveIndex(p *Bac_Ntk_t) {
	var i, iObj, iTerm int32
	Vec_IntFill(&p.vIndex, Bac_NtkObjNum(p), -1)
	// inlining of macro Bac_NtkForEachPi
	for i = 0; i < Bac_NtkPiNum(p); i++ {
		iObj = Bac_NtkPi(p, i)
		Bac_ObjSetIndex(p, iObj, i)
	}
	// inlining of macro Bac_NtkForEachPo
	for i = 0; i < Bac_NtkPoNum(p); i++ {
		iObj = Bac_NtkPo(p, i)
		Bac_ObjSetIndex(p, iObj, i)
	}
	// inlining of macro Bac_NtkForEachBox
	for iObj = 0; iObj < Bac_NtkObjNum(p); iObj++ {
		if !(Bac_ObjIsBox(p, iObj) != 0) {
		} else {
			// inlining of macro Bac_BoxForEachBi
			for iTerm, i = iObj-1, 0; iTerm >= 0 && (Bac_ObjIsBi(p, iTerm) != 0); iTerm, i = iTerm-1, i+1 {
				Bac_ObjSetIndex(p, iTerm, i)
			}
			// inlining of macro Bac_BoxForEachBo
			for iTerm, i = iObj+1, 0; iTerm < Bac_NtkObjNum(p) && (Bac_ObjIsBo(p, iTerm) != 0); iTerm, i = iTerm+1, i+1 {
				Bac_ObjSetIndex(p, iTerm, i)
			}
		}
	}
}
func Bac_NtkPrint(p *Bac_Ntk_t) {
	var i, Type, Beg, End int32
	fmt.Printf("Interface (%d):\n", Bac_NtkInfoNum(p))
	// inlining of macro Vec_IntForEachEntryTriple
	for i = 0; int(i+2) < vec.Vec_IntSize(&p.vInfo); i += 3 {
		Beg = Vec_IntEntry(&p.vInfo, i+1)
		End = Vec_IntEntry(&p.vInfo, i+2)
		fmt.Printf("%6d : ", i)
		fmt.Printf("Type =%3d  ", Bac_NtkInfoType(p, i/3))
		if Beg >= 0 {
			fmt.Printf("[%d:%d]   ", End, Beg)
		} else {
			fmt.Printf("        ")
		}
		fmt.Printf("Name =%3d   ", Bac_NtkInfoName(p, i/3))
		if Bac_NtkInfoName(p, i/3) != 0 {
			fmt.Printf("%s", Bac_NtkStr(p, Bac_NtkInfoName(p, i/3)))
		}
		fmt.Printf("\n")
	}
	fmt.Printf("Objects (%d):\n", Bac_NtkObjNum(p))
	// inlining of macro Bac_NtkForEachObjType
	for i = 0; i < Bac_NtkObjNum(p); i++ {
		Type = int32(Bac_ObjType(p, i))
		if Type != 0 {
			fmt.Printf("%6d : ", i)
			fmt.Printf("Type =%3d  ", Type)
			if Bac_ObjIsCo(p, i) != 0 {
				fmt.Printf("Fanin =%6d  ", Bac_ObjFanin(p, i))
			} else if (Bac_NtkHasNames(p) != 0) && (Bac_ObjName(p, i) != 0) {
				fmt.Printf("Name  =%6d(%d)  ", Bac_ObjNameId(p, i), Bac_ObjNameType(p, i))
				if Bac_ObjNameType(p, i) <= BAC_NAME_WORD {
					fmt.Printf("%s", Bac_ObjNameStr(p, i))
				}
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
func Bac_ManAlloc(pFileName string, nNtks int32) *Bac_Man_t {
	// Bac_ManSetupTypes is defined in bacNtk.go.
	var pNtk *Bac_Ntk_t
	var i int32
	// inlining of macro ABC_CALLOC
	var pNew *Bac_Man_t = &Bac_Man_t{}
	pNew.pName = Extra_FileDesignName(pFileName)
	pNew.pSpec = Abc_UtilStrsav(pFileName)
	pNew.pStrs = Abc_NamStart(1000, 24)
	pNew.pMods = Abc_NamStart(1000, 24)
	pNew.iRoot = 1
	pNew.nNtks = nNtks
	// inlining of macro ABC_CALLOC
	pNew.pNtks = make([]Bac_Ntk_t, pNew.nNtks+1)
	// inlining of macro Bac_ManForEachNtk
	for i = 1; i <= Bac_ManNtkNum(pNew); i++ {
		pNtk = Bac_ManNtk(pNew, i)
		pNtk.pDesign = pNew
		pNtk.ntkID = i
	}
	Bac_ManSetupTypes(&pNew.pPrimNames, &pNew.pPrimSymbs)
	return pNew
}
func Bac_ManStart(p *Bac_Man_t, nNtks int32) *Bac_Man_t {
	var pNtk *Bac_Ntk_t
	var i int32
	// inlining of macro ABC_CALLOC
	var pNew *Bac_Man_t = &Bac_Man_t{}
	pNew.pName = Abc_UtilStrsav(Bac_ManName(p))
	pNew.pSpec = Abc_UtilStrsav(Bac_ManSpec(p))
	pNew.pStrs = Abc_NamRef(p.pStrs)
	pNew.pMods = Abc_NamStart(1000, 24)
	pNew.iRoot = 1
	pNew.nNtks = nNtks
	// inlining of macro ABC_CALLOC
	pNew.pNtks = make([]Bac_Ntk_t, nNtks+1)
	// inlining of macro Bac_ManForEachNtk
	for i = 1; i <= Bac_ManNtkNum(pNew); i++ {
		pNtk = Bac_ManNtk(pNew, i)
		pNtk.pDesign = pNew
		pNtk.ntkID = i
	}
	return pNew
}
func Bac_ManDup(p *Bac_Man_t) *Bac_Man_t {
	var pNtk, pHost *Bac_Ntk_t
	var i int32
	var pNew *Bac_Man_t = Bac_ManStart(p, Bac_ManNtkNum(p))
	// inlining of macro Bac_ManForEachNtk
	for i = 1; i <= Bac_ManNtkNum(p); i++ {
		pNtk = Bac_ManNtk(p, i)
		Bac_NtkSetCopy(pNtk, i)
	}
	// inlining of macro Bac_ManForEachNtk
	for i = 1; i <= Bac_ManNtkNum(p); i++ {
		pNtk = Bac_ManNtk(p, i)
		Bac_NtkDup(Bac_NtkCopyNtk(pNew, pNtk), pNtk)
	}
	// inlining of macro Bac_ManForEachNtk
	for i = 1; i <= Bac_ManNtkNum(p); i++ {
		pNtk = Bac_ManNtk(p, i)
		pHost = Bac_NtkHostNtk(pNtk)
		if pHost != nil {
			Bac_NtkSetHost(Bac_NtkCopyNtk(pNew, pNtk), Bac_NtkCopy(pHost), Bac_ObjCopy(pHost, Bac_NtkHostObj(pNtk)))
		}
	}
	return pNew
}
func Bac_ManDupUserBoxes(p *Bac_Man_t) *Bac_Man_t {
	var pNtk, pHost *Bac_Ntk_t
	var i int32
	var pNew *Bac_Man_t = Bac_ManStart(p, Bac_ManNtkNum(p))
	// inlining of macro Bac_ManForEachNtk
	for i = 1; i <= Bac_ManNtkNum(p); i++ {
		pNtk = Bac_ManNtk(p, i)
		Bac_NtkSetCopy(pNtk, i)
	}
	// inlining of macro Bac_ManForEachNtk
	for i = 1; i <= Bac_ManNtkNum(p); i++ {
		pNtk = Bac_ManNtk(p, i)
		Bac_NtkDupUserBoxes(Bac_NtkCopyNtk(pNew, pNtk), pNtk)
	}
	// inlining of macro Bac_ManForEachNtk
	for i = 1; i <= Bac_ManNtkNum(p); i++ {
		pNtk = Bac_ManNtk(p, i)
		pHost = Bac_NtkHostNtk(pNtk)
		if pHost != nil {
			Bac_NtkSetHost(Bac_NtkCopyNtk(pNew, pNtk), Bac_NtkCopy(pHost), Bac_ObjCopy(pHost, Bac_NtkHostObj(pNtk)))
		}
	}
	return pNew
}
func Bac_ManMoveNames(pNew *Bac_Man_t, p *Bac_Man_t) {
	var pNtk *Bac_Ntk_t
	var i int32
	// inlining of macro Bac_ManForEachNtk
	for i = 1; i <= Bac_ManNtkNum(p); i++ {
		pNtk = Bac_ManNtk(p, i)
		Bac_NtkMoveNames(Bac_NtkCopyNtk(pNew, pNtk), pNtk)
	}
}

func Bac_ManFree(p *Bac_Man_t) {
	var pNtk *Bac_Ntk_t
	var i int32
	// inlining of macro Bac_ManForEachNtk
	for i = 1; i <= Bac_ManNtkNum(p); i++ {
		pNtk = Bac_ManNtk(p, i)
		Bac_NtkFree(pNtk)
	}
	Vec_IntErase(&p.vBuf2LeafNtk)
	Vec_IntErase(&p.vBuf2LeafObj)
	Vec_IntErase(&p.vBuf2RootNtk)
	Vec_IntErase(&p.vBuf2RootObj)
	Abc_NamDeref(p.pStrs)
	Abc_NamDeref(p.pMods)
	p.pName = ""
	p.pSpec = ""
	p.pNtks = nil
	p = nil
}
func Bac_ManMemory(p *Bac_Man_t) int32 {
	var pNtk *Bac_Ntk_t
	var i int32
	var nMem int32 = int32(unsafe.Sizeof(Bac_Man_t{}))
	if p.pName != "" {
		nMem += int32(len(p.pName))
	}
	if p.pSpec != "" {
		nMem += int32(len(p.pSpec))
	}
	nMem += Abc_NamMemUsed(p.pStrs)
	nMem += Abc_NamMemUsed(p.pMods)
	// inlining of macro Bac_ManForEachNtk
	for i = 1; i <= Bac_ManNtkNum(p); i++ {
		pNtk = Bac_ManNtk(p, i)
		nMem += Bac_NtkMemory(pNtk)
	}
	return nMem
}
func Bac_ManObjNum(p *Bac_Man_t) int32 {
	var pNtk *Bac_Ntk_t
	var i, Count int32
	// inlining of macro Bac_ManForEachNtk
	for i = 1; i <= Bac_ManNtkNum(p); i++ {
		pNtk = Bac_ManNtk(p, i)
		Count += Bac_NtkObjNum(pNtk)
	}
	return Count
}
func Bac_ManNodeNum(p *Bac_Man_t) int32 {
	var pNtk *Bac_Ntk_t
	var i, Count int32
	// inlining of macro Bac_ManForEachNtk
	for i = 1; i <= Bac_ManNtkNum(p); i++ {
		pNtk = Bac_ManNtk(p, i)
		Count += Bac_NtkBoxNum(pNtk)
	}
	return Count
}
func Bac_ManBoxNum_rec(p *Bac_Ntk_t) int32 {
	var iObj, Counter int32
	if p.Count >= 0 {
		return p.Count
	}
	// inlining of macro Bac_NtkForEachBox
	for iObj = 0; iObj < Bac_NtkObjNum(p); iObj++ {
		if !(Bac_ObjIsBox(p, iObj) != 0) {
		} else {
			if Bac_ObjIsBoxUser(p, iObj) != 0 {
				Counter += Bac_ManBoxNum_rec(Bac_BoxNtk(p, iObj))
			} else {
				Counter++
			}
		}
	}
	p.Count = Counter
	return Counter
}
func Bac_ManBoxNum(p *Bac_Man_t) int32 {
	var pNtk *Bac_Ntk_t
	var i int32
	// inlining of macro Bac_ManForEachNtk
	for i = 1; i <= Bac_ManNtkNum(p); i++ {
		pNtk = Bac_ManNtk(p, i)
		pNtk.Count = -1
	}
	return Bac_ManBoxNum_rec(Bac_ManRoot(p))
}
func Bac_ManPrintStats(p *Bac_Man_t, nModules int32, fVerbose int32) {
	var pNtk *Bac_Ntk_t
	var i int32
	var pRoot *Bac_Ntk_t = Bac_ManRoot(p)
	fmt.Printf("%-12s : ", Bac_ManName(p))
	fmt.Printf("pi =%5d  ", Bac_NtkPiNum(pRoot))
	fmt.Printf("po =%5d  ", Bac_NtkPoNum(pRoot))
	fmt.Printf("pri =%4d  ", Bac_ManPrimNum(p))
	fmt.Printf("mod =%6d  ", Bac_ManNtkNum(p))
	fmt.Printf("box =%7d  ", Bac_ManNodeNum(p))
	fmt.Printf("obj =%7d  ", Bac_ManObjNum(p))
	fmt.Printf("mem =%6.3f MB", float64(Bac_ManMemory(p))/(1<<20))
	fmt.Printf("\n")
	Bac_ManBoxNum(p)
	// inlining of macro Bac_ManForEachNtk
	for i = 1; i <= Bac_ManNtkNum(p); i++ {
		pNtk = Bac_ManNtk(p, i)
		if i == nModules+1 {
			break
		}
		fmt.Printf("Module %5d : ", i)
		Bac_NtkPrintStats(pNtk)
	}
}

/*
*Function*************************************************************

	Synopsis    [Other APIs.]

	Description []

	SideEffects []

	SeeAlso     []

**********************************************************************
*/
func Ptr_SopToType(pSop string) Bac_ObjType_t {
	if !(int32(strings.Compare(pSop, " 0\n")) != 0) {
		return BAC_BOX_CF
	}
	if !(int32(strings.Compare(pSop, " 1\n")) != 0) {
		return BAC_BOX_CT
	}
	if !(int32(strings.Compare(pSop, "1 1\n")) != 0) {
		return BAC_BOX_BUF
	}
	if !(int32(strings.Compare(pSop, "0 1\n")) != 0) {
		return BAC_BOX_INV
	}
	if !(int32(strings.Compare(pSop, "11 1\n")) != 0) {
		return BAC_BOX_AND
	}
	if !(int32(strings.Compare(pSop, "00 1\n")) != 0) {
		return BAC_BOX_NOR
	}
	if !(int32(strings.Compare(pSop, "00 0\n")) != 0) {
		return BAC_BOX_OR
	}
	if !(int32(strings.Compare(pSop, "-1 1\n1- 1\n")) != 0) {
		return BAC_BOX_OR
	}
	if !(int32(strings.Compare(pSop, "1- 1\n-1 1\n")) != 0) {
		return BAC_BOX_OR
	}
	if !(int32(strings.Compare(pSop, "01 1\n10 1\n")) != 0) {
		return BAC_BOX_XOR
	}
	if !(int32(strings.Compare(pSop, "10 1\n01 1\n")) != 0) {
		return BAC_BOX_XOR
	}
	if !(int32(strings.Compare(pSop, "11 1\n00 1\n")) != 0) {
		return BAC_BOX_XNOR
	}
	if !(int32(strings.Compare(pSop, "00 1\n11 1\n")) != 0) {
		return BAC_BOX_XNOR
	}
	if !(int32(strings.Compare(pSop, "10 1\n")) != 0) {
		return BAC_BOX_SHARP
	}
	if !(int32(strings.Compare(pSop, "01 1\n")) != 0) {
		return BAC_BOX_SHARPL
	}
	assert(false, "0")
	return BAC_OBJ_NONE
}
func Ptr_SopToTypeName(pSop string) string {
	if !(int32(strings.Compare(pSop, " 0\n")) != 0) {
		return "BAC_BOX_C0"
	}
	if !(int32(strings.Compare(pSop, " 1\n")) != 0) {
		return "BAC_BOX_C1"
	}
	if !(int32(strings.Compare(pSop, "1 1\n")) != 0) {
		return "BAC_BOX_BUF"
	}
	if !(int32(strings.Compare(pSop, "0 1\n")) != 0) {
		return "BAC_BOX_INV"
	}
	if !(int32(strings.Compare(pSop, "11 1\n")) != 0) {
		return "BAC_BOX_AND"
	}
	if !(int32(strings.Compare(pSop, "00 1\n")) != 0) {
		return "BAC_BOX_NOR"
	}
	if !(int32(strings.Compare(pSop, "00 0\n")) != 0) {
		return "BAC_BOX_OR"
	}
	if !(int32(strings.Compare(pSop, "-1 1\n1- 1\n")) != 0) {
		return "BAC_BOX_OR"
	}
	if !(int32(strings.Compare(pSop, "1- 1\n-1 1\n")) != 0) {
		return "BAC_BOX_OR"
	}
	if !(int32(strings.Compare(pSop, "01 1\n10 1\n")) != 0) {
		return "BAC_BOX_XOR"
	}
	if !(int32(strings.Compare(pSop, "10 1\n01 1\n")) != 0) {
		return "BAC_BOX_XOR"
	}
	if !(int32(strings.Compare(pSop, "11 1\n00 1\n")) != 0) {
		return "BAC_BOX_XNOR"
	}
	if !(int32(strings.Compare(pSop, "00 1\n11 1\n")) != 0) {
		return "BAC_BOX_XNOR"
	}
	if !(int32(strings.Compare(pSop, "10 1\n")) != 0) {
		return "BAC_BOX_SHARP"
	}
	if !(int32(strings.Compare(pSop, "01 1\n")) != 0) {
		return "BAC_BOX_SHARPL"
	}
	assert(false, "0")
	return ""
}
func Ptr_TypeToName(Type Bac_ObjType_t) string {
	if Type == BAC_BOX_CF {
		return "const0"
	}
	if Type == BAC_BOX_CT {
		return "const1"
	}
	if Type == BAC_BOX_CX {
		return "constX"
	}
	if Type == BAC_BOX_CZ {
		return "constZ"
	}
	if Type == BAC_BOX_BUF {
		return "buf"
	}
	if Type == BAC_BOX_INV {
		return "not"
	}
	if Type == BAC_BOX_AND {
		return "and"
	}
	if Type == BAC_BOX_NAND {
		return "nand"
	}
	if Type == BAC_BOX_OR {
		return "or"
	}
	if Type == BAC_BOX_NOR {
		return "nor"
	}
	if Type == BAC_BOX_XOR {
		return "xor"
	}
	if Type == BAC_BOX_XNOR {
		return "xnor"
	}
	if Type == BAC_BOX_MUX {
		return "mux"
	}
	if Type == BAC_BOX_MAJ {
		return "maj"
	}
	if Type == BAC_BOX_SHARP {
		return "sharp"
	}
	if Type == BAC_BOX_SHARPL {
		return "sharpl"
	}
	assert(false, "0")
	return "???"
}
func Ptr_TypeToSop(Type Bac_ObjType_t) string {
	if Type == BAC_BOX_CF {
		return " 0\n"
	}
	if Type == BAC_BOX_CT {
		return " 1\n"
	}
	if Type == BAC_BOX_CX {
		return " 0\n"
	}
	if Type == BAC_BOX_CZ {
		return " 0\n"
	}
	if Type == BAC_BOX_BUF {
		return "1 1\n"
	}
	if Type == BAC_BOX_INV {
		return "0 1\n"
	}
	if Type == BAC_BOX_AND {
		return "11 1\n"
	}
	if Type == BAC_BOX_NAND {
		return "11 0\n"
	}
	if Type == BAC_BOX_OR {
		return "00 0\n"
	}
	if Type == BAC_BOX_NOR {
		return "00 1\n"
	}
	if Type == BAC_BOX_XOR {
		return "01 1\n10 1\n"
	}
	if Type == BAC_BOX_XNOR {
		return "00 1\n11 1\n"
	}
	if Type == BAC_BOX_SHARP {
		return "10 1\n"
	}
	if Type == BAC_BOX_SHARPL {
		return "01 1\n"
	}
	if Type == BAC_BOX_MUX {
		return "11- 1\n0-1 1\n"
	}
	if Type == BAC_BOX_MAJ {
		return "11- 1\n1-1 1\n-11 1\n"
	}
	assert(false, "0")
	return "???"
}

/*=== bacCom.c ===============================================================*/
// extern void          Abc_FrameImportPtr( Vec_Ptr_t * vPtr );

// extern Vec_Ptr_t *   Abc_FrameExportPtr();

/*=== bacBlast.c =============================================================*/
// extern int           Bac_NtkBuildLibrary( Bac_Man_t * p );

// extern Gia_Man_t *   Bac_ManExtract( Bac_Man_t * p, int fBuffers, int fVerbose );

// extern Bac_Man_t *   Bac_ManInsertGia( Bac_Man_t * p, Gia_Man_t * pGia );

// extern void *        Bac_ManInsertAbc( Bac_Man_t * p, void * pAbc );

/*=== bacCba.c ===============================================================*/
// extern Bac_Man_t *   Bac_ManReadBac( char * pFileName );

// extern void          Bac_ManWriteBac( char * pFileName, Bac_Man_t * p );

/*=== bacNtk.c ===============================================================*/
// extern char *        Bac_NtkGenerateName( Bac_Ntk_t * p, Bac_ObjType_t Type, vec.Vec_Int_t * vBits );

// extern Bac_ObjType_t Bac_NameToType( char * pName );

// extern vec.Vec_Int_t *   Bac_NameToRanges( char * pName );

// extern void          Bac_NtkUpdateFanout( Bac_Ntk_t * p, int iOld, int iNew );

// extern void          Bac_ManDeriveFanout( Bac_Man_t * p );

//extern void          Bac_ManAssignInternNames( Bac_Man_t * p );
// extern void          Bac_ManAssignInternWordNames( Bac_Man_t * p );

// extern Bac_Man_t *   Bac_ManCollapse( Bac_Man_t * p );

// extern void          Bac_ManSetupTypes( char ** pNames, char ** pSymbs );

/*=== bacPtr.c ===============================================================*/
// extern void          Bac_PtrFree( Vec_Ptr_t * vDes );

// extern int           Bac_PtrMemory( Vec_Ptr_t * vDes );

// extern void          Bac_PtrDumpBlif( char * pFileName, Vec_Ptr_t * vDes );

// extern void          Bac_PtrDumpVerilog( char * pFileName, Vec_Ptr_t * vDes );

// extern Vec_Ptr_t *   Bac_PtrTransformTest( Vec_Ptr_t * vDes );

/*=== bacPtrAbc.c ============================================================*/
// extern Bac_Man_t *   Bac_PtrTransformToCba( Vec_Ptr_t * vDes );

// extern Vec_Ptr_t *   Bac_PtrDeriveFromCba( Bac_Man_t * p );

/*=== bacPrsBuild.c ==========================================================*/
// extern Bac_Man_t *   Psr_ManBuildCba( char * pFileName, Vec_Ptr_t * vDes );

/*=== bacReadBlif.c ==========================================================*/
// extern Vec_Ptr_t *   Psr_ManReadBlif( char * pFileName );

/*=== bacReadSmt.c ===========================================================*/
// extern Vec_Ptr_t *   Psr_ManReadSmt( char * pFileName );

/*=== bacReadVer.c ===========================================================*/
// extern Vec_Ptr_t *   Psr_ManReadVerilog( char * pFileName );

/*=== bacWriteBlif.c =========================================================*/
// extern void          Psr_ManWriteBlif( char * pFileName, Vec_Ptr_t * p );

// extern void          Bac_ManWriteBlif( char * pFileName, Bac_Man_t * p );

/*=== bacWriteVer.c ==========================================================*/
// extern void          Psr_ManWriteVerilog( char * pFileName, Vec_Ptr_t * p );

// extern void          Bac_ManWriteVerilog( char * pFileName, Bac_Man_t * p, int fUseAssign );

//ABC_NAMESPACE_HEADER_END

////////////////////////////////////////////////////////////////////////
///                       END OF FILE                                ///
////////////////////////////////////////////////////////////////////////
