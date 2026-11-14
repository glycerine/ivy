package wlc

import "unsafe"

/**CFile****************************************************************

  FileName    [wlc.h]

  SystemName  [ABC: Logic synthesis and verification system.]

  PackageName [Verilog parser.]

  Synopsis    [External declarations.]

  Author      [Alan Mishchenko]

  Affiliation [UC Berkeley]

  Date        [Ver. 1.0. Started - August 22, 2014.]

  Revision    [$Id: wlc.h,v 1.00 2014/09/12 00:00:00 alanmi Exp $]

***********************************************************************/

////////////////////////////////////////////////////////////////////////
///                          INCLUDES                                ///
////////////////////////////////////////////////////////////////////////

//#include "aig/gia/gia.h"
//#include "misc/extra/extra.h"
//#include "misc/util/utilNam.h"
//#include "misc/mem/mem.h"
//#include "misc/extra/extra.h"
//#include "misc/util/utilTruth.h"
//#include "base/main/mainInt.h"

////////////////////////////////////////////////////////////////////////
///                         PARAMETERS                               ///
////////////////////////////////////////////////////////////////////////

//ABC_NAMESPACE_HEADER_START

// object types
type Wlc_ObjType_t int32

const (
	WLC_OBJ_NONE Wlc_ObjType_t = 0
	// 00: unknown
	WLC_OBJ_PI Wlc_ObjType_t = WLC_OBJ_NONE + 1
	// 01: primary input
	WLC_OBJ_PO Wlc_ObjType_t = WLC_OBJ_PI + 1
	// 02: primary output
	WLC_OBJ_FO Wlc_ObjType_t = WLC_OBJ_PO + 1
	// 03: flop output
	WLC_OBJ_FI Wlc_ObjType_t = WLC_OBJ_FO + 1
	// 04: flop input (unused)
	WLC_OBJ_FF Wlc_ObjType_t = WLC_OBJ_FI + 1
	// 05: flop
	WLC_OBJ_CONST Wlc_ObjType_t = WLC_OBJ_FF + 1
	// 06: constant
	WLC_OBJ_BUF Wlc_ObjType_t = WLC_OBJ_CONST + 1
	// 07: buffer
	WLC_OBJ_MUX Wlc_ObjType_t = WLC_OBJ_BUF + 1
	// 08: multiplexer
	WLC_OBJ_SHIFT_R Wlc_ObjType_t = WLC_OBJ_MUX + 1
	// 09: shift right
	WLC_OBJ_SHIFT_RA Wlc_ObjType_t = WLC_OBJ_SHIFT_R + 1
	// 10: shift right (arithmetic)
	WLC_OBJ_SHIFT_L Wlc_ObjType_t = WLC_OBJ_SHIFT_RA + 1
	// 11: shift left
	WLC_OBJ_SHIFT_LA Wlc_ObjType_t = WLC_OBJ_SHIFT_L + 1
	// 12: shift left (arithmetic)
	WLC_OBJ_ROTATE_R Wlc_ObjType_t = WLC_OBJ_SHIFT_LA + 1
	// 13: rotate right
	WLC_OBJ_ROTATE_L Wlc_ObjType_t = WLC_OBJ_ROTATE_R + 1
	// 14: rotate left
	WLC_OBJ_BIT_NOT Wlc_ObjType_t = WLC_OBJ_ROTATE_L + 1
	// 15: bitwise NOT
	WLC_OBJ_BIT_AND Wlc_ObjType_t = WLC_OBJ_BIT_NOT + 1
	// 16: bitwise AND
	WLC_OBJ_BIT_OR Wlc_ObjType_t = WLC_OBJ_BIT_AND + 1
	// 17: bitwise OR
	WLC_OBJ_BIT_XOR Wlc_ObjType_t = WLC_OBJ_BIT_OR + 1
	// 18: bitwise XOR
	WLC_OBJ_BIT_NAND Wlc_ObjType_t = WLC_OBJ_BIT_XOR + 1
	// 19: bitwise NAND
	WLC_OBJ_BIT_NOR Wlc_ObjType_t = WLC_OBJ_BIT_NAND + 1
	// 20: bitwise OR
	WLC_OBJ_BIT_NXOR Wlc_ObjType_t = WLC_OBJ_BIT_NOR + 1
	// 21: bitwise NXOR
	WLC_OBJ_BIT_SELECT Wlc_ObjType_t = WLC_OBJ_BIT_NXOR + 1
	// 22: bit selection
	WLC_OBJ_BIT_CONCAT Wlc_ObjType_t = WLC_OBJ_BIT_SELECT + 1
	// 23: bit concatenation
	WLC_OBJ_BIT_ZEROPAD Wlc_ObjType_t = WLC_OBJ_BIT_CONCAT + 1
	// 24: zero padding
	WLC_OBJ_BIT_SIGNEXT Wlc_ObjType_t = WLC_OBJ_BIT_ZEROPAD + 1
	// 25: sign extension
	WLC_OBJ_LOGIC_NOT Wlc_ObjType_t = WLC_OBJ_BIT_SIGNEXT + 1
	// 26: logic NOT
	WLC_OBJ_LOGIC_IMPL Wlc_ObjType_t = WLC_OBJ_LOGIC_NOT + 1
	// 27: logic implication
	WLC_OBJ_LOGIC_AND Wlc_ObjType_t = WLC_OBJ_LOGIC_IMPL + 1
	// 28: logic AND
	WLC_OBJ_LOGIC_OR Wlc_ObjType_t = WLC_OBJ_LOGIC_AND + 1
	// 29: logic OR
	WLC_OBJ_LOGIC_XOR Wlc_ObjType_t = WLC_OBJ_LOGIC_OR + 1
	// 30: logic XOR
	WLC_OBJ_COMP_EQU Wlc_ObjType_t = WLC_OBJ_LOGIC_XOR + 1
	// 31: compare equal
	WLC_OBJ_COMP_NOTEQU Wlc_ObjType_t = WLC_OBJ_COMP_EQU + 1
	// 32: compare not equal
	WLC_OBJ_COMP_LESS Wlc_ObjType_t = WLC_OBJ_COMP_NOTEQU + 1
	// 33: compare less
	WLC_OBJ_COMP_MORE Wlc_ObjType_t = WLC_OBJ_COMP_LESS + 1
	// 34: compare more
	WLC_OBJ_COMP_LESSEQU Wlc_ObjType_t = WLC_OBJ_COMP_MORE + 1
	// 35: compare less or equal
	WLC_OBJ_COMP_MOREEQU Wlc_ObjType_t = WLC_OBJ_COMP_LESSEQU + 1
	// 36: compare more or equal
	WLC_OBJ_REDUCT_AND Wlc_ObjType_t = WLC_OBJ_COMP_MOREEQU + 1
	// 37: reduction AND
	WLC_OBJ_REDUCT_OR Wlc_ObjType_t = WLC_OBJ_REDUCT_AND + 1
	// 38: reduction OR
	WLC_OBJ_REDUCT_XOR Wlc_ObjType_t = WLC_OBJ_REDUCT_OR + 1
	// 39: reduction XOR
	WLC_OBJ_REDUCT_NAND Wlc_ObjType_t = WLC_OBJ_REDUCT_XOR + 1
	// 40: reduction NAND
	WLC_OBJ_REDUCT_NOR Wlc_ObjType_t = WLC_OBJ_REDUCT_NAND + 1
	// 41: reduction NOR
	WLC_OBJ_REDUCT_NXOR Wlc_ObjType_t = WLC_OBJ_REDUCT_NOR + 1
	// 42: reduction NXOR
	WLC_OBJ_ARI_ADD Wlc_ObjType_t = WLC_OBJ_REDUCT_NXOR + 1
	// 43: arithmetic addition
	WLC_OBJ_ARI_SUB Wlc_ObjType_t = WLC_OBJ_ARI_ADD + 1
	// 44: arithmetic subtraction
	WLC_OBJ_ARI_MULTI Wlc_ObjType_t = WLC_OBJ_ARI_SUB + 1
	// 45: arithmetic multiplier
	WLC_OBJ_ARI_DIVIDE Wlc_ObjType_t = WLC_OBJ_ARI_MULTI + 1
	// 46: arithmetic division
	WLC_OBJ_ARI_REM Wlc_ObjType_t = WLC_OBJ_ARI_DIVIDE + 1
	// 47: arithmetic remainder
	WLC_OBJ_ARI_MODULUS Wlc_ObjType_t = WLC_OBJ_ARI_REM + 1
	// 48: arithmetic modulus
	WLC_OBJ_ARI_POWER Wlc_ObjType_t = WLC_OBJ_ARI_MODULUS + 1
	// 49: arithmetic power
	WLC_OBJ_ARI_MINUS Wlc_ObjType_t = WLC_OBJ_ARI_POWER + 1
	// 50: arithmetic minus
	WLC_OBJ_ARI_SQRT Wlc_ObjType_t = WLC_OBJ_ARI_MINUS + 1
	// 51: integer square root
	WLC_OBJ_ARI_SQUARE Wlc_ObjType_t = WLC_OBJ_ARI_SQRT + 1
	// 52: integer square
	WLC_OBJ_TABLE Wlc_ObjType_t = WLC_OBJ_ARI_SQUARE + 1
	// 53: bit table
	WLC_OBJ_READ Wlc_ObjType_t = WLC_OBJ_TABLE + 1
	// 54: read port
	WLC_OBJ_WRITE Wlc_ObjType_t = WLC_OBJ_READ + 1
	// 55: write port
	WLC_OBJ_ARI_ADDSUB Wlc_ObjType_t = WLC_OBJ_WRITE + 1
	// 56: adder-subtractor
	WLC_OBJ_SEL Wlc_ObjType_t = WLC_OBJ_ARI_ADDSUB + 1
	// 57: positionally encoded selector
	WLC_OBJ_DEC Wlc_ObjType_t = WLC_OBJ_SEL + 1
	// 58: decoder
	WLC_OBJ_LUT Wlc_ObjType_t = WLC_OBJ_DEC + 1
	// 59: lookup table
	WLC_OBJ_NUMBER Wlc_ObjType_t = WLC_OBJ_LUT + 1 // 59: unused

)

