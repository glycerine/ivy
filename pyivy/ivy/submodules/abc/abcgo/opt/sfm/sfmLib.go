package sfm

import "fmt"

/**CFile****************************************************************

  FileName    [sfmLib.c]

  SystemName  [ABC: Logic synthesis and verification system.]

  PackageName [SAT-based optimization using internal don't-cares.]

  Synopsis    [Preprocessing genlib library.]

  Author      [Alan Mishchenko]
  
  Affiliation [UC Berkeley]

  Date        [Ver. 1.0. Started - June 20, 2005.]

  Revision    [$Id: sfmLib.c,v 1.00 2005/06/20 00:00:00 alanmi Exp $]

***********************************************************************/

//#include "sfmInt.h"
//#include "misc/st/st.h"
//#include "map/mio/mio.h"
//#include "misc/vec/vecMem.h"
//#include "misc/util/utilTruth.h"
//#include "misc/extra/extra.h"
//#include "map/mio/exp.h"
//#include "opt/dau/dau.h"
//#include "base/main/main.h"

//ABC_NAMESPACE_IMPL_START


////////////////////////////////////////////////////////////////////////
///                        DECLARATIONS                              ///
////////////////////////////////////////////////////////////////////////

type Sfm_Fun_t_ struct {
    Next             int32;          // next function in the list
    Area             int32;          // area of this function
    pFansT [SFM_SUPP_MAX+1]byte;     // top gate ID, followed by fanin perm
    pFansB [SFM_SUPP_MAX+1]byte;     // bottom gate ID, followed by fanin perm
}
type Sfm_Lib_t_ struct {
    nVars             int32;         // variable count
    nWords             int32;        // truth table words
    fVerbose             int32;      // verbose statistics
    pCells    *Mio_Cell2_t;        // library gates
    nCells             int32;        // library gate count
    fDelay             int32;        // uses delay profile
    nObjs             int32;         // object count
    nObjsAlloc             int32;    // object count
    pObjs      *Sfm_Fun_t;         // objects  
    vTtMem      *Vec_Mem_t;        // truth tables
    vLists       Vec_Int_t;        // lists of funcs for each truth table
    vCounts       Vec_Int_t;       // counters of functions for each truth table
    vHits       Vec_Int_t;         // the number of times this function was used
    vProfs       Vec_Int_t;        // area/delay profiles
    vStore       Vec_Int_t;        // storage for area/delay profiles
    vTemp       Vec_Int_t;         // temporary storage for candidates
    nObjSkipped             int32;
    nObjRemoved             int32;
}

func Sfm_LibFun(p *Sfm_Lib_t, i int32) *Sfm_Fun_t { if i == -1 {
    return nil;
} else {
    return p.pObjs + i;
} }
func Sfm_LibFunId(p *Sfm_Lib_t, pFun *Sfm_Fun_t) int32 { return pFun - p.pObjs;               }

// Original macro:
// #define Sfm_LibForEachSuper( p, pObj, Func ) \
//     for ( pObj = Sfm_LibFun(p, Vec_IntEntry(&p->vLists, Func)); pObj; pObj = Sfm_LibFun(p, pObj->Next) )



