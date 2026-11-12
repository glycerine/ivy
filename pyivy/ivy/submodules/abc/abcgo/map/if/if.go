package if_

/**CFile****************************************************************

  FileName    [if.h]

  SystemName  [ABC: Logic synthesis and verification system.]

  PackageName [FPGA mapping based on priority cuts.]

  Synopsis    [External declarations.]

  Author      [Alan Mishchenko]
  
  Affiliation [UC Berkeley]

  Date        [Ver. 1.0. Started - November 21, 2006.]

  Revision    [$Id: if.h,v 1.00 2006/11/21 00:00:00 alanmi Exp $]

***********************************************************************/
 




////////////////////////////////////////////////////////////////////////
///                          INCLUDES                                ///
////////////////////////////////////////////////////////////////////////
 
//#include <stdio.h>
//#include <stdlib.h>
//#include <string.h>
//#include <assert.h>

//#include "misc/vec/vec.h"
//#include "misc/mem/mem.h"
//#include "misc/tim/tim.h"
//#include "misc/util/utilNam.h"
//#include "misc/vec/vecMem.h"
//#include "misc/util/utilTruth.h"
//#include "opt/dau/dau.h"
//#include "misc/vec/vecHash.h"
//#include "misc/vec/vecWec.h"

//ABC_NAMESPACE_HEADER_START


////////////////////////////////////////////////////////////////////////
///                         PARAMETERS                               ///
////////////////////////////////////////////////////////////////////////
 
// the maximum size of LUTs used for mapping (should be the same as FPGA_MAX_LUTSIZE defined in "fpga.h"!!!)
const IF_MAX_LUTSIZE =       32
// the largest possible number of LUT inputs when funtionality of the LUTs are computed
const IF_MAX_FUNC_LUTSIZE =  15
// a very large number
const IF_INFINITY =          100000000  
// the largest possible user cut cost
const IF_COST_MAX =          4095 // ((1<<12)-1)

#define IF_BIG_CHAR ((char)120)

// object types
 type If_Type_t int32
const ( 
    IF_NONE If_Type_t = iota
     // 0: non-existent object
    IF_CONST1
   // 1: constant 1 
    IF_CI
       // 2: combinational input
    IF_CO
       // 3: combinational output
    IF_AND
      // 4: AND node
    IF_VOID      // 5: unused object

)

 

////////////////////////////////////////////////////////////////////////
///                         BASIC TYPES                              ///
////////////////////////////////////////////////////////////////////////

type If_Man_t =     If_Man_t_;
type If_Par_t =     If_Par_t_;
type If_Obj_t =     If_Obj_t_;
type If_Cut_t =     If_Cut_t_;
type If_Set_t =     If_Set_t_;
type If_LibLut_t =  If_LibLut_t_;
type If_LibBox_t =  If_LibBox_t_;
type If_DsdMan_t =  If_DsdMan_t_;
type Ifn_Ntk_t =    Ifn_Ntk_t_;

     
type Ifif_Par_t struct {
    nLutSize                int32;      // the LUT size
    pLutLib       *If_LibLut_t;       // the LUT library
    pLutDelays [IF_MAX_LUTSIZE]float32;  // pin-to-pin delays of the max LUT
    DelayWire              float32;     // wire delay
    nDegree                int32;       // structure degree 
    fCascade                int32;      // cascade
    fVerbose                int32;      // verbose
    fVeryVerbose                int32;  // verbose
}

// parameters
type If_Par_t_ struct {
    // user-controlable parameters
    nLutSize                int32;      // the LUT size
    nCutsMax                int32;      // the max number of cuts
    nFlowIters                int32;    // the number of iterations of area recovery
    nAreaIters                int32;    // the number of iterations of area recovery
    nGateSize                int32;     // the max size of the AND/OR gate to map into
    nNonDecLimit                int32;  // the max size of non-dec nodes
    DelayTarget              float32;   // delay target
    Epsilon              float32;       // value used in comparison floating point numbers
    nRelaxRatio                int32;   // delay relaxation ratio
    nStructType                int32;   // type of the structure
    nAndDelay                int32;     // delay of AND-gate in LUT library units
    nAndArea                int32;      // area of AND-gate in LUT library units
    fPreprocess                int32;   // preprossing
    fArea                int32;         // area-oriented mapping
    fFancy                int32;        // a fancy feature
    fExpRed                int32;       // expand/reduce of the best cuts
    fLatchPaths                int32;   // reset timing on latch paths
    fLut6Filter                int32;   // uses filtering of 6-LUT functions
    fEdge                int32;         // uses edge-based cut selection heuristics
    fPower                int32;        // uses power-aware cut selection heuristics
    fCutMin                int32;       // performs cut minimization by removing functionally reducdant variables
    fDelayOpt                int32;     // special delay optimization
    fDelayOptLut                int32;  // delay optimization for LUTs
    fDsdBalance                int32;   // special delay optimization
    fUserRecLib                int32;   // use recorded library
    fUserSesLib                int32;   // use SAT-based synthesis
    fBidec                int32;        // use bi-decomposition
    fUse34Spec                int32;    // use specialized matching
    fUseBat                int32;       // use one specialized feature
    fUseBuffs                int32;     // use buffers to decouple outputs
    fEnableCheck07                int32;// enable additional checking
    fEnableCheck08                int32;// enable additional checking
    fEnableCheck10                int32;// enable additional checking
    fEnableCheck75                int32;// enable additional checking
    fEnableCheck75u                int32;// enable additional checking
    fUseDsd                int32;       // compute DSD of the cut functions
    fUseDsdTune                int32;   // use matching based on precomputed manager
    fUseCofVars                int32;   // use cofactoring variables
    fUseAndVars                int32;   // use bi-decomposition
    fUseTtPerm                int32;    // compute truth tables of the cut functions
    fDeriveLuts                int32;   // enables deriving LUT structures
    fDoAverage                int32;    // optimize average rather than maximum level
    fHashMapping                int32;  // perform AIG hashing after mapping
    fVerbose                int32;      // the verbosity flag
    fVerboseTrace                int32; // the verbosity flag
    pLutStruct              string;    // LUT structure
    WireDelay              float32;     // wire delay
    // internal parameters
    fSkipCutFilter                int32;// skip cut filter
    fAreaOnly                int32;     // area only mode
    fTruth                int32;        // truth table computation enabled
    fUsePerm                int32;      // use permutation (delay info)
    fUseBdds                int32;      // use local BDDs as a cost function
    fUseSops                int32;      // use local SOPs as a cost function
    fUseCnfs                int32;      // use local CNFs as a cost function
    fUseMv                int32;        // use local MV-SOPs as a cost function
    nLatchesCi                int32;    // the number of latches among the CIs
    nLatchesCo                int32;    // the number of latches among the COs
    nLatchesCiBox                int32; // the number of white box outputs among the CIs
    nLatchesCoBox                int32; // the number of white box inputs among the COs
    fLiftLeaves                int32;   // shift the leaves for seq mapping
    fUseCoAttrs                int32;   // use CO attributes
    DelayTargetNew              float32;// new delay target
    FinalDelay              float32;    // final delay after mapping
    FinalArea              float32;     // final area after mapping
    pLutLib       *If_LibLut_t;       // the LUT library
    pTimesArr             *float32;     // arrival times
    pTimesReq             *float32;     // required times
    int (* pFuncCost)  (If_Man_t *, If_Cut_t *);  // procedure to compute the user's cost of a cut
    int (* pFuncUser)  (If_Man_t *, If_Obj_t *, If_Cut_t *);            //  procedure called for each cut when cut computation is finished
    int (* pFuncCell)  (If_Man_t *, unsigned *, int, int, char *);      //  procedure called for cut functions
    int (* pFuncCell2) (If_Man_t *, word *, int, Vec_Str_t *, char **); //  procedure called for cut functions
    int (* pFuncWrite) (If_Man_t *);                                    //  procedure called for cut functions
    pReoMan              any;       // reordering manager
}

