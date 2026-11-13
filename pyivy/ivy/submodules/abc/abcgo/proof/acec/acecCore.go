package acec

import (
	"fmt"
	"os"
)

/**CFile****************************************************************

  FileName    [acecCore.c]

  SystemName  [ABC: Logic synthesis and verification system.]

  PackageName [CEC for arithmetic circuits.]

  Synopsis    [Core procedures.]

  Author      [Alan Mishchenko]
  
  Affiliation [UC Berkeley]

  Date        [Ver. 1.0. Started - June 20, 2005.]

  Revision    [$Id: acecCore.c,v 1.00 2005/06/20 00:00:00 alanmi Exp $]

***********************************************************************/

//#include "acecInt.h"
//#include "proof/cec/cec.h"
//#include "misc/util/utilTruth.h"
//#include "misc/extra/extra.h"

//ABC_NAMESPACE_IMPL_START


////////////////////////////////////////////////////////////////////////
///                        DECLARATIONS                              ///
////////////////////////////////////////////////////////////////////////

const TRUTH_UNUSED = 0x1234567812345678

////////////////////////////////////////////////////////////////////////
///                     FUNCTION DEFINITIONS                         ///
////////////////////////////////////////////////////////////////////////

/**Function*************************************************************

  Synopsis    [This procedure sets default parameters.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Acec_VerifyClasses(p *Gia_Man_t, vLits *Vec_Wec_t, vReprs *Vec_Wec_t) {
    var vFunc *Vec_Ptr_t  = Vec_PtrAlloc( Vec_WecSize(vLits) );
    var vSupp *Vec_Int_t  = Vec_IntAlloc( 100 );
    var vTemp *Vec_Wrd_t  = Vec_WrdStart( Gia_ManObjNum(p) );
    var vLevel *Vec_Int_t;
    var i, j, k, Entry, Entry2, nOvers, nErrors int32;
    // inlining of macro Vec_WecForEachLevel
    for  i = 0; (i < Vec_WecSize(vLits)) && (((vLevel) = Vec_WecEntry(vLits, i)), 1); i++  {
        var vTruths *Vec_Wrd_t  = Vec_WrdAlloc( Vec_IntSize(vLevel) );
        // inlining of macro Vec_IntForEachEntry
        for  k = 0; (k < Vec_IntSize(vLevel)) && (((Entry) = Vec_IntEntry(vLevel, k)), 1); k++  {
            var Truth word  = Gia_ObjComputeTruth6Cis( p, Entry, vSupp, vTemp );
            if Vec_IntSize(vSupp) > 6 {
                nOvers++;
                Vec_WrdPush( vTruths, TRUTH_UNUSED );
                continue;
            }
            vSupp.nSize = Abc_Tt6MinBase( &Truth, vSupp.pArray, vSupp.nSize );
            if Vec_IntSize(vSupp) > 5 {
                nOvers++;
                Vec_WrdPush( vTruths, TRUTH_UNUSED );
                continue;
            }
            Vec_WrdPush( vTruths, Truth );
        }
        Vec_PtrPush( vFunc, vTruths );
    }
    if nOvers {
        fmt.Printf("Detected %d oversize support nodes.\n", nOvers); }
    Vec_IntFree( vSupp );
    Vec_WrdFree( vTemp );
    // verify the classes
    // inlining of macro Vec_WecForEachLevel
    for  i = 0; (i < Vec_WecSize(vReprs)) && (((vLevel) = Vec_WecEntry(vReprs, i)), 1); i++  {
        var vTruths *Vec_Wrd_t  = (Vec_Wrd_t *)Vec_PtrEntry( vFunc, i );
        // inlining of macro Vec_IntForEachEntry
        for  k = 0; (k < Vec_IntSize(vLevel)) && (((Entry) = Vec_IntEntry(vLevel, k)), 1); k++  {
        // inlining of macro Vec_IntForEachEntryStart
        for  j = k+1; (j < Vec_IntSize(vLevel)) && (((Entry2) = Vec_IntEntry(vLevel, j)), 1); j++  {
            var Truth word  = Vec_WrdEntry( vTruths, k );
            var Truth2 word  = Vec_WrdEntry( vTruths, j );
            if Entry == Entry2 {
                nErrors++;
                if Truth != Truth2 && Truth != TRUTH_UNUSED && Truth2 != TRUTH_UNUSED {
                    fmt.Printf("Rank %d:  Lit %d and %d do not pass verification.\n", i, k, j); }
            }
            if Entry == Abc_LitNot(Entry2) {
                nErrors++;
                if Truth != ~Truth2 && Truth != TRUTH_UNUSED && Truth2 != TRUTH_UNUSED {
                    fmt.Printf("Rank %d:  Lit %d and %d do not pass verification.\n", i, k, j); }
            }
        }
        }
    }
    if nErrors != 0 {
        fmt.Printf("Total errors in equivalence classes = %d.\n", nErrors); }
    Vec_VecFree( (Vec_Vec_t *)vFunc );
}

/**Function*************************************************************

  Synopsis    []

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Acec_ComputeEquivClasses(pOne *Gia_Man_t, pTwo *Gia_Man_t, pvMap1 **Vec_Int_t, pvMap2 **Vec_Int_t) {
    var clk abctime  = Abc_Clock();
    var pBase, pRepr *Gia_Man_t;
    pBase = Acec_CommonStart( nil, pOne );
    pBase = Acec_CommonStart( pBase, pTwo );
    Acec_CommonFinish( pBase );
    //Gia_ManShow( pBase, NULL, 0, 0, 0 );
    pRepr = Gia_ManComputeGiaEquivs( pBase, 100, 0 );
    *pvMap1 = Acec_CountRemap( pOne, pBase );
    *pvMap2 = Acec_CountRemap( pTwo, pBase );
    Gia_ManStop( pBase );
    Gia_ManStop( pRepr );
    fmt.Printf("Finished computing equivalent nodes.  ");
    Abc_PrintTime( 1, "Time", Abc_Clock() - clk );
}
func Acec_MatchPrintEquivLits(p *Gia_Man_t, vLits *Vec_Wec_t, pCostLits *int32, fVerbose int32) {
    var vSupp *Vec_Int_t;
    var vTemp *Vec_Wrd_t;
    var vLevel *Vec_Int_t;
    var i, k, Entry int32;
    fmt.Printf("Leaf literals and their classes:\n");
    // inlining of macro Vec_WecForEachLevel
    for  i = 0; (i < Vec_WecSize(vLits)) && (((vLevel) = Vec_WecEntry(vLits, i)), 1); i++  {
        if Vec_IntSize(vLevel) == 0 {
            continue; }
        fmt.Printf("Rank %2d : %2d  ", i, Vec_IntSize(vLevel));
        // inlining of macro Vec_IntForEachEntry
        for  k = 0; (k < Vec_IntSize(vLevel)) && (((Entry) = Vec_IntEntry(vLevel, k)), 1); k++  {
            fmt.Printf("%s%d(%d) ", Abc_LitIsCompl(Entry) ? "-":"+", Abc_Lit2Var(Entry), Abc_Lit2LitL(pCostLits, Entry));
        }
        fmt.Printf("\n");
    }
    if !(fVerbose != 0) {
        return; }
    vSupp = Vec_IntAlloc( 100 );
    vTemp = Vec_WrdStart( Gia_ManObjNum(p) );
    // inlining of macro Vec_WecForEachLevel
    for  i = 0; (i < Vec_WecSize(vLits)) && (((vLevel) = Vec_WecEntry(vLits, i)), 1); i++  {
        //if ( i != 20 )
        //    continue;
        if Vec_IntSize(vLevel) == 0 {
            continue; }
        // inlining of macro Vec_IntForEachEntry
        for  k = 0; (k < Vec_IntSize(vLevel)) && (((Entry) = Vec_IntEntry(vLevel, k)), 1); k++  {
            var Truth word  = Gia_ObjComputeTruth6Cis( p, Entry, vSupp, vTemp );
/*
            {
                int iObj = Abc_Lit2Var(Entry);
                Gia_Man_t * pGia0 = Gia_ManDupAndCones( p, &iObj, 1, 1 );
                Gia_ManShow( pGia0, NULL, 0, 0, 0 );
                Gia_ManStop( pGia0 );
            }
*/
            fmt.Printf("Rank = %4d : ", i);
            fmt.Printf("Obj = %4d  ", Abc_Lit2Var(Entry));
            if Vec_IntSize(vSupp) > 6 {
                fmt.Printf("Supp = %d.\n", Vec_IntSize(vSupp));
                continue;
            }
            vSupp.nSize = Abc_Tt6MinBase( &Truth, vSupp.pArray, vSupp.nSize );
            if Vec_IntSize(vSupp) > 5 {
                fmt.Printf("Supp = %d.\n", Vec_IntSize(vSupp));
                continue;
            }
            Extra_PrintHex( os.Stdout, (unsigned*)&Truth, Vec_IntSize(vSupp) );
            if Vec_IntSize(vSupp) == 4 { fmt.Printf("    "); }
            if Vec_IntSize(vSupp) == 3 { fmt.Printf("      "); }
            if Vec_IntSize(vSupp) <= 2 { fmt.Printf("       "); }
            fmt.Printf("  ");
            Vec_IntPrint( vSupp );
        }
        fmt.Printf("\n");
    }
    Vec_IntFree( vSupp );
    Vec_WrdFree( vTemp );
}
func Acec_MatchBoxes(pBox0 *Acec_Box_t, pBox1 *Acec_Box_t) int32 {
    var vMap0, vMap1, vLevel *Vec_Int_t; 
    var i, nSize, nTotal int32;
    Acec_ComputeEquivClasses( pBox0.pGia, pBox1.pGia, &vMap0, &vMap1 );
    // sort nodes in the classes by their equivalences
    // inlining of macro Vec_WecForEachLevel
    for  i = 0; (i < Vec_WecSize(pBox0.vLeafLits)) && (((vLevel) = Vec_WecEntry(pBox0.vLeafLits, i)), 1); i++  {
        Acec_MatchBoxesSort( Vec_IntArray(vLevel), Vec_IntSize(vLevel), Vec_IntArray(vMap0) );
    }
    // inlining of macro Vec_WecForEachLevel
    for  i = 0; (i < Vec_WecSize(pBox1.vLeafLits)) && (((vLevel) = Vec_WecEntry(pBox1.vLeafLits, i)), 1); i++  {
        Acec_MatchBoxesSort( Vec_IntArray(vLevel), Vec_IntSize(vLevel), Vec_IntArray(vMap1) );
    }
    Acec_MatchCheckShift( pBox0.pGia, pBox1.pGia, pBox0.vLeafLits, pBox1.vLeafLits, vMap0, vMap1, pBox0.vRootLits, pBox1.vRootLits );
    
    //Acec_MatchPrintEquivLits( pBox0->pGia, pBox0->vLeafLits, Vec_IntArray(vMap0), 0 );
    //Acec_MatchPrintEquivLits( pBox1->pGia, pBox1->vLeafLits, Vec_IntArray(vMap1), 0 );
    //printf( "Outputs:\n" );
    //Vec_WecPrintLits( pBox0->vRootLits );
    //printf( "Outputs:\n" );
    //Vec_WecPrintLits( pBox1->vRootLits );

    // reorder nodes to have the same order
    assert(pBox0.vShared == nil, "pBox0.vShared == nil");
    assert(pBox1.vShared == nil, "pBox1.vShared == nil");
    pBox0.vShared = Vec_WecStart( Vec_WecSize(pBox0.vLeafLits) );
    pBox1.vShared = Vec_WecStart( Vec_WecSize(pBox1.vLeafLits) );
    pBox0.vUnique = Vec_WecStart( Vec_WecSize(pBox0.vLeafLits) );
    pBox1.vUnique = Vec_WecStart( Vec_WecSize(pBox1.vLeafLits) );
    nSize = Abc_MinInt( Vec_WecSize(pBox0.vLeafLits), Vec_WecSize(pBox1.vLeafLits) );
    // inlining of macro Vec_WecForEachLevelStart
    for  i = nSize; (i < Vec_WecSize(pBox0.vLeafLits)) && (((vLevel) = Vec_WecEntry(pBox0.vLeafLits, i)), 1); i++  {
        Vec_IntAppend( Vec_WecEntry(pBox0.vUnique, i), vLevel );
    }
    // inlining of macro Vec_WecForEachLevelStart
    for  i = nSize; (i < Vec_WecSize(pBox1.vLeafLits)) && (((vLevel) = Vec_WecEntry(pBox1.vLeafLits, i)), 1); i++  {
        Vec_IntAppend( Vec_WecEntry(pBox1.vUnique, i), vLevel );
    }
    for  i = 0; i < nSize; i++  {
        var vShared0 *Vec_Int_t  = Vec_WecEntry( pBox0.vShared, i );
        var vShared1 *Vec_Int_t  = Vec_WecEntry( pBox1.vShared, i );
        var vUnique0 *Vec_Int_t  = Vec_WecEntry( pBox0.vUnique, i );
        var vUnique1 *Vec_Int_t  = Vec_WecEntry( pBox1.vUnique, i );

        var vLevel0 *Vec_Int_t  = Vec_WecEntry( pBox0.vLeafLits, i );
        var vLevel1 *Vec_Int_t  = Vec_WecEntry( pBox1.vLeafLits, i );
        var pBeg0 *int32  = Vec_IntArray(vLevel0);
        var pBeg1 *int32  = Vec_IntArray(vLevel1);
        var pEnd0 *int32  = Vec_IntLimit(vLevel0);
        var pEnd1 *int32  = Vec_IntLimit(vLevel1);
        for  pBeg0 < pEnd0 && pBeg1 < pEnd1  {
            var Entry0 int32  = Abc_Lit2LitL( Vec_IntArray(vMap0), *pBeg0 );
            var Entry1 int32  = Abc_Lit2LitL( Vec_IntArray(vMap1), *pBeg1 );
            assert(*pBeg0 && *pBeg1, "*pBeg0 && *pBeg1");
            if Entry0 == Entry1 {
                Vec_IntPush( vShared0, *pBeg0++ );
                Vec_IntPush( vShared1, *pBeg1++ );
            
            } else if Entry0 > Entry1 {
                Vec_IntPush( vUnique0, *pBeg0++ );
            } else { 
                Vec_IntPush( vUnique1, *pBeg1++ ); }
        }
        for  pBeg0 < pEnd0  {
            Vec_IntPush( vUnique0, *pBeg0++ );
        }
        for  pBeg1 < pEnd1  {
            Vec_IntPush( vUnique1, *pBeg1++ );
        }
        assert(Vec_IntSize(vShared0) == Vec_IntSize(vShared1), "Vec_IntSize(vShared0) == Vec_IntSize(vShared1)");
        assert(Vec_IntSize(vShared0) + Vec_IntSize(vUnique0) == Vec_IntSize(vLevel0), "Vec_IntSize(vShared0) + Vec_IntSize(vUnique0) == Vec_IntSize(vLevel0)");
        assert(Vec_IntSize(vShared1) + Vec_IntSize(vUnique1) == Vec_IntSize(vLevel1), "Vec_IntSize(vShared1) + Vec_IntSize(vUnique1) == Vec_IntSize(vLevel1)");
    }
    nTotal = Vec_WecSizeSize(pBox0.vShared);
    fmt.Printf("Box0: Matched %d entries out of %d.\n", nTotal, Vec_WecSizeSize(pBox0.vLeafLits));
    fmt.Printf("Box1: Matched %d entries out of %d.\n", nTotal, Vec_WecSizeSize(pBox1.vLeafLits));

    //Acec_MatchPrintEquivLits( pBox0->pGia, pBox0->vShared, Vec_IntArray(vMap0), 0 );
    //Acec_MatchPrintEquivLits( pBox1->pGia, pBox1->vShared, Vec_IntArray(vMap1), 0 );
    //printf( "\n" );

    //Acec_MatchPrintEquivLits( pBox0->pGia, pBox0->vUnique, Vec_IntArray(vMap0), 0 );
    //Acec_MatchPrintEquivLits( pBox1->pGia, pBox1->vUnique, Vec_IntArray(vMap1), 0 );

    Vec_IntFree( vMap0 );
    Vec_IntFree( vMap1 );
    return nTotal;
}

