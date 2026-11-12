package saig

import "fmt"

import "os"

/**CFile****************************************************************

  FileName    [saigPhase.c]

  SystemName  [ABC: Logic synthesis and verification system.]

  PackageName [Sequential AIG package.]

  Synopsis    [Automated phase abstraction.]

  Author      [Alan Mishchenko]
  
  Affiliation [UC Berkeley]

  Date        [Ver. 1.0. Started - June 20, 2005.]

  Revision    [$Id: saigPhase.c,v 1.00 2005/06/20 00:00:00 alanmi Exp $]

***********************************************************************/

//#include "saig.h"

//ABC_NAMESPACE_IMPL_START


/*
    The algorithm is described in the paper: Per Bjesse and Jim Kukula,
    "Automatic Phase Abstraction for Formal Verification", ICCAD 2005
    http://www.iccad.com/data2/iccad/iccad_05acceptedpapers.nsf/9cfb1ebaaf59043587256a6a00031f78/1701ecf34b149e958725702f00708828?OpenDocument
*/

// the maximum number of cycles of termiry simulation
const TSIM_MAX_ROUNDS =    10000
const TSIM_ONE_SERIES =     3000

const SAIG_XVS0 =   1
const SAIG_XVS1 =   2
const SAIG_XVSX =   3

func Saig_XsimConvertValue(v int32) int32 { if v == 0 {
    return SAIG_XVS0;
} else {
    if v == 1 {
        return SAIG_XVS1;
    } else {
        if v == 2 {
            return SAIG_XVSX;
        } else {
            return -1;
        }
    }
}  }

func Saig_ObjSetXsim(pObj *Aig_Obj_t, Value int32) { pObj.nCuts = Value;  }
func Saig_ObjGetXsim(pObj *Aig_Obj_t) int32 { return pObj.nCuts;   }
func Saig_XsimInv(Value int32) int32 { 
    if Value == SAIG_XVS0 {
        return SAIG_XVS1; }
    if Value == SAIG_XVS1 {
        return SAIG_XVS0; }
    assert(Value == SAIG_XVSX, "Value == SAIG_XVSX");       
    return SAIG_XVSX;
}
func Saig_XsimAnd(Value0 int32, Value1 int32) int32 { 
    if Value0 == SAIG_XVS0 || Value1 == SAIG_XVS0 {
        return SAIG_XVS0; }
    if Value0 == SAIG_XVSX || Value1 == SAIG_XVSX {
        return SAIG_XVSX; }
    assert(Value0 == SAIG_XVS1 && Value1 == SAIG_XVS1, "Value0 == SAIG_XVS1 && Value1 == SAIG_XVS1");
    return SAIG_XVS1;
}
func Saig_XsimRand2() int32 {
    if Aig_ManRandom(0) & 1 {
        return SAIG_XVS1;
    } else {
        return SAIG_XVS0;
    }
}
func Saig_XsimRand3() int32 {
    var RetValue int32;
    for { 
        RetValue = Aig_ManRandom(0) & 3; 
    
if !( RetValue == 0 ) { break }
} 
    return RetValue;
}
func Saig_ObjGetXsimFanin0(pObj *Aig_Obj_t) int32 { 
    var RetValue int32;
    RetValue = Saig_ObjGetXsim(Aig_ObjFanin0(pObj));
    if Aig_ObjFaninC0(pObj) {
        return Saig_XsimInv(RetValue);
    } else {
        return RetValue;
    }
}
func Saig_ObjGetXsimFanin1(pObj *Aig_Obj_t) int32 { 
    var RetValue int32;
    RetValue = Saig_ObjGetXsim(Aig_ObjFanin1(pObj));
    if Aig_ObjFaninC1(pObj) {
        return Saig_XsimInv(RetValue);
    } else {
        return RetValue;
    }
}
func Saig_XsimPrint(pFile *os.File, Value int32) { 
    if Value == SAIG_XVS0 {
        fmt.Fprintf(pFile, "0");
        return;
    }
    if Value == SAIG_XVS1 {
        fmt.Fprintf(pFile, "1");
        return;
    }
    assert(Value == SAIG_XVSX, "Value == SAIG_XVSX");       
    fmt.Fprintf(pFile, "x");
}

// simulation manager
   
type Saig_Tsim_t struct {
    pAig       *Aig_Man_t;              // the original AIG manager
    nWords              int32;            // the number of words in the states
    // ternary state representation
    vStates       *Vec_Ptr_t;           // the collection of ternary states
    pMem   *Aig_MmFixed_t;              // memory for ternary states
    nPrefix              int32;           // prefix of the ternary state space
    nCycle              int32;            // cycle of the ternary state space
    nNonXRegs              int32;         // the number of candidate registers
    vNonXRegs       *Vec_Int_t;         // the candidate registers
    // hash table for terminary states
    pBins       **uint32;
    nBins              int32;
}

func Saig_TsiNext(pState *uint32, nWords int32) *uint32 { return *((unsigned **)(pState + nWords));  }
func Saig_TsiSetNext(pState *uint32, nWords int32, pNext *uint32) { *((unsigned **)(pState + nWords)) = pNext; }

////////////////////////////////////////////////////////////////////////
///                        DECLARATIONS                              ///
////////////////////////////////////////////////////////////////////////

////////////////////////////////////////////////////////////////////////
///                     FUNCTION DEFINITIONS                         ///
////////////////////////////////////////////////////////////////////////