// the LUT library
type If_LibLut_t_ struct {
    pName              string;         // the name of the LUT library
    LutMax                int32;        // the maximum LUT size 
    fVarPinDelays                int32; // set to 1 if variable pin delays are specified
    pLutAreas [IF_MAX_LUTSIZE+1]float32; // the areas of LUTs
    pLutDelays [IF_MAX_LUTSIZE+1][IF_MAX_LUTSIZE+1]float32;// the delays of LUTs
}

// manager
type If_Man_t_ struct {
    pName              string;
    // mapping parameters
    pPars          *If_Par_t;
    // mapping nodes
    pConst1          *If_Obj_t;       // the constant 1 node
    vCis         *Vec_Ptr_t;          // the primary inputs
    vCos         *Vec_Ptr_t;          // the primary outputs
    vObjs         *Vec_Ptr_t;         // all objects
    vObjsRev         *Vec_Ptr_t;      // reverse topological order of objects
    vTemp         *Vec_Ptr_t;         // temporary array
    nObjs [IF_VOID]int32;// the number of objects by type
    // various data
    nLevelMax                int32;     // the max number of AIG levels
    fEpsilon              float32;      // epsilon used for comparison
    RequiredGlo              float32;   // global required times
    RequiredGlo2              float32;  // global required times
    AreaGlo              float32;       // global area
    nNets                int32;         // the sum total of fanins of all LUTs in the mapping
    dPower              float32;        // the sum total of switching activities of all LUTs in the mapping
    nCutsUsed                int32;     // the number of cuts currently used
    nCutsMerged                int32;   // the total number of cuts merged
    puTemp [4]*uint32;     // used for the truth table computation
    puTempW              *word;       // used for the truth table computation
    SortMode                int32;      // one of the three sorting modes
    fNextRound                int32;    // set to 1 after the first round
    nChoices                int32;      // the number of choice nodes
    vSwitching         *Vec_Int_t;    // switching activity of each node
    pPerm [3][IF_MAX_LUTSIZE]int32; // permutations
    uSharedMask           uint32;   // mask of shared variables
    nShared                int32;       // the number of shared variables
    fReqTimeWarn                int32;  // warning about exceeding required times was printed
    // SOP balancing
    vCover         *Vec_Int_t;        // used to compute ISOP
    vArray         *Vec_Int_t;        // intermediate storage
    vAnds         *Vec_Wrd_t;         // intermediate storage
    vOrGate         *Vec_Wrd_t;       // intermediate storage
    vAndGate         *Vec_Wrd_t;      // intermediate storage
    // sequential mapping
    vLatchOrder         *Vec_Ptr_t;   // topological ordering of latches
    vLags         *Vec_Int_t;         // sequentail lags of all nodes
    nAttempts                int32;     // the number of attempts in binary search
    nMaxIters                int32;     // the maximum number of iterations
    Period                int32;        // the current value of the clock period (for seq mapping)
    // memory management
    nTruth6Words [IF_MAX_FUNC_LUTSIZE+1]int32;  // the size of the truth table if allocated
    nPermWords                int32;    // the size of the permutation array (in words)
    nObjBytes                int32;     // the size of the object
    nCutBytes                int32;     // the size of the cut
    nSetBytes                int32;     // the size of the cut set
    pMemObj       *Mem_Fixed_t;       // memory manager for objects (entrysize = nEntrySize)
    pMemSet       *Mem_Fixed_t;       // memory manager for sets of cuts (entrysize = nCutSize*(nCutsMax+1))
    pMemCi          *If_Set_t;        // memory for CI cutsets
    pMemAnd          *If_Set_t;       // memory for AND cutsets
    pFreeList          *If_Set_t;     // the list of free cutsets
    nSmallSupp                int32;    // the small support
    nCutsTotal                int32;
    nCutsUseless [32]int32;
    nCutsCount [32]int32;
    nCutsCountAll                int32;
    nCutsUselessAll                int32;
    nCuts5, nCuts5a int32;
    pIfDsdMan       *If_DsdMan_t;     // DSD manager
    Vec_Mem_t *        vTtMem[IF_MAX_FUNC_LUTSIZE+1];   // truth table memory and hash table
    Vec_Wec_t *        vTtIsops[IF_MAX_FUNC_LUTSIZE+1]; // mapping of truth table into DSD
    Vec_Int_t *        vTtDsds[IF_MAX_FUNC_LUTSIZE+1];  // mapping of truth table into DSD
    Vec_Str_t *        vTtPerms[IF_MAX_FUNC_LUTSIZE+1]; // mapping of truth table into permutations
    Vec_Str_t *        vTtVars[IF_MAX_FUNC_LUTSIZE+1];  // mapping of truth table into selected vars
    Vec_Int_t *        vTtDecs[IF_MAX_FUNC_LUTSIZE+1];  // mapping of truth table into decomposition pattern
    Vec_Int_t *        vTtOccurs[IF_MAX_FUNC_LUTSIZE+1];// truth table occurange counters
    vPairHash     *Hash_IntMan_t;     // hashing pairs of truth tables
    vPairRes         *Vec_Int_t;      // resulting truth table
    vPairPerms         *Vec_Str_t;    // resulting permutation
    vTtMem6         *Vec_Mem_t;
    pCanonPerm [IF_MAX_LUTSIZE]byte;
    uCanonPhase           uint32;
    nCacheHits                int32;
    nCacheMisses                int32;
    abctime            timeCache[6];
    nBestCutSmall [2]int32;
    nCountNonDec [2]int32;
    vCutData         *Vec_Int_t;      // cut data storage
    pArrTimeProfile [IF_MAX_FUNC_LUTSIZE]int32;
    vVisited         *Vec_Ptr_t;
    pUserMan              any;
    vDump         *Vec_Int_t;
    pDumpIns [16]int32;

    // timing manager
    pManTim         *Tim_Man_t;
    vCoAttrs         *Vec_Int_t;      // CO attributes   0=optimize; 1=keep; 2=relax
    // hash table for functions
    nTableSize [2]int32;    // hash table size
    nTableEntries [2]int32; // hash table entries
    void **            pHashTable[2];    // hash table bins
    pMemEntries       *Mem_Fixed_t;      // memory manager for hash table entries
    // statistics 
//    abctime                timeTruth;
}

