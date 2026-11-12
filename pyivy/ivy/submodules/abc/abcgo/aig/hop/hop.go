package hop

/**CFile****************************************************************

  FileName    [hop.h]

  SystemName  [ABC: Logic synthesis and verification system.]

  PackageName [Minimalistic And-Inverter Graph package.]

  Synopsis    [External declarations.]

  Author      [Alan Mishchenko]
  
  Affiliation [UC Berkeley]

  Date        [Ver. 1.0. Started - May 11, 2006.]

  Revision    [$Id: hop.h,v 1.00 2006/05/11 00:00:00 alanmi Exp $]

***********************************************************************/





////////////////////////////////////////////////////////////////////////
///                          INCLUDES                                ///
////////////////////////////////////////////////////////////////////////

//#include <stdio.h>
//#include <stdlib.h>
//#include <string.h>
//#include <assert.h>

//#include "misc/vec/vec.h"

////////////////////////////////////////////////////////////////////////
///                         PARAMETERS                               ///
////////////////////////////////////////////////////////////////////////



//ABC_NAMESPACE_HEADER_START
 

////////////////////////////////////////////////////////////////////////
///                         BASIC TYPES                              ///
////////////////////////////////////////////////////////////////////////

type Hop_Man_t =            Hop_Man_t_;
type Hop_Obj_t =            Hop_Obj_t_;
typedef int                          Hop_Edge_t;

// object types
 type Hop_Type_t int32
const ( 
    AIG_NONE Hop_Type_t = iota
                        // 0: non-existent object
    AIG_CONST1
                      // 1: constant 1 
    AIG_PI
                          // 2: primary input
    AIG_PO
                          // 3: primary output
    AIG_AND
                         // 4: AND node
    AIG_EXOR
                        // 5: EXOR node
    AIG_VOID                         // 6: unused object

)

 

// the AIG node
type Hop_Obj_t_ struct {  // 6 words

    union { 
        void *       pData;          // misc
        int          iData; };       // misc
    union { 
        Hop_Obj_t *  pNext;          // strashing table
        int          PioNum; };      // the number of PI/PO
    pFanin0       *Hop_Obj_t;        // fanin
    pFanin1       *Hop_Obj_t;        // fanin
    Type uint32; /* C bitfield: 3 bits; review packing and truncation */   // object type
    fPhase uint32; /* C bitfield: 1 bits; review packing and truncation */   // value under 000...0 pattern
    fMarkA uint32; /* C bitfield: 1 bits; review packing and truncation */   // multipurpose mask
    fMarkB uint32; /* C bitfield: 1 bits; review packing and truncation */   // multipurpose mask
    nRefs uint32; /* C bitfield: 26 bits; review packing and truncation */   // reference count (level)
    Id              int32;             // unique ID of the node
}

// the AIG manager
type Hop_Man_t_ struct {
    // AIG nodes
    vPis       *Vec_Ptr_t;           // the array of PIs
    vPos       *Vec_Ptr_t;           // the array of POs
    vObjs       *Vec_Ptr_t;          // the array of all nodes (optional)
    pConst1       *Hop_Obj_t;        // the constant 1 node
    Ghost        Hop_Obj_t;          // the ghost node
    // AIG node counters
    nObjs [AIG_VOID]int32;// the number of objects by type
    nCreated              int32;       // the number of created objects
    nDeleted              int32;       // the number of deleted objects
    // stuctural hash table
    pTable      **Hop_Obj_t;         // structural hash table
    nTableSize              int32;     // structural hash table size
    // various data members
    pData            any;          // the temporary data
    nTravIds              int32;       // the current traversal ID
    fRefCount              int32;      // enables reference counting
    fCatchExor              int32;     // enables EXOR nodes
    // memory management
    vChunks       *Vec_Ptr_t;        // allocated memory pieces
    vPages       *Vec_Ptr_t;         // memory pages used by nodes
    pListFree       *Hop_Obj_t;      // the list of free nodes 
    // timing statistics
    time1          abctime;
    time2          abctime;
}

