package miniaig

import "fmt"

import "os"

/**CFile****************************************************************

  FileName    [ndr.h]

  SystemName  [ABC: Logic synthesis and verification system.]

  PackageName [Format for word-level design representation.]

  Synopsis    [External declarations.]

  Author      [Alan Mishchenko]
  
  Affiliation [UC Berkeley]

  Date        [Ver. 1.0. Started - August 22, 2014.]

  Revision    [$Id: ndr.h,v 1.00 2014/09/12 00:00:00 alanmi Exp $]

***********************************************************************/





////////////////////////////////////////////////////////////////////////
///                          INCLUDES                                ///
////////////////////////////////////////////////////////////////////////

//#include <stdio.h>
//#include <stdlib.h>
//#include <string.h>
//#include <assert.h>

//#include "abcOper.h"

//ABC_NAMESPACE_HEADER_START 

#ifdef _WIN32
#define inline __inline
#endif

/*
    For the lack of a better name, this format is called New Data Representation (NDR).

    NDR is designed as an interchange format to pass hierarchical word-level designs between the tools.
    It is relatively simple, uses little memory, and can be easily converted into other ABC data-structures.

    This tutorial discusses how to construct the NDR representation of a hierarchical word-level design.

    First, all names used in the design (including the design name, module names, port names, net names, 
    instance names, etc) are hashed into 1-based integers called "name IDs". Nets are not explicitly represented. 
    The connectivity of a design object is established by specifying name IDs of the nets connected to the object. 
    Object inputs are name IDs of the driving nets; object outputs are name IDs of the driven nets.

    The design is initialized using procedure Ndr_Create(), which takes the design name as an argument.
    A module in the design is initialized using procedure Ndr_AddModule(), which takes the design and 
    the module name as arguments. Objects are added to a module in any order using procedure Ndr_AddObject().

    Primary input and primary output objects should be explicitly created, as shown in the examples below.

    Instances of known operators listed in file "abcOper.h" are assumed to have one output. The only known 
    issue due to this restriction concerns the adder, which produces the sum and the carry-out. To make sure the 
    adder instance has only one output, the carry-out has to be concatenated with the sum before the adder 
    instance is created in the NDR format.

    Instances of hierarchical modules defined by the user can have multiple outputs. 

    Bit-slice and concatenation operators should be represented as separate objects.

    If the ordering of inputs/outputs/flops of a module is not provided as a separate record in NDR format, 
    their ordering is determined by the order of their appearance in the NDR description of the module.

    If left limit and right limit of a bit-range are equal, it is assumed that the range contains one bit
    
    Word-level constants are represented as char-strings given in the same way as they would appear in a Verilog 
    file. For example, the 16-bit constant 10 is represented as a string "4'b1010" and is given as an argument 
    (char * pFunction) to the procedure Ndr_AddObject().

    Currently two types of flops are supported: a simple flop with implicit clock and two fanins (data and init) 
    and a complex flop with 8 fanins (clock, data, reset, set, enable, async, sre, init), as shown in the examples below.

    The initial value of a flop is represented by input "init", which can be driven by a constant or by a primary 
    input of the module. If it is a primary input, is it assumed that the flop is not initialized. If the input 
    "init" is not driven, it is assumed that the flop is initialized to 0.

    Memory read and write ports are supported, as shown in the example below.

    (to be continued)
*/

////////////////////////////////////////////////////////////////////////
///                         PARAMETERS                               ///
////////////////////////////////////////////////////////////////////////

// record types
 type Ndr_RecordType_t int32
const (
    NDR_NONE Ndr_RecordType_t = 0
          // 0:  unused
    NDR_DESIGN Ndr_RecordType_t = NDR_NONE + 1
            // 1:  design (or library of modules)
    NDR_MODULE Ndr_RecordType_t = NDR_DESIGN + 1
            // 2:  one module
    NDR_OBJECT Ndr_RecordType_t = NDR_MODULE + 1
            // 3:  object
    NDR_INPUT Ndr_RecordType_t = NDR_OBJECT + 1
             // 4:  input
    NDR_OUTPUT Ndr_RecordType_t = NDR_INPUT + 1
            // 5:  output
    NDR_OPERTYPE Ndr_RecordType_t = NDR_OUTPUT + 1
          // 6:  operator type (buffer, shifter, adder, etc)
    NDR_NAME Ndr_RecordType_t = NDR_OPERTYPE + 1
              // 7:  name
    NDR_RANGE Ndr_RecordType_t = NDR_NAME + 1
             // 8:  bit range
    NDR_FUNCTION Ndr_RecordType_t = NDR_RANGE + 1
          // 9:  specified for some operators (PLAs, etc)
    NDR_TARGET Ndr_RecordType_t = NDR_FUNCTION + 1
            // 10: target
    NDR_UNKNOWN Ndr_RecordType_t = NDR_TARGET + 1            // 11: unknown

)

  


////////////////////////////////////////////////////////////////////////
///                         BASIC TYPES                              ///
////////////////////////////////////////////////////////////////////////

// this is an internal procedure, which is not seen by the user
    
type Ndr_Data_t struct {
    nSize                     int32;
    nCap                     int32;
    unsigned char *         pHead;
    unsigned int *          pBody;
}

func Ndr_DataType(p *Ndr_Data_t, i int32) int32 { assert(p.pHead[i], "p.pHead[i]"); return int32(p.pHead[i]);                }
func Ndr_DataSize(p *Ndr_Data_t, i int32) int32 { if Ndr_DataType(p, i) > NDR_OBJECT {
    return 1;
} else {
    return p.pBody[i];
}      }
func Ndr_DataEntry(p *Ndr_Data_t, i int32) int32 { return int32(p.pBody[i]);                                       }
func Ndr_DataEntryP(p *Ndr_Data_t, i int32) *int32 { return (int *)p.pBody + i;                                    }
func Ndr_DataEnd(p *Ndr_Data_t, i int32) int32 { return i + p.pBody[i];                                        }
func Ndr_DataAddTo(p *Ndr_Data_t, i int32, Add int32) { assert(Ndr_DataType(p, i) <= NDR_OBJECT, "Ndr_DataType(p, i) <= NDR_OBJECT"); p.pBody[i] += Add;  } 
func Ndr_DataPush(p *Ndr_Data_t, Type int32, Entry int32) { p.pHead[p.nSize] = Type; p.pBody[p.nSize++] = Entry;       }

// Original macro:
// #define NDR_ALLOC(type, num)     ((type *) malloc(sizeof(type) * (size_t)(num)))


////////////////////////////////////////////////////////////////////////
///                          ITERATORS                               ///
////////////////////////////////////////////////////////////////////////

// iterates over modules in the design
// Original macro:
// #define Ndr_DesForEachMod( p, Mod )                                   \
//     for ( Mod = 1; Mod < Ndr_DataEntry(p, 0); Mod += Ndr_DataSize(p, Mod) ) if (Ndr_DataType(p, Mod) != NDR_MODULE) {} else


// iterates over objects in a module
// Original macro:
// #define Ndr_ModForEachObj( p, Mod, Obj )                              \
//     for ( Obj = Mod + 1; Obj < Ndr_DataEnd(p, Mod); Obj += Ndr_DataSize(p, Obj) ) if (Ndr_DataType(p, Obj) != NDR_OBJECT) {} else


// iterates over records in an object
// Original macro:
// #define Ndr_ObjForEachEntry( p, Obj, Ent )                            \
//     for ( Ent = Obj + 1; Ent < Ndr_DataEnd(p, Obj); Ent += Ndr_DataSize(p, Ent) )


// iterates over primary inputs of a module
// Original macro:
// #define Ndr_ModForEachPi( p, Mod, Obj )                               \
//     Ndr_ModForEachObj( p, Mod, Obj ) if ( !Ndr_ObjIsType(p, Obj, ABC_OPER_CI) ) {} else


// iteraots over primary outputs of a module
// Original macro:
// #define Ndr_ModForEachPo( p, Mod, Obj )                               \
//     Ndr_ModForEachObj( p, Mod, Obj ) if ( !Ndr_ObjIsType(p, Obj, ABC_OPER_CO) ) {} else


