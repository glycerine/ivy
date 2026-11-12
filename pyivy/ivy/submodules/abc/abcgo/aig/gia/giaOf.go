package gia

import (
	"fmt"
	"os"
)

/**CFile****************************************************************

  FileName    [giaOf.c]

  SystemName  [ABC: Logic synthesis and verification system.]

  PackageName [Scalable AIG package.]

  Synopsis    [LUT structure mapper.]

  Author      [Alan Mishchenko]
  
  Affiliation [UC Berkeley]

  Date        [Ver. 1.0. Started - June 20, 2005.]

  Revision    [$Id: giaOf.c,v 1.00 2005/06/20 00:00:00 alanmi Exp $]

***********************************************************************/

//#include "gia.h"
//#include "misc/st/st.h"
//#include "map/mio/mio.h"
//#include "misc/util/utilTruth.h"
//#include "misc/extra/extra.h"
//#include "base/main/main.h"
//#include "misc/vec/vecMem.h"
//#include "misc/vec/vecWec.h"
//#include "opt/dau/dau.h"
//#include "sat/bsat/satStore.h"

//ABC_NAMESPACE_IMPL_START

////////////////////////////////////////////////////////////////////////
///                        DECLARATIONS                              ///
////////////////////////////////////////////////////////////////////////

const OF_LEAF_MAX =  6
const OF_CUT_MAX =  32
const OF_NO_LEAF =  31
const OF_NO_FUNC =  0x7FFFFFF
const OF_CUT_EXTRA = 4 // size; delay1, delay2; area

    
type Of_Cut_t struct {
    Sign            word;           // signature
    Delay             int32;          // delay
    Flow             int32;           // flow
    iFunc uint32; /* C bitfield: 27 bits; review packing and truncation */   // function (OF_NO_FUNC)
    nLeaves uint32; /* C bitfield: 5 bits; review packing and truncation */   // leaf number (OF_NO_LEAF)
    pLeaves [OF_LEAF_MAX+1]int32; // leaves
}
    
type Of_Obj_t struct {
    iCutH             int32;          // best cut
    iCutH2             int32;         // best cut
    Delay1             int32;         // arrival time
    Delay2             int32;         // arrival time
    Required             int32;       // required
    nRefs             int32;          // references 
    Flow             int32;           // area flow
    Temp             int32;           // unused
}
    
type Of_Man_t struct {
    // user data
    pGia      *Gia_Man_t;           // derived manager
    pPars       *Jf_Par_t;          // parameters
    // cut data
    vTtMem      *Vec_Mem_t;         // truth tables
    vPages       Vec_Ptr_t;         // cut memory
    vCutSets       Vec_Int_t;       // cut offsets
    vCutFlows       Vec_Int_t;      // temporary cut area
    vCutDelays       Vec_Int_t;     // temporary cut delay
    vCutRefs       Vec_Int_t;       // temporary cut referebces
    iCur             int32;           // current position
    Iter             int32;           // mapping iterations
    // object data
    pObjs       *Of_Obj_t;          
    // statistics
    clkStart         abctime;       // starting time
    CutCount [6]float64;    // cut counts
}

const OF_NUM =       10
const OF_NUMINV =   0.1

func Of_Flt2Int(f float32) int32 { return int32((OF_NUM*f));                                          }
func Of_Int2Flt(i int32) float32 { return OF_NUMINV*i;                                              }

func Of_ManCutSet(p *Of_Man_t, i int32) *int32 { return (int *)Vec_PtrEntry(&p.vPages, i >> 16) + (i & 0xFFFF);  }
func Of_ObjCutSetId(p *Of_Man_t, i int32) int32 { return Vec_IntEntry( &p.vCutSets, i );                          }
func Of_ObjCutSet(p *Of_Man_t, i int32) *int32 { return Of_ManCutSet(p, Of_ObjCutSetId(p, i));                    }
func Of_ObjHasCuts(p *Of_Man_t, i int32) int32 { return int32((Vec_IntEntry(&p.vCutSets, i) > 0));                 }

func Of_ObjCutFlow(p *Of_Man_t, i int32) int32 { return Vec_IntEntry(&p.vCutFlows, i);                           } 
func Of_ObjCutDelay(p *Of_Man_t, i int32) int32 { return Vec_IntEntry(&p.vCutDelays, i);                          } 
func Of_ObjSetCutFlow(p *Of_Man_t, i int32, a int32) { Vec_IntWriteEntry(&p.vCutFlows, i, a);                          } 
func Of_ObjSetCutDelay(p *Of_Man_t, i int32, d int32) { Vec_IntWriteEntry(&p.vCutDelays, i, d);                         } 

func Of_CutSize(pCut *int32) int32 { return pCut[0] & OF_NO_LEAF;                                     }
func Of_CutFunc(pCut *int32) int32 { return (uint32(pCut[0]) >> 5);                                 }
func Of_CutLeaves(pCut *int32) *int32 { return pCut + 1;                                                 }
func Of_CutSetBoth(n int32, f int32) int32 { return n | (f << 5);                                             }
func Of_CutHandle(pCutSet *int32, pCut *int32) int32 { assert(pCut > pCutSet, "pCut > pCutSet"); return pCut - pCutSet;                 } 
func Of_CutFromHandle(pCutSet *int32, h int32) *int32 { assert(h > 0, "h > 0"); return pCutSet + h;                             }

func Of_CutDelay1(pCut *int32) int32 { return pCut[1 + Of_CutSize(pCut)];                               }
func Of_CutDelay2(pCut *int32) int32 { return pCut[2 + Of_CutSize(pCut)];                               }
func Of_CutAreaFlow(pCut *int32) int32 { return pCut[3 + Of_CutSize(pCut)];                               }
func Of_CutSetDelay1(pCut *int32, d int32) { pCut[1 + Of_CutSize(pCut)] = d;                                  } 
func Of_CutSetDelay2(pCut *int32, d int32) { pCut[2 + Of_CutSize(pCut)] = d;                                  } 
func Of_CutSetAreaFlow(pCut *int32, d int32) { pCut[3 + Of_CutSize(pCut)] = d;                                  } 

func Of_CutVar(pCut *int32, v int32) int32 { return Abc_Lit2Var(Of_CutLeaves(pCut)[v]);                       } 
func Of_CutFlag(pCut *int32, v int32) int32 { return Abc_LitIsCompl(Of_CutLeaves(pCut)[v]);                    } 
func Of_CutCleanFlag(pCut *int32, v int32) { Of_CutLeaves(pCut)[v] = Abc_LitRegular(Of_CutLeaves(pCut)[v]);   } 
func Of_CutSetFlag(pCut *int32, v int32, x int32) { Of_CutLeaves(pCut)[v] = Abc_Var2Lit(Of_CutVar(pCut, v), x);      } 

func Of_ObjData(p *Of_Man_t, i int32) *Of_Obj_t { return p.pObjs + i;                                             }

func Of_ObjCutBest(p *Of_Man_t, i int32) int32 { return Of_ObjData(p, i).iCutH;                                  }
func Of_ObjCutBest2(p *Of_Man_t, i int32) int32 { return Of_ObjData(p, i).iCutH2;                                 }
func Of_ObjDelay1(p *Of_Man_t, i int32) int32 { return Of_ObjData(p, i).Delay1;                                 }
func Of_ObjDelay2(p *Of_Man_t, i int32) int32 { return Of_ObjData(p, i).Delay2;                                 }
func Of_ObjRequired(p *Of_Man_t, i int32) int32 { return Of_ObjData(p, i).Required;                               }
func Of_ObjRefNum(p *Of_Man_t, i int32) int32 { return Of_ObjData(p, i).nRefs;                                  }
func Of_ObjFlow(p *Of_Man_t, i int32) int32 { return Of_ObjData(p, i).Flow;                                   }

func Of_ObjSetCutBest(p *Of_Man_t, i int32, x int32) { Of_ObjData(p, i).iCutH = x;                                     }
func Of_ObjSetCutBest2(p *Of_Man_t, i int32, x int32) { Of_ObjData(p, i).iCutH2 = x;                                    }
func Of_ObjSetDelay1(p *Of_Man_t, i int32, x int32) { Of_ObjData(p, i).Delay1 = x;                                    }
func Of_ObjSetDelay2(p *Of_Man_t, i int32, x int32) { Of_ObjData(p, i).Delay2 = x;                                    }
func Of_ObjSetRequired(p *Of_Man_t, i int32, x int32) { Of_ObjData(p, i).Required = x;                                  }
func Of_ObjSetRefNum(p *Of_Man_t, i int32, x int32) { Of_ObjData(p, i).nRefs = x;                                     }
func Of_ObjSetFlow(p *Of_Man_t, i int32, x int32) { Of_ObjData(p, i).Flow = x;                                      }
func Of_ObjUpdateRequired(p *Of_Man_t, i int32, x int32) { if Of_ObjRequired(p, i) > x { Of_ObjSetRequired(p, i, x); }      }
func Of_ObjRefInc(p *Of_Man_t, i int32) int32 { return Of_ObjData(p, i).nRefs++;                                }
func Of_ObjRefDec(p *Of_Man_t, i int32) int32 { return --Of_ObjData(p, i).nRefs;                                }

func Of_ObjCutBestP(p *Of_Man_t, iObj int32) *int32 { assert(iObj>0 && iObj<Gia_ManObjNum(p.pGia), "iObj>0 && iObj<Gia_ManObjNum(p.pGia)");return Of_ManCutSet( p, Of_ObjCutBest(p, iObj) );  }
func Of_ObjSetCutBestP(p *Of_Man_t, pCutSet *int32, iObj int32, pCut *int32) { Of_ObjSetCutBest( p, iObj, Of_ObjCutSetId(p, iObj) + Of_CutHandle(pCutSet, pCut) ); }

func Of_ObjCutBestP2(p *Of_Man_t, iObj int32) *int32 { assert(iObj>0 && iObj<Gia_ManObjNum(p.pGia), "iObj>0 && iObj<Gia_ManObjNum(p.pGia)");return Of_ManCutSet( p, Of_ObjCutBest2(p, iObj) );  }
func Of_ObjSetCutBestP2(p *Of_Man_t, pCutSet *int32, iObj int32, pCut *int32) { Of_ObjSetCutBest2( p, iObj, Of_ObjCutSetId(p, iObj) + Of_CutHandle(pCutSet, pCut) ); }

// Original macro:
// #define Of_SetForEachCut( pList, pCut, i )          for ( i = 0, pCut = pList + 1; i < pList[0]; i++, pCut += Of_CutSize(pCut) + OF_CUT_EXTRA )

// Original macro:
// #define Of_ObjForEachCut( pCuts, i, nCuts )         for ( i = 0, i < nCuts; i++ )

// Original macro:
// #define Of_CutForEachVar( pCut, iVar, i )           for ( i = 0; i < Of_CutSize(pCut) && (iVar = Of_CutVar(pCut,i)); i++ )

// Original macro:
// #define Of_CutForEachVarFlag( pCut, iVar, Flag, i ) for ( i = 0; i < Of_CutSize(pCut) && (iVar = Of_CutVar(pCut,i)) && ((Flag = Of_CutFlag(pCut,i)), 1); i++ )


////////////////////////////////////////////////////////////////////////
///                     FUNCTION DEFINITIONS                         ///
////////////////////////////////////////////////////////////////////////

