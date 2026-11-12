package aig

/**CFile****************************************************************

  FileName    [aig.h]

  SystemName  [ABC: Logic synthesis and verification system.]

  PackageName [AIG package.]

  Synopsis    [External declarations.]

  Author      [Alan Mishchenko]
  
  Affiliation [UC Berkeley]

  Date        [Ver. 1.0. Started - April 28, 2007.]

  Revision    [$Id: aig.h,v 1.00 2007/04/28 00:00:00 alanmi Exp $]

***********************************************************************/





////////////////////////////////////////////////////////////////////////
///                          INCLUDES                                ///
////////////////////////////////////////////////////////////////////////

//#include <stdio.h>
//#include <stdlib.h>
//#include <string.h>
//#include <assert.h>

//#include "misc/vec/vec.h"
//#include "misc/util/utilCex.h"

////////////////////////////////////////////////////////////////////////
///                         PARAMETERS                               ///
////////////////////////////////////////////////////////////////////////



//ABC_NAMESPACE_HEADER_START
 

////////////////////////////////////////////////////////////////////////
///                         BASIC TYPES                              ///
////////////////////////////////////////////////////////////////////////

type Aig_Man_t =            Aig_Man_t_;
type Aig_Obj_t =            Aig_Obj_t_;
type Aig_MmFixed_t =        Aig_MmFixed_t_;    
type Aig_MmFlex_t =         Aig_MmFlex_t_;     
type Aig_MmStep_t =         Aig_MmStep_t_;     

// object types
 type Aig_Type_t int32
const ( 
    AIG_OBJ_NONE Aig_Type_t = iota
                    // 0: non-existent object
    AIG_OBJ_CONST1
                  // 1: constant 1 
    AIG_OBJ_CI
                      // 2: combinational input
    AIG_OBJ_CO
                      // 3: combinational output
    AIG_OBJ_BUF
                     // 4: buffer node
    AIG_OBJ_AND
                     // 5: AND node
    AIG_OBJ_EXOR
                    // 6: EXOR node
    AIG_OBJ_VOID                     // 7: unused object

)

 

// the AIG node
type Aig_Obj_t_ struct {  // 8 words

    union {                         
        Aig_Obj_t *  pNext;          // strashing table
        int          CioId;          // 0-based number of CI/CO
    };
    pFanin0       *Aig_Obj_t;        // fanin
    pFanin1       *Aig_Obj_t;        // fanin
    Type uint32; /* C bitfield: 3 bits; review packing and truncation */   // object type
    fPhase uint32; /* C bitfield: 1 bits; review packing and truncation */   // value under 000...0 pattern
    fMarkA uint32; /* C bitfield: 1 bits; review packing and truncation */   // multipurpose mask
    fMarkB uint32; /* C bitfield: 1 bits; review packing and truncation */   // multipurpose mask
    nRefs uint32; /* C bitfield: 26 bits; review packing and truncation */   // reference count 
    Level uint32; /* C bitfield: 24 bits; review packing and truncation */   // the level of this node
    nCuts uint32; /* C bitfield: 8 bits; review packing and truncation */   // the number of cuts
    TravId              int32;         // unique ID of last traversal involving the node
    Id              int32;             // unique ID of the node
    union {                          // temporary store for user's data
        void *       pData;
        int          iData;
        float        dData;
    };
}

// the AIG manager
type Aig_Man_t_ struct {
    pName            string;          // the design name
    pSpec            string;          // the input file name
    // AIG nodes
    vCis       *Vec_Ptr_t;           // the array of PIs
    vCos       *Vec_Ptr_t;           // the array of POs
    vObjs       *Vec_Ptr_t;          // the array of all nodes (optional)
    vBufs       *Vec_Ptr_t;          // the array of buffers
    pConst1       *Aig_Obj_t;        // the constant 1 node
    Ghost        Aig_Obj_t;          // the ghost node
    nRegs              int32;          // the number of registers (registers are last POs)
    nTruePis              int32;       // the number of true primary inputs
    nTruePos              int32;       // the number of true primary outputs
    nAsserts              int32;       // the number of asserts among POs (asserts are first POs)
    nConstrs              int32;       // the number of constraints (model checking only)
    nBarBufs              int32;       // the number of barrier buffers
    // AIG node counters
    nObjs [AIG_OBJ_VOID]int32;// the number of objects by type
    nDeleted              int32;       // the number of deleted objects
    // structural hash table
    pTable      **Aig_Obj_t;         // structural hash table
    nTableSize              int32;     // structural hash table size
    // representation of fanouts
    pFanData             *int32;       // the database to store fanout information
    nFansAlloc              int32;     // the size of fanout representation
    vLevels       *Vec_Vec_t;        // used to update timing information
    nBufReplaces              int32;   // the number of times replacement led to a buffer
    nBufFixes              int32;      // the number of times buffers were propagated
    nBufMax              int32;        // the maximum number of buffers during computation
    // topological order
    pOrderData        *uint32;
    nOrderAlloc              int32;
    iPrev              int32;
    iNext              int32;
    nAndTotal              int32;
    nAndPrev              int32;
    // representatives
    pEquivs      **Aig_Obj_t;        // linked list of equivalent nodes (when choices are used)
    pReprs      **Aig_Obj_t;         // representatives of each node
    nReprsAlloc              int32;    // the number of allocated representatives
    // various data members
    pMemObjs   *Aig_MmFixed_t;       // memory manager for objects
    vLevelR       *Vec_Int_t;        // the reverse level of the nodes
    nLevelMax              int32;      // maximum number of levels
    pData            any;          // the temporary data
    pData2            any;         // the temporary data
    nTravIds              int32;       // the current traversal ID
    fCatchExor              int32;     // enables EXOR nodes
    fAddStrash              int32;     // performs additional strashing
    pObjCopies      **Aig_Obj_t;     // mapping of AIG nodes into FRAIG nodes
    void (*pImpFunc) (void*, void*); // implication checking precedure
    pImpData            any;       // implication checking data
    pManTime            any;       // the timing manager
    pManCuts            any;
    pFastSim             *int32; 
    pTerSimData        *uint32;    // ternary simulation data
    vMapped       *Vec_Ptr_t;
    vFlopNums       *Vec_Int_t;      
    vFlopReprs       *Vec_Int_t;      
    pSeqModel       *Abc_Cex_t;
    vSeqModelVec       *Vec_Ptr_t;   // vector of counter-examples (for sequential miters)
    pManExdc       *Aig_Man_t;
    vOnehots       *Vec_Ptr_t;
    fCreatePios              int32;
    vEquPairs       *Vec_Int_t;   
    vClockDoms       *Vec_Vec_t; 
    vProbs       *Vec_Int_t;         // probability of node being 1 
    vCiNumsOrig       *Vec_Int_t;    // original CI names
    nComplEdges              int32;    // complemented edges
    Time2Quit          abctime;
    // timing statistics
    time1          abctime;
    time2          abctime;
  //-- jlong -- begin
  unfold2_type_I       *Vec_Ptr_t;
  unfold2_type_II       *Vec_Ptr_t;
  //-- jlong -- end
}

// cut computation
type Aig_ManCut_t =         Aig_ManCut_t_;
              

// the cut used to represent node in the AIG
type Aig_Cut_t struct {
    pNext      *Aig_Cut_t;           // the next cut in the table 
    Cost             int32;            // the cost of the cut
    uSign        uint32;           // cut signature
    iNode             int32;           // the node, for which it is the cut
    nCutSize           int16;        // the number of bytes in the cut
    nLeafMax            byte;        // the maximum number of fanins
    nFanins            byte;         // the current number of fanins
    pFanins [0]int32;      // the fanins (followed by the truth table)
}

