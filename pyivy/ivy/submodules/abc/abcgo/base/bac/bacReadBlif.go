package bac

import (
	"fmt"
	"strings"

	"github.com/glycerine/abcgo/misc/util"
	"github.com/glycerine/abcgo/misc/vec"
)

/**CFile****************************************************************

  FileName    [bacReadBlif.c]

  SystemName  [ABC: Logic synthesis and verification system.]

  PackageName [Hierarchical word-level netlist.]

  Synopsis    [BLIF parser.]

  Author      [Alan Mishchenko]

  Affiliation [UC Berkeley]

  Date        [Ver. 1.0. Started - November 29, 2014.]

  Revision    [$Id: bacReadBlif.c,v 1.00 2014/11/29 00:00:00 alanmi Exp $]

***********************************************************************/

//#include "bac.h"
//#include "bacPrs.h"

//ABC_NAMESPACE_IMPL_START

////////////////////////////////////////////////////////////////////////
///                        DECLARATIONS                              ///
////////////////////////////////////////////////////////////////////////

// BLIF keywords
type Bac_BlifType_t int32

const (
	PRS_BLIF_NONE Bac_BlifType_t = 0
	// 0:   unused
	PRS_BLIF_MODEL Bac_BlifType_t = PRS_BLIF_NONE + 1
	// 1:   .model
	PRS_BLIF_INOUTS Bac_BlifType_t = PRS_BLIF_MODEL + 1
	// 2:   .inouts
	PRS_BLIF_INPUTS Bac_BlifType_t = PRS_BLIF_INOUTS + 1
	// 3:   .inputs
	PRS_BLIF_OUTPUTS Bac_BlifType_t = PRS_BLIF_INPUTS + 1
	// 4:   .outputs
	PRS_BLIF_NAMES Bac_BlifType_t = PRS_BLIF_OUTPUTS + 1
	// 5:   .names
	PRS_BLIF_SUBCKT Bac_BlifType_t = PRS_BLIF_NAMES + 1
	// 6:   .subckt
	PRS_BLIF_GATE Bac_BlifType_t = PRS_BLIF_SUBCKT + 1
	// 7:   .gate
	PRS_BLIF_LATCH Bac_BlifType_t = PRS_BLIF_GATE + 1
	// 8:   .latch
	PRS_BLIF_SHORT Bac_BlifType_t = PRS_BLIF_LATCH + 1
	// 9:   .short
	PRS_BLIF_END Bac_BlifType_t = PRS_BLIF_SHORT + 1
	// 10:  .end
	PRS_BLIF_UNKNOWN Bac_BlifType_t = PRS_BLIF_END + 1 // 11:  unknown

)

var s_BlifTypes = []string{
	"",         // 0:   unused
	".model",   // 1:   .model
	".inouts",  // 2:   .inputs
	".inputs",  // 3:   .inputs
	".outputs", // 4:   .outputs
	".names",   // 5:   .names
	".subckt",  // 6:   .subckt
	".gate",    // 7:   .gate
	".latch",   // 8:   .latch
	".short",   // 9:   .short
	".end",     // 10:  .end
	"",         // 11:  unknown
}

func Psr_NtkAddBlifDirectives(p *Psr_Man_t) {
	var i int32
	for i = 1; i < int32(len(s_BlifTypes)) && s_BlifTypes[i] != ""; i++ {
		Abc_NamStrFindOrAdd(p.pStrs, s_BlifTypes[i], nil)
	}
	assert(Abc_NamObjNumMax(p.pStrs) == i, "Abc_NamObjNumMax(p.pStrs) == i")
}

////////////////////////////////////////////////////////////////////////
///                     FUNCTION DEFINITIONS                         ///
////////////////////////////////////////////////////////////////////////

/*
*Function*************************************************************

	Synopsis    [Reading characters.]

	Description []

	SideEffects []

	SeeAlso     []

**********************************************************************
*/
func psrBool(v bool) int32 {
	if v {
		return 1
	}
	return 0
}
func Psr_BlifCharIsSpace(c byte) int32 { return psrBool(c == ' ' || c == '\t' || c == '\r') }
func Psr_CharIsStop(c byte) int32      { return psrBool(c == '#' || c == '\\' || c == '\n' || c == '=') }
func Psr_CharIsLit(c byte) int32       { return psrBool(c == '0' || c == '1' || c == '-') }

