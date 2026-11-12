package saig

import (
	"fmt"
	"strconv"
)

/**CFile****************************************************************

  FileName    [saigSwitch.c]

  SystemName  [ABC: Logic synthesis and verification system.]

  PackageName [Sequential AIG package.]

  Synopsis    [Returns switching propabilities.]

  Author      [Alan Mishchenko]
  
  Affiliation [UC Berkeley]

  Date        [Ver. 1.0. Started - June 20, 2005.]

  Revision    [$Id: saigSwitch.c,v 1.00 2005/06/20 00:00:00 alanmi Exp $]

***********************************************************************/

//#include "saig.h"

//#include "base/main/main.h"

//ABC_NAMESPACE_IMPL_START


////////////////////////////////////////////////////////////////////////
///                        DECLARATIONS                              ///
////////////////////////////////////////////////////////////////////////

   
type Saig_SimObj_t struct {
    iFan0      int32;
    iFan1      int32;
    Type uint32; /* C bitfield: 8 bits; review packing and truncation */
    Number uint32; /* C bitfield: 24 bits; review packing and truncation */
    pData [1]uint32;
}

func Saig_SimObjFaninC0(pObj *Saig_SimObj_t) int32 { return pObj.iFan0 & 1;  }
func Saig_SimObjFaninC1(pObj *Saig_SimObj_t) int32 { return pObj.iFan1 & 1;  }
func Saig_SimObjFanin0(pObj *Saig_SimObj_t) int32 { return pObj.iFan0 >> 1; }
func Saig_SimObjFanin1(pObj *Saig_SimObj_t) int32 { return pObj.iFan1 >> 1; }

//typedef struct Aig_CMan_t_ Aig_CMan_t;

//static Aig_CMan_t * Aig_CManCreate( Aig_Man_t * p );

////////////////////////////////////////////////////////////////////////
///                     FUNCTION DEFINITIONS                         ///
////////////////////////////////////////////////////////////////////////

