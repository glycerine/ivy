package saig

import (
	"fmt"
	"time"
)

/**CFile****************************************************************

  FileName    [saigSimMv.c]

  SystemName  [ABC: Logic synthesis and verification system.]

  PackageName [Sequential AIG package.]

  Synopsis    [Multi-valued simulation.]

  Author      [Alan Mishchenko]
  
  Affiliation [UC Berkeley]

  Date        [Ver. 1.0. Started - June 20, 2005.]

  Revision    [$Id: saigSimMv.c,v 1.00 2005/06/20 00:00:00 alanmi Exp $]

***********************************************************************/

//#include "saig.h"

//ABC_NAMESPACE_IMPL_START


////////////////////////////////////////////////////////////////////////
///                        DECLARATIONS                              ///
////////////////////////////////////////////////////////////////////////

const SAIG_DIFF_VALUES =  8
const SAIG_UNDEF_VALUE =  0x1ffffffe  //536870910

// old AIG
   
type Saig_MvObj_t struct {
    iFan0              int32;
    iFan1              int32;
    Type uint32; /* C bitfield: 3 bits; review packing and truncation */
    Value uint32; /* C bitfield: 29 bits; review packing and truncation */
}

// new AIG
   
type Saig_MvAnd_t struct {
    iFan0              int32;
    iFan1              int32;
    iNext              int32;
}

// simulation manager
   
type Saig_MvMan_t struct {
    // user data
    pAig       *Aig_Man_t;         // original AIG    
    // parameters
    nStatesMax              int32;   // maximum number of states
    nLevelsMax              int32;   // maximum number of levels
    nValuesMax              int32;   // maximum number of values
    nFlops              int32;       // number of flops
    // compacted AIG
    pAigOld    *Saig_MvObj_t;      // AIG objects
    vFlops       *Vec_Ptr_t;       // collected flops
    vXFlops       *Vec_Int_t;      // flops that had at least one X-value
    vTired       *Vec_Ptr_t;       // collected flops
    pTStates        *uint32;     // hash table for states
    nTStatesSize              int32; // hash table size
    pMemStates   *Aig_MmFixed_t;   // memory for states
    vStates       *Vec_Ptr_t;      // reached states
    pRegsUndef             *int32;   // count the number of undef values
    pRegsValues            **int32;  // write the first different values
    nRegsValues             *int32;  // count the number of different values
    nRUndefs              int32;     // the number of undef registers
    nRValues [SAIG_DIFF_VALUES+1]int32; // the number of registers with given values
    // internal AIG
    pAigNew    *Saig_MvAnd_t;      // AIG nodes
    nObjsAlloc              int32;   // the number of objects allocated 
    nObjs              int32;        // the number of objects
    nPis              int32;         // the number of primary inputs
    pTNodes             *int32;      // hash table
    nTNodesSize              int32;  // hash table size
    unsigned char *  pLevels;      // levels of AIG nodes
}

func Saig_MvObjFaninC0(pObj *Saig_MvObj_t) int32 { return pObj.iFan0 & 1;           }
func Saig_MvObjFaninC1(pObj *Saig_MvObj_t) int32 { return pObj.iFan1 & 1;           }
func Saig_MvObjFanin0(pObj *Saig_MvObj_t) int32 { return pObj.iFan0 >> 1;          }
func Saig_MvObjFanin1(pObj *Saig_MvObj_t) int32 { return pObj.iFan1 >> 1;          }

func Saig_MvConst0() int32 { return 1;                         }
func Saig_MvConst1() int32 { return 0;                         }
func Saig_MvConst(c int32) int32 { return !c;                        }
func Saig_MvUndef() int32 { return SAIG_UNDEF_VALUE;          }

func Saig_MvIsConst0(iNode int32) int32 { return iNode == 1;                }
func Saig_MvIsConst1(iNode int32) int32 { return iNode == 0;                }
func Saig_MvIsConst(iNode int32) int32 { return iNode  < 2;                }
func Saig_MvIsUndef(iNode int32) int32 { return iNode == SAIG_UNDEF_VALUE; }

func Saig_MvRegular(iNode int32) int32 { return (iNode & ~01);             }
func Saig_MvNot(iNode int32) int32 { return (iNode ^  01);             }
func Saig_MvNotCond(iNode int32, c int32) int32 { return (iNode ^ (c));             }
func Saig_MvIsComplement(iNode int32) int32 { return int32((iNode & 01));         }

func Saig_MvLit2Var(iNode int32) int32 { return (iNode >> 1);              }
func Saig_MvVar2Lit(iVar int32) int32 { return (iVar << 1);               }
func Saig_MvLev(p *Saig_MvMan_t, iNode int32) int32 { return p.pLevels[iNode >> 1];    }

// iterator over compacted objects
// Original macro:
// #define Saig_MvManForEachObj( pAig, pEntry ) \
//     for ( pEntry = pAig; pEntry->Type != AIG_OBJ_VOID; pEntry++ )


////////////////////////////////////////////////////////////////////////
///                     FUNCTION DEFINITIONS                         ///
////////////////////////////////////////////////////////////////////////