/**Function*************************************************************

  Synopsis    [Area flow.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Of_ManAreaFlow(p *Of_Man_t) {
    var AreaUnit int32  = 1000;
    var i, Id, Total int32;
    var pObj *Gia_Obj_t;
    assert(p.pGia.pRefs == nil, "p.pGia.pRefs == nil");
    Gia_ManCreateRefs( p.pGia );
    Of_ObjSetFlow( p, 0, 0 );
    // inlining of macro Gia_ManForEachCiId
    for  i = 0; (i < Vec_IntSize(p.pGia.vCis)) && ((Id) = Gia_ObjId(p.pGia, Gia_ManCi(p.pGia, i))); i++  {
        Of_ObjSetFlow( p, Id, 0 );
    }
    // inlining of macro Gia_ManForEachAnd
    for  Id = 0; (Id < p.pGia.nObjs) && ((pObj) = Gia_ManObj(p.pGia, Id)); Id++  {      if !Gia_ObjIsAnd(pObj) {} else {
        Of_ObjSetFlow( p, Id, (Gia_ObjFanin0(pObj).Value + Gia_ObjFanin1(pObj).Value + AreaUnit) / Gia_ObjRefNum(p.pGia, pObj) ); }
    }
    // inlining of macro Gia_ManForEachCo
    for  i = 0; (i < Vec_IntSize(p.pGia.vCos)) && ((pObj) = Gia_ManCo(p.pGia, i)); i++  {
        Total += Gia_ObjFanin0(pObj).Value;
    }
    // inlining of macro ABC_FREE
    ((p.pGia.pRefs) ? (free((char *) (p.pGia.pRefs)), (p.pGia.pRefs) = 0) : 0);
    if true {
        return; }
    fmt.Printf("CI = %5d.  ", Gia_ManCiNum(p.pGia));
    fmt.Printf("CO = %5d.  ", Gia_ManCoNum(p.pGia));
    fmt.Printf("And = %8d.  ", Gia_ManAndNum(p.pGia));
    fmt.Printf("Area = %8d.  ", Total/AreaUnit);
    fmt.Printf("\n");
}

/**Function*************************************************************

  Synopsis    []

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Of_StoCreate(pGia *Gia_Man_t, pPars *Jf_Par_t) *Of_Man_t {
    extern void Mf_ManSetFlowRefs( Gia_Man_t * p, Vec_Int_t * vRefs );
    var p *Of_Man_t;
    var vFlowRefs *Vec_Int_t;
    var pRefs *int32  = nil;
    assert(pPars.nCutNum > 1  && pPars.nCutNum <= OF_CUT_MAX, "pPars.nCutNum > 1  && pPars.nCutNum <= OF_CUT_MAX");
    assert(pPars.nLutSize > 1 && pPars.nLutSize <= OF_LEAF_MAX, "pPars.nLutSize > 1 && pPars.nLutSize <= OF_LEAF_MAX");
    // inlining of macro ABC_FREE
    ((pGia.pRefs) ? (free((char *) (pGia.pRefs)), (pGia.pRefs) = 0) : 0);
    Vec_IntFreeP( &pGia.vCellMapping );
    if Gia_ManHasChoices(pGia) {
        Gia_ManSetPhase(pGia); }
    // create references
    // inlining of macro ABC_FREE
    ((pGia.pRefs) ? (free((char *) (pGia.pRefs)), (pGia.pRefs) = 0) : 0);
    vFlowRefs = Vec_IntAlloc(0);
    Mf_ManSetFlowRefs( pGia, vFlowRefs );
    pGia.pRefs= Vec_IntReleaseArray(vFlowRefs);
    Vec_IntFree(vFlowRefs);
    // create
    // inlining of macro ABC_CALLOC
    p = &Of_Man_t{};
    p.clkStart = Abc_Clock();
    p.pGia     = pGia;
    p.pPars    = pPars;
    // inlining of macro ABC_CALLOC
    p.pObjs    = ((Of_Obj_t *) calloc((size_t)(Gia_ManObjNum(pGia)), sizeof(Of_Obj_t)));
    p.iCur     = 2;
    // other
    Vec_PtrGrow( &p.vPages, 256 );                                    // cut memory
    Vec_IntFill( &p.vCutSets,  Gia_ManObjNum(pGia), 0 );              // cut offsets
    Vec_IntFill( &p.vCutFlows, Gia_ManObjNum(pGia), 0 );              // cut area
    Vec_IntFill( &p.vCutDelays,Gia_ManObjNum(pGia), 0 );              // cut delay
    Vec_IntGrow( &p.vCutRefs,  1000 );                                // cut references
    if pPars.fCutMin {
        p.vTtMem = Vec_MemAllocForTT( 6, 0 ); }          
    // compute area flow
    pRefs = pGia.pRefs; pGia.pRefs = nil;
    Of_ManAreaFlow( p );
    pGia.pRefs = pRefs;
    return p;
}
func Of_StoDelete(p *Of_Man_t) {
    Vec_PtrFreeData( &p.vPages );
    Vec_PtrErase( &p.vPages );
    Vec_IntErase( &p.vCutSets );
    Vec_IntErase( &p.vCutFlows );
    Vec_IntErase( &p.vCutDelays );
    Vec_IntErase( &p.vCutRefs );
    // inlining of macro ABC_FREE
    ((p.pObjs) ? (free((char *) (p.pObjs)), (p.pObjs) = 0) : 0);
    // matching
    if p.pPars.fCutMin {
        Vec_MemHashFree( p.vTtMem ); }
    if p.pPars.fCutMin {
        Vec_MemFree( p.vTtMem ); }
    // inlining of macro ABC_FREE
    ((p) ? (free((char *) (p)), (p) = 0) : 0);
}




/**Function*************************************************************

  Synopsis    []

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Of_CutComputeTruth6(p *Of_Man_t, pCut0 *Of_Cut_t, pCut1 *Of_Cut_t, fCompl0 int32, fCompl1 int32, pCutR *Of_Cut_t, fIsXor int32) int32 {
//    extern int Of_ManTruthCanonicize( word * t, int nVars );
    var nOldSupp int32 = pCutR.nLeaves;
    var truthId int32;
    var fCompl int32; var t word;
    var t0 word  = *Vec_MemReadEntry(p.vTtMem, Abc_Lit2Var(pCut0.iFunc));
    var t1 word  = *Vec_MemReadEntry(p.vTtMem, Abc_Lit2Var(pCut1.iFunc));
    if Abc_LitIsCompl(pCut0.iFunc) ^ fCompl0 { t0 = ~t0; }
    if Abc_LitIsCompl(pCut1.iFunc) ^ fCompl1 { t1 = ~t1; }
    t0 = Abc_Tt6Expand( t0, pCut0.pLeaves, pCut0.nLeaves, pCutR.pLeaves, pCutR.nLeaves );
    t1 = Abc_Tt6Expand( t1, pCut1.pLeaves, pCut1.nLeaves, pCutR.pLeaves, pCutR.nLeaves );
    if fIsXor != 0 {
        t = t0 ^ t1;
    } else {
        t = t0 & t1;
    }
    if fCompl = int32((t & 1)) { t = ~t; }
    pCutR.nLeaves = Abc_Tt6MinBase( &t, pCutR.pLeaves, pCutR.nLeaves );
    assert(int32((t & 1)) == 0, "int32((t & 1)) == 0");
    truthId        = Vec_MemHashInsert(p.vTtMem, &t);
    pCutR.iFunc   = Abc_Var2Lit( truthId, fCompl );
    assert(int32(pCutR.nLeaves) <= nOldSupp, "int32(pCutR.nLeaves) <= nOldSupp");
    return int32(pCutR.nLeaves) < nOldSupp;
}
func Of_CutComputeTruthMux6(p *Of_Man_t, pCut0 *Of_Cut_t, pCut1 *Of_Cut_t, pCutC *Of_Cut_t, fCompl0 int32, fCompl1 int32, fComplC int32, pCutR *Of_Cut_t) int32 {
    var nOldSupp int32 = pCutR.nLeaves;
    var truthId int32;
    var fCompl int32; var t word;
    var t0 word  = *Vec_MemReadEntry(p.vTtMem, Abc_Lit2Var(pCut0.iFunc));
    var t1 word  = *Vec_MemReadEntry(p.vTtMem, Abc_Lit2Var(pCut1.iFunc));
    var tC word  = *Vec_MemReadEntry(p.vTtMem, Abc_Lit2Var(pCutC.iFunc));
    if Abc_LitIsCompl(pCut0.iFunc) ^ fCompl0 { t0 = ~t0; }
    if Abc_LitIsCompl(pCut1.iFunc) ^ fCompl1 { t1 = ~t1; }
    if Abc_LitIsCompl(pCutC.iFunc) ^ fComplC { tC = ~tC; }
    t0 = Abc_Tt6Expand( t0, pCut0.pLeaves, pCut0.nLeaves, pCutR.pLeaves, pCutR.nLeaves );
    t1 = Abc_Tt6Expand( t1, pCut1.pLeaves, pCut1.nLeaves, pCutR.pLeaves, pCutR.nLeaves );
    tC = Abc_Tt6Expand( tC, pCutC.pLeaves, pCutC.nLeaves, pCutR.pLeaves, pCutR.nLeaves );
    t = (tC & t1) | (~tC & t0);
    if fCompl = int32((t & 1)) { t = ~t; }
    pCutR.nLeaves = Abc_Tt6MinBase( &t, pCutR.pLeaves, pCutR.nLeaves );
    assert(int32((t & 1)) == 0, "int32((t & 1)) == 0");
    truthId        = Vec_MemHashInsert(p.vTtMem, &t);
    pCutR.iFunc   = Abc_Var2Lit( truthId, fCompl );
    assert(int32(pCutR.nLeaves) <= nOldSupp, "int32(pCutR.nLeaves) <= nOldSupp");
    return int32(pCutR.nLeaves) < nOldSupp;
}


/**Function*************************************************************

  Synopsis    []

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Of_CutCountBits(i word) int32 {
    i = i - ((i >> 1) & 0x5555555555555555);
    i = (i & 0x3333333333333333) + ((i >> 2) & 0x3333333333333333);
    i = ((i + (i >> 4)) & 0x0F0F0F0F0F0F0F0F);
    return (i*(0x0101010101010101))>>56;
}
func Of_CutGetSign(pLeaves *int32, nLeaves int32) word {
    var Sign word; var i int32; 
    for  i = 0; i < nLeaves; i++  {
        Sign |= ((word)1) << (pLeaves[i] & 0x3F);
    }
    return Sign;
}
func Of_CutCreateUnit(p *Of_Cut_t, i int32) int32 {
    p.Delay      = 0;
    p.Flow       = 0;
    p.iFunc      = 2;
    p.nLeaves    = 1;
    p.pLeaves[0] = i;
    p.Sign       = ((word)1) << (i & 0x3F);
    return 1;
}
func Of_Cutprintf(p *Of_Man_t, pCut *Of_Cut_t) {
    var i int32;
    var nDigits int32 = Abc_Base10Log(Gia_ManObjNum(p.pGia)); 
    fmt.Printf("%d  {", pCut.nLeaves);
    for  i = 0; i < int32(pCut.nLeaves); i++  {
        fmt.Printf(" %*d", nDigits, pCut.pLeaves[i]);
    }
    for  ; i < int32(p.pPars.nLutSize); i++  {
        fmt.Printf(" %*s", nDigits, " ");
    }
    fmt.Printf("  }   D = %4d  A = %9d  F = %6d  ", pCut.Delay, pCut.Flow, pCut.iFunc);
    if p.vTtMem {
        Dau_DsdPrintFromTruth( Vec_MemReadEntry(p.vTtMem, Abc_Lit2Var(pCut.iFunc)), pCut.nLeaves );
    } else {
        fmt.Printf("\n"); }
}
func Of_ManPrepareCuts(pCuts *Of_Cut_t, p *Of_Man_t, iObj int32, fAddUnit int32) int32 {
    if Of_ObjHasCuts(p, iObj) != 0 {
        var pMfCut *Of_Cut_t  = pCuts;
        var i int32;
        var pCut *int32;
        var pList *int32 = Of_ObjCutSet(p, iObj);
        // inlining of macro Of_SetForEachCut
        for  i, pCut = 0, pList + 1; i < pList[0]; i++, pCut += Of_CutSize(pCut) + OF_CUT_EXTRA  {
            pMfCut.Delay   = 0;
            pMfCut.Flow    = 0;
            pMfCut.iFunc   = Of_CutFunc( pCut );
            pMfCut.nLeaves = Of_CutSize( pCut );
            pMfCut.Sign    = Of_CutGetSign( pCut+1, Of_CutSize(pCut) );
            memcpy( pMfCut.pLeaves, pCut+1, sizeof(int) * Of_CutSize(pCut) );
            pMfCut++;
        }
        if (fAddUnit != 0) && pCuts.nLeaves > 1 {
            return pList[0] + Of_CutCreateUnit( pMfCut, iObj ); }
        return pList[0];
    }
    return Of_CutCreateUnit( pCuts, iObj );
}
func Of_ManSaveCuts(p *Of_Man_t, pCuts **Of_Cut_t, nCuts int32) int32 {
    var i int32;
    var pPlace *int32;
    var iCur int32;
    var nInts int32 = 1;
    var nCutsNew int32;
    for  i = 0; i < nCuts; i++  {
        nInts += pCuts[i].nLeaves + OF_CUT_EXTRA, nCutsNew++;
    }
    if (p.iCur & 0xFFFF) + nInts > 0xFFFF {
        p.iCur = ((p.iCur >> 16) + 1) << 16; }
    if Vec_PtrSize(&p.vPages) == (p.iCur >> 16) {
        // inlining of macro ABC_CALLOC
        Vec_PtrPush( &p.vPages, ((int *) calloc((size_t)((1<<16)), sizeof(int))) ); }
    iCur = p.iCur; p.iCur += nInts;
    pPlace = Of_ManCutSet( p, iCur );
    *pPlace++ = nCutsNew;
    for  i = 0; i < nCuts; i++  {
        *pPlace++ = Of_CutSetBoth( pCuts[i].nLeaves, pCuts[i].iFunc );
        memcpy( pPlace, pCuts[i].pLeaves, sizeof(int) * pCuts[i].nLeaves );
        pPlace += pCuts[i].nLeaves;
        memset( pPlace, 0xFF, sizeof(int) * (OF_CUT_EXTRA - 1) );
        pPlace += OF_CUT_EXTRA - 1;
    }
    return iCur;
}
func Of_ManLiftCuts(p *Of_Man_t, iObj int32) {
    var i int32;
    var k int32;
    var pCut *int32;
    var pList *int32 = Of_ObjCutSet(p, iObj);
    assert((Of_ObjHasCuts(p, iObj) != 0), "Of_ObjHasCuts(p, iObj)");
    // inlining of macro Of_SetForEachCut
    for  i, pCut = 0, pList + 1; i < pList[0]; i++, pCut += Of_CutSize(pCut) + OF_CUT_EXTRA  {
        for  k = 1; k <= Of_CutSize(pCut); k++  {
            pCut[k] = Abc_Var2Lit(pCut[k], 0);
        }
    }
}
func Of_CutPrint(pCut *int32) {
    var k, iVar int32;
    fmt.Printf("Cut with %d inputs and function %3d : { ", Of_CutSize(pCut), Of_CutFunc(pCut) == OF_NO_FUNC ? 0 : Of_CutFunc(pCut));
    // inlining of macro Of_CutForEachVar
    for  k = 0; k < Of_CutSize(pCut) && (iVar = Of_CutVar(pCut,k)); k++  {
        fmt.Printf("%d ", iVar);
    }
    fmt.Printf("}\n");
}

/**Function*************************************************************

  Synopsis    [Check correctness of cuts.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
static inline int Of_CutCheck( Of_Cut_t * pBase, Of_Cut_t * pCut ) // check if pCut is contained in pBase
{
    var nSizeB int32  = pBase.nLeaves;
    var nSizeC int32  = pCut.nLeaves;
    var i int32;
    var pB *int32 = pBase.pLeaves;
    var k int32;
    var pC *int32 = pCut.pLeaves;
    for  i = 0; i < nSizeC; i++  {
        for  k = 0; k < nSizeB; k++  {
            if pC[i] == pB[k] {
                break; }
        }
        if k == nSizeB {
            return 0; }
    }
    return 1;
}
func Of_SetCheckArray(ppCuts **Of_Cut_t, nCuts int32) int32 {
    var pCut0, pCut1 *Of_Cut_t; 
    var i, k, m, n, Value int32;
    assert(nCuts > 0, "nCuts > 0");
    for  i = 0; i < nCuts; i++  {
        pCut0 = ppCuts[i];
        assert(pCut0.nLeaves <= OF_LEAF_MAX, "pCut0.nLeaves <= OF_LEAF_MAX");
        assert(pCut0.Sign == Of_CutGetSign(pCut0.pLeaves, pCut0.nLeaves), "pCut0.Sign == Of_CutGetSign(pCut0.pLeaves, pCut0.nLeaves)");
        // check duplicates
        for  m = 0; m < int32(pCut0.nLeaves); m++  {
        for  n = m + 1; n < int32(pCut0.nLeaves); n++  {
            assert(pCut0.pLeaves[m] < pCut0.pLeaves[n], "pCut0.pLeaves[m] < pCut0.pLeaves[n]");
        }
        }
        // check pairs
        for  k = 0; k < nCuts; k++  {
            pCut1 = ppCuts[k];
            if pCut0 == pCut1 {
                continue; }
            // check containments
            Value = Of_CutCheck( pCut0, pCut1 );
            assert(Value == 0, "Value == 0");
        }
    }
    return 1;
}


/**Function*************************************************************

  Synopsis    []

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Of_CutMergeOrder(pCut0 *Of_Cut_t, pCut1 *Of_Cut_t, pCut *Of_Cut_t, nLutSize int32) int32 { 
    var nSize0 int32    = pCut0.nLeaves;
    var nSize1 int32    = pCut1.nLeaves;
    var i int32;
    var pC0 *int32 = pCut0.pLeaves;
    var k int32;
    var pC1 *int32 = pCut1.pLeaves;
    var c int32;
    var pC *int32 = pCut.pLeaves;
    // the case of the largest cut sizes
    if nSize0 == nLutSize && nSize1 == nLutSize {
        for  i = 0; i < nSize0; i++  {
            if pC0[i] != pC1[i] {  return 0; }
            pC[i] = pC0[i];
        }
        pCut.nLeaves = nLutSize;
        pCut.iFunc = OF_NO_FUNC;
        pCut.Sign = pCut0.Sign | pCut1.Sign;
        return 1;
    }
    // compare two cuts with different numbers
    i = k = c = 0;
    if nSize0 == 0 { goto FlushCut1; }
    if nSize1 == 0 { goto FlushCut0; }
    for {
        if c == nLutSize { return 0; }
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
    if c + nSize0 > nLutSize + i { return 0; }
    for  i < nSize0  {
        pC[c++] = pC0[i++];
    }
    pCut.nLeaves = c;
    pCut.iFunc = OF_NO_FUNC;
    pCut.Sign = pCut0.Sign | pCut1.Sign;
    return 1;

FlushCut1:
    if c + nSize1 > nLutSize + k { return 0; }
    for  k < nSize1  {
        pC[c++] = pC1[k++];
    }
    pCut.nLeaves = c;
    pCut.iFunc = OF_NO_FUNC;
    pCut.Sign = pCut0.Sign | pCut1.Sign;
    return 1;
}
func Of_CutMergeOrderMux(pCut0 *Of_Cut_t, pCut1 *Of_Cut_t, pCut2 *Of_Cut_t, pCut *Of_Cut_t, nLutSize int32) int32 { 
    var x0 int32;
    var i0 int32;
    var nSize0 int32 = pCut0.nLeaves;
    var pC0 *int32 = pCut0.pLeaves;
    var x1 int32;
    var i1 int32;
    var nSize1 int32 = pCut1.nLeaves;
    var pC1 *int32 = pCut1.pLeaves;
    var x2 int32;
    var i2 int32;
    var nSize2 int32 = pCut2.nLeaves;
    var pC2 *int32 = pCut2.pLeaves;
    var xMin int32;
    var c int32;
    var pC *int32 = pCut.pLeaves;
    for {
        if i0 == nSize0 {
            x0 = ABC_INFINITY;
        } else {
            x0 = pC0[i0];
        }
        if i1 == nSize1 {
            x1 = ABC_INFINITY;
        } else {
            x1 = pC1[i1];
        }
        if i2 == nSize2 {
            x2 = ABC_INFINITY;
        } else {
            x2 = pC2[i2];
        }
        xMin = Abc_MinInt( Abc_MinInt(x0, x1), x2 );
        if xMin == ABC_INFINITY { break; }
        if c == nLutSize { return 0; }
        pC[c++] = xMin;
        if x0 == xMin { i0++; }
        if x1 == xMin { i1++; }
        if x2 == xMin { i2++; }
    }
    pCut.nLeaves = c;
    pCut.iFunc = OF_NO_FUNC;
    pCut.Sign = pCut0.Sign | pCut1.Sign | pCut2.Sign;
    return 1;
}
static inline int Of_SetCutIsContainedOrder( Of_Cut_t * pBase, Of_Cut_t * pCut ) // check if pCut is contained in pBase
{
    var i int32;
    var nSizeB int32 = pBase.nLeaves;
    var k int32;
    var nSizeC int32 = pCut.nLeaves;
    if nSizeB == nSizeC {
        for  i = 0; i < nSizeB; i++  {
            if pBase.pLeaves[i] != pCut.pLeaves[i] {
                return 0; }
        }
        return 1;
    }
    assert(nSizeB > nSizeC, "nSizeB > nSizeC"); 
    if nSizeC == 0 {
        return 1; }
    for  i = k = 0; i < nSizeB; i++  {
        if pBase.pLeaves[i] > pCut.pLeaves[k] {
            return 0; }
        if pBase.pLeaves[i] == pCut.pLeaves[k] {
            if k++; k == nSizeC {
                return 1; }
        }
    }
    return 0;
}
func Of_SetLastCutIsContained(pCuts **Of_Cut_t, nCuts int32) int32 {
    var i int32;
    for  i = 0; i < nCuts; i++  {
        if pCuts[i].nLeaves <= pCuts[nCuts].nLeaves && (pCuts[i].Sign & pCuts[nCuts].Sign) == pCuts[i].Sign && Of_SetCutIsContainedOrder(pCuts[nCuts], pCuts[i]) {
            return 1; }
    }
    return 0;
}
func Of_SetLastCutContainsArea(pCuts **Of_Cut_t, nCuts int32) int32 {
    var i, k, fChanges int32;
    for  i = 0; i < nCuts; i++  {
        if pCuts[nCuts].nLeaves < pCuts[i].nLeaves && (pCuts[nCuts].Sign & pCuts[i].Sign) == pCuts[nCuts].Sign && Of_SetCutIsContainedOrder(pCuts[i], pCuts[nCuts]) {
            pCuts[i].nLeaves = OF_NO_LEAF, fChanges = 1; }
    }
    if !(fChanges != 0) {
        return nCuts; }
    for  i = k = 0; i <= nCuts; i++  {
        if pCuts[i].nLeaves == OF_NO_LEAF {
            continue; }
        if k < i {
            // inlining of macro ABC_SWAP
             var t *Of_Cut_t  = pCuts[k]; pCuts[k] = pCuts[i]; pCuts[i] = t; };
        k++;
    }
    return k - 1;
}
func Of_CutCompareArea(pCut0 *Of_Cut_t, pCut1 *Of_Cut_t) int32 {
    if pCut0.Delay   < pCut1.Delay {  return -1; }
    if pCut0.Delay   > pCut1.Delay {  return  1; }
    if pCut0.Flow    < pCut1.Flow {  return -1; }
    if pCut0.Flow    > pCut1.Flow {  return  1; }
    if pCut0.nLeaves < pCut1.nLeaves {  return -1; }
    if pCut0.nLeaves > pCut1.nLeaves {  return  1; }
    return 0;
}
func Of_SetSortByArea(pCuts **Of_Cut_t, nCuts int32) {
    var i int32;
    for  i = nCuts; i > 0; i--  {
        if Of_CutCompareArea(pCuts[i - 1], pCuts[i]) < 0 {//!= 1 )
            return; }
        // inlining of macro ABC_SWAP
        { var t *Of_Cut_t  = pCuts[i - 1]; pCuts[i - 1] = pCuts[i]; pCuts[i] = t; };
    }
}
func Of_SetAddCut(pCuts **Of_Cut_t, nCuts int32, nCutNum int32) int32 {
    if nCuts == 0 {
        return 1; }
    nCuts = Of_SetLastCutContainsArea(pCuts, nCuts);
    Of_SetSortByArea( pCuts, nCuts );
    return Abc_MinInt( nCuts + 1, nCutNum - 1 );
}
func Of_CutArea(p *Of_Man_t, nLeaves int32) int32 {
    if nLeaves < 2 {
        return 0; }
    return nLeaves + p.pPars.nAreaTuner;
}
func Of_CutParams(p *Of_Man_t, pCut *Of_Cut_t, nGiaRefs int32) {
    var i int32;
    var nLeaves int32 = pCut.nLeaves; 
    assert(nLeaves <= p.pPars.nLutSize, "nLeaves <= p.pPars.nLutSize");
    pCut.Delay = 0;
    pCut.Flow  = 0;
    for  i = 0; i < nLeaves; i++  {
        pCut.Delay = Abc_MaxInt( pCut.Delay, Of_ObjCutDelay(p, pCut.pLeaves[i]) );
        pCut.Flow += Of_ObjCutFlow(p, pCut.pLeaves[i]);
    }
    pCut.Delay += int32((nLeaves > 1));
    pCut.Flow = (pCut.Flow + 100 * Of_CutArea(p, nLeaves)) / (nGiaRefs ? nGiaRefs : 1);
}
func Of_ObjMergeOrder(p *Of_Man_t, iObj int32) {
    var pCuts0 [OF_CUT_MAX]Of_Cut_t;
    var pCuts1 [OF_CUT_MAX]Of_Cut_t;
    var pCuts [OF_CUT_MAX]Of_Cut_t;
    var pCutsR [OF_CUT_MAX]*Of_Cut_t;
    var pObj *Gia_Obj_t  = Gia_ManObj(p.pGia, iObj);
    var nGiaRefs int32  = 2*Gia_ObjRefNumId(p.pGia, iObj);
    var nLutSize int32  = p.pPars.nLutSize;
    var nCutNum int32   = p.pPars.nCutNum;
    var nCuts0 int32    = Of_ManPrepareCuts(pCuts0, p, Gia_ObjFaninId0(pObj, iObj), 1);
    var nCuts1 int32    = Of_ManPrepareCuts(pCuts1, p, Gia_ObjFaninId1(pObj, iObj), 1);
    var fComp0 int32    = Gia_ObjFaninC0(pObj);
    var fComp1 int32    = Gia_ObjFaninC1(pObj);
    var iSibl int32     = Gia_ObjSibl(p.pGia, iObj);
    var pCut0 *Of_Cut_t;
    var pCut1 *Of_Cut_t;
    var pCut0Lim *Of_Cut_t = pCuts0 + nCuts0;
    var pCut1Lim *Of_Cut_t = pCuts1 + nCuts1;
    var i, nCutsR int32;
    assert(!Gia_ObjIsBuf(pObj), "!Gia_ObjIsBuf(pObj)");
    for  i = 0; i < nCutNum; i++  {
        pCutsR[i] = pCuts + i;
    }
    if iSibl != 0 {
        var pCuts2 [OF_CUT_MAX]Of_Cut_t;
        var pObjE *Gia_Obj_t  = Gia_ObjSiblObj(p.pGia, iObj);
        var fCompE int32  = Gia_ObjPhase(pObj) ^ Gia_ObjPhase(pObjE);
        var nCuts2 int32  = Of_ManPrepareCuts(pCuts2, p, iSibl, 0);
        var pCut2 *Of_Cut_t;
        var pCut2Lim *Of_Cut_t = pCuts2 + nCuts2;
        for  pCut2 = pCuts2; pCut2 < pCut2Lim; pCut2++  {
            *pCutsR[nCutsR] = *pCut2;
            if p.pPars.fCutMin {
                pCutsR[nCutsR].iFunc = Abc_LitNotCond( pCutsR[nCutsR].iFunc, fCompE ); }
            Of_CutParams( p, pCutsR[nCutsR], nGiaRefs );
            nCutsR = Of_SetAddCut( pCutsR, nCutsR, nCutNum );
        }
    }
    if Gia_ObjIsMuxId(p.pGia, iObj) {
        var pCuts2 [OF_CUT_MAX]Of_Cut_t;
        var nCuts2 int32   = Of_ManPrepareCuts(pCuts2, p, Gia_ObjFaninId2(p.pGia, iObj), 1);
        var fComp2 int32   = Gia_ObjFaninC2(p.pGia, pObj);
        var pCut2 *Of_Cut_t;
        var pCut2Lim *Of_Cut_t = pCuts2 + nCuts2;
        p.CutCount[0] += nCuts0 * nCuts1 * nCuts2;
        for  pCut0 = pCuts0; pCut0 < pCut0Lim; pCut0++  {
        for  pCut1 = pCuts1; pCut1 < pCut1Lim; pCut1++  {
        for  pCut2 = pCuts2; pCut2 < pCut2Lim; pCut2++  {
            if Of_CutCountBits(pCut0.Sign | pCut1.Sign | pCut2.Sign) > nLutSize {
                continue; }
            p.CutCount[1]++; 
            if !(Of_CutMergeOrderMux(pCut0, pCut1, pCut2, pCutsR[nCutsR], nLutSize) != 0) {
                continue; }
            if Of_SetLastCutIsContained(pCutsR, nCutsR) != 0 {
                continue; }
            p.CutCount[2]++;
            if p.pPars.fCutMin && (Of_CutComputeTruthMux6(p, pCut0, pCut1, pCut2, fComp0, fComp1, fComp2, pCutsR[nCutsR]) != 0) {
                pCutsR[nCutsR].Sign = Of_CutGetSign(pCutsR[nCutsR].pLeaves, pCutsR[nCutsR].nLeaves); }
            Of_CutParams( p, pCutsR[nCutsR], nGiaRefs );
            nCutsR = Of_SetAddCut( pCutsR, nCutsR, nCutNum );
        }
        }
        }
    
    } else {
        var fIsXor int32  = Gia_ObjIsXor(pObj);
        p.CutCount[0] += nCuts0 * nCuts1;
        for  pCut0 = pCuts0; pCut0 < pCut0Lim; pCut0++  {
        for  pCut1 = pCuts1; pCut1 < pCut1Lim; pCut1++  {
            if int32((pCut0.nLeaves + pCut1.nLeaves)) > nLutSize && Of_CutCountBits(pCut0.Sign | pCut1.Sign) > nLutSize {
                continue; }
            p.CutCount[1]++; 
            if !(Of_CutMergeOrder(pCut0, pCut1, pCutsR[nCutsR], nLutSize) != 0) {
                continue; }
            if Of_SetLastCutIsContained(pCutsR, nCutsR) != 0 {
                continue; }
            p.CutCount[2]++;
            if p.pPars.fCutMin && (Of_CutComputeTruth6(p, pCut0, pCut1, fComp0, fComp1, pCutsR[nCutsR], fIsXor) != 0) {
                pCutsR[nCutsR].Sign = Of_CutGetSign(pCutsR[nCutsR].pLeaves, pCutsR[nCutsR].nLeaves); }
            Of_CutParams( p, pCutsR[nCutsR], nGiaRefs );
            nCutsR = Of_SetAddCut( pCutsR, nCutsR, nCutNum );
        }
        }
    }
    // debug printout
    if false {
        fmt.Printf("*** Obj = %d\n", iObj);
        for  i = 0; i < nCutsR; i++  {
            Of_Cutprintf( p, pCutsR[i] );
        }
        fmt.Printf("\n");
    } 
    // verify
    assert(nCutsR > 0 && nCutsR < nCutNum, "nCutsR > 0 && nCutsR < nCutNum");
    //assert( Of_SetCheckArray(pCutsR, nCutsR) );
    // store the cutset
    Of_ObjSetCutFlow( p, iObj, pCutsR[0].Flow );
    Of_ObjSetCutDelay( p, iObj, pCutsR[0].Delay );
    *Vec_IntEntryP(&p.vCutSets, iObj) = Of_ManSaveCuts(p, pCutsR, nCutsR);
    p.CutCount[3] += nCutsR;
}
func Of_ManComputeCuts(p *Of_Man_t) {
    var pObj *Gia_Obj_t; var i, iFanin int32;
    // inlining of macro Gia_ManForEachAnd
    for  i = 0; (i < p.pGia.nObjs) && ((pObj) = Gia_ManObj(p.pGia, i)); i++  {      if !Gia_ObjIsAnd(pObj) {} else
        if Gia_ObjIsBuf(pObj) {
            iFanin = Gia_ObjFaninId0(pObj, i);
            Of_ObjSetCutFlow( p, i,  Of_ObjCutFlow(p, iFanin) );
            Of_ObjSetCutDelay( p, i, Of_ObjCutDelay(p, iFanin) );
        
        } else {
            Of_ObjMergeOrder( p, i ); }
    }
    // inlining of macro Gia_ManForEachAnd
    for  i = 0; (i < p.pGia.nObjs) && ((pObj) = Gia_ManObj(p.pGia, i)); i++  {      if !Gia_ObjIsAnd(pObj) {} else
        if !Gia_ObjIsBuf(pObj) {
            Of_ManLiftCuts( p, i ); }
    }
}




/**Function*************************************************************

  Synopsis    []

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Of_ManPrintStats(p *Of_Man_t, pTitle string) {
    if !p.pPars.fVerbose {
        return; }
    fmt.Printf("%s :  ", pTitle);
    fmt.Printf("Delay =%8.2f ", Of_Int2Flt(int32(p.pPars.Delay)));
    fmt.Printf("Area =%8d  ", int32(p.pPars.Area));
    fmt.Printf("Edge =%9d  ", int32(p.pPars.Edge));
    Abc_PrintTime( 1, "Time", Abc_Clock() - p.clkStart );
    fflush( os.Stdout );
}
func Of_ManPrintInit(p *Of_Man_t) {
    var nChoices int32;
    if !p.pPars.fVerbose {
        return; }
    fmt.Printf("LutSize = %d  ", p.pPars.nLutSize);
    fmt.Printf("CutNum = %d  ", p.pPars.nCutNum);
    fmt.Printf("Iter = %d  ", p.pPars.nRounds + p.pPars.nRoundsEla);
    fmt.Printf("Coarse = %d   ", p.pPars.fCoarsen);
    if p.pPars.fCutMin {
    fmt.Printf("Funcs = %d  ", Vec_MemEntryNum(p.vTtMem)); }
    nChoices = Gia_ManChoiceNum( p.pGia );
    if nChoices != 0 {
    fmt.Printf("Choices = %d  ", nChoices); }
    fmt.Printf("\n");
    fmt.Printf("Computing cuts...\r");
    fflush( os.Stdout );
}
func Of_ManPrintQuit(p *Of_Man_t) {
    var MemGia float32   = Gia_ManMemory(p.pGia) / (1<<20);
    var MemMan float32   = 1.0 * sizeof(Of_Obj_t) * Gia_ManObjNum(p.pGia) / (1<<20);
    var MemCuts float32  = 1.0 * sizeof(int) * (1 << 16) * Vec_PtrSize(&p.vPages) / (1<<20);
    var MemTt float32;
    if p.vTtMem {
        MemTt = Vec_MemMemory(p.vTtMem) / (1<<20);
    } else {
        MemTt = 0;
    }
    if p.CutCount[0] == 0 {
        p.CutCount[0] = 1; }
    if !p.pPars.fVerbose {
        return; }
    fmt.Printf("CutPair = %.0f  ", p.CutCount[0]);
    fmt.Printf("Merge = %.0f (%.1f)  ", p.CutCount[1], 1.0*p.CutCount[1]/Gia_ManAndNum(p.pGia));
    fmt.Printf("Eval = %.0f (%.1f)  ", p.CutCount[2], 1.0*p.CutCount[2]/Gia_ManAndNum(p.pGia));
    fmt.Printf("Cut = %.0f (%.1f)  ", p.CutCount[3], 1.0*p.CutCount[3]/Gia_ManAndNum(p.pGia));
//    printf( "Use = %.0f (%.1f)  ",      p->CutCount[4], 1.0*p->CutCount[4]/Gia_ManAndNum(p->pGia) );
//    printf( "Mat = %.0f (%.1f)  ",      p->CutCount[5], 1.0*p->CutCount[5]/Gia_ManAndNum(p->pGia) );
//    printf( "Equ = %d (%.2f %%)  ",     p->nCutUseAll,  100.0*p->nCutUseAll /p->CutCount[0] );
    fmt.Printf("\n");
    fmt.Printf("Gia = %.2f MB  ", MemGia);
    fmt.Printf("Man = %.2f MB  ", MemMan); 
    fmt.Printf("Cut = %.2f MB   ", MemCuts);
    if p.pPars.fCutMin {
    fmt.Printf("TT = %.2f MB  ", MemTt); } 
    fmt.Printf("Total = %.2f MB   ", MemGia + MemMan + MemCuts + MemTt); 
//    printf( "\n" );
    Abc_PrintTime( 1, "Time", Abc_Clock() - p.clkStart );
    fflush( os.Stdout );
}


/**Function*************************************************************

  Synopsis    [Technology mappping.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
/*
static inline int Of_ManComputeForwardCut( Of_Man_t * p, int iObj, int * pCut )
{
    int k, iVar, Delay = 0, Area = Of_CutArea(p, Of_CutSize(pCut));
    int DelayLut1 = p->pPars->nDelayLut1;
    Of_CutForEachVar( pCut, iVar, k )
    {
        Delay = Abc_MaxInt( Delay, Of_ObjDelay1(p, iVar) + DelayLut1 );
        if ( p->Iter )
            Area += Of_ObjRefNum(p, iVar) ? 0 : Of_ObjFlow(p, iVar);
    }
    Of_CutSetDelay1( pCut, Delay );
    if ( p->Iter )
        Of_CutSetAreaFlow( pCut, Area );
    return Delay;
}
static inline void Of_ManComputeForwardObj( Of_Man_t * p, int iObj )
{
    int Delay1 = ABC_INFINITY, Area1 = ABC_INFINITY;
    int * pList = Of_ObjCutSet(p, iObj);
    int i, * pCut, * pCutMin = NULL, * pCutMin2 = NULL;
    // compute cut arrivals
    Of_SetForEachCut( pList, pCut, i )
    {
        int Delay1This = Of_ManComputeForwardCut(p, iObj, pCut);
        if ( Delay1 > Delay1This )
        {
            Delay1  = Delay1This;
            pCutMin = pCut;
        }
        if ( p->Iter && Area1 > Of_CutAreaFlow(pCut) )
        {
            Area1    = Of_CutAreaFlow(pCut);
            pCutMin2 = pCut;
        }
    }
    // if mapping is present, set object arrival equal to cut arrival
    if ( Of_ObjRefNum(p, iObj) )
    {
        pCutMin = Of_ObjCutBestP(p, iObj);
        Delay1 = Of_CutDelay1( pCutMin );
        Of_ObjSetDelay1( p, iObj, Delay1 );
        if ( p->Iter )
            Of_ObjSetFlow( p, iObj, Of_CutAreaFlow(pCutMin) );
    }
    else
    {
        if ( p->Iter == 0 )
        {
            Of_ObjSetCutBestP( p, pList, iObj, pCutMin );
            Of_ObjSetDelay1( p, iObj, Delay1 );
        }
        else
        {
            Of_ObjSetCutBestP( p, pList, iObj, pCutMin2 );
            Of_ObjSetDelay1( p, iObj, Of_CutDelay1(pCutMin2) );
            Of_ObjSetFlow( p, iObj, Of_CutAreaFlow(pCutMin2) );
        }
    }
}
*/

