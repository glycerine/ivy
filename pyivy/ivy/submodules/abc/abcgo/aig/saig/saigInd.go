package saig

import (
	"fmt"
	"time"
)

/**CFile****************************************************************

  FileName    [saigLoc.c]

  SystemName  [ABC: Logic synthesis and verification system.]

  PackageName [Sequential AIG package.]

  Synopsis    [K-step induction for one property only.]

  Author      [Alan Mishchenko]
  
  Affiliation [UC Berkeley]

  Date        [Ver. 1.0. Started - June 20, 2005.]

  Revision    [$Id: saigLoc.c,v 1.00 2005/06/20 00:00:00 alanmi Exp $]

***********************************************************************/

//#include "saig.h"
//#include "sat/cnf/cnf.h"
//#include "sat/bsat/satSolver.h"

//ABC_NAMESPACE_IMPL_START


////////////////////////////////////////////////////////////////////////
///                        DECLARATIONS                              ///
////////////////////////////////////////////////////////////////////////

////////////////////////////////////////////////////////////////////////
///                     FUNCTION DEFINITIONS                         ///
////////////////////////////////////////////////////////////////////////

/**Function*************************************************************

  Synopsis    [Returns 1 if two state are equal.]

  Description [Array vState contains indexes of CNF variables for each
  flop in the first N time frames (0 < i < k, i < N, k < N).]
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Saig_ManStatesAreEqual(pSat *sat_solver, vState *Vec_Int_t, nRegs int32, i int32, k int32) int32 {
    var pStateI *int32  = (int *)Vec_IntArray(vState) + nRegs * i;
    var pStateK *int32  = (int *)Vec_IntArray(vState) + nRegs * k;
    var v int32;
    assert((i != 0) && (k != 0) && i < k, "i && k && i < k");
    assert(nRegs * k <= Vec_IntSize(vState), "nRegs * k <= Vec_IntSize(vState)");
    // check if the states are available
    for  v = 0; v < nRegs; v++  {
        if pStateI[v] >= 0 && pStateK[v] == -1 {
            return 0; }
    }
/*
    printf( "\nchecking uniqueness\n" );
    printf( "%3d : ", i );
    for ( v = 0; v < nRegs; v++ )
        printf( "%d", sat_solver_var_value(pSat, pStateI[v]) );
    printf( "\n" );

    printf( "%3d : ", k );
    for ( v = 0; v < nRegs; v++ )
        printf( "%d", sat_solver_var_value(pSat, pStateK[v]) );
    printf( "\n" );
*/
    for  v = 0; v < nRegs; v++  {
        if pStateI[v] >= 0 {
            if sat_solver_var_value(pSat, pStateI[v]) != sat_solver_var_value(pSat, pStateK[v]) {
                return 0; }
        }
    }
    return 1;
}

