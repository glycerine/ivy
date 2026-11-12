package res

/**CFile****************************************************************

  FileName    [resSat.c]

  SystemName  [ABC: Logic synthesis and verification system.]

  PackageName [Resynthesis package.]

  Synopsis    [Interface with the SAT solver.]

  Author      [Alan Mishchenko]
  
  Affiliation [UC Berkeley]

  Date        [Ver. 1.0. Started - January 15, 2007.]

  Revision    [$Id: resSat.c,v 1.00 2007/01/15 00:00:00 alanmi Exp $]

***********************************************************************/

//#include "base/abc/abc.h"
//#include "resInt.h"
//#include "aig/hop/hop.h"
//#include "sat/bsat/satSolver.h"

//ABC_NAMESPACE_IMPL_START


////////////////////////////////////////////////////////////////////////
///                        DECLARATIONS                              ///
////////////////////////////////////////////////////////////////////////

// extern int Res_SatAddConst1( sat_solver * pSat, int iVar, int fCompl );

// extern int Res_SatAddEqual( sat_solver * pSat, int iVar0, int iVar1, int fCompl );

// extern int Res_SatAddAnd( sat_solver * pSat, int iVar, int iVar0, int iVar1, int fCompl0, int fCompl1 );


////////////////////////////////////////////////////////////////////////
///                     FUNCTION DEFINITIONS                         ///
////////////////////////////////////////////////////////////////////////

