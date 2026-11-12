package gia

import "fmt"

import "os"

/**CFile****************************************************************

  FileName    [gia.h]

  SystemName  [ABC: Logic synthesis and verification system.]

  PackageName [Scalable AIG package.]

  Synopsis    [External declarations.]

  Author      [Alan Mishchenko]
  
  Affiliation [UC Berkeley]

  Date        [Ver. 1.0. Started - June 20, 2005.]

  Revision    [$Id: gia.h,v 1.00 2005/06/20 00:00:00 alanmi Exp $]

***********************************************************************/
 




////////////////////////////////////////////////////////////////////////
///                          INCLUDES                                ///
////////////////////////////////////////////////////////////////////////

//#include <stdio.h>
//#include <stdlib.h>
//#include <string.h>
//#include <assert.h>

//#include "misc/vec/vec.h"
//#include "misc/vec/vecWec.h"
//#include "misc/util/utilCex.h"

////////////////////////////////////////////////////////////////////////
///                         PARAMETERS                               ///
////////////////////////////////////////////////////////////////////////


//ABC_NAMESPACE_HEADER_START

const GIA_NONE = 0x1FFFFFFF
const GIA_VOID = 0x0FFFFFFF

////////////////////////////////////////////////////////////////////////
///                         BASIC TYPES                              ///
////////////////////////////////////////////////////////////////////////

type Gia_MmFixed_t =        Gia_MmFixed_t_;    
type Gia_MmFlex_t =         Gia_MmFlex_t_;     
type Gia_MmStep_t =         Gia_MmStep_t_;     
type Gia_Dat_t =            Gia_Dat_t_;

   
type Gia_Rpr_t struct {
    iRepr uint32; /* C bitfield: 28 bits; review packing and truncation */  // representative node
    fProved uint32; /* C bitfield: 1 bits; review packing and truncation */  // marks the proved equivalence
    fFailed uint32; /* C bitfield: 1 bits; review packing and truncation */  // marks the failed equivalence
    fColorA uint32; /* C bitfield: 1 bits; review packing and truncation */  // marks cone of A
    fColorB uint32; /* C bitfield: 1 bits; review packing and truncation */  // marks cone of B
}

   
type Gia_Plc_t struct {
    fFixed uint32; /* C bitfield: 1 bits; review packing and truncation */  // the placement of this object is fixed
    xCoord uint32; /* C bitfield: 15 bits; review packing and truncation */  // x-ooordinate of the placement
    fUndef uint32; /* C bitfield: 1 bits; review packing and truncation */  // the placement of this object is not assigned
    yCoord uint32; /* C bitfield: 15 bits; review packing and truncation */  // y-ooordinate of the placement
}

   
type Gia_Obj_t struct {
    iDiff0 uint32; /* C bitfield: 29 bits; review packing and truncation */  // the diff of the first fanin
    fCompl0 uint32; /* C bitfield: 1 bits; review packing and truncation */  // the complemented attribute
    fMark0 uint32; /* C bitfield: 1 bits; review packing and truncation */  // first user-controlled mark
    fTerm uint32; /* C bitfield: 1 bits; review packing and truncation */  // terminal node (CI/CO)

    iDiff1 uint32; /* C bitfield: 29 bits; review packing and truncation */  // the diff of the second fanin
    fCompl1 uint32; /* C bitfield: 1 bits; review packing and truncation */  // the complemented attribute
    fMark1 uint32; /* C bitfield: 1 bits; review packing and truncation */  // second user-controlled mark
    fPhase uint32; /* C bitfield: 1 bits; review packing and truncation */  // value under 000 pattern

    Value       uint32;         // application-specific value
}
// Value is currently used to store several types of information
// - pointer to the next node in the hash table during structural hashing
// - pointer to the node copy during duplication 

// new AIG manager
   
type Gia_Man_t struct {
    pName          string;         // name of the AIG
    pSpec          string;         // name of the input file
    nRegs            int32;         // number of registers
    nRegsAlloc            int32;    // number of allocated registers
    nObjs            int32;         // number of objects
    nObjsAlloc            int32;    // number of allocated objects
    pObjs     *Gia_Obj_t;         // the array of objects
    pMuxes      *uint32;        // control signals of MUXes
    nXors            int32;         // the number of XORs
    nMuxes            int32;        // the number of MUXes 
    nBufs            int32;         // the number of buffers
    vCis     *Vec_Int_t;          // the vector of CIs (PIs + LOs)
    vCos     *Vec_Int_t;          // the vector of COs (POs + LIs)
    vHash      Vec_Int_t;         // hash links
    vHTable      Vec_Int_t;       // hash table
    fAddStrash            int32;    // performs additional structural hashing
    fSweeper            int32;      // sweeper is running
    fGiaSimple            int32;    // simple mode (no const-propagation and strashing)
    vRefs      Vec_Int_t;         // the reference count
    pRefs           *int32;         // the reference count
    pLutRefs           *int32;      // the reference count
    vLevels     *Vec_Int_t;       // levels of the nodes
    nLevels            int32;       // the mamixum level
    nConstrs            int32;      // the number of constraints
    nTravIds            int32;      // the current traversal ID
    nFront            int32;        // frontier size 
    pReprsOld           *int32;     // representatives (for CIs and ANDs)
    pReprs     *Gia_Rpr_t;        // representatives (for CIs and ANDs)
    pNexts           *int32;        // next nodes in the equivalence classes
    pSibls           *int32;        // next nodes in the choice nodes
    pIso           *int32;          // pairs of structurally isomorphic nodes
    nTerLoop            int32;      // the state where loop begins  
    nTerStates            int32;    // the total number of ternary states
    pFanData           *int32;      // the database to store fanout information
    nFansAlloc            int32;    // the size of fanout representation
    vFanoutNums     *Vec_Int_t;   // static fanout
    vFanout     *Vec_Int_t;       // static fanout
    vMapping     *Vec_Int_t;      // mapping for each node
    vMapping2     *Vec_Wec_t;     // mapping for each node
    vFanouts2     *Vec_Wec_t;     // mapping fanouts 
    vCellMapping     *Vec_Int_t;  // mapping for each node
    pSatlutWinman          any; // windowing for SAT-based mapping
    vPacking     *Vec_Int_t;      // packing information
    vConfigs     *Vec_Int_t;      // cell configurations
    pCellStr          string;      // cell description
    vLutConfigs     *Vec_Int_t;   // LUT configurations
    vEdgeDelay     *Vec_Int_t;    // special edge information
    vEdgeDelayR     *Vec_Int_t;   // special edge information
    vEdge1     *Vec_Int_t;        // special edge information
    vEdge2     *Vec_Int_t;        // special edge information
    pCexComb     *Abc_Cex_t;      // combinational counter-example
    pCexSeq     *Abc_Cex_t;       // sequential counter-example
    vSeqModelVec     *Vec_Ptr_t;  // sequential counter-examples
    vCopies      Vec_Int_t;       // intermediate copies
    vCopies2      Vec_Int_t;      // intermediate copies
    vVar2Obj     *Vec_Int_t;      // mapping of variables into objects
    vTruths     *Vec_Int_t;       // used for truth table computation
    vFlopClasses     *Vec_Int_t;  // classes of flops for retiming/merging/etc
    vGateClasses     *Vec_Int_t;  // classes of gates for abstraction
    vObjClasses     *Vec_Int_t;   // classes of objects for abstraction
    vInitClasses     *Vec_Int_t;  // classes of flops for retiming/merging/etc
    vRegClasses     *Vec_Int_t;   // classes of registers for sequential synthesis
    vRegInits     *Vec_Int_t;     // initial state
    vDoms     *Vec_Int_t;         // dominators
    vBarBufs     *Vec_Int_t;      // barrier buffers
    vXors     *Vec_Int_t;         // temporary XORs
    unsigned char* pSwitching;    // switching activity for each object
    pPlacement     *Gia_Plc_t;    // placement of the objects
    pAigExtra     *Gia_Man_t;     // combinational logic of holes
    vInArrs     *Vec_Flt_t;       // PI arrival times
    vOutReqs     *Vec_Flt_t;      // PO required times
    vCiArrs     *Vec_Int_t;       // CI arrival times
    vCoReqs     *Vec_Int_t;       // CO required times
    vCoArrs     *Vec_Int_t;       // CO arrival times
    vCoAttrs     *Vec_Int_t;      // CO attributes
    And2Delay            int32;     // delay of the AND gate 
    DefInArrs          float32;     // default PI arrival times
    DefOutReqs          float32;    // default PO required times
    vSwitching     *Vec_Int_t;    // switching activity
    pTravIds           *int32;      // separate traversal ID representation
    nTravIdsAlloc            int32; // the number of trav IDs allocated
    vNamesIn     *Vec_Ptr_t;      // the input names 
    vNamesOut     *Vec_Ptr_t;     // the output names
    vUserPiIds     *Vec_Int_t;    // numbers assigned to PIs by the user
    vUserPoIds     *Vec_Int_t;    // numbers assigned to POs by the user
    vUserFfIds     *Vec_Int_t;    // numbers assigned to FFs by the user
    vCiNumsOrig     *Vec_Int_t;   // original CI names
    vCoNumsOrig     *Vec_Int_t;   // original CO names
    vIdsOrig     *Vec_Int_t;      // original object IDs
    vIdsEquiv     *Vec_Int_t;     // original object IDs proved equivalent
    vCofVars     *Vec_Int_t;      // cofactoring variables
    vClockDoms     *Vec_Vec_t;    // clock domains
    vTiming     *Vec_Flt_t;       // arrival/required/slack
    pManTime          any;      // the timing manager
    pLutLib          any;       // LUT library
    nHashHit           word;      // hash table hit
    nHashMiss           word;     // hash table miss
    pData          any;         // various user data
    pData2      *uint32;        // various user data
    iData            int32;         // various user data
    iData2            int32;        // various user data
    nAnd2Delay            int32;    // AND2 delay scaled to match delay numbers used
    fVerbose            int32;      // verbose reports
    MappedArea            int32;    // area after mapping
    MappedDelay            int32;   // delay after mapping
    // bit-parallel simulation
    fBuiltInSim            int32;
    iPatsPi            int32;
    nSimWords            int32;
    nSimWordsT            int32;
    iPastPiMax            int32;
    nSimWordsMax            int32;
    vSims     *Vec_Wrd_t;
    vSimsT     *Vec_Wrd_t;
    vSimsPi     *Vec_Wrd_t;
    vSimsPo     *Vec_Wrd_t;
    vClassOld     *Vec_Int_t;
    vClassNew     *Vec_Int_t;
    // incremental simulation
    fIncrSim            int32;
    iNextPi            int32;
    iTimeStamp            int32;
    vTimeStamps     *Vec_Int_t;
    // truth table computation for small functions
    nTtVars            int32;       // truth table variables
    nTtWords            int32;      // truth table words
    vTtNums     *Vec_Int_t;       // object numbers
    vTtNodes     *Vec_Int_t;      // internal nodes
    vTtInputs     *Vec_Ptr_t;     // truth tables for constant and primary inputs
    vTtMemory     *Vec_Wrd_t;     // truth tables for internal nodes
    // balancing
    vSuper     *Vec_Int_t;        // supergate
    vStore     *Vec_Int_t;        // node storage  
    // existential quantification
    iSuppPi            int32;       // the number of support variables
    nSuppWords            int32;    // the number of support words
    vSuppWords     *Vec_Wrd_t;    // support information
    vCopiesTwo      Vec_Int_t;    // intermediate copies
    vSuppVars      Vec_Int_t;     // used variables
    vVarMap      Vec_Int_t;       // used variables
    pUData     *Gia_Dat_t;
}


   
type Gps_Par_t struct {
    fTents            int32;
    fSwitch            int32;
    fCut            int32;
    fNpn            int32;
    fLutProf            int32;
    fMuxXor            int32;
    fMiter            int32;
    fSkipMap            int32;
    fSlacks            int32;
    fNoColor            int32;
    pDumpFile          string;
}

   
type Emb_Par_t struct {
    nDims            int32;         // the number of dimension
    nSols            int32;         // the number of solutions (typically, 2)
    nIters            int32;        // the number of iterations of FORCE
    fRefine            int32;       // use refinement by FORCE
    fCluster            int32;      // use clustered representation 
    fDump            int32;         // dump Gnuplot file
    fDumpLarge            int32;    // dump Gnuplot file for large benchmarks
    fShowImage            int32;    // shows image if Gnuplot is installed
    fVerbose            int32;      // verbose flag  
}


// frames parameters
   
type Gia_ParFra_t struct {
    nFrames            int32;       // the number of frames to unroll
    fInit            int32;         // initialize the timeframes
    fSaveLastLit            int32;  // adds POs for outputs of each frame
    fDisableSt            int32;    // disables strashing
    fOrPos            int32;        // ORs respective POs in each timeframe
    fVerbose            int32;      // enables verbose output
}


// simulation parameters
   
type Gia_ParSim_t struct {
    // user-controlled parameters
    nWords            int32;        // the number of machine words
    nIters            int32;        // the number of timeframes
    RandSeed            int32;      // seed to generate random numbers
    TimeLimit            int32;     // time limit in seconds
    fCheckMiter            int32;   // check if miter outputs are non-zero
    fVerbose            int32;      // enables verbose output
    iOutFail            int32;      // index of the failed output
}

   
type Gia_ManSim_t struct {
    pAig     *Gia_Man_t;
    pPars  *Gia_ParSim_t; 
    nWords            int32;
    vCis2Ids     *Vec_Int_t;
    vConsts     *Vec_Int_t;
    // simulation information
    pDataSim      *uint32;     // simulation data
    pDataSimCis      *uint32;  // simulation data for CIs
    pDataSimCos      *uint32;  // simulation data for COs
}

    
type Jf_Par_t struct {
    nLutSize            int32;
    nCutNum            int32;
    nProcNum            int32;
    nRounds            int32;
    nRoundsEla            int32;
    nRelaxRatio            int32;
    nCoarseLimit            int32;
    nAreaTuner            int32;
    nReqTimeFlex            int32;
    nVerbLimit            int32;
    nDelayLut1            int32;
    nDelayLut2            int32;
    nFastEdges            int32;
    DelayTarget            int32;
    fAreaOnly            int32;
    fPinPerm            int32;
    fPinQuick            int32;
    fPinFilter            int32;
    fOptEdge            int32;
    fUseMux7            int32;
    fPower            int32;
    fCoarsen            int32;
    fCutMin            int32;
    fFuncDsd            int32;
    fGenCnf            int32;
    fGenLit            int32;
    fCnfObjIds            int32;
    fAddOrCla            int32;
    fCnfMapping            int32;
    fPureAig            int32;
    fDoAverage            int32;
    fCutHashing            int32;
    fCutSimple            int32;
    fCutGroup            int32;
    fVerbose            int32;
    fVeryVerbose            int32;
    nLutSizeMax            int32;
    nCutNumMax            int32;
    nProcNumMax            int32;
    nLutSizeMux            int32;
    Delay           word;
    Area           word;
    Edge           word;
    Clause           word;
    Mux7           word;
    WordMapDelay           word;
    WordMapArea           word;
    WordMapDelayTarget           word;
    MapDelay            int32;
    MapArea          float32;
    MapAreaF          float32;
    MapDelayTarget          float32;
    Epsilon          float32;
    pTimesArr         *float32;
    pTimesReq         *float32;
}

func Gia_ObjCutSign(ObjId uint32) uint32 { return (1 << (ObjId & 31));                                 }
func Gia_WordHasOneBit(uWord uint32) int32 { return (uWord & (uWord-1)) == 0;                            }
func Gia_WordHasOnePair(uWord uint32) int32 { return Gia_WordHasOneBit(uWord & (uWord>>1) & 0x55555555);  }
func Gia_WordCountOnes(uWord uint32) int32 {
    uWord = (uWord & 0x55555555) + ((uWord>>1) & 0x55555555);
    uWord = (uWord & 0x33333333) + ((uWord>>2) & 0x33333333);
    uWord = (uWord & 0x0F0F0F0F) + ((uWord>>4) & 0x0F0F0F0F);
    uWord = (uWord & 0x00FF00FF) + ((uWord>>8) & 0x00FF00FF);
    return  (uWord & 0x0000FFFF) + (uWord>>16);
}
func Gia_WordFindFirstBit(uWord uint32) int32 {
    var i int32;
    for  i = 0; i < 32; i++  {
        if uWord & (1 << i) {
            return i; }
    }
    return -1;
}

func Gia_ManTruthIsConst0(pIn *uint32, nVars int32) int32 {
    var w int32;
    for  w = Abc_TruthWordNum(nVars)-1; w >= 0; w--  {
        if pIn[w] {
            return 0; }
    }
    return 1;
}
func Gia_ManTruthIsConst1(pIn *uint32, nVars int32) int32 {
    var w int32;
    for  w = Abc_TruthWordNum(nVars)-1; w >= 0; w--  {
        if pIn[w] != ~uint32(0) {
            return 0; }
    }
    return 1;
}
func Gia_ManTruthCopy(pOut *uint32, pIn *uint32, nVars int32) {
    var w int32;
    for  w = Abc_TruthWordNum(nVars)-1; w >= 0; w--  {
        pOut[w] = pIn[w];
    }
}
func Gia_ManTruthClear(pOut *uint32, nVars int32) {
    var w int32;
    for  w = Abc_TruthWordNum(nVars)-1; w >= 0; w--  {
        pOut[w] = 0;
    }
}
func Gia_ManTruthFill(pOut *uint32, nVars int32) {
    var w int32;
    for  w = Abc_TruthWordNum(nVars)-1; w >= 0; w--  {
        pOut[w] = ~uint32(0);
    }
}
func Gia_ManTruthNot(pOut *uint32, pIn *uint32, nVars int32) {
    var w int32;
    for  w = Abc_TruthWordNum(nVars)-1; w >= 0; w--  {
        pOut[w] = ~pIn[w];
    }
}

func Gia_ManConst0Lit() int32 { return 0; }
func Gia_ManConst1Lit() int32 { return 1; }
func Gia_ManIsConst0Lit(iLit int32) int32 { return (iLit == 0); }
func Gia_ManIsConst1Lit(iLit int32) int32 { return (iLit == 1); }
func Gia_ManIsConstLit(iLit int32) int32 { return (iLit <= 1); }

func Gia_Regular(p *Gia_Obj_t) *Gia_Obj_t { return (Gia_Obj_t *)((ABC_PTRUINT_T)(p) & ~01);                           }
func Gia_Not(p *Gia_Obj_t) *Gia_Obj_t { return (Gia_Obj_t *)((ABC_PTRUINT_T)(p) ^  01);                           }
func Gia_NotCond(p *Gia_Obj_t, c int32) *Gia_Obj_t { return (Gia_Obj_t *)((ABC_PTRUINT_T)(p) ^ (c));                           }
func Gia_IsComplement(p *Gia_Obj_t) int32 { return int32(((ABC_PTRUINT_T)(p) & 01));                                    }

