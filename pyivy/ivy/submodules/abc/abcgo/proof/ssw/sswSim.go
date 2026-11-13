package ssw

/**CFile****************************************************************

  FileName    [sswSim.c]

  SystemName  [ABC: Logic synthesis and verification system.]

  PackageName [Inductive prover with constraints.]

  Synopsis    [Sequential simulator used by the inductive prover.]

  Author      [Alan Mishchenko]
  
  Affiliation [UC Berkeley]

  Date        [Ver. 1.0. Started - September 1, 2008.]

  Revision    [$Id: sswSim.c,v 1.00 2008/09/01 00:00:00 alanmi Exp $]

***********************************************************************/

//#include "sswInt.h"

//ABC_NAMESPACE_IMPL_START


////////////////////////////////////////////////////////////////////////
///                        DECLARATIONS                              ///
////////////////////////////////////////////////////////////////////////

// simulation manager





////////////////////////////////////////////////////////////////////////
///                     FUNCTION DEFINITIONS                         ///
////////////////////////////////////////////////////////////////////////

/**Function*************************************************************

  Synopsis    [Returns 1 if the node appears to be constant 1 candidate.]

  Description []

  SideEffects []

  SeeAlso     []

***********************************************************************/
func Ssw_SmlObjIsConstBit(p any, pObj *Aig_Obj_t) int32 {
    return pObj.fPhase == pObj.fMarkB;
}

/**Function*************************************************************

  Synopsis    [Returns 1 if the nodes appear equal.]

  Description []

  SideEffects []

  SeeAlso     []

***********************************************************************/
func Ssw_SmlObjsAreEqualBit(p any, pObj0 *Aig_Obj_t, pObj1 *Aig_Obj_t) int32 {
    return (pObj0.fPhase == pObj1.fPhase) == (pObj0.fMarkB == pObj1.fMarkB);
}


/**Function*************************************************************

  Synopsis    [Counts the number of 1s in the XOR of simulation data.]

  Description []

  SideEffects []

  SeeAlso     []

***********************************************************************/


/**Function*************************************************************

  Synopsis    [Checks implication.]

  Description []

  SideEffects []

  SeeAlso     []

***********************************************************************/
func Ssw_SmlCheckXorImplication(p *Ssw_Sml_t, pObjLi *Aig_Obj_t, pObjLo *Aig_Obj_t, pCand *Aig_Obj_t) int32 {
    var pSimLi, pSimLo, pSimCand *uint32;
    var k int32;
    assert(pObjLo.fPhase == 0, "pObjLo.fPhase == 0");
    // pObjLi->fPhase may be 1, but the LI simulation data is not complemented!
    pSimCand = Ssw_ObjSim( p, Aig_Regular(pCand).Id );
    pSimLi   = Ssw_ObjSim( p, pObjLi.Id );
    pSimLo   = Ssw_ObjSim( p, pObjLo.Id );
    if Aig_Regular(pCand).fPhase ^ Aig_IsComplement(pCand) {
        for  k = p.nWordsPref; k < p.nWordsTotal; k++  {
            if ~pSimCand[k] & (pSimLi[k] ^ pSimLo[k]) {
                return 0; }
        }
    
    } else {
        for  k = p.nWordsPref; k < p.nWordsTotal; k++  {
            if pSimCand[k] & (pSimLi[k] ^ pSimLo[k]) {
                return 0; }
        }
    }
    return 1;
}

/**Function*************************************************************

  Synopsis    [Counts the number of 1s in the implication.]

  Description []

  SideEffects []

  SeeAlso     []

***********************************************************************/
func Ssw_SmlCountXorImplication(p *Ssw_Sml_t, pObjLi *Aig_Obj_t, pObjLo *Aig_Obj_t, pCand *Aig_Obj_t) int32 {
    var pSimLi, pSimLo, pSimCand *uint32;
    var k, Counter int32;
    assert(pObjLo.fPhase == 0, "pObjLo.fPhase == 0");
    // pObjLi->fPhase may be 1, but the LI simulation data is not complemented!
    pSimCand = Ssw_ObjSim( p, Aig_Regular(pCand).Id );
    pSimLi   = Ssw_ObjSim( p, pObjLi.Id );
    pSimLo   = Ssw_ObjSim( p, pObjLo.Id );
    if Aig_Regular(pCand).fPhase ^ Aig_IsComplement(pCand) {
        for  k = p.nWordsPref; k < p.nWordsTotal; k++  {
            Counter += Aig_WordCountOnes(~pSimCand[k] & ~(pSimLi[k] ^ pSimLo[k]));
        }
    
    } else {
        for  k = p.nWordsPref; k < p.nWordsTotal; k++  {
            Counter += Aig_WordCountOnes(pSimCand[k] & ~(pSimLi[k] ^ pSimLo[k]));
        }
    }
    return Counter;
}

