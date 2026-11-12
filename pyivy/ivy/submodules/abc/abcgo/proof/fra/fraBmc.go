package fra

import (
	"fmt"
	"time"
)

/**CFile****************************************************************

  FileName    [fraBmc.c]

  SystemName  [ABC: Logic synthesis and verification system.]

  PackageName [New FRAIG package.]

  Synopsis    [Bounded model checking.]

  Author      [Alan Mishchenko]
  
  Affiliation [UC Berkeley]

  Date        [Ver. 1.0. Started - June 30, 2007.]

  Revision    [$Id: fraBmc.c,v 1.00 2007/06/30 00:00:00 alanmi Exp $]

***********************************************************************/

//#include "fra.h"

//ABC_NAMESPACE_IMPL_START


////////////////////////////////////////////////////////////////////////
///                        DECLARATIONS                              ///
////////////////////////////////////////////////////////////////////////

// simulation manager
type Fra_Bmc_t_ struct {
    // parameters
    nPref              int32;             // the size of the prefix
    nDepth              int32;            // the depth of the frames
    nFramesAll              int32;        // the total number of timeframes
    // implications to be filtered
    vImps       *Vec_Int_t;
    // AIG managers
    pAig       *Aig_Man_t;              // the original AIG manager
    pAigFrames       *Aig_Man_t;        // initialized timeframes
    pAigFraig       *Aig_Man_t;         // the fraiged initialized timeframes
    // mapping of nodes
    pObjToFrames      **Aig_Obj_t;      // mapping of the original node into frames
    pObjToFraig      **Aig_Obj_t;       // mapping of the frames node into fraig
}

func Bmc_ObjFrames(pObj *Aig_Obj_t, i int32) *Aig_Obj_t { return ((Fra_Man_t *)pObj.pData).pBmc.pObjToFrames[((Fra_Man_t *)pObj.pData).pBmc.nFramesAll*pObj.Id + i];  }
func Bmc_ObjSetFrames(pObj *Aig_Obj_t, i int32, pNode *Aig_Obj_t) { ((Fra_Man_t *)pObj.pData).pBmc.pObjToFrames[((Fra_Man_t *)pObj.pData).pBmc.nFramesAll*pObj.Id + i] = pNode; }

func Bmc_ObjFraig(pObj *Aig_Obj_t) *Aig_Obj_t { return ((Fra_Man_t *)pObj.pData).pBmc.pObjToFraig[pObj.Id];  }
func Bmc_ObjSetFraig(pObj *Aig_Obj_t, pNode *Aig_Obj_t) { ((Fra_Man_t *)pObj.pData).pBmc.pObjToFraig[pObj.Id] = pNode; }

func Bmc_ObjChild0Frames(pObj *Aig_Obj_t, i int32) *Aig_Obj_t { assert(!Aig_IsComplement(pObj), "!Aig_IsComplement(pObj)"); if Aig_ObjFanin0(pObj) {
    return Aig_NotCond(Bmc_ObjFrames(Aig_ObjFanin0(pObj),i), Aig_ObjFaninC0(pObj));
} else {
    return nil;
}  }
func Bmc_ObjChild1Frames(pObj *Aig_Obj_t, i int32) *Aig_Obj_t { assert(!Aig_IsComplement(pObj), "!Aig_IsComplement(pObj)"); if Aig_ObjFanin1(pObj) {
    return Aig_NotCond(Bmc_ObjFrames(Aig_ObjFanin1(pObj),i), Aig_ObjFaninC1(pObj));
} else {
    return nil;
}  }

////////////////////////////////////////////////////////////////////////
///                     FUNCTION DEFINITIONS                         ///
////////////////////////////////////////////////////////////////////////

