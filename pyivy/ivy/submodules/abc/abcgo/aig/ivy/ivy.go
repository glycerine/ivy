package ivy

/**CFile****************************************************************

  FileName    [ivy.h]

  SystemName  [ABC: Logic synthesis and verification system.]

  PackageName [And-Inverter Graph package.]

  Synopsis    [External declarations.]

  Author      [Alan Mishchenko]
  
  Affiliation [UC Berkeley]

  Date        [Ver. 1.0. Started - May 11, 2006.]

  Revision    [$Id: ivy.c,v 1.00 2006/05/11 00:00:00 alanmi Exp $]

***********************************************************************/





////////////////////////////////////////////////////////////////////////
///                          INCLUDES                                ///
////////////////////////////////////////////////////////////////////////

//#include <stdio.h>
//#include "misc/extra/extra.h"
//#include "misc/vec/vec.h"

////////////////////////////////////////////////////////////////////////
///                         PARAMETERS                               ///
////////////////////////////////////////////////////////////////////////



//ABC_NAMESPACE_HEADER_START
 

////////////////////////////////////////////////////////////////////////
///                         BASIC TYPES                              ///
////////////////////////////////////////////////////////////////////////

type Ivy_Man_t =            Ivy_Man_t_;
type Ivy_Obj_t =            Ivy_Obj_t_;
typedef int                          Ivy_Edge_t;
type Ivy_FraigParams_t =    Ivy_FraigParams_t_;

// object types
 type Ivy_Type_t int32
const ( 
    IVY_NONE Ivy_Type_t = iota
                        // 0: non-existent object
    IVY_PI
                          // 1: primary input (and constant 1 node)
    IVY_PO
                          // 2: primary output
    IVY_ASSERT
                      // 3: assertion
    IVY_LATCH
                       // 4: sequential element
    IVY_AND
                         // 5: AND node
    IVY_EXOR
                        // 6: EXOR node
    IVY_BUF
                         // 7: buffer (temporary)
    IVY_VOID                         // 8: unused object

)

 

// latch initial values
 type Ivy_Init_t int32
const ( 
    IVY_INIT_NONE Ivy_Init_t = iota
                   // 0: not a latch
    IVY_INIT_0
                      // 1: zero
    IVY_INIT_1
                      // 2: one
    IVY_INIT_DC                      // 3: don't-care

)

 

// the AIG node
type Ivy_Obj_t_ struct {  // 24 bytes (32-bit) or 32 bytes (64-bit)   // 10 words - 16 words

    Id              int32;             // integer ID
    TravId              int32;         // traversal ID
    Type uint32; /* C bitfield: 4 bits; review packing and truncation */  // object type
    fMarkA uint32; /* C bitfield: 1 bits; review packing and truncation */  // multipurpose mask
    fMarkB uint32; /* C bitfield: 1 bits; review packing and truncation */  // multipurpose mask
    fExFan uint32; /* C bitfield: 1 bits; review packing and truncation */  // set to 1 if last fanout added is EXOR
    fPhase uint32; /* C bitfield: 1 bits; review packing and truncation */  // value under 000...0 pattern
    fFailTfo uint32; /* C bitfield: 1 bits; review packing and truncation */  // the TFO of the failed node
    Init uint32; /* C bitfield: 2 bits; review packing and truncation */  // latch initial value
    Level uint32; /* C bitfield: 21 bits; review packing and truncation */  // logic level
    nRefs              int32;          // reference counter
    pFanin0       *Ivy_Obj_t;        // fanin
    pFanin1       *Ivy_Obj_t;        // fanin
    pFanout       *Ivy_Obj_t;        // fanout
    pNextFan0       *Ivy_Obj_t;      // next fanout of the first fanin
    pNextFan1       *Ivy_Obj_t;      // next fanout of the second fanin
    pPrevFan0       *Ivy_Obj_t;      // prev fanout of the first fanin
    pPrevFan1       *Ivy_Obj_t;      // prev fanout of the second fanin
    pEquiv       *Ivy_Obj_t;         // equivalent node
}

// the AIG manager
type Ivy_Man_t_ struct {
    // AIG nodes
    vPis       *Vec_Ptr_t;           // the array of PIs
    vPos       *Vec_Ptr_t;           // the array of POs
    vBufs       *Vec_Ptr_t;          // the array of buffers
    vObjs       *Vec_Ptr_t;          // the array of objects
    pConst1       *Ivy_Obj_t;        // the constant 1 node
    Ghost        Ivy_Obj_t;          // the ghost node
    // AIG node counters
    nObjs [IVY_VOID]int32;// the number of objects by type
    nCreated              int32;       // the number of created objects
    nDeleted              int32;       // the number of deleted objects
    // stuctural hash table
    pTable             *int32;         // structural hash table
    nTableSize              int32;     // structural hash table size
    // various data members
    fCatchExor              int32;     // set to 1 to detect EXORs
    nTravIds              int32;       // the traversal ID
    nLevelMax              int32;      // the maximum level
    vRequired       *Vec_Int_t;      // required times
    fFanout              int32;        // fanout is allocated
    pData            any;          // the temporary data
    pCopy            any;          // the temporary data
    pHaig       *Ivy_Man_t;          // history AIG if present
    nClassesSkip              int32;   // the number of skipped classes
    // memory management
    vChunks       *Vec_Ptr_t;        // allocated memory pieces
    vPages       *Vec_Ptr_t;         // memory pages used by nodes
    pListFree       *Ivy_Obj_t;      // the list of free nodes 
    // timing statistics
    time1          abctime;
    time2          abctime;
}