func Gia_ManName(p *Gia_Man_t) string { return p.pName;                                                          }
func Gia_ManCiNum(p *Gia_Man_t) int32 { return Vec_IntSize(p.vCis);                                              }
func Gia_ManCoNum(p *Gia_Man_t) int32 { return Vec_IntSize(p.vCos);                                              }
func Gia_ManPiNum(p *Gia_Man_t) int32 { return Vec_IntSize(p.vCis) - p.nRegs;                                   }
func Gia_ManPoNum(p *Gia_Man_t) int32 { return Vec_IntSize(p.vCos) - p.nRegs;                                   }
func Gia_ManRegNum(p *Gia_Man_t) int32 { return p.nRegs;                                                          }
func Gia_ManObjNum(p *Gia_Man_t) int32 { return p.nObjs;                                                          }
func Gia_ManAndNum(p *Gia_Man_t) int32 { return p.nObjs - Vec_IntSize(p.vCis) - Vec_IntSize(p.vCos) - 1;        }
func Gia_ManXorNum(p *Gia_Man_t) int32 { return p.nXors;                                                          }
func Gia_ManMuxNum(p *Gia_Man_t) int32 { return p.nMuxes;                                                         }
func Gia_ManBufNum(p *Gia_Man_t) int32 { return p.nBufs;                                                          }
func Gia_ManAndNotBufNum(p *Gia_Man_t) int32 { return Gia_ManAndNum(p) - Gia_ManBufNum(p);                               }
func Gia_ManCandNum(p *Gia_Man_t) int32 { return Gia_ManCiNum(p) + Gia_ManAndNum(p);                                }
func Gia_ManConstrNum(p *Gia_Man_t) int32 { return p.nConstrs;                                                       }
func Gia_ManFlipVerbose(p *Gia_Man_t) { p.fVerbose ^= 1;                                                         } 
func Gia_ManHasChoices(p *Gia_Man_t) int32 { return p.pSibls != nil;                                                 } 
func Gia_ManChoiceNum(p *Gia_Man_t) int32 { var c int32; if p.pSibls { var i int32; for i = 0; i < p.nObjs; i++ { c += int32((p.pSibls[i] > 0)); } }
 return c; } 

func Gia_ManConst0(p *Gia_Man_t) *Gia_Obj_t { return p.pObjs;                                                          }
func Gia_ManConst1(p *Gia_Man_t) *Gia_Obj_t { return Gia_Not(Gia_ManConst0(p));                                         }
func Gia_ManObj(p *Gia_Man_t, v int32) *Gia_Obj_t { assert(v >= 0 && v < p.nObjs, "v >= 0 && v < p.nObjs"); return p.pObjs + v;                    }
func Gia_ManCi(p *Gia_Man_t, v int32) *Gia_Obj_t { return Gia_ManObj( p, Vec_IntEntry(p.vCis,v) );                          }
func Gia_ManCo(p *Gia_Man_t, v int32) *Gia_Obj_t { return Gia_ManObj( p, Vec_IntEntry(p.vCos,v) );                          }
func Gia_ManPi(p *Gia_Man_t, v int32) *Gia_Obj_t { assert(v < Gia_ManPiNum(p), "v < Gia_ManPiNum(p)");  return Gia_ManCi( p, v );                 }
func Gia_ManPo(p *Gia_Man_t, v int32) *Gia_Obj_t { assert(v < Gia_ManPoNum(p), "v < Gia_ManPoNum(p)");  return Gia_ManCo( p, v );                 }
func Gia_ManRo(p *Gia_Man_t, v int32) *Gia_Obj_t { assert(v < Gia_ManRegNum(p), "v < Gia_ManRegNum(p)"); return Gia_ManCi( p, Gia_ManPiNum(p)+v ); }
func Gia_ManRi(p *Gia_Man_t, v int32) *Gia_Obj_t { assert(v < Gia_ManRegNum(p), "v < Gia_ManRegNum(p)"); return Gia_ManCo( p, Gia_ManPoNum(p)+v ); }

func Gia_ObjId(p *Gia_Man_t, pObj *Gia_Obj_t) int32 { assert(p.pObjs <= pObj && pObj < p.pObjs + p.nObjs, "p.pObjs <= pObj && pObj < p.pObjs + p.nObjs"); return pObj - p.pObjs; }
func Gia_ObjCioId(pObj *Gia_Obj_t) int32 { assert(pObj.fTerm, "pObj.fTerm"); return pObj.iDiff1;                 }
func Gia_ObjSetCioId(pObj *Gia_Obj_t, v int32) { assert(pObj.fTerm, "pObj.fTerm"); pObj.iDiff1 = v;                    }
func Gia_ObjValue(pObj *Gia_Obj_t) int32 { return pObj.Value;                                         }
func Gia_ObjSetValue(pObj *Gia_Obj_t, i int32) { pObj.Value = i;                                            }
func Gia_ObjPhase(pObj *Gia_Obj_t) int32 { return pObj.fPhase;                                        }
func Gia_ObjPhaseReal(pObj *Gia_Obj_t) int32 { return Gia_Regular(pObj).fPhase ^ Gia_IsComplement(pObj);  }
func Gia_ObjPhaseDiff(p *Gia_Man_t, i int32, k int32) int32 { return Gia_ManObj(p, i).fPhase ^ Gia_ManObj(p, k).fPhase; }

func Gia_ObjIsTerm(pObj *Gia_Obj_t) int32 { return pObj.fTerm;                             } 
func Gia_ObjIsAndOrConst0(pObj *Gia_Obj_t) int32 { return!pObj.fTerm;                             } 
func Gia_ObjIsCi(pObj *Gia_Obj_t) int32 { return pObj.fTerm && pObj.iDiff0 == GIA_NONE; } 
func Gia_ObjIsCo(pObj *Gia_Obj_t) int32 { return pObj.fTerm && pObj.iDiff0 != GIA_NONE; } 
func Gia_ObjIsAnd(pObj *Gia_Obj_t) int32 { return!pObj.fTerm && pObj.iDiff0 != GIA_NONE; } 
func Gia_ObjIsXor(pObj *Gia_Obj_t) int32 { return Gia_ObjIsAnd(pObj) && pObj.iDiff0 < pObj.iDiff1; } 
func Gia_ObjIsMuxId(p *Gia_Man_t, iObj int32) int32 { return p.pMuxes && p.pMuxes[iObj] > 0;        } 
func Gia_ObjIsMux(p *Gia_Man_t, pObj *Gia_Obj_t) int32 { return Gia_ObjIsMuxId( p, Gia_ObjId(p, pObj) ); } 
func Gia_ObjIsAndReal(p *Gia_Man_t, pObj *Gia_Obj_t) int32 { return Gia_ObjIsAnd(pObj) && pObj.iDiff0 > pObj.iDiff1 && !Gia_ObjIsMux(p, pObj); } 
func Gia_ObjIsBuf(pObj *Gia_Obj_t) int32 { return pObj.iDiff0 == pObj.iDiff1 && pObj.iDiff0 != GIA_NONE && !pObj.fTerm;    } 
func Gia_ObjIsAndNotBuf(pObj *Gia_Obj_t) int32 { return Gia_ObjIsAnd(pObj) && pObj.iDiff0 != pObj.iDiff1; } 
func Gia_ObjIsCand(pObj *Gia_Obj_t) int32 { return Gia_ObjIsAnd(pObj) || Gia_ObjIsCi(pObj); } 
func Gia_ObjIsConst0(pObj *Gia_Obj_t) int32 { return pObj.iDiff0 == GIA_NONE && pObj.iDiff1 == GIA_NONE;     } 
func Gia_ManObjIsConst0(p *Gia_Man_t, pObj *Gia_Obj_t) int32 { return pObj == p.pObjs;                        } 

func Gia_Obj2Lit(p *Gia_Man_t, pObj *Gia_Obj_t) int32 { return Abc_Var2Lit(Gia_ObjId(p, Gia_Regular(pObj)), Gia_IsComplement(pObj)); }
func Gia_Lit2Obj(p *Gia_Man_t, iLit int32) *Gia_Obj_t { return Gia_NotCond(Gia_ManObj(p, Abc_Lit2Var(iLit)), Abc_LitIsCompl(iLit));  }
func Gia_ManCiLit(p *Gia_Man_t, CiId int32) int32 { return Gia_Obj2Lit( p, Gia_ManCi(p, CiId) );                }

func Gia_ManIdToCioId(p *Gia_Man_t, Id int32) int32 { return Gia_ObjCioId( Gia_ManObj(p, Id) );                   }
func Gia_ManCiIdToId(p *Gia_Man_t, CiId int32) int32 { return Gia_ObjId( p, Gia_ManCi(p, CiId) );                  }
func Gia_ManCoIdToId(p *Gia_Man_t, CoId int32) int32 { return Gia_ObjId( p, Gia_ManCo(p, CoId) );                  }

func Gia_ObjIsPi(p *Gia_Man_t, pObj *Gia_Obj_t) int32 { return Gia_ObjIsCi(pObj) && Gia_ObjCioId(pObj) < Gia_ManPiNum(p);   } 
func Gia_ObjIsPo(p *Gia_Man_t, pObj *Gia_Obj_t) int32 { return Gia_ObjIsCo(pObj) && Gia_ObjCioId(pObj) < Gia_ManPoNum(p);   } 
func Gia_ObjIsRo(p *Gia_Man_t, pObj *Gia_Obj_t) int32 { return Gia_ObjIsCi(pObj) && Gia_ObjCioId(pObj) >= Gia_ManPiNum(p);  } 
func Gia_ObjIsRi(p *Gia_Man_t, pObj *Gia_Obj_t) int32 { return Gia_ObjIsCo(pObj) && Gia_ObjCioId(pObj) >= Gia_ManPoNum(p);  } 

func Gia_ObjRoToRi(p *Gia_Man_t, pObj *Gia_Obj_t) *Gia_Obj_t { assert((Gia_ObjIsRo(p, pObj) != 0), "Gia_ObjIsRo(p, pObj)"); return Gia_ManCo(p, Gia_ManCoNum(p) - Gia_ManCiNum(p) + Gia_ObjCioId(pObj)); } 
func Gia_ObjRiToRo(p *Gia_Man_t, pObj *Gia_Obj_t) *Gia_Obj_t { assert((Gia_ObjIsRi(p, pObj) != 0), "Gia_ObjIsRi(p, pObj)"); return Gia_ManCi(p, Gia_ManCiNum(p) - Gia_ManCoNum(p) + Gia_ObjCioId(pObj)); } 
func Gia_ObjRoToRiId(p *Gia_Man_t, ObjId int32) int32 { return Gia_ObjId( p, Gia_ObjRoToRi( p, Gia_ManObj(p, ObjId) ) );    } 
func Gia_ObjRiToRoId(p *Gia_Man_t, ObjId int32) int32 { return Gia_ObjId( p, Gia_ObjRiToRo( p, Gia_ManObj(p, ObjId) ) );    }

func Gia_ObjDiff0(pObj *Gia_Obj_t) int32 { return pObj.iDiff0;         }
func Gia_ObjDiff1(pObj *Gia_Obj_t) int32 { return pObj.iDiff1;         }
func Gia_ObjFaninC0(pObj *Gia_Obj_t) int32 { return pObj.fCompl0;        }
func Gia_ObjFaninC1(pObj *Gia_Obj_t) int32 { return pObj.fCompl1;        }
func Gia_ObjFaninC2(p *Gia_Man_t, pObj *Gia_Obj_t) int32 { return p.pMuxes && Abc_LitIsCompl(p.pMuxes[Gia_ObjId(p, pObj)]);  }
func Gia_ObjFanin0(pObj *Gia_Obj_t) *Gia_Obj_t { return pObj - pObj.iDiff0;  }
func Gia_ObjFanin1(pObj *Gia_Obj_t) *Gia_Obj_t { return pObj - pObj.iDiff1;  }
func Gia_ObjFanin2(p *Gia_Man_t, pObj *Gia_Obj_t) *Gia_Obj_t { if p.pMuxes {
    return Gia_ManObj(p, Abc_Lit2Var(p.pMuxes[Gia_ObjId(p, pObj)]));
} else {
    return nil;
}  }
func Gia_ObjChild0(pObj *Gia_Obj_t) *Gia_Obj_t { return Gia_NotCond( Gia_ObjFanin0(pObj), Gia_ObjFaninC0(pObj) ); }
func Gia_ObjChild1(pObj *Gia_Obj_t) *Gia_Obj_t { return Gia_NotCond( Gia_ObjFanin1(pObj), Gia_ObjFaninC1(pObj) ); }
func Gia_ObjChild2(p *Gia_Man_t, pObj *Gia_Obj_t) *Gia_Obj_t { return Gia_NotCond( Gia_ObjFanin2(p, pObj), Gia_ObjFaninC2(p, pObj) ); }
func Gia_ObjFaninId0(pObj *Gia_Obj_t, ObjId int32) int32 { return ObjId - pObj.iDiff0;    }
func Gia_ObjFaninId1(pObj *Gia_Obj_t, ObjId int32) int32 { return ObjId - pObj.iDiff1;    }
func Gia_ObjFaninId2(p *Gia_Man_t, ObjId int32) int32 { if p.pMuxes && p.pMuxes[ObjId] {
    return Abc_Lit2Var(p.pMuxes[ObjId]);
} else {
    return -1;
} }
func Gia_ObjFaninId0p(p *Gia_Man_t, pObj *Gia_Obj_t) int32 { return Gia_ObjFaninId0( pObj, Gia_ObjId(p, pObj) );              }
func Gia_ObjFaninId1p(p *Gia_Man_t, pObj *Gia_Obj_t) int32 { return Gia_ObjFaninId1( pObj, Gia_ObjId(p, pObj) );              }
func Gia_ObjFaninId2p(p *Gia_Man_t, pObj *Gia_Obj_t) int32 { if p.pMuxes && p.pMuxes[Gia_ObjId(p, pObj)] {
    return Abc_Lit2Var(p.pMuxes[Gia_ObjId(p, pObj)]);
} else {
    return -1;
} }
func Gia_ObjFaninLit0(pObj *Gia_Obj_t, ObjId int32) int32 { return Abc_Var2Lit( Gia_ObjFaninId0(pObj, ObjId), Gia_ObjFaninC0(pObj) ); }
func Gia_ObjFaninLit1(pObj *Gia_Obj_t, ObjId int32) int32 { return Abc_Var2Lit( Gia_ObjFaninId1(pObj, ObjId), Gia_ObjFaninC1(pObj) ); }
func Gia_ObjFaninLit2(p *Gia_Man_t, ObjId int32) int32 { if p.pMuxes && p.pMuxes[ObjId] {
    return p.pMuxes[ObjId];
} else {
    return -1;
}           }
func Gia_ObjFaninLit0p(p *Gia_Man_t, pObj *Gia_Obj_t) int32 { return Abc_Var2Lit( Gia_ObjFaninId0p(p, pObj), Gia_ObjFaninC0(pObj) );    }
func Gia_ObjFaninLit1p(p *Gia_Man_t, pObj *Gia_Obj_t) int32 { return Abc_Var2Lit( Gia_ObjFaninId1p(p, pObj), Gia_ObjFaninC1(pObj) );    }
func Gia_ObjFaninLit2p(p *Gia_Man_t, pObj *Gia_Obj_t) int32 { if p.pMuxes && p.pMuxes[Gia_ObjId(p, pObj)] {
    return p.pMuxes[Gia_ObjId(p, pObj)];
} else {
    return -1;
}    }
func Gia_ObjFlipFaninC0(pObj *Gia_Obj_t) { assert((Gia_ObjIsCo(pObj) != 0), "Gia_ObjIsCo(pObj)"); pObj.fCompl0 ^= 1;          }
func Gia_ObjFaninNum(p *Gia_Man_t, pObj *Gia_Obj_t) int32 { if Gia_ObjIsMux(p, pObj) != 0 { return 3; }
 if Gia_ObjIsAnd(pObj) != 0 { return 2; }
 if Gia_ObjIsCo(pObj) != 0 { return 1; }
 return 0; }
func Gia_ObjWhatFanin(p *Gia_Man_t, pObj *Gia_Obj_t, pFanin *Gia_Obj_t) int32 { if Gia_ObjFanin0(pObj) == pFanin { return 0; }
 if Gia_ObjFanin1(pObj) == pFanin { return 1; }
 if Gia_ObjFanin2(p, pObj) == pFanin { return 2; }
 assert(false, "0"); return -1; }

func Gia_ManPoIsConst(p *Gia_Man_t, iPoIndex int32) int32 { return Gia_ObjFaninId0p(p, Gia_ManPo(p, iPoIndex)) == 0;                   }
func Gia_ManPoIsConst0(p *Gia_Man_t, iPoIndex int32) int32 { return Gia_ManIsConst0Lit( Gia_ObjFaninLit0p(p, Gia_ManPo(p, iPoIndex)) ); }
func Gia_ManPoIsConst1(p *Gia_Man_t, iPoIndex int32) int32 { return Gia_ManIsConst1Lit( Gia_ObjFaninLit0p(p, Gia_ManPo(p, iPoIndex)) ); }

func Gia_ObjCopy(p *Gia_Man_t, pObj *Gia_Obj_t) *Gia_Obj_t { return Gia_ManObj( p, Abc_Lit2Var(pObj.Value) );                              }
func Gia_ObjLitCopy(p *Gia_Man_t, iLit int32) int32 { return Abc_LitNotCond( Gia_ManObj(p, Abc_Lit2Var(iLit)).Value, Abc_LitIsCompl(iLit));     }

func Gia_ObjFanin0Copy(pObj *Gia_Obj_t) int32 { return Abc_LitNotCond( Gia_ObjFanin0(pObj).Value, Gia_ObjFaninC0(pObj) );     }
func Gia_ObjFanin1Copy(pObj *Gia_Obj_t) int32 { return Abc_LitNotCond( Gia_ObjFanin1(pObj).Value, Gia_ObjFaninC1(pObj) );     }
func Gia_ObjFanin2Copy(p *Gia_Man_t, pObj *Gia_Obj_t) int32 { return Abc_LitNotCond(Gia_ObjFanin2(p, pObj).Value, Gia_ObjFaninC2(p, pObj)); }

func Gia_ObjCopyF(p *Gia_Man_t, f int32, pObj *Gia_Obj_t) int32 { return Vec_IntEntry(&p.vCopies, Gia_ManObjNum(p) * f + Gia_ObjId(p,pObj));      }
func Gia_ObjSetCopyF(p *Gia_Man_t, f int32, pObj *Gia_Obj_t, iLit int32) { Vec_IntWriteEntry(&p.vCopies, Gia_ManObjNum(p) * f + Gia_ObjId(p,pObj), iLit);  }
func Gia_ObjCopyArray(p *Gia_Man_t, iObj int32) int32 { return Vec_IntEntry(&p.vCopies, iObj);                                          }
func Gia_ObjSetCopyArray(p *Gia_Man_t, iObj int32, iLit int32) { Vec_IntWriteEntry(&p.vCopies, iObj, iLit);                                      }
func Gia_ManCleanCopyArray(p *Gia_Man_t) { Vec_IntFill( &p.vCopies, Gia_ManObjNum(p), -1 );                                }

func Gia_ObjCopy2Array(p *Gia_Man_t, iObj int32) int32 { return Vec_IntEntry(&p.vCopies2, iObj);                                          }
func Gia_ObjSetCopy2Array(p *Gia_Man_t, iObj int32, iLit int32) { Vec_IntWriteEntry(&p.vCopies2, iObj, iLit);                                      }
func Gia_ManCleanCopy2Array(p *Gia_Man_t) { Vec_IntFill( &p.vCopies2, Gia_ManObjNum(p), -1 );                                }

func Gia_ObjFanin0CopyF(p *Gia_Man_t, f int32, pObj *Gia_Obj_t) int32 { return Abc_LitNotCond(Gia_ObjCopyF(p, f, Gia_ObjFanin0(pObj)), Gia_ObjFaninC0(pObj));   }
func Gia_ObjFanin1CopyF(p *Gia_Man_t, f int32, pObj *Gia_Obj_t) int32 { return Abc_LitNotCond(Gia_ObjCopyF(p, f, Gia_ObjFanin1(pObj)), Gia_ObjFaninC1(pObj));   }
func Gia_ObjFanin0CopyArray(p *Gia_Man_t, pObj *Gia_Obj_t) int32 { return Abc_LitNotCond(Gia_ObjCopyArray(p, Gia_ObjFaninId0p(p,pObj)), Gia_ObjFaninC0(pObj));  }
func Gia_ObjFanin1CopyArray(p *Gia_Man_t, pObj *Gia_Obj_t) int32 { return Abc_LitNotCond(Gia_ObjCopyArray(p, Gia_ObjFaninId1p(p,pObj)), Gia_ObjFaninC1(pObj));  }