// priority cut
type If_Cut_t_ struct {
    Area              float32;          // area (or area-flow) of the cut
    Edge              float32;          // the edge flow
    Power              float32;         // the power flow
    Delay              float32;         // delay of the cut
    iCutFunc                int32;      // TT ID of the cut
    uMaskFunc                int32;     // polarity bitmask
    uSign           uint32;         // cut signature
    Cost uint32; /* C bitfield: 12 bits; review packing and truncation */  // the user's cost of the cut (related to IF_COST_MAX)
    fCompl uint32; /* C bitfield: 1 bits; review packing and truncation */  // the complemented attribute 
    fUser uint32; /* C bitfield: 1 bits; review packing and truncation */  // using the user's area and delay
    fUseless uint32; /* C bitfield: 1 bits; review packing and truncation */  // cannot be used in the mapping
    fAndCut uint32; /* C bitfield: 1 bits; review packing and truncation */  // matched with AND gate
    nLimit uint32; /* C bitfield: 8 bits; review packing and truncation */  // the maximum number of leaves
    nLeaves uint32; /* C bitfield: 8 bits; review packing and truncation */  // the number of leaves
    pLeaves [0]int32;
}

// set of priority cut
type If_Set_t_ struct {
    nCutsMax              int16;      // the max number of cuts
    nCuts              int16;         // the current number of cuts
    pNext          *If_Set_t;         // next cutset in the free list
    ppCuts         **If_Cut_t;        // the array of pointers to the cuts
}

// node extension
type If_Obj_t_ struct {
    Type uint32; /* C bitfield: 4 bits; review packing and truncation */  // object
    fCompl0 uint32; /* C bitfield: 1 bits; review packing and truncation */  // complemented attribute
    fCompl1 uint32; /* C bitfield: 1 bits; review packing and truncation */  // complemented attribute
    fPhase uint32; /* C bitfield: 1 bits; review packing and truncation */  // phase of the node
    fRepr uint32; /* C bitfield: 1 bits; review packing and truncation */  // representative of the equivalence class
    fMark uint32; /* C bitfield: 1 bits; review packing and truncation */  // multipurpose mark
    fVisit uint32; /* C bitfield: 1 bits; review packing and truncation */  // multipurpose mark
    fSpec uint32; /* C bitfield: 1 bits; review packing and truncation */  // multipurpose mark
    fDriver uint32; /* C bitfield: 1 bits; review packing and truncation */  // multipurpose mark
    fSkipCut uint32; /* C bitfield: 1 bits; review packing and truncation */  // multipurpose mark
    Level uint32; /* C bitfield: 19 bits; review packing and truncation */  // logic level of the node
    Id                int32;            // integer ID
    IdPio                int32;         // integer ID of PIs/POs
    nRefs                int32;         // the number of references
    nVisits                int32;       // the number of visits to this node
    nVisitsCopy                int32;   // the number of visits to this node
    pFanin0          *If_Obj_t;       // the first fanin 
    pFanin1          *If_Obj_t;       // the second fanin
    pEquiv          *If_Obj_t;        // the choice node
    EstRefs              float32;       // estimated reference counter
    Required              float32;      // required time of the onde
    LValue              float32;        // sequential arrival time of the node
    union{
    void *             pCopy;         // used for object duplication
    int                iCopy;
    };
    
    pCutSet          *If_Set_t;       // the pointer to the cutset
    CutBest           If_Cut_t;       // the best cut selected 
}

   
type If_Box_t struct {
    pName              string;
    fSeq               byte;
    fBlack               byte;
    fOuter               byte;
    fUnused               byte;
    Id                int32;
    nPis                int32;
    nPos                int32;
    pDelays               *int32;
}

type If_LibBox_t_ struct {
    nBoxes                int32;
    vBoxes         *Vec_Ptr_t;
}

func If_Regular(p *If_Obj_t) *If_Obj_t { return (If_Obj_t *)((ABC_PTRUINT_T)(p) & ~01);  }
func If_Not(p *If_Obj_t) *If_Obj_t { return (If_Obj_t *)((ABC_PTRUINT_T)(p) ^  01);  }
func If_NotCond(p *If_Obj_t, c int32) *If_Obj_t { return (If_Obj_t *)((ABC_PTRUINT_T)(p) ^ (c));  }
func If_IsComplement(p *If_Obj_t) int32 { return int32( (((ABC_PTRUINT_T)p) & 01));         }

func If_ManCiNum(p *If_Man_t) int32 { return p.nObjs[IF_CI];               }
func If_ManCoNum(p *If_Man_t) int32 { return p.nObjs[IF_CO];               }
func If_ManAndNum(p *If_Man_t) int32 { return p.nObjs[IF_AND];              }
func If_ManObjNum(p *If_Man_t) int32 { return Vec_PtrSize(p.vObjs);         }

