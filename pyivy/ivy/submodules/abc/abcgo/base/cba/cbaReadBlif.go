package cba

import (
	"fmt"
	"strings"
)

/**CFile****************************************************************

  FileName    [cbaReadBlif.c]

  SystemName  [ABC: Logic synthesis and verification system.]

  PackageName [Hierarchical word-level netlist.]

  Synopsis    [BLIF parser.]

  Author      [Alan Mishchenko]

  Affiliation [UC Berkeley]

  Date        [Ver. 1.0. Started - November 29, 2014.]

  Revision    [$Id: cbaReadBlif.c,v 1.00 2014/11/29 00:00:00 alanmi Exp $]

***********************************************************************/

//#include "cba.h"
//#include "cbaPrs.h"

//ABC_NAMESPACE_IMPL_START

////////////////////////////////////////////////////////////////////////
///                        DECLARATIONS                              ///
////////////////////////////////////////////////////////////////////////

// BLIF keywords
type Cba_BlifType_t int32

const (
	PRS_BLIF_NONE Cba_BlifType_t = 0
	// 0:   unused
	PRS_BLIF_MODEL Cba_BlifType_t = PRS_BLIF_NONE + 1
	// 1:   .model
	PRS_BLIF_INOUTS Cba_BlifType_t = PRS_BLIF_MODEL + 1
	// 2:   .inouts
	PRS_BLIF_INPUTS Cba_BlifType_t = PRS_BLIF_INOUTS + 1
	// 3:   .inputs
	PRS_BLIF_OUTPUTS Cba_BlifType_t = PRS_BLIF_INPUTS + 1
	// 4:   .outputs
	PRS_BLIF_NAMES Cba_BlifType_t = PRS_BLIF_OUTPUTS + 1
	// 5:   .names
	PRS_BLIF_SUBCKT Cba_BlifType_t = PRS_BLIF_NAMES + 1
	// 6:   .subckt
	PRS_BLIF_GATE Cba_BlifType_t = PRS_BLIF_SUBCKT + 1
	// 7:   .gate
	PRS_BLIF_LATCH Cba_BlifType_t = PRS_BLIF_GATE + 1
	// 8:   .latch
	PRS_BLIF_SHORT Cba_BlifType_t = PRS_BLIF_LATCH + 1
	// 9:   .short
	PRS_BLIF_END Cba_BlifType_t = PRS_BLIF_SHORT + 1
	// 10:  .end
	PRS_BLIF_UNKNOWN Cba_BlifType_t = PRS_BLIF_END + 1 // 11:  unknown

)

var s_BlifTypes = []string{"", ".model", ".inouts", ".inputs", ".outputs", ".names", ".subckt", ".gate", ".latch", ".short", ".end", ""}