////////////////////////////////////////////////////////////////////////
///                      MACRO DEFINITIONS                           ///
////////////////////////////////////////////////////////////////////////
// extern void Hop_ManAddMemory( Hop_Man_t * p );


func Hop_BitWordNum(nBits int32) int32 { return (nBits>>5) + ((nBits&31) > 0);           }
func Hop_TruthWordNum(nVars int32) int32 { if nVars <= 5 {
    return 1;
} else {
    return (1 << (nVars - 5));
}     }
func Hop_InfoHasBit(p *uint32, i int32) int32 { return (p[(i)>>5] & (1<<((i) & 31))) > 0;       }
func Hop_InfoSetBit(p *uint32, i int32) { p[(i)>>5] |= (1<<((i) & 31));                   }
func Hop_InfoXorBit(p *uint32, i int32) { p[(i)>>5] ^= (1<<((i) & 31));                   }
func Hop_Base2Log(n uint32) int32 { var r int32; if n < 2 { return n; }
 for  r, n = 0, n-1; (n != 0); n >>= 1, r++  {}; return r; }
func Hop_Base10Log(n uint32) int32 { var r int32; if n < 2 { return n; }
 for  r, n = 0, n-1; (n != 0); n /= 10, r++  {}; return r; }

func Hop_Regular(p *Hop_Obj_t) *Hop_Obj_t { return (Hop_Obj_t *)((ABC_PTRUINT_T)(p) & ~01); }
func Hop_Not(p *Hop_Obj_t) *Hop_Obj_t { return (Hop_Obj_t *)((ABC_PTRUINT_T)(p) ^  01); }
func Hop_NotCond(p *Hop_Obj_t, c int32) *Hop_Obj_t { return (Hop_Obj_t *)((ABC_PTRUINT_T)(p) ^ (c)); }
func Hop_IsComplement(p *Hop_Obj_t) int32 { return int32(((ABC_PTRUINT_T)(p) & 01));          }

func Hop_ManConst0(p *Hop_Man_t) *Hop_Obj_t { return Hop_Not(p.pConst1);                     }
func Hop_ManConst1(p *Hop_Man_t) *Hop_Obj_t { return p.pConst1;                              }
func Hop_ManGhost(p *Hop_Man_t) *Hop_Obj_t { return &p.Ghost;                               }
func Hop_ManPi(p *Hop_Man_t, i int32) *Hop_Obj_t { return (Hop_Obj_t *)Vec_PtrEntry(p.vPis, i);   }
func Hop_ManPo(p *Hop_Man_t, i int32) *Hop_Obj_t { return (Hop_Obj_t *)Vec_PtrEntry(p.vPos, i);   }
func Hop_ManObj(p *Hop_Man_t, i int32) *Hop_Obj_t { if p.vObjs {
    return (Hop_Obj_t *)Vec_PtrEntry(p.vObjs, i);
} else {
    return nil;
}  }

func Hop_EdgeCreate(Id int32, fCompl int32) Hop_Edge_t { return (Id << 1) | fCompl;             }
func Hop_EdgeId(Edge Hop_Edge_t) int32 { return Edge >> 1;                      }
func Hop_EdgeIsComplement(Edge Hop_Edge_t) int32 { return Edge & 1;                       }
func Hop_EdgeRegular(Edge Hop_Edge_t) Hop_Edge_t { return (Edge >> 1) << 1;               }
func Hop_EdgeNot(Edge Hop_Edge_t) Hop_Edge_t { return Edge ^ 1;                       }
func Hop_EdgeNotCond(Edge Hop_Edge_t, fCond int32) Hop_Edge_t { return Edge ^ fCond;                   }