func Psr_ManIsSpace(p *Psr_Man_t) int32 { return Psr_BlifCharIsSpace(p.pCur[0]) }
func Psr_ManIsStop(p *Psr_Man_t) int32  { return Psr_CharIsStop(p.pCur[0]) }
func Psr_ManIsLit(p *Psr_Man_t) int32   { return Psr_CharIsLit(p.pCur[0]) }

func Psr_BlifManIsChar(p *Psr_Man_t, c byte) int32 { return psrBool(p.pCur[0] == c) }
func Psr_BlifManIsChar2(p *Psr_Man_t, c byte) int32 {
	if p.pCur[0] != c {
		return 0
	}
	p.pCur = p.pCur[1:]
	return 1
}

func Psr_ManSkip(p *Psr_Man_t) {
	if len(p.pCur) > 0 {
		p.pCur = p.pCur[1:]
	}
}
func Psr_ManSkip2(p *Psr_Man_t) byte { c := p.pCur[0]; Psr_ManSkip(p); return c }

/*
*Function*************************************************************

	Synopsis    [Reading names.]

	Description []

	SideEffects []

	SeeAlso     []

**********************************************************************
*/
func Psr_ManSkipToChar(p *Psr_Man_t, c byte) {
	for !(Psr_BlifManIsChar(p, c) != 0) {
		Psr_ManSkip(p)
	}
}
func Psr_ManSkipSpaces(p *Psr_Man_t) {
	for {
		for Psr_ManIsSpace(p) != 0 {
			Psr_ManSkip(p)
		}
		if Psr_BlifManIsChar(p, '\\') != 0 {
			Psr_ManSkipToChar(p, '\n')
			Psr_ManSkip(p)
			continue
		}
		if Psr_BlifManIsChar(p, '#') != 0 {
			Psr_ManSkipToChar(p, '\n')
		}
		break
	}
	assert(!(Psr_ManIsSpace(p) != 0), "!Psr_ManIsSpace(p)")
}
func Psr_BlifReadName(p *Psr_Man_t) int32 {
	var pStart string
	Psr_ManSkipSpaces(p)
	if Psr_BlifManIsChar(p, '\n') != 0 {
		return 0
	}
	pStart = p.pCur
	for !(Psr_ManIsSpace(p) != 0) && !(Psr_ManIsStop(p) != 0) {
		Psr_ManSkip(p)
	}
	if pStart == p.pCur {
		return 0
	}
	return Abc_NamStrFindOrAddLim(p.pStrs, pStart, len(pStart)-len(p.pCur), nil)
}
func Psr_ManReadList(p *Psr_Man_t, vOrder *vec.Vec_Int_t, Type int32) int32 {
	var iToken int32
	Vec_IntClear(&p.vTemp)
	for iToken = Psr_BlifReadName(p); iToken != 0; iToken = Psr_BlifReadName(p) {
		Vec_IntPush(&p.vTemp, iToken)
		Vec_IntPush(vOrder, Abc_Var2Lit2(iToken, Type))
	}
	if Vec_IntSize(&p.vTemp) == 0 {
		return Psr_ManErrorSet(p, "Signal list is empty.", 1)
	}
	return 0
}
func Psr_ManReadList2(p *Psr_Man_t) int32 {
	var iToken int32
	Vec_IntClear(&p.vTemp)
	for iToken = Psr_BlifReadName(p); iToken != 0; iToken = Psr_BlifReadName(p) {
		Vec_IntPushTwo(&p.vTemp, 0, iToken)
	}
	if Vec_IntSize(&p.vTemp) == 0 {
		return Psr_ManErrorSet(p, "Signal list is empty.", 1)
	}
	return 0
}
func Psr_ManReadList3(p *Psr_Man_t) int32 {
	Vec_IntClear(&p.vTemp)
	for !(Psr_BlifManIsChar(p, '\n') != 0) {
		var iToken int32 = Psr_BlifReadName(p)
		if iToken == 0 {
			return Psr_ManErrorSet(p, "Cannot read formal name.", 1)
		}
		Vec_IntPush(&p.vTemp, iToken)
		Psr_ManSkipSpaces(p)
		if !(Psr_BlifManIsChar2(p, '=') != 0) {
			return Psr_ManErrorSet(p, "Cannot find symbol \"=\".", 1)
		}
		iToken = Psr_BlifReadName(p)
		if iToken == 0 {
			return Psr_ManErrorSet(p, "Cannot read actual name.", 1)
		}
		Vec_IntPush(&p.vTemp, iToken)
		Psr_ManSkipSpaces(p)
	}
	if Vec_IntSize(&p.vTemp) == 0 {
		return Psr_ManErrorSet(p, "Cannot read a list of formal/actual names.", 1)
	}
	if Vec_IntSize(&p.vTemp)%2 != 0 {
		return Psr_ManErrorSet(p, "The number of formal/actual names is not even.", 1)
	}
	return 0
}

