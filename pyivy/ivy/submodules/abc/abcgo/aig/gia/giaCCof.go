package gia

import (
	"fmt"
	"time"
)

/**CFile****************************************************************

  FileName    [giaCCof.c]

  SystemName  [ABC: Logic synthesis and verification system.]

  PackageName [Scalable AIG package.]

  Synopsis    [Backward reachability using circuit cofactoring.]

  Author      [Alan Mishchenko]
  
  Affiliation [UC Berkeley]

  Date        [Ver. 1.0. Started - June 20, 2005.]

  Revision    [$Id: giaCCof.c,v 1.00 2005/06/20 00:00:00 alanmi Exp $]

***********************************************************************/

//#include "gia.h"
//#include "sat/bsat/satSolver.h"

//ABC_NAMESPACE_IMPL_START


////////////////////////////////////////////////////////////////////////
///                        DECLARATIONS                              ///
////////////////////////////////////////////////////////////////////////

    // manager
type Ccf_Man_t struct {
    // user data
    pGia   *Gia_Man_t;       // single-output AIG manager
    nFrameMax          int32;  // maximum number of frames
    nConfMax          int32;   // maximum number of conflicts
    nTimeMax          int32;   // maximum runtime in seconds
    fVerbose          int32;   // verbose flag
    // internal data
    pUnr        any;       // unrolling manager
    pFrames   *Gia_Man_t;    // unrolled timeframes
    vCopies   *Vec_Int_t;    // copy pointers of the AIG
    pSat  *sat_solver;       // SAT solver
}

////////////////////////////////////////////////////////////////////////
///                     FUNCTION DEFINITIONS                         ///
////////////////////////////////////////////////////////////////////////

