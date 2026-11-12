package fra

import "fmt"

/**CFile****************************************************************

  FileName    [fraClau.c]

  SystemName  [ABC: Logic synthesis and verification system.]

  PackageName [New FRAIG package.]

  Synopsis    [Induction with clause strengthening.]

  Author      [Alan Mishchenko]
  
  Affiliation [UC Berkeley]

  Date        [Ver. 1.0. Started - June 30, 2007.]

  Revision    [$Id: fraClau.c,v 1.00 2007/06/30 00:00:00 alanmi Exp $]

***********************************************************************/

//#include "fra.h"
//#include "sat/cnf/cnf.h"
//#include "sat/bsat/satSolver.h"

//ABC_NAMESPACE_IMPL_START


/*
    This code is inspired by the paper: Aaron Bradley and Zohar Manna, 
    "Checking safety by inductive generalization of counterexamples to 
    induction", FMCAD '07.
*/

////////////////////////////////////////////////////////////////////////
///                        DECLARATIONS                              ///
////////////////////////////////////////////////////////////////////////

      
type Cla_Man_t struct {
    // SAT solvers
    pSatMain      *sat_solver;
    pSatTest      *sat_solver;
    pSatBmc      *sat_solver;
    // CNF for the test solver
//    Cnf_Dat_t *      pCnfTest;
    // SAT variables
    vSatVarsMainCs       *Vec_Int_t;
    vSatVarsTestCs       *Vec_Int_t;
    vSatVarsTestNs       *Vec_Int_t;
    vSatVarsBmcNs       *Vec_Int_t;
    // helper variables
    nSatVarsTestBeg              int32;
    nSatVarsTestCur              int32;
    // counter-examples
    vCexMain0       *Vec_Int_t;
    vCexMain       *Vec_Int_t;
    vCexTest       *Vec_Int_t;
    vCexBase       *Vec_Int_t;
    vCexAssm       *Vec_Int_t;
    vCexBmc       *Vec_Int_t;
    // mapping of CS into NS var numbers
    pMapCsMainToCsTest             *int32; 
    pMapCsTestToCsMain             *int32; 
    pMapCsTestToNsTest             *int32; 
    pMapCsTestToNsBmc             *int32;  
}

////////////////////////////////////////////////////////////////////////
///                     FUNCTION DEFINITIONS                         ///
////////////////////////////////////////////////////////////////////////

