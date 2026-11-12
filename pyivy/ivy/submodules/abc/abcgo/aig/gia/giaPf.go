package gia

import (
	"fmt"
	"os"
)

/**CFile****************************************************************

  FileName    [giaNf.c]

  SystemName  [ABC: Logic synthesis and verification system.]

  PackageName [Scalable AIG package.]

  Synopsis    [Standard-cell mapper.]

  Author      [Alan Mishchenko]
  
  Affiliation [UC Berkeley]

  Date        [Ver. 1.0. Started - June 20, 2005.]

  Revision    [$Id: giaNf.c,v 1.00 2005/06/20 00:00:00 alanmi Exp $]

***********************************************************************/

//#include <float.h>
//#include "gia.h"
//#include "misc/st/st.h"
//#include "map/mio/mio.h"
//#include "misc/util/utilTruth.h"
//#include "misc/extra/extra.h"
//#include "base/main/main.h"
//#include "misc/vec/vecMem.h"
//#include "misc/vec/vecWec.h"
//#include "opt/dau/dau.h"

//ABC_NAMESPACE_IMPL_START

////////////////////////////////////////////////////////////////////////
///                        DECLARATIONS                              ///
////////////////////////////////////////////////////////////////////////

const PF_LEAF_MAX =  6
const PF_CUT_MAX =  32
const PF_NO_LEAF =  31
const PF_NO_FUNC =  0x3FFFFFF
#define PF_INFINITY FLT_MAX

    
type Pf_Cut_t struct {
    Sign            word;           // signature
    Delay             int32;          // delay
    Flow           float32;           // flow
    iFunc uint32; /* C bitfield: 26 bits; review packing and truncation */   // function (PF_NO_FUNC)
    Useless uint32; /* C bitfield: 1 bits; review packing and truncation */   // function
    nLeaves uint32; /* C bitfield: 5 bits; review packing and truncation */   // leaf number (PF_NO_LEAF)
    pLeaves [PF_LEAF_MAX+1]int32; // leaves
}
    
type Pf_Mat_t struct {
    fCompl uint32; /* C bitfield: 8 bits; review packing and truncation */   // complemented
    Phase uint32; /* C bitfield: 6 bits; review packing and truncation */   // match phase
    Perm uint32; /* C bitfield: 18 bits; review packing and truncation */   // match permutation
}
    
type Pf_Obj_t struct {
    Area           float32;
    Gate uint32; /* C bitfield: 7 bits; review packing and truncation */   // gate
    nLeaves uint32; /* C bitfield: 3 bits; review packing and truncation */   // fanin count  
    nRefs uint32; /* C bitfield: 22 bits; review packing and truncation */   // ref count
    pLeaves [6]int32;     // leaf literals
}
    
type Pf_Man_t struct {
    // user data
    pGia      *Gia_Man_t;           // derived manager
    pPars       *Jf_Par_t;          // parameters
    // matching
    vTtMem      *Vec_Mem_t;         // truth tables
    vTt2Match      *Vec_Wec_t;      // matches for truth tables
    pCells     *Mio_Cell_t;         // library gates
    nCells             int32;         // library gate count
    // cut data
    pPfObjs       *Pf_Obj_t;        // best cuts
    vPages       Vec_Ptr_t;         // cut memory
    vCutSets       Vec_Int_t;       // cut offsets
    vCutFlows       Vec_Flt_t;      // temporary cut area
    vCutDelays       Vec_Int_t;     // temporary cut delay
    iCur             int32;           // current position
    Iter             int32;           // mapping iterations
    fUseEla             int32;        // use exact area
    nInvs             int32;          // the inverter count
    InvDelay           float32;       // inverter delay
    InvArea           float32;        // inverter area 
    // statistics
    clkStart         abctime;       // starting time
    CutCount [6]float64;    // cut counts
    nCutUseAll             int32;     // objects with useful cuts
}

func Pf_Mat2Int(Mat Pf_Mat_t) int32 { union { var x int32; var y Pf_Mat_t; }
 v; v.y = Mat; return v.x;           }
func Pf_Int2Mat(Int int32) Pf_Mat_t { union { var x int32; var y Pf_Mat_t; }
 v; v.x = Int; return v.y;           }

func Pf_ManObj(p *Pf_Man_t, i int32) *Pf_Obj_t { return p.pPfObjs + i;                                           }
func Pf_ManCell(p *Pf_Man_t, i int32) *Mio_Cell_t { return p.pCells + i;                                            }
func Pf_ManCutSet(p *Pf_Man_t, i int32) *int32 { return (int *)Vec_PtrEntry(&p.vPages, i >> 16) + (i & 0xFFFF);  }
func Pf_ObjCutSetId(p *Pf_Man_t, i int32) int32 { return Vec_IntEntry( &p.vCutSets, i );                          }
func Pf_ObjCutSet(p *Pf_Man_t, i int32) *int32 { return Pf_ManCutSet(p, Pf_ObjCutSetId(p, i));                    }
func Pf_ObjHasCuts(p *Pf_Man_t, i int32) int32 { return int32((Vec_IntEntry(&p.vCutSets, i) > 0));                 }
func Pf_ObjCutUseless(p *Pf_Man_t, TruthId int32) int32 { return int32((TruthId >= Vec_WecSize(p.vTt2Match)));              } 

func Pf_ObjCutFlow(p *Pf_Man_t, i int32) float32 { return Vec_FltEntry(&p.vCutFlows, i);                           } 
func Pf_ObjCutDelay(p *Pf_Man_t, i int32) int32 { return Vec_IntEntry(&p.vCutDelays, i);                          } 
func Pf_ObjSetCutFlow(p *Pf_Man_t, i int32, a float32) { Vec_FltWriteEntry(&p.vCutFlows, i, a);                          } 
func Pf_ObjSetCutDelay(p *Pf_Man_t, i int32, d int32) { Vec_IntWriteEntry(&p.vCutDelays, i, d);                         } 

func Pf_CutSize(pCut *int32) int32 { return pCut[0] & PF_NO_LEAF;                                     }
func Pf_CutFunc(pCut *int32) int32 { return (uint32(pCut[0]) >> 5);                                 }
func Pf_CutLeaves(pCut *int32) *int32 { return pCut + 1;                                                 }
func Pf_CutSetBoth(n int32, f int32) int32 { return n | (f << 5);                                             }
func Pf_CutIsTriv(pCut *int32, i int32) int32 { return Pf_CutSize(pCut) == 1 && pCut[1] == i;                    } 
func Pf_CutHandle(pCutSet *int32, pCut *int32) int32 { assert(pCut > pCutSet, "pCut > pCutSet"); return pCut - pCutSet;                 } 
func Pf_CutFromHandle(pCutSet *int32, h int32) *int32 { assert(h > 0, "h > 0"); return pCutSet + h;                             }
func Pf_CutConfLit(Conf int32, i int32) int32 { return 15 & (Conf >> (i << 2));                                  }
func Pf_CutConfVar(Conf int32, i int32) int32 { return Abc_Lit2Var( Pf_CutConfLit(Conf, i) );                    }
func Pf_CutConfC(Conf int32, i int32) int32 { return Abc_LitIsCompl( Pf_CutConfLit(Conf, i) );                 }

// Original macro:
// #define Pf_SetForEachCut( pList, pCut, i )         for ( i = 0, pCut = pList + 1; i < pList[0]; i++, pCut += Pf_CutSize(pCut) + 1 )

// Original macro:
// #define Pf_ObjForEachCut( pCuts, i, nCuts )        for ( i = 0, i < nCuts; i++ )