type Ivy_FraigParams_t_ struct {
    nSimWords              int32;         // the number of words in the simulation info
    dSimSatur           float64;         // the ratio of refined classes when saturation is reached
    fPatScores              int32;        // enables simulation pattern scoring
    MaxScore              int32;          // max score after which resimulation is used
    dActConeRatio           float64;     // the ratio of cone to be bumped
    dActConeBumpMax           float64;   // the largest bump in activity
    fProve              int32;            // prove the miter outputs
    fVerbose              int32;          // verbose output
    fDoSparse              int32;         // skip sparse functions
    nBTLimitNode              int32;      // conflict limit at a node
    nBTLimitMiter              int32;     // conflict limit at an output
//    int              nBTLimitGlobal;    // conflict limit global
//    int              nInsLimitNode;     // inspection limit at a node
//    int              nInsLimitMiter;    // inspection limit at an output
//    int              nInsLimitGlobal;   // inspection limit global
}


const IVY_CUT_LIMIT =     256
const IVY_CUT_INPUT =       6

   
type Ivy_Cut_t struct {
    nLatches         int32;
    nSize       int16;
    nSizeMax       int16;
    pArray [IVY_CUT_INPUT]int32;
    uHash    uint32;
}

   
type Ivy_Store_t struct {
    nCuts         int32;
    nCutsM         int32;
    nCutsMax         int32;
    fSatur         int32;
    pCuts [IVY_CUT_LIMIT]Ivy_Cut_t; // storage for cuts
}

const IVY_LEAF_MASK =     255
const IVY_LEAF_BITS =       8

////////////////////////////////////////////////////////////////////////
///                      MACRO DEFINITIONS                           ///
////////////////////////////////////////////////////////////////////////

// Original macro:
// #define IVY_MIN(a,b)       (((a) < (b))? (a) : (b))

// Original macro:
// #define IVY_MAX(a,b)       (((a) > (b))? (a) : (b))


// extern void Ivy_ManAddMemory( Ivy_Man_t * p );


func Ivy_BitWordNum(nBits int32) int32 { return (nBits>>5) + ((nBits&31) > 0);           }
func Ivy_TruthWordNum(nVars int32) int32 { if nVars <= 5 {
    return 1;
} else {
    return (1 << (nVars - 5));
}     }
func Ivy_InfoHasBit(p *uint32, i int32) int32 { return (p[(i)>>5] & (1<<((i) & 31))) > 0;       }
func Ivy_InfoSetBit(p *uint32, i int32) { p[(i)>>5] |= (1<<((i) & 31));                   }
func Ivy_InfoXorBit(p *uint32, i int32) { p[(i)>>5] ^= (1<<((i) & 31));                   }

func Ivy_Regular(p *Ivy_Obj_t) *Ivy_Obj_t { return (Ivy_Obj_t *)((ABC_PTRUINT_T)(p) & ~01); }
func Ivy_Not(p *Ivy_Obj_t) *Ivy_Obj_t { return (Ivy_Obj_t *)((ABC_PTRUINT_T)(p) ^  01); }
func Ivy_NotCond(p *Ivy_Obj_t, c int32) *Ivy_Obj_t { return (Ivy_Obj_t *)((ABC_PTRUINT_T)(p) ^ (c)); }
func Ivy_IsComplement(p *Ivy_Obj_t) int32 { return int32(((ABC_PTRUINT_T)(p) & 01));          }

func Ivy_ManConst0(p *Ivy_Man_t) *Ivy_Obj_t { return Ivy_Not(p.pConst1);                     }
func Ivy_ManConst1(p *Ivy_Man_t) *Ivy_Obj_t { return p.pConst1;                              }
func Ivy_ManGhost(p *Ivy_Man_t) *Ivy_Obj_t { return &p.Ghost;                               }
func Ivy_ManPi(p *Ivy_Man_t, i int32) *Ivy_Obj_t { return (Ivy_Obj_t *)Vec_PtrEntry(p.vPis, i);   }
func Ivy_ManPo(p *Ivy_Man_t, i int32) *Ivy_Obj_t { return (Ivy_Obj_t *)Vec_PtrEntry(p.vPos, i);   }
func Ivy_ManObj(p *Ivy_Man_t, i int32) *Ivy_Obj_t { return (Ivy_Obj_t *)Vec_PtrEntry(p.vObjs, i);  }

func Ivy_EdgeCreate(Id int32, fCompl int32) Ivy_Edge_t { return (Id << 1) | fCompl;             }
func Ivy_EdgeId(Edge Ivy_Edge_t) int32 { return Edge >> 1;                      }
func Ivy_EdgeIsComplement(Edge Ivy_Edge_t) int32 { return Edge & 1;                       }
func Ivy_EdgeRegular(Edge Ivy_Edge_t) Ivy_Edge_t { return (Edge >> 1) << 1;               }
func Ivy_EdgeNot(Edge Ivy_Edge_t) Ivy_Edge_t { return Edge ^ 1;                       }
func Ivy_EdgeNotCond(Edge Ivy_Edge_t, fCond int32) Ivy_Edge_t { return Edge ^ fCond;                   }
func Ivy_EdgeFromNode(pNode *Ivy_Obj_t) Ivy_Edge_t { return Ivy_EdgeCreate( Ivy_Regular(pNode).Id, Ivy_IsComplement(pNode) );          }
func Ivy_EdgeToNode(p *Ivy_Man_t, Edge Ivy_Edge_t) *Ivy_Obj_t { return Ivy_NotCond( Ivy_ManObj(p, Ivy_EdgeId(Edge)), Ivy_EdgeIsComplement(Edge) ); }