var s_Truth8 [8][4]word = [8][4]word{
    { ABC_CONST(0xAAAAAAAAAAAAAAAA),ABC_CONST(0xAAAAAAAAAAAAAAAA),ABC_CONST(0xAAAAAAAAAAAAAAAA),ABC_CONST(0xAAAAAAAAAAAAAAAA) },
    { ABC_CONST(0xCCCCCCCCCCCCCCCC),ABC_CONST(0xCCCCCCCCCCCCCCCC),ABC_CONST(0xCCCCCCCCCCCCCCCC),ABC_CONST(0xCCCCCCCCCCCCCCCC) },
    { ABC_CONST(0xF0F0F0F0F0F0F0F0),ABC_CONST(0xF0F0F0F0F0F0F0F0),ABC_CONST(0xF0F0F0F0F0F0F0F0),ABC_CONST(0xF0F0F0F0F0F0F0F0) },
    { ABC_CONST(0xFF00FF00FF00FF00),ABC_CONST(0xFF00FF00FF00FF00),ABC_CONST(0xFF00FF00FF00FF00),ABC_CONST(0xFF00FF00FF00FF00) },
    { ABC_CONST(0xFFFF0000FFFF0000),ABC_CONST(0xFFFF0000FFFF0000),ABC_CONST(0xFFFF0000FFFF0000),ABC_CONST(0xFFFF0000FFFF0000) },
    { ABC_CONST(0xFFFFFFFF00000000),ABC_CONST(0xFFFFFFFF00000000),ABC_CONST(0xFFFFFFFF00000000),ABC_CONST(0xFFFFFFFF00000000) },
    { ABC_CONST(0x0000000000000000),ABC_CONST(0xFFFFFFFFFFFFFFFF),ABC_CONST(0x0000000000000000),ABC_CONST(0xFFFFFFFFFFFFFFFF) },
    { ABC_CONST(0x0000000000000000),ABC_CONST(0x0000000000000000),ABC_CONST(0xFFFFFFFFFFFFFFFF),ABC_CONST(0xFFFFFFFFFFFFFFFF) },
};

////////////////////////////////////////////////////////////////////////
///                     FUNCTION DEFINITIONS                         ///
////////////////////////////////////////////////////////////////////////

/**Function*************************************************************

  Synopsis    []

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/


/**Function*************************************************************

  Synopsis    [Preprocess the library.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/



/**Function*************************************************************

  Synopsis    []

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/



/**Function*************************************************************

  Synopsis    []

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Sfm_LibStart(nVars int32, fDelay int32, fVerbose int32) *Sfm_Lib_t {
    // inlining of macro ABC_CALLOC
    var p *Sfm_Lib_t  = &Sfm_Lib_t{};
    assert(nVars <= SFM_SUPP_MAX, "nVars <= SFM_SUPP_MAX");
    p.vTtMem = Vec_MemAllocForTT( nVars, 0 );   
    Vec_IntGrow( &p.vLists,  (1 << 16) );
    Vec_IntGrow( &p.vCounts, (1 << 16) );
    Vec_IntGrow( &p.vHits,   (1 << 16) );
    Vec_IntFill( &p.vLists,  2, -1 );
    Vec_IntFill( &p.vCounts, 2, -1 );
    Vec_IntFill( &p.vHits,   2, -1 );
    p.nObjsAlloc = (1 << 16);
    // inlining of macro ABC_CALLOC
    p.pObjs = ((Sfm_Fun_t *) calloc((size_t)(p.nObjsAlloc), sizeof(Sfm_Fun_t)));
    p.fDelay = fDelay;
    if fDelay != 0 { Vec_IntGrow( &p.vProfs,  (1 << 16) ); }
    if fDelay != 0 { Vec_IntGrow( &p.vStore,  (1 << 18) ); }
    Vec_IntGrow( &p.vTemp, 16 );
    p.nVars = nVars;
    p.nWords = Abc_TtWordNum( nVars );
    p.fVerbose = fVerbose;
    return p;
}
func Sfm_LibStop(p *Sfm_Lib_t) {
    Vec_MemHashFree( p.vTtMem );
    Vec_MemFree( p.vTtMem );
    Vec_IntErase( &p.vLists );
    Vec_IntErase( &p.vCounts );
    Vec_IntErase( &p.vHits );
    Vec_IntErase( &p.vProfs );
    Vec_IntErase( &p.vStore );
    Vec_IntErase( &p.vTemp );
    p.pCells = nil;
    p.pObjs = nil;
    p = nil;
}

/**Function*************************************************************

  Synopsis    []

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/



/**Function*************************************************************

  Synopsis    []

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
/*
void Sfm_LibCellProfile( Mio_Cell2_t * pCellBot, Mio_Cell2_t * pCellTop, int InTop, int nFanins, int * Perm, int * pProf )
{
    int i, DelayAdd = pCellTop ? pCellTop->iDelays[InTop] : 0;
    for ( i = 0; i < nFanins; i++ )
        if ( Perm[i] < (int)pCellBot->nFanins )
            pProf[i] = pCellBot->iDelays[Perm[i]] + DelayAdd;
        else if ( Perm[i] < (int)pCellBot->nFanins + InTop )
            pProf[i] = pCellTop->iDelays[Perm[i] - (int)pCellBot->nFanins];
        else // if ( Perm[i] >= (int)pCellBot->nFanins + InTop )
            pProf[i] = pCellTop->iDelays[Perm[i] - (int)pCellBot->nFanins + 1];
}
*/



