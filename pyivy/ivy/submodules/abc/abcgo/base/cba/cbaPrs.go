package cba

import (
	"fmt"
	"strings"
)

import "os"

/**CFile****************************************************************

  FileName    [cbaPrs.h]

  SystemName  [ABC: Logic synthesis and verification system.]

  PackageName [Hierarchical word-level netlist.]

  Synopsis    [Parser declarations.]

  Author      [Alan Mishchenko]
  
  Affiliation [UC Berkeley]

  Date        [Ver. 1.0. Started - November 29, 2014.]

  Revision    [$Id: cbaPrs.h,v 1.00 2014/11/29 00:00:00 alanmi Exp $]

***********************************************************************/





////////////////////////////////////////////////////////////////////////
///                          INCLUDES                                ///
////////////////////////////////////////////////////////////////////////

//#include "aig/gia/gia.h"
//#include "misc/util/utilNam.h"

////////////////////////////////////////////////////////////////////////
///                         PARAMETERS                               ///
////////////////////////////////////////////////////////////////////////

//ABC_NAMESPACE_HEADER_START 

// Verilog keywords
 type Cba_VerType_t int32
const ( 
    PRS_VER_NONE Cba_VerType_t = 0
  // 0:  unused
    PRS_VER_INPUT Cba_VerType_t = PRS_VER_NONE + 1
     // 1:  input
    PRS_VER_OUTPUT Cba_VerType_t = PRS_VER_INPUT + 1
    // 2:  output
    PRS_VER_INOUT Cba_VerType_t = PRS_VER_OUTPUT + 1
     // 3:  inout
    PRS_VER_WIRE Cba_VerType_t = PRS_VER_INOUT + 1
      // 4:  wire
    PRS_VER_REG Cba_VerType_t = PRS_VER_WIRE + 1
       // 5:  reg
    PRS_VER_MODULE Cba_VerType_t = PRS_VER_REG + 1
    // 6:  module
    PRS_VER_ASSIGN Cba_VerType_t = PRS_VER_MODULE + 1
    // 7:  assign
    PRS_VER_ALWAYS Cba_VerType_t = PRS_VER_ASSIGN + 1
    // 8:  always
    PRS_VER_FUNCTION Cba_VerType_t = PRS_VER_ALWAYS + 1
  // 9:  function
    PRS_VER_DEFPARAM Cba_VerType_t = PRS_VER_FUNCTION + 1
  // 10: defparam
    PRS_VER_BEGIN Cba_VerType_t = PRS_VER_DEFPARAM + 1
     // 11: begin
    PRS_VER_END Cba_VerType_t = PRS_VER_BEGIN + 1
       // 12: end
    PRS_VER_CASE Cba_VerType_t = PRS_VER_END + 1
      // 13: case
    PRS_VER_ENDCASE Cba_VerType_t = PRS_VER_CASE + 1
   // 14: endcase
    PRS_VER_SIGNED Cba_VerType_t = PRS_VER_ENDCASE + 1
    // 15: signed
    PRS_VER_ENDMODULE Cba_VerType_t = PRS_VER_SIGNED + 1
 // 16: endmodule
    PRS_VER_UNKNOWN Cba_VerType_t = PRS_VER_ENDMODULE + 1    // 17: unknown

)

 

// parser name types
 type Prs_ManType_t int32
const ( 
    CBA_PRS_NAME Prs_ManType_t = 0
  // 0:  name/variable
    CBA_PRS_SLICE Prs_ManType_t = CBA_PRS_NAME + 1
     // 1:  slice
    CBA_PRS_CONST Prs_ManType_t = CBA_PRS_SLICE + 1
     // 2:  constant
    CBA_PRS_CONCAT Prs_ManType_t = CBA_PRS_CONST + 1
    // 3:  concatentation

)

  

////////////////////////////////////////////////////////////////////////
///                         BASIC TYPES                              ///
////////////////////////////////////////////////////////////////////////

