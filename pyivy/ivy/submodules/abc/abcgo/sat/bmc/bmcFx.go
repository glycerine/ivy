package bmc

import "fmt"

/**CFile****************************************************************

  FileName    [bmcFx.c]

  SystemName  [ABC: Logic synthesis and verification system.]

  PackageName [SAT-based bounded model checking.]

  Synopsis    [INT-FX: Interpolation-based logic sharing extraction.]

  Author      [Alan Mishchenko]
  
  Affiliation [UC Berkeley]

  Date        [Ver. 1.0. Started - June 20, 2005.]

  Revision    [$Id: bmcFx.c,v 1.00 2005/06/20 00:00:00 alanmi Exp $]

***********************************************************************/

//#include "bmc.h"
//#include "misc/vec/vecWec.h"
//#include "sat/cnf/cnf.h"
//#include "sat/bsat/satStore.h"

//ABC_NAMESPACE_IMPL_START


////////////////////////////////////////////////////////////////////////
///                        DECLARATIONS                              ///
////////////////////////////////////////////////////////////////////////

////////////////////////////////////////////////////////////////////////
///                     FUNCTION DEFINITIONS                         ///
////////////////////////////////////////////////////////////////////////

/**Function*************************************************************

  Synopsis    [Create hash table to hash divisors.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
const TAB_UNUSED = 0x7FFF
    // 16 bytes
type Tab_Obj_t struct {
    Table         int32;
    Next         int32;
    Cost uint32; /* C bitfield: 17 bits; review packing and truncation */
    LitA uint32; /* C bitfield: 15 bits; review packing and truncation */
    LitB uint32; /* C bitfield: 15 bits; review packing and truncation */
    LitC uint32; /* C bitfield: 15 bits; review packing and truncation */
    Func uint32; /* C bitfield: 2 bits; review packing and truncation */
}
    // 16 bytes
