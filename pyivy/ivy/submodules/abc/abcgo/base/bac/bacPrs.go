package bac

import "fmt"

import "os"

/**CFile****************************************************************

  FileName    [bacPrs.h]

  SystemName  [ABC: Logic synthesis and verification system.]

  PackageName [Hierarchical word-level netlist.]

  Synopsis    [Parser declarations.]

  Author      [Alan Mishchenko]
  
  Affiliation [UC Berkeley]

  Date        [Ver. 1.0. Started - November 29, 2014.]

  Revision    [$Id: bacPrs.h,v 1.00 2014/11/29 00:00:00 alanmi Exp $]

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

// parser name types
 type Psr_ManType_t int32
const ( 
    BAC_PRS_NAME Psr_ManType_t = 0
        // 0:  name/variable
    BAC_PRS_SLICE Psr_ManType_t = BAC_PRS_NAME + 1
           // 1:  slice
    BAC_PRS_CONST Psr_ManType_t = BAC_PRS_SLICE + 1
           // 2:  constant
    BAC_PRS_CONCAT Psr_ManType_t = BAC_PRS_CONST + 1
          // 3:  concatentation

)

  

////////////////////////////////////////////////////////////////////////
///                         BASIC TYPES                              ///
////////////////////////////////////////////////////////////////////////

// network
   
type Psr_Ntk_t struct {
    // general info
    iModuleName          int32;
    fMapped uint32; /* C bitfield: 1 bits; review packing and truncation */
    fSlices uint32; /* C bitfield: 1 bits; review packing and truncation */
    fHasC0s uint32; /* C bitfield: 1 bits; review packing and truncation */
    fHasC1s uint32; /* C bitfield: 1 bits; review packing and truncation */
    fHasCXs uint32; /* C bitfield: 1 bits; review packing and truncation */
    fHasCZs uint32; /* C bitfield: 1 bits; review packing and truncation */
    pStrs   *Abc_Nam_t;
    // interface
    vOrder    Vec_Int_t;     // order of signals
    // signal names
    vInouts    Vec_Int_t;    // inouts 
    vInputs    Vec_Int_t;    // inputs 
    vOutputs    Vec_Int_t;   // outputs
    vWires    Vec_Int_t;     // wires  
    // signal ranges
    vInoutsR    Vec_Int_t;   // inouts 
    vInputsR    Vec_Int_t;   // inputs 
    vOutputsR    Vec_Int_t;  // outputs
    vWiresR    Vec_Int_t;    // wires  
    // slices/concatenations/objects
    vSlices    Vec_Int_t;    // NameId + RangeId
    vConcats    Vec_Int_t;   // array of NameId/SliceId/ConstId
    vBoxes    Vec_Int_t;     // ModuleId + InstId + array of pairs {FormNameId, ActSignalId(NameId/SliceId/ConstId/ConcatId)}
    vObjs    Vec_Int_t;      // box handles
}

// parser
   
type Psr_Man_t struct {
    // input data
    pName        string;       // file name
    pBuffer        string;     // file contents
    pLimit        string;      // end of file
    pCur        string;        // current position
    pStrs   *Abc_Nam_t;       // string manager
    pNtk   *Psr_Ntk_t;        // current network
    vNtks   *Vec_Ptr_t;       // input networks
    // temporary data
    vCover    Vec_Str_t;      // one SOP cover
    vTemp    Vec_Int_t;       // array of tokens
    vTemp2    Vec_Int_t;      // array of tokens
    // statistics
    vKnown    Vec_Int_t;
    vFailed    Vec_Int_t;
    vSucceeded    Vec_Int_t;
    // error handling
    fUsingTemp2          int32; // vTemp2 is in use
    ErrorStr [1000]byte;      // error
}