// network
   
type Prs_Ntk_t struct {
    // general info
    iModuleName             int32;
    fMapped uint32; /* C bitfield: 1 bits; review packing and truncation */
    fSlices uint32; /* C bitfield: 1 bits; review packing and truncation */
    fHasC0s uint32; /* C bitfield: 1 bits; review packing and truncation */
    fHasC1s uint32; /* C bitfield: 1 bits; review packing and truncation */
    fHasCXs uint32; /* C bitfield: 1 bits; review packing and truncation */
    fHasCZs uint32; /* C bitfield: 1 bits; review packing and truncation */
    pStrs      *Abc_Nam_t;
    pFuns      *Abc_Nam_t;
    vHash  *Hash_IntMan_t;
    // interface
    vOrder       Vec_Int_t;     // order of signals
    // signal names
    vInouts       Vec_Int_t;    // inouts 
    vInputs       Vec_Int_t;    // inputs 
    vOutputs       Vec_Int_t;   // outputs
    vWires       Vec_Int_t;     // wires  
    // signal ranges
    vInoutsR       Vec_Int_t;   // inouts 
    vInputsR       Vec_Int_t;   // inputs 
    vOutputsR       Vec_Int_t;  // outputs
    vWiresR       Vec_Int_t;    // wires  
    // slices/concatenations/objects
    vSlices       Vec_Int_t;    // NameId + RangeId
    vConcats       Vec_Int_t;   // array of NameId/SliceId/ConstId
    vBoxes       Vec_Int_t;     // ModuleId + InstId + array of pairs {FormNameId, ActSignalId(NameId/SliceId/ConstId/ConcatId)}
    vObjs       Vec_Int_t;      // box handles
}

// parser
   
type Prs_Man_t struct {
    // input data
    pName           string;       // file name
    pBuffer           string;     // file contents
    pLimit           string;      // end of file
    pCur           string;        // current position
    pStrs      *Abc_Nam_t;       // string manager
    pFuns      *Abc_Nam_t;       // cover manager
    vHash  *Hash_IntMan_t;       // variable ranges
    pNtk      *Prs_Ntk_t;        // current network
    vNtks      *Vec_Ptr_t;       // input networks
    // temporary data
    vCover       Vec_Str_t;      // one SOP cover
    vTemp       Vec_Int_t;       // array of tokens
    vTemp2       Vec_Int_t;      // array of tokens
    vTemp3       Vec_Int_t;      // array of tokens
    vTemp4       Vec_Int_t;      // array of tokens
    // statistics
    vKnown       Vec_Int_t;
    vFailed       Vec_Int_t;
    vSucceeded       Vec_Int_t;
    // error handling
    nOpens             int32;      // open port counter
    fUsingTemp2             int32; // vTemp2 is in use
    FuncNameId             int32;  // temp value
    FuncRangeId             int32; // temp value
    ErrorStr [1000]byte;         // error
}

func Prs_ManNtk(vPrs *Vec_Ptr_t, i int32) *Prs_Ntk_t { if i >= 0 && i < Vec_PtrSize(vPrs) {
    return (Prs_Ntk_t *)Vec_PtrEntry(vPrs, i);
} else {
    return nil;
} }
func Prs_ManRoot(vPrs *Vec_Ptr_t) *Prs_Ntk_t { return Prs_ManNtk(vPrs, 0);                             }
func Prs_ManNameMan(vPrs *Vec_Ptr_t) *Abc_Nam_t { return Prs_ManRoot(vPrs).pStrs;                        }
func Prs_ManFuncMan(vPrs *Vec_Ptr_t) *Abc_Nam_t { return Prs_ManRoot(vPrs).pFuns;                        }