/**Function*************************************************************

  Synopsis    [Counts the number of 1s in the implication.]

  Description []

  SideEffects []

  SeeAlso     []

***********************************************************************/


/**Function*************************************************************

  Synopsis    [Returns 1 if simulation info is composed of all zeros.]

  Description []

  SideEffects []

  SeeAlso     []

***********************************************************************/


/**Function*************************************************************

  Synopsis    [Returns 1 if simulation info is composed of all zeros.]

  Description []

  SideEffects []

  SeeAlso     []

***********************************************************************/


/**Function*************************************************************

  Synopsis    [Counts the number of one's in the pattern of the object.]

  Description []

  SideEffects []

  SeeAlso     []

***********************************************************************/


/**Function*************************************************************

  Synopsis    [Counts the number of one's in the pattern of the objects.]

  Description []

  SideEffects []

  SeeAlso     []

***********************************************************************/
func Ssw_SmlNodeCountOnesRealVec(p *Ssw_Sml_t, vObjs *Vec_Ptr_t) int32 {
    var pObj *Aig_Obj_t;
    var pSims *uint32;
    var uWord uint32;
    var i, k, Counter int32;
    if Vec_PtrSize(vObjs) == 0 {
        return 0; }
    for  i = 0; i < p.nWordsTotal; i++  {
        uWord = 0;
        // inlining of macro Vec_PtrForEachEntry
        for  k = 0; (k < Vec_PtrSize(vObjs)) && (((pObj) = (Aig_Obj_t *)Vec_PtrEntry(vObjs, k)), 1); k++  {
            pSims = Ssw_ObjSim(p, Aig_Regular(pObj).Id);
            if Aig_Regular(pObj).fPhase ^ Aig_IsComplement(pObj) {
                uWord |= ~pSims[i];
            } else {
                uWord |= pSims[i]; }
        }
        Counter += Aig_WordCountOnes( uWord );
    }
    return Counter;
}



/**Function*************************************************************

  Synopsis    [Generated const 0 pattern.]

  Description []

  SideEffects []

  SeeAlso     []

***********************************************************************/
func Ssw_SmlSavePattern0(p *Ssw_Man_t, fInit int32) {
    memset( p.pPatWords, 0, sizeof(unsigned) * p.nPatWords );
}

/**Function*************************************************************

  Synopsis    [[Generated const 1 pattern.]

  Description []

  SideEffects []

  SeeAlso     []

***********************************************************************/
func Ssw_SmlSavePattern1(p *Ssw_Man_t, fInit int32) {
    var pObj *Aig_Obj_t;
    var i, k, nTruePis int32;
    memset( p.pPatWords, 0xff, sizeof(unsigned) * p.nPatWords );
    if !(fInit != 0) {
        return; }
    // clear the state bits to correspond to all-0 initial state
    nTruePis = Saig_ManPiNum(p.pAig);
    k = 0;
    // inlining of macro Saig_ManForEachLo
    for  i = 0; (i < Saig_ManRegNum(p.pAig)) && (((pObj) = (Aig_Obj_t *)Vec_PtrEntry(p.pAig.vCis, i+Saig_ManPiNum(p.pAig))), 1); i++  {
        Abc_InfoXorBit( p.pPatWords, nTruePis * p.nFrames + k++ );
    }
}


