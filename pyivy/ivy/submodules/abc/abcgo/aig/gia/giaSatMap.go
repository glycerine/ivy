package gia

import "fmt"

/**CFile****************************************************************

  FileName    [giaSatMap.c]

  SystemName  [ABC: Logic synthesis and verification system.]

  PackageName [Scalable AIG package.]

  Synopsis    []

  Author      [Alan Mishchenko]
  
  Affiliation [UC Berkeley]

  Date        [Ver. 1.0. Started - June 20, 2005.]

  Revision    [$Id: giaSatMap.c,v 1.00 2005/06/20 00:00:00 alanmi Exp $]

***********************************************************************/

//#include "gia.h"
//#include "sat/bsat/satStore.h"
//#include "misc/vec/vecWec.h"
//#include "misc/util/utilNam.h"
//#include "map/scl/sclCon.h"

//ABC_NAMESPACE_IMPL_START


////////////////////////////////////////////////////////////////////////
///                        DECLARATIONS                              ///
////////////////////////////////////////////////////////////////////////


// operation manager
   
type Sbm_Man_t struct {
    pSat      *sat_solver;      // SAT solver
    vCardVars       *Vec_Int_t; // candinality variables
    LogN              int32;      // max vars
    FirstVar              int32;  // first variable to be used
    LitShift              int32;  // shift in terms of literals (2*Gia_ManCiNum(pGia)+2)
    nInputs              int32;   // the number of inputs
    // window
    vRoots       *Vec_Int_t;    // output drivers to be mapped (root -> lit)
    vCuts       *Vec_Wec_t;     // cuts (cut -> node lit + fanin lits)
    vObjCuts       *Vec_Wec_t;  // cuts (obj -> node lit + cut lits)
    vSolCuts       *Vec_Int_t;  // current solution (index -> cut)
    vCutGates       *Vec_Int_t; // gates (cut -> gate)
    vCutAreas       *Vec_Wrd_t; // area of each cut
    // solver
    vAssump       *Vec_Int_t;   // assumptions (root nodes)
    vPolar       *Vec_Int_t;    // polarity of nodes and cuts
    // timing
    vArrs       *Vec_Int_t;     // arrivals for the inputs (input -> time)
    vReqs       *Vec_Int_t;     // required for the outputs (root -> time)
    // internal
    vLit2Used       *Vec_Int_t; // current solution (lit -> used)
    vDelays       *Vec_Int_t;   // node arrivals (lit -> time)
    vReason       *Vec_Int_t;   // timing reasons (lit -> cut)
}

/*
    Cuts in p->vCuts have 0-based numbers and are expressed in terms of object literals.
    Cuts in p->vObjCuts are expressed in terms of the obj-lit + cut-lits (i + p->FirstVar)
    Unit cuts for each polarity are ordered in the end.
*/

////////////////////////////////////////////////////////////////////////
///                     FUNCTION DEFINITIONS                         ///
////////////////////////////////////////////////////////////////////////