/**Function*************************************************************

  Synopsis    [Allocates simulation manager.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Saig_TsiStart(pAig *Aig_Man_t) *Saig_Tsim_t {
    var p *Saig_Tsim_t;
    // inlining of macro ABC_ALLOC
    p = (Saig_Tsim_t *)((char *) malloc(sizeof(char) * (size_t)(sizeof(Saig_Tsim_t))));
    memset( p, 0, sizeof(Saig_Tsim_t) );
    p.pAig    = pAig;
    p.nWords  = Abc_BitWordNum( 2*Aig_ManRegNum(pAig) );
    p.vStates = Vec_PtrAlloc( 1000 );
    p.pMem    = Aig_MmFixedStart( sizeof(unsigned) * p.nWords + sizeof(unsigned *), 10000 );
    p.nBins   = Abc_PrimeCudd(TSIM_MAX_ROUNDS/2);
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
func Saig_TsiStop(p *Saig_Tsim_t) {
    if p.vNonXRegs {
        Vec_IntFree( p.vNonXRegs ); }
    Aig_MmFixedStop( p.pMem, 0 );
    Vec_PtrFree( p.vStates );
    p.pBins = nil;
    p = nil;
}

/**Function*************************************************************

  Synopsis    [Computes hash value of the node using its simulation info.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
var spStatic_Saig_TsiStateHash_s_FPrimes [128]int32 = [128]int32{ 
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
func Saig_TsiStateHash(pState *uint32, nWords int32, nTableSize int32) int32 {
    
    var uHash uint32;
    var i int32;
    uHash = 0;
    for  i = 0; i < nWords; i++  {
        uHash ^= pState[i] * spStatic_Saig_TsiStateHash_s_FPrimes[i & 0x7F];
    }
    return uHash % nTableSize;
}

/**Function*************************************************************

  Synopsis    [Count non-X-valued registers in the simulation data.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Saig_TsiCountNonXValuedRegisters(p *Saig_Tsim_t, nPref int32) int32 {
    var pState *uint32;
    var nRegs int32  = p.pAig.nRegs;
    var Value, i, k int32;
    assert(p.vNonXRegs == nil, "p.vNonXRegs == nil");
    p.vNonXRegs = Vec_IntAlloc( 10 );
    for  i = 0; i < nRegs; i++  {
        // inlining of macro Vec_PtrForEachEntryStart
        for  k = nPref; (k < Vec_PtrSize(p.vStates)) && (((pState) = (unsigned *)Vec_PtrEntry(p.vStates, k)), 1); k++  {
            Value = (Abc_InfoHasBit( pState, 2 * i + 1 ) << 1) | Abc_InfoHasBit( pState, 2 * i );
            assert(Value != 0, "Value != 0");
            if Value == SAIG_XVSX {
                break; }
        }
        if k == Vec_PtrSize(p.vStates) {
            Vec_IntPush( p.vNonXRegs, i ); }
    }
    return Vec_IntSize(p.vNonXRegs);
} 

/**Function*************************************************************

  Synopsis    [Computes flops that are stuck-at constant.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Saig_TsiComputeTransient(p *Saig_Tsim_t, nPref int32) *Vec_Int_t {
    var vCounters *Vec_Int_t;
    var pState *uint32;
    var ValueThis int32 = -1;
    var ValuePrev int32 = -1;
    var StepPrev int32 = -1;
    var i int32;
    var k int32;
    var nRegs int32 = p.pAig.nRegs;
    vCounters = Vec_IntStart( nPref );
    for  i = 0; i < nRegs; i++  {
        // inlining of macro Vec_PtrForEachEntry
        for  k = 0; (k < Vec_PtrSize(p.vStates)) && (((pState) = (unsigned *)Vec_PtrEntry(p.vStates, k)), 1); k++  {
            ValueThis = (Abc_InfoHasBit( pState, 2 * i + 1 ) << 1) | Abc_InfoHasBit( pState, 2 * i );
//printf( "%s", (ValueThis == 1)? "0" : ((ValueThis == 2)? "1" : "x") );
            assert(ValueThis != 0, "ValueThis != 0");
            if ValuePrev != ValueThis {
                ValuePrev = ValueThis;
                StepPrev  = k;
            }
        }
//printf( "\n" );
        if ValueThis == SAIG_XVSX {
            continue; }
        if StepPrev >= nPref {
            continue; }
        Vec_IntAddToEntry( vCounters, StepPrev, 1 );
    }
    // inlining of macro Vec_IntForEachEntry
    for  i = 0; (i < Vec_IntSize(vCounters)) && (((ValueThis) = Vec_IntEntry(vCounters, i)), 1); i++  {
        if ValueThis == 0 {
            continue; }
//        printf( "%3d : %3d\n", i, ValueThis );
    }
    return vCounters;
}

/**Function*************************************************************

  Synopsis    [Count non-X-valued registers in the simulation data.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Saig_TsiPrintTraces(p *Saig_Tsim_t, nWords int32, nPrefix int32, nLoop int32) {
    var pState *uint32;
    var nRegs int32  = p.pAig.nRegs;
    var Value, i, k, Counter int32;
    fmt.Printf("Ternary traces for each flop:\n");
    fmt.Printf("      : ");
    for  i = 0; i < Vec_PtrSize(p.vStates) - nLoop - 1; i++  {
        fmt.Printf("%d", i%10);
    }
    fmt.Printf("  ");
    for  i = 0; i < nLoop; i++  {
        fmt.Printf("%d", i%10);
    }
    fmt.Printf("\n");
    for  i = 0; i < nRegs; i++  {
/*
        Vec_PtrForEachEntry( unsigned *, p->vStates, pState, k )
        {
            Value = (Abc_InfoHasBit( pState, 2 * i + 1 ) << 1) | Abc_InfoHasBit( pState, 2 * i );
            if ( Value == SAIG_XVSX )
                break;
        }
        if ( k == Vec_PtrSize(p->vStates) )
            Counter++;
        else
            continue;
*/

        // print trace
