package gia

import (
	"fmt"
	"os"
)

/**CFile****************************************************************

  FileName    [giaResub.c]

  SystemName  [ABC: Logic synthesis and verification system.]

  PackageName [Scalable AIG package.]

  Synopsis    [Resubstitution.]

  Author      [Alan Mishchenko]
  
  Affiliation [UC Berkeley]

  Date        [Ver. 1.0. Started - June 20, 2005.]

  Revision    [$Id: giaResub.c,v 1.00 2005/06/20 00:00:00 alanmi Exp $]

***********************************************************************/

//#include "gia.h"
//#include "misc/vec/vecWec.h"
//#include "misc/vec/vecQue.h"
//#include "misc/vec/vecHsh.h"
//#include "misc/util/utilTruth.h"

//ABC_NAMESPACE_IMPL_START

////////////////////////////////////////////////////////////////////////
///                        DECLARATIONS                              ///
////////////////////////////////////////////////////////////////////////

////////////////////////////////////////////////////////////////////////
///                     FUNCTION DEFINITIONS                         ///
////////////////////////////////////////////////////////////////////////

/**Function*************************************************************

  Synopsis    [Computes MFFCs of all qualifying nodes.]

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

  Synopsis    [Compute divisors and Boolean functions for the nodes.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/

func Gia_ManResubTest(p *Gia_Man_t) {
    var vMffcs *Vec_Wec_t;
    var pNew *Gia_Man_t  = Gia_ManDupMuxes( p, 2 );
var clkStart abctime  = Abc_Clock();
    vMffcs = Gia_ManComputeMffcs( pNew, 4, 100, 8, 100 );
    Gia_ManAddDivisors( pNew, vMffcs );
    Vec_WecFree( vMffcs );
Abc_PrintTime( 1, "Time", Abc_Clock() - clkStart );
    Gia_ManStop( pNew );
}





/**Function*************************************************************

  Synopsis    [Resubstitution data-structure.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
   





/**Function*************************************************************

  Synopsis    [Print resubstitution.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Gia_ManResubPrintNode(vRes *Vec_Int_t, nVars int32, Node int32, fCompl int32) {
    extern void Gia_ManResubPrintLit( Vec_Int_t * vRes, int nVars, int iLit );
    var iLit0 int32  = Vec_IntEntry( vRes, 2*Node + 0 );
    var iLit1 int32  = Vec_IntEntry( vRes, 2*Node + 1 );
    assert(iLit0 != iLit1, "iLit0 != iLit1");
    if iLit0 > iLit1 && Abc_LitIsCompl(fCompl) { // xor
    
        fmt.Printf("~");
        fCompl = 0;
    }
    fmt.Printf("(");
    Gia_ManResubPrintLit( vRes, nVars, Abc_LitNotCond(iLit0, fCompl) );
    fmt.Printf(" %c ", iLit0 > iLit1 ? '^' : (fCompl ? '|' : '&'));
    Gia_ManResubPrintLit( vRes, nVars, Abc_LitNotCond(iLit1, fCompl) );
    fmt.Printf(")");
}
func Gia_ManResubPrintLit(vRes *Vec_Int_t, nVars int32, iLit int32) {
    if Abc_Lit2Var(iLit) < nVars {
        if nVars < 26 {
            fmt.Printf("%s%c", Abc_LitIsCompl(iLit) ? "~":"", 'a' + Abc_Lit2Var(iLit)-2);
        } else {
            fmt.Printf("%si%d", Abc_LitIsCompl(iLit) ? "~":"", Abc_Lit2Var(iLit)-2); }
    
    } else {
        Gia_ManResubPrintNode( vRes, nVars, Abc_Lit2Var(iLit) - nVars, Abc_LitIsCompl(iLit) ); }
}
func Gia_ManResubPrint(vRes *Vec_Int_t, nVars int32) int32 {
    var iTopLit int32;
    if Vec_IntSize(vRes) == 0 {
        return fmt.Printf("none"); }
    assert(Vec_IntSize(vRes) % 2 == 1, "Vec_IntSize(vRes) % 2 == 1");
    iTopLit = Vec_IntEntryLast(vRes);
    if iTopLit == 0 {
        return fmt.Printf("const0"); }
    if iTopLit == 1 {
        return fmt.Printf("const1"); }
    Gia_ManResubPrintLit( vRes, nVars, iTopLit );
    return 0;
}

/**Function*************************************************************

  Synopsis    [Verify resubstitution.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/


/**Function*************************************************************

  Synopsis    [Construct AIG manager from gates.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/



/**Function*************************************************************

  Synopsis    [Construct AIG manager from gates.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/




/**Function*************************************************************

  Synopsis    [Perform resubstitution.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/






























/**Function*************************************************************

  Synopsis    [Perform resubstitution.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/




/**Function*************************************************************

  Synopsis    [Perform resubstitution.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/


func Gia_ManResubOne(vDivs *Vec_Ptr_t, nWords int32, nLimit int32, nDivsMax int32, iChoice int32, fUseXor int32, fDebug int32, fVerbose int32, pFunc *word) *Vec_Int_t {
    var vRes *Vec_Int_t;
    var p *Gia_ResbMan_t  = Gia_ResbAlloc( nWords );
    Gia_ManResubPerform( p, vDivs, nWords, nLimit, nDivsMax, iChoice, fUseXor, fDebug, fVerbose );
    if fVerbose != 0 {
        Gia_ManResubPrint( p.vGates, Vec_PtrSize(vDivs) ); }
    if fVerbose != 0 {
        fmt.Printf("\n"); }
    if !(Gia_ManResubVerify(p, pFunc) != 0) {
        Gia_ManResubPrint( p.vGates, Vec_PtrSize(vDivs) );
        fmt.Printf("Verification FAILED.\n");
    
    } else if (fDebug != 0) && (fVerbose != 0) {
        fmt.Printf("Verification succeeded."); }
    if fVerbose != 0 {
        fmt.Printf("\n"); }
    vRes = Vec_IntDup( p.vGates );
    Gia_ResbFree( p );
    return vRes;
}

/**Function*************************************************************

  Synopsis    [Top level.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
var s_pResbMan *Gia_ResbMan_t = nil;

func Abc_ResubPrepareManager(nWords int32) {
    if s_pResbMan != nil {
        Gia_ResbFree( s_pResbMan ); }
    s_pResbMan = nil;
    if nWords > 0 {
        s_pResbMan = Gia_ResbAlloc( nWords ); }
}

func Abc_ResubComputeFunction(ppDivs *any, nDivs int32, nWords int32, nLimit int32, nDivsMax int32, iChoice int32, fUseXor int32, fDebug int32, fVerbose int32, ppArray **int32) int32 {
    var Divs Vec_Ptr_t  = Vec_Ptr_t{ nDivs, nDivs, ppDivs };
    assert(s_pResbMan != nil, "s_pResbMan != nil"); // first call Abc_ResubPrepareManager()
    Gia_ManResubPerform( s_pResbMan, &Divs, nWords, nLimit, nDivsMax, iChoice, fUseXor, fDebug, fVerbose==2 );
    if fVerbose != 0 {
        var nGates int32  = Vec_IntSize(s_pResbMan.vGates)/2;
        if nGates != 0 {
            fmt.Printf("      Gain = %2d  Gates = %2d  __________  F = ", nLimit+1-nGates, nGates);
            Gia_ManResubPrint( s_pResbMan.vGates, nDivs );
            fmt.Printf("\n");
        }
    }
    if fDebug != 0 {
        if !(Gia_ManResubVerify(s_pResbMan, nil) != 0) {
            Gia_ManResubPrint( s_pResbMan.vGates, nDivs );
            fmt.Printf("Verification FAILED.\n");
        }
        //else
        //    printf( "Verification succeeded.\n" );
    }
    *ppArray = Vec_IntArray(s_pResbMan.vGates);
    assert(Vec_IntSize(s_pResbMan.vGates)/2 <= nLimit, "Vec_IntSize(s_pResbMan.vGates)/2 <= nLimit");
    return Vec_IntSize(s_pResbMan.vGates);
}

func Abc_ResubDumpProblem(pFileName string, ppDivs *any, nDivs int32, nWords int32) {
    var vSims *Vec_Wrd_t  = Vec_WrdAlloc( nDivs * nWords );
    var pDivs **word  = (word **)ppDivs;
    var d, w int32;
    for  d = 0; d < nDivs;  d++  {
    for  w = 0; w < nWords; w++  {
        Vec_WrdPush( vSims, pDivs[d][w] );
    }
    }
    Vec_WrdDumpHex( pFileName, vSims, nWords, 1 );
    Vec_WrdFree( vSims );
}

/**Function*************************************************************

  Synopsis    [Top level.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/

// extern void Extra_PrintHex( FILE * pFile, unsigned * pTruth, int nVars );

// extern void Dau_DsdPrintFromTruth2( word * pTruth, int nVarsInit );


func Gia_ManResubTest3() {
    var nVars int32  = 4;
    var fVerbose int32  = 1;
    var Divs [6]word = [6]word{ 0, 0, 
        ABC_CONST(0xAAAAAAAAAAAAAAAA),
        ABC_CONST(0xCCCCCCCCCCCCCCCC),
        ABC_CONST(0xF0F0F0F0F0F0F0F0),
        ABC_CONST(0xFF00FF00FF00FF00), 
    };
    var vDivs *Vec_Ptr_t  = Vec_PtrAlloc( 6 );
    var vRes *Vec_Int_t  = Vec_IntAlloc( 100 );
    var i int32;
    var k int32;
    var ArraySize int32;
    var pArray *int32; 
    for  i = 0; i < 6; i++  {
        Vec_PtrPush( vDivs, Divs+i );
    }
    Abc_ResubPrepareManager( 1 );
    for  i = 0; i < (1<<(1<<nVars)); i++  { //if ( i == 0xCA ) 
    
        var Truth word  = Abc_Tt6Stretch( i, nVars );
        Divs[0] = ~Truth;
        Divs[1] =  Truth;
        fmt.Printf("%3d : ", i);
        Extra_PrintHex( os.Stdout, (unsigned*)&Truth, nVars );
        fmt.Printf(" ");
        Dau_DsdPrintFromTruth2( &Truth, nVars );
        fmt.Printf("           ");

        //Abc_ResubDumpProblem( "temp.resub", (void **)Vec_PtrArray(vDivs), Vec_PtrSize(vDivs), 1 );
        ArraySize = Abc_ResubComputeFunction( (void **)Vec_PtrArray(vDivs), Vec_PtrSize(vDivs), 1, 16, 50, 0, 0, 1, fVerbose, &pArray );
        fmt.Printf("\n");

        Vec_IntClear( vRes );
        for  k = 0; k < ArraySize; k++  {
            Vec_IntPush( vRes, pArray[k] );
        }

        if i == 1000 {
            break; }
    }
    Abc_ResubPrepareManager( 0 );
    Vec_IntFree( vRes );
    Vec_PtrFree( vDivs );
}
func Gia_ManResubTest3_() {
    var p *Gia_ResbMan_t  = Gia_ResbAlloc( 1 );
    var Divs [6]word = [6]word{ 0, 0, 
        ABC_CONST(0xAAAAAAAAAAAAAAAA),
        ABC_CONST(0xCCCCCCCCCCCCCCCC),
        ABC_CONST(0xF0F0F0F0F0F0F0F0),
        ABC_CONST(0xFF00FF00FF00FF00), 
    };
    var vDivs *Vec_Ptr_t  = Vec_PtrAlloc( 6 );
    var vRes *Vec_Int_t  = Vec_IntAlloc( 100 );
    var i int32; 
    for  i = 0; i < 6; i++  {
        Vec_PtrPush( vDivs, Divs+i );
    }

    {
        var Truth word  = (Divs[2] | Divs[3]) & (Divs[4] & Divs[5]);
//        word Truth = (~Divs[2] | Divs[3]) | ~Divs[4];
        Divs[0] = ~Truth;
        Divs[1] =  Truth;
        Extra_PrintHex( os.Stdout, (unsigned*)&Truth, 6 );
        fmt.Printf(" ");
        Dau_DsdPrintFromTruth2( &Truth, 6 );
        fmt.Printf("       ");
        Gia_ManResubPerform( p, vDivs, 1, 100, 0, 50, 1, 1, 0 );
    }
    Gia_ResbFree( p );
    Vec_IntFree( vRes );
    Vec_PtrFree( vDivs );
}

/**Function*************************************************************

  Synopsis    [Top level.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Gia_ManCheckResub(vDivs *Vec_Ptr_t, nWords int32) {
    //int i, nVars = 6, pVarSet[10] = { 2, 189, 2127, 2125, 177, 178 };
    var i int32;
    var nVars int32 = 3;
    var pVarSet [10]int32 = [10]int32{ 2, 3, 4 };
    var pOff *word  = (word *)Vec_PtrEntry( vDivs, 0 );
    var pOn *word   = (word *)Vec_PtrEntry( vDivs, 1 );
    var vValue *Vec_Int_t  = Vec_IntStartFull( 1 << 6 );
    fmt.Printf("Verifying resub:\n");
    for  i = 0; i < 64*nWords; i++  {
        var v int32;
        var Mint int32;
        var Value int32 = Abc_TtGetBit(pOn, i);
        if !Abc_TtGetBit(pOff, i) && !(Value != 0) {
            continue; }
        for  v = 0; v < nVars; v++  {
            if Abc_TtGetBit((word *)Vec_PtrEntry(vDivs, pVarSet[v]), i) {
                Mint |= 1 << v; }
        }
        if Vec_IntEntry(vValue, Mint) == -1 {
            Vec_IntWriteEntry(vValue, Mint, Value);
        } else if Vec_IntEntry(vValue, Mint) != Value {
            fmt.Printf("Mismatch in pattern %d\n", i); }
    }
    fmt.Printf("Finished verifying resub.\n");
    Vec_IntFree( vValue );
}
func Gia_ManDeriveDivs(vSims *Vec_Wrd_t, nWords int32) *Vec_Ptr_t {
    var i int32;
    var nDivs int32 = Vec_WrdSize(vSims)/nWords;
    var vDivs *Vec_Ptr_t  = Vec_PtrAlloc( nDivs );
    for  i = 0; i < nDivs; i++  {
        Vec_PtrPush( vDivs, Vec_WrdEntryP(vSims, nWords*i) );
    }
    return vDivs;
}
func Gia_ManResub2(pGia *Gia_Man_t, nNodes int32, nSupp int32, nDivs int32, iChoice int32, fUseXor int32, fVerbose int32, fVeryVerbose int32) *Gia_Man_t {
    return nil;
}
func Gia_ManResub1(pFileName string, nNodes int32, nSupp int32, nDivs int32, iChoice int32, fUseXor int32, fVerbose int32, fVeryVerbose int32) *Gia_Man_t {
    var nWords int32;
    var pMan *Gia_Man_t    = nil;
    var vSims *Vec_Wrd_t   = Vec_WrdReadHex( pFileName, &nWords, 1 );
    var vDivs *Vec_Ptr_t;
    if vSims != nil {
        vDivs = Gia_ManDeriveDivs( vSims, nWords );
    } else {
        vDivs = nil;
    }
    var p *Gia_ResbMan_t  = Gia_ResbAlloc( nWords );
    //Gia_ManCheckResub( vDivs, nWords );
    if Vec_PtrSize(vDivs) >= (1<<14) {
        fmt.Printf("Reducing all divs from %d to %d.\n", Vec_PtrSize(vDivs), (1<<14)-1);
        Vec_PtrShrink( vDivs, (1<<14)-1 );
    }
    assert(Vec_PtrSize(vDivs) < (1<<14), "Vec_PtrSize(vDivs) < (1<<14)");
    Gia_ManResubPerform( p, vDivs, nWords, 100, 50, iChoice, fUseXor, 1, 1 );
    if Vec_IntSize(p.vGates) {
        var vGates *Vec_Wec_t  = Vec_WecStart(1);
        Vec_IntAppend( Vec_WecEntry(vGates, 0), p.vGates );
        pMan = Gia_ManConstructFromGates( vGates, Vec_PtrSize(vDivs) );
        Vec_WecFree( vGates );
    
    } else {
        fmt.Printf("Decomposition did not succeed.\n"); }
    Gia_ResbFree( p );
    Vec_PtrFree( vDivs );
    Vec_WrdFree( vSims );
    return pMan;
}

/**Function*************************************************************

  Synopsis    []

  Description []

  SideEffects []

  SeeAlso     []

***********************************************************************/
func Gia_ManUnivTfo_rec(p *Gia_Man_t, iObj int32, vNodes *Vec_Int_t, vPos *Vec_Int_t) int32 {
    var i int32;
    var iFan int32;
    var Count int32 = 1;
    if Gia_ObjIsTravIdCurrentId(p, iObj) {
        return 0; }
    Gia_ObjSetTravIdCurrentId(p, iObj);
    if (vNodes != nil) && Gia_ObjIsCo(Gia_ManObj(p, iObj)) {
        Vec_IntPush( vNodes, iObj ); }    
    if (vPos != nil) && Gia_ObjIsCo(Gia_ManObj(p, iObj)) {
        Vec_IntPush( vPos, iObj ); }
    // inlining of macro Gia_ObjForEachFanoutStaticId
    for  i = 0; (i < Gia_ObjFanoutNumId(p, iObj))   && (((iFan) = Gia_ObjFanoutId(p, iObj, i)), 1); i++  {
        Count += Gia_ManUnivTfo_rec( p, iFan, vNodes, vPos );
    }
    return Count;
}
func Gia_ManUnivTfo(p *Gia_Man_t, pObjs *int32, nObjs int32, pvNodes **Vec_Int_t, pvPos **Vec_Int_t) int32 {
    var i, Count int32;
    if pvNodes {
        if *pvNodes {
            Vec_IntClear( *pvNodes );
        } else {
            *pvNodes = Vec_IntAlloc( 100 ); }
    }
    if pvPos {
        if *pvPos {
            Vec_IntClear( *pvPos );
        } else {
            *pvPos = Vec_IntAlloc( 100 ); }
    }
    Gia_ManIncrementTravId( p );
    for  i = 0; i < nObjs; i++  {
        Count += Gia_ManUnivTfo_rec( p, pObjs[i], pvNodes ? *pvNodes : nil, pvPos ? *pvPos : nil );
    }
    if pvNodes {
        Vec_IntSort( *pvNodes, 0 ); }
    if pvPos {
        Vec_IntSort( *pvPos, 0 ); }
    return Count;
}