// when adding new types, remember to update table Wlc_Names in "wlcNtk.c"

// Unlike AIG managers and logic networks in ABC, this network treats POs and FIs
// as attributes of internal nodes and *not* as separate types of objects.

////////////////////////////////////////////////////////////////////////
///                         BASIC TYPES                              ///
////////////////////////////////////////////////////////////////////////

type Wlc_Obj_t struct { // 24 bytes

	Type    uint32 /* C bitfield: 6 bits; review packing and truncation */ // node type
	Signed  uint32 /* C bitfield: 1 bits; review packing and truncation */ // signed
	Mark    uint32 /* C bitfield: 1 bits; review packing and truncation */ // user mark
	Mark2   uint32 /* C bitfield: 1 bits; review packing and truncation */ // user mark
	fIsPo   uint32 /* C bitfield: 1 bits; review packing and truncation */ // this is PO
	fIsFi   uint32 /* C bitfield: 1 bits; review packing and truncation */ // this is FI
	fXConst uint32 /* C bitfield: 1 bits; review packing and truncation */ // this is X-valued constant
	nFanins uint32 // fanin count
	End     int32  // range end
	Beg     int32  // range begin
	Fanins  [2]int32
	pFanins [1][]int32
}

type Wlc_Ntk_t struct {
	pName     string                // model name
	pSpec     string                // input file name
	vPis      Vec_Int_t             // primary inputs
	vPos      Vec_Int_t             // primary outputs
	vCis      Vec_Int_t             // combinational inputs
	vCos      Vec_Int_t             // combinational outputs
	vFfs      Vec_Int_t             // flops
	vFfs2     Vec_Int_t             // flops
	vArsts    *Vec_Int_t            // async resets
	vInits    *Vec_Int_t            // initial values
	pInits    string                // initial values
	nObjs     [WLC_OBJ_NUMBER]int32 // counter of objects of each type
	nAnds     [WLC_OBJ_NUMBER]int32 // counter of AND gates after blasting
	fSmtLib   int32                 // the network comes from an SMT-LIB file
	fAsyncRst int32                 // the network has asynchronous reset
	fMemPorts int32                 // the network contains memory ports
	fEasyFfs  int32                 // the network contains simple flops
	nAssert   int32                 // the number of asserts
	// memory for objects
	pObjs      *Wlc_Obj_t
	objStore   []Wlc_Obj_t
	iObj       int32
	nObjsAlloc int32
	pMemFanin  *Mem_Flex_t
	pMemTable  *Mem_Flex_t
	vTables    *Vec_Ptr_t
	vLutTruths *Vec_Wrd_t
	// object names
	pManName *Abc_Nam_t // object names
	vNameIds Vec_Int_t  // object name IDs
	vValues  Vec_Int_t  // value objects
	// object attributes
	nTravIds int32     // counter of traversal IDs
	vTravIds Vec_Int_t // trav IDs of the objects
	vCopies  Vec_Int_t // object first bits
	vBits    Vec_Int_t // object mapping into AIG nodes
	vLevels  Vec_Int_t // object levels
	vRefs    Vec_Int_t // object reference counters
	vPoPairs Vec_Int_t // pairs of primary outputs
}

