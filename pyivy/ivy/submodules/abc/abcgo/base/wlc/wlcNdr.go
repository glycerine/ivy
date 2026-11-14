package wlc

import (
	"encoding/binary"
	"fmt"
	"strings"

	"github.com/glycerine/abcgo/aig/miniaig"
	"github.com/glycerine/abcgo/misc/util"
	"github.com/glycerine/abcgo/misc/vec"
)

/**CFile****************************************************************

  FileName    [wlcNdr.c]

  SystemName  [ABC: Logic synthesis and verification system.]

  PackageName [Experimental procedures.]

  Synopsis    [Constructing WLC network from NDR data structure.]

  Author      [Alan Mishchenko]
  
  Affiliation [UC Berkeley]

  Date        [Ver. 1.0. Started - August 22, 2014.]

  Revision    [$Id: wlcNdr.c,v 1.00 2014/09/12 00:00:00 alanmi Exp $]

***********************************************************************/

//#include "wlc.h"
//#include "aig/miniaig/ndr.h"

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
func Ndr_TypeNdr2Wlc(typ int32) int32 {
	switch typ {
	case miniaig.ABC_OPER_CONST:
		return int32(WLC_OBJ_CONST)
	case miniaig.ABC_OPER_BIT_BUF:
		return int32(WLC_OBJ_BUF)
	case miniaig.ABC_OPER_BIT_MUX:
		return int32(WLC_OBJ_MUX)
	case miniaig.ABC_OPER_SHIFT_R:
		return int32(WLC_OBJ_SHIFT_R)
	case miniaig.ABC_OPER_SHIFT_RA:
		return int32(WLC_OBJ_SHIFT_RA)
	case miniaig.ABC_OPER_SHIFT_L:
		return int32(WLC_OBJ_SHIFT_L)
	case miniaig.ABC_OPER_SHIFT_LA:
		return int32(WLC_OBJ_SHIFT_LA)
	case miniaig.ABC_OPER_SHIFT_ROTR:
		return int32(WLC_OBJ_ROTATE_R)
	case miniaig.ABC_OPER_SHIFT_ROTL:
		return int32(WLC_OBJ_ROTATE_L)
	case miniaig.ABC_OPER_BIT_INV:
		return int32(WLC_OBJ_BIT_NOT)
	case miniaig.ABC_OPER_BIT_AND:
		return int32(WLC_OBJ_BIT_AND)
	case miniaig.ABC_OPER_BIT_OR:
		return int32(WLC_OBJ_BIT_OR)
	case miniaig.ABC_OPER_BIT_XOR:
		return int32(WLC_OBJ_BIT_XOR)
	case miniaig.ABC_OPER_BIT_NAND:
		return int32(WLC_OBJ_BIT_NAND)
	case miniaig.ABC_OPER_BIT_NOR:
		return int32(WLC_OBJ_BIT_NOR)
	case miniaig.ABC_OPER_BIT_NXOR:
		return int32(WLC_OBJ_BIT_NXOR)
	case miniaig.ABC_OPER_SLICE:
		return int32(WLC_OBJ_BIT_SELECT)
	case miniaig.ABC_OPER_CONCAT:
		return int32(WLC_OBJ_BIT_CONCAT)
	case miniaig.ABC_OPER_ZEROPAD:
		return int32(WLC_OBJ_BIT_ZEROPAD)
	case miniaig.ABC_OPER_SIGNEXT:
		return int32(WLC_OBJ_BIT_SIGNEXT)
	case miniaig.ABC_OPER_LOGIC_NOT:
		return int32(WLC_OBJ_LOGIC_NOT)
	case miniaig.ABC_OPER_LOGIC_IMPL:
		return int32(WLC_OBJ_LOGIC_IMPL)
	case miniaig.ABC_OPER_LOGIC_AND:
		return int32(WLC_OBJ_LOGIC_AND)
	case miniaig.ABC_OPER_LOGIC_OR:
		return int32(WLC_OBJ_LOGIC_OR)
	case miniaig.ABC_OPER_LOGIC_XOR:
		return int32(WLC_OBJ_LOGIC_XOR)
	case miniaig.ABC_OPER_SEL_NMUX:
		return int32(WLC_OBJ_MUX)
	case miniaig.ABC_OPER_SEL_SEL:
		return int32(WLC_OBJ_SEL)
	case miniaig.ABC_OPER_SEL_DEC:
		return int32(WLC_OBJ_DEC)
	case miniaig.ABC_OPER_COMP_EQU:
		return int32(WLC_OBJ_COMP_EQU)
	case miniaig.ABC_OPER_COMP_NOTEQU:
		return int32(WLC_OBJ_COMP_NOTEQU)
	case miniaig.ABC_OPER_COMP_LESS:
		return int32(WLC_OBJ_COMP_LESS)
	case miniaig.ABC_OPER_COMP_MORE:
		return int32(WLC_OBJ_COMP_MORE)
	case miniaig.ABC_OPER_COMP_LESSEQU:
		return int32(WLC_OBJ_COMP_LESSEQU)
	case miniaig.ABC_OPER_COMP_MOREEQU:
		return int32(WLC_OBJ_COMP_MOREEQU)
	case miniaig.ABC_OPER_RED_AND:
		return int32(WLC_OBJ_REDUCT_AND)
	case miniaig.ABC_OPER_RED_OR:
		return int32(WLC_OBJ_REDUCT_OR)
	case miniaig.ABC_OPER_RED_XOR:
		return int32(WLC_OBJ_REDUCT_XOR)
	case miniaig.ABC_OPER_RED_NAND:
		return int32(WLC_OBJ_REDUCT_NAND)
	case miniaig.ABC_OPER_RED_NOR:
		return int32(WLC_OBJ_REDUCT_NOR)
	case miniaig.ABC_OPER_RED_NXOR:
		return int32(WLC_OBJ_REDUCT_NXOR)
	case miniaig.ABC_OPER_ARI_ADD:
		return int32(WLC_OBJ_ARI_ADD)
	case miniaig.ABC_OPER_ARI_SUB:
		return int32(WLC_OBJ_ARI_SUB)
	case miniaig.ABC_OPER_ARI_MUL:
		return int32(WLC_OBJ_ARI_MULTI)
	case miniaig.ABC_OPER_ARI_DIV:
		return int32(WLC_OBJ_ARI_DIVIDE)
	case miniaig.ABC_OPER_ARI_REM:
		return int32(WLC_OBJ_ARI_REM)
	case miniaig.ABC_OPER_ARI_MOD:
		return int32(WLC_OBJ_ARI_MODULUS)
	case miniaig.ABC_OPER_ARI_POW:
		return int32(WLC_OBJ_ARI_POWER)
	case miniaig.ABC_OPER_ARI_MIN:
		return int32(WLC_OBJ_ARI_MINUS)
	case miniaig.ABC_OPER_ARI_SQRT:
		return int32(WLC_OBJ_ARI_SQRT)
	case miniaig.ABC_OPER_ARI_SQUARE:
		return int32(WLC_OBJ_ARI_SQUARE)
	case miniaig.ABC_OPER_ARI_ADDSUB:
		return int32(WLC_OBJ_ARI_ADDSUB)
	case miniaig.ABC_OPER_ARI_SMUL:
		return int32(WLC_OBJ_ARI_MULTI)
	case miniaig.ABC_OPER_DFF:
		return int32(WLC_OBJ_FO)
	case miniaig.ABC_OPER_DFFRSE:
		return int32(WLC_OBJ_FF)
	case miniaig.ABC_OPER_RAMR:
		return int32(WLC_OBJ_READ)
	case miniaig.ABC_OPER_RAMW:
		return int32(WLC_OBJ_WRITE)
	case miniaig.ABC_OPER_LUT:
		return int32(WLC_OBJ_LUT)
	default:
		return -1
	}
}
func Ndr_TypeWlc2Ndr(typ int32) int32 {
	switch typ {
	case int32(WLC_OBJ_CONST):
		return miniaig.ABC_OPER_CONST
	case int32(WLC_OBJ_BUF):
		return miniaig.ABC_OPER_BIT_BUF
	case int32(WLC_OBJ_MUX):
		return miniaig.ABC_OPER_BIT_MUX
	case int32(WLC_OBJ_SHIFT_R):
		return miniaig.ABC_OPER_SHIFT_R
	case int32(WLC_OBJ_SHIFT_RA):
		return miniaig.ABC_OPER_SHIFT_RA
	case int32(WLC_OBJ_SHIFT_L):
		return miniaig.ABC_OPER_SHIFT_L
	case int32(WLC_OBJ_SHIFT_LA):
		return miniaig.ABC_OPER_SHIFT_LA
	case int32(WLC_OBJ_ROTATE_R):
		return miniaig.ABC_OPER_SHIFT_ROTR
	case int32(WLC_OBJ_ROTATE_L):
		return miniaig.ABC_OPER_SHIFT_ROTL
	case int32(WLC_OBJ_BIT_NOT):
		return miniaig.ABC_OPER_BIT_INV
	case int32(WLC_OBJ_BIT_AND):
		return miniaig.ABC_OPER_BIT_AND
	case int32(WLC_OBJ_BIT_OR):
		return miniaig.ABC_OPER_BIT_OR
	case int32(WLC_OBJ_BIT_XOR):
		return miniaig.ABC_OPER_BIT_XOR
	case int32(WLC_OBJ_BIT_NAND):
		return miniaig.ABC_OPER_BIT_NAND
	case int32(WLC_OBJ_BIT_NOR):
		return miniaig.ABC_OPER_BIT_NOR
	case int32(WLC_OBJ_BIT_NXOR):
		return miniaig.ABC_OPER_BIT_NXOR
	case int32(WLC_OBJ_BIT_SELECT):
		return miniaig.ABC_OPER_SLICE
	case int32(WLC_OBJ_BIT_CONCAT):
		return miniaig.ABC_OPER_CONCAT
	case int32(WLC_OBJ_BIT_ZEROPAD):
		return miniaig.ABC_OPER_ZEROPAD
	case int32(WLC_OBJ_BIT_SIGNEXT):
		return miniaig.ABC_OPER_SIGNEXT
	case int32(WLC_OBJ_LOGIC_NOT):
		return miniaig.ABC_OPER_LOGIC_NOT
	case int32(WLC_OBJ_LOGIC_IMPL):
		return miniaig.ABC_OPER_LOGIC_IMPL
	case int32(WLC_OBJ_LOGIC_AND):
		return miniaig.ABC_OPER_LOGIC_AND
	case int32(WLC_OBJ_LOGIC_OR):
		return miniaig.ABC_OPER_LOGIC_OR
	case int32(WLC_OBJ_LOGIC_XOR):
		return miniaig.ABC_OPER_LOGIC_XOR
	case int32(WLC_OBJ_SEL):
		return miniaig.ABC_OPER_SEL_SEL
	case int32(WLC_OBJ_DEC):
		return miniaig.ABC_OPER_SEL_DEC
	case int32(WLC_OBJ_COMP_EQU):
		return miniaig.ABC_OPER_COMP_EQU
	case int32(WLC_OBJ_COMP_NOTEQU):
		return miniaig.ABC_OPER_COMP_NOTEQU
	case int32(WLC_OBJ_COMP_LESS):
		return miniaig.ABC_OPER_COMP_LESS
	case int32(WLC_OBJ_COMP_MORE):
		return miniaig.ABC_OPER_COMP_MORE
	case int32(WLC_OBJ_COMP_LESSEQU):
		return miniaig.ABC_OPER_COMP_LESSEQU
	case int32(WLC_OBJ_COMP_MOREEQU):
		return miniaig.ABC_OPER_COMP_MOREEQU
	case int32(WLC_OBJ_REDUCT_AND):
		return miniaig.ABC_OPER_RED_AND
	case int32(WLC_OBJ_REDUCT_OR):
		return miniaig.ABC_OPER_RED_OR
	case int32(WLC_OBJ_REDUCT_XOR):
		return miniaig.ABC_OPER_RED_XOR
	case int32(WLC_OBJ_REDUCT_NAND):
		return miniaig.ABC_OPER_RED_NAND
	case int32(WLC_OBJ_REDUCT_NOR):
		return miniaig.ABC_OPER_RED_NOR
	case int32(WLC_OBJ_REDUCT_NXOR):
		return miniaig.ABC_OPER_RED_NXOR
	case int32(WLC_OBJ_ARI_ADD):
		return miniaig.ABC_OPER_ARI_ADD
	case int32(WLC_OBJ_ARI_SUB):
		return miniaig.ABC_OPER_ARI_SUB
	case int32(WLC_OBJ_ARI_MULTI):
		return miniaig.ABC_OPER_ARI_MUL
	case int32(WLC_OBJ_ARI_DIVIDE):
		return miniaig.ABC_OPER_ARI_DIV
	case int32(WLC_OBJ_ARI_REM):
		return miniaig.ABC_OPER_ARI_REM
	case int32(WLC_OBJ_ARI_MODULUS):
		return miniaig.ABC_OPER_ARI_MOD
	case int32(WLC_OBJ_ARI_POWER):
		return miniaig.ABC_OPER_ARI_POW
	case int32(WLC_OBJ_ARI_MINUS):
		return miniaig.ABC_OPER_ARI_MIN
	case int32(WLC_OBJ_ARI_SQRT):
		return miniaig.ABC_OPER_ARI_SQRT
	case int32(WLC_OBJ_ARI_SQUARE):
		return miniaig.ABC_OPER_ARI_SQUARE
	case int32(WLC_OBJ_ARI_ADDSUB):
		return miniaig.ABC_OPER_ARI_ADDSUB
	case int32(WLC_OBJ_FO):
		return miniaig.ABC_OPER_DFFRSE
	case int32(WLC_OBJ_FF):
		return miniaig.ABC_OPER_DFFRSE
	case int32(WLC_OBJ_READ):
		return miniaig.ABC_OPER_RAMR
	case int32(WLC_OBJ_WRITE):
		return miniaig.ABC_OPER_RAMW
	case int32(WLC_OBJ_LUT):
		return miniaig.ABC_OPER_LUT
	default:
		return -1
	}
}


