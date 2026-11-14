package acb

import (
	"fmt"
	baseabc "github.com/glycerine/abcgo/base/abc"
	basemain "github.com/glycerine/abcgo/base/main"
	"github.com/glycerine/abcgo/misc/extra"
	"github.com/glycerine/abcgo/misc/util"
	"github.com/glycerine/abcgo/misc/vec"
	"strconv"
	"strings"
	"sync"
	"time"
)

import "os"

/**CFile****************************************************************

  FileName    [abcFunc.c]

  SystemName  [ABC: Logic synthesis and verification system.]

  PackageName [Network and node package.]

  Synopsis    [Experimental procedures.]

  Author      [Alan Mishchenko]

  Affiliation [UC Berkeley]

  Date        [Ver. 1.0. Started - June 20, 2005.]

  Revision    [$Id: abcFunc.c,v 1.00 2005/06/20 00:00:00 alanmi Exp $]

***********************************************************************/

//#include "acb.h"
//#include "aig/miniaig/ndr.h"
//#include "sat/cnf/cnf.h"
//#include "sat/bsat/satStore.h"
//#include "sat/satoko/satoko.h"
//#include "map/mio/mio.h"
//#include "misc/util/utilTruth.h"
//#include "aig/gia/giaAig.h"
//#include "base/main/main.h"
//#include "base/cmd/cmd.h"

//ABC_NAMESPACE_IMPL_START

////////////////////////////////////////////////////////////////////////
///                        DECLARATIONS                              ///
////////////////////////////////////////////////////////////////////////

type Acb_KeyWords_t = int32

const (
	ACB_NONE Acb_KeyWords_t = 0
	// 0:  unused
	ACB_MODULE Acb_KeyWords_t = ACB_NONE + 1
	// 1:  "module"
	ACB_ENDMODULE Acb_KeyWords_t = ACB_MODULE + 1
	// 2:  "endmodule"
	ACB_INPUT Acb_KeyWords_t = ACB_ENDMODULE + 1
	// 3:  "input"
	ACB_OUTPUT Acb_KeyWords_t = ACB_INPUT + 1
	// 4:  "output"
	ACB_WIRE Acb_KeyWords_t = ACB_OUTPUT + 1
	// 5:  "wire"
	ACB_BUF Acb_KeyWords_t = ACB_WIRE + 1
	// 6:  "buf"
	ACB_NOT Acb_KeyWords_t = ACB_BUF + 1
	// 7:  "not"
	ACB_AND Acb_KeyWords_t = ACB_NOT + 1
	// 8:  "and"
	ACB_NAND Acb_KeyWords_t = ACB_AND + 1
	// 9:  "nand"
	ACB_OR Acb_KeyWords_t = ACB_NAND + 1
	// 10: "or"
	ACB_NOR Acb_KeyWords_t = ACB_OR + 1
	// 11: "nor"
	ACB_XOR Acb_KeyWords_t = ACB_NOR + 1
	// 12: "xor"
	ACB_XNOR Acb_KeyWords_t = ACB_XOR + 1
	// 13: "xnor"
	ACB_MUX Acb_KeyWords_t = ACB_XNOR + 1
	// 14: "_HMUX"
	ACB_DC Acb_KeyWords_t = ACB_MUX + 1
	// 15: "_DC"
	ACB_UNUSED Acb_KeyWords_t = ACB_DC + 1 // 14: unused

)

func Acb_Num2Name(i int32) string {
	if i == 1 {
		return "module"
	}
	if i == 2 {
		return "endmodule"
	}
	if i == 3 {
		return "input"
	}
	if i == 4 {
		return "output"
	}
	if i == 5 {
		return "wire"
	}
	if i == 6 {
		return "buf"
	}
	if i == 7 {
		return "not"
	}
	if i == 8 {
		return "and"
	}
	if i == 9 {
		return "nand"
	}
	if i == 10 {
		return "or"
	}
	if i == 11 {
		return "nor"
	}
	if i == 12 {
		return "xor"
	}
	if i == 13 {
		return "xnor"
	}
	if i == 14 {
		return "_HMUX"
	}
	if i == 15 {
		return "_DC"
	}
	return ""
}

func Acb_Type2Oper(i int32) int32 {
	if i == 6 {
		return ABC_OPER_BIT_BUF
	}
	if i == 7 {
		return ABC_OPER_BIT_INV
	}
	if i == 8 {
		return ABC_OPER_BIT_AND
	}
	if i == 9 {
		return ABC_OPER_BIT_NAND
	}
	if i == 10 {
		return ABC_OPER_BIT_OR
	}
	if i == 11 {
		return ABC_OPER_BIT_NOR
	}
	if i == 12 {
		return ABC_OPER_BIT_XOR
	}
	if i == 13 {
		return ABC_OPER_BIT_NXOR
	}
	if i == 14 {
		return ABC_OPER_BIT_MUX
	}
	if i == 15 {
		return ABC_OPER_TRI
	}
	assert(false, "0")
	return -1
}

func Acb_Oper2Name(i int32) string {
	if i == ABC_OPER_CONST_F {
		return "const0"
	}
	if i == ABC_OPER_CONST_T {
		return "const1"
	}
	if i == ABC_OPER_CONST_X {
		return "constX"
	}
	if i == ABC_OPER_BIT_BUF {
		return "buf"
	}
	if i == ABC_OPER_BIT_INV {
		return "not"
	}
	if i == ABC_OPER_BIT_AND {
		return "and"
	}
	if i == ABC_OPER_BIT_NAND {
		return "nand"
	}
	if i == ABC_OPER_BIT_OR {
		return "or"
	}
	if i == ABC_OPER_BIT_NOR {
		return "nor"
	}
	if i == ABC_OPER_BIT_XOR {
		return "xor"
	}
	if i == ABC_OPER_BIT_NXOR {
		return "xnor"
	}
	if i == ABC_OPER_BIT_MUX {
		return "mux"
	}
	if i == ABC_OPER_TRI {
		return "_DC"
	}
	assert(false, "0")
	return ""
}

////////////////////////////////////////////////////////////////////////
///                     FUNCTION DEFINITIONS                         ///
////////////////////////////////////////////////////////////////////////

/*
*Function*************************************************************

	Synopsis    [Installs the required standard cell library.]

	Description []

	SideEffects []

	SeeAlso     []

**********************************************************************
*/
var pLibStr [25]string = [25]string{
	"GATE buf        1       O=a;            PIN * INV 1 999 1.0 0.0 1.0 0.0\n" +
		"GATE inv        1       O=!a;           PIN * INV 1 999 1.0 0.0 1.0 0.0\n" +
		"GATE and2       1       O=a*b;          PIN * INV 1 999 1.0 0.0 1.0 0.0\n" +
		"GATE and3       1       O=a*b*c;        PIN * INV 1 999 1.0 0.0 1.0 0.0\n" +
		"GATE and4       1       O=a*b*c*d;      PIN * INV 1 999 1.0 0.0 1.0 0.0\n" +
		"GATE or2        1       O=a+b;          PIN * INV 1 999 1.0 0.0 1.0 0.0\n" +
		"GATE or3        1       O=a+b+c;        PIN * INV 1 999 1.0 0.0 1.0 0.0\n" +
		"GATE or4        1       O=a+b+c+d;      PIN * INV 1 999 1.0 0.0 1.0 0.0\n" +
		"GATE nand2      1       O=!(a*b);       PIN * INV 1 999 1.0 0.0 1.0 0.0\n" +
		"GATE nand3      1       O=!(a*b*c);     PIN * INV 1 999 1.0 0.0 1.0 0.0\n" +
		"GATE nand4      1       O=!(a*b*c*d);   PIN * INV 1 999 1.0 0.0 1.0 0.0\n" +
		"GATE nor2       1       O=!(a+b);       PIN * INV 1 999 1.0 0.0 1.0 0.0\n" +
		"GATE nor3       1       O=!(a+b+c);     PIN * INV 1 999 1.0 0.0 1.0 0.0\n" +
		"GATE nor4       1       O=!(a+b+c+d);   PIN * INV 1 999 1.0 0.0 1.0 0.0\n" +
		"GATE xor        1       O=!a*b+a*!b;    PIN * INV 1 999 1.0 0.0 1.0 0.0\n" +
		"GATE xnor       1       O=a*b+!a*!b;    PIN * INV 1 999 1.0 0.0 1.0 0.0\n" +
		"GATE zero       0       O=CONST0;\n" +
		"GATE one        0       O=CONST1;\n",
}
var pLibStr2 [25]string = [25]string{
	"GATE buf        1       O=a;            PIN * INV 1 999 1.0 0.0 1.0 0.0\n" +
		"GATE inv        1       O=!a;           PIN * INV 1 999 1.0 0.0 1.0 0.0\n" +
		"GATE and2       1       O=a*b;          PIN * INV 1 999 1.0 0.0 1.0 0.0\n" +
		"GATE or2        1       O=a+b;          PIN * INV 1 999 1.0 0.0 1.0 0.0\n" +
		"GATE nand2      1       O=!(a*b);       PIN * INV 1 999 1.0 0.0 1.0 0.0\n" +
		"GATE nor2       1       O=!(a+b);       PIN * INV 1 999 1.0 0.0 1.0 0.0\n" +
		"GATE xor        1       O=!a*b+a*!b;    PIN * INV 1 999 1.0 0.0 1.0 0.0\n" +
		"GATE xnor       1       O=a*b+!a*!b;    PIN * INV 1 999 1.0 0.0 1.0 0.0\n" +
		"GATE zero       0       O=CONST0;\n" +
		"GATE one        0       O=CONST1;\n",
}

func Acb_IntallLibrary(f2Ins int32) {

	var pLib *Mio_Library_t
	var i int32
	// create library string
	var vLibStr *Vec_Str_t = Vec_StrAlloc(1000)
	var ppLibStr []string
	if f2Ins != 0 {
		ppLibStr = pLibStr2[:]
	} else {
		ppLibStr = pLibStr[:]
	}
	for i = 0; int(i) < len(ppLibStr) && ppLibStr[i] != ""; i++ {
		Vec_StrAppend(vLibStr, ppLibStr[i])
	}
	Vec_StrPush(vLibStr, 0)
	// create library
	pLib = Mio_LibraryReadBuffer(string(Vec_StrArray(vLibStr)[:Vec_StrSize(vLibStr)-1]), 0, nil, 0)
	Mio_LibrarySetName(pLib, Abc_UtilStrsav("iccad17.genlib"))
	Mio_UpdateGenlib(pLib)
	Vec_StrFree(vLibStr)
}

/*
*Function*************************************************************

	Synopsis    [Parse Verilog file into an intermediate representation.]

	Description []

	SideEffects []

	SeeAlso     []

**********************************************************************
*/
func Acb_VerilogStartNames() *Abc_Nam_t {
	var pNames *Abc_Nam_t = Abc_NamStart(100, 16)
	var i, NameId, fFound int32
	for i = 1; i < ACB_UNUSED; i++ {
		NameId = Abc_NamStrFindOrAdd(pNames, Acb_Num2Name(i), &fFound)
		assert(i == NameId && !(fFound != 0), "i == NameId && !fFound")
	}
	return pNames
}
func Acb_VerilogRemoveComments(buffer []byte) {
	acbRemoveComments(buffer)
}
func Acb_VerilogSimpleLex(pFileName string, pNames *Abc_Nam_t) *Vec_Int_t {
	buffer, err := os.ReadFile(pFileName)
	if err != nil {
		return nil
	}
	result := vec.Vec_IntAlloc(1000)
	tokens := acbLexBuffer(buffer, func(token string) int32 {
		return util.Abc_NamStrFindOrAdd(pNames, token, nil)
	})
	for _, token := range tokens {
		vec.Vec_IntPush(result, token)
	}
	return result
}
func Acb_WireIsTarget(Token int32, pNames *Abc_Nam_t) int32 {
	var pStr string = Abc_NamStr(pNames, Token)
	if len(pStr) >= 2 && pStr[0] == 't' && pStr[1] == '_' {
		return 1
	}
	return 0
}
func Acb_VerilogSimpleParse(vBuffer *Vec_Int_t, pNames *Abc_Nam_t) any {
	var pDesign *Ndr_Data_t = nil
	var vInputs *Vec_Int_t = Vec_IntAlloc(100)
	var vOutputs *Vec_Int_t = Vec_IntAlloc(100)
	var vWires *Vec_Int_t = Vec_IntAlloc(100)
	var vTypes *Vec_Int_t = Vec_IntAlloc(100)
	var vFanins *Vec_Int_t = Vec_IntAlloc(100)
	var vCur *Vec_Int_t = nil
	var i, ModuleID, Token, Size, Count int32
	assert(Vec_IntEntry(vBuffer, 0) == ACB_MODULE, "Vec_IntEntry(vBuffer, 0) == ACB_MODULE")
	// Vec_IntForEachEntryStart(vBuffer, Token, i, 2)
	for i = 2; i < Vec_IntSize(vBuffer); i++ {
		Token = Vec_IntEntry(vBuffer, i)
		if vCur == nil && Token >= ACB_UNUSED {
			continue
		}
		if Token == ACB_ENDMODULE {
			break
		}
		if Token == ACB_INPUT {
			vCur = vInputs
		} else if Token == ACB_OUTPUT {
			vCur = vOutputs
		} else if Token == ACB_WIRE {
			vCur = vWires
		} else if Token >= ACB_BUF && Token < ACB_UNUSED {
			Vec_IntPush(vTypes, Token)
			Vec_IntPush(vTypes, Vec_IntSize(vFanins))
			vCur = vFanins

		} else {
			Vec_IntPush(vCur, Token)
		}
	}
	Vec_IntPush(vTypes, -1)
	Vec_IntPush(vTypes, Vec_IntSize(vFanins))
	// create design
	pDesign = Ndr_Create(Vec_IntEntry(vBuffer, 1))
	ModuleID = Ndr_AddModule(pDesign, Vec_IntEntry(vBuffer, 1))
	// create inputs
	Ndr_DataResize(pDesign, Vec_IntSize(vInputs))
	for i = 0; i < Vec_IntSize(vInputs); i++ {
		Token = Vec_IntEntry(vInputs, i)
		Ndr_DataPush(pDesign, NDR_INPUT, Token)
	}
	Ndr_DataAddTo(pDesign, ModuleID-256, Vec_IntSize(vInputs))
	Ndr_DataAddTo(pDesign, 0, Vec_IntSize(vInputs))
	// create outputs
	Ndr_DataResize(pDesign, Vec_IntSize(vOutputs))
	for i = 0; i < Vec_IntSize(vOutputs); i++ {
		Token = Vec_IntEntry(vOutputs, i)
		Ndr_DataPush(pDesign, NDR_OUTPUT, Token)
	}
	Ndr_DataAddTo(pDesign, ModuleID-256, Vec_IntSize(vOutputs))
	Ndr_DataAddTo(pDesign, 0, Vec_IntSize(vOutputs))
	// create targets
	Ndr_DataResize(pDesign, Vec_IntSize(vWires))
	for i = 0; i < Vec_IntSize(vWires); i++ {
		Token = Vec_IntEntry(vWires, i)
		if Acb_WireIsTarget(Token, pNames) != 0 {
			Ndr_DataPush(pDesign, NDR_TARGET, Token)
			Count++
		}
	}
	Ndr_DataAddTo(pDesign, ModuleID-256, Count)
	Ndr_DataAddTo(pDesign, 0, Count)
	// create nodes
	for i = 0; i < Vec_IntSize(vInputs); i++ {
		Token = Vec_IntEntry(vInputs, i)
		Ndr_AddObject(pDesign, ModuleID, ABC_OPER_CI, 0, 0, 0, 0, 0, nil, 1, &Token, nil)
	} // no fanins
	if Token = Abc_NamStrFind(pNames, "1'b0"); Token != 0 {
		Ndr_AddObject(pDesign, ModuleID, ABC_OPER_CONST_F, 0, 0, 0, 0, 0, nil, 1, &Token, nil)
	} // no fanins
	if Token = Abc_NamStrFind(pNames, "1'b1"); Token != 0 {
		Ndr_AddObject(pDesign, ModuleID, ABC_OPER_CONST_T, 0, 0, 0, 0, 0, nil, 1, &Token, nil)
	} // no fanins
	if Token = Abc_NamStrFind(pNames, "1'bx"); Token != 0 {
		Ndr_AddObject(pDesign, ModuleID, ABC_OPER_CONST_X, 0, 0, 0, 0, 0, nil, 1, &Token, nil)
	} // no fanins
	for i = 0; i+1 < Vec_IntSize(vTypes); i += 2 {
		Token, Size = Vec_IntEntry(vTypes, i), Vec_IntEntry(vTypes, i+1)
		if Token > 0 {
			var Output int32 = Vec_IntEntry(vFanins, Size)
			var nFanins int32 = Vec_IntEntry(vTypes, i+3) - Size - 1
			var pFanins *int32 = Vec_IntEntryP(vFanins, Size+1)
			Ndr_AddObject(pDesign, ModuleID, Acb_Type2Oper(Token), 0, 0, 0, 0, nFanins, pFanins, 1, &Output, nil) // many fanins
		}
	}
	for i = 0; i < Vec_IntSize(vOutputs); i++ {
		Token = Vec_IntEntry(vOutputs, i)
		Ndr_AddObject(pDesign, ModuleID, ABC_OPER_CO, 0, 0, 0, 0, 1, &Token, 1, &Token, nil)
	} // one fanin
	// cleanup
	Vec_IntFree(vInputs)
	Vec_IntFree(vOutputs)
	Vec_IntFree(vWires)
	Vec_IntFree(vTypes)
	Vec_IntFree(vFanins)
	return pDesign
}

/*
*Function*************************************************************

	Synopsis    [Read weights of nodes from the weight file.]

	Description []

	SideEffects []

	SeeAlso     []

**********************************************************************
*/
func Acb_ReadWeightMap(pFileName string, pNames *Abc_Nam_t) *Vec_Int_t {
	var vWeights *Vec_Int_t = Vec_IntStartFull(Abc_NamObjNumMax(pNames))
	var pBuffer string = Extra_FileReadContents(pFileName)
	if pBuffer == "" {
		return nil
	}
	// The C reader alternates name and value tokens. Its first pair uses
	// whitespace, parentheses and commas; subsequent pairs also accept ';'.
	firstDelimiters := " \n\r(),"
	nextDelimiters := " \n\r(),;"
	first := true
	for {
		delimiters := nextDelimiters
		if first {
			delimiters = firstDelimiters
			first = false
		}
		fields := strings.FieldsFunc(pBuffer, func(r rune) bool { return strings.ContainsRune(delimiters, r) })
		if len(fields) < 2 {
			break
		}
		pToken, pNum := fields[0], fields[1]
		// Consume through the end of the pair, matching successive strtok calls.
		pBuffer = strings.TrimLeft(pBuffer, " \n\r(),")
		for _, token := range fields[:2] {
			at := strings.Index(pBuffer, token)
			if at < 0 {
				break
			}
			pBuffer = pBuffer[at+len(token):]
		}
		var NameId int32 = Abc_NamStrFind(pNames, pToken)
		numeric := strings.TrimSpace(pNum)
		end := 0
		if len(numeric) > 0 && (numeric[0] == '+' || numeric[0] == '-') {
			end++
		}
		for end < len(numeric) && numeric[end] >= '0' && numeric[end] <= '9' {
			end++
		}
		number, _ := strconv.Atoi(numeric[:end]) // C atoi accepts a numeric prefix.
		var Number int32 = int32(number)
		if NameId <= 0 {
			fmt.Printf("Cannot find name \"%s\" among node names of this network.\n", pToken)
			continue
		}
		Vec_IntWriteEntry(vWeights, NameId, Number)
	}
	pBuffer = ""
	return vWeights
}