func Gia_ObjFromLit(p *Gia_Man_t, iLit int32) *Gia_Obj_t { return Gia_NotCond( Gia_ManObj(p, Abc_Lit2Var(iLit)), Abc_LitIsCompl(iLit) );  }
func Gia_ObjToLit(p *Gia_Man_t, pObj *Gia_Obj_t) int32 { return Abc_Var2Lit( Gia_ObjId(p, Gia_Regular(pObj)), Gia_IsComplement(pObj) ); }
func Gia_ObjPhaseRealLit(p *Gia_Man_t, iLit int32) int32 { return Gia_ObjPhaseReal( Gia_ObjFromLit(p, iLit) );                            }

func Gia_ObjLevelId(p *Gia_Man_t, Id int32) int32 { return Vec_IntGetEntry(p.vLevels, Id);                    }
func Gia_ObjLevel(p *Gia_Man_t, pObj *Gia_Obj_t) int32 { return Gia_ObjLevelId( p, Gia_ObjId(p,pObj) );             }
func Gia_ObjUpdateLevelId(p *Gia_Man_t, Id int32, l int32) { Vec_IntSetEntry(p.vLevels, Id, Abc_MaxInt(Vec_IntEntry(p.vLevels, Id), l));                        }
func Gia_ObjSetLevelId(p *Gia_Man_t, Id int32, l int32) { Vec_IntSetEntry(p.vLevels, Id, l);                        }
func Gia_ObjSetLevel(p *Gia_Man_t, pObj *Gia_Obj_t, l int32) { Gia_ObjSetLevelId( p, Gia_ObjId(p,pObj), l );       }
func Gia_ObjSetCoLevel(p *Gia_Man_t, pObj *Gia_Obj_t) { assert((Gia_ObjIsCo(pObj) != 0), "Gia_ObjIsCo(pObj)"); Gia_ObjSetLevel( p, pObj, Gia_ObjLevel(p,Gia_ObjFanin0(pObj)) );                                                }
func Gia_ObjSetBufLevel(p *Gia_Man_t, pObj *Gia_Obj_t) { assert((Gia_ObjIsAnd(pObj) != 0), "Gia_ObjIsAnd(pObj)"); Gia_ObjSetLevel( p, pObj, Gia_ObjLevel(p,Gia_ObjFanin0(pObj)) );                                                }
func Gia_ObjSetAndLevel(p *Gia_Man_t, pObj *Gia_Obj_t) { assert((Gia_ObjIsAnd(pObj) != 0), "Gia_ObjIsAnd(pObj)"); Gia_ObjSetLevel( p, pObj, 1+Abc_MaxInt(Gia_ObjLevel(p,Gia_ObjFanin0(pObj)),Gia_ObjLevel(p,Gia_ObjFanin1(pObj))) ); }
func Gia_ObjSetXorLevel(p *Gia_Man_t, pObj *Gia_Obj_t) { assert((Gia_ObjIsXor(pObj) != 0), "Gia_ObjIsXor(pObj)"); Gia_ObjSetLevel( p, pObj, 2+Abc_MaxInt(Gia_ObjLevel(p,Gia_ObjFanin0(pObj)),Gia_ObjLevel(p,Gia_ObjFanin1(pObj))) ); }
func Gia_ObjSetMuxLevel(p *Gia_Man_t, pObj *Gia_Obj_t) { assert((Gia_ObjIsMux(p,pObj) != 0), "Gia_ObjIsMux(p,pObj)"); Gia_ObjSetLevel( p, pObj, 2+Abc_MaxInt( Abc_MaxInt(Gia_ObjLevel(p,Gia_ObjFanin0(pObj)),Gia_ObjLevel(p,Gia_ObjFanin1(pObj))), Gia_ObjLevel(p,Gia_ObjFanin2(p,pObj))) ); }
func Gia_ObjSetGateLevel(p *Gia_Man_t, pObj *Gia_Obj_t) { if !p.fGiaSimple && (Gia_ObjIsBuf(pObj) != 0) { Gia_ObjSetBufLevel(p, pObj); } else if Gia_ObjIsMux(p,pObj) != 0 { Gia_ObjSetMuxLevel(p, pObj); } else if Gia_ObjIsXor(pObj) != 0 { Gia_ObjSetXorLevel(p, pObj); } else if Gia_ObjIsAnd(pObj) != 0 { Gia_ObjSetAndLevel(p, pObj); } }

func Gia_ObjHasNumId(p *Gia_Man_t, Id int32) int32 { return Vec_IntEntry(p.vTtNums, Id) > -ABC_INFINITY;     }
func Gia_ObjNumId(p *Gia_Man_t, Id int32) int32 { return Vec_IntEntry(p.vTtNums, Id);                     }
func Gia_ObjNum(p *Gia_Man_t, pObj *Gia_Obj_t) int32 { return Vec_IntEntry(p.vTtNums, Gia_ObjId(p,pObj));      }
func Gia_ObjSetNumId(p *Gia_Man_t, Id int32, n int32) { Vec_IntWriteEntry(p.vTtNums, Id, n);                    }
func Gia_ObjSetNum(p *Gia_Man_t, pObj *Gia_Obj_t, n int32) { Vec_IntWriteEntry(p.vTtNums, Gia_ObjId(p,pObj), n);     }
func Gia_ObjResetNumId(p *Gia_Man_t, Id int32) { Vec_IntWriteEntry(p.vTtNums, Id, -ABC_INFINITY);        }

func Gia_ObjRefNumId(p *Gia_Man_t, Id int32) int32 { return p.pRefs[Id];                              }
func Gia_ObjRefIncId(p *Gia_Man_t, Id int32) int32 { return p.pRefs[Id]++;                            }
func Gia_ObjRefDecId(p *Gia_Man_t, Id int32) int32 { return --p.pRefs[Id];                            }
func Gia_ObjRefNum(p *Gia_Man_t, pObj *Gia_Obj_t) int32 { return Gia_ObjRefNumId( p, Gia_ObjId(p, pObj) );  }
func Gia_ObjRefInc(p *Gia_Man_t, pObj *Gia_Obj_t) int32 { return Gia_ObjRefIncId( p, Gia_ObjId(p, pObj) );  }
func Gia_ObjRefDec(p *Gia_Man_t, pObj *Gia_Obj_t) int32 { return Gia_ObjRefDecId( p, Gia_ObjId(p, pObj) );  }
func Gia_ObjRefFanin0Inc(p *Gia_Man_t, pObj *Gia_Obj_t) { Gia_ObjRefInc(p, Gia_ObjFanin0(pObj));            }
func Gia_ObjRefFanin1Inc(p *Gia_Man_t, pObj *Gia_Obj_t) { Gia_ObjRefInc(p, Gia_ObjFanin1(pObj));            }
func Gia_ObjRefFanin2Inc(p *Gia_Man_t, pObj *Gia_Obj_t) { Gia_ObjRefInc(p, Gia_ObjFanin2(p, pObj));         }
func Gia_ObjRefFanin0Dec(p *Gia_Man_t, pObj *Gia_Obj_t) { Gia_ObjRefDec(p, Gia_ObjFanin0(pObj));            }
func Gia_ObjRefFanin1Dec(p *Gia_Man_t, pObj *Gia_Obj_t) { Gia_ObjRefDec(p, Gia_ObjFanin1(pObj));            }
func Gia_ObjRefFanin2Dec(p *Gia_Man_t, pObj *Gia_Obj_t) { Gia_ObjRefDec(p, Gia_ObjFanin2(p, pObj));         }

func Gia_ObjLutRefNumId(p *Gia_Man_t, Id int32) int32 { assert(p.pLutRefs, "p.pLutRefs"); return p.pLutRefs[Id];                    }
func Gia_ObjLutRefIncId(p *Gia_Man_t, Id int32) int32 { assert(p.pLutRefs, "p.pLutRefs"); return p.pLutRefs[Id]++;                  }
func Gia_ObjLutRefDecId(p *Gia_Man_t, Id int32) int32 { assert(p.pLutRefs, "p.pLutRefs"); return --p.pLutRefs[Id];                  }
func Gia_ObjLutRefNum(p *Gia_Man_t, pObj *Gia_Obj_t) int32 { assert(p.pLutRefs, "p.pLutRefs"); return p.pLutRefs[Gia_ObjId(p, pObj)];    }
func Gia_ObjLutRefInc(p *Gia_Man_t, pObj *Gia_Obj_t) int32 { assert(p.pLutRefs, "p.pLutRefs"); return p.pLutRefs[Gia_ObjId(p, pObj)]++;  }
func Gia_ObjLutRefDec(p *Gia_Man_t, pObj *Gia_Obj_t) int32 { assert(p.pLutRefs, "p.pLutRefs"); return --p.pLutRefs[Gia_ObjId(p, pObj)];  }

func Gia_ObjSetTravIdCurrent(p *Gia_Man_t, pObj *Gia_Obj_t) { assert(Gia_ObjId(p, pObj) < p.nTravIdsAlloc, "Gia_ObjId(p, pObj) < p.nTravIdsAlloc"); p.pTravIds[Gia_ObjId(p, pObj)] = p.nTravIds;                    }
func Gia_ObjSetTravIdPrevious(p *Gia_Man_t, pObj *Gia_Obj_t) { assert(Gia_ObjId(p, pObj) < p.nTravIdsAlloc, "Gia_ObjId(p, pObj) < p.nTravIdsAlloc"); p.pTravIds[Gia_ObjId(p, pObj)] = p.nTravIds - 1;                }
func Gia_ObjIsTravIdCurrent(p *Gia_Man_t, pObj *Gia_Obj_t) int32 { assert(Gia_ObjId(p, pObj) < p.nTravIdsAlloc, "Gia_ObjId(p, pObj) < p.nTravIdsAlloc"); return (p.pTravIds[Gia_ObjId(p, pObj)] == p.nTravIds);          }
func Gia_ObjIsTravIdPrevious(p *Gia_Man_t, pObj *Gia_Obj_t) int32 { assert(Gia_ObjId(p, pObj) < p.nTravIdsAlloc, "Gia_ObjId(p, pObj) < p.nTravIdsAlloc"); return (p.pTravIds[Gia_ObjId(p, pObj)] == p.nTravIds - 1);      }
func Gia_ObjSetTravIdCurrentId(p *Gia_Man_t, Id int32) { assert(Id < p.nTravIdsAlloc, "Id < p.nTravIdsAlloc"); p.pTravIds[Id] = p.nTravIds;                     }
func Gia_ObjSetTravIdPreviousId(p *Gia_Man_t, Id int32) { assert(Id < p.nTravIdsAlloc, "Id < p.nTravIdsAlloc"); p.pTravIds[Id] = p.nTravIds - 1;                 }
func Gia_ObjIsTravIdCurrentId(p *Gia_Man_t, Id int32) int32 { assert(Id < p.nTravIdsAlloc, "Id < p.nTravIdsAlloc"); return (p.pTravIds[Id] == p.nTravIds);           }
func Gia_ObjIsTravIdPreviousId(p *Gia_Man_t, Id int32) int32 { assert(Id < p.nTravIdsAlloc, "Id < p.nTravIdsAlloc"); return (p.pTravIds[Id] == p.nTravIds - 1);       }

func Gia_ManTimeClean(p *Gia_Man_t) { var i int32; assert(p.vTiming != nil, "p.vTiming != nil"); Vec_FltFill(p.vTiming, 3*Gia_ManObjNum(p), 0); for  i = 0; i < Gia_ManObjNum(p); i++  {  Vec_FltWriteEntry( p.vTiming, 3*i+1, float32((ABC_INFINITY)) ); } }
func Gia_ManTimeStart(p *Gia_Man_t) { assert(p.vTiming == nil, "p.vTiming == nil"); p.vTiming = Vec_FltAlloc(0); Gia_ManTimeClean( p );  }
func Gia_ManTimeStop(p *Gia_Man_t) { assert(p.vTiming != nil, "p.vTiming != nil"); Vec_FltFreeP(&p.vTiming);                            }
func Gia_ObjTimeArrival(p *Gia_Man_t, Id int32) float32 { return Vec_FltEntry(p.vTiming, 3*Id+0);                                            }
func Gia_ObjTimeRequired(p *Gia_Man_t, Id int32) float32 { return Vec_FltEntry(p.vTiming, 3*Id+1);                                            }
func Gia_ObjTimeSlack(p *Gia_Man_t, Id int32) float32 { return Vec_FltEntry(p.vTiming, 3*Id+2);                                            }
func Gia_ObjTimeArrivalObj(p *Gia_Man_t, pObj *Gia_Obj_t) float32 { return Gia_ObjTimeArrival( p, Gia_ObjId(p, pObj) );                                 }
func Gia_ObjTimeRequiredObj(p *Gia_Man_t, pObj *Gia_Obj_t) float32 { return Gia_ObjTimeRequired( p, Gia_ObjId(p, pObj) );                                }
func Gia_ObjTimeSlackObj(p *Gia_Man_t, pObj *Gia_Obj_t) float32 { return Gia_ObjTimeSlack( p, Gia_ObjId(p, pObj) );                                   }
func Gia_ObjSetTimeArrival(p *Gia_Man_t, Id int32, t float32) { Vec_FltWriteEntry( p.vTiming, 3*Id+0, t );                                         }
func Gia_ObjSetTimeRequired(p *Gia_Man_t, Id int32, t float32) { Vec_FltWriteEntry( p.vTiming, 3*Id+1, t );                                         }
func Gia_ObjSetTimeSlack(p *Gia_Man_t, Id int32, t float32) { Vec_FltWriteEntry( p.vTiming, 3*Id+2, t );                                         }
func Gia_ObjSetTimeArrivalObj(p *Gia_Man_t, pObj *Gia_Obj_t, t float32) { Gia_ObjSetTimeArrival( p, Gia_ObjId(p, pObj), t );                               }
func Gia_ObjSetTimeRequiredObj(p *Gia_Man_t, pObj *Gia_Obj_t, t float32) { Gia_ObjSetTimeRequired( p, Gia_ObjId(p, pObj), t );                              }
func Gia_ObjSetTimeSlackObj(p *Gia_Man_t, pObj *Gia_Obj_t, t float32) { Gia_ObjSetTimeSlack( p, Gia_ObjId(p, pObj), t );                                 }

func Gia_ObjSimWords(p *Gia_Man_t) int32 { return Vec_WrdSize( p.vSimsPi ) / Gia_ManPiNum( p );          }
func Gia_ObjSimPi(p *Gia_Man_t, PiId int32) *word { return Vec_WrdEntryP( p.vSimsPi, PiId * Gia_ObjSimWords(p) ); }
func Gia_ObjSim(p *Gia_Man_t, Id int32) *word { return Vec_WrdEntryP( p.vSims, Id * Gia_ObjSimWords(p) );     }
func Gia_ObjSimObj(p *Gia_Man_t, pObj *Gia_Obj_t) *word { return Gia_ObjSim( p, Gia_ObjId(p, pObj) );                    }

// AIG construction
// extern void Gia_ObjAddFanout( Gia_Man_t * p, Gia_Obj_t * pObj, Gia_Obj_t * pFanout );

func Gia_ManAppendObj(p *Gia_Man_t) *Gia_Obj_t { 
    if p.nObjs == p.nObjsAlloc {
        var nObjNew int32  = Abc_MinInt( 2 * p.nObjsAlloc, (1 << 29) );
        if p.nObjs == (1 << 29) {
            fmt.Printf("Hard limit on the number of nodes (2^29) is reached. Quitting...\n");
os.Exit(int(1)); }
        assert(p.nObjs < nObjNew, "p.nObjs < nObjNew");
        if p.fVerbose {
            fmt.Printf("Extending GIA object storage: %d -> %d.\n", p.nObjsAlloc, nObjNew); }
        assert(p.nObjsAlloc > 0, "p.nObjsAlloc > 0");
        // inlining of macro ABC_REALLOC
        
        if p.pObjs {
            p.pObjs = ((Gia_Obj_t *) realloc((char *)(p.pObjs), sizeof(Gia_Obj_t) * (size_t)(nObjNew)));
        } else {
            p.pObjs = ((Gia_Obj_t *) malloc(sizeof(Gia_Obj_t) * (size_t)(nObjNew)));
        }
        memset( p.pObjs + p.nObjsAlloc, 0, sizeof(Gia_Obj_t) * (nObjNew - p.nObjsAlloc) );
        if p.pMuxes {
            // inlining of macro ABC_REALLOC
            
            if p.pMuxes {
                p.pMuxes = ((unsigned *) realloc((char *)(p.pMuxes), sizeof(unsigned) * (size_t)(nObjNew)));
            } else {
                p.pMuxes = ((unsigned *) malloc(sizeof(unsigned) * (size_t)(nObjNew)));
            }
            memset( p.pMuxes + p.nObjsAlloc, 0, sizeof(unsigned) * (nObjNew - p.nObjsAlloc) );
        }
        p.nObjsAlloc = nObjNew;
    }
    if Vec_IntSize(&p.vHTable) { Vec_IntPush( &p.vHash, 0 ); }
    return Gia_ManObj( p, p.nObjs++ );
}
func Gia_ManAppendCi(p *Gia_Man_t) int32 { 
    var pObj *Gia_Obj_t  = Gia_ManAppendObj( p );
    pObj.fTerm = 1;
    pObj.iDiff0 = GIA_NONE;
    pObj.iDiff1 = Vec_IntSize( p.vCis );
    Vec_IntPush( p.vCis, Gia_ObjId(p, pObj) );
    return Gia_ObjId( p, pObj ) << 1;
}

// extern void Gia_ManQuantSetSuppAnd( Gia_Man_t * p, Gia_Obj_t * pObj );

// extern void Gia_ManBuiltInSimPerform( Gia_Man_t * p, int iObj );


