package if_

/**CFile****************************************************************

  FileName    [ifTime.c]

  SystemName  [ABC: Logic synthesis and verification system.]

  PackageName [FPGA mapping based on priority cuts.]

  Synopsis    [Computation of delay paramters depending on the library.]

  Author      [Alan Mishchenko]
  
  Affiliation [UC Berkeley]

  Date        [Ver. 1.0. Started - November 21, 2006.]

  Revision    [$Id: ifTime.c,v 1.00 2006/11/21 00:00:00 alanmi Exp $]

***********************************************************************/

//#include "if.h"

//ABC_NAMESPACE_IMPL_START

////////////////////////////////////////////////////////////////////////
///                        DECLARATIONS                              ///
////////////////////////////////////////////////////////////////////////

////////////////////////////////////////////////////////////////////////
///                     FUNCTION DEFINITIONS                         ///
////////////////////////////////////////////////////////////////////////

/**Function*************************************************************

  Synopsis    [Sorts the pins in the decreasing order of delays.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func If_CutSortInputPins(p *If_Man_t, pCut *If_Cut_t, pPinPerm *int32, pPinDelays *float32) {
    var pLeaf *If_Obj_t;
    var i, j, best_i, temp int32;
    // start the trivial permutation and collect pin delays
    // inlining of macro If_CutForEachLeaf
    for  i = 0; (i < int32((pCut).nLeaves)) && ((pLeaf) = If_ManObj(p, (pCut).pLeaves[i])); i++  {
        pPinPerm[i] = i;
        pPinDelays[i] = If_ObjCutBest(pLeaf).Delay;
    }
    // selection sort the pins in the decreasible order of delays
    // this order will match the increasing order of LUT input pins
    for  i = 0; i < int32(pCut.nLeaves)-1; i++  {
        best_i = i;
        for  j = i+1; j < int32(pCut.nLeaves); j++  {
            if pPinDelays[pPinPerm[j]] > pPinDelays[pPinPerm[best_i]] {
                best_i = j; }
        }
        if best_i == i {
            continue; }
        temp = pPinPerm[i]; 
        pPinPerm[i] = pPinPerm[best_i]; 
        pPinPerm[best_i] = temp;
    }
/*
    // verify
    assert( pPinPerm[0] < (int)pCut->nLeaves );
    for ( i = 1; i < (int)pCut->nLeaves; i++ )
    {
        assert( pPinPerm[i] < (int)pCut->nLeaves );
        assert( pPinDelays[pPinPerm[i-1]] >= pPinDelays[pPinPerm[i]] );
    }
*/
}


