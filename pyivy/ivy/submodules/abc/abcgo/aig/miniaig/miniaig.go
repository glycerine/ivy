package miniaig

import "fmt"

import "os"

/**CFile****************************************************************

  FileName    [miniaig.h]

  SystemName  [ABC: Logic synthesis and verification system.]

  PackageName [Minimalistic AIG package.]

  Synopsis    [External declarations.]

  Author      [Alan Mishchenko]
  
  Affiliation [UC Berkeley]

  Date        [Ver. 1.0. Started - September 29, 2012.]

  Revision    [$Id: miniaig.h,v 1.00 2012/09/29 00:00:00 alanmi Exp $]

***********************************************************************/
 



////////////////////////////////////////////////////////////////////////
///                          INCLUDES                                ///
////////////////////////////////////////////////////////////////////////

//#include <stdio.h>
//#include <stdlib.h>
//#include <string.h>
//#include <assert.h>

//ABC_NAMESPACE_HEADER_START

////////////////////////////////////////////////////////////////////////
///                         PARAMETERS                               ///
////////////////////////////////////////////////////////////////////////

#define MINI_AIG_NULL       (0x7FFFFFFF)
#define MINI_AIG_START_SIZE (0x000000FF)

////////////////////////////////////////////////////////////////////////
///                         BASIC TYPES                              ///
////////////////////////////////////////////////////////////////////////

         
type Mini_Aig_t struct {
    nCap           int32;
    nSize           int32;
    nRegs           int32;
    pArray          *int32;
}

////////////////////////////////////////////////////////////////////////
///                      MACRO DEFINITIONS                           ///
////////////////////////////////////////////////////////////////////////

// memory management
// Original macro:
// #define MINI_AIG_ALLOC(type, num)     ((type *) malloc(sizeof(type) * (num)))

// Original macro:
// #define MINI_AIG_CALLOC(type, num)    ((type *) calloc((num), sizeof(type)))

// Original macro:
// #define MINI_AIG_FALLOC(type, num)    ((type *) memset(malloc(sizeof(type) * (num)), 0xff, sizeof(type) * (num)))

// Original macro:
// #define MINI_AIG_FREE(obj)            ((obj) ? (free((char *) (obj)), (obj) = 0) : 0)

// Original macro:
// #define MINI_AIG_REALLOC(type, obj, num) \
//         ((obj) ? ((type *) realloc((char *)(obj), sizeof(type) * (num))) : \
//          ((type *) malloc(sizeof(type) * (num))))


// internal procedures
func Mini_AigGrow(p *Mini_Aig_t, nCapMin int32) {
    if p.nCap >= nCapMin {
        return; }
    // inlining of macro MINI_AIG_REALLOC
    
    if p.pArray {
        p.pArray = ((int *) realloc((char *)(p.pArray), sizeof(int) * (nCapMin)));
    } else {
        p.pArray = ((int *) malloc(sizeof(int) * (nCapMin)));
    } 
    assert(p.pArray, "p.pArray");
    p.nCap   = nCapMin;
}
func Mini_AigPush(p *Mini_Aig_t, Lit0 int32, Lit1 int32) {
    if p.nSize + 2 > p.nCap {
        assert(p.nSize < MINI_AIG_NULL/4, "p.nSize < MINI_AIG_NULL/4");
        if p.nCap < MINI_AIG_START_SIZE {
            Mini_AigGrow( p, MINI_AIG_START_SIZE );
        } else {
            Mini_AigGrow( p, 2 * p.nCap ); }
    }
    p.pArray[p.nSize++] = Lit0;
    p.pArray[p.nSize++] = Lit1;
}

// accessing fanins
func Mini_AigNodeFanin0(p *Mini_Aig_t, Id int32) int32 {
    assert(Id >= 0 && 2*Id < p.nSize, "Id >= 0 && 2*Id < p.nSize");
    assert(p.pArray[2*Id] == 0x7FFFFFFF || p.pArray[2*Id] < 2*Id, "p.pArray[2*Id] == 0x7FFFFFFF || p.pArray[2*Id] < 2*Id");
    return p.pArray[2*Id];
}
func Mini_AigNodeFanin1(p *Mini_Aig_t, Id int32) int32 {
    assert(Id >= 0 && 2*Id < p.nSize, "Id >= 0 && 2*Id < p.nSize");
    assert(p.pArray[2*Id+1] == 0x7FFFFFFF || p.pArray[2*Id+1] < 2*Id, "p.pArray[2*Id+1] == 0x7FFFFFFF || p.pArray[2*Id+1] < 2*Id");
    return p.pArray[2*Id+1];
}