/**Function*************************************************************

  Synopsis    [Create manager.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
var spStatic_Ccf_ManStart_Pars Gia_ParFra_t;
var spStatic_Ccf_ManStart_pPars *Gia_ParFra_t = &spStatic_Ccf_ManStart_Pars;
func Ccf_ManStart(pGia *Gia_Man_t, nFrameMax int32, nConfMax int32, nTimeMax int32, fVerbose int32) *Ccf_Man_t {
    
    
    assert(nFrameMax > 0, "nFrameMax > 0");
    // inlining of macro ABC_CALLOC
    p := &Ccf_Man_t{};
    p.pGia      = pGia;
    p.nFrameMax = nFrameMax;    
    p.nConfMax  = nConfMax;
    p.nTimeMax  = nTimeMax;
    p.fVerbose  = fVerbose;
    // create unrolling manager
    memset( spStatic_Ccf_ManStart_pPars, 0, sizeof(Gia_ParFra_t) );
    spStatic_Ccf_ManStart_pPars.fVerbose     = fVerbose;
    spStatic_Ccf_ManStart_pPars.nFrames      = nFrameMax;
    spStatic_Ccf_ManStart_pPars.fSaveLastLit = 1;
    p.pUnr = Gia_ManUnrollStart( pGia, spStatic_Ccf_ManStart_pPars );
    p.vCopies = Vec_IntAlloc( 1000 );
    // internal data
    p.pSat = sat_solver_new();
//    sat_solver_setnvars( p->pSat, 10000 );
    return p;
}

/**Function*************************************************************

  Synopsis    [Delete manager.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Ccf_ManStop(p *Ccf_Man_t) {
    Vec_IntFree( p.vCopies );
    Gia_ManUnrollStop( p.pUnr );
    sat_solver_delete( p.pSat );
    Gia_ManStopP( &p.pFrames );
    // inlining of macro ABC_FREE
    p = nil;
}


/**Function*************************************************************

  Synopsis    [Extends the solver.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Gia_ManCofExtendSolver(p *Ccf_Man_t) {
    var pObj *Gia_Obj_t;
    var i int32;
    // add SAT clauses
    for  i = sat_solver_nvars(p.pSat); i < Gia_ManObjNum(p.pFrames); i++  {
        pObj = Gia_ManObj( p.pFrames, i );
        if Gia_ObjIsAnd(pObj) {
            sat_solver_add_and( p.pSat, i, 
                Gia_ObjFaninId0(pObj, i), 
                Gia_ObjFaninId1(pObj, i), 
                Gia_ObjFaninC0(pObj), 
                Gia_ObjFaninC1(pObj), 0 ); } 
    }
    sat_solver_setnvars( p.pSat, Gia_ManObjNum(p.pFrames) );
}

func Gia_Obj0Copy(vCopies *Vec_Int_t, Fan0 int32, fCompl0 int32) int32 { return Abc_LitNotCond( Vec_IntEntry(vCopies, Fan0), fCompl0 ); }

func Gia_Obj1Copy(vCopies *Vec_Int_t, Fan1 int32, fCompl1 int32) int32 { return Abc_LitNotCond( Vec_IntEntry(vCopies, Fan1), fCompl1 ); }
 
/**Function*************************************************************

  Synopsis    [Cofactor the circuit w.r.t. the given assignment.]

  Description [Assumes that the solver has just returned SAT.]
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Gia_ManCofOneDerive_rec(p *Ccf_Man_t, Id int32) {
    var pObj *Gia_Obj_t;
    var Res int32;
    if Vec_IntEntry(p.vCopies, Id) != -1 {
        return; }
    pObj = Gia_ManObj(p.pFrames, Id);
    assert(Gia_ObjIsCi(pObj) || Gia_ObjIsAnd(pObj), "Gia_ObjIsCi(pObj) || Gia_ObjIsAnd(pObj)");
    if Gia_ObjIsAnd(pObj) {
        var fCompl0 int32  = Gia_ObjFaninC0(pObj);
        var fCompl1 int32  = Gia_ObjFaninC1(pObj);
        var Fan0 int32  = Gia_ObjFaninId0p(p.pFrames, pObj);
        var Fan1 int32  = Gia_ObjFaninId1p(p.pFrames, pObj);
        Gia_ManCofOneDerive_rec( p, Fan0 );
        Gia_ManCofOneDerive_rec( p, Fan1 );
        Res = Gia_ManHashAnd( p.pFrames, 
            Gia_Obj0Copy(p.vCopies, Fan0, fCompl0), 
            Gia_Obj1Copy(p.vCopies, Fan1, fCompl1) );
    
    } else if Gia_ObjCioId(pObj) >= Gia_ManRegNum(p.pGia) { // PI
        Res = sat_solver_var_value( p.pSat, Id );
    } else {
        Res = Abc_Var2Lit( Id, 0 ); }
    Vec_IntWriteEntry( p.vCopies, Id, Res );
}

/**Function*************************************************************

  Synopsis    [Cofactor the circuit w.r.t. the given assignment.]

  Description [Assumes that the solver has just returned SAT.]
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Gia_ManCofOneDerive(p *Ccf_Man_t, LitProp int32) int32 {
    var LitOut int32;
    // derive the cofactor of the property node
    Vec_IntFill( p.vCopies, Gia_ManObjNum(p.pFrames), -1 );
    Gia_ManCofOneDerive_rec( p, Abc_Lit2Var(LitProp) );
    LitOut = Vec_IntEntry( p.vCopies, Abc_Lit2Var(LitProp) );
    LitOut = Abc_LitNotCond( LitOut, Abc_LitIsCompl(LitProp) );
    // add new PO for the cofactor
    Gia_ManAppendCo( p.pFrames, LitOut );
    // add SAT clauses
    Gia_ManCofExtendSolver( p );
    // return negative literal of the cofactor
    return Abc_LitNot(LitOut);
} 

/**Function*************************************************************

  Synopsis    [Enumerates backward reachable states.]

  Description [Return -1 if resource limit is reached. Returns 1 
  if computation converged (there is no more reachable states).
  Returns 0 if no more states to enumerate.]
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Gia_ManCofGetReachable(p *Ccf_Man_t, Lit int32) int32 {
    var ObjPrev, ConfPrev int32;
    var Count, LitOut, RetValue int32;
    var clk abctime;
    // try solving for the first time and quit if converged
    RetValue = sat_solver_solve( p.pSat, &Lit, &Lit + 1, p.nConfMax, 0, 0, 0 );
    if RetValue == l_False {
        return 1; }
    // iterate circuit cofactoring
    for  RetValue == l_True  {
        clk = Abc_Clock();
        // derive cofactor
        LitOut = Gia_ManCofOneDerive( p, Lit );
        // add the blocking clause
        RetValue = sat_solver_addclause( p.pSat, &LitOut, &LitOut + 1 );
        assert((RetValue != 0), "RetValue");
        // try solving again
        RetValue = sat_solver_solve( p.pSat, &Lit, &Lit + 1, p.nConfMax, 0, 0, 0 );
        // derive cofactors
        if p.fVerbose {
            fmt.Printf("%3d : AIG =%7d  Conf =%7d.  ", Count++, Gia_ManObjNum(p.pFrames) - ObjPrev, sat_solver_nconflicts(p.pSat) - ConfPrev);
            Abc_PrintTime( 1, "Time", Abc_Clock() - clk );
            ObjPrev = Gia_ManObjNum(p.pFrames);
            ConfPrev = sat_solver_nconflicts(p.pSat);
        }
    }
    if RetValue == l_Undef {
        return -1; }
    return 0;
}

/**Function*************************************************************

  Synopsis    []

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Gia_ManCofTest(pGia *Gia_Man_t, nFrameMax int32, nConfMax int32, nTimeMax int32, fVerbose int32) *Gia_Man_t { 
    var pNew *Gia_Man_t;
    var p *Ccf_Man_t;
    var pObj *Gia_Obj_t;
    var f int32;
    var i int32;
    var Lit int32;
    var RetValue int32 = -1;
    var fFailed int32;
    var nTimeToStop abctime  = Abc_Clock() + nTimeMax * int64(time.Second);
    var clk abctime  = Abc_Clock();
    assert(Gia_ManPoNum(pGia) == 1, "Gia_ManPoNum(pGia) == 1");

    // create reachability manager
    p = Ccf_ManStart( pGia, nFrameMax, nConfMax, nTimeMax, fVerbose );

    // set runtime limit
    if nTimeMax {
        sat_solver_set_runtime_limit( p.pSat, nTimeToStop ); }

    // perform backward image computation
    for  f = 0; f < nFrameMax; f++  {
        if fVerbose != 0 { 
            fmt.Printf("ITER %3d :\n", f); }
        // add to the mapping of nodes
        p.pFrames = (Gia_Man_t *)Gia_ManUnrollAdd( p.pUnr, f+1 );
        // add SAT clauses
        Gia_ManCofExtendSolver( p );
        // return output literal
        Lit = Gia_ManUnrollLastLit( p.pUnr );
        // derives cofactors of the property literal till all states are blocked
        RetValue = Gia_ManCofGetReachable( p, Lit );
        if RetValue != 0 {
            break; }

        // check the property output
        Gia_ManSetPhase( p.pFrames );
        // inlining of macro Gia_ManForEachPo
        for  i = 0; (i < Gia_ManPoNum(p.pFrames)) && ((pObj) = Gia_ManCo(p.pFrames, i)); i++  {
            if pObj.fPhase {
                fmt.Printf("Property failed in frame %d.\n", f);
                fFailed = 1;
                break;
            }
        }
        if i < Gia_ManPoNum(p.pFrames) {
            break; }
    }

    // report the result
    if nTimeToStop && Abc_Clock() > nTimeToStop {
        fmt.Printf("Runtime limit (%d sec) is reached after %d frames.  ", nTimeMax, f);
    } else if f == nFrameMax {
        fmt.Printf("Completed %d frames without converging.  ", f);
    } else if RetValue == 1 {
        fmt.Printf("Backward reachability converged after %d iterations.  ", f-1);
    } else if RetValue == -1 {
        fmt.Printf("Conflict limit or timeout is reached after %d frames.  ", f-1); }
    Abc_PrintTime( 1, "Runtime", Abc_Clock() - clk );

    if !(fFailed != 0) && RetValue == 1 {
        fmt.Printf("Property holds.\n");
    } else if !(fFailed != 0) {
        fmt.Printf("Property is undecided.\n"); }

    // get the resulting AIG manager
    Gia_ManHashStop( p.pFrames );
    pNew = p.pFrames;  p.pFrames = nil;
    Ccf_ManStop( p );

    // cleanup
//    if ( fVerbose )
//        Gia_ManPrintStats( pNew, 0 );
    pNew = Gia_ManCleanup( pGia = pNew );
    Gia_ManStop( pGia );
//    if ( fVerbose )
//        Gia_ManPrintStats( pNew, 0 );
    return pNew;   
}

////////////////////////////////////////////////////////////////////////
///                       END OF FILE                                ///
////////////////////////////////////////////////////////////////////////


//ABC_NAMESPACE_IMPL_END