type Wlc_Par_t struct {
	nBitsAdd        int32             // adder bit-width
	nBitsMul        int32             // multiplier bit-widht
	nBitsMux        int32             // MUX bit-width
	nBitsFlop       int32             // flop bit-width
	nIterMax        int32             // the max number of iterations
	nLimit          int32             // the max number of signals
	fXorOutput      int32             // XOR outputs of word-level miter
	fCheckClauses   int32             // Check clauses in the reloaded trace
	fPushClauses    int32             // Push clauses in the reloaded trace
	fMFFC           int32             // Refine the entire MFFC of a PPI
	fPdra           int32             // Use pdr -nct
	fLoadTrace      int32             // Load previous traces if any
	fProofRefine    int32             // Use proof-based refinement
	fHybrid         int32             // Use a hybrid of CBR and PBR
	fCheckCombUnsat int32             // Check if ABS becomes comb. unsat
	fAbs2           int32             // Use UFAR style createAbs
	fProofUsePPI    int32             // Use PPI values in PBR
	fUseBmc3        int32             // Run BMC3 in parallel
	fShrinkAbs      int32             // Shrink Abs with BMC
	fShrinkScratch  int32             // Restart pdr from scratch after shrinking
	fVerbose        int32             // verbose output
	fPdrVerbose     int32             // verbose output
	RunId           int32             // id in this run
	pFuncStop       func(int32) int32 // callback to terminate
}

type Wlc_BstPar_t struct {
	iOutput          int32
	nOutputRange     int32
	nAdderLimit      int32
	nMultLimit       int32
	fGiaSimple       int32
	fAddOutputs      int32
	fMulti           int32
	fBooth           int32
	fNonRest         int32
	fCla             int32
	fDivBy0          int32
	fNoCleanup       int32
	fCreateMiter     int32
	fCreateWordMiter int32
	fDecMuxes        int32
	fSaveFfNames     int32
	fVerbose         int32
	vBoxIds          *Vec_Int_t
}

