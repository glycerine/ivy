package bsat

import "os"

/**CFile****************************************************************

  FileName    [satSolver2i.c]

  SystemName  [ABC: Logic synthesis and verification system.]

  PackageName [SAT solver.]

  Synopsis    [Records the trace of SAT solving in the CNF form.]

  Author      [Alan Mishchenko]
  
  Affiliation [UC Berkeley]

  Date        [Ver. 1.0. Started - September 2, 2013.]

  Revision    [$Id: satSolver2i.c,v 1.4 2013/09/02 00:00:00 casem Exp $]

***********************************************************************/

//#include "satSolver2.h"
//#include "aig/gia/gia.h"
//#include "aig/gia/giaAig.h"
//#include "sat/cnf/cnf.h"

//ABC_NAMESPACE_IMPL_START


////////////////////////////////////////////////////////////////////////
///                        DECLARATIONS                              ///
////////////////////////////////////////////////////////////////////////

type Int2_Man_t_ struct {
    pSat    *sat_solver2;      // user's SAT solver
    vGloVars      *Vec_Int_t;  // IDs of global variables
    vVar2Glo      *Vec_Int_t;  // mapping of SAT variables into their global IDs
    pGia      *Gia_Man_t;      // AIG manager to store the interpolant
}

////////////////////////////////////////////////////////////////////////
///                     FUNCTION DEFINITIONS                         ///
////////////////////////////////////////////////////////////////////////

