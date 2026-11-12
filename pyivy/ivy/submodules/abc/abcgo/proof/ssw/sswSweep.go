package ssw

import "fmt"

import "os"

/**CFile****************************************************************

  FileName    [sswSweep.c]

  SystemName  [ABC: Logic synthesis and verification system.]

  PackageName [Inductive prover with constraints.]

  Synopsis    [One round of SAT sweeping.]

  Author      [Alan Mishchenko]
  
  Affiliation [UC Berkeley]

  Date        [Ver. 1.0. Started - September 1, 2008.]

  Revision    [$Id: sswSweep.c,v 1.00 2008/09/01 00:00:00 alanmi Exp $]

***********************************************************************/

//#include "sswInt.h"
//#include "misc/bar/bar.h"

//ABC_NAMESPACE_IMPL_START


////////////////////////////////////////////////////////////////////////
///                        DECLARATIONS                              ///
////////////////////////////////////////////////////////////////////////

////////////////////////////////////////////////////////////////////////
///                     FUNCTION DEFINITIONS                         ///
////////////////////////////////////////////////////////////////////////

/**Function*************************************************************

  Synopsis    [Retrives value of the PI in the original AIG.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Ssw_ManGetSatVarValue(p *Ssw_Man_t, pObj *Aig_Obj_t, f int32) int32 {
    var fUseNoBoundary int32;
    var pObjFraig *Aig_Obj_t;
    var Value int32;
//    assert( Aig_ObjIsCi(pObj) );
    pObjFraig = Ssw_ObjFrame( p, pObj, f );
    if fUseNoBoundary != 0 {
        Value = Ssw_CnfGetNodeValue( p.pMSat, Aig_Regular(pObjFraig) );
        Value ^= Aig_IsComplement(pObjFraig);
    
    } else {
        var nVarNum int32  = Ssw_ObjSatNum( p.pMSat, Aig_Regular(pObjFraig) );
        if !(nVarNum != 0) {
            Value = 0;
        } else {
            Value = (Aig_IsComplement(pObjFraig) ^ sat_solver_var_value( p.pMSat.pSat, nVarNum ));
        }
    }

//    Value = (Aig_IsComplement(pObjFraig) ^ ((!nVarNum)? 0 : sat_solver_var_value( p->pSat, nVarNum )));
//    Value = (!nVarNum)? Aig_ManRandom(0) & 1 : (Aig_IsComplement(pObjFraig) ^ sat_solver_var_value( p->pSat, nVarNum ));
    if p.pPars.fPolarFlip {
        if Aig_Regular(pObjFraig).fPhase {  Value ^= 1; }
    }
    return Value;
}

/**Function*************************************************************

  Synopsis    [Performs fraiging for the internal nodes.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Ssw_CheckConstraints(p *Ssw_Man_t) {
    var pObj, pObj2 *Aig_Obj_t;
    var nConstrPairs, i int32;
    var Counter int32;
    nConstrPairs = Aig_ManCoNum(p.pFrames)-Aig_ManRegNum(p.pAig);
    assert((nConstrPairs & 1) == 0, "(nConstrPairs & 1) == 0");
    for  i = 0; i < nConstrPairs; i += 2  {
        pObj  = Aig_ManCo( p.pFrames, i   );
        pObj2 = Aig_ManCo( p.pFrames, i+1 );
        if Ssw_NodesAreEquiv( p, Aig_ObjFanin0(pObj), Aig_ObjFanin0(pObj2) ) != 1 {
            Ssw_NodesAreConstrained( p, Aig_ObjChild0(pObj), Aig_ObjChild0(pObj2) );
            Counter++;
        }
    }
    Abc_Print( 1, "Total constraints = %d. Added constraints = %d.\n", nConstrPairs/2, Counter );
}

/**Function*************************************************************

  Synopsis    [Copy pattern from the solver into the internal storage.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Ssw_SmlSavePatternAigPhase(p *Ssw_Man_t, f int32) {
    var pObj *Aig_Obj_t;
    var i int32;
    memset( p.pPatWords, 0, sizeof(unsigned) * p.nPatWords );
    // inlining of macro Aig_ManForEachCi
    // inlining of macro Vec_PtrForEachEntry
    for  i = 0; (i < Vec_PtrSize(p.pAig.vCis)) && (((pObj) = (Aig_Obj_t *)Vec_PtrEntry(p.pAig.vCis, i)), 1); i++  {
        if Aig_ObjPhaseReal( Ssw_ObjFrame(p, pObj, f) ) {
            Abc_InfoSetBit( p.pPatWords, i ); }
    }
}

/**Function*************************************************************

  Synopsis    [Copy pattern from the solver into the internal storage.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Ssw_SmlSavePatternAig(p *Ssw_Man_t, f int32) {
    var pObj *Aig_Obj_t;
    var i int32;
    memset( p.pPatWords, 0, sizeof(unsigned) * p.nPatWords );
    // inlining of macro Aig_ManForEachCi
    // inlining of macro Vec_PtrForEachEntry
    for  i = 0; (i < Vec_PtrSize(p.pAig.vCis)) && (((pObj) = (Aig_Obj_t *)Vec_PtrEntry(p.pAig.vCis, i)), 1); i++  {
        if Ssw_ManGetSatVarValue( p, pObj, f ) != 0 {
            Abc_InfoSetBit( p.pPatWords, i ); }
    }
}

/**Function*************************************************************

  Synopsis    [Saves one counter-example into internal storage.]

  Description []

  SideEffects []

  SeeAlso     []

***********************************************************************/
func Ssw_SmlAddPatternDyn(p *Ssw_Man_t) {
    var pObj *Aig_Obj_t;
    var pInfo *uint32;
    var i, nVarNum int32;
    // iterate through the PIs of the frames
    // inlining of macro Vec_PtrForEachEntry
    for  i = 0; (i < Vec_PtrSize(p.pMSat.vUsedPis)) && (((pObj) = (Aig_Obj_t *)Vec_PtrEntry(p.pMSat.vUsedPis, i)), 1); i++  {
        assert(Aig_ObjIsCi(pObj), "Aig_ObjIsCi(pObj)");
        nVarNum = Ssw_ObjSatNum( p.pMSat, pObj );
        assert(nVarNum > 0, "nVarNum > 0");
        if sat_solver_var_value( p.pMSat.pSat, nVarNum ) {
            pInfo = (unsigned *)Vec_PtrEntry( p.vSimInfo, Aig_ObjCioId(pObj) );
            Abc_InfoSetBit( pInfo, p.nPatterns );
        }
    }
}