/**Function*************************************************************

  Synopsis    [Creates the counter-example from the successful pattern.]

  Description []

  SideEffects []

  SeeAlso     []

***********************************************************************/
func Ssw_SmlCheckOutputSavePattern(p *Ssw_Sml_t, pObjPo *Aig_Obj_t) *int32 {
    var pFanin, pObjPi *Aig_Obj_t;
    var pSims *uint32;
    var i int32;
    var k int32;
    var BestPat int32;
    var pModel *int32;
    // find the word of the pattern
    pFanin = Aig_ObjFanin0(pObjPo);
    pSims = Ssw_ObjSim(p, pFanin.Id);
    for  i = 0; i < p.nWordsTotal; i++  {
        if pSims[i] {
            break; }
    }
    assert(i < p.nWordsTotal, "i < p.nWordsTotal");
    // find the bit of the pattern
    for  k = 0; k < 32; k++  {
        if pSims[i] & (1 << k) {
            break; }
    }
    assert(k < 32, "k < 32");
    // determine the best pattern
    BestPat = i * 32 + k;
    // fill in the counter-example data
    // inlining of macro ABC_ALLOC
    pModel = ((int *) malloc(sizeof(int) * (size_t)(Aig_ManCiNum(p.pAig)+1)));
    // inlining of macro Aig_ManForEachCi
    // inlining of macro Vec_PtrForEachEntry
    for  i = 0; (i < Vec_PtrSize(p.pAig.vCis)) && (((pObjPi) = (Aig_Obj_t *)Vec_PtrEntry(p.pAig.vCis, i)), 1); i++  {
        pModel[i] = Abc_InfoHasBit(Ssw_ObjSim(p, pObjPi.Id), BestPat);
//        Abc_Print( 1, "%d", pModel[i] );
    }
    pModel[Aig_ManCiNum(p.pAig)] = pObjPo.Id;
//    Abc_Print( 1, "\n" );
    return pModel;
}

/**Function*************************************************************

  Synopsis    [Returns 1 if the one of the output is already non-constant 0.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Ssw_SmlCheckOutput(p *Ssw_Sml_t) *int32 {
    var pObj *Aig_Obj_t;
    var i int32;
    // make sure the reference simulation pattern does not detect the bug
    pObj = Aig_ManCo( p.pAig, 0 );
    assert(Aig_ObjFanin0(pObj).fPhase == uint32(Aig_ObjFaninC0(pObj)), "Aig_ObjFanin0(pObj).fPhase == uint32(Aig_ObjFaninC0(pObj))");
    // inlining of macro Aig_ManForEachCo
    // inlining of macro Vec_PtrForEachEntry
    for  i = 0; (i < Vec_PtrSize(p.pAig.vCos)) && (((pObj) = (Aig_Obj_t *)Vec_PtrEntry(p.pAig.vCos, i)), 1); i++  {
        if !(Ssw_SmlObjIsConstWord( p, Aig_ObjFanin0(pObj) ) != 0) {
            // create the counter-example from this pattern
            return Ssw_SmlCheckOutputSavePattern( p, pObj );
        }
    }
    return nil;
}



/**Function*************************************************************

  Synopsis    [Assigns random patterns to the PI node.]

  Description []

  SideEffects []

  SeeAlso     []

***********************************************************************/


/**Function*************************************************************

  Synopsis    [Assigns random patterns to the PI node.]

  Description []

  SideEffects []

  SeeAlso     []

***********************************************************************/


/**Function*************************************************************

  Synopsis    [Assigns constant patterns to the PI node.]

  Description []

  SideEffects []

  SeeAlso     []

***********************************************************************/


/**Function*************************************************************

  Synopsis    [Assigns constant patterns to the PI node.]

  Description []

  SideEffects []

  SeeAlso     []

***********************************************************************/


/**Function*************************************************************

  Synopsis    [Assigns constant patterns to the PI node.]

  Description []

  SideEffects []

  SeeAlso     []

***********************************************************************/


/**Function*************************************************************

  Synopsis    [Assings distance-1 simulation info for the PIs.]

  Description []

  SideEffects []

  SeeAlso     []

***********************************************************************/


/**Function*************************************************************

  Synopsis    [Assings distance-1 simulation info for the PIs.]

  Description []

  SideEffects []

  SeeAlso     []

***********************************************************************/


/**Function*************************************************************

  Synopsis    [Simulates one node.]

  Description []

  SideEffects []

  SeeAlso     []

***********************************************************************/


/**Function*************************************************************

  Synopsis    [Simulates one node.]

  Description []

  SideEffects []

  SeeAlso     []

***********************************************************************/