// Original macro:
// #define Pf_CutForEachLit( pCut, Conf, iLit, i )    for ( i = 0; i < Pf_CutSize(pCut) && (iLit = Abc_Lit2LitV(Pf_CutLeaves(pCut), Pf_CutConfLit(Conf, i))); i++ )

// Original macro:
// #define Pf_CutForEachVar( pCut, Conf, iVar, c, i ) for ( i = 0; i < Pf_CutSize(pCut) && (iVar = Pf_CutLeaves(pCut)[Pf_CutConfVar(Conf, i)]) && ((c = Pf_CutConfC(Conf, i)), 1); i++ )


////////////////////////////////////////////////////////////////////////
///                     FUNCTION DEFINITIONS                         ///
////////////////////////////////////////////////////////////////////////

/**Function*************************************************************

  Synopsis    []

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Pf_StoCreateGateAdd(pMan *Pf_Man_t, uTruth word, pFans *int32, nFans int32, CellId int32) {
    var vArray *Vec_Int_t;
    var Mat Pf_Mat_t  = Pf_Int2Mat(0);
    var i int32;
    var GateId int32;
    var Entry int32;
    var fCompl int32 = int32((uTruth & 1));
    var uFunc word;
    if fCompl {
        uFunc = ~uTruth;
    } else {
        uFunc = uTruth;
    }
    var iFunc int32  = Vec_MemHashInsert( pMan.vTtMem, &uFunc );
    if iFunc == Vec_WecSize(pMan.vTt2Match) {
        Vec_WecPushLevel( pMan.vTt2Match ); }
    vArray = Vec_WecEntry( pMan.vTt2Match, iFunc );
    Mat.fCompl = fCompl;
    assert(nFans < 7, "nFans < 7");
    for  i = 0; i < nFans; i++  {
        Mat.Perm  |= uint32((Abc_Lit2Var(pFans[i]) << (3*i)));
        Mat.Phase |= uint32((Abc_LitIsCompl(pFans[i]) << i));
    }
    // check if the same one exists
    // inlining of macro Vec_IntForEachEntryDouble
    for  i = 0; (i+1 < Vec_IntSize(vArray)) && (((GateId) = Vec_IntEntry(vArray, i)), 1) && (((Entry) = Vec_IntEntry(vArray, i+1)), 1); i += 2  {
        if GateId == CellId && Pf_Int2Mat(Entry).Phase == Mat.Phase {
            break; }
    }
    if i == Vec_IntSize(vArray) {
        Vec_IntPush( vArray, CellId );
        Vec_IntPush( vArray, Pf_Mat2Int(Mat) );
    }
}
func Pf_StoCreateGate(pMan *Pf_Man_t, pCell *Mio_Cell_t, pComp **int32, pPerm **int32, pnPerms *int32) {
    var Perm [PF_LEAF_MAX]int32;
    var Perm1 *int32;
    var Perm2 *int32;
    var nPerms int32  = pnPerms[pCell.nFanins];
    var nMints int32  = (1 << pCell.nFanins);
    var tCur, tTemp1, tTemp2 word;
    var i, p, c int32;
    for  i = 0; i < int32(pCell.nFanins); i++  {
        Perm[i] = Abc_Var2Lit( i, 0 );
    }
    tCur = tTemp1 = pCell.uTruth;
    for  p = 0; p < nPerms; p++  {
        tTemp2 = tCur;
        for  c = 0; c < nMints; c++  {
            Pf_StoCreateGateAdd( pMan, tCur, Perm, pCell.nFanins, pCell.Id );
            // update
            tCur  = Abc_Tt6Flip( tCur, pComp[pCell.nFanins][c] );
            Perm1 = Perm + pComp[pCell.nFanins][c];
            *Perm1 = Abc_LitNot( *Perm1 );
        }
        assert(tTemp2 == tCur, "tTemp2 == tCur");
        // update
        tCur = Abc_Tt6SwapAdjacent( tCur, pPerm[pCell.nFanins][p] );
        Perm1 = Perm + pPerm[pCell.nFanins][p];
        Perm2 = Perm1 + 1;
        // inlining of macro ABC_SWAP
        { var t int32  = *Perm1; *Perm1 = *Perm2; *Perm2 = t; };
    }
    assert(tTemp1 == tCur, "tTemp1 == tCur");
}
func Pf_StoDeriveMatches(p *Pf_Man_t, fVerbose int32) {
//    abctime clk = Abc_Clock();
    var pComp [7]*int32;
    var pPerm [7]*int32;
    var nPerms [7]int32;
    var i int32;
    for  i = 2; i <= 6; i++  {
        pComp[i] = Extra_GreyCodeSchedule( i );
    }
    for  i = 2; i <= 6; i++  {
        pPerm[i] = Extra_PermSchedule( i );
    }
    for  i = 2; i <= 6; i++  {
        nPerms[i] = Extra_Factorial( i );
    }
    p.pCells = Mio_CollectRootsNewDefault( 6, &p.nCells, fVerbose );
    for  i = 4; i < p.nCells; i++  {
        Pf_StoCreateGate( p, p.pCells + i, pComp, pPerm, nPerms );
    }
    for  i = 2; i <= 6; i++  {
        // inlining of macro ABC_FREE
        ((pComp[i]) ? (free((char *) (pComp[i])), (pComp[i]) = 0) : 0);
    }
    for  i = 2; i <= 6; i++  {
        // inlining of macro ABC_FREE
        ((pPerm[i]) ? (free((char *) (pPerm[i])), (pPerm[i]) = 0) : 0);
    }
//    Abc_PrintTime( 1, "Time", Abc_Clock() - clk );
}
func Pf_StoPrintOne(p *Pf_Man_t, Count int32, t int32, i int32, GateId int32, Mat Pf_Mat_t) {
    var pC *Mio_Cell_t  = p.pCells + GateId;
    var pTruth *word  = Vec_MemReadEntry(p.vTtMem, t);
    var k int32;
    var nSuppSize int32 = Abc_TtSupportSize(pTruth, 6);
    fmt.Printf("%6d : ", Count);
    fmt.Printf("%6d : ", t);
    fmt.Printf("%6d : ", i);
    fmt.Printf("Gate %16s  ", pC.pName);
    fmt.Printf("Area =%8.2f  ", pC.Area);
    fmt.Printf("In = %d   ", pC.nFanins);
    if Mat.fCompl {
        fmt.Printf(" compl ");
    } else {
        fmt.Printf("       "); }
    for  k = 0; k < int32(pC.nFanins); k++  {
        var fComplF int32  = (Mat.Phase >> k) & 1;
        var iFanin int32   = (Mat.Perm >> (3*k)) & 7;
        fmt.Printf("%c", 'a' + iFanin - fComplF * ('a' - 'A'));
    }
    fmt.Printf("  ");
    Dau_DsdPrintFromTruth( pTruth, nSuppSize );
}
func Pf_StoPrint(p *Pf_Man_t, fVerbose int32) {
    var t, i, GateId, Entry, Count int32;
    for  t = 2; t < Vec_WecSize(p.vTt2Match); t++  {
        var vArr *Vec_Int_t  = Vec_WecEntry( p.vTt2Match, t );
        // inlining of macro Vec_IntForEachEntryDouble
        for  i = 0; (i+1 < Vec_IntSize(vArr)) && (((GateId) = Vec_IntEntry(vArr, i)), 1) && (((Entry) = Vec_IntEntry(vArr, i+1)), 1); i += 2  {
            Count++;
            if !(fVerbose != 0) {
                continue; }
            if t < 10 {
                Pf_StoPrintOne( p, Count, t, i/2, GateId, Pf_Int2Mat(Entry) ); }
        }
    }
    fmt.Printf("Gates = %d.  Truths = %d.  Matches = %d.\n", p.nCells, Vec_MemEntryNum(p.vTtMem), Count);
}
/*
void Pf_ManPrepareLibraryTest()
{
    int fVerbose = 0;
    abctime clk = Abc_Clock();
    Pf_Man_t * p;
    p = Pf_StoCreate( NULL, NULL, fVerbose );
    Pf_StoPrint( p, fVerbose );
    Pf_StoDelete(p);
    Abc_PrintTime( 1, "Time", Abc_Clock() - clk );
}
*/