/**Function*************************************************************

  Synopsis    [Add uniqueness constraint.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Saig_ManAddUniqueness(pSat *sat_solver, vState *Vec_Int_t, nRegs int32, i int32, k int32, pnSatVarNum *int32, pnClauses *int32, fVerbose int32) int32 {
    var pStateI *int32  = (int *)Vec_IntArray(vState) + nRegs * i;
    var pStateK *int32  = (int *)Vec_IntArray(vState) + nRegs * k;
    var v int32;
    var iVars int32;
    var nSatVarsOld int32;
    var RetValue int32;
    var pClause *int32;
    assert((i != 0) && (k != 0) && i < k, "i && k && i < k");
    assert(nRegs * k <= Vec_IntSize(vState), "nRegs * k <= Vec_IntSize(vState)");
    // check if the states are available
    for  v = 0; v < nRegs; v++  {
        if pStateI[v] >= 0 && pStateK[v] == -1 {
            if fVerbose != 0 {
                fmt.Printf("Cannot constrain an incomplete state.\n"); }
            return 0;
        }
    }
    // add XORs
    nSatVarsOld = *pnSatVarNum;
    for  v = 0; v < nRegs; v++  {
        if pStateI[v] >= 0 {
            *pnClauses += 4;
            RetValue = Cnf_DataAddXorClause( pSat, pStateI[v], pStateK[v], (*pnSatVarNum)++ );
            if RetValue == 0 {
                if fVerbose != 0 {
                    fmt.Printf("SAT solver became UNSAT after adding a uniqueness constraint.\n"); }
                return 1;
            }
        }
    }
    // add OR clause
    (*pnClauses)++;
    iVars = 0;
    // inlining of macro ABC_ALLOC
    pClause = ((int *) malloc(sizeof(int) * (size_t)(nRegs)));
    for  v = nSatVarsOld; v < *pnSatVarNum; v++  {
        pClause[iVars++] = toLitCond( v, 0 );
    }
    assert(iVars <= nRegs, "iVars <= nRegs");
    RetValue = sat_solver_addclause( pSat, pClause, pClause + iVars );
    // inlining of macro ABC_FREE
    ((pClause) ? (free((char *) (pClause)), (pClause) = 0) : 0);
    if RetValue == 0 {
        if fVerbose != 0 {
            fmt.Printf("SAT solver became UNSAT after adding a uniqueness constraint.\n"); }
        return 1;
    }
    return 0;
}

/**Function*************************************************************

  Synopsis    [Performs induction by unrolling timeframes backward.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Saig_ManInduction(p *Aig_Man_t, nTimeOut int32, nFramesMax int32, nConfMax int32, fUnique int32, fUniqueAll int32, fGetCex int32, fVerbose int32, fVeryVerbose int32) int32 {
    var pSat *sat_solver;
    var pAigPart *Aig_Man_t  = nil;
    var pCnfPart *Cnf_Dat_t  = nil;
    var vTopVarNums *Vec_Int_t;
    var vState *Vec_Int_t;
    var vTopVarIds *Vec_Int_t = nil;
    var vTop, vBot *Vec_Ptr_t;
    var pObjPi, pObjPiCopy, pObjPo *Aig_Obj_t;
    var i int32;
    var k int32;
    var f int32;
    var Lits [2]int32;
    var status int32 = -1;
    var RetValue int32;
    var nSatVarNum int32;
    var nConfPrev int32;
    var nOldSize, iReg, iLast, fAdded, nConstrs, nClauses int32;
    var clk abctime;
    var nTimeToStop abctime;
    if nTimeOut {
        nTimeToStop = nTimeOut * int64(time.Second) + Abc_Clock();
    } else {
        nTimeToStop = 0;
    }
    assert(fUnique == 0 || fUniqueAll == 0, "fUnique == 0 || fUniqueAll == 0");
    assert(Saig_ManPoNum(p) == 1, "Saig_ManPoNum(p) == 1");
    Aig_ManSetCioIds( p );

    // start the top by including the PO
    vBot = Vec_PtrAlloc( 100 );
    vTop = Vec_PtrAlloc( 100 );
    vState = Vec_IntAlloc( 1000 );
    Vec_PtrPush( vTop, Aig_ManCo(p, 0) );
    // start the array of CNF variables
    vTopVarNums = Vec_IntAlloc( 100 );
    // start the solver
    pSat = sat_solver_new();
    sat_solver_setnvars( pSat, 1000 );

    // set runtime limit
    if nTimeToStop {
        sat_solver_set_runtime_limit( pSat, nTimeToStop ); }

    // iterate backward unrolling
    RetValue = -1;
    nSatVarNum = 0;
    if fVerbose != 0 {
        fmt.Printf("Induction parameters: FramesMax = %5d. ConflictMax = %6d.\n", nFramesMax, nConfMax); }
    for  f = 0; ; f++  { 
        if f > 0 {
            Aig_ManStop( pAigPart );
            Cnf_DataFree( pCnfPart );
        }
        clk = Abc_Clock();
        // get the bottom
        Aig_SupportNodes( p, (Aig_Obj_t **)Vec_PtrArray(vTop), Vec_PtrSize(vTop), vBot );
        // derive AIG for the part between top and bottom
        pAigPart = Aig_ManDupSimpleDfsPart( p, vBot, vTop );
        // convert it into CNF
        pCnfPart = Cnf_Derive( pAigPart, Aig_ManCoNum(pAigPart) );
        Cnf_DataLift( pCnfPart, nSatVarNum );
        nSatVarNum += pCnfPart.nVars;
        nClauses   += pCnfPart.nClauses;

        // remember top frame var IDs
        if (fGetCex != 0) && vTopVarIds == nil {
            vTopVarIds = Vec_IntStartFull( Aig_ManCiNum(p) );
            // inlining of macro Aig_ManForEachCi
            // inlining of macro Vec_PtrForEachEntry
            for  i = 0; (i < Vec_PtrSize(p.vCis)) && (((pObjPi) = (Aig_Obj_t *)Vec_PtrEntry(p.vCis, i)), 1); i++  {
                if pObjPi.pData == nil {
                    continue; }
                pObjPiCopy = (Aig_Obj_t *)pObjPi.pData;
                assert(Aig_ObjIsCi(pObjPiCopy), "Aig_ObjIsCi(pObjPiCopy)");
                if Saig_ObjIsPi(p, pObjPi) {
                    Vec_IntWriteEntry( vTopVarIds, Aig_ObjCioId(pObjPi) + Saig_ManRegNum(p), pCnfPart.pVarNums[Aig_ObjId(pObjPiCopy)] );
                } else if Saig_ObjIsLo(p, pObjPi) {
                    Vec_IntWriteEntry( vTopVarIds, Aig_ObjCioId(pObjPi) - Saig_ManPiNum(p), pCnfPart.pVarNums[Aig_ObjId(pObjPiCopy)] );
                } else { assert(false, "0"); }
            }
        }

        // stitch variables of top and bot
        assert(Aig_ManCoNum(pAigPart)-1 == Vec_IntSize(vTopVarNums), "Aig_ManCoNum(pAigPart)-1 == Vec_IntSize(vTopVarNums)");
        // inlining of macro Aig_ManForEachCo
        // inlining of macro Vec_PtrForEachEntry
        for  i = 0; (i < Vec_PtrSize(pAigPart.vCos)) && (((pObjPo) = (Aig_Obj_t *)Vec_PtrEntry(pAigPart.vCos, i)), 1); i++  {
            if i == 0 {
                // do not perform inductive strengthening
//                if ( f > 0 )
//                    continue;
                // add topmost literal
                Lits[0] = toLitCond( pCnfPart.pVarNums[pObjPo.Id], f>0 );
                if !sat_solver_addclause( pSat, Lits, Lits+1 ) {
                    assert(false, "0"); }
                nClauses++;
                continue;
            }
            Lits[0] = toLitCond( Vec_IntEntry(vTopVarNums, i-1), 0 );
            Lits[1] = toLitCond( pCnfPart.pVarNums[pObjPo.Id], 1 );
            if !sat_solver_addclause( pSat, Lits, Lits+2 ) {
                assert(false, "0"); }
            Lits[0] = toLitCond( Vec_IntEntry(vTopVarNums, i-1), 1 );
            Lits[1] = toLitCond( pCnfPart.pVarNums[pObjPo.Id], 0 );
            if !sat_solver_addclause( pSat, Lits, Lits+2 ) {
                assert(false, "0"); }
            nClauses += 2;
        }
        // add CNF to the SAT solver
        for  i = 0; i < pCnfPart.nClauses; i++  {
            if !sat_solver_addclause( pSat, pCnfPart.pClauses[i], pCnfPart.pClauses[i+1] ) {
                break; }
        }
        if i < pCnfPart.nClauses {
//            printf( "SAT solver became UNSAT after adding clauses.\n" );
            RetValue = 1;
            break;
        }

        // create new set of POs to derive new top
        Vec_PtrClear( vTop );
        Vec_PtrPush( vTop, Aig_ManCo(p, 0) );
        Vec_IntClear( vTopVarNums );
        nOldSize = Vec_IntSize(vState);
        Vec_IntFillExtra( vState, nOldSize + Aig_ManRegNum(p), -1 );
        // inlining of macro Vec_PtrForEachEntry
        for  i = 0; (i < Vec_PtrSize(vBot)) && (((pObjPi) = (Aig_Obj_t *)Vec_PtrEntry(vBot, i)), 1); i++  {
            assert(Aig_ObjIsCi(pObjPi), "Aig_ObjIsCi(pObjPi)");
            if Saig_ObjIsLo(p, pObjPi) {
                pObjPiCopy = (Aig_Obj_t *)pObjPi.pData;
                assert(pObjPiCopy != nil, "pObjPiCopy != nil");
                Vec_PtrPush( vTop, Saig_ObjLoToLi(p, pObjPi) );
                Vec_IntPush( vTopVarNums, pCnfPart.pVarNums[pObjPiCopy.Id] );

                iReg = pObjPi.CioId - Saig_ManPiNum(p);
                assert(iReg >= 0 && iReg < Aig_ManRegNum(p), "iReg >= 0 && iReg < Aig_ManRegNum(p)");
                Vec_IntWriteEntry( vState, nOldSize+iReg, pCnfPart.pVarNums[pObjPiCopy.Id] );
            }
        } 
        assert(Vec_IntSize(vState)%Aig_ManRegNum(p) == 0, "Vec_IntSize(vState)%Aig_ManRegNum(p) == 0");
        iLast = Vec_IntSize(vState)/Aig_ManRegNum(p);
        if fUniqueAll != 0 {
            for  i = 1; i < iLast-1; i++  {
                nConstrs++;
                if fVeryVerbose != 0 {
                    fmt.Printf("Adding constaint for state %2d and state %2d.\n", i, iLast-1); }
                if Saig_ManAddUniqueness( pSat, vState, Aig_ManRegNum(p), i, iLast-1, &nSatVarNum, &nClauses, fVerbose ) != 0 {
                    break; }
            }
            if i < iLast-1 {
                RetValue = 1;
                break;
            }
        }

nextrun:
        fAdded = 0;
        // run the SAT solver
        nConfPrev = pSat.stats.conflicts;
        status = sat_solver_solve( pSat, nil, nil, (ABC_INT64_T)nConfMax, 0, 0, 0 );
        if fVerbose != 0 {
            fmt.Printf("Frame %4d : PI =%5d. PO =%5d. AIG =%5d. Var =%7d. Clau =%7d. Conf =%7d. ", f, Aig_ManCiNum(pAigPart), Aig_ManCoNum(pAigPart), Aig_ManNodeNum(pAigPart), nSatVarNum, nClauses, int32(pSat.stats.conflicts)-nConfPrev);
            // inlining of macro ABC_PRT
            Abc_Print(1, "%s =", ("Time"));
Abc_Print(1, "%9.2f sec\n", 1.0*(float64((Abc_Clock() - clk)))/(float64(int64(time.Second))));;
        }
        if status == l_Undef {
            break; }
        if status == l_False {
            RetValue = 1;
            break;
        }
        assert(status == l_True, "status == l_True");
        // the problem is SAT - add more clauses
        if fVeryVerbose != 0 {
            // inlining of macro Vec_IntForEachEntry
            for  i = 0; (i < Vec_IntSize(vState)) && (((iReg) = Vec_IntEntry(vState, i)), 1); i++  {
                if (i != 0) && (i % Aig_ManRegNum(p)) == 0 {
                    fmt.Printf("\n"); }
                if (i % Aig_ManRegNum(p)) == 0 {
                    fmt.Printf("       State %3d : ", i/Aig_ManRegNum(p)); }
                fmt.Printf("%c", (iReg >= 0) ? ('0' + sat_solver_var_value(pSat, iReg)) : 'x');
            }
            fmt.Printf("\n");
        }
        if (nFramesMax != 0) && f == nFramesMax - 1 {
            // derive counter-example
            assert(status == l_True, "status == l_True");
            if fGetCex != 0 {
                var VarNum, iBit int32;
                var pCex *Abc_Cex_t  = Abc_CexAlloc( Aig_ManRegNum(p)-1, Saig_ManPiNum(p), 1 );
                pCex.iFrame = 0;
                pCex.iPo = 0;
                // inlining of macro Vec_IntForEachEntryStart
                for  i = 1; (i < Vec_IntSize(vTopVarIds)) && (((VarNum) = Vec_IntEntry(vTopVarIds, i)), 1); i++  {
                    if VarNum >= 0 && sat_solver_var_value( pSat, VarNum ) {
                        Abc_InfoSetBit( pCex.pData, iBit ); }
                    iBit++;
                }
                assert(iBit == pCex.nBits, "iBit == pCex.nBits");
                Abc_CexFree( p.pSeqModel );
                p.pSeqModel = pCex;
            }
            break;
        }
        if fUnique != 0 {
            for  i = 1; i < iLast; i++  {
                for  k = i+1; k < iLast; k++  {
                    if !(Saig_ManStatesAreEqual( pSat, vState, Aig_ManRegNum(p), i, k ) != 0) {
                        continue; }
                    nConstrs++;
                    fAdded = 1;
                    if fVeryVerbose != 0 {
                        fmt.Printf("Adding constaint for state %2d and state %2d.\n", i, k); }
                    if Saig_ManAddUniqueness( pSat, vState, Aig_ManRegNum(p), i, k, &nSatVarNum, &nClauses, fVerbose ) != 0 {
                        break; }
                }
                if k < iLast {
                    break; }
            }
            if i < iLast {
                RetValue = 1;
                break;
            }
        }
        if fAdded {
            goto nextrun; }
    }
    if fVerbose != 0 {
        if nTimeToStop && Abc_Clock() >= nTimeToStop {
            fmt.Printf("Timeout (%d sec) was reached during iteration %d.\n", nTimeOut, f+1);
        } else if status == l_Undef {
            fmt.Printf("Conflict limit (%d) was reached during iteration %d.\n", nConfMax, f+1);
        } else if (fUnique != 0) || (fUniqueAll != 0) {
            fmt.Printf("Completed %d iterations and added %d uniqueness constraints.\n", f+1, nConstrs);
        } else {
            fmt.Printf("Completed %d iterations.\n", f+1); }
    }
    // cleanup
    sat_solver_delete( pSat );
    Aig_ManStop( pAigPart );
    Cnf_DataFree( pCnfPart );
    Vec_IntFree( vTopVarNums );
    Vec_PtrFree( vTop );
    Vec_PtrFree( vBot );
    Vec_IntFree( vState );
    Vec_IntFreeP( &vTopVarIds );
    return RetValue;
}


////////////////////////////////////////////////////////////////////////
///                       END OF FILE                                ///
////////////////////////////////////////////////////////////////////////


//ABC_NAMESPACE_IMPL_END