/**Function*************************************************************

  Synopsis    [Simulates one node.]

  Description []

  SideEffects []

  SeeAlso     []

***********************************************************************/


/**Function*************************************************************

  Synopsis    [Simulates one node.]

  Description []

  SideEffects []

  SeeAlso     []

***********************************************************************/


/**Function*************************************************************

  Synopsis    [Simulates one node.]

  Description []

  SideEffects []

  SeeAlso     []

***********************************************************************/



/**Function*************************************************************

  Synopsis    [Assings random simulation info for the PIs.]

  Description []

  SideEffects []

  SeeAlso     []

***********************************************************************/
func Ssw_SmlInitialize(p *Ssw_Sml_t, fInit int32) {
    var pObj *Aig_Obj_t;
    var i int32;
    if fInit != 0 {
        assert(Aig_ManRegNum(p.pAig) > 0, "Aig_ManRegNum(p.pAig) > 0");
        assert(Aig_ManRegNum(p.pAig) <= Aig_ManCiNum(p.pAig), "Aig_ManRegNum(p.pAig) <= Aig_ManCiNum(p.pAig)");
        // assign random info for primary inputs
        // inlining of macro Saig_ManForEachPi
        // inlining of macro Vec_PtrForEachEntryStop
        for  i = 0; (i < Saig_ManPiNum(p.pAig)) && (((pObj) = (Aig_Obj_t *)Vec_PtrEntry(p.pAig.vCis, i)), 1); i++  {
            Ssw_SmlAssignRandom( p, pObj );
        }
        // assign the initial state for the latches
        // inlining of macro Saig_ManForEachLo
        for  i = 0; (i < Saig_ManRegNum(p.pAig)) && (((pObj) = (Aig_Obj_t *)Vec_PtrEntry(p.pAig.vCis, i+Saig_ManPiNum(p.pAig))), 1); i++  {
            Ssw_SmlObjAssignConst( p, pObj, 0, 0 );
        }
    
    } else {
        // inlining of macro Aig_ManForEachCi
        // inlining of macro Vec_PtrForEachEntry
        for  i = 0; (i < Vec_PtrSize(p.pAig.vCis)) && (((pObj) = (Aig_Obj_t *)Vec_PtrEntry(p.pAig.vCis, i)), 1); i++  {
            Ssw_SmlAssignRandom( p, pObj );
        }
    }
}

/**Function*************************************************************

  Synopsis    [Assings random simulation info for the PIs.]

  Description []

  SideEffects []

  SeeAlso     []

***********************************************************************/
func Ssw_SmlInitializeSpecial(p *Ssw_Sml_t, vInit *Vec_Int_t) {
    var pObj *Aig_Obj_t;
    var Entry, i, nRegs int32;
    nRegs = Aig_ManRegNum(p.pAig);
    assert(nRegs > 0, "nRegs > 0");
    assert(nRegs <= Aig_ManCiNum(p.pAig), "nRegs <= Aig_ManCiNum(p.pAig)");
    assert(Vec_IntSize(vInit) == nRegs * p.nWordsFrame, "Vec_IntSize(vInit) == nRegs * p.nWordsFrame");
    // assign random info for primary inputs
    // inlining of macro Saig_ManForEachPi
    // inlining of macro Vec_PtrForEachEntryStop
    for  i = 0; (i < Saig_ManPiNum(p.pAig)) && (((pObj) = (Aig_Obj_t *)Vec_PtrEntry(p.pAig.vCis, i)), 1); i++  {
        Ssw_SmlAssignRandom( p, pObj );
    }
    // assign the initial state for the latches
    // inlining of macro Vec_IntForEachEntry
    for  i = 0; (i < Vec_IntSize(vInit)) && (((Entry) = Vec_IntEntry(vInit, i)), 1); i++  {
        Ssw_SmlObjAssignConstWord( p, Saig_ManLo(p.pAig, i % nRegs), Entry, 0, i / nRegs );
    }
}