// the CNF computation manager
type Aig_ManCut_t_ struct {
    // AIG manager
    pAig      *Aig_Man_t;            // the input AIG manager
    pCuts     **Aig_Cut_t;           // the cuts for each node in the output manager
    // parameters
    nCutsMax             int32;        // the max number of cuts at the node
    nLeafMax             int32;        // the max number of leaves of a cut
    fTruth             int32;          // enables truth table computation
    fVerbose             int32;        // enables verbose output
    // internal variables
    nCutSize             int32;        // the number of bytes needed to store one cut
    nTruthWords             int32;     // the number of truth table words
    pMemCuts  *Aig_MmFixed_t;        // memory manager for cuts
    puTemp [4]*uint32;       // used for the truth table computation
}

func Aig_ObjCuts(p *Aig_ManCut_t, pObj *Aig_Obj_t) *Aig_Cut_t { return p.pCuts[pObj.Id];  }
func Aig_ObjSetCuts(p *Aig_ManCut_t, pObj *Aig_Obj_t, pCuts *Aig_Cut_t) { p.pCuts[pObj.Id] = pCuts; }

func Aig_CutLeaveNum(pCut *Aig_Cut_t) int32 { return pCut.nFanins;                                    }
func Aig_CutLeaves(pCut *Aig_Cut_t) *int32 { return pCut.pFanins;                                    }
func Aig_CutTruth(pCut *Aig_Cut_t) *uint32 { return (unsigned *)(pCut.pFanins + pCut.nLeafMax);     }
func Aig_CutNext(pCut *Aig_Cut_t) *Aig_Cut_t { return (Aig_Cut_t *)(((char *)pCut) + pCut.nCutSize);   }

// iterator over cuts of the node
// Original macro:
// #define Aig_ObjForEachCut( p, pObj, pCut, i )                           \
//     for ( i = 0, pCut = Aig_ObjCuts(p, pObj); i < p->nCutsMax; i++, pCut = Aig_CutNext(pCut) ) 

// iterator over leaves of the cut
// Original macro:
// #define Aig_CutForEachLeaf( p, pCut, pLeaf, i )                         \
//     for ( i = 0; (i < (int)(pCut)->nFanins) && ((pLeaf) = Aig_ManObj(p, (pCut)->pFanins[i])); i++ )


////////////////////////////////////////////////////////////////////////
///                      MACRO DEFINITIONS                           ///
////////////////////////////////////////////////////////////////////////

func Aig_ObjCutSign(ObjId uint32) uint32 { return (1 << (ObjId & 31));                            }
func Aig_WordCountOnes(uWord uint32) int32 {
    uWord = (uWord & 0x55555555) + ((uWord>>1) & 0x55555555);
    uWord = (uWord & 0x33333333) + ((uWord>>2) & 0x33333333);
    uWord = (uWord & 0x0F0F0F0F) + ((uWord>>4) & 0x0F0F0F0F);
    uWord = (uWord & 0x00FF00FF) + ((uWord>>8) & 0x00FF00FF);
    return  (uWord & 0x0000FFFF) + (uWord>>16);
}
func Aig_WordFindFirstBit(uWord uint32) int32 {
    var i int32;
    for  i = 0; i < 32; i++  {
        if uWord & (1 << i) {
            return i; }
    }
    return -1;
}

func Aig_Regular(p *Aig_Obj_t) *Aig_Obj_t { return (Aig_Obj_t *)((ABC_PTRUINT_T)(p) & ~01);  }
func Aig_Not(p *Aig_Obj_t) *Aig_Obj_t { return (Aig_Obj_t *)((ABC_PTRUINT_T)(p) ^  01);  }
func Aig_NotCond(p *Aig_Obj_t, c int32) *Aig_Obj_t { return (Aig_Obj_t *)((ABC_PTRUINT_T)(p) ^ (c));  }
func Aig_IsComplement(p *Aig_Obj_t) int32 { return int32(((ABC_PTRUINT_T)(p) & 01));           }

func Aig_ManCiNum(p *Aig_Man_t) int32 { return p.nObjs[AIG_OBJ_CI];                     }
func Aig_ManCoNum(p *Aig_Man_t) int32 { return p.nObjs[AIG_OBJ_CO];                     }
func Aig_ManBufNum(p *Aig_Man_t) int32 { return p.nObjs[AIG_OBJ_BUF];                    }
func Aig_ManAndNum(p *Aig_Man_t) int32 { return p.nObjs[AIG_OBJ_AND];                    }
func Aig_ManExorNum(p *Aig_Man_t) int32 { return p.nObjs[AIG_OBJ_EXOR];                   }
func Aig_ManNodeNum(p *Aig_Man_t) int32 { return p.nObjs[AIG_OBJ_AND]+p.nObjs[AIG_OBJ_EXOR];   }
func Aig_ManGetCost(p *Aig_Man_t) int32 { return p.nObjs[AIG_OBJ_AND]+3*p.nObjs[AIG_OBJ_EXOR]; }
func Aig_ManObjNum(p *Aig_Man_t) int32 { return Vec_PtrSize(p.vObjs) - p.nDeleted;      }
func Aig_ManObjNumMax(p *Aig_Man_t) int32 { return Vec_PtrSize(p.vObjs);                    }
func Aig_ManRegNum(p *Aig_Man_t) int32 { return p.nRegs;                                 }
func Aig_ManConstrNum(p *Aig_Man_t) int32 { return p.nConstrs;                              }

func Aig_ManConst0(p *Aig_Man_t) *Aig_Obj_t { return Aig_Not(p.pConst1);                      }
func Aig_ManConst1(p *Aig_Man_t) *Aig_Obj_t { return p.pConst1;                               }
func Aig_ManGhost(p *Aig_Man_t) *Aig_Obj_t { return &p.Ghost;                                }
func Aig_ManCi(p *Aig_Man_t, i int32) *Aig_Obj_t { return (Aig_Obj_t *)Vec_PtrEntry(p.vCis, i);    }
func Aig_ManCo(p *Aig_Man_t, i int32) *Aig_Obj_t { return (Aig_Obj_t *)Vec_PtrEntry(p.vCos, i);    }
func Aig_ManLo(p *Aig_Man_t, i int32) *Aig_Obj_t { return (Aig_Obj_t *)Vec_PtrEntry(p.vCis, Aig_ManCiNum(p)-Aig_ManRegNum(p)+i);   }
func Aig_ManLi(p *Aig_Man_t, i int32) *Aig_Obj_t { return (Aig_Obj_t *)Vec_PtrEntry(p.vCos, Aig_ManCoNum(p)-Aig_ManRegNum(p)+i);   }
func Aig_ManObj(p *Aig_Man_t, i int32) *Aig_Obj_t { if p.vObjs {
    return (Aig_Obj_t *)Vec_PtrEntry(p.vObjs, i);
} else {
    return nil;
}  }