/**Function*************************************************************

  Synopsis    [Saves variables corresponding to latch outputs.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Fra_ClauSaveLatchVars(pMan *Aig_Man_t, pCnf *Cnf_Dat_t, fCsVars int32) *Vec_Int_t {
    var vVars *Vec_Int_t;
    var pObjLo, pObjLi *Aig_Obj_t;
    var i int32;
    vVars = Vec_IntAlloc( Aig_ManRegNum(pMan) );
    // inlining of macro Aig_ManForEachLiLoSeq
    for  i = 0; (i < Aig_ManRegNum(pMan)) && (((pObjLi) = Aig_ManLi(pMan, i)), 1)            && (((pObjLo)=Aig_ManLo(pMan, i)), 1); i++  {
        Vec_IntPush( vVars, pCnf.pVarNums[fCsVars? pObjLo.Id : pObjLi.Id] );
    }
    return vVars;
}

/**Function*************************************************************

  Synopsis    [Saves variables corresponding to latch outputs.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Fra_ClauSaveOutputVars(pMan *Aig_Man_t, pCnf *Cnf_Dat_t) *Vec_Int_t {
    var vVars *Vec_Int_t;
    var pObj *Aig_Obj_t;
    var i int32;
    vVars = Vec_IntAlloc( Aig_ManCoNum(pMan) );
    // inlining of macro Aig_ManForEachCo
    // inlining of macro Vec_PtrForEachEntry
    for  i = 0; (i < Vec_PtrSize(pMan.vCos)) && (((pObj) = (Aig_Obj_t *)Vec_PtrEntry(pMan.vCos, i)), 1); i++  {
        Vec_IntPush( vVars, pCnf.pVarNums[pObj.Id] );
    }
    return vVars;
}

/**Function*************************************************************

  Synopsis    [Saves variables corresponding to latch outputs.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Fra_ClauSaveInputVars(pMan *Aig_Man_t, pCnf *Cnf_Dat_t, nStarting int32) *Vec_Int_t {
    var vVars *Vec_Int_t;
    var pObj *Aig_Obj_t;
    var i int32;
    vVars = Vec_IntAlloc( Aig_ManCiNum(pMan) - nStarting );
    // inlining of macro Aig_ManForEachCi
    // inlining of macro Vec_PtrForEachEntry
    for  i = 0; (i < Vec_PtrSize(pMan.vCis)) && (((pObj) = (Aig_Obj_t *)Vec_PtrEntry(pMan.vCis, i)), 1); i++  {
        if i < nStarting {
            continue; }
        Vec_IntPush( vVars, pCnf.pVarNums[pObj.Id] );
    }
    return vVars;
}

/**Function*************************************************************

  Synopsis    [Saves variables corresponding to latch outputs.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Fra_ClauCreateMapping(vSatVarsFrom *Vec_Int_t, vSatVarsTo *Vec_Int_t, nVarsMax int32) *int32 {
    var pMapping *int32;
    var Var int32;
    var i int32;
    assert(Vec_IntSize(vSatVarsFrom) == Vec_IntSize(vSatVarsTo), "Vec_IntSize(vSatVarsFrom) == Vec_IntSize(vSatVarsTo)");
    // inlining of macro ABC_ALLOC
    pMapping = ((int *) malloc(sizeof(int) * (size_t)(nVarsMax)));
    for  i = 0; i < nVarsMax; i++  {
        pMapping[i] = -1;
    }
    // inlining of macro Vec_IntForEachEntry
    for  i = 0; (i < Vec_IntSize(vSatVarsFrom)) && (((Var) = Vec_IntEntry(vSatVarsFrom, i)), 1); i++  {
        pMapping[Var] = Vec_IntEntry(vSatVarsTo,i);
    }
    return pMapping;
}


/**Function*************************************************************

  Synopsis    [Deletes the manager.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Fra_ClauStop(p *Cla_Man_t) {
    // inlining of macro ABC_FREE
    ((p.pMapCsMainToCsTest) ? (free((char *) (p.pMapCsMainToCsTest)), (p.pMapCsMainToCsTest) = 0) : 0);
    // inlining of macro ABC_FREE
    ((p.pMapCsTestToCsMain) ? (free((char *) (p.pMapCsTestToCsMain)), (p.pMapCsTestToCsMain) = 0) : 0);
    // inlining of macro ABC_FREE
    ((p.pMapCsTestToNsTest) ? (free((char *) (p.pMapCsTestToNsTest)), (p.pMapCsTestToNsTest) = 0) : 0);
    // inlining of macro ABC_FREE
    ((p.pMapCsTestToNsBmc) ? (free((char *) (p.pMapCsTestToNsBmc)), (p.pMapCsTestToNsBmc) = 0) : 0);
    Vec_IntFree( p.vSatVarsMainCs );
    Vec_IntFree( p.vSatVarsTestCs );
    Vec_IntFree( p.vSatVarsTestNs );
    Vec_IntFree( p.vSatVarsBmcNs );
    Vec_IntFree( p.vCexMain0 );
    Vec_IntFree( p.vCexMain );
    Vec_IntFree( p.vCexTest );
    Vec_IntFree( p.vCexBase );
    Vec_IntFree( p.vCexAssm );
    Vec_IntFree( p.vCexBmc  );
    if p.pSatMain { sat_solver_delete( p.pSatMain ); }
    if p.pSatTest { sat_solver_delete( p.pSatTest ); }
    if p.pSatBmc {  sat_solver_delete( p.pSatBmc ); }
    // inlining of macro ABC_FREE
    ((p) ? (free((char *) (p)), (p) = 0) : 0);
}

/**Function*************************************************************

  Synopsis    [Takes the AIG with the single output to be checked.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Fra_ClauStart(pMan *Aig_Man_t) *Cla_Man_t {
    var p *Cla_Man_t;
    var pCnfMain *Cnf_Dat_t;
    var pCnfTest *Cnf_Dat_t;
    var pCnfBmc *Cnf_Dat_t;
    var pFramesMain *Aig_Man_t;
    var pFramesTest *Aig_Man_t;
    var pFramesBmc *Aig_Man_t;
    assert(Aig_ManCoNum(pMan) - Aig_ManRegNum(pMan) == 1, "Aig_ManCoNum(pMan) - Aig_ManRegNum(pMan) == 1");

    // start the manager
    // inlining of macro ABC_ALLOC
    p = ((Cla_Man_t *) malloc(sizeof(Cla_Man_t) * (size_t)(1)));
    *p = Cla_Man_t{};
    p.vCexMain0 = Vec_IntAlloc( Aig_ManRegNum(pMan) );
    p.vCexMain  = Vec_IntAlloc( Aig_ManRegNum(pMan) );
    p.vCexTest  = Vec_IntAlloc( Aig_ManRegNum(pMan) );
    p.vCexBase  = Vec_IntAlloc( Aig_ManRegNum(pMan) );
    p.vCexAssm  = Vec_IntAlloc( Aig_ManRegNum(pMan) );
    p.vCexBmc   = Vec_IntAlloc( Aig_ManRegNum(pMan) );

    // derive two timeframes to be checked
    pFramesMain = Aig_ManFrames( pMan, 2, 0, 1, 0, 0, nil ); // nFrames, fInit, fOuts, fRegs
//Aig_ManShow( pFramesMain, 0, NULL );
    assert(Aig_ManCoNum(pFramesMain) == 2, "Aig_ManCoNum(pFramesMain) == 2");
    Aig_ObjChild0Flip( Aig_ManCo(pFramesMain, 0) ); // complement the first output
    pCnfMain = Cnf_DeriveSimple( pFramesMain, 0 );
//Cnf_DataWriteIntoFile( pCnfMain, "temp.cnf", 1 );
    p.pSatMain = (sat_solver *)Cnf_DataWriteIntoSolver( pCnfMain, 1, 0 );
/*
    {
        int i;
        Aig_Obj_t * pObj;
        Aig_ManForEachObj( pFramesMain, pObj, i )
            printf( "%d -> %d  \n", pObj->Id, pCnfMain->pVarNums[pObj->Id] );
        printf( "\n" );
    }
*/

    // derive one timeframe to be checked
    pFramesTest = Aig_ManFrames( pMan, 1, 0, 0, 1, 0, nil );
    assert(Aig_ManCoNum(pFramesTest) == Aig_ManRegNum(pMan), "Aig_ManCoNum(pFramesTest) == Aig_ManRegNum(pMan)");
    pCnfTest = Cnf_DeriveSimple( pFramesTest, Aig_ManRegNum(pMan) );
    p.pSatTest = (sat_solver *)Cnf_DataWriteIntoSolver( pCnfTest, 1, 0 );
    p.nSatVarsTestBeg = p.nSatVarsTestCur = sat_solver_nvars( p.pSatTest );

    // derive one timeframe to be checked for BMC
    pFramesBmc = Aig_ManFrames( pMan, 1, 1, 0, 1, 0, nil );
