package ssw

import "os"

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
//#include "sat/cnf/cnf.h"
//#include "misc/bar/bar.h"

//ABC_NAMESPACE_IMPL_START


////////////////////////////////////////////////////////////////////////
///                        DECLARATIONS                              ///
////////////////////////////////////////////////////////////////////////

////////////////////////////////////////////////////////////////////////
///                     FUNCTION DEFINITIONS                         ///
////////////////////////////////////////////////////////////////////////

/**Function*************************************************************

  Synopsis    [Constructs initialized timeframes with constraints as POs.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/


/**Function*************************************************************

  Synopsis    [Finds one satisfiable assignment of the timeframes.]

  Description []

  SideEffects []

  SeeAlso     []

***********************************************************************/


/**Function*************************************************************

  Synopsis    [Performs fraiging for one node.]

  Description [Returns the fraiged node.]
               
  SideEffects []

  SeeAlso     []

***********************************************************************/


/**Function*************************************************************

  Synopsis    [Performs fraiging for one node.]

  Description [Returns the fraiged node.]
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Ssw_ManPrintPolarity(p *Aig_Man_t) {
    var pObj *Aig_Obj_t;
    var i int32;
    // inlining of macro Aig_ManForEachObj
    // inlining of macro Vec_PtrForEachEntry
    for  i = 0; (i < Vec_PtrSize(p.vObjs)) && (((pObj) = (Aig_Obj_t *)Vec_PtrEntry(p.vObjs, i)), 1); i++  { if (pObj) == nil {} else {
        Abc_Print( 1, "%d", pObj.fPhase ); }
    }
    Abc_Print( 1, "\n" );
}

/**Function*************************************************************

  Synopsis    [Performs fraiging for one node.]

  Description [Returns the fraiged node.]
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Ssw_ManRefineByConstrSim(p *Ssw_Man_t) {
    var pObj, pObjLi *Aig_Obj_t;
    var f, i, iLits, RetValue1, RetValue2 int32;
    var nFrames int32  = Vec_IntSize(p.vInits) / Saig_ManPiNum(p.pAig);
    assert(Vec_IntSize(p.vInits) % Saig_ManPiNum(p.pAig) == 0, "Vec_IntSize(p.vInits) % Saig_ManPiNum(p.pAig) == 0");
    // assign register outputs
    // inlining of macro Saig_ManForEachLi
    for  i = 0; (i < Saig_ManRegNum(p.pAig)) && (((pObj) = (Aig_Obj_t *)Vec_PtrEntry(p.pAig.vCos, i+Saig_ManPoNum(p.pAig))), 1); i++  {
        pObj.fMarkB = 0;
    }
    // simulate the timeframes
    iLits = 0;
    for  f = 0; f < nFrames; f++  {
        // set the PI simulation information
        Aig_ManConst1(p.pAig).fMarkB = 1;
        // inlining of macro Saig_ManForEachPi
        // inlining of macro Vec_PtrForEachEntryStop
        for  i = 0; (i < Saig_ManPiNum(p.pAig)) && (((pObj) = (Aig_Obj_t *)Vec_PtrEntry(p.pAig.vCis, i)), 1); i++  {
            pObj.fMarkB = Vec_IntEntry( p.vInits, iLits++ );
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
        // check the outputs
        // inlining of macro Saig_ManForEachPo
        // inlining of macro Vec_PtrForEachEntryStop
        for  i = 0; (i < Saig_ManPoNum(p.pAig)) && (((pObj) = (Aig_Obj_t *)Vec_PtrEntry(p.pAig.vCos, i)), 1); i++  {
            if i < Saig_ManPoNum(p.pAig) - Saig_ManConstrNum(p.pAig) {
                if pObj.fMarkB && Saig_ManConstrNum(p.pAig) {
                    Abc_Print( 1, "output %d failed in frame %d.\n", i, f ); }
            
            } else {
                if pObj.fMarkB && Saig_ManConstrNum(p.pAig) {
                    Abc_Print( 1, "constraint %d failed in frame %d.\n", i, f ); }
            }
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
    assert(iLits == Vec_IntSize(p.vInits), "iLits == Vec_IntSize(p.vInits)");
}


/**Function*************************************************************

  Synopsis    [Performs fraiging for one node.]

  Description [Returns the fraiged node.]
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Ssw_ManSweepNodeConstr(p *Ssw_Man_t, pObj *Aig_Obj_t, f int32, fBmc int32) int32 {
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
    
        Ssw_ClassesRemoveNode( p.ppClasses, pObj );
        return 1;
    }
    // disproved equivalence
    Ssw_SmlSavePatternAig( p, f );
    Ssw_ManResimulateBit( p, pObj, pObjRepr );
    assert(Aig_ObjRepr( p.pAig, pObj ) != pObjRepr, "Aig_ObjRepr( p.pAig, pObj ) != pObjRepr");
    if Aig_ObjRepr( p.pAig, pObj ) == pObjRepr {
        Abc_Print( 1, "Ssw_ManSweepNodeConstr(): Failed to refine representative.\n" );
    }
    return 1;
}

/**Function*************************************************************

  Synopsis    [Performs fraiging for the internal nodes.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Ssw_ManSweepBmcConstr_rec(p *Ssw_Man_t, pObj *Aig_Obj_t, f int32) *Aig_Obj_t {
    var pObjNew, pObjLi *Aig_Obj_t;
    pObjNew = Ssw_ObjFrame( p, pObj, f );
    if pObjNew != nil {
        return pObjNew; }
    assert(!Saig_ObjIsPi(p.pAig, pObj), "!Saig_ObjIsPi(p.pAig, pObj)");
    if Saig_ObjIsLo(p.pAig, pObj) {
        assert(f > 0, "f > 0");
        pObjLi  = Saig_ObjLoToLi( p.pAig, pObj );
        pObjNew = Ssw_ManSweepBmcConstr_rec( p, Aig_ObjFanin0(pObjLi), f-1 );
        pObjNew = Aig_NotCond( pObjNew, Aig_ObjFaninC0(pObjLi) );
    
    } else {
        assert(Aig_ObjIsNode(pObj), "Aig_ObjIsNode(pObj)");
        Ssw_ManSweepBmcConstr_rec( p, Aig_ObjFanin0(pObj), f );
        Ssw_ManSweepBmcConstr_rec( p, Aig_ObjFanin1(pObj), f );
        pObjNew = Aig_And( p.pFrames, Ssw_ObjChild0Fra(p, pObj, f), Ssw_ObjChild1Fra(p, pObj, f) );
    }
    Ssw_ObjSetFrame( p, pObj, f, pObjNew );
    assert(pObjNew != nil, "pObjNew != nil");
    return pObjNew;
}

/**Function*************************************************************

  Synopsis    [Performs fraiging for the internal nodes.]

  Description []

  SideEffects []

  SeeAlso     []

***********************************************************************/
func Ssw_ManSweepBmcConstr_old(p *Ssw_Man_t) int32 {
    var pProgress *Bar_Progress_t  = nil;
    var pObj, pObjNew, pObjLi, pObjLo *Aig_Obj_t;
    var i, f, iLits int32;
    var clk abctime;
clk = Abc_Clock();

    // start initialized timeframes
    p.pFrames = Aig_ManStart( Aig_ManObjNumMax(p.pAig) * p.pPars.nFramesK );
    // inlining of macro Saig_ManForEachLo
    for  i = 0; (i < Saig_ManRegNum(p.pAig)) && (((pObj) = (Aig_Obj_t *)Vec_PtrEntry(p.pAig.vCis, i+Saig_ManPiNum(p.pAig))), 1); i++  {
        Ssw_ObjSetFrame( p, pObj, 0, Aig_ManConst0(p.pFrames) );
    }

    // build the constraint outputs
    iLits = 0;
    for  f = 0; f < p.pPars.nFramesK; f++  {
        // map constants and PIs
        Ssw_ObjSetFrame( p, Aig_ManConst1(p.pAig), f, Aig_ManConst1(p.pFrames) );
        // inlining of macro Saig_ManForEachPi
        // inlining of macro Vec_PtrForEachEntryStop
        for  i = 0; (i < Saig_ManPiNum(p.pAig)) && (((pObj) = (Aig_Obj_t *)Vec_PtrEntry(p.pAig.vCis, i)), 1); i++  {
            pObjNew = Aig_ObjCreateCi(p.pFrames);
            pObjNew.fPhase = Vec_IntEntry( p.vInits, iLits++ );
            Ssw_ObjSetFrame( p, pObj, f, pObjNew );
        }
        // build the constraint cones
        // inlining of macro Saig_ManForEachPo
        // inlining of macro Vec_PtrForEachEntryStop
        for  i = 0; (i < Saig_ManPoNum(p.pAig)) && (((pObj) = (Aig_Obj_t *)Vec_PtrEntry(p.pAig.vCos, i)), 1); i++  {
            if i < Saig_ManPoNum(p.pAig) - Saig_ManConstrNum(p.pAig) {
                continue; }
            pObjNew = Ssw_ManSweepBmcConstr_rec( p, Aig_ObjFanin0(pObj), f );
            pObjNew = Aig_NotCond( pObjNew, Aig_ObjFaninC0(pObj) );
            if Aig_Regular(pObjNew) == Aig_ManConst1(p.pFrames) {
                assert(Aig_IsComplement(pObjNew), "Aig_IsComplement(pObjNew)");
                continue;
            }
            Ssw_NodesAreConstrained( p, pObjNew, Aig_ManConst0(p.pFrames) );
        }
    }
    assert(Vec_IntSize(p.vInits) == iLits + Saig_ManPiNum(p.pAig), "Vec_IntSize(p.vInits) == iLits + Saig_ManPiNum(p.pAig)");

    // sweep internal nodes
    p.fRefined = 0;
    if p.pPars.fVerbose {
        pProgress = Bar_ProgressStart( os.Stdout, Aig_ManObjNumMax(p.pAig) * p.pPars.nFramesK ); }
    for  f = 0; f < p.pPars.nFramesK; f++  {
        // sweep internal nodes
        // inlining of macro Aig_ManForEachNode
        // inlining of macro Vec_PtrForEachEntry
        for  i = 0; (i < Vec_PtrSize(p.pAig.vObjs)) && (((pObj) = (Aig_Obj_t *)Vec_PtrEntry(p.pAig.vObjs, i)), 1); i++  { if (pObj) == nil || !Aig_ObjIsNode(pObj) {} else {
            if p.pPars.fVerbose {
                Bar_ProgressUpdate( pProgress, Aig_ManObjNumMax(p.pAig) * f + i, nil ); }
            pObjNew = Aig_And( p.pFrames, Ssw_ObjChild0Fra(p, pObj, f), Ssw_ObjChild1Fra(p, pObj, f) );
            Ssw_ObjSetFrame( p, pObj, f, pObjNew );
            p.fRefined |= Ssw_ManSweepNodeConstr( p, pObj, f, 1 );
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

  Synopsis    [Performs fraiging for the internal nodes.]

  Description []

  SideEffects []

  SeeAlso     []

***********************************************************************/
func Ssw_ManSweepBmcConstr(p *Ssw_Man_t) int32 {
    var pObj, pObjNew, pObjLi, pObjLo *Aig_Obj_t;
    var i, f, iLits int32;
    var clk abctime;
clk = Abc_Clock();

    // start initialized timeframes
    p.pFrames = Aig_ManStart( Aig_ManObjNumMax(p.pAig) * p.pPars.nFramesK );
    // inlining of macro Saig_ManForEachLo
    for  i = 0; (i < Saig_ManRegNum(p.pAig)) && (((pObj) = (Aig_Obj_t *)Vec_PtrEntry(p.pAig.vCis, i+Saig_ManPiNum(p.pAig))), 1); i++  {
        Ssw_ObjSetFrame( p, pObj, 0, Aig_ManConst0(p.pFrames) );
    }

    // build the constraint outputs
    iLits = 0;
    p.fRefined = 0;
    for  f = 0; f < p.pPars.nFramesK; f++  {
        // map constants and PIs
        Ssw_ObjSetFrame( p, Aig_ManConst1(p.pAig), f, Aig_ManConst1(p.pFrames) );
        // inlining of macro Saig_ManForEachPi
        // inlining of macro Vec_PtrForEachEntryStop
        for  i = 0; (i < Saig_ManPiNum(p.pAig)) && (((pObj) = (Aig_Obj_t *)Vec_PtrEntry(p.pAig.vCis, i)), 1); i++  {
            pObjNew = Aig_ObjCreateCi(p.pFrames);
            pObjNew.fPhase = Vec_IntEntry( p.vInits, iLits++ );
            Ssw_ObjSetFrame( p, pObj, f, pObjNew );
        }
        // build the constraint cones
        // inlining of macro Saig_ManForEachPo
        // inlining of macro Vec_PtrForEachEntryStop
        for  i = 0; (i < Saig_ManPoNum(p.pAig)) && (((pObj) = (Aig_Obj_t *)Vec_PtrEntry(p.pAig.vCos, i)), 1); i++  {
            if i < Saig_ManPoNum(p.pAig) - Saig_ManConstrNum(p.pAig) {
                continue; }
            pObjNew = Ssw_ManSweepBmcConstr_rec( p, Aig_ObjFanin0(pObj), f );
            pObjNew = Aig_NotCond( pObjNew, Aig_ObjFaninC0(pObj) );
            if Aig_Regular(pObjNew) == Aig_ManConst1(p.pFrames) {
                assert(Aig_IsComplement(pObjNew), "Aig_IsComplement(pObjNew)");
                continue;
            }
            Ssw_NodesAreConstrained( p, pObjNew, Aig_ManConst0(p.pFrames) );
        }        
        // sweep flops
        // inlining of macro Saig_ManForEachLo
        for  i = 0; (i < Saig_ManRegNum(p.pAig)) && (((pObj) = (Aig_Obj_t *)Vec_PtrEntry(p.pAig.vCis, i+Saig_ManPiNum(p.pAig))), 1); i++  {
            p.fRefined |= Ssw_ManSweepNodeConstr( p, pObj, f, 1 );
        }
        // sweep internal nodes
        // inlining of macro Aig_ManForEachNode
        // inlining of macro Vec_PtrForEachEntry
        for  i = 0; (i < Vec_PtrSize(p.pAig.vObjs)) && (((pObj) = (Aig_Obj_t *)Vec_PtrEntry(p.pAig.vObjs, i)), 1); i++  { if (pObj) == nil || !Aig_ObjIsNode(pObj) {} else {
            pObjNew = Aig_And( p.pFrames, Ssw_ObjChild0Fra(p, pObj, f), Ssw_ObjChild1Fra(p, pObj, f) );
            Ssw_ObjSetFrame( p, pObj, f, pObjNew );
            p.fRefined |= Ssw_ManSweepNodeConstr( p, pObj, f, 1 );
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
    assert(Vec_IntSize(p.vInits) == iLits + Saig_ManPiNum(p.pAig), "Vec_IntSize(p.vInits) == iLits + Saig_ManPiNum(p.pAig)");

    // cleanup
//    Ssw_ClassesCheck( p->ppClasses );
p.timeBmc += Abc_Clock() - clk;
    return p.fRefined;
}




/**Function*************************************************************

  Synopsis    [Performs fraiging for the internal nodes.]

  Description []

  SideEffects []

  SeeAlso     []

***********************************************************************/
func Ssw_FramesWithClasses_rec(p *Ssw_Man_t, pObj *Aig_Obj_t, f int32) *Aig_Obj_t {
    var pObjNew, pObjLi *Aig_Obj_t;
    pObjNew = Ssw_ObjFrame( p, pObj, f );
    if pObjNew != nil {
        return pObjNew; }
    assert(!Saig_ObjIsPi(p.pAig, pObj), "!Saig_ObjIsPi(p.pAig, pObj)");
    if Saig_ObjIsLo(p.pAig, pObj) {
        assert(f > 0, "f > 0");
        pObjLi  = Saig_ObjLoToLi( p.pAig, pObj );
        pObjNew = Ssw_FramesWithClasses_rec( p, Aig_ObjFanin0(pObjLi), f-1 );
        pObjNew = Aig_NotCond( pObjNew, Aig_ObjFaninC0(pObjLi) );
    
    } else {
        assert(Aig_ObjIsNode(pObj), "Aig_ObjIsNode(pObj)");
        Ssw_FramesWithClasses_rec( p, Aig_ObjFanin0(pObj), f );
        Ssw_FramesWithClasses_rec( p, Aig_ObjFanin1(pObj), f );
        pObjNew = Aig_And( p.pFrames, Ssw_ObjChild0Fra(p, pObj, f), Ssw_ObjChild1Fra(p, pObj, f) );
    }
    Ssw_ObjSetFrame( p, pObj, f, pObjNew );
    assert(pObjNew != nil, "pObjNew != nil");
    return pObjNew;
}


/**Function*************************************************************

  Synopsis    [Performs fraiging for the internal nodes.]

  Description []

  SideEffects []

  SeeAlso     []

***********************************************************************/
func Ssw_ManSweepConstr(p *Ssw_Man_t) int32 {
    var pProgress *Bar_Progress_t  = nil;
    var pObj, pObj2, pObjNew *Aig_Obj_t;
    var nConstrPairs, i, f, iLits int32;
    var clk abctime;
//Ssw_ManPrintPolarity( p->pAig );

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

    // map constants and PIs of the last frame
    f = p.pPars.nFramesK;
//    iLits = 0;
    iLits = f * Saig_ManPiNum(p.pAig);
    Ssw_ObjSetFrame( p, Aig_ManConst1(p.pAig), f, Aig_ManConst1(p.pFrames) );
    // inlining of macro Saig_ManForEachPi
    // inlining of macro Vec_PtrForEachEntryStop
    for  i = 0; (i < Saig_ManPiNum(p.pAig)) && (((pObj) = (Aig_Obj_t *)Vec_PtrEntry(p.pAig.vCis, i)), 1); i++  {
        pObjNew = Aig_ObjCreateCi(p.pFrames);
        pObjNew.fPhase = (p.vInits != nil) && Vec_IntEntry(p.vInits, iLits++);
        Ssw_ObjSetFrame( p, pObj, f, pObjNew );
    }
    assert(Vec_IntSize(p.vInits) == iLits, "Vec_IntSize(p.vInits) == iLits");
p.timeReduce += Abc_Clock() - clk;

    // add constraints to all timeframes
    for  f = 0; f <= p.pPars.nFramesK; f++  {
        // inlining of macro Saig_ManForEachPo
        // inlining of macro Vec_PtrForEachEntryStop
        for  i = 0; (i < Saig_ManPoNum(p.pAig)) && (((pObj) = (Aig_Obj_t *)Vec_PtrEntry(p.pAig.vCos, i)), 1); i++  {
            if i < Saig_ManPoNum(p.pAig) - Saig_ManConstrNum(p.pAig) {
                continue; }
            Ssw_FramesWithClasses_rec( p, Aig_ObjFanin0(pObj), f );
//            if ( Aig_Regular(Ssw_ObjChild0Fra(p,pObj,f)) == Aig_ManConst1(p->pFrames) )
            if Ssw_ObjChild0Fra(p,pObj,f) == Aig_ManConst0(p.pFrames) {
                continue; }
            assert(Ssw_ObjChild0Fra(p,pObj,f) != Aig_ManConst1(p.pFrames), "Ssw_ObjChild0Fra(p,pObj,f) != Aig_ManConst1(p.pFrames)");
            if Ssw_ObjChild0Fra(p,pObj,f) == Aig_ManConst1(p.pFrames) {
                Abc_Print( 1, "Polarity violation.\n" );
                continue;
            }
            Ssw_NodesAreConstrained( p, Ssw_ObjChild0Fra(p,pObj,f), Aig_ManConst0(p.pFrames) );
        }
    }
    f = p.pPars.nFramesK;
    // clean the solver
    sat_solver_simplify( p.pMSat.pSat );


    // sweep internal nodes
    p.fRefined = 0;
    Ssw_ClassesClearRefined( p.ppClasses );
    if p.pPars.fVerbose {
        pProgress = Bar_ProgressStart( os.Stdout, Aig_ManObjNumMax(p.pAig) ); }
    // inlining of macro Aig_ManForEachObj
    // inlining of macro Vec_PtrForEachEntry
    for  i = 0; (i < Vec_PtrSize(p.pAig.vObjs)) && (((pObj) = (Aig_Obj_t *)Vec_PtrEntry(p.pAig.vObjs, i)), 1); i++  { if (pObj) == nil {} else {
        if p.pPars.fVerbose {
            Bar_ProgressUpdate( pProgress, i, nil ); }
        if Saig_ObjIsLo(p.pAig, pObj) {
            p.fRefined |= Ssw_ManSweepNodeConstr( p, pObj, f, 0 );
        } else if Aig_ObjIsNode(pObj) {
            pObjNew = Aig_And( p.pFrames, Ssw_ObjChild0Fra(p, pObj, f), Ssw_ObjChild1Fra(p, pObj, f) );
            Ssw_ObjSetFrame( p, pObj, f, pObjNew );
            p.fRefined |= Ssw_ManSweepNodeConstr( p, pObj, f, 0 );
        }
    }
    }
    if p.pPars.fVerbose {
        Bar_ProgressStop( pProgress ); }
    // cleanup
//    Ssw_ClassesCheck( p->ppClasses );
    return p.fRefined;
}

////////////////////////////////////////////////////////////////////////
///                       END OF FILE                                ///
////////////////////////////////////////////////////////////////////////


//ABC_NAMESPACE_IMPL_END