/**Function*************************************************************

  Synopsis    [Performs fraiging for one node.]

  Description [Returns the fraiged node.]

  SideEffects []

  SeeAlso     []

***********************************************************************/
func Ssw_ManSweepNode(p *Ssw_Man_t, pObj *Aig_Obj_t, f int32, fBmc int32, vPairs *Vec_Int_t) int32 {
    var pObjRepr, pObjFraig, pObjFraig2, pObjReprFraig *Aig_Obj_t;
    var RetValue int32;
    var clk abctime;
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
    // add constraints on demand
    if !(fBmc != 0) && p.pPars.fDynamic {
clk = Abc_Clock();
        Ssw_ManLoadSolver( p, pObjRepr, pObj );
        p.nRecycleCalls++;
p.timeMarkCones += Abc_Clock() - clk;
    }
    // call equivalence checking
    if Aig_Regular(pObjFraig) != Aig_ManConst1(p.pFrames) {
        RetValue = Ssw_NodesAreEquiv( p, Aig_Regular(pObjReprFraig), Aig_Regular(pObjFraig) );
    } else {
        RetValue = Ssw_NodesAreEquiv( p, Aig_Regular(pObjFraig), Aig_Regular(pObjReprFraig) ); }
    if RetValue == 1 {  // proved equivalent
    
        pObjFraig2 = Aig_NotCond( pObjReprFraig, pObj.fPhase ^ pObjRepr.fPhase );
        Ssw_ObjSetFrame( p, pObj, f, pObjFraig2 );
        if p.pPars.fEquivDump2 && (vPairs != nil) {
            Vec_IntPush( vPairs, pObjRepr.Id );
            Vec_IntPush( vPairs, pObj.Id );
        }
        return 0;
    }
    if p.pPars.fEquivDump && (vPairs != nil) {
        Vec_IntPush( vPairs, pObjRepr.Id );
        Vec_IntPush( vPairs, pObj.Id );
    }
    if RetValue == -1 { // timed out
    
        Ssw_ClassesRemoveNode( p.ppClasses, pObj );
        return 1;
    }
    // disproved the equivalence
    if !(fBmc != 0) && p.pPars.fDynamic {
        Ssw_SmlAddPatternDyn( p );
        p.nPatterns++;
        return 1;
    
    } else {
        Ssw_SmlSavePatternAig( p, f ); }
    if !p.pPars.fConstrs {
        Ssw_ManResimulateWord( p, pObj, pObjRepr, f );
    } else {
        Ssw_ManResimulateBit( p, pObj, pObjRepr ); }
    assert(Aig_ObjRepr( p.pAig, pObj ) != pObjRepr, "Aig_ObjRepr( p.pAig, pObj ) != pObjRepr");
    if Aig_ObjRepr( p.pAig, pObj ) == pObjRepr {
        Abc_Print( 1, "Ssw_ManSweepNode(): Failed to refine representative.\n" );
    }
    return 1;
}