func Prs_NtkId(p *Prs_Ntk_t) int32 { return p.iModuleName;                                  }
func Prs_NtkPioNum(p *Prs_Ntk_t) int32 { return Vec_IntSize(&p.vInouts);                        }
func Prs_NtkPiNum(p *Prs_Ntk_t) int32 { return Vec_IntSize(&p.vInputs);                        }
func Prs_NtkPoNum(p *Prs_Ntk_t) int32 { return Vec_IntSize(&p.vOutputs);                       }
func Prs_NtkBoxNum(p *Prs_Ntk_t) int32 { return Vec_IntSize(&p.vObjs);                          }
func Prs_NtkObjNum(p *Prs_Ntk_t) int32 { return Prs_NtkPioNum(p) + Prs_NtkPiNum(p) + Prs_NtkPoNum(p) + Prs_NtkBoxNum(p); }
func Prs_NtkStr(p *Prs_Ntk_t, h int32) string { return Abc_NamStr(p.pStrs, h);                         }
func Prs_NtkSop(p *Prs_Ntk_t, h int32) string { return Abc_NamStr(p.pFuns, h);                         }
func Prs_NtkConst(p *Prs_Ntk_t, h int32) string { return Abc_NamStr(p.pFuns, h);                         }
func Prs_NtkName(p *Prs_Ntk_t) string { return Prs_NtkStr(p, Prs_NtkId(p));                     }
func Prs_NtkSigName(p *Prs_Ntk_t, i int32) int32 { if !p.fSlices { return i; }
 assert(Abc_Lit2Att2(i) == CBA_PRS_NAME, "Abc_Lit2Att2(i) == CBA_PRS_NAME"); return Abc_Lit2Var2(i); }
func Ptr_NtkRangeSize(p *Prs_Ntk_t, h int32) int32 { var l int32 = Hash_IntObjData0(p.vHash, h);
var r int32 = Hash_IntObjData1(p.vHash, h); return 1 + (l > r ? l-r : r-l); }

func Prs_SliceName(p *Prs_Ntk_t, h int32) int32 { return Vec_IntEntry(&p.vSlices, h);                    }
func Prs_SliceRange(p *Prs_Ntk_t, h int32) int32 { return Vec_IntEntry(&p.vSlices, h+1);                  }

func Prs_CatSize(p *Prs_Ntk_t, h int32) int32 { return Vec_IntEntry(&p.vConcats, h);                   }
func Prs_CatArray(p *Prs_Ntk_t, h int32) *int32 { return Vec_IntEntryP(&p.vConcats, h+1);                }
var spStatic_Prs_CatSignals_V Vec_Int_t;
func Prs_CatSignals(p *Prs_Ntk_t, h int32) *Vec_Int_t {  spStatic_Prs_CatSignals_V.nSize = spStatic_Prs_CatSignals_V.nCap = Prs_CatSize(p, h); spStatic_Prs_CatSignals_V.pArray = Prs_CatArray(p, h); return &spStatic_Prs_CatSignals_V; }

func Prs_BoxHand(p *Prs_Ntk_t, i int32) int32 { return Vec_IntEntry(&p.vObjs, i);                      }
func Prs_BoxSize(p *Prs_Ntk_t, i int32) int32 { return Vec_IntEntry(&p.vBoxes, Prs_BoxHand(p, i))-2;   }
func Prs_BoxIONum(p *Prs_Ntk_t, i int32) int32 { return Prs_BoxSize(p, i) / 2;                           }
func Prs_BoxNtk(p *Prs_Ntk_t, i int32) int32 { return Vec_IntEntry(&p.vBoxes, Prs_BoxHand(p, i)+1);   }
func Prs_BoxSetNtk(p *Prs_Ntk_t, i int32, m int32) { Vec_IntWriteEntry(&p.vBoxes, Prs_BoxHand(p, i)+1, m);  }
func Prs_BoxName(p *Prs_Ntk_t, i int32) int32 { return Vec_IntEntry(&p.vBoxes, Prs_BoxHand(p, i)+2);   }
func Prs_BoxIsNode(p *Prs_Ntk_t, i int32) int32 { return!Vec_IntEntry(&p.vBoxes, Prs_BoxHand(p, i)+3);   } // no formal names
func Prs_BoxArray(p *Prs_Ntk_t, i int32) *int32 { return Vec_IntEntryP(&p.vBoxes, Prs_BoxHand(p, i)+3);  }
var spStatic_Prs_BoxSignals_V Vec_Int_t;
func Prs_BoxSignals(p *Prs_Ntk_t, i int32) *Vec_Int_t {  spStatic_Prs_BoxSignals_V.nSize = spStatic_Prs_BoxSignals_V.nCap = Prs_BoxSize(p, i); spStatic_Prs_BoxSignals_V.pArray = Prs_BoxArray(p, i); return &spStatic_Prs_BoxSignals_V; }