func Ivy_LeafCreate(Id int32, Lat int32) int32 { return (Id << IVY_LEAF_BITS) | Lat;         }
func Ivy_LeafId(Leaf int32) int32 { return Leaf >> IVY_LEAF_BITS;               }
func Ivy_LeafLat(Leaf int32) int32 { return Leaf & IVY_LEAF_MASK;                }

func Ivy_ManPiNum(p *Ivy_Man_t) int32 { return p.nObjs[IVY_PI];                    }
func Ivy_ManPoNum(p *Ivy_Man_t) int32 { return p.nObjs[IVY_PO];                    }
func Ivy_ManAssertNum(p *Ivy_Man_t) int32 { return p.nObjs[IVY_ASSERT];                }
func Ivy_ManLatchNum(p *Ivy_Man_t) int32 { return p.nObjs[IVY_LATCH];                 }
func Ivy_ManAndNum(p *Ivy_Man_t) int32 { return p.nObjs[IVY_AND];                   }
func Ivy_ManExorNum(p *Ivy_Man_t) int32 { return p.nObjs[IVY_EXOR];                  }
func Ivy_ManBufNum(p *Ivy_Man_t) int32 { return p.nObjs[IVY_BUF];                   }
func Ivy_ManObjNum(p *Ivy_Man_t) int32 { return p.nCreated - p.nDeleted;           }
func Ivy_ManObjIdMax(p *Ivy_Man_t) int32 { return Vec_PtrSize(p.vObjs)-1;             }
func Ivy_ManNodeNum(p *Ivy_Man_t) int32 { return p.nObjs[IVY_AND]+p.nObjs[IVY_EXOR];}
func Ivy_ManHashObjNum(p *Ivy_Man_t) int32 { return p.nObjs[IVY_AND]+p.nObjs[IVY_EXOR]+p.nObjs[IVY_LATCH];     }
func Ivy_ManGetCost(p *Ivy_Man_t) int32 { return p.nObjs[IVY_AND]+3*p.nObjs[IVY_EXOR]+8*p.nObjs[IVY_LATCH]; }

func Ivy_ObjType(pObj *Ivy_Obj_t) Ivy_Type_t { return (Ivy_Type_t)pObj.Type;               }
func Ivy_ObjInit(pObj *Ivy_Obj_t) Ivy_Init_t { return (Ivy_Init_t)pObj.Init;               }
func Ivy_ObjIsConst1(pObj *Ivy_Obj_t) int32 { return pObj.Id == 0;            }
func Ivy_ObjIsGhost(pObj *Ivy_Obj_t) int32 { return pObj.Id < 0;             }
func Ivy_ObjIsNone(pObj *Ivy_Obj_t) int32 { return pObj.Type == IVY_NONE;   }
func Ivy_ObjIsPi(pObj *Ivy_Obj_t) int32 { return pObj.Type == IVY_PI;     }
func Ivy_ObjIsPo(pObj *Ivy_Obj_t) int32 { return pObj.Type == IVY_PO;     }
func Ivy_ObjIsCi(pObj *Ivy_Obj_t) int32 { return pObj.Type == IVY_PI || pObj.Type == IVY_LATCH; }
func Ivy_ObjIsCo(pObj *Ivy_Obj_t) int32 { return pObj.Type == IVY_PO || pObj.Type == IVY_LATCH; }
func Ivy_ObjIsAssert(pObj *Ivy_Obj_t) int32 { return pObj.Type == IVY_ASSERT; }
func Ivy_ObjIsLatch(pObj *Ivy_Obj_t) int32 { return pObj.Type == IVY_LATCH;  }
func Ivy_ObjIsAnd(pObj *Ivy_Obj_t) int32 { return pObj.Type == IVY_AND;    }
func Ivy_ObjIsExor(pObj *Ivy_Obj_t) int32 { return pObj.Type == IVY_EXOR;   }
func Ivy_ObjIsBuf(pObj *Ivy_Obj_t) int32 { return pObj.Type == IVY_BUF;    }
func Ivy_ObjIsNode(pObj *Ivy_Obj_t) int32 { return pObj.Type == IVY_AND || pObj.Type == IVY_EXOR; }
func Ivy_ObjIsTerm(pObj *Ivy_Obj_t) int32 { return pObj.Type == IVY_PI  || pObj.Type == IVY_PO || pObj.Type == IVY_ASSERT; }
func Ivy_ObjIsHash(pObj *Ivy_Obj_t) int32 { return pObj.Type == IVY_AND || pObj.Type == IVY_EXOR || pObj.Type == IVY_LATCH; }
func Ivy_ObjIsOneFanin(pObj *Ivy_Obj_t) int32 { return pObj.Type == IVY_PO  || pObj.Type == IVY_ASSERT || pObj.Type == IVY_BUF || pObj.Type == IVY_LATCH; }

func Ivy_ObjIsMarkA(pObj *Ivy_Obj_t) int32 { return pObj.fMarkA;  }
func Ivy_ObjSetMarkA(pObj *Ivy_Obj_t) { pObj.fMarkA = 1;     }
func Ivy_ObjClearMarkA(pObj *Ivy_Obj_t) { pObj.fMarkA = 0;     }
 
func Ivy_ObjSetTravId(pObj *Ivy_Obj_t, TravId int32) { pObj.TravId = TravId;                               }
func Ivy_ObjSetTravIdCurrent(p *Ivy_Man_t, pObj *Ivy_Obj_t) { pObj.TravId = p.nTravIds;                          }
func Ivy_ObjSetTravIdPrevious(p *Ivy_Man_t, pObj *Ivy_Obj_t) { pObj.TravId = p.nTravIds - 1;                      }
func Ivy_ObjIsTravIdCurrent(p *Ivy_Man_t, pObj *Ivy_Obj_t) int32 { return int32( (int32(pObj.TravId) == p.nTravIds));     }
func Ivy_ObjIsTravIdPrevious(p *Ivy_Man_t, pObj *Ivy_Obj_t) int32 { return int32( (int32(pObj.TravId) == p.nTravIds - 1)); }

