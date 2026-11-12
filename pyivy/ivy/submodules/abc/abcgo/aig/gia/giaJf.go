package gia

import (
	"fmt"
	"os"
)

/**CFile****************************************************************

  FileName    [giaJf.c]

  SystemName  [ABC: Logic synthesis and verification system.]

  PackageName [Scalable AIG package.]

  Synopsis    []

  Author      [Alan Mishchenko]
  
  Affiliation [UC Berkeley]

  Date        [Ver. 1.0. Started - June 20, 2005.]

  Revision    [$Id: giaJf.c,v 1.00 2005/06/20 00:00:00 alanmi Exp $]

***********************************************************************/

//#include "gia.h"
//#include "misc/vec/vecSet.h"
//#include "misc/vec/vecMem.h"
//#include "misc/extra/extra.h"
//#include "bool/kit/kit.h"
//#include "misc/util/utilTruth.h"
//#include "opt/dau/dau.h"
//#include "sat/cnf/cnf.h"

//ABC_NAMESPACE_IMPL_START

////////////////////////////////////////////////////////////////////////
///                        DECLARATIONS                              ///
////////////////////////////////////////////////////////////////////////

const JF_LEAF_MAX =   8
#define JF_WORD_MAX  ((JF_LEAF_MAX > 6) ? 1 << (JF_LEAF_MAX-6) : 1)
const JF_CUT_MAX =   16
const JF_EPSILON = 0.005

    
type Jf_Cut_t struct {
    Sign             word;        // signature
    Flow            float32;        // flow
    Time              int32;        // arrival time
    iFunc              int32;       // function 
    Cost              int32;        // cut cost
    pCut [JF_LEAF_MAX+2]int32; // cut
}

    
type Jf_Man_t struct {
    pGia       *Gia_Man_t;        // user's manager
    pPars        *Jf_Par_t;       // users parameter
    pDsd       *Sdm_Man_t;        // extern DSD manager
    vCnfs       *Vec_Int_t;       // costs of elementary CNFs
    vTtMem       *Vec_Mem_t;      // truth table memory and hash table
    vCuts        Vec_Int_t;       // cuts for each node
    vArr        Vec_Int_t;        // arrival time
    vDep        Vec_Int_t;        // departure time
    vFlow        Vec_Flt_t;       // area flow
    vRefs        Vec_Flt_t;       // ref counters
    pMem        Vec_Set_t;        // cut storage
    vTemp       *Vec_Int_t;       // temporary
    float (*pCutCmp) (Jf_Cut_t *, Jf_Cut_t *);// procedure to compare cuts
    clkStart          abctime;    // starting time
    word             CutCount[4]; // statistics
    nCoarse              int32;     // coarse nodes
}

func Jf_ObjIsUnit(p *Gia_Obj_t) int32 { return !p.fMark0;                                       }
func Jf_ObjCleanUnit(p *Gia_Obj_t) { assert((Jf_ObjIsUnit(p) != 0), "Jf_ObjIsUnit(p)"); p.fMark0 = 1;                  }
func Jf_ObjSetUnit(p *Gia_Obj_t) { p.fMark0 = 0;                                           }

func Jf_ObjCutH(p *Jf_Man_t, i int32) int32 { return Vec_IntEntry(&p.vCuts, i);                       }
func Jf_ObjCuts(p *Jf_Man_t, i int32) *int32 { return (int *)Vec_SetEntry(&p.pMem, Jf_ObjCutH(p, i));  }
func Jf_ObjCutBest(p *Jf_Man_t, i int32) *int32 { return Jf_ObjCuts(p, i) + 1;                             }
func Jf_ObjArr(p *Jf_Man_t, i int32) int32 { return Vec_IntEntry(&p.vArr, i);                        }
func Jf_ObjDep(p *Jf_Man_t, i int32) int32 { return Vec_IntEntry(&p.vDep, i);                        }
func Jf_ObjFlow(p *Jf_Man_t, i int32) float32 { return Vec_FltEntry(&p.vFlow, i);                       }
func Jf_ObjRefs(p *Jf_Man_t, i int32) float32 { return Vec_FltEntry(&p.vRefs, i);                       }
//static inline int    Jf_ObjLit( int i, int c )            { return i;                                                }
func Jf_ObjLit(i int32, c int32) int32 { return Abc_Var2Lit( i, c );                              }

func Jf_CutSize(pCut *int32) int32 { return pCut[0] & 0xF;                                    }  //  4 bits
func Jf_CutCost(pCut *int32) int32 { return (pCut[0] >> 4) & 0xF;                             }  //  4 bits
func Jf_CutFunc(pCut *int32) int32 { return (uint32(pCut[0]) >> 8);                         }  // 24 bits
func Jf_CutSetAll(f int32, c int32, s int32) int32 { return (f << 8) | (c << 4) | s;                          }
func Jf_CutSetSize(pCut *int32, s int32) { assert(s>=0 && s<16, "s>=0 && s<16"); pCut[0] ^= (Jf_CutSize(pCut) ^ s); }
func Jf_CutSetCost(pCut *int32, c int32) { assert(c>=0 && c<16, "c>=0 && c<16"); pCut[0] ^=((Jf_CutCost(pCut) ^ c) << 4); }
func Jf_CutSetFunc(pCut *int32, f int32) { assert(f>=0, "f>=0"); pCut[0] ^=((Jf_CutFunc(pCut) ^ f) << 8);   }

func Jf_CutFuncClass(pCut *int32) int32 { return Abc_Lit2Var(Jf_CutFunc(pCut));                    }
func Jf_CutFuncCompl(pCut *int32) int32 { return Abc_LitIsCompl(Jf_CutFunc(pCut));                 }
func Jf_CutLits(pCut *int32) *int32 { return pCut + 1;                                         }
func Jf_CutLit(pCut *int32, i int32) int32 { assert(i, "i");return pCut[i];                                }
//static inline int    Jf_CutVar( int * pCut, int i )         { assert(i); return pCut[i];                               }
func Jf_CutVar(pCut *int32, i int32) int32 {  assert(i, "i");return Abc_Lit2Var(pCut[i]);                  }
func Jf_CutIsTriv(pCut *int32, i int32) int32 { return Jf_CutSize(pCut) == 1 && Jf_CutVar(pCut, 1) == i; } 
func Jf_CutCnfSizeF(p *Jf_Man_t, f int32) int32 { return Vec_IntEntry( p.vCnfs, f );                      }
func Jf_CutCnfSize(p *Jf_Man_t, c *int32) int32 { return Jf_CutCnfSizeF( p, Jf_CutFuncClass(c) );          }

func Jf_ObjFunc0(p *Gia_Obj_t, c *int32) int32 { return Abc_LitNotCond(Jf_CutFunc(c), Gia_ObjFaninC0(p)); } 
func Jf_ObjFunc1(p *Gia_Obj_t, c *int32) int32 { return Abc_LitNotCond(Jf_CutFunc(c), Gia_ObjFaninC1(p)); } 

// Original macro:
// #define Jf_ObjForEachCut( pList, pCut, i )   for ( i = 0, pCut = pList + 1; i < pList[0]; i++, pCut += Jf_CutSize(pCut) + 1 )

// Original macro:
// #define Jf_CutForEachLit( pCut, Lit, i )     for ( i = 1; i <= Jf_CutSize(pCut) && (Lit = Jf_CutLit(pCut, i)); i++ )

// Original macro:
// #define Jf_CutForEachVar( pCut, Var, i )     for ( i = 1; i <= Jf_CutSize(pCut) && (Var = Jf_CutVar(pCut, i)); i++ )


// extern int Kit_TruthToGia( Gia_Man_t * pMan, unsigned * pTruth, int nVars, Vec_Int_t * vMemory, Vec_Int_t * vLeaves, int fHash );


////////////////////////////////////////////////////////////////////////
///                     FUNCTION DEFINITIONS                         ///
////////////////////////////////////////////////////////////////////////