/**Function*************************************************************

  Synopsis    [Assings random simulation info for the PIs.]

  Description []

  SideEffects []

  SeeAlso     []

***********************************************************************/
func Ssw_SmlReinitialize(p *Ssw_Sml_t) {
    var pObj, pObjLi, pObjLo *Aig_Obj_t;
    var i int32;
    assert(Aig_ManRegNum(p.pAig) > 0, "Aig_ManRegNum(p.pAig) > 0");
    assert(Aig_ManRegNum(p.pAig) < Aig_ManCiNum(p.pAig), "Aig_ManRegNum(p.pAig) < Aig_ManCiNum(p.pAig)");
    // assign random info for primary inputs
    // inlining of macro Saig_ManForEachPi
    // inlining of macro Vec_PtrForEachEntryStop
    for  i = 0; (i < Saig_ManPiNum(p.pAig)) && (((pObj) = (Aig_Obj_t *)Vec_PtrEntry(p.pAig.vCis, i)), 1); i++  {
        Ssw_SmlAssignRandom( p, pObj );
    }
    // copy simulation info into the inputs
    // inlining of macro Saig_ManForEachLiLo
    for  i = 0; (i < Saig_ManRegNum(p.pAig)) && (((pObjLi) = Saig_ManLi(p.pAig, i)), 1)            && (((pObjLo)=Saig_ManLo(p.pAig, i)), 1); i++  {
        Ssw_SmlNodeTransferFirst( p, pObjLi, pObjLo );
    }
}

/**Function*************************************************************

  Synopsis    [Check if any of the POs becomes non-constant.]

  Description []

  SideEffects []

  SeeAlso     []

***********************************************************************/
func Ssw_SmlCheckNonConstOutputs(p *Ssw_Sml_t) int32 {
    var pObj *Aig_Obj_t;
    var i int32;
    // inlining of macro Saig_ManForEachPo
    // inlining of macro Vec_PtrForEachEntryStop
    for  i = 0; (i < Saig_ManPoNum(p.pAig)) && (((pObj) = (Aig_Obj_t *)Vec_PtrEntry(p.pAig.vCos, i)), 1); i++  {
        if p.pAig.nConstrs && i >= Saig_ManPoNum(p.pAig) - p.pAig.nConstrs {
            return 0; }
        if !(Ssw_SmlNodeIsZero(p, pObj) != 0) {
            return 1; }
    }
    return 0;
}

/**Function*************************************************************

  Synopsis    [Simulates AIG manager.]

  Description [Assumes that the PI simulation info is attached.]

  SideEffects []

  SeeAlso     []

***********************************************************************/


/**Function*************************************************************

  Synopsis    [Converts simulation information to be not normallized.]

  Description []

  SideEffects []

  SeeAlso     []

***********************************************************************/
func Ssw_SmlUnnormalize(p *Ssw_Sml_t) {
    var pObj *Aig_Obj_t;
    var pSims *uint32;
    var i, k int32;
    // convert constant 1
    pSims  = Ssw_ObjSim( p, 0 );
    for  i = 0; i < p.nWordsFrame; i++  {
        pSims[i] = ~pSims[i];
    }
    // convert internal nodes
    // inlining of macro Aig_ManForEachNode
    // inlining of macro Vec_PtrForEachEntry
    for  k = 0; (k < Vec_PtrSize(p.pAig.vObjs)) && (((pObj) = (Aig_Obj_t *)Vec_PtrEntry(p.pAig.vObjs, k)), 1); k++  { if (pObj) == nil || !Aig_ObjIsNode(pObj) {} else {
        if pObj.fPhase == 0 {
            continue; }
        pSims  = Ssw_ObjSim( p, pObj.Id );
        for  i = 0; i < p.nWordsFrame; i++  {
            pSims[i] = ~pSims[i];
        }
    }
    }
    // PIs/POs are always stored in their natural state
}

