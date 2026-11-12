package amap

/**CFile****************************************************************

  FileName    [amapInt.h]

  SystemName  [ABC: Logic synthesis and verification system.]

  PackageName [Technology mapper for standard cells.]

  Synopsis    [Internal declarations.]

  Author      [Alan Mishchenko]
  
  Affiliation [UC Berkeley]

  Date        [Ver. 1.0. Started - June 20, 2005.]

  Revision    [$Id: amapInt.h,v 1.00 2005/06/20 00:00:00 alanmi Exp $]

***********************************************************************/
 




////////////////////////////////////////////////////////////////////////
///                          INCLUDES                                ///
////////////////////////////////////////////////////////////////////////

//#include "misc/extra/extra.h"
//#include "aig/aig/aig.h"
//#include "amap.h"

////////////////////////////////////////////////////////////////////////
///                         PARAMETERS                               ///
////////////////////////////////////////////////////////////////////////



//ABC_NAMESPACE_HEADER_START


// the largest gate size in the library
// (gates above this size will be ignored)
const AMAP_MAXINS = 15

const AMAP_STRING_CONST0 =     "CONST0"
const AMAP_STRING_CONST1 =     "CONST1"

// object types
 type Amap_Type_t int32
const ( 
    AMAP_OBJ_NONE Amap_Type_t = iota
    // 0: non-existent object
    AMAP_OBJ_CONST1
  // 1: constant 1 
    AMAP_OBJ_PI
      // 2: primary input
    AMAP_OBJ_PO
      // 3: primary output
    AMAP_OBJ_AND
     // 4: AND node
    AMAP_OBJ_XOR
     // 5: XOR node
    AMAP_OBJ_MUX
     // 6: MUX node
    AMAP_OBJ_VOID     // 7: unused object

)

 

////////////////////////////////////////////////////////////////////////
///                         BASIC TYPES                              ///
////////////////////////////////////////////////////////////////////////

type Amap_Pin_t = Amap_Pin_t_;
type Amap_Gat_t = Amap_Gat_t_;
type Amap_Nod_t = Amap_Nod_t_;
type Amap_Set_t = Amap_Set_t_;

type Amap_Man_t = Amap_Man_t_;
type Amap_Obj_t = Amap_Obj_t_;
type Amap_Cut_t = Amap_Cut_t_;
type Amap_Mat_t = Amap_Mat_t_;