// Original macro:
// #define Prs_ManForEachNameVec( vVec, p, pName, i )   \
//     for ( i = 0; (i < Vec_IntSize(vVec)) && ((pName) = Abc_NamStr(p->pStrs, Vec_IntEntry(vVec,i))); i++ )


// Original macro:
// #define Prs_NtkForEachPio( p, NameId, i )            \
//     for ( i = 0; i < Prs_NtkPioNum(p) && ((NameId) = Vec_IntEntry(&p->vInouts, i)); i++ )

// Original macro:
// #define Prs_NtkForEachPi( p, NameId, i )             \
//     for ( i = 0; i < Prs_NtkPiNum(p) && ((NameId) = Vec_IntEntry(&p->vInputs, i)); i++ )

// Original macro:
// #define Prs_NtkForEachPo( p, NameId, i )             \
//     for ( i = 0; i < Prs_NtkPoNum(p) && ((NameId) = Vec_IntEntry(&p->vOutputs, i)); i++ )

// Original macro:
// #define Prs_NtkForEachBox( p, vVec, i )              \
//     for ( i = 0; i < Prs_NtkBoxNum(p) && ((vVec) = Prs_BoxSignals(p, i)); i++ )


////////////////////////////////////////////////////////////////////////
///                      MACRO DEFINITIONS                           ///
////////////////////////////////////////////////////////////////////////

// create error message
func Prs_ManErrorSet(p *Prs_Man_t, pError string, Value int32) int32 {
    assert(!p.ErrorStr[0], "!p.ErrorStr[0]");
    sprintf( p.ErrorStr, "%s", pError );
    return Value;
}
// clear error message
func Prs_ManErrorClear(p *Prs_Man_t) {
    p.ErrorStr[0] = '\0';
}
// print error message
func Prs_ManErrorPrint(p *Prs_Man_t) int32 {
    var pThis string; var iLine int32;
    if !p.ErrorStr[0] { return 1; }
    for  pThis = p.pBuffer; pThis < p.pCur; pThis++  {
        iLine += int32((*pThis == '\n'));
    }
    fmt.Printf("Line %d: %s\n", iLine, p.ErrorStr);
    return 0;
}

// parsing network
func Prs_ManInitializeNtk(p *Prs_Man_t, iName int32, fSlices int32) {
    assert(p.pNtk == nil, "p.pNtk == nil");
    // inlining of macro ABC_CALLOC
    p.pNtk = &Prs_Ntk_t{};
    p.pNtk.iModuleName = iName;
    p.pNtk.fSlices = fSlices;
    p.pNtk.pStrs = Abc_NamRef( p.pStrs );
    p.pNtk.pFuns = Abc_NamRef( p.pFuns );
    p.pNtk.vHash = Hash_IntManRef( p.vHash );
    Vec_PtrPush( p.vNtks, p.pNtk );
}
func Prs_ManFinalizeNtk(p *Prs_Man_t) {
    assert(p.pNtk != nil, "p.pNtk != nil");
    p.pNtk = nil;
}

