package bac

import (
	"fmt"
	"github.com/glycerine/abcgo/misc/vec"
)

import "os"

/**CFile****************************************************************

  FileName    [bacPtr.c]

  SystemName  [ABC: Logic synthesis and verification system.]

  PackageName [Hierarchical word-level netlist.]

  Synopsis    [Simple interface with external tools.]

  Author      [Alan Mishchenko]

  Affiliation [UC Berkeley]

  Date        [Ver. 1.0. Started - November 29, 2014.]

  Revision    [$Id: bacPtr.c,v 1.00 2014/11/29 00:00:00 alanmi Exp $]

***********************************************************************/

//#include "base/abc/abc.h"
//#include "base/main/mainInt.h"
//#include "map/mio/mio.h"
//#include "bac.h"

//ABC_NAMESPACE_IMPL_START

////////////////////////////////////////////////////////////////////////
///                        DECLARATIONS                              ///
////////////////////////////////////////////////////////////////////////

/*
design = array containing design name (as the first entry in the array) followed by pointers to modules
module = array containing module name (as the first entry in the array) followed by pointers to 6 arrays:
         {array of input names; array of output names; array of nodes; array of boxes,
          array of floating-point input-arrival times; array of floating-point output-required times}
node   = array containing output name, followed by node type, followed by input names
box    = array containing model name, instance name, followed by pairs of formal/actual names for each port

  Comments:
  - in describing boxes
     - input formal/actual name pairs should be listed before output name pairs
     - the order of formal names should be the same as the order of inputs/outputs in the module description
     - all formal names present in the module description should be listed
     - if an input pin is not driven or an output pin has no fanout, the actual pin name is NULL
     - word-level formal name "a" is written as bit-level names (a[0]. a[1], etc) ordered LSB to MSB
     - the boxes can appear in any order (topological order is not expected)
  - in description of nodes and boxes, primitive names should be given as char*-strings ("AndT", "OrT", etc)
  - constant 0/1 nets should be driven by constant nodes having primitive names "Const0T" and "Const1T"
  - primitive modules should not be written, but the list of primitives and formal names should be provided
  - currently only "boxes" are supported (the array of "nodes" should contain no entries)
  - arrays of input-arrival/output-required times in the module description are optional
*/

// elementary gates
type Ptr_ObjType_t int32

const (
	PTR_GATE_NONE Ptr_ObjType_t = 0

	PTR_GATE_C0 Ptr_ObjType_t = PTR_GATE_NONE + 1
	// Const0T
	PTR_GATE_C1 Ptr_ObjType_t = PTR_GATE_C0 + 1
	// Const1T
	PTR_GATE_BUF Ptr_ObjType_t = PTR_GATE_C1 + 1
	// BufT
	PTR_GATE_INV Ptr_ObjType_t = PTR_GATE_BUF + 1
	// InvT
	PTR_GATE_AND Ptr_ObjType_t = PTR_GATE_INV + 1
	// AndT
	PTR_GATE_NAND Ptr_ObjType_t = PTR_GATE_AND + 1
	// NandT
	PTR_GATE_OR Ptr_ObjType_t = PTR_GATE_NAND + 1
	// OrT
	PTR_GATE_NOR Ptr_ObjType_t = PTR_GATE_OR + 1
	// NorT
	PTR_GATE_XOR Ptr_ObjType_t = PTR_GATE_NOR + 1
	// XorT
	PTR_GATE_XNOR Ptr_ObjType_t = PTR_GATE_XOR + 1
	// XnorT
	PTR_GATE_UNKNOWN Ptr_ObjType_t = PTR_GATE_XNOR + 1
)

////////////////////////////////////////////////////////////////////////
///                     FUNCTION DEFINITIONS                         ///
////////////////////////////////////////////////////////////////////////

