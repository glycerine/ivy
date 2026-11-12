package sbd

import "fmt"

/**CFile****************************************************************

  FileName    [sbdLut.c]

  SystemName  [ABC: Logic synthesis and verification system.]

  PackageName [SAT-based optimization using internal don't-cares.]

  Synopsis    [CNF computation.]

  Author      [Alan Mishchenko]
  
  Affiliation [UC Berkeley]

  Date        [Ver. 1.0. Started - June 20, 2005.]

  Revision    [$Id: sbdLut.c,v 1.00 2005/06/20 00:00:00 alanmi Exp $]

***********************************************************************/

//#include "sbdInt.h"
//#include "misc/util/utilTruth.h"

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
// count the number of parameter variables in the structure
func Sbd_ProblemCountParams(nStrs int32, pStr0 *Sbd_Str_t) int32 {
    var pStr *Sbd_Str_t;  var nPars int32;
    for  pStr = pStr0; pStr < pStr0 + nStrs; pStr++  {
        nPars += pStr.fLut ? 1 << pStr.nVarIns : pStr.nVarIns;
    }
    return nPars;
}
// add clauses for the structure
func Sbd_ProblemAddClauses(pSat *sat_solver, nVars int32, nStrs int32, pVars *int32, pStr0 *Sbd_Str_t) int32 {   
    // variable order:  inputs, structure outputs, parameters
    var pStr *Sbd_Str_t;
    var VarOut int32  = nVars;
    var VarPar int32  = nVars + nStrs;
    var m int32;
    var k int32;
    var n int32;
    var status int32;
    var pLits [SBD_SIZE_MAX+2]int32;
//printf( "Start par = %d.  ", VarPar );
    for  pStr = pStr0; pStr < pStr0 + nStrs; pStr, VarOut = pStr+1, VarOut+1  {
        if pStr.fLut {
            var nMints int32  = 1 << pStr.nVarIns;
            assert(pStr.nVarIns <= 6, "pStr.nVarIns <= 6");
            for  m = 0; m < nMints; m, VarPar = m+1, VarPar+1  {
                for  k = 0; k < pStr.nVarIns; k++  {
                    pLits[k] = Abc_Var2Lit( pVars[pStr.VarIns[k]], (m >> k) & 1 );
                } 
                for  n = 0; n < 2; n++  {
                    pLits[pStr.nVarIns]   = Abc_Var2Lit( pVars[VarPar], n );
                    pLits[pStr.nVarIns+1] = Abc_Var2Lit( pVars[VarOut], !n );
                    status = sat_solver_addclause( pSat, pLits, pLits + pStr.nVarIns + 2 );
                    if !(status != 0) {
                        return 0; }
                }
            }
        
        } else {
            assert(pStr.nVarIns <= SBD_DIV_MAX, "pStr.nVarIns <= SBD_DIV_MAX");
            for  k = 0; k < pStr.nVarIns; k, VarPar = k+1, VarPar+1  {
                for  n = 0; n < 2; n++  {
                    pLits[0] = Abc_Var2Lit( pVars[VarPar], 1 );
                    pLits[1] = Abc_Var2Lit( pVars[VarOut], n );
                    pLits[2] = Abc_Var2Lit( pVars[pStr.VarIns[k]], !n );
                    status = sat_solver_addclause( pSat, pLits, pLits + 3 );
                    if !(status != 0) {
                        return 0; }
                }
            }
        }
    }
    return 1;
}
func Sbd_ProblemAddClausesInit(pSat *sat_solver, nVars int32, nStrs int32, pVars *int32, pStr0 *Sbd_Str_t) {
    var pStr *Sbd_Str_t;
    var VarPar int32  = nVars + nStrs;
    var m int32;
    var m2 int32;
    var status int32;
    var pLits [SBD_DIV_MAX]int32;
    // make sure selector parameters are mutually exclusive
    for  pStr = pStr0; pStr < pStr0 + nStrs; VarPar += pStr.fLut ? 1 << pStr.nVarIns : pStr.nVarIns, pStr++  {
        if pStr.fLut {
            continue; }
        // one variable should be selected
        assert(pStr.nVarIns <= SBD_DIV_MAX, "pStr.nVarIns <= SBD_DIV_MAX");
        for  m = 0; m < pStr.nVarIns; m++  {
            pLits[m] = Abc_Var2Lit( pVars[VarPar + m], 0 );
        }
        status = sat_solver_addclause( pSat, pLits, pLits + pStr.nVarIns );
        assert((status != 0), "status");
        // two variables cannot be selected
        for  m = 0; m < pStr.nVarIns; m++  {
        for  m2 = m+1; m2 < pStr.nVarIns; m2++  {
            pLits[0] = Abc_Var2Lit( pVars[VarPar + m], 1 );
            pLits[1] = Abc_Var2Lit( pVars[VarPar + m2], 1 );
            status = sat_solver_addclause( pSat, pLits, pLits + 2 );
            assert((status != 0), "status");
        }
        }
    }
}
func Sbd_ProblemPrintSolution(nStrs int32, pStr0 *Sbd_Str_t, vLits *Vec_Int_t) {
    var pStr *Sbd_Str_t;
    var m, nIters, iLit int32;
    fmt.Printf("Solution found:\n");
    for  pStr = pStr0; pStr < pStr0 + nStrs; pStr++  {
        if pStr.fLut {
            nIters = 1 << pStr.nVarIns;
        } else {
            nIters = pStr.nVarIns;
        }
        fmt.Printf("%s%d : ", pStr.fLut ? "LUT":"SEL", int32((pStr-pStr0)));
        for  m = 0; m < nIters; m, iLit = m+1, iLit+1  {
            fmt.Printf("%d", !Abc_LitIsCompl(Vec_IntEntry(vLits, iLit)));
        }
        fmt.Printf("    {");
        for  m = 0; m < pStr.nVarIns; m++  {
            fmt.Printf(" %d", pStr.VarIns[m]);
        }
        fmt.Printf(" }\n");
    }
    assert(iLit == Vec_IntSize(vLits), "iLit == Vec_IntSize(vLits)");
}
func Sbd_ProblemCollectSolution(nStrs int32, pStr0 *Sbd_Str_t, vLits *Vec_Int_t) {
    var pStr *Sbd_Str_t;
    var m, nIters, iLit int32;
    for  pStr = pStr0; pStr < pStr0 + nStrs; pStr++  {
        pStr.Res = 0;
        if pStr.fLut {
            nIters = 1 << pStr.nVarIns;
            for  m = 0; m < nIters; m, iLit = m+1, iLit+1  {
                if !Abc_LitIsCompl(Vec_IntEntry(vLits, iLit)) {
                    Abc_TtSetBit( &pStr.Res, m ); }
            }
            pStr.Res = Abc_Tt6Stretch( pStr.Res, pStr.nVarIns );
        
        } else {
            nIters = 0;
            for  m = 0; m < pStr.nVarIns; m, iLit = m+1, iLit+1  {
                if !Abc_LitIsCompl(Vec_IntEntry(vLits, iLit)) {
                    pStr.Res = pStr.VarIns[m];
                    nIters++;
                }
            }
            assert(nIters == 1, "nIters == 1");
        }
    }
    assert(iLit == Vec_IntSize(vLits), "iLit == Vec_IntSize(vLits)");
}