/**Function*************************************************************

  Synopsis    [Verify solution. Create polarity and assumptions.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Sbm_ManCheckSol(p *Sbm_Man_t, vSol *Vec_Int_t) int32 {
    //int K = Vec_IntSize(vSol) - 1;
    var i, j, Lit, Cut int32;
    var RetValue int32  = 1;
    var vCut *Vec_Int_t;
    // clear polarity and assumptions
    Vec_IntClear( p.vPolar );
    // mark used literals
    Vec_IntFill( p.vLit2Used, Vec_WecSize(p.vObjCuts) + p.nInputs, 0 );
    // inlining of macro Vec_IntForEachEntry
    for  i = 0; (i < Vec_IntSize(p.vSolCuts)) && (((Cut) = Vec_IntEntry(p.vSolCuts, i)), 1); i++  {
        if Cut < 0 { // input inverter variable
        
            Vec_IntWriteEntry( p.vLit2Used, -Cut, 1 );
            Vec_IntPush( p.vPolar, -Cut ); 
            continue;
        }
        Vec_IntPush( p.vPolar, p.FirstVar + Cut ); 
        vCut = Vec_WecEntry( p.vCuts, Cut );
        Lit = Vec_IntEntry( vCut, 0 ) - p.LitShift; // obj literal
        if Vec_IntEntry(p.vLit2Used, Lit) {
            continue; }
        Vec_IntWriteEntry( p.vLit2Used, Lit, 1 );
        Vec_IntPush( p.vPolar, Lit ); // literal variable
    }
    // check that the output literals are used
    // inlining of macro Vec_IntForEachEntry
    for  i = 0; (i < Vec_IntSize(p.vRoots)) && (((Lit) = Vec_IntEntry(p.vRoots, i)), 1); i++  {
        if Vec_IntEntry(p.vLit2Used, Lit) == 0 {
            fmt.Printf("Output literal %d has no cut.\n", Lit);
RetValue = 0; }
    }
    // check internal nodes
    // inlining of macro Vec_IntForEachEntry
    for  i = 0; (i < Vec_IntSize(p.vSolCuts)) && (((Cut) = Vec_IntEntry(p.vSolCuts, i)), 1); i++  {
        if Cut < 0 {
            continue; }
        vCut = Vec_WecEntry( p.vCuts, Cut );
        // inlining of macro Vec_IntForEachEntryStart
        for  j = 1; (j < Vec_IntSize(vCut)) && (((Lit) = Vec_IntEntry(vCut, j)), 1); j++  {
            if Lit - p.LitShift < 0 {
                assert(Abc_LitIsCompl(Lit), "Abc_LitIsCompl(Lit)");
                if Vec_IntEntry(p.vLit2Used, Vec_WecSize(p.vObjCuts) + Abc_Lit2Var(Lit)-1) == 0 {
                    fmt.Printf("Inverter of input %d of cut %d is not mapped.\n", Abc_Lit2Var(Lit)-1, Cut);
RetValue = 0; }
            
            } else if Vec_IntEntry(p.vLit2Used, Lit - p.LitShift) == 0 {
                fmt.Printf("Internal literal %d of cut %d is not mapped.\n", Lit - p.LitShift, Cut);
RetValue = 0; }
        }
        // create polarity
        Vec_IntPush( p.vPolar, p.FirstVar + Cut ); // cut variable
    }
    return RetValue;
}

/**Function*************************************************************

  Synopsis    []

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Sbm_ManCreateCnf(p *Sbm_Man_t) int32 {
    var i int32;
    var k int32;
    var Lit int32;
    var Lits [2]int32;
    var value int32;
    var vLits, vCutOne, vLitsPrev *Vec_Int_t;
    // sat_solver_rollback( p->Sat );
    if !(Sbm_ManCheckSol(p, p.vSolCuts) != 0) {
        return 0; }
    // increase var count
    assert(p.FirstVar == sat_solver_nvars(p.pSat), "p.FirstVar == sat_solver_nvars(p.pSat)");
    sat_solver_setnvars( p.pSat, sat_solver_nvars(p.pSat) + Vec_WecSize(p.vCuts) );
    // iterate over arrays containing obj-lit cuts (neg-obj-lit cut-lits followed by pos-obj-lit cut-lits)
    vLitsPrev = nil;
    // inlining of macro Vec_WecForEachLevel
    for  i = 0; (i < Vec_WecSize(p.vObjCuts)) && (((vLits) = Vec_WecEntry(p.vObjCuts, i)), 1); i++  {
        assert(Vec_IntSize(vLits) >= 2, "Vec_IntSize(vLits) >= 2");
        value = sat_solver_addclause( p.pSat, Vec_IntArray(vLits), Vec_IntLimit(vLits)  );
        assert((value != 0), "value");
/*
        // for each cut, add implied nodes
        Lits[0] = Abc_LitNot( Vec_IntEntry(vLits, 0) );
        assert( Lits[0] < 2*p->FirstVar );
        Vec_IntForEachEntryStart( vLits, Lit, k, 1 )
        {
            assert( Lit >= 2*p->FirstVar );
            Lits[1] = Abc_LitNot( Lit );
            value = sat_solver_addclause( p->pSat, Lits, Lits + 2 );
            assert( value );
            //printf( "Adding clause %d + %d.\n", Lits[0], Lits[1]-2*p->FirstVar );
        }
*/
        //printf( "\n" );
        // create invertor exclusivity clause
        if i & 1 {
            Lits[0] = Abc_LitNot( Vec_IntEntryLast(vLits) );
            Lits[1] = Abc_LitNot( Vec_IntEntryLast(vLitsPrev) );
            value = sat_solver_addclause( p.pSat, Lits, Lits + 2 );
            assert((value != 0), "value");
            //printf( "Adding exclusivity clause %d + %d.\n", Lits[0]-2*p->FirstVar, Lits[1]-2*p->FirstVar );
        }
        vLitsPrev = vLits;
    }
    // add inverters
    // inlining of macro Vec_WecForEachLevel
    for  i = 0; (i < Vec_WecSize(p.vCuts)) && (((vCutOne) = Vec_WecEntry(p.vCuts, i)), 1); i++  {
        // inlining of macro Vec_IntForEachEntry
        for  k = 0; (k < Vec_IntSize(vCutOne)) && (((Lit) = Vec_IntEntry(vCutOne, k)), 1); k++  {
            if Abc_Lit2Var(Lit)-1 < p.nInputs { // input
            
                assert(k > 0, "k > 0");
                Lits[0] = Abc_Var2Lit( Vec_WecSize(p.vObjCuts) + Abc_Lit2Var(Lit)-1, 0 );
                Lits[1] = Abc_Var2Lit( p.FirstVar + i, 1 );
                value = sat_solver_addclause( p.pSat, Lits, Lits + 2 );
                assert((value != 0), "value");
            
            } else { // internal node
            
                Lits[0] = Abc_Var2Lit( Lit-p.LitShift, 0 );
                Lits[1] = Abc_Var2Lit( p.FirstVar + i, 1 );
                value = sat_solver_addclause( p.pSat, Lits, Lits + 2 );
                assert((value != 0), "value");
            }
        }
    }

    sat_solver_set_polarity( p.pSat, Vec_IntArray(p.vPolar), Vec_IntSize(p.vPolar) );
    return 1;
}

