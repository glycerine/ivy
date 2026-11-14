package wln

import (
	"fmt"
	"os"

	"github.com/glycerine/abcgo/misc/extra"
)

func signedness(p *Wln_Ntk_t, iObj int32) string {
	if p.fSmtLib == 0 && Wln_ObjIsSigned(p, iObj) != 0 {
		return "signed "
	}
	return "       "
}

func faninSeparator(k int32) string {
	if k != 0 {
		return " or "
	}
	return ""
}

func bitIndexPrefix(index int32) string {
	if index != 0 {
		return " & "
	}
	return ""
}
func bitPolarityPrefix(bit int32) string {
	if bit != 0 {
		return " "
	}
	return "~"
}
func commaIfMore(index, count int32) string {
	if index < count-1 {
		return ", "
	}
	return ""
}
func commaIfLast(index, count int32) string {
	if index == count-1 {
		return ""
	}
	return ","
}
func ramModuleName(kind int32, read, write string) string {
	if kind == ABC_OPER_RAMR {
		return read
	}
	return write
}
func ramPortName(index int32, memory, address, data string) string {
	if index == 0 {
		return memory
	}
	if index == 1 {
		return address
	}
	return data
}
func smtSignedComment(p *Wln_Ntk_t, obj int32) string {
	if p.fSmtLib != 0 && Wln_ObjIsSigned(p, obj) != 0 {
		return " // signed SMT-LIB operator"
	}
	return ""
}
func specOrUnknown(spec string) string {
	if spec != "" {
		return spec
	}
	return "unknown"
}

/**CFile****************************************************************

  FileName    [wlnWriteVer.c]

  SystemName  [ABC: Logic synthesis and verification system.]

  PackageName [Word-level network.]

  Synopsis    [Writing Verilog.]

  Author      [Alan Mishchenko]

  Affiliation [UC Berkeley]

  Date        [Ver. 1.0. Started - September 23, 2018.]

  Revision    [$Id: wlnWriteVer.c,v 1.00 2018/09/23 00:00:00 alanmi Exp $]

***********************************************************************/

//#include "wln.h"

//ABC_NAMESPACE_IMPL_START