func Ivy_ObjId(pObj *Ivy_Obj_t) int32 { return pObj.Id;                               }
func Ivy_ObjTravId(pObj *Ivy_Obj_t) int32 { return pObj.TravId;                           }
func Ivy_ObjPhase(pObj *Ivy_Obj_t) int32 { return pObj.fPhase;                           }
func Ivy_ObjExorFanout(pObj *Ivy_Obj_t) int32 { return pObj.fExFan;                           }
func Ivy_ObjRefs(pObj *Ivy_Obj_t) int32 { return pObj.nRefs;                            }
func Ivy_ObjRefsInc(pObj *Ivy_Obj_t) { pObj.nRefs++;                                 }
func Ivy_ObjRefsDec(pObj *Ivy_Obj_t) { assert(pObj.nRefs > 0, "pObj.nRefs > 0"); pObj.nRefs--;      }
func Ivy_ObjFaninId0(pObj *Ivy_Obj_t) int32 { if pObj.pFanin0 {
    return Ivy_ObjId(Ivy_Regular(pObj.pFanin0));
} else {
    return 0;
} }
func Ivy_ObjFaninId1(pObj *Ivy_Obj_t) int32 { if pObj.pFanin1 {
    return Ivy_ObjId(Ivy_Regular(pObj.pFanin1));
} else {
    return 0;
} }
func Ivy_ObjFaninC0(pObj *Ivy_Obj_t) int32 { return Ivy_IsComplement(pObj.pFanin0);        }
func Ivy_ObjFaninC1(pObj *Ivy_Obj_t) int32 { return Ivy_IsComplement(pObj.pFanin1);        }
func Ivy_ObjFanin0(pObj *Ivy_Obj_t) *Ivy_Obj_t { return Ivy_Regular(pObj.pFanin0);             }
func Ivy_ObjFanin1(pObj *Ivy_Obj_t) *Ivy_Obj_t { return Ivy_Regular(pObj.pFanin1);             }
func Ivy_ObjChild0(pObj *Ivy_Obj_t) *Ivy_Obj_t { return pObj.pFanin0;                          }
func Ivy_ObjChild1(pObj *Ivy_Obj_t) *Ivy_Obj_t { return pObj.pFanin1;                          }
func Ivy_ObjChild0Equiv(pObj *Ivy_Obj_t) *Ivy_Obj_t { assert(!(Ivy_IsComplement(pObj) != 0), "!Ivy_IsComplement(pObj)"); if Ivy_ObjFanin0(pObj) != nil {
    return Ivy_NotCond(Ivy_ObjFanin0(pObj).pEquiv, Ivy_ObjFaninC0(pObj));
} else {
    return nil;
}  }
func Ivy_ObjChild1Equiv(pObj *Ivy_Obj_t) *Ivy_Obj_t { assert(!(Ivy_IsComplement(pObj) != 0), "!Ivy_IsComplement(pObj)"); if Ivy_ObjFanin1(pObj) != nil {
    return Ivy_NotCond(Ivy_ObjFanin1(pObj).pEquiv, Ivy_ObjFaninC1(pObj));
} else {
    return nil;
}  }
func Ivy_ObjEquiv(pObj *Ivy_Obj_t) *Ivy_Obj_t { if Ivy_Regular(pObj).pEquiv {
    return Ivy_NotCond(Ivy_Regular(pObj).pEquiv, Ivy_IsComplement(pObj));
} else {
    return nil;
} }
func Ivy_ObjLevel(pObj *Ivy_Obj_t) int32 { return pObj.Level;                            }
// inlining of macro IVY_MAX
func Ivy_ObjLevelNew(pObj *Ivy_Obj_t) int32 { return 1 + Ivy_ObjIsExor(pObj) + (((Ivy_ObjFanin0(pObj).Level) > (Ivy_ObjFanin1(pObj).Level))? (Ivy_ObjFanin0(pObj).Level) : (Ivy_ObjFanin1(pObj).Level));       }
func Ivy_ObjFaninPhase(pObj *Ivy_Obj_t) int32 { if Ivy_IsComplement(pObj) != 0 {
    return !Ivy_Regular(pObj).fPhase;
} else {
    return pObj.fPhase;
} }

func Ivy_ObjClean(pObj *Ivy_Obj_t) { 
    var IdSaved int32  = pObj.Id; 
    *pObj = Ivy_Obj_t{}; 
    pObj.Id = IdSaved; 
}
func Ivy_ObjOverwrite(pBase *Ivy_Obj_t, pData *Ivy_Obj_t) { 
    var IdSaved int32  = pBase.Id; 
    memcpy( pBase, pData, sizeof(Ivy_Obj_t) ); 
    pBase.Id = IdSaved;         
}
func Ivy_ObjWhatFanin(pObj *Ivy_Obj_t, pFanin *Ivy_Obj_t) int32 { 
    if Ivy_ObjFanin0(pObj) == pFanin { return 0; } 
    if Ivy_ObjFanin1(pObj) == pFanin { return 1; } 
    assert(false, "0"); return -1; 
}
func Ivy_ObjFanoutC(pObj *Ivy_Obj_t, pFanout *Ivy_Obj_t) int32 { 
    if Ivy_ObjFanin0(pFanout) == pObj { return Ivy_ObjFaninC0(pObj); } 
    if Ivy_ObjFanin1(pFanout) == pObj { return Ivy_ObjFaninC1(pObj); } 
    assert(false, "0"); return -1; 
}