/**Function*************************************************************

  Synopsis    []

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func sat_solver_add_and1(pSat *sat_solver, iVar int32, iVar0 int32, iVar1 int32, fCompl0 int32, fCompl1 int32, fCompl int32) int32 {
    var Lits [3]lit;
    var Cid int32;

    Lits[0] = toLitCond( iVar, !fCompl );
    Lits[1] = toLitCond( iVar0, fCompl0 );
    Cid = sat_solver_addclause( pSat, Lits, Lits + 2 );
    assert((Cid != 0), "Cid");

    Lits[0] = toLitCond( iVar, !fCompl );
    Lits[1] = toLitCond( iVar1, fCompl1 );
    Cid = sat_solver_addclause( pSat, Lits, Lits + 2 );
    assert((Cid != 0), "Cid");
/*
    Lits[0] = toLitCond( iVar, fCompl );
    Lits[1] = toLitCond( iVar0, !fCompl0 );
    Lits[2] = toLitCond( iVar1, !fCompl1 );
    Cid = sat_solver_addclause( pSat, Lits, Lits + 3 );
    assert( Cid );
*/
    return 3;
}

func sat_solver_add_and2(pSat *sat_solver, iVar int32, iVar0 int32, iVar1 int32, fCompl0 int32, fCompl1 int32, fCompl int32) int32 {
    var Lits [3]lit;
    var Cid int32;
/*
    Lits[0] = toLitCond( iVar, !fCompl );
    Lits[1] = toLitCond( iVar0, fCompl0 );
    Cid = sat_solver_addclause( pSat, Lits, Lits + 2 );
    assert( Cid );

    Lits[0] = toLitCond( iVar, !fCompl );
    Lits[1] = toLitCond( iVar1, fCompl1 );
    Cid = sat_solver_addclause( pSat, Lits, Lits + 2 );
    assert( Cid );
*/
    Lits[0] = toLitCond( iVar, fCompl );
    Lits[1] = toLitCond( iVar0, !fCompl0 );
    Lits[2] = toLitCond( iVar1, !fCompl1 );
    Cid = sat_solver_addclause( pSat, Lits, Lits + 3 );
    assert((Cid != 0), "Cid");
    return 3;
}