func Aig_ObjType(pObj *Aig_Obj_t) Aig_Type_t { return (Aig_Type_t)pObj.Type;       }
func Aig_ObjIsNone(pObj *Aig_Obj_t) int32 { return pObj.Type == AIG_OBJ_NONE;   }
func Aig_ObjIsConst1(pObj *Aig_Obj_t) int32 { assert(!(Aig_IsComplement(pObj) != 0), "!Aig_IsComplement(pObj)"); return pObj.Type == AIG_OBJ_CONST1; }
func Aig_ObjIsCi(pObj *Aig_Obj_t) int32 { return pObj.Type == AIG_OBJ_CI;     }
func Aig_ObjIsCo(pObj *Aig_Obj_t) int32 { return pObj.Type == AIG_OBJ_CO;     }
func Aig_ObjIsBuf(pObj *Aig_Obj_t) int32 { return pObj.Type == AIG_OBJ_BUF;    }
func Aig_ObjIsAnd(pObj *Aig_Obj_t) int32 { return pObj.Type == AIG_OBJ_AND;    }
func Aig_ObjIsExor(pObj *Aig_Obj_t) int32 { return pObj.Type == AIG_OBJ_EXOR;   }
func Aig_ObjIsNode(pObj *Aig_Obj_t) int32 { return pObj.Type == AIG_OBJ_AND || pObj.Type == AIG_OBJ_EXOR;   }
func Aig_ObjIsTerm(pObj *Aig_Obj_t) int32 { return pObj.Type == AIG_OBJ_CI  || pObj.Type == AIG_OBJ_CO || pObj.Type == AIG_OBJ_CONST1;   }
func Aig_ObjIsHash(pObj *Aig_Obj_t) int32 { return pObj.Type == AIG_OBJ_AND || pObj.Type == AIG_OBJ_EXOR;                                 }
func Aig_ObjIsChoice(p *Aig_Man_t, pObj *Aig_Obj_t) int32 { return p.pEquivs && p.pEquivs[pObj.Id] && pObj.nRefs > 0;                    }
func Aig_ObjIsCand(pObj *Aig_Obj_t) int32 { return pObj.Type == AIG_OBJ_CI || pObj.Type == AIG_OBJ_AND || pObj.Type == AIG_OBJ_EXOR;     }
func Aig_ObjCioId(pObj *Aig_Obj_t) int32 { assert(!(Aig_ObjIsNode(pObj) != 0), "!Aig_ObjIsNode(pObj)"); return pObj.CioId;                                            }
func Aig_ObjId(pObj *Aig_Obj_t) int32 { return pObj.Id;                     }

func Aig_ObjIsMarkA(pObj *Aig_Obj_t) int32 { return pObj.fMarkA;  }
func Aig_ObjSetMarkA(pObj *Aig_Obj_t) { pObj.fMarkA = 1;     }
func Aig_ObjClearMarkA(pObj *Aig_Obj_t) { pObj.fMarkA = 0;     }
 
func Aig_ObjSetTravId(pObj *Aig_Obj_t, TravId int32) { pObj.TravId = TravId;                         }
func Aig_ObjSetTravIdCurrent(p *Aig_Man_t, pObj *Aig_Obj_t) { pObj.TravId = p.nTravIds;                    }
func Aig_ObjSetTravIdPrevious(p *Aig_Man_t, pObj *Aig_Obj_t) { pObj.TravId = p.nTravIds - 1;                }
func Aig_ObjIsTravIdCurrent(p *Aig_Man_t, pObj *Aig_Obj_t) int32 { return int32((pObj.TravId == p.nTravIds));     }
func Aig_ObjIsTravIdPrevious(p *Aig_Man_t, pObj *Aig_Obj_t) int32 { return int32((pObj.TravId == p.nTravIds - 1)); }

func Aig_ObjPhase(pObj *Aig_Obj_t) int32 { return pObj.fPhase;                           }
func Aig_ObjPhaseReal(pObj *Aig_Obj_t) int32 { if pObj != nil {
    return Aig_Regular(pObj).fPhase ^ Aig_IsComplement(pObj);
} else {
    return 1;
}                              }
func Aig_ObjRefs(pObj *Aig_Obj_t) int32 { return pObj.nRefs;                            }
func Aig_ObjRef(pObj *Aig_Obj_t) { pObj.nRefs++;                                 }
func Aig_ObjDeref(pObj *Aig_Obj_t) { assert(pObj.nRefs > 0, "pObj.nRefs > 0"); pObj.nRefs--;      }
func Aig_ObjClearRef(pObj *Aig_Obj_t) { pObj.nRefs = 0;                               }
func Aig_ObjFaninId0(pObj *Aig_Obj_t) int32 { if pObj.pFanin0 {
    return Aig_Regular(pObj.pFanin0).Id;
} else {
    return -1;
} }
func Aig_ObjFaninId1(pObj *Aig_Obj_t) int32 { if pObj.pFanin1 {
    return Aig_Regular(pObj.pFanin1).Id;
} else {
    return -1;
} }
func Aig_ObjFaninC0(pObj *Aig_Obj_t) int32 { return Aig_IsComplement(pObj.pFanin0);        }
func Aig_ObjFaninC1(pObj *Aig_Obj_t) int32 { return Aig_IsComplement(pObj.pFanin1);        }
func Aig_ObjFanin0(pObj *Aig_Obj_t) *Aig_Obj_t { return Aig_Regular(pObj.pFanin0);             }
func Aig_ObjFanin1(pObj *Aig_Obj_t) *Aig_Obj_t { return Aig_Regular(pObj.pFanin1);             }
func Aig_ObjChild0(pObj *Aig_Obj_t) *Aig_Obj_t { return pObj.pFanin0;                          }
func Aig_ObjChild1(pObj *Aig_Obj_t) *Aig_Obj_t { return pObj.pFanin1;                          }
func Aig_ObjChild0Copy(pObj *Aig_Obj_t) *Aig_Obj_t { assert(!(Aig_IsComplement(pObj) != 0), "!Aig_IsComplement(pObj)"); if Aig_ObjFanin0(pObj) != nil {
    return Aig_NotCond((Aig_Obj_t *)Aig_ObjFanin0(pObj).pData, Aig_ObjFaninC0(pObj));
} else {
    return nil;
}  }
func Aig_ObjChild1Copy(pObj *Aig_Obj_t) *Aig_Obj_t { assert(!(Aig_IsComplement(pObj) != 0), "!Aig_IsComplement(pObj)"); if Aig_ObjFanin1(pObj) != nil {
    return Aig_NotCond((Aig_Obj_t *)Aig_ObjFanin1(pObj).pData, Aig_ObjFaninC1(pObj));
} else {
    return nil;
}  }
func Aig_ObjChild0Next(pObj *Aig_Obj_t) *Aig_Obj_t { assert(!(Aig_IsComplement(pObj) != 0), "!Aig_IsComplement(pObj)"); if Aig_ObjFanin0(pObj) != nil {
    return Aig_NotCond((Aig_Obj_t *)Aig_ObjFanin0(pObj).pNext, Aig_ObjFaninC0(pObj));
} else {
    return nil;
}  }
func Aig_ObjChild1Next(pObj *Aig_Obj_t) *Aig_Obj_t { assert(!(Aig_IsComplement(pObj) != 0), "!Aig_IsComplement(pObj)"); if Aig_ObjFanin1(pObj) != nil {
    return Aig_NotCond((Aig_Obj_t *)Aig_ObjFanin1(pObj).pNext, Aig_ObjFaninC1(pObj));
} else {
    return nil;
}  }
func Aig_ObjChild0Flip(pObj *Aig_Obj_t) { assert(!(Aig_IsComplement(pObj) != 0), "!Aig_IsComplement(pObj)"); pObj.pFanin0 = Aig_Not(pObj.pFanin0);        }
func Aig_ObjChild1Flip(pObj *Aig_Obj_t) { assert(!(Aig_IsComplement(pObj) != 0), "!Aig_IsComplement(pObj)"); pObj.pFanin1 = Aig_Not(pObj.pFanin1);        }
func Aig_ObjCopy(pObj *Aig_Obj_t) *Aig_Obj_t { assert(!(Aig_IsComplement(pObj) != 0), "!Aig_IsComplement(pObj)"); return (Aig_Obj_t *)pObj.pData;               } 
func Aig_ObjSetCopy(pObj *Aig_Obj_t, pCopy *Aig_Obj_t) {  assert(!(Aig_IsComplement(pObj) != 0), "!Aig_IsComplement(pObj)"); pObj.pData = pCopy;       } 
func Aig_ObjRealCopy(pObj *Aig_Obj_t) *Aig_Obj_t { return Aig_NotCond((Aig_Obj_t *)Aig_Regular(pObj).pData, Aig_IsComplement(pObj));}
func Aig_ObjToLit(pObj *Aig_Obj_t) int32 { return Abc_Var2Lit( Aig_ObjId(Aig_Regular(pObj)), Aig_IsComplement(pObj) );       }
func Aig_ObjFromLit(p *Aig_Man_t, iLit int32) *Aig_Obj_t { return Aig_NotCond( Aig_ManObj(p, Abc_Lit2Var(iLit)), Abc_LitIsCompl(iLit) );     }
func Aig_ObjLevel(pObj *Aig_Obj_t) int32 { assert(!(Aig_IsComplement(pObj) != 0), "!Aig_IsComplement(pObj)"); return pObj.Level;                            }
func Aig_ObjLevelNew(pObj *Aig_Obj_t) int32 { assert(!(Aig_IsComplement(pObj) != 0), "!Aig_IsComplement(pObj)"); if Aig_ObjFanin1(pObj) != nil {
    return 1 + Aig_ObjIsExor(pObj) + Abc_MaxInt(Aig_ObjFanin0(pObj).Level, Aig_ObjFanin1(pObj).Level);
} else {
    return Aig_ObjFanin0(pObj).Level;
} }
func Aig_ObjSetLevel(pObj *Aig_Obj_t, i int32) int32 { assert(!(Aig_IsComplement(pObj) != 0), "!Aig_IsComplement(pObj)"); return pObj.Level = i;                    }
func Aig_ObjClean(pObj *Aig_Obj_t) { *pObj = Aig_Obj_t{};                                                             }
func Aig_ObjFanout0(p *Aig_Man_t, pObj *Aig_Obj_t) *Aig_Obj_t { assert(p.pFanData && pObj.Id < p.nFansAlloc, "p.pFanData && pObj.Id < p.nFansAlloc"); return Aig_ManObj(p, p.pFanData[5*pObj.Id] >> 1); } 
func Aig_ObjEquiv(p *Aig_Man_t, pObj *Aig_Obj_t) *Aig_Obj_t { if p.pEquivs {
    return p.pEquivs[pObj.Id];
} else {
    return nil;
}           } 
func Aig_ObjSetEquiv(p *Aig_Man_t, pObj *Aig_Obj_t, pEqu *Aig_Obj_t) { assert(p.pEquivs, "p.pEquivs"); p.pEquivs[pObj.Id] = pEqu;                  }
func Aig_ObjRepr(p *Aig_Man_t, pObj *Aig_Obj_t) *Aig_Obj_t { if p.pReprs {
    return p.pReprs[pObj.Id];
} else {
    return nil;
}             } 
func Aig_ObjSetRepr(p *Aig_Man_t, pObj *Aig_Obj_t, pRepr *Aig_Obj_t) { assert(p.pReprs, "p.pReprs"); p.pReprs[pObj.Id] = pRepr;                                } 
func Aig_ObjWhatFanin(pObj *Aig_Obj_t, pFanin *Aig_Obj_t) int32 { 
    if Aig_ObjFanin0(pObj) == pFanin { return 0; } 
    if Aig_ObjFanin1(pObj) == pFanin { return 1; } 
    assert(false, "0"); return -1; 
}
func Aig_ObjFanoutC(pObj *Aig_Obj_t, pFanout *Aig_Obj_t) int32 { 
    if Aig_ObjFanin0(pFanout) == pObj { return Aig_ObjFaninC0(pObj); } 
    if Aig_ObjFanin1(pFanout) == pObj { return Aig_ObjFaninC1(pObj); } 
    assert(false, "0"); return -1; 
}

