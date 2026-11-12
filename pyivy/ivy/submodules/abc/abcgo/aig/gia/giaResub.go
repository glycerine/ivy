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




/**Function*************************************************************

  Synopsis    [Top level.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/








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





/**Function*************************************************************

  Synopsis    []

  Description []

  SideEffects []

  SeeAlso     []

***********************************************************************/



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

