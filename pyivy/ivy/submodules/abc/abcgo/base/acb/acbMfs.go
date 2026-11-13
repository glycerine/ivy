package acb

import (
	"fmt"
	"os"
)

/**CFile****************************************************************

  FileName    [abcMfs.c]

  SystemName  [ABC: Logic synthesis and verification system.]

  PackageName [Hierarchical word-level netlist.]

  Synopsis    [Optimization with don't-cares.]

  Author      [Alan Mishchenko]
  
  Affiliation [UC Berkeley]

  Date        [Ver. 1.0. Started - July 21, 2015.]

  Revision    [$Id: abcMfs.c,v 1.00 2014/11/29 00:00:00 alanmi Exp $]

***********************************************************************/

//#include "acb.h"
//#include "bool/kit/kit.h"
//#include "sat/bsat/satSolver.h"
//#include "sat/cnf/cnf.h"
//#include "misc/util/utilTruth.h"
//#include "acbPar.h"

//ABC_NAMESPACE_IMPL_START

////////////////////////////////////////////////////////////////////////
///                        DECLARATIONS                              ///
////////////////////////////////////////////////////////////////////////

func Acb_ObjIsDelayCriticalFanin(p *Acb_Ntk_t, i int32, f int32) int32 { return !Acb_ObjIsCi(p, f) && Acb_ObjLevelR(p, i) + Acb_ObjLevelD(p, f) == p.LevelMax; }
func Acb_ObjIsAreaCritical(p *Acb_Ntk_t, f int32) int32 { return !Acb_ObjIsCi(p, f) && Acb_ObjFanoutNum(p, f) == 1;                              }
func Acb_ObjIsCritical(p *Acb_Ntk_t, i int32, f int32, fDel int32) int32 { if fDel != 0 {
    return Acb_ObjIsDelayCriticalFanin(p, i, f);
} else {
    return Acb_ObjIsAreaCritical(p, f);
}      }

////////////////////////////////////////////////////////////////////////
///                     FUNCTION DEFINITIONS                         ///
////////////////////////////////////////////////////////////////////////