// iterates over internal nodes of a module
// Original macro:
// #define Ndr_ModForEachNode( p, Mod, Obj )                             \
//     Ndr_ModForEachObj( p, Mod, Obj ) if ( Ndr_ObjIsType(p, Obj, ABC_OPER_CI) || Ndr_ObjIsType(p, Obj, ABC_OPER_CO) ) {} else


// iterates over target signals of a module
// Original macro:
// #define Ndr_ModForEachTarget( p, Mod, Obj )                           \
//     for ( Obj = Mod + 1; Obj < Ndr_DataEnd(p, Mod); Obj += Ndr_DataSize(p, Obj) ) if (Ndr_DataType(p, Obj) != NDR_TARGET) {} else


////////////////////////////////////////////////////////////////////////
///                    INTERNAL PROCEDURES                           ///
////////////////////////////////////////////////////////////////////////


func Ndr_DataResize(p *Ndr_Data_t, Add int32) {
    if p.nSize + Add <= p.nCap {
        return; }
    if 2 * p.nCap > p.nSize + Add {
        p.nCap = 2 * p.nCap;
    } else {
        p.nCap = p.nSize + Add;
    }
    p.pHead = (unsigned char*)realloc( p.pHead,   p.nCap );
    p.pBody = (unsigned int *)realloc( p.pBody, 4*p.nCap );
}
func Ndr_DataPushRange(p *Ndr_Data_t, RangeLeft int32, RangeRight int32, fSignedness int32) { 
    if fSignedness != 0 {
        Ndr_DataPush( p, NDR_RANGE, RangeLeft );
        Ndr_DataPush( p, NDR_RANGE, RangeRight );
        Ndr_DataPush( p, NDR_RANGE, fSignedness );
        return;
    }
    if !(RangeLeft != 0) && !(RangeRight != 0) {
        return; }
    if RangeLeft == RangeRight {
        Ndr_DataPush( p, NDR_RANGE, RangeLeft );
    } else {
        Ndr_DataPush( p, NDR_RANGE, RangeLeft );
        Ndr_DataPush( p, NDR_RANGE, RangeRight );
    }
}
func Ndr_DataPushArray(p *Ndr_Data_t, Type int32, nArray int32, pArray *int32) { 
    if !(nArray != 0) {
        return; }
    assert(nArray > 0, "nArray > 0");
    Ndr_DataResize( p, nArray );
    memset( p.pHead + p.nSize, Type, (size_t)nArray );
    memcpy( p.pBody + p.nSize, pArray, (size_t)4*nArray );
    p.nSize += nArray;
}
func Ndr_DataPushString(p *Ndr_Data_t, ObjType int32, Type int32, pFunc string) { 
    var nBuffInts int32;
    var pBuff *int32;
    if !pFunc {
        return; }
    if ObjType == ABC_OPER_LUT {
        var Truth word  = (word)pFunc;
        Ndr_DataPushArray( p, Type, 2, (int *)&Truth );
    
    } else {
        nBuffInts = (int32(len(pFunc)) + 4) / 4;
        pBuff = (int *)calloc( 1, 4*nBuffInts );
        memcpy( pBuff, pFunc, len(pFunc) );
        Ndr_DataPushArray( p, Type, nBuffInts, pBuff );
        free( pBuff );
    }
}

////////////////////////////////////////////////////////////////////////
///                     VERILOG WRITING                              ///
////////////////////////////////////////////////////////////////////////

func Ndr_ObjReadEntry(p *Ndr_Data_t, Obj int32, Type int32) int32 {
    var Ent int32;
    // inlining of macro Ndr_ObjForEachEntry
    for  Ent = Obj + 1; Ent < Ndr_DataEnd(p, Obj); Ent += Ndr_DataSize(p, Ent)  {
        if Ndr_DataType(p, Ent) == Type {
            return Ndr_DataEntry(p, Ent); }
    }
    return -1;
}
func Ndr_ObjReadArray(p *Ndr_Data_t, Obj int32, Type int32, ppStart **int32) int32 {
    var Ent, Counter int32; *ppStart = nil;
    // inlining of macro Ndr_ObjForEachEntry
    for  Ent = Obj + 1; Ent < Ndr_DataEnd(p, Obj); Ent += Ndr_DataSize(p, Ent)  {
        if Ndr_DataType(p, Ent) == Type {
            Counter++;
            if *ppStart == nil {
                *ppStart = (int *)p.pBody + Ent; }
        
        } else if *ppStart {
            return Counter; }
    }
    return Counter;
} 
func Ndr_ObjIsType(p *Ndr_Data_t, Obj int32, Type int32) int32 {
    var Ent int32;
    // inlining of macro Ndr_ObjForEachEntry
    for  Ent = Obj + 1; Ent < Ndr_DataEnd(p, Obj); Ent += Ndr_DataSize(p, Ent)  {
        if Ndr_DataType(p, Ent) == NDR_OPERTYPE {
            return int32((Ndr_DataEntry(p, Ent) == Type)); }
    }
    return -1;
}
func Ndr_ObjReadBody(p *Ndr_Data_t, Obj int32, Type int32) int32 {
    var Ent int32;
    // inlining of macro Ndr_ObjForEachEntry
    for  Ent = Obj + 1; Ent < Ndr_DataEnd(p, Obj); Ent += Ndr_DataSize(p, Ent)  {
        if Ndr_DataType(p, Ent) == Type {
            return Ndr_DataEntry(p, Ent); }
    }
    return -1;
}
func Ndr_ObjReadBodyP(p *Ndr_Data_t, Obj int32, Type int32) *int32 {
    var Ent int32;
    // inlining of macro Ndr_ObjForEachEntry
    for  Ent = Obj + 1; Ent < Ndr_DataEnd(p, Obj); Ent += Ndr_DataSize(p, Ent)  {
        if Ndr_DataType(p, Ent) == Type {
            return Ndr_DataEntryP(p, Ent); }
    }
    return nil;
}
func Ndr_ObjWriteRange(p *Ndr_Data_t, Obj int32, pFile *os.File, fSkipBin int32) {
    var pArray *int32;
    var nArray int32 = Ndr_ObjReadArray( p, Obj, NDR_RANGE, &pArray );
    if (nArray == 0 || nArray == 1) && (fSkipBin != 0) {
        return; }
    if nArray == 3 && (fSkipBin != 0) {
        fmt.Fprintf(pFile, "signed "); 
    } else if nArray == 1 {
        if fSkipBin != 0 {
            fmt.Fprintf(pFile, "[%d:%d]", pArray[0], pArray[0]);
        } else {
            fmt.Fprintf(pFile, "[%d]", pArray[0]); }
    
    } else if nArray == 0 {
        if fSkipBin != 0 {
            fmt.Fprintf(pFile, "[%d:%d]", 0, 0);
        } else {
            fmt.Fprintf(pFile, "[%d]", 0); }
    
    } else {
        fmt.Fprintf(pFile, "[%d:%d]", pArray[0], pArray[1]); }
}
func Ndr_ObjReadOutName(p *Ndr_Data_t, Obj int32, pNames *string) string {
    return pNames[Ndr_ObjReadBody(p, Obj, NDR_OUTPUT)];
}
func Ndr_ObjReadInName(p *Ndr_Data_t, Obj int32, pNames *string) string {
    return pNames[Ndr_ObjReadBody(p, Obj, NDR_INPUT)];
}

func Ndr_DataCiNum(p *Ndr_Data_t, Mod int32) int32 { 
    var Obj, Count int32;
    // inlining of macro Ndr_ModForEachPi
    // inlining of macro Ndr_ModForEachObj
    for  Obj = Mod + 1; Obj < Ndr_DataEnd(p, Mod); Obj += Ndr_DataSize(p, Obj)  { if Ndr_DataType(p, Obj) != NDR_OBJECT {} else if !(Ndr_ObjIsType(p, Obj, ABC_OPER_CI) != 0) {} else {
        Count++; }
    }
    return Count;
}
func Ndr_DataCoNum(p *Ndr_Data_t, Mod int32) int32 { 
    var Obj, Count int32;
    // inlining of macro Ndr_ModForEachPo
    // inlining of macro Ndr_ModForEachObj
    for  Obj = Mod + 1; Obj < Ndr_DataEnd(p, Mod); Obj += Ndr_DataSize(p, Obj)  { if Ndr_DataType(p, Obj) != NDR_OBJECT {} else if !(Ndr_ObjIsType(p, Obj, ABC_OPER_CO) != 0) {} else {
        Count++; }
    }
    return Count;
}
func Ndr_DataObjNum(p *Ndr_Data_t, Mod int32) int32 { 
    var Obj, Count int32;
    // inlining of macro Ndr_ModForEachObj
    for  Obj = Mod + 1; Obj < Ndr_DataEnd(p, Mod); Obj += Ndr_DataSize(p, Obj)  { if Ndr_DataType(p, Obj) != NDR_OBJECT {} else {
        Count++; }
    }
    return Count;
}