func Gia_ManAppendAnd(p *Gia_Man_t, iLit0 int32, iLit1 int32) int32 { 
    var pObj *Gia_Obj_t  = Gia_ManAppendObj( p );
    assert(iLit0 >= 0 && Abc_Lit2Var(iLit0) < Gia_ManObjNum(p), "iLit0 >= 0 && Abc_Lit2Var(iLit0) < Gia_ManObjNum(p)");
    assert(iLit1 >= 0 && Abc_Lit2Var(iLit1) < Gia_ManObjNum(p), "iLit1 >= 0 && Abc_Lit2Var(iLit1) < Gia_ManObjNum(p)");
    assert(p.fGiaSimple || Abc_Lit2Var(iLit0) != Abc_Lit2Var(iLit1), "p.fGiaSimple || Abc_Lit2Var(iLit0) != Abc_Lit2Var(iLit1)");
    if iLit0 < iLit1 {
        pObj.iDiff0  = uint32((Gia_ObjId(p, pObj) - Abc_Lit2Var(iLit0)));
        pObj.fCompl0 = uint32((Abc_LitIsCompl(iLit0)));
        pObj.iDiff1  = uint32((Gia_ObjId(p, pObj) - Abc_Lit2Var(iLit1)));
        pObj.fCompl1 = uint32((Abc_LitIsCompl(iLit1)));
    
    } else {
        pObj.iDiff1  = uint32((Gia_ObjId(p, pObj) - Abc_Lit2Var(iLit0)));
        pObj.fCompl1 = uint32((Abc_LitIsCompl(iLit0)));
        pObj.iDiff0  = uint32((Gia_ObjId(p, pObj) - Abc_Lit2Var(iLit1)));
        pObj.fCompl0 = uint32((Abc_LitIsCompl(iLit1)));
    }
    if p.pFanData {
        Gia_ObjAddFanout( p, Gia_ObjFanin0(pObj), pObj );
        Gia_ObjAddFanout( p, Gia_ObjFanin1(pObj), pObj );
    }
    if p.fSweeper {
        var pFan0 *Gia_Obj_t  = Gia_ObjFanin0(pObj);
        var pFan1 *Gia_Obj_t  = Gia_ObjFanin1(pObj);
        if pFan0.fMark0 { pFan0.fMark1 = 1; } else { pFan0.fMark0 = 1; }
        if pFan1.fMark0 { pFan1.fMark1 = 1; } else { pFan1.fMark0 = 1; }
        pObj.fPhase = (Gia_ObjPhase(pFan0) ^ Gia_ObjFaninC0(pObj)) & (Gia_ObjPhase(pFan1) ^ Gia_ObjFaninC1(pObj));
    }
    if p.fBuiltInSim {
        var pFan0 *Gia_Obj_t  = Gia_ObjFanin0(pObj);
        var pFan1 *Gia_Obj_t  = Gia_ObjFanin1(pObj);
        pObj.fPhase = (Gia_ObjPhase(pFan0) ^ Gia_ObjFaninC0(pObj)) & (Gia_ObjPhase(pFan1) ^ Gia_ObjFaninC1(pObj));
        Gia_ManBuiltInSimPerform( p, Gia_ObjId( p, pObj ) );
    }
    if p.vSuppWords {
        Gia_ManQuantSetSuppAnd( p, pObj ); }
    return Gia_ObjId( p, pObj ) << 1;
}
func Gia_ManAppendXorReal(p *Gia_Man_t, iLit0 int32, iLit1 int32) int32 { 
    var pObj *Gia_Obj_t  = Gia_ManAppendObj( p );
    assert(iLit0 >= 0 && Abc_Lit2Var(iLit0) < Gia_ManObjNum(p), "iLit0 >= 0 && Abc_Lit2Var(iLit0) < Gia_ManObjNum(p)");
    assert(iLit1 >= 0 && Abc_Lit2Var(iLit1) < Gia_ManObjNum(p), "iLit1 >= 0 && Abc_Lit2Var(iLit1) < Gia_ManObjNum(p)");
    assert(Abc_Lit2Var(iLit0) != Abc_Lit2Var(iLit1), "Abc_Lit2Var(iLit0) != Abc_Lit2Var(iLit1)");
    //assert( !Abc_LitIsCompl(iLit0) );
    //assert( !Abc_LitIsCompl(iLit1) );
    if Abc_Lit2Var(iLit0) > Abc_Lit2Var(iLit1) {
        pObj.iDiff0  = uint32((Gia_ObjId(p, pObj) - Abc_Lit2Var(iLit0)));
        pObj.fCompl0 = uint32((Abc_LitIsCompl(iLit0)));
        pObj.iDiff1  = uint32((Gia_ObjId(p, pObj) - Abc_Lit2Var(iLit1)));
        pObj.fCompl1 = uint32((Abc_LitIsCompl(iLit1)));
    
    } else {
        pObj.iDiff1  = uint32((Gia_ObjId(p, pObj) - Abc_Lit2Var(iLit0)));
        pObj.fCompl1 = uint32((Abc_LitIsCompl(iLit0)));
        pObj.iDiff0  = uint32((Gia_ObjId(p, pObj) - Abc_Lit2Var(iLit1)));
        pObj.fCompl0 = uint32((Abc_LitIsCompl(iLit1)));
    }
    p.nXors++;
    return Gia_ObjId( p, pObj ) << 1;
}
func Gia_ManAppendMuxReal(p *Gia_Man_t, iLitC int32, iLit1 int32, iLit0 int32) int32 { 
    var pObj *Gia_Obj_t  = Gia_ManAppendObj( p );
    assert(p.pMuxes != nil, "p.pMuxes != nil");
    assert(iLit0 >= 0 && Abc_Lit2Var(iLit0) < Gia_ManObjNum(p), "iLit0 >= 0 && Abc_Lit2Var(iLit0) < Gia_ManObjNum(p)");
    assert(iLit1 >= 0 && Abc_Lit2Var(iLit1) < Gia_ManObjNum(p), "iLit1 >= 0 && Abc_Lit2Var(iLit1) < Gia_ManObjNum(p)");
    assert(iLitC >= 0 && Abc_Lit2Var(iLitC) < Gia_ManObjNum(p), "iLitC >= 0 && Abc_Lit2Var(iLitC) < Gia_ManObjNum(p)");
    assert(Abc_Lit2Var(iLit0) != Abc_Lit2Var(iLit1), "Abc_Lit2Var(iLit0) != Abc_Lit2Var(iLit1)");
    assert(Abc_Lit2Var(iLitC) != Abc_Lit2Var(iLit0), "Abc_Lit2Var(iLitC) != Abc_Lit2Var(iLit0)");
    assert(Abc_Lit2Var(iLitC) != Abc_Lit2Var(iLit1), "Abc_Lit2Var(iLitC) != Abc_Lit2Var(iLit1)");
    assert(!Vec_IntSize(&p.vHTable) || !Abc_LitIsCompl(iLit1), "!Vec_IntSize(&p.vHTable) || !Abc_LitIsCompl(iLit1)");
    if Abc_Lit2Var(iLit0) < Abc_Lit2Var(iLit1) {
        pObj.iDiff0  = uint32((Gia_ObjId(p, pObj) - Abc_Lit2Var(iLit0)));
        pObj.fCompl0 = uint32((Abc_LitIsCompl(iLit0)));
        pObj.iDiff1  = uint32((Gia_ObjId(p, pObj) - Abc_Lit2Var(iLit1)));
        pObj.fCompl1 = uint32((Abc_LitIsCompl(iLit1)));
        p.pMuxes[Gia_ObjId(p, pObj)] = iLitC;
    
    } else {
        pObj.iDiff1  = uint32((Gia_ObjId(p, pObj) - Abc_Lit2Var(iLit0)));
        pObj.fCompl1 = uint32((Abc_LitIsCompl(iLit0)));
        pObj.iDiff0  = uint32((Gia_ObjId(p, pObj) - Abc_Lit2Var(iLit1)));
        pObj.fCompl0 = uint32((Abc_LitIsCompl(iLit1)));
        p.pMuxes[Gia_ObjId(p, pObj)] = Abc_LitNot(iLitC);
    }
    p.nMuxes++;
    return Gia_ObjId( p, pObj ) << 1;
}
func Gia_ManAppendBuf(p *Gia_Man_t, iLit int32) int32 { 
    var pObj *Gia_Obj_t  = Gia_ManAppendObj( p );
    assert(iLit >= 0 && Abc_Lit2Var(iLit) < Gia_ManObjNum(p), "iLit >= 0 && Abc_Lit2Var(iLit) < Gia_ManObjNum(p)");
    pObj.iDiff0  = pObj.iDiff1  = Gia_ObjId(p, pObj) - Abc_Lit2Var(iLit);
    pObj.fCompl0 = pObj.fCompl1 = Abc_LitIsCompl(iLit);
    p.nBufs++;
    return Gia_ObjId( p, pObj ) << 1;
}
func Gia_ManAppendCo(p *Gia_Man_t, iLit0 int32) int32 { 
    var pObj *Gia_Obj_t;
    assert(iLit0 >= 0 && Abc_Lit2Var(iLit0) < Gia_ManObjNum(p), "iLit0 >= 0 && Abc_Lit2Var(iLit0) < Gia_ManObjNum(p)");
    assert(!(Gia_ObjIsCo(Gia_ManObj(p, Abc_Lit2Var(iLit0))) != 0), "!Gia_ObjIsCo(Gia_ManObj(p, Abc_Lit2Var(iLit0)))");
    pObj = Gia_ManAppendObj( p );    
    pObj.fTerm = 1;
    pObj.iDiff0  = Gia_ObjId(p, pObj) - Abc_Lit2Var(iLit0);
    pObj.fCompl0 = Abc_LitIsCompl(iLit0);
    pObj.iDiff1  = Vec_IntSize( p.vCos );
    Vec_IntPush( p.vCos, Gia_ObjId(p, pObj) );
    if p.pFanData {
        Gia_ObjAddFanout( p, Gia_ObjFanin0(pObj), pObj ); }
    return Gia_ObjId( p, pObj ) << 1;
}
func Gia_ManAppendOr(p *Gia_Man_t, iLit0 int32, iLit1 int32) int32 {
    return Abc_LitNot(Gia_ManAppendAnd( p, Abc_LitNot(iLit0), Abc_LitNot(iLit1) ));
}
func Gia_ManAppendMux(p *Gia_Man_t, iCtrl int32, iData1 int32, iData0 int32) int32 { 
    var iTemp0 int32  = Gia_ManAppendAnd( p, Abc_LitNot(iCtrl), iData0 );
    var iTemp1 int32  = Gia_ManAppendAnd( p, iCtrl, iData1 );
    return Abc_LitNotCond( Gia_ManAppendAnd( p, Abc_LitNot(iTemp0), Abc_LitNot(iTemp1) ), 1 );
}
func Gia_ManAppendMaj(p *Gia_Man_t, iData0 int32, iData1 int32, iData2 int32) int32 { 
    var iTemp0 int32  = Gia_ManAppendOr( p, iData1, iData2 );
    var iTemp1 int32  = Gia_ManAppendAnd( p, iData0, iTemp0 );
    var iTemp2 int32  = Gia_ManAppendAnd( p, iData1, iData2 );
    return Gia_ManAppendOr( p, iTemp1, iTemp2 );
}
func Gia_ManAppendXor(p *Gia_Man_t, iLit0 int32, iLit1 int32) int32 { 
    return Gia_ManAppendMux( p, iLit0, Abc_LitNot(iLit1), iLit1 );
}

func Gia_ManAppendAnd2(p *Gia_Man_t, iLit0 int32, iLit1 int32) int32 { 
    if !p.fGiaSimple {
        if iLit0 < 2 {
            if iLit0 != 0 {
                return iLit1;
            } else {
                return 0;
            } }
        if iLit1 < 2 {
            if iLit1 != 0 {
                return iLit0;
            } else {
                return 0;
            } }
        if iLit0 == iLit1 {
            return iLit1; }
        if iLit0 == Abc_LitNot(iLit1) {
            return 0; }
    }
    return Gia_ManAppendAnd( p, iLit0, iLit1 );
}
func Gia_ManAppendOr2(p *Gia_Man_t, iLit0 int32, iLit1 int32) int32 {
    return Abc_LitNot(Gia_ManAppendAnd2( p, Abc_LitNot(iLit0), Abc_LitNot(iLit1) ));
}
func Gia_ManAppendMux2(p *Gia_Man_t, iCtrl int32, iData1 int32, iData0 int32) int32 { 
    var iTemp0 int32  = Gia_ManAppendAnd2( p, Abc_LitNot(iCtrl), iData0 );
    var iTemp1 int32  = Gia_ManAppendAnd2( p, iCtrl, iData1 );
    return Abc_LitNotCond( Gia_ManAppendAnd2( p, Abc_LitNot(iTemp0), Abc_LitNot(iTemp1) ), 1 );
}
func Gia_ManAppendMaj2(p *Gia_Man_t, iData0 int32, iData1 int32, iData2 int32) int32 { 
    var iTemp0 int32  = Gia_ManAppendOr2( p, iData1, iData2 );
    var iTemp1 int32  = Gia_ManAppendAnd2( p, iData0, iTemp0 );
    var iTemp2 int32  = Gia_ManAppendAnd2( p, iData1, iData2 );
    return Gia_ManAppendOr2( p, iTemp1, iTemp2 );
}
func Gia_ManAppendXor2(p *Gia_Man_t, iLit0 int32, iLit1 int32) int32 { 
    return Gia_ManAppendMux2( p, iLit0, Abc_LitNot(iLit1), iLit1 );
}

func Gia_ManAppendXorReal2(p *Gia_Man_t, iLit0 int32, iLit1 int32) int32 { 
    if !p.fGiaSimple {
        if iLit0 < 2 {
            if iLit0 != 0 {
                return Abc_LitNot(iLit1);
            } else {
                return iLit1;
            } }
        if iLit1 < 2 {
            if iLit1 != 0 {
                return Abc_LitNot(iLit0);
            } else {
                return iLit0;
            } }
        if iLit0 == iLit1 {
            return 0; }
        if iLit0 == Abc_LitNot(iLit1) {
            return 1; }
    }
    return Gia_ManAppendXorReal( p, iLit0, iLit1 );
}

func Gia_ManPatchCoDriver(p *Gia_Man_t, iCoIndex int32, iLit0 int32) {
    var pObjCo *Gia_Obj_t   = Gia_ManCo( p, iCoIndex );
    assert(Gia_ObjId(p, pObjCo) > Abc_Lit2Var(iLit0), "Gia_ObjId(p, pObjCo) > Abc_Lit2Var(iLit0)");
    pObjCo.iDiff0  = Gia_ObjId(p, pObjCo) - Abc_Lit2Var(iLit0);
    pObjCo.fCompl0 = Abc_LitIsCompl(iLit0);
}

const GIA_ZER = 1
const GIA_ONE = 2
const GIA_UND = 3

func Gia_XsimNotCond(Value int32, fCompl int32) int32 { 
    if Value == GIA_UND {
        return GIA_UND; }
    if Value == GIA_ZER + fCompl {
        return GIA_ZER; }
    return GIA_ONE;
}
func Gia_XsimAndCond(Value0 int32, fCompl0 int32, Value1 int32, fCompl1 int32) int32 { 
    if Value0 == GIA_ZER + fCompl0 || Value1 == GIA_ZER + fCompl1 {
        return GIA_ZER; }
    if Value0 == GIA_UND || Value1 == GIA_UND {
        return GIA_UND; }
    return GIA_ONE;
}


func Gia_ObjTerSimSetC(pObj *Gia_Obj_t) { pObj.fMark0 = 0; pObj.fMark1 = 0;    }
func Gia_ObjTerSimSet0(pObj *Gia_Obj_t) { pObj.fMark0 = 1; pObj.fMark1 = 0;    }
func Gia_ObjTerSimSet1(pObj *Gia_Obj_t) { pObj.fMark0 = 0; pObj.fMark1 = 1;    }
func Gia_ObjTerSimSetX(pObj *Gia_Obj_t) { pObj.fMark0 = 1; pObj.fMark1 = 1;    }

func Gia_ObjTerSimGetC(pObj *Gia_Obj_t) int32 { return !pObj.fMark0 && !pObj.fMark1; }
func Gia_ObjTerSimGet0(pObj *Gia_Obj_t) int32 { return  pObj.fMark0 && !pObj.fMark1; }
func Gia_ObjTerSimGet1(pObj *Gia_Obj_t) int32 { return !pObj.fMark0 &&  pObj.fMark1; }
func Gia_ObjTerSimGetX(pObj *Gia_Obj_t) int32 { return  pObj.fMark0 &&  pObj.fMark1; }

func Gia_ObjTerSimGet0Fanin0(pObj *Gia_Obj_t) int32 { return (Gia_ObjTerSimGet1(Gia_ObjFanin0(pObj)) && Gia_ObjFaninC0(pObj)) || (Gia_ObjTerSimGet0(Gia_ObjFanin0(pObj)) && !Gia_ObjFaninC0(pObj)); }
func Gia_ObjTerSimGet1Fanin0(pObj *Gia_Obj_t) int32 { return (Gia_ObjTerSimGet0(Gia_ObjFanin0(pObj)) && Gia_ObjFaninC0(pObj)) || (Gia_ObjTerSimGet1(Gia_ObjFanin0(pObj)) && !Gia_ObjFaninC0(pObj)); }

func Gia_ObjTerSimGet0Fanin1(pObj *Gia_Obj_t) int32 { return (Gia_ObjTerSimGet1(Gia_ObjFanin1(pObj)) && Gia_ObjFaninC1(pObj)) || (Gia_ObjTerSimGet0(Gia_ObjFanin1(pObj)) && !Gia_ObjFaninC1(pObj)); }
func Gia_ObjTerSimGet1Fanin1(pObj *Gia_Obj_t) int32 { return (Gia_ObjTerSimGet0(Gia_ObjFanin1(pObj)) && Gia_ObjFaninC1(pObj)) || (Gia_ObjTerSimGet1(Gia_ObjFanin1(pObj)) && !Gia_ObjFaninC1(pObj)); }

func Gia_ObjTerSimAnd(pObj *Gia_Obj_t) {
    assert((Gia_ObjIsAnd(pObj) != 0), "Gia_ObjIsAnd(pObj)");
    assert(!(Gia_ObjTerSimGetC( Gia_ObjFanin0(pObj) ) != 0), "!Gia_ObjTerSimGetC( Gia_ObjFanin0(pObj) )");
    assert(!(Gia_ObjTerSimGetC( Gia_ObjFanin1(pObj) ) != 0), "!Gia_ObjTerSimGetC( Gia_ObjFanin1(pObj) )");
    if (Gia_ObjTerSimGet0Fanin0(pObj) != 0) || (Gia_ObjTerSimGet0Fanin1(pObj) != 0) {
        Gia_ObjTerSimSet0( pObj );
    } else if (Gia_ObjTerSimGet1Fanin0(pObj) != 0) && (Gia_ObjTerSimGet1Fanin1(pObj) != 0) {
        Gia_ObjTerSimSet1( pObj );
    } else { 
        Gia_ObjTerSimSetX( pObj ); }
}
func Gia_ObjTerSimCo(pObj *Gia_Obj_t) {
    assert((Gia_ObjIsCo(pObj) != 0), "Gia_ObjIsCo(pObj)");
    assert(!(Gia_ObjTerSimGetC( Gia_ObjFanin0(pObj) ) != 0), "!Gia_ObjTerSimGetC( Gia_ObjFanin0(pObj) )");
    if Gia_ObjTerSimGet0Fanin0(pObj) != 0 {
        Gia_ObjTerSimSet0( pObj );
    } else if Gia_ObjTerSimGet1Fanin0(pObj) != 0 {
        Gia_ObjTerSimSet1( pObj );
    } else { 
        Gia_ObjTerSimSetX( pObj ); }
}
func Gia_ObjTerSimRo(p *Gia_Man_t, pObj *Gia_Obj_t) {
    var pTemp *Gia_Obj_t  = Gia_ObjRoToRi(p, pObj);
    assert((Gia_ObjIsRo(p, pObj) != 0), "Gia_ObjIsRo(p, pObj)");
    assert(!(Gia_ObjTerSimGetC( pTemp ) != 0), "!Gia_ObjTerSimGetC( pTemp )");
    pObj.fMark0 = pTemp.fMark0;
    pObj.fMark1 = pTemp.fMark1;
}

func Gia_ObjTerSimPrint(pObj *Gia_Obj_t) {
    if Gia_ObjTerSimGet0(pObj) != 0 {
        fmt.Printf("0");
    } else if Gia_ObjTerSimGet1(pObj) != 0 {
        fmt.Printf("1");
    } else if Gia_ObjTerSimGetX(pObj) != 0 {
        fmt.Printf("X"); }
}

static inline int Gia_AigerReadInt( unsigned char * pPos )
{
    var i, Value int32;
    for  i = 0; i < 4; i++  {
        Value = (Value << 8) | *pPos++;
    }
    return Value;
}
static inline void Gia_AigerWriteInt( unsigned char * pPos, int Value )
{
    var i int32;
    for  i = 3; i >= 0; i--  {
        *pPos++ = (Value >> (8*i)) & 255;
    }
}
static inline unsigned Gia_AigerReadUnsigned( unsigned char ** ppPos )
{
    var x, i uint32;
    unsigned char ch;
    for (ch = *(*ppPos)++) & 0x80 {
        x |= (ch & 0x7f) << (7 * i++);
    }
    return x | (ch << (7 * i));
}
func Gia_AigerWriteUnsigned(vStr *Vec_Str_t, x uint32) {
    unsigned char ch;
    for x & ~0x7f {
        ch = (x & 0x7f) | 0x80;
        Vec_StrPush( vStr, ch );
        x >>= 7;
    }
    ch = x;
    Vec_StrPush( vStr, ch );
}
func Gia_AigerWriteUnsignedFile(pFile *os.File, x uint32) {
    unsigned char ch;
    for x & ~0x7f {
        ch = (x & 0x7f) | 0x80;
        fputc( ch, pFile );
        x >>= 7;
    }
    ch = x;
    fputc( ch, pFile );
}
static inline int Gia_AigerWriteUnsignedBuffer( unsigned char * pBuffer, int Pos, unsigned x )
{
    unsigned char ch;
    for x & ~0x7f {
        ch = (x & 0x7f) | 0x80;
        pBuffer[Pos++] = ch;
        x >>= 7;
    }
    ch = x;
    pBuffer[Pos++] = ch;
    return Pos;
}