/**Function*************************************************************

  Synopsis    [Managing interpolation manager.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Int2_ManStart(pSat *sat_solver2, pGloVars *int32, nGloVars int32) *Int2_Man_t {
    
    var i int32;
    // inlining of macro ABC_CALLOC
    p := &Int2_Man_t{};
    p.pSat     = pSat;
    p.vGloVars = Vec_IntAllocArrayCopy( pGloVars, nGloVars );
    p.vVar2Glo = Vec_IntInvert( p.vGloVars, -1 );
    p.pGia     = Gia_ManStart( 10 * Vec_IntSize(p.vGloVars) );
    p.pGia.pName = Abc_UtilStrsav( "interpolant" );
    for  i = 0; i < nGloVars; i++  {
        Gia_ManAppendCi( p.pGia );
    }
    Gia_ManHashStart( p.pGia );
    return p;
}
func Int2_ManStop(p *Int2_Man_t) {
    if p == nil {
        return; }
    Gia_ManStopP( &p.pGia );
    Vec_IntFree( p.vGloVars );
    Vec_IntFree( p.vVar2Glo );
    p = nil;
}
func Int2_ManReadInterpolant(pSat *sat_solver2) any {
    var p *Int2_Man_t  = pSat.pInt2;
    var pTemp *Gia_Man_t;
    var pGia *Gia_Man_t = p.pGia; p.pGia = nil;
    // return NULL, if the interpolant is not ready (for example, when the solver returned 'sat')
    if pSat.hProofLast == -1 {
        return nil; }
    // create AIG with one primary output
    assert(Gia_ManPoNum(pGia) == 0, "Gia_ManPoNum(pGia) == 0");
    Gia_ManAppendCo( pGia, pSat.hProofLast );  
    pSat.hProofLast = -1;
    // cleanup the resulting AIG
    pGia = Gia_ManCleanup( pTemp = pGia );
    Gia_ManStop( pTemp );
    return (void *)pGia;
}

/**Function*************************************************************

  Synopsis    [Computing interpolant for a clause.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Int2_ManChainStart(p *Int2_Man_t, c *clause) int32 {
    if c.lrn {
        return veci_begin(&p.pSat.claProofs)[clause_id(c)]; }
    if !c.partA {
        return 1; }
    if c.lits[c.size] < 0 {
        var i, Var, CiId, Res int32;
        for  i = 0; i < int32(c.size); i++  {
            // get ID of the global variable
            if Abc_Lit2Var(c.lits[i]) >= Vec_IntSize(p.vVar2Glo) {
                continue; }
            Var = Vec_IntEntry( p.vVar2Glo, Abc_Lit2Var(c.lits[i]) );
            if Var < 0 {
                continue; }
            // get literal of the AIG node
            CiId = Gia_ObjId( p.pGia, Gia_ManCi(p.pGia, Var) );
            // compute interpolant of the clause
            Res = Gia_ManHashOr( p.pGia, Res, Abc_Var2Lit(CiId, Abc_LitIsCompl(c.lits[i])) );
        }
        c.lits[c.size] = Res;
    }
    return c.lits[c.size];
}
func Int2_ManChainResolve(p *Int2_Man_t, c *clause, iLit int32, varA int32) int32 {
    var iLit2 int32  = Int2_ManChainStart( p, c );
    assert(iLit >= 0, "iLit >= 0");
    if varA != 0 {
        iLit = Gia_ManHashOr( p.pGia, iLit, iLit2 );
    } else {
        iLit = Gia_ManHashAnd( p.pGia, iLit, iLit2 ); }
    return iLit;
}

/**Function*************************************************************

  Synopsis    [Test for the interpolation procedure.]

  Description [The input AIG can be any n-input comb circuit with one PO 
  (not necessarily a comb miter).  The interpolant depends on n+1 variables
  and equal to the relation f = F(x0,x1,...,xn).]
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Gia_ManInterTest(p *Gia_Man_t) *Gia_Man_t {
    var pSat *sat_solver2;
    var pInter *Gia_Man_t;
    var pMan *Aig_Man_t;
    var vGVars *Vec_Int_t;
    var pCnf *Cnf_Dat_t;
    var pObj *Aig_Obj_t;
    var Lit, Cid, Var, status, i int32;
    var clk abctime  = Abc_Clock();
    assert(Gia_ManRegNum(p) == 0, "Gia_ManRegNum(p) == 0");
    assert(Gia_ManCoNum(p) == 1, "Gia_ManCoNum(p) == 1");

    // derive CNFs
    pMan = Gia_ManToAigSimple( p );
    pCnf = Cnf_Derive( pMan, 1 );

    // start the solver
    pSat = sat_solver2_new();
    pSat.fVerbose = 1;
    sat_solver2_setnvars( pSat, 2*pCnf.nVars+1 );

    // set A-variables (all used except PI/PO, which will be global variables)
    // inlining of macro Aig_ManForEachObj
    // inlining of macro Vec_PtrForEachEntry
    for  i = 0; (i < Vec_PtrSize(pMan.vObjs)) && (((pObj) = (Aig_Obj_t *)Vec_PtrEntry(pMan.vObjs, i)), 1); i++  { if (pObj) == nil {} else
        if pCnf.pVarNums[pObj.Id] >= 0 && !Aig_ObjIsCi(pObj) && !Aig_ObjIsCo(pObj) {
            var_set_partA( pSat, pCnf.pVarNums[pObj.Id], 1 ); }
    }

    // add clauses of A
    for  i = 0; i < pCnf.nClauses; i++  {
        Cid = sat_solver2_addclause( pSat, pCnf.pClauses[i], pCnf.pClauses[i+1], -1 );
        clause2_set_partA( pSat, Cid, 1 ); // this API should be called for each clause of A
    }

    // add clauses of B (after shifting all CNF variables by pCnf->nVars)
    Cnf_DataLift( pCnf, pCnf.nVars );
    for  i = 0; i < pCnf.nClauses; i++  {
        sat_solver2_addclause( pSat, pCnf.pClauses[i], pCnf.pClauses[i+1], -1 );
    }
    Cnf_DataLift( pCnf, -pCnf.nVars );

    // add PI equality clauses
    vGVars = Vec_IntAlloc( Aig_ManCoNum(pMan)+1 );
    // inlining of macro Aig_ManForEachCi
    // inlining of macro Vec_PtrForEachEntry
    for  i = 0; (i < Vec_PtrSize(pMan.vCis)) && (((pObj) = (Aig_Obj_t *)Vec_PtrEntry(pMan.vCis, i)), 1); i++  {
        Var = pCnf.pVarNums[pObj.Id];
        sat_solver2_add_buffer( pSat, Var, pCnf.nVars + Var, 0, 0, -1 );
        Vec_IntPush( vGVars, Var );
    }

    // add an XOR clause in the end
    Var = pCnf.pVarNums[Aig_ManCo(pMan,0).Id];
    sat_solver2_add_xor( pSat, Var, pCnf.nVars + Var, 2*pCnf.nVars, 0, 0, -1 );
    Vec_IntPush( vGVars, Var );

    // start the interpolation manager
    pSat.pInt2 = Int2_ManStart( pSat, Vec_IntArray(vGVars), Vec_IntSize(vGVars) );

    // solve the problem
    Lit = toLitCond( 2*pCnf.nVars, 0 );
    status = sat_solver2_solve( pSat, &Lit, &Lit + 1, 0, 0, 0, 0 );
    assert(status == l_False, "status == l_False");
    Sat_Solver2PrintStats( os.Stdout, pSat );

    // derive interpolant
    pInter = (Gia_Man_t *)Int2_ManReadInterpolant( pSat );
    Gia_ManPrintStats( pInter, nil );
    Abc_PrintTime( 1, "Total interpolation time", Abc_Clock() - clk );

    // clean up
    Vec_IntFree( vGVars );
    Cnf_DataFree( pCnf );
    Aig_ManStop( pMan );
    sat_solver2_delete( pSat );

    // return interpolant
    return pInter;
}

////////////////////////////////////////////////////////////////////////
///                       END OF FILE                                ///
////////////////////////////////////////////////////////////////////////


//ABC_NAMESPACE_IMPL_END

