package cba

import (
	"fmt"
	"github.com/glycerine/abcgo/misc/util"
	"strconv"
	"strings"
)

/**CFile****************************************************************

  FileName    [cbaReadVer.c]

  SystemName  [ABC: Logic synthesis and verification system.]

  PackageName [Hierarchical word-level netlist.]

  Synopsis    [BLIF writer.]

  Author      [Alan Mishchenko]

  Affiliation [UC Berkeley]

  Date        [Ver. 1.0. Started - November 29, 2014.]

  Revision    [$Id: cbaReadVer.c,v 1.00 2014/11/29 00:00:00 alanmi Exp $]

***********************************************************************/

//#include "cba.h"
//#include "cbaPrs.h"

//ABC_NAMESPACE_IMPL_START

////////////////////////////////////////////////////////////////////////
///                        DECLARATIONS                              ///
////////////////////////////////////////////////////////////////////////

var s_VerTypes = []string{
	"",          // 0: unused
	"input",     // 1:  input
	"output",    // 2:  output
	"inout",     // 3:  inout
	"wire",      // 4:  wire
	"reg",       // 5:  reg
	"module",    // 6:  module
	"assign",    // 7:  assign
	"always",    // 8:  always
	"function",  // 9:  function
	"defparam",  // 10: defparam
	"begin",     // 11: begin
	"end",       // 12: end
	"case",      // 13: case
	"endcase",   // 14: endcase
	"signed",    // 15: signed
	"endmodule", // 16: endmodule
	"",          // 17: unknown
}

func Prs_NtkAddVerilogDirectives(p *Prs_Man_t) {
	var i int32
	for i = 1; i < int32(len(s_VerTypes)-1); i++ {
		Abc_NamStrFindOrAdd(p.pStrs, s_VerTypes[i], nil)
	}
	assert(Abc_NamObjNumMax(p.pStrs) == i, "Abc_NamObjNumMax(p.pStrs) == i")
}

// Prs_ManAt and Prs_ManPeek model C pointer reads while keeping the cursor as
// an explicit byte offset. Reads beyond the loaded buffer behave like NUL.
func Prs_ManAt(p *Prs_Man_t, offset int) byte {
	if offset < 0 || offset >= len(p.pBuffer) {
		return 0
	}
	return p.pBuffer[offset]
}
func Prs_ManPeek(p *Prs_Man_t, delta int) byte { return Prs_ManAt(p, p.pCur+delta) }
func Prs_ManSlice(p *Prs_Man_t, start, end int) string {
	if start < 0 {
		start = 0
	}
	if end > len(p.pBuffer) {
		end = len(p.pBuffer)
	}
	if end < start {
		return ""
	}
	return p.pBuffer[start:end]
}
func Prs_ManIsChar(p *Prs_Man_t, c byte) int32  { return boolToInt(Prs_ManPeek(p, 0) == c) }
func Prs_ManIsChar1(p *Prs_Man_t, c byte) int32 { return boolToInt(Prs_ManPeek(p, 1) == c) }
func Prs_ManIsDigit(p *Prs_Man_t) int32         { return Prs_CharIsDigit(Prs_ManPeek(p, 0)) }

////////////////////////////////////////////////////////////////////////
///                     FUNCTION DEFINITIONS                         ///
////////////////////////////////////////////////////////////////////////

/**Function*************************************************************

  Synopsis    []

  Description []

  SideEffects []

  SeeAlso     []

***********************************************************************/

// predefined primitives

type Prs_VerPrim_t struct {
	Type  int32
	pName string
}

var s_VerilogPrims [16]Prs_VerPrim_t = [16]Prs_VerPrim_t{
	{int32(CBA_BOX_BUF), "buf"},
	{int32(CBA_BOX_INV), "not"},
	{int32(CBA_BOX_AND), "and"},
	{int32(CBA_BOX_NAND), "nand"},
	{int32(CBA_BOX_OR), "or"},
	{int32(CBA_BOX_NOR), "nor"},
	{int32(CBA_BOX_XOR), "xor"},
	{int32(CBA_BOX_XNOR), "xnor"},
	{int32(CBA_BOX_TRI), "bufif1"},
	{0, ""},
}

// predefined operator names
var s_VerNames [100]string = [100]string{
	"",
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
	"abs_",
	"CPL_NMACROFF",
	"CPL_MACROFF",
	"CPL_FF",
	"",
}

type Prs_VerInfo_t struct {
	Type      int32
	nInputs   int32
	pTypeName string
	pSigNames [6]string
}

var s_VerInfo = []Prs_VerInfo_t{
	{-1, 0, "" /* "PRIM_NONE"               */, [6]string{""}},
	{int32(CBA_BOX_CT), 0, "VERIFIC_PWR" /* "PRIM_PWR"                */, [6]string{"o"}},
	{int32(CBA_BOX_CF), 0, "VERIFIC_GND" /* "PRIM_GND"                */, [6]string{"o"}},
	{int32(CBA_BOX_CX), 0, "VERIFIC_X" /* "PRIM_X"                  */, [6]string{"o"}},
	{int32(CBA_BOX_CZ), 0, "VERIFIC_Z" /* "PRIM_Z"                  */, [6]string{"o"}},
	{int32(CBA_BOX_INV), 1, "VERIFIC_INV" /* "PRIM_INV"                */, [6]string{"i", "o"}},
	{int32(CBA_BOX_BUF), 1, "VERIFIC_BUF" /* "PRIM_BUF"                */, [6]string{"i", "o"}},
	{int32(CBA_BOX_AND), 1, "VERIFIC_AND" /* "PRIM_AND"                */, [6]string{"a0", "a1", "o"}},
	{int32(CBA_BOX_NAND), 2, "VERIFIC_NAND" /* "PRIM_NAND"               */, [6]string{"a0", "a1", "o"}},
	{int32(CBA_BOX_OR), 2, "VERIFIC_OR" /* "PRIM_OR"                 */, [6]string{"a0", "a1", "o"}},
	{int32(CBA_BOX_NOR), 2, "VERIFIC_NOR" /* "PRIM_NOR"                */, [6]string{"a0", "a1", "o"}},
	{int32(CBA_BOX_XOR), 2, "VERIFIC_XOR" /* "PRIM_XOR"                */, [6]string{"a0", "a1", "o"}},
	{int32(CBA_BOX_XNOR), 2, "VERIFIC_XNOR" /* "PRIM_XNOR"               */, [6]string{"a0", "a1", "o"}},
	{int32(CBA_BOX_MUX), 3, "VERIFIC_MUX" /* "PRIM_MUX"                */, [6]string{"c", "a1", "a0", "o"}}, // changed order
	{-1, 0, "VERIFIC_PULLUP" /* "PRIM_PULLUP"             */, [6]string{"o"}},
	{-1, 0, "VERIFIC_PULLDOWN" /* "PRIM_PULLDOWN"           */, [6]string{"o"}},
	{int32(CBA_BOX_TRI), 3, "VERIFIC_TRI" /* "PRIM_TRI"                */, [6]string{"i", "c", "o"}},
	{int32(CBA_BOX_LATCHRS), 4, "VERIFIC_DLATCHRS" /* "PRIM_DLATCHRS"           */, [6]string{"d", "s", "r", "gate", "q"}},              // changed order
	{int32(CBA_BOX_LATCH), 4, "VERIFIC_DLATCH" /* "PRIM_DLATCH"             */, [6]string{"d", "async_val", "async_cond", "gate", "q"}}, // changed order
	{int32(CBA_BOX_DFFRS), 4, "VERIFIC_DFFRS" /* "PRIM_DFFRS"              */, [6]string{"d", "s", "r", "clk", "q"}},                    // changed order
	{int32(CBA_BOX_DFF), 4, "VERIFIC_DFF" /* "PRIM_DFF"                */, [6]string{"d", "async_val", "async_cond", "clk", "q"}},       // changed order
	{-1, 2, "VERIFIC_NMOS" /* "PRIM_NMOS"               */, [6]string{"c", "d", "o"}},
	{-1, 2, "VERIFIC_PMOS" /* "PRIM_PMOS"               */, [6]string{"c", "d", "o"}},
	{-1, 3, "VERIFIC_CMOS" /* "PRIM_CMOS"               */, [6]string{"d", "nc", "pc", "o"}},
	{-1, 2, "VERIFIC_TRAN" /* "PRIM_TRAN"               */, [6]string{"inout1", "inout2", "control"}},
	{int32(CBA_BOX_ADD), 3, "VERIFIC_FADD" /* "PRIM_FADD"               */, [6]string{"cin", "a", "b", "o", "cout"}},
	{-1, 3, "VERIFIC_RCMOS" /* "PRIM_RCMOS"              */, [6]string{"d", "nc", "pc", "o"}},
	{-1, 2, "VERIFIC_RNMOS" /* "PRIM_RNMOS"              */, [6]string{"c", "d", "o"}},
	{-1, 2, "VERIFIC_RPMOS" /* "PRIM_RPMOS"              */, [6]string{"c", "d", "o"}},
	{-1, 2, "VERIFIC_RTRAN" /* "PRIM_RTRAN"              */, [6]string{"inout1", "inout2", "control"}},
	{-1, 0, "VERIFIC_HDL_ASSERTION" /* "PRIM_HDL_ASSERTION"      */, [6]string{"condition"}},
	{int32(CBA_BOX_ADD), 3, "add_" /* "OPER_ADDER"              */, [6]string{"cin", "a", "b", "o", "cout"}},
	{int32(CBA_BOX_MUL), 2, "mult_" /* "OPER_MULTIPLIER"         */, [6]string{"a", "b", "o"}},
	{int32(CBA_BOX_DIV), 2, "div_" /* "OPER_DIVIDER"            */, [6]string{"a", "b", "o"}},
	{int32(CBA_BOX_MOD), 2, "mod_" /* "OPER_MODULO"             */, [6]string{"a", "b", "o"}},
	{int32(CBA_BOX_REM), 2, "rem_" /* "OPER_REMAINDER"          */, [6]string{"a", "b", "o"}},
	{int32(CBA_BOX_SHIL), 3, "shift_left_" /* "OPER_SHIFT_LEFT"         */, [6]string{"cin", "a", "amount", "o"}},
	{int32(CBA_BOX_SHIR), 3, "shift_right_" /* "OPER_SHIFT_RIGHT"        */, [6]string{"cin", "a", "amount", "o"}},
	{int32(CBA_BOX_ROTL), 2, "rotate_left_" /* "OPER_ROTATE_LEFT"        */, [6]string{"a", "amount", "o"}},
	{int32(CBA_BOX_ROTR), 2, "rotate_right_" /* "OPER_ROTATE_RIGHT"       */, [6]string{"a", "amount", "o"}},
	{int32(CBA_BOX_RAND), 1, "reduce_and_" /* "OPER_REDUCE_AND"         */, [6]string{"a", "o"}},
	{int32(CBA_BOX_ROR), 1, "reduce_or_" /* "OPER_REDUCE_OR"          */, [6]string{"a", "o"}},
	{int32(CBA_BOX_RXOR), 1, "reduce_xor_" /* "OPER_REDUCE_XOR"         */, [6]string{"a", "o"}},
	{int32(CBA_BOX_RNAND), 1, "reduce_nand_" /* "OPER_REDUCE_NAND"        */, [6]string{"a", "o"}},
	{int32(CBA_BOX_RNOR), 1, "reduce_nor_" /* "OPER_REDUCE_NOR"         */, [6]string{"a", "o"}},
	{int32(CBA_BOX_RXNOR), 1, "reduce_xnor_" /* "OPER_REDUCE_XNOR"        */, [6]string{"a", "o"}},
	{int32(CBA_BOX_LTHAN), 3, "LessThan_" /* "OPER_LESSTHAN"           */, [6]string{"cin", "a", "b", "o"}},
	{int32(CBA_BOX_NMUX), 2, "Mux_" /* "OPER_NTO1MUX"            */, [6]string{"sel", "data", "o"}},
	{int32(CBA_BOX_SEL), 2, "Select_" /* "OPER_SELECTOR"           */, [6]string{"sel", "data", "o"}},
	{int32(CBA_BOX_DEC), 1, "Decoder_" /* "OPER_DECODER"            */, [6]string{"a", "o"}},
	{int32(CBA_BOX_EDEC), 2, "EnabledDecoder_" /* "OPER_ENABLED_DECODER"    */, [6]string{"en", "i", "o"}},
	{int32(CBA_BOX_PSEL), 3, "PrioSelect_" /* "OPER_PRIO_SELECTOR"      */, [6]string{"cin", "sel", "data", "o"}},
	{int32(CBA_BOX_RAM), 4, "DualPortRam_" /* "OPER_DUAL_PORT_RAM"      */, [6]string{"write_enable", "write_address", "write_data", "read_address", "read_data"}},
	{int32(CBA_BOX_RAMR), 3, "ReadPort_" /* "OPER_READ_PORT"          */, [6]string{"read_enable", "read_address", "Ram", "read_data"}},
	{int32(CBA_BOX_RAMW), 3, "WritePort_" /* "OPER_WRITE_PORT"         */, [6]string{"write_enable", "write_address", "write_data", "Ram"}},
	{int32(CBA_BOX_RAMWC), 4, "ClockedWritePort_" /* "OPER_CLOCKED_WRITE_PORT" */, [6]string{"clk", "write_enable", "write_address", "write_data", "Ram"}},
	{int32(CBA_BOX_LUT), 1, "lut" /* "OPER_LUT"                */, [6]string{"i", "o"}},
	{int32(CBA_BOX_AND), 2, "and_" /* "OPER_WIDE_AND"           */, [6]string{"a", "b", "o"}},
	{int32(CBA_BOX_OR), 2, "or_" /* "OPER_WIDE_OR"            */, [6]string{"a", "b", "o"}},
	{int32(CBA_BOX_XOR), 2, "xor_" /* "OPER_WIDE_XOR"           */, [6]string{"a", "b", "o"}},
	{int32(CBA_BOX_NAND), 2, "nand_" /* "OPER_WIDE_NAND"          */, [6]string{"a", "b", "o"}},
	{int32(CBA_BOX_NOR), 2, "nor_" /* "OPER_WIDE_NOR"           */, [6]string{"a", "b", "o"}},
	{int32(CBA_BOX_XNOR), 2, "xnor_" /* "OPER_WIDE_XNOR"          */, [6]string{"a", "b", "o"}},
	{int32(CBA_BOX_BUF), 1, "buf_" /* "OPER_WIDE_BUF"           */, [6]string{"i", "o"}},
	{int32(CBA_BOX_INV), 1, "inv_" /* "OPER_WIDE_INV"           */, [6]string{"i", "o"}},
	{int32(CBA_BOX_TRI), 2, "tri_" /* "OPER_WIDE_TRI"           */, [6]string{"i", "c", "o"}},
	{int32(CBA_BOX_SUB), 2, "sub_" /* "OPER_MINUS"              */, [6]string{"a", "b", "o"}},
	{int32(CBA_BOX_MIN), 1, "unary_minus_" /* "OPER_UMINUS"             */, [6]string{"i", "o"}},
	{int32(CBA_BOX_EQU), 2, "equal_" /* "OPER_EQUAL"              */, [6]string{"a", "b", "o"}},
	{int32(CBA_BOX_NEQU), 2, "not_equal_" /* "OPER_NEQUAL"             */, [6]string{"a", "b", "o"}},
	{int32(CBA_BOX_MUX), 3, "mux_" /* "OPER_WIDE_MUX"           */, [6]string{"cond", "d1", "d0", "o"}}, // changed order
	{int32(CBA_BOX_NMUX), 2, "wide_mux_" /* "OPER_WIDE_NTO1MUX"       */, [6]string{"sel", "data", "o"}},
	{int32(CBA_BOX_SEL), 2, "wide_select_" /* "OPER_WIDE_SELECTOR"      */, [6]string{"sel", "data", "o"}},
	{int32(CBA_BOX_DFF), 4, "wide_dff_" /* "OPER_WIDE_DFF"           */, [6]string{"d", "async_val", "async_cond", "clock", "q"}},
	{int32(CBA_BOX_DFFRS), 4, "wide_dffrs_" /* "OPER_WIDE_DFFRS"         */, [6]string{"d", "set", "reset", "clock", "q"}},
	{int32(CBA_BOX_LATCHRS), 4, "wide_dlatchrs_" /* "OPER_WIDE_DLATCHRS"      */, [6]string{"d", "set", "reset", "clock", "q"}},
	{int32(CBA_BOX_LATCH), 4, "wide_dlatch_" /* "OPER_WIDE_DLATCH"        */, [6]string{"d", "async_val", "async_cond", "clock", "q"}},
	{int32(CBA_BOX_PSEL), 3, "wide_prio_select_" /* "OPER_WIDE_PRIO_SELECTOR" */, [6]string{"sel", "data", "carry_in", "o"}},
	{int32(CBA_BOX_POW), 2, "pow_" /* "OPER_POW"                */, [6]string{"a", "b", "o"}},
	{int32(CBA_BOX_PENC), 1, "PrioEncoder_" /* "OPER_PRIO_ENCODER"       */, [6]string{"sel", "o"}},
	{int32(CBA_BOX_ABS), 1, "abs_" /* "OPER_ABS"                */, [6]string{"i", "o"}},
	{int32(CBA_BOX_DFFCPL), 4, "CPL_FF" /* "OPER_WIDE_DFF - 2"       */, [6]string{"d", "arstval", "arst", "clk", "q", "qbar"}},
	{-1, 0, "" /* "PRIM_END"                */, [6]string{""}},
}