// create the ghost of the new node
func Aig_ObjCreateGhost(p *Aig_Man_t, p0 *Aig_Obj_t, p1 *Aig_Obj_t, Type Aig_Type_t) *Aig_Obj_t {
    var pGhost *Aig_Obj_t;
    assert(Type != AIG_OBJ_AND || !(Aig_ObjIsConst1(Aig_Regular(p0)) != 0), "Type != AIG_OBJ_AND || !Aig_ObjIsConst1(Aig_Regular(p0))");
    assert(p1 == nil || !(Aig_ObjIsConst1(Aig_Regular(p1)) != 0), "p1 == nil || !Aig_ObjIsConst1(Aig_Regular(p1))");
    assert(Type == AIG_OBJ_CI || Aig_Regular(p0) != Aig_Regular(p1), "Type == AIG_OBJ_CI || Aig_Regular(p0) != Aig_Regular(p1)");
    pGhost = Aig_ManGhost(p);
    pGhost.Type = Type;
    if p1 == nil || Aig_Regular(p0).Id < Aig_Regular(p1).Id {
        pGhost.pFanin0 = p0;
        pGhost.pFanin1 = p1;
    
    } else {
        pGhost.pFanin0 = p1;
        pGhost.pFanin1 = p0;
    }
    return pGhost;
}

// internal memory manager
func Aig_ManFetchMemory(p *Aig_Man_t) *Aig_Obj_t {
    extern char * Aig_MmFixedEntryFetch( Aig_MmFixed_t * p );
    var pTemp *Aig_Obj_t;
    pTemp = (Aig_Obj_t *)Aig_MmFixedEntryFetch( p.pMemObjs );
    *pTemp = Aig_Obj_t{}; 
    pTemp.Id = Vec_PtrSize(p.vObjs);
    Vec_PtrPush( p.vObjs, pTemp );
    return pTemp;
}
func Aig_ManRecycleMemory(p *Aig_Man_t, pEntry *Aig_Obj_t) {
    extern void Aig_MmFixedEntryRecycle( Aig_MmFixed_t * p, char * pEntry );
    assert(pEntry.nRefs == 0, "pEntry.nRefs == 0");
    pEntry.Type = AIG_OBJ_NONE; // distinquishes a dead node from a live node
    Aig_MmFixedEntryRecycle( p.pMemObjs, (char *)pEntry );
    p.nDeleted++;
}


////////////////////////////////////////////////////////////////////////
///                             ITERATORS                            ///
////////////////////////////////////////////////////////////////////////

// iterator over the combinational inputs
// Original macro:
// #define Aig_ManForEachCi( p, pObj, i )                                          \
//     Vec_PtrForEachEntry( Aig_Obj_t *, p->vCis, pObj, i )

// Original macro:
// #define Aig_ManForEachCiReverse( p, pObj, i )                                   \
//     Vec_PtrForEachEntryReverse( Aig_Obj_t *, p->vCis, pObj, i )

// iterator over the combinational outputs
// Original macro:
// #define Aig_ManForEachCo( p, pObj, i )                                          \
//     Vec_PtrForEachEntry( Aig_Obj_t *, p->vCos, pObj, i )

// Original macro:
// #define Aig_ManForEachCoReverse( p, pObj, i )                                   \
//     Vec_PtrForEachEntryReverse( Aig_Obj_t *, p->vCos, pObj, i )

// iterators over all objects, including those currently not used
// Original macro:
// #define Aig_ManForEachObj( p, pObj, i )                                         \
//     Vec_PtrForEachEntry( Aig_Obj_t *, p->vObjs, pObj, i ) if ( (pObj) == NULL ) {} else

// Original macro:
// #define Aig_ManForEachObjReverse( p, pObj, i )                                  \
//     Vec_PtrForEachEntryReverse( Aig_Obj_t *, p->vObjs, pObj, i ) if ( (pObj) == NULL ) {} else

