package bmc

import "fmt"

/**CFile****************************************************************

  FileName    [bmcEnum.c]

  SystemName  [ABC: Logic synthesis and verification system.]

  PackageName [SAT-based bounded model checking.]

  Synopsis    [Enumeration.]

  Author      [Alan Mishchenko]
  
  Affiliation [UC Berkeley]

  Date        [Ver. 1.0. Started - June 20, 2005.]

  Revision    [$Id: bmcEnum.c,v 1.00 2005/06/20 00:00:00 alanmi Exp $]

***********************************************************************/

//#include "bmc.h"
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

  Synopsis    []

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Gia_ManDeriveOne(p *Gia_Man_t, vValues *Vec_Int_t) *Gia_Man_t {
    var pNew *Gia_Man_t;
    var pObj *Gia_Obj_t; var i, fPhase0, fPhase1 int32;
    assert(Vec_IntSize(vValues) == Gia_ManCiNum(p), "Vec_IntSize(vValues) == Gia_ManCiNum(p)");
    // propagate forward
    // inlining of macro Gia_ManForEachCi
    for i = 0; i < Vec_IntSize(p.vCis); i++ {
    	pObj = Gia_ManCi(p, i)
    	if pObj == nil { break }
        pObj.fPhase = Vec_IntEntry(vValues, i);
    }
    // inlining of macro Gia_ManForEachAnd
    for i = 0; i < p.nObjs; i++ {
    	pObj = Gia_ManObj(p, i)
    	if pObj == nil { break }
    	if !Gia_ObjIsAnd(pObj) {} else {
        fPhase0 = Gia_ObjPhase(Gia_ObjFanin0(pObj)) ^ Gia_ObjFaninC0(pObj);
        fPhase1 = Gia_ObjPhase(Gia_ObjFanin1(pObj)) ^ Gia_ObjFaninC1(pObj);
        pObj.fPhase = fPhase0 & fPhase1;
    }
    }
    // propagate backward
    Gia_ManCleanMark0(p);
    // inlining of macro Gia_ManForEachCo
    for i = 0; i < Vec_IntSize(p.vCos); i++ {
    	pObj = Gia_ManCo(p, i)
    	if pObj == nil { break }
        Gia_ObjFanin0(pObj).fMark0 = 1;
    }
    // inlining of macro Gia_ManForEachAndReverse
    for i = p.nObjs - 1; i > 0; i-- {
    	pObj = Gia_ManObj(p, i)
    	if pObj == nil { break }
    	if !Gia_ObjIsAnd(pObj) {} else {
        if !pObj.fMark0 {
            continue; }
        fPhase0 = Gia_ObjPhase(Gia_ObjFanin0(pObj)) ^ Gia_ObjFaninC0(pObj);
        fPhase1 = Gia_ObjPhase(Gia_ObjFanin1(pObj)) ^ Gia_ObjFaninC1(pObj);
        if fPhase0 == fPhase1 {
            assert(int32(pObj.fPhase) == fPhase0, "int32(pObj.fPhase) == fPhase0");
            Gia_ObjFanin0(pObj).fMark0 = 1;
            Gia_ObjFanin1(pObj).fMark0 = 1;
        
        } else if fPhase0 {
            assert(fPhase1 == 0, "fPhase1 == 0");
            assert(pObj.fPhase == 0, "pObj.fPhase == 0");
            Gia_ObjFanin1(pObj).fMark0 = 1;
        
        } else if fPhase1 != 0 {
            assert(fPhase0 == 0, "fPhase0 == 0");
            assert(pObj.fPhase == 0, "pObj.fPhase == 0");
            Gia_ObjFanin0(pObj).fMark0 = 1;
        }
    }
    }
    // create new
    Gia_ManFillValue( p );
    pNew = Gia_ManStart( Gia_ManObjNum(p) );
    pNew.pName = Abc_UtilStrsav( p.pName );
    pNew.pSpec = Abc_UtilStrsav( p.pSpec );
    Gia_ManConst0(p).Value = 0;
    // inlining of macro Gia_ManForEachCi
    for i = 0; i < Vec_IntSize(p.vCis); i++ {
    	pObj = Gia_ManCi(p, i)
    	if pObj == nil { break }
        pObj.Value = Abc_LitNotCond( Gia_ManAppendCi(pNew), !Vec_IntEntry(vValues, i) );
    }
    // inlining of macro Gia_ManForEachAnd
    for i = 0; i < p.nObjs; i++ {
    	pObj = Gia_ManObj(p, i)
    	if pObj == nil { break }
    	if !Gia_ObjIsAnd(pObj) {} else {
        if !pObj.fMark0 {
            continue; }
        fPhase0 = Gia_ObjPhase(Gia_ObjFanin0(pObj)) ^ Gia_ObjFaninC0(pObj);
        fPhase1 = Gia_ObjPhase(Gia_ObjFanin1(pObj)) ^ Gia_ObjFaninC1(pObj);
        if fPhase0 == fPhase1 {
            assert(int32(pObj.fPhase) == fPhase0, "int32(pObj.fPhase) == fPhase0");
            if pObj.fPhase {
                pObj.Value = Gia_ManAppendAnd( pNew, Gia_ObjFanin0(pObj).Value, Gia_ObjFanin1(pObj).Value );
            } else {
                pObj.Value = Gia_ManAppendOr( pNew, Gia_ObjFanin0(pObj).Value, Gia_ObjFanin1(pObj).Value ); }
        
        } else if fPhase0 {
            assert(fPhase1 == 0, "fPhase1 == 0");
            assert(pObj.fPhase == 0, "pObj.fPhase == 0");
            pObj.Value = Gia_ObjFanin1(pObj).Value;
        
        } else if fPhase1 != 0 {
            assert(fPhase0 == 0, "fPhase0 == 0");
            assert(pObj.fPhase == 0, "pObj.fPhase == 0");
            pObj.Value = Gia_ObjFanin0(pObj).Value;
        }
    }
    }
    // inlining of macro Gia_ManForEachCo
    for i = 0; i < Vec_IntSize(p.vCos); i++ {
    	pObj = Gia_ManCo(p, i)
    	if pObj == nil { break }
        Gia_ManAppendCo( pNew, Gia_ObjFanin0(pObj).Value );
    }
    Gia_ManSetRegNum( pNew, Gia_ManRegNum(p) );
    Gia_ManCleanMark0(p);
    return pNew;
}