/*
*Function*************************************************************

	Synopsis    [Create network from the intermediate representation.]

	Description []

	SideEffects []

	SeeAlso     []

**********************************************************************
*/
func Acb_VerilogSimpleRead(pFileName string, pFileNameW string) *Acb_Ntk_t {
	var pNtk *Acb_Ntk_t
	var pNames *Abc_Nam_t = Acb_VerilogStartNames()
	var vBuffer *Vec_Int_t = Acb_VerilogSimpleLex(pFileName, pNames)
	var pModule any
	if vBuffer != nil {
		pModule = Acb_VerilogSimpleParse(vBuffer, pNames)
	} else {
		pModule = nil
	}
	var vWeights *Vec_Int_t
	if pFileNameW != "" {
		vWeights = Acb_ReadWeightMap(pFileNameW, pNames)
	} else {
		vWeights = nil
	}
	if pFileName != "" && pModule == nil {
		fmt.Printf("Cannot read input file \"%s\".\n", pFileName)
		return nil
	}
	if pFileNameW != "" && vWeights == nil {
		fmt.Printf("Cannot read weight file \"%s\".\n", pFileNameW)
		return nil
	}
	//    Ndr_ModuleWriteVerilog( "iccad17/unit1/test.v", pModule, pNameStrs );
	pNtk = Acb_NtkFromNdr(pFileName, pModule, pNames, vWeights, Abc_NamObjNumMax(pNames))
	Ndr_Delete(pModule.(*Ndr_Data_t))
	Vec_IntFree(vBuffer)
	Vec_IntFreeP(&vWeights)
	Abc_NamDeref(pNames)
	return pNtk
}

/*
*Function*************************************************************

	Synopsis    [Write Verilog for sanity checking.]

	Description []

	SideEffects []

	SeeAlso     []

**********************************************************************
*/
func Acb_VerilogSimpleWrite(p *Acb_Ntk_t, pFileName string) {
	var i int32
	var iObj int32
	var fFirst int32 = 1
	pFile, openErr := os.OpenFile(pFileName, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, 0666)
	if openErr != nil {
		fmt.Printf("Cannot open file \"%s\" for writing.\n", pFileName)
		return
	}

	fmt.Fprintf(pFile, "\nmodule %s (\n  ", Acb_NtkName(p))

	// inlining of macro Acb_NtkForEachPi
	for i = 0; i < Acb_NtkPiNum(p); i++ {
		iObj = Acb_NtkCi(p, i)
		fmt.Fprintf(pFile, "%s, ", Acb_ObjNameStr(p, iObj))
	}
	fmt.Fprintf(pFile, "\n  ")

	// inlining of macro Acb_NtkForEachPo
	for i = 0; i < Acb_NtkPoNum(p); i++ {
		iObj = Acb_NtkCo(p, i)
		prefix := ", "
		if fFirst != 0 {
			prefix = ""
		}
		fmt.Fprintf(pFile, "%s%s", prefix, Acb_ObjNameStr(p, iObj))
		fFirst = 0
	}
	fmt.Fprintf(pFile, "\n);\n\n")

	// inlining of macro Acb_NtkForEachPi
	for i = 0; i < Acb_NtkPiNum(p); i++ {
		iObj = Acb_NtkCi(p, i)
		fmt.Fprintf(pFile, "  input %s;\n", Acb_ObjNameStr(p, iObj))
	}
	fmt.Fprintf(pFile, "\n")

	// inlining of macro Acb_NtkForEachPo
	for i = 0; i < Acb_NtkPoNum(p); i++ {
		iObj = Acb_NtkCo(p, i)
		fmt.Fprintf(pFile, "  output %s;\n", Acb_ObjNameStr(p, iObj))
	}
	fmt.Fprintf(pFile, "\n")

	// inlining of macro Acb_NtkForEachNode
	for iObj = 1; iObj < Vec_StrSize(&p.vObjType); iObj++ {
		if Acb_ObjType(p, iObj) == 0 || Acb_ObjIsCio(p, iObj) != 0 {
		} else if Acb_ObjFaninNum(p, iObj) > 0 {
			fmt.Fprintf(pFile, "  wire %s;\n", Acb_ObjNameStr(p, iObj))
		}
	}
	fmt.Fprintf(pFile, "\n")

	// inlining of macro Acb_NtkForEachNode
	for iObj = 1; iObj < Vec_StrSize(&p.vObjType); iObj++ {
		if Acb_ObjType(p, iObj) == 0 || Acb_ObjIsCio(p, iObj) != 0 {
		} else if Acb_ObjFaninNum(p, iObj) > 0 {
			var pFanin []int32
			var iFanin int32
			var k int32
			var start int32 = fileTell(pFile) + 55
			fmt.Fprintf(pFile, "  %s (", Acb_Oper2Name(int32(Acb_ObjType(p, iObj))))
			fmt.Fprintf(pFile, " %s", Acb_ObjNameStr(p, iObj))
			// inlining of macro Acb_ObjForEachFaninFast
			pFanin = Acb_ObjFanins(p, iObj)
			for k = 0; k < pFanin[0]; k++ {
				iFanin = pFanin[k+1]
				//if ( iFanin == 0 ) fprintf( pFile, ", <zero>" ); else
				fmt.Fprintf(pFile, ", %s", Acb_ObjNameStr(p, iFanin))
			}
			fmt.Fprintf(pFile, " );")
			if Acb_NtkHasObjWeights(p) != 0 && Acb_ObjWeight(p, iObj) > 0 {
				fmt.Fprintf(pFile, " %*s // weight = %d", int32((start - fileTell(pFile))), "", Acb_ObjWeight(p, iObj))
			}
			fmt.Fprintf(pFile, "\n")

		} else { // constant nodes

			assert(Acb_ObjType(p, iObj) == ABC_OPER_CONST_F || Acb_ObjType(p, iObj) == ABC_OPER_CONST_T, "Acb_ObjType(p, iObj) == ABC_OPER_CONST_F || Acb_ObjType(p, iObj) == ABC_OPER_CONST_T")
			fmt.Fprintf(pFile, "  %s (", Acb_Oper2Name(ABC_OPER_BIT_BUF))
			if Acb_ObjType(p, iObj) == ABC_OPER_CONST_X {
				fmt.Fprintf(pFile, " 1'bx")
			} else {
				bit := 0
				if Acb_ObjType(p, iObj) == ABC_OPER_CONST_T {
					bit = 1
				}
				fmt.Fprintf(pFile, " 1'b%d", bit)
			}
			fmt.Fprintf(pFile, " );\n")
		}
	}

	fmt.Fprintf(pFile, "\nendmodule\n\n")
	_ = (pFile).Close()
}

/*
*Function*************************************************************

	Synopsis    [Compute roots (PO nodes in the TFO of the targets in F).]

	Description []

	SideEffects []

	SeeAlso     []

**********************************************************************
*/
func Acb_NtkFindRoots_rec(p *Acb_Ntk_t, iObj int32, vBlock *Vec_Bit_t) int32 {
	var pFanin []int32
	var iFanin int32
	var i int32
	var Res int32
	if Acb_ObjIsTravIdPrev(p, iObj) != 0 {
		return 1
	}
	if Acb_ObjSetTravIdCur(p, iObj) != 0 {
		return 0
	}
	assert(Acb_ObjIsCi(p, iObj) == 0, "Acb_ObjIsCi(p, iObj) == 0")
	// inlining of macro Acb_ObjForEachFaninFast
	pFanin = Acb_ObjFanins(p, iObj)
	for i = 0; i < pFanin[0]; i++ {
		iFanin = pFanin[i+1]
		Res |= Acb_NtkFindRoots_rec(p, iFanin, vBlock)
	}
	if Res != 0 {
		Acb_ObjSetTravIdPrev(p, iObj)
	}
	if Res != 0 {
		Vec_BitWriteEntry(vBlock, iObj, 1)
	}
	return Res
}
func Acb_NtkFindRoots(p *Acb_Ntk_t, vTargets *Vec_Int_t, pvBlock **Vec_Bit_t) *Vec_Int_t {
	var i, iObj int32
	var vRoots *Vec_Int_t = Vec_IntAlloc(1000)
	var vBlock *Vec_Bit_t = Vec_BitStart(Acb_NtkObjNum(p))
	*pvBlock = vBlock
	// mark targets
	Acb_NtkIncTravId(p)
	assert(Vec_IntSize(vTargets) > 0, "Vec_IntSize(vTargets) > 0")
	// inlining of macro Vec_IntForEachEntry
	for i = 0; i < Vec_IntSize(vTargets); i++ {
		iObj = Vec_IntEntry(vTargets, i)
		Acb_ObjSetTravIdCur(p, iObj)
		Vec_BitWriteEntry(vBlock, iObj, 1)
	}
	// mark inputs
	Acb_NtkIncTravId(p)
	// inlining of macro Acb_NtkForEachCi
	for i = 0; i < Acb_NtkCiNum(p); i++ {
		iObj = Acb_NtkCi(p, i)
		Acb_ObjSetTravIdCur(p, iObj)
	}
	// visit internal nodes
	// inlining of macro Acb_NtkForEachNode
	for iObj = 1; iObj < Vec_StrSize(&p.vObjType); iObj++ {
		if Acb_ObjType(p, iObj) == 0 || Acb_ObjIsCio(p, iObj) != 0 {
		} else {
			Acb_NtkFindRoots_rec(p, iObj, vBlock)
		}
	}
	// collect outputs reachable from targets
	// inlining of macro Acb_NtkForEachCoDriver
	for i = 0; i < Acb_NtkCoNum(p); i++ {
		iObj = Acb_ObjFanin(p, Acb_NtkCo(p, i), 0)
		if Acb_NtkFindRoots_rec(p, iObj, vBlock) != 0 {
			Vec_IntPush(vRoots, i)
		}
	}
	//Vec_IntPrint( vRoots );
	return vRoots
}

/*
*Function*************************************************************

	Synopsis    [Compute support in the TFI of the roots.]

	Description []

	SideEffects []

	SeeAlso     []

**********************************************************************
*/
func Acb_NtkFindSupp_rec(p *Acb_Ntk_t, iObj int32, vSupp *Vec_Int_t) {
	var iFanin int32
	var i int32
	if Acb_ObjSetTravIdCur(p, iObj) != 0 {
		return
	}
	if Acb_ObjIsCi(p, iObj) != 0 {
		Vec_IntPush(vSupp, Acb_ObjCioId(p, iObj))
	} else {
		// inlining of macro Acb_ObjForEachFaninFast
		pFanins := Acb_ObjFanins(p, iObj)
		for i = 0; i < pFanins[0]; i++ {
			iFanin = pFanins[i+1]
			Acb_NtkFindSupp_rec(p, iFanin, vSupp)
		}
	}
}
func Acb_NtkFindSupp(p *Acb_Ntk_t, vRoots *Vec_Int_t) *Vec_Int_t {
	var i, iObj int32
	var vSupp *Vec_Int_t = Vec_IntAlloc(1000)
	Acb_NtkIncTravId(p)
	// inlining of macro Acb_NtkForEachCoDriverVec
	for i = 0; i < Vec_IntSize(vRoots); i++ {
		iObj = Acb_ObjFanin(p, Acb_NtkCo(p, Vec_IntEntry(vRoots, i)), 0)
		Acb_NtkFindSupp_rec(p, iObj, vSupp)
	}
	Vec_IntSort(vSupp, 0)
	//Vec_IntPrint( vSupp );
	return vSupp
}

/*
*Function*************************************************************

	Synopsis    [Collect divisors whose support is completely contained in vSupp.]

	Description []

	SideEffects []

	SeeAlso     []

**********************************************************************
*/
func Acb_NtkFindDivs_rec(p *Acb_Ntk_t, iObj int32) int32 {
	var pFanin []int32
	var iFanin int32
	var i int32
	var Res int32 = 1
	if Acb_ObjIsTravIdPrev(p, iObj) != 0 {
		return 1
	}
	if Acb_ObjSetTravIdCur(p, iObj) != 0 {
		return 0
	}
	if Acb_ObjIsCi(p, iObj) != 0 {
		return 0
	}
	// inlining of macro Acb_ObjForEachFaninFast
	pFanin = Acb_ObjFanins(p, iObj)
	for i = 0; i < pFanin[0]; i++ {
		iFanin = pFanin[i+1]
		Res &= Acb_NtkFindDivs_rec(p, iFanin)
	}
	if Res != 0 {
		Acb_ObjSetTravIdPrev(p, iObj)
	}
	return Res
}
func Acb_NtkFindDivsCis(p *Acb_Ntk_t, vSupp *Vec_Int_t) *Vec_Int_t {
	var i, iObj int32
	var vDivs *Vec_Int_t = Vec_IntAlloc(Vec_IntSize(vSupp))
	// inlining of macro Acb_NtkForEachCiVec
	for i = 0; i < Vec_IntSize(vSupp); i++ {
		iObj = Acb_NtkCi(p, Vec_IntEntry(vSupp, i))
		assert(Acb_ObjWeight(p, iObj) > 0, "Acb_ObjWeight(p, iObj) > 0")
		Vec_IntPush(vDivs, iObj)
	}
	fmt.Printf("Divisors are %d support variables (CIs in the TFO of the targets).\n", Vec_IntSize(vSupp))
	return vDivs
}
func Acb_NtkFindDivs(p *Acb_Ntk_t, vSupp *Vec_Int_t, vBlock *Vec_Bit_t, fVerbose int32) *Vec_Int_t {
	var fPrintWeights int32
	var nDivLimit int32 = 5000
	var i, iObj int32
	var vDivs *Vec_Int_t = Vec_IntAlloc(1000)
	// mark inputs
	Acb_NtkIncTravId(p)
	// inlining of macro Acb_NtkForEachCiVec
	for i = 0; i < Vec_IntSize(vSupp); i++ {
		iObj = Acb_NtkCi(p, Vec_IntEntry(vSupp, i))
		Acb_ObjSetTravIdCur(p, iObj)
		if Acb_ObjWeight(p, iObj) > 0 {
			Vec_IntPush(vDivs, iObj)
		}
	}
	// collect nodes whose support is contained in vSupp
	Acb_NtkIncTravId(p)
	// inlining of macro Acb_NtkForEachNode
	for iObj = 1; iObj < Vec_StrSize(&p.vObjType); iObj++ {
		if Acb_ObjType(p, iObj) == 0 || Acb_ObjIsCio(p, iObj) != 0 {
		} else if Vec_BitEntry(vBlock, iObj) == 0 && Acb_ObjWeight(p, iObj) > 0 && (Acb_NtkFindDivs_rec(p, iObj) != 0) {
			Vec_IntPush(vDivs, iObj)
		}
	}
	// sort divisors by cost (first cheap ones; later expensive ones)
	Vec_IntSelectSortCost(Vec_IntArray(vDivs), Vec_IntSize(vDivs), &p.vObjWeight)
	//Vec_IntPrint( vDivs );
	nDivLimit = Abc_MinInt(Vec_IntSize(vDivs), nDivLimit)
	if fPrintWeights != 0 {
		//        Vec_IntForEachEntryStop( vDivs, iObj, i, nDivLimit )
		//            printf( "%d:%d (w=%d)   ", i, iObj, Vec_IntEntry(&p->vObjWeight, iObj) );
		//        printf( "\n" );

		fmt.Printf("    : ")
		// inlining of macro Vec_IntForEachEntryStop
		for i = 0; i < nDivLimit; i++ {
			iObj = Vec_IntEntry(vDivs, i)
			fmt.Printf("%d", (Vec_IntEntry(&p.vObjWeight, iObj)/100)%10)
		}
		fmt.Printf("\n")
		fmt.Printf("    : ")
		// inlining of macro Vec_IntForEachEntryStop
		for i = 0; i < nDivLimit; i++ {
			iObj = Vec_IntEntry(vDivs, i)
			fmt.Printf("%d", (Vec_IntEntry(&p.vObjWeight, iObj)/10)%10)
		}
		fmt.Printf("\n")
		fmt.Printf("    : ")
		// inlining of macro Vec_IntForEachEntryStop
		for i = 0; i < nDivLimit; i++ {
			iObj = Vec_IntEntry(vDivs, i)
			fmt.Printf("%d", (Vec_IntEntry(&p.vObjWeight, iObj)/1)%10)
		}
		fmt.Printf("\n")
	}
	if fVerbose != 0 {
		fmt.Printf("Reducing divisor set from %d to ", Vec_IntSize(vDivs))
	}
	Vec_IntShrink(vDivs, nDivLimit)
	if fVerbose != 0 {
		fmt.Printf("%d.\n", Vec_IntSize(vDivs))
	}
	return vDivs
}

/*
*Function*************************************************************

	Synopsis    [Compute support and internal nodes in the TFI of the roots.]

	Description []

	SideEffects []

	SeeAlso     []

**********************************************************************
*/
func Acb_NtkFindNodes_rec(p *Acb_Ntk_t, iObj int32, vNodes *Vec_Int_t) {
	var pFanin []int32
	var iFanin int32
	var i int32
	if Acb_ObjSetTravIdCur(p, iObj) != 0 {
		return
	}
	if Acb_ObjIsCi(p, iObj) != 0 {
		return
	}
	// inlining of macro Acb_ObjForEachFaninFast
	pFanin = Acb_ObjFanins(p, iObj)
	for i = 0; i < pFanin[0]; i++ {
		iFanin = pFanin[i+1]
		Acb_NtkFindNodes_rec(p, iFanin, vNodes)
	}
	assert(Acb_ObjIsCo(p, iObj) == 0, "Acb_ObjIsCo(p, iObj) == 0")
	Vec_IntPush(vNodes, iObj)
}
func Acb_NtkFindNodes(p *Acb_Ntk_t, vRoots *Vec_Int_t, vDivs *Vec_Int_t) *Vec_Int_t {
	var i, iObj int32
	var vNodes *Vec_Int_t = Vec_IntAlloc(1000)
	Acb_NtkIncTravId(p)
	// inlining of macro Acb_NtkForEachCoDriverVec
	for i = 0; i < Vec_IntSize(vRoots); i++ {
		iObj = Acb_ObjFanin(p, Acb_NtkCo(p, Vec_IntEntry(vRoots, i)), 0)
		Acb_NtkFindNodes_rec(p, iObj, vNodes)
	}
	if vDivs != nil {
		// inlining of macro Vec_IntForEachEntry
		for i = 0; i < Vec_IntSize(vDivs); i++ {
			iObj = Vec_IntEntry(vDivs, i)
			Acb_NtkFindNodes_rec(p, iObj, vNodes)
		}
	}
	return vNodes
}

