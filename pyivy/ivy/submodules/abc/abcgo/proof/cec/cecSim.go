package cec

import "fmt"

/**CFile****************************************************************

  FileName    [cecSim.c]

  SystemName  [ABC: Logic synthesis and verification system.]

  PackageName [Combinational equivalence checking.]

  Synopsis    [Simulation.]

  Author      [Alan Mishchenko]
  
  Affiliation [UC Berkeley]

  Date        [Ver. 1.0. Started - June 20, 2005.]

  Revision    [$Id: cecSim.c,v 1.00 2005/06/20 00:00:00 alanmi Exp $]

***********************************************************************/

//#include "cecInt.h"
//#include "aig/gia/giaAig.h"
//#include "misc/util/utilTruth.h"


//ABC_NAMESPACE_IMPL_START

////////////////////////////////////////////////////////////////////////
///                        DECLARATIONS                              ///
////////////////////////////////////////////////////////////////////////

const SIM_RANDS =    113

   
type Cec_ManS_t struct {
    nWords              int32;
    nLevelMax              int32;
    nLevelMin              int32;
    iRand              int32;
    pAig       *Gia_Man_t;       
    vInputs       *Vec_Int_t;
    vLevels       *Vec_Wec_t;
    vSims       *Vec_Wrd_t;
    word *           pTemp[4];
    word             Rands[SIM_RANDS];
    nSkipped              int32;
    nVisited              int32;
    nCexes              int32;
    clkSat          abctime;
    clkUnsat          abctime;
}

func Cec_ManSSim(p *Cec_ManS_t, iNode int32, Value int32) *word { return Vec_WrdEntryP(p.vSims, p.nWords*(iNode+iNode+Value)); }

////////////////////////////////////////////////////////////////////////
///                     FUNCTION DEFINITIONS                         ///
////////////////////////////////////////////////////////////////////////

