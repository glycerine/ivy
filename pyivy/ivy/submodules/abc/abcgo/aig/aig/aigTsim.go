package aig

import "fmt"

import "os"

/**CFile****************************************************************

  FileName    [aigTsim.c]

  SystemName  [ABC: Logic synthesis and verification system.]

  PackageName [AIG package.]

  Synopsis    [Ternary simulation.]

  Author      [Alan Mishchenko]
  
  Affiliation [UC Berkeley]

  Date        [Ver. 1.0. Started - April 28, 2007.]

  Revision    [$Id: aigTsim.c,v 1.00 2007/04/28 00:00:00 alanmi Exp $]

***********************************************************************/

//#include "aig.h"
//#include "aig/saig/saig.h"

//ABC_NAMESPACE_IMPL_START


////////////////////////////////////////////////////////////////////////
///                        DECLARATIONS                              ///
////////////////////////////////////////////////////////////////////////

const TSI_MAX_ROUNDS =    1000
const TSI_ONE_SERIES =     300

const AIG_XVS0 =   1
const AIG_XVS1 =   2
const AIG_XVSX =   3

func Aig_ObjSetXsim(pObj *Aig_Obj_t, Value int32) { pObj.nCuts = Value;  }
func Aig_ObjGetXsim(pObj *Aig_Obj_t) int32 { return pObj.nCuts;   }
func Aig_XsimInv(Value int32) int32 { 
    if Value == AIG_XVS0 {
        return AIG_XVS1; }
    if Value == AIG_XVS1 {
        return AIG_XVS0; }
    assert(Value == AIG_XVSX, "Value == AIG_XVSX");       
    return AIG_XVSX;
}
func Aig_XsimAnd(Value0 int32, Value1 int32) int32 { 
    if Value0 == AIG_XVS0 || Value1 == AIG_XVS0 {
        return AIG_XVS0; }
    if Value0 == AIG_XVSX || Value1 == AIG_XVSX {
        return AIG_XVSX; }
    assert(Value0 == AIG_XVS1 && Value1 == AIG_XVS1, "Value0 == AIG_XVS1 && Value1 == AIG_XVS1");
    return AIG_XVS1;
}
func Aig_XsimRand2() int32 {
    if Aig_ManRandom(0) & 1 {
        return AIG_XVS1;
    } else {
        return AIG_XVS0;
    }
}
func Aig_XsimRand3() int32 {
    var RetValue int32;
    for { 
        RetValue = Aig_ManRandom(0) & 3; 
    
if !( RetValue == 0 ) { break }
} 
    return RetValue;
}
func Aig_ObjGetXsimFanin0(pObj *Aig_Obj_t) int32 { 
    var RetValue int32;
    RetValue = Aig_ObjGetXsim(Aig_ObjFanin0(pObj));
    if Aig_ObjFaninC0(pObj) {
        return Aig_XsimInv(RetValue);
    } else {
        return RetValue;
    }
}
func Aig_ObjGetXsimFanin1(pObj *Aig_Obj_t) int32 { 
    var RetValue int32;
    RetValue = Aig_ObjGetXsim(Aig_ObjFanin1(pObj));
    if Aig_ObjFaninC1(pObj) {
        return Aig_XsimInv(RetValue);
    } else {
        return RetValue;
    }
}
func Aig_XsimPrint(pFile *os.File, Value int32) { 
    if Value == AIG_XVS0 {
        fmt.Fprintf(pFile, "0");
        return;
    }
    if Value == AIG_XVS1 {
        fmt.Fprintf(pFile, "1");
        return;
    }
    assert(Value == AIG_XVSX, "Value == AIG_XVSX");       
    fmt.Fprintf(pFile, "x");
}

// simulation manager
   
type Aig_Tsi_t struct {
    pAig       *Aig_Man_t;              // the original AIG manager
    // ternary state representation
    nWords              int32;            // the number of words in the states
    vStates       *Vec_Ptr_t;           // the collection of ternary states
    pMem   *Aig_MmFixed_t;              // memory for ternary states
    // hash table for terminary states
    pBins       **uint32;
    nBins              int32;
}

