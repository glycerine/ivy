package if_

import (
	"fmt"
	"math"
)

import "os"

/**CFile****************************************************************

  FileName    [ifDsd.c]

  SystemName  [ABC: Logic synthesis and verification system.]

  PackageName [FPGA mapping based on priority cuts.]

  Synopsis    [Computation of DSD representation.]

  Author      [Alan Mishchenko]
  
  Affiliation [UC Berkeley]

  Date        [Ver. 1.0. Started - November 21, 2006.]

  Revision    [$Id: ifTruth.c,v 1.00 2006/11/21 00:00:00 alanmi Exp $]

***********************************************************************/

//#include <math.h>
//#include "if.h"
//#include "ifCount.h"
//#include "misc/extra/extra.h"
//#include "sat/bsat/satSolver.h"
//#include "aig/gia/gia.h"
//#include "bool/kit/kit.h"

#ifdef ABC_USE_CUDD
#include "bdd/extrab/extraBdd.h"
#endif

#ifdef ABC_USE_PTHREADS

#ifdef _WIN32
#include "../lib/pthread.h"
#else
//#include <pthread.h>
#include <unistd.h>
#endif

#endif

//ABC_NAMESPACE_IMPL_START


////////////////////////////////////////////////////////////////////////
///                        DECLARATIONS                              ///
////////////////////////////////////////////////////////////////////////

const DSD_VERSION = "dsd1"

// network types
 type If_DsdType_t int32
const ( 
    IF_DSD_NONE If_DsdType_t = 0
               // 0:  unknown
    IF_DSD_CONST0 If_DsdType_t = IF_DSD_NONE + 1
                 // 1:  constant
    IF_DSD_VAR If_DsdType_t = IF_DSD_CONST0 + 1
                    // 2:  variable
    IF_DSD_AND If_DsdType_t = IF_DSD_VAR + 1
                    // 3:  AND
    IF_DSD_XOR If_DsdType_t = IF_DSD_AND + 1
                    // 4:  XOR
    IF_DSD_MUX If_DsdType_t = IF_DSD_XOR + 1
                    // 5:  MUX
    IF_DSD_PRIME If_DsdType_t = IF_DSD_MUX + 1                   // 6:  PRIME

)

 

   
type If_DsdObj_t struct {
    Id       uint32;             // node ID
    Type uint32; /* C bitfield: 3 bits; review packing and truncation */   // node type
    nSupp uint32; /* C bitfield: 5 bits; review packing and truncation */   // variable
    fMark uint32; /* C bitfield: 1 bits; review packing and truncation */   // user mark
    Count uint32; /* C bitfield: 18 bits; review packing and truncation */   // variable
    nFans uint32; /* C bitfield: 5 bits; review packing and truncation */   // fanin count
    pFans [0]uint32;       // fanins
}

type If_DsdMan_t_ struct {
    pStore          string;         // input/output file
    nVars            int32;          // max var number
    LutSize            int32;        // LUT size
    nWords            int32;         // word number
    nBins            int32;          // table size
    pBins      *uint32;          // hash table
    pMem    *Mem_Flex_t;           // memory for nodes
    vObjs      Vec_Ptr_t;          // objects
    vNexts      Vec_Int_t;         // next pointers
    vTruths      Vec_Int_t;        // truth IDs of prime nodes
    vTemp1     *Vec_Int_t;         // temp
    vTemp2     *Vec_Int_t;         // temp
    pTtElems         **word;       // elementary TTs
    Vec_Mem_t *    vTtMem[IF_MAX_FUNC_LUTSIZE+1];  // truth table memory and hash table
    Vec_Ptr_t *    vTtDecs[IF_MAX_FUNC_LUTSIZE+1]; // truth table decompositions
    Vec_Wec_t *    vIsops[IF_MAX_FUNC_LUTSIZE+1];  // ISOP for each function
    pSched [IF_MAX_FUNC_LUTSIZE]*int32;    // grey code schedules
    nTtBits            int32;        // the number of truth table bits
    nConfigWords            int32;   // the number of words for config data per node
    vConfigs     *Vec_Wrd_t;       // permutations
    pTtGia     *Gia_Man_t;         // GIA to represent truth tables
    vCover     *Vec_Int_t;         // temporary memory
    pSat          any;           // SAT solver
    pCellStr          string;       // symbolic cell description
    nObjsPrev            int32;      // previous number of objects
    fNewAsUseless            int32;  // set new as useless
    nUniqueHits            int32;    // statistics
    nUniqueMisses            int32;  // statistics
    timeDsd        abctime;        // statistics
    timeCanon        abctime;      // statistics
    timeCheck        abctime;      // statistics
    timeCheck2        abctime;     // statistics
    timeVerify        abctime;     // statistics
}

func If_DsdObjWordNum(nFans int32) int32 { return sizeof(If_DsdObj_t) / 8 + nFans / 2 + ((nFans & 1) > 0);              }
func If_DsdObjTruthId(p *If_DsdMan_t, pObj *If_DsdObj_t) int32 { if pObj.Type == IF_DSD_PRIME && pObj.nFans > 2 {
    return Vec_IntEntry(&p.vTruths, pObj.Id);
} else {
    return -1;
}     }
func If_DsdObjTruth(p *If_DsdMan_t, pObj *If_DsdObj_t) *word { return Vec_MemReadEntry(p.vTtMem[pObj.nFans], If_DsdObjTruthId(p, pObj));  }
func If_DsdObjSetTruth(p *If_DsdMan_t, pObj *If_DsdObj_t, Id int32) { assert(pObj.Type == IF_DSD_PRIME && pObj.nFans > 2, "pObj.Type == IF_DSD_PRIME && pObj.nFans > 2"); Vec_IntWriteEntry(&p.vTruths, pObj.Id, Id); }

func If_DsdObjClean(pObj *If_DsdObj_t) { *pObj = If_DsdObj_t{};                                            }
func If_DsdObjId(pObj *If_DsdObj_t) int32 { return pObj.Id;                                                                   }
func If_DsdObjType(pObj *If_DsdObj_t) int32 { return pObj.Type;                                                                 }
func If_DsdObjIsVar(pObj *If_DsdObj_t) int32 { return int32((pObj.Type == IF_DSD_VAR));                                            }
func If_DsdObjSuppSize(pObj *If_DsdObj_t) int32 { return pObj.nSupp;                                                                }
func If_DsdObjFaninNum(pObj *If_DsdObj_t) int32 { return pObj.nFans;                                                                }
func If_DsdObjFaninC(pObj *If_DsdObj_t, i int32) int32 { assert(i < int32(pObj.nFans), "i < int32(pObj.nFans)"); return Abc_LitIsCompl(pObj.pFans[i]);               }
func If_DsdObjFaninLit(pObj *If_DsdObj_t, i int32) int32 { assert(i < int32(pObj.nFans), "i < int32(pObj.nFans)"); return pObj.pFans[i];                               }

func If_DsdVecObj(p *Vec_Ptr_t, Id int32) *If_DsdObj_t { return (If_DsdObj_t *)Vec_PtrEntry(p, Id);                                         }
func If_DsdVecConst0(p *Vec_Ptr_t) *If_DsdObj_t { return If_DsdVecObj( p, 0 );                                                       }
func If_DsdVecVar(p *Vec_Ptr_t, v int32) *If_DsdObj_t { return If_DsdVecObj( p, v+1 );                                                     }
func If_DsdVecObjSuppSize(p *Vec_Ptr_t, iObj int32) int32 { return If_DsdVecObj( p, iObj ).nSupp;                                             }
func If_DsdVecLitSuppSize(p *Vec_Ptr_t, iLit int32) int32 { return If_DsdVecObjSuppSize( p, Abc_Lit2Var(iLit) );                               }
func If_DsdVecObjRef(p *Vec_Ptr_t, iObj int32) int32 { return If_DsdVecObj( p, iObj ).Count;                                             }
func If_DsdVecObjIncRef(p *Vec_Ptr_t, iObj int32) { if If_DsdVecObjRef(p, iObj) < 0x3FFFF { If_DsdVecObj( p, iObj ).Count++; }        }
func If_DsdObjFanin(p *Vec_Ptr_t, pObj *If_DsdObj_t, i int32) *If_DsdObj_t { assert(i < int32(pObj.nFans), "i < int32(pObj.nFans)"); return If_DsdVecObj(p, Abc_Lit2Var(pObj.pFans[i])); }
func If_DsdVecObjMark(p *Vec_Ptr_t, iObj int32) int32 { return If_DsdVecObj( p, iObj ).fMark;                                             }
func If_DsdVecObjSetMark(p *Vec_Ptr_t, iObj int32) { If_DsdVecObj( p, iObj ).fMark = 1;                                                }
func If_DsdVecObjClearMark(p *Vec_Ptr_t, iObj int32) { If_DsdVecObj( p, iObj ).fMark = 0;                                                }

// Original macro:
// #define If_DsdVecForEachObj( vVec, pObj, i )                \
//     Vec_PtrForEachEntry( If_DsdObj_t *, vVec, pObj, i )

// Original macro:
// #define If_DsdVecForEachObjStart( vVec, pObj, i, Start )    \
//     Vec_PtrForEachEntryStart( If_DsdObj_t *, vVec, pObj, i, Start )

// Original macro:
// #define If_DsdVecForEachObjVec( vNodes, vVec, pObj, i )     \
//     for ( i = 0; (i < Vec_IntSize(vNodes)) && ((pObj) = If_DsdVecObj(vVec, Vec_IntEntry(vNodes,i))); i++ )

// Original macro:
// #define If_DsdVecForEachNode( vVec, pObj, i )               \
//     Vec_PtrForEachEntryStart( If_DsdObj_t *, vVec, pObj, i, 2 )

// Original macro:
// #define If_DsdObjForEachFanin( vVec, pObj, pFanin, i )      \
//     for ( i = 0; (i < If_DsdObjFaninNum(pObj)) && ((pFanin) = If_DsdObjFanin(vVec, pObj, i)); i++ )

// Original macro:
// #define If_DsdObjForEachFaninLit( vVec, pObj, iLit, i )     \
//     for ( i = 0; (i < If_DsdObjFaninNum(pObj)) && ((iLit) = If_DsdObjFaninLit(pObj, i)); i++ )


// extern int Kit_TruthToGia( Gia_Man_t * pMan, unsigned * pTruth, int nVars, Vec_Int_t * vMemory, Vec_Int_t * vLeaves, int fHash );


////////////////////////////////////////////////////////////////////////
///                     FUNCTION DEFINITIONS                         ///
////////////////////////////////////////////////////////////////////////

/**Function*************************************************************

  Synopsis    []

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func If_DsdManFileName(p *If_DsdMan_t) string {
    return p.pStore;
}
func If_DsdManVarNum(p *If_DsdMan_t) int32 {
    return p.nVars;
}
func If_DsdManObjNum(p *If_DsdMan_t) int32 {
    return Vec_PtrSize( &p.vObjs );
}
func If_DsdManLutSize(p *If_DsdMan_t) int32 {
    return p.LutSize;
}
func If_DsdManTtBitNum(p *If_DsdMan_t) int32 {
    return p.nTtBits;
}
func If_DsdManPermBitNum(p *If_DsdMan_t) int32 {
    return (Abc_Base2Log(p.nVars + 1) + 1) * p.nVars;
}
func If_DsdManSetLutSize(p *If_DsdMan_t, nLutSize int32) {
    p.LutSize = nLutSize;
}
func If_DsdManSuppSize(p *If_DsdMan_t, iDsd int32) int32 {
    return If_DsdVecLitSuppSize( &p.vObjs, iDsd );
}
func If_DsdManCheckDec(p *If_DsdMan_t, iDsd int32) int32 {
    return If_DsdVecObjMark( &p.vObjs, Abc_Lit2Var(iDsd) );
}
func If_DsdManReadMark(p *If_DsdMan_t, iDsd int32) int32 {
    return If_DsdVecObjMark( &p.vObjs, Abc_Lit2Var(iDsd) );
}
func If_DsdManSetNewAsUseless(p *If_DsdMan_t) {
    if p.nObjsPrev == 0 {
        p.nObjsPrev = If_DsdManObjNum(p); }
    p.fNewAsUseless = 1;
}
func If_DsdManGetFuncConfig(p *If_DsdMan_t, iDsd int32) *word {
    if p.vConfigs {
        return Vec_WrdEntryP(p.vConfigs, p.nConfigWords * Abc_Lit2Var(iDsd));
    } else {
        return nil;
    }
}
func If_DsdManGetCellStr(p *If_DsdMan_t) string {
    return p.pCellStr;
}

/**Function*************************************************************

  Synopsis    [DSD manager.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
var spStatic_If_ManDsdTtElems_TtElems [DAU_MAX_VAR+1][DAU_MAX_WORD]word;
var spStatic_If_ManDsdTtElems_pTtElems [DAU_MAX_VAR+1]*word = [DAU_MAX_VAR+1]*word{nil};
func If_ManDsdTtElems() **word {
    
    if spStatic_If_ManDsdTtElems_pTtElems[0] == nil {
        var v int32;
        for  v = 0; v <= DAU_MAX_VAR; v++  {
            spStatic_If_ManDsdTtElems_pTtElems[v] = spStatic_If_ManDsdTtElems_TtElems[v];
        }
        Abc_TtElemInit( spStatic_If_ManDsdTtElems_pTtElems, DAU_MAX_VAR );
    }
    return spStatic_If_ManDsdTtElems_pTtElems;
}
func If_DsdObjAlloc(p *If_DsdMan_t, Type int32, nFans int32) *If_DsdObj_t {
    var nWords int32  = If_DsdObjWordNum( nFans );
    var pObj *If_DsdObj_t  = (If_DsdObj_t *)Mem_FlexEntryFetch( p.pMem, sizeof(word) * nWords );
    If_DsdObjClean( pObj );
    pObj.Type   = Type;
    pObj.nFans  = nFans;
    pObj.Id     = Vec_PtrSize( &p.vObjs );
    pObj.fMark  = p.fNewAsUseless;
    pObj.Count  = 0;
    Vec_PtrPush( &p.vObjs, pObj );
    Vec_IntPush( &p.vNexts, 0 );
    Vec_IntPush( &p.vTruths, -1 );
    assert(Vec_IntSize(&p.vNexts) == Vec_PtrSize(&p.vObjs), "Vec_IntSize(&p.vNexts) == Vec_PtrSize(&p.vObjs)");
    assert(Vec_IntSize(&p.vTruths) == Vec_PtrSize(&p.vObjs), "Vec_IntSize(&p.vTruths) == Vec_PtrSize(&p.vObjs)");
    return pObj;
}
func If_DsdManAlloc(nVars int32, LutSize int32) *If_DsdMan_t {
    var p *If_DsdMan_t; var v int32;
    var pFileName string // was [100]byte in C

    assert(nVars <= DAU_MAX_VAR, "nVars <= DAU_MAX_VAR");
    pFileName = fmt.Sprintf("%02d.dsd", nVars);
    p = new(If_DsdMan_t);
    p.pStore  = Abc_UtilStrsav( pFileName );
    p.nVars   = nVars;
    p.LutSize = LutSize;
    p.nWords  = Abc_TtWordNum( nVars );
    p.nBins   = Abc_PrimeCudd( 100000 );
    p.pBins   = ABC_CALLOC( unsigned, p.nBins );
    p.pMem    = Mem_FlexStart();
    p.nConfigWords = 1;
    Vec_PtrGrow( &p.vObjs, 10000 );
    Vec_IntGrow( &p.vNexts, 10000 );
    Vec_IntGrow( &p.vTruths, 10000 );
    If_DsdObjAlloc( p, IF_DSD_CONST0, 0 );
    If_DsdObjAlloc( p, IF_DSD_VAR, 0 ).nSupp = 1;
    p.vTemp1   = Vec_IntAlloc( 32 );
    p.vTemp2   = Vec_IntAlloc( 32 );
    p.pTtElems = If_ManDsdTtElems();
    for  v = 3; v <= nVars; v++  {
        p.vTtMem[v] = Vec_MemAlloc( Abc_TtWordNum(v), 12 );
        Vec_MemHashAlloc( p.vTtMem[v], 10000 );
        p.vTtDecs[v] = Vec_PtrAlloc( 1000 );
    }
/*
    p->pTtGia   = Gia_ManStart( nVars );
    Gia_ManHashAlloc( p->pTtGia );
    for ( v = 0; v < nVars; v++ )
        Gia_ManAppendCi( p->pTtGia );
*/
    for  v = 2; v < nVars; v++  {
        p.pSched[v] = Extra_GreyCodeSchedule( v );
    }
    if LutSize != 0 {
    p.pSat     = If_ManSatBuildXY( LutSize ); }
    p.vCover   = Vec_IntAlloc( 0 );
    return p;
}
func If_DsdManAllocIsops(p *If_DsdMan_t, nLutSize int32) {
    var vLevel *Vec_Int_t;
    var v, i, fCompl int32;
    var pTruth *word;
    if p.vIsops[3] != nil {
        return; }
    if Vec_PtrSize(&p.vObjs) > 2 {
        fmt.Printf("Warning: DSD manager is already started without ISOPs.\n"); }
    for  v = 3; v <= nLutSize; v++  {
        p.vIsops[v] = Vec_WecAlloc( 100 );
        Vec_MemForEachEntry( p.vTtMem[v], pTruth, i )
        {
            vLevel = Vec_WecPushLevel( p.vIsops[v] );
            fCompl = Kit_TruthIsop( (unsigned *)pTruth, v, p.vCover, 1 );
            if fCompl >= 0 && Vec_IntSize(p.vCover) <= 8 {
                Vec_IntGrow( vLevel, Vec_IntSize(p.vCover) );
                Vec_IntAppend( vLevel, p.vCover );
                if fCompl != 0 {
                    vLevel.nCap ^= (1<<16); } // hack to remember complemented attribute
            }
        }
        assert(Vec_WecSize(p.vIsops[v]) == Vec_MemEntryNum(p.vTtMem[v]), "Vec_WecSize(p.vIsops[v]) == Vec_MemEntryNum(p.vTtMem[v])");
    }
}
func If_DsdManFree(p *If_DsdMan_t, fVerbose int32) {
    var v int32;
//    If_DsdManDumpDsd( p );
    if fVerbose != 0 {
        If_DsdManPrint( p, nil, 0, 0, 0, 0, 0 ); }
    if fVerbose != 0 {
        var FileName string // was [10]byte in C

        for  v = 3; v <= p.nVars; v++  {
            FileName = fmt.Sprintf("dumpdsd%02d", v);
            Vec_MemDumpTruthTables( p.vTtMem[v], FileName, v );
        }
    }
    for  v = 2; v < p.nVars; v++  {
        ABC_FREE( p.pSched[v] );
    }
    for  v = 3; v <= p.nVars; v++  {
        Vec_MemHashFree( p.vTtMem[v] );
        Vec_MemFree( p.vTtMem[v] );
        Vec_VecFree( (Vec_Vec_t *)(p.vTtDecs[v]) );
        if p.vIsops[v] {
            Vec_WecFree( p.vIsops[v] ); }
    }
    Vec_WrdFreeP( &p.vConfigs );
    Vec_IntFreeP( &p.vTemp1 );
    Vec_IntFreeP( &p.vTemp2 );
    ABC_FREE( p.vObjs.pArray );
    ABC_FREE( p.vNexts.pArray );
    ABC_FREE( p.vTruths.pArray );
    Mem_FlexStop( p.pMem, 0 );
    Gia_ManStopP( &p.pTtGia );
    Vec_IntFreeP( &p.vCover );
    If_ManSatUnbuild( p.pSat );
    ABC_FREE( p.pCellStr );
    ABC_FREE( p.pStore );
    ABC_FREE( p.pBins );
    ABC_FREE( p );
}
func If_DsdManDumpDsd(p *If_DsdMan_t, Support int32) {
    var pFileName string  = "tts_nondsd.txt";
    var pObj *If_DsdObj_t; 
    var vMap *Vec_Int_t;
    var pFile *os.File  = os.OpenFile(pFileName, os.O_WRONLY | os.O_CREATE | os.O_TRUNC, 0666);
    var v, i int32;
    if pFile == nil {
        fmt.Printf("Cannot open file \"%s\".\n", pFileName);
        return;
    }
    for  v = 3; v <= p.nVars; v++  {
        vMap = Vec_IntStart( Vec_MemEntryNum(p.vTtMem[v]) );
        // inlining of macro If_DsdVecForEachObj
        Vec_PtrForEachEntry( If_DsdObj_t *, &p.vObjs, pObj, i )
        {
            if (Support != 0) && Support != If_DsdObjSuppSize(pObj) {
                continue; }
            if If_DsdObjType(pObj) != IF_DSD_PRIME {
                continue; }
            if Vec_IntEntry(vMap, If_DsdObjTruthId(p, pObj)) {
                continue; }
            Vec_IntWriteEntry(vMap, If_DsdObjTruthId(p, pObj), 1);
            fmt.Fprintf(pFile, "0x");
            Abc_TtPrintHexRev( pFile, If_DsdObjTruth(p, pObj), Support ? Abc_MaxInt(Support, 6) : v );
            fmt.Fprintf(pFile, "\n");
            //printf( "    " );
            //Dau_DsdPrintFromTruth( If_DsdObjTruth(p, pObj), p->nVars );
        }
        Vec_IntFree( vMap );
    }
    _ = (pFile).Close();
}
func If_DsdManDumpAll(p *If_DsdMan_t, Support int32) {
    extern word * If_DsdManComputeTruth( If_DsdMan_t * p, int iDsd, unsigned char * pPermLits );
    var pFileName string  = "tts_all.txt";
    var pObj *If_DsdObj_t;
    var pRes *word; var i int32;
    var pFile *os.File  = os.OpenFile(pFileName, os.O_WRONLY | os.O_CREATE | os.O_TRUNC, 0666);
    if pFile == nil {
        fmt.Printf("Cannot open file \"%s\".\n", pFileName);
        return;
    }
    // inlining of macro If_DsdVecForEachObj
    Vec_PtrForEachEntry( If_DsdObj_t *, &p.vObjs, pObj, i )
    {
        if (Support != 0) && Support != If_DsdObjSuppSize(pObj) {
            continue; }
        pRes = If_DsdManComputeTruth( p, Abc_Var2Lit(i, 0), nil );
        fmt.Fprintf(pFile, "0x");
        Abc_TtPrintHexRev( pFile, pRes, Support ? Abc_MaxInt(Support, 6) : p.nVars );
        fmt.Fprintf(pFile, "\n");
//        printf( "    " );
//        Dau_DsdPrintFromTruth( pRes, p->nVars );
    }
    _ = (pFile).Close();
}
func If_DsdManHasMarks(p *If_DsdMan_t) int32 {
    var pObj *If_DsdObj_t;
    var i int32;
    // inlining of macro If_DsdVecForEachObj
    Vec_PtrForEachEntry( If_DsdObj_t *, &p.vObjs, pObj, i )
        if ( pObj.fMark )
            return 1;
    return 0;
}