func Sfm_LibPrepareAdd(p *Sfm_Lib_t, pTruth *word, Perm *int32, nFanins int32, pCellBot *Mio_Cell2_t, pCellTop *Mio_Cell2_t, InTop int32) {
    var pObj *Sfm_Fun_t;
    var InvPerm, Profile [SFM_SUPP_MAX]int32;
    var Area int32  = int32(pCellBot.AreaW) + (pCellTop ? int32(pCellTop.AreaW) : 0);
    var i int32;
    var k int32;
    var Id int32;
    var Prev int32;
    var Offset int32;
    var pProf *int32;
    var iFunc int32 = Vec_MemHashInsert( p.vTtMem, pTruth );
    if iFunc == Vec_IntSize(&p.vLists) {
        Vec_IntPush( &p.vLists, -1 );
        Vec_IntPush( &p.vCounts, 0 );
        Vec_IntPush( &p.vHits,   0 );
    }
    assert(pCellBot != nil, "pCellBot != nil");
    // iterate through the supergates of this truth table
    if p.fDelay {
        assert(Vec_IntSize(&p.vProfs) == p.nObjs, "Vec_IntSize(&p.vProfs) == p.nObjs");
        Sfm_LibCellProfile( pCellBot, pCellTop, InTop, nFanins, Perm, Profile );
        // check if new one is contained in old ones
        Vec_IntClear( &p.vTemp );
        // inlining of macro Sfm_LibForEachSuper
        for  pObj = Sfm_LibFun(p, Vec_IntEntry(&p.vLists, iFunc)); (pObj != nil); pObj = Sfm_LibFun(p, pObj.Next)  {
            Vec_IntPush( &p.vTemp, Sfm_LibFunId(p, pObj) );
            Offset = Vec_IntEntry( &p.vProfs, Sfm_LibFunId(p, pObj) );
            pProf  = Vec_IntEntryP( &p.vStore, Offset );
            if Sfm_LibNewIsContained(pObj, pProf, Area, Profile, nFanins) != 0 {
                p.nObjSkipped++;
                return;
            }
        }
        // check if old ones are contained in new one
        k = 0;
        // inlining of macro Vec_IntForEachEntry
        for  i = 0; (i < Vec_IntSize(&p.vTemp)) && (((Id) = Vec_IntEntry(&p.vTemp, i)), 1); i++  {
            Offset = Vec_IntEntry( &p.vProfs, Id );
            pProf  = Vec_IntEntryP( &p.vStore, Offset );
            if !(Sfm_LibNewContains(Sfm_LibFun(p, Id), pProf, Area, Profile, nFanins) != 0) {
                Vec_IntWriteEntry( &p.vTemp, k++, Id );
            } else {
                p.nObjRemoved++; }
        }
        if k < i { // change
        
            if k == 0 {
                Vec_IntWriteEntry( &p.vLists, iFunc, -1 );
            } else {
                Vec_IntShrink( &p.vTemp, k );
                Prev = Vec_IntEntry(&p.vTemp, 0);
                Vec_IntWriteEntry( &p.vLists, iFunc, Prev );
                // inlining of macro Vec_IntForEachEntryStart
                for  i = 1; (i < Vec_IntSize(&p.vTemp)) && (((Id) = Vec_IntEntry(&p.vTemp, i)), 1); i++  {
                    Sfm_LibFun(p, Prev).Next = Id;
                    Prev = Id;
                }
                Sfm_LibFun(p, Prev).Next = -1;
            }
        }
    
    } else {
        // inlining of macro Sfm_LibForEachSuper
        for  pObj = Sfm_LibFun(p, Vec_IntEntry(&p.vLists, iFunc)); (pObj != nil); pObj = Sfm_LibFun(p, pObj.Next)  {
            if Area >= pObj.Area {
                return; }
        }
    }
    for  k = 0; k < nFanins; k++  {
        InvPerm[Perm[k]] = k;
    }
    // create delay profile
    if p.fDelay {
        Vec_IntPush( &p.vProfs, Vec_IntSize(&p.vStore) );
        for  k = 0; k < nFanins; k++  {
            Vec_IntPush( &p.vStore, Profile[k] );
        }
    }
    // create new object
    if p.nObjs == p.nObjsAlloc {
        var nObjsAlloc int32  = 2 * p.nObjsAlloc;
        // inlining of macro ABC_REALLOC
        
        if p.pObjs {
            p.pObjs = ((Sfm_Fun_t *) realloc((char *)(p.pObjs), sizeof(Sfm_Fun_t) * (size_t)(nObjsAlloc)));
        } else {
            p.pObjs = ((Sfm_Fun_t *) malloc(sizeof(Sfm_Fun_t) * (size_t)(nObjsAlloc)));
        }
        memset( p.pObjs + p.nObjsAlloc, 0, sizeof(Sfm_Fun_t) * p.nObjsAlloc );
        p.nObjsAlloc = nObjsAlloc;
    }
    pObj = p.pObjs + p.nObjs;
    pObj.Area = Area;
    pObj.Next = Vec_IntEntry(&p.vLists, iFunc);
    Vec_IntWriteEntry( &p.vLists, iFunc, p.nObjs++ );
    Vec_IntAddToEntry( &p.vCounts, iFunc, 1 );
    // create gate
    assert(pCellBot.Id < 128, "pCellBot.Id < 128");
    pObj.pFansB[0] = byte(pCellBot.Id);
    for  k = 0; k < int32(pCellBot.nFanins); k++  {
        pObj.pFansB[k+1] = InvPerm[k];
    }
    if pCellTop == nil {
        return; }
    assert(pCellTop.Id < 128, "pCellTop.Id < 128");
    pObj.pFansT[0] = byte(pCellTop.Id);
    for  i = 0; i < int32(pCellTop.nFanins); i++  {
        pObj.pFansT[i+1] = byte((i == InTop ? 16 : InvPerm[k++]));
    }
    assert(k == nFanins, "k == nFanins");
}
func Sfm_LibPrepare(nVars int32, fTwo int32, fDelay int32, fVerbose int32, fLibVerbose int32) *Sfm_Lib_t {
    var clk abctime  = Abc_Clock();
    var p *Sfm_Lib_t  = Sfm_LibStart( nVars, fDelay, fLibVerbose );
    var pCell1, pCell2, pLimit *Mio_Cell2_t;
    var pPerm [SFM_SUPP_MAX+1]*int32;
    var Perm1 *int32;
    var Perm2 *int32;
    var Perm [SFM_SUPP_MAX]int32;
    var nPerms [SFM_SUPP_MAX+1]int32;
    var i int32;
    var f int32;
    var n int32;
    var tTemp1, tCur [4]word;
    var pRes [1000]byte;
    assert(nVars <= SFM_SUPP_MAX, "nVars <= SFM_SUPP_MAX");
    // precompute gates
    p.pCells = Mio_CollectRootsNewDefault2( Abc_MinInt(6, nVars), &p.nCells, 0 );
    pLimit = p.pCells + p.nCells;
    // find useful ones
    for  pCell1 = p.pCells + 4; pCell1 < pLimit; pCell1++  {
        var uTruth word  = pCell1.uTruth;
        pCell1.Type = 0;
        if Abc_Tt6IsAndType(uTruth, pCell1.nFanins) || Abc_Tt6IsOrType(uTruth, pCell1.nFanins) {
            pCell1.Type = 1;
        } else if Dau_DsdDecompose(&uTruth, pCell1.nFanins, 0, 0, pRes) <= 3 {
            pCell1.Type = 2;
        } else if fLibVerbose != 0 {
            fmt.Printf("Skipping gate \"%s\" with non-DSD function %s\n", pCell1.pName, pRes); }
    }
    // generate permutations
    for  i = 2; i <= nVars; i++  {
        pPerm[i] = Extra_PermSchedule( i );
    }
    for  i = 2; i <= nVars; i++  {
        nPerms[i] = Extra_Factorial( i );
    }
    // add single cells
    for  pCell1 = p.pCells + 4; pCell1 < pLimit; pCell1++  {
        var nFanins int32  = pCell1.nFanins;
        assert(nFanins >= 2 && nFanins <= nVars, "nFanins >= 2 && nFanins <= nVars");
        for  i = 0; i < nFanins; i++  {
            Perm[i] = i;
        }
        // permute truth table
        tCur[0] = tTemp1[0] = pCell1.uTruth;
        if p.nVars > 6 {
            tTemp1[1] = tTemp1[2] = tTemp1[3] = tCur[1] = tCur[2] = tCur[3] = tCur[0]; }
        for  n = 0; n < nPerms[nFanins]; n++  {
            Sfm_LibPrepareAdd( p, tCur, Perm, nFanins, pCell1, nil, -1 );
            // update
            Abc_TtSwapAdjacent( tCur, p.nWords, pPerm[nFanins][n] );
            Perm1 = Perm + pPerm[nFanins][n];
            Perm2 = Perm1 + 1;
            // inlining of macro ABC_SWAP
            { var t int32  = *Perm1; *Perm1 = *Perm2; *Perm2 = t; };
        }
        assert(Abc_TtEqual(tTemp1, tCur, p.nWords), "Abc_TtEqual(tTemp1, tCur, p.nWords)");
    }
    // add double cells
    if fTwo != 0 {
    for  pCell1 = p.pCells + 4; pCell1 < pLimit; pCell1++  { // Bot
    if pCell1.Type > 0 {
    for  pCell2 = p.pCells + 4; pCell2 < pLimit; pCell2++  { // Top
    if pCell2.Type > 0 {//&& pCell1->Type + pCell2->Type <= 2 )
    if int32(pCell1.nFanins) + int32(pCell2.nFanins) <= nVars + 1 {
    for  f = 0; f < int32(pCell2.nFanins); f++  {
        var nFanins int32  = pCell1.nFanins + pCell2.nFanins - 1;
        assert(nFanins >= 2 && nFanins <= nVars, "nFanins >= 2 && nFanins <= nVars");
        for  i = 0; i < nFanins; i++  {
            Perm[i] = i;
        }
        // permute truth table
        if p.nVars > 6 {
            Sfm_LibTruth8Two( pCell1, pCell2, f, tCur );
            Abc_TtCopy( tTemp1, tCur, p.nWords, 0 );
        
        } else {
            tCur[0] = tTemp1[0] = Sfm_LibTruth6Two( pCell1, pCell2, f ); }
        for  n = 0; n < nPerms[nFanins]; n++  {
            Sfm_LibPrepareAdd( p, tCur, Perm, nFanins, pCell1, pCell2, f );
            if nFanins > 5 {
                break; }
            // update
            Abc_TtSwapAdjacent( tCur, p.nWords, pPerm[nFanins][n] );
            Perm1 = Perm + pPerm[nFanins][n];
            Perm2 = Perm1 + 1;
            // inlining of macro ABC_SWAP
            { var t int32  = *Perm1; *Perm1 = *Perm2; *Perm2 = t; };
        }
        assert(Abc_TtEqual(tTemp1, tCur, p.nWords), "Abc_TtEqual(tTemp1, tCur, p.nWords)");
    } } }
    } }
    } }
    // cleanup
    for  i = 2; i <= nVars; i++  {
        pPerm[i] = nil;
    }
    if fVerbose != 0 {
        fmt.Printf("Library processing: Var = %d. Cell = %d.  Fun = %d. Obj = %d. Ave = %.2f.  Skip = %d. Rem = %d.  ", nVars, p.nCells, Vec_MemEntryNum(p.vTtMem)-2, p.nObjs-p.nObjRemoved, 1.0*(p.nObjs-p.nObjRemoved)/(Vec_MemEntryNum(p.vTtMem)-2), p.nObjSkipped, p.nObjRemoved);
        Abc_PrintTime( 1, "Time", Abc_Clock() - clk );
    }
    return p;
}
func Sfm_LibPrintGate(pCell *Mio_Cell2_t, pFanins string, pCell2 *Mio_Cell2_t, pFanins2 string) {
    var k int32;
    fmt.Printf(" %-20s(", pCell.pName);
    for  k = 0; k < int32(pCell.nFanins); k++  {
        if pFanins[k] == byte(16) {
            Sfm_LibPrintGate( pCell2, pFanins2, nil, nil );
        } else {
            fmt.Printf(" %c", 'a' + pFanins[k]); }
    }
    fmt.Printf(" )");
}
func Sfm_LibPrintObj(p *Sfm_Lib_t, pObj *Sfm_Fun_t) {
    var pCellB *Mio_Cell2_t  = p.pCells + int32(pObj.pFansB[0]);
    var pCellT *Mio_Cell2_t  = p.pCells + int32(pObj.pFansT[0]);
    var i int32;
    var nFanins int32 = pCellB.nFanins + (pCellT == p.pCells ? 0 : pCellT.nFanins - 1);
    fmt.Printf("F = %d  A =%6.2f  ", nFanins, Scl_Int2Flt(pObj.Area));
    if pCellT == p.pCells {
        Sfm_LibPrintGate( pCellB, pObj.pFansB + 1, nil, nil );
    } else {
        Sfm_LibPrintGate( pCellT, pObj.pFansT + 1, pCellB, pObj.pFansB + 1 ); }
    // get hold of delay info
    if p.fDelay {
        var Offset int32   = Vec_IntEntry( &p.vProfs, Sfm_LibFunId(p, pObj) );
        var pProf *int32  = Vec_IntEntryP( &p.vStore, Offset );
        for  i = 0; i < nFanins; i++  {
            fmt.Printf("%6.2f ", Scl_Int2Flt(pProf[i]));
        }
    }
}
func Sfm_LibPrint(p *Sfm_Lib_t) {
    var pObj *Sfm_Fun_t; var pTruth *word; var i, nFanins int32;   
    // inlining of macro Vec_MemForEachEntry
    for  i = 0; (i < Vec_MemEntryNum(p.vTtMem)) && ((pTruth) = Vec_MemReadEntry(p.vTtMem, i)); i++  {
        if i < 2 || Vec_IntEntry(&p.vHits, i) == 0 {
            continue; }
        nFanins = Abc_TtSupportSize(pTruth, p.nVars);
        fmt.Printf("%8d : ", i);
        fmt.Printf("Num =%5d  ", Vec_IntEntry(&p.vCounts, i));
        fmt.Printf("Hit =%4d  ", Vec_IntEntry(&p.vHits, i));
        // inlining of macro Sfm_LibForEachSuper
        for  pObj = Sfm_LibFun(p, Vec_IntEntry(&p.vLists, i)); (pObj != nil); pObj = Sfm_LibFun(p, pObj.Next)  {
            Sfm_LibPrintObj( p, pObj );
            break;
        }
        fmt.Printf("    ");
        Dau_DsdPrintFromTruth( pTruth, nFanins );
    }
}
func Sfm_LibTest() {
    var p *Sfm_Lib_t;
    var fVerbose int32  = 1;
    if Abc_FrameReadLibGen() == nil {
        fmt.Printf("There is no current library.\n");
        return;
    }
    p = Sfm_LibPrepare( 7, 1, 1, 1, fVerbose );
    if fVerbose != 0 {
        Sfm_LibPrint( p ); }
    Sfm_LibStop( p );
}