/**Function*************************************************************

  Synopsis    [Returns 1 if the nodes are equivalent.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Fra_BmcNodesAreEqual(pObj0 *Aig_Obj_t, pObj1 *Aig_Obj_t) int32 {
    var p *Fra_Man_t  = (Fra_Man_t *)pObj0.pData;
    var pObjFrames0, pObjFrames1 *Aig_Obj_t;
    var pObjFraig0, pObjFraig1 *Aig_Obj_t;
    var i int32;
    for  i = p.pBmc.nPref; i < p.pBmc.nFramesAll; i++  {
        pObjFrames0 = Aig_Regular( Bmc_ObjFrames(pObj0, i) );
        pObjFrames1 = Aig_Regular( Bmc_ObjFrames(pObj1, i) );
        if pObjFrames0 == pObjFrames1 {
            continue; }
        pObjFraig0 = Aig_Regular( Bmc_ObjFraig(pObjFrames0) );
        pObjFraig1 = Aig_Regular( Bmc_ObjFraig(pObjFrames1) );
        if pObjFraig0 != pObjFraig1 {
            return 0; }
    }
    return 1;
}

/**Function*************************************************************

  Synopsis    [Returns 1 if the node is costant.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Fra_BmcNodeIsConst(pObj *Aig_Obj_t) int32 {
    var p *Fra_Man_t  = (Fra_Man_t *)pObj.pData;
    return Fra_BmcNodesAreEqual( pObj, Aig_ManConst1(p.pManAig) );
}

/**Function*************************************************************

  Synopsis    [Refines implications using BMC.]

  Description [The input is the combinational FRAIG manager,
  which is used to FRAIG the timeframes. ]
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Fra_BmcFilterImplications(p *Fra_Man_t, pBmc *Fra_Bmc_t) {
    var pLeft, pRight *Aig_Obj_t;
    var pLeftT, pRightT *Aig_Obj_t;
    var pLeftF, pRightF *Aig_Obj_t;
    var i, f, Imp, Left, Right int32;
    var fComplL, fComplR int32;
    assert(p.nFramesAll == 1, "p.nFramesAll == 1");
    assert(p.pManAig == pBmc.pAigFrames, "p.pManAig == pBmc.pAigFrames");
    // inlining of macro Vec_IntForEachEntry
    for  i = 0; (i < Vec_IntSize(pBmc.vImps)) && (((Imp) = Vec_IntEntry(pBmc.vImps, i)), 1); i++  {
        if Imp == 0 {
            continue; }
        Left  = Fra_ImpLeft(Imp);
        Right = Fra_ImpRight(Imp);
        // get the corresponding nodes
        pLeft  = Aig_ManObj( pBmc.pAig, Left );
        pRight = Aig_ManObj( pBmc.pAig, Right );
        // iterate through the timeframes
        for  f = pBmc.nPref; f < pBmc.nFramesAll; f++  {
            // get timeframe nodes
            pLeftT  = Bmc_ObjFrames( pLeft, f );
            pRightT = Bmc_ObjFrames( pRight, f );
            // get the corresponding FRAIG nodes
            pLeftF  = Fra_ObjFraig( Aig_Regular(pLeftT), 0 );
            pRightF = Fra_ObjFraig( Aig_Regular(pRightT), 0 );
            // get the complemented attributes
            fComplL = pLeft.fPhase ^ Aig_IsComplement(pLeftF) ^ Aig_IsComplement(pLeftT);
            fComplR = pRight.fPhase ^ Aig_IsComplement(pRightF) ^ Aig_IsComplement(pRightT);
            // check equality
            if Aig_Regular(pLeftF) == Aig_Regular(pRightF) {
                if fComplL == fComplR { // x => x  - always true
                    continue; }
                assert(fComplL != fComplR, "fComplL != fComplR");
                // consider 4 possibilities:
                // NOT(1) => 1    or   0 => 1  - always true
                // 1 => NOT(1)    or   1 => 0  - never true
                // NOT(x) => x    or   x       - not always true
                // x => NOT(x)    or   NOT(x)  - not always true
                if Aig_ObjIsConst1(Aig_Regular(pLeftF)) && fComplL { // proved implication
                    continue; }
                // disproved implication
                Vec_IntWriteEntry( pBmc.vImps, i, 0 ); 
                break;
            }
            // check the implication 
            if Fra_NodesAreImp( p, Aig_Regular(pLeftF), Aig_Regular(pRightF), fComplL, fComplR ) != 1 {
                Vec_IntWriteEntry( pBmc.vImps, i, 0 );
                break;
            }
        }
    }
    Fra_ImpCompactArray( pBmc.vImps );
}


/**Function*************************************************************

  Synopsis    [Starts the BMC manager.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Fra_BmcStart(pAig *Aig_Man_t, nPref int32, nDepth int32) *Fra_Bmc_t {
    var p *Fra_Bmc_t;
    // inlining of macro ABC_ALLOC
    p = ((Fra_Bmc_t *) malloc(sizeof(Fra_Bmc_t) * (size_t)(1)));
    memset( p, 0, sizeof(Fra_Bmc_t) );
    p.pAig = pAig;
    p.nPref = nPref;
    p.nDepth = nDepth;
    p.nFramesAll = nPref + nDepth;
    // inlining of macro ABC_ALLOC
    p.pObjToFrames  = ((Aig_Obj_t * *) malloc(sizeof(Aig_Obj_t *) * (size_t)(p.nFramesAll * Aig_ManObjNumMax(pAig))));
    memset( p.pObjToFrames, 0, sizeof(Aig_Obj_t *) * p.nFramesAll * Aig_ManObjNumMax(pAig) );
    return p;
}

/**Function*************************************************************

  Synopsis    [Stops the BMC manager.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Fra_BmcStop(p *Fra_Bmc_t) {
    Aig_ManStop( p.pAigFrames );
    if p.pAigFraig {
        Aig_ManStop( p.pAigFraig ); }
    p.pObjToFrames = nil;
    p.pObjToFraig = nil;
    p = nil;
}

/**Function*************************************************************

  Synopsis    [Constructs initialized timeframes of the AIG.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Fra_BmcFrames(p *Fra_Bmc_t, fKeepPos int32) *Aig_Man_t {
    var pAigFrames *Aig_Man_t;
    var pObj, pObjNew *Aig_Obj_t;
    var pLatches **Aig_Obj_t;
    var i, k, f int32;

    // start the fraig package
    pAigFrames = Aig_ManStart( Aig_ManObjNumMax(p.pAig) * p.nFramesAll );
    pAigFrames.pName = Abc_UtilStrsav( p.pAig.pName );
    pAigFrames.pSpec = Abc_UtilStrsav( p.pAig.pSpec );
    // create PI nodes for the frames
    for  f = 0; f < p.nFramesAll; f++  {
        Bmc_ObjSetFrames( Aig_ManConst1(p.pAig), f, Aig_ManConst1(pAigFrames) );
    }
    for  f = 0; f < p.nFramesAll; f++  {
        // inlining of macro Aig_ManForEachPiSeq
        // inlining of macro Vec_PtrForEachEntryStop
        for  i = 0; (i < Aig_ManCiNum(p.pAig)-Aig_ManRegNum(p.pAig)) && (((pObj) = (Aig_Obj_t *)Vec_PtrEntry(p.pAig.vCis, i)), 1); i++  {
            Bmc_ObjSetFrames( pObj, f, Aig_ObjCreateCi(pAigFrames) );
        }
    }
    // set initial state for the latches
    // inlining of macro Aig_ManForEachLoSeq
    // inlining of macro Vec_PtrForEachEntryStart
    for  i = Aig_ManCiNum(p.pAig)-Aig_ManRegNum(p.pAig); (i < Vec_PtrSize(p.pAig.vCis)) && (((pObj) = (Aig_Obj_t *)Vec_PtrEntry(p.pAig.vCis, i)), 1); i++  {
        Bmc_ObjSetFrames( pObj, 0, Aig_ManConst0(pAigFrames) );
    }

    // add timeframes
    // inlining of macro ABC_ALLOC
    pLatches = ((Aig_Obj_t * *) malloc(sizeof(Aig_Obj_t *) * (size_t)(Aig_ManRegNum(p.pAig))));
    for  f = 0; f < p.nFramesAll; f++  {
        // add internal nodes of this frame
        // inlining of macro Aig_ManForEachNode
        // inlining of macro Vec_PtrForEachEntry
        for  i = 0; (i < Vec_PtrSize(p.pAig.vObjs)) && (((pObj) = (Aig_Obj_t *)Vec_PtrEntry(p.pAig.vObjs, i)), 1); i++  { if (pObj) == nil || !Aig_ObjIsNode(pObj) {} else {
            pObjNew = Aig_And( pAigFrames, Bmc_ObjChild0Frames(pObj,f), Bmc_ObjChild1Frames(pObj,f) );
            Bmc_ObjSetFrames( pObj, f, pObjNew );
        }
        }
        if f == p.nFramesAll - 1 {
            break; }
        // save the latch input values
        k = 0;
        // inlining of macro Aig_ManForEachLiSeq
        // inlining of macro Vec_PtrForEachEntryStart
        for  i = Aig_ManCoNum(p.pAig)-Aig_ManRegNum(p.pAig); (i < Vec_PtrSize(p.pAig.vCos)) && (((pObj) = (Aig_Obj_t *)Vec_PtrEntry(p.pAig.vCos, i)), 1); i++  {
            pLatches[k++] = Bmc_ObjChild0Frames(pObj,f);
        }
        assert(k == Aig_ManRegNum(p.pAig), "k == Aig_ManRegNum(p.pAig)");
        // insert them to the latch output values
        k = 0;
        // inlining of macro Aig_ManForEachLoSeq
        // inlining of macro Vec_PtrForEachEntryStart
        for  i = Aig_ManCiNum(p.pAig)-Aig_ManRegNum(p.pAig); (i < Vec_PtrSize(p.pAig.vCis)) && (((pObj) = (Aig_Obj_t *)Vec_PtrEntry(p.pAig.vCis, i)), 1); i++  {
            Bmc_ObjSetFrames( pObj, f+1, pLatches[k++] );
        }
        assert(k == Aig_ManRegNum(p.pAig), "k == Aig_ManRegNum(p.pAig)");
    }
    pLatches = nil;
    if fKeepPos != 0 {
        for  f = 0; f < p.nFramesAll; f++  {
            // inlining of macro Aig_ManForEachPoSeq
            // inlining of macro Vec_PtrForEachEntryStop
            for  i = 0; (i < Aig_ManCoNum(p.pAig)-Aig_ManRegNum(p.pAig)) && (((pObj) = (Aig_Obj_t *)Vec_PtrEntry(p.pAig.vCos, i)), 1); i++  {
                Aig_ObjCreateCo( pAigFrames, Bmc_ObjChild0Frames(pObj,f) );
            }
        }
        Aig_ManCleanup( pAigFrames );
    
    } else {
        // add POs to all the dangling nodes
        // inlining of macro Aig_ManForEachObj
        // inlining of macro Vec_PtrForEachEntry
        for  i = 0; (i < Vec_PtrSize(pAigFrames.vObjs)) && (((pObjNew) = (Aig_Obj_t *)Vec_PtrEntry(pAigFrames.vObjs, i)), 1); i++  { if (pObjNew) == nil {} else
            if Aig_ObjIsNode(pObjNew) && pObjNew.nRefs == 0 {
                Aig_ObjCreateCo( pAigFrames, pObjNew ); }
        }
    }
    // return the new manager
    return pAigFrames;
}

/**Function*************************************************************

  Synopsis    [Performs BMC for the given AIG.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Fra_BmcPerform(p *Fra_Man_t, nPref int32, nDepth int32) {
    var pObj *Aig_Obj_t;
    var i, nImpsOld int32;
    var clk abctime  = Abc_Clock();
    assert(p.pBmc == nil, "p.pBmc == nil");
    // derive and fraig the frames
    p.pBmc = Fra_BmcStart( p.pManAig, nPref, nDepth );
    p.pBmc.pAigFrames = Fra_BmcFrames( p.pBmc, 0 );
    // if implications are present, configure the AIG manager to check them
    if p.pCla.vImps {
        p.pBmc.pAigFrames.pImpFunc = (void (*) (void*, void*))Fra_BmcFilterImplications;
        p.pBmc.pAigFrames.pImpData = p.pBmc;
        p.pBmc.vImps = p.pCla.vImps;
        nImpsOld = Vec_IntSize(p.pCla.vImps);
    } 
    p.pBmc.pAigFraig = Fra_FraigEquivence( p.pBmc.pAigFrames, 1000000, 0 );
    p.pBmc.pObjToFraig = p.pBmc.pAigFrames.pObjCopies;
    p.pBmc.pAigFrames.pObjCopies = nil;
    // annotate frames nodes with pointers to the manager
    // inlining of macro Aig_ManForEachObj
    // inlining of macro Vec_PtrForEachEntry
    for  i = 0; (i < Vec_PtrSize(p.pBmc.pAigFrames.vObjs)) && (((pObj) = (Aig_Obj_t *)Vec_PtrEntry(p.pBmc.pAigFrames.vObjs, i)), 1); i++  { if (pObj) == nil {} else {
        pObj.pData = p; }
    }
    // report the results
    if p.pPars.fVerbose {
        fmt.Printf("Original AIG = %d. Init %d frames = %d. Fraig = %d.  ", Aig_ManNodeNum(p.pBmc.pAig), p.pBmc.nFramesAll, Aig_ManNodeNum(p.pBmc.pAigFrames), Aig_ManNodeNum(p.pBmc.pAigFraig));
        // inlining of macro ABC_PRT
        Abc_Print(1, "%s =", ("Time"));
Abc_Print(1, "%9.2f sec\n", 1.0*(float64((Abc_Clock() - clk)))/(float64(int64(time.Second))));;
        fmt.Printf("Before BMC: ");  
//        Fra_ClassesPrint( p->pCla, 0 );
        fmt.Printf("Const = %5d. Class = %5d. Lit = %5d. ", Vec_PtrSize(p.pCla.vClasses1), Vec_PtrSize(p.pCla.vClasses), Fra_ClassesCountLits(p.pCla));
        if p.pCla.vImps {
            fmt.Printf("Imp = %5d. ", nImpsOld); }
        fmt.Printf("\n");
    }
    // refine the classes
    p.pCla.pFuncNodeIsConst   = Fra_BmcNodeIsConst;
    p.pCla.pFuncNodesAreEqual = Fra_BmcNodesAreEqual;
    Fra_ClassesRefine( p.pCla );
    Fra_ClassesRefine1( p.pCla, 1, nil );
    p.pCla.pFuncNodeIsConst   = Fra_SmlNodeIsConst;
    p.pCla.pFuncNodesAreEqual = Fra_SmlNodesAreEqual;
    // report the results
    if p.pPars.fVerbose {
        fmt.Printf("After  BMC: ");  
//        Fra_ClassesPrint( p->pCla, 0 );
        fmt.Printf("Const = %5d. Class = %5d. Lit = %5d. ", Vec_PtrSize(p.pCla.vClasses1), Vec_PtrSize(p.pCla.vClasses), Fra_ClassesCountLits(p.pCla));
        if p.pCla.vImps {
            fmt.Printf("Imp = %5d. ", Vec_IntSize(p.pCla.vImps)); }
        fmt.Printf("\n");
    }
    // free the BMC manager
    Fra_BmcStop( p.pBmc );  
    p.pBmc = nil;
}

/**Function*************************************************************

  Synopsis    [Performs BMC for the given AIG.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Fra_BmcPerformSimple(pAig *Aig_Man_t, nFrames int32, nBTLimit int32, fRewrite int32, fVerbose int32) {
    extern Fra_Man_t * Fra_LcrAigPrepare( Aig_Man_t * pAig );
    var pTemp *Fra_Man_t;
    var pBmc *Fra_Bmc_t;
    var pAigTemp *Aig_Man_t;
    var clk abctime;
    var iOutput int32;
    // derive and fraig the frames
    clk = Abc_Clock();
    pBmc = Fra_BmcStart( pAig, 0, nFrames );
    pTemp = Fra_LcrAigPrepare( pAig );
    pTemp.pBmc = pBmc;
    pBmc.pAigFrames = Fra_BmcFrames( pBmc, 1 );
    if fVerbose != 0 {
        fmt.Printf("AIG:  PI/PO/Reg = %d/%d/%d.  Node = %6d. Lev = %5d.\n", Aig_ManCiNum(pAig)-Aig_ManRegNum(pAig), Aig_ManCoNum(pAig)-Aig_ManRegNum(pAig), Aig_ManRegNum(pAig), Aig_ManNodeNum(pAig), Aig_ManLevelNum(pAig));
        fmt.Printf("Time-frames (%d):  PI/PO = %d/%d.  Node = %6d. Lev = %5d.  ", nFrames, Aig_ManCiNum(pBmc.pAigFrames), Aig_ManCoNum(pBmc.pAigFrames), Aig_ManNodeNum(pBmc.pAigFrames), Aig_ManLevelNum(pBmc.pAigFrames));
        // inlining of macro ABC_PRT
        Abc_Print(1, "%s =", ("Time"));
Abc_Print(1, "%9.2f sec\n", 1.0*(float64((Abc_Clock() - clk)))/(float64(int64(time.Second))));;
    }
    if fRewrite != 0 {
        clk = Abc_Clock();
        pBmc.pAigFrames = Dar_ManRwsat( pAigTemp = pBmc.pAigFrames, 1, 0 );
        Aig_ManStop( pAigTemp );
        if fVerbose != 0 {
            fmt.Printf("Time-frames after rewriting:  Node = %6d. Lev = %5d.  ", Aig_ManNodeNum(pBmc.pAigFrames), Aig_ManLevelNum(pBmc.pAigFrames));
            // inlining of macro ABC_PRT
            Abc_Print(1, "%s =", ("Time"));
Abc_Print(1, "%9.2f sec\n", 1.0*(float64((Abc_Clock() - clk)))/(float64(int64(time.Second))));;
        }
    }
    clk = Abc_Clock();
    iOutput = Fra_FraigMiterAssertedOutput( pBmc.pAigFrames );
    if iOutput >= 0 {
        pAig.pSeqModel = Abc_CexMakeTriv( Aig_ManRegNum(pAig), Aig_ManCiNum(pAig)-Aig_ManRegNum(pAig), Aig_ManCoNum(pAig)-Aig_ManRegNum(pAig), iOutput );
    } else {
        pBmc.pAigFraig = Fra_FraigEquivence( pBmc.pAigFrames, nBTLimit, 1 );
        iOutput = Fra_FraigMiterAssertedOutput( pBmc.pAigFraig );
        if pBmc.pAigFraig.pData {
            pAig.pSeqModel = Fra_SmlCopyCounterExample( pAig, pBmc.pAigFrames, (int *)pBmc.pAigFraig.pData );
            pBmc.pAigFraig.pData = nil;
        
        } else if iOutput >= 0 {
            pAig.pSeqModel = Abc_CexMakeTriv( Aig_ManRegNum(pAig), Aig_ManCiNum(pAig)-Aig_ManRegNum(pAig), Aig_ManCoNum(pAig)-Aig_ManRegNum(pAig), iOutput ); }
    }
    if fVerbose != 0 {
        fmt.Printf("Fraiged init frames: Node = %6d. Lev = %5d.  ", pBmc.pAigFraig? Aig_ManNodeNum(pBmc.pAigFraig) : -1, pBmc.pAigFraig? Aig_ManLevelNum(pBmc.pAigFraig) : -1);
        // inlining of macro ABC_PRT
        Abc_Print(1, "%s =", ("Time"));
Abc_Print(1, "%9.2f sec\n", 1.0*(float64((Abc_Clock() - clk)))/(float64(int64(time.Second))));;
    }
    Fra_BmcStop( pBmc );  
    pTemp = nil;
}


////////////////////////////////////////////////////////////////////////
///                       END OF FILE                                ///
////////////////////////////////////////////////////////////////////////


//ABC_NAMESPACE_IMPL_END

