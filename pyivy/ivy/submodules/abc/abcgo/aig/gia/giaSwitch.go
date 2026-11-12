package gia

import (
	"fmt"
	"strconv"
	"time"
)

/**CFile****************************************************************

  FileName    [giaSwitch.c]

  SystemName  [ABC: Logic synthesis and verification system.]

  PackageName [Scalable AIG package.]

  Synopsis    [Computing switching activity.]

  Author      [Alan Mishchenko]
  
  Affiliation [UC Berkeley]

  Date        [Ver. 1.0. Started - June 20, 2005.]

  Revision    [$Id: giaSwitch.c,v 1.00 2005/06/20 00:00:00 alanmi Exp $]

***********************************************************************/

//#include "giaAig.h"
//#include "base/main/main.h"

//ABC_NAMESPACE_IMPL_START


////////////////////////////////////////////////////////////////////////
///                        DECLARATIONS                              ///
////////////////////////////////////////////////////////////////////////

// switching estimation parameters
   
func Gia_ManSwiComputeSwitching(nOnes int32, nSimWords int32) float32 {
    var nTotal int32  = 32 * nSimWords;
    return float32(2.0) * nOnes / nTotal * (nTotal - nOnes) / nTotal;
}

/**Function*************************************************************

  Synopsis    [Computes switching activity of one node.]

  Description [Uses the formula: Switching = 2 * nOnes * nZeros / (nTotal ^ 2) ]
               
  SideEffects []

  SeeAlso     [] 

***********************************************************************/
func Gia_ManSwiComputeProbOne(nOnes int32, nSimWords int32) float32 {
    var nTotal int32  = 32 * nSimWords;
    return float32(nOnes) / nTotal;
}