// working with variables and literals
func Mini_AigVar2Lit(Var int32, fCompl int32) int32 { return Var + Var + fCompl;   }
func Mini_AigLit2Var(Lit int32) int32 { return Lit >> 1;             }
func Mini_AigLitIsCompl(Lit int32) int32 { return Lit & 1;              }
func Mini_AigLitNot(Lit int32) int32 { return Lit ^ 1;              }
func Mini_AigLitNotCond(Lit int32, c int32) int32 { return Lit ^ int32((c > 0));   }
func Mini_AigLitRegular(Lit int32) int32 { return Lit & ~01;            }

func Mini_AigLitConst0() int32 { return 0;                    }
func Mini_AigLitConst1() int32 { return 1;                    }
func Mini_AigLitIsConst0(Lit int32) int32 { return Lit == 0;             }
func Mini_AigLitIsConst1(Lit int32) int32 { return Lit == 1;             }
func Mini_AigLitIsConst(Lit int32) int32 { return Lit == 0 || Lit == 1; }

func Mini_AigNodeNum(p *Mini_Aig_t) int32 { return p.nSize/2;           }
func Mini_AigNodeIsConst(p *Mini_Aig_t, Id int32) int32 { assert(Id >= 0, "Id >= 0"); return Id == 0; }
func Mini_AigNodeIsPi(p *Mini_Aig_t, Id int32) int32 { assert(Id >= 0, "Id >= 0"); return Id > 0 && Mini_AigNodeFanin0( p, Id ) == MINI_AIG_NULL; }
func Mini_AigNodeIsPo(p *Mini_Aig_t, Id int32) int32 { assert(Id >= 0, "Id >= 0"); return Id > 0 && Mini_AigNodeFanin0( p, Id ) != MINI_AIG_NULL && Mini_AigNodeFanin1( p, Id ) == MINI_AIG_NULL; }
func Mini_AigNodeIsAnd(p *Mini_Aig_t, Id int32) int32 { assert(Id >= 0, "Id >= 0"); return Id > 0 && Mini_AigNodeFanin0( p, Id ) != MINI_AIG_NULL && Mini_AigNodeFanin1( p, Id ) != MINI_AIG_NULL; }

// working with sequential AIGs
func Mini_AigRegNum(p *Mini_Aig_t) int32 { return p.nRegs;             }
func Mini_AigSetRegNum(p *Mini_Aig_t, n int32) { p.nRegs = n;                }

// iterators through objects
// Original macro:
// #define Mini_AigForEachPi( p, i )  for (i = 1; i < Mini_AigNodeNum(p); i++) if ( !Mini_AigNodeIsPi(p, i) ) {} else 

// Original macro:
// #define Mini_AigForEachPo( p, i )  for (i = 1; i < Mini_AigNodeNum(p); i++) if ( !Mini_AigNodeIsPo(p, i) ) {} else 

// Original macro:
// #define Mini_AigForEachAnd( p, i ) for (i = 1; i < Mini_AigNodeNum(p); i++) if ( !Mini_AigNodeIsAnd(p, i) ) {} else



// constructor/destructor
func Mini_AigStart() *Mini_Aig_t {
    
    // inlining of macro MINI_AIG_CALLOC
    p := &Mini_Aig_t{};
    p.nCap   = MINI_AIG_START_SIZE;
    // inlining of macro MINI_AIG_ALLOC
    p.pArray = ((int *) malloc(sizeof(int) * (p.nCap)));
    Mini_AigPush( p, MINI_AIG_NULL, MINI_AIG_NULL );
    return p;
}
func Mini_AigStop(p *Mini_Aig_t) {
    // inlining of macro MINI_AIG_FREE
    ((p.pArray) ? (free((char *) (p.pArray)), (p.pArray) = 0) : 0);
    // inlining of macro MINI_AIG_FREE
    ((p) ? (free((char *) (p)), (p) = 0) : 0);
}
func Mini_AigPiNum(p *Mini_Aig_t) int32 {
    var i, nPis int32;
    // inlining of macro Mini_AigForEachPi
    for i = 1; i < Mini_AigNodeNum(p); i++ { if !(Mini_AigNodeIsPi(p, i) != 0) {} else { 
        nPis++; }
    }
    return nPis;
}
func Mini_AigPoNum(p *Mini_Aig_t) int32 {
    var i, nPos int32;
    // inlining of macro Mini_AigForEachPo
    for i = 1; i < Mini_AigNodeNum(p); i++ { if !(Mini_AigNodeIsPo(p, i) != 0) {} else { 
        nPos++; }
    }
    return nPos;
}
func Mini_AigAndNum(p *Mini_Aig_t) int32 {
    var i, nNodes int32;
    // inlining of macro Mini_AigForEachAnd
    for i = 1; i < Mini_AigNodeNum(p); i++ { if !(Mini_AigNodeIsAnd(p, i) != 0) {} else {
        nNodes++; }
    }
    return nNodes;
}
func Mini_AigPrintStats(p *Mini_Aig_t) {
    fmt.Printf("PI = %d. PO = %d. Node = %d.\n", Mini_AigPiNum(p), Mini_AigPoNum(p), Mini_AigAndNum(p));
}