/*
int * Of_CutReferChooseCut( Of_Man_t * p, int Var, int Required, int fSetBest )
{
    int i, CostMin = ABC_INFINITY;
    int * pCutMin = NULL, * pList = Of_ObjCutSet(p, Var);
    int * pCut = Of_ObjCutBestP(p, Var);
    assert( Of_CutDelay1(pCut) <= Required );
//    return pCut;
    // choose cut with smaller area
    Of_SetForEachCut( pList, pCut, i )
    {
        if ( Of_CutDelay1(pCut) > Required )
            continue;
        if ( CostMin > Of_CutAreaFlow(pCut) )
        {
            CostMin = Of_CutAreaFlow(pCut);
            pCutMin = pCut;
        }
    }
    assert( pCutMin != NULL );
    assert( Of_CutDelay1(pCutMin) <= Required );
    if ( fSetBest )
        Of_ObjSetCutBestP( p, pList, Var, pCutMin );
    return pCutMin;
}
int Of_CutRef2_rec( Of_Man_t * p, int * pCut, int Required, int fSetBest )
{
    int i, Var, Count = Of_CutArea(p, Of_CutSize(pCut));
    assert( Of_CutDelay1(pCut) <= Required );
    Required -= p->pPars->nDelayLut1;
    Of_CutForEachVar( pCut, Var, i )
    {
        if ( !Of_ObjCutBest(p, Var) )
            continue;
        if ( !fSetBest )
            Vec_IntPush( &p->vCutRefs, Var );
        if ( Of_ObjRefInc(p, Var) )
            continue;
        Count += Of_CutRef2_rec( p, Of_CutReferChooseCut(p, Var, Required, fSetBest), Required, fSetBest );
    }
    return Count;
}
static inline int Of_CutAreaDerefed2( Of_Man_t * p, int * pCut, int Required )
{
    int Ela1, i, iObj;
    assert( Vec_IntSize(&p->vCutRefs) == 0 );
    Ela1 = Of_CutRef2_rec( p, pCut, Required, 0 );
    Vec_IntForEachEntry( &p->vCutRefs, iObj, i )
        Of_ObjRefDec(p, iObj);
    Vec_IntClear( &p->vCutRefs );
    return Ela1;
}
*/


