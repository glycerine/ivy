package ssw

import "time"

/**CFile****************************************************************

  FileName    [sswRarity.c]

  SystemName  [ABC: Logic synthesis and verification system.]

  PackageName [Inductive prover with constraints.]

  Synopsis    [Rarity-driven refinement of equivalence classes.]

  Author      [Alan Mishchenko]
  
  Affiliation [UC Berkeley]

  Date        [Ver. 1.0. Started - September 1, 2008.]

  Revision    [$Id: sswRarity.c,v 1.00 2008/09/01 00:00:00 alanmi Exp $]

***********************************************************************/

//#include "sswInt.h"
//#include "aig/gia/giaAig.h"

//ABC_NAMESPACE_IMPL_START


////////////////////////////////////////////////////////////////////////
///                        DECLARATIONS                              ///
////////////////////////////////////////////////////////////////////////

   
type sswRar2Man struct {
    // parameters
    nWords            int32;       // the number of words to simulate
    nFrames            int32;      // the number of frames to simulate
    nBinSize            int32;     // the number of flops in one group
    fVerbose            int32;     // the verbosiness flag
    nGroups            int32;      // the number of flop groups
    // internal data
    pAig     *Aig_Man_t;         // AIG with equivalence classes
    ppClasses     *Ssw_Cla_t;    // equivalence classes
    pSml     *Ssw_Sml_t;         // simulation manager
    vSimInfo     *Vec_Ptr_t;     // simulation info from pSml manager
    vInits     *Vec_Int_t;       // initial state
    // rarity data
    pRarity           *int32;      // occur counts for patterns in groups
    pGroupValues           *int32; // occur counts in each group
    pPatCosts        *float64;    // pattern costs

}

func sswRar2GetBinPat(p *sswRar2Man, iBin int32, iPat int32) int32 {
    assert(iBin >= 0 && iBin < Aig_ManRegNum(p.pAig) / p.nBinSize, "iBin >= 0 && iBin < Aig_ManRegNum(p.pAig) / p.nBinSize");
    assert(iPat >= 0 && iPat < (1 << p.nBinSize), "iPat >= 0 && iPat < (1 << p.nBinSize)");
    return p.pRarity[iBin * (1 << p.nBinSize) + iPat];
}
func sswRar2SetBinPat(p *sswRar2Man, iBin int32, iPat int32, Value int32) {
    assert(iBin >= 0 && iBin < Aig_ManRegNum(p.pAig) / p.nBinSize, "iBin >= 0 && iBin < Aig_ManRegNum(p.pAig) / p.nBinSize");
    assert(iPat >= 0 && iPat < (1 << p.nBinSize), "iPat >= 0 && iPat < (1 << p.nBinSize)");
    p.pRarity[iBin * (1 << p.nBinSize) + iPat] = Value;
}
func sswRar2AddToBinPat(p *sswRar2Man, iBin int32, iPat int32) {
    assert(iBin >= 0 && iBin < Aig_ManRegNum(p.pAig) / p.nBinSize, "iBin >= 0 && iBin < Aig_ManRegNum(p.pAig) / p.nBinSize");
    assert(iPat >= 0 && iPat < (1 << p.nBinSize), "iPat >= 0 && iPat < (1 << p.nBinSize)");
    p.pRarity[iBin * (1 << p.nBinSize) + iPat]++;
}

////////////////////////////////////////////////////////////////////////
///                     FUNCTION DEFINITIONS                         ///
////////////////////////////////////////////////////////////////////////

