package ssc

import "fmt"

/**CFile****************************************************************

  FileName    [sscUtil.c]

  SystemName  [ABC: Logic synthesis and verification system.]

  PackageName [SAT sweeping under constraints.]

  Synopsis    [Various utilities.]

  Author      [Alan Mishchenko]
  
  Affiliation [UC Berkeley]

  Date        [Ver. 1.0. Started - June 29, 2008.]

  Revision    [$Id: sscUtil.c,v 1.00 2008/07/29 00:00:00 alanmi Exp $]

***********************************************************************/

//#include "sscInt.h"
//#include "sat/cnf/cnf.h"
//#include "aig/gia/giaAig.h"

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
func Gia_ManDropContained(p *Gia_Man_t) *Gia_Man_t {
    var pNew *Gia_Man_t;
    var pMan *Aig_Man_t  = Gia_ManToAigSimple( p );
    var pDat *Cnf_Dat_t  = Cnf_Derive( pMan, Gia_ManPoNum(p) );
    var pObj *Gia_Obj_t;
    var vLits, vKeep *Vec_Int_t;
    var pSat *sat_solver;
    var ConstLit int32  = Abc_Var2Lit(pDat.pVarNums[0], 0);
    var i, status int32;//, Count = 0;
    Aig_ManStop( pMan );

    vLits = Vec_IntAlloc( Gia_ManPoNum(p) );
    // inlining of macro Gia_ManForEachPo
    for  i = 0; (i < Gia_ManPoNum(p)) && ((pObj) = Gia_ManCo(p, i)); i++  {
        var iObj int32  = Gia_ObjId( p, pObj );
        Vec_IntPush( vLits, Abc_Var2Lit(pDat.pVarNums[iObj], 1) );
    }

    // start the SAT solver
    pSat = sat_solver_new();
    sat_solver_setnvars( pSat, pDat.nVars );
    for  i = 0; i < pDat.nClauses; i++  {
        if !sat_solver_addclause( pSat, pDat.pClauses[i], pDat.pClauses[i+1] ) {
            Cnf_DataFree( pDat );
            sat_solver_delete( pSat );
            Vec_IntFree( vLits );
            return nil;
        }
    }
    Cnf_DataFree( pDat );
    status = sat_solver_simplify( pSat );
    if status == 0 {
        sat_solver_delete( pSat );
        Vec_IntFree( vLits );
        return nil;
    }

    // iterate over POs
    vKeep = Vec_IntAlloc( Gia_ManPoNum(p) );
    // inlining of macro Gia_ManForEachPo
    for  i = 0; (i < Gia_ManPoNum(p)) && ((pObj) = Gia_ManCo(p, i)); i++  {
        Vec_IntWriteEntry( vLits, i, Abc_LitNot(Vec_IntEntry(vLits,i)) );
        status = sat_solver_solve( pSat, Vec_IntArray(vLits), Vec_IntArray(vLits) + Vec_IntSize(vLits), 0, 0, 0, 0 );
        Vec_IntWriteEntry( vLits, i, Abc_LitNot(Vec_IntEntry(vLits,i)) );
        if status == l_False {
            Vec_IntWriteEntry( vLits, i, ConstLit ); // const1 SAT var is always true
        } else {
            assert(status == l_True, "status == l_True");
            Vec_IntPush( vKeep, i );
        }
    }
    sat_solver_delete( pSat );
    Vec_IntFree( vLits );
    if Vec_IntSize(vKeep) == Gia_ManPoNum(p) {
        Vec_IntFree( vKeep );
        return Gia_ManDup(p);
    }
    pNew = Gia_ManDupCones( p, Vec_IntArray(vKeep), Vec_IntSize(vKeep), 0 );
    Vec_IntFree( vKeep );
    return pNew;
}

/**Function*************************************************************

  Synopsis    []

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Gia_ManOptimizeRing(p *Gia_Man_t) *Gia_Man_t {
    var Pars Ssc_Pars_t;
    var pPars *Ssc_Pars_t = &Pars;
    var pTemp, pAux *Gia_Man_t;
    var i int32;
    assert(p.nConstrs == 0, "p.nConstrs == 0");
    fmt.Printf("User AIG: ");
    Gia_ManPrintStats( p, nil );
    pTemp = Gia_ManDropContained( p );
    fmt.Printf("Drop AIG: ");
    Gia_ManPrintStats( pTemp, nil );
//    return pTemp;
    if Gia_ManPoNum(pTemp) == 1 {
        return pTemp; }
    Ssc_ManSetDefaultParams( pPars );
    pPars.fAppend  = 1;
    pPars.fVerbose = 0;
    pTemp.nConstrs = Gia_ManPoNum(pTemp) - 1;    
    for  i = 0; i < Gia_ManPoNum(pTemp); i++  {
        // move i-th PO forward
        Gia_ManSwapPos( pTemp, i );
        pTemp = Gia_ManDupDfs( pAux = pTemp );
        Gia_ManStop( pAux );
        // minimize this PO
        pTemp = Ssc_PerformSweepingConstr( pAux = pTemp, pPars );
        Gia_ManStop( pAux );
        pTemp = Gia_ManDupDfs( pAux = pTemp );
        Gia_ManStop( pAux );
        // move i-th PO back
        Gia_ManSwapPos( pTemp, i );
        pTemp = Gia_ManDupDfs( pAux = pTemp );
        Gia_ManStop( pAux );
        // report results
        fmt.Printf("AIG%3d  : ", i);
        Gia_ManPrintStats( pTemp, nil );
    }
    pTemp.nConstrs = 0;
    return pTemp;
}

////////////////////////////////////////////////////////////////////////
///                       END OF FILE                                ///
////////////////////////////////////////////////////////////////////////


//ABC_NAMESPACE_IMPL_END

