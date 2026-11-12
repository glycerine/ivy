package bmc

import "fmt"

/**CFile****************************************************************

  FileName    [bmcLoad.c]

  SystemName  [ABC: Logic synthesis and verification system.]

  PackageName [SAT-based bounded model checking.]

  Synopsis    [Experiments with CNF loading.]

  Author      [Alan Mishchenko]
  
  Affiliation [UC Berkeley]

  Date        [Ver. 1.0. Started - June 20, 2005.]

  Revision    [$Id: bmcLoad.c,v 1.00 2005/06/20 00:00:00 alanmi Exp $]

***********************************************************************/

//#include "bmc.h"
//#include "sat/bsat/satStore.h"

//ABC_NAMESPACE_IMPL_START


////////////////////////////////////////////////////////////////////////
///                        DECLARATIONS                              ///
////////////////////////////////////////////////////////////////////////

     
type Bmc_Load_t struct {
    pPars    *Bmc_AndPar_t;     // parameters
    pGia       *Gia_Man_t;      // unrolled AIG
    pSat      *sat_solver;      // SAT solvers
    vSat2Id       *Vec_Int_t;   // maps SAT var into its node
//    Vec_Int_t *      vCut;      // cut in terms of GIA IDs
//    Vec_Int_t *      vCnf;      // CNF for the cut
    nCallBacks1              int32; 
    nCallBacks2              int32; 
}

////////////////////////////////////////////////////////////////////////
///                     FUNCTION DEFINITIONS                         ///
////////////////////////////////////////////////////////////////////////


