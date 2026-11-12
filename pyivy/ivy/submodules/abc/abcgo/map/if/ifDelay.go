package if_

import "fmt"

/**CFile****************************************************************

  FileName    [ifDelay.c]

  SystemName  [ABC: Logic synthesis and verification system.]

  PackageName [FPGA mapping based on priority cuts.]

  Synopsis    [Delay balancing for cut functions.]

  Author      [Alan Mishchenko]
  
  Affiliation [UC Berkeley]

  Date        [Ver. 1.0. Started - November 21, 2006.]

  Revision    [$Id: ifDelay.c,v 1.00 2006/11/21 00:00:00 alanmi Exp $]

***********************************************************************/

//#include "if.h"
//#include "ifCount.h"
//#include "bool/kit/kit.h"

//ABC_NAMESPACE_IMPL_START

////////////////////////////////////////////////////////////////////////
///                        DECLARATIONS                              ///
////////////////////////////////////////////////////////////////////////

const IF_MAX_CUBES = 70

////////////////////////////////////////////////////////////////////////
///                     FUNCTION DEFINITIONS                         ///
////////////////////////////////////////////////////////////////////////
 
/**Function*************************************************************

  Synopsis    [Computes the SOP delay using balanced AND decomposition.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func If_CutMaxCubeSize(vCover *Vec_Int_t, nVars int32) int32 {
    var i, k, Entry, Literal, Count, CountMax int32;
    Vec_IntForEachEntry( vCover, Entry, i )
    {
        Count = 0;
        for  k = 0; k < nVars; k++  {
            Literal = (3 & (Entry >> (k << 1)));
            if Literal == 1 || Literal == 2 {
                Count++; }
        }
        CountMax = Abc_MaxInt( CountMax, Count );
    }
    return CountMax;
}
var spStatic_If_CutDelaySop_GateDelays [20]float64 = [20]float64{ 1.00, 1.00, 2.00, 2.58, 3.00, 3.32, 3.58, 3.81, 4.00, 4.17, 4.32, 4.46, 4.58, 4.70, 4.81, 4.91, 5.00, 5.09, 5.17, 5.25 };
func If_CutDelaySop(p *If_Man_t, pCut *If_Cut_t) int32 {
    var pPerm string  = If_CutPerm( pCut );
    // delay is calculated using 1+log2(NumFanins)
    
    var vCover *Vec_Int_t;
    var pLeaf *If_Obj_t;
    var i, nLitMax, Delay, DelayMax int32;
    // mark cut as a user cut
    pCut.fUser = 1;
    if pCut.nLeaves == 0 {
        return 0; }
    if pCut.nLeaves == 1 {
        return int32(If_ObjCutBest(If_CutLeaf(p, pCut, 0)).Delay); }
    vCover = Vec_WecEntry( p.vTtIsops[pCut.nLeaves], Abc_Lit2Var(If_CutTruthLit(pCut)) );
    if Vec_IntSize(vCover) == 0 {
        return -1; }
    // mark the output as complemented
//    vAnds = If_CutDelaySopAnds( p, pCut, vCover, RetValue ^ pCut->fCompl );
    if Vec_IntSize(vCover) > p.pPars.nGateSize {
        return -1; }
    // set the area cost
    assert(If_CutLeaveNum(pCut) >= 0 && If_CutLeaveNum(pCut) <= 16, "If_CutLeaveNum(pCut) >= 0 && If_CutLeaveNum(pCut) <= 16");
    // compute the gate delay
    nLitMax = If_CutMaxCubeSize( vCover, If_CutLeaveNum(pCut) );
    if Vec_IntSize(vCover) < 2 {
        pCut.Cost = Vec_IntSize(vCover);
        Delay = int32((spStatic_If_CutDelaySop_GateDelays[If_CutLeaveNum(pCut)] + 0.5));
        DelayMax = 0;
        If_CutForEachLeaf( p, pCut, pLeaf, i )
            DelayMax = Abc_MaxInt( DelayMax, If_ObjCutBest(pLeaf).Delay + (pPerm[i] = byte(Delay)) );
    
    } else {
        pCut.Cost = Vec_IntSize(vCover) + 1;
        Delay = int32((spStatic_If_CutDelaySop_GateDelays[If_CutLeaveNum(pCut)] + spStatic_If_CutDelaySop_GateDelays[nLitMax] + 0.5));
        DelayMax = 0;
        If_CutForEachLeaf( p, pCut, pLeaf, i )
            DelayMax = Abc_MaxInt( DelayMax, If_ObjCutBest(pLeaf).Delay + (pPerm[i] = byte(Delay)) );
    }
    return DelayMax;
}


/**Function*************************************************************

  Synopsis    [Compute pin delays.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func If_CutSopBalancePinDelaysInt(vCover *Vec_Int_t, pTimes *int32, pFaninRes *word, nSuppAll int32, pRes *word) int32 {
    var pPinDelsAnd [IF_MAX_FUNC_LUTSIZE]word;
    var pPinDelsOr [IF_MAX_CUBES]word;
    var nCounterAnd int32;
    var pCounterAnd [IF_MAX_FUNC_LUTSIZE]int32;
    var nCounterOr int32;
    var pCounterOr [IF_MAX_CUBES]int32;
    var i, k, Entry, Literal, Delay int32;
    var ResAnd word;
    if Vec_IntSize(vCover) > IF_MAX_CUBES {
        return -1; }
    nCounterOr = 0;
    Vec_IntForEachEntry( vCover, Entry, i )
    { 
        nCounterAnd = 0;
        for  k = 0; k < nSuppAll; k++  {
            Literal = 3 & (Entry >> (k << 1));
            if Literal == 1 || Literal == 2 { // neg or pos literal
                Delay = If_LogCounterPinDelays( pCounterAnd, &nCounterAnd, pPinDelsAnd, pTimes[k], pFaninRes[k], nSuppAll, 0 );
            } else if Literal != 0 { 
                assert(false, "0"); }
        }
        assert(nCounterAnd > 0, "nCounterAnd > 0");
        ResAnd = If_LogPinDelaysMulti( pPinDelsAnd, nCounterAnd, nSuppAll, 0 );
        Delay = If_LogCounterPinDelays( pCounterOr, &nCounterOr, pPinDelsOr, Delay, ResAnd, nSuppAll, 0 );
    }
    assert(nCounterOr > 0, "nCounterOr > 0");
    *pRes = If_LogPinDelaysMulti( pPinDelsOr, nCounterOr, nSuppAll, 0 );
    return Delay;
}
func If_CutSopBalancePinDelaysIntInt(vCover *Vec_Int_t, pTimes *int32, nSuppAll int32, pPerm string) int32 {
    var i, Delay int32;
    var Res word;
    var FaninRes [IF_MAX_FUNC_LUTSIZE]word;
    for  i = 0; i < nSuppAll; i++  {
        FaninRes[i] = If_CutPinDelayInit(i);
    }
    Delay = If_CutSopBalancePinDelaysInt( vCover, pTimes, FaninRes, nSuppAll, &Res );
    If_CutPinDelayTranslate( Res, nSuppAll, pPerm );
    return Delay;
}
func If_CutSopBalancePinDelays(p *If_Man_t, pCut *If_Cut_t, pPerm string) int32 {
    if pCut.nLeaves == 0 { // const
        return 0; }
    if pCut.nLeaves == 1 { // variable
    
        pPerm[0] = 0;
        return int32(If_ObjCutBest(If_CutLeaf(p, pCut, 0)).Delay);
    
    } else {
        var vCover *Vec_Int_t;
        var i int32;
        var pTimes [IF_MAX_FUNC_LUTSIZE]int32;
        vCover = Vec_WecEntry( p.vTtIsops[pCut.nLeaves], Abc_Lit2Var(If_CutTruthLit(pCut)) );
        if Vec_IntSize(vCover) == 0 {
            return -1; }
        for  i = 0; i < If_CutLeaveNum(pCut); i++  {
            pTimes[i] = int32(If_ObjCutBest(If_CutLeaf(p, pCut, i)).Delay);
        } 
        return If_CutSopBalancePinDelaysIntInt( vCover, pTimes, If_CutLeaveNum(pCut), pPerm );
    }
}

/**Function*************************************************************

  Synopsis    [Evaluate delay using SOP balancing.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func If_CutSopBalanceEvalInt(vCover *Vec_Int_t, pTimes *int32, pFaninLits *int32, vAig *Vec_Int_t, piRes *int32, nSuppAll int32, pArea *int32) int32 {
    var nCounterAnd int32;
    var pCounterAnd [IF_MAX_FUNC_LUTSIZE]int32;
    var pFaninLitsAnd [IF_MAX_FUNC_LUTSIZE]int32;
    var nCounterOr int32;
    var pCounterOr [IF_MAX_CUBES]int32;
    var pFaninLitsOr [IF_MAX_CUBES]int32;
    var i, k, Entry, Literal, nLits, Delay, iRes int32;
    if Vec_IntSize(vCover) > IF_MAX_CUBES {
        return -1; }
    nCounterOr = 0;
    Vec_IntForEachEntry( vCover, Entry, i )
    { 
        nCounterAnd = nLits = 0;
        for  k = 0; k < nSuppAll; k++  {
            Literal = 3 & (Entry >> (k << 1));
            if Literal == 1 { // neg literal
                nLits++, Delay = If_LogCounterAddAig( pCounterAnd, &nCounterAnd, pFaninLitsAnd, pTimes[k], vAig ? Abc_LitNot(pFaninLits[k]) : -1, vAig, nSuppAll, 0, 0 );
            } else if Literal == 2 { // pos literal
                nLits++, Delay = If_LogCounterAddAig( pCounterAnd, &nCounterAnd, pFaninLitsAnd, pTimes[k], vAig ? pFaninLits[k] : -1, vAig, nSuppAll, 0, 0 );
            } else if Literal != 0 { 
                assert(false, "0"); }
        }
        assert(nCounterAnd > 0, "nCounterAnd > 0");
        assert(nLits > 0, "nLits > 0");
        if vAig != nil {
            iRes = If_LogCreateAndXorMulti( vAig, pFaninLitsAnd, nCounterAnd, nSuppAll, 0 );
        } else {
            *pArea += nLits == 1 ? 0 : nLits - 1; }
        Delay = If_LogCounterAddAig( pCounterOr, &nCounterOr, pFaninLitsOr, Delay, vAig ? Abc_LitNot(iRes) : -1, vAig, nSuppAll, 0, 0 );
    }
    assert(nCounterOr > 0, "nCounterOr > 0");
    if vAig != nil {
        *piRes = Abc_LitNot( If_LogCreateAndXorMulti( vAig, pFaninLitsOr, nCounterOr, nSuppAll, 0 ) );
        if (vCover.nCap >> 16) & 1 {  // hack to remember complemented attribute
            *piRes = Abc_LitNot( *piRes ); }
    
    } else {       
        *pArea += Vec_IntSize(vCover) == 1 ? 0 : Vec_IntSize(vCover) - 1; }
    return Delay;
}
func If_CutSopBalanceEvalIntInt(vCover *Vec_Int_t, nLeaves int32, pTimes *int32, vAig *Vec_Int_t, fCompl int32, pArea *int32) int32 {
    var pFaninLits [IF_MAX_FUNC_LUTSIZE]int32;
    var iRes, Res, k int32;
    if vAig != nil {
        for  k = 0; k < nLeaves; k++  {
            pFaninLits[k] = Abc_Var2Lit(k, 0);
        } }
    Res = If_CutSopBalanceEvalInt( vCover, pTimes, pFaninLits, vAig, &iRes, nLeaves, pArea );
    if Res == -1 {
        return -1; }
    assert(vAig == nil || Abc_Lit2Var(iRes) == nLeaves + Abc_Lit2Var(Vec_IntSize(vAig)) - 1, "vAig == nil || Abc_Lit2Var(iRes) == nLeaves + Abc_Lit2Var(Vec_IntSize(vAig)) - 1");
    if vAig != nil {
        Vec_IntPush( vAig, Abc_LitIsCompl(iRes) ^ fCompl ); }
    assert(vAig == nil || (Vec_IntSize(vAig) & 1), "vAig == nil || (Vec_IntSize(vAig) & 1)");
    return Res;
}
func If_CutSopBalanceEval(p *If_Man_t, pCut *If_Cut_t, vAig *Vec_Int_t) int32 {
    pCut.fUser = 1;
    if vAig != nil {
        Vec_IntClear( vAig ); }
    if pCut.nLeaves == 0 { // const
    
        assert(Abc_Lit2Var(If_CutTruthLit(pCut)) == 0, "Abc_Lit2Var(If_CutTruthLit(pCut)) == 0");
        if vAig != nil {
            Vec_IntPush( vAig, Abc_LitIsCompl(If_CutTruthLit(pCut)) ); }
        pCut.Cost = 0;
        return 0;
    }
    if pCut.nLeaves == 1 { // variable
    
        assert(Abc_Lit2Var(If_CutTruthLit(pCut)) == 1, "Abc_Lit2Var(If_CutTruthLit(pCut)) == 1");
        if vAig != nil {
            Vec_IntPush( vAig, 0 ); }
        if vAig != nil {
            Vec_IntPush( vAig, Abc_LitIsCompl(If_CutTruthLit(pCut)) ); }
        pCut.Cost = 0;
        return int32(If_ObjCutBest(If_CutLeaf(p, pCut, 0)).Delay);
    
    } else {
        var fVerbose int32;
        var vCover *Vec_Int_t  = Vec_WecEntry( p.vTtIsops[pCut.nLeaves], Abc_Lit2Var(If_CutTruthLit(pCut)) );
        var Delay, Area int32;
        var i int32;
        var pTimes [IF_MAX_FUNC_LUTSIZE]int32;
        if vCover == nil {
            return -1; }
        assert(Vec_IntSize(vCover) > 0, "Vec_IntSize(vCover) > 0");
        for  i = 0; i < If_CutLeaveNum(pCut); i++  {
            pTimes[i] = int32(If_ObjCutBest(If_CutLeaf(p, pCut, i)).Delay);
        } 
        Delay = If_CutSopBalanceEvalIntInt( vCover, If_CutLeaveNum(pCut), pTimes, vAig, Abc_LitIsCompl(If_CutTruthLit(pCut)) ^ pCut.fCompl, &Area );
        pCut.Cost = Area;
        if fVerbose != 0 {
            var Max, Two int32;
            for  i = 0; i < If_CutLeaveNum(pCut); i++  {
                Max = Abc_MaxInt( Max, pTimes[i] );
            }
            for  i = 0; i < If_CutLeaveNum(pCut); i++  {
                if pTimes[i] != Max {
                    Two = Abc_MaxInt( Two, pTimes[i] ); }
            }
            if Two + 2 < Max && Max + 3 < Delay {
                for  i = 0; i < If_CutLeaveNum(pCut); i++  {
                    fmt.Printf("%3d ", pTimes[i]);
                }
                for  ; i < p.pPars.nLutSize; i++  {
                    fmt.Printf("    ");
                }
                fmt.Printf("-> %3d   ", Delay);
                Dau_DsdPrintFromTruth( If_CutTruthW(p, pCut), If_CutLeaveNum(pCut) );
                Kit_TruthIsopPrintCover( vCover, If_CutLeaveNum(pCut), Abc_LitIsCompl(If_CutTruthLit(pCut)) ^ pCut.fCompl );
                {
                    var vIsop Vec_Int_t;
                    var pIsop [64]int32;
                    vIsop.nCap = vIsop.nSize = Abc_Tt6Esop( *If_CutTruthW(p, pCut), pCut.nLeaves, pIsop );
                    vIsop.pArray = pIsop;
                    fmt.Printf("ESOP (%d -> %d)\n", Vec_IntSize(vCover), vIsop.nSize);
                    Kit_TruthIsopPrintCover( &vIsop, If_CutLeaveNum(pCut), 0 );
                }
                fmt.Printf("\n");
            }
        }
        return Delay;
    }
}

/**Function*************************************************************

  Synopsis    [Evaluate delay using SOP balancing.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func If_CutLutBalancePinDelays(p *If_Man_t, pCut *If_Cut_t, pPerm string) int32 {
    if pCut.nLeaves == 0 { // const
        return 0; }
    if pCut.nLeaves == 1 { // variable
    
        pPerm[0] = 0;
        return int32(If_ObjCutBest(If_CutLeaf(p, pCut, 0)).Delay);
    
    } else {
        var pCutPerm string  = If_CutDsdPerm( p, pCut );
        var LutSize int32  = p.pPars.pLutStruct[0] - '0';
        var i int32;
        var Delay int32;
        var DelayMax int32 = -1;
        assert((If_CutLeaveNum(pCut) > LutSize) == (pCut.uMaskFunc > 0), "(If_CutLeaveNum(pCut) > LutSize) == (pCut.uMaskFunc > 0)");
        for  i = 0; i < If_CutLeaveNum(pCut); i++  {
            if If_CutLeaveNum(pCut) > LutSize && ((pCut.uMaskFunc >> (i << 1)) & 1) {
                pPerm[Abc_Lit2Var(int32(pCutPerm[i]))] = 2;
            } else {
                pPerm[Abc_Lit2Var(int32(pCutPerm[i]))] = 1; }
        }
        for  i = 0; i < If_CutLeaveNum(pCut); i++  {
            Delay = int32(If_ObjCutBest(If_CutLeaf(p, pCut, i)).Delay);
            DelayMax = Abc_MaxInt( DelayMax, Delay + int32(pPerm[i]) );
        }
        return DelayMax;
    }
}

/**Function*************************************************************

  Synopsis    [Evaluate delay using SOP balancing.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func If_CutLutBalanceEval(p *If_Man_t, pCut *If_Cut_t) int32 {
    pCut.fUser = 1;
    if pCut.nLeaves > 1 {
        pCut.Cost = 1;
    } else {
        pCut.Cost = 0;
    }
    pCut.uMaskFunc = 0;
    if pCut.nLeaves == 0 { // const
    
        assert(Abc_Lit2Var(If_CutTruthLit(pCut)) == 0, "Abc_Lit2Var(If_CutTruthLit(pCut)) == 0");
        return 0;
    }
    if pCut.nLeaves == 1 { // variable
    
        assert(Abc_Lit2Var(If_CutTruthLit(pCut)) == 1, "Abc_Lit2Var(If_CutTruthLit(pCut)) == 1");
        return int32(If_ObjCutBest(If_CutLeaf(p, pCut, 0)).Delay);
    
    } else {
        var pCutPerm string  = If_CutDsdPerm( p, pCut );
        var LutSize int32  = p.pPars.pLutStruct[0] - '0';
        var i int32;
        var pTimes [IF_MAX_FUNC_LUTSIZE]int32;
        var DelayMax int32 = -1;
        var nLeafMax int32;
        var uLeafMask uint32;
        for  i = 0; i < If_CutLeaveNum(pCut); i++  {
            pTimes[i] = int32(If_ObjCutBest(If_CutLeaf(p, pCut, Abc_Lit2Var(int32(pCutPerm[i])))).Delay); 
            if DelayMax < pTimes[i] {
                DelayMax = pTimes[i];
nLeafMax = 1;
uLeafMask = (1 << (i << 1));
            } else if DelayMax == pTimes[i] {
                nLeafMax++, uLeafMask |= (1 << (i << 1)); }
        }
        if If_CutLeaveNum(pCut) <= LutSize {
            return DelayMax + 1; }
        pCut.Cost = 2;
        if nLeafMax <= LutSize - 1 {
            pCut.uMaskFunc = If_DsdManCheckXY( p.pIfDsdMan, If_CutDsdLit(p, pCut), LutSize, 1, uLeafMask, 0, 0 );
            if pCut.uMaskFunc > 0 {
                return DelayMax + 1; }
        }
        pCut.uMaskFunc = If_DsdManCheckXY( p.pIfDsdMan, If_CutDsdLit(p, pCut), LutSize, 1, 0, 0, 0 );
        if pCut.uMaskFunc == 0 {
            return -1; }
        return DelayMax + 2;
    }
}
/*
int If_CutLutBalanceEval( If_Man_t * p, If_Cut_t * pCut )
{
    char pPerm[16];
    int Delay2, Delay = If_CutLutBalanceEvalInt( p, pCut );
    if ( Delay == -1 )
        return -1;
    Delay2 = If_CutLutBalancePinDelays( p, pCut, pPerm );
    if ( Delay2 != Delay )
    {
        int s = 0;
        char * pCutPerm = If_CutDsdPerm( p, pCut );
        If_DsdManPrintNode( p->pIfDsdMan, If_CutDsdLit(p, pCut) );        Dau_DecPrintSet( pCut->uMaskFunc, pCut->nLeaves, 1 );
        Kit_DsdPrintFromTruth( If_CutTruthUR(p, pCut), pCut->nLeaves ); printf( "\n" );
        for ( s = 0; s < pCut->nLeaves; s++ )
//            printf( "%d ", (int)If_ObjCutBest(If_CutLeaf(p, pCut, Abc_Lit2Var((int)pCutPerm[s])))->Delay );
            printf( "%d ", (int)If_ObjCutBest(If_CutLeaf(p, pCut, s))->Delay );
        printf( "\n" );

        Delay  = If_CutLutBalanceEvalInt( p, pCut );
        Delay2 = If_CutLutBalancePinDelays( p, pCut, pPerm );
    }

    return Delay;
}
*/

////////////////////////////////////////////////////////////////////////
///                       END OF FILE                                ///
////////////////////////////////////////////////////////////////////////


//ABC_NAMESPACE_IMPL_END