// create the ghost of the new node
func Ivy_ObjCreateGhost(p *Ivy_Man_t, p0 *Ivy_Obj_t, p1 *Ivy_Obj_t, Type Ivy_Type_t, Init Ivy_Init_t) *Ivy_Obj_t {
    var pGhost, pTemp *Ivy_Obj_t;
    assert(Type != IVY_AND || !(Ivy_ObjIsConst1(Ivy_Regular(p0)) != 0), "Type != IVY_AND || !Ivy_ObjIsConst1(Ivy_Regular(p0))");
    assert(p1 == nil || !(Ivy_ObjIsConst1(Ivy_Regular(p1)) != 0), "p1 == nil || !Ivy_ObjIsConst1(Ivy_Regular(p1))");
    assert(Type == IVY_PI || Ivy_Regular(p0) != Ivy_Regular(p1), "Type == IVY_PI || Ivy_Regular(p0) != Ivy_Regular(p1)");
    assert(Type != IVY_LATCH || !(Ivy_IsComplement(p0) != 0), "Type != IVY_LATCH || !Ivy_IsComplement(p0)");
//    assert( p1 == NULL || (!Ivy_ObjIsLatch(Ivy_Regular(p0)) || !Ivy_ObjIsLatch(Ivy_Regular(p1))) );
    pGhost = Ivy_ManGhost(p);
    pGhost.Type = Type;
    pGhost.Init = Init;
    pGhost.pFanin0 = p0;
    pGhost.pFanin1 = p1;
    if (p1 != nil) && Ivy_ObjFaninId0(pGhost) > Ivy_ObjFaninId1(pGhost) {
        pTemp = pGhost.pFanin0;
pGhost.pFanin0 = pGhost.pFanin1;
pGhost.pFanin1 = pTemp; }
    return pGhost;
}

// get the complemented initial state
func Ivy_InitNotCond(Init Ivy_Init_t, fCompl int32) Ivy_Init_t {
    assert(Init != IVY_INIT_NONE, "Init != IVY_INIT_NONE");
    if fCompl == 0 {
        return Init; }
    if Init == IVY_INIT_0 {
        return IVY_INIT_1; }
    if Init == IVY_INIT_1 {
        return IVY_INIT_0; }
    return IVY_INIT_DC;
}

// get the initial state after forward retiming over AND gate
func Ivy_InitAnd(InitA Ivy_Init_t, InitB Ivy_Init_t) Ivy_Init_t {
    assert(InitA != IVY_INIT_NONE && InitB != IVY_INIT_NONE, "InitA != IVY_INIT_NONE && InitB != IVY_INIT_NONE");
    if InitA == IVY_INIT_0 || InitB == IVY_INIT_0 {
        return IVY_INIT_0; }
    if InitA == IVY_INIT_DC || InitB == IVY_INIT_DC {
        return IVY_INIT_DC; }
    return IVY_INIT_1;
}

// get the initial state after forward retiming over EXOR gate
func Ivy_InitExor(InitA Ivy_Init_t, InitB Ivy_Init_t) Ivy_Init_t {
    assert(InitA != IVY_INIT_NONE && InitB != IVY_INIT_NONE, "InitA != IVY_INIT_NONE && InitB != IVY_INIT_NONE");
    if InitA == IVY_INIT_DC || InitB == IVY_INIT_DC {
        return IVY_INIT_DC; }
    if InitA == IVY_INIT_0 && InitB == IVY_INIT_1 {
        return IVY_INIT_1; }
    if InitA == IVY_INIT_1 && InitB == IVY_INIT_0 {
        return IVY_INIT_1; }
    return IVY_INIT_0;
}

// internal memory manager
func Ivy_ManFetchMemory(p *Ivy_Man_t) *Ivy_Obj_t { 
    var pTemp *Ivy_Obj_t;
    if p.pListFree == nil {
        Ivy_ManAddMemory( p ); }
    pTemp = p.pListFree;
    p.pListFree = *((Ivy_Obj_t **)pTemp);
    *pTemp = Ivy_Obj_t{}; 
    return pTemp;
}
func Ivy_ManRecycleMemory(p *Ivy_Man_t, pEntry *Ivy_Obj_t) {
    pEntry.Type = IVY_NONE; // distinquishes dead node from live node
    *((Ivy_Obj_t **)pEntry) = p.pListFree;
    p.pListFree = pEntry;
}


////////////////////////////////////////////////////////////////////////
///                             ITERATORS                            ///
////////////////////////////////////////////////////////////////////////

// iterator over the primary inputs
// Original macro:
// #define Ivy_ManForEachPi( p, pObj, i )                                          \
//     Vec_PtrForEachEntry( Ivy_Obj_t *, p->vPis, pObj, i )

// iterator over the primary outputs
// Original macro:
// #define Ivy_ManForEachPo( p, pObj, i )                                          \
//     Vec_PtrForEachEntry( Ivy_Obj_t *, p->vPos, pObj, i )

// iterator over all objects, including those currently not used
// Original macro:
// #define Ivy_ManForEachObj( p, pObj, i )                                         \
//     Vec_PtrForEachEntry( Ivy_Obj_t *, p->vObjs, pObj, i ) if ( (pObj) == NULL ) {} else

// iterator over the combinational inputs
// Original macro:
// #define Ivy_ManForEachCi( p, pObj, i )                                          \
//     Ivy_ManForEachObj( p, pObj, i ) if ( !Ivy_ObjIsCi(pObj) ) {} else

// iterator over the combinational outputs
// Original macro:
// #define Ivy_ManForEachCo( p, pObj, i )                                          \
//     Ivy_ManForEachObj( p, pObj, i ) if ( !Ivy_ObjIsCo(pObj) ) {} else