func Psr_ManNtk(vPrs *Vec_Ptr_t, i int32) *Psr_Ntk_t { if i >= 0 && i < Vec_PtrSize(vPrs) {
    return (Psr_Ntk_t *)Vec_PtrEntry(vPrs, i);
} else {
    return nil;
} }
func Psr_ManRoot(vPrs *Vec_Ptr_t) *Psr_Ntk_t { return Psr_ManNtk(vPrs, 0);                             }
func Psr_ManNameMan(vPrs *Vec_Ptr_t) *Abc_Nam_t { return Psr_ManRoot(vPrs).pStrs;                        }

func Psr_NtkId(p *Psr_Ntk_t) int32 { return p.iModuleName;                                  }
func Psr_NtkPioNum(p *Psr_Ntk_t) int32 { return Vec_IntSize(&p.vInouts);                        }
func Psr_NtkPiNum(p *Psr_Ntk_t) int32 { return Vec_IntSize(&p.vInputs);                        }
func Psr_NtkPoNum(p *Psr_Ntk_t) int32 { return Vec_IntSize(&p.vOutputs);                       }
func Psr_NtkBoxNum(p *Psr_Ntk_t) int32 { return Vec_IntSize(&p.vObjs);                          }
func Psr_NtkObjNum(p *Psr_Ntk_t) int32 { return Psr_NtkPioNum(p) + Psr_NtkPiNum(p) + Psr_NtkPoNum(p) + Psr_NtkBoxNum(p); }
func Psr_NtkStr(p *Psr_Ntk_t, h int32) string { return Abc_NamStr(p.pStrs, h);                         }
func Psr_NtkName(p *Psr_Ntk_t) string { return Psr_NtkStr(p, Psr_NtkId(p));                     }
func Psr_NtkSigName(p *Psr_Ntk_t, i int32) int32 { if !p.fSlices { return i; }
 assert(Abc_Lit2Att2(i) == BAC_PRS_NAME, "Abc_Lit2Att2(i) == BAC_PRS_NAME"); return Abc_Lit2Var2(i); }

func Psr_SliceName(p *Psr_Ntk_t, h int32) int32 { return Vec_IntEntry(&p.vSlices, h);                    }
func Psr_SliceRange(p *Psr_Ntk_t, h int32) int32 { return Vec_IntEntry(&p.vSlices, h+1);                  }

func Psr_CatSize(p *Psr_Ntk_t, h int32) int32 { return Vec_IntEntry(&p.vConcats, h);                   }
func Psr_CatArray(p *Psr_Ntk_t, h int32) *int32 { return Vec_IntEntryP(&p.vConcats, h+1);                }
var spStatic_Psr_CatSignals_V Vec_Int_t;
func Psr_CatSignals(p *Psr_Ntk_t, h int32) *Vec_Int_t {  spStatic_Psr_CatSignals_V.nSize = spStatic_Psr_CatSignals_V.nCap = Psr_CatSize(p, h); spStatic_Psr_CatSignals_V.pArray = Psr_CatArray(p, h); return &spStatic_Psr_CatSignals_V; }

func Psr_BoxHand(p *Psr_Ntk_t, i int32) int32 { return Vec_IntEntry(&p.vObjs, i);                      }
func Psr_BoxSize(p *Psr_Ntk_t, i int32) int32 { return Vec_IntEntry(&p.vBoxes, Psr_BoxHand(p, i))-2;   }
func Psr_BoxIONum(p *Psr_Ntk_t, i int32) int32 { return Psr_BoxSize(p, i) / 2;                           }
func Psr_BoxNtk(p *Psr_Ntk_t, i int32) int32 { return Vec_IntEntry(&p.vBoxes, Psr_BoxHand(p, i)+1);   }
func Psr_BoxSetNtk(p *Psr_Ntk_t, i int32, m int32) { Vec_IntWriteEntry(&p.vBoxes, Psr_BoxHand(p, i)+1, m);  }
func Psr_BoxName(p *Psr_Ntk_t, i int32) int32 { return Vec_IntEntry(&p.vBoxes, Psr_BoxHand(p, i)+2);   }
func Psr_BoxIsNode(p *Psr_Ntk_t, i int32) int32 { return!Vec_IntEntry(&p.vBoxes, Psr_BoxHand(p, i)+3);   } // no formal names
func Psr_BoxArray(p *Psr_Ntk_t, i int32) *int32 { return Vec_IntEntryP(&p.vBoxes, Psr_BoxHand(p, i)+3);  }
var spStatic_Psr_BoxSignals_V Vec_Int_t;
func Psr_BoxSignals(p *Psr_Ntk_t, i int32) *Vec_Int_t {  spStatic_Psr_BoxSignals_V.nSize = spStatic_Psr_BoxSignals_V.nCap = Psr_BoxSize(p, i); spStatic_Psr_BoxSignals_V.pArray = Psr_BoxArray(p, i); return &spStatic_Psr_BoxSignals_V; }