/**Function*************************************************************

  Synopsis    [Printing.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func If_DsdManHashProfile(p *If_DsdMan_t) {
    var pObj *If_DsdObj_t;
    var pSpot *uint32;
    var i, Counter int32;
    for  i = 0; i < p.nBins; i++  {
        Counter = 0;
        for  pSpot = p.pBins + i; *pSpot; pSpot = (unsigned *)Vec_IntEntryP(&p.vNexts, pObj.Id), Counter++  {
             pObj = If_DsdVecObj( &p.vObjs, *pSpot );
        }
//        if ( Counter > 5 )
//            printf( "%d ", Counter );
//        if ( i > 10000 )
//            break;
    }
//    printf( "\n" );
}
func If_DsdManCheckNonDec_rec(p *If_DsdMan_t, Id int32) int32 {
    var pObj *If_DsdObj_t;
    var i, iFanin int32;
    pObj = If_DsdVecObj( &p.vObjs, Id );
    if If_DsdObjType(pObj) == IF_DSD_CONST0 {
        return 0; }
    if If_DsdObjType(pObj) == IF_DSD_VAR {
        return 0; }
    if If_DsdObjType(pObj) == IF_DSD_PRIME {
        return 1; }
    // inlining of macro If_DsdObjForEachFaninLit
    for  i = 0; (i < If_DsdObjFaninNum(pObj)) && ((iFanin) = If_DsdObjFaninLit(pObj, i)); i++  {
        if If_DsdManCheckNonDec_rec( p, Abc_Lit2Var(iFanin) ) != 0 {
            return 1; }
    }
    return 0;
}
void If_DsdManPrint_rec( FILE * pFile, If_DsdMan_t * p, int iDsdLit, unsigned char * pPermLits, int * pnSupp )
{
    var OpenType [7]byte = [7]byte{0, 0, 0, '(', '[', '<', '{'};
    var CloseType [7]byte = [7]byte{0, 0, 0, ')', ']', '>', '}'};
    var pObj *If_DsdObj_t;
    var i, iFanin int32;
    fmt.Fprintf(pFile, "%s", Abc_LitIsCompl(iDsdLit) ? "!" : "");
    pObj = If_DsdVecObj( &p.vObjs, Abc_Lit2Var(iDsdLit) );
    if If_DsdObjType(pObj) == IF_DSD_CONST0 { fmt.Fprintf(pFile, "0"); return; }
    if If_DsdObjType(pObj) == IF_DSD_VAR {
        var iPermLit int32;
        if pPermLits {
            iPermLit = int32(pPermLits[(*pnSupp)++]);
        } else {
            iPermLit = Abc_Var2Lit((*pnSupp)++, 0);
        }
        fmt.Fprintf(pFile, "%s%c", Abc_LitIsCompl(iPermLit)? "!":"", 'a' + Abc_Lit2Var(iPermLit));
        return;
    }
    if If_DsdObjType(pObj) == IF_DSD_PRIME {
        Abc_TtPrintHexRev( pFile, If_DsdObjTruth(p, pObj), If_DsdObjFaninNum(pObj) ); }
    fmt.Fprintf(pFile, "%c", OpenType[If_DsdObjType(pObj)]);
    // inlining of macro If_DsdObjForEachFaninLit
    for  i = 0; (i < If_DsdObjFaninNum(pObj)) && ((iFanin) = If_DsdObjFaninLit(pObj, i)); i++  {
        If_DsdManPrint_rec( pFile, p, iFanin, pPermLits, pnSupp );
    }
    fmt.Fprintf(pFile, "%c", CloseType[If_DsdObjType(pObj)]);
}
void If_DsdManPrintOne( FILE * pFile, If_DsdMan_t * p, int iObjId, unsigned char * pPermLits, int fNewLine )
{
    var nSupp int32;
    fmt.Fprintf(pFile, "%6d : ", iObjId);
    fmt.Fprintf(pFile, "%2d ", If_DsdVecObjSuppSize(&p.vObjs, iObjId));
    fmt.Fprintf(pFile, "%8d ", If_DsdVecObjRef(&p.vObjs, iObjId));
    fmt.Fprintf(pFile, "%d  ", If_DsdVecObjMark(&p.vObjs, iObjId));
    If_DsdManPrint_rec( pFile, p, Abc_Var2Lit(iObjId, 0), pPermLits, &nSupp );
    if fNewLine {
        fmt.Fprintf(pFile, "\n"); }
    assert(nSupp == If_DsdVecObjSuppSize(&p.vObjs, iObjId), "nSupp == If_DsdVecObjSuppSize(&p.vObjs, iObjId)");
}
const DSD_ARRAY_LIMIT = 16
func If_DsdManPrintDecs(pFile *os.File, p *If_DsdMan_t) {
    var vDecs *Vec_Int_t;
    var i, k, v, nSuppSize, nDecMax int32;
    var pDecMax [IF_MAX_FUNC_LUTSIZE]int32 ;
    var pCountsAll [IF_MAX_FUNC_LUTSIZE]int32 ;
    var pCountsSSizes [IF_MAX_FUNC_LUTSIZE]int32 ;
    var pCounts [IF_MAX_FUNC_LUTSIZE][DSD_ARRAY_LIMIT+2]int32 = [IF_MAX_FUNC_LUTSIZE][DSD_ARRAY_LIMIT+2]int32{{0}};
    var pTruth *word;
    for  v = 3; v <= p.nVars; v++  {
        assert(Vec_MemEntryNum(p.vTtMem[v]) == Vec_PtrSize(p.vTtDecs[v]), "Vec_MemEntryNum(p.vTtMem[v]) == Vec_PtrSize(p.vTtDecs[v])");
        // find max number of decompositions
        Vec_PtrForEachEntry( Vec_Int_t *, p.vTtDecs[v], vDecs, i )
        {
            pTruth = Vec_MemReadEntry( p.vTtMem[v], i );
            nSuppSize = Abc_TtSupportSize( pTruth, p.nVars );
            pDecMax[nSuppSize] = Abc_MaxInt( pDecMax[nSuppSize], Vec_IntSize(vDecs) );
            nDecMax = Abc_MaxInt( nDecMax, Vec_IntSize(vDecs) );
        }
        // fill up
        Vec_PtrForEachEntry( Vec_Int_t *, p.vTtDecs[v], vDecs, i )
        {
            pTruth = Vec_MemReadEntry( p.vTtMem[v], i );
            nSuppSize = Abc_TtSupportSize( pTruth, p.nVars );
            pCountsAll[nSuppSize]++;
            pCountsSSizes[nSuppSize] += Vec_IntSize(vDecs);
            pCounts[nSuppSize][Abc_MinInt(DSD_ARRAY_LIMIT+1,Vec_IntSize(vDecs))]++;
    //        pCounts[nSuppSize][Abc_MinInt(DSD_ARRAY_LIMIT+1,Vec_IntSize(vDecs)?1+(Vec_IntSize(vDecs)/10):0)]++;
    /*
            if ( nSuppSize == 6 && Vec_IntSize(vDecs) == pDecMax[6] )
            {
                fprintf( pFile, "0x" );
                Abc_TtPrintHex( pTruth, nSuppSize );
                Dau_DecPrintSets( vDecs, nSuppSize );
            }
    */
        }
    }
    // print header
    fmt.Fprintf(pFile, " N :  ");
    fmt.Fprintf(pFile, " Total  ");
    for  k = 0; k <= DSD_ARRAY_LIMIT; k++  {
        fmt.Fprintf(pFile, "%6d", k);
    }
    fmt.Fprintf(pFile, "  ");
    fmt.Fprintf(pFile, "  More");
    fmt.Fprintf(pFile, "     Ave");
    fmt.Fprintf(pFile, "     Max");
    fmt.Fprintf(pFile, "\n");
    // print rows
    for  i = 0; i <= p.nVars; i++  {
        fmt.Fprintf(pFile, "%2d :  ", i);
        fmt.Fprintf(pFile, "%6d  ", pCountsAll[i]);
        for  k = 0; k <= DSD_ARRAY_LIMIT; k++  {
//            fprintf( pFile, "%6d", pCounts[i][k] );
            fmt.Fprintf(pFile, "%6.1f", 100.0*pCounts[i][k]/Abc_MaxInt(1,pCountsAll[i]));
        }
        fmt.Fprintf(pFile, "  ");
//        fprintf( pFile, "%6d", pCounts[i][k] );
        fmt.Fprintf(pFile, "%6.1f", 100.0*pCounts[i][k]/Abc_MaxInt(1,pCountsAll[i]));
        fmt.Fprintf(pFile, "  ");
        fmt.Fprintf(pFile, "%6.1f", 1.0*pCountsSSizes[i]/Abc_MaxInt(1,pCountsAll[i]));
        fmt.Fprintf(pFile, "  ");
        fmt.Fprintf(pFile, "%6d", pDecMax[i]);
        fmt.Fprintf(pFile, "\n");
    }
}
func If_DsdManPrintOccurs(pFile *os.File, p *If_DsdMan_t) {
    var Buffer string // was [100]byte in C

    var pObj *If_DsdObj_t;
    var vOccurs *Vec_Int_t;
    var nOccurs, nOccursMax, nOccursAll int32;
    var i, k, nSizeMax, Counter int32;
    // determine the largest fanin and fanout
    nOccursMax = nOccursAll = 0;
    // inlining of macro If_DsdVecForEachNode
    Vec_PtrForEachEntryStart( If_DsdObj_t *, &p.vObjs, pObj, i, 2 )
    {
        nOccurs = pObj.Count;
        nOccursAll += nOccurs;
        nOccursMax  = Abc_MaxInt( nOccursMax, nOccurs );
    }
    // allocate storage for fanin/fanout numbers
    nSizeMax = 10 * (Abc_Base10Log(nOccursMax) + 1);
    vOccurs  = Vec_IntStart( nSizeMax );
    // count the number of fanins and fanouts
    // inlining of macro If_DsdVecForEachNode
    Vec_PtrForEachEntryStart( If_DsdObj_t *, &p.vObjs, pObj, i, 2 )
    {
        nOccurs = pObj.Count;
        if nOccurs < 10 {
            Vec_IntAddToEntry( vOccurs, nOccurs, 1 );
        } else if nOccurs < 100 {
            Vec_IntAddToEntry( vOccurs, 10 + nOccurs/10, 1 );
        } else if nOccurs < 1000 {
            Vec_IntAddToEntry( vOccurs, 20 + nOccurs/100, 1 );
        } else if nOccurs < 10000 {
            Vec_IntAddToEntry( vOccurs, 30 + nOccurs/1000, 1 );
        } else if nOccurs < 100000 {
            Vec_IntAddToEntry( vOccurs, 40 + nOccurs/10000, 1 );
        } else if nOccurs < 1000000 {
            Vec_IntAddToEntry( vOccurs, 50 + nOccurs/100000, 1 );
        } else if nOccurs < 10000000 {
            Vec_IntAddToEntry( vOccurs, 60 + nOccurs/1000000, 1 ); }
    }
    fmt.Fprintf(pFile, "The distribution of object occurrences:\n");
    for  k = 0; k < nSizeMax; k++  {
        if Vec_IntEntry(vOccurs, k) == 0 {
            continue; }
        if k < 10 {
            fmt.Fprintf(pFile, "%15d : ", k);
        } else {
            Buffer = fmt.Sprintf("%d - %d", int32(math.Pow(float64(float64(10)), float64(k/10))) * (k%10), int32(math.Pow(float64(float64(10)), float64(k/10))) * (k%10+1) - 1);
            fprintf( pFile, "%15s : ", Buffer );
        }
        fmt.Fprintf(pFile, "%12d   ", Vec_IntEntry(vOccurs, k));
        Counter += Vec_IntEntry(vOccurs, k);
        fmt.Fprintf(pFile, "(%6.2f %%)", 100.0*Counter/Vec_PtrSize(&p.vObjs));
        fmt.Fprintf(pFile, "\n");
    }
    Vec_IntFree( vOccurs );
    fmt.Fprintf(pFile, "Fanins: Max = %d. Ave = %.2f.\n", nOccursMax, 1.0*nOccursAll/Vec_PtrSize(&p.vObjs));
}

