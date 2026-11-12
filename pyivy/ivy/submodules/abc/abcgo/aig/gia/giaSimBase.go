package gia

import "fmt"

import "os"

/**CFile****************************************************************

  FileName    [giaSim.c]

  SystemName  [ABC: Logic synthesis and verification system.]

  PackageName [Scalable AIG package.]

  Synopsis    [Fast sequential simulator.]

  Author      [Alan Mishchenko]
  
  Affiliation [UC Berkeley]

  Date        [Ver. 1.0. Started - June 20, 2005.]

  Revision    [$Id: giaSim.c,v 1.00 2005/06/20 00:00:00 alanmi Exp $]

***********************************************************************/

//#include "gia.h"
//#include "misc/util/utilTruth.h"
//#include "misc/extra/extra.h"
//#include <immintrin.h>

//ABC_NAMESPACE_IMPL_START


////////////////////////////////////////////////////////////////////////
///                        DECLARATIONS                              ///
////////////////////////////////////////////////////////////////////////


   

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





/**Function*************************************************************

  Synopsis    []

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/

func Gia_ManPatSatImprove(p *Gia_Man_t, nWords0 int32, fVerbose int32) {
    extern Vec_Int_t * Cbs2_ManSolveMiterNc( Gia_Man_t * pAig, int nConfs, Vec_Str_t ** pvStatus, int fVerbose );
    var i int32;
    var Status int32;
    var Counts [3]int32 ;
    var pGia *Gia_Man_t;
    var vSimsIn *Vec_Wrd_t  = nil;
    var vStatus *Vec_Str_t  = nil;
    var vCexStore *Vec_Int_t  = nil;
    var vSims *Vec_Wrd_t  = Gia_ManSimPatSim( p );
    //Gia_ManSimProfile( p );
    pGia  = Gia_ManSimPatGenMiter( p, vSims );
    vCexStore = Cbs2_ManSolveMiterNc( pGia, 1000, &vStatus, 0 );
    Gia_ManStop( pGia );
    // inlining of macro Vec_StrForEachEntry
    for  i = 0; (i < Vec_StrSize(vStatus)) && (((Status) = Vec_StrEntry(vStatus, i)), 1); i++  {
        assert(Status >= -1 && Status <= 1, "Status >= -1 && Status <= 1");
        Counts[Status+1]++;
    }
    if fVerbose != 0 {
        fmt.Printf("Total = %d : SAT = %d.  UNSAT = %d.  UNDEC = %d.\n", Counts[1]+Counts[2]+Counts[0], Counts[1], Counts[2], Counts[0]); }
    if Counts[1] == 0 {
        fmt.Printf("There are no counter-examples.  No need for more simulation.\n");
    } else {
        vSimsIn = Gia_ManSimBitPacking( p, vCexStore, Counts[1], Counts[0] );
        Vec_WrdFreeP( &p.vSimsPi );
        p.vSimsPi = vSimsIn;
        //Gia_ManSimProfile( p );
    }
    Vec_StrFree( vStatus );
    Vec_IntFree( vCexStore );
    Vec_WrdFree( vSims );
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
func Gia_ManSimRelPrint(p *Gia_Man_t, vRel *Vec_Wrd_t, vOutMints *Vec_Int_t) {
    var nWords int32  = Vec_WrdSize(p.vSimsPi) / Gia_ManCiNum(p);
    var nMints int32  = Vec_WrdSize(vRel) / nWords; 
    var i, m, Count int32;
/*
    for ( i = 0; i < 64 * nWords; i++ )
    {
        int k;
        for ( k = 0; k < Gia_ManCiNum(p); k++ )
            printf( "%d", Abc_TtGetBit( Vec_WrdEntryP(p->vSimsPi, k), i ) );
        printf( " " );
        Count = 0;
        for ( m = 0; m < nMints; m++ )
        {
            printf( "%d", Abc_TtGetBit( Vec_WrdArray(vRel), i*nMints+m ) );
            Count += Abc_TtGetBit( Vec_WrdArray(vRel), i*nMints+m );
        }
        printf( "  Count = %2d ", Count );
        if ( vOutMints )
        {
            printf( "   %3d ", Vec_IntEntry(vOutMints, i) );
            if ( Abc_TtGetBit( Vec_WrdArray(vRel), i*nMints+Vec_IntEntry(vOutMints, i) ) )
                printf( "yes" );
            else
                printf( "no" );
        }
        printf( "\n" );
    }
*/
/*
    for ( i = 0; i < 64 * nWords; i++ )
    {
        Count = 0;
        for ( m = 0; m < nMints; m++ )
            Count += Abc_TtGetBit( Vec_WrdArray(vRel), i*nMints+m );
        printf( "%d ", Count );
    }
    printf( "\n" );
*/
    for  i = 0; i < 64 * nWords; i++  {
        Count = 0;
        for  m = 0; m < nMints; m++  {
            fmt.Printf("%d", Abc_TtGetBit( Vec_WrdArray(vRel), i*nMints+m ));
            Count += Abc_TtGetBit( Vec_WrdArray(vRel), i*nMints+m );
        }
        fmt.Printf("  Count = %2d \n", Count);
    }
}
func Gia_ManSimPatStart(nItems int32) *Vec_Int_t {
    var vValues *Vec_Int_t  = Vec_IntAlloc( nItems );
    Vec_IntPush( vValues, 17 );
    Vec_IntPush( vValues, 39 );
    Vec_IntPush( vValues, 56 );
    Vec_IntPush( vValues, 221 );
    return vValues;
}
func Gia_ManSimRelTest(p *Gia_Man_t) {
    //int nWords        = Vec_WrdSize(p->vSimsPi) / Gia_ManCiNum(p);
    var vObjs *Vec_Int_t  = Gia_ManSimPatStart( 4 ); // can be CI/AND/CO
    var vVals *Vec_Wrd_t  = Gia_ManSimPatValues( p );
    var vRel *Vec_Wrd_t   = Gia_ManSimRel( p, vObjs, vVals );
    assert(p.vSimsPi != nil, "p.vSimsPi != nil");
    Gia_ManSimRelPrint( p, vRel, nil );
    Vec_IntFree( vObjs );
    Vec_WrdFree( vVals );
    Vec_WrdFree( vRel );
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

  Synopsis    [Improving quality of simulation patterns.]

  Description []

  SideEffects []

  SeeAlso     []

***********************************************************************/
func Gia_SimQualityOne(p *Gia_Man_t, vPat *Vec_Int_t, fPoOnly int32) *Vec_Int_t {
    var i int32;
    var Id int32;
    var Value int32;
    var nWords int32 = Abc_Bit6WordNum( 1+Gia_ManCiNum(p) );
    var vTemp *Vec_Wrd_t;
    var vSims *Vec_Wrd_t;
    var vSimsPi *Vec_Wrd_t = Vec_WrdStart( Gia_ManCiNum(p) * nWords );
    var vRes *Vec_Int_t;
    assert(Vec_IntSize(vPat) == Gia_ManCiNum(p), "Vec_IntSize(vPat) == Gia_ManCiNum(p)");
    // inlining of macro Vec_IntForEachEntry
    for  i = 0; (i < Vec_IntSize(vPat)) && (((Value) = Vec_IntEntry(vPat, i)), 1); i++  {
        var pSim *word  = Vec_WrdEntryP( vSimsPi, i*nWords );
        if Value != 0 {
            Abc_TtFill( pSim, nWords ); }
        Abc_TtXorBit( pSim, i+1 );
    }
    vTemp = p.vSimsPi;
    p.vSimsPi = vSimsPi;
    vSims = Gia_ManSimPatSim( p );
    p.vSimsPi = vTemp;
    if fPoOnly != 0 {
        vRes = Vec_IntStart( Gia_ManCoNum(p) );
        // inlining of macro Gia_ManForEachCoId
        for  i = 0; (i < Vec_IntSize(p.vCos)) && ((Id) = Gia_ObjId(p, Gia_ManCo(p, i))); i++  {
            var pSim *word  = Vec_WrdEntryP( vSims, Id*nWords );
            if pSim[0] & 1 {
                Abc_TtNot( pSim, nWords ); }
            Vec_IntWriteEntry( vRes, i, Abc_TtCountOnesVec(pSim, nWords) );
        }
        assert(Vec_IntSize(vRes) == Gia_ManCoNum(p), "Vec_IntSize(vRes) == Gia_ManCoNum(p)");
    
    } else {
        vRes = Vec_IntStart( Gia_ManObjNum(p) );
        // inlining of macro Gia_ManForEachAndId
        for  Id = 0; (Id < p.nObjs); Id++  {                                     if !Gia_ObjIsAnd(Gia_ManObj(p, Id)) {} else {
            var pSim *word  = Vec_WrdEntryP( vSims, Id*nWords );
            if pSim[0] & 1 {
                Abc_TtNot( pSim, nWords ); }
            Vec_IntWriteEntry( vRes, Id, Abc_TtCountOnesVec(pSim, nWords) );
        }
        }
        assert(Vec_IntSize(vRes) == Gia_ManObjNum(p), "Vec_IntSize(vRes) == Gia_ManObjNum(p)");
    }
    Vec_WrdFree( vSims );
    Vec_WrdFree( vSimsPi );
    return vRes;
}
func Gia_SimQualityTest(p *Gia_Man_t) {
    var vPat, vRes *Vec_Int_t;
    var k int32;
    var m int32;
    var nMints int32 = (1 << Gia_ManCiNum(p));
    assert(Gia_ManCiNum(p) <= 10, "Gia_ManCiNum(p) <= 10");
    for  m = 0; m < nMints; m++  {
        fmt.Printf("%d : ", m);
        Extra_PrintBinary( os.Stdout, (unsigned*)&m, Gia_ManCiNum(p) );
        fmt.Printf(" ");
        vPat = Vec_IntAlloc( Gia_ManCiNum(p) );
        for  k = 0; k < Gia_ManCiNum(p); k++  {
            Vec_IntPush( vPat, (m >> k) & 1 );
        }
        vRes = Gia_SimQualityOne( p, vPat, 1 );
        fmt.Printf("%d ", Vec_IntSum(vRes));
        Vec_IntFree( vRes );
        Vec_IntFree( vPat );
        fmt.Printf("\n");
    }
}
func Gia_SimGenerateStats(p *Gia_Man_t) *Vec_Int_t {
    var vTotal *Vec_Int_t  = Vec_IntStart( Gia_ManObjNum(p) );
    var vRes, vPat *Vec_Int_t;
    var i, k, Value int32;
    Abc_Random(1);
    for  i = 0; i < 1000; i++  {
        vPat = Vec_IntAlloc( Gia_ManCiNum(p) );
        for  k = 0; k < Gia_ManCiNum(p); k++  {
            Vec_IntPush( vPat, Abc_Random(0) & 1 );
        }
        vRes = Gia_SimQualityOne( p, vPat, 0 );
        assert(Vec_IntSize(vRes) == Gia_ManObjNum(p), "Vec_IntSize(vRes) == Gia_ManObjNum(p)");
        // inlining of macro Vec_IntForEachEntry
        for  k = 0; (k < Vec_IntSize(vRes)) && (((Value) = Vec_IntEntry(vRes, k)), 1); k++  {
            Vec_IntAddToEntry( vTotal, k, Value );
        }
        Vec_IntFree( vRes );
        Vec_IntFree( vPat );
    }
    //Vec_IntPrint( vTotal );
    return vTotal;
}   
func Gia_SimComputeScore(p *Gia_Man_t, vTotal *Vec_Int_t, vThis *Vec_Int_t) float64 {
    var TotalScore float64;
    var i, Total, This int32; 
    assert(Vec_IntSize(vTotal) == Vec_IntSize(vThis), "Vec_IntSize(vTotal) == Vec_IntSize(vThis)");
    // inlining of macro Vec_IntForEachEntryTwo
    for  i = 0; (i < Vec_IntSize(vTotal)) && (((Total) = Vec_IntEntry(vTotal, i)), 1) && (((This) = Vec_IntEntry(vThis, i)), 1); i++  {
        if Total == 0 { 
            Total = 1; }
        TotalScore += 1000.0*This/Total;
    }
    if TotalScore == 0 {
        return 1.0;
    } else {
        return TotalScore/Gia_ManAndNum(p);
    }
}
func Gia_SimQualityPatternsMax(p *Gia_Man_t, vPat *Vec_Int_t, Iter int32, fVerbose int32, vStats *Vec_Int_t) int32 {
    var k int32;
    var MaxIn int32 = -1;
    var vTries *Vec_Int_t  = Vec_IntAlloc( 100 );
    var vRes *Vec_Int_t  = Gia_SimQualityOne( p, vPat, 0 );
    var Value float64;
    var InitValue float64;
    var MaxValue float64 = InitValue = Gia_SimComputeScore( p, vStats, vRes );
    Vec_IntFree( vRes );

    if fVerbose != 0 {
        fmt.Printf("Iter %5d : Init = %6.3f  ", Iter, InitValue); }

    for  k = 0; k < Gia_ManCiNum(p); k++  {
        Vec_IntArray(vPat)[k] ^= 1;
        //Vec_IntPrint( vPat );

        vRes = Gia_SimQualityOne( p, vPat, 0 );
        Value = Gia_SimComputeScore( p, vStats, vRes );
        if MaxValue <= Value {
            if MaxValue < Value {
                Vec_IntClear( vTries ); }
            Vec_IntPush( vTries, k );
            MaxValue = Value;
            MaxIn = k;
        }
        Vec_IntFree( vRes );

        Vec_IntArray(vPat)[k] ^= 1;
    }
    if Vec_IntSize(vTries) {
        MaxIn = Vec_IntEntry( vTries, rand()%Vec_IntSize(vTries) );
    } else {
        MaxIn = -1;
    }
    if fVerbose != 0 {
        fmt.Printf("Final = %6.3f  Ratio = %4.2f  Tries = %5d  ", MaxValue, MaxValue/InitValue, Vec_IntSize(vTries));
        fmt.Printf("Choosing %5d\r", MaxIn);
    }
    Vec_IntFree( vTries );
    return MaxIn;
}
func Gia_ManPatCollectOne(p *Gia_Man_t, vPatterns *Vec_Wrd_t, n int32, nWords int32) *Vec_Int_t {
    var vPat *Vec_Int_t  = Vec_IntAlloc( Gia_ManCiNum(p) ); var k int32;
    for  k = 0; k < Gia_ManCiNum(p); k++  {
        Vec_IntPush( vPat, Abc_TtGetBit( Vec_WrdEntryP(vPatterns, k*nWords), n ) );
    }
    return vPat;
}
func Gia_ManPatUpdateOne(p *Gia_Man_t, vPatterns *Vec_Wrd_t, n int32, nWords int32, vPat *Vec_Int_t) {
    var k, Value int32;
    // inlining of macro Vec_IntForEachEntry
    for  k = 0; (k < Vec_IntSize(vPat)) && (((Value) = Vec_IntEntry(vPat, k)), 1); k++  {
        var pSim *word  = Vec_WrdEntryP( vPatterns, k*nWords );
        if Abc_TtGetBit(pSim, n) != Value {
            Abc_TtXorBit( pSim, n ); }
    }
}
func Gia_ManPatDistImprove(p *Gia_Man_t, fVerbose int32) {
    var n int32;
    var k int32;
    var nWords int32 = Vec_WrdSize(p.vSimsPi) / Gia_ManCiNum(p);
    var InitValue, InitTotal, FinalValue, FinalTotal float64;
    var vPat *Vec_Int_t;
    var vRes *Vec_Int_t;
    var vStats *Vec_Int_t = Gia_SimGenerateStats( p );
    var vPatterns *Vec_Wrd_t  = p.vSimsPi; p.vSimsPi = nil;
    Abc_Random(1);
    for  n = 0; n < 64*nWords; n++  {
        var clk abctime  = Abc_Clock();
//        if ( n == 32 )
//            break;

        vPat = Gia_ManPatCollectOne( p, vPatterns, n, nWords );
        vRes = Gia_SimQualityOne( p, vPat, 0 );
        InitValue = Gia_SimComputeScore(p, vStats, vRes);
        InitTotal += InitValue;
        Vec_IntFree( vRes );

        for  k = 0; k < 100; k++  {
            var MaxIn int32  = Gia_SimQualityPatternsMax( p, vPat, k, fVerbose, vStats );
            if MaxIn == -1 {
                break; }
            assert(MaxIn >= 0 && MaxIn < Gia_ManCiNum(p), "MaxIn >= 0 && MaxIn < Gia_ManCiNum(p)");
            Vec_IntArray(vPat)[MaxIn] ^= 1;            
        }
        //Vec_IntPrint( vPat );

        vRes = Gia_SimQualityOne( p, vPat, 0 );
        FinalValue = Gia_SimComputeScore(p, vStats, vRes);
        FinalTotal += FinalValue;
        Vec_IntFree( vRes );

        if fVerbose != 0 {
            fmt.Printf("Pat %5d : Tries = %5d  InitValue = %6.3f  FinalValue = %6.3f  Ratio = %4.2f  ", n, k, InitValue, FinalValue, FinalValue/InitValue);
            Abc_PrintTime( 1, "Time", Abc_Clock() - clk );
        }

        Gia_ManPatUpdateOne( p, vPatterns, n, nWords, vPat );
        Vec_IntFree( vPat );
    }
    Vec_IntFree( vStats );
    if fVerbose != 0 {
        fmt.Printf("\n"); }
    fmt.Printf("Improved %d patterns with average init value %.2f and average final value %.2f.\n", 64*nWords, 1.0*InitTotal/(64*nWords), 1.0*FinalTotal/(64*nWords));
    p.vSimsPi = vPatterns;
}

/**Function*************************************************************

  Synopsis    [Improving quality of simulation patterns.]

  Description []

  SideEffects []

  SeeAlso     []

***********************************************************************/
func Gia_SimCollectRare(p *Gia_Man_t, vPatterns *Vec_Wrd_t, RareLimit int32) *Vec_Int_t {
    var vRareCounts *Vec_Int_t  = Vec_IntAlloc( 100 );  // (node, rare_count) pairs
    var Id int32;
    var nWords int32 = Vec_WrdSize(vPatterns) / Gia_ManCiNum(p);
    var TotalBits int32 = 64*nWords;
    var vSims *Vec_Wrd_t;
    var vTemp *Vec_Wrd_t = p.vSimsPi;
    assert(Vec_WrdSize(vPatterns) % Gia_ManCiNum(p) == 0, "Vec_WrdSize(vPatterns) % Gia_ManCiNum(p) == 0");
    p.vSimsPi = vPatterns;
    vSims = Gia_ManSimPatSim( p );
    p.vSimsPi = vTemp;
    // inlining of macro Gia_ManForEachAndId
    for  Id = 0; (Id < p.nObjs); Id++  {                                     if !Gia_ObjIsAnd(Gia_ManObj(p, Id)) {} else {
        var pSim *word    = Vec_WrdEntryP( vSims, Id*nWords );
        var Count int32      = Abc_TtCountOnesVec( pSim, nWords );
        var fRareOne int32   = Count < TotalBits/2; // fRareOne is 1 if rare value is 1
        var CountRare int32;
        if fRareOne != 0 {
            CountRare = Count;
        } else {
            CountRare = TotalBits - Count;
        }
        assert(CountRare <= TotalBits/2, "CountRare <= TotalBits/2");
        if CountRare <= RareLimit {
            Vec_IntPushTwo( vRareCounts, Abc_Var2Lit(Id, fRareOne), CountRare ); }
    }
    }
    Vec_WrdFree( vSims );
    return vRareCounts;
}
func Gia_SimQualityImpact(p *Gia_Man_t, vPat *Vec_Int_t, vRareCounts *Vec_Int_t) *Vec_Flt_t {
    var vQuoIncs *Vec_Flt_t  = Vec_FltStart( Gia_ManCiNum(p) );
    var nWordsNew int32  = Abc_Bit6WordNum( 1+Gia_ManCiNum(p) );
    var vSimsPiNew *Vec_Wrd_t  = Vec_WrdStart( Gia_ManCiNum(p) * nWordsNew );
    var vTemp, vSims *Vec_Wrd_t;
    var i, k, Value, RareLit, RareCount int32;
    assert(Vec_IntSize(vPat) == Gia_ManCiNum(p), "Vec_IntSize(vPat) == Gia_ManCiNum(p)");
    // inlining of macro Vec_IntForEachEntry
    for  i = 0; (i < Vec_IntSize(vPat)) && (((Value) = Vec_IntEntry(vPat, i)), 1); i++  {
        var pSim *word  = Vec_WrdEntryP( vSimsPiNew, i*nWordsNew );
        if Value != 0 {
            Abc_TtFill( pSim, nWordsNew ); }
        Abc_TtXorBit( pSim, i+1 );
    }
    vTemp = p.vSimsPi;
    p.vSimsPi = vSimsPiNew;
    vSims = Gia_ManSimPatSim( p );
    p.vSimsPi = vTemp;
    // inlining of macro Vec_IntForEachEntryDouble
    for  i = 0; (i+1 < Vec_IntSize(vRareCounts)) && (((RareLit) = Vec_IntEntry(vRareCounts, i)), 1) && (((RareCount) = Vec_IntEntry(vRareCounts, i+1)), 1); i += 2  {
        var Incrm float32  = float32(1.0)/(RareCount+1);
        var RareObj int32  = Abc_Lit2Var(RareLit);
        var RareVal int32  = Abc_LitIsCompl(RareLit);
        var pSim *word  = Vec_WrdEntryP( vSims, RareObj*nWordsNew );
        var OrigVal int32  = pSim[0] & 1;
        if OrigVal != 0 {
            Abc_TtNot( pSim, nWordsNew ); }
        for  k = 0; k < Gia_ManCiNum(p); k++  {
            if Abc_TtGetBit(pSim, k+1) { // value changed
                Vec_FltAddToEntry( vQuoIncs, k, OrigVal != RareVal ? Incrm : -Incrm ); }
        }
    }
    Vec_WrdFree( vSims );
    Vec_WrdFree( vSimsPiNew );
    return vQuoIncs;
}
func Gia_SimCollectBest(vQuo *Vec_Flt_t) *Vec_Int_t {
    var vRes *Vec_Int_t; var i int32;
    var Value float32;
    var ValueMax float32 = Vec_FltFindMax( vQuo );
    if ValueMax <= 0 {
        return nil; }
    vRes = Vec_IntAlloc( 100 );  // variables with max quo
    // inlining of macro Vec_FltForEachEntry
    for  i = 0; (i < Vec_FltSize(vQuo)) && (((Value) = Vec_FltEntry(vQuo, i)), 1); i++  {
        if Value == ValueMax {
            Vec_IntPush( vRes, i ); }
    }
    return vRes;
}
func Gia_ManPatGetQuo(p *Gia_Man_t, vRareCounts *Vec_Int_t, vSims *Vec_Wrd_t, n int32, nWords int32) float32 {
    var Quality float32;
    var RareLit, RareCount, i int32; 
    assert(Vec_WrdSize(vSims) == Gia_ManObjNum(p), "Vec_WrdSize(vSims) == Gia_ManObjNum(p)");
    // inlining of macro Vec_IntForEachEntryDouble
    for  i = 0; (i+1 < Vec_IntSize(vRareCounts)) && (((RareLit) = Vec_IntEntry(vRareCounts, i)), 1) && (((RareCount) = Vec_IntEntry(vRareCounts, i+1)), 1); i += 2  {
        var Incrm float32  = float32(1.0)/(RareCount+1);
        var RareObj int32  = Abc_Lit2Var(RareLit);
        var RareVal int32  = Abc_LitIsCompl(RareLit);
        var pSim *word  = Vec_WrdEntryP( vSims, RareObj*nWords );
        if Abc_TtGetBit(pSim, n) == RareVal {
            Quality += Incrm; }
    }
    return Quality;
}
func Gia_ManPatGetTotalQuo(p *Gia_Man_t, RareLimit int32, vPatterns *Vec_Wrd_t, nWords int32) float32 {
    var Total float32; var n int32;
    var vRareCounts *Vec_Int_t  = Gia_SimCollectRare( p, vPatterns, RareLimit );
    var vSims *Vec_Wrd_t;
    var vTemp *Vec_Wrd_t = p.vSimsPi;
    p.vSimsPi = vPatterns;
    vSims = Gia_ManSimPatSim( p );
    p.vSimsPi = vTemp;
    for  n = 0; n < 64*nWords; n++  {
        Total += Gia_ManPatGetQuo( p, vRareCounts, vSims, n, nWords );
    }
    Vec_IntFree( vRareCounts );
    Vec_WrdFree( vSims );
    return Total;
}
func Gia_ManPatGetOneQuo(p *Gia_Man_t, RareLimit int32, vPatterns *Vec_Wrd_t, nWords int32, n int32) float32 {
    var Total float32; 
    var vRareCounts *Vec_Int_t  = Gia_SimCollectRare( p, vPatterns, RareLimit );
    var vSims *Vec_Wrd_t;
    var vTemp *Vec_Wrd_t = p.vSimsPi;
    p.vSimsPi = vPatterns;
    vSims = Gia_ManSimPatSim( p );
    p.vSimsPi = vTemp;
    Total += Gia_ManPatGetQuo( p, vRareCounts, vSims, n, nWords );
    Vec_IntFree( vRareCounts );
    Vec_WrdFree( vSims );
    return Total;
}
func Gia_ManPatRareImprove(p *Gia_Man_t, RareLimit int32, fVerbose int32) {
    var clk abctime  = Abc_Clock();
    var FinalTotal, InitTotal float32;
    var n int32;
    var nRares int32;
    var nChanges int32;
    var nWords int32 = Vec_WrdSize(p.vSimsPi) / Gia_ManCiNum(p);
    var vPatterns *Vec_Wrd_t  = p.vSimsPi; p.vSimsPi = nil;
    InitTotal = Gia_ManPatGetTotalQuo( p, RareLimit, vPatterns, nWords );
    for  n = 0; n < 64*nWords; n++  {
        var clk abctime  = Abc_Clock();
        var vRareCounts *Vec_Int_t  = Gia_SimCollectRare( p, vPatterns, RareLimit );
        var vPat *Vec_Int_t         = Gia_ManPatCollectOne( p, vPatterns, n, nWords );
        var vQuoIncs *Vec_Flt_t     = Gia_SimQualityImpact( p, vPat, vRareCounts );
        var vBest *Vec_Int_t        = Gia_SimCollectBest( vQuoIncs );
        if fVerbose != 0 {
            var PatQuo float32  = Gia_ManPatGetOneQuo( p, RareLimit, vPatterns, nWords, n );
            fmt.Printf("Pat %5d : Rare = %4d  Cands = %3d  Value = %8.3f  Change = %8.3f  ", n, Vec_IntSize(vRareCounts)/2, vBest ? Vec_IntSize(vBest) : 0, PatQuo, vBest ? Vec_FltEntry(vQuoIncs, Vec_IntEntry(vBest,0)) : 0);
            Abc_PrintTime( 1, "Time", Abc_Clock() - clk );
        }
        if vBest != nil {
            var VarBest int32  = Vec_IntEntry( vBest, rand()%Vec_IntSize(vBest) );
            Abc_TtXorBit( Vec_WrdEntryP(vPatterns, VarBest*nWords), n );
            nChanges++;
        }
        nRares = Vec_IntSize(vRareCounts)/2;
        Vec_IntFree( vRareCounts );
        Vec_IntFree( vPat );
        Vec_FltFree( vQuoIncs );
        Vec_IntFreeP( &vBest );
    }
    if fVerbose != 0 {
        fmt.Printf("\n"); }
    FinalTotal = Gia_ManPatGetTotalQuo( p, RareLimit, vPatterns, nWords );
    p.vSimsPi = vPatterns;

    fmt.Printf("Improved %d out of %d patterns using %d rare nodes: %.2f -> %.2f.  ", nChanges, 64*nWords, nRares, InitTotal, FinalTotal);
    Abc_PrintTime( 1, "Time", Abc_Clock() - clk );
}

/**Function*************************************************************

  Synopsis    [Trying vectorized simulation.]

  Description []

  SideEffects []

  SeeAlso     []

***********************************************************************/
func Gia_ManSimTest(pGia *Gia_Man_t) {
    var n int32;
    var nWords int32 = 4;
    var vSim1, vSim2 *Vec_Wrd_t;
    var vSim0 *Vec_Wrd_t  = Vec_WrdStartRandom( Gia_ManCiNum(pGia) * nWords );
    var clk abctime  = Abc_Clock();

    pGia.vSimsPi = vSim0;
    for  n = 0; n < 20; n++  {
        vSim1 = Gia_ManSimPatSim( pGia );
        Vec_WrdFree( vSim1 );
    }
    Abc_PrintTime( 1, "Time1", Abc_Clock() - clk );

    clk = Abc_Clock();
    for  n = 0; n < 20; n++  {
        vSim2 = Gia_ManSimPatSim2( pGia );
        Vec_WrdFree( vSim2 );
    }
    Abc_PrintTime( 1, "Time2", Abc_Clock() - clk );

    pGia.vSimsPi = nil;
    Vec_WrdFree( vSim0 );
}

/**Function*************************************************************

  Synopsis    [Trying compiled simulation.]

  Description []

  SideEffects []

  SeeAlso     []

***********************************************************************/
func Gia_ManSimGen(pGia *Gia_Man_t) {
    var nWords int32  = 4;
    var pObj *Gia_Obj_t;
    var vSim0 *Vec_Wrd_t  = Vec_WrdStartRandom( Gia_ManCiNum(pGia) * nWords );
    var pFile *os.File  = os.OpenFile("comp_sim.c", os.O_WRONLY | os.O_CREATE | os.O_TRUNC, 0666);
    var i, k, Id int32;
    fmt.Fprintf(pFile, "#include <stdio.h>\n");
    fmt.Fprintf(pFile, "#include <stdlib.h>\n");
    fmt.Fprintf(pFile, "#include <time.h>\n");
    fmt.Fprintf(pFile, "int main()\n");
    fmt.Fprintf(pFile, "{\n");
    fmt.Fprintf(pFile, "  clock_t clkThis = clock();\n");
    fmt.Fprintf(pFile, "  unsigned long Res = 0;\n");
    fmt.Fprintf(pFile, "  int i;\n");
    fmt.Fprintf(pFile, "  srand(time(NULL));\n");
    fmt.Fprintf(pFile, "  for ( i = 0; i < 2000; i++ )\n");
    fmt.Fprintf(pFile, "  {\n");
    for  k = 0; k < nWords; k++  {
        fmt.Fprintf(pFile, "  unsigned long s%07d_%d = 0x%08x%08x;\n", 0, k, 0, 0);
    }
    // inlining of macro Gia_ManForEachCiId
    for  i = 0; (i < Vec_IntSize(pGia.vCis)) && ((Id) = Gia_ObjId(pGia, Gia_ManCi(pGia, i))); i++  {
        //word * pSim = Vec_WrdEntryP(vSim0, i*nWords);
        //unsigned * pSimU = (unsigned *)pSim;
        for  k = 0; k < nWords; k++  {
            fmt.Fprintf(pFile, "  unsigned long s%07d_%d = ((unsigned long)rand() << 48) | ((unsigned long)rand() << 32) | ((unsigned long)rand() << 16) | (unsigned long)rand();\n", Id, k);
        }
    }
    // inlining of macro Gia_ManForEachAnd
    for  Id = 0; (Id < pGia.nObjs) && ((pObj) = Gia_ManObj(pGia, Id)); Id++  {      if !Gia_ObjIsAnd(pObj) {} else {
        for  k = 0; k < nWords; k++  {
            fmt.Fprintf(pFile, "  unsigned long s%07d_%d = %cs%07d_%d & %cs%07d_%d;\n", Id, k, Gia_ObjFaninC0(pObj) ? '~' : ' ', Gia_ObjFaninId0(pObj, Id), k, Gia_ObjFaninC1(pObj) ? ' ' : '~', Gia_ObjFaninId1(pObj, Id), k);
        }
    }
    }
    // inlining of macro Gia_ManForEachCoId
    for  i = 0; (i < Vec_IntSize(pGia.vCos)) && ((Id) = Gia_ObjId(pGia, Gia_ManCo(pGia, i))); i++  {
        pObj = Gia_ManObj(pGia, Id);
        for  k = 0; k < nWords; k++  {
            fmt.Fprintf(pFile, "  Res ^= %cs%07d_%d;\n", Gia_ObjFaninC0(pObj) ? '~' : ' ', Gia_ObjFaninId0(pObj, Id), k);
        }
    }
    Vec_WrdFree( vSim0 );
    fmt.Fprintf(pFile, "  }\n");
    fmt.Fprintf(pFile, "  printf( \"Res = 0x%%08x    \", (unsigned)Res );\n");
    fmt.Fprintf(pFile, "  printf( \"Time = %%6.2f sec\\n\", (float)(clock() - clkThis)/CLOCKS_PER_SEC );\n");
    fmt.Fprintf(pFile, "  return 1;\n");
    fmt.Fprintf(pFile, "}\n");
    _ = (pFile).Close();
}

////////////////////////////////////////////////////////////////////////
///                       END OF FILE                                ///
////////////////////////////////////////////////////////////////////////


//ABC_NAMESPACE_IMPL_END