type Amap_Man_t_ struct {
    // user data
    pPars        *Amap_Par_t;
    pLib        *Amap_Lib_t;
    // internal parameters
    fEpsilonInternal              float32;
    fAreaInv              float32;
    fUseXor                int32;    
    fUseMux                int32;    
    // internal AIG with choices
    vPis         *Vec_Ptr_t;
    vPos         *Vec_Ptr_t;
    vObjs         *Vec_Ptr_t;
    pMemObj     *Aig_MmFixed_t;
    pMemCuts      *Aig_MmFlex_t;
    pMemCutBest      *Aig_MmFlex_t;
    pMemTemp      *Aig_MmFlex_t;
    pConst1        *Amap_Obj_t;
    nObjs [AMAP_OBJ_VOID]int32;
    nLevelMax                int32;
    nChoicesGiven                int32;
    nChoicesAdded                int32;
    // mapping data-structures
    vTemp         *Vec_Int_t;
    pMatsTemp               *int32;
    ppCutsTemp       **Amap_Cut_t;
    pCutsPi        *Amap_Cut_t;
    vCuts0         *Vec_Ptr_t;
    vCuts1         *Vec_Ptr_t;
    vCuts2         *Vec_Ptr_t;
    vTempP         *Vec_Ptr_t;
    // statistics
    nCutsUsed                int32;
    nCutsTried                int32;
    nCutsTried3                int32;
    nBytesUsed                int32;
}
type Amap_Lib_t_ struct {
    pName              string;       // library name
    vGates         *Vec_Ptr_t;      // represenation of gates
    vSorted         *Vec_Ptr_t;     // gates sorted for area-only mapping
    vSelect         *Vec_Ptr_t;     // gates selected for area-only mapping
    pGate0        *Amap_Gat_t;      // the constant zero gate
    pGate1        *Amap_Gat_t;      // the constant one gate
    pGateBuf        *Amap_Gat_t;    // the buffer
    pGateInv        *Amap_Gat_t;    // the inverter
    pMemGates      *Aig_MmFlex_t;   // memory manager for objects
    fHasXor                int32;     // XOR/NXOR gates are present
    fHasMux                int32;     // MUX/NMUX gates are present
    // structural representation
    fVerbose                int32;    // enable detailed statistics
    pNodes        *Amap_Nod_t;      // representation nodes
    nNodes                int32;      // the number of nodes used
    nNodesAlloc                int32; // the number of nodes allocated
    vRules         *Vec_Ptr_t;      // the rule of AND gate
    vRulesX         *Vec_Ptr_t;     // the rule of XOR gate
    vRules3         *Vec_Int_t;     // the rule of MUX gate
    pRules              **int32;      // simplified representation
    pRulesX              **int32;     // simplified representation
    pMemSet      *Aig_MmFlex_t;     // memory manager for sets
    nSets                int32;       // the number of sets created
}
type Amap_Pin_t_ struct {
    pName              string;
    Phase                int32;
    dLoadInput             float64;
    dLoadMax             float64;
    dDelayBlockRise             float64;
    dDelayFanoutRise             float64;
    dDelayBlockFall             float64;
    dDelayFanoutFall             float64;
    dDelayBlockMax             float64;
}
type Amap_Gat_t_ struct {
    pLib        *Amap_Lib_t;            // library   
    pTwin        *Amap_Gat_t;           // twin gate
    pName              string;           // the name of the gate
    pOutName              string;        // name of the output
    dArea             float64;           // the area of the gate
    pForm              string;           // the formula describing functionality
    pFunc          *uint32;           // truth table
    Id uint32; /* C bitfield: 23 bits; review packing and truncation */   // unique number of the gate
    fMux uint32; /* C bitfield: 1 bits; review packing and truncation */   // denotes MUX-gates
    nPins uint32; /* C bitfield: 8 bits; review packing and truncation */   // number of inputs
    Pins [0]Amap_Pin_t;         // description of inputs
}
type Amap_Set_t_ struct {
    pNext        *Amap_Set_t;
    iGate uint32; /* C bitfield: 16 bits; review packing and truncation */
    fInv uint32; /* C bitfield: 1 bits; review packing and truncation */
    nIns uint32; /* C bitfield: 15 bits; review packing and truncation */
    Ins [AMAP_MAXINS]byte;// mapping from gate inputs into fanins
}
type Amap_Nod_t_ struct {
    Id uint32; /* C bitfield: 16 bits; review packing and truncation */   // ID of the node
    nSuppSize uint32; /* C bitfield: 8 bits; review packing and truncation */   // support size
    Type uint32; /* C bitfield: 8 bits; review packing and truncation */   // the type of node
    iFan0              int16;           // fanin0
    iFan1              int16;           // fanin1
    iFan2              int16;           // fanin2
    Unused              int16;          // 
    pSets        *Amap_Set_t;           // implementable gates
}
type Amap_Cut_t_ struct {
    iMat uint32; /* C bitfield: 16 bits; review packing and truncation */
    fInv uint32; /* C bitfield: 1 bits; review packing and truncation */
    nFans uint32; /* C bitfield: 15 bits; review packing and truncation */
    Fans [0]int32;
}
type Amap_Mat_t_ struct {
    pCut        *Amap_Cut_t;           // the cut
    pSet        *Amap_Set_t;           // the set
    Area              float32;           // area flow / exact area of the node
    AveFan              float32;         // edge flow of the node
    Delay              float32;          // delay of the node
}
type Amap_Obj_t_ struct {
    Type uint32; /* C bitfield: 3 bits; review packing and truncation */   
    Id uint32; /* C bitfield: 29 bits; review packing and truncation */
    IdPio uint32; /* C bitfield: 29 bits; review packing and truncation */
    fPhase uint32; /* C bitfield: 1 bits; review packing and truncation */
    fRepr uint32; /* C bitfield: 1 bits; review packing and truncation */
    fPolar uint32; /* C bitfield: 1 bits; review packing and truncation */  // pCutBest->fInv ^ pSetBest->fInv
    Level           uint32;
    nCuts           uint32;
    nRefs                int32;
    Equiv                int32;
    Fan [3]int32;
    union {
    void *             pData;
    int                iData;
    };
    // match of the node
    EstRefs              float32;        // the number of estimated fanouts
    nFouts [2]int32;      // the number of refs in each polarity
    Best         Amap_Mat_t;           // the best match of the node
}

func Amap_Regular(p *Amap_Obj_t) *Amap_Obj_t { return (Amap_Obj_t *)((ABC_PTRUINT_T)(p) & ~01);  }
func Amap_Not(p *Amap_Obj_t) *Amap_Obj_t { return (Amap_Obj_t *)((ABC_PTRUINT_T)(p) ^  01);  }
func Amap_NotCond(p *Amap_Obj_t, c int32) *Amap_Obj_t { return (Amap_Obj_t *)((ABC_PTRUINT_T)(p) ^ (c));  }
func Amap_IsComplement(p *Amap_Obj_t) int32 { return int32( (((ABC_PTRUINT_T)p) & 01));           }