/**Function*************************************************************

  Synopsis    []

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Sfm_LibFindAreaMatch(p *Sfm_Lib_t, pTruth *word, nFanins int32, piObj *int32) int32 {
    var pObj *Sfm_Fun_t  = nil;
    var iFunc int32  = *Vec_MemHashLookup( p.vTtMem,  pTruth );
    if iFunc == -1 {
        return -1; }
    // inlining of macro Sfm_LibForEachSuper
    for  pObj = Sfm_LibFun(p, Vec_IntEntry(&p.vLists, iFunc)); (pObj != nil); pObj = Sfm_LibFun(p, pObj.Next)  {
        break;
    }
    if piObj != nil {
        *piObj = pObj - p.pObjs; }
    return pObj.Area;
}
func Sfm_LibFindDelayMatches(p *Sfm_Lib_t, pTruth *word, pFanins *int32, nFanins int32, vGates *Vec_Ptr_t, vFans *Vec_Ptr_t) int32 {
    var pObj *Sfm_Fun_t;
    var pCellB, pCellT *Mio_Cell2_t;
    var iFunc int32;
    if nFanins > 6 {
        var pCopy [4]word;
        Abc_TtCopy( pCopy, pTruth, 4, 0 );
        Dau_DsdPrintFromTruth( pCopy, p.nVars );
    }
    Vec_PtrClear( vGates );
    Vec_PtrClear( vFans );
    // look for gate
    assert(!Abc_TtIsConst0(pTruth, p.nWords) && 
                !Abc_TtIsConst1(pTruth, p.nWords) && 
                !Abc_TtEqual(pTruth, s_Truth8[0], p.nWords) && 
                !Abc_TtOpposite(pTruth, s_Truth8[0], p.nWords), "!Abc_TtIsConst0(pTruth, p.nWords) && \n            !Abc_TtIsConst1(pTruth, p.nWords) && \n            !Abc_TtEqual(pTruth, s_Truth8[0], p.nWords) && \n            !Abc_TtOpposite(pTruth, s_Truth8[0], p.nWords)");
    iFunc = *Vec_MemHashLookup( p.vTtMem,  pTruth );
    if iFunc == -1 {
        // print functions not found in the library
        if p.fVerbose || nFanins > 6 {
            fmt.Printf("Not found in the precomputed library: ");
            Dau_DsdPrintFromTruth( pTruth, nFanins );
        }
        return 0;
    }
    Vec_IntAddToEntry( &p.vHits, iFunc, 1 );
    // collect matches
    // inlining of macro Sfm_LibForEachSuper
    for  pObj = Sfm_LibFun(p, Vec_IntEntry(&p.vLists, iFunc)); (pObj != nil); pObj = Sfm_LibFun(p, pObj.Next)  {
        pCellB = p.pCells + int32(pObj.pFansB[0]);
        pCellT = p.pCells + int32(pObj.pFansT[0]);
        Vec_PtrPush( vGates, pCellB.pMioGate );
        Vec_PtrPush( vGates, pCellT == p.pCells ? nil : pCellT.pMioGate );
        Vec_PtrPush( vFans, pObj.pFansB + 1 );
        Vec_PtrPush( vFans, pCellT == p.pCells ? nil : pObj.pFansT + 1 );
    }
    return Vec_PtrSize(vGates) / 2;
}

/**Function*************************************************************

  Synopsis    []

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Sfm_LibImplementSimple(p *Sfm_Lib_t, pTruth *word, pFanins *int32, nFanins int32, vGates *Vec_Int_t, vFanins *Vec_Wec_t) int32 {
    var pLib *Mio_Library_t  = (Mio_Library_t *)Abc_FrameReadLibGen();
    var pGate *Mio_Gate_t;
    var vLevel *Vec_Int_t;
    if Abc_TtIsConst0(pTruth, p.nWords) || Abc_TtIsConst1(pTruth, p.nWords) {
        assert(nFanins == 0, "nFanins == 0");
        if Abc_TtIsConst1(pTruth, p.nWords) {
            pGate = Mio_LibraryReadConst1(pLib);
        } else {
            pGate = Mio_LibraryReadConst0(pLib);
        }
        Vec_IntPush( vGates, Mio_GateReadValue(pGate) );
        vLevel = Vec_WecPushLevel( vFanins );
        return 1;
    }
    if Abc_TtEqual(pTruth, s_Truth8[0], p.nWords) || Abc_TtOpposite(pTruth, s_Truth8[0], p.nWords) {
        assert(nFanins == 1, "nFanins == 1");
        if Abc_TtEqual(pTruth, s_Truth8[0], p.nWords) {
            pGate = Mio_LibraryReadBuf(pLib);
        } else {
            pGate = Mio_LibraryReadInv(pLib);
        }
        Vec_IntPush( vGates, Mio_GateReadValue(pGate) );
        vLevel = Vec_WecPushLevel( vFanins );
        Vec_IntPush( vLevel, pFanins[0] );
        return 1;
    }
    assert(false, "0");
    return -1;
}
func Sfm_LibImplementGatesArea(p *Sfm_Lib_t, pFanins *int32, nFanins int32, iObj int32, vGates *Vec_Int_t, vFanins *Vec_Wec_t) int32 {
    var pLib *Mio_Library_t  = (Mio_Library_t *)Abc_FrameReadLibGen();
    var pObjMin *Sfm_Fun_t  = p.pObjs + iObj;
    var pCellB, pCellT *Mio_Cell2_t;
    var pGate *Mio_Gate_t;
    var vLevel *Vec_Int_t;
    var i int32;
    // get the gates
    pCellB = p.pCells + int32(pObjMin.pFansB[0]);
    pCellT = p.pCells + int32(pObjMin.pFansT[0]);
    // create bottom gate
    pGate = Mio_LibraryReadGateByName( pLib, pCellB.pName, nil );
    assert(pGate == pCellB.pMioGate, "pGate == pCellB.pMioGate");
    Vec_IntPush( vGates, Mio_GateReadValue(pGate) );
    vLevel = Vec_WecPushLevel( vFanins );
    for  i = 0; i < int32(pCellB.nFanins); i++  {
        Vec_IntPush( vLevel, pFanins[int32(pObjMin.pFansB[i+1])] );
    }
    if pCellT == p.pCells {
        return 1; }
    // create top gate
    pGate = Mio_LibraryReadGateByName( pLib, pCellT.pName, nil );
    assert(pGate == pCellT.pMioGate, "pGate == pCellT.pMioGate");
    Vec_IntPush( vGates, Mio_GateReadValue(pGate) );
    vLevel = Vec_WecPushLevel( vFanins );
    for  i = 0; i < int32(pCellT.nFanins); i++  {
        if pObjMin.pFansT[i+1] == byte(16) {
            Vec_IntPush( vLevel, Vec_WecSize(vFanins)-2 );
        } else {
            Vec_IntPush( vLevel, pFanins[int32(pObjMin.pFansT[i+1])] ); }
    }
    return 2;
}
func Sfm_LibImplementGatesDelay(p *Sfm_Lib_t, pFanins *int32, pGateB *Mio_Gate_t, pGateT *Mio_Gate_t, pFansB string, pFansT string, vGates *Vec_Int_t, vFanins *Vec_Wec_t) int32 {
    var vLevel *Vec_Int_t;
    var i, nFanins int32;
    // create bottom gate
    Vec_IntPush( vGates, Mio_GateReadValue(pGateB) );
    vLevel  = Vec_WecPushLevel( vFanins );
    nFanins = Mio_GateReadPinNum( pGateB );
    for  i = 0; i < nFanins; i++  {
        Vec_IntPush( vLevel, pFanins[int32(pFansB[i])] );
    }
    if pGateT == nil {
        return 1; }
    // create top gate
    Vec_IntPush( vGates, Mio_GateReadValue(pGateT) );
    vLevel  = Vec_WecPushLevel( vFanins );
    nFanins = Mio_GateReadPinNum( pGateT );
    for  i = 0; i < nFanins; i++  {
        if pFansT[i] == byte(16) {
            Vec_IntPush( vLevel, Vec_WecSize(vFanins)-2 );
        } else {
            Vec_IntPush( vLevel, pFanins[int32(pFansT[i])] ); }
    }
    return 2;
}

////////////////////////////////////////////////////////////////////////
///                       END OF FILE                                ///
////////////////////////////////////////////////////////////////////////


//ABC_NAMESPACE_IMPL_END