//Aig_ManShow( pFramesBmc, 0, NULL );
    assert(Aig_ManCoNum(pFramesBmc) == Aig_ManRegNum(pMan), "Aig_ManCoNum(pFramesBmc) == Aig_ManRegNum(pMan)");
    pCnfBmc = Cnf_DeriveSimple( pFramesBmc, Aig_ManRegNum(pMan) );
    p.pSatBmc = (sat_solver *)Cnf_DataWriteIntoSolver( pCnfBmc, 1, 0 );

    // create variable sets
    p.vSatVarsMainCs = Fra_ClauSaveInputVars( pFramesMain, pCnfMain, 2 * (Aig_ManCiNum(pMan)-Aig_ManRegNum(pMan)) );
    p.vSatVarsTestCs = Fra_ClauSaveLatchVars( pFramesTest, pCnfTest, 1 );
    p.vSatVarsTestNs = Fra_ClauSaveLatchVars( pFramesTest, pCnfTest, 0 );
    p.vSatVarsBmcNs  = Fra_ClauSaveOutputVars( pFramesBmc, pCnfBmc );
    assert(Vec_IntSize(p.vSatVarsTestCs) == Vec_IntSize(p.vSatVarsMainCs), "Vec_IntSize(p.vSatVarsTestCs) == Vec_IntSize(p.vSatVarsMainCs)");
    assert(Vec_IntSize(p.vSatVarsTestCs) == Vec_IntSize(p.vSatVarsBmcNs), "Vec_IntSize(p.vSatVarsTestCs) == Vec_IntSize(p.vSatVarsBmcNs)");

    // create mapping of CS into NS vars
    p.pMapCsMainToCsTest = Fra_ClauCreateMapping( p.vSatVarsMainCs, p.vSatVarsTestCs, Aig_ManObjNumMax(pFramesMain) );
    p.pMapCsTestToCsMain = Fra_ClauCreateMapping( p.vSatVarsTestCs, p.vSatVarsMainCs, Aig_ManObjNumMax(pFramesTest) );
    p.pMapCsTestToNsTest = Fra_ClauCreateMapping( p.vSatVarsTestCs, p.vSatVarsTestNs, Aig_ManObjNumMax(pFramesTest) );
    p.pMapCsTestToNsBmc  = Fra_ClauCreateMapping( p.vSatVarsTestCs, p.vSatVarsBmcNs,  Aig_ManObjNumMax(pFramesTest) );

    // cleanup
    Cnf_DataFree( pCnfMain );
    Cnf_DataFree( pCnfTest );
    Cnf_DataFree( pCnfBmc );
    Aig_ManStop( pFramesMain );
    Aig_ManStop( pFramesTest );
    Aig_ManStop( pFramesBmc );
    if p.pSatMain == nil || p.pSatTest == nil || p.pSatBmc == nil {
        Fra_ClauStop( p );
        return nil;
    }
    return p;
}