type Tab_Tab_t struct {
    SizeMask         int32;
    nBins         int32;
    pBins  *Tab_Obj_t;
}
func Tab_TabAlloc(LogSize int32) *Tab_Tab_t {
    // inlining of macro ABC_CALLOC
    var p *Tab_Tab_t  = &Tab_Tab_t{};
    assert(LogSize >= 4 && LogSize <= 31, "LogSize >= 4 && LogSize <= 31");
    p.SizeMask = (1 << LogSize) - 1;
    // inlining of macro ABC_CALLOC
    p.pBins = ((Tab_Obj_t *) calloc((size_t)(p.SizeMask + 1), sizeof(Tab_Obj_t)));
    p.nBins = 1;
    return p;
}
func Tab_TabFree(p *Tab_Tab_t) {
    p.pBins = nil;
    p = nil;
}
func Tab_TabFindBest(p *Tab_Tab_t, nDivs int32) *Vec_Int_t {
    var pNames [5]string = [5]string{"const1", "and", "xor", "mux", "none"};
    var pOrder *int32;
    var i int32;
    var vDivs *Vec_Int_t  = Vec_IntAlloc( 100 ); 
    var vCosts *Vec_Int_t  = Vec_IntAlloc( p.nBins ); 
    var pEnt *Tab_Obj_t;
    var pLimit *Tab_Obj_t = p.pBins + p.nBins;
    for  pEnt = p.pBins; pEnt < pLimit; pEnt++  {
        Vec_IntPush( vCosts, -int32(pEnt.Cost) );
    }
    pOrder = Abc_MergeSortCost( Vec_IntArray(vCosts), Vec_IntSize(vCosts) );
    for  i = 0; i < Vec_IntSize(vCosts); i++  {
        pEnt = p.pBins + pOrder[i];
        if i == nDivs || pEnt.Cost == 1 {
            break; }
        fmt.Printf("Lit0 = %5d.  Lit1 = %5d.  Lit2 = %5d.  Func = %s.  Cost = %3d.\n", pEnt.LitA, pEnt.LitB, pEnt.LitC, pNames[pEnt.Func], pEnt.Cost);
        Vec_IntPushTwo( vDivs, pEnt.Func, pEnt.LitA );
        Vec_IntPushTwo( vDivs, pEnt.LitB, pEnt.LitC );
    }
    Vec_IntFree( vCosts );
    pOrder = nil;
    return vDivs;
}
func Tab_Hash(LitA int32, LitB int32, LitC int32, Func int32, Mask int32) int32 {
    return (LitA * 50331653 + LitB * 100663319 + LitC * 201326611 + Func * 402653189) & Mask;
}
func Tab_TabRehash(p *Tab_Tab_t) {
    var pEnt, pLimit, pBin *Tab_Obj_t;
    assert(p.nBins == p.SizeMask + 1, "p.nBins == p.SizeMask + 1");
    // realloc memory
    // inlining of macro ABC_REALLOC
    
    if p.pBins {
        p.pBins = ((Tab_Obj_t *) realloc((char *)(p.pBins), sizeof(Tab_Obj_t) * (size_t)(2 * (p.SizeMask + 1))));
    } else {
        p.pBins = ((Tab_Obj_t *) malloc(sizeof(Tab_Obj_t) * (size_t)(2 * (p.SizeMask + 1))));
    }
    memset( p.pBins + p.SizeMask + 1, 0, sizeof(Tab_Obj_t) * (p.SizeMask + 1) );
    // clean entries
    pLimit = p.pBins + p.SizeMask + 1;
    for  pEnt = p.pBins; pEnt < pLimit; pEnt++  {
        pEnt.Table = pEnt.Next = 0;
    }
    // rehash entries
    p.SizeMask = 2 * p.SizeMask + 1;
    for  pEnt = p.pBins + 1; pEnt < pLimit; pEnt++  {
        pBin = p.pBins + Tab_Hash( pEnt.LitA, pEnt.LitB, pEnt.LitC, pEnt.Func, p.SizeMask );
        pEnt.Next  = pBin.Table;
        pBin.Table = pEnt - p.pBins;
        assert(!pEnt.Next || pEnt.Next != pBin.Table, "!pEnt.Next || pEnt.Next != pBin.Table");
    }
}
func Tab_TabEntry(p *Tab_Tab_t, i int32) *Tab_Obj_t { if i {
    return p.pBins + i;
} else {
    return nil;
} }
func Tab_TabHashAdd(p *Tab_Tab_t, pLits *int32, Func int32, Cost int32) int32 {
    if p.nBins == p.SizeMask + 1 {
        Tab_TabRehash( p ); }
    assert(p.nBins < p.SizeMask + 1, "p.nBins < p.SizeMask + 1");
    {
        var pEnt *Tab_Obj_t;
        var pBin *Tab_Obj_t = p.pBins + Tab_Hash(pLits[0], pLits[1], pLits[2], Func, p.SizeMask);
        for  pEnt = Tab_TabEntry(p, pBin.Table); (pEnt != nil); pEnt = Tab_TabEntry(p, pEnt.Next)  {
            if int32(pEnt.LitA) == pLits[0] && int32(pEnt.LitB) == pLits[1] && int32(pEnt.LitC) == pLits[2] && int32(pEnt.Func) == Func {  pEnt.Cost += Cost; return 1; }
        }
        pEnt = p.pBins + p.nBins;
        pEnt.LitA  = pLits[0];
        pEnt.LitB  = pLits[1];
        pEnt.LitC  = pLits[2];
        pEnt.Func  = Func;
        pEnt.Cost  = Cost;
        pEnt.Next  = pBin.Table;
        pBin.Table = p.nBins++;
        assert(!pEnt.Next || pEnt.Next != pBin.Table, "!pEnt.Next || pEnt.Next != pBin.Table");
        return 0;
    }
}