/*
*Function*************************************************************

	Synopsis    [Free Ptr.]

	Description []

	SideEffects []

	SeeAlso     []

**********************************************************************
*/
func Bac_PtrFreeNtk(vNtk *vec.Vec_Ptr_t) {
	vec.Vec_PtrFree(vec.Vec_PtrEntry(vNtk, 1).(*vec.Vec_Ptr_t))
	vec.Vec_PtrFree(vec.Vec_PtrEntry(vNtk, 2).(*vec.Vec_Ptr_t))
	vec.Vec_VecFree(vec.Vec_PtrEntry(vNtk, 3).(*vec.Vec_Vec_t))
	vec.Vec_VecFree(vec.Vec_PtrEntry(vNtk, 4).(*vec.Vec_Vec_t))
	if vec.Vec_PtrSize(vNtk) > 5 {
		vec.Vec_FltFree(vec.Vec_PtrEntry(vNtk, 5).(*vec.Vec_Flt_t))
	}
	if vec.Vec_PtrSize(vNtk) > 6 {
		vec.Vec_FltFree(vec.Vec_PtrEntry(vNtk, 6).(*vec.Vec_Flt_t))
	}
	vec.Vec_PtrFree(vNtk)
}
func Bac_PtrFree(vDes *vec.Vec_Ptr_t) {
	var vNtk *vec.Vec_Ptr_t
	var i int32
	if !(vDes != nil) {
		return
	}
	for i = 1; int(i) < vec.Vec_PtrSize(vDes); i++ {
		vNtk = vec.Vec_PtrEntry(vDes, int(i)).(*vec.Vec_Ptr_t)
		Bac_PtrFreeNtk(vNtk)
	}
	vec.Vec_PtrFree(vDes)
}

/*
*Function*************************************************************

	Synopsis    [Count memory used by Ptr.]

	Description []

	SideEffects []

	SeeAlso     []

**********************************************************************
*/
func Bac_PtrMemoryArray(vArray *vec.Vec_Ptr_t) int32 {
	return int32(vec.Vec_PtrMemory(vArray))
}
func Bac_PtrMemoryArrayArray(vArrayArray *vec.Vec_Ptr_t) int32 {
	var vArray *vec.Vec_Ptr_t
	var i, nBytes int32
	for i = 0; int(i) < vec.Vec_PtrSize(vArrayArray); i++ {
		vArray = vec.Vec_PtrEntry(vArrayArray, int(i)).(*vec.Vec_Ptr_t)
		nBytes += Bac_PtrMemoryArray(vArray)
	}
	return nBytes
}
func Bac_PtrMemoryNtk(vNtk *vec.Vec_Ptr_t) int32 {
	var nBytes int32 = int32(vec.Vec_PtrMemory(vNtk))
	nBytes += Bac_PtrMemoryArray(vec.Vec_PtrEntry(vNtk, 1).(*vec.Vec_Ptr_t))
	nBytes += Bac_PtrMemoryArray(vec.Vec_PtrEntry(vNtk, 2).(*vec.Vec_Ptr_t))
	nBytes += Bac_PtrMemoryArrayArray(vec.Vec_PtrEntry(vNtk, 3).(*vec.Vec_Ptr_t))
	nBytes += Bac_PtrMemoryArrayArray(vec.Vec_PtrEntry(vNtk, 4).(*vec.Vec_Ptr_t))
	return nBytes
}
func Bac_PtrMemory(vDes *vec.Vec_Ptr_t) int32 {
	var vNtk *vec.Vec_Ptr_t
	var i int32
	var nBytes int32 = int32(vec.Vec_PtrMemory(vDes))
	for i = 1; int(i) < vec.Vec_PtrSize(vDes); i++ {
		vNtk = vec.Vec_PtrEntry(vDes, int(i)).(*vec.Vec_Ptr_t)
		nBytes += Bac_PtrMemoryNtk(vNtk)
	}
	return nBytes
}