/**Function*************************************************************

  Synopsis    []

  Description []

  SideEffects []

  SeeAlso     []

***********************************************************************/
func sswRar2ManStart(pAig *Aig_Man_t, nWords int32, nFrames int32, nBinSize int32, fVerbose int32) *sswRar2Man {
    var p *sswRar2Man;
    if Aig_ManRegNum(pAig) < nBinSize || nBinSize <= 0 {
        return nil; }
    // inlining of macro ABC_CALLOC
    p = &sswRar2Man{};
    p.pAig = pAig;
    p.nWords = nWords;
    p.nFrames = nFrames;
    p.nBinSize = nBinSize;
    p.fVerbose = fVerbose;
    p.nGroups  = Aig_ManRegNum(pAig) / nBinSize;
    // inlining of macro ABC_CALLOC
    p.pRarity  = ((int *) calloc((size_t)((1 << nBinSize) * p.nGroups), sizeof(int)));
    // inlining of macro ABC_CALLOC
    p.pGroupValues = ((int *) calloc((size_t)(p.nGroups), sizeof(int)));
    // inlining of macro ABC_CALLOC
    p.pPatCosts = ((double *) calloc((size_t)(p.nWords * 32), sizeof(double)));
    p.pSml = Ssw_SmlStart( pAig, 0, nFrames, nWords );
    p.vSimInfo = Ssw_SmlSimDataPointers( p.pSml );
    return p;
}

/**Function*************************************************************

  Synopsis    []

  Description []

  SideEffects []

  SeeAlso     []

***********************************************************************/
func sswRar2ManStop(p *sswRar2Man) {
    if p.pSml { Ssw_SmlStop( p.pSml ); }
    if p.ppClasses { Ssw_ClassesStop( p.ppClasses ); }
    Vec_PtrFreeP( &p.vSimInfo );
    Vec_IntFreeP( &p.vInits );
    p.pGroupValues = nil;
    p.pPatCosts = nil;
    p.pRarity = nil;
    p = nil;
}


/**Function*************************************************************

  Synopsis    [Updates rarity counters.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Ssw_RarUpdateCounters(p *sswRar2Man) {
    var pObj *Aig_Obj_t;
    var pData *uint32;
    var i, k int32;
/*
    Saig_ManForEachLi( p->pAig, pObj, i )
    {
        pData = (unsigned *)Vec_PtrEntry( p->vSimInfo, Aig_ObjId(pObj) ) + p->nWords * (p->nFrames - 1);
        Extra_PrintBinary( stdout, pData, 32 );  Abc_Print( 1, "\n" );
    }
*/
    for  k = 0; k < p.nWords * 32; k++  {
        for  i = 0; i < p.nGroups; i++  {
            p.pGroupValues[i] = 0;
        }
        // inlining of macro Saig_ManForEachLi
        for  i = 0; (i < Saig_ManRegNum(p.pAig)) && (((pObj) = (Aig_Obj_t *)Vec_PtrEntry(p.pAig.vCos, i+Saig_ManPoNum(p.pAig))), 1); i++  {
            pData = (unsigned *)Vec_PtrEntry( p.vSimInfo, Aig_ObjId(pObj) ) + p.nWords * (p.nFrames - 1);
            if Abc_InfoHasBit(pData, k) && i / p.nBinSize < p.nGroups {
                p.pGroupValues[i / p.nBinSize] |= (1 << (i % p.nBinSize)); }
        }
        for  i = 0; i < p.nGroups; i++  {
            sswRar2AddToBinPat( p, i, p.pGroupValues[i] );
        }
    }
/*
    for ( i = 0; i < p->nGroups; i++ )
    {
        for ( k = 0; k < (1 << p->nBinSize); k++ )
            Abc_Print( 1, "%d ", sswRar2GetBinPat(p, i, k) );
        Abc_Print( 1, "\n" );
    }
*/
}