/**Function*************************************************************

  Synopsis    [Creates reduced manager.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Saig_ManCreateReducedAig(p *Aig_Man_t, pvFlops **Vec_Ptr_t) *Saig_MvObj_t {
    var pAig, pEntry *Saig_MvObj_t;
    var pObj *Aig_Obj_t;
    var i int32;
    *pvFlops = Vec_PtrAlloc( Aig_ManRegNum(p) );
    // inlining of macro ABC_CALLOC
    pAig = ((Saig_MvObj_t *) calloc((size_t)(Aig_ManObjNumMax(p)+1), sizeof(Saig_MvObj_t)));
    // inlining of macro Aig_ManForEachObj
    // inlining of macro Vec_PtrForEachEntry
    for  i = 0; (i < Vec_PtrSize(p.vObjs)) && (((pObj) = (Aig_Obj_t *)Vec_PtrEntry(p.vObjs, i)), 1); i++  { if (pObj) == nil {} else {
        pEntry = pAig + i;
        pEntry.Type = pObj.Type;
        if Aig_ObjIsCi(pObj) || i == 0 {
            if Saig_ObjIsLo(p, pObj) {
                pEntry.iFan0 = (Saig_ObjLoToLi(p, pObj).Id << 1);
                pEntry.iFan1 = -1;
                Vec_PtrPush( *pvFlops, pEntry );
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

  Synopsis    [Creates a new node.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Saig_MvCreateObj(p *Saig_MvMan_t, iFan0 int32, iFan1 int32) int32 {
    var pNode *Saig_MvAnd_t;
    if p.nObjs == p.nObjsAlloc {
        // inlining of macro ABC_REALLOC
        
        if p.pAigNew {
            p.pAigNew = ((Saig_MvAnd_t *) realloc((char *)(p.pAigNew), sizeof(Saig_MvAnd_t) * (size_t)(2*p.nObjsAlloc)));
        } else {
            p.pAigNew = ((Saig_MvAnd_t *) malloc(sizeof(Saig_MvAnd_t) * (size_t)(2*p.nObjsAlloc)));
        }
        // inlining of macro ABC_REALLOC
        
        if p.pLevels {
            p.pLevels = ((unsigned char *) realloc((char *)(p.pLevels), sizeof(unsigned char) * (size_t)(2*p.nObjsAlloc)));
        } else {
            p.pLevels = ((unsigned char *) malloc(sizeof(unsigned char) * (size_t)(2*p.nObjsAlloc)));
        }
        p.nObjsAlloc *= 2;
    }
    pNode = p.pAigNew + p.nObjs;
    pNode.iFan0 = iFan0;
    pNode.iFan1 = iFan1;
    pNode.iNext = 0;
    if (iFan0 != 0) || (iFan1 != 0) {
        p.pLevels[p.nObjs] = 1 + Abc_MaxInt( Saig_MvLev(p, iFan0), Saig_MvLev(p, iFan1) );
    } else {
        p.pLevels[p.nObjs] = 0, p.nPis++; }
    return p.nObjs++;
}

/**Function*************************************************************

  Synopsis    [Creates multi-valued simulation manager.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Saig_MvManStart(pAig *Aig_Man_t, nFramesSatur int32) *Saig_MvMan_t {
    var p *Saig_MvMan_t;
    var i int32;
    assert(Aig_ManRegNum(pAig) > 0, "Aig_ManRegNum(pAig) > 0");
    // inlining of macro ABC_ALLOC
    p = (Saig_MvMan_t *)((Saig_MvMan_t *) malloc(sizeof(Saig_MvMan_t) * (size_t)(1)));
    memset( p, 0, sizeof(Saig_MvMan_t) );
    // set parameters
    p.pAig         = pAig;
    p.nStatesMax   = 2 * nFramesSatur + 100;
    p.nLevelsMax   = 4;
    p.nValuesMax   = SAIG_DIFF_VALUES;
    p.nFlops       = Aig_ManRegNum(pAig);
    // compacted AIG
    p.pAigOld      = Saig_ManCreateReducedAig( pAig, &p.vFlops );
    p.nTStatesSize = Abc_PrimeCudd( p.nStatesMax );
    // inlining of macro ABC_CALLOC
    p.pTStates     = ((unsigned *) calloc((size_t)(p.nTStatesSize), sizeof(unsigned)));
    p.pMemStates   = Aig_MmFixedStart( sizeof(int) * (p.nFlops+1), p.nStatesMax );
    p.vStates      = Vec_PtrAlloc( p.nStatesMax );
    Vec_PtrPush( p.vStates, nil );
    // inlining of macro ABC_CALLOC
    p.pRegsUndef   = ((int *) calloc((size_t)(p.nFlops), sizeof(int)));
    // inlining of macro ABC_ALLOC
    p.pRegsValues  = ((int * *) malloc(sizeof(int *) * (size_t)(p.nFlops)));
    // inlining of macro ABC_ALLOC
    p.pRegsValues[0] = ((int *) malloc(sizeof(int) * (size_t)(p.nValuesMax * p.nFlops)));
    for  i = 1; i < p.nFlops; i++  {
        p.pRegsValues[i] = p.pRegsValues[i-1] + p.nValuesMax;
    }
    // inlining of macro ABC_CALLOC
    p.nRegsValues  = ((int *) calloc((size_t)(p.nFlops), sizeof(int)));
    p.vTired       = Vec_PtrAlloc( 100 );
    // internal AIG
    p.nObjsAlloc   = 1000000;
    // inlining of macro ABC_ALLOC
    p.pAigNew      = ((Saig_MvAnd_t *) malloc(sizeof(Saig_MvAnd_t) * (size_t)(p.nObjsAlloc)));
    p.nTNodesSize  = Abc_PrimeCudd( p.nObjsAlloc / 3 );
    // inlining of macro ABC_CALLOC
    p.pTNodes      = ((int *) calloc((size_t)(p.nTNodesSize), sizeof(int)));
    // inlining of macro ABC_ALLOC
    p.pLevels      = ((unsigned char *) malloc(sizeof(unsigned char) * (size_t)(p.nObjsAlloc)));
    Saig_MvCreateObj( p, 0, 0 );
    return p;
}

/**Function*************************************************************

  Synopsis    [Destroys multi-valued simulation manager.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Saig_MvManStop(p *Saig_MvMan_t) {
    Aig_MmFixedStop( p.pMemStates, 0 );
    Vec_PtrFree( p.vStates );
    Vec_IntFreeP( &p.vXFlops );
    Vec_PtrFree( p.vFlops );
    Vec_PtrFree( p.vTired );
    // inlining of macro ABC_FREE
    ((p.pRegsValues[0]) ? (free((char *) (p.pRegsValues[0])), (p.pRegsValues[0]) = 0) : 0);
    // inlining of macro ABC_FREE
    ((p.pRegsValues) ? (free((char *) (p.pRegsValues)), (p.pRegsValues) = 0) : 0);
    // inlining of macro ABC_FREE
    ((p.nRegsValues) ? (free((char *) (p.nRegsValues)), (p.nRegsValues) = 0) : 0);
    // inlining of macro ABC_FREE
    ((p.pRegsUndef) ? (free((char *) (p.pRegsUndef)), (p.pRegsUndef) = 0) : 0);
    // inlining of macro ABC_FREE
    ((p.pAigOld) ? (free((char *) (p.pAigOld)), (p.pAigOld) = 0) : 0);
    // inlining of macro ABC_FREE
    ((p.pTStates) ? (free((char *) (p.pTStates)), (p.pTStates) = 0) : 0);
    // inlining of macro ABC_FREE
    ((p.pAigNew) ? (free((char *) (p.pAigNew)), (p.pAigNew) = 0) : 0);
    // inlining of macro ABC_FREE
    ((p.pTNodes) ? (free((char *) (p.pTNodes)), (p.pTNodes) = 0) : 0);
    // inlining of macro ABC_FREE
    ((p.pLevels) ? (free((char *) (p.pLevels)), (p.pLevels) = 0) : 0);
    // inlining of macro ABC_FREE
    ((p) ? (free((char *) (p)), (p) = 0) : 0);
}

/**Function*************************************************************

  Synopsis    [Hashing the node.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Saig_MvHash(iFan0 int32, iFan1 int32, TableSize int32) int32 {
    var Key uint32;
    assert(iFan0 < iFan1, "iFan0 < iFan1");
    Key ^= Saig_MvLit2Var(iFan0) * 7937;
    Key ^= Saig_MvLit2Var(iFan1) * 2971;
    Key ^= Saig_MvIsComplement(iFan0) * 911;
    Key ^= Saig_MvIsComplement(iFan1) * 353;
    return int32((Key % TableSize));
}

/**Function*************************************************************

  Synopsis    [Returns the place where this node is stored (or should be stored).]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Saig_MvTableFind(p *Saig_MvMan_t, iFan0 int32, iFan1 int32) *int32 {
    var pEntry *Saig_MvAnd_t;
    var pPlace *int32  = p.pTNodes + Saig_MvHash( iFan0, iFan1, p.nTNodesSize );
    for  pEntry = (*pPlace)? p.pAigNew + *pPlace : nil; pEntry; 
          pPlace = &pEntry.iNext, pEntry = (*pPlace)? p.pAigNew + *pPlace : nil  {
              if pEntry.iFan0 == iFan0 && pEntry.iFan1 == iFan1 {
                  break; }
    }
    return pPlace;
}

/**Function*************************************************************

  Synopsis    [Performs an AND-operation.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Saig_MvAnd(p *Saig_MvMan_t, iFan0 int32, iFan1 int32, fFirst int32) int32 {
    if iFan0 == iFan1 {
        return iFan0; }
    if iFan0 == Saig_MvNot(iFan1) {
        return Saig_MvConst0(); }
    if Saig_MvIsConst(iFan0) != 0 {
        if Saig_MvIsConst1(iFan0) != 0 {
            return iFan1;
        } else {
            return Saig_MvConst0();
        } }
    if Saig_MvIsConst(iFan1) != 0 {
        if Saig_MvIsConst1(iFan1) != 0 {
            return iFan0;
        } else {
            return Saig_MvConst0();
        } }
    if (Saig_MvIsUndef(iFan0) != 0) || (Saig_MvIsUndef(iFan1) != 0) {
        return Saig_MvUndef(); }
//    if ( Saig_MvLev(p, iFan0) >= p->nLevelsMax || Saig_MvLev(p, iFan1) >= p->nLevelsMax )
//        return Saig_MvUndef();

    // go undef after the first frame
    if !(fFirst != 0) {
        return Saig_MvUndef(); }

    if iFan0 > iFan1 {
        var Temp int32  = iFan0;
        iFan0 = iFan1;
        iFan1 = Temp;
    }
    {
        var pPlace *int32;
        pPlace = Saig_MvTableFind( p, iFan0, iFan1 );
        if *pPlace == 0 {
            if pPlace >= (int*)p.pAigNew && pPlace < (int*)(p.pAigNew + p.nObjsAlloc) {
                var iPlace int32  = pPlace - (int*)p.pAigNew;
                var iNode int32   = Saig_MvCreateObj( p, iFan0, iFan1 );
                ((int*)p.pAigNew)[iPlace] = iNode;
                return Saig_MvVar2Lit( iNode );
            
            } else {
                *pPlace = Saig_MvCreateObj( p, iFan0, iFan1 ); }
        }
        return Saig_MvVar2Lit( *pPlace );
    }
}

/**Function*************************************************************

  Synopsis    [Propagates one edge.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Saig_MvSimulateValue0(pAig *Saig_MvObj_t, pObj *Saig_MvObj_t) int32 {
    var pObj0 *Saig_MvObj_t  = pAig + Saig_MvObjFanin0(pObj);
    if Saig_MvIsUndef( pObj0.Value ) != 0 {
        return Saig_MvUndef(); }
    return Saig_MvNotCond( pObj0.Value, Saig_MvObjFaninC0(pObj) );
}
func Saig_MvSimulateValue1(pAig *Saig_MvObj_t, pObj *Saig_MvObj_t) int32 {
    var pObj1 *Saig_MvObj_t  = pAig + Saig_MvObjFanin1(pObj);
    if Saig_MvIsUndef( pObj1.Value ) != 0 {
        return Saig_MvUndef(); }
    return Saig_MvNotCond( pObj1.Value, Saig_MvObjFaninC1(pObj) );
}

/**Function*************************************************************

  Synopsis    [Prints MV state.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Saig_MvPrintState(Iter int32, p *Saig_MvMan_t) {
    var pEntry *Saig_MvObj_t;
    var i int32;
    fmt.Printf("%3d : ", Iter);
    // inlining of macro Vec_PtrForEachEntry
    for  i = 0; (i < Vec_PtrSize(p.vFlops)) && (((pEntry) = (Saig_MvObj_t *)Vec_PtrEntry(p.vFlops, i)), 1); i++  {
        if pEntry.Value == SAIG_UNDEF_VALUE {
            fmt.Printf("    *");
        } else {
            fmt.Printf("%5d", pEntry.Value); }
    }
    fmt.Printf("\n");
}

/**Function*************************************************************

  Synopsis    [Performs one iteration of simulation.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Saig_MvSimulateFrame(p *Saig_MvMan_t, fFirst int32, fVerbose int32) {
    var pEntry *Saig_MvObj_t;
    var i int32;
    // inlining of macro Saig_MvManForEachObj
    for  pEntry = p.pAigOld; pEntry.Type != AIG_OBJ_VOID; pEntry++  {
        if pEntry.Type == AIG_OBJ_AND {
            pEntry.Value = Saig_MvAnd( p, 
                Saig_MvSimulateValue0(p.pAigOld, pEntry),
                Saig_MvSimulateValue1(p.pAigOld, pEntry), fFirst );
        
        } else if pEntry.Type == AIG_OBJ_CO {
            pEntry.Value = Saig_MvSimulateValue0(p.pAigOld, pEntry);
        } else if pEntry.Type == AIG_OBJ_CI {
            if pEntry.iFan1 == 0 { // true PI
            
                if fFirst != 0 {
                    pEntry.Value = Saig_MvVar2Lit( Saig_MvCreateObj( p, 0, 0 ) );
                } else {
                    pEntry.Value = SAIG_UNDEF_VALUE; }
            }
//            else if ( fFirst ) // register output
//                pEntry->Value = Saig_MvConst0();
//            else
//                pEntry->Value = Saig_MvSimulateValue0(p->pAigOld, pEntry);
        
        } else if pEntry.Type == AIG_OBJ_CONST1 {
            pEntry.Value = Saig_MvConst1();
        } else if pEntry.Type != AIG_OBJ_NONE {
            assert(false, "0"); }
    }
    // transfer to registers
    // inlining of macro Vec_PtrForEachEntry
    for  i = 0; (i < Vec_PtrSize(p.vFlops)) && (((pEntry) = (Saig_MvObj_t *)Vec_PtrEntry(p.vFlops, i)), 1); i++  {
        pEntry.Value = Saig_MvSimulateValue0( p.pAigOld, pEntry );
    }
}


/**Function*************************************************************

  Synopsis    [Computes hash value of the node using its simulation info.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
var spStatic_Saig_MvSimHash_s_SPrimes [16]int32 = [16]int32{ 
     1610612741,
     805306457,
     402653189,
     201326611,
     100663319,
     50331653,
     25165843,
     12582917,
     6291469,
     3145739,
     1572869,
     786433,
     393241,
     196613,
     98317,
     49157,
    };
func Saig_MvSimHash(pState *uint32, nFlops int32, TableSize int32) int32 {
    
    var uHash uint32;
    var i int32;
    for  i = 0; i < nFlops; i++  {
        uHash ^= pState[i] * spStatic_Saig_MvSimHash_s_SPrimes[i & 0xF];
    }
    return int32((uHash % TableSize));
}

/**Function*************************************************************

  Synopsis    [Returns the place where this state is stored (or should be stored).]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Saig_MvSimTableFind(p *Saig_MvMan_t, pState *uint32) *uint32 {
    var pEntry *uint32;
    var pPlace *uint32  = p.pTStates + Saig_MvSimHash( pState+1, p.nFlops, p.nTStatesSize );
    for  pEntry = (*pPlace)? (unsigned *)Vec_PtrEntry(p.vStates, *pPlace) : nil; pEntry; 
          pPlace = pEntry, pEntry = (*pPlace)? (unsigned *)Vec_PtrEntry(p.vStates, *pPlace) : nil  {
              if memcmp( pEntry+1, pState+1, sizeof(int)*p.nFlops ) == 0 {
                  break; }
    }
    return pPlace;
}

/**Function*************************************************************

  Synopsis    [Saves current state.]

  Description [Returns -1 if there is no fixed point.]
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Saig_MvSaveState(p *Saig_MvMan_t) int32 {
    var pEntry *Saig_MvObj_t;
    var pState, pPlace *uint32;
    var i int32;
    pState = (unsigned *)Aig_MmFixedEntryFetch( p.pMemStates );
    pState[0] = 0;
    // inlining of macro Vec_PtrForEachEntry
    for  i = 0; (i < Vec_PtrSize(p.vFlops)) && (((pEntry) = (Saig_MvObj_t *)Vec_PtrEntry(p.vFlops, i)), 1); i++  {
        pState[i+1] = pEntry.Value;
    }
    pPlace = Saig_MvSimTableFind( p, pState );
    if *pPlace {
        return *pPlace; }
    *pPlace = Vec_PtrSize( p.vStates );
    Vec_PtrPush( p.vStates, pState );
    return -1;
}

/**Function*************************************************************

  Synopsis    [Performs multi-valued simulation.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Saig_MvManPostProcess(p *Saig_MvMan_t, iState int32) {
    var pEntry *Saig_MvObj_t;
    var pState *uint32;
    var i, k, j, nTotal, iFlop int32;
    var vUniques *Vec_Int_t  = Vec_IntAlloc( 100 );
    var vCounter *Vec_Int_t  = Vec_IntAlloc( 100 );
    // count registers that never became undef
    // inlining of macro Vec_PtrForEachEntry
    for  i = 0; (i < Vec_PtrSize(p.vFlops)) && (((pEntry) = (Saig_MvObj_t *)Vec_PtrEntry(p.vFlops, i)), 1); i++  {
        if p.pRegsUndef[i] == 0 {
            nTotal++; }
    }
    fmt.Printf("The number of registers that never became undef = %d. (Total = %d.)\n", nTotal, p.nFlops);
    // inlining of macro Vec_PtrForEachEntry
    for  i = 0; (i < Vec_PtrSize(p.vFlops)) && (((pEntry) = (Saig_MvObj_t *)Vec_PtrEntry(p.vFlops, i)), 1); i++  {
        if p.pRegsUndef[i] {
            continue; }
        // inlining of macro Vec_IntForEachEntry
        for  k = 0; (k < Vec_IntSize(vUniques)) && (((iFlop) = Vec_IntEntry(vUniques, k)), 1); k++  {
            // inlining of macro Vec_PtrForEachEntryStart
            for  j = 1; (j < Vec_PtrSize(p.vStates)) && (((pState) = (unsigned *)Vec_PtrEntry(p.vStates, j)), 1); j++  {
                if pState[iFlop+1] != pState[i+1] {
                    break; }
            }
            if j == Vec_PtrSize(p.vStates) {
                Vec_IntAddToEntry( vCounter, k, 1 );
                break;
            }
        }
        if k == Vec_IntSize(vUniques) {
            Vec_IntPush( vUniques, i );
            Vec_IntPush( vCounter, 1 );
        }
    }
    // inlining of macro Vec_IntForEachEntry
    for  i = 0; (i < Vec_IntSize(vUniques)) && (((iFlop) = Vec_IntEntry(vUniques, i)), 1); i++  {
        fmt.Printf("FLOP %5d : (%3d) ", iFlop, Vec_IntEntry(vCounter,i));
/*
        for ( k = 0; k < p->nRegsValues[iFlop]; k++ )
            if ( p->pRegsValues[iFlop][k] == SAIG_UNDEF_VALUE )
                printf( "* " );
            else
                printf( "%d ", p->pRegsValues[iFlop][k] );
        printf( "\n" );
*/
        // inlining of macro Vec_PtrForEachEntryStart
        for  k = 1; (k < Vec_PtrSize(p.vStates)) && (((pState) = (unsigned *)Vec_PtrEntry(p.vStates, k)), 1); k++  {
            if k == iState+1 {
                fmt.Printf(" # "); }
            if pState[iFlop+1] == SAIG_UNDEF_VALUE {
                fmt.Printf("*");
            } else {
                fmt.Printf("%d", pState[iFlop+1]); }
        }
        fmt.Printf("\n");