/**Function*************************************************************

  Synopsis    [Splits off second half and returns it as a new vector.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Vec_IntSplitHalf(vVec *Vec_Int_t) *Vec_Int_t {
    var vPart *Vec_Int_t;
    var Entry, i int32;
    assert(Vec_IntSize(vVec) > 1, "Vec_IntSize(vVec) > 1");
    vPart = Vec_IntAlloc( Vec_IntSize(vVec) / 2 + 1 );
    // inlining of macro Vec_IntForEachEntryStart
    for  i = Vec_IntSize(vVec) / 2; (i < Vec_IntSize(vVec)) && (((Entry) = Vec_IntEntry(vVec, i)), 1); i++  {
        Vec_IntPush( vPart, Entry );
    }
    Vec_IntShrink( vVec, Vec_IntSize(vVec) / 2 );
    return vPart;
}

/**Function*************************************************************

  Synopsis    [Complements all literals in the clause.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Vec_IntComplement(vVec *Vec_Int_t) {
    var i int32;
    for  i = 0; i < Vec_IntSize(vVec); i++  {
        vVec.pArray[i] = lit_neg( vVec.pArray[i] );
    }
}

/**Function*************************************************************

  Synopsis    [Checks if the property holds. Returns counter-example if not.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Fra_ClauCheckProperty(p *Cla_Man_t, vCex *Vec_Int_t) int32 {
    var nBTLimit int32;
    var RetValue, iVar, i int32;
    sat_solver_act_var_clear( p.pSatMain );
    RetValue = sat_solver_solve( p.pSatMain, nil, nil, (ABC_INT64_T)nBTLimit, (ABC_INT64_T)0, (ABC_INT64_T)0, (ABC_INT64_T)0 );
    Vec_IntClear( vCex );
    if RetValue == l_False {
        return 1; }
    assert(RetValue == l_True, "RetValue == l_True");
    // inlining of macro Vec_IntForEachEntry
    for  i = 0; (i < Vec_IntSize(p.vSatVarsMainCs)) && (((iVar) = Vec_IntEntry(p.vSatVarsMainCs, i)), 1); i++  {
        Vec_IntPush( vCex, sat_solver_var_literal(p.pSatMain, iVar) );
    }
/*
    {
        int i;
        for (i = 0; i < p->pSatMain->size; i++)
            printf( "%d=%d ", i, p->pSatMain->model.ptr[i] == l_True );
        printf( "\n" );
    }
*/
    return 0;
}