/*
*Function*************************************************************

	Synopsis    [Derive AIG for one network.]

	Description []

	SideEffects []

	SeeAlso     []

**********************************************************************
*/
func Acb_ObjToGia(pNew *Gia_Man_t, p *Acb_Ntk_t, iObj int32, vTemp *Vec_Int_t) int32 {
	//char * pName = Abc_NamStr( p->pDesign->pStrs, Acb_ObjName(p, iObj) );
	var pFanin []int32
	var iFanin int32
	var k int32
	var Type int32
	var Res int32
	assert(Acb_ObjIsCio(p, iObj) == 0, "Acb_ObjIsCio(p, iObj) == 0")
	Vec_IntClear(vTemp)
	// inlining of macro Acb_ObjForEachFaninFast
	pFanin = Acb_ObjFanins(p, iObj)
	for k = 0; k < pFanin[0]; k++ {
		iFanin = pFanin[k+1]
		assert(Acb_ObjCopy(p, iFanin) >= 0, "Acb_ObjCopy(p, iFanin) >= 0")
		Vec_IntPush(vTemp, Acb_ObjCopy(p, iFanin))
	}
	Type = int32(Acb_ObjType(p, iObj))
	if Type == ABC_OPER_CONST_F {
		return 0
	}
	if Type == ABC_OPER_CONST_T {
		return 1
	}
	if Type == ABC_OPER_BIT_BUF {
		return Vec_IntEntry(vTemp, 0)
	}
	if Type == ABC_OPER_BIT_INV {
		return Abc_LitNot(Vec_IntEntry(vTemp, 0))
	}
	if Type == ABC_OPER_BIT_AND || Type == ABC_OPER_BIT_NAND {
		Res = 1
		// inlining of macro Vec_IntForEachEntry
		for k = 0; k < Vec_IntSize(vTemp); k++ {
			iFanin = Vec_IntEntry(vTemp, k)
			Res = Gia_ManHashAnd(pNew, Res, iFanin)
		}
		return Abc_LitNotCond(Res, Type == ABC_OPER_BIT_NAND)
	}
	if Type == ABC_OPER_BIT_OR || Type == ABC_OPER_BIT_NOR {
		Res = 0
		// inlining of macro Vec_IntForEachEntry
		for k = 0; k < Vec_IntSize(vTemp); k++ {
			iFanin = Vec_IntEntry(vTemp, k)
			Res = Gia_ManHashOr(pNew, Res, iFanin)
		}
		return Abc_LitNotCond(Res, Type == ABC_OPER_BIT_NOR)
	}
	if Type == ABC_OPER_BIT_XOR || Type == ABC_OPER_BIT_NXOR {
		Res = 0
		// inlining of macro Vec_IntForEachEntry
		for k := int32(0); k < Vec_IntSize(vTemp); k++ {
			iFanin = Vec_IntEntry(vTemp, k)
			Res = Gia_ManHashXor(pNew, Res, iFanin)
		}
		return Abc_LitNotCond(Res, Type == ABC_OPER_BIT_NXOR)
	}
	assert(false, "0")
	return -1
}
func Acb_NtkToGia(p *Acb_Ntk_t, vSupp *Vec_Int_t, vNodes *Vec_Int_t, vRoots *Vec_Int_t, vDivs *Vec_Int_t, vTargets *Vec_Int_t) *Gia_Man_t {
	var pNew, pOne *Gia_Man_t
	var vFanins *Vec_Int_t
	var i, iObj int32
	pNew = Gia_ManStart(2*Acb_NtkObjNum(p) + 1000)
	Gia_ManHashAlloc(pNew)
	Acb_NtkCleanObjCopies(p)
	// create primary inputs
	// inlining of macro Acb_NtkForEachCiVec
	for i = 0; i < Vec_IntSize(vSupp); i++ {
		iObj = Acb_NtkCi(p, Vec_IntEntry(vSupp, i))
		Acb_ObjSetCopy(p, iObj, Gia_ManAppendCi(pNew))
	}
	// add targets as primary inputs
	if vTargets != nil {
		// inlining of macro Vec_IntForEachEntry
		for i = 0; i < Vec_IntSize(vTargets); i++ {
			iObj = Vec_IntEntry(vTargets, i)
			Acb_ObjSetCopy(p, iObj, Gia_ManAppendCi(pNew))
		}
	}
	// bit-blast internal nodes
	vFanins = Vec_IntAlloc(4)
	// inlining of macro Vec_IntForEachEntry
	for i = 0; i < Vec_IntSize(vNodes); i++ {
		iObj = Vec_IntEntry(vNodes, i)
		if Acb_ObjCopy(p, iObj) == -1 { // skip targets assigned above
			Acb_ObjSetCopy(p, iObj, Acb_ObjToGia(pNew, p, iObj, vFanins))
		}
	}
	Vec_IntFree(vFanins)
	// create primary outputs
	// inlining of macro Acb_NtkForEachCoDriverVec
	for i = 0; i < Vec_IntSize(vRoots); i++ {
		iObj = Acb_ObjFanin(p, Acb_NtkCo(p, Vec_IntEntry(vRoots, i)), 0)
		Gia_ManAppendCo(pNew, Acb_ObjCopy(p, iObj))
	}
	// add divisor as primary outputs (if the divisors are not primary inputs)
	if vDivs != nil { //&& Vec_IntSize(vDivs) > Vec_IntSize(vSupp) )
		// inlining of macro Vec_IntForEachEntry
		for i = 0; i < Vec_IntSize(vDivs); i++ {
			iObj = Vec_IntEntry(vDivs, i)
			Gia_ManAppendCo(pNew, Acb_ObjCopy(p, iObj))
		}
	}
	Gia_ManHashStop(pNew)
	pOne = pNew
	pNew = Gia_ManCleanup(pNew)
	Gia_ManStop(pOne)
	return pNew
}

/*
*Function*************************************************************

	Synopsis    [Derive miter of two AIGs.]

	Description []

	SideEffects []

	SeeAlso     []

**********************************************************************
*/
func Acb_CreateMiter(pF *Gia_Man_t, pG *Gia_Man_t) *Gia_Man_t {
	var pNew, pOne *Gia_Man_t
	var pObj *Gia_Obj_t
	var i, iMiter, iXor int32
	Gia_ManFillValue(pF)
	Gia_ManFillValue(pG)
	pNew = Gia_ManStart(Gia_ManObjNum(pF) + Gia_ManObjNum(pF) + 1000)
	Gia_ManHashAlloc(pNew)
	Gia_ManConst0(pF).Value = 0
	Gia_ManConst0(pG).Value = 0
	// inlining of macro Gia_ManForEachCi
	for i = 0; i < Gia_ManCiNum(pF); i++ {
		pObj = Gia_ManCi(pF, i)
		pObj.Value = uint32(Gia_ManAppendCi(pNew))
	}
	// inlining of macro Gia_ManForEachCi
	for i = 0; i < Gia_ManCiNum(pG); i++ {
		pObj = Gia_ManCi(pG, i)
		pObj.Value = Gia_ManCi(pF, i).Value
	}
	// inlining of macro Gia_ManForEachAnd
	for i = 0; i < Gia_ManObjNum(pF); i++ {
		pObj = Gia_ManObj(pF, i)
		if Gia_ObjIsAnd(pObj) == 0 {
		} else {
			pObj.Value = uint32(Gia_ManHashAnd(pNew, Gia_ObjFanin0Copy(pObj), Gia_ObjFanin1Copy(pObj)))
		}
	}
	// inlining of macro Gia_ManForEachAnd
	for i = 0; i < Gia_ManObjNum(pG); i++ {
		pObj = Gia_ManObj(pG, i)
		if Gia_ObjIsAnd(pObj) == 0 {
		} else {
			pObj.Value = uint32(Gia_ManHashAnd(pNew, Gia_ObjFanin0Copy(pObj), Gia_ObjFanin1Copy(pObj)))
		}
	}
	// inlining of macro Gia_ManForEachCo
	for i = 0; i < Gia_ManCoNum(pG); i++ {
		pObj = Gia_ManCo(pG, i)
		iXor = Gia_ManHashXor(pNew, Gia_ObjFanin0Copy(pObj), Gia_ObjFanin0Copy(Gia_ManCo(pF, i)))
		iMiter = Gia_ManHashOr(pNew, iMiter, iXor)
	}
	Gia_ManAppendCo(pNew, iMiter)
	// inlining of macro Gia_ManForEachCo
	for i = 0; i < Gia_ManCoNum(pF); i++ {
		pObj = Gia_ManCo(pF, i)
		if i >= Gia_ManCoNum(pG) {
			Gia_ManAppendCo(pNew, Gia_ObjFanin0Copy(pObj))
		}
	}
	Gia_ManHashStop(pNew)
	pOne = pNew
	pNew = Gia_ManCleanup(pNew)
	Gia_ManStop(pOne)
	return pNew
}

/*
*Function*************************************************************

	Synopsis    []

	Description []

	SideEffects []

	SeeAlso     []

**********************************************************************
*/
const NWORDS = 256

// acbCRand mirrors Linux/glibc's srand/rand TYPE_3 generator used by ABC's
// Vec_IntPermute functions. The C source seeds with time(NULL), then consumes
// one rand() result per swap.
type acbCRand struct {
	mu          sync.Mutex
	state       [31]uint32
	front, rear int
}

func (r *acbCRand) seed(seed uint32) {
	r.mu.Lock()
	defer r.mu.Unlock()
	if seed == 0 {
		seed = 1
	}
	r.front, r.rear = 3, 0
	r.state[0] = seed
	value := int64(int32(seed))
	for i := 1; i < len(r.state); i++ {
		value = (16807 * value) % 2147483647
		if value < 0 {
			value += 2147483647
		}
		r.state[i] = uint32(value)
	}
	for i := 0; i < 310; i++ {
		r.nextUnlocked()
	}
}

func (r *acbCRand) nextUnlocked() uint32 {
	r.state[r.front] += r.state[r.rear]
	value := r.state[r.front] >> 1
	r.front = (r.front + 1) % len(r.state)
	r.rear = (r.rear + 1) % len(r.state)
	return value
}

func (r *acbCRand) next() uint32 {
	r.mu.Lock()
	defer r.mu.Unlock()
	return r.nextUnlocked()
}

var acbGlobalRand acbCRand

func vecIntPermuteWithRand(p *Vec_Int_t, draw func() uint32) {
	nSize := Vec_IntSize(p)
	pArray := Vec_IntArray(p)
	for i := int32(0); i < nSize; i++ {
		j := int32(draw() % uint32(nSize))
		pArray[i], pArray[j] = pArray[j], pArray[i]
	}
}

func Vec_IntPermute(p *Vec_Int_t) {
	acbGlobalRand.seed(uint32(time.Now().Unix()))
	vecIntPermuteWithRand(p, acbGlobalRand.next)
}

func Vec_IntPermute2(p *Vec_Int_t) {
	acbGlobalRand.seed(uint32(time.Now().Unix()))
	nSize := Vec_IntSize(p)
	pArray := Vec_IntArray(p)
	for i := int32(0); i < nSize-1; i++ {
		if acbGlobalRand.next()%3 == 0 {
			fmt.Printf("Permuting %d and %d\n", i, i+1)
			pArray[i], pArray[i+1] = pArray[i+1], pArray[i]
		}
	}
}

func Acb_PrintPatterns(vPats *Vec_Wrd_t, nPats int32, vWeights *Vec_Int_t) {
	var i int32
	var k int32
	var Entry int32
	var nDivs int32 = Vec_IntSize(vWeights)
	fmt.Printf("    : ")
	// inlining of macro Vec_IntForEachEntry
	for i = 0; i < Vec_IntSize(vWeights); i++ {
		Entry = Vec_IntEntry(vWeights, i)
		fmt.Printf("%d", (Entry/100)%10)
	}
	fmt.Printf("\n")
	fmt.Printf("    : ")
	// inlining of macro Vec_IntForEachEntry
	for i = 0; i < Vec_IntSize(vWeights); i++ {
		Entry = Vec_IntEntry(vWeights, i)
		fmt.Printf("%d", (Entry/10)%10)
	}
	fmt.Printf("\n")
	fmt.Printf("    : ")
	// inlining of macro Vec_IntForEachEntry
	for i = 0; i < Vec_IntSize(vWeights); i++ {
		Entry = Vec_IntEntry(vWeights, i)
		fmt.Printf("%d", (Entry/1)%10)
	}
	fmt.Printf("\n")
	fmt.Printf("\n")

	fmt.Printf("    : ")
	for i = 0; i < nDivs; i++ {
		fmt.Printf("%d", (i/100)%10)
	}
	fmt.Printf("\n")
	fmt.Printf("    : ")
	for i = 0; i < nDivs; i++ {
		fmt.Printf("%d", (i/10)%10)
	}
	fmt.Printf("\n")
	fmt.Printf("    : ")
	for i = 0; i < nDivs; i++ {
		fmt.Printf("%d", i%10)
	}
	fmt.Printf("\n")
	fmt.Printf("\n")

	for k = 0; k < nPats; k++ {
		fmt.Printf("%3d : ", k)
		for i = 0; i < nDivs; i++ {
			bit := '|'
			if Abc_TtGetBit(Vec_WrdEntryP(vPats, NWORDS*i), k) != 0 {
				bit = '*'
			}
			fmt.Printf("%c", bit)
		}
		fmt.Printf("\n")
	}
	fmt.Printf("\n")
}

func Acb_DeriveWeights(vDivs *Vec_Int_t, pNtkF *Acb_Ntk_t) *Vec_Int_t {
	var i, iDiv int32
	var vWeights *Vec_Int_t = Vec_IntAlloc(Vec_IntSize(vDivs))
	// inlining of macro Vec_IntForEachEntry
	for i = 0; i < Vec_IntSize(vDivs); i++ {
		iDiv = Vec_IntEntry(vDivs, i)
		Vec_IntPush(vWeights, Vec_IntEntry(&pNtkF.vObjWeight, iDiv))
	}
	return vWeights
}
func Acb_ComputeSuppCost(vSupp *Vec_Int_t, vWeights *Vec_Int_t, iFirstDiv int32) int32 {
	var i, Entry, Cost int32
	// inlining of macro Vec_IntForEachEntry
	for i = 0; i < Vec_IntSize(vSupp); i++ {
		Entry = Vec_IntEntry(vSupp, i)
		Cost += Vec_IntEntry(vWeights, Abc_Lit2Var(Entry)-iFirstDiv)
	}
	return Cost
}
func Acb_FindSupportStart(pSat *sat_solver, iFirstDiv int32, vWeights *Vec_Int_t, pvPats **Vec_Wrd_t, piPats *int32) *Vec_Int_t {
	var i int32
	var status int32
	var nDivs int32 = Vec_IntSize(vWeights)
	var vSupp *Vec_Int_t = Vec_IntAlloc(100)
	var vPats *Vec_Wrd_t = Vec_WrdStart(NWORDS * nDivs)
	var iPat int32
	for {
		var fFound int32
		// try one run
		status = sat_solver_solve(pSat, Vec_IntArray(vSupp)[:Vec_IntSize(vSupp)], nil, 0, 0, 0, 0)
		if status == l_False {
			break
		}
		assert(status == l_True, "status == l_True")
		// collect pattern
		for i = 0; i < nDivs; i++ {
			if sat_solver_var_value(pSat, iFirstDiv+i) == 0 {
				continue
			}
			Abc_TtSetBit(Vec_WrdEntryP(vPats, NWORDS*i), iPat)
			if fFound != 0 {
				continue
			}
			// process new divisor
			Vec_IntPush(vSupp, Abc_Var2Lit(iFirstDiv+i, 1))
			//printf( "Selecting divisor %d with weight %d\n", i, Vec_IntEntry(vWeights, i) );
			fFound = 1
		}
		if fFound == 0 {
			break
		}
		/*
		   if ( fFound == 0 )
		   {
		       printf( "For some reason, cannot find a divisor.\n" );
		       Vec_WrdFree( vPats );
		       Vec_IntFree( vSupp );
		       return NULL;
		   }
		*/
		assert((fFound != 0), "fFound")
		iPat++
	}
	*piPats = iPat
	*pvPats = vPats
	Vec_IntSort(vSupp, 0)
	return vSupp
}
func Acb_FindArgMaxUnderMask(vPats *Vec_Wrd_t, Mask []word, vWeights *Vec_Int_t, nPats int32) int32 {
	var nDivs int32 = Vec_WrdSize(vPats) / NWORDS
	var nWords int32 = Abc_Bit6WordNum(nPats)
	var i int32
	var iBest int32 = -1
	var Cost int32
	var CostBest int32 = -1
	for i = 0; i < nDivs; i++ {
		var pPat *word = Vec_WrdEntryP(vPats, NWORDS*i)
		Cost = Abc_TtCountOnesVecMask(Mask, pPat, nWords, 0)
		if CostBest < Cost {
			//        if ( CostBest == -1 || (float)CostBest/Cost < 0.6*(float)Vec_IntEntry(vWeights, iBest)/Vec_IntEntry(vWeights, i) )
			//        if ( CostBest == -1 || (float)CostBest/Cost < 0.67*(float)Vec_IntEntry(vWeights, iBest)/Vec_IntEntry(vWeights, i) )

			CostBest = Cost
			iBest = i
		}
	}
	return iBest
}
func Acb_FindArgMaxUnderMask2(vPats *Vec_Wrd_t, Mask []word, vWeights *Vec_Int_t, nPats int32) int32 {
	var nDivs int32 = Vec_WrdSize(vPats) / NWORDS
	var i int32
	var b int32
	var iBest int32 = -1
	var Cost int32
	var CostBest int32 = -1
	// count how many times each of them appears
	var vCounts *Vec_Int_t = Vec_IntStart(nPats)
	for i = 0; i < nDivs; i++ {
		var pPat *word = Vec_WrdEntryP(vPats, NWORDS*i)
		for b = 0; b < nPats; b++ {
			if Abc_TtGetBit(pPat, b) != 0 {
				Vec_IntAddToEntry(vCounts, b, 1)
			}
		}
	}
	for i = 0; i < nDivs; i++ {
		var pPat *word = Vec_WrdEntryP(vPats, NWORDS*i)
		//        Cost = Abc_TtCountOnesVecMask(Mask, pPat, NWORDS, 0);
		Cost = 0
		for b = 0; b < nPats; b++ {
			if Abc_TtGetBit(pPat, b) != 0 && Abc_TtGetBit(Mask, b) != 0 {
				Cost += 1000000 / Vec_IntEntry(vCounts, b)
			}
		}
		if CostBest < Cost {
			CostBest = Cost
			iBest = i
		}
	}
	Vec_IntFree(vCounts)
	return iBest
}