/**Function*************************************************************

  Synopsis    []

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Gia_ManDeriveOneTest2(p *Gia_Man_t) *Gia_Man_t {
    var pNew *Gia_Man_t;
    var vValues *Vec_Int_t  = Vec_IntStart( Gia_ManCiNum(p) );
    //Vec_IntFill( vValues, Gia_ManCiNum(p), 1 );
    pNew = Gia_ManDeriveOne( p, vValues );
    Vec_IntFree( vValues );
    return pNew;
}

/**Function*************************************************************

  Synopsis    []

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Gia_ManDeriveOneTest(p *Gia_Man_t) {
    var fVerbose int32  = 1;
    var pNew *Gia_Man_t;
    var pObj, pRoot *Gia_Obj_t;
    var vValues *Vec_Int_t  = Vec_IntStart( Gia_ManCiNum(p) );
    var pCnf *Cnf_Dat_t  = (Cnf_Dat_t *)Mf_ManGenerateCnf( p, 8, 0, 0, 0, 0 );
    var i int32;
    var iVar int32;
    var nIter int32;
    var iPoVarBeg int32 = pCnf.nVars - Gia_ManCiNum(p);
    var pSat *sat_solver  = (sat_solver *)Cnf_DataWriteIntoSolver( pCnf, 1, 0 );
    var vLits *Vec_Int_t  = Vec_IntAlloc( 100 );
    Vec_IntPush( vLits, Abc_Var2Lit( 1, 1 ) );
    for  nIter = 0; nIter < 10000; nIter++  {
        var status int32  = sat_solver_solve( pSat, Vec_IntArray(vLits), Vec_IntLimit(vLits), 0, 0, 0, 0 );
        if status == l_False { 
            break; }
        // derive new set
        assert(status == l_True, "status == l_True");
        for  i = 0; i < Gia_ManCiNum(p); i++  {
            Vec_IntWriteEntry( vValues, i, sat_solver_var_value(pSat, iPoVarBeg+i) );
            fmt.Printf("%d", sat_solver_var_value(pSat, iPoVarBeg+i));
        }
        fmt.Printf(" : ");

        pNew = Gia_ManDeriveOne( p, vValues );
        // assign variables
        // inlining of macro Gia_ManForEachCi
        for i = 0; i < Vec_IntSize(pNew.vCis); i++ {
        	pObj = Gia_ManCi(pNew, i)
        	if pObj == nil { break }
            pObj.Value = iPoVarBeg+i;
        }
        iVar = sat_solver_nvars(pSat);
        // inlining of macro Gia_ManForEachAnd
        for i = 0; i < pNew.nObjs; i++ {
        	pObj = Gia_ManObj(pNew, i)
        	if pObj == nil { break }
        	if !Gia_ObjIsAnd(pObj) {} else {
            pObj.Value = iVar++; }
        }
        sat_solver_setnvars( pSat, iVar );
        // create new clauses
        // inlining of macro Gia_ManForEachAnd
        for i = 0; i < pNew.nObjs; i++ {
        	pObj = Gia_ManObj(pNew, i)
        	if pObj == nil { break }
        	if !Gia_ObjIsAnd(pObj) {} else {
            sat_solver_add_and( pSat, pObj.Value, Gia_ObjFanin0(pObj).Value, Gia_ObjFanin1(pObj).Value, Gia_ObjFaninC0(pObj), Gia_ObjFaninC1(pObj), 0 ); }
        }
        // add to the assumptions
        pRoot = Gia_ManCo(pNew, 0);
        Vec_IntPush( vLits, Abc_Var2Lit(Gia_ObjFanin0(pRoot).Value, !Gia_ObjFaninC0(pRoot)) );
        if fVerbose != 0 {
            fmt.Printf("Iter = %5d : ", nIter);
            fmt.Printf("And = %4d ", Gia_ManAndNum(pNew));
            fmt.Printf("\n");
        }
        Gia_ManStop( pNew );
    }
    Vec_IntFree( vLits );
    Vec_IntFree( vValues );
    Cnf_DataFree( pCnf );
    sat_solver_delete( pSat );
}

////////////////////////////////////////////////////////////////////////
///                       END OF FILE                                ///
////////////////////////////////////////////////////////////////////////


//ABC_NAMESPACE_IMPL_END

