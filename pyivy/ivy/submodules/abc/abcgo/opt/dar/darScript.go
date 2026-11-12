package dar

import (
	"fmt"
	"time"
)

/**CFile****************************************************************

  FileName    [darScript.c]

  SystemName  [ABC: Logic synthesis and verification system.]

  PackageName [DAG-aware AIG rewriting.]

  Synopsis    [Rewriting scripts.]

  Author      [Alan Mishchenko]
  
  Affiliation [UC Berkeley]

  Date        [Ver. 1.0. Started - April 28, 2007.]

  Revision    [$Id: darScript.c,v 1.00 2007/04/28 00:00:00 alanmi Exp $]

***********************************************************************/

//#include "darInt.h"
//#include "proof/dch/dch.h"
//#include "aig/gia/gia.h"
//#include "aig/gia/giaAig.h"

//ABC_NAMESPACE_IMPL_START


////////////////////////////////////////////////////////////////////////
///                        DECLARATIONS                              ///
////////////////////////////////////////////////////////////////////////

////////////////////////////////////////////////////////////////////////
///                     FUNCTION DEFINITIONS                         ///
////////////////////////////////////////////////////////////////////////

/**Function*************************************************************

  Synopsis    [Performs one iteration of AIG rewriting.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Dar_ManRewriteDefault(pAig *Aig_Man_t) *Aig_Man_t {
    var pTemp *Aig_Man_t;
    var Pars Dar_RwrPar_t;
    var pPars *Dar_RwrPar_t = &Pars;
    Dar_ManDefaultRwrParams( pPars );
    pAig = Aig_ManDupDfs( pAig ); 
    Dar_ManRewrite( pAig, pPars );
    pAig = Aig_ManDupDfs( pTemp = pAig ); 
    Aig_ManStop( pTemp );
    return pAig;
}

/**Function*************************************************************

  Synopsis    [Reproduces script "rwsat".]

  Description []
               
  SideEffects [This procedure does not tighten level during restructuring.]

  SeeAlso     []

***********************************************************************/
Aig_Man_t * Dar_ManRwsat( Aig_Man_t * pAig, int fBalance, int fVerbose )
//alias rwsat       "st; rw -l; b -l; rw -l; rf -l"
{
    var pTemp *Aig_Man_t;
    var Time abctime  = pAig.Time2Quit;

    var ParsRwr Dar_RwrPar_t;
    var pParsRwr *Dar_RwrPar_t = &ParsRwr;
    var ParsRef Dar_RefPar_t;
    var pParsRef *Dar_RefPar_t = &ParsRef;

    Dar_ManDefaultRwrParams( pParsRwr );
    Dar_ManDefaultRefParams( pParsRef );

    pParsRwr.fUpdateLevel = 0;
    pParsRef.fUpdateLevel = 0;

    pParsRwr.fVerbose = fVerbose;
    pParsRef.fVerbose = fVerbose;
//printf( "1" );
    pAig = Aig_ManDupDfs( pAig ); 
    if fVerbose != 0 { fmt.Printf("Starting:  ");
Aig_ManPrintStats( pAig ); }

//printf( "2" );
    // balance
    if fBalance != 0 {
    pAig.Time2Quit = Time;
    pAig = Dar_ManBalance( pTemp = pAig, 0 );
    Aig_ManStop( pTemp );
    if fVerbose != 0 { fmt.Printf("Balance:   ");
Aig_ManPrintStats( pAig ); }
    if Time && Abc_Clock() > Time { if pAig != nil { Aig_ManStop( pAig ); }
 return nil; }
    }
   
//Aig_ManDumpBlif( pAig, "inter.blif", NULL, NULL );
//printf( "3" );
    // rewrite
    pAig.Time2Quit = Time;
    Dar_ManRewrite( pAig, pParsRwr );
    pAig = Aig_ManDupDfs( pTemp = pAig ); 
    Aig_ManStop( pTemp );
    if fVerbose != 0 { fmt.Printf("Rewrite:   ");
Aig_ManPrintStats( pAig ); }
    if Time && Abc_Clock() > Time { if pAig != nil { Aig_ManStop( pAig ); }
 return nil; }

//printf( "4" );
    // refactor
    pAig.Time2Quit = Time;
    Dar_ManRefactor( pAig, pParsRef );
    pAig = Aig_ManDupDfs( pTemp = pAig ); 
    Aig_ManStop( pTemp );
    if fVerbose != 0 { fmt.Printf("Refactor:  ");
Aig_ManPrintStats( pAig ); }
    if Time && Abc_Clock() > Time { if pAig != nil { Aig_ManStop( pAig ); }
 return nil; }

//printf( "5" );
    // balance
    if fBalance != 0 {
    pAig.Time2Quit = Time;
    pAig = Dar_ManBalance( pTemp = pAig, 0 );
    Aig_ManStop( pTemp );
    if fVerbose != 0 { fmt.Printf("Balance:   ");
Aig_ManPrintStats( pAig ); }
    if Time && Abc_Clock() > Time { if pAig != nil { Aig_ManStop( pAig ); }
 return nil; }
    }

//printf( "6" );
    // rewrite
    pAig.Time2Quit = Time;
    Dar_ManRewrite( pAig, pParsRwr );
    pAig = Aig_ManDupDfs( pTemp = pAig ); 
    Aig_ManStop( pTemp );
    if fVerbose != 0 { fmt.Printf("Rewrite:   ");
Aig_ManPrintStats( pAig ); }
    if Time && Abc_Clock() > Time { if pAig != nil { Aig_ManStop( pAig ); }
 return nil; }

//printf( "7" );
    return pAig;
}