func Of_ManComputeForwardCut(p *Of_Man_t, iObj int32, pCut *int32) int32 {
    var k, iVar, Delay int32;
    var DelayLut1 int32  = p.pPars.nDelayLut1;
    // inlining of macro Of_CutForEachVar
    for  k = 0; k < Of_CutSize(pCut) && (iVar = Of_CutVar(pCut,k)); k++  {
        Delay = Abc_MaxInt( Delay, Of_ObjDelay1(p, iVar) + DelayLut1 );
    }
    Of_CutSetDelay1( pCut, Delay );
    return Delay;
}
func Of_ManComputeForwardCutArea(p *Of_Man_t, iObj int32, pCut *int32) int32 {
    var k int32;
    var iVar int32;
    var Area int32 = 100 * Of_CutArea(p, Of_CutSize(pCut));
    // inlining of macro Of_CutForEachVar
    for  k = 0; k < Of_CutSize(pCut) && (iVar = Of_CutVar(pCut,k)); k++  {
         Area += Of_ObjFlow(p, iVar);
    }
    return Area / Abc_MaxInt(1, Of_ObjRefNum(p, iObj));
}
func Of_ManComputeForwardObj(p *Of_Man_t, iObj int32) {
    var Delay1 int32  = ABC_INFINITY;
    var i int32;
    var pCut *int32;
    var pCutMin *int32 = nil;
    var pList *int32 = Of_ObjCutSet(p, iObj);
    // compute cut arrivals
    // inlining of macro Of_SetForEachCut
    for  i, pCut = 0, pList + 1; i < pList[0]; i++, pCut += Of_CutSize(pCut) + OF_CUT_EXTRA  {
        var Delay1This int32  = Of_ManComputeForwardCut(p, iObj, pCut);
        if Delay1 > Delay1This {
            Delay1  = Delay1This;
            pCutMin = pCut;
        }
    }
    // if mapping is present, set object arrival equal to cut arrival
    if Of_ObjRefNum(p, iObj) != 0 {
        pCutMin = Of_ObjCutBestP(p, iObj); }
    Of_ObjSetCutBestP( p, pList, iObj, pCutMin );
    Of_ObjSetDelay1( p, iObj, Of_CutDelay1(pCutMin) );
    if p.Iter {
        Of_ObjSetFlow( p, iObj, Of_ManComputeForwardCutArea(p, iObj, pCutMin) ); }
}
func Of_ManComputeForward1(p *Of_Man_t) {
    var pObj *Gia_Obj_t; var i int32;
    // inlining of macro Gia_ManForEachAnd
    for  i = 0; (i < p.pGia.nObjs) && ((pObj) = Gia_ManObj(p.pGia, i)); i++  {      if !Gia_ObjIsAnd(pObj) {} else
        if Gia_ObjIsBuf(pObj) {
            Of_ObjSetDelay1( p, i, Of_ObjDelay1(p, Gia_ObjFaninId0(pObj, i)) );
        } else {
            Of_ManComputeForwardObj( p, i ); }
    }
}