/*
*Function*************************************************************

	Synopsis    [Dumping Ptr into a BLIF file.]

	Description []

	SideEffects []

	SeeAlso     []

**********************************************************************
*/
func Bac_PtrDumpSignalsBlif(pFile *os.File, vSigs *vec.Vec_Ptr_t, fSkipLastComma int32) {
	var pSig string
	var i int32
	for i = 0; int(i) < vec.Vec_PtrSize(vSigs); i++ {
		pSig = vec.Vec_PtrEntry(vSigs, int(i)).(string)
		fmt.Fprintf(pFile, " %s", pSig)
	}
}
func Bac_PtrDumpBoxBlif(pFile *os.File, vBox *vec.Vec_Ptr_t) {
	var pName string
	var i int32
	fmt.Fprintf(pFile, ".subckt")
	fmt.Fprintf(pFile, " %s", vec.Vec_PtrEntry(vBox, 0).(string))
	//fprintf( pFile, " %s", (char *)vec.Vec_PtrEntry(vBox, 1) ); // do not write intance name in BLIF
	for i = 2; int(i) < vec.Vec_PtrSize(vBox); i += 2 {
		pName = vec.Vec_PtrEntry(vBox, int(i)).(string)
		fmt.Fprintf(pFile, " %s=%s", pName, vec.Vec_PtrEntry(vBox, int(i+1)).(string))
	}
	fmt.Fprintf(pFile, "\n")
}
func Bac_PtrDumpBoxesBlif(pFile *os.File, vBoxes *vec.Vec_Ptr_t) {
	var vBox *vec.Vec_Ptr_t
	var i int32
	for i = 0; int(i) < vec.Vec_PtrSize(vBoxes); i++ {
		vBox = vec.Vec_PtrEntry(vBoxes, int(i)).(*vec.Vec_Ptr_t)
		Bac_PtrDumpBoxBlif(pFile, vBox)
	}
}
func Bac_PtrDumpModuleBlif(pFile *os.File, vNtk *vec.Vec_Ptr_t) {
	fmt.Fprintf(pFile, ".model %s\n", vec.Vec_PtrEntry(vNtk, 0).(string))
	fmt.Fprintf(pFile, ".inputs")
	Bac_PtrDumpSignalsBlif(pFile, vec.Vec_PtrEntry(vNtk, 1).(*vec.Vec_Ptr_t), 0)
	fmt.Fprintf(pFile, "\n")
	fmt.Fprintf(pFile, ".outputs")
	Bac_PtrDumpSignalsBlif(pFile, vec.Vec_PtrEntry(vNtk, 2).(*vec.Vec_Ptr_t), 1)
	fmt.Fprintf(pFile, "\n")
	assert(vec.Vec_PtrSize(vec.Vec_PtrEntry(vNtk, 3).(*vec.Vec_Ptr_t)) == 0, "no nodes; only boxes")
	Bac_PtrDumpBoxesBlif(pFile, vec.Vec_PtrEntry(vNtk, 4).(*vec.Vec_Ptr_t))
	fmt.Fprintf(pFile, ".end\n\n")
}
func Bac_PtrDumpBlif(pFileName string, vDes *vec.Vec_Ptr_t) {
	var pFile *os.File
	var vNtk *vec.Vec_Ptr_t
	var i int32
	pFile, _ = os.OpenFile(pFileName, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, 0666)
	if pFile == nil {
		fmt.Printf("Cannot open output file \"%s\".\n", pFileName)
		return
	}
	fmt.Fprintf(pFile, "// Design \"%s\" written via Ptr in ABC on %s\n\n", vec.Vec_PtrEntry(vDes, 0).(string), Extra_TimeStamp())
	for i = 1; int(i) < vec.Vec_PtrSize(vDes); i++ {
		vNtk = vec.Vec_PtrEntry(vDes, int(i)).(*vec.Vec_Ptr_t)
		Bac_PtrDumpModuleBlif(pFile, vNtk)
	}
	_ = (pFile).Close()
}

