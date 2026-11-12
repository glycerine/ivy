package gia

import (
	"fmt"
	"strconv"
)

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
//#include "misc/extra/extra.h"
//#include "sat/satoko/satoko.h"

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
func Gia_ManCollectVars_rec(Var int32, vMapping *Vec_Int_t, vRes *Vec_Int_t, vVisit *Vec_Bit_t) {
    var pCut *int32;
    var i int32;
    if Vec_BitEntry(vVisit, Var) {
        return; }
    Vec_BitWriteEntry(vVisit, Var, 1);
    if Vec_IntEntry(vMapping, Var) { // primary input or constant 0
    
        pCut = Vec_IntEntryP( vMapping, Vec_IntEntry(vMapping, Var) );
        for  i = 1; i <= pCut[0]; i++  {
            Gia_ManCollectVars_rec( pCut[i], vMapping, vRes, vVisit );
        }
    }
    Vec_IntPush( vRes, Var );        
}
func Gia_ManCollectVars(Root int32, vMapping *Vec_Int_t, nVars int32) *Vec_Int_t {
    var vRes *Vec_Int_t  = Vec_IntAlloc( 100 );
    var vVisit *Vec_Bit_t  = Vec_BitStart( nVars );
    assert(Vec_IntEntry(vMapping, Root), "Vec_IntEntry(vMapping, Root)");
    Gia_ManCollectVars_rec( Root, vMapping, vRes, vVisit );
    Vec_BitFree( vVisit );
    return vRes;
}