/**Function*************************************************************

  Synopsis    [Reproduces script "compress".]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
Aig_Man_t * Dar_ManCompress( Aig_Man_t * pAig, int fBalance, int fUpdateLevel, int fPower, int fVerbose )
//alias compress2   "b -l; rw -l; rwz -l; b -l; rwz -l; b -l"
{
    var pTemp *Aig_Man_t;

    var ParsRwr Dar_RwrPar_t;
    var pParsRwr *Dar_RwrPar_t = &ParsRwr;
    var ParsRef Dar_RefPar_t;
    var pParsRef *Dar_RefPar_t = &ParsRef;

    Dar_ManDefaultRwrParams( pParsRwr );
    Dar_ManDefaultRefParams( pParsRef );

    pParsRwr.fUpdateLevel = fUpdateLevel;
    pParsRef.fUpdateLevel = fUpdateLevel;

    pParsRwr.fPower = fPower;

    pParsRwr.fVerbose = 0;//fVerbose;
    pParsRef.fVerbose = 0;//fVerbose;

    pAig = Aig_ManDupDfs( pAig ); 
    if fVerbose != 0 { fmt.Printf("Starting:  ");
Aig_ManPrintStats( pAig ); }
/*
    // balance
    if ( fBalance )
    {
    pAig = Dar_ManBalance( pTemp = pAig, fUpdateLevel );
    Aig_ManStop( pTemp );
    if ( fVerbose ) printf( "Balance:   " ), Aig_ManPrintStats( pAig );
    }
*/    
    // rewrite
    Dar_ManRewrite( pAig, pParsRwr );
    pAig = Aig_ManDupDfs( pTemp = pAig ); 
    Aig_ManStop( pTemp );
    if fVerbose != 0 { fmt.Printf("Rewrite:   ");
Aig_ManPrintStats( pAig ); }
    
    // refactor
    Dar_ManRefactor( pAig, pParsRef );
    pAig = Aig_ManDupDfs( pTemp = pAig ); 
    Aig_ManStop( pTemp );
    if fVerbose != 0 { fmt.Printf("Refactor:  ");
Aig_ManPrintStats( pAig ); }

    // balance
    if fBalance != 0 {
    pAig = Dar_ManBalance( pTemp = pAig, fUpdateLevel );
    Aig_ManStop( pTemp );
    if fVerbose != 0 { fmt.Printf("Balance:   ");
Aig_ManPrintStats( pAig ); }
    }

    pParsRwr.fUseZeros = 1;
    pParsRef.fUseZeros = 1;
    
    // rewrite
    Dar_ManRewrite( pAig, pParsRwr );
    pAig = Aig_ManDupDfs( pTemp = pAig ); 
    Aig_ManStop( pTemp );
    if fVerbose != 0 { fmt.Printf("RewriteZ:  ");
Aig_ManPrintStats( pAig ); }

    return pAig;
}

