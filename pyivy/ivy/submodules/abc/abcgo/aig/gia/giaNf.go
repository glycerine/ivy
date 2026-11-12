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
//#include "misc/util/utilNam.h"
//#include "map/scl/sclCon.h"

//ABC_NAMESPACE_IMPL_START

////////////////////////////////////////////////////////////////////////
///                        DECLARATIONS                              ///
////////////////////////////////////////////////////////////////////////

type Nf_Mat_t struct {
    Gate uint32; /* C bitfield: 20 bits; review packing and truncation */   // gate
    CutH uint32; /* C bitfield: 10 bits; review packing and truncation */   // cut handle
    fCompl uint32; /* C bitfield: 1 bits; review packing and truncation */   // complemented
    fBest uint32; /* C bitfield: 1 bits; review packing and truncation */   // best cut
    Cfg        Nf_Cfg_t;            // input literals
    D             int32;              // delay
    F           float32;              // area 
}
    
type Nf_Obj_t struct {
    M [2][2]Nf_Mat_t;         // del/area (2x)
}
    
type Nf_Man_t struct {
    // user data
    pGia      *Gia_Man_t;           // derived manager
    pPars       *Jf_Par_t;          // parameters
    // matching
    vTtMem      *Vec_Mem_t;         // truth tables
    vTt2Match      *Vec_Wec_t;      // matches for truth tables
    pCells    *Mio_Cell2_t;         // library gates
    nCells             int32;         // library gate count
    // cut data
    pNfObjs       *Nf_Obj_t;        // best cuts
    vPages       Vec_Ptr_t;         // cut memory
    vCutSets       Vec_Int_t;       // cut offsets
    vMapRefs       Vec_Int_t;       // mapping refs   (2x)
    vFlowRefs       Vec_Flt_t;      // flow refs      (2x)
    vRequired       Vec_Int_t;      // required times (2x)
    vCutFlows       Vec_Flt_t;      // temporary cut area
    vCutDelays       Vec_Int_t;     // temporary cut delay
    vBackup       Vec_Int_t;        // backup literals
    iCur             int32;           // current position
    Iter             int32;           // mapping iterations
    fUseEla             int32;        // use exact area
    nInvs             int32;          // the inverter count
    InvDelayI             int32;      // inverter delay
    InvAreaW            word;       // inverter delay
    InvAreaF           float32;       // inverter area 
    // statistics
    clkStart         abctime;       // starting time
    CutCount [6]float64;    // cut counts
    nCutUseAll             int32;     // objects with useful cuts
}

func Nf_ManObj(p *Nf_Man_t, i int32) *Nf_Obj_t { return p.pNfObjs + i;                                           }
func Nf_ManCell(p *Nf_Man_t, i int32) *Mio_Cell2_t { return p.pCells + i;                                            }
func Nf_ManCutSet(p *Nf_Man_t, i int32) *int32 { return (int *)Vec_PtrEntry(&p.vPages, i >> 16) + (i & 0xFFFF);  }
func Nf_ObjCutSetId(p *Nf_Man_t, i int32) int32 { return Vec_IntEntry( &p.vCutSets, i );                          }
func Nf_ObjCutSet(p *Nf_Man_t, i int32) *int32 { return Nf_ManCutSet(p, Nf_ObjCutSetId(p, i));                    }
func Nf_ObjHasCuts(p *Nf_Man_t, i int32) int32 { return int32((Vec_IntEntry(&p.vCutSets, i) > 0));                 }
func Nf_ObjCutBest(p *Nf_Man_t, i int32) *int32 { return nil;                                                     } 
func Nf_ObjCutUseless(p *Nf_Man_t, TruthId int32) int32 { return int32((TruthId >= Vec_WecSize(p.vTt2Match)));              } 

func Nf_ObjCutFlow(p *Nf_Man_t, i int32) float32 { return Vec_FltEntry(&p.vCutFlows, i);                           } 
func Nf_ObjCutDelay(p *Nf_Man_t, i int32) int32 { return Vec_IntEntry(&p.vCutDelays, i);                          } 
func Nf_ObjSetCutFlow(p *Nf_Man_t, i int32, a float32) { Vec_FltWriteEntry(&p.vCutFlows, i, a);                          } 
func Nf_ObjSetCutDelay(p *Nf_Man_t, i int32, d int32) { Vec_IntWriteEntry(&p.vCutDelays, i, d);                         } 

func Nf_ObjMapRefNum(p *Nf_Man_t, i int32, c int32) int32 { return Vec_IntEntry(&p.vMapRefs, Abc_Var2Lit(i,c));             }
func Nf_ObjMapRefInc(p *Nf_Man_t, i int32, c int32) int32 { return (*Vec_IntEntryP(&p.vMapRefs, Abc_Var2Lit(i,c)))++;       }
func Nf_ObjMapRefDec(p *Nf_Man_t, i int32, c int32) int32 { return --(*Vec_IntEntryP(&p.vMapRefs, Abc_Var2Lit(i,c)));       }
func Nf_ObjFlowRefs(p *Nf_Man_t, i int32, c int32) float32 { return Vec_FltEntry(&p.vFlowRefs, Abc_Var2Lit(i,c));            }
func Nf_ObjRequired(p *Nf_Man_t, i int32, c int32) int32 { return Vec_IntEntry(&p.vRequired, Abc_Var2Lit(i,c));            }
func Nf_ObjSetRequired(p *Nf_Man_t, i int32, c int32, f int32) { Vec_IntWriteEntry(&p.vRequired, Abc_Var2Lit(i,c), f);           }
func Nf_ObjUpdateRequired(p *Nf_Man_t, i int32, c int32, f int32) { if Nf_ObjRequired(p, i, c) > f { Nf_ObjSetRequired(p, i, c, f); }  }

func Nf_ObjMatchD(p *Nf_Man_t, i int32, c int32) *Nf_Mat_t { return &Nf_ManObj(p, i).M[c][0];                                }
func Nf_ObjMatchA(p *Nf_Man_t, i int32, c int32) *Nf_Mat_t { return &Nf_ManObj(p, i).M[c][1];                                }





 
func Nf_CutHandle(pCutSet *int32, pCut *int32) int32 { assert(pCut > pCutSet, "pCut > pCutSet"); return pCut - pCutSet;                 } 
func Nf_CutFromHandle(pCutSet *int32, h int32) *int32 { assert(h > 0, "h > 0"); return pCutSet + h;                             }




// Original macro:
// #define Nf_SetForEachCut( pList, pCut, i )                   for ( i = 0, pCut = pList + 1; i < pList[0]; i++, pCut += Nf_CutSize(pCut) + 1 )

// Original macro:
// #define Nf_CutForEachVarCompl( pCut, Cfg, iVar, fCompl, i )  for ( i = 0; i < Nf_CutSize(pCut) && (iVar = Nf_CutLeaves(pCut)[Nf_CfgVar(Cfg, i)]) && ((fCompl = Nf_CfgCompl(Cfg, i)), 1); i++ )

// Original macro:
// #define Nf_CfgForEachVarCompl( Cfg, Size, iVar, fCompl, i )  for ( i = 0; i < Size && ((iVar = Nf_CfgVar(Cfg, i)), 1) && ((fCompl = Nf_CfgCompl(Cfg, i)), 1); i++ )


////////////////////////////////////////////////////////////////////////
///                     FUNCTION DEFINITIONS                         ///
////////////////////////////////////////////////////////////////////////