// to write signal names, this procedure takes a mapping of name IDs into actual char-strings (pNames)
func Ndr_WriteVerilogModule(pFile *os.File, pDesign any, Mod int32, pNames *string) {
    var p *Ndr_Data_t  = (Ndr_Data_t *)pDesign; 
    // inlining of macro NDR_ALLOC
    var pOuts *int32  = ((int *) malloc(sizeof(int) * (size_t)(Ndr_DataCoNum(p, Mod))));
    var i int32;
    var k int32;
    var Obj int32;
    var nArray int32;
    var pArray *int32;
    var fFirst int32 = 1;

    fmt.Fprintf(pFile, "\nmodule %s (\n  ", pNames[Ndr_ObjReadEntry(p, Mod, NDR_NAME)]);

    // inlining of macro Ndr_ModForEachPi
    // inlining of macro Ndr_ModForEachObj
    for  Obj = Mod + 1; Obj < Ndr_DataEnd(p, Mod); Obj += Ndr_DataSize(p, Obj)  { if Ndr_DataType(p, Obj) != NDR_OBJECT {} else if !(Ndr_ObjIsType(p, Obj, ABC_OPER_CI) != 0) {} else {
        fmt.Fprintf(pFile, "%s, ", Ndr_ObjReadOutName(p, Obj, pNames)); }
    }

    fmt.Fprintf(pFile, "\n  ");

    // inlining of macro Ndr_ModForEachPo
    // inlining of macro Ndr_ModForEachObj
    for  Obj = Mod + 1; Obj < Ndr_DataEnd(p, Mod); Obj += Ndr_DataSize(p, Obj)  { if Ndr_DataType(p, Obj) != NDR_OBJECT {} else if !(Ndr_ObjIsType(p, Obj, ABC_OPER_CO) != 0) {} else {
        fmt.Fprintf(pFile, "%s%s", fFirst ? "":", ", Ndr_ObjReadInName(p, Obj, pNames));
fFirst = 0; }
    }

    fmt.Fprintf(pFile, "\n);\n\n");

    // inlining of macro Ndr_ModForEachPi
    // inlining of macro Ndr_ModForEachObj
    for  Obj = Mod + 1; Obj < Ndr_DataEnd(p, Mod); Obj += Ndr_DataSize(p, Obj)  { if Ndr_DataType(p, Obj) != NDR_OBJECT {} else if !(Ndr_ObjIsType(p, Obj, ABC_OPER_CI) != 0) {} else {
        fmt.Fprintf(pFile, "  input ");
        Ndr_ObjWriteRange( p, Obj, pFile, 1 );
        fmt.Fprintf(pFile, " %s;\n", Ndr_ObjReadOutName(p, Obj, pNames));
    }
    }

    i = 0;
    // inlining of macro Ndr_ModForEachPo
    // inlining of macro Ndr_ModForEachObj
    for  Obj = Mod + 1; Obj < Ndr_DataEnd(p, Mod); Obj += Ndr_DataSize(p, Obj)  { if Ndr_DataType(p, Obj) != NDR_OBJECT {} else if !(Ndr_ObjIsType(p, Obj, ABC_OPER_CO) != 0) {} else {
        fmt.Fprintf(pFile, "  output ");
        Ndr_ObjWriteRange( p, Obj, pFile, 1 );
        fmt.Fprintf(pFile, " %s;\n", Ndr_ObjReadInName(p, Obj, pNames));
        pOuts[i++] = Ndr_ObjReadBody(p, Obj, NDR_INPUT);
    }
    }

    fmt.Fprintf(pFile, "\n");

    // inlining of macro Ndr_ModForEachNode
    // inlining of macro Ndr_ModForEachObj
    for  Obj = Mod + 1; Obj < Ndr_DataEnd(p, Mod); Obj += Ndr_DataSize(p, Obj)  { if Ndr_DataType(p, Obj) != NDR_OBJECT {} else if (Ndr_ObjIsType(p, Obj, ABC_OPER_CI) != 0) || (Ndr_ObjIsType(p, Obj, ABC_OPER_CO) != 0) {} else {
        for  k = 0; k < i; k++  {
            if pOuts[k] == Ndr_ObjReadBody(p, Obj, NDR_OUTPUT) {
                break; }
        }
        if k < i {
            continue; }
        fmt.Fprintf(pFile, "  wire ");
        Ndr_ObjWriteRange( p, Obj, pFile, 1 );
        fmt.Fprintf(pFile, " %s;\n", Ndr_ObjReadOutName(p, Obj, pNames));
    }
    }
    free( pOuts );

    fmt.Fprintf(pFile, "\n");

    // inlining of macro Ndr_ModForEachNode
    // inlining of macro Ndr_ModForEachObj
    for  Obj = Mod + 1; Obj < Ndr_DataEnd(p, Mod); Obj += Ndr_DataSize(p, Obj)  { if Ndr_DataType(p, Obj) != NDR_OBJECT {} else if (Ndr_ObjIsType(p, Obj, ABC_OPER_CI) != 0) || (Ndr_ObjIsType(p, Obj, ABC_OPER_CO) != 0) {} else {
        var i int32;
        var Type int32 = Ndr_ObjReadBody(p, Obj, NDR_OPERTYPE);
        if Type >= 256 {
            fmt.Fprintf(pFile, "  %s ", pNames[Ndr_ObjReadEntry(p, Type-256, NDR_NAME)]);
            if Ndr_ObjReadBody(p, Obj, NDR_NAME) > 0 {
                fmt.Fprintf(pFile, "%s ", pNames[Ndr_ObjReadBody(p, Obj, NDR_NAME)]); }
            fmt.Fprintf(pFile, "( ");
            nArray = Ndr_ObjReadArray( p, Obj, NDR_INPUT, &pArray );
            for  i = 0; i < nArray; i++  {
                fmt.Fprintf(pFile, "%s%s ", pNames[pArray[i]], i==nArray-1 ? "":",");
            }
            fmt.Fprintf(pFile, ");\n");
            continue;
        }
        if Type == ABC_OPER_DFF {
            fmt.Fprintf(pFile, "  %s ", "ABC_DFF");
            if Ndr_ObjReadBody(p, Obj, NDR_NAME) > 0 {
                fmt.Fprintf(pFile, "%s ", pNames[Ndr_ObjReadBody(p, Obj, NDR_NAME)]); }
            fmt.Fprintf(pFile, "( ");
            nArray = Ndr_ObjReadArray( p, Obj, NDR_INPUT, &pArray );
            fmt.Fprintf(pFile, ".q(%s), ", Ndr_ObjReadOutName(p, Obj, pNames));
            fmt.Fprintf(pFile, ".d(%s), ", pNames[pArray[0]]);
            fmt.Fprintf(pFile, ".init(%s) ", pNames[pArray[1]]);
            fmt.Fprintf(pFile, ");\n");
            continue;
        }
        if Type == ABC_OPER_DFFRSE {
            fmt.Fprintf(pFile, "  %s ", "ABC_DFFRSE");
            if Ndr_ObjReadBody(p, Obj, NDR_NAME) > 0 {
                fmt.Fprintf(pFile, "%s ", pNames[Ndr_ObjReadBody(p, Obj, NDR_NAME)]); }
            fmt.Fprintf(pFile, "( ");
            nArray = Ndr_ObjReadArray( p, Obj, NDR_INPUT, &pArray );
            fmt.Fprintf(pFile, ".q(%s), ", Ndr_ObjReadOutName(p, Obj, pNames));
            fmt.Fprintf(pFile, ".d(%s), ", pNames[pArray[0]]);
            fmt.Fprintf(pFile, ".clk(%s), ", pNames[pArray[1]]);
            fmt.Fprintf(pFile, ".reset(%s), ", pNames[pArray[2]]);
            fmt.Fprintf(pFile, ".set(%s), ", pNames[pArray[3]]);
            fmt.Fprintf(pFile, ".enable(%s), ", pNames[pArray[4]]);
            fmt.Fprintf(pFile, ".async(%s), ", pNames[pArray[5]]);
            fmt.Fprintf(pFile, ".sre(%s), ", pNames[pArray[6]]);
            fmt.Fprintf(pFile, ".init(%s) ", pNames[pArray[7]]);
            fmt.Fprintf(pFile, ");\n");
            continue;
        }
        if Type == ABC_OPER_RAMR {
            fmt.Fprintf(pFile, "  %s ", "ABC_READ");
            if Ndr_ObjReadBody(p, Obj, NDR_NAME) > 0 {
                fmt.Fprintf(pFile, "%s ", pNames[Ndr_ObjReadBody(p, Obj, NDR_NAME)]); }
            fmt.Fprintf(pFile, "( ");
            nArray = Ndr_ObjReadArray( p, Obj, NDR_INPUT, &pArray );
            fmt.Fprintf(pFile, ".data(%s), ", Ndr_ObjReadOutName(p, Obj, pNames));
            fmt.Fprintf(pFile, ".mem_in(%s), ", pNames[pArray[0]]);
            fmt.Fprintf(pFile, ".addr(%s) ", pNames[pArray[1]]);
            fmt.Fprintf(pFile, ");\n");
            continue;
        }
        if Type == ABC_OPER_RAMW {
            fmt.Fprintf(pFile, "  %s ", "ABC_WRITE");
            if Ndr_ObjReadBody(p, Obj, NDR_NAME) > 0 {
                fmt.Fprintf(pFile, "%s ", pNames[Ndr_ObjReadBody(p, Obj, NDR_NAME)]); }
            fmt.Fprintf(pFile, "( ");
            nArray = Ndr_ObjReadArray( p, Obj, NDR_INPUT, &pArray );
            fmt.Fprintf(pFile, ".mem_out(%s), ", Ndr_ObjReadOutName(p, Obj, pNames));
            fmt.Fprintf(pFile, ".mem_in(%s), ", pNames[pArray[0]]);
            fmt.Fprintf(pFile, ".addr(%s), ", pNames[pArray[1]]);
            fmt.Fprintf(pFile, ".data(%s) ", pNames[pArray[2]]);
            fmt.Fprintf(pFile, ");\n");
            continue;
        }
        fmt.Fprintf(pFile, "  assign %s = ", Ndr_ObjReadOutName(p, Obj, pNames));
        nArray = Ndr_ObjReadArray( p, Obj, NDR_INPUT, &pArray );
        if nArray == 0 {
            fmt.Fprintf(pFile, "%s;\n", (char *)Ndr_ObjReadBodyP(p, Obj, NDR_FUNCTION));
        } else if nArray == 1 && Ndr_ObjReadBody(p, Obj, NDR_OPERTYPE) == ABC_OPER_BIT_BUF {
            fmt.Fprintf(pFile, "%s;\n", pNames[pArray[0]]);
        } else if Type == ABC_OPER_SLICE {
            fmt.Fprintf(pFile, "%s", pNames[pArray[0]]);
Ndr_ObjWriteRange( p, Obj, pFile, 0 );
fmt.Fprintf(pFile, ";\n");
        } else if Type == ABC_OPER_CONCAT {
            fmt.Fprintf(pFile, "{");
            for  i = 0; i < nArray; i++  {
                fmt.Fprintf(pFile, "%s%s", pNames[pArray[i]], i==nArray-1 ? "":", ");
            }
            fmt.Fprintf(pFile, "};\n");
        
        } else if nArray == 1 {
            fmt.Fprintf(pFile, "%s %s;\n", Abc_OperName(Ndr_ObjReadBody(p, Obj, NDR_OPERTYPE)), pNames[pArray[0]]);
        } else if nArray == 2 {
            fmt.Fprintf(pFile, "%s %s %s;\n", pNames[pArray[0]], Abc_OperName(Ndr_ObjReadBody(p, Obj, NDR_OPERTYPE)), pNames[pArray[1]]);
        } else if nArray == 3 && Type == ABC_OPER_ARI_ADD {
            fmt.Fprintf(pFile, "%s + %s + %s;\n", pNames[pArray[0]], pNames[pArray[1]], pNames[pArray[2]]);
        } else if Type == ABC_OPER_BIT_MUX {
            fmt.Fprintf(pFile, "%s ? %s : %s;\n", pNames[pArray[0]], pNames[pArray[1]], pNames[pArray[2]]);
        } else {
            fmt.Fprintf(pFile, "<cannot write operation %s>;\n", Abc_OperName(Ndr_ObjReadBody(p, Obj, NDR_OPERTYPE))); }
    }
    }

    fmt.Fprintf(pFile, "\nendmodule\n\n");
}