func Hop_ManPiNum(p *Hop_Man_t) int32 { return p.nObjs[AIG_PI];                    }
func Hop_ManPoNum(p *Hop_Man_t) int32 { return p.nObjs[AIG_PO];                    }
func Hop_ManAndNum(p *Hop_Man_t) int32 { return p.nObjs[AIG_AND];                   }
func Hop_ManExorNum(p *Hop_Man_t) int32 { return p.nObjs[AIG_EXOR];                  }
func Hop_ManNodeNum(p *Hop_Man_t) int32 { return p.nObjs[AIG_AND]+p.nObjs[AIG_EXOR];}
func Hop_ManGetCost(p *Hop_Man_t) int32 { return p.nObjs[AIG_AND]+3*p.nObjs[AIG_EXOR]; }
func Hop_ManObjNum(p *Hop_Man_t) int32 { return p.nCreated - p.nDeleted;           }

func Hop_ObjType(pObj *Hop_Obj_t) Hop_Type_t { return (Hop_Type_t)pObj.Type;               }
func Hop_ObjIsNone(pObj *Hop_Obj_t) int32 { return pObj.Type == AIG_NONE;   }
func Hop_ObjIsConst1(pObj *Hop_Obj_t) int32 { assert(!(Hop_IsComplement(pObj) != 0), "!Hop_IsComplement(pObj)"); return pObj.Type == AIG_CONST1; }
func Hop_ObjIsPi(pObj *Hop_Obj_t) int32 { return pObj.Type == AIG_PI;     }
func Hop_ObjIsPo(pObj *Hop_Obj_t) int32 { return pObj.Type == AIG_PO;     }
func Hop_ObjIsAnd(pObj *Hop_Obj_t) int32 { return pObj.Type == AIG_AND;    }
func Hop_ObjIsExor(pObj *Hop_Obj_t) int32 { return pObj.Type == AIG_EXOR;   }
func Hop_ObjIsNode(pObj *Hop_Obj_t) int32 { return pObj.Type == AIG_AND || pObj.Type == AIG_EXOR;   }
func Hop_ObjIsTerm(pObj *Hop_Obj_t) int32 { return pObj.Type == AIG_PI  || pObj.Type == AIG_PO || pObj.Type == AIG_CONST1; }
func Hop_ObjIsHash(pObj *Hop_Obj_t) int32 { return pObj.Type == AIG_AND || pObj.Type == AIG_EXOR;   }

func Hop_ObjIsMarkA(pObj *Hop_Obj_t) int32 { return pObj.fMarkA;  }
func Hop_ObjSetMarkA(pObj *Hop_Obj_t) { pObj.fMarkA = 1;     }
func Hop_ObjClearMarkA(pObj *Hop_Obj_t) { pObj.fMarkA = 0;     }
 
func Hop_ObjSetTravId(pObj *Hop_Obj_t, TravId int32) { pObj.pData = (void *)(ABC_PTRINT_T)TravId;                      }
func Hop_ObjSetTravIdCurrent(p *Hop_Man_t, pObj *Hop_Obj_t) { pObj.pData = (void *)(ABC_PTRINT_T)p.nTravIds;                 }
func Hop_ObjSetTravIdPrevious(p *Hop_Man_t, pObj *Hop_Obj_t) { pObj.pData = (void *)(ABC_PTRINT_T)(p.nTravIds - 1);           }
func Hop_ObjIsTravIdCurrent(p *Hop_Man_t, pObj *Hop_Obj_t) int32 { return int32((int32((ABC_PTRINT_T))pObj.pData == p.nTravIds));     }
func Hop_ObjIsTravIdPrevious(p *Hop_Man_t, pObj *Hop_Obj_t) int32 { return int32((int32((ABC_PTRINT_T))pObj.pData == p.nTravIds - 1)); }

