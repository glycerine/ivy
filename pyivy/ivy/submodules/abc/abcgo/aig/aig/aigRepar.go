package aig

import (
	"os"
	"time"
)

/**CFile****************************************************************

  FileName    [aigRepar.c]

  SystemName  [ABC: Logic synthesis and verification system.]

  PackageName [AIG package.]

  Synopsis    [Interpolation-based reparametrization.]

  Author      [Alan Mishchenko]
  
  Affiliation [UC Berkeley]

  Date        [Ver. 1.0. Started - April 28, 2007.]

  Revision    [$Id: aigRepar.c,v 1.00 2007/04/28 00:00:00 alanmi Exp $]

***********************************************************************/

//#include "aig.h"
//#include "sat/cnf/cnf.h"
//#include "sat/bsat/satSolver2.h"

//ABC_NAMESPACE_IMPL_START


////////////////////////////////////////////////////////////////////////
///                        DECLARATIONS                              ///
////////////////////////////////////////////////////////////////////////

////////////////////////////////////////////////////////////////////////
///                     FUNCTION DEFINITIONS                         ///
////////////////////////////////////////////////////////////////////////

/**Function*************************************************************

  Synopsis    [Adds buffer to the solver.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Aig_ManInterAddBuffer(pSat *sat_solver2, iVarA int32, iVarB int32, fCompl int32, fMark int32) {
    var Lits [2]lit;
    var Cid int32;
    assert(iVarA >= 0 && iVarB >= 0, "iVarA >= 0 && iVarB >= 0");

    Lits[0] = toLitCond( iVarA, 0 );
    Lits[1] = toLitCond( iVarB, !fCompl );
    Cid = sat_solver2_addclause( pSat, Lits, Lits + 2, 0 );
    if fMark != 0 {
        clause2_set_partA( pSat, Cid, 1 ); }

    Lits[0] = toLitCond( iVarA, 1 );
    Lits[1] = toLitCond( iVarB, fCompl );
    Cid = sat_solver2_addclause( pSat, Lits, Lits + 2, 0 );
    if fMark != 0 {
        clause2_set_partA( pSat, Cid, 1 ); }
}

/**Function*************************************************************

  Synopsis    [Adds constraints for the two-input AND-gate.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Aig_ManInterAddXor(pSat *sat_solver2, iVarA int32, iVarB int32, iVarC int32, fCompl int32, fMark int32) {
    var Lits [3]lit;
    var Cid int32;
    assert(iVarA >= 0 && iVarB >= 0 && iVarC >= 0, "iVarA >= 0 && iVarB >= 0 && iVarC >= 0");

    Lits[0] = toLitCond( iVarA, !fCompl );
    Lits[1] = toLitCond( iVarB, 1 );
    Lits[2] = toLitCond( iVarC, 1 );
    Cid = sat_solver2_addclause( pSat, Lits, Lits + 3, 0 );
    if fMark != 0 {
        clause2_set_partA( pSat, Cid, 1 ); }

    Lits[0] = toLitCond( iVarA, !fCompl );
    Lits[1] = toLitCond( iVarB, 0 );
    Lits[2] = toLitCond( iVarC, 0 );
    Cid = sat_solver2_addclause( pSat, Lits, Lits + 3, 0 );
    if fMark != 0 {
        clause2_set_partA( pSat, Cid, 1 ); }

    Lits[0] = toLitCond( iVarA, fCompl );
    Lits[1] = toLitCond( iVarB, 1 );
    Lits[2] = toLitCond( iVarC, 0 );
    Cid = sat_solver2_addclause( pSat, Lits, Lits + 3, 0 );
    if fMark != 0 {
        clause2_set_partA( pSat, Cid, 1 ); }

    Lits[0] = toLitCond( iVarA, fCompl );
    Lits[1] = toLitCond( iVarB, 0 );
    Lits[2] = toLitCond( iVarC, 1 );
    Cid = sat_solver2_addclause( pSat, Lits, Lits + 3, 0 );
    if fMark != 0 {
        clause2_set_partA( pSat, Cid, 1 ); }
}

/**Function*************************************************************

  Synopsis    []

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Aig_ManInterTest(pMan *Aig_Man_t, fVerbose int32) {
    var pSat *sat_solver2;
//    Aig_Man_t * pInter;
    var pInter *word;
    var vVars *Vec_Int_t;
    var pCnf *Cnf_Dat_t;
    var pObj *Aig_Obj_t;
    var Lit, Cid, Var, status, i int32;
    var clk int64  = time.Now().UnixNano();
    assert(Aig_ManRegNum(pMan) == 0, "Aig_ManRegNum(pMan) == 0");
    assert(Aig_ManCoNum(pMan) == 1, "Aig_ManCoNum(pMan) == 1");

    // derive CNFs
    pCnf = Cnf_Derive( pMan, 1 );

    // start the solver
    pSat = sat_solver2_new();
    sat_solver2_setnvars( pSat, 2*pCnf.nVars+1 );
    // set A-variables (all used except PI/PO)
    // inlining of macro Aig_ManForEachObj
    // inlining of macro Vec_PtrForEachEntry
    for  i = 0; (i < Vec_PtrSize(pMan.vObjs)) && (((pObj) = (Aig_Obj_t *)Vec_PtrEntry(pMan.vObjs, i)), 1); i++  { if (pObj) == nil {} else {
        if pCnf.pVarNums[pObj.Id] < 0 {
            continue; }
        if !Aig_ObjIsCi(pObj) && !Aig_ObjIsCo(pObj) {
            var_set_partA( pSat, pCnf.pVarNums[pObj.Id], 1 ); }
    }
    }

    // add clauses of A
    for  i = 0; i < pCnf.nClauses; i++  {
        Cid = sat_solver2_addclause( pSat, pCnf.pClauses[i], pCnf.pClauses[i+1], 0 );
        clause2_set_partA( pSat, Cid, 1 );
    }

    // add clauses of B
    Cnf_DataLift( pCnf, pCnf.nVars );
    for  i = 0; i < pCnf.nClauses; i++  {
        sat_solver2_addclause( pSat, pCnf.pClauses[i], pCnf.pClauses[i+1], 0 );
    }
    Cnf_DataLift( pCnf, -pCnf.nVars );

    // add PI equality clauses
    vVars = Vec_IntAlloc( Aig_ManCoNum(pMan)+1 );
    // inlining of macro Aig_ManForEachCi
    // inlining of macro Vec_PtrForEachEntry
    for  i = 0; (i < Vec_PtrSize(pMan.vCis)) && (((pObj) = (Aig_Obj_t *)Vec_PtrEntry(pMan.vCis, i)), 1); i++  {
        if Aig_ObjRefs(pObj) == 0 {
            continue; }
        Var = pCnf.pVarNums[pObj.Id];
        Aig_ManInterAddBuffer( pSat, Var, pCnf.nVars + Var, 0, 0 );
        Vec_IntPush( vVars, Var );
    }

    // add an XOR clause in the end
    Var = pCnf.pVarNums[Aig_ManCo(pMan,0).Id];
    Aig_ManInterAddXor( pSat, Var, pCnf.nVars + Var, 2*pCnf.nVars, 0, 0 );
    Vec_IntPush( vVars, Var );

    // solve the problem
    Lit = toLitCond( 2*pCnf.nVars, 0 );
    status = sat_solver2_solve( pSat, &Lit, &Lit + 1, 0, 0, 0, 0 );
    assert(status == l_False, "status == l_False");
    Sat_Solver2PrintStats( os.Stdout, pSat );

    // derive interpolant
//    pInter = Sat_ProofInterpolant( pSat, vVars );
//    Aig_ManPrintStats( pInter );
//    Aig_ManDumpBlif( pInter, "int.blif", NULL, NULL );
//pInter = Sat_ProofInterpolantTruth( pSat, vVars );
    pInter = nil;
//    Extra_PrintHex( stdout, pInter, Vec_IntSize(vVars) ); printf( "\n" );

    // clean up
//    Aig_ManStop( pInter );
    // inlining of macro ABC_FREE
    ((pInter) ? (free((char *) (pInter)), (pInter) = 0) : 0);

    Vec_IntFree( vVars );
    Cnf_DataFree( pCnf );
    sat_solver2_delete( pSat );
    // inlining of macro ABC_PRT
    Abc_Print(1, "%s =", ("Total interpolation time"));
Abc_Print(1, "%9.2f sec\n", 1.0*(float64((time.Now().UnixNano() - clk)))/(float64(int64(time.Second))));;
}

/**Function*************************************************************

  Synopsis    [Duplicates AIG while mapping PIs into the given array.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Aig_ManAppend(pBase *Aig_Man_t, pNew *Aig_Man_t) {
    var pObj *Aig_Obj_t;
    var i int32;
    assert(Aig_ManCoNum(pNew) == 1, "Aig_ManCoNum(pNew) == 1");
    assert(Aig_ManCiNum(pNew) == Aig_ManCiNum(pBase), "Aig_ManCiNum(pNew) == Aig_ManCiNum(pBase)");
    // create the PIs
    Aig_ManCleanData( pNew );
    Aig_ManConst1(pNew).pData = Aig_ManConst1(pBase);
    // inlining of macro Aig_ManForEachCi
    // inlining of macro Vec_PtrForEachEntry
    for  i = 0; (i < Vec_PtrSize(pNew.vCis)) && (((pObj) = (Aig_Obj_t *)Vec_PtrEntry(pNew.vCis, i)), 1); i++  {
        pObj.pData = Aig_IthVar(pBase, i);
    }
    // duplicate internal nodes
    // inlining of macro Aig_ManForEachNode
    // inlining of macro Vec_PtrForEachEntry
    for  i = 0; (i < Vec_PtrSize(pNew.vObjs)) && (((pObj) = (Aig_Obj_t *)Vec_PtrEntry(pNew.vObjs, i)), 1); i++  { if (pObj) == nil || !Aig_ObjIsNode(pObj) {} else {
        pObj.pData = Aig_And( pBase, Aig_ObjChild0Copy(pObj), Aig_ObjChild1Copy(pObj) ); }
    }
    // add one PO to base
    pObj = Aig_ManCo( pNew, 0 );
    Aig_ObjCreateCo( pBase, Aig_ObjChild0Copy(pObj) );
}

/**Function*************************************************************

  Synopsis    []

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Aig_ManInterRepar(pMan *Aig_Man_t, fVerbose int32) *Aig_Man_t {
    var pAigTemp *Aig_Man_t;
    var pInter *Aig_Man_t;
    var pBase *Aig_Man_t = nil;
    var pSat *sat_solver2;
    var vVars *Vec_Int_t;
    var pCnf, pCnfInter *Cnf_Dat_t;
    var pObj *Aig_Obj_t;
    var nOuts int32  = Aig_ManCoNum(pMan);
    var ShiftP [2]int32;
    var ShiftCnf [2]int32;
    var ShiftOr [2]int32;
    var ShiftAssume int32;
    var Cid, Lit, status, i, k, c int32;
    var clk int64  = time.Now().UnixNano();
    assert(Aig_ManRegNum(pMan) == 0, "Aig_ManRegNum(pMan) == 0");

    // derive CNFs
    pCnf = Cnf_Derive( pMan, nOuts );

    // start the solver
    pSat = sat_solver2_new();
    sat_solver2_setnvars( pSat, 4*pCnf.nVars + 6*nOuts );
    // vars: pGlobal + (p0 + A1 + A2 + or0) + (p1 + B1 + B2 + or1) + pAssume;
    ShiftP[0] = nOuts;
    ShiftP[1] = 2*pCnf.nVars + 3*nOuts;
    ShiftCnf[0] = ShiftP[0] + nOuts;
    ShiftCnf[1] = ShiftP[1] + nOuts;
    ShiftOr[0] = ShiftCnf[0] + 2*pCnf.nVars;
    ShiftOr[1] = ShiftCnf[1] + 2*pCnf.nVars;
    ShiftAssume = ShiftOr[1] + nOuts;
    assert(ShiftAssume + nOuts == pSat.size, "ShiftAssume + nOuts == pSat.size");

    // mark variables of A
    for  i = ShiftCnf[0]; i < ShiftP[1]; i++  {
        var_set_partA( pSat, i, 1 );
    }

    // add clauses of A, then B
    vVars = Vec_IntAlloc( 2*nOuts );
    for  k = 0; k < 2; k++  {
        // copy A1
        Cnf_DataLift( pCnf, ShiftCnf[k] );
        for  i = 0; i < pCnf.nClauses; i++  {
            Cid = sat_solver2_addclause( pSat, pCnf.pClauses[i], pCnf.pClauses[i+1], 0 );
            clause2_set_partA( pSat, Cid, k==0 );
        }
        // add equality p[k] == A1/B1
        // inlining of macro Aig_ManForEachCo
        // inlining of macro Vec_PtrForEachEntry
        for  i = 0; (i < Vec_PtrSize(pMan.vCos)) && (((pObj) = (Aig_Obj_t *)Vec_PtrEntry(pMan.vCos, i)), 1); i++  {
            Aig_ManInterAddBuffer( pSat, ShiftP[k] + i, pCnf.pVarNums[pObj.Id], k==1, k==0 );
        }
        // copy A2
        Cnf_DataLift( pCnf, pCnf.nVars );
        for  i = 0; i < pCnf.nClauses; i++  {
            Cid = sat_solver2_addclause( pSat, pCnf.pClauses[i], pCnf.pClauses[i+1], 0 );
            clause2_set_partA( pSat, Cid, k==0 );
        }
        // add comparator (!p[k] ^ A2/B2) == or[k]
        Vec_IntClear( vVars );
        // inlining of macro Aig_ManForEachCo
        // inlining of macro Vec_PtrForEachEntry
        for  i = 0; (i < Vec_PtrSize(pMan.vCos)) && (((pObj) = (Aig_Obj_t *)Vec_PtrEntry(pMan.vCos, i)), 1); i++  {
            Aig_ManInterAddXor( pSat, ShiftP[k] + i, pCnf.pVarNums[pObj.Id], ShiftOr[k] + i, k==1, k==0 );
            Vec_IntPush( vVars, toLitCond(ShiftOr[k] + i, 1) );
        }
        Cid = sat_solver2_addclause( pSat, Vec_IntArray(vVars), Vec_IntArray(vVars) + Vec_IntSize(vVars), 0 );
        clause2_set_partA( pSat, Cid, k==0 );
        // return to normal
        Cnf_DataLift( pCnf, -ShiftCnf[k]-pCnf.nVars );
    }
    // add clauses to constrain p[0] and p[1]
    for  k = 0; k < nOuts; k++  {
        Aig_ManInterAddXor( pSat, ShiftP[0] + k, ShiftP[1] + k, ShiftAssume + k, 0, 0 );
    }

    // start the interpolant
    pBase = Aig_ManStart( 1000 );
    pBase.pName = Abc_UtilStrsav( "repar" );
    for  k = 0; k < 2*nOuts; k++  {
        Aig_IthVar(pBase, i);
    }

    // start global variables (pGlobal and p[0])
    Vec_IntClear( vVars );
    for  k = 0; k < 2*nOuts; k++  {
        Vec_IntPush( vVars, k );
    }

    // perform iterative solving
    // vars: pGlobal + (p0 + A1 + A2 + or0) + (p1 + B1 + B2 + or1) + pAssume;
    for  k = 0; k < nOuts; k++  {
        // swap k-th variables
        var Temp int32  = Vec_IntEntry( vVars, k );
        Vec_IntWriteEntry( vVars, k, Vec_IntEntry(vVars, nOuts+k) );
        Vec_IntWriteEntry( vVars, nOuts+k, Temp );

        // solve incrementally 
        Lit = toLitCond( ShiftAssume + k, 1 ); // XOR output is 0 ==> p1 == p2
        status = sat_solver2_solve( pSat, &Lit, &Lit + 1, 0, 0, 0, 0 );
        assert(status == l_False, "status == l_False");
        Sat_Solver2PrintStats( os.Stdout, pSat );

        // derive interpolant
        pInter = (Aig_Man_t *)Sat_ProofInterpolant( pSat, vVars );
        Aig_ManPrintStats( pInter );
        // make sure interpolant does not depend on useless vars
        // inlining of macro Aig_ManForEachCi
        // inlining of macro Vec_PtrForEachEntry
        for  i = 0; (i < Vec_PtrSize(pInter.vCis)) && (((pObj) = (Aig_Obj_t *)Vec_PtrEntry(pInter.vCis, i)), 1); i++  {
            assert(i <= k || Aig_ObjRefs(pObj) == 0, "i <= k || Aig_ObjRefs(pObj) == 0");
        }

        // simplify
        pInter = Dar_ManRwsat( pAigTemp = pInter, 1, 0 );
        Aig_ManStop( pAigTemp );

        // add interpolant to the solver
        pCnfInter = Cnf_Derive( pInter, 1 );
        Cnf_DataLift( pCnfInter, pSat.size );
        sat_solver2_setnvars( pSat, pSat.size + 2*pCnfInter.nVars );
        for  i = 0; i < pCnfInter.nVars; i++  {
            var_set_partA( pSat, pSat.size-2*pCnfInter.nVars + i, 1 );
        }
        for  c = 0; c < 2; c++  {
            if c == 1 {
                Cnf_DataLift( pCnfInter, pCnfInter.nVars ); }
            // add to A
            for  i = 0; i < pCnfInter.nClauses; i++  {
                Cid = sat_solver2_addclause( pSat, pCnfInter.pClauses[i], pCnfInter.pClauses[i+1], 0 );
                clause2_set_partA( pSat, Cid, c==0 );
            }
            // connect to the inputs
            // inlining of macro Aig_ManForEachCi
            // inlining of macro Vec_PtrForEachEntry
            for  i = 0; (i < Vec_PtrSize(pInter.vCis)) && (((pObj) = (Aig_Obj_t *)Vec_PtrEntry(pInter.vCis, i)), 1); i++  {
                if i <= k {
                    Aig_ManInterAddBuffer( pSat, i, pCnf.pVarNums[pObj.Id], 0, c==0 ); }
            }
            // connect to the outputs
            pObj = Aig_ManCo(pInter, 0);
            Aig_ManInterAddBuffer( pSat, ShiftP[c] + k, pCnf.pVarNums[pObj.Id], 0, c==0 );
            if c == 1 {
                Cnf_DataLift( pCnfInter, -pCnfInter.nVars ); }
        }
        Cnf_DataFree( pCnfInter );

        // accumulate
        Aig_ManAppend( pBase, pInter );
        Aig_ManStop( pInter );

        // update global variables
        Temp = Vec_IntEntry( vVars, k );
        Vec_IntWriteEntry( vVars, k, Vec_IntEntry(vVars, nOuts+k) );
        Vec_IntWriteEntry( vVars, nOuts+k, Temp );
    }

    Vec_IntFree( vVars );
    Cnf_DataFree( pCnf );
    sat_solver2_delete( pSat );
    // inlining of macro ABC_PRT
    Abc_Print(1, "%s =", ("Reparameterization time"));
Abc_Print(1, "%9.2f sec\n", 1.0*(float64((time.Now().UnixNano() - clk)))/(float64(int64(time.Second))));;
    return pBase;
}

////////////////////////////////////////////////////////////////////////
///                       END OF FILE                                ///
////////////////////////////////////////////////////////////////////////


//ABC_NAMESPACE_IMPL_END