/**Function*************************************************************

  Synopsis    [Loads AIG into the SAT solver for checking resubstitution.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Res_SatProveUnsat(pAig *Abc_Ntk_t, vFanins *Vec_Ptr_t) any {
    var pCnf any  = nil;
    var pSat *sat_solver;
    var vNodes *Vec_Ptr_t;
    var pObj *Abc_Obj_t;
    var i, nNodes, status int32;

    // make sure fanins contain POs of the AIG
    pObj = (Abc_Obj_t *)Vec_PtrEntry( vFanins, 0 );
    assert(pObj.pNtk == pAig && Abc_ObjIsPo(pObj), "pObj.pNtk == pAig && Abc_ObjIsPo(pObj)");

    // collect reachable nodes
    vNodes = Abc_NtkDfsNodes( pAig, (Abc_Obj_t **)vFanins.pArray, vFanins.nSize );

    // assign unique numbers to each node
    nNodes = 0;
    Abc_AigConst1(pAig).pCopy = (Abc_Obj_t *)(ABC_PTRUINT_T)nNodes++;
    // inlining of macro Abc_NtkForEachPi
    for i = 0; (i < Abc_NtkPiNum(pAig)); i++ { pObj = Abc_NtkPi(pAig, i);
        pObj.pCopy = (Abc_Obj_t *)(ABC_PTRUINT_T)nNodes++;
    }
    // inlining of macro Vec_PtrForEachEntry
    for i = 0; (i < Vec_PtrSize(vNodes)); i++ { pObj = Vec_PtrEntry(vNodes, i);
        pObj.pCopy = (Abc_Obj_t *)(ABC_PTRUINT_T)nNodes++;
    }
    // inlining of macro Vec_PtrForEachEntry
    for i = 0; (i < Vec_PtrSize(vFanins)); i++ { pObj = Vec_PtrEntry(vFanins, i); // useful POs
        pObj.pCopy = (Abc_Obj_t *)(ABC_PTRUINT_T)nNodes++;
    }

    // start the solver
    pSat = sat_solver_new();
    sat_solver_store_alloc( pSat );

    // add clause for the constant node
    Res_SatAddConst1( pSat, int32((ABC_PTRUINT_T))Abc_AigConst1(pAig).pCopy, 0 );
    // add clauses for AND gates
    // inlining of macro Vec_PtrForEachEntry
    for i = 0; (i < Vec_PtrSize(vNodes)); i++ { pObj = Vec_PtrEntry(vNodes, i);
        Res_SatAddAnd( pSat, int32((ABC_PTRUINT_T))pObj.pCopy, 
            int32((ABC_PTRUINT_T))Abc_ObjFanin0(pObj).pCopy, int32((ABC_PTRUINT_T))Abc_ObjFanin1(pObj).pCopy, Abc_ObjFaninC0(pObj), Abc_ObjFaninC1(pObj) );
    }
    Vec_PtrFree( vNodes );
    // add clauses for POs
    // inlining of macro Vec_PtrForEachEntry
    for i = 0; (i < Vec_PtrSize(vFanins)); i++ { pObj = Vec_PtrEntry(vFanins, i);
        Res_SatAddEqual( pSat, int32((ABC_PTRUINT_T))pObj.pCopy, 
            int32((ABC_PTRUINT_T))Abc_ObjFanin0(pObj).pCopy, Abc_ObjFaninC0(pObj) );
    }
    // add trivial clauses
    pObj = (Abc_Obj_t *)Vec_PtrEntry(vFanins, 0);
    Res_SatAddConst1( pSat, int32((ABC_PTRUINT_T))pObj.pCopy, 0 ); // care-set
    pObj = (Abc_Obj_t *)Vec_PtrEntry(vFanins, 1);
    Res_SatAddConst1( pSat, int32((ABC_PTRUINT_T))pObj.pCopy, 0 ); // on-set

    // bookmark the clauses of A
    sat_solver_store_mark_clauses_a( pSat );

    // duplicate the clauses
    pObj = (Abc_Obj_t *)Vec_PtrEntry(vFanins, 1);
    Sat_SolverDoubleClauses( pSat, int32((ABC_PTRUINT_T))pObj.pCopy );
    // add the equality constraints
    // inlining of macro Vec_PtrForEachEntryStart
    for i = 2; (i < Vec_PtrSize(vFanins)); i++ { pObj = Vec_PtrEntry(vFanins, i);
        Res_SatAddEqual( pSat, int32((ABC_PTRUINT_T))pObj.pCopy, (int32((ABC_PTRUINT_T))pObj.pCopy) + nNodes, 0 );
    }

    // bookmark the roots
    sat_solver_store_mark_roots( pSat );

    // solve the problem
    status = sat_solver_solve( pSat, nil, nil, (ABC_INT64_T)10000, (ABC_INT64_T)0, (ABC_INT64_T)0, (ABC_INT64_T)0 );
    if status == l_False {
        pCnf = sat_solver_store_release( pSat );
//        printf( "unsat\n" );
    
    } else if status == l_True {
//        printf( "sat\n" );
    
    } else {
//        printf( "undef\n" );
    }
    sat_solver_delete( pSat );
    return pCnf;
}

/**Function*************************************************************

  Synopsis    [Loads AIG into the SAT solver for constrained simulation.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Res_SatSimulateConstr(pAig *Abc_Ntk_t, fOnSet int32) any {
    var pSat *sat_solver;
    var vFanins *Vec_Ptr_t;
    var vNodes *Vec_Ptr_t;
    var pObj *Abc_Obj_t;
    var i, nNodes int32;

    // start the array
    vFanins = Vec_PtrAlloc( 2 );
    pObj = Abc_NtkPo( pAig, 0 );
    Vec_PtrPush( vFanins, pObj );
    pObj = Abc_NtkPo( pAig, 1 );
    Vec_PtrPush( vFanins, pObj );

    // collect reachable nodes
    vNodes = Abc_NtkDfsNodes( pAig, (Abc_Obj_t **)vFanins.pArray, vFanins.nSize );

    // assign unique numbers to each node
    nNodes = 0;
    Abc_AigConst1(pAig).pCopy = (Abc_Obj_t *)(ABC_PTRUINT_T)nNodes++;
    // inlining of macro Abc_NtkForEachPi
    for i = 0; (i < Abc_NtkPiNum(pAig)); i++ { pObj = Abc_NtkPi(pAig, i);
        pObj.pCopy = (Abc_Obj_t *)(ABC_PTRUINT_T)nNodes++;
    }
    // inlining of macro Vec_PtrForEachEntry
    for i = 0; (i < Vec_PtrSize(vNodes)); i++ { pObj = Vec_PtrEntry(vNodes, i);
        pObj.pCopy = (Abc_Obj_t *)(ABC_PTRUINT_T)nNodes++;
    }
    // inlining of macro Vec_PtrForEachEntry
    for i = 0; (i < Vec_PtrSize(vFanins)); i++ { pObj = Vec_PtrEntry(vFanins, i); // useful POs
        pObj.pCopy = (Abc_Obj_t *)(ABC_PTRUINT_T)nNodes++;
    }

    // start the solver
    pSat = sat_solver_new();

    // add clause for the constant node
    Res_SatAddConst1( pSat, int32((ABC_PTRUINT_T))Abc_AigConst1(pAig).pCopy, 0 );
    // add clauses for AND gates
    // inlining of macro Vec_PtrForEachEntry
    for i = 0; (i < Vec_PtrSize(vNodes)); i++ { pObj = Vec_PtrEntry(vNodes, i);
        Res_SatAddAnd( pSat, int32((ABC_PTRUINT_T))pObj.pCopy, 
            int32((ABC_PTRUINT_T))Abc_ObjFanin0(pObj).pCopy, int32((ABC_PTRUINT_T))Abc_ObjFanin1(pObj).pCopy, Abc_ObjFaninC0(pObj), Abc_ObjFaninC1(pObj) );
    }
    Vec_PtrFree( vNodes );
    // add clauses for the first PO
    pObj = Abc_NtkPo( pAig, 0 );
    Res_SatAddEqual( pSat, int32((ABC_PTRUINT_T))pObj.pCopy, 
        int32((ABC_PTRUINT_T))Abc_ObjFanin0(pObj).pCopy, Abc_ObjFaninC0(pObj) );
    // add clauses for the second PO
    pObj = Abc_NtkPo( pAig, 1 );
    Res_SatAddEqual( pSat, int32((ABC_PTRUINT_T))pObj.pCopy, 
        int32((ABC_PTRUINT_T))Abc_ObjFanin0(pObj).pCopy, Abc_ObjFaninC0(pObj) );

    // add trivial clauses
    pObj = Abc_NtkPo( pAig, 0 );
    Res_SatAddConst1( pSat, int32((ABC_PTRUINT_T))pObj.pCopy, 0 ); // care-set

    pObj = Abc_NtkPo( pAig, 1 );
    Res_SatAddConst1( pSat, int32((ABC_PTRUINT_T))pObj.pCopy, !fOnSet ); // on-set

    Vec_PtrFree( vFanins );
    return pSat;
}

/**Function*************************************************************

  Synopsis    [Loads AIG into the SAT solver for constrained simulation.]

  Description [Returns 1 if the required number of patterns are found. 
  Returns 0 if the solver ran out of time or proved a constant. 
  In the latter, case one of the flags, fConst0 or fConst1, are set to 1.]
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Res_SatSimulate(p *Res_Sim_t, nPatsLimit int32, fOnSet int32) int32 {
    var vLits *Vec_Int_t;
    var vPats *Vec_Ptr_t;
    var pSat *sat_solver;
    var RetValue int32  = -1; // Suppress "might be used uninitialized"
    var i, k, value, status, Lit, Var, iPat int32;
    var clk abctime  = Abc_Clock();

//printf( "Looking for %s:  ", fOnSet? "onset " : "offset" );

    // decide what problem should be solved
    Lit = toLitCond( int32((ABC_PTRUINT_T))Abc_NtkPo(p.pAig,1).pCopy, !fOnSet );
    if fOnSet != 0 {
        iPat = p.nPats1;
        vPats = p.vPats1;
    
    } else {
        iPat = p.nPats0;
        vPats = p.vPats0;
    }
    assert(iPat < nPatsLimit, "iPat < nPatsLimit");

    // derive the SAT solver
    pSat = (sat_solver *)Res_SatSimulateConstr( p.pAig, fOnSet );
    pSat.fSkipSimplify = 1;
    status = sat_solver_simplify( pSat );
    if status == 0 {
        if iPat == 0 {
//            if ( fOnSet )
//                p->fConst0 = 1;
//            else
//                p->fConst1 = 1;
            RetValue = 0;
        }
        goto finish;
    }
 
    // enumerate through the SAT assignments
    RetValue = 1;
    vLits = Vec_IntAlloc( 32 );
    for  k = iPat; k < nPatsLimit; k++  {
        // solve with the assumption
//        status = sat_solver_solve( pSat, &Lit, &Lit + 1, (ABC_INT64_T)10000, (ABC_INT64_T)0, (ABC_INT64_T)0, (ABC_INT64_T)0 );
        status = sat_solver_solve( pSat, nil, nil, (ABC_INT64_T)10000, (ABC_INT64_T)0, (ABC_INT64_T)0, (ABC_INT64_T)0 );
        if status == l_False {
//printf( "Const %d\n", !fOnSet );
            if k == 0 {
                if fOnSet != 0 {
                    p.fConst0 = 1;
                } else {
                    p.fConst1 = 1; }
                RetValue = 0;
            }
            break;
        
        } else if status == l_True {
            // save the pattern
            Vec_IntClear( vLits );
            for  i = 0; i < p.nTruePis; i++  {
                Var = int32((ABC_PTRUINT_T))Abc_NtkPi(p.pAig,i).pCopy;
//                value = (int)(pSat->model.ptr[Var] == l_True);
                value = sat_solver_var_value(pSat, Var);
                if value {
                    Abc_InfoSetBit( resWords(Vec_PtrEntry(vPats, i)), k ); }
                Lit = toLitCond( Var, value );
                Vec_IntPush( vLits, Lit );
//                printf( "%d", value );
            }
//            printf( "\n" );

            status = sat_solver_addclause( pSat, Vec_IntArray(vLits), Vec_IntArray(vLits) + Vec_IntSize(vLits) );
            if status == 0 {
                k++;
                RetValue = 1; 
                break;
            }
        
        } else {
//printf( "Undecided\n" );
            if k == 0 {
                RetValue = 0;
            } else {
                RetValue = 1; } 
            break;
        }
    }
    Vec_IntFree( vLits );
//printf( "Found %d patterns\n", k - iPat );

    // set the new number of patterns
    if fOnSet != 0 {
        p.nPats1 = k;
    } else {
        p.nPats0 = k; }

finish:

    sat_solver_delete( pSat );
p.timeSat += Abc_Clock() - clk;
    return RetValue;
}


/**Function*************************************************************

  Synopsis    [Asserts equality of the variable to a constant.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
////////////////////////////////////////////////////////////////////////
///                       END OF FILE                                ///
////////////////////////////////////////////////////////////////////////


//ABC_NAMESPACE_IMPL_END