static inline int Prs_ManNewStrId( Prs_Man_t * p, char * format, ...  )
{
    var pStrs *Abc_Nam_t  = p.pStrs;
    var vBuf *Vec_Str_t  = Abc_NamBuffer( pStrs );
    var nAdded int32;
    var nSize int32 = 1000; 
    va_list args;   va_start( args, format );
    Vec_StrGrow( vBuf, Vec_StrSize(vBuf) + nSize );
    nAdded = vsnprintf( Vec_StrLimit(vBuf), nSize, format, args );
    if nAdded > nSize {
        Vec_StrGrow( vBuf, Vec_StrSize(vBuf) + nAdded + nSize );
        nSize = vsnprintf( Vec_StrLimit(vBuf), nAdded, format, args );
        assert(nSize == nAdded, "nSize == nAdded");
    }
    va_end( args );
    return Abc_NamStrFindOrAddLim( pStrs, Vec_StrLimit(vBuf), Vec_StrLimit(vBuf) + nAdded, nil );
}

// parsing slice/concatentation/box
func Prs_NtkAddSlice(p *Prs_Ntk_t, Name int32, Range int32) int32 {
    var Value int32  = Vec_IntSize(&p.vSlices);
    Vec_IntPushTwo( &p.vSlices, Name, Range );
    return Value;
}
func Prs_NtkAddConcat(p *Prs_Ntk_t, vTemp *Vec_Int_t) int32 {
    var Value int32;
    if !(Vec_IntSize(&p.vConcats) & 1) {
        Vec_IntPush(&p.vConcats, -1); }
    Value = Vec_IntSize(&p.vConcats);
    assert(Value & 1, "Value & 1");
    Vec_IntPush( &p.vConcats, Vec_IntSize(vTemp) );
    Vec_IntAppend( &p.vConcats, vTemp );
    return Value;
}
func Prs_NtkAddBox(p *Prs_Ntk_t, ModName int32, InstName int32, vTemp *Vec_Int_t) {
    var Value int32;
    assert(Vec_IntSize(vTemp) % 2 == 0, "Vec_IntSize(vTemp) % 2 == 0");
    if !(Vec_IntSize(&p.vBoxes) & 1) {
        Vec_IntPush(&p.vBoxes, -1); }
    Value = Vec_IntSize(&p.vBoxes);
    assert(Value & 1, "Value & 1");
    Vec_IntPush( &p.vObjs, Value );
    // create entry
    Vec_IntPush( &p.vBoxes, Vec_IntSize(vTemp)+2 );
    Vec_IntPush( &p.vBoxes, ModName );
    Vec_IntPush( &p.vBoxes, InstName );
    Vec_IntAppend( &p.vBoxes, vTemp );
}
func Prs_ManLoadFile(pFileName string, ppLimit *string) string {
    var pBuffer string;
    var nFileSize, RetValue int32;
    var pFile *os.File  = os.Open(pFileName);
    if pFile == nil {
        fmt.Printf("Cannot open input file.\n");
        return nil;
    }
    // get the file size, in bytes
    fseek( pFile, 0, SEEK_END );  
    nFileSize = ftell( pFile );  
    // move the file current reading position to the beginning
    rewind( pFile ); 
    // load the contents of the file into memory
    // inlining of macro ABC_ALLOC
    pBuffer = ((char *) malloc(sizeof(char) * (size_t)(nFileSize + 16)));
    pBuffer[0] = '\n';
    RetValue = fread( pBuffer+1, nFileSize, 1, pFile );
    _ = (pFile).Close();
    // terminate the string with '\0'
    pBuffer[nFileSize + 1] = '\n';
    pBuffer[nFileSize + 2] = '\0';
    *ppLimit = pBuffer + nFileSize + 3;
    return pBuffer;
}
func Prs_ManAlloc(pFileName string) *Prs_Man_t {
    
    // inlining of macro ABC_CALLOC
    p := &Prs_Man_t{};
    if pFileName {
        var pBuffer, pLimit string;
        pBuffer = Prs_ManLoadFile( pFileName, &pLimit );
        if pBuffer == nil {
            return nil; }
        p.pName   = pFileName;
        p.pBuffer = pBuffer;
        p.pLimit  = pLimit;
        p.pCur    = pBuffer;
    }
    p.pStrs = Abc_NamStart( 1000, 24 );
    p.pFuns = Abc_NamStart( 100, 24 );
    p.vHash = Hash_IntManStart( 1000 );
    p.vNtks = Vec_PtrAlloc( 100 );
    return p;
}

