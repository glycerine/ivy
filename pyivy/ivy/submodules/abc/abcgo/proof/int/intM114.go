package int

/**CFile****************************************************************

  FileName    [intM114.c]

  SystemName  [ABC: Logic synthesis and verification system.]

  PackageName [Interpolation engine.]

  Synopsis    [Intepolation using ABC's proof generator added to MiniSat-1.14c.]

  Author      [Alan Mishchenko]
  
  Affiliation [UC Berkeley]

  Date        [Ver. 1.0. Started - June 24, 2008.]

  Revision    [$Id: intM114.c,v 1.00 2005/06/20 00:00:00 alanmi Exp $]

***********************************************************************/

//#include "intInt.h"

//ABC_NAMESPACE_IMPL_START


////////////////////////////////////////////////////////////////////////
///                        DECLARATIONS                              ///
////////////////////////////////////////////////////////////////////////

////////////////////////////////////////////////////////////////////////
///                     FUNCTION DEFINITIONS                         ///
////////////////////////////////////////////////////////////////////////

/**Function*************************************************************

  Synopsis    [Returns the SAT solver for one interpolation run.]

  Description [pInter is the previous interpolant. pAig is one time frame.
  pFrames is the unrolled time frames.]
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Inter_ManDeriveSatSolver(pInter *Aig_Man_t, pCnfInter *Cnf_Dat_t, pAig *Aig_Man_t, pCnfAig *Cnf_Dat_t, pFrames *Aig_Man_t, pCnfFrames *Cnf_Dat_t, vVarsAB *Vec_Int_t, fUseBackward int32) *sat_solver {
    var pSat *sat_solver;
    var pObj, pObj2 *Aig_Obj_t;
    var i int32;
    var Lits [2]int32;

//Aig_ManDumpBlif( pInter,  "out_inter.blif", NULL, NULL );
//Aig_ManDumpBlif( pAig,    "out_aig.blif", NULL, NULL );
//Aig_ManDumpBlif( pFrames, "out_frames.blif", NULL, NULL );

    // sanity checks
    assert(Aig_ManRegNum(pInter) == 0, "Aig_ManRegNum(pInter) == 0");
    assert(Aig_ManRegNum(pAig) > 0, "Aig_ManRegNum(pAig) > 0");
    assert(Aig_ManRegNum(pFrames) == 0, "Aig_ManRegNum(pFrames) == 0");
    assert(Aig_ManCoNum(pInter) == 1, "Aig_ManCoNum(pInter) == 1");
    assert(Aig_ManCoNum(pFrames) == fUseBackward? Saig_ManRegNum(pAig) : 1, "Aig_ManCoNum(pFrames) == fUseBackward? Saig_ManRegNum(pAig) : 1");
    assert((fUseBackward != 0) || Aig_ManCiNum(pInter) == Aig_ManRegNum(pAig), "fUseBackward || Aig_ManCiNum(pInter) == Aig_ManRegNum(pAig)");
//    assert( (Aig_ManCiNum(pFrames) - Aig_ManRegNum(pAig)) % Saig_ManPiNum(pAig) == 0 );

    // prepare CNFs
    Cnf_DataLift( pCnfAig,   pCnfFrames.nVars );
    Cnf_DataLift( pCnfInter, pCnfFrames.nVars + pCnfAig.nVars );

    // start the solver
    pSat = sat_solver_new();
    sat_solver_store_alloc( pSat );
    sat_solver_setnvars( pSat, pCnfInter.nVars + pCnfAig.nVars + pCnfFrames.nVars );

    // add clauses of A
    // interpolant
    for  i = 0; i < pCnfInter.nClauses; i++  {
        if !sat_solver_addclause( pSat, pCnfInter.pClauses[i], pCnfInter.pClauses[i+1] ) {
            sat_solver_delete( pSat );
            // return clauses to the original state
            Cnf_DataLift( pCnfAig, -pCnfFrames.nVars );
            Cnf_DataLift( pCnfInter, -pCnfFrames.nVars -pCnfAig.nVars );
            return nil;
        }
    }
    // connector clauses
    if fUseBackward != 0 {
        // inlining of macro Saig_ManForEachLi
        for  i = 0; (i < Saig_ManRegNum(pAig)) && (((pObj2) = (Aig_Obj_t *)Vec_PtrEntry(pAig.vCos, i+Saig_ManPoNum(pAig))), 1); i++  {
            if Saig_ManRegNum(pAig) == Aig_ManCiNum(pInter) {
                pObj = Aig_ManCi( pInter, i );
            } else {
                assert(Aig_ManCiNum(pAig) == Aig_ManCiNum(pInter), "Aig_ManCiNum(pAig) == Aig_ManCiNum(pInter)");
                pObj = Aig_ManCi( pInter, Aig_ManCiNum(pAig)-Saig_ManRegNum(pAig) + i );
            }

            Lits[0] = toLitCond( pCnfInter.pVarNums[pObj.Id], 0 );
            Lits[1] = toLitCond( pCnfAig.pVarNums[pObj2.Id], 1 );
            if !sat_solver_addclause( pSat, Lits, Lits+2 ) {
                assert(false, "0"); }
            Lits[0] = toLitCond( pCnfInter.pVarNums[pObj.Id], 1 );
            Lits[1] = toLitCond( pCnfAig.pVarNums[pObj2.Id], 0 );
            if !sat_solver_addclause( pSat, Lits, Lits+2 ) {
                assert(false, "0"); }
        }
    
    } else {
        // inlining of macro Aig_ManForEachCi
        // inlining of macro Vec_PtrForEachEntry
        for  i = 0; (i < Vec_PtrSize(pInter.vCis)) && (((pObj) = (Aig_Obj_t *)Vec_PtrEntry(pInter.vCis, i)), 1); i++  {
            pObj2 = Saig_ManLo( pAig, i );

            Lits[0] = toLitCond( pCnfInter.pVarNums[pObj.Id], 0 );
            Lits[1] = toLitCond( pCnfAig.pVarNums[pObj2.Id], 1 );
            if !sat_solver_addclause( pSat, Lits, Lits+2 ) {
                assert(false, "0"); }
            Lits[0] = toLitCond( pCnfInter.pVarNums[pObj.Id], 1 );
            Lits[1] = toLitCond( pCnfAig.pVarNums[pObj2.Id], 0 );
            if !sat_solver_addclause( pSat, Lits, Lits+2 ) {
                assert(false, "0"); }
        }
    }
    // one timeframe
    for  i = 0; i < pCnfAig.nClauses; i++  {
        if !sat_solver_addclause( pSat, pCnfAig.pClauses[i], pCnfAig.pClauses[i+1] ) {
            assert(false, "0"); }
    }
    // connector clauses
    Vec_IntClear( vVarsAB );
    if fUseBackward != 0 {
        // inlining of macro Aig_ManForEachCo
        // inlining of macro Vec_PtrForEachEntry
        for  i = 0; (i < Vec_PtrSize(pFrames.vCos)) && (((pObj) = (Aig_Obj_t *)Vec_PtrEntry(pFrames.vCos, i)), 1); i++  {
            assert(pCnfFrames.pVarNums[pObj.Id] >= 0, "pCnfFrames.pVarNums[pObj.Id] >= 0");
            Vec_IntPush( vVarsAB, pCnfFrames.pVarNums[pObj.Id] );

            pObj2 = Saig_ManLo( pAig, i );
            Lits[0] = toLitCond( pCnfFrames.pVarNums[pObj.Id], 0 );
            Lits[1] = toLitCond( pCnfAig.pVarNums[pObj2.Id], 1 );
            if !sat_solver_addclause( pSat, Lits, Lits+2 ) {
                assert(false, "0"); }
            Lits[0] = toLitCond( pCnfFrames.pVarNums[pObj.Id], 1 );
            Lits[1] = toLitCond( pCnfAig.pVarNums[pObj2.Id], 0 );
            if !sat_solver_addclause( pSat, Lits, Lits+2 ) {
                assert(false, "0"); }
        }
    
    } else {
        // inlining of macro Aig_ManForEachCi
        // inlining of macro Vec_PtrForEachEntry
        for  i = 0; (i < Vec_PtrSize(pFrames.vCis)) && (((pObj) = (Aig_Obj_t *)Vec_PtrEntry(pFrames.vCis, i)), 1); i++  {
            if i == Aig_ManRegNum(pAig) {
                break; }
            Vec_IntPush( vVarsAB, pCnfFrames.pVarNums[pObj.Id] );

            pObj2 = Saig_ManLi( pAig, i );
            Lits[0] = toLitCond( pCnfFrames.pVarNums[pObj.Id], 0 );
            Lits[1] = toLitCond( pCnfAig.pVarNums[pObj2.Id], 1 );
            if !sat_solver_addclause( pSat, Lits, Lits+2 ) {
                assert(false, "0"); }
            Lits[0] = toLitCond( pCnfFrames.pVarNums[pObj.Id], 1 );
            Lits[1] = toLitCond( pCnfAig.pVarNums[pObj2.Id], 0 );
            if !sat_solver_addclause( pSat, Lits, Lits+2 ) {
                assert(false, "0"); }
        }
    }
    // add clauses of B
    sat_solver_store_mark_clauses_a( pSat );
    for  i = 0; i < pCnfFrames.nClauses; i++  {
        if !sat_solver_addclause( pSat, pCnfFrames.pClauses[i], pCnfFrames.pClauses[i+1] ) {
            pSat.fSolved = 1;
            break;
        }
    }
    sat_solver_store_mark_roots( pSat );
    // return clauses to the original state
    Cnf_DataLift( pCnfAig, -pCnfFrames.nVars );
    Cnf_DataLift( pCnfInter, -pCnfFrames.nVars -pCnfAig.nVars );
    return pSat;
}

/**Function*************************************************************

  Synopsis    [Performs one SAT run with interpolation.]

  Description [Returns 1 if proven. 0 if failed. -1 if undecided.]
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Inter_ManPerformOneStep(p *Inter_Man_t, fUseBias int32, fUseBackward int32, nTimeNewOut abctime) int32 {
    var pSat *sat_solver;
    var pSatCnf any  = nil;
    var pManInterA *Inta_Man_t; 
//    Intb_Man_t * pManInterB; 
    var pGlobalVars *int32;
    var status, RetValue int32;
    var i, Var int32;
    var clk abctime;
//    assert( p->pInterNew == NULL );

    // derive the SAT solver
    pSat = Inter_ManDeriveSatSolver( p.pInter, p.pCnfInter, p.pAigTrans, p.pCnfAig, p.pFrames, p.pCnfFrames, p.vVarsAB, fUseBackward );
    if pSat == nil {
        p.pInterNew = nil;
        return 1;
    }

    // set runtime limit
    if nTimeNewOut {
        sat_solver_set_runtime_limit( pSat, nTimeNewOut ); }

    // collect global variables
    // inlining of macro ABC_CALLOC
    pGlobalVars = ((int *) calloc((size_t)(sat_solver_nvars(pSat)), sizeof(int)));
    // inlining of macro Vec_IntForEachEntry
    for  i = 0; (i < Vec_IntSize(p.vVarsAB)) && (((Var) = Vec_IntEntry(p.vVarsAB, i)), 1); i++  {
        pGlobalVars[Var] = 1;
    }
    if fUseBias != 0 {
        pSat.pGlobalVars = pGlobalVars;
    } else {
        pSat.pGlobalVars = nil;
    }

    // solve the problem
clk = Abc_Clock();
    status = sat_solver_solve( pSat, nil, nil, (ABC_INT64_T)p.nConfLimit, (ABC_INT64_T)0, (ABC_INT64_T)0, (ABC_INT64_T)0 );
    p.nConfCur = pSat.stats.conflicts;
p.timeSat += Abc_Clock() - clk;

    pSat.pGlobalVars = nil;
    pGlobalVars = nil;
    if status == l_False {
        pSatCnf = sat_solver_store_release( pSat );
        RetValue = 1;
    
    } else if status == l_True {
        RetValue = 0;
     
    } else {
        RetValue = -1;
    }
    sat_solver_delete( pSat );
    if pSatCnf == nil {
        return RetValue; }

    // create the resulting manager
clk = Abc_Clock();
/*
    if ( !fUseIp )
    {
        pManInterA = Inta_ManAlloc();
        p->pInterNew = Inta_ManInterpolate( pManInterA, pSatCnf, p->vVarsAB, 0 );
        Inta_ManFree( pManInterA );
    }
    else
    {
        Aig_Man_t * pInterNew2;
        int RetValue;

        pManInterA = Inta_ManAlloc();
        p->pInterNew = Inta_ManInterpolate( pManInterA, pSatCnf, p->vVarsAB, 0 );
//        Inter_ManVerifyInterpolant1( pManInterA, pSatCnf, p->pInterNew );
        Inta_ManFree( pManInterA );

        pManInterB = Intb_ManAlloc();
        pInterNew2 = Intb_ManInterpolate( pManInterB, pSatCnf, p->vVarsAB, 0 );
        Inter_ManVerifyInterpolant2( pManInterB, pSatCnf, pInterNew2 );
        Intb_ManFree( pManInterB );

        // check relationship
        RetValue = Inter_ManCheckEquivalence( pInterNew2, p->pInterNew );
        if ( RetValue )
            printf( "Equivalence \"Ip == Im\" holds\n" );
        else
        {
//            printf( "Equivalence \"Ip == Im\" does not hold\n" );
            RetValue = Inter_ManCheckContainment( pInterNew2, p->pInterNew );
            if ( RetValue )
                printf( "Containment \"Ip -> Im\" holds\n" );
            else
                printf( "Containment \"Ip -> Im\" does not hold\n" );

            RetValue = Inter_ManCheckContainment( p->pInterNew, pInterNew2 );
            if ( RetValue )
                printf( "Containment \"Im -> Ip\" holds\n" );
            else
                printf( "Containment \"Im -> Ip\" does not hold\n" );
        }

        Aig_ManStop( pInterNew2 );
    }
*/

    pManInterA = Inta_ManAlloc();
    p.pInterNew = (Aig_Man_t *)Inta_ManInterpolate( pManInterA, (Sto_Man_t *)pSatCnf, nTimeNewOut, p.vVarsAB, 0 );
    Inta_ManFree( pManInterA );

p.timeInt += Abc_Clock() - clk;
    Sto_ManFree( (Sto_Man_t *)pSatCnf );
    if p.pInterNew == nil {
        RetValue = -1; }
    return RetValue;
}

////////////////////////////////////////////////////////////////////////
///                       END OF FILE                                ///
////////////////////////////////////////////////////////////////////////


//ABC_NAMESPACE_IMPL_END

