package bmc

import "fmt"

import "os"

/**CFile****************************************************************

  FileName    [bmcEco.c]

  SystemName  [ABC: Logic synthesis and verification system.]

  PackageName [SAT-based bounded model checking.]

  Synopsis    []

  Author      [Alan Mishchenko]
  
  Affiliation [UC Berkeley]

  Date        [Ver. 1.0. Started - June 20, 2005.]

  Revision    [$Id: bmcEco.c,v 1.00 2005/06/20 00:00:00 alanmi Exp $]

***********************************************************************/

//#include "bmc.h"
//#include "sat/cnf/cnf.h"
//#include "sat/bsat/satSolver.h"
//#include "aig/gia/giaAig.h"

//ABC_NAMESPACE_IMPL_START


////////////////////////////////////////////////////////////////////////
///                        DECLARATIONS                              ///
////////////////////////////////////////////////////////////////////////

////////////////////////////////////////////////////////////////////////
///                     FUNCTION DEFINITIONS                         ///
////////////////////////////////////////////////////////////////////////

/**Function*************************************************************

  Synopsis    [Computes miter for ECO with given root node and fanins.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Bmc_EcoMiter(pGold *Gia_Man_t, pOld *Gia_Man_t, vFans *Vec_Int_t) *Gia_Man_t {
    var pNew, pTemp *Gia_Man_t;
    var pRoot *Gia_Obj_t  = Gia_ObjFanin0( Gia_ManPo(pOld, Gia_ManPoNum(pOld)-1) ); // fanin of the last PO
    var pObj *Gia_Obj_t;
    var i, NewPi, Miter int32;
    assert(Gia_ManCiNum(pGold) == Gia_ManCiNum(pOld), "Gia_ManCiNum(pGold) == Gia_ManCiNum(pOld)");
    assert(Gia_ManCoNum(pGold) == Gia_ManCoNum(pOld) - 1, "Gia_ManCoNum(pGold) == Gia_ManCoNum(pOld) - 1");
    assert(Gia_ObjIsAnd(pRoot), "Gia_ObjIsAnd(pRoot)");
    // create the miter
    pNew = Gia_ManStart( 3 * Gia_ManObjNum(pGold) );
    pNew.pName = Abc_UtilStrsav( pGold.pName );
    Gia_ManHashAlloc( pNew );
    // copy gold
    Gia_ManConst0(pGold).Value = 0;
    // inlining of macro Gia_ManForEachCi
    for  i = 0; (i < Vec_IntSize(pGold.vCis)) && ((pObj) = Gia_ManCi(pGold, i)); i++  {
        pObj.Value = Gia_ManAppendCi( pNew );
    }
    NewPi = Gia_ManAppendCi( pNew );
    // inlining of macro Gia_ManForEachAnd
    for  i = 0; (i < pGold.nObjs) && ((pObj) = Gia_ManObj(pGold, i)); i++  {      if !Gia_ObjIsAnd(pObj) {} else {
        pObj.Value = Gia_ManHashAnd( pNew, Gia_ObjFanin0Copy(pObj), Gia_ObjFanin1Copy(pObj) ); }
    }
    // inlining of macro Gia_ManForEachCo
    for  i = 0; (i < Vec_IntSize(pGold.vCos)) && ((pObj) = Gia_ManCo(pGold, i)); i++  {
        pObj.Value = Gia_ObjFanin0Copy( pObj );
    }
    // create onset
    Gia_ManConst0(pOld).Value = 0;
    // inlining of macro Gia_ManForEachCi
    for  i = 0; (i < Vec_IntSize(pOld.vCis)) && ((pObj) = Gia_ManCi(pOld, i)); i++  {
        pObj.Value = Gia_ManCi(pGold, i).Value;
    }
    // inlining of macro Gia_ManForEachAnd
    for  i = 0; (i < pOld.nObjs) && ((pObj) = Gia_ManObj(pOld, i)); i++  {      if !Gia_ObjIsAnd(pObj) {} else
        if pObj == pRoot {
            pObj.Value = NewPi;
        } else {
            pObj.Value = Gia_ManHashAnd( pNew, Gia_ObjFanin0Copy(pObj), Gia_ObjFanin1Copy(pObj) ); }
    }
    // inlining of macro Gia_ManForEachCo
    for  i = 0; (i < Vec_IntSize(pOld.vCos)) && ((pObj) = Gia_ManCo(pOld, i)); i++  {
        pObj.Value = Gia_ObjFanin0Copy( pObj );
    }
    // inlining of macro Gia_ManForEachCo
    for  i = 0; (i < Vec_IntSize(pGold.vCos)) && ((pObj) = Gia_ManCo(pGold, i)); i++  {
        Gia_ManAppendCo( pNew, Gia_ManHashXor(pNew, pObj.Value, Gia_ManCo(pOld, i).Value) );
    }
    // create offset
    // inlining of macro Gia_ManForEachAnd
    for  i = 0; (i < pOld.nObjs) && ((pObj) = Gia_ManObj(pOld, i)); i++  {      if !Gia_ObjIsAnd(pObj) {} else
        if pObj == pRoot {
            pObj.Value = Abc_LitNot(NewPi);
        } else {
            pObj.Value = Gia_ManHashAnd( pNew, Gia_ObjFanin0Copy(pObj), Gia_ObjFanin1Copy(pObj) ); }
    }
    // inlining of macro Gia_ManForEachCo
    for  i = 0; (i < Vec_IntSize(pOld.vCos)) && ((pObj) = Gia_ManCo(pOld, i)); i++  {
        pObj.Value = Gia_ObjFanin0Copy( pObj );
    }
    Miter = 0;
    // inlining of macro Gia_ManForEachCo
    for  i = 0; (i < Vec_IntSize(pGold.vCos)) && ((pObj) = Gia_ManCo(pGold, i)); i++  {
        Miter = Gia_ManHashOr( pNew, Miter, Gia_ManHashXor(pNew, pObj.Value, Gia_ManCo(pOld, i).Value) );
    }
    Gia_ManAppendCo( pNew, Miter );
    // add outputs for the nodes
    // inlining of macro Gia_ManForEachObjVec
    for  i = 0; (i < Vec_IntSize(vFans)) && ((pObj) = Gia_ManObj(pOld, Vec_IntEntry(vFans,i))); i++  {
        Gia_ManAppendCo( pNew, pObj.Value );
    }
    // cleanup
    pNew = Gia_ManCleanup( pTemp = pNew );
    Gia_ManStop( pTemp );
    assert(Gia_ManPiNum(pNew) == Gia_ManCiNum(pGold) + 1, "Gia_ManPiNum(pNew) == Gia_ManCiNum(pGold) + 1");
    assert(Gia_ManPoNum(pNew) == Gia_ManCoNum(pGold) + 1 + Vec_IntSize(vFans), "Gia_ManPoNum(pNew) == Gia_ManCoNum(pGold) + 1 + Vec_IntSize(vFans)");
    return pNew;
}

/**Function*************************************************************

  Synopsis    []

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
/**Function*************************************************************

  Synopsis    [Solve the enumeration problem.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Bmc_EcoSolve(pSat *sat_solver, Root int32, vVars *Vec_Int_t) int32 {
    var nBTLimit int32  = 1000000;
    var vLits *Vec_Int_t    = Vec_IntAlloc( Vec_IntSize(vVars) );
    var status int32;
    var i int32;
    var Div int32;
    var iVar int32;
    var nFinal int32;
    var pFinal *int32;
    var nIter int32;
    var RetValue int32;
    var pLits [2]int32;
    var nVars int32 = sat_solver_nvars( pSat );
    sat_solver_setnvars( pSat, nVars + 1 );
    pLits[0] = Abc_Var2Lit( Root, 0 );  // F = 1
    pLits[1] = Abc_Var2Lit( nVars, 0 ); // iNewLit
    for {
        // find onset minterm
        status = sat_solver_solve( pSat, pLits, pLits + 2, nBTLimit, 0, 0, 0 );
        if status == l_Undef { RetValue = -1; break; }
        if status == l_False { RetValue = 1; break; }
        assert(status == l_True, "status == l_True");
        // collect divisor literals
        Vec_IntClear( vLits );
        Vec_IntPush( vLits, Abc_LitNot(pLits[0]) ); // F = 0
        // inlining of macro Vec_IntForEachEntry
        for  i = 0; (i < Vec_IntSize(vVars)) && (((Div) = Vec_IntEntry(vVars, i)), 1); i++  {
            Vec_IntPush( vLits, sat_solver_var_literal(pSat, Div) );
        }
        // check against offset
        status = sat_solver_solve( pSat, Vec_IntArray(vLits), Vec_IntArray(vLits) + Vec_IntSize(vLits), nBTLimit, 0, 0, 0 );
        if status == l_Undef { RetValue = -1; break; }
        if status == l_True {
            break; }
        assert(status == l_False, "status == l_False");
        // compute cube and add clause
        nFinal = sat_solver_final( pSat, &pFinal );
        Vec_IntClear( vLits );
        Vec_IntPush( vLits, Abc_LitNot(pLits[1]) ); // NOT(iNewLit)
        fmt.Printf("Cube %d : ", nIter);
        for  i = 0; i < nFinal; i++  {
            if pFinal[i] == pLits[0] {
                continue; }
            Vec_IntPush( vLits, pFinal[i] );
            iVar = Vec_IntFind( vVars, Abc_Lit2Var(pFinal[i]) );   assert(iVar >= 0, "iVar >= 0");
            fmt.Printf("%s%d ", Abc_LitIsCompl(pFinal[i]) ? "+":"-", iVar);
        }
        fmt.Printf("\n");
        status = sat_solver_addclause( pSat, Vec_IntArray(vLits), Vec_IntArray(vLits) + Vec_IntSize(vLits) );
        assert((status != 0), "status");
        nIter++;
    }
//    assert( status == l_True );
    Vec_IntFree( vLits );
    return RetValue;
}

/**Function*************************************************************

  Synopsis    [Compute the patch function.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Bmc_EcoPatch(p *Gia_Man_t, nIns int32, nOuts int32) int32 {
    var i, Lit, RetValue, Root int32;
    var pObj *Gia_Obj_t;
    var vVars *Vec_Int_t;
    // generate CNF and solver
    var pCnf *Cnf_Dat_t  = Cnf_DeriveGiaRemapped( p );
    var pSat *sat_solver  = sat_solver_new();
    sat_solver_setnvars( pSat, pCnf.nVars );
    // add timeframe clauses
    for  i = 0; i < pCnf.nClauses; i++  {
        if !sat_solver_addclause( pSat, pCnf.pClauses[i], pCnf.pClauses[i+1] ) {
            assert(false, "0"); }
    }
    // add equality constraints
    assert(Gia_ManPoNum(p) == nOuts + 1 + nIns, "Gia_ManPoNum(p) == nOuts + 1 + nIns");
    // inlining of macro Gia_ManForEachPo
    for  i = 0; (i < Gia_ManPoNum(p)) && ((pObj) = Gia_ManCo(p, i)); i++  {
        if i == nOuts {
            break; }
        Lit = Abc_Var2Lit( pCnf.pVarNums[Gia_ObjId(p, pObj)], 1 ); // neg lit
        RetValue = sat_solver_addclause( pSat, &Lit, &Lit + 1 );
        assert((RetValue != 0), "RetValue");
    }
    // add inequality constraint
    pObj = Gia_ManPo( p, nOuts );
    Lit = Abc_Var2Lit( pCnf.pVarNums[Gia_ObjId(p, pObj)], 0 ); // pos lit
    RetValue = sat_solver_addclause( pSat, &Lit, &Lit + 1 );
    assert((RetValue != 0), "RetValue");
    // simplify the SAT solver
    RetValue = sat_solver_simplify( pSat );
    assert((RetValue != 0), "RetValue");
    // collect input variables
    vVars = Vec_IntAlloc( nIns );
    // inlining of macro Gia_ManForEachPo
    for  i = 0; (i < Gia_ManPoNum(p)) && ((pObj) = Gia_ManCo(p, i)); i++  {
        if i >= nOuts + 1 {
            Vec_IntPush( vVars, pCnf.pVarNums[Gia_ObjId(p, pObj)] ); }
    }
    assert(Vec_IntSize(vVars) == nIns, "Vec_IntSize(vVars) == nIns");
    // derive the root variable
    pObj = Gia_ManPi( p, Gia_ManPiNum(p) - 1 );
    Root = pCnf.pVarNums[Gia_ObjId(p, pObj)];
    // solve the problem
    RetValue = Bmc_EcoSolve( pSat, Root, vVars );
    Vec_IntFree( vVars );
    return RetValue;
}

/**Function*************************************************************

  Synopsis    [Tests the ECO miter.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Bmc_EcoMiterTest() {
    var pFileGold string  = "eco_gold.aig";
    var pFileOld string  =  "eco_old.aig";
    var vFans *Vec_Int_t;
    var pFile *os.File;
    var pMiter *Gia_Man_t;
    var pObj *Gia_Obj_t;
    var pGold *Gia_Man_t;
    var pOld *Gia_Man_t;
    var i, RetValue int32;
    // check that the files exist
    pFile = os.Open(pFileGold);
    if pFile == nil {
        fmt.Printf("File \"%s\" does not exist.\n", pFileGold);
        return;
    }
    _ = (pFile).Close();
    pFile = os.Open(pFileOld);
    if pFile == nil {
        fmt.Printf("File \"%s\" does not exist.\n", pFileOld);
        return;
    }
    _ = (pFile).Close();
    // read files
    pGold = Gia_AigerRead( pFileGold, 0, 0, 0 );
    pOld  = Gia_AigerRead( pFileOld, 0, 0, 0 );
    // create ECO miter
    vFans = Vec_IntAlloc( Gia_ManCiNum(pOld) );
    // inlining of macro Gia_ManForEachCi
    for  i = 0; (i < Vec_IntSize(pOld.vCis)) && ((pObj) = Gia_ManCi(pOld, i)); i++  {
        Vec_IntPush( vFans, Gia_ObjId(pOld, pObj) );
    }
    pMiter = Bmc_EcoMiter( pGold, pOld, vFans );
    Vec_IntFree( vFans );
    Gia_AigerWrite( pMiter, "eco_miter.aig", 0, 0, 0 );
    // find the patch
    RetValue = Bmc_EcoPatch( pMiter, Gia_ManCiNum(pGold), Gia_ManCoNum(pGold) );
    if RetValue == 1 {
        fmt.Printf("Patch is computed.\n"); }
    if RetValue == 0 {
        fmt.Printf("Cannot be patched.\n"); }
    if RetValue == -1 {
        fmt.Printf("Resource limit exceeded.\n"); }
    Gia_ManStop( pMiter );
}

////////////////////////////////////////////////////////////////////////
///                       END OF FILE                                ///
////////////////////////////////////////////////////////////////////////


//ABC_NAMESPACE_IMPL_END

