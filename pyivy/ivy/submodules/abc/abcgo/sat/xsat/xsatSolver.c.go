package xsat

/**CFile****************************************************************

  FileName    [xsatSolver.c]

  SystemName  [ABC: Logic synthesis and verification system.]

  PackageName [xSAT - A SAT solver written in C.
               Read the license file for more info.]

  Synopsis    [Solver internal functions implementation.]

  Author      [Bruno Schmitt <boschmitt@inf.ufrgs.br>]

  Affiliation [UC Berkeley / UFRGS]

  Date        [Ver. 1.0. Started - November 10, 2016.]

  Revision    []

***********************************************************************/
////////////////////////////////////////////////////////////////////////
///                          INCLUDES                                ///
////////////////////////////////////////////////////////////////////////
//#include <stdio.h>
//#include <assert.h>
//#include <string.h>
//#include <math.h>

//#include "xsatHeap.h"
//#include "xsatSolver.h"
//#include "xsatUtils.h"

//ABC_NAMESPACE_IMPL_START

////////////////////////////////////////////////////////////////////////
///                     FUNCTION DECLARATIONS                        ///
////////////////////////////////////////////////////////////////////////
/**Function*************************************************************

  Synopsis    []

  Description []

  SideEffects []

  SeeAlso     []

***********************************************************************/
func xSAT_SolverDecide(s *xSAT_Solver_t) int32 {
    var NextVar int32  = VarUndef;

    for  NextVar == VarUndef || Vec_StrEntry( s.vAssigns, NextVar ) != VarX  {
        if xSAT_HeapSize( s.hOrder ) == 0 {
            NextVar = VarUndef;
            break;
        
        } else {
            NextVar = xSAT_HeapRemoveMin( s.hOrder ); }
    }
    return NextVar;
}

/**Function*************************************************************

  Synopsis    []

  Description []

  SideEffects []

  SeeAlso     []

***********************************************************************/
func xSAT_SolverRebuildOrderHeap(s *xSAT_Solver_t) {
    var vTemp *Vec_Int_t  = Vec_IntAlloc( Vec_StrSize( s.vAssigns ) );
    var Var int32;

    for  Var = 0; Var < Vec_StrSize( s.vAssigns ); Var++  {
        if Vec_StrEntry( s.vAssigns, Var ) == VarX {
            Vec_IntPush( vTemp, Var ); }
    }

    xSAT_HeapBuild( s.hOrder, vTemp );
    Vec_IntFree( vTemp );
}

/**Function*************************************************************

  Synopsis    []

  Description []

  SideEffects []

  SeeAlso     []

***********************************************************************/
func xSAT_SolverVarActRescale(s *xSAT_Solver_t) {
    var i int32;
    var pActivity *uint32  = ( unsigned * ) Vec_IntArray( s.vActivity );

    for  i = 0; i < Vec_IntSize( s.vActivity ); i++  {
        pActivity[i] >>= 19;
    }

    s.nVarActInc >>= 19;
    s.nVarActInc = Abc_MaxInt( s.nVarActInc, ( 1 << 5 ) );
}

/**Function*************************************************************

  Synopsis    []

  Description []

  SideEffects []

  SeeAlso     []

***********************************************************************/
func xSAT_SolverVarActBump(s *xSAT_Solver_t, Var int32) {
    var pActivity *uint32  = ( unsigned * ) Vec_IntArray( s.vActivity );

    pActivity[Var] += s.nVarActInc;
    if pActivity[Var] & 0x80000000 {
        xSAT_SolverVarActRescale( s ); }

    if xSAT_HeapInHeap( s.hOrder, Var ) {
        xSAT_HeapDecrease( s.hOrder, Var ); }
}

/**Function*************************************************************

  Synopsis    []

  Description []

  SideEffects []

  SeeAlso     []

***********************************************************************/
func xSAT_SolverVarActDecay(s *xSAT_Solver_t) {
    s.nVarActInc += ( s.nVarActInc >> 4 );
}

/**Function*************************************************************

  Synopsis    []

  Description []

  SideEffects []

  SeeAlso     []

***********************************************************************/
func xSAT_SolverClaActRescale(s *xSAT_Solver_t) {
    var pC *xSAT_Clause_t;
    var i, CRef int32;

    // inlining of macro Vec_IntForEachEntry
    for  i = 0; (i < Vec_IntSize(s.vLearnts)) && (((CRef) = Vec_IntEntry(s.vLearnts, i)), 1); i++  {
        pC = xSAT_SolverReadClause( s, uint32(   CRef) );
        pC.pData[pC.nSize].Act >>= 14;
    }
    s.nClaActInc >>= 14;
    s.nClaActInc = Abc_MaxInt( s.nClaActInc, ( 1 << 10 ) );
}

