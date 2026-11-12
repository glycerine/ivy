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



////////////////////////////////////////////////////////////////////////
///                       END OF FILE                                ///
////////////////////////////////////////////////////////////////////////


//ABC_NAMESPACE_IMPL_END