/**Function*************************************************************

  Synopsis    []

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Nf_StoCellIsDominated(pCell *Mio_Cell2_t, pFans *int32, pProf *int32) int32 {
    var k int32;
    if pCell.AreaF + NF_EPSILON < Abc_Int2Float(pProf[0]) {
        return 0; }
    for  k = 0; k < int32(pCell.nFanins); k++  {
        if pCell.iDelays[Abc_Lit2Var(pFans[k])] < pProf[k+1] {
            return 0; }
    }
    return 1; // pCell is dominated
}
func Nf_StoCreateGateAdd(vTtMem *Vec_Mem_t, vTt2Match *Vec_Wec_t, pCell *Mio_Cell2_t, uTruth word, pFans *int32, nFans int32, vProfs *Vec_Wec_t, vStore *Vec_Int_t, fPinFilter int32, fPinPerm int32, fPinQuick int32) {
    var vArray *Vec_Int_t;
    var vArrayProfs *Vec_Int_t = nil;
    var i int32;
    var k int32;
    var GateId int32;
    var Entry int32;
    var fCompl int32 = int32((uTruth & 1));
    var uFunc word;
    if fCompl {
        uFunc = ~uTruth;
    } else {
        uFunc = uTruth;
    }
    var iFunc int32  = Vec_MemHashInsert( vTtMem, &uFunc );
    var Mat Nf_Cfg_t  = Nf_Int2Cfg(0);
    // get match array
    if iFunc == Vec_WecSize(vTt2Match) {
        Vec_WecPushLevel( vTt2Match ); }
    vArray = Vec_WecEntry( vTt2Match, iFunc );
    // create match
    Mat.fCompl = fCompl;
    assert(nFans == int32(pCell.nFanins), "nFans == int32(pCell.nFanins)");
    for  i = 0; i < nFans; i++  {
        Mat.Perm  |= uint32((i << (Abc_Lit2Var(pFans[i]) << 2)));
        Mat.Phase |= uint32((Abc_LitIsCompl(pFans[i]) << Abc_Lit2Var(pFans[i])));
    }
    // check other profiles
    if fPinFilter != 0 {
        // get profile array
        assert(Vec_WecSize(vTt2Match) == Vec_WecSize(vProfs), "Vec_WecSize(vTt2Match) == Vec_WecSize(vProfs)");
        if iFunc == Vec_WecSize(vProfs) {
            Vec_WecPushLevel( vProfs ); }
        vArrayProfs = Vec_WecEntry( vProfs, iFunc );
        assert(Vec_IntSize(vArray) == 2 * Vec_IntSize(vArrayProfs), "Vec_IntSize(vArray) == 2 * Vec_IntSize(vArrayProfs)");
        // skip dominated matches
        // inlining of macro Vec_IntForEachEntryDouble
        for  i = 0; (i+1 < Vec_IntSize(vArray)) && (((GateId) = Vec_IntEntry(vArray, i)), 1) && (((Entry) = Vec_IntEntry(vArray, i+1)), 1); i += 2  {
            if Nf_Int2Cfg(Entry).Phase == Mat.Phase && Nf_Int2Cfg(Entry).fCompl == Mat.fCompl {
                var Offset int32  = Vec_IntEntry(vArrayProfs, i/2);
                var pProf *int32  = Vec_IntEntryP(vStore, Offset);
                if Nf_StoCellIsDominated(pCell, pFans, pProf) != 0 {
                    return; }
            }
        }
    }
    // check pin permutation
    if !(fPinPerm != 0) { // do not use  pin-permutation (improves delay when pin-delays differ)
    
        if fPinQuick != 0 { // reduce the number of matches agressively
        
            // inlining of macro Vec_IntForEachEntryDouble
            for  i = 0; (i+1 < Vec_IntSize(vArray)) && (((GateId) = Vec_IntEntry(vArray, i)), 1) && (((Entry) = Vec_IntEntry(vArray, i+1)), 1); i += 2  {
                if GateId == int32(pCell.Id) && Abc_TtBitCount8[Nf_Int2Cfg(Entry).Phase] == Abc_TtBitCount8[Mat.Phase] {
                    return; }
            }
        
        } else { // reduce the number of matches less agressively
        
            // inlining of macro Vec_IntForEachEntryDouble
            for  i = 0; (i+1 < Vec_IntSize(vArray)) && (((GateId) = Vec_IntEntry(vArray, i)), 1) && (((Entry) = Vec_IntEntry(vArray, i+1)), 1); i += 2  {
                if GateId == int32(pCell.Id) && Nf_Int2Cfg(Entry).Phase == Mat.Phase {
                    return; }
            }
        }
    }
    // save data and profile
    Vec_IntPush( vArray, pCell.Id );
    Vec_IntPush( vArray, Nf_Cfg2Int(Mat) );
    // add delay profile
    if fPinFilter != 0 {
        Vec_IntPush( vArrayProfs, Vec_IntSize(vStore) );
        Vec_IntPush( vStore, Abc_Float2Int(pCell.AreaF) );
        for  k = 0; k < nFans; k++  {
            Vec_IntPush( vStore, pCell.iDelays[Abc_Lit2Var(pFans[k])] );
        }
    }
}
func Nf_StoCreateGateMaches(vTtMem *Vec_Mem_t, vTt2Match *Vec_Wec_t, pCell *Mio_Cell2_t, pComp **int32, pPerm **int32, pnPerms *int32, vProfs *Vec_Wec_t, vStore *Vec_Int_t, fPinFilter int32, fPinPerm int32, fPinQuick int32) {
    var Perm [NF_LEAF_MAX]int32;
    var Perm1 *int32;
    var Perm2 *int32;
    var nPerms int32  = pnPerms[pCell.nFanins];
    var nMints int32  = (1 << pCell.nFanins);
    var tCur, tTemp1, tTemp2 word;
    var i, p, c int32;
    assert(pCell.nFanins <= 6, "pCell.nFanins <= 6");
    for  i = 0; i < int32(pCell.nFanins); i++  {
        Perm[i] = Abc_Var2Lit( i, 0 );
    }
    tCur = tTemp1 = pCell.uTruth;
    for  p = 0; p < nPerms; p++  {
        tTemp2 = tCur;
        for  c = 0; c < nMints; c++  {
            Nf_StoCreateGateAdd( vTtMem, vTt2Match, pCell, tCur, Perm, pCell.nFanins, vProfs, vStore, fPinFilter, fPinPerm, fPinQuick );
            // update
            tCur  = Abc_Tt6Flip( tCur, pComp[pCell.nFanins][c] );
            Perm1 = Perm + pComp[pCell.nFanins][c];
            *Perm1 = Abc_LitNot( *Perm1 );
        }
        assert(tTemp2 == tCur, "tTemp2 == tCur");
        if nPerms == 1 {
            continue; }
        // update
        tCur = Abc_Tt6SwapAdjacent( tCur, pPerm[pCell.nFanins][p] );
        Perm1 = Perm + pPerm[pCell.nFanins][p];
        Perm2 = Perm1 + 1;
        // inlining of macro ABC_SWAP
        { var t int32  = *Perm1; *Perm1 = *Perm2; *Perm2 = t; };
    }
    assert(tTemp1 == tCur, "tTemp1 == tCur");
}
func Nf_StoDeriveMatches(vTtMem *Vec_Mem_t, vTt2Match *Vec_Wec_t, pnCells *int32, fPinFilter int32, fPinPerm int32, fPinQuick int32) *Mio_Cell2_t {
    var fVerbose int32;
    //abctime clk = Abc_Clock();
    var vProfs *Vec_Wec_t  = Vec_WecAlloc( 1000 );
    var vStore *Vec_Int_t  = Vec_IntAlloc( 10000 );
    var pComp [7]*int32;
    var pPerm [7]*int32;
    var nPerms [7]int32;
    var i int32;
    var pCells *Mio_Cell2_t;
    Vec_WecPushLevel( vProfs );
    Vec_WecPushLevel( vProfs );
    for  i = 1; i <= 6; i++  {
        pComp[i] = Extra_GreyCodeSchedule( i );
    }
    for  i = 1; i <= 6; i++  {
        pPerm[i] = Extra_PermSchedule( i );
    }
    for  i = 1; i <= 6; i++  {
        nPerms[i] = Extra_Factorial( i );
    }
    pCells = Mio_CollectRootsNewDefault2( 6, pnCells, fVerbose );
    if pCells != nil {
    for  i = 2; i < *pnCells; i++  {
        Nf_StoCreateGateMaches( vTtMem, vTt2Match, pCells+i, pComp, pPerm, nPerms, vProfs, vStore, fPinFilter, fPinPerm, fPinQuick );
    } }
    for  i = 1; i <= 6; i++  {
        // inlining of macro ABC_FREE
        pComp[i] = nil;
    }
    for  i = 1; i <= 6; i++  {
        // inlining of macro ABC_FREE
        pPerm[i] = nil;
    }
    Vec_WecFree( vProfs );
    Vec_IntFree( vStore );
    //Abc_PrintTime( 1, "Time", Abc_Clock() - clk );
    return pCells;
}
func Nf_StoPrintOne(p *Nf_Man_t, Count int32, t int32, i int32, GateId int32, Mat Nf_Cfg_t) {
    var pC *Mio_Cell2_t  = p.pCells + GateId;
    var pTruth *word  = Vec_MemReadEntry(p.vTtMem, t);
    var k int32;
    var nSuppSize int32 = Abc_TtSupportSize(pTruth, 6);
    fmt.Printf("%6d : ", Count);
    fmt.Printf("%6d : ", t);
    fmt.Printf("%6d : ", i);
    fmt.Printf("Gate %16s  ", pC.pName);
    fmt.Printf("Area =%8.2f  ", pC.AreaF);
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
func Nf_StoPrint(p *Nf_Man_t, fVerbose int32) {
    var t, i, GateId, Entry, Count int32;
    for  t = 2; t < Vec_WecSize(p.vTt2Match); t++  {
        var vArr *Vec_Int_t  = Vec_WecEntry( p.vTt2Match, t );
        // inlining of macro Vec_IntForEachEntryDouble
        for  i = 0; (i+1 < Vec_IntSize(vArr)) && (((GateId) = Vec_IntEntry(vArr, i)), 1) && (((Entry) = Vec_IntEntry(vArr, i+1)), 1); i += 2  {
            Count++;
            if !(fVerbose != 0) {
                continue; }
            //if ( t < 10 )
            //    Nf_StoPrintOne( p, Count, t, i/2, GateId, Pf_Int2Mat(Entry) );
        }
    }
    fmt.Printf("Gates = %d.  Truths = %d.  Matches = %d.\n", p.nCells, Vec_MemEntryNum(p.vTtMem), Count);
}



/**Function*************************************************************

  Synopsis    []

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Nf_StoCreate(pGia *Gia_Man_t, pPars *Jf_Par_t) *Nf_Man_t {
    extern void Mf_ManSetFlowRefs( Gia_Man_t * p, Vec_Int_t * vRefs );
    var vFlowRefs *Vec_Int_t;
    var p *Nf_Man_t;
    var i, Entry int32;
    assert(pPars.nCutNum > 1  && pPars.nCutNum <= NF_CUT_MAX, "pPars.nCutNum > 1  && pPars.nCutNum <= NF_CUT_MAX");
    assert(pPars.nLutSize > 1 && pPars.nLutSize <= NF_LEAF_MAX, "pPars.nLutSize > 1 && pPars.nLutSize <= NF_LEAF_MAX");
    // inlining of macro ABC_FREE
    pGia.Refs = nil;
    Vec_IntFreeP( &pGia.vCellMapping );
    if Gia_ManHasChoices(pGia) {
        Gia_ManSetPhase(pGia); }
    // create
    // inlining of macro ABC_CALLOC
    p = &Nf_Man_t{};
    p.clkStart = Abc_Clock();
    p.pGia     = pGia;
    p.pPars    = pPars;
    // inlining of macro ABC_CALLOC
    p.pNfObjs  = ((Nf_Obj_t *) calloc((size_t)(Gia_ManObjNum(pGia)), sizeof(Nf_Obj_t)));
    p.iCur     = 2;
    // other
    Vec_PtrGrow( &p.vPages, 256 );                                    // cut memory
    Vec_IntFill( &p.vMapRefs,  2*Gia_ManObjNum(pGia), 0 );            // mapping refs   (2x)
    Vec_FltFill( &p.vFlowRefs, 2*Gia_ManObjNum(pGia), 0 );            // flow refs      (2x)
    Vec_IntFill( &p.vRequired, 2*Gia_ManObjNum(pGia), SCL_INFINITY ); // required times (2x)
    Vec_IntFill( &p.vCutSets,  Gia_ManObjNum(pGia), 0 );              // cut offsets
    Vec_FltFill( &p.vCutFlows, Gia_ManObjNum(pGia), 0 );              // cut area
    Vec_IntFill( &p.vCutDelays,Gia_ManObjNum(pGia), 0 );              // cut delay
    Vec_IntGrow( &p.vBackup, 1000 );
    // references
    vFlowRefs = Vec_IntAlloc(0);
    Mf_ManSetFlowRefs( pGia, vFlowRefs );
    // inlining of macro Vec_IntForEachEntry
    for  i = 0; (i < Vec_IntSize(vFlowRefs)) && (((Entry) = Vec_IntEntry(vFlowRefs, i)), 1); i++  {
        Vec_FltWriteEntry( &p.vFlowRefs, 2*i,   /*0.5* */Entry );
        Vec_FltWriteEntry( &p.vFlowRefs, 2*i+1, /*0.5* */Entry );
    }
    Vec_IntFree(vFlowRefs);
    // matching
    Mio_LibraryMatchesFetch( (Mio_Library_t *)Abc_FrameReadLibGen(), &p.vTtMem, &p.vTt2Match, &p.pCells, &p.nCells, p.pPars.fPinFilter, p.pPars.fPinPerm, p.pPars.fPinQuick );
    if p.pCells == nil {
        return nil; }
    p.InvDelayI = p.pCells[3].iDelays[0];
    p.InvAreaW  = p.pCells[3].AreaW;
    p.InvAreaF  = p.pCells[3].AreaF;
    Nf_ObjMatchD(p, 0, 0).Gate = 0;
    Nf_ObjMatchD(p, 0, 1).Gate = 1;
    // prepare cuts
    return p;
}
func Nf_StoDelete(p *Nf_Man_t) {
    Vec_PtrFreeData( &p.vPages );
    // inlining of macro ABC_FREE
    p.vPages.pArray = nil;
    // inlining of macro ABC_FREE
    p.vMapRefs.pArray = nil;
    // inlining of macro ABC_FREE
    p.vFlowRefs.pArray = nil;
    // inlining of macro ABC_FREE
    p.vRequired.pArray = nil;
    // inlining of macro ABC_FREE
    p.vCutSets.pArray = nil;
    // inlining of macro ABC_FREE
    p.vCutFlows.pArray = nil;
    // inlining of macro ABC_FREE
    p.vCutDelays.pArray = nil;
    // inlining of macro ABC_FREE
    p.vBackup.pArray = nil;
    // inlining of macro ABC_FREE
    p.pNfObjs = nil;
    // inlining of macro ABC_FREE
    p = nil;
}