func If_ManConst1(p *If_Man_t) *If_Obj_t { return p.pConst1;                              }
func If_ManCi(p *If_Man_t, i int32) *If_Obj_t { return (If_Obj_t *)Vec_PtrEntry( p.vCis, i );  }
func If_ManCo(p *If_Man_t, i int32) *If_Obj_t { return (If_Obj_t *)Vec_PtrEntry( p.vCos, i );  }
func If_ManLi(p *If_Man_t, i int32) *If_Obj_t { return (If_Obj_t *)Vec_PtrEntry( p.vCos, If_ManCoNum(p) - p.pPars.nLatchesCo + i );  }
func If_ManLo(p *If_Man_t, i int32) *If_Obj_t { return (If_Obj_t *)Vec_PtrEntry( p.vCis, If_ManCiNum(p) - p.pPars.nLatchesCi + i );  }
func If_ManObj(p *If_Man_t, i int32) *If_Obj_t { return (If_Obj_t *)Vec_PtrEntry( p.vObjs, i ); }

func If_ObjIsConst1(pObj *If_Obj_t) int32 { return pObj.Type == IF_CONST1;       }
func If_ObjIsCi(pObj *If_Obj_t) int32 { return pObj.Type == IF_CI;           }
func If_ObjIsCo(pObj *If_Obj_t) int32 { return pObj.Type == IF_CO;           }
func If_ObjIsTerm(pObj *If_Obj_t) int32 { return pObj.Type == IF_CI || pObj.Type == IF_CO; }
func If_ObjIsLatch(pObj *If_Obj_t) int32 { return If_ObjIsCi(pObj) && pObj.pFanin0 != nil;  }
func If_ObjIsAnd(pObj *If_Obj_t) int32 { return pObj.Type == IF_AND;          }

func If_ObjId(pObj *If_Obj_t) int32 { return pObj.Id;                      }
func If_ObjFanin0(pObj *If_Obj_t) *If_Obj_t { return pObj.pFanin0;                 }
func If_ObjFanin1(pObj *If_Obj_t) *If_Obj_t { return pObj.pFanin1;                 }
func If_ObjFaninC0(pObj *If_Obj_t) int32 { return pObj.fCompl0;                 }
func If_ObjFaninC1(pObj *If_Obj_t) int32 { return pObj.fCompl1;                 }
func If_ObjCopy(pObj *If_Obj_t) any { return pObj.pCopy;                   }
func If_ObjLevel(pObj *If_Obj_t) int32 { return pObj.Level;                   }
func If_ObjSetLevel(pObj *If_Obj_t, Level int32) { pObj.Level = Level;                  }
func If_ObjSetCopy(pObj *If_Obj_t, pCopy any) { pObj.pCopy = pCopy;                  }
func If_ObjSetChoice(pObj *If_Obj_t, pEqu *If_Obj_t) { pObj.pEquiv = pEqu;                  }

func If_CutLeaveNum(pCut *If_Cut_t) int32 { return pCut.nLeaves;                             }
func If_CutLeaves(pCut *If_Cut_t) *int32 { return pCut.pLeaves;                             }
func If_CutLeaf(p *If_Man_t, pCut *If_Cut_t, i int32) *If_Obj_t { assert(i >= 0 && i < int32(pCut.nLeaves), "i >= 0 && i < int32(pCut.nLeaves)"); return If_ManObj(p, pCut.pLeaves[i]);                         }
func If_CutSuppMask(pCut *If_Cut_t) uint32 { return (~uint32(0)) >> (32-pCut.nLeaves);      }
func If_CutTruthWords(nVarsMax int32) int32 { if nVarsMax <= 5 {
    return 2;
} else {
    return (1 << (nVarsMax - 5));
} }
func If_CutPermWords(nVarsMax int32) int32 { return nVarsMax / sizeof(int) + ((nVarsMax % sizeof(int)) > 0); }
func If_CutLeafBit(pCut *If_Cut_t, i int32) int32 { return (pCut.uMaskFunc >> i) & 1;                }
func If_CutPerm(pCut *If_Cut_t) string { return (char *)(pCut.pLeaves + pCut.nLeaves);   }
func If_CutCopy(p *If_Man_t, pDst *If_Cut_t, pSrc *If_Cut_t) { memcpy( pDst, pSrc, (size_t)p.nCutBytes );      }
func If_CutSetup(p *If_Man_t, pCut *If_Cut_t) { memset(pCut, 0, (size_t)p.nCutBytes); pCut.nLimit = p.pPars.nLutSize; }

func If_ObjCutBest(pObj *If_Obj_t) *If_Cut_t { return &pObj.CutBest;                }
func If_ObjCutSign(ObjId uint32) uint32 { return (1 << (ObjId % 31));           }
func If_ObjCutSignCompute(p *If_Cut_t) uint32 { var s uint32; var i int32; for  i = 0; i < If_CutLeaveNum(p); i++  { s |= If_ObjCutSign(p.pLeaves[i]); }
 return s; }

func If_ObjArrTime(pObj *If_Obj_t) float32 { return If_ObjCutBest(pObj).Delay;    }
func If_ObjSetArrTime(pObj *If_Obj_t, ArrTime float32) { If_ObjCutBest(pObj).Delay = ArrTime; }

func If_ObjLValue(pObj *If_Obj_t) float32 { return pObj.LValue;                  }
func If_ObjSetLValue(pObj *If_Obj_t, LValue float32) { pObj.LValue = LValue;                }

func If_CutData(pCut *If_Cut_t) any { return *(void **)pCut;                }
func If_CutSetData(pCut *If_Cut_t, pData any) { *(void **)pCut = pData;               }

func If_CutDataInt(pCut *If_Cut_t) int32 { return *(int *)pCut;                  }
func If_CutSetDataInt(pCut *If_Cut_t, Data int32) { *(int *)pCut = Data;                  }