/**Function*************************************************************

  Synopsis    [Computes delay.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
var spStatic_If_CutDelay_pPinPerm [IF_MAX_LUTSIZE]int32;
var spStatic_If_CutDelay_pPinDelays [IF_MAX_LUTSIZE]float32;
func If_CutDelay(p *If_Man_t, pObj *If_Obj_t, pCut *If_Cut_t) float32 {
    
    
    var pPerm string  = If_CutPerm( pCut );
    var pLeaf *If_Obj_t;
    var Delay, DelayCur float32;
    var pLutDelays *float32;
    var i, Shift, Pin2PinDelay int32;//, iLeaf;
    Delay = -IF_FLOAT_LARGE;
    if pCut.fAndCut {
        // inlining of macro If_CutForEachLeaf
        for  i = 0; (i < int32((pCut).nLeaves)) && ((pLeaf) = If_ManObj(p, (pCut).pLeaves[i])); i++  {
            DelayCur = If_ObjCutBest(pLeaf).Delay + p.pPars.nAndDelay;
            // inlining of macro IF_MAX
            
            if (Delay) > (DelayCur) {
                Delay = (Delay);
            } else {
                Delay = (DelayCur);
            }
        }
    
    } else if p.pPars.pLutLib {
        assert(!p.pPars.fLiftLeaves, "!p.pPars.fLiftLeaves");
        pLutDelays = p.pPars.pLutLib.pLutDelays[pCut.nLeaves];
        if p.pPars.pLutLib.fVarPinDelays {
            // compute the delay using sorted pins
            If_CutSortInputPins( p, pCut, spStatic_If_CutDelay_pPinPerm, spStatic_If_CutDelay_pPinDelays );
            for  i = 0; i < int32(pCut.nLeaves); i++  {
                DelayCur = spStatic_If_CutDelay_pPinDelays[spStatic_If_CutDelay_pPinPerm[i]] + pLutDelays[i];
                // inlining of macro IF_MAX
                
                if (Delay) > (DelayCur) {
                    Delay = (Delay);
                } else {
                    Delay = (DelayCur);
                }
            }
        
        } else {
            // inlining of macro If_CutForEachLeaf
            for  i = 0; (i < int32((pCut).nLeaves)) && ((pLeaf) = If_ManObj(p, (pCut).pLeaves[i])); i++  {
                DelayCur = If_ObjCutBest(pLeaf).Delay + pLutDelays[0];
                // inlining of macro IF_MAX
                
                if (Delay) > (DelayCur) {
                    Delay = (Delay);
                } else {
                    Delay = (DelayCur);
                }
            }
        }
    
    } else {
        if pCut.fUser {
            assert(!p.pPars.fLiftLeaves, "!p.pPars.fLiftLeaves");
            // inlining of macro If_CutForEachLeaf
            for  i = 0; (i < int32((pCut).nLeaves)) && ((pLeaf) = If_ManObj(p, (pCut).pLeaves[i])); i++  {
                if pPerm {
                    if pPerm[i] == IF_BIG_CHAR {
                        Pin2PinDelay = -IF_BIG_CHAR;
                    } else {
                        Pin2PinDelay = pPerm[i];
                    }
                } else {
                    Pin2PinDelay = 1;
                }
                DelayCur = If_ObjCutBest(pLeaf).Delay + float32(Pin2PinDelay);
                // inlining of macro IF_MAX
                
                if (Delay) > (DelayCur) {
                    Delay = (Delay);
                } else {
                    Delay = (DelayCur);
                }
            }
        
        } else {
            if p.pPars.fLiftLeaves {
                // inlining of macro If_CutForEachLeafSeq
                for  i = 0; (i < int32((pCut).nLeaves)) && ((pLeaf) = If_ManObj(p, (pCut).pLeaves[i] >> 8)) && (((Shift) = ((pCut).pLeaves[i] & 255)) >= 0); i++  {
                    DelayCur = If_ObjCutBest(pLeaf).Delay - Shift * p.Period;
                    // inlining of macro IF_MAX
                    
                    if (Delay) > (DelayCur + 1.0) {
                        Delay = (Delay);
                    } else {
                        Delay = (DelayCur + 1.0);
                    }
                }
            
            } else {
                // inlining of macro If_CutForEachLeaf
                for  i = 0; (i < int32((pCut).nLeaves)) && ((pLeaf) = If_ManObj(p, (pCut).pLeaves[i])); i++  {
                    DelayCur = If_ObjCutBest(pLeaf).Delay + 1.0;
                    // inlining of macro IF_MAX
                    
                    if (Delay) > (DelayCur) {
                        Delay = (Delay);
                    } else {
                        Delay = (DelayCur);
                    }
                }
            }
        }
    }
    return Delay;
}

/**Function*************************************************************

  Synopsis    []

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
var spStatic_If_CutPropagateRequired_pPinPerm [IF_MAX_LUTSIZE]int32;
var spStatic_If_CutPropagateRequired_pPinDelays [IF_MAX_LUTSIZE]float32;
func If_CutPropagateRequired(p *If_Man_t, pObj *If_Obj_t, pCut *If_Cut_t, ObjRequired float32) {
    
    
    var pLeaf *If_Obj_t;
    var pLutDelays *float32;
    var Required float32;
    var i, Pin2PinDelay int32;//, iLeaf;
    assert(!p.pPars.fLiftLeaves, "!p.pPars.fLiftLeaves");
    // compute the pins
    if pCut.fAndCut {
        // inlining of macro If_CutForEachLeaf
        for  i = 0; (i < int32((pCut).nLeaves)) && ((pLeaf) = If_ManObj(p, (pCut).pLeaves[i])); i++  {
            // inlining of macro IF_MIN
            
            if (pLeaf.Required) < (ObjRequired - p.pPars.nAndDelay) {
                pLeaf.Required = (pLeaf.Required);
            } else {
                pLeaf.Required = (ObjRequired - p.pPars.nAndDelay);
            }
        }
    
    } else if p.pPars.pLutLib {
        pLutDelays = p.pPars.pLutLib.pLutDelays[pCut.nLeaves];
        if p.pPars.pLutLib.fVarPinDelays {
            // compute the delay using sorted pins
            If_CutSortInputPins( p, pCut, spStatic_If_CutPropagateRequired_pPinPerm, spStatic_If_CutPropagateRequired_pPinDelays );
            for  i = 0; i < int32(pCut.nLeaves); i++  {
                Required = ObjRequired - pLutDelays[i];
                pLeaf = If_ManObj( p, pCut.pLeaves[spStatic_If_CutPropagateRequired_pPinPerm[i]] );
                // inlining of macro IF_MIN
                
                if (pLeaf.Required) < (Required) {
                    pLeaf.Required = (pLeaf.Required);
                } else {
                    pLeaf.Required = (Required);
                }
            }
        
        } else {
            Required = ObjRequired;
            // inlining of macro If_CutForEachLeaf
            for  i = 0; (i < int32((pCut).nLeaves)) && ((pLeaf) = If_ManObj(p, (pCut).pLeaves[i])); i++  {
                // inlining of macro IF_MIN
                
                if (pLeaf.Required) < (Required - pLutDelays[0]) {
                    pLeaf.Required = (pLeaf.Required);
                } else {
                    pLeaf.Required = (Required - pLutDelays[0]);
                }
            }
        }
    
    } else {
        if pCut.fUser {
            var Perm [IF_MAX_FUNC_LUTSIZE]byte;
            var pPerm string = Perm;
            if p.pPars.fDelayOpt {
                var Delay int32  = If_CutSopBalancePinDelays( p, pCut, pPerm );
                assert(Delay == int32(pCut.Delay), "Delay == int32(pCut.Delay)");
            
            } else if p.pPars.fDelayOptLut {
                var Delay int32  = If_CutLutBalancePinDelays( p, pCut, pPerm );
                assert(Delay == int32(pCut.Delay), "Delay == int32(pCut.Delay)");
            
            } else if p.pPars.fDsdBalance {
                var Delay int32  = If_CutDsdBalancePinDelays( p, pCut, pPerm );
                assert(Delay == int32(pCut.Delay), "Delay == int32(pCut.Delay)");
            
            } else {
                pPerm = If_CutPerm(pCut); }
            // inlining of macro If_CutForEachLeaf
            for  i = 0; (i < int32((pCut).nLeaves)) && ((pLeaf) = If_ManObj(p, (pCut).pLeaves[i])); i++  {
                if pPerm {
                    if pPerm[i] == IF_BIG_CHAR {
                        Pin2PinDelay = -IF_BIG_CHAR;
                    } else {
                        Pin2PinDelay = pPerm[i];
                    }
                } else {
                    Pin2PinDelay = 1;
                }
                Required = ObjRequired - float32(Pin2PinDelay);
                // inlining of macro IF_MIN
                
                if (pLeaf.Required) < (Required) {
                    pLeaf.Required = (pLeaf.Required);
                } else {
                    pLeaf.Required = (Required);
                }
            }
        
        } else {
            Required = ObjRequired;
            // inlining of macro If_CutForEachLeaf
            for  i = 0; (i < int32((pCut).nLeaves)) && ((pLeaf) = If_ManObj(p, (pCut).pLeaves[i])); i++  {
                // inlining of macro IF_MIN
                
                if (pLeaf.Required) < (Required - float32(1.0)) {
                    pLeaf.Required = (pLeaf.Required);
                } else {
                    pLeaf.Required = (Required - float32(1.0));
                }
            }
        }
    }
}

/**Function*************************************************************

  Synopsis    [Returns the max delay of the POs.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func If_ManDelayMax(p *If_Man_t, fSeq int32) float32 {
    var pObj *If_Obj_t;
    var DelayBest float32;
    var i int32;
    if p.pPars.fLatchPaths && (p.pPars.nLatchesCi == 0 || p.pPars.nLatchesCo == 0) {
        Abc_Print( 0, "Delay optimization of latch path is not performed because there is no latches.\n" );
        p.pPars.fLatchPaths = 0;
    }
    DelayBest = -IF_FLOAT_LARGE;
    if fSeq != 0 {
        assert(p.pPars.nLatchesCi > 0, "p.pPars.nLatchesCi > 0");
        // inlining of macro If_ManForEachPo
        // inlining of macro Vec_PtrForEachEntryStartStop
        for  i = p.pPars.nLatchesCoBox; (i < If_ManCoNum(p) - p.pPars.nLatchesCo) && (((pObj) = (If_Obj_t *)Vec_PtrEntry(p.vCos, i)), 1); i++  {
            if DelayBest < If_ObjArrTime(If_ObjFanin0(pObj)) {
                 DelayBest = If_ObjArrTime(If_ObjFanin0(pObj)); }
        }
    
    } else if p.pPars.fLatchPaths {
        // inlining of macro If_ManForEachLatchInput
        // inlining of macro Vec_PtrForEachEntryStart
        for  i = If_ManCoNum(p) - p.pPars.nLatchesCo; (i < Vec_PtrSize(p.vCos)) && (((pObj) = (If_Obj_t *)Vec_PtrEntry(p.vCos, i)), 1); i++  {
            if DelayBest < If_ObjArrTime(If_ObjFanin0(pObj)) {
                 DelayBest = If_ObjArrTime(If_ObjFanin0(pObj)); }
        }
    
    } else {
        // inlining of macro If_ManForEachCo
        // inlining of macro Vec_PtrForEachEntry
        for  i = 0; (i < Vec_PtrSize(p.vCos)) && (((pObj) = (If_Obj_t *)Vec_PtrEntry(p.vCos, i)), 1); i++  {
            if DelayBest < If_ObjArrTime(If_ObjFanin0(pObj)) {
                 DelayBest = If_ObjArrTime(If_ObjFanin0(pObj)); }
        }
    }
    return DelayBest;
}

/**Function*************************************************************

  Synopsis    [Computes the required times of all nodes.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func If_ManComputeRequired(p *If_Man_t) {
    var pObj *If_Obj_t;
    var i, Counter int32;
    var reqTime float32;

    // compute area, clean required times, collect nodes used in the mapping
//    p->AreaGlo = If_ManScanMapping( p );
    If_ManMarkMapping( p );
    if p.pManTim == nil {
        // get the global required times
        p.RequiredGlo = If_ManDelayMax( p, 0 );

        // consider the case when the required times are given
        if p.pPars.pTimesReq && !p.pPars.fAreaOnly {
            // make sure that the required time hold
            Counter = 0;
            // inlining of macro If_ManForEachCo
            // inlining of macro Vec_PtrForEachEntry
            for  i = 0; (i < Vec_PtrSize(p.vCos)) && (((pObj) = (If_Obj_t *)Vec_PtrEntry(p.vCos, i)), 1); i++  {
                if If_ObjArrTime(If_ObjFanin0(pObj)) > p.pPars.pTimesReq[i] + p.fEpsilon {
                    If_ObjFanin0(pObj).Required = If_ObjArrTime(If_ObjFanin0(pObj));
                    Counter++;
    //                Abc_Print( 0, "Required times are violated for output %d (arr = %d; req = %d).\n", 
    //                    i, (int)If_ObjArrTime(If_ObjFanin0(pObj)), (int)p->pPars->pTimesReq[i] );
                
                } else {
                    If_ObjFanin0(pObj).Required = p.pPars.pTimesReq[i]; }
            }
            if (Counter != 0) && !p.fReqTimeWarn {
                Abc_Print( 0, "Required times are exceeded at %d output%s. The earliest arrival times are used.\n", Counter, Counter > 1 ? "s":"" );
                p.fReqTimeWarn = 1;
            }
        
        } else {
            // find new delay target
            if p.pPars.nRelaxRatio && p.pPars.DelayTargetNew == 0 {
                p.pPars.DelayTargetNew = p.RequiredGlo * (100.0 + p.pPars.nRelaxRatio) / 100.0; } 

            // update the required times according to the target
            if p.pPars.DelayTarget != -1 {
                if p.RequiredGlo > p.pPars.DelayTarget + p.fEpsilon {
                    if p.fNextRound == 0 {
                        p.fNextRound = 1;
                        Abc_Print( 0, "Cannot meet the target required times (%4.2f). Mapping continues anyway.\n", p.pPars.DelayTarget );
                    }
                
                } else if p.RequiredGlo < p.pPars.DelayTarget - p.fEpsilon {
                    if p.fNextRound == 0 {
                        p.fNextRound = 1;
//                        Abc_Print( 0, "Relaxing the required times from (%4.2f) to the target (%4.2f).\n", p->RequiredGlo, p->pPars->DelayTarget );
                    }
                    p.RequiredGlo = p.pPars.DelayTarget;
                }
            
            } else if p.pPars.DelayTargetNew > 0 { // relax the required times 
                p.RequiredGlo = p.pPars.DelayTargetNew; }
            // do not propagate required times if area minimization is requested
            if p.pPars.fAreaOnly { 
                return; }
            // set the required times for the POs
            if p.pPars.fDoAverage {
                if p.pPars.nRelaxRatio {
                    // inlining of macro If_ManForEachCo
                    // inlining of macro Vec_PtrForEachEntry
                    for  i = 0; (i < Vec_PtrSize(p.vCos)) && (((pObj) = (If_Obj_t *)Vec_PtrEntry(p.vCos, i)), 1); i++  {
                        If_ObjFanin0(pObj).Required = If_ObjArrTime(If_ObjFanin0(pObj)) * (100.0 + p.pPars.nRelaxRatio) / 100.0;
                    }
                
                } else {
                    // inlining of macro If_ManForEachCo
                    // inlining of macro Vec_PtrForEachEntry
                    for  i = 0; (i < Vec_PtrSize(p.vCos)) && (((pObj) = (If_Obj_t *)Vec_PtrEntry(p.vCos, i)), 1); i++  {
                        If_ObjFanin0(pObj).Required = If_ObjArrTime(If_ObjFanin0(pObj));
                    }
                }
            
            } else if p.pPars.fLatchPaths {
                // inlining of macro If_ManForEachLatchInput
                // inlining of macro Vec_PtrForEachEntryStart
                for  i = If_ManCoNum(p) - p.pPars.nLatchesCo; (i < Vec_PtrSize(p.vCos)) && (((pObj) = (If_Obj_t *)Vec_PtrEntry(p.vCos, i)), 1); i++  {
                    If_ObjFanin0(pObj).Required = p.RequiredGlo;
                }
            
            } else {
                // inlining of macro If_ManForEachCo
                // inlining of macro Vec_PtrForEachEntry
                for  i = 0; (i < Vec_PtrSize(p.vCos)) && (((pObj) = (If_Obj_t *)Vec_PtrEntry(p.vCos, i)), 1); i++  {
                    If_ObjFanin0(pObj).Required = p.RequiredGlo;
                }
            }
        }
        // go through the nodes in the reverse topological order
    //    Vec_PtrForEachEntry( If_Obj_t *, p->vMapped, pObj, i )
    //        If_CutPropagateRequired( p, pObj, If_ObjCutBest(pObj), pObj->Required );
        // inlining of macro If_ManForEachObjReverse
        // inlining of macro Vec_PtrForEachEntry
        for  i = 0; (i < Vec_PtrSize(p.vObjsRev)) && (((pObj) = (If_Obj_t *)Vec_PtrEntry(p.vObjsRev, i)), 1); i++  {
            if pObj.nRefs == 0 {
                continue; }
            If_CutPropagateRequired( p, pObj, If_ObjCutBest(pObj), pObj.Required );
        }
    
    } else {
        // get the global required times
        p.RequiredGlo = If_ManDelayMax( p, 0 );

        // find new delay target
        if p.pPars.nRelaxRatio && p.pPars.DelayTargetNew == 0 {
            p.pPars.DelayTargetNew = p.RequiredGlo * (100.0 + p.pPars.nRelaxRatio) / 100.0; } 

        // update the required times according to the target
        if p.pPars.DelayTarget != -1 {
            if p.RequiredGlo > p.pPars.DelayTarget + p.fEpsilon {
                if p.fNextRound == 0 {
                    p.fNextRound = 1;
                    Abc_Print( 0, "Cannot meet the target required times (%4.2f). Mapping continues anyway.\n", p.pPars.DelayTarget );
                }
            
            } else if p.RequiredGlo < p.pPars.DelayTarget - p.fEpsilon {
                if p.fNextRound == 0 {
                    p.fNextRound = 1;
//                    Abc_Print( 0, "Relaxing the required times from (%4.2f) to the target (%4.2f).\n", p->RequiredGlo, p->pPars->DelayTarget );
                }
                p.RequiredGlo = p.pPars.DelayTarget;
            }
        
        } else if p.pPars.DelayTargetNew > 0 { // relax the required times 
            p.RequiredGlo = p.pPars.DelayTargetNew; }

        // do not propagate required times if area minimization is requested
        if p.pPars.fAreaOnly { 
            return; }
        // set the required times for the POs
        Tim_ManIncrementTravId( p.pManTim );
        if p.vCoAttrs {
            assert(If_ManCoNum(p) == Vec_IntSize(p.vCoAttrs), "If_ManCoNum(p) == Vec_IntSize(p.vCoAttrs)");
            // inlining of macro If_ManForEachCo
            // inlining of macro Vec_PtrForEachEntry
            for  i = 0; (i < Vec_PtrSize(p.vCos)) && (((pObj) = (If_Obj_t *)Vec_PtrEntry(p.vCos, i)), 1); i++  { 
                if Vec_IntEntry(p.vCoAttrs, i) == -1 {       // -1=internal
                    continue; }
                if Vec_IntEntry(p.vCoAttrs, i) == 0 {        //  0=optimize
                    Tim_ManSetCoRequired( p.pManTim, i, p.RequiredGlo );
                } else if Vec_IntEntry(p.vCoAttrs, i) == 1 {   //  1=keep
                    Tim_ManSetCoRequired( p.pManTim, i, If_ObjArrTime(If_ObjFanin0(pObj)) );
                } else if Vec_IntEntry(p.vCoAttrs, i) == 2 {   //  2=relax
                    Tim_ManSetCoRequired( p.pManTim, i, IF_FLOAT_LARGE );
                } else { assert(false, "0"); }
            }
        
        } else if p.pPars.fDoAverage {
            if p.pPars.nRelaxRatio {
                // inlining of macro If_ManForEachCo
                // inlining of macro Vec_PtrForEachEntry
                for  i = 0; (i < Vec_PtrSize(p.vCos)) && (((pObj) = (If_Obj_t *)Vec_PtrEntry(p.vCos, i)), 1); i++  {
                    Tim_ManSetCoRequired( p.pManTim, i, If_ObjArrTime(If_ObjFanin0(pObj)) * (100.0 + p.pPars.nRelaxRatio) / 100.0 );
                }
            
            } else {
                // inlining of macro If_ManForEachCo
                // inlining of macro Vec_PtrForEachEntry
                for  i = 0; (i < Vec_PtrSize(p.vCos)) && (((pObj) = (If_Obj_t *)Vec_PtrEntry(p.vCos, i)), 1); i++  {
                    Tim_ManSetCoRequired( p.pManTim, i, If_ObjArrTime(If_ObjFanin0(pObj)) );
                }
            }
        
        } else if p.pPars.fLatchPaths {
            // inlining of macro If_ManForEachPo
            // inlining of macro Vec_PtrForEachEntryStartStop
            for  i = p.pPars.nLatchesCoBox; (i < If_ManCoNum(p) - p.pPars.nLatchesCo) && (((pObj) = (If_Obj_t *)Vec_PtrEntry(p.vCos, i)), 1); i++  {
                Tim_ManSetCoRequired( p.pManTim, i, IF_FLOAT_LARGE );
            }
            // inlining of macro If_ManForEachLatchInput
            // inlining of macro Vec_PtrForEachEntryStart
            for  i = If_ManCoNum(p) - p.pPars.nLatchesCo; (i < Vec_PtrSize(p.vCos)) && (((pObj) = (If_Obj_t *)Vec_PtrEntry(p.vCos, i)), 1); i++  {
                Tim_ManSetCoRequired( p.pManTim, i, p.RequiredGlo );
            }
        
        } else {
            Tim_ManInitPoRequiredAll( p.pManTim, p.RequiredGlo );
//            If_ManForEachCo( p, pObj, i )
//                Tim_ManSetCoRequired( p->pManTim, pObj->IdPio, p->RequiredGlo );
        }
        // go through the nodes in the reverse topological order
        // inlining of macro If_ManForEachObjReverse
        // inlining of macro Vec_PtrForEachEntry
        for  i = 0; (i < Vec_PtrSize(p.vObjsRev)) && (((pObj) = (If_Obj_t *)Vec_PtrEntry(p.vObjsRev, i)), 1); i++  {
            if If_ObjIsAnd(pObj) {
                if pObj.nRefs == 0 {
                    continue; }
                If_CutPropagateRequired( p, pObj, If_ObjCutBest(pObj), pObj.Required );
            
            } else if If_ObjIsCi(pObj) {
                reqTime = pObj.Required;
                Tim_ManSetCiRequired( p.pManTim, pObj.IdPio, reqTime );
            
            } else if If_ObjIsCo(pObj) {
                reqTime = Tim_ManGetCoRequired( p.pManTim, pObj.IdPio );
                // inlining of macro IF_MIN
                If_ObjFanin0(pObj).Required = (((reqTime) < (If_ObjFanin0(pObj).Required))? (reqTime) : (If_ObjFanin0(pObj).Required));
            
            } else if If_ObjIsConst1(pObj) {
            
            } else { // add the node to the mapper
                assert(false, "0"); }
        }
    }
}

////////////////////////////////////////////////////////////////////////
///                       END OF FILE                                ///
////////////////////////////////////////////////////////////////////////


//ABC_NAMESPACE_IMPL_END