func Acb_FindSupportNext(pSat *sat_solver, iFirstDiv int32, vWeights *Vec_Int_t, vPats *Vec_Wrd_t, pnPats *int32) *Vec_Int_t {
	var i int32
	var status int32
	var nDivs int32 = Vec_IntSize(vWeights)
	var vSupp *Vec_Int_t = Vec_IntAlloc(100)
	var pMask *word
	Mask := make([]word, NWORDS)
	Abc_TtConst(Mask, NWORDS, 1)
	for {
		var iDivBest int32 = Acb_FindArgMaxUnderMask(vPats, Mask, vWeights, *pnPats)
		Vec_IntPush(vSupp, Abc_Var2Lit(iFirstDiv+iDivBest, 1))
		//printf( "Selecting divisor %d with weight %d\n", iDivBest, Vec_IntEntry(vWeights, iDivBest) );
		//        Mask &= ~Vec_WrdEntry( vPats, iDivBest );
		pMask = Vec_WrdEntryP(vPats, NWORDS*iDivBest)
		Abc_TtAndSharp(Mask, Mask, pMask, NWORDS, 1)

		// try one run
		status = sat_solver_solve(pSat, Vec_IntArray(vSupp)[:Vec_IntSize(vSupp)], nil, 0, 0, 0, 0)
		if status == l_False {
			break
		}
		assert(status == l_True, "status == l_True")
		// collect pattern
		for i = 0; i < nDivs; i++ {
			if sat_solver_var_value(pSat, iFirstDiv+i) == 0 {
				continue
			}
			Abc_TtSetBit(Vec_WrdEntryP(vPats, NWORDS*i), *pnPats)
		}
		(*pnPats)++
		if *pnPats == NWORDS*64 {
			fmt.Printf("Exceeded %d words.\n", NWORDS)
			Vec_IntFreeP(&vSupp)
			return nil
		}
		assert(*pnPats < NWORDS*64, "*pnPats < NWORDS*64")
		//Acb_PrintPatterns( vPats, *pnPats, vWeights );
		//i = i;
	}
	Vec_IntSort(vSupp, 0)
	return vSupp
}
func Acb_FindSupportMinOne(pSat *sat_solver, iFirstDiv int32, vPats *Vec_Wrd_t, pnPats *int32, vSupp *Vec_Int_t, iVar int32) *Vec_Int_t {
	var i, iLit, status int32
	var nDivs int32 = Vec_WrdSize(vPats) / NWORDS
	var vLits *Vec_Int_t = Vec_IntAlloc(Vec_IntSize(vSupp))
	// inlining of macro Vec_IntForEachEntry
	for i = 0; i < Vec_IntSize(vSupp); i++ {
		iLit = Vec_IntEntry(vSupp, i)
		if i != iVar {
			Vec_IntPush(vLits, iLit)
		}
	}
	// try one run
	status = sat_solver_solve(pSat, Vec_IntArray(vLits)[:Vec_IntSize(vLits)], nil, 0, 0, 0, 0)
	if status == l_False {
		return vLits
	}
	Vec_IntFree(vLits)
	assert(status == l_True, "status == l_True")
	// collect pattern
	for i = 0; i < nDivs; i++ {
		if sat_solver_var_value(pSat, iFirstDiv+i) == 0 {
			continue
		}
		Abc_TtSetBit(Vec_WrdEntryP(vPats, NWORDS*i), *pnPats)
	}
	(*pnPats)++
	if *pnPats == NWORDS*64 {
		return nil
	}
	return vSupp
}
func Acb_FindSupportMin(pSat *sat_solver, iFirstDiv int32, vPats *Vec_Wrd_t, pnPats *int32, vSuppStart *Vec_Int_t) *Vec_Int_t {
	var vTemp *Vec_Int_t
	var vSupp *Vec_Int_t = Vec_IntDup(vSuppStart)
	var i int32
	for i = Vec_IntSize(vSupp) - 1; i >= 0; i-- {
		vTemp = vSupp
		vSupp = Acb_FindSupportMinOne(pSat, iFirstDiv, vPats, pnPats, vTemp, i)
		if vTemp != vSupp {
			Vec_IntFree(vTemp)
		}
		if vSupp == nil {
			return nil
		}
	}
	return vSupp
}
func Acb_FindReplace(pSat *sat_solver, iFirstDiv int32, vWeights *Vec_Int_t, vPats *Vec_Wrd_t, nPats int32, vSupp *Vec_Int_t) {
	var i int32
	var k int32
	var iLit int32
	var iLit2 int32
	var status int32
	var nWords int32 = Abc_Bit6WordNum(nPats)
	Covered := make([]word, NWORDS)
	Both := make([]word, NWORDS)
	Mask := make([]word, NWORDS)
	var pMask *word
	assert(nWords <= NWORDS, "nWords <= NWORDS")
	// prepare constant pattern
	Abc_TtConst(Mask, nWords, 0)
	for i = 0; i < nPats; i++ {
		Abc_TtSetBit(Mask, i)
	}
	// try to replace each by a cheaper one
	// inlining of macro Vec_IntForEachEntry
	for i = 0; i < Vec_IntSize(vSupp); i++ {
		iLit = Vec_IntEntry(vSupp, i)
		var iDiv int32 = Abc_Lit2Var(iLit) - iFirstDiv
		// collect covered except by this one
		Abc_TtConst(Covered, nWords, 0)
		// inlining of macro Vec_IntForEachEntry
		for k = 0; k < Vec_IntSize(vSupp); k++ {
			iLit2 = Vec_IntEntry(vSupp, k)
			if iLit2 == iLit {
				continue
			}
			pMask = Vec_WrdEntryP(vPats, NWORDS*(Abc_Lit2Var(iLit2)-iFirstDiv))
			Abc_TtOr(Covered, Covered, pMask, nWords)
		}
		// consider any cheaper ones that this one
		for k = 0; k < iDiv; k++ {
			if Vec_IntEntry(vWeights, k) == Vec_IntEntry(vWeights, iDiv) {
				continue
			}
			assert(Vec_IntEntry(vWeights, k) < Vec_IntEntry(vWeights, iDiv), "Vec_IntEntry(vWeights, k) < Vec_IntEntry(vWeights, iDiv)")
			pMask = Vec_WrdEntryP(vPats, NWORDS*k)
			// check if it covers the remaining ones
			Abc_TtOr(Both, Covered, pMask, nWords)
			if !Abc_TtEqual(Both, Mask, nWords) {
				continue
			}
			// try this one
			Vec_IntWriteEntry(vSupp, i, Abc_Var2Lit(iFirstDiv+k, 1))
			status = sat_solver_solve(pSat, Vec_IntArray(vSupp)[:Vec_IntSize(vSupp)], nil, 0, 0, 0, 0)
			if status == l_False { // success

				//printf( "Replacing %d(%d) by %d(%d) with const difference %d.\n",
				//    iDiv, Vec_IntEntry(vWeights, iDiv), k, Vec_IntEntry(vWeights, k),
				//    Vec_IntEntry(vWeights, iDiv) - Vec_IntEntry(vWeights, k) );
				break
			}
			Vec_IntWriteEntry(vSupp, i, iLit)
		}
	}
}

func Acb_FindSupport(pSat *sat_solver, iFirstDiv int32, vWeights *Vec_Int_t, vSuppStart *Vec_Int_t, TimeOut int32) *Vec_Int_t {
	var clkLimit abctime = int64(TimeOut)*int64(time.Second) + Abc_Clock()
	var vPats *Vec_Wrd_t = nil
	var nPats int32
	var vSuppBest, vSupp, vTemp *Vec_Int_t
	var CostBest, Cost int32
	var Iter int32

	// find initial best
	CostBest = Acb_ComputeSuppCost(vSuppStart, vWeights, iFirstDiv)
	vSuppBest = Vec_IntDup(vSuppStart)
	fmt.Printf("Starting cost = %d.\n", CostBest)

	// iteratively find the one with the most ones in the uncovered rows
	for Iter = 0; Iter < 500; Iter++ {
		if Abc_Clock() > clkLimit {
			fmt.Printf("Timeout after %d sec.\n", TimeOut)
			break
		}
		if Iter == 0 {
			vSupp = Acb_FindSupportStart(pSat, iFirstDiv, vWeights, &vPats, &nPats)
		} else {
			vSupp = Acb_FindSupportNext(pSat, iFirstDiv, vWeights, vPats, &nPats)
		}
		if vSupp == nil {
			break
		}
		vTemp = vSupp
		vSupp = Acb_FindSupportMin(pSat, iFirstDiv, vPats, &nPats, vTemp)
		Vec_IntFree(vTemp)
		if vSupp == nil {
			break
		}
		Cost = Acb_ComputeSuppCost(vSupp, vWeights, iFirstDiv)
		//Acb_PrintPatterns( vPats, nPats, vWeights );
		if CostBest > Cost {
			CostBest = Cost
			// inlining of macro ABC_SWAP
			{
				var t *Vec_Int_t = vSuppBest
				vSuppBest = vSupp
				vSupp = t
			}
			fmt.Printf("Iter %4d:  Next cost = %5d.  ", Iter, Cost)
			fmt.Printf("Updating best solution.\n")
		}
		Vec_IntFree(vSupp)
	}
	if vPats != nil {
		Acb_FindReplace(pSat, iFirstDiv, vWeights, vPats, nPats, vSuppBest)
	}
	//printf( "Number of patterns = %d.\n", nPats );
	Vec_WrdFreeP(&vPats)
	return vSuppBest
}

/*
*Function*************************************************************

	Synopsis    [Compute the best support of the targets.]

	Description []

	SideEffects []

	SeeAlso     []

**********************************************************************
*/
func Acb_DerivePatchSupport(pCnf *Cnf_Dat_t, iTar int32, nTargets int32, nCoDivs int32, vDivs *Vec_Int_t, pNtkF *Acb_Ntk_t, vSuppOld *Vec_Int_t, TimeOut int32) *Vec_Int_t {
	var vSupp *Vec_Int_t = Vec_IntAlloc(100)
	var i, Lit int32
	var iCoVarBeg int32 = 1
	var iCiVarBeg int32 = Cnf_DataVarNum(pCnf) - nTargets
	var pSat *sat_solver = sat_solver_new()
	sat_solver_setnvars(pSat, 2*Cnf_DataVarNum(pCnf)+nCoDivs)
	// add clauses
	for i = 0; i < Cnf_DataClauseNum(pCnf); i++ {
		if !sat_solver_addclause(pSat, Cnf_DataClause(pCnf, i)) {
			return nil
		}
	}
	// add output clause
	Lit = Abc_Var2Lit(iCoVarBeg, 0)
	if !sat_solver_addclause(pSat, []int32{Lit}) {
		return nil
	}
	// add clauses
	Cnf_DataSetManNil(pCnf)
	Cnf_DataLift(pCnf, Cnf_DataVarNum(pCnf))
	for i = 0; i < Cnf_DataClauseNum(pCnf); i++ {
		if !sat_solver_addclause(pSat, Cnf_DataClause(pCnf, i)) {
			return nil
		}
	}
	Cnf_DataLift(pCnf, -Cnf_DataVarNum(pCnf))
	// add output clause
	Lit = Abc_Var2Lit(iCoVarBeg+Cnf_DataVarNum(pCnf), 0)
	if !sat_solver_addclause(pSat, []int32{Lit}) {
		return nil
	}
	// create XORs for targets
	// add negative literal
	Lit = Abc_Var2Lit(iCiVarBeg+iTar, 1)
	if !sat_solver_addclause(pSat, []int32{Lit}) {
		return nil
	}
	// add positive literal
	Lit = Abc_Var2Lit(iCiVarBeg+Cnf_DataVarNum(pCnf)+iTar, 0)
	if !sat_solver_addclause(pSat, []int32{Lit}) {
		return nil
	}
	// create XORs for divisors
	if nCoDivs > 0 {
		var clk abctime = Abc_Clock()
		var fUseMinAssump int32 = 1
		var iLit, nSuppNew, status int32
		var iDivVar int32 = 2 * Cnf_DataVarNum(pCnf)
		var pLits [2]int32
		var j, iDivOld int32
		Vec_IntClear(vSupp)
		if vSuppOld != nil {
			// start with predefined support
			// inlining of macro Vec_IntForEachEntry
			for j = 0; j < Vec_IntSize(vSuppOld); j++ {
				iDivOld = Vec_IntEntry(vSuppOld, j)
				var iVar0 int32 = iCoVarBeg + 1 + iDivOld
				var iVar1 int32 = iCoVarBeg + 1 + iDivOld + Cnf_DataVarNum(pCnf)
				//printf( "Selecting predefined divisor %d with weight %d\n",
				//    iDivOld, Vec_IntEntry(&pNtkF->vObjWeight, Vec_IntEntry(vDivs, iDivOld)) );
				// add equality clauses
				pLits[0] = Abc_Var2Lit(iVar0, 0)
				pLits[1] = Abc_Var2Lit(iVar1, 1)
				if !sat_solver_addclause(pSat, pLits[:]) {
					fmt.Printf("Unsat is detected earlier.\n")
					status = l_False
					break
				}
				pLits[0] = Abc_Var2Lit(iVar0, 1)
				pLits[1] = Abc_Var2Lit(iVar1, 0)
				if !sat_solver_addclause(pSat, pLits[:]) {
					fmt.Printf("Unsat is detected earlier.\n")
					status = l_False
					break
				}
			}
		}
		if vSuppOld == nil || j == Vec_IntSize(vSuppOld) {
			for i = 0; i < nCoDivs; i++ {
				sat_solver_add_xor(pSat, iDivVar+i, iCoVarBeg+1+i, iCoVarBeg+1+i+Cnf_DataVarNum(pCnf), 0)
				Vec_IntPush(vSupp, Abc_Var2Lit(iDivVar+i, 1))
			}
			// try one run
			if TimeOut != 0 {
				sat_solver_set_runtime_limit(pSat, int64(TimeOut)*int64(time.Second)+Abc_Clock())
			}
			status = sat_solver_solve(pSat, Vec_IntArray(vSupp)[:Vec_IntSize(vSupp)], nil, 0, 0, 0, 0)
			if TimeOut != 0 {
				sat_solver_set_runtime_limit(pSat, 0)
			}
			if status == l_True {
				fmt.Printf("ECO does not exist.\n")
				sat_solver_delete(pSat)
				Vec_IntFree(vSupp)
				return nil
			}
			if status == l_Undef {
				fmt.Printf("Support computation timed out after %d sec.\n", TimeOut)
				sat_solver_delete(pSat)
				Vec_IntFree(vSupp)
				return nil
			}
			assert(status == l_False, "status == l_False")
			fmt.Printf("Proved that the problem has a solution.  ")
			Abc_PrintTime(1, "Time", Abc_Clock()-clk)
			// find minimum subset
			if fUseMinAssump != 0 {
				var fUseSuppMin int32 = 1
				// solve in a standard way
				var clk abctime = Abc_Clock()
				nSuppNew = sat_solver_minimize_assumptions(pSat, Vec_IntArray(vSupp), Vec_IntSize(vSupp), 0)
				Vec_IntShrink(vSupp, nSuppNew)
				Vec_IntSort(vSupp, 0)
				fmt.Printf("Found one feasible set of %d divisors.  ", Vec_IntSize(vSupp))
				Abc_PrintTime(1, "Time", Abc_Clock()-clk)

				// perform minimization
				if (fUseSuppMin != 0) && Vec_IntSize(vSupp) > 0 {
					var clk abctime = Abc_Clock()
					var vSupp2 *Vec_Int_t = Vec_IntDup(vSupp)
					var vTemp *Vec_Int_t
					var vWeights *Vec_Int_t = Acb_DeriveWeights(vDivs, pNtkF)
					vTemp = vSupp
					vSupp = Acb_FindSupport(pSat, iDivVar, vWeights, vTemp, TimeOut)
					Vec_IntFree(vWeights)
					Vec_IntFree(vTemp)
					if vSupp == nil {
						fmt.Printf("Support minimization did not succeed.  ")
						//sat_solver_delete( pSat );
						vSupp = vSupp2

					} else {
						Vec_IntFree(vSupp2)
						fmt.Printf("Minimized support to %d supp vars.  ", Vec_IntSize(vSupp))
					}
					Abc_PrintTime(1, "Time", Abc_Clock()-clk)
				}

			} else {
				var pFinal []int32
				var nFinal int32 = sat_solver_final(pSat, &pFinal)
				fmt.Printf("AnalyzeFinal returned %d (out of %d).\n", nFinal, Vec_IntSize(vSupp))
				Vec_IntClear(vSupp)
				for i = 0; i < nFinal; i++ {
					Vec_IntPush(vSupp, Abc_LitNot(pFinal[i]))
				}
				// try one run
				status = sat_solver_solve(pSat, Vec_IntArray(vSupp)[:Vec_IntSize(vSupp)], nil, 0, 0, 0, 0)
				assert(status == l_False, "status == l_False")
				// try again
				nFinal = sat_solver_final(pSat, &pFinal)
				fmt.Printf("AnalyzeFinal returned %d (out of %d).\n", nFinal, Vec_IntSize(vSupp))
			}
			// remap them into numbers
			// inlining of macro Vec_IntForEachEntry
			for i = 0; i < Vec_IntSize(vSupp); i++ {
				iLit = Vec_IntEntry(vSupp, i)
				Vec_IntWriteEntry(vSupp, i, Abc_Lit2Var(iLit)-iDivVar)
			}
			Vec_IntSort(vSupp, 0)
		}
	}
	sat_solver_delete(pSat)
	if vSupp != nil {
		Vec_IntSort(vSupp, 0)
	}
	return vSupp
}

func satoko_add_xor(pSat *satoko_t, iVarA int32, iVarB int32, iVarC int32, fCompl int32) int32 {
	var Lits [3]int32
	var Cid int32
	assert(iVarA >= 0 && iVarB >= 0 && iVarC >= 0, "iVarA >= 0 && iVarB >= 0 && iVarC >= 0")

	Lits[0] = toLitCond(iVarA, fCompl == 0)
	Lits[1] = toLitCond(iVarB, 1)
	Lits[2] = toLitCond(iVarC, 1)
	Cid = satoko_add_clause(pSat, Lits, 3)
	assert((Cid != 0), "Cid")

	Lits[0] = toLitCond(iVarA, fCompl == 0)
	Lits[1] = toLitCond(iVarB, 0)
	Lits[2] = toLitCond(iVarC, 0)
	Cid = satoko_add_clause(pSat, Lits, 3)
	assert((Cid != 0), "Cid")

	Lits[0] = toLitCond(iVarA, fCompl)
	Lits[1] = toLitCond(iVarB, 1)
	Lits[2] = toLitCond(iVarC, 0)
	Cid = satoko_add_clause(pSat, Lits, 3)
	assert((Cid != 0), "Cid")

	Lits[0] = toLitCond(iVarA, fCompl)
	Lits[1] = toLitCond(iVarB, 0)
	Lits[2] = toLitCond(iVarC, 1)
	Cid = satoko_add_clause(pSat, Lits, 3)
	assert((Cid != 0), "Cid")
	return 4
}
func Acb_DerivePatchSupportS(pCnf *Cnf_Dat_t, nCiTars int32, nCoDivs int32, vDivs *Vec_Int_t, pNtkF *Acb_Ntk_t, vSuppOld *Vec_Int_t, TimeOut int32) *Vec_Int_t {
	var vSupp *Vec_Int_t = Vec_IntAlloc(100)
	var i, Lit int32
	var iCoVarBeg int32 = 1
	var iCiVarBeg int32 = Cnf_DataVarNum(pCnf) - nCiTars
	var pSat *satoko_t = satoko_create()
	satoko_setnvars(pSat, 2*Cnf_DataVarNum(pCnf)+nCiTars+nCoDivs)
	satoko_options(pSat).NoSimplify = true
	// add clauses
	for i = 0; i < Cnf_DataClauseNum(pCnf); i++ {
		if satoko_add_clause(pSat, Cnf_DataClause(pCnf, i)) == 0 {
			return nil
		}
	}
	// add output clause
	Lit = Abc_Var2Lit(iCoVarBeg, 0)
	if satoko_add_clause(pSat, &Lit, 1) == 0 {
		return nil
	}
	// add clauses
	Cnf_DataSetManNil(pCnf)
	Cnf_DataLift(pCnf, Cnf_DataVarNum(pCnf))
	for i = 0; i < Cnf_DataClauseNum(pCnf); i++ {
		if satoko_add_clause(pSat, Cnf_DataClause(pCnf, i)) == 0 {
			return nil
		}
	}
	Cnf_DataLift(pCnf, -Cnf_DataVarNum(pCnf))
	// add output clause
	Lit = Abc_Var2Lit(iCoVarBeg+Cnf_DataVarNum(pCnf), 0)
	if satoko_add_clause(pSat, &Lit, 1) == 0 {
		return nil
	}
	// create XORs for targets
	if nCiTars > 0 {
		/*
		   int iXorVar = 2 * pCnf->nVars;
		   int Lit;
		   Vec_IntClear( vSupp );
		   for ( i = 0; i < nCiTars; i++ )
		   {
		       satoko_add_xor( pSat, iXorVar+i, iCiVarBeg+i, iCiVarBeg+i+pCnf->nVars, 0 );
		       Vec_IntPush( vSupp, Abc_Var2Lit(iXorVar+i, 0) );
		   }
		   if ( satoko_add_clause( pSat, Vec_IntArray(vSupp), Vec_IntSize(vSupp) ) )
		       return NULL;
		*/
		// add negative literal
		Lit = Abc_Var2Lit(iCiVarBeg, 1)
		if satoko_add_clause(pSat, &Lit, 1) == 0 {
			return nil
		}
		// add positive literal
		Lit = Abc_Var2Lit(iCiVarBeg+Cnf_DataVarNum(pCnf), 0)
		if satoko_add_clause(pSat, &Lit, 1) == 0 {
			return nil
		}
	}
	// create XORs for divisors
	if nCoDivs > 0 {
		var clk abctime = Abc_Clock()
		var fUseMinAssump int32 = 1
		var iLit, status, nSuppNew int32
		var iDivVar int32 = 2*Cnf_DataVarNum(pCnf) + nCiTars
		Vec_IntClear(vSupp)
		for i = 0; i < nCoDivs; i++ {
			satoko_add_xor(pSat, iDivVar+i, iCoVarBeg+1+i, iCoVarBeg+1+i+Cnf_DataVarNum(pCnf), 0)
			Vec_IntPush(vSupp, Abc_Var2Lit(iDivVar+i, 1))
			/*
			   pLits[0] = Abc_Var2Lit(iCoVarBeg+1+i,             1);
			   pLits[1] = Abc_Var2Lit(iCoVarBeg+1+i+pCnf->nVars, 0);
			   pLits[2] = Abc_Var2Lit(iDivVar+i,                 1);
			   if ( satoko_add_clause( pSat, pLits, 3 ) )
			       assert( 0 );

			   pLits[0] = Abc_Var2Lit(iCoVarBeg+1+i,             0);
			   pLits[1] = Abc_Var2Lit(iCoVarBeg+1+i+pCnf->nVars, 1);
			   pLits[2] = Abc_Var2Lit(iDivVar+i,                 1);
			   if ( satoko_add_clause( pSat, pLits, 3 ) )
			       assert( 0 );

			   Vec_IntPush( vSupp, Abc_Var2Lit(iDivVar+i, 0) );
			*/
		}

		// try one run
		status = satoko_solve_assumptions(pSat, Vec_IntArray(vSupp), Vec_IntSize(vSupp))
		if status != l_False {
			fmt.Printf("Demonstrated that the problem has NO solution.  ")
			Abc_PrintTime(1, "Time", Abc_Clock()-clk)
			satoko_destroy(pSat)
			Vec_IntFree(vSupp)
			return nil
		}
		assert(status == l_False, "status == l_False")
		fmt.Printf("Proved that the problem has a solution.  ")
		Abc_PrintTime(1, "Time", Abc_Clock()-clk)

		// find minimum subset
		if fUseMinAssump != 0 {
			var clk abctime = Abc_Clock()
			nSuppNew = satoko_minimize_assumptions(pSat, Vec_IntArray(vSupp), Vec_IntSize(vSupp), 0)
			Vec_IntShrink(vSupp, nSuppNew)
			fmt.Printf("Solved the problem with %d supp vars.  ", Vec_IntSize(vSupp))
			Abc_PrintTime(1, "Time", Abc_Clock()-clk)

		} else {
			var pFinal []int32
			var nFinal int32 = satoko_final_conflict(pSat, &pFinal)
			fmt.Printf("AnalyzeFinal returned %d (out of %d).\n", nFinal, Vec_IntSize(vSupp))
			Vec_IntClear(vSupp)
			for i = 0; i < nFinal; i++ {
				Vec_IntPush(vSupp, Abc_LitNot(pFinal[i]))
			}
			// try one run
			status = satoko_solve_assumptions(pSat, Vec_IntArray(vSupp), Vec_IntSize(vSupp))
			assert(status == l_False, "status == l_False")
			// try again
			nFinal = satoko_final_conflict(pSat, &pFinal)
			fmt.Printf("AnalyzeFinal returned %d (out of %d).\n", nFinal, Vec_IntSize(vSupp))
		}

		// remap them into numbers
		// inlining of macro Vec_IntForEachEntry
		for i = 0; i < Vec_IntSize(vSupp); i++ {
			iLit = Vec_IntEntry(vSupp, i)
			Vec_IntWriteEntry(vSupp, i, Abc_Lit2Var(iLit)-iDivVar)
		}
		Vec_IntSort(vSupp, 0)
		//Vec_IntPrint( vSupp );
	}
	satoko_destroy(pSat)
	Vec_IntSort(vSupp, 0)
	return vSupp
}