/**Function*************************************************************

  Synopsis    []

  Description []

  SideEffects []

  SeeAlso     []

***********************************************************************/
func xSAT_SolverClaActBump(s *xSAT_Solver_t, pCla *xSAT_Clause_t) {
    pCla.pData[pCla.nSize].Act += s.nClaActInc;
    if pCla.pData[pCla.nSize].Act & 0x80000000 {
        xSAT_SolverClaActRescale( s ); }
}

/**Function*************************************************************

  Synopsis    []

  Description []

  SideEffects []

  SeeAlso     []

***********************************************************************/
func xSAT_SolverClaActDecay(s *xSAT_Solver_t) {
    s.nClaActInc += ( s.nClaActInc >> 10 );
}

/**Function*************************************************************

  Synopsis    []

  Description []

  SideEffects []

  SeeAlso     []

***********************************************************************/
func xSAT_SolverClaCalcLBD(s *xSAT_Solver_t, pCla *xSAT_Clause_t) int32 {
    var i int32;
    var nLBD int32;

    s.nStamp++;
    for  i = 0; i < pCla.nSize; i++  {
        var Level int32  = Vec_IntEntry( s.vLevels, xSAT_Lit2Var( pCla.pData[i].Lit ) );
        if uint32(   Vec_IntEntry( s.vStamp, Level )) != s.nStamp {
            Vec_IntWriteEntry( s.vStamp, Level, int32(   s.nStamp) );
            nLBD++;
        }
    }
    return nLBD;
}

func xSAT_SolverClaCalcLBD2(s *xSAT_Solver_t, vLits *Vec_Int_t) int32 {
    var i int32;
    var nLBD int32;

    s.nStamp++;
    for  i = 0; i < Vec_IntSize( vLits ); i++  {
        var Level int32  = Vec_IntEntry( s.vLevels, xSAT_Lit2Var( Vec_IntEntry( vLits, i ) ) );
        if uint32(   Vec_IntEntry( s.vStamp, Level )) != s.nStamp {
            Vec_IntWriteEntry( s.vStamp, Level, int32(   s.nStamp) );
            nLBD++;
        }
    }
    return nLBD;
}

/**Function*************************************************************

  Synopsis    []

  Description []

  SideEffects []

  SeeAlso     []

***********************************************************************/
func xSAT_SolverClaNew(s *xSAT_Solver_t, vLits *Vec_Int_t, fLearnt int32) uint32 {
    var CRef uint32;
    var pCla *xSAT_Clause_t;
    var w1 xSAT_Watcher_t;
    var w2 xSAT_Watcher_t;
    var nWords uint32;

    assert(Vec_IntSize( vLits ) > 1, "Vec_IntSize( vLits ) > 1");
    assert(fLearnt == 0 || fLearnt == 1, "fLearnt == 0 || fLearnt == 1");

    nWords = 3 + fLearnt + Vec_IntSize( vLits );
    CRef = xSAT_MemAppend( s.pMemory, nWords );
    pCla = xSAT_SolverReadClause( s, CRef );
    pCla.fLearnt = fLearnt;
    pCla.fMark = 0;
    pCla.fReallocd = 0;
    pCla.fCanBeDel = fLearnt;
    pCla.nSize = Vec_IntSize( vLits );
    memcpy( &( pCla.pData[0].Lit ), Vec_IntArray( vLits ), sizeof( int ) * Vec_IntSize( vLits ) );

    if fLearnt != 0 {
        Vec_IntPush( s.vLearnts, CRef );
        pCla.nLBD = xSAT_SolverClaCalcLBD2( s, vLits );
        pCla.pData[pCla.nSize].Act = 0;
        s.Stats.nLearntLits += Vec_IntSize( vLits );
        xSAT_SolverClaActBump(s, pCla);
    
    } else {
        Vec_IntPush( s.vClauses, CRef );
        s.Stats.nClauseLits += Vec_IntSize( vLits );
    }

    w1.CRef = CRef;
    w2.CRef = CRef;
    w1.Blocker = pCla.pData[1].Lit;
    w2.Blocker = pCla.pData[0].Lit;

    if Vec_IntSize( vLits ) == 2 {
        xSAT_WatchListPush( xSAT_VecWatchListEntry( s.vBinWatches, xSAT_NegLit( pCla.pData[0].Lit ) ), w1 );
        xSAT_WatchListPush( xSAT_VecWatchListEntry( s.vBinWatches, xSAT_NegLit( pCla.pData[1].Lit ) ), w2 );
    
    } else {
        xSAT_WatchListPush( xSAT_VecWatchListEntry( s.vWatches, xSAT_NegLit( pCla.pData[0].Lit ) ), w1 );
        xSAT_WatchListPush( xSAT_VecWatchListEntry( s.vWatches, xSAT_NegLit( pCla.pData[1].Lit ) ), w2 );
    }
    return CRef;
}

