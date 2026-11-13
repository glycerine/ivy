package bmc

import "fmt"

/**CFile****************************************************************

  FileName    [bmcClp.c]

  SystemName  [ABC: Logic synthesis and verification system.]

  PackageName [SAT-based bounded model checking.]

  Synopsis    [INT-FX: Interpolation-based logic sharing extraction.]

  Author      [Alan Mishchenko]
  
  Affiliation [UC Berkeley]

  Date        [Ver. 1.0. Started - June 20, 2005.]

  Revision    [$Id: bmcClp.c,v 1.00 2005/06/20 00:00:00 alanmi Exp $]

***********************************************************************/

//#include "bmc.h"
//#include "misc/vec/vecWec.h"
//#include "sat/cnf/cnf.h"
//#include "sat/bsat/satStore.h"

//ABC_NAMESPACE_IMPL_START


////////////////////////////////////////////////////////////////////////
///                        DECLARATIONS                              ///
////////////////////////////////////////////////////////////////////////

////////////////////////////////////////////////////////////////////////
///                     FUNCTION DEFINITIONS                         ///
////////////////////////////////////////////////////////////////////////

var clkCheck1 abctime;
var clkCheck2 abctime;
var clkCheckS abctime;
var clkCheckU abctime;

// iterator thought the cubes
// Original macro:
// #define Bmc_SopForEachCube( pSop, nVars, pCube )  for ( pCube = (pSop); *pCube; pCube += (nVars) + 3 )