/**Function*************************************************************

  Synopsis    [Derives CNF for the mapped GIA.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Jf_ManGenCnf(uTruth word, iLitOut int32, vLeaves *Vec_Int_t, vLits *Vec_Int_t, vClas *Vec_Int_t, vCover *Vec_Int_t) {
    if uTruth == 0 || ~uTruth == 0 {
        Vec_IntPush( vClas, Vec_IntSize(vLits) );
        Vec_IntPush( vLits, Abc_LitNotCond(iLitOut, (uTruth == 0)) );
    
    } else {
        var i, k, c, Literal, Cube int32;
        assert(Vec_IntSize(vLeaves) > 0, "Vec_IntSize(vLeaves) > 0");
        for  c = 0; c < 2; c ++  {
            var RetValue int32  = Kit_TruthIsop( (unsigned *)&uTruth, Vec_IntSize(vLeaves), vCover, 0 );
            assert(RetValue == 0, "RetValue == 0");
            Vec_IntForEachEntry( vCover, Cube, i )
            {
                Vec_IntPush( vClas, Vec_IntSize(vLits) );
                Vec_IntPush( vLits, Abc_LitNotCond(iLitOut, c) );
                for  k = 0; k < Vec_IntSize(vLeaves); k++  {
                    Literal = 3 & (Cube >> (k << 1));
                    if Literal == 1 {      // '0'  -> pos lit
                        Vec_IntPush( vLits, Abc_LitNotCond(Vec_IntEntry(vLeaves, k), 0) );
                    } else if Literal == 2 { // '1'  -> neg lit
                        Vec_IntPush( vLits, Abc_LitNotCond(Vec_IntEntry(vLeaves, k), 1) );
                    } else if Literal != 0 {
                        assert(false, "0"); }
                }
            }
            uTruth = ~uTruth;
        }
    }
}
func Jf_ManCreateCnfRemap(p *Gia_Man_t, vLits *Vec_Int_t, vClas *Vec_Int_t, fAddOrCla int32) *Cnf_Dat_t {
    var pCnf *Cnf_Dat_t; 
    var pObj *Gia_Obj_t;
    var i int32;
    var Entry int32;
    var pMap *int32;
    var nVars int32;
    if fAddOrCla != 0 {
        Vec_IntPush( vClas, Vec_IntSize(vLits) );
        Gia_ManForEachPo( p, pObj, i )
            Vec_IntPush( vLits, Abc_Var2Lit(Gia_ObjId(p, pObj), 0) );
    }
    // label nodes present in the mapping
    Vec_IntForEachEntry( vLits, Entry, i )
        Gia_ManObj(p, Abc_Lit2Var(Entry)).fMark0 = 1;
    // create variable map
    pMap = ABC_FALLOC( int, Gia_ManObjNum(p) );
    Gia_ManForEachObjReverse( p, pObj, i )
        if ( pObj.fMark0 )
            pObj.fMark0 = 0;
pMap[i] = nVars++;
    // relabel literals
    Vec_IntForEachEntry( vLits, Entry, i )
        Vec_IntWriteEntry( vLits, i, Abc_Lit2LitV(pMap, Entry) );
    // generate CNF
    pCnf = new(Cnf_Dat_t);
    pCnf.pMan = (Aig_Man_t *)p;
    pCnf.nVars = nVars;
    pCnf.nLiterals = Vec_IntSize(vLits);
    pCnf.nClauses = Vec_IntSize(vClas);
    pCnf.pClauses = ABC_ALLOC( int *, pCnf.nClauses+1 );
    pCnf.pClauses[0] = Vec_IntReleaseArray(vLits);
    Vec_IntForEachEntry( vClas, Entry, i )
        pCnf.pClauses[i] = pCnf.pClauses[0] + Entry;
    pCnf.pClauses[i] = pCnf.pClauses[0] + pCnf.nLiterals;
    pCnf.pVarNums = pMap;
    return pCnf;
}
func Jf_ManCreateCnf(p *Gia_Man_t, vLits *Vec_Int_t, vClas *Vec_Int_t) *Cnf_Dat_t {
    var pCnf *Cnf_Dat_t; 
    var i, Entry, iOut int32;
    // generate CNF
    pCnf = new(Cnf_Dat_t);
    pCnf.pMan = (Aig_Man_t *)p;
    pCnf.nVars = Gia_ManObjNum(p);
    pCnf.nLiterals = Vec_IntSize(vLits);
    pCnf.nClauses = Vec_IntSize(vClas);
    pCnf.pClauses = ABC_ALLOC( int *, pCnf.nClauses+1 );
    pCnf.pClauses[0] = Vec_IntReleaseArray(vLits);
    Vec_IntForEachEntry( vClas, Entry, i )
        pCnf.pClauses[i] = pCnf.pClauses[0] + Entry;
    pCnf.pClauses[i] = pCnf.pClauses[0] + pCnf.nLiterals;
    // create mapping of objects into their clauses
    pCnf.pObj2Clause = ABC_FALLOC( int, Gia_ManObjNum(p) );
    pCnf.pObj2Count  = ABC_FALLOC( int, Gia_ManObjNum(p) );
    for  i = 0; i < pCnf.nClauses; i++  {
        iOut = Abc_Lit2Var(pCnf.pClauses[i][0]);
        if pCnf.pObj2Clause[iOut] == -1 {
            pCnf.pObj2Clause[iOut] = i;
            pCnf.pObj2Count[iOut] = 1;
        
        } else {
            assert(pCnf.pObj2Count[iOut] > 0, "pCnf.pObj2Count[iOut] > 0");
            pCnf.pObj2Count[iOut]++;
        }
    }
    return pCnf;
}

/**Function*************************************************************

  Synopsis    [Computing references while discounting XOR/MUX.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Jf_ManInitRefs(pMan *Jf_Man_t) *float32 {
    var p *Gia_Man_t  = pMan.pGia;
    var pObj, pCtrl, pData0, pData1 *Gia_Obj_t;
    var pRes *float32; var i int32;
    assert(p.Refs == nil, "p.Refs == nil");
    p.Refs = ABC_CALLOC( int, Gia_ManObjNum(p) );
    Gia_ManForEachAnd( p, pObj, i )
    {
        Gia_ObjRefFanin0Inc( p, pObj );
        if Gia_ObjIsBuf(pObj) {
            continue; }
        Gia_ObjRefFanin1Inc( p, pObj );
        if !Gia_ObjIsMuxType(pObj) {
            continue; }
        // discount XOR/MUX
        pCtrl = Gia_ObjRecognizeMux( pObj, &pData1, &pData0 );
        Gia_ObjRefDec( p, Gia_Regular(pCtrl) );
        if Gia_Regular(pData1) == Gia_Regular(pData0) {
            Gia_ObjRefDec( p, Gia_Regular(pData1) ); }
    }
    Gia_ManForEachCo( p, pObj, i )
        Gia_ObjRefFanin0Inc( p, pObj );
    // mark XOR/MUX internal nodes, which are not used elsewhere
    if pMan.pPars.fCoarsen {
        pMan.nCoarse = 0;
        Gia_ManForEachAnd( p, pObj, i )
        {
            if !Gia_ObjIsMuxType(pObj) {
                continue; }
            if Gia_ObjRefNum(p, Gia_ObjFanin0(pObj)) == 1 {
                Jf_ObjSetUnit(Gia_ObjFanin0(Gia_ObjFanin0(pObj)));
                Jf_ObjSetUnit(Gia_ObjFanin0(Gia_ObjFanin1(pObj)));
                Jf_ObjCleanUnit(Gia_ObjFanin0(pObj));
pMan.nCoarse++;
            }
            if Gia_ObjRefNum(p, Gia_ObjFanin1(pObj)) == 1 {
                Jf_ObjSetUnit(Gia_ObjFanin1(Gia_ObjFanin0(pObj)));
                Jf_ObjSetUnit(Gia_ObjFanin1(Gia_ObjFanin1(pObj)));
                Jf_ObjCleanUnit(Gia_ObjFanin1(pObj));
pMan.nCoarse++;
            }
        }
    }
    // multiply by factor
    pRes = ABC_ALLOC( float, Gia_ManObjNum(p) );
    for  i = 0; i < Gia_ManObjNum(p); i++  {
        pRes[i] = Abc_MaxInt( 1, p.Refs[i] );
    }
    return pRes;
}

/**Function*************************************************************

  Synopsis    []

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Jf_ManProfileClasses(p *Jf_Man_t) {
    var pObj *Gia_Obj_t; 
    var Counts [595]int32 ;
    var Costs [595]int32 ;
    var i, iFunc, Total, CostTotal, Other, CostOther int32;
    fmt.Printf("DSD classes that appear in more than %.1f %% of mapped nodes:\n", 0.1 * p.pPars.nVerbLimit);
    Gia_ManForEachAnd( p.pGia, pObj, i )
        if !Gia_ObjIsBuf(pObj) && Gia_ObjRefNumId(p.pGia, i) {
            iFunc = Jf_CutFuncClass( Jf_ObjCutBest(p, i) );
            assert(iFunc < 595, "iFunc < 595");
            if p.pPars.fGenCnf {
                Costs[iFunc] += Jf_CutCnfSizeF(p, iFunc);
                CostTotal += Jf_CutCnfSizeF(p, iFunc);
            }
            Counts[iFunc]++;
            Total++;
        }
    CostTotal = Abc_MaxInt(CostTotal, 1);
    Total = Abc_MaxInt(Total, 1);
    for  i = 0; i < 595; i++  {
        if Counts[i] && 100.0 * Counts[i] / Total >= 0.1 * p.pPars.nVerbLimit {
            fmt.Printf("%5d  :  ", i);
            fmt.Printf("%-20s   ", Sdm_ManReadDsdStr(p.pDsd, i));
            fmt.Printf("%8d  ", Counts[i]);
            fmt.Printf("%5.1f %%   ", 100.0 * Counts[i] / Total);
            fmt.Printf("%8d  ", Costs[i]);
            fmt.Printf("%5.1f %%", 100.0 * Costs[i] / CostTotal);
            fmt.Printf("\n");
        
        } else {
            Other += Counts[i];
            CostOther += Costs[i];
        }
    }
    fmt.Printf("Other  :  ");
    fmt.Printf("%-20s   ", "");
    fmt.Printf("%8d  ", Other);
    fmt.Printf("%5.1f %%   ", 100.0 * Other / Total);
    fmt.Printf("%8d  ", CostOther);
    fmt.Printf("%5.1f %%", 100.0 * CostOther / CostTotal);
    fmt.Printf("\n");
}

/**Function*************************************************************

  Synopsis    [Manager manipulation.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Jf_ManAlloc(pGia *Gia_Man_t, pPars *Jf_Par_t) *Jf_Man_t {
    var p *Jf_Man_t;
    assert(pPars.nLutSize <= JF_LEAF_MAX, "pPars.nLutSize <= JF_LEAF_MAX");
    assert(pPars.nCutNum <= JF_CUT_MAX, "pPars.nCutNum <= JF_CUT_MAX");
    Vec_IntFreeP( &pGia.vMapping );
    p = new(Jf_Man_t);
    p.pGia      = pGia;
    p.pPars     = pPars;
    if pPars.fCutMin && !pPars.fFuncDsd {
        p.vTtMem = Vec_MemAllocForTT( pPars.nLutSize, 0 );
    } else if pPars.fCutMin && pPars.fFuncDsd {
        p.pDsd = Sdm_ManRead();
        if pPars.fGenCnf {
            p.vCnfs = Vec_IntStart( 595 );
            Sdm_ManReadCnfCosts( p.pDsd, Vec_IntArray(p.vCnfs), Vec_IntSize(p.vCnfs) );
        }
    }
    Vec_IntFill( &p.vCuts, Gia_ManObjNum(pGia), 0 );
    Vec_IntFill( &p.vArr,  Gia_ManObjNum(pGia), 0 );
    Vec_IntFill( &p.vDep,  Gia_ManObjNum(pGia), 0 );
    Vec_FltFill( &p.vFlow, Gia_ManObjNum(pGia), 0 );
    p.vRefs.nCap = p.vRefs.nSize = Gia_ManObjNum(pGia);
    p.vRefs.pArray = Jf_ManInitRefs( p );
    Vec_SetAlloc_( &p.pMem, 20 );
    p.vTemp     = Vec_IntAlloc( 1000 );
    p.clkStart  = Abc_Clock();
    return p;
}
func Jf_ManFree(p *Jf_Man_t) {
    if p.pPars.fVerbose && p.pDsd {
        Sdm_ManPrintDsdStats( p.pDsd, 0 ); }
    if p.pPars.fVerbose && p.vTtMem {
        fmt.Printf("Unique truth tables = %d. Memory = %.2f MB   ", Vec_MemEntryNum(p.vTtMem), Vec_MemMemory(p.vTtMem) / (1<<20)); 
        Abc_PrintTime( 1, "Time", Abc_Clock() - p.clkStart );
    }
    if p.pPars.fVeryVerbose && p.pPars.fCutMin && p.pPars.fFuncDsd {
        Jf_ManProfileClasses( p ); }
    if p.pPars.fCoarsen {
        Gia_ManCleanMark0( p.pGia ); }
    p.pGia.Refs = nil;
    p.vCuts.pArray = nil;
    p.vArr.pArray = nil;
    p.vDep.pArray = nil;
    p.vFlow.pArray = nil;
    p.vRefs.pArray = nil;
    if p.pPars.fCutMin && !p.pPars.fFuncDsd {
        Vec_MemHashFree( p.vTtMem );
        Vec_MemFree( p.vTtMem );
    }
    Vec_IntFreeP( &p.vCnfs );
    Vec_SetFree_( &p.pMem );
    Vec_IntFreeP( &p.vTemp );
    p = nil;
}

/**Function*************************************************************

  Synopsis    [Cut functions.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Jf_CutPrint(pCut *int32) {
    var i int32; 
    fmt.Printf("%d {", Jf_CutSize(pCut));
    for  i = 1; i <= Jf_CutSize(pCut); i++  {
        fmt.Printf(" %d", Jf_CutLit(pCut, i));
    }
    fmt.Printf(" } Func = %d\n", Jf_CutFunc(pCut));
}
func Jf_ObjCutPrint(pCuts *int32) {
    var i int32;
    var pCut *int32; 
    // inlining of macro Jf_ObjForEachCut
    for  i, pCut = 0, pCuts + 1; i < pCuts[0]; i++, pCut += Jf_CutSize(pCut) + 1  {
        Jf_CutPrint( pCut );
    }
    fmt.Printf("\n");
}
func Jf_ObjBestCutConePrint(p *Jf_Man_t, pObj *Gia_Obj_t) {
    var pCut *int32  = Jf_ObjCutBest( p, Gia_ObjId(p.pGia, pObj) );
    fmt.Printf("Best cut of node %d : ", Gia_ObjId(p.pGia, pObj));
    Jf_CutPrint( pCut );
    Gia_ManPrintCone( p.pGia, pObj, Jf_CutLits(pCut), Jf_CutSize(pCut), p.vTemp );
}
func Jf_CutCheck(pCut *int32) {
    var i, k int32;
    for  i = 2; i <= Jf_CutSize(pCut); i++  {
        for  k = 1; k < i; k++  {
            assert(Jf_CutLit(pCut, i) != Jf_CutLit(pCut, k), "Jf_CutLit(pCut, i) != Jf_CutLit(pCut, k)");
        }
    }
}
func Jf_CountBitsSimple(n uint32) int32 {
    var i, Count int32;
    for  i = 0; i < 32; i++  {
        Count += ((n >> i) & 1);
    }
    return Count;
}
func Jf_CountBits32(i uint32) int32 {
    i = i - ((i >> 1) & 0x55555555);
    i = (i & 0x33333333) + ((i >> 2) & 0x33333333);
    i = ((i + (i >> 4)) & 0x0F0F0F0F);
    return (i*(0x01010101))>>24;
}
func Jf_CountBits(i word) int32 {
    i = i - ((i >> 1) & 0x5555555555555555);
    i = (i & 0x3333333333333333) + ((i >> 2) & 0x3333333333333333);
    i = ((i + (i >> 4)) & 0x0F0F0F0F0F0F0F0F);
    return (i*(0x0101010101010101))>>56;
}
func Jf_CutGetSign32(pCut *int32) uint32 {
    var Sign uint32; var i int32; 
    for  i = 1; i <= Jf_CutSize(pCut); i++  {
        Sign |= 1 << (Jf_CutVar(pCut, i) & 0x1F);
    }
    return Sign;
}
func Jf_CutGetSign(pCut *int32) word {
    var Sign word; var i int32; 
    for  i = 1; i <= Jf_CutSize(pCut); i++  {
        Sign |= ((word)1) << (Jf_CutVar(pCut, i) & 0x3F);
    }
    return Sign;
}
func Jf_CutArr(p *Jf_Man_t, pCut *int32) int32 {
    var i, Time int32;
    for  i = 1; i <= Jf_CutSize(pCut); i++  {
        Time = Abc_MaxInt( Time, Jf_ObjArr(p, Jf_CutVar(pCut, i)) );
    }
    return Time + 1; 
}

func Jf_ObjSetBestCut(pCuts *int32, pCut *int32, vTemp *Vec_Int_t) {
    assert(pCuts < pCut, "pCuts < pCut");
    if pCuts++; pCuts < pCut {
        var nBlock int32  = pCut - pCuts;
        var nSize int32  = Jf_CutSize(pCut) + 1;
        Vec_IntGrow( vTemp, nBlock );
        memmove( Vec_IntArray(vTemp), pCuts, sizeof(int) * nBlock );
        memmove( pCuts, pCut, sizeof(int) * nSize );
        memmove( pCuts + nSize, Vec_IntArray(vTemp), sizeof(int) * nBlock );
    }
}
func Jf_CutRef(p *Jf_Man_t, pCut *int32) {
    var i int32;
    for  i = 1; i <= Jf_CutSize(pCut); i++  {
        Gia_ObjRefIncId( p.pGia, Jf_CutVar(pCut, i) );
    }
}
func Jf_CutDeref(p *Jf_Man_t, pCut *int32) {
    var i int32;
    for  i = 1; i <= Jf_CutSize(pCut); i++  {
        Gia_ObjRefDecId( p.pGia, Jf_CutVar(pCut, i) );
    }
}
func Jf_CutFlow(p *Jf_Man_t, pCut *int32) float32 {
    var Flow float32; var i int32; 
    for  i = 1; i <= Jf_CutSize(pCut); i++  {
        Flow += Jf_ObjFlow( p, Jf_CutVar(pCut, i) );
    }
    assert(Flow >= 0, "Flow >= 0");
    return Flow; 
}

/**Function*************************************************************

  Synopsis    [Cut merging.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
static inline int Jf_CutIsContainedOrder( int * pBase, int * pCut ) // check if pCut is contained pBase
{
    var nSizeB int32  = Jf_CutSize(pBase);
    var nSizeC int32  = Jf_CutSize(pCut);
    var i, k int32;
    if nSizeB == nSizeC {
        for  i = 1; i <= nSizeB; i++  {
            if pBase[i] != pCut[i] {
                return 0; }
        }
        return 1;
    }
    assert(nSizeB > nSizeC, "nSizeB > nSizeC"); 
    for  i = k = 1; i <= nSizeB; i++  {
        if pBase[i] > pCut[k] {
            return 0; }
        if pBase[i] == pCut[k] {
            if k++ == nSizeC {
                return 1; }
        }
    }
    return 0;
}
func Jf_CutMergeOrder(pCut0 *int32, pCut1 *int32, pCut *int32, LutSize int32) int32 { 
    var nSize0 int32  = Jf_CutSize(pCut0);
    var nSize1 int32  = Jf_CutSize(pCut1);
    var pC0 *int32  = pCut0 + 1;
    var pC1 *int32  = pCut1 + 1;
    var pC *int32  = pCut + 1;
    var i, k, c, s int32;
    // the case of the largest cut sizes
    if nSize0 == LutSize && nSize1 == LutSize {
        for  i = 0; i < nSize0; i++  {
            if pC0[i] != pC1[i] {
                return 0; }
            pC[i] = pC0[i];
        }
        pCut[0] = LutSize;
        return 1;
    }
    // compare two cuts with different numbers
    i = k = c = s = 0;
    if nSize0 == 0 { goto FlushCut1; }
    if nSize1 == 0 { goto FlushCut0; }
    for {
        if c == LutSize { return 0; }
        if pC0[i] < pC1[k] {
            pC[c++] = pC0[i++];
            if i >= nSize0 { goto FlushCut1; }
        
        } else if pC0[i] > pC1[k] {
            pC[c++] = pC1[k++];
            if k >= nSize1 { goto FlushCut0; }
        
        } else {
            pC[c++] = pC0[i++]; k++;
            if i >= nSize0 { goto FlushCut1; }
            if k >= nSize1 { goto FlushCut0; }
        }
    }

FlushCut0:
    if c + nSize0 > LutSize + i { return 0; }
    for  i < nSize0  {
        pC[c++] = pC0[i++];
    }
    pCut[0] = c;
    return 1;

FlushCut1:
    if c + nSize1 > LutSize + k { return 0; }
    for  k < nSize1  {
        pC[c++] = pC1[k++];
    }
    pCut[0] = c;
    return 1;
}

/**Function*************************************************************

  Synopsis    [Cut merging.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Jf_CutFindLeaf0(pCut *int32, iObj int32) int32 {
    var i int32;
    var nLits int32 = Jf_CutSize(pCut);
    for  i = 1; i <= nLits; i++  {
        if pCut[i] == iObj {
            return i; }
    }
    return i;
}
static inline int Jf_CutIsContained0( int * pBase, int * pCut ) // check if pCut is contained pBase
{
    var i int32;
    var nLits int32 = Jf_CutSize(pCut);
    for  i = 1; i <= nLits; i++  {
        if Jf_CutFindLeaf0(pBase, pCut[i]) > pBase[0] {
            return 0; }
    }
    return 1;
}
func Jf_CutMerge0(pCut0 *int32, pCut1 *int32, pCut *int32, LutSize int32) int32 {
    var nSize0 int32  = Jf_CutSize(pCut0);
    var nSize1 int32 = Jf_CutSize(pCut1);
    var i int32;
    pCut[0] = nSize0;
    for  i = 1; i <= nSize1; i++  {
        if Jf_CutFindLeaf0(pCut0, pCut1[i]) > nSize0 {
            if pCut[0] == LutSize {
                return 0; }
            pCut[++pCut[0]] = pCut1[i];
        }
    }
    memcpy( pCut + 1, pCut0 + 1, sizeof(int) * nSize0 );
    return 1;
}

/**Function*************************************************************

  Synopsis    [Cut merging.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Jf_CutFindLeaf1(pCut *int32, iLit int32) int32 {
    var i int32;
    var nLits int32 = Jf_CutSize(pCut);
    for  i = 1; i <= nLits; i++  {
        if Abc_Lit2Var(pCut[i]) == iLit {
            return i; }
    }
    return i;
}
static inline int Jf_CutIsContained1( int * pBase, int * pCut ) // check if pCut is contained pBase
{
    var i int32;
    var nLits int32 = Jf_CutSize(pCut);
    for  i = 1; i <= nLits; i++  {
        if Jf_CutFindLeaf1(pBase, Abc_Lit2Var(pCut[i])) > pBase[0] {
            return 0; }
    }
    return 1;
}
func Jf_CutMerge1(pCut0 *int32, pCut1 *int32, pCut *int32, LutSize int32) int32 {
    var nSize0 int32  = Jf_CutSize(pCut0);
    var nSize1 int32 = Jf_CutSize(pCut1);
    var i int32;
    pCut[0] = nSize0;
    for  i = 1; i <= nSize1; i++  {
        if Jf_CutFindLeaf1(pCut0, Abc_Lit2Var(pCut1[i])) > nSize0 {
            if pCut[0] == LutSize {
                return 0; }
            pCut[++pCut[0]] = pCut1[i];
        }
    }
    memcpy( pCut + 1, pCut0 + 1, sizeof(int) * nSize0 );
    return 1;
}
func Jf_CutMerge2(pCut0 *int32, pCut1 *int32, pCut *int32, LutSize int32) int32 {
    var ConfigMask int32  = 0x3FFFF; // 18 bits
    var nSize0 int32  = Jf_CutSize(pCut0);
    var nSize1 int32  = Jf_CutSize(pCut1);
    var i, iPlace int32;
    pCut[0] = nSize0;
    for  i = 1; i <= nSize1; i++  {
        iPlace = Jf_CutFindLeaf1(pCut0, Abc_Lit2Var(pCut1[i]));
        if iPlace > nSize0 {
            if pCut[0] == LutSize {
                return 0; }
            pCut[(iPlace = ++pCut[0])] = pCut1[i];
        
        } else if pCut0[iPlace] != pCut1[i] {
            ConfigMask |= (1 << (iPlace+17)); }
        ConfigMask ^= (((i-1) ^ 7) << (3*(iPlace-1)));
    }
    memcpy( pCut + 1, pCut0 + 1, sizeof(int) * nSize0 );
    return ConfigMask;
}


/**Function*************************************************************

  Synopsis    [Cut filtering.]

  Description [Returns the number of cuts after filtering and the last
  cut in the last entry.  If the cut is filtered, its size is set to -1.]
               
  SideEffects [This was found to be 15% slower.]

  SeeAlso     []

***********************************************************************/
func Jf_ObjCutFilterBoth(p *Jf_Man_t, pSto **Jf_Cut_t, c int32) int32 {
    var k, last int32;
    // filter this cut using other cuts
    for  k = 0; k < c; k++  {
        if pSto[c].pCut[0] >= pSto[k].pCut[0] && 
            (pSto[c].Sign & pSto[k].Sign) == pSto[k].Sign && 
             Jf_CutIsContained1(pSto[c].pCut, pSto[k].pCut) {
                pSto[c].pCut[0] = -1;
                return c;
        }
    }
    // filter other cuts using this cut
    for  k = last = 0; k < c; k++  {
        if !(pSto[c].pCut[0] < pSto[k].pCut[0] && 
              (pSto[c].Sign & pSto[k].Sign) == pSto[c].Sign && 
               Jf_CutIsContained1(pSto[k].pCut, pSto[c].pCut)) {
            if last++ == k {
                continue; }
            ABC_SWAP( Jf_Cut_t *, pSto[last-1], pSto[k] );
        }
    }
    assert(last <= c, "last <= c");
    if last < c {
        ABC_SWAP( Jf_Cut_t *, pSto[last], pSto[c] ); }
    return last;
}
func Jf_ObjCutFilter(p *Jf_Man_t, pSto **Jf_Cut_t, c int32) int32 {
    var k int32;
    if p.pPars.fCutMin {
        for  k = 0; k < c; k++  {
            if pSto[c].pCut[0] >= pSto[k].pCut[0] && 
                (pSto[c].Sign & pSto[k].Sign) == pSto[k].Sign && 
                 Jf_CutIsContained1(pSto[c].pCut, pSto[k].pCut) {
                 return 0; }
        }
    
    } else {
        for  k = 0; k < c; k++  {
            if pSto[c].pCut[0] >= pSto[k].pCut[0] && 
                (pSto[c].Sign & pSto[k].Sign) == pSto[k].Sign && 
                 Jf_CutIsContainedOrder(pSto[c].pCut, pSto[k].pCut) {
                 return 0; }
        }
    }
    return 1;
}

