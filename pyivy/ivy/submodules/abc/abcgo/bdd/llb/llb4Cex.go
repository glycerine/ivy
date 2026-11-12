package llb

import "fmt"

/**CFile****************************************************************

  FileName    [llb2Cex.c]

  SystemName  [ABC: Logic synthesis and verification system.]

  PackageName [BDD based reachability.]

  Synopsis    [Non-linear quantification scheduling.]

  Author      [Alan Mishchenko]
  
  Affiliation [UC Berkeley]

  Date        [Ver. 1.0. Started - June 20, 2005.]

  Revision    [$Id: llb2Cex.c,v 1.00 2005/06/20 00:00:00 alanmi Exp $]

***********************************************************************/

//#include "llbInt.h"
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

  Synopsis    [Translates a sequence of states into a counter-example.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Llb4_Nonlin4TransformCex(pAig *Aig_Man_t, vStates *Vec_Ptr_t, iCexPo int32, fVerbose int32) *Abc_Cex_t {
    var pCex *Abc_Cex_t;
    var pCnf *Cnf_Dat_t;
    var vAssumps *Vec_Int_t;
    var pSat *sat_solver;
    var pObj *Aig_Obj_t;
    var pNext, pThis *uint32;
    var i, k, iBit, status, nRegs int32;//, clk = Abc_Clock();
/*
    Vec_PtrForEachEntry( unsigned *, vStates, pNext, i )
    {
        printf( "%4d : ", i );
        Extra_PrintBinary( stdout, pNext, Aig_ManRegNum(pAig) );
        printf( "\n" );
    }
*/
    // derive SAT solver
    nRegs = Aig_ManRegNum(pAig); pAig.nRegs = 0;
    pCnf = Cnf_Derive( pAig, Aig_ManCoNum(pAig) );
    pAig.nRegs = nRegs;
//    Cnf_DataTranformPolarity( pCnf, 0 );
    // convert into SAT solver
    pSat = (sat_solver *)Cnf_DataWriteIntoSolver( pCnf, 1, 0 );
    if pSat == nil {
        fmt.Printf("Llb4_Nonlin4TransformCex(): Counter-example generation has failed.\n");
        Cnf_DataFree( pCnf );
        return nil;
    }
    // simplify the problem
    status = sat_solver_simplify(pSat);
    if status == 0 {
        fmt.Printf("Llb4_Nonlin4TransformCex(): SAT solver is invalid.\n");
        sat_solver_delete( pSat );
        Cnf_DataFree( pCnf );
        return nil;
    }
    // start the counter-example
    pCex = Abc_CexAlloc( Saig_ManRegNum(pAig), Saig_ManPiNum(pAig), Vec_PtrSize(vStates) );
    pCex.iFrame = Vec_PtrSize(vStates)-1;
    pCex.iPo = -1;

    // solve each time frame
    iBit = Saig_ManRegNum(pAig);
    pThis = (unsigned *)Vec_PtrEntry( vStates, 0 );
    vAssumps = Vec_IntAlloc( 2 * Aig_ManRegNum(pAig) );
    Vec_PtrForEachEntryStart( unsigned *, vStates, pNext, i, 1 )
    {
        // create assumptions
        Vec_IntClear( vAssumps );
        Saig_ManForEachLo( pAig, pObj, k )
            Vec_IntPush( vAssumps, toLitCond( pCnf.pVarNums[Aig_ObjId(pObj)], !Abc_InfoHasBit(pThis,k) ) );
        Saig_ManForEachLi( pAig, pObj, k )
            Vec_IntPush( vAssumps, toLitCond( pCnf.pVarNums[Aig_ObjId(pObj)], !Abc_InfoHasBit(pNext,k) ) );
        // solve SAT problem
        status = sat_solver_solve( pSat, Vec_IntArray(vAssumps), Vec_IntArray(vAssumps) + Vec_IntSize(vAssumps), 
            (ABC_INT64_T)0, (ABC_INT64_T)0, (ABC_INT64_T)0, (ABC_INT64_T)0 );
        // if the problem is SAT, get the counterexample
        if status != l_True {
            fmt.Printf("Llb4_Nonlin4TransformCex(): There is no transition between state %d and %d.\n", i-1, i);
            Vec_IntFree( vAssumps );
            sat_solver_delete( pSat );
            Cnf_DataFree( pCnf );
            ABC_FREE( pCex );
            return nil;
        }
        // get the assignment of PIs
        Saig_ManForEachPi( pAig, pObj, k )
            if ( sat_solver_var_value(pSat, pCnf.pVarNums[Aig_ObjId(pObj)]) )
                Abc_InfoSetBit( pCex.pData, iBit + k );
        // update the counter
        iBit += Saig_ManPiNum(pAig);
        pThis = pNext;
    }

    // add the last frame when the property fails
    Vec_IntClear( vAssumps );
    if iCexPo >= 0 {
        Saig_ManForEachPo( pAig, pObj, k )
            if ( k == iCexPo )
                Vec_IntPush( vAssumps, toLitCond( pCnf.pVarNums[Aig_ObjId(pObj)], 0 ) );
    
    } else {
        Saig_ManForEachPo( pAig, pObj, k )
            Vec_IntPush( vAssumps, toLitCond( pCnf.pVarNums[Aig_ObjId(pObj)], 0 ) );
    }

    // add clause
    status = sat_solver_addclause( pSat, Vec_IntArray(vAssumps), Vec_IntArray(vAssumps) + Vec_IntSize(vAssumps) );
    if status == 0 {
        fmt.Printf("Llb4_Nonlin4TransformCex(): The SAT solver is unsat after adding last clause.\n");
        Vec_IntFree( vAssumps );
        sat_solver_delete( pSat );
        Cnf_DataFree( pCnf );
        ABC_FREE( pCex );
        return nil;
    }
    // create assumptions
    Vec_IntClear( vAssumps );
    Saig_ManForEachLo( pAig, pObj, k )
        Vec_IntPush( vAssumps, toLitCond( pCnf.pVarNums[Aig_ObjId(pObj)], !Abc_InfoHasBit(pThis,k) ) );
    // solve the last frame
    status = sat_solver_solve( pSat, Vec_IntArray(vAssumps), Vec_IntArray(vAssumps) + Vec_IntSize(vAssumps), 
        (ABC_INT64_T)0, (ABC_INT64_T)0, (ABC_INT64_T)0, (ABC_INT64_T)0 );
    if status != l_True {
        fmt.Printf("Llb4_Nonlin4TransformCex(): There is no last transition that makes the property fail.\n");
        Vec_IntFree( vAssumps );
        sat_solver_delete( pSat );
        Cnf_DataFree( pCnf );
        ABC_FREE( pCex );
        return nil;
    }
    // get the assignment of PIs
    Saig_ManForEachPi( pAig, pObj, k )
        if ( sat_solver_var_value(pSat, pCnf.pVarNums[Aig_ObjId(pObj)]) )
            Abc_InfoSetBit( pCex.pData, iBit + k );
    iBit += Saig_ManPiNum(pAig);
    assert(iBit == pCex.nBits, "iBit == pCex.nBits");

    // free the sat_solver
    Vec_IntFree( vAssumps );
    sat_solver_delete( pSat );
    Cnf_DataFree( pCnf );

    // verify counter-example
    status = Saig_ManFindFailedPoCex( pAig, pCex );
    if status >= 0 && status < Saig_ManPoNum(pAig) {
        pCex.iPo = status;
    } else {
        fmt.Printf("Llb4_Nonlin4TransformCex(): Counter-example verification has FAILED.\n");
        ABC_FREE( pCex );
        return nil;
    }
    // report the results
//    if ( fVerbose )
//       Abc_PrintTime( 1, "SAT-based cex generation time", Abc_Clock() - clk );
    return pCex;
}


/**Function*************************************************************

  Synopsis    [Resimulates the counter-example.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Llb4_Nonlin4VerifyCex(pAig *Aig_Man_t, p *Abc_Cex_t) *Vec_Ptr_t {
    var vStates *Vec_Ptr_t;
    var pObj, pObjRi, pObjRo *Aig_Obj_t;
    var i, k, iBit int32;
    // create storage for states
    vStates = Vec_PtrAllocSimInfo( p.iFrame+1, Abc_BitWordNum(Aig_ManRegNum(pAig)) );
    Vec_PtrCleanSimInfo( vStates, 0, Abc_BitWordNum(Aig_ManRegNum(pAig)) );
    // verify counter-example
    Aig_ManCleanMarkB(pAig);
    Aig_ManConst1(pAig).fMarkB = 1;
    Saig_ManForEachLo( pAig, pObj, i )
        pObj.fMarkB = 0; //Abc_InfoHasBit(p->pData, iBit++);
    // do not require equal flop count in the AIG and in the CEX
    iBit = p.nRegs;
    for  i = 0; i <= p.iFrame; i++  {
        // save current state
        Saig_ManForEachLo( pAig, pObj, k )
            if ( pObj.fMarkB )
                Abc_InfoSetBit( (unsigned *)Vec_PtrEntry(vStates, i), k );
        // compute new state
        Saig_ManForEachPi( pAig, pObj, k )
            pObj.fMarkB = Abc_InfoHasBit(p.pData, iBit++);
        Aig_ManForEachNode( pAig, pObj, k )
            pObj.fMarkB = (Aig_ObjFanin0(pObj).fMarkB ^ Aig_ObjFaninC0(pObj)) & 
                           (Aig_ObjFanin1(pObj).fMarkB ^ Aig_ObjFaninC1(pObj));
        Aig_ManForEachCo( pAig, pObj, k )
            pObj.fMarkB = Aig_ObjFanin0(pObj).fMarkB ^ Aig_ObjFaninC0(pObj);
        if i == p.iFrame {
            break; }
        Saig_ManForEachLiLo( pAig, pObjRi, pObjRo, k )
            pObjRo.fMarkB = pObjRi.fMarkB;
    }
/*
    {
        unsigned * pNext;
        Vec_PtrForEachEntry( unsigned *, vStates, pNext, i )
        {
            printf( "%4d : ", i );
            Extra_PrintBinary( stdout, pNext, Aig_ManRegNum(pAig) );
            printf( "\n" );
        }
    }
*/
    assert(iBit == p.nBits, "iBit == p.nBits");
//    if ( Aig_ManCo(pAig, p->iPo)->fMarkB == 0 )
//        Vec_PtrFreeP( &vStates );
    for  i = Saig_ManPoNum(pAig) - 1; i >= 0; i--  {
        if Aig_ManCo(pAig, i).fMarkB {
            p.iPo = i;
            break;
        }
    }
    if i == -1 {
        Vec_PtrFreeP( &vStates ); }
    Aig_ManCleanMarkB(pAig);
    return vStates;
}