/**Function*************************************************************

  Synopsis    []

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Nf_CutComputeTruth6(p *Nf_Man_t, pCut0 *Nf_Cut_t, pCut1 *Nf_Cut_t, fCompl0 int32, fCompl1 int32, pCutR *Nf_Cut_t, fIsXor int32) int32 {
//    extern int Nf_ManTruthCanonicize( word * t, int nVars );
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
    pCutR.Useless = Nf_ObjCutUseless( p, truthId );
    assert(int32(pCutR.nLeaves) <= nOldSupp, "int32(pCutR.nLeaves) <= nOldSupp");
    return int32(pCutR.nLeaves) < nOldSupp;
}
func Nf_CutComputeTruthMux6(p *Nf_Man_t, pCut0 *Nf_Cut_t, pCut1 *Nf_Cut_t, pCutC *Nf_Cut_t, fCompl0 int32, fCompl1 int32, fComplC int32, pCutR *Nf_Cut_t) int32 {
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
    pCutR.Useless = Nf_ObjCutUseless( p, truthId );
    assert(int32(pCutR.nLeaves) <= nOldSupp, "int32(pCutR.nLeaves) <= nOldSupp");
    return int32(pCutR.nLeaves) < nOldSupp;
}


/**Function*************************************************************

  Synopsis    []

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/



func Nf_CutPrint(p *Nf_Man_t, pCut *Nf_Cut_t) {
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
func Nf_ManPrepareCuts(pCuts *Nf_Cut_t, p *Nf_Man_t, iObj int32, fAddUnit int32) int32 {
    if Nf_ObjHasCuts(p, iObj) != 0 {
        var pMfCut *Nf_Cut_t  = pCuts;
        var i int32;
        var pCut *int32;
        var pList *int32 = Nf_ObjCutSet(p, iObj);
        // inlining of macro Nf_SetForEachCut
        for  i, pCut = 0, pList + 1; i < pList[0]; i++, pCut += Nf_CutSize(pCut) + 1  {
            pMfCut.Delay   = 0;
            pMfCut.Flow    = 0;
            pMfCut.iFunc   = Nf_CutFunc( pCut );
            pMfCut.nLeaves = Nf_CutSize( pCut );
            pMfCut.Sign    = Nf_CutGetSign( pCut+1, Nf_CutSize(pCut) );
            pMfCut.Useless = Nf_ObjCutUseless( p, Abc_Lit2Var(pMfCut.iFunc) );
            memcpy( pMfCut.pLeaves, pCut+1, sizeof(int) * Nf_CutSize(pCut) );
            pMfCut++;
        }
        if (fAddUnit != 0) && pCuts.nLeaves > 1 {
            return pList[0] + Nf_CutCreateUnit( pMfCut, iObj ); }
        return pList[0];
    }
    return Nf_CutCreateUnit( pCuts, iObj );
}
func Nf_ManSaveCuts(p *Nf_Man_t, pCuts **Nf_Cut_t, nCuts int32, fUseful int32) int32 {
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
    pPlace = Nf_ManCutSet( p, iCur );
    *pPlace++ = nCutsNew;
    for  i = 0; i < nCuts; i++  {
        if !(fUseful != 0) || !pCuts[i].Useless {
            *pPlace++ = Nf_CutSetBoth( pCuts[i].nLeaves, pCuts[i].iFunc );
            memcpy( pPlace, pCuts[i].pLeaves, sizeof(int) * pCuts[i].nLeaves );
            pPlace += pCuts[i].nLeaves;
        }
    }
    return iCur;
}
func Nf_ManCountUseful(pCuts **Nf_Cut_t, nCuts int32) int32 {
    var i, Count int32;
    for  i = 0; i < nCuts; i++  {
        Count += !pCuts[i].Useless;
    }
    return Count;
}
func Nf_ManCountMatches(p *Nf_Man_t, pCuts **Nf_Cut_t, nCuts int32) int32 {
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
static inline int Nf_CutCheck( Nf_Cut_t * pBase, Nf_Cut_t * pCut ) // check if pCut is contained in pBase
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
func Nf_SetCheckArray(ppCuts **Nf_Cut_t, nCuts int32) int32 {
    var pCut0, pCut1 *Nf_Cut_t; 
    var i, k, m, n, Value int32;
    assert(nCuts > 0, "nCuts > 0");
    for  i = 0; i < nCuts; i++  {
        pCut0 = ppCuts[i];
        assert(pCut0.nLeaves <= NF_LEAF_MAX, "pCut0.nLeaves <= NF_LEAF_MAX");
        assert(pCut0.Sign == Nf_CutGetSign(pCut0.pLeaves, pCut0.nLeaves), "pCut0.Sign == Nf_CutGetSign(pCut0.pLeaves, pCut0.nLeaves)");
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
            Value = Nf_CutCheck( pCut0, pCut1 );
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


static inline int Nf_SetCutIsContainedOrder( Nf_Cut_t * pBase, Nf_Cut_t * pCut ) // check if pCut is contained in pBase
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





func Nf_CutArea(p *Nf_Man_t, nLeaves int32) int32 {
    if nLeaves < 2 {
        return 0; }
    return nLeaves + p.pPars.nAreaTuner;
}
func Nf_CutParams(p *Nf_Man_t, pCut *Nf_Cut_t, FlowRefs float32) {
    var i int32;
    var nLeaves int32 = pCut.nLeaves; 
    assert(nLeaves <= p.pPars.nLutSize, "nLeaves <= p.pPars.nLutSize");
    pCut.Delay = 0;
    pCut.Flow  = 0;
    for  i = 0; i < nLeaves; i++  {
        pCut.Delay = Abc_MaxInt( pCut.Delay, Nf_ObjCutDelay(p, pCut.pLeaves[i]) );
        pCut.Flow += Nf_ObjCutFlow(p, pCut.pLeaves[i]);
    }
    pCut.Delay += int32((nLeaves > 1));
    pCut.Flow = (pCut.Flow + Nf_CutArea(p, nLeaves)) / FlowRefs;
}
func Nf_ObjMergeOrder(p *Nf_Man_t, iObj int32) {
    var pCuts0 [NF_CUT_MAX]Nf_Cut_t;
    var pCuts1 [NF_CUT_MAX]Nf_Cut_t;
    var pCuts [NF_CUT_MAX]Nf_Cut_t;
    var pCutsR [NF_CUT_MAX]*Nf_Cut_t;
    var pObj *Gia_Obj_t  = Gia_ManObj(p.pGia, iObj);
    //Nf_Obj_t * pBest = Nf_ManObj(p, iObj);
    var dFlowRefs float32   = Nf_ObjFlowRefs(p, iObj, 0) + Nf_ObjFlowRefs(p, iObj, 1);
    var nLutSize int32  = p.pPars.nLutSize;
    var nCutNum int32   = p.pPars.nCutNum;
    var nCuts0 int32    = Nf_ManPrepareCuts(pCuts0, p, Gia_ObjFaninId0(pObj, iObj), 1);
    var nCuts1 int32    = Nf_ManPrepareCuts(pCuts1, p, Gia_ObjFaninId1(pObj, iObj), 1);
    var fComp0 int32    = Gia_ObjFaninC0(pObj);
    var fComp1 int32    = Gia_ObjFaninC1(pObj);
    var iSibl int32     = Gia_ObjSibl(p.pGia, iObj);
    var pCut0 *Nf_Cut_t;
    var pCut1 *Nf_Cut_t;
    var pCut0Lim *Nf_Cut_t = pCuts0 + nCuts0;
    var pCut1Lim *Nf_Cut_t = pCuts1 + nCuts1;
    var i, nCutsUse, nCutsR int32;
    assert(!Gia_ObjIsBuf(pObj), "!Gia_ObjIsBuf(pObj)");
    for  i = 0; i < nCutNum; i++  {
        pCutsR[i] = pCuts + i;
    }
    if iSibl != 0 {
        var pCuts2 [NF_CUT_MAX]Nf_Cut_t;
        var pObjE *Gia_Obj_t  = Gia_ObjSiblObj(p.pGia, iObj);
        var fCompE int32  = Gia_ObjPhase(pObj) ^ Gia_ObjPhase(pObjE);
        var nCuts2 int32  = Nf_ManPrepareCuts(pCuts2, p, iSibl, 0);
        var pCut2 *Nf_Cut_t;
        var pCut2Lim *Nf_Cut_t = pCuts2 + nCuts2;
        for  pCut2 = pCuts2; pCut2 < pCut2Lim; pCut2++  {
            *pCutsR[nCutsR] = *pCut2;
            pCutsR[nCutsR].iFunc = Abc_LitNotCond( pCutsR[nCutsR].iFunc, fCompE );
            Nf_CutParams( p, pCutsR[nCutsR], dFlowRefs );
            nCutsR = Nf_SetAddCut( pCutsR, nCutsR, nCutNum );
        }
    }
    if Gia_ObjIsMuxId(p.pGia, iObj) {
        var pCuts2 [NF_CUT_MAX]Nf_Cut_t;
        var nCuts2 int32   = Nf_ManPrepareCuts(pCuts2, p, Gia_ObjFaninId2(p.pGia, iObj), 1);
        var fComp2 int32   = Gia_ObjFaninC2(p.pGia, pObj);
        var pCut2 *Nf_Cut_t;
        var pCut2Lim *Nf_Cut_t = pCuts2 + nCuts2;
        p.CutCount[0] += nCuts0 * nCuts1 * nCuts2;
        for  pCut0 = pCuts0; pCut0 < pCut0Lim; pCut0++  {
        for  pCut1 = pCuts1; pCut1 < pCut1Lim; pCut1++  {
        for  pCut2 = pCuts2; pCut2 < pCut2Lim; pCut2++  {
            if Nf_CutCountBits(pCut0.Sign | pCut1.Sign | pCut2.Sign) > nLutSize {
                continue; }
            p.CutCount[1]++; 
            if !(Nf_CutMergeOrderMux(pCut0, pCut1, pCut2, pCutsR[nCutsR], nLutSize) != 0) {
                continue; }
            if Nf_SetLastCutIsContained(pCutsR, nCutsR) != 0 {
                continue; }
            p.CutCount[2]++;
            if Nf_CutComputeTruthMux6(p, pCut0, pCut1, pCut2, fComp0, fComp1, fComp2, pCutsR[nCutsR]) != 0 {
                pCutsR[nCutsR].Sign = Nf_CutGetSign(pCutsR[nCutsR].pLeaves, pCutsR[nCutsR].nLeaves); }
            Nf_CutParams( p, pCutsR[nCutsR], dFlowRefs );
            nCutsR = Nf_SetAddCut( pCutsR, nCutsR, nCutNum );
        }
        }
        }
    
    } else {
        var fIsXor int32  = Gia_ObjIsXor(pObj);
        p.CutCount[0] += nCuts0 * nCuts1;
        for  pCut0 = pCuts0; pCut0 < pCut0Lim; pCut0++  {
        for  pCut1 = pCuts1; pCut1 < pCut1Lim; pCut1++  {
            if int32((pCut0.nLeaves + pCut1.nLeaves)) > nLutSize && Nf_CutCountBits(pCut0.Sign | pCut1.Sign) > nLutSize {
                continue; }
            p.CutCount[1]++; 
            if !(Nf_CutMergeOrder(pCut0, pCut1, pCutsR[nCutsR], nLutSize) != 0) {
                continue; }
            if Nf_SetLastCutIsContained(pCutsR, nCutsR) != 0 {
                continue; }
            p.CutCount[2]++;
            if Nf_CutComputeTruth6(p, pCut0, pCut1, fComp0, fComp1, pCutsR[nCutsR], fIsXor) != 0 {
                pCutsR[nCutsR].Sign = Nf_CutGetSign(pCutsR[nCutsR].pLeaves, pCutsR[nCutsR].nLeaves); }
            Nf_CutParams( p, pCutsR[nCutsR], dFlowRefs );
            nCutsR = Nf_SetAddCut( pCutsR, nCutsR, nCutNum );
        }
        }
    }
    // debug printout
    if false {
//    if ( iObj % 10000 == 0 )
//    if ( iObj == 1090 )
    
        fmt.Printf("*** Obj = %d  Useful = %d\n", iObj, Nf_ManCountUseful(pCutsR, nCutsR));
        for  i = 0; i < nCutsR; i++  {
            Nf_CutPrint( p, pCutsR[i] );
        }
        fmt.Printf("\n");
    } 
    // verify
    assert(nCutsR > 0 && nCutsR < nCutNum, "nCutsR > 0 && nCutsR < nCutNum");
//    assert( Nf_SetCheckArray(pCutsR, nCutsR) );
    // store the cutset
    Nf_ObjSetCutFlow( p, iObj, pCutsR[0].Flow );
    Nf_ObjSetCutDelay( p, iObj, pCutsR[0].Delay );
    *Vec_IntEntryP(&p.vCutSets, iObj) = Nf_ManSaveCuts(p, pCutsR, nCutsR, 0);
    p.CutCount[3] += nCutsR;
    nCutsUse = Nf_ManCountUseful(pCutsR, nCutsR);
    p.CutCount[4] += nCutsUse;
    p.nCutUseAll  += nCutsUse == nCutsR;
    p.CutCount[5] += Nf_ManCountMatches(p, pCutsR, nCutsR);
}
func Nf_ManComputeCuts(p *Nf_Man_t) {
    var pObj *Gia_Obj_t; var i, iFanin int32;
    // inlining of macro Gia_ManForEachAnd
    for  i = 0; (i < p.pGia.nObjs) && ((pObj) = Gia_ManObj(p.pGia, i)); i++  {      if !Gia_ObjIsAnd(pObj) {} else
        if Gia_ObjIsBuf(pObj) {
            iFanin = Gia_ObjFaninId0(pObj, i);
            Nf_ObjSetCutFlow( p, i,  Nf_ObjCutFlow(p, iFanin) );
            Nf_ObjSetCutDelay( p, i, Nf_ObjCutDelay(p, iFanin) );
        
        } else {
            Nf_ObjMergeOrder( p, i ); }
    }
}




/**Function*************************************************************

  Synopsis    []

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Nf_ManPrintStats(p *Nf_Man_t, pTitle string) {
    if !p.pPars.fVerbose {
        return; }
    fmt.Printf("%s :  ", pTitle);
    fmt.Printf("Delay =%8.2f  ", Scl_Int2Flt(p.pPars.MapDelay));
    fmt.Printf("Area =%12.2f  ", p.pPars.MapAreaF);
    fmt.Printf("Gate =%6d  ", int32(p.pPars.Area));
    fmt.Printf("Inv =%6d  ", int32(p.nInvs));
    fmt.Printf("Edge =%7d  ", int32(p.pPars.Edge));
    Abc_PrintTime( 1, "Time", Abc_Clock() - p.clkStart );
    fflush( os.Stdout );
}
func Nf_ManPrintInit(p *Nf_Man_t) {
    var nChoices int32;
    if !p.pPars.fVerbose {
        return; }
    fmt.Printf("LutSize = %d  ", p.pPars.nLutSize);
    fmt.Printf("CutNum = %d  ", p.pPars.nCutNum);
    fmt.Printf("Iter = %d  ", p.pPars.nRounds);//+ p->pPars->nRoundsEla );
    fmt.Printf("Coarse = %d   ", p.pPars.fCoarsen);
    fmt.Printf("Cells = %d  ", p.nCells);
    fmt.Printf("Funcs = %d  ", Vec_MemEntryNum(p.vTtMem));
    fmt.Printf("Matches = %d  ", Vec_WecSizeSize(p.vTt2Match)/2);
    fmt.Printf("And = %d  ", Gia_ManAndNum(p.pGia));
    nChoices = Gia_ManChoiceNum( p.pGia );
    if nChoices != 0 {
    fmt.Printf("Choices = %d  ", nChoices); }
    fmt.Printf("\n");
    fmt.Printf("Computing cuts...\r");
    fflush( os.Stdout );
}
func Nf_ManPrintQuit(p *Nf_Man_t) {
    var MemGia float32    = Gia_ManMemory(p.pGia) / (1<<20);
    var MemMan float32    =(1.0 * sizeof(Nf_Obj_t) + 8.0 * sizeof(int)) * Gia_ManObjNum(p.pGia) / (1<<20);
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
func Nf_ManCutMatchPrint(p *Nf_Man_t, iObj int32, pStr string, pM *Nf_Mat_t) {
    var pCell *Mio_Cell2_t;
    var i int32;
    var pCut *int32;
    fmt.Printf("%5d %s : ", iObj, pStr);
    if pM.CutH == 0 {
        fmt.Printf("Unassigned\n");
        return;
    }
    pCell = Nf_ManCell( p, pM.Gate );
    pCut = Nf_CutFromHandle( Nf_ObjCutSet(p, iObj), pM.CutH );
    fmt.Printf("D =%6.2f  ", Scl_Int2Flt(pM.D));
    fmt.Printf("A =%6.2f  ", pM.F);
    fmt.Printf("C = %d ", pM.fCompl);
//    printf( "B = %d ", pM->fBest );
    fmt.Printf("  ");
    fmt.Printf("Cut = {");
    for  i = 0; i < int32(pCell.nFanins); i++  {
        fmt.Printf("%4d ", Nf_CutLeaves(pCut)[i]);
    }
    for  ; i < 6; i++  {
        fmt.Printf("     ");
    }
    fmt.Printf("}  ");
    fmt.Printf("%10s ", pCell.pName);
    fmt.Printf("%d  ", pCell.nFanins);
    fmt.Printf("{");
    for  i = 0; i < int32(pCell.nFanins); i++  {
        fmt.Printf("%6.2f ", Scl_Int2Flt(pCell.iDelays[i]));
    }
    for  ; i < 6; i++  {
        fmt.Printf("       ");
    }
    fmt.Printf(" } ");
    for  i = 0; i < int32(pCell.nFanins); i++  {
        fmt.Printf("%s%d ", Nf_CfgCompl(pM.Cfg, i) ? "!":" ", Nf_CfgVar(pM.Cfg, i));
    }
    for  ; i < 6; i++  {
        fmt.Printf("  ");
    }
    Dau_DsdPrintFromTruth( &pCell.uTruth, pCell.nFanins );
}
func Nf_ManCutMatchOne(p *Nf_Man_t, iObj int32, pCut *int32, pCutSet *int32) {
    var pBest *Nf_Obj_t  = Nf_ManObj(p, iObj);
    var pFans *int32       = Nf_CutLeaves(pCut);
    var nFans int32         = Nf_CutSize(pCut);
    var iFuncLit int32      = Nf_CutFunc(pCut);
    var fComplExt int32     = Abc_LitIsCompl(iFuncLit);
    var vArr *Vec_Int_t  = Vec_WecEntry( p.vTt2Match, Abc_Lit2Var(iFuncLit) );
    var i, k, c, Info, Offset, iFanin, fComplF int32;
    var ArrivalD, ArrivalA int32;
    var pD, pA *Nf_Mat_t;
    // assign fanins matches
    var pBestF [NF_LEAF_MAX]*Nf_Obj_t;
    for  i = 0; i < nFans; i++  {
        pBestF[i] = Nf_ManObj( p, pFans[i] );
    }
    // special cases
    if nFans == 0 {
        var Const int32  = (iFuncLit == 1);
        assert(iFuncLit == 0 || iFuncLit == 1, "iFuncLit == 0 || iFuncLit == 1");
        for  c = 0; c < 2; c++  { 
            pD = Nf_ObjMatchD( p, iObj, c );
            pA = Nf_ObjMatchA( p, iObj, c );
            pD.D = pA.D = 0;
            pD.F = pA.F = p.pCells[c ^ Const].AreaF;
            pD.CutH = pA.CutH = Nf_CutHandle(pCutSet, pCut);
            pD.Gate = pA.Gate = c ^ Const;
//            pD->Conf = pA->Conf = 0;
            pD.Cfg = pA.Cfg = Nf_Int2Cfg(0);
        }
        return;
    }
    // consider matches of this function
    // inlining of macro Vec_IntForEachEntryDouble
    for  i = 0; (i+1 < Vec_IntSize(vArr)) && (((Info) = Vec_IntEntry(vArr, i)), 1) && (((Offset) = Vec_IntEntry(vArr, i+1)), 1); i += 2  {
        var Cfg Nf_Cfg_t    = Nf_Int2Cfg(Offset);
        var pC *Mio_Cell2_t  = Nf_ManCell( p, Info );
        var fCompl int32      = Cfg.fCompl ^ fComplExt;
        var Required int32 = Nf_ObjRequired( p, iObj, fCompl );
        var Delay int32;
        var pD *Nf_Mat_t   = &pBest.M[fCompl][0];
        var pA *Nf_Mat_t   = &pBest.M[fCompl][1];
        var AreaF float32     = pC.AreaF;
        assert(nFans == int32(pC.nFanins), "nFans == int32(pC.nFanins)");
        // inlining of macro Nf_CfgForEachVarCompl
        for  k = 0; k < nFans && ((iFanin = Nf_CfgVar(Cfg, k)), 1) && ((fComplF = Nf_CfgCompl(Cfg, k)), 1); k++  {
            ArrivalD  = pBestF[iFanin].M[fComplF][0].D;
            ArrivalA  = pBestF[iFanin].M[fComplF][1].D;
            if ArrivalA + pC.iDelays[k] <= Required && Required != SCL_INFINITY {
                Delay = Abc_MaxInt( Delay, ArrivalA + pC.iDelays[k] );
                AreaF += pBestF[iFanin].M[fComplF][1].F;
            
            } else {
                if pD.D < SCL_INFINITY && pA.D < SCL_INFINITY && ArrivalD + pC.iDelays[k] > Required {
                    break; }
                Delay = Abc_MaxInt( Delay, ArrivalD + pC.iDelays[k] );
                //AreaF += pBestF[iFanin]->M[fComplF][0].F;
                if AreaF >= float32(1e32) || pBestF[iFanin].M[fComplF][0].F >= float32(1e32) {
                    AreaF = float32(1e32);
                } else { 
                    AreaF += pBestF[iFanin].M[fComplF][0].F; }
            }
        }
        if k < nFans {
            continue; }
        // select best Cfgch
        if pD.D > Delay {
            pD.D = Delay;
            pD.F = AreaF;
            pD.CutH = Nf_CutHandle(pCutSet, pCut);
            pD.Gate = pC.Id;
            pD.Cfg = Cfg;
            pD.Cfg.fCompl = 0;
        }

        if pA.F > AreaF + NF_EPSILON {
            pA.D = Delay;
            pA.F = AreaF;
            pA.CutH = Nf_CutHandle(pCutSet, pCut);
            pA.Gate = pC.Id;
            pA.Cfg = Cfg;
            pA.Cfg.fCompl = 0;
        }
    }
}
func Nf_ObjPrepareCi(p *Nf_Man_t, iObj int32, Time int32) {
    var pD0 *Nf_Mat_t  = Nf_ObjMatchD( p, iObj, 0 );
    var pA0 *Nf_Mat_t  = Nf_ObjMatchA( p, iObj, 0 );
    var pD *Nf_Mat_t  = Nf_ObjMatchD( p, iObj, 1 );
    var pA *Nf_Mat_t  = Nf_ObjMatchA( p, iObj, 1 );
    pD0.D = pA0.D = pD.D = pA.D = Time;
    pD.fCompl = 1;
    pD.D += p.InvDelayI;
    pD.F = p.InvAreaF;
    pA.fCompl = 1;
    pA.D += p.InvDelayI;
    pA.F = p.InvAreaF;
    Nf_ObjMatchD( p, iObj, 0 ).fBest = 1;
    Nf_ObjMatchD( p, iObj, 1 ).fBest = 1;
}
func Nf_ObjPrepareBuf(p *Nf_Man_t, pObj *Gia_Obj_t) {
    // get fanin info
    var iObj int32  = Gia_ObjId( p.pGia, pObj );
    var iFanin int32  = Gia_ObjFaninId0( pObj, iObj );
    var pDf *Nf_Mat_t  = Nf_ObjMatchD( p, iFanin, Gia_ObjFaninC0(pObj) );
    //Nf_Mat_t * pAf = Nf_ObjMatchA( p, iFanin, Gia_ObjFaninC0(pObj) );
    // set the direct phase
    var pDp *Nf_Mat_t  = Nf_ObjMatchD( p, iObj, 0 );
    var pAp *Nf_Mat_t  = Nf_ObjMatchA( p, iObj, 0 );
    var pDn *Nf_Mat_t  = Nf_ObjMatchD( p, iObj, 1 );
    var pAn *Nf_Mat_t  = Nf_ObjMatchA( p, iObj, 1 );
    assert(Gia_ObjIsBuf(pObj), "Gia_ObjIsBuf(pObj)");
    memset( Nf_ManObj(p, iObj), 0, sizeof(Nf_Obj_t) );
    // set the direct phase
    pDp.D = pAp.D = pDf.D;
    pDp.F = pAp.F = pDf.F; // do not pass flow???
    pDp.fBest = 1;
    // set the inverted phase
    pDn.D = pAn.D = pDf.D + p.InvDelayI;
    pDn.F = pAn.F = pDf.F + p.InvAreaF;
    pDn.fCompl = pAn.fCompl = 1;
    pDn.fBest = 1;
}
func Nf_CutRequired(p *Nf_Man_t, pM *Nf_Mat_t, pCutSet *int32) int32 {
    var pCell *Mio_Cell2_t  = Nf_ManCell( p, pM.Gate );
    var pCut *int32  = Nf_CutFromHandle( pCutSet, pM.CutH );
    var i, iVar, fCompl int32;
    var Arr, Req, Arrival, Required int32;
    // inlining of macro Nf_CutForEachVarCompl
    for  i = 0; i < Nf_CutSize(pCut) && (iVar = Nf_CutLeaves(pCut)[Nf_CfgVar(pM.Cfg, i)]) && ((fCompl = Nf_CfgCompl(pM.Cfg, i)), 1); i++  {
        Arr     = Nf_ManObj(p, iVar).M[fCompl][0].D + pCell.iDelays[i];
        Req     = Nf_ObjRequired(p, iVar, fCompl);
        Arrival = Abc_MaxInt( Arrival, Arr );
        if Req < SCL_INFINITY {
            Required = Abc_MaxInt( Required, Req + pCell.iDelays[i] ); }
    }
    return Abc_MaxInt( Required + p.pPars.nReqTimeFlex*p.InvDelayI, Arrival ); 
}
func Nf_ObjComputeRequired(p *Nf_Man_t, iObj int32) {
    var pBest *Nf_Obj_t  = Nf_ManObj(p, iObj);
    var c int32;
    var pCutSet *int32 = Nf_ObjCutSet( p, iObj );
    for  c = 0; c < 2; c++  {
        if Nf_ObjRequired(p, iObj, c) == SCL_INFINITY {
            Nf_ObjSetRequired( p, iObj, c, Nf_CutRequired(p, &pBest.M[c][0], pCutSet) ); }
    }
}
func Nf_ManCutMatch(p *Nf_Man_t, iObj int32) {
    var pBest *Nf_Obj_t  = Nf_ManObj(p, iObj);
    var pDp *Nf_Mat_t  = &pBest.M[0][0];
    var pDn *Nf_Mat_t  = &pBest.M[1][0];
    var pAp *Nf_Mat_t  = &pBest.M[0][1];
    var pAn *Nf_Mat_t  = &pBest.M[1][1];
    var FlowRefPf float32   = Nf_ObjFlowRefs(p, iObj, 0);
    var FlowRefNf float32   = Nf_ObjFlowRefs(p, iObj, 1);
    var i int32;
    var pCut *int32;
    var pCutSet *int32 = Nf_ObjCutSet( p, iObj );
    var Required [2]int32 ;
    if p.Iter {
        Nf_ObjComputeRequired( p, iObj );
        Required[0] = Nf_ObjRequired( p, iObj, 0 );
        Required[1] = Nf_ObjRequired( p, iObj, 1 );
    }
    *pBest = Nf_Obj_t{};
    pDp.D = SCL_INFINITY; pDp.F = FLT_MAX;
    pDn.D = SCL_INFINITY; pDn.F = FLT_MAX;
    pAp.D = SCL_INFINITY; pAp.F = FLT_MAX;
    pAn.D = SCL_INFINITY; pAn.F = FLT_MAX;
    // inlining of macro Nf_SetForEachCut
    for  i, pCut = 0, pCutSet + 1; i < pCutSet[0]; i++, pCut += Nf_CutSize(pCut) + 1  {
        if Abc_Lit2Var(Nf_CutFunc(pCut)) >= Vec_WecSize(p.vTt2Match) {
            continue; }
        assert(!(Nf_CutIsTriv(pCut, iObj) != 0), "!Nf_CutIsTriv(pCut, iObj)");
        assert(Nf_CutSize(pCut) <= p.pPars.nLutSize, "Nf_CutSize(pCut) <= p.pPars.nLutSize");
        assert(Abc_Lit2Var(Nf_CutFunc(pCut)) < Vec_WecSize(p.vTt2Match), "Abc_Lit2Var(Nf_CutFunc(pCut)) < Vec_WecSize(p.vTt2Match)");
        Nf_ManCutMatchOne( p, iObj, pCut, pCutSet );
    }

/*
    if ( 461 == iObj && p->Iter == 0 )
    {
        printf( "\nObj %6d (%.2f %.2f):\n", iObj, Scl_Int2Flt(Required[0]), Scl_Int2Flt(Required[1]) );
        Nf_ManCutMatchPrint( p, iObj, "Dp", &pBest->M[0][0] );
        Nf_ManCutMatchPrint( p, iObj, "Dn", &pBest->M[1][0] );
        Nf_ManCutMatchPrint( p, iObj, "Ap", &pBest->M[0][1] );
        Nf_ManCutMatchPrint( p, iObj, "An", &pBest->M[1][1] );
        printf( "\n" );
    }
*/
    // divide by ref count
    pDp.F = pDp.F / FlowRefPf;
    pAp.F = pAp.F / FlowRefPf;
    pDn.F = pDn.F / FlowRefNf;
    pAn.F = pAn.F / FlowRefNf;

    // add the inverters
    assert(pDp.D < SCL_INFINITY || pDn.D < SCL_INFINITY, "pDp.D < SCL_INFINITY || pDn.D < SCL_INFINITY");
    if pDp.D > pDn.D + p.InvDelayI {
        *pDp = *pDn;
        pDp.D += p.InvDelayI;
        pDp.F += p.InvAreaF;
        pDp.fCompl = 1;
        if pAp.D == SCL_INFINITY {
            *pAp = *pDp; }
        //printf( "Using inverter to improve delay at node %d in phase %d.\n", iObj, 1 );
    
    } else if pDn.D > pDp.D + p.InvDelayI {
        *pDn = *pDp;
        pDn.D += p.InvDelayI;
        pDn.F += p.InvAreaF;
        pDn.fCompl = 1;
        if pAn.D == SCL_INFINITY {
            *pAn = *pDn; }
        //printf( "Using inverter to improve delay at node %d in phase %d.\n", iObj, 0 );
    }
    //assert( pAp->F < FLT_MAX || pAn->F < FLT_MAX );
    // try replacing pos with neg
    if pAp.D == SCL_INFINITY || (pAp.F > pAn.F + p.InvAreaF + NF_EPSILON && pAn.D + p.InvDelayI <= Required[0]) {
        assert(p.Iter > 0, "p.Iter > 0");
        *pAp = *pAn;
        pAp.D += p.InvDelayI;
        pAp.F += p.InvAreaF;
        pAp.fCompl = 1;
        if pDp.D == SCL_INFINITY {
            *pDp = *pAp; }
        //printf( "Using inverter to improve area at node %d in phase %d.\n", iObj, 1 );
    
    // try replacing neg with pos
    } else if pAn.D == SCL_INFINITY || (pAn.F > pAp.F + p.InvAreaF + NF_EPSILON && pAp.D + p.InvDelayI <= Required[1]) {
        assert(p.Iter > 0, "p.Iter > 0");
        *pAn = *pAp;
        pAn.D += p.InvDelayI;
        pAn.F += p.InvAreaF;
        pAn.fCompl = 1;
        if pDn.D == SCL_INFINITY {
            *pDn = *pAn; }
        //printf( "Using inverter to improve area at node %d in phase %d.\n", iObj, 0 );
    }

    if pDp.D == SCL_INFINITY {
        fmt.Printf("Object %d has pDp unassigned.\n", iObj); }
    if pDn.D == SCL_INFINITY {
        fmt.Printf("Object %d has pDn unassigned.\n", iObj); }
    if pAp.D == SCL_INFINITY {
        fmt.Printf("Object %d has pAp unassigned.\n", iObj); }
    if pAn.D == SCL_INFINITY {
        fmt.Printf("Object %d has pAn unassigned.\n", iObj); }