// check if it is a Verilog predefined module
func Prs_ManIsVerilogPrim(pName string) int32 {
	var i int32
	for i = 0; i < int32(len(s_VerilogPrims)) && s_VerilogPrims[i].pName != ""; i++ {
		if strings.Compare(pName, s_VerilogPrims[i].pName) == 0 {
			return s_VerilogPrims[i].Type
		}
	}
	return 0
}

// check if it is a known module
func Prs_ManIsKnownModule(pName string) int32 {
	var i, Length int32
	for i = int32(1); i < int32(len(s_VerNames)) && s_VerNames[i] != ""; i++ {
		Length = int32(len(s_VerNames[i]))
		//        if ( !strncmp(pName, s_VerNames[i], Length) && (i == 1 || (pName[Length] >= '0' && pName[Length] <= '9')) )
		if strncmp(pName, s_VerNames[i], Length) == 0 {
			return i
		}
	}
	return 0
}

// check if it is a known module
func Prs_ManFindType(pName string, pInputs *int32, fOut int32, ppNames **string) int32 {
	var i, Length int32
	*pInputs = -1
	for i = int32(1); i < int32(len(s_VerInfo)) && s_VerInfo[i].pTypeName != ""; i++ {
		Length = int32(len(s_VerInfo[i].pTypeName))
		if strncmp(pName, s_VerInfo[i].pTypeName, Length) == 0 {
			*pInputs = s_VerInfo[i].nInputs
			offset := int32(0)
			if fOut != 0 {
				offset = s_VerInfo[i].nInputs
			}
			*ppNames = &s_VerInfo[i].pSigNames[offset]
			return s_VerInfo[i].Type
		}
	}
	return int32(CBA_OBJ_BOX)
}

/**Function*************************************************************

  Synopsis    []

  Description []

  SideEffects []

  SeeAlso     []

***********************************************************************/

// skips Verilog comments (returns 1 if some comments were skipped)
func Prs_ManUtilSkipComments(p *Prs_Man_t) int32 {
	if !(Prs_ManIsChar(p, '/') != 0) {
		return 0
	}
	if Prs_ManIsChar1(p, '/') != 0 {
		for p.pCur += 2; p.pCur < p.pLimit; p.pCur++ {
			if Prs_ManIsChar(p, '\n') != 0 {
				p.pCur++
				return 1
			}
		}

	} else if Prs_ManIsChar1(p, '*') != 0 {
		for p.pCur += 2; p.pCur < p.pLimit; p.pCur++ {
			if (Prs_ManIsChar(p, '*') != 0) && (Prs_ManIsChar1(p, '/') != 0) {
				p.pCur++
				p.pCur++
				return 1
			}
		}
	}
	return 0
}
func Prs_ManUtilSkipName(p *Prs_Man_t) int32 {
	if !(Prs_ManIsChar(p, '\\') != 0) {
		return 0
	}
	for p.pCur++; p.pCur < p.pLimit; p.pCur++ {
		if Prs_ManIsChar(p, ' ') != 0 {
			p.pCur++
			return 1
		}
	}
	return 0
}

// skip any number of spaces and comments
func Prs_ManUtilSkipSpaces(p *Prs_Man_t) int32 {
	for p.pCur < p.pLimit {
		for Prs_CharIsSpace(Prs_ManPeek(p, 0)) != 0 {
			p.pCur++
		}
		if Prs_ManPeek(p, 0) == 0 {
			return Prs_ManErrorSet(p, "Unexpectedly reached end-of-file.", 1)
		}
		if !(Prs_ManUtilSkipComments(p) != 0) {
			return 0
		}
	}
	return Prs_ManErrorSet(p, "Unexpectedly reached end-of-file.", 1)
}

// skip everything including comments until the given char
func Prs_ManUtilSkipUntil(p *Prs_Man_t, c byte) int32 {
	for p.pCur < p.pLimit {
		if Prs_ManIsChar(p, c) != 0 {
			return 1
		}
		if Prs_ManUtilSkipComments(p) != 0 {
			continue
		}
		if Prs_ManUtilSkipName(p) != 0 {
			continue
		}
		p.pCur++
	}
	return 0
}

// skip everything including comments until the given word
func Prs_ManUtilSkipUntilWord(p *Prs_Man_t, pWord string) int32 {
	if p.pCur > p.pLimit {
		return 1
	}
	rel := strings.Index(p.pBuffer[p.pCur:p.pLimit], pWord)
	if rel < 0 {
		return 1
	}
	p.pCur += rel + len(pWord)
	return 0
}

// detect two symbols on the same line
func Prs_ManUtilDetectTwo(p *Prs_Man_t, Sym1 byte, Sym2 byte) int32 {
	for i := p.pCur; i < p.pLimit && Prs_ManAt(p, i) != ';'; i++ {
		if Prs_ManAt(p, i) == Sym1 && Prs_ManAt(p, i) == Sym2 {
			return 1
		}
	}
	return 0
}

// find closing paren
func Prs_ManFindClosingParenthesis(p *Prs_Man_t, Open byte, Close byte) int {
	var Counter int32
	var fNotName int32 = 1
	assert((Prs_ManIsChar(p, Open) != 0), "Prs_ManIsChar(p, Open)")
	for i := p.pCur; Prs_ManAt(p, i) != 0; i++ {
		c := Prs_ManAt(p, i)
		if fNotName != 0 {
			if c == Open {
				Counter++
			}
			if c == Close {
				Counter--
			}
			if Counter == 0 {
				return i
			}
		}
		if c == '\\' {
			fNotName = 0
		} else if !(fNotName != 0) && c == ' ' {
			fNotName = 1
		}
	}
	return -1
}