func Of_CutRef_rec(p *Of_Man_t, pCut *int32) int32 {
    var i int32;
    var Var int32;
    var Count int32;
    if p.Iter & 1 {
        Count = 1;
    } else {
        Count = Of_CutArea(p, Of_CutSize(pCut));
    }
    // inlining of macro Of_CutForEachVar
    for  i = 0; i < Of_CutSize(pCut) && (Var = Of_CutVar(pCut,i)); i++  {
        if (Of_ObjCutBest(p, Var) != 0) && !(Of_ObjRefInc(p, Var) != 0) {
            Count += Of_CutRef_rec( p, Of_ObjCutBestP(p, Var) ); }
    }
    return Count;
}
func Of_CutDeref_rec(p *Of_Man_t, pCut *int32) int32 {
    var i int32;
    var Var int32;
    var Count int32;
    if p.Iter & 1 {
        Count = 1;
    } else {
        Count = Of_CutArea(p, Of_CutSize(pCut));
    }
    // inlining of macro Of_CutForEachVar
    for  i = 0; i < Of_CutSize(pCut) && (Var = Of_CutVar(pCut,i)); i++  {
        if (Of_ObjCutBest(p, Var) != 0) && !(Of_ObjRefDec(p, Var) != 0) {
            Count += Of_CutDeref_rec( p, Of_ObjCutBestP(p, Var) ); }
    }
    return Count;
}
func Of_CutAreaDerefed(p *Of_Man_t, pCut *int32) int32 {
    var Ela1 int32  = Of_CutRef_rec( p, pCut );
    var Ela2 int32  = Of_CutDeref_rec( p, pCut );
    assert(Ela1 == Ela2, "Ela1 == Ela2");
    return Ela1;
}

func Of_CutRef2_rec(p *Of_Man_t, pCut *int32) int32 {
    var i int32;
    var Var int32;
    var Count int32;
    if p.Iter & 1 {
        Count = 1;
    } else {
        Count = Of_CutArea(p, Of_CutSize(pCut));
    }
    // inlining of macro Of_CutForEachVar
    for  i = 0; i < Of_CutSize(pCut) && (Var = Of_CutVar(pCut,i)); i++  {
        if !(Of_ObjCutBest(p, Var) != 0) {
            continue; }
        Vec_IntPush( &p.vCutRefs, Var );
        if Of_ObjRefInc(p, Var) != 0 {
            continue; }
        Count += Of_CutRef2_rec( p, Of_ObjCutBestP(p, Var) );
    }
    return Count;
}
func Of_CutAreaDerefed2(p *Of_Man_t, pCut *int32) int32 {
    var Ela1, i, iObj int32;
    assert(Vec_IntSize(&p.vCutRefs) == 0, "Vec_IntSize(&p.vCutRefs) == 0");
    Ela1 = Of_CutRef2_rec( p, pCut );
    // inlining of macro Vec_IntForEachEntry
    for  i = 0; (i < Vec_IntSize(&p.vCutRefs)) && (((iObj) = Vec_IntEntry(&p.vCutRefs, i)), 1); i++  {
        Of_ObjRefDec(p, iObj);
    }
    Vec_IntClear( &p.vCutRefs );
    return Ela1;
}


func Of_ManComputeForwardObj2(p *Of_Man_t, iObj int32) {
    var Delay int32;
    var Required int32 = Of_ObjRequired(p, iObj);
    var AreaBef int32;
    var AreaAft int32;
    var Area int32;
    var AreaMin int32 = ABC_INFINITY;
    var k int32;
    var pCut *int32;
    var pCutMin *int32 = nil;
    var pList *int32 = Of_ObjCutSet(p, iObj);
    if Of_ObjRefNum(p, iObj) != 0 {
        AreaBef = Of_CutDeref_rec( p, Of_ObjCutBestP(p, iObj) ); }        
    // inlining of macro Of_SetForEachCut
    for  k, pCut = 0, pList + 1; k < pList[0]; k++, pCut += Of_CutSize(pCut) + OF_CUT_EXTRA  {
        Delay = Of_ManComputeForwardCut(p, iObj, pCut);
        if Delay > Required {
            continue; }
        Area = Of_CutAreaDerefed2( p, pCut );
        if AreaMin > Area {
            AreaMin = Area;
            pCutMin = pCut;
        }
    }
    assert(pCutMin != nil, "pCutMin != nil");
    Of_ObjSetCutBestP( p, pList, iObj, pCutMin );
    if Of_ObjRefNum(p, iObj) != 0 {
        AreaAft = Of_CutRef_rec( p, pCutMin ); }
    assert(AreaAft <= AreaBef, "AreaAft <= AreaBef");
    Delay = Of_CutDelay1(pCutMin);
    assert(Delay <= Required, "Delay <= Required");
    Of_ObjSetDelay1( p, iObj, Delay );
}
func Of_ManComputeForward2(p *Of_Man_t) {
    var pObj *Gia_Obj_t; var i int32;
    // inlining of macro Gia_ManForEachAnd
    for  i = 0; (i < p.pGia.nObjs) && ((pObj) = Gia_ManObj(p.pGia, i)); i++  {      if !Gia_ObjIsAnd(pObj) {} else
        if Gia_ObjIsBuf(pObj) {
            Of_ObjSetDelay1( p, i, Of_ObjDelay1(p, Gia_ObjFaninId0(pObj, i)) );
        } else {
            Of_ManComputeForwardObj2( p, i ); }
    }
}