func Amap_ManPiNum(p *Amap_Man_t) int32 { return p.nObjs[AMAP_OBJ_PI];                      }
func Amap_ManPoNum(p *Amap_Man_t) int32 { return p.nObjs[AMAP_OBJ_PO];                      }
func Amap_ManAndNum(p *Amap_Man_t) int32 { return p.nObjs[AMAP_OBJ_AND];                     }
func Amap_ManXorNum(p *Amap_Man_t) int32 { return p.nObjs[AMAP_OBJ_XOR];                     }
func Amap_ManMuxNum(p *Amap_Man_t) int32 { return p.nObjs[AMAP_OBJ_MUX];                     }
func Amap_ManObjNum(p *Amap_Man_t) int32 { return Vec_PtrSize(p.vObjs);                      }
func Amap_ManNodeNum(p *Amap_Man_t) int32 { return p.nObjs[AMAP_OBJ_AND] + p.nObjs[AMAP_OBJ_XOR] + p.nObjs[AMAP_OBJ_MUX];   }

func Amap_ManConst1(p *Amap_Man_t) *Amap_Obj_t { return p.pConst1;                                 }
func Amap_ManPi(p *Amap_Man_t, i int32) *Amap_Obj_t { return (Amap_Obj_t *)Vec_PtrEntry( p.vPis, i );   }
func Amap_ManPo(p *Amap_Man_t, i int32) *Amap_Obj_t { return (Amap_Obj_t *)Vec_PtrEntry( p.vPos, i );   }
func Amap_ManObj(p *Amap_Man_t, i int32) *Amap_Obj_t { return (Amap_Obj_t *)Vec_PtrEntry( p.vObjs, i );  }

func Amap_ObjIsConst1(pObj *Amap_Obj_t) int32 { return pObj.Type == AMAP_OBJ_CONST1;              }
func Amap_ObjIsPi(pObj *Amap_Obj_t) int32 { return pObj.Type == AMAP_OBJ_PI;                  }
func Amap_ObjIsPo(pObj *Amap_Obj_t) int32 { return pObj.Type == AMAP_OBJ_PO;                  }
func Amap_ObjIsAnd(pObj *Amap_Obj_t) int32 { return pObj.Type == AMAP_OBJ_AND;                 }
func Amap_ObjIsXor(pObj *Amap_Obj_t) int32 { return pObj.Type == AMAP_OBJ_XOR;                 }
func Amap_ObjIsMux(pObj *Amap_Obj_t) int32 { return pObj.Type == AMAP_OBJ_MUX;                 }
func Amap_ObjIsNode(pObj *Amap_Obj_t) int32 { return pObj.Type == AMAP_OBJ_AND || pObj.Type == AMAP_OBJ_XOR || pObj.Type == AMAP_OBJ_MUX;  }

func Amap_ObjToLit(pObj *Amap_Obj_t) int32 { return Abc_Var2Lit( Amap_Regular(pObj).Id, Amap_IsComplement(pObj) ); }
func Amap_ObjFanin0(p *Amap_Man_t, pObj *Amap_Obj_t) *Amap_Obj_t { return Amap_ManObj(p, Abc_Lit2Var(pObj.Fan[0])); }
func Amap_ObjFanin1(p *Amap_Man_t, pObj *Amap_Obj_t) *Amap_Obj_t { return Amap_ManObj(p, Abc_Lit2Var(pObj.Fan[1])); }
func Amap_ObjFanin2(p *Amap_Man_t, pObj *Amap_Obj_t) *Amap_Obj_t { return Amap_ManObj(p, Abc_Lit2Var(pObj.Fan[2])); }
func Amap_ObjFaninC0(pObj *Amap_Obj_t) int32 { return Abc_LitIsCompl(pObj.Fan[0]);              }
func Amap_ObjFaninC1(pObj *Amap_Obj_t) int32 { return Abc_LitIsCompl(pObj.Fan[1]);              }
func Amap_ObjFaninC2(pObj *Amap_Obj_t) int32 { return Abc_LitIsCompl(pObj.Fan[2]);              }
func Amap_ObjCopy(pObj *Amap_Obj_t) any { return pObj.pData;                                }
func Amap_ObjLevel(pObj *Amap_Obj_t) int32 { return pObj.Level;                                }
func Amap_ObjSetLevel(pObj *Amap_Obj_t, Level int32) { pObj.Level = Level;                               }
func Amap_ObjSetCopy(pObj *Amap_Obj_t, pCopy any) { pObj.pData = pCopy;                               }
func Amap_ObjChoice(p *Amap_Man_t, pObj *Amap_Obj_t) *Amap_Obj_t { if pObj.Equiv {
    return Amap_ManObj(p, pObj.Equiv);
} else {
    return nil;
} }
func Amap_ObjSetChoice(pObj *Amap_Obj_t, pEqu *Amap_Obj_t) { assert(pObj.Equiv==0, "pObj.Equiv==0"); if pObj.Id != pEqu.Id { pObj.Equiv = pEqu.Id; } }
func Amap_ObjPhaseReal(pObj *Amap_Obj_t) int32 { return Amap_Regular(pObj).fPhase ^ Amap_IsComplement(pObj);   }
func Amap_ObjRefsTotal(pObj *Amap_Obj_t) int32 { return pObj.nFouts[0] + pObj.nFouts[1];          }