/**Function*************************************************************

  Synopsis    []

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Ndr_ObjWriteConstant(bits []uint32, nBits int32) string {
	assert(nBits+10 < 10000, "nBits + 10 < 10000")
	var result strings.Builder
	if nBits > 0 {
		result.Grow(int(nBits) + 10)
	}
	fmt.Fprintf(&result, "%d'b", nBits)
	for i := nBits - 1; i >= 0; i-- {
		bit := (bits[i>>5] >> uint(i&31)) & 1
		result.WriteByte(byte('0' + bit))
	}
	return result.String()
}
func Wlc_NtkToNdr(pNtk *Wlc_Ntk_t) *miniaig.Ndr_Data_t {
	design := miniaig.Ndr_Create(1)
	moduleID := miniaig.Ndr_AddModule(design, 1)

	// Add primary inputs first, matching Wlc_NtkForEachPi in the C source.
	for i := int32(0); i < Wlc_NtkPiNum(pNtk); i++ {
		obj := Wlc_NtkPi(pNtk, i)
		id := Wlc_ObjId(pNtk, obj)
		miniaig.Ndr_AddObject(design, moduleID, miniaig.ABC_OPER_CI, 0,
			int(obj.End), int(obj.Beg), int(obj.Signed), 0, nil, 1, []int{int(id)}, nil)
	}

	// Add internal nodes in object ID order so every NDR fanin refers to an
	// object already emitted, as in the C traversal.
	for id := int32(1); id < Wlc_NtkObjNumMax(pNtk); id++ {
		obj := Wlc_NtkObj(pNtk, id)
		if Wlc_ObjIsPi(obj) != 0 || obj.Type == 0 {
			continue
		}

		fanins := make([]int, Wlc_ObjFaninNum(obj))
		for k := range fanins {
			fanins[k] = int(Wlc_ObjFaninId(obj, int32(k)))
		}

		var function any
		if int32(obj.Type) == int32(WLC_OBJ_CONST) {
			wordCount := (int(Wlc_ObjRange(obj)) + 31) / 32
			bits := make([]uint32, wordCount)
			for i := range bits {
				bits[i] = uint32(Wlc_ObjFaninId(obj, int32(i)))
			}
			function = Ndr_ObjWriteConstant(bits, Wlc_ObjRange(obj))
		}

		ndrType := int32(0)
		switch {
		case int32(obj.Type) == int32(WLC_OBJ_MUX) && Wlc_ObjRange(Wlc_ObjFanin0(pNtk, obj)) > 1:
			ndrType = miniaig.ABC_OPER_SEL_NMUX
		case int32(obj.Type) == int32(WLC_OBJ_FO):
			pFi := Wlc_ObjFo2Fi(pNtk, obj)
			assert(len(fanins) == 0, "len(fanins) == 0")
			fanins = append(fanins, int(Wlc_ObjId(pNtk, pFi)))
			for len(fanins) < 7 {
				fanins = append(fanins, 0)
			}
			ndrType = miniaig.ABC_OPER_DFFRSE
		default:
			ndrType = Ndr_TypeWlc2Ndr(int32(obj.Type))
		}
		assert(ndrType > 0, "ndrType > 0")
		miniaig.Ndr_AddObject(design, moduleID, int(ndrType), 0,
			int(obj.End), int(obj.Beg), int(obj.Signed), len(fanins), fanins, 1, []int{int(id)}, function)
	}

	// Add primary outputs after all internal nodes, preserving source order.
	for id := int32(1); id < Wlc_NtkObjNumMax(pNtk); id++ {
		obj := Wlc_NtkObj(pNtk, id)
		if Wlc_ObjIsPo(obj) == 0 {
			continue
		}
		miniaig.Ndr_AddObject(design, moduleID, miniaig.ABC_OPER_CO, 0,
			int(obj.End), int(obj.Beg), int(obj.Signed), 1, []int{int(id)}, 0, nil, nil)
	}
	return design
}
func Wlc_WriteNdr(pNtk *Wlc_Ntk_t, fileName string) {
	design := Wlc_NtkToNdr(pNtk)
	miniaig.Ndr_Write(fileName, design)
	miniaig.Ndr_Delete(design)
	fmt.Printf("Dumped the current design into file \"%s\".\n", fileName)
}
func Wlc_NtkToNdrTest(pNtk *Wlc_Ntk_t) {
	design := Wlc_NtkToNdr(pNtk)
	defer miniaig.Ndr_Delete(design)

	// NDR name IDs are one-based; preserve the object ID-to-name table passed
	// to the C Verilog writer.
	names := make([]string, Wlc_NtkObjNum(pNtk)+1)
	for id := int32(1); id < Wlc_NtkObjNumMax(pNtk); id++ {
		names[id] = Wlc_ObjName(pNtk, id)
	}
	miniaig.Ndr_WriteVerilog("", design, names)
	miniaig.Ndr_Write("test.ndr", design)
}

/**Function*************************************************************

  Synopsis    []

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Ndr_ObjReadRange(p *miniaig.Ndr_Data_t, obj int, end, beg *int32) int32 {
	var values []int
	n := miniaig.Ndr_ObjReadArray(p, obj, miniaig.NDR_RANGE, &values)
	*end, *beg = 0, 0
	if n == 0 {
		return 0
	}
	signed := int32(0)
	if n == 3 {
		signed = 1
	}
	if n == 1 {
		*end, *beg = int32(values[0]), int32(values[0])
	} else {
		*end, *beg = int32(values[0]), int32(values[1])
	}
	return signed
}
func Ndr_ObjReadConstant(vFanins *vec.Vec_Int_t, str string) {
	k := strings.IndexByte(str, 'b')
	if k < 0 {
		fmt.Printf("Constants should be represented in binary Verilog notation <nbits>'b<bits> as char strings (for example, \"4'b1010\").\n")
		return
	}
	nBits := len(str) - k - 1
	vec.Vec_IntFill(vFanins, (nBits+31)/32, 0)
	for i := k + 1; i < len(str); i++ {
		bitIndex := len(str) - i - 1
		if str[i] == '1' {
			wordIndex := bitIndex >> 5
			value := vec.Vec_IntEntry(vFanins, wordIndex) | int32(1)<<uint(bitIndex&31)
			vec.Vec_IntWriteEntry(vFanins, wordIndex, value)
		} else if str[i] != '0' {
			fmt.Printf("Wrongn symbol (%c) in binary Verilog constant \"%s\".\n", str[i], str)
		}
	}
}
func Ndr_NtkPrintNodes(pNtk *Wlc_Ntk_t) {
	fmt.Printf("Node IDs and their fanins:\n")
	for id := int32(1); id < Wlc_NtkObjNumMax(pNtk); id++ {
		obj := Wlc_NtkObj(pNtk, id)
		fmt.Printf("%5d = ", id)
		faninCount := Wlc_ObjFaninNum(obj)
		for k := int32(0); k < faninCount; k++ {
			fmt.Printf("%5d ", Wlc_ObjFaninId(obj, k))
		}
		for k := faninCount; k < 4; k++ {
			fmt.Printf("      ")
		}
		fmt.Printf("    Name Id %d ", Wlc_ObjNameId(pNtk, id))
		if Wlc_ObjIsPi(obj) != 0 {
			fmt.Printf("  pi  ")
		}
		if Wlc_ObjIsPo(obj) != 0 {
			fmt.Printf("  po  ")
		}
		fmt.Printf("\n")
	}
}
func Wlc_NtkCheckIntegrity(p *miniaig.Ndr_Data_t) {
	const module = 2
	vMap := vec.Vec_IntAlloc(100)
	defer vec.Vec_IntFree(vMap)

	for obj := module + 1; obj < miniaig.Ndr_DataEnd(p, module); obj += miniaig.Ndr_DataSize(p, obj) {
		if miniaig.Ndr_DataType(p, obj) != miniaig.NDR_OBJECT {
			continue
		}
		nameID := miniaig.Ndr_ObjReadBody(p, obj, miniaig.NDR_OUTPUT)
		if nameID == -1 {
			typ := miniaig.Ndr_ObjReadBody(p, obj, miniaig.NDR_OPERTYPE)
			if typ != miniaig.ABC_OPER_CO {
				fmt.Printf("Internal object %d of type %s has no output name.\n", obj, miniaig.Abc_OperName(typ))
			}
			continue
		}
		if vec.Vec_IntGetEntry(vMap, nameID) > 0 {
			fmt.Printf("Output name %d is used more than once (obj %d and obj %d).\n", nameID, vec.Vec_IntGetEntry(vMap, nameID), obj)
		}
		vec.Vec_IntSetEntry(vMap, nameID, int32(obj))
	}

	for obj := module + 1; obj < miniaig.Ndr_DataEnd(p, module); obj += miniaig.Ndr_DataSize(p, obj) {
		if miniaig.Ndr_DataType(p, obj) != miniaig.NDR_OBJECT {
			continue
		}
		typ := miniaig.Ndr_ObjReadBody(p, obj, miniaig.NDR_OPERTYPE)
		var inputs []int
		miniaig.Ndr_ObjReadArray(p, obj, miniaig.NDR_INPUT, &inputs)
		for i, nameID := range inputs {
			if vec.Vec_IntGetEntry(vMap, nameID) == 0 && !(typ == miniaig.ABC_OPER_DFFRSE && i >= 5 && i <= 7) {
				fmt.Printf("Input name %d appearing as fanin %d of obj %d is not used as output name in any object.\n", nameID, i, obj)
			}
		}
	}
}
func Wlc_NtkFromNdr(p *miniaig.Ndr_Data_t) *Wlc_Ntk_t {
	const module = 2
	vFanins := vec.Vec_IntAlloc(100)
	pNtk := Wlc_NtkAlloc("top", int32(miniaig.Ndr_DataObjNum(p, module)+1))
	Wlc_NtkCheckIntegrity(p)
	vec.Vec_IntClear(&pNtk.vFfs)
	Wlc_NtkCleanNameId(pNtk)

	for objID := module + 1; objID < miniaig.Ndr_DataEnd(p, module); objID += miniaig.Ndr_DataSize(p, objID) {
		if miniaig.Ndr_DataType(p, objID) != miniaig.NDR_OBJECT || !miniaig.Ndr_ObjIsType(p, objID, miniaig.ABC_OPER_CI) {
			continue
		}
		var end, beg int32
		signed := Ndr_ObjReadRange(p, objID, &end, &beg)
		iObj := Wlc_ObjAlloc(pNtk, int32(WLC_OBJ_PI), signed, end, beg)
		nameID := miniaig.Ndr_ObjReadBody(p, objID, miniaig.NDR_OUTPUT)
		Wlc_ObjSetNameId(pNtk, iObj, int32(nameID))
	}

	var vTruths *vec.Vec_Wrd_t
	for objID := module + 1; objID < miniaig.Ndr_DataEnd(p, module); objID += miniaig.Ndr_DataSize(p, objID) {
		if miniaig.Ndr_DataType(p, objID) != miniaig.NDR_OBJECT || miniaig.Ndr_ObjIsType(p, objID, miniaig.ABC_OPER_CI) || miniaig.Ndr_ObjIsType(p, objID, miniaig.ABC_OPER_CO) {
			continue
		}

		var end, beg int32
		signed := Ndr_ObjReadRange(p, objID, &end, &beg)
		typ := miniaig.Ndr_ObjReadBody(p, objID, miniaig.NDR_OPERTYPE)
		var inputNames []int
		miniaig.Ndr_ObjReadArray(p, objID, miniaig.NDR_INPUT, &inputNames)
		vec.Vec_IntClear(vFanins)
		for _, nameID := range inputNames {
			vec.Vec_IntPush(vFanins, int32(nameID))
		}
		iObj := Wlc_ObjAlloc(pNtk, Ndr_TypeNdr2Wlc(int32(typ)), signed, end, beg)
		nameID := miniaig.Ndr_ObjReadBody(p, objID, miniaig.NDR_OUTPUT)

		if typ == miniaig.ABC_OPER_DFF {
			if pNtk.vInits == nil {
				pNtk.vInits = vec.Vec_IntAlloc(100)
			}
			if vec.Vec_IntSize(vFanins) == 2 {
				vec.Vec_IntPush(pNtk.vInits, vec.Vec_IntPop(vFanins))
			} else {
				vec.Vec_IntPush(pNtk.vInits, -(end - beg + 1))
			}
			obj := Wlc_NtkObj(pNtk, iObj)
			assert(Wlc_ObjType(obj) == int32(WLC_OBJ_FO), "Wlc_ObjType(obj) == int32(WLC_OBJ_FO)")
			Wlc_ObjSetNameId(pNtk, iObj, int32(nameID))
			vec.Vec_IntPush(&pNtk.vFfs, int32(nameID))
			assert(vec.Vec_IntSize(vFanins) == 1, "Vec_IntSize(vFanins) == 1")
			vec.Vec_IntPush(&pNtk.vFfs, vec.Vec_IntEntry(vFanins, 0))
			continue
		}

		if typ == miniaig.ABC_OPER_DFFRSE {
			vec.Vec_IntPush(&pNtk.vFfs2, iObj)
		}
		if typ == miniaig.ABC_OPER_LUT {
			if vTruths == nil {
				vTruths = vec.Vec_WrdStart(1000)
			}
			if int(nameID) >= vec.Vec_WrdSize(vTruths) {
				vec.Vec_WrdFillExtra(vTruths, 2*nameID, 0)
			}
			truthWords := miniaig.Ndr_ObjReadBodyP(p, objID, miniaig.NDR_FUNCTION)
			var truth uint64
			if len(truthWords) > 0 {
				truth = uint64(uint32(truthWords[0]))
			}
			if len(truthWords) > 1 {
				truth |= uint64(uint32(truthWords[1])) << 32
			}
			vec.Vec_WrdWriteEntry(vTruths, nameID, truth)
		}

		switch {
		case typ == miniaig.ABC_OPER_SLICE:
			vec.Vec_IntPushTwo(vFanins, end, beg)
		case typ == miniaig.ABC_OPER_CONST:
			Ndr_ObjReadConstant(vFanins, ndrFunction(p, objID))
		case typ == miniaig.ABC_OPER_BIT_MUX && vec.Vec_IntSize(vFanins) == 3:
			first := vec.Vec_IntEntry(vFanins, 1)
			vec.Vec_IntWriteEntry(vFanins, 1, vec.Vec_IntEntry(vFanins, 2))
			vec.Vec_IntWriteEntry(vFanins, 2, first)
		}
		obj := Wlc_NtkObj(pNtk, iObj)
		Wlc_ObjAddFanins(pNtk, obj, vFanins)
		Wlc_ObjSetNameId(pNtk, iObj, int32(nameID))
		if typ == miniaig.ABC_OPER_ARI_SMUL {
			assert(Wlc_ObjFaninNum(obj) == 2, "Wlc_ObjFaninNum(obj) == 2")
			Wlc_ObjFanin0(pNtk, obj).Signed = 1
			Wlc_ObjFanin1(pNtk, obj).Signed = 1
		}
	}

	for objID := module + 1; objID < miniaig.Ndr_DataEnd(p, module); objID += miniaig.Ndr_DataSize(p, objID) {
		if miniaig.Ndr_DataType(p, objID) != miniaig.NDR_OBJECT || !miniaig.Ndr_ObjIsType(p, objID, miniaig.ABC_OPER_CO) {
			continue
		}
		var end, beg int32
		signed := Ndr_ObjReadRange(p, objID, &end, &beg)
		var inputNames []int
		nInputs := miniaig.Ndr_ObjReadArray(p, objID, miniaig.NDR_INPUT, &inputNames)
		iObj := Wlc_ObjAlloc(pNtk, int32(WLC_OBJ_BUF), signed, end, beg)
		nameID := miniaig.Ndr_ObjReadBody(p, objID, miniaig.NDR_OUTPUT)
		assert(nInputs == 1 && nameID == -1, "nInputs == 1 && nameID == -1")
		vec.Vec_IntFill(vFanins, 1, int32(inputNames[0]))
		obj := Wlc_NtkObj(pNtk, iObj)
		Wlc_ObjAddFanins(pNtk, obj, vFanins)
		Wlc_ObjSetCo(pNtk, obj, 0)
	}
	vec.Vec_IntFree(vFanins)

	nameToObj := vec.Vec_IntInvert(&pNtk.vNameIds, 0)
	for i := int32(1); i < Wlc_NtkObjNumMax(pNtk); i++ {
		obj := Wlc_NtkObj(pNtk, i)
		for k := int32(0); k < Wlc_ObjFaninNum(obj); k++ {
			faninID := Wlc_ObjFaninId(obj, k)
			Wlc_ObjFanins(obj)[k] = vec.Vec_IntEntry(nameToObj, int(faninID))
		}
	}
	if pNtk.vInits != nil {
		for i := 0; i < vec.Vec_IntSize(&pNtk.vFfs); i++ {
			nameID := vec.Vec_IntEntry(&pNtk.vFfs, i)
			vec.Vec_IntWriteEntry(&pNtk.vFfs, i, vec.Vec_IntEntry(nameToObj, int(nameID)))
		}
		for i := 0; i < vec.Vec_IntSize(pNtk.vInits); i++ {
			nameID := vec.Vec_IntEntry(pNtk.vInits, i)
			if nameID > 0 {
				vec.Vec_IntWriteEntry(pNtk.vInits, i, vec.Vec_IntEntry(nameToObj, int(nameID)))
			}
		}
		assert(vec.Vec_IntSize(&pNtk.vFfs)%2 == 0, "Vec_IntSize(&pNtk.vFfs)%2 == 0")
		assert(vec.Vec_IntSize(&pNtk.vFfs) == 2*vec.Vec_IntSize(pNtk.vInits), "Vec_IntSize(&pNtk.vFfs) == 2*Vec_IntSize(pNtk.vInits)")
		for i := int32(0); i < vec.Vec_IntSize(&pNtk.vFfs); i++ {
			if i&1 != 0 {
				Wlc_ObjSetCo(pNtk, Wlc_NtkFf(pNtk, i), 1)
			}
		}
		vec.Vec_IntClear(&pNtk.vFfs)
		pNtk.pInits = Wlc_PrsConvertInitValues(pNtk)
	}
	vec.Vec_IntFree(nameToObj)

	nameIDMax := vec.Vec_IntFindMax(&pNtk.vNameIds)
	digits := util.Abc_Base10Log(uint32(nameIDMax + 1))
	pNtk.pManName = util.Abc_NamStart(nameIDMax+1, 10)
	for i := int32(1); i <= nameIDMax; i++ {
		name := fmt.Sprintf("s%0*d", digits, i)
		found := int32(0)
		nameID := util.Abc_NamStrFindOrAdd(pNtk.pManName, name, &found)
		assert(found == 0 && i == nameID, "found == 0 && i == nameID")
	}

	oldNtk := pNtk
	pNtk = Wlc_NtkDupDfs(oldNtk, 0, 1)
	Wlc_NtkFree(oldNtk)
	if vTruths != nil {
		pNtk.vLutTruths = vec.Vec_WrdStart(int(Wlc_NtkObjNumMax(pNtk)))
		for i := int32(1); i < Wlc_NtkObjNumMax(pNtk); i++ {
			obj := Wlc_NtkObj(pNtk, i)
			nameID := Wlc_ObjNameId(pNtk, i)
			if int32(obj.Type) != int32(WLC_OBJ_LUT) || nameID == 0 {
				continue
			}
			truth := vec.Vec_WrdEntry(vTruths, int(nameID))
			assert(^uint(0)>>32 != 0 || Wlc_ObjFaninNum(obj) < 6, "64-bit pointers or LUT fanin count < 6")
			vec.Vec_WrdWriteEntry(pNtk.vLutTruths, int(i), truth)
		}
		vec.Vec_WrdFree(vTruths)
	}
	pNtk.fMemPorts = 1
	pNtk.fEasyFfs = 1
	return pNtk
}

/**Function*************************************************************

  Synopsis    []

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func ndrFunction(p *miniaig.Ndr_Data_t, obj int) string {
	var data []byte
	for _, value := range miniaig.Ndr_ObjReadBodyP(p, obj, miniaig.NDR_FUNCTION) {
		var word [4]byte
		binary.LittleEndian.PutUint32(word[:], uint32(value))
		data = append(data, word[:]...)
	}
	return strings.TrimRight(string(data), "\x00")
}

func Wlc_ReadNdr(pFileName string) *Wlc_Ntk_t {
	pData := miniaig.Ndr_Read(pFileName)
	var pNtk *Wlc_Ntk_t = Wlc_NtkFromNdr(pData)
	//char * ppNames[10] = { NULL, "a", "b", "c", "d", "e", "f", "g", "h", "i" };
	//Ndr_WriteVerilog( NULL, pData, ppNames );
	//Ndr_Delete( pData );
	Abc_FrameInputNdr(Abc_FrameGetGlobalFrame(), pData)
	return pNtk
}
func Wlc_ReadNdrTest() {
    var pNtk *Wlc_Ntk_t  = Wlc_ReadNdr( "top.ndr" );
    Wlc_WriteVer( pNtk, "top.v", 0, 0 );
    Wlc_NtkFree( pNtk );
}

////////////////////////////////////////////////////////////////////////
///                       END OF FILE                                ///
////////////////////////////////////////////////////////////////////////


//ABC_NAMESPACE_IMPL_END