/**Function*************************************************************

  Synopsis    []

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Pf_StoCreate(pGia *Gia_Man_t, pPars *Jf_Par_t) *Pf_Man_t {
    extern void Mf_ManSetFlowRefs( Gia_Man_t * p, Vec_Int_t * vRefs );
    var p *Pf_Man_t;
    var vFlowRefs *Vec_Int_t;
    assert(pPars.nCutNum > 1  && pPars.nCutNum <= PF_CUT_MAX, "pPars.nCutNum > 1  && pPars.nCutNum <= PF_CUT_MAX");
    assert(pPars.nLutSize > 1 && pPars.nLutSize <= PF_LEAF_MAX, "pPars.nLutSize > 1 && pPars.nLutSize <= PF_LEAF_MAX");
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
    p = &Pf_Man_t{};
    p.clkStart = Abc_Clock();
    p.pGia     = pGia;
    p.pPars    = pPars;
    // inlining of macro ABC_CALLOC
    p.pPfObjs  = ((Pf_Obj_t *) calloc((size_t)(Gia_ManObjNum(pGia)), sizeof(Pf_Obj_t)));
    p.iCur     = 2;
    // other
    Vec_PtrGrow( &p.vPages, 256 );                                    // cut memory
    Vec_IntFill( &p.vCutSets,  Gia_ManObjNum(pGia), 0 );              // cut offsets
    Vec_FltFill( &p.vCutFlows, Gia_ManObjNum(pGia), 0 );              // cut area
    Vec_IntFill( &p.vCutDelays,Gia_ManObjNum(pGia), 0 );              // cut delay
    // matching
    p.vTtMem    = Vec_MemAllocForTT( 6, 0 );          
    p.vTt2Match = Vec_WecAlloc( 1000 ); 
    Vec_WecPushLevel( p.vTt2Match );
    Vec_WecPushLevel( p.vTt2Match );
    assert(Vec_WecSize(p.vTt2Match) == Vec_MemEntryNum(p.vTtMem), "Vec_WecSize(p.vTt2Match) == Vec_MemEntryNum(p.vTtMem)");
    Pf_StoDeriveMatches( p, 0 );//pPars->fVerbose );
    p.InvDelay = p.pCells[3].Delays[0];
    p.InvArea  = p.pCells[3].Area;
    //Pf_ObjMatchD(p, 0, 0)->Gate = 0;
    //Pf_ObjMatchD(p, 0, 1)->Gate = 1;
    // prepare cuts
    return p;
}
func Pf_StoDelete(p *Pf_Man_t) {
    Vec_PtrFreeData( &p.vPages );
    // inlining of macro ABC_FREE
    ((p.vPages.pArray) ? (free((char *) (p.vPages.pArray)), (p.vPages.pArray) = 0) : 0);
    // inlining of macro ABC_FREE
    ((p.vCutSets.pArray) ? (free((char *) (p.vCutSets.pArray)), (p.vCutSets.pArray) = 0) : 0);
    // inlining of macro ABC_FREE
    ((p.vCutFlows.pArray) ? (free((char *) (p.vCutFlows.pArray)), (p.vCutFlows.pArray) = 0) : 0);
    // inlining of macro ABC_FREE
    ((p.vCutDelays.pArray) ? (free((char *) (p.vCutDelays.pArray)), (p.vCutDelays.pArray) = 0) : 0);
    // inlining of macro ABC_FREE
    ((p.pPfObjs) ? (free((char *) (p.pPfObjs)), (p.pPfObjs) = 0) : 0);
    // matching
    Vec_WecFree( p.vTt2Match );
    Vec_MemHashFree( p.vTtMem );
    Vec_MemFree( p.vTtMem );
    // inlining of macro ABC_FREE
    ((p.pCells) ? (free((char *) (p.pCells)), (p.pCells) = 0) : 0);
    // inlining of macro ABC_FREE
    ((p) ? (free((char *) (p)), (p) = 0) : 0);
}




/**Function*************************************************************

  Synopsis    []

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Pf_CutComputeTruth6(p *Pf_Man_t, pCut0 *Pf_Cut_t, pCut1 *Pf_Cut_t, fCompl0 int32, fCompl1 int32, pCutR *Pf_Cut_t, fIsXor int32) int32 {
//    extern int Pf_ManTruthCanonicize( word * t, int nVars );
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
    pCutR.Useless = Pf_ObjCutUseless( p, truthId );
    assert(int32(pCutR.nLeaves) <= nOldSupp, "int32(pCutR.nLeaves) <= nOldSupp");
    return int32(pCutR.nLeaves) < nOldSupp;
}
func Pf_CutComputeTruthMux6(p *Pf_Man_t, pCut0 *Pf_Cut_t, pCut1 *Pf_Cut_t, pCutC *Pf_Cut_t, fCompl0 int32, fCompl1 int32, fComplC int32, pCutR *Pf_Cut_t) int32 {
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
    pCutR.Useless = Pf_ObjCutUseless( p, truthId );
    assert(int32(pCutR.nLeaves) <= nOldSupp, "int32(pCutR.nLeaves) <= nOldSupp");
    return int32(pCutR.nLeaves) < nOldSupp;
}


/**Function*************************************************************

  Synopsis    []

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Pf_CutCountBits(i word) int32 {
    i = i - ((i >> 1) & 0x5555555555555555);
    i = (i & 0x3333333333333333) + ((i >> 2) & 0x3333333333333333);
    i = ((i + (i >> 4)) & 0x0F0F0F0F0F0F0F0F);
    return (i*(0x0101010101010101))>>56;
}
func Pf_CutGetSign(pLeaves *int32, nLeaves int32) word {
    var Sign word; var i int32; 
    for  i = 0; i < nLeaves; i++  {
        Sign |= ((word)1) << (pLeaves[i] & 0x3F);
    }
    return Sign;
}
func Pf_CutCreateUnit(p *Pf_Cut_t, i int32) int32 {
    p.Delay      = 0;
    p.Flow       = 0;
    p.iFunc      = 2;
    p.nLeaves    = 1;
    p.pLeaves[0] = i;
    p.Sign       = ((word)1) << (i & 0x3F);
    return 1;
}
func Pf_Cutprintf(p *Pf_Man_t, pCut *Pf_Cut_t) {
    var i int32;
    var nDigits int32 = Abc_Base10Log(Gia_ManObjNum(p.pGia)); 
    fmt.Printf("%d  {", pCut.nLeaves);
    for  i = 0; i < int32(pCut.nLeaves); i++  {
        fmt.Printf(" %*d", nDigits, pCut.pLeaves[i]);
    }
    for  ; i < int32(p.pPars.nLutSize); i++  {
        fmt.Printf(" %*s", nDigits, " ");
    }
    fmt.Printf("  }   Useless = %d. D = %4d  A = %9.4f  F = %6d  ", pCut.Useless, pCut.Delay, pCut.Flow, pCut.iFunc);
    if p.vTtMem {
        Dau_DsdPrintFromTruth( Vec_MemReadEntry(p.vTtMem, Abc_Lit2Var(pCut.iFunc)), pCut.nLeaves );
    } else {
        fmt.Printf("\n"); }
}
func Pf_ManPrepareCuts(pCuts *Pf_Cut_t, p *Pf_Man_t, iObj int32, fAddUnit int32) int32 {
    if Pf_ObjHasCuts(p, iObj) != 0 {
        var pMfCut *Pf_Cut_t  = pCuts;
        var i int32;
        var pCut *int32;
        var pList *int32 = Pf_ObjCutSet(p, iObj);
        // inlining of macro Pf_SetForEachCut
        for  i, pCut = 0, pList + 1; i < pList[0]; i++, pCut += Pf_CutSize(pCut) + 1  {
            pMfCut.Delay   = 0;
            pMfCut.Flow    = 0;
            pMfCut.iFunc   = Pf_CutFunc( pCut );
            pMfCut.nLeaves = Pf_CutSize( pCut );
            pMfCut.Sign    = Pf_CutGetSign( pCut+1, Pf_CutSize(pCut) );
            pMfCut.Useless = Pf_ObjCutUseless( p, Abc_Lit2Var(pMfCut.iFunc) );
            memcpy( pMfCut.pLeaves, pCut+1, sizeof(int) * Pf_CutSize(pCut) );
            pMfCut++;
        }
        if (fAddUnit != 0) && pCuts.nLeaves > 1 {
            return pList[0] + Pf_CutCreateUnit( pMfCut, iObj ); }
        return pList[0];
    }
    return Pf_CutCreateUnit( pCuts, iObj );
}
func Pf_ManSaveCuts(p *Pf_Man_t, pCuts **Pf_Cut_t, nCuts int32, fUseful int32) int32 {
    var i int32;
    var pPlace *int32;
    var iCur int32;
    var nInts int32 = 1;
    var nCutsNew int32;
    for  i = 0; i < nCuts; i++  {
        if !(fUseful != 0) || !pCuts[i].Useless {
            nInts += pCuts[i].nLeaves + 1, nCutsNew++; }
    }
    if (p.iCur & 0xFFFF) + nInts > 0xFFFF {
        p.iCur = ((p.iCur >> 16) + 1) << 16; }
    if Vec_PtrSize(&p.vPages) == (p.iCur >> 16) {
        // inlining of macro ABC_ALLOC
        Vec_PtrPush( &p.vPages, ((int *) malloc(sizeof(int) * (size_t)((1<<16)))) ); }
    iCur = p.iCur; p.iCur += nInts;
    pPlace = Pf_ManCutSet( p, iCur );
    *pPlace++ = nCutsNew;
    for  i = 0; i < nCuts; i++  {
        if !(fUseful != 0) || !pCuts[i].Useless {
            *pPlace++ = Pf_CutSetBoth( pCuts[i].nLeaves, pCuts[i].iFunc );
            memcpy( pPlace, pCuts[i].pLeaves, sizeof(int) * pCuts[i].nLeaves );
            pPlace += pCuts[i].nLeaves;
        }
    }
    return iCur;
}
func Pf_ManCountUseful(pCuts **Pf_Cut_t, nCuts int32) int32 {
    var i, Count int32;
    for  i = 0; i < nCuts; i++  {
        Count += !pCuts[i].Useless;
    }
    return Count;
}
func Pf_ManCountMatches(p *Pf_Man_t, pCuts **Pf_Cut_t, nCuts int32) int32 {
    var i, Count int32;
    for  i = 0; i < nCuts; i++  {
        if !pCuts[i].Useless {
            Count += Vec_IntSize(Vec_WecEntry(p.vTt2Match, Abc_Lit2Var(pCuts[i].iFunc))) / 2; }
    }
    return Count;
}

/**Function*************************************************************

  Synopsis    [Check correctness of cuts.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
static inline int Pf_CutCheck( Pf_Cut_t * pBase, Pf_Cut_t * pCut ) // check if pCut is contained in pBase
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
func Pf_SetCheckArray(ppCuts **Pf_Cut_t, nCuts int32) int32 {
    var pCut0, pCut1 *Pf_Cut_t; 
    var i, k, m, n, Value int32;
    assert(nCuts > 0, "nCuts > 0");
    for  i = 0; i < nCuts; i++  {
        pCut0 = ppCuts[i];
        assert(pCut0.nLeaves <= PF_LEAF_MAX, "pCut0.nLeaves <= PF_LEAF_MAX");
        assert(pCut0.Sign == Pf_CutGetSign(pCut0.pLeaves, pCut0.nLeaves), "pCut0.Sign == Pf_CutGetSign(pCut0.pLeaves, pCut0.nLeaves)");
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
            Value = Pf_CutCheck( pCut0, pCut1 );
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
func Pf_CutMergeOrder(pCut0 *Pf_Cut_t, pCut1 *Pf_Cut_t, pCut *Pf_Cut_t, nLutSize int32) int32 { 
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
        pCut.iFunc = PF_NO_FUNC;
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
    pCut.iFunc = PF_NO_FUNC;
    pCut.Sign = pCut0.Sign | pCut1.Sign;
    return 1;

FlushCut1:
    if c + nSize1 > nLutSize + k { return 0; }
    for  k < nSize1  {
        pC[c++] = pC1[k++];
    }
    pCut.nLeaves = c;
    pCut.iFunc = PF_NO_FUNC;
    pCut.Sign = pCut0.Sign | pCut1.Sign;
    return 1;
}
func Pf_CutMergeOrderMux(pCut0 *Pf_Cut_t, pCut1 *Pf_Cut_t, pCut2 *Pf_Cut_t, pCut *Pf_Cut_t, nLutSize int32) int32 { 
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
    pCut.iFunc = PF_NO_FUNC;
    pCut.Sign = pCut0.Sign | pCut1.Sign | pCut2.Sign;
    return 1;
}
static inline int Pf_SetCutIsContainedOrder( Pf_Cut_t * pBase, Pf_Cut_t * pCut ) // check if pCut is contained in pBase
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
func Pf_SetLastCutIsContained(pCuts **Pf_Cut_t, nCuts int32) int32 {
    var i int32;
    for  i = 0; i < nCuts; i++  {
        if pCuts[i].nLeaves <= pCuts[nCuts].nLeaves && (pCuts[i].Sign & pCuts[nCuts].Sign) == pCuts[i].Sign && Pf_SetCutIsContainedOrder(pCuts[nCuts], pCuts[i]) {
            return 1; }
    }
    return 0;
}
func Pf_SetLastCutContainsArea(pCuts **Pf_Cut_t, nCuts int32) int32 {
    var i, k, fChanges int32;
    for  i = 0; i < nCuts; i++  {
        if pCuts[nCuts].nLeaves < pCuts[i].nLeaves && (pCuts[nCuts].Sign & pCuts[i].Sign) == pCuts[nCuts].Sign && Pf_SetCutIsContainedOrder(pCuts[i], pCuts[nCuts]) {
            pCuts[i].nLeaves = PF_NO_LEAF, fChanges = 1; }
    }
    if !(fChanges != 0) {
        return nCuts; }
    for  i = k = 0; i <= nCuts; i++  {
        if pCuts[i].nLeaves == PF_NO_LEAF {
            continue; }
        if k < i {
            // inlining of macro ABC_SWAP
             var t *Pf_Cut_t  = pCuts[k]; pCuts[k] = pCuts[i]; pCuts[i] = t; };
        k++;
    }
    return k - 1;
}
func Pf_CutCompareArea(pCut0 *Pf_Cut_t, pCut1 *Pf_Cut_t) int32 {
    if pCut0.Useless < pCut1.Useless {  return -1; }
    if pCut0.Useless > pCut1.Useless {  return  1; }
    if pCut0.Flow    < pCut1.Flow {  return -1; }
    if pCut0.Flow    > pCut1.Flow {  return  1; }
    if pCut0.Delay   < pCut1.Delay {  return -1; }
    if pCut0.Delay   > pCut1.Delay {  return  1; }
    if pCut0.nLeaves < pCut1.nLeaves {  return -1; }
    if pCut0.nLeaves > pCut1.nLeaves {  return  1; }
    return 0;
}
func Pf_SetSortByArea(pCuts **Pf_Cut_t, nCuts int32) {
    var i int32;
    for  i = nCuts; i > 0; i--  {
        if Pf_CutCompareArea(pCuts[i - 1], pCuts[i]) < 0 {//!= 1 )
            return; }
        // inlining of macro ABC_SWAP
        { var t *Pf_Cut_t  = pCuts[i - 1]; pCuts[i - 1] = pCuts[i]; pCuts[i] = t; };
    }
}
func Pf_SetAddCut(pCuts **Pf_Cut_t, nCuts int32, nCutNum int32) int32 {
    if nCuts == 0 {
        return 1; }
    nCuts = Pf_SetLastCutContainsArea(pCuts, nCuts);
    Pf_SetSortByArea( pCuts, nCuts );
    return Abc_MinInt( nCuts + 1, nCutNum - 1 );
}
func Pf_CutArea(p *Pf_Man_t, nLeaves int32) int32 {
    if nLeaves < 2 {
        return 0; }
    return nLeaves + p.pPars.nAreaTuner;
}
func Pf_CutParams(p *Pf_Man_t, pCut *Pf_Cut_t, nGiaRefs int32) {
    var i int32;
    var nLeaves int32 = pCut.nLeaves; 
    assert(nLeaves <= p.pPars.nLutSize, "nLeaves <= p.pPars.nLutSize");
    pCut.Delay = 0;
    pCut.Flow  = 0;
    for  i = 0; i < nLeaves; i++  {
        pCut.Delay = Abc_MaxInt( pCut.Delay, Pf_ObjCutDelay(p, pCut.pLeaves[i]) );
        pCut.Flow += Pf_ObjCutFlow(p, pCut.pLeaves[i]);
    }
    pCut.Delay += int32((nLeaves > 1));
    pCut.Flow = (pCut.Flow + Pf_CutArea(p, nLeaves)) / (nGiaRefs ? nGiaRefs : 1);
}
func Pf_ObjMergeOrder(p *Pf_Man_t, iObj int32) {
    var pCuts0 [PF_CUT_MAX]Pf_Cut_t;
    var pCuts1 [PF_CUT_MAX]Pf_Cut_t;
    var pCuts [PF_CUT_MAX]Pf_Cut_t;
    var pCutsR [PF_CUT_MAX]*Pf_Cut_t;
    var pObj *Gia_Obj_t  = Gia_ManObj(p.pGia, iObj);
    var nGiaRefs int32  = 2*Gia_ObjRefNumId(p.pGia, iObj);
    var nLutSize int32  = p.pPars.nLutSize;
    var nCutNum int32   = p.pPars.nCutNum;
    var nCuts0 int32    = Pf_ManPrepareCuts(pCuts0, p, Gia_ObjFaninId0(pObj, iObj), 1);
    var nCuts1 int32    = Pf_ManPrepareCuts(pCuts1, p, Gia_ObjFaninId1(pObj, iObj), 1);
    var fComp0 int32    = Gia_ObjFaninC0(pObj);
    var fComp1 int32    = Gia_ObjFaninC1(pObj);
    var iSibl int32     = Gia_ObjSibl(p.pGia, iObj);
    var pCut0 *Pf_Cut_t;
    var pCut1 *Pf_Cut_t;
    var pCut0Lim *Pf_Cut_t = pCuts0 + nCuts0;
    var pCut1Lim *Pf_Cut_t = pCuts1 + nCuts1;
    var i, nCutsUse, nCutsR int32;
    assert(!Gia_ObjIsBuf(pObj), "!Gia_ObjIsBuf(pObj)");
    for  i = 0; i < nCutNum; i++  {
        pCutsR[i] = pCuts + i;
    }
    if iSibl != 0 {
        var pCuts2 [PF_CUT_MAX]Pf_Cut_t;
        var pObjE *Gia_Obj_t  = Gia_ObjSiblObj(p.pGia, iObj);
        var fCompE int32  = Gia_ObjPhase(pObj) ^ Gia_ObjPhase(pObjE);
        var nCuts2 int32  = Pf_ManPrepareCuts(pCuts2, p, iSibl, 0);
        var pCut2 *Pf_Cut_t;
        var pCut2Lim *Pf_Cut_t = pCuts2 + nCuts2;
        for  pCut2 = pCuts2; pCut2 < pCut2Lim; pCut2++  {
            *pCutsR[nCutsR] = *pCut2;
            pCutsR[nCutsR].iFunc = Abc_LitNotCond( pCutsR[nCutsR].iFunc, fCompE );
            Pf_CutParams( p, pCutsR[nCutsR], nGiaRefs );
            nCutsR = Pf_SetAddCut( pCutsR, nCutsR, nCutNum );
        }
    }
    if Gia_ObjIsMuxId(p.pGia, iObj) {
        var pCuts2 [PF_CUT_MAX]Pf_Cut_t;
        var nCuts2 int32   = Pf_ManPrepareCuts(pCuts2, p, Gia_ObjFaninId2(p.pGia, iObj), 1);
        var fComp2 int32   = Gia_ObjFaninC2(p.pGia, pObj);
        var pCut2 *Pf_Cut_t;
        var pCut2Lim *Pf_Cut_t = pCuts2 + nCuts2;
        p.CutCount[0] += nCuts0 * nCuts1 * nCuts2;
        for  pCut0 = pCuts0; pCut0 < pCut0Lim; pCut0++  {
        for  pCut1 = pCuts1; pCut1 < pCut1Lim; pCut1++  {
        for  pCut2 = pCuts2; pCut2 < pCut2Lim; pCut2++  {
            if Pf_CutCountBits(pCut0.Sign | pCut1.Sign | pCut2.Sign) > nLutSize {
                continue; }
            p.CutCount[1]++; 
            if !(Pf_CutMergeOrderMux(pCut0, pCut1, pCut2, pCutsR[nCutsR], nLutSize) != 0) {
                continue; }
            if Pf_SetLastCutIsContained(pCutsR, nCutsR) != 0 {
                continue; }
            p.CutCount[2]++;
            if Pf_CutComputeTruthMux6(p, pCut0, pCut1, pCut2, fComp0, fComp1, fComp2, pCutsR[nCutsR]) != 0 {
                pCutsR[nCutsR].Sign = Pf_CutGetSign(pCutsR[nCutsR].pLeaves, pCutsR[nCutsR].nLeaves); }
            Pf_CutParams( p, pCutsR[nCutsR], nGiaRefs );
            nCutsR = Pf_SetAddCut( pCutsR, nCutsR, nCutNum );
        }
        }
        }
    
    } else {
        var fIsXor int32  = Gia_ObjIsXor(pObj);
        p.CutCount[0] += nCuts0 * nCuts1;
        for  pCut0 = pCuts0; pCut0 < pCut0Lim; pCut0++  {
        for  pCut1 = pCuts1; pCut1 < pCut1Lim; pCut1++  {
            if int32((pCut0.nLeaves + pCut1.nLeaves)) > nLutSize && Pf_CutCountBits(pCut0.Sign | pCut1.Sign) > nLutSize {
                continue; }
            p.CutCount[1]++; 
            if !(Pf_CutMergeOrder(pCut0, pCut1, pCutsR[nCutsR], nLutSize) != 0) {
                continue; }
            if Pf_SetLastCutIsContained(pCutsR, nCutsR) != 0 {
                continue; }
            p.CutCount[2]++;
            if Pf_CutComputeTruth6(p, pCut0, pCut1, fComp0, fComp1, pCutsR[nCutsR], fIsXor) != 0 {
                pCutsR[nCutsR].Sign = Pf_CutGetSign(pCutsR[nCutsR].pLeaves, pCutsR[nCutsR].nLeaves); }
            Pf_CutParams( p, pCutsR[nCutsR], nGiaRefs );
            nCutsR = Pf_SetAddCut( pCutsR, nCutsR, nCutNum );
        }
        }
    }
    // debug printout
    if false {
//    if ( iObj % 10000 == 0 )
//    if ( iObj == 1090 )
    
        fmt.Printf("*** Obj = %d  Useful = %d\n", iObj, Pf_ManCountUseful(pCutsR, nCutsR));
        for  i = 0; i < nCutsR; i++  {
            Pf_Cutprintf( p, pCutsR[i] );
        }
        fmt.Printf("\n");
    } 
    // verify
    assert(nCutsR > 0 && nCutsR < nCutNum, "nCutsR > 0 && nCutsR < nCutNum");
//    assert( Pf_SetCheckArray(pCutsR, nCutsR) );
    // store the cutset
    Pf_ObjSetCutFlow( p, iObj, pCutsR[0].Flow );
    Pf_ObjSetCutDelay( p, iObj, pCutsR[0].Delay );
    *Vec_IntEntryP(&p.vCutSets, iObj) = Pf_ManSaveCuts(p, pCutsR, nCutsR, 0);
    p.CutCount[3] += nCutsR;
    nCutsUse = Pf_ManCountUseful(pCutsR, nCutsR);
    p.CutCount[4] += nCutsUse;
    p.nCutUseAll  += nCutsUse == nCutsR;
    p.CutCount[5] += Pf_ManCountMatches(p, pCutsR, nCutsR);
}
func Pf_ManComputeCuts(p *Pf_Man_t) {
    var pObj *Gia_Obj_t; var i, iFanin int32;
    // inlining of macro Gia_ManForEachAnd
    for  i = 0; (i < p.pGia.nObjs) && ((pObj) = Gia_ManObj(p.pGia, i)); i++  {      if !Gia_ObjIsAnd(pObj) {} else
        if Gia_ObjIsBuf(pObj) {
            iFanin = Gia_ObjFaninId0(pObj, i);
            Pf_ObjSetCutFlow( p, i,  Pf_ObjCutFlow(p, iFanin) );
            Pf_ObjSetCutDelay( p, i, Pf_ObjCutDelay(p, iFanin) );
        
        } else {
            Pf_ObjMergeOrder( p, i ); }
    }
}




/**Function*************************************************************

  Synopsis    []

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Pf_ManPrintStats(p *Pf_Man_t, pTitle string) {
    if !p.pPars.fVerbose {
        return; }
    fmt.Printf("%s :  ", pTitle);
    fmt.Printf("Delay =%8.2f  ", float32(p.pPars.MapDelay));
    fmt.Printf("Area =%12.2f  ", p.pPars.MapArea);
    fmt.Printf("Gate =%6d  ", int32(p.pPars.Area));
    fmt.Printf("Inv =%6d  ", int32(p.nInvs));
    fmt.Printf("Edge =%7d  ", int32(p.pPars.Edge));
    Abc_PrintTime( 1, "Time", Abc_Clock() - p.clkStart );
    fflush( os.Stdout );
}
func Pf_ManPrintInit(p *Pf_Man_t) {
    var nChoices int32;
    if !p.pPars.fVerbose {
        return; }
    fmt.Printf("LutSize = %d  ", p.pPars.nLutSize);
    fmt.Printf("CutNum = %d  ", p.pPars.nCutNum);
    fmt.Printf("Iter = %d  ", p.pPars.nRounds + p.pPars.nRoundsEla);
    fmt.Printf("Coarse = %d   ", p.pPars.fCoarsen);
    fmt.Printf("Cells = %d  ", p.nCells);
    fmt.Printf("Funcs = %d  ", Vec_MemEntryNum(p.vTtMem));
    fmt.Printf("Matches = %d  ", Vec_WecSizeSize(p.vTt2Match)/2);
    nChoices = Gia_ManChoiceNum( p.pGia );
    if nChoices != 0 {
    fmt.Printf("Choices = %d  ", nChoices); }
    fmt.Printf("\n");
    fmt.Printf("Computing cuts...\r");
    fflush( os.Stdout );
}
func Pf_ManPrintQuit(p *Pf_Man_t) {
    var MemGia float32    = Gia_ManMemory(p.pGia) / (1<<20);
    var MemMan float32    =(1.0 * sizeof(Pf_Obj_t) + 3.0 * sizeof(int)) * Gia_ManObjNum(p.pGia) / (1<<20);
    var MemCuts float32   = 1.0 * sizeof(int) * (1 << 16) * Vec_PtrSize(&p.vPages) / (1<<20);
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
    fmt.Printf("Use = %.0f (%.1f)  ", p.CutCount[4], 1.0*p.CutCount[4]/Gia_ManAndNum(p.pGia));
    fmt.Printf("Mat = %.0f (%.1f)  ", p.CutCount[5], 1.0*p.CutCount[5]/Gia_ManAndNum(p.pGia));
//    printf( "Equ = %d (%.2f %%)  ",     p->nCutUseAll,  100.0*p->nCutUseAll /p->CutCount[0] );
    fmt.Printf("\n");
    fmt.Printf("Gia = %.2f MB  ", MemGia);
    fmt.Printf("Man = %.2f MB  ", MemMan); 
    fmt.Printf("Cut = %.2f MB   ", MemCuts);
    fmt.Printf("TT = %.2f MB  ", MemTt); 
    fmt.Printf("Total = %.2f MB   ", MemGia + MemMan + MemCuts + MemTt); 
//    printf( "\n" );
    Abc_PrintTime( 1, "Time", Abc_Clock() - p.clkStart );
    fflush( os.Stdout );
}


/**Function*************************************************************

  Synopsis    []

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
/*
void Pf_ManSetMapRefsGate( Pf_Man_t * p, int iObj, float Required, Pf_Mat_t * pM )
{
    int k, iVar, fCompl;
    Mio_Cell_t * pCell = Pf_ManCell( p, pM->Gate );
    int * pCut = Pf_CutFromHandle( Pf_ObjCutSet(p, iObj), pM->CutH );
    Pf_CutForEachVar( pCut, pM->Conf, iVar, fCompl, k )
    {
        Pf_ObjMapRefInc( p, iVar, fCompl );
        Pf_ObjUpdateRequired( p, iVar, fCompl, Required - pCell->Delays[k] );
    }
    assert( Pf_CutSize(pCut) == (int)pCell->nFanins );
    // update global stats
    p->pPars->MapArea += pCell->Area;
    p->pPars->Edge += Pf_CutSize(pCut);
    p->pPars->Area++;
    // update status of the gate
    assert( pM->fBest == 0 );
    pM->fBest = 1;
}
int Pf_ManSetMapRefs( Pf_Man_t * p )
{
    float Coef = 1.0 / (1.0 + (p->Iter + 1) * (p->Iter + 1));
    float * pFlowRefs = Vec_FltArray( &p->vFlowRefs );
    int * pMapRefs = Vec_IntArray( &p->vMapRefs );
    float Epsilon = p->pPars->Epsilon;
    int nLits = 2*Gia_ManObjNum(p->pGia);
    int i, c, Id, nRefs[2];
    Pf_Mat_t * pD, * pA, * pM;
    Pf_Mat_t * pDs[2], * pAs[2], * pMs[2];
    Gia_Obj_t * pObj;
    float Required = 0, Requireds[2];
    // check references
    assert( !p->fUseEla );
    memset( pMapRefs, 0, sizeof(int) * nLits );
    Vec_FltFill( &p->vRequired, nLits, PF_INFINITY );
//    for ( i = 0; i < Gia_ManObjNum(p->pGia); i++ )
//        assert( !Pf_ObjMapRefNum(p, i, 0) && !Pf_ObjMapRefNum(p, i, 1) );
    // compute delay
    p->pPars->MapDelay = 0;
    Gia_ManForEachCo( p->pGia, pObj, i )
    {
        Required = Pf_ObjMatchD( p, Gia_ObjFaninId0p(p->pGia, pObj), Gia_ObjFaninC0(pObj) )->D;
        if ( Required == PF_INFINITY )
        {
            Pf_ManCutMatchprintf( p, Gia_ObjFaninId0p(p->pGia, pObj), Gia_ObjFaninC0(pObj), Pf_ObjMatchD( p, Gia_ObjFaninId0p(p->pGia, pObj), Gia_ObjFaninC0(pObj) ) );
        }
        p->pPars->MapDelay = Abc_MaxFloat( p->pPars->MapDelay, Required );
    }
    // check delay target
    if ( p->pPars->MapDelayTarget == -1 && p->pPars->nRelaxRatio )
        p->pPars->MapDelayTarget = (int)((float)p->pPars->MapDelay * (100.0 + p->pPars->nRelaxRatio) / 100.0);
    if ( p->pPars->MapDelayTarget != -1 )
    {
        if ( p->pPars->MapDelay < p->pPars->MapDelayTarget + Epsilon )
            p->pPars->MapDelay = p->pPars->MapDelayTarget;
        else if ( p->pPars->nRelaxRatio == 0 )
            Abc_Print( 0, "Relaxing user-specified delay target from %.2f to %.2f.\n", p->pPars->MapDelayTarget, p->pPars->MapDelay );
    }
    // set required times
    Gia_ManForEachCo( p->pGia, pObj, i )
    {
        Required = Pf_ObjMatchD( p, Gia_ObjFaninId0p(p->pGia, pObj), Gia_ObjFaninC0(pObj) )->D;
        Required = p->pPars->fDoAverage ? Required * (100.0 + p->pPars->nRelaxRatio) / 100.0 : p->pPars->MapDelay;
        Pf_ObjUpdateRequired( p, Gia_ObjFaninId0p(p->pGia, pObj), Gia_ObjFaninC0(pObj), Required );
        Pf_ObjMapRefInc( p, Gia_ObjFaninId0p(p->pGia, pObj), Gia_ObjFaninC0(pObj));
    }
    // compute area and edges
    p->nInvs = 0;
    p->pPars->MapArea = 0; 
    p->pPars->Area = p->pPars->Edge = 0;
    Gia_ManForEachAndReverse( p->pGia, pObj, i )
    {
        if ( Gia_ObjIsBuf(pObj) )
        {
            if ( Pf_ObjMapRefNum(p, i, 1) )
            {
                Pf_ObjMapRefInc( p, i, 0 );
                Pf_ObjUpdateRequired( p, i, 0, Pf_ObjRequired(p, i, 1) - p->InvDelay );
                p->pPars->MapArea += p->InvArea;
                p->pPars->Edge++;
                p->pPars->Area++;
                p->nInvs++;
            }
            Pf_ObjUpdateRequired( p, Gia_ObjFaninId0(pObj, i), Gia_ObjFaninC0(pObj), Pf_ObjRequired(p, i, 0) );
            Pf_ObjMapRefInc( p, Gia_ObjFaninId0(pObj, i), Gia_ObjFaninC0(pObj));
            continue;
        }
        // skip if this node is not used
        for ( c = 0; c < 2; c++ )
        {
            nRefs[c] = Pf_ObjMapRefNum(p, i, c);

            //if ( Pf_ObjMatchD( p, i, c )->fCompl )
            //    printf( "Match D of node %d has inv in phase %d.\n", i, c );
            //if ( Pf_ObjMatchA( p, i, c )->fCompl )
            //    printf( "Match A of node %d has inv in phase %d.\n", i, c );
        }
        if ( !nRefs[0] && !nRefs[1] )
            continue;

        // consider two cases
        if ( nRefs[0] && nRefs[1] )
        {
            // find best matches for both phases
            for ( c = 0; c < 2; c++ )
            {
                Requireds[c] = Pf_ObjRequired( p, i, c );
                //assert( Requireds[c] < PF_INFINITY );
                pDs[c] = Pf_ObjMatchD( p, i, c );
                pAs[c] = Pf_ObjMatchA( p, i, c );
                pMs[c] = (pAs[c]->D < Requireds[c] + Epsilon) ? pAs[c] : pDs[c];
            }
            // swap complemented matches
            if ( pMs[0]->fCompl && pMs[1]->fCompl )
            {
                pMs[0]->fCompl = pMs[1]->fCompl = 0;
                ABC_SWAP( Pf_Mat_t *, pMs[0], pMs[1] );
            }
            // check if intervers are involved
            if ( !pMs[0]->fCompl && !pMs[1]->fCompl )
            {
                // no inverters
                for ( c = 0; c < 2; c++ )
                    Pf_ManSetMapRefsGate( p, i, Requireds[c], pMs[c] );
            }
            else 
            {
                // one interver
                assert( !pMs[0]->fCompl || !pMs[1]->fCompl );
                c = pMs[1]->fCompl;
                assert( pMs[c]->fCompl && !pMs[!c]->fCompl );
                //printf( "Using inverter at node %d in phase %d\n", i, c );

                // update this phase phase
                pM = pMs[c];
                pM->fBest = 1;
                Required = Requireds[c];

                // update opposite phase
                Pf_ObjMapRefInc( p, i, !c );
                Pf_ObjUpdateRequired( p, i, !c, Required - p->InvDelay );

                // select oppositve phase
                Required = Pf_ObjRequired( p, i, !c );
                //assert( Required < PF_INFINITY );
                pD = Pf_ObjMatchD( p, i, !c );
                pA = Pf_ObjMatchA( p, i, !c );
                pM = (pA->D < Required + Epsilon) ? pA : pD;
                assert( !pM->fCompl );

                // account for the inverter
                p->pPars->MapArea += p->InvArea;
                p->pPars->Edge++;
                p->pPars->Area++;
                p->nInvs++;

                // create gate
                Pf_ManSetMapRefsGate( p, i, Required, pM );
            }
        }
        else
        {
            c = (int)(nRefs[1] > 0);
            assert( nRefs[c] && !nRefs[!c] );
            // consider this phase
            Required = Pf_ObjRequired( p, i, c );
            //assert( Required < PF_INFINITY );
            pD = Pf_ObjMatchD( p, i, c );
            pA = Pf_ObjMatchA( p, i, c );
            pM = (pA->D < Required + Epsilon) ? pA : pD;

            if ( pM->fCompl ) // use inverter
            {
                p->nInvs++;
                //printf( "Using inverter at node %d in phase %d\n", i, c );
                pM->fBest = 1;
                // update opposite phase
                Pf_ObjMapRefInc( p, i, !c );
                Pf_ObjUpdateRequired( p, i, !c, Required - p->InvDelay );
                // select oppositve phase
                Required = Pf_ObjRequired( p, i, !c );
                //assert( Required < PF_INFINITY );
                pD = Pf_ObjMatchD( p, i, !c );
                pA = Pf_ObjMatchA( p, i, !c );
                pM = (pA->D < Required + Epsilon) ? pA : pD;
                assert( !pM->fCompl );

                // account for the inverter
                p->pPars->MapArea += p->InvArea;
                p->pPars->Edge++;
                p->pPars->Area++;
            }

            // create gate
            Pf_ManSetMapRefsGate( p, i, Required, pM );
        }


        // the result of this:
        // - only one phase can be implemented as inverter of the other phase
        // - required times are propagated correctly
        // - references are set correctly
    }
    Gia_ManForEachCiId( p->pGia, Id, i )
        if ( Pf_ObjMapRefNum(p, Id, 1) )
        {
            Pf_ObjMapRefInc( p, Id, 0 );
            Pf_ObjUpdateRequired( p, Id, 0, Required - p->InvDelay );
            p->pPars->MapArea += p->InvArea;
            p->pPars->Edge++;
            p->pPars->Area++;
            p->nInvs++;
        }
    // blend references
    for ( i = 0; i < nLits; i++ )
//        pFlowRefs[i] = Abc_MaxFloat(1.0, pMapRefs[i]);
        pFlowRefs[i] = Abc_MaxFloat(1.0, Coef * pFlowRefs[i] + (1.0 - Coef) * Abc_MaxFloat(1, pMapRefs[i]));
//        pFlowRefs[i] = 0.2 * pFlowRefs[i] + 0.8 * Abc_MaxFloat(1, pMapRefs[i]);
//    memset( pMapRefs, 0, sizeof(int) * nLits );
    return p->pPars->Area;
}
Gia_Man_t * Pf_ManDeriveMapping( Pf_Man_t * p )
{
    Vec_Int_t * vMapping;
    Pf_Mat_t * pM;
    int i, k, c, Id, iLit, * pCut;
    assert( p->pGia->vCellMapping == NULL );
    vMapping = Vec_IntAlloc( 2*Gia_ManObjNum(p->pGia) + (int)p->pPars->Edge + (int)p->pPars->Area * 2 );
    Vec_IntFill( vMapping, 2*Gia_ManObjNum(p->pGia), 0 );
    // create CI inverters
    Gia_ManForEachCiId( p->pGia, Id, i )
    if ( Pf_ObjMapRefNum(p, Id, 1) )
        Vec_IntWriteEntry( vMapping, Abc_Var2Lit(Id, 1), -1 );
    // create internal nodes
    Gia_ManForEachAndId( p->pGia, i )
    {
        Gia_Obj_t * pObj = Gia_ManObj(p->pGia, i);
        if ( Gia_ObjIsBuf(pObj) )
        {
            if ( Pf_ObjMapRefNum(p, i, 1) )
                Vec_IntWriteEntry( vMapping, Abc_Var2Lit(i, 1), -1 );
            Vec_IntWriteEntry( vMapping, Abc_Var2Lit(i, 0), -2 );
            continue;
        }
        for ( c = 0; c < 2; c++ )
        if ( Pf_ObjMapRefNum(p, i, c) )
        {
    //        printf( "Using %d %d\n", i, c );
            pM = Pf_ObjMatchBest( p, i, c );
            // remember inverter
            if ( pM->fCompl )
            {
                Vec_IntWriteEntry( vMapping, Abc_Var2Lit(i, c), -1 );
                continue;
            }
    //        Pf_ManCutMatchprintf( p, i, c, pM );
            pCut = Pf_CutFromHandle( Pf_ObjCutSet(p, i), pM->CutH );
            // create mapping
            Vec_IntWriteEntry( vMapping, Abc_Var2Lit(i, c), Vec_IntSize(vMapping) );
            Vec_IntPush( vMapping, Pf_CutSize(pCut) );
            Pf_CutForEachLit( pCut, pM->Conf, iLit, k )
                Vec_IntPush( vMapping, iLit );
            Vec_IntPush( vMapping, pM->Gate );
        }
    }
//    assert( Vec_IntCap(vMapping) == 16 || Vec_IntSize(vMapping) == Vec_IntCap(vMapping) );
    p->pGia->vCellMapping = vMapping;
    return p->pGia;
}
*/