/**Function*************************************************************

  Synopsis    [Reproduces script "compress2".]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
Aig_Man_t * Dar_ManCompress2( Aig_Man_t * pAig, int fBalance, int fUpdateLevel, int fFanout, int fPower, int fVerbose )
//alias compress2   "b -l; rw -l; rf -l; b -l; rw -l; rwz -l; b -l; rfz -l; rwz -l; b -l"
{
    var pTemp *Aig_Man_t;

    var ParsRwr Dar_RwrPar_t;
    var pParsRwr *Dar_RwrPar_t = &ParsRwr;
    var ParsRef Dar_RefPar_t;
    var pParsRef *Dar_RefPar_t = &ParsRef;

    Dar_ManDefaultRwrParams( pParsRwr );
    Dar_ManDefaultRefParams( pParsRef );

    pParsRwr.fUpdateLevel = fUpdateLevel;
    pParsRef.fUpdateLevel = fUpdateLevel;
    pParsRwr.fFanout = fFanout;
    pParsRwr.fPower = fPower;

    pParsRwr.fVerbose = 0;//fVerbose;
    pParsRef.fVerbose = 0;//fVerbose;

    pAig = Aig_ManDupDfs( pAig ); 
    if fVerbose != 0 { fmt.Printf("Starting:  ");
Aig_ManPrintStats( pAig ); }
/*
    // balance
    if ( fBalance )
    {
    pAig = Dar_ManBalance( pTemp = pAig, fUpdateLevel );
    Aig_ManStop( pTemp );
    if ( fVerbose ) printf( "Balance:   " ), Aig_ManPrintStats( pAig );
    }
*/
    // rewrite
//    Dar_ManRewrite( pAig, pParsRwr );
    pParsRwr.fUpdateLevel = 0;  // disable level update
    Dar_ManRewrite( pAig, pParsRwr );
    pParsRwr.fUpdateLevel = fUpdateLevel;  // reenable level update if needed

    pAig = Aig_ManDupDfs( pTemp = pAig ); 
    Aig_ManStop( pTemp );
    if fVerbose != 0 { fmt.Printf("Rewrite:   ");
Aig_ManPrintStats( pAig ); }
    
    // refactor
    Dar_ManRefactor( pAig, pParsRef );
    pAig = Aig_ManDupDfs( pTemp = pAig ); 
    Aig_ManStop( pTemp );
    if fVerbose != 0 { fmt.Printf("Refactor:  ");
Aig_ManPrintStats( pAig ); }

    // balance
//    if ( fBalance )
    {
    pAig = Dar_ManBalance( pTemp = pAig, fUpdateLevel );
    Aig_ManStop( pTemp );
    if fVerbose != 0 { fmt.Printf("Balance:   ");
Aig_ManPrintStats( pAig ); }
    }
    
    // rewrite
    Dar_ManRewrite( pAig, pParsRwr );
    pAig = Aig_ManDupDfs( pTemp = pAig ); 
    Aig_ManStop( pTemp );
    if fVerbose != 0 { fmt.Printf("Rewrite:   ");
Aig_ManPrintStats( pAig ); }

    pParsRwr.fUseZeros = 1;
    pParsRef.fUseZeros = 1;
    
    // rewrite
    Dar_ManRewrite( pAig, pParsRwr );
    pAig = Aig_ManDupDfs( pTemp = pAig ); 
    Aig_ManStop( pTemp );
    if fVerbose != 0 { fmt.Printf("RewriteZ:  ");
Aig_ManPrintStats( pAig ); }

    // balance
    if fBalance != 0 {
    pAig = Dar_ManBalance( pTemp = pAig, fUpdateLevel );
    Aig_ManStop( pTemp );
    if fVerbose != 0 { fmt.Printf("Balance:   ");
Aig_ManPrintStats( pAig ); }
    }
    
    // refactor
    Dar_ManRefactor( pAig, pParsRef );
    pAig = Aig_ManDupDfs( pTemp = pAig ); 
    Aig_ManStop( pTemp );
    if fVerbose != 0 { fmt.Printf("RefactorZ: ");
Aig_ManPrintStats( pAig ); }
    
    // rewrite
    Dar_ManRewrite( pAig, pParsRwr );
    pAig = Aig_ManDupDfs( pTemp = pAig ); 
    Aig_ManStop( pTemp );
    if fVerbose != 0 { fmt.Printf("RewriteZ:  ");
Aig_ManPrintStats( pAig ); }

    // balance
    if fBalance != 0 {
    pAig = Dar_ManBalance( pTemp = pAig, fUpdateLevel );
    Aig_ManStop( pTemp );
    if fVerbose != 0 { fmt.Printf("Balance:   ");
Aig_ManPrintStats( pAig ); }
    }
    return pAig;
}