/**Function*************************************************************

  Synopsis    [Translates a sequence of states into a counter-example.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Llb4_Nonlin4NormalizeCex(pAigOrg *Aig_Man_t, pAigRpm *Aig_Man_t, pCexRpm *Abc_Cex_t) *Abc_Cex_t {
    var pCexOrg *Abc_Cex_t;
    var vStates *Vec_Ptr_t;
    // check parameters of the AIG
    if Saig_ManRegNum(pAigOrg) != Saig_ManRegNum(pAigRpm) {
        fmt.Printf("Llb4_Nonlin4NormalizeCex(): The number of flops in the original and reparametrized AIGs do not agree.\n");
        return nil;
    }
/*
    if ( Saig_ManRegNum(pAigRpm) != pCexRpm->nRegs )
    {
        printf( "Llb4_Nonlin4NormalizeCex(): The number of flops in the reparametrized AIG and in the CEX do not agree.\n" );
        return NULL;
    }
*/
    if Saig_ManPiNum(pAigRpm) != pCexRpm.nPis {
        fmt.Printf("Llb4_Nonlin4NormalizeCex(): The number of PIs in the reparametrized AIG and in the CEX do not agree.\n");
        return nil;
    }
    // get the sequence of states
    vStates = Llb4_Nonlin4VerifyCex( pAigRpm, pCexRpm );
    if vStates == nil {
        Abc_Print( 1, "Llb4_Nonlin4NormalizeCex(): The given CEX does not fail outputs of pAigRpm.\n" );
        return nil;
    }
    // derive updated counter-example
    pCexOrg = Llb4_Nonlin4TransformCex( pAigOrg, vStates, pCexRpm.iPo, 0 );
    Vec_PtrFree( vStates );
    return pCexOrg;
}

////////////////////////////////////////////////////////////////////////
///                       END OF FILE                                ///
////////////////////////////////////////////////////////////////////////


//ABC_NAMESPACE_IMPL_END