func Gia_ObjReprObj(p *Gia_Man_t, Id int32) *Gia_Obj_t { if p.pReprs[Id].iRepr == GIA_VOID {
    return nil;
} else {
    return Gia_ManObj( p, p.pReprs[Id].iRepr );
}                  }
func Gia_ObjRepr(p *Gia_Man_t, Id int32) int32 { return p.pReprs[Id].iRepr;                                                }
func Gia_ObjSetRepr(p *Gia_Man_t, Id int32, Num int32) { assert(Num == GIA_VOID || Num < Id, "Num == GIA_VOID || Num < Id"); p.pReprs[Id].iRepr = Num;          }
func Gia_ObjSetReprRev(p *Gia_Man_t, Id int32, Num int32) { assert(Num == GIA_VOID || Num > Id, "Num == GIA_VOID || Num > Id"); p.pReprs[Id].iRepr = Num;          }
func Gia_ObjUnsetRepr(p *Gia_Man_t, Id int32) { p.pReprs[Id].iRepr = GIA_VOID;                                            }
func Gia_ObjHasRepr(p *Gia_Man_t, Id int32) int32 { return p.pReprs[Id].iRepr != GIA_VOID;                                    }
func Gia_ObjReprSelf(p *Gia_Man_t, Id int32) int32 { if Gia_ObjHasRepr(p, Id) != 0 {
    return Gia_ObjRepr(p, Id);
} else {
    return Id;
}                    }
func Gia_ObjSibl(p *Gia_Man_t, Id int32) int32 { if p.pSibls {
    return p.pSibls[Id];
} else {
    return 0;
}                                      }
func Gia_ObjSiblObj(p *Gia_Man_t, Id int32) *Gia_Obj_t { if p.pSibls && p.pSibls[Id] {
    return Gia_ManObj(p, p.pSibls[Id]);
} else {
    return nil;
} }

func Gia_ObjProved(p *Gia_Man_t, Id int32) int32 { return p.pReprs[Id].fProved;       }
func Gia_ObjSetProved(p *Gia_Man_t, Id int32) { p.pReprs[Id].fProved = 1;          }
func Gia_ObjUnsetProved(p *Gia_Man_t, Id int32) { p.pReprs[Id].fProved = 0;          }

func Gia_ObjFailed(p *Gia_Man_t, Id int32) int32 { return p.pReprs[Id].fFailed;       }
func Gia_ObjSetFailed(p *Gia_Man_t, Id int32) { p.pReprs[Id].fFailed = 1;          }

func Gia_ObjColor(p *Gia_Man_t, Id int32, c int32) int32 { if c != 0 {
    return p.pReprs[Id].fColorB;
} else {
    return p.pReprs[Id].fColorA;
}          }
func Gia_ObjColors(p *Gia_Man_t, Id int32) int32 { return p.pReprs[Id].fColorB * 2 + p.pReprs[Id].fColorA;         }
func Gia_ObjSetColor(p *Gia_Man_t, Id int32, c int32) { if c != 0 { p.pReprs[Id].fColorB = 1; } else { p.pReprs[Id].fColorA = 1; } }
func Gia_ObjSetColors(p *Gia_Man_t, Id int32) { p.pReprs[Id].fColorB = p.pReprs[Id].fColorA = 1;                }
func Gia_ObjVisitColor(p *Gia_Man_t, Id int32, c int32) int32 { var x int32; if c != 0 { x = p.pReprs[Id].fColorB; p.pReprs[Id].fColorB = 1; } else { x = p.pReprs[Id].fColorA; p.pReprs[Id].fColorA = 1; }
 return x; }
func Gia_ObjDiffColors(p *Gia_Man_t, i int32, j int32) int32 { return (p.pReprs[i].fColorA ^ p.pReprs[j].fColorA) && (p.pReprs[i].fColorB ^ p.pReprs[j].fColorB); }
func Gia_ObjDiffColors2(p *Gia_Man_t, i int32, j int32) int32 { return (p.pReprs[i].fColorA ^ p.pReprs[j].fColorA) || (p.pReprs[i].fColorB ^ p.pReprs[j].fColorB); }

func Gia_ObjNextObj(p *Gia_Man_t, Id int32) *Gia_Obj_t { if p.pNexts[Id] == 0 {
    return nil;
} else {
    return Gia_ManObj( p, p.pNexts[Id] );
}}
func Gia_ObjNext(p *Gia_Man_t, Id int32) int32 { return p.pNexts[Id];                                             }
func Gia_ObjSetNext(p *Gia_Man_t, Id int32, Num int32) { p.pNexts[Id] = Num;                                              }

func Gia_ObjIsConst(p *Gia_Man_t, Id int32) int32 { return Gia_ObjRepr(p, Id) == 0;                                   }
func Gia_ObjIsHead(p *Gia_Man_t, Id int32) int32 { return Gia_ObjRepr(p, Id) == GIA_VOID && Gia_ObjNext(p, Id) > 0;  }
func Gia_ObjIsNone(p *Gia_Man_t, Id int32) int32 { return Gia_ObjRepr(p, Id) == GIA_VOID && Gia_ObjNext(p, Id) <= 0; }
func Gia_ObjIsTail(p *Gia_Man_t, Id int32) int32 { return (Gia_ObjRepr(p, Id) > 0 && Gia_ObjRepr(p, Id) != GIA_VOID) && Gia_ObjNext(p, Id) <= 0;                  }
func Gia_ObjIsClass(p *Gia_Man_t, Id int32) int32 { return (Gia_ObjRepr(p, Id) > 0 && Gia_ObjRepr(p, Id) != GIA_VOID) || Gia_ObjNext(p, Id) > 0;                   }
func Gia_ObjHasSameRepr(p *Gia_Man_t, i int32, k int32) int32 { assert((k != 0), "k"); if i {
    return (Gia_ObjRepr(p, i) == Gia_ObjRepr(p, k) && Gia_ObjRepr(p, i) != GIA_VOID);
} else {
    return Gia_ObjRepr(p, k) == 0;
}  }
func Gia_ObjIsFailedPair(p *Gia_Man_t, i int32, k int32) int32 { assert((k != 0), "k"); if i {
    return (Gia_ObjFailed(p, i) || Gia_ObjFailed(p, k));
} else {
    return Gia_ObjFailed(p, k);
}                     }
func Gia_ClassIsPair(p *Gia_Man_t, i int32) int32 { assert((Gia_ObjIsHead(p, i) != 0), "Gia_ObjIsHead(p, i)"); assert((Gia_ObjNext(p, i) != 0), "Gia_ObjNext(p, i)"); return Gia_ObjNext(p, Gia_ObjNext(p, i)) <= 0;     }
func Gia_ClassUndoPair(p *Gia_Man_t, i int32) { assert((Gia_ClassIsPair(p,i) != 0), "Gia_ClassIsPair(p,i)"); Gia_ObjSetRepr(p, Gia_ObjNext(p, i), GIA_VOID); Gia_ObjSetNext(p, i, 0);       }

// Original macro:
// #define Gia_ManForEachConst( p, i )                            \
//     for ( i = 1; i < Gia_ManObjNum(p); i++ ) if ( !Gia_ObjIsConst(p, i) ) {} else

// Original macro:
// #define Gia_ManForEachClass( p, i )                            \
//     for ( i = 1; i < Gia_ManObjNum(p); i++ ) if ( !Gia_ObjIsHead(p, i) ) {} else

// Original macro:
// #define Gia_ManForEachClass0( p, i )                            \
//     for ( i = 0; i < Gia_ManObjNum(p); i++ ) if ( !Gia_ObjIsHead(p, i) ) {} else

// Original macro:
// #define Gia_ManForEachClassReverse( p, i )                     \
//     for ( i = Gia_ManObjNum(p) - 1; i > 0; i-- ) if ( !Gia_ObjIsHead(p, i) ) {} else

// Original macro:
// #define Gia_ClassForEachObj( p, i, iObj )                      \
//     for ( assert(Gia_ObjIsHead(p, i) && i), iObj = i; iObj > 0; iObj = Gia_ObjNext(p, iObj) )

// Original macro:
// #define Gia_ClassForEachObj1( p, i, iObj )                     \
//     for ( assert(Gia_ObjIsHead(p, i)), iObj = Gia_ObjNext(p, i); iObj > 0; iObj = Gia_ObjNext(p, iObj) )

// Original macro:
// #define Gia_ClassForEachObjStart( p, i, iObj, Start )          \
//     for ( assert(Gia_ObjIsHead(p, i)), iObj = Gia_ObjNext(p, Start); iObj > 0; iObj = Gia_ObjNext(p, iObj) )



func Gia_ObjFoffsetId(p *Gia_Man_t, Id int32) int32 { return Vec_IntEntry( p.vFanout, Id );                                 }
func Gia_ObjFoffset(p *Gia_Man_t, pObj *Gia_Obj_t) int32 { return Gia_ObjFoffsetId( p, Gia_ObjId(p, pObj) );                      }
func Gia_ObjFanoutNumId(p *Gia_Man_t, Id int32) int32 { return Vec_IntEntry( p.vFanoutNums, Id );                             }
func Gia_ObjFanoutNum(p *Gia_Man_t, pObj *Gia_Obj_t) int32 { return Gia_ObjFanoutNumId( p, Gia_ObjId(p, pObj) );                    }
func Gia_ObjFanoutId(p *Gia_Man_t, Id int32, i int32) int32 { return Vec_IntEntry( p.vFanout, Gia_ObjFoffsetId(p, Id) + i );        }
func Gia_ObjFanout0(p *Gia_Man_t, pObj *Gia_Obj_t) *Gia_Obj_t { return Gia_ManObj( p, Gia_ObjFanoutId(p, Gia_ObjId(p, pObj), 0) );     }
func Gia_ObjFanout(p *Gia_Man_t, pObj *Gia_Obj_t, i int32) *Gia_Obj_t { return Gia_ManObj( p, Gia_ObjFanoutId(p, Gia_ObjId(p, pObj), i) );     }
func Gia_ObjSetFanout(p *Gia_Man_t, pObj *Gia_Obj_t, i int32, pFan *Gia_Obj_t) { Vec_IntWriteEntry( p.vFanout, Gia_ObjFoffset(p, pObj) + i, Gia_ObjId(p, pFan) ); }
func Gia_ObjSetFanoutInt(p *Gia_Man_t, pObj *Gia_Obj_t, i int32, x int32) { Vec_IntWriteEntry( p.vFanout, Gia_ObjFoffset(p, pObj) + i, x );                  }

// Original macro:
// #define Gia_ObjForEachFanoutStatic( p, pObj, pFanout, i )      \
//     for ( i = 0; (i < Gia_ObjFanoutNum(p, pObj))   && (((pFanout) = Gia_ObjFanout(p, pObj, i)), 1); i++ )

// Original macro:
// #define Gia_ObjForEachFanoutStaticId( p, Id, FanId, i )      \
//     for ( i = 0; (i < Gia_ObjFanoutNumId(p, Id))   && (((FanId) = Gia_ObjFanoutId(p, Id, i)), 1); i++ )


func Gia_ManHasMapping(p *Gia_Man_t) int32 { return p.vMapping != nil;                                                   }
func Gia_ObjIsLut(p *Gia_Man_t, Id int32) int32 { return Vec_IntEntry(p.vMapping, Id) != 0;                                    }
func Gia_ObjLutSize(p *Gia_Man_t, Id int32) int32 { return Vec_IntEntry(p.vMapping, Vec_IntEntry(p.vMapping, Id));              }
func Gia_ObjLutFanins(p *Gia_Man_t, Id int32) *int32 { return Vec_IntEntryP(p.vMapping, Vec_IntEntry(p.vMapping, Id)) + 1;         }
func Gia_ObjLutFanin(p *Gia_Man_t, Id int32, i int32) int32 { return Gia_ObjLutFanins(p, Id)[i];                                            }
func Gia_ObjLutMuxId(p *Gia_Man_t, Id int32) int32 { return Gia_ObjLutFanins(p, Id)[Gia_ObjLutSize(p, Id)];                        }
func Gia_ObjLutIsMux(p *Gia_Man_t, Id int32) int32 { return int32((Gia_ObjLutMuxId(p, Id) < 0));                                     }

func Gia_ManHasMapping2(p *Gia_Man_t) int32 { return p.vMapping2 != nil;                                                  }
func Gia_ObjIsLut2(p *Gia_Man_t, Id int32) int32 { return Vec_IntSize(Vec_WecEntry(p.vMapping2, Id)) != 0;                      }
func Gia_ObjLutSize2(p *Gia_Man_t, Id int32) int32 { return Vec_IntSize(Vec_WecEntry(p.vMapping2, Id));                           }
func Gia_ObjLutFanins2(p *Gia_Man_t, Id int32) *Vec_Int_t { return Vec_WecEntry(p.vMapping2, Id);                                        }
func Gia_ObjLutFanin2(p *Gia_Man_t, Id int32, i int32) int32 { return Vec_IntEntry(Vec_WecEntry(p.vMapping2, Id), i);                       }
func Gia_ObjLutFanoutNum2(p *Gia_Man_t, Id int32) int32 { return Vec_IntSize(Vec_WecEntry(p.vFanouts2, Id));                           }
func Gia_ObjLutFanout2(p *Gia_Man_t, Id int32, i int32) int32 { return Vec_IntEntry(Vec_WecEntry(p.vFanouts2, Id), i);                       }

func Gia_ManHasCellMapping(p *Gia_Man_t) int32 { return p.vCellMapping != nil;                                               }
func Gia_ObjIsCell(p *Gia_Man_t, iLit int32) int32 { return Vec_IntEntry(p.vCellMapping, iLit) != 0;                              }
func Gia_ObjIsCellInv(p *Gia_Man_t, iLit int32) int32 { return Vec_IntEntry(p.vCellMapping, iLit) == -1;                             }
func Gia_ObjIsCellBuf(p *Gia_Man_t, iLit int32) int32 { return Vec_IntEntry(p.vCellMapping, iLit) == -2;                             }
func Gia_ObjCellSize(p *Gia_Man_t, iLit int32) int32 { return Vec_IntEntry(p.vCellMapping, Vec_IntEntry(p.vCellMapping, iLit));    }
func Gia_ObjCellFanins(p *Gia_Man_t, iLit int32) *int32 { return Vec_IntEntryP(p.vCellMapping, Vec_IntEntry(p.vCellMapping, iLit))+1; }
func Gia_ObjCellFanin(p *Gia_Man_t, iLit int32, i int32) int32 { return Gia_ObjCellFanins(p, iLit)[i];                                         }
func Gia_ObjCellId(p *Gia_Man_t, iLit int32) int32 { return Gia_ObjCellFanins(p, iLit)[Gia_ObjCellSize(p, iLit)];                  }

// Original macro:
// #define Gia_ManForEachLut( p, i )                                       \
//     for ( i = 1; i < Gia_ManObjNum(p); i++ ) if ( !Gia_ObjIsLut(p, i) ) {} else

// Original macro:
// #define Gia_ManForEachLutReverse( p, i )                                \
//     for ( i = Gia_ManObjNum(p) - 1; i > 0; i-- ) if ( !Gia_ObjIsLut(p, i) ) {} else

// Original macro:
// #define Gia_LutForEachFanin( p, i, iFan, k )                            \
//     for ( k = 0; k < Gia_ObjLutSize(p,i) && ((iFan = Gia_ObjLutFanins(p,i)[k]),1); k++ )

// Original macro:
// #define Gia_LutForEachFaninObj( p, i, pFanin, k )                       \
//     for ( k = 0; k < Gia_ObjLutSize(p,i) && ((pFanin = Gia_ManObj(p, Gia_ObjLutFanins(p,i)[k])),1); k++ )


// Original macro:
// #define Gia_ManForEachLut2( p, i )                                      \
//     for ( i = 1; i < Gia_ManObjNum(p); i++ ) if ( !Gia_ObjIsLut2(p, i) ) {} else

// Original macro:
// #define Gia_ManForEachLut2Reverse( p, i )                               \
//     for ( i = Gia_ManObjNum(p) - 1; i > 0; i-- ) if ( !Gia_ObjIsLut2(p, i) ) {} else

// Original macro:
// #define Gia_ManForEachLut2Vec( vIds, p, vVec, iObj, i )                 \
//     for ( i = 0; i < Vec_IntSize(vIds) && (vVec = Vec_WecEntry(p->vMapping2, (iObj = Vec_IntEntry(vIds, i)))); i++ )

// Original macro:
// #define Gia_ManForEachLut2VecReverse( vIds, p, vVec, iObj, i )          \
//     for ( i = Vec_IntSize(vIds)-1; i >= 0 && (vVec = Vec_WecEntry(p->vMapping2, (iObj = Vec_IntEntry(vIds, i)))); i-- )

// Original macro:
// #define Gia_LutForEachFanin2( p, i, iFan, k )                           \
//     for ( k = 0; k < Gia_ObjLutSize2(p,i) && ((iFan = Gia_ObjLutFanin2(p,i,k)),1); k++ )

// Original macro:
// #define Gia_LutForEachFanout2( p, i, iFan, k )                          \
//     for ( k = 0; k < Gia_ObjLutFanoutNum2(p,i) && ((iFan = Gia_ObjLutFanout2(p,i,k)),1); k++ )


// Original macro:
// #define Gia_ManForEachCell( p, i )                                      \
//     for ( i = 2; i < 2*Gia_ManObjNum(p); i++ ) if ( !Gia_ObjIsCell(p, i) ) {} else

// Original macro:
// #define Gia_CellForEachFanin( p, i, iFanLit, k )                        \
//     for ( k = 0; k < Gia_ObjCellSize(p,i) && ((iFanLit = Gia_ObjCellFanins(p,i)[k]),1); k++ )


////////////////////////////////////////////////////////////////////////
///                      MACRO DEFINITIONS                           ///
////////////////////////////////////////////////////////////////////////

// Original macro:
// #define Gia_ManForEachObj( p, pObj, i )                                 \
//     for ( i = 0; (i < p->nObjs) && ((pObj) = Gia_ManObj(p, i)); i++ )

// Original macro:
// #define Gia_ManForEachObj1( p, pObj, i )                                \
//     for ( i = 1; (i < p->nObjs) && ((pObj) = Gia_ManObj(p, i)); i++ )

// Original macro:
// #define Gia_ManForEachObjVec( vVec, p, pObj, i )                        \
//     for ( i = 0; (i < Vec_IntSize(vVec)) && ((pObj) = Gia_ManObj(p, Vec_IntEntry(vVec,i))); i++ )

// Original macro:
// #define Gia_ManForEachObjVecReverse( vVec, p, pObj, i )                        \
//     for ( i = Vec_IntSize(vVec) - 1; (i >= 0) && ((pObj) = Gia_ManObj(p, Vec_IntEntry(vVec,i))); i-- )

// Original macro:
// #define Gia_ManForEachObjVecLit( vVec, p, pObj, fCompl, i )             \
//     for ( i = 0; (i < Vec_IntSize(vVec)) && ((pObj) = Gia_ManObj(p, Abc_Lit2Var(Vec_IntEntry(vVec,i)))) && (((fCompl) = Abc_LitIsCompl(Vec_IntEntry(vVec,i))),1); i++ )

// Original macro:
// #define Gia_ManForEachObjReverse( p, pObj, i )                          \
//     for ( i = p->nObjs - 1; (i >= 0) && ((pObj) = Gia_ManObj(p, i)); i-- )

// Original macro:
// #define Gia_ManForEachObjReverse1( p, pObj, i )                         \
//     for ( i = p->nObjs - 1; (i > 0) && ((pObj) = Gia_ManObj(p, i)); i-- )

// Original macro:
// #define Gia_ManForEachBuf( p, pObj, i )                                 \
//     for ( i = Gia_ManBufNum(p) ? 0 : p->nObjs; (i < p->nObjs) && ((pObj) = Gia_ManObj(p, i)); i++ )      if ( !Gia_ObjIsBuf(pObj) ) {} else