/**Function*************************************************************

  Synopsis    []

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Gia_ManSwiSimulate(pAig *Gia_Man_t, pPars *Gia_ParSwi_t) *Vec_Int_t {
    var p *Gia_ManSwi_t;
    var pObj *Gia_Obj_t;
    var vSwitching *Vec_Int_t;
    var pSwitching *float32;
    var i int32;
    var clk abctime;
    var clkTotal abctime = Abc_Clock();
    if pPars.fProbOne && pPars.fProbTrans {
        fmt.Printf("Conflict of options: Can either compute probability of 1, or probability of switching by observing transitions.\n"); }
    // create manager
    clk = Abc_Clock();
    p = Gia_ManSwiCreate( pAig, pPars );
    if pPars.fVerbose {
        fmt.Printf("Obj = %8d (%8d). F = %6d. ", pAig.nObjs, Gia_ManCiNum(pAig) + Gia_ManAndNum(pAig), p.pAig.nFront);
        fmt.Printf("AIG = %7.2f MB. F-mem = %7.2f MB. Other = %7.2f MB.  ", 12.0*Gia_ManObjNum(p.pAig)/(1<<20), 4.0*p.nWords*p.pAig.nFront/(1<<20), 4.0*p.nWords*(Gia_ManCiNum(p.pAig) + Gia_ManCoNum(p.pAig))/(1<<20));
        // inlining of macro ABC_PRT
        Abc_Print(1, "%s =", ("Time"));
Abc_Print(1, "%9.2f sec\n", 1.0*(float64((Abc_Clock() - clk)))/(float64(int64(time.Second))));;
    }
    // perform simulation
    Gia_ManRandom( 1 );
    Gia_ManSwiSimInfoInit( p );
    for  i = 0; i < pPars.nIters; i++  {
        Gia_ManSwiSimulateRound( p, i >= pPars.nPref );
        if i == pPars.nIters - 1 {
            break; }
        if pPars.fProbTrans {
            Gia_ManSwiSimInfoTransferShift( p, pPars.nRandPiFactor );
        } else {
            Gia_ManSwiSimInfoTransfer( p, pPars.nRandPiFactor ); }
    }
    if pPars.fVerbose {
        fmt.Printf("Simulated %d frames with %d words. ", pPars.nIters, pPars.nWords);
        // inlining of macro ABC_PRT
        Abc_Print(1, "%s =", ("Simulation time"));
Abc_Print(1, "%9.2f sec\n", 1.0*(float64((Abc_Clock() - clkTotal)))/(float64(int64(time.Second))));;
    }
    // derive the result
    vSwitching = Vec_IntStart( Gia_ManObjNum(pAig) );
    pSwitching = (float *)vSwitching.pArray;
    if pPars.fProbOne {
        // inlining of macro Gia_ManForEachObj
        for  i = 0; (i < pAig.nObjs) && ((pObj) = Gia_ManObj(pAig, i)); i++  {
            pSwitching[i] = Gia_ManSwiComputeProbOne( p.pData1[i], pPars.nWords*(pPars.nIters-pPars.nPref) );
        }
        // inlining of macro Gia_ManForEachCo
        for  i = 0; (i < Vec_IntSize(pAig.vCos)) && ((pObj) = Gia_ManCo(pAig, i)); i++  {
            if Gia_ObjFaninC0(pObj) {
                pSwitching[Gia_ObjId(pAig,pObj)] = float32(1.0)-pSwitching[Gia_ObjId(pAig,Gia_ObjFanin0(pObj))];
            } else {
                pSwitching[Gia_ObjId(pAig,pObj)] = pSwitching[Gia_ObjId(pAig,Gia_ObjFanin0(pObj))]; }
        }
    
    } else if pPars.fProbTrans {
        // inlining of macro Gia_ManForEachObj
        for  i = 0; (i < pAig.nObjs) && ((pObj) = Gia_ManObj(pAig, i)); i++  {
            pSwitching[i] = Gia_ManSwiComputeProbOne( p.pData1[i], pPars.nWords*(pPars.nIters-pPars.nPref) );
        }
    
    } else {
        // inlining of macro Gia_ManForEachObj
        for  i = 0; (i < pAig.nObjs) && ((pObj) = Gia_ManObj(pAig, i)); i++  {
            pSwitching[i] = Gia_ManSwiComputeSwitching( p.pData1[i], pPars.nWords*(pPars.nIters-pPars.nPref) );
        }
    }
/*
    printf( "PI: " );
    Gia_ManForEachPi( pAig, pObj, i )
        printf( "%d=%d (%f)  ", i, p->pData1[Gia_ObjId(pAig,pObj)], pSwitching[Gia_ObjId(pAig,pObj)] );
    printf( "\n" );

    printf( "LO: " );
    Gia_ManForEachRo( pAig, pObj, i )
        printf( "%d=%d (%f)  ", i, p->pData1[Gia_ObjId(pAig,pObj)], pSwitching[Gia_ObjId(pAig,pObj)] );
    printf( "\n" );

    printf( "PO: " );
    Gia_ManForEachPo( pAig, pObj, i )
        printf( "%d=%d (%f)  ", i, p->pData1[Gia_ObjId(pAig,pObj)], pSwitching[Gia_ObjId(pAig,pObj)] );
    printf( "\n" );

    printf( "LI: " );
    Gia_ManForEachRi( pAig, pObj, i )
        printf( "%d=%d (%f)  ", i, p->pData1[Gia_ObjId(pAig,pObj)], pSwitching[Gia_ObjId(pAig,pObj)] );
    printf( "\n" );
*/
    Gia_ManSwiDelete( p );
    return vSwitching;

}
/**Function*************************************************************

  Synopsis    [Computes probability of switching (or of being 1).]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Gia_ManComputeSwitchProbs(pGia *Gia_Man_t, nFrames int32, nPref int32, fProbOne int32) *Vec_Int_t {
    var Pars Gia_ParSwi_t;
    var pPars *Gia_ParSwi_t = &Pars;
    // set the default parameters
    Gia_ManSetDefaultParamsSwi( pPars ); 
    // override some of the defaults
    pPars.nIters   = nFrames;  // set number of total timeframes
    if Abc_FrameReadFlag("seqsimframes") {
        pPars.nIters = strconv.Atoi(Abc_FrameReadFlag("seqsimframes")); }
    pPars.nPref    = nPref;    // set number of first timeframes to skip  
    // decide what should be computed
    if fProbOne != 0 {
        // if the user asked to compute propability of 1, we do not need transition information
        pPars.fProbOne   = 1;  // enable computing probabiblity of being one
        pPars.fProbTrans = 0;  // disable computing transition probability 
    
    } else {
        // if the user asked for transition propabability, we do not need to compute probability of 1
        pPars.fProbOne   = 0;  // disable computing probabiblity of being one
        pPars.fProbTrans = 1;  // enable computing transition probability 
    }
    // perform the computation of switching activity
    return Gia_ManSwiSimulate( pGia, pPars );
}
func Saig_ManComputeSwitchProbs(pAig *Aig_Man_t, nFrames int32, nPref int32, fProbOne int32) *Vec_Int_t {
    var vSwitching, vResult *Vec_Int_t;
    var p *Gia_Man_t;
    var pObj *Aig_Obj_t;
    var i int32;
    // translate AIG into the intermediate form (takes care of choices if present!)
    p = Gia_ManFromAigSwitch( pAig );
    // perform the computation of switching activity
    vSwitching = Gia_ManComputeSwitchProbs( p, nFrames, nPref, fProbOne );
    // transfer the computed result to the original AIG
    vResult = Vec_IntStart( Aig_ManObjNumMax(pAig) );
    // inlining of macro Aig_ManForEachObj
    // inlining of macro Vec_PtrForEachEntry
    for  i = 0; (i < Vec_PtrSize(pAig.vObjs)) && (((pObj) = (Aig_Obj_t *)Vec_PtrEntry(pAig.vObjs, i)), 1); i++  { if (pObj) == nil {} else {
//        if ( Aig_ObjIsCo(pObj) )
//            printf( "%d=%f\n", i, Abc_Int2Float( Vec_IntEntry(vSwitching, Abc_Lit2Var(pObj->iData)) ) );
        Vec_IntWriteEntry( vResult, i, Vec_IntEntry(vSwitching, Abc_Lit2Var(pObj.iData)) );
    }
    }
    // delete intermediate results
    Vec_IntFree( vSwitching );
    Gia_ManStop( p );
    return vResult;
}

/**Function*************************************************************

  Synopsis    [Computes probability of switching (or of being 1).]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Gia_ManEvaluateSwitching(p *Gia_Man_t) float32 {
    var pObj *Gia_Obj_t;
    var SwitchTotal float32  = 0.0;
    var i int32;
    assert(p.pSwitching, "p.pSwitching");
    // inlining of macro ABC_FREE
    p.Refs = nil;
    Gia_ManCreateRefs( p );
    // inlining of macro Gia_ManForEachObj
    for  i = 0; (i < p.nObjs) && ((pObj) = Gia_ManObj(p, i)); i++  {
        SwitchTotal += float32(Gia_ObjRefNum(p, pObj)) * p.pSwitching[i] / 255;
    }
    return SwitchTotal;
}

/**Function*************************************************************

  Synopsis    [Computes probability of switching (or of being 1).]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
/*
float Gia_ManComputeSwitching( Gia_Man_t * p, int nFrames, int nPref, int fProbOne )
{
    Gia_Man_t * pDfs;
    Gia_Obj_t * pObj, * pObjDfs;
    Vec_Int_t * vSwitching;
    float * pSwitching, Switch, SwitchTotal = 0.0;
    int i;
    // derives the DFS ordered AIG
    if ( Gia_ManHasMapping(p) )
        Gia_ManSetRefsMapped(p);
    else
        Gia_ManCreateRefs( p );
//    pDfs = Gia_ManDupOrderDfs( p );
    pDfs = Gia_ManDup( p );
    assert( Gia_ManObjNum(pDfs) == Gia_ManObjNum(p) );
    // perform the computation of switching activity
    vSwitching = Gia_ManComputeSwitchProbs( pDfs, nFrames, nPref, fProbOne );
    // transfer the computed result to the original AIG
    ABC_FREE( p->pSwitching );
    p->pSwitching = ABC_CALLOC( unsigned char, Gia_ManObjNum(p) );
    pSwitching = (float *)vSwitching->pArray;
    Gia_ManForEachObj( p, pObj, i )
    {
        pObjDfs = Gia_ObjFromLit( pDfs, pObj->Value );
        Switch = pSwitching[ Gia_ObjId(pDfs, pObjDfs) ];
        p->pSwitching[i] = (char)((Switch >= 1.0) ? 255 : (int)((0.002 + Switch) * 255)); // 0.00196 = (1/255)/2
        if ( Gia_ObjIsCi(pObj) || (Gia_ObjIsAnd(pObj) && (!Gia_ManHasMapping(p) || Gia_ObjIsLut(p, i))) )
        {
            SwitchTotal += (float)Gia_ObjRefNum(p, pObj) * p->pSwitching[i] / 255;
//            printf( "%d = %.2f\n", i, (float)Gia_ObjRefNum(p, pObj) * p->pSwitching[i] / 255 );            
        }
    }
    Vec_IntFree( vSwitching );
    Gia_ManStop( pDfs );
    return SwitchTotal;
}
*/
/**Function*************************************************************

  Synopsis    [Determine probability of being 1 at the outputs.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Gia_ManPrintOutputProb(p *Gia_Man_t) *Vec_Flt_t {
    var vSimData *Vec_Flt_t;
    var pDfs *Gia_Man_t  = Gia_ManDup( p );
    assert(Gia_ManObjNum(pDfs) == Gia_ManObjNum(p), "Gia_ManObjNum(pDfs) == Gia_ManObjNum(p)");
    vSimData = (Vec_Flt_t *)Gia_ManComputeSwitchProbs( pDfs, (Gia_ManRegNum(p) ? 16 : 1), 0, 1 );
    Gia_ManStop( pDfs );
    return vSimData;
}

////////////////////////////////////////////////////////////////////////
///                       END OF FILE                                ///
////////////////////////////////////////////////////////////////////////


//ABC_NAMESPACE_IMPL_END