/**Function*************************************************************

  Synopsis    [Load CNF for the cone.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Bmc_LoadGetSatVar(p *Bmc_Load_t, Id int32) int32 {
    var pObj *Gia_Obj_t  = Gia_ManObj( p.pGia, Id );
    if pObj.Value == 0 {
        pObj.Value = Vec_IntSize( p.vSat2Id );
        Vec_IntPush( p.vSat2Id, Id );
        sat_solver_setnvars( p.pSat, Vec_IntSize(p.vSat2Id) );
    }
    return pObj.Value;
}
func Bmc_LoadAddCnf(pMan any, iLit int32) int32 {
    var p *Bmc_Load_t  = (Bmc_Load_t *)pMan;
    var Lits [3]int32;
    var iVar int32 = Abc_Lit2Var(iLit);
    var pObj *Gia_Obj_t  = Gia_ManObj( p.pGia, Vec_IntEntry(p.vSat2Id, iVar) );
    p.nCallBacks1++;
    if Gia_ObjIsCi(pObj) || Gia_ObjIsConst0(pObj) {
        return 0; }
    assert(Gia_ObjIsAnd(pObj), "Gia_ObjIsAnd(pObj)");
    if Abc_LitIsCompl(iLit) ? pObj.fMark1 : pObj.fMark0 {
        return 0; }
    Lits[0] = Abc_LitNot(iLit);
    if Abc_LitIsCompl(iLit) {
        Lits[1] = Abc_Var2Lit( Bmc_LoadGetSatVar(p, Gia_ObjFaninId0p(p.pGia, pObj)), !Gia_ObjFaninC0(pObj) );
        Lits[2] = Abc_Var2Lit( Bmc_LoadGetSatVar(p, Gia_ObjFaninId1p(p.pGia, pObj)), !Gia_ObjFaninC1(pObj) );
        sat_solver_clause_new( p.pSat, Lits, Lits + 3, 0 );
        pObj.fMark1 = 1;
    
    } else {
        Lits[1] = Abc_Var2Lit( Bmc_LoadGetSatVar(p, Gia_ObjFaninId0p(p.pGia, pObj)), Gia_ObjFaninC0(pObj) );
        sat_solver_clause_new( p.pSat, Lits, Lits + 2, 0 );
        Lits[1] = Abc_Var2Lit( Bmc_LoadGetSatVar(p, Gia_ObjFaninId1p(p.pGia, pObj)), Gia_ObjFaninC1(pObj) );
        sat_solver_clause_new( p.pSat, Lits, Lits + 2, 0 );
        pObj.fMark0 = 1;
    }
    p.nCallBacks2++;
    return 1;
}
func Bmc_LoadAddCnf_rec(p *Bmc_Load_t, Id int32) int32 {
    var iVar int32  = Bmc_LoadGetSatVar( p, Id );
    var pObj *Gia_Obj_t  = Gia_ManObj( p.pGia, Id );
    if Gia_ObjIsAnd(pObj) && !(pObj.fMark0 && pObj.fMark1) {
        Bmc_LoadAddCnf( p, Abc_Var2Lit(iVar, 0) );
        Bmc_LoadAddCnf( p, Abc_Var2Lit(iVar, 1) );
        Bmc_LoadAddCnf_rec( p, Gia_ObjFaninId0(pObj, Id) );
        Bmc_LoadAddCnf_rec( p, Gia_ObjFaninId1(pObj, Id) );
    }
    return iVar;
}

/**Function*************************************************************

  Synopsis    []

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Bmc_LoadStart(pGia *Gia_Man_t) *Bmc_Load_t {
    
    var Lit int32;
    Gia_ManSetPhase( pGia );
    Gia_ManCleanValue( pGia );
    Gia_ManCreateRefs( pGia );
    // inlining of macro ABC_CALLOC
    p := &Bmc_Load_t{};
    p.pGia      = pGia;
    p.pSat      = sat_solver_new();
    p.vSat2Id   = Vec_IntAlloc( 1000 );
    Vec_IntPush( p.vSat2Id, 0 );
    // create constant node
    Lit = Abc_Var2Lit( Bmc_LoadGetSatVar(p, 0), 1 );
    sat_solver_addclause( p.pSat, &Lit, &Lit + 1 );
    return p;
}
func Bmc_LoadStop(p *Bmc_Load_t) {
    Vec_IntFree( p.vSat2Id );
    sat_solver_delete( p.pSat );
    p = nil;
}

/**Function*************************************************************

  Synopsis    []

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Bmc_LoadTest(pGia *Gia_Man_t, fLoadCnf int32, fVerbose int32) {
    var nConfLimit int32;
    var p *Bmc_Load_t;
    var pObj *Gia_Obj_t;
    var i, status, Lit int32;
    var clk abctime  = Abc_Clock();
    // create the loading manager
    p = Bmc_LoadStart( pGia );
    // add callback for CNF loading
    if fLoadCnf != 0 {
        p.pSat.pCnfMan  = p;
        p.pSat.pCnfFunc = Bmc_LoadAddCnf;
    }
    // solve SAT problem for each PO
    // inlining of macro Gia_ManForEachPo
    for  i = 0; (i < Gia_ManPoNum(pGia)) && ((pObj) = Gia_ManCo(pGia, i)); i++  {
        if fLoadCnf != 0 {
            Lit = Abc_Var2Lit( Bmc_LoadGetSatVar(p, Gia_ObjFaninId0p(pGia, pObj)), Gia_ObjFaninC0(pObj) );
        } else {
            Lit = Abc_Var2Lit( Bmc_LoadAddCnf_rec(p, Gia_ObjFaninId0p(pGia, pObj)), Gia_ObjFaninC0(pObj) ); }
        if fVerbose != 0 {
            fmt.Printf("Frame%4d :  ", i);
            fmt.Printf("Vars = %6d  ", Vec_IntSize(p.vSat2Id));
            fmt.Printf("Clas = %6d  ", sat_solver_nclauses(p.pSat));
        }
        status = sat_solver_solve( p.pSat, &Lit, &Lit + 1, (ABC_INT64_T)nConfLimit, (ABC_INT64_T)0, (ABC_INT64_T)0, (ABC_INT64_T)0 );
        if fVerbose != 0 {
            fmt.Printf("Conf = %6d  ", sat_solver_nconflicts(p.pSat));
            if status == l_False {
                fmt.Printf("UNSAT  ");
            } else if status == l_True {
                fmt.Printf("SAT    ");
            } else { // if ( status == l_Undec )
                fmt.Printf("UNDEC  "); }
            Abc_PrintTime( 1, "Time", Abc_Clock() - clk );
        }
    }
    fmt.Printf("Callbacks = %d.  Loadings = %d.\n", p.nCallBacks1, p.nCallBacks2);
    Bmc_LoadStop( p );
}

////////////////////////////////////////////////////////////////////////
///                       END OF FILE                                ///
////////////////////////////////////////////////////////////////////////


//ABC_NAMESPACE_IMPL_END