// Original macro:
// #define Gia_ManForEachBufId( p, i )                                     \
//     for ( i = 0; (i < p->nObjs); i++ )                                     if ( !Gia_ObjIsBuf(Gia_ManObj(p, i)) ) {} else

// Original macro:
// #define Gia_ManForEachAnd( p, pObj, i )                                 \
//     for ( i = 0; (i < p->nObjs) && ((pObj) = Gia_ManObj(p, i)); i++ )      if ( !Gia_ObjIsAnd(pObj) ) {} else

// Original macro:
// #define Gia_ManForEachAndId( p, i )                                     \
//     for ( i = 0; (i < p->nObjs); i++ )                                     if ( !Gia_ObjIsAnd(Gia_ManObj(p, i)) ) {} else

// Original macro:
// #define Gia_ManForEachMuxId( p, i )                                     \
//     for ( i = 0; (i < p->nObjs); i++ )                                     if ( !Gia_ObjIsMuxId(p, i) ) {} else

// Original macro:
// #define Gia_ManForEachCand( p, pObj, i )                                \
//     for ( i = 0; (i < p->nObjs) && ((pObj) = Gia_ManObj(p, i)); i++ )      if ( !Gia_ObjIsCand(pObj) ) {} else

// Original macro:
// #define Gia_ManForEachAndReverse( p, pObj, i )                          \
//     for ( i = p->nObjs - 1; (i > 0) && ((pObj) = Gia_ManObj(p, i)); i-- )  if ( !Gia_ObjIsAnd(pObj) ) {} else

// Original macro:
// #define Gia_ManForEachAndReverseId( p, i )                              \
//     for ( i = p->nObjs - 1; (i > 0); i-- )                                 if ( !Gia_ObjIsAnd(Gia_ManObj(p, i)) ) {} else

// Original macro:
// #define Gia_ManForEachMux( p, pObj, i )                                 \
//     for ( i = 0; (i < p->nObjs) && ((pObj) = Gia_ManObj(p, i)); i++ )      if ( !Gia_ObjIsMuxId(p, i) ) {} else

// Original macro:
// #define Gia_ManForEachCi( p, pObj, i )                                  \
//     for ( i = 0; (i < Vec_IntSize(p->vCis)) && ((pObj) = Gia_ManCi(p, i)); i++ )

// Original macro:
// #define Gia_ManForEachCiId( p, Id, i )                                  \
//     for ( i = 0; (i < Vec_IntSize(p->vCis)) && ((Id) = Gia_ObjId(p, Gia_ManCi(p, i))); i++ )

// Original macro:
// #define Gia_ManForEachCiVec( vVec, p, pObj, i )                         \
//     for ( i = 0; (i < Vec_IntSize(vVec)) && ((pObj) = Gia_ManCi(p, Vec_IntEntry(vVec,i))); i++ )

// Original macro:
// #define Gia_ManForEachCiReverse( p, pObj, i )                           \
//     for ( i = Vec_IntSize(p->vCis) - 1; (i >= 0) && ((pObj) = Gia_ManCi(p, i)); i-- )

// Original macro:
// #define Gia_ManForEachCo( p, pObj, i )                                  \
//     for ( i = 0; (i < Vec_IntSize(p->vCos)) && ((pObj) = Gia_ManCo(p, i)); i++ )

// Original macro:
// #define Gia_ManForEachCoVec( vVec, p, pObj, i )                         \
//     for ( i = 0; (i < Vec_IntSize(vVec)) && ((pObj) = Gia_ManCo(p, Vec_IntEntry(vVec,i))); i++ )

// Original macro:
// #define Gia_ManForEachCoId( p, Id, i )                                  \
//     for ( i = 0; (i < Vec_IntSize(p->vCos)) && ((Id) = Gia_ObjId(p, Gia_ManCo(p, i))); i++ )

// Original macro:
// #define Gia_ManForEachCoReverse( p, pObj, i )                           \
//     for ( i = Vec_IntSize(p->vCos) - 1; (i >= 0) && ((pObj) = Gia_ManCo(p, i)); i-- )

// Original macro:
// #define Gia_ManForEachCoDriver( p, pObj, i )                            \
//     for ( i = 0; (i < Vec_IntSize(p->vCos)) && ((pObj) = Gia_ObjFanin0(Gia_ManCo(p, i))); i++ )

// Original macro:
// #define Gia_ManForEachCoDriverId( p, DriverId, i )                      \
//     for ( i = 0; (i < Vec_IntSize(p->vCos)) && (((DriverId) = Gia_ObjFaninId0p(p, Gia_ManCo(p, i))), 1); i++ )

// Original macro:
// #define Gia_ManForEachPi( p, pObj, i )                                  \
//     for ( i = 0; (i < Gia_ManPiNum(p)) && ((pObj) = Gia_ManCi(p, i)); i++ )

// Original macro:
// #define Gia_ManForEachPo( p, pObj, i )                                  \
//     for ( i = 0; (i < Gia_ManPoNum(p)) && ((pObj) = Gia_ManCo(p, i)); i++ )

// Original macro:
// #define Gia_ManForEachRo( p, pObj, i )                                  \
//     for ( i = 0; (i < Gia_ManRegNum(p)) && ((pObj) = Gia_ManCi(p, Gia_ManPiNum(p)+i)); i++ )

// Original macro:
// #define Gia_ManForEachRi( p, pObj, i )                                  \
//     for ( i = 0; (i < Gia_ManRegNum(p)) && ((pObj) = Gia_ManCo(p, Gia_ManPoNum(p)+i)); i++ )

// Original macro:
// #define Gia_ManForEachRiRo( p, pObjRi, pObjRo, i )                      \
//     for ( i = 0; (i < Gia_ManRegNum(p)) && ((pObjRi) = Gia_ManCo(p, Gia_ManPoNum(p)+i)) && ((pObjRo) = Gia_ManCi(p, Gia_ManPiNum(p)+i)); i++ )

 
////////////////////////////////////////////////////////////////////////
///                    FUNCTION DECLARATIONS                         ///
////////////////////////////////////////////////////////////////////////

/*=== giaAiger.c ===========================================================*/
// extern int                 Gia_FileSize( char * pFileName );

// extern Gia_Man_t *         Gia_AigerReadFromMemory( char * pContents, int nFileSize, int fGiaSimple, int fSkipStrash, int fCheck );

// extern Gia_Man_t *         Gia_AigerRead( char * pFileName, int fGiaSimple, int fSkipStrash, int fCheck );

// extern void                Gia_AigerWrite( Gia_Man_t * p, char * pFileName, int fWriteSymbols, int fCompact, int fWriteNewLine );

// extern void                Gia_DumpAiger( Gia_Man_t * p, char * pFilePrefix, int iFileNum, int nFileNumDigits );

// extern Vec_Str_t *         Gia_AigerWriteIntoMemoryStr( Gia_Man_t * p );

// extern Vec_Str_t *         Gia_AigerWriteIntoMemoryStrPart( Gia_Man_t * p, Vec_Int_t * vCis, Vec_Int_t * vAnds, Vec_Int_t * vCos, int nRegs );

// extern void                Gia_AigerWriteSimple( Gia_Man_t * pInit, char * pFileName );

/*=== giaBalance.c ===========================================================*/
// extern Gia_Man_t *         Gia_ManBalance( Gia_Man_t * p, int fSimpleAnd, int fStrict, int fVerbose );

// extern Gia_Man_t *         Gia_ManAreaBalance( Gia_Man_t * p, int fSimpleAnd, int nNewNodesMax, int fVerbose, int fVeryVerbose );

// extern Gia_Man_t *         Gia_ManAigSyn2( Gia_Man_t * p, int fOldAlgo, int fCoarsen, int fCutMin, int nRelaxRatio, int fDelayMin, int fVerbose, int fVeryVerbose );

// extern Gia_Man_t *         Gia_ManAigSyn3( Gia_Man_t * p, int fVerbose, int fVeryVerbose );

// extern Gia_Man_t *         Gia_ManAigSyn4( Gia_Man_t * p, int fVerbose, int fVeryVerbose );

/*=== giaBidec.c ===========================================================*/
// extern unsigned *          Gia_ManConvertAigToTruth( Gia_Man_t * p, Gia_Obj_t * pRoot, Vec_Int_t * vLeaves, Vec_Int_t * vTruth, Vec_Int_t * vVisited );

// extern Gia_Man_t *         Gia_ManPerformBidec( Gia_Man_t * p, int fVerbose );

/*=== giaCex.c ============================================================*/
// extern int                 Gia_ManVerifyCex( Gia_Man_t * pAig, Abc_Cex_t * p, int fDualOut );

// extern int                 Gia_ManFindFailedPoCex( Gia_Man_t * pAig, Abc_Cex_t * p, int nOutputs );

// extern int                 Gia_ManSetFailedPoCex( Gia_Man_t * pAig, Abc_Cex_t * p );

// extern void                Gia_ManCounterExampleValueStart( Gia_Man_t * pGia, Abc_Cex_t * pCex );

// extern void                Gia_ManCounterExampleValueStop( Gia_Man_t * pGia );

// extern int                 Gia_ManCounterExampleValueLookup( Gia_Man_t * pGia, int Id, int iFrame );

// extern Abc_Cex_t *         Gia_ManCexExtendToIncludeCurrentStates( Gia_Man_t * p, Abc_Cex_t * pCex );

// extern Abc_Cex_t *         Gia_ManCexExtendToIncludeAllObjects( Gia_Man_t * p, Abc_Cex_t * pCex );

/*=== giaCsatOld.c ============================================================*/
// extern Vec_Int_t *         Cbs_ManSolveMiter( Gia_Man_t * pGia, int nConfs, Vec_Str_t ** pvStatus, int fVerbose );

/*=== giaCsat.c ============================================================*/
type Cbs_Man_t =  Cbs_Man_t_;
// extern Cbs_Man_t *         Cbs_ManAlloc( Gia_Man_t * pGia );

// extern void                Cbs_ManStop( Cbs_Man_t * p );

// extern int                 Cbs_ManSolve( Cbs_Man_t * p, Gia_Obj_t * pObj );

// extern int                 Cbs_ManSolve2( Cbs_Man_t * p, Gia_Obj_t * pObj, Gia_Obj_t * pObj2 );

// extern Vec_Int_t *         Cbs_ManSolveMiterNc( Gia_Man_t * pGia, int nConfs, Vec_Str_t ** pvStatus, int f0Proved, int fVerbose );

// extern void                Cbs_ManSetConflictNum( Cbs_Man_t * p, int Num );

// extern Vec_Int_t *         Cbs_ReadModel( Cbs_Man_t * p );

/*=== giaCTas.c ============================================================*/
// extern Vec_Int_t *         Tas_ManSolveMiterNc( Gia_Man_t * pGia, int nConfs, Vec_Str_t ** pvStatus, int fVerbose );

/*=== giaCof.c =============================================================*/
// extern void                Gia_ManPrintFanio( Gia_Man_t * pGia, int nNodes );

// extern Gia_Man_t *         Gia_ManDupCof( Gia_Man_t * p, int iVar );

// extern Gia_Man_t *         Gia_ManDupCofAllInt( Gia_Man_t * p, Vec_Int_t * vSigs, int fVerbose );

// extern Gia_Man_t *         Gia_ManDupCofAll( Gia_Man_t * p, int nFanLim, int fVerbose );

/*=== giaDfs.c ============================================================*/
// extern void                Gia_ManCollectCis( Gia_Man_t * p, int * pNodes, int nNodes, Vec_Int_t * vSupp );

// extern void                Gia_ManCollectAnds_rec( Gia_Man_t * p, int iObj, Vec_Int_t * vNodes );

// extern void                Gia_ManCollectAnds( Gia_Man_t * p, int * pNodes, int nNodes, Vec_Int_t * vNodes, Vec_Int_t * vLeaves );

// extern Vec_Int_t *         Gia_ManCollectAndsAll( Gia_Man_t * p );

// extern Vec_Int_t *         Gia_ManCollectNodesCis( Gia_Man_t * p, int * pNodes, int nNodes );

// extern int                 Gia_ManSuppSize( Gia_Man_t * p, int * pNodes, int nNodes );

// extern int                 Gia_ManConeSize( Gia_Man_t * p, int * pNodes, int nNodes );

// extern Vec_Vec_t *         Gia_ManLevelize( Gia_Man_t * p );

// extern Vec_Wec_t *         Gia_ManLevelizeR( Gia_Man_t * p );

// extern Vec_Int_t *         Gia_ManOrderReverse( Gia_Man_t * p );

// extern void                Gia_ManCollectTfi( Gia_Man_t * p, Vec_Int_t * vRoots, Vec_Int_t * vNodes );

// extern void                Gia_ManCollectTfo( Gia_Man_t * p, Vec_Int_t * vRoots, Vec_Int_t * vNodes );

/*=== giaDup.c ============================================================*/
// extern void                Gia_ManDupRemapLiterals( Vec_Int_t * vLits, Gia_Man_t * p );

// extern void                Gia_ManDupRemapEquiv( Gia_Man_t * pNew, Gia_Man_t * p );

// extern Gia_Man_t *         Gia_ManDupOrderDfs( Gia_Man_t * p );

// extern Gia_Man_t *         Gia_ManDupOrderDfsChoices( Gia_Man_t * p );

// extern Gia_Man_t *         Gia_ManDupOrderDfsReverse( Gia_Man_t * p );

// extern Gia_Man_t *         Gia_ManDupOutputGroup( Gia_Man_t * p, int iOutStart, int iOutStop );

// extern Gia_Man_t *         Gia_ManDupOutputVec( Gia_Man_t * p, Vec_Int_t * vOutPres );

// extern Gia_Man_t *         Gia_ManDupSelectedOutputs( Gia_Man_t * p, Vec_Int_t * vOutsLeft );

// extern Gia_Man_t *         Gia_ManDupOrderAiger( Gia_Man_t * p );

// extern Gia_Man_t *         Gia_ManDupLastPis( Gia_Man_t * p, int nLastPis );

// extern Gia_Man_t *         Gia_ManDupFlip( Gia_Man_t * p, int * pInitState );

// extern Gia_Man_t *         Gia_ManDupCycled( Gia_Man_t * pAig, Abc_Cex_t * pCex, int nFrames );

// extern Gia_Man_t *         Gia_ManDup( Gia_Man_t * p );
  
// extern Gia_Man_t *         Gia_ManDup2( Gia_Man_t * p1, Gia_Man_t * p2 );

// extern Gia_Man_t *         Gia_ManDupWithAttributes( Gia_Man_t * p );
  
// extern Gia_Man_t *         Gia_ManDupRemovePis( Gia_Man_t * p, int nRemPis );

// extern Gia_Man_t *         Gia_ManDupZero( Gia_Man_t * p );

// extern Gia_Man_t *         Gia_ManDupPerm( Gia_Man_t * p, Vec_Int_t * vPiPerm );

// extern Gia_Man_t *         Gia_ManDupPermFlop( Gia_Man_t * p, Vec_Int_t * vFfPerm );

// extern Gia_Man_t *         Gia_ManDupPermFlopGap( Gia_Man_t * p, Vec_Int_t * vFfPerm );

// extern void                Gia_ManDupAppend( Gia_Man_t * p, Gia_Man_t * pTwo );

// extern void                Gia_ManDupAppendShare( Gia_Man_t * p, Gia_Man_t * pTwo );

// extern Gia_Man_t *         Gia_ManDupAppendNew( Gia_Man_t * pOne, Gia_Man_t * pTwo );

// extern Gia_Man_t *         Gia_ManDupAppendCones( Gia_Man_t * p, Gia_Man_t ** ppCones, int nCones, int fOnlyRegs );

// extern Gia_Man_t *         Gia_ManDupSelf( Gia_Man_t * p );

// extern Gia_Man_t *         Gia_ManDupFlopClass( Gia_Man_t * p, int iClass );

// extern Gia_Man_t *         Gia_ManDupMarked( Gia_Man_t * p );

// extern Gia_Man_t *         Gia_ManDupTimes( Gia_Man_t * p, int nTimes );
  
// extern Gia_Man_t *         Gia_ManDupDfs( Gia_Man_t * p );
  
// extern Gia_Man_t *         Gia_ManDupDfsOnePo( Gia_Man_t * p, int iPo );

// extern Gia_Man_t *         Gia_ManDupCofactorVar( Gia_Man_t * p, int iVar, int Value );
  
// extern Gia_Man_t *         Gia_ManDupCofactorObj( Gia_Man_t * p, int iObj, int Value );
  
// extern Gia_Man_t *         Gia_ManDupMux( int iVar, Gia_Man_t * pCof1, Gia_Man_t * pCof0 );

// extern Gia_Man_t *         Gia_ManDupBlock( Gia_Man_t * p, int nBlock );

// extern Gia_Man_t *         Gia_ManDupExist( Gia_Man_t * p, int iVar );

// extern Gia_Man_t *         Gia_ManDupUniv( Gia_Man_t * p, int iVar );

// extern Gia_Man_t *         Gia_ManDupDfsSkip( Gia_Man_t * p );

// extern Gia_Man_t *         Gia_ManDupDfsCone( Gia_Man_t * p, Gia_Obj_t * pObj );

// extern Gia_Man_t *         Gia_ManDupConeSupp( Gia_Man_t * p, int iLit, Vec_Int_t * vCiIds );

// extern int                 Gia_ManDupConeBack( Gia_Man_t * p, Gia_Man_t * pNew, Vec_Int_t * vCiIds );

// extern int                 Gia_ManDupConeBackObjs( Gia_Man_t * p, Gia_Man_t * pNew, Vec_Int_t * vObjs );

// extern Gia_Man_t *         Gia_ManDupDfsNode( Gia_Man_t * p, Gia_Obj_t * pObj );

// extern Gia_Man_t *         Gia_ManDupDfsLitArray( Gia_Man_t * p, Vec_Int_t * vLits );

// extern Gia_Man_t *         Gia_ManDupTrimmed( Gia_Man_t * p, int fTrimCis, int fTrimCos, int fDualOut, int OutValue );

// extern Gia_Man_t *         Gia_ManDupOntop( Gia_Man_t * p, Gia_Man_t * p2 );

// extern Gia_Man_t *         Gia_ManDupWithNewPo( Gia_Man_t * p1, Gia_Man_t * p2 );

// extern Gia_Man_t *         Gia_ManDupDfsCiMap( Gia_Man_t * p, int * pCi2Lit, Vec_Int_t * vLits );

// extern Gia_Man_t *         Gia_ManPermuteInputs( Gia_Man_t * p, int nPpis, int nExtra );

// extern Gia_Man_t *         Gia_ManDupDfsClasses( Gia_Man_t * p );

// extern Gia_Man_t *         Gia_ManDupTopAnd( Gia_Man_t * p, int fVerbose );

// extern Gia_Man_t *         Gia_ManMiter( Gia_Man_t * pAig0, Gia_Man_t * pAig1, int nInsDup, int fDualOut, int fSeq, int fImplic, int fVerbose );

// extern Gia_Man_t *         Gia_ManDupAndOr( Gia_Man_t * p, int nOuts, int fUseOr, int fCompl );

// extern Gia_Man_t *         Gia_ManDupZeroUndc( Gia_Man_t * p, char * pInit, int nNewPis, int fGiaSimple, int fVerbose );

// extern Gia_Man_t *         Gia_ManMiter2( Gia_Man_t * p, char * pInit, int fVerbose );

// extern Gia_Man_t *         Gia_ManTransformMiter( Gia_Man_t * p );

// extern Gia_Man_t *         Gia_ManTransformMiter2( Gia_Man_t * p );

// extern Gia_Man_t *         Gia_ManTransformToDual( Gia_Man_t * p );

// extern Gia_Man_t *         Gia_ManTransformTwoWord2DualOutput( Gia_Man_t * p );