/**Function*************************************************************

  Synopsis    [Sorting cuts by size.]

  Description []
               
  SideEffects [Did not really help.]

  SeeAlso     []

***********************************************************************/
func Jf_ObjSortCuts(pSto **Jf_Cut_t, nSize int32) {
    var i, j, best_i int32;
    for  i = 0; i < nSize-1; i++  {
        best_i = i;
        for  j = i+1; j < nSize; j++  {
            if pSto[j].pCut[0] < pSto[best_i].pCut[0] {
                best_i = j; }
        }
        ABC_SWAP( Jf_Cut_t *, pSto[i], pSto[best_i] );
    }
}

/**Function*************************************************************

  Synopsis    [Reference counting.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Jf_CutRef_rec(p *Jf_Man_t, pCut *int32) int32 {
    var i int32;
    var Var int32;
    var Count int32 = Jf_CutCost(pCut);
    // inlining of macro Jf_CutForEachVar
    for  i = 1; i <= Jf_CutSize(pCut) && (Var = Jf_CutVar(pCut, i)); i++  {
        if !Gia_ObjRefIncId(p.pGia, Var) && !(Jf_CutIsTriv(Jf_ObjCutBest(p, Var), Var) != 0) {
            Count += Jf_CutRef_rec( p, Jf_ObjCutBest(p, Var) ); }
    }
    return Count;
}
func Jf_CutDeref_rec(p *Jf_Man_t, pCut *int32) int32 {
    var i int32;
    var Var int32;
    var Count int32 = Jf_CutCost(pCut);
    // inlining of macro Jf_CutForEachVar
    for  i = 1; i <= Jf_CutSize(pCut) && (Var = Jf_CutVar(pCut, i)); i++  {
        if !Gia_ObjRefDecId(p.pGia, Var) && !(Jf_CutIsTriv(Jf_ObjCutBest(p, Var), Var) != 0) {
            Count += Jf_CutDeref_rec( p, Jf_ObjCutBest(p, Var) ); }
    }
    return Count;
}
func Jf_CutAreaOld(p *Jf_Man_t, pCut *int32) int32 {
    var Ela1, Ela2 int32;
    Ela1 = Jf_CutRef_rec( p, pCut );
    Ela2 = Jf_CutDeref_rec( p, pCut );
    assert(Ela1 == Ela2, "Ela1 == Ela2");
    return Ela1;
}

func Jf_CutAreaRef_rec(p *Jf_Man_t, pCut *int32) int32 {
    var i int32;
    var Var int32;
    var Count int32 = Jf_CutCost(pCut);
    // inlining of macro Jf_CutForEachVar
    for  i = 1; i <= Jf_CutSize(pCut) && (Var = Jf_CutVar(pCut, i)); i++  {
        if !Gia_ObjRefIncId(p.pGia, Var) && !(Jf_CutIsTriv(Jf_ObjCutBest(p, Var), Var) != 0) {
            Count += Jf_CutAreaRef_rec( p, Jf_ObjCutBest(p, Var) ); }
        Vec_IntPush( p.vTemp, Var );
    }
    return Count;
}
func Jf_CutAreaRefEdge_rec(p *Jf_Man_t, pCut *int32) int32 {
    var i int32;
    var Var int32;
    var Count int32 = (Jf_CutCost(pCut) << 4) | Jf_CutSize(pCut);
    // inlining of macro Jf_CutForEachVar
    for  i = 1; i <= Jf_CutSize(pCut) && (Var = Jf_CutVar(pCut, i)); i++  {
        if !Gia_ObjRefIncId(p.pGia, Var) && !(Jf_CutIsTriv(Jf_ObjCutBest(p, Var), Var) != 0) {
            Count += Jf_CutAreaRefEdge_rec( p, Jf_ObjCutBest(p, Var) ); }
        Vec_IntPush( p.vTemp, Var );
    }
    return Count;
}
func Jf_CutArea(p *Jf_Man_t, pCut *int32, fEdge int32) int32 {
    var Ela, Entry, i int32;
    Vec_IntClear( p.vTemp );
    if fEdge != 0 {
        Ela = Jf_CutAreaRefEdge_rec( p, pCut );
    } else {
        Ela = Jf_CutAreaRef_rec( p, pCut ); }
    Vec_IntForEachEntry( p.vTemp, Entry, i )
        Gia_ObjRefDecId( p.pGia, Entry );
    return Ela;
}
// returns 1 if MFFC size is less than limit
func Jf_CutCheckMffc_rec(p *Jf_Man_t, pCut *int32, Limit int32) int32 {
    var i, Var int32;
    // inlining of macro Jf_CutForEachVar
    for  i = 1; i <= Jf_CutSize(pCut) && (Var = Jf_CutVar(pCut, i)); i++  {
        var fRecur int32  = (!Gia_ObjRefDecId(p.pGia, Var) && !Jf_CutIsTriv(Jf_ObjCutBest(p, Var), Var));
        Vec_IntPush( p.vTemp, Var );
        if Vec_IntSize(p.vTemp) >= Limit {
            return 0; }
        if (fRecur != 0) && !(Jf_CutCheckMffc_rec( p, Jf_ObjCutBest(p, Var), Limit ) != 0) {
            return 0; }
    }
    return 1;
}
func Jf_CutCheckMffc(p *Jf_Man_t, pCut *int32, Limit int32) int32 {
    var RetValue, Entry, i int32;
    Vec_IntClear( p.vTemp );
    RetValue = Jf_CutCheckMffc_rec( p, pCut, Limit );
    Vec_IntForEachEntry( p.vTemp, Entry, i )
        Gia_ObjRefIncId( p.pGia, Entry );
    return RetValue;
}

/**Function*************************************************************

  Synopsis    [Comparison procedures.]

  Description [Return positive value if the new cut is better than the old cut.]
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Jf_CutCompareDelay(pOld *Jf_Cut_t, pNew *Jf_Cut_t) float32 {
    if pOld.Time    != pNew.Time { return pOld.Time    - pNew.Time; }
    if pOld.pCut[0] != pNew.pCut[0] { return pOld.pCut[0] - pNew.pCut[0]; }
//    if ( pOld->Flow    != pNew->Flow    ) return pOld->Flow    - pNew->Flow;
    if pOld.Flow < pNew.Flow - JF_EPSILON { return -1; }
    if pOld.Flow > pNew.Flow + JF_EPSILON { return 1; }
    return 0;
}
func Jf_CutCompareArea(pOld *Jf_Cut_t, pNew *Jf_Cut_t) float32 {
//    if ( pOld->Flow    != pNew->Flow    ) return pOld->Flow    - pNew->Flow;
    if pOld.Flow < pNew.Flow - JF_EPSILON { return -1; }
    if pOld.Flow > pNew.Flow + JF_EPSILON { return 1; }
    if pOld.pCut[0] != pNew.pCut[0] { return pOld.pCut[0] - pNew.pCut[0]; }
    if pOld.Time    != pNew.Time { return pOld.Time    - pNew.Time; }
    return 0;
}
func Jf_ObjAddCutToStore(p *Jf_Man_t, pSto **Jf_Cut_t, c int32, cMax int32) int32 {
    var pTemp *Jf_Cut_t;
    var k, last, iPivot int32;
    // if the store is empty, add anything
    if c == 0 {
        return 1; }
    // special case when the cut store is full and last cut is better than new cut
    if c == cMax && p.pCutCmp(pSto[c-1], pSto[c]) <= 0 {
        return c; }
    // find place of the given cut in the store
    assert(c <= cMax, "c <= cMax");
    for  iPivot = c-1; iPivot >= 0; iPivot--  {
        if p.pCutCmp(pSto[iPivot], pSto[c]) < 0 { // iPivot-th cut is better than new cut
            break; }
    }
    // filter this cut using other cuts
    if p.pPars.fCutMin {
        for  k = 0; k <= iPivot; k++  {
            if pSto[c].pCut[0] >= pSto[k].pCut[0] && 
                (pSto[c].Sign & pSto[k].Sign) == pSto[k].Sign && 
                 Jf_CutIsContained1(pSto[c].pCut, pSto[k].pCut) {
                    return c; }
        }
    
    } else {
        for  k = 0; k <= iPivot; k++  {
            if pSto[c].pCut[0] >= pSto[k].pCut[0] && 
                (pSto[c].Sign & pSto[k].Sign) == pSto[k].Sign && 
                 Jf_CutIsContainedOrder(pSto[c].pCut, pSto[k].pCut) {
                    return c; }
        }
    }
    // insert this cut after iPivot
    pTemp = pSto[c];
    for  ++iPivot, k = c++; k > iPivot; k--  {
        pSto[k] = pSto[k-1];
    }
    pSto[iPivot] = pTemp;
    // filter other cuts using this cut
    if p.pPars.fCutMin {
        for  k = last = iPivot+1; k < c; k++  {
            if !(pSto[iPivot].pCut[0] <= pSto[k].pCut[0] && 
                  (pSto[iPivot].Sign & pSto[k].Sign) == pSto[iPivot].Sign && 
                   Jf_CutIsContained1(pSto[k].pCut, pSto[iPivot].pCut)) {
                if last++ == k {
                    continue; }
                ABC_SWAP( Jf_Cut_t *, pSto[last-1], pSto[k] );
            }
        }
    
    } else {
        for  k = last = iPivot+1; k < c; k++  {
            if !(pSto[iPivot].pCut[0] <= pSto[k].pCut[0] && 
                  (pSto[iPivot].Sign & pSto[k].Sign) == pSto[iPivot].Sign && 
                   Jf_CutIsContainedOrder(pSto[k].pCut, pSto[iPivot].pCut)) {
                if last++ == k {
                    continue; }
                ABC_SWAP( Jf_Cut_t *, pSto[last-1], pSto[k] );
            }
        }
    }
    c = last;
    // remove the last cut if too many
    if c == cMax + 1 {
        return c - 1; }
    return c;
}
func Jf_ObjPrintStore(p *Jf_Man_t, pSto **Jf_Cut_t, c int32) {
    var i int32;
    for  i = 0; i < c; i++  {
        fmt.Printf("Flow =%9.5f  ", pSto[i].Flow);
        fmt.Printf("Time = %5d   ", pSto[i].Time);
        fmt.Printf("Func = %5d   ", pSto[i].iFunc);
        fmt.Printf("  ");
        Jf_CutPrint( pSto[i].pCut );
    }
    fmt.Printf("\n");
}
func Jf_ObjCheckPtrs(pSto **Jf_Cut_t, c int32) {
    var i, k int32;
    for  i = 1; i < c; i++  {
        for  k = 0; k < i; k++  {
            assert(pSto[k] != pSto[i], "pSto[k] != pSto[i]");
        }
    }
}
func Jf_ObjCheckStore(p *Jf_Man_t, pSto **Jf_Cut_t, c int32, iObj int32) {
    var i, k int32;
    for  i = 1; i < c; i++  {
        assert(p.pCutCmp(pSto[i-1], pSto[i]) <= 0, "p.pCutCmp(pSto[i-1], pSto[i]) <= 0");
    }
    for  i = 1; i < c; i++  {
        for  k = 0; k < i; k++  {
            assert(!Jf_CutIsContained1(pSto[k].pCut, pSto[i].pCut), "!Jf_CutIsContained1(pSto[k].pCut, pSto[i].pCut)");
            assert(!Jf_CutIsContained1(pSto[i].pCut, pSto[k].pCut), "!Jf_CutIsContained1(pSto[i].pCut, pSto[k].pCut)");
        }
    }
}

/**Function*************************************************************

  Synopsis    [Cut minimization.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Jf_TtComputeForCut(p *Jf_Man_t, iFuncLit0 int32, iFuncLit1 int32, pCut0 *int32, pCut1 *int32, pCutOut *int32) int32 {
    var uTruth, uTruth0, uTruth1 [JF_WORD_MAX]word;
    var fCompl, truthId int32;
    var LutSize int32     = p.pPars.nLutSize;
    var nWords int32      = Abc_Truth6WordNum(p.pPars.nLutSize);
    var pTruth0 *word  = Vec_MemReadEntry(p.vTtMem, Abc_Lit2Var(iFuncLit0));
    var pTruth1 *word  = Vec_MemReadEntry(p.vTtMem, Abc_Lit2Var(iFuncLit1));
    Abc_TtCopy( uTruth0, pTruth0, nWords, Abc_LitIsCompl(iFuncLit0) );
    Abc_TtCopy( uTruth1, pTruth1, nWords, Abc_LitIsCompl(iFuncLit1) );
    Abc_TtExpand( uTruth0, LutSize, pCut0 + 1, Jf_CutSize(pCut0), pCutOut + 1, Jf_CutSize(pCutOut) );
    Abc_TtExpand( uTruth1, LutSize, pCut1 + 1, Jf_CutSize(pCut1), pCutOut + 1, Jf_CutSize(pCutOut) );
    fCompl         = int32((uTruth0[0] & uTruth1[0] & 1));
    Abc_TtAnd( uTruth, uTruth0, uTruth1, nWords, fCompl );
    pCutOut[0]     = Abc_TtMinBase( uTruth, pCutOut + 1, pCutOut[0], LutSize );
    assert((uTruth[0] & 1) == 0, "(uTruth[0] & 1) == 0");
    truthId        = Vec_MemHashInsert(p.vTtMem, uTruth);
    return Abc_Var2Lit( truthId, fCompl );
}

/**Function*************************************************************

  Synopsis    [Cut enumeration.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Jf_ObjAssignCut(p *Jf_Man_t, pObj *Gia_Obj_t) {
    var iObj int32  = Gia_ObjId(p.pGia, pObj);
    var pClause [3]int32 = [3]int32{ 1, Jf_CutSetAll(2, 0, 1), Jf_ObjLit(iObj, 0) }; // set function
    assert(Gia_ObjIsCi(pObj) || Gia_ObjIsBuf(pObj), "Gia_ObjIsCi(pObj) || Gia_ObjIsBuf(pObj)");
    Vec_IntWriteEntry( &p.vCuts, iObj, Vec_SetAppend( &p.pMem, pClause, 3 ) );
}
func Jf_ObjPropagateBuf(p *Jf_Man_t, pObj *Gia_Obj_t, fReverse int32) {
    var iObj int32  = Gia_ObjId( p.pGia, pObj );
    var iFanin int32  = Gia_ObjFaninId0( pObj, iObj );
    assert(false, "0");
    assert(Gia_ObjIsBuf(pObj), "Gia_ObjIsBuf(pObj)");
    if fReverse != 0 {
        ABC_SWAP( int, iObj, iFanin ); }
    Vec_IntWriteEntry( &p.vArr,  iObj, Jf_ObjArr(p, iFanin) );
    Vec_FltWriteEntry( &p.vFlow, iObj, Jf_ObjFlow(p, iFanin) );
}
func Jf_ObjHasCutWithSize(pSto **Jf_Cut_t, c int32, nSize int32) int32 {
    var i int32;
    for  i = 0; i < c; i++  {
        if pSto[i].pCut[0] <= nSize {
            return 1; }
    }
    return 0;
}
func Jf_ObjComputeCuts(p *Jf_Man_t, pObj *Gia_Obj_t, fEdge int32) {
    var LutSize int32  = p.pPars.nLutSize;
    var CutNum int32  = p.pPars.nCutNum;
    var iObj int32  = Gia_ObjId(p.pGia, pObj);
    var Sign0 [JF_CUT_MAX+2]word; // signatures of the first cut
    var Sign1 [JF_CUT_MAX+2]word; // signatures of the second cut
    var Sto [JF_CUT_MAX+2]Jf_Cut_t;   // cut storage
    var pSto [JF_CUT_MAX+2]*Jf_Cut_t;  // pointers to cut storage
    var pCut0, pCut1, pCuts0, pCuts1 *int32;
    var nOldSupp, Config, i, k, c int32;
    // prepare cuts
    for  i = 0; i <= CutNum+1; i++  {
        pSto[i] = Sto + i, pSto[i].Cost = 0, pSto[i].iFunc = ~0;
    }
    // compute signatures
    pCuts0 = Jf_ObjCuts( p, Gia_ObjFaninId0(pObj, iObj) );
    // inlining of macro Jf_ObjForEachCut
    for  i, pCut0 = 0, pCuts0 + 1; i < pCuts0[0]; i++, pCut0 += Jf_CutSize(pCut0) + 1  {
        Sign0[i] = Jf_CutGetSign( pCut0 );
    }
    // compute signatures
    pCuts1 = Jf_ObjCuts( p, Gia_ObjFaninId1(pObj, iObj) );
    // inlining of macro Jf_ObjForEachCut
    for  i, pCut1 = 0, pCuts1 + 1; i < pCuts1[0]; i++, pCut1 += Jf_CutSize(pCut1) + 1  {
        Sign1[i] = Jf_CutGetSign( pCut1 );
    }
    // merge cuts
    p.CutCount[0] += pCuts0[0] * pCuts1[0];
    // inlining of macro Jf_ObjForEachCut
    for  i, pCut0 = 0, pCuts0 + 1; i < pCuts0[0]; i++, pCut0 += Jf_CutSize(pCut0) + 1  {
    // inlining of macro Jf_ObjForEachCut
    for  k, pCut1 = 0, pCuts1 + 1; k < pCuts1[0]; k++, pCut1 += Jf_CutSize(pCut1) + 1  {
        if Jf_CountBits(Sign0[i] | Sign1[k]) > LutSize {
            continue; }
        p.CutCount[1]++;        
        if !p.pPars.fCutMin {
            if !(Jf_CutMergeOrder(pCut0, pCut1, pSto[c].pCut, LutSize) != 0) {
                continue; }
            pSto[c].Sign = Sign0[i] | Sign1[k];
        
        } else if p.pPars.fFuncDsd {
            if !(Config = Jf_CutMerge2(pCut0, pCut1, pSto[c].pCut, LutSize)) {
                continue; }
            pSto[c].Sign = Sign0[i] | Sign1[k];
            nOldSupp = pSto[c].pCut[0];
            pSto[c].iFunc = Sdm_ManComputeFunc( p.pDsd, Jf_ObjFunc0(pObj, pCut0), Jf_ObjFunc1(pObj, pCut1), pSto[c].pCut, Config, 0 );
            if pSto[c].iFunc == -1 {
                continue; }
            if p.pPars.fGenCnf && Jf_CutCnfSizeF(p, Abc_Lit2Var(pSto[c].iFunc)) >= 12 { // no more than 15
                continue; }
            assert(pSto[c].pCut[0] <= nOldSupp, "pSto[c].pCut[0] <= nOldSupp");
            if pSto[c].pCut[0] < nOldSupp {
                pSto[c].Sign = Jf_CutGetSign( pSto[c].pCut ); }
        
        } else {
            if !(Jf_CutMergeOrder(pCut0, pCut1, pSto[c].pCut, LutSize) != 0) {
                continue; }
            pSto[c].Sign = Sign0[i] | Sign1[k];
            nOldSupp = pSto[c].pCut[0];
            pSto[c].iFunc = Jf_TtComputeForCut( p, Jf_ObjFunc0(pObj, pCut0), Jf_ObjFunc1(pObj, pCut1), pCut0, pCut1, pSto[c].pCut );
            assert(pSto[c].pCut[0] <= nOldSupp, "pSto[c].pCut[0] <= nOldSupp");
            if pSto[c].pCut[0] < nOldSupp {
                pSto[c].Sign = Jf_CutGetSign( pSto[c].pCut ); }
            if pSto[c].iFunc >= (1 << 24) {
                fmt.Printf("Hard limit on the number of different Boolean functions (2^23) is reached. Quitting...\n");
os.Exit(int(1)); }
        }
        p.CutCount[2]++;
        pSto[c].Time = p.pPars.fAreaOnly ? 0 : Jf_CutArr(p, pSto[c].pCut);
        pSto[c].Flow = Jf_CutFlow(p, pSto[c].pCut);
        c = Jf_ObjAddCutToStore( p, pSto, c, CutNum );
        assert(c <= CutNum, "c <= CutNum");
    }
    }
//    Jf_ObjPrintStore( p, pSto, c );
//    Jf_ObjCheckStore( p, pSto, c, iObj );
    // add two variable cut
    if !(Jf_ObjIsUnit(pObj) != 0) && !(Jf_ObjHasCutWithSize(pSto, c, 2) != 0) {
        assert((Jf_ObjIsUnit(Gia_ObjFanin0(pObj)) != 0) && (Jf_ObjIsUnit(Gia_ObjFanin1(pObj)) != 0), "Jf_ObjIsUnit(Gia_ObjFanin0(pObj)) && Jf_ObjIsUnit(Gia_ObjFanin1(pObj))");
        if p.pPars.fCutMin { pSto[c].iFunc = 4; }  // set function (DSD only!)
        pSto[c].pCut[0] = 2; 
        pSto[c].pCut[1] = Jf_ObjLit(Gia_ObjFaninId0(pObj, iObj), Gia_ObjFaninC0(pObj)); 
        pSto[c].pCut[2] = Jf_ObjLit(Gia_ObjFaninId1(pObj, iObj), Gia_ObjFaninC1(pObj)); 
        c++;
    }
    // add elementary cut
    if (Jf_ObjIsUnit(pObj) != 0) && !(p.pPars.fCutMin && (Jf_ObjHasCutWithSize(pSto, c, 1) != 0)) {
        if p.pPars.fCutMin { pSto[c].iFunc = 2; }  // set function
        pSto[c].pCut[0] = 1;
        pSto[c].pCut[1] = Jf_ObjLit(iObj, 0);
        c++;
    }
    // reorder cuts
//    Jf_ObjSortCuts( pSto + 1, c - 1 );
//    Jf_ObjCheckPtrs( pSto, CutNum );
    // find cost of the best cut
    pSto[0].Cost = p.pPars.fGenCnf ? Jf_CutCnfSizeF(p, Abc_Lit2Var(pSto[0].iFunc)) : 1;
    assert(pSto[0].Cost >= 0, "pSto[0].Cost >= 0");
    // save best info
    assert(pSto[0].Flow >= 0, "pSto[0].Flow >= 0");
    Vec_IntWriteEntry( &p.vArr,  iObj, pSto[0].Time );
    Vec_FltWriteEntry( &p.vFlow, iObj, (pSto[0].Flow + (fEdge ? pSto[0].pCut[0] : pSto[0].Cost)) / Jf_ObjRefs(p, iObj) );
    // add cuts to storage cuts
    Vec_IntClear( p.vTemp );
    Vec_IntPush( p.vTemp, c );
    for  i = 0; i < c; i++  {
        pSto[i].Cost = p.pPars.fGenCnf ? Jf_CutCnfSizeF(p, Abc_Lit2Var(pSto[i].iFunc)) : 1;
        Vec_IntPush( p.vTemp, Jf_CutSetAll(pSto[i].iFunc, pSto[i].Cost, pSto[i].pCut[0]) );
        for  k = 1; k <= pSto[i].pCut[0]; k++  {
            Vec_IntPush( p.vTemp, pSto[i].pCut[k] );
        }
    }
    Vec_IntWriteEntry( &p.vCuts, iObj, Vec_SetAppend(&p.pMem, Vec_IntArray(p.vTemp), Vec_IntSize(p.vTemp)) );
    p.CutCount[3] += c;
}
func Jf_ManComputeCuts(p *Jf_Man_t, fEdge int32) {
    var pObj *Gia_Obj_t; var i int32;
    if p.pPars.fVerbose {
        fmt.Printf("Aig: CI = %d  CO = %d  AND = %d    ", Gia_ManCiNum(p.pGia), Gia_ManCoNum(p.pGia), Gia_ManAndNum(p.pGia));
        fmt.Printf("LutSize = %d  CutMax = %d  Rounds = %d\n", p.pPars.nLutSize, p.pPars.nCutNum, p.pPars.nRounds);
        fmt.Printf("Computing cuts...\r");
        fflush( os.Stdout );
    }
    Gia_ManForEachObj( p.pGia, pObj, i )
    {
        if Gia_ObjIsCi(pObj) || Gia_ObjIsBuf(pObj) {
            Jf_ObjAssignCut( p, pObj ); }
        if Gia_ObjIsBuf(pObj) {
            Jf_ObjPropagateBuf( p, pObj, 0 );
        } else if Gia_ObjIsAnd(pObj) {
            Jf_ObjComputeCuts( p, pObj, fEdge ); }
    }
    if p.pPars.fVerbose {
        fmt.Printf("CutPair = %d  ", int64(p.CutCount[0]));
        fmt.Printf("Merge = %d  ", int64(p.CutCount[1]));
        fmt.Printf("Eval = %d  ", int64(p.CutCount[2]));
        fmt.Printf("Cut = %d  ", int64(p.CutCount[3]));
        Abc_PrintTime( 1, "Time",  Abc_Clock() - p.clkStart );
        fmt.Printf("Memory:  ");
        fmt.Printf("Gia = %.2f MB  ", Gia_ManMemory(p.pGia) / (1<<20));
        fmt.Printf("Man = %.2f MB  ", 6.0 * sizeof(int) * Gia_ManObjNum(p.pGia) / (1<<20));
        fmt.Printf("Cuts = %.2f MB", Vec_ReportMemory(&p.pMem) / (1<<20));
        if p.nCoarse {
        fmt.Printf("   Coarse = %d (%.1f %%)", p.nCoarse, 100.0 * p.nCoarse / Gia_ManObjNum(p.pGia)); }
        fmt.Printf("\n");
        fflush( os.Stdout );
    }
}


/**Function*************************************************************

  Synopsis    [Computing delay/area.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Jf_ManComputeDelay(p *Jf_Man_t, fEval int32) int32 {
    var pObj *Gia_Obj_t; 
    var i, Delay int32;
    if fEval != 0 {
        Gia_ManForEachObj( p.pGia, pObj, i )
            if ( Gia_ObjIsBuf(pObj) )
                Jf_ObjPropagateBuf( p, pObj, 0 );
            else if ( Gia_ObjIsAnd(pObj) && Gia_ObjRefNum(p.pGia, pObj) > 0 )
                Vec_IntWriteEntry( &p.vArr, i, Jf_CutArr(p, Jf_ObjCutBest(p, i)) );
    }
    Gia_ManForEachCoDriver( p.pGia, pObj, i )
    {
        assert(Gia_ObjRefNum(p.pGia, pObj) > 0, "Gia_ObjRefNum(p.pGia, pObj) > 0");
        Delay = Abc_MaxInt( Delay, Jf_ObjArr(p, Gia_ObjId(p.pGia, pObj)) );
    }
    return Delay;
}
func Jf_ManComputeRefs(p *Jf_Man_t) int32 {
    var pObj *Gia_Obj_t; 
    var nRefsNew float32; var i int32;
var pCut *int32;
    var pRefs *float32  = Vec_FltArray(&p.vRefs);
    var pFlow *float32  = Vec_FltArray(&p.vFlow);
    assert(p.pGia.Refs != nil, "p.pGia.Refs != nil");
    memset( p.pGia.Refs, 0, sizeof(int) * Gia_ManObjNum(p.pGia) );
    p.pPars.Area = p.pPars.Edge = 0;
    Gia_ManForEachObjReverse( p.pGia, pObj, i )
    {
        if Gia_ObjIsCo(pObj) || Gia_ObjIsBuf(pObj) {
            Gia_ObjRefInc( p.pGia, Gia_ObjFanin0(pObj) );
        } else if Gia_ObjIsAnd(pObj) && Gia_ObjRefNum(p.pGia, pObj) > 0 {
            assert((Jf_ObjIsUnit(pObj) != 0), "Jf_ObjIsUnit(pObj)");
            pCut = Jf_ObjCutBest(p, i);
            Jf_CutRef( p, pCut );
            if p.pPars.fGenCnf {
            p.pPars.Clause += Jf_CutCnfSize(p, pCut); }
            p.pPars.Edge += Jf_CutSize(pCut);
            p.pPars.Area++;
        }
    }
    // blend references and normalize flow
    for  i = 0; i < Gia_ManObjNum(p.pGia); i++  {
        if p.pPars.fOptEdge {
            nRefsNew = Abc_MaxFloat( 1, 0.8 * pRefs[i] + 0.2 * p.pGia.Refs[i] );
        } else {
            nRefsNew = Abc_MaxFloat( 1, 0.2 * pRefs[i] + 0.8 * p.pGia.Refs[i] ); }
        pFlow[i] = pFlow[i] * pRefs[i] / nRefsNew;
        pRefs[i] = nRefsNew;
        assert(pFlow[i] >= 0, "pFlow[i] >= 0");
    }
    // compute delay
    p.pPars.Delay = Jf_ManComputeDelay( p, 1 );
    return p.pPars.Area;
}

/**Function*************************************************************

  Synopsis    [Mapping rounds.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Jf_ObjComputeBestCut(p *Jf_Man_t, pObj *Gia_Obj_t, fEdge int32, fEla int32) {
    var i int32;
    var iObj int32 = Gia_ObjId( p.pGia, pObj );
    var pCuts *int32  = Jf_ObjCuts( p, iObj );
    var pCut *int32;
    var pCutBest *int32 = nil;
    var Time int32 = ABC_INFINITY;
    var TimeBest int32 = ABC_INFINITY;
    var Area float32;
    var AreaBest float32 = ABC_INFINITY;
    // inlining of macro Jf_ObjForEachCut
    for  i, pCut = 0, pCuts + 1; i < pCuts[0]; i++, pCut += Jf_CutSize(pCut) + 1  {
        if Jf_CutIsTriv(pCut, iObj) != 0 { continue; }
        if (fEdge != 0) && !(fEla != 0) { 
            Jf_CutSetCost(pCut, Jf_CutSize(pCut)); }
        if fEla != 0 {
            Area = Jf_CutArea(p, pCut, fEdge);
        } else {
            Area = Jf_CutFlow(p, pCut) + Jf_CutCost(pCut);
        }
        if pCutBest == nil || AreaBest > Area + JF_EPSILON || (AreaBest > Area - JF_EPSILON && TimeBest > (Time = Jf_CutArr(p, pCut))) {
            pCutBest = pCut;
AreaBest = Area;
TimeBest = Time; }
    }
    Vec_IntWriteEntry( &p.vArr,  iObj, Jf_CutArr(p, pCutBest) );
    if !(fEla != 0) {
    Vec_FltWriteEntry( &p.vFlow, iObj, AreaBest / Jf_ObjRefs(p, iObj) ); }
    Jf_ObjSetBestCut( pCuts, pCutBest, p.vTemp );
//    Jf_CutPrint( Jf_ObjCutBest(p, iObj) ); printf( "\n" );
}
func Jf_ManPropagateFlow(p *Jf_Man_t, fEdge int32) {
    var pObj *Gia_Obj_t; 
    var i int32;
    Gia_ManForEachObj( p.pGia, pObj, i )
        if ( Gia_ObjIsBuf(pObj) )
            Jf_ObjPropagateBuf( p, pObj, 0 );
        else if ( Gia_ObjIsAnd(pObj) && Jf_ObjIsUnit(pObj) )
            Jf_ObjComputeBestCut( p, pObj, fEdge, 0 );
    Jf_ManComputeRefs( p );
}
func Jf_ManPropagateEla(p *Jf_Man_t, fEdge int32) {
    var pObj *Gia_Obj_t; 
    var i, CostBef, CostAft int32;
    p.pPars.Area = p.pPars.Edge = p.pPars.Clause = 0;
    Gia_ManForEachObjReverse( p.pGia, pObj, i )
        if ( Gia_ObjIsBuf(pObj) )
            Jf_ObjPropagateBuf( p, pObj, 1 );
        else if Gia_ObjIsAnd(pObj) && Gia_ObjRefNum(p.pGia, pObj) > 0 {
            assert((Jf_ObjIsUnit(pObj) != 0), "Jf_ObjIsUnit(pObj)");
            if Jf_CutCheckMffc(p, Jf_ObjCutBest(p, i), 50) != 0 {
                CostBef = Jf_CutDeref_rec( p, Jf_ObjCutBest(p, i) );
                Jf_ObjComputeBestCut( p, pObj, fEdge, 1 );
                CostAft = Jf_CutRef_rec( p, Jf_ObjCutBest(p, i) );
    //            if ( CostBef != CostAft )  printf( "%d -> %d   ", CostBef, CostAft );
                assert(CostBef >= CostAft, "CostBef >= CostAft"); // does not hold because of JF_EDGE_LIM
            }
            if p.pPars.fGenCnf {
            p.pPars.Clause += Jf_CutCnfSize(p, Jf_ObjCutBest(p, i)); }
            p.pPars.Edge += Jf_CutSize(Jf_ObjCutBest(p, i));
            p.pPars.Area++;
        }
    p.pPars.Delay = Jf_ManComputeDelay( p, 1 );
//    printf( "\n" );
}

/**Function*************************************************************

  Synopsis    [Derives the result of mapping.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Jf_ManDeriveMappingGia(p *Jf_Man_t) *Gia_Man_t {
    var pNew *Gia_Man_t;
    var pObj *Gia_Obj_t; 
    var vCopies *Vec_Int_t    = Vec_IntStartFull( Gia_ManObjNum(p.pGia) );
    var vMapping *Vec_Int_t   = Vec_IntStart( 2 * Gia_ManObjNum(p.pGia) + int32(p.pPars.Edge) + 2 * int32(p.pPars.Area) );
    var vMapping2 *Vec_Int_t  = Vec_IntStart( int32(p.pPars.Edge) + 2 * int32(p.pPars.Area) + 1000 );
    var vCover *Vec_Int_t     = Vec_IntAlloc( 1 << 16 );
    var vLeaves *Vec_Int_t    = Vec_IntAlloc( 16 );
    var vLits *Vec_Int_t = nil;
    var vClas *Vec_Int_t = nil;
    var i int32;
    var k int32;
    var iLit int32;
    var Class int32;
    var pCut *int32;
    var uTruth word;
    var pTruth *word = &uTruth;
    assert(p.pPars.fCutMin, "p.pPars.fCutMin");
    if p.pPars.fGenCnf {
        vLits = Vec_IntAlloc( 1000 );
        vClas = Vec_IntAlloc( 1000 );
        Vec_IntPush( vClas, Vec_IntSize(vLits) );
        Vec_IntPush( vLits, 1 );
    }
    // create new manager
    pNew = Gia_ManStart( Gia_ManObjNum(p.pGia) );
    pNew.pName = Abc_UtilStrsav( p.pGia.pName );
    pNew.pSpec = Abc_UtilStrsav( p.pGia.pSpec );
    // map primary inputs
    Vec_IntWriteEntry( vCopies, 0, 0 );
    Gia_ManForEachCi( p.pGia, pObj, i )
        Vec_IntWriteEntry( vCopies, Gia_ObjId(p.pGia, pObj), Gia_ManAppendCi(pNew) );
    // iterate through nodes used in the mapping
    Gia_ManForEachAnd( p.pGia, pObj, i )
    {
       if Gia_ObjIsBuf(pObj) || Gia_ObjRefNum(p.pGia, pObj) == 0 {
            continue; }
        pCut = Jf_ObjCutBest( p, i );
//        printf( "Best cut of node %d:  ", i );  Jf_CutPrint(pCut);
        Class = Jf_CutFuncClass( pCut );
        if Jf_CutSize(pCut) == 0 {
            assert(Class == 0, "Class == 0");
            Vec_IntWriteEntry( vCopies, i, Jf_CutFunc(pCut) );
            continue;
        }
        if Jf_CutSize(pCut) == 1 {
            assert(Class == 1, "Class == 1");
            iLit = Abc_LitNotCond( Jf_CutLit(pCut, 1) , Jf_CutFuncCompl(pCut) );
            iLit = Abc_Lit2LitL( Vec_IntArray(vCopies), iLit );
            Vec_IntWriteEntry( vCopies, i, iLit );
            continue;
        }
        if p.pPars.fFuncDsd {
            uTruth = Sdm_ManReadDsdTruth(p.pDsd, Class);
        } else {
            pTruth = Vec_MemReadEntry(p.vTtMem, Class); }
        assert(p.pDsd == nil || Sdm_ManReadDsdVarNum(p.pDsd, Class) == Jf_CutSize(pCut), "p.pDsd == nil || Sdm_ManReadDsdVarNum(p.pDsd, Class) == Jf_CutSize(pCut)");
        // collect leaves
        Vec_IntClear( vLeaves );
        // inlining of macro Jf_CutForEachLit
        for  k = 1; k <= Jf_CutSize(pCut) && (iLit = Jf_CutLit(pCut, k)); k++  {
            Vec_IntPush( vLeaves, Abc_Lit2LitL(Vec_IntArray(vCopies), iLit) );
        }
        // create GIA
        iLit = Kit_TruthToGia( pNew, (unsigned *)pTruth, Vec_IntSize(vLeaves), vCover, vLeaves, 0 );
        if p.pPars.fGenCnf {
            Jf_ManGenCnf( uTruth, iLit, vLeaves, vLits, vClas, vCover ); }
        iLit = Abc_LitNotCond( iLit, Jf_CutFuncCompl(pCut) );
        Vec_IntWriteEntry( vCopies, i, iLit );
        // create mapping
        Vec_IntSetEntry( vMapping, Abc_Lit2Var(iLit), Vec_IntSize(vMapping2) );
        Vec_IntPush( vMapping2, Vec_IntSize(vLeaves) );
        Vec_IntForEachEntry( vLeaves, iLit, k )
            Vec_IntPush( vMapping2, Abc_Lit2Var(iLit) );
        Vec_IntPush( vMapping2, Abc_Lit2Var(Vec_IntEntry(vCopies, i)) );
    }
    Gia_ManForEachCo( p.pGia, pObj, i )
    {
        if p.pPars.fGenCnf {
            Vec_IntClear( vLeaves ); }
        iLit = Vec_IntEntry( vCopies, Gia_ObjFaninId0p(p.pGia, pObj) );
        if p.pPars.fGenCnf {
            Vec_IntPush( vLeaves, Abc_LitNotCond(iLit, Gia_ObjFaninC0(pObj)) ); }
        iLit = Gia_ManAppendCo( pNew, Abc_LitNotCond(iLit, Gia_ObjFaninC0(pObj)) );
        if p.pPars.fGenCnf {
            Jf_ManGenCnf( ABC_CONST(0xAAAAAAAAAAAAAAAA), iLit, vLeaves, vLits, vClas, vCover ); }
    }
    Vec_IntFree( vCopies );
    Vec_IntFree( vCover );
    Vec_IntFree( vLeaves );
    // finish mapping 
    if Vec_IntSize(vMapping) > Gia_ManObjNum(pNew) {
        Vec_IntShrink( vMapping, Gia_ManObjNum(pNew) );
    } else {
        Vec_IntFillExtra( vMapping, Gia_ManObjNum(pNew), 0 ); }
    assert(Vec_IntSize(vMapping) == Gia_ManObjNum(pNew), "Vec_IntSize(vMapping) == Gia_ManObjNum(pNew)");
    Vec_IntForEachEntry( vMapping, iLit, i )
        if ( iLit > 0 )
            Vec_IntAddToEntry( vMapping, i, Gia_ManObjNum(pNew) );
    Vec_IntAppend( vMapping, vMapping2 );
    Vec_IntFree( vMapping2 );
    // attach mapping and packing
    assert(pNew.vMapping == nil, "pNew.vMapping == nil");
    pNew.vMapping = vMapping;
    Gia_ManSetRegNum( pNew, Gia_ManRegNum(p.pGia) );
    // derive CNF
    if p.pPars.fGenCnf {
        if p.pPars.fCnfObjIds {
            pNew.pData = Jf_ManCreateCnf( pNew, vLits, vClas );
        } else {
            pNew.pData = Jf_ManCreateCnfRemap( pNew, vLits, vClas, p.pPars.fAddOrCla ); }
    }
    Vec_IntFreeP( &vLits );
    Vec_IntFreeP( &vClas );
    return pNew;
}
func Jf_ManDeriveMapping(p *Jf_Man_t) {
    var vMapping *Vec_Int_t;
    var pObj *Gia_Obj_t; 
    var i int32;
    var k int32;
    var pCut *int32;
    assert(!p.pPars.fCutMin, "!p.pPars.fCutMin");
    vMapping = Vec_IntAlloc( Gia_ManObjNum(p.pGia) + int32(p.pPars.Edge) + int32(p.pPars.Area) * 2 );
    Vec_IntFill( vMapping, Gia_ManObjNum(p.pGia), 0 );
    Gia_ManForEachAnd( p.pGia, pObj, i )
    {
        if Gia_ObjIsBuf(pObj) || Gia_ObjRefNum(p.pGia, pObj) == 0 {
            continue; }
        pCut = Jf_ObjCutBest( p, i );
        Vec_IntWriteEntry( vMapping, i, Vec_IntSize(vMapping) );
        assert(!p.pPars.fCutMin || Jf_CutSize(pCut) <= 6, "!p.pPars.fCutMin || Jf_CutSize(pCut) <= 6");
        Vec_IntPush( vMapping, Jf_CutSize(pCut) );
        for  k = 1; k <= Jf_CutSize(pCut); k++  {
            Vec_IntPush( vMapping, Jf_CutVar(pCut, k) );
        }
        Vec_IntPush( vMapping, i );
    }
    assert(Vec_IntCap(vMapping) == 16 || Vec_IntSize(vMapping) == Vec_IntCap(vMapping), "Vec_IntCap(vMapping) == 16 || Vec_IntSize(vMapping) == Vec_IntCap(vMapping)");
    p.pGia.vMapping = vMapping;
//    Gia_ManMappingVerify( p->pGia );
}

/**Function*************************************************************

  Synopsis    [Derive GIA without mapping.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Jf_ManDeriveGia(p *Jf_Man_t) *Gia_Man_t {
    var pNew, pTemp *Gia_Man_t;
    var pObj *Gia_Obj_t; 
    var vCopies *Vec_Int_t    = Vec_IntStartFull( Gia_ManObjNum(p.pGia) );
    var vCover *Vec_Int_t     = Vec_IntAlloc( 1 << 16 );
    var vLeaves *Vec_Int_t    = Vec_IntAlloc( 16 );
    var i int32;
    var k int32;
    var iLit int32;
    var Class int32;
    var pCut *int32;
    var nWords int32  = Abc_Truth6WordNum(p.pPars.nLutSize);
    var uTruth word;
    var pTruth *word = &uTruth;
    var Truth [JF_WORD_MAX]word;
    // create new manager
    pNew = Gia_ManStart( Gia_ManObjNum(p.pGia) );
    pNew.pName = Abc_UtilStrsav( p.pGia.pName );
    pNew.pSpec = Abc_UtilStrsav( p.pGia.pSpec );
    pNew.vLevels = Vec_IntStart( 6*Gia_ManObjNum(p.pGia)/5 + 100 );
    // map primary inputs
    Vec_IntWriteEntry( vCopies, 0, 0 );
    Gia_ManForEachCi( p.pGia, pObj, i )
        Vec_IntWriteEntry( vCopies, Gia_ObjId(p.pGia, pObj), Gia_ManAppendCi(pNew) );
    // iterate through nodes used in the mapping
    if !p.pPars.fCutMin {
        Gia_ObjComputeTruthTableStart( p.pGia, p.pPars.nLutSize ); }
    Gia_ManHashStart( pNew );
    Gia_ManForEachAnd( p.pGia, pObj, i )
    {
       if Gia_ObjIsBuf(pObj) || Gia_ObjRefNum(p.pGia, pObj) == 0 {
            continue; }
        pCut = Jf_ObjCutBest( p, i );
//        printf( "Best cut of node %d:  ", i );  Jf_CutPrint(pCut);
        // get the truth table
        if p.pPars.fCutMin {
            Class = Jf_CutFuncClass( pCut );
            if Jf_CutSize(pCut) == 0 {
                assert(Class == 0, "Class == 0");
                Vec_IntWriteEntry( vCopies, i, Jf_CutFunc(pCut) );
                continue;
            }
            if Jf_CutSize(pCut) == 1 {
                assert(Class == 1, "Class == 1");
                iLit = Abc_LitNotCond( Jf_CutLit(pCut, 1) , Jf_CutFuncCompl(pCut) );
                iLit = Abc_Lit2LitL( Vec_IntArray(vCopies), iLit );
                Vec_IntWriteEntry( vCopies, i, iLit );
                continue;
            }
            if p.pPars.fFuncDsd {
                uTruth = Sdm_ManReadDsdTruth(p.pDsd, Class);
            } else {
                Abc_TtCopy( (pTruth = Truth), Vec_MemReadEntry(p.vTtMem, Class), nWords, 0 ); }
            assert(p.pDsd == nil || Sdm_ManReadDsdVarNum(p.pDsd, Class) == Jf_CutSize(pCut), "p.pDsd == nil || Sdm_ManReadDsdVarNum(p.pDsd, Class) == Jf_CutSize(pCut)");
        
        } else {
            Vec_IntClear( vLeaves );
            // inlining of macro Jf_CutForEachLit
            for  k = 1; k <= Jf_CutSize(pCut) && (iLit = Jf_CutLit(pCut, k)); k++  {
                Vec_IntPush( vLeaves, Abc_Lit2Var(iLit) );
            }
            pTruth = Gia_ObjComputeTruthTableCut( p.pGia, pObj, vLeaves );
        }
        // collect incoming literals
        Vec_IntClear( vLeaves );
        // inlining of macro Jf_CutForEachLit
        for  k = 1; k <= Jf_CutSize(pCut) && (iLit = Jf_CutLit(pCut, k)); k++  {
            Vec_IntPush( vLeaves, Abc_Lit2LitL(Vec_IntArray(vCopies), iLit) );
        }
        // create GIA
        iLit = Dsm_ManTruthToGia( pNew, pTruth, vLeaves, vCover );
        iLit = Abc_LitNotCond( iLit, (p.pPars.fCutMin && Jf_CutFuncCompl(pCut)) );
        Vec_IntWriteEntry( vCopies, i, iLit );
    }
    Gia_ManForEachCo( p.pGia, pObj, i )
    {
        iLit = Vec_IntEntry( vCopies, Gia_ObjFaninId0p(p.pGia, pObj) );
        Gia_ManAppendCo( pNew, Abc_LitNotCond(iLit, Gia_ObjFaninC0(pObj)) );
    }
    if !p.pPars.fCutMin {
        Gia_ObjComputeTruthTableStop( p.pGia ); }
    Vec_IntFree( vCopies );
    Vec_IntFree( vLeaves );
    Vec_IntFree( vCover );
    Gia_ManHashStop( pNew );
    Gia_ManSetRegNum( pNew, Gia_ManRegNum(p.pGia) );
//    Dsm_ManReportStats();
    // perform cleanup
    if !p.pPars.fCutMin {
        pNew = Gia_ManCleanup( pTemp = pNew );
        Gia_ManStop( pTemp );
    }
    return pNew;
}

/**Function*************************************************************

  Synopsis    []

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Jf_ManSetDefaultPars(pPars *Jf_Par_t) {
    memset( pPars, 0, sizeof(Jf_Par_t) );
    pPars.nLutSize     =  6;
    pPars.nCutNum      =  8;
    pPars.nRounds      =  1;
    pPars.nVerbLimit   =  5;
    pPars.DelayTarget  = -1;
    pPars.fAreaOnly    =  1;
    pPars.fOptEdge     =  1; 
    pPars.fCoarsen     =  0;
    pPars.fCutMin      =  0;
    pPars.fFuncDsd     =  0;
    pPars.fGenCnf      =  0;
    pPars.fPureAig     =  0;
    pPars.fVerbose     =  0;
    pPars.fVeryVerbose =  0;
    pPars.nLutSizeMax  =  JF_LEAF_MAX;
    pPars.nCutNumMax   =  JF_CUT_MAX;
}
func Jf_ManPrintStats(p *Jf_Man_t, pTitle string) {
    if !p.pPars.fVerbose {
        return; }
    fmt.Printf("%s :  ", pTitle);
    fmt.Printf("Level =%6d   ", int64(p.pPars.Delay));
    fmt.Printf("Area =%9d   ", int64(p.pPars.Area));
    fmt.Printf("Edge =%9d   ", int64(p.pPars.Edge));
    if p.pPars.fGenCnf {
    fmt.Printf("Cnf =%9d   ", int64(p.pPars.Clause)); }
    Abc_PrintTime( 1, "Time", Abc_Clock() - p.clkStart );
    fflush( os.Stdout );
}
func Jf_ManPerformMapping(pGia *Gia_Man_t, pPars *Jf_Par_t) *Gia_Man_t {
    var pNew *Gia_Man_t  = pGia;
    var p *Jf_Man_t; var i int32;
    assert(!Gia_ManBufNum(pGia), "!Gia_ManBufNum(pGia)");
    assert(!pPars.fCutMin || !pPars.fFuncDsd || pPars.nLutSize <= 6, "!pPars.fCutMin || !pPars.fFuncDsd || pPars.nLutSize <= 6");
    if pPars.fGenCnf {
        pPars.fCutMin = 1, pPars.fFuncDsd = 1, pPars.fOptEdge = 0; }
    if pPars.fCutMin && !pPars.fFuncDsd {
        pPars.fCoarsen = 0; }
    p = Jf_ManAlloc( pGia, pPars );
    if pPars.fAreaOnly {
        p.pCutCmp = Jf_CutCompareArea;
    } else {
        p.pCutCmp = Jf_CutCompareDelay;
    }
    Jf_ManComputeCuts( p, 0 );
    Jf_ManComputeRefs( p );                         Jf_ManPrintStats( p, "Start" );
    for  i = 0; i < pPars.nRounds; i++  {
        if !p.pPars.fGenCnf {
        Jf_ManPropagateFlow( p, pPars.fOptEdge );  Jf_ManPrintStats( p, "Flow " );
        }
        Jf_ManPropagateEla( p, 0 );                 Jf_ManPrintStats( p, "Area " );
        Jf_ManPropagateEla( p, 1 );                 Jf_ManPrintStats( p, "Edge " );
    }
    if p.pPars.fVeryVerbose && p.pPars.fCutMin && !p.pPars.fFuncDsd {
        Vec_MemDumpTruthTables( p.vTtMem, Gia_ManName(p.pGia), p.pPars.nLutSize ); }
    if p.pPars.fPureAig {
        pNew = Jf_ManDeriveGia(p);
    } else if p.pPars.fCutMin {
        pNew = Jf_ManDeriveMappingGia(p);
    } else {
        Jf_ManDeriveMapping(p); }
    Jf_ManFree( p );
    return pNew;
}
func Jf_ManDeriveCnf(p *Gia_Man_t, fCnfObjIds int32) *Gia_Man_t {
    var Pars Jf_Par_t;
    var pPars *Jf_Par_t = &Pars;
    Jf_ManSetDefaultPars( pPars );
    pPars.fGenCnf = 1;
    pPars.fCnfObjIds = fCnfObjIds;
    return Jf_ManPerformMapping( p, pPars );
}
func Jf_ManDeriveCnfMiter(p *Gia_Man_t, fVerbose int32) *Gia_Man_t {
    var Pars Jf_Par_t;
    var pPars *Jf_Par_t = &Pars;
    Jf_ManSetDefaultPars( pPars );
    pPars.fGenCnf = 1;
    pPars.fCnfObjIds = 0;
    pPars.fAddOrCla = 1;
    pPars.fVerbose = fVerbose;
    return Jf_ManPerformMapping( p, pPars );
}
func Jf_ManDumpCnf(p *Gia_Man_t, pFileName string, fVerbose int32) {
    var clk abctime  = Abc_Clock();
    var pNew *Gia_Man_t;
    var pCnf *Cnf_Dat_t;
    pNew = Jf_ManDeriveCnfMiter( p, fVerbose );
    pCnf = (Cnf_Dat_t *)pNew.pData; pNew.pData = nil;
    Cnf_DataWriteIntoFile( pCnf, pFileName, 0, nil, nil );
    Gia_ManStop( pNew );
//    if ( fVerbose )
    {
        fmt.Printf("CNF stats: Vars = %6d. Clauses = %7d. Literals = %8d. ", pCnf.nVars, pCnf.nClauses, pCnf.nLiterals);
        Abc_PrintTime( 1, "Time", Abc_Clock() - clk );
    }
    Cnf_DataFree(pCnf);
}

func Jf_ManTestCnf(p *Gia_Man_t) {
    var pNew *Gia_Man_t;
    var pCnf *Cnf_Dat_t;
    var i int32;
//    Cnf_Dat_t * pCnf = Cnf_DeriveGia( p );
    pNew = Jf_ManDeriveCnf( p, 1 );
    pCnf = (Cnf_Dat_t *)pNew.pData; pNew.pData = nil;
    Cnf_DataWriteIntoFile( pCnf, "test.cnf", 0, nil, nil );
    for  i = 0; i < pCnf.nVars; i++  {
        fmt.Printf("%d : %d %d\n", i, pCnf.pObj2Count[i], pCnf.pObj2Clause[i]);
    }
    Gia_ManStop( pNew );
    Cnf_DataFree(pCnf);
}

////////////////////////////////////////////////////////////////////////
///                       END OF FILE                                ///
////////////////////////////////////////////////////////////////////////


//ABC_NAMESPACE_IMPL_END