func If_CutTruthLit(pCut *If_Cut_t) int32 { assert(pCut.iCutFunc >= 0, "pCut.iCutFunc >= 0"); return pCut.iCutFunc;             }
func If_CutTruthIsCompl(pCut *If_Cut_t) int32 { assert(pCut.iCutFunc >= 0, "pCut.iCutFunc >= 0"); return Abc_LitIsCompl(pCut.iCutFunc);                               }
func If_CutTruthWR(p *If_Man_t, pCut *If_Cut_t) *word { if p.vTtMem[pCut.nLeaves] {
    return Vec_MemReadEntry(p.vTtMem[pCut.nLeaves], Abc_Lit2Var(pCut.iCutFunc));
} else {
    return nil;
}  }
func If_CutTruthUR(p *If_Man_t, pCut *If_Cut_t) *uint32 { return (unsigned *)If_CutTruthWR(p, pCut);                        }
func If_CutTruthW(p *If_Man_t, pCut *If_Cut_t) *word { assert(pCut.iCutFunc >= 0, "pCut.iCutFunc >= 0"); Abc_TtCopy( p.puTempW, If_CutTruthWR(p, pCut), p.nTruth6Words[pCut.nLeaves], If_CutTruthIsCompl(pCut) ); return p.puTempW;  }
func If_CutTruth(p *If_Man_t, pCut *If_Cut_t) *uint32 { return (unsigned *)If_CutTruthW(p, pCut);                         }

func If_CutDsdLit(p *If_Man_t, pCut *If_Cut_t) int32 { return Abc_Lit2LitL( Vec_IntArray(p.vTtDsds[pCut.nLeaves]), If_CutTruthLit(pCut) );               }
func If_CutDsdIsCompl(p *If_Man_t, pCut *If_Cut_t) int32 { return Abc_LitIsCompl( If_CutDsdLit(p, pCut) );                                                     }
func If_CutDsdPerm(p *If_Man_t, pCut *If_Cut_t) string { return Vec_StrEntryP( p.vTtPerms[pCut.nLeaves], Abc_Lit2Var(pCut.iCutFunc) * Abc_MaxInt(6, pCut.nLeaves) );           }

func If_CutLutArea(p *If_Man_t, pCut *If_Cut_t) float32 { if pCut.fAndCut {
    return p.pPars.nAndArea;
} else {
    if pCut.fUser {
        return float32(pCut.Cost);
    } else {
        if p.pPars.pLutLib {
            return p.pPars.pLutLib.pLutAreas[pCut.nLeaves];
        } else {
            return float32(1.0);
        }
    }
}  }
func If_CutLutDelay(p *If_LibLut_t, Size int32, iPin int32) float32 { if p {
    if p.fVarPinDelays {
        return p.pLutDelays[Size][iPin];
    } else {
        return p.pLutDelays[Size][0];
    }
} else {
    return 1.0;
}                              }


////////////////////////////////////////////////////////////////////////
///                      MACRO DEFINITIONS                           ///
////////////////////////////////////////////////////////////////////////

// Original macro:
// #define IF_MIN(a,b)      (((a) < (b))? (a) : (b))

// Original macro:
// #define IF_MAX(a,b)      (((a) > (b))? (a) : (b))


// the small and large numbers (min/max float are 1.17e-38/3.40e+38)
#define IF_FLOAT_LARGE   ((float)1.0e+20)
#define IF_FLOAT_SMALL   ((float)1.0e-20)
#define IF_INT_LARGE     (10000000)

// iterator over the primary inputs
// Original macro:
// #define If_ManForEachCi( p, pObj, i )                                          \
//     Vec_PtrForEachEntry( If_Obj_t *, p->vCis, pObj, i )

// iterator over the primary outputs
// Original macro:
// #define If_ManForEachCo( p, pObj, i )                                          \
//     Vec_PtrForEachEntry( If_Obj_t *, p->vCos, pObj, i )

// iterator over the primary inputs
// Original macro:
// #define If_ManForEachPi( p, pObj, i )                                          \
//     Vec_PtrForEachEntryStop( If_Obj_t *, p->vCis, pObj, i, If_ManCiNum(p) - p->pPars->nLatchesCi - p->pPars->nLatchesCiBox )

// iterator over the primary outputs
// Original macro:
// #define If_ManForEachPo( p, pObj, i )                                          \
//     Vec_PtrForEachEntryStartStop( If_Obj_t *, p->vCos, pObj, i, p->pPars->nLatchesCoBox, If_ManCoNum(p) - p->pPars->nLatchesCo )

// iterator over the latches 
// Original macro:
// #define If_ManForEachLatchInput( p, pObj, i )                                  \
//     Vec_PtrForEachEntryStart( If_Obj_t *, p->vCos, pObj, i, If_ManCoNum(p) - p->pPars->nLatchesCo )

// Original macro:
// #define If_ManForEachLatchOutput( p, pObj, i )                                 \
//     Vec_PtrForEachEntryStartStop( If_Obj_t *, p->vCis, pObj, i, If_ManCiNum(p) - p->pPars->nLatchesCi - p->pPars->nLatchesCiBox, If_ManCiNum(p) - p->pPars->nLatchesCiBox )

// iterator over all objects in topological order
// Original macro:
// #define If_ManForEachObj( p, pObj, i )                                         \
//     Vec_PtrForEachEntry( If_Obj_t *, p->vObjs, pObj, i )

// iterator over all objects in reverse topological order
// Original macro:
// #define If_ManForEachObjReverse( p, pObj, i )                                  \
//     Vec_PtrForEachEntry( If_Obj_t *, p->vObjsRev, pObj, i )

// iterator over logic nodes 
// Original macro:
// #define If_ManForEachNode( p, pObj, i )                                        \
//     If_ManForEachObj( p, pObj, i ) if ( pObj->Type != IF_AND ) {} else

// iterator over cuts of the node
// Original macro:
// #define If_ObjForEachCut( pObj, pCut, i )                                      \
//     for ( i = 0; (i < (pObj)->pCutSet->nCuts) && ((pCut) = (pObj)->pCutSet->ppCuts[i]); i++ )

// iterator over the leaves of the cut
// Original macro:
// #define If_CutForEachLeaf( p, pCut, pLeaf, i )                                 \
//     for ( i = 0; (i < (int)(pCut)->nLeaves) && ((pLeaf) = If_ManObj(p, (pCut)->pLeaves[i])); i++ )

// Original macro:
// #define If_CutForEachLeafReverse( p, pCut, pLeaf, i )                                 \
//     for ( i = (int)(pCut)->nLeaves - 1; (i >= 0) && ((pLeaf) = If_ManObj(p, (pCut)->pLeaves[i])); i-- )