/**Function*************************************************************

  Synopsis    [Solves QBF problem for the given window.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
int Sbd_ProblemSolve( Gia_Man_t * p, Vec_Int_t * vMirrors, 
                       int Pivot, Vec_Int_t * vWinObjs, Vec_Int_t * vObj2Var, 
                       Vec_Int_t * vTfo, Vec_Int_t * vRoots, 
                       Vec_Int_t * vDivSet, int nStrs, Sbd_Str_t * pStr0 )  // divisors, structures
{
    extern sat_solver * Sbd_ManSatSolver( sat_solver * pSat, Gia_Man_t * p, Vec_Int_t * vMirrors, int Pivot, Vec_Int_t * vWinObjs, Vec_Int_t * vObj2Var, Vec_Int_t * vTfo, Vec_Int_t * vRoots, int fQbf );
    
    var fVerbose int32;
    var clk abctime  = Abc_Clock();
    var vLits *Vec_Int_t     = Vec_IntAlloc( 100 );
    var pSatCec *sat_solver  = Sbd_ManSatSolver( nil, p, vMirrors, Pivot, vWinObjs, vObj2Var, vTfo, vRoots, 1 );
    var pSatQbf *sat_solver  = sat_solver_new();

    var nVars int32       = Vec_IntSize( vDivSet );
    var nPars int32       = Sbd_ProblemCountParams( nStrs, pStr0 );

    var VarCecOut int32   = Vec_IntSize(vWinObjs) + Vec_IntSize(vTfo) + Vec_IntSize(vRoots);
    var VarCecPar int32   = VarCecOut + nStrs;

    var VarQbfPar int32;
    var VarQbfFree int32  = nPars;

    var pVarsCec [256]int32;
    var pVarsQbf [256]int32;
    var i, iVar, iLit, nIters int32;
    var RetValue int32;

    assert(Vec_IntSize(vDivSet) <= SBD_DIV_MAX, "Vec_IntSize(vDivSet) <= SBD_DIV_MAX");
    assert(nVars + nStrs + nPars <= 256, "nVars + nStrs + nPars <= 256");

    // collect CEC variables
    Vec_IntForEachEntry( vDivSet, iVar, i )
        pVarsCec[i] = iVar;
    for  i = 0; i < nStrs; i++  {
        pVarsCec[nVars + i] = VarCecOut + i;
    }
    for  i = 0; i < nPars; i++  {
        pVarsCec[nVars + nStrs + i] = VarCecPar + i;
    }

    // collect QBF variables
    for  i = 0; i < nVars + nStrs; i++  {
        pVarsQbf[i] = -1;
    }
    for  i = 0; i < nPars; i++  {
        pVarsQbf[nVars + nStrs + i] = VarQbfPar + i;
    }

    // add clauses to the CEC problem
    Sbd_ProblemAddClauses( pSatCec, nVars, nStrs, pVarsCec, pStr0 );

    // create QBF solver
    sat_solver_setnvars( pSatQbf, 1000 );
    Sbd_ProblemAddClausesInit( pSatQbf, nVars, nStrs, pVarsQbf, pStr0 );

    // assume all parameter variables are 0
    Vec_IntClear( vLits );
    for  i = 0; i < nPars; i++  {
        Vec_IntPush( vLits, Abc_Var2Lit(VarCecPar + i, 1) );
    }
    for  nIters = 0; nIters < (1 << nVars); nIters++  {
        // check if these parameters solve the problem
        var status int32  = sat_solver_solve( pSatCec, Vec_IntArray(vLits), Vec_IntLimit(vLits), 0, 0, 0, 0 );
        if status == l_False { // solution found
            break; }
        assert(status == l_True, "status == l_True");

        if fVerbose != 0 {
            fmt.Printf("Iter %3d : ", nIters);
            for  i = 0; i < nPars; i++  {
                fmt.Printf("%d", !Abc_LitIsCompl(Vec_IntEntry(vLits, i)));
            }
            fmt.Printf("    ");
        }

        Vec_IntClear( vLits );
        // create new QBF variables
        for  i = 0; i < nVars + nStrs; i++  {
            pVarsQbf[i] = VarQbfFree++;
        }
        // set their values
        Vec_IntForEachEntry( vDivSet, iVar, i )
        {
            iLit = Abc_Var2Lit( pVarsQbf[i], !sat_solver_var_value(pSatCec, iVar) );
            status = sat_solver_addclause( pSatQbf, &iLit, &iLit + 1 );
            assert((status != 0), "status");
            if fVerbose != 0 {
                fmt.Printf("%d", sat_solver_var_value(pSatCec, iVar)); }
        }
        iLit = Abc_Var2Lit( pVarsQbf[nVars], sat_solver_var_value(pSatCec, VarCecOut) );
        status = sat_solver_addclause( pSatQbf, &iLit, &iLit + 1 );
        assert((status != 0), "status");
        if fVerbose != 0 {
            fmt.Printf(" %d\n", !sat_solver_var_value(pSatCec, VarCecOut)); }
        // add clauses to the QBF problem
        if !(Sbd_ProblemAddClauses( pSatQbf, nVars, nStrs, pVarsQbf, pStr0 ) != 0) {
            break; } // solution does not exist
        // check if solution still exists
        status = sat_solver_solve( pSatQbf, nil, nil, 0, 0, 0, 0 );
        if status == l_False { // solution does not exist
            break; }
        assert(status == l_True, "status == l_True"); 
        // find the new values of parameters
        assert(Vec_IntSize(vLits) == 0, "Vec_IntSize(vLits) == 0");
        for  i = 0; i < nPars; i++  {
            Vec_IntPush( vLits, Abc_Var2Lit(VarCecPar + i, !sat_solver_var_value(pSatQbf, VarQbfPar + i)) );
        }
    }
    if Vec_IntSize(vLits) > 0 {
        //Sbd_ProblemPrintSolution( nStrs, pStr0, vLits );
        Sbd_ProblemCollectSolution( nStrs, pStr0, vLits );
        RetValue = 1;
    }
    sat_solver_delete( pSatCec );
    sat_solver_delete( pSatQbf );
    Vec_IntFree( vLits );

    if fVerbose != 0 {
        Abc_PrintTime( 1, "Time", Abc_Clock() - clk ); }
    return RetValue;
}


////////////////////////////////////////////////////////////////////////
///                       END OF FILE                                ///
////////////////////////////////////////////////////////////////////////


//ABC_NAMESPACE_IMPL_END