// extern Gia_Man_t *         Gia_ManChoiceMiter( Vec_Ptr_t * vGias );

// extern Gia_Man_t *         Gia_ManDupWithConstraints( Gia_Man_t * p, Vec_Int_t * vPoTypes );

// extern Gia_Man_t *         Gia_ManDupCones( Gia_Man_t * p, int * pPos, int nPos, int fTrimPis );

// extern Gia_Man_t *         Gia_ManDupAndCones( Gia_Man_t * p, int * pAnds, int nAnds, int fTrimPis );

// extern Gia_Man_t *         Gia_ManDupAndConesLimit( Gia_Man_t * p, int * pAnds, int nAnds, int Level );

// extern Gia_Man_t *         Gia_ManDupAndConesLimit2( Gia_Man_t * p, int * pAnds, int nAnds, int Level );

// extern Gia_Man_t *         Gia_ManDupOneHot( Gia_Man_t * p );

// extern Gia_Man_t *         Gia_ManDupLevelized( Gia_Man_t * p );

// extern Gia_Man_t *         Gia_ManDupFromVecs( Gia_Man_t * p, Vec_Int_t * vCis, Vec_Int_t * vAnds, Vec_Int_t * vCos, int nRegs );

// extern Gia_Man_t *         Gia_ManDupSliced( Gia_Man_t * p, int nSuppMax );

// extern Gia_Man_t *         Gia_ManDupDemiter( Gia_Man_t * p, int fVerbose );

// extern Gia_Man_t *         Gia_ManDemiterToDual( Gia_Man_t * p );

// extern int                 Gia_ManDemiterDual( Gia_Man_t * p, Gia_Man_t ** pp0, Gia_Man_t ** pp1 );

// extern int                 Gia_ManDemiterTwoWords( Gia_Man_t * p, Gia_Man_t ** pp0, Gia_Man_t ** pp1 );

/*=== giaEdge.c ==========================================================*/
// extern void                Gia_ManEdgeFromArray( Gia_Man_t * p, Vec_Int_t * vArray );

// extern Vec_Int_t *         Gia_ManEdgeToArray( Gia_Man_t * p );

// extern void                Gia_ManConvertPackingToEdges( Gia_Man_t * p );

// extern int                 Gia_ObjCheckEdge( Gia_Man_t * p, int iObj, int iNext );

// extern int                 Gia_ManEvalEdgeDelay( Gia_Man_t * p );

// extern int                 Gia_ManEvalEdgeCount( Gia_Man_t * p );

// extern int                 Gia_ManComputeEdgeDelay( Gia_Man_t * p, int fUseTwo );

// extern int                 Gia_ManComputeEdgeDelay2( Gia_Man_t * p );

// extern void                Gia_ManUpdateMapping( Gia_Man_t * p, Vec_Int_t * vNodes, Vec_Wec_t * vWin );

// extern int                 Gia_ManEvalWindow( Gia_Man_t * p, Vec_Int_t * vLeaves, Vec_Int_t * vNodes, Vec_Wec_t * vWin, Vec_Int_t * vTemp, int fUseTwo );

/*=== giaEnable.c ==========================================================*/
// extern void                Gia_ManDetectSeqSignals( Gia_Man_t * p, int fSetReset, int fVerbose );

// extern Gia_Man_t *         Gia_ManUnrollAndCofactor( Gia_Man_t * p, int nFrames, int nFanMax, int fVerbose );

// extern Gia_Man_t *         Gia_ManRemoveEnables( Gia_Man_t * p );

/*=== giaEquiv.c ==========================================================*/
// extern void                Gia_ManOrigIdsInit( Gia_Man_t * p );

// extern void                Gia_ManOrigIdsStart( Gia_Man_t * p );

// extern void                Gia_ManOrigIdsRemap( Gia_Man_t * p, Gia_Man_t * pNew );

// extern Gia_Man_t *         Gia_ManOrigIdsReduce( Gia_Man_t * p, Vec_Int_t * vPairs );

// extern Gia_Man_t *         Gia_ManComputeGiaEquivs( Gia_Man_t * pGia, int nConfs, int fVerbose );

// extern void                Gia_ManEquivFixOutputPairs( Gia_Man_t * p );

// extern int                 Gia_ManCheckTopoOrder( Gia_Man_t * p );

// extern int *               Gia_ManDeriveNexts( Gia_Man_t * p );

// extern void                Gia_ManDeriveReprs( Gia_Man_t * p );

// extern int                 Gia_ManEquivCountLits( Gia_Man_t * p );

// extern int                 Gia_ManEquivCountLitsAll( Gia_Man_t * p );

// extern int                 Gia_ManEquivCountClasses( Gia_Man_t * p );

// extern void                Gia_ManEquivPrintOne( Gia_Man_t * p, int i, int Counter );

// extern void                Gia_ManEquivPrintClasses( Gia_Man_t * p, int fVerbose, float Mem );

// extern Gia_Man_t *         Gia_ManEquivReduce( Gia_Man_t * p, int fUseAll, int fDualOut, int fSkipPhase, int fVerbose );

// extern Gia_Man_t *         Gia_ManEquivReduceAndRemap( Gia_Man_t * p, int fSeq, int fMiterPairs );

// extern int                 Gia_ManEquivSetColors( Gia_Man_t * p, int fVerbose );

// extern Gia_Man_t *         Gia_ManSpecReduce( Gia_Man_t * p, int fDualOut, int fSynthesis, int fReduce, int fSkipSome, int fVerbose );

// extern Gia_Man_t *         Gia_ManSpecReduceInit( Gia_Man_t * p, Abc_Cex_t * pInit, int nFrames, int fDualOut );

// extern Gia_Man_t *         Gia_ManSpecReduceInitFrames( Gia_Man_t * p, Abc_Cex_t * pInit, int nFramesMax, int * pnFrames, int fDualOut, int nMinOutputs );

// extern void                Gia_ManEquivTransform( Gia_Man_t * p, int fVerbose );

// extern void                Gia_ManEquivImprove( Gia_Man_t * p );

// extern Gia_Man_t *         Gia_ManEquivToChoices( Gia_Man_t * p, int nSnapshots );

// extern int                 Gia_ManCountChoiceNodes( Gia_Man_t * p );

// extern int                 Gia_ManCountChoices( Gia_Man_t * p );

// extern int                 Gia_ManFilterEquivsForSpeculation( Gia_Man_t * pGia, char * pName1, char * pName2, int fLatchA, int fLatchB );

// extern int                 Gia_ManFilterEquivsUsingParts( Gia_Man_t * pGia, char * pName1, char * pName2 );

// extern void                Gia_ManFilterEquivsUsingLatches( Gia_Man_t * pGia, int fFlopsOnly, int fFlopsWith, int fUseRiDrivers );

/*=== giaExist.c =========================================================*/
// extern void                Gia_ManQuantSetSuppStart( Gia_Man_t * p );

// extern void                Gia_ManQuantSetSuppZero( Gia_Man_t * p );

// extern void                Gia_ManQuantSetSuppCi( Gia_Man_t * p, Gia_Obj_t * pObj );

// extern void                Gia_ManQuantUpdateCiSupp( Gia_Man_t * p, int iObj );

// extern int                 Gia_ManQuantExist( Gia_Man_t * p, int iLit, int(*pFuncCiToKeep)(void *, int), void * pData );

/*=== giaFanout.c =========================================================*/
// extern void                Gia_ObjAddFanout( Gia_Man_t * p, Gia_Obj_t * pObj, Gia_Obj_t * pFanout );

// extern void                Gia_ObjRemoveFanout( Gia_Man_t * p, Gia_Obj_t * pObj, Gia_Obj_t * pFanout );

// extern void                Gia_ManFanoutStart( Gia_Man_t * p );

// extern void                Gia_ManFanoutStop( Gia_Man_t * p );

// extern void                Gia_ManStaticFanoutStart( Gia_Man_t * p );

// extern void                Gia_ManStaticFanoutStop( Gia_Man_t * p );

// extern void                Gia_ManStaticMappingFanoutStart( Gia_Man_t * p );

/*=== giaForce.c =========================================================*/
// extern void                For_ManExperiment( Gia_Man_t * pGia, int nIters, int fClustered, int fVerbose );

/*=== giaFrames.c =========================================================*/
// extern Gia_Man_t *         Gia_ManUnrollDup( Gia_Man_t * p, Vec_Int_t * vLimit );

// extern Vec_Ptr_t *         Gia_ManUnrollAbs( Gia_Man_t * p, int nFrames );

// extern void *              Gia_ManUnrollStart( Gia_Man_t * pAig, Gia_ParFra_t * pPars );

// extern void *              Gia_ManUnrollAdd( void * pMan, int fMax );

// extern void                Gia_ManUnrollStop( void * pMan );

// extern int                 Gia_ManUnrollLastLit( void * pMan );

// extern void                Gia_ManFraSetDefaultParams( Gia_ParFra_t * p );

// extern Gia_Man_t *         Gia_ManFrames( Gia_Man_t * pAig, Gia_ParFra_t * pPars );
  
// extern Gia_Man_t *         Gia_ManFramesInitSpecial( Gia_Man_t * pAig, int nFrames, int fVerbose );

/*=== giaFront.c ==========================================================*/
// extern Gia_Man_t *         Gia_ManFront( Gia_Man_t * p );

// extern void                Gia_ManFrontTest( Gia_Man_t * p );

/*=== giaFx.c ==========================================================*/
// extern Gia_Man_t *         Gia_ManPerformFx( Gia_Man_t * p, int nNewNodesMax, int LitCountMax, int fReverse, int fVerbose, int fVeryVerbose );

/*=== giaHash.c ===========================================================*/
// extern void                Gia_ManHashAlloc( Gia_Man_t * p );
 
// extern void                Gia_ManHashStart( Gia_Man_t * p );
 
// extern void                Gia_ManHashStop( Gia_Man_t * p );

// extern int                 Gia_ManHashXorReal( Gia_Man_t * p, int iLit0, int iLit1 );

// extern int                 Gia_ManHashMuxReal( Gia_Man_t * p, int iLitC, int iLit1, int iLit0 );

// extern int                 Gia_ManHashAnd( Gia_Man_t * p, int iLit0, int iLit1 );
 
// extern int                 Gia_ManHashOr( Gia_Man_t * p, int iLit0, int iLit1 );
 
// extern int                 Gia_ManHashXor( Gia_Man_t * p, int iLit0, int iLit1 );
 
// extern int                 Gia_ManHashMux( Gia_Man_t * p, int iCtrl, int iData1, int iData0 );

// extern int                 Gia_ManHashMaj( Gia_Man_t * p, int iData0, int iData1, int iData2 );

// extern int                 Gia_ManHashAndTry( Gia_Man_t * p, int iLit0, int iLit1 );

// extern Gia_Man_t *         Gia_ManRehash( Gia_Man_t * p, int fAddStrash );

// extern void                Gia_ManHashProfile( Gia_Man_t * p );

// extern int                 Gia_ManHashLookupInt( Gia_Man_t * p, int iLit0, int iLit1 );

// extern int                 Gia_ManHashLookup( Gia_Man_t * p, Gia_Obj_t * p0, Gia_Obj_t * p1 );

// extern int                 Gia_ManHashAndMulti( Gia_Man_t * p, Vec_Int_t * vLits );

// extern int                 Gia_ManHashAndMulti2( Gia_Man_t * p, Vec_Int_t * vLits );

// extern int                 Gia_ManHashDualMiter( Gia_Man_t * p, Vec_Int_t * vOuts );

/*=== giaIf.c ===========================================================*/
// extern void                Gia_ManPrintMappingStats( Gia_Man_t * p, char * pDumpFile );

// extern void                Gia_ManPrintPackingStats( Gia_Man_t * p );

// extern void                Gia_ManPrintLutStats( Gia_Man_t * p );

// extern int                 Gia_ManLutFaninCount( Gia_Man_t * p );

// extern int                 Gia_ManLutSizeMax( Gia_Man_t * p );

// extern int                 Gia_ManLutNum( Gia_Man_t * p );

// extern int                 Gia_ManLutLevel( Gia_Man_t * p, int ** ppLevels );

// extern void                Gia_ManLutParams( Gia_Man_t * p, int * pnCurLuts, int * pnCurEdges, int * pnCurLevels );

// extern void                Gia_ManSetRefsMapped( Gia_Man_t * p );

// extern void                Gia_ManSetLutRefs( Gia_Man_t * p );
  
// extern void                Gia_ManSetIfParsDefault( void * pIfPars );

// extern void                Gia_ManMappingVerify( Gia_Man_t * p );

// extern void                Gia_ManTransferMapping( Gia_Man_t * p, Gia_Man_t * pGia );

// extern void                Gia_ManTransferPacking( Gia_Man_t * p, Gia_Man_t * pGia );

// extern void                Gia_ManTransferTiming( Gia_Man_t * p, Gia_Man_t * pGia );

// extern Gia_Man_t *         Gia_ManPerformMapping( Gia_Man_t * p, void * pIfPars );

// extern Gia_Man_t *         Gia_ManPerformSopBalance( Gia_Man_t * p, int nCutNum, int nRelaxRatio, int fVerbose );

// extern Gia_Man_t *         Gia_ManPerformDsdBalance( Gia_Man_t * p, int nLutSize, int nCutNum, int nRelaxRatio, int fVerbose );

// extern Gia_Man_t *         Gia_ManDupHashMapping( Gia_Man_t * p );

/*=== giaJf.c ===========================================================*/
// extern void                Jf_ManSetDefaultPars( Jf_Par_t * pPars );

// extern Gia_Man_t *         Jf_ManPerformMapping( Gia_Man_t * pGia, Jf_Par_t * pPars );

// extern Gia_Man_t *         Jf_ManDeriveCnf( Gia_Man_t * p, int fCnfObjIds );

/*=== giaIso.c ===========================================================*/
// extern Gia_Man_t *         Gia_ManIsoCanonicize( Gia_Man_t * p, int fVerbose );

// extern Gia_Man_t *         Gia_ManIsoReduce( Gia_Man_t * p, Vec_Ptr_t ** pvPosEquivs, Vec_Ptr_t ** pvPiPerms, int fEstimate, int fDualOut, int fVerbose, int fVeryVerbose );

// extern Gia_Man_t *         Gia_ManIsoReduce2( Gia_Man_t * p, Vec_Ptr_t ** pvPosEquivs, Vec_Ptr_t ** pvPiPerms, int fEstimate, int fBetterQual, int fDualOut, int fVerbose, int fVeryVerbose );

/*=== giaLf.c ===========================================================*/
// extern void                Lf_ManSetDefaultPars( Jf_Par_t * pPars );

// extern Gia_Man_t *         Lf_ManPerformMapping( Gia_Man_t * pGia, Jf_Par_t * pPars );

// extern Gia_Man_t *         Gia_ManPerformLfMapping( Gia_Man_t * p, Jf_Par_t * pPars, int fNormalized );

/*=== giaLogic.c ===========================================================*/
// extern void                Gia_ManTestDistance( Gia_Man_t * p );

// extern void                Gia_ManSolveProblem( Gia_Man_t * pGia, Emb_Par_t * pPars );

 /*=== giaMan.c ===========================================================*/
// extern Gia_Man_t *         Gia_ManStart( int nObjsMax );
 
// extern void                Gia_ManStop( Gia_Man_t * p );
  
// extern void                Gia_ManStopP( Gia_Man_t ** p );
  
// extern double              Gia_ManMemory( Gia_Man_t * p );

// extern void                Gia_ManPrintStats( Gia_Man_t * p, Gps_Par_t * pPars );
 
// extern void                Gia_ManPrintStatsShort( Gia_Man_t * p );
 
// extern void                Gia_ManPrintMiterStatus( Gia_Man_t * p );
 
// extern void                Gia_ManPrintStatsMiter( Gia_Man_t * p, int fVerbose );

// extern void                Gia_ManSetRegNum( Gia_Man_t * p, int nRegs );

// extern void                Gia_ManReportImprovement( Gia_Man_t * p, Gia_Man_t * pNew );

// extern void                Gia_ManPrintNpnClasses( Gia_Man_t * p );

// extern void                Gia_ManDumpVerilog( Gia_Man_t * p, char * pFileName, Vec_Int_t * vObjs );

/*=== giaMem.c ===========================================================*/
// extern Gia_MmFixed_t *     Gia_MmFixedStart( int nEntrySize, int nEntriesMax );

// extern void                Gia_MmFixedStop( Gia_MmFixed_t * p, int fVerbose );

// extern char *              Gia_MmFixedEntryFetch( Gia_MmFixed_t * p );

// extern void                Gia_MmFixedEntryRecycle( Gia_MmFixed_t * p, char * pEntry );

// extern void                Gia_MmFixedRestart( Gia_MmFixed_t * p );

// extern int                 Gia_MmFixedReadMemUsage( Gia_MmFixed_t * p );

// extern int                 Gia_MmFixedReadMaxEntriesUsed( Gia_MmFixed_t * p );

// extern Gia_MmFlex_t *      Gia_MmFlexStart();

// extern void                Gia_MmFlexStop( Gia_MmFlex_t * p, int fVerbose );

// extern char *              Gia_MmFlexEntryFetch( Gia_MmFlex_t * p, int nBytes );

// extern void                Gia_MmFlexRestart( Gia_MmFlex_t * p );

// extern int                 Gia_MmFlexReadMemUsage( Gia_MmFlex_t * p );

// extern Gia_MmStep_t *      Gia_MmStepStart( int nSteps );

// extern void                Gia_MmStepStop( Gia_MmStep_t * p, int fVerbose );

// extern char *              Gia_MmStepEntryFetch( Gia_MmStep_t * p, int nBytes );

// extern void                Gia_MmStepEntryRecycle( Gia_MmStep_t * p, char * pEntry, int nBytes );

// extern int                 Gia_MmStepReadMemUsage( Gia_MmStep_t * p );

/*=== giaMf.c ===========================================================*/
// extern void                Mf_ManSetDefaultPars( Jf_Par_t * pPars );

// extern Gia_Man_t *         Mf_ManPerformMapping( Gia_Man_t * pGia, Jf_Par_t * pPars );

// extern void *              Mf_ManGenerateCnf( Gia_Man_t * pGia, int nLutSize, int fCnfObjIds, int fAddOrCla, int fMapping, int fVerbose );

/*=== giaMini.c ===========================================================*/
// extern Gia_Man_t *         Gia_ManReadMiniAig( char * pFileName );

// extern void                Gia_ManWriteMiniAig( Gia_Man_t * pGia, char * pFileName );

// extern Gia_Man_t *         Gia_ManReadMiniLut( char * pFileName );

// extern void                Gia_ManWriteMiniLut( Gia_Man_t * pGia, char * pFileName );

/*=== giaMuxes.c ===========================================================*/
// extern void                Gia_ManCountMuxXor( Gia_Man_t * p, int * pnMuxes, int * pnXors );

// extern void                Gia_ManPrintMuxStats( Gia_Man_t * p );

// extern Gia_Man_t *         Gia_ManDupMuxes( Gia_Man_t * p, int Limit );

// extern Gia_Man_t *         Gia_ManDupNoMuxes( Gia_Man_t * p, int fSkipBufs );

/*=== giaPat.c ===========================================================*/
// extern void                Gia_SatVerifyPattern( Gia_Man_t * p, Gia_Obj_t * pRoot, Vec_Int_t * vCex, Vec_Int_t * vVisit );

/*=== giaRetime.c ===========================================================*/
// extern Gia_Man_t *         Gia_ManRetimeForward( Gia_Man_t * p, int nMaxIters, int fVerbose );

/*=== giaSat.c ============================================================*/
// extern int                 Sat_ManTest( Gia_Man_t * pGia, Gia_Obj_t * pObj, int nConfsMax );

/*=== giaScl.c ============================================================*/
// extern int                 Gia_ManSeqMarkUsed( Gia_Man_t * p );

// extern int                 Gia_ManCombMarkUsed( Gia_Man_t * p );

// extern Gia_Man_t *         Gia_ManCleanup( Gia_Man_t * p );