func Hop_ObjTravId(pObj *Hop_Obj_t) int32 { return int32((ABC_PTRINT_T))pObj.pData;        }
func Hop_ObjPhase(pObj *Hop_Obj_t) int32 { return pObj.fPhase;                           }
func Hop_ObjRefs(pObj *Hop_Obj_t) int32 { return pObj.nRefs;                            }
func Hop_ObjRef(pObj *Hop_Obj_t) { pObj.nRefs++;                                 }
func Hop_ObjDeref(pObj *Hop_Obj_t) { assert(pObj.nRefs > 0, "pObj.nRefs > 0"); pObj.nRefs--;      }
func Hop_ObjClearRef(pObj *Hop_Obj_t) { pObj.nRefs = 0;                               }
func Hop_ObjFaninC0(pObj *Hop_Obj_t) int32 { return Hop_IsComplement(pObj.pFanin0);        }
func Hop_ObjFaninC1(pObj *Hop_Obj_t) int32 { return Hop_IsComplement(pObj.pFanin1);        }
func Hop_ObjFanin0(pObj *Hop_Obj_t) *Hop_Obj_t { return Hop_Regular(pObj.pFanin0);             }
func Hop_ObjFanin1(pObj *Hop_Obj_t) *Hop_Obj_t { return Hop_Regular(pObj.pFanin1);             }
func Hop_ObjChild0(pObj *Hop_Obj_t) *Hop_Obj_t { return pObj.pFanin0;                          }
func Hop_ObjChild1(pObj *Hop_Obj_t) *Hop_Obj_t { return pObj.pFanin1;                          }
func Hop_ObjChild0Copy(pObj *Hop_Obj_t) *Hop_Obj_t { assert(!(Hop_IsComplement(pObj) != 0), "!Hop_IsComplement(pObj)"); if Hop_ObjFanin0(pObj) != nil {
    return Hop_NotCond((Hop_Obj_t *)Hop_ObjFanin0(pObj).pData, Hop_ObjFaninC0(pObj));
} else {
    return nil;
}  }
func Hop_ObjChild1Copy(pObj *Hop_Obj_t) *Hop_Obj_t { assert(!(Hop_IsComplement(pObj) != 0), "!Hop_IsComplement(pObj)"); if Hop_ObjFanin1(pObj) != nil {
    return Hop_NotCond((Hop_Obj_t *)Hop_ObjFanin1(pObj).pData, Hop_ObjFaninC1(pObj));
} else {
    return nil;
}  }
func Hop_ObjChild0CopyI(pObj *Hop_Obj_t) int32 { assert(!(Hop_IsComplement(pObj) != 0), "!Hop_IsComplement(pObj)"); if Hop_ObjFanin0(pObj) != nil {
    return Abc_LitNotCond(Hop_ObjFanin0(pObj).iData, Hop_ObjFaninC0(pObj));
} else {
    return -1;
}              }
func Hop_ObjChild1CopyI(pObj *Hop_Obj_t) int32 { assert(!(Hop_IsComplement(pObj) != 0), "!Hop_IsComplement(pObj)"); if Hop_ObjFanin1(pObj) != nil {
    return Abc_LitNotCond(Hop_ObjFanin1(pObj).iData, Hop_ObjFaninC1(pObj));
} else {
    return -1;
}              }
func Hop_ObjLevel(pObj *Hop_Obj_t) int32 { return pObj.nRefs;                            }
func Hop_ObjLevelNew(pObj *Hop_Obj_t) int32 { return 1 + Hop_ObjIsExor(pObj) + Abc_MaxInt(Hop_ObjFanin0(pObj).nRefs, Hop_ObjFanin1(pObj).nRefs);       }
func Hop_ObjPhaseCompl(pObj *Hop_Obj_t) int32 { if Hop_IsComplement(pObj) != 0 {
    return !Hop_Regular(pObj).fPhase;
} else {
    return pObj.fPhase;
} }
func Hop_ObjClean(pObj *Hop_Obj_t) { *pObj = Hop_Obj_t{}; }
func Hop_ObjWhatFanin(pObj *Hop_Obj_t, pFanin *Hop_Obj_t) int32 { 
    if Hop_ObjFanin0(pObj) == pFanin { return 0; } 
    if Hop_ObjFanin1(pObj) == pFanin { return 1; } 
    assert(false, "0"); return -1; 
}
func Hop_ObjFanoutC(pObj *Hop_Obj_t, pFanout *Hop_Obj_t) int32 { 
    if Hop_ObjFanin0(pFanout) == pObj { return Hop_ObjFaninC0(pObj); } 
    if Hop_ObjFanin1(pFanout) == pObj { return Hop_ObjFaninC1(pObj); } 
    assert(false, "0"); return -1; 
}

