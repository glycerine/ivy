package fra

import "fmt"

import "os"

/**CFile****************************************************************

  FileName    [fraSim.c]

  SystemName  [ABC: Logic synthesis and verification system.]

  PackageName [New FRAIG package.]

  Synopsis    []

  Author      [Alan Mishchenko]
  
  Affiliation [UC Berkeley]

  Date        [Ver. 1.0. Started - June 30, 2007.]

  Revision    [$Id: fraSim.c,v 1.00 2007/06/30 00:00:00 alanmi Exp $]

***********************************************************************/

//#include "fra.h"
//#include "aig/saig/saig.h"

//ABC_NAMESPACE_IMPL_START


////////////////////////////////////////////////////////////////////////
///                        DECLARATIONS                              ///
////////////////////////////////////////////////////////////////////////

////////////////////////////////////////////////////////////////////////
///                     FUNCTION DEFINITIONS                         ///
////////////////////////////////////////////////////////////////////////

/**Function*************************************************************

  Synopsis    [Computes hash value of the node using its simulation info.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
var spStatic_Fra_SmlNodeHash_s_FPrimes [128]int32 = [128]int32{ 
        1009, 1049, 1093, 1151, 1201, 1249, 1297, 1361, 1427, 1459, 
        1499, 1559, 1607, 1657, 1709, 1759, 1823, 1877, 1933, 1997, 
        2039, 2089, 2141, 2213, 2269, 2311, 2371, 2411, 2467, 2543, 
        2609, 2663, 2699, 2741, 2797, 2851, 2909, 2969, 3037, 3089, 
        3169, 3221, 3299, 3331, 3389, 3461, 3517, 3557, 3613, 3671, 
        3719, 3779, 3847, 3907, 3943, 4013, 4073, 4129, 4201, 4243, 
        4289, 4363, 4441, 4493, 4549, 4621, 4663, 4729, 4793, 4871, 
        4933, 4973, 5021, 5087, 5153, 5227, 5281, 5351, 5417, 5471, 
        5519, 5573, 5651, 5693, 5749, 5821, 5861, 5923, 6011, 6073, 
        6131, 6199, 6257, 6301, 6353, 6397, 6481, 6563, 6619, 6689, 
        6737, 6803, 6863, 6917, 6977, 7027, 7109, 7187, 7237, 7309, 
        7393, 7477, 7523, 7561, 7607, 7681, 7727, 7817, 7877, 7933, 
        8011, 8039, 8059, 8081, 8093, 8111, 8123, 8147,
    };
func Fra_SmlNodeHash(pObj *Aig_Obj_t, nTableSize int32) int32 {
    var p *Fra_Man_t  = (Fra_Man_t *)pObj.pData;
    
    var pSims *uint32;
    var uHash uint32;
    var i int32;
//    assert( p->pSml->nWordsTotal <= 128 );
    uHash = 0;
    pSims = Fra_ObjSim(p.pSml, pObj.Id);
    for  i = p.pSml.nWordsPref; i < p.pSml.nWordsTotal; i++  {
        uHash ^= pSims[i] * spStatic_Fra_SmlNodeHash_s_FPrimes[i & 0x7F];
    }
    return uHash % nTableSize;
}

/**Function*************************************************************

  Synopsis    [Returns 1 if simulation info is composed of all zeros.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Fra_SmlNodeIsConst(pObj *Aig_Obj_t) int32 {
    var p *Fra_Man_t  = (Fra_Man_t *)pObj.pData;
    var pSims *uint32;
    var i int32;
    pSims = Fra_ObjSim(p.pSml, pObj.Id);
    for  i = p.pSml.nWordsPref; i < p.pSml.nWordsTotal; i++  {
        if pSims[i] {
            return 0; }
    }
    return 1;
}

/**Function*************************************************************

  Synopsis    [Returns 1 if simulation infos are equal.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Fra_SmlNodesAreEqual(pObj0 *Aig_Obj_t, pObj1 *Aig_Obj_t) int32 {
    var p *Fra_Man_t  = (Fra_Man_t *)pObj0.pData;
    var pSims0, pSims1 *uint32;
    var i int32;
    pSims0 = Fra_ObjSim(p.pSml, pObj0.Id);
    pSims1 = Fra_ObjSim(p.pSml, pObj1.Id);
    for  i = p.pSml.nWordsPref; i < p.pSml.nWordsTotal; i++  {
        if pSims0[i] != pSims1[i] {
            return 0; }
    }
    return 1;
}

/**Function*************************************************************

  Synopsis    [Counts the number of 1s in the XOR of simulation data.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Fra_SmlNodeNotEquWeight(p *Fra_Sml_t, Left int32, Right int32) int32 {
    var pSimL, pSimR *uint32;
    var k, Counter int32;
    pSimL = Fra_ObjSim( p, Left );
    pSimR = Fra_ObjSim( p, Right );
    for  k = p.nWordsPref; k < p.nWordsTotal; k++  {
        Counter += Aig_WordCountOnes( pSimL[k] ^ pSimR[k] );
    }
    return Counter;
}

/**Function*************************************************************

  Synopsis    [Returns 1 if simulation info is composed of all zeros.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Fra_SmlNodeIsZero(p *Fra_Sml_t, pObj *Aig_Obj_t) int32 {
    var pSims *uint32;
    var i int32;
    pSims = Fra_ObjSim(p, pObj.Id);
    for  i = p.nWordsPref; i < p.nWordsTotal; i++  {
        if pSims[i] {
            return 0; }
    }
    return 1;
}

/**Function*************************************************************

  Synopsis    [Counts the number of one's in the patten of the output.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Fra_SmlNodeCountOnes(p *Fra_Sml_t, pObj *Aig_Obj_t) int32 {
    var pSims *uint32;
    var i, Counter int32;
    pSims = Fra_ObjSim(p, pObj.Id);
    for  i = 0; i < p.nWordsTotal; i++  {
        Counter += Aig_WordCountOnes( pSims[i] );
    }
    return Counter;
}



/**Function*************************************************************

  Synopsis    [Generated const 0 pattern.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Fra_SmlSavePattern0(p *Fra_Man_t, fInit int32) {
    memset( p.pPatWords, 0, sizeof(unsigned) * p.nPatWords );
}

/**Function*************************************************************

  Synopsis    [[Generated const 1 pattern.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Fra_SmlSavePattern1(p *Fra_Man_t, fInit int32) {
    var pObj *Aig_Obj_t;
    var i, k, nTruePis int32;
    memset( p.pPatWords, 0xff, sizeof(unsigned) * p.nPatWords );
    if !(fInit != 0) {
        return; }
    // clear the state bits to correspond to all-0 initial state
    nTruePis = Aig_ManCiNum(p.pManAig) - Aig_ManRegNum(p.pManAig);
    k = 0;
    // inlining of macro Aig_ManForEachLoSeq
    // inlining of macro Vec_PtrForEachEntryStart
    for  i = Aig_ManCiNum(p.pManAig)-Aig_ManRegNum(p.pManAig); (i < Vec_PtrSize(p.pManAig.vCis)) && (((pObj) = (Aig_Obj_t *)Vec_PtrEntry(p.pManAig.vCis, i)), 1); i++  {
        Abc_InfoXorBit( p.pPatWords, nTruePis * p.nFramesAll + k++ );
    }
}

/**Function*************************************************************

  Synopsis    [Copy pattern from the solver into the internal storage.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Fra_SmlSavePattern(p *Fra_Man_t) {
    var pObj *Aig_Obj_t;
    var i int32;
    memset( p.pPatWords, 0, sizeof(unsigned) * p.nPatWords );
    // inlining of macro Aig_ManForEachCi
    // inlining of macro Vec_PtrForEachEntry
    for  i = 0; (i < Vec_PtrSize(p.pManFraig.vCis)) && (((pObj) = (Aig_Obj_t *)Vec_PtrEntry(p.pManFraig.vCis, i)), 1); i++  {
//        if ( p->pSat->model.ptr[Fra_ObjSatNum(pObj)] == l_True )
        if sat_solver_var_value(p.pSat, Fra_ObjSatNum(pObj)) {
            Abc_InfoSetBit( p.pPatWords, i ); }
    }

    if p.vCex {
        Vec_IntClear( p.vCex );
        for  i = 0; i < Aig_ManCiNum(p.pManAig) - Aig_ManRegNum(p.pManAig); i++  {
            Vec_IntPush( p.vCex, Abc_InfoHasBit( p.pPatWords, i ) );
        }
        for  i = Aig_ManCiNum(p.pManFraig) - Aig_ManRegNum(p.pManFraig); i < Aig_ManCiNum(p.pManFraig); i++  {
            Vec_IntPush( p.vCex, Abc_InfoHasBit( p.pPatWords, i ) );
        }
    }

/*
    printf( "Pattern: " );
    Aig_ManForEachCi( p->pManFraig, pObj, i )
        printf( "%d", Abc_InfoHasBit( p->pPatWords, i ) );
    printf( "\n" );
*/
}