// iterators over the objects whose IDs are stored in an array
// Original macro:
// #define Aig_ManForEachObjVec( vIds, p, pObj, i )                                \
//     for ( i = 0; i < Vec_IntSize(vIds) && (((pObj) = Aig_ManObj(p, Vec_IntEntry(vIds,i))), 1); i++ )

// Original macro:
// #define Aig_ManForEachObjVecReverse( vIds, p, pObj, i )                         \
//     for ( i = Vec_IntSize(vIds) - 1; i >= 0 && (((pObj) = Aig_ManObj(p, Vec_IntEntry(vIds,i))), 1); i-- )

// iterators over all nodes
// Original macro:
// #define Aig_ManForEachNode( p, pObj, i )                                        \
//     Vec_PtrForEachEntry( Aig_Obj_t *, p->vObjs, pObj, i ) if ( (pObj) == NULL || !Aig_ObjIsNode(pObj) ) {} else

// Original macro:
// #define Aig_ManForEachNodeReverse( p, pObj, i )                                 \
//     Vec_PtrForEachEntryReverse( Aig_Obj_t *, p->vObjs, pObj, i ) if ( (pObj) == NULL || !Aig_ObjIsNode(pObj) ) {} else

// iterator over all nodes
// Original macro:
// #define Aig_ManForEachExor( p, pObj, i )                                        \
//     Vec_PtrForEachEntry( Aig_Obj_t *, p->vObjs, pObj, i ) if ( (pObj) == NULL || !Aig_ObjIsExor(pObj) ) {} else

// Original macro:
// #define Aig_ManForEachExorReverse( p, pObj, i )                                 \
//     Vec_PtrForEachEntryReverse( Aig_Obj_t *, p->vObjs, pObj, i ) if ( (pObj) == NULL || !Aig_ObjIsExor(pObj) ) {} else


// these two procedures are only here for the use inside the iterator
func Aig_ObjFanout0Int(p *Aig_Man_t, ObjId int32) int32 { assert(ObjId < p.nFansAlloc, "ObjId < p.nFansAlloc");  return p.pFanData[5*ObjId];                         }
func Aig_ObjFanoutNext(p *Aig_Man_t, iFan int32) int32 { assert(iFan/2 < p.nFansAlloc, "iFan/2 < p.nFansAlloc"); return p.pFanData[5*(iFan >> 1) + 3 + (iFan & 1)];  }
// iterator over the fanouts
// Original macro:
// #define Aig_ObjForEachFanout( p, pObj, pFanout, iFan, i )                       \
//     for ( assert(p->pFanData), i = 0; (i < (int)(pObj)->nRefs) &&               \
//           (((iFan) = i? Aig_ObjFanoutNext(p, iFan) : Aig_ObjFanout0Int(p, pObj->Id)), 1) && \
//           (((pFanout) = Aig_ManObj(p, iFan>>1)), 1); i++ )



////////////////////////////////////////////////////////////////////////
///                     SEQUENTIAL ITERATORS                         ///
////////////////////////////////////////////////////////////////////////

// iterator over the primary inputs
// Original macro:
// #define Aig_ManForEachPiSeq( p, pObj, i )                                       \
//     Vec_PtrForEachEntryStop( Aig_Obj_t *, p->vCis, pObj, i, Aig_ManCiNum(p)-Aig_ManRegNum(p) )

// iterator over the latch outputs
// Original macro:
// #define Aig_ManForEachLoSeq( p, pObj, i )                                       \
//     Vec_PtrForEachEntryStart( Aig_Obj_t *, p->vCis, pObj, i, Aig_ManCiNum(p)-Aig_ManRegNum(p) )

// iterator over the primary outputs
// Original macro:
// #define Aig_ManForEachPoSeq( p, pObj, i )                                       \
//     Vec_PtrForEachEntryStop( Aig_Obj_t *, p->vCos, pObj, i, Aig_ManCoNum(p)-Aig_ManRegNum(p) )

// iterator over the latch inputs
// Original macro:
// #define Aig_ManForEachLiSeq( p, pObj, i )                                       \
//     Vec_PtrForEachEntryStart( Aig_Obj_t *, p->vCos, pObj, i, Aig_ManCoNum(p)-Aig_ManRegNum(p) )

// iterator over the latch input and outputs
// Original macro:
// #define Aig_ManForEachLiLoSeq( p, pObjLi, pObjLo, k )                           \
//     for ( k = 0; (k < Aig_ManRegNum(p)) && (((pObjLi) = Aig_ManLi(p, k)), 1)    \
//         && (((pObjLo)=Aig_ManLo(p, k)), 1); k++ )


////////////////////////////////////////////////////////////////////////
///                    FUNCTION DECLARATIONS                         ///
////////////////////////////////////////////////////////////////////////

/*=== aigCheck.c ========================================================*/
extern ABC_DLL int     Aig_ManCheck( Aig_Man_t * p );
// extern void            Aig_ManCheckMarkA( Aig_Man_t * p );

// extern void            Aig_ManCheckPhase( Aig_Man_t * p );

/*=== aigCuts.c ========================================================*/
// extern Aig_ManCut_t *  Aig_ComputeCuts( Aig_Man_t * pAig, int nCutsMax, int nLeafMax, int fTruth, int fVerbose );

// extern void            Aig_ManCutStop( Aig_ManCut_t * p );

/*=== aigDfs.c ==========================================================*/
// extern int             Aig_ManVerifyTopoOrder( Aig_Man_t * p );

// extern Vec_Ptr_t *     Aig_ManDfs( Aig_Man_t * p, int fNodesOnly );

// extern Vec_Ptr_t *     Aig_ManDfsAll( Aig_Man_t * p );

// extern Vec_Ptr_t *     Aig_ManDfsPreorder( Aig_Man_t * p, int fNodesOnly );

// extern Vec_Vec_t *     Aig_ManLevelize( Aig_Man_t * p );

// extern Vec_Ptr_t *     Aig_ManDfsNodes( Aig_Man_t * p, Aig_Obj_t ** ppNodes, int nNodes );

// extern Vec_Ptr_t *     Aig_ManDfsChoices( Aig_Man_t * p );

// extern Vec_Ptr_t *     Aig_ManDfsReverse( Aig_Man_t * p );

// extern int             Aig_ManLevelNum( Aig_Man_t * p );

// extern int             Aig_ManChoiceLevel( Aig_Man_t * p );

// extern int             Aig_DagSize( Aig_Obj_t * pObj );

// extern int             Aig_SupportSize( Aig_Man_t * p, Aig_Obj_t * pObj );

// extern Vec_Ptr_t *     Aig_Support( Aig_Man_t * p, Aig_Obj_t * pObj );

// extern void            Aig_SupportNodes( Aig_Man_t * p, Aig_Obj_t ** ppObjs, int nObjs, Vec_Ptr_t * vSupp );

// extern void            Aig_ConeUnmark_rec( Aig_Obj_t * pObj );

// extern Aig_Obj_t *     Aig_Transfer( Aig_Man_t * pSour, Aig_Man_t * pDest, Aig_Obj_t * pObj, int nVars );

// extern Aig_Obj_t *     Aig_Compose( Aig_Man_t * p, Aig_Obj_t * pRoot, Aig_Obj_t * pFunc, int iVar );

// extern void            Aig_ObjCollectCut( Aig_Obj_t * pRoot, Vec_Ptr_t * vLeaves, Vec_Ptr_t * vNodes );

// extern int             Aig_ObjCollectSuper( Aig_Obj_t * pObj, Vec_Ptr_t * vSuper );