// create the ghost of the new node
func Hop_ObjCreateGhost(p *Hop_Man_t, p0 *Hop_Obj_t, p1 *Hop_Obj_t, Type Hop_Type_t) *Hop_Obj_t {
    var pGhost *Hop_Obj_t;
    assert(Type != AIG_AND || !(Hop_ObjIsConst1(Hop_Regular(p0)) != 0), "Type != AIG_AND || !Hop_ObjIsConst1(Hop_Regular(p0))");
    assert(p1 == nil || !(Hop_ObjIsConst1(Hop_Regular(p1)) != 0), "p1 == nil || !Hop_ObjIsConst1(Hop_Regular(p1))");
    assert(Type == AIG_PI || Hop_Regular(p0) != Hop_Regular(p1), "Type == AIG_PI || Hop_Regular(p0) != Hop_Regular(p1)");
    pGhost = Hop_ManGhost(p);
    pGhost.Type = Type;
    if Hop_Regular(p0).Id < Hop_Regular(p1).Id {
        pGhost.pFanin0 = p0;
        pGhost.pFanin1 = p1;
    
    } else {
        pGhost.pFanin0 = p1;
        pGhost.pFanin1 = p0;
    }
    return pGhost;
}

// internal memory manager
func Hop_ManFetchMemory(p *Hop_Man_t) *Hop_Obj_t { 
    var pTemp *Hop_Obj_t;
    if p.pListFree == nil {
        Hop_ManAddMemory( p ); }
    pTemp = p.pListFree;
    p.pListFree = *((Hop_Obj_t **)pTemp);
    *pTemp = Hop_Obj_t{}; 
    if p.vObjs {
        assert(p.nCreated == Vec_PtrSize(p.vObjs), "p.nCreated == Vec_PtrSize(p.vObjs)");
        Vec_PtrPush( p.vObjs, pTemp );
    }
    pTemp.Id = p.nCreated++;
    return pTemp;
}
func Hop_ManRecycleMemory(p *Hop_Man_t, pEntry *Hop_Obj_t) {
    pEntry.Type = AIG_NONE; // distinquishes dead node from live node
    *((Hop_Obj_t **)pEntry) = p.pListFree;
    p.pListFree = pEntry;
}


////////////////////////////////////////////////////////////////////////
///                             ITERATORS                            ///
////////////////////////////////////////////////////////////////////////

// iterator over the primary inputs
// Original macro:
// #define Hop_ManForEachPi( p, pObj, i )                                          \
//     Vec_PtrForEachEntry( Hop_Obj_t *, p->vPis, pObj, i )

// iterator over the primary outputs
// Original macro:
// #define Hop_ManForEachPo( p, pObj, i )                                          \
//     Vec_PtrForEachEntry( Hop_Obj_t *, p->vPos, pObj, i )

// iterator over all objects, including those currently not used
// Original macro:
// #define Hop_ManForEachNode( p, pObj, i )                                        \
//     for ( i = 0; i < p->nTableSize; i++ )                                       \
//         if ( ((pObj) = p->pTable[i]) == NULL ) {} else


////////////////////////////////////////////////////////////////////////
///                    FUNCTION DECLARATIONS                         ///
////////////////////////////////////////////////////////////////////////

