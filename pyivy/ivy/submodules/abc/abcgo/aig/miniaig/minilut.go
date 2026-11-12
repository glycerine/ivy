package miniaig

import "fmt"

import "os"

/**CFile****************************************************************

  FileName    [minilut.h]

  SystemName  [ABC: Logic synthesis and verification system.]

  PackageName [Minimalistic representation of LUT mapped network.]

  Synopsis    [External declarations.]

  Author      [Alan Mishchenko]
  
  Affiliation [UC Berkeley]

  Date        [Ver. 1.0. Started - September 29, 2012.]

  Revision    [$Id: minilut.h,v 1.00 2012/09/29 00:00:00 alanmi Exp $]

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

#define MINI_LUT_NULL       (0x7FFFFFFF)
#define MINI_LUT_NULL2      (0x7FFFFFFE)
#define MINI_LUT_START_SIZE (0x000000FF)

////////////////////////////////////////////////////////////////////////
///                         BASIC TYPES                              ///
////////////////////////////////////////////////////////////////////////

         
type Mini_Lut_t struct {
    nCap           int32;
    nSize           int32;
    nRegs           int32;
    LutSize           int32;
    pArray          *int32;
    pTruths     *uint32;
}

////////////////////////////////////////////////////////////////////////
///                      MACRO DEFINITIONS                           ///
////////////////////////////////////////////////////////////////////////

// memory management
// Original macro:
// #define MINI_LUT_ALLOC(type, num)     ((type *) malloc(sizeof(type) * (num)))

// Original macro:
// #define MINI_LUT_CALLOC(type, num)    ((type *) calloc((num), sizeof(type)))

// Original macro:
// #define MINI_LUT_FALLOC(type, num)    ((type *) memset(malloc(sizeof(type) * (num)), 0xff, sizeof(type) * (num)))

// Original macro:
// #define MINI_LUT_FREE(obj)            ((obj) ? (free((char *) (obj)), (obj) = 0) : 0)

// Original macro:
// #define MINI_LUT_REALLOC(type, obj, num) \
//         ((obj) ? ((type *) realloc((char *)(obj), sizeof(type) * (num))) : \
//          ((type *) malloc(sizeof(type) * (num))))


// compute truth table size measured in unsigned's
func Mini_LutWordNum(LutSize int32) int32 {
    if LutSize > 5 {
        return 1 << (LutSize-5);
    } else {
        return 1;
    }
}

// internal procedures
func Mini_LutGrow(p *Mini_Lut_t, nCapMin int32) {
    if p.nCap >= nCapMin {
        return; }
    // inlining of macro MINI_LUT_REALLOC
    
    if p.pArray {
        p.pArray = ((int *) realloc((char *)(p.pArray), sizeof(int) * (nCapMin * p.LutSize)));
    } else {
        p.pArray = ((int *) malloc(sizeof(int) * (nCapMin * p.LutSize)));
    } 
    // inlining of macro MINI_LUT_REALLOC
    
    if p.pTruths {
        p.pTruths = ((unsigned *) realloc((char *)(p.pTruths), sizeof(unsigned) * (nCapMin * Mini_LutWordNum(p.LutSize))));
    } else {
        p.pTruths = ((unsigned *) malloc(sizeof(unsigned) * (nCapMin * Mini_LutWordNum(p.LutSize))));
    } 
    p.nCap   = nCapMin;
    assert(p.pArray, "p.pArray");
    assert(p.pTruths, "p.pTruths");
}
func Mini_LutPush(p *Mini_Lut_t, nVars int32, pVars *int32, pTruth *uint32) {
    var i int32;
    var nWords int32 = Mini_LutWordNum(p.LutSize);
    if p.nSize == p.nCap {
        assert(p.LutSize*p.nSize < MINI_LUT_NULL/2, "p.LutSize*p.nSize < MINI_LUT_NULL/2");
        if p.nCap < MINI_LUT_START_SIZE {
            Mini_LutGrow( p, MINI_LUT_START_SIZE );
        } else {
            Mini_LutGrow( p, 2 * p.nCap ); }
    }
    for  i = 0; i < nVars; i++  {
        assert(pVars[i] >= 0 && pVars[i] < p.nSize, "pVars[i] >= 0 && pVars[i] < p.nSize");
    }
    for  i = 0; i < nVars; i++  {
        p.pArray[p.LutSize * p.nSize + i] = pVars[i];
    }
    for  ; i < p.LutSize; i++  {
        p.pArray[p.LutSize * p.nSize + i] = MINI_LUT_NULL;
    }
    for  i = 0; i < nWords; i++  {
        p.pTruths[nWords * p.nSize + i] = pTruth? pTruth[i] : 0;
    }
    p.nSize++;
}

// accessing fanins
func Mini_LutNodeFanin(p *Mini_Lut_t, Id int32, k int32) int32 {
    assert(Id >= 0 && Id < p.nSize, "Id >= 0 && Id < p.nSize");
    return p.pArray[p.LutSize*Id+k];
}
func Mini_LutNodeTruth(p *Mini_Lut_t, Id int32) *uint32 { 
    assert(Id >= 0 && Id < p.nSize, "Id >= 0 && Id < p.nSize"); 
    return p.pTruths + Id * Mini_LutWordNum(p.LutSize);
}

// working with LUTs
func Mini_LutNodeConst0() int32 { return 0;                    }
func Mini_LutNodeConst1() int32 { return 1;                    }

func Mini_LutNodeNum(p *Mini_Lut_t) int32 { return p.nSize;             }
func Mini_LutNodeIsConst(p *Mini_Lut_t, Id int32) int32 { assert(Id >= 0, "Id >= 0"); return Id == 0 || Id == 1; }
func Mini_LutNodeIsPi(p *Mini_Lut_t, Id int32) int32 { assert(Id >= 0, "Id >= 0"); return Id > 1 && Mini_LutNodeFanin( p, Id, 0 ) == MINI_LUT_NULL; }
func Mini_LutNodeIsPo(p *Mini_Lut_t, Id int32) int32 { assert(Id >= 0, "Id >= 0"); return Id > 1 && Mini_LutNodeFanin( p, Id, 0 ) != MINI_LUT_NULL && Mini_LutNodeFanin( p, Id, 1 ) == MINI_LUT_NULL2; }
func Mini_LutNodeIsNode(p *Mini_Lut_t, Id int32) int32 { assert(Id >= 0, "Id >= 0"); return Id > 1 && Mini_LutNodeFanin( p, Id, 0 ) != MINI_LUT_NULL && Mini_LutNodeFanin( p, Id, 1 ) != MINI_LUT_NULL2; }

func Mini_LutSize(p *Mini_Lut_t) int32 { return p.LutSize;           }

// working with sequential AIGs
func Mini_LutRegNum(p *Mini_Lut_t) int32 { return p.nRegs;             }
func Mini_LutSetRegNum(p *Mini_Lut_t, n int32) { p.nRegs = n;                }

// iterators through objects
// Original macro:
// #define Mini_LutForEachPi( p, i )    for (i = 2; i < Mini_LutNodeNum(p); i++) if ( !Mini_LutNodeIsPi(p, i) )   {} else 

// Original macro:
// #define Mini_LutForEachPo( p, i )    for (i = 2; i < Mini_LutNodeNum(p); i++) if ( !Mini_LutNodeIsPo(p, i) )   {} else 

// Original macro:
// #define Mini_LutForEachNode( p, i )  for (i = 2; i < Mini_LutNodeNum(p); i++) if ( !Mini_LutNodeIsNode(p, i) ) {} else


// iterator through fanins
// Original macro:
// #define Mini_LutForEachFanin( p, i, Fan, k ) for (k = 0; (k < p->LutSize) && (Fan = Mini_LutNodeFanin(p, i, k)) < MINI_LUT_NULL2; k++) 


// constructor/destructor
func Mini_LutStart(LutSize int32) *Mini_Lut_t {
     var i int32;
    assert(LutSize >= 2 && LutSize <= 16, "LutSize >= 2 && LutSize <= 16");
    // inlining of macro MINI_LUT_CALLOC
    p := &Mini_Lut_t{};
    p.LutSize = LutSize;
    p.nCap    = MINI_LUT_START_SIZE;
    // inlining of macro MINI_LUT_ALLOC
    p.pArray  = ((int *) malloc(sizeof(int) * (p.nCap * p.LutSize)));
    // inlining of macro MINI_LUT_ALLOC
    p.pTruths = ((unsigned *) malloc(sizeof(unsigned) * (p.nCap * Mini_LutWordNum(p.LutSize))));
    Mini_LutPush( p, 0, nil, nil ); // const0
    Mini_LutPush( p, 0, nil, nil ); // const1
    for  i = 0; i < Mini_LutWordNum(p.LutSize); i++  {
        p.pTruths[i] = 0;
    }
    for  i = 0; i < Mini_LutWordNum(p.LutSize); i++  {
        p.pTruths[Mini_LutWordNum(p.LutSize) + i] = ~0;
    }
    return p;
}
func Mini_LutStop(p *Mini_Lut_t) {
    // inlining of macro MINI_LUT_FREE
    ((p.pArray) ? (free((char *) (p.pArray)), (p.pArray) = 0) : 0);
    // inlining of macro MINI_LUT_FREE
    ((p.pTruths) ? (free((char *) (p.pTruths)), (p.pTruths) = 0) : 0);
    // inlining of macro MINI_LUT_FREE
    ((p) ? (free((char *) (p)), (p) = 0) : 0);
}
func Mini_LutPrintStats(p *Mini_Lut_t) {
    var i, nPis, nPos, nNodes int32;
    nPis = 0;
    // inlining of macro Mini_LutForEachPi
    for i = 2; i < Mini_LutNodeNum(p); i++ { if !(Mini_LutNodeIsPi(p, i) != 0)   {} else { 
        nPis++; }
    }
    nPos = 0;
    // inlining of macro Mini_LutForEachPo
    for i = 2; i < Mini_LutNodeNum(p); i++ { if !(Mini_LutNodeIsPo(p, i) != 0)   {} else { 
        nPos++; }
    }
    nNodes = 0;
    // inlining of macro Mini_LutForEachNode
    for i = 2; i < Mini_LutNodeNum(p); i++ { if !(Mini_LutNodeIsNode(p, i) != 0) {} else {
        nNodes++; }
    }
    fmt.Printf("PI = %d. PO = %d. LUT = %d.\n", nPis, nPos, nNodes);
}

// serialization
func Mini_LutDump(p *Mini_Lut_t, pFileName string) {
    var pFile *os.File;
    var RetValue int32;
    pFile = os.OpenFile(pFileName, os.O_WRONLY | os.O_CREATE | os.O_TRUNC, 0666);
    if pFile == nil {
        fmt.Printf("Cannot open file for writing \"%s\".\n", pFileName);
        return;
    }
    RetValue = int32(fwrite( &p.nSize,   sizeof(int), 1, pFile ));
    RetValue = int32(fwrite( &p.nRegs,   sizeof(int), 1, pFile ));
    RetValue = int32(fwrite( &p.LutSize, sizeof(int), 1, pFile ));
    RetValue = int32(fwrite( p.pArray,   sizeof(int), p.nSize * p.LutSize, pFile ));
    RetValue = int32(fwrite( p.pTruths,  sizeof(int), p.nSize * Mini_LutWordNum(p.LutSize), pFile ));
    _ = (pFile).Close();
}
func Mini_LutLoad(pFileName string) *Mini_Lut_t {
    var p *Mini_Lut_t;
    var pFile *os.File;
    var RetValue, nSize int32;
    pFile = os.Open(pFileName);
    if pFile == nil {
        fmt.Printf("Cannot open file for reading \"%s\".\n", pFileName);
        return nil;
    }
    RetValue = int32(fread( &nSize, sizeof(int), 1, pFile ));
    // inlining of macro MINI_LUT_CALLOC
    p = &Mini_Lut_t{};
    p.nSize   = p.nCap = nSize;
    RetValue = int32(fread( &p.nRegs,   sizeof(int), 1, pFile ));
    RetValue = int32(fread( &p.LutSize, sizeof(int), 1, pFile ));
    // inlining of macro MINI_LUT_ALLOC
    p.pArray  = ((int *) malloc(sizeof(int) * (p.nCap * p.LutSize)));
    // inlining of macro MINI_LUT_ALLOC
    p.pTruths = ((unsigned *) malloc(sizeof(unsigned) * (p.nCap * Mini_LutWordNum(p.LutSize))));
    RetValue = int32(fread( p.pArray,   sizeof(int), p.nCap * p.LutSize, pFile ));
    RetValue = int32(fread( p.pTruths,  sizeof(int), p.nCap * Mini_LutWordNum(p.LutSize), pFile ));
    _ = (pFile).Close();
    return p;
}


// creating nodes 
// (constant nodes are created when LUT manager is created)
func Mini_LutCreatePi(p *Mini_Lut_t) int32 {
    Mini_LutPush( p, 0, nil, nil );
    return p.nSize - 1;
}
func Mini_LutCreatePo(p *Mini_Lut_t, Var0 int32) int32 {
    assert(Var0 >= 0 && Var0 < p.nSize, "Var0 >= 0 && Var0 < p.nSize");
    Mini_LutPush( p, 1, &Var0, nil );
    // mark PO by setting its 2nd fanin to the special number
    p.pArray[p.LutSize*(p.nSize - 1)+1] = MINI_LUT_NULL2;
    return p.nSize - 1;
}

// create LUT
func Mini_LutCreateNode(p *Mini_Lut_t, nVars int32, pVars *int32, pTruth *uint32) int32 {
    assert(nVars >= 0 && nVars <= p.LutSize, "nVars >= 0 && nVars <= p.LutSize");
    Mini_LutPush( p, nVars, pVars, pTruth );
    return p.nSize - 1;
}

// procedure to check the topological order during AIG construction
func Mini_LutCheck(p *Mini_Lut_t) int32 {
    var status int32  = 1;
    var i, k, iFaninVar int32;
    // inlining of macro Mini_LutForEachNode
    for i = 2; i < Mini_LutNodeNum(p); i++ { if !(Mini_LutNodeIsNode(p, i) != 0) {} else {
        for  k = 0; k < p.LutSize; k++  {
            iFaninVar = Mini_LutNodeFanin( p, i, k );
            if iFaninVar == MINI_LUT_NULL {
                continue; }
            if iFaninVar >= p.LutSize * i {
                fmt.Printf("Fanin %d of LUT node %d is not in a topological order.\n", k, i);
status = 0; }
        }
    }
    }
    // inlining of macro Mini_LutForEachPo
    for i = 2; i < Mini_LutNodeNum(p); i++ { if !(Mini_LutNodeIsPo(p, i) != 0)   {} else {
        iFaninVar = Mini_LutNodeFanin( p, i, 0 );
        if iFaninVar >= p.LutSize * i {
            fmt.Printf("Fanin %d of PO node %d is not in a topological order.\n", k, i);
status = 0; }
    }
    }
    return status;
}



////////////////////////////////////////////////////////////////////////
///                    FUNCTION DECLARATIONS                         ///
////////////////////////////////////////////////////////////////////////

//ABC_NAMESPACE_HEADER_END



////////////////////////////////////////////////////////////////////////
///                       END OF FILE                                ///
////////////////////////////////////////////////////////////////////////