/**Function*************************************************************

  Synopsis    [Performs fraiging for the internal nodes.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Ssw_ManSweepBmc(p *Ssw_Man_t) int32 {
    var pProgress *Bar_Progress_t  = nil;
    var pObj, pObjNew, pObjLi, pObjLo *Aig_Obj_t;
    var i, f int32;
    var clk abctime;
clk = Abc_Clock();

    // start initialized timeframes
    p.pFrames = Aig_ManStart( Aig_ManObjNumMax(p.pAig) * p.pPars.nFramesK );
    // inlining of macro Saig_ManForEachLo
    for  i = 0; (i < Saig_ManRegNum(p.pAig)) && (((pObj) = (Aig_Obj_t *)Vec_PtrEntry(p.pAig.vCis, i+Saig_ManPiNum(p.pAig))), 1); i++  {
        Ssw_ObjSetFrame( p, pObj, 0, Aig_ManConst0(p.pFrames) );
    }

    // sweep internal nodes
    p.fRefined = 0;
    if p.pPars.fVerbose {
        pProgress = Bar_ProgressStart( os.Stdout, Aig_ManObjNumMax(p.pAig) * p.pPars.nFramesK ); }
    for  f = 0; f < p.pPars.nFramesK; f++  {
        // map constants and PIs
        Ssw_ObjSetFrame( p, Aig_ManConst1(p.pAig), f, Aig_ManConst1(p.pFrames) );
        // inlining of macro Saig_ManForEachPi
        // inlining of macro Vec_PtrForEachEntryStop
        for  i = 0; (i < Saig_ManPiNum(p.pAig)) && (((pObj) = (Aig_Obj_t *)Vec_PtrEntry(p.pAig.vCis, i)), 1); i++  {
            Ssw_ObjSetFrame( p, pObj, f, Aig_ObjCreateCi(p.pFrames) );
        }
        // sweep flops
        // inlining of macro Saig_ManForEachLo
        for  i = 0; (i < Saig_ManRegNum(p.pAig)) && (((pObj) = (Aig_Obj_t *)Vec_PtrEntry(p.pAig.vCis, i+Saig_ManPiNum(p.pAig))), 1); i++  {
            p.fRefined |= Ssw_ManSweepNode( p, pObj, f, 1, nil );
        }
        // sweep internal nodes
        // inlining of macro Aig_ManForEachNode
        // inlining of macro Vec_PtrForEachEntry
        for  i = 0; (i < Vec_PtrSize(p.pAig.vObjs)) && (((pObj) = (Aig_Obj_t *)Vec_PtrEntry(p.pAig.vObjs, i)), 1); i++  { if (pObj) == nil || !Aig_ObjIsNode(pObj) {} else {
            if p.pPars.fVerbose {
                Bar_ProgressUpdate( pProgress, Aig_ManObjNumMax(p.pAig) * f + i, nil ); }
            pObjNew = Aig_And( p.pFrames, Ssw_ObjChild0Fra(p, pObj, f), Ssw_ObjChild1Fra(p, pObj, f) );
            Ssw_ObjSetFrame( p, pObj, f, pObjNew );
            p.fRefined |= Ssw_ManSweepNode( p, pObj, f, 1, nil );
        }
        }
        // quit if this is the last timeframe
        if f == p.pPars.nFramesK - 1 {
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
    if p.pPars.fVerbose {
        Bar_ProgressStop( pProgress ); }

    // cleanup
//    Ssw_ClassesCheck( p->ppClasses );
p.timeBmc += Abc_Clock() - clk;
    return p.fRefined;
}


/**Function*************************************************************

  Synopsis    [Generates AIG with the following nodes put into seq miters.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Ssw_ManDumpEquivMiter(p *Aig_Man_t, vPairs *Vec_Int_t, Num int32, fAddOuts int32) {
    var pFile *os.File;
    var pBuffer string // was [16]byte in C

    var pNew *Aig_Man_t;
    pBuffer = fmt.Sprintf("equiv%03d.aig", Num);
    pFile = fopen( pBuffer, "w" );
    if pFile == nil {
        Abc_Print( 1, "Cannot open file %s for writing.\n", pBuffer );
        return;
    }
    _ = (pFile).Close();
    pNew = Saig_ManCreateEquivMiter( p, vPairs, fAddOuts );
    Ioa_WriteAiger( pNew, pBuffer, 0, 0 );
    Aig_ManStop( pNew );
    Abc_Print( 1, "AIG with %4d disproved equivs is dumped into file \"%s\".\n", Vec_IntSize(vPairs)/2, pBuffer );
}


/**Function*************************************************************

  Synopsis    [Performs fraiging for the internal nodes.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
var spStatic_Ssw_ManSweep_Counter int32;
func Ssw_ManSweep(p *Ssw_Man_t) int32 {
    
    var pProgress *Bar_Progress_t  = nil;
    var pObj, pObj2, pObjNew *Aig_Obj_t;
    var nConstrPairs, i, f int32;
    var clk abctime;
    var vObjPairs *Vec_Int_t;

    // perform speculative reduction
clk = Abc_Clock();
    // create timeframes
    p.pFrames = Ssw_FramesWithClasses( p );
    // add constants
    nConstrPairs = Aig_ManCoNum(p.pFrames)-Aig_ManRegNum(p.pAig);
    assert((nConstrPairs & 1) == 0, "(nConstrPairs & 1) == 0");
    for  i = 0; i < nConstrPairs; i += 2  {
        pObj  = Aig_ManCo( p.pFrames, i   );
        pObj2 = Aig_ManCo( p.pFrames, i+1 );
        Ssw_NodesAreConstrained( p, Aig_ObjChild0(pObj), Aig_ObjChild0(pObj2) );
    }
    // build logic cones for register inputs
    for  i = 0; i < Aig_ManRegNum(p.pAig); i++  {
        pObj  = Aig_ManCo( p.pFrames, nConstrPairs + i );
        Ssw_CnfNodeAddToSolver( p.pMSat, Aig_ObjFanin0(pObj) );//
    }
    sat_solver_simplify( p.pMSat.pSat );

    // map constants and PIs of the last frame
    f = p.pPars.nFramesK;
    Ssw_ObjSetFrame( p, Aig_ManConst1(p.pAig), f, Aig_ManConst1(p.pFrames) );
    // inlining of macro Saig_ManForEachPi
    // inlining of macro Vec_PtrForEachEntryStop
    for  i = 0; (i < Saig_ManPiNum(p.pAig)) && (((pObj) = (Aig_Obj_t *)Vec_PtrEntry(p.pAig.vCis, i)), 1); i++  {
        Ssw_ObjSetFrame( p, pObj, f, Aig_ObjCreateCi(p.pFrames) );
    }
p.timeReduce += Abc_Clock() - clk;

    // sweep internal nodes
    p.fRefined = 0;
    Ssw_ClassesClearRefined( p.ppClasses );
    if p.pPars.fVerbose {
        pProgress = Bar_ProgressStart( os.Stdout, Aig_ManObjNumMax(p.pAig) ); }
    if p.pPars.fEquivDump || p.pPars.fEquivDump2 {
        vObjPairs = Vec_IntAlloc(1000);
    } else {
        vObjPairs = nil;
    }
    // inlining of macro Aig_ManForEachObj
    // inlining of macro Vec_PtrForEachEntry
    for  i = 0; (i < Vec_PtrSize(p.pAig.vObjs)) && (((pObj) = (Aig_Obj_t *)Vec_PtrEntry(p.pAig.vObjs, i)), 1); i++  { if (pObj) == nil {} else {
        if p.pPars.fVerbose {
            Bar_ProgressUpdate( pProgress, i, nil ); }
        if Saig_ObjIsLo(p.pAig, pObj) {
            p.fRefined |= Ssw_ManSweepNode( p, pObj, f, 0, vObjPairs );
        } else if Aig_ObjIsNode(pObj) {
            pObjNew = Aig_And( p.pFrames, Ssw_ObjChild0Fra(p, pObj, f), Ssw_ObjChild1Fra(p, pObj, f) );
            Ssw_ObjSetFrame( p, pObj, f, pObjNew );
            p.fRefined |= Ssw_ManSweepNode( p, pObj, f, 0, vObjPairs );
        }
    }
    }
    if p.pPars.fVerbose {
        Bar_ProgressStop( pProgress ); }

    // cleanup
//    Ssw_ClassesCheck( p->ppClasses );
    if p.pPars.fEquivDump {
        Ssw_ManDumpEquivMiter( p.pAig, vObjPairs, spStatic_Ssw_ManSweep_Counter++, 1 ); }
    if p.pPars.fEquivDump2 && !p.fRefined {
        Ssw_ManDumpEquivMiter( p.pAig, vObjPairs, 0, 0 ); }
    Vec_IntFreeP( &vObjPairs );
    return p.fRefined;
}

////////////////////////////////////////////////////////////////////////
///                       END OF FILE                                ///
////////////////////////////////////////////////////////////////////////


//ABC_NAMESPACE_IMPL_END