/*=== hopBalance.c ========================================================*/
// extern Hop_Man_t *     Hop_ManBalance( Hop_Man_t * p, int fUpdateLevel );

// extern Hop_Obj_t *     Hop_NodeBalanceBuildSuper( Hop_Man_t * p, Vec_Ptr_t * vSuper, Hop_Type_t Type, int fUpdateLevel );

/*=== hopCheck.c ========================================================*/
// extern int             Hop_ManCheck( Hop_Man_t * p );

/*=== hopDfs.c ==========================================================*/
// extern Vec_Ptr_t *     Hop_ManDfs( Hop_Man_t * p );

// extern Vec_Ptr_t *     Hop_ManDfsNode( Hop_Man_t * p, Hop_Obj_t * pNode );

// extern int             Hop_ManCountLevels( Hop_Man_t * p );

// extern void            Hop_ManCreateRefs( Hop_Man_t * p );

// extern int             Hop_DagSize( Hop_Obj_t * pObj );

// extern int             Hop_ObjFanoutCount( Hop_Obj_t * pObj, Hop_Obj_t * pPivot );

// extern void            Hop_ConeUnmark_rec( Hop_Obj_t * pObj );

// extern Hop_Obj_t *     Hop_Transfer( Hop_Man_t * pSour, Hop_Man_t * pDest, Hop_Obj_t * pObj, int nVars );

// extern Hop_Obj_t *     Hop_Compose( Hop_Man_t * p, Hop_Obj_t * pRoot, Hop_Obj_t * pFunc, int iVar );

// extern Hop_Obj_t *     Hop_Complement( Hop_Man_t * p, Hop_Obj_t * pRoot, int iVar );

// extern Hop_Obj_t *     Hop_Remap( Hop_Man_t * p, Hop_Obj_t * pRoot, unsigned uSupp, int nVars );

// extern Hop_Obj_t *     Hop_Permute( Hop_Man_t * p, Hop_Obj_t * pRoot, int nRootVars, int * pPermute );

/*=== hopMan.c ==========================================================*/
// extern Hop_Man_t *     Hop_ManStart();

// extern Hop_Man_t *     Hop_ManDup( Hop_Man_t * p );

// extern void            Hop_ManStop( Hop_Man_t * p );

// extern int             Hop_ManCleanup( Hop_Man_t * p );

// extern void            Hop_ManPrintStats( Hop_Man_t * p );

/*=== hopMem.c ==========================================================*/
// extern void            Hop_ManStartMemory( Hop_Man_t * p );

// extern void            Hop_ManStopMemory( Hop_Man_t * p );

/*=== hopObj.c ==========================================================*/
// extern Hop_Obj_t *     Hop_ObjCreatePi( Hop_Man_t * p );

// extern Hop_Obj_t *     Hop_ObjCreatePo( Hop_Man_t * p, Hop_Obj_t * pDriver );

// extern Hop_Obj_t *     Hop_ObjCreate( Hop_Man_t * p, Hop_Obj_t * pGhost );

// extern void            Hop_ObjConnect( Hop_Man_t * p, Hop_Obj_t * pObj, Hop_Obj_t * pFan0, Hop_Obj_t * pFan1 );

// extern void            Hop_ObjDisconnect( Hop_Man_t * p, Hop_Obj_t * pObj );

// extern void            Hop_ObjDelete( Hop_Man_t * p, Hop_Obj_t * pObj );

// extern void            Hop_ObjDelete_rec( Hop_Man_t * p, Hop_Obj_t * pObj );

// extern Hop_Obj_t *     Hop_ObjRepr( Hop_Obj_t * pObj );

// extern void            Hop_ObjCreateChoice( Hop_Obj_t * pOld, Hop_Obj_t * pNew );

/*=== hopOper.c =========================================================*/
// extern Hop_Obj_t *     Hop_IthVar( Hop_Man_t * p, int i );

// extern Hop_Obj_t *     Hop_Oper( Hop_Man_t * p, Hop_Obj_t * p0, Hop_Obj_t * p1, Hop_Type_t Type );