/**Function*************************************************************

  Synopsis    [Creates fast simulation manager.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Saig_ManCreateMan(p *Aig_Man_t) *Saig_SimObj_t {
    var pAig, pEntry *Saig_SimObj_t;
    var pObj *Aig_Obj_t;
    var i int32;
    // inlining of macro ABC_CALLOC
    pAig = ((Saig_SimObj_t *) calloc((size_t)(Aig_ManObjNumMax(p)+1), sizeof(Saig_SimObj_t)));
//    printf( "Allocating %7.2f MB.\n", 1.0 * sizeof(Saig_SimObj_t) * (Aig_ManObjNumMax(p)+1)/(1<<20) );
    // inlining of macro Aig_ManForEachObj
    // inlining of macro Vec_PtrForEachEntry
    for  i = 0; (i < Vec_PtrSize(p.vObjs)) && (((pObj) = (Aig_Obj_t *)Vec_PtrEntry(p.vObjs, i)), 1); i++  { if (pObj) == nil {} else {
        pEntry = pAig + i;
        pEntry.Type = pObj.Type;
        if Aig_ObjIsCi(pObj) || i == 0 {
            if Saig_ObjIsLo(p, pObj) {
                pEntry.iFan0 = (Saig_ObjLoToLi(p, pObj).Id << 1);
                pEntry.iFan1 = -1;
            }
            continue;
        }
        pEntry.iFan0 = (Aig_ObjFaninId0(pObj) << 1) | Aig_ObjFaninC0(pObj);
        if Aig_ObjIsCo(pObj) {
            continue; }
        assert(Aig_ObjIsNode(pObj), "Aig_ObjIsNode(pObj)");
        pEntry.iFan1 = (Aig_ObjFaninId1(pObj) << 1) | Aig_ObjFaninC1(pObj);
    }
    }
    pEntry = pAig + Aig_ManObjNumMax(p);
    pEntry.Type = AIG_OBJ_VOID;
    return pAig;
}

/**Function*************************************************************

  Synopsis    [Simulated one node.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Saig_ManSimulateNode2(pAig *Saig_SimObj_t, pObj *Saig_SimObj_t) {
    var pObj0 *Saig_SimObj_t  = pAig + Saig_SimObjFanin0( pObj );
    var pObj1 *Saig_SimObj_t  = pAig + Saig_SimObjFanin1( pObj );
    if (Saig_SimObjFaninC0(pObj) != 0) && (Saig_SimObjFaninC1(pObj) != 0) {
        pObj.pData[0] = ~(pObj0.pData[0] | pObj1.pData[0]);
    } else if (Saig_SimObjFaninC0(pObj) != 0) && !(Saig_SimObjFaninC1(pObj) != 0) {
        pObj.pData[0] = (~pObj0.pData[0] & pObj1.pData[0]);
    } else if !(Saig_SimObjFaninC0(pObj) != 0) && (Saig_SimObjFaninC1(pObj) != 0) {
        pObj.pData[0] = (pObj0.pData[0] & ~pObj1.pData[0]);
    } else { // if ( !Saig_SimObjFaninC0(pObj) && !Saig_SimObjFaninC1(pObj) )
        pObj.pData[0] = (pObj0.pData[0] & pObj1.pData[0]); }
}

/**Function*************************************************************

  Synopsis    [Simulated one node.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Saig_ManSimulateNode(pAig *Saig_SimObj_t, pObj *Saig_SimObj_t) {
    var pObj0 *Saig_SimObj_t  = pAig + Saig_SimObjFanin0( pObj );
    var pObj1 *Saig_SimObj_t  = pAig + Saig_SimObjFanin1( pObj );
    pObj.pData[0] = (Saig_SimObjFaninC0(pObj)? ~pObj0.pData[0] : pObj0.pData[0])
        & (Saig_SimObjFaninC1(pObj)? ~pObj1.pData[0] : pObj1.pData[0]);
}

/**Function*************************************************************

  Synopsis    [Simulated buffer/inverter.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Saig_ManSimulateOneInput(pAig *Saig_SimObj_t, pObj *Saig_SimObj_t) {
    var pObj0 *Saig_SimObj_t  = pAig + Saig_SimObjFanin0( pObj );
    if Saig_SimObjFaninC0(pObj) != 0 {
        pObj.pData[0] = ~pObj0.pData[0];
    } else { // if ( !Saig_SimObjFaninC0(pObj) )
        pObj.pData[0] = pObj0.pData[0]; }
}

/**Function*************************************************************

  Synopsis    [Simulates the timeframes.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Saig_ManSimulateFrames(pAig *Saig_SimObj_t, nFrames int32, nPref int32) {
    var pEntry *Saig_SimObj_t;
    var f int32;
    for  f = 0; f < nFrames; f++  {
        for  pEntry = pAig; pEntry.Type != AIG_OBJ_VOID; pEntry++  {
            if pEntry.Type == AIG_OBJ_AND {
                Saig_ManSimulateNode( pAig, pEntry );
            } else if pEntry.Type == AIG_OBJ_CO {
                Saig_ManSimulateOneInput( pAig, pEntry );
            } else if pEntry.Type == AIG_OBJ_CI {
                if pEntry.iFan0 == 0 { // true PI
                    pEntry.pData[0] = Aig_ManRandom( 0 );
                } else if f > 0 { // register output
                    Saig_ManSimulateOneInput( pAig, pEntry ); }
            
            } else if pEntry.Type == AIG_OBJ_CONST1 {
                pEntry.pData[0] = ~0;
            } else if pEntry.Type != AIG_OBJ_NONE {
                assert(false, "0"); }
            if f >= nPref {
                pEntry.Number += Aig_WordCountOnes( pEntry.pData[0] ); }
        }
    }
}

/**Function*************************************************************

  Synopsis    [Computes switching activity of one node.]

  Description [Uses the formula: Switching = 2 * nOnes * nZeros / (nTotal ^ 2) ]
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Saig_ManComputeSwitching(nOnes int32, nSimWords int32) float32 {
    var nTotal int32  = 32 * nSimWords;
    return float32(2.0) * nOnes / nTotal * (nTotal - nOnes) / nTotal;
}

/**Function*************************************************************

  Synopsis    [Computes switching activity of one node.]

  Description [Uses the formula: Switching = 2 * nOnes * nZeros / (nTotal ^ 2) ]
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Saig_ManComputeProbOne(nOnes int32, nSimWords int32) float32 {
    var nTotal int32  = 32 * nSimWords;
    return float32(nOnes) / nTotal;
}

/**Function*************************************************************

  Synopsis    [Computes switching activity of one node.]

  Description [Uses the formula: Switching = 2 * nOnes * nZeros / (nTotal ^ 2) ]
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Saig_ManComputeProbOnePlus(nOnes int32, nSimWords int32, fCompl int32) float32 {
    var nTotal int32  = 32 * nSimWords;
    if fCompl != 0 {
        return float32((nTotal-nOnes)) / nTotal;
    } else {
        return float32(nOnes) / nTotal; }
}

/**Function*************************************************************

  Synopsis    [Compute switching probabilities of all nodes.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Saig_ManComputeSwitchProb4s(p *Aig_Man_t, nFrames int32, nPref int32, fProbOne int32) *Vec_Int_t {
    var pAig, pEntry *Saig_SimObj_t;
    var vSwitching *Vec_Int_t;
    var pSwitching *float32;
    var nFramesReal int32;
    var clk abctime;//, clkTotal = Abc_Clock();
    vSwitching = Vec_IntStart( Aig_ManObjNumMax(p) );
    pSwitching = (float *)vSwitching.pArray;
clk = Abc_Clock();
    pAig = Saig_ManCreateMan( p );
//ABC_PRT( "\nCreation  ", Abc_Clock() - clk );

    Aig_ManRandom( 1 );
    // get the number of  frames to simulate
    // if the parameter "seqsimframes" is defined, use it
    // otherwise, use the given number of frames "nFrames"
    nFramesReal = nFrames;
    if Abc_FrameReadFlag("seqsimframes") {
        nFramesReal = strconv.Atoi(Abc_FrameReadFlag("seqsimframes")); }
    if nFramesReal <= nPref {
        fmt.Printf("The total number of frames (%d) should exceed prefix (%d).\n", nFramesReal, nPref);\
        fmt.Printf("Setting the total number of frames to be %d.\n", nFrames);
        nFramesReal = nFrames;
    }
//printf( "Simulating %d frames.\n", nFramesReal );
clk = Abc_Clock();
    Saig_ManSimulateFrames( pAig, nFramesReal, nPref );
//ABC_PRT( "Simulation", Abc_Clock() - clk );
clk = Abc_Clock();
    for  pEntry = pAig; pEntry.Type != AIG_OBJ_VOID; pEntry++  {
/*
        if ( pEntry->Type == AIG_OBJ_AND )
        {
        Saig_SimObj_t * pObj0 = pAig + Saig_SimObjFanin0( pEntry );
        Saig_SimObj_t * pObj1 = pAig + Saig_SimObjFanin1( pEntry );
        printf( "%5.2f = %5.2f * %5.2f  (%7.4f)\n", 
            Saig_ManComputeProbOnePlus( pEntry->Number, nFrames - nPref, 0 ),
            Saig_ManComputeProbOnePlus( pObj0->Number, nFrames - nPref, Saig_SimObjFaninC0(pEntry) ),
            Saig_ManComputeProbOnePlus( pObj1->Number, nFrames - nPref, Saig_SimObjFaninC1(pEntry) ),
            Saig_ManComputeProbOnePlus( pEntry->Number, nFrames - nPref, 0 ) -
            Saig_ManComputeProbOnePlus( pObj0->Number, nFrames - nPref, Saig_SimObjFaninC0(pEntry) ) *
            Saig_ManComputeProbOnePlus( pObj1->Number, nFrames - nPref, Saig_SimObjFaninC1(pEntry) )
            );
        }
*/
        if fProbOne != 0 {
            pSwitching[pEntry-pAig] = Saig_ManComputeProbOne( pEntry.Number, nFramesReal - nPref );
        } else {
            pSwitching[pEntry-pAig] = Saig_ManComputeSwitching( pEntry.Number, nFramesReal - nPref ); }