func Bmc_CollapseOne_int3(pSat0 *sat_solver, pSat1 *sat_solver, pSat2 *sat_solver, pSat3 *sat_solver, nVars int32, nCubeLim int32, nBTLimit int32, fCanon int32, fReverse int32, fVerbose int32) *Vec_Str_t {
    var fVeryVerbose int32  = fVerbose;
    var pSat [2]*sat_solver = [2]*sat_solver{ pSat0, pSat1 };
    var pSatClean [2]*sat_solver = [2]*sat_solver{ pSat2, pSat3 };
    var vSop [2]*Vec_Str_t = [2]*Vec_Str_t{ Vec_StrAlloc(1000),  Vec_StrAlloc(1000)  };
    var vRes *Vec_Str_t = nil;
    var vLitsC [2]*Vec_Int_t = [2]*Vec_Int_t{ Vec_IntAlloc(nVars), Vec_IntAlloc(nVars) };
    var vVars *Vec_Int_t  = Vec_IntAlloc( nVars );
    var vLits *Vec_Int_t  = Vec_IntAlloc( nVars );
    var vNums *Vec_Int_t  = Vec_IntAlloc( nVars );
    var vCube *Vec_Int_t  = Vec_IntAlloc( nVars );
    var n, v, iVar, iLit, iCiVarBeg, iCube, Start, status int32;
    var clk abctime;
    var Time [2][2]abctime = [2][2]abctime{{0}};
    var fComplete [2]int32 ;

    // collect CI variables
//    iCiVarBeg = pCnf->nVars - nVars;// - 1;
    iCiVarBeg = sat_solver_nvars(pSat0) - nVars;
    if fReverse != 0 {
        for  v = nVars - 1; v >= 0; v--  {
            Vec_IntPush( vVars, iCiVarBeg + v );
        }
    } else {
        for  v = 0; v < nVars; v++  {
            Vec_IntPush( vVars, iCiVarBeg + v );
        } }

    // check that on-set/off-set is sat
    for  n = 0; n < 2; n++  {
        iLit = Abc_Var2Lit( 1, n ); // n=0 => F=1   n=1 => F=0
        status = sat_solver_solve( pSat[n], &iLit, &iLit + 1, nBTLimit, 0, 0, 0 );
        if status == l_Undef {
            goto cleanup; } // timeout
        if status == l_False {
            Vec_StrClear( vSop[0] );
            Vec_StrPrintStr( vSop[0], n ? " 1\n" : " 0\n" );
            Vec_StrPush( vSop[0], '\0' );
            fComplete[0] = 1;
            goto cleanup; // const0/1
        }
        // start with all negative literals
        // inlining of macro Vec_IntForEachEntry
        for  v = 0; (v < Vec_IntSize(vVars)) && (((iVar) = Vec_IntEntry(vVars, v)), 1); v++  {
            Vec_IntPush( vLitsC[n], Abc_Var2Lit(iVar, 1) );
        }
        // add literals to the solver
        status = sat_solver_addclause( pSat[n], &iLit, &iLit + 1 );
        assert((status != 0), "status");
        status = sat_solver_addclause( pSatClean[n], &iLit, &iLit + 1 );
        assert((status != 0), "status");
        // start cover
        Vec_StrPush( vSop[n], '\0' );
    }

    // compute cube pairs
    for  iCube = 0; nCubeLim == 0 || iCube < nCubeLim; iCube++  {
        for  n = 0; n < 2; n++  {
            if fVeryVerbose != 0 { clk = Abc_Clock(); }
            // get the assignment
            if fCanon != 0 {
                status = Bmc_ComputeCanonical( pSat[n], vLitsC[n], vCube, nBTLimit );
            } else {
                sat_solver_clean_polarity( pSat[n], Vec_IntArray(vVars), Vec_IntSize(vVars) );
                status = sat_solver_solve( pSat[n], nil, nil, 0, 0, 0, 0 );
            }
            if fVeryVerbose != 0 { Time[n][0] += Abc_Clock() - clk; }
            if status == l_Undef {
                goto cleanup; } // timeout
            if status == l_False {
                fComplete[n] = 1;
                break;
            }
            // collect values
            Vec_IntClear( vLits );
            Vec_IntClear( vLitsC[n] );
            // inlining of macro Vec_IntForEachEntry
            for  v = 0; (v < Vec_IntSize(vVars)) && (((iVar) = Vec_IntEntry(vVars, v)), 1); v++  {
                iLit = Abc_Var2Lit(iVar, !sat_solver_var_value(pSat[n], iVar));
                Vec_IntPush( vLits, iLit );
                Vec_IntPush( vLitsC[n], iLit );
            }
            // expand the values
            if fVeryVerbose != 0 { clk = Abc_Clock(); }
            status = Bmc_CollapseExpand( pSatClean[!n], pSat[n], vLits, vNums, vCube, nBTLimit, fCanon, -1 );
            if fVeryVerbose != 0 { Time[n][1] += Abc_Clock() - clk; }
            if status < 0 {
                goto cleanup; } // timeout
            // collect cube
            Vec_StrPop( vSop[n] );
            Start = Vec_StrSize( vSop[n] );
            Vec_StrFillExtra( vSop[n], Start + nVars + 4, '-' );
            Vec_StrWriteEntry( vSop[n], Start + nVars + 0, ' ' );
            Vec_StrWriteEntry( vSop[n], Start + nVars + 1, byte((n ? '0' : '1')) );
            Vec_StrWriteEntry( vSop[n], Start + nVars + 2, '\n' );
            Vec_StrWriteEntry( vSop[n], Start + nVars + 3, '\0' );
            Vec_IntClear( vCube );
            // inlining of macro Vec_IntForEachEntry
            for  v = 0; (v < Vec_IntSize(vNums)) && (((iVar) = Vec_IntEntry(vNums, v)), 1); v++  {
                iLit = Vec_IntEntry( vLits, iVar );
                Vec_IntPush( vCube, Abc_LitNot(iLit) );
                if fReverse != 0 {
                    Vec_StrWriteEntry( vSop[n], Start + nVars - iVar - 1, byte(('0' + !Abc_LitIsCompl(iLit))) );
                } else { 
                    Vec_StrWriteEntry( vSop[n], Start + iVar, byte(('0' + !Abc_LitIsCompl(iLit))) ); }
            }
            // add cube
            status = sat_solver_addclause( pSat[n], Vec_IntArray(vCube), Vec_IntLimit(vCube) );
            if status == 0 {
                fComplete[n] = 1;
                break;
            }
            assert(status == 1, "status == 1");
        }
        if fComplete[0] || fComplete[1] {
            break; }
    }
cleanup:
    Vec_IntFree( vVars );
    Vec_IntFree( vLits );
    Vec_IntFree( vLitsC[0] );
    Vec_IntFree( vLitsC[1] );
    Vec_IntFree( vNums );
    Vec_IntFree( vCube );
    assert(!fComplete[0] || !fComplete[1], "!fComplete[0] || !fComplete[1]");
    if fComplete[0] || fComplete[1] { // one of the cover is computed
    
        vRes = vSop[fComplete[1]]; vSop[fComplete[1]] = nil;
        if iCube > 1 {
//            Bmc_CollapseIrredundant( vRes, Vec_StrSize(vRes)/(nVars +3), nVars );
            Bmc_CollapseIrredundantFull( vRes, Vec_StrSize(vRes)/(nVars +3), nVars ); }
    }
    if fVeryVerbose != 0 {
        var fProfile int32;
        fmt.Printf("Processed output with %d supp vars. ", nVars);
        if vRes == nil {
            fmt.Printf("The resulting SOP exceeded %d cubes.\n", nCubeLim);
        } else { 
            fmt.Printf("The best cover contains %d cubes.\n", Vec_StrSize(vRes)/(nVars +3)); }
        Abc_PrintTime( 1, "Onset  minterm", Time[0][0] );
        Abc_PrintTime( 1, "Onset  expand ", Time[0][1] );
        Abc_PrintTime( 1, "Offset minterm", Time[1][0] );
        Abc_PrintTime( 1, "Offset expand ", Time[1][1] );
        if fProfile != 0 {
            Abc_PrintTime( 1, "Expand check1 ", clkCheck1 ); clkCheck1 = 0;
            Abc_PrintTime( 1, "Expand check2 ", clkCheck2 ); clkCheck2 = 0;
            Abc_PrintTime( 1, "Expand sat    ", clkCheckS ); clkCheckS = 0;
            Abc_PrintTime( 1, "Expand unsat  ", clkCheckU ); clkCheckU = 0;
        }
    }
    Vec_StrFreeP( &vSop[0] );
    Vec_StrFreeP( &vSop[1] );
    return vRes;
}
func Bmc_CollapseOne_int2(pSat1 *sat_solver, pSat2 *sat_solver, nVars int32, nCubeLim int32, nBTLimit int32, fCanon int32, fReverse int32, fVerbose int32) *Vec_Str_t {
    var fVeryVerbose int32  = fVerbose;
    var pSat [2]*sat_solver = [2]*sat_solver{ pSat1, pSat2 };
    var vSop [2]*Vec_Str_t = [2]*Vec_Str_t{ Vec_StrAlloc(1000),  Vec_StrAlloc(1000)  };
    var vRes *Vec_Str_t = nil;
    var vVars *Vec_Int_t  = Vec_IntAlloc( nVars+1 );
    var vLits *Vec_Int_t  = Vec_IntAlloc( nVars+1 );
    var vNums *Vec_Int_t  = Vec_IntAlloc( nVars+1 );
    var vCube *Vec_Int_t  = Vec_IntAlloc( nVars+1 );
    var n int32;
    var v int32;
    var iVar int32;
    var pLits [2]int32;
    var iCube int32;
    var Start int32;
    var status int32;
    var clk abctime;
    var Time [2][2]abctime = [2][2]abctime{{0}};
    var fComplete [2]int32 ;
    // variables
    var iOutVar int32     = 2;
    var iOOVars [2]int32 = [2]int32{0, 1};
//    int iOutVar    = 1;
//    int iOOVars[2] = {sat_solver_nvars(pSat) - 5, sat_solver_nvars(pSat) - 5 + 1};

    // collect CI variables (0 = onset enable, 1 = offset enable, 2 = output)
    var iCiVarBeg int32  = 3;
//    int iCiVarBeg = sat_solver_nvars(pSat) - 5 - nVars;
    if fReverse != 0 {
        for  v = nVars - 1; v >= 0; v--  {
            Vec_IntPush( vVars, iCiVarBeg + v );
        }
    } else {
        for  v = 0; v < nVars; v++  {
            Vec_IntPush( vVars, iCiVarBeg + v );
        } }

    // check that on-set/off-set is sat
    for  n = 0; n < 2; n++  {
        pLits[0] = Abc_Var2Lit( iOutVar, n ); // n=0 => F=1   n=1 => F=0
        status = sat_solver_solve( pSat[n], pLits, pLits + 1, nBTLimit, 0, 0, 0 );
        if status == l_Undef {
            goto cleanup; } // timeout
        if status == l_False {
            Vec_StrClear( vSop[0] );
            Vec_StrPrintStr( vSop[0], n ? " 1\n" : " 0\n" );
            Vec_StrPush( vSop[0], '\0' );
            fComplete[0] = 1;
            goto cleanup; // const0/1
        }
        // add literals to the solver
        status = sat_solver_addclause( pSat[n], pLits, pLits + 1 );
        assert((status != 0), "status");
        // start cover
        Vec_StrPush( vSop[n], '\0' );
    }

    // compute cube pairs
    for  iCube = 0; nCubeLim == 0 || iCube < nCubeLim; iCube++  {
        for  n = 0; n < 2; n++  {
            if fVeryVerbose != 0 { clk = Abc_Clock(); }
            // get the assignment
            sat_solver_clean_polarity( pSat[n], Vec_IntArray(vVars), Vec_IntSize(vVars) );
            pLits[0] = Abc_Var2Lit( iOOVars[n], 1 ); // enable clauses
//            pLits[1] = Abc_Var2Lit( iOutVar, n );    // set output
//            status = sat_solver_solve( pSat, pLits, pLits + 2, 0, 0, 0, 0 );
            status = sat_solver_solve( pSat[n], pLits, pLits + 1, 0, 0, 0, 0 );
            if fVeryVerbose != 0 { Time[n][0] += Abc_Clock() - clk; }
            if status == l_Undef {
                goto cleanup; } // timeout
            if status == l_False {
                fComplete[n] = 1;
                break;
            }
            // collect values
            Vec_IntClear( vLits );
            // inlining of macro Vec_IntForEachEntry
            for  v = 0; (v < Vec_IntSize(vVars)) && (((iVar) = Vec_IntEntry(vVars, v)), 1); v++  {
                Vec_IntPush( vLits, Abc_Var2Lit(iVar, !sat_solver_var_value(pSat[n], iVar)) );
            }
            // expand the values
            if fVeryVerbose != 0 { clk = Abc_Clock(); }
//            status = Bmc_CollapseExpand( pSat, NULL, vLits, vNums, vCube, nBTLimit, fCanon, Abc_Var2Lit(iOutVar, !n) );
            status = Bmc_CollapseExpand( pSat[!n], nil, vLits, vNums, vCube, nBTLimit, fCanon, -1 );
            if fVeryVerbose != 0 { Time[n][1] += Abc_Clock() - clk; }
            if status < 0 {
                goto cleanup; } // timeout
            // collect cube
            Vec_StrPop( vSop[n] );
            Start = Vec_StrSize( vSop[n] );
            Vec_StrFillExtra( vSop[n], Start + nVars + 4, '-' );
            Vec_StrWriteEntry( vSop[n], Start + nVars + 0, ' ' );
            Vec_StrWriteEntry( vSop[n], Start + nVars + 1, byte((n ? '0' : '1')) );
            Vec_StrWriteEntry( vSop[n], Start + nVars + 2, '\n' );
            Vec_StrWriteEntry( vSop[n], Start + nVars + 3, '\0' );
            Vec_IntClear( vCube );
            Vec_IntPush( vCube, Abc_Var2Lit( iOOVars[n], 0 ) );
            // inlining of macro Vec_IntForEachEntry
            for  v = 0; (v < Vec_IntSize(vNums)) && (((iVar) = Vec_IntEntry(vNums, v)), 1); v++  {
                var iLit int32  = Vec_IntEntry( vLits, iVar );
                Vec_IntPush( vCube, Abc_LitNot(iLit) );
                if fReverse != 0 {
                    Vec_StrWriteEntry( vSop[n], Start + nVars - iVar - 1, byte(('0' + !Abc_LitIsCompl(iLit))) );
                } else { 
                    Vec_StrWriteEntry( vSop[n], Start + iVar, byte(('0' + !Abc_LitIsCompl(iLit))) ); }
            }
            // add cube
//            status = sat_solver_addclause( pSat, Vec_IntArray(vCube), Vec_IntLimit(vCube) );
            status = sat_solver_addclause( pSat[n], Vec_IntArray(vCube), Vec_IntLimit(vCube) );
            if status == 0 {
                fComplete[n] = 1;
                break;
            }
            assert(status == 1, "status == 1");
        }
        if fComplete[0] || fComplete[1] {
            break; }
    }
cleanup:
    Vec_IntFree( vVars );
    Vec_IntFree( vLits );
    Vec_IntFree( vNums );
    Vec_IntFree( vCube );
    assert(!fComplete[0] || !fComplete[1], "!fComplete[0] || !fComplete[1]");
    if fComplete[0] || fComplete[1] { // one of the cover is computed
    
        vRes = vSop[fComplete[1]]; vSop[fComplete[1]] = nil;
        if iCube > 1 {
//            Bmc_CollapseIrredundant( vRes, Vec_StrSize(vRes)/(nVars +3), nVars );
            Bmc_CollapseIrredundantFull( vRes, Vec_StrSize(vRes)/(nVars +3), nVars ); }
    }
    if fVeryVerbose != 0 {
        var fProfile int32;
        fmt.Printf("Processed output with %d supp vars. ", nVars);
        if vRes == nil {
            fmt.Printf("The resulting SOP exceeded %d cubes.\n", nCubeLim);
        } else { 
            fmt.Printf("The best cover contains %d cubes.\n", Vec_StrSize(vRes)/(nVars +3)); }
        Abc_PrintTime( 1, "Onset  minterm", Time[0][0] );
        Abc_PrintTime( 1, "Onset  expand ", Time[0][1] );
        Abc_PrintTime( 1, "Offset minterm", Time[1][0] );
        Abc_PrintTime( 1, "Offset expand ", Time[1][1] );
        if fProfile != 0 {
            Abc_PrintTime( 1, "Expand check1 ", clkCheck1 ); clkCheck1 = 0;
            Abc_PrintTime( 1, "Expand check2 ", clkCheck2 ); clkCheck2 = 0;
            Abc_PrintTime( 1, "Expand sat    ", clkCheckS ); clkCheckS = 0;
            Abc_PrintTime( 1, "Expand unsat  ", clkCheckU ); clkCheckU = 0;
        }
    }
    Vec_StrFreeP( &vSop[0] );
    Vec_StrFreeP( &vSop[1] );
    return vRes;
}