/*
*Function*************************************************************

	Synopsis    [Dumping Ptr into a Verilog file.]

	Description []

	SideEffects []

	SeeAlso     []

**********************************************************************
*/
func Bac_PtrDumpSignalsVerilog(pFile *os.File, vSigs *vec.Vec_Ptr_t, fAlwaysComma int32) {
	var pSig string
	var i int32
	for i = 0; int(i) < vec.Vec_PtrSize(vSigs); i++ {
		pSig = vec.Vec_PtrEntry(vSigs, int(i)).(string)
		comma := ""
		if fAlwaysComma != 0 || int(i) < vec.Vec_PtrSize(vSigs)-1 {
			comma = ","
		}
		fmt.Fprintf(pFile, " %s%s", pSig, comma)
	}
}
func Bac_PtrDumpBoxVerilog(pFile *os.File, vBox *vec.Vec_Ptr_t) {
	var pName string
	var i int32
	fmt.Fprintf(pFile, "  %s", vec.Vec_PtrEntry(vBox, 0).(string))
	fmt.Fprintf(pFile, " %s (", vec.Vec_PtrEntry(vBox, 1).(string))
	for i = 2; int(i) < vec.Vec_PtrSize(vBox); i += 2 {
		pName = vec.Vec_PtrEntry(vBox, int(i)).(string)
		comma := ""
		if int(i) < vec.Vec_PtrSize(vBox)-2 {
			comma = ", "
		}
		fmt.Fprintf(pFile, ".%s(%s)%s", pName, vec.Vec_PtrEntry(vBox, int(i+1)).(string), comma)
	}
	fmt.Fprintf(pFile, ");\n")
}
func Bac_PtrDumpBoxesVerilog(pFile *os.File, vBoxes *vec.Vec_Ptr_t) {
	var vBox *vec.Vec_Ptr_t
	var i int32
	for i = 0; int(i) < vec.Vec_PtrSize(vBoxes); i++ {
		vBox = vec.Vec_PtrEntry(vBoxes, int(i)).(*vec.Vec_Ptr_t)
		Bac_PtrDumpBoxVerilog(pFile, vBox)
	}
}
func Bac_PtrDumpModuleVerilog(pFile *os.File, vNtk *vec.Vec_Ptr_t) {
	fmt.Fprintf(pFile, "module %s (\n    ", vec.Vec_PtrEntry(vNtk, 0).(string))
	Bac_PtrDumpSignalsVerilog(pFile, vec.Vec_PtrEntry(vNtk, 1).(*vec.Vec_Ptr_t), 1)
	Bac_PtrDumpSignalsVerilog(pFile, vec.Vec_PtrEntry(vNtk, 2).(*vec.Vec_Ptr_t), 0)
	fmt.Fprintf(pFile, "\n  );\n")
	fmt.Fprintf(pFile, "  input")
	Bac_PtrDumpSignalsVerilog(pFile, vec.Vec_PtrEntry(vNtk, 1).(*vec.Vec_Ptr_t), 0)
	fmt.Fprintf(pFile, ";\n")
	fmt.Fprintf(pFile, "  output")
	Bac_PtrDumpSignalsVerilog(pFile, vec.Vec_PtrEntry(vNtk, 2).(*vec.Vec_Ptr_t), 0)
	fmt.Fprintf(pFile, ";\n")
	assert(vec.Vec_PtrSize(vec.Vec_PtrEntry(vNtk, 3).(*vec.Vec_Ptr_t)) == 0, "no nodes; only boxes")
	Bac_PtrDumpBoxesVerilog(pFile, vec.Vec_PtrEntry(vNtk, 4).(*vec.Vec_Ptr_t))
	fmt.Fprintf(pFile, "endmodule\n\n")
}
func Bac_PtrDumpVerilog(pFileName string, vDes *vec.Vec_Ptr_t) {
	var pFile *os.File
	var vNtk *vec.Vec_Ptr_t
	var i int32
	pFile, _ = os.OpenFile(pFileName, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, 0666)
	if pFile == nil {
		fmt.Printf("Cannot open output file \"%s\".\n", pFileName)
		return
	}
	fmt.Fprintf(pFile, "// Design \"%s\" written via Ptr in ABC on %s\n\n", vec.Vec_PtrEntry(vDes, 0).(string), Extra_TimeStamp())
	for i = 1; int(i) < vec.Vec_PtrSize(vDes); i++ {
		vNtk = vec.Vec_PtrEntry(vDes, int(i)).(*vec.Vec_Ptr_t)
		Bac_PtrDumpModuleVerilog(pFile, vNtk)
	}
	_ = (pFile).Close()
}