// serialization
func Mini_AigDump(p *Mini_Aig_t, pFileName string) {
    var pFile *os.File;
    var RetValue int32;
    pFile = os.OpenFile(pFileName, os.O_WRONLY | os.O_CREATE | os.O_TRUNC, 0666);
    if pFile == nil {
        fmt.Printf("Cannot open file for writing \"%s\".\n", pFileName);
        return;
    }
    RetValue = int32(fwrite( &p.nSize, sizeof(int), 1, pFile ));
    RetValue = int32(fwrite( &p.nRegs, sizeof(int), 1, pFile ));
    RetValue = int32(fwrite( p.pArray, sizeof(int), p.nSize, pFile ));
    _ = (pFile).Close();
}
func Mini_AigLoad(pFileName string) *Mini_Aig_t {
    var p *Mini_Aig_t;
    var pFile *os.File;
    var RetValue, nSize int32;
    pFile = os.Open(pFileName);
    if pFile == nil {
        fmt.Printf("Cannot open file for reading \"%s\".\n", pFileName);
        return nil;
    }
    RetValue = int32(fread( &nSize, sizeof(int), 1, pFile ));
    // inlining of macro MINI_AIG_CALLOC
    p = &Mini_Aig_t{};
    p.nSize = p.nCap = nSize;
    // inlining of macro MINI_AIG_ALLOC
    p.pArray = ((int *) malloc(sizeof(int) * (p.nCap)));
    RetValue = int32(fread( &p.nRegs, sizeof(int), 1, pFile ));
    RetValue = int32(fread( p.pArray, sizeof(int), p.nSize, pFile ));
    _ = (pFile).Close();
    return p;
}


// creating nodes 
// (constant node is created when AIG manager is created)
func Mini_AigCreatePi(p *Mini_Aig_t) int32 {
    var Lit int32  = p.nSize;
    Mini_AigPush( p, MINI_AIG_NULL, MINI_AIG_NULL );
    return Lit;
}
func Mini_AigCreatePo(p *Mini_Aig_t, Lit0 int32) int32 {
    var Lit int32  = p.nSize;
    assert(Lit0 >= 0 && Lit0 < Lit, "Lit0 >= 0 && Lit0 < Lit");
    Mini_AigPush( p, Lit0, MINI_AIG_NULL );
    return Lit;
}

// boolean operations
func Mini_AigAnd(p *Mini_Aig_t, Lit0 int32, Lit1 int32) int32 {
    var Lit int32  = p.nSize;
    assert(Lit0 >= 0 && Lit0 < Lit, "Lit0 >= 0 && Lit0 < Lit");
    assert(Lit1 >= 0 && Lit1 < Lit, "Lit1 >= 0 && Lit1 < Lit");
    Mini_AigPush( p, Lit0, Lit1 );
    return Lit;
}
func Mini_AigOr(p *Mini_Aig_t, Lit0 int32, Lit1 int32) int32 {
    return Mini_AigLitNot( Mini_AigAnd( p, Mini_AigLitNot(Lit0), Mini_AigLitNot(Lit1) ) );
}
func Mini_AigMux(p *Mini_Aig_t, LitC int32, Lit1 int32, Lit0 int32) int32 {
    var Res0 int32  = Mini_AigAnd( p, LitC, Lit1 );
    var Res1 int32  = Mini_AigAnd( p, Mini_AigLitNot(LitC), Lit0 );
    return Mini_AigOr( p, Res0, Res1 );
}
func Mini_AigXor(p *Mini_Aig_t, Lit0 int32, Lit1 int32) int32 {
    return Mini_AigMux( p, Lit0, Mini_AigLitNot(Lit1), Lit1 );
}