// iterator over logic nodes (AND and EXOR gates)
// Original macro:
// #define Ivy_ManForEachNode( p, pObj, i )                                        \
//     Ivy_ManForEachObj( p, pObj, i ) if ( !Ivy_ObjIsNode(pObj) ) {} else

// iterator over logic latches
// Original macro:
// #define Ivy_ManForEachLatch( p, pObj, i )                                       \
//     Ivy_ManForEachObj( p, pObj, i ) if ( !Ivy_ObjIsLatch(pObj) ) {} else

// iterator over the nodes whose IDs are stored in the array
// Original macro:
// #define Ivy_ManForEachNodeVec( p, vIds, pObj, i )                               \
//     for ( i = 0; i < Vec_IntSize(vIds) && ((pObj) = Ivy_ManObj(p, Vec_IntEntry(vIds,i))); i++ )

// iterator over the fanouts of an object
// Original macro:
// #define Ivy_ObjForEachFanout( p, pObj, vArray, pFanout, i )                     \
//     for ( i = 0, Ivy_ObjCollectFanouts(p, pObj, vArray);                        \
//           i < Vec_PtrSize(vArray) && ((pFanout) = (Ivy_Obj_t *)Vec_PtrEntry(vArray,i)); i++ )


////////////////////////////////////////////////////////////////////////
///                    FUNCTION DECLARATIONS                         ///
////////////////////////////////////////////////////////////////////////

/*=== ivyBalance.c ========================================================*/
// extern Ivy_Man_t *     Ivy_ManBalance( Ivy_Man_t * p, int fUpdateLevel );

// extern Ivy_Obj_t *     Ivy_NodeBalanceBuildSuper( Ivy_Man_t * p, Vec_Ptr_t * vSuper, Ivy_Type_t Type, int fUpdateLevel );

/*=== ivyCanon.c ========================================================*/
// extern Ivy_Obj_t *     Ivy_CanonAnd( Ivy_Man_t * p, Ivy_Obj_t * p0, Ivy_Obj_t * p1 );

// extern Ivy_Obj_t *     Ivy_CanonExor( Ivy_Man_t * p, Ivy_Obj_t * p0, Ivy_Obj_t * p1 );

// extern Ivy_Obj_t *     Ivy_CanonLatch( Ivy_Man_t * p, Ivy_Obj_t * pObj, Ivy_Init_t Init );

/*=== ivyCheck.c ========================================================*/
// extern int             Ivy_ManCheck( Ivy_Man_t * p );

// extern int             Ivy_ManCheckFanoutNums( Ivy_Man_t * p );

// extern int             Ivy_ManCheckFanouts( Ivy_Man_t * p );

// extern int             Ivy_ManCheckChoices( Ivy_Man_t * p );

/*=== ivyCut.c ==========================================================*/
// extern void            Ivy_ManSeqFindCut( Ivy_Man_t * p, Ivy_Obj_t * pNode, Vec_Int_t * vFront, Vec_Int_t * vInside, int nSize );

// extern Ivy_Store_t *   Ivy_NodeFindCutsAll( Ivy_Man_t * p, Ivy_Obj_t * pObj, int nLeaves );

/*=== ivyDfs.c ==========================================================*/
// extern Vec_Int_t *     Ivy_ManDfs( Ivy_Man_t * p );

// extern Vec_Int_t *     Ivy_ManDfsSeq( Ivy_Man_t * p, Vec_Int_t ** pvLatches );

// extern void            Ivy_ManCollectCone( Ivy_Obj_t * pObj, Vec_Ptr_t * vFront, Vec_Ptr_t * vCone );

// extern Vec_Vec_t *     Ivy_ManLevelize( Ivy_Man_t * p );

// extern Vec_Int_t *     Ivy_ManRequiredLevels( Ivy_Man_t * p );

// extern int             Ivy_ManIsAcyclic( Ivy_Man_t * p );

// extern int             Ivy_ManSetLevels( Ivy_Man_t * p, int fHaig );

/*=== ivyDsd.c ==========================================================*/
// extern int             Ivy_TruthDsd( unsigned uTruth, Vec_Int_t * vTree );

// extern void            Ivy_TruthDsdPrint( FILE * pFile, Vec_Int_t * vTree );

// extern unsigned        Ivy_TruthDsdCompute( Vec_Int_t * vTree );

// extern void            Ivy_TruthDsdComputePrint( unsigned uTruth );

// extern Ivy_Obj_t *     Ivy_ManDsdConstruct( Ivy_Man_t * p, Vec_Int_t * vFront, Vec_Int_t * vTree );

/*=== ivyFanout.c ==========================================================*/
// extern void            Ivy_ManStartFanout( Ivy_Man_t * p );

// extern void            Ivy_ManStopFanout( Ivy_Man_t * p );

// extern void            Ivy_ObjAddFanout( Ivy_Man_t * p, Ivy_Obj_t * pObj, Ivy_Obj_t * pFanout );

// extern void            Ivy_ObjDeleteFanout( Ivy_Man_t * p, Ivy_Obj_t * pObj, Ivy_Obj_t * pFanout );

// extern void            Ivy_ObjPatchFanout( Ivy_Man_t * p, Ivy_Obj_t * pObj, Ivy_Obj_t * pFanoutOld, Ivy_Obj_t * pFanoutNew );

// extern void            Ivy_ObjCollectFanouts( Ivy_Man_t * p, Ivy_Obj_t * pObj, Vec_Ptr_t * vArray );

// extern Ivy_Obj_t *     Ivy_ObjReadFirstFanout( Ivy_Man_t * p, Ivy_Obj_t * pObj );