/*
*Function*************************************************************

	Synopsis    []

	Description []

	SideEffects []

	SeeAlso     []

**********************************************************************
*/
func Prs_ManReadName(p *Prs_Man_t) int32 {
	pStart := p.pCur
	if Prs_ManIsChar(p, '\\') != 0 { // escaped name

		p.pCur++
		pStart = p.pCur
		for p.pCur < p.pLimit && !(Prs_ManIsChar(p, ' ') != 0) && Prs_ManPeek(p, 0) != 0 {
			p.pCur++
		}

	} else if Prs_CharIsSymb1(Prs_ManPeek(p, 0)) != 0 { // simple name

		p.pCur++
		for Prs_CharIsSymb2(Prs_ManPeek(p, 0)) != 0 {
			p.pCur++
		}

	} else {
		return 0
	}
	s := Prs_ManSlice(p, pStart, p.pCur)
	return Abc_NamStrFindOrAddLim(p.pStrs, s, len(s), nil)
}
func Prs_ManReadNameList(p *Prs_Man_t, vTemp *Vec_Int_t, LastSymb byte) int32 {
	Vec_IntClear(vTemp)
	for {
		var Item int32 = Prs_ManReadName(p)
		if Item == 0 {
			return Prs_ManErrorSet(p, "Cannot read name in the list.", 0)
		}
		if Prs_ManUtilSkipSpaces(p) != 0 {
			return Prs_ManErrorSet(p, "Error number 1.", 0)
		}
		if Item == int32(PRS_VER_WIRE) {
			continue
		}
		Vec_IntPush(vTemp, Item)
		if Prs_ManIsChar(p, LastSymb) != 0 {
			break
		}
		if !(Prs_ManIsChar(p, ',') != 0) {
			return Prs_ManErrorSet(p, "Expecting comma in the list.", 0)
		}
		p.pCur++
		if Prs_ManUtilSkipSpaces(p) != 0 {
			return Prs_ManErrorSet(p, "Error number 2.", 0)
		}
	}
	return 1
}
func Prs_ManReadConstant(p *Prs_Man_t) int32 {
	pStart := p.pCur
	assert((Prs_ManIsDigit(p) != 0), "Prs_ManIsDigit(p)")
	for Prs_ManIsDigit(p) != 0 {
		p.pCur++
	}
	if !(Prs_ManIsChar(p, '\'') != 0) {
		s := Prs_ManSlice(p, pStart, p.pCur)
		return Abc_NamStrFindOrAddLim(p.pFuns, s, len(s), nil)
	}
	p.pCur++
	if Prs_ManIsChar(p, 's') != 0 {
		p.pCur++
	}
	if Prs_ManIsChar(p, 'b') != 0 {
		p.pCur++
		for Prs_CharIsDigitB(Prs_ManPeek(p, 0)) != 0 {
			if Prs_ManPeek(p, 0) == '0' {
				p.pNtk.fHasC0s = 1
			} else if Prs_ManPeek(p, 0) == '1' {
				p.pNtk.fHasC1s = 1
			} else if Prs_ManPeek(p, 0) == 'x' {
				p.pNtk.fHasCXs = 1
			} else if Prs_ManPeek(p, 0) == 'z' {
				p.pNtk.fHasCZs = 1
			}
			p.pCur++
		}

	} else if Prs_ManIsChar(p, 'h') != 0 {
		p.pCur++
		p.pNtk.fHasC0s = 1
		for Prs_CharIsDigitH(Prs_ManPeek(p, 0)) != 0 {
			if Prs_ManPeek(p, 0) != '0' {
				p.pNtk.fHasC1s = 1
			}
			p.pCur++
		}

	} else if Prs_ManIsChar(p, 'd') != 0 {
		p.pCur++
		p.pNtk.fHasC0s = 1
		for Prs_ManIsDigit(p) != 0 {
			if Prs_ManPeek(p, 0) != '0' {
				p.pNtk.fHasC1s = 1
			}
			p.pCur++
		}

	} else {
		return Prs_ManErrorSet(p, "Cannot read radix of constant.", 0)
	}
	s := Prs_ManSlice(p, pStart, p.pCur)
	return Abc_NamStrFindOrAddLim(p.pFuns, s, len(s), nil)
}
func Prs_ManReadRange(p *Prs_Man_t) int32 {
	var Left, Right int32
	assert((Prs_ManIsChar(p, '[') != 0), "Prs_ManIsChar(p, '[')")
	p.pCur++
	if Prs_ManUtilSkipSpaces(p) != 0 {
		return Prs_ManErrorSet(p, "Error number 3.", 0)
	}
	if !(Prs_ManIsDigit(p) != 0) {
		return Prs_ManErrorSet(p, "Cannot read digit in range specification.", 0)
	}
	start := p.pCur
	for Prs_ManIsDigit(p) != 0 {
		p.pCur++
	}
	n, _ := strconv.Atoi(Prs_ManSlice(p, start, p.pCur))
	Left = int32(n)
	Right = Left
	if Prs_ManUtilSkipSpaces(p) != 0 {
		return Prs_ManErrorSet(p, "Error number 4.", 0)
	}
	if Prs_ManIsChar(p, ':') != 0 {
		p.pCur++
		if Prs_ManUtilSkipSpaces(p) != 0 {
			return Prs_ManErrorSet(p, "Error number 5.", 0)
		}
		if !(Prs_ManIsDigit(p) != 0) {
			return Prs_ManErrorSet(p, "Cannot read digit in range specification.", 0)
		}
		start = p.pCur
		for Prs_ManIsDigit(p) != 0 {
			p.pCur++
		}
		n, _ = strconv.Atoi(Prs_ManSlice(p, start, p.pCur))
		Right = int32(n)
		if Prs_ManUtilSkipSpaces(p) != 0 {
			return Prs_ManErrorSet(p, "Error number 6.", 0)
		}
	}
	if !(Prs_ManIsChar(p, ']') != 0) {
		return Prs_ManErrorSet(p, "Cannot read closing brace in range specification.", 0)
	}
	p.pCur++
	if Prs_ManUtilSkipSpaces(p) != 0 {
		return Prs_ManErrorSet(p, "Error number 6a.", 0)
	}
	return Hash_Int2ManInsert(p.vHash, Left, Right, 0)
}
func Prs_ManReadConcat(p *Prs_Man_t, vTemp2 *Vec_Int_t) int32 {
	assert((Prs_ManIsChar(p, '{') != 0), "Prs_ManIsChar(p, '{')")
	p.pCur++
	if !(Prs_ManReadSignalList(p, vTemp2, '}', 0) != 0) {
		return Prs_ManErrorSet(p, "Error number 7.", 0)
	}
	// check final
	assert((Prs_ManIsChar(p, '}') != 0), "Prs_ManIsChar(p, '}')")
	p.pCur++
	// return special case
	assert(Vec_IntSize(vTemp2) > 0, "Vec_IntSize(vTemp2) > 0")
	if Vec_IntSize(vTemp2) == 1 {
		return Vec_IntEntry(vTemp2, 0)
	}
	return Abc_Var2Lit2(Prs_NtkAddConcat(p.pNtk, vTemp2), int32(CBA_PRS_CONCAT))
}
func Prs_ManReadSignal(p *Prs_Man_t) int32 {
	var Item int32
	if Prs_ManUtilSkipSpaces(p) != 0 {
		return Prs_ManErrorSet(p, "Error number 8.", 0)
	}
	if Prs_ManIsDigit(p) != 0 {
		Item = Prs_ManReadConstant(p)
		if Item == 0 {
			return 0
		}
		if Prs_ManUtilSkipSpaces(p) != 0 {
			return Prs_ManErrorSet(p, "Error number 10.", 0)
		}
		return Abc_Var2Lit2(Item, int32(CBA_PRS_CONST))
	}
	if Prs_ManIsChar(p, '{') != 0 {
		if Prs_CharIsDigit(Prs_ManPeek(p, 1)) != 0 {
			p.pCur++
			if Prs_ManIsDigit(p) != 0 {
				var i int32
				start := p.pCur
				for Prs_ManIsDigit(p) != 0 {
					p.pCur++
				}
				Num, _ := strconv.Atoi(Prs_ManSlice(p, start, p.pCur))
				if Prs_ManUtilSkipSpaces(p) != 0 {
					return Prs_ManErrorSet(p, "Error number 10.", 0)
				}
				assert((Prs_ManIsChar(p, '{') != 0), "Prs_ManIsChar(p, '{')")
				p.pCur++
				if Prs_ManUtilSkipSpaces(p) != 0 {
					return Prs_ManErrorSet(p, "Error number 10.", 0)
				}
				Item = Prs_ManReadSignal(p)
				assert((Prs_ManIsChar(p, '}') != 0), "Prs_ManIsChar(p, '}')")
				p.pCur++
				if Prs_ManUtilSkipSpaces(p) != 0 {
					return Prs_ManErrorSet(p, "Error number 10.", 0)
				}
				// add to concat all, expect the last one
				assert(p.fUsingTemp2 != 0, "p.fUsingTemp2")
				for i = 0; i < int32(Num-1); i++ {
					Vec_IntPush(&p.vTemp2, Item)
				}
				if Prs_ManUtilSkipSpaces(p) != 0 {
					return Prs_ManErrorSet(p, "Error number 10.", 0)
				}
				assert((Prs_ManIsChar(p, '}') != 0), "Prs_ManIsChar(p, '}')")
				p.pCur++
				if Prs_ManUtilSkipSpaces(p) != 0 {
					return Prs_ManErrorSet(p, "Error number 10.", 0)
				}
				return Item
			}
		}
		if p.fUsingTemp2 != 0 {
			return Prs_ManErrorSet(p, "Cannot read nested concatenations.", 0)
		}
		p.fUsingTemp2 = 1
		Item = Prs_ManReadConcat(p, &p.vTemp2)
		p.fUsingTemp2 = 0
		if Item == 0 {
			return 0
		}
		if Prs_ManUtilSkipSpaces(p) != 0 {
			return Prs_ManErrorSet(p, "Error number 12.", 0)
		}
		return Item

	} else {
		Item = Prs_ManReadName(p)
		if Item == 0 {
			return 1
		} // no actual name
		if Prs_ManUtilSkipSpaces(p) != 0 {
			return Prs_ManErrorSet(p, "Error number 14.", 0)
		}
		if Prs_ManIsChar(p, '[') != 0 {
			var Range int32 = Prs_ManReadRange(p)
			if Range == 0 {
				return Prs_ManErrorSet(p, "Error number 15.", 0)
			}
			if Prs_ManUtilSkipSpaces(p) != 0 {
				return Prs_ManErrorSet(p, "Error number 16.", 0)
			}
			return Abc_Var2Lit2(Prs_NtkAddSlice(p.pNtk, Item, Range), int32(CBA_PRS_SLICE))
		}
		return Abc_Var2Lit2(Item, int32(CBA_PRS_NAME))
	}
}
func Prs_ManReadSignalList(p *Prs_Man_t, vTemp *Vec_Int_t, LastSymb byte, fAddForm int32) int32 {
	Vec_IntClear(vTemp)
	for {
		var Item int32 = Prs_ManReadSignal(p)
		if Item == 0 {
			return Prs_ManErrorSet(p, "Cannot read signal in the list.", 0)
		}
		if fAddForm != 0 {
			Vec_IntPush(vTemp, 0)
		}
		Vec_IntPush(vTemp, Item)
		if Prs_ManIsChar(p, LastSymb) != 0 {
			break
		}
		if !(Prs_ManIsChar(p, ',') != 0) {
			return Prs_ManErrorSet(p, "Expecting comma in the list.", 0)
		}
		p.pCur++
	}
	return 1
}
func Prs_ManReadSignalList2(p *Prs_Man_t, vTemp *Vec_Int_t) int32 {
	var FormId, ActItem int32
	Vec_IntClear(vTemp)
	assert((Prs_ManIsChar(p, '.') != 0), "Prs_ManIsChar(p, '.')")
	for Prs_ManIsChar(p, '.') != 0 {
		p.pCur++
		FormId = Prs_ManReadName(p)
		if FormId == 0 {
			return Prs_ManErrorSet(p, "Cannot read formal name of the instance.", 0)
		}
		if Prs_ManUtilSkipSpaces(p) != 0 {
			return Prs_ManErrorSet(p, "Error number 17.", 0)
		}
		if !(Prs_ManIsChar(p, '(') != 0) {
			return Prs_ManErrorSet(p, "Cannot read \"(\" in the instance.", 0)
		}
		p.pCur++
		if Prs_ManUtilSkipSpaces(p) != 0 {
			return Prs_ManErrorSet(p, "Error number 17.", 0)
		}
		ActItem = Prs_ManReadSignal(p)
		if ActItem == 0 {
			return Prs_ManErrorSet(p, "Cannot read actual name of an instance.", 0)
		}
		if !(Prs_ManIsChar(p, ')') != 0) {
			return Prs_ManErrorSet(p, "Cannot read \")\" in the instance.", 0)
		}
		p.pCur++
		if ActItem != 1 {
			Vec_IntPushTwo(vTemp, FormId, ActItem)
		}
		if Prs_ManUtilSkipSpaces(p) != 0 {
			return Prs_ManErrorSet(p, "Error number 18.", 0)
		}
		if Prs_ManIsChar(p, ')') != 0 {
			break
		}
		if !(Prs_ManIsChar(p, ',') != 0) {
			return Prs_ManErrorSet(p, "Expecting comma in the instance.", 0)
		}
		p.pCur++
		if Prs_ManUtilSkipSpaces(p) != 0 {
			return Prs_ManErrorSet(p, "Error number 19.", 0)
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
func Prs_ManReadFunction(p *Prs_Man_t) int32 {
	// this is a hack to read functions produced by ABC Verilog writer
	p.FuncNameId = 0
	p.FuncRangeId = 0
	if Prs_ManUtilSkipUntilWord(p, "_func_") != 0 {
		return Prs_ManErrorSet(p, "Cannot find \"_func_\" keyword.", 0)
	}
	p.pCur -= 6
	p.FuncNameId = Prs_ManReadName(p)
	if p.FuncNameId == 0 {
		return Prs_ManErrorSet(p, "Error number 30a.", 0)
	}
	if Prs_ManUtilSkipUntilWord(p, "input") != 0 {
		return Prs_ManErrorSet(p, "Cannot find \"input\" keyword.", 0)
	}
	if Prs_ManUtilSkipSpaces(p) != 0 {
		return Prs_ManErrorSet(p, "Error number 30b.", 0)
	}
	if Prs_ManIsChar(p, '[') != 0 {
		p.FuncRangeId = Prs_ManReadRange(p)
	} else if Prs_ManReadName(p) == int32(PRS_VER_SIGNED) {
		if Prs_ManUtilSkipSpaces(p) != 0 {
			return Prs_ManErrorSet(p, "Error number 30c.", 0)
		}
		if Prs_ManIsChar(p, '[') != 0 {
			p.FuncRangeId = Prs_ManReadRange(p)
		}
	}
	if Prs_ManUtilSkipUntilWord(p, "endfunction") != 0 {
		return Prs_ManErrorSet(p, "Cannot find \"endfunction\" keyword.", 0)
	}
	return 1
}
func Prs_ManReadAlways(p *Prs_Man_t) int32 {
	// this is a hack to read always-statement representing case-statement
	var iToken int32
	var pClose int
	if Prs_ManUtilSkipSpaces(p) != 0 {
		return Prs_ManErrorSet(p, "Error number 23.", 0)
	}
	if !(Prs_ManIsChar(p, '@') != 0) {
		return Prs_ManErrorSet(p, "Cannot parse always statement.", 0)
	}
	p.pCur++
	if Prs_ManUtilSkipSpaces(p) != 0 {
		return Prs_ManErrorSet(p, "Error number 23a.", 0)
	}
	if !(Prs_ManIsChar(p, '(') != 0) {
		return Prs_ManErrorSet(p, "Cannot parse always statement.", 0)
	}
	pClose = Prs_ManFindClosingParenthesis(p, '(', ')')
	if pClose < 0 {
		return Prs_ManErrorSet(p, "Expecting closing parenthesis 1.", 0)
	}
	p.pCur = pClose
	if !(Prs_ManIsChar(p, ')') != 0) {
		return Prs_ManErrorSet(p, "Cannot parse always statement.", 0)
	}
	p.pCur++
	// read begin
	if Prs_ManUtilSkipSpaces(p) != 0 {
		return Prs_ManErrorSet(p, "Error number 23a.", 0)
	}
	iToken = Prs_ManReadName(p)
	if iToken != int32(PRS_VER_BEGIN) {
		return Prs_ManErrorSet(p, "Cannot read \"begin\" keyword.", 0)
	}
	// read case
	if Prs_ManUtilSkipSpaces(p) != 0 {
		return Prs_ManErrorSet(p, "Error number 23a.", 0)
	}
	iToken = Prs_ManReadName(p)
	if iToken != int32(PRS_VER_CASE) {
		return Prs_ManErrorSet(p, "Cannot read \"case\" keyword.", 0)
	}
	// read control
	if Prs_ManUtilSkipSpaces(p) != 0 {
		return Prs_ManErrorSet(p, "Error number 23a.", 0)
	}
	if !(Prs_ManIsChar(p, '(') != 0) {
		return Prs_ManErrorSet(p, "Cannot parse always statement.", 0)
	}
	p.pCur++
	if Prs_ManUtilSkipSpaces(p) != 0 {
		return Prs_ManErrorSet(p, "Error number 23a.", 0)
	}
	iToken = Prs_ManReadSignal(p)
	if iToken == 0 {
		return Prs_ManErrorSet(p, "Cannot read output in assign-statement.", 0)
	}
	if Prs_ManUtilSkipSpaces(p) != 0 {
		return Prs_ManErrorSet(p, "Error number 23a.", 0)
	}
	if !(Prs_ManIsChar(p, ')') != 0) {
		return Prs_ManErrorSet(p, "Cannot parse always statement.", 0)
	}
	p.pCur++
	// save control
	Vec_IntClear(&p.vTemp3)
	Vec_IntPushTwo(&p.vTemp3, 0, 0) // output will go here
	Vec_IntPushTwo(&p.vTemp3, 0, iToken)
	// read conditions
	if Prs_ManUtilSkipSpaces(p) != 0 {
		return Prs_ManErrorSet(p, "Error number 23a.", 0)
	}
	if !(Prs_ManIsDigit(p) != 0) {
		return Prs_ManErrorSet(p, "Cannot parse always statement.", 0)
	}
	for Prs_ManIsDigit(p) != 0 {
		for Prs_ManIsDigit(p) != 0 {
			p.pCur++
		}
		if Prs_ManUtilSkipSpaces(p) != 0 {
			return Prs_ManErrorSet(p, "Error number 23a.", 0)
		}
		if !(Prs_ManIsChar(p, ':') != 0) {
			return Prs_ManErrorSet(p, "Cannot parse always statement.", 0)
		}
		p.pCur++
		if Prs_ManUtilSkipSpaces(p) != 0 {
			return Prs_ManErrorSet(p, "Error number 23a.", 0)
		}
		// read output
		iToken = Prs_ManReadSignal(p)
		if iToken == 0 {
			return Prs_ManErrorSet(p, "Cannot read output in assign-statement.", 0)
		}
		if Prs_ManUtilSkipSpaces(p) != 0 {
			return Prs_ManErrorSet(p, "Error number 23a.", 0)
		}
		if !(Prs_ManIsChar(p, '=') != 0) {
			return Prs_ManErrorSet(p, "Cannot parse always statement.", 0)
		}
		p.pCur++
		// save output
		Vec_IntWriteEntry(&p.vTemp3, 1, iToken)
		// read input
		iToken = Prs_ManReadSignal(p)
		if iToken == 0 {
			return Prs_ManErrorSet(p, "Cannot read output in assign-statement.", 0)
		}
		if Prs_ManUtilSkipSpaces(p) != 0 {
			return Prs_ManErrorSet(p, "Error number 23a.", 0)
		}
		if !(Prs_ManIsChar(p, ';') != 0) {
			return Prs_ManErrorSet(p, "Cannot parse always statement.", 0)
		}
		p.pCur++
		if Prs_ManUtilSkipSpaces(p) != 0 {
			return Prs_ManErrorSet(p, "Error number 23a.", 0)
		}
		// save input
		Vec_IntPushTwo(&p.vTemp3, 0, iToken)
	}
	// read endcase
	if Prs_ManUtilSkipSpaces(p) != 0 {
		return Prs_ManErrorSet(p, "Error number 23a.", 0)
	}
	iToken = Prs_ManReadName(p)
	if iToken != int32(PRS_VER_ENDCASE) {
		return Prs_ManErrorSet(p, "Cannot read \"endcase\" keyword.", 0)
	}
	// read end
	if Prs_ManUtilSkipSpaces(p) != 0 {
		return Prs_ManErrorSet(p, "Error number 23a.", 0)
	}
	iToken = Prs_ManReadName(p)
	if iToken != int32(PRS_VER_END) {
		return Prs_ManErrorSet(p, "Cannot read \"end\" keyword.", 0)
	}
	// save binary operator
	Prs_NtkAddBox(p.pNtk, int32(CBA_BOX_NMUX), 0, &p.vTemp3)
	return 1
}

/*
static inline int Prs_ManReadExpression( Prs_Man_t * p, int OutItem )

	{
	    int InItem, fCompl = 0, fCompl2 = 0, Oper = 0;
	    // read output name
	    if ( Prs_ManIsChar(p, '~') )
	    {
	        fCompl = 1;
	        p->pCur++;
	    }
	    // write output name
	    Vec_IntClear( &p->vTemp );
	    Vec_IntPush( &p->vTemp, 0 );
	    Vec_IntPush( &p->vTemp, OutItem );
	    // read first name
	    InItem = Prs_ManReadSignal( p );
	    if ( InItem == 0 )                      return Prs_ManErrorSet(p, "Cannot read first input name in the assign-statement.", 0);
	    Vec_IntPush( &p->vTemp, 0 );
	    Vec_IntPush( &p->vTemp, InItem );
	    // check unary operator
	    if ( Prs_ManIsChar(p, ';') )
	    {
	        Oper = fCompl ? CBA_BOX_INV : CBA_BOX_BUF;
	        Prs_NtkAddBox( p->pNtk, Oper, 0, &p->vTemp );
	        return 1;
	    }
	    if ( Prs_ManIsChar(p, '&') )
	        Oper = CBA_BOX_AND;
	    else if ( Prs_ManIsChar(p, '|') )
	        Oper = CBA_BOX_OR;
	    else if ( Prs_ManIsChar(p, '^') )
	        Oper = CBA_BOX_XOR;
	    else if ( Prs_ManIsChar(p, '?') )
	        Oper = CBA_BOX_MUX;
	    else                                    return Prs_ManErrorSet(p, "Unrecognized operator in the assign-statement.", 0);
	    p->pCur++;
	    if ( Prs_ManUtilSkipSpaces(p) )         return Prs_ManErrorSet(p, "Error number 24.", 0);
	    if ( Prs_ManIsChar(p, '~') )
	    {
	        fCompl2 = 1;
	        p->pCur++;
	    }
	    // read second name
	    InItem = Prs_ManReadSignal( p );
	    if ( InItem == 0 )                      return Prs_ManErrorSet(p, "Cannot read second input name in the assign-statement.", 0);
	    Vec_IntPush( &p->vTemp, 0 );
	    Vec_IntPush( &p->vTemp, InItem );
	    // read third argument
	    if ( Oper == CBA_BOX_MUX )
	    {
	        assert( fCompl == 0 );
	        if ( !Prs_ManIsChar(p, ':') )       return Prs_ManErrorSet(p, "Expected colon in the MUX assignment.", 0);
	        p->pCur++;
	        // read third name
	        InItem = Prs_ManReadSignal( p );
	        if ( InItem == 0 )                  return Prs_ManErrorSet(p, "Cannot read third input name in the assign-statement.", 0);
	        Vec_IntPush( &p->vTemp, 0 );
	        Vec_IntPush( &p->vTemp, InItem );
	        if ( !Prs_ManIsChar(p, ';') )       return Prs_ManErrorSet(p, "Expected semicolon at the end of the assign-statement.", 0);
	    }
	    else
	    {
	        // figure out operator
	        if ( Oper == CBA_BOX_AND )
	        {
	            if ( fCompl && !fCompl2 )
	                Oper = CBA_BOX_SHARPL;
	            else if ( !fCompl && fCompl2 )
	                Oper = CBA_BOX_SHARP;
	            else if ( fCompl && fCompl2 )
	                Oper = CBA_BOX_NOR;
	        }
	        else if ( Oper == CBA_BOX_OR )
	        {
	            if ( fCompl && fCompl2 )
	                Oper = CBA_BOX_NAND;
	            else assert( !fCompl && !fCompl2 );
	        }
	        else if ( Oper == CBA_BOX_XOR )
	        {
	            if ( fCompl && !fCompl2 )
	                Oper = CBA_BOX_XNOR;
	            else assert( !fCompl && !fCompl2 );
	        }
	    }
	    // save binary operator
	    Prs_NtkAddBox( p->pNtk, Oper, 0, &p->vTemp );
	    return 1;
	}
*/
func Prs_ManExpressionOperator(p *Prs_Man_t, rotating int32) (int32, int) {
	a, b, c := Prs_ManPeek(p, 0), Prs_ManPeek(p, 1), Prs_ManPeek(p, 2)
	switch {
	case a == '>' && b == '>' && c == '>':
		return int32(CBA_BOX_SHIRA), 3
	case a == '<' && b == '<' && c == '<':
		return int32(CBA_BOX_SHILA), 3
	case a == '>' && b == '>':
		if rotating != 0 {
			return int32(CBA_BOX_ROTR), 2
		}
		return int32(CBA_BOX_SHIR), 2
	case a == '<' && b == '<':
		if rotating != 0 {
			return int32(CBA_BOX_ROTL), 2
		}
		return int32(CBA_BOX_SHIL), 2
	case a == '&' && b == '&':
		return int32(CBA_BOX_LAND), 2
	case a == '|' && b == '|':
		return int32(CBA_BOX_LOR), 2
	case a == '=' && b == '=':
		return int32(CBA_BOX_EQU), 2
	case a == '!' && b == '=':
		return int32(CBA_BOX_NEQU), 2
	case a == '<' && b == '=':
		return CBA_BOX_LETHAN, 2
	case a == '>' && b == '=':
		return CBA_BOX_METHAN, 2
	case a == '*' && b == '*':
		return CBA_BOX_POW, 2
	case a == '&':
		return int32(CBA_BOX_AND), 1
	case a == '|':
		return int32(CBA_BOX_OR), 1
	case a == '^':
		return int32(CBA_BOX_XOR), 1
	case a == '<':
		return int32(CBA_BOX_LTHAN), 1
	case a == '>':
		return int32(CBA_BOX_MTHAN), 1
	case a == '+':
		return int32(CBA_BOX_ADD), 1
	case a == '-':
		return int32(CBA_BOX_SUB), 1
	case a == '*':
		return int32(CBA_BOX_MUL), 1
	case a == '/':
		return int32(CBA_BOX_DIV), 1
	case a == '%':
		return int32(CBA_BOX_MOD), 1
	}
	return CBA_OBJ_NONE, 0
}

func Prs_ManReadExpression(p *Prs_Man_t, OutItem int32) int32 {
	// The C parser masks the outer parentheses in-place.  Use a private
	// mutable view so offsets stay shared while caller input remains intact.
	originalBuffer := p.pBuffer
	localBuffer := []byte(originalBuffer)
	p.pBuffer = string(localBuffer)
	defer func() { p.pBuffer = originalBuffer }()
	var pClose int
	var Item int32
	var Type int32 = CBA_OBJ_NONE
	var fRotating int32

	// write output name
	Vec_IntClear(&p.vTemp)
	Vec_IntPush(&p.vTemp, 0)
	Vec_IntPush(&p.vTemp, OutItem)
	if Prs_ManUtilSkipSpaces(p) != 0 {
		return Prs_ManErrorSet(p, "Error number 24.", 0)
	}
	if Prs_ManIsChar(p, '(') != 0 {
		// THIS IS A HACK TO DETECT rotating shifters:  try to find both << and >> on the same line
		if Prs_ManUtilDetectTwo(p, '>', '>') != 0 && Prs_ManUtilDetectTwo(p, '<', '<') != 0 {
			fRotating = 1
		}
		pClose = Prs_ManFindClosingParenthesis(p, '(', ')')
		if pClose < 0 {
			return Prs_ManErrorSet(p, "Expecting closing parenthesis 1.", 0)
		}
		localBuffer[p.pCur] = ' '
		localBuffer[pClose] = ' '
		p.pBuffer = string(localBuffer)
	}
	if Prs_ManUtilSkipSpaces(p) != 0 {
		return Prs_ManErrorSet(p, "Error number 24.", 0)
	}
	// read constant or concatenation
	if (Prs_ManIsDigit(p) != 0) || (Prs_ManIsChar(p, '{') != 0) {
		Item = Prs_ManReadSignal(p)
		// write constant
		Vec_IntPush(&p.vTemp, 0)
		Vec_IntPush(&p.vTemp, Item)
		Type = CBA_BOX_BUF

	} else if (Prs_ManIsChar(p, '!') != 0) || (Prs_ManIsChar(p, '~') != 0) || (Prs_ManIsChar(p, '@') != 0) ||
		(Prs_ManIsChar(p, '&') != 0) || (Prs_ManIsChar(p, '|') != 0) || (Prs_ManIsChar(p, '^') != 0) || (Prs_ManIsChar(p, '-') != 0) {
		if Prs_ManIsChar(p, '!') != 0 {
			Type = CBA_BOX_LNOT
		} else if Prs_ManIsChar(p, '~') != 0 {
			Type = CBA_BOX_INV
		} else if Prs_ManIsChar(p, '@') != 0 {
			Type = CBA_BOX_SQRT
		} else if Prs_ManIsChar(p, '&') != 0 {
			Type = CBA_BOX_RAND
		} else if Prs_ManIsChar(p, '|') != 0 {
			Type = CBA_BOX_ROR
		} else if Prs_ManIsChar(p, '^') != 0 {
			Type = CBA_BOX_RXOR
		} else if Prs_ManIsChar(p, '-') != 0 {
			Type = CBA_BOX_MIN
		} else {
			assert(false, "0")
		}
		p.pCur++
		if Prs_ManUtilSkipSpaces(p) != 0 {
			return Prs_ManErrorSet(p, "Error number 24.", 0)
		}
		// skip parentheses
		if Prs_ManIsChar(p, '(') != 0 {
			pClose = Prs_ManFindClosingParenthesis(p, '(', ')')
			if pClose < 0 {
				return Prs_ManErrorSet(p, "Expecting closing parenthesis 2.", 0)
			}
			localBuffer[p.pCur] = ' '
			localBuffer[pClose] = ' '
			p.pBuffer = string(localBuffer)
		}
		if Prs_ManUtilSkipSpaces(p) != 0 {
			return Prs_ManErrorSet(p, "Error number 24.", 0)
		}
		// read first name
		Item = Prs_ManReadSignal(p)
		if Item == 0 {
			return Prs_ManErrorSet(p, "Cannot read name after a unary operator.", 0)
		}
		Vec_IntPush(&p.vTemp, 0)
		Vec_IntPush(&p.vTemp, Item)

	} else {
		// read first name
		Item = Prs_ManReadSignal(p)
		if Item == 0 {
			return Prs_ManErrorSet(p, "Cannot read name after a binary operator.", 0)
		}
		// check if this is a recent function
		if Abc_Lit2Var2(Item) == p.FuncNameId {
			var Status, nInputs, RangeSize int32
			if Prs_ManUtilSkipSpaces(p) != 0 {
				return Prs_ManErrorSet(p, "Error number 24.", 0)
			}
			if !(Prs_ManIsChar(p, '(') != 0) {
				return Prs_ManErrorSet(p, "Error number 24.", 0)
			}
			p.pCur++
			Status = Prs_ManReadSignalList(p, &p.vTemp, ')', 1)
			nInputs = Vec_IntSize(&p.vTemp) / 2
			if p.FuncRangeId != 0 {
				RangeSize = Ptr_NtkRangeSize(p.pNtk, p.FuncRangeId)
			} else {
				RangeSize = 1
			}
			p.FuncNameId, p.FuncRangeId = 0, 0
			if Status == 0 {
				return 0
			}
			if nInputs == 1 {
				Type = CBA_BOX_DEC
			} else if nInputs == RangeSize+1 {
				Type = CBA_BOX_SEL
			} else if nInputs == (1<<RangeSize)+1 {
				Type = CBA_BOX_NMUX
			} else {
				return Prs_ManErrorSet(p, "Cannot determine word-level operator.", 0)
			}
			p.pCur++
			if Prs_ManUtilSkipSpaces(p) != 0 {
				return Prs_ManErrorSet(p, "Error number 24.", 0)
			}
			// save word-level operator
			Vec_IntInsert(&p.vTemp, 0, 0)
			Vec_IntInsert(&p.vTemp, 1, OutItem)
			Prs_NtkAddBox(p.pNtk, Type, 0, &p.vTemp)
			return 1
		}
		Vec_IntPush(&p.vTemp, 0)
		Vec_IntPush(&p.vTemp, Item)
		if Prs_ManUtilSkipSpaces(p) != 0 {
			return Prs_ManErrorSet(p, "Error number 24.", 0)
		}
		assert(!(Prs_ManIsChar(p, '[') != 0), "!Prs_ManIsChar(p, '[')")

		// get the next symbol
		if (Prs_ManIsChar(p, ',') != 0) || (Prs_ManIsChar(p, ';') != 0) {
			Type = CBA_BOX_BUF
		} else if Prs_ManIsChar(p, '?') != 0 {
			p.pCur++
			Item = Prs_ManReadSignal(p)
			if Item == 0 {
				return 0
			}
			Vec_IntPush(&p.vTemp, 0)
			Vec_IntPush(&p.vTemp, Item)
			if Prs_ManUtilSkipSpaces(p) != 0 {
				return Prs_ManErrorSet(p, "Error number 24.", 0)
			}
			if !(Prs_ManIsChar(p, ':') != 0) {
				return Prs_ManErrorSet(p, "MUX lacks the colon symbol (:).", 0)
			}

			p.pCur++
			Item = Prs_ManReadSignal(p)
			if Item == 0 {
				return 0
			}
			Vec_IntPush(&p.vTemp, 0)
			Vec_IntPush(&p.vTemp, Item)
			if Prs_ManUtilSkipSpaces(p) != 0 {
				return Prs_ManErrorSet(p, "Error number 24.", 0)
			}
			assert(Vec_IntSize(&p.vTemp) == 8, "Vec_IntSize(&p.vTemp) == 8")
			//ABC_SWAP( int, Vec_IntArray(&p->vTemp)[3], Vec_IntArray(&p->vTemp)[5] );
			Type = CBA_BOX_MUX

		} else {
			if op, n := Prs_ManExpressionOperator(p, fRotating); op != CBA_OBJ_NONE {
				Type = op
				p.pCur += n
			} else {
				return Prs_ManErrorSet(p, "Unsupported operation.", 0)
			}

			Item = Prs_ManReadSignal(p)
			if Item == 0 {
				return 0
			}
			Vec_IntPush(&p.vTemp, 0)
			Vec_IntPush(&p.vTemp, Item)
			// for adder insert carry-in
			if Type == CBA_BOX_ADD {
				Vec_IntInsert(&p.vTemp, 2, 0)
			}
			if Type == CBA_BOX_ADD {
				Vec_IntInsert(&p.vTemp, 3, 0)
			}
		}
	}
	if Prs_ManUtilSkipSpaces(p) != 0 {
		return Prs_ManErrorSet(p, "Error number 24.", 0)
	}
	// make sure there is nothing left there
	if fRotating != 0 {
		Prs_ManUtilSkipUntilWord(p, ";")
		p.pCur--

	} else if !(Prs_ManIsChar(p, ',') != 0) && !(Prs_ManIsChar(p, ';') != 0) {
		return Prs_ManErrorSet(p, "Trailing symbols on this line.", 0)
	}
	// save binary operator
	Prs_NtkAddBox(p.pNtk, Type, 0, &p.vTemp)
	return 1
}
func Prs_ManReadDeclaration(p *Prs_Man_t, Type int32) int32 {
	var i, Item, NameId, RangeId, fSigned int32
	var vNames [4]*Vec_Int_t = [4]*Vec_Int_t{&p.pNtk.vInputs, &p.pNtk.vOutputs, &p.pNtk.vInouts, &p.pNtk.vWires}
	var vNamesR [4]*Vec_Int_t = [4]*Vec_Int_t{&p.pNtk.vInputsR, &p.pNtk.vOutputsR, &p.pNtk.vInoutsR, &p.pNtk.vWiresR}
	assert(Type >= PRS_VER_INPUT && Type <= PRS_VER_WIRE, "Type >= PRS_VER_INPUT && Type <= PRS_VER_WIRE")
	// read first word
	if Prs_ManUtilSkipSpaces(p) != 0 {
		return Prs_ManErrorSet(p, "Error number 20.", 0)
	}
	if Prs_ManIsChar(p, '[') != 0 {
		RangeId = Prs_ManReadRange(p)
		if RangeId == 0 {
			return Prs_ManErrorSet(p, "Error number 21.", 0)
		}
	}
	Item = Prs_ManReadName(p)
	if Item == int32(PRS_VER_SIGNED) {
		fSigned = 1
		if Prs_ManUtilSkipSpaces(p) != 0 {
			return Prs_ManErrorSet(p, "Error number 20.", 0)
		}
		if Prs_ManIsChar(p, '[') != 0 {
			RangeId = Prs_ManReadRange(p)
			if RangeId == 0 {
				return Prs_ManErrorSet(p, "Error number 21.", 0)
			}
		}
		Item = Prs_ManReadName(p)
	}
	if Item == int32(PRS_VER_WIRE) {
		if Prs_ManUtilSkipSpaces(p) != 0 {
			return Prs_ManErrorSet(p, "Error number 20.", 0)
		}
		if Prs_ManIsChar(p, '[') != 0 {
			RangeId = Prs_ManReadRange(p)
			if RangeId == 0 {
				return Prs_ManErrorSet(p, "Error number 21.", 0)
			}
		}
		Item = Prs_ManReadName(p)
	}
	// read variable names
	Vec_IntClear(&p.vTemp3)
	for {
		if Item == 0 {
			return Prs_ManErrorSet(p, "Cannot read name in the list.", 0)
		}
		if Prs_ManUtilSkipSpaces(p) != 0 {
			return Prs_ManErrorSet(p, "Error number 22a", 0)
		}
		if Item == int32(PRS_VER_WIRE) {
			continue
		}
		Vec_IntPush(&p.vTemp3, Item)
		if Prs_ManIsChar(p, '=') != 0 {
			if Type == PRS_VER_INPUT {
				return Prs_ManErrorSet(p, "Input cannot be defined", 0)
			}
			p.pCur++
			if Prs_ManUtilSkipSpaces(p) != 0 {
				return Prs_ManErrorSet(p, "Error number 23.", 0)
			}
			if !(Prs_ManReadExpression(p, Abc_Var2Lit2(Item, int32(CBA_PRS_NAME))) != 0) {
				return 0
			}
		}
		if Prs_ManIsChar(p, ';') != 0 {
			break
		}
		if !(Prs_ManIsChar(p, ',') != 0) {
			return Prs_ManErrorSet(p, "Expecting comma in the list.", 0)
		}
		p.pCur++
		if Prs_ManUtilSkipSpaces(p) != 0 {
			return Prs_ManErrorSet(p, "Error number 22b.", 0)
		}
		Item = Prs_ManReadName(p)
	}
	// inlining of macro Vec_IntForEachEntry
	for i = int32(0); i < Vec_IntSize(&p.vTemp3); i++ {
		NameId = Vec_IntEntry(&p.vTemp3, i)
		Vec_IntPush(vNames[Type-PRS_VER_INPUT], NameId)
		Vec_IntPush(vNamesR[Type-PRS_VER_INPUT], Abc_Var2Lit(RangeId, fSigned))
		if Type < PRS_VER_WIRE {
			Vec_IntPush(&p.pNtk.vOrder, Abc_Var2Lit2(NameId, Type))
		}
	}
	return 1
}
func Prs_ManReadInstance(p *Prs_Man_t, Func int32) int32 {
	var InstId, Status int32
	if Prs_ManUtilSkipSpaces(p) != 0 {
		return Prs_ManErrorSet(p, "Error number 25.", 0)
	}
	if Prs_ManIsChar(p, '#') != 0 {
		p.pCur++
		for Prs_ManIsDigit(p) != 0 {
			p.pCur++
		}
		if Prs_ManUtilSkipSpaces(p) != 0 {
			return Prs_ManErrorSet(p, "Error number 25.", 0)
		}
	}
	InstId = Prs_ManReadName(p)
	if InstId != 0 {
		if Prs_ManUtilSkipSpaces(p) != 0 {
			return Prs_ManErrorSet(p, "Error number 26.", 0)
		}
	}
	if !(Prs_ManIsChar(p, '(') != 0) {
		return Prs_ManErrorSet(p, "Expecting \"(\" in module instantiation.", 0)
	}
	p.pCur++
	if Prs_ManUtilSkipSpaces(p) != 0 {
		return Prs_ManErrorSet(p, "Error number 27.", 0)
	}
	if Prs_ManIsChar(p, '.') != 0 { // box
		Status = Prs_ManReadSignalList2(p, &p.vTemp)
	} else { // node

		//char * s = Abc_NamStr(p->pStrs, Func);
		// translate elementary gate
		var iFuncNew int32 = Prs_ManIsVerilogPrim(Abc_NamStr(p.pStrs, Func))
		if iFuncNew == 0 {
			return Prs_ManErrorSet(p, "Cannot find elementary gate.", 0)
		}
		Func = iFuncNew
		Status = Prs_ManReadSignalList(p, &p.vTemp, ')', 1)
	}
	if Status == 0 {
		return Prs_ManErrorSet(p, "Error number 28.", 0)
	}
	assert((Prs_ManIsChar(p, ')') != 0), "Prs_ManIsChar(p, ')')")
	p.pCur++
	if Prs_ManUtilSkipSpaces(p) != 0 {
		return Prs_ManErrorSet(p, "Error number 29.", 0)
	}
	if !(Prs_ManIsChar(p, ';') != 0) {
		return Prs_ManErrorSet(p, "Expecting semicolon in the instance.", 0)
	}
	// add box
	Prs_NtkAddBox(p.pNtk, Func, InstId, &p.vTemp)
	return 1
}
func Prs_ManReadArguments(p *Prs_Man_t) int32 {
	var iRange int32
	var iType int32 = -1
	var vSigs [3]*Vec_Int_t = [3]*Vec_Int_t{&p.pNtk.vInputs, &p.pNtk.vOutputs, &p.pNtk.vInouts}
	var vSigsR [3]*Vec_Int_t = [3]*Vec_Int_t{&p.pNtk.vInputsR, &p.pNtk.vOutputsR, &p.pNtk.vInoutsR}
	assert((Prs_ManIsChar(p, '(') != 0), "Prs_ManIsChar(p, '(')")
	p.pCur++
	if Prs_ManUtilSkipSpaces(p) != 0 {
		return Prs_ManErrorSet(p, "Error number 30.", 0)
	}
	if Prs_ManIsChar(p, ')') != 0 {
		return 1
	}
	for {
		var fEscape int32 = Prs_ManIsChar(p, '\\')
		var iName int32 = Prs_ManReadName(p)
		var fSigned int32
		if iName == 0 {
			return Prs_ManErrorSet(p, "Error number 31.", 0)
		}
		if Prs_ManUtilSkipSpaces(p) != 0 {
			return Prs_ManErrorSet(p, "Error number 32.", 0)
		}
		if iName >= PRS_VER_INPUT && iName <= PRS_VER_INOUT && !(fEscape != 0) { // declaration

			iType = iName
			if Prs_ManIsChar(p, '[') != 0 {
				iRange = Prs_ManReadRange(p)
				if iRange == 0 {
					return Prs_ManErrorSet(p, "Error number 33.", 0)
				}
				if Prs_ManUtilSkipSpaces(p) != 0 {
					return Prs_ManErrorSet(p, "Error number 34.", 0)
				}
			}
			iName = Prs_ManReadName(p)
			if iName == 0 {
				return Prs_ManErrorSet(p, "Error number 35.", 0)
			}
			if iName == int32(PRS_VER_SIGNED) {
				fSigned = 1
				if Prs_ManUtilSkipSpaces(p) != 0 {
					return Prs_ManErrorSet(p, "Error number 32.", 0)
				}
				if Prs_ManIsChar(p, '[') != 0 {
					iRange = Prs_ManReadRange(p)
					if iRange == 0 {
						return Prs_ManErrorSet(p, "Error number 33.", 0)
					}
					if Prs_ManUtilSkipSpaces(p) != 0 {
						return Prs_ManErrorSet(p, "Error number 34.", 0)
					}
				}
				iName = Prs_ManReadName(p)
				if iName == 0 {
					return Prs_ManErrorSet(p, "Error number 35.", 0)
				}
			}
		}
		if iType > 0 {
			Vec_IntPush(vSigs[iType-PRS_VER_INPUT], iName)
			Vec_IntPush(vSigsR[iType-PRS_VER_INPUT], Abc_Var2Lit(iRange, fSigned))
			Vec_IntPush(&p.pNtk.vOrder, Abc_Var2Lit2(iName, iType))
		}
		if Prs_ManUtilSkipSpaces(p) != 0 {
			return Prs_ManErrorSet(p, "Error number 36.", 0)
		}
		if Prs_ManIsChar(p, ')') != 0 {
			break
		}
		if !(Prs_ManIsChar(p, ',') != 0) {
			return Prs_ManErrorSet(p, "Expecting comma in the instance.", 0)
		}
		p.pCur++
		if Prs_ManUtilSkipSpaces(p) != 0 {
			return Prs_ManErrorSet(p, "Error number 36.", 0)
		}
	}
	// check final
	assert((Prs_ManIsChar(p, ')') != 0), "Prs_ManIsChar(p, ')')")
	return 1
}

// this procedure can return:
// 0 = reached end-of-file; 1 = successfully parsed; 2 = recognized as primitive; 3 = failed and skipped; 4 = error (failed and could not skip)
func Prs_ManReadModule(p *Prs_Man_t) int32 {
	var iToken int32
	var Status int32 = -1
	var fAlways int32
	if p.pNtk != nil {
		return Prs_ManErrorSet(p, "Parsing previous module is unfinished.", 4)
	}
	if Prs_ManUtilSkipSpaces(p) != 0 {
		Prs_ManErrorClear(p)
		return 0
	}
	// read keyword
	for Prs_ManIsChar(p, '`') != 0 {
		Prs_ManUtilSkipUntilWord(p, "\n")
		if Prs_ManUtilSkipSpaces(p) != 0 {
			Prs_ManErrorClear(p)
			return 0
		}
	}
	iToken = Prs_ManReadName(p)
	if iToken != PRS_VER_MODULE {
		return Prs_ManErrorSet(p, "Cannot read \"module\" keyword.", 4)
	}
	if Prs_ManUtilSkipSpaces(p) != 0 {
		return 4
	}
	// read module name
	iToken = Prs_ManReadName(p)
	if iToken == 0 {
		return Prs_ManErrorSet(p, "Cannot read module name.", 4)
	}
	if Prs_ManIsKnownModule(Abc_NamStr(p.pStrs, iToken)) != 0 {
		if Prs_ManUtilSkipUntilWord(p, "endmodule") != 0 {
			return Prs_ManErrorSet(p, "Cannot find \"endmodule\" keyword.", 4)
		}
		//printf( "Warning! Skipped known module \"%s\".\n", Abc_NamStr(p->pStrs, iToken) );
		Vec_IntPush(&p.vKnown, iToken)
		return 2
	}
	Prs_ManInitializeNtk(p, iToken, 1)
	// skip arguments
	if Prs_ManUtilSkipSpaces(p) != 0 {
		return 4
	}
	if !(Prs_ManIsChar(p, '(') != 0) {
		return Prs_ManErrorSet(p, "Cannot find \"(\" in the argument declaration.", 4)
	}
	if !(Prs_ManReadArguments(p) != 0) {
		return 4
	}
	assert(Prs_ManPeek(p, 0) == ')', "*p.pCur == ')'")
	p.pCur++
	if Prs_ManUtilSkipSpaces(p) != 0 {
		return 4
	}
	// read declarations and instances
	for (Prs_ManIsChar(p, ';') != 0) || (fAlways != 0) {
		if !(fAlways != 0) {
			p.pCur++
		}
		fAlways = 0
		if Prs_ManUtilSkipSpaces(p) != 0 {
			return 4
		}
		iToken = Prs_ManReadName(p)
		if iToken == PRS_VER_ENDMODULE {
			Vec_IntPush(&p.vSucceeded, p.pNtk.iModuleName)
			Prs_ManFinalizeNtk(p)
			return 1
		}
		if iToken >= PRS_VER_INPUT && iToken <= PRS_VER_REG { // declaration
			declType := iToken
			if iToken == PRS_VER_REG {
				declType = PRS_VER_WIRE
			}
			Status = Prs_ManReadDeclaration(p, declType)
		} else if iToken == PRS_VER_REG || iToken == PRS_VER_DEFPARAM { // unsupported keywords
			Status = Prs_ManUtilSkipUntil(p, ';')
		} else { // read instance

			if iToken == PRS_VER_ASSIGN {
				// read output name
				var OutItem int32 = Prs_ManReadSignal(p)
				if OutItem == 0 {
					return Prs_ManErrorSet(p, "Cannot read output in assign-statement.", 0)
				}
				if !(Prs_ManIsChar(p, '=') != 0) {
					return Prs_ManErrorSet(p, "Expecting \"=\" in assign-statement.", 0)
				}
				p.pCur++
				if Prs_ManUtilSkipSpaces(p) != 0 {
					return Prs_ManErrorSet(p, "Error number 23.", 0)
				}
				// read expression
				for {
					if !(Prs_ManReadExpression(p, OutItem) != 0) {
						return 0
					}
					if Prs_ManIsChar(p, ';') != 0 {
						break
					}
					assert((Prs_ManIsChar(p, ',') != 0), "Prs_ManIsChar(p, ',')")
					p.pCur++
					if Prs_ManUtilSkipSpaces(p) != 0 {
						return Prs_ManErrorSet(p, "Error number 23a.", 0)
					}
					// read output name
					OutItem = Prs_ManReadSignal(p)
					if OutItem == 0 {
						return Prs_ManErrorSet(p, "Cannot read output in assign-statement.", 0)
					}
					if !(Prs_ManIsChar(p, '=') != 0) {
						return Prs_ManErrorSet(p, "Expecting \"=\" in assign-statement.", 0)
					}
					p.pCur++
					if Prs_ManUtilSkipSpaces(p) != 0 {
						return Prs_ManErrorSet(p, "Error number 23.", 0)
					}
				}

			} else if iToken == PRS_VER_ALWAYS {
				Status = Prs_ManReadAlways(p)
				fAlways = 1
			} else if iToken == PRS_VER_FUNCTION {
				Status = Prs_ManReadFunction(p)
				fAlways = 1
			} else {
				Status = Prs_ManReadInstance(p, iToken)
			}
			if Status == 0 {
				return 4

				if Prs_ManUtilSkipUntilWord(p, "endmodule") != 0 {
					return Prs_ManErrorSet(p, "Cannot find \"endmodule\" keyword.", 4)
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
				Prs_ManFinalizeNtk(p)
				Prs_ManErrorClear(p)
				return 3
			}
		}
		if !(Status != 0) {
			return 4
		}
		if Prs_ManUtilSkipSpaces(p) != 0 {
			return 4
		}
	}
	return Prs_ManErrorSet(p, "Cannot find \";\" in the module definition.", 4)
}
func Prs_ManReadDesign(p *Prs_Man_t) int32 {
	for {
		var RetValue int32 = Prs_ManReadModule(p)
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
func Prs_ManPrintModules(p *Prs_Man_t) {
	var pName string
	var i int32
	fmt.Printf("Succeeded parsing %d models:\n", Vec_IntSize(&p.vSucceeded))
	// inlining of macro Prs_ManForEachNameVec
	for i = int32(0); i < Vec_IntSize(&p.vSucceeded); i++ {
		pName = Abc_NamStr(p.pStrs, Vec_IntEntry(&p.vSucceeded, i))
		fmt.Printf(" %s", pName)
	}
	fmt.Printf("\n")
	fmt.Printf("Skipped %d known models:\n", Vec_IntSize(&p.vKnown))
	// inlining of macro Prs_ManForEachNameVec
	for i = int32(0); i < Vec_IntSize(&p.vKnown); i++ {
		pName = Abc_NamStr(p.pStrs, Vec_IntEntry(&p.vKnown, i))
		fmt.Printf(" %s", pName)
	}
	fmt.Printf("\n")
	fmt.Printf("Skipped %d failed models:\n", Vec_IntSize(&p.vFailed))
	// inlining of macro Prs_ManForEachNameVec
	for i = int32(0); i < Vec_IntSize(&p.vFailed); i++ {
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
func Prs_ManReadVerilog(pFileName string) *Vec_Ptr_t {
	var vPrs *Vec_Ptr_t = nil
	var p *Prs_Man_t = Prs_ManAlloc(pFileName)
	if p == nil {
		return nil
	}
	Abc_NamStrFindOrAdd(p.pFuns, "1'b0", nil)
	Abc_NamStrFindOrAdd(p.pFuns, "1'b1", nil)
	Abc_NamStrFindOrAdd(p.pFuns, "1'bx", nil)
	Abc_NamStrFindOrAdd(p.pFuns, "1'bz", nil)
	Prs_NtkAddVerilogDirectives(p)
	Prs_ManReadDesign(p)
	Prs_ManPrintModules(p)
	if Prs_ManErrorPrint(p) != 0 {
		// inlining of macro ABC_SWAP
		var t *Vec_Ptr_t = vPrs
		vPrs = p.vNtks
		p.vNtks = t
	}
	Prs_ManFree(p)
	return vPrs
}

func Prs_ManReadVerilogTest(pFileName string) {
	var clk abctime = Abc_Clock()
	var vPrs *Vec_Ptr_t = Prs_ManReadVerilog(pFileName)
	if !(vPrs != nil) {
		return
	}
	fmt.Printf("Finished reading %d networks. ", Vec_PtrSize(vPrs))
	fmt.Printf("NameIDs = %d. ", Abc_NamObjNumMax(Prs_ManNameMan(vPrs)))
	fmt.Printf("Memory = %.2f MB. ", float64(Prs_ManMemory(vPrs))/float64(1<<20))
	Abc_PrintTime(1, "Time", Abc_Clock()-clk)
	Prs_ManWriteVerilog(Extra_FileNameGenericAppend(pFileName, "_out.v"), vPrs)
	//    Abc_NamPrint( Prs_ManNameMan(vPrs) );
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
func Prs_CreateVerilogFindFon(p *Cba_Ntk_t, NameId int32) int32 {
	var iFon int32 = Cba_NtkGetMap(p, NameId)
	if iFon != 0 {
		return iFon
	}
	fmt.Printf("Network \"%s\": Signal \"%s\" is not driven.\n", Cba_NtkName(p), Cba_NtkStr(p, NameId))
	return 0
}
func Prs_CreateSlice(p *Cba_Ntk_t, iFon int32, pNtk *Prs_Ntk_t, Range int32) int32 {
	var iObj, iFonNew, NameId int32
	assert(Cba_FonIsReal(iFon) != 0, "Cba_FonIsReal(iFon)")
	// check existing slice
	NameId = Cba_NtkNewStrId(p, "%s", Cba_ManGetSliceName(p, iFon, Range))
	iFonNew = Cba_NtkGetMap(p, NameId)
	if iFonNew != 0 {
		return iFonNew
	}
	// create slice
	iObj = Cba_ObjAlloc(p, CBA_BOX_SLICE, 1, 1)
	Cba_ObjSetName(p, iObj, NameId)
	Cba_ObjSetFinFon(p, iObj, 0, iFon)
	iFonNew = Cba_ObjFon0(p, iObj)
	Cba_FonSetRange(p, iFonNew, Range)
	Cba_FonSetName(p, iFonNew, NameId)
	Cba_NtkSetMap(p, NameId, iFonNew)
	return iFonNew
}
func Prs_CreateCatIn(p *Cba_Ntk_t, pNtk *Prs_Ntk_t, Con int32) int32 {
	var i, Sig, iObj, iFon, NameId, nBits int32
	var vSigs *Vec_Int_t = Prs_CatSignals(pNtk, Con)
	// create input concatenation
	iObj = Cba_ObjAlloc(p, CBA_BOX_CONCAT, Vec_IntSize(vSigs), 1)
	iFon = Cba_ObjFon0(p, iObj)
	//sprintf( Buffer, "_icc%d_", iObj );
	//NameId = Cba_NtkNewStrId( p, Buffer );
	NameId = Cba_NtkNewStrId(p, "_icc%d_", iObj)
	Cba_FonSetName(p, iFon, NameId)
	Cba_NtkSetMap(p, NameId, iFon)
	// set inputs
	// inlining of macro Vec_IntForEachEntry
	for i = int32(0); i < Vec_IntSize(vSigs); i++ {
		Sig = Vec_IntEntry(vSigs, i)
		iFon = Prs_CreateSignalIn(p, pNtk, Sig)
		if iFon != 0 {
			Cba_ObjSetFinFon(p, iObj, i, iFon)
		}
		if iFon != 0 {
			nBits += Cba_FonRangeSize(p, iFon)
		}
	}
	iFon = Cba_ObjFon0(p, iObj)
	Cba_FonSetRange(p, iFon, Cba_NtkHashRange(p, nBits-1, 0))
	return Cba_ObjFon0(p, iObj)
}
func Prs_CreateSignalIn(p *Cba_Ntk_t, pNtk *Prs_Ntk_t, Sig int32) int32 {
	var iFon int32
	var Value int32 = Abc_Lit2Var2(Sig)
	var Type Prs_ManType_t = Prs_ManType_t(Abc_Lit2Att2(Sig))
	if !(Sig != 0) {
		return 0
	}
	if Type == CBA_PRS_NAME {
		return Prs_CreateVerilogFindFon(p, Cba_NtkNewStrId(p, "%s", Prs_NtkStr(pNtk, Value)))
	}
	if Type == CBA_PRS_CONST {
		return Cba_FonFromConst(Value)
	}
	if Type == CBA_PRS_SLICE {
		iFon = Prs_CreateVerilogFindFon(p, Cba_NtkNewStrId(p, "%s", Prs_NtkStr(pNtk, Prs_SliceName(pNtk, Value))))
		if !(iFon != 0) {
			return 0
		}
		return Prs_CreateSlice(p, iFon, pNtk, Prs_SliceRange(pNtk, Value))
	}
	assert(Type == CBA_PRS_CONCAT, "Type == CBA_PRS_CONCAT")
	return Prs_CreateCatIn(p, pNtk, Value)
}
func Prs_CreateRange(p *Cba_Ntk_t, iFon int32, NameId int32) int32 {
	var RangeId int32 = -Cba_NtkGetMap(p, NameId)
	if RangeId < 0 { // this variable is already created
		return Cba_FonRangeSize(p, -RangeId)
	}
	Cba_NtkUnsetMap(p, NameId)
	Cba_NtkSetMap(p, NameId, iFon)
	if RangeId == 0 {
		return 1
	}
	assert(RangeId > 0, "RangeId > 0")
	Cba_FonSetRangeSign(p, iFon, RangeId)
	return Cba_FonRangeSize(p, iFon)
}
func Prs_CreateSignalOut(p *Cba_Ntk_t, iFon int32, pNtk *Prs_Ntk_t, Sig int32) {
	var i, iFonNew, NameOut, RangeOut, NameId, RangeId, RangeSize, nBits int32
	var SigType Prs_ManType_t = Prs_ManType_t(Abc_Lit2Att2(Sig))
	var SigValue int32 = Abc_Lit2Var2(Sig)
	if !(Sig != 0) {
		return
	}
	if SigType == CBA_PRS_NAME {
		NameId = SigValue
		if strncmp(Cba_NtkStr(p, NameId), "Open_", 5) == 0 {
			return
		}
		Cba_FonSetName(p, iFon, NameId)
		Prs_CreateRange(p, iFon, NameId)
		return
	}
	// create name for this fan
	NameOut = Cba_NtkNewStrId(p, "_occ%d_", iFon)
	Cba_FonSetName(p, iFon, NameOut)
	Cba_NtkSetMap(p, NameOut, iFon)
	// consider special cases
	if SigType == CBA_PRS_SLICE {
		NameId = Prs_SliceName(pNtk, SigValue)
		RangeId = Prs_SliceRange(pNtk, SigValue)
		nBits = Cba_NtkRangeSize(p, RangeId)
		// save this slice
		Vec_IntPushThree(&p.vArray0, NameId, RangeId, iFon)

	} else if SigType == CBA_PRS_CONCAT {
		var vSigs *Vec_Int_t = Prs_CatSignals(pNtk, SigValue)
		Vec_IntReverseOrder(vSigs)
		// inlining of macro Vec_IntForEachEntry
		for i = int32(0); i < Vec_IntSize(vSigs); i++ {
			Sig = Vec_IntEntry(vSigs, i)
			SigType = Prs_ManType_t(Abc_Lit2Att2(Sig))
			SigValue = Abc_Lit2Var2(Sig)
			if SigType == CBA_PRS_NAME {
				var iObjBuf, iFonBuf int32
				// create buffer
				NameId = SigValue
				if strncmp(Cba_NtkStr(p, NameId), "Open_", 5) == 0 {
					nBits++
					continue
				}
				iObjBuf = Cba_ObjAlloc(p, CBA_BOX_BUF, 1, 1)
				iFonBuf = Cba_ObjFon0(p, iObjBuf)
				Cba_FonSetName(p, iFonBuf, NameId)
				RangeSize = Prs_CreateRange(p, iFonBuf, NameId)
				RangeOut = Cba_NtkHashRange(p, nBits+RangeSize-1, nBits)
				// create slice
				iFonNew = Prs_CreateSlice(p, iFon, pNtk, RangeOut)
				Cba_ObjSetFinFon(p, iObjBuf, 0, iFonNew)

			} else if SigType == CBA_PRS_SLICE {
				NameId = Prs_SliceName(pNtk, SigValue)
				RangeId = Prs_SliceRange(pNtk, SigValue)
				RangeSize = Cba_NtkRangeSize(p, RangeId)
				RangeOut = Cba_NtkHashRange(p, nBits+RangeSize-1, nBits)
				// create slice
				iFonNew = Prs_CreateSlice(p, iFon, pNtk, RangeOut)
				// save this slice
				Vec_IntPushThree(&p.vArray0, NameId, RangeId, iFonNew)

			} else {
				assert(false, "0")
			}
			// increment complete range
			nBits += RangeSize
		}
		Vec_IntReverseOrder(vSigs)

	} else {
		assert(false, "0")
	}
	// set the range for the output
	Cba_FonHashRange(p, iFon, nBits-1, 0)
}

func Prs_CreateOutConcat(p *Cba_Ntk_t, pSlices []int32, nSlices int32) {
	var vBits *Vec_Int_t = &p.vArray1
	var NameId int32 = pSlices[0]
	var RangeId int32 = -Cba_NtkGetMap(p, NameId)
	var LeftId int32 = Cba_NtkRangeLeft(p, RangeId)
	var RightId int32 = Cba_NtkRangeRight(p, RangeId)
	var BotId int32 = Abc_MinInt(LeftId, RightId)
	var TopId int32 = Abc_MaxInt(LeftId, RightId)
	var i, k, iObj, iFon, nParts, Prev, nBits int32
	assert(RangeId > 0, "RangeId > 0")
	Vec_IntFill(vBits, Abc_MaxInt(LeftId, RightId)+1, 0)
	// fill up with slices
	for i = 0; i < nSlices; i++ {
		var Range int32 = pSlices[3*i+1]
		var iFon int32 = pSlices[3*i+2]
		var Left int32 = Cba_NtkRangeLeft(p, Range)
		var Right int32 = Cba_NtkRangeRight(p, Range)
		var Bot int32 = Abc_MinInt(Left, Right)
		var Top int32 = Abc_MaxInt(Left, Right)
		assert(NameId == pSlices[3*i+0] && iFon > 0, "NameId == pSlices[3*i+0] && iFon > 0")
		assert(BotId <= Bot && Top <= TopId, "BotId <= Bot && Top <= TopId")
		for k = Bot; k <= Top; k++ {
			assert(Vec_IntEntry(vBits, k) == 0, "Vec_IntEntry(vBits, k) == 0")
			Vec_IntWriteEntry(vBits, k, iFon)
		}
	}
	// check how many parts we have
	Prev = -1
	nParts = 0
	// inlining of macro Vec_IntForEachEntryStartStop
	for i = BotId; i < TopId+1; i++ {
		iFon = Vec_IntEntry(vBits, i)
		if Prev != iFon {
			nParts++
		}
		Prev = iFon
	}
	// create new concatenation
	iObj = Cba_ObjAlloc(p, CBA_BOX_CONCAT, nParts, 1)
	iFon = Cba_ObjFon0(p, iObj)
	Cba_FonSetName(p, iFon, NameId)
	Prs_CreateRange(p, iFon, NameId)
	// set inputs
	k = 0
	Prev = -1
	nBits = 0
	// inlining of macro Vec_IntForEachEntryStartStop
	for i = BotId; i < TopId+1; i++ {
		iFon = Vec_IntEntry(vBits, i)
		if Prev == -1 || Prev == iFon {
			nBits++
		} else {
			if Prev == 0 { // create constant
				Prev = Cba_ManNewConstZero(p, nBits)
			}
			assert(nBits == Cba_FonRangeSize(p, Prev), "nBits == Cba_FonRangeSize(p, Prev)")
			Cba_ObjSetFinFon(p, iObj, nParts-1-k, Prev)
			k++
			nBits = 1
		}
		Prev = iFon
	}
	assert(nBits == Cba_FonRangeSize(p, Prev), "nBits == Cba_FonRangeSize(p, Prev)")
	Cba_ObjSetFinFon(p, iObj, nParts-1-k, Prev)
	k++
	assert(k == nParts, "k == nParts")
}

// looks at multi-bit signal; if one bit is repeated, returns this bit; otherwise, returns -1
func Prs_CreateBitSignal(pNtk *Prs_Ntk_t, Sig int32) int32 {
	var vSigs *Vec_Int_t
	var i int32
	var SigTemp int32
	var SigOne int32 = -1
	var Value int32 = Abc_Lit2Var2(Sig)
	var Type Prs_ManType_t = Prs_ManType_t(Abc_Lit2Att2(Sig))
	if Type == CBA_PRS_NAME || Type == CBA_PRS_SLICE {
		return -1
	}
	if Type == CBA_PRS_CONST {
		var fOnly0 int32 = 1
		var fOnly1 int32 = 1
		var pConst string = Prs_NtkConst(pNtk, Value)
		if k := strings.IndexByte(pConst, '\''); k >= 0 {
			pConst = pConst[k+1:]
		}
		assert(len(pConst) > 0 && pConst[0] == 'b', "constant binary prefix")
		for _, c := range pConst[1:] {
			if c == '0' {
				fOnly1 = 0
			} else if c == '1' {
				fOnly0 = 0
			}
		}
		if fOnly0 != 0 {
			return Abc_Var2Lit2(1, int32(CBA_PRS_CONST))
		} // const0
		if fOnly1 != 0 {
			return Abc_Var2Lit2(2, int32(CBA_PRS_CONST))
		} // const1
		return -1
	}
	assert(Type == CBA_PRS_CONCAT, "Type == CBA_PRS_CONCAT")
	vSigs = Prs_CatSignals(pNtk, Value)
	// inlining of macro Vec_IntForEachEntry
	for i = int32(0); i < Vec_IntSize(vSigs); i++ {
		SigTemp = Vec_IntEntry(vSigs, i)
		Value = Abc_Lit2Var2(SigTemp)
		Type = Prs_ManType_t(Abc_Lit2Att2(SigTemp))
		if Type != CBA_PRS_NAME {
			return -1
		}
		if SigOne == -1 {
			SigOne = Value
		} else if SigOne != Value {
			return -1
		}
	}
	assert(SigOne >= 0, "SigOne >= 0")
	return Abc_Var2Lit2(SigOne, int32(CBA_PRS_NAME))
}

func Prs_CreateFlopSetReset(p *Cba_Ntk_t, pNtk *Prs_Ntk_t, vBox *Vec_Int_t, pIndexSet *int32, pIndexRst *int32, pBitSet *int32, pBitRst *int32) int32 {
	var iSigSet int32 = -1
	var iSigRst int32 = -1
	var IndexSet int32 = -1
	var IndexRst int32 = -1
	var FormId, ActId, k int32
	// mark set and reset
	Cba_NtkCleanMap2(p)
	Cba_NtkSetMap2(p, Cba_NtkStrId(p, "set"), 1)
	Cba_NtkSetMap2(p, Cba_NtkStrId(p, "reset"), 2)
	// check the inputs
	// inlining of macro Vec_IntForEachEntryDouble
	for k = int32(0); k+1 < Vec_IntSize(vBox); k += 2 {
		FormId, ActId = Vec_IntEntry(vBox, k), Vec_IntEntry(vBox, k+1)
		if Cba_NtkGetMap2(p, FormId) == 1 { // set
			iSigSet = ActId
			IndexSet = k + 1
		} else if Cba_NtkGetMap2(p, FormId) == 2 { // reset
			iSigRst = ActId
			IndexRst = k + 1
		}
	}
	assert(iSigSet >= 0 && iSigRst >= 0, "iSigSet >= 0 && iSigRst >= 0")
	if pIndexSet != nil {
		*pBitSet = 0
	}
	if pIndexRst != nil {
		*pBitRst = 0
	}
	if pBitSet != nil {
		*pBitSet = 0
	}
	if pBitRst != nil {
		*pBitRst = 0
	}
	if iSigSet == -1 || iSigRst == -1 {
		return 0
	}
	iSigSet = Prs_CreateBitSignal(pNtk, iSigSet)
	iSigRst = Prs_CreateBitSignal(pNtk, iSigRst)
	if iSigSet == -1 || iSigRst == -1 {
		return 0
	}
	if pIndexSet != nil {
		*pIndexSet = IndexSet
	}
	if pIndexRst != nil {
		*pIndexRst = IndexRst
	}
	if pBitSet != nil {
		*pBitSet = iSigSet
	}
	if pBitRst != nil {
		*pBitRst = iSigRst
	}
	return 1
}
func Prs_CreateDetectRamPort(pNtk *Prs_Ntk_t, vBox *Vec_Int_t, NameRamId int32) string {
	var i, FormId, ActId int32
	// inlining of macro Vec_IntForEachEntryDouble
	for i = int32(0); i+1 < Vec_IntSize(vBox); i += 2 {
		FormId, ActId = Vec_IntEntry(vBox, i), Vec_IntEntry(vBox, i+1)
		if FormId == NameRamId {
			return Abc_NamStr(pNtk.pStrs, Abc_Lit2Var2(ActId))
		}
	}
	return ""
}
func Prs_CreateGetMemSize(pName string) int32 {
	first := strings.IndexByte(pName, '_')
	if first < 0 {
		return 0
	}
	secondRel := strings.IndexByte(pName[first+1:], '_')
	if secondRel < 0 {
		return 0
	}
	second := first + 1 + secondRel
	num1, _ := strconv.Atoi(pName[first+1 : second])
	num2, _ := strconv.Atoi(pName[second+1:])
	assert(int32(num1)+Abc_Base2Log(uint32(num2)) < 32, "Num1 + Abc_Base2Log(Num2) < 32")
	return (1 << uint(num1)) * int32(num2)
}
func Prs_CreateDetectRams(pNtk *Prs_Ntk_t) *Vec_Ptr_t {
	var vAllRams *Vec_Ptr_t = nil
	var vRam *Vec_Ptr_t
	var vBox, vBoxCopy *Vec_Int_t
	var pNtkName, pRamName string
	var NameRamId int32 = Abc_NamStrFind(pNtk.pStrs, "Ram")
	var i, k, fWrite int32
	// inlining of macro Prs_NtkForEachBox
	for i = int32(0); i < Prs_NtkBoxNum(pNtk); i++ {
		vBox = Prs_BoxSignals(pNtk, i)
		if Prs_BoxIsNode(pNtk, i) != 0 { // node
			continue
		}
		pNtkName = Prs_NtkStr(pNtk, Prs_BoxNtk(pNtk, i))
		fWrite = boolToInt(strncmp(pNtkName, "ClockedWritePort_", int32(len("ClockedWritePort_"))) == 0)
		if fWrite != 0 || strings.HasPrefix(pNtkName, "ReadPort_") {
			pRamName = Prs_CreateDetectRamPort(pNtk, vBox, NameRamId)
			assert(pRamName != "", "pRamName")
			if vAllRams == nil {
				vAllRams = Vec_PtrAlloc(4)
			}
			// inlining of macro Vec_PtrForEachEntry
			for k = int32(0); k < Vec_PtrSize(vAllRams); k++ {
				vRam = Vec_PtrEntry(vAllRams, k).(*Vec_Ptr_t)
				if pRamName == Vec_PtrEntry(vRam, 0).(string) {
					if fWrite != 0 {
						vBoxCopy = Vec_IntDup(vBox)
						Vec_IntPush(vBoxCopy, i)
						Vec_PtrPush(vRam, vBoxCopy)
					}
					break
				}
			}
			if k < Vec_PtrSize(vAllRams) {
				continue
			}
			vRam = Vec_PtrAlloc(4)
			Vec_PtrPush(vRam, pRamName)
			Vec_PtrPush(vRam, Abc_Int2Ptr(Prs_CreateGetMemSize(pNtkName)))
			if fWrite != 0 {
				vBoxCopy = Vec_IntDup(vBox)
				Vec_IntPush(vBoxCopy, i)
				Vec_PtrPush(vRam, vBoxCopy)
			}
			Vec_PtrPush(vAllRams, vRam)
		}
	}
	return vAllRams
}
func Prs_CreateVerilogPio(p *Cba_Ntk_t, pNtk *Prs_Ntk_t) {
	var i, NameId, RangeId, iObj, iFon int32
	Cba_NtkCleanObjFuncs(p)
	Cba_NtkCleanObjNames(p)
	Cba_NtkCleanFonNames(p)
	Cba_NtkCleanFonRanges(p)
	// create inputs
	Cba_NtkCleanMap(p)
	assert(Vec_IntSize(&pNtk.vInouts) == 0, "Vec_IntSize(&pNtk.vInouts) == 0")
	// inlining of macro Vec_IntForEachEntryTwo
	for i = int32(0); i < Vec_IntSize(&pNtk.vInputs); i++ {
		NameId, RangeId = Vec_IntEntry(&pNtk.vInputs, i), Vec_IntEntry(&pNtk.vInputsR, i)
		iObj = Cba_ObjAlloc(p, CBA_OBJ_PI, 0, 1)
		Cba_ObjSetName(p, iObj, NameId) // direct name
		iFon = Cba_ObjFon0(p, iObj)
		Cba_FonSetRangeSign(p, iFon, RangeId)
		Cba_FonSetName(p, iFon, NameId)
		Cba_NtkSetMap(p, NameId, iObj)
	}
	// create outputs
	// inlining of macro Vec_IntForEachEntryTwo
	for i = int32(0); i < Vec_IntSize(&pNtk.vOutputs); i++ {
		NameId, RangeId = Vec_IntEntry(&pNtk.vOutputs, i), Vec_IntEntry(&pNtk.vOutputsR, i)
		iObj = Cba_ObjAlloc(p, CBA_OBJ_PO, 1, 0)
		Cba_ObjSetName(p, iObj, NameId) // direct name
		Cba_NtkSetMap(p, NameId, iObj)
	}
	// create order
	// inlining of macro Vec_IntForEachEntry
	for i = int32(0); i < Vec_IntSize(&pNtk.vOrder); i++ {
		NameId = Vec_IntEntry(&pNtk.vOrder, i)
		iObj = Prs_CreateVerilogFindFon(p, Abc_Lit2Var2(NameId)) // labeled name
		if iObj != 0 {
			Vec_IntPush(&p.vOrder, iObj)
		}
	}
}

func Prs_CreateVerilogNtk(p *Cba_Ntk_t, pNtk *Prs_Ntk_t) int32 {
	vBox2Obj := Vec_IntStart(Prs_NtkBoxNum(pNtk))
	var vBox *Vec_Int_t
	var i, k, iObj, iTerm, iFon, FormId, ActId, RangeId, NameId, Type int32
	formalName := func(module string, output bool, index int32) string {
		for j := 1; j < len(s_VerInfo) && s_VerInfo[j].pTypeName != ""; j++ {
			info := &s_VerInfo[j]
			if strings.HasPrefix(module, info.pTypeName) {
				offset := 0
				if output {
					offset = int(info.nInputs)
				}
				return info.pSigNames[offset+int(index)]
			}
		}
		return ""
	}
	Cba_NtkCleanMap(p)
	for i = 0; i < Cba_NtkPiNum(p); i++ {
		iObj = Cba_NtkPi(p, i)
		Cba_NtkSetMap(p, Cba_ObjName(p, iObj), Cba_ObjFon0(p, iObj))
	}
	for i = 0; i < Vec_IntSize(&pNtk.vWires); i++ {
		NameId, RangeId = Vec_IntEntry(&pNtk.vWires, i), Vec_IntEntry(&pNtk.vWiresR, i)
		Cba_NtkSetMap(p, NameId, -RangeId)
	}
	for i = 0; i < Vec_IntSize(&pNtk.vOutputs); i++ {
		NameId, RangeId = Vec_IntEntry(&pNtk.vOutputs, i), Vec_IntEntry(&pNtk.vOutputsR, i)
		Cba_NtkSetMap(p, NameId, -RangeId)
	}

	// RAM groups own their read object and write-port objects. Each write
	// group's saved box has its line number appended as its last entry.
	vAllRams := Prs_CreateDetectRams(pNtk)
	if vAllRams != nil {
		for i = 0; i < int32(Vec_PtrSize(vAllRams)); i++ {
			vRam := Vec_PtrEntry(vAllRams, i).(*Vec_Ptr_t)
			pRamName := Vec_PtrEntry(vRam, 0).(string)
			MemSize := Abc_Ptr2Int(Vec_PtrEntry(vRam, 1).(util.AbcPointerValue))
			NameId = Cba_NtkNewStrId(p, "%s_box", pRamName)
			iObj = Cba_ObjAlloc(p, CBA_BOX_RAMBOX, int32(Vec_PtrSize(vRam))-2, 1)
			Cba_ObjSetName(p, iObj, NameId)
			iFon = Cba_ObjFon0(p, iObj)
			NameId = Cba_NtkNewStrId(p, "%s", pRamName)
			Cba_FonSetName(p, iFon, NameId)
			Prs_CreateRange(p, iFon, NameId)
			assert(Cba_FonLeft(p, iFon) <= MemSize-1, "RAM range left")
			assert(Cba_FonRight(p, iFon) == 0, "RAM range right")
			for k = 2; k < int32(Vec_PtrSize(vRam)); k++ {
				vBox = Vec_PtrEntry(vRam, k).(*Vec_Int_t)
				iObjNew := Cba_ObjAlloc(p, CBA_BOX_RAMWC, 4, 1)
				Line := Vec_IntPop(vBox)
				Vec_IntWriteEntry(vBox2Obj, Line, iObjNew)
				if Prs_BoxName(pNtk, Line) != 0 {
					Cba_ObjSetName(p, iObjNew, Prs_BoxName(pNtk, Line))
				}
				iFon = Cba_ObjFon0(p, iObjNew)
				Cba_FonSetRange(p, iFon, Cba_NtkHashRange(p, MemSize-1, 0))
				NameId = Cba_NtkNewStrId(p, "%s_wp%d", pRamName, k-2)
				Cba_FonSetName(p, iFon, NameId)
				Cba_NtkSetMap(p, NameId, iFon)
				Cba_ObjSetFinFon(p, iObj, k-2, iFon)
				k++ // mirrors the original RAM-port loop index increment
				Vec_IntFree(vBox)
			}
			Vec_PtrFree(vRam)
		}
		Vec_PtrFreeP(&vAllRams)
	}

	Vec_IntClear(&p.vArray0)
	for i = 0; i < Prs_NtkBoxNum(pNtk); i++ {
		vBox = Prs_BoxSignals(pNtk, i)
		if Prs_BoxIsNode(pNtk, i) != 0 {
			Type = Prs_BoxNtk(pNtk, i)
			iObj = Cba_ObjAlloc(p, Cba_ObjType_t(Type), Prs_BoxIONum(pNtk, i)-1, boolToInt(Type == CBA_BOX_ADD)+1)
			Prs_CreateSignalOut(p, Cba_ObjFon0(p, iObj), pNtk, Vec_IntEntry(vBox, 1))
		} else {
			var pBox *Cba_Ntk_t
			var nInputs int32
			nOutputs := int32(1)
			var pOutNames *string
			pNtkName := Prs_NtkStr(pNtk, Prs_BoxNtk(pNtk, i))
			Type = Prs_ManFindType(pNtkName, &nInputs, 1, &pOutNames)
			if Type == CBA_BOX_RAMWC {
				continue
			}
			if Type == CBA_OBJ_BOX {
				pBox = Cba_ManNtkFind(p.pDesign, pNtkName)
				if pBox == nil {
					fmt.Printf("Fatal error: Cannot find module %q.\n", pNtkName)
					continue
				}
				nInputs, nOutputs = Cba_NtkPiNum(pBox), Cba_NtkPoNum(pBox)
			} else if Type == CBA_BOX_ADD || Type == CBA_BOX_DFFCPL {
				nOutputs = 2
			} else if Type == CBA_BOX_NMUX {
				prefix := "Mux_"
				if strings.HasPrefix(pNtkName, "wide_mux_") {
					prefix = "wide_mux_"
				}
				if !strings.HasPrefix(pNtkName, prefix) {
					assert(false, "mux name")
				}
				width, _ := strconv.Atoi(pNtkName[len(prefix):])
				nInputs = 1 + int32(1<<uint(width))
			} else if Type == CBA_BOX_SEL {
				prefix := "Select_"
				if strings.HasPrefix(pNtkName, "wide_select_") {
					prefix = "wide_select_"
				}
				width, _ := strconv.Atoi(pNtkName[len(prefix):])
				nInputs = 1 + int32(width)
			} else if (Type == CBA_BOX_DFFRS || Type == CBA_BOX_LATCHRS) && strings.HasPrefix(pNtkName, "wide_") && Prs_CreateFlopSetReset(p, pNtk, vBox, nil, nil, nil, nil) == 0 {
				prefix := "wide_dffrs_"
				if Type == CBA_BOX_LATCHRS {
					prefix = "wide_latchrs_"
				}
				width, _ := strconv.Atoi(pNtkName[len(prefix):])
				nInputs, nOutputs, Type = int32(width), 1, CBA_BOX_CONCAT
			}
			iObj = Cba_ObjAlloc(p, Cba_ObjType_t(Type), nInputs, nOutputs)
			if pBox != nil {
				Cba_ObjSetFunc(p, iObj, Cba_NtkId(pBox))
			}
			Cba_NtkCleanMap2(p)
			if pBox != nil {
				for k = 0; k < Cba_NtkPoNum(pBox); k++ {
					iTerm = Cba_NtkPo(pBox, k)
					Cba_NtkSetMap2(p, Cba_ObjName(pBox, iTerm), k+1)
				}
			} else {
				for k = 0; k < nOutputs; k++ {
					Cba_NtkSetMap2(p, Cba_NtkStrId(p, formalName(pNtkName, true, k)), k+1)
				}
			}
			for k = 0; k+1 < Vec_IntSize(vBox); k += 2 {
				FormId, ActId = Vec_IntEntry(vBox, k), Vec_IntEntry(vBox, k+1)
				if Cba_NtkGetMap2(p, FormId) != 0 {
					iFon = Cba_ObjFon(p, iObj, Cba_NtkGetMap2(p, FormId)-1)
					Prs_CreateSignalOut(p, iFon, pNtk, ActId)
				}
			}
		}
		Vec_IntWriteEntry(vBox2Obj, i, iObj)
		if Prs_BoxName(pNtk, i) != 0 {
			Cba_ObjSetName(p, iObj, Prs_BoxName(pNtk, i))
		}
	}

	// Join split output signals in the parser's range/name order.
	if Vec_IntSize(&p.vArray0) != 0 {
		Prev, Index := int32(-1), int32(0)
		Vec_IntSortMulti(&p.vArray0, 3, 0)
		for i = 0; i+2 < Vec_IntSize(&p.vArray0); i += 3 {
			NameId, RangeId, iFon = Vec_IntEntry(&p.vArray0, i), Vec_IntEntry(&p.vArray0, i+1), Vec_IntEntry(&p.vArray0, i+2)
			if Prev != -1 && Prev != NameId {
				Prs_CreateOutConcat(p, Vec_IntArray(&p.vArray0)[Index:], (i-Index)/3)
				Index = i
			}
			Prev = NameId
		}
		Prs_CreateOutConcat(p, Vec_IntArray(&p.vArray0)[Index:], (i-Index)/3)
	}

	for i = 0; i < Prs_NtkBoxNum(pNtk); i++ {
		vBox = Prs_BoxSignals(pNtk, i)
		iObj = Vec_IntEntry(vBox2Obj, i)
		if Prs_BoxIsNode(pNtk, i) != 0 {
			Type = Prs_BoxNtk(pNtk, i)
			for k = 2; k+1 < Vec_IntSize(vBox); k += 2 {
				ActId = Vec_IntEntry(vBox, k+1)
				iFon = Prs_CreateSignalIn(p, pNtk, ActId)
				if iFon != 0 {
					Cba_ObjSetFinFon(p, iObj, k/2-1, iFon)
				}
			}
		} else {
			var nInputs int32 = -1
			var pInNames *string
			pNtkName := Prs_NtkStr(pNtk, Prs_BoxNtk(pNtk, i))
			Type = Prs_ManFindType(pNtkName, &nInputs, 0, &pInNames)
			if (Type == CBA_BOX_DFFRS || Type == CBA_BOX_LATCHRS) && strings.HasPrefix(pNtkName, "wide_") {
				var IndexSet, IndexRst, iBitSet, iBitRst int32 = -1, -1, -1, -1
				if Prs_CreateFlopSetReset(p, pNtk, vBox, &IndexSet, &IndexRst, &iBitSet, &iBitRst) != 0 {
					Vec_IntWriteEntry(vBox, IndexSet, iBitSet)
					Vec_IntWriteEntry(vBox, IndexRst, iBitRst)
				} else {
					prefix := "wide_dffrs_"
					if Type == CBA_BOX_LATCHRS {
						prefix = "wide_latchrs_"
					}
					width, _ := strconv.Atoi(pNtkName[len(prefix):])
					Width := int32(width)
					assert(Cba_ObjType(p, iObj) == CBA_BOX_CONCAT, "wide flop concat")
					assert(nInputs >= 0, "wide flop formal count")
					Cba_NtkCleanMap2(p)
					for k = 0; k < nInputs; k++ {
						Cba_NtkSetMap2(p, Cba_NtkStrId(p, formalName(pNtkName, false, k)), k+1)
					}
					for w := int32(0); w < int32(Width); w++ {
						iObjNew := Cba_ObjAlloc(p, Cba_ObjType_t(Type), 4, 1)
						if Prs_BoxName(pNtk, i) != 0 {
							NameId = Cba_NtkNewStrId(p, "%s[%d]", Prs_NtkStr(pNtk, Prs_BoxName(pNtk, i)), w)
							Cba_ObjSetName(p, iObjNew, NameId)
						}
						iFon = Cba_ObjFon0(p, iObjNew)
						NameId = Cba_NtkNewStrId(p, "%s[%d]", Cba_FonNameStr(p, Cba_ObjFon0(p, iObj)), w)
						Cba_FonSetName(p, iFon, NameId)
						Cba_ObjSetFinFon(p, iObj, Width-1-w, iFon)
						for k = 0; k+1 < Vec_IntSize(vBox); k += 2 {
							FormId, ActId = Vec_IntEntry(vBox, k), Vec_IntEntry(vBox, k+1)
							if Cba_NtkGetMap2(p, FormId) != 0 {
								Index := Cba_NtkGetMap2(p, FormId) - 1
								iFon = Prs_CreateSignalIn(p, pNtk, ActId)
								assert(iFon != 0, "wide flop input")
								if Index < 3 {
									iFon = Prs_CreateSlice(p, iFon, pNtk, 0)
								}
								Cba_ObjSetFinFon(p, iObjNew, Index, iFon)
							}
						}
					}
					continue
				}
			}
			assert(Type == Cba_ObjType(p, iObj), "box type")
			Cba_NtkCleanMap2(p)
			if Type == CBA_OBJ_BOX {
				pBox := Cba_ObjNtk(p, iObj)
				assert(Cba_NtkPiNum(pBox) == Cba_ObjFinNum(p, iObj), "box input count")
				assert(Cba_NtkPoNum(pBox) == Cba_ObjFonNum(p, iObj), "box output count")
				for k = 0; k < Cba_NtkPiNum(pBox); k++ {
					iTerm = Cba_NtkPi(pBox, k)
					Cba_NtkSetMap2(p, Cba_ObjName(pBox, iTerm), k+1)
				}
			} else {
				assert(nInputs >= 0, "box formal count")
				for k = 0; k < nInputs; k++ {
					Cba_NtkSetMap2(p, Cba_NtkStrId(p, formalName(pNtkName, false, k)), k+1)
				}
			}
			for k = 0; k+1 < Vec_IntSize(vBox); k += 2 {
				FormId, ActId = Vec_IntEntry(vBox, k), Vec_IntEntry(vBox, k+1)
				if Cba_NtkGetMap2(p, FormId) != 0 {
					Index := Cba_NtkGetMap2(p, FormId) - 1
					assert(Index < Cba_ObjFinNum(p, iObj), "box fanin index")
					iFon = Prs_CreateSignalIn(p, pNtk, ActId)
					if iFon != 0 {
						Cba_ObjSetFinFon(p, iObj, Index, iFon)
					}
				}
			}
			if Type == CBA_BOX_NMUX || Type == CBA_BOX_SEL {
				FonCat := Cba_ObjFinFon(p, iObj, 1)
				nBits := Cba_FonRangeSize(p, FonCat)
				nParts := Cba_ObjFinNum(p, iObj) - 1
				assert(nParts > 0, "mux parts")
				Slice := nBits / nParts
				assert(Slice*nParts == nBits, "mux slice width")
				assert(Cba_ObjFinNum(p, iObj) == 1+nParts, "mux fins")
				Cba_ObjCleanFinFon(p, iObj, 1)
				if FonCat < 0 {
					iObjNew := Cba_ObjAlloc(p, CBA_BOX_BUF, 1, 1)
					Cba_ObjSetFinFon(p, iObjNew, 0, FonCat)
					FonCat = Cba_ObjFon0(p, iObjNew)
					NameId = Cba_NtkNewStrId(p, "_buf_const_%d", iObjNew)
					Cba_FonSetName(p, FonCat, NameId)
					Cba_FonSetRange(p, FonCat, Cba_NtkHashRange(p, nBits-1, 0))
				}
				for k = 0; k < nParts; k++ {
					iFon = Prs_CreateSlice(p, FonCat, pNtk, Cba_NtkHashRange(p, k*Slice+Slice-1, k*Slice))
					Cba_ObjSetFinFon(p, iObj, k+1, iFon)
				}
			}
		}
		if Type == CBA_BOX_ADD && Cba_ObjFinFon(p, iObj, 0) == 0 {
			Cba_ObjSetFinFon(p, iObj, 0, Cba_FonFromConst(1))
		}
		if Type == CBA_BOX_DFFRS && Cba_ObjFinFon(p, iObj, 1) == 0 {
			Cba_ObjSetFinFon(p, iObj, 1, Cba_FonFromConst(1))
		}
		if Type == CBA_BOX_DFFRS && Cba_ObjFinFon(p, iObj, 2) == 0 {
			Cba_ObjSetFinFon(p, iObj, 2, Cba_FonFromConst(1))
		}
	}
	Vec_IntFree(vBox2Obj)
	for i = 0; i < Vec_IntSize(&pNtk.vOutputs); i++ {
		NameId, RangeId = Vec_IntEntry(&pNtk.vOutputs, i), Vec_IntEntry(&pNtk.vOutputsR, i)
		iObj = Cba_NtkPo(p, i)
		assert(NameId == Cba_ObjName(p, iObj), "output name")
		iFon = Prs_CreateVerilogFindFon(p, NameId)
		if iFon == 0 {
			continue
		}
		Cba_ObjSetFinFon(p, iObj, 0, iFon)
		if RangeId != 0 {
			assert(Cba_NtkRangeLeft(p, Abc_Lit2Var(RangeId)) == Cba_FonLeft(p, iFon), "output range left")
			assert(Cba_NtkRangeRight(p, Abc_Lit2Var(RangeId)) == Cba_FonRight(p, iFon), "output range right")
		}
	}
	return 0
}
func Prs_ManBuildCbaVerilog(pFileName string, vDes *Vec_Ptr_t) *Cba_Man_t {
	var i, fError int32
	pPrsRoot := Prs_ManRoot(vDes)
	pStrs := Abc_NamRef(pPrsRoot.pStrs)
	pFuns := Abc_NamRef(pPrsRoot.pFuns)
	pMods := Abc_NamStart(100, 24)
	p := Cba_ManAlloc(pFileName, int32(Vec_PtrSize(vDes)), pStrs, pFuns, pMods, Hash_IntManRef(pPrsRoot.vHash))

	// Allocate all module networks before building any instances so that
	// references to modules declared later in the file can be resolved.
	for i = 0; i < Vec_PtrSize(vDes); i++ {
		pPrsNtk := Prs_ManNtk(vDes, i)
		pNtk := Cba_NtkAlloc(p, Prs_NtkId(pPrsNtk), Prs_NtkPiNum(pPrsNtk), Prs_NtkPoNum(pPrsNtk), Prs_NtkObjNum(pPrsNtk), 100, 100)
		Prs_CreateVerilogPio(pNtk, pPrsNtk)
		Cba_NtkAdd(p, pNtk)
	}

	for i = 0; i < Vec_PtrSize(vDes); i++ {
		pPrsNtk := Prs_ManNtk(vDes, i)
		fmt.Printf("Building module %q...\n", Prs_NtkName(pPrsNtk))
		fError = Prs_CreateVerilogNtk(Cba_ManNtk(p, i+1), pPrsNtk)
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

func Cba_ManReadVerilog(pFileName string) *Cba_Man_t {
	var p *Cba_Man_t
	vDes := Prs_ManReadVerilog(pFileName)
	if vDes != nil && Vec_PtrSize(vDes) != 0 {
		p = Prs_ManBuildCbaVerilog(pFileName, vDes)
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