/**Function*************************************************************

  Synopsis    [Creates the counter-example from the successful pattern.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Fra_SmlCheckOutputSavePattern(p *Fra_Man_t, pObjPo *Aig_Obj_t) { 
    var pFanin, pObjPi *Aig_Obj_t;
    var pSims *uint32;
    var i int32;
    var k int32;
    var BestPat int32;
    var pModel *int32;
    // find the word of the pattern
    pFanin = Aig_ObjFanin0(pObjPo);
    pSims = Fra_ObjSim(p.pSml, pFanin.Id);
    for  i = 0; i < p.pSml.nWordsTotal; i++  {
        if pSims[i] {
            break; }
    }
    assert(i < p.pSml.nWordsTotal, "i < p.pSml.nWordsTotal");
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
    pModel = ((int *) malloc(sizeof(int) * (size_t)(Aig_ManCiNum(p.pManFraig)+1)));
    // inlining of macro Aig_ManForEachCi
    // inlining of macro Vec_PtrForEachEntry
    for  i = 0; (i < Vec_PtrSize(p.pManAig.vCis)) && (((pObjPi) = (Aig_Obj_t *)Vec_PtrEntry(p.pManAig.vCis, i)), 1); i++  {
        pModel[i] = Abc_InfoHasBit(Fra_ObjSim(p.pSml, pObjPi.Id), BestPat);
//        printf( "%d", pModel[i] );
    }
    pModel[Aig_ManCiNum(p.pManAig)] = pObjPo.Id;
//    printf( "\n" );
    // set the model
    assert(p.pManFraig.pData == nil, "p.pManFraig.pData == nil");
    p.pManFraig.pData = pModel;
    return;
}

/**Function*************************************************************

  Synopsis    [Returns 1 if the one of the output is already non-constant 0.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Fra_SmlCheckOutput(p *Fra_Man_t) int32 {
    var pObj *Aig_Obj_t;
    var i int32;
    // make sure the reference simulation pattern does not detect the bug
    pObj = Aig_ManCo( p.pManAig, 0 );
    assert(Aig_ObjFanin0(pObj).fPhase == uint32(Aig_ObjFaninC0(pObj)), "Aig_ObjFanin0(pObj).fPhase == uint32(Aig_ObjFaninC0(pObj))"); 
    // inlining of macro Aig_ManForEachCo
    // inlining of macro Vec_PtrForEachEntry
    for  i = 0; (i < Vec_PtrSize(p.pManAig.vCos)) && (((pObj) = (Aig_Obj_t *)Vec_PtrEntry(p.pManAig.vCos, i)), 1); i++  {
        if !(Fra_SmlNodeIsConst( Aig_ObjFanin0(pObj) ) != 0) {
            // create the counter-example from this pattern
            Fra_SmlCheckOutputSavePattern( p, pObj );
            return 1;
        }
    }
    return 0;
}



/**Function*************************************************************

  Synopsis    [Assigns random patterns to the PI node.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Fra_SmlAssignRandom(p *Fra_Sml_t, pObj *Aig_Obj_t) {
    var pSims *uint32;
    var i int32;
    assert(Aig_ObjIsCi(pObj), "Aig_ObjIsCi(pObj)");
    pSims = Fra_ObjSim( p, pObj.Id );
    for  i = 0; i < p.nWordsTotal; i++  {
        pSims[i] = Fra_ObjRandomSim();
    }
}

/**Function*************************************************************

  Synopsis    [Assigns constant patterns to the PI node.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Fra_SmlAssignConst(p *Fra_Sml_t, pObj *Aig_Obj_t, fConst1 int32, iFrame int32) {
    var pSims *uint32;
    var i int32;
    assert(Aig_ObjIsCi(pObj) || Aig_ObjIsConst1(pObj), "Aig_ObjIsCi(pObj) || Aig_ObjIsConst1(pObj)");
    pSims = Fra_ObjSim( p, pObj.Id ) + p.nWordsFrame * iFrame;
    for  i = 0; i < p.nWordsFrame; i++  {
        pSims[i] = fConst1? ~uint32(0) : 0;
    }
}

/**Function*************************************************************

  Synopsis    [Assings random simulation info for the PIs.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Fra_SmlInitialize(p *Fra_Sml_t, fInit int32) {
    var pObj *Aig_Obj_t;
    var i int32;
    if fInit != 0 {
        assert(Aig_ManRegNum(p.pAig) > 0, "Aig_ManRegNum(p.pAig) > 0");
        assert(Aig_ManRegNum(p.pAig) < Aig_ManCiNum(p.pAig), "Aig_ManRegNum(p.pAig) < Aig_ManCiNum(p.pAig)");
        // assign random info for primary inputs
        // inlining of macro Aig_ManForEachPiSeq
        // inlining of macro Vec_PtrForEachEntryStop
        for  i = 0; (i < Aig_ManCiNum(p.pAig)-Aig_ManRegNum(p.pAig)) && (((pObj) = (Aig_Obj_t *)Vec_PtrEntry(p.pAig.vCis, i)), 1); i++  {
            Fra_SmlAssignRandom( p, pObj );
        }
        // assign the initial state for the latches
        // inlining of macro Aig_ManForEachLoSeq
        // inlining of macro Vec_PtrForEachEntryStart
        for  i = Aig_ManCiNum(p.pAig)-Aig_ManRegNum(p.pAig); (i < Vec_PtrSize(p.pAig.vCis)) && (((pObj) = (Aig_Obj_t *)Vec_PtrEntry(p.pAig.vCis, i)), 1); i++  {
            Fra_SmlAssignConst( p, pObj, 0, 0 );
        }
    
    } else {
        // inlining of macro Aig_ManForEachCi
        // inlining of macro Vec_PtrForEachEntry
        for  i = 0; (i < Vec_PtrSize(p.pAig.vCis)) && (((pObj) = (Aig_Obj_t *)Vec_PtrEntry(p.pAig.vCis, i)), 1); i++  {
            Fra_SmlAssignRandom( p, pObj );
        }
    }
}

/**Function*************************************************************

  Synopsis    [Assings distance-1 simulation info for the PIs.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Fra_SmlAssignDist1(p *Fra_Sml_t, pPat *uint32) {
    var pObj *Aig_Obj_t;
    var f, i, k, Limit, nTruePis int32;
    assert(p.nFrames > 0, "p.nFrames > 0");
    if p.nFrames == 1 {
        // copy the PI info 
        // inlining of macro Aig_ManForEachCi
        // inlining of macro Vec_PtrForEachEntry
        for  i = 0; (i < Vec_PtrSize(p.pAig.vCis)) && (((pObj) = (Aig_Obj_t *)Vec_PtrEntry(p.pAig.vCis, i)), 1); i++  {
            Fra_SmlAssignConst( p, pObj, Abc_InfoHasBit(pPat, i), 0 );
        }
        // flip one bit
        Limit = Abc_MinInt( Aig_ManCiNum(p.pAig), p.nWordsTotal * 32 - 1 );
        for  i = 0; i < Limit; i++  {
            Abc_InfoXorBit( Fra_ObjSim( p, Aig_ManCi(p.pAig,i).Id ), i+1 );
        }
    
    } else {
        var fUseDist1 int32;

        // copy the PI info for each frame
        nTruePis = Aig_ManCiNum(p.pAig) - Aig_ManRegNum(p.pAig);
        for  f = 0; f < p.nFrames; f++  {
            // inlining of macro Aig_ManForEachPiSeq
            // inlining of macro Vec_PtrForEachEntryStop
            for  i = 0; (i < Aig_ManCiNum(p.pAig)-Aig_ManRegNum(p.pAig)) && (((pObj) = (Aig_Obj_t *)Vec_PtrEntry(p.pAig.vCis, i)), 1); i++  {
                Fra_SmlAssignConst( p, pObj, Abc_InfoHasBit(pPat, nTruePis * f + i), f );
            }
        }
        // copy the latch info 
        k = 0;
        // inlining of macro Aig_ManForEachLoSeq
        // inlining of macro Vec_PtrForEachEntryStart
        for  i = Aig_ManCiNum(p.pAig)-Aig_ManRegNum(p.pAig); (i < Vec_PtrSize(p.pAig.vCis)) && (((pObj) = (Aig_Obj_t *)Vec_PtrEntry(p.pAig.vCis, i)), 1); i++  {
            Fra_SmlAssignConst( p, pObj, Abc_InfoHasBit(pPat, nTruePis * p.nFrames + k++), 0 );
        }
//        assert( p->pManFraig == NULL || nTruePis * p->nFrames + k == Aig_ManCiNum(p->pManFraig) );

        // flip one bit of the last frame
        if fUseDist1 != 0 { //&& p->nFrames == 2 )
        
            Limit = Abc_MinInt( nTruePis, p.nWordsFrame * 32 - 1 );
            for  i = 0; i < Limit; i++  {
                Abc_InfoXorBit( Fra_ObjSim( p, Aig_ManCi(p.pAig, i).Id ) + p.nWordsFrame*(p.nFrames-1), i+1 );
            }
        }
    }
}


/**Function*************************************************************

  Synopsis    [Simulates one node.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Fra_SmlNodeSimulate(p *Fra_Sml_t, pObj *Aig_Obj_t, iFrame int32) {
    var pSims, pSims0, pSims1 *uint32;
    var fCompl, fCompl0, fCompl1, i int32;
    assert(!Aig_IsComplement(pObj), "!Aig_IsComplement(pObj)");
    assert(Aig_ObjIsNode(pObj), "Aig_ObjIsNode(pObj)");
    assert(iFrame == 0 || p.nWordsFrame < p.nWordsTotal, "iFrame == 0 || p.nWordsFrame < p.nWordsTotal");
    // get hold of the simulation information
    pSims  = Fra_ObjSim(p, pObj.Id) + p.nWordsFrame * iFrame;
    pSims0 = Fra_ObjSim(p, Aig_ObjFanin0(pObj).Id) + p.nWordsFrame * iFrame;
    pSims1 = Fra_ObjSim(p, Aig_ObjFanin1(pObj).Id) + p.nWordsFrame * iFrame;
    // get complemented attributes of the children using their random info
    fCompl  = pObj.fPhase;
    fCompl0 = Aig_ObjPhaseReal(Aig_ObjChild0(pObj));
    fCompl1 = Aig_ObjPhaseReal(Aig_ObjChild1(pObj));
    // simulate
    if fCompl0 && fCompl1 {
        if fCompl {
            for  i = 0; i < p.nWordsFrame; i++  {
                pSims[i] = (pSims0[i] | pSims1[i]);
            }
        } else {
            for  i = 0; i < p.nWordsFrame; i++  {
                pSims[i] = ~(pSims0[i] | pSims1[i]);
            } }
    
    } else if fCompl0 && !fCompl1 {
        if fCompl {
            for  i = 0; i < p.nWordsFrame; i++  {
                pSims[i] = (pSims0[i] | ~pSims1[i]);
            }
        } else {
            for  i = 0; i < p.nWordsFrame; i++  {
                pSims[i] = (~pSims0[i] & pSims1[i]);
            } }
    
    } else if !fCompl0 && fCompl1 {
        if fCompl {
            for  i = 0; i < p.nWordsFrame; i++  {
                pSims[i] = (~pSims0[i] | pSims1[i]);
            }
        } else {
            for  i = 0; i < p.nWordsFrame; i++  {
                pSims[i] = (pSims0[i] & ~pSims1[i]);
            } }
    
    } else { // if ( !fCompl0 && !fCompl1 )
    
        if fCompl {
            for  i = 0; i < p.nWordsFrame; i++  {
                pSims[i] = ~(pSims0[i] & pSims1[i]);
            }
        } else {
            for  i = 0; i < p.nWordsFrame; i++  {
                pSims[i] = (pSims0[i] & pSims1[i]);
            } }
    }
}

/**Function*************************************************************

  Synopsis    [Simulates one node.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Fra_SmlNodesCompareInFrame(p *Fra_Sml_t, pObj0 *Aig_Obj_t, pObj1 *Aig_Obj_t, iFrame0 int32, iFrame1 int32) int32 {
    var pSims0, pSims1 *uint32;
    var i int32;
    assert(!Aig_IsComplement(pObj0), "!Aig_IsComplement(pObj0)");
    assert(!Aig_IsComplement(pObj1), "!Aig_IsComplement(pObj1)");
    assert(iFrame0 == 0 || p.nWordsFrame < p.nWordsTotal, "iFrame0 == 0 || p.nWordsFrame < p.nWordsTotal");
    assert(iFrame1 == 0 || p.nWordsFrame < p.nWordsTotal, "iFrame1 == 0 || p.nWordsFrame < p.nWordsTotal");
    // get hold of the simulation information
    pSims0  = Fra_ObjSim(p, pObj0.Id) + p.nWordsFrame * iFrame0;
    pSims1  = Fra_ObjSim(p, pObj1.Id) + p.nWordsFrame * iFrame1;
    // compare
    for  i = 0; i < p.nWordsFrame; i++  {
        if pSims0[i] != pSims1[i] {
            return 0; }
    }
    return 1;
}

/**Function*************************************************************

  Synopsis    [Simulates one node.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Fra_SmlNodeCopyFanin(p *Fra_Sml_t, pObj *Aig_Obj_t, iFrame int32) {
    var pSims, pSims0 *uint32;
    var fCompl, fCompl0, i int32;
    assert(!Aig_IsComplement(pObj), "!Aig_IsComplement(pObj)");
    assert(Aig_ObjIsCo(pObj), "Aig_ObjIsCo(pObj)");
    assert(iFrame == 0 || p.nWordsFrame < p.nWordsTotal, "iFrame == 0 || p.nWordsFrame < p.nWordsTotal");
    // get hold of the simulation information
    pSims  = Fra_ObjSim(p, pObj.Id) + p.nWordsFrame * iFrame;
    pSims0 = Fra_ObjSim(p, Aig_ObjFanin0(pObj).Id) + p.nWordsFrame * iFrame;
    // get complemented attributes of the children using their random info
    fCompl  = pObj.fPhase;
    fCompl0 = Aig_ObjPhaseReal(Aig_ObjChild0(pObj));
    // copy information as it is
//    if ( Aig_ObjFaninC0(pObj) )
    if fCompl0 {
        for  i = 0; i < p.nWordsFrame; i++  {
            pSims[i] = ~pSims0[i];
        }
    } else {
        for  i = 0; i < p.nWordsFrame; i++  {
            pSims[i] = pSims0[i];
        } }
}

/**Function*************************************************************

  Synopsis    [Simulates one node.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Fra_SmlNodeTransferNext(p *Fra_Sml_t, pOut *Aig_Obj_t, pIn *Aig_Obj_t, iFrame int32) {
    var pSims0, pSims1 *uint32;
    var i int32;
    assert(!Aig_IsComplement(pOut), "!Aig_IsComplement(pOut)");
    assert(!Aig_IsComplement(pIn), "!Aig_IsComplement(pIn)");
    assert(Aig_ObjIsCo(pOut), "Aig_ObjIsCo(pOut)");
    assert(Aig_ObjIsCi(pIn), "Aig_ObjIsCi(pIn)");
    assert(iFrame == 0 || p.nWordsFrame < p.nWordsTotal, "iFrame == 0 || p.nWordsFrame < p.nWordsTotal");
    // get hold of the simulation information
    pSims0 = Fra_ObjSim(p, pOut.Id) + p.nWordsFrame * iFrame;
    pSims1 = Fra_ObjSim(p, pIn.Id) + p.nWordsFrame * (iFrame+1);
    // copy information as it is
    for  i = 0; i < p.nWordsFrame; i++  {
        pSims1[i] = pSims0[i];
    }
}


/**Function*************************************************************

  Synopsis    [Check if any of the POs becomes non-constant.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Fra_SmlCheckNonConstOutputs(p *Fra_Sml_t) int32 {
    var pObj *Aig_Obj_t;
    var i int32;
    // inlining of macro Aig_ManForEachPoSeq
    // inlining of macro Vec_PtrForEachEntryStop
    for  i = 0; (i < Aig_ManCoNum(p.pAig)-Aig_ManRegNum(p.pAig)) && (((pObj) = (Aig_Obj_t *)Vec_PtrEntry(p.pAig.vCos, i)), 1); i++  {
        if !(Fra_SmlNodeIsZero(p, pObj) != 0) {
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
func Fra_SmlSimulateOne(p *Fra_Sml_t) {
    var pObj, pObjLi, pObjLo *Aig_Obj_t;
    var f, i int32;
    var clk abctime;
clk = Abc_Clock();
    for  f = 0; f < p.nFrames; f++  {
        // simulate the nodes
        // inlining of macro Aig_ManForEachNode
        // inlining of macro Vec_PtrForEachEntry
        for  i = 0; (i < Vec_PtrSize(p.pAig.vObjs)) && (((pObj) = (Aig_Obj_t *)Vec_PtrEntry(p.pAig.vObjs, i)), 1); i++  { if (pObj) == nil || !Aig_ObjIsNode(pObj) {} else {
            Fra_SmlNodeSimulate( p, pObj, f ); }
        }
        // copy simulation info into outputs
        // inlining of macro Aig_ManForEachPoSeq
        // inlining of macro Vec_PtrForEachEntryStop
        for  i = 0; (i < Aig_ManCoNum(p.pAig)-Aig_ManRegNum(p.pAig)) && (((pObj) = (Aig_Obj_t *)Vec_PtrEntry(p.pAig.vCos, i)), 1); i++  {
            Fra_SmlNodeCopyFanin( p, pObj, f );
        }
        // quit if this is the last timeframe
        if f == p.nFrames - 1 {
            break; }
        // copy simulation info into outputs
        // inlining of macro Aig_ManForEachLiSeq
        // inlining of macro Vec_PtrForEachEntryStart
        for  i = Aig_ManCoNum(p.pAig)-Aig_ManRegNum(p.pAig); (i < Vec_PtrSize(p.pAig.vCos)) && (((pObj) = (Aig_Obj_t *)Vec_PtrEntry(p.pAig.vCos, i)), 1); i++  {
            Fra_SmlNodeCopyFanin( p, pObj, f );
        }
        // copy simulation info into the inputs
        // inlining of macro Aig_ManForEachLiLoSeq
        for  i = 0; (i < Aig_ManRegNum(p.pAig)) && (((pObjLi) = Aig_ManLi(p.pAig, i)), 1)            && (((pObjLo)=Aig_ManLo(p.pAig, i)), 1); i++  {
            Fra_SmlNodeTransferNext( p, pObjLi, pObjLo, f );
        }
    }
p.timeSim += Abc_Clock() - clk;
p.nSimRounds++;
}


/**Function*************************************************************

  Synopsis    [Resimulates fraiging manager after finding a counter-example.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Fra_SmlResimulate(p *Fra_Man_t) {
    var nChanges int32;
    var clk abctime;
    Fra_SmlAssignDist1( p.pSml, p.pPatWords );
    Fra_SmlSimulateOne( p.pSml );
//    if ( p->pPars->fPatScores )
//        Fra_CleanPatScores( p );
    if p.pPars.fProve && (Fra_SmlCheckOutput(p) != 0) {
        return; }
clk = Abc_Clock();
    nChanges = Fra_ClassesRefine( p.pCla );
    nChanges += Fra_ClassesRefine1( p.pCla, 1, nil );
    if p.pCla.vImps {
        nChanges += Fra_ImpRefineUsingCex( p, p.pCla.vImps ); }
    if p.vOneHots {
        nChanges += Fra_OneHotRefineUsingCex( p, p.vOneHots ); }
p.timeRef += Abc_Clock() - clk;
    if !p.pPars.nFramesK && nChanges < 1 {
        fmt.Printf("Error: A counter-example did not refine classes!\n"); }
//    assert( nChanges >= 1 );
//printf( "Refined classes = %5d.   Changes = %4d.\n", Vec_PtrSize(p->vClasses), nChanges );
}

/**Function*************************************************************

  Synopsis    [Performs simulation of the manager.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Fra_SmlSimulate(p *Fra_Man_t, fInit int32) {
    var fVerbose int32;
    var nChanges, nClasses int32;
    var clk abctime;
    assert(!(fInit != 0) || Aig_ManRegNum(p.pManAig), "!fInit || Aig_ManRegNum(p.pManAig)");
    // start the classes
    Fra_SmlInitialize( p.pSml, fInit );
    Fra_SmlSimulateOne( p.pSml );
    if p.pPars.fProve && (Fra_SmlCheckOutput(p) != 0) {
        return; }
    Fra_ClassesPrepare( p.pCla, p.pPars.fLatchCorr, 0 );
//    Fra_ClassesPrint( p->pCla, 0 );
if fVerbose != 0 {
fmt.Printf("Starting classes = %5d.   Lits = %6d.\n", Vec_PtrSize(p.pCla.vClasses), Fra_ClassesCountLits(p.pCla)); }

//return;

    // refine classes by walking 0/1 patterns
    Fra_SmlSavePattern0( p, fInit );
    Fra_SmlAssignDist1( p.pSml, p.pPatWords );
    Fra_SmlSimulateOne( p.pSml );
    if p.pPars.fProve && (Fra_SmlCheckOutput(p) != 0) {
        return; }
clk = Abc_Clock();
    nChanges = Fra_ClassesRefine( p.pCla );
    nChanges += Fra_ClassesRefine1( p.pCla, 1, nil );
p.timeRef += Abc_Clock() - clk;
if fVerbose != 0 {
fmt.Printf("Refined classes  = %5d.   Changes = %4d.   Lits = %6d.\n", Vec_PtrSize(p.pCla.vClasses), nChanges, Fra_ClassesCountLits(p.pCla)); }
    Fra_SmlSavePattern1( p, fInit );
    Fra_SmlAssignDist1( p.pSml, p.pPatWords );
    Fra_SmlSimulateOne( p.pSml );
    if p.pPars.fProve && (Fra_SmlCheckOutput(p) != 0) {
        return; }
clk = Abc_Clock();
    nChanges = Fra_ClassesRefine( p.pCla );
    nChanges += Fra_ClassesRefine1( p.pCla, 1, nil );
p.timeRef += Abc_Clock() - clk;

if fVerbose != 0 {
fmt.Printf("Refined classes  = %5d.   Changes = %4d.   Lits = %6d.\n", Vec_PtrSize(p.pCla.vClasses), nChanges, Fra_ClassesCountLits(p.pCla)); }
    // refine classes by random simulation
    for {
        Fra_SmlInitialize( p.pSml, fInit );
        Fra_SmlSimulateOne( p.pSml );
        nClasses = Vec_PtrSize(p.pCla.vClasses);
        if p.pPars.fProve && (Fra_SmlCheckOutput(p) != 0) {
            return; }
clk = Abc_Clock();
        nChanges = Fra_ClassesRefine( p.pCla );
        nChanges += Fra_ClassesRefine1( p.pCla, 1, nil );
p.timeRef += Abc_Clock() - clk;
if fVerbose != 0 {
fmt.Printf("Refined classes  = %5d.   Changes = %4d.   Lits = %6d.\n", Vec_PtrSize(p.pCla.vClasses), nChanges, Fra_ClassesCountLits(p.pCla)); }
    
if !( float64(nChanges) / nClasses > p.pPars.dSimSatur ) { break }
} 

//    if ( p->pPars->fVerbose )
//    printf( "Consts = %6d. Classes = %6d. Literals = %6d.\n", 
//        Vec_PtrSize(p->pCla->vClasses1), Vec_PtrSize(p->pCla->vClasses), Fra_ClassesCountLits(p->pCla) );
//    Fra_ClassesPrint( p->pCla, 0 );
}
 

/**Function*************************************************************

  Synopsis    [Allocates simulation manager.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Fra_SmlStart(pAig *Aig_Man_t, nPref int32, nFrames int32, nWordsFrame int32) *Fra_Sml_t {
    var p *Fra_Sml_t;
    // inlining of macro ABC_ALLOC
    p = (Fra_Sml_t *)((char *) malloc(sizeof(char) * (size_t)(sizeof(Fra_Sml_t) + sizeof(unsigned) * Aig_ManObjNumMax(pAig) * (nPref + nFrames) * nWordsFrame)));
    memset( p, 0, sizeof(Fra_Sml_t) + sizeof(unsigned) * (nPref + nFrames) * nWordsFrame );
    p.pAig        = pAig;
    p.nPref       = nPref;
    p.nFrames     = nPref + nFrames;
    p.nWordsFrame = nWordsFrame;
    p.nWordsTotal = (nPref + nFrames) * nWordsFrame;
    p.nWordsPref  = nPref * nWordsFrame;
    // constant 1 is initialized to 0 because we store values modulus phase (pObj->fPhase)
    return p;
}

/**Function*************************************************************

  Synopsis    [Deallocates simulation manager.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Fra_SmlStop(p *Fra_Sml_t) {
    p = nil;
}


/**Function*************************************************************

  Synopsis    [Performs simulation of the uninitialized circuit.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Fra_SmlSimulateComb(pAig *Aig_Man_t, nWords int32, fCheckMiter int32) *Fra_Sml_t {
    var p *Fra_Sml_t;
    p = Fra_SmlStart( pAig, 0, 1, nWords );
    Fra_SmlInitialize( p, 0 );
    Fra_SmlSimulateOne( p );
    if fCheckMiter != 0 {
        p.fNonConstOut = Fra_SmlCheckNonConstOutputs( p ); }
    return p;
}

/**Function*************************************************************

  Synopsis    [Reads simulation patterns from file.]

  Description [Each pattern contains the given number (nInputs) of binary digits.
  No other symbols (except spaces and line endings) are allowed in the file.]
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Fra_SmlSimulateReadFile(pFileName string) *Vec_Str_t {
    var vRes *Vec_Str_t;
    var pFile *os.File;
    var c int32;
    pFile = os.Open(pFileName);
    if pFile == nil {
        fmt.Printf("Cannot open file \"%s\" with simulation patterns.\n", pFileName);
        return nil;
    }
    vRes = Vec_StrAlloc( 1000 );
    for  (c = fgetc(pFile)) != -1  {
        if c == '0' || c == '1' {
            Vec_StrPush( vRes, byte((c - '0')) );
        } else if c != ' ' && c != '\r' && c != '\n' && c != '\t' {
            fmt.Printf("File \"%s\" contains symbol (%c) other than \'0\' or \'1\'.\n", pFileName, byte(c));
            Vec_StrFreeP( &vRes );
            break;
        }
    }
    _ = (pFile).Close();
    return vRes;
}

/**Function*************************************************************

  Synopsis    [Assigns simulation patters derived from file.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Fra_SmlInitializeGiven(p *Fra_Sml_t, vSimInfo *Vec_Str_t) {
    var pObj *Aig_Obj_t;
    var pSims *uint32;
    var i int32;
    var k int32;
    var nPats int32 = Vec_StrSize(vSimInfo) / Aig_ManCiNum(p.pAig);
    var nPatsPadded int32  = p.nWordsTotal * 32;
    assert(Aig_ManRegNum(p.pAig) == 0, "Aig_ManRegNum(p.pAig) == 0");
    assert(Vec_StrSize(vSimInfo) % Aig_ManCiNum(p.pAig) == 0, "Vec_StrSize(vSimInfo) % Aig_ManCiNum(p.pAig) == 0");
    assert(nPats <= nPatsPadded, "nPats <= nPatsPadded");
    // inlining of macro Aig_ManForEachCi
    // inlining of macro Vec_PtrForEachEntry
    for  i = 0; (i < Vec_PtrSize(p.pAig.vCis)) && (((pObj) = (Aig_Obj_t *)Vec_PtrEntry(p.pAig.vCis, i)), 1); i++  {
        pSims = Fra_ObjSim( p, pObj.Id );
        // clean data
        for  k = 0; k < p.nWordsTotal; k++  {
            pSims[k] = 0;
        }
        // load patterns
        for  k = 0; k < nPats; k++  {
            if Vec_StrEntry(vSimInfo, k * Aig_ManCiNum(p.pAig) + i) {
                Abc_InfoSetBit( pSims, k ); }
        }
        // pad the remaining bits with the value of the last pattern
        for  ; k < nPatsPadded; k++  {
            if Vec_StrEntry(vSimInfo, (nPats-1) * Aig_ManCiNum(p.pAig) + i) {
                Abc_InfoSetBit( pSims, k ); }
        }
    }
}

/**Function*************************************************************

  Synopsis    [Prints output values.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Fra_SmlPrintOutputs(p *Fra_Sml_t, nPatterns int32) {
    var pObj *Aig_Obj_t;
    var pSims *uint32;
    var i, k int32;
    for  k = 0; k < nPatterns; k++  {
        // inlining of macro Aig_ManForEachCo
        // inlining of macro Vec_PtrForEachEntry
        for  i = 0; (i < Vec_PtrSize(p.pAig.vCos)) && (((pObj) = (Aig_Obj_t *)Vec_PtrEntry(p.pAig.vCos, i)), 1); i++  {
            pSims = Fra_ObjSim( p, pObj.Id );
            fmt.Printf("%d", Abc_InfoHasBit( pSims, k ));
        }
        fmt.Printf("\n");               ;
    }
}

/**Function*************************************************************

  Synopsis    [Assigns simulation patters derived from file.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Fra_SmlSimulateCombGiven(pAig *Aig_Man_t, pFileName string, fCheckMiter int32, fVerbose int32) *Fra_Sml_t {
    var vSimInfo *Vec_Str_t;
    var p *Fra_Sml_t;
    var nPatterns int32;
    assert(Aig_ManRegNum(pAig) == 0, "Aig_ManRegNum(pAig) == 0");
    // read comb patterns from file
    vSimInfo = Fra_SmlSimulateReadFile( pFileName );
    if vSimInfo == nil {
        return nil; }
    if Vec_StrSize(vSimInfo) % Aig_ManCiNum(pAig) != 0 {
        fmt.Printf("File \"%s\": The number of binary digits (%d) is not divisible by the number of primary inputs (%d).\n", pFileName, Vec_StrSize(vSimInfo), Aig_ManCiNum(pAig));
        Vec_StrFree( vSimInfo );
        return nil;
    }
    p = Fra_SmlStart( pAig, 0, 1, Abc_BitWordNum(Vec_StrSize(vSimInfo) / Aig_ManCiNum(pAig)) );
    Fra_SmlInitializeGiven( p, vSimInfo );
    nPatterns = Vec_StrSize(vSimInfo) / Aig_ManCiNum(pAig);
    Vec_StrFree( vSimInfo );
    Fra_SmlSimulateOne( p );
    if fCheckMiter != 0 {
        p.fNonConstOut = Fra_SmlCheckNonConstOutputs( p ); }
    if fVerbose != 0 {
        Fra_SmlPrintOutputs( p, nPatterns ); }
    return p;
}

/**Function*************************************************************

  Synopsis    [Performs simulation of the initialized circuit.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Fra_SmlSimulateSeq(pAig *Aig_Man_t, nPref int32, nFrames int32, nWords int32, fCheckMiter int32) *Fra_Sml_t {
    var p *Fra_Sml_t;
    p = Fra_SmlStart( pAig, nPref, nFrames, nWords );
    Fra_SmlInitialize( p, 1 );
    Fra_SmlSimulateOne( p );
    if fCheckMiter != 0 {
        p.fNonConstOut = Fra_SmlCheckNonConstOutputs( p ); }
    return p;
}

/**Function*************************************************************

  Synopsis    [Creates sequential counter-example from the simulation info.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Fra_SmlGetCounterExample(p *Fra_Sml_t) *Abc_Cex_t {
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
    // inlining of macro Aig_ManForEachPoSeq
    // inlining of macro Vec_PtrForEachEntryStop
    for  iPo = 0; (iPo < Aig_ManCoNum(p.pAig)-Aig_ManRegNum(p.pAig)) && (((pObj) = (Aig_Obj_t *)Vec_PtrEntry(p.pAig.vCos, iPo)), 1); iPo++  {
        if Fra_SmlNodeIsZero(p, pObj) != 0 {
            continue; }
        pSims = Fra_ObjSim( p, pObj.Id );
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
    // inlining of macro Aig_ManForEachLoSeq
    // inlining of macro Vec_PtrForEachEntryStart
    for  k = Aig_ManCiNum(p.pAig)-Aig_ManRegNum(p.pAig); (k < Vec_PtrSize(p.pAig.vCis)) && (((pObj) = (Aig_Obj_t *)Vec_PtrEntry(p.pAig.vCis, k)), 1); k++  {
        pSims = Fra_ObjSim( p, pObj.Id );
        if Abc_InfoHasBit( pSims, iBit ) {
            Abc_InfoSetBit( pCex.pData, k ); }
    }
    for  i = 0; i <= iFrame; i++  {
        // inlining of macro Aig_ManForEachPiSeq
        // inlining of macro Vec_PtrForEachEntryStop
        for  k = 0; (k < Aig_ManCiNum(p.pAig)-Aig_ManRegNum(p.pAig)) && (((pObj) = (Aig_Obj_t *)Vec_PtrEntry(p.pAig.vCis, k)), 1); k++  {
            pSims = Fra_ObjSim( p, pObj.Id );
            if Abc_InfoHasBit( pSims, 32 * p.nWordsFrame * i + iBit ) {
                Abc_InfoSetBit( pCex.pData, pCex.nRegs + pCex.nPis * i + k ); }
        }
    }
    // verify the counter example
    if !Saig_ManVerifyCex( p.pAig, pCex ) {
        fmt.Printf("Fra_SmlGetCounterExample(): Counter-example is invalid.\n");
        Abc_CexFree( pCex );
        pCex = nil;
    }
    return pCex;
}
 
/**Function*************************************************************

  Synopsis    [Generates seq counter-example from the combinational one.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Fra_SmlCopyCounterExample(pAig *Aig_Man_t, pFrames *Aig_Man_t, pModel *int32) *Abc_Cex_t {
    var pCex *Abc_Cex_t;
    var pObj *Aig_Obj_t;
    var i, nFrames, nTruePis, nTruePos, iPo, iFrame int32;
    // get the number of frames
    assert(Aig_ManRegNum(pAig) > 0, "Aig_ManRegNum(pAig) > 0");
    assert(Aig_ManRegNum(pFrames) == 0, "Aig_ManRegNum(pFrames) == 0");
    nTruePis = Aig_ManCiNum(pAig)-Aig_ManRegNum(pAig);
    nTruePos = Aig_ManCoNum(pAig)-Aig_ManRegNum(pAig);
    nFrames = Aig_ManCiNum(pFrames) / nTruePis;
    assert(nTruePis * nFrames == Aig_ManCiNum(pFrames), "nTruePis * nFrames == Aig_ManCiNum(pFrames)");
    assert(nTruePos * nFrames == Aig_ManCoNum(pFrames), "nTruePos * nFrames == Aig_ManCoNum(pFrames)");
    // find the PO that failed
    iPo = -1;
    iFrame = -1;
    // inlining of macro Aig_ManForEachCo
    // inlining of macro Vec_PtrForEachEntry
    for  i = 0; (i < Vec_PtrSize(pFrames.vCos)) && (((pObj) = (Aig_Obj_t *)Vec_PtrEntry(pFrames.vCos, i)), 1); i++  {
        if pObj.Id == pModel[Aig_ManCiNum(pFrames)] {
            iPo = i % nTruePos;
            iFrame = i / nTruePos;
            break;
        }
    }
    assert(iPo >= 0, "iPo >= 0");
    // allocate the counter example
    pCex = Abc_CexAlloc( Aig_ManRegNum(pAig), nTruePis, iFrame + 1 );
    pCex.iPo    = iPo;
    pCex.iFrame = iFrame;

    // copy the bit data
    for  i = 0; i < Aig_ManCiNum(pFrames); i++  {
        if pModel[i] {
            Abc_InfoSetBit( pCex.pData, pCex.nRegs + i ); }
        if pCex.nRegs + i == pCex.nBits - 1 {
            break; }
    }

    // verify the counter example
    if !Saig_ManVerifyCex( pAig, pCex ) {
        fmt.Printf("Fra_SmlGetCounterExample(): Counter-example is invalid.\n");
        Abc_CexFree( pCex );
        pCex = nil;
    }
    return pCex;

}


////////////////////////////////////////////////////////////////////////
///                       END OF FILE                                ///
////////////////////////////////////////////////////////////////////////


//ABC_NAMESPACE_IMPL_END

