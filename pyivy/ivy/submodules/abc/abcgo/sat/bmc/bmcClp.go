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


/**Function*************************************************************

  Synopsis    []

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Bmc_CollapseOneInt2(p *Gia_Man_t, nCubeLim int32, nBTLimit int32, fCanon int32, fReverse int32, fVerbose int32, fCompl int32) *Vec_Str_t {
    var fPrintMinterm int32;
    var nVars int32  = Gia_ManCiNum(p);
    var vVars *Vec_Int_t  = Vec_IntAlloc( nVars );
    var vLits *Vec_Int_t  = Vec_IntAlloc( nVars );
    var vLitsC *Vec_Int_t = Vec_IntAlloc( nVars );
    var vNums *Vec_Int_t  = Vec_IntAlloc( nVars );
    var vCube *Vec_Int_t  = Vec_IntAlloc( nVars );
    var vSop *Vec_Str_t   = Vec_StrAlloc( 100 );
    var iOut, iLit, iVar, status, n, Count, Start int32;

    // create SAT solver
    var pCnf *Cnf_Dat_t  = (Cnf_Dat_t *)Mf_ManGenerateCnf( p, 8, 0, 0, 0, 0 );
    var pSat [3]*sat_solver = [3]*sat_solver{ 
        (sat_solver *)Cnf_DataWriteIntoSolver(pCnf, 1, 0), 
        (sat_solver *)Cnf_DataWriteIntoSolver(pCnf, 1, 0),
        fCanon ? (sat_solver *)Cnf_DataWriteIntoSolver(pCnf, 1, 0) : nil,
    };

    // collect CI variables
    var iCiVarBeg int32  = pCnf.nVars - nVars;// - 1;
    if fReverse != 0 {
        for  n = nVars - 1; n >= 0; n--  {
            Vec_IntPush( vVars, iCiVarBeg + n );
        }
    } else {
        for  n = 0; n < nVars; n++  {
            Vec_IntPush( vVars, iCiVarBeg + n );
        } }

    // start with all negative literals
    // inlining of macro Vec_IntForEachEntry
    for  n = 0; (n < Vec_IntSize(vVars)) && (((iVar) = Vec_IntEntry(vVars, n)), 1); n++  {
        Vec_IntPush( vLitsC, Abc_Var2Lit(iVar, 1) );
    }

    // check that on-set/off-set is sat
    for  n = 0; n < 2 + fCanon; n++  {
        iLit = Abc_Var2Lit( iOut + 1, n&1 ); // n=0 => F=1   n=1 => F=0
        status = sat_solver_addclause( pSat[n], &iLit, &iLit + 1 );
        if status == 0 {
            Vec_StrPrintStr( vSop, ((n&1) ^ fCompl) ? " 1\n" : " 0\n" );
            Vec_StrPush( vSop, '\0' );
            goto cleanup; // const0/1
        }
        status = sat_solver_solve( pSat[n], nil, nil, nBTLimit, 0, 0, 0 );
        if status == l_Undef {
            Vec_StrFreeP( &vSop );
            goto cleanup; // timeout
        }
        if status == l_False {
            Vec_StrPrintStr( vSop, ((n&1) ^ fCompl) ? " 1\n" : " 0\n" );
            Vec_StrPush( vSop, '\0' );
            goto cleanup; // const0/1
        }
    }
    Vec_StrPush( vSop, '\0' );

    // prepare on-set for solving
//    if ( fCanon )
//        sat_solver_prepare_enum( pSat[0], Vec_IntArray(vVars), Vec_IntSize(vVars) );
    Count = 0;
    for {
        // get the assignment
        if fCanon != 0 {
            status = Bmc_ComputeCanonical( pSat[0], vLitsC, vCube, nBTLimit );
        } else {
            sat_solver_clean_polarity( pSat[0], Vec_IntArray(vVars), Vec_IntSize(vVars) );
            status = sat_solver_solve( pSat[0], nil, nil, 0, 0, 0, 0 );
        }
        if status == l_Undef {
            Vec_StrFreeP( &vSop );
            goto cleanup; // timeout
        }
        if status == l_False {
            break; }
        // check number of cubes
        if nCubeLim > 0 && Count == nCubeLim {
            //printf( "The number of cubes exceeded the limit (%d).\n", nCubeLim );
            Vec_StrFreeP( &vSop );
            goto cleanup; // cube out
        }
        // collect values
        Vec_IntClear( vLits );
        Vec_IntClear( vLitsC );
        // inlining of macro Vec_IntForEachEntry
        for  n = 0; (n < Vec_IntSize(vVars)) && (((iVar) = Vec_IntEntry(vVars, n)), 1); n++  {
            iLit = Abc_Var2Lit(iVar, !sat_solver_var_value(pSat[0], iVar));
            Vec_IntPush( vLits, iLit );
            Vec_IntPush( vLitsC, iLit );
        }
        // print minterm
        if fPrintMinterm != 0 {
            fmt.Printf("Mint: ");
            // inlining of macro Vec_IntForEachEntry
            for  n = 0; (n < Vec_IntSize(vLits)) && (((iLit) = Vec_IntEntry(vLits, n)), 1); n++  {
                fmt.Printf("%d", !Abc_LitIsCompl(iLit));
            }
            fmt.Printf("\n");
        }
        // expand the values
        status = Bmc_CollapseExpand( pSat[1], fCanon ? pSat[2] : pSat[0], vLits, vNums, vCube, nBTLimit, fCanon, -1 );
        if status < 0 {
            Vec_StrFreeP( &vSop );
            goto cleanup; // timeout
        }
        // collect cube
        Vec_StrPop( vSop );
        Start = Vec_StrSize( vSop );
        Vec_StrFillExtra( vSop, Start + nVars + 4, '-' );
        Vec_StrWriteEntry( vSop, Start + nVars + 0, ' ' );
        Vec_StrWriteEntry( vSop, Start + nVars + 1, byte((fCompl ? '0' : '1')) );
        Vec_StrWriteEntry( vSop, Start + nVars + 2, '\n' );
        Vec_StrWriteEntry( vSop, Start + nVars + 3, '\0' );
        Vec_IntClear( vCube );
        // inlining of macro Vec_IntForEachEntry
        for  n = 0; (n < Vec_IntSize(vNums)) && (((iVar) = Vec_IntEntry(vNums, n)), 1); n++  {
            iLit = Vec_IntEntry( vLits, iVar );
            Vec_IntPush( vCube, Abc_LitNot(iLit) );
            if fReverse != 0 {
                Vec_StrWriteEntry( vSop, Start + nVars - iVar - 1, byte(('0' + !Abc_LitIsCompl(iLit))) );
            } else { 
                Vec_StrWriteEntry( vSop, Start + iVar, byte(('0' + !Abc_LitIsCompl(iLit))) ); }
        }
        //if ( fVerbose )
        //    printf( "Cube %4d: %s", Count, Vec_StrArray(vSop) + Start );
        Count++;
        // add cube
        status = sat_solver_addclause( pSat[0], Vec_IntArray(vCube), Vec_IntLimit(vCube) );
        if status == 0 {
            break; }
        // add cube
        if fCanon != 0 {
            status = sat_solver_addclause( pSat[2], Vec_IntArray(vCube), Vec_IntLimit(vCube) ); }
        assert(status == 1, "status == 1");
    }
    //printf( "Finished enumerating %d assignments.\n", Count );
cleanup:
    Vec_IntFree( vVars );
    Vec_IntFree( vLits );
    Vec_IntFree( vLitsC );
    Vec_IntFree( vNums );
    Vec_IntFree( vCube );
    sat_solver_delete( pSat[0] );
    sat_solver_delete( pSat[1] );
    if fCanon != 0 {
    sat_solver_delete( pSat[2] ); }
    Cnf_DataFree( pCnf );
    // quickly reduce contained cubes
    if vSop != nil {
        Bmc_CollapseIrredundant( vSop, Vec_StrSize(vSop)/(nVars +3), nVars ); }
    return vSop;
}
func Bmc_CollapseOneOld2(p *Gia_Man_t, nCubeLim int32, nBTLimit int32, fCanon int32, fReverse int32, fVerbose int32) *Vec_Str_t {
    var vSopOn, vSopOff *Vec_Str_t;
    var nCubesOn int32  = ABC_INFINITY;
    var nCubesOff int32  = ABC_INFINITY;
    vSopOn = Bmc_CollapseOneInt2( p, nCubeLim, nBTLimit, fCanon, fReverse, fVerbose, 0 );
    if vSopOn {
        nCubesOn = Vec_StrCountEntry(vSopOn,'\n'); }
    Gia_ObjFlipFaninC0( Gia_ManPo(p, 0) );
    vSopOff = Bmc_CollapseOneInt2( p, Abc_MinInt(nCubeLim, nCubesOn), nBTLimit, fCanon, fReverse, fVerbose, 1 );
    Gia_ObjFlipFaninC0( Gia_ManPo(p, 0) );
    if vSopOff != nil {
        nCubesOff = Vec_StrCountEntry(vSopOff,'\n'); }
    if vSopOn == nil {
        return vSopOff; }
    if vSopOff == nil {
        return vSopOn; }
    if nCubesOn <= nCubesOff {
        Vec_StrFree( vSopOff );
        return vSopOn;
    
    } else {
        Vec_StrFree( vSopOn );
        return vSopOff;
    }
}


/**Function*************************************************************

  Synopsis    [This code computes on-set and off-set simultaneously.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Bmc_CollapseOneOld(p *Gia_Man_t, nCubeLim int32, nBTLimit int32, fCanon int32, fReverse int32, fVerbose int32) *Vec_Str_t {
    var fVeryVerbose int32  = fVerbose;
    var nVars int32  = Gia_ManCiNum(p);
    var pCnf *Cnf_Dat_t  = (Cnf_Dat_t *)Mf_ManGenerateCnf( p, 8, 0, 0, 0, 0 );
    var pSat [2]*sat_solver = [2]*sat_solver{ (sat_solver *)Cnf_DataWriteIntoSolver(pCnf, 1, 0), (sat_solver *)Cnf_DataWriteIntoSolver(pCnf, 1, 0) };
    var pSatClean [2]*sat_solver = [2]*sat_solver{ (sat_solver *)Cnf_DataWriteIntoSolver(pCnf, 1, 0), (sat_solver *)Cnf_DataWriteIntoSolver(pCnf, 1, 0) };
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
    iCiVarBeg = pCnf.nVars - nVars;// - 1;
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
    Cnf_DataFree( pCnf );
    sat_solver_delete( pSat[0] );
    sat_solver_delete( pSat[1] );
    sat_solver_delete( pSatClean[0] );
    sat_solver_delete( pSatClean[1] );
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
func Bmc_CollapseOne3(p *Gia_Man_t, nCubeLim int32, nBTLimit int32, fCanon int32, fReverse int32, fVerbose int32) *Vec_Str_t {
    var pCnf *Cnf_Dat_t  = (Cnf_Dat_t *)Mf_ManGenerateCnf( p, 8, 0, 0, 0, 0 );
    var pSat0 *sat_solver  = (sat_solver *)Cnf_DataWriteIntoSolver(pCnf, 1, 0);
    var pSat1 *sat_solver  = (sat_solver *)Cnf_DataWriteIntoSolver(pCnf, 1, 0);
    var pSat2 *sat_solver  = (sat_solver *)Cnf_DataWriteIntoSolver(pCnf, 1, 0);
    var pSat3 *sat_solver  = (sat_solver *)Cnf_DataWriteIntoSolver(pCnf, 1, 0);
    var vSop *Vec_Str_t  = Bmc_CollapseOne_int3( pSat0, pSat1, pSat2, pSat3, Gia_ManCiNum(p), nCubeLim, nBTLimit, fCanon, fReverse, fVerbose );
    sat_solver_delete( pSat0 );
    sat_solver_delete( pSat1 );
    sat_solver_delete( pSat2 );
    sat_solver_delete( pSat3 );
    Cnf_DataFree( pCnf );
    return vSop;
}


/**Function*************************************************************

  Synopsis    [This code computes on-set and off-set simultaneously.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
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
func Bmc_CollapseOne(p *Gia_Man_t, nCubeLim int32, nBTLimit int32, fCanon int32, fReverse int32, fVerbose int32) *Vec_Str_t {
    var pCnf *Cnf_Dat_t  = (Cnf_Dat_t *)Mf_ManGenerateCnf( p, 8, 0, 0, 0, 0 );
    var pSat *sat_solver  = (sat_solver *)Cnf_DataWriteIntoSolver(pCnf, 1, 0);
    var vSop *Vec_Str_t  = Bmc_CollapseOne_int( pSat, Gia_ManCiNum(p), nCubeLim, nBTLimit, fCanon, fReverse, fVerbose );
    sat_solver_delete( pSat );
    Cnf_DataFree( pCnf );
    return vSop;
}

////////////////////////////////////////////////////////////////////////
///                       END OF FILE                                ///
////////////////////////////////////////////////////////////////////////


//ABC_NAMESPACE_IMPL_END

