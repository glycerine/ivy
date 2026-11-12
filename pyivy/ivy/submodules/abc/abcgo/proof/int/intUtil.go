package int

import "time"

/**CFile****************************************************************

  FileName    [intUtil.c]

  SystemName  [ABC: Logic synthesis and verification system.]

  PackageName [Interpolation engine.]

  Synopsis    [Various interpolation utilities.]

  Author      [Alan Mishchenko]
  
  Affiliation [UC Berkeley]

  Date        [Ver. 1.0. Started - June 24, 2008.]

  Revision    [$Id: intUtil.c,v 1.00 2005/06/20 00:00:00 alanmi Exp $]

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

  Synopsis    [Returns 1 if the property fails in the initial state.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Inter_ManCheckInitialState(p *Aig_Man_t) int32 {
    var pCnf *Cnf_Dat_t;
    var pObj *Aig_Obj_t;
    var pSat *sat_solver;
    var i, status int32;
    //abctime clk = Abc_Clock();
    pCnf = Cnf_Derive( p, Saig_ManRegNum(p) ); 
    pSat = (sat_solver *)Cnf_DataWriteIntoSolver( pCnf, 1, 1 );
    if pSat == nil {
        Cnf_DataFree( pCnf );
        return 0;
    }
    status = sat_solver_solve( pSat, nil, nil, (ABC_INT64_T)0, (ABC_INT64_T)0, (ABC_INT64_T)0, (ABC_INT64_T)0 );
    //ABC_PRT( "Time", Abc_Clock() - clk );
    if status == l_True {
        p.pSeqModel = Abc_CexAlloc( Aig_ManRegNum(p), Saig_ManPiNum(p), 1 );
        // inlining of macro Saig_ManForEachPi
        // inlining of macro Vec_PtrForEachEntryStop
        for  i = 0; (i < Saig_ManPiNum(p)) && (((pObj) = (Aig_Obj_t *)Vec_PtrEntry(p.vCis, i)), 1); i++  {
            if sat_solver_var_value( pSat, pCnf.pVarNums[Aig_ObjId(pObj)] ) {
                Abc_InfoSetBit( p.pSeqModel.pData, Aig_ManRegNum(p) + i ); }
        }
    }
    Cnf_DataFree( pCnf );
    sat_solver_delete( pSat );
    return status == l_True;
}

/**Function*************************************************************

  Synopsis    [Returns 1 if the property holds in all states.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Inter_ManCheckAllStates(p *Aig_Man_t) int32 {
    var pCnf *Cnf_Dat_t;
    var pSat *sat_solver;
    var status int32;
    var clk abctime  = Abc_Clock();
    pCnf = Cnf_Derive( p, Saig_ManRegNum(p) ); 
    pSat = (sat_solver *)Cnf_DataWriteIntoSolver( pCnf, 1, 0 );
    Cnf_DataFree( pCnf );
    if pSat == nil {
        return 1; }
    status = sat_solver_solve( pSat, nil, nil, (ABC_INT64_T)0, (ABC_INT64_T)0, (ABC_INT64_T)0, (ABC_INT64_T)0 );
    sat_solver_delete( pSat );
    // inlining of macro ABC_PRT
    Abc_Print(1, "%s =", ("Time"));
Abc_Print(1, "%9.2f sec\n", 1.0*(float64((Abc_Clock() - clk)))/(float64(int64(time.Second))));;
    return status == l_False;
}

////////////////////////////////////////////////////////////////////////
///                       END OF FILE                                ///
////////////////////////////////////////////////////////////////////////


//ABC_NAMESPACE_IMPL_END