/*
*Function*************************************************************

	Synopsis    [Compute functions of the targets.]

	Description []

	SideEffects []

	SeeAlso     []

**********************************************************************
*/
func Acb_EnumerateSatAssigns(pSat *sat_solver, PivotVar int32, FreeVar int32, vDivVars *Vec_Int_t, vTempLits *Vec_Int_t, vTempSop *Vec_Str_t) string {
	var fCreatePrime int32 = 1
	var status int32
	var i int32
	var iMint int32
	var iVar int32
	var iLit int32
	var nFinal int32
	var pFinal []int32
	var pLits [2]int32
	var vTemp, vLits *Vec_Int_t
	assert(FreeVar < sat_solver_nvars(pSat), "FreeVar < sat_solver_nvars(pSat)")
	pLits[0] = Abc_Var2Lit(PivotVar, 1) // F = 1
	pLits[1] = Abc_Var2Lit(FreeVar, 0)  // iNewLit
	Vec_StrClear(vTempSop)
	Vec_StrGrow(vTempSop, 8*(Vec_IntSize(vDivVars)+3)+1)
	// check constant 0
	status = sat_solver_solve(pSat, pLits[:], nil, 0, 0, 0, 0)
	if status == l_False {
		Vec_StrPush(vTempSop, ' ')
		Vec_StrPush(vTempSop, '0')
		Vec_StrPush(vTempSop, '\n')
		Vec_StrPush(vTempSop, 0)
		return strings.TrimRight(string(Vec_StrReleaseArray(vTempSop)), "\x00")
	}
	// check constant 1
	pLits[0] = Abc_LitNot(pLits[0])
	status = sat_solver_solve(pSat, pLits[:], nil, 0, 0, 0, 0)
	pLits[0] = Abc_LitNot(pLits[0])
	if status == l_False || Vec_IntSize(vDivVars) == 0 {
		Vec_StrPush(vTempSop, ' ')
		Vec_StrPush(vTempSop, '1')
		Vec_StrPush(vTempSop, '\n')
		Vec_StrPush(vTempSop, 0)
		return strings.TrimRight(string(Vec_StrReleaseArray(vTempSop)), "\x00")
	}
	//Vec_IntPrint( vDivVars );
	vTemp = Vec_IntAlloc(100)
	vLits = Vec_IntAlloc(100)
	for iMint = 0; ; iMint++ {
		if iMint == 1000 {
			if Vec_IntSize(vDivVars) == 0 {
				fmt.Printf("Assuming constant 0 function.\n")
				Vec_StrClear(vTempSop)
				Vec_StrPush(vTempSop, ' ')
				Vec_StrPush(vTempSop, '0')
				Vec_StrPush(vTempSop, '\n')
				Vec_StrPush(vTempSop, 0)
				return strings.TrimRight(string(Vec_StrReleaseArray(vTempSop)), "\x00")
			}

			fmt.Printf("Reached the limit on the number of cubes (1000).\n")
			Vec_IntFree(vTemp)
			Vec_IntFree(vLits)
			return ""
		}
		//int Offset = Vec_StrSize(vTempSop);
		// find onset minterm
		status = sat_solver_solve(pSat, pLits[:], nil, 0, 0, 0, 0)
		if status == l_False {
			fmt.Printf("Finished enumerating %d cubes.\n", iMint)
			Vec_IntFree(vTemp)
			Vec_IntFree(vLits)
			Vec_StrPush(vTempSop, 0)
			return strings.TrimRight(string(Vec_StrReleaseArray(vTempSop)), "\x00")
		}
		assert(status == l_True, "status == l_True")
		// collect divisor literals
		Vec_IntClear(vTempLits)
		Vec_IntPush(vTempLits, Abc_LitNot(pLits[0])) // F = 0
		//printf( "%8d %3d  ", 0, 0 );
		//        Vec_IntForEachEntry( vDivVars, iVar, i )
		// inlining of macro Vec_IntForEachEntryReverse
		for i = Vec_IntSize(vDivVars) - 1; i >= 0; i-- {
			iVar = Vec_IntEntry(vDivVars, i)
			Vec_IntPush(vTempLits, sat_solver_var_literal(pSat, iVar))
			//printf( "%c", '0' + sat_solver_var_value(pSat, iVar) );
		}
		//printf( "\n" );
		// create new cube
		for i = 0; i < Vec_IntSize(vDivVars); i++ {
			Vec_StrPush(vTempSop, '-')
		}
		if fCreatePrime != 0 {
			// expand against offset
			status = sat_solver_push(pSat, Vec_IntEntry(vTempLits, 0))
			assert(status == 1, "status == 1")
			nFinal = sat_solver_minimize_assumptions(pSat, Vec_IntArray(vTempLits)[1:], Vec_IntSize(vTempLits)-1, 0)
			Vec_IntShrink(vTempLits, nFinal+1)
			sat_solver_pop(pSat)
			// compute cube and add clause
			Vec_IntWriteEntry(vTempLits, 0, Abc_LitNot(pLits[1])) // NOT(iNewLit)
			// inlining of macro Vec_IntForEachEntryStart
			for i = 1; i < Vec_IntSize(vTempLits); i++ {
				iLit = Vec_IntEntry(vTempLits, i)
				Vec_IntWriteEntry(vTempLits, i, Abc_LitNot(iLit))
				iVar = Vec_IntFind(vDivVars, Abc_Lit2Var(iLit))
				assert(iVar >= 0, "iVar >= 0")
				//uCube &= Abc_LitIsCompl(pFinal[i]) ? s_Truths6[iVar] : ~s_Truths6[iVar];
				Vec_StrWriteEntry(vTempSop, Vec_StrSize(vTempSop)-Vec_IntSize(vDivVars)+iVar, byte('0'+(1-Abc_LitIsCompl(iLit))))
			}

		} else {
			// expand against offset
			status = sat_solver_solve(pSat, Vec_IntArray(vTempLits)[:Vec_IntSize(vTempLits)], nil, 0, 0, 0, 0)
			if status != l_False {
				fmt.Printf("Selected onset minterm number %d belongs to the offset (this is a bug).\n", iMint)
			}
			assert(status == l_False, "status == l_False")

			// compute cube and add clause
			nFinal = sat_solver_final(pSat, &pFinal)
			Vec_IntSelectSort(pFinal, nFinal)
			/*
			   // pretend that this is final
			   veci_resize(&pSat->conf_final,0);
			   Vec_IntForEachEntry( vTempLits, iLit, i )
			       veci_push(&pSat->conf_final, lit_neg(iLit));
			   pFinal = pSat->conf_final.ptr;
			   nFinal = Vec_IntSize(vTempLits);
			*/
			////////////////////////////////////////////////////////
			// create new cube
			Vec_IntClear(vTemp)
			Vec_IntPush(vTemp, Abc_LitNot(pLits[0])) // F = 0
			for i = 0; i < nFinal; i++ {
				if pFinal[i] == pLits[0] {
					continue
				}
				Vec_IntPush(vTemp, Abc_LitNot(pFinal[i]))
			}

			//Vec_IntPrint( vTemp );
			// try removing each one starting with the last one
			//printf( "Started with %d lits   ", nFinal-1 );
			for i = nFinal - 1; i > 0; i-- {
				var iLit int32 = Vec_IntEntry(vTemp, i)
				Vec_IntDrop(vTemp, i)
				// try SAT
				status = sat_solver_solve(pSat, Vec_IntArray(vTemp)[:Vec_IntSize(vTemp)], nil, 0, 0, 0, 0)
				if status == l_False {
					//    printf( "U" );
					continue
				}
				//if ( status == l_True )
				//    printf( "S" );
				//else if ( status == l_Undef )
				//    printf( "T" );
				Vec_IntInsert(vTemp, i, iLit)
			}
			//printf( "   Ended up with %d lits\n", Vec_IntSize(vTemp)-1 );
			//Vec_IntPrint( vTemp );

			// inlining of macro Vec_IntForEachEntry
			for i = 0; i < Vec_IntSize(vTemp); i++ {
				iLit = Vec_IntEntry(vTemp, i)
				pFinal[i] = Abc_LitNot(iLit)
			}
			nFinal = Vec_IntSize(vTemp)
			////////////////////////////////////////////////////////

			Vec_IntClear(vTempLits)
			Vec_IntPush(vTempLits, Abc_LitNot(pLits[1])) // NOT(iNewLit)
			for i = 0; i < nFinal; i++ {
				if pFinal[i] == pLits[0] {
					continue
				}
				Vec_IntPush(vTempLits, pFinal[i])
				iVar = Vec_IntFind(vDivVars, Abc_Lit2Var(pFinal[i]))
				assert(iVar >= 0, "iVar >= 0")
				//uCube &= Abc_LitIsCompl(pFinal[i]) ? s_Truths6[iVar] : ~s_Truths6[iVar];
				Vec_StrWriteEntry(vTempSop, Vec_StrSize(vTempSop)-Vec_IntSize(vDivVars)+iVar, byte(('0' + Abc_LitIsCompl(pFinal[i]))))
			}
		}
		//printf( "%6d : %8d %3d  ", iMint, (int)pSat->stats.conflicts, Vec_IntSize(vTempLits)-1 );

		Vec_StrAppend(vTempSop, " 1\n")
		status = boolInt(sat_solver_addclause(pSat, Vec_IntArray(vTempLits)[:Vec_IntSize(vTempLits)]), 1, 0)
		assert((status != 0), "status")

		//Vec_StrPush( vTempSop, '\0' );
		//printf( "%s", Vec_StrEntryP(vTempSop, Offset) );
		//Vec_StrPop( vTempSop );
	}
	assert(false, "0")
	return ""
}
func Acb_DeriveOnePatchFunction(pCnf *Cnf_Dat_t, iTar int32, nTargets int32, nCoDivs int32, vUsed *Vec_Int_t, fCisOnly int32) string {
	var pSop string
	var vTempLits *Vec_Int_t = Vec_IntAlloc(Vec_IntSize(vUsed) + 1)
	var vTempSop *Vec_Str_t = Vec_StrAlloc(0)
	var i, Lit int32
	var iCiVarBeg int32 = Cnf_DataVarNum(pCnf) - nTargets - Vec_IntSize(vUsed)
	var iCoVarBeg int32 = 1
	var Index int32
	var pSat *sat_solver = sat_solver_new()
	sat_solver_setnvars(pSat, Cnf_DataVarNum(pCnf)+1)
	// add clauses
	for i = 0; i < Cnf_DataClauseNum(pCnf); i++ {
		if !sat_solver_addclause(pSat, Cnf_DataClause(pCnf, i)) {
			return ""
		}
	}
	// add output clause
	Lit = Abc_Var2Lit(iCoVarBeg, 0)
	if !sat_solver_addclause(pSat, []int32{Lit}) {
		return ""
	}
	// remap vUsed to be in terms of divisor variables
	if fCisOnly != 0 {
		// inlining of macro Vec_IntForEachEntry
		for i = 0; i < Vec_IntSize(vUsed); i++ {
			Index = Vec_IntEntry(vUsed, i)
			Vec_IntWriteEntry(vUsed, i, iCiVarBeg+Index)
		}

	} else {
		// inlining of macro Vec_IntForEachEntry
		for i = 0; i < Vec_IntSize(vUsed); i++ {
			Index = Vec_IntEntry(vUsed, i)
			Vec_IntWriteEntry(vUsed, i, iCoVarBeg+1+Index)
		}
	}
	// enumerate assignments for each target in terms of used divisors
	pSop = Acb_EnumerateSatAssigns(pSat, Cnf_DataVarNum(pCnf)-nTargets+iTar, Cnf_DataVarNum(pCnf), vUsed, vTempLits, vTempSop)
	Vec_IntFree(vTempLits)
	Vec_StrFree(vTempSop)
	sat_solver_delete(pSat)
	if pSop == "" {
		return ""
	}
	//printf( "Function %d:\n%s", i, pSop );
	// remap vUsed to be in terms of original divisors
	if fCisOnly != 0 {
		// inlining of macro Vec_IntForEachEntry
		for i = 0; i < Vec_IntSize(vUsed); i++ {
			Index = Vec_IntEntry(vUsed, i)
			Vec_IntWriteEntry(vUsed, i, Index-iCiVarBeg)
		}

	} else {
		// inlining of macro Vec_IntForEachEntry
		for i = 0; i < Vec_IntSize(vUsed); i++ {
			Index = Vec_IntEntry(vUsed, i)
			Vec_IntWriteEntry(vUsed, i, Index-(iCoVarBeg+1))
		}
	}
	return pSop
}
func Acb_CheckMiter(pCnf *Cnf_Dat_t) int32 {
	var iCoVarBeg int32 = 1
	var i int32
	var Lit int32
	var status int32
	var pSat *sat_solver = sat_solver_new()
	sat_solver_setnvars(pSat, Cnf_DataVarNum(pCnf))
	// add clauses
	for i = 0; i < Cnf_DataClauseNum(pCnf); i++ {
		if !sat_solver_addclause(pSat, Cnf_DataClause(pCnf, i)) {
			return 1
		}
	}
	// add output clause
	Lit = Abc_Var2Lit(iCoVarBeg, 0)
	if !sat_solver_addclause(pSat, []int32{Lit}) {
		return 1
	}
	status = sat_solver_solve(pSat, nil, nil, 0, 0, 0, 0)
	sat_solver_delete(pSat)
	return boolInt(status == l_False, 1, 0)
}

/*
*Function*************************************************************

	Synopsis    [Update miter by substituting the last target by a given function.]

	Description []

	SideEffects []

	SeeAlso     []

**********************************************************************
*/
func Acb_CollectIntNodes_rec(p *Gia_Man_t, pObj *Gia_Obj_t, vNodes *Vec_Int_t) {
	if Gia_ObjIsTravIdCurrent(p, pObj) {
		return
	}
	Gia_ObjSetTravIdCurrent(p, pObj)
	assert(Gia_ObjIsAnd(pObj) != 0, "Gia_ObjIsAnd(pObj)")
	Acb_CollectIntNodes_rec(p, Gia_ObjFanin0(pObj), vNodes)
	Acb_CollectIntNodes_rec(p, Gia_ObjFanin1(pObj), vNodes)
	Vec_IntPush(vNodes, Gia_ObjId(p, pObj))
}
func Acb_CollectIntNodes(p *Gia_Man_t, vNodes0 *Vec_Int_t, vNodes1 *Vec_Int_t) {
	var pObj *Gia_Obj_t
	var i int32
	Vec_IntClear(vNodes0)
	Vec_IntClear(vNodes1)
	Gia_ManIncrementTravId(p)
	Gia_ObjSetTravIdCurrent(p, Gia_ManConst0(p))
	// inlining of macro Gia_ManForEachCi
	for i = 0; i < Gia_ManCiNum(p); i++ {
		pObj = Gia_ManCi(p, i)
		Gia_ObjSetTravIdCurrent(p, pObj)
	}
	// inlining of macro Gia_ManForEachCo
	for i = 0; i < Gia_ManCoNum(p); i++ {
		pObj = Gia_ManCo(p, i)
		if i > 0 {
			Acb_CollectIntNodes_rec(p, Gia_ObjFanin0(pObj), vNodes1)
		}
	}
	// inlining of macro Gia_ManForEachCo
	for i = 0; i < Gia_ManCoNum(p); i++ {
		pObj = Gia_ManCo(p, i)
		if i == 0 {
			Acb_CollectIntNodes_rec(p, Gia_ObjFanin0(pObj), vNodes0)
		}
	}
}
func Acb_UpdateMiter(pM *Gia_Man_t, pOne *Gia_Man_t, iTar int32, nTargets int32, vUsed *Vec_Int_t, fCisOnly int32) *Gia_Man_t {
	var pRes, pTemp *Gia_Man_t
	var pObj *Gia_Obj_t
	var i int32
	var vNodes0 *Vec_Int_t = Vec_IntAlloc(Gia_ManAndNum(pM))
	var vNodes1 *Vec_Int_t = Vec_IntAlloc(Gia_ManAndNum(pM))
	Acb_CollectIntNodes(pM, vNodes0, vNodes1)
	Gia_ManFillValue(pM)
	Gia_ManFillValue(pOne)
	// create new
	pRes = Gia_ManStart(Gia_ManObjNum(pM) + Gia_ManObjNum(pOne))
	Gia_ManHashAlloc(pRes)
	Gia_ManConst0(pM).Value = 0
	Gia_ManConst0(pOne).Value = 0
	// copy first part of the miter
	// inlining of macro Gia_ManForEachCi
	for i = 0; i < Gia_ManCiNum(pM); i++ {
		pObj = Gia_ManCi(pM, i)
		//        if ( i < Gia_ManCiNum(pM) - 1 )
		pObj.Value = uint32(Gia_ManAppendCi(pRes))
	}
	// inlining of macro Gia_ManForEachObjVec
	for i = 0; i < Vec_IntSize(vNodes1); i++ {
		pObj = Gia_ManObj(pM, Vec_IntEntry(vNodes1, i))
		pObj.Value = uint32(Gia_ManHashAnd(pRes, Gia_ObjFanin0Copy(pObj), Gia_ObjFanin1Copy(pObj)))
	}
	// inlining of macro Gia_ManForEachCo
	for i = 0; i < Gia_ManCoNum(pM); i++ {
		pObj = Gia_ManCo(pM, i)
		if i > 0 {
			pObj.Value = uint32(Gia_ObjFanin0Copy(pObj))
		}
	}
	// transfer to New
	//assert( Gia_ManCiNum(pOne) <= Vec_IntSize(vUsed) );
	assert(Gia_ManCoNum(pOne) == 1, "Gia_ManCoNum(pOne) == 1")
	if fCisOnly != 0 {
		// inlining of macro Gia_ManForEachCi
		for i = 0; i < Gia_ManCiNum(pOne); i++ {
			pObj = Gia_ManCi(pOne, i)
			if i < Vec_IntSize(vUsed) {
				pObj.Value = Gia_ManCi(pM, Vec_IntEntry(vUsed, i)).Value
			}
		}

	} else {
		// inlining of macro Gia_ManForEachCi
		for i = 0; i < Gia_ManCiNum(pOne); i++ {
			pObj = Gia_ManCi(pOne, i)
			pObj.Value = Gia_ManCo(pM, 1+Vec_IntEntry(vUsed, i)).Value
		}
	}
	// inlining of macro Gia_ManForEachAnd
	for i = 0; i < Gia_ManObjNum(pOne); i++ {
		pObj = Gia_ManObj(pOne, i)
		if Gia_ObjIsAnd(pObj) == 0 {
		} else {
			pObj.Value = uint32(Gia_ManHashAnd(pRes, Gia_ObjFanin0Copy(pObj), Gia_ObjFanin1Copy(pObj)))
		}
	}
	// transfer to miter
	pObj = Gia_ManCi(pM, Gia_ManCiNum(pM)-nTargets+iTar)
	pObj.Value = uint32(Gia_ObjFanin0Copy(Gia_ManCo(pOne, 0)))
	// inlining of macro Gia_ManForEachObjVec
	for i = 0; i < Vec_IntSize(vNodes0); i++ {
		pObj = Gia_ManObj(pM, Vec_IntEntry(vNodes0, i))
		pObj.Value = uint32(Gia_ManHashAnd(pRes, Gia_ObjFanin0Copy(pObj), Gia_ObjFanin1Copy(pObj)))
	}
	// inlining of macro Gia_ManForEachCo
	for i = 0; i < Gia_ManCoNum(pM); i++ {
		pObj = Gia_ManCo(pM, i)
		Gia_ManAppendCo(pRes, Gia_ObjFanin0Copy(pObj))
	}
	// cleanup
	Vec_IntFree(vNodes0)
	Vec_IntFree(vNodes1)
	Gia_ManHashStop(pRes)
	pTemp = pRes
	pRes = Gia_ManCleanup(pRes)
	Gia_ManStop(pTemp)
	assert(Gia_ManCiNum(pRes) == Gia_ManCiNum(pM), "Gia_ManCiNum(pRes) == Gia_ManCiNum(pM)")
	assert(Gia_ManCoNum(pRes) == Gia_ManCoNum(pM), "Gia_ManCoNum(pRes) == Gia_ManCoNum(pM)")
	return pRes
}