var s_MiniTruths5 [5]uint32 = [5]uint32{
    0xAAAAAAAA,
    0xCCCCCCCC,
    0xF0F0F0F0,
    0xFF00FF00,
    0xFFFF0000,
};
func Mini_AigTt5HasVar(t uint32, iVar int32) int32 {
    return ((t << (1<<iVar)) & s_MiniTruths5[iVar]) != (t & s_MiniTruths5[iVar]);
}
func Mini_AigTt5Cofactor0(t uint32, iVar int32) uint32 {
    assert(iVar >= 0 && iVar < 6, "iVar >= 0 && iVar < 6");
    return (t & ~s_MiniTruths5[iVar]) | ((t & ~s_MiniTruths5[iVar]) << (1<<iVar));
}
func Mini_AigTt5Cofactor1(t uint32, iVar int32) uint32 {
    assert(iVar >= 0 && iVar < 6, "iVar >= 0 && iVar < 6");
    return (t & s_MiniTruths5[iVar]) | ((t & s_MiniTruths5[iVar]) >> (1<<iVar));
}
func Mini_AigAndProp(p *Mini_Aig_t, iLit0 int32, iLit1 int32) int32 { 
    if iLit0 < 2 {
        if iLit0 != 0 {
            return iLit1;
        } else {
            return 0;
        } }
    if iLit1 < 2 {
        if iLit1 != 0 {
            return iLit0;
        } else {
            return 0;
        } }
    if iLit0 == iLit1 {
        return iLit1; }
    if iLit0 == Mini_AigLitNot(iLit1) {
        return 0; }
    return Mini_AigAnd( p, iLit0, iLit1 );
}
func Mini_AigMuxProp(p *Mini_Aig_t, iCtrl int32, iData1 int32, iData0 int32) int32 { 
    var iTemp0 int32  = Mini_AigAndProp( p, Mini_AigLitNot(iCtrl), iData0 );
    var iTemp1 int32  = Mini_AigAndProp( p, iCtrl, iData1 );
    return Mini_AigLitNot( Mini_AigAndProp( p, Mini_AigLitNot(iTemp0), Mini_AigLitNot(iTemp1) ) );
}
func Mini_AigTruth(p *Mini_Aig_t, pVarLits *int32, nVars int32, Truth uint32) int32 {
    var Var, Lit0, Lit1 int32; 
    if Truth == 0 {
        return 0; }
    if ~Truth == 0 {
        return 1; }
    assert(nVars > 0, "nVars > 0");
    // find the topmost var
    for  Var = nVars-1; Var >= 0; Var--  {
        if Mini_AigTt5HasVar( Truth, Var ) != 0 {
             break; }
    }
    assert(Var >= 0, "Var >= 0");
    // cofactor
    Lit0 = Mini_AigTruth( p, pVarLits, Var, Mini_AigTt5Cofactor0(Truth, Var) );
    Lit1 = Mini_AigTruth( p, pVarLits, Var, Mini_AigTt5Cofactor1(Truth, Var) );
    return Mini_AigMuxProp( p, pVarLits[Var], Lit1, Lit0 );
}

// procedure to check the topological order during AIG construction
func Mini_AigCheck(p *Mini_Aig_t) int32 {
    var status int32  = 1;
    var i, iFaninLit0, iFaninLit1 int32;
    // inlining of macro Mini_AigForEachAnd
    for i = 1; i < Mini_AigNodeNum(p); i++ { if !(Mini_AigNodeIsAnd(p, i) != 0) {} else {
        // get the fanin literals of this AND node
        iFaninLit0 = Mini_AigNodeFanin0( p, i );
        iFaninLit1 = Mini_AigNodeFanin1( p, i );
        // compare the fanin literals with the literal of the current node (2 * i)
        if iFaninLit0 >= 2 * i {
            fmt.Printf("Fanin0 of AND node %d is not in a topological order.\n", i);
status = 0; }
        if iFaninLit1 >= 2 * i {
            fmt.Printf("Fanin0 of AND node %d is not in a topological order.\n", i);
status = 0; }
    }
    }
    // inlining of macro Mini_AigForEachPo
    for i = 1; i < Mini_AigNodeNum(p); i++ { if !(Mini_AigNodeIsPo(p, i) != 0) {} else {
        // get the fanin literal of this PO node
        iFaninLit0 = Mini_AigNodeFanin0( p, i );
        // compare the fanin literal with the literal of the current node (2 * i)
        if iFaninLit0 >= 2 * i {
            fmt.Printf("Fanin0 of PO node %d is not in a topological order.\n", i);
status = 0; }
    }
    }
    return status;
}