////////////////////////////////////////////////////////////////////////
///                        DECLARATIONS                              ///
////////////////////////////////////////////////////////////////////////

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
func Wln_WriteVerInt(pFile *os.File, p *Wln_Ntk_t) {
	var k, j, iObj, iFanin int32
	var Range string // was [100]byte in C

	fmt.Fprintf(pFile, "module %s ( ", p.pName)
	fmt.Fprintf(pFile, "\n   ")
	if Wln_NtkCiNum(p) > 0 {
		Wln_WriteVerIntVec(pFile, p, &p.vCis, 3)
		fmt.Fprintf(pFile, ",\n   ")
	}
	if Wln_NtkCoNum(p) > 0 {
		Wln_WriteVerIntVec(pFile, p, &p.vCos, 3)
	}
	fmt.Fprintf(pFile, "  );\n")
	// inlining of macro Wln_NtkForEachObj
	for iObj = 1; iObj < Wln_NtkObjNum(p); iObj++ {
		var End int32 = Wln_ObjRangeEnd(p, iObj)
		var Beg int32 = Wln_ObjRangeBeg(p, iObj)
		var nDigits int32 = Abc_Base10Log(uint32(Abc_AbsInt(End)+1)) + Abc_Base10Log(uint32(Abc_AbsInt(Beg)+1)) + wlnBool(End < 0) + wlnBool(Beg < 0)
		Range = fmt.Sprintf("%s[%d:%d]%*s", signedness(p, iObj), End, Beg, 8-nDigits, "")
		fmt.Fprintf(pFile, "  ")
		if Wln_ObjIsCi(p, iObj) != 0 {
			fmt.Fprintf(pFile, "input  ")
		} else if Wln_ObjIsCo(p, iObj) != 0 {
			fmt.Fprintf(pFile, "output ")
		} else {
			fmt.Fprintf(pFile, "       ")
		}
		if Wln_ObjIsCio(p, iObj) != 0 {
			fmt.Fprintf(pFile, "wire %s %s ;\n", Range, Wln_ObjName(p, iObj))
			if Wln_ObjIsCi(p, iObj) != 0 {
				continue
			}
			fmt.Fprintf(pFile, "  assign                         ")
			fmt.Fprintf(pFile, "%-16s = %s ;\n", Wln_ObjName(p, iObj), Wln_ObjName(p, Wln_ObjFanin0(p, iObj)))
			continue
		}
		if Wln_ObjType(p, iObj) == ABC_OPER_SEL_NMUX || Wln_ObjType(p, iObj) == ABC_OPER_SEL_SEL {
			fmt.Fprintf(pFile, "reg  %s ", Range)
		} else {
			fmt.Fprintf(pFile, "wire %s ", Range)
		}
		if Wln_ObjType(p, iObj) == ABC_OPER_TABLE {
			// wire [3:0] s4972; table0 s4972_Index(s4971, s4972);
			fmt.Fprintf(pFile, "%s ;              table%d", Wln_ObjName(p, iObj), Wln_ObjFanin1(p, iObj))
			fmt.Fprintf(pFile, " s%d_Index(%s, ", iObj, Wln_ObjName(p, Wln_ObjFanin0(p, iObj)))
			fmt.Fprintf(pFile, "%s)", Wln_ObjName(p, iObj))

		} else if Wln_ObjType(p, iObj) == ABC_OPER_LUT {
			// wire [3:0] s4972; LUT lut4972_Index(s4971, s4972);
			fmt.Fprintf(pFile, "%s ;           LUT", Wln_ObjName(p, iObj))
			fmt.Fprintf(pFile, " lut%d (%s, ", iObj, Wln_ObjName(p, Wln_ObjFanin0(p, iObj)))
			for k = 1; k < Wln_ObjFaninNum(p, iObj); k++ {
				fmt.Fprintf(pFile, "%s, ", Wln_ObjName(p, Wln_ObjFanin(p, iObj, k)))
			}
			fmt.Fprintf(pFile, "%s)", Wln_ObjName(p, iObj))

		} else if Wln_ObjIsConst(p, iObj) != 0 {
			fmt.Fprintf(pFile, "%-16s = %s", Wln_ObjName(p, iObj), Wln_ObjConstString(p, iObj))
		} else if Wln_ObjType(p, iObj) == ABC_OPER_SHIFT_ROTR || Wln_ObjType(p, iObj) == ABC_OPER_SHIFT_ROTL {
			//  wire [27:0] s4960 = (s57 >> 17) | (s57 << 11);
			var Num0 int32 = Wln_ObjFanin1(p, iObj)
			var Num1 int32 = Wln_ObjRange(p, iObj) - Num0
			assert(Num0 > 0 && Num0 < Wln_ObjRange(p, iObj), "Num0 > 0 && Num0 < Wln_ObjRange(p, iObj)")
			fmt.Fprintf(pFile, "%-16s = ", Wln_ObjName(p, iObj))
			if Wln_ObjType(p, iObj) == ABC_OPER_SHIFT_ROTR {
				fmt.Fprintf(pFile, "(%s >> %d) | (%s << %d)", Wln_ObjName(p, Wln_ObjFanin0(p, iObj)), Num0, Wln_ObjName(p, Wln_ObjFanin0(p, iObj)), Num1)
			} else {
				fmt.Fprintf(pFile, "(%s << %d) | (%s >> %d)", Wln_ObjName(p, Wln_ObjFanin0(p, iObj)), Num0, Wln_ObjName(p, Wln_ObjFanin0(p, iObj)), Num1)
			}

		} else if Wln_ObjType(p, iObj) == ABC_OPER_SEL_NMUX {
			fmt.Fprintf(pFile, "%s ;\n", Wln_ObjName(p, iObj))
			fmt.Fprintf(pFile, "         ")
			fmt.Fprintf(pFile, "always @( ")
			// inlining of macro Wln_ObjForEachFanin
			for k = 0; k < Wln_ObjFaninNum(p, iObj); k++ {
				iFanin = Wln_ObjFanin(p, iObj, k)
				if !(iFanin != 0) {
				} else {
					fmt.Fprintf(pFile, "%s%s", faninSeparator(k), Wln_ObjName(p, Wln_ObjFanin(p, iObj, k)))
				}
			}
			fmt.Fprintf(pFile, " )\n")
			fmt.Fprintf(pFile, "           ")
			fmt.Fprintf(pFile, "begin\n")
			fmt.Fprintf(pFile, "             ")
			fmt.Fprintf(pFile, "case ( %s )\n", Wln_ObjName(p, Wln_ObjFanin(p, iObj, 0)))
			// inlining of macro Wln_ObjForEachFanin
			for k = 0; k < Wln_ObjFaninNum(p, iObj); k++ {
				iFanin = Wln_ObjFanin(p, iObj, k)
				if !(iFanin != 0) {
				} else {
					if k == 0 {
						continue
					}
					fmt.Fprintf(pFile, "               ")
					fmt.Fprintf(pFile, "%d : %s = ", k-1, Wln_ObjName(p, iObj))
					fmt.Fprintf(pFile, "%s ;\n", Wln_ObjName(p, Wln_ObjFanin(p, iObj, k)))
				}
			}
			fmt.Fprintf(pFile, "             ")
			fmt.Fprintf(pFile, "endcase\n")
			fmt.Fprintf(pFile, "           ")
			fmt.Fprintf(pFile, "end\n")
			continue

		} else if Wln_ObjType(p, iObj) == ABC_OPER_SEL_SEL {
			fmt.Fprintf(pFile, "%s ;\n", Wln_ObjName(p, iObj))
			fmt.Fprintf(pFile, "         ")
			fmt.Fprintf(pFile, "always @( ")
			// inlining of macro Wln_ObjForEachFanin
			for k = 0; k < Wln_ObjFaninNum(p, iObj); k++ {
				iFanin = Wln_ObjFanin(p, iObj, k)
				if !(iFanin != 0) {
				} else {
					fmt.Fprintf(pFile, "%s%s", faninSeparator(k), Wln_ObjName(p, Wln_ObjFanin(p, iObj, k)))
				}
			}
			fmt.Fprintf(pFile, " )\n")
			fmt.Fprintf(pFile, "           ")
			fmt.Fprintf(pFile, "begin\n")
			fmt.Fprintf(pFile, "             ")
			fmt.Fprintf(pFile, "case ( %s )\n", Wln_ObjName(p, Wln_ObjFanin(p, iObj, 0)))
			// inlining of macro Wln_ObjForEachFanin
			for k = 0; k < Wln_ObjFaninNum(p, iObj); k++ {
				iFanin = Wln_ObjFanin(p, iObj, k)
				if !(iFanin != 0) {
				} else {
					if k == 0 {
						continue
					}
					fmt.Fprintf(pFile, "               ")
					fmt.Fprintf(pFile, "%d'b", Wln_ObjFaninNum(p, iObj)-1)
					for j = Wln_ObjFaninNum(p, iObj) - 1; j > 0; j-- {
						fmt.Fprintf(pFile, "%d", wlnBool(j == k))
					}
					fmt.Fprintf(pFile, " : %s = ", Wln_ObjName(p, iObj))
					fmt.Fprintf(pFile, "%s ;\n", Wln_ObjName(p, Wln_ObjFanin(p, iObj, k)))
				}
			}
			fmt.Fprintf(pFile, "               ")
			fmt.Fprintf(pFile, "default")
			fmt.Fprintf(pFile, " : %s = ", Wln_ObjName(p, iObj))
			fmt.Fprintf(pFile, "%d'b", Wln_ObjRange(p, iObj))
			for j = Wln_ObjRange(p, iObj) - 1; j >= 0; j-- {
				fmt.Fprintf(pFile, "%d", 0)
			}
			fmt.Fprintf(pFile, " ;\n")
			fmt.Fprintf(pFile, "             ")
			fmt.Fprintf(pFile, "endcase\n")
			fmt.Fprintf(pFile, "           ")
			fmt.Fprintf(pFile, "end\n")
			continue

		} else if Wln_ObjType(p, iObj) == ABC_OPER_SEL_DEC {
			var nRange int32 = Wln_ObjRange(p, Wln_ObjFanin0(p, iObj))
			assert((1<<nRange) == Wln_ObjRange(p, iObj), "(1 << nRange) == Wln_ObjRange(p, iObj)")
			fmt.Fprintf(pFile, "%s ;\n", Wln_ObjName(p, iObj))
			for k = 0; k < Wln_ObjRange(p, iObj); k++ {
				fmt.Fprintf(pFile, "         ")
				fmt.Fprintf(pFile, "wire ")
				fmt.Fprintf(pFile, "%s_", Wln_ObjName(p, iObj))
				for j = 0; j < nRange; j++ {
					fmt.Fprintf(pFile, "%d", (k>>(nRange-1-j))&1)
				}
				fmt.Fprintf(pFile, " = ")
				for j = 0; j < nRange; j++ {
					fmt.Fprintf(pFile, "%s%s%s[%d]", bitIndexPrefix(j), bitPolarityPrefix((k>>(nRange-1-j))&1), Wln_ObjName(p, Wln_ObjFanin(p, iObj, 0)), nRange-1-j)
				}
				fmt.Fprintf(pFile, " ;\n")
			}
			fmt.Fprintf(pFile, "         ")
			fmt.Fprintf(pFile, "assign %s = { ", Wln_ObjName(p, iObj))
			for k = Wln_ObjRange(p, iObj) - 1; k >= 0; k-- {
				fmt.Fprintf(pFile, "%s%s_", commaIfMore(k, Wln_ObjRange(p, iObj)), Wln_ObjName(p, iObj))
				for j = 0; j < nRange; j++ {
					fmt.Fprintf(pFile, "%d", (k>>(nRange-1-j))&1)
				}
			}
			fmt.Fprintf(pFile, " } ;\n")
			continue

		} else if Wln_ObjType(p, iObj) == ABC_OPER_ARI_ADDSUB {
			// out = mode ? a+b+cin : a-b-cin
			fmt.Fprintf(pFile, "%s ;\n", Wln_ObjName(p, iObj))
			fmt.Fprintf(pFile, "         ")
			fmt.Fprintf(pFile, "assign ")
			fmt.Fprintf(pFile, "%s = %s ? %s + %s + %s : %s - %s - %s ;\n", Wln_ObjName(p, iObj), Wln_ObjName(p, Wln_ObjFanin0(p, iObj)), Wln_ObjName(p, Wln_ObjFanin2(p, iObj)), Wln_ObjName(p, Wln_ObjFanin(p, iObj, 3)), Wln_ObjName(p, Wln_ObjFanin1(p, iObj)), Wln_ObjName(p, Wln_ObjFanin2(p, iObj)), Wln_ObjName(p, Wln_ObjFanin(p, iObj, 3)), Wln_ObjName(p, Wln_ObjFanin1(p, iObj)))
			continue

		} else if Wln_ObjType(p, iObj) == ABC_OPER_RAMR || Wln_ObjType(p, iObj) == ABC_OPER_RAMW {
			if true {
				fmt.Fprintf(pFile, "%s ;\n", Wln_ObjName(p, iObj))
				fmt.Fprintf(pFile, "         ")
				fmt.Fprintf(pFile, "%s (", ramModuleName(Wln_ObjType(p, iObj), "ABC_READ", "ABC_WRITE"))
				// inlining of macro Wln_ObjForEachFanin
				for k = 0; k < Wln_ObjFaninNum(p, iObj); k++ {
					iFanin = Wln_ObjFanin(p, iObj, k)
					if !(iFanin != 0) {
					} else {
						fmt.Fprintf(pFile, " .%s(%s),", ramPortName(k, "mem_in", "addr", "data"), Wln_ObjName(p, iFanin))
					}
				}
				fmt.Fprintf(pFile, " .%s(%s) ) ;\n", ramModuleName(Wln_ObjType(p, iObj), "data", "mem_out"), Wln_ObjName(p, iObj))
				continue

			} else {
				var nBitsMem int32 = Wln_ObjRange(p, Wln_ObjFanin(p, iObj, 0))
				//int nBitsAddr = Wln_ObjRange(p,  Wln_ObjFanin(p, iObj, 1) );
				var nBitsDat int32
				if Wln_ObjType(p, iObj) == ABC_OPER_RAMR {
					nBitsDat = Wln_ObjRange(p, iObj)
				} else {
					nBitsDat = Wln_ObjRange(p, Wln_ObjFanin(p, iObj, 2))
				}
				var Depth int32 = nBitsMem / nBitsDat
				assert(nBitsMem%nBitsDat == 0, "nBitsMem % nBitsDat == 0")
				fmt.Fprintf(pFile, "%s ;\n", Wln_ObjName(p, iObj))
				fmt.Fprintf(pFile, "         ")
				fmt.Fprintf(pFile, "%s_%d (", ramModuleName(Wln_ObjType(p, iObj), "CPL_MEM_READ", "CPL_MEM_WRITE"), Depth)
				// inlining of macro Wln_ObjForEachFanin
				for k = 0; k < Wln_ObjFaninNum(p, iObj); k++ {
					iFanin = Wln_ObjFanin(p, iObj, k)
					if !(iFanin != 0) {
					} else {
						fmt.Fprintf(pFile, " .%s(%s),", ramPortName(k, "mem_data_in", "addr_in", "data_in"), Wln_ObjName(p, iFanin))
					}
				}
				fmt.Fprintf(pFile, " .%s(%s) ) ;\n", "data_out", Wln_ObjName(p, iObj))
				continue
			}

		} else if Wln_ObjType(p, iObj) == ABC_OPER_DFFRSE {
			fmt.Fprintf(pFile, "%s ;\n", Wln_ObjName(p, iObj))
			continue

		} else {
			fmt.Fprintf(pFile, "%-16s = ", Wln_ObjName(p, iObj))
			if Wln_ObjType(p, iObj) == ABC_OPER_BIT_BUF {
				fmt.Fprintf(pFile, "%s", Wln_ObjName(p, Wln_ObjFanin0(p, iObj)))
			} else if Wln_ObjType(p, iObj) == ABC_OPER_BIT_MUX {
				fmt.Fprintf(pFile, "%s ? ", Wln_ObjName(p, Wln_ObjFanin0(p, iObj)))
				fmt.Fprintf(pFile, "%s : ", Wln_ObjName(p, Wln_ObjFanin1(p, iObj)))
				fmt.Fprintf(pFile, "%s", Wln_ObjName(p, Wln_ObjFanin2(p, iObj)))

			} else if Wln_ObjType(p, iObj) == ABC_OPER_ARI_MIN {
				fmt.Fprintf(pFile, "-%s", Wln_ObjName(p, Wln_ObjFanin0(p, iObj)))
			} else if Wln_ObjType(p, iObj) == ABC_OPER_BIT_INV {
				fmt.Fprintf(pFile, "~%s", Wln_ObjName(p, Wln_ObjFanin0(p, iObj)))
			} else if Wln_ObjType(p, iObj) == ABC_OPER_LOGIC_NOT {
				fmt.Fprintf(pFile, "!%s", Wln_ObjName(p, Wln_ObjFanin0(p, iObj)))
			} else if Wln_ObjType(p, iObj) == ABC_OPER_RED_AND {
				fmt.Fprintf(pFile, "&%s", Wln_ObjName(p, Wln_ObjFanin0(p, iObj)))
			} else if Wln_ObjType(p, iObj) == ABC_OPER_RED_OR {
				fmt.Fprintf(pFile, "|%s", Wln_ObjName(p, Wln_ObjFanin0(p, iObj)))
			} else if Wln_ObjType(p, iObj) == ABC_OPER_RED_XOR {
				fmt.Fprintf(pFile, "^%s", Wln_ObjName(p, Wln_ObjFanin0(p, iObj)))
			} else if Wln_ObjType(p, iObj) == ABC_OPER_RED_NAND {
				fmt.Fprintf(pFile, "~&%s", Wln_ObjName(p, Wln_ObjFanin0(p, iObj)))
			} else if Wln_ObjType(p, iObj) == ABC_OPER_RED_NOR {
				fmt.Fprintf(pFile, "~|%s", Wln_ObjName(p, Wln_ObjFanin0(p, iObj)))
			} else if Wln_ObjType(p, iObj) == ABC_OPER_RED_NXOR {
				fmt.Fprintf(pFile, "~^%s", Wln_ObjName(p, Wln_ObjFanin0(p, iObj)))
			} else if Wln_ObjType(p, iObj) == ABC_OPER_SLICE {
				fmt.Fprintf(pFile, "%s [%d:%d]", Wln_ObjName(p, Wln_ObjFanin0(p, iObj)), Wln_ObjRangeEnd(p, iObj), Wln_ObjRangeBeg(p, iObj))
			} else if Wln_ObjType(p, iObj) == ABC_OPER_SIGNEXT {
				fmt.Fprintf(pFile, "{ {%d{%s[%d]}}, %s }", Wln_ObjRange(p, iObj)-Wln_ObjRange(p, Wln_ObjFanin0(p, iObj)), Wln_ObjName(p, Wln_ObjFanin0(p, iObj)), Wln_ObjRange(p, Wln_ObjFanin0(p, iObj))-1, Wln_ObjName(p, Wln_ObjFanin0(p, iObj)))
			} else if Wln_ObjType(p, iObj) == ABC_OPER_ZEROPAD {
				fmt.Fprintf(pFile, "{ {%d{1'b0}}, %s }", Wln_ObjRange(p, iObj)-Wln_ObjRange(p, Wln_ObjFanin0(p, iObj)), Wln_ObjName(p, Wln_ObjFanin0(p, iObj)))
			} else if Wln_ObjType(p, iObj) == ABC_OPER_CONCAT {
				fmt.Fprintf(pFile, "{")
				// inlining of macro Wln_ObjForEachFanin
				for k = 0; k < Wln_ObjFaninNum(p, iObj); k++ {
					iFanin = Wln_ObjFanin(p, iObj, k)
					if !(iFanin != 0) {
					} else {
						fmt.Fprintf(pFile, " %s%s", Wln_ObjName(p, Wln_ObjFanin(p, iObj, k)), commaIfLast(k, Wln_ObjFaninNum(p, iObj)))
					}
				}
				fmt.Fprintf(pFile, " }")

			} else {
				fmt.Fprintf(pFile, "%s ", Wln_ObjName(p, Wln_ObjFanin(p, iObj, 0)))
				if Wln_ObjType(p, iObj) == ABC_OPER_SHIFT_R {
					fmt.Fprintf(pFile, ">>")
				} else if Wln_ObjType(p, iObj) == ABC_OPER_SHIFT_RA {
					fmt.Fprintf(pFile, ">>>")
				} else if Wln_ObjType(p, iObj) == ABC_OPER_SHIFT_L {
					fmt.Fprintf(pFile, "<<")
				} else if Wln_ObjType(p, iObj) == ABC_OPER_SHIFT_LA {
					fmt.Fprintf(pFile, "<<<")
				} else if Wln_ObjType(p, iObj) == ABC_OPER_BIT_AND {
					fmt.Fprintf(pFile, "&")
				} else if Wln_ObjType(p, iObj) == ABC_OPER_BIT_OR {
					fmt.Fprintf(pFile, "|")
				} else if Wln_ObjType(p, iObj) == ABC_OPER_BIT_XOR {
					fmt.Fprintf(pFile, "^")
				} else if Wln_ObjType(p, iObj) == ABC_OPER_BIT_NAND {
					fmt.Fprintf(pFile, "~&")
				} else if Wln_ObjType(p, iObj) == ABC_OPER_BIT_NOR {
					fmt.Fprintf(pFile, "~|")
				} else if Wln_ObjType(p, iObj) == ABC_OPER_BIT_NXOR {
					fmt.Fprintf(pFile, "~^")
				} else if Wln_ObjType(p, iObj) == ABC_OPER_LOGIC_IMPL {
					fmt.Fprintf(pFile, "=>")
				} else if Wln_ObjType(p, iObj) == ABC_OPER_LOGIC_AND {
					fmt.Fprintf(pFile, "&&")
				} else if Wln_ObjType(p, iObj) == ABC_OPER_LOGIC_OR {
					fmt.Fprintf(pFile, "||")
				} else if Wln_ObjType(p, iObj) == ABC_OPER_LOGIC_XOR {
					fmt.Fprintf(pFile, "^^")
				} else if Wln_ObjType(p, iObj) == ABC_OPER_COMP_EQU {
					fmt.Fprintf(pFile, "==")
				} else if Wln_ObjType(p, iObj) == ABC_OPER_COMP_NOTEQU {
					fmt.Fprintf(pFile, "!=")
				} else if Wln_ObjType(p, iObj) == ABC_OPER_COMP_LESS {
					fmt.Fprintf(pFile, "<")
				} else if Wln_ObjType(p, iObj) == ABC_OPER_COMP_MORE {
					fmt.Fprintf(pFile, ">")
				} else if Wln_ObjType(p, iObj) == ABC_OPER_COMP_LESSEQU {
					fmt.Fprintf(pFile, "<=")
				} else if Wln_ObjType(p, iObj) == ABC_OPER_COMP_MOREEQU {
					fmt.Fprintf(pFile, ">=")
				} else if Wln_ObjType(p, iObj) == ABC_OPER_ARI_ADD {
					fmt.Fprintf(pFile, "+")
				} else if Wln_ObjType(p, iObj) == ABC_OPER_ARI_SUB {
					fmt.Fprintf(pFile, "-")
				} else if Wln_ObjType(p, iObj) == ABC_OPER_ARI_MUL {
					fmt.Fprintf(pFile, "*")
				} else if Wln_ObjType(p, iObj) == ABC_OPER_ARI_DIV {
					fmt.Fprintf(pFile, "/")
				} else if Wln_ObjType(p, iObj) == ABC_OPER_ARI_REM {
					fmt.Fprintf(pFile, "%%")
				} else if Wln_ObjType(p, iObj) == ABC_OPER_ARI_MOD {
					fmt.Fprintf(pFile, "%%")
				} else if Wln_ObjType(p, iObj) == ABC_OPER_ARI_POW {
					fmt.Fprintf(pFile, "**")
				} else if Wln_ObjType(p, iObj) == ABC_OPER_ARI_SQRT {
					fmt.Fprintf(pFile, "@")
				} else if Wln_ObjType(p, iObj) == ABC_OPER_ARI_SQUARE {
					fmt.Fprintf(pFile, "#")
				} else {
					//assert( 0 );
					fmt.Printf("Failed to write node \"%s\" with unknown operator type (%d).\n", Wln_ObjName(p, iObj), Wln_ObjType(p, iObj))
					fmt.Fprintf(pFile, "???\n")
					continue
				}
				fmt.Fprintf(pFile, " %s", Wln_ObjName(p, Wln_ObjFanin(p, iObj, 1)))
				if Wln_ObjFaninNum(p, iObj) == 3 && Wln_ObjType(p, iObj) == ABC_OPER_ARI_ADD {
					fmt.Fprintf(pFile, " + %s", Wln_ObjName(p, Wln_ObjFanin(p, iObj, 2)))
				}
			}
		}
		fmt.Fprintf(pFile, " ;%s\n", smtSignedComment(p, iObj))
	}
	iFanin = 0
	// write DFFs in the end
	fmt.Fprintf(pFile, "\n")
	// inlining of macro Wln_NtkForEachFf
	for j = 0; j < Wln_NtkFfNum(p); j++ {
		iObj = Wln_NtkFf(p, j)
		var pInNames [8]string = [8]string{"d", "clk", "reset", "set", "enable", "async", "sre", "init"}
		fmt.Fprintf(pFile, "         ")
		fmt.Fprintf(pFile, "%s (", "ABC_DFFRSE")
		// inlining of macro Wln_ObjForEachFanin
		for k = 0; k < Wln_ObjFaninNum(p, iObj); k++ {
			iFanin = Wln_ObjFanin(p, iObj, k)
			if !(iFanin != 0) {
			} else if iFanin != 0 {
				fmt.Fprintf(pFile, " .%s(%s),", pInNames[k], Wln_ObjName(p, iFanin))
			}
		}
		fmt.Fprintf(pFile, " .%s(%s) ) ;\n", "q", Wln_ObjName(p, iObj))
	}
	fmt.Fprintf(pFile, "\n")
	fmt.Fprintf(pFile, "endmodule\n\n")
}
func Wln_WriteVer(p *Wln_Ntk_t, pFileName string) {
	pFile, err := os.OpenFile(pFileName, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, 0666)
	if err != nil {
		fmt.Fprintf(os.Stdout, "Wln_WriteVer(): Cannot open the output file \"%s\".\n", pFileName)
		return
	}
	fmt.Fprintf(pFile, "// Benchmark \"%s\" from file \"%s\" written by ABC on %s\n", p.pName, specOrUnknown(p.pSpec), extra.Extra_TimeStamp())
	fmt.Fprintf(pFile, "\n")
	Wln_WriteTables(pFile, p)
	//    if ( fAddCos )
	//        Wln_WriteAddPos( p );
	Wln_WriteVerInt(pFile, p)
	fmt.Fprintf(pFile, "\n")
	_ = (pFile).Close()
}

////////////////////////////////////////////////////////////////////////
///                       END OF FILE                                ///
////////////////////////////////////////////////////////////////////////

//ABC_NAMESPACE_IMPL_END