/*
*Function*************************************************************

	Synopsis    [Generate strings representing instance and the patch.]

	Description []

	SideEffects []

	SeeAlso     []

**********************************************************************
*/
func Acb_GenerateInstance(p *Acb_Ntk_t, vDivs *Vec_Int_t, vUsed *Vec_Int_t, vTars *Vec_Int_t) *Vec_Str_t {
	var i, iObj int32
	var vStr *Vec_Str_t = Vec_StrAlloc(100)
	Vec_StrAppend(vStr, "  patch p0 (")
	// inlining of macro Vec_IntForEachEntry
	for i = 0; i < Vec_IntSize(vTars); i++ {
		iObj = Vec_IntEntry(vTars, i)
		prefix := ""
		if i > 0 {
			prefix = ","
		}
		Vec_StrPrintF(vStr, "%s .%s(%s)", prefix, Acb_ObjNameStr(p, iObj), Acb_ObjNameStr(p, iObj))
	}
	// inlining of macro Vec_IntForEachEntryInVec
	for i = 0; i < Vec_IntSize(vUsed); i++ {
		iObj = Vec_IntEntry(vDivs, Vec_IntEntry(vUsed, i))
		Vec_StrPrintF(vStr, ", .%s(%s)", Acb_ObjNameStr(p, iObj), Acb_ObjNameStr(p, iObj))
	}
	Vec_StrAppend(vStr, " );\n\n")
	Vec_StrPush(vStr, 0)
	return vStr
}
func Acb_GenerateSignalNames(p *Acb_Ntk_t, vDivs *Vec_Int_t, vUsed *Vec_Int_t, nNodes int32, vTars *Vec_Int_t, vGates *Vec_Wec_t) *Vec_Ptr_t {
	var vRes *Vec_Ptr_t = Vec_PtrStart(Vec_IntSize(vUsed) + nNodes)
	var vStr *Vec_Str_t = Vec_StrAlloc(1000)
	var i int32
	var iObj int32
	var nWires int32 = 1
	// create input names
	// inlining of macro Vec_IntForEachEntryInVec
	for i = 0; i < Vec_IntSize(vUsed); i++ {
		iObj = Vec_IntEntry(vDivs, Vec_IntEntry(vUsed, i))
		Vec_PtrWriteEntry(vRes, i, Abc_UtilStrsav(Acb_ObjNameStr(p, iObj)))
	}
	// create names for nodes driving outputs
	assert(Vec_WecSize(vGates) == Vec_IntSize(vUsed)+nNodes+Vec_IntSize(vTars), "Vec_WecSize(vGates) == Vec_IntSize(vUsed) + nNodes + Vec_IntSize(vTars)")
	// inlining of macro Vec_IntForEachEntry
	for i = 0; i < Vec_IntSize(vTars); i++ {
		iObj = Vec_IntEntry(vTars, i)
		var vGate *Vec_Int_t = Vec_WecEntry(vGates, Vec_IntSize(vUsed)+nNodes+i)
		assert(Vec_IntEntry(vGate, 0) == ABC_OPER_BIT_BUF, "Vec_IntEntry(vGate, 0) == ABC_OPER_BIT_BUF")
		Vec_PtrWriteEntry(vRes, Vec_IntEntry(vGate, 1), Abc_UtilStrsav(Acb_ObjNameStr(p, iObj)))
	}
	for i = Vec_IntSize(vUsed); i < Vec_IntSize(vUsed)+nNodes; i++ {
		if Vec_PtrEntry(vRes, i) == nil {
			Vec_StrPrintF(vStr, "ww%d", nWires)
			nWires++
			Vec_StrPush(vStr, 0)
			Vec_PtrWriteEntry(vRes, i, Vec_StrReleaseArray(vStr))
		}
	}
	Vec_StrFree(vStr)
	return vRes
}
func Acb_GetUsedDivs(vDivs *Vec_Int_t, vUsed *Vec_Int_t) *Vec_Int_t {
	var i, iObj int32
	var vRes *Vec_Int_t = Vec_IntAlloc(Vec_IntSize(vUsed))
	// inlining of macro Vec_IntForEachEntryInVec
	for i = 0; i < Vec_IntSize(vUsed); i++ {
		iObj = Vec_IntEntry(vDivs, Vec_IntEntry(vUsed, i))
		Vec_IntPush(vRes, iObj)
	}
	return vRes
}
func Acb_SignalNames(p *Acb_Ntk_t, vObjs *Vec_Int_t) *Vec_Ptr_t {
	var vNames *Vec_Ptr_t = Vec_PtrAlloc(Vec_IntSize(vObjs))
	var i, iObj int32
	// inlining of macro Vec_IntForEachEntry
	for i = 0; i < Vec_IntSize(vObjs); i++ {
		iObj = Vec_IntEntry(vObjs, i)
		Vec_PtrPush(vNames, Acb_ObjNameStr(p, iObj))
	}
	return vNames
}
func Acb_GeneratePatch(p *Acb_Ntk_t, vDivs *Vec_Int_t, vUsed *Vec_Int_t, vSops *Vec_Ptr_t, vGias *Vec_Ptr_t, vTars *Vec_Int_t) *Vec_Str_t {
	var vGates *Vec_Wec_t
	if vGias != nil {
		vGates = Abc_GiaSynthesize(vGias, nil)
	} else {
		vGates = Abc_SopSynthesize(vSops)
	}
	var vGate *Vec_Int_t
	var nOuts int32
	if vGias != nil {
		nOuts = Vec_PtrSize(vGias)
	} else {
		nOuts = Vec_PtrSize(vSops)
	}
	var i int32
	var k int32
	var iObj int32
	var nWires int32 = Vec_WecSize(vGates) - Vec_IntSize(vUsed) - nOuts
	var fFirst int32 = 1
	var nGates1 [5]int32
	var nGates0 [5]int32
	var vNames *Vec_Ptr_t = Acb_GenerateSignalNames(p, vDivs, vUsed, nWires, vTars, vGates)
	var vStr *Vec_Str_t = Vec_StrAlloc(100)
	var vSup *Vec_Int_t = Acb_GetUsedDivs(vDivs, vUsed)
	var vSpN *Vec_Ptr_t = Acb_SignalNames(p, vSup)
	var vTfi *Vec_Int_t = Acb_ObjCollectTfiVec(p, vSup)
	var vTfo *Vec_Int_t = Acb_ObjCollectTfoVec(p, vTars)
	var nPiCount int32 = Acb_NtkCountPiBuffers(p, vSup)
	var nPoCount int32 = Acb_NtkCountPoDrivers(p, vTars)
	var nMffc int32
	if Abc_FrameReadSpecName() != "" {
		nMffc = Acb_NtkCollectMfsGates(Abc_FrameReadSpecName(), vSpN, Abc_FrameReadSignalNames(), nGates1[:])
	} else {
		nMffc = 0
	}
	Vec_PtrFree(vSpN)
	Vec_IntFree(vSup)
	// inlining of macro Vec_WecForEachLevelStartStop
	for i = Vec_IntSize(vUsed); i < Vec_IntSize(vUsed)+nWires; i++ {
		vGate = Vec_WecEntry(vGates, i)
		if Vec_IntSize(vGate) > 2 {
			var pName string = Acb_Oper2Name(Vec_IntEntry(vGate, 0))
			if !(int32(strings.Compare(pName, "buf")) != 0) {
				nGates0[2]++
			} else if !(int32(strings.Compare(pName, "not")) != 0) {
				nGates0[3]++
			} else {
				nGates0[4] += Vec_IntSize(vGate) - 3
			}

		} else {
			index := 0
			if Vec_IntEntry(vGate, 0) == ABC_OPER_CONST_T {
				index = 1
			}
			nGates0[index]++
		}
	}

	Vec_StrPrintF(vStr, "// Patch   : in = %d  out = %d : pi_in = %d  po_out = %d : tfi = %d  tfo = %d\n", Vec_IntSize(vUsed), nOuts, nPiCount, nPoCount, Vec_IntSize(vTfi), Vec_IntSize(vTfo))
	Vec_StrPrintF(vStr, "// Added   : gate =%4d : c0 =%2d  c1 =%2d  buf =%3d  inv =%3d  two-input =%4d\n", nWires, nGates0[0], nGates0[1], nGates0[2], nGates0[3], nGates0[4])
	if Abc_FrameReadSpecName() != "" {
		Vec_StrPrintF(vStr, "// Removed : gate =%4d : c0 =%2d  c1 =%2d  buf =%3d  inv =%3d  two-input =%4d\n", nMffc, nGates1[0], nGates1[1], nGates1[2], nGates1[3], nGates1[4])
	}
	if Abc_FrameReadSpecName() != "" {
		Vec_StrPrintF(vStr, "// TOTAL   : gate =%4d : c0 =%2d  c1 =%2d  buf =%3d  inv =%3d  two-input =%4d\n", nWires-nMffc, nGates0[0]-nGates1[0], nGates0[1]-nGates1[1], nGates0[2]-nGates1[2], nGates0[3]-nGates1[3], nGates0[4]-nGates1[4])
	}
	Vec_StrPrintF(vStr, "\n")

	Vec_StrAppend(vStr, "module patch (")

	assert(Vec_IntSize(vTars) == nOuts, "Vec_IntSize(vTars) == nOuts")
	// inlining of macro Vec_IntForEachEntry
	for i = 0; i < Vec_IntSize(vTars); i++ {
		iObj = Vec_IntEntry(vTars, i)
		sep := ""
		if i > 0 {
			sep = ","
		}
		Vec_StrPrintF(vStr, "%s %s", sep, Acb_ObjNameStr(p, iObj))
	}
	// inlining of macro Vec_IntForEachEntryInVec
	for i = 0; i < Vec_IntSize(vUsed); i++ {
		iObj = Vec_IntEntry(vDivs, Vec_IntEntry(vUsed, i))
		Vec_StrPrintF(vStr, ", %s", Acb_ObjNameStr(p, iObj))
	}
	Vec_StrAppend(vStr, " );\n\n")

	Vec_StrAppend(vStr, "  output")
	// inlining of macro Vec_IntForEachEntry
	for i = 0; i < Vec_IntSize(vTars); i++ {
		iObj = Vec_IntEntry(vTars, i)
		sep := ""
		if i > 0 {
			sep = ","
		}
		Vec_StrPrintF(vStr, "%s %s", sep, Acb_ObjNameStr(p, iObj))
	}
	Vec_StrAppend(vStr, ";\n")

	Vec_StrAppend(vStr, "  input")
	// inlining of macro Vec_IntForEachEntryInVec
	for i = 0; i < Vec_IntSize(vUsed); i++ {
		iObj = Vec_IntEntry(vDivs, Vec_IntEntry(vUsed, i))
		sep := ""
		if i > 0 {
			sep = ","
		}
		Vec_StrPrintF(vStr, "%s %s", sep, Acb_ObjNameStr(p, iObj))
	}
	Vec_StrAppend(vStr, ";\n")

	if nWires > nOuts {
		Vec_StrAppend(vStr, "  wire")
		for i = 0; i < nWires; i++ {
			var pName string = Vec_PtrEntry(vNames, Vec_IntSize(vUsed)+i).(string)
			if strncmp(pName, "ww", 2) == 0 {
				sep := ","
				if fFirst != 0 {
					sep = ""
				}
				Vec_StrPrintF(vStr, "%s %s", sep, pName)
				fFirst = 0
			}
		}
		Vec_StrAppend(vStr, ";\n\n")
	}

	// create internal nodes
	// inlining of macro Vec_WecForEachLevelStartStop
	for i = Vec_IntSize(vUsed); i < Vec_IntSize(vUsed)+nWires; i++ {
		vGate = Vec_WecEntry(vGates, i)
		if Vec_IntSize(vGate) > 2 {
			Vec_StrPrintF(vStr, "  %s (", Acb_Oper2Name(Vec_IntEntry(vGate, 0)))
			// inlining of macro Vec_IntForEachEntryStart
			for k = 1; k < Vec_IntSize(vGate); k++ {
				iObj = Vec_IntEntry(vGate, k)
				sep := ""
				if k > 1 {
					sep = ","
				}
				Vec_StrPrintF(vStr, "%s %s", sep, Vec_PtrEntry(vNames, iObj))
			}
			Vec_StrAppend(vStr, " );\n")

		} else {
			assert(Vec_IntEntry(vGate, 0) == ABC_OPER_CONST_F || Vec_IntEntry(vGate, 0) == ABC_OPER_CONST_T, "Vec_IntEntry(vGate, 0) == ABC_OPER_CONST_F || Vec_IntEntry(vGate, 0) == ABC_OPER_CONST_T")
			Vec_StrPrintF(vStr, "  %s (", Acb_Oper2Name(ABC_OPER_BIT_BUF))
			Vec_StrPrintF(vStr, " %s, ", Vec_PtrEntry(vNames, Vec_IntEntry(vGate, 1)))
			bit := 0
			if Vec_IntEntry(vGate, 0) == ABC_OPER_CONST_T {
				bit = 1
			}
			Vec_StrPrintF(vStr, " 1'b%d", bit)
			Vec_StrPrintF(vStr, " );\n")
		}
	}
	Vec_StrAppend(vStr, "\nendmodule\n\n")
	Vec_StrPush(vStr, 0)
	Vec_PtrFreeFree(vNames)
	Vec_WecFree(vGates)

	//    printf( "Synthesized patch with %d inputs, %d outputs and %d gates (const = %d  buf = %d  inv = %d  other = %d).\n",
	//        Vec_IntSize(vUsed), nOuts, nWires, nConst, nBufs, nInvs, nNodes );
	//    printf( "Summary of the results\n" );
	fmt.Printf("\n")
	fmt.Printf("Patch   : in = %d  out = %d : pi_in = %d  po_out = %d : tfi = %d  tfo = %d\n", Vec_IntSize(vUsed), nOuts, nPiCount, nPoCount, Vec_IntSize(vTfi), Vec_IntSize(vTfo))
	fmt.Printf("Added   : gate =%4d : c0 =%2d  c1 =%2d  buf =%3d  inv =%3d  two-input =%4d\n", nWires, nGates0[0], nGates0[1], nGates0[2], nGates0[3], nGates0[4])
	if Abc_FrameReadSpecName() != "" {
		fmt.Printf("Removed : gate =%4d : c0 =%2d  c1 =%2d  buf =%3d  inv =%3d  two-input =%4d\n", nMffc, nGates1[0], nGates1[1], nGates1[2], nGates1[3], nGates1[4])
	}
	if Abc_FrameReadSpecName() != "" {
		fmt.Printf("TOTAL   : gate =%4d : c0 =%2d  c1 =%2d  buf =%3d  inv =%3d  two-input =%4d\n", nWires-nMffc, nGates0[0]-nGates1[0], nGates0[1]-nGates1[1], nGates0[2]-nGates1[2], nGates0[3]-nGates1[3], nGates0[4]-nGates1[4])
	}
	fmt.Printf("\n")
	return vStr
}