/**Function*************************************************************

  Synopsis    [Technology mappping.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Pf_ManComputeMapping(p *Pf_Man_t) {
}

/**Function*************************************************************

  Synopsis    [Technology mappping.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Pf_ManSetDefaultPars(pPars *Jf_Par_t) {
    memset( pPars, 0, sizeof(Jf_Par_t) );
    pPars.nLutSize     =  6;
    pPars.nCutNum      = 16;
    pPars.nProcNum     =  0;
    pPars.nRounds      =  3;
    pPars.nRoundsEla   =  0;
    pPars.nRelaxRatio  =  0;
    pPars.nCoarseLimit =  3;
    pPars.nAreaTuner   =  1;
    pPars.nVerbLimit   =  5;
    pPars.DelayTarget  = -1;
    pPars.fAreaOnly    =  0;
    pPars.fOptEdge     =  1; 
    pPars.fCoarsen     =  0;
    pPars.fCutMin      =  1;
    pPars.fGenCnf      =  0;
    pPars.fPureAig     =  0;
    pPars.fVerbose     =  0;
    pPars.fVeryVerbose =  0;
    pPars.nLutSizeMax  =  PF_LEAF_MAX;
    pPars.nCutNumMax   =  PF_CUT_MAX;
    pPars.MapDelayTarget = -1;
    pPars.Epsilon      = float32(0.01);
}
func Pf_ManPerformMapping(pGia *Gia_Man_t, pPars *Jf_Par_t) *Gia_Man_t {
    var pNew *Gia_Man_t = nil;
    var pCls *Gia_Man_t;
    var p *Pf_Man_t; 
    if Gia_ManHasChoices(pGia) {
        pPars.fCoarsen = 0; } 
    if pPars.fCoarsen {
        pCls = Gia_ManDupMuxes(pGia, pPars.nCoarseLimit);
    } else {
        pCls = pGia;
    }
    p = Pf_StoCreate( pCls, pPars );
//    if ( pPars->fVeryVerbose )
        Pf_StoPrint( p, 1 );
    if pPars.fVerbose && pPars.fCoarsen {
        fmt.Printf("Initial ");  Gia_ManPrintMuxStats( pGia );  fmt.Printf("\n");
        fmt.Printf("Derived ");  Gia_ManPrintMuxStats( pCls );  fmt.Printf("\n");
    }
    Pf_ManPrintInit( p );
    Pf_ManComputeCuts( p );
    Pf_ManPrintQuit( p );
/*
    Gia_ManForEachCiId( p->pGia, Id, i )
        Pf_ObjPrepareCi( p, Id );
    for ( p->Iter = 0; p->Iter < p->pPars->nRounds; p->Iter++ )
    {
        Pf_ManComputeMapping( p );
        //Pf_ManSetMapRefs( p );
        Pf_ManPrintStats( p, p->Iter ? "Area " : "Delay" );
    }
    p->fUseEla = 1;
    for ( ; p->Iter < p->pPars->nRounds + pPars->nRoundsEla; p->Iter++ )
    {
        Pf_ManComputeMapping( p );
        //Pf_ManUpdateStats( p );
        Pf_ManPrintStats( p, "Ela  " );
    }
*/
    pNew = nil; //Pf_ManDeriveMapping( p );
//    Gia_ManMappingVerify( pNew );
    Pf_StoDelete( p );
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