/**Function*************************************************************

  Synopsis    []

  Description []

  SideEffects []

  SeeAlso     []

***********************************************************************/
func xSAT_SolverEnqueue(s *xSAT_Solver_t, Lit int32, Reason uint32) int32 {
    var Var int32  = xSAT_Lit2Var( Lit );

    Vec_StrWriteEntry( s.vAssigns, Var, byte(xSAT_LitSign( Lit )) );
    Vec_IntWriteEntry( s.vLevels, Var, xSAT_SolverDecisionLevel( s ) );
    Vec_IntWriteEntry( s.vReasons, Var, int32(   Reason) );
    Vec_IntPush( s.vTrail, Lit );

    return true;
}

/**Function*************************************************************

  Synopsis    []

  Description []

  SideEffects []

  SeeAlso     []

***********************************************************************/
func xSAT_SolverNewDecision(s *xSAT_Solver_t, Lit int32) {
    assert(Vec_StrEntry( s.vAssigns, xSAT_Lit2Var( Lit ) ) == VarX, "Vec_StrEntry( s.vAssigns, xSAT_Lit2Var( Lit ) ) == VarX");
    s.Stats.nDecisions++;
    Vec_IntPush( s.vTrailLim, Vec_IntSize( s.vTrail ) );
    xSAT_SolverEnqueue( s, Lit, CRefUndef );
}

/**Function*************************************************************

  Synopsis    []

  Description []

  SideEffects []

  SeeAlso     []

***********************************************************************/
func xSAT_SolverCancelUntil(s *xSAT_Solver_t, Level int32) {
    var c int32;

    if xSAT_SolverDecisionLevel( s ) <= Level {
        return; }

    for  c = Vec_IntSize( s.vTrail ) - 1; c >= Vec_IntEntry( s.vTrailLim, Level ); c--  {
        var Var int32  = xSAT_Lit2Var( Vec_IntEntry( s.vTrail, c ) );

        Vec_StrWriteEntry( s.vAssigns, Var, VarX );
        Vec_IntWriteEntry( s.vReasons, Var, int32(   CRefUndef) );
        Vec_StrWriteEntry( s.vPolarity, Var, byte(  xSAT_LitSign( Vec_IntEntry( s.vTrail, c ) )) );

        if !xSAT_HeapInHeap( s.hOrder, Var ) {
            xSAT_HeapInsert( s.hOrder, Var ); }
    }

    s.iQhead = Vec_IntEntry( s.vTrailLim, Level );
    Vec_IntShrink( s.vTrail, Vec_IntEntry( s.vTrailLim, Level ) );
    Vec_IntShrink( s.vTrailLim, Level );
}