// extern int             Ivy_ObjFanoutNum( Ivy_Man_t * p, Ivy_Obj_t * pObj );

/*=== ivyFastMap.c =============================================================*/
// extern void            Ivy_FastMapPerform( Ivy_Man_t * pAig, int nLimit, int fRecovery, int fVerbose );

// extern void            Ivy_FastMapStop( Ivy_Man_t * pAig );

// extern void            Ivy_FastMapReadSupp( Ivy_Man_t * pAig, Ivy_Obj_t * pObj, Vec_Int_t * vLeaves );

// extern void            Ivy_FastMapReverseLevel( Ivy_Man_t * pAig );

/*=== ivyFraig.c ==========================================================*/
// extern int             Ivy_FraigProve( Ivy_Man_t ** ppManAig, void * pPars );

// extern Ivy_Man_t *     Ivy_FraigPerform( Ivy_Man_t * pManAig, Ivy_FraigParams_t * pParams );

// extern Ivy_Man_t *     Ivy_FraigMiter( Ivy_Man_t * pManAig, Ivy_FraigParams_t * pParams );

// extern void            Ivy_FraigParamsDefault( Ivy_FraigParams_t * pParams );

/*=== ivyHaig.c ==========================================================*/
// extern void            Ivy_ManHaigStart( Ivy_Man_t * p, int fVerbose );

// extern void            Ivy_ManHaigTrasfer( Ivy_Man_t * p, Ivy_Man_t * pNew );

// extern void            Ivy_ManHaigStop( Ivy_Man_t * p );

// extern void            Ivy_ManHaigPostprocess( Ivy_Man_t * p, int fVerbose );

// extern void            Ivy_ManHaigCreateObj( Ivy_Man_t * p, Ivy_Obj_t * pObj );

// extern void            Ivy_ManHaigCreateChoice( Ivy_Man_t * p, Ivy_Obj_t * pObjOld, Ivy_Obj_t * pObjNew );

// extern void            Ivy_ManHaigSimulate( Ivy_Man_t * p );

/*=== ivyMan.c ==========================================================*/
// extern Ivy_Man_t *     Ivy_ManStart();

// extern Ivy_Man_t *     Ivy_ManStartFrom( Ivy_Man_t * p );

// extern Ivy_Man_t *     Ivy_ManDup( Ivy_Man_t * p );

// extern Ivy_Man_t *     Ivy_ManFrames( Ivy_Man_t * pMan, int nLatches, int nFrames, int fInit, Vec_Ptr_t ** pvMapping );

// extern void            Ivy_ManStop( Ivy_Man_t * p );

// extern int             Ivy_ManCleanup( Ivy_Man_t * p );

// extern int             Ivy_ManPropagateBuffers( Ivy_Man_t * p, int fUpdateLevel );

// extern void            Ivy_ManPrintStats( Ivy_Man_t * p );

// extern void            Ivy_ManMakeSeq( Ivy_Man_t * p, int nLatches, int * pInits );

/*=== ivyMem.c ==========================================================*/
// extern void            Ivy_ManStartMemory( Ivy_Man_t * p );

// extern void            Ivy_ManStopMemory( Ivy_Man_t * p );

/*=== ivyMulti.c ==========================================================*/
// extern Ivy_Obj_t *     Ivy_Multi( Ivy_Man_t * p, Ivy_Obj_t ** pArgs, int nArgs, Ivy_Type_t Type );

// extern Ivy_Obj_t *     Ivy_Multi1( Ivy_Man_t * p, Ivy_Obj_t ** pArgs, int nArgs, Ivy_Type_t Type );

// extern Ivy_Obj_t *     Ivy_Multi_rec( Ivy_Man_t * p, Ivy_Obj_t ** ppObjs, int nObjs, Ivy_Type_t Type );

// extern Ivy_Obj_t *     Ivy_MultiBalance_rec( Ivy_Man_t * p, Ivy_Obj_t ** pArgs, int nArgs, Ivy_Type_t Type );

// extern int             Ivy_MultiPlus( Ivy_Man_t * p, Vec_Ptr_t * vLeaves, Vec_Ptr_t * vCone, Ivy_Type_t Type, int nLimit, Vec_Ptr_t * vSol );

/*=== ivyObj.c ==========================================================*/
// extern Ivy_Obj_t *     Ivy_ObjCreatePi( Ivy_Man_t * p );

// extern Ivy_Obj_t *     Ivy_ObjCreatePo( Ivy_Man_t * p, Ivy_Obj_t * pDriver );

// extern Ivy_Obj_t *     Ivy_ObjCreate( Ivy_Man_t * p, Ivy_Obj_t * pGhost );

// extern void            Ivy_ObjConnect( Ivy_Man_t * p, Ivy_Obj_t * pObj, Ivy_Obj_t * pFan0, Ivy_Obj_t * pFan1 );

// extern void            Ivy_ObjDisconnect( Ivy_Man_t * p, Ivy_Obj_t * pObj );

// extern void            Ivy_ObjPatchFanin0( Ivy_Man_t * p, Ivy_Obj_t * pObj, Ivy_Obj_t * pFaninNew );

// extern void            Ivy_ObjDelete( Ivy_Man_t * p, Ivy_Obj_t * pObj, int fFreeTop );

// extern void            Ivy_ObjDelete_rec( Ivy_Man_t * p, Ivy_Obj_t * pObj, int fFreeTop );

// extern void            Ivy_ObjReplace( Ivy_Man_t * p, Ivy_Obj_t * pObjOld, Ivy_Obj_t * pObjNew, int fDeleteOld, int fFreeTop, int fUpdateLevel );

// extern void            Ivy_NodeFixBufferFanins( Ivy_Man_t * p, Ivy_Obj_t * pNode, int fUpdateLevel );