func Of_ManComputeOutputRequired(p *Of_Man_t, fCleanRefs int32) int32 {
    var i, Id, Delay int32;
    for  i = 0; i < Gia_ManObjNum(p.pGia); i++  {
        Of_ObjSetRequired( p, i, ABC_INFINITY );
        if fCleanRefs != 0 {
            Of_ObjSetRefNum( p, i, 0 ); }
    }
    // inlining of macro Gia_ManForEachCoDriverId
    for  i = 0; (i < Vec_IntSize(p.pGia.vCos)) && (((Id) = Gia_ObjFaninId0p(p.pGia, Gia_ManCo(p.pGia, i))), 1); i++  {
        Delay = Abc_MaxInt( Delay, Of_ObjDelay1(p, Id) );
    }
    // inlining of macro Gia_ManForEachCoDriverId
    for  i = 0; (i < Vec_IntSize(p.pGia.vCos)) && (((Id) = Gia_ObjFaninId0p(p.pGia, Gia_ManCo(p.pGia, i))), 1); i++  {
        Of_ObjUpdateRequired( p, Id, Delay );
        if fCleanRefs != 0 {
            Of_ObjRefInc( p, Id ); }
    }
    if p.pPars.Delay && p.pPars.Delay < Delay {
        fmt.Printf("Error: Delay violation.\n"); }
    p.pPars.Delay = Delay;
    return Delay;
}
func Of_ManComputeBackwardCut(p *Of_Man_t, pCut *int32) int32 {
    var k, iVar, Cost int32;
    // inlining of macro Of_CutForEachVar
    for  k = 0; k < Of_CutSize(pCut) && (iVar = Of_CutVar(pCut,k)); k++  {
        if !(Of_ObjRefNum(p, iVar) != 0) {
            Cost += Of_ObjFlow( p, iVar ); }
    }
    return Cost;
}
func Of_ManComputeBackward1(p *Of_Man_t) {
    var pObj *Gia_Obj_t; 
    var DelayLut1 int32  = p.pPars.nDelayLut1;
    var i int32;
    var k int32;
    var iVar int32;
    var pList *int32;
    var pCut *int32;
    var pCutMin *int32;
    Of_ManComputeOutputRequired( p, 1 );
    // compute area and edges
    p.pPars.Area = p.pPars.Edge = 0;
    // inlining of macro Gia_ManForEachAndReverse
    for  i = p.pGia.nObjs - 1; (i > 0) && ((pObj) = Gia_ManObj(p.pGia, i)); i--  {  if !Gia_ObjIsAnd(pObj) {} else {
        var CostMin int32;
        var Cost int32;
        var Required int32 = Of_ObjRequired(p, i);
        if Gia_ObjIsBuf(pObj) {
            var FaninId int32  = Gia_ObjFaninId0(pObj, i);
            Of_ObjUpdateRequired( p, FaninId, Required );
            Of_ObjRefInc( p, FaninId );
            continue;
        }
        if !(Of_ObjRefNum(p, i) != 0) {
            continue; }
        // select the best cut
        pCutMin = nil;
        CostMin = ABC_INFINITY;
        pList = Of_ObjCutSet( p, i );
        // inlining of macro Of_SetForEachCut
        for  k, pCut = 0, pList + 1; k < pList[0]; k++, pCut += Of_CutSize(pCut) + OF_CUT_EXTRA  {
            if Of_CutDelay1(pCut) > Required {
                continue; }
            Cost = Of_ManComputeBackwardCut( p, pCut );
            if CostMin > Cost {
                CostMin = Cost;
                pCutMin = pCut;
            }
        }
        // the cut is selected
        assert(pCutMin != nil, "pCutMin != nil");
        Of_ObjSetCutBestP( p, pList, i, pCutMin );
        // inlining of macro Of_CutForEachVar
        for  k = 0; k < Of_CutSize(pCutMin) && (iVar = Of_CutVar(pCutMin,k)); k++  {
            Of_ObjUpdateRequired( p, iVar, Required - DelayLut1 );
            Of_ObjRefInc( p, iVar );
        }
        // update parameters
        p.pPars.Edge += Of_CutSize(pCutMin);
        p.pPars.Area++;
    }
    }
}
func Of_ManComputeBackward2(p *Of_Man_t) {
    var pObj *Gia_Obj_t; 
    var DelayLut1 int32  = p.pPars.nDelayLut1;
    var i int32;
    var k int32;
    var iVar int32;
    var pCutMin *int32;
    Of_ManComputeOutputRequired( p, 0 );
    // compute area and edges
    p.pPars.Area = p.pPars.Edge = 0;
    // inlining of macro Gia_ManForEachAndReverse
    for  i = p.pGia.nObjs - 1; (i > 0) && ((pObj) = Gia_ManObj(p.pGia, i)); i--  {  if !Gia_ObjIsAnd(pObj) {} else {
        var Required int32  = Of_ObjRequired(p, i);
        if Gia_ObjIsBuf(pObj) {
            var FaninId int32  = Gia_ObjFaninId0(pObj, i);
            Of_ObjUpdateRequired( p, FaninId, Required );
            continue;
        }
        if !(Of_ObjRefNum(p, i) != 0) {
            continue; }
        // lookup for the cut
        pCutMin = Of_ObjCutBestP( p, i );
        // inlining of macro Of_CutForEachVar
        for  k = 0; k < Of_CutSize(pCutMin) && (iVar = Of_CutVar(pCutMin,k)); k++  {
            Of_ObjUpdateRequired( p, iVar, Required - DelayLut1 );
        }
        // update parameters
        p.pPars.Edge += Of_CutSize(pCutMin);
        p.pPars.Area++;
    }
    }
}
func Of_ManComputeBackward3(p *Of_Man_t) {
    var pObj *Gia_Obj_t; 
    var DelayLut1 int32  = p.pPars.nDelayLut1;
    var i int32;
    var k int32;
    var iVar int32;
    var pList *int32;
    var pCut *int32;
    var pCutMin *int32;
    var AreaBef, AreaAft int32;
    Of_ManComputeOutputRequired( p, 0 );
    // compute area and edges
    p.pPars.Area = p.pPars.Edge = 0;
    // inlining of macro Gia_ManForEachAndReverse
    for  i = p.pGia.nObjs - 1; (i > 0) && ((pObj) = Gia_ManObj(p.pGia, i)); i--  {  if !Gia_ObjIsAnd(pObj) {} else {
        var CostMin int32;
        var Cost int32;
        var Required int32 = Of_ObjRequired(p, i);
        if Gia_ObjIsBuf(pObj) {
            var FaninId int32  = Gia_ObjFaninId0(pObj, i);
            Of_ObjUpdateRequired( p, FaninId, Required );
            continue;
        }
        if !(Of_ObjRefNum(p, i) != 0) {
            continue; }
        // deref best cut
        AreaBef = Of_CutDeref_rec( p, Of_ObjCutBestP(p, i) );        
        // select the best cut
        pCutMin = nil;
        CostMin = ABC_INFINITY;
        pList = Of_ObjCutSet( p, i );
        // inlining of macro Of_SetForEachCut
        for  k, pCut = 0, pList + 1; k < pList[0]; k++, pCut += Of_CutSize(pCut) + OF_CUT_EXTRA  {
            if Of_CutDelay1(pCut) > Required {
                continue; }
            Cost = Of_CutAreaDerefed2( p, pCut );
            if CostMin > Cost {
                CostMin = Cost;
                pCutMin = pCut;
            }
        }
        // the cut is selected
        assert(pCutMin != nil, "pCutMin != nil");
        Of_ObjSetCutBestP( p, pList, i, pCutMin );
        // inlining of macro Of_CutForEachVar
        for  k = 0; k < Of_CutSize(pCutMin) && (iVar = Of_CutVar(pCutMin,k)); k++  {
            Of_ObjUpdateRequired( p, iVar, Required - DelayLut1 );
        }
        // ref best cut
        AreaAft = Of_CutRef_rec( p, pCutMin );
        assert(AreaAft <= AreaBef, "AreaAft <= AreaBef");
        // update parameters
        p.pPars.Edge += Of_CutSize(pCutMin);
        p.pPars.Area++;
    }
    }
}

