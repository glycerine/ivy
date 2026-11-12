package ssw

import "time"

/**CFile****************************************************************

  FileName    [sswConstr.c]

  SystemName  [ABC: Logic synthesis and verification system.]

  PackageName [Inductive prover with constraints.]

  Synopsis    [One round of SAT sweeping.]

  Author      [Alan Mishchenko]
  
  Affiliation [UC Berkeley]

  Date        [Ver. 1.0. Started - September 1, 2008.]

  Revision    [$Id: sswConstr.c,v 1.00 2008/09/01 00:00:00 alanmi Exp $]

***********************************************************************/

//#include "sswInt.h"
//#include "aig/gia/giaAig.h"

//ABC_NAMESPACE_IMPL_START


////////////////////////////////////////////////////////////////////////
///                        DECLARATIONS                              ///
////////////////////////////////////////////////////////////////////////

////////////////////////////////////////////////////////////////////////
///                     FUNCTION DEFINITIONS                         ///
////////////////////////////////////////////////////////////////////////

/**Function*************************************************************

  Synopsis    [Performs fraiging for one node.]

  Description [Returns the fraiged node.]
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Ssw_ManRefineByFilterSim(p *Ssw_Man_t, nFrames int32) {
    var pObj, pObjLi *Aig_Obj_t;
    var f, i, RetValue1, RetValue2 int32;
    assert(nFrames > 0, "nFrames > 0");
    // assign register outputs
    // inlining of macro Saig_ManForEachLi
    for  i = 0; (i < Saig_ManRegNum(p.pAig)) && (((pObj) = (Aig_Obj_t *)Vec_PtrEntry(p.pAig.vCos, i+Saig_ManPoNum(p.pAig))), 1); i++  {
        pObj.fMarkB = Abc_InfoHasBit( p.pPatWords, Saig_ManPiNum(p.pAig) + i );
    }
    // simulate the timeframes
    for  f = 0; f < nFrames; f++  {
        // set the PI simulation information
        Aig_ManConst1(p.pAig).fMarkB = 1;
        // inlining of macro Saig_ManForEachPi
        // inlining of macro Vec_PtrForEachEntryStop
        for  i = 0; (i < Saig_ManPiNum(p.pAig)) && (((pObj) = (Aig_Obj_t *)Vec_PtrEntry(p.pAig.vCis, i)), 1); i++  {
            pObj.fMarkB = 0;
        }
        // inlining of macro Saig_ManForEachLiLo
        for  i = 0; (i < Saig_ManRegNum(p.pAig)) && (((pObjLi) = Saig_ManLi(p.pAig, i)), 1)            && (((pObj)=Saig_ManLo(p.pAig, i)), 1); i++  {
            pObj.fMarkB = pObjLi.fMarkB;
        }
        // simulate internal nodes
        // inlining of macro Aig_ManForEachNode
        // inlining of macro Vec_PtrForEachEntry
        for  i = 0; (i < Vec_PtrSize(p.pAig.vObjs)) && (((pObj) = (Aig_Obj_t *)Vec_PtrEntry(p.pAig.vObjs, i)), 1); i++  { if (pObj) == nil || !Aig_ObjIsNode(pObj) {} else {
            pObj.fMarkB = ( Aig_ObjFanin0(pObj).fMarkB ^ Aig_ObjFaninC0(pObj) )
                         & ( Aig_ObjFanin1(pObj).fMarkB ^ Aig_ObjFaninC1(pObj) ); }
        }
        // assign the COs
        // inlining of macro Aig_ManForEachCo
        // inlining of macro Vec_PtrForEachEntry
        for  i = 0; (i < Vec_PtrSize(p.pAig.vCos)) && (((pObj) = (Aig_Obj_t *)Vec_PtrEntry(p.pAig.vCos, i)), 1); i++  {
            pObj.fMarkB = ( Aig_ObjFanin0(pObj).fMarkB ^ Aig_ObjFaninC0(pObj) );
        }
        // transfer
        if f == 0 { // copy markB into phase
            // inlining of macro Aig_ManForEachObj
            // inlining of macro Vec_PtrForEachEntry
            for  i = 0; (i < Vec_PtrSize(p.pAig.vObjs)) && (((pObj) = (Aig_Obj_t *)Vec_PtrEntry(p.pAig.vObjs, i)), 1); i++  { if (pObj) == nil {} else {
                pObj.fPhase = pObj.fMarkB; }
            }
        
        } else { // refine classes
            RetValue1 = Ssw_ClassesRefineConst1( p.ppClasses, 0 );
            RetValue2 = Ssw_ClassesRefine( p.ppClasses, 0 );
        }
    }
}


/**Function*************************************************************

  Synopsis    [Performs fraiging for one node.]

  Description [Returns the fraiged node.]

  SideEffects []

  SeeAlso     []

***********************************************************************/
func Ssw_ManRollForward(p *Ssw_Man_t, nFrames int32) {
    var pObj, pObjLi *Aig_Obj_t;
    var f, i int32;
    assert(nFrames > 0, "nFrames > 0");
    // assign register outputs
    // inlining of macro Saig_ManForEachLi
    for  i = 0; (i < Saig_ManRegNum(p.pAig)) && (((pObj) = (Aig_Obj_t *)Vec_PtrEntry(p.pAig.vCos, i+Saig_ManPoNum(p.pAig))), 1); i++  {
        pObj.fMarkB = Abc_InfoHasBit( p.pPatWords, Saig_ManPiNum(p.pAig) + i );
    }
    // simulate the timeframes
    for  f = 0; f < nFrames; f++  {
        // set the PI simulation information
        Aig_ManConst1(p.pAig).fMarkB = 1;
        // inlining of macro Saig_ManForEachPi
        // inlining of macro Vec_PtrForEachEntryStop
        for  i = 0; (i < Saig_ManPiNum(p.pAig)) && (((pObj) = (Aig_Obj_t *)Vec_PtrEntry(p.pAig.vCis, i)), 1); i++  {
            pObj.fMarkB = Aig_ManRandom(0) & 1;
        }
        // inlining of macro Saig_ManForEachLiLo
        for  i = 0; (i < Saig_ManRegNum(p.pAig)) && (((pObjLi) = Saig_ManLi(p.pAig, i)), 1)            && (((pObj)=Saig_ManLo(p.pAig, i)), 1); i++  {
            pObj.fMarkB = pObjLi.fMarkB;
        }
        // simulate internal nodes
        // inlining of macro Aig_ManForEachNode
        // inlining of macro Vec_PtrForEachEntry
        for  i = 0; (i < Vec_PtrSize(p.pAig.vObjs)) && (((pObj) = (Aig_Obj_t *)Vec_PtrEntry(p.pAig.vObjs, i)), 1); i++  { if (pObj) == nil || !Aig_ObjIsNode(pObj) {} else {
            pObj.fMarkB = ( Aig_ObjFanin0(pObj).fMarkB ^ Aig_ObjFaninC0(pObj) )
                         & ( Aig_ObjFanin1(pObj).fMarkB ^ Aig_ObjFaninC1(pObj) ); }
        }
        // assign the COs
        // inlining of macro Aig_ManForEachCo
        // inlining of macro Vec_PtrForEachEntry
        for  i = 0; (i < Vec_PtrSize(p.pAig.vCos)) && (((pObj) = (Aig_Obj_t *)Vec_PtrEntry(p.pAig.vCos, i)), 1); i++  {
            pObj.fMarkB = ( Aig_ObjFanin0(pObj).fMarkB ^ Aig_ObjFaninC0(pObj) );
        }
    }
    // record the new pattern
    // inlining of macro Saig_ManForEachLi
    for  i = 0; (i < Saig_ManRegNum(p.pAig)) && (((pObj) = (Aig_Obj_t *)Vec_PtrEntry(p.pAig.vCos, i+Saig_ManPoNum(p.pAig))), 1); i++  {
        if pObj.fMarkB ^ Abc_InfoHasBit(p.pPatWords, Saig_ManPiNum(p.pAig) + i) {
            Abc_InfoXorBit( p.pPatWords, Saig_ManPiNum(p.pAig) + i ); }
    }
}