func Prs_NtkAddBlifDirectives(p *Prs_Man_t) {
	var i int32
	for i = 1; i < int32(len(s_BlifTypes)-1); i++ {
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
func Prs_ManIsSpace(p *Prs_Man_t) int32 { return prsBlifCharIsSpace(Prs_ManAt(p, p.pCur)) }
func Prs_ManIsStop(p *Prs_Man_t) int32  { return Prs_CharIsStop(Prs_ManAt(p, p.pCur)) }
func Prs_ManIsLit(p *Prs_Man_t) int32   { return Prs_CharIsLit(Prs_ManAt(p, p.pCur)) }

func prsBlifManIsChar(p *Prs_Man_t, c byte) int32 { return boolToInt(Prs_ManAt(p, p.pCur) == c) }
func Prs_ManIsChar2(p *Prs_Man_t, c byte) int32 {
	result := boolToInt(Prs_ManAt(p, p.pCur) == c)
	p.pCur++
	return result
}

func Prs_ManSkip(p *Prs_Man_t)       { p.pCur++ }
func Prs_ManSkip2(p *Prs_Man_t) byte { result := Prs_ManAt(p, p.pCur); p.pCur++; return result }

/*
*Function*************************************************************

	Synopsis    [Reading names.]

	Description []

	SideEffects []

	SeeAlso     []

**********************************************************************
*/
func Prs_ManSkipToChar(p *Prs_Man_t, c byte) {
	for !(prsBlifManIsChar(p, c) != 0) {
		Prs_ManSkip(p)
	}
}
func Prs_ManSkipSpaces(p *Prs_Man_t) {
	for {
		for Prs_ManIsSpace(p) != 0 {
			Prs_ManSkip(p)
		}
		if prsBlifManIsChar(p, '\\') != 0 {
			Prs_ManSkipToChar(p, '\n')
			Prs_ManSkip(p)
			continue
		}
		if prsBlifManIsChar(p, '#') != 0 {
			Prs_ManSkipToChar(p, '\n')
		}
		break
	}
	assert(!(Prs_ManIsSpace(p) != 0), "!Prs_ManIsSpace(p)")
}
func prsBlifManReadName(p *Prs_Man_t) int32 {
	var pStart int
	Prs_ManSkipSpaces(p)
	if prsBlifManIsChar(p, '\n') != 0 {
		return 0
	}
	pStart = p.pCur
	for !(Prs_ManIsSpace(p) != 0) && !(Prs_ManIsStop(p) != 0) {
		Prs_ManSkip(p)
	}
	if pStart == p.pCur {
		return 0
	}
	return Abc_NamStrFindOrAdd(p.pStrs, Prs_ManSlice(p, pStart, p.pCur), nil)
}
func Prs_ManReadList(p *Prs_Man_t, vOrder *Vec_Int_t, Type int32) int32 {
	var iToken int32
	Vec_IntClear(&p.vTemp)
	for {
		iToken = prsBlifManReadName(p)
		if iToken == 0 {
			break
		}
		Vec_IntPush(&p.vTemp, iToken)
		Vec_IntPush(vOrder, Abc_Var2Lit2(iToken, Type))
	}
	if Vec_IntSize(&p.vTemp) == 0 {
		return Prs_ManErrorSet(p, "Signal list is empty.", 1)
	}
	return 0
}
func Prs_ManReadList2(p *Prs_Man_t) int32 {
	var iToken int32
	Vec_IntClear(&p.vTemp)
	for {
		iToken = prsBlifManReadName(p)
		if iToken == 0 {
			break
		}
		Vec_IntPushTwo(&p.vTemp, 0, iToken)
	}
	if Vec_IntSize(&p.vTemp) == 0 {
		return Prs_ManErrorSet(p, "Signal list is empty.", 1)
	}
	return 0
}
func Prs_ManReadList3(p *Prs_Man_t) int32 {
	Vec_IntClear(&p.vTemp)
	for !(prsBlifManIsChar(p, '\n') != 0) {
		var iToken int32 = prsBlifManReadName(p)
		if iToken == 0 {
			return Prs_ManErrorSet(p, "Cannot read formal name.", 1)
		}
		Vec_IntPush(&p.vTemp, iToken)
		Prs_ManSkipSpaces(p)
		if !(Prs_ManIsChar2(p, '=') != 0) {
			return Prs_ManErrorSet(p, "Cannot find symbol \"=\".", 1)
		}
		iToken = prsBlifManReadName(p)
		if iToken == 0 {
			return Prs_ManErrorSet(p, "Cannot read actual name.", 1)
		}
		Vec_IntPush(&p.vTemp, iToken)
		Prs_ManSkipSpaces(p)
	}
	if Vec_IntSize(&p.vTemp) == 0 {
		return Prs_ManErrorSet(p, "Cannot read a list of formal/actual names.", 1)
	}
	if Vec_IntSize(&p.vTemp)%2 != 0 {
		return Prs_ManErrorSet(p, "The number of formal/actual names is not even.", 1)
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
func Prs_ManReadCube(p *Prs_Man_t) int32 {
	assert((Prs_ManIsLit(p) != 0), "Prs_ManIsLit(p)")
	for Prs_ManIsLit(p) != 0 {
		Vec_StrPush(&p.vCover, Prs_ManSkip2(p))
	}
	Prs_ManSkipSpaces(p)
	if prsBlifManIsChar(p, '\n') != 0 {
		if Vec_StrSize(&p.vCover) != 1 {
			return Prs_ManErrorSet(p, "Cannot read cube.", 1)
		}
		// fix single literal cube by adding space
		Vec_StrPush(&p.vCover, Vec_StrEntry(&p.vCover, 0))
		Vec_StrWriteEntry(&p.vCover, 0, ' ')
		Vec_StrPush(&p.vCover, '\n')
		return 0
	}
	if !(Prs_ManIsLit(p) != 0) {
		return Prs_ManErrorSet(p, "Cannot read output literal.", 1)
	}
	Vec_StrPush(&p.vCover, ' ')
	Vec_StrPush(&p.vCover, Prs_ManSkip2(p))
	Vec_StrPush(&p.vCover, '\n')
	Prs_ManSkipSpaces(p)
	if !(prsBlifManIsChar(p, '\n') != 0) {
		return Prs_ManErrorSet(p, "Cannot read end of cube.", 1)
	}
	return 0
}
func Prs_ManSaveCover(p *Prs_Man_t) {
	var iToken int32
	if Vec_StrSize(&p.vCover) == 0 {
		p.pNtk.fHasC0s = 1
	} else if Vec_StrSize(&p.vCover) == 2 {
		if Vec_StrEntry(&p.vCover, Vec_StrSize(&p.vCover)-1) == '0' {
			p.pNtk.fHasC0s = 1
		} else if Vec_StrEntry(&p.vCover, Vec_StrSize(&p.vCover)-1) == '1' {
			p.pNtk.fHasC1s = 1
		} else {
			assert(false, "0")
		}
	}
	assert(Vec_StrSize(&p.vCover) > 0, "Vec_StrSize(&p.vCover) > 0")
	Vec_StrPush(&p.vCover, '\x00')
	//    iToken = Ptr_SopToType( Vec_StrArray(&p->vCover) );
	iToken = Abc_NamStrFindOrAdd(p.pFuns, strings.TrimRight(string(Vec_StrArray(&p.vCover)), "\x00"), nil)
	Vec_StrClear(&p.vCover)
	// set the cover to the module of this box
	assert(Prs_BoxNtk(p.pNtk, Prs_NtkBoxNum(p.pNtk)-1) == 1, "Prs_BoxNtk(p.pNtk, Prs_NtkBoxNum(p.pNtk)-1) == 1") // default const 0
	Prs_BoxSetNtk(p.pNtk, Prs_NtkBoxNum(p.pNtk)-1, iToken)
}

/*
*Function*************************************************************

	Synopsis    []

	Description []

	SideEffects []

	SeeAlso     []

**********************************************************************
*/
func Prs_ManReadInouts(p *Prs_Man_t) int32 {
	if Prs_ManReadList(p, &p.pNtk.vOrder, 3) != 0 {
		return 1
	}
	Vec_IntAppend(&p.pNtk.vInouts, &p.vTemp)
	return 0
}
func Prs_ManReadInputs(p *Prs_Man_t) int32 {
	if Prs_ManReadList(p, &p.pNtk.vOrder, 1) != 0 {
		return 1
	}
	Vec_IntAppend(&p.pNtk.vInputs, &p.vTemp)
	return 0
}
func Prs_ManReadOutputs(p *Prs_Man_t) int32 {
	if Prs_ManReadList(p, &p.pNtk.vOrder, 2) != 0 {
		return 1
	}
	Vec_IntAppend(&p.pNtk.vOutputs, &p.vTemp)
	return 0
}
func Prs_ManReadNode(p *Prs_Man_t) int32 {
	if Prs_ManReadList2(p) != 0 {
		return 1
	}
	// save results
	Prs_NtkAddBox(p.pNtk, 1, 0, &p.vTemp) // default const 0 function
	return 0
}
func Prs_ManReadBox(p *Prs_Man_t, fGate int32) int32 {
	var iToken int32 = prsBlifManReadName(p)
	if iToken == 0 {
		return Prs_ManErrorSet(p, "Cannot read model name.", 1)
	}
	if Prs_ManReadList3(p) != 0 {
		return 1
	}
	// save results
	Prs_NtkAddBox(p.pNtk, iToken, 0, &p.vTemp)
	if fGate != 0 {
		p.pNtk.fMapped = 1
	}
	return 0
}
func Prs_ManReadLatch(p *Prs_Man_t) int32 {
	var iToken int32 = prsBlifManReadName(p)
	Vec_IntClear(&p.vTemp)
	if iToken == 0 {
		return Prs_ManErrorSet(p, "Cannot read latch input.", 1)
	}
	Vec_IntWriteEntry(&p.vTemp, 1, iToken)
	iToken = prsBlifManReadName(p)
	if iToken == 0 {
		return Prs_ManErrorSet(p, "Cannot read latch output.", 1)
	}
	Vec_IntWriteEntry(&p.vTemp, 0, iToken)
	Prs_ManSkipSpaces(p)
	if prsBlifManIsChar(p, '0') != 0 {
		iToken = 0
	} else if prsBlifManIsChar(p, '1') != 0 {
		iToken = 1
	} else {
		iToken = 2
	}
	Prs_ManSkipToChar(p, '\n')
	// save results
	Prs_NtkAddBox(p.pNtk, -1, iToken, &p.vTemp) // -1 stands for latch
	return 0
}
func Prs_ManReadShort(p *Prs_Man_t) int32 {
	var iToken int32 = prsBlifManReadName(p)
	Vec_IntClear(&p.vTemp)
	if iToken == 0 {
		return Prs_ManErrorSet(p, "Cannot read .short input.", 1)
	}
	Vec_IntWriteEntry(&p.vTemp, 1, iToken)
	iToken = prsBlifManReadName(p)
	if iToken == 0 {
		return Prs_ManErrorSet(p, "Cannot read .short output.", 1)
	}
	Vec_IntWriteEntry(&p.vTemp, 0, iToken)
	Prs_ManSkipSpaces(p)
	if !(prsBlifManIsChar(p, '\n') != 0) {
		return Prs_ManErrorSet(p, "Trailing symbols on .short line.", 1)
	}
	// save results
	iToken = Abc_NamStrFindOrAdd(p.pStrs, "1 1\n", nil)
	Prs_NtkAddBox(p.pNtk, iToken, 0, &p.vTemp)
	return 0
}
func Prs_ManReadModel(p *Prs_Man_t) int32 {
	var iToken int32
	if p.pNtk != nil {
		return Prs_ManErrorSet(p, "Parsing previous model is unfinished.", 1)
	}
	iToken = prsBlifManReadName(p)
	if iToken == 0 {
		return Prs_ManErrorSet(p, "Cannot read model name.", 1)
	}
	Prs_ManInitializeNtk(p, iToken, 0)
	Prs_ManSkipSpaces(p)
	if !(prsBlifManIsChar(p, '\n') != 0) {
		return Prs_ManErrorSet(p, "Trailing symbols on .model line.", 1)
	}
	return 0
}
func Prs_ManReadEnd(p *Prs_Man_t) int32 {
	if p.pNtk == nil {
		return Prs_ManErrorSet(p, "Directive .end without .model.", 1)
	}
	//printf( "Saving model \"%s\".\n", Abc_NamStr(p->pStrs, p->iModuleName) );
	Prs_ManFinalizeNtk(p)
	Prs_ManSkipSpaces(p)
	if !(prsBlifManIsChar(p, '\n') != 0) {
		return Prs_ManErrorSet(p, "Trailing symbols on .end line.", 1)
	}
	return 0
}

func Prs_ManReadDirective(p *Prs_Man_t) int32 {
	var iToken int32
	if !(prsBlifManIsChar(p, '.') != 0) {
		return Prs_ManReadCube(p)
	}
	if Vec_StrSize(&p.vCover) > 0 { // SOP was specified for the previous node
		Prs_ManSaveCover(p)
	}
	iToken = prsBlifManReadName(p)
	if iToken == int32(PRS_BLIF_MODEL) {
		return Prs_ManReadModel(p)
	}
	if iToken == int32(PRS_BLIF_INOUTS) {
		return Prs_ManReadInouts(p)
	}
	if iToken == int32(PRS_BLIF_INPUTS) {
		return Prs_ManReadInputs(p)
	}
	if iToken == int32(PRS_BLIF_OUTPUTS) {
		return Prs_ManReadOutputs(p)
	}
	if iToken == int32(PRS_BLIF_NAMES) {
		return Prs_ManReadNode(p)
	}
	if iToken == int32(PRS_BLIF_SUBCKT) {
		return Prs_ManReadBox(p, 0)
	}
	if iToken == int32(PRS_BLIF_GATE) {
		return Prs_ManReadBox(p, 1)
	}
	if iToken == int32(PRS_BLIF_LATCH) {
		return Prs_ManReadLatch(p)
	}
	if iToken == int32(PRS_BLIF_SHORT) {
		return Prs_ManReadShort(p)
	}
	if iToken == int32(PRS_BLIF_END) {
		return Prs_ManReadEnd(p)
	}
	fmt.Printf("Cannot read directive \"%s\".\n", Abc_NamStr(p.pStrs, iToken))
	return 1
}
func Prs_ManReadLines(p *Prs_Man_t) int32 {
	for Prs_ManPeek(p, 1) != '\x00' {
		assert((prsBlifManIsChar(p, '\n') != 0), "prsBlifManIsChar(p, '\\n')")
		Prs_ManSkip(p)
		Prs_ManSkipSpaces(p)
		if prsBlifManIsChar(p, '\n') != 0 {
			continue
		}
		if Prs_ManReadDirective(p) != 0 {
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
func Prs_ManReadBlif(pFileName string) *Vec_Ptr_t {
	var vPrs *Vec_Ptr_t = nil
	var p *Prs_Man_t = Prs_ManAlloc(pFileName)
	if p == nil {
		return nil
	}
	Abc_NamStrFindOrAdd(p.pFuns, " 0\n", nil)
	Abc_NamStrFindOrAdd(p.pFuns, " 1\n", nil)
	Prs_NtkAddBlifDirectives(p)
	Prs_ManReadLines(p)
	if Prs_ManErrorPrint(p) != 0 {
		// inlining of macro ABC_SWAP
		var t *Vec_Ptr_t = vPrs
		vPrs = p.vNtks
		p.vNtks = t
	}
	Prs_ManFree(p)
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
func Prs_ManReadBlifTest(pFileName string) {
	var clk abctime = Abc_Clock()
	var vPrs *Vec_Ptr_t = Prs_ManReadBlif(pFileName)
	if !(vPrs != nil) {
		return
	}
	fmt.Printf("Finished reading %d networks. ", Vec_PtrSize(vPrs))
	fmt.Printf("NameIDs = %d. ", Abc_NamObjNumMax(Prs_ManNameMan(vPrs)))
	fmt.Printf("Memory = %.2f MB. ", float64(Prs_ManMemory(vPrs))/float64(1<<20))
	Abc_PrintTime(1, "Time", Abc_Clock()-clk)
	//    Abc_NamPrint( p->pStrs );
	Prs_ManWriteBlif(Extra_FileNameGenericAppend(pFileName, "_out.blif"), vPrs)
	Prs_ManVecFree(vPrs)
}

/*
*Function*************************************************************

	Synopsis    []

	Description []

	SideEffects []

	SeeAlso     []

**********************************************************************
*/
func Prs_CreateBlifFindFon(p *Cba_Ntk_t, NameId int32) int32 {
	var iFon int32 = Cba_NtkGetMap(p, NameId)
	if iFon != 0 {
		return iFon
	}
	fmt.Printf("Network \"%s\": Signal \"%s\" is not driven.\n", Cba_NtkName(p), Cba_NtkStr(p, NameId))
	return 0
}
func Prs_CreateBlifPio(p *Cba_Ntk_t, pNtk *Prs_Ntk_t) {
	var i, NameId, iObj, iFon int32
	Cba_NtkCleanObjFuncs(p)
	Cba_NtkCleanObjNames(p)
	Cba_NtkCleanFonNames(p)
	// create inputs
	// inlining of macro Prs_NtkForEachPi
	for i = 0; i < Prs_NtkPiNum(pNtk); i++ {
		NameId = Vec_IntEntry(&pNtk.vInputs, i)
		if NameId == 0 {
			break
		}
		iObj = Cba_ObjAlloc(p, CBA_OBJ_PI, 0, 1)
		Cba_ObjSetName(p, iObj, NameId)
		iFon = Cba_ObjFon0(p, iObj)
		Cba_FonSetName(p, iFon, NameId)
		Cba_NtkSetMap(p, NameId, iFon)
		Vec_IntPush(&p.vOrder, iObj)
	}
	// create outputs
	// inlining of macro Prs_NtkForEachPo
	for i = 0; i < Prs_NtkPoNum(pNtk); i++ {
		NameId = Vec_IntEntry(&pNtk.vOutputs, i)
		if NameId == 0 {
			break
		}
		iObj = Cba_ObjAlloc(p, CBA_OBJ_PO, 1, 0)
		Cba_ObjSetName(p, iObj, NameId)
		Vec_IntPush(&p.vOrder, iObj)
	}
}
func Prs_CreateBlifNtk(p *Cba_Ntk_t, pNtk *Prs_Ntk_t) int32 {
	var vBox *Vec_Int_t
	var i, k, iObj, iTerm, iFon, FormId, ActId int32
	// map inputs
	Cba_NtkCleanMap(p)
	// inlining of macro Cba_NtkForEachPi
	for i = 0; i < Cba_NtkPiNum(p); i++ {
		iObj = Cba_NtkPi(p, i)
		Cba_NtkSetMap(p, Cba_ObjName(p, iObj), Cba_ObjFon0(p, iObj))
	}
	// create objects
	// inlining of macro Prs_NtkForEachBox
	for i = 0; i < Prs_NtkBoxNum(pNtk); i++ {
		vBox = Prs_BoxSignals(pNtk, i)
		var FuncId int32 = Prs_BoxNtk(pNtk, i)
		assert(Prs_BoxIONum(pNtk, i) > 0, "Prs_BoxIONum(pNtk, i) > 0")
		assert(Vec_IntSize(vBox)%2 == 0, "Vec_IntSize(vBox) % 2 == 0")
		if FuncId == -1 { // latch

			iObj = Cba_ObjAlloc(p, CBA_BOX_DFFRS, 4, 1)
			Cba_NtkSetMap(p, Vec_IntEntry(vBox, 3), Cba_ObjFon0(p, iObj)) // latch output
			Cba_ObjSetFunc(p, iObj, Prs_BoxName(pNtk, i)+1)               // init + 1

		} else if Prs_BoxIsNode(pNtk, i) != 0 { // node

			iObj = Cba_ObjAlloc(p, CBA_BOX_NODE, Prs_BoxIONum(pNtk, i)-1, 1)
			Cba_FonSetName(p, Cba_ObjFon0(p, iObj), Vec_IntEntryLast(vBox)) // node output
			Cba_NtkSetMap(p, Vec_IntEntryLast(vBox), Cba_ObjFon0(p, iObj))
			Cba_ObjSetFunc(p, iObj, FuncId)

		} else { // box

			var pBox *Cba_Ntk_t = Cba_ManNtkFind(p.pDesign, Prs_NtkStr(pNtk, FuncId))
			iObj = Cba_ObjAlloc(p, CBA_OBJ_BOX, Cba_NtkPiNum(pBox), Cba_NtkPoNum(pBox))
			Cba_ObjSetFunc(p, iObj, Cba_NtkId(pBox))
			// mark PO objects
			Cba_NtkCleanMap2(p)
			// inlining of macro Cba_NtkForEachPo
			for k = 0; k < Cba_NtkPoNum(pBox); k++ {
				iTerm = Cba_NtkPo(pBox, k)
				Cba_NtkSetMap2(p, Cba_ObjName(pBox, iTerm), k+1)
			}
			// map box fons
			// inlining of macro Vec_IntForEachEntryDouble
			for k = 0; k+1 < Vec_IntSize(vBox); k += 2 {
				FormId = Vec_IntEntry(vBox, k)
				ActId = Vec_IntEntry(vBox, k+1)
				if Cba_NtkGetMap2(p, FormId) != 0 {
					iFon = Cba_ObjFon(p, iObj, Cba_NtkGetMap2(p, FormId)-1)
					Cba_FonSetName(p, iFon, ActId)
					Cba_NtkSetMap(p, ActId, iFon)
				}
			}
		}
	}
	// connect objects
	// inlining of macro Prs_NtkForEachBox
	for i = 0; i < Prs_NtkBoxNum(pNtk); i++ {
		vBox = Prs_BoxSignals(pNtk, i)
		iObj = Cba_NtkPiNum(p) + Cba_NtkPoNum(p) + i + 1
		if Prs_BoxNtk(pNtk, i) == -1 { // latch

			assert(Cba_ObjType(p, iObj) == CBA_BOX_DFFRS, "Cba_ObjType(p, iObj) == CBA_BOX_DFFRS")
			iFon = Prs_CreateBlifFindFon(p, Vec_IntEntry(vBox, 1)) // latch input
			if iFon != 0 {
				Cba_ObjSetFinFon(p, iObj, 0, iFon)
			}

		} else if Prs_BoxIsNode(pNtk, i) != 0 { // node

			assert(Cba_ObjType(p, iObj) == CBA_BOX_NODE, "Cba_ObjType(p, iObj) == CBA_BOX_NODE")
			// inlining of macro Vec_IntForEachEntryDouble
			for k = 0; k+1 < Vec_IntSize(vBox); k += 2 {
				FormId = Vec_IntEntry(vBox, k)
				ActId = Vec_IntEntry(vBox, k+1)
				if k == 2*Cba_ObjFinNum(p, iObj) { // stop at node output
					break
				}
				iFon = Prs_CreateBlifFindFon(p, ActId)
				if iFon != 0 {
					Cba_ObjSetFinFon(p, iObj, k/2, iFon)
				}
			}

		} else { // box

			// mark PI objects
			var pBox *Cba_Ntk_t = Cba_ObjNtk(p, iObj)
			assert(Cba_NtkPiNum(pBox) == Cba_ObjFinNum(p, iObj), "Cba_NtkPiNum(pBox) == Cba_ObjFinNum(p, iObj)")
			assert(Cba_NtkPoNum(pBox) == Cba_ObjFonNum(p, iObj), "Cba_NtkPoNum(pBox) == Cba_ObjFonNum(p, iObj)")
			Cba_NtkCleanMap2(p)
			// inlining of macro Cba_NtkForEachPi
			for k = 0; k < Cba_NtkPiNum(pBox); k++ {
				iTerm = Cba_NtkPi(pBox, k)
				Cba_NtkSetMap2(p, Cba_ObjName(pBox, iTerm), k+1)
			}
			// connect box fins
			// inlining of macro Vec_IntForEachEntryDouble
			for k = 0; k+1 < Vec_IntSize(vBox); k += 2 {
				FormId = Vec_IntEntry(vBox, k)
				ActId = Vec_IntEntry(vBox, k+1)
				if Cba_NtkGetMap2(p, FormId) != 0 {
					iFon = Prs_CreateBlifFindFon(p, ActId)
					if iFon != 0 {
						Cba_ObjSetFinFon(p, iObj, Cba_NtkGetMap2(p, FormId)-1, iFon)
					}
				}
			}
		}
	}
	// connect outputs
	// inlining of macro Cba_NtkForEachPo
	for i = 0; i < Cba_NtkPoNum(p); i++ {
		iObj = Cba_NtkPo(p, i)
		iFon = Prs_CreateBlifFindFon(p, Cba_ObjName(p, iObj))
		if iFon != 0 {
			Cba_ObjSetFinFon(p, iObj, 0, iFon)
		}
	}
	return 0
}
func Prs_ManBuildCbaBlif(pFileName string, vDes *Vec_Ptr_t) *Cba_Man_t {
	var pPrsNtk *Prs_Ntk_t
	var i, fError int32
	var pPrsRoot *Prs_Ntk_t = Prs_ManRoot(vDes)
	// start the manager
	var pStrs *Abc_Nam_t = Abc_NamRef(pPrsRoot.pStrs)
	var pFuns *Abc_Nam_t = Abc_NamRef(pPrsRoot.pFuns)
	var pMods *Abc_Nam_t = Abc_NamStart(100, 24)
	var p *Cba_Man_t = Cba_ManAlloc(pFileName, Vec_PtrSize(vDes), pStrs, pFuns, pMods, nil)
	// initialize networks
	// inlining of macro Vec_PtrForEachEntry
	for i = 0; i < Vec_PtrSize(vDes); i++ {
		pPrsNtk, _ = Vec_PtrEntry(vDes, i).(*Prs_Ntk_t)
		var pNtk *Cba_Ntk_t = Cba_NtkAlloc(p, Prs_NtkId(pPrsNtk), Prs_NtkPiNum(pPrsNtk), Prs_NtkPoNum(pPrsNtk), Prs_NtkObjNum(pPrsNtk), 100, 100)
		Prs_CreateBlifPio(pNtk, pPrsNtk)
		Cba_NtkAdd(p, pNtk)
	}
	// create networks
	// inlining of macro Vec_PtrForEachEntry
	for i = 0; i < Vec_PtrSize(vDes); i++ {
		pPrsNtk, _ = Vec_PtrEntry(vDes, i).(*Prs_Ntk_t)
		fmt.Printf("Elaboration module \"%s\"...\n", Prs_NtkName(pPrsNtk))
		fError = Prs_CreateBlifNtk(Cba_ManNtk(p, i+1), pPrsNtk)
		if fError != 0 {
			break
		}
	}
	if fError != 0 {
		fmt.Printf("Quitting because of errors.\n")
	} else {
		Cba_ManPrepareSeq(p)
	}
	return p
}

/*
*Function*************************************************************

	Synopsis    []

	Description []

	SideEffects []

	SeeAlso     []

**********************************************************************
*/
func Cba_ManReadBlif(pFileName string) *Cba_Man_t {
	var p *Cba_Man_t = nil
	var vDes *Vec_Ptr_t = Prs_ManReadBlif(pFileName)
	if (vDes != nil) && Vec_PtrSize(vDes) != 0 {
		p = Prs_ManBuildCbaBlif(pFileName, vDes)
	}
	if vDes != nil {
		Prs_ManVecFree(vDes)
	}
	return p
}

////////////////////////////////////////////////////////////////////////
///                       END OF FILE                                ///
////////////////////////////////////////////////////////////////////////

//ABC_NAMESPACE_IMPL_END