/**Function*************************************************************

  Synopsis    [Technology mappping.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Of_ManComputeForwardDirconCut(p *Of_Man_t, iObj int32, pCut *int32, pDelay1 *int32, pDelay2 *int32) {
    // Delay1 - main delay;  Delay2 - precomputed LUT delay in terms of Delay1 for the fanins
    var Delays [6]int32;
    var Perm [6]int32 = [6]int32{0, 1, 2, 3, 4, 5}; 
    var DelayLut1 int32  = p.pPars.nDelayLut1;
    var DelayLut2 int32  = p.pPars.nDelayLut2;
    var nSize int32  = Of_CutSize(pCut);
    var k, iVar, Flag, SlowCon, Delay, DelayAfter, fDirConWorks int32;
    // inlining of macro Of_CutForEachVar
    for  k = 0; k < Of_CutSize(pCut) && (iVar = Of_CutVar(pCut,k)); k++  {
        Delays[k] = Of_ObjDelay1(p, iVar) + DelayLut1;
//        printf( "%3d%s ", iVar, Flag ? "*" : " " );
    }
    for  ; k < p.pPars.nLutSize; k++  {
        Delays[k] = -ABC_INFINITY;
//        printf( "     " );
    }
    Vec_IntSelectSortCost2Reverse( Perm, nSize, Delays );
    assert(nSize < 2 || Delays[0] >= Delays[nSize-1], "nSize < 2 || Delays[0] >= Delays[nSize-1]");
    assert(Delays[0] >= 0 && Delays[nSize-1] >= 0, "Delays[0] >= 0 && Delays[nSize-1] >= 0");
    // consider speedup due to dircons
    fDirConWorks = 1;
    *pDelay1 = *pDelay2 = 0;
    if p.pPars.nFastEdges < nSize {
        SlowCon = Delays[p.pPars.nFastEdges];
    } else {
        SlowCon = 0;
    }
    for  k = 0; k < nSize; k++  {
        // use dircon if the following is true
        // - the input is eligible for dircon (does not exceed the limit)
        // - there is an expected gain in delay, compared the largest delay without dircon
        // - the dircon delay is indeed lower than the largest delay without dircon
        // - all previous dircons worked out well
        // - the node is an AND-gate 
        iVar = Of_CutVar( pCut, Perm[k] );
        assert(Delays[k] == Of_ObjDelay1(p, iVar) + DelayLut1, "Delays[k] == Of_ObjDelay1(p, iVar) + DelayLut1");
        DelayAfter = Of_ObjDelay2(p, iVar) + DelayLut2;
        if k < p.pPars.nFastEdges && Delays[k] > SlowCon && DelayAfter < Delays[k] && (fDirConWorks != 0) && Gia_ObjIsAndNotBuf(Gia_ManObj(p.pGia, iVar)) {
            Delay = DelayAfter;
            Of_CutSetFlag( pCut, Perm[k], 1 );
        
        } else {
            Delay = Delays[k];// + DelayLut2;
            Of_CutSetFlag( pCut, Perm[k], 0 );
            fDirConWorks = 0;
        }
        *pDelay1 = Abc_MaxInt( *pDelay1, Delay );
        *pDelay2 = Abc_MaxInt( *pDelay2, Delays[k] );
    }
//    printf( "   %5.2f",   Of_Int2Flt(*pDelay1) );
//    printf( "   %5.2f\n", Of_Int2Flt(*pDelay2) );
    // do not use the structure if simple LUT is better
    if *pDelay1 > *pDelay2 {
        for  k = 0; k < nSize; k++  {
            Of_CutSetFlag( pCut, k, 0 );
        }
        *pDelay1 = *pDelay2;
    }
    assert(*pDelay1 <= *pDelay2, "*pDelay1 <= *pDelay2");
    Of_CutSetDelay1( pCut, *pDelay1 );
    Of_CutSetDelay2( pCut, *pDelay2 );
    // verify
    // inlining of macro Of_CutForEachVarFlag
    for  k = 0; k < Of_CutSize(pCut) && (iVar = Of_CutVar(pCut,k)) && ((Flag = Of_CutFlag(pCut,k)), 1); k++  {
        if Flag != 0 {
            assert(Of_ObjDelay2(p, iVar) + DelayLut2 <= *pDelay1, "Of_ObjDelay2(p, iVar) + DelayLut2 <= *pDelay1");
        } else {
            assert(Of_ObjDelay1(p, iVar) + DelayLut1 <= *pDelay1, "Of_ObjDelay1(p, iVar) + DelayLut1 <= *pDelay1"); }
        assert(Of_ObjDelay1(p, iVar) + DelayLut1 <= *pDelay2, "Of_ObjDelay1(p, iVar) + DelayLut1 <= *pDelay2");
    }
}
func Of_ManComputeForwardDirconObj(p *Of_Man_t, iObj int32) int32 {
    var Delay1 int32 = ABC_INFINITY;
    var Delay2 int32 = ABC_INFINITY;
    var i int32;
    var pCut *int32;
    var pCutMin *int32 = nil;
    var pCutMin2 *int32 = nil;
    var pList *int32 = Of_ObjCutSet(p, iObj);
    // inlining of macro Of_SetForEachCut
    for  i, pCut = 0, pList + 1; i < pList[0]; i++, pCut += Of_CutSize(pCut) + OF_CUT_EXTRA  {
        var Delay1This, Delay2This int32;
        Of_ManComputeForwardDirconCut( p, iObj, pCut, &Delay1This, &Delay2This );
        if Delay1 > Delay1This {
            pCutMin = pCut; }
        if Delay2 > Delay2This {
            pCutMin2 = pCut; }
        Delay1 = Abc_MinInt( Delay1, Delay1This );
        Delay2 = Abc_MinInt( Delay2, Delay2This );
    }
    Of_ObjSetDelay1( p, iObj, Delay1 );
    Of_ObjSetDelay2( p, iObj, Delay2 );
    Of_ObjSetCutBestP( p, pList, iObj, pCutMin );
    Of_ObjSetCutBestP2( p, pList, iObj, pCutMin2 );
    return Delay1;
}
func Of_ManComputeForwardDircon1(p *Of_Man_t) {
    var pObj *Gia_Obj_t; var i int32;
    // inlining of macro Gia_ManForEachAnd
    for  i = 0; (i < p.pGia.nObjs) && ((pObj) = Gia_ManObj(p.pGia, i)); i++  {      if !Gia_ObjIsAnd(pObj) {} else
        if Gia_ObjIsBuf(pObj) {
            Of_ObjSetDelay1( p, i, Of_ObjDelay1(p, Gia_ObjFaninId0(pObj, i)) );
            Of_ObjSetDelay2( p, i, Of_ObjDelay2(p, Gia_ObjFaninId0(pObj, i)) );
        
        } else {
            Of_ManComputeForwardDirconObj( p, i ); }
    }
}
func Of_ManComputeBackwardDircon1(p *Of_Man_t) {
    var pObj *Gia_Obj_t; 
    var vPointed *Vec_Bit_t; 
    var DelayLut1 int32  = p.pPars.nDelayLut1;
    var DelayLut2 int32  = p.pPars.nDelayLut2;
    var i int32;
    var k int32;
    var iVar int32;
    var Flag int32;
    var pList *int32;
    var pCutMin *int32;
    var CountNodes, CountEdges int32;
    Of_ManComputeOutputRequired( p, 1 );
    fmt.Printf("Global delay =%8.2f\n", Of_Int2Flt(int32(p.pPars.Delay)));
    //return;
    // compute area and edges
    vPointed = Vec_BitStart( Gia_ManObjNum(p.pGia) );
    p.pPars.Area = p.pPars.Edge = 0;
    // inlining of macro Gia_ManForEachAndReverse
    for  i = p.pGia.nObjs - 1; (i > 0) && ((pObj) = Gia_ManObj(p.pGia, i)); i--  {  if !Gia_ObjIsAnd(pObj) {} else {
        var CostMin int32;
        var fPointed int32;
        var Required int32 = Of_ObjRequired(p, i);
        if Gia_ObjIsBuf(pObj) {
            var FaninId int32  = Gia_ObjFaninId0(pObj, i);
            Of_ObjUpdateRequired( p, FaninId, Required );
            Of_ObjRefInc( p, FaninId );
            continue;
        }
        if !(Of_ObjRefNum(p, i) != 0) {
            continue; }
        // check if the LUT is has an outgoing dircon edge
        fPointed = Vec_BitEntry(vPointed, i);
        CountNodes += fPointed;

/*
        // select the best cut
        {
            int * pCut;
            pCutMin = NULL;
            CostMin = ABC_INFINITY;
            pList = Of_ObjCutSet( p, i );
            Of_SetForEachCut( pList, pCut, k )
            {
                int Cost;
                if ( (fPointed ? Of_CutDelay2(pCut) : Of_CutDelay1(pCut)) > Required )
                    continue;
                Cost = Of_ManComputeBackwardCut( p, pCut );
                if ( CostMin > Cost )
                {
                    CostMin = Cost;
                    pCutMin = pCut;
                }
            }
        }
*/

        if fPointed != 0 {
            pCutMin = Of_ObjCutBestP2( p, i );
            CostMin = Of_CutDelay2(pCutMin);
            //assert( Of_CutDelay2(pCutMin) <= Required );
        
        } else {
            pCutMin = Of_ObjCutBestP( p, i );
            CostMin = Of_CutDelay1(pCutMin);
            //assert( Of_CutDelay1(pCutMin) <= Required );
        }

        // remove dircon markers
        //if ( fPointed )
        //    Of_CutForEachVarFlag( pCutMin, iVar, Flag, k )
        //        Of_CutSetFlag( pCutMin, k, 0 );

        // the cut is selected
        assert(pCutMin != nil, "pCutMin != nil");
        pList = Of_ObjCutSet( p, i );
        Of_ObjSetCutBestP( p, pList, i, pCutMin );  ///// SET THE BEST CUT
        // inlining of macro Of_CutForEachVarFlag
        for  k = 0; k < Of_CutSize(pCutMin) && (iVar = Of_CutVar(pCutMin,k)) && ((Flag = Of_CutFlag(pCutMin,k)), 1); k++  {
            Of_ObjUpdateRequired( p, iVar, Required - ((Flag && !fPointed) ? DelayLut2 : DelayLut1) );
            Of_ObjRefInc( p, iVar );           
            if (Flag != 0) && !(fPointed != 0) {             
                Vec_BitWriteEntry( vPointed, iVar, 1 );
                CountEdges++;
            }
        }
        // update parameters
        p.pPars.Edge += Of_CutSize(pCutMin);
        p.pPars.Area++;
    }
    }
    Vec_BitFree( vPointed );
    //printf( "Dircon nodes = %d.  Dircon edges = %d.\n", CountNodes, CountEdges );
}