// to write signal names, this procedure takes a mapping of name IDs into actual char-strings (pNames)
func Ndr_WriteVerilog(pFileName string, pDesign any, pNames *string) {
    var p *Ndr_Data_t  = (Ndr_Data_t *)pDesign; var Mod int32;

    var pFile *os.File  = pFileName ? os.OpenFile(pFileName, os.O_WRONLY | os.O_CREATE | os.O_TRUNC, 0666) : os.Stdout;
    if pFile == nil { fmt.Printf("Cannot open file \"%s\" for writing.\n", pFileName); return; }

    // inlining of macro Ndr_DesForEachMod
    for  Mod = 1; Mod < Ndr_DataEntry(p, 0); Mod += Ndr_DataSize(p, Mod)  { if Ndr_DataType(p, Mod) != NDR_MODULE {} else {
        Ndr_WriteVerilogModule( pFile, p, Mod, pNames ); }
    }
    
    if pFileName { _ = (pFile).Close(); }
}


////////////////////////////////////////////////////////////////////////
///                     EXTERNAL PROCEDURES                          ///
////////////////////////////////////////////////////////////////////////

// creating a new module (returns pointer to the memory buffer storing the module info)
func Ndr_Create(Name int32) any {
    // inlining of macro NDR_ALLOC
    var p *Ndr_Data_t  = ((Ndr_Data_t *) malloc(sizeof(Ndr_Data_t) * (size_t)(1)));
    p.nSize = 0;
    p.nCap  = 16;
    // inlining of macro NDR_ALLOC
    p.pHead = ((unsigned char *) malloc(sizeof(unsigned char) * (size_t)(p.nCap)));
    // inlining of macro NDR_ALLOC
    p.pBody = ((unsigned int *) malloc(sizeof(unsigned int) * (size_t)(p.nCap * 4)));
    Ndr_DataPush( p, NDR_DESIGN, 0 );
    Ndr_DataPush( p, NDR_NAME, Name );
    Ndr_DataAddTo( p, 0, p.nSize );
    assert(p.nSize == 2, "p.nSize == 2");
    assert((Name != 0), "Name");
    return p;
}

// creating a new module in an already started design 
// returns module ID to be used when adding objects to the module
func Ndr_AddModule(pDesign any, Name int32) int32 {
    var p *Ndr_Data_t  = (Ndr_Data_t *)pDesign;
    var Mod int32  = p.nSize;  
    Ndr_DataResize( p, 6 );
    Ndr_DataPush( p, NDR_MODULE, 0 );
    Ndr_DataPush( p, NDR_NAME, Name );
    Ndr_DataAddTo( p, Mod, p.nSize - Mod );
    Ndr_DataAddTo( p, 0, p.nSize - Mod );
    assert(int32(p.pBody[0]) == p.nSize, "int32(p.pBody[0]) == p.nSize");
    return Mod + 256;
}