/**Function*************************************************************

  Synopsis    [Adds a general cardinality constraint in terms of vVars.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Sbm_AddSorter(p *sat_solver, pVars *int32, i int32, k int32, pnVars *int32) {
    var iVar1 int32  = (*pnVars)++;
    var iVar2 int32  = (*pnVars)++;
    var fVerbose int32;
    if fVerbose != 0 {
        var v int32;
        for  v = 0; v < i; v++  {
            fmt.Printf(" ");
        }
        fmt.Printf("<");
        for  v = i+1; v < k; v++  {
            fmt.Printf("-");
        }
        fmt.Printf(">");
        for  v = k+1; v < 8; v++  {
            fmt.Printf(" ");
        }
        fmt.Printf("    ");
        fmt.Printf("[%3d :%3d ] -> [%3d :%3d ]\n", pVars[i], pVars[k], iVar1, iVar2);
    }
//    sat_solver_add_and1( p, iVar1, pVars[i], pVars[k], 1, 1, 1 );
//    sat_solver_add_and2( p, iVar2, pVars[i], pVars[k], 0, 0, 0 );
    sat_solver_add_half_sorter( p, iVar1, iVar2, pVars[i], pVars[k] );
    pVars[i] = iVar1;
    pVars[k] = iVar2;
}
func Sbm_AddCardinConstrMerge(p *sat_solver, pVars *int32, lo int32, hi int32, r int32, pnVars *int32) {
    var i int32;
    var step int32 = r * 2;
    if step < hi - lo {
        Sbm_AddCardinConstrMerge( p, pVars, lo, hi-r, step, pnVars );
        Sbm_AddCardinConstrMerge( p, pVars, lo+r, hi, step, pnVars );
        for  i = lo+r; i < hi-r; i += step  {
            Sbm_AddSorter( p, pVars, i, i+r, pnVars );
        }
        for  i = lo+r; i < hi-r-1; i += r  {
            var Lits [2]lit = [2]lit{ Abc_Var2Lit(pVars[i], 0), Abc_Var2Lit(pVars[i+r], 1) };
            var Cid int32  = sat_solver_addclause( p, Lits, Lits + 2 );
            assert((Cid != 0), "Cid");
        }
    }
}
func Sbm_AddCardinConstrRange(p *sat_solver, pVars *int32, lo int32, hi int32, pnVars *int32) {
    if hi - lo >= 1 {
        var i int32;
        var mid int32 = lo + (hi - lo) / 2;
        for  i = lo; i <= mid; i++  {
            Sbm_AddSorter( p, pVars, i, i + (hi - lo + 1) / 2, pnVars );
        }
        Sbm_AddCardinConstrRange( p, pVars, lo, mid, pnVars );
        Sbm_AddCardinConstrRange( p, pVars, mid+1, hi, pnVars );
        Sbm_AddCardinConstrMerge( p, pVars, lo, hi, 1, pnVars );
    }
}
func Sbm_AddCardinConstrPairWise(p *sat_solver, vVars *Vec_Int_t, K int32) int32 {
    var nVars int32  = Vec_IntSize(vVars);
    Sbm_AddCardinConstrRange( p, Vec_IntArray(vVars), 0, nVars - 1, &nVars );
    sat_solver_bookmark( p );
    return nVars;
}
func Sbm_AddCardinSolver(LogN int32, pvVars **Vec_Int_t) *sat_solver {
    var nVars int32  = 1 << LogN;
    var nVarsAlloc int32 = nVars + 2 * (nVars * LogN * (LogN-1) / 4 + nVars - 1);
    var nVarsReal int32;
    var vVars *Vec_Int_t  = Vec_IntStartNatural( nVars );
    var pSat *sat_solver  = sat_solver_new();
    sat_solver_setnvars( pSat, nVarsAlloc );
    nVarsReal = Sbm_AddCardinConstrPairWise( pSat, vVars, 2 );
    assert(nVarsReal == nVarsAlloc, "nVarsReal == nVarsAlloc");
    *pvVars = vVars;
    return pSat;
}

func Sbm_AddCardinConstrTest() {
    var LogN int32 = 3;
    var nVars int32 = 1 << LogN;
    var K int32 = 2;
    var Count int32 = 1;
    var vVars *Vec_Int_t;
    var vLits *Vec_Int_t = Vec_IntAlloc( nVars );
    var pSat *sat_solver  = Sbm_AddCardinSolver( LogN, &vVars );
    var nVarsReal int32  = sat_solver_nvars( pSat );

    var Lit int32  = Abc_Var2Lit( Vec_IntEntry(vVars, K), 1 );
    fmt.Printf("LogN = %d. N = %3d.   Vars = %5d. Clauses = %6d.  Comb = %d.\n", LogN, nVars, nVarsReal, sat_solver_nclauses(pSat), nVars * (nVars-1)/2 + nVars + 1);
    for {
        var i int32;
        var status int32 = sat_solver_solve( pSat, &Lit, &Lit+1, 0, 0, 0, 0 );
        if status != l_True {
            break; }
        Vec_IntClear( vLits );
        fmt.Printf("%3d : ", Count++);
        for  i = 0; i < nVars; i++  {
            Vec_IntPush( vLits, Abc_Var2Lit(i, sat_solver_var_value(pSat, i)) );
            fmt.Printf("%d", sat_solver_var_value(pSat, i));
        }
        fmt.Printf("\n");
        status = sat_solver_addclause( pSat, Vec_IntArray(vLits), Vec_IntArray(vLits) + nVars );
        if status == 0 {
            break; }
    }

    sat_solver_delete( pSat );
    Vec_IntFree( vVars );
    Vec_IntFree( vLits );
}

/**Function*************************************************************

  Synopsis    []

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Sbm_ManAlloc(LogN int32) *Sbm_Man_t {
    // inlining of macro ABC_CALLOC
    var p *Sbm_Man_t  = &Sbm_Man_t{};
    p.pSat = Sbm_AddCardinSolver( LogN, &p.vCardVars );
    p.LogN = LogN;
    p.FirstVar  = sat_solver_nvars( p.pSat );
    // window
    p.vRoots    = Vec_IntAlloc( 100 );
    p.vCuts     = Vec_WecAlloc( 1000 );
    p.vObjCuts  = Vec_WecAlloc( 1000 );
    p.vSolCuts  = Vec_IntAlloc( 100 );
    p.vCutGates = Vec_IntAlloc( 100 );
    p.vCutAreas = Vec_WrdAlloc( 100 );
    // solver
    p.vAssump   = Vec_IntAlloc( 100 );
    p.vPolar    = Vec_IntAlloc( 100 );
    // timing
    p.vArrs     = Vec_IntAlloc( 100 );
    p.vReqs     = Vec_IntAlloc( 100 );
    // internal
    p.vLit2Used = Vec_IntAlloc( 100 );
    p.vDelays   = Vec_IntAlloc( 100 );
    p.vReason   = Vec_IntAlloc( 100 );
    return p;
}

func Sbm_ManStop(p *Sbm_Man_t) {
    sat_solver_delete( p.pSat );
    Vec_IntFree( p.vCardVars );
    // internal
    Vec_IntFree( p.vRoots );
    Vec_WecFree( p.vCuts );
    Vec_WecFree( p.vObjCuts );
    Vec_IntFree( p.vSolCuts );
    Vec_IntFree( p.vCutGates );
    Vec_WrdFree( p.vCutAreas );
    // internal
    Vec_IntFree( p.vAssump );
    Vec_IntFree( p.vPolar );
    // internal
    Vec_IntFree( p.vArrs );
    Vec_IntFree( p.vReqs );
    // internal
    Vec_IntFree( p.vLit2Used );
    Vec_IntFree( p.vDelays );
    Vec_IntFree( p.vReason );
    // inlining of macro ABC_FREE
    p = nil;
}

func Sbm_ManTestSat(pMan any) int32 {
    var clk abctime = Abc_Clock();
    var clk2 abctime;
    var i int32;
    var k int32;
    var Lit int32;
    var LogN int32 = 7;
    var nVars int32 = 1 << LogN;
    var status int32;
    var Root int32;
    var p *Sbm_Man_t  = Sbm_ManAlloc( LogN );
    var InvArea word;
    var fKeepTrying int32  = 1;
    var StartSol int32;
    // derive window
    extern int Nf_ManExtractWindow( void * pMan, Vec_Int_t * vRoots, Vec_Wec_t * vCuts, Vec_Wec_t * vObjCuts, Vec_Int_t * vSolCuts, Vec_Int_t * vCutGates, Vec_Wrd_t * vCutAreas, word * pInvArea, int StartVar, int nVars );
    p.nInputs = Nf_ManExtractWindow( pMan, p.vRoots, p.vCuts, p.vObjCuts, p.vSolCuts, p.vCutGates, p.vCutAreas, &InvArea, p.FirstVar, nVars );
    p.LitShift = 2*p.nInputs+2;
    assert(Vec_WecSize(p.vObjCuts) + p.nInputs <= nVars, "Vec_WecSize(p.vObjCuts) + p.nInputs <= nVars");

    // print-out
//    Vec_WecPrint( p->vCuts, 0 );
//    printf( "\n" );
//    Vec_WecPrint( p->vObjCuts, 0 );
//    printf( "\n" );
    Vec_IntPrint( p.vSolCuts );
    fmt.Printf("All clauses = %d.  Multi clauses = %d.  Binary clauses = %d.  Other clauses = %d.\n", sat_solver_nclauses(p.pSat), Vec_WecSize(p.vObjCuts), Vec_WecSizeSize(p.vCuts), sat_solver_nclauses(p.pSat) - Vec_WecSize(p.vObjCuts) - Vec_WecSizeSize(p.vCuts));

    // creating CNF
    if !(Sbm_ManCreateCnf(p) != 0) {
        return 0; }

    // create assumptions
    // cardinality
    Vec_IntClear( p.vAssump );
    Vec_IntPush( p.vAssump, -1 );
    // unused variables
    for  i = Vec_WecSize(p.vObjCuts) + p.nInputs; i < nVars; i++  {
        Vec_IntPush( p.vAssump, Abc_Var2Lit(i, 1) );
    }
    // root variables
    // inlining of macro Vec_IntForEachEntry
    for  i = 0; (i < Vec_IntSize(p.vRoots)) && (((Root) = Vec_IntEntry(p.vRoots, i)), 1); i++  {
        Vec_IntPush( p.vAssump, Abc_Var2Lit(Root, 0) );
    }
//    Vec_IntPrint( p->vAssump );

    StartSol = Vec_IntSize(p.vSolCuts);
//    StartSol = 30;
    for  (fKeepTrying != 0) && StartSol-fKeepTrying > 0  {
        fmt.Printf("Trying to find mapping with %d gates.\n", StartSol-fKeepTrying);
    //    for ( i = Vec_IntSize(p->vSolCuts)-5; i < nVars; i++ )
    //        Vec_IntPush( p->vAssump, Abc_Var2Lit(Vec_IntEntry(p->vCardVars, i), 1) );
        Vec_IntWriteEntry( p.vAssump, 0, Abc_Var2Lit(Vec_IntEntry(p.vCardVars, StartSol-fKeepTrying), 1) );
        // solve the problem
        clk2 = Abc_Clock();
        status = sat_solver_solve( p.pSat, Vec_IntArray(p.vAssump), Vec_IntLimit(p.vAssump), 0, 0, 0, 0 );
        if status == l_True {
            var AreaNew word;
            var Count int32;
            fmt.Printf("AND Lits = %d.  Inputs = %d.  Vars = %d.  All vars = %d.\n", Vec_WecSize(p.vObjCuts), p.nInputs, Vec_WecSize(p.vObjCuts) + p.nInputs, nVars);
//            for ( i = 0; i < nVars; i++ )
//                printf( "%d", sat_solver_var_value(p->pSat, i) );
//            printf( "\n" );
            for  i = 0; i < nVars; i++  {
                if sat_solver_var_value(p.pSat, i) {
                    fmt.Printf("%d=%d ", i, sat_solver_var_value(p.pSat, i));
                    Count++;
                    if i >= Vec_WecSize(p.vObjCuts) {
                        AreaNew += InvArea; }
                }
            }
            fmt.Printf("Count = %d\n", Count);
//            for ( i = p->FirstVar; i < sat_solver_nvars(p->pSat); i++ )
//                printf( "%d", sat_solver_var_value(p->pSat, i) );
//            printf( "\n" );
            Count = 1;
            for  i = p.FirstVar; i < sat_solver_nvars(p.pSat); i++  {
                if sat_solver_var_value(p.pSat, i) {
                    var vCutOne *Vec_Int_t  = Vec_WecEntry(p.vCuts, i-p.FirstVar);
                    fmt.Printf("%2d : Cut %3d  (Gate %2d)  ", Count, i-p.FirstVar, Vec_IntEntry(p.vCutGates, i-p.FirstVar));
                    // inlining of macro Vec_IntForEachEntry
                    for  k = 0; (k < Vec_IntSize(vCutOne)) && (((Lit) = Vec_IntEntry(vCutOne, k)), 1); k++  {
                        fmt.Printf("%d(%d) ", Lit - 2*(p.nInputs+1), Abc_Lit2Var(Lit));
                    }
                    fmt.Printf("\n");
                    Count++;
                    AreaNew += Vec_WrdEntry(p.vCutAreas, i-p.FirstVar);
                }
            }
            fmt.Printf("Area = %7.2f\n", Scl_Int2Flt(int32(AreaNew)));
        }
        if status == l_False {
            fmt.Printf("UNSAT ");
fKeepTrying = 0;
        } else if status == l_True {
            fmt.Printf("SAT   ");
fKeepTrying++; }
        Abc_PrintTime( 1, "Time", Abc_Clock() - clk2 );
        Abc_PrintTime( 1, "Total time", Abc_Clock() - clk );
        fmt.Printf("\n");
    }
    Sbm_ManStop( p );
    return 1;
}


////////////////////////////////////////////////////////////////////////
///                       END OF FILE                                ///
////////////////////////////////////////////////////////////////////////


//ABC_NAMESPACE_IMPL_END

