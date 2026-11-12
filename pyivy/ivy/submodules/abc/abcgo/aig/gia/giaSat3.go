package gia

/**CFile****************************************************************

  FileName    [giaSatoko.c]

  SystemName  [ABC: Logic synthesis and verification system.]

  PackageName [Scalable AIG package.]

  Synopsis    [Interface to Satoko solver.]

  Author      [Alan Mishchenko, Bruno Schmitt]
  
  Affiliation [UC Berkeley]

  Date        [Ver. 1.0. Started - June 20, 2005.]

  Revision    [$Id: giaSatoko.c,v 1.00 2005/06/20 00:00:00 alanmi Exp $]

***********************************************************************/

//#include "gia.h"
//#include "sat/cnf/cnf.h"
//#include "sat/bsat/satSolver3.h"

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
func Gia_ManSat3Init(pCnf *Cnf_Dat_t) *sat_solver3 {
    var pSat *sat_solver3  = sat_solver3_new();
    var i int32;
    //sat_solver_setnvars( pSat, p->nVars );
    for  i = 0; i < pCnf.nClauses; i++  {
        if !sat_solver3_addclause( pSat, pCnf.pClauses[i], pCnf.pClauses[i+1] ) {
            sat_solver3_delete( pSat );
            return nil;
        }
    }
    return pSat;
}
func Gia_ManSat3Report(iOutput int32, status int32, clk abctime) {
    if iOutput >= 0 {
        Abc_Print( 1, "Output %6d : ", iOutput );
    } else {
        Abc_Print( 1, "Total: " ); }

    if status == l_Undef {
        Abc_Print( 1, "UNDECIDED      " );
    } else if status == l_True {
        Abc_Print( 1, "SATISFIABLE    " );
    } else {
        Abc_Print( 1, "UNSATISFIABLE  " ); }

    Abc_PrintTime( 1, "Time", clk );
}
func Gia_ManSat3Create(p *Gia_Man_t) *sat_solver3 {
    var pCnf *Cnf_Dat_t  = (Cnf_Dat_t *)Mf_ManGenerateCnf( p, 8, 0, 1, 0, 0 );
    var pSat *sat_solver3  = Gia_ManSat3Init( pCnf );
    var status int32;
    if pSat != nil {
        status = sat_solver3_simplify(pSat);
    } else {
        status = 0;
    }
    Cnf_DataFree( pCnf );
    if status != 0 {
        return pSat; }
    if pSat != nil {
        sat_solver3_delete( pSat ); }
    return nil;
}
func Gia_ManSat3CallOne(p *Gia_Man_t, iOutput int32) int32 {
    var clk abctime  = Abc_Clock();
    var pSat *sat_solver3;
    var status, Cost int32;

    pSat = Gia_ManSat3Create( p );
    if pSat != nil {
        status = sat_solver3_solve( pSat, nil, nil, 0, 0, 0, 0 );
        Cost = uint32(pSat.stats.conflicts);
        sat_solver3_delete( pSat );
    
    } else { 
        status = l_False; }

    Gia_ManSat3Report( iOutput, status, Abc_Clock() - clk );
    return Cost;
}
func Gia_ManSat3Call(p *Gia_Man_t, fSplit int32) {
    var pOne *Gia_Man_t;
    var pRoot *Gia_Obj_t;
    var i int32;
    if fSplit != 0 {
        var clk abctime  = Abc_Clock();
        // inlining of macro Gia_ManForEachCo
        for  i = 0; (i < Vec_IntSize(p.vCos)) && ((pRoot) = Gia_ManCo(p, i)); i++  {
            pOne = Gia_ManDupDfsCone( p, pRoot );
            Gia_ManSat3CallOne( pOne, i );
            Gia_ManStop( pOne );
        }
        Abc_PrintTime( 1, "Total time", Abc_Clock() - clk );
        return;
    }
    Gia_ManSat3CallOne( p, -1 );
}    


////////////////////////////////////////////////////////////////////////
///                       END OF FILE                                ///
////////////////////////////////////////////////////////////////////////


//ABC_NAMESPACE_IMPL_END