func Wlc_BstParDefault(pPar *Wlc_BstPar_t) {
	*pPar = Wlc_BstPar_t{}
	pPar.iOutput = -1
	pPar.nOutputRange = 0
	pPar.nAdderLimit = 0
	pPar.nMultLimit = 0
	pPar.fGiaSimple = 0
	pPar.fAddOutputs = 0
	pPar.fMulti = 0
	pPar.fBooth = 0
	pPar.fCla = 0
	pPar.fDivBy0 = 0
	pPar.fCreateMiter = 0
	pPar.fCreateWordMiter = 0
	pPar.fDecMuxes = 0
	pPar.fVerbose = 0
}

type Wla_Man_t struct {
	p        *Wlc_Ntk_t
	pPars    *Wlc_Par_t
	vClauses *Vec_Vec_t
	vBlacks  *Vec_Int_t
	vSignals *Vec_Int_t
	pCex     *Abc_Cex_t
	pGia     *Gia_Man_t
	vUnmark  *Vec_Bit_t
	pPdrPars any
	pThread  any

	iCexFrame int32
	fNewAbs   int32

	nIters    int32
	nTotalCla int32
	nDisj     int32
	nNDisj    int32

	tPdr abctime
	tCbr abctime
	tPbr abctime
}

func Wlc_NtkObjNum(p *Wlc_Ntk_t) int32    { return p.iObj - 1 }
func Wlc_NtkObjNumMax(p *Wlc_Ntk_t) int32 { return p.iObj }
func Wlc_NtkPiNum(p *Wlc_Ntk_t) int32     { return Vec_IntSize(&p.vPis) }
func Wlc_NtkPoNum(p *Wlc_Ntk_t) int32     { return Vec_IntSize(&p.vPos) }
func Wlc_NtkCiNum(p *Wlc_Ntk_t) int32     { return Vec_IntSize(&p.vCis) }
func Wlc_NtkCoNum(p *Wlc_Ntk_t) int32     { return Vec_IntSize(&p.vCos) }
func Wlc_NtkFfNum(p *Wlc_Ntk_t) int32     { return Vec_IntSize(&p.vCis) - Vec_IntSize(&p.vPis) }

func Wlc_NtkObj(p *Wlc_Ntk_t, Id int32) *Wlc_Obj_t {
	assert(Id > 0 && Id < p.nObjsAlloc, "Id > 0 && Id < p.nObjsAlloc")
	return (*Wlc_Obj_t)(unsafe.Add(unsafe.Pointer(p.pObjs), uintptr(Id)*unsafe.Sizeof(Wlc_Obj_t{})))
}
func Wlc_NtkPi(p *Wlc_Ntk_t, i int32) *Wlc_Obj_t  { return Wlc_NtkObj(p, Vec_IntEntry(&p.vPis, i)) }
func Wlc_NtkPo(p *Wlc_Ntk_t, i int32) *Wlc_Obj_t  { return Wlc_NtkObj(p, Vec_IntEntry(&p.vPos, i)) }
func Wlc_NtkCi(p *Wlc_Ntk_t, i int32) *Wlc_Obj_t  { return Wlc_NtkObj(p, Vec_IntEntry(&p.vCis, i)) }
func Wlc_NtkCo(p *Wlc_Ntk_t, i int32) *Wlc_Obj_t  { return Wlc_NtkObj(p, Vec_IntEntry(&p.vCos, i)) }
func Wlc_NtkFf(p *Wlc_Ntk_t, i int32) *Wlc_Obj_t  { return Wlc_NtkObj(p, Vec_IntEntry(&p.vFfs, i)) }
func Wlc_NtkFf2(p *Wlc_Ntk_t, i int32) *Wlc_Obj_t { return Wlc_NtkObj(p, Vec_IntEntry(&p.vFfs2, i)) }

func Wlc_ObjIsPi(p *Wlc_Obj_t) int32 { return boolToInt32(int32(p.Type) == int32(WLC_OBJ_PI)) }
func Wlc_ObjIsPo(p *Wlc_Obj_t) int32 { return int32(p.fIsPo) }
func Wlc_ObjIsCi(p *Wlc_Obj_t) int32 {
	return boolToInt32(int32(p.Type) == int32(WLC_OBJ_PI) || int32(p.Type) == int32(WLC_OBJ_FO))
}
func Wlc_ObjIsCo(p *Wlc_Obj_t) int32    { return boolToInt32(p.fIsPo != 0 || p.fIsFi != 0) }
func Wlc_ObjIsRead(p *Wlc_Obj_t) int32  { return boolToInt32(int32(p.Type) == int32(WLC_OBJ_READ)) }
func Wlc_ObjIsWrite(p *Wlc_Obj_t) int32 { return boolToInt32(int32(p.Type) == int32(WLC_OBJ_WRITE)) }
func Wlc_ObjIsMux(p *Wlc_Obj_t) int32   { return boolToInt32(int32(p.Type) == int32(WLC_OBJ_MUX)) }
func Wlc_ObjIsBuf(p *Wlc_Obj_t) int32   { return boolToInt32(int32(p.Type) == int32(WLC_OBJ_BUF)) }
func Wlc_ObjIsFf(p *Wlc_Ntk_t, i int32) int32 {
	return boolToInt32(int32(Wlc_NtkObj(p, i).Type) == int32(WLC_OBJ_FF))
}