// extern Gia_Man_t *         Gia_ManCleanupOutputs( Gia_Man_t * p, int nOutputs );

// extern Gia_Man_t *         Gia_ManSeqCleanup( Gia_Man_t * p );

// extern Gia_Man_t *         Gia_ManSeqStructSweep( Gia_Man_t * p, int fConst, int fEquiv, int fVerbose );

/*=== giaShow.c ===========================================================*/
// extern void                Gia_ManShow( Gia_Man_t * pMan, Vec_Int_t * vBold, int fAdders, int fFadds, int fPath );

/*=== giaShrink.c ===========================================================*/
// extern Gia_Man_t *         Gia_ManMapShrink4( Gia_Man_t * p, int fKeepLevel, int fVerbose );

// extern Gia_Man_t *         Gia_ManMapShrink6( Gia_Man_t * p, int nFanoutMax, int fKeepLevel, int fVerbose );

/*=== giaSopb.c ============================================================*/
// extern Gia_Man_t *         Gia_ManExtractWindow( Gia_Man_t * p, int LevelMax, int nTimeWindow, int fVerbose );

// extern Gia_Man_t *         Gia_ManPerformSopBalanceWin( Gia_Man_t * p, int LevelMax, int nTimeWindow, int nCutNum, int nRelaxRatio, int fVerbose );

// extern Gia_Man_t *         Gia_ManPerformDsdBalanceWin( Gia_Man_t * p, int LevelMax, int nTimeWindow, int nLutSize, int nCutNum, int nRelaxRatio, int fVerbose );

/*=== giaSort.c ============================================================*/
// extern int *               Gia_SortFloats( float * pArray, int * pPerm, int nSize );

/*=== giaSim.c ============================================================*/
// extern void                Gia_ManSimSetDefaultParams( Gia_ParSim_t * p );

// extern int                 Gia_ManSimSimulate( Gia_Man_t * pAig, Gia_ParSim_t * pPars );

// extern unsigned *          Gia_SimDataExt( Gia_ManSim_t * p, int i );

// extern unsigned *          Gia_SimDataCiExt( Gia_ManSim_t * p, int i );

// extern unsigned *          Gia_SimDataCoExt( Gia_ManSim_t * p, int i );

// extern void                Gia_ManSimInfoInit( Gia_ManSim_t * p );

// extern void                Gia_ManSimInfoTransfer( Gia_ManSim_t * p );

// extern void                Gia_ManSimulateRound( Gia_ManSim_t * p );

// extern void                Gia_ManBuiltInSimStart( Gia_Man_t * p, int nWords, int nObjs );

// extern void                Gia_ManBuiltInSimPerform( Gia_Man_t * p, int iObj );

// extern int                 Gia_ManBuiltInSimCheckOver( Gia_Man_t * p, int iLit0, int iLit1 );

// extern int                 Gia_ManBuiltInSimCheckEqual( Gia_Man_t * p, int iLit0, int iLit1 );

// extern void                Gia_ManBuiltInSimResimulateCone( Gia_Man_t * p, int iLit0, int iLit1 );

// extern void                Gia_ManBuiltInSimResimulate( Gia_Man_t * p );

// extern int                 Gia_ManBuiltInSimAddPat( Gia_Man_t * p, Vec_Int_t * vPat );

// extern void                Gia_ManIncrSimStart( Gia_Man_t * p, int nWords, int nObjs );

// extern void                Gia_ManIncrSimSet( Gia_Man_t * p, Vec_Int_t * vObjLits );

// extern int                 Gia_ManIncrSimCheckOver( Gia_Man_t * p, int iLit0, int iLit1 );

// extern int                 Gia_ManIncrSimCheckEqual( Gia_Man_t * p, int iLit0, int iLit1 );

/*=== giaSimBase.c ============================================================*/
// extern Vec_Wrd_t *         Gia_ManSimPatSim( Gia_Man_t * p );

/*=== giaSpeedup.c ============================================================*/
// extern float               Gia_ManDelayTraceLut( Gia_Man_t * p );

// extern float               Gia_ManDelayTraceLutPrint( Gia_Man_t * p, int fVerbose );

// extern Gia_Man_t *         Gia_ManSpeedup( Gia_Man_t * p, int Percentage, int Degree, int fVerbose, int fVeryVerbose );

/*=== giaSplit.c ============================================================*/
// extern void                Gia_ManComputeOneWinStart( Gia_Man_t * p, int nAnds, int fReverse );

// extern int                 Gia_ManComputeOneWin( Gia_Man_t * p, int iPivot, Vec_Int_t ** pvRoots, Vec_Int_t ** pvNodes, Vec_Int_t ** pvLeaves, Vec_Int_t ** pvAnds );

/*=== giaStg.c ============================================================*/
// extern void                Gia_ManStgPrint( FILE * pFile, Vec_Int_t * vLines, int nIns, int nOuts, int nStates );

// extern Gia_Man_t *         Gia_ManStgRead( char * pFileName, int kHot, int fVerbose );

/*=== giaSupp.c ============================================================*/
type Gia_ManMin_t = Gia_ManMin_t_;
// extern Gia_ManMin_t *      Gia_ManSuppStart( Gia_Man_t * pGia );

// extern void                Gia_ManSuppStop( Gia_ManMin_t * p );

// extern int                 Gia_ManSupportAnd( Gia_ManMin_t * p, int iLit0, int iLit1 );

type Gia_Man2Min_t = Gia_Man2Min_t_;
// extern Gia_Man2Min_t *     Gia_Man2SuppStart( Gia_Man_t * pGia );

// extern void                Gia_Man2SuppStop( Gia_Man2Min_t * p );

// extern int                 Gia_Man2SupportAnd( Gia_Man2Min_t * p, int iLit0, int iLit1 );

/*=== giaSweep.c ============================================================*/
// extern Gia_Man_t *         Gia_ManFraigSweepSimple( Gia_Man_t * p, void * pPars );

// extern Gia_Man_t *         Gia_ManSweepWithBoxes( Gia_Man_t * p, void * pParsC, void * pParsS, int fConst, int fEquiv, int fVerbose, int fVerbEquivs );

// extern void                Gia_ManCheckIntegrityWithBoxes( Gia_Man_t * p );

/*=== giaSweeper.c ============================================================*/
// extern Gia_Man_t *         Gia_SweeperStart( Gia_Man_t * p );

// extern void                Gia_SweeperStop( Gia_Man_t * p );

// extern int                 Gia_SweeperIsRunning( Gia_Man_t * p );

// extern void                Gia_SweeperPrintStats( Gia_Man_t * p );

// extern void                Gia_SweeperSetConflictLimit( Gia_Man_t * p, int nConfMax );

// extern void                Gia_SweeperSetRuntimeLimit( Gia_Man_t * p, int nSeconds );

// extern Vec_Int_t *         Gia_SweeperGetCex( Gia_Man_t * p );

// extern int                 Gia_SweeperProbeCreate( Gia_Man_t * p, int iLit );

// extern int                 Gia_SweeperProbeDelete( Gia_Man_t * p, int ProbeId );

// extern int                 Gia_SweeperProbeUpdate( Gia_Man_t * p, int ProbeId, int iLitNew );

// extern int                 Gia_SweeperProbeLit( Gia_Man_t * p, int ProbeId );

// extern Vec_Int_t *         Gia_SweeperCollectValidProbeIds( Gia_Man_t * p );

// extern int                 Gia_SweeperCondPop( Gia_Man_t * p );

// extern void                Gia_SweeperCondPush( Gia_Man_t * p, int ProbeId );

// extern Vec_Int_t *         Gia_SweeperCondVector( Gia_Man_t * p );

// extern int                 Gia_SweeperCondCheckUnsat( Gia_Man_t * p );

// extern int                 Gia_SweeperCheckEquiv( Gia_Man_t * p, int ProbeId1, int ProbeId2 );

// extern Gia_Man_t *         Gia_SweeperExtractUserLogic( Gia_Man_t * p, Vec_Int_t * vProbeIds, Vec_Ptr_t * vInNames, Vec_Ptr_t * vOutNames );

// extern void                Gia_SweeperLogicDump( Gia_Man_t * p, Vec_Int_t * vProbeIds, int fDumpConds, char * pFileName );

// extern Gia_Man_t *         Gia_SweeperCleanup( Gia_Man_t * p, char * pCommLime );

// extern Vec_Int_t *         Gia_SweeperGraft( Gia_Man_t * pDst, Vec_Int_t * vProbes, Gia_Man_t * pSrc );

// extern int                 Gia_SweeperFraig( Gia_Man_t * p, Vec_Int_t * vProbeIds, char * pCommLime, int nWords, int nConfs, int fVerify, int fVerbose );

// extern int                 Gia_SweeperRun( Gia_Man_t * p, Vec_Int_t * vProbeIds, char * pCommLime, int fVerbose );

/*=== giaSwitch.c ============================================================*/
// extern float               Gia_ManEvaluateSwitching( Gia_Man_t * p );

// extern float               Gia_ManComputeSwitching( Gia_Man_t * p, int nFrames, int nPref, int fProbOne );

// extern Vec_Int_t *         Gia_ManComputeSwitchProbs( Gia_Man_t * pGia, int nFrames, int nPref, int fProbOne );

// extern Vec_Flt_t *         Gia_ManPrintOutputProb( Gia_Man_t * p );

/*=== giaTim.c ===========================================================*/
// extern int                 Gia_ManBoxNum( Gia_Man_t * p );

// extern int                 Gia_ManRegBoxNum( Gia_Man_t * p );

// extern int                 Gia_ManNonRegBoxNum( Gia_Man_t * p );

// extern int                 Gia_ManBlackBoxNum( Gia_Man_t * p );

// extern int                 Gia_ManBoxCiNum( Gia_Man_t * p );

// extern int                 Gia_ManBoxCoNum( Gia_Man_t * p );

// extern int                 Gia_ManClockDomainNum( Gia_Man_t * p );

// extern int                 Gia_ManIsSeqWithBoxes( Gia_Man_t * p );

// extern int                 Gia_ManIsNormalized( Gia_Man_t * p );

// extern Vec_Int_t *         Gia_ManOrderWithBoxes( Gia_Man_t * p );

// extern Gia_Man_t *         Gia_ManDupNormalize( Gia_Man_t * p, int fHashMapping );

// extern Gia_Man_t *         Gia_ManDupUnnormalize( Gia_Man_t * p );

// extern Gia_Man_t *         Gia_ManDupUnshuffleInputs( Gia_Man_t * p );

// extern int                 Gia_ManLevelWithBoxes( Gia_Man_t * p );

// extern int                 Gia_ManLutLevelWithBoxes( Gia_Man_t * p );

// extern void *              Gia_ManUpdateTimMan( Gia_Man_t * p, Vec_Int_t * vBoxPres );

// extern void *              Gia_ManUpdateTimMan2( Gia_Man_t * p, Vec_Int_t * vBoxesLeft, int nTermsDiff );

// extern Gia_Man_t *         Gia_ManUpdateExtraAig( void * pTime, Gia_Man_t * pAig, Vec_Int_t * vBoxPres );

// extern Gia_Man_t *         Gia_ManUpdateExtraAig2( void * pTime, Gia_Man_t * pAig, Vec_Int_t * vBoxesLeft );

// extern Gia_Man_t *         Gia_ManDupCollapse( Gia_Man_t * p, Gia_Man_t * pBoxes, Vec_Int_t * vBoxPres, int fSeq );

// extern int                 Gia_ManVerifyWithBoxes( Gia_Man_t * pGia, int nBTLimit, int nTimeLim, int fSeq, int fDumpFiles, int fVerbose, char * pFileSpec );

/*=== giaTruth.c ===========================================================*/
// extern word                Gia_LutComputeTruth6( Gia_Man_t * p, int iObj, Vec_Wrd_t * vTruths );

// extern word                Gia_ObjComputeTruthTable6Lut( Gia_Man_t * p, int iObj, Vec_Wrd_t * vTemp );

// extern word                Gia_ObjComputeTruthTable6( Gia_Man_t * p, Gia_Obj_t * pObj, Vec_Int_t * vSupp, Vec_Wrd_t * vTruths );

// extern word                Gia_ObjComputeTruth6Cis( Gia_Man_t * p, int iLit, Vec_Int_t * vSupp, Vec_Wrd_t * vTemp );

// extern void                Gia_ObjCollectInternal( Gia_Man_t * p, Gia_Obj_t * pObj );

// extern word *              Gia_ObjComputeTruthTable( Gia_Man_t * p, Gia_Obj_t * pObj );

// extern void                Gia_ObjComputeTruthTableStart( Gia_Man_t * p, int nVarsMax );

// extern void                Gia_ObjComputeTruthTableStop( Gia_Man_t * p );

// extern word *              Gia_ObjComputeTruthTableCut( Gia_Man_t * p, Gia_Obj_t * pObj, Vec_Int_t * vLeaves );

/*=== giaTsim.c ============================================================*/
// extern Gia_Man_t *         Gia_ManReduceConst( Gia_Man_t * pAig, int fVerbose );

/*=== giaUtil.c ===========================================================*/
// extern unsigned            Gia_ManRandom( int fReset );

// extern word                Gia_ManRandomW( int fReset );

// extern void                Gia_ManRandomInfo( Vec_Ptr_t * vInfo, int iInputStart, int iWordStart, int iWordStop );

// extern char *              Gia_TimeStamp();

// extern char *              Gia_FileNameGenericAppend( char * pBase, char * pSuffix );

// extern void                Gia_ManIncrementTravId( Gia_Man_t * p );

// extern void                Gia_ManCleanMark01( Gia_Man_t * p );

// extern void                Gia_ManSetMark0( Gia_Man_t * p );

// extern void                Gia_ManCleanMark0( Gia_Man_t * p );

// extern void                Gia_ManCheckMark0( Gia_Man_t * p );

// extern void                Gia_ManSetMark1( Gia_Man_t * p );

// extern void                Gia_ManCleanMark1( Gia_Man_t * p );

// extern void                Gia_ManCheckMark1( Gia_Man_t * p );

// extern void                Gia_ManCleanValue( Gia_Man_t * p );

// extern void                Gia_ManCleanLevels( Gia_Man_t * p, int Size );

// extern void                Gia_ManCleanTruth( Gia_Man_t * p );

// extern void                Gia_ManFillValue( Gia_Man_t * p );

// extern void                Gia_ObjSetPhase( Gia_Man_t * p, Gia_Obj_t * pObj );

// extern void                Gia_ManSetPhase( Gia_Man_t * p );

// extern void                Gia_ManSetPhasePattern( Gia_Man_t * p, Vec_Int_t * vCiValues );

// extern void                Gia_ManSetPhase1( Gia_Man_t * p );

// extern void                Gia_ManCleanPhase( Gia_Man_t * p );

// extern int                 Gia_ManCheckCoPhase( Gia_Man_t * p );

// extern int                 Gia_ManLevelNum( Gia_Man_t * p );

// extern int                 Gia_ManLevelRNum( Gia_Man_t * p );

// extern Vec_Int_t *         Gia_ManGetCiLevels( Gia_Man_t * p );

// extern int                 Gia_ManSetLevels( Gia_Man_t * p, Vec_Int_t * vCiLevels );

// extern Vec_Int_t *         Gia_ManReverseLevel( Gia_Man_t * p );

// extern Vec_Int_t *         Gia_ManRequiredLevel( Gia_Man_t * p );

// extern void                Gia_ManCreateValueRefs( Gia_Man_t * p );

// extern void                Gia_ManCreateRefs( Gia_Man_t * p );

// extern int *               Gia_ManCreateMuxRefs( Gia_Man_t * p );

// extern int                 Gia_ManCrossCut( Gia_Man_t * p, int fReverse );

// extern Vec_Int_t *         Gia_ManCollectPoIds( Gia_Man_t * p );

// extern int                 Gia_ObjIsMuxType( Gia_Obj_t * pNode );

// extern int                 Gia_ObjRecognizeExor( Gia_Obj_t * pObj, Gia_Obj_t ** ppFan0, Gia_Obj_t ** ppFan1 );

// extern Gia_Obj_t *         Gia_ObjRecognizeMux( Gia_Obj_t * pNode, Gia_Obj_t ** ppNodeT, Gia_Obj_t ** ppNodeE );

// extern int                 Gia_ObjRecognizeMuxLits( Gia_Man_t * p, Gia_Obj_t * pNode, int * iLitT, int * iLitE );

// extern int                 Gia_NodeMffcSize( Gia_Man_t * p, Gia_Obj_t * pNode );

// extern int                 Gia_NodeMffcSizeMark( Gia_Man_t * p, Gia_Obj_t * pNode );

// extern int                 Gia_NodeMffcSizeSupp( Gia_Man_t * p, Gia_Obj_t * pNode, Vec_Int_t * vSupp );

// extern int                 Gia_NodeMffcMapping( Gia_Man_t * p );

// extern int                 Gia_ManHasDangling( Gia_Man_t * p );

// extern int                 Gia_ManMarkDangling( Gia_Man_t * p );

// extern Vec_Int_t *         Gia_ManGetDangling( Gia_Man_t * p );

// extern void                Gia_ObjPrint( Gia_Man_t * p, Gia_Obj_t * pObj );

// extern void                Gia_ManPrint( Gia_Man_t * p );

// extern void                Gia_ManPrintCo( Gia_Man_t * p, Gia_Obj_t * pObj );

// extern void                Gia_ManPrintCone( Gia_Man_t * p, Gia_Obj_t * pObj, int * pLeaves, int nLeaves, Vec_Int_t * vNodes );

// extern void                Gia_ManPrintConeMulti( Gia_Man_t * p, Vec_Int_t * vObjs, Vec_Int_t * vLeaves, Vec_Int_t * vNodes );

// extern void                Gia_ManPrintCone2( Gia_Man_t * p, Gia_Obj_t * pObj );

// extern void                Gia_ManInvertConstraints( Gia_Man_t * pAig );

// extern void                Gia_ManInvertPos( Gia_Man_t * pAig );

// extern int                 Gia_ManCompare( Gia_Man_t * p1, Gia_Man_t * p2 );

// extern void                Gia_ManMarkFanoutDrivers( Gia_Man_t * p );

// extern void                Gia_ManSwapPos( Gia_Man_t * p, int i );

// extern Vec_Int_t *         Gia_ManSaveValue( Gia_Man_t * p );

// extern void                Gia_ManLoadValue( Gia_Man_t * p, Vec_Int_t * vValues );

// extern Vec_Int_t *         Gia_ManFirstFanouts( Gia_Man_t * p );

// extern int                 Gia_ManCheckSuppOverlap( Gia_Man_t * p, int iNode1, int iNode2 );

// extern int                 Gia_ManCountPisWithFanout( Gia_Man_t * p );

// extern int                 Gia_ManCountPosWithNonZeroDrivers( Gia_Man_t * p );

// extern void                Gia_ManUpdateCopy( Vec_Int_t * vCopy, Gia_Man_t * p );

// extern Vec_Int_t *         Gia_ManComputeDistance( Gia_Man_t * p, int iObj, Vec_Int_t * vObjs, int fVerbose );


/*=== giaCTas.c ===========================================================*/
type Tas_Man_t =  Tas_Man_t_;
// extern Tas_Man_t *         Tas_ManAlloc( Gia_Man_t * pAig, int nBTLimit );

// extern void                Tas_ManStop( Tas_Man_t * p );

// extern Vec_Int_t *         Tas_ReadModel( Tas_Man_t * p );

// extern void                Tas_ManSatPrintStats( Tas_Man_t * p );

// extern int                 Tas_ManSolve( Tas_Man_t * p, Gia_Obj_t * pObj, Gia_Obj_t * pObj2 );

// extern int                 Tas_ManSolveArray( Tas_Man_t * p, Vec_Ptr_t * vObjs );



//ABC_NAMESPACE_HEADER_END




////////////////////////////////////////////////////////////////////////
///                       END OF FILE                                ///
////////////////////////////////////////////////////////////////////////