// adding a new object (input/output/flop/intenal node) to an already started module
// this procedure takes the design, the module ID, and the parameters of the boject
// (please note that all objects should be added to a given module before starting a new module)
func Ndr_AddObject(pDesign any, ModuleId int32, ObjType int32, InstName int32, RangeLeft int32, RangeRight int32, fSignedness int32, nInputs int32, pInputs *int32, nOutputs int32, pOutputs *int32, pFunction string) {
    var p *Ndr_Data_t  = (Ndr_Data_t *)pDesign;
    var Mod int32  = ModuleId - 256;  
    var Obj int32  = p.nSize;  assert(ObjType != 0, "ObjType != 0");
    Ndr_DataResize( p, 6 );
    Ndr_DataPush( p, NDR_OBJECT, 0 );
    Ndr_DataPush( p, NDR_OPERTYPE, ObjType );
    Ndr_DataPushRange( p, RangeLeft, RangeRight, fSignedness );
    if InstName != 0 {
        Ndr_DataPush( p, NDR_NAME, InstName ); }
    Ndr_DataPushArray( p, NDR_INPUT, nInputs, pInputs );
    Ndr_DataPushArray( p, NDR_OUTPUT, nOutputs, pOutputs );
    Ndr_DataPushString( p, ObjType, NDR_FUNCTION, pFunction );
    Ndr_DataAddTo( p, Obj, p.nSize - Obj );
    Ndr_DataAddTo( p, Mod, p.nSize - Obj );
    Ndr_DataAddTo( p, 0, p.nSize - Obj );
    assert(int32(p.pBody[0]) == p.nSize, "int32(p.pBody[0]) == p.nSize");
}

// deallocate the memory buffer
func Ndr_Delete(pDesign any) {
    var p *Ndr_Data_t  = (Ndr_Data_t *)pDesign;
    if !(p != nil) { return; }
    free( p.pHead );
    free( p.pBody );
    free( p );
}


////////////////////////////////////////////////////////////////////////
///                  FILE READING AND WRITING                        ///
////////////////////////////////////////////////////////////////////////

// file reading/writing
func Ndr_Read(pFileName string) any {
    var p *Ndr_Data_t; var nFileSize, RetValue int32;
    var pFile *os.File  = os.Open(pFileName);
    if pFile == nil { fmt.Printf("Cannot open file \"%s\" for reading.\n", pFileName); return nil; }
    // check file size
    fseek( pFile, 0, SEEK_END );
    nFileSize = ftell( pFile ); 
    if nFileSize % 5 != 0 {
        return nil; }
    assert(nFileSize % 5 == 0, "nFileSize % 5 == 0");
    rewind( pFile );
    // create structure
    // inlining of macro NDR_ALLOC
    p = ((Ndr_Data_t *) malloc(sizeof(Ndr_Data_t) * (size_t)(1)));
    p.nSize = p.nCap = nFileSize / 5;
    // inlining of macro NDR_ALLOC
    p.pHead = ((unsigned char *) malloc(sizeof(unsigned char) * (size_t)(p.nCap)));
    // inlining of macro NDR_ALLOC
    p.pBody = ((unsigned int *) malloc(sizeof(unsigned int) * (size_t)(p.nCap * 4)));
    RetValue = int32(fread( p.pBody, 4, p.nCap, pFile ));
    RetValue = int32(fread( p.pHead, 1, p.nCap, pFile ));
    assert(p.nSize == int32(p.pBody[0]), "p.nSize == int32(p.pBody[0])");
    _ = (pFile).Close();
    //printf( "Read the design from file \"%s\".\n", pFileName );
    return p;
}
func Ndr_Write(pFileName string, pDesign any) {
    var p *Ndr_Data_t  = (Ndr_Data_t *)pDesign; var RetValue int32;
    var pFile *os.File  = os.OpenFile(pFileName, os.O_WRONLY | os.O_CREATE | os.O_TRUNC, 0666);
    if pFile == nil { fmt.Printf("Cannot open file \"%s\" for writing.\n", pFileName); return; }
    RetValue = int32(fwrite( p.pBody, 4, p.pBody[0], pFile ));
    RetValue = int32(fwrite( p.pHead, 1, p.pBody[0], pFile ));
    _ = (pFile).Close();
    //printf( "Dumped the design into file \"%s\".\n", pFileName );
}


////////////////////////////////////////////////////////////////////////
///                     TESTING PROCEDURE                            ///
////////////////////////////////////////////////////////////////////////

// This testing procedure creates and writes into a Verilog file 
// the following design composed of one module

// module add10 ( input [3:0] a, output [3:0] s );
//   wire [3:0] const10 = 4'b1010;
//   assign s = a + const10;
// endmodule

func Ndr_ModuleTest() {
    // name IDs
    var NameIdA int32  = 2;
    var NameIdS int32  = 3;
    var NameIdC int32  = 4;
    // array of fanins of node s
    var Fanins [2]int32 = [2]int32{ NameIdA, NameIdC };
    // map name IDs into char strings
    var ppNames [5]string = [5]string{ nil, "add10", "a", "s", "const10" };

    // create a new module
    var pDesign any  = Ndr_Create( 1 );

    var ModuleID int32  = Ndr_AddModule( pDesign, 1 );

    // add objects to the modele
    Ndr_AddObject( pDesign, ModuleID, ABC_OPER_CI,       0,   3, 0, 0,   0, nil,      1, &NameIdA,   nil      ); // no fanins
    Ndr_AddObject( pDesign, ModuleID, ABC_OPER_CONST,    0,   3, 0, 0,   0, nil,      1, &NameIdC,   "4'b1010" ); // no fanins
    Ndr_AddObject( pDesign, ModuleID, ABC_OPER_ARI_ADD,  0,   3, 0, 0,   2, Fanins,    1, &NameIdS,   nil      ); // fanins are a and const10 
    Ndr_AddObject( pDesign, ModuleID, ABC_OPER_CO,       0,   3, 0, 0,   1, &NameIdS,  0, nil,       nil      ); // fanin is a

    // write Verilog for verification
    Ndr_WriteVerilog( nil, pDesign, ppNames );
    Ndr_Write( "add4.ndr", pDesign );
    Ndr_Delete( pDesign );
}



// This testing procedure creates and writes into a Verilog file 
// the following design composed of one adder divided into two

// module add8 ( input [7:0] a, input [7:0] b, output [7:0] s, output co );
//   wire [3:0] a0 = a[3:0];
//   wire [3:0] b0 = b[3:0];

//   wire [7:4] a1 = a[7:4];
//   wire [7:4] b1 = b[7:4];

//   wire [4:0] r0 = a0 + b0;
//   wire [3:0] s0 = r0[3:0];
//   wire rco = r0[4];

//   wire [4:0] r1 = a1 + b1 + rco;
//   wire [3:0] s1 = r1[3:0];
//   assign co = r1[4];

//   assign s = {s1, s0};
// endmodule