func If_DsdManPrintDistrib(p *If_DsdMan_t) {
    var pObj *If_DsdObj_t; var i int32;
    var CountObj [IF_MAX_FUNC_LUTSIZE+2]int32 ;
    var CountObjNon [IF_MAX_FUNC_LUTSIZE+2]int32 ;
    var CountObjNpn [IF_MAX_FUNC_LUTSIZE+2]int32 ;
    var CountStr [IF_MAX_FUNC_LUTSIZE+2]int32 ;
    var CountStrNon [IF_MAX_FUNC_LUTSIZE+2]int32 ;
    var CountMarked [IF_MAX_FUNC_LUTSIZE+2]int32 ;
    for  i = 3; i <= p.nVars; i++  {
        CountObjNpn[i] = Vec_MemEntryNum(p.vTtMem[i]);
        CountObjNpn[p.nVars+1] += Vec_MemEntryNum(p.vTtMem[i]);
    }
    // inlining of macro If_DsdVecForEachObj
    Vec_PtrForEachEntry( If_DsdObj_t *, &p.vObjs, pObj, i )
    {
        CountObj[If_DsdObjFaninNum(pObj)]++,        CountObj[p.nVars+1]++;
        if If_DsdObjType(pObj) == IF_DSD_PRIME {
            CountObjNon[If_DsdObjFaninNum(pObj)]++, CountObjNon[p.nVars+1]++; }
        CountStr[If_DsdObjSuppSize(pObj)]++,        CountStr[p.nVars+1]++;
        if If_DsdManCheckNonDec_rec(p, i) != 0 {
            CountStrNon[If_DsdObjSuppSize(pObj)]++, CountStrNon[p.nVars+1]++; }
        if If_DsdVecObjMark(&p.vObjs, i) != 0 {
            CountMarked[If_DsdObjSuppSize(pObj)]++, CountMarked[p.nVars+1]++; }
    }
    fmt.Printf("***** DSD MANAGER STATISTICS *****\n");
    fmt.Printf("Support     ");
    fmt.Printf("Obj   ");
    fmt.Printf("ObjNDSD            ");
    fmt.Printf("NPNNDSD                  ");
    fmt.Printf("Str   ");
    fmt.Printf("StrNDSD             ");
    fmt.Printf("Marked  ");
    fmt.Printf("\n");
    for  i = 0; i <= p.nVars + 1; i++  {
        if i == p.nVars + 1 {
            fmt.Printf("All : ");
        } else {
            fmt.Printf("%3d : ", i); }
        fmt.Printf("%9d ", CountObj[i]);
        fmt.Printf("%9d ", CountObjNon[i]);
        fmt.Printf("%6.2f %% ", 100.0 * CountObjNon[i] / Abc_MaxInt(1, CountObj[i]));
        fmt.Printf("%9d ", CountObjNpn[i]);
        fmt.Printf("%6.2f %% ", 100.0 * CountObjNpn[i] / Abc_MaxInt(1, CountObj[i]));
        fmt.Printf("  ");
        fmt.Printf("%9d ", CountStr[i]);
        fmt.Printf("%9d ", CountStrNon[i]);
        fmt.Printf("%6.2f %% ", 100.0 * CountStrNon[i] / Abc_MaxInt(1, CountStr[i]));
        fmt.Printf("%9d ", CountMarked[i]);
        fmt.Printf("%6.2f %%", 100.0 * CountMarked[i] / Abc_MaxInt(1, CountStr[i]));
        fmt.Printf("\n");
    }
}
func If_DsdManPrint(p *If_DsdMan_t, pFileName string, Number int32, Support int32, fOccurs int32, fTtDump int32, fVerbose int32) {
    var pObj *If_DsdObj_t;
    var vStructs, vCounts *Vec_Int_t;
    var CountUsed, CountNonDsd, CountNonDsdStr, CountMarked, CountPrime int32;
    var i int32;
    var v int32;
    var pPerm *int32;
    var DsdMax int32;
    var MemSizeTTs int32;
    var MemSizeDecs int32;
    var pFile *os.File;
    if pFileName {
        pFile = os.OpenFile(pFileName, os.O_WRONLY | os.O_CREATE | os.O_TRUNC, 0666);
    } else {
        pFile = os.Stdout;
    }
    if pFileName && pFile == nil {
        fmt.Printf("cannot open output file\n");
        return;
    }
    if fVerbose != 0 {
        fmt.Fprintf(pFile, "*****  NOTATIONS USED BELOW  *****\n");
        fmt.Fprintf(pFile, "Support -- the support size\n");
        fmt.Fprintf(pFile, "Obj     -- the number of nodes in the DSD manager for each support size\n");
        fmt.Fprintf(pFile, "           (the constant node and the primary input node have no support)\n");
        fmt.Fprintf(pFile, "ObjNDSD -- the number of prime nodes (that is, nodes whose function has no DSD)\n");
        fmt.Fprintf(pFile, "           (percentage is relative to the number of all nodes of that size)\n");
        fmt.Fprintf(pFile, "NPNNDSD -- the number of different NPN classes of prime nodes\n");
        fmt.Fprintf(pFile, "           (Each NPN class may appear more than once. For example: F1 = 17(ab(cd))\n"); 
        fmt.Fprintf(pFile, "           and F2 = 17(ab[cd]) both have prime majority node (hex TT is 17),\n"); 
        fmt.Fprintf(pFile, "           but in one case the majority node is fed by AND, and in another by XOR.\n");
        fmt.Fprintf(pFile, "           These two majority nodes are different nodes in the DSD manager\n");
        fmt.Fprintf(pFile, "Str     -- the number of structures for each support size\n");
        fmt.Fprintf(pFile, "           (each structure is composed of one or more nodes)\n");
        fmt.Fprintf(pFile, "StrNDSD -- the number of DSD structures containing at least one prime node\n");
        fmt.Fprintf(pFile, "Marked  -- the number of DSD structures matchable with the LUT structure (say, \"44\")\n");
    }
    // inlining of macro If_DsdVecForEachObj
    Vec_PtrForEachEntry( If_DsdObj_t *, &p.vObjs, pObj, i )
    {
        if If_DsdObjType(pObj) == IF_DSD_PRIME {
            DsdMax = Abc_MaxInt( DsdMax, pObj.nFans ); } 
        CountPrime += If_DsdObjType(pObj) == IF_DSD_PRIME;
        CountNonDsdStr += If_DsdManCheckNonDec_rec( p, pObj.Id );
        CountUsed += ( If_DsdVecObjRef(&p.vObjs, pObj.Id) > 0 );
        CountMarked += If_DsdVecObjMark( &p.vObjs, i );
    }
    for  v = 3; v <= p.nVars; v++  {
        CountNonDsd += Vec_MemEntryNum(p.vTtMem[v]);
        MemSizeTTs += Vec_MemEntrySize(p.vTtMem[v]) * Vec_MemEntryNum(p.vTtMem[v]);
        MemSizeDecs += int32(Vec_VecMemoryInt((Vec_Vec_t *)(p.vTtDecs[v])));
    }
    If_DsdManPrintDistrib( p );
    fmt.Printf("Number of inputs = %d.  LUT size = %d.  Marks = %s.  NewAsUseless = %s.  Bookmark = %d.\n", p.nVars, p.LutSize, If_DsdManHasMarks(p)? "yes" : "no", p.fNewAsUseless? "yes" : "no", p.nObjsPrev);
    if p.pCellStr {
        fmt.Printf("Symbolic cell description: %s\n", p.pCellStr); }
    if p.pTtGia {
    fmt.Fprintf(pFile, "Non-DSD AIG nodes          = %8d\n", Gia_ManAndNum(p.pTtGia)); }
    fmt.Fprintf(pFile, "Unique table misses        = %8d\n", p.nUniqueMisses);
    fmt.Fprintf(pFile, "Unique table hits          = %8d\n", p.nUniqueHits);
    fmt.Fprintf(pFile, "Memory used for objects    = %8.2f MB.\n", 1.0*Mem_FlexReadMemUsage(p.pMem)/(1<<20));
    fmt.Fprintf(pFile, "Memory used for functions  = %8.2f MB.\n", 8.0*(MemSizeTTs+sizeof(int)*Vec_IntCap(&p.vTruths))/(1<<20));
    fmt.Fprintf(pFile, "Memory used for hash table = %8.2f MB.\n", 1.0*sizeof(int)*(p.nBins+Vec_IntCap(&p.vNexts))/(1<<20));
    fmt.Fprintf(pFile, "Memory used for bound sets = %8.2f MB.\n", 1.0*MemSizeDecs/(1<<20));
    fmt.Fprintf(pFile, "Memory used for array      = %8.2f MB.\n", 1.0*sizeof(void *)*Vec_PtrCap(&p.vObjs)/(1<<20));
    if p.pTtGia {
    fmt.Fprintf(pFile, "Memory used for AIG        = %8.2f MB.\n", 8.0*Gia_ManAndNum(p.pTtGia)/(1<<20)); }
    if p.timeDsd {
        Abc_PrintTime( 1, "Time DSD   ", p.timeDsd    );
        Abc_PrintTime( 1, "Time canon ", p.timeCanon-p.timeCheck  );
        Abc_PrintTime( 1, "Time check ", p.timeCheck  );
        Abc_PrintTime( 1, "Time check2", p.timeCheck2 );
        Abc_PrintTime( 1, "Time verify", p.timeVerify );
    }
    if fOccurs != 0 {
        If_DsdManPrintOccurs( os.Stdout, p ); }
//    If_DsdManHashProfile( p );
    if fTtDump != 0 {
        If_DsdManDumpDsd( p, Support ); }
    if fTtDump != 0 {
        If_DsdManDumpAll( p, Support ); }
//    If_DsdManPrintDecs( stdout, p );
    if !(fVerbose != 0) {
        return; }
    vStructs = Vec_IntAlloc( 1000 );
    vCounts  = Vec_IntAlloc( 1000 );
    // inlining of macro If_DsdVecForEachObj
    Vec_PtrForEachEntry( If_DsdObj_t *, &p.vObjs, pObj, i )
    {
        if (Number != 0) && i % Number {
            continue; }
        if (Support != 0) && Support != If_DsdObjSuppSize(pObj) {
            continue; }
        Vec_IntPush( vStructs, i );
        Vec_IntPush( vCounts, -int32(pObj.Count) );
//        If_DsdManPrintOne( pFile, p, pObj->Id, NULL, 1 );
    }
//    fprintf( pFile, "\n" );
    pPerm = Abc_MergeSortCost( Vec_IntArray(vCounts), Vec_IntSize(vCounts) );
    for  i = 0; i < Abc_MinInt(Vec_IntSize(vCounts), 20); i++  {
        fmt.Printf("%2d : ", i+1);
        pObj = If_DsdVecObj( &p.vObjs, Vec_IntEntry(vStructs, pPerm[i]) );
        If_DsdManPrintOne( pFile, p, pObj.Id, nil, 1 );
    }
    ABC_FREE( pPerm );
    Vec_IntFree( vStructs );
    Vec_IntFree( vCounts );
    if pFileName {
        _ = (pFile).Close(); }
}
 