/**Function*************************************************************

  Synopsis    [Derive CNF for nodes in the window.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Acb_DeriveCnfFromTruth(Truth word, nVars int32, vCover *Vec_Int_t, vCnf *Vec_Str_t) int32 {
    Vec_StrClear( vCnf );
    if Truth == 0 || ~Truth == 0 {
//        assert( nVars == 0 );
        Vec_StrPush( vCnf, byte((Truth == 0)) );
        Vec_StrPush( vCnf, byte(-1) );
        return 1;
    
    } else {
        var i, k, c, RetValue, Literal, Cube, nCubes int32;
        assert(nVars > 0, "nVars > 0");
        for  c = 0; c < 2; c ++  {
            if c != 0 {
                Truth = ~Truth;
            } else {
                Truth = Truth;
            }
            RetValue = Kit_TruthIsop( (unsigned *)&Truth, nVars, vCover, 0 );
            assert(RetValue == 0, "RetValue == 0");
            nCubes += Vec_IntSize( vCover );
            Vec_IntForEachEntry( vCover, Cube, i )
            {
                for  k = 0; k < nVars; k++  {
                    Literal = 3 & (Cube >> (k << 1));
                    if Literal == 1 {      // '0'  -> pos lit
                        Vec_StrPush( vCnf, byte(Abc_Var2Lit(k, 0)) );
                    } else if Literal == 2 { // '1'  -> neg lit
                        Vec_StrPush( vCnf, byte(Abc_Var2Lit(k, 1)) );
                    } else if Literal != 0 {
                        assert(false, "0"); }
                }
                Vec_StrPush( vCnf, byte(Abc_Var2Lit(nVars, c)) );
                Vec_StrPush( vCnf, byte(-1) );
            }
        }
        return nCubes;
    }
}

func Acb_DeriveCnfForWindowOne(p *Acb_Ntk_t, iObj int32) {
    var vCnfs []Vec_Str_t = p.vCnfs;
    var vCnfBase *Vec_Str_t  = Acb_ObjCnfs( p, iObj );
    assert(Vec_StrSize(vCnfBase) == 0, "Vec_StrSize(vCnfBase) == 0"); // unassigned
    assert(len(vCnfs) == int(Acb_NtkObjNumMax(p)), "len(vCnfs) == int(Acb_NtkObjNumMax(p))");
    Acb_DeriveCnfFromTruth( Acb_ObjTruth(p, iObj), Acb_ObjFaninNum(p, iObj), &p.vCover, &p.vCnf );
    Vec_StrGrow( vCnfBase, Vec_StrSize(&p.vCnf) );
    memcpy( Vec_StrArray(vCnfBase), Vec_StrArray(&p.vCnf), (size_t)Vec_StrSize(&p.vCnf) );
    vCnfBase.nSize = Vec_StrSize(&p.vCnf);
}
func Acb_DeriveCnfForWindow(p *Acb_Ntk_t, vWin *Vec_Int_t, PivotVar int32) []Vec_Str_t {
    var vCnfs []Vec_Str_t = p.vCnfs;
    var vCnfBase *Vec_Str_t; var i, iObj int32;
    assert(len(vCnfs) == int(Acb_NtkObjNumMax(p)), "len(vCnfs) == int(Acb_NtkObjNumMax(p))");
    Vec_IntForEachEntry( vWin, iObj, i )
    {
        if Abc_LitIsCompl(iObj) && i < PivotVar {
            continue; }
        iObj = Abc_Lit2Var(iObj);
        vCnfBase = Acb_ObjCnfs( p, iObj );
        if Vec_StrSize(vCnfBase) > 0 {
            continue; }
        Acb_DeriveCnfForWindowOne( p, iObj );
    }
    return vCnfs;
}

/**Function*************************************************************

  Synopsis    [Constructs CNF for the window.]

  Description [The window for the pivot node is represented as a DFS ordered array 
  of objects (vWinObjs) whose indexes are used as SAT variable IDs (stored in p->vCopies).
  PivotVar is the index of the pivot node in array vWinObjs.
  The nodes before (after) PivotVar are TFI (TFO) nodes.
  The leaf (root) nodes are labeled with Abc_LitIsCompl().
  If fQbf is 1, returns the instance meant for QBF solving. It uses the last 
  variable (LastVar) as the placeholder for the second copy of the pivot node.]
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Acb_TranslateCnf(vClas *Vec_Int_t, vLits *Vec_Int_t, vCnf *Vec_Str_t, vSatVars *Vec_Int_t, iPivotVar int32) {
    signed char Entry;
    var i, Lit int32;
    Vec_StrForEachEntry( vCnf, Entry, i )
    {
        if int32(Entry) == -1 {
            Vec_IntPush( vClas, Vec_IntSize(vLits) );
            continue;
        }
        Lit = Abc_Lit2LitV( Vec_IntArray(vSatVars), int32(Entry) );
        Lit = Abc_LitNotCond( Lit, Abc_Lit2Var(Lit) == iPivotVar );
        Vec_IntPush( vLits, Lit );
    }
}
func Acb_NtkCountRoots(vWinObjs *Vec_Int_t, PivotVar int32) int32 {
    var i, iObjLit, nRoots int32;
    Vec_IntForEachEntryStart( vWinObjs, iObjLit, i, PivotVar + 1 )
        nRoots += Abc_LitIsCompl(iObjLit);
    return nRoots;
}
func Acb_DeriveCnfForNode(p *Acb_Ntk_t, iObj int32, pSat *sat_solver, OutVar int32) {
    var vCnfs []Vec_Str_t = p.vCnfs;
    var vFaninVars *Vec_Int_t  = &p.vCover;
    var vClas *Vec_Int_t  = Vec_IntAlloc( 100 );
    var vLits *Vec_Int_t  = Vec_IntAlloc( 100 );
    var k int32;
    var iFanin int32;
    var pFanins *int32;
    var Prev int32;
    var This int32;
    // collect SAT variables
    Vec_IntClear( vFaninVars );
    Acb_ObjForEachFaninFast( p, iObj, pFanins, iFanin, k )
    {
        assert(Acb_ObjFunc(p, iFanin) >= 0, "Acb_ObjFunc(p, iFanin) >= 0");
        Vec_IntPush( vFaninVars, Acb_ObjFunc(p, iFanin) );
    }
    Vec_IntPush( vFaninVars, OutVar );
    // derive CNF for the node
    Acb_TranslateCnf( vClas, vLits, &vCnfs[iObj], vFaninVars, -1 );
    // add clauses
    Prev = 0;
    Vec_IntForEachEntry( vClas, This, k )
    {
        if !sat_solver_addclause( pSat, Vec_IntArray(vLits) + Prev, Vec_IntArray(vLits) + This ) {
            fmt.Printf("Error: SAT solver became UNSAT at a wrong place (while adding new CNF).\n"); }
        Prev = This;
    }
    Vec_IntFree( vClas );
    Vec_IntFree( vLits );
}
func Acb_NtkWindow2Cnf(p *Acb_Ntk_t, vWinObjs *Vec_Int_t, Pivot int32) *Cnf_Dat_t {
    var pCnf *Cnf_Dat_t;
    var vFaninVars *Vec_Int_t  = Vec_IntAlloc( 8 );
    var PivotVar int32  = Vec_IntFind(vWinObjs, Abc_Var2Lit(Pivot, 0));
    var nRoots int32    = Acb_NtkCountRoots(vWinObjs, PivotVar);
    var TfoStart int32  = PivotVar + 1;
    var nTfoSize int32  = Vec_IntSize(vWinObjs) - TfoStart;
    var nVarsAll int32  = Vec_IntSize(vWinObjs) + nTfoSize + nRoots;
    var i int32;
    var k int32;
    var iObj int32;
    var iObjLit int32;
    var iFanin int32;
    var pFanins *int32;
    var Entry int32;
    var vCnfs []Vec_Str_t = Acb_DeriveCnfForWindow( p, vWinObjs, PivotVar );
    var vClas *Vec_Int_t  = Vec_IntAlloc( 100 );
    var vLits *Vec_Int_t  = Vec_IntAlloc( 1000 );
    // mark new SAT variables
    Vec_IntForEachEntry( vWinObjs, iObj, i )
        Acb_ObjSetFunc( p, Abc_Lit2Var(iObj), i );
    // add clauses for all nodes
    Vec_IntPush( vClas, Vec_IntSize(vLits) );
    Vec_IntForEachEntry( vWinObjs, iObjLit, i )
    {
        if Abc_LitIsCompl(iObjLit) && i < PivotVar {
            continue; }
        iObj = Abc_Lit2Var(iObjLit);
        assert(!Acb_ObjIsCio(p, iObj), "!Acb_ObjIsCio(p, iObj)");
        // collect SAT variables
        Vec_IntClear( vFaninVars );
        Acb_ObjForEachFaninFast( p, iObj, pFanins, iFanin, k )
            Vec_IntPush( vFaninVars, Acb_ObjFunc(p, iFanin) );
        Vec_IntPush( vFaninVars, Acb_ObjFunc(p, iObj) );
        // derive CNF for the node
        Acb_TranslateCnf( vClas, vLits, &vCnfs[iObj], vFaninVars, -1 );
    }
    // add second clauses for the TFO
    Vec_IntForEachEntryStart( vWinObjs, iObjLit, i, TfoStart )
    {
        iObj = Abc_Lit2Var(iObjLit);
        assert(!Acb_ObjIsCio(p, iObj), "!Acb_ObjIsCio(p, iObj)");
        // collect SAT variables
        Vec_IntClear( vFaninVars );
        Acb_ObjForEachFaninFast( p, iObj, pFanins, iFanin, k )
            Vec_IntPush( vFaninVars, Acb_ObjFunc(p, iFanin) + (Acb_ObjFunc(p, iFanin) > PivotVar) * nTfoSize );
        Vec_IntPush( vFaninVars, Acb_ObjFunc(p, iObj) + nTfoSize );
        // derive CNF for the node
        Acb_TranslateCnf( vClas, vLits, &vCnfs[iObj], vFaninVars, PivotVar );
    }
    if nRoots > 0 {
        // create XOR clauses for the roots
        var nVars int32  = Vec_IntSize(vWinObjs) + nTfoSize;
        Vec_IntClear( vFaninVars );
        Vec_IntForEachEntryStart( vWinObjs, iObjLit, i, TfoStart )
        {
            if !Abc_LitIsCompl(iObjLit) {
                continue; }
            iObj = Abc_Lit2Var(iObjLit);
            // add clauses
            Vec_IntPushThree( vLits, Abc_Var2Lit(Acb_ObjFunc(p, iObj), 1), Abc_Var2Lit(Acb_ObjFunc(p, iObj) + nTfoSize, 0), Abc_Var2Lit(nVars, 0) );
            Vec_IntPush( vClas, Vec_IntSize(vLits) );
            Vec_IntPushThree( vLits, Abc_Var2Lit(Acb_ObjFunc(p, iObj), 0), Abc_Var2Lit(Acb_ObjFunc(p, iObj) + nTfoSize, 1), Abc_Var2Lit(nVars, 0) );
            Vec_IntPush( vClas, Vec_IntSize(vLits) );
            Vec_IntPushThree( vLits, Abc_Var2Lit(Acb_ObjFunc(p, iObj), 0), Abc_Var2Lit(Acb_ObjFunc(p, iObj) + nTfoSize, 0), Abc_Var2Lit(nVars, 1) );
            Vec_IntPush( vClas, Vec_IntSize(vLits) );
            Vec_IntPushThree( vLits, Abc_Var2Lit(Acb_ObjFunc(p, iObj), 1), Abc_Var2Lit(Acb_ObjFunc(p, iObj) + nTfoSize, 1), Abc_Var2Lit(nVars, 1) );
            Vec_IntPush( vClas, Vec_IntSize(vLits) );
            Vec_IntPush( vFaninVars, Abc_Var2Lit(nVars++, 0) );
        }
        Vec_IntAppend( vLits, vFaninVars );
        Vec_IntPush( vClas, Vec_IntSize(vLits) );
        assert(nRoots == Vec_IntSize(vFaninVars), "nRoots == Vec_IntSize(vFaninVars)");
        assert(nVars == nVarsAll, "nVars == nVarsAll");
    }
    Vec_IntFree( vFaninVars );
    // create CNF structure
    pCnf = new(Cnf_Dat_t);
    pCnf.nVars     = nVarsAll;
    pCnf.nClauses  = Vec_IntSize(vClas)-1;
    pCnf.nLiterals = Vec_IntSize(vLits);
    pCnf.pClauses  = ABC_ALLOC( int *, Vec_IntSize(vClas) );
    pCnf.pClauses[0] = Vec_IntReleaseArray(vLits);
    Vec_IntForEachEntry( vClas, Entry, i )
        pCnf.pClauses[i] = pCnf.pClauses[0] + Entry;
    // cleanup
    Vec_IntFree( vClas );
    Vec_IntFree( vLits );
    //Cnf_DataPrint( pCnf, 1 );
    return pCnf;
}
func Acb_NtkWindowUndo(p *Acb_Ntk_t, vWin *Vec_Int_t) {
    var i, iObj int32;
    Vec_IntForEachEntry( vWin, iObj, i )
    {
        assert(Vec_IntEntry(&p.vObjFunc, Abc_Lit2Var(iObj)) != -1, "Vec_IntEntry(&p.vObjFunc, Abc_Lit2Var(iObj)) != -1");
        Vec_IntWriteEntry( &p.vObjFunc, Abc_Lit2Var(iObj), -1 );
    }
}

/**Function*************************************************************

  Synopsis    [Creates SAT solver containing several copies of the window.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Acb_NtkWindow2Solver(pSat *sat_solver, pCnf *Cnf_Dat_t, vFlip *Vec_Int_t, PivotVar int32, nDivs int32, nTimes int32) int32 {
    var n int32;
    var i int32;
    var RetValue int32;
    var Test int32 = pCnf.pClauses[0][0];
    var nGroups int32;
    if nTimes <= 2 {
        nGroups = nTimes-1;
    } else {
        nGroups = 2;
    }
    var nRounds int32;
    if nTimes <= 2 {
        nRounds = nTimes-1;
    } else {
        nRounds = nTimes;
    }
    assert(sat_solver_nvars(pSat) == 0, "sat_solver_nvars(pSat) == 0");
    sat_solver_setnvars( pSat, nTimes * pCnf.nVars + nGroups * nDivs + 2 );
    assert(nTimes == 1 || nTimes == 2 || nTimes == 6, "nTimes == 1 || nTimes == 2 || nTimes == 6");
    for  n = 0; n < nTimes; n++  {
        if n & 1 {
            Cnf_DataLiftAndFlipLits( pCnf, -pCnf.nVars, vFlip ); }
        for  i = 0; i < pCnf.nClauses; i++  {
            if !sat_solver_addclause( pSat, pCnf.pClauses[i], pCnf.pClauses[i+1] ) {
                fmt.Printf("Error: SAT solver became UNSAT at a wrong place.\n"); }
        }
        if n & 1 {
            Cnf_DataLiftAndFlipLits( pCnf, pCnf.nVars, vFlip ); }
        if n < nTimes - 1 {
            Cnf_DataLift( pCnf, pCnf.nVars );
        } else if n != 0 { // if ( n == nTimes - 1 )
            Cnf_DataLift( pCnf, -(nTimes - 1) * pCnf.nVars ); }
    }
    assert(Test == pCnf.pClauses[0][0], "Test == pCnf.pClauses[0][0]");
    // add conditional buffers
    for  n = 0; n < nRounds; n++  {
        var BaseA int32  = n * pCnf.nVars;
        var BaseB int32  = ((n + 1) % nTimes) * pCnf.nVars;
        var BaseC int32  = nTimes * pCnf.nVars + (n & 1) * nDivs;
        for  i = 0; i < nDivs; i++  {
            sat_solver_add_buffer_enable( pSat, BaseA + i, BaseB + i, BaseC + i, 0 );
        }
    }
    // finalize
    RetValue = sat_solver_simplify( pSat );
    if !(RetValue != 0) { fmt.Printf("Error: SAT solver became UNSAT at a wrong place.\n"); }
    return 1;
}

/**Function*************************************************************

  Synopsis    [Computes function of the node]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Acb_ComputeFunction(pSat *sat_solver, PivotVar int32, FreeVar int32, vDivVars *Vec_Int_t, fCompl int32) word {
    var fExpand int32;
    var uCube, uTruth word;
    var vTempLits *Vec_Int_t  = Vec_IntAlloc( 100 );
    var status int32;
    var i int32;
    var iVar int32;
    var iLit int32;
    var nFinal int32;
    var pFinal *int32;
    var pLits [2]int32;
    assert(FreeVar < sat_solver_nvars(pSat), "FreeVar < sat_solver_nvars(pSat)");
//    if ( fCompl )
//        pLits[0] = Abc_Var2Lit( sat_solver_nvars(pSat)-2, 0 ); // F = 1
//    else
        pLits[0] = Abc_Var2Lit( PivotVar, fCompl ); // F = 1
    pLits[1] = Abc_Var2Lit( FreeVar, 0 );  // iNewLit
    for {
        // find onset minterm
        status = sat_solver_solve( pSat, pLits, pLits + 2, 0, 0, 0, 0 );
        if status == l_False {
            Vec_IntFree( vTempLits );
            return uTruth;
        }
        assert(status == l_True, "status == l_True");
        if fExpand != 0 {
            // collect divisor literals
            Vec_IntFill( vTempLits, 1, Abc_LitNot(pLits[0]) ); // F = 0
            Vec_IntForEachEntry( vDivVars, iVar, i )
                Vec_IntPush( vTempLits, sat_solver_var_literal(pSat, iVar) );
            // check against offset
            status = sat_solver_solve( pSat, Vec_IntArray(vTempLits), Vec_IntLimit(vTempLits), 0, 0, 0, 0 );
            if status != l_False {
                fmt.Printf("Failed internal check during function comptutation.\n"); }
            assert(status == l_False, "status == l_False");
            // compute cube and add clause
            nFinal = sat_solver_final( pSat, &pFinal );
            Vec_IntFill( vTempLits, 1, Abc_LitNot(pLits[1]) ); // NOT(iNewLit)
            for  i = 0; i < nFinal; i++  {
                if pFinal[i] != pLits[0] {
                    Vec_IntPush( vTempLits, pFinal[i] ); }
            }
        
        } else {
            // collect divisor literals
            Vec_IntFill( vTempLits, 1, Abc_LitNot(pLits[1]) );// NOT(iNewLit)
            Vec_IntForEachEntry( vDivVars, iVar, i )
                Vec_IntPush( vTempLits, Abc_LitNot(sat_solver_var_literal(pSat, iVar)) );
        }
        uCube = ~(word)0;
        Vec_IntForEachEntryStart( vTempLits, iLit, i, 1 )
        {
            iVar = Vec_IntFind( vDivVars, Abc_Lit2Var(iLit) );   assert(iVar >= 0, "iVar >= 0");
            uCube &= Abc_LitIsCompl(iLit) ? s_Truths6[iVar] : ~s_Truths6[iVar];
        }
        uTruth |= uCube;
        status = sat_solver_addclause( pSat, Vec_IntArray(vTempLits), Vec_IntLimit(vTempLits) );
        if status == 0 {
            Vec_IntFree( vTempLits );
            return uTruth;
        }
    }
    Vec_IntFree( vTempLits );
    assert(false, "0"); 
    return ~(word)0;
}





/**Function*************************************************************

  Synopsis    []

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Acb_NtkPrintVec(p *Acb_Ntk_t, vVec *Vec_Int_t, pName string) {
    var i int32;
    fmt.Printf("%s: ", pName);
    for  i = 0; i < vVec.nSize; i++  {
        fmt.Printf("%d ", vVec.pArray[i]);
    }
    fmt.Printf("\n");
}
func Acb_NtkPrintVec2(p *Acb_Ntk_t, vVec *Vec_Int_t, pName string) {
    var i int32;
    fmt.Printf("%s: \n", pName);
    for  i = 0; i < vVec.nSize; i++  {
        Acb_NtkPrintNode( p, vVec.pArray[i] );
    }
    fmt.Printf("\n");
}
func Acb_NtkPrintVecWin(p *Acb_Ntk_t, vVec *Vec_Int_t, pName string) {
    var i int32;
    fmt.Printf("%s: \n", pName);
    for  i = 0; i < vVec.nSize; i++  {
        Acb_NtkPrintNode( p, Abc_Lit2Var(vVec.pArray[i]) );
    }
    fmt.Printf("\n");
}

/**Function*************************************************************

  Synopsis    [Collects divisors in a non-topo order.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Acb_NtkDivisors_rec(p *Acb_Ntk_t, iObj int32, nTfiLevMin int32, vDivs *Vec_Int_t) {
    var k int32;
    var iFanin int32;
    var pFanins *int32;
//    if ( !Acb_ObjIsCi(p, iObj) && Acb_ObjLevelD(p, iObj) < nTfiLevMin )
    if !Acb_ObjIsCi(p, iObj) && nTfiLevMin < 0 {
        return; }
    if Acb_ObjSetTravIdCur(p, iObj) {
        return; }
    Acb_ObjForEachFaninFast( p, iObj, pFanins, iFanin, k )
        Acb_NtkDivisors_rec( p, iFanin, nTfiLevMin-1, vDivs );
    Vec_IntPush( vDivs, iObj );
}
func Acb_NtkDivisors(p *Acb_Ntk_t, Pivot int32, nTfiLevMin int32, fDelay int32) *Vec_Int_t {
    var k int32;
    var iFanin int32;
    var pFanins *int32;
    var vDivs *Vec_Int_t  = Vec_IntAlloc( 100 );
    Acb_NtkIncTravId( p );
//    if ( fDelay ) // delay-oriented
    if false { // delay-oriented
    
        // start from critical fanins
        assert(Acb_ObjLevelD( p, Pivot ) > 1, "Acb_ObjLevelD( p, Pivot ) > 1");
        Acb_ObjForEachFaninFast( p, Pivot, pFanins, iFanin, k )
            if ( Acb_ObjIsDelayCriticalFanin( p, Pivot, iFanin ) )
                Acb_NtkDivisors_rec( p, iFanin, nTfiLevMin, vDivs );
        // add non-critical fanins
        Acb_ObjForEachFaninFast( p, Pivot, pFanins, iFanin, k )
            if ( !Acb_ObjIsDelayCriticalFanin( p, Pivot, iFanin ) )
                if ( !Acb_ObjSetTravIdCur(p, iFanin) )
                    Vec_IntPush( vDivs, iFanin );
    
    } else {
        Acb_NtkDivisors_rec( p, Pivot, nTfiLevMin, vDivs );
        assert(Vec_IntEntryLast(vDivs) == Pivot, "Vec_IntEntryLast(vDivs) == Pivot");
        Vec_IntPop( vDivs );
        // add remaining fanins of the node
        Acb_ObjForEachFaninFast( p, Pivot, pFanins, iFanin, k )
            if ( !Acb_ObjSetTravIdCur(p, iFanin) )
                Vec_IntPush( vDivs, iFanin );
/*
        // start from critical fanins
        assert( Acb_ObjLevelD( p, Pivot ) > 1 );
        Acb_ObjForEachFaninFast( p, Pivot, pFanins, iFanin, k )
            if ( Acb_ObjIsAreaCritical( p, iFanin ) )
                Acb_NtkDivisors_rec( p, iFanin, nTfiLevMin, vDivs );
        // add non-critical fanins
        Acb_ObjForEachFaninFast( p, Pivot, pFanins, iFanin, k )
            if ( !Acb_ObjIsAreaCritical( p, iFanin ) )
                if ( !Acb_ObjSetTravIdCur(p, iFanin) )
                    Vec_IntPush( vDivs, iFanin );
*/
    }
    return vDivs;
}