func Ndr_ModuleTestAdder() {
    // map name IDs into char strings
    char * ppNames[20] = {  nil, 
                           "a", "b", "s", "co",          // 1,  2,  3,  4
                           "a0",  "a1",  "b0",  "b1",    // 5,  6,  7,  8
                           "r0", "s0", "rco",            // 9,  10, 11
                           "r1", "s1", "add8"            // 12, 13, 14
                         };
    // fanins 
    var FaninA int32         =  1;
    var FaninB int32         =  2;
    var FaninS int32         =  3;
    var FaninCO int32        =  4;

    var FaninA0 int32        =  5;
    var FaninA1 int32        =  6;
    var FaninB0 int32        =  7;
    var FaninB1 int32        =  8;

    var FaninR0 int32        =  9;
    var FaninS0 int32        = 10;
    var FaninRCO int32       = 11;

    var FaninR1 int32        = 12;
    var FaninS1 int32        = 13;

    var Fanins1 [2]int32 = [2]int32{ FaninA0, FaninB0 };
    var Fanins2 [3]int32 = [3]int32{ FaninA1, FaninB1, FaninRCO };
    var Fanins3 [4]int32 = [4]int32{ FaninS1, FaninS0 };

    // create a new module
    var pDesign any  = Ndr_Create( 14 );

    var ModuleID int32  = Ndr_AddModule( pDesign, 14 );

    // add objects to the modele
    Ndr_AddObject( pDesign, ModuleID, ABC_OPER_CI,       0,   7, 0, 0,   0, nil,      1, &FaninA,   nil  );  // no fanins
    Ndr_AddObject( pDesign, ModuleID, ABC_OPER_CI,       0,   7, 0, 0,   0, nil,      1, &FaninB,   nil  );  // no fanins

    Ndr_AddObject( pDesign, ModuleID, ABC_OPER_SLICE,    0,   3, 0, 0,   1, &FaninA,   1, &FaninA0,   nil );  // wire [3:0] a0 = a[3:0];
    Ndr_AddObject( pDesign, ModuleID, ABC_OPER_SLICE,    0,   3, 0, 0,   1, &FaninB,   1, &FaninB0,   nil );  // wire [3:0] b0 = a[3:0];

    Ndr_AddObject( pDesign, ModuleID, ABC_OPER_SLICE,    0,   7, 4, 0,   1, &FaninA,   1, &FaninA1,   nil );  // wire [7:4] a1 = a[7:4];
    Ndr_AddObject( pDesign, ModuleID, ABC_OPER_SLICE,    0,   7, 4, 0,   1, &FaninB,   1, &FaninB1,   nil );  // wire [7:4] b1 = b[7:4];

    Ndr_AddObject( pDesign, ModuleID, ABC_OPER_ARI_ADD,  0,   4, 0, 0,   2, Fanins1,   1, &FaninR0,   nil );  // wire [4:0] r0 = a0 + b0;
    Ndr_AddObject( pDesign, ModuleID, ABC_OPER_SLICE,    0,   3, 0, 0,   1, &FaninR0,  1, &FaninS0,   nil );  // wire [3:0] s0 = r0[3:0];
    Ndr_AddObject( pDesign, ModuleID, ABC_OPER_SLICE,    0,   4, 4, 0,   1, &FaninR0,  1, &FaninRCO,  nil );  // wire rco = r0[4];

    Ndr_AddObject( pDesign, ModuleID, ABC_OPER_ARI_ADD,  0,   4, 0, 0,   3, Fanins2,   1, &FaninR1,   nil );  // wire [4:0] r1 = a1 + b1 + rco;
    Ndr_AddObject( pDesign, ModuleID, ABC_OPER_SLICE,    0,   3, 0, 0,   1, &FaninR1,  1, &FaninS1,   nil );  // wire [3:0] s1 = r1[3:0];
    Ndr_AddObject( pDesign, ModuleID, ABC_OPER_SLICE,    0,   4, 4, 0,   1, &FaninR1,  1, &FaninCO,   nil );  // assign co = r1[4];

    Ndr_AddObject( pDesign, ModuleID, ABC_OPER_CONCAT,   0,   7, 0, 0,   2, Fanins3,   1, &FaninS,    nil );  // s = {s1, s0};

    Ndr_AddObject( pDesign, ModuleID, ABC_OPER_CO,       0,   7, 0, 0,   1, &FaninS,   0, nil,       nil ); 
    Ndr_AddObject( pDesign, ModuleID, ABC_OPER_CO,       0,   0, 0, 0,   1, &FaninCO,  0, nil,       nil ); 

    // write Verilog for verification
    Ndr_WriteVerilog( nil, pDesign, ppNames );
    Ndr_Write( "add8.ndr", pDesign );
    Ndr_Delete( pDesign );

}

// This testing procedure creates and writes into a Verilog file 
// the following hierarchical design composed of two modules

// module mux21w ( input sel, input [3:0] d1, input [3:0] d0, output [3:0] out );
//   assign out = sel ? d1 : d0;
// endmodule

// module mux41w ( input [1:0] sel, input [15:0] d, output [3:0] out );
//   wire [3:0] t0, t1;
//   wire [3:0] d0 = d[3:0];
//   wire [3:0] d1 = d[7:4];
//   wire [3:0] d2 = d[11:8];
//   wire [3:0] d3 = d[15:12];
//   wire sel0 = sel[0];
//   wire sel1 = sel[1];
//   mux21w i0 ( sel0, d1, d0, t0 );
//   mux21w i1 ( sel0, d3, d2, t1 );
//   mux21w i2 ( sel1, t1, t0, out );
// endmodule

func Ndr_ModuleTestHierarchy() {
    // map name IDs into char strings
    char * ppNames[20] = {  nil, 
                           "mux21w", "mux41w",     // 1,  2
                           "sel",  "d",  "out",    // 3,  4,  5
                           "d0", "d1", "d2", "d3", // 6,  7,  8,  9
                           "sel0", "sel1",         // 10, 11,
                           "t0", "t1",             // 12, 13
                           "i0", "i1", "i2"        // 14, 15, 16
                         };
    // fanins 
    var FaninSel int32       =  3;
    var FaninSel0 int32      = 10;
    var FaninSel1 int32      = 11;
    var FaninD int32         =  4;
    var FaninD0 int32        =  6;
    var FaninD1 int32        =  7;
    var FaninD2 int32        =  8;
    var FaninD3 int32        =  9;
    var FaninT0 int32        = 12;
    var FaninT1 int32        = 13;
    var FaninOut int32       =  5;
    var Fanins1 [3]int32 = [3]int32{  FaninSel,  FaninD1, FaninD0 };
    var Fanins3 [3][3]int32 = [3][3]int32{ {FaninSel0, FaninD1, FaninD0 },  
                          {FaninSel0, FaninD3, FaninD2 },  
                          {FaninSel1, FaninT1, FaninT0 } };

    // create a new module
    var pDesign any  = Ndr_Create( 2 );

    var Module21, Module41 int32;

    Module21 = Ndr_AddModule( pDesign, 1 );
    
    Ndr_AddObject( pDesign, Module21, ABC_OPER_CI,        0,   0, 0, 0,   0, nil,        1, &FaninSel,  nil ); 
    Ndr_AddObject( pDesign, Module21, ABC_OPER_CI,        0,   3, 0, 0,   0, nil,        1, &FaninD1,   nil ); 
    Ndr_AddObject( pDesign, Module21, ABC_OPER_CI,        0,   3, 0, 0,   0, nil,        1, &FaninD0,   nil ); 
    Ndr_AddObject( pDesign, Module21, ABC_OPER_BIT_MUX,   0,   3, 0, 0,   3, Fanins1,     1, &FaninOut,  nil ); 
    Ndr_AddObject( pDesign, Module21, ABC_OPER_CO,        0,   3, 0, 0,   1, &FaninOut,   0, nil,       nil ); 

    Module41 = Ndr_AddModule( pDesign, 2 );

    Ndr_AddObject( pDesign, Module41, ABC_OPER_CI,        0,   1, 0, 0,   0, nil,        1, &FaninSel,  nil ); 
    Ndr_AddObject( pDesign, Module41, ABC_OPER_CI,        0,   15,0, 0,   0, nil,        1, &FaninD,    nil ); 

    Ndr_AddObject( pDesign, Module41, ABC_OPER_SLICE,     0,   3, 0, 0,   1, &FaninD,     1, &FaninD0,   nil ); 
    Ndr_AddObject( pDesign, Module41, ABC_OPER_SLICE,     0,   7, 4, 0,   1, &FaninD,     1, &FaninD1,   nil ); 
    Ndr_AddObject( pDesign, Module41, ABC_OPER_SLICE,     0,   11,8, 0,   1, &FaninD,     1, &FaninD2,   nil ); 
    Ndr_AddObject( pDesign, Module41, ABC_OPER_SLICE,     0,   15,12,0,   1, &FaninD,     1, &FaninD3,   nil ); 

    Ndr_AddObject( pDesign, Module41, ABC_OPER_SLICE,     0,   0, 0, 0,   1, &FaninSel,   1, &FaninSel0, nil ); 
    Ndr_AddObject( pDesign, Module41, ABC_OPER_SLICE,     0,   1, 1, 0,   1, &FaninSel,   1, &FaninSel1, nil ); 

    Ndr_AddObject( pDesign, Module41, Module21,          14,   3, 0, 0,   3, Fanins3[0],  1, &FaninT0,   nil ); 
    Ndr_AddObject( pDesign, Module41, Module21,          15,   3, 0, 0,   3, Fanins3[1],  1, &FaninT1,   nil ); 
    Ndr_AddObject( pDesign, Module41, Module21,          16,   3, 0, 0,   3, Fanins3[2],  1, &FaninOut,  nil ); 
    Ndr_AddObject( pDesign, Module41, ABC_OPER_CO,        0,   3, 0, 0,   1, &FaninOut,   0, nil,       nil ); 

    // write Verilog for verification
    Ndr_WriteVerilog( nil, pDesign, ppNames );
    Ndr_Write( "mux41w.ndr", pDesign );
    Ndr_Delete( pDesign );
}


// This testing procedure creates and writes into a Verilog file 
// the following design with read/write memory ports

