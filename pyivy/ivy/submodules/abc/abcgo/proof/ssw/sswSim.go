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
type Ssw_Sml_t_ struct {
    pAig       *Aig_Man_t;              // the original AIG manager
    nPref              int32;             // the number of timeframes in the prefix
    nFrames              int32;           // the number of timeframes 
    nWordsFrame              int32;       // the number of words in each timeframe
    nWordsTotal              int32;       // the total number of words at a node
    nWordsPref              int32;        // the number of word in the prefix
    fNonConstOut              int32;      // have seen a non-const-0 output during simulation
    nSimRounds              int32;        // statistics
    timeSim          abctime;           // statistics
    pData [0]uint32;          // simulation data for the nodes
}

func Ssw_ObjSim(p *Ssw_Sml_t, Id int32) *uint32 { return p.pData + p.nWordsTotal * Id; }
func Ssw_ObjRandomSim() uint32 { return Aig_ManRandom(0);               }

////////////////////////////////////////////////////////////////////////
///                     FUNCTION DEFINITIONS                         ///
////////////////////////////////////////////////////////////////////////

/**Function*************************************************************

  Synopsis    [Computes hash value of the node using its simulation info.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
var spStatic_Ssw_SmlObjHashWord_s_SPrimes [128]int32 = [128]int32{
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
func Ssw_SmlObjHashWord(p *Ssw_Sml_t, pObj *Aig_Obj_t) uint32 {
    
    var pSims *uint32;
    var uHash uint32;
    var i int32;
//    assert( p->nWordsTotal <= 128 );
    uHash = 0;
    pSims = Ssw_ObjSim(p, pObj.Id);
    for  i = p.nWordsPref; i < p.nWordsTotal; i++  {
        uHash ^= pSims[i] * spStatic_Ssw_SmlObjHashWord_s_SPrimes[i & 0x7F];
    }
    return uHash;
}

/**Function*************************************************************

  Synopsis    [Returns 1 if simulation info is composed of all zeros.]

  Description []

  SideEffects []

  SeeAlso     []

***********************************************************************/
func Ssw_SmlObjIsConstWord(p *Ssw_Sml_t, pObj *Aig_Obj_t) int32 {
    var pSims *uint32;
    var i int32;
    pSims = Ssw_ObjSim(p, pObj.Id);
    for  i = p.nWordsPref; i < p.nWordsTotal; i++  {
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
func Ssw_SmlObjsAreEqualWord(p *Ssw_Sml_t, pObj0 *Aig_Obj_t, pObj1 *Aig_Obj_t) int32 {
    var pSims0, pSims1 *uint32;
    var i int32;
    pSims0 = Ssw_ObjSim(p, pObj0.Id);
    pSims1 = Ssw_ObjSim(p, pObj1.Id);
    for  i = p.nWordsPref; i < p.nWordsTotal; i++  {
        if pSims0[i] != pSims1[i] {
            return 0; }
    }
    return 1;
}

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
func Ssw_SmlNodeNotEquWeight(p *Ssw_Sml_t, Left int32, Right int32) int32 {
    var pSimL, pSimR *uint32;
    var k, Counter int32;
    pSimL = Ssw_ObjSim( p, Left );
    pSimR = Ssw_ObjSim( p, Right );
    for  k = p.nWordsPref; k < p.nWordsTotal; k++  {
        Counter += Aig_WordCountOnes( pSimL[k] ^ pSimR[k] );
    }
    return Counter;
}

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
func Ssw_SmlCountEqual(p *Ssw_Sml_t, pObjLi *Aig_Obj_t, pObjLo *Aig_Obj_t) int32 {
    var pSimLi, pSimLo *uint32;
    var k, Counter int32;
    assert(pObjLo.fPhase == 0, "pObjLo.fPhase == 0");
    // pObjLi->fPhase may be 1, but the LI simulation data is not complemented!
    pSimLi   = Ssw_ObjSim( p, pObjLi.Id );
    pSimLo   = Ssw_ObjSim( p, pObjLo.Id );
    for  k = p.nWordsPref; k < p.nWordsTotal; k++  {
        Counter += Aig_WordCountOnes( ~(pSimLi[k] ^ pSimLo[k]) );
    }
    return Counter;
}

/**Function*************************************************************

  Synopsis    [Returns 1 if simulation info is composed of all zeros.]

  Description []

  SideEffects []

  SeeAlso     []

***********************************************************************/
func Ssw_SmlNodeIsZero(p *Ssw_Sml_t, pObj *Aig_Obj_t) int32 {
    var pSims *uint32;
    var i int32;
    pSims = Ssw_ObjSim(p, pObj.Id);
    for  i = p.nWordsPref; i < p.nWordsTotal; i++  {
        if pSims[i] {
            return 0; }
    }
    return 1;
}

/**Function*************************************************************

  Synopsis    [Returns 1 if simulation info is composed of all zeros.]

  Description []

  SideEffects []

  SeeAlso     []

***********************************************************************/
func Ssw_SmlNodeIsZeroFrame(p *Ssw_Sml_t, pObj *Aig_Obj_t, f int32) int32 {
    var pSims *uint32  = Ssw_ObjSim(p, pObj.Id);
    return pSims[f] == 0;
}

/**Function*************************************************************

  Synopsis    [Counts the number of one's in the pattern of the object.]

  Description []

  SideEffects []

  SeeAlso     []

***********************************************************************/
func Ssw_SmlNodeCountOnesReal(p *Ssw_Sml_t, pObj *Aig_Obj_t) int32 {
    var pSims *uint32;
    var i, Counter int32;
    pSims = Ssw_ObjSim(p, Aig_Regular(pObj).Id);
    if Aig_Regular(pObj).fPhase ^ Aig_IsComplement(pObj) {
        for  i = 0; i < p.nWordsTotal; i++  {
            Counter += Aig_WordCountOnes( ~pSims[i] );
        }
    
    } else {
        for  i = 0; i < p.nWordsTotal; i++  {
            Counter += Aig_WordCountOnes( pSims[i] );
        }
    }
    return Counter;
}

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
func Ssw_SmlAssignRandom(p *Ssw_Sml_t, pObj *Aig_Obj_t) {
    var pSims *uint32;
    var i, f int32;
    assert(Aig_ObjIsCi(pObj), "Aig_ObjIsCi(pObj)");
    pSims = Ssw_ObjSim( p, pObj.Id );
    for  i = 0; i < p.nWordsTotal; i++  {
        pSims[i] = Ssw_ObjRandomSim();
    }
    // set the first bit 0 in each frame
    assert(p.nWordsFrame * p.nFrames == p.nWordsTotal, "p.nWordsFrame * p.nFrames == p.nWordsTotal");
    for  f = 0; f < p.nFrames; f++  {
        pSims[p.nWordsFrame*f] <<= 1;
    }
}

/**Function*************************************************************

  Synopsis    [Assigns random patterns to the PI node.]

  Description []

  SideEffects []

  SeeAlso     []

***********************************************************************/
func Ssw_SmlAssignRandomFrame(p *Ssw_Sml_t, pObj *Aig_Obj_t, iFrame int32) {
    var pSims *uint32;
    var i int32;
    assert(iFrame < p.nFrames, "iFrame < p.nFrames");
    assert(Aig_ObjIsCi(pObj), "Aig_ObjIsCi(pObj)");
    pSims = Ssw_ObjSim( p, pObj.Id ) + p.nWordsFrame * iFrame;
    for  i = 0; i < p.nWordsFrame; i++  {
        pSims[i] = Ssw_ObjRandomSim();
    }
}

/**Function*************************************************************

  Synopsis    [Assigns constant patterns to the PI node.]

  Description []

  SideEffects []

  SeeAlso     []

***********************************************************************/
func Ssw_SmlObjAssignConst(p *Ssw_Sml_t, pObj *Aig_Obj_t, fConst1 int32, iFrame int32) {
    var pSims *uint32;
    var i int32;
    assert(iFrame < p.nFrames, "iFrame < p.nFrames");
    assert(Aig_ObjIsCi(pObj), "Aig_ObjIsCi(pObj)");
    pSims = Ssw_ObjSim( p, pObj.Id ) + p.nWordsFrame * iFrame;
    for  i = 0; i < p.nWordsFrame; i++  {
        pSims[i] = fConst1? ~uint32(0) : 0;
    }
}

/**Function*************************************************************

  Synopsis    [Assigns constant patterns to the PI node.]

  Description []

  SideEffects []

  SeeAlso     []

***********************************************************************/
func Ssw_SmlObjAssignConstWord(p *Ssw_Sml_t, pObj *Aig_Obj_t, fConst1 int32, iFrame int32, iWord int32) {
    var pSims *uint32;
    assert(iFrame < p.nFrames, "iFrame < p.nFrames");
    assert(iWord < p.nWordsFrame, "iWord < p.nWordsFrame");
    assert(Aig_ObjIsCi(pObj), "Aig_ObjIsCi(pObj)");
    pSims = Ssw_ObjSim( p, pObj.Id ) + p.nWordsFrame * iFrame;
    pSims[iWord] = fConst1? ~uint32(0) : 0;
}

/**Function*************************************************************

  Synopsis    [Assigns constant patterns to the PI node.]

  Description []

  SideEffects []

  SeeAlso     []

***********************************************************************/
func Ssw_SmlObjSetWord(p *Ssw_Sml_t, pObj *Aig_Obj_t, Word uint32, iWord int32, iFrame int32) {
    var pSims *uint32;
    assert(iFrame < p.nFrames, "iFrame < p.nFrames");
    assert(Aig_ObjIsCi(pObj), "Aig_ObjIsCi(pObj)");
    pSims = Ssw_ObjSim( p, pObj.Id ) + p.nWordsFrame * iFrame;
    pSims[iWord] = Word;
}

/**Function*************************************************************

  Synopsis    [Assings distance-1 simulation info for the PIs.]

  Description []

  SideEffects []

  SeeAlso     []

***********************************************************************/
func Ssw_SmlAssignDist1(p *Ssw_Sml_t, pPat *uint32) {
    var pObj *Aig_Obj_t;
    var f, i, k, Limit, nTruePis int32;
    assert(p.nFrames > 0, "p.nFrames > 0");
    if p.nFrames == 1 {
        // copy the PI info
        // inlining of macro Aig_ManForEachCi
        // inlining of macro Vec_PtrForEachEntry
        for  i = 0; (i < Vec_PtrSize(p.pAig.vCis)) && (((pObj) = (Aig_Obj_t *)Vec_PtrEntry(p.pAig.vCis, i)), 1); i++  {
            Ssw_SmlObjAssignConst( p, pObj, Abc_InfoHasBit(pPat, i), 0 );
        }
        // flip one bit
        Limit = Abc_MinInt( Aig_ManCiNum(p.pAig), p.nWordsTotal * 32 - 1 );
        for  i = 0; i < Limit; i++  {
            Abc_InfoXorBit( Ssw_ObjSim( p, Aig_ManCi(p.pAig,i).Id ), i+1 );
        }
    
    } else {
        var fUseDist1 int32;

        // copy the PI info for each frame
        nTruePis = Aig_ManCiNum(p.pAig) - Aig_ManRegNum(p.pAig);
        for  f = 0; f < p.nFrames; f++  {
            // inlining of macro Saig_ManForEachPi
            // inlining of macro Vec_PtrForEachEntryStop
            for  i = 0; (i < Saig_ManPiNum(p.pAig)) && (((pObj) = (Aig_Obj_t *)Vec_PtrEntry(p.pAig.vCis, i)), 1); i++  {
                Ssw_SmlObjAssignConst( p, pObj, Abc_InfoHasBit(pPat, nTruePis * f + i), f );
            }
        }
        // copy the latch info
        k = 0;
        // inlining of macro Saig_ManForEachLo
        for  i = 0; (i < Saig_ManRegNum(p.pAig)) && (((pObj) = (Aig_Obj_t *)Vec_PtrEntry(p.pAig.vCis, i+Saig_ManPiNum(p.pAig))), 1); i++  {
            Ssw_SmlObjAssignConst( p, pObj, Abc_InfoHasBit(pPat, nTruePis * p.nFrames + k++), 0 );
        }
//        assert( p->pFrames == NULL || nTruePis * p->nFrames + k == Aig_ManCiNum(p->pFrames) );

        // flip one bit of the last frame
        if fUseDist1 != 0 { //&& p->nFrames == 2 )
        
            Limit = Abc_MinInt( nTruePis, p.nWordsFrame * 32 - 1 );
            for  i = 0; i < Limit; i++  {
                Abc_InfoXorBit( Ssw_ObjSim( p, Aig_ManCi(p.pAig, i).Id ) + p.nWordsFrame*(p.nFrames-1), i+1 );
            }
        }
    }
}

/**Function*************************************************************

  Synopsis    [Assings distance-1 simulation info for the PIs.]

  Description []

  SideEffects []

  SeeAlso     []

***********************************************************************/
func Ssw_SmlAssignDist1Plus(p *Ssw_Sml_t, pPat *uint32) {
    var pObj *Aig_Obj_t;
    var f, i, Limit int32;
    assert(p.nFrames > 0, "p.nFrames > 0");

    // copy the pattern into the primary inputs
    // inlining of macro Aig_ManForEachCi
    // inlining of macro Vec_PtrForEachEntry
    for  i = 0; (i < Vec_PtrSize(p.pAig.vCis)) && (((pObj) = (Aig_Obj_t *)Vec_PtrEntry(p.pAig.vCis, i)), 1); i++  {
        Ssw_SmlObjAssignConst( p, pObj, Abc_InfoHasBit(pPat, i), 0 );
    }

    // set distance one PIs for the first frame
    Limit = Abc_MinInt( Saig_ManPiNum(p.pAig), p.nWordsFrame * 32 - 1 );
    for  i = 0; i < Limit; i++  {
        Abc_InfoXorBit( Ssw_ObjSim( p, Aig_ManCi(p.pAig, i).Id ), i+1 );
    }

    // create random info for the remaining timeframes
    for  f = 1; f < p.nFrames; f++  {
        // inlining of macro Saig_ManForEachPi
        // inlining of macro Vec_PtrForEachEntryStop
        for  i = 0; (i < Saig_ManPiNum(p.pAig)) && (((pObj) = (Aig_Obj_t *)Vec_PtrEntry(p.pAig.vCis, i)), 1); i++  {
            Ssw_SmlAssignRandomFrame( p, pObj, f );
        }
    }
}

/**Function*************************************************************

  Synopsis    [Simulates one node.]

  Description []

  SideEffects []

  SeeAlso     []

***********************************************************************/
func Ssw_SmlNodeSimulate(p *Ssw_Sml_t, pObj *Aig_Obj_t, iFrame int32) {
    var pSims, pSims0, pSims1 *uint32;
    var fCompl, fCompl0, fCompl1, i int32;
    assert(iFrame < p.nFrames, "iFrame < p.nFrames");
    assert(!Aig_IsComplement(pObj), "!Aig_IsComplement(pObj)");
    assert(Aig_ObjIsNode(pObj), "Aig_ObjIsNode(pObj)");
    assert(iFrame == 0 || p.nWordsFrame < p.nWordsTotal, "iFrame == 0 || p.nWordsFrame < p.nWordsTotal");
    // get hold of the simulation information
    pSims  = Ssw_ObjSim(p, pObj.Id) + p.nWordsFrame * iFrame;
    pSims0 = Ssw_ObjSim(p, Aig_ObjFanin0(pObj).Id) + p.nWordsFrame * iFrame;
    pSims1 = Ssw_ObjSim(p, Aig_ObjFanin1(pObj).Id) + p.nWordsFrame * iFrame;
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
func Ssw_SmlNodesCompareInFrame(p *Ssw_Sml_t, pObj0 *Aig_Obj_t, pObj1 *Aig_Obj_t, iFrame0 int32, iFrame1 int32) int32 {
    var pSims0, pSims1 *uint32;
    var i int32;
    assert(iFrame0 < p.nFrames, "iFrame0 < p.nFrames");
    assert(iFrame1 < p.nFrames, "iFrame1 < p.nFrames");
    assert(!Aig_IsComplement(pObj0), "!Aig_IsComplement(pObj0)");
    assert(!Aig_IsComplement(pObj1), "!Aig_IsComplement(pObj1)");
    assert(iFrame0 == 0 || p.nWordsFrame < p.nWordsTotal, "iFrame0 == 0 || p.nWordsFrame < p.nWordsTotal");
    assert(iFrame1 == 0 || p.nWordsFrame < p.nWordsTotal, "iFrame1 == 0 || p.nWordsFrame < p.nWordsTotal");
    // get hold of the simulation information
    pSims0  = Ssw_ObjSim(p, pObj0.Id) + p.nWordsFrame * iFrame0;
    pSims1  = Ssw_ObjSim(p, pObj1.Id) + p.nWordsFrame * iFrame1;
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
func Ssw_SmlNodeCopyFanin(p *Ssw_Sml_t, pObj *Aig_Obj_t, iFrame int32) {
    var pSims, pSims0 *uint32;
    var fCompl, fCompl0, i int32;
    assert(iFrame < p.nFrames, "iFrame < p.nFrames");
    assert(!Aig_IsComplement(pObj), "!Aig_IsComplement(pObj)");
    assert(Aig_ObjIsCo(pObj), "Aig_ObjIsCo(pObj)");
    assert(iFrame == 0 || p.nWordsFrame < p.nWordsTotal, "iFrame == 0 || p.nWordsFrame < p.nWordsTotal");
    // get hold of the simulation information
    pSims  = Ssw_ObjSim(p, pObj.Id) + p.nWordsFrame * iFrame;
    pSims0 = Ssw_ObjSim(p, Aig_ObjFanin0(pObj).Id) + p.nWordsFrame * iFrame;
    // get complemented attributes of the children using their random info
    fCompl  = pObj.fPhase;
    fCompl0 = Aig_ObjPhaseReal(Aig_ObjChild0(pObj));
    // copy information as it is
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
func Ssw_SmlNodeTransferNext(p *Ssw_Sml_t, pOut *Aig_Obj_t, pIn *Aig_Obj_t, iFrame int32) {
    var pSims0, pSims1 *uint32;
    var i int32;
    assert(iFrame < p.nFrames, "iFrame < p.nFrames");
    assert(!Aig_IsComplement(pOut), "!Aig_IsComplement(pOut)");
    assert(!Aig_IsComplement(pIn), "!Aig_IsComplement(pIn)");
    assert(Aig_ObjIsCo(pOut), "Aig_ObjIsCo(pOut)");
    assert(Aig_ObjIsCi(pIn), "Aig_ObjIsCi(pIn)");
    assert(iFrame == 0 || p.nWordsFrame < p.nWordsTotal, "iFrame == 0 || p.nWordsFrame < p.nWordsTotal");
    // get hold of the simulation information
    pSims0 = Ssw_ObjSim(p, pOut.Id) + p.nWordsFrame * iFrame;
    pSims1 = Ssw_ObjSim(p, pIn.Id) + p.nWordsFrame * (iFrame+1);
    // copy information as it is
    for  i = 0; i < p.nWordsFrame; i++  {
        pSims1[i] = pSims0[i];
    }
}

/**Function*************************************************************

  Synopsis    [Simulates one node.]

  Description []

  SideEffects []

  SeeAlso     []

***********************************************************************/
func Ssw_SmlNodeTransferFirst(p *Ssw_Sml_t, pOut *Aig_Obj_t, pIn *Aig_Obj_t) {
    var pSims0, pSims1 *uint32;
    var i int32;
    assert(!Aig_IsComplement(pOut), "!Aig_IsComplement(pOut)");
    assert(!Aig_IsComplement(pIn), "!Aig_IsComplement(pIn)");
    assert(Aig_ObjIsCo(pOut), "Aig_ObjIsCo(pOut)");
    assert(Aig_ObjIsCi(pIn), "Aig_ObjIsCi(pIn)");
    assert(p.nWordsFrame < p.nWordsTotal, "p.nWordsFrame < p.nWordsTotal");
    // get hold of the simulation information
    pSims0 = Ssw_ObjSim(p, pOut.Id) + p.nWordsFrame * (p.nFrames-1);
    pSims1 = Ssw_ObjSim(p, pIn.Id);
    // copy information as it is
    for  i = 0; i < p.nWordsFrame; i++  {
        pSims1[i] = pSims0[i];
    }
}


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
func Ssw_SmlSimulateOne(p *Ssw_Sml_t) {
    var pObj, pObjLi, pObjLo *Aig_Obj_t;
    var f, i int32;
    var clk abctime;
clk = Abc_Clock();
    for  f = 0; f < p.nFrames; f++  {
        // simulate the nodes
        // inlining of macro Aig_ManForEachNode
        // inlining of macro Vec_PtrForEachEntry
        for  i = 0; (i < Vec_PtrSize(p.pAig.vObjs)) && (((pObj) = (Aig_Obj_t *)Vec_PtrEntry(p.pAig.vObjs, i)), 1); i++  { if (pObj) == nil || !Aig_ObjIsNode(pObj) {} else {
            Ssw_SmlNodeSimulate( p, pObj, f ); }
        }
        // copy simulation info into outputs
        // inlining of macro Saig_ManForEachPo
        // inlining of macro Vec_PtrForEachEntryStop
        for  i = 0; (i < Saig_ManPoNum(p.pAig)) && (((pObj) = (Aig_Obj_t *)Vec_PtrEntry(p.pAig.vCos, i)), 1); i++  {
            Ssw_SmlNodeCopyFanin( p, pObj, f );
        }
        // copy simulation info into outputs
        // inlining of macro Saig_ManForEachLi
        for  i = 0; (i < Saig_ManRegNum(p.pAig)) && (((pObj) = (Aig_Obj_t *)Vec_PtrEntry(p.pAig.vCos, i+Saig_ManPoNum(p.pAig))), 1); i++  {
            Ssw_SmlNodeCopyFanin( p, pObj, f );
        }
        // quit if this is the last timeframe
        if f == p.nFrames - 1 {
            break; }
        // copy simulation info into the inputs
        // inlining of macro Saig_ManForEachLiLo
        for  i = 0; (i < Saig_ManRegNum(p.pAig)) && (((pObjLi) = Saig_ManLi(p.pAig, i)), 1)            && (((pObjLo)=Saig_ManLo(p.pAig, i)), 1); i++  {
            Ssw_SmlNodeTransferNext( p, pObjLi, pObjLo, f );
        }
    }
p.timeSim += Abc_Clock() - clk;
p.nSimRounds++;
}

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

  Synopsis    [Allocates simulation manager.]

  Description []

  SideEffects []

  SeeAlso     []

***********************************************************************/
func Ssw_SmlStart(pAig *Aig_Man_t, nPref int32, nFrames int32, nWordsFrame int32) *Ssw_Sml_t {
    var p *Ssw_Sml_t;
    // inlining of macro ABC_ALLOC
    p = (Ssw_Sml_t *)((char *) malloc(sizeof(char) * (size_t)(sizeof(Ssw_Sml_t) + sizeof(unsigned) * Aig_ManObjNumMax(pAig) * (nPref + nFrames) * nWordsFrame)));
    memset( p, 0, sizeof(Ssw_Sml_t) + sizeof(unsigned) * (nPref + nFrames) * nWordsFrame );
    p.pAig        = pAig;
    p.nPref       = nPref;
    p.nFrames     = nPref + nFrames;
    p.nWordsFrame = nWordsFrame;
    p.nWordsTotal = (nPref + nFrames) * nWordsFrame;
    p.nWordsPref  = nPref * nWordsFrame;
    return p;
}

/**Function*************************************************************

  Synopsis    [Allocates simulation manager.]

  Description []

  SideEffects []

  SeeAlso     []

***********************************************************************/
func Ssw_SmlClean(p *Ssw_Sml_t) {
    memset( p.pData, 0, sizeof(unsigned) * Aig_ManObjNumMax(p.pAig) * p.nWordsTotal );
}

/**Function*************************************************************

  Synopsis    [Get simulation data.]

  Description []

  SideEffects []

  SeeAlso     []

***********************************************************************/
func Ssw_SmlSimDataPointers(p *Ssw_Sml_t) *Vec_Ptr_t {
    var vSimInfo *Vec_Ptr_t;
    var pObj *Aig_Obj_t;
    var i int32;
    vSimInfo = Vec_PtrStart( Aig_ManObjNumMax(p.pAig) );
    // inlining of macro Aig_ManForEachObj
    // inlining of macro Vec_PtrForEachEntry
    for  i = 0; (i < Vec_PtrSize(p.pAig.vObjs)) && (((pObj) = (Aig_Obj_t *)Vec_PtrEntry(p.pAig.vObjs, i)), 1); i++  { if (pObj) == nil {} else {
        Vec_PtrWriteEntry( vSimInfo, i, Ssw_ObjSim(p, i) ); }
    }
    return vSimInfo;
}

/**Function*************************************************************

  Synopsis    [Deallocates simulation manager.]

  Description []

  SideEffects []

  SeeAlso     []

***********************************************************************/
func Ssw_SmlStop(p *Ssw_Sml_t) {
    // inlining of macro ABC_FREE
    ((p) ? (free((char *) (p)), (p) = 0) : 0);
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

  Synopsis    [Returns the number of frames simulated in the manager.]

  Description []

  SideEffects []

  SeeAlso     []

***********************************************************************/
func Ssw_SmlNumFrames(p *Ssw_Sml_t) int32 {
    return p.nFrames;
}

/**Function*************************************************************

  Synopsis    [Returns the total number of simulation words.]

  Description []

  SideEffects []

  SeeAlso     []

***********************************************************************/
func Ssw_SmlNumWordsTotal(p *Ssw_Sml_t) int32 {
    return p.nWordsTotal;
}

/**Function*************************************************************

  Synopsis    [Returns the pointer to the simulation info of the node.]

  Description [The simulation info is normalized unless procedure
  Ssw_SmlUnnormalize() is called in advance.]

  SideEffects []

  SeeAlso     []

***********************************************************************/
func Ssw_SmlSimInfo(p *Ssw_Sml_t, pObj *Aig_Obj_t) *uint32 {
    assert(!Aig_IsComplement(pObj), "!Aig_IsComplement(pObj)");
    return Ssw_ObjSim( p, pObj.Id );
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