// extern Hop_Obj_t *     Hop_And( Hop_Man_t * p, Hop_Obj_t * p0, Hop_Obj_t * p1 );

// extern Hop_Obj_t *     Hop_Or( Hop_Man_t * p, Hop_Obj_t * p0, Hop_Obj_t * p1 );

// extern Hop_Obj_t *     Hop_Exor( Hop_Man_t * p, Hop_Obj_t * p0, Hop_Obj_t * p1 );

// extern Hop_Obj_t *     Hop_Mux( Hop_Man_t * p, Hop_Obj_t * pC, Hop_Obj_t * p1, Hop_Obj_t * p0 );

// extern Hop_Obj_t *     Hop_Maj( Hop_Man_t * p, Hop_Obj_t * pA, Hop_Obj_t * pB, Hop_Obj_t * pC );

// extern Hop_Obj_t *     Hop_Miter( Hop_Man_t * p, Vec_Ptr_t * vPairs );

// extern Hop_Obj_t *     Hop_CreateAnd( Hop_Man_t * p, int nVars );

// extern Hop_Obj_t *     Hop_CreateOr( Hop_Man_t * p, int nVars );

// extern Hop_Obj_t *     Hop_CreateExor( Hop_Man_t * p, int nVars );

/*=== hopTable.c ========================================================*/
// extern Hop_Obj_t *     Hop_TableLookup( Hop_Man_t * p, Hop_Obj_t * pGhost );

// extern void            Hop_TableInsert( Hop_Man_t * p, Hop_Obj_t * pObj );

// extern void            Hop_TableDelete( Hop_Man_t * p, Hop_Obj_t * pObj );

// extern int             Hop_TableCountEntries( Hop_Man_t * p );

// extern void            Hop_TableProfile( Hop_Man_t * p );

/*=== hopTruth.c ========================================================*/
// extern unsigned *      Hop_ManConvertAigToTruth( Hop_Man_t * p, Hop_Obj_t * pRoot, int nVars, Vec_Int_t * vTruth, int fMsbFirst );

// extern word            Hop_ManComputeTruth6( Hop_Man_t * p, Hop_Obj_t * pObj, int nVars );

/*=== hopUtil.c =========================================================*/
// extern void            Hop_ManIncrementTravId( Hop_Man_t * p );

// extern void            Hop_ManCleanData( Hop_Man_t * p );

// extern void            Hop_ObjCleanData_rec( Hop_Obj_t * pObj );

// extern void            Hop_ObjCollectMulti( Hop_Obj_t * pFunc, Vec_Ptr_t * vSuper );

// extern int             Hop_ObjIsMuxType( Hop_Obj_t * pObj );

// extern int             Hop_ObjRecognizeExor( Hop_Obj_t * pObj, Hop_Obj_t ** ppFan0, Hop_Obj_t ** ppFan1 );

// extern Hop_Obj_t *     Hop_ObjRecognizeMux( Hop_Obj_t * pObj, Hop_Obj_t ** ppObjT, Hop_Obj_t ** ppObjE );

// extern void            Hop_ObjPrintEqn( FILE * pFile, Hop_Obj_t * pObj, Vec_Vec_t * vLevels, int Level );

// extern void            Hop_ObjPrintVerilog( FILE * pFile, Hop_Obj_t * pObj, Vec_Vec_t * vLevels, int Level, int fOnlyAnds );

// extern void            Hop_ObjPrintVerbose( Hop_Obj_t * pObj, int fHaig );

// extern void            Hop_ManPrintVerbose( Hop_Man_t * p, int fHaig );

// extern void            Hop_ManDumpBlif( Hop_Man_t * p, char * pFileName );




//ABC_NAMESPACE_HEADER_END





////////////////////////////////////////////////////////////////////////
///                       END OF FILE                                ///
////////////////////////////////////////////////////////////////////////