//#define If_CutForEachLeaf( p, pCut, pLeaf, i )                                 \ \\prevent multiline comment
//    for ( i = 0; (i < (int)(pCut)->nLeaves) && ((pLeaf) = If_ManObj(p, p->pPars->fLiftLeaves? (pCut)->pLeaves[i] >> 8 : (pCut)->pLeaves[i])); i++ )
// iterator over the leaves of the sequential cut
// Original macro:
// #define If_CutForEachLeafSeq( p, pCut, pLeaf, Shift, i )                       \
//     for ( i = 0; (i < (int)(pCut)->nLeaves) && ((pLeaf) = If_ManObj(p, (pCut)->pLeaves[i] >> 8)) && (((Shift) = ((pCut)->pLeaves[i] & 255)) >= 0); i++ )


////////////////////////////////////////////////////////////////////////
///                    FUNCTION DECLARATIONS                         ///
////////////////////////////////////////////////////////////////////////

/*=== ifCore.c ===========================================================*/
// extern void            If_ManSetDefaultPars( If_Par_t * pPars );

// extern int             If_ManPerformMapping( If_Man_t * p );

// extern int             If_ManPerformMappingComb( If_Man_t * p );

// extern void            If_ManComputeSwitching( If_Man_t * p );

/*=== ifCut.c ============================================================*/
// extern int             If_CutVerifyCuts( If_Set_t * pCutSet, int fOrdered );

// extern int             If_CutFilter( If_Set_t * pCutSet, If_Cut_t * pCut, int fSaveCut0 );

// extern void            If_CutSort( If_Man_t * p, If_Set_t * pCutSet, If_Cut_t * pCut );

// extern void            If_CutOrder( If_Cut_t * pCut );

// extern int             If_CutMergeOrdered( If_Man_t * p, If_Cut_t * pCut0, If_Cut_t * pCut1, If_Cut_t * pCut );

// extern int             If_CutMerge( If_Man_t * p, If_Cut_t * pCut0, If_Cut_t * pCut1, If_Cut_t * pCut );

// extern int             If_CutCheck( If_Cut_t * pCut );

// extern void            If_CutPrint( If_Cut_t * pCut );

// extern void            If_CutPrintTiming( If_Man_t * p, If_Cut_t * pCut );

// extern void            If_CutLift( If_Cut_t * pCut );

// extern void            If_CutCopy( If_Man_t * p, If_Cut_t * pCutDest, If_Cut_t * pCutSrc );

// extern float           If_CutAreaFlow( If_Man_t * p, If_Cut_t * pCut );

// extern float           If_CutEdgeFlow( If_Man_t * p, If_Cut_t * pCut );

// extern float           If_CutPowerFlow( If_Man_t * p, If_Cut_t * pCut, If_Obj_t * pRoot );

// extern float           If_CutAverageRefs( If_Man_t * p, If_Cut_t * pCut );

// extern float           If_CutAreaDeref( If_Man_t * p, If_Cut_t * pCut );

// extern float           If_CutAreaRef( If_Man_t * p, If_Cut_t * pCut );

// extern float           If_CutAreaDerefed( If_Man_t * p, If_Cut_t * pCut );

// extern float           If_CutAreaRefed( If_Man_t * p, If_Cut_t * pCut );

// extern float           If_CutEdgeDeref( If_Man_t * p, If_Cut_t * pCut );

// extern float           If_CutEdgeRef( If_Man_t * p, If_Cut_t * pCut );

// extern float           If_CutEdgeDerefed( If_Man_t * p, If_Cut_t * pCut );

// extern float           If_CutEdgeRefed( If_Man_t * p, If_Cut_t * pCut );

// extern float           If_CutPowerDeref( If_Man_t * p, If_Cut_t * pCut, If_Obj_t * pRoot );

// extern float           If_CutPowerRef( If_Man_t * p, If_Cut_t * pCut, If_Obj_t * pRoot );

// extern float           If_CutPowerDerefed( If_Man_t * p, If_Cut_t * pCut, If_Obj_t * pRoot );

// extern float           If_CutPowerRefed( If_Man_t * p, If_Cut_t * pCut, If_Obj_t * pRoot );

/*=== ifDec.c =============================================================*/
// extern word            If_CutPerformDerive07( If_Man_t * p, unsigned * pTruth, int nVars, int nLeaves, char * pStr );

// extern int             If_CutPerformCheck07( If_Man_t * p, unsigned * pTruth, int nVars, int nLeaves, char * pStr );

// extern int             If_CutPerformCheck08( If_Man_t * p, unsigned * pTruth, int nVars, int nLeaves, char * pStr );

// extern int             If_CutPerformCheck10( If_Man_t * p, unsigned * pTruth, int nVars, int nLeaves, char * pStr );

// extern int             If_CutPerformCheck16( If_Man_t * p, unsigned * pTruth, int nVars, int nLeaves, char * pStr );

// extern int             If_CutPerformCheck45( If_Man_t * p, unsigned * pTruth, int nVars, int nLeaves, char * pStr );

// extern int             If_CutPerformCheck54( If_Man_t * p, unsigned * pTruth, int nVars, int nLeaves, char * pStr );

// extern int             If_CutPerformCheck75( If_Man_t * p, unsigned * pTruth, int nVars, int nLeaves, char * pStr );

// extern float           If_CutDelayLutStruct( If_Man_t * p, If_Cut_t * pCut, char * pStr, float WireDelay );

// extern int             If_CluCheckExt( void * p, word * pTruth, int nVars, int nLutLeaf, int nLutRoot, 
//                            char * pLut0, char * pLut1, word * pFunc0, word * pFunc1 );

// extern int             If_CluCheckExt3( void * p, word * pTruth, int nVars, int nLutLeaf, int nLutLeaf2, int nLutRoot, 
//                            char * pLut0, char * pLut1, char * pLut2, word * pFunc0, word * pFunc1, word * pFunc2 );

/*=== ifDelay.c =============================================================*/
// extern int             If_CutDelaySop( If_Man_t * p, If_Cut_t * pCut );

// extern int             If_CutSopBalanceEvalInt( Vec_Int_t * vCover, int * pTimes, int * pFaninLits, Vec_Int_t * vAig, int * piRes, int nSuppAll, int * pArea );

