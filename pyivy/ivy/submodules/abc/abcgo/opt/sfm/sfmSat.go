package sfm

import "fmt"

/**CFile****************************************************************

  FileName    [sfmSat.c]

  SystemName  [ABC: Logic synthesis and verification system.]

  PackageName [SAT-based optimization using internal don't-cares.]

  Synopsis    [SAT-based procedures.]

  Author      [Alan Mishchenko]
  
  Affiliation [UC Berkeley]

  Date        [Ver. 1.0. Started - June 20, 2005.]

  Revision    [$Id: sfmSat.c,v 1.00 2005/06/20 00:00:00 alanmi Exp $]

***********************************************************************/

//#include "sfmInt.h"
//#include "misc/util/utilTruth.h"

//ABC_NAMESPACE_IMPL_START


////////////////////////////////////////////////////////////////////////
///                        DECLARATIONS                              ///
////////////////////////////////////////////////////////////////////////

////////////////////////////////////////////////////////////////////////
///                     FUNCTION DEFINITIONS                         ///
////////////////////////////////////////////////////////////////////////

/**Function*************************************************************

  Synopsis    [Converts a window into a SAT solver.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Sfm_NtkWindowToSolver(p *Sfm_Ntk_t) int32 {
    // p->vOrder contains all variables in the window in a good order
    // p->vDivs is a subset of nodes in p->vOrder used as divisor candidates
    // p->vTfo contains TFO of the node (does not include node)
    // p->vRoots contains roots of the TFO of the node (may include node)
    var vClause *Vec_Int_t;
    var RetValue int32;
    var iNode int32 = -1;
    var iFanin int32;
    var i int32;
    var k int32;
    var clk abctime  = Abc_Clock();
//    if ( p->pSat )
//        printf( "%d  ", p->pSat->stats.learnts );
    sat_solver_restart( p.pSat );
    sat_solver_setnvars( p.pSat, 1 + Vec_IntSize(p.vOrder) + Vec_IntSize(p.vTfo) + Vec_IntSize(p.vRoots) + 10 );
    // create SAT variables
    Sfm_NtkCleanVars( p );
    p.nSatVars = 1;
    // inlining of macro Vec_IntForEachEntry
    for  i = 0; (i < Vec_IntSize(p.vOrder)) && (((iNode) = Vec_IntEntry(p.vOrder, i)), 1); i++  {
        Sfm_ObjSetSatVar( p, iNode, p.nSatVars++ );
    }
    // collect divisor variables
    Vec_IntClear( p.vDivVars );
    // inlining of macro Vec_IntForEachEntry
    for  i = 0; (i < Vec_IntSize(p.vDivs)) && (((iNode) = Vec_IntEntry(p.vDivs, i)), 1); i++  {
        Vec_IntPush( p.vDivVars, Sfm_ObjSatVar(p, iNode) );
    }
    // add CNF clauses for the TFI
    // inlining of macro Vec_IntForEachEntry
    for  i = 0; (i < Vec_IntSize(p.vOrder)) && (((iNode) = Vec_IntEntry(p.vOrder, i)), 1); i++  {
        if Sfm_ObjIsPi(p, iNode) {
            continue; }
        // collect fanin variables
        Vec_IntClear( p.vFaninMap );
        // inlining of macro Sfm_ObjForEachFanin
        for  k = 0; k < Sfm_ObjFaninNum(p, iNode)  && ((iFanin = Sfm_ObjFanin(p, iNode, k)), 1);  k++  {
            Vec_IntPush( p.vFaninMap, Sfm_ObjSatVar(p, iFanin) );
        }
        Vec_IntPush( p.vFaninMap, Sfm_ObjSatVar(p, iNode) );
        // generate CNF 
        Sfm_TranslateCnf( p.vClauses, (Vec_Str_t *)Vec_WecEntry(p.vCnfs, iNode), p.vFaninMap, -1 );
        // add clauses
        // inlining of macro Vec_WecForEachLevel
        for  k = 0; (k < Vec_WecSize(p.vClauses)) && (((vClause) = Vec_WecEntry(p.vClauses, k)), 1); k++  {
            if Vec_IntSize(vClause) == 0 {
                break; }
            RetValue = sat_solver_addclause( p.pSat, Vec_IntArray(vClause), Vec_IntArray(vClause) + Vec_IntSize(vClause) );
            if RetValue == 0 {
                return 0; }
        }
    }
    if Vec_IntSize(p.vTfo) > 0 {
        assert(p.pPars.nTfoLevMax > 0, "p.pPars.nTfoLevMax > 0");
        assert(Vec_IntSize(p.vRoots) > 0, "Vec_IntSize(p.vRoots) > 0");
        assert(Vec_IntEntry(p.vTfo, 0) != p.iPivotNode, "Vec_IntEntry(p.vTfo, 0) != p.iPivotNode");
        // collect variables of root nodes
        Vec_IntClear( p.vLits );
        // inlining of macro Vec_IntForEachEntry
        for  i = 0; (i < Vec_IntSize(p.vRoots)) && (((iNode) = Vec_IntEntry(p.vRoots, i)), 1); i++  {
            Vec_IntPush( p.vLits, Sfm_ObjSatVar(p, iNode) );
        }
        // assign new variables to the TFO nodes
        // inlining of macro Vec_IntForEachEntry
        for  i = 0; (i < Vec_IntSize(p.vTfo)) && (((iNode) = Vec_IntEntry(p.vTfo, i)), 1); i++  {
            Sfm_ObjCleanSatVar( p, Sfm_ObjSatVar(p, iNode) );
            Sfm_ObjSetSatVar( p, iNode, p.nSatVars++ );
        }
        // add CNF clauses for the TFO
        // inlining of macro Vec_IntForEachEntry
        for  i = 0; (i < Vec_IntSize(p.vTfo)) && (((iNode) = Vec_IntEntry(p.vTfo, i)), 1); i++  {
            assert(Sfm_ObjIsNode(p, iNode), "Sfm_ObjIsNode(p, iNode)");
            // collect fanin variables
            Vec_IntClear( p.vFaninMap );
            // inlining of macro Sfm_ObjForEachFanin
            for  k = 0; k < Sfm_ObjFaninNum(p, iNode)  && ((iFanin = Sfm_ObjFanin(p, iNode, k)), 1);  k++  {
                Vec_IntPush( p.vFaninMap, Sfm_ObjSatVar(p, iFanin) );
            }
            Vec_IntPush( p.vFaninMap, Sfm_ObjSatVar(p, iNode) );
            // generate CNF 
            Sfm_TranslateCnf( p.vClauses, (Vec_Str_t *)Vec_WecEntry(p.vCnfs, iNode), p.vFaninMap, Sfm_ObjSatVar(p, p.iPivotNode) );
            // add clauses
            // inlining of macro Vec_WecForEachLevel
            for  k = 0; (k < Vec_WecSize(p.vClauses)) && (((vClause) = Vec_WecEntry(p.vClauses, k)), 1); k++  {
                if Vec_IntSize(vClause) == 0 {
                    break; }
                RetValue = sat_solver_addclause( p.pSat, Vec_IntArray(vClause), Vec_IntArray(vClause) + Vec_IntSize(vClause) );
                if RetValue == 0 {
                    return 0; }
            }
        }
        // create XOR clauses for the roots
        // inlining of macro Vec_IntForEachEntry
        for  i = 0; (i < Vec_IntSize(p.vRoots)) && (((iNode) = Vec_IntEntry(p.vRoots, i)), 1); i++  {
            sat_solver_add_xor( p.pSat, Vec_IntEntry(p.vLits, i), Sfm_ObjSatVar(p, iNode), p.nSatVars++, 0 );
            Vec_IntWriteEntry( p.vLits, i, Abc_Var2Lit(p.nSatVars-1, 0) );
        }
        // make OR clause for the last nRoots variables
        RetValue = sat_solver_addclause( p.pSat, Vec_IntArray(p.vLits), Vec_IntArray(p.vLits) + Vec_IntSize(p.vLits) );
        if RetValue == 0 {
            return 0; }
    }
    // finalize
    RetValue = sat_solver_simplify( p.pSat );
    p.timeCnf += Abc_Clock() - clk;
    return RetValue;
} 

/**Function*************************************************************

  Synopsis    [Takes SAT solver and returns interpolant.]

  Description [If interpolant does not exist, records difference variables.]
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Sfm_ComputeInterpolant(p *Sfm_Ntk_t) word {
    var pSign *word;
    var uCube word;
    var uTruth word;
    var status int32;
    var i int32;
    var Div int32;
    var iVar int32;
    var nFinal int32;
    var pFinal *int32;
    var nIter int32;
    var pLits [2]int32;
    var nVars int32 = sat_solver_nvars( p.pSat );
    sat_solver_setnvars( p.pSat, nVars + 1 );
    pLits[0] = Abc_Var2Lit( Sfm_ObjSatVar(p, p.iPivotNode), 0 ); // F = 1
    pLits[1] = Abc_Var2Lit( nVars, 0 ); // iNewLit
    assert(Vec_IntSize(p.vDivIds) <= 6, "Vec_IntSize(p.vDivIds) <= 6");
    for {
        // find onset minterm
        p.nSatCalls++;
        status = sat_solver_solve( p.pSat, pLits, pLits + 2, p.pPars.nBTLimit, 0, 0, 0 );
        if status == l_Undef {
            return SFM_SAT_UNDEC; }
        if status == l_False {
            return uTruth; }
        assert(status == l_True, "status == l_True");
        // remember variable values
        Vec_IntClear( p.vValues );
        // inlining of macro Vec_IntForEachEntry
        for  i = 0; (i < Vec_IntSize(p.vDivVars)) && (((iVar) = Vec_IntEntry(p.vDivVars, i)), 1); i++  {
            Vec_IntPush( p.vValues, sat_solver_var_value(p.pSat, iVar) );
        }
        // collect divisor literals
        Vec_IntClear( p.vLits );
        Vec_IntPush( p.vLits, Abc_LitNot(pLits[0]) ); // F = 0
        // inlining of macro Vec_IntForEachEntry
        for  i = 0; (i < Vec_IntSize(p.vDivIds)) && (((Div) = Vec_IntEntry(p.vDivIds, i)), 1); i++  {
            Vec_IntPush( p.vLits, sat_solver_var_literal(p.pSat, Div) );
        }
        // check against offset
        p.nSatCalls++;
        status = sat_solver_solve( p.pSat, Vec_IntArray(p.vLits), Vec_IntArray(p.vLits) + Vec_IntSize(p.vLits), p.pPars.nBTLimit, 0, 0, 0 );
        if status == l_Undef {
            return SFM_SAT_UNDEC; }
        if status == l_True {
            break; }
        assert(status == l_False, "status == l_False");
        // compute cube and add clause
        nFinal = sat_solver_final( p.pSat, &pFinal );
        uCube = ~(word)0;
        Vec_IntClear( p.vLits );
        Vec_IntPush( p.vLits, Abc_LitNot(pLits[1]) ); // NOT(iNewLit)
        for  i = 0; i < nFinal; i++  {
            if pFinal[i] == pLits[0] {
                continue; }
            Vec_IntPush( p.vLits, pFinal[i] );
            iVar = Vec_IntFind( p.vDivIds, Abc_Lit2Var(pFinal[i]) );   assert(iVar >= 0, "iVar >= 0");
            uCube &= Abc_LitIsCompl(pFinal[i]) ? s_Truths6[iVar] : ~s_Truths6[iVar];
        }
        uTruth |= uCube;
        status = sat_solver_addclause( p.pSat, Vec_IntArray(p.vLits), Vec_IntArray(p.vLits) + Vec_IntSize(p.vLits) );
        assert((status != 0), "status");
        nIter++;
    }
    assert(status == l_True, "status == l_True");
    // store the counter-example
    // inlining of macro Vec_IntForEachEntry
    for  i = 0; (i < Vec_IntSize(p.vDivVars)) && (((iVar) = Vec_IntEntry(p.vDivVars, i)), 1); i++  {
        if Vec_IntEntry(p.vValues, i) ^ sat_solver_var_value(p.pSat, iVar) { // insert 1
        
            pSign = Vec_WrdEntryP( p.vDivCexes, i );
            assert(!Abc_InfoHasBit( (unsigned *)pSign, p.nCexes), "!Abc_InfoHasBit( (unsigned *)pSign, p.nCexes)");
            Abc_InfoXorBit( (unsigned *)pSign, p.nCexes );
        }
    }
    p.nCexes++;
    return SFM_SAT_SAT;
}

/**Function*************************************************************

  Synopsis    [Takes SAT solver and returns interpolant.]

  Description [If interpolant does not exist, records difference variables.]
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Sfm_ComputeInterpolantInt(p *Sfm_Ntk_t, Truth []word) int32 {
    var fOnSet int32;
    var iMint int32;
    var iVar int32;
    var nVars int32 = sat_solver_nvars( p.pSat );
    var iVarPivot int32  = Sfm_ObjSatVar( p, p.iPivotNode );
    var status, iNewLit, i, Div, nIter int32;
    Truth[0] = Truth[1] = 0;
    sat_solver_setnvars( p.pSat, nVars + 1 );
    iNewLit = Abc_Var2Lit( nVars, 0 ); // iNewLit
    assert(Vec_IntSize(p.vDivIds) <= 6, "Vec_IntSize(p.vDivIds) <= 6");
    Vec_IntFill( p.vValues, (1 << Vec_IntSize(p.vDivIds)) * Vec_IntSize(p.vDivVars), -1 );
    for {
        // find care minterm
        p.nSatCalls++; nIter++;
        status = sat_solver_solve( p.pSat, &iNewLit, &iNewLit + 1, p.pPars.nBTLimit, 0, 0, 0 );
        if status == l_Undef {
            return l_Undef; }
        if status == l_False {
            return l_False; }
        assert(status == l_True, "status == l_True");
        // collect values
        iMint = 0;
        fOnSet = sat_solver_var_value(p.pSat, iVarPivot);
        Vec_IntClear( p.vLits );
        Vec_IntPush( p.vLits, Abc_LitNot(iNewLit) ); // NOT(iNewLit)
        Vec_IntPush( p.vLits, Abc_LitNot(sat_solver_var_literal(p.pSat, iVarPivot)) );
        // inlining of macro Vec_IntForEachEntry
        for  i = 0; (i < Vec_IntSize(p.vDivIds)) && (((Div) = Vec_IntEntry(p.vDivIds, i)), 1); i++  {
            Vec_IntPush( p.vLits, Abc_LitNot(sat_solver_var_literal(p.pSat, Div)) );
            if sat_solver_var_value(p.pSat, Div) {
                iMint |= 1 << i; }
        }
        if Truth[!fOnSet] & ((word)1 << iMint) {
            break; } 
        assert(!(Truth[fOnSet] & ((word)1 << iMint)), "!(Truth[fOnSet] & ((word)1 << iMint))");
        Truth[fOnSet] |= ((word)1 << iMint);
        // remember variable values
        // inlining of macro Vec_IntForEachEntry
        for  i = 0; (i < Vec_IntSize(p.vDivVars)) && (((iVar) = Vec_IntEntry(p.vDivVars, i)), 1); i++  {
            Vec_IntWriteEntry( p.vValues, iMint * Vec_IntSize(p.vDivVars) + i, sat_solver_var_value(p.pSat, iVar) );
        }
        status = sat_solver_addclause( p.pSat, Vec_IntArray(p.vLits), Vec_IntArray(p.vLits) + Vec_IntSize(p.vLits) );
        if status == 0 {
            return l_False; }
    }
    assert(status == l_True, "status == l_True");
    // store the counter-example
    assert(iMint < (1 << Vec_IntSize(p.vDivIds)), "iMint < (1 << Vec_IntSize(p.vDivIds))");
    // inlining of macro Vec_IntForEachEntry
    for  i = 0; (i < Vec_IntSize(p.vDivVars)) && (((iVar) = Vec_IntEntry(p.vDivVars, i)), 1); i++  {
        var Value int32  = Vec_IntEntry(p.vValues, iMint * Vec_IntSize(p.vDivVars) + i);
        assert(Value != -1, "Value != -1");
        if Value ^ sat_solver_var_value(p.pSat, iVar) { // insert 1
        
            var pSign *word  = Vec_WrdEntryP( p.vDivCexes, i );
            assert(!Abc_InfoHasBit( (unsigned *)pSign, p.nCexes), "!Abc_InfoHasBit( (unsigned *)pSign, p.nCexes)");
            Abc_InfoXorBit( (unsigned *)pSign, p.nCexes );
        }
    }
    p.nCexes++;
    return l_True;
}
func Sfm_ComputeInterpolant2(p *Sfm_Ntk_t) word {
    var Res word;
    var ResP word;
    var ResN word;
    var Truth [2]word;
    var nCubesP, nCubesN int32;
    var RetValue int32  = Sfm_ComputeInterpolantInt( p, Truth );
    if RetValue == l_Undef {
        return SFM_SAT_UNDEC; }
    if RetValue == l_True {
        return SFM_SAT_SAT; }
    assert(RetValue == l_False, "RetValue == l_False");
    //printf( "Offset = %2d.  Onset = %2d.  DC = %2d.  Total = %2d.\n", 
    //    Abc_TtCountOnes(Truth[0]), Abc_TtCountOnes(Truth[1]), 
    //    (1<<Vec_IntSize(p->vDivIds)) - (Abc_TtCountOnes(Truth[0]) + Abc_TtCountOnes(Truth[1])), 
    //     1<<Vec_IntSize(p->vDivIds) );
    Truth[0] = Abc_Tt6Stretch( Truth[0], Vec_IntSize(p.vDivIds) );
    Truth[1] = Abc_Tt6Stretch( Truth[1], Vec_IntSize(p.vDivIds) );
    ResP = Abc_Tt6Isop( Truth[1], ~Truth[0], Vec_IntSize(p.vDivIds), &nCubesP );
    ResN = Abc_Tt6Isop( Truth[0], ~Truth[1], Vec_IntSize(p.vDivIds), &nCubesN );
    if nCubesP <= nCubesN {
        Res = ResP;
    } else {
        Res = ~ResN;
    }
    //Dau_DsdPrintFromTruth( &Res, Vec_IntSize(p->vDivIds) );            
    return Res;
}

/**Function*************************************************************

  Synopsis    [Checks resubstitution.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Sfm_ComputeInterpolantCheck(p *Sfm_Ntk_t) {
    var iNode int32  = 3;
    var iDiv0 int32  = 6;
    var iDiv1 int32  = 7;
    var uTruth word;
//    int i;
//    Sfm_NtkForEachNode( p, i )
    {
        Sfm_NtkCreateWindow( p, iNode, 1 );
        Sfm_NtkWindowToSolver( p );

        // collect SAT variables of divisors
        Vec_IntClear( p.vDivIds );
        Vec_IntPush( p.vDivIds, Sfm_ObjSatVar(p, iDiv0) );
        Vec_IntPush( p.vDivIds, Sfm_ObjSatVar(p, iDiv1) );

        uTruth = Sfm_ComputeInterpolant( p );

        if uTruth == SFM_SAT_SAT {
            fmt.Printf("The problem is SAT.\n");
        } else if uTruth == SFM_SAT_UNDEC {
            fmt.Printf("The problem is UNDEC.\n");
        } else {
            Kit_DsdPrintFromTruth( (unsigned *)&uTruth, 2 );
fmt.Printf("\n"); }
    }
}

////////////////////////////////////////////////////////////////////////
///                       END OF FILE                                ///
////////////////////////////////////////////////////////////////////////


//ABC_NAMESPACE_IMPL_END