/**Function*************************************************************

  Synopsis    [Check if the function is non-trivial.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func If_DsdManCheckNonTriv(p *If_DsdMan_t, Id int32, nVars int32, iVarMax int32) int32 {
    var pObj *If_DsdObj_t; var i, iFanin int32;
    pObj = If_DsdVecObj( &p.vObjs, Id );
    if If_DsdObjType(pObj) == IF_DSD_PRIME {
        return 1; }
    if If_DsdObjFaninNum(pObj) == nVars {
        return 0; }
    // inlining of macro If_DsdObjForEachFaninLit
    for  i = 0; (i < If_DsdObjFaninNum(pObj)) && ((iFanin) = If_DsdObjFaninLit(pObj, i)); i++  {
        if Abc_Lit2Var(iFanin) == 1 && i == iVarMax {
            return 0; }
    }
    return 1;
}

/**Function*************************************************************

  Synopsis    [Sorting DSD literals.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func If_DsdObjCompare(pMan *If_DsdMan_t, p *Vec_Ptr_t, iLit0 int32, iLit1 int32) int32 {
    var p0 *If_DsdObj_t  = If_DsdVecObj(p, Abc_Lit2Var(iLit0));
    var p1 *If_DsdObj_t  = If_DsdVecObj(p, Abc_Lit2Var(iLit1));
    var i, Res int32;
    if If_DsdObjType(p0) < If_DsdObjType(p1) {
        return -1; }
    if If_DsdObjType(p0) > If_DsdObjType(p1) {
        return 1; }
    if If_DsdObjType(p0) < IF_DSD_AND {
        return 0; }
    if If_DsdObjFaninNum(p0) < If_DsdObjFaninNum(p1) {
        return -1; }
    if If_DsdObjFaninNum(p0) > If_DsdObjFaninNum(p1) {
        return 1; }
    if If_DsdObjType(p0) == IF_DSD_PRIME {
        if If_DsdObjTruthId(pMan, p0) < If_DsdObjTruthId(pMan, p1) {
            return -1; }
        if If_DsdObjTruthId(pMan, p0) > If_DsdObjTruthId(pMan, p1) {
            return 1; }
    }
    for  i = 0; i < If_DsdObjFaninNum(p0); i++  {
        Res = If_DsdObjCompare( pMan, p, If_DsdObjFaninLit(p0, i), If_DsdObjFaninLit(p1, i) );
        if Res != 0 {
            return Res; }
    }
    if Abc_LitIsCompl(iLit0) > Abc_LitIsCompl(iLit1) {
        return -1; }
    if Abc_LitIsCompl(iLit0) < Abc_LitIsCompl(iLit1) {
        return 1; }
    assert(iLit0 == iLit1, "iLit0 == iLit1");
    return 0;
}
func If_DsdObjSort(pMan *If_DsdMan_t, p *Vec_Ptr_t, pLits *int32, nLits int32, pPerm *int32) {
    var i, j, best_i int32;
    for  i = 0; i < nLits-1; i++  {
        best_i = i;
        for  j = i+1; j < nLits; j++  {
            if If_DsdObjCompare(pMan, p, pLits[best_i], pLits[j]) == 1 {
                best_i = j; }
        }
        if i == best_i {
            continue; }
        ABC_SWAP( int, pLits[i], pLits[best_i] );
        if pPerm {
            ABC_SWAP( int, pPerm[i], pPerm[best_i] ); }
    }
}

/**Function*************************************************************

  Synopsis    []

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
var spStatic_If_DsdObjHashKey_s_Primes [24]int32 = [24]int32{ 1049, 1297, 1559, 1823, 2089, 2371, 2663, 2909, 
                                3221, 3517, 3779, 4073, 4363, 4663, 4973, 5281, 
                                5573, 5861, 6199, 6481, 6803, 7109, 7477, 7727 };
func If_DsdObjHashKey(p *If_DsdMan_t, Type int32, pLits *int32, nLits int32, truthId int32) uint32 {
    
    var i int32;
    var uHash uint32  = Type * 7873 + nLits * 8147;
    for  i = 0; i < nLits; i++  {
        uHash += pLits[i] * spStatic_If_DsdObjHashKey_s_Primes[i & 0xF];
    }
    if Type == IF_DSD_PRIME {
        uHash += truthId * spStatic_If_DsdObjHashKey_s_Primes[i & 0xF]; }
    return uHash % p.nBins;
}
func If_DsdObjHashLookup(p *If_DsdMan_t, Type int32, pLits *int32, nLits int32, truthId int32) *uint32 {
    var pObj *If_DsdObj_t;
    var pSpot *uint32  = p.pBins + If_DsdObjHashKey(p, Type, pLits, nLits, truthId);
    for  ; *pSpot; pSpot = (unsigned *)Vec_IntEntryP(&p.vNexts, pObj.Id)  {
        pObj = If_DsdVecObj( &p.vObjs, *pSpot );
        if If_DsdObjType(pObj) == Type && 
             If_DsdObjFaninNum(pObj) == nLits && 
             !memcmp(pObj.pFans, pLits, sizeof(int)*If_DsdObjFaninNum(pObj)) &&
             truthId == If_DsdObjTruthId(p, pObj) {
            p.nUniqueHits++;
            return pSpot;
        }
    }
    p.nUniqueMisses++;
    return pSpot;
}
func If_DsdObjHashResize(p *If_DsdMan_t) {
    var pObj *If_DsdObj_t;
    var pSpot *uint32;
    var i int32;
    var Prev int32 = p.nUniqueMisses;
    p.nBins = Abc_PrimeCudd( 2 * p.nBins );
    p.pBins = ABC_REALLOC( unsigned, p.pBins, p.nBins );
    memset( p.pBins, 0, sizeof(unsigned) * p.nBins );
    Vec_IntFill( &p.vNexts, Vec_PtrSize(&p.vObjs), 0 );
    // inlining of macro If_DsdVecForEachNode
    Vec_PtrForEachEntryStart( If_DsdObj_t *, &p.vObjs, pObj, i, 2 )
    {
        pSpot = If_DsdObjHashLookup( p, pObj.Type, (int *)pObj.pFans, pObj.nFans, If_DsdObjTruthId(p, pObj) );
        assert(*pSpot == 0, "*pSpot == 0");
        *pSpot = pObj.Id;
    }
    assert(p.nUniqueMisses - Prev == Vec_PtrSize(&p.vObjs) - 2, "p.nUniqueMisses - Prev == Vec_PtrSize(&p.vObjs) - 2");
    p.nUniqueMisses = Prev;
}

func If_DsdObjCreate(p *If_DsdMan_t, Type int32, pLits *int32, nLits int32, truthId int32) int32 {
    var pObj, pFanin *If_DsdObj_t;
    var i int32;
    var iPrev int32 = -1;
    // check structural canonicity
    assert(Type != DAU_DSD_MUX || nLits == 3, "Type != DAU_DSD_MUX || nLits == 3");
//    assert( Type != DAU_DSD_MUX || !Abc_LitIsCompl(pLits[0]) );
    assert(Type != DAU_DSD_MUX || !Abc_LitIsCompl(pLits[1]) || !Abc_LitIsCompl(pLits[2]), "Type != DAU_DSD_MUX || !Abc_LitIsCompl(pLits[1]) || !Abc_LitIsCompl(pLits[2])");
    // check that leaves are in good order
    if Type == DAU_DSD_AND || Type == DAU_DSD_XOR {
        for  i = 0; i < nLits; i++  {
            pFanin = If_DsdVecObj( &p.vObjs, Abc_Lit2Var(pLits[i]) );
            assert(Type != DAU_DSD_AND || Abc_LitIsCompl(pLits[i]) || If_DsdObjType(pFanin) != DAU_DSD_AND, "Type != DAU_DSD_AND || Abc_LitIsCompl(pLits[i]) || If_DsdObjType(pFanin) != DAU_DSD_AND");
            assert(Type != DAU_DSD_XOR || If_DsdObjType(pFanin) != DAU_DSD_XOR, "Type != DAU_DSD_XOR || If_DsdObjType(pFanin) != DAU_DSD_XOR");
            assert(iPrev == -1 || If_DsdObjCompare(p, &p.vObjs, iPrev, pLits[i]) <= 0, "iPrev == -1 || If_DsdObjCompare(p, &p.vObjs, iPrev, pLits[i]) <= 0");
            iPrev = pLits[i];
        }
    }
    // create new node
    pObj = If_DsdObjAlloc( p, Type, nLits );
    if Type == DAU_DSD_PRIME {
        If_DsdObjSetTruth( p, pObj, truthId ); }
    assert(pObj.nSupp == 0, "pObj.nSupp == 0");
    for  i = 0; i < nLits; i++  {
        pObj.pFans[i] = pLits[i];
        pObj.nSupp += If_DsdVecLitSuppSize(&p.vObjs, pLits[i]);
    }
    // check decomposability
    if p.LutSize && !(If_DsdManCheckXY(p, Abc_Var2Lit(pObj.Id, 0), p.LutSize, 0, 0, 0, 0) != 0) {
        If_DsdVecObjSetMark( &p.vObjs, pObj.Id ); }
    return pObj.Id;
}
func If_DsdObjFindOrAdd(p *If_DsdMan_t, Type int32, pLits *int32, nLits int32, pTruth *word) int32 {
    var PrevSize int32;
    if Type == IF_DSD_PRIME {
        PrevSize = Vec_MemEntryNum( p.vTtMem[nLits] );
    } else {
        PrevSize = -1;
    }   
    var objId int32;
    var truthId int32;
    if Type == IF_DSD_PRIME {
        truthId = Vec_MemHashInsert(p.vTtMem[nLits], pTruth);
    } else {
        truthId = -1;
    }
    var pSpot *uint32  = If_DsdObjHashLookup( p, Type, pLits, nLits, truthId );
//abctime clk;
    if *pSpot {
        return int32(*pSpot); }
//clk = Abc_Clock();
    if p.LutSize && truthId >= 0 && truthId == Vec_PtrSize(p.vTtDecs[nLits]) {
        var vSets *Vec_Int_t  = Dau_DecFindSets_int( pTruth, nLits, p.pSched );
        assert(truthId == Vec_MemEntryNum(p.vTtMem[nLits])-1, "truthId == Vec_MemEntryNum(p.vTtMem[nLits])-1");
        Vec_PtrPush( p.vTtDecs[nLits], vSets );
//        Dau_DecPrintSets( vSets, nLits );
    }
    if p.vIsops[nLits] && truthId >= 0 && PrevSize != Vec_MemEntryNum(p.vTtMem[nLits]) {
        var vLevel *Vec_Int_t  = Vec_WecPushLevel( p.vIsops[nLits] );
        var fCompl int32  = Kit_TruthIsop( (unsigned *)pTruth, nLits, p.vCover, 1 );
        if fCompl >= 0 && Vec_IntSize(p.vCover) <= 8 {
            Vec_IntGrow( vLevel, Vec_IntSize(p.vCover) );
            Vec_IntAppend( vLevel, p.vCover );
            if fCompl != 0 {
                vLevel.nCap ^= (1<<16); } // hack to remember complemented attribute
        }
        assert(Vec_WecSize(p.vIsops[nLits]) == Vec_MemEntryNum(p.vTtMem[nLits]), "Vec_WecSize(p.vIsops[nLits]) == Vec_MemEntryNum(p.vTtMem[nLits])");
    }
    if p.pTtGia && truthId >= 0 && truthId == Vec_MemEntryNum(p.vTtMem[nLits])-1 {
//        int nObjOld = Gia_ManAndNum(p->pTtGia);
        var Lit int32  = Kit_TruthToGia( p.pTtGia, (unsigned *)pTruth, nLits, p.vCover, nil, 1 );
//        printf( "%d ", Gia_ManAndNum(p->pTtGia)-nObjOld );
        Gia_ManAppendCo( p.pTtGia, Lit );
    }
//p->timeCheck += Abc_Clock() - clk;
    *pSpot = Vec_PtrSize( &p.vObjs );
    objId = If_DsdObjCreate( p, Type, pLits, nLits, truthId );
    if Vec_PtrSize(&p.vObjs) > p.nBins {
        If_DsdObjHashResize( p ); }
    return objId;
}


/**Function*************************************************************

  Synopsis    [Saving/loading DSD manager.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func If_DsdManSave(p *If_DsdMan_t, pFileName string) {
    var pObj *If_DsdObj_t; 
    var vSets *Vec_Int_t;
    var pTruth *word; 
    var i, v, Num int32;
    var pFile *os.File  = os.OpenFile(pFileName ? pFileName : p.pStore, os.O_WRONLY | os.O_CREATE | os.O_TRUNC, 0666);
    if pFile == nil {
        fmt.Printf("Writing DSD manager file \"%s\" has failed.\n", pFileName ? pFileName : p.pStore);
        return;
    }
    fwrite( DSD_VERSION, 4, 1, pFile );
    Num = p.nVars;
    fwrite( &Num, 4, 1, pFile );
    Num = p.LutSize;
    fwrite( &Num, 4, 1, pFile );
    Num = Vec_PtrSize(&p.vObjs);
    fwrite( &Num, 4, 1, pFile );
    Vec_PtrForEachEntryStart( If_DsdObj_t *, &p.vObjs, pObj, i, 2 )
    {
        Num = If_DsdObjWordNum( pObj.nFans );
        fwrite( &Num, 4, 1, pFile );
        fwrite( pObj, sizeof(word)*Num, 1, pFile );
        if pObj.Type == IF_DSD_PRIME {
            fwrite( Vec_IntEntryP(&p.vTruths, i), 4, 1, pFile ); }
    }
    for  v = 3; v <= p.nVars; v++  {
        var nBytes int32  = sizeof(word)*Vec_MemEntrySize(p.vTtMem[v]);
        Num = Vec_MemEntryNum(p.vTtMem[v]);
        fwrite( &Num, 4, 1, pFile );
        Vec_MemForEachEntry( p.vTtMem[v], pTruth, i )
            fwrite( pTruth, nBytes, 1, pFile );
        Num = Vec_PtrSize(p.vTtDecs[v]);
        fwrite( &Num, 4, 1, pFile );
        Vec_PtrForEachEntry( Vec_Int_t *, p.vTtDecs[v], vSets, i )
        {
            Num = Vec_IntSize(vSets);
            fwrite( &Num, 4, 1, pFile );
            fwrite( Vec_IntArray(vSets), sizeof(int)*Num, 1, pFile );
        }
    }
    Num = p.nConfigWords;
    fwrite( &Num, 4, 1, pFile );
    Num = p.nTtBits;
    fwrite( &Num, 4, 1, pFile );
    if p.vConfigs {
        Num = Vec_WrdSize(p.vConfigs);
    } else {
        Num = 0;
    }
    fwrite( &Num, 4, 1, pFile );
    if Num != 0 {
        fwrite( Vec_WrdArray(p.vConfigs), sizeof(word)*Num, 1, pFile ); }
    if p.pCellStr {
        Num = len(p.pCellStr);
    } else {
        Num = 0;
    }
    fwrite( &Num, 4, 1, pFile );
    if Num != 0 {
        fwrite( p.pCellStr, sizeof(char)*Num, 1, pFile ); }
    _ = (pFile).Close();
}
func If_DsdManLoad(pFileName string) *If_DsdMan_t {
    var p *If_DsdMan_t;
    var pObj *If_DsdObj_t; 
    var vSets *Vec_Int_t;
    var pBuffer [10]byte;
    var pSpot *uint32;
    var pTruth *word;
    var i, v, Num, Num2, RetValue int32;
    var pFile *os.File  = os.Open(pFileName);
    if pFile == nil {
        fmt.Printf("Reading DSD manager file \"%s\" has failed.\n", pFileName);
        return nil;
    }
    RetValue = fread( pBuffer, 4, 1, pFile );
    if strncmp(pBuffer, DSD_VERSION, len(DSD_VERSION)) {
        fmt.Printf("Unrecognized format of file \"%s\".\n", pFileName);
        return nil;
    }
    RetValue = fread( &Num, 4, 1, pFile );
    p = If_DsdManAlloc( Num, 0 );
    ABC_FREE( p.pStore );
    p.pStore = Abc_UtilStrsav( pFileName );
    RetValue = fread( &Num, 4, 1, pFile );
    p.LutSize = Num;
    p.pSat  = If_ManSatBuildXY( p.LutSize );
    RetValue = fread( &Num, 4, 1, pFile );
    assert(Num >= 2, "Num >= 2");
    Vec_PtrFillExtra( &p.vObjs, Num, nil );
    Vec_IntFill( &p.vNexts, Num, 0 );
    Vec_IntFill( &p.vTruths, Num, -1 );
    p.nBins = Abc_PrimeCudd( 2*Num );
    p.pBins = ABC_REALLOC( unsigned, p.pBins, p.nBins );
    memset( p.pBins, 0, sizeof(unsigned) * p.nBins );
    for  i = 2; i < Vec_PtrSize(&p.vObjs); i++  {
        RetValue = fread( &Num, 4, 1, pFile );
        pObj = (If_DsdObj_t *)Mem_FlexEntryFetch( p.pMem, sizeof(word) * Num );
        RetValue = fread( pObj, sizeof(word)*Num, 1, pFile );
        Vec_PtrWriteEntry( &p.vObjs, i, pObj );
        if pObj.Type == IF_DSD_PRIME {
            RetValue = fread( &Num, 4, 1, pFile );
            Vec_IntWriteEntry( &p.vTruths, i, Num );
        }
        pSpot = If_DsdObjHashLookup( p, pObj.Type, (int *)pObj.pFans, pObj.nFans, If_DsdObjTruthId(p, pObj) );
        assert(*pSpot == 0, "*pSpot == 0");
        *pSpot = pObj.Id;
    }
    assert(p.nUniqueMisses == Vec_PtrSize(&p.vObjs) - 2, "p.nUniqueMisses == Vec_PtrSize(&p.vObjs) - 2");
    p.nUniqueMisses = 0;
    pTruth = ABC_ALLOC( word, p.nWords );
    for  v = 3; v <= p.nVars; v++  {
        var nBytes int32  = sizeof(word)*Vec_MemEntrySize(p.vTtMem[v]);
        RetValue = fread( &Num, 4, 1, pFile );
        for  i = 0; i < Num; i++  {
            RetValue = fread( pTruth, nBytes, 1, pFile );
            Vec_MemHashInsert( p.vTtMem[v], pTruth );
        }
        assert(Num == Vec_MemEntryNum(p.vTtMem[v]), "Num == Vec_MemEntryNum(p.vTtMem[v])");
        RetValue = fread( &Num2, 4, 1, pFile );
        for  i = 0; i < Num2; i++  {
            RetValue = fread( &Num, 4, 1, pFile );
            vSets = Vec_IntAlloc( Num );
            RetValue = fread( Vec_IntArray(vSets), sizeof(int)*Num, 1, pFile );
            vSets.nSize = Num;
            Vec_PtrPush( p.vTtDecs[v], vSets );
        }
        assert(Num2 == Vec_PtrSize(p.vTtDecs[v]), "Num2 == Vec_PtrSize(p.vTtDecs[v])"); 
    }
    ABC_FREE( pTruth );
    RetValue = fread( &Num, 4, 1, pFile );
    p.nConfigWords = Num;
    RetValue = fread( &Num, 4, 1, pFile );
    p.nTtBits = Num;
    RetValue = fread( &Num, 4, 1, pFile );
    if (RetValue != 0) && (Num != 0) {
        p.vConfigs = Vec_WrdStart( Num );
        RetValue = fread( Vec_WrdArray(p.vConfigs), sizeof(word)*Num, 1, pFile );
    }
    RetValue = fread( &Num, 4, 1, pFile );
    if (RetValue != 0) && (Num != 0) {
        p.pCellStr = ABC_CALLOC( char, Num + 1 );
        RetValue = fread( p.pCellStr, sizeof(char)*Num, 1, pFile );
    }
    _ = (pFile).Close();
    return p;
}
func If_DsdManMerge(p *If_DsdMan_t, pNew *If_DsdMan_t) {
    var pObj *If_DsdObj_t; 
    var vMap *Vec_Int_t;
    var pFanins [DAU_MAX_VAR]int32;
    var i, k, iFanin, Id int32;
    if p.nVars < pNew.nVars {
        fmt.Printf("The number of variables should be the same or smaller.\n");
        return;
    }
    if p.LutSize != pNew.LutSize {
        fmt.Printf("LUT size should be the same.\n");
        return;
    }
    assert(p.nTtBits == pNew.nTtBits, "p.nTtBits == pNew.nTtBits");
    assert(p.nConfigWords == pNew.nConfigWords, "p.nConfigWords == pNew.nConfigWords");
    if If_DsdManHasMarks(p) != If_DsdManHasMarks(pNew) {
        fmt.Printf("Warning! Old manager has %smarks while new manager has %smarks.\n", If_DsdManHasMarks(p) ? "" : "no ", If_DsdManHasMarks(pNew) ? "" : "no "); }
    vMap = Vec_IntAlloc( Vec_PtrSize(&pNew.vObjs) );
    Vec_IntPush( vMap, 0 );
    Vec_IntPush( vMap, 1 );
    if p.vConfigs && pNew.vConfigs {
        Vec_WrdFillExtra( p.vConfigs, p.nConfigWords * (Vec_PtrSize(&p.vObjs) + Vec_PtrSize(&pNew.vObjs)), 0 ); }
    // inlining of macro If_DsdVecForEachNode
    Vec_PtrForEachEntryStart( If_DsdObj_t *, &pNew.vObjs, pObj, i, 2 )
    {
        // inlining of macro If_DsdObjForEachFaninLit
        for  k = 0; (k < If_DsdObjFaninNum(pObj)) && ((iFanin) = If_DsdObjFaninLit(pObj, k)); k++  {
            pFanins[k] = Abc_Lit2LitV( Vec_IntArray(vMap), iFanin );
        }
        Id = If_DsdObjFindOrAdd( p, pObj.Type, pFanins, pObj.nFans, pObj.Type == IF_DSD_PRIME ? If_DsdObjTruth(pNew, pObj) : nil );
        if pObj.fMark {
            If_DsdVecObjSetMark( &p.vObjs, Id ); }
        if p.vConfigs && pNew.vConfigs && p.nConfigWords * i < Vec_WrdSize(pNew.vConfigs) {
            //Vec_WrdFillExtra( p->vConfigs, Id, Vec_WrdEntry(pNew->vConfigs, i) );
            var pConfigNew *word  = Vec_WrdEntryP(pNew.vConfigs, p.nConfigWords * i);
            var pConfigOld *word  = Vec_WrdEntryP(p.vConfigs, p.nConfigWords * Id);
            memcpy( pConfigOld, pConfigNew, sizeof(word) * p.nConfigWords );
        }
        Vec_IntPush( vMap, Id );
    }
    assert(Vec_IntSize(vMap) == Vec_PtrSize(&pNew.vObjs), "Vec_IntSize(vMap) == Vec_PtrSize(&pNew.vObjs)");
    Vec_IntFree( vMap );
    if p.vConfigs && pNew.vConfigs {
        Vec_WrdShrink( p.vConfigs, p.nConfigWords * Vec_PtrSize(&p.vObjs) ); }
}
func If_DsdManCleanOccur(p *If_DsdMan_t, fVerbose int32) {
    var pObj *If_DsdObj_t; 
    var i int32;
    // inlining of macro If_DsdVecForEachObj
    Vec_PtrForEachEntry( If_DsdObj_t *, &p.vObjs, pObj, i )
        pObj.Count = 0;
}
func If_DsdManCleanMarks(p *If_DsdMan_t, fVerbose int32) {
    var pObj *If_DsdObj_t; 
    var i int32;
    ABC_FREE( p.pCellStr );
    Vec_WrdFreeP( &p.vConfigs );
    // inlining of macro If_DsdVecForEachObj
    Vec_PtrForEachEntry( If_DsdObj_t *, &p.vObjs, pObj, i )
        pObj.fMark = 0;
}
func If_DsdManInvertMarks(p *If_DsdMan_t, fVerbose int32) {
    var pObj *If_DsdObj_t; 
    var i int32;
    ABC_FREE( p.pCellStr );
    //Vec_WrdFreeP( &p->vConfigs );
    // inlining of macro If_DsdVecForEachObj
    Vec_PtrForEachEntry( If_DsdObj_t *, &p.vObjs, pObj, i )
        pObj.fMark = !pObj.fMark;
}
func If_DsdManFilter_rec(pNew *If_DsdMan_t, p *If_DsdMan_t, i int32, vMap *Vec_Int_t) {
    var pObj *If_DsdObj_t;
    var pFanins [DAU_MAX_VAR]int32;
    var k, iFanin, Id int32;
    if Vec_IntEntry(vMap, i) >= 0 {
        return; }
    // call recursively
    pObj = If_DsdVecObj( &p.vObjs, i );
    // inlining of macro If_DsdObjForEachFaninLit
    for  k = 0; (k < If_DsdObjFaninNum(pObj)) && ((iFanin) = If_DsdObjFaninLit(pObj, k)); k++  {
        If_DsdManFilter_rec( pNew, p, Abc_Lit2Var(iFanin), vMap );
    }
    // duplicate this one
    // inlining of macro If_DsdObjForEachFaninLit
    for  k = 0; (k < If_DsdObjFaninNum(pObj)) && ((iFanin) = If_DsdObjFaninLit(pObj, k)); k++  {
        pFanins[k] = Abc_Lit2LitV( Vec_IntArray(vMap), iFanin );
    }
    Id = If_DsdObjFindOrAdd( pNew, pObj.Type, pFanins, pObj.nFans, pObj.Type == IF_DSD_PRIME ? If_DsdObjTruth(p, pObj) : nil );
    if pObj.fMark {
        If_DsdVecObjSetMark( &pNew.vObjs, Id ); }
    If_DsdVecObj( &pNew.vObjs, Id ).Count = pObj.Count;
    // save the result
    Vec_IntWriteEntry( vMap, i, Id );
}
func If_DsdManFilter(p *If_DsdMan_t, Limit int32) *If_DsdMan_t {
    var pNew *If_DsdMan_t  = If_DsdManAlloc( p.nVars, p.LutSize );
    var pObj *If_DsdObj_t; 
    var vMap *Vec_Int_t;
    var i int32;
    vMap = Vec_IntStartFull( Vec_PtrSize(&p.vObjs) );
    Vec_IntWriteEntry( vMap, 0, 0 );
    Vec_IntWriteEntry( vMap, 1, 1 );
    // inlining of macro If_DsdVecForEachNode
    Vec_PtrForEachEntryStart( If_DsdObj_t *, &p.vObjs, pObj, i, 2 )
        if ( int32(pObj.Count) >= Limit )
            If_DsdManFilter_rec( pNew, p, i, vMap );
    Vec_IntFree( vMap );
    return pNew;
}

/**Function*************************************************************

  Synopsis    [Collect nodes of the tree.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func If_DsdManCollect_rec(p *If_DsdMan_t, Id int32, vNodes *Vec_Int_t, vFirsts *Vec_Int_t, pnSupp *int32) {
    var i, iFanin, iFirst int32;
    var pObj *If_DsdObj_t  = If_DsdVecObj( &p.vObjs, Id );
    if If_DsdObjType(pObj) == IF_DSD_CONST0 {
        return; }
    if If_DsdObjType(pObj) == IF_DSD_VAR {
        (*pnSupp)++;
        return;
    }
    iFirst = *pnSupp;
    // inlining of macro If_DsdObjForEachFaninLit
    for  i = 0; (i < If_DsdObjFaninNum(pObj)) && ((iFanin) = If_DsdObjFaninLit(pObj, i)); i++  {
        If_DsdManCollect_rec( p, Abc_Lit2Var(iFanin), vNodes, vFirsts, pnSupp );
    }
    Vec_IntPush( vNodes, Id );
    Vec_IntPush( vFirsts, iFirst );
}
func If_DsdManCollect(p *If_DsdMan_t, Id int32, vNodes *Vec_Int_t, vFirsts *Vec_Int_t) {
    var nSupp int32;
    Vec_IntClear( vNodes );
    Vec_IntClear( vFirsts );
    If_DsdManCollect_rec( p, Id, vNodes, vFirsts, &nSupp );
}

/**Function*************************************************************

  Synopsis    []

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
void If_DsdManComputeTruth_rec( If_DsdMan_t * p, int iDsd, word * pRes, unsigned char * pPermLits, int * pnSupp )
{
    var i int32;
    var iFanin int32;
    var fCompl int32 = Abc_LitIsCompl(iDsd);
    var pObj *If_DsdObj_t  = If_DsdVecObj( &p.vObjs, Abc_Lit2Var(iDsd) );
    if If_DsdObjType(pObj) == IF_DSD_VAR {
        var iPermLit int32;
        if pPermLits {
            iPermLit = int32(pPermLits[*pnSupp]);
        } else {
            iPermLit = Abc_Var2Lit(*pnSupp, 0);
        }
        (*pnSupp)++;
        assert((*pnSupp) <= p.nVars, "(*pnSupp) <= p.nVars");
        Abc_TtCopy( pRes, p.pTtElems[Abc_Lit2Var(iPermLit)], p.nWords, fCompl ^ Abc_LitIsCompl(iPermLit) );
        return;
    }
    if If_DsdObjType(pObj) == IF_DSD_AND || If_DsdObjType(pObj) == IF_DSD_XOR {
        var pTtTemp [DAU_MAX_WORD]word;
        if If_DsdObjType(pObj) == IF_DSD_AND {
            Abc_TtConst1( pRes, p.nWords );
        } else {
            Abc_TtConst0( pRes, p.nWords ); }
        // inlining of macro If_DsdObjForEachFaninLit
        for  i = 0; (i < If_DsdObjFaninNum(pObj)) && ((iFanin) = If_DsdObjFaninLit(pObj, i)); i++  {
            If_DsdManComputeTruth_rec( p, iFanin, pTtTemp, pPermLits, pnSupp );
            if If_DsdObjType(pObj) == IF_DSD_AND {
                Abc_TtAnd( pRes, pRes, pTtTemp, p.nWords, 0 );
            } else {
                Abc_TtXor( pRes, pRes, pTtTemp, p.nWords, 0 ); }
        }
        if fCompl != 0 { Abc_TtNot( pRes, p.nWords ); }
        return;
    }
    if If_DsdObjType(pObj) == IF_DSD_MUX { // mux
    
        var pTtTemp [3][DAU_MAX_WORD]word;
        // inlining of macro If_DsdObjForEachFaninLit
        for  i = 0; (i < If_DsdObjFaninNum(pObj)) && ((iFanin) = If_DsdObjFaninLit(pObj, i)); i++  {
            If_DsdManComputeTruth_rec( p, iFanin, pTtTemp[i], pPermLits, pnSupp );
        }
        assert(i == 3, "i == 3");
        Abc_TtMux( pRes, pTtTemp[0], pTtTemp[1], pTtTemp[2], p.nWords );
        if fCompl != 0 { Abc_TtNot( pRes, p.nWords ); }
        return;
    }
    if If_DsdObjType(pObj) == IF_DSD_PRIME { // function
    
        var pFanins [DAU_MAX_VAR][DAU_MAX_WORD]word;
        // inlining of macro If_DsdObjForEachFaninLit
        for  i = 0; (i < If_DsdObjFaninNum(pObj)) && ((iFanin) = If_DsdObjFaninLit(pObj, i)); i++  {
            If_DsdManComputeTruth_rec( p, iFanin, pFanins[i], pPermLits, pnSupp );
        }
        Dau_DsdTruthCompose_rec( If_DsdObjTruth(p, pObj), pFanins, pRes, pObj.nFans, p.nWords );
        if fCompl != 0 { Abc_TtNot( pRes, p.nWords ); }
        return;
    }
    assert(false, "0");

}
void If_DsdManComputeTruthPtr( If_DsdMan_t * p, int iDsd, unsigned char * pPermLits, word * pRes )
{
    var nSupp int32;
    var pObj *If_DsdObj_t  = If_DsdVecObj( &p.vObjs, Abc_Lit2Var(iDsd) );
    if iDsd == 0 {
        Abc_TtConst0( pRes, p.nWords );
    } else if iDsd == 1 {
        Abc_TtConst1( pRes, p.nWords );
    } else if pObj.Type == IF_DSD_VAR {
        var iPermLit int32;
        if pPermLits {
            iPermLit = int32(pPermLits[nSupp]);
        } else {
            iPermLit = Abc_Var2Lit(nSupp, 0);
        }
        nSupp++;
        Abc_TtCopy( pRes, p.pTtElems[Abc_Lit2Var(iPermLit)], p.nWords, Abc_LitIsCompl(iDsd) ^ Abc_LitIsCompl(iPermLit) );
    
    } else {
        If_DsdManComputeTruth_rec( p, iDsd, pRes, pPermLits, &nSupp ); }
    assert(nSupp == If_DsdVecLitSuppSize(&p.vObjs, iDsd), "nSupp == If_DsdVecLitSuppSize(&p.vObjs, iDsd)");
}
word * If_DsdManComputeTruth( If_DsdMan_t * p, int iDsd, unsigned char * pPermLits )
{
    var pRes *word  = p.pTtElems[DAU_MAX_VAR];
    If_DsdManComputeTruthPtr( p, iDsd, pPermLits, pRes );
    return pRes;
}

/**Function*************************************************************

  Synopsis    [Procedures to propagate the invertor.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func If_DsdManCheckInv_rec(p *If_DsdMan_t, iLit int32) int32 {
    var pObj *If_DsdObj_t;
    var i, iFanin int32;
    pObj = If_DsdVecObj( &p.vObjs, Abc_Lit2Var(iLit) );
    if If_DsdObjType(pObj) == IF_DSD_VAR {
        return 1; }
    if If_DsdObjType(pObj) == IF_DSD_AND || If_DsdObjType(pObj) == IF_DSD_PRIME {
        return 0; }
    if If_DsdObjType(pObj) == IF_DSD_XOR {
        // inlining of macro If_DsdObjForEachFaninLit
        for  i = 0; (i < If_DsdObjFaninNum(pObj)) && ((iFanin) = If_DsdObjFaninLit(pObj, i)); i++  {
            if If_DsdManCheckInv_rec(p, iFanin) != 0 {
                return 1; }
        }
        return 0;
    }
    if If_DsdObjType(pObj) == IF_DSD_MUX {
        return If_DsdManCheckInv_rec(p, pObj.pFans[1]) && If_DsdManCheckInv_rec(p, pObj.pFans[2]); }
    assert(false, "0");
    return 0;
}
int If_DsdManPushInv_rec( If_DsdMan_t * p, int iLit, unsigned char * pPerm )
{
    var pObj *If_DsdObj_t;
    var i, iFanin int32;
    pObj = If_DsdVecObj( &p.vObjs, Abc_Lit2Var(iLit) );
    if If_DsdObjType(pObj) == IF_DSD_VAR {
        pPerm[0] = (unsigned char)Abc_LitNot(int32(pPerm[0]));
    } else if If_DsdObjType(pObj) == IF_DSD_XOR {
        // inlining of macro If_DsdObjForEachFaninLit
        for  i = 0; (i < If_DsdObjFaninNum(pObj)) && ((iFanin) = If_DsdObjFaninLit(pObj, i)); i++  {
            if If_DsdManCheckInv_rec(p, iFanin) != 0 {
                If_DsdManPushInv_rec( p, iFanin, pPerm );
                break;
            }
            pPerm += If_DsdVecLitSuppSize(&p.vObjs, iFanin);
        }
    
    } else if If_DsdObjType(pObj) == IF_DSD_MUX {
        assert((If_DsdManCheckInv_rec(p, pObj.pFans[1]) != 0) && (If_DsdManCheckInv_rec(p, pObj.pFans[2]) != 0), "If_DsdManCheckInv_rec(p, pObj.pFans[1]) && If_DsdManCheckInv_rec(p, pObj.pFans[2])");
        pPerm += If_DsdVecLitSuppSize(&p.vObjs, pObj.pFans[0]);
        If_DsdManPushInv_rec(p, pObj.pFans[1], pPerm);
        pPerm += If_DsdVecLitSuppSize(&p.vObjs, pObj.pFans[1]);
        If_DsdManPushInv_rec(p, pObj.pFans[2], pPerm);
    
    } else { assert(false, "0"); }
    return 1;
}
int If_DsdManPushInv( If_DsdMan_t * p, int iLit, unsigned char * pPerm )
{
    if Abc_LitIsCompl(iLit) && (If_DsdManCheckInv_rec(p, iLit) != 0) {
        return If_DsdManPushInv_rec( p, iLit, pPerm ); }
    return 0;
}

/**Function*************************************************************

  Synopsis    [Performs DSD operation.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func If_DsdManComputeFirstArray(p *If_DsdMan_t, pLits *int32, nLits int32, pFirsts *int32) int32 {
    var i, nSSize int32;
    for  i = 0; i < nLits; i++  {
        pFirsts[i] = nSSize;
        nSSize += If_DsdVecLitSuppSize(&p.vObjs, pLits[i]);
    }
    return nSSize;
}
func If_DsdManComputeFirst(p *If_DsdMan_t, pObj *If_DsdObj_t, pFirsts *int32) int32 {
    return If_DsdManComputeFirstArray( p, (int *)pObj.pFans, pObj.nFans, pFirsts );
}
int If_DsdManOperation( If_DsdMan_t * p, int Type, int * pLits, int nLits, unsigned char * pPerm, word * pTruth )
{
    var pObj, pFanin *If_DsdObj_t;
    unsigned char pPermNew[DAU_MAX_VAR], * pPermStart = pPerm;
    var nChildren int32;
    var pChildren [DAU_MAX_VAR]int32;
    var pBegEnd [DAU_MAX_VAR]int32;
    var i, k, j, Id, iFanin, fCompl, nSSize int32;
    if Type == IF_DSD_AND || Type == IF_DSD_XOR {
        for  k = 0; k < nLits; k++  {
            if Type == IF_DSD_XOR && Abc_LitIsCompl(pLits[k]) {
                pLits[k] = Abc_LitNot(pLits[k]);
                fCompl ^= 1;
            }
            pObj = If_DsdVecObj( &p.vObjs, Abc_Lit2Var(pLits[k]) );
            if Type == If_DsdObjType(pObj) && (Type == IF_DSD_XOR || !Abc_LitIsCompl(pLits[k])) {
                // inlining of macro If_DsdObjForEachFaninLit
                for  i = 0; (i < If_DsdObjFaninNum(pObj)) && ((iFanin) = If_DsdObjFaninLit(pObj, i)); i++  {
                    assert(Type == IF_DSD_AND || !Abc_LitIsCompl(iFanin), "Type == IF_DSD_AND || !Abc_LitIsCompl(iFanin)");
                    pChildren[nChildren] = iFanin;
                    pBegEnd[nChildren++] = (nSSize << 16) | (nSSize + If_DsdVecLitSuppSize(&p.vObjs, iFanin));
                    nSSize += If_DsdVecLitSuppSize(&p.vObjs, iFanin);
                }
            
            } else {
                pChildren[nChildren] = Abc_LitNotCond( pLits[k], If_DsdManPushInv(p, pLits[k], pPermStart) );
                pBegEnd[nChildren++] = (nSSize << 16) | (nSSize + If_DsdObjSuppSize(pObj));
                nSSize += If_DsdObjSuppSize(pObj);
            }
            pPermStart += If_DsdObjSuppSize(pObj);
        }
        If_DsdObjSort( p, &p.vObjs, pChildren, nChildren, pBegEnd );
        // create permutation
        for  j = i = 0; i < nChildren; i++  {
            for  k = (pBegEnd[i] >> 16); k < (pBegEnd[i] & 0xFF); k++  {
                pPermNew[j++] = pPerm[k];
            }
        }
        assert(j == nSSize, "j == nSSize");
        for  j = 0; j < nSSize; j++  {
            pPerm[j] = pPermNew[j];
        }
    
    } else if Type == IF_DSD_MUX {
        var RetValue int32;
        assert(nLits == 3, "nLits == 3");
        for  k = 0; k < nLits; k++  {
            pFanin = If_DsdVecObj( &p.vObjs, Abc_Lit2Var(pLits[k]) );
            pLits[k] = Abc_LitNotCond( pLits[k], If_DsdManPushInv(p, pLits[k], pPermStart) );
            pPermStart += pFanin.nSupp;
        }
        RetValue = If_DsdObjCompare( p, &p.vObjs, pLits[1], pLits[2] );
        if RetValue == 1 || (RetValue == 0 && Abc_LitIsCompl(pLits[0])) {
            var nSupp0 int32  = If_DsdVecLitSuppSize( &p.vObjs, pLits[0] );
            var nSupp1 int32  = If_DsdVecLitSuppSize( &p.vObjs, pLits[1] );
            var nSupp2 int32  = If_DsdVecLitSuppSize( &p.vObjs, pLits[2] );
            pLits[0] = Abc_LitNot(pLits[0]);
            ABC_SWAP( int, pLits[1], pLits[2] );
            for  j = k = 0; k < nSupp0; k++  {
                pPermNew[j++] = pPerm[k];
            }
            for  k = 0; k < nSupp2; k++  {
                pPermNew[j++] = pPerm[nSupp0 + nSupp1 + k];
            }
            for  k = 0; k < nSupp1; k++  {
                pPermNew[j++] = pPerm[nSupp0 + k];
            }
            for  j = 0; j < nSupp0 + nSupp1 + nSupp2; j++  {
                pPerm[j] = pPermNew[j];
            }
        }
        if Abc_LitIsCompl(pLits[1]) {
            pLits[1] = Abc_LitNot(pLits[1]);
            pLits[2] = Abc_LitNot(pLits[2]);
            fCompl ^= 1;
        }
        pPermStart = pPerm;
        for  k = 0; k < nLits; k++  {
            pFanin = If_DsdVecObj( &p.vObjs, Abc_Lit2Var(pLits[k]) );
            pChildren[nChildren++] = Abc_LitNotCond( pLits[k], If_DsdManPushInv(p, pLits[k], pPermStart) );
            pPermStart += pFanin.nSupp;
        }
    
    } else if Type == IF_DSD_PRIME {
        var pCanonPerm [DAU_MAX_VAR]byte;
        var i int32;
        var uCanonPhase int32;
        var pFirsts [DAU_MAX_VAR]int32;
        uCanonPhase = Abc_TtCanonicize( pTruth, nLits, pCanonPerm );
        fCompl = ((uCanonPhase >> nLits) & 1);
        nSSize = If_DsdManComputeFirstArray( p, pLits, nLits, pFirsts );
        for  j = i = 0; i < nLits; i++  {
            var iLitNew int32  = Abc_LitNotCond( pLits[int32(pCanonPerm[i])], ((uCanonPhase>>i)&1) );
            pFanin = If_DsdVecObj( &p.vObjs, Abc_Lit2Var(iLitNew) );
            pPermStart = pPerm + pFirsts[int32(pCanonPerm[i])];
            pChildren[nChildren++] = Abc_LitNotCond( iLitNew, If_DsdManPushInv(p, iLitNew, pPermStart) );
            for  k = 0; k < int32(pFanin.nSupp); k++  {            
                pPermNew[j++] = pPermStart[k];
            }
        }
        assert(j == nSSize, "j == nSSize");
        for  j = 0; j < nSSize; j++  {
            pPerm[j] = pPermNew[j];
        }
        Abc_TtStretch6( pTruth, nLits, p.nVars );
    
    } else { assert(false, "0"); }
    // create new graph
    Id = If_DsdObjFindOrAdd( p, Type, pChildren, nChildren, pTruth );
    return Abc_Var2Lit( Id, fCompl );
}

/**Function*************************************************************

  Synopsis    [Creating DSD network from SOP.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func If_DsdMergeMatches(pDsd string, pMatches *int32) {
    var pNested [DAU_MAX_VAR]int32;
    var i, nNested int32;
    for  i = 0; pDsd[i]; i++  {
        pMatches[i] = 0;
        if pDsd[i] == '(' || pDsd[i] == '[' || pDsd[i] == '<' || pDsd[i] == '{' {
            pNested[nNested++] = i;
        } else if pDsd[i] == ')' || pDsd[i] == ']' || pDsd[i] == '>' || pDsd[i] == '}' {
            pMatches[pNested[--nNested]] = i; }
        assert(nNested < DAU_MAX_VAR, "nNested < DAU_MAX_VAR");
    }
    assert(nNested == 0, "nNested == 0");
}
int If_DsdManAddDsd_rec( char * pStr, char ** p, int * pMatches, If_DsdMan_t * pMan, word * pTruth, unsigned char * pPerm, int * pnSupp )
{
    unsigned char * pPermStart = pPerm + *pnSupp;
    var iRes int32 = -1;
    var fCompl int32;
    if **p == '!' {
        fCompl = 1;
        (*p)++;
    }
    if **p >= 'a' && **p <= 'z' { // var
    
        pPerm[(*pnSupp)++] = Abc_Var2Lit( **p - 'a', fCompl );
        return 2;
    }
    if **p == '(' || **p == '[' || **p == '<' || **p == '{' { // and/or/xor
    
        var Type int32;
        var nLits int32;
        var pLits [DAU_MAX_VAR]int32;
        var q string  = pStr + pMatches[ *p - pStr ];
        if **p == '(' {
            Type = DAU_DSD_AND;
        } else if **p == '[' {
            Type = DAU_DSD_XOR;
        } else if **p == '<' {
            Type = DAU_DSD_MUX;
        } else if **p == '{' {
            Type = DAU_DSD_PRIME;
        } else { assert(false, "0"); }
        assert(*q == **p + 1 + (**p != '('), "*q == **p + 1 + (**p != '(')");
        for  (*p)++; *p < q; (*p)++  {
            pLits[nLits++] = If_DsdManAddDsd_rec( pStr, p, pMatches, pMan, pTruth, pPerm, pnSupp );
        }
        assert(*p == q, "*p == q");
        iRes = If_DsdManOperation( pMan, Type, pLits, nLits, pPermStart, pTruth );
        return Abc_LitNotCond( iRes, fCompl );
    }
    if (**p >= 'A' && **p <= 'F') || (**p >= '0' && **p <= '9') {
        var pFunc [DAU_MAX_WORD]word;
        var nLits int32;
        var pLits [DAU_MAX_VAR]int32;
        var q string;
        var i int32;
        var nVarsF int32 = Abc_TtReadHex( pFunc, *p );
        *p += Abc_TtHexDigitNum( nVarsF );
        q = pStr + pMatches[ *p - pStr ];
        assert(**p == '{' && *q == '}', "**p == '{' && *q == '}'");
        for  i = 0, (*p)++; *p < q; (*p)++, i++  {
            pLits[nLits++] = If_DsdManAddDsd_rec( pStr, p, pMatches, pMan, pTruth, pPerm, pnSupp );
        }
        assert(i == nVarsF, "i == nVarsF");
        assert(*p == q, "*p == q");
        iRes = If_DsdManOperation( pMan, DAU_DSD_PRIME, pLits, nLits, pPermStart, pFunc );
        return Abc_LitNotCond( iRes, fCompl );
    }
    assert(false, "0");
    return -1;
}
int If_DsdManAddDsd( If_DsdMan_t * p, char * pDsd, word * pTruth, unsigned char * pPerm, int * pnSupp )
{
    var iRes int32 = -1;
    var fCompl int32;
    if *pDsd == '!' {
         pDsd++, fCompl = 1; }
    if Dau_DsdIsConst0(pDsd) {
        iRes = 0;
    } else if Dau_DsdIsConst1(pDsd) {
        iRes = 1;
    } else if Dau_DsdIsVar(pDsd) {
        pPerm[(*pnSupp)++] = Dau_DsdReadVar(pDsd);
        iRes = 2;
    
    } else {
        var pMatches [DAU_MAX_STR]int32;
        If_DsdMergeMatches( pDsd, pMatches );
        iRes = If_DsdManAddDsd_rec( pDsd, &pDsd, pMatches, p, pTruth, pPerm, pnSupp );
    }
    return Abc_LitNotCond( iRes, fCompl );
}

/**Function*************************************************************

  Synopsis    [Returns 1 if XY-decomposability holds to this LUT size.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
// create signature of the support of the node
func If_DsdSign_rec(p *If_DsdMan_t, pObj *If_DsdObj_t, pnSupp *int32) uint32 {
    var uSign uint32; var i int32;
    var pFanin *If_DsdObj_t;
    if If_DsdObjType(pObj) == IF_DSD_VAR {
        return (1 << (2*(*pnSupp)++)); }
    // inlining of macro If_DsdObjForEachFanin
    for  i = 0; (i < If_DsdObjFaninNum(pObj)) && ((pFanin) = If_DsdObjFanin(&p.vObjs, pObj, i)); i++  {
        uSign |= If_DsdSign_rec( p, pFanin, pnSupp );
    }
    return uSign;
}
func If_DsdSign(p *If_DsdMan_t, pObj *If_DsdObj_t, iFan int32, iFirst int32, fShared int32) uint32 {
    var pFanin *If_DsdObj_t  = If_DsdObjFanin( &p.vObjs, pObj, iFan );
    var uSign uint32  = If_DsdSign_rec( p, pFanin, &iFirst );
    if fShared != 0 {
        return (uSign << 1) | uSign;
    } else {
        return uSign;
    }
}
// collect supports of the node
func If_DsdManGetSuppSizes(p *If_DsdMan_t, pObj *If_DsdObj_t, pSSizes *int32) {
    var pFanin *If_DsdObj_t; var i int32;
    // inlining of macro If_DsdObjForEachFanin
    for  i = 0; (i < If_DsdObjFaninNum(pObj)) && ((pFanin) = If_DsdObjFanin(&p.vObjs, pObj, i)); i++  {
        pSSizes[i] = If_DsdObjSuppSize(pFanin);
    }    
}
// checks if there is a way to package some fanins 
func If_DsdManCheckAndXor(p *If_DsdMan_t, iFirst int32, uMaskNot uint32, pObj *If_DsdObj_t, nSuppAll int32, LutSize int32, fDerive int32, fVerbose int32) uint32 {
    var i [6]int32;
    var LimitOut int32;
    var SizeIn int32;
    var SizeOut int32;
    var pSSizes [DAU_MAX_VAR]int32;
    var nFans int32 = If_DsdObjFaninNum(pObj);
    var pFirsts [DAU_MAX_VAR]int32;
    var uRes uint32;
    assert(pObj.nFans > 2, "pObj.nFans > 2");
    assert(If_DsdObjSuppSize(pObj) > LutSize, "If_DsdObjSuppSize(pObj) > LutSize");
    If_DsdManGetSuppSizes( p, pObj, pSSizes );
    LimitOut = LutSize - (nSuppAll - pObj.nSupp + 1);
    assert(LimitOut < LutSize, "LimitOut < LutSize");
    for  i[0] = 0;      i[0] < nFans; i[0]++  {
    for  i[1] = i[0]+1; i[1] < nFans; i[1]++  {
        SizeIn = pSSizes[i[0]] + pSSizes[i[1]];
        SizeOut = pObj.nSupp - SizeIn;
        if SizeIn > LutSize || SizeOut > LimitOut {
            continue; }
        if !(fDerive != 0) {
            return ~0; }
        If_DsdManComputeFirst( p, pObj, pFirsts );
        uRes = If_DsdSign(p, pObj, i[0], iFirst + pFirsts[i[0]], 0) | 
               If_DsdSign(p, pObj, i[1], iFirst + pFirsts[i[1]], 0);
        if uRes & uMaskNot {
            continue; }
        return uRes;
    }
    }
    if pObj.nFans == 3 {
        return 0; }
    for  i[0] = 0;      i[0] < nFans; i[0]++  {
    for  i[1] = i[0]+1; i[1] < nFans; i[1]++  {
    for  i[2] = i[1]+1; i[2] < nFans; i[2]++  {
        SizeIn = pSSizes[i[0]] + pSSizes[i[1]] + pSSizes[i[2]];
        SizeOut = pObj.nSupp - SizeIn;
        if SizeIn > LutSize || SizeOut > LimitOut {
            continue; }
        if !(fDerive != 0) {
            return ~0; }
        If_DsdManComputeFirst( p, pObj, pFirsts );
        uRes = If_DsdSign(p, pObj, i[0], iFirst + pFirsts[i[0]], 0) | 
               If_DsdSign(p, pObj, i[1], iFirst + pFirsts[i[1]], 0) | 
               If_DsdSign(p, pObj, i[2], iFirst + pFirsts[i[2]], 0);
        if uRes & uMaskNot {
            continue; }
        return uRes;
    }
    }
    }
    if pObj.nFans == 4 {
        return 0; }
    for  i[0] = 0;      i[0] < nFans; i[0]++  {
    for  i[1] = i[0]+1; i[1] < nFans; i[1]++  {
    for  i[2] = i[1]+1; i[2] < nFans; i[2]++  {
    for  i[3] = i[2]+1; i[3] < nFans; i[3]++  {
        SizeIn = pSSizes[i[0]] + pSSizes[i[1]] + pSSizes[i[2]] + pSSizes[i[3]];
        SizeOut = pObj.nSupp - SizeIn;
        if SizeIn > LutSize || SizeOut > LimitOut {
            continue; }
        if !(fDerive != 0) {
            return ~0; }
        If_DsdManComputeFirst( p, pObj, pFirsts );
        uRes = If_DsdSign(p, pObj, i[0], iFirst + pFirsts[i[0]], 0) | 
               If_DsdSign(p, pObj, i[1], iFirst + pFirsts[i[1]], 0) | 
               If_DsdSign(p, pObj, i[2], iFirst + pFirsts[i[2]], 0) | 
               If_DsdSign(p, pObj, i[3], iFirst + pFirsts[i[3]], 0);
        if uRes & uMaskNot {
            continue; }
        return uRes;
    }
    }
    }
    }
    return 0;
}
// checks if there is a way to package some fanins 
func If_DsdManCheckMux(p *If_DsdMan_t, iFirst int32, uMaskNot uint32, pObj *If_DsdObj_t, nSuppAll int32, LutSize int32, fDerive int32, fVerbose int32) uint32 {
    var LimitOut int32;
    var SizeIn int32;
    var SizeOut int32;
    var pSSizes [DAU_MAX_VAR]int32;
    var pFirsts [DAU_MAX_VAR]int32;
    var uRes uint32;
    assert(If_DsdObjFaninNum(pObj) == 3, "If_DsdObjFaninNum(pObj) == 3");
    assert(If_DsdObjSuppSize(pObj) > LutSize, "If_DsdObjSuppSize(pObj) > LutSize");
    If_DsdManGetSuppSizes( p, pObj, pSSizes );
    LimitOut = LutSize - (nSuppAll - If_DsdObjSuppSize(pObj) + 1);
    assert(LimitOut < LutSize, "LimitOut < LutSize");
    // first input
    SizeIn = pSSizes[0] + pSSizes[1];
    SizeOut = pSSizes[0] + pSSizes[2] + 1;
    if SizeIn <= LutSize && SizeOut <= LimitOut {
        if !(fDerive != 0) {
            return ~0; }
        If_DsdManComputeFirst( p, pObj, pFirsts );
        uRes = If_DsdSign(p, pObj, 0, iFirst + pFirsts[0], 1) | If_DsdSign(p, pObj, 1, iFirst + pFirsts[1], 0);
        if (uRes & uMaskNot) == 0 {
            return uRes; }
    }
    // second input
    SizeIn = pSSizes[0] + pSSizes[2];
    SizeOut = pSSizes[0] + pSSizes[1] + 1;
    if SizeIn <= LutSize && SizeOut <= LimitOut {
        if !(fDerive != 0) {
            return ~0; }
        If_DsdManComputeFirst( p, pObj, pFirsts );
        uRes = If_DsdSign(p, pObj, 0, iFirst + pFirsts[0], 1) | If_DsdSign(p, pObj, 2, iFirst + pFirsts[2], 0);
        if (uRes & uMaskNot) == 0 {
            return uRes; }
    }
    return 0;
}
// checks if there is a way to package some fanins 
func If_DsdManCheckPrime(p *If_DsdMan_t, iFirst int32, uMaskNot uint32, pObj *If_DsdObj_t, nSuppAll int32, LutSize int32, fDerive int32, fVerbose int32) uint32 {
    var i int32;
    var v int32;
    var set int32;
    var LimitOut int32;
    var SizeIn int32;
    var SizeOut int32;
    var pSSizes [DAU_MAX_VAR]int32;
    var pFirsts [DAU_MAX_VAR]int32;
    var truthId int32  = If_DsdObjTruthId(p, pObj);
    var nFans int32  = If_DsdObjFaninNum(pObj);
    var vSets *Vec_Int_t  = (Vec_Int_t *)Vec_PtrEntry(p.vTtDecs[pObj.nFans], truthId);
if fVerbose != 0 {
fmt.Printf("\n"); }
if fVerbose != 0 {
Dau_DecPrintSets( vSets, nFans ); }
    assert(If_DsdObjFaninNum(pObj) > 2, "If_DsdObjFaninNum(pObj) > 2");
    assert(If_DsdObjSuppSize(pObj) > LutSize, "If_DsdObjSuppSize(pObj) > LutSize");
    If_DsdManGetSuppSizes( p, pObj, pSSizes );
    LimitOut = LutSize - (nSuppAll - If_DsdObjSuppSize(pObj) + 1);
    assert(LimitOut < LutSize, "LimitOut < LutSize");
    Vec_IntForEachEntry( vSets, set, i )
    {
        SizeIn = SizeOut = 0;
        for  v = 0; v < nFans; v++  {
            var Value int32  = ((set >> (v << 1)) & 3);
            if Value == 0 {
                SizeOut += pSSizes[v];
            } else if Value == 1 {
                SizeIn += pSSizes[v];
            } else if Value == 3 {
                SizeIn += pSSizes[v];
                SizeOut += pSSizes[v];
            
            } else { assert(false, "0"); }
            if SizeIn > LutSize || SizeOut > LimitOut {
                break; }
        }
        if v == nFans {
            var uRes uint32;
            if !(fDerive != 0) {
                return ~0; }
            If_DsdManComputeFirst( p, pObj, pFirsts );
            for  v = 0; v < nFans; v++  {
                var Value int32  = ((set >> (v << 1)) & 3);
                if Value == 0 {
                } else if Value == 1 {
                    uRes |= If_DsdSign(p, pObj, v, iFirst + pFirsts[v], 0);
                } else if Value == 3 {
                    uRes |= If_DsdSign(p, pObj, v, iFirst + pFirsts[v], 1);
                } else { assert(false, "0"); }
            }
            if uRes & uMaskNot {
                continue; }
            return uRes;
        }
    }
    return 0;
}
func If_DsdManCheckXY_int(p *If_DsdMan_t, iDsd int32, LutSize int32, fDerive int32, uMaskNot uint32, fVerbose int32) uint32 {
    var pObj, pTemp *If_DsdObj_t; 
    var i, Mask, iFirst int32;
    var uRes uint32;
    pObj = If_DsdVecObj( &p.vObjs, Abc_Lit2Var(iDsd) );
    if fVerbose != 0 {
    If_DsdManPrintOne( os.Stdout, p, Abc_Lit2Var(iDsd), nil, 0 ); }
    if If_DsdObjSuppSize(pObj) <= LutSize {
        if fVerbose != 0 {
        fmt.Printf("    Trivial\n"); }
        return ~0;
    }
    If_DsdManCollect( p, pObj.Id, p.vTemp1, p.vTemp2 );
    // inlining of macro If_DsdVecForEachObjVec
    for  i = 0; (i < Vec_IntSize(p.vTemp1)) && ((pTemp) = If_DsdVecObj(&p.vObjs, Vec_IntEntry(p.vTemp1,i))); i++  {
        if If_DsdObjSuppSize(pTemp) <= LutSize && If_DsdObjSuppSize(pObj) - If_DsdObjSuppSize(pTemp) <= LutSize - 1 {
            if fVerbose != 0 {
            fmt.Printf("    Dec using node "); }
            if fVerbose != 0 {
            If_DsdManPrintOne( os.Stdout, p, pTemp.Id, nil, 1 ); }
            iFirst = Vec_IntEntry(p.vTemp2, i);
            uRes = If_DsdSign_rec(p, pTemp, &iFirst);
            if uRes & uMaskNot {
                continue; }
            return uRes;
        }
    }
    // inlining of macro If_DsdVecForEachObjVec
    for  i = 0; (i < Vec_IntSize(p.vTemp1)) && ((pTemp) = If_DsdVecObj(&p.vObjs, Vec_IntEntry(p.vTemp1,i))); i++  {
        if (If_DsdObjType(pTemp) == IF_DSD_AND || If_DsdObjType(pTemp) == IF_DSD_XOR) && If_DsdObjFaninNum(pTemp) > 2 && If_DsdObjSuppSize(pTemp) > LutSize {
            if Mask = If_DsdManCheckAndXor(p, Vec_IntEntry(p.vTemp2, i), uMaskNot, pTemp, If_DsdObjSuppSize(pObj), LutSize, fDerive, fVerbose) {
                if fVerbose != 0 {
                fmt.Printf("    "); }
                if fVerbose != 0 {
                Abc_TtPrintBinary( (word *)&Mask, 4 ); } 
                if fVerbose != 0 {
                fmt.Printf("    Using multi-input AND/XOR node\n"); }
                return Mask;
            }
        }
    }
    // inlining of macro If_DsdVecForEachObjVec
    for  i = 0; (i < Vec_IntSize(p.vTemp1)) && ((pTemp) = If_DsdVecObj(&p.vObjs, Vec_IntEntry(p.vTemp1,i))); i++  {
        if If_DsdObjType(pTemp) == IF_DSD_MUX && If_DsdObjSuppSize(pTemp) > LutSize {
            if Mask = If_DsdManCheckMux(p, Vec_IntEntry(p.vTemp2, i), uMaskNot, pTemp, If_DsdObjSuppSize(pObj), LutSize, fDerive, fVerbose) {
                if fVerbose != 0 {
                fmt.Printf("    "); }
                if fVerbose != 0 {
                Abc_TtPrintBinary( (word *)&Mask, 4 ); } 
                if fVerbose != 0 {
                fmt.Printf("    Using multi-input MUX node\n"); }
                return Mask;
            }
        }
    }
    // inlining of macro If_DsdVecForEachObjVec
    for  i = 0; (i < Vec_IntSize(p.vTemp1)) && ((pTemp) = If_DsdVecObj(&p.vObjs, Vec_IntEntry(p.vTemp1,i))); i++  {
        if If_DsdObjType(pTemp) == IF_DSD_PRIME && If_DsdObjSuppSize(pTemp) > LutSize {
            if Mask = If_DsdManCheckPrime(p, Vec_IntEntry(p.vTemp2, i), uMaskNot, pTemp, If_DsdObjSuppSize(pObj), LutSize, fDerive, fVerbose) {
                if fVerbose != 0 {
                fmt.Printf("    "); }
                if fVerbose != 0 {
                Dau_DecPrintSet( Mask, If_DsdObjFaninNum(pTemp), 0 ); }
                if fVerbose != 0 {
                fmt.Printf("    Using prime node\n"); }
                return Mask;
            }
        }
    }
    if fVerbose != 0 {
    fmt.Printf("    UNDEC\n"); }
//    If_DsdManPrintOne( stdout, p, Abc_Lit2Var(iDsd), NULL, 1 );
    return 0;
}
func If_DsdManCheckXY(p *If_DsdMan_t, iDsd int32, LutSize int32, fDerive int32, uMaskNot uint32, fHighEffort int32, fVerbose int32) uint32 {
    var uSet uint32  = If_DsdManCheckXY_int( p, iDsd, LutSize, fDerive, uMaskNot, fVerbose );
    if uSet == 0 && (fHighEffort != 0) {
//        abctime clk = Abc_Clock();
        var nVars int32  = If_DsdVecLitSuppSize( &p.vObjs, iDsd );
        var pRes *word  = If_DsdManComputeTruth( p, iDsd, nil );
        uSet = If_ManSatCheckXYall( p.pSat, LutSize, pRes, nVars, p.vTemp1 );
        if uSet != 0 {
//            If_DsdManPrintOne( stdout, p, Abc_Lit2Var(iDsd), NULL, 1 );
//            Dau_DecPrintSet( uSet, nVars, 1 );
        }
//        p->timeCheck2 += Abc_Clock() - clk;
    }
    return uSet;
}

/**Function*************************************************************

  Synopsis    [Checks existence of decomposition.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func If_DsdManCheckXYZ(p *If_DsdMan_t, iDsd int32, LutSize int32, fDerive int32, fVerbose int32) uint32 {
    return ~0;
}

/**Function*************************************************************

  Synopsis    [Add the function to the DSD manager.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
int If_DsdManCompute( If_DsdMan_t * p, word * pTruth, int nLeaves, unsigned char * pPerm, char * pLutStruct )
{
    var pCopy [DAU_MAX_WORD]word;
    var pRes *word;
    var pDsd [DAU_MAX_STR]byte;
    var iDsd, nSizeNonDec, nSupp int32;
    var nWords int32  = Abc_TtWordNum(nLeaves);
//    abctime clk = 0;
    assert(nLeaves <= DAU_MAX_VAR, "nLeaves <= DAU_MAX_VAR");
    Abc_TtCopy( pCopy, pTruth, nWords, 0 );
//clk = Abc_Clock();
    nSizeNonDec = Dau_DsdDecompose( pCopy, nLeaves, 0, 1, pDsd );
//p->timeDsd += Abc_Clock() - clk;
    if nSizeNonDec > 0 {
        Abc_TtStretch6( pCopy, nSizeNonDec, p.nVars ); }
    memset( pPerm, 0xFF, (size_t)nLeaves );
//clk = Abc_Clock();
    iDsd = If_DsdManAddDsd( p, pDsd, pCopy, pPerm, &nSupp );
//p->timeCanon += Abc_Clock() - clk;
    assert(nSupp == nLeaves, "nSupp == nLeaves");
    // verify the result
//clk = Abc_Clock();
    pRes = If_DsdManComputeTruth( p, iDsd, pPerm );
//p->timeVerify += Abc_Clock() - clk;
    if !Abc_TtEqual(pRes, pTruth, nWords) {
//        If_DsdManPrint( p, NULL );
        fmt.Printf("\n");
        fmt.Printf("Verification failed!\n");
        fmt.Printf("%s\n", pDsd);
        Dau_DsdPrintFromTruth( pTruth, nLeaves );
        Dau_DsdPrintFromTruth( pRes, nLeaves );
        If_DsdManPrintOne( os.Stdout, p, Abc_Lit2Var(iDsd), pPerm, 1 );
        fmt.Printf("\n");
    }
    If_DsdVecObjIncRef( &p.vObjs, Abc_Lit2Var(iDsd) );
    assert(If_DsdVecLitSuppSize(&p.vObjs, iDsd) == nLeaves, "If_DsdVecLitSuppSize(&p.vObjs, iDsd) == nLeaves");
    return iDsd;
}

/**Function*************************************************************

  Synopsis    [Checks existence of decomposition.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func If_DsdManTest() {
    var vSets *Vec_Int_t;
    var t word  = 0x5277;
    t = Abc_Tt6Stretch( t, 4 );
//    word t = 0xD9D900D900D900001010001000100000;
    vSets = Dau_DecFindSets( &t, 6 );
    Vec_IntFree( vSets );
}


/**Function*************************************************************

  Synopsis    [Compute pin delays.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func If_CutDsdBalancePinDelays_rec(p *If_DsdMan_t, Id int32, pTimes *int32, pRes *word, pnSupp *int32, nSuppAll int32, pPermLits string) int32 {
    var pObj *If_DsdObj_t  = If_DsdVecObj( &p.vObjs, Id );
    if If_DsdObjType(pObj) == IF_DSD_VAR {
        var iCutVar int32  = Abc_Lit2Var(pPermLits[(*pnSupp)++]);
        *pRes = If_CutPinDelayInit(iCutVar);
        return pTimes[iCutVar];
    }
    if If_DsdObjType(pObj) == IF_DSD_MUX {
        var pFaninRes [3]word;
        var Res0 word;
        var Res1 word;
        var i int32;
        var iFanin int32;
        var Delays [3]int32;
        // inlining of macro If_DsdObjForEachFaninLit
        for  i = 0; (i < If_DsdObjFaninNum(pObj)) && ((iFanin) = If_DsdObjFaninLit(pObj, i)); i++  {
            Delays[i] = If_CutDsdBalancePinDelays_rec( p, Abc_Lit2Var(iFanin), pTimes, pFaninRes+i, pnSupp, nSuppAll, pPermLits );
        }
        Res0 = If_CutPinDelayMax( pFaninRes[0], pFaninRes[1], nSuppAll, 1 );
        Res1 = If_CutPinDelayMax( pFaninRes[0], pFaninRes[2], nSuppAll, 1 );
        *pRes = If_CutPinDelayMax( Res0, Res1, nSuppAll, 1 );
        return 2 + Abc_MaxInt(Delays[0], Abc_MaxInt(Delays[1], Delays[2]));
    }
    if If_DsdObjType(pObj) == IF_DSD_PRIME {
        var pFaninRes [IF_MAX_FUNC_LUTSIZE]word;
        var i int32;
        var iFanin int32;
        var Delays [IF_MAX_FUNC_LUTSIZE]int32;
        var vCover *Vec_Int_t  = Vec_WecEntry( p.vIsops[pObj.nFans], If_DsdObjTruthId(p, pObj) );
        assert(Vec_IntSize(vCover) > 0, "Vec_IntSize(vCover) > 0");
        // inlining of macro If_DsdObjForEachFaninLit
        for  i = 0; (i < If_DsdObjFaninNum(pObj)) && ((iFanin) = If_DsdObjFaninLit(pObj, i)); i++  {
            Delays[i] = If_CutDsdBalancePinDelays_rec( p, Abc_Lit2Var(iFanin), pTimes, pFaninRes+i, pnSupp, nSuppAll, pPermLits );
        }
        return If_CutSopBalancePinDelaysInt( vCover, Delays, pFaninRes, nSuppAll, pRes );
    }
    assert(If_DsdObjType(pObj) == IF_DSD_AND || If_DsdObjType(pObj) == IF_DSD_XOR, "If_DsdObjType(pObj) == IF_DSD_AND || If_DsdObjType(pObj) == IF_DSD_XOR");
    {
        var pFaninRes [IF_MAX_FUNC_LUTSIZE]word;
        var i, iFanin, Delay, Result int32;
        var fXor int32;//(If_DsdObjType(pObj) == IF_DSD_XOR);
        var nCounter int32;
        var pCounter [IF_MAX_FUNC_LUTSIZE]int32;
        // inlining of macro If_DsdObjForEachFaninLit
        for  i = 0; (i < If_DsdObjFaninNum(pObj)) && ((iFanin) = If_DsdObjFaninLit(pObj, i)); i++  {
            Delay = If_CutDsdBalancePinDelays_rec( p, Abc_Lit2Var(iFanin), pTimes, pFaninRes+i, pnSupp, nSuppAll, pPermLits );
            Result = If_LogCounterPinDelays( pCounter, &nCounter, pFaninRes, Delay, pFaninRes[i], nSuppAll, fXor );
        }
        assert(nCounter > 0, "nCounter > 0");
        if fXor != 0 {
            Result = If_LogCounterDelayXor( pCounter, nCounter ); } // estimation
        *pRes = If_LogPinDelaysMulti( pFaninRes, nCounter, nSuppAll, fXor );
        return Result;
    }
}
func If_CutDsdBalancePinDelays(p *If_Man_t, pCut *If_Cut_t, pPerm string) int32 {
    if pCut.nLeaves == 0 { // const
        return 0; }
    if pCut.nLeaves == 1 { // variable
    
        pPerm[0] = 0;
        return int32(If_ObjCutBest(If_CutLeaf(p, pCut, 0)).Delay);
    
    } else {
        var Result word;
        var i int32;
        var Delay int32;
        var nSupp int32;
        var pTimes [IF_MAX_FUNC_LUTSIZE]int32;
        for  i = 0; i < If_CutLeaveNum(pCut); i++  {
            pTimes[i] = int32(If_ObjCutBest(If_CutLeaf(p, pCut, i)).Delay);
        } 
        Delay = If_CutDsdBalancePinDelays_rec( p.pIfDsdMan, Abc_Lit2Var(If_CutDsdLit(p, pCut)), pTimes, &Result, &nSupp, If_CutLeaveNum(pCut), If_CutDsdPerm(p, pCut) );
        assert(nSupp == If_CutLeaveNum(pCut), "nSupp == If_CutLeaveNum(pCut)");
        If_CutPinDelayTranslate( Result, If_CutLeaveNum(pCut), pPerm );
        return Delay;
    }
}


/**Function*************************************************************

  Synopsis    []

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func If_CutDsdPermLitMax(pPermLits string, nVars int32, iVar int32) int32 {
    var i int32;
    assert(iVar >= 0 && iVar < nVars, "iVar >= 0 && iVar < nVars");
    for  i = 0; i < nVars; i++  {
        if iVar == Abc_Lit2Var(int32(pPermLits[i])) {
            return i; }
    }
    assert(false, "0");
    return -1;
}

/**Function*************************************************************

  Synopsis    [Evaluate delay using DSD balancing.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func If_CutDsdBalanceEval_rec(p *If_DsdMan_t, Id int32, pTimes *int32, pnSupp *int32, vAig *Vec_Int_t, piLit *int32, nSuppAll int32, pArea *int32, pPermLits string) int32 {
    var pObj *If_DsdObj_t  = If_DsdVecObj( &p.vObjs, Id );
    if If_DsdObjType(pObj) == IF_DSD_VAR {
        var iCutVar int32  = Abc_Lit2Var( pPermLits[*pnSupp] );
        if vAig != nil { 
            *piLit = Abc_Var2Lit( iCutVar, Abc_LitIsCompl(pPermLits[*pnSupp]) ); }
        (*pnSupp)++;
        return pTimes[iCutVar];
    }
    if If_DsdObjType(pObj) == IF_DSD_MUX {
        var i int32;
        var iFanin int32;
        var Delays [3]int32;
        var pFaninLits [3]int32;
        // inlining of macro If_DsdObjForEachFaninLit
        for  i = 0; (i < If_DsdObjFaninNum(pObj)) && ((iFanin) = If_DsdObjFaninLit(pObj, i)); i++  {
            Delays[i] = If_CutDsdBalanceEval_rec( p, Abc_Lit2Var(iFanin), pTimes, pnSupp, vAig, pFaninLits+i, nSuppAll, pArea, pPermLits );
            if Delays[i] == -1 {
                return -1; }
            if vAig != nil { 
                pFaninLits[i] = Abc_LitNotCond( pFaninLits[i], Abc_LitIsCompl(iFanin) ); }
        }
        if vAig != nil {
            *piLit = If_LogCreateMux( vAig, pFaninLits[0], pFaninLits[1], pFaninLits[2], nSuppAll );
        } else {
            *pArea += 3; }
        return 2 + Abc_MaxInt(Delays[0], Abc_MaxInt(Delays[1], Delays[2]));
    }
    if If_DsdObjType(pObj) == IF_DSD_PRIME {
        var i int32;
        var iFanin int32;
        var Delays [IF_MAX_FUNC_LUTSIZE]int32;
        var pFaninLits [IF_MAX_FUNC_LUTSIZE]int32;
        var vCover *Vec_Int_t  = Vec_WecEntry( p.vIsops[pObj.nFans], If_DsdObjTruthId(p, pObj) );
        if Vec_IntSize(vCover) == 0 {
            return -1; }
        // inlining of macro If_DsdObjForEachFaninLit
        for  i = 0; (i < If_DsdObjFaninNum(pObj)) && ((iFanin) = If_DsdObjFaninLit(pObj, i)); i++  {
            Delays[i] = If_CutDsdBalanceEval_rec( p, Abc_Lit2Var(iFanin), pTimes, pnSupp, vAig, pFaninLits+i, nSuppAll, pArea, pPermLits );
            if Delays[i] == -1 {
                return -1; }
            if vAig != nil { 
                pFaninLits[i] = Abc_LitNotCond( pFaninLits[i], Abc_LitIsCompl(iFanin) ); }
        }
        return If_CutSopBalanceEvalInt( vCover, Delays, pFaninLits, vAig, piLit, nSuppAll, pArea );
    }
    assert(If_DsdObjType(pObj) == IF_DSD_AND || If_DsdObjType(pObj) == IF_DSD_XOR, "If_DsdObjType(pObj) == IF_DSD_AND || If_DsdObjType(pObj) == IF_DSD_XOR");
    {
        var i, iFanin, Delay, Result int32;
        var fXor int32;//(If_DsdObjType(pObj) == IF_DSD_XOR);
        var fXorFunc int32  = (If_DsdObjType(pObj) == IF_DSD_XOR);
        var nCounter int32;
        var pCounter [IF_MAX_FUNC_LUTSIZE]int32;
        var pFaninLits [IF_MAX_FUNC_LUTSIZE]int32;
        // inlining of macro If_DsdObjForEachFaninLit
        for  i = 0; (i < If_DsdObjFaninNum(pObj)) && ((iFanin) = If_DsdObjFaninLit(pObj, i)); i++  {
            Delay = If_CutDsdBalanceEval_rec( p, Abc_Lit2Var(iFanin), pTimes, pnSupp, vAig, pFaninLits+i, nSuppAll, pArea, pPermLits );
            if Delay == -1 {
                return -1; }
            if vAig != nil { 
                pFaninLits[i] = Abc_LitNotCond( pFaninLits[i], Abc_LitIsCompl(iFanin) ); }
            Result = If_LogCounterAddAig( pCounter, &nCounter, pFaninLits, Delay, vAig ? pFaninLits[i] : -1, vAig, nSuppAll, fXor, fXorFunc );
        }
        assert(nCounter > 0, "nCounter > 0");
        if fXor != 0 {
            Result = If_LogCounterDelayXor( pCounter, nCounter ); } // estimation
        if vAig != nil {
            *piLit = If_LogCreateAndXorMulti( vAig, pFaninLits, nCounter, nSuppAll, fXorFunc );
        } else {
            *pArea += (pObj.nFans - 1) * (1 + 2 * fXor); }
        return Result;
    }
}
func If_CutDsdBalanceEvalInt(p *If_DsdMan_t, iDsd int32, pTimes *int32, vAig *Vec_Int_t, pArea *int32, pPermLits string) int32 {
    var nSupp, iLit int32;
    var nSuppAll int32  = If_DsdVecLitSuppSize( &p.vObjs, iDsd );
    var Res int32  = If_CutDsdBalanceEval_rec( p, Abc_Lit2Var(iDsd), pTimes, &nSupp, vAig, &iLit, nSuppAll, pArea, pPermLits );
    if Res == -1 {
        return -1; }
    assert(nSupp == nSuppAll, "nSupp == nSuppAll");
    assert(vAig == nil || Abc_Lit2Var(iLit) == nSupp + Abc_Lit2Var(Vec_IntSize(vAig)) - 1, "vAig == nil || Abc_Lit2Var(iLit) == nSupp + Abc_Lit2Var(Vec_IntSize(vAig)) - 1");
    if vAig != nil {
        Vec_IntPush( vAig, Abc_LitIsCompl(iLit) ^ Abc_LitIsCompl(iDsd) ); }
    assert(vAig == nil || (Vec_IntSize(vAig) & 1), "vAig == nil || (Vec_IntSize(vAig) & 1)");
    return Res;
}
func If_CutDsdBalanceEval(p *If_Man_t, pCut *If_Cut_t, vAig *Vec_Int_t) int32 {
    var fUseCofs int32;
    pCut.fUser = 1;
    if vAig != nil {
        Vec_IntClear( vAig ); }
    if pCut.nLeaves == 0 { // const
    
        assert(Abc_Lit2Var(If_CutDsdLit(p, pCut)) == 0, "Abc_Lit2Var(If_CutDsdLit(p, pCut)) == 0");
        if vAig != nil {
            Vec_IntPush( vAig, Abc_LitIsCompl(If_CutDsdLit(p, pCut)) ); }
        pCut.Cost = 0;
        return 0;
    }
    if pCut.nLeaves == 1 { // variable
    
        assert(Abc_Lit2Var(If_CutDsdLit(p, pCut)) == 1, "Abc_Lit2Var(If_CutDsdLit(p, pCut)) == 1");
        if vAig != nil {
            Vec_IntPush( vAig, 0 ); }
        if vAig != nil {
            Vec_IntPush( vAig, Abc_LitIsCompl(If_CutDsdLit(p, pCut)) ); }
        pCut.Cost = 0;
        return int32(If_ObjCutBest(If_CutLeaf(p, pCut, 0)).Delay);
    
    } else {
        var fVerbose int32;
        var i int32;
        var pTimes [IF_MAX_FUNC_LUTSIZE]int32;
        var Delay, Area int32;
        var pPermLits string  = If_CutDsdPerm(p, pCut);
        for  i = 0; i < If_CutLeaveNum(pCut); i++  {
            pTimes[i] = int32(If_ObjCutBest(If_CutLeaf(p, pCut, i)).Delay);
        } 
        Delay = If_CutDsdBalanceEvalInt( p.pIfDsdMan, Abc_LitNotCond(If_CutDsdLit(p, pCut), pCut.fCompl), pTimes, vAig, &Area, If_CutDsdPerm(p, pCut) );
        pCut.Cost = Area;
        // try cofactoring 
        if fUseCofs != 0 {
            // count how many times the max one appears
            var iMax int32;
            var nCountMax int32 = 1;
            for  i = 1; i < If_CutLeaveNum(pCut); i++  {
                if pTimes[i] > pTimes[iMax] {
                    iMax = i;
nCountMax = 1;
                } else if pTimes[i] == pTimes[iMax] {
                    nCountMax++; }
            }
            // decide when to try the decomposition
            if nCountMax == 1 && pTimes[iMax] + 2 < Delay && (If_DsdManCheckNonTriv( p.pIfDsdMan, Abc_Lit2Var(If_CutDsdLit(p, pCut)), 
                 If_CutLeaveNum(pCut), If_CutDsdPermLitMax(pPermLits, If_CutLeaveNum(pCut), iMax)) != 0) {
//                fVerbose = 1;
                Delay = pTimes[iMax] + 2;
            }
        }
        // report the result
        if fVerbose != 0 {
/*
            int Max = 0, Two = 0;
            for ( i = 0; i < If_CutLeaveNum(pCut); i++ )
                Max = Abc_MaxInt( Max, pTimes[i] );
            for ( i = 0; i < If_CutLeaveNum(pCut); i++ )
                if ( pTimes[i] != Max )
                    Two = Abc_MaxInt( Two, pTimes[i] );
            if ( Two + 2 < Max && Max + 3 < Delay )
*/
            {
                for  i = 0; i < If_CutLeaveNum(pCut); i++  {
                    fmt.Printf("%3d ", pTimes[Abc_Lit2Var(pPermLits[i])]);
                }
                for  ; i < p.pPars.nLutSize; i++  {
                    fmt.Printf("    ");
                }
                fmt.Printf("-> %3d   ", Delay);
                If_DsdManPrintOne( os.Stdout, p.pIfDsdMan, Abc_Lit2Var(If_CutDsdLit(p, pCut)), nil, 0 );
                fmt.Printf("\n");
            }
        }
        return Delay;
    }
}