func Aig_TsiNext(pState *uint32, nWords int32) *uint32 { return *((unsigned **)(pState + nWords));  }
func Aig_TsiSetNext(pState *uint32, nWords int32, pNext *uint32) { *((unsigned **)(pState + nWords)) = pNext; }

////////////////////////////////////////////////////////////////////////
///                     FUNCTION DEFINITIONS                         ///
////////////////////////////////////////////////////////////////////////

/**Function*************************************************************

  Synopsis    [Allocates simulation manager.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Aig_TsiStart(pAig *Aig_Man_t) *Aig_Tsi_t {
    var p *Aig_Tsi_t;
    // inlining of macro ABC_ALLOC
    p = ((Aig_Tsi_t *) malloc(sizeof(Aig_Tsi_t) * (size_t)(1)));
    memset( p, 0, sizeof(Aig_Tsi_t) );
    p.pAig    = pAig;
    p.nWords  = Abc_BitWordNum( 2*Aig_ManRegNum(pAig) );
    p.vStates = Vec_PtrAlloc( 1000 );
    p.pMem    = Aig_MmFixedStart( sizeof(unsigned) * p.nWords + sizeof(unsigned *), 10000 );
    p.nBins   = Abc_PrimeCudd(TSI_MAX_ROUNDS/2);
    // inlining of macro ABC_ALLOC
    p.pBins   = ((unsigned * *) malloc(sizeof(unsigned *) * (size_t)(p.nBins)));
    memset( p.pBins, 0, sizeof(unsigned *) * p.nBins );
    return p;
}

/**Function*************************************************************

  Synopsis    [Deallocates simulation manager.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Aig_TsiStop(p *Aig_Tsi_t) {
    Aig_MmFixedStop( p.pMem, 0 );
    Vec_PtrFree( p.vStates );
    // inlining of macro ABC_FREE
    ((p.pBins) ? (free((char *) (p.pBins)), (p.pBins) = 0) : 0);
    // inlining of macro ABC_FREE
    ((p) ? (free((char *) (p)), (p) = 0) : 0);
}

/**Function*************************************************************

  Synopsis    [Computes hash value of the node using its simulation info.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
var spStatic_Aig_TsiStateHash_s_FPrimes [128]int32 = [128]int32{ 
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
func Aig_TsiStateHash(pState *uint32, nWords int32, nTableSize int32) int32 {
    
    var uHash uint32;
    var i int32;
    uHash = 0;
    for  i = 0; i < nWords; i++  {
        uHash ^= pState[i] * spStatic_Aig_TsiStateHash_s_FPrimes[i & 0x7F];
    }
    return uHash % nTableSize;
}

/**Function*************************************************************

  Synopsis    [Inserts value into the table.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Aig_TsiStateLookup(p *Aig_Tsi_t, pState *uint32, nWords int32) int32 {
    var pEntry *uint32;
    var Hash int32;
    Hash = Aig_TsiStateHash( pState, nWords, p.nBins );
    for  pEntry = p.pBins[Hash]; (pEntry != nil); pEntry = Aig_TsiNext(pEntry, nWords)  {
        if !memcmp( pEntry, pState, sizeof(unsigned) * nWords ) {
            return 1; }
    }
    return 0;
}

/**Function*************************************************************

  Synopsis    [Inserts value into the table.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Aig_TsiStateInsert(p *Aig_Tsi_t, pState *uint32, nWords int32) {
    var Hash int32  = Aig_TsiStateHash( pState, nWords, p.nBins );
    assert(!(Aig_TsiStateLookup( p, pState, nWords ) != 0), "!Aig_TsiStateLookup( p, pState, nWords )");
    Aig_TsiSetNext( pState, nWords, p.pBins[Hash] );
    p.pBins[Hash] = pState;    
}

/**Function*************************************************************

  Synopsis    [Inserts value into the table.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Aig_TsiStateNew(p *Aig_Tsi_t) *uint32 {
    var pState *uint32;
    pState = (unsigned *)Aig_MmFixedEntryFetch( p.pMem );
    memset( pState, 0, sizeof(unsigned) * p.nWords );
    Vec_PtrPush( p.vStates, pState );
    return pState;
}

/**Function*************************************************************

  Synopsis    [Inserts value into the table.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Aig_TsiStatePrint(p *Aig_Tsi_t, pState *uint32) {
    var i, Value, nZeros, nOnes, nDcs int32;
    for  i = 0; i < Aig_ManRegNum(p.pAig); i++  {
        Value = (Abc_InfoHasBit( pState, 2 * i + 1 ) << 1) | Abc_InfoHasBit( pState, 2 * i );
        if Value == 1 {
            fmt.Printf("0");
nZeros++;
        } else if Value == 2 {
            fmt.Printf("1");
nOnes++;
        } else if Value == 3 {
            fmt.Printf("x");
nDcs++;
        } else {
            assert(false, "0"); }
    }
    fmt.Printf(" (0=%5d, 1=%5d, x=%5d)\n", nZeros, nOnes, nDcs);
}

/**Function*************************************************************

  Synopsis    [Count constant values in the state.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Aig_TsiStateCount(p *Aig_Tsi_t, pState *uint32) int32 {
    var pObjLi, pObjLo *Aig_Obj_t;
    var i, Value, nCounter int32;
    // inlining of macro Aig_ManForEachLiLoSeq
    for  i = 0; (i < Aig_ManRegNum(p.pAig)) && (((pObjLi) = Aig_ManLi(p.pAig, i)), 1)            && (((pObjLo)=Aig_ManLo(p.pAig, i)), 1); i++  {
        Value = (Abc_InfoHasBit( pState, 2 * i + 1 ) << 1) | Abc_InfoHasBit( pState, 2 * i );
        nCounter += (Value == 1 || Value == 2);
    }
    return nCounter;
}

/**Function*************************************************************

  Synopsis    [Count constant values in the state.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Aig_TsiStateOrAll(pTsi *Aig_Tsi_t, pState *uint32) {
    var pPrev *uint32;
    var i, k int32;
    // inlining of macro Vec_PtrForEachEntry
    for  i = 0; (i < Vec_PtrSize(pTsi.vStates)) && (((pPrev) = (unsigned *)Vec_PtrEntry(pTsi.vStates, i)), 1); i++  {
        for  k = 0; k < pTsi.nWords; k++  {
            pState[k] |= pPrev[k];
        }
    }
}


/**Function*************************************************************

  Synopsis    [Cycles the circuit to create a new initial state.]

  Description [Simulates the circuit with random input for the given 
  number of timeframes to get a better initial state.]
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Aig_ManTernarySimulate(p *Aig_Man_t, fVerbose int32, fVeryVerbose int32) *Vec_Ptr_t {
    var pTsi *Aig_Tsi_t;
    var vMap *Vec_Ptr_t;
    var pObj, pObjLi, pObjLo *Aig_Obj_t;
    var pState *uint32;//, * pPrev;
    var i, f, fConstants, Value, nCounter, nRetired int32;
    // allocate the simulation manager
    pTsi = Aig_TsiStart( p );
    // initialize the values
    Aig_ObjSetXsim( Aig_ManConst1(p), AIG_XVS1 );
    // inlining of macro Aig_ManForEachPiSeq
    // inlining of macro Vec_PtrForEachEntryStop
    for  i = 0; (i < Aig_ManCiNum(p)-Aig_ManRegNum(p)) && (((pObj) = (Aig_Obj_t *)Vec_PtrEntry(p.vCis, i)), 1); i++  {
        Aig_ObjSetXsim( pObj, AIG_XVSX );
    }
    // inlining of macro Aig_ManForEachLoSeq
    // inlining of macro Vec_PtrForEachEntryStart
    for  i = Aig_ManCiNum(p)-Aig_ManRegNum(p); (i < Vec_PtrSize(p.vCis)) && (((pObj) = (Aig_Obj_t *)Vec_PtrEntry(p.vCis, i)), 1); i++  {
        Aig_ObjSetXsim( pObj, AIG_XVS0 );
    }
    // simulate for the given number of timeframes
    for  f = 0; f < TSI_MAX_ROUNDS; f++  {
        // collect this state
        pState = Aig_TsiStateNew( pTsi );
        // inlining of macro Aig_ManForEachLiLoSeq
        for  i = 0; (i < Aig_ManRegNum(p)) && (((pObjLi) = Aig_ManLi(p, i)), 1)            && (((pObjLo)=Aig_ManLo(p, i)), 1); i++  {
            Value = Aig_ObjGetXsim(pObjLo);
            if Value & 1 {
                Abc_InfoSetBit( pState, 2 * i ); }
            if Value & 2 {
                Abc_InfoSetBit( pState, 2 * i + 1 ); }
        }

//        printf( "%d ", Aig_TsiStateCount(pTsi, pState) );
if fVeryVerbose != 0 {
fmt.Printf("%3d : ", f);
Aig_TsiStatePrint( pTsi, pState );
}
        // check if this state exists
        if Aig_TsiStateLookup( pTsi, pState, pTsi.nWords ) != 0 {
            break; }
//        nCounter = 0;
//        Aig_ManForEachLiLoSeq( p, pObjLi, pObjLo, i )
//            nCounter += (Aig_ObjGetXsim(pObjLo) == AIG_XVS0);
//printf( "%d -> ", nCounter );
        // insert this state
        Aig_TsiStateInsert( pTsi, pState, pTsi.nWords );
        // simulate internal nodes
        // inlining of macro Aig_ManForEachNode
        // inlining of macro Vec_PtrForEachEntry
        for  i = 0; (i < Vec_PtrSize(p.vObjs)) && (((pObj) = (Aig_Obj_t *)Vec_PtrEntry(p.vObjs, i)), 1); i++  { if (pObj) == nil || !Aig_ObjIsNode(pObj) {} else {
            Aig_ObjSetXsim( pObj, Aig_XsimAnd(Aig_ObjGetXsimFanin0(pObj), Aig_ObjGetXsimFanin1(pObj)) );
//            printf( "%d %d    Id = %2d.  Value = %d.\n", 
//                Aig_ObjGetXsimFanin0(pObj), Aig_ObjGetXsimFanin1(pObj),
//                i, Aig_XsimAnd(Aig_ObjGetXsimFanin0(pObj), Aig_ObjGetXsimFanin1(pObj)) );
        }
        }
        // transfer the latch values
        // inlining of macro Aig_ManForEachLiSeq
        // inlining of macro Vec_PtrForEachEntryStart
        for  i = Aig_ManCoNum(p)-Aig_ManRegNum(p); (i < Vec_PtrSize(p.vCos)) && (((pObj) = (Aig_Obj_t *)Vec_PtrEntry(p.vCos, i)), 1); i++  {
            Aig_ObjSetXsim( pObj, Aig_ObjGetXsimFanin0(pObj) );
        }
        nCounter = 0;
        nRetired = 0;
        // inlining of macro Aig_ManForEachLiLoSeq
        for  i = 0; (i < Aig_ManRegNum(p)) && (((pObjLi) = Aig_ManLi(p, i)), 1)            && (((pObjLo)=Aig_ManLo(p, i)), 1); i++  {
            if f < TSI_ONE_SERIES {
                Aig_ObjSetXsim( pObjLo, Aig_ObjGetXsim(pObjLi) );
            } else {
                if Aig_ObjGetXsim(pObjLi) != Aig_ObjGetXsim(pObjLo) {
                    Aig_ObjSetXsim( pObjLo, AIG_XVSX );
                    nRetired++;
                }
            }
            nCounter += (Aig_ObjGetXsim(pObjLo) == AIG_XVS0);
        }
//        if ( nRetired )
//        printf( "Retired %d registers.\n", nRetired );

//        if ( f && (f % 1000 == 0) )
//            printf( "%d \n", f );
//printf( "%d  ", nCounter );
    }
//printf( "\n" );
    if f == TSI_MAX_ROUNDS {
        fmt.Printf("Aig_ManTernarySimulate(): Did not reach a fixed point after %d iterations (not a bug).\n", TSI_MAX_ROUNDS);
        Aig_TsiStop( pTsi );
        return nil;
    }
    // OR all the states
    pState = (unsigned *)Vec_PtrEntry( pTsi.vStates, 0 );
    Aig_TsiStateOrAll( pTsi, pState );
    // check if there are constants
    fConstants = 0;
    if 2*Aig_ManRegNum(p) == 32*pTsi.nWords {
        for  i = 0; i < pTsi.nWords; i++  {
            if pState[i] != ~0 {
                fConstants = 1; }
        }
    
    } else {
        for  i = 0; i < pTsi.nWords - 1; i++  {
            if pState[i] != ~0 {
                fConstants = 1; }
        }
        if pState[i] != Abc_InfoMask( 2*Aig_ManRegNum(p) - 32*(pTsi.nWords-1) ) {
            fConstants = 1; }
    }
    if fConstants == 0 {
        if fVerbose != 0 {
        fmt.Printf("Detected 0 constants after %d iterations of ternary simulation.\n", f); }
        Aig_TsiStop( pTsi );
        return nil;
    }

    // start mapping by adding the true PIs
    vMap = Vec_PtrAlloc( Aig_ManCiNum(p) );
    // inlining of macro Aig_ManForEachPiSeq
    // inlining of macro Vec_PtrForEachEntryStop
    for  i = 0; (i < Aig_ManCiNum(p)-Aig_ManRegNum(p)) && (((pObj) = (Aig_Obj_t *)Vec_PtrEntry(p.vCis, i)), 1); i++  {
        Vec_PtrPush( vMap, pObj );
    }
    // find constant registers
    nCounter = 0;
    // inlining of macro Aig_ManForEachLiLoSeq
    for  i = 0; (i < Aig_ManRegNum(p)) && (((pObjLi) = Aig_ManLi(p, i)), 1)            && (((pObjLo)=Aig_ManLo(p, i)), 1); i++  {
        Value = (Abc_InfoHasBit( pState, 2 * i + 1 ) << 1) | Abc_InfoHasBit( pState, 2 * i );
        nCounter += (Value == 1 || Value == 2);
        if Value == 1 {
            Vec_PtrPush( vMap, Aig_ManConst0(p) );
        } else if Value == 2 {
            Vec_PtrPush( vMap, Aig_ManConst1(p) );
        } else if Value == 3 {
            Vec_PtrPush( vMap, pObjLo );
        } else {
            assert(false, "0"); }
//        Aig_XsimPrint( stdout, Value );
    }
//    printf( "\n" );
    Aig_TsiStop( pTsi );
    if fVerbose != 0 {
    fmt.Printf("Detected %d constants after %d iterations of ternary simulation.\n", nCounter, f); }
    return vMap;
}

/**Function*************************************************************

  Synopsis    [Reduces the circuit using ternary simulation.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Aig_ManConstReduce(p *Aig_Man_t, fUseMvSweep int32, nFramesSymb int32, nFramesSatur int32, fVerbose int32, fVeryVerbose int32) *Aig_Man_t {
    var pTemp *Aig_Man_t;
    var vMap *Vec_Ptr_t;
    for  Aig_ManRegNum(p) > 0  {
        if fUseMvSweep != 0 {
            vMap = Saig_MvManSimulate( p, nFramesSymb, nFramesSatur, fVerbose, fVeryVerbose );
        } else {
            vMap = Aig_ManTernarySimulate( p, fVerbose, fVeryVerbose ); }
        if vMap == nil {
            break; }
        p = Aig_ManRemap( pTemp = p, vMap );
        Vec_PtrFree( vMap );
        Aig_ManSeqCleanup( p );
        if fVerbose != 0 {
            Aig_ManReportImprovement( pTemp, p ); }
        Aig_ManStop( pTemp );
    }
    return p;
}

////////////////////////////////////////////////////////////////////////
///                       END OF FILE                                ///
////////////////////////////////////////////////////////////////////////


//ABC_NAMESPACE_IMPL_END