/*
*Function*************************************************************

	Synopsis    [Collect elementary gates from the library.]

	Description []

	SideEffects []

	SeeAlso     []

**********************************************************************
*/
func Bac_ManCollectGateNameOne(pLib *Mio_Library_t, Type Ptr_ObjType_t, Truth uint64, vGateNames *vec.Vec_Ptr_t) {
	var pGate *Mio_Gate_t = Mio_LibraryReadGateByTruth(pLib, Truth)
	if pGate != nil {
		vec.Vec_PtrWriteEntry(vGateNames, int(Type), Mio_GateReadName(pGate))
	}
}
func Bac_ManCollectGateNamesByTruth(pLib *Mio_Library_t) *vec.Vec_Ptr_t {
	uTruths6 := [3]uint64{0xAAAAAAAAAAAAAAAA, 0xCCCCCCCCCCCCCCCC, 0xF0F0F0F0F0F0F0F0}
	var vGateNames *vec.Vec_Ptr_t = vec.Vec_PtrStart(int(PTR_GATE_UNKNOWN))
	Bac_ManCollectGateNameOne(pLib, PTR_GATE_C0, 0, vGateNames)
	Bac_ManCollectGateNameOne(pLib, PTR_GATE_C1, ^uint64(0), vGateNames)
	Bac_ManCollectGateNameOne(pLib, PTR_GATE_BUF, uTruths6[0], vGateNames)
	Bac_ManCollectGateNameOne(pLib, PTR_GATE_INV, ^uTruths6[0], vGateNames)
	Bac_ManCollectGateNameOne(pLib, PTR_GATE_AND, (uTruths6[0] & uTruths6[1]), vGateNames)
	Bac_ManCollectGateNameOne(pLib, PTR_GATE_NAND, ^(uTruths6[0] & uTruths6[1]), vGateNames)
	Bac_ManCollectGateNameOne(pLib, PTR_GATE_OR, (uTruths6[0] | uTruths6[1]), vGateNames)
	Bac_ManCollectGateNameOne(pLib, PTR_GATE_NOR, ^(uTruths6[0] | uTruths6[1]), vGateNames)
	Bac_ManCollectGateNameOne(pLib, PTR_GATE_XOR, (uTruths6[0] ^ uTruths6[1]), vGateNames)
	Bac_ManCollectGateNameOne(pLib, PTR_GATE_XNOR, ^(uTruths6[0] ^ uTruths6[1]), vGateNames)
	return vGateNames
}