// Original macro:
// #define Psr_ManForEachNameVec( vVec, p, pName, i )   \
//     for ( i = 0; (i < Vec_IntSize(vVec)) && ((pName) = Abc_NamStr(p->pStrs, Vec_IntEntry(vVec,i))); i++ )


// Original macro:
// #define Psr_NtkForEachPio( p, NameId, i )            \
//     for ( i = 0; i < Psr_NtkPioNum(p) && ((NameId) = Vec_IntEntry(&p->vInouts, i)); i++ )

// Original macro:
// #define Psr_NtkForEachPi( p, NameId, i )             \
//     for ( i = 0; i < Psr_NtkPiNum(p) && ((NameId) = Vec_IntEntry(&p->vInputs, i)); i++ )

// Original macro:
// #define Psr_NtkForEachPo( p, NameId, i )             \
//     for ( i = 0; i < Psr_NtkPoNum(p) && ((NameId) = Vec_IntEntry(&p->vOutputs, i)); i++ )

// Original macro:
// #define Psr_NtkForEachBox( p, vVec, i )              \
//     for ( i = 0; i < Psr_NtkBoxNum(p) && ((vVec) = Psr_BoxSignals(p, i)); i++ )


////////////////////////////////////////////////////////////////////////
///                      MACRO DEFINITIONS                           ///
////////////////////////////////////////////////////////////////////////

// create error message
func Psr_ManErrorSet(p *Psr_Man_t, pError string, Value int32) int32 {
    assert(!p.ErrorStr[0], "!p.ErrorStr[0]");
    sprintf( p.ErrorStr, "%s", pError );
    return Value;
}
// clear error message
func Psr_ManErrorClear(p *Psr_Man_t) {
    p.ErrorStr[0] = '\0';
}
// print error message
func Psr_ManErrorPrint(p *Psr_Man_t) int32 {
    var pThis string; var iLine int32;
    if !p.ErrorStr[0] { return 1; }
    for  pThis = p.pBuffer; pThis < p.pCur; pThis++  {
        iLine += int32((*pThis == '\n'));
    }
    fmt.Printf("Line %d: %s\n", iLine, p.ErrorStr);
    return 0;
}

// parsing network
func Psr_ManInitializeNtk(p *Psr_Man_t, iName int32, fSlices int32) {
    assert(p.pNtk == nil, "p.pNtk == nil");
    // inlining of macro ABC_CALLOC
    p.pNtk = &Psr_Ntk_t{};
    p.pNtk.iModuleName = iName;
    p.pNtk.fSlices = fSlices;
    p.pNtk.pStrs = Abc_NamRef( p.pStrs );
    Vec_PtrPush( p.vNtks, p.pNtk );
}
func Psr_ManFinalizeNtk(p *Psr_Man_t) {
    assert(p.pNtk != nil, "p.pNtk != nil");
    p.pNtk = nil;
}

