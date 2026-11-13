package bmc

import (
	"fmt"
	"time"
)

import "os"

/**CFile****************************************************************

  FileName    [bmcFault.c]

  SystemName  [ABC: Logic synthesis and verification system.]

  PackageName [SAT-based bounded model checking.]

  Synopsis    [Checking for functional faults.]

  Author      [Alan Mishchenko]
  
  Affiliation [UC Berkeley]

  Date        [Ver. 1.0. Started - June 20, 2005.]

  Revision    [$Id: bmcFault.c,v 1.00 2005/06/20 00:00:00 alanmi Exp $]

***********************************************************************/

//#include "bmc.h"
//#include "sat/cnf/cnf.h"
//#include "sat/bsat/satStore.h"
//#include "aig/gia/giaAig.h"
//#include "misc/extra/extra.h"

//ABC_NAMESPACE_IMPL_START


////////////////////////////////////////////////////////////////////////
///                        DECLARATIONS                              ///
////////////////////////////////////////////////////////////////////////

const FFTEST_MAX_VARS = 2
const FFTEST_MAX_PARS = 8

////////////////////////////////////////////////////////////////////////
///                     FUNCTION DEFINITIONS                         ///
////////////////////////////////////////////////////////////////////////

/**Function*************************************************************

  Synopsis    [This procedure sets default parameters.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Gia_DeriveFormula_rec(pGia *Gia_Man_t, ppNamesIn *string, vStr *Vec_Str_t, iLit int32) {
    var pObj *Gia_Obj_t  = Gia_ManObj( pGia, Abc_Lit2Var(iLit) );
    var fCompl int32  = Abc_LitIsCompl(iLit);
    if Gia_ObjIsAnd(pObj) {
        Vec_StrPush( vStr, '(' );
        if Gia_ObjIsMux(pGia, pObj) {
            Gia_DeriveFormula_rec( pGia, ppNamesIn, vStr, Gia_ObjFaninLit0p(pGia, pObj) );
            Vec_StrPush( vStr, '?' );
            Gia_DeriveFormula_rec( pGia, ppNamesIn, vStr, Abc_LitNotCond( Gia_ObjFaninLit1p(pGia, pObj), fCompl ) );
            Vec_StrPush( vStr, ':' );
            Gia_DeriveFormula_rec( pGia, ppNamesIn, vStr, Abc_LitNotCond( Gia_ObjFaninLit2p(pGia, pObj), fCompl ) );
        
        } else {
            Gia_DeriveFormula_rec( pGia, ppNamesIn, vStr, Abc_LitNotCond( Gia_ObjFaninLit0p(pGia, pObj), fCompl ) );
            Vec_StrPush( vStr, byte((Gia_ObjIsXor(pObj) ? '^' : byte((fCompl ? '|' : '&')))) );
            Gia_DeriveFormula_rec( pGia, ppNamesIn, vStr, Abc_LitNotCond( Gia_ObjFaninLit1p(pGia, pObj), fCompl ) );
        }
        Vec_StrPush( vStr, ')' );
    
    } else {
        if fCompl != 0 { Vec_StrPush( vStr, '~' ); }
        Vec_StrPrintF( vStr, "%s", ppNamesIn[Gia_ObjCioId(pObj)] );
    }
}
func Gia_DeriveFormula(pGia *Gia_Man_t, ppNamesIn *string) string {
    var pResult string;
    var vStr *Vec_Str_t    = Vec_StrAlloc( 1000 );
    var pMuxes *Gia_Man_t  = Gia_ManDupMuxes( pGia, 2 );
    var pObj *Gia_Obj_t    = Gia_ManCo( pGia, 0 );
    Vec_StrPush( vStr, '(' );
    Gia_DeriveFormula_rec( pGia, ppNamesIn, vStr, Gia_ObjFaninLit0p(pGia, pObj) );
    Vec_StrPush( vStr, ')' );
    Vec_StrPush( vStr, '\0' );
    Gia_ManStop( pMuxes );
    pResult = Vec_StrReleaseArray( vStr );
    Vec_StrFree( vStr );
    return pResult;
}

/**Function*************************************************************

  Synopsis    [This procedure sets default parameters.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Gia_ParFfSetDefault(p *Bmc_ParFf_t) {
    memset( p, 0, sizeof(Bmc_ParFf_t) );
    p.pFileName     =  nil; 
    p.Algo          =     0; 
    p.fStartPats    =     0; 
    p.nTimeOut      =     0; 
    p.nIterCheck    =     0;
    p.fBasic        =     0; 
    p.fFfOnly       =     0;
    p.fCheckUntest  =     0; 
    p.fDump         =     0; 
    p.fDumpUntest   =     0; 
    p.fVerbose      =     0; 
}

/**Function*************************************************************

  Synopsis    [Adds a general cardinality constraint in terms of vVars.]

  Description [Strict is "exactly K out of N". Non-strict is 
  "less or equal than K out of N".]
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Cnf_AddCardinConstrTest() {
    var i int32;
    var status int32;
    var Count int32 = 1;
    var nVars int32 = 8;
    var vVars *Vec_Int_t  = Vec_IntStartNatural( nVars );
    var pSat *sat_solver  = sat_solver_new();
    sat_solver_setnvars( pSat, nVars );
    //Cnf_AddCardinConstr( pSat, vVars );
    Cnf_AddCardinConstrPairWise( pSat, vVars, 2, 1 );
    for {
        status = sat_solver_solve( pSat, nil, nil, 0, 0, 0, 0 );
        if status != l_True {
            break; }
        Vec_IntClear( vVars );
        fmt.Printf("%3d : ", Count++);
        for  i = 0; i < nVars; i++  {
            Vec_IntPush( vVars, Abc_Var2Lit(i, sat_solver_var_value(pSat, i)) );
            fmt.Printf("%d", sat_solver_var_value(pSat, i));
        }
        fmt.Printf("\n");
        status = sat_solver_addclause( pSat, Vec_IntArray(vVars), Vec_IntArray(vVars) + Vec_IntSize(vVars) );
        if status == 0 {
            break; }
    }
    sat_solver_delete( pSat );
    Vec_IntFree( vVars );
}

/**Function*************************************************************

  Synopsis    []

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
/**Function*************************************************************

  Synopsis    []

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
/**Function*************************************************************

  Synopsis    []

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Gia_ManFaultUnfold(p *Gia_Man_t, fUseMuxes int32, fFfOnly int32) *Gia_Man_t {
    var pNew, pTemp *Gia_Man_t;
    var pObj *Gia_Obj_t;
    var i, iCtrl, iThis int32;
    pNew = Gia_ManStart( (2 + 3 * fUseMuxes) * Gia_ManObjNum(p) );
    pNew.pName = Abc_UtilStrsav( p.pName );
    Gia_ManHashAlloc( pNew );
    Gia_ManConst0(p).Value = 0;
    // add first timeframe
    // inlining of macro Gia_ManForEachRo
    for i = 0; i < Gia_ManRegNum(p); i++ {
    	pObj = Gia_ManCi(p, Gia_ManPiNum(p)+i)
    	if pObj == nil { break }
        pObj.Value = Gia_ManAppendCi( pNew );
    }
    // inlining of macro Gia_ManForEachPi
    for i = 0; i < Gia_ManPiNum(p); i++ {
    	pObj = Gia_ManCi(p, i)
    	if pObj == nil { break }
        pObj.Value = Gia_ManAppendCi( pNew );
    }
    // inlining of macro Gia_ManForEachAnd
    for i = 0; i < p.nObjs; i++ {
    	pObj = Gia_ManObj(p, i)
    	if pObj == nil { break }
    	if !Gia_ObjIsAnd(pObj) {} else {
        pObj.Value = Gia_ManHashAnd( pNew, Gia_ObjFanin0Copy(pObj), Gia_ObjFanin1Copy(pObj) ); }
    }
    // inlining of macro Gia_ManForEachCo
    for i = 0; i < Vec_IntSize(p.vCos); i++ {
    	pObj = Gia_ManCo(p, i)
    	if pObj == nil { break }
        pObj.Value = Gia_ObjFanin0Copy(pObj);
    }
    // add second timeframe
    // inlining of macro Gia_ManForEachRo
    for i = 0; i < Gia_ManRegNum(p); i++ {
    	pObj = Gia_ManCi(p, Gia_ManPiNum(p)+i)
    	if pObj == nil { break }
        pObj.Value = Gia_ObjRoToRi(p, pObj).Value;
    }
    // inlining of macro Gia_ManForEachPi
    for i = 0; i < Gia_ManPiNum(p); i++ {
    	pObj = Gia_ManCi(p, i)
    	if pObj == nil { break }
        pObj.Value = Gia_ManAppendCi( pNew );
    }
    if fFfOnly != 0 {
        // inlining of macro Gia_ManForEachAnd
        for i = 0; i < p.nObjs; i++ {
        	pObj = Gia_ManObj(p, i)
        	if pObj == nil { break }
        	if !Gia_ObjIsAnd(pObj) {} else {
            pObj.Value = Gia_ManHashAnd( pNew, Gia_ObjFanin0Copy(pObj), Gia_ObjFanin1Copy(pObj) ); }
        }
        // inlining of macro Gia_ManForEachPo
        for i = 0; i < Gia_ManPoNum(p); i++ {
        	pObj = Gia_ManCo(p, i)
        	if pObj == nil { break }
            pObj.Value = Gia_ManAppendCo( pNew, Gia_ObjFanin0Copy(pObj) );
        }
        // inlining of macro Gia_ManForEachRi
        for i = 0; i < Gia_ManRegNum(p); i++ {
        	pObj = Gia_ManCo(p, Gia_ManPoNum(p)+i)
        	if pObj == nil { break }
            iCtrl = Gia_ManAppendCi(pNew);
            iThis = Gia_ObjFanin0Copy(pObj);
            if fUseMuxes != 0 {
                pObj.Value = Gia_ManHashMux( pNew, iCtrl, pObj.Value, iThis );
            } else {
                pObj.Value = iThis; }
            pObj.Value = Gia_ManAppendCo( pNew, pObj.Value );
        }
    
    } else {
        // inlining of macro Gia_ManForEachAnd
        for i = 0; i < p.nObjs; i++ {
        	pObj = Gia_ManObj(p, i)
        	if pObj == nil { break }
        	if !Gia_ObjIsAnd(pObj) {} else {
            iCtrl = Gia_ManAppendCi(pNew);
            iThis = Gia_ManHashAnd( pNew, Gia_ObjFanin0Copy(pObj), Gia_ObjFanin1Copy(pObj) );
            if fUseMuxes != 0 {
                pObj.Value = Gia_ManHashMux( pNew, iCtrl, pObj.Value, iThis );
            } else {
                pObj.Value = iThis; }
        }
        }
        // inlining of macro Gia_ManForEachCo
        for i = 0; i < Vec_IntSize(p.vCos); i++ {
        	pObj = Gia_ManCo(p, i)
        	if pObj == nil { break }
            pObj.Value = Gia_ManAppendCo( pNew, Gia_ObjFanin0Copy(pObj) );
        }
    }
    pNew = Gia_ManCleanup( pTemp = pNew );
    Gia_ManStop( pTemp );
    assert(Gia_ManPiNum(pNew) == Gia_ManRegNum(p) + 2 * Gia_ManPiNum(p) + (fFfOnly ? Gia_ManRegNum(p) : Gia_ManAndNum(p)), "Gia_ManPiNum(pNew) == Gia_ManRegNum(p) + 2 * Gia_ManPiNum(p) + (fFfOnly ? Gia_ManRegNum(p) : Gia_ManAndNum(p))");
    return pNew;
}

/**Function*************************************************************

  Synopsis    []

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Gia_ManStuckAtUnfold(p *Gia_Man_t, vMap *Vec_Int_t) *Gia_Man_t {
    var pNew, pTemp *Gia_Man_t;
    var pObj *Gia_Obj_t;
    var i, iFuncVars int32;
    pNew = Gia_ManStart( 3 * Gia_ManObjNum(p) );
    pNew.pName = Abc_UtilStrsav( p.pName );
    Gia_ManHashAlloc( pNew );
    Gia_ManConst0(p).Value = 0;
    // inlining of macro Gia_ManForEachCi
    for i = 0; i < Vec_IntSize(p.vCis); i++ {
    	pObj = Gia_ManCi(p, i)
    	if pObj == nil { break }
        pObj.Value = Gia_ManAppendCi( pNew );
    }
    // inlining of macro Gia_ManForEachAnd
    for i = 0; i < p.nObjs; i++ {
    	pObj = Gia_ManObj(p, i)
    	if pObj == nil { break }
    	if !Gia_ObjIsAnd(pObj) {} else {
        pObj.Value = Gia_ManHashAnd( pNew, Gia_ObjFanin0Copy(pObj), Gia_ObjFanin1Copy(pObj) );

        if Vec_IntEntry(vMap, iFuncVars++) {
            pObj.Value = Gia_ManHashAnd( pNew, Abc_LitNot(Gia_ManAppendCi(pNew)), pObj.Value );
        } else {
            Gia_ManAppendCi(pNew); }

        if Vec_IntEntry(vMap, iFuncVars++) {
            pObj.Value = Gia_ManHashOr( pNew, Gia_ManAppendCi(pNew), pObj.Value );
        } else {
            Gia_ManAppendCi(pNew); }
    }
    }
    assert(iFuncVars == Vec_IntSize(vMap), "iFuncVars == Vec_IntSize(vMap)");
    // inlining of macro Gia_ManForEachCo
    for i = 0; i < Vec_IntSize(p.vCos); i++ {
    	pObj = Gia_ManCo(p, i)
    	if pObj == nil { break }
        pObj.Value = Gia_ManAppendCo( pNew, Gia_ObjFanin0Copy(pObj) );
    }
    pNew = Gia_ManCleanup( pTemp = pNew );
    Gia_ManStop( pTemp );
    assert(Gia_ManPiNum(pNew) == Gia_ManCiNum(p) + 2 * Gia_ManAndNum(p), "Gia_ManPiNum(pNew) == Gia_ManCiNum(p) + 2 * Gia_ManAndNum(p)");
    return pNew;
}

/**Function*************************************************************

  Synopsis    []

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Gia_ManFlipUnfold(p *Gia_Man_t, vMap *Vec_Int_t) *Gia_Man_t {
    var pNew, pTemp *Gia_Man_t;
    var pObj *Gia_Obj_t;
    var i, iFuncVars int32;
    pNew = Gia_ManStart( 4 * Gia_ManObjNum(p) );
    pNew.pName = Abc_UtilStrsav( p.pName );
    Gia_ManHashAlloc( pNew );
    Gia_ManConst0(p).Value = 0;
    // inlining of macro Gia_ManForEachCi
    for i = 0; i < Vec_IntSize(p.vCis); i++ {
    	pObj = Gia_ManCi(p, i)
    	if pObj == nil { break }
        pObj.Value = Gia_ManAppendCi( pNew );
    }
    // inlining of macro Gia_ManForEachAnd
    for i = 0; i < p.nObjs; i++ {
    	pObj = Gia_ManObj(p, i)
    	if pObj == nil { break }
    	if !Gia_ObjIsAnd(pObj) {} else {
        pObj.Value = Gia_ManHashAnd( pNew, Gia_ObjFanin0Copy(pObj), Gia_ObjFanin1Copy(pObj) );
        if Vec_IntEntry(vMap, iFuncVars++) {
            pObj.Value = Gia_ManHashXor( pNew, Gia_ManAppendCi(pNew), pObj.Value );
        } else {
            Gia_ManAppendCi(pNew); }
    }
    }
    assert(iFuncVars == Vec_IntSize(vMap), "iFuncVars == Vec_IntSize(vMap)");
    // inlining of macro Gia_ManForEachCo
    for i = 0; i < Vec_IntSize(p.vCos); i++ {
    	pObj = Gia_ManCo(p, i)
    	if pObj == nil { break }
        pObj.Value = Gia_ManAppendCo( pNew, Gia_ObjFanin0Copy(pObj) );
    }
    pNew = Gia_ManCleanup( pTemp = pNew );
    Gia_ManStop( pTemp );
    assert(Gia_ManPiNum(pNew) == Gia_ManCiNum(p) + Gia_ManAndNum(p), "Gia_ManPiNum(pNew) == Gia_ManCiNum(p) + Gia_ManAndNum(p)");
    return pNew;
}

/**Function*************************************************************

  Synopsis    []

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Gia_ManFOFUnfold(p *Gia_Man_t, vMap *Vec_Int_t) *Gia_Man_t {
    var pNew, pTemp *Gia_Man_t;
    var pObj *Gia_Obj_t;
    var i, iCtrl0, iCtrl1, iCtrl2, iCtrl3, iMuxA, iMuxB, iFuncVars int32;
    var VarLimit int32  = 4 * Gia_ManAndNum(p);
    pNew = Gia_ManStart( 9 * Gia_ManObjNum(p) );
    pNew.pName = Abc_UtilStrsav( p.pName );
    Gia_ManHashAlloc( pNew );
    Gia_ManConst0(p).Value = 0;
    // inlining of macro Gia_ManForEachCi
    for i = 0; i < Vec_IntSize(p.vCis); i++ {
    	pObj = Gia_ManCi(p, i)
    	if pObj == nil { break }
        pObj.Value = Gia_ManAppendCi( pNew );
    }
    // inlining of macro Gia_ManForEachAnd
    for i = 0; i < p.nObjs; i++ {
    	pObj = Gia_ManObj(p, i)
    	if pObj == nil { break }
    	if !Gia_ObjIsAnd(pObj) {} else {
        if Vec_IntEntry(vMap, iFuncVars++) && iFuncVars < VarLimit {
            iCtrl0 = Gia_ManAppendCi(pNew);
        } else {
            iCtrl0 = 0;
Gia_ManAppendCi(pNew); }

        if Vec_IntEntry(vMap, iFuncVars++) && iFuncVars < VarLimit {
            iCtrl1 = Gia_ManAppendCi(pNew);
        } else {
            iCtrl1 = 0;
Gia_ManAppendCi(pNew); }

        if Vec_IntEntry(vMap, iFuncVars++) && iFuncVars < VarLimit {
            iCtrl2 = Gia_ManAppendCi(pNew);
        } else {
            iCtrl2 = 0;
Gia_ManAppendCi(pNew); }

        if Vec_IntEntry(vMap, iFuncVars++) && iFuncVars < VarLimit {
            iCtrl3 = Gia_ManAppendCi(pNew);
        } else {
            iCtrl3 = 0;
Gia_ManAppendCi(pNew); }

        if Gia_ObjFaninC0(pObj) && Gia_ObjFaninC1(pObj) {
            iCtrl0 = Abc_LitNot(iCtrl0);
        } else if !Gia_ObjFaninC0(pObj) && Gia_ObjFaninC1(pObj) {
            iCtrl1 = Abc_LitNot(iCtrl1);
        } else if Gia_ObjFaninC0(pObj) && !Gia_ObjFaninC1(pObj) {
            iCtrl2 = Abc_LitNot(iCtrl2);
        } else { //if ( !Gia_ObjFaninC0(pObj) && !Gia_ObjFaninC1(pObj) )
            iCtrl3 = Abc_LitNot(iCtrl3); }

        iMuxA       = Gia_ManHashMux( pNew, Gia_ObjFanin0(pObj).Value, iCtrl1, iCtrl0 );
        iMuxB       = Gia_ManHashMux( pNew, Gia_ObjFanin0(pObj).Value, iCtrl3, iCtrl2 );
        pObj.Value = Gia_ManHashMux( pNew, Gia_ObjFanin1(pObj).Value, iMuxB,  iMuxA );
    }
    }
    assert(iFuncVars == Vec_IntSize(vMap), "iFuncVars == Vec_IntSize(vMap)");
    // inlining of macro Gia_ManForEachCo
    for i = 0; i < Vec_IntSize(p.vCos); i++ {
    	pObj = Gia_ManCo(p, i)
    	if pObj == nil { break }
        pObj.Value = Gia_ManAppendCo( pNew, Gia_ObjFanin0Copy(pObj) );
    }
    pNew = Gia_ManCleanup( pTemp = pNew );
    Gia_ManStop( pTemp );
    assert(Gia_ManPiNum(pNew) == Gia_ManCiNum(p) + 4 * Gia_ManAndNum(p), "Gia_ManPiNum(pNew) == Gia_ManCiNum(p) + 4 * Gia_ManAndNum(p)");
    return pNew;
}

/**Function*************************************************************

  Synopsis    [This procedure sets default parameters.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Gia_FormStrCount(pStr string, pnVars *int32, pnPars *int32) int32 {
    var i, Counter int32;
    if pStr[0] != '(' {
        fmt.Printf("The first symbol should be the opening parenthesis \"(\".\n");
        return 1;
    }
    if pStr[len(pStr)-1] != ')' {
        fmt.Printf("The last symbol should be the closing parenthesis \")\".\n");
        return 1;
    }
    for  i = 0; pStr[i]; i++  {
        if pStr[i] == '(' {
            Counter++;
        } else if pStr[i] == ')' {
            Counter--; }
    }
    if Counter != 0 {
        fmt.Printf("The number of opening and closing parentheses is not equal.\n");
        return 1;
    }
    *pnVars = 0;
    *pnPars = 0;
    for  i = 0; pStr[i]; i++  {
        if pStr[i] >= 'a' && pStr[i] <= 'b' {
            *pnVars = Abc_MaxInt( *pnVars, pStr[i] - 'a' + 1 );
        } else if pStr[i] >= 'p' && pStr[i] <= 's' {
            *pnPars = Abc_MaxInt( *pnPars, pStr[i] - 'p' + 1 );
        } else if pStr[i] == '(' || pStr[i] == ')' {
        } else if pStr[i] == '&' || pStr[i] == '|' || pStr[i] == '^' {
        } else if pStr[i] == '?' || pStr[i] == ':' {
        } else if pStr[i] == '~' {
            if pStr[i+1] < 'a' || pStr[i+1] > 'z' {
                fmt.Printf("Expecting alphabetic symbol (instead of \"%c\") after negation (~)\n", pStr[i+1]);
                return 1;
            }
        
        } else {
            fmt.Printf("Unknown symbol (%c) in the formula (%s)\n", pStr[i], pStr);
            return 1;
        }
    }
    if *pnVars != FFTEST_MAX_VARS { fmt.Printf("The number of input variables (%d) should be 2\n", *pnVars); return 1; }
    if *pnPars < 1 || *pnPars > FFTEST_MAX_PARS { fmt.Printf("The number of parameters should be between 1 and %d\n", FFTEST_MAX_PARS); return 1; }
    return 0;
}
func Gia_FormStrTransform(pStr string, pForm string) {
    var i, k int32;
    for  k = i = 0; pForm[i]; i++  {
        if pForm[i] == '~' {
            i++;
            assert(pForm[i] >= 'a' && pForm[i] <= 'z', "pForm[i] >= 'a' && pForm[i] <= 'z'");
            pStr[k++] = 'A' + pForm[i] - 'a';
        
        } else {
            pStr[k++] = pForm[i]; }
    }
    pStr[k] = 0; 
}   

/**Function*************************************************************

  Synopsis    [Print formula.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Gia_ManFormulaEndToken(pForm string) string {
    var Counter int32;
    var pThis string;
    for  pThis = pForm; *pThis; pThis++  {
        assert(*pThis != '~', "*pThis != '~'");
        if *pThis == '(' {
            Counter++;
        } else if *pThis == ')' {
            Counter--; }
        if Counter == 0 {
            return pThis + 1; }
    }
    assert(false, "0");
    return nil;
}
func Gia_ManPrintFormula_rec(pBeg string, pEnd string) {
    var Oper int32  = -1;
    var pEndNew string;
    if pBeg + 1 == pEnd {
        if pBeg[0] >= 'a' && pBeg[0] <= 'b' {
            fmt.Printf("%c", pBeg[0]);
        } else if pBeg[0] >= 'A' && pBeg[0] <= 'B' {
            fmt.Printf("~%c", pBeg[0]-'A'+'a');
        } else if pBeg[0] >= 'p' && pBeg[0] <= 'w' { // pqrstuvw
            fmt.Printf("%c", pBeg[0]);
        } else if pBeg[0] >= 'P' && pBeg[0] <= 'W' {
            fmt.Printf("~%c", pBeg[0]-'A'+'a'); }
        return;
    }
    if pBeg[0] == '(' {
        pEndNew = Gia_ManFormulaEndToken( pBeg );
        if pEndNew == pEnd {
            assert(pBeg[0] == '(', "pBeg[0] == '('");
            assert(pBeg[pEnd-pBeg-1] == ')', "pBeg[pEnd-pBeg-1] == ')'");
            Gia_ManPrintFormula_rec( pBeg + 1, pEnd - 1 );
            return;
        }
    }
    // get first part
    pEndNew  = Gia_ManFormulaEndToken( pBeg );
    fmt.Printf("(");
    Gia_ManPrintFormula_rec( pBeg, pEndNew );
    fmt.Printf(")");
    Oper     = pEndNew[0];
    // derive the formula
    if Oper == '&' {
        fmt.Printf("&");
    } else if Oper == '|' {
        fmt.Printf("|");
    } else if Oper == '^' {
        fmt.Printf("^");
    } else if Oper == '?' {
        fmt.Printf("?");
    } else { assert(false, "0"); }
    // get second part
    pBeg     = pEndNew + 1;
    pEndNew  = Gia_ManFormulaEndToken( pBeg );
    fmt.Printf("(");
    Gia_ManPrintFormula_rec( pBeg, pEndNew );
    fmt.Printf(")");
    if Oper == '?' {
        fmt.Printf(":");
        // get third part
        assert(Oper == '?', "Oper == '?'");
        assert(pEndNew[0] == ':', "pEndNew[0] == ':'");
        pBeg     = pEndNew + 1;
        pEndNew  = Gia_ManFormulaEndToken( pBeg );
        fmt.Printf("(");
        Gia_ManPrintFormula_rec( pBeg, pEndNew );
        fmt.Printf(")");
    }
}
func Gia_ManPrintFormula(pStr string) {
    fmt.Printf("Using formula: ");
    fmt.Printf("(");
    Gia_ManPrintFormula_rec( pStr, pStr + len(pStr) );
    fmt.Printf(")");
    fmt.Printf("\n");
}

/**Function*************************************************************

  Synopsis    [Implements fault model formula using functional/parameter vars.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Gia_ManRealizeFormula_rec(p *Gia_Man_t, pVars *int32, pPars *int32, pBeg string, pEnd string, nPars int32) int32 {
    var iFans [3]int32;
    var Oper int32 = -1;
    var pEndNew string;
    if pBeg + 1 == pEnd {
        if pBeg[0] >= 'a' && pBeg[0] <= 'b' {
            return pVars[pBeg[0] - 'a']; }
        if pBeg[0] >= 'A' && pBeg[0] <= 'B' {
            return Abc_LitNot( pVars[pBeg[0] - 'A'] ); }
        if pBeg[0] >= 'p' && pBeg[0] <= 'w' { // pqrstuvw
            return pPars[pBeg[0] - 'p']; }
        if pBeg[0] >= 'P' && pBeg[0] <= 'W' {
            return Abc_LitNot( pPars[pBeg[0] - 'P'] ); }
        assert(false, "0");
        return -1;
    }
    if pBeg[0] == '(' {
        pEndNew = Gia_ManFormulaEndToken( pBeg );
        if pEndNew == pEnd {
            assert(pBeg[0] == '(', "pBeg[0] == '('");
            assert(pBeg[pEnd-pBeg-1] == ')', "pBeg[pEnd-pBeg-1] == ')'");
            return Gia_ManRealizeFormula_rec( p, pVars, pPars, pBeg + 1, pEnd - 1, nPars );
        }
    }
    // get first part
    pEndNew  = Gia_ManFormulaEndToken( pBeg );
    iFans[0] = Gia_ManRealizeFormula_rec( p, pVars, pPars, pBeg, pEndNew, nPars );
    Oper     = pEndNew[0];
    // get second part
    pBeg     = pEndNew + 1;
    pEndNew  = Gia_ManFormulaEndToken( pBeg );
    iFans[1] = Gia_ManRealizeFormula_rec( p, pVars, pPars, pBeg, pEndNew, nPars );
    // derive the formula
    if Oper == '&' {
        return Gia_ManHashAnd( p, iFans[0], iFans[1] ); }
    if Oper == '|' {
        return Gia_ManHashOr( p, iFans[0], iFans[1] ); }
    if Oper == '^' {
        return Gia_ManHashXor( p, iFans[0], iFans[1] ); }
    // get third part
    assert(Oper == '?', "Oper == '?'");
    assert(pEndNew[0] == ':', "pEndNew[0] == ':'");
    pBeg     = pEndNew + 1;
    pEndNew  = Gia_ManFormulaEndToken( pBeg );
    iFans[2] = Gia_ManRealizeFormula_rec( p, pVars, pPars, pBeg, pEndNew, nPars );
    return Gia_ManHashMux( p, iFans[0], iFans[1], iFans[2] );
}
func Gia_ManRealizeFormula(p *Gia_Man_t, pVars *int32, pPars *int32, pStr string, nPars int32) int32 {
    return Gia_ManRealizeFormula_rec( p, pVars, pPars, pStr, pStr + len(pStr), nPars );
}
func Gia_ManFormulaUnfold(p *Gia_Man_t, pForm string, fFfOnly int32) *Gia_Man_t {
    var pStr [100]byte;
    var Count int32;
    var pNew, pTemp *Gia_Man_t;
    var pObj *Gia_Obj_t;
    var i int32;
    var k int32;
    var iCtrl [FFTEST_MAX_PARS]int32;
    var iFans [FFTEST_MAX_VARS]int32;
    var nVars, nPars int32;
    assert(len(pForm) < 100, "strlen(pForm) < 100");
    Gia_FormStrCount( pForm, &nVars, &nPars );
    assert(nVars == 2, "nVars == 2");
    Gia_FormStrTransform( pStr, pForm );
    Gia_ManPrintFormula( pStr );
    pNew = Gia_ManStart( 5 * Gia_ManObjNum(p) );
    pNew.pName = Abc_UtilStrsav( p.pName );
    Gia_ManHashAlloc( pNew );
    Gia_ManConst0(p).Value = 0;
    // inlining of macro Gia_ManForEachCi
    for i = 0; i < Vec_IntSize(p.vCis); i++ {
    	pObj = Gia_ManCi(p, i)
    	if pObj == nil { break }
        pObj.Value = Gia_ManAppendCi( pNew );
    }
    if fFfOnly != 0 {
        Gia_ManCleanMark0( p );
        // inlining of macro Gia_ManForEachRi
        for i = 0; i < Gia_ManRegNum(p); i++ {
        	pObj = Gia_ManCo(p, Gia_ManPoNum(p)+i)
        	if pObj == nil { break }
            Gia_ObjFanin0(pObj).fMark0 = 1;
        }
        // inlining of macro Gia_ManForEachAnd
        for i = 0; i < p.nObjs; i++ {
        	pObj = Gia_ManObj(p, i)
        	if pObj == nil { break }
        	if !Gia_ObjIsAnd(pObj) {} else {
            if pObj.fMark0 {
                for  k = 0; k < nPars; k++  {
                    iCtrl[k] = Gia_ManAppendCi(pNew);
                }
                iFans[0] = Gia_ObjFanin0Copy(pObj);
                iFans[1] = Gia_ObjFanin1Copy(pObj);
                pObj.Value = Gia_ManRealizeFormula( pNew, iFans, iCtrl, pStr, nPars );
                Count++;
            
            } else {
                pObj.Value = Gia_ManHashAnd( pNew, Gia_ObjFanin0Copy(pObj), Gia_ObjFanin1Copy(pObj) ); }
        }
        }
        // inlining of macro Gia_ManForEachRi
        for i = 0; i < Gia_ManRegNum(p); i++ {
        	pObj = Gia_ManCo(p, Gia_ManPoNum(p)+i)
        	if pObj == nil { break }
            Gia_ObjFanin0(pObj).fMark0 = 0;
        }
    
    } else {
        // inlining of macro Gia_ManForEachAnd
        for i = 0; i < p.nObjs; i++ {
        	pObj = Gia_ManObj(p, i)
        	if pObj == nil { break }
        	if !Gia_ObjIsAnd(pObj) {} else {
            for  k = 0; k < nPars; k++  {
                iCtrl[k] = Gia_ManAppendCi(pNew);
            }
            iFans[0] = Gia_ObjFanin0Copy(pObj);
            iFans[1] = Gia_ObjFanin1Copy(pObj);
            pObj.Value = Gia_ManRealizeFormula( pNew, iFans, iCtrl, pStr, nPars );
        }
        }
    }
    // inlining of macro Gia_ManForEachCo
    for i = 0; i < Vec_IntSize(p.vCos); i++ {
    	pObj = Gia_ManCo(p, i)
    	if pObj == nil { break }
        pObj.Value = Gia_ManAppendCo( pNew, Gia_ObjFanin0Copy(pObj) );
    }
    pNew = Gia_ManCleanup( pTemp = pNew );
    Gia_ManStop( pTemp );
    assert(Gia_ManPiNum(pNew) == Gia_ManCiNum(p) + nPars * (fFfOnly ? Count : Gia_ManAndNum(p)), "Gia_ManPiNum(pNew) == Gia_ManCiNum(p) + nPars * (fFfOnly ? Count : Gia_ManAndNum(p))");
//    if ( fUseFaults )
//        Gia_AigerWrite( pNew, "newfault.aig", 0, 0, 0 );
    return pNew;
}

/**Function*************************************************************

  Synopsis    []

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Gia_ManFaultCofactor(p *Gia_Man_t, vValues *Vec_Int_t) *Gia_Man_t {
    var pNew, pTemp *Gia_Man_t;
    var pObj *Gia_Obj_t;
    var i int32;
    pNew = Gia_ManStart( Gia_ManObjNum(p) );
    pNew.pName = Abc_UtilStrsav( p.pName );
    Gia_ManHashAlloc( pNew );
    Gia_ManConst0(p).Value = 0;
    // inlining of macro Gia_ManForEachPi
    for i = 0; i < Gia_ManPiNum(p); i++ {
    	pObj = Gia_ManCi(p, i)
    	if pObj == nil { break }
        pObj.Value = Gia_ManAppendCi( pNew );
        if i < Vec_IntSize(vValues) {
            pObj.Value = Vec_IntEntry( vValues, i ); }
    }
    // inlining of macro Gia_ManForEachAnd
    for i = 0; i < p.nObjs; i++ {
    	pObj = Gia_ManObj(p, i)
    	if pObj == nil { break }
    	if !Gia_ObjIsAnd(pObj) {} else {
        pObj.Value = Gia_ManHashAnd( pNew, Gia_ObjFanin0Copy(pObj), Gia_ObjFanin1Copy(pObj) ); }
    }
    // inlining of macro Gia_ManForEachCo
    for i = 0; i < Vec_IntSize(p.vCos); i++ {
    	pObj = Gia_ManCo(p, i)
    	if pObj == nil { break }
        pObj.Value = Gia_ManAppendCo( pNew, Gia_ObjFanin0Copy(pObj) );
    }
    pNew = Gia_ManCleanup( pTemp = pNew );
    Gia_ManStop( pTemp );
    assert(Gia_ManPiNum(pNew) == Gia_ManPiNum(p), "Gia_ManPiNum(pNew) == Gia_ManPiNum(p)");
    return pNew;

}

/**Function*************************************************************

  Synopsis    []

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Gia_ManDumpTests(vTests *Vec_Int_t, nIter int32, pFileName string) {
    var pFile *os.File  = os.OpenFile(pFileName, os.O_WRONLY | os.O_CREATE | os.O_TRUNC, 0666);
    var i int32;
    var k int32;
    var v int32;
    var nVars int32 = Vec_IntSize(vTests) / nIter;
    assert(Vec_IntSize(vTests) % nIter == 0, "Vec_IntSize(vTests) % nIter == 0");
    for  v = i = 0; i < nIter; i++, fprintf(pFile, "\n")  {
        for  k = 0; k < nVars; k++  {
            fmt.Fprintf(pFile, "%d", Vec_IntEntry(vTests, v++));
        }
    }
    _ = (pFile).Close();
}

/**Function*************************************************************

  Synopsis    []

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Gia_ManDumpTestsSimulate(p *Gia_Man_t, vValues *Vec_Int_t) {
    var pObj *Gia_Obj_t; var k int32;
    assert(Vec_IntSize(vValues) == Gia_ManCiNum(p), "Vec_IntSize(vValues) == Gia_ManCiNum(p)");
    Gia_ManConst0(p).fMark0 = 0;
    // inlining of macro Gia_ManForEachCi
    for k = 0; k < Vec_IntSize(p.vCis); k++ {
    	pObj = Gia_ManCi(p, k)
    	if pObj == nil { break }
        pObj.fMark0 = Vec_IntEntry( vValues, k );
    }
    // inlining of macro Gia_ManForEachAnd
    for k = 0; k < p.nObjs; k++ {
    	pObj = Gia_ManObj(p, k)
    	if pObj == nil { break }
    	if !Gia_ObjIsAnd(pObj) {} else {
        pObj.fMark0 = (Gia_ObjFanin0(pObj).fMark0 ^ Gia_ObjFaninC0(pObj)) & 
                       (Gia_ObjFanin1(pObj).fMark0 ^ Gia_ObjFaninC1(pObj)); }
    }
    // inlining of macro Gia_ManForEachCo
    for k = 0; k < Vec_IntSize(p.vCos); k++ {
    	pObj = Gia_ManCo(p, k)
    	if pObj == nil { break }
        pObj.fMark0 = Gia_ObjFanin0(pObj).fMark0 ^ Gia_ObjFaninC0(pObj);
    }
    // collect flop input values
    Vec_IntClear( vValues );
    // inlining of macro Gia_ManForEachRi
    for k = 0; k < Gia_ManRegNum(p); k++ {
    	pObj = Gia_ManCo(p, Gia_ManPoNum(p)+k)
    	if pObj == nil { break }
        Vec_IntPush( vValues, pObj.fMark0 );
    }
    assert(Vec_IntSize(vValues) == Gia_ManRegNum(p), "Vec_IntSize(vValues) == Gia_ManRegNum(p)");
}
func Gia_ManDumpTestsDelay(vTests *Vec_Int_t, nIter int32, pFileName string, p *Gia_Man_t) {
    var pFile *os.File  = os.OpenFile(pFileName, os.O_WRONLY | os.O_CREATE | os.O_TRUNC, 0666);
    var vValues *Vec_Int_t  = Vec_IntAlloc( Gia_ManCiNum(p) );
    var i int32;
    var v int32;
    var nVars int32 = Vec_IntSize(vTests) / nIter;
    assert(Vec_IntSize(vTests) % nIter == 0, "Vec_IntSize(vTests) % nIter == 0");
    assert(nVars == 2 * Gia_ManPiNum(p) + Gia_ManRegNum(p), "nVars == 2 * Gia_ManPiNum(p) + Gia_ManRegNum(p)");
    for  i = 0; i < nIter; i++  {
        // collect PIs followed by flops
        Vec_IntClear( vValues );
        for  v = Gia_ManRegNum(p); v < Gia_ManCiNum(p); v++  {
            fmt.Fprintf(pFile, "%d", Vec_IntEntry(vTests, i * nVars + v));
            Vec_IntPush( vValues, Vec_IntEntry(vTests, i * nVars + v) );
        }
        for  v = 0; v < Gia_ManRegNum(p); v++  {
            fmt.Fprintf(pFile, "%d", Vec_IntEntry(vTests, i * nVars + v));
            Vec_IntPush( vValues, Vec_IntEntry(vTests, i * nVars + v) );
        }
        fmt.Fprintf(pFile, "\n");
        // derive next-state values
        Gia_ManDumpTestsSimulate( p, vValues );
        // collect PIs followed by flops
        for  v = Gia_ManCiNum(p); v < nVars; v++  {
            fmt.Fprintf(pFile, "%d", Vec_IntEntry(vTests, i * nVars + v));
        }
        for  v = 0; v < Vec_IntSize(vValues); v++  {
            fmt.Fprintf(pFile, "%d", Vec_IntEntry(vValues, v));
        }
        fmt.Fprintf(pFile, "\n");
    }
    Gia_ManCleanMark0(p);
    _ = (pFile).Close();
    Vec_IntFree( vValues );
}

/**Function*************************************************************

  Synopsis    []

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Gia_ManPrintResults(p *Gia_Man_t, pSat *sat_solver, nIter int32, clk abctime) {
    var pTable *os.File  = os.OpenFile("fault_stats.txt", os.O_RDWR | os.O_CREATE | os.O_APPEND, 0666);

    fmt.Fprintf(pTable, "%s ", Gia_ManName(p));
    fmt.Fprintf(pTable, "%d ", Gia_ManPiNum(p));
    fmt.Fprintf(pTable, "%d ", Gia_ManPoNum(p));
    fmt.Fprintf(pTable, "%d ", Gia_ManRegNum(p));
    fmt.Fprintf(pTable, "%d ", Gia_ManAndNum(p));

    fmt.Fprintf(pTable, "%d ", sat_solver_nvars(pSat));
    fmt.Fprintf(pTable, "%d ", sat_solver_nclauses(pSat));
    fmt.Fprintf(pTable, "%d ", sat_solver_nconflicts(pSat));

    fmt.Fprintf(pTable, "%d ", nIter);
    fmt.Fprintf(pTable, "%.2f", 1.0*clk/int64(time.Second));
    fmt.Fprintf(pTable, "\n");
    _ = (pTable).Close();
}

/**Function*************************************************************

  Synopsis    []

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Gia_ManFaultAddOne(pM *Gia_Man_t, pCnf *Cnf_Dat_t, pSat *sat_solver, vLits *Vec_Int_t, nFuncVars int32, fAddOr int32, pGiaCnf *Gia_Man_t) int32 {
    var pC *Gia_Man_t;//, * pTemp;
    var pCnf2 *Cnf_Dat_t;
    var pObj *Gia_Obj_t;
    var i, Lit int32;
    // derive the cofactor
    pC = Gia_ManFaultCofactor( pM, vLits );
    // derive new CNF
    pCnf2 = Cnf_DeriveGiaRemapped( pC );
    Cnf_DataLiftGia( pCnf2, pC, sat_solver_nvars(pSat) );
    // add timeframe clauses
    for  i = 0; i < pCnf2.nClauses; i++  {
        if !sat_solver_addclause( pSat, pCnf2.pClauses[i], pCnf2.pClauses[i+1] ) {
            Cnf_DataFree( pCnf2 );
            Gia_ManStop( pC );
            return 0;
        }
    }
    // add constraint clauses
    if fAddOr != 0 {
        // add one OR gate to assert difference of at least one output of the miter
        var vOrGate *Vec_Int_t  = Vec_IntAlloc( Gia_ManPoNum(pC) );
        // inlining of macro Gia_ManForEachPo
        for i = 0; i < Gia_ManPoNum(pC); i++ {
        	pObj = Gia_ManCo(pC, i)
        	if pObj == nil { break }
            Lit = Abc_Var2Lit( pCnf2.pVarNums[Gia_ObjId(pC, pObj)], 0 );
            Vec_IntPush( vOrGate, Lit );
        }
        if !sat_solver_addclause( pSat, Vec_IntArray(vOrGate), Vec_IntArray(vOrGate) + Vec_IntSize(vOrGate) ) {
            assert(false, "0"); }
        Vec_IntFree( vOrGate );
    
    } else {
        // add negative literals to assert equality of all outputs of the miter
        // inlining of macro Gia_ManForEachPo
        for i = 0; i < Gia_ManPoNum(pC); i++ {
        	pObj = Gia_ManCo(pC, i)
        	if pObj == nil { break }
            Lit = Abc_Var2Lit( pCnf2.pVarNums[Gia_ObjId(pC, pObj)], 1 );
            if !sat_solver_addclause( pSat, &Lit, &Lit+1 ) {
                Cnf_DataFree( pCnf2 );
                Gia_ManStop( pC );
                return 0;
            }
        }
    }
    // add connection clauses
    if pGiaCnf != nil {
        // inlining of macro Gia_ManForEachPi
        for i = 0; i < Gia_ManPiNum(pGiaCnf); i++ {
        	pObj = Gia_ManCi(pGiaCnf, i)
        	if pObj == nil { break }
            if i >= nFuncVars {
                sat_solver_add_buffer( pSat, pCnf.pVarNums[Gia_ObjId(pGiaCnf, pObj)], pCnf2.pVarNums[Gia_ObjId(pC, Gia_ManPi(pC, i))], 0 ); }
        }
    }
    Cnf_DataFree( pCnf2 );
    Gia_ManStop( pC );
    return 1;
}


/**Function*************************************************************

  Synopsis    []

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Gia_ManDumpUntests(pM *Gia_Man_t, pCnf *Cnf_Dat_t, pSat *sat_solver, nFuncVars int32, pFileName string, fVerbose int32) int32 {
    var pFile *os.File  = os.OpenFile(pFileName, os.O_WRONLY | os.O_CREATE | os.O_TRUNC, 0666);
    var vLits *Vec_Int_t;
    var pObj *Gia_Obj_t;
    var nItersMax int32  = 10000;
    var i, nIters, status, Value, Count int32;
    vLits = Vec_IntAlloc( Gia_ManPiNum(pM) - nFuncVars );
    for  nIters = 0; nIters < nItersMax; nIters++  {
        status = sat_solver_solve( pSat, nil, nil, (ABC_INT64_T)0, (ABC_INT64_T)0, (ABC_INT64_T)0, (ABC_INT64_T)0 );
        if status == l_Undef {
            fmt.Printf("Timeout reached after dumping %d untestable faults.\n", nIters); }
        if status == l_Undef {
            break; }
        if status == l_False {
            break; }
        // collect literals
        Vec_IntClear( vLits );
        // inlining of macro Gia_ManForEachPi
        for i = 0; i < Gia_ManPiNum(pM); i++ {
        	pObj = Gia_ManCi(pM, i)
        	if pObj == nil { break }
            if i >= nFuncVars {
                Vec_IntPush( vLits, Abc_Var2Lit(pCnf.pVarNums[Gia_ObjId(pM, pObj)], sat_solver_var_value(pSat, pCnf.pVarNums[Gia_ObjId(pM, pObj)])) ); }
        }
        // dump the fault
        // inlining of macro Vec_IntForEachEntry
        for  i = 0; (i < Vec_IntSize(vLits)) && (((Value) = Vec_IntEntry(vLits, i)), 1); i++  {
            if Abc_LitIsCompl(Value) {
                break; }
        }
        if i < Vec_IntSize(vLits) {
            if fVerbose != 0 {
                fmt.Printf("Untestable fault %4d : ", ++Count);
                // inlining of macro Vec_IntForEachEntry
                for  i = 0; (i < Vec_IntSize(vLits)) && (((Value) = Vec_IntEntry(vLits, i)), 1); i++  {
                    if Abc_LitIsCompl(Value) {
                        fmt.Printf("%d ", i); }
                }
                fmt.Printf("\n");
            }
            // inlining of macro Vec_IntForEachEntry
            for  i = 0; (i < Vec_IntSize(vLits)) && (((Value) = Vec_IntEntry(vLits, i)), 1); i++  {
                if Abc_LitIsCompl(Value) {
                    fmt.Fprintf(pFile, "%d ", i); }
            }
            fmt.Fprintf(pFile, "\n");
        }
        // add this clause
        if !sat_solver_addclause( pSat, Vec_IntArray(vLits), Vec_IntArray(vLits) + Vec_IntSize(vLits) ) {
            break; }
    }
    Vec_IntFree( vLits );
    _ = (pFile).Close();
    return nIters;
}

/**Function*************************************************************

  Synopsis    []

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Gia_ManGetTestPatterns(pFileName string) *Vec_Int_t {
    var pFile *os.File  = os.Open(pFileName);
    var vTests *Vec_Int_t; var c int32;
    if pFile == nil {
        fmt.Printf("Cannot open input file \"%s\".\n", pFileName);
        return nil;
    }
    vTests = Vec_IntAlloc( 10000 );
    for  (c = fgetc(pFile)) != -1  {
        if c == ' ' || c == '\t' || c == '\r' || c == '\n' {
            continue; }
        if c != '0' && c != '1' {
            fmt.Printf("Wrong symbol (%c) in the input file.\n", c);
            Vec_IntFreeP( &vTests );
            break;
        }
        Vec_IntPush( vTests, c - '0' );
    }
    _ = (pFile).Close();
    return vTests;
}

/**Function*************************************************************

  Synopsis    [Derive the second AIG.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Gia_ManDeriveDup(p *Gia_Man_t, nPisNew int32) *Gia_Man_t {
    var i int32;
    var pNew *Gia_Man_t  = Gia_ManDup(p);
    for  i = 0; i < nPisNew; i++  {
        Gia_ManAppendCi( pNew );
    }
    return pNew;
}

/**Function*************************************************************

  Synopsis    []

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Gia_ManFaultAnalyze(pSat *sat_solver, vPars *Vec_Int_t, vMap *Vec_Int_t, vLits *Vec_Int_t, Iter int32) int32 {
    var nConfLimit int32  = 100;
    var status, i, v, iVar, Lit int32;
    var nUnsats, nRuns int32;
    var clk abctime  = Abc_Clock();
    assert(Vec_IntSize(vPars) == Vec_IntSize(vMap), "Vec_IntSize(vPars) == Vec_IntSize(vMap)");
    // check presence of each variable
    Vec_IntClear( vLits );
    Vec_IntAppend( vLits, vMap );
    for  v = 0; v < Vec_IntSize(vPars); v++  {
        if !Vec_IntEntry(vLits, v) {
            continue; }
        assert(Vec_IntEntry(vLits, v) == 1, "Vec_IntEntry(vLits, v) == 1");
        nRuns++;
        Lit = Abc_Var2Lit( Vec_IntEntry(vPars, v), 0 );
        status = sat_solver_solve( pSat, &Lit, &Lit+1, (ABC_INT64_T)nConfLimit, (ABC_INT64_T)0, (ABC_INT64_T)0, (ABC_INT64_T)0 );
        if status == l_Undef {
            continue; }
        if status == l_False {
            nUnsats++;
            assert(Vec_IntEntry(vMap, v) == 1, "Vec_IntEntry(vMap, v) == 1");
            Vec_IntWriteEntry( vMap, v, 0 );
            Lit = Abc_LitNot(Lit);
            //status = sat_solver_addclause( pSat, &Lit, &Lit+1 );
            //assert( status );
            continue;
        }
        // inlining of macro Vec_IntForEachEntry
        for  i = 0; (i < Vec_IntSize(vPars)) && (((iVar) = Vec_IntEntry(vPars, i)), 1); i++  {
            if Vec_IntEntry(vLits, i) && sat_solver_var_value(pSat, iVar) {
                Vec_IntWriteEntry( vLits, i, 0 ); }
        }
        assert(Vec_IntEntry(vLits, v) == 0, "Vec_IntEntry(vLits, v) == 0");
    }
    fmt.Printf("Iteration %3d has determined %5d (out of %5d) parameters after %6d SAT calls.  ", Iter, Vec_IntSize(vMap) - Vec_IntCountPositive(vMap), Vec_IntSize(vPars), nRuns);
    Abc_PrintTime( 1, "Time", Abc_Clock() - clk );
    return nUnsats;
}

/**Function*************************************************************

  Synopsis    [Dump faults detected by the new test, which are not detected by previous tests.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Gia_ManFaultDumpNewFaults(pM *Gia_Man_t, nFuncVars int32, vTests *Vec_Int_t, vTestNew *Vec_Int_t, pPars *Bmc_ParFf_t) int32 {
    var pFileName string  = "newfaults.txt";
    var clk abctime;
    var pC *Gia_Man_t;
    var pCnf2 *Cnf_Dat_t;
    var pSat *sat_solver;
    var vLits *Vec_Int_t;
    var i, Iter, IterMax, nNewFaults int32;

    // derive the cofactor
    pC = Gia_ManFaultCofactor( pM, vTestNew );
    // derive new CNF
    pCnf2 = Cnf_DeriveGiaRemapped( pC );

    // create SAT solver
    pSat = sat_solver_new();
    sat_solver_setnvars( pSat, 1 );
    sat_solver_set_runtime_limit( pSat, pPars.nTimeOut ? pPars.nTimeOut * int64(time.Second) + Abc_Clock(): 0 );
    // create the first cofactor
    Gia_ManFaultAddOne( pM, nil, pSat, vTestNew, nFuncVars, 1, nil );

    // add other test patterns
    assert(Vec_IntSize(vTests) % nFuncVars == 0, "Vec_IntSize(vTests) % nFuncVars == 0");
    IterMax = Vec_IntSize(vTests) / nFuncVars;
    vLits = Vec_IntAlloc( nFuncVars );
    for  Iter = 0; Iter < IterMax; Iter++  {
        // get pattern
        Vec_IntClear( vLits );
        for  i = 0; i < nFuncVars; i++  {
            Vec_IntPush( vLits, Vec_IntEntry(vTests, Iter*nFuncVars + i) );
        }
        // the resulting problem cannot be UNSAT, because the new test is guaranteed 
        // to detect something that the given test set does not detect
        if !(Gia_ManFaultAddOne( pM, pCnf2, pSat, vLits, nFuncVars, 0, pC ) != 0) {
            assert(false, "0"); }
    }
    Vec_IntFree( vLits );

    // dump the new faults
    clk = Abc_Clock();
    nNewFaults = Gia_ManDumpUntests( pC, pCnf2, pSat, nFuncVars, pFileName, pPars.fVerbose );
    fmt.Printf("Dumped %d new multiple faults into file \"%s\".  ", nNewFaults, pFileName);
    Abc_PrintTime( 1, "Time", Abc_Clock() - clk );

    // cleanup
    sat_solver_delete( pSat );
    Cnf_DataFree( pCnf2 );
    Gia_ManStop( pC );
    return 1;
}

/**Function*************************************************************

  Synopsis    [Generate miter, CNF and solver.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Gia_ManFaultPrepare(p *Gia_Man_t, pG *Gia_Man_t, pPars *Bmc_ParFf_t, nFuncVars int32, vMap *Vec_Int_t, vTests *Vec_Int_t, vLits *Vec_Int_t, ppMiter **Gia_Man_t, ppCnf **Cnf_Dat_t, ppSat **sat_solver, fWarmUp int32) int32 {
    var p0 *Gia_Man_t = nil;
    var p1 *Gia_Man_t = nil;
    var pM *Gia_Man_t;
    var pObj *Gia_Obj_t;
    var pCnf *Cnf_Dat_t;
    var pSat *sat_solver;
    var i, Iter, status int32;
    var clkSat abctime;

    if Vec_IntSize(vTests) && (Vec_IntSize(vTests) % nFuncVars != 0) {
        fmt.Printf("The number of symbols in the input patterns (%d) does not divide evenly on the number of test variables (%d).\n", Vec_IntSize(vTests), nFuncVars);
        Vec_IntFree( vTests );
        return 0;
    }

    // select algorithm
    if pPars.Algo == 0 {
        p1 = Gia_ManFormulaUnfold( p, pPars.pFormStr, pPars.fFfOnly );
    } else if pPars.Algo == 1 {
        assert(Gia_ManRegNum(p) > 0, "Gia_ManRegNum(p) > 0");
        p0 = Gia_ManFaultUnfold( pG, 0, pPars.fFfOnly );
        p1 = Gia_ManFaultUnfold( p, 1, pPars.fFfOnly );
    
    } else if pPars.Algo == 2 {
        p1 = Gia_ManStuckAtUnfold( p, vMap );
    } else if pPars.Algo == 3 {
        p1 = Gia_ManFlipUnfold( p, vMap );
    } else if pPars.Algo == 4 {
        p1 = Gia_ManFOFUnfold( p, vMap ); }
    if pPars.Algo != 1 {
        p0 = Gia_ManDeriveDup( pG, Gia_ManCiNum(p1) - Gia_ManCiNum(pG) ); }
//    Gia_AigerWrite( p1, "newfault.aig", 0, 0, 0 );
//    printf( "Dumped circuit with fault parameters into file \"newfault.aig\".\n" );

    // create miter
    pM = Gia_ManMiter( p0, p1, 0, 0, 0, 0, 0 );
    pCnf = Cnf_DeriveGiaRemapped( pM );
    Gia_ManStop( p0 );
    Gia_ManStop( p1 );

    // start the SAT solver
    pSat = sat_solver_new();
    sat_solver_setnvars( pSat, pCnf.nVars );
    sat_solver_set_runtime_limit( pSat, pPars.nTimeOut ? pPars.nTimeOut * int64(time.Second) + Abc_Clock(): 0 );
    // add timeframe clauses
    for  i = 0; i < pCnf.nClauses; i++  {
        if !sat_solver_addclause( pSat, pCnf.pClauses[i], pCnf.pClauses[i+1] ) {
            assert(false, "0"); }
    }

    // add one large OR clause
    Vec_IntClear( vLits );
    // inlining of macro Gia_ManForEachCo
    for i = 0; i < Vec_IntSize(pM.vCos); i++ {
    	pObj = Gia_ManCo(pM, i)
    	if pObj == nil { break }
        Vec_IntPush( vLits, Abc_Var2Lit(pCnf.pVarNums[Gia_ObjId(pM, pObj)], 0) );
    }
    sat_solver_addclause( pSat, Vec_IntArray(vLits), Vec_IntArray(vLits) + Vec_IntSize(vLits) );

    // save return data
    *ppMiter = pM;
    *ppCnf = pCnf;
    *ppSat = pSat;

    // add cardinality constraint
    if pPars.fBasic {
        Vec_IntClear( vLits );
        // inlining of macro Gia_ManForEachPi
        for i = 0; i < Gia_ManPiNum(pM); i++ {
        	pObj = Gia_ManCi(pM, i)
        	if pObj == nil { break }
            if i >= nFuncVars {
                Vec_IntPush( vLits, pCnf.pVarNums[Gia_ObjId(pM, pObj)] ); }
        }
        Cnf_AddCardinConstr( pSat, vLits );
    
    } else if pPars.nCardConstr {
        Vec_IntClear( vLits );
        // inlining of macro Gia_ManForEachPi
        for i = 0; i < Gia_ManPiNum(pM); i++ {
        	pObj = Gia_ManCi(pM, i)
        	if pObj == nil { break }
            if i >= nFuncVars {
                Vec_IntPush( vLits, pCnf.pVarNums[Gia_ObjId(pM, pObj)] ); }
        }
        Cnf_AddCardinConstrGeneral( pSat, vLits, pPars.nCardConstr, !pPars.fNonStrict );
    }

    // add available test-patterns
    if Vec_IntSize(vTests) > 0 {
        var nTests int32  = Vec_IntSize(vTests) / nFuncVars;
        assert(Vec_IntSize(vTests) % nFuncVars == 0, "Vec_IntSize(vTests) % nFuncVars == 0");
        if pPars.pFileName {
            fmt.Printf("Reading %d pre-computed test patterns from file \"%s\".\n", Vec_IntSize(vTests) / nFuncVars, pPars.pFileName);
        } else {
            fmt.Printf("Reading %d pre-computed test patterns from previous rounds.\n", Vec_IntSize(vTests) / nFuncVars); }
        for  Iter = 0; Iter < nTests; Iter++  {
            if fWarmUp != 0 {
                var clk abctime  = Abc_Clock();
                status = sat_solver_solve( pSat, nil, nil, (ABC_INT64_T)0, (ABC_INT64_T)0, (ABC_INT64_T)0, (ABC_INT64_T)0 );
                clkSat += Abc_Clock() - clk;
                if status == l_Undef {
                    if pPars.fVerbose {
                        fmt.Printf("\n"); }
                    fmt.Printf("Timeout reached after %d seconds and adding %d tests.\n", pPars.nTimeOut, Iter);
                    Vec_IntShrink( vTests, Iter * nFuncVars );
                    return 0;
                }
                if status == l_False {
                    if pPars.fVerbose {
                        fmt.Printf("\n"); }
                    fmt.Printf("The problem is UNSAT after adding %d tests.\n", Iter);
                    Vec_IntShrink( vTests, Iter * nFuncVars );
                    return 0;
                }
            }
            // get pattern
            Vec_IntClear( vLits );
            for  i = 0; i < nFuncVars; i++  {
                Vec_IntPush( vLits, Vec_IntEntry(vTests, Iter*nFuncVars + i) );
            }
            if !(Gia_ManFaultAddOne( pM, pCnf, pSat, vLits, nFuncVars, 0, pM ) != 0) {
                if pPars.fVerbose {
                    fmt.Printf("\n"); }
                fmt.Printf("The problem is UNSAT after adding %d tests.\n", Iter);
                Vec_IntShrink( vTests, Iter * nFuncVars );
                return 0;
            }
            if pPars.fVerbose {
                fmt.Printf("Iter%6d : ", Iter);
                fmt.Printf("Var =%10d  ", sat_solver_nvars(pSat));
                fmt.Printf("Clause =%10d  ", sat_solver_nclauses(pSat));
                fmt.Printf("Conflict =%10d  ", sat_solver_nconflicts(pSat));
                //Abc_PrintTime( 1, "Time", clkSat );
                // inlining of macro ABC_PRTr
                Abc_Print(1, "%s =", ("Solver time"));
Abc_Print(1, "%9.2f sec\r", 1.0*(float64((clkSat)))/(float64(int64(time.Second))));;
            }
        }
    
    } else if pPars.fStartPats {
        for  Iter = 0; Iter < 2; Iter++  {
            status = sat_solver_solve( pSat, nil, nil, (ABC_INT64_T)0, (ABC_INT64_T)0, (ABC_INT64_T)0, (ABC_INT64_T)0 );
            if status == l_Undef {
                if pPars.fVerbose {
                    fmt.Printf("\n"); }
                fmt.Printf("Timeout reached after %d seconds and %d iterations.\n", pPars.nTimeOut, Iter);
                Vec_IntShrink( vTests, Iter * nFuncVars );
                return 0;
            }
            if status == l_False {
                if pPars.fVerbose {
                    fmt.Printf("\n"); }
                fmt.Printf("The problem is UNSAT after %d iterations.\n", Iter);
                Vec_IntShrink( vTests, Iter * nFuncVars );
                return 0;
            }
            // initialize simple pattern
            Vec_IntFill( vLits, nFuncVars, Iter );
            Vec_IntAppend( vTests, vLits );
            if !(Gia_ManFaultAddOne( pM, pCnf, pSat, vLits, nFuncVars, 0, pM ) != 0) {
                if pPars.fVerbose {
                    fmt.Printf("\n"); }
                fmt.Printf("The problem is UNSAT after adding %d tests.\n", Iter);
                Vec_IntShrink( vTests, Iter * nFuncVars );
                return 0;
            }
        }
    }

    fmt.Printf("Using miter with:  AIG nodes = %6d.  CNF variables = %6d.  CNF clauses = %8d.\n", Gia_ManAndNum(pM), pCnf.nVars, pCnf.nClauses);
    return 1;
}

/**Function*************************************************************

  Synopsis    []

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Gia_ManFaultTest(p *Gia_Man_t, pG *Gia_Man_t, pPars *Bmc_ParFf_t) {
    var nIterMax int32 = 1000000;
    var nVars int32;
    var nPars int32;
    var i int32;
    var Iter int32;
    var Iter2 int32;
    var status int32;
    var nFuncVars int32 = -1;
    var clk abctime;
    var clkSat abctime;
    var clkTotal abctime = Abc_Clock();
    var vLits *Vec_Int_t;
    var vMap *Vec_Int_t = nil;
    var vTests *Vec_Int_t;
    var vPars *Vec_Int_t = nil;
    var pM *Gia_Man_t;
    var pObj *Gia_Obj_t;
    var pCnf *Cnf_Dat_t;
    var pSat *sat_solver  = nil;

    if pPars.Algo == 0 && (Gia_FormStrCount( pPars.pFormStr, &nVars, &nPars ) != 0) {
        return; }

    // select algorithm
    if pPars.Algo == 0 {
        fmt.Printf("FFTEST is computing test patterns for fault model \"%s\"...\n", pPars.pFormStr);
    } else if pPars.Algo == 1 {
        fmt.Printf("FFTEST is computing test patterns for %sdelay faults...\n", pPars.fBasic ? "single " : "");
    } else if pPars.Algo == 2 {
        fmt.Printf("FFTEST is computing test patterns for %sstuck-at faults...\n", pPars.fBasic ? "single " : "");
    } else if pPars.Algo == 3 {
        fmt.Printf("FFTEST is computing test patterns for %scomplement faults...\n", pPars.fBasic ? "single " : "");
    } else if pPars.Algo == 4 {
        fmt.Printf("FFTEST is computing test patterns for %sfunctionally observable faults...\n", pPars.fBasic ? "single " : "");
    } else {
        fmt.Printf("Unrecognized algorithm (%d).\n", pPars.Algo);
        return;
    }

    fmt.Printf("Options: ");
    fmt.Printf("Untestable faults = %s. ", pPars.fCheckUntest || pPars.fDumpDelay ? "yes": "no");
    if pPars.nCardConstr {
        fmt.Printf("Using %sstrict cardinality %d. ", pPars.fNonStrict ? "non-" : "", pPars.nCardConstr); }
    if pPars.fFfOnly {
        fmt.Printf("Faults at FF outputs only = yes. "); }
    if pPars.nTimeOut {
        fmt.Printf("Runtime limit = %d sec.  ", pPars.nTimeOut); }
    if p != pG && pG.pSpec {
        fmt.Printf("Golden model = %s. ", pG.pSpec); }
    fmt.Printf("Verbose = %s. ", pPars.fVerbose ? "yes": "no");
    fmt.Printf("\n");

    // select algorithm
    if pPars.Algo == 0 {
        nFuncVars = Gia_ManCiNum(p);
    } else if pPars.Algo == 1 {
        nFuncVars = Gia_ManRegNum(p) + 2 * Gia_ManPiNum(p);
    } else if pPars.Algo == 2 {
        nFuncVars = Gia_ManCiNum(p);
    } else if pPars.Algo == 3 {
        nFuncVars = Gia_ManCiNum(p);
    } else if pPars.Algo == 4 {
        nFuncVars = Gia_ManCiNum(p); }

    // collect test patterns from file
    if pPars.pFileName {
        vTests = Gia_ManGetTestPatterns( pPars.pFileName );
    } else {
        vTests = Vec_IntAlloc( 10000 ); }
    if vTests == nil {
        return; }

    // select algorithm
    vMap = Vec_IntAlloc( 0 );
    if pPars.Algo == 2 {
        Vec_IntFill( vMap, 2 * Gia_ManAndNum(p), 1 );
    } else if pPars.Algo == 3 {
        Vec_IntFill( vMap,     Gia_ManAndNum(p), 1 );
    } else if pPars.Algo == 4 {
        Vec_IntFill( vMap, 4 * Gia_ManAndNum(p), 1 ); }

    // prepare SAT solver
    vLits = Vec_IntAlloc( Gia_ManCoNum(p) );

    // add user-speicified test-vectors (if given) and create missing ones (if needed)
    if Gia_ManFaultPrepare(p, pG, pPars, nFuncVars, vMap, vTests, vLits, &pM, &pCnf, &pSat, 1) != 0 {
    for  Iter = pPars.fStartPats ? 2 : Vec_IntSize(vTests) / nFuncVars; Iter < nIterMax; Iter++  {
        // collect parameter variables
        if pPars.nIterCheck && vPars == nil {
            vPars = Vec_IntAlloc( Gia_ManPiNum(pM) - nFuncVars );
            // inlining of macro Gia_ManForEachPi
            for i = 0; i < Gia_ManPiNum(pM); i++ {
            	pObj = Gia_ManCi(pM, i)
            	if pObj == nil { break }
                if i >= nFuncVars {
                    Vec_IntPush( vPars, pCnf.pVarNums[Gia_ObjId(pM, pObj)] ); }
            }
            assert(Vec_IntSize(vPars) == Gia_ManPiNum(pM) - nFuncVars, "Vec_IntSize(vPars) == Gia_ManPiNum(pM) - nFuncVars");
        }
        // derive unit parameter variables
        if Iter && pPars.nIterCheck && (Iter % pPars.nIterCheck) == 0 {
            Gia_ManFaultAnalyze( pSat, vPars, vMap, vLits, Iter );
            // cleanup
            Gia_ManStop( pM );
            Cnf_DataFree( pCnf );
            sat_solver_delete( pSat );
            // recompute
            if !(Gia_ManFaultPrepare(p, pG, pPars, nFuncVars, vMap, vTests, vLits, &pM, &pCnf, &pSat, 0) != 0) {
                fmt.Printf("This should never happen.\n");
                return;
            }
            Vec_IntFreeP( &vPars );
        }
        // solve
        clk = Abc_Clock();
        status = sat_solver_solve( pSat, nil, nil, (ABC_INT64_T)0, (ABC_INT64_T)0, (ABC_INT64_T)0, (ABC_INT64_T)0 );
        clkSat += Abc_Clock() - clk;
        if pPars.fVerbose {
            fmt.Printf("Iter%6d : ", Iter);
            fmt.Printf("Var =%10d  ", sat_solver_nvars(pSat));
            fmt.Printf("Clause =%10d  ", sat_solver_nclauses(pSat));
            fmt.Printf("Conflict =%10d  ", sat_solver_nconflicts(pSat));
            //Abc_PrintTime( 1, "Time", clkSat );
            // inlining of macro ABC_PRTr
            Abc_Print(1, "%s =", ("Solver time"));
Abc_Print(1, "%9.2f sec\r", 1.0*(float64((clkSat)))/(float64(int64(time.Second))));;
        }
        if status == l_Undef {
            if pPars.fVerbose {
                fmt.Printf("\n"); }
            fmt.Printf("Timeout reached after %d seconds and %d iterations.\n", pPars.nTimeOut, Iter);
            goto finish;
        }
        if status == l_False {
            if pPars.fVerbose {
                fmt.Printf("\n"); }
            fmt.Printf("The problem is UNSAT after %d iterations.  ", Iter);
            break;
        }
        assert(status == l_True, "status == l_True");
        // collect SAT assignment
        Vec_IntClear( vLits );
        // inlining of macro Gia_ManForEachPi
        for i = 0; i < Gia_ManPiNum(pM); i++ {
        	pObj = Gia_ManCi(pM, i)
        	if pObj == nil { break }
            if i < nFuncVars {
                Vec_IntPush( vLits, sat_solver_var_value(pSat, pCnf.pVarNums[Gia_ObjId(pM, pObj)]) ); }
        }
        // if the user selected to generate new faults detected by this test (vLits)
        // and not detected by the given test set (vTests), we compute the new faults here and quit
        if pPars.fDumpNewFaults {
            if Vec_IntSize(vTests) == 0 {
                fmt.Printf("The input test patterns are not given.\n");
            } else {
                Gia_ManFaultDumpNewFaults( pM, nFuncVars, vTests, vLits, pPars ); }
            goto finish_all;
        }
        Vec_IntAppend( vTests, vLits );
        // add constraint
        if !(Gia_ManFaultAddOne( pM, pCnf, pSat, vLits, nFuncVars, 0, pM ) != 0) {
            if pPars.fVerbose {
                fmt.Printf("\n"); }
            fmt.Printf("The problem is UNSAT after adding %d tests.\n", Iter);
            break;
        }
    }
    } else { Iter = Vec_IntSize(vTests) / nFuncVars; }
finish:
    // print results
//    if ( status == l_False )
//        Gia_ManPrintResults( p, pSat, Iter, Abc_Clock() - clkTotal );
    // cleanup
    Abc_PrintTime( 1, "Testing runtime", Abc_Clock() - clkTotal );
    // dump the test suite
    if pPars.fDump {
        var pFileName string;
        if p.pSpec {
            pFileName = Extra_FileNameGenericAppend(p.pSpec, "_tests.txt");
        } else {
            pFileName = (char *)"tests.txt";
        }
        if pPars.fDumpDelay && pPars.Algo == 1 {
            Gia_ManDumpTestsDelay( vTests, Iter, pFileName, p );
            fmt.Printf("Dumping %d pairs of test patterns (total %d pattern) into file \"%s\".\n", Vec_IntSize(vTests) / nFuncVars, 2*Vec_IntSize(vTests) / nFuncVars, pFileName);
        
        } else {
            Gia_ManDumpTests( vTests, Iter, pFileName );
            fmt.Printf("Dumping %d test patterns into file \"%s\".\n", Vec_IntSize(vTests) / nFuncVars, pFileName);
        }
    }

    // compute untestable faults
    if Iter && (p != pG || pPars.fDumpUntest || pPars.fCheckUntest) {
        var clkTotal abctime  = Abc_Clock();
        // restart the SAT solver
        sat_solver_delete( pSat );
        pSat = sat_solver_new();
        sat_solver_setnvars( pSat, pCnf.nVars );
        sat_solver_set_runtime_limit( pSat, pPars.nTimeOut ? pPars.nTimeOut * int64(time.Second) + Abc_Clock(): 0 );
        // add timeframe clauses
        for  i = 0; i < pCnf.nClauses; i++  {
            if !sat_solver_addclause( pSat, pCnf.pClauses[i], pCnf.pClauses[i+1] ) {
                assert(false, "0"); }
        }
        // add constraint to rule out no fault
//        if ( p == pG )
        {
            Vec_IntClear( vLits );
            // inlining of macro Gia_ManForEachPi
            for i = 0; i < Gia_ManPiNum(pM); i++ {
            	pObj = Gia_ManCi(pM, i)
            	if pObj == nil { break }
                if i >= nFuncVars {
                    Vec_IntPush( vLits, Abc_Var2Lit(pCnf.pVarNums[Gia_ObjId(pM, pObj)], 0) ); }
            }
            sat_solver_addclause( pSat, Vec_IntArray(vLits), Vec_IntArray(vLits) + Vec_IntSize(vLits) );
        }
        // add cardinality constraint
        if pPars.fBasic {
            Vec_IntClear( vLits );
            // inlining of macro Gia_ManForEachPi
            for i = 0; i < Gia_ManPiNum(pM); i++ {
            	pObj = Gia_ManCi(pM, i)
            	if pObj == nil { break }
                if i >= nFuncVars {
                    Vec_IntPush( vLits, pCnf.pVarNums[Gia_ObjId(pM, pObj)] ); }
            }
            Cnf_AddCardinConstr( pSat, vLits );
        }
        // add output clauses
        // inlining of macro Gia_ManForEachCo
        for i = 0; i < Vec_IntSize(pM.vCos); i++ {
        	pObj = Gia_ManCo(pM, i)
        	if pObj == nil { break }
            var Lit int32  = Abc_Var2Lit( pCnf.pVarNums[Gia_ObjId(pM, pObj)], 1 );
            sat_solver_addclause( pSat, &Lit, &Lit + 1 );
        }
        // simplify the SAT solver
        status = sat_solver_simplify( pSat );
        assert((status != 0), "status");

        // add test patterns
        assert(Vec_IntSize(vTests) == Iter * nFuncVars, "Vec_IntSize(vTests) == Iter * nFuncVars");
        for  Iter2 = 0; ; Iter2++  {
            var clk abctime  = Abc_Clock();
            status = sat_solver_solve( pSat, nil, nil, (ABC_INT64_T)0, (ABC_INT64_T)0, (ABC_INT64_T)0, (ABC_INT64_T)0 );
            clkSat += Abc_Clock() - clk;
            if pPars.fVerbose {
                fmt.Printf("Iter%6d : ", Iter2);
                fmt.Printf("Var =%10d  ", sat_solver_nvars(pSat));
                fmt.Printf("Clause =%10d  ", sat_solver_nclauses(pSat));
                fmt.Printf("Conflict =%10d  ", sat_solver_nconflicts(pSat));
                //Abc_PrintTime( 1, "Time", clkSat );
                // inlining of macro ABC_PRTr
                Abc_Print(1, "%s =", ("Solver time"));
Abc_Print(1, "%9.2f sec\r", 1.0*(float64((clkSat)))/(float64(int64(time.Second))));;
            }
            if status == l_Undef {
                if pPars.fVerbose {
                    fmt.Printf("\n"); }
                fmt.Printf("Timeout reached after %d seconds and %d iterations.\n", pPars.nTimeOut, Iter2);
                goto finish;
            }
            if Iter2 == Iter {
                break; }
            assert(status == l_True, "status == l_True");
            // get pattern
            Vec_IntClear( vLits );
            for  i = 0; i < nFuncVars; i++  {
                Vec_IntPush( vLits, Vec_IntEntry(vTests, Iter2*nFuncVars + i) );
            }
            if !(Gia_ManFaultAddOne( pM, pCnf, pSat, vLits, nFuncVars, 0, pM ) != 0) {
                fmt.Printf("The problem is UNSAT after adding %d tests.\n", Iter2);
                goto finish;
            }
        }
        assert(Iter2 == Iter, "Iter2 == Iter");
        if pPars.fVerbose {
            fmt.Printf("\n"); }
        if p == pG {
            if status == l_True {
                fmt.Printf("There are untestable faults.  ");
            } else if status == l_False {
                fmt.Printf("There is no untestable faults.  ");
            } else { assert(false, "0"); }
            Abc_PrintTime( 1, "Fault computation runtime", Abc_Clock() - clkTotal );
        
        } else {
            if status == l_True {
                fmt.Printf("The circuit is rectifiable.  ");
            } else if status == l_False {
                fmt.Printf("The circuit is not rectifiable (or equivalent to the golden one).  ");
            } else { assert(false, "0"); }
            Abc_PrintTime( 1, "Rectification runtime", Abc_Clock() - clkTotal );
        }
        // dump untestable faults
        if pPars.fDumpUntest && status == l_True {
            var clk abctime  = Abc_Clock();
            var pFileName string;
            if p.pSpec {
                pFileName = Extra_FileNameGenericAppend(p.pSpec, "_untest.txt");
            } else {
                pFileName = (char *)"untest.txt";
            }
            var nUntests int32  = Gia_ManDumpUntests( pM, pCnf, pSat, nFuncVars, pFileName, pPars.fVerbose );
            if p == pG {
                fmt.Printf("Dumped %d untestable multiple faults into file \"%s\".  ", nUntests, pFileName);
            } else {
                fmt.Printf("Dumped %d ways of rectifying the circuit into file \"%s\".  ", nUntests, pFileName); }
            Abc_PrintTime( 1, "Time", Abc_Clock() - clk );
        }
    }
finish_all:
    sat_solver_delete( pSat );
    Cnf_DataFree( pCnf );
    Gia_ManStop( pM );
    Vec_IntFree( vTests );
    Vec_IntFree( vMap );
    Vec_IntFree( vLits );
    Vec_IntFreeP( &vPars );
}



////////////////////////////////////////////////////////////////////////
///                       END OF FILE                                ///
////////////////////////////////////////////////////////////////////////


//ABC_NAMESPACE_IMPL_END