/**Function*************************************************************

  Synopsis    []

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Of_ManCreateSat(p *Of_Man_t, nCutsAll int32, vFirst *Vec_Int_t, vCutNum *Vec_Int_t, vBestNode *Vec_Int_t, vBestCut *Vec_Int_t) {
    extern void Cnf_AddCardinConstrPairWise( sat_solver * p, Vec_Int_t * vVars, int K, int fStrict );

    var pObj, pVar *Gia_Obj_t;
    var pCutSet, pCut *int32;
    var i int32;
    var k int32;
    var v int32;
    var c int32;
    var Var int32;
    var Lit int32;
    var pLits [2]int32;
    var status int32;
    var RetValue int32;
    var nCutCount int32;
    var nClauses int32;
    var vLits *Vec_Int_t  = Vec_IntAlloc( 100 );
    var clk abctime  = Abc_Clock();

    // start solver
    var pSat *sat_solver  = sat_solver_new();
    sat_solver_setnvars( pSat, Gia_ManAndNum(p.pGia) + nCutsAll );

    // set polarity
    Vec_IntAppend( vBestNode, vBestCut );
    //Vec_IntPrint( vBestNode );
    sat_solver_set_polarity( pSat, Vec_IntArray(vBestNode), Vec_IntSize(vBestNode) );
    Vec_IntShrink( vBestNode, Vec_IntSize(vBestNode) - Vec_IntSize(vBestCut) );

    // add clauses for nodes
    // inlining of macro Gia_ManForEachAnd
    for  i = 0; (i < p.pGia.nObjs) && ((pObj) = Gia_ManObj(p.pGia, i)); i++  {      if !Gia_ObjIsAnd(pObj) {} else {
        var iFirst int32  = Vec_IntEntry(vFirst, i);
        var nCuts int32  = Vec_IntEntry(vCutNum, i);
        Vec_IntClear( vLits );
        Vec_IntPush( vLits, Abc_Var2Lit(pObj.Value, 1) );
        for  c = 0; c < nCuts; c++  {
            Vec_IntPush( vLits, Abc_Var2Lit(iFirst + c, 0) );
        }
        RetValue = sat_solver_addclause( pSat, Vec_IntArray(vLits), Vec_IntLimit(vLits) );
        assert((RetValue != 0), "RetValue");
    }
    }

    // add clauses for cuts
    nCutCount = 0;
    // inlining of macro Gia_ManForEachAnd
    for  i = 0; (i < p.pGia.nObjs) && ((pObj) = Gia_ManObj(p.pGia, i)); i++  {      if !Gia_ObjIsAnd(pObj) {} else {
        pCutSet = Of_ObjCutSet(p, i);
        // inlining of macro Of_SetForEachCut
        for  k, pCut = 0, pCutSet + 1; k < pCutSet[0]; k++, pCut += Of_CutSize(pCut) + OF_CUT_EXTRA  {
            pLits[0] = Abc_Var2Lit( Gia_ManAndNum(p.pGia) + nCutCount, 1 );
            pLits[1] = Abc_Var2Lit( pObj.Value, 0 );
            RetValue = sat_solver_addclause( pSat, pLits, pLits+2 );
            assert((RetValue != 0), "RetValue");
            // inlining of macro Of_CutForEachVar
            for  v = 0; v < Of_CutSize(pCut) && (Var = Of_CutVar(pCut,v)); v++  {
                pVar = Gia_ManObj(p.pGia, Var);
                if !Gia_ObjIsAnd(pVar) {
                    continue; }
                pLits[1] = Abc_Var2Lit( pVar.Value, 0 );
                RetValue = sat_solver_addclause( pSat, pLits, pLits+2 );
                assert((RetValue != 0), "RetValue");
            }
            nCutCount++;
        }
    }
    }
    assert(nCutCount == nCutsAll, "nCutCount == nCutsAll");

    // mark CO drivers
    // inlining of macro Gia_ManForEachCo
    for  i = 0; (i < Vec_IntSize(p.pGia.vCos)) && ((pObj) = Gia_ManCo(p.pGia, i)); i++  {
        Gia_ObjFanin0(pObj).fMark0 = 1;
    }
    // set used nodes to 1
    // inlining of macro Gia_ManForEachAnd
    for  i = 0; (i < p.pGia.nObjs) && ((pObj) = Gia_ManObj(p.pGia, i)); i++  {      if !Gia_ObjIsAnd(pObj) {} else
        if pObj.fMark0 {
            Lit = Abc_Var2Lit( pObj.Value, 0 );
            RetValue = sat_solver_addclause( pSat, &Lit, &Lit + 1 );
            assert((RetValue != 0), "RetValue");
        }
    }
    // unmark CO drivers
    // inlining of macro Gia_ManForEachCo
    for  i = 0; (i < Vec_IntSize(p.pGia.vCos)) && ((pObj) = Gia_ManCo(p.pGia, i)); i++  {
        Gia_ObjFanin0(pObj).fMark0 = 0;
    }

//    Sat_SolverWriteDimacs( pSat, "temp.cnf", NULL, NULL, 0 );

    // add cardinality constraint
    nClauses = pSat.stats.clauses;
    Vec_IntClear( vLits );
    Vec_IntFillNatural( vLits, Gia_ManAndNum(p.pGia) );
    Cnf_AddCardinConstrPairWise( pSat, vLits, Vec_IntSize(vBestNode)-2, 0 );
    fmt.Printf("Problem clauses = %d.  Cardinality clauses = %d.\n", nClauses, pSat.stats.clauses - nClauses);

    // solve the problem
    status = sat_solver_solve( pSat, nil, nil, 1000000, 0, 0, 0 );
    if status == l_Undef {
        fmt.Printf("Undecided.  "); }
    if status == l_True {
        fmt.Printf("Satisfiable.  "); }
    if status == l_False {
        fmt.Printf("Unsatisfiable. "); }
    Abc_PrintTime( 1, "Time", Abc_Clock() - clk );
    Sat_SolverPrintStats( os.Stdout, pSat );
    if status == l_True {
        var nOnes int32;
        for  v = 0; v < Gia_ManAndNum(p.pGia); v++  {
            fmt.Printf("%d", sat_solver_var_value(pSat, v));
            nOnes += sat_solver_var_value(pSat, v);
        }
        fmt.Printf(" Nodes = %d\n", nOnes);

        nOnes = 0;
        for  ; v < Gia_ManAndNum(p.pGia) + nCutsAll; v++  {
            fmt.Printf("%d", sat_solver_var_value(pSat, v));
            nOnes += sat_solver_var_value(pSat, v);
        }
        fmt.Printf(" LUTs = %d\n", nOnes);
    }

    // cleanup
    sat_solver_delete( pSat );
    Vec_IntFree( vLits );
}
func Of_ManPrintCuts(p *Of_Man_t) {
    var fVerbose int32;
    var pObj *Gia_Obj_t;
    var pCutSet, pCut, pCutBest *int32;
    var i, k, v, Var, nCuts int32;
    var vFirst *Vec_Int_t   = Vec_IntStartFull( Gia_ManObjNum(p.pGia) );
    var vCutNum *Vec_Int_t  = Vec_IntStartFull( Gia_ManObjNum(p.pGia) );
    var vBestNode *Vec_Int_t  = Vec_IntAlloc( 100 );
    var vBestCut *Vec_Int_t   = Vec_IntAlloc( 100 );
    var nAndsAll int32;
    var nCutsAll int32;
    var Shift int32 = Gia_ManAndNum(p.pGia);
    Gia_ManFillValue( p.pGia );
    // inlining of macro Gia_ManForEachAnd
    for  i = 0; (i < p.pGia.nObjs) && ((pObj) = Gia_ManObj(p.pGia, i)); i++  {      if !Gia_ObjIsAnd(pObj) {} else {
        // get the best cut
        pCutBest = nil;
        if Of_ObjRefNum(p, i) != 0 {
            Vec_IntPush( vBestNode, nAndsAll );
            pCutBest = Of_ObjCutBestP( p, i ); 
        }
        pObj.Value = nAndsAll++;
        // get the cutset
        pCutSet = Of_ObjCutSet(p, i);
        // count cuts
        nCuts = 0;
        // inlining of macro Of_SetForEachCut
        for  k, pCut = 0, pCutSet + 1; k < pCutSet[0]; k++, pCut += Of_CutSize(pCut) + OF_CUT_EXTRA  {
            nCuts++;
        }
        // save
        Vec_IntWriteEntry( vFirst,  i, Shift + nCutsAll );
        Vec_IntWriteEntry( vCutNum, i, nCuts );
        // print cuts
        if fVerbose != 0 {
            fmt.Printf("Node %d. Cuts %d.\n", i, nCuts); }
        // inlining of macro Of_SetForEachCut
        for  k, pCut = 0, pCutSet + 1; k < pCutSet[0]; k++, pCut += Of_CutSize(pCut) + OF_CUT_EXTRA  {
            if fVerbose != 0 {
                fmt.Printf("{ ");
                // inlining of macro Of_CutForEachVar
                for  v = 0; v < Of_CutSize(pCut) && (Var = Of_CutVar(pCut,v)); v++  {
                    fmt.Printf("%d ", Var);
                }
                fmt.Printf("} %s\n", pCutBest == pCut ? "best" :"");
            }
            if pCutBest == pCut {
                Vec_IntPush( vBestCut, Shift + nCutsAll ); }
            nCutsAll++;
        }
    }
    }
    assert(nAndsAll == Shift, "nAndsAll == Shift");
    fmt.Printf("Total:   Ands = %d.  Luts = %d.  Cuts = %d.\n", nAndsAll, Vec_IntSize(vBestNode), nCutsAll);

    // create SAT problem
    Of_ManCreateSat( p, nCutsAll, vFirst, vCutNum, vBestNode, vBestCut );

    Vec_IntFree( vFirst );
    Vec_IntFree( vCutNum );
    Vec_IntFree( vBestNode );
    Vec_IntFree( vBestCut );
}

/**Function*************************************************************

  Synopsis    [Technology mappping.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Of_ManSetDefaultPars(pPars *Jf_Par_t) {
    memset( pPars, 0, sizeof(Jf_Par_t) );
    pPars.nLutSize     =  4;
    pPars.nCutNum      = 16;
    pPars.nProcNum     =  0;
    pPars.nRounds      =  3;
    pPars.nRoundsEla   =  4;
    pPars.nRelaxRatio  =  0;
    pPars.nCoarseLimit =  3;
    pPars.nAreaTuner   = 10;
    pPars.DelayTarget  = -1;
    pPars.nDelayLut1   = 10;
    pPars.nDelayLut2   =  2;
    pPars.nFastEdges   =  0; //
    pPars.fAreaOnly    =  0;
    pPars.fOptEdge     =  1; 
    pPars.fCoarsen     =  0;
    pPars.fCutMin      =  0;
    pPars.fGenCnf      =  0;
    pPars.fPureAig     =  0;
    pPars.fVerbose     =  0;
    pPars.fVeryVerbose =  0;
    pPars.nLutSizeMax  =  OF_LEAF_MAX;
    pPars.nCutNumMax   =  OF_CUT_MAX;
    pPars.MapDelayTarget = -1;
}
func Of_ManDeriveMapping(p *Of_Man_t) *Gia_Man_t {
    var vMapping *Vec_Int_t;
    var vPacking *Vec_Int_t = nil;
    var vPointed *Vec_Bit_t; 
    var i int32;
    var k int32;
    var iVar int32;
    var pCut *int32;
    var Place int32;
    var Flag int32;
    assert(!p.pPars.fCutMin && p.pGia.vMapping == nil, "!p.pPars.fCutMin && p.pGia.vMapping == nil");
    vMapping = Vec_IntAlloc( Gia_ManObjNum(p.pGia) + int32(p.pPars.Edge) + int32(p.pPars.Area) * 2 );
    Vec_IntFill( vMapping, Gia_ManObjNum(p.pGia), 0 );
    if p.pPars.nFastEdges {
        vPacking = Vec_IntAlloc( 1000 );
        Vec_IntPush( vPacking, 0 );
    }
    vPointed = Vec_BitStart( Gia_ManObjNum(p.pGia) );
    // inlining of macro Gia_ManForEachAndId
    for  i = 0; (i < p.pGia.nObjs); i++  {                                     if !Gia_ObjIsAnd(Gia_ManObj(p.pGia, i)) {} else {
        if !(Of_ObjRefNum(p, i) != 0) {
            continue; }
        assert(!Gia_ObjIsBuf(Gia_ManObj(p.pGia,i)), "!Gia_ObjIsBuf(Gia_ManObj(p.pGia,i))");
        pCut = Of_ObjCutBestP( p, i );
        Vec_IntWriteEntry( vMapping, i, Vec_IntSize(vMapping) );
        Vec_IntPush( vMapping, Of_CutSize(pCut) );
        // inlining of macro Of_CutForEachVar
        for  k = 0; k < Of_CutSize(pCut) && (iVar = Of_CutVar(pCut,k)); k++  {
            Vec_IntPush( vMapping, iVar );
        }
        Vec_IntPush( vMapping, i );
        if vPacking == nil || Vec_BitEntry(vPointed, i) {
            continue; }
        Place = Vec_IntSize( vPacking );
        Vec_IntPush( vPacking, 0 );
        Vec_IntPush( vPacking, i );
        // inlining of macro Of_CutForEachVarFlag
        for  k = 0; k < Of_CutSize(pCut) && (iVar = Of_CutVar(pCut,k)) && ((Flag = Of_CutFlag(pCut,k)), 1); k++  {
            if Flag != 0 {
                Vec_IntPush( vPacking, iVar );
                Vec_BitWriteEntry( vPointed, iVar, 1 );
            }
        }
        Vec_IntAddToEntry( vPacking, Place, Vec_IntSize(vPacking)-Place-1 );
        Vec_IntAddToEntry( vPacking, 0, 1 );
    }
    }
    assert(Vec_IntCap(vMapping) == 16 || Vec_IntSize(vMapping) == Vec_IntCap(vMapping), "Vec_IntCap(vMapping) == 16 || Vec_IntSize(vMapping) == Vec_IntCap(vMapping)");
    p.pGia.vMapping = vMapping;
    p.pGia.vPacking = vPacking;
    Vec_BitFree( vPointed );
    return p.pGia;
}
func Of_ManPerformMapping(pGia *Gia_Man_t, pPars *Jf_Par_t) *Gia_Man_t {
    var pNew *Gia_Man_t = nil;
    var pCls *Gia_Man_t;
    var p *Of_Man_t; var i, Id int32;
    if Gia_ManHasChoices(pGia) {
        pPars.fCoarsen = 0, pPars.fCutMin = 1; } 
    if pPars.fCoarsen {
        pCls = Gia_ManDupMuxes(pGia, pPars.nCoarseLimit);
    } else {
        pCls = pGia;
    }
    p = Of_StoCreate( pCls, pPars );
    if pPars.fVerbose && pPars.fCoarsen {
        fmt.Printf("Initial ");  Gia_ManPrintMuxStats( pGia );  fmt.Printf("\n");
        fmt.Printf("Derived ");  Gia_ManPrintMuxStats( pCls );  fmt.Printf("\n");
    }
    Of_ManPrintInit( p );
    Of_ManComputeCuts( p );
    Of_ManPrintQuit( p );

    // inlining of macro Gia_ManForEachCiId
    for  i = 0; (i < Vec_IntSize(p.pGia.vCis)) && ((Id) = Gia_ObjId(p.pGia, Gia_ManCi(p.pGia, i))); i++  {
        var Time int32  = Of_Flt2Int(p.pGia.vInArrs ? Abc_MaxFloat(0.0, Vec_FltEntry(p.pGia.vInArrs, i)) : 0.0);
        Of_ObjSetDelay1( p, Id, Time );
        Of_ObjSetDelay2( p, Id, Time );
    }

    if p.pPars.nFastEdges {
        p.pPars.nRounds = 1;
        for  p.Iter = 0; p.Iter < p.pPars.nRounds; p.Iter++  {
            if p.Iter == 0 {
                Of_ManComputeForwardDircon1( p );
                Of_ManComputeBackwardDircon1( p );
                Of_ManPrintStats( p, "Delay" );
            
            } else {
                Of_ManComputeForwardDircon1( p );
                Of_ManComputeBackwardDircon1( p );
                Of_ManPrintStats( p, "Flow " );
            }
        }
    
    } else {
        for  p.Iter = 0; p.Iter < p.pPars.nRounds; p.Iter++  {
            if p.Iter == 0 {
                Of_ManComputeForward1( p );
                Of_ManComputeBackward1( p );
                Of_ManPrintStats( p, "Delay" );
            
            } else {
                Of_ManComputeForward1( p );
                Of_ManComputeBackward1( p );
                Of_ManPrintStats( p, "Flow " );
            }
        }
        for  ; p.Iter < p.pPars.nRounds + p.pPars.nRoundsEla; p.Iter++  {
            if p.Iter < p.pPars.nRounds + p.pPars.nRoundsEla - 1 {
                Of_ManComputeForward2( p );
                Of_ManComputeBackward3( p );
                Of_ManPrintStats( p, "Area " );
            
            } else {
                Of_ManComputeForward1( p );
                Of_ManComputeBackward3( p );
                Of_ManPrintStats( p, "Area " );
            }
        }
    }

    pNew = Of_ManDeriveMapping( p );
    Gia_ManMappingVerify( pNew );
    if pNew.vPacking {
        Gia_ManConvertPackingToEdges( pNew ); }
    //Of_ManPrintCuts( p );
    Of_StoDelete( p );
    if pCls != pGia {
        Gia_ManStop( pCls ); }
    if pNew == nil {
        return Gia_ManDup( pGia ); }
    return pNew;
}

////////////////////////////////////////////////////////////////////////
///                       END OF FILE                                ///
////////////////////////////////////////////////////////////////////////


//ABC_NAMESPACE_IMPL_END

