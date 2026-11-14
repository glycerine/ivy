package bac

import (
	"fmt"
	"github.com/glycerine/abcgo/misc/vec"
	"strings"
)

/**CFile****************************************************************

  FileName    [bacReadVer.c]

  SystemName  [ABC: Logic synthesis and verification system.]

  PackageName [Hierarchical word-level netlist.]

  Synopsis    [BLIF writer.]

  Author      [Alan Mishchenko]

  Affiliation [UC Berkeley]

  Date        [Ver. 1.0. Started - November 29, 2014.]

  Revision    [$Id: bacReadVer.c,v 1.00 2014/11/29 00:00:00 alanmi Exp $]

***********************************************************************/

//#include "bac.h"
//#include "bacPrs.h"

//ABC_NAMESPACE_IMPL_START

////////////////////////////////////////////////////////////////////////
///                        DECLARATIONS                              ///
////////////////////////////////////////////////////////////////////////

// Verilog keywords
type Bac_VerType_t int32

const (
	PRS_VER_NONE Bac_VerType_t = 0
	// 0:  unused
	PRS_VER_INPUT Bac_VerType_t = PRS_VER_NONE + 1
	// 1:  input
	PRS_VER_OUTPUT Bac_VerType_t = PRS_VER_INPUT + 1
	// 2:  output
	PRS_VER_INOUT Bac_VerType_t = PRS_VER_OUTPUT + 1
	// 3:  inout
	PRS_VER_WIRE Bac_VerType_t = PRS_VER_INOUT + 1
	// 4:  wire
	PRS_VER_MODULE Bac_VerType_t = PRS_VER_WIRE + 1
	// 5:  module
	PRS_VER_ASSIGN Bac_VerType_t = PRS_VER_MODULE + 1
	// 6:  assign
	PRS_VER_REG Bac_VerType_t = PRS_VER_ASSIGN + 1
	// 7:  reg
	PRS_VER_ALWAYS Bac_VerType_t = PRS_VER_REG + 1
	// 8:  always
	PRS_VER_DEFPARAM Bac_VerType_t = PRS_VER_ALWAYS + 1
	// 9:  always
	PRS_VER_BEGIN Bac_VerType_t = PRS_VER_DEFPARAM + 1
	// 10: begin
	PRS_VER_END Bac_VerType_t = PRS_VER_BEGIN + 1
	// 11: end
	PRS_VER_ENDMODULE Bac_VerType_t = PRS_VER_END + 1
	// 12: endmodule
	PRS_VER_UNKNOWN Bac_VerType_t = PRS_VER_ENDMODULE + 1 // 13: unknown

)

var s_VerTypes = []string{
	"",          // 0:  unused
	"input",     // 1:  input
	"output",    // 2:  output
	"inout",     // 3:  inout
	"wire",      // 4:  wire
	"module",    // 5:  module
	"assign",    // 6:  assign
	"reg",       // 7:  reg
	"always",    // 8:  always
	"defparam",  // 9:  defparam
	"begin",     // 10: begin
	"end",       // 11: end
	"endmodule", // 12: endmodule
	"",          // 13: unknown
}

func Psr_NtkAddVerilogDirectives(p *Psr_Man_t) {
	var i int32
	for i = 1; i < int32(len(s_VerTypes))-1; i++ {
		Abc_NamStrFindOrAdd(p.pStrs, s_VerTypes[i], nil)
	}
	assert(Abc_NamObjNumMax(p.pStrs) == i, "Abc_NamObjNumMax(p.pStrs) == i")
}

// character recognition
func Psr_CharIsSpace(c byte) int32  { return psrBool(c == ' ' || c == '\t' || c == '\r' || c == '\n') }
func Psr_CharIsDigit(c byte) int32  { return psrBool(c >= '0' && c <= '9') }
func Psr_CharIsDigitB(c byte) int32 { return psrBool(c == '0' || c == '1' || c == 'x' || c == 'z') }
func Psr_CharIsDigitH(c byte) int32 {
	return psrBool(c >= '0' && c <= '9' || c >= 'A' && c <= 'F' || c >= 'a' && c <= 'f')
}
func Psr_CharIsChar(c byte) int32  { return psrBool(c >= 'a' && c <= 'z' || c >= 'A' && c <= 'Z') }
func Psr_CharIsSymb1(c byte) int32 { return psrBool(Psr_CharIsChar(c) != 0 || c == '_') }
func Psr_CharIsSymb2(c byte) int32 {
	return psrBool(Psr_CharIsSymb1(c) != 0 || Psr_CharIsDigit(c) != 0 || c == '$')
}

func Psr_ManIsChar(p *Psr_Man_t, c byte) int32 {
	if len(p.pCur) > 0 && p.pCur[0] == c {
		return 1
	}
	return 0
}
func Psr_ManIsChar1(p *Psr_Man_t, c byte) int32 {
	if len(p.pCur) > 1 && p.pCur[1] == c {
		return 1
	}
	return 0
}
func Psr_ManIsDigit(p *Psr_Man_t) int32 {
	if len(p.pCur) == 0 {
		return 0
	}
	return Psr_CharIsDigit(p.pCur[0])
}

////////////////////////////////////////////////////////////////////////
///                     FUNCTION DEFINITIONS                         ///
////////////////////////////////////////////////////////////////////////

/**Function*************************************************************

  Synopsis    []

  Description []

  SideEffects []

  SeeAlso     []

***********************************************************************/