/**Function*************************************************************

  Synopsis    []

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Acec_Solve(pGia0 *Gia_Man_t, pGia1 *Gia_Man_t, pPars *Acec_ParCec_t) int32 {
    var status int32  = -1;
    var clk abctime  = Abc_Clock();
    var pMiter *Gia_Man_t;
    var pGia0n *Gia_Man_t = pGia0;
    var pGia1n *Gia_Man_t = pGia1;
    var ParsCec Cec_ParCec_t;
    var pCecPars *Cec_ParCec_t = &ParsCec;
//    Vec_Bit_t * vIgnore0 = pPars->fBooth ? Acec_BoothFindPPG(pGia0) : NULL;
//    Vec_Bit_t * vIgnore1 = pPars->fBooth ? Acec_BoothFindPPG(pGia1) : NULL;
//    Acec_Box_t * pBox0 = Acec_DeriveBox( pGia0, vIgnore0, 0, 0, pPars->fVerbose );
//    Acec_Box_t * pBox1 = Acec_DeriveBox( pGia1, vIgnore1, 0, 0, pPars->fVerbose );
//    Vec_BitFreeP( &vIgnore0 );
//    Vec_BitFreeP( &vIgnore1 );
    var pBox0 *Acec_Box_t  = Acec_ProduceBox( pGia0, pPars.fVerbose );
    var pBox1 *Acec_Box_t  = Acec_ProduceBox( pGia1, pPars.fVerbose );
    if pBox0 == nil || pBox1 == nil { // cannot match
        fmt.Printf("Cannot find arithmetic boxes in both LHS and RHS. Trying regular CEC.\n");
    } else if !(Acec_MatchBoxes( pBox0, pBox1 ) != 0) { // cannot find matching
        fmt.Printf("Cannot match arithmetic boxes in LHS and RHS. Trying regular CEC.\n");
    } else {
        pGia0n = Acec_InsertBox( pBox0, 0 );
        pGia1n = Acec_InsertBox( pBox1, 0 );
        fmt.Printf("Matching of adder trees in LHS and RHS succeeded.  ");
        Abc_PrintTime( 1, "Time", Abc_Clock() - clk );
        // remove the last output
        Gia_ManPatchCoDriver( pGia0n, Gia_ManCoNum(pGia0n)-1, 0 );
        Gia_ManPatchCoDriver( pGia1n, Gia_ManCoNum(pGia1n)-1, 0 );

        Gia_ManPatchCoDriver( pGia0n, Gia_ManCoNum(pGia0n)-2, 0 );
        Gia_ManPatchCoDriver( pGia1n, Gia_ManCoNum(pGia1n)-2, 0 );
    }
    // solve regular CEC problem 
    Cec_ManCecSetDefaultParams( pCecPars );
    pCecPars.nBTLimit = pPars.nBTLimit;
    pMiter = Gia_ManMiter( pGia0n, pGia1n, 0, 1, 0, 0, pPars.fVerbose );
    if pMiter != nil {
        var fDumpMiter int32;
        if fDumpMiter != 0 {
            Abc_Print( 0, "The verification miter is written into file \"%s\".\n", "acec_miter.aig" );
            Gia_AigerWrite( pMiter, "acec_miter.aig", 0, 0, 0 );
        }
        status = Cec_ManVerify( pMiter, pCecPars );
        // inlining of macro ABC_SWAP
        { var t *Abc_Cex_t  = pGia0.pCexComb; pGia0.pCexComb = pMiter.pCexComb; pMiter.pCexComb = t; };
        Gia_ManStop( pMiter );
    
    } else {
        fmt.Printf("Miter computation has failed.\n"); }
    if pGia0n != pGia0 {
        Gia_ManStop( pGia0n ); }
    if pGia1n != pGia1 {
        Gia_ManStop( pGia1n ); }
    Acec_BoxFreeP( &pBox0 );
    Acec_BoxFreeP( &pBox1 );
    return status;
}

////////////////////////////////////////////////////////////////////////
///                       END OF FILE                                ///
////////////////////////////////////////////////////////////////////////


//ABC_NAMESPACE_IMPL_END