/**Function*************************************************************

  Synopsis    [Input is four literals. Output is type, polarity and fanins.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
// name types
 type Div_Type_t int32
const ( 
    DIV_CST Div_Type_t = 0
   // 0: constant 1
    DIV_AND Div_Type_t = DIV_CST + 1
       // 1: and (ordered fanins)
    DIV_XOR Div_Type_t = DIV_AND + 1
       // 2: xor (ordered fanins)
    DIV_MUX Div_Type_t = DIV_XOR + 1
       // 3: mux (c, d1, d0)
    DIV_NONE Div_Type_t = DIV_MUX + 1       // 4: not used

)

  

func Bmc_FxDivOr(LitA int32, LitB int32, pLits *int32, pPhase *int32) Div_Type_t {
    assert(LitA != LitB, "LitA != LitB");
    if Abc_Lit2Var(LitA) == Abc_Lit2Var(LitB) {
        return DIV_CST; }
    if LitA > LitB {
        // inlining of macro ABC_SWAP
         var t int32  = LitA; LitA = LitB; LitB = t; };
    pLits[0] = Abc_LitNot(LitA);
    pLits[1] = Abc_LitNot(LitB);
    *pPhase = 1;
    return DIV_AND;
}
func Bmc_FxDivXor(LitA int32, LitB int32, pLits *int32, pPhase *int32) Div_Type_t {
    assert(LitA != LitB, "LitA != LitB");
    *pPhase ^= Abc_LitIsCompl(LitA);
    *pPhase ^= Abc_LitIsCompl(LitB);
    pLits[0] = Abc_LitRegular(LitA);
    pLits[1] = Abc_LitRegular(LitB);
    return DIV_XOR;
}
func Bmc_FxDivMux(LitC int32, LitCn int32, LitT int32, LitE int32, pLits *int32, pPhase *int32) Div_Type_t {
    assert(LitC != LitCn, "LitC != LitCn");
    assert(Abc_Lit2Var(LitC) == Abc_Lit2Var(LitCn), "Abc_Lit2Var(LitC) == Abc_Lit2Var(LitCn)");
    assert(Abc_Lit2Var(LitC) != Abc_Lit2Var(LitT), "Abc_Lit2Var(LitC) != Abc_Lit2Var(LitT)");
    assert(Abc_Lit2Var(LitC) != Abc_Lit2Var(LitE), "Abc_Lit2Var(LitC) != Abc_Lit2Var(LitE)");
    assert(Abc_Lit2Var(LitC) != Abc_Lit2Var(LitE), "Abc_Lit2Var(LitC) != Abc_Lit2Var(LitE)");
    if Abc_LitIsCompl(LitC) {
        LitC = Abc_LitRegular(LitC);
        // inlining of macro ABC_SWAP
        { var t int32  = LitT; LitT = LitE; LitE = t; };
    }
    if Abc_LitIsCompl(LitT) {
        *pPhase ^= 1;
        LitT = Abc_LitNot(LitT);
        LitE = Abc_LitNot(LitE);
    }
    pLits[0] = LitC;
    pLits[1] = LitT;
    pLits[2] = LitE;
    return DIV_MUX;
}
func Div_FindType(LitA []int32, LitB []int32, pLits *int32, pPhase *int32) Div_Type_t {
    *pPhase = 0;
    pLits[0] = pLits[1] = pLits[2] = TAB_UNUSED;
    if LitA[0] == -1 && LitA[1] == -1 { return DIV_CST; }
    if LitB[0] == -1 && LitB[1] == -1 { return DIV_CST; }
    assert(LitA[0] >= 0 && LitB[0] >= 0, "LitA[0] >= 0 && LitB[0] >= 0");
    assert(LitA[0] != LitB[0], "LitA[0] != LitB[0]");
    if LitA[1] == -1 && LitB[1] == -1 {
        return Bmc_FxDivOr( LitA[0], LitB[0], pLits, pPhase ); }
    assert(LitA[1] != LitB[1], "LitA[1] != LitB[1]");
    if LitA[1] == -1 || LitB[1] == -1 {
        if LitA[1] == -1 {
            // inlining of macro ABC_SWAP
            { var t int32  = LitA[0]; LitA[0] = LitB[0]; LitB[0] = t; };
            // inlining of macro ABC_SWAP
            { var t int32  = LitA[1]; LitA[1] = LitB[1]; LitB[1] = t; };
        }
        assert(LitA[0] >= 0 && LitA[1] >= 0, "LitA[0] >= 0 && LitA[1] >= 0");
        assert(LitB[0] >= 0 && LitB[1] == -1, "LitB[0] >= 0 && LitB[1] == -1");
        if Abc_Lit2Var(LitB[0]) == Abc_Lit2Var(LitA[0]) {
            return Bmc_FxDivOr( LitB[0], LitA[1], pLits, pPhase ); }
        if Abc_Lit2Var(LitB[0]) == Abc_Lit2Var(LitA[1]) {
            return Bmc_FxDivOr( LitB[0], LitA[0], pLits, pPhase ); }
        return DIV_NONE;
    }
    if Abc_Lit2Var(LitA[0]) == Abc_Lit2Var(LitB[0]) && Abc_Lit2Var(LitA[1]) == Abc_Lit2Var(LitB[1]) {
        return Bmc_FxDivXor( LitA[0], LitB[1], pLits, pPhase ); }
    if Abc_Lit2Var(LitA[0]) == Abc_Lit2Var(LitB[0]) {
        return Bmc_FxDivMux( LitA[0], LitB[0], LitA[1], LitB[1], pLits, pPhase ); }
    if Abc_Lit2Var(LitA[0]) == Abc_Lit2Var(LitB[1]) {
        return Bmc_FxDivMux( LitA[0], LitB[1], LitA[1], LitB[0], pLits, pPhase ); }
    if Abc_Lit2Var(LitA[1]) == Abc_Lit2Var(LitB[0]) {
        return Bmc_FxDivMux( LitA[1], LitB[0], LitA[0], LitB[1], pLits, pPhase ); }
    if Abc_Lit2Var(LitA[1]) == Abc_Lit2Var(LitB[1]) {
        return Bmc_FxDivMux( LitA[1], LitB[1], LitA[0], LitB[0], pLits, pPhase ); }
    return DIV_NONE;
}

/**Function*************************************************************

  Synopsis    [Returns the number of shared variables, or -1 if failed.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Div_AddLit(Lit int32, pLits []int32) int32 {
    if pLits[0] == -1 {
        pLits[0] = Lit;
    } else if pLits[1] == -1 {
        pLits[1] = Lit;
    } else { return 1; }
    return 0;
}
func Div_FindDiv(vA *Vec_Int_t, vB *Vec_Int_t, pLitsA []int32, pLitsB []int32) int32 {
    var Counter int32;
    var pBegA *int32 = vA.pArray;
    var pEndA *int32 = vA.pArray + vA.nSize;
    var pBegB *int32 = vB.pArray;
    var pEndB *int32 = vB.pArray + vB.nSize; 
    pLitsA[0] = pLitsA[1] = pLitsB[0] = pLitsB[1] = -1;
    for  pBegA < pEndA && pBegB < pEndB  {
        if *pBegA == *pBegB {
            pBegA++, pBegB++, Counter++;
        } else if *pBegA < *pBegB {
            if Div_AddLit( *pBegA++, pLitsA ) {
                return -1; }
        
        } else {
            if Div_AddLit( *pBegB++, pLitsB ) {
                return -1; }
        }
    }
    for  pBegA < pEndA  {
        if Div_AddLit( *pBegA++, pLitsA ) {
            return -1; }
    }
    for  pBegB < pEndB  {
        if Div_AddLit( *pBegB++, pLitsB ) {
            return -1; }
    }
    return Counter;
}

func Div_CubePrintOne(vCube *Vec_Int_t, vStr *Vec_Str_t, nVars int32) {
    var i, Lit int32;
    Vec_StrFill( vStr, nVars, '-' );
    // inlining of macro Vec_IntForEachEntry
    for  i = 0; (i < Vec_IntSize(vCube)) && (((Lit) = Vec_IntEntry(vCube, i)), 1); i++  {
        Vec_StrWriteEntry( vStr, Abc_Lit2Var(Lit), byte((Abc_LitIsCompl(Lit) ? '0' : '1')) );
    }
    fmt.Printf("%s\n", Vec_StrArray(vStr));
}
func Div_CubePrint(p *Vec_Wec_t, nVars int32) {
    var vCube *Vec_Int_t; var i int32;
    var vStr *Vec_Str_t  = Vec_StrStart( nVars + 1 );
    // inlining of macro Vec_WecForEachLevel
    for  i = 0; (i < Vec_WecSize(p)) && (((vCube) = Vec_WecEntry(p, i)), 1); i++  {
        Div_CubePrintOne( vCube, vStr, nVars );
    }
    Vec_StrFree( vStr );
}

func Div_CubePairs(p *Vec_Wec_t, nVars int32, nDivs int32) *Vec_Int_t {
    var fVerbose int32;
    var pNames [5]string = [5]string{"const1", "and", "xor", "mux", "none"};
    var vCube1, vCube2, vDivs *Vec_Int_t; 
    var i1 int32;
    var i2 int32;
    var i int32;
    var k int32;
    var pLitsA [2]int32;
    var pLitsB [2]int32;
    var pLits [4]int32;
    var Type int32;
    var Phase int32;
    var nBase int32;
    var Count int32;
    var vStr *Vec_Str_t  = Vec_StrStart( nVars + 1 );
    var pTab *Tab_Tab_t  = Tab_TabAlloc( 5 );
    // inlining of macro Vec_WecForEachLevel
    for  i = 0; (i < Vec_WecSize(p)) && (((vCube1) = Vec_WecEntry(p, i)), 1); i++  {
        // add lit pairs
        pLits[2] = TAB_UNUSED;
        // inlining of macro Vec_IntForEachEntry
        for  i1 = 0; (i1 < Vec_IntSize(vCube1)) && (((pLits[0]) = Vec_IntEntry(vCube1, i1)), 1); i1++  {
        // inlining of macro Vec_IntForEachEntryStart
        for  i2 = i1+1; (i2 < Vec_IntSize(vCube1)) && (((pLits[1]) = Vec_IntEntry(vCube1, i2)), 1); i2++  {
            Tab_TabHashAdd( pTab, pLits, DIV_AND, 1 );
        }
        }

        // inlining of macro Vec_WecForEachLevelStart
        for  k = i+1; (k < Vec_WecSize(p)) && (((vCube2) = Vec_WecEntry(p, k)), 1); k++  {
            nBase = Div_FindDiv( vCube1, vCube2, pLitsA, pLitsB );
            if nBase == -1 {
                continue; }
            Type = Div_FindType(pLitsA, pLitsB, pLits, &Phase);
            if Type >= DIV_AND && Type <= DIV_MUX {
                Tab_TabHashAdd( pTab, pLits, Type, nBase ); }

            if fVerbose != 0 {
                fmt.Printf("Pair %d:\n", Count++);
                Div_CubePrintOne( vCube1, vStr, nVars );
                Div_CubePrintOne( vCube2, vStr, nVars );
                fmt.Printf("Result = %d   ", nBase);
                assert(nBase > 0, "nBase > 0");

                fmt.Printf("Type = %s  ", pNames[Type]);
                fmt.Printf("LitA = %d ", pLits[0]);
                fmt.Printf("LitB = %d ", pLits[1]);
                fmt.Printf("LitC = %d ", pLits[2]);
                fmt.Printf("Phase = %d ", Phase);
                fmt.Printf("\n");
            }
        }
    }
    // print the table
    fmt.Printf("Divisors = %d.\n", pTab.nBins);
    vDivs = Tab_TabFindBest( pTab, nDivs );
    // cleanup
    Vec_StrFree( vStr );
    Tab_TabFree( pTab );
    return vDivs;
}

/**Function*************************************************************

  Synopsis    [Solve the enumeration problem.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Bmc_FxSolve(pSat *sat_solver, iOut int32, iAuxVar int32, vVars *Vec_Int_t, fDumpPla int32, fVerbose int32, pCounter *int32, vCubes *Vec_Wec_t) int32 {
    var nBTLimit int32  = 1000000;
    var fUseOrder int32  = 1;
    var vLevel *Vec_Int_t   = nil;
    var vLits *Vec_Int_t    = Vec_IntAlloc( Vec_IntSize(vVars) );
    var vLits2 *Vec_Int_t   = Vec_IntAlloc( Vec_IntSize(vVars) );
    var vCube *Vec_Str_t    = Vec_StrStart( Vec_IntSize(vVars)+1 );
    var status int32;
    var i int32;
    var k int32;
    var n int32;
    var Lit int32;
    var Lit2 int32;
    var iVar int32;
    var nFinal int32;
    var pFinal *int32;
    var pLits [2]int32;
    var nIter int32;
    var RetValue int32;
    var Before, After, Total, nLits int32;
    pLits[0] = Abc_Var2Lit( iOut + 1, 0 ); // F = 1
    pLits[1] = Abc_Var2Lit( iAuxVar, 0 );  // iNewLit
    if vCubes != nil {
        Vec_WecClear( vCubes ); }
    if fDumpPla != 0 {
    fmt.Printf(".i %d\n", Vec_IntSize(vVars)); }
    if fDumpPla != 0 {
    fmt.Printf(".o %d\n", 1); }
    for {
        // find onset minterm
        status = sat_solver_solve( pSat, pLits, pLits + 2, nBTLimit, 0, 0, 0 );
        if status == l_Undef { RetValue = -1; break; }
        if status == l_False { RetValue = 1; break; }
        assert(status == l_True, "status == l_True");
        // collect divisor literals
        Vec_IntClear( vLits );
        Vec_IntPush( vLits, Abc_LitNot(pLits[0]) ); // F = 0
//        Vec_IntForEachEntryReverse( vVars, iVar, i )
        // inlining of macro Vec_IntForEachEntry
        for  i = 0; (i < Vec_IntSize(vVars)) && (((iVar) = Vec_IntEntry(vVars, i)), 1); i++  {
            Vec_IntPush( vLits, sat_solver_var_literal(pSat, iVar) );
        }

        if fUseOrder != 0 {
            //////////////////////////////////////////////////////////////
            // save these literals
            Vec_IntClear( vLits2 );
            Vec_IntAppend( vLits2, vLits );
            Before = Vec_IntSize(vLits2);

            // try removing literals from the cube
            // inlining of macro Vec_IntForEachEntry
            for  k = 0; (k < Vec_IntSize(vLits2)) && (((Lit2) = Vec_IntEntry(vLits2, k)), 1); k++  {
                if Lit2 == Abc_LitNot(pLits[0]) {
                    continue; }
                Vec_IntClear( vLits );
                // inlining of macro Vec_IntForEachEntry
                for  n = 0; (n < Vec_IntSize(vLits2)) && (((Lit) = Vec_IntEntry(vLits2, n)), 1); n++  {
                    if Lit != -1 && Lit != Lit2 {
                        Vec_IntPush( vLits, Lit ); }
                }
                // call sat
                status = sat_solver_solve( pSat, Vec_IntArray(vLits), Vec_IntArray(vLits) + Vec_IntSize(vLits), nBTLimit, 0, 0, 0 );
                if status == l_Undef {
                    assert(false, "0"); }
                if status == l_True { // SAT
                    continue; }
                // Lit2 can be removed
                Vec_IntWriteEntry( vLits2, k, -1 );
            }

            // make one final run
            Vec_IntClear( vLits );
            // inlining of macro Vec_IntForEachEntry
            for  k = 0; (k < Vec_IntSize(vLits2)) && (((Lit2) = Vec_IntEntry(vLits2, k)), 1); k++  {
                if Lit2 != -1 {
                    Vec_IntPush( vLits, Lit2 ); }
            }
            status = sat_solver_solve( pSat, Vec_IntArray(vLits), Vec_IntArray(vLits) + Vec_IntSize(vLits), nBTLimit, 0, 0, 0 );
            assert(status == l_False, "status == l_False");

            // get subset of literals
            nFinal = sat_solver_final( pSat, &pFinal );
            //////////////////////////////////////////////////////////////
        
        } else {
            ///////////////////////////////////////////////////////////////
            // check against offset
            status = sat_solver_solve( pSat, Vec_IntArray(vLits), Vec_IntArray(vLits) + Vec_IntSize(vLits), nBTLimit, 0, 0, 0 );
            if status == l_Undef { RetValue = -1; break; }
            if status == l_True {
                break; }
            assert(status == l_False, "status == l_False");
            // get subset of literals
            nFinal = sat_solver_final( pSat, &pFinal );
            Before = nFinal;
            //printf( "Before %d. ", nFinal );

/*
            // save these literals
            Vec_IntClear( vLits );
            for ( i = 0; i < nFinal; i++ )
                Vec_IntPush( vLits, Abc_LitNot(pFinal[i]) );
            Vec_IntReverseOrder( vLits );

            // make one final run
            status = sat_solver_solve( pSat, Vec_IntArray(vLits), Vec_IntArray(vLits) + Vec_IntSize(vLits), nBTLimit, 0, 0, 0 );
            assert( status == l_False );
            nFinal = sat_solver_final( pSat, &pFinal );
*/

            // save these literals
            Vec_IntClear( vLits2 );
            for  i = 0; i < nFinal; i++  {
                Vec_IntPush( vLits2, Abc_LitNot(pFinal[i]) );
            }
            Vec_IntSort( vLits2, 1 );

            // try removing literals from the cube
            // inlining of macro Vec_IntForEachEntry
            for  k = 0; (k < Vec_IntSize(vLits2)) && (((Lit2) = Vec_IntEntry(vLits2, k)), 1); k++  {
                if Lit2 == Abc_LitNot(pLits[0]) {
                    continue; }
                Vec_IntClear( vLits );
                // inlining of macro Vec_IntForEachEntry
                for  n = 0; (n < Vec_IntSize(vLits2)) && (((Lit) = Vec_IntEntry(vLits2, n)), 1); n++  {
                    if Lit != -1 && Lit != Lit2 {
                        Vec_IntPush( vLits, Lit ); }
                }
                // call sat
                status = sat_solver_solve( pSat, Vec_IntArray(vLits), Vec_IntArray(vLits) + Vec_IntSize(vLits), nBTLimit, 0, 0, 0 );
                if status == l_Undef {
                    assert(false, "0"); }
                if status == l_True { // SAT
                    continue; }
                // Lit2 can be removed
                Vec_IntWriteEntry( vLits2, k, -1 );
            }

            // make one final run
            Vec_IntClear( vLits );
            // inlining of macro Vec_IntForEachEntry
            for  k = 0; (k < Vec_IntSize(vLits2)) && (((Lit2) = Vec_IntEntry(vLits2, k)), 1); k++  {
                if Lit2 != -1 {
                    Vec_IntPush( vLits, Lit2 ); }
            }
            status = sat_solver_solve( pSat, Vec_IntArray(vLits), Vec_IntArray(vLits) + Vec_IntSize(vLits), nBTLimit, 0, 0, 0 );
            assert(status == l_False, "status == l_False");

            // put them back
            nFinal = 0;
            // inlining of macro Vec_IntForEachEntry
            for  k = 0; (k < Vec_IntSize(vLits2)) && (((Lit2) = Vec_IntEntry(vLits2, k)), 1); k++  {
                if Lit2 != -1 {
                    pFinal[nFinal++] = Abc_LitNot(Lit2); }
            }        
            /////////////////////////////////////////////////////////
        }


        //printf( "After %d. \n", nFinal );
        After  = nFinal;
        Total += Before - After;

        // get subset of literals
        //nFinal = sat_solver_final( pSat, &pFinal );
        // compute cube and add clause
        Vec_IntClear( vLits );
        Vec_IntPush( vLits, Abc_LitNot(pLits[1]) ); // NOT(iNewLit)
        if fDumpPla != 0 {
            Vec_StrFill( vCube, Vec_IntSize(vVars), '-' ); }
        if vCubes != nil {
            vLevel = Vec_WecPushLevel( vCubes ); }
        for  i = 0; i < nFinal; i++  {
            if pFinal[i] == pLits[0] {
                continue; }
            Vec_IntPush( vLits, pFinal[i] );
            iVar = Vec_IntFind( vVars, Abc_Lit2Var(pFinal[i]) );   
            assert(iVar >= 0 && iVar < Vec_IntSize(vVars), "iVar >= 0 && iVar < Vec_IntSize(vVars)");
            //printf( "%s%d ", Abc_LitIsCompl(pFinal[i]) ? "+":"-", iVar );
            if fDumpPla != 0 {
                Vec_StrWriteEntry( vCube, iVar, byte((!Abc_LitIsCompl(pFinal[i]) ? '0' : '1')) ); }
            if vLevel != nil {
                Vec_IntPush( vLevel, Abc_Var2Lit(iVar, !Abc_LitIsCompl(pFinal[i])) ); }
        }
        if vCubes != nil {
            Vec_IntSort( vLevel, 0 ); }
        if fDumpPla != 0 {
            fmt.Printf("%s 1\n", Vec_StrArray(vCube)); }
        status = sat_solver_addclause( pSat, Vec_IntArray(vLits), Vec_IntArray(vLits) + Vec_IntSize(vLits) );
        assert((status != 0), "status");
        nLits += Vec_IntSize(vLevel);
        nIter++;
    }
    if fDumpPla != 0 {
    fmt.Printf(".e\n"); }
    if fDumpPla != 0 {
    fmt.Printf(".p %d\n\n", nIter); }
    if fVerbose != 0 {
    fmt.Printf("Cubes = %d.  Reduced = %d.  Lits = %d\n", nIter, Total, nLits); }
    if pCounter != nil {
        *pCounter = nIter; }