/*=== aigDup.c ==========================================================*/
// extern Aig_Obj_t *     Aig_ManDupSimpleDfs_rec( Aig_Man_t * pNew, Aig_Man_t * p, Aig_Obj_t * pObj );

// extern Aig_Man_t *     Aig_ManDupSimple( Aig_Man_t * p );

// extern Aig_Man_t *     Aig_ManDupSimpleWithHints( Aig_Man_t * p, Vec_Int_t * vHints );

// extern Aig_Man_t *     Aig_ManDupSimpleDfs( Aig_Man_t * p );

// extern Aig_Man_t *     Aig_ManDupSimpleDfsPart( Aig_Man_t * p, Vec_Ptr_t * vPis, Vec_Ptr_t * vCos );

// extern Aig_Man_t *     Aig_ManDupOrdered( Aig_Man_t * p );

// extern Aig_Man_t *     Aig_ManDupCof( Aig_Man_t * p, int iInput, int Value );

// extern Aig_Man_t *     Aig_ManDupTrim( Aig_Man_t * p );

// extern Aig_Man_t *     Aig_ManDupExor( Aig_Man_t * p );

// extern Aig_Man_t *     Aig_ManDupDfs( Aig_Man_t * p );

// extern Vec_Ptr_t *     Aig_ManOrderPios( Aig_Man_t * p, Aig_Man_t * pOrder );

// extern Aig_Man_t *     Aig_ManDupDfsGuided( Aig_Man_t * p, Vec_Ptr_t * vPios );

// extern Aig_Man_t *     Aig_ManDupLevelized( Aig_Man_t * p );

// extern Aig_Man_t *     Aig_ManDupWithoutPos( Aig_Man_t * p );

// extern Aig_Man_t *     Aig_ManDupFlopsOnly( Aig_Man_t * p );

// extern Aig_Man_t *     Aig_ManDupRepres( Aig_Man_t * p );

// extern Aig_Man_t *     Aig_ManDupRepresDfs( Aig_Man_t * p );

// extern Aig_Man_t *     Aig_ManCreateMiter( Aig_Man_t * p1, Aig_Man_t * p2, int fImpl );

// extern Aig_Man_t *     Aig_ManDupOrpos( Aig_Man_t * p, int fAddRegs );

// extern Aig_Man_t *     Aig_ManDupOneOutput( Aig_Man_t * p, int iPoNum, int fAddRegs );

// extern Aig_Man_t *     Aig_ManDupUnsolvedOutputs( Aig_Man_t * p, int fAddRegs );

// extern Aig_Man_t *     Aig_ManDupArray( Vec_Ptr_t * vArray );

// extern Aig_Man_t *     Aig_ManDupNodes( Aig_Man_t * pMan, Vec_Ptr_t * vArray );

/*=== aigFanout.c ==========================================================*/
// extern void            Aig_ObjAddFanout( Aig_Man_t * p, Aig_Obj_t * pObj, Aig_Obj_t * pFanout );

// extern void            Aig_ObjRemoveFanout( Aig_Man_t * p, Aig_Obj_t * pObj, Aig_Obj_t * pFanout );

// extern void            Aig_ManFanoutStart( Aig_Man_t * p );

// extern void            Aig_ManFanoutStop( Aig_Man_t * p );

/*=== aigFrames.c ==========================================================*/
// extern Aig_Man_t *     Aig_ManFrames( Aig_Man_t * pAig, int nFs, int fInit, int fOuts, int fRegs, int fEnlarge, Aig_Obj_t *** ppObjMap );

/*=== aigMan.c ==========================================================*/
// extern Aig_Man_t *     Aig_ManStart( int nNodesMax );

// extern Aig_Man_t *     Aig_ManStartFrom( Aig_Man_t * p );

// extern Aig_Man_t *     Aig_ManExtractMiter( Aig_Man_t * p, Aig_Obj_t * pNode1, Aig_Obj_t * pNode2 );

// extern void            Aig_ManStop( Aig_Man_t * p );

// extern void            Aig_ManStopP( Aig_Man_t ** p );

// extern int             Aig_ManCleanup( Aig_Man_t * p );

// extern int             Aig_ManAntiCleanup( Aig_Man_t * p );

// extern int             Aig_ManCiCleanup( Aig_Man_t * p );

// extern int             Aig_ManCoCleanup( Aig_Man_t * p );

// extern void            Aig_ManPrintStats( Aig_Man_t * p );

// extern void            Aig_ManReportImprovement( Aig_Man_t * p, Aig_Man_t * pNew );

// extern void            Aig_ManSetRegNum( Aig_Man_t * p, int nRegs );

// extern void            Aig_ManFlipFirstPo( Aig_Man_t * p );

// extern void *          Aig_ManReleaseData( Aig_Man_t * p );

/*=== aigMem.c ==========================================================*/
// extern void            Aig_ManStartMemory( Aig_Man_t * p );

// extern void            Aig_ManStopMemory( Aig_Man_t * p );

/*=== aigMffc.c ==========================================================*/
// extern int             Aig_NodeRef_rec( Aig_Obj_t * pNode, unsigned LevelMin );

// extern int             Aig_NodeDeref_rec( Aig_Obj_t * pNode, unsigned LevelMin, float * pPower, float * pProbs );

// extern int             Aig_NodeMffcSupp( Aig_Man_t * p, Aig_Obj_t * pNode, int LevelMin, Vec_Ptr_t * vSupp );

// extern int             Aig_NodeMffcLabel( Aig_Man_t * p, Aig_Obj_t * pNode, float * pPower );

// extern int             Aig_NodeMffcLabelCut( Aig_Man_t * p, Aig_Obj_t * pNode, Vec_Ptr_t * vLeaves );

// extern int             Aig_NodeMffcExtendCut( Aig_Man_t * p, Aig_Obj_t * pNode, Vec_Ptr_t * vLeaves, Vec_Ptr_t * vResult );

/*=== aigObj.c ==========================================================*/
// extern Aig_Obj_t *     Aig_ObjCreateCi( Aig_Man_t * p );

// extern Aig_Obj_t *     Aig_ObjCreateCo( Aig_Man_t * p, Aig_Obj_t * pDriver );

// extern Aig_Obj_t *     Aig_ObjCreate( Aig_Man_t * p, Aig_Obj_t * pGhost );

// extern void            Aig_ObjConnect( Aig_Man_t * p, Aig_Obj_t * pObj, Aig_Obj_t * pFan0, Aig_Obj_t * pFan1 );

// extern void            Aig_ObjDisconnect( Aig_Man_t * p, Aig_Obj_t * pObj );

// extern void            Aig_ObjDelete( Aig_Man_t * p, Aig_Obj_t * pObj );

// extern void            Aig_ObjDelete_rec( Aig_Man_t * p, Aig_Obj_t * pObj, int fFreeTop );

// extern void            Aig_ObjDeletePo( Aig_Man_t * p, Aig_Obj_t * pObj );

// extern void            Aig_ObjPrint( Aig_Man_t * p, Aig_Obj_t * pObj );

// extern void            Aig_ObjPatchFanin0( Aig_Man_t * p, Aig_Obj_t * pObj, Aig_Obj_t * pFaninNew );

// extern void            Aig_ObjReplace( Aig_Man_t * p, Aig_Obj_t * pObjOld, Aig_Obj_t * pObjNew, int fUpdateLevel );

/*=== aigOper.c =========================================================*/
// extern Aig_Obj_t *     Aig_IthVar( Aig_Man_t * p, int i );

// extern Aig_Obj_t *     Aig_Oper( Aig_Man_t * p, Aig_Obj_t * p0, Aig_Obj_t * p1, Aig_Type_t Type );