/**Function*************************************************************

  Synopsis    [Simulates AIG manager.]

  Description [Assumes that the PI simulation info is attached.]

  SideEffects []

  SeeAlso     []

***********************************************************************/
func Ssw_SmlSimulateOneDyn_rec(p *Ssw_Sml_t, pObj *Aig_Obj_t, f int32, pVisited *int32, nVisCounter int32) {
//    if ( Aig_ObjIsTravIdCurrent(p->pAig, pObj) )
//        return;
//    Aig_ObjSetTravIdCurrent(p->pAig, pObj);
    if pVisited[p.nFrames*pObj.Id+f] == nVisCounter {
        return; }
    pVisited[p.nFrames*pObj.Id+f] = nVisCounter;
    if Saig_ObjIsPi( p.pAig, pObj ) || Aig_ObjIsConst1(pObj) {
        return; }
    if Saig_ObjIsLo( p.pAig, pObj ) {
        if f == 0 {
            return; }
        Ssw_SmlSimulateOneDyn_rec( p, Saig_ObjLoToLi(p.pAig, pObj), f-1, pVisited, nVisCounter );
        Ssw_SmlNodeTransferNext( p, Saig_ObjLoToLi(p.pAig, pObj), pObj, f-1 );
        return;
    }
    if Saig_ObjIsLi( p.pAig, pObj ) {
        Ssw_SmlSimulateOneDyn_rec( p, Aig_ObjFanin0(pObj), f, pVisited, nVisCounter );
        Ssw_SmlNodeCopyFanin( p, pObj, f );
        return;
    }
    assert(Aig_ObjIsNode(pObj), "Aig_ObjIsNode(pObj)");
    Ssw_SmlSimulateOneDyn_rec( p, Aig_ObjFanin0(pObj), f, pVisited, nVisCounter );
    Ssw_SmlSimulateOneDyn_rec( p, Aig_ObjFanin1(pObj), f, pVisited, nVisCounter );
    Ssw_SmlNodeSimulate( p, pObj, f );
}

/**Function*************************************************************

  Synopsis    [Simulates AIG manager.]

  Description [Assumes that the PI simulation info is attached.]

  SideEffects []

  SeeAlso     []

***********************************************************************/
func Ssw_SmlSimulateOneFrame(p *Ssw_Sml_t) {
    var pObj, pObjLi, pObjLo *Aig_Obj_t;
    var i int32;
    var clk abctime;
clk = Abc_Clock();
    // simulate the nodes
    // inlining of macro Aig_ManForEachNode
    // inlining of macro Vec_PtrForEachEntry
    for  i = 0; (i < Vec_PtrSize(p.pAig.vObjs)) && (((pObj) = (Aig_Obj_t *)Vec_PtrEntry(p.pAig.vObjs, i)), 1); i++  { if (pObj) == nil || !Aig_ObjIsNode(pObj) {} else {
        Ssw_SmlNodeSimulate( p, pObj, 0 ); }
    }
    // copy simulation info into outputs
    // inlining of macro Saig_ManForEachLi
    for  i = 0; (i < Saig_ManRegNum(p.pAig)) && (((pObj) = (Aig_Obj_t *)Vec_PtrEntry(p.pAig.vCos, i+Saig_ManPoNum(p.pAig))), 1); i++  {
        Ssw_SmlNodeCopyFanin( p, pObj, 0 );
    }
    // copy simulation info into the inputs
    // inlining of macro Saig_ManForEachLiLo
    for  i = 0; (i < Saig_ManRegNum(p.pAig)) && (((pObjLi) = Saig_ManLi(p.pAig, i)), 1)            && (((pObjLo)=Saig_ManLo(p.pAig, i)), 1); i++  {
        Ssw_SmlNodeTransferNext( p, pObjLi, pObjLo, 0 );
    }
p.timeSim += Abc_Clock() - clk;
p.nSimRounds++;
}


/**Function*************************************************************

  Synopsis    [Performs simulation of the uninitialized circuit.]

  Description []

  SideEffects []

  SeeAlso     []

***********************************************************************/
func Ssw_SmlSimulateComb(pAig *Aig_Man_t, nWords int32) *Ssw_Sml_t {
    var p *Ssw_Sml_t;
    p = Ssw_SmlStart( pAig, 0, 1, nWords );
    Ssw_SmlInitialize( p, 0 );
    Ssw_SmlSimulateOne( p );
    return p;
}

/**Function*************************************************************

  Synopsis    [Performs simulation of the initialized circuit.]

  Description []

  SideEffects []

  SeeAlso     []

***********************************************************************/
func Ssw_SmlSimulateSeq(pAig *Aig_Man_t, nPref int32, nFrames int32, nWords int32) *Ssw_Sml_t {
    var p *Ssw_Sml_t;
    p = Ssw_SmlStart( pAig, nPref, nFrames, nWords );
    Ssw_SmlInitialize( p, 1 );
    Ssw_SmlSimulateOne( p );
    p.fNonConstOut = Ssw_SmlCheckNonConstOutputs( p );
    return p;
}