/*
*Function*************************************************************

	Synopsis    []

	Description []

	SideEffects []

	SeeAlso     []

**********************************************************************
*/
func Psr_ManReadCube(p *Psr_Man_t) int32 {
	assert((Psr_ManIsLit(p) != 0), "Psr_ManIsLit(p)")
	for Psr_ManIsLit(p) != 0 {
		Vec_StrPush(&p.vCover, Psr_ManSkip2(p))
	}
	Psr_ManSkipSpaces(p)
	if Psr_BlifManIsChar(p, '\n') != 0 {
		if Vec_StrSize(&p.vCover) != 1 {
			return Psr_ManErrorSet(p, "Cannot read cube.", 1)
		}
		// fix single literal cube by adding space
		Vec_StrPush(&p.vCover, Vec_StrEntry(&p.vCover, 0))
		Vec_StrWriteEntry(&p.vCover, 0, ' ')
		Vec_StrPush(&p.vCover, '\n')
		return 0
	}
	if !(Psr_ManIsLit(p) != 0) {
		return Psr_ManErrorSet(p, "Cannot read output literal.", 1)
	}
	Vec_StrPush(&p.vCover, ' ')
	Vec_StrPush(&p.vCover, Psr_ManSkip2(p))
	Vec_StrPush(&p.vCover, '\n')
	Psr_ManSkipSpaces(p)
	if !(Psr_BlifManIsChar(p, '\n') != 0) {
		return Psr_ManErrorSet(p, "Cannot read end of cube.", 1)
	}
	return 0
}
func Psr_ManSaveCover(p *Psr_Man_t) {
	var iToken int32
	if Vec_StrSize(&p.vCover) == 0 {
		p.pNtk.fHasC0s = 1
	} else if Vec_StrSize(&p.vCover) == 2 {
		if Vec_StrEntryLast(&p.vCover) == '0' {
			p.pNtk.fHasC0s = 1
		} else if Vec_StrEntryLast(&p.vCover) == '1' {
			p.pNtk.fHasC1s = 1
		} else {
			assert(false, "0")
		}
	}
	assert(Vec_StrSize(&p.vCover) > 0, "Vec_StrSize(&p.vCover) > 0")
	Vec_StrPush(&p.vCover, 0)
	//iToken = Abc_NamStrFindOrAdd( p->pStrs, Vec_StrArray(&p->vCover), NULL );
	iToken = int32(Ptr_SopToType(strings.TrimRight(string(vec.Vec_StrArray(&p.vCover)), "\x00")))
	Vec_StrClear(&p.vCover)
	// set the cover to the module of this box
	assert(Psr_BoxNtk(p.pNtk, Psr_NtkBoxNum(p.pNtk)-1) == 1, "Psr_BoxNtk(p.pNtk, Psr_NtkBoxNum(p.pNtk)-1) == 1") // default const 0
	Psr_BoxSetNtk(p.pNtk, Psr_NtkBoxNum(p.pNtk)-1, iToken)
}

