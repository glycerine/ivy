package wlc

import "fmt"

import "os"

func boolString(cond bool, yes, no string) string {
	if cond {
		return yes
	}
	return no
}
func signedSuffix(obj *Wlc_Obj_t) string {
	if Wlc_ObjIsSigned(obj) != 0 {
		return "s"
	}
	return ""
}
func commaUnlessLast(i, n int32) string {
	if i == n-1 {
		return ""
	}
	return ","
}
func faninPort(k int32, first, second, rest string) string {
	if k == 0 {
		return first
	}
	if k == 1 {
		return second
	}
	return rest
}

/**CFile****************************************************************

  FileName    [wlcWriteVer.c]

  SystemName  [ABC: Logic synthesis and verification system.]

  PackageName [Verilog parser.]

  Synopsis    [Writes word-level Verilog.]

  Author      [Alan Mishchenko]

  Affiliation [UC Berkeley]

  Date        [Ver. 1.0. Started - August 22, 2014.]

  Revision    [$Id: wlcWriteVer.c,v 1.00 2014/09/12 00:00:00 alanmi Exp $]

***********************************************************************/

//#include "wlc.h"

//ABC_NAMESPACE_IMPL_START

////////////////////////////////////////////////////////////////////////
///                        DECLARATIONS                              ///
////////////////////////////////////////////////////////////////////////

////////////////////////////////////////////////////////////////////////
///                     FUNCTION DEFINITIONS                         ///
////////////////////////////////////////////////////////////////////////

/**Function*************************************************************

  Synopsis    []

  Description []

  SideEffects []

  SeeAlso     []

***********************************************************************/

func Wlc_WriteTables(pFile *os.File, p *Wlc_Ntk_t) {
	var vNodes *Vec_Int_t
	var pObj, pFanin *Wlc_Obj_t
	var pTable *word
	var i int32
	if p.vTables == nil || Vec_PtrSize(p.vTables) == 0 {
		return
	}
	// map tables into their nodes
	vNodes = Vec_IntStart(Vec_PtrSize(p.vTables))
	// inlining of macro Wlc_NtkForEachObj
	for i = 1; i < Wlc_NtkObjNumMax(p); i++ {
		pObj = Wlc_NtkObj(p, i)
		if pObj.Type == WLC_OBJ_TABLE {
			Vec_IntWriteEntry(vNodes, Wlc_ObjTableId(pObj), i)
		}
	}
	// write tables
	for i = 0; i < Vec_PtrSize(p.vTables); i++ {
		pTable = Vec_PtrEntry(p.vTables, i)
		pObj = Wlc_NtkObj(p, Vec_IntEntry(vNodes, i))
		assert(pObj.Type == WLC_OBJ_TABLE, "pObj.Type == WLC_OBJ_TABLE")
		pFanin = Wlc_ObjFanin0(p, pObj)
		Wlc_WriteTableOne(pFile, Wlc_ObjRange(pFanin), Wlc_ObjRange(pObj), pTable, i)
	}
	Vec_IntFree(vNodes)
}