func Wlc_ObjId(p *Wlc_Ntk_t, pObj *Wlc_Obj_t) int32 {
	return int32((uintptr(unsafe.Pointer(pObj)) - uintptr(unsafe.Pointer(p.pObjs))) / unsafe.Sizeof(Wlc_Obj_t{}))
}
func Wlc_ObjCiId(p *Wlc_Obj_t) int32 {
	assert((Wlc_ObjIsCi(p) != 0), "Wlc_ObjIsCi(p)")
	return p.Fanins[1]
}
func Wlc_ObjType(pObj *Wlc_Obj_t) int32  { return int32(pObj.Type) }
func Wlc_ObjFaninNum(p *Wlc_Obj_t) int32 { return int32(p.nFanins) }
func Wlc_ObjHasArray(p *Wlc_Obj_t) int32 {
	return boolToInt32(p.nFanins > 2 || int32(p.Type) == int32(WLC_OBJ_CONST) || int32(p.Type) == int32(WLC_OBJ_BIT_SELECT))
}
func Wlc_ObjFanins(p *Wlc_Obj_t) []int32 {
	if Wlc_ObjHasArray(p) != 0 {
		return p.pFanins[0]
	}
	return p.Fanins[:]
}
func Wlc_ObjFaninId(p *Wlc_Obj_t, i int32) int32 { return Wlc_ObjFanins(p)[i] }
func Wlc_ObjFaninId0(p *Wlc_Obj_t) int32         { return Wlc_ObjFanins(p)[0] }
func Wlc_ObjFaninId1(p *Wlc_Obj_t) int32         { return Wlc_ObjFanins(p)[1] }
func Wlc_ObjFaninId2(p *Wlc_Obj_t) int32         { return Wlc_ObjFanins(p)[2] }
func Wlc_ObjFanin(p *Wlc_Ntk_t, pObj *Wlc_Obj_t, i int32) *Wlc_Obj_t {
	return Wlc_NtkObj(p, Wlc_ObjFaninId(pObj, i))
}
func Wlc_ObjFanin0(p *Wlc_Ntk_t, pObj *Wlc_Obj_t) *Wlc_Obj_t {
	return Wlc_NtkObj(p, Wlc_ObjFaninId(pObj, 0))
}
func Wlc_ObjFanin1(p *Wlc_Ntk_t, pObj *Wlc_Obj_t) *Wlc_Obj_t {
	return Wlc_NtkObj(p, Wlc_ObjFaninId(pObj, 1))
}
func Wlc_ObjFanin2(p *Wlc_Ntk_t, pObj *Wlc_Obj_t) *Wlc_Obj_t {
	return Wlc_NtkObj(p, Wlc_ObjFaninId(pObj, 2))
}

func Wlc_ObjRange(p *Wlc_Obj_t) int32 {
	if p.End >= p.Beg {
		return 1 + p.End - p.Beg
	}
	return 1 + p.Beg - p.End
}
func Wlc_ObjRangeEnd(p *Wlc_Obj_t) int32 {
	assert(int32(p.Type) == int32(WLC_OBJ_BIT_SELECT), "int32(p.Type) == int32(WLC_OBJ_BIT_SELECT)")
	return p.pFanins[0][1]
}
func Wlc_ObjRangeBeg(p *Wlc_Obj_t) int32 {
	assert(int32(p.Type) == int32(WLC_OBJ_BIT_SELECT), "int32(p.Type) == int32(WLC_OBJ_BIT_SELECT)")
	return p.pFanins[0][2]
}
func Wlc_ObjRangeIsReversed(p *Wlc_Obj_t) int32 { return boolToInt32(p.End < p.Beg) }

func Wlc_ObjIsSigned(p *Wlc_Obj_t) int32 { return int32(p.Signed) }
func Wlc_ObjIsSignedFanin01(p *Wlc_Ntk_t, pObj *Wlc_Obj_t) int32 {
	if p.fSmtLib != 0 {
		return Wlc_ObjIsSigned(pObj)
	} else {
		return boolToInt32(Wlc_ObjFanin0(p, pObj).Signed != 0 && Wlc_ObjFanin1(p, pObj).Signed != 0)
	}
}
func Wlc_ObjIsSignedFanin0(p *Wlc_Ntk_t, pObj *Wlc_Obj_t) int32 {
	if p.fSmtLib != 0 {
		return Wlc_ObjIsSigned(pObj)
	} else {
		return int32(Wlc_ObjFanin0(p, pObj).Signed)
	}
}
func Wlc_ObjIsSignedFanin1(p *Wlc_Ntk_t, pObj *Wlc_Obj_t) int32 {
	if p.fSmtLib != 0 {
		return Wlc_ObjIsSigned(pObj)
	} else {
		return int32(Wlc_ObjFanin1(p, pObj).Signed)
	}
}
func Wlc_ObjSign(p *Wlc_Obj_t) int32 { return Abc_Var2Lit(Wlc_ObjRange(p), Wlc_ObjIsSigned(p)) }
func Wlc_ObjConstValue(p *Wlc_Obj_t) *int32 {
	assert(int32(p.Type) == int32(WLC_OBJ_CONST), "int32(p.Type) == int32(WLC_OBJ_CONST)")
	return &Wlc_ObjFanins(p)[0]
}
func Wlc_ObjTableId(p *Wlc_Obj_t) int32 {
	assert(int32(p.Type) == int32(WLC_OBJ_TABLE), "int32(p.Type) == int32(WLC_OBJ_TABLE)")
	return p.Fanins[1]
}
func Wlc_ObjTable(p *Wlc_Ntk_t, pObj *Wlc_Obj_t) *word {
	return Vec_PtrEntry(p.vTables, Wlc_ObjTableId(pObj)).(*word)
}
func Wlc_ObjLevelId(p *Wlc_Ntk_t, iObj int32) int32    { return Vec_IntEntry(&p.vLevels, iObj) }
func Wlc_ObjLevel(p *Wlc_Ntk_t, pObj *Wlc_Obj_t) int32 { return Wlc_ObjLevelId(p, Wlc_ObjId(p, pObj)) }