/**Function*************************************************************

  Synopsis    [This code computes on-set and off-set simultaneously.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Bmc_CollapseOne_int(pSat *sat_solver, nVars int32, nCubeLim int32, nBTLimit int32, fCanon int32, fReverse int32, fVerbose int32) *Vec_Str_t {
    var fVeryVerbose int32  = fVerbose;
    var vSop [2]*Vec_Str_t = [2]*Vec_Str_t{ Vec_StrAlloc(1000),  Vec_StrAlloc(1000)  };
    var vRes *Vec_Str_t = nil;
    var vVars *Vec_Int_t  = Vec_IntAlloc( nVars+1 );
    var vLits *Vec_Int_t  = Vec_IntAlloc( nVars+1 );
    var vNums *Vec_Int_t  = Vec_IntAlloc( nVars+1 );
    var vCube *Vec_Int_t  = Vec_IntAlloc( nVars+1 );
    var n int32;
    var v int32;
    var iVar int32;
    var pLits [2]int32;
    var iCube int32;
    var Start int32;
    var status int32;
    var clk abctime;
    var Time [2][2]abctime = [2][2]abctime{{0}};
    var fComplete [2]int32 ;
    // variables
    var iOutVar int32     = 2;
    var iOOVars [2]int32 = [2]int32{0, 1};
//    int iOutVar    = 1;
//    int iOOVars[2] = {sat_solver_nvars(pSat) - 5, sat_solver_nvars(pSat) - 5 + 1};

    // collect CI variables (0 = onset enable, 1 = offset enable, 2 = output)
    var iCiVarBeg int32  = 3;
//    int iCiVarBeg = sat_solver_nvars(pSat) - 5 - nVars;
    if fReverse != 0 {
        for  v = nVars - 1; v >= 0; v--  {
            Vec_IntPush( vVars, iCiVarBeg + v );
        }
    } else {
        for  v = 0; v < nVars; v++  {
            Vec_IntPush( vVars, iCiVarBeg + v );
        } }

    // check that on-set/off-set is sat
    for  n = 0; n < 2; n++  {
        pLits[0] = Abc_Var2Lit( iOutVar, n ); // n=0 => F=1   n=1 => F=0
        status = sat_solver_solve( pSat, pLits, pLits + 1, nBTLimit, 0, 0, 0 );
        if status == l_Undef {
            goto cleanup; } // timeout
        if status == l_False {
            Vec_StrClear( vSop[0] );
            Vec_StrPrintStr( vSop[0], n ? " 1\n" : " 0\n" );
            Vec_StrPush( vSop[0], '\0' );
            fComplete[0] = 1;
            goto cleanup; // const0/1
        }
        // start cover
        Vec_StrPush( vSop[n], '\0' );
    }

    // compute cube pairs
    for  iCube = 0; nCubeLim == 0 || iCube < nCubeLim; iCube++  {
        for  n = 0; n < 2; n++  {
            if fVeryVerbose != 0 { clk = Abc_Clock(); }
            // get the assignment
            sat_solver_clean_polarity( pSat, Vec_IntArray(vVars), Vec_IntSize(vVars) );
            pLits[0] = Abc_Var2Lit( iOutVar, n );    // set output
            pLits[1] = Abc_Var2Lit( iOOVars[n], 1 ); // enable clauses
            status = sat_solver_solve( pSat, pLits, pLits + 2, 0, 0, 0, 0 );
            if fVeryVerbose != 0 { Time[n][0] += Abc_Clock() - clk; }
            if status == l_Undef {
                goto cleanup; } // timeout
            if status == l_False {
                fComplete[n] = 1;
                break;
            }
            // collect values
            Vec_IntClear( vLits );
            // inlining of macro Vec_IntForEachEntry
            for  v = 0; (v < Vec_IntSize(vVars)) && (((iVar) = Vec_IntEntry(vVars, v)), 1); v++  {
                Vec_IntPush( vLits, Abc_Var2Lit(iVar, !sat_solver_var_value(pSat, iVar)) );
            }
            // expand the values
            if fVeryVerbose != 0 { clk = Abc_Clock(); }
            status = Bmc_CollapseExpand( pSat, nil, vLits, vNums, vCube, nBTLimit, fCanon, Abc_Var2Lit(iOutVar, !n) );
            if fVeryVerbose != 0 { Time[n][1] += Abc_Clock() - clk; }
            if status < 0 {
                goto cleanup; } // timeout
            // collect cube
            Vec_StrPop( vSop[n] );
            Start = Vec_StrSize( vSop[n] );
            Vec_StrFillExtra( vSop[n], Start + nVars + 4, '-' );
            Vec_StrWriteEntry( vSop[n], Start + nVars + 0, ' ' );
            Vec_StrWriteEntry( vSop[n], Start + nVars + 1, byte((n ? '0' : '1')) );
            Vec_StrWriteEntry( vSop[n], Start + nVars + 2, '\n' );
            Vec_StrWriteEntry( vSop[n], Start + nVars + 3, '\0' );
            Vec_IntClear( vCube );
            Vec_IntPush( vCube, Abc_Var2Lit( iOOVars[n], 0 ) );
            // inlining of macro Vec_IntForEachEntry
            for  v = 0; (v < Vec_IntSize(vNums)) && (((iVar) = Vec_IntEntry(vNums, v)), 1); v++  {
                var iLit int32  = Vec_IntEntry( vLits, iVar );
                Vec_IntPush( vCube, Abc_LitNot(iLit) );
                if fReverse != 0 {
                    Vec_StrWriteEntry( vSop[n], Start + nVars - iVar - 1, byte(('0' + !Abc_LitIsCompl(iLit))) );
                } else { 
                    Vec_StrWriteEntry( vSop[n], Start + iVar, byte(('0' + !Abc_LitIsCompl(iLit))) ); }
            }
            // add cube
            status = sat_solver_addclause( pSat, Vec_IntArray(vCube), Vec_IntLimit(vCube) );
            if status == 0 {
                fComplete[n] = 1;
                break;
            }
            assert(status == 1, "status == 1");
        }
        if fComplete[0] || fComplete[1] {
            break; }
    }
cleanup:
    Vec_IntFree( vVars );
    Vec_IntFree( vLits );
    Vec_IntFree( vNums );
    Vec_IntFree( vCube );
    assert(!fComplete[0] || !fComplete[1], "!fComplete[0] || !fComplete[1]");
    if fComplete[0] || fComplete[1] { // one of the cover is computed
    
        vRes = vSop[fComplete[1]]; vSop[fComplete[1]] = nil;
        if iCube > 1 {
//            Bmc_CollapseIrredundant( vRes, Vec_StrSize(vRes)/(nVars +3), nVars );
            Bmc_CollapseIrredundantFull( vRes, Vec_StrSize(vRes)/(nVars +3), nVars ); }
    }
    if fVeryVerbose != 0 {
        var fProfile int32;
        fmt.Printf("Processed output with %d supp vars. ", nVars);
        if vRes == nil {
            fmt.Printf("The resulting SOP exceeded %d cubes.\n", nCubeLim);
        } else { 
            fmt.Printf("The best cover contains %d cubes.\n", Vec_StrSize(vRes)/(nVars +3)); }
        Abc_PrintTime( 1, "Onset  minterm", Time[0][0] );
        Abc_PrintTime( 1, "Onset  expand ", Time[0][1] );
        Abc_PrintTime( 1, "Offset minterm", Time[1][0] );
        Abc_PrintTime( 1, "Offset expand ", Time[1][1] );
        if fProfile != 0 {
            Abc_PrintTime( 1, "Expand check1 ", clkCheck1 ); clkCheck1 = 0;
            Abc_PrintTime( 1, "Expand check2 ", clkCheck2 ); clkCheck2 = 0;
            Abc_PrintTime( 1, "Expand sat    ", clkCheckS ); clkCheckS = 0;
            Abc_PrintTime( 1, "Expand unsat  ", clkCheckU ); clkCheckU = 0;
        }
    }
    Vec_StrFreeP( &vSop[0] );
    Vec_StrFreeP( &vSop[1] );
    return vRes;
}