/*
*Function*************************************************************

	Synopsis    []

	Description []

	SideEffects []

	SeeAlso     []

**********************************************************************
*/
func Psr_ManReadInouts(p *Psr_Man_t) int32 {
	if Psr_ManReadList(p, &p.pNtk.vOrder, 3) != 0 {
		return 1
	}
	Vec_IntAppend(&p.pNtk.vInouts, &p.vTemp)
	return 0
}
func Psr_ManReadInputs(p *Psr_Man_t) int32 {
	if Psr_ManReadList(p, &p.pNtk.vOrder, 1) != 0 {
		return 1
	}
	Vec_IntAppend(&p.pNtk.vInputs, &p.vTemp)
	return 0
}
func Psr_ManReadOutputs(p *Psr_Man_t) int32 {
	if Psr_ManReadList(p, &p.pNtk.vOrder, 2) != 0 {
		return 1
	}
	Vec_IntAppend(&p.pNtk.vOutputs, &p.vTemp)
	return 0
}
func Psr_ManReadNode(p *Psr_Man_t) int32 {
	if Psr_ManReadList2(p) != 0 {
		return 1
	}
	// save results
	Psr_NtkAddBox(p.pNtk, 1, 0, &p.vTemp) // default const 0 function
	return 0
}
func Psr_ManReadBox(p *Psr_Man_t, fGate int32) int32 {
	var iToken int32 = Psr_BlifReadName(p)
	if iToken == 0 {
		return Psr_ManErrorSet(p, "Cannot read model name.", 1)
	}
	if Psr_ManReadList3(p) != 0 {
		return 1
	}
	// save results
	Psr_NtkAddBox(p.pNtk, iToken, 0, &p.vTemp)
	if fGate != 0 {
		p.pNtk.fMapped = 1
	}
	return 0
}
func Psr_ManReadLatch(p *Psr_Man_t) int32 {
	var iToken int32 = Psr_BlifReadName(p)
	Vec_IntClear(&p.vTemp)
	if iToken == 0 {
		return Psr_ManErrorSet(p, "Cannot read latch input.", 1)
	}
	Vec_IntWriteEntry(&p.vTemp, 1, iToken)
	iToken = Psr_BlifReadName(p)
	if iToken == 0 {
		return Psr_ManErrorSet(p, "Cannot read latch output.", 1)
	}
	Vec_IntWriteEntry(&p.vTemp, 0, iToken)
	Psr_ManSkipSpaces(p)
	if Psr_BlifManIsChar(p, '0') != 0 {
		iToken = 0
	} else if Psr_BlifManIsChar(p, '1') != 0 {
		iToken = 1
	} else {
		iToken = 2
	}
	Psr_ManSkipToChar(p, '\n')
	// save results
	Psr_NtkAddBox(p.pNtk, -1, iToken, &p.vTemp) // -1 stands for latch
	return 0
}
func Psr_ManReadShort(p *Psr_Man_t) int32 {
	var iToken int32 = Psr_BlifReadName(p)
	Vec_IntClear(&p.vTemp)
	if iToken == 0 {
		return Psr_ManErrorSet(p, "Cannot read .short input.", 1)
	}
	Vec_IntWriteEntry(&p.vTemp, 1, iToken)
	iToken = Psr_BlifReadName(p)
	if iToken == 0 {
		return Psr_ManErrorSet(p, "Cannot read .short output.", 1)
	}
	Vec_IntWriteEntry(&p.vTemp, 0, iToken)
	Psr_ManSkipSpaces(p)
	if !(Psr_BlifManIsChar(p, '\n') != 0) {
		return Psr_ManErrorSet(p, "Trailing symbols on .short line.", 1)
	}
	// save results
	iToken = Abc_NamStrFindOrAdd(p.pStrs, "1 1\n", nil)
	Psr_NtkAddBox(p.pNtk, iToken, 0, &p.vTemp)
	return 0
}
func Psr_ManReadModel(p *Psr_Man_t) int32 {
	var iToken int32
	if p.pNtk != nil {
		return Psr_ManErrorSet(p, "Parsing previous model is unfinished.", 1)
	}
	iToken = Psr_BlifReadName(p)
	if iToken == 0 {
		return Psr_ManErrorSet(p, "Cannot read model name.", 1)
	}
	Psr_ManInitializeNtk(p, iToken, 0)
	Psr_ManSkipSpaces(p)
	if !(Psr_BlifManIsChar(p, '\n') != 0) {
		return Psr_ManErrorSet(p, "Trailing symbols on .model line.", 1)
	}
	return 0
}
func Psr_ManReadEnd(p *Psr_Man_t) int32 {
	if p.pNtk == nil {
		return Psr_ManErrorSet(p, "Directive .end without .model.", 1)
	}
	//printf( "Saving model \"%s\".\n", Abc_NamStr(p->pStrs, p->iModuleName) );
	Psr_ManFinalizeNtk(p)
	Psr_ManSkipSpaces(p)
	if !(Psr_BlifManIsChar(p, '\n') != 0) {
		return Psr_ManErrorSet(p, "Trailing symbols on .end line.", 1)
	}
	return 0
}