/**Function*************************************************************

  Synopsis    [Select best patterns.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func sswRar2TransferPatterns(p *sswRar2Man, vInits *Vec_Int_t) {
    var pObj *Aig_Obj_t;
    var pData *uint32;
    var i, k, Value int32;

    // for each pattern
    for  k = 0; k < p.nWords * 32; k++  {
        for  i = 0; i < p.nGroups; i++  {
            p.pGroupValues[i] = 0;
        }
        // compute its group values
        // inlining of macro Saig_ManForEachLi
        for  i = 0; (i < Saig_ManRegNum(p.pAig)) && (((pObj) = (Aig_Obj_t *)Vec_PtrEntry(p.pAig.vCos, i+Saig_ManPoNum(p.pAig))), 1); i++  {
            pData = (unsigned *)Vec_PtrEntry( p.vSimInfo, Aig_ObjId(pObj) ) + p.nWords * (p.nFrames - 1);
            if Abc_InfoHasBit(pData, k) && i / p.nBinSize < p.nGroups {
                p.pGroupValues[i / p.nBinSize] |= (1 << (i % p.nBinSize)); }
        }
        // find the cost of its values
        p.pPatCosts[k] = 0.0;
        for  i = 0; i < p.nGroups; i++  {
            Value = sswRar2GetBinPat( p, i, p.pGroupValues[i] );
            assert(Value > 0, "Value > 0");
            p.pPatCosts[k] += 1.0/(Value*Value);
        }
        // print the result
//        Abc_Print( 1, "%3d : %9.6f\n", k, p->pPatCosts[k] );
    }

    // choose as many as there are words
    Vec_IntClear( vInits );
    for  i = 0; i < p.nWords; i++  {
        // select the best
        var iPatBest int32  = -1;
        var iCostBest float64  = -ABC_INFINITY;
        for  k = 0; k < p.nWords * 32; k++  {
            if iCostBest < p.pPatCosts[k] {
                iCostBest = p.pPatCosts[k];
                iPatBest  = k;
            }
        }
        // remove from costs
        assert(iPatBest >= 0, "iPatBest >= 0");
        p.pPatCosts[iPatBest] = -ABC_INFINITY;
        // set the flops
        // inlining of macro Saig_ManForEachLi
        for  k = 0; (k < Saig_ManRegNum(p.pAig)) && (((pObj) = (Aig_Obj_t *)Vec_PtrEntry(p.pAig.vCos, k+Saig_ManPoNum(p.pAig))), 1); k++  {
            pData = (unsigned *)Vec_PtrEntry( p.vSimInfo, Aig_ObjId(pObj) ) + p.nWords * (p.nFrames - 1);
            Vec_IntPush( vInits, Abc_InfoHasBit(pData, iPatBest) );
        }
//Abc_Print( 1, "Best pattern %5d\n", iPatBest );
    }
    assert(Vec_IntSize(vInits) == Aig_ManRegNum(p.pAig) * p.nWords, "Vec_IntSize(vInits) == Aig_ManRegNum(p.pAig) * p.nWords");
}


/**Function*************************************************************

  Synopsis    [Performs fraiging for one node.]

  Description [Returns the fraiged node.]
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func sswRar2FindStartingState(pAig *Aig_Man_t, pCex *Abc_Cex_t) *Vec_Int_t {
    var vInit *Vec_Int_t;
    var pObj, pObjLi *Aig_Obj_t;
    var f, i, iBit int32;
    // assign register outputs
    // inlining of macro Saig_ManForEachLi
    for  i = 0; (i < Saig_ManRegNum(pAig)) && (((pObj) = (Aig_Obj_t *)Vec_PtrEntry(pAig.vCos, i+Saig_ManPoNum(pAig))), 1); i++  {
        pObj.fMarkB = Abc_InfoHasBit( pCex.pData, i );
    }
    // simulate the timeframes
    iBit = pCex.nRegs;
    for  f = 0; f <= pCex.iFrame; f++  {
        // set the PI simulation information
        Aig_ManConst1(pAig).fMarkB = 1;
        // inlining of macro Saig_ManForEachPi
        // inlining of macro Vec_PtrForEachEntryStop
        for  i = 0; (i < Saig_ManPiNum(pAig)) && (((pObj) = (Aig_Obj_t *)Vec_PtrEntry(pAig.vCis, i)), 1); i++  {
            pObj.fMarkB = Abc_InfoHasBit( pCex.pData, iBit++ );
        }
        // inlining of macro Saig_ManForEachLiLo
        for  i = 0; (i < Saig_ManRegNum(pAig)) && (((pObjLi) = Saig_ManLi(pAig, i)), 1)            && (((pObj)=Saig_ManLo(pAig, i)), 1); i++  {
            pObj.fMarkB = pObjLi.fMarkB;
        }
        // simulate internal nodes
        // inlining of macro Aig_ManForEachNode
        // inlining of macro Vec_PtrForEachEntry
        for  i = 0; (i < Vec_PtrSize(pAig.vObjs)) && (((pObj) = (Aig_Obj_t *)Vec_PtrEntry(pAig.vObjs, i)), 1); i++  { if (pObj) == nil || !Aig_ObjIsNode(pObj) {} else {
            pObj.fMarkB = ( Aig_ObjFanin0(pObj).fMarkB ^ Aig_ObjFaninC0(pObj) )
                         & ( Aig_ObjFanin1(pObj).fMarkB ^ Aig_ObjFaninC1(pObj) ); }
        }
        // assign the COs
        // inlining of macro Aig_ManForEachCo
        // inlining of macro Vec_PtrForEachEntry
        for  i = 0; (i < Vec_PtrSize(pAig.vCos)) && (((pObj) = (Aig_Obj_t *)Vec_PtrEntry(pAig.vCos, i)), 1); i++  {
            pObj.fMarkB = ( Aig_ObjFanin0(pObj).fMarkB ^ Aig_ObjFaninC0(pObj) );
        }
    }
    assert(iBit == pCex.nBits, "iBit == pCex.nBits");
    // check that the output failed as expected -- cannot check because it is not an SRM!
//    pObj = Aig_ManCo( pAig, pCex->iPo );
//    if ( pObj->fMarkB != 1 )
//        Abc_Print( 1, "The counter-example does not refine the output.\n" );
    // record the new pattern
    vInit = Vec_IntAlloc( Saig_ManRegNum(pAig) );
    // inlining of macro Saig_ManForEachLo
    for  i = 0; (i < Saig_ManRegNum(pAig)) && (((pObj) = (Aig_Obj_t *)Vec_PtrEntry(pAig.vCis, i+Saig_ManPiNum(pAig))), 1); i++  {
        Vec_IntPush( vInit, pObj.fMarkB );
    }
    return vInit;
}


/**Function*************************************************************

  Synopsis    [Perform sequential simulation.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Ssw_RarSimulate2(pAig *Aig_Man_t, nFrames int32, nWords int32, nBinSize int32, nRounds int32, TimeOut int32, fVerbose int32) int32 {
    var fMiter int32  = 1;
    var p *sswRar2Man;
    var r int32;
    var clk abctime;
    var clkTotal abctime = Abc_Clock();
    var nTimeToStop abctime;
    if TimeOut {
        nTimeToStop = TimeOut * int64(time.Second) + Abc_Clock();
    } else {
        nTimeToStop = 0;
    }
    var RetValue int32  = -1;
    assert(Aig_ManRegNum(pAig) > 0, "Aig_ManRegNum(pAig) > 0");
    assert(Aig_ManConstrNum(pAig) == 0, "Aig_ManConstrNum(pAig) == 0");
    // consider the case of empty AIG
    if Aig_ManNodeNum(pAig) == 0 {
        return -1; }
    if fVerbose != 0 {
        Abc_Print( 1, "Simulating %d words through %d frames with %d binsize, %d rounds, and %d sec timeout.\n",
            nWords, nFrames, nBinSize, nRounds, TimeOut ); }
    // reset random numbers
    Aig_ManRandom( 1 );

    // create manager
    p = sswRar2ManStart( pAig, nWords, nFrames, nBinSize, fVerbose );
    p.vInits = Vec_IntStart( Aig_ManRegNum(pAig) * nWords );
    Ssw_SmlInitializeSpecial( p.pSml, p.vInits );

    // perform simulation rounds
    for  r = 0; r < nRounds; r++  {
        clk = Abc_Clock();
        // simulate
        Ssw_SmlSimulateOne( p.pSml );
        if (fMiter != 0) && Ssw_SmlCheckNonConstOutputs(p.pSml) {
            if fVerbose != 0 { Abc_Print( 1, "\n" ); }
            Abc_Print( 1, "Simulation asserted a PO in frame f: %d <= f < %d.\n", r * nFrames, (r+1) * nFrames );
            RetValue = 0;
            break;
        }
        // get initialization patterns
        Ssw_RarUpdateCounters( p );
        sswRar2TransferPatterns( p, p.vInits );
        Ssw_SmlInitializeSpecial( p.pSml, p.vInits );
        // printout
        if fVerbose != 0 {
//            Abc_Print( 1, "Round %3d:  ", r );
//            Abc_PrintTime( 1, "Time", Abc_Clock() - clk );
            Abc_Print( 1, "." );
        }
        // check timeout
        if TimeOut && Abc_Clock() > nTimeToStop {
            if fVerbose != 0 { Abc_Print( 1, "\n" ); }
            Abc_Print( 1, "Reached timeout (%d seconds).\n",  TimeOut );
            break;
        }
    }
    if r == nRounds {
        if fVerbose != 0 { Abc_Print( 1, "\n" ); }
        Abc_Print( 1, "Simulation did not assert POs in the first %d frames.  ", nRounds * nFrames );
        Abc_PrintTime( 1, "Time", Abc_Clock() - clkTotal );
    }
    // cleanup
    sswRar2ManStop( p );
    return RetValue;
}


/**Function*************************************************************

  Synopsis    [Filter equivalence classes of nodes.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Ssw_RarSignalFilter2(pAig *Aig_Man_t, nFrames int32, nWords int32, nBinSize int32, nRounds int32, TimeOut int32, pCex *Abc_Cex_t, fLatchOnly int32, fVerbose int32) int32 {
    var fMiter int32;
    var p *sswRar2Man;
    var r, i, k int32;
    var clkTotal abctime  = Abc_Clock();
    var nTimeToStop abctime;
    if TimeOut {
        nTimeToStop = TimeOut * int64(time.Second) + Abc_Clock();
    } else {
        nTimeToStop = 0;
    }
    var RetValue int32  = -1;
    assert(Aig_ManRegNum(pAig) > 0, "Aig_ManRegNum(pAig) > 0");
    assert(Aig_ManConstrNum(pAig) == 0, "Aig_ManConstrNum(pAig) == 0");
    // consider the case of empty AIG
    if Aig_ManNodeNum(pAig) == 0 {
        return -1; }
    if fVerbose != 0 {
        Abc_Print( 1, "Filtering equivs with %d words through %d frames with %d binsize, %d rounds, and %d sec timeout.\n",
            nWords, nFrames, nBinSize, nRounds, TimeOut ); }
    // reset random numbers
    Aig_ManRandom( 1 );

    // create manager
    p = sswRar2ManStart( pAig, nWords, nFrames, nBinSize, fVerbose );
    // compute starting state if needed
    assert(p.vInits == nil, "p.vInits == nil");
    if pCex != nil {
        p.vInits = sswRar2FindStartingState( pAig, pCex );
    } else {
        p.vInits = Vec_IntStart( Aig_ManRegNum(pAig) ); }
    // duplicate the array
    for  i = 1; i < nWords; i++  {
        for  k = 0; k < Aig_ManRegNum(pAig); k++  {
            Vec_IntPush( p.vInits, Vec_IntEntry(p.vInits, k) );
        }
    }
    assert(Vec_IntSize(p.vInits) == Aig_ManRegNum(pAig) * nWords, "Vec_IntSize(p.vInits) == Aig_ManRegNum(pAig) * nWords");
    // initialize simulation manager
    Ssw_SmlInitializeSpecial( p.pSml, p.vInits );

    // create trivial equivalence classes with all nodes being candidates for constant 1
    if pAig.pReprs == nil {
        p.ppClasses = Ssw_ClassesPrepareSimple( pAig, fLatchOnly, 0 );
    } else {
        p.ppClasses = Ssw_ClassesPrepareFromReprs( pAig ); }
    Ssw_ClassesSetData( p.ppClasses, p.pSml, nil, (int(*)(void *,Aig_Obj_t *))Ssw_SmlObjIsConstWord, (int(*)(void *,Aig_Obj_t *,Aig_Obj_t *))Ssw_SmlObjsAreEqualWord );
    // print the stats
    if fVerbose != 0 {
        Abc_Print( 1, "Initial  :  " );
        Ssw_ClassesPrint( p.ppClasses, 0 );
    }
    // refine classes using BMC
    for  r = 0; r < nRounds; r++  {
        // start filtering equivalence classes
        if Ssw_ClassesCand1Num(p.ppClasses) == 0 && Ssw_ClassesClassNum(p.ppClasses) == 0 {
            Abc_Print( 1, "All equivalences are refined away.\n" );
            break;
        }
        // simulate
        Ssw_SmlSimulateOne( p.pSml );
        if (fMiter != 0) && Ssw_SmlCheckNonConstOutputs(p.pSml) {
            if fVerbose != 0 { Abc_Print( 1, "\n" ); }
            Abc_Print( 1, "Simulation asserted a PO in frame f: %d <= f < %d.\n", r * nFrames, (r+1) * nFrames );
            RetValue = 0;
            break;
        }
        // check equivalence classes
        Ssw_ClassesRefineConst1( p.ppClasses, 1 );
        Ssw_ClassesRefine( p.ppClasses, 1 );
        // printout
        if fVerbose != 0 {
            Abc_Print( 1, "Round %3d:  ", r );
            Ssw_ClassesPrint( p.ppClasses, 0 );
        }
        // get initialization patterns
        Ssw_RarUpdateCounters( p );
        sswRar2TransferPatterns( p, p.vInits );
        Ssw_SmlInitializeSpecial( p.pSml, p.vInits );
        // check timeout
        if TimeOut && Abc_Clock() > nTimeToStop {
            if fVerbose != 0 { Abc_Print( 1, "\n" ); }
            Abc_Print( 1, "Reached timeout (%d seconds).\n",  TimeOut );
            break;
        }
    }
    if r == nRounds {
        Abc_Print( 1, "Simulation did not assert POs in the first %d frames.  ", nRounds * nFrames );
        Abc_PrintTime( 1, "Time", Abc_Clock() - clkTotal );
    }
    // cleanup
    sswRar2ManStop( p );
    return -1;
}

/**Function*************************************************************

  Synopsis    [Filter equivalence classes of nodes.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Ssw_RarSignalFilterGia2(p *Gia_Man_t, nFrames int32, nWords int32, nBinSize int32, nRounds int32, TimeOut int32, pCex *Abc_Cex_t, fLatchOnly int32, fVerbose int32) int32 {
    var pAig *Aig_Man_t;
    var RetValue int32;
    pAig = Gia_ManToAigSimple( p );
    if p.pReprs != nil {
        Gia_ManReprToAigRepr2( pAig, p );
        p.pReprs = nil;
        p.pNexts = nil;
    }
    RetValue = Ssw_RarSignalFilter2( pAig, nFrames, nWords, nBinSize, nRounds, TimeOut, pCex, fLatchOnly, fVerbose );
    Gia_ManReprFromAigRepr( pAig, p );
    Aig_ManStop( pAig );
    return RetValue;
}




////////////////////////////////////////////////////////////////////////
///                       END OF FILE                                ///
////////////////////////////////////////////////////////////////////////


//ABC_NAMESPACE_IMPL_END
