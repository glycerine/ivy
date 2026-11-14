package bac

import "fmt"

import "os"
import "github.com/glycerine/abcgo/misc/vec"

/**CFile****************************************************************

  FileName    [bacWriteVer.c]

  SystemName  [ABC: Logic synthesis and verification system.]

  PackageName [Hierarchical word-level netlist.]

  Synopsis    [Verilog writer.]

  Author      [Alan Mishchenko]

  Affiliation [UC Berkeley]

  Date        [Ver. 1.0. Started - November 29, 2014.]

  Revision    [$Id: bacWriteVer.c,v 1.00 2014/11/29 00:00:00 alanmi Exp $]

***********************************************************************/

//#include "bac.h"
//#include "bacPrs.h"
//#include "map/mio/mio.h"
//#include "base/main/main.h"

//ABC_NAMESPACE_IMPL_START

////////////////////////////////////////////////////////////////////////
///                        DECLARATIONS                              ///
////////////////////////////////////////////////////////////////////////

////////////////////////////////////////////////////////////////////////
///                     FUNCTION DEFINITIONS                         ///
////////////////////////////////////////////////////////////////////////

/*
*Function*************************************************************

	Synopsis    [Writing parser state into a file.]

	Description []

	SideEffects []

	SeeAlso     []

**********************************************************************
*/
func Psr_ManWriteVerilogConcat(pFile *os.File, p *Psr_Ntk_t, Con int32) {
	var vSigs *vec.Vec_Int_t = Psr_CatSignals(p, Con)
	fmt.Fprintf(pFile, "{")
	Psr_ManWriteVerilogArray(pFile, p, vSigs, 0, Vec_IntSize(vSigs), 0)
	fmt.Fprintf(pFile, "}")
}
func Psr_ManWriteVerilogSignal(pFile *os.File, p *Psr_Ntk_t, Sig int32) {
	var Value int32 = Abc_Lit2Var2(Sig)
	var Type Psr_ManType_t = Psr_ManType_t(Abc_Lit2Att2(Sig))
	if Type == BAC_PRS_NAME || Type == BAC_PRS_CONST {
		fmt.Fprintf(pFile, "%s", Psr_NtkStr(p, Value))
	} else if Type == BAC_PRS_SLICE {
		fmt.Fprintf(pFile, "%s%s", Psr_NtkStr(p, Psr_SliceName(p, Value)), Psr_NtkStr(p, Psr_SliceRange(p, Value)))
	} else if Type == BAC_PRS_CONCAT {
		Psr_ManWriteVerilogConcat(pFile, p, Value)
	} else {
		assert(false, "0")
	}
}
func Psr_ManWriteVerilogArray(pFile *os.File, p *Psr_Ntk_t, vSigs *vec.Vec_Int_t, Start int32, Stop int32, fOdd int32) {
	var i, Sig int32
	assert(Vec_IntSize(vSigs) > 0, "Vec_IntSize(vSigs) > 0")
	// inlining of macro Vec_IntForEachEntryStartStop
	for i = Start; i < Stop; i++ {
		Sig = Vec_IntEntry(vSigs, i)
		if (fOdd != 0) && (i&1) == 0 {
			continue
		}
		Psr_ManWriteVerilogSignal(pFile, p, Sig)
		if i != Stop-1 {
			fmt.Fprintf(pFile, ", ")
		}
	}
}
func Psr_ManWriteVerilogArray2(pFile *os.File, p *Psr_Ntk_t, vSigs *vec.Vec_Int_t) {
	var i, FormId, ActSig int32
	assert(Vec_IntSize(vSigs)%2 == 0, "Vec_IntSize(vSigs) % 2 == 0")
	// inlining of macro Vec_IntForEachEntryDouble
	for i = 0; i+1 < Vec_IntSize(vSigs); i += 2 {
		FormId, ActSig = Vec_IntEntry(vSigs, i), Vec_IntEntry(vSigs, i+1)
		fmt.Fprintf(pFile, ".")
		fmt.Fprintf(pFile, "%s", Psr_NtkStr(p, FormId))
		fmt.Fprintf(pFile, "(")
		Psr_ManWriteVerilogSignal(pFile, p, ActSig)
		if i == Vec_IntSize(vSigs)-2 {
			fmt.Fprintf(pFile, ")")
		} else {
			fmt.Fprintf(pFile, "), ")
		}
	}
}
func Psr_ManWriteVerilogMux(pFile *os.File, p *Psr_Ntk_t, vSigs *vec.Vec_Int_t) {
	var i, ActSig int32
	var pStrs [4]string = [4]string{" = ", " ? ", " : ", ";\n"}
	assert(Vec_IntSize(vSigs) == 8, "Vec_IntSize(vSigs) == 8")
	fmt.Fprintf(pFile, "  assign ")
	Psr_ManWriteVerilogSignal(pFile, p, Vec_IntEntryLast(vSigs))
	fmt.Fprintf(pFile, "%s", pStrs[0])
	// inlining of macro Vec_IntForEachEntryDouble
	for i = 0; i+1 < Vec_IntSize(vSigs); i += 2 {
		ActSig = Vec_IntEntry(vSigs, i+1)
		Psr_ManWriteVerilogSignal(pFile, p, ActSig)
		fmt.Fprintf(pFile, "%s", pStrs[1+i/2])
		if i == 4 {
			break
		}
	}
}
func Psr_ManWriteVerilogBoxes(pFile *os.File, p *Psr_Ntk_t) {
	var vBox *vec.Vec_Int_t
	var i int32
	// inlining of macro Psr_NtkForEachBox
	for i = 0; i < Psr_NtkBoxNum(p); i++ {
		vBox = Psr_BoxSignals(p, i)
		var NtkId Bac_ObjType_t = Bac_ObjType_t(Psr_BoxNtk(p, i))
		if NtkId == BAC_BOX_MUX {
			Psr_ManWriteVerilogMux(pFile, p, vBox)
		} else if Psr_BoxIsNode(p, i) != 0 { // node   ------- check order of fanins

			fmt.Fprintf(pFile, "  %s (", Ptr_TypeToName(NtkId))
			Psr_ManWriteVerilogSignal(pFile, p, Vec_IntEntryLast(vBox))
			if Psr_BoxIONum(p, i) > 1 {
				fmt.Fprintf(pFile, ", ")
			}
			Psr_ManWriteVerilogArray(pFile, p, vBox, 0, Vec_IntSize(vBox)-2, 1)
			fmt.Fprintf(pFile, ");\n")

		} else { // box

			//char * s = Psr_NtkStr(p, Vec_IntEntry(vBox, 0));
			boxName := ""
			if Psr_BoxName(p, i) != 0 {
				boxName = Psr_NtkStr(p, Psr_BoxName(p, i))
			}
			fmt.Fprintf(pFile, "  %s %s (", Psr_NtkStr(p, int32(NtkId)), boxName)
			Psr_ManWriteVerilogArray2(pFile, p, vBox)
			fmt.Fprintf(pFile, ");\n")
		}
	}
}
func Psr_ManWriteVerilogIos(pFile *os.File, p *Psr_Ntk_t, SigType int32) {
	var NameId, RangeId, i int32
	var pSigNames [4]string = [4]string{"inout", "input", "output", "wire"}
	var vSigs [4]*vec.Vec_Int_t = [4]*vec.Vec_Int_t{&p.vInouts, &p.vInputs, &p.vOutputs, &p.vWires}
	var vSigsR [4]*vec.Vec_Int_t = [4]*vec.Vec_Int_t{&p.vInoutsR, &p.vInputsR, &p.vOutputsR, &p.vWiresR}
	if SigType == 3 {
		fmt.Fprintf(pFile, "\n")
	}
	// inlining of macro Vec_IntForEachEntryTwo
	for i = 0; i < Vec_IntSize(vSigs[SigType]); i++ {
		NameId, RangeId = Vec_IntEntry(vSigs[SigType], i), Vec_IntEntry(vSigsR[SigType], i)
		rangeName := ""
		if RangeId != 0 {
			rangeName = Psr_NtkStr(p, RangeId)
		}
		fmt.Fprintf(pFile, "  %s %s%s;\n", pSigNames[SigType], rangeName, Psr_NtkStr(p, NameId))
	}
}
func Psr_ManWriteVerilogIoOrder(pFile *os.File, p *Psr_Ntk_t, vOrder *vec.Vec_Int_t) {
	var i, NameId int32
	// inlining of macro Vec_IntForEachEntry
	for i = 0; i < Vec_IntSize(vOrder); i++ {
		NameId = Vec_IntEntry(vOrder, i)
		fmt.Fprintf(pFile, "%s", Psr_NtkStr(p, NameId))
		if i != Vec_IntSize(vOrder)-1 {
			fmt.Fprintf(pFile, ", ")
		}
	}
}
func Psr_ManWriteVerilogNtk(pFile *os.File, p *Psr_Ntk_t) {
	var s int32
	// write header
	fmt.Fprintf(pFile, "module %s (\n    ", Psr_NtkStr(p, p.iModuleName))
	Psr_ManWriteVerilogIoOrder(pFile, p, &p.vOrder)
	fmt.Fprintf(pFile, "\n  );\n")
	// write declarations
	for s = 0; s < 4; s++ {
		Psr_ManWriteVerilogIos(pFile, p, s)
	}
	fmt.Fprintf(pFile, "\n")
	// write objects
	Psr_ManWriteVerilogBoxes(pFile, p)
	fmt.Fprintf(pFile, "endmodule\n\n")
}
func Psr_ManWriteVerilog(pFileName string, vPrs *vec.Vec_Ptr_t) {
	var pNtk *Psr_Ntk_t = Psr_ManRoot(vPrs)
	var i int32
	pFile, openErr := os.OpenFile(pFileName, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, 0666)
	if openErr != nil {
		fmt.Printf("Cannot open output file \"%s\".\n", pFileName)
		return
	}
	fmt.Fprintf(pFile, "// Design \"%s\" written by ABC on %s\n\n", Psr_NtkStr(pNtk, pNtk.iModuleName), Extra_TimeStamp())
	// inlining of macro Vec_PtrForEachEntry
	for i = 0; i < Vec_PtrSize(vPrs); i++ {
		pNtk, _ = Vec_PtrEntry(vPrs, i).(*Psr_Ntk_t)
		Psr_ManWriteVerilogNtk(pFile, pNtk)
	}
	_ = (pFile).Close()
}