/*
*Function*************************************************************

	Synopsis    [Patch generation.]

	Description []

	SideEffects []

	SeeAlso     []

**********************************************************************
*/
func Acb_GenerateInstance2(vIns *Vec_Ptr_t, vOuts *Vec_Ptr_t) *Vec_Str_t {
	var pName string
	var i int32
	var vStr *Vec_Str_t = Vec_StrAlloc(100)
	Vec_StrAppend(vStr, "  patch p0 (")
	// inlining of macro Vec_PtrForEachEntry
	for i = 0; i < Vec_PtrSize(vOuts); i++ {
		pName = Vec_PtrEntry(vOuts, i).(string)
		sep := ""
		if i > 0 {
			sep = ","
		}
		Vec_StrPrintF(vStr, "%s .%s(t%d_%s)", sep, pName, i, pName)
	}
	// inlining of macro Vec_PtrForEachEntry
	for i = 0; i < Vec_PtrSize(vIns); i++ {
		pName = Vec_PtrEntry(vIns, i).(string)
		Vec_StrPrintF(vStr, ", .%s(%s)", pName, pName)
	}
	Vec_StrAppend(vStr, " );\n\n")
	Vec_StrPush(vStr, 0)
	return vStr
}
func Acb_GenerateSignalNames2(vGates *Vec_Wec_t, vIns *Vec_Ptr_t, vOuts *Vec_Ptr_t) *Vec_Ptr_t {
	var nIns int32 = Vec_PtrSize(vIns)
	var nOuts int32 = Vec_PtrSize(vOuts)
	var nNodes int32 = Vec_WecSize(vGates) - nIns - nOuts
	var vRes *Vec_Ptr_t = Vec_PtrStart(Vec_WecSize(vGates))
	var pName string
	var vStr *Vec_Str_t = Vec_StrAlloc(1000)
	var i int32
	var nWires int32 = 1
	// create input names
	// inlining of macro Vec_PtrForEachEntry
	for i = 0; i < Vec_PtrSize(vIns); i++ {
		pName = Vec_PtrEntry(vIns, i).(string)
		Vec_PtrWriteEntry(vRes, i, Abc_UtilStrsav(pName))
	}
	// create names for nodes driving outputs
	// inlining of macro Vec_PtrForEachEntry
	for i = 0; i < Vec_PtrSize(vOuts); i++ {
		pName = Vec_PtrEntry(vOuts, i).(string)
		var vGate *Vec_Int_t = Vec_WecEntry(vGates, nIns+nNodes+i)
		assert(Vec_IntEntry(vGate, 0) == ABC_OPER_BIT_BUF, "Vec_IntEntry(vGate, 0) == ABC_OPER_BIT_BUF")
		Vec_PtrWriteEntry(vRes, Vec_IntEntry(vGate, 1), Abc_UtilStrsav(pName))
	}
	for i = nIns; i < nIns+nNodes; i++ {
		if Vec_PtrEntry(vRes, i) == nil {
			Vec_StrPrintF(vStr, "ww%d", nWires)
			nWires++
			Vec_StrPush(vStr, 0)
			Vec_PtrWriteEntry(vRes, i, Vec_StrReleaseArray(vStr))
		}
	}
	Vec_StrFree(vStr)
	return vRes
}
func Acb_GeneratePatch2(pGia *Gia_Man_t, vIns *Vec_Ptr_t, vOuts *Vec_Ptr_t) *Vec_Str_t {
	var vGates *Vec_Wec_t = Abc_GiaSynthesize(nil, pGia)
	var vGate *Vec_Int_t
	var nIns int32 = Vec_PtrSize(vIns)
	var nOuts int32 = Vec_PtrSize(vOuts)
	var pName string
	var i int32
	var k int32
	var iObj int32
	var nWires int32 = Vec_WecSize(vGates) - nIns - nOuts
	var fFirst int32 = 1
	var vNames *Vec_Ptr_t = Acb_GenerateSignalNames2(vGates, vIns, vOuts)

	var vStr *Vec_Str_t = Vec_StrAlloc(100)
	Vec_StrAppend(vStr, "module patch (")

	// inlining of macro Vec_PtrForEachEntry
	for i = 0; i < Vec_PtrSize(vOuts); i++ {
		pName = Vec_PtrEntry(vOuts, i).(string)
		sep := ""
		if i > 0 {
			sep = ","
		}
		Vec_StrPrintF(vStr, "%s %s", sep, pName)
	}
	// inlining of macro Vec_PtrForEachEntry
	for i = 0; i < Vec_PtrSize(vIns); i++ {
		pName = Vec_PtrEntry(vIns, i).(string)
		Vec_StrPrintF(vStr, ", %s", pName)
	}
	Vec_StrAppend(vStr, " );\n\n")

	Vec_StrAppend(vStr, "  output")
	// inlining of macro Vec_PtrForEachEntry
	for i = 0; i < Vec_PtrSize(vOuts); i++ {
		pName = Vec_PtrEntry(vOuts, i).(string)
		sep := ""
		if i > 0 {
			sep = ","
		}
		Vec_StrPrintF(vStr, "%s %s", sep, pName)
	}
	Vec_StrAppend(vStr, ";\n")

	Vec_StrAppend(vStr, "  input")
	// inlining of macro Vec_PtrForEachEntry
	for i = 0; i < Vec_PtrSize(vIns); i++ {
		pName = Vec_PtrEntry(vIns, i).(string)
		sep := ""
		if i > 0 {
			sep = ","
		}
		Vec_StrPrintF(vStr, "%s %s", sep, pName)
	}
	Vec_StrAppend(vStr, ";\n")

	if nWires > nOuts {
		Vec_StrAppend(vStr, "  wire")
		for i = 0; i < nWires; i++ {
			pName = Vec_PtrEntry(vNames, nIns+i).(string)
			if strncmp(pName, "ww", 2) == 0 {
				sep := ","
				if fFirst != 0 {
					sep = ""
				}
				Vec_StrPrintF(vStr, "%s %s", sep, pName)
				fFirst = 0
			}
		}
		Vec_StrAppend(vStr, ";\n\n")
	}

	// create internal nodes
	// inlining of macro Vec_WecForEachLevelStartStop
	for i = nIns; i < nIns+nWires; i++ {
		vGate = Vec_WecEntry(vGates, i)
		if Vec_IntSize(vGate) > 2 {
			Vec_StrPrintF(vStr, "  %s (", Acb_Oper2Name(Vec_IntEntry(vGate, 0)))
			// inlining of macro Vec_IntForEachEntryStart
			for k = 1; k < Vec_IntSize(vGate); k++ {
				iObj = Vec_IntEntry(vGate, k)
				sep := ""
				if k > 1 {
					sep = ","
				}
				Vec_StrPrintF(vStr, "%s %s", sep, Vec_PtrEntry(vNames, iObj))
			}
			Vec_StrAppend(vStr, " );\n")

		} else {
			assert(Vec_IntEntry(vGate, 0) == ABC_OPER_CONST_F || Vec_IntEntry(vGate, 0) == ABC_OPER_CONST_T, "Vec_IntEntry(vGate, 0) == ABC_OPER_CONST_F || Vec_IntEntry(vGate, 0) == ABC_OPER_CONST_T")
			Vec_StrPrintF(vStr, "  %s (", Acb_Oper2Name(ABC_OPER_BIT_BUF))
			Vec_StrPrintF(vStr, " %s, ", Vec_PtrEntry(vNames, Vec_IntEntry(vGate, 1)))
			bit := 0
			if Vec_IntEntry(vGate, 0) == ABC_OPER_CONST_T {
				bit = 1
			}
			Vec_StrPrintF(vStr, " 1'b%d", bit)
			Vec_StrPrintF(vStr, " );\n")
		}
	}
	Vec_StrAppend(vStr, "\nendmodule\n\n")
	Vec_StrPush(vStr, 0)
	Vec_PtrFreeFree(vNames)
	Vec_WecFree(vGates)

	fmt.Printf("Synthesized patch with %d inputs, %d outputs and %d gates.\n", nIns, nOuts, nWires)
	return vStr
}
func Acb_GenerateFile2(pGia *Gia_Man_t, vIns *Vec_Ptr_t, vOuts *Vec_Ptr_t, pFileName string, pFileNameOut string, fSkipMffc int32) {
	var vInst *Vec_Str_t = Acb_GenerateInstance2(vIns, vOuts)
	var vPatch *Vec_Str_t = Acb_GeneratePatch2(pGia, vIns, vOuts)
	//printf( "%s", Vec_StrArray(vPatch) );
	//Gia_AigerWrite( pGia, "test.aig", 0, 0, 0 );
	// generate output files
	Acb_GenerateFilePatch(vPatch, "patch.v")
	fmt.Printf("Finished dumping patch file \"%s\".\n", "patch.v")
	Acb_NtkInsert(pFileName, "temp.v", vOuts, 0, fSkipMffc)
	fmt.Printf("Finished dumping intermediate file \"%s\".\n", "temp.v")
	Acb_GenerateFileOut(vInst, "temp.v", pFileNameOut, vPatch)
	fmt.Printf("Finished dumping the resulting file \"%s\".\n", pFileNameOut)
	Vec_StrFree(vInst)
	Vec_StrFree(vPatch)
}

/*
*Function*************************************************************

	Synopsis    [Produce output files.]

	Description []

	SideEffects []

	SeeAlso     []

**********************************************************************
*/
func Acb_GenerateFilePatch(p *Vec_Str_t, pFileNamePatch string) {
	pFile, err := os.OpenFile(pFileNamePatch, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, 0666)
	if err != nil {
		return
	}
	_, _ = pFile.Write(Vec_StrArray(p))
	_ = pFile.Close()
}
func Acb_GenerateFileOut(vPatchLine *Vec_Str_t, pFileNameF string, pFileNameOut string, vPatch *Vec_Str_t) {
	pBuffer := Extra_FileReadContents(pFileNameF)
	if pBuffer == "" {
		return
	}
	pFileOut, err := os.OpenFile(pFileNameOut, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, 0666)
	if err != nil {
		return
	}
	defer pFileOut.Close()
	nFirst := strings.Index(pBuffer, "endmodule")
	if nFirst < 0 {
		return
	}
	_, _ = fmt.Fprint(pFileOut, pBuffer[:nFirst])
	_, _ = fmt.Fprintf(pFileOut, "\n%s", Vec_StrArray(vPatchLine))
	_, _ = fmt.Fprint(pFileOut, pBuffer[nFirst:])
	if vPatch != nil {
		_, _ = fmt.Fprintf(pFileOut, "\n%s\n", Vec_StrArray(vPatch))
	}
}

/*
*Function*************************************************************

	Synopsis    [Print parameters of the patch.]

	Description []

	SideEffects []

	SeeAlso     []

**********************************************************************
*/
func Acb_PrintPatch(pNtkF *Acb_Ntk_t, vDivs *Vec_Int_t, vUsed *Vec_Int_t, clk abctime) {
	var i, iObj, Weight int32
	fmt.Printf("Patch has %d inputs: ", Vec_IntSize(vUsed))
	// inlining of macro Vec_IntForEachEntryInVec
	for i = 0; i < Vec_IntSize(vUsed); i++ {
		iObj = Vec_IntEntry(vDivs, Vec_IntEntry(vUsed, i))
		fmt.Printf("%d=%s(w=%d) ", Vec_IntEntry(vUsed, i), Acb_ObjNameStr(pNtkF, iObj), Acb_ObjWeight(pNtkF, iObj))
		Weight += Acb_ObjWeight(pNtkF, iObj)
	}
	fmt.Printf("\nTotal weight = %d  ", Weight)
	Abc_PrintTime(1, "Total runtime", Abc_Clock()-clk)
	fmt.Printf("\n")
}

/*
*Function*************************************************************

	Synopsis    [Quantifies targets 0 up to iTar (out of nTars).]

	Description []

	SideEffects []

	SeeAlso     []

**********************************************************************
*/
func Acb_NtkEcoSynthesize(p *Gia_Man_t) *Gia_Man_t {
	var Iter, fVerbose int32
	var pNew *Gia_Man_t = Gia_ManDup(p)

	if fVerbose != 0 {
		fmt.Printf("M_quo: ")
	}
	if fVerbose != 0 {
		Gia_ManPrintStats(pNew, nil)
	}

	p = pNew
	pNew = Gia_ManAreaBalance(p, 0, 0, 0, 0)
	Gia_ManStop(p)

	if fVerbose != 0 {
		fmt.Printf("M_bal: ")
	}
	if fVerbose != 0 {
		Gia_ManPrintStats(pNew, nil)
	}

	for Iter = 0; Iter < 2; Iter++ {
		p = pNew
		pNew = Gia_ManCompress2(p, 1, 0)
		Gia_ManStop(p)

		if fVerbose != 0 {
			fmt.Printf("M_dc2: ")
		}
		if fVerbose != 0 {
			Gia_ManPrintStats(pNew, nil)
		}
	}

	p = pNew
	pNew = Gia_ManAigSyn2(p, 0, 1, 0, 100, 0, 0, 0)
	Gia_ManStop(p)

	if fVerbose != 0 {
		fmt.Printf("M_sn2: ")
	}
	if fVerbose != 0 {
		Gia_ManPrintStats(pNew, nil)
	}

	for Iter = 0; Iter < 2; Iter++ {
		p = pNew
		pNew = Gia_ManCompress2(p, 1, 0)
		Gia_ManStop(p)

		if fVerbose != 0 {
			fmt.Printf("M_dc2: ")
		}
		if fVerbose != 0 {
			Gia_ManPrintStats(pNew, nil)
		}
	}
	return pNew
}
func Acb_NtkDeriveMiterCnf(p *Gia_Man_t, iTar int32, nTars int32, fVerbose int32) *Cnf_Dat_t {
	var pCof *Gia_Man_t = Gia_ManDup(p)
	var pCnf *Cnf_Dat_t
	var v int32
	for v = 0; v < iTar; v++ {
		var pTemp *Gia_Man_t
		p = pCof
		pCof = Gia_ManDupUniv(p, Gia_ManCiNum(pCof)-nTars+v)
		//pCof = Acb_NtkEcoSynthesize( pTemp = pCof );
		//pCof = Gia_ManCompress2( pTemp = pCof, 1, 0 );
		pTemp = pCof
		pCof = Gia_ManAigSyn2(pTemp, 0, 1, 0, 100, 0, 0, 0)
		Gia_ManStop(pTemp)
		if Gia_ManAndNum(pCof) > 10000 {
			fmt.Printf("Quantifying target %3d ", v)
			Gia_ManPrintStats(pCof, nil)
		}
		assert(Gia_ManCiNum(pCof) == Gia_ManCiNum(p), "Gia_ManCiNum(pCof) == Gia_ManCiNum(p)")
		Gia_ManStop(p)
	}
	if fVerbose != 0 {
		fmt.Printf("M_quo: ")
	}
	if fVerbose != 0 {
		Gia_ManPrintStats(pCof, nil)
	}
	//pCof = Acb_NtkEcoSynthesize( p = pCof );
	//Gia_ManStop( p );
	if fVerbose != 0 {
		fmt.Printf("M_syn: ")
	}
	if fVerbose != 0 {
		Gia_ManPrintStats(pCof, nil)
	}
	if false && iTar < nTars {
		var pCof0 *Gia_Man_t = Gia_ManDupCofactorVar(pCof, Gia_ManCiNum(pCof)-nTars+iTar, 0)
		var pCof1 *Gia_Man_t = Gia_ManDupCofactorVar(pCof, Gia_ManCiNum(pCof)-nTars+iTar, 1)
		p = pCof0
		pCof0 = Acb_NtkEcoSynthesize(p)
		Gia_ManStop(p)
		p = pCof1
		pCof1 = Acb_NtkEcoSynthesize(p)
		Gia_ManStop(p)
		Gia_AigerWrite(pCof0, "eco_qbf0.aig", 0, 0, 0)
		Gia_AigerWrite(pCof1, "eco_qbf1.aig", 0, 0, 0)
		Gia_ManStop(pCof0)
		Gia_ManStop(pCof1)
		fmt.Printf("Dumped cof0 into file \"%s\".\n", "eco_qbf0.aig")
		fmt.Printf("Dumped cof1 into file \"%s\".\n", "eco_qbf1.aig")
	}
	//    Gia_AigerWrite( pCof, "eco_qbf.aig", 0, 0, 0 );
	//    printf( "Dumped the result of quantification into file \"%s\".\n", "eco_qbf.aig" );
	pCnf = Mf_ManGenerateCnf(pCof, 8, 0, 0, 0, 0)
	Gia_ManStop(pCof)
	return pCnf
}
func Gia_ManInterOneInt(pCof1 *Gia_Man_t, pCof0 *Gia_Man_t, Depth int32) *Gia_Man_t {
	var pGia [2]*Gia_Man_t = [2]*Gia_Man_t{pCof0, pCof1}
	var pCof [2][2]*Gia_Man_t
	var pTemp *Gia_Man_t
	var pInter [2]*Gia_Man_t
	var pFinal *Gia_Man_t
	var pObj *Gia_Obj_t
	var i int32
	var n int32
	var m int32
	var Count int32
	var CountBest int32
	var iVarBest int32 = -1
	// find PIs with the highest fanout
	var vFanCount *Vec_Int_t
	if Gia_ManAndNum(pCof1) == 0 || Gia_ManAndNum(pCof0) == 0 {
		return Gia_ManDup(pCof1)
	}
	vFanCount = Vec_IntStart(Gia_ManCiNum(pCof0))
	for n = 0; n < 2; n++ {
		// inlining of macro Gia_ManForEachAnd
		for i = 0; i < Gia_ManObjNum(pGia[n]); i++ {
			pObj = Gia_ManObj(pGia[n], i)
			if Gia_ObjIsAnd(pObj) == 0 {
			} else {
				if Gia_ObjIsCi(Gia_ObjFanin0(pObj)) != 0 {
					Vec_IntAddToEntry(vFanCount, Gia_ObjCioId(Gia_ObjFanin0(pObj)), 1)
				}
				if Gia_ObjIsCi(Gia_ObjFanin1(pObj)) != 0 {
					Vec_IntAddToEntry(vFanCount, Gia_ObjCioId(Gia_ObjFanin1(pObj)), 1)
				}
			}
		}
	}
	// inlining of macro Vec_IntForEachEntry
	for i = 0; i < Vec_IntSize(vFanCount); i++ {
		Count = Vec_IntEntry(vFanCount, i)
		if CountBest < Count {
			CountBest = Count
			iVarBest = i
		}
	}
	Vec_IntFree(vFanCount)
	// Gia_Man_t * Gia_ManDupCofactorVar( Gia_Man_t * p, int iVar, int Value )
	for n = 0; n < 2; n++ {
		for m = 0; m < 2; m++ {
			pCof[n][m] = Gia_ManDupCofactorVar(pGia[n], iVarBest, m)
			pTemp = pCof[n][m]
			pCof[n][m] = Acb_NtkEcoSynthesize(pTemp)
			Gia_ManStop(pTemp)
			fmt.Printf("%*sCof%d%d : ", 8-Depth, "", n, m)
			Gia_ManPrintStats(pCof[n][m], nil)
		}
	}
	for m = 0; m < 2; m++ {
		if Gia_ManAndNum(pCof[1][m]) == 0 || Gia_ManAndNum(pCof[0][m]) == 0 {
			pInter[m] = Gia_ManDup(pCof[1][m])
		} else if Depth == 1 {
			pInter[m] = Gia_ManInterOne(pCof[1][m], pCof[0][m], 1)
		} else {
			pInter[m] = Gia_ManInterOneInt(pCof[1][m], pCof[0][m], Depth-1)
		}
		fmt.Printf("%*sInter%d : ", 8-Depth, "", m)
		Gia_ManPrintStats(pInter[m], nil)
		pTemp = pInter[m]
		pInter[m] = Abc_GiaSynthesizeInter(pTemp)
		Gia_ManStop(pTemp)
		fmt.Printf("%*sInter%d : ", 8-Depth, "", m)
		Gia_ManPrintStats(pInter[m], nil)
	}
	for n = 0; n < 2; n++ {
		for m = 0; m < 2; m++ {
			Gia_ManStop(pCof[n][m])
		}
	}
	pFinal = Gia_ManDupMux(iVarBest, pInter[1], pInter[0])
	for m = 0; m < 2; m++ {
		Gia_ManStop(pInter[m])
	}
	return pFinal
}
func Acb_NtkDeriveMiterCnfInter2(p *Gia_Man_t, iTar int32, nTars int32) *Gia_Man_t {
	//    extern Gia_Man_t * Gia_ManInterOne( Gia_Man_t * pNtkOn, Gia_Man_t * pNtkOff, int fVerbose );
	var pInter *Gia_Man_t
	var pCof0 *Gia_Man_t
	var pCof1 *Gia_Man_t
	var pCof *Gia_Man_t = Gia_ManDup(p)
	var v int32
	for v = 0; v < iTar; v++ {
		p = pCof
		pCof = Gia_ManDupUniv(p, Gia_ManCiNum(pCof)-nTars+v)
		assert(Gia_ManCiNum(pCof) == Gia_ManCiNum(p), "Gia_ManCiNum(pCof) == Gia_ManCiNum(p)")
		Gia_ManStop(p)

		p = pCof
		pCof = Acb_NtkEcoSynthesize(p)
		Gia_ManStop(p)
	}
	pCof0 = Gia_ManDupCofactorVar(pCof, Gia_ManCiNum(pCof)-nTars+iTar, 1)
	pCof1 = Gia_ManDupCofactorVar(pCof, Gia_ManCiNum(pCof)-nTars+iTar, 0)
	Gia_ManStop(pCof)
	p = pCof0
	pCof0 = Acb_NtkEcoSynthesize(p)
	Gia_ManStop(p)
	p = pCof1
	pCof1 = Acb_NtkEcoSynthesize(p)
	Gia_ManStop(p)

	fmt.Printf("Cof0 : ")
	Gia_ManPrintStats(pCof0, nil)
	fmt.Printf("Cof1 : ")
	Gia_ManPrintStats(pCof1, nil)

	if Gia_ManAndNum(pCof1) == 0 || Gia_ManAndNum(pCof0) == 0 {
		pInter = Gia_ManDup(pCof1)
	} else {
		pInter = Gia_ManInterOneInt(pCof1, pCof0, 7)
	}
	Gia_ManStop(pCof0)
	Gia_ManStop(pCof1)
	p = pInter
	pInter = Abc_GiaSynthesizeInter(p)
	Gia_ManStop(p)
	//Gia_ManPrintStats( pInter, NULL );
	p = pInter
	pInter = Gia_ManDupRemovePis(p, nTars)
	Gia_ManStop(p)
	//Gia_ManPrintStats( pInter, NULL );
	return pInter
}
func Acb_NtkDeriveMiterCnfInter(p *Gia_Man_t, iTar int32, nTars int32) *Gia_Man_t {
	var pCof1 *Gia_Man_t
	var pCof *Gia_Man_t = Gia_ManDup(p)
	var v int32
	for v = 0; v < iTar; v++ {
		p = pCof
		pCof = Gia_ManDupUniv(p, Gia_ManCiNum(pCof)-nTars+v)
		assert(Gia_ManCiNum(pCof) == Gia_ManCiNum(p), "Gia_ManCiNum(pCof) == Gia_ManCiNum(p)")
		Gia_ManStop(p)

		p = pCof
		pCof = Acb_NtkEcoSynthesize(p)
		Gia_ManStop(p)
	}
	pCof1 = Gia_ManDupCofactorVar(pCof, Gia_ManCiNum(pCof)-nTars+iTar, 0)
	Gia_ManStop(pCof)

	p = pCof1
	pCof1 = Acb_NtkEcoSynthesize(p)
	Gia_ManStop(p)

	p = pCof1
	pCof1 = Gia_ManDupRemovePis(p, nTars)
	Gia_ManStop(p)
	return pCof1
}