// extern int             If_CutSopBalanceEval( If_Man_t * p, If_Cut_t * pCut, Vec_Int_t * vAig );

// extern int             If_CutSopBalancePinDelaysInt( Vec_Int_t * vCover, int * pTimes, word * pFaninRes, int nSuppAll, word * pRes );

// extern int             If_CutSopBalancePinDelays( If_Man_t * p, If_Cut_t * pCut, char * pPerm );

// extern int             If_CutLutBalanceEval( If_Man_t * p, If_Cut_t * pCut );

// extern int             If_CutLutBalancePinDelays( If_Man_t * p, If_Cut_t * pCut, char * pPerm );

/*=== ifDsd.c =============================================================*/
// extern If_DsdMan_t *   If_DsdManAlloc( int nVars, int nLutSize );

// extern void            If_DsdManAllocIsops( If_DsdMan_t * p, int nLutSize );

// extern void            If_DsdManPrint( If_DsdMan_t * p, char * pFileName, int Number, int Support, int fOccurs, int fTtDump, int fVerbose );

// extern void            If_DsdManTune( If_DsdMan_t * p, int LutSize, int fFast, int fAdd, int fSpec, int fVerbose );

// extern void            Id_DsdManTuneStr( If_DsdMan_t * p, char * pStruct, int nConfls, int nProcs, int fVerbose );

// extern void            If_DsdManFree( If_DsdMan_t * p, int fVerbose );

// extern void            If_DsdManSave( If_DsdMan_t * p, char * pFileName );

// extern If_DsdMan_t *   If_DsdManLoad( char * pFileName );

// extern void            If_DsdManMerge( If_DsdMan_t * p, If_DsdMan_t * pNew );

// extern void            If_DsdManCleanOccur( If_DsdMan_t * p, int fVerbose );

// extern void            If_DsdManCleanMarks( If_DsdMan_t * p, int fVerbose );

// extern void            If_DsdManInvertMarks( If_DsdMan_t * p, int fVerbose );

// extern If_DsdMan_t *   If_DsdManFilter( If_DsdMan_t * p, int Limit );

// extern int             If_DsdManCompute( If_DsdMan_t * p, word * pTruth, int nLeaves, unsigned char * pPerm, char * pLutStruct );

// extern char *          If_DsdManFileName( If_DsdMan_t * p );

// extern int             If_DsdManVarNum( If_DsdMan_t * p );

// extern int             If_DsdManObjNum( If_DsdMan_t * p );

// extern int             If_DsdManLutSize( If_DsdMan_t * p );

// extern int             If_DsdManTtBitNum( If_DsdMan_t * p );

// extern int             If_DsdManPermBitNum( If_DsdMan_t * p );

// extern void            If_DsdManSetLutSize( If_DsdMan_t * p, int nLutSize );

// extern int             If_DsdManSuppSize( If_DsdMan_t * p, int iDsd );

// extern int             If_DsdManCheckDec( If_DsdMan_t * p, int iDsd );

// extern int             If_DsdManReadMark( If_DsdMan_t * p, int iDsd );

// extern void            If_DsdManSetNewAsUseless( If_DsdMan_t * p );

// extern word *          If_DsdManGetFuncConfig( If_DsdMan_t * p, int iDsd );

// extern char *          If_DsdManGetCellStr( If_DsdMan_t * p );

// extern unsigned        If_DsdManCheckXY( If_DsdMan_t * p, int iDsd, int LutSize, int fDerive, unsigned uMaskNot, int fHighEffort, int fVerbose );

// extern int             If_CutDsdBalanceEval( If_Man_t * p, If_Cut_t * pCut, Vec_Int_t * vAig );

// extern int             If_CutDsdBalancePinDelays( If_Man_t * p, If_Cut_t * pCut, char * pPerm );

// extern void            Id_DsdManTuneThresh( If_DsdMan_t * p, int fUnate, int fThresh, int fThreshHeuristic, int fVerbose );

/*=== ifLib.c =============================================================*/
// extern If_LibLut_t *   If_LibLutRead( char * FileName );

// extern If_LibLut_t *   If_LibLutDup( If_LibLut_t * p );

// extern void            If_LibLutFree( If_LibLut_t * pLutLib );

// extern void            If_LibLutPrint( If_LibLut_t * pLutLib );

// extern int             If_LibLutDelaysAreDiscrete( If_LibLut_t * pLutLib );

// extern int             If_LibLutDelaysAreDifferent( If_LibLut_t * pLutLib );

// extern If_LibLut_t *   If_LibLutSetSimple( int nLutSize );

// extern float           If_LibLutFastestPinDelay( If_LibLut_t * p );

// extern float           If_LibLutSlowestPinDelay( If_LibLut_t * p );

/*=== ifLibBox.c =============================================================*/
// extern If_LibBox_t *   If_LibBoxStart();

// extern void            If_LibBoxFree( If_LibBox_t * p );

// extern int             If_LibBoxNum( If_LibBox_t * p );

// extern If_Box_t *      If_LibBoxReadBox( If_LibBox_t * p, int Id );

// extern If_Box_t *      If_LibBoxFindBox( If_LibBox_t * p, char * pName );

// extern void            If_LibBoxAdd( If_LibBox_t * p, If_Box_t * pBox );

// extern If_LibBox_t *   If_LibBoxRead( char * pFileName );

// extern If_LibBox_t *   If_LibBoxRead2( char * pFileName );

// extern void            If_LibBoxPrint( FILE * pFile, If_LibBox_t * p );

// extern void            If_LibBoxWrite( char * pFileName, If_LibBox_t * p );

// extern int             If_LibBoxLoad( char * pFileName );

// extern If_Box_t *      If_BoxStart( char * pName, int Id, int nPis, int nPos, int fSeq, int fBlack, int fOuter );

/*=== ifMan.c =============================================================*/
// extern If_Man_t *      If_ManStart( If_Par_t * pPars );

// extern void            If_ManRestart( If_Man_t * p );

// extern void            If_ManStop( If_Man_t * p );

// extern If_Obj_t *      If_ManCreateCi( If_Man_t * p );

// extern If_Obj_t *      If_ManCreateCo( If_Man_t * p, If_Obj_t * pDriver );

// extern If_Obj_t *      If_ManCreateAnd( If_Man_t * p, If_Obj_t * pFan0, If_Obj_t * pFan1 );