/**Function*************************************************************

  Synopsis    [Marks TFO of divisors.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Acb_ObjMarkTfo_rec(p *Acb_Ntk_t, iObj int32, nTfoLevMax int32, nFanMax int32, vMarked *Vec_Int_t) {
    var iFanout, i int32;
    if Acb_ObjSetTravIdCur(p, iObj) {
        return; }
    Vec_IntPush( vMarked, iObj );
    if Acb_ObjLevelD(p, iObj) > nTfoLevMax || Acb_ObjFanoutNum(p, iObj) > nFanMax {
        return; }
    Acb_ObjForEachFanout( p, iObj, iFanout, i )
        Acb_ObjMarkTfo_rec( p, iFanout, nTfoLevMax, nFanMax, vMarked );
}
func Acb_ObjMarkTfo(p *Acb_Ntk_t, vDivs *Vec_Int_t, Pivot int32, nTfoLevMax int32, nFanMax int32) *Vec_Int_t {
    var vMarked *Vec_Int_t  = Vec_IntAlloc( 1000 );
    var i, iObj int32;
    Acb_NtkIncTravId( p );
    Acb_ObjSetTravIdCur( p, Pivot );
    Vec_IntPush( vMarked, Pivot );
    Vec_IntForEachEntry( vDivs, iObj, i )
        Acb_ObjMarkTfo_rec( p, iObj, nTfoLevMax, nFanMax, vMarked );
    return vMarked;
}
func Acb_ObjMarkTfo2(p *Acb_Ntk_t, vMarked *Vec_Int_t) {
    var i, Node int32;
    Acb_NtkIncTravId( p );
    Vec_IntForEachEntry( vMarked, Node, i )
        Acb_ObjSetTravIdCur( p, Node );
}

/**Function*************************************************************

  Synopsis    [Labels TFO nodes with {none, root, inner} based on their type.]

  Description [Assuming TFO of TFI is marked with the current trav ID.]
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Acb_ObjLabelTfo_rec(p *Acb_Ntk_t, iObj int32, nTfoLevMax int32, nFanMax int32, fFirst int32) int32 {
    var iFanout, i, Diff, fHasNone int32;
    if (Diff = Acb_ObjTravIdDiff(p, iObj)) <= 2 {
        return Diff; }
    Acb_ObjSetTravIdDiff( p, iObj, 2 );
    if Acb_ObjIsCo(p, iObj) || Acb_ObjLevelD(p, iObj) > nTfoLevMax {
        return 2; }
    if Acb_ObjLevelD(p, iObj) == nTfoLevMax || Acb_ObjFanoutNum(p, iObj) > nFanMax {
        if Diff == 3 {  // belongs to TFO of TFI
            Acb_ObjSetTravIdDiff( p, iObj, 1 ); } // root
        return Acb_ObjTravIdDiff(p, iObj);
    }
    Acb_ObjForEachFanout( p, iObj, iFanout, i )
        if ( !fFirst || Acb_ObjIsDelayCriticalFanin(p, iFanout, iObj) )
            fHasNone |= 2 == Acb_ObjLabelTfo_rec( p, iFanout, nTfoLevMax, nFanMax, 0 );
    if (fHasNone != 0) && Diff == 3 {  // belongs to TFO of TFI
        Acb_ObjSetTravIdDiff( p, iObj, 1 ); // root
    } else if !(fHasNone != 0) {
        Acb_ObjSetTravIdDiff( p, iObj, 0 ); } // inner
    return Acb_ObjTravIdDiff(p, iObj);
}
func Acb_ObjLabelTfo(p *Acb_Ntk_t, Root int32, nTfoLevMax int32, nFanMax int32, fDelay int32) int32 {
    Acb_NtkIncTravId( p ); // none  (2)    marked (3)  unmarked (4)
    Acb_NtkIncTravId( p ); // root  (1)
    Acb_NtkIncTravId( p ); // inner (0)
    assert(Acb_ObjTravIdDiff(p, Root) > 2, "Acb_ObjTravIdDiff(p, Root) > 2");
    return Acb_ObjLabelTfo_rec( p, Root, nTfoLevMax, nFanMax, fDelay );
}

/**Function*************************************************************

  Synopsis    [Collects labeled TFO.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Acb_ObjDeriveTfo_rec(p *Acb_Ntk_t, iObj int32, vTfo *Vec_Int_t, vRoots *Vec_Int_t, fFirst int32) {
    var iFanout int32;
    var i int32;
    var Diff int32 = Acb_ObjTravIdDiff(p, iObj);
    if Acb_ObjSetTravIdCur(p, iObj) {
        return; }
    if Diff == 2 { // root
    
        Vec_IntPush( vRoots, iObj );
        Vec_IntPush( vTfo, iObj );
        return;
    }
    assert(Diff == 1, "Diff == 1");
    Acb_ObjForEachFanout( p, iObj, iFanout, i )
        if ( !fFirst || Acb_ObjIsDelayCriticalFanin(p, iFanout, iObj) )
            Acb_ObjDeriveTfo_rec( p, iFanout, vTfo, vRoots, 0 );
    Vec_IntPush( vTfo, iObj );
}
func Acb_ObjDeriveTfo(p *Acb_Ntk_t, Pivot int32, nTfoLevMax int32, nFanMax int32, pvTfo **Vec_Int_t, pvRoots **Vec_Int_t, fDelay int32) {
    var Res int32  = Acb_ObjLabelTfo( p, Pivot, nTfoLevMax, nFanMax, fDelay );
    var vTfo *Vec_Int_t    = *pvTfo   = Vec_IntAlloc( 10 );
    var vRoots *Vec_Int_t  = *pvRoots = Vec_IntAlloc( 10 );
    if Res != 0 { // none or root
        return; }
    Acb_NtkIncTravId( p ); // root (2)   inner (1)  visited (0)
    Acb_ObjDeriveTfo_rec( p, Pivot, vTfo, vRoots, fDelay );
    assert(Vec_IntEntryLast(vTfo) == Pivot, "Vec_IntEntryLast(vTfo) == Pivot");
    Vec_IntPop( vTfo );
    assert(Vec_IntEntryLast(vRoots) != Pivot, "Vec_IntEntryLast(vRoots) != Pivot");
    Vec_IntReverseOrder( vTfo );
    Vec_IntReverseOrder( vRoots );
}


/**Function*************************************************************

  Synopsis    [Collect side-inputs of the TFO, except the node.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Acb_NtkCollectTfoSideInputs(p *Acb_Ntk_t, Pivot int32, vTfo *Vec_Int_t) *Vec_Int_t {
    var vSide *Vec_Int_t   = Vec_IntAlloc( 100 );
    var i int32;
    var k int32;
    var Node int32;
    var iFanin int32;
    var pFanins *int32;
    Acb_NtkIncTravId( p );
    Vec_IntPush( vTfo, Pivot );
    Vec_IntForEachEntry( vTfo, Node, i )
        Acb_ObjSetTravIdCur( p, Node );
    Vec_IntForEachEntry( vTfo, Node, i )
        Acb_ObjForEachFaninFast( p, Node, pFanins, iFanin, k )
            if ( !Acb_ObjSetTravIdCur(p, iFanin) && iFanin != Pivot )
                Vec_IntPush( vSide, iFanin );
    Vec_IntPop( vTfo );
    return vSide;
}

/**Function*************************************************************

  Synopsis    [From side inputs, collect marked nodes and their unmarked fanins.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Acb_NtkCollectNewTfi1_rec(p *Acb_Ntk_t, iObj int32, vTfiNew *Vec_Int_t) {
    var i int32;
    var iFanin int32;
    var pFanins *int32;
    if !Acb_ObjIsTravIdPrev(p, iObj) {
        return; }
    if Acb_ObjSetTravIdCur(p, iObj) {
        return; }
    Acb_ObjForEachFaninFast( p, iObj, pFanins, iFanin, i )
        Acb_NtkCollectNewTfi1_rec( p, iFanin, vTfiNew );
    Vec_IntPush( vTfiNew, iObj );
}
func Acb_NtkCollectNewTfi2_rec(p *Acb_Ntk_t, iObj int32, vTfiNew *Vec_Int_t) {
    var i int32;
    var iFanin int32;
    var pFanins *int32;
    var fTravIdPrev int32  = Acb_ObjIsTravIdPrev(p, iObj);
    if Acb_ObjSetTravIdCur(p, iObj) {
        return; }
    if (fTravIdPrev != 0) && !Acb_ObjIsCi(p, iObj) {
        Acb_ObjForEachFaninFast( p, iObj, pFanins, iFanin, i )
            Acb_NtkCollectNewTfi2_rec( p, iFanin, vTfiNew ); }
    Vec_IntPush( vTfiNew, iObj );
}
func Acb_NtkCollectNewTfi(p *Acb_Ntk_t, Pivot int32, vDivs *Vec_Int_t, vSide *Vec_Int_t, pnDivs *int32) *Vec_Int_t {
    var vTfiNew *Vec_Int_t   = Vec_IntAlloc( 100 );
    var i, Node int32;
    Acb_NtkIncTravId( p );
    //Acb_NtkPrintVec( p, vDivs, "vDivs" );
    Vec_IntForEachEntry( vDivs, Node, i )
        Acb_NtkCollectNewTfi1_rec( p, Node, vTfiNew );
//Acb_NtkPrintVec( p, vTfiNew, "vTfiNew" );
    Acb_NtkCollectNewTfi1_rec( p, Pivot, vTfiNew );
//Acb_NtkPrintVec( p, vTfiNew, "vTfiNew" );
    assert(Vec_IntEntryLast(vTfiNew) == Pivot, "Vec_IntEntryLast(vTfiNew) == Pivot");
    Vec_IntPop( vTfiNew );
/*
    Vec_IntForEachEntry( vDivs, Node, i )
    {
        Acb_ObjSetTravIdCur( p, Node );
        Vec_IntPush( vTfiNew, Node );
    }
*/
    *pnDivs = Vec_IntSize(vTfiNew);
    Vec_IntForEachEntry( vSide, Node, i )
        Acb_NtkCollectNewTfi2_rec( p, Node, vTfiNew );
    Vec_IntPush( vTfiNew, Pivot );
    return vTfiNew;
}