/*
*Function*************************************************************

	Synopsis    [Transform patch functions to have common support.]

	Description []

	SideEffects []

	SeeAlso     []

**********************************************************************
*/
func Acb_RemapOneFunction(pStr string, vSupp *Vec_Int_t, vMap *Vec_Int_t, nVars int32) string {
	vTempSop := Vec_StrAlloc(100)
	for _, pToken := range strings.Split(pStr, "\n") {
		if pToken == "" {
			continue
		}
		for i := int32(0); i < nVars; i++ {
			Vec_StrPush(vTempSop, '-')
		}
		i := 0
		for i < len(pToken) && pToken[i] != ' ' {
			if pToken[i] != '-' {
				iVar := Vec_IntEntry(vMap, Vec_IntEntry(vSupp, int32(i)))
				assert(iVar >= 0 && iVar < nVars, "iVar >= 0 && iVar < nVars")
				Vec_StrWriteEntry(vTempSop, Vec_StrSize(vTempSop)-nVars+iVar, pToken[i])
			}
			i++
		}
		assert(i+1 < len(pToken), "SOP row has output literal")
		Vec_StrPrintF(vTempSop, " %d\n", pToken[i+1]-'0')
	}
	Vec_StrPush(vTempSop, 0)
	result := strings.TrimRight(string(Vec_StrReleaseArray(vTempSop)), "\x00")
	Vec_StrFree(vTempSop)
	return result
}

func Acb_TransformPatchFunctions(vSops *Vec_Ptr_t, vSupps *Vec_Wec_t, pvUsed **Vec_Int_t, nDivs int32) *Vec_Ptr_t {
	var vFuncs *Vec_Ptr_t = Vec_PtrAlloc(Vec_PtrSize(vSops))
	var vUsed *Vec_Int_t = Vec_IntAlloc(100)
	var vMap *Vec_Int_t = Vec_IntStartFull(nDivs)
	var vPres *Vec_Int_t = Vec_IntStart(nDivs)
	var vLevel *Vec_Int_t
	var i, k, iVar int32
	// check what divisors are used
	// inlining of macro Vec_WecForEachLevel
	for i = 0; i < Vec_WecSize(vSupps); i++ {
		vLevel = Vec_WecEntry(vSupps, i)
		pSop := Vec_PtrEntry(vSops, i).(string)
		for _, pToken := range strings.Split(pSop, "\n") {
			if pToken == "" {
				continue
			}
			for k = 0; pToken[k] != ' '; k++ {
				if pToken[k] != '-' {
					Vec_IntWriteEntry(vPres, Vec_IntEntry(vLevel, k), 1)
				}
			}
		}
	}
	// create common order
	// inlining of macro Vec_WecForEachLevel
	for i = 0; i < Vec_WecSize(vSupps); i++ {
		vLevel = Vec_WecEntry(vSupps, i)
		// inlining of macro Vec_IntForEachEntry
		for k = 0; k < Vec_IntSize(vLevel); k++ {
			iVar = Vec_IntEntry(vLevel, k)
			if Vec_IntEntry(vPres, iVar) == 0 {
				continue
			}
			if Vec_IntEntry(vMap, iVar) >= 0 {
				continue
			}
			Vec_IntWriteEntry(vMap, iVar, Vec_IntSize(vUsed))
			Vec_IntPush(vUsed, iVar)
		}
	}
	//printf( "The number of used variables %d (out of %d).\n", Vec_IntSum(vPres), Vec_IntSize(vPres) );
	// remap SOPs
	// inlining of macro Vec_WecForEachLevel
	for i = 0; i < Vec_WecSize(vSupps); i++ {
		vLevel = Vec_WecEntry(vSupps, i)
		pSop := Vec_PtrEntry(vSops, i).(string)
		pSop = Acb_RemapOneFunction(pSop, vLevel, vMap, Vec_IntSize(vUsed))
		//printf( "Function %d\n%s", i, pSop );
		Vec_PtrPush(vFuncs, pSop)
	}
	Vec_IntFree(vPres)
	Vec_IntFree(vMap)
	*pvUsed = vUsed
	return vFuncs
}

/*
*Function*************************************************************

	Synopsis    [Performs ECO for two networks.]

	Description []

	SideEffects []

	SeeAlso     []

**********************************************************************
*/
func Acb_NtkEcoPerform(pNtkF *Acb_Ntk_t, pNtkG *Acb_Ntk_t, pFileName []string, nTimeout int32, fCisOnly int32, fInputs int32, fCheck int32, fVerbose int32, fVeryVerbose int32) int32 {

	var clkStart abctime = Abc_Clock()
	var clk abctime = Abc_Clock()
	var nTargets int32 = Vec_IntSize(&pNtkF.vTargets)
	var TimeOut int32
	if fCisOnly != 0 {
		TimeOut = 0
	} else {
		TimeOut = 120
	} // 60 seconds
	var RetValue int32 = 1

	// compute various sets of nodes
	var vBlock *Vec_Bit_t
	var vRoots *Vec_Int_t = Acb_NtkFindRoots(pNtkF, &pNtkF.vTargets, &vBlock)
	var vSuppF *Vec_Int_t = Acb_NtkFindSupp(pNtkF, vRoots)
	var vSuppG *Vec_Int_t = Acb_NtkFindSupp(pNtkG, vRoots)
	var vSupp *Vec_Int_t = Vec_IntTwoMerge(vSuppF, vSuppG)
	var vDivs *Vec_Int_t
	if (fCisOnly != 0) || (fInputs != 0) {
		vDivs = Acb_NtkFindDivsCis(pNtkF, vSupp)
	} else {
		vDivs = Acb_NtkFindDivs(pNtkF, vSupp, vBlock, fVerbose)
	}
	var vNodesF *Vec_Int_t = Acb_NtkFindNodes(pNtkF, vRoots, vDivs)
	var vNodesG *Vec_Int_t = Acb_NtkFindNodes(pNtkG, vRoots, nil)

	// create AIGs
	var pGiaF *Gia_Man_t = Acb_NtkToGia(pNtkF, vSupp, vNodesF, vRoots, vDivs, &pNtkF.vTargets)
	var pGiaG *Gia_Man_t = Acb_NtkToGia(pNtkG, vSupp, vNodesG, vRoots, nil, nil)
	var pGiaM *Gia_Man_t = Acb_CreateMiter(pGiaF, pGiaG)

	var pCnf *Cnf_Dat_t
	var pTemp, pOne *Gia_Man_t
	var vSops *Vec_Ptr_t = Vec_PtrAlloc(nTargets)
	var vSupps *Vec_Wec_t = Vec_WecAlloc(nTargets)
	var vSuppOld *Vec_Int_t = Vec_IntAlloc(100)

	var vUsed *Vec_Int_t = nil
	var vFuncs *Vec_Ptr_t = nil
	var vGias *Vec_Ptr_t
	if fCisOnly != 0 {
		vGias = Vec_PtrAlloc(nTargets)
	} else {
		vGias = nil
	}
	var vInst *Vec_Str_t = nil
	var vPatch *Vec_Str_t = nil

	var pSop string
	var outFile string
	var i int32

	if fVerbose != 0 {
		fmt.Printf("The number of targets = %d.\n", nTargets)

		fmt.Printf("NtkF:  ")
		Gia_ManPrintStats(pGiaF, nil)
		fmt.Printf("NtkG:  ")
		Gia_ManPrintStats(pGiaG, nil)
		fmt.Printf("Miter: ")
		Gia_ManPrintStats(pGiaM, nil)
	}

	// check that the problem has a solution
	if fCheck != 0 { //fCisOnly )

		var Lit, status int32
		var pSat *sat_solver
		pCnf = Acb_NtkDeriveMiterCnf(pGiaM, nTargets, nTargets, fVerbose)
		pSat = Cnf_DataWriteIntoSolver(pCnf, 1, 0)
		Cnf_DataFree(pCnf)
		// add output clause
		Lit = Abc_Var2Lit(1, 0)
		if !sat_solver_addclause(pSat, []int32{Lit}) {
			status = l_False
		} else {
			status = sat_solver_solve(pSat, nil, nil, 0, 0, 0, 0)
		}
		sat_solver_delete(pSat)
		result := "NO"
		if status == l_False {
			result = "a"
		}
		fmt.Printf("The ECO problem has %s solution. ", result)
		Abc_PrintTime(1, "Time", Abc_Clock()-clk)
		if status != l_False {
			RetValue = 0
			goto cleanup
		}
	}

	for i = nTargets - 1; i >= 0; i-- {
		var vSupp *Vec_Int_t = nil
		fmt.Printf("\nConsidering target %d (out of %d)...\n", i, nTargets)
		// compute support of this target
		if fCisOnly != 0 {
			vSupp = Vec_IntStartNatural(Vec_IntSize(vDivs))
			fmt.Printf("Target %d has support with %d variables.\n", i, Vec_IntSize(vSupp))

			pOne = Acb_NtkDeriveMiterCnfInter(pGiaM, i, nTargets)
			fmt.Printf("Tar%02d: ", i)
			Gia_ManPrintStats(pOne, nil)

			// update miter
			pTemp = pGiaM
			pGiaM = Acb_UpdateMiter(pTemp, pOne, i, nTargets, vSupp, fCisOnly)
			Gia_ManStop(pTemp)

			// add to functions
			Vec_PtrPush(vGias, pOne)

		} else {
			pCnf = Acb_NtkDeriveMiterCnf(pGiaM, i, nTargets, fVerbose)
			//            vSupp = Acb_DerivePatchSupportS( pCnf, i, nTargets, Vec_IntSize(vDivs), vDivs, pNtkF, NULL, TimeOut );
			vSupp = Acb_DerivePatchSupport(pCnf, i, nTargets, Vec_IntSize(vDivs), vDivs, pNtkF, vSuppOld, TimeOut)
			if vSupp == nil {
				Cnf_DataFree(pCnf)
				RetValue = 0
				goto cleanup
			}
			Vec_IntAppend(vSuppOld, vSupp)
			Vec_IntClear(vSupp)
			Vec_IntAppend(vSupp, vSuppOld)
			//Vec_IntClear( vSuppOld );

			// derive function of this target
			pSop = Acb_DeriveOnePatchFunction(pCnf, i, nTargets, Vec_IntSize(vDivs), vSupp, fCisOnly)
			Cnf_DataFree(pCnf)
			if pSop == "" {
				RetValue = 0
				goto cleanup
			}
			if (nTimeout != 0) && int32((Abc_Clock()-clkStart)/int64(time.Second)) >= nTimeout {
				Vec_IntFreeP(&vSupp)
				pSop = ""
				fmt.Printf("The target computation timed out after %d seconds.\n", nTimeout)
				RetValue = 0
				goto cleanup
			}

			// add new function to the miter
			pOne = baseabc.Abc_SopSynthesizeOne(pSop, 1)
			fmt.Printf("Tar%02d: ", i)
			Gia_ManPrintStats(pOne, nil)

			// update miter
			pTemp = pGiaM
			pGiaM = Acb_UpdateMiter(pTemp, pOne, i, nTargets, vSupp, fCisOnly)
			Gia_ManStop(pTemp)
			Gia_ManStop(pOne)

			// add to functions
			Vec_PtrPush(vSops, pSop)
			if fVeryVerbose != 0 {
				fmt.Printf("Function %d\n%s", i, pSop)
			}
		}
		// add to supports
		Vec_IntAppend(Vec_WecPushLevel(vSupps), vSupp)
		Vec_IntFree(vSupp)
	}

	// make sure the function is UNSAT
	fmt.Printf("\n")
	if !(fCisOnly != 0) {
		var Res int32
		var clk abctime = Abc_Clock()
		pCnf = Mf_ManGenerateCnf(pGiaM, 8, 0, 0, 0, 0)
		Res = Acb_CheckMiter(pCnf)
		Cnf_DataFree(pCnf)
		if Res == 1 {
			fmt.Printf("The ECO solution was verified successfully.  ")
		} else {
			fmt.Printf("The ECO solution verification FAILED.  ")
		}
		Abc_PrintTime(1, "Time", Abc_Clock()-clk)
	}

	// derive new patch functions
	if fCisOnly != 0 {
		vUsed = Vec_IntStartNatural(Vec_IntSize(vDivs))
		Vec_PtrReverseOrder(vGias)

	} else {
		vFuncs = Acb_TransformPatchFunctions(vSops, vSupps, &vUsed, Vec_IntSize(vDivs))
		Vec_PtrReverseOrder(vFuncs)
	}

	// generate instance and patch
	vInst = Acb_GenerateInstance(pNtkF, vDivs, vUsed, &pNtkF.vTargets)
	vPatch = Acb_GeneratePatch(pNtkF, vDivs, vUsed, vFuncs, vGias, &pNtkF.vTargets)

	// print the results
	//printf( "%s", Vec_StrArray(vPatch) );
	Acb_PrintPatch(pNtkF, vDivs, vUsed, clk)

	// generate output files
	if pFileName[3] == "" {
		Acb_GenerateFilePatch(vPatch, "patch.v")
	}
	outFile = pFileName[3]
	if outFile == "" {
		outFile = "out.v"
	}
	Acb_GenerateFileOut(vInst, pFileName[0], outFile, vPatch)
	fmt.Printf("Finished dumping resulting file \"%s\".\n\n", outFile)
	//Gia_AigerWrite( pGiaG, "test.aig", 0, 0, 0 );
cleanup:
	// cleanup
	if vGias != nil {
		var pTemp *Gia_Man_t
		var i int32
		// inlining of macro Vec_PtrForEachEntry
		for i = 0; i < Vec_PtrSize(vGias); i++ {
			pTemp = Vec_PtrEntry(vGias, i).(*Gia_Man_t)
			Gia_ManStop(pTemp)
		}
		Vec_PtrFree(vGias)
	}
	Vec_StrFreeP(&vPatch)
	Vec_StrFreeP(&vInst)

	Vec_PtrFreeFree(vSops)
	Vec_WecFree(vSupps)
	Vec_IntFree(vSuppOld)
	Vec_IntFreeP(&vUsed)
	if vFuncs != nil {
		Vec_PtrFreeFree(vFuncs)
	}

	Gia_ManStop(pGiaF)
	Gia_ManStop(pGiaG)
	Gia_ManStop(pGiaM)

	Vec_IntFreeP(&vSuppF)
	Vec_IntFreeP(&vSuppG)
	Vec_IntFreeP(&vSupp)
	Vec_IntFreeP(&vNodesF)
	Vec_IntFreeP(&vNodesG)
	Vec_IntFreeP(&vRoots)
	Vec_IntFreeP(&vDivs)
	Vec_BitFreeP(&vBlock)
	return RetValue
}

/*
*Function*************************************************************

	Synopsis    [Read/write test.]

	Description []

	SideEffects []

	SeeAlso     []

**********************************************************************
*/
func Acb_NtkTestRun2(pFileNames []string, fVerbose int32) {
	var pFileNameOut string = extra.Extra_FileNameGenericAppend(pFileNames[0], "_w.v")
	var pNtk *Acb_Ntk_t = Acb_VerilogSimpleRead(pFileNames[0], pFileNames[2])
	Acb_VerilogSimpleWrite(pNtk, pFileNameOut)
	Acb_ManFree(pNtk.pDesign)
	Acb_IntallLibrary(0)
}

/*
*Function*************************************************************

	Synopsis    [Top level procedure.]

	Description []

	SideEffects []

	SeeAlso     []

**********************************************************************
*/
func Acb_NtkRunEco(pFileNames []string, nTimeout int32, fCheck int32, fRandom int32, fInputs int32, fVerbose int32, fVeryVerbose int32) {
	var Command string // was [1000]byte in C
	var Result int32 = 1
	var pNtkF *Acb_Ntk_t = Acb_VerilogSimpleRead(pFileNames[0], pFileNames[2])
	var pNtkG *Acb_Ntk_t = Acb_VerilogSimpleRead(pFileNames[1], "")
	if !(pNtkF != nil) || !(pNtkG != nil) {
		return
	}
	//int * pArray = Vec_IntArray( &pNtkF->vTargets );
	//ABC_SWAP( int, pArray[7], pArray[4] );
	//Vec_IntReverseOrder( &pNtkF->vTargets );
	if fRandom != 0 {
		fmt.Printf("Permuting targets as follows:   ")
		Vec_IntPermute(&pNtkF.vTargets)
		Vec_IntPrint(&pNtkF.vTargets)
	}

	assert(Acb_NtkCiNum(pNtkF) == Acb_NtkCiNum(pNtkG), "Acb_NtkCiNum(pNtkF) == Acb_NtkCiNum(pNtkG)")
	assert(Acb_NtkCoNum(pNtkF) == Acb_NtkCoNum(pNtkG), "Acb_NtkCoNum(pNtkF) == Acb_NtkCoNum(pNtkG)")

	if Abc_FrameReadSignalNames() != nil {
		Acb_IntallLibrary(1)
	} else {
		Acb_IntallLibrary(0)
	}

	if !(Acb_NtkEcoPerform(pNtkF, pNtkG, pFileNames, nTimeout, 0, fInputs, fCheck, fVerbose, fVeryVerbose) != 0) {
		//        printf( "General computation timed out. Trying inputs only.\n\n" );
		//        if ( !Acb_NtkEcoPerform( pNtkF, pNtkG, pFileNames, nTimeout, 1, fInputs, fCheck, fVerbose, fVeryVerbose ) )
		//            printf( "Input-only computation also timed out.\n\n" );
		fmt.Printf("Computation did not succeed.\n")
		Result = 0
	}

	Acb_ManFree(pNtkF.pDesign)
	Acb_ManFree(pNtkG.pDesign)

	// verify the result
	output := pFileNames[3]
	if output == "" {
		output = "out.v"
	}
	Command = fmt.Sprintf("read %s; strash; write temp1.aig; read %s; strash; write temp2.aig; &cec temp1.aig temp2.aig", pFileNames[1], output)
	if (Result != 0) && basemain.Cmd_CommandExecute(basemain.Abc_FrameGetGlobalFrame(), Command) != 0 {
		fmt.Fprintf(os.Stdout, "Cannot execute command \"%s\".\n", Command)
	}
	fmt.Printf("\n")
}

////////////////////////////////////////////////////////////////////////
///                       END OF FILE                                ///
////////////////////////////////////////////////////////////////////////

//ABC_NAMESPACE_IMPL_END