/**Function*************************************************************

  Synopsis    [Checks if the clause holds using BMC.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Fra_ClauCheckBmc(p *Cla_Man_t, vClause *Vec_Int_t) int32 {
    var nBTLimit int32;
    var RetValue int32;
    RetValue = sat_solver_solve( p.pSatBmc, Vec_IntArray(vClause), Vec_IntArray(vClause) + Vec_IntSize(vClause), 
        (ABC_INT64_T)nBTLimit, (ABC_INT64_T)0, (ABC_INT64_T)0, (ABC_INT64_T)0 );
    if RetValue == l_False {
        return 1; }
    assert(RetValue == l_True, "RetValue == l_True");
    return 0;
}

/**Function*************************************************************

  Synopsis    [Lifts the clause to depend on NS variables.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Fra_ClauRemapClause(pMap *int32, vClause *Vec_Int_t, vRemapped *Vec_Int_t, fInv int32) {
    var iLit, i int32;
    Vec_IntClear( vRemapped );
    // inlining of macro Vec_IntForEachEntry
    for  i = 0; (i < Vec_IntSize(vClause)) && (((iLit) = Vec_IntEntry(vClause, i)), 1); i++  {
        assert(pMap[lit_var(iLit)] >= 0, "pMap[lit_var(iLit)] >= 0");
        iLit = toLitCond( pMap[lit_var(iLit)], lit_sign(iLit) ^ fInv );
        Vec_IntPush( vRemapped, iLit );
    }
}

/**Function*************************************************************

  Synopsis    [Checks if the clause holds. Returns counter example if not.]

  Description [Uses test SAT solver.]
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Fra_ClauCheckClause(p *Cla_Man_t, vClause *Vec_Int_t, vCex *Vec_Int_t) int32 {
    var nBTLimit int32;
    var RetValue, iVar, i int32;
    // complement literals
    Vec_IntPush( vClause, toLit( p.nSatVarsTestCur++ ) ); // helper positive
    Vec_IntComplement( vClause ); // helper negative (the clause is C v h')
    // add the clause
    RetValue = sat_solver_addclause( p.pSatTest, Vec_IntArray(vClause), Vec_IntArray(vClause) + Vec_IntSize(vClause) );
    assert(RetValue == 1, "RetValue == 1");
    // complement all literals
    Vec_IntPop( vClause );  // helper removed
    Vec_IntComplement( vClause ); 
    // create the assumption in terms of NS variables
    Fra_ClauRemapClause( p.pMapCsTestToNsTest, vClause, p.vCexAssm, 0 );
    // add helper literals
    for  i = p.nSatVarsTestBeg; i < p.nSatVarsTestCur - 1; i++  {
        Vec_IntPush( p.vCexAssm, toLitCond(i,1) );
    } // other helpers negative
    Vec_IntPush( p.vCexAssm, toLitCond(i,0) ); // positive helper
    // try to solve
    RetValue = sat_solver_solve( p.pSatTest, Vec_IntArray(p.vCexAssm), Vec_IntArray(p.vCexAssm) + Vec_IntSize(p.vCexAssm), 
        (ABC_INT64_T)nBTLimit, (ABC_INT64_T)0, (ABC_INT64_T)0, (ABC_INT64_T)0 );
    if vCex != nil {
        Vec_IntClear( vCex ); }
    if RetValue == l_False {
        return 1; }
    assert(RetValue == l_True, "RetValue == l_True");
    if vCex != nil {
        // inlining of macro Vec_IntForEachEntry
        for  i = 0; (i < Vec_IntSize(p.vSatVarsTestCs)) && (((iVar) = Vec_IntEntry(p.vSatVarsTestCs, i)), 1); i++  {
            Vec_IntPush( vCex, sat_solver_var_literal(p.pSatTest, iVar) );
        }
    }
    return 0;
}

/**Function*************************************************************

  Synopsis    [Reduces the counter-example by removing complemented literals.]

  Description [Removes literals from vMain that differ from those in the  
  counter-example (vNew). Relies on the fact that the PI variables are
  assigned in the increasing order.]
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Fra_ClauReduceClause(vMain *Vec_Int_t, vNew *Vec_Int_t) {
    var LitM, LitN, VarM, VarN, i, j, k int32;
    assert(Vec_IntSize(vMain) <= Vec_IntSize(vNew), "Vec_IntSize(vMain) <= Vec_IntSize(vNew)");
    for  i = j = k = 0; i < Vec_IntSize(vMain) && j < Vec_IntSize(vNew);  {
        LitM = Vec_IntEntry( vMain, i );
        LitN = Vec_IntEntry( vNew, j );
        VarM = lit_var( LitM );
        VarN = lit_var( LitN );
        if VarM < VarN {
            assert(false, "0");
        
        } else if VarM > VarN {
            j++;
        
        } else { // if ( VarM == VarN )
        
            i++;
            j++;
            if LitM == LitN {
                Vec_IntWriteEntry( vMain, k++, LitM ); }
        }
    }
    assert(i == Vec_IntSize(vMain), "i == Vec_IntSize(vMain)");
    Vec_IntShrink( vMain, k );
}

/**Function*************************************************************

  Synopsis    [Computes the minimal invariant that holds.]

  Description [On entrace, vBasis does not hold, vBasis+vExtra holds but
  is not minimal. On exit, vBasis is unchanged, vBasis+vExtra is minimal.]
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Fra_ClauMinimizeClause_rec(p *Cla_Man_t, vBasis *Vec_Int_t, vExtra *Vec_Int_t) {
    var vExtra2 *Vec_Int_t;
    var nSizeOld int32;
    if Vec_IntSize(vExtra) == 1 {
        return; }
    nSizeOld = Vec_IntSize( vBasis );
    vExtra2 = Vec_IntSplitHalf( vExtra );

    // try the first half
    Vec_IntAppend( vBasis, vExtra );
    if Fra_ClauCheckClause( p, vBasis, nil ) != 0 {
        Vec_IntShrink( vBasis, nSizeOld );
        Fra_ClauMinimizeClause_rec( p, vBasis, vExtra );
        return;
    }
    Vec_IntShrink( vBasis, nSizeOld );

    // try the second half
    Vec_IntAppend( vBasis, vExtra2 );
    if Fra_ClauCheckClause( p, vBasis, nil ) != 0 {
        Vec_IntShrink( vBasis, nSizeOld );
        Fra_ClauMinimizeClause_rec( p, vBasis, vExtra2 );
        return;
    }
//    Vec_IntShrink( vBasis, nSizeOld );

    // find the smallest with the second half added
    Fra_ClauMinimizeClause_rec( p, vBasis, vExtra );
    Vec_IntShrink( vBasis, nSizeOld );
    Vec_IntAppend( vBasis, vExtra );
    // find the smallest with the second half added
    Fra_ClauMinimizeClause_rec( p, vBasis, vExtra2 );
    Vec_IntShrink( vBasis, nSizeOld );
    Vec_IntAppend( vExtra, vExtra2 );
    Vec_IntFree( vExtra2 );
}

/**Function*************************************************************

  Synopsis    [Minimizes the clauses using a simple method.]

  Description [The input and output clause are in vExtra.]
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Fra_ClauMinimizeClause(p *Cla_Man_t, vBasis *Vec_Int_t, vExtra *Vec_Int_t) {
    var iLit, iLit2, i, k int32;
    // inlining of macro Vec_IntForEachEntryReverse
    for  i = Vec_IntSize(vExtra) - 1; (i >= 0) && (((iLit) = Vec_IntEntry(vExtra, i)), 1); i--  {
        // copy literals without the given one
        Vec_IntClear( vBasis );
        // inlining of macro Vec_IntForEachEntry
        for  k = 0; (k < Vec_IntSize(vExtra)) && (((iLit2) = Vec_IntEntry(vExtra, k)), 1); k++  {
            if k != i {
                Vec_IntPush( vBasis, iLit2 ); }
        }
        // try whether it is inductive
        if !(Fra_ClauCheckClause( p, vBasis, nil ) != 0) {
            continue; }
        // the clause is inductive
        // remove the literal
        for  k = i; k < Vec_IntSize(vExtra)-1; k++  {
            Vec_IntWriteEntry( vExtra, k, Vec_IntEntry(vExtra,k+1) );
        }
        Vec_IntShrink( vExtra, Vec_IntSize(vExtra)-1 );
    }
}

/**Function*************************************************************

  Synopsis    [Prints the clause.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Fra_ClauPrintClause(vSatCsVars *Vec_Int_t, vCex *Vec_Int_t) {
    var LitM, VarM, VarN, i, j, k int32;
    assert(Vec_IntSize(vCex) <= Vec_IntSize(vSatCsVars), "Vec_IntSize(vCex) <= Vec_IntSize(vSatCsVars)");
    for  i = j = k = 0; i < Vec_IntSize(vCex) && j < Vec_IntSize(vSatCsVars);  {
        LitM = Vec_IntEntry( vCex, i );
        VarM = lit_var( LitM );
        VarN = Vec_IntEntry( vSatCsVars, j );
        if VarM < VarN {
            assert(false, "0");
        
        } else if VarM > VarN {
            j++;
            fmt.Printf("-");
        
        } else { // if ( VarM == VarN )
        
            i++;
            j++;
            fmt.Printf("%d", !lit_sign(LitM));
        }
    }
    assert(i == Vec_IntSize(vCex), "i == Vec_IntSize(vCex)");
}

/**Function*************************************************************

  Synopsis    [Takes the AIG with the single output to be checked.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Fra_Clau(pMan *Aig_Man_t, nIters int32, fVerbose int32, fVeryVerbose int32) int32 {
    var p *Cla_Man_t;
    var Iter, RetValue, fFailed, i int32;
    assert(Aig_ManCoNum(pMan) - Aig_ManRegNum(pMan) == 1, "Aig_ManCoNum(pMan) - Aig_ManRegNum(pMan) == 1");
    // create the manager
    p = Fra_ClauStart( pMan );
    if p == nil {
        fmt.Printf("The property is trivially inductive.\n");
        return 1;
    }
    // generate counter-examples and expand them
    for  Iter = 0; !(Fra_ClauCheckProperty( p, p.vCexMain0 ) != 0) && Iter < nIters; Iter++  {
        if fVerbose != 0 {
            fmt.Printf("%4d : ", Iter); }
        // remap clause into the test manager
        Fra_ClauRemapClause( p.pMapCsMainToCsTest, p.vCexMain0, p.vCexMain, 0 );
        if (fVerbose != 0) && (fVeryVerbose != 0) {
            Fra_ClauPrintClause( p.vSatVarsTestCs, p.vCexMain ); }
        // the main counter-example is in p->vCexMain
        // intermediate counter-examples are in p->vCexTest
        // generate the reduced counter-example to the inductive property
        fFailed = 0;
        for  i = 0; !(Fra_ClauCheckClause( p, p.vCexMain, p.vCexTest ) != 0); i++  {
            Fra_ClauReduceClause( p.vCexMain, p.vCexTest );
            Fra_ClauRemapClause( p.pMapCsTestToNsBmc, p.vCexMain, p.vCexBmc, 0 );

//            if ( !Fra_ClauCheckBmc(p, p->vCexBmc) )
            if Vec_IntSize(p.vCexMain) < 1 {
                Vec_IntComplement( p.vCexMain0 ); 
                RetValue = sat_solver_addclause( p.pSatMain, Vec_IntArray(p.vCexMain0), Vec_IntArray(p.vCexMain0) + Vec_IntSize(p.vCexMain0) );
                if RetValue == 0 {
                    fmt.Printf("\nProperty is proved after %d iterations.\n", Iter+1);
                    return 0;
                }
                fFailed = 1;
                break;
            }
        }
        if fFailed {
            if fVerbose != 0 {
                fmt.Printf(" Reducing failed after %d iterations (BMC failed).\n", i); }
            continue;
        }
        if Vec_IntSize(p.vCexMain) == 0 {
            if fVerbose != 0 {
                fmt.Printf(" Reducing failed after %d iterations (nothing left).\n", i); }
            continue;
        }
        if fVerbose != 0 {
            fmt.Printf("  "); }
        if (fVerbose != 0) && (fVeryVerbose != 0) {
            Fra_ClauPrintClause( p.vSatVarsTestCs, p.vCexMain ); }
        if fVerbose != 0 {
            fmt.Printf(" LitsInd = %3d.  ", Vec_IntSize(p.vCexMain)); }
        // minimize the inductive property
        Vec_IntClear( p.vCexBase );
        if Vec_IntSize(p.vCexMain) > 1 {
//        Fra_ClauMinimizeClause_rec( p, p->vCexBase, p->vCexMain );
            Fra_ClauMinimizeClause( p, p.vCexBase, p.vCexMain ); }
        assert(Vec_IntSize(p.vCexMain) > 0, "Vec_IntSize(p.vCexMain) > 0");
        if (fVerbose != 0) && (fVeryVerbose != 0) {
            Fra_ClauPrintClause( p.vSatVarsTestCs, p.vCexMain ); }
        if fVerbose != 0 {
            fmt.Printf(" LitsRed = %3d.  ", Vec_IntSize(p.vCexMain)); }
        if fVerbose != 0 {
            fmt.Printf("\n"); }
        // add the clause to the solver
        Fra_ClauRemapClause( p.pMapCsTestToCsMain, p.vCexMain, p.vCexAssm, 1 );
        RetValue = sat_solver_addclause( p.pSatMain, Vec_IntArray(p.vCexAssm), Vec_IntArray(p.vCexAssm) + Vec_IntSize(p.vCexAssm) );
        if RetValue == 0 {
            Iter++;
            break;
        }
        if p.pSatMain.qtail != p.pSatMain.qhead {
            RetValue = sat_solver_simplify(p.pSatMain);
            assert(RetValue != 0, "RetValue != 0");
            assert(p.pSatMain.qtail == p.pSatMain.qhead, "p.pSatMain.qtail == p.pSatMain.qhead");
        }
    }

    // report the results
    if Iter == nIters {
        fmt.Printf("Property is not proved after %d iterations.\n", nIters);
        return 0;
    }
    fmt.Printf("Property is proved after %d iterations.\n", Iter);
    Fra_ClauStop( p );
    return 1;
}


////////////////////////////////////////////////////////////////////////
///                       END OF FILE                                ///
////////////////////////////////////////////////////////////////////////


//ABC_NAMESPACE_IMPL_END