/**Function*************************************************************

  Synopsis    [Writing word-level Verilog.]

  Description []

  SideEffects []

  SeeAlso     []

***********************************************************************/
// compute range of a name (different from range of a multi-bit wire)
func Bac_ObjGetRange(p *Bac_Ntk_t, iObj int32) int32 {
	var i int32
	var NameId int32 = Bac_ObjName(p, iObj)
	assert(Bac_ObjIsCi(p, iObj) != 0, "Bac_ObjIsCi(p, iObj)")
	//    if ( Bac_NameType(NameId) == BAC_NAME_INDEX )
	//        NameId = Bac_ObjName(p, iObj - Abc_Lit2Var2(NameId));
	assert(Bac_NameType(NameId) == BAC_NAME_WORD || Bac_NameType(NameId) == BAC_NAME_INFO, "Bac_NameType(NameId) == BAC_NAME_WORD || Bac_NameType(NameId) == BAC_NAME_INFO")
	for i = iObj + 1; i < Bac_NtkObjNum(p); i++ {
		if Bac_ObjIsCi(p, i) == 0 || Bac_ObjNameType(p, i) != BAC_NAME_INDEX {
			break
		}
	}
	return i - iObj
}

func Bac_ManWriteVar(p *Bac_Ntk_t, RealName int32) {
	Vec_StrPrintStr(p.pDesign.vOut, Bac_NtkStr(p, RealName))
}
func Bac_ManWriteRange(p *Bac_Ntk_t, Beg int32, End int32) {
	var vStr *Vec_Str_t = p.pDesign.vOut
	Vec_StrPrintStr(vStr, "[")
	if End >= 0 {
		Vec_StrPrintNum(vStr, End)
		Vec_StrPrintStr(vStr, ":")
	}
	Vec_StrPrintNum(vStr, Beg)
	Vec_StrPrintStr(vStr, "]")
}
func Bac_ManWriteConstBit(p *Bac_Ntk_t, iObj int32, fHead int32) {
	var vStr *Vec_Str_t = p.pDesign.vOut
	var Const int32 = Bac_ObjGetConst(p, iObj)
	assert((Const != 0), "Const")
	if fHead != 0 {
		Vec_StrPrintStr(vStr, "1'b")
	}
	if Const == int32(BAC_BOX_CF) {
		Vec_StrPush(vStr, '0')
	} else if Const == int32(BAC_BOX_CT) {
		Vec_StrPush(vStr, '1')
	} else if Const == int32(BAC_BOX_CX) {
		Vec_StrPush(vStr, 'x')
	} else if Const == int32(BAC_BOX_CZ) {
		Vec_StrPush(vStr, 'z')
	} else {
		assert(false, "0")
	}
}
func Bac_ManFindRealNameId(p *Bac_Ntk_t, iObj int32) int32 {
	var NameId int32 = Bac_ObjName(p, iObj)
	assert(Bac_ObjIsCi(p, iObj) != 0, "Bac_ObjIsCi(p, iObj)")
	if Bac_NameType(NameId) == BAC_NAME_INDEX {
		NameId = Bac_ObjName(p, iObj-Abc_Lit2Var2(NameId))
	}
	if Bac_NameType(NameId) == BAC_NAME_INFO {
		return Bac_NtkInfoName(p, Abc_Lit2Var2(NameId))
	}
	assert(Bac_NameType(NameId) == BAC_NAME_BIN || Bac_NameType(NameId) == BAC_NAME_WORD, "Bac_NameType(NameId) == BAC_NAME_BIN || Bac_NameType(NameId) == BAC_NAME_WORD")
	return Abc_Lit2Var2(NameId)
}
func Bac_ManFindRealIndex(p *Bac_Ntk_t, iObj int32) int32 {
	var iBit int32 = 0
	var NameId int32 = Bac_ObjName(p, iObj)
	assert(Bac_ObjIsCi(p, iObj) != 0, "Bac_ObjIsCi(p, iObj)")
	assert(Bac_NameType(NameId) != BAC_NAME_BIN, "Bac_NameType(NameId) != BAC_NAME_BIN")
	if Bac_NameType(NameId) == BAC_NAME_INDEX {
		iBit = Abc_Lit2Var2(NameId)
		NameId = Bac_ObjName(p, iObj-iBit)
	}
	if Bac_NameType(NameId) == BAC_NAME_INFO {
		return Bac_NtkInfoIndex(p, Abc_Lit2Var2(NameId), iBit)
	}
	assert(Bac_NameType(NameId) == BAC_NAME_WORD, "Bac_NameType(NameId) == BAC_NAME_WORD")
	return iBit
}
func Bac_ManWriteSig(p *Bac_Ntk_t, iObj int32) {
	if Bac_ObjIsCo(p, iObj) != 0 {
		iObj = Bac_ObjFanin(p, iObj)
	}
	assert(Bac_ObjIsCi(p, iObj) != 0, "Bac_ObjIsCi(p, iObj)")
	if Bac_ObjGetConst(p, iObj) != 0 {
		Bac_ManWriteConstBit(p, iObj, 1)
	} else {
		var NameId int32 = Bac_ObjName(p, iObj)
		if Bac_NameType(NameId) == BAC_NAME_BIN {
			Bac_ManWriteVar(p, Abc_Lit2Var2(NameId))
		} else {
			Bac_ManWriteVar(p, Bac_ManFindRealNameId(p, iObj))
			Bac_ManWriteRange(p, Bac_ManFindRealIndex(p, iObj), -1)
		}
	}
}
func Bac_ManWriteConcat(p *Bac_Ntk_t, iStart int32, nObjs int32) {
	var vStr *Vec_Str_t = p.pDesign.vOut
	assert(nObjs >= 1, "nObjs >= 1")
	if nObjs == 1 {
		Bac_ManWriteSig(p, iStart)
		return
	}
	Vec_StrPrintStr(vStr, "{")
	if Bac_ObjIsBo(p, iStart) != 0 { // box output

		var i int32
		for i = iStart + nObjs - 1; i >= iStart; i-- {
			if Bac_ObjNameType(p, i) == BAC_NAME_INDEX {
				continue
			}
			if Vec_StrEntryLast(vStr) != '{' {
				Vec_StrPrintStr(vStr, ", ")
			}
			Bac_ManWriteVar(p, Bac_ManFindRealNameId(p, i))
		}

	} else if Bac_ObjIsBi(p, iStart) != 0 { // box input

		var e, b, k, NameId int32
		for e = iStart - nObjs + 1; e <= iStart; {
			if Vec_StrEntryLast(vStr) != '{' {
				Vec_StrPrintStr(vStr, ", ")
			}
			// write constant
			if Bac_ObjGetConst(p, Bac_ObjFanin(p, e)) != 0 {
				var fBinary int32 = Bac_ObjIsConstBin(p, Bac_ObjFanin(p, e)-1)
				for b = e + 1; b <= iStart; b++ {
					if Bac_ObjGetConst(p, Bac_ObjFanin(p, b)) == 0 {
						break
					}
					if Bac_ObjIsConstBin(p, Bac_ObjFanin(p, b)-1) == 0 {
						fBinary = 0
					}
				}
				Vec_StrPrintNum(vStr, b-e)
				if (fBinary != 0) && b-e > 8 { // write hex if more than 8 bits

					var Digit int32 = 0
					var nBits int32
					if (b-e)&3 != 0 {
						nBits = (b - e) & 3
					} else {
						nBits = 4
					}
					Vec_StrPrintStr(vStr, "'h")
					for k = e; k < b; k++ {
						Digit = 2*Digit + Bac_ObjGetConst(p, Bac_ObjFanin(p, k)) - int32(BAC_BOX_CF)
						assert(Digit < 16, "Digit < 16")
						nBits--
						if nBits == 0 {
							if Digit < 10 {
								Vec_StrPush(vStr, byte('0'+Digit))
							} else {
								Vec_StrPush(vStr, byte('a'+Digit-10))
							}
							nBits = 4
							Digit = 0
						}
					}
					assert(nBits == 4, "nBits == 4")
					assert(Digit == 0, "Digit == 0")

				} else {
					Vec_StrPrintStr(vStr, "'b")
					for k = e; k < b; k++ {
						Bac_ManWriteConstBit(p, Bac_ObjFanin(p, k), 0)
					}
				}
				e = b
				continue
			}
			// try replication
			for b = e + 1; b <= iStart; b++ {
				if Bac_ObjFanin(p, b) != Bac_ObjFanin(p, e) {
					break
				}
			}
			if b > e+2 { // more than two

				Vec_StrPrintNum(vStr, b-e)
				Vec_StrPrintStr(vStr, "{")
				Bac_ManWriteSig(p, e)
				Vec_StrPrintStr(vStr, "}")
				e = b
				continue
			}
			NameId = Bac_ObjName(p, Bac_ObjFanin(p, e))
			if Bac_NameType(NameId) == BAC_NAME_BIN {
				Bac_ManWriteVar(p, Abc_Lit2Var2(NameId))
				e++
				continue
			}
			// find end of the slice
			for b = e + 1; b <= iStart; b++ {
				if Bac_ObjFanin(p, e)-Bac_ObjFanin(p, b) != b-e {
					break
				}
			}
			// write signal name
			Bac_ManWriteVar(p, Bac_ManFindRealNameId(p, Bac_ObjFanin(p, e)))
			if b == e+1 { // literal
				Bac_ManWriteRange(p, Bac_ManFindRealIndex(p, Bac_ObjFanin(p, e)), -1)
			} else { // slice or complete variable

				// consider first variable of the slice
				var f int32 = Bac_ObjFanin(p, b-1)
				assert(Bac_ObjNameType(p, f) != BAC_NAME_BIN, "Bac_ObjNameType(p, f) != BAC_NAME_BIN")
				if Bac_ObjNameType(p, f) == BAC_NAME_INDEX || Bac_ObjGetRange(p, f) != b-e { // slice
					Bac_ManWriteRange(p, Bac_ManFindRealIndex(p, f), Bac_ManFindRealIndex(p, Bac_ObjFanin(p, e)))
				}
				// else this is complete variable
			}
			e = b
		}

	} else {
		assert(false, "0")
	}
	Vec_StrPrintStr(vStr, "}")
}
func Bac_ManWriteGate(p *Bac_Ntk_t, iObj int32) {
	vStr := p.pDesign.vOut
	pGateName := Abc_NamStr(p.pDesign.pMods, Bac_BoxNtkId(p, iObj))
	pLib := Abc_FrameReadLibGen()
	pGate := Mio_LibraryReadGateByName(pLib, pGateName, "")
	Vec_StrPrintStr(vStr, "  ")
	Vec_StrPrintStr(vStr, pGateName)
	Vec_StrPrintStr(vStr, " ")
	if Bac_ObjName(p, iObj) != 0 {
		Vec_StrPrintStr(vStr, Bac_ObjNameStr(p, iObj))
	}
	Vec_StrPrintStr(vStr, " (")
	for iTerm, k := iObj-1, int32(0); iTerm >= 0 && Bac_ObjIsBi(p, iTerm) != 0; iTerm, k = iTerm-1, k+1 {
		if k != 0 {
			Vec_StrPrintStr(vStr, ", .")
		} else {
			Vec_StrPrintStr(vStr, ".")
		}
		Vec_StrPrintStr(vStr, Mio_GateReadPinName(pGate, k))
		Vec_StrPrintStr(vStr, "(")
		Bac_ManWriteSig(p, iTerm)
		Vec_StrPrintStr(vStr, ")")
	}
	for iTerm := iObj + 1; iTerm < Bac_NtkObjNum(p) && Bac_ObjIsBo(p, iTerm) != 0; iTerm++ {
		if Bac_BoxBiNum(p, iObj) != 0 {
			Vec_StrPrintStr(vStr, ", .")
		} else {
			Vec_StrPrintStr(vStr, ".")
		}
		Vec_StrPrintStr(vStr, Mio_GateReadOutName(pGate))
		Vec_StrPrintStr(vStr, "(")
		Bac_ManWriteSig(p, iTerm)
		Vec_StrPrintStr(vStr, ")")
	}
	Vec_StrPrintStr(vStr, ");\n")
}
func Bac_ManWriteAssign(p *Bac_Ntk_t, iObj int32) {
	vStr := p.pDesign.vOut
	Type := Bac_ObjType(p, iObj)
	nInputs, nOutputs := Bac_BoxBiNum(p, iObj), Bac_BoxBoNum(p, iObj)
	assert(nOutputs == 1, "nOutputs == 1")
	Vec_StrPrintStr(vStr, "  assign ")
	Bac_ManWriteSig(p, iObj+1)
	Vec_StrPrintStr(vStr, " = ")
	if nInputs == 0 {
		if Type == BAC_BOX_CF {
			Vec_StrPrintStr(vStr, "1'b0")
		} else if Type == BAC_BOX_CT {
			Vec_StrPrintStr(vStr, "1'b1")
		} else if Type == BAC_BOX_CX {
			Vec_StrPrintStr(vStr, "1'bx")
		} else if Type == BAC_BOX_CZ {
			Vec_StrPrintStr(vStr, "1'bz")
		} else {
			assert(false, "0")
		}

	} else if nInputs == 1 {
		if Type == BAC_BOX_INV {
			Vec_StrPrintStr(vStr, "~")
		} else {
			assert(Type == BAC_BOX_BUF, "Type == BAC_BOX_BUF")
		}
		Bac_ManWriteSig(p, iObj-1)

	} else if nInputs == 2 {
		if Type == BAC_BOX_NAND || Type == BAC_BOX_NOR || Type == BAC_BOX_XNOR || Type == BAC_BOX_SHARPL {
			Vec_StrPrintStr(vStr, "~")
		}
		Bac_ManWriteSig(p, iObj-1)
		if Type == BAC_BOX_AND || Type == BAC_BOX_SHARPL {
			Vec_StrPrintStr(vStr, " & ")
		} else if Type == BAC_BOX_SHARP || Type == BAC_BOX_NOR {
			Vec_StrPrintStr(vStr, " & ~")
		} else if Type == BAC_BOX_OR {
			Vec_StrPrintStr(vStr, " | ")
		} else if Type == BAC_BOX_NAND {
			Vec_StrPrintStr(vStr, " | ~")
		} else if Type == BAC_BOX_XOR || Type == BAC_BOX_XNOR {
			Vec_StrPrintStr(vStr, " ^ ")
		} else {
			assert(false, "0")
		}
		Bac_ManWriteSig(p, iObj-2)
	}
	Vec_StrPrintStr(vStr, ";\n")
}
func Bac_ManWriteVerilogBoxes(p *Bac_Ntk_t, fUseAssign int32) {
	vStr := p.pDesign.vOut
	for iObj := int32(0); iObj < Bac_NtkObjNum(p); iObj++ {
		if Bac_ObjIsBox(p, iObj) == 0 || Bac_ObjIsConst(p, iObj) != 0 {
			continue
		}
		if Bac_ObjIsGate(p, iObj) != 0 {
			Bac_ManWriteGate(p, iObj)
			continue
		}
		if Bac_ObjIsBoxUser(p, iObj) == 0 && fUseAssign != 0 {
			Bac_ManWriteAssign(p, iObj)
			continue
		}
		startPos := Vec_StrSize(vStr)
		if Bac_ObjIsBoxUser(p, iObj) != 0 {
			pModel := Bac_BoxNtk(p, iObj)
			Vec_StrPrintStr(vStr, "  ")
			Vec_StrPrintStr(vStr, Bac_NtkName(pModel))
			Vec_StrPrintStr(vStr, " ")
			if Bac_ObjName(p, iObj) != 0 {
				Vec_StrPrintStr(vStr, Bac_ObjNameStr(p, iObj))
			}
			Vec_StrPrintStr(vStr, " (")
			i, o := int32(0), int32(0)
			assert(Bac_NtkInfoNum(pModel) != 0, "Bac_NtkInfoNum(pModel)")
			for k := int32(0); k+2 < Vec_IntSize(&pModel.vInfo); k += 3 {
				value, beg, end := Vec_IntEntry(&pModel.vInfo, k), Vec_IntEntry(&pModel.vInfo, k+1), Vec_IntEntry(&pModel.vInfo, k+2)
				nameId, typ := Abc_Lit2Var2(value), Abc_Lit2Att2(value)
				if k != 0 {
					Vec_StrPrintStr(vStr, ", ")
				}
				if Vec_StrSize(vStr) > startPos+70 {
					startPos = Vec_StrSize(vStr)
					Vec_StrPrintStr(vStr, "\n    ")
				}
				Vec_StrPrintStr(vStr, ".")
				Vec_StrPrintStr(vStr, Bac_NtkStr(p, nameId))
				Vec_StrPrintStr(vStr, "(")
				ran := Bac_InfoRange(beg, end)
				assert(ran > 0, "Range > 0")
				if typ == 1 {
					Bac_ManWriteConcat(p, Bac_BoxBi(p, iObj, i), ran)
					i += ran
				} else if typ == 2 {
					Bac_ManWriteConcat(p, Bac_BoxBo(p, iObj, o), ran)
					o += ran
				} else {
					assert(false, "0")
				}
				Vec_StrPrintStr(vStr, ")")
			}
			assert(i == Bac_BoxBiNum(p, iObj), "i == Bac_BoxBiNum(p, iObj)")
			assert(o == Bac_BoxBoNum(p, iObj), "o == Bac_BoxBoNum(p, iObj)")
		} else {
			vBits := Bac_BoxCollectRanges(p, iObj)
			typ := Bac_ObjType(p, iObj)
			pName, pSymbs := Bac_NtkGenerateName(p, typ, vBits), Bac_ManPrimSymb(p.pDesign, typ)
			Vec_StrPrintStr(vStr, "  ")
			Vec_StrPrintStr(vStr, pName)
			Vec_StrPrintStr(vStr, " ")
			if Bac_ObjName(p, iObj) != 0 {
				Vec_StrPrintStr(vStr, Bac_ObjNameStr(p, iObj))
			}
			Vec_StrPrintStr(vStr, " (")
			iSig := int32(0)
			writeMain := func(iTerm int32) {
				ran := Vec_IntEntry(vBits, iSig)
				if iSig != 0 {
					Vec_StrPrintStr(vStr, ", ")
				}
				if Vec_StrSize(vStr) > startPos+70 {
					startPos = Vec_StrSize(vStr)
					Vec_StrPrintStr(vStr, "\n    ")
				}
				Vec_StrPrintStr(vStr, ".")
				Vec_StrPush(vStr, pSymbs[iSig])
				Vec_StrPrintStr(vStr, "(")
				Bac_ManWriteConcat(p, iTerm, ran)
				Vec_StrPrintStr(vStr, ")")
				iSig++
			}
			for iTerm := iObj - 1; iTerm >= 0 && Bac_ObjIsBi(p, iTerm) != 0; iTerm-- {
				if Bac_ObjBit(p, iTerm) == 0 {
					writeMain(iTerm)
				}
			}
			for iTerm := iObj + 1; iTerm < Bac_NtkObjNum(p) && Bac_ObjIsBo(p, iTerm) != 0; iTerm++ {
				if Bac_ObjBit(p, iTerm) == 0 {
					writeMain(iTerm)
				}
			}
			assert(iSig == Vec_IntSize(vBits), "iSig == Vec_IntSize(vBits)")
		}
		Vec_StrPrintStr(vStr, ");\n")
	}
}
func Bac_ManWriteVerilogNtk(p *Bac_Ntk_t, fUseAssign int32) {
	keywords := [4]string{"wire ", "input ", "output ", "inout "}
	vStr := p.pDesign.vOut
	assert(Vec_IntSize(&p.vFanin) == Bac_NtkObjNum(p), "Vec_IntSize(&p.vFanin) == Bac_NtkObjNum(p)")
	Vec_StrPrintStr(vStr, "module ")
	Vec_StrPrintStr(vStr, Bac_NtkName(p))
	Vec_StrPrintStr(vStr, " (\n    ")
	startPos := Vec_StrSize(vStr)
	for k := int32(0); k+2 < Vec_IntSize(&p.vInfo); k += 3 {
		value := Vec_IntEntry(&p.vInfo, k)
		if Abc_Lit2Att2(value) != 0 {
			if k != 0 {
				Vec_StrPrintStr(vStr, ", ")
			}
			if Vec_StrSize(vStr) > startPos+70 {
				startPos = Vec_StrSize(vStr)
				Vec_StrPrintStr(vStr, "\n    ")
			}
			Bac_ManWriteVar(p, Abc_Lit2Var2(value))
		}
	}
	Vec_StrPrintStr(vStr, "\n  );\n")
	for k := int32(0); k+2 < Vec_IntSize(&p.vInfo); k += 3 {
		value, beg, end := Vec_IntEntry(&p.vInfo, k), Vec_IntEntry(&p.vInfo, k+1), Vec_IntEntry(&p.vInfo, k+2)
		attr := Abc_Lit2Att2(value)
		if attr != 0 {
			Vec_StrPrintStr(vStr, "  ")
			Vec_StrPrintStr(vStr, keywords[attr])
			if beg >= 0 {
				Bac_ManWriteRange(p, beg, end)
			}
			Bac_ManWriteVar(p, Abc_Lit2Var2(value))
			Vec_StrPrintStr(vStr, ";\n")
		}
	}
	Vec_StrPrintStr(vStr, "\n")
	for iObj := int32(0); iObj < Bac_NtkObjNum(p); iObj++ {
		if Bac_ObjIsBox(p, iObj) == 0 || Bac_ObjIsConst(p, iObj) != 0 {
			continue
		}
		for iTerm := iObj + 1; iTerm < Bac_NtkObjNum(p) && Bac_ObjIsBo(p, iTerm) != 0; iTerm++ {
			if Bac_ObjNameType(p, iTerm) == BAC_NAME_WORD || Bac_ObjNameType(p, iTerm) == BAC_NAME_INFO {
				ran := Bac_ObjGetRange(p, iTerm)
				Vec_StrPrintStr(vStr, "  wire ")
				Bac_ManWriteRange(p, Bac_ManFindRealIndex(p, iTerm), Bac_ManFindRealIndex(p, iTerm+ran-1))
				Bac_ManWriteVar(p, Bac_ManFindRealNameId(p, iTerm))
				Vec_StrPrintStr(vStr, ";\n")
			}
		}
	}
	haveWires := false
	for iObj := int32(0); iObj < Bac_NtkObjNum(p) && !haveWires; iObj++ {
		if Bac_ObjIsBox(p, iObj) == 0 || Bac_ObjIsConst(p, iObj) != 0 {
			continue
		}
		for iTerm := iObj + 1; iTerm < Bac_NtkObjNum(p) && Bac_ObjIsBo(p, iTerm) != 0; iTerm++ {
			if Bac_ObjNameType(p, iTerm) == BAC_NAME_BIN {
				haveWires = true
				break
			}
		}
	}
	if haveWires {
		length := int32(7)
		Vec_StrPrintStr(vStr, "\n  wire ")
		for iObj := int32(0); iObj < Bac_NtkObjNum(p); iObj++ {
			if Bac_ObjIsBox(p, iObj) == 0 || Bac_ObjIsConst(p, iObj) != 0 {
				continue
			}
			for iTerm := iObj + 1; iTerm < Bac_NtkObjNum(p) && Bac_ObjIsBo(p, iTerm) != 0; iTerm++ {
				if Bac_ObjNameType(p, iTerm) != BAC_NAME_BIN {
					continue
				}
				name := Bac_ObjNameStr(p, iTerm)
				if length > 72 {
					Vec_StrPrintStr(vStr, ";\n  wire ")
					length = 7
				}
				if length > 7 {
					Vec_StrPrintStr(vStr, ", ")
				}
				Vec_StrPrintStr(vStr, name)
				length += int32(len(name))
			}
		}
		Vec_StrPrintStr(vStr, ";\n")
	}
	Vec_StrPrintStr(vStr, "\n")
	Bac_ManWriteVerilogBoxes(p, fUseAssign)
	Vec_StrPrintStr(vStr, "endmodule\n\n")
}
func Bac_ManWriteVerilog(pFileName string, p *Bac_Man_t, fUseAssign int32) {
	if p.pMioLib != nil && p.pMioLib != Abc_FrameReadLibGen() {
		fmt.Printf("Genlib library used in the mapped design is not longer a current library.\n")
		return
	}
	p.vOut = Vec_StrAlloc(10000)
	p.vOut2 = Vec_StrAlloc(1000)
	Vec_StrPrintStr(p.vOut, "// Design \"")
	Vec_StrPrintStr(p.vOut, Bac_ManName(p))
	Vec_StrPrintStr(p.vOut, "\" written via CBA package in ABC on ")
	Vec_StrPrintStr(p.vOut, Extra_TimeStamp())
	Vec_StrPrintStr(p.vOut, "\n\n")
	Bac_ManAssignInternWordNames(p)
	for i := int32(1); i <= Bac_ManNtkNum(p); i++ {
		Bac_ManWriteVerilogNtk(Bac_ManNtk(p, i), fUseAssign)
	}
	if p.vOut != nil && Vec_StrSize(p.vOut) > 0 {
		file, err := os.OpenFile(pFileName, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, 0666)
		if err != nil {
			fmt.Printf("Cannot open file \"%s\" for writing.\n", pFileName)
		} else {
			_, _ = file.Write(Vec_StrArray(p.vOut)[:Vec_StrSize(p.vOut)])
			_ = file.Close()
		}
	}
	Vec_StrFreeP(&p.vOut)
	Vec_StrFreeP(&p.vOut2)
}

////////////////////////////////////////////////////////////////////////
///                       END OF FILE                                ///
////////////////////////////////////////////////////////////////////////

//ABC_NAMESPACE_IMPL_END