// module test ( input clk, input [8:0] raddr, input [8:0] waddr, input [31:0] data, input [16383:0] mem_init, output out );
// 
//    wire [31:0] read1, read2;
// 
//    wire [16383:0] mem_fo1, mem_fo2,  mem_fi1, mem_fi2;
// 
//    ABC_FF    i_reg1   ( .q(mem_fo1), .d(mem_fi1), .init(mem_init) );
//    ABC_FF    i_reg2   ( .q(mem_fo2), .d(mem_fi2), .init(mem_init) );
// 
//    ABC_WRITE i_write1 ( .mem_out(mem_fi1), .mem_in(mem_fo1), .addr(waddr), .data(data) );
//    ABC_WRITE i_write2 ( .mem_out(mem_fi2), .mem_in(mem_fo2), .addr(waddr), .data(data) );
// 
//    ABC_READ  i_read1  ( .data(read1), .mem_in(mem_fi1), .addr(raddr) );
//    ABC_READ  i_read2  ( .data(read2), .mem_in(mem_fi2), .addr(raddr) );
//
//    assign out = read1 != read2;
//endmodule

func Ndr_ModuleTestMemory() {
    // map name IDs into char strings
    char * ppNames[20] = {  nil, 
                           "clk", "raddr", "waddr", "data", "mem_init", "out",  // 1, 2, 3, 4, 5, 6
                           "read1",  "read2",                                   // 7. 8
                           "mem_fo1", "mem_fo2", "mem_fi1", "mem_fi2",          // 9, 10, 11, 12
                           "i_reg1", "i_reg2",                                  // 13, 14
                           "i_read1", "i_read2",                                // 15, 16
                           "i_write1", "i_write2", "memtest"                    // 17, 18, 19
                         };
    // inputs
    var NameIdClk int32      = 1;
    var NameIdRaddr int32    = 2;
    var NameIdWaddr int32    = 3;
    var NameIdData int32     = 4;
    var NameIdMemInit int32  = 5;
    // flops
    var NameIdFF1 int32      =  9;
    var NameIdFF2 int32      = 10;
    var FaninsFF1 [2]int32 = [2]int32{ 11, 5 };
    var FaninsFF2 [2]int32 = [2]int32{ 12, 5 };
    // writes
    var NameIdWrite1 int32     = 11;
    var NameIdWrite2 int32     = 12;
    var FaninsWrite1 [3]int32 = [3]int32{  9, 3, 4 };
    var FaninsWrite2 [3]int32 = [3]int32{ 10, 3, 4 };
    // reads
    var NameIdRead1 int32     =  7;
    var NameIdRead2 int32     =  8;
    var FaninsRead1 [2]int32 = [2]int32{ 11, 2 };
    var FaninsRead2 [2]int32 = [2]int32{ 12, 2 };
    // compare
    var NameIdComp int32       = 6;
    var FaninsComp [2]int32 = [2]int32{ 7, 8 };

    // create a new module 
    var pDesign any  = Ndr_Create( 19 );           // create design named "memtest"

    var ModuleID int32  = Ndr_AddModule( pDesign, 19 ); // create module named "memtest"

    // add objects to the module
    Ndr_AddObject( pDesign, ModuleID, ABC_OPER_CI,           0,     0, 0, 0,   0, nil,         1, &NameIdClk,     nil );
    Ndr_AddObject( pDesign, ModuleID, ABC_OPER_CI,           0,     8, 0, 0,   0, nil,         1, &NameIdRaddr,   nil );
    Ndr_AddObject( pDesign, ModuleID, ABC_OPER_CI,           0,     8, 0, 0,   0, nil,         1, &NameIdWaddr,   nil );
    Ndr_AddObject( pDesign, ModuleID, ABC_OPER_CI,           0,    31, 0, 0,   0, nil,         1, &NameIdData,    nil );
    Ndr_AddObject( pDesign, ModuleID, ABC_OPER_CI,           0, 16383, 0, 0,   0, nil,         1, &NameIdMemInit, nil );

    Ndr_AddObject( pDesign, ModuleID, ABC_OPER_CO,           0,     0, 0, 0,   1, &NameIdComp,  0, nil,           nil );

    Ndr_AddObject( pDesign, ModuleID, ABC_OPER_DFF,         13, 16383, 0, 0,   2, FaninsFF1,    1, &NameIdFF1,     nil );
    Ndr_AddObject( pDesign, ModuleID, ABC_OPER_DFF,         14, 16383, 0, 0,   2, FaninsFF2,    1, &NameIdFF2,     nil );

    Ndr_AddObject( pDesign, ModuleID, ABC_OPER_RAMW,        17, 16383, 0, 0,   3, FaninsWrite1, 1, &NameIdWrite1,  nil );
    Ndr_AddObject( pDesign, ModuleID, ABC_OPER_RAMW,        18, 16383, 0, 0,   3, FaninsWrite2, 1, &NameIdWrite2,  nil );

    Ndr_AddObject( pDesign, ModuleID, ABC_OPER_RAMR,        15,    31, 0, 0,   2, FaninsRead1,  1, &NameIdRead1,   nil );
    Ndr_AddObject( pDesign, ModuleID, ABC_OPER_RAMR,        16,    31, 0, 0,   2, FaninsRead2,  1, &NameIdRead2,   nil );

    Ndr_AddObject( pDesign, ModuleID, ABC_OPER_COMP_NOTEQU,  0,     0, 0, 0,   2, FaninsComp,   1, &NameIdComp,    nil );

    // write Verilog for verification
    Ndr_WriteVerilog( nil, pDesign, ppNames );
    Ndr_Write( "memtest.ndr", pDesign );
    Ndr_Delete( pDesign );
}

// This testing procedure creates and writes into a Verilog file 
// the following design composed of one word-level flop

// module flop ( input [3:0] data, input clk, input reset, input set, input enable, input async, input sre, input [3:0] init, output [3:0] q );
//   ABC_DFFRSE reg1 ( .d(data), .clk(clk), .reset(reset), .set(set), .enable(enable), .async(async), .sre(sre), .init(init), .q(q) ) ;
// endmodule

func Ndr_ModuleTestFlop() {
    // map name IDs into char strings
    var ppNames [12]string = [12]string{ nil, "flop", "data", "clk", "reset", "set", "enable", "async", "sre", "init", "q" };
    // name IDs
    var NameIdData int32    =  2;
    var NameIdClk int32     =  3;
    var NameIdReset int32   =  4;
    var NameIdSet int32     =  5;
    var NameIdEnable int32  =  6;
    var NameIdAsync int32   =  7;
    var NameIdSre int32     =  8;
    var NameIdInit int32    =  9;
    var NameIdQ int32       = 10;
    // array of fanins of node s
    var Fanins [8]int32 = [8]int32{ NameIdData, NameIdClk, NameIdReset, NameIdSet, NameIdEnable, NameIdAsync, NameIdSre, NameIdInit };

    // create a new module
    var pDesign any  = Ndr_Create( 1 );

    var ModuleID int32  = Ndr_AddModule( pDesign, 1 );

    // add objects to the modele
    Ndr_AddObject( pDesign, ModuleID, ABC_OPER_CI,       0,   3, 0, 0,   0, nil,      1, &NameIdData,   nil );
    Ndr_AddObject( pDesign, ModuleID, ABC_OPER_CI,       0,   0, 0, 0,   0, nil,      1, &NameIdClk,    nil );
    Ndr_AddObject( pDesign, ModuleID, ABC_OPER_CI,       0,   0, 0, 0,   0, nil,      1, &NameIdReset,  nil );
    Ndr_AddObject( pDesign, ModuleID, ABC_OPER_CI,       0,   0, 0, 0,   0, nil,      1, &NameIdSet,    nil );
    Ndr_AddObject( pDesign, ModuleID, ABC_OPER_CI,       0,   0, 0, 0,   0, nil,      1, &NameIdEnable, nil );
    Ndr_AddObject( pDesign, ModuleID, ABC_OPER_CI,       0,   0, 0, 0,   0, nil,      1, &NameIdAsync,  nil );
    Ndr_AddObject( pDesign, ModuleID, ABC_OPER_CI,       0,   0, 0, 0,   0, nil,      1, &NameIdSre,    nil );
    Ndr_AddObject( pDesign, ModuleID, ABC_OPER_CI,       0,   3, 0, 0,   0, nil,      1, &NameIdInit,   nil );

    Ndr_AddObject( pDesign, ModuleID, ABC_OPER_DFFRSE,   0,   3, 0, 0,   8, Fanins,    1, &NameIdQ,      nil );

    Ndr_AddObject( pDesign, ModuleID, ABC_OPER_CO,       0,   3, 0, 0,   1, &NameIdQ,  0, nil,          nil );

    // write Verilog for verification
    Ndr_WriteVerilog( nil, pDesign, ppNames );
    Ndr_Write( "flop.ndr", pDesign );
    Ndr_Delete( pDesign );
}


