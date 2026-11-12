package gia

import (
	"fmt"
	"os"
)

/**CFile****************************************************************

  FileName    [giaRex.c]

  SystemName  [ABC: Logic synthesis and verification system.]

  PackageName [Scalable AIG package.]

  Synopsis    [Regular expressions.]

  Author      [Alan Mishchenko]
  
  Affiliation [UC Berkeley]

  Date        [Ver. 1.0. Started - June 20, 2005.]

  Revision    [$Id: giaRex.c,v 1.00 2005/06/20 00:00:00 alanmi Exp $]

***********************************************************************/

//#include "gia.h"
//#include "misc/extra/extra.h"

//ABC_NAMESPACE_IMPL_START


////////////////////////////////////////////////////////////////////////
///                        DECLARATIONS                              ///
////////////////////////////////////////////////////////////////////////

////////////////////////////////////////////////////////////////////////
///                     FUNCTION DEFINITIONS                         ///
////////////////////////////////////////////////////////////////////////

/**Function*************************************************************

  Synopsis    [Simulate AIG with the given sequence.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Gia_ManAutomSimulate(p *Gia_Man_t, vAlpha *Vec_Int_t, pSim string) {
    var pObj, pObjRi, pObjRo *Gia_Obj_t;
    var nInputs int32  = Vec_IntSize(vAlpha);
    var nFrames int32  = len(pSim);
    var i, k int32;
    assert(Gia_ManPiNum(p) == nInputs, "Gia_ManPiNum(p) == nInputs");
    fmt.Printf("Simulating string \"%s\":\n", pSim);
    Gia_ManCleanMark0(p);
    // inlining of macro Gia_ManForEachRo
    for  i = 0; (i < Gia_ManRegNum(p)) && ((pObj) = Gia_ManCi(p, Gia_ManPiNum(p)+i)); i++  {
        pObj.fMark0 = 0;
    }
    for  i = 0; i < nFrames; i++  {
        // inlining of macro Gia_ManForEachPi
        for  k = 0; (k < Gia_ManPiNum(p)) && ((pObj) = Gia_ManCi(p, k)); k++  {
            pObj.fMark0 = int32((Vec_IntFind(vAlpha, pSim[i]) == k));
        }
        // inlining of macro Gia_ManForEachAnd
        for  k = 0; (k < p.nObjs) && ((pObj) = Gia_ManObj(p, k)); k++  {      if !Gia_ObjIsAnd(pObj) {} else {
            pObj.fMark0 = (Gia_ObjFanin0(pObj).fMark0 ^ Gia_ObjFaninC0(pObj)) & 
                           (Gia_ObjFanin1(pObj).fMark0 ^ Gia_ObjFaninC1(pObj)); }
        }
        // inlining of macro Gia_ManForEachCo
        for  k = 0; (k < Vec_IntSize(p.vCos)) && ((pObj) = Gia_ManCo(p, k)); k++  {
            pObj.fMark0 = Gia_ObjFanin0(pObj).fMark0 ^ Gia_ObjFaninC0(pObj);
        }
        // inlining of macro Gia_ManForEachRiRo
        for  k = 0; (k < Gia_ManRegNum(p)) && ((pObjRi) = Gia_ManCo(p, Gia_ManPoNum(p)+k)) && ((pObjRo) = Gia_ManCi(p, Gia_ManPiNum(p)+k)); k++  {
            pObjRo.fMark0 = pObjRi.fMark0;
        }

        fmt.Printf("Frame %d : %c %d\n", i, pSim[i], Gia_ManPo(p, 0).fMark0);
    }
    Gia_ManCleanMark0(p);
}

/**Function*************************************************************

  Synopsis    [Builds 1-hotness contraint.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
/**Function*************************************************************

  Synopsis    [Converting regular expressions into sequential AIGs.]

  Description [http://algs4.cs.princeton.edu/lectures/54RegularExpressions.pdf]
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
// preprocesses the regular expression
// construct sequential AIG for the automaton
func Gia_ManRex2Gia(pStrInit string, fOrder int32, fVerbose int32) *Gia_Man_t {
    var pNew *Gia_Man_t = nil;
    var pTemp *Gia_Man_t; 
    var vAlphas, vStr2Sta, vStaLits *Vec_Int_t;
    var vStaTrans, vStack, vVisited *Vec_Int_t;
    var vInit *Vec_Str_t;
    var pStr string  = Gia_ManRexPreprocess( pStrInit );
    var nStates int32  = Gia_ManRexNumInputs( pStr, &vAlphas, &vStr2Sta );
    var i int32;
    var k int32;
    var iLit int32;
    var Entry int32;
    var nLength int32 = len(pStr);
    var nTravId int32 = 1;
    if fOrder != 0 {
        Vec_IntSort( vAlphas, 0 ); }
//    if ( fVerbose )
    {
        fmt.Printf("Input variable order: ");
        // inlining of macro Vec_IntForEachEntry
        for  k = 0; (k < Vec_IntSize(vAlphas)) && (((Entry) = Vec_IntEntry(vAlphas, k)), 1); k++  {
            fmt.Printf("%c", byte(Entry));
        }
        fmt.Printf("\n");
    }
    // start AIG
    pNew = Gia_ManStart( 1000 );
    pNew.pName = Abc_UtilStrsav( pStrInit );
    for  i = 0; i < Vec_IntSize(vAlphas) + nStates; i++  {
        Gia_ManAppendCi( pNew );
    }
    // prepare automaton
    vStaLits  = Vec_IntStart( nStates + 1 );
    vStaTrans = Vec_IntStartFull( nLength );
    vStack    = Vec_IntAlloc( nLength );
    vVisited  = Vec_IntStartFull( nLength + 1 );
    for  i = 0; i < nLength; i++  {
        var Lp int32  = i;
        if pStr[i] == '(' || pStr[i] == '|' {
            Vec_IntPush( vStack, i );
        } else if pStr[i] == ')' {
            var Or int32  = Vec_IntPop( vStack );
            if pStr[Or] == '|' {
                Lp = Vec_IntPop( vStack );
                Vec_IntWriteEntry( vStaTrans, Lp, Or + 1 );
                Vec_IntWriteEntry( vStaTrans, Or, i );
            
            } else {
                Lp = Or; }
        }
        if i < nLength - 1 && pStr[i+1] == '*' {
            Vec_IntWriteEntry( vStaTrans, Lp, i+1 );
            Vec_IntWriteEntry( vStaTrans, i+1, Lp );
        }
    }
    assert(Vec_IntSize(vStack) == 0, "Vec_IntSize(vStack) == 0");
    if fVerbose != 0 {
        Gia_ManPrintAutom( pStr, vStaTrans ); }

    // create next-state functions for each state
    Gia_ManHashAlloc( pNew );
    for  i = 1; i < nLength; i++  {
        var iThis, iThat, iThisLit, iInputLit int32;
        if Gia_SymbSpecial(pStr[i]) != 0 {
            continue; }
        Gia_ManCollectReached( pStr, vStaTrans, i, vStack, vVisited, nTravId++ );
        if fVerbose != 0 {
            Gia_ManPrintReached( pStr, i, vStack ); }
        // create transitions from this state under this input
        iThis = Vec_IntEntry(vStr2Sta, i);
        iThisLit = Gia_Obj2Lit(pNew, Gia_ManPi(pNew, Vec_IntSize(vAlphas) + iThis));
        iInputLit = Gia_Obj2Lit(pNew, Gia_ManPi(pNew, Vec_IntFind(vAlphas, pStr[i])));
        iLit = Gia_ManHashAnd( pNew, iThisLit, iInputLit );
        // inlining of macro Vec_IntForEachEntry
        for  k = 0; (k < Vec_IntSize(vStack)) && (((Entry) = Vec_IntEntry(vStack, k)), 1); k++  {
            iThat = Vec_IntEntry(vStr2Sta, Entry);
            iLit = Gia_ManHashOr( pNew, iLit, Vec_IntEntry(vStaLits, iThat) );
            Vec_IntWriteEntry( vStaLits, iThat, iLit );
        }
    }
    // create one-hotness
    Vec_IntClear( vStack );
    for  i = 0; i < Vec_IntSize(vAlphas); i++  {
        Vec_IntPush( vStack, Gia_Obj2Lit(pNew, Gia_ManPi(pNew, i)) );
    }
    iLit = Gia_ManBuild1Hot( pNew, vStack );
    // combine with outputs
    // inlining of macro Vec_IntForEachEntry
    for  k = 0; (k < Vec_IntSize(vStaLits)) && (((Entry) = Vec_IntEntry(vStaLits, k)), 1); k++  {
        Vec_IntWriteEntry( vStaLits, k, Gia_ManHashAnd(pNew, iLit, Entry) );
    }
    Gia_ManHashStop( pNew );

    // collect initial state
    Gia_ManCollectReached( pStr, vStaTrans, 0, vStack, vVisited, nTravId++ );
    if fVerbose != 0 {
        Gia_ManPrintReached( pStr, 0, vStack ); }
    vInit = Vec_StrStart( nStates + 1 );
    Vec_StrFill( vInit, nStates, '0' );
    // inlining of macro Vec_IntForEachEntry
    for  k = 0; (k < Vec_IntSize(vStack)) && (((Entry) = Vec_IntEntry(vStack, k)), 1); k++  {
        if pStr[Entry] != '\0' {
            Vec_StrWriteEntry( vInit, Vec_IntEntry(vStr2Sta, Entry), '1' ); }
    }
    if fVerbose != 0 {
        fmt.Printf("Init state = %s\n", Vec_StrArray(vInit)); }

    // add outputs
    Vec_IntPushFirst( vStaLits, Vec_IntPop(vStaLits) );
    assert(Vec_IntSize(vStaLits) == nStates + 1, "Vec_IntSize(vStaLits) == nStates + 1");
    // inlining of macro Vec_IntForEachEntry
    for  i = 0; (i < Vec_IntSize(vStaLits)) && (((iLit) = Vec_IntEntry(vStaLits, i)), 1); i++  {
        Gia_ManAppendCo( pNew, iLit );
    }
    Gia_ManSetRegNum( pNew, nStates );
    pNew = Gia_ManCleanup( pTemp = pNew );
    Gia_ManStop( pTemp );

    // add initial state
    pNew = Gia_ManDupZeroUndc( pTemp = pNew, Vec_StrArray(vInit), 0, 0, 0 );
    Gia_ManStop( pTemp );
    Vec_StrFree( vInit );
/*
    Gia_ManAutomSimulate( pNew, vAlphas, "0" );
    Gia_ManAutomSimulate( pNew, vAlphas, "01" );
    Gia_ManAutomSimulate( pNew, vAlphas, "110" );
    Gia_ManAutomSimulate( pNew, vAlphas, "011" );
    Gia_ManAutomSimulate( pNew, vAlphas, "111" );
    Gia_ManAutomSimulate( pNew, vAlphas, "1111" );
    Gia_ManAutomSimulate( pNew, vAlphas, "1010" );

    Gia_ManAutomSimulate( pNew, vAlphas, "A" );
    Gia_ManAutomSimulate( pNew, vAlphas, "AD" );
    Gia_ManAutomSimulate( pNew, vAlphas, "ABCD" );
    Gia_ManAutomSimulate( pNew, vAlphas, "BCD" );
    Gia_ManAutomSimulate( pNew, vAlphas, "CD" );
*/

    // cleanup
    Vec_IntFree( vAlphas );
    Vec_IntFree( vStr2Sta );
    Vec_IntFree( vStaLits );
    Vec_IntFree( vStaTrans );
    Vec_IntFree( vStack );
    Vec_IntFree( vVisited );
    // inlining of macro ABC_FREE
    pStr = "";
    return pNew;
}