/**Function*************************************************************

  Synopsis    []

  Description []

  SideEffects []

  SeeAlso     []

***********************************************************************/
func xSAT_SolverIsLitRemovable(s *xSAT_Solver_t, Lit int32, MinLevel int32) int32 {
    var top int32  = Vec_IntSize( s.vTagged );

    assert(uint32(   Vec_IntEntry( s.vReasons, xSAT_Lit2Var( Lit ) )) != CRefUndef, "uint32(   Vec_IntEntry( s.vReasons, xSAT_Lit2Var( Lit ) )) != CRefUndef");
    Vec_IntClear( s.vStack );
    Vec_IntPush( s.vStack, xSAT_Lit2Var( Lit ) );

    for  Vec_IntSize( s.vStack )  {
        var i int32;
        var v int32  = Vec_IntPop(  s.vStack );
        var c *xSAT_Clause_t  = xSAT_SolverReadClause(s, uint32(   Vec_IntEntry( s.vReasons, v )) );
        var Lits *int32  = &( c.pData[0].Lit );

        assert(uint32( Vec_IntEntry( s.vReasons, v )) != CRefUndef, "uint32( Vec_IntEntry( s.vReasons, v )) != CRefUndef");
        if c.nSize == 2 && Vec_StrEntry( s.vAssigns, xSAT_Lit2Var( Lits[0] ) ) == xSAT_LitSign( xSAT_NegLit( Lits[0] ) ) {
            assert(Vec_StrEntry( s.vAssigns, xSAT_Lit2Var( Lits[1] ) ) == xSAT_LitSign( ( Lits[1] ) ), "Vec_StrEntry( s.vAssigns, xSAT_Lit2Var( Lits[1] ) ) == xSAT_LitSign( ( Lits[1] ) )");
            // inlining of macro ABC_SWAP
            { var t int32  = Lits[0]; Lits[0] = Lits[1]; Lits[1] = t; };
        }

        for  i = 1; i < c.nSize; i++  {
            var v int32  = xSAT_Lit2Var( Lits[i] );
            if !Vec_StrEntry( s.vSeen, v ) && Vec_IntEntry( s.vLevels, v ) {
                if uint32(   Vec_IntEntry( s.vReasons, v )) != CRefUndef && ( ( 1 << (Vec_IntEntry( s.vLevels, v ) & 31 ) ) & MinLevel ) {
                    Vec_IntPush( s.vStack, v );
                    Vec_IntPush( s.vTagged, Lits[i] );
                    Vec_StrWriteEntry( s.vSeen, v, 1 );
                
                } else {
                    var Lit int32;
                    // inlining of macro Vec_IntForEachEntryStart
                    for  i = top; (i < Vec_IntSize(s.vTagged)) && (((Lit) = Vec_IntEntry(s.vTagged, i)), 1); i++  {
                        Vec_StrWriteEntry( s.vSeen, xSAT_Lit2Var( Lit ), 0 );
                    }
                    Vec_IntShrink( s.vTagged, top );
                    return 0;
                }
            }
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
func xSAT_SolverClaMinimisation(s *xSAT_Solver_t, vLits *Vec_Int_t) {
    var pLits *int32  = Vec_IntArray( vLits );
    var MinLevel int32;
    var i, j int32;

    for  i = 1; i < Vec_IntSize( vLits ); i++  {
        var Level int32  = Vec_IntEntry( s.vLevels, xSAT_Lit2Var( pLits[i] ) );
        MinLevel |= 1 << ( Level & 31 );
    }

    /* Remove reduntant literals */
    Vec_IntAppend( s.vTagged, vLits );
    for  i = j = 1; i < Vec_IntSize( vLits ); i++  {
        if uint32(   Vec_IntEntry( s.vReasons, xSAT_Lit2Var( pLits[i] ) )) == CRefUndef || !(xSAT_SolverIsLitRemovable( s, pLits[i], MinLevel ) != 0) {
            pLits[j++] = pLits[i]; }
    }
    Vec_IntShrink( vLits, j );

    /* Binary Resolution */
    if Vec_IntSize( vLits ) <= 30 && xSAT_SolverClaCalcLBD2( s, vLits ) <= 6 {
        var nb, l int32;
        var Lit int32;
        var FlaseLit int32  = xSAT_NegLit( pLits[0] );
        var ws *xSAT_WatchList_t  = xSAT_VecWatchListEntry( s.vBinWatches, FlaseLit );
        var begin *xSAT_Watcher_t  = xSAT_WatchListArray( ws );
        var end *xSAT_Watcher_t    = begin + xSAT_WatchListSize( ws );
        var pWatcher *xSAT_Watcher_t;

        s.nStamp++;
        // inlining of macro Vec_IntForEachEntry
        for  i = 0; (i < Vec_IntSize(vLits)) && (((Lit) = Vec_IntEntry(vLits, i)), 1); i++  {
            Vec_IntWriteEntry( s.vStamp, xSAT_Lit2Var( Lit ), int32(   s.nStamp) );
        }

        nb = 0;
        for  pWatcher = begin; pWatcher < end; pWatcher++  {
            var ImpLit int32  = pWatcher.Blocker;

            if uint32(   Vec_IntEntry( s.vStamp, xSAT_Lit2Var( ImpLit ) )) == s.nStamp && Vec_StrEntry( s.vAssigns, xSAT_Lit2Var( ImpLit ) ) == xSAT_LitSign( ImpLit ) {
                nb++;
                Vec_IntWriteEntry( s.vStamp, xSAT_Lit2Var( ImpLit ), int32(  ( s.nStamp - 1 )) );
            }
        }

        l = Vec_IntSize( vLits ) - 1;
        if nb > 0 {
            for  i = 1; i < Vec_IntSize( vLits ) - nb; i++  {
                if uint32(   Vec_IntEntry( s.vStamp, xSAT_Lit2Var( pLits[i] ) )) != s.nStamp {
                    var TempLit int32  = pLits[l];
                    pLits[l] = pLits[i];
                    pLits[i] = TempLit;
                    i--; l--;
                }
            }

            Vec_IntShrink( vLits, Vec_IntSize( vLits ) - nb );
        }
    }
}

/**Function*************************************************************

  Synopsis    []

  Description []

  SideEffects []

  SeeAlso     []

***********************************************************************/
func xSAT_SolverAnalyze(s *xSAT_Solver_t, ConfCRef uint32, vLearnt *Vec_Int_t, OutBtLevel *int32, nLBD *uint32) {
    var trail *int32  = Vec_IntArray( s.vTrail );
    var Count int32;
    var p int32        = LitUndef;
    var Idx int32      = Vec_IntSize( s.vTrail ) - 1;
    var Lits *int32;
    var Lit int32;
    var i, j int32;

    Vec_IntPush( vLearnt, LitUndef );
    for
    {
        var pCla *xSAT_Clause_t;

        assert(ConfCRef != CRefUndef, "ConfCRef != CRefUndef");
        pCla = xSAT_SolverReadClause(s, ConfCRef);
        Lits = &( pCla.pData[0].Lit );

        if p != LitUndef && pCla.nSize == 2 && Vec_StrEntry( s.vAssigns, xSAT_Lit2Var( Lits[0] ) ) == xSAT_LitSign( xSAT_NegLit( Lits[0] ) ) {
            assert(Vec_StrEntry( s.vAssigns, xSAT_Lit2Var( Lits[1] ) ) == xSAT_LitSign( ( Lits[1] ) ), "Vec_StrEntry( s.vAssigns, xSAT_Lit2Var( Lits[1] ) ) == xSAT_LitSign( ( Lits[1] ) )");
            // inlining of macro ABC_SWAP
            { var t int32  = Lits[0]; Lits[0] = Lits[1]; Lits[1] = t; };
        }

        if pCla.fLearnt {
            xSAT_SolverClaActBump( s, pCla ); }

        if pCla.fLearnt && pCla.nLBD > 2 {
            unsigned int nLevels = xSAT_SolverClaCalcLBD( s, pCla );
            if nLevels + 1 < pCla.nLBD {
                if pCla.nLBD <= s.Config.nLBDFrozenClause {
                    pCla.fCanBeDel = 0; }
                pCla.nLBD = nLevels;
            }
        }

        for  j = ( p == LitUndef ? 0 : 1 ); j < pCla.nSize; j++  {
            var Var int32  = xSAT_Lit2Var( Lits[j] );

            if Vec_StrEntry( s.vSeen, Var ) == 0 && Vec_IntEntry( s.vLevels, Var ) > 0 {
                Vec_StrWriteEntry( s.vSeen, Var, 1 );
                xSAT_SolverVarActBump( s, Var );
                if Vec_IntEntry( s.vLevels, Var ) >= xSAT_SolverDecisionLevel( s ) {
                    Count++;
                    if Vec_IntEntry( s.vReasons, Var ) != CRefUndef && xSAT_SolverReadClause( s, Vec_IntEntry( s.vReasons, Var ) ).fLearnt {
                        Vec_IntPush( s.vLastDLevel, Var ); }
                
                } else {
                    Vec_IntPush( vLearnt, Lits[j] ); }
            }
        }

        for  !Vec_StrEntry( s.vSeen, xSAT_Lit2Var( trail[Idx--] ) )  {; }

        // Next clause to look at
        p = trail[Idx+1];
        ConfCRef = uint32(   Vec_IntEntry( s.vReasons, xSAT_Lit2Var( p ) ));
        Vec_StrWriteEntry( s.vSeen, xSAT_Lit2Var( p ), 0 );
        Count--;

    
if !( Count > 0 ) { break }
} 

    Vec_IntArray( vLearnt )[0] = xSAT_NegLit( p );
    xSAT_SolverClaMinimisation( s, vLearnt );

    // Find the backtrack level
    Lits = Vec_IntArray( vLearnt );
    if Vec_IntSize( vLearnt ) == 1 {
        *OutBtLevel = 0;
    } else {
        var iMax int32  = 1;
        var Max int32    = Vec_IntEntry( s.vLevels, xSAT_Lit2Var( Lits[1] ) );
        var Tmp int32;

        for i = 2; i < Vec_IntSize( vLearnt ); i++ {
            if Vec_IntEntry( s.vLevels, xSAT_Lit2Var( Lits[i]) ) > Max {
                Max   = Vec_IntEntry( s.vLevels, xSAT_Lit2Var( Lits[i]) );
                iMax = i;
            }
        }

        Tmp         = Lits[1];
        Lits[1]     = Lits[iMax];
        Lits[iMax]  = Tmp;
        *OutBtLevel = Vec_IntEntry( s.vLevels, xSAT_Lit2Var( Lits[1] ) );
    }

    *nLBD = xSAT_SolverClaCalcLBD2( s, vLearnt );
    if Vec_IntSize( s.vLastDLevel ) > 0 {
        var Var int32;
        // inlining of macro Vec_IntForEachEntry
        for  i = 0; (i < Vec_IntSize(s.vLastDLevel)) && (((Var) = Vec_IntEntry(s.vLastDLevel, i)), 1); i++  {
            if xSAT_SolverReadClause( s, Vec_IntEntry( s.vReasons, Var ) ).nLBD < *nLBD {
                xSAT_SolverVarActBump( s, Var ); }
        }

        Vec_IntClear( s.vLastDLevel );
    }

    // inlining of macro Vec_IntForEachEntry
    for  i = 0; (i < Vec_IntSize(s.vTagged)) && (((Lit) = Vec_IntEntry(s.vTagged, i)), 1); i++  {
        Vec_StrWriteEntry( s.vSeen, xSAT_Lit2Var( Lit ), 0 );
    }
    Vec_IntClear( s.vTagged );
}

/**Function*************************************************************

  Synopsis    []

  Description []

  SideEffects []

  SeeAlso     []

***********************************************************************/
func xSAT_SolverPropagate(s *xSAT_Solver_t) uint32 {
    var hConfl uint32  = CRefUndef;
    var Lits *int32;
    var NegLit int32;
    var nProp int32;

    for  s.iQhead < Vec_IntSize( s.vTrail )  {
        var p int32  = Vec_IntEntry( s.vTrail, s.iQhead++ );
        var ws *xSAT_WatchList_t  = xSAT_VecWatchListEntry( s.vBinWatches, p );
        var begin *xSAT_Watcher_t  = xSAT_WatchListArray( ws );
        var end *xSAT_Watcher_t    = begin + xSAT_WatchListSize( ws );
        var i, j *xSAT_Watcher_t;

        nProp++;
        for  i = begin; i < end; i++  {
            if Vec_StrEntry( s.vAssigns, xSAT_Lit2Var( i.Blocker ) ) == VarX {
                xSAT_SolverEnqueue( s, i.Blocker, i.CRef );
            } else if Vec_StrEntry( s.vAssigns, xSAT_Lit2Var( i.Blocker ) ) == xSAT_LitSign( xSAT_NegLit( i.Blocker ) ) {
                return i.CRef; }
        }

        ws    = xSAT_VecWatchListEntry( s.vWatches, p );
        begin = xSAT_WatchListArray( ws );
        end   = begin + xSAT_WatchListSize( ws );

        for  i = j = begin; i < end;  {
            var pCla *xSAT_Clause_t;
            var w xSAT_Watcher_t;
            if Vec_StrEntry( s.vAssigns, xSAT_Lit2Var( i.Blocker ) ) == xSAT_LitSign( i.Blocker ) {
                *j++ = *i++;
                continue;
            }

            pCla = xSAT_SolverReadClause( s, i.CRef );
            Lits = &( pCla.pData[0].Lit );

            // Make sure the false literal is data[1]:
            NegLit = xSAT_NegLit( p );
            if Lits[0] == NegLit {
                Lits[0] = Lits[1];
                Lits[1] = NegLit;
            }
            assert(Lits[1] == NegLit, "Lits[1] == NegLit");

            w.CRef = i.CRef;
            w.Blocker = Lits[0];

            // If 0th watch is true, then clause is already satisfied.
            if Lits[0] != i.Blocker && Vec_StrEntry( s.vAssigns, xSAT_Lit2Var( Lits[0] ) ) == xSAT_LitSign( Lits[0] ) {
                *j++ = w;
            } else {
                // Look for new watch:
                var stop *int32  = Lits + pCla.nSize;
                var k *int32;
                for  k = Lits + 2; k < stop; k++  {
                    if Vec_StrEntry( s.vAssigns, xSAT_Lit2Var( *k ) ) != !xSAT_LitSign( *k ) {
                        Lits[1] = *k;
                        *k = NegLit;
                        xSAT_WatchListPush( xSAT_VecWatchListEntry( s.vWatches, xSAT_NegLit( Lits[1] ) ), w );
                        goto next;
                    }
                }

                *j++ = w;

                // Clause is unit under assignment:
                if Vec_StrEntry( s.vAssigns, xSAT_Lit2Var( Lits[0] ) ) == xSAT_LitSign( xSAT_NegLit( Lits[0] ) ) {
                    hConfl = i.CRef;
                    i++;
                    s.iQhead = Vec_IntSize( s.vTrail );
                    // Copy the remaining watches:
                    for i < end {
                        *j++ = *i++;
                    }
                
                } else {
                    xSAT_SolverEnqueue( s, Lits[0], i.CRef ); }
            }
        next:
            i++;
        }

        s.Stats.nInspects += j - xSAT_WatchListArray( ws );
        xSAT_WatchListShrink( ws, j - xSAT_WatchListArray( ws ) );
    }

    s.Stats.nPropagations += nProp;
    s.nPropSimplify -= nProp;

    return hConfl;
}

/**Function*************************************************************

  Synopsis    []

  Description []

  SideEffects []

  SeeAlso     []

***********************************************************************/
var spStatic_xSAT_SolverReduceDB_TimeTotal abctime;
func xSAT_SolverReduceDB(s *xSAT_Solver_t) {
    
    var clk abctime  = Abc_Clock();
    var nLearnedOld int32  = Vec_IntSize( s.vLearnts );
    var i, limit int32;
    var CRef uint32;
    var pCla *xSAT_Clause_t;
    var learnts_cls **xSAT_Clause_t;

    // inlining of macro ABC_ALLOC
    learnts_cls = ((xSAT_Clause_t * *) malloc(sizeof(xSAT_Clause_t *) * (size_t)(nLearnedOld)));
    // inlining of macro Vec_IntForEachEntry
    for  i = 0; (i < Vec_IntSize(s.vLearnts)) && (((CRef) = Vec_IntEntry(s.vLearnts, i)), 1); i++  {
        learnts_cls[i] = xSAT_SolverReadClause(s, CRef);
    }

    limit = nLearnedOld / 2;

    xSAT_UtilSort((void **) learnts_cls, nLearnedOld,
                  (int (*)( void *, void * )) xSAT_ClauseCompare);

    if learnts_cls[nLearnedOld / 2].nLBD <= 3 {
        s.nRC2 += s.Config.nSpecialIncReduce; }
    if learnts_cls[nLearnedOld - 1].nLBD <= 5 {
        s.nRC2 += s.Config.nSpecialIncReduce; }

    Vec_IntClear( s.vLearnts );
    for  i = 0; i < nLearnedOld; i++  {
        var CRef uint32;

        pCla = learnts_cls[i];
        CRef = xSAT_MemCRef( s.pMemory, ( unsigned * ) pCla );
        assert(pCla.fMark == 0, "pCla.fMark == 0");
        if pCla.fCanBeDel && pCla.nLBD > 2 && pCla.nSize > 2 && uint32( Vec_IntEntry( s.vReasons, xSAT_Lit2Var( pCla.pData[0].Lit ) )) != CRef && ( i < limit ) {
            pCla.fMark = 1;
            s.Stats.nLearntLits -= pCla.nSize;
            xSAT_WatchListRemove( xSAT_VecWatchListEntry( s.vWatches, xSAT_NegLit( pCla.pData[0].Lit ) ), CRef );
            xSAT_WatchListRemove( xSAT_VecWatchListEntry( s.vWatches, xSAT_NegLit( pCla.pData[1].Lit ) ), CRef );
        
        } else {
            if !pCla.fCanBeDel {
                limit++; }
            pCla.fCanBeDel = 1;
            Vec_IntPush( s.vLearnts, CRef );
        }
    }
    learnts_cls = nil;

    spStatic_xSAT_SolverReduceDB_TimeTotal += Abc_Clock() - clk;
    if s.Config.fVerbose {
        Abc_Print(1, "reduceDB: Keeping %7d out of %7d clauses (%5.2f %%)  ",
            Vec_IntSize( s.vLearnts ), nLearnedOld, 100.0 * Vec_IntSize( s.vLearnts ) / nLearnedOld );
        Abc_PrintTime( 1, "Time", spStatic_xSAT_SolverReduceDB_TimeTotal );
    }
    xSAT_SolverGarbageCollect(s);
}

/**Function*************************************************************

  Synopsis    []

  Description []

  SideEffects []

  SeeAlso     []

***********************************************************************/
func xSAT_SolverSearch(s *xSAT_Solver_t) byte {
    var conflictC iword;

    s.Stats.nStarts++;
    for {
        var hConfl uint32  = xSAT_SolverPropagate( s );

        if hConfl != CRefUndef {
            /* Conflict */
            var BacktrackLevel int32;
            var nLBD uint32;
            var CRef uint32;

            s.Stats.nConflicts++;
            conflictC++;

            if xSAT_SolverDecisionLevel( s ) == 0 {
                return LBoolFalse; }

            xSAT_BQueuePush( s.bqTrail, Vec_IntSize( s.vTrail ) );
            if s.Stats.nConflicts > s.Config.nFirstBlockRestart && xSAT_BQueueIsValid( s.bqLBD ) && ( Vec_IntSize( s.vTrail ) > ( s.Config.R * ( iword ) xSAT_BQueueAvg( s.bqTrail ) ) ) {
                xSAT_BQueueClean(s.bqLBD); }

            Vec_IntClear( s.vLearntClause );
            xSAT_SolverAnalyze( s, hConfl, s.vLearntClause, &BacktrackLevel, &nLBD );

            s.nSumLBD += nLBD;
            xSAT_BQueuePush( s.bqLBD, nLBD );
            xSAT_SolverCancelUntil( s, BacktrackLevel );

            if Vec_IntSize( s.vLearntClause ) == 1 {
                CRef = CRefUndef;
            } else {
                CRef = xSAT_SolverClaNew( s, s.vLearntClause , 1 );
            }
            xSAT_SolverEnqueue( s, Vec_IntEntry( s.vLearntClause , 0 ), CRef );

            xSAT_SolverVarActDecay( s );
            xSAT_SolverClaActDecay( s );
        
        } else {
            /* No conflict */
            var NextVar int32;
            if xSAT_BQueueIsValid( s.bqLBD ) && ( ( ( iword )xSAT_BQueueAvg( s.bqLBD ) * s.Config.K ) > ( s.nSumLBD / s.Stats.nConflicts ) ) {
                xSAT_BQueueClean( s.bqLBD );
                xSAT_SolverCancelUntil( s, 0 );
                return LBoolUndef;
            }

            // Simplify the set of problem clauses:
            if xSAT_SolverDecisionLevel( s ) == 0 {
                xSAT_SolverSimplify( s ); }

            // Reduce the set of learnt clauses:
            if s.Stats.nConflicts >= s.nConfBeforeReduce {
                s.nRC1 = ( s.Stats.nConflicts / s.nRC2 ) + 1;
                xSAT_SolverReduceDB(s);
                s.nRC2 += s.Config.nIncReduce;
                s.nConfBeforeReduce = s.nRC1 * s.nRC2;
            }

            // New variable decision:
            NextVar = xSAT_SolverDecide( s );

            if NextVar == VarUndef {
                return LBoolTrue; }

            xSAT_SolverNewDecision( s, xSAT_Var2Lit( NextVar, int32(   Vec_StrEntry( s.vPolarity, NextVar )) ) );
        }
    }

    return LBoolUndef; // cannot happen
}

/**Function*************************************************************

  Synopsis    []

  Description []

  SideEffects []

  SeeAlso     []

***********************************************************************/
func xSAT_SolverClaRealloc(pDest *xSAT_Mem_t, pSrc *xSAT_Mem_t, pCRef *uint32) {
    var nNewCRef uint32;
    var pNewCla *xSAT_Clause_t;
    var pOldCla *xSAT_Clause_t  = xSAT_MemClauseHand( pSrc, *pCRef );

    if pOldCla.fReallocd {
        *pCRef = uint32(   pOldCla.nSize);
        return;
    }
    nNewCRef = xSAT_MemAppend( pDest, 3 + pOldCla.fLearnt + pOldCla.nSize );
    pNewCla = xSAT_MemClauseHand( pDest, nNewCRef );
    memcpy( pNewCla, pOldCla, (size_t)(( 3 + pOldCla.fLearnt + pOldCla.nSize ) * 4) );
    pOldCla.fReallocd = 1;
    pOldCla.nSize = uint32(   nNewCRef);
    *pCRef = nNewCRef;
}

/**Function*************************************************************

  Synopsis    []

  Description []

  SideEffects []

  SeeAlso     []

***********************************************************************/
func xSAT_SolverGarbageCollect(s *xSAT_Solver_t) {
    var i int32;
    var pArray *uint32;
    var pNewMemMngr *xSAT_Mem_t  = xSAT_MemAlloc(  xSAT_MemCap( s.pMemory ) - xSAT_MemWastedCap( s.pMemory ) );

    for  i = 0; i < 2 * Vec_StrSize( s.vAssigns ); i++  {
        var ws *xSAT_WatchList_t  = xSAT_VecWatchListEntry( s.vWatches, i);
        var begin *xSAT_Watcher_t  = xSAT_WatchListArray(ws);
        var end *xSAT_Watcher_t    = begin + xSAT_WatchListSize(ws);
        var w *xSAT_Watcher_t;

        for  w = begin; w != end; w++  {
            xSAT_SolverClaRealloc( pNewMemMngr, s.pMemory, &(w.CRef) );
        }

        ws = xSAT_VecWatchListEntry( s.vBinWatches, i);
        begin = xSAT_WatchListArray(ws);
        end   = begin + xSAT_WatchListSize(ws);
        for  w = begin; w != end; w++  {
            xSAT_SolverClaRealloc( pNewMemMngr, s.pMemory, &(w.CRef) );
        }
    }

    for  i = 0; i < Vec_IntSize( s.vTrail ); i++  {
        if uint32(   Vec_IntEntry( s.vReasons, xSAT_Lit2Var( Vec_IntEntry( s.vTrail, i ) ) )) != CRefUndef {
            xSAT_SolverClaRealloc( pNewMemMngr, s.pMemory, ( unsigned * ) &( Vec_IntArray( s.vReasons )[xSAT_Lit2Var( Vec_IntEntry( s.vTrail, i ) )] ) ); }
    }

    pArray = ( unsigned * ) Vec_IntArray( s.vLearnts );
    for  i = 0; i < Vec_IntSize( s.vLearnts ); i++  {
        xSAT_SolverClaRealloc( pNewMemMngr, s.pMemory, &(pArray[i]) );
    }

    pArray = ( unsigned * ) Vec_IntArray( s.vClauses );
    for  i = 0; i < Vec_IntSize( s.vClauses ); i++  {
        xSAT_SolverClaRealloc( pNewMemMngr, s.pMemory, &(pArray[i]) );
    }

    xSAT_MemFree( s.pMemory );
    s.pMemory = pNewMemMngr;
}

//ABC_NAMESPACE_IMPL_END