func Amap_LibGate(p *Amap_Lib_t, i int32) *Amap_Gat_t { return (Amap_Gat_t *)Vec_PtrEntry(p.vGates, i); }
func Amap_LibNod(p *Amap_Lib_t, i int32) *Amap_Nod_t { return p.pNodes + i;           }

// returns pointer to the next cut (internal cuts only)
func Amap_ManCutNext(pCut *Amap_Cut_t) *Amap_Cut_t { return (Amap_Cut_t *)(((int *)pCut)+pCut.nFans+1); }
// returns pointer to the place of the next cut (temporary cuts only)
func Amap_ManCutNextP(pCut *Amap_Cut_t) **Amap_Cut_t { return (Amap_Cut_t **)(((int *)pCut)+pCut.nFans+1); }

// extern void Kit_DsdPrintFromTruth( unsigned * pTruth, int nVars );


////////////////////////////////////////////////////////////////////////
///                      MACRO DEFINITIONS                           ///
////////////////////////////////////////////////////////////////////////

// iterator over the primary inputs
// Original macro:
// #define Amap_ManForEachPi( p, pObj, i )                                    \
//     Vec_PtrForEachEntry( Amap_Obj_t *, p->vPis, pObj, i )

// iterator over the primary outputs
// Original macro:
// #define Amap_ManForEachPo( p, pObj, i )                                    \
//     Vec_PtrForEachEntry( Amap_Obj_t *, p->vPos, pObj, i )

// iterator over all objects, including those currently not used
// Original macro:
// #define Amap_ManForEachObj( p, pObj, i )                                   \
//     Vec_PtrForEachEntry( Amap_Obj_t *, p->vObjs, pObj, i ) if ( (pObj) == NULL ) {} else

// iterator over all nodes
// Original macro:
// #define Amap_ManForEachNode( p, pObj, i )                                  \
//     Vec_PtrForEachEntry( Amap_Obj_t *, p->vObjs, pObj, i ) if ( (pObj) == NULL || !Amap_ObjIsNode(pObj) ) {} else


// iterator through all gates of the library
// Original macro:
// #define Amap_LibForEachGate( pLib, pGate, i )                              \
//     Vec_PtrForEachEntry( Amap_Gat_t *, pLib->vGates, pGate, i )

// iterator through all pins of the gate
// Original macro:
// #define Amap_GateForEachPin( pGate, pPin )                                 \
//     for ( pPin = pGate->Pins; pPin < pGate->Pins + pGate->nPins; pPin++ )


// iterator through all cuts of the node
// Original macro:
// #define Amap_NodeForEachCut( pNode, pCut, i )                              \
//     for ( i = 0, pCut = (Amap_Cut_t *)pNode->pData; i < (int)pNode->nCuts; \
//         i++, pCut = Amap_ManCutNext(pCut) )


// iterator through all sets of one library node
// Original macro:
// #define Amap_LibNodeForEachSet( pNod, pSet )                               \
//     for ( pSet = pNod->pSets; pSet; pSet = pSet->pNext )


// iterates through each fanin of the match
// Original macro:
// #define Amap_MatchForEachFaninCompl( p, pM, pFanin, fCompl, i )            \
//     for ( i = 0; i < (int)(pM)->pCut->nFans &&                             \
//         ((pFanin = Amap_ManObj((p), Abc_Lit2Var((pM)->pCut->Fans[Abc_Lit2Var((pM)->pSet->Ins[i])]))), 1) && \
//         ((fCompl = Abc_LitIsCompl((pM)->pSet->Ins[i]) ^ Abc_LitIsCompl((pM)->pCut->Fans[Abc_Lit2Var((pM)->pSet->Ins[i])])), 1); \
//         i++ )