/**Function*************************************************************

  Synopsis    [Tuning resub.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Gia_ManTryResub(p *Gia_Man_t) {
    var nLimit int32    =  20;
    var nDivsMax int32  = 200;
    var iChoice int32;
    var fUseXor int32   =   1;
    var fDebug int32    =   1;
    var fVerbose int32;
    var clk abctime;
    var clkResub abctime;
    var clkStart abctime = Abc_Clock();
    var vvSims *Vec_Ptr_t  = Vec_PtrAlloc( 100 );
    var vSims *Vec_Wrd_t;
    var pSets [2]*word;
    var pFunc *word;
    var pObj, pObj2 *Gia_Obj_t; 
    var i, i2, nWords, nNonDec, nTotal int32;
    assert(Gia_ManCiNum(p) < 16, "Gia_ManCiNum(p) < 16");
    Vec_WrdFreeP( &p.vSimsPi );
    p.vSimsPi = Vec_WrdStartTruthTables( Gia_ManCiNum(p) );
    nWords = Vec_WrdSize(p.vSimsPi) / Gia_ManCiNum(p);
    //Vec_WrdPrintHex( p->vSimsPi, nWords );
    // inlining of macro ABC_CALLOC
    pSets[0] = ((word *) calloc((size_t)(nWords), sizeof(word)));
    // inlining of macro ABC_CALLOC
    pSets[1] = ((word *) calloc((size_t)(nWords), sizeof(word)));
    vSims = Gia_ManSimPatSim( p );
    Gia_ManLevelNum(p);
    Gia_ManCreateRefs(p);
    Abc_ResubPrepareManager( nWords );
    Gia_ManStaticFanoutStart( p );
    // inlining of macro Gia_ManForEachAnd
    for  i = 0; (i < p.nObjs) && ((pObj) = Gia_ManObj(p, i)); i++  {      if !Gia_ObjIsAnd(pObj) {} else {
        var vGates Vec_Int_t;
        var pArray *int32;
        var nArray int32;
        var nTfo int32;
        var iObj int32 = Gia_ObjId(p, pObj);
        var Level int32  = Gia_ObjLevel(p, pObj);
        var nMffc int32  = Gia_NodeMffcSizeMark(p, pObj);
        pFunc = Vec_WrdEntryP( vSims, nWords*iObj );
        Abc_TtCopy( pSets[0], pFunc, nWords, 1 );
        Abc_TtCopy( pSets[1], pFunc, nWords, 0 );
        Vec_PtrClear( vvSims );
        Vec_PtrPushTwo( vvSims, pSets[0], pSets[1] );
        nTfo = Gia_ManUnivTfo( p, &iObj, 1, nil, nil );
        // inlining of macro Gia_ManForEachCi
        for  i2 = 0; (i2 < Vec_IntSize(p.vCis)) && ((pObj2) = Gia_ManCi(p, i2)); i2++  {
            Vec_PtrPush( vvSims, Vec_WrdEntryP(vSims, nWords*Gia_ObjId(p, pObj2)) );
        }
        // inlining of macro Gia_ManForEachAnd
        for  i2 = 0; (i2 < p.nObjs) && ((pObj2) = Gia_ManObj(p, i2)); i2++  {      if !Gia_ObjIsAnd(pObj2) {} else
            if !Gia_ObjIsTravIdCurrent(p, pObj2) && !Gia_ObjIsTravIdPrevious(p, pObj2) && Gia_ObjLevel(p, pObj2) <= Level {
                Vec_PtrPush( vvSims, Vec_WrdEntryP(vSims, nWords*Gia_ObjId(p, pObj2)) ); }
        }
        if fVerbose != 0 {
        fmt.Printf("%3d : Lev = %2d  Mffc = %2d  Divs = %3d  Tfo = %3d\n", iObj, Level, nMffc, Vec_PtrSize(vvSims)-2, nTfo); }
        clk = Abc_Clock();
        nArray = Abc_ResubComputeFunction( (void **)Vec_PtrArray(vvSims), Vec_PtrSize(vvSims), nWords, Abc_MinInt(nMffc-1, nLimit), nDivsMax, iChoice, fUseXor, fDebug, fVerbose, &pArray );
        clkResub += Abc_Clock() - clk;
        vGates.nSize = vGates.nCap = nArray;
        vGates.pArray = pArray;
        assert(nMffc > Vec_IntSize(&vGates)/2, "nMffc > Vec_IntSize(&vGates)/2");
        if Vec_IntSize(&vGates) > 0 {
            nTotal += nMffc - Vec_IntSize(&vGates)/2; }
        nNonDec += Vec_IntSize(&vGates) == 0;
    }
    }
    fmt.Printf("Total nodes = %5d.  Non-realizable = %5d.  Gain = %6d.  ", Gia_ManAndNum(p), nNonDec, nTotal);
    Abc_PrintTime( 1, "Time", Abc_Clock() - clkStart );
    Abc_PrintTime( 1, "Pure resub time", clkResub );
    Abc_ResubPrepareManager( 0 );
    Gia_ManStaticFanoutStop( p );
    Vec_PtrFree( vvSims );
    Vec_WrdFree( vSims );
    // inlining of macro ABC_FREE
    pSets[0] = nil;
    // inlining of macro ABC_FREE
    pSets[1] = nil;
}


////////////////////////////////////////////////////////////////////////
///                       END OF FILE                                ///
////////////////////////////////////////////////////////////////////////


//ABC_NAMESPACE_IMPL_END