/**Function*************************************************************

  Synopsis    []

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func If_DsdManTune(p *If_DsdMan_t, LutSize int32, fFast int32, fAdd int32, fSpec int32, fVerbose int32) {
    var pProgress *ProgressBar  = nil;
    var pSat *sat_solver  = nil;
    var pObj *If_DsdObj_t;
    var vLits *Vec_Int_t;
    var i, Value, nVars int32;
    var pTruth *word;
    if !(fAdd != 0) || !(LutSize != 0) {
        // inlining of macro If_DsdVecForEachObj
        Vec_PtrForEachEntry( If_DsdObj_t *, &p.vObjs, pObj, i )
            pObj.fMark = 0; }
    if LutSize == 0 {
        return; }
    vLits = Vec_IntAlloc( 1000 );
    pSat = (sat_solver *)If_ManSatBuildXY( LutSize );
    pProgress = Extra_ProgressBarStart( os.Stdout, Vec_PtrSize(&p.vObjs) );
    // inlining of macro If_DsdVecForEachObj
    Vec_PtrForEachEntry( If_DsdObj_t *, &p.vObjs, pObj, i )
    {
        Extra_ProgressBarUpdate( pProgress, i, nil );
        nVars = If_DsdObjSuppSize(pObj);
        if nVars <= LutSize {
            continue; }
        if (fAdd != 0) && !pObj.fMark {
            continue; }
        pObj.fMark = 0;
        if If_DsdManCheckXY(p, Abc_Var2Lit(i, 0), LutSize, 0, 0, 0, 0) != 0 {
            continue; }
        if fFast != 0 {
            Value = 0;
        } else {
            pTruth = If_DsdManComputeTruth( p, Abc_Var2Lit(i, 0), nil );
            Value = If_ManSatCheckXYall( pSat, LutSize, pTruth, nVars, vLits );
        }
        if Value != 0 {
            continue; }
        If_DsdVecObjSetMark( &p.vObjs, i );
    }
    Extra_ProgressBarStop( pProgress );
    If_ManSatUnbuild( pSat );
    Vec_IntFree( vLits );
    if fVerbose != 0 {
        If_DsdManPrintDistrib( p ); }
}



/**Function*************************************************************

  Synopsis    []

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Id_DsdManTuneStr1(p *If_DsdMan_t, pStruct string, nConfls int32, fVerbose int32) {
    var fVeryVerbose int32;
    var pProgress *ProgressBar  = nil;
    var pObj *If_DsdObj_t;
    var pTruth, pConfig *word;
    var i, nVars, Value, LutSize int32;
    var clk abctime  = Abc_Clock();
    // parse the structure
    var pNtk *Ifn_Ntk_t  = Ifn_NtkParse( pStruct );
    if pNtk == nil {
        return; }
    if If_DsdManVarNum(p) > Ifn_NtkInputNum(pNtk) {
        fmt.Printf("The support of DSD manager (%d) exceeds the support of the structure (%d).\n", If_DsdManVarNum(p), Ifn_NtkInputNum(pNtk));
        ABC_FREE( pNtk );
        return;
    }
    ABC_FREE( p.pCellStr );
    p.pCellStr = Abc_UtilStrsav( pStruct );
    if If_DsdManVarNum(p) < Ifn_NtkInputNum(pNtk) {
        fmt.Printf("Warning: The support of DSD manager (%d) is less than the support of the structure (%d).\n", If_DsdManVarNum(p), Ifn_NtkInputNum(pNtk)); }
    LutSize = Ifn_NtkLutSizeMax(pNtk);
    p.nTtBits = Ifn_NtkTtBits( pStruct );
    p.nConfigWords = 1 + Abc_Bit6WordNum( p.nTtBits );
    // print
    if fVerbose != 0 {
        fmt.Printf("Considering programmable cell: ");
        Ifn_NtkPrint( pNtk );
        fmt.Printf("Largest LUT size = %d.\n", LutSize);
    }
    if p.nObjsPrev > 0 {
        fmt.Printf("Starting the tuning process from object %d (out of %d).\n", p.nObjsPrev, Vec_PtrSize(&p.vObjs)); }
    // clean the attributes
    // inlining of macro If_DsdVecForEachObj
    Vec_PtrForEachEntry( If_DsdObj_t *, &p.vObjs, pObj, i )
        if ( i >= p.nObjsPrev )
            pObj.fMark = 0;
    if p.vConfigs == nil {
        p.vConfigs = Vec_WrdStart( p.nConfigWords * Vec_PtrSize(&p.vObjs) );
    } else {
        Vec_WrdFillExtra( p.vConfigs, p.nConfigWords * Vec_PtrSize(&p.vObjs), 0 ); }
    pProgress = Extra_ProgressBarStart( os.Stdout, Vec_PtrSize(&p.vObjs) );
    // inlining of macro If_DsdVecForEachObjStart
    Vec_PtrForEachEntryStart( If_DsdObj_t *, &p.vObjs, pObj, i, p.nObjsPrev )
    {
        if (i & 0xFF) == 0 {
            Extra_ProgressBarUpdate( pProgress, i, nil ); }
        nVars = If_DsdObjSuppSize(pObj);
        //if ( nVars <= LutSize )
        //    continue;
        pTruth = If_DsdManComputeTruth( p, Abc_Var2Lit(i, 0), nil );
        if fVeryVerbose != 0 {
            Dau_DsdPrintFromTruth( pTruth, nVars ); }
        if fVerbose != 0 {
            fmt.Printf("%6d : %2d ", i, nVars); }
        pConfig = Vec_WrdEntryP( p.vConfigs, p.nConfigWords * i );
        Value = Ifn_NtkMatch( pNtk, pTruth, nVars, nConfls, fVerbose, fVeryVerbose, pConfig );
        if fVeryVerbose != 0 {
            fmt.Printf("\n"); }
        if Value == 0 {
            If_DsdVecObjSetMark( &p.vObjs, i );
            memset( pConfig, 0, sizeof(word) * p.nConfigWords );
        }
    }
    p.nObjsPrev = 0;
    p.LutSize = 0;
    Extra_ProgressBarStop( pProgress );
    fmt.Printf("Finished matching %d functions. ", Vec_PtrSize(&p.vObjs));
    Abc_PrintTime( 1, "Time", Abc_Clock() - clk );
    if fVeryVerbose != 0 {
        If_DsdManPrintDistrib( p ); }
    ABC_FREE( pNtk );
}


/**Function*************************************************************

  Synopsis    []

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
#ifndef ABC_USE_PTHREADS
void Id_DsdManTuneStr( If_DsdMan_t * p, char * pStruct, int nConfls, int nProcs, int fVerbose )
{
    Id_DsdManTuneStr1( p, pStruct, nConfls, fVerbose );
}
#else // pthreads are used


#define PAR_THR_MAX 100
typedef struct Ifn_ThData_t_
{
    Ifn_Ntk_t * pNtk;    // network
    word        pTruth[DAU_MAX_WORD];
    word        pConfig[10]; // configuration data
    int         nConfigWords;// configuration data word count
    int         nVars;   // support
    int         Id;      // object
    int         nConfls; // conflicts
    int         Result;  // result
    int         Status;  // state
    abctime     clkUsed; // total runtime
} Ifn_ThData_t;
void * Ifn_WorkerThread( void * pArg )
{
    Ifn_ThData_t * pThData = (Ifn_ThData_t *)pArg;
    volatile int * pPlace = &pThData->Status;
    abctime clk;
    while ( 1 )
    {
        while ( *pPlace == 0 );
        assert( pThData->Status == 1 );
        if ( pThData->Id == -1 )
        {
            pthread_exit( NULL );
            assert( 0 );
            return NULL;
        }
        clk = Abc_Clock();
        memset( pThData->pConfig, 0, sizeof(word) * pThData->nConfigWords );
        pThData->Result = Ifn_NtkMatch( pThData->pNtk, pThData->pTruth, pThData->nVars, pThData->nConfls, 0, 0, pThData->pConfig );
        pThData->clkUsed += Abc_Clock() - clk;
        pThData->Status = 0;
//        printf( "Finished object %d\n", pThData->Id );
    }
    assert( 0 );
    return NULL;
}
void Id_DsdManTuneStr( If_DsdMan_t * p, char * pStruct, int nConfls, int nProcs, int fVerbose )
{
    int fVeryVerbose = 0;
    ProgressBar * pProgress = NULL;
    int i, k, nVars, LutSize;
    abctime clk = Abc_Clock();
    Ifn_Ntk_t * pNtk;
    If_DsdObj_t * pObj;
    if ( nProcs == 1 )
    {
        Id_DsdManTuneStr1( p, pStruct, nConfls, fVerbose );
        return;
    }
    if ( nProcs > PAR_THR_MAX )
    {
        printf( "The number of processes (%d) exceeds the precompiled limit (%d).\n", nProcs, PAR_THR_MAX );
        return;
    }
    // parse the structure
    pNtk = Ifn_NtkParse( pStruct );
    if ( pNtk == NULL )
        return;
    if ( If_DsdManVarNum(p) > Ifn_NtkInputNum(pNtk) )
    {
        printf( "The support of DSD manager (%d) exceeds the support of the structure (%d).\n", If_DsdManVarNum(p), Ifn_NtkInputNum(pNtk) );
        ABC_FREE( pNtk );
        return;
    }
    ABC_FREE( p->pCellStr );
    p->pCellStr = Abc_UtilStrsav( pStruct );
    if ( If_DsdManVarNum(p) < Ifn_NtkInputNum(pNtk) )
        printf( "Warning: The support of DSD manager (%d) is less than the support of the structure (%d).\n", If_DsdManVarNum(p), Ifn_NtkInputNum(pNtk) );
    // check the largest LUT
    LutSize = Ifn_NtkLutSizeMax(pNtk);
    p->nTtBits = Ifn_NtkTtBits( pStruct );
    p->nConfigWords = 1 + Abc_Bit6WordNum( p->nTtBits );
    assert( p->nConfigWords <= 10 );
    if ( fVerbose )
    {
        printf( "Considering programmable cell: " );
        Ifn_NtkPrint( pNtk );
        printf( "Largest LUT size = %d.\n", LutSize );
    }
    ABC_FREE( pNtk );
    if ( p->nObjsPrev > 0 )
        printf( "Starting the tuning process from object %d (out of %d).\n", p->nObjsPrev, Vec_PtrSize(&p->vObjs) );
    // clean the attributes
    If_DsdVecForEachObj( &p->vObjs, pObj, i )
        if ( i >= p->nObjsPrev )
            pObj->fMark = 0;
    if ( p->vConfigs == NULL )
        p->vConfigs = Vec_WrdStart( p->nConfigWords * Vec_PtrSize(&p->vObjs) );
    else
        Vec_WrdFillExtra( p->vConfigs, p->nConfigWords * Vec_PtrSize(&p->vObjs), 0 );
    pProgress = Extra_ProgressBarStart( stdout, Vec_PtrSize(&p->vObjs) );

    // perform concurrent solving
    {
        pthread_t WorkerThread[PAR_THR_MAX];
        Ifn_ThData_t ThData[PAR_THR_MAX];
        abctime clk, clkUsed = 0;
        int status, fRunning = 1, iCurrentObj = p->nObjsPrev;
        // start the threads
        for ( i = 0; i < nProcs; i++ )
        {
            ThData[i].pNtk     = Ifn_NtkParse( pStruct );
            ThData[i].nVars    = -1;      // support
            ThData[i].Id       = -1;      // object
            ThData[i].nConfls  = nConfls; // conflicts
            ThData[i].Result   = -1;      // result
            ThData[i].Status   =  0;      // state
            ThData[i].clkUsed  =  0;      // total runtime
            ThData[i].nConfigWords = p->nConfigWords;
            status = pthread_create( WorkerThread + i, NULL, Ifn_WorkerThread, (void *)(ThData + i) );  assert( status == 0 );
        }
        // run the threads
        while ( fRunning || iCurrentObj < Vec_PtrSize(&p->vObjs) )
        {
            for ( i = 0; i < nProcs; i++ )
            {
                if ( ThData[i].Status )
                    continue;
                assert( ThData[i].Status == 0 );
                if ( ThData[i].Id >= 0 )
                {
                    //printf( "Closing obj %d with Thread %d:\n", ThData[i].Id, i );
                    assert( ThData[i].Result == 0 || ThData[i].Result == 1 );
                    if ( ThData[i].Result == 0 )
                        If_DsdVecObjSetMark( &p->vObjs, ThData[i].Id );
                    else
                    {
                        word * pTtWords = Vec_WrdEntryP( p->vConfigs, p->nConfigWords * ThData[i].Id );
                        memcpy( pTtWords, ThData[i].pConfig, sizeof(word) * p->nConfigWords );
                    }
                    ThData[i].Id     = -1;
                    ThData[i].Result = -1;
                }
                for ( k = iCurrentObj; k < Vec_PtrSize(&p->vObjs); k++ )
                {
                    if ( (k & 0xFF) == 0 )
                        Extra_ProgressBarUpdate( pProgress, k, NULL );
                    pObj  = If_DsdVecObj( &p->vObjs, k );
                    nVars = If_DsdObjSuppSize(pObj);
                    //if ( nVars <= LutSize )
                    //    continue;
                    clk = Abc_Clock();
                    If_DsdManComputeTruthPtr( p, Abc_Var2Lit(k, 0), NULL, ThData[i].pTruth );
                    clkUsed += Abc_Clock() - clk;
                    ThData[i].nVars  = nVars;
                    ThData[i].Id     =  k;
                    ThData[i].Result = -1;
                    ThData[i].Status =  1;
                    //printf( "Scheduling %d for Thread %d\n", ThData[i].Id, i );
                    iCurrentObj = k+1;
                    break;
                }
            }
            fRunning = 0;
            for ( i = 0; i < nProcs; i++ )
                if ( ThData[i].Status == 1 || (ThData[i].Status == 0 && ThData[i].Id >= 0) )
                    fRunning = 1;
            //printf( "fRunning %d\n", fRunning );
        }
        // stop the threads
        for ( i = 0; i < nProcs; i++ )
        {
            assert( ThData[i].Status == 0 );
            ThData[i].Id = -1;
            ThData[i].Status = 1;
            ABC_FREE( ThData[i].pNtk );
        }
        if ( fVerbose )
        {
            printf( "Main     : " );
            Abc_PrintTime( 1, "Time", clkUsed );
            for ( i = 0; i < nProcs; i++ )
            {
                printf( "Thread %d : ", i );
                Abc_PrintTime( 1, "Time", ThData[i].clkUsed );
            }
        }
    }

    p->nObjsPrev = 0;
    p->LutSize = 0;
    Extra_ProgressBarStop( pProgress );
    printf( "Finished matching %d functions. ", Vec_PtrSize(&p->vObjs) );
    Abc_PrintTime( 1, "Time", Abc_Clock() - clk );
    if ( fVeryVerbose )
        If_DsdManPrintDistrib( p );
}

#endif // pthreads are used

/**Function*************************************************************

  Synopsis    []

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
#ifdef ABC_USE_CUDD

void Id_DsdManTuneThresh( If_DsdMan_t * p, int fUnate, int fThresh, int fThreshHeuristic, int fVerbose )
{
    extern int Extra_ThreshCheck( word * t, int nVars, int * pW );
    extern int Extra_ThreshHeuristic( word * t, int nVars, int * pW );
    int fVeryVerbose = 0;
    int pW[16];
    ProgressBar * pProgress = NULL;
    If_DsdObj_t * pObj;
    word * pTruth, Perm;
    int i, nVars, Value;
    abctime clk = Abc_Clock();
    assert( fUnate + fThresh + fThreshHeuristic <= 1 );
    if ( p->nObjsPrev > 0 )
        printf( "Starting the tuning process from object %d (out of %d).\n", p->nObjsPrev, Vec_PtrSize(&p->vObjs) );
    // clean the attributes
    If_DsdVecForEachObj( &p->vObjs, pObj, i )
        if ( i >= p->nObjsPrev )
            pObj->fMark = 0;
    if ( p->vConfigs == NULL )
        p->vConfigs = Vec_WrdStart( Vec_PtrSize(&p->vObjs) );
    else
        Vec_WrdFillExtra( p->vConfigs, Vec_PtrSize(&p->vObjs), 0 );
    pProgress = Extra_ProgressBarStart( stdout, Vec_PtrSize(&p->vObjs) );
    If_DsdVecForEachObjStart( &p->vObjs, pObj, i, p->nObjsPrev )
    {
        if ( (i & 0xFF) == 0 )
            Extra_ProgressBarUpdate( pProgress, i, NULL );
        nVars = If_DsdObjSuppSize(pObj);
        if ( nVars > 8 )
            continue;
        pTruth = If_DsdManComputeTruth( p, Abc_Var2Lit(i, 0), NULL );
        if ( fVeryVerbose )
            Dau_DsdPrintFromTruth( pTruth, nVars );
        if ( fVerbose )
            printf( "%6d : %2d ", i, nVars );
        if ( fUnate )
            Value = Abc_TtIsUnate( pTruth, nVars );
        else if ( fThresh )
            Value = Extra_ThreshCheck( pTruth, nVars, pW );
        else if ( fThreshHeuristic )
            Value = Extra_ThreshHeuristic( pTruth, nVars, pW );
        else
            Value = 0;
        Perm = 0;
        if ( fVeryVerbose )
            printf( "\n" );
        if ( Value )
            If_DsdVecObjSetMark( &p->vObjs, i );
        else
            Vec_WrdWriteEntry( p->vConfigs, i, Perm );
    }
    p->nObjsPrev = 0;
    p->LutSize = 0;
    Extra_ProgressBarStop( pProgress );
    printf( "Finished matching %d functions. ", Vec_PtrSize(&p->vObjs) );
    Abc_PrintTime( 1, "Time", Abc_Clock() - clk );
    if ( fVeryVerbose )
        If_DsdManPrintDistrib( p );
}

#endif // ABC_USE_CUDD are used

////////////////////////////////////////////////////////////////////////
///                       END OF FILE                                ///
////////////////////////////////////////////////////////////////////////


//ABC_NAMESPACE_IMPL_END

