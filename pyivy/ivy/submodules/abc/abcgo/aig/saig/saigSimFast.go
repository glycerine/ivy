package saig

import (
	"fmt"
	"strconv"
)

/**CFile****************************************************************

  FileName    [saigSimFast.c]

  SystemName  [ABC: Logic synthesis and verification system.]

  PackageName [Sequential AIG package.]

  Synopsis    [Fast sequential AIG simulator.]

  Author      [Alan Mishchenko]
  
  Affiliation [UC Berkeley]

  Date        [Ver. 1.0. Started - June 20, 2005.]

  Revision    [$Id: saigSimFast.c,v 1.00 2005/06/20 00:00:00 alanmi Exp $]

***********************************************************************/

//#include "saig.h"

//#include "base/main/main.h"

//ABC_NAMESPACE_IMPL_START


////////////////////////////////////////////////////////////////////////
///                        DECLARATIONS                              ///
////////////////////////////////////////////////////////////////////////

// the AIG manager
   
type Faig_Man_t struct {
    // parameters
    nPis            int32;
    nPos            int32;
    nCis            int32;
    nCos            int32;
    nFfs            int32;
    nNos            int32;
    // offsets
    nPis1            int32;
    nCis1            int32;
    nCisNos1            int32;
    nCisNosPos1            int32;
    nObjs            int32;
    // allocated data
    nWords            int32;
    pObjs [0]int32;
}

func Faig_CheckIdPi(p *Faig_Man_t, i int32) int32 { return i >= 1 && i < p.nPis1;                     }
func Faig_CheckIdLo(p *Faig_Man_t, i int32) int32 { return i >= p.nPis1 && i < p.nCis1;              }
func Faig_CheckIdNo(p *Faig_Man_t, i int32) int32 { return i >= p.nCis1 && i < p.nCisNos1;           }
func Faig_CheckIdPo(p *Faig_Man_t, i int32) int32 { return i >= p.nCisNos1 && i < p.nCisNosPos1;     }
func Faig_CheckIdLi(p *Faig_Man_t, i int32) int32 { return i >= p.nCisNosPos1 && i < p.nObjs;        }
func Faig_CheckIdCo(p *Faig_Man_t, i int32) int32 { return i >= p.nCisNos1 && i < p.nObjs;           }
func Faig_CheckIdObj(p *Faig_Man_t, i int32) int32 { return i >= 0 && i < p.nObjs;                     }

func Faig_ObjIdToNumPi(p *Faig_Man_t, i int32) int32 { assert((Faig_CheckIdPi(p,i) != 0), "Faig_CheckIdPi(p,i)"); return i - 1;              }
func Faig_ObjIdToNumLo(p *Faig_Man_t, i int32) int32 { assert((Faig_CheckIdLo(p,i) != 0), "Faig_CheckIdLo(p,i)"); return i - p.nPis1;       }
func Faig_ObjIdToNumNo(p *Faig_Man_t, i int32) int32 { assert((Faig_CheckIdNo(p,i) != 0), "Faig_CheckIdNo(p,i)"); return i - p.nCis1;       }
func Faig_ObjIdToNumPo(p *Faig_Man_t, i int32) int32 { assert((Faig_CheckIdPo(p,i) != 0), "Faig_CheckIdPo(p,i)"); return i - p.nCisNos1;    }
func Faig_ObjIdToNumLi(p *Faig_Man_t, i int32) int32 { assert((Faig_CheckIdLi(p,i) != 0), "Faig_CheckIdLi(p,i)"); return i - p.nCisNosPos1; }
func Faig_ObjIdToNumCo(p *Faig_Man_t, i int32) int32 { assert((Faig_CheckIdCo(p,i) != 0), "Faig_CheckIdCo(p,i)"); return i - p.nCisNos1;    }

func Faig_ObjLoToLi(p *Faig_Man_t, i int32) int32 { assert((Faig_CheckIdLo(p,i) != 0), "Faig_CheckIdLo(p,i)"); return p.nObjs - (p.nCis1 - i); }
func Faig_ObjLiToLo(p *Faig_Man_t, i int32) int32 { assert((Faig_CheckIdLi(p,i) != 0), "Faig_CheckIdLi(p,i)"); return p.nCis1 - (p.nObjs - i); }