// iterates through each fanin of the match
// Original macro:
// #define Amap_MatchForEachFanin( p, pM, pFanin, i )                         \
//     for ( i = 0; i < (int)(pM)->pCut->nFans &&                             \
//         ((pFanin = Amap_ManObj((p), Abc_Lit2Var((pM)->pCut->Fans[Abc_Lit2Var((pM)->pSet->Ins[i])]))), 1); \
//         i++ )


////////////////////////////////////////////////////////////////////////
///                    FUNCTION DECLARATIONS                         ///
////////////////////////////////////////////////////////////////////////

/*=== amapCore.c ==========================================================*/
/*=== amapGraph.c ==========================================================*/
// extern Amap_Obj_t *  Amap_ManCreatePi( Amap_Man_t * p );

// extern Amap_Obj_t *  Amap_ManCreatePo( Amap_Man_t * p, Amap_Obj_t * pFan0 );

// extern Amap_Obj_t *  Amap_ManCreateAnd( Amap_Man_t * p, Amap_Obj_t * pFan0, Amap_Obj_t * pFan1 );

// extern Amap_Obj_t *  Amap_ManCreateXor( Amap_Man_t * p, Amap_Obj_t * pFan0, Amap_Obj_t * pFan1 );

// extern Amap_Obj_t *  Amap_ManCreateMux( Amap_Man_t * p, Amap_Obj_t * pFanC, Amap_Obj_t * pFan1, Amap_Obj_t * pFan0 );

// extern void          Amap_ManCreateChoice( Amap_Man_t * p, Amap_Obj_t * pObj );

// extern void          Amap_ManCreate( Amap_Man_t * p, Aig_Man_t * pAig );

/*=== amapLib.c ==========================================================*/
// extern Amap_Lib_t *  Amap_LibAlloc();

// extern int           Amap_LibNumPinsMax( Amap_Lib_t * p );

// extern void          Amap_LibWrite( FILE * pFile, Amap_Lib_t * pLib, int fPrintDsd );

// extern Vec_Ptr_t *   Amap_LibSelectGates( Amap_Lib_t * p, int fVerbose );

/*=== amapMan.c ==========================================================*/
// extern Amap_Man_t *  Amap_ManStart( int nNodes );

// extern void          Amap_ManStop( Amap_Man_t * p );

/*=== amapMatch.c ==========================================================*/
// extern void          Amap_ManMap( Amap_Man_t * p );

/*=== amapMerge.c ==========================================================*/
// extern void          Amap_ManMerge( Amap_Man_t * p );

/*=== amapOutput.c ==========================================================*/
// extern Vec_Ptr_t *   Amap_ManProduceMapped( Amap_Man_t * p );

/*=== amapParse.c ==========================================================*/
// extern int           Amap_LibParseEquations( Amap_Lib_t * p, int fVerbose );

/*=== amapPerm.c ==========================================================*/
/*=== amapRead.c ==========================================================*/
// extern Amap_Lib_t *  Amap_LibReadBuffer( char * pBuffer, int fVerbose );

// extern Amap_Lib_t *  Amap_LibReadFile( char * pFileName, int fVerbose );

/*=== amapRule.c ==========================================================*/
// extern short *       Amap_LibTableFindNode( Amap_Lib_t * p, int iFan0, int iFan1, int fXor );

// extern void          Amap_LibCreateRules( Amap_Lib_t * p, int fVeryVerbose );

/*=== amapUniq.c ==========================================================*/
// extern int           Amap_LibFindNode( Amap_Lib_t * pLib, int iFan0, int iFan1, int fXor );

// extern int           Amap_LibFindMux( Amap_Lib_t * p, int iFan0, int iFan1, int iFan2 );

// extern int           Amap_LibCreateVar( Amap_Lib_t * p );

// extern int           Amap_LibCreateNode( Amap_Lib_t * p, int iFan0, int iFan1, int fXor );

// extern int           Amap_LibCreateMux( Amap_Lib_t * p, int iFan0, int iFan1, int iFan2 );

// extern int **        Amap_LibLookupTableAlloc( Vec_Ptr_t * vVec, int fVerbose );




//ABC_NAMESPACE_HEADER_END





////////////////////////////////////////////////////////////////////////
///                       END OF FILE                                ///
////////////////////////////////////////////////////////////////////////