func Psr_ManReadDirective(p *Psr_Man_t) int32 {
	var iToken int32
	if !(Psr_BlifManIsChar(p, '.') != 0) {
		return Psr_ManReadCube(p)
	}
	if Vec_StrSize(&p.vCover) > 0 { // SOP was specified for the previous node
		Psr_ManSaveCover(p)
	}
	iToken = Psr_BlifReadName(p)
	if iToken == int32(PRS_BLIF_MODEL) {
		return Psr_ManReadModel(p)
	}
	if iToken == int32(PRS_BLIF_INOUTS) {
		return Psr_ManReadInouts(p)
	}
	if iToken == int32(PRS_BLIF_INPUTS) {
		return Psr_ManReadInputs(p)
	}
	if iToken == int32(PRS_BLIF_OUTPUTS) {
		return Psr_ManReadOutputs(p)
	}
	if iToken == int32(PRS_BLIF_NAMES) {
		return Psr_ManReadNode(p)
	}
	if iToken == int32(PRS_BLIF_SUBCKT) {
		return Psr_ManReadBox(p, 0)
	}
	if iToken == int32(PRS_BLIF_GATE) {
		return Psr_ManReadBox(p, 1)
	}
	if iToken == int32(PRS_BLIF_LATCH) {
		return Psr_ManReadLatch(p)
	}
	if iToken == int32(PRS_BLIF_SHORT) {
		return Psr_ManReadShort(p)
	}
	if iToken == int32(PRS_BLIF_END) {
		return Psr_ManReadEnd(p)
	}
	fmt.Printf("Cannot read directive \"%s\".\n", Abc_NamStr(p.pStrs, iToken))
	return 1
}
func Psr_ManReadLines(p *Psr_Man_t) int32 {
	for p.pCur[1] != 0 {
		assert(Psr_BlifManIsChar(p, '\n') != 0, "Psr_BlifManIsChar(p, '\\n')")
		Psr_ManSkip(p)
		Psr_ManSkipSpaces(p)
		if Psr_BlifManIsChar(p, '\n') != 0 {
			continue
		}
		if Psr_ManReadDirective(p) != 0 {
			return 1
		}
	}
	return 0
}

/*
*Function*************************************************************

	Synopsis    []

	Description []

	SideEffects []

	SeeAlso     []

**********************************************************************
*/
func Psr_ManReadBlif(pFileName string) *vec.Vec_Ptr_t {
	var vPrs *vec.Vec_Ptr_t = nil
	var p *Psr_Man_t = Psr_ManAlloc(pFileName)
	if p == nil {
		return nil
	}
	Psr_NtkAddBlifDirectives(p)
	Psr_ManReadLines(p)
	if Psr_ManErrorPrint(p) != 0 {
		// inlining of macro ABC_SWAP
		var t *vec.Vec_Ptr_t = vPrs
		vPrs = p.vNtks
		p.vNtks = t
	}
	Psr_ManFree(p)
	return vPrs
}

/*
*Function*************************************************************

	Synopsis    []

	Description []

	SideEffects []

	SeeAlso     []

**********************************************************************
*/
func Psr_ManReadBlifTest() {
	var clk int64 = util.Abc_Clock()
	//    Vec_Ptr_t * vPrs = Psr_ManReadBlif( "aga/ray/ray_hie_oper.blif" );
	var vPrs *vec.Vec_Ptr_t = Psr_ManReadBlif("c/hie/dump/1/netlist_1_out8.blif")
	if !(vPrs != nil) {
		return
	}
	fmt.Printf("Finished reading %d networks. ", vec.Vec_PtrSize(vPrs))
	fmt.Printf("NameIDs = %d. ", Abc_NamObjNumMax(Psr_ManNameMan(vPrs)))
	fmt.Printf("Memory = %.2f MB. ", float64(Psr_ManMemory(vPrs))/(1<<20))
	util.Abc_PrintTime(1, "Time", util.Abc_Clock()-clk)
	//    Abc_NamPrint( p->pStrs );
	Psr_ManWriteBlif("c/hie/dump/1/netlist_1_out8_out.blif", vPrs)
	Psr_ManVecFree(vPrs)
}

////////////////////////////////////////////////////////////////////////
///                       END OF FILE                                ///
////////////////////////////////////////////////////////////////////////

//ABC_NAMESPACE_IMPL_END