// collect predefined modules names
var s_VerilogModules = []string{
	"const0", // BAC_BOX_CF,
	"const1", // BAC_BOX_CT,
	"constX", // BAC_BOX_CX,
	"constZ", // BAC_BOX_CZ,
	"buf",    // int32(BAC_BOX_BUF),
	"not",    // int32(BAC_BOX_INV),
	"and",    // int32(BAC_BOX_AND),
	"nand",   // int32(BAC_BOX_NAND),
	"or",     // int32(BAC_BOX_OR),
	"nor",    // int32(BAC_BOX_NOR),
	"xor",    // int32(BAC_BOX_XOR),
	"xnor",   // int32(BAC_BOX_XNOR),
	"sharp",  // int32(BAC_BOX_SHARP),
	"mux",    // int32(BAC_BOX_MUX),
	"maj",    // BAC_BOX_MAJ,
}
var s_KnownModules [100]string = [100]string{
	"VERIFIC_",
	"add_",
	"mult_",
	"div_",
	"mod_",
	"rem_",
	"shift_left_",
	"shift_right_",
	"rotate_left_",
	"rotate_right_",
	"reduce_and_",
	"reduce_or_",
	"reduce_xor_",
	"reduce_nand_",
	"reduce_nor_",
	"reduce_xnor_",
	"LessThan_",
	"Mux_",
	"Select_",
	"Decoder_",
	"EnabledDecoder_",
	"PrioSelect_",
	"DualPortRam_",
	"ReadPort_",
	"WritePort_",
	"ClockedWritePort_",
	"lut",
	"and_",
	"or_",
	"xor_",
	"nand_",
	"nor_",
	"xnor_",
	"buf_",
	"inv_",
	"tri_",
	"sub_",
	"unary_minus_",
	"equal_",
	"not_equal_",
	"mux_",
	"wide_mux_",
	"wide_select_",
	"wide_dff_",
	"wide_dlatch_",
	"wide_dffrs_",
	"wide_dlatchrs_",
	"wide_prio_select_",
	"pow_",
	"PrioEncoder_",
	"abs",
	"",
}

// check if it is a Verilog predefined module
func Psr_ManIsVerilogModule(p *Psr_Man_t, pName string) int32 {
	for i := 0; i < len(s_VerilogModules) && s_VerilogModules[i] != ""; i++ {
		if strings.Compare(pName, s_VerilogModules[i]) == 0 {
			return int32(BAC_BOX_CF) + int32(i)
		}
	}
	return 0
}

// check if it is a known module
func Psr_ManIsKnownModule(p *Psr_Man_t, pName string) int32 {
	for i := 0; i < len(s_KnownModules) && s_KnownModules[i] != ""; i++ {
		if strings.HasPrefix(pName, s_KnownModules[i]) {
			return int32(i)
		}
	}
	return 0
}

/**Function*************************************************************

  Synopsis    []

  Description []

  SideEffects []

  SeeAlso     []

***********************************************************************/

// skips Verilog comments (returns 1 if some comments were skipped)
func Psr_ManUtilSkipComments(p *Psr_Man_t) int32 {
	if Psr_ManIsChar(p, '/') == 0 || len(p.pCur) < 2 {
		return 0
	}
	if Psr_ManIsChar1(p, '/') != 0 {
		p.pCur = p.pCur[2:]
		for len(p.pCur) > 0 {
			if Psr_ManIsChar(p, '\n') != 0 {
				p.pCur = p.pCur[1:]
				return 1
			}
			p.pCur = p.pCur[1:]
		}
	} else if Psr_ManIsChar1(p, '*') != 0 {
		p.pCur = p.pCur[2:]
		for len(p.pCur) > 1 {
			if Psr_ManIsChar(p, '*') != 0 && Psr_ManIsChar1(p, '/') != 0 {
				p.pCur = p.pCur[2:]
				return 1
			}
			p.pCur = p.pCur[1:]
		}
	}
	return 0
}
func Psr_ManUtilSkipName(p *Psr_Man_t) int32 {
	if Psr_ManIsChar(p, '\\') == 0 {
		return 0
	}
	p.pCur = p.pCur[1:]
	for len(p.pCur) > 0 {
		if Psr_ManIsChar(p, ' ') != 0 {
			p.pCur = p.pCur[1:]
			return 1
		}
		p.pCur = p.pCur[1:]
	}
	return 0
}

// skip any number of spaces and comments
func Psr_ManUtilSkipSpaces(p *Psr_Man_t) int32 {
	for len(p.pCur) > 0 {
		for len(p.pCur) > 0 && Psr_CharIsSpace(p.pCur[0]) != 0 {
			p.pCur = p.pCur[1:]
		}
		if len(p.pCur) == 0 || p.pCur[0] == 0 {
			return Psr_ManErrorSet(p, "Unexpectedly reached end-of-file.", 1)
		}
		if !(Psr_ManUtilSkipComments(p) != 0) {
			return 0
		}
	}
	return Psr_ManErrorSet(p, "Unexpectedly reached end-of-file.", 1)
}

// skip everything including comments until the given char
func Psr_ManUtilSkipUntil(p *Psr_Man_t, c byte) int32 {
	for len(p.pCur) > 0 {
		if Psr_ManIsChar(p, c) != 0 {
			return 1
		}
		if Psr_ManUtilSkipComments(p) != 0 {
			continue
		}
		if Psr_ManUtilSkipName(p) != 0 {
			continue
		}
		p.pCur = p.pCur[1:]
	}
	return 0
}