func Wlc_NtkCleanCopy(p *Wlc_Ntk_t)                    { Vec_IntFill(&p.vCopies, p.nObjsAlloc, 0) }
func Wlc_NtkHasCopy(p *Wlc_Ntk_t) int32                { return boolToInt32(Vec_IntSize(&p.vCopies) > 0) }
func Wlc_ObjSetCopy(p *Wlc_Ntk_t, iObj int32, i int32) { Vec_IntWriteEntry(&p.vCopies, iObj, i) }
func Wlc_ObjCopy(p *Wlc_Ntk_t, iObj int32) int32       { return Vec_IntEntry(&p.vCopies, iObj) }
func Wlc_ObjCopyObj(pNew *Wlc_Ntk_t, p *Wlc_Ntk_t, pObj *Wlc_Obj_t) *Wlc_Obj_t {
	return Wlc_NtkObj(pNew, Wlc_ObjCopy(p, Wlc_ObjId(p, pObj)))
}

func Wlc_NtkCleanNameId(p *Wlc_Ntk_t)                    { Vec_IntFill(&p.vNameIds, p.nObjsAlloc, 0) }
func Wlc_NtkHasNameId(p *Wlc_Ntk_t) int32                { return boolToInt32(Vec_IntSize(&p.vNameIds) > 0) }
func Wlc_ObjSetNameId(p *Wlc_Ntk_t, iObj int32, i int32) { Vec_IntWriteEntry(&p.vNameIds, iObj, i) }
func Wlc_ObjNameId(p *Wlc_Ntk_t, iObj int32) int32       { return Vec_IntEntry(&p.vNameIds, iObj) }

func Wlc_ObjFo2Fi(p *Wlc_Ntk_t, pObj *Wlc_Obj_t) *Wlc_Obj_t {
	assert(int32(pObj.Type) == int32(WLC_OBJ_FO), "int32(pObj.Type) == int32(WLC_OBJ_FO)")
	return Wlc_NtkCo(p, Wlc_NtkPoNum(p)+Wlc_ObjCiId(pObj)-Wlc_NtkPiNum(p))
}
func Wlc_ObjCo2PoFo(p *Wlc_Ntk_t, iCoId int32) *Wlc_Obj_t {
	if iCoId < Wlc_NtkPoNum(p) {
		return Wlc_NtkPo(p, iCoId)
	} else {
		return Wlc_NtkCi(p, Wlc_NtkPiNum(p)+iCoId-Wlc_NtkPoNum(p))
	}
}

////////////////////////////////////////////////////////////////////////
///                      MACRO DEFINITIONS                           ///
////////////////////////////////////////////////////////////////////////

////////////////////////////////////////////////////////////////////////
///                             ITERATORS                            ///
////////////////////////////////////////////////////////////////////////

// Original macro:
// #define Wlc_NtkForEachObj( p, pObj, i )                                             \
//     for ( i = 1; (i < Wlc_NtkObjNumMax(p)) && (((pObj) = Wlc_NtkObj(p, i)), 1); i++ )

// Original macro:
// #define Wlc_NtkForEachObjReverse( p, pObj, i )                                      \
//     for ( i = Wlc_NtkObjNumMax(p) - 1; (i > 0) && (((pObj) = Wlc_NtkObj(p, i)), 1); i-- )

// Original macro:
// #define Wlc_NtkForEachObjVec( vVec, p, pObj, i )                                    \
//     for ( i = 0; (i < Vec_IntSize(vVec)) && (((pObj) = Wlc_NtkObj(p, Vec_IntEntry(vVec, i))), 1); i++ )

// Original macro:
// #define Wlc_NtkForEachPi( p, pPi, i )                                               \
//     for ( i = 0; (i < Wlc_NtkPiNum(p)) && (((pPi) = Wlc_NtkPi(p, i)), 1); i++ )

// Original macro:
// #define Wlc_NtkForEachPo( p, pPo, i )                                               \
//     for ( i = 0; (i < Wlc_NtkPoNum(p)) && (((pPo) = Wlc_NtkPo(p, i)), 1); i++ )

// Original macro:
// #define Wlc_NtkForEachCi( p, pCi, i )                                               \
//     for ( i = 0; (i < Wlc_NtkCiNum(p)) && (((pCi) = Wlc_NtkCi(p, i)), 1); i++ )