func Prs_NtkFree(p *Prs_Ntk_t) {
    if p.pStrs { Abc_NamDeref( p.pStrs ); }
    if p.pFuns { Abc_NamDeref( p.pFuns ); }
    if p.vHash { Hash_IntManDeref( p.vHash ); }
    Vec_IntErase( &p.vOrder );
    Vec_IntErase( &p.vInouts );
    Vec_IntErase( &p.vInputs );
    Vec_IntErase( &p.vOutputs );
    Vec_IntErase( &p.vWires );
    Vec_IntErase( &p.vInoutsR );
    Vec_IntErase( &p.vInputsR );
    Vec_IntErase( &p.vOutputsR );
    Vec_IntErase( &p.vWiresR );
    Vec_IntErase( &p.vSlices );
    Vec_IntErase( &p.vConcats );
    Vec_IntErase( &p.vBoxes );
    Vec_IntErase( &p.vObjs );
    // inlining of macro ABC_FREE
    ((p) ? (free((char *) (p)), (p) = 0) : 0);
}

func Prs_ManVecFree(vPrs *Vec_Ptr_t) {
    var pNtk *Prs_Ntk_t; var i int32;
    // inlining of macro Vec_PtrForEachEntry
    for  i = 0; (i < Vec_PtrSize(vPrs)) && (((pNtk) = (Prs_Ntk_t *)Vec_PtrEntry(vPrs, i)), 1); i++  {
        Prs_NtkFree( pNtk );
    }
    Vec_PtrFree( vPrs );
}

func Prs_ManFree(p *Prs_Man_t) {
    if p.pStrs { Abc_NamDeref( p.pStrs ); }
    if p.pFuns { Abc_NamDeref( p.pFuns ); }
    if p.vHash { Hash_IntManDeref( p.vHash ); }
    if p.vNtks { Prs_ManVecFree( p.vNtks ); }
    // temporary
    Vec_StrErase( &p.vCover );
    Vec_IntErase( &p.vTemp );
    Vec_IntErase( &p.vTemp2 );
    Vec_IntErase( &p.vTemp3 );
    Vec_IntErase( &p.vTemp4 );
    Vec_IntErase( &p.vKnown );
    Vec_IntErase( &p.vFailed );
    Vec_IntErase( &p.vSucceeded );
    // inlining of macro ABC_FREE
    ((p.pBuffer) ? (free((char *) (p.pBuffer)), (p.pBuffer) = 0) : 0);
    // inlining of macro ABC_FREE
    ((p) ? (free((char *) (p)), (p) = 0) : 0);
}

func Prs_NtkMemory(p *Prs_Ntk_t) int32 {
    var nMem int32  = sizeof(Prs_Ntk_t);
    nMem += Vec_IntMemory( &p.vOrder );
    nMem += Vec_IntMemory( &p.vInouts );
    nMem += Vec_IntMemory( &p.vInputs );
    nMem += Vec_IntMemory( &p.vOutputs );
    nMem += Vec_IntMemory( &p.vWires );
    nMem += Vec_IntMemory( &p.vInoutsR );
    nMem += Vec_IntMemory( &p.vInputsR );
    nMem += Vec_IntMemory( &p.vOutputsR );
    nMem += Vec_IntMemory( &p.vWiresR );
    nMem += Vec_IntMemory( &p.vSlices );
    nMem += Vec_IntMemory( &p.vBoxes );
    nMem += Vec_IntMemory( &p.vConcats );
    return nMem;
}
func Prs_ManMemory(vPrs *Vec_Ptr_t) int32 {
    var pNtk *Prs_Ntk_t; var i int32;
    var nMem int32  = Vec_PtrMemory(vPrs);
    // inlining of macro Vec_PtrForEachEntry
    for  i = 0; (i < Vec_PtrSize(vPrs)) && (((pNtk) = (Prs_Ntk_t *)Vec_PtrEntry(vPrs, i)), 1); i++  {
        nMem += Prs_NtkMemory( pNtk );
    }
    nMem += Abc_NamMemUsed(Prs_ManNameMan(vPrs));
    return nMem;
}