/*
    pDp->F = Abc_MinFloat( pDp->F, FLT_MAX/SCL_NUM );
    pDn->F = Abc_MinFloat( pDn->F, FLT_MAX/SCL_NUM );
    pAp->F = Abc_MinFloat( pAp->F, FLT_MAX/SCL_NUM );  
    pAn->F = Abc_MinFloat( pAn->F, FLT_MAX/SCL_NUM );
*/    
    assert(pDp.D < SCL_INFINITY, "pDp.D < SCL_INFINITY");
    assert(pDn.D < SCL_INFINITY, "pDn.D < SCL_INFINITY");
    assert(pAp.D < SCL_INFINITY, "pAp.D < SCL_INFINITY");
    assert(pAn.D < SCL_INFINITY, "pAn.D < SCL_INFINITY");

    assert(pDp.F < FLT_MAX, "pDp.F < FLT_MAX");
    assert(pDn.F < FLT_MAX, "pDn.F < FLT_MAX");
    assert(pAp.F < FLT_MAX, "pAp.F < FLT_MAX");
    assert(pAn.F < FLT_MAX, "pAn.F < FLT_MAX");

/*
    if ( p->Iter && (pDp->D > Required[0] || pDn->D > Required[1]) )
    {
        printf( "%5d : ", iObj );
        printf( "Dp = %6.2f  ", Scl_Int2Flt(pDp->D) );
        printf( "Dn = %6.2f  ", Scl_Int2Flt(pDn->D) );
        printf( "  " );
        printf( "Ap = %6.2f  ", Scl_Int2Flt(pAp->D) );
        printf( "An = %6.2f  ", Scl_Int2Flt(pAn->D) );
        printf( "  " );
        printf( "Rp = %6.2f  ", Scl_Int2Flt(Required[0]) );
        printf( "Rn = %6.2f  ", Scl_Int2Flt(Required[1]) );
        printf( "\n" );
    }
*/
}
func Nf_ManComputeMapping(p *Nf_Man_t) {
    var pObj *Gia_Obj_t; var i int32;
    // inlining of macro Gia_ManForEachAnd
    for  i = 0; (i < p.pGia.nObjs) && ((pObj) = Gia_ManObj(p.pGia, i)); i++  {      if !Gia_ObjIsAnd(pObj) {} else
        if Gia_ObjIsBuf(pObj) {
            Nf_ObjPrepareBuf( p, pObj );
        } else {
            Nf_ManCutMatch( p, i ); }
    }
}