// procedure to dump MiniAIG into a Verilog file
func Mini_AigDumpVerilog(pFileName string, pModuleName string, p *Mini_Aig_t) {
    var i int32;
    var k int32;
    var iFaninLit0 int32;
    var iFaninLit1 int32;
    var Length int32 = len(pModuleName);
    var nPos int32 = Mini_AigPoNum(p); 
    var vObjIsPi *Vec_Bit_t  = Vec_BitStart( Mini_AigNodeNum(p) );
    var pFile *os.File  = os.OpenFile(pFileName, os.O_WRONLY | os.O_CREATE | os.O_TRUNC, 0666);
    if pFile == nil { fmt.Printf("Cannot open output file %s\n", pFileName); return; }
    // write interface
    fmt.Fprintf(pFile, "// This MiniAIG dump was produced by ABC on %s\n\n", Extra_TimeStamp());
    fmt.Fprintf(pFile, "module %s (\n", pModuleName);
    if Mini_AigPiNum(p) > 0 {
        fmt.Fprintf(pFile, "%*sinput wire", Length+10, "");
        k = 0;
        // inlining of macro Mini_AigForEachPi
        for i = 1; i < Mini_AigNodeNum(p); i++ { if !(Mini_AigNodeIsPi(p, i) != 0) {} else {
            if k++ % 12 == 0 { fmt.Fprintf(pFile, "\n%*s", Length+10, ""); }
            fmt.Fprintf(pFile, "i%d, ", i);
            Vec_BitWriteEntry( vObjIsPi, i, 1 );
        }
        }
    }
    fmt.Fprintf(pFile, "\n%*soutput wire", Length+10, "");
    k = 0;
    // inlining of macro Mini_AigForEachPo
    for i = 1; i < Mini_AigNodeNum(p); i++ { if !(Mini_AigNodeIsPo(p, i) != 0) {} else {
        if k++ % 12 == 0 { fmt.Fprintf(pFile, "\n%*s", Length+10, ""); }
        fmt.Fprintf(pFile, "o%d%s", i, k==nPos ? "":", ");
    }
    }
    fmt.Fprintf(pFile, "\n%*s);\n\n", Length+8, "");
    // write LUTs
    // inlining of macro Mini_AigForEachAnd
    for i = 1; i < Mini_AigNodeNum(p); i++ { if !(Mini_AigNodeIsAnd(p, i) != 0) {} else {
        iFaninLit0 = Mini_AigNodeFanin0( p, i );
        iFaninLit1 = Mini_AigNodeFanin1( p, i );
        fmt.Fprintf(pFile, "  assign n%d = ", i);
        fmt.Fprintf(pFile, "%s%c%d", (iFaninLit0 & 1) ? "~":"", Vec_BitEntry(vObjIsPi, iFaninLit0 >> 1) ? 'i':'n', iFaninLit0 >> 1);
        fmt.Fprintf(pFile, " & ");
        fmt.Fprintf(pFile, "%s%c%d", (iFaninLit1 & 1) ? "~":"", Vec_BitEntry(vObjIsPi, iFaninLit1 >> 1) ? 'i':'n', iFaninLit1 >> 1);
        fmt.Fprintf(pFile, ";\n");
    }
    }
    // write assigns
    fmt.Fprintf(pFile, "\n");
    // inlining of macro Mini_AigForEachPo
    for i = 1; i < Mini_AigNodeNum(p); i++ { if !(Mini_AigNodeIsPo(p, i) != 0) {} else {
        iFaninLit0 = Mini_AigNodeFanin0( p, i );
        fmt.Fprintf(pFile, "  assign o%d = ", i);
        fmt.Fprintf(pFile, "%s%c%d", (iFaninLit0 & 1) ? "~":"", Vec_BitEntry(vObjIsPi, iFaninLit0 >> 1) ? 'i':'n', iFaninLit0 >> 1);
        fmt.Fprintf(pFile, ";\n");
    }
    }
    fmt.Fprintf(pFile, "\nendmodule // %s \n\n\n", pModuleName);
    Vec_BitFree( vObjIsPi );
    _ = (pFile).Close();
}

////////////////////////////////////////////////////////////////////////
///                    FUNCTION DECLARATIONS                         ///
////////////////////////////////////////////////////////////////////////

//ABC_NAMESPACE_HEADER_END



////////////////////////////////////////////////////////////////////////
///                       END OF FILE                                ///
////////////////////////////////////////////////////////////////////////