// extern Aig_Obj_t *     Aig_And( Aig_Man_t * p, Aig_Obj_t * p0, Aig_Obj_t * p1 );

// extern Aig_Obj_t *     Aig_Or( Aig_Man_t * p, Aig_Obj_t * p0, Aig_Obj_t * p1 );

// extern Aig_Obj_t *     Aig_Exor( Aig_Man_t * p, Aig_Obj_t * p0, Aig_Obj_t * p1 );

// extern Aig_Obj_t *     Aig_Mux( Aig_Man_t * p, Aig_Obj_t * pC, Aig_Obj_t * p1, Aig_Obj_t * p0 );

// extern Aig_Obj_t *     Aig_Maj( Aig_Man_t * p, Aig_Obj_t * pA, Aig_Obj_t * pB, Aig_Obj_t * pC );

// extern Aig_Obj_t *     Aig_Multi( Aig_Man_t * p, Aig_Obj_t ** pArgs, int nArgs, Aig_Type_t Type );

// extern Aig_Obj_t *     Aig_Miter( Aig_Man_t * p, Vec_Ptr_t * vPairs );

// extern Aig_Obj_t *     Aig_MiterTwo( Aig_Man_t * p, Vec_Ptr_t * vNodes1, Vec_Ptr_t * vNodes2 );

// extern Aig_Obj_t *     Aig_CreateAnd( Aig_Man_t * p, int nVars );

// extern Aig_Obj_t *     Aig_CreateOr( Aig_Man_t * p, int nVars );

// extern Aig_Obj_t *     Aig_CreateExor( Aig_Man_t * p, int nVars );

/*=== aigOrder.c =========================================================*/
// extern void            Aig_ManOrderStart( Aig_Man_t * p );

// extern void            Aig_ManOrderStop( Aig_Man_t * p );

// extern void            Aig_ObjOrderInsert( Aig_Man_t * p, int ObjId );

// extern void            Aig_ObjOrderRemove( Aig_Man_t * p, int ObjId );

// extern void            Aig_ObjOrderAdvance( Aig_Man_t * p );

/*=== aigPart.c =========================================================*/
// extern Vec_Ptr_t *     Aig_ManSupports( Aig_Man_t * p );

// extern Vec_Ptr_t *     Aig_ManSupportsInverse( Aig_Man_t * p );

// extern Vec_Ptr_t *     Aig_ManSupportsRegisters( Aig_Man_t * p );

// extern Vec_Ptr_t *     Aig_ManPartitionSmart( Aig_Man_t * p, int nPartSizeLimit, int fVerbose, Vec_Ptr_t ** pvPartSupps );

// extern Vec_Ptr_t *     Aig_ManPartitionSmartRegisters( Aig_Man_t * pAig, int nSuppSizeLimit, int fVerbose );

// extern Vec_Ptr_t *     Aig_ManPartitionNaive( Aig_Man_t * p, int nPartSize );

// extern Vec_Ptr_t *     Aig_ManMiterPartitioned( Aig_Man_t * p1, Aig_Man_t * p2, int nPartSize, int fSmart );

// extern Aig_Man_t *     Aig_ManChoicePartitioned( Vec_Ptr_t * vAigs, int nPartSize, int nConfMax, int nLevelMax, int fVerbose );

// extern Aig_Man_t *     Aig_ManFraigPartitioned( Aig_Man_t * pAig, int nPartSize, int nConfMax, int nLevelMax, int fVerbose );

// extern Aig_Man_t *     Aig_ManChoiceConstructive( Vec_Ptr_t * vAigs, int fVerbose );

/*=== aigPartReg.c =========================================================*/
// extern Vec_Ptr_t *     Aig_ManRegPartitionSimple( Aig_Man_t * pAig, int nPartSize, int nOverSize );

// extern void            Aig_ManPartDivide( Vec_Ptr_t * vResult, Vec_Int_t * vDomain, int nPartSize, int nOverSize );

// extern Vec_Ptr_t *     Aig_ManRegPartitionSmart( Aig_Man_t * pAig, int nPartSize );

// extern Aig_Man_t *     Aig_ManRegCreatePart( Aig_Man_t * pAig, Vec_Int_t * vPart, int * pnCountPis, int * pnCountRegs, int ** ppMapBack );

// extern Vec_Ptr_t *     Aig_ManRegProjectOnehots( Aig_Man_t * pAig, Aig_Man_t * pPart, Vec_Ptr_t * vOnehots, int fVerbose );

/*=== aigRepr.c =========================================================*/
// extern void            Aig_ManReprStart( Aig_Man_t * p, int nIdMax );

// extern void            Aig_ManReprStop( Aig_Man_t * p );

// extern void            Aig_ObjCreateRepr( Aig_Man_t * p, Aig_Obj_t * pNode1, Aig_Obj_t * pNode2 );

// extern void            Aig_ManTransferRepr( Aig_Man_t * pNew, Aig_Man_t * p );

// extern Aig_Man_t *     Aig_ManDupRepr( Aig_Man_t * p, int fOrdered );

// extern Aig_Man_t *     Aig_ManDupReprBasic( Aig_Man_t * p );

// extern int             Aig_ManCountReprs( Aig_Man_t * p );

// extern Aig_Man_t *     Aig_ManRehash( Aig_Man_t * p );

// extern int             Aig_ObjCheckTfi( Aig_Man_t * p, Aig_Obj_t * pNew, Aig_Obj_t * pOld );

// extern void            Aig_ManMarkValidChoices( Aig_Man_t * p );

// extern int             Aig_TransferMappedClasses( Aig_Man_t * pAig, Aig_Man_t * pPart, int * pMapBack );

/*=== aigRet.c ========================================================*/
// extern Aig_Man_t *     Rtm_ManRetime( Aig_Man_t * p, int fForward, int nStepsMax, int fVerbose );

/*=== aigRetF.c ========================================================*/
// extern Aig_Man_t *     Aig_ManRetimeFrontier( Aig_Man_t * p, int nStepsMax );

/*=== aigScl.c ==========================================================*/
// extern Aig_Man_t *     Aig_ManRemap( Aig_Man_t * p, Vec_Ptr_t * vMap );

// extern int             Aig_ManSeqCleanup( Aig_Man_t * p );

// extern int             Aig_ManSeqCleanupBasic( Aig_Man_t * p );

// extern int             Aig_ManCountMergeRegs( Aig_Man_t * p );

// extern Aig_Man_t *     Aig_ManReduceLaches( Aig_Man_t * p, int fVerbose );

// extern void            Aig_ManComputeSccs( Aig_Man_t * p );
 
// extern Aig_Man_t *     Aig_ManScl( Aig_Man_t * pAig, int fLatchConst, int fLatchEqual, int fUseMvSweep, int nFramesSymb, int nFramesSatur, int fVerbose, int fVeryVerbose );

/*=== aigShow.c ========================================================*/
// extern void            Aig_ManShow( Aig_Man_t * pMan, int fHaig, Vec_Ptr_t * vBold );

/*=== aigTable.c ========================================================*/
// extern Aig_Obj_t *     Aig_TableLookup( Aig_Man_t * p, Aig_Obj_t * pGhost );

// extern Aig_Obj_t *     Aig_TableLookupTwo( Aig_Man_t * p, Aig_Obj_t * pFanin0, Aig_Obj_t * pFanin1 );

// extern void            Aig_TableInsert( Aig_Man_t * p, Aig_Obj_t * pObj );

// extern void            Aig_TableDelete( Aig_Man_t * p, Aig_Obj_t * pObj );

// extern int             Aig_TableCountEntries( Aig_Man_t * p );