/**Function*************************************************************

  Synopsis    []

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Nf_ObjMatchBest(p *Nf_Man_t, i int32, c int32) *Nf_Mat_t {
    var pD *Nf_Mat_t  = Nf_ObjMatchD(p, i, c);
    var pA *Nf_Mat_t  = Nf_ObjMatchA(p, i, c);
    assert(pD.fBest != pA.fBest, "pD.fBest != pA.fBest");
    //assert( Nf_ObjMapRefNum(p, i, c) > 0 );
    if pA.fBest {
        return pA; }
    if pD.fBest {
        return pD; }
    return nil;
}
func Nf_ManSetOutputRequireds(p *Nf_Man_t, fPropCompl int32) {
    var pObj *Gia_Obj_t;
    var Required int32;
    var MapDelayOld int32 = p.pPars.MapDelay;
    var fUseConMan int32  = Scl_ConIsRunning() && Scl_ConHasOutReqs();
    var i int32;
    var iObj int32;
    var fCompl int32;
    var nLits int32 = 2*Gia_ManObjNum(p.pGia);
    Vec_IntFill( &p.vRequired, nLits, SCL_INFINITY );
    // compute delay
    p.pPars.MapDelay = 0;
    // inlining of macro Gia_ManForEachCo
    for  i = 0; (i < Vec_IntSize(p.pGia.vCos)) && ((pObj) = Gia_ManCo(p.pGia, i)); i++  {
        Required = Nf_ObjMatchD( p, Gia_ObjFaninId0p(p.pGia, pObj), Gia_ObjFaninC0(pObj) ).D;
        p.pPars.MapDelay = Abc_MaxInt( p.pPars.MapDelay, Required );
    }
    if p.Iter && MapDelayOld < p.pPars.MapDelay && p.pGia.vOutReqs == nil {
        fmt.Printf("******** Critical delay violation %.2f -> %.2f ********\n", Scl_Int2Flt(MapDelayOld), Scl_Int2Flt(p.pPars.MapDelay)); } 
    p.pPars.MapDelay = Abc_MaxInt( p.pPars.MapDelay, MapDelayOld );
    // check delay target
    if p.pPars.MapDelayTarget == 0 && p.pPars.nRelaxRatio {
        p.pPars.MapDelayTarget = p.pPars.MapDelay * (100 + p.pPars.nRelaxRatio) / 100; }
    if p.pPars.MapDelayTarget > 0 {
        if p.pPars.MapDelay < p.pPars.MapDelayTarget {
            p.pPars.MapDelay = p.pPars.MapDelayTarget;
        } else if p.pPars.nRelaxRatio == 0 {
            Abc_Print( 0, "Relaxing user-specified delay target from %.2f to %.2f.\n", Scl_Int2Flt(p.pPars.MapDelayTarget), Scl_Int2Flt(p.pPars.MapDelay) ); }
    }
    //assert( p->pPars->MapDelayTarget == 0 );
    // set required times
    // inlining of macro Gia_ManForEachCo
    for  i = 0; (i < Vec_IntSize(p.pGia.vCos)) && ((pObj) = Gia_ManCo(p.pGia, i)); i++  {
        iObj     = Gia_ObjFaninId0p(p.pGia, pObj);
        fCompl   = Gia_ObjFaninC0(pObj);
        Required = Nf_ObjMatchD(p, iObj, fCompl).D;
        if p.pPars.fDoAverage {
            Required = Required * (100 + p.pPars.nRelaxRatio) / 100;
        } else {
            Required = p.pPars.MapDelay;
        }
        // if external required time can be achieved, use it
        if fUseConMan != 0 {
            if Scl_ConGetOutReq(i) > 0 && Required <= Scl_ConGetOutReq(i) {
                Required = Scl_ConGetOutReq(i); }
        
        } else if p.pGia.vOutReqs {
            var NewRequired int32  = Scl_Flt2Int(Vec_FltEntry(p.pGia.vOutReqs, i));
            if NewRequired > 0 && Required <= NewRequired {
                Required = Abc_MinInt( 2*Required, NewRequired ); }
        }
        // if external required cannot be achieved, set the earliest possible arrival time
//        else if ( p->pGia->vOutReqs && Vec_FltEntry(p->pGia->vOutReqs, i) > 0 && Required > Vec_FltEntry(p->pGia->vOutReqs, i) )
//            ptTime->Rise = ptTime->Fall = ptTime->Worst = Required;
        // otherwise, set the global required time
        Nf_ObjUpdateRequired( p, iObj, fCompl, Required );
        if (fPropCompl != 0) && iObj > 0 && Nf_ObjMatchBest(p, iObj, fCompl).fCompl {
            Nf_ObjUpdateRequired( p, iObj, !fCompl, Required - p.InvDelayI ); }
        //Nf_ObjMapRefInc( p, Gia_ObjFaninId0p(p->pGia, pObj), Gia_ObjFaninC0(pObj));
    }
}
func Nf_ManSetMapRefsGate(p *Nf_Man_t, iObj int32, Required int32, pM *Nf_Mat_t) {
    var k, iVar, fCompl int32;
    var pCell *Mio_Cell2_t  = Nf_ManCell( p, pM.Gate );
    var pCut *int32  = Nf_CutFromHandle( Nf_ObjCutSet(p, iObj), pM.CutH );
    // inlining of macro Nf_CutForEachVarCompl
    for  k = 0; k < Nf_CutSize(pCut) && (iVar = Nf_CutLeaves(pCut)[Nf_CfgVar(pM.Cfg, k)]) && ((fCompl = Nf_CfgCompl(pM.Cfg, k)), 1); k++  {
        Nf_ObjMapRefInc( p, iVar, fCompl );
        Nf_ObjUpdateRequired( p, iVar, fCompl, Required - pCell.iDelays[k] );
    }
    assert(Nf_CutSize(pCut) == int32(pCell.nFanins), "Nf_CutSize(pCut) == int32(pCell.nFanins)");
    // update global stats
    p.pPars.MapAreaF += pCell.AreaF;
    p.pPars.Edge += Nf_CutSize(pCut);
    p.pPars.Area++;
    // update status of the gate
    assert(pM.fBest == 0, "pM.fBest == 0");
    pM.fBest = 1;
}
func Nf_ManPrintMatches(p *Nf_Man_t) {
    var pObj *Gia_Obj_t; var i int32;
    // inlining of macro Gia_ManForEachAnd
    for  i = 0; (i < p.pGia.nObjs) && ((pObj) = Gia_ManObj(p.pGia, i)); i++  {      if !Gia_ObjIsAnd(pObj) {} else {
        var pDp *Nf_Mat_t  = Nf_ObjMatchD( p, i, 0 );
        var pAp *Nf_Mat_t  = Nf_ObjMatchA( p, i, 0 );
        var pDn *Nf_Mat_t  = Nf_ObjMatchD( p, i, 1 );
        var pAn *Nf_Mat_t  = Nf_ObjMatchA( p, i, 1 );

        fmt.Printf("%5d : ", i);
        fmt.Printf("Dp = %6.2f  ", Scl_Int2Flt(pDp.D));
        fmt.Printf("Dn = %6.2f  ", Scl_Int2Flt(pDn.D));
        fmt.Printf("  ");
        fmt.Printf("Ap = %6.2f  ", Scl_Int2Flt(pAp.D));
        fmt.Printf("An = %6.2f  ", Scl_Int2Flt(pAn.D));
        fmt.Printf("  ");
        fmt.Printf("Dp = %8s ", Nf_ManCell(p, pDp.Gate).pName);
        fmt.Printf("Dn = %8s ", Nf_ManCell(p, pDn.Gate).pName);
        fmt.Printf("Ap = %8s ", Nf_ManCell(p, pAp.Gate).pName);
        fmt.Printf("An = %8s ", Nf_ManCell(p, pAn.Gate).pName);
        fmt.Printf("\n");

    }
    }
}
func Nf_ManSetMapRefs(p *Nf_Man_t) int32 {
    var Coef float32  = 1.0 / (1.0 + (p.Iter + 1) * (p.Iter + 1));
    var pFlowRefs *float32  = Vec_FltArray( &p.vFlowRefs );
    var pMapRefs *int32  = Vec_IntArray( &p.vMapRefs );
    var nLits int32  = 2*Gia_ManObjNum(p.pGia);
    var i int32;
    var c int32;
    var Id int32;
    var nRefs [2]int32;
    var pObj *Gia_Obj_t;
    var pD, pA, pM *Nf_Mat_t;
    var pDs, pAs, pMs [2]*Nf_Mat_t;
    var Required int32;
    var Requireds [2]int32;
    assert(!p.fUseEla, "!p.fUseEla");
//    if ( p->Iter == 0 )
//        Nf_ManPrintMatches( p );
    Nf_ManSetOutputRequireds( p, 0 );
    // set output references
    memset( pMapRefs, 0, sizeof(int) * nLits );
    // inlining of macro Gia_ManForEachCo
    for  i = 0; (i < Vec_IntSize(p.pGia.vCos)) && ((pObj) = Gia_ManCo(p.pGia, i)); i++  {
        Nf_ObjMapRefInc( p, Gia_ObjFaninId0p(p.pGia, pObj), Gia_ObjFaninC0(pObj));
    }

    // compute area and edges
    p.nInvs = 0;
    p.pPars.MapAreaF = 0; 
    p.pPars.Area = p.pPars.Edge = 0;
    // inlining of macro Gia_ManForEachAndReverse
    for  i = p.pGia.nObjs - 1; (i > 0) && ((pObj) = Gia_ManObj(p.pGia, i)); i--  {  if !Gia_ObjIsAnd(pObj) {} else {
        if Gia_ObjIsBuf(pObj) {
            if Nf_ObjMapRefNum(p, i, 1) != 0 {
                Nf_ObjMapRefInc( p, i, 0 );
                Nf_ObjUpdateRequired( p, i, 0, Nf_ObjRequired(p, i, 1) - p.InvDelayI );
                p.pPars.MapAreaF += p.InvAreaF;
                p.pPars.Edge++;
                p.pPars.Area++;
                p.nInvs++;
            }
            Nf_ObjUpdateRequired( p, Gia_ObjFaninId0(pObj, i), Gia_ObjFaninC0(pObj), Nf_ObjRequired(p, i, 0) );
            Nf_ObjMapRefInc( p, Gia_ObjFaninId0(pObj, i), Gia_ObjFaninC0(pObj));
            continue;
        }
        // skip if this node is not used
        for  c = 0; c < 2; c++  {
            nRefs[c] = Nf_ObjMapRefNum(p, i, c);
        }
        if !nRefs[0] && !nRefs[1] {
            continue; }

        // consider two cases
        if nRefs[0] && nRefs[1] {
            // find best matches for both phases
            for  c = 0; c < 2; c++  {
                Requireds[c] = Nf_ObjRequired( p, i, c );
                //assert( Requireds[c] < SCL_INFINITY );
                pDs[c] = Nf_ObjMatchD( p, i, c );
                pAs[c] = Nf_ObjMatchA( p, i, c );
                pMs[c] = (pAs[c].D <= Requireds[c]) ? pAs[c] : pDs[c];
            }
            // swap complemented matches
            if pMs[0].fCompl && pMs[1].fCompl {
//                pMs[0]->fCompl = pMs[1]->fCompl = 0;
//                ABC_SWAP( Nf_Mat_t *, pMs[0], pMs[1] );
                // find best matches for both phases
                pMs[0] = Nf_ObjMatchD( p, i, 0 );
                pMs[1] = Nf_ObjMatchD( p, i, 1 );
                assert(!pMs[0].fCompl || !pMs[1].fCompl, "!pMs[0].fCompl || !pMs[1].fCompl");
            }
            // check if intervers are involved
            if !pMs[0].fCompl && !pMs[1].fCompl { // no inverters
            
                for  c = 0; c < 2; c++  {
                    Nf_ManSetMapRefsGate( p, i, Requireds[c], pMs[c] );
                }
            
            } else {
                // one interver
                assert(!pMs[0].fCompl || !pMs[1].fCompl, "!pMs[0].fCompl || !pMs[1].fCompl");
                c = pMs[1].fCompl;
                assert(pMs[c].fCompl && !pMs[!c].fCompl, "pMs[c].fCompl && !pMs[!c].fCompl");
                //printf( "Using inverter at node %d in phase %d\n", i, c );
                // update this phase
                pM = pMs[c];
                pM.fBest = 1;
                Required = Requireds[c];
                // update opposite phase
                Nf_ObjMapRefInc( p, i, !c );
                Nf_ObjUpdateRequired( p, i, !c, Required - p.InvDelayI );
                // select opposite phase
                Required = Nf_ObjRequired( p, i, !c );
                //assert( Required < SCL_INFINITY );
                pD = Nf_ObjMatchD( p, i, !c );
                pA = Nf_ObjMatchA( p, i, !c );
                if pA.D <= Required {
                    pM = pA;
                } else {
                    pM = pD;
                }
                assert(!pM.fCompl, "!pM.fCompl");
                // create gate
                Nf_ManSetMapRefsGate( p, i, Required, pM );
                // account for the inverter
                p.pPars.MapAreaF += p.InvAreaF;
                p.pPars.Edge++;
                p.pPars.Area++;
                p.nInvs++;
            }
        
        } else {
            c = int32((nRefs[1] > 0));
            assert(nRefs[c] && !nRefs[!c], "nRefs[c] && !nRefs[!c]");
            // consider this phase
            Required = Nf_ObjRequired( p, i, c );
            //assert( Required < SCL_INFINITY );
            pD = Nf_ObjMatchD( p, i, c );
            pA = Nf_ObjMatchA( p, i, c );
            if pA.D <= Required {
                pM = pA;
            } else {
                pM = pD;
            }
            if pM.fCompl { // use inverter
            
                p.nInvs++;
                //printf( "Using inverter at node %d in phase %d\n", i, c );
                pM.fBest = 1;
                // update opposite phase
                Nf_ObjMapRefInc( p, i, !c );
                Nf_ObjUpdateRequired( p, i, !c, Required - p.InvDelayI );
                // select opposite phase
                Required = Nf_ObjRequired( p, i, !c );
                //assert( Required < SCL_INFINITY );
                pD = Nf_ObjMatchD( p, i, !c );
                pA = Nf_ObjMatchA( p, i, !c );
                if pA.D <= Required {
                    pM = pA;
                } else {
                    pM = pD;
                }
                assert(!pM.fCompl, "!pM.fCompl");
                // account for the inverter
                p.pPars.MapAreaF += p.InvAreaF;
                p.pPars.Edge++;
                p.pPars.Area++;
            }
            // create gate
            Nf_ManSetMapRefsGate( p, i, Required, pM );
        }
        // the result of this:
        // - only one phase can be implemented as inverter of the other phase
        // - required times are propagated correctly
        // - references are set correctly
    }
    }
    // inlining of macro Gia_ManForEachCiId
    for  i = 0; (i < Vec_IntSize(p.pGia.vCis)) && ((Id) = Gia_ObjId(p.pGia, Gia_ManCi(p.pGia, i))); i++  {
        if Nf_ObjMapRefNum(p, Id, 1) != 0 {
            Nf_ObjMapRefInc( p, Id, 0 );
            Nf_ObjUpdateRequired( p, Id, 0, Required - p.InvDelayI );
            p.pPars.MapAreaF += p.InvAreaF;
            p.pPars.Edge++;
            p.pPars.Area++;
            p.nInvs++;
        }
    }
    // blend references
    for  i = 0; i < nLits; i++  {
        pFlowRefs[i] = Abc_MaxFloat(1.0, Coef * pFlowRefs[i] + (1.0 - Coef) * Abc_MaxFloat(1, pMapRefs[i]));
    }
//        pFlowRefs[i] = 0.2 * pFlowRefs[i] + 0.8 * Abc_MaxFloat(1, pMapRefs[i]);
    return p.pPars.Area;
}



/**Function*************************************************************

  Synopsis    [Area recovery.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Nf_MatchDeref_rec(p *Nf_Man_t, i int32, c int32, pM *Nf_Mat_t) word {
    var Area word;
    var k int32;
    var iVar int32;
    var fCompl int32;
    var pCut *int32;
    assert(pM.fBest, "pM.fBest");
    if pM.fCompl {
        assert(Nf_ObjMapRefNum(p, i, !c) > 0, "Nf_ObjMapRefNum(p, i, !c) > 0");
        if !(Nf_ObjMapRefDec(p, i, !c) != 0) {
            Area += Nf_MatchDeref_rec( p, i, !c, Nf_ObjMatchD(p, i, !c) ); }
        return Area + p.InvAreaW;
    }
    if Nf_ObjCutSetId(p, i) == 0 {
        return 0; }
    pCut = Nf_CutFromHandle( Nf_ObjCutSet(p, i), pM.CutH );
    // inlining of macro Nf_CutForEachVarCompl
    for  k = 0; k < Nf_CutSize(pCut) && (iVar = Nf_CutLeaves(pCut)[Nf_CfgVar(pM.Cfg, k)]) && ((fCompl = Nf_CfgCompl(pM.Cfg, k)), 1); k++  {
        assert(Nf_ObjMapRefNum(p, iVar, fCompl) > 0, "Nf_ObjMapRefNum(p, iVar, fCompl) > 0");
        if !(Nf_ObjMapRefDec(p, iVar, fCompl) != 0) {
            Area += Nf_MatchDeref_rec( p, iVar, fCompl, Nf_ObjMatchD(p, iVar, fCompl) ); }
    }
    return Area + Nf_ManCell(p, pM.Gate).AreaW;
}
func Nf_MatchRef_rec(p *Nf_Man_t, i int32, c int32, pM *Nf_Mat_t, Required int32, vBackup *Vec_Int_t) word {
    var Area word;
    var ReqFanin int32;
    var k int32;
    var iVar int32;
    var fCompl int32;
    var pCut *int32;
    assert(pM.fBest, "pM.fBest");
    assert(pM.D <= Required, "pM.D <= Required");
    if pM.fCompl {
        ReqFanin = Required - p.InvDelayI;
        if vBackup != nil {
            Vec_IntPush( vBackup, Abc_Var2Lit(i, !c) ); }
        assert(Nf_ObjMapRefNum(p, i, !c) >= 0, "Nf_ObjMapRefNum(p, i, !c) >= 0");
        if !(Nf_ObjMapRefInc(p, i, !c) != 0) {
            Area += Nf_MatchRef_rec( p, i, !c, Nf_ObjMatchD(p, i, !c), ReqFanin, vBackup ); }
        return Area + p.InvAreaW;
    }
    if Nf_ObjCutSetId(p, i) == 0 {
        return 0; }
    pCut = Nf_CutFromHandle( Nf_ObjCutSet(p, i), pM.CutH );
    // inlining of macro Nf_CutForEachVarCompl
    for  k = 0; k < Nf_CutSize(pCut) && (iVar = Nf_CutLeaves(pCut)[Nf_CfgVar(pM.Cfg, k)]) && ((fCompl = Nf_CfgCompl(pM.Cfg, k)), 1); k++  {
        ReqFanin = Required - Nf_ManCell(p, pM.Gate).iDelays[k];
        if vBackup != nil {
            Vec_IntPush( vBackup, Abc_Var2Lit(iVar, fCompl) ); }
        assert(Nf_ObjMapRefNum(p, iVar, fCompl) >= 0, "Nf_ObjMapRefNum(p, iVar, fCompl) >= 0");
        if !(Nf_ObjMapRefInc(p, iVar, fCompl) != 0) {
            Area += Nf_MatchRef_rec( p, iVar, fCompl, Nf_ObjMatchD(p, iVar, fCompl), ReqFanin, vBackup ); }
    }
    return Area + Nf_ManCell(p, pM.Gate).AreaW;
}
func Nf_MatchRefArea(p *Nf_Man_t, i int32, c int32, pM *Nf_Mat_t, Required int32) word {
    var Area word;  var iLit, k int32; 
    Vec_IntClear( &p.vBackup );
    Area = Nf_MatchRef_rec( p, i, c, pM, Required, &p.vBackup );
    // inlining of macro Vec_IntForEachEntry
    for  k = 0; (k < Vec_IntSize(&p.vBackup)) && (((iLit) = Vec_IntEntry(&p.vBackup, k)), 1); k++  {
        assert(Nf_ObjMapRefNum(p, Abc_Lit2Var(iLit), Abc_LitIsCompl(iLit)) > 0, "Nf_ObjMapRefNum(p, Abc_Lit2Var(iLit), Abc_LitIsCompl(iLit)) > 0");
        Nf_ObjMapRefDec( p, Abc_Lit2Var(iLit), Abc_LitIsCompl(iLit) );
    }
    return Area;
}
func Nf_ManElaBestMatchOne(p *Nf_Man_t, iObj int32, c int32, pCut *int32, pCutSet *int32, pRes *Nf_Mat_t, Required int32) {
    var Mb Nf_Mat_t;
    var pMb *Nf_Mat_t = &Mb;
    var pMd *Nf_Mat_t;
    var pFans *int32       = Nf_CutLeaves(pCut);
    var nFans int32         = Nf_CutSize(pCut);
    var iFuncLit int32      = Nf_CutFunc(pCut);
    var fComplExt int32     = Abc_LitIsCompl(iFuncLit);
    var vArr *Vec_Int_t  = Vec_WecEntry( p.vTt2Match, Abc_Lit2Var(iFuncLit) );
    var i, k, Info, Offset, iFanin, fComplF int32;
    // assign fanins matches
    var pBestF [NF_LEAF_MAX]*Nf_Obj_t;
    for  i = 0; i < nFans; i++  {
        pBestF[i] = Nf_ManObj( p, pFans[i] );
    }
    // consider matches of this function
    *pMb = Nf_Mat_t{};
    pMb.D = SCL_INFINITY; pMb.F = FLT_MAX;
    // special cases
    if nFans == 0 {
        var Const int32  = (iFuncLit == 1);
        //printf( "Node %d(%d) is const\n", iObj, c );
        assert(iFuncLit == 0 || iFuncLit == 1, "iFuncLit == 0 || iFuncLit == 1");
        pMb.D = 0;
        pMb.F = p.pCells[c ^ Const].AreaF;
        pMb.CutH = Nf_CutHandle(pCutSet, pCut);
        pMb.Gate = c ^ Const;
//        pMb->Conf = 0;
        pMb.Cfg = Nf_Int2Cfg(0);
        pMb.fBest = 1;
        // compare
        if pRes.F > pMb.F + NF_EPSILON || (pRes.F > pMb.F - NF_EPSILON && pRes.D > pMb.D) {
            *pRes = *pMb; }
        return;
    }
    // consider matches of this function
    // inlining of macro Vec_IntForEachEntryDouble
    for  i = 0; (i+1 < Vec_IntSize(vArr)) && (((Info) = Vec_IntEntry(vArr, i)), 1) && (((Offset) = Vec_IntEntry(vArr, i+1)), 1); i += 2  {
        var Cfg Nf_Cfg_t    = Nf_Int2Cfg(Offset);
        var pC *Mio_Cell2_t  = Nf_ManCell( p, Info );
        var fCompl int32      = Cfg.fCompl ^ fComplExt;
        var Delay int32;
        assert(nFans == int32(pC.nFanins), "nFans == int32(pC.nFanins)");
        if fCompl != c {
            continue; }
        // inlining of macro Nf_CfgForEachVarCompl
        for  k = 0; k < nFans && ((iFanin = Nf_CfgVar(Cfg, k)), 1) && ((fComplF = Nf_CfgCompl(Cfg, k)), 1); k++  {
            pMd     = &pBestF[iFanin].M[fComplF][0];
            assert(pMd.fBest, "pMd.fBest");
            Delay = Abc_MaxInt( Delay, pMd.D + pC.iDelays[k] );
            if Delay > Required {
                break; }
        }
        if k < nFans {
            continue; }
        // create match
        pMb.D = Delay;
        pMb.F = FLT_MAX;
        pMb.fBest = 1;
        pMb.fCompl = 0;
        pMb.CutH = Nf_CutHandle(pCutSet, pCut);
        pMb.Gate = pC.Id;
        pMb.Cfg = Cfg;
        pMb.Cfg.fCompl = 0;
        // compute area
        pMb.F = Scl_Int2Flt(int32(Nf_MatchRefArea(p, iObj, c, pMb, Required)));
        // compare
        if pRes.F > pMb.F + NF_EPSILON || (pRes.F > pMb.F - NF_EPSILON && pRes.D > pMb.D) {
            *pRes = *pMb; }
    }
}
func Nf_ManElaBestMatch(p *Nf_Man_t, iObj int32, c int32, pRes *Nf_Mat_t, Required int32) {
    var k int32;
    var pCut *int32;
    var pCutSet *int32 = Nf_ObjCutSet( p, iObj );
    *pRes = Nf_Mat_t{};
    pRes.D = SCL_INFINITY; pRes.F = FLT_MAX;
    // inlining of macro Nf_SetForEachCut
    for  k, pCut = 0, pCutSet + 1; k < pCutSet[0]; k++, pCut += Nf_CutSize(pCut) + 1  {
        if Abc_Lit2Var(Nf_CutFunc(pCut)) >= Vec_WecSize(p.vTt2Match) {
            continue; }
        Nf_ManElaBestMatchOne( p, iObj, c, pCut, pCutSet, pRes, Required );
    }
}
func Nf_ManComputeArrival(p *Nf_Man_t, pM *Nf_Mat_t, pCutSet *int32) int32 {
    var Delay int32; 
    var pMfan *Nf_Mat_t;
    var iVar, fCompl, k int32;
    var pCell *Mio_Cell2_t  = Nf_ManCell( p, pM.Gate );
    var pCut *int32  = Nf_CutFromHandle( pCutSet, pM.CutH );
    assert(!pM.fCompl, "!pM.fCompl");
    // inlining of macro Nf_CutForEachVarCompl
    for  k = 0; k < Nf_CutSize(pCut) && (iVar = Nf_CutLeaves(pCut)[Nf_CfgVar(pM.Cfg, k)]) && ((fCompl = Nf_CfgCompl(pM.Cfg, k)), 1); k++  {
        pMfan = Nf_ObjMatchBest( p, iVar, fCompl );
        Delay = Abc_MaxInt( Delay, pMfan.D + pCell.iDelays[k] );
    }
    //if ( pM->fCompl ) Delay += p->InvDelayI;
    return Delay;
}
func Nf_ManResetMatches(p *Nf_Man_t, Round int32) {
    var pObj *Gia_Obj_t;
    var pDc *Nf_Mat_t;
    var pAc *Nf_Mat_t;
    var pMfan *Nf_Mat_t;
    var pM [2]*Nf_Mat_t; 
    var i, c, Arrival int32;
    // go through matches in the topo order
    // inlining of macro Gia_ManForEachAnd
    for  i = 0; (i < p.pGia.nObjs) && ((pObj) = Gia_ManObj(p.pGia, i)); i++  {      if !Gia_ObjIsAnd(pObj) {} else {
        if Gia_ObjIsBuf(pObj) {
            pMfan = Nf_ObjMatchBest( p, Gia_ObjFaninId0(pObj, i), Gia_ObjFaninC0(pObj) );
            for  c = 0; c < 2; c++  {
                pDc = Nf_ObjMatchD( p, i, c );
                pAc = Nf_ObjMatchA( p, i, c );
                pDc.F = pAc.F = 0;
                pDc.D = pMfan.D + (c ? p.InvDelayI : 0); 
                assert(pDc.fBest, "pDc.fBest");
                assert(!pAc.fBest, "!pAc.fBest");
                assert(c==0 || pDc.fCompl, "c==0 || pDc.fCompl");
            }
            continue;
        }
        // select the best match for each phase
        for  c = 0; c < 2; c++  {
            pDc = Nf_ObjMatchD( p, i, c );
            pAc = Nf_ObjMatchA( p, i, c );
            pDc.F = pAc.F = 0;
            if Nf_ObjMapRefNum(p, i, c) != 0 {
                assert(pDc.fBest != pAc.fBest, "pDc.fBest != pAc.fBest");
                if pAc.fBest {
                    // inlining of macro ABC_SWAP
                     var t Nf_Mat_t  = *pDc; *pDc = *pAc; *pAc = t; };
                assert(pDc.fBest, "pDc.fBest");
                assert(!pAc.fBest, "!pAc.fBest");
            
            } else {
                assert(Round > 0 || (!pDc.fBest && !pAc.fBest), "Round > 0 || (!pDc.fBest && !pAc.fBest)");
//                if ( (p->pPars->fAreaOnly || (Round & 1)) && !pAc->fCompl )
                if (Round & 1) && !pAc.fCompl {
                    // inlining of macro ABC_SWAP
                     var t Nf_Mat_t  = *pDc; *pDc = *pAc; *pAc = t; };
                pDc.fBest = 1;
                pAc.fBest = 0;
            }
        }
        // consider best matches of both phases
        pM[0] = Nf_ObjMatchD( p, i, 0 );
        pM[1] = Nf_ObjMatchD( p, i, 1 );
        assert(pM[0].fBest && pM[1].fBest, "pM[0].fBest && pM[1].fBest");
        // swap complemented matches
        if pM[0].fCompl && pM[1].fCompl {
//            pM[0]->fCompl = pM[1]->fCompl = 0;
//            ABC_SWAP( Nf_Mat_t *, pM[0], pM[1] );
            assert(false, "0");
        }
        if !pM[0].fCompl && !pM[1].fCompl {
            for  c = 0; c < 2; c++  {
                Arrival = Nf_ManComputeArrival( p, pM[c], Nf_ObjCutSet(p, i) );
                //if ( Nf_ObjMapRefNum(p, i, c) )
                //    assert( Round || Arrival <= pM[c]->D );
                pM[c].D = Arrival;
            }
        
        } else {
            // consider non-complemented match
            c = !pM[1].fCompl;
            assert(!pM[c].fCompl, "!pM[c].fCompl");
            assert(pM[!c].fCompl, "pM[!c].fCompl");
            Arrival = Nf_ManComputeArrival( p, pM[c], Nf_ObjCutSet(p, i) );
            //if ( Nf_ObjMapRefNum(p, i, c) )
            //    assert( Round || Arrival <= pM[c]->D );
            pM[c].D = Arrival;
            // consider complemented match
            Arrival = pM[!c].D;
            *pM[!c] = *pM[c];
            pM[!c].D += p.InvDelayI;
            pM[!c].fCompl = 1;
            //if ( Nf_ObjMapRefNum(p, i, !c) )
            //    assert( Round || pM[!c]->D <= Arrival );
        }
    }
    }
}
func Nf_ManComputeMappingEla(p *Nf_Man_t) {
    var fVerbose int32;
    var pObj *Gia_Obj_t;
    var pCell *Mio_Cell2_t;
    var Mb Nf_Mat_t;
    var pMb *Nf_Mat_t = &Mb;
    var pM *Nf_Mat_t;
    var AreaBef, AreaAft, Gain word;
    var i int32;
    var c int32;
    var iVar int32;
    var Id int32;
    var fCompl int32;
    var k int32;
    var pCut *int32;
    var Required int32;
    Nf_ManSetOutputRequireds( p, 1 );
    Nf_ManResetMatches( p, p.Iter - p.pPars.nRounds );
    // inlining of macro Gia_ManForEachAndReverse
    for  i = p.pGia.nObjs - 1; (i > 0) && ((pObj) = Gia_ManObj(p.pGia, i)); i--  {  if !Gia_ObjIsAnd(pObj) {} else {
        if Gia_ObjIsBuf(pObj) {
            if Nf_ObjMapRefNum(p, i, 1) != 0 {
                Nf_ObjUpdateRequired( p, i, 0, Nf_ObjRequired(p, i, 1) - p.InvDelayI ); }
            Nf_ObjUpdateRequired( p, Gia_ObjFaninId0(pObj, i), Gia_ObjFaninC0(pObj), Nf_ObjRequired(p, i, 0) );
            continue;
        }
        for  c = 0; c < 2; c++  {
        if Nf_ObjMapRefNum(p, i, c) != 0 {
            pM = Nf_ObjMatchBest( p, i, c );
            Required = Nf_ObjRequired( p, i, c );
            assert(pM.D <= Required, "pM.D <= Required");
            if pM.fCompl {
                continue; }
            // search for a better match
            assert(!pM.fCompl, "!pM.fCompl");
            AreaBef = Nf_MatchDeref_rec( p, i, c, pM );
            assert(pM.fBest, "pM.fBest");
            Nf_ManElaBestMatch( p, i, c, pMb, Required );
            AreaAft = Nf_MatchRef_rec( p, i, c, pMb, Required, nil );
            Gain += AreaBef - AreaAft;
            // print area recover progress
            if (fVerbose != 0) && Nf_ManCell(p, pM.Gate).pName != Nf_ManCell(p, pMb.Gate).pName {
                fmt.Printf("%4d (%d)  ", i, c);
                fmt.Printf("%8s ->%8s  ", Nf_ManCell(p, pM.Gate).pName, Nf_ManCell(p, pMb.Gate).pName);
                fmt.Printf("%d -> %d  ", Nf_ManCell(p, pM.Gate).nFanins, Nf_ManCell(p, pMb.Gate).nFanins);
                fmt.Printf("D: %7.2f -> %7.2f  ", Scl_Int2Flt(pM.D), Scl_Int2Flt(pMb.D));
                fmt.Printf("R: %7.2f  ", Required == SCL_INFINITY ? 9999.99 : Scl_Int2Flt(Required));
                fmt.Printf("A: %7.2f -> %7.2f  ", Scl_Int2Flt(int32(AreaBef)), Scl_Int2Flt(int32(AreaAft)));
                fmt.Printf("G: %7.2f (%7.2f) ", Scl_Int2Flt(int32(AreaBef) - int32(AreaAft)), Scl_Int2Flt(int32(Gain)));
                fmt.Printf("\n");
            }
            // set best match
            assert(pMb.fBest, "pMb.fBest");
            assert(pMb.D <= Required, "pMb.D <= Required");
            //assert( Scl_Flt2Int(pMb->F) == (int)AreaAft );
            //assert( AreaBef >= AreaAft );
            *pM = *pMb;
            // update timing
            pCell = Nf_ManCell( p, pMb.Gate );
            pCut = Nf_CutFromHandle( Nf_ObjCutSet(p, i), pMb.CutH );
            // inlining of macro Nf_CutForEachVarCompl
            for  k = 0; k < Nf_CutSize(pCut) && (iVar = Nf_CutLeaves(pCut)[Nf_CfgVar(pMb.Cfg, k)]) && ((fCompl = Nf_CfgCompl(pMb.Cfg, k)), 1); k++  {
                pM = Nf_ObjMatchBest( p, iVar, fCompl );
                assert(pM.D <= Required - pCell.iDelays[k], "pM.D <= Required - pCell.iDelays[k]");
                Nf_ObjUpdateRequired( p, iVar, fCompl, Required - pCell.iDelays[k] );
                if pM.fCompl {
                    pM = Nf_ObjMatchBest( p, iVar, !fCompl );
                    assert(pM.D <= Required - pCell.iDelays[k] - p.InvDelayI, "pM.D <= Required - pCell.iDelays[k] - p.InvDelayI");
                    Nf_ObjUpdateRequired( p, iVar, !fCompl, Required - pCell.iDelays[k] - p.InvDelayI );
                }
            }
        }
        }
    }
    }
    // inlining of macro Gia_ManForEachCiId
    for  i = 0; (i < Vec_IntSize(p.pGia.vCis)) && ((Id) = Gia_ObjId(p.pGia, Gia_ManCi(p.pGia, i))); i++  {
        if Nf_ObjMapRefNum(p, Id, 1) != 0 {
            Required = Nf_ObjRequired( p, i, 1 );
            Nf_ObjUpdateRequired( p, Id, 0, Required - p.InvDelayI );
        }
    }
}
func Nf_ManFixPoDrivers(p *Nf_Man_t) {
    var pObj *Gia_Obj_t;
    var pM, pMc *Nf_Mat_t;
    var i, iDriver, Count int32;
    // inlining of macro Gia_ManForEachCo
    for  i = 0; (i < Vec_IntSize(p.pGia.vCos)) && ((pObj) = Gia_ManCo(p.pGia, i)); i++  {
        iDriver = Gia_ObjFaninId0p(p.pGia, pObj);
        if !Gia_ObjIsAnd(Gia_ManObj(p.pGia, iDriver)) {
            continue; }
        // skip unless both are used
        if !(Nf_ObjMapRefNum(p, iDriver, 0) != 0) || !(Nf_ObjMapRefNum(p, iDriver, 1) != 0) {
            continue; }
        pM  = Nf_ObjMatchD( p, iDriver,  Gia_ObjFaninC0(pObj) );
        pMc = Nf_ObjMatchD( p, iDriver, !Gia_ObjFaninC0(pObj) );
        // skip unless both are non-complemented
        if pM.fCompl || pMc.fCompl {
            continue; }
        // skip if arrival time exceeds the required time
        if pMc.D + p.InvDelayI > p.pPars.MapDelay {
            continue; }
        // update references
        Nf_MatchDeref_rec( p, iDriver, Gia_ObjFaninC0(pObj), pM );
        Nf_ObjMapRefInc( p, iDriver, !Gia_ObjFaninC0(pObj) );
        // add inverter
        *pM = *pMc;
        pM.D += p.InvDelayI;
        pM.fCompl = 1;
        pM.fBest = 1;
        pMc.fBest = 1;
        Count++;
    }
    //printf( "Fixed %d PO drivers.\n", Count );
}

/**Function*************************************************************

  Synopsis    [Deriving mapping.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Nf_ManDeriveMapping(p *Nf_Man_t) *Gia_Man_t {
    var vMapping *Vec_Int_t;
    var pM *Nf_Mat_t;
    var i int32;
    var k int32;
    var c int32;
    var Id int32;
    var iVar int32;
    var fCompl int32;
    var pCut *int32;
    assert(p.pGia.vCellMapping == nil, "p.pGia.vCellMapping == nil");
    vMapping = Vec_IntAlloc( 2*Gia_ManObjNum(p.pGia) + int32(p.pPars.Edge) + int32(p.pPars.Area) * 2 );
    Vec_IntFill( vMapping, 2*Gia_ManObjNum(p.pGia), 0 );
    // create CI inverters
    // inlining of macro Gia_ManForEachCiId
    for  i = 0; (i < Vec_IntSize(p.pGia.vCis)) && ((Id) = Gia_ObjId(p.pGia, Gia_ManCi(p.pGia, i))); i++  {
    if Nf_ObjMapRefNum(p, Id, 1) != 0 {
        Vec_IntWriteEntry( vMapping, Abc_Var2Lit(Id, 1), -1 ); }
    }
    // create internal nodes
    // inlining of macro Gia_ManForEachAndId
    for  i = 0; (i < p.pGia.nObjs); i++  {                                     if !Gia_ObjIsAnd(Gia_ManObj(p.pGia, i)) {} else {
        var pObj *Gia_Obj_t  = Gia_ManObj(p.pGia, i);
        if Gia_ObjIsBuf(pObj) {
            if Nf_ObjMapRefNum(p, i, 1) != 0 {
                Vec_IntWriteEntry( vMapping, Abc_Var2Lit(i, 1), -1 ); }
            Vec_IntWriteEntry( vMapping, Abc_Var2Lit(i, 0), -2 );
            continue;
        }
        for  c = 0; c < 2; c++  {
        if Nf_ObjMapRefNum(p, i, c) != 0 {
            pM = Nf_ObjMatchBest( p, i, c );
            // remember inverter
            if pM.fCompl {
                Vec_IntWriteEntry( vMapping, Abc_Var2Lit(i, c), -1 );
                continue;
            }
    //        Nf_ManCutMatchPrint( p, i, c, pM );
            pCut = Nf_CutFromHandle( Nf_ObjCutSet(p, i), pM.CutH );
            Vec_IntWriteEntry( vMapping, Abc_Var2Lit(i, c), Vec_IntSize(vMapping) );
            Vec_IntPush( vMapping, Nf_CutSize(pCut) );
            // inlining of macro Nf_CutForEachVarCompl
            for  k = 0; k < Nf_CutSize(pCut) && (iVar = Nf_CutLeaves(pCut)[Nf_CfgVar(pM.Cfg, k)]) && ((fCompl = Nf_CfgCompl(pM.Cfg, k)), 1); k++  {
                Vec_IntPush( vMapping, Abc_Var2Lit(iVar, fCompl) );
            }
            Vec_IntPush( vMapping, pM.Gate );
        }
        }
    }
    }
//    assert( Vec_IntCap(vMapping) == 16 || Vec_IntSize(vMapping) == Vec_IntCap(vMapping) );
    p.pGia.vCellMapping = vMapping;
    return p.pGia;
}
func Nf_ManUpdateStats(p *Nf_Man_t) {
    var pM *Nf_Mat_t;
    var pObj *Gia_Obj_t;
    var pCell *Mio_Cell2_t;
    var i int32;
    var c int32;
    var Id int32;
    var pCut *int32;
    p.pPars.MapAreaF = 0; p.nInvs = 0;
    p.pPars.Area = p.pPars.Edge = 0;
    // inlining of macro Gia_ManForEachAndReverse
    for  i = p.pGia.nObjs - 1; (i > 0) && ((pObj) = Gia_ManObj(p.pGia, i)); i--  {  if !Gia_ObjIsAnd(pObj) {} else {
        if Gia_ObjIsBuf(pObj) {
            if Nf_ObjMapRefNum(p, i, 1) != 0 {
                p.pPars.MapAreaF += p.InvAreaF;
                p.pPars.Edge++;
                p.pPars.Area++;
                p.nInvs++;
            }
            continue;
        }
        for  c = 0; c < 2; c++  {
        if Nf_ObjMapRefNum(p, i, c) != 0 {
            pM = Nf_ObjMatchBest( p, i, c );
            if pM.fCompl {
                p.pPars.MapAreaF += p.InvAreaF;
                p.pPars.Edge++;
                p.pPars.Area++;
                p.nInvs++;
                continue;
            }
            pCut = Nf_CutFromHandle( Nf_ObjCutSet(p, i), pM.CutH );
            pCell = Nf_ManCell( p, pM.Gate );
            assert(Nf_CutSize(pCut) == int32(pCell.nFanins), "Nf_CutSize(pCut) == int32(pCell.nFanins)");
            p.pPars.MapAreaF += pCell.AreaF;
            p.pPars.Edge += Nf_CutSize(pCut);
            p.pPars.Area++;
            //printf( "%5d (%d) : Gate = %7s \n", i, c, pCell->pName );
        }
        }
    }
    }
    // inlining of macro Gia_ManForEachCiId
    for  i = 0; (i < Vec_IntSize(p.pGia.vCis)) && ((Id) = Gia_ObjId(p.pGia, Gia_ManCi(p.pGia, i))); i++  {
        if Nf_ObjMapRefNum(p, Id, 1) != 0 {
            p.pPars.MapAreaF += p.InvAreaF;
            p.pPars.Edge++;
            p.pPars.Area++;
            p.nInvs++;
        }
    }
}

/**Function*************************************************************

  Synopsis    [Extract window.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/

/*
    int              nInputs;   // the number of inputs
    int              nObjs;     // number of all objects
    Vec_Int_t *      vRoots;    // output drivers to be mapped (root -> obj lit)
    Vec_Wec_t *      vCuts;     // cuts (cut -> obj lit + fanin lits)
    Vec_Wec_t *      vObjCuts;  // cuts (obj lit -> obj lit + cut lits)
    Vec_Int_t *      vSolCuts;  // current solution (index -> cut)
    Vec_Int_t *      vCutGates; // gates (cut -> gate)
*/