/*
*Function*************************************************************

	Synopsis    [This procedure transforms tech-ind Ptr into mapped Ptr.]

	Description []

	SideEffects []

	SeeAlso     []

**********************************************************************
*/
func Bac_PtrUpdateBox(vBox *vec.Vec_Ptr_t, vGatesNames *vec.Vec_Ptr_t) {
	var pGate *Mio_Gate_t
	var pPin *Mio_Pin_t
	var i int32 = 1
	var pLib *Mio_Library_t = Abc_FrameReadLibGen()
	// update gate name
	var pNameNew string
	var pName string = vec.Vec_PtrEntry(vBox, 0).(string)
	if pName == "Const0T" {
		pNameNew = vec.Vec_PtrEntry(vGatesNames, int(PTR_GATE_C0)).(string)
	} else if pName == "Const1T" {
		pNameNew = vec.Vec_PtrEntry(vGatesNames, int(PTR_GATE_C1)).(string)
	} else if pName == "BufT" {
		pNameNew = vec.Vec_PtrEntry(vGatesNames, int(PTR_GATE_BUF)).(string)
	} else if pName == "InvT" {
		pNameNew = vec.Vec_PtrEntry(vGatesNames, int(PTR_GATE_INV)).(string)
	} else if pName == "AndT" {
		pNameNew = vec.Vec_PtrEntry(vGatesNames, int(PTR_GATE_AND)).(string)
	} else if pName == "NandT" {
		pNameNew = vec.Vec_PtrEntry(vGatesNames, int(PTR_GATE_NAND)).(string)
	} else if pName == "OrT" {
		pNameNew = vec.Vec_PtrEntry(vGatesNames, int(PTR_GATE_OR)).(string)
	} else if pName == "NorT" {
		pNameNew = vec.Vec_PtrEntry(vGatesNames, int(PTR_GATE_NOR)).(string)
	} else if pName == "XorT" {
		pNameNew = vec.Vec_PtrEntry(vGatesNames, int(PTR_GATE_XOR)).(string)
	} else if pName == "XnorT" {
		pNameNew = vec.Vec_PtrEntry(vGatesNames, int(PTR_GATE_XNOR)).(string)
	} else { // user hierarchy
		return
	}
	vec.Vec_PtrWriteEntry(vBox, 0, Abc_UtilStrsav(pNameNew))
	// remove instance name
	pName = vec.Vec_PtrEntry(vBox, 1).(string)
	vec.Vec_PtrWriteEntry(vBox, 1, nil)
	// update formal input names
	pGate = Mio_LibraryReadGateByName(pLib, pNameNew, "")
	// inlining of macro Mio_GateForEachPin
	for pPin = Mio_GateReadPins(pGate); pPin != nil; pPin = Mio_PinReadNext(pPin) {
		pName = vec.Vec_PtrEntry(vBox, int(2*i)).(string)
		pNameNew = Mio_PinReadName(pPin)
		vec.Vec_PtrWriteEntry(vBox, int(2*i), Abc_UtilStrsav(pNameNew))
		i++
	}
	// update output name
	pName = vec.Vec_PtrEntry(vBox, int(2*i)).(string)
	pNameNew = Mio_GateReadOutName(pGate)
	vec.Vec_PtrWriteEntry(vBox, int(2*i), Abc_UtilStrsav(pNameNew))
	i++
	assert(int(2*i) == vec.Vec_PtrSize(vBox), "2 * i == vec.Vec_PtrSize(vBox)")
}
func Bac_PtrTransformSigs(vSig *vec.Vec_Ptr_t) *vec.Vec_Ptr_t {
	var pName string
	var i int32
	var vNew *vec.Vec_Ptr_t = vec.Vec_PtrAllocExact(vec.Vec_PtrSize(vSig))
	for i = 0; int(i) < vec.Vec_PtrSize(vSig); i++ {
		pName = vec.Vec_PtrEntry(vSig, int(i)).(string)
		vec.Vec_PtrPush(vNew, Abc_UtilStrsav(pName))
	}
	return vNew
}
func Bac_PtrTransformBox(vBox *vec.Vec_Ptr_t, vGatesNames *vec.Vec_Ptr_t) *vec.Vec_Ptr_t {
	var pName string
	var i int32
	var vNew *vec.Vec_Ptr_t = vec.Vec_PtrAllocExact(vec.Vec_PtrSize(vBox))
	for i = 0; int(i) < vec.Vec_PtrSize(vBox); i++ {
		pName = vec.Vec_PtrEntry(vBox, int(i)).(string)
		vec.Vec_PtrPush(vNew, Abc_UtilStrsav(pName))
	}
	if vGatesNames != nil {
		Bac_PtrUpdateBox(vNew, vGatesNames)
	}
	return vNew
}
func Bac_PtrTransformBoxes(vBoxes *vec.Vec_Ptr_t, vGatesNames *vec.Vec_Ptr_t) *vec.Vec_Ptr_t {
	var vBox *vec.Vec_Ptr_t
	var i int32
	var vNew *vec.Vec_Ptr_t = vec.Vec_PtrAllocExact(vec.Vec_PtrSize(vBoxes))
	for i = 0; int(i) < vec.Vec_PtrSize(vBoxes); i++ {
		vBox = vec.Vec_PtrEntry(vBoxes, int(i)).(*vec.Vec_Ptr_t)
		vec.Vec_PtrPush(vNew, Bac_PtrTransformBox(vBox, vGatesNames))
	}
	return vNew
}
func Bac_PtrTransformNtk(vNtk *vec.Vec_Ptr_t, vGatesNames *vec.Vec_Ptr_t) *vec.Vec_Ptr_t {
	var pName string = vec.Vec_PtrEntry(vNtk, 0).(string)
	var vInputs *vec.Vec_Ptr_t = vec.Vec_PtrEntry(vNtk, 1).(*vec.Vec_Ptr_t)
	var vOutputs *vec.Vec_Ptr_t = vec.Vec_PtrEntry(vNtk, 2).(*vec.Vec_Ptr_t)
	var vBoxes *vec.Vec_Ptr_t = vec.Vec_PtrEntry(vNtk, 4).(*vec.Vec_Ptr_t)
	var vNew *vec.Vec_Ptr_t = vec.Vec_PtrAllocExact(vec.Vec_PtrSize(vNtk))
	vec.Vec_PtrPush(vNew, Abc_UtilStrsav(pName))
	vec.Vec_PtrPush(vNew, Bac_PtrTransformSigs(vInputs))
	vec.Vec_PtrPush(vNew, Bac_PtrTransformSigs(vOutputs))
	vec.Vec_PtrPush(vNew, vec.Vec_PtrAllocExact(0))
	vec.Vec_PtrPush(vNew, Bac_PtrTransformBoxes(vBoxes, vGatesNames))
	return vNew
}
func Bac_PtrTransformTest(vDes *vec.Vec_Ptr_t) *vec.Vec_Ptr_t {
	var pLib *Mio_Library_t
	var vGatesNames *vec.Vec_Ptr_t
	var vNtk, vNew *vec.Vec_Ptr_t
	var i int32
	// dump BLIF before transformation
	Bac_PtrDumpBlif("test1.blif", vDes)
	if Abc_FrameGetGlobalFrame() == nil {
		fmt.Printf("ABC framework is not started.\n")
		return nil
	}
	pLib = Abc_FrameReadLibGen()
	if pLib == nil {
		fmt.Printf("Standard cell library is not entered.\n")
		return nil
	}
	vGatesNames = Bac_ManCollectGateNamesByTruth(pLib)
	// transform
	vNew = vec.Vec_PtrAllocExact(vec.Vec_PtrSize(vDes))
	vec.Vec_PtrPush(vNew, Abc_UtilStrsav(vec.Vec_PtrEntry(vDes, 0).(string)))
	for i = 1; int(i) < vec.Vec_PtrSize(vDes); i++ {
		vNtk = vec.Vec_PtrEntry(vDes, int(i)).(*vec.Vec_Ptr_t)
		vec.Vec_PtrPush(vNew, Bac_PtrTransformNtk(vNtk, vGatesNames))
	}
	// dump BLIF after transformation
	Bac_PtrDumpBlif("test2.blif", vNew)
	vec.Vec_PtrFree(vGatesNames)
	return vNew
}

/*
*Function*************************************************************

	Synopsis    [Test the testing procedure.]

	Description []

	SideEffects []

	SeeAlso     []

**********************************************************************
*/
func Bac_PtrTransformTestTest() {
	var pFileName string = "c/hie/dump/1/netlist_1.v"
	var pNtk *Abc_Ntk_t = Io_ReadNetlist(pFileName, Io_ReadFileType(pFileName), 0)
	var vDes *vec.Vec_Ptr_t = Ptr_AbcDeriveDes(pNtk)
	var vNew *vec.Vec_Ptr_t = Bac_PtrTransformTest(vDes)
	Bac_PtrFree(vDes)
	Bac_PtrFree(vNew)
}

////////////////////////////////////////////////////////////////////////
///                       END OF FILE                                ///
////////////////////////////////////////////////////////////////////////

//ABC_NAMESPACE_IMPL_END
