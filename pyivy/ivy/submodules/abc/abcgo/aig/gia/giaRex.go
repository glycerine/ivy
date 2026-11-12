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
/**Function*************************************************************

  Synopsis    [Transposing 64-bit matrix.]

  Description [Borrowed from "Hacker's Delight", by Henry Warren.]
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
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