/**Function*************************************************************

  Synopsis    []

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Cec_ManSStart(pAig *Gia_Man_t, nWords int32) *Cec_ManS_t {
    var p *Cec_ManS_t; var i int32;
    // inlining of macro ABC_ALLOC
    p = ((Cec_ManS_t *) malloc(sizeof(Cec_ManS_t) * (size_t)(1)));
    memset( p, 0, sizeof(Cec_ManS_t) );
    p.nWords    = nWords;
    p.pAig      = pAig;
    p.vInputs   = Vec_IntAlloc( 100 );
    p.vLevels   = Vec_WecStart( Gia_ManLevelNum(pAig) + 1 );
    p.vSims     = Vec_WrdStart( Gia_ManObjNum(pAig) * nWords * 2 );
    // inlining of macro ABC_ALLOC
    p.pTemp[0]  = ((word *) malloc(sizeof(word) * (size_t)(4*nWords)));
    for  i = 1; i < 4; i++  {
        p.pTemp[i] = p.pTemp[i-1] + nWords;
    }
    for  i = 0; i < SIM_RANDS; i++  {
        p.Rands[i] = Gia_ManRandomW(0);
    }
    return p;
}
func Cec_ManSStop(p *Cec_ManS_t) {
    Vec_IntFree( p.vInputs );
    Vec_WecFree( p.vLevels );
    Vec_WrdFree( p.vSims );
    // inlining of macro ABC_FREE
    ((p.pTemp[0]) ? (free((char *) (p.pTemp[0])), (p.pTemp[0]) = 0) : 0);
    // inlining of macro ABC_FREE
    ((p) ? (free((char *) (p)), (p) = 0) : 0);
}

/**Function*************************************************************

  Synopsis    [Verify counter-example.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Cec_ManSVerify_rec(p *Gia_Man_t, iObj int32) int32 {
    var Value0, Value1 int32;
    var pObj *Gia_Obj_t  = Gia_ManObj( p, iObj );
    if iObj == 0 { return 0; }
    if Gia_ObjIsTravIdCurrentId(p, iObj) {
        return pObj.fMark1; }
    Gia_ObjSetTravIdCurrentId(p, iObj);
    if Gia_ObjIsCi(pObj) {
        return pObj.fMark1; }
    assert(Gia_ObjIsAnd(pObj), "Gia_ObjIsAnd(pObj)");
    Value0 = Cec_ManSVerify_rec( p, Gia_ObjFaninId0(pObj, iObj) ) ^ Gia_ObjFaninC0(pObj);
    Value1 = Cec_ManSVerify_rec( p, Gia_ObjFaninId1(pObj, iObj) ) ^ Gia_ObjFaninC1(pObj);
    if pObj.fMark1 = Gia_ObjIsXor(pObj) {
        return Value0 ^ Value1;
    } else {
        return Value0 & Value1;
    }
}
func Cec_ManSVerifyTwo(p *Gia_Man_t, iObj0 int32, iObj1 int32) {
    var Value0, Value1 int32;
    Gia_ManIncrementTravId( p );
    Value0 = Cec_ManSVerify_rec( p, iObj0 );
    Value1 = Cec_ManSVerify_rec( p, iObj1 );
    if (Value0 ^ Value1) == Gia_ObjPhaseDiff(p, iObj0, iObj1) {
        fmt.Printf("CEX verification FAILED for obj %d and obj %d.\n", iObj0, iObj1); }
//    else
//        printf( "CEX verification succeeded for obj %d and obj %d.\n", iObj0, iObj1 );;
}
func Cec_ManSVerify(p *Cec_ManS_t, iObj0 int32, iObj1 int32) {
    var fDoVerify int32;
    var w, i, iObj, nCares int32;
    var pCare *word  = Vec_WrdArray(p.vSims);
    if Vec_IntSize(p.vInputs) == 0 {
        fmt.Printf("No primary inputs.\n");
        return;
    }
    // inlining of macro Vec_IntForEachEntry
    for  i = 0; (i < Vec_IntSize(p.vInputs)) && (((iObj) = Vec_IntEntry(p.vInputs, i)), 1); i++  {
        var pSim_0 *word  = Cec_ManSSim( p, iObj, 0 );
        var pSim_1 *word  = Cec_ManSSim( p, iObj, 1 );
        if p.nWords == 1 {
            pCare[0] |= pSim_0[0] & pSim_1[0];
        } else {
            Abc_TtOrAnd( pCare, pSim_0, pSim_1, p.nWords ); }
    }
    nCares = Abc_TtCountOnesVec( pCare, p.nWords );
    if nCares == 64*p.nWords {
        fmt.Printf("No CEXes.\n");
        return;
    }
    assert(Vec_IntSize(p.vInputs) > 0, "Vec_IntSize(p.vInputs) > 0");
    for  w = 0; w < 64*p.nWords; w++  {
        if Abc_TtGetBit(pCare, w) {
            continue; }
        if !(fDoVerify != 0) {
            continue; }
        // inlining of macro Vec_IntForEachEntry
        for  i = 0; (i < Vec_IntSize(p.vInputs)) && (((iObj) = Vec_IntEntry(p.vInputs, i)), 1); i++  {
            Gia_ManObj(p.pAig, iObj).fMark1 = Abc_TtGetBit( Cec_ManSSim(p, iObj, 1), w );
        }
        Cec_ManSVerifyTwo( p.pAig, iObj0, iObj1 );
    }
    fmt.Printf("Considered %d CEXes of nodes %d and %d.\n", 64*p.nWords - nCares, iObj0, iObj1);
}

/**Function*************************************************************

  Synopsis    []

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Cec_ManSRunImply(p *Cec_ManS_t, iNode int32) {
    var pNode *Gia_Obj_t  = Gia_ManObj( p.pAig, iNode );
    if Gia_ObjIsAnd(pNode) {
        var iFan0 int32  = Gia_ObjFaninId0( pNode, iNode );
        var iFan1 int32  = Gia_ObjFaninId1( pNode, iNode );
        var pSim__ *word  = Cec_ManSSim( p, 0, 0 );
        var pSim_0 *word  = Cec_ManSSim( p, iNode, 0 );
        var pSim_1 *word  = Cec_ManSSim( p, iNode, 1 );
        var pSim00 *word  = Cec_ManSSim( p, iFan0,  Gia_ObjFaninC0(pNode) );
        var pSim01 *word  = Cec_ManSSim( p, iFan0, !Gia_ObjFaninC0(pNode) );
        var pSim10 *word  = Cec_ManSSim( p, iFan1,  Gia_ObjFaninC1(pNode) );
        var pSim11 *word  = Cec_ManSSim( p, iFan1, !Gia_ObjFaninC1(pNode) );
        if p.nWords == 1 {
            pSim_0[0] |= pSim00[0] | pSim10[0];
            pSim_1[0] |= pSim01[0] & pSim11[0];
            pSim__[0] |= pSim_0[0] & pSim_1[0];
            pSim_0[0] &= ~pSim__[0];
            pSim_1[0] &= ~pSim__[0];
        
        } else {
            Abc_TtOr( pSim_0, pSim_0, pSim00, p.nWords );
            Abc_TtOr( pSim_0, pSim_0, pSim10, p.nWords );
            Abc_TtOrAnd( pSim_1, pSim01, pSim11, p.nWords );
            Abc_TtOrAnd( pSim__, pSim_0, pSim_1, p.nWords );
            Abc_TtAndSharp( pSim_0, pSim_0, pSim__, p.nWords, 1 );
            Abc_TtAndSharp( pSim_1, pSim_1, pSim__, p.nWords, 1 );
        }
    }
}
func Cec_ManSRunPropagate(p *Cec_ManS_t, iNode int32) int32 {
    var pNode *Gia_Obj_t  = Gia_ManObj( p.pAig, iNode );
    var iFan0 int32  = Gia_ObjFaninId0( pNode, iNode );
    var iFan1 int32  = Gia_ObjFaninId1( pNode, iNode );
    var pSim_0 *word  = Cec_ManSSim( p, iNode, 0 );
    var pSim_1 *word  = Cec_ManSSim( p, iNode, 1 );
    if Abc_TtIsConst0(pSim_0, p.nWords) && Abc_TtIsConst0(pSim_1, p.nWords) {
        p.nSkipped++;
        return 0;
    }
    p.nVisited++;
    //Cec_ManSRunImply( p, iFan0 );
    //Cec_ManSRunImply( p, iFan1 );
    {
        var pSim__ *word  = Cec_ManSSim( p, 0, 0 );
        var pSim00 *word  = Cec_ManSSim( p, iFan0,  Gia_ObjFaninC0(pNode) );
        var pSim01 *word  = Cec_ManSSim( p, iFan0, !Gia_ObjFaninC0(pNode) );
        var pSim10 *word  = Cec_ManSSim( p, iFan1,  Gia_ObjFaninC1(pNode) );
        var pSim11 *word  = Cec_ManSSim( p, iFan1, !Gia_ObjFaninC1(pNode) );
        if p.iRand == SIM_RANDS-1 {
            p.iRand = 0;
        } else {
            p.iRand = p.iRand + 1;
        }
        if p.nWords == 1 {
            pSim01[0] |= pSim_1[0];
            pSim11[0] |= pSim_1[0];

            pSim00[0] |= pSim_0[0] & (pSim11[0] | ~p.Rands[p.iRand]);
            pSim10[0] |= pSim_0[0] & (pSim01[0] |  p.Rands[p.iRand]);

            pSim__[0] |= pSim00[0] & pSim01[0];
            pSim__[0] |= pSim10[0] & pSim11[0];

            pSim00[0] &= ~pSim__[0];
            pSim01[0] &= ~pSim__[0];
            pSim10[0] &= ~pSim__[0];
            pSim11[0] &= ~pSim__[0];
        
        } else {
            var w int32;
            for  w = 0; w < p.nWords; w++  {
                p.pTemp[0][w] = ~p.Rands[(p.iRand + w) % SIM_RANDS];
            }

            Abc_TtOr( pSim01, pSim01, pSim_1, p.nWords );
            Abc_TtOr( pSim11, pSim11, pSim_1, p.nWords );

            Abc_TtOr( p.pTemp[1], pSim11, p.pTemp[0], p.nWords );
            Abc_TtOrAnd( pSim00, pSim_0, p.pTemp[1], p.nWords );

            Abc_TtNot( p.pTemp[0], p.nWords );
            Abc_TtOr( p.pTemp[1], pSim01, p.pTemp[0], p.nWords );
            Abc_TtOrAnd( pSim10, pSim_0, p.pTemp[1], p.nWords );

            Abc_TtOrAnd( pSim__, pSim00, pSim01, p.nWords );
            Abc_TtOrAnd( pSim__, pSim10, pSim11, p.nWords );

            Abc_TtAndSharp( pSim00, pSim00, pSim__, p.nWords, 1 );
            Abc_TtAndSharp( pSim01, pSim01, pSim__, p.nWords, 1 );
            Abc_TtAndSharp( pSim10, pSim10, pSim__, p.nWords, 1 );
            Abc_TtAndSharp( pSim11, pSim11, pSim__, p.nWords, 1 );
        }
    }
    return 1;
}
func Cec_ManSInsert(p *Cec_ManS_t, iNode int32) {
    var pNode *Gia_Obj_t; var Level int32;
    assert(iNode > 0, "iNode > 0");
    if Gia_ObjIsTravIdCurrentId(p.pAig, iNode) {
        return; }
    Gia_ObjSetTravIdCurrentId(p.pAig, iNode);
    pNode = Gia_ManObj( p.pAig, iNode );
    if Gia_ObjIsCi(pNode) {
        Vec_IntPush( p.vInputs, iNode );
        return;
    }
    assert(Gia_ObjIsAnd(pNode), "Gia_ObjIsAnd(pNode)");
    Level = Gia_ObjLevelId( p.pAig, iNode );
    assert(Level > 0, "Level > 0");
    Vec_WecPush( p.vLevels, Level, iNode );
    p.nLevelMax = Abc_MaxInt( p.nLevelMax, Level );
    p.nLevelMin = Abc_MinInt( p.nLevelMin, Level );
    assert(p.nLevelMin <= p.nLevelMax, "p.nLevelMin <= p.nLevelMax");
}
func Cec_ManSRunSimInt(p *Cec_ManS_t) int32 {
    var vLevel *Vec_Int_t; 
    var i, k, iNode, fSolved int32;
    // inlining of macro Vec_WecForEachLevelReverseStartStop
    for  i = p.nLevelMax+1-1; (i >= p.nLevelMin) && (((vLevel) = Vec_WecEntry(p.vLevels, i)), 1); i--  {
        // inlining of macro Vec_IntForEachEntry
        for  k = 0; (k < Vec_IntSize(vLevel)) && (((iNode) = Vec_IntEntry(vLevel, k)), 1); k++  {
            var pNode *Gia_Obj_t  = Gia_ManObj( p.pAig, iNode );
            if !(fSolved != 0) && (Cec_ManSRunPropagate( p, iNode ) != 0) {
                Cec_ManSInsert( p, Gia_ObjFaninId0(pNode, iNode) );
                Cec_ManSInsert( p, Gia_ObjFaninId1(pNode, iNode) );
                if Abc_TtIsConst1(Vec_WrdArray(p.vSims), p.nWords) {
                    fSolved = 1; }
            }
            Abc_TtClear( Cec_ManSSim(p, iNode, 0), 2*p.nWords );
        }
        Vec_IntClear( vLevel );
    }
    //Vec_WecForEachLevel( p->vLevels, vLevel, i )
    //    assert( Vec_IntSize(vLevel) == 0 );
    return fSolved;
}
func Cec_ManSRunSim(p *Cec_ManS_t, iNode0 int32, iNode1 int32) int32 {
    var clk abctime  = Abc_Clock();
    //Vec_Int_t * vLevel; 
    var pNodes [2]int32 = [2]int32{ iNode0, iNode1 };
    var i int32;
    var iNode int32;
    var Status int32;
    var fDiff int32 = Gia_ObjPhaseDiff( p.pAig, iNode0, iNode1 );
    var pSim00 *word  = Cec_ManSSim( p, iNode0, 0 );
    var pSim01 *word  = Cec_ManSSim( p, iNode0, 1 );
    var pSim10 *word  = Cec_ManSSim( p, iNode1,  fDiff );
    var pSim11 *word  = Cec_ManSSim( p, iNode1, !fDiff );
    Abc_TtClear( Vec_WrdArray(p.vSims), p.nWords );
    //for ( i = 0; i < Vec_WrdSize(p->vSims); i++ )
    //    assert( p->vSims->pArray[i] == 0 );
    assert(Vec_IntSize(p.vInputs) == 0, "Vec_IntSize(p.vInputs) == 0");
    if iNode0 == 0 {
        Abc_TtFill( pSim11, p.nWords );
    } else {
        if p.nWords == 1 {
            pSim00[0] = (word)0xFFFFFFFF;
            pSim11[0] = (word)0xFFFFFFFF;
            pSim01[0] = pSim00[0] << 32;
            pSim10[0] = pSim11[0] << 32;
        
        } else {
            assert(p.nWords % 2 == 0, "p.nWords % 2 == 0");
            Abc_TtFill( pSim00,               p.nWords/2 );
            Abc_TtFill( pSim11,               p.nWords/2 );
            Abc_TtFill( pSim01 + p.nWords/2, p.nWords/2 );
            Abc_TtFill( pSim10 + p.nWords/2, p.nWords/2 );
        }
    }
    p.nLevelMin = ABC_INFINITY;
    p.nLevelMax = 0;
    Gia_ManIncrementTravId( p.pAig );
    if iNode0 != 0 {
    Cec_ManSInsert( p, iNode0 ); }
    Cec_ManSInsert( p, iNode1 );
    p.nSkipped = p.nVisited = 0;
    Status = Cec_ManSRunSimInt( p );
    if Status == 0 {
        p.clkSat += Abc_Clock() - clk;
    } else {
        p.clkUnsat += Abc_Clock() - clk; }
//    if ( Status == 0 )
//        printf( "Solving %6d and %6d.  Skipped = %6d.  Visited = %6d.  Cone = %6d.  Min = %3d.  Max = %3d.\n", 
//            iNode0, iNode1, p->nSkipped, p->nVisited, Gia_ManConeSize(p->pAig, pNodes, 2), p->nLevelMin, p->nLevelMax );
    if Status == 0 {
        Cec_ManSVerify( p, iNode0, iNode1 );
p.nCexes++; }
    // inlining of macro Vec_IntForEachEntry
    for  i = 0; (i < Vec_IntSize(p.vInputs)) && (((iNode) = Vec_IntEntry(p.vInputs, i)), 1); i++  {
        Abc_TtClear( Cec_ManSSim(p, iNode, 0), 2*p.nWords );
    }
    Vec_IntClear( p.vInputs );
    return Status;
}
func Cec_ManSRunTest(pAig *Gia_Man_t) {
    var clk abctime  = Abc_Clock();
    var p *Cec_ManS_t;
    var i int32;
    var k int32;
    var nWords int32 = 1;
    Gia_ManRandomW( 1 );
    p = Cec_ManSStart( pAig, nWords );
    // inlining of macro Gia_ManForEachClass0
    for  i = 0; i < Gia_ManObjNum(p.pAig); i++  { if !Gia_ObjIsHead(p.pAig, i) {} else {
        // inlining of macro Gia_ClassForEachObj1
        for  assert(Gia_ObjIsHead(p.pAig, i), "Gia_ObjIsHead(p.pAig, i)"), k = Gia_ObjNext(p.pAig, i); k > 0; k = Gia_ObjNext(p.pAig, k)  {
            Cec_ManSRunSim( p, i, k );
        } }
    }
    fmt.Printf("Detected %d CEXes.  ", p.nCexes);
    Abc_PrintTime( 1, "Time ", Abc_Clock() - clk );
    Abc_PrintTime( 1, "Sat  ", p.clkSat );
    Abc_PrintTime( 1, "Unsat", p.clkUnsat );
    Cec_ManSStop( p );
}


////////////////////////////////////////////////////////////////////////
///                       END OF FILE                                ///
////////////////////////////////////////////////////////////////////////


//ABC_NAMESPACE_IMPL_END