// extern void            Aig_TableProfile( Aig_Man_t * p );

// extern void            Aig_TableClear( Aig_Man_t * p );

/*=== aigTiming.c ========================================================*/
// extern void            Aig_ObjClearReverseLevel( Aig_Man_t * p, Aig_Obj_t * pObj );

// extern int             Aig_ObjRequiredLevel( Aig_Man_t * p, Aig_Obj_t * pObj );

// extern void            Aig_ManStartReverseLevels( Aig_Man_t * p, int nMaxLevelIncrease );

// extern void            Aig_ManStopReverseLevels( Aig_Man_t * p );

// extern void            Aig_ManUpdateLevel( Aig_Man_t * p, Aig_Obj_t * pObjNew );

// extern void            Aig_ManUpdateReverseLevel( Aig_Man_t * p, Aig_Obj_t * pObjNew );

// extern void            Aig_ManVerifyLevel( Aig_Man_t * p );

// extern void            Aig_ManVerifyReverseLevel( Aig_Man_t * p );

/*=== aigTruth.c ========================================================*/
// extern unsigned *      Aig_ManCutTruth( Aig_Obj_t * pRoot, Vec_Ptr_t * vLeaves, Vec_Ptr_t * vNodes, Vec_Ptr_t * vTruthElem, Vec_Ptr_t * vTruthStore );

/*=== aigTsim.c ========================================================*/
// extern Aig_Man_t *     Aig_ManConstReduce( Aig_Man_t * p, int fUseMvSweep, int nFramesSymb, int nFramesSatur, int fVerbose, int fVeryVerbose );

/*=== aigUtil.c =========================================================*/
// extern void            Aig_ManIncrementTravId( Aig_Man_t * p );

// extern char *          Aig_TimeStamp();

// extern int             Aig_ManHasNoGaps( Aig_Man_t * p );

// extern int             Aig_ManLevels( Aig_Man_t * p );

// extern void            Aig_ManResetRefs( Aig_Man_t * p );

// extern void            Aig_ManCleanMarkA( Aig_Man_t * p );

// extern void            Aig_ManCleanMarkB( Aig_Man_t * p );

// extern void            Aig_ManCleanMarkAB( Aig_Man_t * p );

// extern void            Aig_ManCleanData( Aig_Man_t * p );

// extern void            Aig_ObjCleanData_rec( Aig_Obj_t * pObj );

// extern void            Aig_ManCleanNext( Aig_Man_t * p );

// extern void            Aig_ObjCollectMulti( Aig_Obj_t * pFunc, Vec_Ptr_t * vSuper );

// extern int             Aig_ObjIsMuxType( Aig_Obj_t * pObj );

// extern int             Aig_ObjRecognizeExor( Aig_Obj_t * pObj, Aig_Obj_t ** ppFan0, Aig_Obj_t ** ppFan1 );

// extern Aig_Obj_t *     Aig_ObjRecognizeMux( Aig_Obj_t * pObj, Aig_Obj_t ** ppObjT, Aig_Obj_t ** ppObjE );

// extern Aig_Obj_t *     Aig_ObjReal_rec( Aig_Obj_t * pObj );

// extern int             Aig_ObjCompareIdIncrease( Aig_Obj_t ** pp1, Aig_Obj_t ** pp2 );

// extern void            Aig_ObjPrintEqn( FILE * pFile, Aig_Obj_t * pObj, Vec_Vec_t * vLevels, int Level );

// extern void            Aig_ObjPrintVerilog( FILE * pFile, Aig_Obj_t * pObj, Vec_Vec_t * vLevels, int Level );

// extern void            Aig_ObjPrintVerbose( Aig_Obj_t * pObj, int fHaig );

// extern void            Aig_ManPrintVerbose( Aig_Man_t * p, int fHaig );

// extern void            Aig_ManDump( Aig_Man_t * p );

// extern void            Aig_ManDumpBlif( Aig_Man_t * p, char * pFileName, Vec_Ptr_t * vPiNames, Vec_Ptr_t * vPoNames );

// extern void            Aig_ManDumpVerilog( Aig_Man_t * p, char * pFileName );

// extern void            Aig_ManSetCioIds( Aig_Man_t * p );

// extern void            Aig_ManCleanCioIds( Aig_Man_t * p );

// extern int             Aig_ManChoiceNum( Aig_Man_t * p );

// extern char *          Aig_FileNameGenericAppend( char * pBase, char * pSuffix );

// extern unsigned        Aig_ManRandom( int fReset );

// extern word            Aig_ManRandom64( int fReset );

// extern void            Aig_ManRandomInfo( Vec_Ptr_t * vInfo, int iInputStart, int iWordStart, int iWordStop );

// extern void            Aig_NodeUnionLists( Vec_Ptr_t * vArr1, Vec_Ptr_t * vArr2, Vec_Ptr_t * vArr );

// extern void            Aig_NodeIntersectLists( Vec_Ptr_t * vArr1, Vec_Ptr_t * vArr2, Vec_Ptr_t * vArr );

// extern void            Aig_ManSetPhase( Aig_Man_t * pAig );

// extern Vec_Ptr_t *     Aig_ManMuxesCollect( Aig_Man_t * pAig );

// extern void            Aig_ManMuxesDeref( Aig_Man_t * pAig, Vec_Ptr_t * vMuxes );

// extern void            Aig_ManMuxesRef( Aig_Man_t * pAig, Vec_Ptr_t * vMuxes );

// extern void            Aig_ManInvertConstraints( Aig_Man_t * pAig );


/*=== aigWin.c =========================================================*/
// extern void            Aig_ManFindCut( Aig_Obj_t * pRoot, Vec_Ptr_t * vFront, Vec_Ptr_t * vVisited, int nSizeLimit, int nFanoutLimit );

 
/*=== aigMem.c ===========================================================*/
// fixed-size-block memory manager
// extern Aig_MmFixed_t * Aig_MmFixedStart( int nEntrySize, int nEntriesMax );

// extern void            Aig_MmFixedStop( Aig_MmFixed_t * p, int fVerbose );

// extern char *          Aig_MmFixedEntryFetch( Aig_MmFixed_t * p );

// extern void            Aig_MmFixedEntryRecycle( Aig_MmFixed_t * p, char * pEntry );

// extern void            Aig_MmFixedRestart( Aig_MmFixed_t * p );

// extern int             Aig_MmFixedReadMemUsage( Aig_MmFixed_t * p );

// extern int             Aig_MmFixedReadMaxEntriesUsed( Aig_MmFixed_t * p );

// flexible-size-block memory manager
// extern Aig_MmFlex_t *  Aig_MmFlexStart();

// extern void            Aig_MmFlexStop( Aig_MmFlex_t * p, int fVerbose );

// extern char *          Aig_MmFlexEntryFetch( Aig_MmFlex_t * p, int nBytes );

// extern void            Aig_MmFlexRestart( Aig_MmFlex_t * p );

// extern int             Aig_MmFlexReadMemUsage( Aig_MmFlex_t * p );

// hierarchical memory manager
// extern Aig_MmStep_t *  Aig_MmStepStart( int nSteps );

// extern void            Aig_MmStepStop( Aig_MmStep_t * p, int fVerbose );

// extern char *          Aig_MmStepEntryFetch( Aig_MmStep_t * p, int nBytes );

// extern void            Aig_MmStepEntryRecycle( Aig_MmStep_t * p, char * pEntry, int nBytes );

// extern int             Aig_MmStepReadMemUsage( Aig_MmStep_t * p );




//ABC_NAMESPACE_HEADER_END





////////////////////////////////////////////////////////////////////////
///                       END OF FILE                                ///
////////////////////////////////////////////////////////////////////////