// parsing slice/concatentation/box
func Psr_NtkAddSlice(p *Psr_Ntk_t, Name int32, Range int32) int32 {
    var Value int32  = Vec_IntSize(&p.vSlices);
    Vec_IntPushTwo( &p.vSlices, Name, Range );
    return Value;
}
func Psr_NtkAddConcat(p *Psr_Ntk_t, vTemp *Vec_Int_t) int32 {
    var Value int32;
    if !(Vec_IntSize(&p.vConcats) & 1) {
        Vec_IntPush(&p.vConcats, -1); }
    Value = Vec_IntSize(&p.vConcats);
    assert(Value & 1, "Value & 1");
    Vec_IntPush( &p.vConcats, Vec_IntSize(vTemp) );
    Vec_IntAppend( &p.vConcats, vTemp );
    return Value;
}
func Psr_NtkAddBox(p *Psr_Ntk_t, ModName int32, InstName int32, vTemp *Vec_Int_t) {
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


func Psr_ManLoadFile(pFileName string, ppLimit *string) string {
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
func Psr_ManAlloc(pFileName string) *Psr_Man_t {
    var p *Psr_Man_t;
    var pBuffer, pLimit string;
    pBuffer = Psr_ManLoadFile( pFileName, &pLimit );
    if pBuffer == nil {
        return nil; }
    // inlining of macro ABC_CALLOC
    p = &Psr_Man_t{};
    p.pName   = pFileName;
    p.pBuffer = pBuffer;
    p.pLimit  = pLimit;
    p.pCur    = pBuffer;
    p.pStrs   = Abc_NamStart( 1000, 24 );
    p.vNtks   = Vec_PtrAlloc( 100 );
    return p;
}

func Psr_NtkFree(p *Psr_Ntk_t) {
    if p.pStrs {
        Abc_NamDeref( p.pStrs ); }
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
    p = nil;
}

func Psr_ManVecFree(vPrs *Vec_Ptr_t) {
    var pNtk *Psr_Ntk_t; var i int32;
    // inlining of macro Vec_PtrForEachEntry
    for  i = 0; (i < Vec_PtrSize(vPrs)) && (((pNtk) = (Psr_Ntk_t *)Vec_PtrEntry(vPrs, i)), 1); i++  {
        Psr_NtkFree( pNtk );
    }
    Vec_PtrFree( vPrs );
}

func Psr_ManFree(p *Psr_Man_t) {
    if p.pStrs {
        Abc_NamDeref( p.pStrs ); }
    if p.vNtks {
        Psr_ManVecFree( p.vNtks ); }
    // temporary
    Vec_StrErase( &p.vCover );
    Vec_IntErase( &p.vTemp );
    Vec_IntErase( &p.vTemp2 );
    Vec_IntErase( &p.vKnown );
    Vec_IntErase( &p.vFailed );
    Vec_IntErase( &p.vSucceeded );
    p.pBuffer = nil;
    p = nil;
}

func Psr_NtkMemory(p *Psr_Ntk_t) int32 {
    var nMem int32  = sizeof(Psr_Ntk_t);
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
func Psr_ManMemory(vPrs *Vec_Ptr_t) int32 {
    var pNtk *Psr_Ntk_t; var i int32;
    var nMem int32  = Vec_PtrMemory(vPrs);
    // inlining of macro Vec_PtrForEachEntry
    for  i = 0; (i < Vec_PtrSize(vPrs)) && (((pNtk) = (Psr_Ntk_t *)Vec_PtrEntry(vPrs, i)), 1); i++  {
        nMem += Psr_NtkMemory( pNtk );
    }
    nMem += Abc_NamMemUsed(Psr_ManNameMan(vPrs));
    return nMem;
}



////////////////////////////////////////////////////////////////////////
///                             ITERATORS                            ///
////////////////////////////////////////////////////////////////////////

////////////////////////////////////////////////////////////////////////
///                    FUNCTION DECLARATIONS                         ///
////////////////////////////////////////////////////////////////////////

/*=== bac.c ========================================================*/


//ABC_NAMESPACE_HEADER_END



////////////////////////////////////////////////////////////////////////
///                       END OF FILE                                ///
////////////////////////////////////////////////////////////////////////