//        if ( ++Counter == 10 )
//            break;
    }

    Vec_IntFree( vUniques );
    Vec_IntFree( vCounter );
}

/**Function*************************************************************

  Synopsis    [Performs multi-valued simulation.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Saig_MvManFindXFlops(p *Saig_MvMan_t) *Vec_Int_t {
    var vXFlops *Vec_Int_t;
    var pState *uint32;
    var i, k int32;
    vXFlops = Vec_IntStart( p.nFlops );
    // inlining of macro Vec_PtrForEachEntryStart
    for  i = 1; (i < Vec_PtrSize(p.vStates)) && (((pState) = (unsigned *)Vec_PtrEntry(p.vStates, i)), 1); i++  {
        for  k = 0; k < p.nFlops; k++  {
            if Saig_MvIsUndef(pState[k+1]) != 0 {
                Vec_IntWriteEntry( vXFlops, k, 1 ); }
        }
    }
    return vXFlops;
}

/**Function*************************************************************

  Synopsis    [Checks if the flop is an oscilator.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Saig_MvManCheckOscilator(p *Saig_MvMan_t, iFlop int32) int32 {
    var vValues *Vec_Int_t;
    var pState *uint32;
    var k, Per, Entry int32;
    // collect values of this flop
    vValues = Vec_IntAlloc( 100 );
    // inlining of macro Vec_PtrForEachEntryStart
    for  k = 1; (k < Vec_PtrSize(p.vStates)) && (((pState) = (unsigned *)Vec_PtrEntry(p.vStates, k)), 1); k++  {
        Vec_IntPush( vValues, pState[iFlop+1] );
//printf( "%d ", pState[iFlop+1] );
    }
//printf( "\n" );
    assert((Saig_MvIsConst0( Vec_IntEntry(vValues,0) ) != 0), "Saig_MvIsConst0( Vec_IntEntry(vValues,0) )");
    // look for constants
    for  Per = 0; Per < Vec_IntSize(vValues)/2; Per++  {
        // find the first non-const0
        // inlining of macro Vec_IntForEachEntryStart
        for  Per = Per; (Per < Vec_IntSize(vValues)) && (((Entry) = Vec_IntEntry(vValues, Per)), 1); Per++  {
            if !(Saig_MvIsConst0(Entry) != 0) {
                break; }
        }
        if Per == Vec_IntSize(vValues) {
            break; }
        // find the first const0
        // inlining of macro Vec_IntForEachEntryStart
        for  Per = Per; (Per < Vec_IntSize(vValues)) && (((Entry) = Vec_IntEntry(vValues, Per)), 1); Per++  {
            if Saig_MvIsConst0(Entry) != 0 {
                break; }
        }
        if Per == Vec_IntSize(vValues) {
            break; }
        // try to determine period
        assert((Saig_MvIsConst0( Vec_IntEntry(vValues,Per) ) != 0), "Saig_MvIsConst0( Vec_IntEntry(vValues,Per) )");
        for  k = Per; k < Vec_IntSize(vValues); k++  {
            if Vec_IntEntry(vValues, k-Per) != Vec_IntEntry(vValues, k) {
                break; }
        }
        if k < Vec_IntSize(vValues) {
            continue; }
        Vec_IntFree( vValues );
//printf( "Period = %d\n", Per );
        return Per;
    }
    Vec_IntFree( vValues );
    return 0;
}

/**Function*************************************************************

  Synopsis    [Returns const0 and binary flops.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Saig_MvManFindConstBinaryFlops(p *Saig_MvMan_t, pvBinary **Vec_Int_t) *Vec_Int_t {
    var pState *uint32;
    var vBinary, vConst0 *Vec_Int_t;
    var i, k, fConst0 int32;
    // detect constant flops
    vConst0 = Vec_IntAlloc( p.nFlops );
    vBinary = Vec_IntAlloc( p.nFlops );
    for  k = 0; k < p.nFlops; k++  {
        // check if this flop is constant 0 in all states
        fConst0 = 1;
        // inlining of macro Vec_PtrForEachEntryStart
        for  i = 1; (i < Vec_PtrSize(p.vStates)) && (((pState) = (unsigned *)Vec_PtrEntry(p.vStates, i)), 1); i++  {
            if !(Saig_MvIsConst0(pState[k+1]) != 0) {
                fConst0 = 0; }
            if Saig_MvIsUndef(pState[k+1]) != 0 {
                break; }
        }
        if i < Vec_PtrSize(p.vStates) {
            continue; }
        // the flop is binary-valued        
        if fConst0 != 0 { 
            Vec_IntPush( vConst0, k );
        } else {
            Vec_IntPush( vBinary, k ); }
    }
    *pvBinary = vBinary;
    return vConst0;
}

/**Function*************************************************************

  Synopsis    [Find oscilators.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Saig_MvManFindOscilators(p *Saig_MvMan_t, pvConst0 **Vec_Int_t) *Vec_Int_t {
    var vBinary, vOscils *Vec_Int_t;
    var Entry, i int32;
    // detect constant flops
    *pvConst0 = Saig_MvManFindConstBinaryFlops( p, &vBinary );
    // check binary flops
    vOscils = Vec_IntAlloc( 100 );
    // inlining of macro Vec_IntForEachEntry
    for  i = 0; (i < Vec_IntSize(vBinary)) && (((Entry) = Vec_IntEntry(vBinary, i)), 1); i++  {
        if Saig_MvManCheckOscilator( p, Entry ) != 0 {
            Vec_IntPush( vOscils, Entry ); }
    }
    Vec_IntFree( vBinary );
    return vOscils;
}

/**Function*************************************************************

  Synopsis    [Find constants and oscilators.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Saig_MvManCreateNextSkip(p *Saig_MvMan_t) *Vec_Int_t {
    var vConst0, vOscils, vXFlops *Vec_Int_t;
    var i, Entry int32;
    vOscils = Saig_MvManFindOscilators( p, &vConst0 );
//printf( "Found %d constants and %d oscilators.\n", Vec_IntSize(vConst0), Vec_IntSize(vOscils) );
    vXFlops = Vec_IntAlloc( p.nFlops );
    Vec_IntFill( vXFlops, p.nFlops, 1 );
    // inlining of macro Vec_IntForEachEntry
    for  i = 0; (i < Vec_IntSize(vConst0)) && (((Entry) = Vec_IntEntry(vConst0, i)), 1); i++  {
        Vec_IntWriteEntry( vXFlops, Entry, 0 );
    }
    // inlining of macro Vec_IntForEachEntry
    for  i = 0; (i < Vec_IntSize(vOscils)) && (((Entry) = Vec_IntEntry(vOscils, i)), 1); i++  {
        Vec_IntWriteEntry( vXFlops, Entry, 0 );
    }
    Vec_IntFree( vOscils );
    Vec_IntFree( vConst0 );
    return vXFlops;
}

/**Function*************************************************************

  Synopsis    [Finds equivalent flops.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Saig_MvManDeriveMap(p *Saig_MvMan_t, fVerbose int32) *Vec_Ptr_t {
    var vConst0, vBinValued *Vec_Int_t;
    var vMap *Vec_Ptr_t  = nil;
    var pObj *Aig_Obj_t;
    var pState *uint32;
    var i, k, j, FlopK, FlopJ int32;
    var Counter1, Counter2 int32;
    // prepare CI map
    vMap = Vec_PtrAlloc( Aig_ManCiNum(p.pAig) );
    // inlining of macro Aig_ManForEachCi
    // inlining of macro Vec_PtrForEachEntry
    for  i = 0; (i < Vec_PtrSize(p.pAig.vCis)) && (((pObj) = (Aig_Obj_t *)Vec_PtrEntry(p.pAig.vCis, i)), 1); i++  {
        Vec_PtrPush( vMap, pObj );
    }
    // detect constant flops
    vConst0 = Saig_MvManFindConstBinaryFlops( p, &vBinValued );
    // set constants
    // inlining of macro Vec_IntForEachEntry
    for  k = 0; (k < Vec_IntSize(vConst0)) && (((FlopK) = Vec_IntEntry(vConst0, k)), 1); k++  {
        Vec_PtrWriteEntry( vMap, Saig_ManPiNum(p.pAig) + FlopK, Aig_ManConst0(p.pAig) );
        Counter1++;
    }
    Vec_IntFree( vConst0 ); 

    // detect equivalent (non-ternary flops)
    // inlining of macro Vec_IntForEachEntry
    for  k = 0; (k < Vec_IntSize(vBinValued)) && (((FlopK) = Vec_IntEntry(vBinValued, k)), 1); k++  {
    if FlopK >= 0 {
    // inlining of macro Vec_IntForEachEntryStart
    for  j = k+1; (j < Vec_IntSize(vBinValued)) && (((FlopJ) = Vec_IntEntry(vBinValued, j)), 1); j++  {
    if FlopJ >= 0 {
        // check if they are equal
        // inlining of macro Vec_PtrForEachEntryStart
        for  i = 1; (i < Vec_PtrSize(p.vStates)) && (((pState) = (unsigned *)Vec_PtrEntry(p.vStates, i)), 1); i++  {
            if pState[FlopK+1] != pState[FlopJ+1] {
                break; }
        }
        if i < Vec_PtrSize(p.vStates) {
            continue; }
        // set the equivalence
        Vec_PtrWriteEntry( vMap, Saig_ManPiNum(p.pAig) + FlopJ, Saig_ManLo(p.pAig, FlopK) );
        Vec_IntWriteEntry( vBinValued, j, -1 );
        Counter2++;
    }
    } }
    }
    if fVerbose != 0 {
    fmt.Printf("Detected %d const0 flops and %d pairs of equiv binary flops.\n", Counter1, Counter2); }
    Vec_IntFree( vBinValued );
    if Counter1 == 0 && Counter2 == 0 {
        Vec_PtrFreeP( &vMap ); }
    return vMap;
}

/**Function*************************************************************

  Synopsis    [Performs multi-valued simulation.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Saig_MvManSimulate(pAig *Aig_Man_t, nFramesSymb int32, nFramesSatur int32, fVerbose int32, fVeryVerbose int32) *Vec_Ptr_t {
    var vMap *Vec_Ptr_t;
    var p *Saig_MvMan_t;
    var pEntry *Saig_MvObj_t;
    var f, i, iState int32;
    var clk abctime  = Abc_Clock();
    assert(nFramesSymb >= 1 && nFramesSymb <= nFramesSatur, "nFramesSymb >= 1 && nFramesSymb <= nFramesSatur");

    // start manager
    p = Saig_MvManStart( pAig, nFramesSatur );
if fVerbose != 0 {
// inlining of macro ABC_PRT
Abc_Print(1, "%s =", ("Constructing the problem"));
Abc_Print(1, "%9.2f sec\n", 1.0*(float64((Abc_Clock() - clk)))/(float64(int64(time.Second))));; }

    // initialize registers
    // inlining of macro Vec_PtrForEachEntry
    for  i = 0; (i < Vec_PtrSize(p.vFlops)) && (((pEntry) = (Saig_MvObj_t *)Vec_PtrEntry(p.vFlops, i)), 1); i++  {
        pEntry.Value = Saig_MvConst0();
    }
    Saig_MvSaveState( p );
    if fVeryVerbose != 0 {
        Saig_MvPrintState( 0, p ); }
    // simulate until convergence
    clk = Abc_Clock();
    for  f = 0; ; f++  { 
        if f == nFramesSatur {
            if fVerbose != 0 {
            fmt.Printf("Begining to saturate simulation after %d frames\n", f); }
            // find all flops that have at least one X value in the past and set them to X forever
            p.vXFlops = Saig_MvManFindXFlops( p );            
        }
        if f == 2 * nFramesSatur {
            if fVerbose != 0 {
            fmt.Printf("Agressively saturating simulation after %d frames\n", f); }
            Vec_IntFree( p.vXFlops );
            p.vXFlops = Saig_MvManCreateNextSkip( p );
        }
        // retire some flops
        if p.vXFlops {
            // inlining of macro Vec_PtrForEachEntry
            for  i = 0; (i < Vec_PtrSize(p.vFlops)) && (((pEntry) = (Saig_MvObj_t *)Vec_PtrEntry(p.vFlops, i)), 1); i++  {
                if Vec_IntEntry( p.vXFlops, i ) {
                    pEntry.Value = SAIG_UNDEF_VALUE; }
            }
        }
        // simulate timeframe
        Saig_MvSimulateFrame( p, int32((f < nFramesSymb)), fVerbose );
        // save and print state
        iState = Saig_MvSaveState( p );
        if fVeryVerbose != 0 {
            Saig_MvPrintState( f+1, p ); }
        if iState >= 0 {
            if fVerbose != 0 {
            fmt.Printf("Converged after %d frames with lasso in state %d. Cycle = %d.\n", f+1, iState-1, f+2-iState); }
//            printf( "Total number of PIs = %d. AND nodes = %d.\n", p->nPis, p->nObjs - p->nPis );
            break;
        }
    }
//    printf( "Coverged after %d frames.\n", f );
if fVerbose != 0 {
// inlining of macro ABC_PRT
Abc_Print(1, "%s =", ("Multi-valued simulation"));
Abc_Print(1, "%9.2f sec\n", 1.0*(float64((Abc_Clock() - clk)))/(float64(int64(time.Second))));; }
    // implement equivalences
//    Saig_MvManPostProcess( p, iState-1 );
    vMap = Saig_MvManDeriveMap( p, fVerbose );
    Saig_MvManStop( p );
//    return Aig_ManDupSimple( pAig );
    return vMap;
}


////////////////////////////////////////////////////////////////////////
///                       END OF FILE                                ///
////////////////////////////////////////////////////////////////////////


//ABC_NAMESPACE_IMPL_END