//        printf( "%5d : %5d %5d  ", Counter, i, Saig_ManLo(p->pAig, i)->Id );
        fmt.Printf("%5d : ", Counter++);
        // inlining of macro Vec_PtrForEachEntryStop
        for  k = 0; (k < Vec_PtrSize(p.vStates)-1) && (((pState) = (unsigned *)Vec_PtrEntry(p.vStates, k)), 1); k++  {
            Value = (Abc_InfoHasBit( pState, 2 * i + 1 ) << 1) | Abc_InfoHasBit( pState, 2 * i );
            if Value == SAIG_XVS0 {
                fmt.Printf("0");
            } else if Value == SAIG_XVS1 {
                fmt.Printf("1");
            } else if Value == SAIG_XVSX {
                fmt.Printf("x");
            } else {
                assert(false, "0"); }
            if k == nPrefix - 1 {
                fmt.Printf("  "); }
        }
        fmt.Printf("\n");
    }
}

/**Function*************************************************************

  Synopsis    [Returns the number of the state.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Saig_TsiComputePrefix(p *Saig_Tsim_t, pState *uint32, nWords int32) int32 {
    var pEntry, pPrev *uint32;
    var Hash, i int32;
    Hash = Saig_TsiStateHash( pState, nWords, p.nBins );
    for  pEntry = p.pBins[Hash]; (pEntry != nil); pEntry = Saig_TsiNext(pEntry, nWords)  {
        if !memcmp( pEntry, pState, sizeof(unsigned) * nWords ) {
            // inlining of macro Vec_PtrForEachEntry
            for  i = 0; (i < Vec_PtrSize(p.vStates)) && (((pPrev) = (unsigned *)Vec_PtrEntry(p.vStates, i)), 1); i++  {
                if pPrev == pEntry {
                    return i; }
            }
            assert(false, "0");
            return -1;
        }
    }
    return -1;
}

/**Function*************************************************************

  Synopsis    [Checks if the value exists in the table.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Saig_TsiStateLookup(p *Saig_Tsim_t, pState *uint32, nWords int32) int32 {
    var pEntry *uint32;
    var Hash int32;
    Hash = Saig_TsiStateHash( pState, nWords, p.nBins );
    for  pEntry = p.pBins[Hash]; (pEntry != nil); pEntry = Saig_TsiNext(pEntry, nWords)  {
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
func Saig_TsiStateInsert(p *Saig_Tsim_t, pState *uint32, nWords int32) {
    var Hash int32  = Saig_TsiStateHash( pState, nWords, p.nBins );
    assert(!(Saig_TsiStateLookup( p, pState, nWords ) != 0), "!Saig_TsiStateLookup( p, pState, nWords )");
    Saig_TsiSetNext( pState, nWords, p.pBins[Hash] );
    p.pBins[Hash] = pState;    
}

/**Function*************************************************************

  Synopsis    [Inserts value into the table.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Saig_TsiStateNew(p *Saig_Tsim_t) *uint32 {
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
func Saig_TsiStatePrint(p *Saig_Tsim_t, pState *uint32) {
    var i, Value, nZeros, nOnes, nDcs int32;
    for  i = 0; i < Aig_ManRegNum(p.pAig); i++  {
        Value = (Abc_InfoHasBit( pState, 2 * i + 1 ) << 1) | Abc_InfoHasBit( pState, 2 * i );
        if Value == SAIG_XVS0 {
            fmt.Printf("0");
nZeros++;
        } else if Value == SAIG_XVS1 {
            fmt.Printf("1");
nOnes++;
        } else if Value == SAIG_XVSX {
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
func Saig_TsiStateCount(p *Saig_Tsim_t, pState *uint32) int32 {
    var pObjLi, pObjLo *Aig_Obj_t;
    var i, Value, nCounter int32;
    // inlining of macro Aig_ManForEachLiLoSeq
    for  i = 0; (i < Aig_ManRegNum(p.pAig)) && (((pObjLi) = Aig_ManLi(p.pAig, i)), 1)            && (((pObjLo)=Aig_ManLo(p.pAig, i)), 1); i++  {
        Value = (Abc_InfoHasBit( pState, 2 * i + 1 ) << 1) | Abc_InfoHasBit( pState, 2 * i );
        nCounter += (Value == SAIG_XVS0 || Value == SAIG_XVS1);
    }
    return nCounter;
}

/**Function*************************************************************

  Synopsis    [Count constant values in the state.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Saig_TsiStateOrAll(pTsi *Saig_Tsim_t, pState *uint32) {
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
func Saig_ManReachableTernary(p *Aig_Man_t, vInits *Vec_Int_t, fVerbose int32) *Saig_Tsim_t {
    var pTsi *Saig_Tsim_t;
    var pObj, pObjLi, pObjLo *Aig_Obj_t;
    var pState *uint32;
    var i, f, Value, nCounter int32;
    // allocate the simulation manager
    pTsi = Saig_TsiStart( p );
    // initialize the values
    Saig_ObjSetXsim( Aig_ManConst1(p), SAIG_XVS1 );
    // inlining of macro Saig_ManForEachPi
    // inlining of macro Vec_PtrForEachEntryStop
    for  i = 0; (i < Saig_ManPiNum(p)) && (((pObj) = (Aig_Obj_t *)Vec_PtrEntry(p.vCis, i)), 1); i++  {
        Saig_ObjSetXsim( pObj, SAIG_XVSX );
    }
    if vInits != nil {
        // inlining of macro Saig_ManForEachLo
        for  i = 0; (i < Saig_ManRegNum(p)) && (((pObj) = (Aig_Obj_t *)Vec_PtrEntry(p.vCis, i+Saig_ManPiNum(p))), 1); i++  {
            Saig_ObjSetXsim( pObj, Saig_XsimConvertValue(Vec_IntEntry(vInits, i)) );
        }
    
    } else {
        // inlining of macro Saig_ManForEachLo
        for  i = 0; (i < Saig_ManRegNum(p)) && (((pObj) = (Aig_Obj_t *)Vec_PtrEntry(p.vCis, i+Saig_ManPiNum(p))), 1); i++  {
            Saig_ObjSetXsim( pObj, SAIG_XVS0 );
        }
    }
    // simulate for the given number of timeframes
    for  f = 0; f < TSIM_MAX_ROUNDS; f++  {
        // collect this state
        pState = Saig_TsiStateNew( pTsi );
        // inlining of macro Saig_ManForEachLiLo
        for  i = 0; (i < Saig_ManRegNum(p)) && (((pObjLi) = Saig_ManLi(p, i)), 1)            && (((pObjLo)=Saig_ManLo(p, i)), 1); i++  {
            Value = Saig_ObjGetXsim(pObjLo);
            if Value & 1 {
                Abc_InfoSetBit( pState, 2 * i ); }
            if Value & 2 {
                Abc_InfoSetBit( pState, 2 * i + 1 ); }
        }
//        printf( "%d ", Saig_TsiStateCount(pTsi, pState) );
//        Saig_TsiStatePrint( pTsi, pState );
        // check if this state exists
        if Saig_TsiStateLookup( pTsi, pState, pTsi.nWords ) != 0 {
            if fVerbose != 0 {
                fmt.Printf("Ternary simulation converged after %d iterations.\n", f); }
            return pTsi;
        }
        // insert this state
        Saig_TsiStateInsert( pTsi, pState, pTsi.nWords );
        // simulate internal nodes
        // inlining of macro Aig_ManForEachNode
        // inlining of macro Vec_PtrForEachEntry
        for  i = 0; (i < Vec_PtrSize(p.vObjs)) && (((pObj) = (Aig_Obj_t *)Vec_PtrEntry(p.vObjs, i)), 1); i++  { if (pObj) == nil || !Aig_ObjIsNode(pObj) {} else {
            Saig_ObjSetXsim( pObj, Saig_XsimAnd(Saig_ObjGetXsimFanin0(pObj), Saig_ObjGetXsimFanin1(pObj)) ); }
        }
        // transfer the latch values
        // inlining of macro Saig_ManForEachLi
        for  i = 0; (i < Saig_ManRegNum(p)) && (((pObj) = (Aig_Obj_t *)Vec_PtrEntry(p.vCos, i+Saig_ManPoNum(p))), 1); i++  {
            Saig_ObjSetXsim( pObj, Saig_ObjGetXsimFanin0(pObj) );
        }
        nCounter = 0;
        // inlining of macro Saig_ManForEachLiLo
        for  i = 0; (i < Saig_ManRegNum(p)) && (((pObjLi) = Saig_ManLi(p, i)), 1)            && (((pObjLo)=Saig_ManLo(p, i)), 1); i++  {
            if f < TSIM_ONE_SERIES {
                Saig_ObjSetXsim( pObjLo, Saig_ObjGetXsim(pObjLi) );
            } else {
                if Saig_ObjGetXsim(pObjLi) != Saig_ObjGetXsim(pObjLo) {
                    Saig_ObjSetXsim( pObjLo, SAIG_XVSX ); }
            }
            nCounter += (Saig_ObjGetXsim(pObjLo) == SAIG_XVS0);
        }
    }
    fmt.Printf("Saig_ManReachableTernary(): Did not reach a fixed point after %d iterations (not a bug).\n", TSIM_MAX_ROUNDS);
    Saig_TsiStop( pTsi );
    return nil;
}

/**Function*************************************************************

  Synopsis    [Analize initial value of the selected register.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Saig_ManAnalizeControl(p *Aig_Man_t, Reg int32) {
    var pObj, pReg, pCtrl, pAnd *Aig_Obj_t;
    var i int32;
    pReg = Saig_ManLo( p, Reg );
    pCtrl = Saig_ManLo( p, Saig_ManRegNum(p)-1 );
    assert(pReg.Id < pCtrl.Id, "pReg.Id < pCtrl.Id");
    // find a node pointing to both
    pAnd = nil;
    // inlining of macro Aig_ManForEachNode
    // inlining of macro Vec_PtrForEachEntry
    for  i = 0; (i < Vec_PtrSize(p.vObjs)) && (((pObj) = (Aig_Obj_t *)Vec_PtrEntry(p.vObjs, i)), 1); i++  { if (pObj) == nil || !Aig_ObjIsNode(pObj) {} else {
        if Aig_ObjFanin0(pObj) == pReg && Aig_ObjFanin1(pObj) == pCtrl {
            pAnd = pObj;
            break;
        }
    }
    }
    if pAnd == nil {
        fmt.Printf("Register is not found.\n");
        return;
    }
    fmt.Printf("Clock-like register: \n");
    Aig_ObjPrint( p, pReg );
    fmt.Printf("\n");
    fmt.Printf("Control register: \n");
    Aig_ObjPrint( p, pCtrl );
    fmt.Printf("\n");
    fmt.Printf("Their fanout: \n");
    Aig_ObjPrint( p, pAnd );
    fmt.Printf("\n");
 
    // find the fanouts of pAnd
    fmt.Printf("Fanouts of the fanout: \n");
    // inlining of macro Aig_ManForEachObj
    // inlining of macro Vec_PtrForEachEntry
    for  i = 0; (i < Vec_PtrSize(p.vObjs)) && (((pObj) = (Aig_Obj_t *)Vec_PtrEntry(p.vObjs, i)), 1); i++  { if (pObj) == nil {} else
        if Aig_ObjFanin0(pObj) == pAnd || Aig_ObjFanin1(pObj) == pAnd {
            Aig_ObjPrint( p, pObj );
            fmt.Printf("\n");
        }
    }
    fmt.Printf("\n");
}

/**Function*************************************************************

  Synopsis    [Finds the registers to phase-abstract.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Saig_ManFindRegisters(pTsi *Saig_Tsim_t, nFrames int32, fIgnore int32, fVerbose int32) int32 {
    var Values [257]int32 ;
    var pState *uint32;
    var r, i, k, Reg, Value int32;
    var nTests int32  = pTsi.nPrefix + 2 * pTsi.nCycle;
    assert(nFrames <= 256, "nFrames <= 256");
    r = 0;
    // inlining of macro Vec_IntForEachEntry
    for  i = 0; (i < Vec_IntSize(pTsi.vNonXRegs)) && (((Reg) = Vec_IntEntry(pTsi.vNonXRegs, i)), 1); i++  {
        for  k = 0; k < nTests; k++  {
            if k < pTsi.nPrefix + pTsi.nCycle {
                pState = (unsigned *)Vec_PtrEntry( pTsi.vStates, k );
            } else {
                pState = (unsigned *)Vec_PtrEntry( pTsi.vStates, k - pTsi.nCycle ); }
            Value = (Abc_InfoHasBit( pState, 2 * Reg + 1 ) << 1) | Abc_InfoHasBit( pState, 2 * Reg );
            assert(Value == SAIG_XVS0 || Value == SAIG_XVS1, "Value == SAIG_XVS0 || Value == SAIG_XVS1");
            if k < nFrames || ((fIgnore != 0) && k == nFrames) {
                Values[k % nFrames] = Value;
            } else if Values[k % nFrames] != Value {
                break; }
        }
        if k < nTests {
            continue; }
        // skip stuck at
        if fIgnore != 0 {
            for  k = 1; k < nFrames; k++  {
                if Values[k] != Values[0] {
                    break; }
            }
            if k == nFrames {
                continue; }
        }
        // report useful register
        Vec_IntWriteEntry( pTsi.vNonXRegs, r++, Reg );
        if fVerbose != 0 {
            fmt.Printf("Register %5d has generator: [", Reg);
            for  k = 0; k < nFrames; k++  {
                Saig_XsimPrint( os.Stdout, Values[k] );
            }
            fmt.Printf("]\n");

            if fVerbose != 0 {
            Saig_ManAnalizeControl( pTsi.pAig, Reg ); }
        }
    }
    Vec_IntShrink( pTsi.vNonXRegs, r );
    if fVerbose != 0 {
        fmt.Printf("Found %3d useful registers.\n", Vec_IntSize(pTsi.vNonXRegs)); }
    return Vec_IntSize(pTsi.vNonXRegs);
}


/**Function*************************************************************

  Synopsis    [Mapping of AIG nodes into frames nodes.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Saig_ObjFrames(pObjMap **Aig_Obj_t, nFs int32, pObj *Aig_Obj_t, i int32) *Aig_Obj_t { return pObjMap[nFs*pObj.Id + i];  }
func Saig_ObjSetFrames(pObjMap **Aig_Obj_t, nFs int32, pObj *Aig_Obj_t, i int32, pNode *Aig_Obj_t) { pObjMap[nFs*pObj.Id + i] = pNode; }

func Saig_ObjChild0Frames(pObjMap **Aig_Obj_t, nFs int32, pObj *Aig_Obj_t, i int32) *Aig_Obj_t { if Aig_ObjFanin0(pObj) {
    return Aig_NotCond(Saig_ObjFrames(pObjMap,nFs,Aig_ObjFanin0(pObj),i), Aig_ObjFaninC0(pObj));
} else {
    return nil;
}  }
func Saig_ObjChild1Frames(pObjMap **Aig_Obj_t, nFs int32, pObj *Aig_Obj_t, i int32) *Aig_Obj_t { if Aig_ObjFanin1(pObj) {
    return Aig_NotCond(Saig_ObjFrames(pObjMap,nFs,Aig_ObjFanin1(pObj),i), Aig_ObjFaninC1(pObj));
} else {
    return nil;
}  }

/**Function*************************************************************

  Synopsis    [Performs phase abstraction by unrolling the circuit.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Saig_ManPerformAbstraction(pTsi *Saig_Tsim_t, nFrames int32, fVerbose int32) *Aig_Man_t {
    var pFrames *Aig_Man_t;
    var pAig *Aig_Man_t = pTsi.pAig;
    var pObj, pObjLi, pObjLo, pObjNew *Aig_Obj_t;
    var pObjMap **Aig_Obj_t;
    var pState *uint32;
    var i, f, Reg, Value int32;

    assert(Vec_IntSize(pTsi.vNonXRegs) > 0, "Vec_IntSize(pTsi.vNonXRegs) > 0");

    // create mapping for the frames nodes
    // inlining of macro ABC_ALLOC
    pObjMap = ((Aig_Obj_t * *) malloc(sizeof(Aig_Obj_t *) * (size_t)(nFrames * Aig_ManObjNumMax(pAig))));
    memset( pObjMap, 0, sizeof(Aig_Obj_t *) * nFrames * Aig_ManObjNumMax(pAig) );

    // start the fraig package
    pFrames = Aig_ManStart( Aig_ManObjNumMax(pAig) * nFrames );
    pFrames.pName = Abc_UtilStrsav( pAig.pName );
    pFrames.pSpec = Abc_UtilStrsav( pAig.pSpec );
    // map constant nodes
    for  f = 0; f < nFrames; f++  {
        Saig_ObjSetFrames( pObjMap, nFrames, Aig_ManConst1(pAig), f, Aig_ManConst1(pFrames) );
    }
    // create PI nodes for the frames
    for  f = 0; f < nFrames; f++  {
        // inlining of macro Aig_ManForEachPiSeq
        // inlining of macro Vec_PtrForEachEntryStop
        for  i = 0; (i < Aig_ManCiNum(pAig)-Aig_ManRegNum(pAig)) && (((pObj) = (Aig_Obj_t *)Vec_PtrEntry(pAig.vCis, i)), 1); i++  {
            Saig_ObjSetFrames( pObjMap, nFrames, pObj, f, Aig_ObjCreateCi(pFrames) );
        }
    }
    // create the latches
    // inlining of macro Aig_ManForEachLoSeq
    // inlining of macro Vec_PtrForEachEntryStart
    for  i = Aig_ManCiNum(pAig)-Aig_ManRegNum(pAig); (i < Vec_PtrSize(pAig.vCis)) && (((pObj) = (Aig_Obj_t *)Vec_PtrEntry(pAig.vCis, i)), 1); i++  {
        Saig_ObjSetFrames( pObjMap, nFrames, pObj, 0, Aig_ObjCreateCi(pFrames) );
    }

    // add timeframes
    for  f = 0; f < nFrames; f++  {
        // replace abstracted registers by constants
        // inlining of macro Vec_IntForEachEntry
        for  i = 0; (i < Vec_IntSize(pTsi.vNonXRegs)) && (((Reg) = Vec_IntEntry(pTsi.vNonXRegs, i)), 1); i++  {
            pObj = Saig_ManLo( pAig, Reg );
            pState = (unsigned *)Vec_PtrEntry( pTsi.vStates, f );
            Value = (Abc_InfoHasBit( pState, 2 * Reg + 1 ) << 1) | Abc_InfoHasBit( pState, 2 * Reg );
            assert(Value == SAIG_XVS0 || Value == SAIG_XVS1, "Value == SAIG_XVS0 || Value == SAIG_XVS1");
            if Value == SAIG_XVS1 {
                pObjNew = Aig_ManConst1(pFrames);
            } else {
                pObjNew = Aig_ManConst0(pFrames);
            }
            Saig_ObjSetFrames( pObjMap, nFrames, pObj, f, pObjNew );
        }
        // add internal nodes of this frame
        // inlining of macro Aig_ManForEachNode
        // inlining of macro Vec_PtrForEachEntry
        for  i = 0; (i < Vec_PtrSize(pAig.vObjs)) && (((pObj) = (Aig_Obj_t *)Vec_PtrEntry(pAig.vObjs, i)), 1); i++  { if (pObj) == nil || !Aig_ObjIsNode(pObj) {} else {
            pObjNew = Aig_And( pFrames, Saig_ObjChild0Frames(pObjMap,nFrames,pObj,f), Saig_ObjChild1Frames(pObjMap,nFrames,pObj,f) );
            Saig_ObjSetFrames( pObjMap, nFrames, pObj, f, pObjNew );
        }
        }
        // set the latch inputs and copy them into the latch outputs of the next frame
        // inlining of macro Aig_ManForEachLiLoSeq
        for  i = 0; (i < Aig_ManRegNum(pAig)) && (((pObjLi) = Aig_ManLi(pAig, i)), 1)            && (((pObjLo)=Aig_ManLo(pAig, i)), 1); i++  {
            pObjNew = Saig_ObjChild0Frames(pObjMap,nFrames,pObjLi,f);
            if f < nFrames - 1 {
                Saig_ObjSetFrames( pObjMap, nFrames, pObjLo, f+1, pObjNew ); }
        }
    }
    for  f = 0; f < nFrames; f++  {
        // inlining of macro Aig_ManForEachPoSeq
        // inlining of macro Vec_PtrForEachEntryStop
        for  i = 0; (i < Aig_ManCoNum(pAig)-Aig_ManRegNum(pAig)) && (((pObj) = (Aig_Obj_t *)Vec_PtrEntry(pAig.vCos, i)), 1); i++  {
            pObjNew = Aig_ObjCreateCo( pFrames, Saig_ObjChild0Frames(pObjMap,nFrames,pObj,f) );
            Saig_ObjSetFrames( pObjMap, nFrames, pObj, f, pObjNew );
        }
    }
    pFrames.nRegs = pAig.nRegs;
    pFrames.nTruePis = Aig_ManCiNum(pFrames) - Aig_ManRegNum(pFrames); 
    pFrames.nTruePos = Aig_ManCoNum(pFrames) - Aig_ManRegNum(pFrames); 
    // inlining of macro Aig_ManForEachLiSeq
    // inlining of macro Vec_PtrForEachEntryStart
    for  i = Aig_ManCoNum(pAig)-Aig_ManRegNum(pAig); (i < Vec_PtrSize(pAig.vCos)) && (((pObj) = (Aig_Obj_t *)Vec_PtrEntry(pAig.vCos, i)), 1); i++  {
        pObjNew = Aig_ObjCreateCo( pFrames, Saig_ObjChild0Frames(pObjMap,nFrames,pObj,nFrames-1) );
        Saig_ObjSetFrames( pObjMap, nFrames, pObj, nFrames-1, pObjNew );
    }
//Aig_ManPrintStats( pFrames );
    Aig_ManSeqCleanup( pFrames );
//Aig_ManPrintStats( pFrames );
//    Aig_ManCiCleanup( pFrames );
//Aig_ManPrintStats( pFrames );
    pObjMap = nil;
    return pFrames;
}


/**Function*************************************************************

  Synopsis    [Performs automated phase abstraction.]

  Description [Takes the AIG manager and the array of initial states.]
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Saig_ManPhaseFrameNum(p *Aig_Man_t, vInits *Vec_Int_t) int32 {
    var pTsi *Saig_Tsim_t;
    var nFrames, nPrefix int32;
    assert(Saig_ManRegNum(p), "Saig_ManRegNum(p)");
    assert(Saig_ManPiNum(p), "Saig_ManPiNum(p)");
    assert(Saig_ManPoNum(p), "Saig_ManPoNum(p)");
    // perform terminary simulation
    pTsi = Saig_ManReachableTernary( p, vInits, 0 );
    if pTsi == nil {
        return 1; }
    // derive information
    nPrefix = Saig_TsiComputePrefix( pTsi, (unsigned *)Vec_PtrEntryLast(pTsi.vStates), pTsi.nWords );
    nFrames = Vec_PtrSize(pTsi.vStates) - 1 - nPrefix;
    Saig_TsiStop( pTsi );
    // potentially, may need to reduce nFrames if nPrefix is less than nFrames
    return nFrames;
}

/**Function*************************************************************

  Synopsis    [Performs automated phase abstraction.]

  Description [Takes the AIG manager and the array of initial states.]
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Saig_ManPhasePrefixLength(p *Aig_Man_t, fVerbose int32, fVeryVerbose int32, pvTrans **Vec_Int_t) int32 {
    var pTsi *Saig_Tsim_t;
    var nFrames, nPrefix, nNonXRegs int32;
    assert(Saig_ManRegNum(p), "Saig_ManRegNum(p)");
    assert(Saig_ManPiNum(p), "Saig_ManPiNum(p)");
    assert(Saig_ManPoNum(p), "Saig_ManPoNum(p)");
    // perform terminary simulation
    pTsi = Saig_ManReachableTernary( p, nil, 0 );
    if pTsi == nil {
        return 0; }
    // derive information
    nPrefix   = Saig_TsiComputePrefix( pTsi, (unsigned *)Vec_PtrEntryLast(pTsi.vStates), pTsi.nWords );
    nFrames   = Vec_PtrSize(pTsi.vStates) - 1 - nPrefix;
    nNonXRegs = Saig_TsiCountNonXValuedRegisters( pTsi, nPrefix );

    if pvTrans {
        *pvTrans = Saig_TsiComputeTransient( pTsi, nPrefix ); }

    // print statistics
    if fVerbose != 0 {
        fmt.Printf("Lead = %5d. Loop = %5d.  Total flops = %5d. Binary flops = %5d.\n", nPrefix, nFrames, p.nRegs, nNonXRegs); }
    if fVeryVerbose != 0 {
        Saig_TsiPrintTraces( pTsi, pTsi.nWords, nPrefix, nFrames ); }
    Saig_TsiStop( pTsi );
    // potentially, may need to reduce nFrames if nPrefix is less than nFrames
    return nPrefix;
}

/**Function*************************************************************

  Synopsis    [Performs automated phase abstraction.]

  Description [Takes the AIG manager and the array of initial states.]
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Saig_ManPhaseAbstract(p *Aig_Man_t, vInits *Vec_Int_t, nFrames int32, nPref int32, fIgnore int32, fPrint int32, fVerbose int32) *Aig_Man_t {
    var pNew *Aig_Man_t  = nil;
    var pTsi *Saig_Tsim_t;
    assert(Saig_ManRegNum(p), "Saig_ManRegNum(p)");
    assert(Saig_ManPiNum(p), "Saig_ManPiNum(p)");
    assert(Saig_ManPoNum(p), "Saig_ManPoNum(p)");
    // perform terminary simulation
    pTsi = Saig_ManReachableTernary( p, vInits, fVerbose );
    if pTsi == nil {
        return nil; }
    // derive information
    pTsi.nPrefix = Saig_TsiComputePrefix( pTsi, (unsigned *)Vec_PtrEntryLast(pTsi.vStates), pTsi.nWords );
    pTsi.nCycle = Vec_PtrSize(pTsi.vStates) - 1 - pTsi.nPrefix;
    pTsi.nNonXRegs = Saig_TsiCountNonXValuedRegisters(pTsi, Abc_MinInt(pTsi.nPrefix,nPref));
    // print statistics
    if fVerbose != 0 {
        fmt.Printf("Lead = %5d. Loop = %5d.  Total flops = %5d. Binary flops = %5d.\n", pTsi.nPrefix, pTsi.nCycle, p.nRegs, pTsi.nNonXRegs);
        if pTsi.nNonXRegs < 100 && Vec_PtrSize(pTsi.vStates) < 80 {
            Saig_TsiPrintTraces( pTsi, pTsi.nWords, pTsi.nPrefix, pTsi.nCycle ); }
    }
    if fPrint != 0 {
        fmt.Printf("Print-out finished. Phase assignment is not performed.\n");
    } else if nFrames < 2 {
        fmt.Printf("The number of frames is less than 2. Phase assignment is not performed.\n");
    } else if nFrames > 256 {
        fmt.Printf("The number of frames is more than 256. Phase assignment is not performed.\n");
    } else if pTsi.nCycle == 1 {
        fmt.Printf("The cycle of ternary states is trivial. Phase abstraction cannot be done.\n");
    } else if pTsi.nCycle % nFrames != 0 {
        fmt.Printf("The cycle (%d) is not modulo the number of frames (%d). Phase abstraction cannot be done.\n", pTsi.nCycle, nFrames);
    } else if pTsi.nNonXRegs == 0 {
        fmt.Printf("All registers have X-valued states. Phase abstraction cannot be done.\n");
    } else if !(Saig_ManFindRegisters( pTsi, nFrames, fIgnore, fVerbose ) != 0) {
        fmt.Printf("There is no registers to abstract with %d frames.\n", nFrames);
    } else {
        pNew = Saig_ManPerformAbstraction( pTsi, nFrames, fVerbose ); }
    Saig_TsiStop( pTsi );
    return pNew;
}

/**Function*************************************************************

  Synopsis    [Performs automated phase abstraction.]

  Description [Takes the AIG manager and the array of initial states.]
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Saig_ManPhaseAbstractAuto(p *Aig_Man_t, fVerbose int32) *Aig_Man_t {
    var pNew *Aig_Man_t  = nil;
    var pTsi *Saig_Tsim_t;
    var fPrint int32;
    var nFrames int32;
    assert(Saig_ManRegNum(p), "Saig_ManRegNum(p)");
    assert(Saig_ManPiNum(p), "Saig_ManPiNum(p)");
    assert(Saig_ManPoNum(p), "Saig_ManPoNum(p)");
    // perform terminary simulation
    pTsi = Saig_ManReachableTernary( p, nil, fVerbose );
    if pTsi == nil {
        return nil; }
    // derive information
    pTsi.nPrefix = Saig_TsiComputePrefix( pTsi, (unsigned *)Vec_PtrEntryLast(pTsi.vStates), pTsi.nWords );
    pTsi.nCycle = Vec_PtrSize(pTsi.vStates) - 1 - pTsi.nPrefix;
    pTsi.nNonXRegs = Saig_TsiCountNonXValuedRegisters(pTsi, 0);
    // print statistics
    if fVerbose != 0 {
        fmt.Printf("Lead = %5d. Loop = %5d.  Total flops = %5d. Binary flops = %5d.\n", pTsi.nPrefix, pTsi.nCycle, p.nRegs, pTsi.nNonXRegs);
        if pTsi.nNonXRegs < 100 && Vec_PtrSize(pTsi.vStates) < 80 {
            Saig_TsiPrintTraces( pTsi, pTsi.nWords, pTsi.nPrefix, pTsi.nCycle ); }
    }
    nFrames = pTsi.nCycle;
    if fPrint != 0 {
        fmt.Printf("Print-out finished. Phase assignment is not performed.\n");
    
    } else if nFrames < 2 {
//        printf( "The number of frames is less than 2. Phase assignment is not performed.\n" );
    
    } else if nFrames > 256 {
//        printf( "The number of frames is more than 256. Phase assignment is not performed.\n" );
    
    } else if pTsi.nCycle == 1 {
//        printf( "The cycle of ternary states is trivial. Phase abstraction cannot be done.\n" );
    
    } else if pTsi.nCycle % nFrames != 0 {
//        printf( "The cycle (%d) is not modulo the number of frames (%d). Phase abstraction cannot be done.\n", pTsi->nCycle, nFrames );
    
    } else if pTsi.nNonXRegs == 0 {
//        printf( "All registers have X-valued states. Phase abstraction cannot be done.\n" );
    
    } else if !(Saig_ManFindRegisters( pTsi, nFrames, 0, fVerbose ) != 0) {
//        printf( "There is no registers to abstract with %d frames.\n", nFrames );
    
    } else {
        pNew = Saig_ManPerformAbstraction( pTsi, nFrames, fVerbose ); }
    Saig_TsiStop( pTsi );
    if pNew == nil {
        pNew = Aig_ManDupSimple( p ); }
    if Aig_ManCiNum(pNew) == Aig_ManRegNum(pNew) {
        Aig_ManStop( pNew);
        pNew = Aig_ManDupSimple( p );
    }
    return pNew;
}


/**Function*************************************************************

  Synopsis    [Derives CEX for the original AIG from CEX of the unrolled AIG.]

  Description [The number of PIs of the given CEX should divide by the number 
  of PIs of the original AIG without a remainder.]
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Saig_PhaseTranslateCex(p *Aig_Man_t, pCex *Abc_Cex_t) *Abc_Cex_t {
    var pNew *Abc_Cex_t;
    var i, k, iFrame, nFrames int32;
    // make sure the PI count of the AIG is a multiple of the PI count of the CEX
    // if this is not true, the CEX is not derived as the result of unrolling of pAig
    // or the unrolled CEX went through transformations that change the PI count
    if pCex.nPis % Saig_ManPiNum(p) != 0 {
        fmt.Printf("The PI count in the AIG and in the CEX do not match.\n");
        return nil;
    }
    // get the number of unrolled frames
    nFrames = pCex.nPis / Saig_ManPiNum(p);
    // get the frame where it fails
    iFrame = pCex.iFrame * nFrames + pCex.iPo / Saig_ManPoNum(p);
    // start a new CEX (assigns: p->nRegs, p->nPis, p->nBits)
    pNew = Abc_CexAlloc( Saig_ManRegNum(p), Saig_ManPiNum(p), iFrame+1 );
    assert(pNew.nBits == pNew.nPis * (iFrame + 1) + pNew.nRegs, "pNew.nBits == pNew.nPis * (iFrame + 1) + pNew.nRegs");
    // now assign the failed frame and the failed PO (p->iFrame and p->iPo)
    pNew.iFrame = iFrame;
    pNew.iPo    = pCex.iPo % Saig_ManPoNum(p);
    // copy the bit data
    for  i, k = pCex.nRegs, pNew.nRegs; k < pNew.nBits; k, i = k+1, i+1  {
        if Abc_InfoHasBit( pCex.pData, i ) {
            Abc_InfoSetBit( pNew.pData, k ); }
    }
    assert(i <= pCex.nBits, "i <= pCex.nBits");
    return pNew;
}

////////////////////////////////////////////////////////////////////////
///                       END OF FILE                                ///
////////////////////////////////////////////////////////////////////////


//ABC_NAMESPACE_IMPL_END