/**Function*************************************************************

  Synopsis    [Reproduces script "compress2".]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
Vec_Ptr_t * Dar_ManChoiceSynthesis( Aig_Man_t * pAig, int fBalance, int fUpdateLevel, int fPower, int fVerbose )
//alias resyn    "b; rw; rwz; b; rwz; b"
//alias resyn2   "b; rw; rf; b; rw; rwz; b; rfz; rwz; b"
{
    var vAigs *Vec_Ptr_t;

    vAigs = Vec_PtrAlloc( 3 );
    pAig = Aig_ManDupDfs(pAig);
    Vec_PtrPush( vAigs, pAig );

    pAig = Dar_ManCompress(pAig, fBalance, fUpdateLevel, fPower, fVerbose);
    Vec_PtrPush( vAigs, pAig );
//Aig_ManPrintStats( pAig );

    pAig = Dar_ManCompress2(pAig, fBalance, fUpdateLevel, 1, fPower, fVerbose);
    Vec_PtrPush( vAigs, pAig );
//Aig_ManPrintStats( pAig );

    pAig = (Aig_Man_t *)Vec_PtrEntry( vAigs, 1 );
    return vAigs;
}

/**Function*************************************************************

  Synopsis    [Reproduces script "compress2".]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Dar_ManChoice(pAig *Aig_Man_t, fBalance int32, fUpdateLevel int32, fConstruct int32, nConfMax int32, nLevelMax int32, fVerbose int32) *Aig_Man_t {
    var pMan, pTemp *Aig_Man_t;
    var vAigs *Vec_Ptr_t;
    var i int32;
    var clk abctime;

clk = Abc_Clock();
//    vAigs = Dar_ManChoiceSynthesisExt();
    vAigs = Dar_ManChoiceSynthesis( pAig, fBalance, fUpdateLevel, 0, fVerbose );

    // swap the first and last network
    // this should lead to the primary choice being "better" because of synthesis
    // (it is also important when constructing choices)
    if !(fConstruct != 0) {
        pMan = (Aig_Man_t *)Vec_PtrPop( vAigs );
        Vec_PtrPush( vAigs, Vec_PtrEntry(vAigs,0) );
        Vec_PtrWriteEntry( vAigs, 0, pMan );
    }

if fVerbose != 0 {
// inlining of macro ABC_PRT
Abc_Print(1, "%s =", ("Synthesis time"));
Abc_Print(1, "%9.2f sec\n", 1.0*(float64((Abc_Clock() - clk)))/(float64(int64(time.Second))));;
}
clk = Abc_Clock();
    if fConstruct != 0 {
        pMan = Aig_ManChoiceConstructive( vAigs, fVerbose );
    } else {
        pMan = Aig_ManChoicePartitioned( vAigs, 300, nConfMax, nLevelMax, fVerbose ); }
    // inlining of macro Vec_PtrForEachEntry
    for  i = 0; (i < Vec_PtrSize(vAigs)) && (((pTemp) = (Aig_Man_t *)Vec_PtrEntry(vAigs, i)), 1); i++  {
        Aig_ManStop( pTemp );
    }
    Vec_PtrFree( vAigs );
if fVerbose != 0 {
// inlining of macro ABC_PRT
Abc_Print(1, "%s =", ("Choicing time "));
Abc_Print(1, "%9.2f sec\n", 1.0*(float64((Abc_Clock() - clk)))/(float64(int64(time.Second))));;
}
    return pMan;
//    return NULL;
}


/**Function*************************************************************

  Synopsis    [Reproduces script "compress".]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
Aig_Man_t * Dar_NewCompress( Aig_Man_t * pAig, int fBalance, int fUpdateLevel, int fPower, int fVerbose )
//alias compress2   "b -l; rw -l; rwz -l; b -l; rwz -l; b -l"
{
    var pTemp *Aig_Man_t;

    var ParsRwr Dar_RwrPar_t;
    var pParsRwr *Dar_RwrPar_t = &ParsRwr;
    var ParsRef Dar_RefPar_t;
    var pParsRef *Dar_RefPar_t = &ParsRef;

    Dar_ManDefaultRwrParams( pParsRwr );
    Dar_ManDefaultRefParams( pParsRef );

    pParsRwr.fUpdateLevel = fUpdateLevel;
    pParsRef.fUpdateLevel = fUpdateLevel;

    pParsRwr.fPower = fPower;

    pParsRwr.fVerbose = 0;//fVerbose;
    pParsRef.fVerbose = 0;//fVerbose;

//    pAig = Aig_ManDupDfs( pAig ); 
    if fVerbose != 0 { fmt.Printf("Starting:  ");
Aig_ManPrintStats( pAig ); }

    // rewrite
    Dar_ManRewrite( pAig, pParsRwr );
    pAig = Aig_ManDupDfs( pTemp = pAig ); 
    Aig_ManStop( pTemp );
    if fVerbose != 0 { fmt.Printf("Rewrite:   ");
Aig_ManPrintStats( pAig ); }
    
    // refactor
    Dar_ManRefactor( pAig, pParsRef );
    pAig = Aig_ManDupDfs( pTemp = pAig ); 
    Aig_ManStop( pTemp );
    if fVerbose != 0 { fmt.Printf("Refactor:  ");
Aig_ManPrintStats( pAig ); }

    // balance
    if fBalance != 0 {
    pAig = Dar_ManBalance( pTemp = pAig, fUpdateLevel );
    Aig_ManStop( pTemp );
    if fVerbose != 0 { fmt.Printf("Balance:   ");
Aig_ManPrintStats( pAig ); }
    }

    pParsRwr.fUseZeros = 1;
    pParsRef.fUseZeros = 1;
    
    // rewrite
    Dar_ManRewrite( pAig, pParsRwr );
    pAig = Aig_ManDupDfs( pTemp = pAig ); 
    Aig_ManStop( pTemp );
    if fVerbose != 0 { fmt.Printf("RewriteZ:  ");
Aig_ManPrintStats( pAig ); }

    return pAig;
}

/**Function*************************************************************

  Synopsis    [Reproduces script "compress2".]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
Aig_Man_t * Dar_NewCompress2( Aig_Man_t * pAig, int fBalance, int fUpdateLevel, int fFanout, int fPower, int fLightSynth, int fVerbose )
//alias compress2   "b -l; rw -l; rf -l; b -l; rw -l; rwz -l; b -l; rfz -l; rwz -l; b -l"
{
    var pTemp *Aig_Man_t;

    var ParsRwr Dar_RwrPar_t;
    var pParsRwr *Dar_RwrPar_t = &ParsRwr;
    var ParsRef Dar_RefPar_t;
    var pParsRef *Dar_RefPar_t = &ParsRef;

    Dar_ManDefaultRwrParams( pParsRwr );
    Dar_ManDefaultRefParams( pParsRef );

    pParsRwr.fUpdateLevel = fUpdateLevel;
    pParsRef.fUpdateLevel = fUpdateLevel;
    pParsRwr.fFanout = fFanout;
    pParsRwr.fPower = fPower;

    pParsRwr.fVerbose = 0;//fVerbose;
    pParsRef.fVerbose = 0;//fVerbose;

//    pAig = Aig_ManDupDfs( pAig ); 
    if fVerbose != 0 { fmt.Printf("Starting:  ");
Aig_ManPrintStats( pAig ); }

    // skip if lighter synthesis is requested
    if !fLightSynth {
        // rewrite
        //Dar_ManRewrite( pAig, pParsRwr );
//        pParsRwr->fUpdateLevel = 0;  // disable level update  // this change was requested in July and later disabled
        Dar_ManRewrite( pAig, pParsRwr );
//        pParsRwr->fUpdateLevel = fUpdateLevel;  // reenable level update if needed

        pAig = Aig_ManDupDfs( pTemp = pAig ); 
        Aig_ManStop( pTemp );
        if fVerbose != 0 { fmt.Printf("Rewrite:   ");
Aig_ManPrintStats( pAig ); }
    
        // refactor
        Dar_ManRefactor( pAig, pParsRef );
        pAig = Aig_ManDupDfs( pTemp = pAig ); 
        Aig_ManStop( pTemp );
        if fVerbose != 0 { fmt.Printf("Refactor:  ");
Aig_ManPrintStats( pAig ); }
    }

    // balance
    pAig = Dar_ManBalance( pTemp = pAig, fUpdateLevel );
    Aig_ManStop( pTemp );
    if fVerbose != 0 { fmt.Printf("Balance:   ");
Aig_ManPrintStats( pAig ); }
    
    // skip if lighter synthesis is requested
    if !fLightSynth {
        // rewrite
        Dar_ManRewrite( pAig, pParsRwr );
        pAig = Aig_ManDupDfs( pTemp = pAig ); 
        Aig_ManStop( pTemp );
        if fVerbose != 0 { fmt.Printf("Rewrite:   ");
Aig_ManPrintStats( pAig ); }
    }

    pParsRwr.fUseZeros = 1;
    pParsRef.fUseZeros = 1;
    
    // rewrite
    Dar_ManRewrite( pAig, pParsRwr );
    pAig = Aig_ManDupDfs( pTemp = pAig ); 
    Aig_ManStop( pTemp );
    if fVerbose != 0 { fmt.Printf("RewriteZ:  ");
Aig_ManPrintStats( pAig ); }

    // skip if lighter synthesis is requested
    if !fLightSynth {
        // balance
        if fBalance != 0 {
        pAig = Dar_ManBalance( pTemp = pAig, fUpdateLevel );
        Aig_ManStop( pTemp );
        if fVerbose != 0 { fmt.Printf("Balance:   ");
Aig_ManPrintStats( pAig ); }
        }
    }
    
    // refactor
    Dar_ManRefactor( pAig, pParsRef );
    pAig = Aig_ManDupDfs( pTemp = pAig ); 
    Aig_ManStop( pTemp );
    if fVerbose != 0 { fmt.Printf("RefactorZ: ");
Aig_ManPrintStats( pAig ); }
    
    // skip if lighter synthesis is requested
    if !fLightSynth {
        // rewrite
        Dar_ManRewrite( pAig, pParsRwr );
        pAig = Aig_ManDupDfs( pTemp = pAig ); 
        Aig_ManStop( pTemp );
        if fVerbose != 0 { fmt.Printf("RewriteZ:  ");
Aig_ManPrintStats( pAig ); }
    }

    // balance
    if fBalance != 0 {
    pAig = Dar_ManBalance( pTemp = pAig, fUpdateLevel );
    Aig_ManStop( pTemp );
    if fVerbose != 0 { fmt.Printf("Balance:   ");
Aig_ManPrintStats( pAig ); }
    }
    return pAig;
}

/**Function*************************************************************

  Synopsis    [Count the number of nodes with very high fanout count.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Dar_NewChoiceSynthesisGuard(pAig *Aig_Man_t) int32 {
    var pObj *Aig_Obj_t;
    var i, Count int32;
    // inlining of macro Aig_ManForEachNode
    // inlining of macro Vec_PtrForEachEntry
    for  i = 0; (i < Vec_PtrSize(pAig.vObjs)) && (((pObj) = (Aig_Obj_t *)Vec_PtrEntry(pAig.vObjs, i)), 1); i++  { if (pObj) == nil || !Aig_ObjIsNode(pObj) {} else
        if Aig_ObjRefs(pObj) > 1000 {
            Count += Aig_ObjRefs(pObj) / 1000; }
    }
    return int32((Count > 10));
}

/**Function*************************************************************

  Synopsis    [Reproduces script "compress2".]

  Description [Takes AIG manager, consumes it, and produces GIA manager.]
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
Gia_Man_t * Dar_NewChoiceSynthesis( Aig_Man_t * pAig, int fBalance, int fUpdateLevel, int fPower, int fLightSynth, int fVerbose )
//alias resyn    "b; rw; rwz; b; rwz; b"
//alias resyn2   "b; rw; rf; b; rw; rwz; b; rfz; rwz; b"
{
    var vGias *Vec_Ptr_t;
    var pGia, pTemp *Gia_Man_t;
    var i int32;

    if (fUpdateLevel != 0) && (Dar_NewChoiceSynthesisGuard(pAig) != 0) {
        if fVerbose != 0 {
            fmt.Printf("Warning: Due to high fanout count of some nodes, level updating is disabled.\n"); }
        fUpdateLevel = 0;
    }

    vGias = Vec_PtrAlloc( 3 );
    pGia = Gia_ManFromAig(pAig);
    Vec_PtrPush( vGias, pGia );

    pAig = Dar_NewCompress( pAig, fBalance, fUpdateLevel, fPower, fVerbose );
    pGia = Gia_ManFromAig(pAig);
    Vec_PtrPush( vGias, pGia );
//Aig_ManPrintStats( pAig );

    pAig = Dar_NewCompress2( pAig, fBalance, fUpdateLevel, 1, fPower, fLightSynth, fVerbose );
    pGia = Gia_ManFromAig(pAig);
    Vec_PtrPush( vGias, pGia );
//Aig_ManPrintStats( pAig );

    Aig_ManStop( pAig );

    // swap around the first and the last
    pTemp = (Gia_Man_t *)Vec_PtrPop( vGias );
    Vec_PtrPush( vGias, Vec_PtrEntry(vGias,0) );
    Vec_PtrWriteEntry( vGias, 0, pTemp );

//    Aig_Man_t * pAig;
//    int i;
//    printf( "Choicing will be performed with %d AIGs:\n", Vec_PtrSize(p->vAigs) );
//    Vec_PtrForEachEntry( Aig_Man_t *, p->vAigs, pAig, i )
//        Aig_ManPrintStats( pAig );

    // derive the miter
    pGia = Gia_ManChoiceMiter( vGias );

    // cleanup
    // inlining of macro Vec_PtrForEachEntry
    for  i = 0; (i < Vec_PtrSize(vGias)) && (((pTemp) = (Gia_Man_t *)Vec_PtrEntry(vGias, i)), 1); i++  {
        Gia_ManStop( pTemp );
    }
    Vec_PtrFree( vGias );
    return pGia;
}

/**Function*************************************************************

  Synopsis    [Reproduces script "compress2".]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
/*
Aig_Man_t * Dar_ManChoiceNew( Aig_Man_t * pAig, Dch_Pars_t * pPars )
{
    extern Aig_Man_t * Dch_ComputeChoices( Vec_Ptr_t * vAigs, Dch_Pars_t * pPars );
    extern Aig_Man_t * Cec_ComputeChoices( Vec_Ptr_t * vAigs, Dch_Pars_t * pPars );

    int fVerbose = pPars->fVerbose;
    int fConstruct = 0;
    Aig_Man_t * pMan, * pTemp;
    Vec_Ptr_t * vAigs;
    int i;
    abctime clk;

clk = Abc_Clock();
//    vAigs = Dar_ManChoiceSynthesisExt();
//    vAigs = Dar_ManChoiceSynthesis( pAig, 1, 1, pPars->fPower, fVerbose );
    vAigs = Dar_ManChoiceSynthesis( pAig, 1, 1, pPars->fPower, 0 );

    // swap the first and last network
    // this should lead to the primary choice being "better" because of synthesis
    // (it is also important when constructing choices)
    if ( !fConstruct )
    {
        pMan = Vec_PtrPop( vAigs );
        Vec_PtrPush( vAigs, Vec_PtrEntry(vAigs,0) );
        Vec_PtrWriteEntry( vAigs, 0, pMan );
    }

if ( fVerbose )
{
//ABC_PRT( "Synthesis time", Abc_Clock() - clk );
}
    pPars->timeSynth = Abc_Clock() - clk;

clk = Abc_Clock();
    // perform choice computation
    if ( pPars->fUseGia )
        pMan = Cec_ComputeChoices( vAigs, pPars );
    else
        pMan = Dch_ComputeChoices( vAigs, pPars );

    // reconstruct the network
    pMan = Aig_ManDupDfsGuided( pTemp = pMan, Vec_PtrEntry(vAigs,0) );
    Aig_ManStop( pTemp );
    // duplicate the timing manager
    pTemp = Vec_PtrEntry( vAigs, 0 );
    if ( pTemp->pManTime )
    {
        extern void * Tim_ManDup( void * p, int fDiscrete );     
        pMan->pManTime = Tim_ManDup( pTemp->pManTime, 0 );
    }
    // reset levels
    Aig_ManChoiceLevel( pMan );
    ABC_FREE( pMan->pName );
    ABC_FREE( pMan->pSpec );
    pMan->pName = Abc_UtilStrsav( pTemp->pName );
    pMan->pSpec = Abc_UtilStrsav( pTemp->pSpec );

    // cleanup
    Vec_PtrForEachEntry( Aig_Man_t *, vAigs, pTemp, i )
        Aig_ManStop( pTemp );
    Vec_PtrFree( vAigs );

if ( fVerbose )
{
//ABC_PRT( "Choicing time ", Abc_Clock() - clk );
}
    return pMan;
//    return NULL;
}
*/