// Original macro:
// #define Wlc_NtkForEachCo( p, pCo, i )                                               \
//     for ( i = 0; (i < Wlc_NtkCoNum(p)) && (((pCo) = Wlc_NtkCo(p, i)), 1); i++ )

// Original macro:
// #define Wlc_NtkForEachFf( p, pFf, i )                                               \
//     for ( i = 0; (i < Vec_IntSize(&p->vFfs)) && (((pFf) = Wlc_NtkFf(p, i)), 1); i++ )

// Original macro:
// #define Wlc_NtkForEachFf2( p, pFf, i )                                              \
//     for ( i = 0; (i < Vec_IntSize(&p->vFfs2)) && (((pFf) = Wlc_NtkFf2(p, i)), 1); i++ )

// Original macro:
// #define Wlc_ObjForEachFanin( pObj, iFanin, i )                                      \
//     for ( i = 0; (i < Wlc_ObjFaninNum(pObj)) && (((iFanin) = Wlc_ObjFaninId(pObj, i)), 1); i++ )

// Original macro:
// #define Wlc_ObjForEachFaninObj( p, pObj, pFanin, i )                                \
//     for ( i = 0; (i < Wlc_ObjFaninNum(pObj)) && (((pFanin) = Wlc_NtkObj(p, Wlc_ObjFaninId(pObj, i))), 1); i++ )

// Original macro:
// #define Wlc_ObjForEachFaninReverse( pObj, iFanin, i )                               \
//     for ( i = Wlc_ObjFaninNum(pObj) - 1; (i >= 0) && (((iFanin) = Wlc_ObjFaninId(pObj, i)), 1); i-- )

////////////////////////////////////////////////////////////////////////
///                    FUNCTION DECLARATIONS                         ///
////////////////////////////////////////////////////////////////////////

/*=== wlcAbs.c ========================================================*/
// extern int            Wlc_NtkAbsCore( Wlc_Ntk_t * p, Wlc_Par_t * pPars );

// extern int            Wlc_NtkPdrAbs( Wlc_Ntk_t * p, Wlc_Par_t * pPars );

/*=== wlcAbs2.c ========================================================*/
// extern int            Wlc_NtkAbsCore2( Wlc_Ntk_t * p, Wlc_Par_t * pPars );

/*=== wlcBlast.c ========================================================*/
// extern Gia_Man_t *    Wlc_NtkBitBlast( Wlc_Ntk_t * p, Wlc_BstPar_t * pPars );

/*=== wlcCom.c ========================================================*/
// extern void           Wlc_SetNtk( Abc_Frame_t * pAbc, Wlc_Ntk_t * pNtk );

/*=== wlcMem.c ========================================================*/
// extern Vec_Int_t *    Wlc_NtkCollectMemory( Wlc_Ntk_t * p, int fClean );

// extern void           Wlc_NtkPrintMemory( Wlc_Ntk_t * p );

// extern Wlc_Ntk_t *    Wlc_NtkMemAbstractTest( Wlc_Ntk_t * p );

// extern int            Wlc_NtkMemAbstract( Wlc_Ntk_t * p, int nIterMax, int fDumpAbs, int fPdrVerbose, int fVerbose );

// extern Wlc_Ntk_t *    Wlc_NtkAbstractMem( Wlc_Ntk_t * p, int nFrames, int fVerbose );

/*=== wlcNdr.c ========================================================*/
// extern Wlc_Ntk_t *    Wlc_ReadNdr( char * pFileName );

// extern void           Wlc_WriteNdr( Wlc_Ntk_t * pNtk, char * pFileName );

// extern Wlc_Ntk_t *    Wlc_NtkFromNdr( void * pData );

// extern void *         Wlc_NtkToNdr( Wlc_Ntk_t * pNtk );

/*=== wlcNtk.c ========================================================*/
// extern void           Wlc_ManSetDefaultParams( Wlc_Par_t * pPars );

// extern char *         Wlc_ObjTypeName( Wlc_Obj_t * p );

// extern Wlc_Ntk_t *    Wlc_NtkAlloc( char * pName, int nObjsAlloc );

// extern int            Wlc_ObjAlloc( Wlc_Ntk_t * p, int Type, int Signed, int End, int Beg );

// extern int            Wlc_ObjCreate( Wlc_Ntk_t * p, int Type, int Signed, int End, int Beg, Vec_Int_t * vFanins );

// extern void           Wlc_ObjSetCi( Wlc_Ntk_t * p, Wlc_Obj_t * pObj );

// extern void           Wlc_ObjSetCo( Wlc_Ntk_t * p, Wlc_Obj_t * pObj, int fFlopInput );

// extern char *         Wlc_ObjName( Wlc_Ntk_t * p, int iObj );

// extern void           Wlc_ObjUpdateType( Wlc_Ntk_t * p, Wlc_Obj_t * pObj, int Type );

// extern void           Wlc_ObjAddFanins( Wlc_Ntk_t * p, Wlc_Obj_t * pObj, Vec_Int_t * vFanins );

// extern int            Wlc_ObjDup( Wlc_Ntk_t * pNew, Wlc_Ntk_t * p, int iObj, Vec_Int_t * vFanins );