/**Function*************************************************************

  Synopsis    [Other APIs.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Ptr_SopToType(pSop string) Cba_ObjType_t {
    if !(int32(strings.Compare(pSop, " 0\n")) != 0) {         return CBA_BOX_CF; }
    if !(int32(strings.Compare(pSop, " 1\n")) != 0) {         return CBA_BOX_CT; }
    if !(int32(strings.Compare(pSop, "1 1\n")) != 0) {        return CBA_BOX_BUF; }
    if !(int32(strings.Compare(pSop, "0 1\n")) != 0) {        return CBA_BOX_INV; }
    if !(int32(strings.Compare(pSop, "11 1\n")) != 0) {       return CBA_BOX_AND; }
    if !(int32(strings.Compare(pSop, "00 1\n")) != 0) {       return CBA_BOX_NOR; }
    if !(int32(strings.Compare(pSop, "00 0\n")) != 0) {       return CBA_BOX_OR; }
    if !(int32(strings.Compare(pSop, "-1 1\n1- 1\n")) != 0) { return CBA_BOX_OR; }
    if !(int32(strings.Compare(pSop, "1- 1\n-1 1\n")) != 0) { return CBA_BOX_OR; }
    if !(int32(strings.Compare(pSop, "01 1\n10 1\n")) != 0) { return CBA_BOX_XOR; }
    if !(int32(strings.Compare(pSop, "10 1\n01 1\n")) != 0) { return CBA_BOX_XOR; }
    if !(int32(strings.Compare(pSop, "11 1\n00 1\n")) != 0) { return CBA_BOX_XNOR; }
    if !(int32(strings.Compare(pSop, "00 1\n11 1\n")) != 0) { return CBA_BOX_XNOR; }
    if !(int32(strings.Compare(pSop, "10 1\n")) != 0) {       return CBA_BOX_SHARP; }
    if !(int32(strings.Compare(pSop, "01 1\n")) != 0) {       return CBA_BOX_SHARPL; }
    assert(false, "0");
    return CBA_OBJ_NONE;
}
func Ptr_SopToTypeName(pSop string) string {
    if !(int32(strings.Compare(pSop, " 0\n")) != 0) {         return "CBA_BOX_C0"; }
    if !(int32(strings.Compare(pSop, " 1\n")) != 0) {         return "CBA_BOX_C1"; }
    if !(int32(strings.Compare(pSop, "1 1\n")) != 0) {        return "CBA_BOX_BUF"; }
    if !(int32(strings.Compare(pSop, "0 1\n")) != 0) {        return "CBA_BOX_INV"; }
    if !(int32(strings.Compare(pSop, "11 1\n")) != 0) {       return "CBA_BOX_AND"; }
    if !(int32(strings.Compare(pSop, "00 1\n")) != 0) {       return "CBA_BOX_NOR"; }
    if !(int32(strings.Compare(pSop, "00 0\n")) != 0) {       return "CBA_BOX_OR"; }
    if !(int32(strings.Compare(pSop, "-1 1\n1- 1\n")) != 0) { return "CBA_BOX_OR"; }
    if !(int32(strings.Compare(pSop, "1- 1\n-1 1\n")) != 0) { return "CBA_BOX_OR"; }
    if !(int32(strings.Compare(pSop, "01 1\n10 1\n")) != 0) { return "CBA_BOX_XOR"; }
    if !(int32(strings.Compare(pSop, "10 1\n01 1\n")) != 0) { return "CBA_BOX_XOR"; }
    if !(int32(strings.Compare(pSop, "11 1\n00 1\n")) != 0) { return "CBA_BOX_XNOR"; }
    if !(int32(strings.Compare(pSop, "00 1\n11 1\n")) != 0) { return "CBA_BOX_XNOR"; }
    if !(int32(strings.Compare(pSop, "10 1\n")) != 0) {       return "CBA_BOX_SHARP"; }
    if !(int32(strings.Compare(pSop, "01 1\n")) != 0) {       return "CBA_BOX_SHARPL"; }
    assert(false, "0");
    return nil;
}
func Ptr_TypeToName(Type Cba_ObjType_t) string {
    if Type == CBA_BOX_CF {    return "const0"; }
    if Type == CBA_BOX_CT {    return "const1"; }
    if Type == CBA_BOX_CX {    return "constX"; }
    if Type == CBA_BOX_CZ {    return "constZ"; }
    if Type == CBA_BOX_BUF {   return "buf"; }
    if Type == CBA_BOX_INV {   return "not"; }
    if Type == CBA_BOX_AND {   return "and"; }
    if Type == CBA_BOX_NAND {  return "nand"; }
    if Type == CBA_BOX_OR {    return "or"; }
    if Type == CBA_BOX_NOR {   return "nor"; }
    if Type == CBA_BOX_XOR {   return "xor"; }
    if Type == CBA_BOX_XNOR {  return "xnor"; }
    if Type == CBA_BOX_MUX {   return "mux"; }
    if Type == CBA_BOX_MAJ {   return "maj"; }
    if Type == CBA_BOX_SHARP { return "sharp"; }
    if Type == CBA_BOX_SHARPL { return "sharpl"; }
    if Type == CBA_BOX_TRI {    return "bufifl"; }
    assert(false, "0");
    return "???";
}
func Ptr_TypeToSop(Type Cba_ObjType_t) string {
    if Type == CBA_BOX_CF {    return " 0\n"; }
    if Type == CBA_BOX_CT {    return " 1\n"; }
    if Type == CBA_BOX_CX {    return " 0\n"; }
    if Type == CBA_BOX_CZ {    return " 0\n"; }
    if Type == CBA_BOX_BUF {   return "1 1\n"; }
    if Type == CBA_BOX_INV {   return "0 1\n"; }
    if Type == CBA_BOX_AND {   return "11 1\n"; }
    if Type == CBA_BOX_NAND {  return "11 0\n"; }
    if Type == CBA_BOX_OR {    return "00 0\n"; }
    if Type == CBA_BOX_NOR {   return "00 1\n"; }
    if Type == CBA_BOX_XOR {   return "01 1\n10 1\n"; }
    if Type == CBA_BOX_XNOR {  return "00 1\n11 1\n"; }
    if Type == CBA_BOX_SHARP { return "10 1\n"; }
    if Type == CBA_BOX_SHARPL { return "01 1\n"; }
    if Type == CBA_BOX_MUX {   return "11- 1\n0-1 1\n"; }
    if Type == CBA_BOX_MAJ {   return "11- 1\n1-1 1\n-11 1\n"; }
    assert(false, "0");
    return "???";
}


////////////////////////////////////////////////////////////////////////
///                             ITERATORS                            ///
////////////////////////////////////////////////////////////////////////

////////////////////////////////////////////////////////////////////////
///                    FUNCTION DECLARATIONS                         ///
////////////////////////////////////////////////////////////////////////

/*=== cbaReadVer.c ========================================================*/
// extern void Prs_NtkAddVerilogDirectives( Prs_Man_t * p );



//ABC_NAMESPACE_HEADER_END



////////////////////////////////////////////////////////////////////////
///                       END OF FILE                                ///
////////////////////////////////////////////////////////////////////////

