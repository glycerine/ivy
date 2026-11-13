package saig

import "time"

/**CFile****************************************************************

  FileName    [saigTrans.c]

  SystemName  [ABC: Logic synthesis and verification system.]

  PackageName [Sequential AIG package.]

  Synopsis    [Dynamic simplication of the transition relation.]

  Author      [Alan Mishchenko]
  
  Affiliation [UC Berkeley]

  Date        [Ver. 1.0. Started - June 20, 2005.]

  Revision    [$Id: saigTrans.c,v 1.00 2005/06/20 00:00:00 alanmi Exp $]

***********************************************************************/

//#include "saig.h"
//#include "proof/fra/fra.h"

//ABC_NAMESPACE_IMPL_START


/* 
    A similar approach is presented in the his paper:
    A. Kuehlmann. Dynamic transition relation simplification for 
    bounded property checking. ICCAD'04, pp. 50-57.
*/

////////////////////////////////////////////////////////////////////////
///                        DECLARATIONS                              ///
////////////////////////////////////////////////////////////////////////

////////////////////////////////////////////////////////////////////////
///                     FUNCTION DEFINITIONS                         ///
////////////////////////////////////////////////////////////////////////

/**Function*************************************************************

  Synopsis    [Maps a node/frame into a node of a different manager.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
/**Function*************************************************************

  Synopsis    [Create mapping for the first nFrames timeframes of pAig.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Saig_ManCreateMapping(pAig *Aig_Man_t, pFrames *Aig_Man_t, nFrames int32) {
    var pObj, pObjFrame, pObjRepr *Aig_Obj_t;
    var i, f, iNum, iFrame int32;
    assert(pFrames.pReprs != nil, "pFrames.pReprs != nil"); // mapping from nodes into their representatives
    // start step mapping for both orignal manager and fraig
    Saig_ManStartMap2( pAig, nFrames );
    Saig_ManStartMap2( pFrames, 1 );
    // for each object in each frame
    for  f = 0; f < nFrames; f++  {
    // inlining of macro Aig_ManForEachObj
    // inlining of macro Vec_PtrForEachEntry
    for  i = 0; (i < Vec_PtrSize(pAig.vObjs)) && (((pObj) = (Aig_Obj_t *)Vec_PtrEntry(pAig.vObjs, i)), 1); i++  { if (pObj) == nil {} else {
        // get the frame object
        iNum = Saig_ManGetMap1( pAig, pObj, f );
        pObjFrame = Aig_ManObj( pFrames, iNum );
        // if the node has no prototype, map it into itself
        if pObjFrame == nil {
            Saig_ManSetMap2( pAig, pObj, f, pObj, f );
            continue;
        }
        // get the representative object
        pObjRepr = Aig_ObjRepr( pFrames, pObjFrame );
        if pObjRepr == nil {
            pObjRepr = pObjFrame; }
        // check if this is the first time this object is reached
        if Saig_ManGetMap2( pFrames, pObjRepr, 0, &iFrame ) == -1 {
            Saig_ManSetMap2( pFrames, pObjRepr, 0, pObj, f ); }
        // set the map for the main object
        iNum = Saig_ManGetMap2( pFrames, pObjRepr, 0, &iFrame );
        Saig_ManSetMap2( pAig, pObj, f, Aig_ManObj(pAig, iNum), iFrame );
    }
    }
    }
    Saig_ManStopMap2( pFrames );
    assert((Saig_ManHasMap2(pAig) != 0), "Saig_ManHasMap2(pAig)");
}

/**Function*************************************************************

  Synopsis    [Unroll without initialization.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Saig_ManFramesNonInitial(pAig *Aig_Man_t, nFrames int32) *Aig_Man_t {
    var pFrames *Aig_Man_t;
    var pObj, pObjLi, pObjLo *Aig_Obj_t;
    var i, f int32;
    assert(Saig_ManRegNum(pAig) > 0, "Saig_ManRegNum(pAig) > 0");
    // start node map
    Saig_ManStartMap1( pAig, nFrames ); 
    // start the new manager
    pFrames = Aig_ManStart( Aig_ManNodeNum(pAig) * nFrames );
    // map the constant node
    Aig_ManConst1(pAig).pData = Aig_ManConst1( pFrames );
    // create variables for register outputs
    // inlining of macro Saig_ManForEachLo
    for  i = 0; (i < Saig_ManRegNum(pAig)) && (((pObj) = (Aig_Obj_t *)Vec_PtrEntry(pAig.vCis, i+Saig_ManPiNum(pAig))), 1); i++  {
        pObj.pData = Aig_ObjCreateCi( pFrames );
    }
    // add timeframes
    for  f = 0; f < nFrames; f++  {
        // create PI nodes for this frame
        // inlining of macro Saig_ManForEachPi
        // inlining of macro Vec_PtrForEachEntryStop
        for  i = 0; (i < Saig_ManPiNum(pAig)) && (((pObj) = (Aig_Obj_t *)Vec_PtrEntry(pAig.vCis, i)), 1); i++  {
            pObj.pData = Aig_ObjCreateCi( pFrames );
        }
        // add internal nodes of this frame
        // inlining of macro Aig_ManForEachNode
        // inlining of macro Vec_PtrForEachEntry
        for  i = 0; (i < Vec_PtrSize(pAig.vObjs)) && (((pObj) = (Aig_Obj_t *)Vec_PtrEntry(pAig.vObjs, i)), 1); i++  { if (pObj) == nil || !Aig_ObjIsNode(pObj) {} else {
            pObj.pData = Aig_And( pFrames, Aig_ObjChild0Copy(pObj), Aig_ObjChild1Copy(pObj) ); }
        }
        // create POs for this frame
        // inlining of macro Saig_ManForEachPo
        // inlining of macro Vec_PtrForEachEntryStop
        for  i = 0; (i < Saig_ManPoNum(pAig)) && (((pObj) = (Aig_Obj_t *)Vec_PtrEntry(pAig.vCos, i)), 1); i++  {
            pObj.pData = Aig_ObjCreateCo( pFrames, Aig_ObjChild0Copy(pObj) );
        }
        // save register inputs
        // inlining of macro Saig_ManForEachLi
        for  i = 0; (i < Saig_ManRegNum(pAig)) && (((pObj) = (Aig_Obj_t *)Vec_PtrEntry(pAig.vCos, i+Saig_ManPoNum(pAig))), 1); i++  {
            pObj.pData = Aig_ObjChild0Copy(pObj);
        }
        // save the mapping
        // inlining of macro Aig_ManForEachObj
        // inlining of macro Vec_PtrForEachEntry
        for  i = 0; (i < Vec_PtrSize(pAig.vObjs)) && (((pObj) = (Aig_Obj_t *)Vec_PtrEntry(pAig.vObjs, i)), 1); i++  { if (pObj) == nil {} else {
            assert(pObj.pData != nil, "pObj.pData != nil");
            Saig_ManSetMap1( pAig, pObj, f, Aig_Regular((Aig_Obj_t *)pObj.pData) );
        }
        }
        // quit if the last frame
        if f == nFrames - 1 {
            break; }
        // transfer to register outputs
        // inlining of macro Saig_ManForEachLiLo
        for  i = 0; (i < Saig_ManRegNum(pAig)) && (((pObjLi) = Saig_ManLi(pAig, i)), 1)            && (((pObjLo)=Saig_ManLo(pAig, i)), 1); i++  {
            pObjLo.pData = pObjLi.pData;
        }
    }
    // remember register outputs
    // inlining of macro Saig_ManForEachLiLo
    for  i = 0; (i < Saig_ManRegNum(pAig)) && (((pObjLi) = Saig_ManLi(pAig, i)), 1)            && (((pObjLo)=Saig_ManLo(pAig, i)), 1); i++  {
        Aig_ObjCreateCo( pFrames, (Aig_Obj_t *)pObjLi.pData );
    }
    Aig_ManCleanup( pFrames );
    return pFrames;
}

/**Function*************************************************************

  Synopsis    [Unroll with initialization and mapping.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Saig_ManFramesInitialMapped(pAig *Aig_Man_t, nFrames int32, nFramesMax int32, fInit int32) *Aig_Man_t {
    var pFrames *Aig_Man_t;
    var pObj, pObjLi, pObjLo, pRepr *Aig_Obj_t;
    var i, f, iNum1, iNum2, iFrame2 int32;
    assert(nFrames <= nFramesMax, "nFrames <= nFramesMax");
    assert(Saig_ManRegNum(pAig) > 0, "Saig_ManRegNum(pAig) > 0");
    // start node map
    Saig_ManStartMap1( pAig, nFramesMax );
    // start the new manager
    pFrames = Aig_ManStart( Aig_ManNodeNum(pAig) * nFramesMax );
    // create variables for register outputs
    if fInit != 0 {
        // inlining of macro Saig_ManForEachLo
        for  i = 0; (i < Saig_ManRegNum(pAig)) && (((pObj) = (Aig_Obj_t *)Vec_PtrEntry(pAig.vCis, i+Saig_ManPiNum(pAig))), 1); i++  {
            pObj.pData = Aig_ManConst0( pFrames );
            Saig_ManSetMap1( pAig, pObj, 0, Aig_Regular((Aig_Obj_t *)pObj.pData) );
        }
    
    } else {
        // create PIs first
        for  f = 0; f < nFramesMax; f++  {
            // inlining of macro Saig_ManForEachPi
            // inlining of macro Vec_PtrForEachEntryStop
            for  i = 0; (i < Saig_ManPiNum(pAig)) && (((pObj) = (Aig_Obj_t *)Vec_PtrEntry(pAig.vCis, i)), 1); i++  {
                Aig_ObjCreateCi( pFrames );
            }
        }
        // create registers second
        // inlining of macro Saig_ManForEachLo
        for  i = 0; (i < Saig_ManRegNum(pAig)) && (((pObj) = (Aig_Obj_t *)Vec_PtrEntry(pAig.vCis, i+Saig_ManPiNum(pAig))), 1); i++  {
            pObj.pData = Aig_ObjCreateCi( pFrames );
            Saig_ManSetMap1( pAig, pObj, 0, Aig_Regular((Aig_Obj_t *)pObj.pData) );
        }
    }
    // add timeframes
    for  f = 0; f < nFramesMax; f++  {
        // map the constant node
        pObj = Aig_ManConst1(pAig);
        pObj.pData = Aig_ManConst1( pFrames );
        Saig_ManSetMap1( pAig, pObj, f, Aig_Regular((Aig_Obj_t *)pObj.pData) );
        // create PI nodes for this frame
        // inlining of macro Saig_ManForEachPi
        // inlining of macro Vec_PtrForEachEntryStop
        for  i = 0; (i < Saig_ManPiNum(pAig)) && (((pObj) = (Aig_Obj_t *)Vec_PtrEntry(pAig.vCis, i)), 1); i++  {
            if fInit != 0 {
                pObj.pData = Aig_ObjCreateCi( pFrames );
            } else {
                pObj.pData = Aig_ManCi( pFrames, f * Saig_ManPiNum(pAig) + i ); }
            Saig_ManSetMap1( pAig, pObj, f, Aig_Regular((Aig_Obj_t *)pObj.pData) );
        }
        // add internal nodes of this frame
        // inlining of macro Aig_ManForEachNode
        // inlining of macro Vec_PtrForEachEntry
        for  i = 0; (i < Vec_PtrSize(pAig.vObjs)) && (((pObj) = (Aig_Obj_t *)Vec_PtrEntry(pAig.vObjs, i)), 1); i++  { if (pObj) == nil || !Aig_ObjIsNode(pObj) {} else {
            pObj.pData = Aig_And( pFrames, Aig_ObjChild0Copy(pObj), Aig_ObjChild1Copy(pObj) );
            Saig_ManSetMap1( pAig, pObj, f, Aig_Regular((Aig_Obj_t *)pObj.pData) );
            if !(Saig_ManHasMap2(pAig) != 0) {
                continue; }
            if f < nFrames {
                // get the mapping for this node
                iNum2 = Saig_ManGetMap2( pAig, pObj, f, &iFrame2 );
            
            } else {
                // get the mapping for this node
                iNum2 = Saig_ManGetMap2( pAig, pObj, nFrames-1, &iFrame2 );
                iFrame2 += f - (nFrames-1);
            }
            assert(iNum2 != -1, "iNum2 != -1");
            assert(f >= iFrame2, "f >= iFrame2");
            // get the corresponding frames node
            iNum1 = Saig_ManGetMap1( pAig, Aig_ManObj(pAig, iNum2), iFrame2 );
            pRepr = Aig_ManObj( pFrames, iNum1 );
            // compare the phases of these nodes
            pObj.pData = Aig_NotCond( pRepr, pRepr.fPhase ^ Aig_ObjPhaseReal((Aig_Obj_t *)pObj.pData) );
        }
        }
        // create POs for this frame
        // inlining of macro Saig_ManForEachPo
        // inlining of macro Vec_PtrForEachEntryStop
        for  i = 0; (i < Saig_ManPoNum(pAig)) && (((pObj) = (Aig_Obj_t *)Vec_PtrEntry(pAig.vCos, i)), 1); i++  {
            pObj.pData = Aig_ObjCreateCo( pFrames, Aig_ObjChild0Copy(pObj) );
            Saig_ManSetMap1( pAig, pObj, f, Aig_Regular((Aig_Obj_t *)pObj.pData) );
        }
        // save register inputs
        // inlining of macro Saig_ManForEachLi
        for  i = 0; (i < Saig_ManRegNum(pAig)) && (((pObj) = (Aig_Obj_t *)Vec_PtrEntry(pAig.vCos, i+Saig_ManPoNum(pAig))), 1); i++  {
            pObj.pData = Aig_ObjChild0Copy(pObj);
            Saig_ManSetMap1( pAig, pObj, f, Aig_Regular((Aig_Obj_t *)pObj.pData) );
        }
        // quit if the last frame
        if f == nFramesMax - 1 {
            break; }
        // transfer to register outputs
        // inlining of macro Saig_ManForEachLiLo
        for  i = 0; (i < Saig_ManRegNum(pAig)) && (((pObjLi) = Saig_ManLi(pAig, i)), 1)            && (((pObjLo)=Saig_ManLo(pAig, i)), 1); i++  {
            pObjLo.pData = pObjLi.pData;
            if !(fInit != 0) {
                Saig_ManSetMap1( pAig, pObjLo, f+1, Aig_Regular((Aig_Obj_t *)pObjLo.pData) ); }
        }
    }
    if !(fInit != 0) {
        // create registers
        // inlining of macro Saig_ManForEachLiLo
        for  i = 0; (i < Saig_ManRegNum(pAig)) && (((pObjLi) = Saig_ManLi(pAig, i)), 1)            && (((pObjLo)=Saig_ManLo(pAig, i)), 1); i++  {
            Aig_ObjCreateCo( pFrames, (Aig_Obj_t *)pObjLi.pData );
        }
        // set register number
        Aig_ManSetRegNum( pFrames, pAig.nRegs );
    }
    Aig_ManCleanup( pFrames );
    Saig_ManStopMap1( pAig );
    return pFrames;
}

/**Function*************************************************************

  Synopsis    [Implements dynamic simplification.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Saig_ManTimeframeSimplify(pAig *Aig_Man_t, nFrames int32, nFramesMax int32, fInit int32, fVerbose int32) *Aig_Man_t {
//    extern Aig_Man_t * Fra_FraigEquivence( Aig_Man_t * pManAig, int nConfMax, int fProve );
    var pFrames, pFraig, pRes1, pRes2 *Aig_Man_t;
    var clk abctime;
    // create uninitialized timeframes with map1
    pFrames = Saig_ManFramesNonInitial( pAig, nFrames );
    // perform fraiging for the unrolled timeframes
clk = Abc_Clock();
    pFraig = Fra_FraigEquivence( pFrames, 1000, 0 );
    // report the results
    if fVerbose != 0 {
        Aig_ManPrintStats( pFrames );
        Aig_ManPrintStats( pFraig );
// inlining of macro ABC_PRT
Abc_Print(1, "%s =", ("Fraiging"));
Abc_Print(1, "%9.2f sec\n", 1.0*(float64((Abc_Clock() - clk)))/(float64(int64(time.Second))));;
    }
    Aig_ManStop( pFraig );
    assert(pFrames.pReprs != nil, "pFrames.pReprs != nil");
    // create AIG with map2
    Saig_ManCreateMapping( pAig, pFrames, nFrames );
    Aig_ManStop( pFrames );
    Saig_ManStopMap1( pAig );
    // create reduced initialized timeframes
clk = Abc_Clock();
    pRes2 = Saig_ManFramesInitialMapped( pAig, nFrames, nFramesMax, fInit );
// inlining of macro ABC_PRT
Abc_Print(1, "%s =", ("Mapped"));
Abc_Print(1, "%9.2f sec\n", 1.0*(float64((Abc_Clock() - clk)))/(float64(int64(time.Second))));;
    // free mapping
    Saig_ManStopMap2( pAig );
clk = Abc_Clock();
    pRes1 = Saig_ManFramesInitialMapped( pAig, nFrames, nFramesMax, fInit );
// inlining of macro ABC_PRT
Abc_Print(1, "%s =", ("Normal"));
Abc_Print(1, "%9.2f sec\n", 1.0*(float64((Abc_Clock() - clk)))/(float64(int64(time.Second))));;
    // report the results
    if fVerbose != 0 {
        Aig_ManPrintStats( pRes1 );
        Aig_ManPrintStats( pRes2 );
    }
    Aig_ManStop( pRes1 );
    assert(!(Saig_ManHasMap1(pAig) != 0), "!Saig_ManHasMap1(pAig)");
    assert(!(Saig_ManHasMap2(pAig) != 0), "!Saig_ManHasMap2(pAig)");
    return pRes2;
}

////////////////////////////////////////////////////////////////////////
///                       END OF FILE                                ///
////////////////////////////////////////////////////////////////////////


//ABC_NAMESPACE_IMPL_END