/**Function*************************************************************

  Synopsis    [Performs next round of sequential simulation.]

  Description []

  SideEffects []

  SeeAlso     []

***********************************************************************/
func Ssw_SmlResimulateSeq(p *Ssw_Sml_t) {
    Ssw_SmlReinitialize( p );
    Ssw_SmlSimulateOne( p );
    p.fNonConstOut = Ssw_SmlCheckNonConstOutputs( p );
}


/**Function*************************************************************

  Synopsis    [Creates sequential counter-example from the simulation info.]

  Description []

  SideEffects []

  SeeAlso     []

***********************************************************************/
func Ssw_SmlGetCounterExample(p *Ssw_Sml_t) *Abc_Cex_t {
    var pCex *Abc_Cex_t;
    var pObj *Aig_Obj_t;
    var pSims *uint32;
    var iPo, iFrame, iBit, i, k int32;

    // make sure the simulation manager has it
    assert(p.fNonConstOut, "p.fNonConstOut");

    // find the first output that failed
    iPo = -1;
    iBit = -1;
    iFrame = -1;
    // inlining of macro Saig_ManForEachPo
    // inlining of macro Vec_PtrForEachEntryStop
    for  iPo = 0; (iPo < Saig_ManPoNum(p.pAig)) && (((pObj) = (Aig_Obj_t *)Vec_PtrEntry(p.pAig.vCos, iPo)), 1); iPo++  {
        if Ssw_SmlNodeIsZero(p, pObj) != 0 {
            continue; }
        pSims = Ssw_ObjSim( p, pObj.Id );
        for  i = p.nWordsPref; i < p.nWordsTotal; i++  {
            if pSims[i] {
                iFrame = i / p.nWordsFrame;
                iBit = 32 * (i % p.nWordsFrame) + Aig_WordFindFirstBit( pSims[i] );
                break;
            }
        }
        break;
    }
    assert(iPo < Aig_ManCoNum(p.pAig)-Aig_ManRegNum(p.pAig), "iPo < Aig_ManCoNum(p.pAig)-Aig_ManRegNum(p.pAig)");
    assert(iFrame < p.nFrames, "iFrame < p.nFrames");
    assert(iBit < 32 * p.nWordsFrame, "iBit < 32 * p.nWordsFrame");

    // allocate the counter example
    pCex = Abc_CexAlloc( Aig_ManRegNum(p.pAig), Aig_ManCiNum(p.pAig) - Aig_ManRegNum(p.pAig), iFrame + 1 );
    pCex.iPo    = iPo;
    pCex.iFrame = iFrame;

    // copy the bit data
    // inlining of macro Saig_ManForEachLo
    for  k = 0; (k < Saig_ManRegNum(p.pAig)) && (((pObj) = (Aig_Obj_t *)Vec_PtrEntry(p.pAig.vCis, k+Saig_ManPiNum(p.pAig))), 1); k++  {
        pSims = Ssw_ObjSim( p, pObj.Id );
        if Abc_InfoHasBit( pSims, iBit ) {
            Abc_InfoSetBit( pCex.pData, k ); }
    }
    for  i = 0; i <= iFrame; i++  {
        // inlining of macro Saig_ManForEachPi
        // inlining of macro Vec_PtrForEachEntryStop
        for  k = 0; (k < Saig_ManPiNum(p.pAig)) && (((pObj) = (Aig_Obj_t *)Vec_PtrEntry(p.pAig.vCis, k)), 1); k++  {
            pSims = Ssw_ObjSim( p, pObj.Id );
            if Abc_InfoHasBit( pSims, 32 * p.nWordsFrame * i + iBit ) {
                Abc_InfoSetBit( pCex.pData, pCex.nRegs + pCex.nPis * i + k ); }
        }
    }
    // verify the counter example
    if !Saig_ManVerifyCex( p.pAig, pCex ) {
        Abc_Print( 1, "Ssw_SmlGetCounterExample(): Counter-example is invalid.\n" );
        Abc_CexFree( pCex );
        pCex = nil;
    }
    return pCex;
}

////////////////////////////////////////////////////////////////////////
///                       END OF FILE                                ///
////////////////////////////////////////////////////////////////////////


//ABC_NAMESPACE_IMPL_END