/**Function*************************************************************

  Synopsis    [Transposing 64-bit matrix.]

  Description [Borrowed from "Hacker's Delight", by Henry Warren.]
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Gia_ManAutomTranspose64(A []word) {
    var j, k int32;
    var t word;
    var m word = 0x00000000FFFFFFFF;
    for  j = 32; j != 0; j = j >> 1, m = m ^ (m << j)  {
        for  k = 0; k < 64; k = (k + j + 1) & ~j  {
            t = (A[k] ^ (A[k+j] >> j)) & m;
            A[k] = A[k] ^ t;
            A[k+j] = A[k+j] ^ (t << j);
        }
    }
}

/**Function*************************************************************

  Synopsis    [Simulate AIG with the given sequence.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Gia_ManAutomSim0(p *Gia_Man_t, pObj *Gia_Obj_t, vTemp *Vec_Wrd_t) word {
    if Gia_ObjFaninC0(pObj) {
        return ~Vec_WrdEntry(vTemp, Gia_ObjFaninId0p(p, pObj));
    } else {
        return Vec_WrdEntry(vTemp, Gia_ObjFaninId0p(p, pObj));
    }
}
func Gia_ManAutomSim1(p *Gia_Man_t, pObj *Gia_Obj_t, vTemp *Vec_Wrd_t) word {
    if Gia_ObjFaninC1(pObj) {
        return ~Vec_WrdEntry(vTemp, Gia_ObjFaninId1p(p, pObj));
    } else {
        return Vec_WrdEntry(vTemp, Gia_ObjFaninId1p(p, pObj));
    }
}
func Gia_ManAutomStep(p *Gia_Man_t, Cur word, pNext *word, vTemp *Vec_Wrd_t) word {
    var pObj *Gia_Obj_t; var i int32;
    assert(Gia_ManPoNum(p) == 1, "Gia_ManPoNum(p) == 1");
    assert(Vec_WrdSize(vTemp) >= Gia_ManObjNum(p), "Vec_WrdSize(vTemp) >= Gia_ManObjNum(p)");
    Vec_WrdWriteEntry( vTemp, 0, 0 );
    // inlining of macro Gia_ManForEachPi
    for  i = 0; (i < Gia_ManPiNum(p)) && ((pObj) = Gia_ManCi(p, i)); i++  {
        Vec_WrdWriteEntry( vTemp, Gia_ObjId(p, pObj), ((word)1) << (63-i) );
    }
    // inlining of macro Gia_ManForEachRo
    for  i = 0; (i < Gia_ManRegNum(p)) && ((pObj) = Gia_ManCi(p, Gia_ManPiNum(p)+i)); i++  {
        Vec_WrdWriteEntry( vTemp, Gia_ObjId(p, pObj), ((Cur >> (63-i)) & 1) ? ~((word)0) : 0 );
    }
    // inlining of macro Gia_ManForEachAnd
    for  i = 0; (i < p.nObjs) && ((pObj) = Gia_ManObj(p, i)); i++  {      if !Gia_ObjIsAnd(pObj) {} else {
        Vec_WrdWriteEntry( vTemp, i, Gia_ManAutomSim0(p, pObj, vTemp) & Gia_ManAutomSim1(p, pObj, vTemp) ); }
    }
    // inlining of macro Gia_ManForEachRi
    for  i = 0; (i < Gia_ManRegNum(p)) && ((pObj) = Gia_ManCo(p, Gia_ManPoNum(p)+i)); i++  {
        pNext[i] = Gia_ManAutomSim0(p, pObj, vTemp);
    }
    for  ; i < 64; i++  {
        pNext[i] = 0;
    }
    // transpose
//    for ( i = 0; i < 64; i++ )
//        Extra_PrintBinary( stdout, (unsigned *)&pNext[i], 64 ), Abc_Print( 1, "\n" );
//    printf( "\n" );
    Gia_ManAutomTranspose64( pNext );
//    for ( i = 0; i < 64; i++ )
//        Extra_PrintBinary( stdout, (unsigned *)&pNext[i], 64 ), Abc_Print( 1, "\n" );
//    printf( "\n" );
    // return output values
    return Gia_ManAutomSim0(p, Gia_ManPo(p, 0), vTemp);
}
func Gia_ManAutomWalkOne(p *Gia_Man_t, nSteps int32, vStates *Vec_Wrd_t, vCounts *Vec_Int_t, vTemp *Vec_Wrd_t, Init word) {
    var iState word;
    var Output word;
    var pNext [64]word; 
    var i, k, kMin, Index, IndexMin int32;
    var Count, CountMin int32; 
    for  i = 0; i < nSteps; i++  {
        Output = Gia_ManAutomStep( p, iState, pNext, vTemp );
        // check visited states
        kMin = -1;
        IndexMin = -1;
        CountMin = ABC_INFINITY;  
        for  k = 0; k < Gia_ManPiNum(p); k++  {
            if pNext[k] == Init {
                continue; }
            Index = Vec_WrdFind( vStates, pNext[k] );
            if Index == -1 {
                Count = 0;
            } else {
                Count = Vec_IntEntry( vCounts, Index );
            }
            if CountMin > Count || (CountMin != ABC_INFINITY && Count && (float32(CountMin) / Count) > float32(rand())/RAND_MAX ) {
                CountMin = Count;
                IndexMin = Index;
                kMin = k;
            }
            if CountMin == 0 {
                break; }
        }
        // choose the best state
        if CountMin == ABC_INFINITY {
            for  k = 0; k < Gia_ManPiNum(p); k++  {
                if (Output >> (63-k)) & 1 {
                    fmt.Printf("%c", 'a' + k);
                    fmt.Printf("!");
                }
            }            
            break;
        }
        assert(CountMin < ABC_INFINITY, "CountMin < ABC_INFINITY");
        if IndexMin == -1 {
            assert(CountMin == 0, "CountMin == 0");
            IndexMin = Vec_IntSize(vCounts);
            Vec_IntPush( vCounts, 0 );
            Vec_WrdPush( vStates, pNext[kMin] );
        }
        Vec_IntAddToEntry( vCounts, IndexMin, 1 );
        iState = pNext[kMin];
//Extra_PrintBinary( stdout, (unsigned *)&iState, 64 ); printf( "\n" );
        // print the transition
        fmt.Printf("%c", 'a' + kMin);
        if (Output >> (63-kMin)) & 1 {
            fmt.Printf("!"); }
    }
    fmt.Printf("\n");
}
// find flop variables pointed to by negative edges
func Gia_ManAutomInit(p *Gia_Man_t) word {
    var pObj *Gia_Obj_t; 
    var i, Index int32;
    var Init word;
    // inlining of macro Gia_ManForEachAnd
    for  i = 0; (i < p.nObjs) && ((pObj) = Gia_ManObj(p, i)); i++  {      if !Gia_ObjIsAnd(pObj) {} else {
        if Gia_ObjFaninC0(pObj) && Gia_ObjIsCi(Gia_ObjFanin0(pObj)) {
            Index = Gia_ObjCioId(Gia_ObjFanin0(pObj)) - Gia_ManPiNum(p);
            if Index >= 0 {
                Init |= ((word)1 << (63-Index)); }
        }
        if Gia_ObjFaninC1(pObj) && Gia_ObjIsCi(Gia_ObjFanin1(pObj)) {
            Index = Gia_ObjCioId(Gia_ObjFanin1(pObj)) - Gia_ManPiNum(p);
            if Index >= 0 {
                Init |= ((word)1 << (63-Index)); }
        }
    }
    }
    return Init;
}
func Gia_ManAutomWalk(p *Gia_Man_t, nSteps int32, nWalks int32, fVerbose int32) {
    var vTemp, vStates *Vec_Wrd_t;
    var vCounts *Vec_Int_t; var i int32; var Init word;
    if Gia_ManPoNum(p) != 1 {
        fmt.Printf("AIG should have one primary output.\n");
        return;
    }
    if Gia_ManPiNum(p) > 64 {
        fmt.Printf("Cannot simulate an automaton with more than 64 inputs.\n");
        return;
    }
    if Gia_ManRegNum(p) > 64 {
        fmt.Printf("Cannot simulate an automaton with more than 63 states.\n");
        return;
    }
    vTemp   = Vec_WrdStart( Gia_ManObjNum(p) );
    vStates = Vec_WrdAlloc( 1000 );
    vCounts = Vec_IntAlloc( 1000 );
    Vec_WrdPush( vStates, 0 );
    Vec_IntPush( vCounts, 1 );
    Init = Gia_ManAutomInit( p );
    for  i = 0; i < nWalks; i++  {
        Gia_ManAutomWalkOne( p, nSteps, vStates, vCounts, vTemp, Init );
    }
    if fVerbose != 0 {
        var State word;
        // inlining of macro Vec_WrdForEachEntry
        for  i = 0; (i < Vec_WrdSize(vStates)) && (((State) = Vec_WrdEntry(vStates, i)), 1); i++  {
            State ^= Init;
            fmt.Printf("%3d : ", i);
            Extra_PrintBinary( os.Stdout, (unsigned *)&State, 64 ); 
            fmt.Printf(" %d  ", Vec_IntEntry(vCounts, i));
            fmt.Printf("\n");
        }
        fmt.Printf("\n");
    }
    Vec_WrdFree( vTemp );
    Vec_WrdFree( vStates );
    Vec_IntFree( vCounts );
}

////////////////////////////////////////////////////////////////////////
///                       END OF FILE                                ///
////////////////////////////////////////////////////////////////////////


//ABC_NAMESPACE_IMPL_END