// extern void           Wlc_NtkFree( Wlc_Ntk_t * p );

// extern int            Wlc_NtkCreateLevels( Wlc_Ntk_t * p );

// extern int            Wlc_NtkCreateLevelsRev( Wlc_Ntk_t * p );

// extern int            Wlc_NtkRemapLevels( Wlc_Ntk_t * p, Vec_Int_t * vObjs, int nLevels );

// extern int            Wlc_NtkCountRealPis( Wlc_Ntk_t * p );

// extern void           Wlc_NtkPrintNode( Wlc_Ntk_t * p, Wlc_Obj_t * pObj );

// extern void           Wlc_NtkPrintNodeArray( Wlc_Ntk_t * p, Vec_Int_t * vArray );

// extern void           Wlc_NtkPrintNodes( Wlc_Ntk_t * p, int Type );

// extern void           Wlc_NtkPrintStats( Wlc_Ntk_t * p, int fDistrib, int fTwoSides, int fVerbose );

// extern void           Wlc_NtkPrintObjects( Wlc_Ntk_t * p );

// extern void           Wlc_NtkTransferNames( Wlc_Ntk_t * pNew, Wlc_Ntk_t * p );

// extern char *         Wlc_NtkNewName( Wlc_Ntk_t * p, int iCoId, int fSeq );

// extern Wlc_Ntk_t *    Wlc_NtkDupDfs( Wlc_Ntk_t * p, int fMarked, int fSeq );

// extern Wlc_Ntk_t *    Wlc_NtkDupDfsAbs( Wlc_Ntk_t * p, Vec_Int_t * vPisOld, Vec_Int_t * vPisNew, Vec_Int_t * vFlops );

// extern Wlc_Ntk_t *    Wlc_NtkDupDfsSimple( Wlc_Ntk_t * p );

// extern void           Wlc_NtkCleanMarks( Wlc_Ntk_t * p );

// extern void           Wlc_NtkMarkCone( Wlc_Ntk_t * p, int iCoId, int Range, int fSeq, int fAllPis );

// extern void           Wlc_NtkProfileCones( Wlc_Ntk_t * p );

// extern Wlc_Ntk_t *    Wlc_NtkDupSingleNodes( Wlc_Ntk_t * p );

// extern void           Wlc_NtkShortNames( Wlc_Ntk_t * p );

// extern int            Wlc_NtkDcFlopNum( Wlc_Ntk_t * p );

// extern void           Wlc_NtkSetRefs( Wlc_Ntk_t * p );

// extern int            Wlc_NtkCountObjBits( Wlc_Ntk_t * p, Vec_Int_t * vPisNew );

/*=== wlcReadSmt.c ========================================================*/
// extern Wlc_Ntk_t *    Wlc_ReadSmtBuffer( char * pFileName, char * pBuffer, char * pLimit, int fOldParser, int fPrintTree );

// extern Wlc_Ntk_t *    Wlc_ReadSmt( char * pFileName, int fOldParser, int fPrintTree );

/*=== wlcSim.c ========================================================*/
// extern Vec_Ptr_t *    Wlc_NtkSimulate( Wlc_Ntk_t * p, Vec_Int_t * vNodes, int nWords, int nFrames );

// extern void           Wlc_NtkDeleteSim( Vec_Ptr_t * p );

/*=== wlcStdin.c ========================================================*/
// extern int            Wlc_StdinProcessSmt( Abc_Frame_t * pAbc, char * pCmd );

/*=== wlcReadVer.c ========================================================*/
// extern char *         Wlc_PrsConvertInitValues( Wlc_Ntk_t * p );

// extern Wlc_Ntk_t *    Wlc_ReadVer( char * pFileName, char * pStr );

/*=== wlcUif.c ========================================================*/
// extern Vec_Int_t *    Wlc_NtkCollectAddMult( Wlc_Ntk_t * p, Wlc_BstPar_t * pPar, int * pCountA, int * CountM );

// extern int            Wlc_NtkPairIsUifable( Wlc_Ntk_t * p, Wlc_Obj_t * pObj, Wlc_Obj_t * pObj2 );

// extern Vec_Int_t *    Wlc_NtkCollectMultipliers( Wlc_Ntk_t * p );

// extern Vec_Int_t *    Wlc_NtkFindUifableMultiplierPairs( Wlc_Ntk_t * p );

// extern Wlc_Ntk_t *    Wlc_NtkAbstractNodes( Wlc_Ntk_t * pNtk, Vec_Int_t * vNodes );

// extern Wlc_Ntk_t *    Wlc_NtkUifNodePairs( Wlc_Ntk_t * pNtk, Vec_Int_t * vPairs );

/*=== wlcWin.c =============================================================*/
// extern void           Wlc_WinProfileArith( Wlc_Ntk_t * p );

/*=== wlcWriteVer.c ========================================================*/
// extern void           Wlc_WriteVer( Wlc_Ntk_t * p, char * pFileName, int fAddCos, int fNoFlops );

//ABC_NAMESPACE_HEADER_END

////////////////////////////////////////////////////////////////////////
///                       END OF FILE                                ///
////////////////////////////////////////////////////////////////////////