//    assert( status == l_True );
    Vec_IntFree( vLits );
    Vec_IntFree( vLits2 );
    Vec_StrFree( vCube );
    return RetValue;
}

/**Function*************************************************************

  Synopsis    []

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Bmc_FxCompute(p *Gia_Man_t) int32 {
    // create dual-output circuit with on-set/off-set
    extern Gia_Man_t * Gia_ManDupOnsetOffset( Gia_Man_t * p );
    var p2 *Gia_Man_t  = Gia_ManDupOnsetOffset( p );
    // create SAT solver
    var pCnf *Cnf_Dat_t  = (Cnf_Dat_t *)Mf_ManGenerateCnf( p2, 8, 0, 0, 0, 0 );
    var pSat *sat_solver  = (sat_solver *)Cnf_DataWriteIntoSolver( pCnf, 1, 0 );
    // compute parameters
    var nOuts int32  = Gia_ManCoNum(p);
    var nCiVars int32 = Gia_ManCiNum(p);
    var iCiVarBeg int32 = pCnf.nVars - nCiVars;// - 1;
    var o int32;
    var i int32;
    var n int32;
    var RetValue int32;
    var nCounter int32;
    var iAuxVarStart int32 = sat_solver_nvars( pSat );
    var nCubes [2][2]int32 = [2][2]int32{{0}};
    // create variables
    var vVars *Vec_Int_t  = Vec_IntAlloc( nCiVars );
    for  n = 0; n < nCiVars; n++  {
        Vec_IntPush( vVars, iCiVarBeg + n );
    }
    sat_solver_setnvars( pSat, iAuxVarStart + 4*nOuts );
    // iterate through outputs
    for  o = 0; o < nOuts; o++  {
    for  i = 0; i < 2; i++  {
    for  n = 0; n < 2; n++  { // 0=onset, 1=offset
//    for ( n = 1; n >= 0; n-- ) // 0=onset, 1=offset
    
        fmt.Printf("Out %3d %sset ", o, n ? "off" : " on");
        RetValue = Bmc_FxSolve( pSat, Abc_Var2Lit(o, n), iAuxVarStart + 2*i*nOuts + Abc_Var2Lit(o, n), vVars, 0, 0, &nCounter, nil );
        if RetValue == 0 {
            fmt.Printf("Mismatch\n"); }
        if RetValue == -1 {
            fmt.Printf("Timeout\n"); }
        nCubes[i][n] += nCounter;
    }
    }
    }
    // cleanup
    Vec_IntFree( vVars );
    sat_solver_delete( pSat );
    Cnf_DataFree( pCnf );
    Gia_ManStop( p2 );
    fmt.Printf("Onset = %5d.   Offset = %5d.      Onset = %5d.   Offset = %5d.\n", nCubes[0][0], nCubes[0][1], nCubes[1][0], nCubes[1][1]);
    return 1;
}

/**Function*************************************************************

  Synopsis    []

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Bmc_FxAddClauses(pSat *sat_solver, vDivs *Vec_Int_t, iCiVarBeg int32, iVarStart int32) {
    var i int32;
    var Func int32;
    var pLits [3]int32;
    var nDivs int32 = Vec_IntSize(vDivs)/4;
    assert(Vec_IntSize(vDivs) % 4 == 0, "Vec_IntSize(vDivs) % 4 == 0");
    // create new var for each divisor
    for  i = 0; i < nDivs; i++  {
        Func     = Vec_IntEntry(vDivs, 4*i+0);
        pLits[0] = Vec_IntEntry(vDivs, 4*i+1);
        pLits[1] = Vec_IntEntry(vDivs, 4*i+2);
        pLits[2] = Vec_IntEntry(vDivs, 4*i+3);
        //printf( "Adding clause with vars %d %d -> %d\n", iCiVarBeg + Abc_Lit2Var(pLits[0]), iCiVarBeg + Abc_Lit2Var(pLits[1]), iVarStart + nDivs - 1 - i );
        if Func == DIV_AND {
            sat_solver_add_and( pSat, 
                iVarStart + nDivs - 1 - i, iCiVarBeg + Abc_Lit2Var(pLits[0]), iCiVarBeg + Abc_Lit2Var(pLits[1]), 
                Abc_LitIsCompl(pLits[0]), Abc_LitIsCompl(pLits[1]), 0 );
        } else if Func == DIV_XOR {
            sat_solver_add_xor( pSat, 
                iVarStart + nDivs - 1 - i, iCiVarBeg + Abc_Lit2Var(pLits[0]), iCiVarBeg + Abc_Lit2Var(pLits[1]), 0 );
        } else if Func == DIV_MUX {
            sat_solver_add_mux( pSat, 
                iVarStart + nDivs - 1 - i, iCiVarBeg + Abc_Lit2Var(pLits[0]), iCiVarBeg + Abc_Lit2Var(pLits[1]), iCiVarBeg + Abc_Lit2Var(pLits[2]), 
                Abc_LitIsCompl(pLits[0]), Abc_LitIsCompl(pLits[1]), Abc_LitIsCompl(pLits[2]), 0 );
        } else { assert(false, "0"); }
    }
}
func Bmc_FxComputeOne(p *Gia_Man_t, nIterMax int32, nDiv2Add int32) int32 {
    var Extra int32     = 1000;
    // create SAT solver
    var pCnf *Cnf_Dat_t  = (Cnf_Dat_t *)Mf_ManGenerateCnf( p, 8, 0, 0, 0, 0 );
    var pSat *sat_solver  = (sat_solver *)Cnf_DataWriteIntoSolver( pCnf, 1, 0 );
    // compute parameters
    var nCiVars int32    = Gia_ManCiNum(p);             // PI count
    var iCiVarBeg int32  = pCnf.nVars - nCiVars;//- 1;   // first PI var
    var iCiVarCur int32  = iCiVarBeg + nCiVars;         // current unused PI var
    var n, Iter, RetValue int32;
    // create variables
    var iAuxVarStart int32  = sat_solver_nvars(pSat) + Extra; // the aux var
    sat_solver_setnvars( pSat, iAuxVarStart + 1 + nIterMax );
    for  Iter = 0; Iter < nIterMax; Iter++  {
        var vCubes *Vec_Wec_t  = Vec_WecAlloc( 1000 );
        // collect variables
        var vVar2Sat *Vec_Int_t = Vec_IntAlloc( iCiVarCur-iCiVarBeg );
        var vDivs *Vec_Int_t;
//        for ( n = iCiVarCur - 1; n >= iCiVarBeg; n-- )
        for  n = iCiVarBeg; n < iCiVarCur; n++  {
            Vec_IntPush( vVar2Sat, n );
        }
        // iterate through outputs
        fmt.Printf("\nIteration %d (Aux = %d)\n", Iter, iAuxVarStart + Iter);
        RetValue = Bmc_FxSolve( pSat, 0, iAuxVarStart + Iter, vVar2Sat, 1, 1, nil, vCubes );
        if RetValue == 0 {
            fmt.Printf("Mismatch\n"); }
        if RetValue == -1 {
            fmt.Printf("Timeout\n"); }
        // print cubes
        //Div_CubePrint( vCubes, nCiVars );
        vDivs = Div_CubePairs( vCubes, nCiVars, nDiv2Add );
        Vec_WecFree( vCubes );
        // add clauses and update variables
        Bmc_FxAddClauses( pSat, vDivs, iCiVarBeg, iCiVarCur );
        iCiVarCur += Vec_IntSize(vDivs)/4;
        Vec_IntFree( vDivs );
        // cleanup
        assert(Vec_IntSize(vVar2Sat) <= nCiVars + Extra, "Vec_IntSize(vVar2Sat) <= nCiVars + Extra");
        Vec_IntFree( vVar2Sat );
    }
    // cleanup
    sat_solver_delete( pSat );
    Cnf_DataFree( pCnf );
    return 1;
}


////////////////////////////////////////////////////////////////////////
///                       END OF FILE                                ///
////////////////////////////////////////////////////////////////////////


//ABC_NAMESPACE_IMPL_END