// extern If_Obj_t *      If_ManCreateXor( If_Man_t * p, If_Obj_t * pFan0, If_Obj_t * pFan1 );

// extern If_Obj_t *      If_ManCreateMux( If_Man_t * p, If_Obj_t * pFan0, If_Obj_t * pFan1, If_Obj_t * pCtrl );

// extern void            If_ManCreateChoice( If_Man_t * p, If_Obj_t * pRepr );

// extern void            If_ManSetupCutTriv( If_Man_t * p, If_Cut_t * pCut, int ObjId );

// extern void            If_ManSetupCiCutSets( If_Man_t * p );

// extern If_Set_t *      If_ManSetupNodeCutSet( If_Man_t * p, If_Obj_t * pObj );

// extern void            If_ManDerefNodeCutSet( If_Man_t * p, If_Obj_t * pObj );

// extern void            If_ManDerefChoiceCutSet( If_Man_t * p, If_Obj_t * pObj );

// extern void            If_ManSetupSetAll( If_Man_t * p, int nCrossCut );

/*=== ifMap.c =============================================================*/
// extern int *           If_CutArrTimeProfile( If_Man_t * p, If_Cut_t * pCut );

// extern void            If_ObjPerformMappingAnd( If_Man_t * p, If_Obj_t * pObj, int Mode, int fPreprocess, int fFirst );

// extern void            If_ObjPerformMappingChoice( If_Man_t * p, If_Obj_t * pObj, int Mode, int fPreprocess );

// extern int             If_ManPerformMappingRound( If_Man_t * p, int nCutsUsed, int Mode, int fPreprocess, int fFirst, char * pLabel );

/*=== ifReduce.c ==========================================================*/
// extern void            If_ManImproveMapping( If_Man_t * p );

/*=== ifSat.c ==========================================================*/
// extern void *          If_ManSatBuildXY( int nLutSize );

// extern void *          If_ManSatBuildXYZ( int nLutSize );

// extern void            If_ManSatUnbuild( void * p );

// extern int             If_ManSatCheckXY( void * pSat, int nLutSize, word * pTruth, int nVars, unsigned uSet, word * pTBound, word * pTFree, Vec_Int_t * vLits );

// extern unsigned        If_ManSatCheckXYall( void * pSat, int nLutSize, word * pTruth, int nVars, Vec_Int_t * vLits );

/*=== ifSeq.c =============================================================*/
// extern int             If_ManPerformMappingSeq( If_Man_t * p );

/*=== ifTime.c ============================================================*/
// extern float           If_CutDelay( If_Man_t * p, If_Obj_t * pObj, If_Cut_t * pCut );

// extern void            If_CutPropagateRequired( If_Man_t * p, If_Obj_t * pObj, If_Cut_t * pCut, float Required );

// extern float           If_ManDelayMax( If_Man_t * p, int fSeq );

// extern void            If_ManComputeRequired( If_Man_t * p );

/*=== ifTruth.c ===========================================================*/
// extern void            If_CutRotatePins( If_Man_t * p, If_Cut_t * pCut );

// extern int             If_CutComputeTruth( If_Man_t * p, If_Cut_t * pCut, If_Cut_t * pCut0, If_Cut_t * pCut1, int fCompl0, int fCompl1 );

// extern int             If_CutComputeTruthPerm( If_Man_t * p, If_Cut_t * pCut, If_Cut_t * pCut0, If_Cut_t * pCut1, int fCompl0, int fCompl1 );

// extern Vec_Mem_t *     If_DeriveHashTable6( int nVars, word Truth );

// extern int             If_CutCheckTruth6( If_Man_t * p, If_Cut_t * pCut );

/*=== ifTune.c ===========================================================*/
// extern Ifn_Ntk_t *     Ifn_NtkParse( char * pStr );

// extern int             Ifn_NtkTtBits( char * pStr );

// extern int             Ifn_NtkMatch( Ifn_Ntk_t * p, word * pTruth, int nVars, int nConfls, int fVerbose, int fVeryVerbose, word * pConfig );

// extern void            Ifn_NtkPrint( Ifn_Ntk_t * p );

// extern int             Ifn_NtkLutSizeMax( Ifn_Ntk_t * p );

// extern int             Ifn_NtkInputNum( Ifn_Ntk_t * p );

// extern void *          If_ManSatBuildFromCell( char * pStr, Vec_Int_t ** pvPiVars, Vec_Int_t ** pvPoVars, Ifn_Ntk_t ** ppNtk );

// extern int             If_ManSatFindCofigBits( void * pSat, Vec_Int_t * vPiVars, Vec_Int_t * vPoVars, word * pTruth, int nVars, word Perm, int nInps, Vec_Int_t * vValues );

// extern int             If_ManSatDeriveGiaFromBits( void * pNew, Ifn_Ntk_t * p, word * pTtData, Vec_Int_t * vLeaves, Vec_Int_t * vValues );

// extern void *          If_ManDeriveGiaFromCells( void * p );

/*=== ifUtil.c ============================================================*/
// extern void            If_ManCleanNodeCopy( If_Man_t * p );

// extern void            If_ManCleanCutData( If_Man_t * p );

// extern void            If_ManCleanMarkV( If_Man_t * p );

// extern float           If_ManScanMapping( If_Man_t * p );

// extern float           If_ManScanMappingDirect( If_Man_t * p );

// extern float           If_ManScanMappingSeq( If_Man_t * p );

// extern void            If_ManResetOriginalRefs( If_Man_t * p );

// extern int             If_ManCrossCut( If_Man_t * p );


// extern Vec_Ptr_t *     If_ManReverseOrder( If_Man_t * p );

// extern void            If_ManMarkMapping( If_Man_t * p );

// extern Vec_Ptr_t *     If_ManCollectMappingDirect( If_Man_t * p );

// extern Vec_Int_t *     If_ManCollectMappingInt( If_Man_t * p );


// extern int             If_ManCountSpecialPos( If_Man_t * p );

// extern void            If_CutTraverse( If_Man_t * p, If_Obj_t * pRoot, If_Cut_t * pCut, Vec_Ptr_t * vNodes );

// extern void            If_ObjPrint( If_Obj_t * pObj );



//ABC_NAMESPACE_HEADER_END



////////////////////////////////////////////////////////////////////////
///                       END OF FILE                                ///
////////////////////////////////////////////////////////////////////////