/**Function*************************************************************

  Synopsis    []

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Gia_ManSatokoReport(iOutput int32, status int32, clk abctime) {
    if iOutput >= 0 {
        Abc_Print( 1, "Output %6d : ", iOutput );
    } else {
        Abc_Print( 1, "Total: " ); }

    if status == SATOKO_UNDEC {
        Abc_Print( 1, "UNDECIDED      " );
    } else if status == SATOKO_SAT {
        Abc_Print( 1, "SATISFIABLE    " );
    } else {
        Abc_Print( 1, "UNSATISFIABLE  " ); }

    Abc_PrintTime( 1, "Time", clk );
}
func Gia_ManSatokoFromDimacs(pFileName string, opts *satoko_opts_t) *satoko_t {
    var pSat *satoko_t  = satoko_create();
    var pBuffer string  = Extra_FileReadContents( pFileName );
    var vLits *Vec_Int_t  = Vec_IntAlloc( 100 );
    var pTemp string; var fComp, Var, VarMax int32;
    for  pTemp = pBuffer; *pTemp; pTemp++  {
        if *pTemp == 'c' || *pTemp == 'p' {
            for  *pTemp != '\n'  {
                pTemp++;
            }
            continue;
        }
        for  *pTemp == ' ' || *pTemp == '\t' || *pTemp == '\r' || *pTemp == '\n'  {
            pTemp++;
        }
        fComp = 0;
        if *pTemp == '-' {
            fComp = 1;
pTemp++; }
        if *pTemp == '+' {
            pTemp++; }
        Var = strconv.Atoi(pTemp);
        if Var == 0 {
            if Vec_IntSize(vLits) > 0 {
                satoko_setnvars( pSat, VarMax+1 );
                if !satoko_add_clause( pSat, Vec_IntArray(vLits), Vec_IntSize(vLits) ) {
                    satoko_destroy(pSat);
                    Vec_IntFree( vLits );
                    // inlining of macro ABC_FREE
                    ((pBuffer) ? (free((char *) (pBuffer)), (pBuffer) = 0) : 0);
                    return nil;
                }
                Vec_IntClear( vLits );
            }
        
        } else {
            Var--;
            VarMax = Abc_MaxInt( VarMax, Var );
            Vec_IntPush( vLits, Abc_Var2Lit(Var, fComp) );
        }
        for  *pTemp >= '0' && *pTemp <= '9'  {
            pTemp++;
        }
    }
    // inlining of macro ABC_FREE
    ((pBuffer) ? (free((char *) (pBuffer)), (pBuffer) = 0) : 0);
    Vec_IntFree( vLits );
    return pSat;
}
func Gia_ManSatokoDimacs(pFileName string, opts *satoko_opts_t) {
    var clk abctime  = Abc_Clock();  
    var status int32  = SATOKO_UNSAT;
    var pSat *satoko_t  = Gia_ManSatokoFromDimacs( pFileName, opts );
    if pSat != nil {
        status = satoko_solve( pSat );
        satoko_destroy( pSat );
    }
    Gia_ManSatokoReport( -1, status, Abc_Clock() - clk );
}

/**Function*************************************************************

  Synopsis    []

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Gia_ManSatokoInit(pCnf *Cnf_Dat_t, opts *satoko_opts_t) *satoko_t {
    var pSat *satoko_t  = satoko_create();
    var i int32;
    //sat_solver_setnvars( pSat, p->nVars );
    for  i = 0; i < pCnf.nClauses; i++  {
        if !satoko_add_clause( pSat, pCnf.pClauses[i], pCnf.pClauses[i+1]-pCnf.pClauses[i] ) {
            satoko_destroy( pSat );
            return nil;
        }
    }
    satoko_configure(pSat, opts);
    return pSat;
}
func Gia_ManSatokoCreate(p *Gia_Man_t, opts *satoko_opts_t) *satoko_t {
    var pCnf *Cnf_Dat_t  = (Cnf_Dat_t *)Mf_ManGenerateCnf( p, 8, 0, 1, 0, 0 );
    var pSat *satoko_t  = Gia_ManSatokoInit( pCnf, opts );
    var status int32;
    if pSat != nil {
        status = satoko_simplify(pSat);
    } else {
        status = SATOKO_OK;
    }
    Cnf_DataFree( pCnf );
    if status == SATOKO_OK {
        return pSat; }
    satoko_destroy( pSat );
    return nil;
}
func Gia_ManSatokoCallOne(p *Gia_Man_t, opts *satoko_opts_t, iOutput int32) int32 {
    var clk abctime  = Abc_Clock();
    var pSat *satoko_t;
    var status int32 = SATOKO_UNSAT;
    var Cost int32;
    pSat = Gia_ManSatokoCreate( p, opts );
    if pSat != nil {
        status = satoko_solve( pSat );
        Cost = satoko_stats(pSat).n_conflicts;
        satoko_destroy( pSat );
    }
    Gia_ManSatokoReport( iOutput, status, Abc_Clock() - clk );
    return Cost;
}
func Gia_ManSatokoCall(p *Gia_Man_t, opts *satoko_opts_t, fSplit int32, fIncrem int32) {
    var fUseCone int32  = 1;
    var pOne *Gia_Man_t;
    var pRoot *Gia_Obj_t;
    var vCone *Vec_Int_t;
    var i, iLit, status int32;
    if fIncrem != 0 {
        var clk abctime  = Abc_Clock();
        var pCnf *Cnf_Dat_t  = (Cnf_Dat_t *)Mf_ManGenerateCnf( p, 8, 0, 0, fUseCone, 0 );
        var pSat *satoko_t  = Gia_ManSatokoInit( pCnf, opts );
        // inlining of macro Gia_ManForEachCo
        for  i = 0; (i < Vec_IntSize(p.vCos)) && ((pRoot) = Gia_ManCo(p, i)); i++  {
            var clk abctime  = Abc_Clock();
            iLit = Abc_Var2Lit( i+1, 0 );
            satoko_assump_push( pSat, iLit );
            if fUseCone != 0 {
                vCone = Gia_ManCollectVars( i+1, pCnf.vMapping, pCnf.nVars );
                satoko_mark_cone( pSat, Vec_IntArray(vCone), Vec_IntSize(vCone) );
                fmt.Printf("Cone has %6d vars (out of %6d).  ", Vec_IntSize(vCone), pCnf.nVars);
                status = satoko_solve( pSat );
                satoko_unmark_cone( pSat, Vec_IntArray(vCone), Vec_IntSize(vCone) );
                Vec_IntFree( vCone );
            
            } else {
                status = satoko_solve( pSat );
            }
            satoko_assump_pop( pSat );
            Gia_ManSatokoReport( i, status, Abc_Clock() - clk );
        }
        Cnf_DataFree( pCnf );
        satoko_destroy( pSat );
        Abc_PrintTime( 1, "Total time", Abc_Clock() - clk );
        return;
    }
    if fSplit != 0 {
        var clk abctime  = Abc_Clock();
        // inlining of macro Gia_ManForEachCo
        for  i = 0; (i < Vec_IntSize(p.vCos)) && ((pRoot) = Gia_ManCo(p, i)); i++  {
            pOne = Gia_ManDupDfsCone( p, pRoot );
            Gia_ManSatokoCallOne( pOne, opts, i );
            Gia_ManStop( pOne );
        }
        Abc_PrintTime( 1, "Total time", Abc_Clock() - clk );
        return;
    }
    Gia_ManSatokoCallOne( p, opts, -1 );
}    


////////////////////////////////////////////////////////////////////////
///                       END OF FILE                                ///
////////////////////////////////////////////////////////////////////////


//ABC_NAMESPACE_IMPL_END