// This testing procedure creates and writes into a Verilog file 
// the following design composed of one selector

// module sel ( input [3:0] c, input [2:0] d0, input [2:0] d1, input [2:0] d2, input [2:0] d3, input [2:0] out );
//     wire [2:0] s7 ;
//     always @( c or d0 or d1 or d2 or d3 )
//       begin
//         case ( c )
//           4'b0001 : s7 = d0 ;
//           4'b0010 : s7 = d1 ;
//           4'b0100 : s7 = d2 ;
//           4'b1000 : s7 = d3 ;
//         endcase
//       end
//     assign  out = s7 ;
// endmodule

func Ndr_ModuleTestSelSel() {
    // map name IDs into char strings
    //char * ppNames[12] = { NULL, "sel", "c", "d0", "d1", "d2", "d3", "out" };
    // name IDs
    var NameIdC int32       =  2;
    var NameIdD0 int32      =  3;
    var NameIdD1 int32      =  4;
    var NameIdD2 int32      =  5;
    var NameIdD3 int32      =  6;
    var NameIdOut int32     =  7;
    // array of fanins of node s
    var Fanins [8]int32 = [8]int32{ NameIdC, NameIdD0, NameIdD1, NameIdD2, NameIdD3 };

    // create a new module
    var pDesign any  = Ndr_Create( 1 );

    var ModuleID int32  = Ndr_AddModule( pDesign, 1 );

    // add objects to the modele
    Ndr_AddObject( pDesign, ModuleID, ABC_OPER_CI,       0,   3, 0, 0,   0, nil,      1, &NameIdC,      nil );
    Ndr_AddObject( pDesign, ModuleID, ABC_OPER_CI,       0,   2, 0, 0,   0, nil,      1, &NameIdD0,     nil );
    Ndr_AddObject( pDesign, ModuleID, ABC_OPER_CI,       0,   2, 0, 0,   0, nil,      1, &NameIdD1,     nil );
    Ndr_AddObject( pDesign, ModuleID, ABC_OPER_CI,       0,   2, 0, 0,   0, nil,      1, &NameIdD2,     nil );
    Ndr_AddObject( pDesign, ModuleID, ABC_OPER_CI,       0,   2, 0, 0,   0, nil,      1, &NameIdD3,     nil );

    Ndr_AddObject( pDesign, ModuleID, ABC_OPER_SEL_SEL,  0,   2, 0, 0,   5, Fanins,    1, &NameIdOut,    nil );

    Ndr_AddObject( pDesign, ModuleID, ABC_OPER_CO,       0,   2, 0, 0,   1, &NameIdOut,0, nil,          nil );

    // write Verilog for verification
    //Ndr_WriteVerilog( NULL, pDesign, ppNames );
    Ndr_Write( "sel.ndr", pDesign );
    Ndr_Delete( pDesign );
}

// This testing procedure creates and writes into a Verilog file 
// the following design composed of one decoder

// module dec ( input [1:0] in, output [3:0] out );
//     wire out0 = ~in[1] & ~in[0] ;
//     wire out1 = ~in[1] &  in[0] ;
//     wire out2 =  in[1] & ~in[0] ;
//     wire out3 =  in[1] &  in[0] ;
//     assign out = { out3, out2, out1, out0 } ;
// endmodule

func Ndr_ModuleTestDec() {
    // map name IDs into char strings
    //char * ppNames[12] = { NULL, "dec", "in", "out" };
    // name IDs
    var NameIdIn int32      =  2;
    var NameIdOut int32     =  3;

    // create a new module
    var pDesign any  = Ndr_Create( 1 );

    var ModuleID int32  = Ndr_AddModule( pDesign, 1 );

    // add objects to the modele
    Ndr_AddObject( pDesign, ModuleID, ABC_OPER_CI,       0,   1, 0, 0,   0, nil,       1, &NameIdIn,     nil );
    Ndr_AddObject( pDesign, ModuleID, ABC_OPER_SEL_DEC,  0,   3, 0, 0,   1, &NameIdIn,  1, &NameIdOut,    nil );
    Ndr_AddObject( pDesign, ModuleID, ABC_OPER_CO,       0,   3, 0, 0,   1, &NameIdOut, 0, nil,          nil );

    Ndr_Write( "dec.ndr", pDesign );
    Ndr_Delete( pDesign );
}

// This testing procedure creates and writes into a Verilog file 
// the following design composed of one adder/subtractor

// module addsub ( input mode, input cin, input [2:0] a, input [2:0] b, output [3:0] out );
//     assign out = mode ? a+b+cin : a-b-cin ;
// endmodule

func Ndr_ModuleTestAddSub() {
    // map name IDs into char strings
    //char * ppNames[12] = { NULL, "addsub", "mode", "cin", "a", "b", "out" };
    // name IDs
    var NameIdInMode int32  =  2;
    var NameIdInCin int32   =  3;
    var NameIdInA int32     =  4;
    var NameIdInB int32     =  5;
    var NameIdOut int32     =  6;
    var Fanins [8]int32 = [8]int32{ 2, 3, 4, 5 };

    // create a new module
    var pDesign any  = Ndr_Create( 1 );

    var ModuleID int32  = Ndr_AddModule( pDesign, 1 );

    // add objects to the modele
    Ndr_AddObject( pDesign, ModuleID, ABC_OPER_CI,          0,   0, 0, 0,   0, nil,       1, &NameIdInMode, nil );
    Ndr_AddObject( pDesign, ModuleID, ABC_OPER_CI,          0,   0, 0, 0,   0, nil,       1, &NameIdInCin,  nil );
    Ndr_AddObject( pDesign, ModuleID, ABC_OPER_CI,          0,   2, 0, 0,   0, nil,       1, &NameIdInA,    nil );
    Ndr_AddObject( pDesign, ModuleID, ABC_OPER_CI,          0,   2, 0, 0,   0, nil,       1, &NameIdInB,    nil );
    Ndr_AddObject( pDesign, ModuleID, ABC_OPER_ARI_ADDSUB,  0,   3, 0, 0,   4, Fanins,     1, &NameIdOut,    nil );
    Ndr_AddObject( pDesign, ModuleID, ABC_OPER_CO,          0,   3, 0, 0,   1, &NameIdOut, 0, nil,          nil );

    Ndr_Write( "addsub.ndr", pDesign );
    Ndr_Delete( pDesign );
}

// This testing procedure creates and writes into a Verilog file 
// the following design composed of one lookup table with function of AND2

// module lut_test ( input [1:0] in, output out );
//     assign out = LUT #(TT=4'h8) lut_inst { in[0], in[1], out } ;
// endmodule

func Ndr_ModuleTestLut() {
    // map name IDs into char strings
    //char * ppNames[12] = { NULL, "lut_test", "in", "out" };
    // name IDs
    var NameIdIn int32      =  2;
    var NameIdOut int32     =  3;

    // create a new module
    var pDesign any  = Ndr_Create( 1 );

    var ModuleID int32  = Ndr_AddModule( pDesign, 1 );

    // add objects to the modele
    Ndr_AddObject( pDesign, ModuleID, ABC_OPER_CI,       0,   1, 0, 0,   0, nil,       1, &NameIdIn,     nil );
    Ndr_AddObject( pDesign, ModuleID, ABC_OPER_LUT,      0,   0, 0, 0,   1, &NameIdIn,  1, &NameIdOut,    (char *)(ABC_CONST(0x8)) );
    Ndr_AddObject( pDesign, ModuleID, ABC_OPER_CO,       0,   0, 0, 0,   1, &NameIdOut, 0, nil,          nil );

    Ndr_Write( "lut_test.ndr", pDesign );
    Ndr_Delete( pDesign );
}

//ABC_NAMESPACE_HEADER_END



////////////////////////////////////////////////////////////////////////
///                       END OF FILE                                ///
////////////////////////////////////////////////////////////////////////