/**Function*************************************************************

  Synopsis    []

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Acb_NtkCollectWindow(p *Acb_Ntk_t, Pivot int32, vTfi *Vec_Int_t, vTfo *Vec_Int_t, vRoots *Vec_Int_t) *Vec_Int_t {
    var vWin *Vec_Int_t  = Vec_IntAlloc( 100 );
    var i int32;
    var k int32;
    var iObj int32;
    var iFanin int32;
    var pFanins *int32;
    assert(Vec_IntEntryLast(vTfi) == Pivot, "Vec_IntEntryLast(vTfi) == Pivot");
    // mark nodes
    Acb_NtkIncTravId( p );
    Vec_IntForEachEntry( vTfi, iObj, i )
        Acb_ObjSetTravIdCur(p, iObj);
    // add TFI
    Vec_IntForEachEntry( vTfi, iObj, i )
    {
        var fIsTfiInput int32;
        Acb_ObjForEachFaninFast( p, iObj, pFanins, iFanin, k )
            if ( !Acb_ObjIsTravIdCur(p, iFanin) ) // fanin is not in TFI
                fIsTfiInput = 1; // mark as leaf
        Vec_IntPush( vWin, Abc_Var2Lit(iObj, Acb_ObjIsCi(p, iObj) || fIsTfiInput) );
    }
    // mark roots
    Acb_NtkIncTravId( p );
    Vec_IntForEachEntry( vRoots, iObj, i )
        Acb_ObjSetTravIdCur(p, iObj);
    // add TFO
    Vec_IntForEachEntry( vTfo, iObj, i )
    {
        assert(!Acb_ObjIsCo(p, iObj), "!Acb_ObjIsCo(p, iObj)");
        Vec_IntPush( vWin, Abc_Var2Lit(iObj, Acb_ObjIsTravIdCur(p, iObj)) );
    }
    return vWin;
}

/**Function*************************************************************

  Synopsis    []

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Acb_NtkWindow(p *Acb_Ntk_t, Pivot int32, nTfiLevs int32, nTfoLevs int32, nFanMax int32, fDelay int32, pnDivs *int32) *Vec_Int_t {
    var fVerbose int32;
    //int nTfiLevMin = Acb_ObjLevelD(p, Pivot) - nTfiLevs;
    var nTfoLevMax int32  = Acb_ObjLevelD(p, Pivot) + nTfoLevs;
    var vWin, vDivs, vMarked, vTfo, vRoots, vSide, vTfi *Vec_Int_t;
    // collect divisors by traversing limited TFI
    vDivs = Acb_NtkDivisors( p, Pivot, nTfiLevs, fDelay );
    if fVerbose != 0 { Acb_NtkPrintVec( p, vDivs, "vDivs" ); }
    // mark limited TFO of the divisors
    vMarked = Acb_ObjMarkTfo( p, vDivs, Pivot, nTfoLevMax, nFanMax );
    // collect TFO and roots
    Acb_ObjDeriveTfo( p, Pivot, nTfoLevMax, nFanMax, &vTfo, &vRoots, 0 );//fDelay );
    if fVerbose != 0 { Acb_NtkPrintVec( p, vTfo, "vTfo" ); }
    if fVerbose != 0 { Acb_NtkPrintVec( p, vRoots, "vRoots" ); }
    // collect side inputs of the TFO
    vSide = Acb_NtkCollectTfoSideInputs( p, Pivot, vTfo );
    if fVerbose != 0 { Acb_NtkPrintVec( p, vSide, "vSide" ); }
    // mark limited TFO of the divisors
    //Acb_ObjMarkTfo( p, vDivs, Pivot, nTfoLevMax, nFanMax );
    Acb_ObjMarkTfo2( p, vMarked );
    Vec_IntFree( vMarked );
    // collect new TFI
    vTfi = Acb_NtkCollectNewTfi( p, Pivot, vDivs, vSide, pnDivs );
    if fVerbose != 0 { Acb_NtkPrintVec( p, vTfi, "vTfi" ); }
    Vec_IntFree( vSide );
    Vec_IntFree( vDivs );
    // collect all nodes
    vWin = Acb_NtkCollectWindow( p, Pivot, vTfi, vTfo, vRoots );
    // cleanup
    Vec_IntFree( vTfi );
    Vec_IntFree( vTfo );
    Vec_IntFree( vRoots );
    return vWin;
}



/**Function*************************************************************

  Synopsis    []

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Vec_IntVars2Vars(p *Vec_Int_t, Shift int32) {
    var i int32;
    for  i = 0; i < p.nSize; i++  {
        p.pArray[i] += Shift;
    }
}
func Vec_IntVars2Lits(p *Vec_Int_t, Shift int32, fCompl int32) {
    var i int32;
    for  i = 0; i < p.nSize; i++  {
        p.pArray[i] = Abc_Var2Lit( p.pArray[i] + Shift, fCompl );
    }
}
func Vec_IntLits2Vars(p *Vec_Int_t, Shift int32) {
    var i int32;
    for  i = 0; i < p.nSize; i++  {
        p.pArray[i] = Abc_Lit2Var( p.pArray[i] ) + Shift;
    }
}
func Vec_IntRemap(p *Vec_Int_t, vMap *Vec_Int_t) {
    var i int32;
    for  i = 0; i < p.nSize; i++  {
        p.pArray[i] = Vec_IntEntry(vMap, p.pArray[i]);
    }
}

func Acb_WinPrint(p *Acb_Ntk_t, vWin *Vec_Int_t, Pivot int32, nDivs int32) {
    var i, Node int32;
    fmt.Printf("Window for node %d with %d divisors:\n", Pivot, nDivs);
    Vec_IntForEachEntry( vWin, Node, i )
    {
        if i == nDivs {
            fmt.Printf(" | "); }
        if Abc_Lit2Var(Node) == Pivot {
            fmt.Printf("(%d) ", Pivot);
        } else {
            fmt.Printf("%s%d ", Abc_LitIsCompl(Node) ? "*":"", Abc_Lit2Var(Node)); }
    }
    fmt.Printf("\n");
}

func Acb_NtkOrderByRefCount(p *Acb_Ntk_t, vSupp *Vec_Int_t) {
    var i int32;
    var j int32;
    var best_i int32;
    var nSize int32 = Vec_IntSize(vSupp);
    var pArray *int32  = Vec_IntArray(vSupp);
    for  i = 0; i < nSize-1; i++  {
        best_i = i;
        for  j = i+1; j < nSize; j++  {
            if Acb_ObjFanoutNum(p, pArray[j]) > Acb_ObjFanoutNum(p, pArray[best_i]) {
                best_i = j; }
        }
        ABC_SWAP( int, pArray[i], pArray[best_i] );
    }
}

func Acb_NtkRemapIntoSatVariables(p *Acb_Ntk_t, vSupp *Vec_Int_t) {
    var k, iFanin int32;
    Vec_IntForEachEntry( vSupp, iFanin, k )
    {
        assert(Acb_ObjFunc(p, iFanin) >= 0, "Acb_ObjFunc(p, iFanin) >= 0");
        Vec_IntWriteEntry( vSupp, k, Acb_ObjFunc(p, iFanin) );
    }
}

/**Function*************************************************************

  Synopsis    []

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Acb_NtkFindSupp1(p *Acb_Ntk_t, Pivot int32, pSat *sat_solver, nVars int32, nDivs int32, vWin *Vec_Int_t, vSupp *Vec_Int_t) int32 {
    var nSuppNew int32;
    var status int32;
    var k int32;
    var iFanin int32;
    var pFanins *int32;
    Vec_IntClear( vSupp );
    Acb_ObjForEachFaninFast( p, Pivot, pFanins, iFanin, k )
        Vec_IntPush( vSupp, iFanin );
    Acb_NtkOrderByRefCount( p, vSupp );
    Acb_NtkRemapIntoSatVariables( p, vSupp );
    Vec_IntVars2Lits( vSupp, 2*nVars, 0 );
    status = sat_solver_solve( pSat, Vec_IntArray(vSupp), Vec_IntLimit(vSupp), 0, 0, 0, 0 );
    if status != l_False {
        fmt.Printf("Failed internal check at node %d.\n", Pivot); }
    assert(status == l_False, "status == l_False");
    nSuppNew = sat_solver_minimize_assumptions( pSat, Vec_IntArray(vSupp), Vec_IntSize(vSupp), 0 );
    Vec_IntShrink( vSupp, nSuppNew );
    Vec_IntLits2Vars( vSupp, -2*nVars );
    return Vec_IntSize(vSupp) < Acb_ObjFaninNum(p, Pivot);
}

var StrCount int32;

func Acb_NtkFindSupp2(p *Acb_Ntk_t, Pivot int32, pSat *sat_solver, nVars int32, nDivs int32, vWin *Vec_Int_t, vSupp *Vec_Int_t, nLutSize int32, fDelay int32) int32 {
    var nSuppNew int32;
    var status int32;
    var k int32;
    var iFanin int32;
    var pFanins *int32;
    var k2 int32;
    var iFanin2 int32;
    var pFanins2 *int32;
    Acb_ObjForEachFaninFast( p, Pivot, pFanins, iFanin, k )
        assert(Acb_ObjFunc(p, iFanin) >= 0 && Acb_ObjFunc(p, iFanin) < nDivs, "Acb_ObjFunc(p, iFanin) >= 0 && Acb_ObjFunc(p, iFanin) < nDivs");
    if fDelay != 0 {
        // add non-timing-critical fanins
        var nNonCrits int32;
        var k2 int32;
        var iFanin2 int32;
        var pFanins2 *int32;
        assert(Acb_ObjLevelD( p, Pivot ) > 1, "Acb_ObjLevelD( p, Pivot ) > 1");
        Vec_IntClear( vSupp );
        Acb_ObjForEachFaninFast( p, Pivot, pFanins, iFanin, k )
            if ( !Acb_ObjIsDelayCriticalFanin( p, Pivot, iFanin ) )
                Vec_IntPush( vSupp, iFanin );
        nNonCrits = Vec_IntSize(vSupp);
        // add fanins of timing critical fanins
        Acb_ObjForEachFaninFast( p, Pivot, pFanins, iFanin, k )
            if ( Acb_ObjIsDelayCriticalFanin( p, Pivot, iFanin ) )
                Acb_ObjForEachFaninFast( p, iFanin, pFanins2, iFanin2, k2 )
                    Vec_IntPushUnique( vSupp, iFanin2 );
        assert(nNonCrits < Vec_IntSize(vSupp), "nNonCrits < Vec_IntSize(vSupp)");
        // sort additional fanins by level
        Vec_IntSelectSortCost( Vec_IntArray(vSupp) + nNonCrits, Vec_IntSize(vSupp) - nNonCrits, &p.vLevelD );
        // translate to SAT vars
        Vec_IntForEachEntry( vSupp, iFanin, k )
        {
            assert(Acb_ObjFunc(p, iFanin) >= 0, "Acb_ObjFunc(p, iFanin) >= 0");
            Vec_IntWriteEntry( vSupp, k, Acb_ObjFunc(p, iFanin) );
        }
        // solve for these fanins
        Vec_IntVars2Lits( vSupp, 2*nVars, 0 );
        status = sat_solver_solve( pSat, Vec_IntArray(vSupp), Vec_IntLimit(vSupp), 0, 0, 0, 0 );
        if status != l_False {
            fmt.Printf("Failed internal check at node %d.\n", Pivot); }
        assert(status == l_False, "status == l_False");
        nSuppNew = sat_solver_minimize_assumptions( pSat, Vec_IntArray(vSupp), Vec_IntSize(vSupp), 0 );
        Vec_IntShrink( vSupp, nSuppNew );
        Vec_IntLits2Vars( vSupp, -2*nVars );
        return Vec_IntSize(vSupp) <= nLutSize;
    }
    // iterate through different fanout free cones
    Acb_ObjForEachFaninFast( p, Pivot, pFanins, iFanin, k )
    {
        if !(Acb_ObjIsAreaCritical(p, iFanin) != 0) {
            continue; }
        // collect fanins of the root node
        Vec_IntClear( vSupp );
        Acb_ObjForEachFaninFast( p, Pivot, pFanins2, iFanin2, k2 )
            if ( iFanin != iFanin2 )
                Vec_IntPush( vSupp, iFanin2 );
        // collect fanins of the selected node
        Acb_ObjForEachFaninFast( p, iFanin, pFanins2, iFanin2, k2 )
            Vec_IntPushUnique( vSupp, iFanin2 );
        // sort fanins by level
        Vec_IntSelectSortCost( Vec_IntArray(vSupp), Vec_IntSize(vSupp), &p.vLevelD );
        //Acb_NtkOrderByRefCount( p, vSupp );
        Acb_NtkRemapIntoSatVariables( p, vSupp );
        // solve for these fanins
        Vec_IntVars2Lits( vSupp, 2*nVars, 0 );
        status = sat_solver_solve( pSat, Vec_IntArray(vSupp), Vec_IntLimit(vSupp), 0, 0, 0, 0 );
        if status != l_False {
            fmt.Printf("Failed internal check at node %d.\n", Pivot); }
        assert(status == l_False, "status == l_False");
        nSuppNew = sat_solver_minimize_assumptions( pSat, Vec_IntArray(vSupp), Vec_IntSize(vSupp), 0 );
        Vec_IntShrink( vSupp, nSuppNew );
        Vec_IntLits2Vars( vSupp, -2*nVars );
        if Vec_IntSize(vSupp) <= nLutSize {
            return 1; }
    }
    return 0;
}

func Acb_NtkFindSupp3(p *Acb_Ntk_t, Pivot int32, pSat *sat_solver, nVars int32, nDivs int32, vWin *Vec_Int_t, vSupp *Vec_Int_t, nLutSize int32, fDelay int32) int32 {
    var nSuppNew int32;
    var status int32;
    var k int32;
    var iFanin int32;
    var pFanins *int32;
    var k2 int32;
    var iFanin2 int32;
    var pFanins2 *int32;
    var k3 int32;
    var iFanin3 int32;
    var pFanins3 *int32;
    var NodeMark int32;

    if fDelay != 0 {
        return 0; }

    // iterate through pairs of fanins with one fanouts
    Acb_ObjForEachFaninFast( p, Pivot, pFanins, iFanin, k )
    {
        if !(Acb_ObjIsAreaCritical(p, iFanin) != 0) {
            continue; }
        Acb_ObjForEachFaninFast( p, Pivot, pFanins2, iFanin2, k2 )
        {
            if !(Acb_ObjIsAreaCritical(p, iFanin2) != 0) || k2 == k {
                continue; }
            // iFanin and iFanin2 have 1 fanout
            assert(iFanin != iFanin2, "iFanin != iFanin2");

            // collect fanins of the root node
            Vec_IntClear( vSupp );
            Acb_ObjForEachFaninFast( p, Pivot, pFanins3, iFanin3, k3 )
                if iFanin3 != iFanin && iFanin3 != iFanin2 {
                    assert(Acb_ObjFunc(p, iFanin3) >= 0, "Acb_ObjFunc(p, iFanin3) >= 0");
                    Vec_IntPush( vSupp, Abc_Var2Lit(Acb_ObjFunc(p, iFanin3) + 6*nVars, 0) );
                }
            NodeMark = Vec_IntSize(vSupp);

            // collect fanins of the second node
            Acb_ObjForEachFaninFast( p, iFanin, pFanins3, iFanin3, k3 )
            {
                assert(Acb_ObjFunc(p, iFanin3) >= 0, "Acb_ObjFunc(p, iFanin3) >= 0");
                Vec_IntPush( vSupp, Abc_Var2Lit(Acb_ObjFunc(p, iFanin3) + 6*nVars + nDivs, 0) );
            }
            // collect fanins of the third node
            Acb_ObjForEachFaninFast( p, iFanin2, pFanins3, iFanin3, k3 )
            {
                assert(Acb_ObjFunc(p, iFanin3) >= 0, "Acb_ObjFunc(p, iFanin3) >= 0");
                Vec_IntPushUnique( vSupp, Abc_Var2Lit(Acb_ObjFunc(p, iFanin3) + 6*nVars + nDivs, 0) );
            }
            assert(Vec_IntCheckUniqueSmall(vSupp), "Vec_IntCheckUniqueSmall(vSupp)");

            // sort fanins by level
            //Vec_IntSelectSortCost( Vec_IntArray(vSupp) + NodeMark, Vec_IntSize(vSupp) - NodeMark, &p->vLevelD );
            // solve for these fanins
            status = sat_solver_solve( pSat, Vec_IntArray(vSupp), Vec_IntLimit(vSupp), 0, 0, 0, 0 );
            if status != l_False {
                continue; }
            assert(status == l_False, "status == l_False");
            nSuppNew = sat_solver_minimize_assumptions( pSat, Vec_IntArray(vSupp), Vec_IntSize(vSupp), 0 );
            Vec_IntShrink( vSupp, nSuppNew );
            Vec_IntLits2Vars( vSupp, -6*nVars );
            Vec_IntSort( vSupp, 1 );
            // count how many belong to H; the rest belong to G
            NodeMark = 0;
            Vec_IntForEachEntry( vSupp, iFanin3, k3 )
                if ( iFanin3 >= nDivs )
                    Vec_IntWriteEntry( vSupp, k3, iFanin3 - nDivs );
                else 
                    NodeMark++;
            if NodeMark == 0 {
                //printf( "Obj %d: Special case 1 (vars = %d)\n", Pivot, Vec_IntSize(vSupp) );
                continue;
            }
            assert(NodeMark > 0, "NodeMark > 0");
            if Vec_IntSize(vSupp) - NodeMark <= nLutSize {
                return NodeMark; }
        }
    }

    // iterate through fanins with one fanout and their fanins with one fanout
    Acb_ObjForEachFaninFast( p, Pivot, pFanins, iFanin, k )
    {
        if !(Acb_ObjIsAreaCritical(p, iFanin) != 0) {
            continue; }
        Acb_ObjForEachFaninFast( p, iFanin, pFanins2, iFanin2, k2 )
        {
            if !(Acb_ObjIsAreaCritical(p, iFanin2) != 0) {
                continue; }
            // iFanin and iFanin2 have 1 fanout
            assert(iFanin != iFanin2, "iFanin != iFanin2");

            // collect fanins of the root node
            Vec_IntClear( vSupp );
            Acb_ObjForEachFaninFast( p, Pivot, pFanins3, iFanin3, k3 )
                if ( iFanin3 != iFanin && iFanin3 != iFanin2 )
                    Vec_IntPush( vSupp, Abc_Var2Lit(Acb_ObjFunc(p, iFanin3) + 6*nVars, 0) );
            NodeMark = Vec_IntSize(vSupp);

            // collect fanins of the second node
            Acb_ObjForEachFaninFast( p, iFanin, pFanins3, iFanin3, k3 )
                if ( iFanin3 != iFanin2 )
                    Vec_IntPush( vSupp, Abc_Var2Lit(Acb_ObjFunc(p, iFanin3) + 6*nVars + nDivs, 0) );
            // collect fanins of the third node
            Acb_ObjForEachFaninFast( p, iFanin2, pFanins3, iFanin3, k3 )
            {
                assert(Acb_ObjFunc(p, iFanin3) >= 0, "Acb_ObjFunc(p, iFanin3) >= 0");
                Vec_IntPushUnique( vSupp, Abc_Var2Lit(Acb_ObjFunc(p, iFanin3) + 6*nVars + nDivs, 0) );
            }
            assert(Vec_IntCheckUniqueSmall(vSupp), "Vec_IntCheckUniqueSmall(vSupp)");

            // sort fanins by level
            //Vec_IntSelectSortCost( Vec_IntArray(vSupp) + NodeMark, Vec_IntSize(vSupp) - NodeMark, &p->vLevelD );
            //Sat_SolverWriteDimacs( pSat, NULL, Vec_IntArray(vSupp), Vec_IntLimit(vSupp), 0 );
            // solve for these fanins
            status = sat_solver_solve( pSat, Vec_IntArray(vSupp), Vec_IntLimit(vSupp), 0, 0, 0, 0 );
            if status != l_False {
                fmt.Printf("Failed internal check at node %d.\n", Pivot); }
            assert(status == l_False, "status == l_False");
            nSuppNew = sat_solver_minimize_assumptions( pSat, Vec_IntArray(vSupp), Vec_IntSize(vSupp), 0 );
            Vec_IntShrink( vSupp, nSuppNew );
            Vec_IntLits2Vars( vSupp, -6*nVars );
            Vec_IntSort( vSupp, 1 );
            // count how many belong to H; the rest belong to G
            NodeMark = 0;
            Vec_IntForEachEntry( vSupp, iFanin3, k3 )
                if ( iFanin3 >= nDivs )
                    Vec_IntWriteEntry( vSupp, k3, iFanin3 - nDivs );
                else 
                    NodeMark++;
            if NodeMark == 0 {
                //printf( "Obj %d: Special case 2 (vars = %d)\n", Pivot, Vec_IntSize(vSupp) );
                continue;
            }
            assert(NodeMark > 0, "NodeMark > 0");
            if Vec_IntSize(vSupp) - NodeMark <= nLutSize {
                return NodeMark; }
        }
    }

    return 0;
}

/**Function*************************************************************

  Synopsis    []

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
   
type Acb_Mfs_t struct {
    pNtk      *Acb_Ntk_t;        // network
    pPars      *Acb_Par_t;       // parameters
    sat_solver *    pSat[3];     // SAT solvers
    vSupp      *Vec_Int_t;       // support
    vFlip      *Vec_Int_t;       // support
    vValues      *Vec_Int_t;     // support
    nNodes             int32;      // nodes
    nWins             int32;       // windows
    nWinsAll             int32;    // windows
    nDivsAll             int32;    // windows
    nChanges [8]int32; // changes
    nOvers             int32;      // overflows
    nTwoNodes             int32;   // two nodes
    timeTotal         abctime;
    timeCnf         abctime;
    timeSol         abctime;
    timeWin         abctime;
    timeSat         abctime;
    timeSatU         abctime;
    timeSatS         abctime;
}
func Acb_MfsStart(pNtk *Acb_Ntk_t, pPars *Acb_Par_t) *Acb_Mfs_t {
    var p *Acb_Mfs_t  = new(Acb_Mfs_t);
    p.pNtk       = pNtk;
    p.pPars      = pPars;
    p.timeTotal  = Abc_Clock();
    p.pSat[0]    = sat_solver_new();
    p.pSat[1]    = sat_solver_new();
    p.pSat[2]    = sat_solver_new();
    p.vSupp      = Vec_IntAlloc(100);
    p.vFlip      = Vec_IntAlloc(100);
    p.vValues    = Vec_IntAlloc(100);
    return p;
}
func Acb_MfsStop(p *Acb_Mfs_t) {
    Vec_IntFree( p.vFlip );
    Vec_IntFree( p.vSupp );
    Vec_IntFree( p.vValues );
    sat_solver_delete( p.pSat[0] );
    sat_solver_delete( p.pSat[1] );
    sat_solver_delete( p.pSat[2] );
    ABC_FREE( p );
}
func Acb_NtkObjMffcEstimate(pNtk *Acb_Ntk_t, iObj int32) int32 {
    var k int32;
    var iFanin int32;
    var pFanins *int32;
    var Count int32;
    var iFaninCrit int32 = -1;
    Acb_ObjForEachFaninFast( pNtk, iObj, pFanins, iFanin, k )
        if ( Acb_ObjIsAreaCritical(pNtk, iFanin) )
            iFaninCrit = iFanin;
Count++;
    if Count != 1 {
        return Count; }
    Acb_ObjForEachFaninFast( pNtk, iFaninCrit, pFanins, iFanin, k )
        if ( Acb_ObjIsAreaCritical(pNtk, iFanin) )
            Count++;
    return Count;
}

/**Function*************************************************************

  Synopsis    []

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Acb_NtkOptNodeAnalyze(p *Acb_Mfs_t, PivotVar int32, nDivs int32, nValues int32, pValues *int32, vSupp *Vec_Int_t) {
    var OnSet [64]word ;
    var OffSet [64]word ;
    var Diffs [64]word ;
    var s int32;
    var nScope int32 = 1 + 2*nDivs;
    var d int32;
    var i int32;
    var f int32;
    var nFrames int32 = nValues / nScope;
    var start int32;
    if nDivs < 64 {
        start = 0;
    } else {
        start = nDivs - 64;
    }
    var stop int32;
    if nDivs < 64 {
        stop = nDivs;
    } else {
        stop = 64;
    }
    assert(nValues % nScope == 0, "nValues % nScope == 0");
    assert(nFrames <= 16, "nFrames <= 16");
    for  f = 0; f < nFrames; f++  {
        var pStart *int32   = pValues + f * nScope;
        var pOnSet *int32   = pStart + 1 + (pStart[0] ? 0 : nDivs);
        var pOffSet *int32  = pStart + 1 + (pStart[0] ? nDivs : 0);

        fmt.Printf("%2d:", f);
        for  s = start; s < stop; s++  {
            fmt.Printf("%d", pOnSet[s]);
        }
        fmt.Printf("\n");

        fmt.Printf("%2d:", f);
        for  s = start; s < stop; s++  {
            fmt.Printf("%d", pOffSet[s]);
        }
        fmt.Printf("\n");

        for  s = start; s < stop; s++  {
            if pOnSet[s] {   OnSet[f]  |= (((word)1) << (s-start)); }
            if pOffSet[s] {  OffSet[f] |= (((word)1) << (s-start)); }
        }
    }
    d = 0;
    for  f = 0; f < nFrames; f++  {
    for  s = 0; s < nFrames; s++  {
        for  i = 0; i < d; i++  {
            if Diffs[i] == (OnSet[f] ^ OffSet[s]) {
                break; }
        }
        if i < d {
            continue; }
        if d < 64 {
            Diffs[d++] = OnSet[f] ^ OffSet[s]; }
    }
    }

    fmt.Printf("Divisors = %d.  Frames = %d.  Patterns = %d.\n", nDivs, nFrames, d);
    fmt.Printf("   ");
    for  s = start; s < stop; s++  {
        fmt.Printf("%d", s / 10);
    }
    fmt.Printf("\n");
    fmt.Printf("   ");
    for  s = start; s < stop; s++  {
        fmt.Printf("%d", s % 10);
    }
    fmt.Printf("\n");
    fmt.Printf("   ");
    for  s = start; s < stop; s++  {
        fmt.Printf("%c", Vec_IntFind(vSupp, s) >= 0 ? 'a' + Vec_IntFind(vSupp, s) : ' ');
    }
    fmt.Printf("\n");
    for  s = 0; s < d; s++  {
        fmt.Printf("%2d:", s);
        for  f = 0; f < stop; f++  {
            fmt.Printf("%c", ((Diffs[s] >> f) & 1) ? '*' : ' ');
        }
        fmt.Printf("\n");
    }
}

func Acb_NtkOptNode(p *Acb_Mfs_t, Pivot int32) int32 {
    var pCnf *Cnf_Dat_t  = nil; var clk abctime;
    var vWin *Vec_Int_t  = nil; var uTruth word;
    var Result, PivotVar, nDivs, RetValue, c int32;
    assert(Acb_ObjFanoutNum(p.pNtk, Pivot) > 0, "Acb_ObjFanoutNum(p.pNtk, Pivot) > 0");
    p.nWins++;

    // compute divisors and window for this target node with these taboo nodes
    clk = Abc_Clock();
    vWin = Acb_NtkWindow( p.pNtk, Pivot, p.pPars.nTfiLevMax, p.pPars.nTfoLevMax, p.pPars.nFanoutMax, !p.pPars.fArea, &nDivs );
    p.nWinsAll += Vec_IntSize(vWin);
    p.nDivsAll += nDivs;
    p.timeWin  += Abc_Clock() - clk;
    PivotVar = Vec_IntFind( vWin, Abc_Var2Lit(Pivot, 0) );
    if p.pPars.fVerbose {
    fmt.Printf("Node %d: Window contains %d objects and %d divisors.  ", Pivot, Vec_IntSize(vWin), nDivs); }
//    Acb_WinPrint( p->pNtk, vWin, Pivot, nDivs );
//    Acb_NtkPrintVecWin( p->pNtk, vWin, "Win" );
    if Vec_IntSize(vWin) > p.pPars.nWinNodeMax {
        p.nOvers++;
        if p.pPars.fVerbose {
            fmt.Printf("Too many divisors.\n"); }
        goto cleanup;
    }

    // derive CNF 
    clk = Abc_Clock();
    pCnf = Acb_NtkWindow2Cnf( p.pNtk, vWin, Pivot );
    assert(PivotVar == Acb_ObjFunc(p.pNtk, Pivot), "PivotVar == Acb_ObjFunc(p.pNtk, Pivot)");
    Cnf_DataCollectFlipLits( pCnf, PivotVar, p.vFlip );
    p.timeCnf += Abc_Clock() - clk;

    // derive SAT solver
    clk = Abc_Clock();
    Acb_NtkWindow2Solver( p.pSat[0], pCnf, p.vFlip, PivotVar, nDivs, 1 );
    p.timeSol += Abc_Clock() - clk;
    // check constants
    for  c = 0; c < 2; c++  {
        var Lit int32  = Abc_Var2Lit( PivotVar, c );
        var status int32  = sat_solver_solve( p.pSat[0], &Lit, &Lit + 1, 0, 0, 0, 0 );
        if status == l_False {
            p.nChanges[0]++;
            if p.pPars.fVerbose {
            fmt.Printf("Found constant %d.\n", c); }
            Acb_NtkUpdateNode( p.pNtk, Pivot, c ? ~(word)0 : 0, nil );
            RetValue = 1;
            goto cleanup;
        }
        assert(status == l_True, "status == l_True");
    }

    // derive SAT solver
    clk = Abc_Clock();
    Acb_NtkWindow2Solver( p.pSat[1], pCnf, p.vFlip, PivotVar, nDivs, 2 );
    p.timeSol += Abc_Clock() - clk;

    // try to remove useless fanins
    if p.pPars.fArea {
        var fEnableProfile int32;
        if fEnableProfile != 0 {
            // alloc
            if p.pSat[1].user_values.cap == 0 {
                veci_new(&p.pSat[1].user_values);
            } else {
                p.pSat[1].user_values.size = 0; }
            if p.pSat[1].user_vars.cap == 0 {
                veci_new(&p.pSat[1].user_vars);
            } else {
                p.pSat[1].user_vars.size = 0; }
            // set variables
            veci_push(&p.pSat[1].user_vars, PivotVar);
            for  c = 0; c < nDivs; c++  {
                veci_push(&p.pSat[1].user_vars, c);
            }
            for  c = 0; c < nDivs; c++  {
                veci_push(&p.pSat[1].user_vars, c+pCnf.nVars);
            }
        }

        // perform solving
        clk = Abc_Clock();
        Result = Acb_NtkFindSupp1( p.pNtk, Pivot, p.pSat[1], pCnf.nVars, nDivs, vWin, p.vSupp );
        p.timeSat += Abc_Clock() - clk;
        // undo variables
        p.pSat[1].user_vars.size = 0;
        if Result {
            if Vec_IntSize(p.vSupp) == 0 {
                p.nChanges[0]++;
            } else {
                p.nChanges[1]++; }
            assert(Vec_IntSize(p.vSupp) < p.pPars.nLutSize, "Vec_IntSize(p.vSupp) < p.pPars.nLutSize");
            if p.pPars.fVerbose {
            fmt.Printf("Found %d inputs: ", Vec_IntSize(p.vSupp)); }
            uTruth = Acb_ComputeFunction( p.pSat[0], PivotVar, sat_solver_nvars(p.pSat[0])-1, p.vSupp, 0 );
            if p.pPars.fVerbose {
            Extra_PrintHex( os.Stdout, (unsigned *)&uTruth, Vec_IntSize(p.vSupp) ); } 
            if p.pPars.fVerbose {
            fmt.Printf("\n"); }
            // create support in terms of nodes
            Vec_IntRemap( p.vSupp, vWin );
            Vec_IntLits2Vars( p.vSupp, 0 );
            Acb_NtkUpdateNode( p.pNtk, Pivot, uTruth, p.vSupp );
            RetValue = 1;
            goto cleanup;
        }
        if fEnableProfile != 0 {
            // analyze the resulting values
            Acb_NtkOptNodeAnalyze( p, PivotVar, nDivs, p.pSat[1].user_values.size, p.pSat[1].user_values.ptr, p.vSupp );
            p.pSat[1].user_values.size = 0;
        }
    }

    if Acb_NtkObjMffcEstimate(p.pNtk, Pivot) >= 1 {
        // check for one-node implementation
        clk = Abc_Clock();
        Result = Acb_NtkFindSupp2( p.pNtk, Pivot, p.pSat[1], pCnf.nVars, nDivs, vWin, p.vSupp, p.pPars.nLutSize, !p.pPars.fArea );
        p.timeSat += Abc_Clock() - clk;
        if Result {
            p.nChanges[2]++;
            assert(Vec_IntSize(p.vSupp) <= p.pPars.nLutSize, "Vec_IntSize(p.vSupp) <= p.pPars.nLutSize");
            if p.pPars.fVerbose {
            fmt.Printf("Found %d inputs: ", Vec_IntSize(p.vSupp)); }
            uTruth = Acb_ComputeFunction( p.pSat[0], PivotVar, sat_solver_nvars(p.pSat[0])-1, p.vSupp, 0 );
            if p.pPars.fVerbose {
            Extra_PrintHex( os.Stdout, (unsigned *)&uTruth, Vec_IntSize(p.vSupp) ); } 
            if p.pPars.fVerbose {
            fmt.Printf("\n"); }
            // create support in terms of nodes
            Vec_IntRemap( p.vSupp, vWin );
            Vec_IntLits2Vars( p.vSupp, 0 );
            Acb_NtkUpdateNode( p.pNtk, Pivot, uTruth, p.vSupp );
            RetValue = 1;
            goto cleanup;
        }
    }

//#if 0
    if p.pPars.fUseAshen && Acb_NtkObjMffcEstimate(p.pNtk, Pivot) >= 2 {// && Pivot != 70 )
    
        p.nTwoNodes++;
        // derive SAT solver
        clk = Abc_Clock();
        Acb_NtkWindow2Solver( p.pSat[2], pCnf, p.vFlip, PivotVar, nDivs, 6 );
        p.timeSol += Abc_Clock() - clk;

        // check for two-node implementation
        clk = Abc_Clock();
        Result = Acb_NtkFindSupp3( p.pNtk, Pivot, p.pSat[2], pCnf.nVars, nDivs, vWin, p.vSupp, p.pPars.nLutSize, !p.pPars.fArea );
        p.timeSat += Abc_Clock() - clk;
        if Result {
            var fVerbose int32  = 1;
            var i, k, Lit, Var, Var2, status, NodeNew, fBecameUnsat, fCompl int32;
            assert(Result                       <  p.pPars.nLutSize, "Result                       <  p.pPars.nLutSize");
            assert(Vec_IntSize(p.vSupp)-Result <= p.pPars.nLutSize, "Vec_IntSize(p.vSupp)-Result <= p.pPars.nLutSize");
            if (fVerbose != 0) || p.pPars.fVerbose {
            fmt.Printf("Obj %5d: Found %d Hvars and %d Gvars: ", Pivot, Result, Vec_IntSize(p.vSupp)-Result); }
            // p->vSupp contains G variables (Vec_IntSize(p->vSupp)-Result) followed by H variables (Result)
            //sat_solver_restart( p->pSat[1] );
            //Acb_NtkWindow2Solver( p->pSat[1], pCnf, p->vFlip, PivotVar, nDivs, 2 );

            // constrain H-variables to be equal
            Vec_IntForEachEntryStart( p.vSupp, Var, i, Vec_IntSize(p.vSupp)-Result ) // H variables
            {
                assert(Var >= 0 && Var < nDivs, "Var >= 0 && Var < nDivs");
                assert(Var + 2*pCnf.nVars < sat_solver_nvars(p.pSat[1]), "Var + 2*pCnf.nVars < sat_solver_nvars(p.pSat[1])");
                Lit = Abc_Var2Lit( Var + 2*pCnf.nVars, 0 ); // HVars are the same
                if !sat_solver_addclause( p.pSat[1], &Lit, &Lit + 1 ) { if (fVerbose != 0) || p.pPars.fVerbose { fmt.Printf("Error: SAT solver became UNSAT at a wrong place (place 2).  "); }
 fBecameUnsat = 1; }
            }
            // find one satisfying assighment
            status = sat_solver_solve( p.pSat[1], nil, nil, 0, 0, 0, 0 );
            assert(status == l_True, "status == l_True");
            // get assignment of the function
            fCompl = !sat_solver_var_value( p.pSat[1], PivotVar );
            // constrain second set of G-vars to have values equal to the assignment
            Vec_IntForEachEntryStop( p.vSupp, Var, i, Vec_IntSize(p.vSupp)-Result ) // G variables
            {
                // check if this is a C-var
                Vec_IntForEachEntryStart( p.vSupp, Var2, k, Vec_IntSize(p.vSupp)-Result ) // G variables
                    if ( Var == Var2 )
                        break;
                if k < Vec_IntSize(p.vSupp) { // do not constrain a C-var
                
                    if (fVerbose != 0) || p.pPars.fVerbose {
                    fmt.Printf("Found C-var in object %d.  ", Pivot); }
                    continue;
                }
                assert(Var >= 0 && Var < nDivs, "Var >= 0 && Var < nDivs");
                Lit = sat_solver_var_literal( p.pSat[1], Var + pCnf.nVars );
                if !sat_solver_addclause( p.pSat[1], &Lit, &Lit + 1 ) { if (fVerbose != 0) || p.pPars.fVerbose { fmt.Printf("Error: SAT solver became UNSAT at a wrong place (place 1).  "); }
 fBecameUnsat = 1; }
            }
            if fBecameUnsat {
                StrCount++;
                if (fVerbose != 0) || p.pPars.fVerbose {
                fmt.Printf(" Quitting.\n"); }
                goto cleanup;
            }
            // consider only G variables
            p.vSupp.nSize -= Result;
            // truth table
            uTruth = Acb_ComputeFunction( p.pSat[1], PivotVar, sat_solver_nvars(p.pSat[1])-1, p.vSupp, fCompl );
            if (fVerbose != 0) || p.pPars.fVerbose {
            Extra_PrintHex( os.Stdout, (unsigned *)&uTruth, Vec_IntSize(p.vSupp) ); } 
            if uTruth == 0 || ~uTruth == 0 {
                if (fVerbose != 0) || p.pPars.fVerbose {
                fmt.Printf(" Quitting.\n"); }
                goto cleanup;
            }
            p.nChanges[3]++;
            // create new node
            Vec_IntRemap( p.vSupp, vWin );
            Vec_IntLits2Vars( p.vSupp, 0 );
            NodeNew = Acb_NtkCreateNode( p.pNtk, uTruth, p.vSupp );
            Acb_DeriveCnfForWindowOne( p.pNtk, NodeNew );
            Acb_DeriveCnfForNode( p.pNtk, NodeNew, p.pSat[0], sat_solver_nvars(p.pSat[0])-2 );
            p.vSupp.nSize += Result;
            // collect new variables
            Vec_IntForEachEntryStart( p.vSupp, Var, i, Vec_IntSize(p.vSupp)-Result )
                Vec_IntWriteEntry( p.vSupp, i-(Vec_IntSize(p.vSupp)-Result), Var );
            Vec_IntShrink( p.vSupp, Result );
            Vec_IntPush( p.vSupp, sat_solver_nvars(p.pSat[0])-2 );
            // truth table
            uTruth = Acb_ComputeFunction( p.pSat[0], PivotVar, sat_solver_nvars(p.pSat[0])-1, p.vSupp, 0 );
            // create new fanins of the node
            if (fVerbose != 0) || p.pPars.fVerbose {
            fmt.Printf("  "); }
            if (fVerbose != 0) || p.pPars.fVerbose {
            Extra_PrintHex( os.Stdout, (unsigned *)&uTruth, Vec_IntSize(p.vSupp) ); } 
            if (fVerbose != 0) || p.pPars.fVerbose {
            fmt.Printf("\n"); }
            // create support in terms of nodes
            Vec_IntPop( p.vSupp );
            Vec_IntRemap( p.vSupp, vWin );
            Vec_IntLits2Vars( p.vSupp, 0 );
            Vec_IntPush( p.vSupp, NodeNew );
            Acb_NtkUpdateNode( p.pNtk, Pivot, uTruth, p.vSupp );
            RetValue = 2;
            goto cleanup;
        }
    }
//#endif

    if p.pPars.fVerbose {
    fmt.Printf("\n"); }

cleanup:
    sat_solver_restart( p.pSat[0] );
    sat_solver_restart( p.pSat[1] );
    sat_solver_restart( p.pSat[2] );
    if pCnf != nil {
        Cnf_DataFree( pCnf );
        Acb_NtkWindowUndo( p.pNtk, vWin );
    }
    Vec_IntFreeP( &vWin );
    return RetValue;
}

/**Function*************************************************************

  Synopsis    []

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Acb_NtkOpt(pNtk *Acb_Ntk_t, pPars *Acb_Par_t) {
    var pMan *Acb_Mfs_t  = Acb_MfsStart( pNtk, pPars );
    if pPars.fVerbose {
        fmt.Printf("%s-optimization parameters: TfiLev(I) = %d  TfoLev(O) = %d  WinMax(W) = %d  LutSize = %d\n", pMan.pPars.fArea ? "Area" : "Delay", pMan.pPars.nTfiLevMax, pMan.pPars.nTfoLevMax, pMan.pPars.nWinNodeMax, pMan.pPars.nLutSize); }
    Acb_NtkCreateFanout( pNtk );  // fanout data structure
    Acb_NtkCleanObjFuncs( pNtk ); // SAT variables
    Acb_NtkCleanObjCnfs( pNtk );  // CNF representations
    if pMan.pPars.fArea {
        var n int32;
        var iObj int32;
        var RetValue int32;
        var nNodes int32 = Acb_NtkObjNumMax(pNtk);
        var vVisited *Vec_Bit_t  = Vec_BitStart( Acb_NtkObjNumMax(pNtk) );
        Acb_NtkUpdateLevelD( pNtk, -1 ); // compute forward logic level
        for ( n = 2; n >= 0; n-- )
            Acb_NtkForEachNode( pNtk, iObj )
                if iObj < nNodes && !Vec_BitEntry(vVisited, iObj) && Acb_NtkObjMffcEstimate(pNtk, iObj) >= n {
                    pMan.nNodes++;
                    //if ( iObj != 103 )
                    //    continue;
                    //Acb_NtkOptNode( pMan, iObj );
                    for  (RetValue = Acb_NtkOptNode(pMan, iObj)) && Acb_ObjFaninNum(pNtk, iObj)  {; }                    
                    Vec_BitWriteEntry( vVisited, iObj, 1 );
                }
        Vec_BitFree( vVisited );
    
    } else {
        var Value int32;
        Acb_NtkUpdateTiming( pNtk, -1 ); // compute delay information
        for  (Value = int32(Vec_QueTopPriority(pNtk.vQue))) > 0  {
            var iObj int32  = Vec_QuePop(pNtk.vQue);
            if !Acb_ObjType(pNtk, iObj) {
                continue; }
            //if ( iObj != 103 )
            //    continue;
            //printf( "Trying node %4d (%4d) ", iObj, Value );
            Acb_NtkOptNode( pMan, iObj ); 
        }
    }
    if pPars.fVerbose {
        pMan.timeTotal = Abc_Clock() - pMan.timeTotal;
        fmt.Printf("Node = %d  Win = %d (Ave = %d)  DivAve = %d   Change = %d  C = %d  N1 = %d  N2 = %d  N3 = %d   Over = %d  Str = %d  2Node = %d.\n", pMan.nNodes, pMan.nWins, pMan.nWinsAll/Abc_MaxInt(1, pMan.nWins), pMan.nDivsAll/Abc_MaxInt(1, pMan.nWins), pMan.nChanges[0] + pMan.nChanges[1] + pMan.nChanges[2] + pMan.nChanges[3], pMan.nChanges[0], pMan.nChanges[1], pMan.nChanges[2], pMan.nChanges[3], pMan.nOvers, StrCount, pMan.nTwoNodes);
        ABC_PRTP( "Windowing  ", pMan.timeWin,    pMan.timeTotal );
        ABC_PRTP( "CNF compute", pMan.timeCnf,    pMan.timeTotal );
        ABC_PRTP( "Make solver", pMan.timeSol,    pMan.timeTotal );
        ABC_PRTP( "SAT solving", pMan.timeSat,    pMan.timeTotal );
//        ABC_PRTP( "  unsat    ", pMan->timeSatU,   pMan->timeTotal );
//        ABC_PRTP( "  sat      ", pMan->timeSatS,   pMan->timeTotal );
        ABC_PRTP( "TOTAL      ", pMan.timeTotal,  pMan.timeTotal );
        fflush( os.Stdout );
    }
    Acb_MfsStop( pMan );
    StrCount = 0;
}

////////////////////////////////////////////////////////////////////////
///                       END OF FILE                                ///
////////////////////////////////////////////////////////////////////////


//ABC_NAMESPACE_IMPL_END