func Nf_ManExtractWindow(pMan any, vRoots *Vec_Int_t, vCuts *Vec_Wec_t, vObjCuts *Vec_Wec_t, vSolCuts *Vec_Int_t, vCutGates *Vec_Int_t, vCutAreas *Vec_Wrd_t, pInvArea *word, StartVar int32, nVars int32) int32 {
    var p *Nf_Man_t  = (Nf_Man_t *)pMan;
    var nInputs int32  = Gia_ManCiNum(p.pGia);
    var LitShift int32  = 2*nInputs+2;
    var pObj *Gia_Obj_t;
    var c, iObj int32;
    if 2*Gia_ManAndNum(p.pGia) + Gia_ManCiNum(p.pGia) > nVars {
        fmt.Printf("The number of variables is too large: 2*%d + %d = %d > %d.\n", Gia_ManAndNum(p.pGia), Gia_ManCiNum(p.pGia), 2*Gia_ManAndNum(p.pGia) + Gia_ManCiNum(p.pGia), nVars);
        return 0;
    }
    *pInvArea = p.InvAreaW;
    // save roots
    Vec_IntClear( vRoots );
    // inlining of macro Gia_ManForEachCo
    for  c = 0; (c < Vec_IntSize(p.pGia.vCos)) && ((pObj) = Gia_ManCo(p.pGia, c)); c++  {
        assert(!Gia_ObjIsCi(Gia_ObjFanin0(pObj)), "!Gia_ObjIsCi(Gia_ObjFanin0(pObj))");
        Vec_IntPush( vRoots, Gia_ObjFaninLit0p(p.pGia, pObj)-LitShift );
    }
    // prepare
    Vec_WecClear( vCuts );
    Vec_WecClear( vObjCuts );
    Vec_IntClear( vSolCuts );
    Vec_IntClear( vCutGates );
    Vec_WrdClear( vCutAreas );
    // collect cuts for each node
    // inlining of macro Gia_ManForEachAndId
    for  iObj = 0; (iObj < p.pGia.nObjs); iObj++  {                                     if !Gia_ObjIsAnd(Gia_ManObj(p.pGia, iObj)) {} else {
        var vObj [2]*Vec_Int_t;
        var vCutOne *Vec_Int_t;
        var iCut int32;
        var pCut *int32;
        var pCutSet *int32;
        var iCutInv [2]int32 = [2]int32{-1, -1};
        // get matches
        var pM [2]*Nf_Mat_t = [2]*Nf_Mat_t{nil, nil};
        for  c = 0; c < 2; c++  {
            if Nf_ObjMapRefNum(p, iObj, c) == 0 {
                continue; }
            if Nf_ObjMatchBest(p, iObj, c).fCompl {
                assert(iCutInv[c] == -1, "iCutInv[c] == -1");
                iCutInv[c] = Vec_IntSize(vSolCuts);
                Vec_IntPush( vSolCuts, -1 );
                continue;
            }
            pM[c] = Nf_ObjMatchBest(p, iObj, c);
        }
        // start collecting cuts of pos-obj and neg-obj
        assert(Vec_WecSize(vObjCuts) == 2*iObj-LitShift, "Vec_WecSize(vObjCuts) == 2*iObj-LitShift");
        for  c = 0; c < 2; c++  {
            vObj[c] = Vec_WecPushLevel( vObjCuts );
            Vec_IntPush( vObj[c], Abc_Var2Lit(Abc_Var2Lit(iObj, c)-LitShift, 1) );
        }
        // enumerate cuts
        pCutSet = Nf_ObjCutSet( p, iObj );
        // inlining of macro Nf_SetForEachCut
        for  iCut, pCut = 0, pCutSet + 1; iCut < pCutSet[0]; iCut++, pCut += Nf_CutSize(pCut) + 1  {
            assert(!(Nf_CutIsTriv(pCut, iObj) != 0), "!Nf_CutIsTriv(pCut, iObj)");
            assert(Nf_CutSize(pCut) <= p.pPars.nLutSize, "Nf_CutSize(pCut) <= p.pPars.nLutSize");
            if Abc_Lit2Var(Nf_CutFunc(pCut)) < Vec_WecSize(p.vTt2Match) {
                var pFans *int32       = Nf_CutLeaves(pCut);
                var nFans int32         = Nf_CutSize(pCut);
                var iFuncLit int32      = Nf_CutFunc(pCut);
                var fComplExt int32     = Abc_LitIsCompl(iFuncLit);
                var vArr *Vec_Int_t  = Vec_WecEntry( p.vTt2Match, Abc_Lit2Var(iFuncLit) );
                var i, k, c, Info, Offset, iFanin, fComplF, iCutLit int32;
                // inlining of macro Vec_IntForEachEntryDouble
                for  i = 0; (i+1 < Vec_IntSize(vArr)) && (((Info) = Vec_IntEntry(vArr, i)), 1) && (((Offset) = Vec_IntEntry(vArr, i+1)), 1); i += 2  {
                    var Cfg Nf_Cfg_t    = Nf_Int2Cfg(Offset);
                    var fCompl int32      = Cfg.fCompl ^ fComplExt;
                    var pC *Mio_Cell2_t  = Nf_ManCell( p, Info );
                    assert(nFans == int32(pC.nFanins), "nFans == int32(pC.nFanins)");
                    Vec_IntPush( vCutGates, Info );
                    Vec_WrdPush( vCutAreas, pC.AreaW );
                    // to make comparison possible
                    Cfg.fCompl = 0;
                    // add solution cut
                    for  c = 0; c < 2; c++  {
                        if pM[c] == nil {
                            continue; }
                        if int32(pM[c].CutH) == Nf_CutHandle(pCutSet, pCut) && int32(pM[c].Gate) == Info && Nf_Cfg2Int(pM[c].Cfg) == Nf_Cfg2Int(Cfg) {
                            Vec_IntPush( vSolCuts, Vec_WecSize(vCuts) );
                            //printf( "adding solution for %d\n", Abc_Var2Lit(iObj, c)-LitShift );
                        }
                    }
                    // add new cut
                    iCutLit = Abc_Var2Lit( StartVar + Vec_WecSize(vCuts), 0 );
                    vCutOne = Vec_WecPushLevel( vCuts );
                    // add literals
                    Vec_IntPush( vCutOne, Abc_Var2Lit(iObj, fCompl) );
                    Vec_IntPush( vObj[fCompl], iCutLit );
                    // inlining of macro Nf_CfgForEachVarCompl
                    for  k = 0; k < nFans && ((iFanin = Nf_CfgVar(Cfg, k)), 1) && ((fComplF = Nf_CfgCompl(Cfg, k)), 1); k++  {
                        if pFans[iFanin] >= nInputs + 1 { // internal node
                        
                            Vec_IntPush( vCutOne, Abc_Var2Lit(pFans[iFanin], fComplF) );
                            //Vec_IntPush( Vec_WecEntry(vObjCuts, Abc_Var2Lit(pFans[iFanin], fComplF)-LitShift), iCutLit );
                        
                        } else if fComplF != 0 { // complemented primary input
                            Vec_IntPush( vCutOne, Abc_Var2Lit(pFans[iFanin], 1) ); }
                    }
                } 
            }
        }
        assert(iCutInv[0] == -1 || iCutInv[1] == -1, "iCutInv[0] == -1 || iCutInv[1] == -1");
        // add inverter cut
        for  c = 0; c < 2; c++  {
            if iCutInv[c] != -1 {
                Vec_IntWriteEntry( vSolCuts, iCutInv[c], Vec_WecSize(vCuts) ); }
            // the obj-lit implies its cut
            Vec_IntPush( Vec_WecEntry(vObjCuts, Abc_Var2Lit(iObj, c)-LitShift), Abc_Var2Lit(StartVar + Vec_WecSize(vCuts), 0) );
            // the cut includes both literals
            vCutOne = Vec_WecPushLevel( vCuts );
            Vec_IntPush( vCutOne, Abc_Var2Lit(iObj, c) );
            Vec_IntPush( vCutOne, Abc_Var2Lit(iObj, !c) );
            Vec_IntPush( vCutGates, 3 );
            Vec_WrdPush( vCutAreas, p.InvAreaW );
        }
    }
    }
//    for ( c = 0; c < p->nCells; c++ )
//        printf( "%d=%s ", c, p->pCells[c].pName );
//    printf( "\n" );
    // add complemented inputs
    // inlining of macro Gia_ManForEachCiId
    for  c = 0; (c < Vec_IntSize(p.pGia.vCis)) && ((iObj) = Gia_ObjId(p.pGia, Gia_ManCi(p.pGia, c))); c++  {
        if Nf_ObjMapRefNum(p, iObj, 1) != 0 {
            Vec_IntPush( vSolCuts, -(2*Gia_ManAndNum(p.pGia)+c) ); }
    }
    assert(Vec_WecSize(vCuts) == Vec_IntSize(vCutGates), "Vec_WecSize(vCuts) == Vec_IntSize(vCutGates)");
    assert(Vec_WecSize(vCuts) == Vec_WrdSize(vCutAreas), "Vec_WecSize(vCuts) == Vec_WrdSize(vCutAreas)");
    assert(Vec_WecSize(vObjCuts) == 2*Gia_ManAndNum(p.pGia), "Vec_WecSize(vObjCuts) == 2*Gia_ManAndNum(p.pGia)");
    return nInputs;
}