// skip everything including comments until the given word
func Psr_ManUtilSkipUntilWord(p *Psr_Man_t, pWord string) int32 {
	i := strings.Index(p.pCur, pWord)
	if i < 0 {
		return 1
	}
	p.pCur = p.pCur[i+len(pWord):]
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
func Psr_ManReadName(p *Psr_Man_t) int32 {
	pStart := p.pCur
	if Psr_ManIsChar(p, '\\') != 0 { // escaped name
		p.pCur = p.pCur[1:]
		pStart = p.pCur
		for len(p.pCur) > 0 && Psr_ManIsChar(p, ' ') == 0 {
			p.pCur = p.pCur[1:]
		}
	} else if len(p.pCur) > 0 && Psr_CharIsSymb1(p.pCur[0]) != 0 { // simple name
		p.pCur = p.pCur[1:]
		for len(p.pCur) > 0 && Psr_CharIsSymb2(p.pCur[0]) != 0 {
			p.pCur = p.pCur[1:]
		}
	} else {
		return 0
	}
	return Abc_NamStrFindOrAddLim(p.pStrs, pStart, len(pStart)-len(p.pCur), nil)
}
func Psr_ManReadNameList(p *Psr_Man_t, vTemp *vec.Vec_Int_t, LastSymb byte) int32 {
	Vec_IntClear(vTemp)
	for {
		var Item int32 = Psr_ManReadName(p)
		if Item == 0 {
			return Psr_ManErrorSet(p, "Cannot read name in the list.", 0)
		}
		if Psr_ManUtilSkipSpaces(p) != 0 {
			return Psr_ManErrorSet(p, "Error number 1.", 0)
		}
		if Item == int32(PRS_VER_WIRE) {
			continue
		}
		Vec_IntPush(vTemp, Item)
		if Psr_ManIsChar(p, LastSymb) != 0 {
			break
		}
		if !(Psr_ManIsChar(p, ',') != 0) {
			return Psr_ManErrorSet(p, "Expecting comma in the list.", 0)
		}
		p.pCur = p.pCur[1:]
		if Psr_ManUtilSkipSpaces(p) != 0 {
			return Psr_ManErrorSet(p, "Error number 2.", 0)
		}
	}
	return 1
}
func Psr_ManReadConstant(p *Psr_Man_t) int32 {
	var pStart string = p.pCur
	assert((Psr_ManIsDigit(p) != 0), "Psr_ManIsDigit(p)")
	for Psr_ManIsDigit(p) != 0 {
		p.pCur = p.pCur[1:]
	}
	if !(Psr_ManIsChar(p, '\'') != 0) {
		return Psr_ManErrorSet(p, "Cannot read constant.", 0)
	}
	p.pCur = p.pCur[1:]
	if Psr_ManIsChar(p, 'b') != 0 {
		p.pCur = p.pCur[1:]
		for Psr_CharIsDigitB(p.pCur[0]) != 0 {
			if p.pCur[0] == '0' {
				p.pNtk.fHasC0s = 1
			} else if p.pCur[0] == '1' {
				p.pNtk.fHasC1s = 1
			} else if p.pCur[0] == 'x' {
				p.pNtk.fHasCXs = 1
			} else if p.pCur[0] == 'z' {
				p.pNtk.fHasCZs = 1
			}
			p.pCur = p.pCur[1:]
		}

	} else if Psr_ManIsChar(p, 'h') != 0 {
		p.pCur = p.pCur[1:]
		p.pNtk.fHasC0s = 1
		for Psr_CharIsDigitH(p.pCur[0]) != 0 {
			if p.pCur[0] != '0' {
				p.pNtk.fHasC1s = 1
			}
			p.pCur = p.pCur[1:]
		}

	} else if Psr_ManIsChar(p, 'd') != 0 {
		p.pCur = p.pCur[1:]
		p.pNtk.fHasC0s = 1
		for Psr_ManIsDigit(p) != 0 {
			if p.pCur[0] != '0' {
				p.pNtk.fHasC1s = 1
			}
			p.pCur = p.pCur[1:]
		}

	} else {
		return Psr_ManErrorSet(p, "Cannot read radix of constant.", 0)
	}
	return Abc_NamStrFindOrAddLim(p.pStrs, pStart, len(pStart)-len(p.pCur), nil)
}
func Psr_ManReadRange(p *Psr_Man_t) int32 {
	assert((Psr_ManIsChar(p, '[') != 0), "Psr_ManIsChar(p, '[')")
	Vec_StrClear(&p.vCover)
	Vec_StrPush(&p.vCover, p.pCur[0])
	p.pCur = p.pCur[1:]
	if Psr_ManUtilSkipSpaces(p) != 0 {
		return Psr_ManErrorSet(p, "Error number 3.", 0)
	}
	if !(Psr_ManIsDigit(p) != 0) {
		return Psr_ManErrorSet(p, "Cannot read digit in range specification.", 0)
	}
	for Psr_ManIsDigit(p) != 0 {
		Vec_StrPush(&p.vCover, p.pCur[0])
		p.pCur = p.pCur[1:]
	}
	if Psr_ManUtilSkipSpaces(p) != 0 {
		return Psr_ManErrorSet(p, "Error number 4.", 0)
	}
	if Psr_ManIsChar(p, ':') != 0 {
		Vec_StrPush(&p.vCover, p.pCur[0])
		p.pCur = p.pCur[1:]
		if Psr_ManUtilSkipSpaces(p) != 0 {
			return Psr_ManErrorSet(p, "Error number 5.", 0)
		}
		if !(Psr_ManIsDigit(p) != 0) {
			return Psr_ManErrorSet(p, "Cannot read digit in range specification.", 0)
		}
		for Psr_ManIsDigit(p) != 0 {
			Vec_StrPush(&p.vCover, p.pCur[0])
			p.pCur = p.pCur[1:]
		}
		if Psr_ManUtilSkipSpaces(p) != 0 {
			return Psr_ManErrorSet(p, "Error number 6.", 0)
		}
	}
	if !(Psr_ManIsChar(p, ']') != 0) {
		return Psr_ManErrorSet(p, "Cannot read closing brace in range specification.", 0)
	}
	Vec_StrPush(&p.vCover, p.pCur[0])
	p.pCur = p.pCur[1:]
	Vec_StrPush(&p.vCover, 0)
	return Abc_NamStrFindOrAdd(p.pStrs, strings.TrimRight(string(Vec_StrArray(&p.vCover)), "\x00"), nil)
}
func Psr_ManReadConcat(p *Psr_Man_t, vTemp2 *vec.Vec_Int_t) int32 {
	assert((Psr_ManIsChar(p, '{') != 0), "Psr_ManIsChar(p, '{')")
	p.pCur = p.pCur[1:]
	if !(Psr_ManReadSignalList(p, vTemp2, '}', 0) != 0) {
		return Psr_ManErrorSet(p, "Error number 7.", 0)
	}
	// check final
	assert((Psr_ManIsChar(p, '}') != 0), "Psr_ManIsChar(p, '}')")
	p.pCur = p.pCur[1:]
	// return special case
	assert(Vec_IntSize(vTemp2) > 0, "Vec_IntSize(vTemp2) > 0")
	if Vec_IntSize(vTemp2) == 1 {
		return Vec_IntEntry(vTemp2, 0)
	}
	return Abc_Var2Lit2(Psr_NtkAddConcat(p.pNtk, vTemp2), int32(BAC_PRS_CONCAT))
}
func Psr_ManReadSignal(p *Psr_Man_t) int32 {
	var Item int32
	if Psr_ManUtilSkipSpaces(p) != 0 {
		return Psr_ManErrorSet(p, "Error number 8.", 0)
	}
	if Psr_ManIsDigit(p) != 0 {
		Item = Psr_ManReadConstant(p)
		if Item == 0 {
			return Psr_ManErrorSet(p, "Error number 9.", 0)
		}
		if Psr_ManUtilSkipSpaces(p) != 0 {
			return Psr_ManErrorSet(p, "Error number 10.", 0)
		}
		return Abc_Var2Lit2(Item, int32(BAC_PRS_CONST))
	}
	if Psr_ManIsChar(p, '{') != 0 {
		if p.fUsingTemp2 != 0 {
			return Psr_ManErrorSet(p, "Cannot read nested concatenations.", 0)
		}
		p.fUsingTemp2 = 1
		Item = Psr_ManReadConcat(p, &p.vTemp2)
		p.fUsingTemp2 = 0
		if Item == 0 {
			return Psr_ManErrorSet(p, "Error number 11.", 0)
		}
		if Psr_ManUtilSkipSpaces(p) != 0 {
			return Psr_ManErrorSet(p, "Error number 12.", 0)
		}
		return Item

	} else {
		Item = Psr_ManReadName(p)
		if Item == 0 {
			return Psr_ManErrorSet(p, "Error number 13.", 0)
		} // was        return 1;
		if Psr_ManUtilSkipSpaces(p) != 0 {
			return Psr_ManErrorSet(p, "Error number 14.", 0)
		}
		if Psr_ManIsChar(p, '[') != 0 {
			var Range int32 = Psr_ManReadRange(p)
			if Range == 0 {
				return Psr_ManErrorSet(p, "Error number 15.", 0)
			}
			if Psr_ManUtilSkipSpaces(p) != 0 {
				return Psr_ManErrorSet(p, "Error number 16.", 0)
			}
			return Abc_Var2Lit2(Psr_NtkAddSlice(p.pNtk, Item, Range), int32(BAC_PRS_SLICE))
		}
		return Abc_Var2Lit2(Item, int32(BAC_PRS_NAME))
	}
}
func Psr_ManReadSignalList(p *Psr_Man_t, vTemp *vec.Vec_Int_t, LastSymb byte, fAddForm int32) int32 {
	Vec_IntClear(vTemp)
	for {
		var Item int32 = Psr_ManReadSignal(p)
		if Item == 0 {
			return Psr_ManErrorSet(p, "Cannot read signal in the list.", 0)
		}
		if fAddForm != 0 {
			Vec_IntPush(vTemp, 0)
		}
		Vec_IntPush(vTemp, Item)
		if Psr_ManIsChar(p, LastSymb) != 0 {
			break
		}
		if !(Psr_ManIsChar(p, ',') != 0) {
			return Psr_ManErrorSet(p, "Expecting comma in the list.", 0)
		}
		p.pCur = p.pCur[1:]
	}
	return 1
}
func Psr_ManReadSignalList2(p *Psr_Man_t, vTemp *vec.Vec_Int_t) int32 {
	var FormId, ActItem int32
	Vec_IntClear(vTemp)
	assert((Psr_ManIsChar(p, '.') != 0), "Psr_ManIsChar(p, '.')")
	for Psr_ManIsChar(p, '.') != 0 {
		p.pCur = p.pCur[1:]
		FormId = Psr_ManReadName(p)
		if FormId == 0 {
			return Psr_ManErrorSet(p, "Cannot read formal name of the instance.", 0)
		}
		if !(Psr_ManIsChar(p, '(') != 0) {
			return Psr_ManErrorSet(p, "Cannot read \"(\" in the instance.", 0)
		}
		p.pCur = p.pCur[1:]
		if Psr_ManUtilSkipSpaces(p) != 0 {
			return Psr_ManErrorSet(p, "Error number 17.", 0)
		}
		ActItem = Psr_ManReadSignal(p)
		if ActItem == 0 {
			return Psr_ManErrorSet(p, "Cannot read actual name of the instance.", 0)
		}
		if !(Psr_ManIsChar(p, ')') != 0) {
			return Psr_ManErrorSet(p, "Cannot read \")\" in the instance.", 0)
		}
		p.pCur = p.pCur[1:]
		Vec_IntPushTwo(vTemp, FormId, ActItem)
		if Psr_ManUtilSkipSpaces(p) != 0 {
			return Psr_ManErrorSet(p, "Error number 18.", 0)
		}
		if Psr_ManIsChar(p, ')') != 0 {
			break
		}
		if !(Psr_ManIsChar(p, ',') != 0) {
			return Psr_ManErrorSet(p, "Expecting comma in the instance.", 0)
		}
		p.pCur = p.pCur[1:]
		if Psr_ManUtilSkipSpaces(p) != 0 {
			return Psr_ManErrorSet(p, "Error number 19.", 0)
		}
	}
	assert(Vec_IntSize(vTemp) > 0, "Vec_IntSize(vTemp) > 0")
	assert(Vec_IntSize(vTemp)%2 == 0, "Vec_IntSize(vTemp) % 2 == 0")
	return 1
}

/*
*Function*************************************************************

	Synopsis    []

	Description []

	SideEffects []

	SeeAlso     []

**********************************************************************
*/
func Psr_ManReadDeclaration(p *Psr_Man_t, Type int32) int32 {
	var i, NameId, RangeId int32
	var vNames [4]*vec.Vec_Int_t = [4]*vec.Vec_Int_t{&p.pNtk.vInputs, &p.pNtk.vOutputs, &p.pNtk.vInouts, &p.pNtk.vWires}
	var vNamesR [4]*vec.Vec_Int_t = [4]*vec.Vec_Int_t{&p.pNtk.vInputsR, &p.pNtk.vOutputsR, &p.pNtk.vInoutsR, &p.pNtk.vWiresR}
	assert(Type >= int32(PRS_VER_INPUT) && Type <= int32(PRS_VER_WIRE), "Type >= int32(PRS_VER_INPUT) && Type <= int32(PRS_VER_WIRE)")
	if Psr_ManUtilSkipSpaces(p) != 0 {
		return Psr_ManErrorSet(p, "Error number 20.", 0)
	}
	if Psr_ManIsChar(p, '[') != 0 {
		RangeId = Psr_ManReadRange(p)
		if RangeId == 0 {
			return Psr_ManErrorSet(p, "Error number 21.", 0)
		}
	}
	if !(Psr_ManReadNameList(p, &p.vTemp, ';') != 0) {
		return Psr_ManErrorSet(p, "Error number 22.", 0)
	}
	// inlining of macro Vec_IntForEachEntry
	for i = 0; i < Vec_IntSize(&p.vTemp); i++ {
		NameId = Vec_IntEntry(&p.vTemp, i)
		Vec_IntPush(vNames[Type-int32(PRS_VER_INPUT)], NameId)
		Vec_IntPush(vNamesR[Type-int32(PRS_VER_INPUT)], RangeId)
		if Type < int32(PRS_VER_WIRE) {
			Vec_IntPush(&p.pNtk.vOrder, Abc_Var2Lit2(NameId, Type))
		}
	}
	return 1
}
func Psr_ManReadAssign(p *Psr_Man_t) int32 {
	var OutItem, InItem, Oper int32
	var fCompl, fCompl2 bool
	// read output name
	OutItem = Psr_ManReadSignal(p)
	if OutItem == 0 {
		return Psr_ManErrorSet(p, "Cannot read output in assign-statement.", 0)
	}
	if !(Psr_ManIsChar(p, '=') != 0) {
		return Psr_ManErrorSet(p, "Expecting \"=\" in assign-statement.", 0)
	}
	p.pCur = p.pCur[1:]
	if Psr_ManUtilSkipSpaces(p) != 0 {
		return Psr_ManErrorSet(p, "Error number 23.", 0)
	}
	if Psr_ManIsChar(p, '~') != 0 {
		fCompl = true
		p.pCur = p.pCur[1:]
	}
	// read first name
	InItem = Psr_ManReadSignal(p)
	if InItem == 0 {
		return Psr_ManErrorSet(p, "Cannot read first input name in the assign-statement.", 0)
	}
	Vec_IntClear(&p.vTemp)
	Vec_IntPush(&p.vTemp, 0)
	Vec_IntPush(&p.vTemp, InItem)
	// check unary operator
	if Psr_ManIsChar(p, ';') != 0 {
		Vec_IntPush(&p.vTemp, 0)
		Vec_IntPush(&p.vTemp, OutItem)
		if fCompl {
			Oper = int32(BAC_BOX_INV)
		} else {
			Oper = int32(BAC_BOX_BUF)
		}
		Psr_NtkAddBox(p.pNtk, Oper, 0, &p.vTemp)
		return 1
	}
	if Psr_ManIsChar(p, '&') != 0 {
		Oper = int32(BAC_BOX_AND)
	} else if Psr_ManIsChar(p, '|') != 0 {
		Oper = int32(BAC_BOX_OR)
	} else if Psr_ManIsChar(p, '^') != 0 {
		Oper = int32(BAC_BOX_XOR)
	} else if Psr_ManIsChar(p, '?') != 0 {
		Oper = int32(BAC_BOX_MUX)
	} else {
		return Psr_ManErrorSet(p, "Unrecognized operator in the assign-statement.", 0)
	}
	p.pCur = p.pCur[1:]
	if Psr_ManUtilSkipSpaces(p) != 0 {
		return Psr_ManErrorSet(p, "Error number 24.", 0)
	}
	if Psr_ManIsChar(p, '~') != 0 {
		fCompl2 = true
		p.pCur = p.pCur[1:]
	}
	// read second name
	InItem = Psr_ManReadSignal(p)
	if InItem == 0 {
		return Psr_ManErrorSet(p, "Cannot read second input name in the assign-statement.", 0)
	}
	Vec_IntPush(&p.vTemp, 0)
	Vec_IntPush(&p.vTemp, InItem)
	// read third argument
	if Oper == int32(BAC_BOX_MUX) {
		assert(!fCompl, "!fCompl")
		if !(Psr_ManIsChar(p, ':') != 0) {
			return Psr_ManErrorSet(p, "Expected colon in the MUX assignment.", 0)
		}
		p.pCur = p.pCur[1:]
		// read third name
		InItem = Psr_ManReadSignal(p)
		if InItem == 0 {
			return Psr_ManErrorSet(p, "Cannot read third input name in the assign-statement.", 0)
		}
		Vec_IntPush(&p.vTemp, 0)
		Vec_IntPush(&p.vTemp, InItem)
		if !(Psr_ManIsChar(p, ';') != 0) {
			return Psr_ManErrorSet(p, "Expected semicolon at the end of the assign-statement.", 0)
		}

	} else {
		// figure out operator
		if Oper == int32(BAC_BOX_AND) {
			if fCompl && !fCompl2 {
				Oper = int32(BAC_BOX_SHARPL)
			} else if !fCompl && fCompl2 {
				Oper = int32(BAC_BOX_SHARP)
			} else if fCompl && fCompl2 {
				Oper = int32(BAC_BOX_NOR)
			}

		} else if Oper == int32(BAC_BOX_OR) {
			if fCompl && fCompl2 {
				Oper = int32(BAC_BOX_NAND)
			} else {
				assert(!fCompl && !fCompl2, "!fCompl && !fCompl2")
			}

		} else if Oper == int32(BAC_BOX_XOR) {
			if fCompl && !fCompl2 {
				Oper = int32(BAC_BOX_XNOR)
			} else {
				assert(!fCompl && !fCompl2, "!fCompl && !fCompl2")
			}
		}
	}
	// write binary operator
	Vec_IntPush(&p.vTemp, 0)
	Vec_IntPush(&p.vTemp, OutItem)
	Psr_NtkAddBox(p.pNtk, Oper, 0, &p.vTemp)
	return 1
}
func Psr_ManReadInstance(p *Psr_Man_t, Func int32) int32 {
	var InstId, Status int32
	/*
	   static Counter = 0;
	   if ( ++Counter == 7 )
	   {
	       int s=0;
	   }
	*/
	if Psr_ManUtilSkipSpaces(p) != 0 {
		return Psr_ManErrorSet(p, "Error number 25.", 0)
	}
	InstId = Psr_ManReadName(p)
	if InstId != 0 {
		if Psr_ManUtilSkipSpaces(p) != 0 {
			return Psr_ManErrorSet(p, "Error number 26.", 0)
		}
	}
	if !(Psr_ManIsChar(p, '(') != 0) {
		return Psr_ManErrorSet(p, "Expecting \"(\" in module instantiation.", 0)
	}
	p.pCur = p.pCur[1:]
	if Psr_ManUtilSkipSpaces(p) != 0 {
		return Psr_ManErrorSet(p, "Error number 27.", 0)
	}
	if Psr_ManIsChar(p, '.') != 0 { // box
		Status = Psr_ManReadSignalList2(p, &p.vTemp)
	} else { // node

		//char * s = Abc_NamStr(p->pStrs, Func);
		// translate elementary gate
		var iFuncNew int32 = Psr_ManIsVerilogModule(p, Abc_NamStr(p.pStrs, Func))
		if iFuncNew == 0 {
			return Psr_ManErrorSet(p, "Cannot find elementary gate.", 0)
		}
		Func = iFuncNew
		Status = Psr_ManReadSignalList(p, &p.vTemp, ')', 1)
	}
	if Status == 0 {
		return Psr_ManErrorSet(p, "Error number 28.", 0)
	}
	assert((Psr_ManIsChar(p, ')') != 0), "Psr_ManIsChar(p, ')')")
	p.pCur = p.pCur[1:]
	if Psr_ManUtilSkipSpaces(p) != 0 {
		return Psr_ManErrorSet(p, "Error number 29.", 0)
	}
	if !(Psr_ManIsChar(p, ';') != 0) {
		return Psr_ManErrorSet(p, "Expecting semicolon in the instance.", 0)
	}
	// add box
	Psr_NtkAddBox(p.pNtk, Func, InstId, &p.vTemp)
	return 1
}
func Psr_ManReadArguments(p *Psr_Man_t) int32 {
	var iRange int32
	var iType int32 = -1
	var vSigs [3]*vec.Vec_Int_t = [3]*vec.Vec_Int_t{&p.pNtk.vInputs, &p.pNtk.vOutputs, &p.pNtk.vInouts}
	var vSigsR [3]*vec.Vec_Int_t = [3]*vec.Vec_Int_t{&p.pNtk.vInputsR, &p.pNtk.vOutputsR, &p.pNtk.vInoutsR}
	assert((Psr_ManIsChar(p, '(') != 0), "Psr_ManIsChar(p, '(')")
	p.pCur = p.pCur[1:]
	if Psr_ManUtilSkipSpaces(p) != 0 {
		return Psr_ManErrorSet(p, "Error number 30.", 0)
	}
	for {
		var iName int32 = Psr_ManReadName(p)
		if iName == 0 {
			return Psr_ManErrorSet(p, "Error number 31.", 0)
		}
		if Psr_ManUtilSkipSpaces(p) != 0 {
			return Psr_ManErrorSet(p, "Error number 32.", 0)
		}
		if iName >= int32(PRS_VER_INPUT) && iName <= int32(PRS_VER_INOUT) { // declaration

			iType = iName
			if Psr_ManIsChar(p, '[') != 0 {
				iRange = Psr_ManReadRange(p)
				if iRange == 0 {
					return Psr_ManErrorSet(p, "Error number 33.", 0)
				}
				if Psr_ManUtilSkipSpaces(p) != 0 {
					return Psr_ManErrorSet(p, "Error number 34.", 0)
				}
			}
			iName = Psr_ManReadName(p)
			if iName == 0 {
				return Psr_ManErrorSet(p, "Error number 35.", 0)
			}
		}
		if iType > 0 {
			Vec_IntPush(vSigs[iType-int32(PRS_VER_INPUT)], iName)
			Vec_IntPush(vSigsR[iType-int32(PRS_VER_INPUT)], iRange)
			Vec_IntPush(&p.pNtk.vOrder, Abc_Var2Lit2(iName, iType))
		}
		if Psr_ManIsChar(p, ')') != 0 {
			break
		}
		if !(Psr_ManIsChar(p, ',') != 0) {
			return Psr_ManErrorSet(p, "Expecting comma in the instance.", 0)
		}
		p.pCur = p.pCur[1:]
		if Psr_ManUtilSkipSpaces(p) != 0 {
			return Psr_ManErrorSet(p, "Error number 36.", 0)
		}
	}
	// check final
	assert((Psr_ManIsChar(p, ')') != 0), "Psr_ManIsChar(p, ')')")
	return 1
}

// this procedure can return:
// 0 = reached end-of-file; 1 = successfully parsed; 2 = recognized as primitive; 3 = failed and skipped; 4 = error (failed and could not skip)
func Psr_ManReadModule(p *Psr_Man_t) int32 {
	var iToken, Status int32
	if p.pNtk != nil {
		return Psr_ManErrorSet(p, "Parsing previous module is unfinished.", 4)
	}
	if Psr_ManUtilSkipSpaces(p) != 0 {
		Psr_ManErrorClear(p)
		return 0
	}
	// read keyword
	iToken = Psr_ManReadName(p)
	if iToken != int32(PRS_VER_MODULE) {
		return Psr_ManErrorSet(p, "Cannot read \"module\" keyword.", 4)
	}
	if Psr_ManUtilSkipSpaces(p) != 0 {
		return 4
	}
	// read module name
	iToken = Psr_ManReadName(p)
	if iToken == 0 {
		return Psr_ManErrorSet(p, "Cannot read module name.", 4)
	}
	if Psr_ManIsKnownModule(p, Abc_NamStr(p.pStrs, iToken)) != 0 {
		if Psr_ManUtilSkipUntilWord(p, "endmodule") != 0 {
			return Psr_ManErrorSet(p, "Cannot find \"endmodule\" keyword.", 4)
		}
		//printf( "Warning! Skipped known module \"%s\".\n", Abc_NamStr(p->pStrs, iToken) );
		Vec_IntPush(&p.vKnown, iToken)
		return 2
	}
	Psr_ManInitializeNtk(p, iToken, 1)
	// skip arguments
	if Psr_ManUtilSkipSpaces(p) != 0 {
		return 4
	}
	if !(Psr_ManIsChar(p, '(') != 0) {
		return Psr_ManErrorSet(p, "Cannot find \"(\" in the argument declaration.", 4)
	}
	if !(Psr_ManReadArguments(p) != 0) {
		return 4
	}
	assert(p.pCur[0] == ')', "p.pCur[0] == ')'")
	p.pCur = p.pCur[1:]
	if Psr_ManUtilSkipSpaces(p) != 0 {
		return 4
	}
	// read declarations and instances
	for Psr_ManIsChar(p, ';') != 0 {
		p.pCur = p.pCur[1:]
		if Psr_ManUtilSkipSpaces(p) != 0 {
			return 4
		}
		iToken = Psr_ManReadName(p)
		if iToken == int32(PRS_VER_ENDMODULE) {
			Vec_IntPush(&p.vSucceeded, p.pNtk.iModuleName)
			Psr_ManFinalizeNtk(p)
			return 1
		}
		if iToken >= int32(PRS_VER_INPUT) && iToken <= int32(PRS_VER_WIRE) { // declaration
			Status = Psr_ManReadDeclaration(p, iToken)
		} else if iToken == int32(PRS_VER_REG) || iToken == int32(PRS_VER_DEFPARAM) { // unsupported keywords
			Status = Psr_ManUtilSkipUntil(p, ';')
		} else { // read instance

			if iToken == int32(PRS_VER_ASSIGN) {
				Status = Psr_ManReadAssign(p)
			} else {
				Status = Psr_ManReadInstance(p, iToken)
			}
			if Status == 0 {
				if Psr_ManUtilSkipUntilWord(p, "endmodule") != 0 {
					return Psr_ManErrorSet(p, "Cannot find \"endmodule\" keyword.", 4)
				}
				//printf( "Warning! Failed to parse \"%s\". Adding module \"%s\" as blackbox.\n",
				//    Abc_NamStr(p->pStrs, iToken), Abc_NamStr(p->pStrs, p->pNtk->iModuleName) );
				Vec_IntPush(&p.vFailed, p.pNtk.iModuleName)
				// cleanup
				Vec_IntErase(&p.pNtk.vWires)
				Vec_IntErase(&p.pNtk.vWiresR)
				Vec_IntErase(&p.pNtk.vSlices)
				Vec_IntErase(&p.pNtk.vConcats)
				Vec_IntErase(&p.pNtk.vBoxes)
				Vec_IntErase(&p.pNtk.vObjs)
				p.fUsingTemp2 = 0
				// add
				Psr_ManFinalizeNtk(p)
				Psr_ManErrorClear(p)
				return 3
			}
		}
		if !(Status != 0) {
			return 4
		}
		if Psr_ManUtilSkipSpaces(p) != 0 {
			return 4
		}
	}
	return Psr_ManErrorSet(p, "Cannot find \";\" in the module definition.", 4)
}
func Psr_ManReadDesign(p *Psr_Man_t) int32 {
	for {
		var RetValue int32 = Psr_ManReadModule(p)
		if RetValue == 0 { // end of file
			break
		}
		if RetValue == 1 { // successfully parsed
			continue
		}
		if RetValue == 2 { // recognized as primitive
			continue
		}
		if RetValue == 3 { // failed and skipped
			continue
		}
		if RetValue == 4 { // error
			return 0
		}
		assert(false, "0")
	}
	return 1
}

/*
*Function*************************************************************

	Synopsis    []

	Description []

	SideEffects []

	SeeAlso     []

**********************************************************************
*/
func Psr_ManPrintModules(p *Psr_Man_t) {
	var pName string
	var i int32
	fmt.Printf("Succeeded parsing %d models:\n", Vec_IntSize(&p.vSucceeded))
	// inlining of macro Psr_ManForEachNameVec
	for i = 0; i < Vec_IntSize(&p.vSucceeded); i++ {
		pName = Abc_NamStr(p.pStrs, Vec_IntEntry(&p.vSucceeded, i))
		fmt.Printf(" %s", pName)
	}
	fmt.Printf("\n")
	fmt.Printf("Skipped %d known models:\n", Vec_IntSize(&p.vKnown))
	// inlining of macro Psr_ManForEachNameVec
	for i = 0; i < Vec_IntSize(&p.vKnown); i++ {
		pName = Abc_NamStr(p.pStrs, Vec_IntEntry(&p.vKnown, i))
		fmt.Printf(" %s", pName)
	}
	fmt.Printf("\n")
	fmt.Printf("Skipped %d failed models:\n", Vec_IntSize(&p.vFailed))
	// inlining of macro Psr_ManForEachNameVec
	for i = 0; i < Vec_IntSize(&p.vFailed); i++ {
		pName = Abc_NamStr(p.pStrs, Vec_IntEntry(&p.vFailed, i))
		fmt.Printf(" %s", pName)
	}
	fmt.Printf("\n")
}

/*
*Function*************************************************************

	Synopsis    []

	Description []

	SideEffects []

	SeeAlso     []

**********************************************************************
*/
func Psr_ManReadVerilog(pFileName string) *vec.Vec_Ptr_t {
	var vPrs *vec.Vec_Ptr_t = nil
	var p *Psr_Man_t = Psr_ManAlloc(pFileName)
	if p == nil {
		return nil
	}
	Psr_NtkAddVerilogDirectives(p)
	Psr_ManReadDesign(p)
	//Psr_ManPrintModules( p );
	if Psr_ManErrorPrint(p) != 0 {
		// inlining of macro ABC_SWAP
		var t *vec.Vec_Ptr_t = vPrs
		vPrs = p.vNtks
		p.vNtks = t
	}
	Psr_ManFree(p)
	return vPrs
}

func Psr_ManReadVerilogTest(pFileName string) {
	var clk abctime = Abc_Clock()
	var vPrs *vec.Vec_Ptr_t = Psr_ManReadVerilog("c/hie/dump/1/netlist_1.v")
	//    Vec_Ptr_t * vPrs = Psr_ManReadVerilog( "aga/me/me_wide.v" );
	//    Vec_Ptr_t * vPrs = Psr_ManReadVerilog( "aga/ray/ray_wide.v" );
	if !(vPrs != nil) {
		return
	}
	fmt.Printf("Finished reading %d networks. ", Vec_PtrSize(vPrs))
	fmt.Printf("NameIDs = %d. ", Abc_NamObjNumMax(Psr_ManNameMan(vPrs)))
	fmt.Printf("Memory = %.2f MB. ", float64(Psr_ManMemory(vPrs))/(1<<20))
	Abc_PrintTime(1, "Time", Abc_Clock()-clk)
	Psr_ManWriteVerilog("c/hie/dump/1/netlist_1_out_new.v", vPrs)
	//    Psr_ManWriteVerilog( "aga/me/me_wide_out.v", vPrs );
	//    Psr_ManWriteVerilog( "aga/ray/ray_wide_out.v", vPrs );
	//    Abc_NamPrint( p->pStrs );
	Psr_ManVecFree(vPrs)
}

////////////////////////////////////////////////////////////////////////
///                       END OF FILE                                ///
////////////////////////////////////////////////////////////////////////

//ABC_NAMESPACE_IMPL_END