/*
*Function*************************************************************

	Synopsis    [This was used to add POs to each node except PIs and MUXes.]

	Description []

	SideEffects []

	SeeAlso     []

**********************************************************************
*/
func Wlc_WriteAddPos(p *Wlc_Ntk_t) {
	var pObj *Wlc_Obj_t
	var i int32
	Vec_IntClear(&p.vPos)
	// inlining of macro Wlc_NtkForEachObj
	for i = 1; i < Wlc_NtkObjNumMax(p); i++ {
		pObj = Wlc_NtkObj(p, i)
		if pObj.Type != WLC_OBJ_PI && pObj.Type != WLC_OBJ_MUX {
			pObj.fIsPo = 1
			Vec_IntPush(&p.vPos, Wlc_ObjId(p, pObj))
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
func Wlc_WriteVerIntVec(pFile *os.File, p *Wlc_Ntk_t, vVec *Vec_Int_t, Start int32) {
	var pName string
	var LineLength int32 = Start
	var NameCounter int32
	var AddedLength, i, iObj int32
	for i = 0; i < Vec_IntSize(vVec); i++ {
		iObj = Vec_IntEntry(vVec, i)
		pName = Wlc_ObjName(p, iObj)
		// get the line length after this name is written
		AddedLength = len(pName) + 2
		if NameCounter != 0 && LineLength+AddedLength+3 > 70 { // write the line extender
			fmt.Fprintf(pFile, "\n   ")
			// reset the line length
			LineLength = Start
			NameCounter = 0
		}
		fmt.Fprintf(pFile, " %s%s", pName, commaUnlessLast(i, Vec_IntSize(vVec)))
		LineLength += AddedLength
		NameCounter++
	}
}

func Wlc_ObjFaninBitNum(p *Wlc_Ntk_t, pObj *Wlc_Obj_t) int32 {
	var pFanin *Wlc_Obj_t
	var i, Count int32
	// inlining of macro Wlc_ObjForEachFaninObj
	for i = 0; i < Wlc_ObjFaninNum(pObj); i++ {
		pFanin = Wlc_NtkObj(p, Wlc_ObjFaninId(pObj, i))
		Count += Wlc_ObjRange(pFanin)
	}
	return Count
}

func Wlc_WriteVerInt(pFile *os.File, p *Wlc_Ntk_t, fNoFlops int32) {
	var pObj *Wlc_Obj_t
	var i, k, j, iFanin int32
	var Range string // was [100]byte in C

	fmt.Fprintf(pFile, "module %s ( ", p.pName)
	fmt.Fprintf(pFile, "\n   ")
	if Wlc_NtkPiNum(p) > 0 || ((fNoFlops != 0) && Wlc_NtkCiNum(p)) {
		if fNoFlops != 0 {
			Wlc_WriteVerIntVec(pFile, p, &p.vCis, 3)
		} else {
			Wlc_WriteVerIntVec(pFile, p, &p.vPis, 3)
		}
		fmt.Fprintf(pFile, ",\n   ")
	}
	if Wlc_NtkPoNum(p) > 0 || ((fNoFlops != 0) && Wlc_NtkCoNum(p)) {
		if fNoFlops != 0 {
			Wlc_WriteVerIntVec(pFile, p, &p.vCos, 3)
		} else {
			Wlc_WriteVerIntVec(pFile, p, &p.vPos, 3)
		}
	}
	fmt.Fprintf(pFile, "  );\n")
	// mark fanins of rotation shifts
	// inlining of macro Wlc_NtkForEachObj
	for i = 1; i < Wlc_NtkObjNumMax(p); i++ {
		pObj = Wlc_NtkObj(p, i)
		if pObj.Type == WLC_OBJ_ROTATE_R || pObj.Type == WLC_OBJ_ROTATE_L {
			Wlc_ObjFanin1(p, pObj).Mark = 1
		}
	}
	// inlining of macro Wlc_NtkForEachObj
	for i = 1; i < Wlc_NtkObjNumMax(p); i++ {
		pObj = Wlc_NtkObj(p, i)
		var nDigits int32 = Abc_Base10Log(Abc_AbsInt(pObj.End)+1) + Abc_Base10Log(Abc_AbsInt(pObj.Beg)+1) + int32((pObj.End < 0)) + int32((pObj.Beg < 0))
		if pObj.Mark != 0 {
			pObj.Mark = 0
			continue
		}
		rangePrefix := "       "
		if p.fSmtLib == 0 && Wlc_ObjIsSigned(pObj) != 0 {
			rangePrefix = "signed "
		}
		Range = fmt.Sprintf("%s[%d:%d]%*s", rangePrefix, pObj.End, pObj.Beg, 8-nDigits, "")
		fmt.Fprintf(pFile, "  ")
		if pObj.Type == WLC_OBJ_PI || ((fNoFlops != 0) && pObj.Type == WLC_OBJ_FO) {
			fmt.Fprintf(pFile, "input  ")
		} else if pObj.fIsPo != 0 || ((fNoFlops != 0) && pObj.fIsFi != 0) {
			fmt.Fprintf(pFile, "output ")
		} else {
			fmt.Fprintf(pFile, "       ")
		}
		if Wlc_ObjIsCi(pObj) != 0 || pObj.fIsPo != 0 || ((fNoFlops != 0) && pObj.fIsFi != 0) {
			fmt.Fprintf(pFile, "wire %s %s ;\n", Range, Wlc_ObjName(p, i))
			if Wlc_ObjIsCi(pObj) != 0 {
				continue
			}
			Range = ""
		}
		if (pObj.fIsPo != 0 || ((fNoFlops != 0) && pObj.fIsFi != 0)) && pObj.Type != WLC_OBJ_FF {
			if Wlc_ObjFaninNum(pObj) == 0 {
				continue
			}
			fmt.Fprintf(pFile, "  assign                         ")

		} else if (pObj.Type == WLC_OBJ_MUX && Wlc_ObjFaninNum(pObj) > 3) || pObj.Type == WLC_OBJ_SEL {
			fmt.Fprintf(pFile, "reg  %s ", Range)
		} else {
			fmt.Fprintf(pFile, "wire %s ", Range)
		}
		if pObj.Type == WLC_OBJ_TABLE {
			// wire [3:0] s4972; table0 s4972_Index(s4971, s4972);
			fmt.Fprintf(pFile, "%s ;              table%d", Wlc_ObjName(p, i), Wlc_ObjTableId(pObj))
			fmt.Fprintf(pFile, " s%d_Index(%s, ", i, Wlc_ObjName(p, Wlc_ObjFaninId0(pObj)))
			fmt.Fprintf(pFile, "%s)", Wlc_ObjName(p, i))

		} else if pObj.Type == WLC_OBJ_LUT {
			// wire [3:0] s4972; LUT lut4972_Index(s4971, s4972);
			fmt.Fprintf(pFile, "%s ;           LUT", Wlc_ObjName(p, i))
			fmt.Fprintf(pFile, " lut%d (%s, ", i, Wlc_ObjName(p, Wlc_ObjFaninId0(pObj)))
			for k = 1; k < Wlc_ObjFaninNum(pObj); k++ {
				fmt.Fprintf(pFile, "%s, ", Wlc_ObjName(p, Wlc_ObjFaninId(pObj, k)))
			}
			fmt.Fprintf(pFile, "%s)", Wlc_ObjName(p, i))
			if p.vLutTruths != nil {
				var Truth word = Vec_WrdEntry(p.vLutTruths, Wlc_ObjId(p, pObj))
				fmt.Fprintf(pFile, " ; // TT = ")
				Extra_PrintHex(pFile, &Truth, Wlc_ObjFaninBitNum(p, pObj))
			}

		} else if pObj.Type == WLC_OBJ_CONST {
			fmt.Fprintf(pFile, "%-16s = %d'%sh", Wlc_ObjName(p, i), Wlc_ObjRange(pObj), signedSuffix(pObj))
			if pObj.fXConst != 0 {
				for k = 0; k < (Wlc_ObjRange(pObj)+3)/4; k++ {
					fmt.Fprintf(pFile, "x")
				}

			} else {
				Abc_TtPrintHexArrayRev(pFile, Wlc_ObjConstValue(pObj), (Wlc_ObjRange(pObj)+3)/4)
			}

		} else if pObj.Type == WLC_OBJ_ROTATE_R || pObj.Type == WLC_OBJ_ROTATE_L {
			//  wire [27:0] s4960 = (s57 >> 17) | (s57 << 11);
			var pShift *Wlc_Obj_t = Wlc_ObjFanin1(p, pObj)
			var Num0 int32 = *Wlc_ObjConstValue(pShift)
			var Num1 int32 = Wlc_ObjRange(pObj) - Num0
			assert(pShift.Type == WLC_OBJ_CONST, "pShift.Type == WLC_OBJ_CONST")
			assert(Num0 > 0 && Num0 < Wlc_ObjRange(pObj), "Num0 > 0 && Num0 < Wlc_ObjRange(pObj)")
			fmt.Fprintf(pFile, "%-16s = ", Wlc_ObjName(p, i))
			if pObj.Type == WLC_OBJ_ROTATE_R {
				fmt.Fprintf(pFile, "(%s >> %d) | (%s << %d)", Wlc_ObjName(p, Wlc_ObjFaninId0(pObj)), Num0, Wlc_ObjName(p, Wlc_ObjFaninId0(pObj)), Num1)
			} else {
				fmt.Fprintf(pFile, "(%s << %d) | (%s >> %d)", Wlc_ObjName(p, Wlc_ObjFaninId0(pObj)), Num0, Wlc_ObjName(p, Wlc_ObjFaninId0(pObj)), Num1)
			}

		} else if pObj.Type == WLC_OBJ_MUX && Wlc_ObjFaninNum(pObj) > 3 {
			fmt.Fprintf(pFile, "%s ;\n", Wlc_ObjName(p, i))
			fmt.Fprintf(pFile, "         ")
			fmt.Fprintf(pFile, "always @( ")
			// inlining of macro Wlc_ObjForEachFanin
			for k = 0; k < Wlc_ObjFaninNum(pObj); k++ {
				iFanin = Wlc_ObjFaninId(pObj, k)
				fmt.Fprintf(pFile, "%s%s", boolString(k != 0, " or ", ""), Wlc_ObjName(p, Wlc_ObjFaninId(pObj, k)))
			}
			fmt.Fprintf(pFile, " )\n")
			fmt.Fprintf(pFile, "           ")
			fmt.Fprintf(pFile, "begin\n")
			fmt.Fprintf(pFile, "             ")
			fmt.Fprintf(pFile, "case ( %s )\n", Wlc_ObjName(p, Wlc_ObjFaninId(pObj, 0)))
			// inlining of macro Wlc_ObjForEachFanin
			for k = 0; k < Wlc_ObjFaninNum(pObj); k++ {
				iFanin = Wlc_ObjFaninId(pObj, k)
				if !k {
					continue
				}
				fmt.Fprintf(pFile, "               ")
				fmt.Fprintf(pFile, "%d : %s = ", k-1, Wlc_ObjName(p, i))
				fmt.Fprintf(pFile, "%s ;\n", Wlc_ObjName(p, Wlc_ObjFaninId(pObj, k)))
			}
			fmt.Fprintf(pFile, "             ")
			fmt.Fprintf(pFile, "endcase\n")
			fmt.Fprintf(pFile, "           ")
			fmt.Fprintf(pFile, "end\n")
			continue

		} else if pObj.Type == WLC_OBJ_SEL {
			fmt.Fprintf(pFile, "%s ;\n", Wlc_ObjName(p, i))
			fmt.Fprintf(pFile, "         ")
			fmt.Fprintf(pFile, "always @( ")
			// inlining of macro Wlc_ObjForEachFanin
			for k = 0; k < Wlc_ObjFaninNum(pObj); k++ {
				iFanin = Wlc_ObjFaninId(pObj, k)
				fmt.Fprintf(pFile, "%s%s", boolString(k != 0, " or ", ""), Wlc_ObjName(p, Wlc_ObjFaninId(pObj, k)))
			}
			fmt.Fprintf(pFile, " )\n")
			fmt.Fprintf(pFile, "           ")
			fmt.Fprintf(pFile, "begin\n")
			fmt.Fprintf(pFile, "             ")
			fmt.Fprintf(pFile, "case ( %s )\n", Wlc_ObjName(p, Wlc_ObjFaninId(pObj, 0)))
			// inlining of macro Wlc_ObjForEachFanin
			for k = 0; k < Wlc_ObjFaninNum(pObj); k++ {
				iFanin = Wlc_ObjFaninId(pObj, k)
				if !k {
					continue
				}
				fmt.Fprintf(pFile, "               ")
				fmt.Fprintf(pFile, "%d'b", Wlc_ObjFaninNum(pObj)-1)
				for j = Wlc_ObjFaninNum(pObj) - 1; j > 0; j-- {
					fmt.Fprintf(pFile, "%d", int32((j == k)))
				}
				fmt.Fprintf(pFile, " : %s = ", Wlc_ObjName(p, i))
				fmt.Fprintf(pFile, "%s ;\n", Wlc_ObjName(p, Wlc_ObjFaninId(pObj, k)))
			}
			fmt.Fprintf(pFile, "               ")
			fmt.Fprintf(pFile, "default")
			fmt.Fprintf(pFile, " : %s = ", Wlc_ObjName(p, i))
			fmt.Fprintf(pFile, "%d'b", Wlc_ObjRange(pObj))
			for j = Wlc_ObjRange(pObj) - 1; j >= 0; j-- {
				fmt.Fprintf(pFile, "%d", 0)
			}
			fmt.Fprintf(pFile, " ;\n")
			fmt.Fprintf(pFile, "             ")
			fmt.Fprintf(pFile, "endcase\n")
			fmt.Fprintf(pFile, "           ")
			fmt.Fprintf(pFile, "end\n")
			continue

		} else if pObj.Type == WLC_OBJ_DEC {
			var nRange int32 = Wlc_ObjRange(Wlc_ObjFanin0(p, pObj))
			assert((1<<nRange) == Wlc_ObjRange(pObj), "(1 << nRange) == Wlc_ObjRange(pObj)")
			fmt.Fprintf(pFile, "%s ;\n", Wlc_ObjName(p, i))
			for k = 0; k < Wlc_ObjRange(pObj); k++ {
				fmt.Fprintf(pFile, "         ")
				fmt.Fprintf(pFile, "wire ")
				fmt.Fprintf(pFile, "%s_", Wlc_ObjName(p, i))
				for j = 0; j < nRange; j++ {
					fmt.Fprintf(pFile, "%d", (k>>(nRange-1-j))&1)
				}
				fmt.Fprintf(pFile, " = ")
				for j = 0; j < nRange; j++ {
					fmt.Fprintf(pFile, "%s%s%s[%d]", boolString(j != 0, " & ", ""), boolString(((k>>(nRange-1-j))&1) != 0, " ", "~"), Wlc_ObjName(p, Wlc_ObjFaninId(pObj, 0)), nRange-1-j)
				}
				fmt.Fprintf(pFile, " ;\n")
			}
			fmt.Fprintf(pFile, "         ")
			fmt.Fprintf(pFile, "assign %s = { ", Wlc_ObjName(p, i))
			for k = Wlc_ObjRange(pObj) - 1; k >= 0; k-- {
				fmt.Fprintf(pFile, "%s%s_", boolString(k < Wlc_ObjRange(pObj)-1, ", ", ""), Wlc_ObjName(p, i))
				for j = 0; j < nRange; j++ {
					fmt.Fprintf(pFile, "%d", (k>>(nRange-1-j))&1)
				}
			}
			fmt.Fprintf(pFile, " } ;\n")
			continue

		} else if pObj.Type == WLC_OBJ_ARI_ADDSUB {
			// out = mode ? a+b+cin : a-b-cin
			fmt.Fprintf(pFile, "%s ;\n", Wlc_ObjName(p, i))
			fmt.Fprintf(pFile, "         ")
			fmt.Fprintf(pFile, "assign ")
			fmt.Fprintf(pFile, "%s = %s ? %s + %s + %s : %s - %s - %s ;\n", Wlc_ObjName(p, i), Wlc_ObjName(p, Wlc_ObjFaninId0(pObj)), Wlc_ObjName(p, Wlc_ObjFaninId2(pObj)), Wlc_ObjName(p, Wlc_ObjFaninId(pObj, 3)), Wlc_ObjName(p, Wlc_ObjFaninId1(pObj)), Wlc_ObjName(p, Wlc_ObjFaninId2(pObj)), Wlc_ObjName(p, Wlc_ObjFaninId(pObj, 3)), Wlc_ObjName(p, Wlc_ObjFaninId1(pObj)))
			continue

		} else if pObj.Type == WLC_OBJ_READ || pObj.Type == WLC_OBJ_WRITE {
			if p.fMemPorts != 0 {
				fmt.Fprintf(pFile, "%s ;\n", Wlc_ObjName(p, i))
				fmt.Fprintf(pFile, "         ")
				fmt.Fprintf(pFile, "%s (", boolString(pObj.Type == WLC_OBJ_READ, "ABC_READ", "ABC_WRITE"))
				// inlining of macro Wlc_ObjForEachFanin
				for k = 0; k < Wlc_ObjFaninNum(pObj); k++ {
					iFanin = Wlc_ObjFaninId(pObj, k)
					fmt.Fprintf(pFile, " .%s(%s),", faninPort(k, "mem_in", "addr", "data"), Wlc_ObjName(p, iFanin))
				}
				fmt.Fprintf(pFile, " .%s(%s) ) ;\n", boolString(pObj.Type == WLC_OBJ_READ, "data", "mem_out"), Wlc_ObjName(p, i))
				continue

			} else {
				var nBitsMem int32 = Wlc_ObjRange(Wlc_ObjFanin(p, pObj, 0))
				//int nBitsAddr = Wlc_ObjRange( Wlc_ObjFanin(p, pObj, 1) );
				var nBitsDat int32
				if pObj.Type == WLC_OBJ_READ {
					nBitsDat = Wlc_ObjRange(pObj)
				} else {
					nBitsDat = Wlc_ObjRange(Wlc_ObjFanin(p, pObj, 2))
				}
				var Depth int32 = nBitsMem / nBitsDat
				assert(nBitsMem%nBitsDat == 0, "nBitsMem % nBitsDat == 0")
				fmt.Fprintf(pFile, "%s ;\n", Wlc_ObjName(p, i))
				fmt.Fprintf(pFile, "         ")
				fmt.Fprintf(pFile, "%s_%d (", boolString(pObj.Type == WLC_OBJ_READ, "CPL_MEM_READ", "CPL_MEM_WRITE"), Depth)
				// inlining of macro Wlc_ObjForEachFanin
				for k = 0; k < Wlc_ObjFaninNum(pObj); k++ {
					iFanin = Wlc_ObjFaninId(pObj, k)
					fmt.Fprintf(pFile, " .%s(%s),", faninPort(k, "mem_data_in", "addr_in", "data_in"), Wlc_ObjName(p, iFanin))
				}
				fmt.Fprintf(pFile, " .%s(%s) ) ;\n", "data_out", Wlc_ObjName(p, i))
				continue
			}

		} else if pObj.Type == WLC_OBJ_FF {
			fmt.Fprintf(pFile, "%s ;\n", Wlc_ObjName(p, i))
			continue

		} else {
			fmt.Fprintf(pFile, "%-16s = ", Wlc_ObjName(p, i))
			if pObj.Type == WLC_OBJ_BUF {
				fmt.Fprintf(pFile, "%s", Wlc_ObjName(p, Wlc_ObjFaninId0(pObj)))
			} else if pObj.Type == WLC_OBJ_MUX {
				fmt.Fprintf(pFile, "%s ? ", Wlc_ObjName(p, Wlc_ObjFaninId0(pObj)))
				fmt.Fprintf(pFile, "%s : ", Wlc_ObjName(p, Wlc_ObjFaninId2(pObj)))
				fmt.Fprintf(pFile, "%s", Wlc_ObjName(p, Wlc_ObjFaninId1(pObj)))

			} else if pObj.Type == WLC_OBJ_ARI_MINUS {
				fmt.Fprintf(pFile, "-%s", Wlc_ObjName(p, Wlc_ObjFaninId0(pObj)))
			} else if pObj.Type == WLC_OBJ_BIT_NOT {
				fmt.Fprintf(pFile, "~%s", Wlc_ObjName(p, Wlc_ObjFaninId0(pObj)))
			} else if pObj.Type == WLC_OBJ_LOGIC_NOT {
				fmt.Fprintf(pFile, "!%s", Wlc_ObjName(p, Wlc_ObjFaninId0(pObj)))
			} else if pObj.Type == WLC_OBJ_REDUCT_AND {
				fmt.Fprintf(pFile, "&%s", Wlc_ObjName(p, Wlc_ObjFaninId0(pObj)))
			} else if pObj.Type == WLC_OBJ_REDUCT_OR {
				fmt.Fprintf(pFile, "|%s", Wlc_ObjName(p, Wlc_ObjFaninId0(pObj)))
			} else if pObj.Type == WLC_OBJ_REDUCT_XOR {
				fmt.Fprintf(pFile, "^%s", Wlc_ObjName(p, Wlc_ObjFaninId0(pObj)))
			} else if pObj.Type == WLC_OBJ_REDUCT_NAND {
				fmt.Fprintf(pFile, "~&%s", Wlc_ObjName(p, Wlc_ObjFaninId0(pObj)))
			} else if pObj.Type == WLC_OBJ_REDUCT_NOR {
				fmt.Fprintf(pFile, "~|%s", Wlc_ObjName(p, Wlc_ObjFaninId0(pObj)))
			} else if pObj.Type == WLC_OBJ_REDUCT_NXOR {
				fmt.Fprintf(pFile, "~^%s", Wlc_ObjName(p, Wlc_ObjFaninId0(pObj)))
			} else if pObj.Type == WLC_OBJ_BIT_SELECT {
				fmt.Fprintf(pFile, "%s [%d:%d]", Wlc_ObjName(p, Wlc_ObjFaninId0(pObj)), Wlc_ObjRangeEnd(pObj), Wlc_ObjRangeBeg(pObj))
			} else if pObj.Type == WLC_OBJ_BIT_SIGNEXT {
				fmt.Fprintf(pFile, "{ {%d{%s[%d]}}, %s }", Wlc_ObjRange(pObj)-Wlc_ObjRange(Wlc_ObjFanin0(p, pObj)), Wlc_ObjName(p, Wlc_ObjFaninId0(pObj)), Wlc_ObjRange(Wlc_ObjFanin0(p, pObj))-1, Wlc_ObjName(p, Wlc_ObjFaninId0(pObj)))
			} else if pObj.Type == WLC_OBJ_BIT_ZEROPAD {
				fmt.Fprintf(pFile, "{ {%d{1'b0}}, %s }", Wlc_ObjRange(pObj)-Wlc_ObjRange(Wlc_ObjFanin0(p, pObj)), Wlc_ObjName(p, Wlc_ObjFaninId0(pObj)))
			} else if pObj.Type == WLC_OBJ_BIT_CONCAT {
				fmt.Fprintf(pFile, "{")
				// inlining of macro Wlc_ObjForEachFanin
				for k = 0; k < Wlc_ObjFaninNum(pObj); k++ {
					iFanin = Wlc_ObjFaninId(pObj, k)
					fmt.Fprintf(pFile, " %s%s", Wlc_ObjName(p, Wlc_ObjFaninId(pObj, k)), boolString(k == Wlc_ObjFaninNum(pObj)-1, "", ","))
				}
				fmt.Fprintf(pFile, " }")

			} else {
				fmt.Fprintf(pFile, "%s ", Wlc_ObjName(p, Wlc_ObjFaninId(pObj, 0)))
				if pObj.Type == WLC_OBJ_SHIFT_R {
					fmt.Fprintf(pFile, ">>")
				} else if pObj.Type == WLC_OBJ_SHIFT_RA {
					fmt.Fprintf(pFile, ">>>")
				} else if pObj.Type == WLC_OBJ_SHIFT_L {
					fmt.Fprintf(pFile, "<<")
				} else if pObj.Type == WLC_OBJ_SHIFT_LA {
					fmt.Fprintf(pFile, "<<<")
				} else if pObj.Type == WLC_OBJ_BIT_AND {
					fmt.Fprintf(pFile, "&")
				} else if pObj.Type == WLC_OBJ_BIT_OR {
					fmt.Fprintf(pFile, "|")
				} else if pObj.Type == WLC_OBJ_BIT_XOR {
					fmt.Fprintf(pFile, "^")
				} else if pObj.Type == WLC_OBJ_BIT_NAND {
					fmt.Fprintf(pFile, "~&")
				} else if pObj.Type == WLC_OBJ_BIT_NOR {
					fmt.Fprintf(pFile, "~|")
				} else if pObj.Type == WLC_OBJ_BIT_NXOR {
					fmt.Fprintf(pFile, "~^")
				} else if pObj.Type == WLC_OBJ_LOGIC_IMPL {
					fmt.Fprintf(pFile, "=>")
				} else if pObj.Type == WLC_OBJ_LOGIC_AND {
					fmt.Fprintf(pFile, "&&")
				} else if pObj.Type == WLC_OBJ_LOGIC_OR {
					fmt.Fprintf(pFile, "||")
				} else if pObj.Type == WLC_OBJ_LOGIC_XOR {
					fmt.Fprintf(pFile, "^^")
				} else if pObj.Type == WLC_OBJ_COMP_EQU {
					fmt.Fprintf(pFile, "==")
				} else if pObj.Type == WLC_OBJ_COMP_NOTEQU {
					fmt.Fprintf(pFile, "!=")
				} else if pObj.Type == WLC_OBJ_COMP_LESS {
					fmt.Fprintf(pFile, "<")
				} else if pObj.Type == WLC_OBJ_COMP_MORE {
					fmt.Fprintf(pFile, ">")
				} else if pObj.Type == WLC_OBJ_COMP_LESSEQU {
					fmt.Fprintf(pFile, "<=")
				} else if pObj.Type == WLC_OBJ_COMP_MOREEQU {
					fmt.Fprintf(pFile, ">=")
				} else if pObj.Type == WLC_OBJ_ARI_ADD {
					fmt.Fprintf(pFile, "+")
				} else if pObj.Type == WLC_OBJ_ARI_SUB {
					fmt.Fprintf(pFile, "-")
				} else if pObj.Type == WLC_OBJ_ARI_MULTI {
					fmt.Fprintf(pFile, "*")
				} else if pObj.Type == WLC_OBJ_ARI_DIVIDE {
					fmt.Fprintf(pFile, "/")
				} else if pObj.Type == WLC_OBJ_ARI_REM {
					fmt.Fprintf(pFile, "%%")
				} else if pObj.Type == WLC_OBJ_ARI_MODULUS {
					fmt.Fprintf(pFile, "%%")
				} else if pObj.Type == WLC_OBJ_ARI_POWER {
					fmt.Fprintf(pFile, "**")
				} else if pObj.Type == WLC_OBJ_ARI_SQRT {
					fmt.Fprintf(pFile, "@")
				} else if pObj.Type == WLC_OBJ_ARI_SQUARE {
					fmt.Fprintf(pFile, "#")
				} else {
					//assert( 0 );
					fmt.Printf("Failed to write node \"%s\" with unknown operator type (%d).\n", Wlc_ObjName(p, i), pObj.Type)
					fmt.Fprintf(pFile, "???\n")
					continue
				}
				fmt.Fprintf(pFile, " %s", Wlc_ObjName(p, Wlc_ObjFaninId(pObj, 1)))
				if Wlc_ObjFaninNum(pObj) == 3 && pObj.Type == WLC_OBJ_ARI_ADD {
					fmt.Fprintf(pFile, " + %s", Wlc_ObjName(p, Wlc_ObjFaninId(pObj, 2)))
				}
			}
		}
		fmt.Fprintf(pFile, " ;%s\n", boolString(p.fSmtLib != 0 && Wlc_ObjIsSigned(pObj) != 0, " // signed SMT-LIB operator", ""))
	}
	iFanin = 0
	assert(p.vInits == nil || Wlc_NtkFfNum(p) == Vec_IntSize(p.vInits), "p.vInits == nil || Wlc_NtkFfNum(p) == Vec_IntSize(p.vInits)")
	if !(fNoFlops != 0) {
		if p.vInits != nil {
			// inlining of macro Wlc_NtkForEachCi
			for i = 0; i < Wlc_NtkCiNum(p); i++ {
				pObj = Wlc_NtkCi(p, i)
				var nDigits int32 = Abc_Base10Log(pObj.End+1) + 1
				var pName string = Wlc_ObjName(p, Wlc_ObjId(p, pObj))
				assert(i == Wlc_ObjCiId(pObj), "i == Wlc_ObjCiId(pObj)")
				if pObj.Type == WLC_OBJ_PI {
					continue
				}
				Range = fmt.Sprintf("       [%d:%d]%*s", Wlc_ObjRange(pObj)-1, 0, 8-nDigits, "")
				fmt.Fprintf(pFile, "         ")
				fmt.Fprintf(pFile, "wire %s ", Range)
				fmt.Fprintf(pFile, "%s_init%*s = ", pName, 11-int32(len(pName)), "")
				if Vec_IntEntry(p.vInits, i-Wlc_NtkPiNum(p)) > 0 {
					fmt.Fprintf(pFile, "%s", Wlc_ObjName(p, Wlc_ObjId(p, Wlc_NtkPi(p, Vec_IntEntry(p.vInits, i-Wlc_NtkPiNum(p))))))
				} else {
					if p.pInits[iFanin] == 'x' || p.pInits[iFanin] == 'X' {
						fmt.Fprintf(pFile, "%d'h", Wlc_ObjRange(pObj))
						for k = 0; k < (Wlc_ObjRange(pObj)+3)/4; k++ {
							fmt.Fprintf(pFile, "x")
						}

					} else {
						fmt.Fprintf(pFile, "%d'b", Wlc_ObjRange(pObj))
						for k = Wlc_ObjRange(pObj) - 1; k >= 0; k-- {
							fmt.Fprintf(pFile, "%c", p.pInits[iFanin+k])
						}
					}
				}
				fmt.Fprintf(pFile, ";\n")
				iFanin += Wlc_ObjRange(pObj)
			}
		}
		// inlining of macro Wlc_NtkForEachCi
		for i = 0; i < Wlc_NtkCiNum(p); i++ {
			pObj = Wlc_NtkCi(p, i)
			assert(i == Wlc_ObjCiId(pObj), "i == Wlc_ObjCiId(pObj)")
			if pObj.Type == WLC_OBJ_PI {
				continue
			}
			fmt.Fprintf(pFile, "         ")
			if p.fEasyFfs != 0 {
				fmt.Fprintf(pFile, "ABC_DFF")
				fmt.Fprintf(pFile, " reg%d (", i)
				fmt.Fprintf(pFile, " .q(%s),", Wlc_ObjName(p, Wlc_ObjId(p, pObj)))
				fmt.Fprintf(pFile, " .d(%s),", Wlc_ObjName(p, Wlc_ObjId(p, Wlc_ObjFo2Fi(p, pObj))))
				if p.vInits != nil {
					fmt.Fprintf(pFile, " .init(%s_init)", Wlc_ObjName(p, Wlc_ObjId(p, pObj)))
				}
				fmt.Fprintf(pFile, " ) ;\n")

			} else {
				fmt.Fprintf(pFile, "CPL_FF")
				if Wlc_ObjRange(pObj) > 1 {
					fmt.Fprintf(pFile, "#%d%*s", Wlc_ObjRange(pObj), 4-Abc_Base10Log(Wlc_ObjRange(pObj)+1), "")
				} else {
					fmt.Fprintf(pFile, "     ")
				}
				fmt.Fprintf(pFile, " reg%d (", i)
				fmt.Fprintf(pFile, " .q(%s),", Wlc_ObjName(p, Wlc_ObjId(p, pObj)))
				fmt.Fprintf(pFile, " .qbar(),")
				fmt.Fprintf(pFile, " .d(%s),", Wlc_ObjName(p, Wlc_ObjId(p, Wlc_ObjFo2Fi(p, pObj))))
				fmt.Fprintf(pFile, " .clk(%s),", "1'b0")
				fmt.Fprintf(pFile, " .arst(%s),", "1'b0")
				if p.vInits != nil {
					fmt.Fprintf(pFile, " .arstval(%s_init)", Wlc_ObjName(p, Wlc_ObjId(p, pObj)))
				} else {
					fmt.Fprintf(pFile, " .arstval(%s)", "1'b0")
				}
				fmt.Fprintf(pFile, " ) ;\n")
			}
		}
		assert(p.vInits == nil || iFanin == int32(len(p.pInits)), "p.vInits == nil || iFanin == int32(strlen(p.pInits))")
	}
	// write DFFs in the end
	fmt.Fprintf(pFile, "\n")
	// inlining of macro Wlc_NtkForEachFf2
	for i = 0; i < Vec_IntSize(&p.vFfs2); i++ {
		pObj = Wlc_NtkFf2(p, i)
		var pInNames [8]string = [8]string{"d", "clk", "reset", "set", "enable", "async", "sre", "init"}
		fmt.Fprintf(pFile, "         ")
		fmt.Fprintf(pFile, "%s (", "ABC_DFFRSE")
		// inlining of macro Wlc_ObjForEachFanin
		for k = 0; k < Wlc_ObjFaninNum(pObj); k++ {
			iFanin = Wlc_ObjFaninId(pObj, k)
			if iFanin != 0 {
				fmt.Fprintf(pFile, " .%s(%s),", pInNames[k], Wlc_ObjName(p, iFanin))
			}
		}
		fmt.Fprintf(pFile, " .%s(%s) ) ;\n", "q", Wlc_ObjName(p, Wlc_ObjId(p, pObj)))
	}
	fmt.Fprintf(pFile, "\n")
	fmt.Fprintf(pFile, "endmodule\n\n")
}
func Wlc_WriteVer(p *Wlc_Ntk_t, pFileName string, fAddCos int32, fNoFlops int32) {
	var pFile *os.File
	pFile = os.OpenFile(pFileName, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, 0666)
	if pFile == nil {
		fmt.Fprintf(os.Stdout, "Wlc_WriteVer(): Cannot open the output file \"%s\".\n", pFileName)
		return
	}
	fmt.Fprintf(pFile, "// Benchmark \"%s\" from file \"%s\" written by ABC on %s\n", p.pName, boolString(p.pSpec != "", p.pSpec, "unknown"), Extra_TimeStamp())
	fmt.Fprintf(pFile, "\n")
	Wlc_WriteTables(pFile, p)
	if fAddCos != 0 {
		Wlc_WriteAddPos(p)
	}
	Wlc_WriteVerInt(pFile, p, fNoFlops)
	fmt.Fprintf(pFile, "\n")
	_ = (pFile).Close()
}

////////////////////////////////////////////////////////////////////////
///                       END OF FILE                                ///
////////////////////////////////////////////////////////////////////////

//ABC_NAMESPACE_IMPL_END
