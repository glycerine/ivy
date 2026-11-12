package bac

import (
	"fmt"
	"strings"
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

/**Function*************************************************************

  Synopsis    [Free Ptr.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Bac_PtrFreeNtk(vNtk *Vec_Ptr_t) {
    Vec_PtrFree( (Vec_Ptr_t *)Vec_PtrEntry(vNtk, 1) );
    Vec_PtrFree( (Vec_Ptr_t *)Vec_PtrEntry(vNtk, 2) );
    Vec_VecFree( (Vec_Vec_t *)Vec_PtrEntry(vNtk, 3) );
    Vec_VecFree( (Vec_Vec_t *)Vec_PtrEntry(vNtk, 4) );
    if Vec_PtrSize(vNtk) > 5 {
        Vec_FltFree( (Vec_Flt_t *)Vec_PtrEntry(vNtk, 5) ); }
    if Vec_PtrSize(vNtk) > 6 {
        Vec_FltFree( (Vec_Flt_t *)Vec_PtrEntry(vNtk, 6) ); }
    Vec_PtrFree( vNtk );
}
func Bac_PtrFree(vDes *Vec_Ptr_t) {
    var vNtk *Vec_Ptr_t; var i int32;
    if !(vDes != nil) { return; }
    Vec_PtrForEachEntryStart( Vec_Ptr_t *, vDes, vNtk, i, 1 )
        Bac_PtrFreeNtk( vNtk );
    Vec_PtrFree( vDes );
}

/**Function*************************************************************

  Synopsis    [Count memory used by Ptr.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Bac_PtrMemoryArray(vArray *Vec_Ptr_t) int32 {
    return int32(Vec_PtrMemory(vArray));
}
func Bac_PtrMemoryArrayArray(vArrayArray *Vec_Ptr_t) int32 {
    var vArray *Vec_Ptr_t; var i, nBytes int32;
    Vec_PtrForEachEntry( Vec_Ptr_t *, vArrayArray, vArray, i )
        nBytes += Bac_PtrMemoryArray(vArray);
    return nBytes;
}
func Bac_PtrMemoryNtk(vNtk *Vec_Ptr_t) int32 {
    var nBytes int32  = int32(Vec_PtrMemory(vNtk));
    nBytes += Bac_PtrMemoryArray( (Vec_Ptr_t *)Vec_PtrEntry(vNtk, 1) );
    nBytes += Bac_PtrMemoryArray( (Vec_Ptr_t *)Vec_PtrEntry(vNtk, 2) );
    nBytes += Bac_PtrMemoryArrayArray( (Vec_Ptr_t *)Vec_PtrEntry(vNtk, 3) );
    nBytes += Bac_PtrMemoryArrayArray( (Vec_Ptr_t *)Vec_PtrEntry(vNtk, 4) );
    return nBytes;
}
func Bac_PtrMemory(vDes *Vec_Ptr_t) int32 {
    var vNtk *Vec_Ptr_t; var i int32;
var nBytes int32 = int32(Vec_PtrMemory(vDes));
    Vec_PtrForEachEntryStart( Vec_Ptr_t *, vDes, vNtk, i, 1 )
        nBytes += Bac_PtrMemoryNtk(vNtk);
    return nBytes;
}

/**Function*************************************************************

  Synopsis    [Dumping Ptr into a BLIF file.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Bac_PtrDumpSignalsBlif(pFile *os.File, vSigs *Vec_Ptr_t, fSkipLastComma int32) {
    var pSig string; var i int32;
    Vec_PtrForEachEntry( char *, vSigs, pSig, i )
        fmt.Fprintf(pFile, " %s", pSig);
}
func Bac_PtrDumpBoxBlif(pFile *os.File, vBox *Vec_Ptr_t) {
    var pName string; var i int32;
    fmt.Fprintf(pFile, ".subckt");
    fmt.Fprintf(pFile, " %s", (char *)Vec_PtrEntry(vBox, 0));
    //fprintf( pFile, " %s", (char *)Vec_PtrEntry(vBox, 1) ); // do not write intance name in BLIF
    Vec_PtrForEachEntryStart( char *, vBox, pName, i, 2 )
        fmt.Fprintf(pFile, " %s=%s", pName, (char *)Vec_PtrEntry(vBox, i+1));
i++;
    fmt.Fprintf(pFile, "\n");
}
func Bac_PtrDumpBoxesBlif(pFile *os.File, vBoxes *Vec_Ptr_t) {
    var vBox *Vec_Ptr_t; var i int32;
    Vec_PtrForEachEntry( Vec_Ptr_t *, vBoxes, vBox, i )
        Bac_PtrDumpBoxBlif( pFile, vBox );
}
func Bac_PtrDumpModuleBlif(pFile *os.File, vNtk *Vec_Ptr_t) {
    fmt.Fprintf(pFile, ".model %s\n", (char *)Vec_PtrEntry(vNtk, 0));
    fmt.Fprintf(pFile, ".inputs");
    Bac_PtrDumpSignalsBlif( pFile, (Vec_Ptr_t *)Vec_PtrEntry(vNtk, 1), 0 );
    fmt.Fprintf(pFile, "\n");
    fmt.Fprintf(pFile, ".outputs");
    Bac_PtrDumpSignalsBlif( pFile, (Vec_Ptr_t *)Vec_PtrEntry(vNtk, 2), 1 );
    fmt.Fprintf(pFile, "\n");
    assert(Vec_PtrSize((Vec_Ptr_t *)Vec_PtrEntry(vNtk, 3)) == 0, "Vec_PtrSize((Vec_Ptr_t *)Vec_PtrEntry(vNtk, 3)) == 0"); // no nodes; only boxes
    Bac_PtrDumpBoxesBlif( pFile, (Vec_Ptr_t *)Vec_PtrEntry(vNtk, 4) );
    fmt.Fprintf(pFile, ".end\n\n");
}
func Bac_PtrDumpBlif(pFileName string, vDes *Vec_Ptr_t) {
    var pFile *os.File;
    var vNtk *Vec_Ptr_t; var i int32;
    pFile = os.OpenFile(pFileName, os.O_WRONLY | os.O_CREATE | os.O_TRUNC, 0666);
    if pFile == nil {
        fmt.Printf("Cannot open output file \"%s\".\n", pFileName);
        return;
    }
    fmt.Fprintf(pFile, "// Design \"%s\" written via Ptr in ABC on %s\n\n", (char *)Vec_PtrEntry(vDes, 0), Extra_TimeStamp());
    Vec_PtrForEachEntryStart( Vec_Ptr_t *, vDes, vNtk, i, 1 )
        Bac_PtrDumpModuleBlif( pFile, vNtk );
    _ = (pFile).Close();
}

/**Function*************************************************************

  Synopsis    [Dumping Ptr into a Verilog file.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Bac_PtrDumpSignalsVerilog(pFile *os.File, vSigs *Vec_Ptr_t, fAlwaysComma int32) {
    var pSig string; var i int32;
    Vec_PtrForEachEntry( char *, vSigs, pSig, i )
        fmt.Fprintf(pFile, " %s%s", pSig, (fAlwaysComma || i < Vec_PtrSize(vSigs) - 1) ? ",":"");
}
func Bac_PtrDumpBoxVerilog(pFile *os.File, vBox *Vec_Ptr_t) {
    var pName string; var i int32;
    fmt.Fprintf(pFile, "  %s", (char *)Vec_PtrEntry(vBox, 0));
    fmt.Fprintf(pFile, " %s (", (char *)Vec_PtrEntry(vBox, 1)); // write intance name in Verilog
    Vec_PtrForEachEntryStart( char *, vBox, pName, i, 2 )
        fmt.Fprintf(pFile, ".%s(%s)%s", pName, (char *)Vec_PtrEntry(vBox, i+1), i < Vec_PtrSize(vBox) - 2 ? ", ":"");
i++;
    fmt.Fprintf(pFile, ");\n");
}
func Bac_PtrDumpBoxesVerilog(pFile *os.File, vBoxes *Vec_Ptr_t) {
    var vBox *Vec_Ptr_t; var i int32;
    Vec_PtrForEachEntry( Vec_Ptr_t *, vBoxes, vBox, i )
        Bac_PtrDumpBoxVerilog( pFile, vBox );
}
func Bac_PtrDumpModuleVerilog(pFile *os.File, vNtk *Vec_Ptr_t) {
    fmt.Fprintf(pFile, "module %s (\n    ", (char *)Vec_PtrEntry(vNtk, 0));
    Bac_PtrDumpSignalsVerilog( pFile, (Vec_Ptr_t *)Vec_PtrEntry(vNtk, 1), 1 );
    Bac_PtrDumpSignalsVerilog( pFile, (Vec_Ptr_t *)Vec_PtrEntry(vNtk, 2), 0 );
    fmt.Fprintf(pFile, "\n  );\n");
    fmt.Fprintf(pFile, "  input");
    Bac_PtrDumpSignalsVerilog( pFile, (Vec_Ptr_t *)Vec_PtrEntry(vNtk, 1), 0 );
    fmt.Fprintf(pFile, ";\n");
    fmt.Fprintf(pFile, "  output");
    Bac_PtrDumpSignalsVerilog( pFile, (Vec_Ptr_t *)Vec_PtrEntry(vNtk, 2), 0 );
    fmt.Fprintf(pFile, ";\n");
    assert(Vec_PtrSize((Vec_Ptr_t *)Vec_PtrEntry(vNtk, 3)) == 0, "Vec_PtrSize((Vec_Ptr_t *)Vec_PtrEntry(vNtk, 3)) == 0"); // no nodes; only boxes
    Bac_PtrDumpBoxesVerilog( pFile, (Vec_Ptr_t *)Vec_PtrEntry(vNtk, 4) );
    fmt.Fprintf(pFile, "endmodule\n\n");
}
func Bac_PtrDumpVerilog(pFileName string, vDes *Vec_Ptr_t) {
    var pFile *os.File;
    var vNtk *Vec_Ptr_t; var i int32;
    pFile = os.OpenFile(pFileName, os.O_WRONLY | os.O_CREATE | os.O_TRUNC, 0666);
    if pFile == nil {
        fmt.Printf("Cannot open output file \"%s\".\n", pFileName);
        return;
    }
    fmt.Fprintf(pFile, "// Design \"%s\" written via Ptr in ABC on %s\n\n", (char *)Vec_PtrEntry(vDes, 0), Extra_TimeStamp());
    Vec_PtrForEachEntryStart( Vec_Ptr_t *, vDes, vNtk, i, 1 )
        Bac_PtrDumpModuleVerilog( pFile, vNtk );
    _ = (pFile).Close();
}


/**Function*************************************************************

  Synopsis    [Collect elementary gates from the library.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Bac_ManCollectGateNameOne(pLib *Mio_Library_t, Type Ptr_ObjType_t, Truth word, vGateNames *Vec_Ptr_t) {
    var pGate *Mio_Gate_t  = Mio_LibraryReadGateByTruth( pLib, Truth );
    if pGate != nil {
        Vec_PtrWriteEntry( vGateNames, Type, Mio_GateReadName(pGate) ); }
}
func Bac_ManCollectGateNamesByTruth(pLib *Mio_Library_t) *Vec_Ptr_t {
    static word uTruths6[3] = {
        ABC_CONST(0xAAAAAAAAAAAAAAAA),
        ABC_CONST(0xCCCCCCCCCCCCCCCC),
        ABC_CONST(0xF0F0F0F0F0F0F0F0),
    };
    var vGateNames *Vec_Ptr_t  = Vec_PtrStart( PTR_GATE_UNKNOWN );
    Bac_ManCollectGateNameOne( pLib, PTR_GATE_C0,              0,                 vGateNames );
    Bac_ManCollectGateNameOne( pLib, PTR_GATE_C1,       ~(word)0,                 vGateNames );
    Bac_ManCollectGateNameOne( pLib, PTR_GATE_BUF,    uTruths6[0],                vGateNames );
    Bac_ManCollectGateNameOne( pLib, PTR_GATE_INV,   ~uTruths6[0],                vGateNames );
    Bac_ManCollectGateNameOne( pLib, PTR_GATE_AND,   (uTruths6[0] & uTruths6[1]), vGateNames );
    Bac_ManCollectGateNameOne( pLib, PTR_GATE_NAND, ~(uTruths6[0] & uTruths6[1]), vGateNames );
    Bac_ManCollectGateNameOne( pLib, PTR_GATE_OR,    (uTruths6[0] | uTruths6[1]), vGateNames );
    Bac_ManCollectGateNameOne( pLib, PTR_GATE_NOR,  ~(uTruths6[0] | uTruths6[1]), vGateNames );
    Bac_ManCollectGateNameOne( pLib, PTR_GATE_XOR,   (uTruths6[0] ^ uTruths6[1]), vGateNames );
    Bac_ManCollectGateNameOne( pLib, PTR_GATE_XNOR, ~(uTruths6[0] ^ uTruths6[1]), vGateNames );
    return vGateNames;
}

/**Function*************************************************************

  Synopsis    [This procedure transforms tech-ind Ptr into mapped Ptr.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Bac_PtrUpdateBox(vBox *Vec_Ptr_t, vGatesNames *Vec_Ptr_t) {
    var pGate *Mio_Gate_t;  var pPin *Mio_Pin_t; var i int32  = 1;
    var pLib *Mio_Library_t  = (Mio_Library_t *)Abc_FrameReadLibGen();
    // update gate name
    var pNameNew string;
    var pName string = (char *)Vec_PtrEntry(vBox, 0);
    if !(int32(strings.Compare(pName, "Const0T")) != 0) {
        pNameNew = (char *)Vec_PtrEntry(vGatesNames, PTR_GATE_C0);
    } else if !(int32(strings.Compare(pName, "Const1T")) != 0) {
        pNameNew = (char *)Vec_PtrEntry(vGatesNames, PTR_GATE_C1);
    } else if !(int32(strings.Compare(pName, "BufT")) != 0) {
        pNameNew = (char *)Vec_PtrEntry(vGatesNames, PTR_GATE_BUF);
    } else if !(int32(strings.Compare(pName, "InvT")) != 0) {
        pNameNew = (char *)Vec_PtrEntry(vGatesNames, PTR_GATE_INV);
    } else if !(int32(strings.Compare(pName, "AndT")) != 0) {
        pNameNew = (char *)Vec_PtrEntry(vGatesNames, PTR_GATE_AND);
    } else if !(int32(strings.Compare(pName, "NandT")) != 0) {
        pNameNew = (char *)Vec_PtrEntry(vGatesNames, PTR_GATE_NAND);
    } else if !(int32(strings.Compare(pName, "OrT")) != 0) {
        pNameNew = (char *)Vec_PtrEntry(vGatesNames, PTR_GATE_OR);
    } else if !(int32(strings.Compare(pName, "NorT")) != 0) {
        pNameNew = (char *)Vec_PtrEntry(vGatesNames, PTR_GATE_NOR);
    } else if !(int32(strings.Compare(pName, "XorT")) != 0) {
        pNameNew = (char *)Vec_PtrEntry(vGatesNames, PTR_GATE_XOR);
    } else if !(int32(strings.Compare(pName, "XnorT")) != 0) {
        pNameNew = (char *)Vec_PtrEntry(vGatesNames, PTR_GATE_XNOR);
    } else { // user hierarchy
        return; }
    ABC_FREE( pName );
    Vec_PtrWriteEntry( vBox, 0, Abc_UtilStrsav(pNameNew) );
    // remove instance name
    pName = (char *)Vec_PtrEntry(vBox, 1);
    ABC_FREE( pName );
    Vec_PtrWriteEntry( vBox, 1, nil );
    // update formal input names
    pGate = Mio_LibraryReadGateByName( pLib, pNameNew, nil );
    // inlining of macro Mio_GateForEachPin
    for  pPin = Mio_GateReadPins(pGate);                                 (pPin != nil);                                                          pPin = Mio_PinReadNext(pPin)  {
        pName = (char *)Vec_PtrEntry( vBox, 2 * i );
        ABC_FREE( pName );
        pNameNew = Mio_PinReadName(pPin);
        Vec_PtrWriteEntry( vBox, 2 * i++, Abc_UtilStrsav(pNameNew) );
    }
    // update output name
    pName = (char *)Vec_PtrEntry( vBox, 2 * i );
    pNameNew = Mio_GateReadOutName(pGate);
    Vec_PtrWriteEntry( vBox, 2 * i++, Abc_UtilStrsav(pNameNew) );
    assert(2 * i == Vec_PtrSize(vBox), "2 * i == Vec_PtrSize(vBox)");
}
func Bac_PtrTransformSigs(vSig *Vec_Ptr_t) *Vec_Ptr_t {
    var pName string; var i int32;
    var vNew *Vec_Ptr_t  = Vec_PtrAllocExact( Vec_PtrSize(vSig) );
    Vec_PtrForEachEntry( char *, vSig, pName, i )
        Vec_PtrPush( vNew, Abc_UtilStrsav(pName) );
    return vNew;
}
func Bac_PtrTransformBox(vBox *Vec_Ptr_t, vGatesNames *Vec_Ptr_t) *Vec_Ptr_t {
    var pName string; var i int32;
    var vNew *Vec_Ptr_t  = Vec_PtrAllocExact( Vec_PtrSize(vBox) );
    Vec_PtrForEachEntry( char *, vBox, pName, i )
        Vec_PtrPush( vNew, Abc_UtilStrsav(pName) );
    if vGatesNames != nil {
        Bac_PtrUpdateBox( vNew, vGatesNames ); }
    return vNew;
}
func Bac_PtrTransformBoxes(vBoxes *Vec_Ptr_t, vGatesNames *Vec_Ptr_t) *Vec_Ptr_t {
    var vBox *Vec_Ptr_t; var i int32;
    var vNew *Vec_Ptr_t  = Vec_PtrAllocExact( Vec_PtrSize(vBoxes) );
    Vec_PtrForEachEntry( Vec_Ptr_t *, vBoxes, vBox, i )
        Vec_PtrPush( vNew, Bac_PtrTransformBox(vBox, vGatesNames) );
    return vNew;
}
func Bac_PtrTransformNtk(vNtk *Vec_Ptr_t, vGatesNames *Vec_Ptr_t) *Vec_Ptr_t {
    var pName string  = (char *)Vec_PtrEntry(vNtk, 0);
    var vInputs *Vec_Ptr_t   = (Vec_Ptr_t *)Vec_PtrEntry(vNtk, 1);
    var vOutputs *Vec_Ptr_t  = (Vec_Ptr_t *)Vec_PtrEntry(vNtk, 2);
    var vBoxes *Vec_Ptr_t    = (Vec_Ptr_t *)Vec_PtrEntry(vNtk, 4);
    var vNew *Vec_Ptr_t      = Vec_PtrAllocExact( Vec_PtrSize(vNtk) );
    Vec_PtrPush( vNew, Abc_UtilStrsav(pName) );
    Vec_PtrPush( vNew, Bac_PtrTransformSigs(vInputs) );
    Vec_PtrPush( vNew, Bac_PtrTransformSigs(vOutputs) );
    Vec_PtrPush( vNew, Vec_PtrAllocExact(0) );
    Vec_PtrPush( vNew, Bac_PtrTransformBoxes(vBoxes, vGatesNames) );
    return vNew;
}
func Bac_PtrTransformTest(vDes *Vec_Ptr_t) *Vec_Ptr_t {
    var pLib *Mio_Library_t;
    var vGatesNames *Vec_Ptr_t;
    var vNtk, vNew *Vec_Ptr_t; var i int32;
    // dump BLIF before transformation
    Bac_PtrDumpBlif( "test1.blif", vDes );
    if Abc_FrameGetGlobalFrame() == nil {
        fmt.Printf("ABC framework is not started.\n");
        return nil;
    }
    pLib = (Mio_Library_t *)Abc_FrameReadLibGen();
    if pLib == nil {
        fmt.Printf("Standard cell library is not entered.\n");
        return nil;
    }
    vGatesNames = Bac_ManCollectGateNamesByTruth( pLib );
    // transform
    vNew = Vec_PtrAllocExact( Vec_PtrSize(vDes) );
    Vec_PtrPush( vNew, Abc_UtilStrsav((char *)Vec_PtrEntry(vDes, 0)) );
    Vec_PtrForEachEntryStart( Vec_Ptr_t *, vDes, vNtk, i, 1 )
        Vec_PtrPush( vNew, Bac_PtrTransformNtk(vNtk, vGatesNames) );
    // dump BLIF after transformation
    Bac_PtrDumpBlif( "test2.blif", vNew );
    Vec_PtrFree( vGatesNames );
    return vNew;
}

/**Function*************************************************************

  Synopsis    [Test the testing procedure.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Bac_PtrTransformTestTest() {
    var pFileName string  = "c/hie/dump/1/netlist_1.v";
    var pNtk *Abc_Ntk_t  = Io_ReadNetlist( pFileName, Io_ReadFileType(pFileName), 0 );
    extern Vec_Ptr_t * Ptr_AbcDeriveDes( Abc_Ntk_t * pNtk );
    var vDes *Vec_Ptr_t  = Ptr_AbcDeriveDes( pNtk );
    var vNew *Vec_Ptr_t  = Bac_PtrTransformTest( vDes );
    Bac_PtrFree( vDes );
    Bac_PtrFree( vNew );
}


////////////////////////////////////////////////////////////////////////
///                       END OF FILE                                ///
////////////////////////////////////////////////////////////////////////


//ABC_NAMESPACE_IMPL_END