//printf( "%3d : %7.2f\n", pEntry-pAig, pSwitching[pEntry-pAig] );
    }
    // inlining of macro ABC_FREE
    ((pAig) ? (free((char *) (pAig)), (pAig) = 0) : 0);
//ABC_PRT( "Switch    ", Abc_Clock() - clk );
//ABC_PRT( "TOTAL     ", Abc_Clock() - clkTotal );

//    Aig_CManCreate( p );
    return vSwitching;
}




   
type Aig_CMan_t struct {
    // parameters
    nIns             int32;
    nNodes             int32;
    nOuts             int32;
    // current state
    iNode             int32;
    iDiff0             int32;
    iDiff1             int32;
    unsigned char * pCur;
    // stored data
    iPrev             int32;
    nBytes             int32;
    unsigned char   Data[0];
}


/**Function*************************************************************

  Synopsis    []

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Aig_CManStart(nIns int32, nNodes int32, nOuts int32) *Aig_CMan_t {
    var p *Aig_CMan_t;
    // inlining of macro ABC_ALLOC
    p = (Aig_CMan_t *)((char *) malloc(sizeof(char) * (size_t)(sizeof(Aig_CMan_t) + 2*(2*nNodes + nOuts))));
    memset( p, 0, sizeof(Aig_CMan_t) );
    // set parameters
    p.nIns = nIns;
    p.nOuts = nOuts;
    p.nNodes = nNodes;
    p.nBytes = 2*(2*nNodes + nOuts);
    // prepare the manager
    p.iNode = 1 + p.nIns;
    p.iPrev = -1;
    p.pCur = p.Data;
    return p;
}

/**Function*************************************************************

  Synopsis    []

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Aig_CManStop(p *Aig_CMan_t) {
    // inlining of macro ABC_FREE
    ((p) ? (free((char *) (p)), (p) = 0) : 0);
}

/**Function*************************************************************

  Synopsis    []

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Aig_CManRestart(p *Aig_CMan_t) {
    assert(p.iNode == 1 + p.nIns + p.nNodes + p.nOuts, "p.iNode == 1 + p.nIns + p.nNodes + p.nOuts");
    p.iNode = 1 + p.nIns;
    p.iPrev = -1;
    p.pCur = p.Data;
}


/**Function*************************************************************

  Synopsis    []

  Description []
  
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Aig_CManStoreNum(p *Aig_CMan_t, x uint32) {
    for  x & ~0x7f  {
        *p.pCur++ = (x & 0x7f) | 0x80;
        x >>= 7;
    }
    *p.pCur++ = x;
    assert(p.pCur - p.Data < p.nBytes - 10, "p.pCur - p.Data < p.nBytes - 10");
}

/**Function*************************************************************

  Synopsis    []

  Description []
  
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Aig_CManRestoreNum(p *Aig_CMan_t) int32 {
    var ch, i, x int32;
    for  i = 0; (ch = *p.pCur++) & 0x80; i++  {
        x |= (ch & 0x7f) << (7 * i);
    }
    return x | (ch << (7 * i));
}

/**Function*************************************************************

  Synopsis    []

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Aig_CManAddNode(p *Aig_CMan_t, iFan0 int32, iFan1 int32) {
    assert(iFan0 < iFan1, "iFan0 < iFan1");
    assert(iFan1 < (p.iNode << 1), "iFan1 < (p.iNode << 1)");
    Aig_CManStoreNum( p, (p.iNode++ << 1) - iFan1 );
    Aig_CManStoreNum( p, iFan1 - iFan0 );
}

/**Function*************************************************************

  Synopsis    []

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Aig_CManAddPo(p *Aig_CMan_t, iFan0 int32) {
    if p.iPrev == -1 {
        Aig_CManStoreNum( p, p.iNode - iFan0 );
    } else if p.iPrev <= iFan0 {
        Aig_CManStoreNum( p, (iFan0 - p.iPrev) << 1 );
    } else { 
        Aig_CManStoreNum( p,((p.iPrev - iFan0) << 1) | 1 ); }
    p.iPrev = iFan0;
    p.iNode++;
}

/**Function*************************************************************

  Synopsis    []

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Aig_CManGetNode(p *Aig_CMan_t, piFan0 *int32, piFan1 *int32) {
    *piFan1 = (p.iNode++ << 1) - Aig_CManRestoreNum( p );
    *piFan0 = *piFan1 - Aig_CManRestoreNum( p );
}

/**Function*************************************************************

  Synopsis    []

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Aig_CManGetPo(p *Aig_CMan_t) int32 {
    var Num int32  = Aig_CManRestoreNum( p );
    if p.iPrev == -1 {
        p.iPrev = p.iNode; }
    p.iNode++;
    if Num & 1 {
        return p.iPrev = p.iPrev + (Num >> 1); }
    return p.iPrev = p.iPrev - (Num >> 1);
}

/**Function*************************************************************

  Synopsis    [Compute switching probabilities of all nodes.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Aig_CManCreate(p *Aig_Man_t) *Aig_CMan_t {
    var pCMan *Aig_CMan_t;
    var pObj *Aig_Obj_t;
    var i int32;
    pCMan = Aig_CManStart( Aig_ManCiNum(p), Aig_ManNodeNum(p), Aig_ManCoNum(p) );
    // inlining of macro Aig_ManForEachNode
    // inlining of macro Vec_PtrForEachEntry
    for  i = 0; (i < Vec_PtrSize(p.vObjs)) && (((pObj) = (Aig_Obj_t *)Vec_PtrEntry(p.vObjs, i)), 1); i++  { if (pObj) == nil || !Aig_ObjIsNode(pObj) {} else {
        Aig_CManAddNode( pCMan, 
            (Aig_ObjFaninId0(pObj) << 1) | Aig_ObjFaninC0(pObj), 
            (Aig_ObjFaninId1(pObj) << 1) | Aig_ObjFaninC1(pObj) ); }
    }
    // inlining of macro Aig_ManForEachCo
    // inlining of macro Vec_PtrForEachEntry
    for  i = 0; (i < Vec_PtrSize(p.vCos)) && (((pObj) = (Aig_Obj_t *)Vec_PtrEntry(p.vCos, i)), 1); i++  {
        Aig_CManAddPo( pCMan, 
            (Aig_ObjFaninId0(pObj) << 1) | Aig_ObjFaninC0(pObj) );
    } 
    fmt.Printf("\nBytes alloc = %5d.  Bytes used = %7d.  Ave per node = %4.2f. \n", pCMan.nBytes, int32((pCMan.pCur - pCMan.Data)), 1.0 * (pCMan.pCur - pCMan.Data) / (pCMan.nNodes + pCMan.nOuts ));
//    Aig_CManStop( pCMan );
    return pCMan;
}

////////////////////////////////////////////////////////////////////////
///                       END OF FILE                                ///
////////////////////////////////////////////////////////////////////////


//ABC_NAMESPACE_IMPL_END