/**Function*************************************************************

  Synopsis    [Performs fraiging for one node.]

  Description [Returns the fraiged node.]

  SideEffects []

  SeeAlso     []

***********************************************************************/
func Ssw_ManFindStartingState(p *Ssw_Man_t, pCex *Abc_Cex_t) {
    var pObj, pObjLi *Aig_Obj_t;
    var f, i, iBit int32;
    // assign register outputs
    // inlining of macro Saig_ManForEachLi
    for  i = 0; (i < Saig_ManRegNum(p.pAig)) && (((pObj) = (Aig_Obj_t *)Vec_PtrEntry(p.pAig.vCos, i+Saig_ManPoNum(p.pAig))), 1); i++  {
        pObj.fMarkB = 0;
    }
    // simulate the timeframes
    iBit = pCex.nRegs;
    for  f = 0; f <= pCex.iFrame; f++  {
        // set the PI simulation information
        Aig_ManConst1(p.pAig).fMarkB = 1;
        // inlining of macro Saig_ManForEachPi
        // inlining of macro Vec_PtrForEachEntryStop
        for  i = 0; (i < Saig_ManPiNum(p.pAig)) && (((pObj) = (Aig_Obj_t *)Vec_PtrEntry(p.pAig.vCis, i)), 1); i++  {
            pObj.fMarkB = Abc_InfoHasBit( pCex.pData, iBit++ );
        }
        // inlining of macro Saig_ManForEachLiLo
        for  i = 0; (i < Saig_ManRegNum(p.pAig)) && (((pObjLi) = Saig_ManLi(p.pAig, i)), 1)            && (((pObj)=Saig_ManLo(p.pAig, i)), 1); i++  {
            pObj.fMarkB = pObjLi.fMarkB;
        }
        // simulate internal nodes
        // inlining of macro Aig_ManForEachNode
        // inlining of macro Vec_PtrForEachEntry
        for  i = 0; (i < Vec_PtrSize(p.pAig.vObjs)) && (((pObj) = (Aig_Obj_t *)Vec_PtrEntry(p.pAig.vObjs, i)), 1); i++  { if (pObj) == nil || !Aig_ObjIsNode(pObj) {} else {
            pObj.fMarkB = ( Aig_ObjFanin0(pObj).fMarkB ^ Aig_ObjFaninC0(pObj) )
                         & ( Aig_ObjFanin1(pObj).fMarkB ^ Aig_ObjFaninC1(pObj) ); }
        }
        // assign the COs
        // inlining of macro Aig_ManForEachCo
        // inlining of macro Vec_PtrForEachEntry
        for  i = 0; (i < Vec_PtrSize(p.pAig.vCos)) && (((pObj) = (Aig_Obj_t *)Vec_PtrEntry(p.pAig.vCos, i)), 1); i++  {
            pObj.fMarkB = ( Aig_ObjFanin0(pObj).fMarkB ^ Aig_ObjFaninC0(pObj) );
        }
    }
    assert(iBit == pCex.nBits, "iBit == pCex.nBits");
    // check that the output failed as expected -- cannot check because it is not an SRM!
//    pObj = Aig_ManCo( p->pAig, pCex->iPo );
//    if ( pObj->fMarkB != 1 )
//        Abc_Print( 1, "The counter-example does not refine the output.\n" );
    // record the new pattern
    // inlining of macro Saig_ManForEachLo
    for  i = 0; (i < Saig_ManRegNum(p.pAig)) && (((pObj) = (Aig_Obj_t *)Vec_PtrEntry(p.pAig.vCis, i+Saig_ManPiNum(p.pAig))), 1); i++  {
        if pObj.fMarkB ^ Abc_InfoHasBit(p.pPatWords, Saig_ManPiNum(p.pAig) + i) {
            Abc_InfoXorBit( p.pPatWords, Saig_ManPiNum(p.pAig) + i ); }
    }
}