/*=== ivyOper.c =========================================================*/
// extern Ivy_Obj_t *     Ivy_Oper( Ivy_Man_t * p, Ivy_Obj_t * p0, Ivy_Obj_t * p1, Ivy_Type_t Type );

// extern Ivy_Obj_t *     Ivy_And( Ivy_Man_t * p, Ivy_Obj_t * p0, Ivy_Obj_t * p1 );

// extern Ivy_Obj_t *     Ivy_Or( Ivy_Man_t * p, Ivy_Obj_t * p0, Ivy_Obj_t * p1 );

// extern Ivy_Obj_t *     Ivy_Exor( Ivy_Man_t * p, Ivy_Obj_t * p0, Ivy_Obj_t * p1 );

// extern Ivy_Obj_t *     Ivy_Mux( Ivy_Man_t * p, Ivy_Obj_t * pC, Ivy_Obj_t * p1, Ivy_Obj_t * p0 );

// extern Ivy_Obj_t *     Ivy_Maj( Ivy_Man_t * p, Ivy_Obj_t * pA, Ivy_Obj_t * pB, Ivy_Obj_t * pC );

// extern Ivy_Obj_t *     Ivy_Miter( Ivy_Man_t * p, Vec_Ptr_t * vPairs );

// extern Ivy_Obj_t *     Ivy_Latch( Ivy_Man_t * p, Ivy_Obj_t * pObj, Ivy_Init_t Init );

/*=== ivyResyn.c =========================================================*/
// extern Ivy_Man_t *     Ivy_ManResyn0( Ivy_Man_t * p, int fUpdateLevel, int fVerbose );

// extern Ivy_Man_t *     Ivy_ManResyn( Ivy_Man_t * p, int fUpdateLevel, int fVerbose );

// extern Ivy_Man_t *     Ivy_ManRwsat( Ivy_Man_t * pMan, int fVerbose );

/*=== ivyRewrite.c =========================================================*/
// extern int             Ivy_ManSeqRewrite( Ivy_Man_t * p, int fUpdateLevel, int fUseZeroCost );

// extern int             Ivy_ManRewriteAlg( Ivy_Man_t * p, int fUpdateLevel, int fUseZeroCost );

// extern int             Ivy_ManRewritePre( Ivy_Man_t * p, int fUpdateLevel, int fUseZeroCost, int fVerbose );

/*=== ivySeq.c =========================================================*/
// extern int             Ivy_ManRewriteSeq( Ivy_Man_t * p, int fUseZeroCost, int fVerbose );

/*=== ivyShow.c =========================================================*/
// extern void            Ivy_ManShow( Ivy_Man_t * pMan, int fHaig, Vec_Ptr_t * vBold );

/*=== ivyTable.c ========================================================*/
// extern Ivy_Obj_t *     Ivy_TableLookup( Ivy_Man_t * p, Ivy_Obj_t * pObj );

// extern void            Ivy_TableInsert( Ivy_Man_t * p, Ivy_Obj_t * pObj );

// extern void            Ivy_TableDelete( Ivy_Man_t * p, Ivy_Obj_t * pObj );

// extern void            Ivy_TableUpdate( Ivy_Man_t * p, Ivy_Obj_t * pObj, int ObjIdNew );

// extern int             Ivy_TableCountEntries( Ivy_Man_t * p );

// extern void            Ivy_TableProfile( Ivy_Man_t * p );

/*=== ivyUtil.c =========================================================*/
// extern void            Ivy_ManIncrementTravId( Ivy_Man_t * p );

// extern void            Ivy_ManCleanTravId( Ivy_Man_t * p );

// extern unsigned *      Ivy_ManCutTruth( Ivy_Man_t * p, Ivy_Obj_t * pRoot, Vec_Int_t * vLeaves, Vec_Int_t * vNodes, Vec_Int_t * vTruth );

// extern void            Ivy_ManCollectCut( Ivy_Man_t * p, Ivy_Obj_t * pRoot, Vec_Int_t * vLeaves, Vec_Int_t * vNodes );

// extern Vec_Int_t *     Ivy_ManLatches( Ivy_Man_t * p );

// extern int             Ivy_ManLevels( Ivy_Man_t * p );

// extern void            Ivy_ManResetLevels( Ivy_Man_t * p );

// extern int             Ivy_ObjMffcLabel( Ivy_Man_t * p, Ivy_Obj_t * pObj );

// extern void            Ivy_ObjUpdateLevel_rec( Ivy_Man_t * p, Ivy_Obj_t * pObj );

// extern void            Ivy_ObjUpdateLevelR_rec( Ivy_Man_t * p, Ivy_Obj_t * pObj, int ReqNew );

// extern int             Ivy_ObjIsMuxType( Ivy_Obj_t * pObj );

// extern Ivy_Obj_t *     Ivy_ObjRecognizeMux( Ivy_Obj_t * pObj, Ivy_Obj_t ** ppObjT, Ivy_Obj_t ** ppObjE );

// extern Ivy_Obj_t *     Ivy_ObjReal( Ivy_Obj_t * pObj );

// extern void            Ivy_ObjPrintVerbose( Ivy_Man_t * p, Ivy_Obj_t * pObj, int fHaig );

// extern void            Ivy_ManPrintVerbose( Ivy_Man_t * p, int fHaig );

// extern int             Ivy_CutTruthPrint( Ivy_Man_t * p, Ivy_Cut_t * pCut, unsigned uTruth );




//ABC_NAMESPACE_HEADER_END





////////////////////////////////////////////////////////////////////////
///                       END OF FILE                                ///
////////////////////////////////////////////////////////////////////////