func Faig_NodeChild0(p *Faig_Man_t, n int32) int32 { return p.pObjs[n<<1];             }
func Faig_NodeChild1(p *Faig_Man_t, n int32) int32 { return p.pObjs[(n<<1)+1];         }
func Faig_CoChild0(p *Faig_Man_t, n int32) int32 { return p.pObjs[(p.nNos<<1)+n];   }
func Faig_ObjFaninC(iFan int32) int32 { return iFan & 1;                   }
func Faig_ObjFanin(iFan int32) int32 { return iFan >> 1;                  }

////////////////////////////////////////////////////////////////////////
///                     FUNCTION DEFINITIONS                         ///
////////////////////////////////////////////////////////////////////////

/**Function*************************************************************

  Synopsis    [Checks if the manager is correct.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Faig_ManIsCorrect(pAig *Aig_Man_t) int32 {
    return Aig_ManObjNumMax(pAig) == 
        1 + Aig_ManCiNum(pAig) + Aig_ManNodeNum(pAig) + Aig_ManCoNum(pAig);
}

/**Function*************************************************************

  Synopsis    [Creates fast simulation manager.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Faig_ManAlloc(pAig *Aig_Man_t) *Faig_Man_t {
    var p *Faig_Man_t;
    var nWords int32;
//    assert( Faig_ManIsCorrect(pAig) );
    nWords = 2 * Aig_ManNodeNum(pAig) + Aig_ManCoNum(pAig);
    // inlining of macro ABC_ALLOC
    p = (Faig_Man_t *)((char *) malloc(sizeof(char) * (size_t)(sizeof(Faig_Man_t) + sizeof(int) * nWords)));
//printf( "Allocating %7.2f MB.\n", 1.0 * (sizeof(Faig_Man_t) + sizeof(int) * nWords)/(1<<20) );
    memset( p, 0, sizeof(Faig_Man_t) );
    p.nPis   = Aig_ManCiNum(pAig) - Aig_ManRegNum(pAig);
    p.nPos   = Aig_ManCoNum(pAig) - Aig_ManRegNum(pAig);
    p.nCis   = Aig_ManCiNum(pAig);
    p.nCos   = Aig_ManCoNum(pAig);
    p.nFfs   = Aig_ManRegNum(pAig);
    p.nNos   = Aig_ManNodeNum(pAig);
    // offsets
    p.nPis1  = p.nPis + 1;
    p.nCis1  = p.nCis + 1;
    p.nCisNos1 = p.nCis + p.nNos + 1;
    p.nCisNosPos1 = p.nCis + p.nNos + p.nPos + 1;
    p.nObjs  = p.nCis + p.nNos + p.nCos + 1;
    p.nWords = nWords;
    return p;
}

/**Function*************************************************************

  Synopsis    [Creates fast simulation manager.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Faig_ManCreate(pAig *Aig_Man_t) *Faig_Man_t {
    var p *Faig_Man_t;
    var pObj *Aig_Obj_t;
    var i, iWord int32;
    p = Faig_ManAlloc( pAig );
    // inlining of macro Aig_ManForEachNode
    // inlining of macro Vec_PtrForEachEntry
    for  i = 0; (i < Vec_PtrSize(pAig.vObjs)) && (((pObj) = (Aig_Obj_t *)Vec_PtrEntry(pAig.vObjs, i)), 1); i++  { if (pObj) == nil || !Aig_ObjIsNode(pObj) {} else {
        p.pObjs[iWord++] = (Aig_ObjFaninId0(pObj) << 1) | Aig_ObjFaninC0(pObj);
        p.pObjs[iWord++] = (Aig_ObjFaninId1(pObj) << 1) | Aig_ObjFaninC1(pObj);
    }
    }
    // inlining of macro Aig_ManForEachCo
    // inlining of macro Vec_PtrForEachEntry
    for  i = 0; (i < Vec_PtrSize(pAig.vCos)) && (((pObj) = (Aig_Obj_t *)Vec_PtrEntry(pAig.vCos, i)), 1); i++  {
        p.pObjs[iWord++] = (Aig_ObjFaninId0(pObj) << 1) | Aig_ObjFaninC0(pObj);
    }
    assert(iWord == p.nWords, "iWord == p.nWords");
    return p;
}

/**Function*************************************************************

  Synopsis    [Simulates one node.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Faig_SimulateNode(p *Faig_Man_t, Id int32, pSimInfo *uint32) uint32 {
    var n int32  = Faig_ObjIdToNumNo( p, Id );
    var iFan0 int32  = Faig_NodeChild0( p, n );
    var iFan1 int32  = Faig_NodeChild1( p, n );
    if (Faig_ObjFaninC(iFan0) != 0) && (Faig_ObjFaninC(iFan1) != 0) {
        return ~(pSimInfo[Faig_ObjFanin(iFan0)] | pSimInfo[Faig_ObjFanin(iFan1)]); }
    if (Faig_ObjFaninC(iFan0) != 0) && !(Faig_ObjFaninC(iFan1) != 0) {
        return (~pSimInfo[Faig_ObjFanin(iFan0)] & pSimInfo[Faig_ObjFanin(iFan1)]); }
    if !(Faig_ObjFaninC(iFan0) != 0) && (Faig_ObjFaninC(iFan1) != 0) {
        return (pSimInfo[Faig_ObjFanin(iFan0)] & ~pSimInfo[Faig_ObjFanin(iFan1)]); }
    // if ( !Faig_ObjFaninC(iFan0) && !Faig_ObjFaninC(iFan1) )
    return (pSimInfo[Faig_ObjFanin(iFan0)] & pSimInfo[Faig_ObjFanin(iFan1)]);
} 

/**Function*************************************************************

  Synopsis    [Simulates one node.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Faig_SimulateCo(p *Faig_Man_t, Id int32, pSimInfo *uint32) uint32 {
    var n int32  = Faig_ObjIdToNumCo( p, Id );
    var iFan0 int32  = Faig_CoChild0( p, n );
    if Faig_ObjFaninC(iFan0) != 0 {
        return ~pSimInfo[Faig_ObjFanin(iFan0)]; }
    // if ( !Faig_ObjFaninC(iFan0) )
    return pSimInfo[Faig_ObjFanin(iFan0)];
}

/**Function*************************************************************

  Synopsis    []

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Faig_SimulateRandomShift(uOld uint32) uint32 {
    return (uOld << 16) | ((uOld ^ Aig_ManRandom(0)) & 0xffff);
}

/**Function*************************************************************

  Synopsis    []

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Faig_SimulateTransferShift(uOld uint32, uNew uint32) uint32 {
    return (uOld << 16) | (uNew & 0xffff);
}

/**Function*************************************************************

  Synopsis    [Simulates the timeframes.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Faig_ManSimulateFrames(p *Faig_Man_t, nFrames int32, nPref int32, fTrans int32) *int32 {
    // inlining of macro ABC_CALLOC
    var pNumOnes *int32  = ((int *) calloc((size_t)(p.nObjs), sizeof(int)));
    // inlining of macro ABC_ALLOC
    var pSimInfo *uint32  = ((unsigned *) malloc(sizeof(unsigned) * (size_t)(p.nObjs)));
    var f, i int32;
//printf( "Allocating %7.2f MB.\n", 1.0 * 4 * p->nObjs/(1<<20) );
//printf( "Allocating %7.2f MB.\n", 1.0 * 4 * p->nObjs/(1<<20) );
    // set constant 1
    pSimInfo[0] = ~0;
    for  f = 0; f < nFrames; f++  {
        if fTrans != 0 {
            for  i = 1; i < p.nPis1; i++  {
                pSimInfo[i] = f? Faig_SimulateRandomShift( pSimInfo[i] ) : Aig_ManRandom( 0 );
            }
            for       ; i < p.nCis1; i++  {
                pSimInfo[i] = f? Faig_SimulateTransferShift( pSimInfo[i], pSimInfo[Faig_ObjLoToLi(p,i)] ) : 0;
            }
        
        } else {
            for  i = 1; i < p.nPis1; i++  {
                pSimInfo[i] = Aig_ManRandom( 0 );
            }
            for       ; i < p.nCis1; i++  {
                pSimInfo[i] = f? pSimInfo[Faig_ObjLoToLi(p,i)] : 0;
            }
        }
        for       ; i < p.nCisNos1; i++  {
            pSimInfo[i] = Faig_SimulateNode( p, i, pSimInfo );
        }
        for       ; i < p.nObjs; i++  {
            pSimInfo[i] = Faig_SimulateCo( p, i, pSimInfo );
        }
        if f < nPref {
            continue; }
        if fTrans != 0 {
            for  i = 0; i < p.nObjs; i++  {
                pNumOnes[i] += Aig_WordCountOnes( (pSimInfo[i] ^ (pSimInfo[i] >> 16)) & 0xffff );
            }
        
        } else {
            for  i = 0; i < p.nObjs; i++  {
                pNumOnes[i] += Aig_WordCountOnes( pSimInfo[i] );
            }
        }
    }
    // inlining of macro ABC_FREE
    ((pSimInfo) ? (free((char *) (pSimInfo)), (pSimInfo) = 0) : 0);
    return pNumOnes;
}

/**Function*************************************************************

  Synopsis    [Computes switching activity of one node.]

  Description [Uses the formula: Switching = 2 * nOnes * nZeros / (nTotal ^ 2) ]
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Faig_ManComputeSwitching(nOnes int32, nSimWords int32) float32 {
    var nTotal int32  = 32 * nSimWords;
    return float32(2.0) * nOnes / nTotal * (nTotal - nOnes) / nTotal;
}

/**Function*************************************************************

  Synopsis    [Computes switching activity of one node.]

  Description [Uses the formula: Switching = 2 * nOnes * nZeros / (nTotal ^ 2) ]
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Faig_ManComputeProbOne(nOnes int32, nSimWords int32) float32 {
    var nTotal int32  = 32 * nSimWords;
    return float32(nOnes) / nTotal;
}

/**Function*************************************************************

  Synopsis    []

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Faig_ManComputeSwitchProbs4(p *Aig_Man_t, nFrames int32, nPref int32, fProbOne int32) *Vec_Int_t {
    var fTrans int32  = 1;
    var pAig *Faig_Man_t;
    var vSwitching *Vec_Int_t;
    var pProbs *int32;
    var pSwitching *float32;
    var nFramesReal int32;
    var clk abctime;//, clkTotal = Abc_Clock();
    if fProbOne != 0 {
        fTrans = 0; }
    vSwitching = Vec_IntStart( Aig_ManObjNumMax(p) );
    pSwitching = (float *)vSwitching.pArray;
clk = Abc_Clock();
    pAig = Faig_ManCreate( p );
//ABC_PRT( "\nCreation  ", Abc_Clock() - clk );
    Aig_ManRandom( 1 );
    // get the number of  frames to simulate
    // if the parameter "seqsimframes" is defined, use it
    // otherwise, use the given number of frames "nFrames"
    nFramesReal = nFrames;
    if Abc_FrameReadFlag("seqsimframes") {
        nFramesReal = strconv.Atoi(Abc_FrameReadFlag("seqsimframes")); }
    if nFramesReal <= nPref {
        fmt.Printf("The total number of frames (%d) should exceed prefix (%d).\n", nFramesReal, nPref);
        fmt.Printf("Setting the total number of frames to be %d.\n", nFrames);
        nFramesReal = nFrames;
    }
//printf( "Simulating %d frames.\n", nFramesReal );
clk = Abc_Clock();
    pProbs = Faig_ManSimulateFrames( pAig, nFramesReal, nPref, fTrans );
//ABC_PRT( "Simulation", Abc_Clock() - clk );
clk = Abc_Clock();
    if fTrans != 0 {
        var pObj *Aig_Obj_t;
        var i, Counter int32;
        pObj = Aig_ManConst1(p);
        pSwitching[pObj.Id] = Faig_ManComputeProbOne( pProbs[Counter++], (nFramesReal - nPref)/2 );
        // inlining of macro Aig_ManForEachCi
        // inlining of macro Vec_PtrForEachEntry
        for  i = 0; (i < Vec_PtrSize(p.vCis)) && (((pObj) = (Aig_Obj_t *)Vec_PtrEntry(p.vCis, i)), 1); i++  {
            pSwitching[pObj.Id] = Faig_ManComputeProbOne( pProbs[Counter++], (nFramesReal - nPref)/2 );
        }
        // inlining of macro Aig_ManForEachNode
        // inlining of macro Vec_PtrForEachEntry
        for  i = 0; (i < Vec_PtrSize(p.vObjs)) && (((pObj) = (Aig_Obj_t *)Vec_PtrEntry(p.vObjs, i)), 1); i++  { if (pObj) == nil || !Aig_ObjIsNode(pObj) {} else {
            pSwitching[pObj.Id] = Faig_ManComputeProbOne( pProbs[Counter++], (nFramesReal - nPref)/2 ); }
        }
        // inlining of macro Aig_ManForEachCo
        // inlining of macro Vec_PtrForEachEntry
        for  i = 0; (i < Vec_PtrSize(p.vCos)) && (((pObj) = (Aig_Obj_t *)Vec_PtrEntry(p.vCos, i)), 1); i++  {
            pSwitching[pObj.Id] = Faig_ManComputeProbOne( pProbs[Counter++], (nFramesReal - nPref)/2 );
        }
        assert(Counter == pAig.nObjs, "Counter == pAig.nObjs");
    
    } else if fProbOne != 0 {
        var pObj *Aig_Obj_t;
        var i, Counter int32;
        pObj = Aig_ManConst1(p);
        pSwitching[pObj.Id] = Faig_ManComputeProbOne( pProbs[Counter++], nFramesReal - nPref );
        // inlining of macro Aig_ManForEachCi
        // inlining of macro Vec_PtrForEachEntry
        for  i = 0; (i < Vec_PtrSize(p.vCis)) && (((pObj) = (Aig_Obj_t *)Vec_PtrEntry(p.vCis, i)), 1); i++  {
            pSwitching[pObj.Id] = Faig_ManComputeProbOne( pProbs[Counter++], nFramesReal - nPref );
        }
        // inlining of macro Aig_ManForEachNode
        // inlining of macro Vec_PtrForEachEntry
        for  i = 0; (i < Vec_PtrSize(p.vObjs)) && (((pObj) = (Aig_Obj_t *)Vec_PtrEntry(p.vObjs, i)), 1); i++  { if (pObj) == nil || !Aig_ObjIsNode(pObj) {} else {
            pSwitching[pObj.Id] = Faig_ManComputeProbOne( pProbs[Counter++], nFramesReal - nPref ); }
        }
        // inlining of macro Aig_ManForEachCo
        // inlining of macro Vec_PtrForEachEntry
        for  i = 0; (i < Vec_PtrSize(p.vCos)) && (((pObj) = (Aig_Obj_t *)Vec_PtrEntry(p.vCos, i)), 1); i++  {
            pSwitching[pObj.Id] = Faig_ManComputeProbOne( pProbs[Counter++], nFramesReal - nPref );
        }
        assert(Counter == pAig.nObjs, "Counter == pAig.nObjs");
    
    } else {
        var pObj *Aig_Obj_t;
        var i, Counter int32;
        pObj = Aig_ManConst1(p);
        pSwitching[pObj.Id] = Faig_ManComputeSwitching( pProbs[Counter++], nFramesReal - nPref );
        // inlining of macro Aig_ManForEachCi
        // inlining of macro Vec_PtrForEachEntry
        for  i = 0; (i < Vec_PtrSize(p.vCis)) && (((pObj) = (Aig_Obj_t *)Vec_PtrEntry(p.vCis, i)), 1); i++  {
            pSwitching[pObj.Id] = Faig_ManComputeSwitching( pProbs[Counter++], nFramesReal - nPref );
        }
        // inlining of macro Aig_ManForEachNode
        // inlining of macro Vec_PtrForEachEntry
        for  i = 0; (i < Vec_PtrSize(p.vObjs)) && (((pObj) = (Aig_Obj_t *)Vec_PtrEntry(p.vObjs, i)), 1); i++  { if (pObj) == nil || !Aig_ObjIsNode(pObj) {} else {
            pSwitching[pObj.Id] = Faig_ManComputeSwitching( pProbs[Counter++], nFramesReal - nPref ); }
        }
        // inlining of macro Aig_ManForEachCo
        // inlining of macro Vec_PtrForEachEntry
        for  i = 0; (i < Vec_PtrSize(p.vCos)) && (((pObj) = (Aig_Obj_t *)Vec_PtrEntry(p.vCos, i)), 1); i++  {
            pSwitching[pObj.Id] = Faig_ManComputeSwitching( pProbs[Counter++], nFramesReal - nPref );
        }
        assert(Counter == pAig.nObjs, "Counter == pAig.nObjs");
    }
    // inlining of macro ABC_FREE
    ((pProbs) ? (free((char *) (pProbs)), (pProbs) = 0) : 0);
    // inlining of macro ABC_FREE
    ((pAig) ? (free((char *) (pAig)), (pAig) = 0) : 0);
//ABC_PRT( "Switch    ", Abc_Clock() - clk );
//ABC_PRT( "TOTAL     ", Abc_Clock() - clkTotal );
    return vSwitching;
}

/**Function*************************************************************

  Synopsis    [Computes probability of switching (or of being 1).]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Saig_ManComputeSwitchProb3s(p *Aig_Man_t, nFrames int32, nPref int32, fProbOne int32) *Vec_Int_t {
//    return Faig_ManComputeSwitchProbs( p, nFrames, nPref, fProbOne );
    return nil;
}


////////////////////////////////////////////////////////////////////////
///                       END OF FILE                                ///
////////////////////////////////////////////////////////////////////////


//ABC_NAMESPACE_IMPL_END