/**Function*************************************************************

  Synopsis    [Reproduces script "compress2".]

  Description [Consumes the input AIG to reduce memory usage.]
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Dar_ManChoiceNewAig(pAig *Aig_Man_t, pPars *Dch_Pars_t) *Aig_Man_t {
//    extern Aig_Man_t * Dch_DeriveTotalAig( Vec_Ptr_t * vAigs );
    extern Aig_Man_t * Dch_ComputeChoices( Aig_Man_t * pAig, Dch_Pars_t * pPars );
    var fVerbose int32  = pPars.fVerbose;
    var pMan, pTemp *Aig_Man_t;
    var vAigs *Vec_Ptr_t;
    var vPios *Vec_Ptr_t;
    var pManTime any;
    var pName, pSpec string;
    var i int32;
    var clk abctime;

clk = Abc_Clock();
    vAigs = Dar_ManChoiceSynthesis( pAig, 1, 1, pPars.fPower, fVerbose );
pPars.timeSynth = Abc_Clock() - clk;
    // swap the first and last network
    // this should lead to the primary choice being "better" because of synthesis
    // (it is also important when constructing choices)
    pMan = (Aig_Man_t *)Vec_PtrPop( vAigs );
    Vec_PtrPush( vAigs, Vec_PtrEntry(vAigs,0) );
    Vec_PtrWriteEntry( vAigs, 0, pMan );

    // derive the total AIG
    pMan = Dch_DeriveTotalAig( vAigs );
    // cleanup
    // inlining of macro Vec_PtrForEachEntry
    for  i = 0; (i < Vec_PtrSize(vAigs)) && (((pTemp) = (Aig_Man_t *)Vec_PtrEntry(vAigs, i)), 1); i++  {
        Aig_ManStop( pTemp );
    }
    Vec_PtrFree( vAigs );

    // compute choices
    pMan = Dch_ComputeChoices( pTemp = pMan, pPars );
    Aig_ManStop( pTemp );

    // save useful things
    pManTime = pAig.pManTime; pAig.pManTime = nil;
    pName = Abc_UtilStrsav( pAig.pName );
    pSpec = Abc_UtilStrsav( pAig.pSpec );

    // create guidence
    vPios = Aig_ManOrderPios( pMan, pAig ); 
    Aig_ManStop( pAig );

    // reconstruct the network
    pMan = Aig_ManDupDfsGuided( pTemp = pMan, vPios );
    Aig_ManStop( pTemp );
    Vec_PtrFree( vPios );

    // reset levels
    pMan.pManTime = pManTime;
    Aig_ManChoiceLevel( pMan );

    // copy names
    // inlining of macro ABC_FREE
    ((pMan.pName) ? (free((char *) (pMan.pName)), (pMan.pName) = 0) : 0);
    // inlining of macro ABC_FREE
    ((pMan.pSpec) ? (free((char *) (pMan.pSpec)), (pMan.pSpec) = 0) : 0);
    pMan.pName = pName;
    pMan.pSpec = pSpec;
    return pMan;
}

/**Function*************************************************************

  Synopsis    [Reproduces script "compress2".]

  Description [Consumes the input AIG to reduce memory usage.]
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Dar_ManChoiceNew(pAig *Aig_Man_t, pPars *Dch_Pars_t) *Aig_Man_t {
    extern Aig_Man_t * Cec_ComputeChoicesNew( Gia_Man_t * pGia, int fVerbose );
    extern Aig_Man_t * Cec_ComputeChoices( Gia_Man_t * pGia, Dch_Pars_t * pPars );
//    extern Aig_Man_t * Dch_DeriveTotalAig( Vec_Ptr_t * vAigs );
    extern Aig_Man_t * Dch_ComputeChoices( Aig_Man_t * pAig, Dch_Pars_t * pPars );
//    int fVerbose = pPars->fVerbose;
    var pMan, pTemp *Aig_Man_t;
    var pGia *Gia_Man_t;
    var vPios *Vec_Ptr_t;
    var pManTime any;
    var pName, pSpec string;
    var clk abctime;

    // save useful things
    pManTime = pAig.pManTime; pAig.pManTime = nil;
    pName = Abc_UtilStrsav( pAig.pName );
    pSpec = Abc_UtilStrsav( pAig.pSpec );

    // perform synthesis
clk = Abc_Clock();
    pGia = Dar_NewChoiceSynthesis( Aig_ManDupDfs(pAig), 1, 1, pPars.fPower, pPars.fLightSynth, pPars.fVerbose );
pPars.timeSynth = Abc_Clock() - clk;

    // perform choice computation
    if pPars.fUseNew {
        pMan = Cec_ComputeChoicesNew( pGia, pPars.fVerbose );
    } else if pPars.fUseGia {
        pMan = Cec_ComputeChoices( pGia, pPars );
    } else {
        pMan = Gia_ManToAigSkip( pGia, 3 );
        pMan = Dch_ComputeChoices( pTemp = pMan, pPars );
        Aig_ManStop( pTemp );
    }
    Gia_ManStop( pGia );

    // create guidence
    vPios = Aig_ManOrderPios( pMan, pAig ); 
    Aig_ManStop( pAig );

    // reconstruct the network
    pMan = Aig_ManDupDfsGuided( pTemp = pMan, vPios );
    Aig_ManStop( pTemp );
    Vec_PtrFree( vPios );

    // reset levels
    pMan.pManTime = pManTime;
    Aig_ManChoiceLevel( pMan );

    // copy names
    // inlining of macro ABC_FREE
    ((pMan.pName) ? (free((char *) (pMan.pName)), (pMan.pName) = 0) : 0);
    // inlining of macro ABC_FREE
    ((pMan.pSpec) ? (free((char *) (pMan.pSpec)), (pMan.pSpec) = 0) : 0);
    pMan.pName = pName;
    pMan.pSpec = pSpec;
    return pMan;
}


////////////////////////////////////////////////////////////////////////
///                       END OF FILE                                ///
////////////////////////////////////////////////////////////////////////


//ABC_NAMESPACE_IMPL_END