/**Function*************************************************************

  Synopsis    [Performs fraiging for one node.]

  Description [Returns the fraiged node.]
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Ssw_ManSweepNodeFilter(p *Ssw_Man_t, pObj *Aig_Obj_t, f int32) int32 {
    var pObjRepr, pObjFraig, pObjFraig2, pObjReprFraig *Aig_Obj_t;
    var RetValue int32;
    // get representative of this class
    pObjRepr = Aig_ObjRepr( p.pAig, pObj );
    if pObjRepr == nil {
        return 0; }
    // get the fraiged node
    pObjFraig = Ssw_ObjFrame( p, pObj, f );
    // get the fraiged representative
    pObjReprFraig = Ssw_ObjFrame( p, pObjRepr, f );
    // check if constant 0 pattern distinquishes these nodes
    assert(pObjFraig != nil && pObjReprFraig != nil, "pObjFraig != nil && pObjReprFraig != nil");
    assert((pObj.fPhase == pObjRepr.fPhase) == (Aig_ObjPhaseReal(pObjFraig) == Aig_ObjPhaseReal(pObjReprFraig)), "(pObj.fPhase == pObjRepr.fPhase) == (Aig_ObjPhaseReal(pObjFraig) == Aig_ObjPhaseReal(pObjReprFraig))");
    // if the fraiged nodes are the same, return
    if Aig_Regular(pObjFraig) == Aig_Regular(pObjReprFraig) {
        return 0; }
    // call equivalence checking
    if Aig_Regular(pObjFraig) != Aig_ManConst1(p.pFrames) {
        RetValue = Ssw_NodesAreEquiv( p, Aig_Regular(pObjReprFraig), Aig_Regular(pObjFraig) );
    } else {
        RetValue = Ssw_NodesAreEquiv( p, Aig_Regular(pObjFraig), Aig_Regular(pObjReprFraig) ); }
    if RetValue == 1 {  // proved equivalent
    
        pObjFraig2 = Aig_NotCond( pObjReprFraig, pObj.fPhase ^ pObjRepr.fPhase );
        Ssw_ObjSetFrame( p, pObj, f, pObjFraig2 );
        return 0;
    }
    if RetValue == -1 { // timed out
    
//        Ssw_ClassesRemoveNode( p->ppClasses, pObj );
        return 1;
    }
    // disproved equivalence
    Ssw_SmlSavePatternAig( p, f );
    Ssw_ManResimulateBit( p, pObj, pObjRepr );
    assert(Aig_ObjRepr( p.pAig, pObj ) != pObjRepr, "Aig_ObjRepr( p.pAig, pObj ) != pObjRepr");
    if Aig_ObjRepr( p.pAig, pObj ) == pObjRepr {
        Abc_Print( 1, "Ssw_ManSweepNodeFilter(): Failed to refine representative.\n" );
    }
    return 0;
}

/**Function*************************************************************

  Synopsis    [Performs fraiging for the internal nodes.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Ssw_ManSweepBmcFilter_rec(p *Ssw_Man_t, pObj *Aig_Obj_t, f int32) *Aig_Obj_t {
    var pObjNew, pObjLi *Aig_Obj_t;
    pObjNew = Ssw_ObjFrame( p, pObj, f );
    if pObjNew {
        return pObjNew; }
    assert(!Saig_ObjIsPi(p.pAig, pObj), "!Saig_ObjIsPi(p.pAig, pObj)");
    if Saig_ObjIsLo(p.pAig, pObj) {
        assert(f > 0, "f > 0");
        pObjLi  = Saig_ObjLoToLi( p.pAig, pObj );
        pObjNew = Ssw_ManSweepBmcFilter_rec( p, Aig_ObjFanin0(pObjLi), f-1 );
        pObjNew = Aig_NotCond( pObjNew, Aig_ObjFaninC0(pObjLi) );
    
    } else {
        assert(Aig_ObjIsNode(pObj), "Aig_ObjIsNode(pObj)");
        Ssw_ManSweepBmcFilter_rec( p, Aig_ObjFanin0(pObj), f );
        Ssw_ManSweepBmcFilter_rec( p, Aig_ObjFanin1(pObj), f );
        pObjNew = Aig_And( p.pFrames, Ssw_ObjChild0Fra(p, pObj, f), Ssw_ObjChild1Fra(p, pObj, f) );
    }
    Ssw_ObjSetFrame( p, pObj, f, pObjNew );
    assert(pObjNew != nil, "pObjNew != nil");
    return pObjNew;
}

/**Function*************************************************************

  Synopsis    [Filter equivalence classes of nodes.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Ssw_ManSweepBmcFilter(p *Ssw_Man_t, TimeLimit int32) int32 {
    var pObj, pObjNew, pObjLi, pObjLo *Aig_Obj_t;
    var f, f1, i int32;
    var clkTotal abctime  = Abc_Clock();
    // start initialized timeframes
    p.pFrames = Aig_ManStart( Aig_ManObjNumMax(p.pAig) * p.pPars.nFramesK );
    // inlining of macro Saig_ManForEachLo
    for  i = 0; (i < Saig_ManRegNum(p.pAig)) && (((pObj) = (Aig_Obj_t *)Vec_PtrEntry(p.pAig.vCis, i+Saig_ManPiNum(p.pAig))), 1); i++  {
        if Abc_InfoHasBit( p.pPatWords, Saig_ManPiNum(p.pAig) + i ) {
            Ssw_ObjSetFrame( p, pObj, 0, Aig_ManConst1(p.pFrames) );
//Abc_Print( 1, "1" );
        
        } else {
            Ssw_ObjSetFrame( p, pObj, 0, Aig_ManConst0(p.pFrames) );
//Abc_Print( 1, "0" );
        }
    }
//Abc_Print( 1, "\n" );

    // sweep internal nodes
    for  f = 0; f < p.pPars.nFramesK; f++  {
        // realloc mapping of timeframes
        if f == p.nFrames-1 {
            var pNodeToFrames **Aig_Obj_t;
            // inlining of macro ABC_CALLOC
            pNodeToFrames = ((Aig_Obj_t * *) calloc((size_t)(Aig_ManObjNumMax(p.pAig) * 2 * p.nFrames), sizeof(Aig_Obj_t *)));
            for  f1 = 0; f1 < p.nFrames; f1++  {
                // inlining of macro Aig_ManForEachObj
                // inlining of macro Vec_PtrForEachEntry
                for  i = 0; (i < Vec_PtrSize(p.pAig.vObjs)) && (((pObj) = (Aig_Obj_t *)Vec_PtrEntry(p.pAig.vObjs, i)), 1); i++  { if (pObj) == nil {} else {
                    pNodeToFrames[2*p.nFrames*pObj.Id + f1] = Ssw_ObjFrame( p, pObj, f1 ); }
                }
            }
            p.pNodeToFrames = nil;
            p.pNodeToFrames = pNodeToFrames;
            p.nFrames *= 2;
        }
        // map constants and PIs
        Ssw_ObjSetFrame( p, Aig_ManConst1(p.pAig), f, Aig_ManConst1(p.pFrames) );
        // inlining of macro Saig_ManForEachPi
        // inlining of macro Vec_PtrForEachEntryStop
        for  i = 0; (i < Saig_ManPiNum(p.pAig)) && (((pObj) = (Aig_Obj_t *)Vec_PtrEntry(p.pAig.vCis, i)), 1); i++  {
            pObjNew = Aig_ObjCreateCi(p.pFrames);
            Ssw_ObjSetFrame( p, pObj, f, pObjNew );
        }
        // sweep internal nodes
        // inlining of macro Aig_ManForEachNode
        // inlining of macro Vec_PtrForEachEntry
        for  i = 0; (i < Vec_PtrSize(p.pAig.vObjs)) && (((pObj) = (Aig_Obj_t *)Vec_PtrEntry(p.pAig.vObjs, i)), 1); i++  { if (pObj) == nil || !Aig_ObjIsNode(pObj) {} else {
            pObjNew = Aig_And( p.pFrames, Ssw_ObjChild0Fra(p, pObj, f), Ssw_ObjChild1Fra(p, pObj, f) );
            Ssw_ObjSetFrame( p, pObj, f, pObjNew );
            if Ssw_ManSweepNodeFilter( p, pObj, f ) != 0 {
                break; }
        }
        }
        // printout
        if p.pPars.fVerbose {
            Abc_Print( 1, "Frame %4d : ", f );
            Ssw_ClassesPrint( p.ppClasses, 0 );
        }
        if i < Vec_PtrSize(p.pAig.vObjs) {
            if p.pPars.fVerbose {
                Abc_Print( 1, "Exceeded the resource limits (%d conflicts). Quitting...\n", p.pPars.nBTLimit ); }
            break;
        }
        // quit if this is the last timeframe
        if f == p.pPars.nFramesK - 1 {
            if p.pPars.fVerbose {
                Abc_Print( 1, "Exceeded the time frame limit (%d time frames). Quitting...\n", p.pPars.nFramesK ); }
            break;
        }
        // check timeout
        if TimeLimit && (float32(TimeLimit) <= float32((Abc_Clock()-clkTotal))/float32((int64(time.Second)))) {
            break; }
        // transfer latch input to the latch outputs 
        // inlining of macro Aig_ManForEachCo
        // inlining of macro Vec_PtrForEachEntry
        for  i = 0; (i < Vec_PtrSize(p.pAig.vCos)) && (((pObj) = (Aig_Obj_t *)Vec_PtrEntry(p.pAig.vCos, i)), 1); i++  {
            Ssw_ObjSetFrame( p, pObj, f, Ssw_ObjChild0Fra(p, pObj, f) );
        }
        // build logic cones for register outputs
        // inlining of macro Saig_ManForEachLiLo
        for  i = 0; (i < Saig_ManRegNum(p.pAig)) && (((pObjLi) = Saig_ManLi(p.pAig, i)), 1)            && (((pObjLo)=Saig_ManLo(p.pAig, i)), 1); i++  {
            pObjNew = Ssw_ObjFrame( p, pObjLi, f );
            Ssw_ObjSetFrame( p, pObjLo, f+1, pObjNew );
            Ssw_CnfNodeAddToSolver( p.pMSat, Aig_Regular(pObjNew) );//
        }
    }
    // verify
//    Ssw_ClassesCheck( p->ppClasses );
    return 1;
}

/**Function*************************************************************

  Synopsis    [Filter equivalence classes of nodes.]

  Description [Unrolls at most nFramesMax frames. Works with nConfMax
  conflicts until the first undefined SAT call. Verbose prints the message.]

  SideEffects []

  SeeAlso     []

***********************************************************************/
func Ssw_SignalFilter(pAig *Aig_Man_t, nFramesMax int32, nConfMax int32, nRounds int32, TimeLimit int32, TimeLimit2 int32, pCex *Abc_Cex_t, fLatchOnly int32, fVerbose int32) {
    var Pars Ssw_Pars_t;
    var pPars *Ssw_Pars_t = &Pars;
    var p *Ssw_Man_t;
    var r, TimeLimitPart int32;//, clkTotal = Abc_Clock();
    var nTimeToStop abctime;
    if TimeLimit {
        nTimeToStop = TimeLimit * int64(time.Second) + Abc_Clock();
    } else {
        nTimeToStop = 0;
    }
    assert(Aig_ManRegNum(pAig) > 0, "Aig_ManRegNum(pAig) > 0");
    assert(Aig_ManConstrNum(pAig) == 0, "Aig_ManConstrNum(pAig) == 0");
    // consider the case of empty AIG
    if Aig_ManNodeNum(pAig) == 0 {
        return; }
    // reset random numbers
    Aig_ManRandom( 1 );
    // if parameters are not given, create them
    Ssw_ManSetDefaultParams( pPars = &Pars );
    pPars.nFramesK  = 3; //nFramesMax;
    pPars.nBTLimit  = nConfMax;
    pPars.TimeLimit = TimeLimit;
    pPars.fVerbose  = fVerbose;
    // start the induction manager
    p = Ssw_ManCreate( pAig, pPars );
    pPars.nFramesK  = nFramesMax;
    // create trivial equivalence classes with all nodes being candidates for constant 1
    if pAig.pReprs == nil {
        p.ppClasses = Ssw_ClassesPrepareSimple( pAig, fLatchOnly, 0 );
    } else {
        p.ppClasses = Ssw_ClassesPrepareFromReprs( pAig ); }
    Ssw_ClassesSetData( p.ppClasses, nil, nil, Ssw_SmlObjIsConstBit, Ssw_SmlObjsAreEqualBit );
    assert(p.vInits == nil, "p.vInits == nil");
    // compute starting state if needed
    if pCex != nil {
        Ssw_ManFindStartingState( p, pCex ); }
    // refine classes using BMC
    for  r = 0; r < nRounds; r++  {
        if p.pPars.fVerbose {
            Abc_Print( 1, "Round %3d:\n", r ); }
        // start filtering equivalence classes
        Ssw_ManRefineByFilterSim( p, p.pPars.nFramesK );
        if Ssw_ClassesCand1Num(p.ppClasses) == 0 && Ssw_ClassesClassNum(p.ppClasses) == 0 {
            Abc_Print( 1, "All equivalences are refined away.\n" );
            break;
        }
        // printout
        if p.pPars.fVerbose {
            Abc_Print( 1, "Initial    : " );
            Ssw_ClassesPrint( p.ppClasses, 0 );
        }
        p.pMSat = Ssw_SatStart( 0 );
        if TimeLimit {
            TimeLimitPart = (nTimeToStop - Abc_Clock()) / int64(time.Second);
        } else {
            TimeLimitPart = 0;
        }
        if TimeLimit2 != 0 {
            if TimeLimitPart != 0 {
                TimeLimitPart = Abc_MinInt( TimeLimitPart, TimeLimit2 );
            } else {
                TimeLimitPart = TimeLimit2; }
        }
        Ssw_ManSweepBmcFilter( p, TimeLimitPart );
        Ssw_SatStop( p.pMSat );
        p.pMSat = nil;
        Ssw_ManCleanup( p );
        // simulate pattern forward
        Ssw_ManRollForward( p, p.pPars.nFramesK );
        // check timeout
        if TimeLimit && Abc_Clock() > nTimeToStop {
            Abc_Print( 1, "Reached timeout (%d seconds).\n",  TimeLimit );
            break;
        }
    }
    // cleanup
    Aig_ManSetPhase( p.pAig );
    Aig_ManCleanMarkB( p.pAig );
    // cleanup
    pPars.fVerbose = 0;
    Ssw_ManStop( p );
}

/**Function*************************************************************

  Synopsis    [Filter equivalence classes of nodes.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Ssw_SignalFilterGia(p *Gia_Man_t, nFramesMax int32, nConfMax int32, nRounds int32, TimeLimit int32, TimeLimit2 int32, pCex *Abc_Cex_t, fLatchOnly int32, fVerbose int32) {
    var pAig *Aig_Man_t;
    pAig = Gia_ManToAigSimple( p );
    if p.pReprs != nil {
        Gia_ManReprToAigRepr2( pAig, p );
        p.pReprs = nil;
        p.pNexts = nil;
    }
    Ssw_SignalFilter( pAig, nFramesMax, nConfMax, nRounds, TimeLimit, TimeLimit2, pCex, fLatchOnly, fVerbose );
    Gia_ManReprFromAigRepr( pAig, p );
    Aig_ManStop( pAig );
}

////////////////////////////////////////////////////////////////////////
///                       END OF FILE                                ///
////////////////////////////////////////////////////////////////////////


//ABC_NAMESPACE_IMPL_END