/**Function*************************************************************

  Synopsis    [Technology mappping.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Nf_ManSetDefaultPars(pPars *Jf_Par_t) {
    memset( pPars, 0, sizeof(Jf_Par_t) );
    pPars.nLutSize     =  6;
    pPars.nCutNum      = 16;
    pPars.nProcNum     =  0;
    pPars.nRounds      =  4;
    pPars.nRoundsEla   =  2;
    pPars.nRelaxRatio  =  0;
    pPars.nCoarseLimit =  3;
    pPars.nAreaTuner   =  0;
    pPars.nReqTimeFlex =  0;
    pPars.nVerbLimit   =  5;
    pPars.DelayTarget  = -1;
    pPars.fAreaOnly    =  0;
    pPars.fPinPerm     =  0;
    pPars.fPinQuick    =  0;
    pPars.fPinFilter   =  0;
    pPars.fOptEdge     =  1; 
    pPars.fCoarsen     =  0;
    pPars.fCutMin      =  1;
    pPars.fGenCnf      =  0;
    pPars.fPureAig     =  0;
    pPars.fVerbose     =  0;
    pPars.fVeryVerbose =  0;
    pPars.nLutSizeMax  =  NF_LEAF_MAX;
    pPars.nCutNumMax   =  NF_CUT_MAX;
    pPars.MapDelayTarget = 0;
}
func Nf_ManPerformMapping(pGia *Gia_Man_t, pPars *Jf_Par_t) *Gia_Man_t {
    var pNew *Gia_Man_t = nil;
    var pCls *Gia_Man_t;
    var p *Nf_Man_t; var i, Id int32;
    if Gia_ManHasChoices(pGia) {
        pPars.fCoarsen = 0; } 
    if pPars.fCoarsen {
        pCls = Gia_ManDupMuxes(pGia, pPars.nCoarseLimit);
    } else {
        pCls = pGia;
    }
    p = Nf_StoCreate( pCls, pPars );
    if p == nil {
        return nil; }
//    if ( pPars->fVeryVerbose )
//        Nf_StoPrint( p, pPars->fVeryVerbose );
    if pPars.fVerbose && pPars.fCoarsen {
        fmt.Printf("Initial ");  Gia_ManPrintMuxStats( pGia );  fmt.Printf("\n");
        fmt.Printf("Derived ");  Gia_ManPrintMuxStats( pCls );  fmt.Printf("\n");
    }
    Nf_ManPrintInit( p );
    Nf_ManComputeCuts( p );
    Nf_ManPrintQuit( p );
    if Scl_ConIsRunning() {
        // inlining of macro Gia_ManForEachCiId
        for  i = 0; (i < Vec_IntSize(p.pGia.vCis)) && ((Id) = Gia_ObjId(p.pGia, Gia_ManCi(p.pGia, i))); i++  {
            Nf_ObjPrepareCi( p, Id, Scl_ConGetInArr(i) );
        }
    
    } else {
        // inlining of macro Gia_ManForEachCiId
        for  i = 0; (i < Vec_IntSize(p.pGia.vCis)) && ((Id) = Gia_ObjId(p.pGia, Gia_ManCi(p.pGia, i))); i++  {
//            Nf_ObjPrepareCi( p, Id, Scl_Flt2Int(p->pGia->vInArrs ? Abc_MaxFloat(0.0, Vec_FltEntry(p->pGia->vInArrs, i)) : 0.0) );
            Nf_ObjPrepareCi( p, Id, Scl_Flt2Int(p.pGia.vInArrs ? Vec_FltEntry(p.pGia.vInArrs, i) : 0.0) );
        }
    }
    for  p.Iter = 0; p.Iter < p.pPars.nRounds; p.Iter++  {
        Nf_ManComputeMapping( p );
        Nf_ManSetMapRefs( p );
        Nf_ManPrintStats( p, (char *)(p.Iter ? "Area " : "Delay") );
    }

    p.fUseEla = 1;
    for  ; p.Iter < p.pPars.nRounds + pPars.nRoundsEla; p.Iter++  {
        Nf_ManComputeMappingEla( p );
        Nf_ManUpdateStats( p );
        Nf_ManPrintStats( p, "Ela  " );
    }
    Nf_ManFixPoDrivers( p );
    pNew = Nf_ManDeriveMapping( p );
/*
    if ( pPars->fAreaOnly )
    {
        int Sbm_ManTestSat( void * pMan );
        Sbm_ManTestSat( p );
    }
*/
    Nf_StoDelete( p );
    return pNew;
}

////////////////////////////////////////////////////////////////////////
///                       END OF FILE                                ///
////////////////////////////////////////////////////////////////////////


//ABC_NAMESPACE_IMPL_END

