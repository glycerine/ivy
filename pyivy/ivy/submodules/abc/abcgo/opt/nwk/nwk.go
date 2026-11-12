package nwk

/**CFile****************************************************************

  FileName    [nwk.h]

  SystemName  [ABC: Logic synthesis and verification system.]

  PackageName [Logic network representation.]

  Synopsis    [External declarations.]

  Author      [Alan Mishchenko]
  
  Affiliation [UC Berkeley]

  Date        [Ver. 1.0. Started - June 20, 2005.]

  Revision    [$Id: nwk.h,v 1.1 2008/05/14 22:13:09 wudenni Exp $]

***********************************************************************/
 



 
////////////////////////////////////////////////////////////////////////
///                          INCLUDES                                ///
////////////////////////////////////////////////////////////////////////

//#include "aig/aig/aig.h"
//#include "aig/hop/hop.h"
//#include "misc/tim/tim.h"
//#include "map/if/if.h"
//#include "bool/bdc/bdc.h"

//#include "proof/fra/fra.h"
//#include "proof/ssw/ssw.h"
//#include "ntlnwk.h"

////////////////////////////////////////////////////////////////////////
///                         PARAMETERS                               ///
////////////////////////////////////////////////////////////////////////

//ABC_NAMESPACE_HEADER_START

////////////////////////////////////////////////////////////////////////
///                         BASIC TYPES                              ///
////////////////////////////////////////////////////////////////////////

type Nwk_Obj_t =    Nwk_Obj_t_;

// object types
 type Nwk_Type_t int32
const ( 
    NWK_OBJ_NONE Nwk_Type_t = iota
                      // 0: non-existant object
    NWK_OBJ_CI
                        // 1: combinational input
    NWK_OBJ_CO
                        // 2: combinational output
    NWK_OBJ_NODE
                      // 3: logic node
    NWK_OBJ_LATCH
                     // 4: register
    NWK_OBJ_VOID                       // 5: unused object

)

 

type Nwk_Man_t_ struct {
    // models of this design
    pName              string;          // the name of this design
    pSpec              string;          // the name of input file
    // node representation
    vCis         *Vec_Ptr_t;           // the primary inputs of the extracted part
    vCos         *Vec_Ptr_t;           // the primary outputs of the extracted part 
    vObjs         *Vec_Ptr_t;          // the objects in the topological order
    nObjs [NWK_OBJ_VOID]int32; // counter of objects of each type
    nFanioPlus                int32;     // the number of extra fanins/fanouts alloc by default
    // functionality, timing, memory, etc
    pManHop         *Hop_Man_t;        // the functionality representation
    pManTime         *Tim_Man_t;       // the timing manager
    pLutLib          *If_LibLut_t;        // the LUT library
    pMemObjs      *Aig_MmFlex_t;       // memory for objects
    vTemp         *Vec_Ptr_t;          // array used for incremental updates
    nTravIds                int32;       // the counter of traversal IDs
    nRealloced                int32;     // the number of realloced nodes
    // sequential information
    nLatches                int32;       // the total number of latches 
    nTruePis                int32;       // the number of true primary inputs
    nTruePos                int32;       // the number of true primary outputs
}

type Nwk_Obj_t_ struct {
    pMan         *Nwk_Man_t;           // the manager  
    pFunc         *Hop_Obj_t;          // functionality
    pCopy              any;          // temporary pointer
    union {
        void *         pNext;          // temporary pointer
        int            iTemp;          // temporary number
    };
    // node information
    Type uint32; /* C bitfield: 3 bits; review packing and truncation */  // object type
    fInvert uint32; /* C bitfield: 1 bits; review packing and truncation */  // complemented attribute
    MarkA uint32; /* C bitfield: 1 bits; review packing and truncation */  // temporary mark  
    MarkB uint32; /* C bitfield: 1 bits; review packing and truncation */  // temporary mark
    MarkC uint32; /* C bitfield: 1 bits; review packing and truncation */  // temporary mark
    PioId uint32; /* C bitfield: 25 bits; review packing and truncation */  // number of this node in the PI/PO list
    Id                int32;             // unique ID
    TravId                int32;         // traversal ID
    // timing information
    Level                int32;          // the topological level
    tArrival              float32;       // the arrival time
    tRequired              float32;      // the required time
    tSlack              float32;         // the slack
    // fanin/fanout representation
    nFanins                int32;        // the number of fanins
    nFanouts                int32;       // the number of fanouts
    nFanioAlloc                int32;    // the number of allocated fanins/fanouts
    pFanio        **Nwk_Obj_t;         // fanins/fanouts
}

////////////////////////////////////////////////////////////////////////
///                      MACRO DEFINITIONS                           ///
////////////////////////////////////////////////////////////////////////


////////////////////////////////////////////////////////////////////////
///                      INLINED FUNCTIONS                           ///
////////////////////////////////////////////////////////////////////////

func Nwk_ManCiNum(p *Nwk_Man_t) int32 { return p.nObjs[NWK_OBJ_CI];                } 
func Nwk_ManCoNum(p *Nwk_Man_t) int32 { return p.nObjs[NWK_OBJ_CO];                } 
func Nwk_ManNodeNum(p *Nwk_Man_t) int32 { return p.nObjs[NWK_OBJ_NODE];              } 
func Nwk_ManLatchNum(p *Nwk_Man_t) int32 { return p.nObjs[NWK_OBJ_LATCH];             } 
func Nwk_ManObjNumMax(p *Nwk_Man_t) int32 { return Vec_PtrSize(p.vObjs);               }

func Nwk_ManCi(p *Nwk_Man_t, i int32) *Nwk_Obj_t { return (Nwk_Obj_t *)Vec_PtrEntry( p.vCis, i );          } 
func Nwk_ManCo(p *Nwk_Man_t, i int32) *Nwk_Obj_t { return (Nwk_Obj_t *)Vec_PtrEntry( p.vCos, i );          } 
func Nwk_ManObj(p *Nwk_Man_t, i int32) *Nwk_Obj_t { return (Nwk_Obj_t *)Vec_PtrEntry( p.vObjs, i );         } 

func Nwk_ObjId(p *Nwk_Obj_t) int32 { return p.Id;                               } 
func Nwk_ObjPioNum(p *Nwk_Obj_t) int32 { return p.PioId;                            } 
func Nwk_ObjFaninNum(p *Nwk_Obj_t) int32 { return p.nFanins;                          } 
func Nwk_ObjFanoutNum(p *Nwk_Obj_t) int32 { return p.nFanouts;                         } 

func Nwk_ObjFanin0(p *Nwk_Obj_t) *Nwk_Obj_t { return p.pFanio[0];                        } 
func Nwk_ObjFanout0(p *Nwk_Obj_t) *Nwk_Obj_t { return p.pFanio[p.nFanins];               } 
func Nwk_ObjFanin(p *Nwk_Obj_t, i int32) *Nwk_Obj_t { return p.pFanio[i];                        } 
func Nwk_ObjFanout(p *Nwk_Obj_t, i int32) *Nwk_Obj_t { return p.pFanio[p.nFanins+1];             } 

func Nwk_ObjIsNone(p *Nwk_Obj_t) int32 { return p.Type == NWK_OBJ_NONE;             } 
func Nwk_ObjIsCi(p *Nwk_Obj_t) int32 { return p.Type == NWK_OBJ_CI;               } 
func Nwk_ObjIsCo(p *Nwk_Obj_t) int32 { return p.Type == NWK_OBJ_CO;               } 
func Nwk_ObjIsNode(p *Nwk_Obj_t) int32 { return p.Type == NWK_OBJ_NODE;             } 
func Nwk_ObjIsLatch(p *Nwk_Obj_t) int32 { return p.Type == NWK_OBJ_LATCH;            } 
func Nwk_ObjIsPi(p *Nwk_Obj_t) int32 { return Nwk_ObjIsCi(p) && (p.pMan.pManTime == nil || Tim_ManBoxForCi(p.pMan.pManTime, p.PioId) == -1); } 
func Nwk_ObjIsPo(p *Nwk_Obj_t) int32 { return Nwk_ObjIsCo(p) && (p.pMan.pManTime == nil || Tim_ManBoxForCo(p.pMan.pManTime, p.PioId) == -1);  }
func Nwk_ObjIsLi(p *Nwk_Obj_t) int32 { return p.pMan.nTruePos && Nwk_ObjIsCo(p) && int32(p.PioId) >= p.pMan.nTruePos;       } 
func Nwk_ObjIsLo(p *Nwk_Obj_t) int32 { return p.pMan.nTruePis && Nwk_ObjIsCi(p) && int32(p.PioId) >= p.pMan.nTruePis;       } 

func Nwk_ObjArrival(pObj *Nwk_Obj_t) float32 { return pObj.tArrival;           }
func Nwk_ObjRequired(pObj *Nwk_Obj_t) float32 { return pObj.tRequired;          }
func Nwk_ObjSlack(pObj *Nwk_Obj_t) float32 { return pObj.tSlack;             }
func Nwk_ObjSetArrival(pObj *Nwk_Obj_t, Time float32) { pObj.tArrival   = Time;         }
func Nwk_ObjSetRequired(pObj *Nwk_Obj_t, Time float32) { pObj.tRequired  = Time;         }
func Nwk_ObjSetSlack(pObj *Nwk_Obj_t, Time float32) { pObj.tSlack     = Time;         }

func Nwk_ObjLevel(pObj *Nwk_Obj_t) int32 { return pObj.Level;              }
func Nwk_ObjSetLevel(pObj *Nwk_Obj_t, Level int32) { pObj.Level = Level;             }

func Nwk_ObjSetTravId(pObj *Nwk_Obj_t, TravId int32) { pObj.TravId = TravId;                           }
func Nwk_ObjSetTravIdCurrent(pObj *Nwk_Obj_t) { pObj.TravId = pObj.pMan.nTravIds;             }
func Nwk_ObjSetTravIdPrevious(pObj *Nwk_Obj_t) { pObj.TravId = pObj.pMan.nTravIds - 1;         }
func Nwk_ObjIsTravIdCurrent(pObj *Nwk_Obj_t) int32 { return pObj.TravId == pObj.pMan.nTravIds;     }
func Nwk_ObjIsTravIdPrevious(pObj *Nwk_Obj_t) int32 { return pObj.TravId == pObj.pMan.nTravIds - 1; }

func Nwk_ManTimeEqual(f1 float32, f2 float32, Eps float32) int32 { return (f1 < f2 + Eps) && (f2 < f1 + Eps);  }
func Nwk_ManTimeLess(f1 float32, f2 float32, Eps float32) int32 { return (f1 < f2 + Eps);                     }
func Nwk_ManTimeMore(f1 float32, f2 float32, Eps float32) int32 { return (f1 + Eps > f2);                     }

////////////////////////////////////////////////////////////////////////
///                         ITERATORS                                ///
////////////////////////////////////////////////////////////////////////

// Original macro:
// #define Nwk_ManForEachCi( p, pObj, i )                                     \
//     Vec_PtrForEachEntry( Nwk_Obj_t *, p->vCis, pObj, i )

// Original macro:
// #define Nwk_ManForEachCo( p, pObj, i )                                     \
//     Vec_PtrForEachEntry( Nwk_Obj_t *, p->vCos, pObj, i )

// Original macro:
// #define Nwk_ManForEachPi( p, pObj, i )                                     \
//     Vec_PtrForEachEntry( Nwk_Obj_t *, p->vCis, pObj, i )                                \
//         if ( !Nwk_ObjIsPi(pObj) ) {} else

// Original macro:
// #define Nwk_ManForEachPo( p, pObj, i )                                     \
//     Vec_PtrForEachEntry( Nwk_Obj_t *, p->vCos, pObj, i )                                \
//         if ( !Nwk_ObjIsPo(pObj) ) {} else

// Original macro:
// #define Nwk_ManForEachObj( p, pObj, i )                                    \
//     for ( i = 0; (i < Vec_PtrSize(p->vObjs)) && (((pObj) = (Nwk_Obj_t *)Vec_PtrEntry(p->vObjs, i)), 1); i++ ) \
//         if ( pObj == NULL ) {} else

// Original macro:
// #define Nwk_ManForEachNode( p, pObj, i )                                   \
//     for ( i = 0; (i < Vec_PtrSize(p->vObjs)) && (((pObj) = (Nwk_Obj_t *)Vec_PtrEntry(p->vObjs, i)), 1); i++ ) \
//         if ( (pObj) == NULL || !Nwk_ObjIsNode(pObj) ) {} else

// Original macro:
// #define Nwk_ManForEachLatch( p, pObj, i )                                  \
//     for ( i = 0; (i < Vec_PtrSize(p->vObjs)) && (((pObj) = (Nwk_Obj_t *)Vec_PtrEntry(p->vObjs, i)), 1); i++ ) \
//         if ( (pObj) == NULL || !Nwk_ObjIsLatch(pObj) ) {} else


// Original macro:
// #define Nwk_ObjForEachFanin( pObj, pFanin, i )                                  \
//     for ( i = 0; (i < (int)(pObj)->nFanins) && ((pFanin) = (pObj)->pFanio[i]); i++ )

// Original macro:
// #define Nwk_ObjForEachFanout( pObj, pFanout, i )                                \
//     for ( i = 0; (i < (int)(pObj)->nFanouts) && ((pFanout) = (pObj)->pFanio[(pObj)->nFanins+i]); i++ )


// sequential iterators
// Original macro:
// #define Nwk_ManForEachPiSeq( p, pObj, i )                                           \
//     Vec_PtrForEachEntryStop( Nwk_Obj_t *, p->vCis, pObj, i, (p)->nTruePis )

// Original macro:
// #define Nwk_ManForEachPoSeq( p, pObj, i )                                           \
//     Vec_PtrForEachEntryStop( Nwk_Obj_t *, p->vCos, pObj, i, (p)->nTruePos )

// Original macro:
// #define Nwk_ManForEachLoSeq( p, pObj, i )                                           \
//     for ( i = 0; (i < (p)->nLatches) && (((pObj) = (Nwk_Obj_t *)Vec_PtrEntry(p->vCis, i+(p)->nTruePis)), 1); i++ )

// Original macro:
// #define Nwk_ManForEachLiSeq( p, pObj, i )                                           \
//     for ( i = 0; (i < (p)->nLatches) && (((pObj) = (Nwk_Obj_t *)Vec_PtrEntry(p->vCos, i+(p)->nTruePos)), 1); i++ )

// Original macro:
// #define Nwk_ManForEachLiLoSeq( p, pObjLi, pObjLo, i )                               \
//     for ( i = 0; (i < (p)->nLatches) && (((pObjLi) = Nwk_ManCo(p, i+(p)->nTruePos)), 1)        \
//         && (((pObjLo) = Nwk_ManCi(p, i+(p)->nTruePis)), 1); i++ )



////////////////////////////////////////////////////////////////////////
///                    FUNCTION DECLARATIONS                         ///
////////////////////////////////////////////////////////////////////////

/*=== nwkAig.c ==========================================================*/
extern ABC_DLL Vec_Ptr_t *     Nwk_ManDeriveRetimingCut( Aig_Man_t * p, int fForward, int fVerbose );
/*=== nwkBidec.c ==========================================================*/
extern ABC_DLL void            Nwk_ManBidecResyn( Nwk_Man_t * pNtk, int fVerbose );
extern ABC_DLL Hop_Obj_t *     Nwk_NodeIfNodeResyn( Bdc_Man_t * p, Hop_Man_t * pHop, Hop_Obj_t * pRoot, int nVars, Vec_Int_t * vTruth, unsigned * puCare, float dProb );
/*=== nwkCheck.c ==========================================================*/
extern ABC_DLL int             Nwk_ManCheck( Nwk_Man_t * p );
/*=== nwkDfs.c ==========================================================*/
extern ABC_DLL int             Nwk_ManVerifyTopoOrder( Nwk_Man_t * pNtk );
extern ABC_DLL int             Nwk_ManLevelBackup( Nwk_Man_t * pNtk );
extern ABC_DLL int             Nwk_ManLevel( Nwk_Man_t * pNtk );
extern ABC_DLL int             Nwk_ManLevelMax( Nwk_Man_t * pNtk );
extern ABC_DLL Vec_Vec_t *     Nwk_ManLevelize( Nwk_Man_t * pNtk );
extern ABC_DLL Vec_Ptr_t *     Nwk_ManDfs( Nwk_Man_t * pNtk );
extern ABC_DLL Vec_Ptr_t *     Nwk_ManDfsNodes( Nwk_Man_t * pNtk, Nwk_Obj_t ** ppNodes, int nNodes );
extern ABC_DLL Vec_Ptr_t *     Nwk_ManDfsReverse( Nwk_Man_t * pNtk );
extern ABC_DLL Vec_Ptr_t *     Nwk_ManSupportNodes( Nwk_Man_t * pNtk, Nwk_Obj_t ** ppNodes, int nNodes );
extern ABC_DLL void            Nwk_ManSupportSum( Nwk_Man_t * pNtk );
extern ABC_DLL int             Nwk_ObjMffcLabel( Nwk_Obj_t * pNode );
/*=== nwkFanio.c ==========================================================*/
extern ABC_DLL void            Nwk_ObjCollectFanins( Nwk_Obj_t * pNode, Vec_Ptr_t * vNodes );
extern ABC_DLL void            Nwk_ObjCollectFanouts( Nwk_Obj_t * pNode, Vec_Ptr_t * vNodes );
extern ABC_DLL int             Nwk_ObjFindFanin( Nwk_Obj_t * pObj, Nwk_Obj_t * pFanin );
extern ABC_DLL int             Nwk_ObjFindFanout( Nwk_Obj_t * pObj, Nwk_Obj_t * pFanout );
extern ABC_DLL void            Nwk_ObjAddFanin( Nwk_Obj_t * pObj, Nwk_Obj_t * pFanin );
extern ABC_DLL void            Nwk_ObjDeleteFanin( Nwk_Obj_t * pObj, Nwk_Obj_t * pFanin );
extern ABC_DLL void            Nwk_ObjPatchFanin( Nwk_Obj_t * pObj, Nwk_Obj_t * pFaninOld, Nwk_Obj_t * pFaninNew );
extern ABC_DLL void            Nwk_ObjTransferFanout( Nwk_Obj_t * pNodeFrom, Nwk_Obj_t * pNodeTo );
extern ABC_DLL void            Nwk_ObjReplace( Nwk_Obj_t * pNodeOld, Nwk_Obj_t * pNodeNew );
/*=== nwkFlow.c ============================================================*/
extern ABC_DLL Vec_Ptr_t *     Nwk_ManRetimeCutForward( Nwk_Man_t * pMan, int nLatches, int fVerbose );
extern ABC_DLL Vec_Ptr_t *     Nwk_ManRetimeCutBackward( Nwk_Man_t * pMan, int nLatches, int fVerbose );
/*=== nwkMan.c ============================================================*/
extern ABC_DLL Nwk_Man_t *     Nwk_ManAlloc();
extern ABC_DLL void            Nwk_ManFree( Nwk_Man_t * p );
extern ABC_DLL float           Nwl_ManComputeTotalSwitching( Nwk_Man_t * pNtk );
extern ABC_DLL void            Nwk_ManPrintStats( Nwk_Man_t * p, If_LibLut_t * pLutLib, int fSaveBest, int fDumpResult, int fPower, Ntl_Man_t * pNtl );
/*=== nwkMap.c ============================================================*/
extern ABC_DLL Nwk_Man_t *     Nwk_MappingIf( Aig_Man_t * p, Tim_Man_t * pManTime, If_Par_t * pPars );
/*=== nwkObj.c ============================================================*/
extern ABC_DLL Nwk_Obj_t *     Nwk_ManCreateCi( Nwk_Man_t * pMan, int nFanouts );
extern ABC_DLL Nwk_Obj_t *     Nwk_ManCreateCo( Nwk_Man_t * pMan );
extern ABC_DLL Nwk_Obj_t *     Nwk_ManCreateNode( Nwk_Man_t * pMan, int nFanins, int nFanouts );
extern ABC_DLL Nwk_Obj_t *     Nwk_ManCreateBox( Nwk_Man_t * pMan, int nFanins, int nFanouts );
extern ABC_DLL Nwk_Obj_t *     Nwk_ManCreateLatch( Nwk_Man_t * pMan );
extern ABC_DLL void            Nwk_ManDeleteNode( Nwk_Obj_t * pObj );
extern ABC_DLL void            Nwk_ManDeleteNode_rec( Nwk_Obj_t * pObj );
/*=== nwkSpeedup.c ============================================================*/
extern ABC_DLL Aig_Man_t *     Nwk_ManSpeedup( Nwk_Man_t * pNtk, int fUseLutLib, int Percentage, int Degree, int fVerbose, int fVeryVerbose );
/*=== nwkStrash.c ============================================================*/
extern ABC_DLL Aig_Man_t *     Nwk_ManStrash( Nwk_Man_t * pNtk );
/*=== nwkTiming.c ============================================================*/
extern ABC_DLL int             Nwk_ManVerifyTiming(  Nwk_Man_t * pNtk );
extern ABC_DLL void            Nwk_ManDelayTraceSortPins( Nwk_Obj_t * pNode, int * pPinPerm, float * pPinDelays );
extern ABC_DLL float           Nwk_ManDelayTraceLut( Nwk_Man_t * pNtk );
extern ABC_DLL void            Nwk_ManDelayTracePrint( Nwk_Man_t * pNtk );
extern ABC_DLL void            Nwk_ManUpdate( Nwk_Obj_t * pObj, Nwk_Obj_t * pObjNew, Vec_Vec_t * vLevels );
extern ABC_DLL int             Nwk_ManVerifyLevel( Nwk_Man_t * pNtk );
/*=== nwkUtil.c ============================================================*/
extern ABC_DLL void            Nwk_ManIncrementTravId( Nwk_Man_t * pNtk );
extern ABC_DLL int             Nwk_ManGetFaninMax( Nwk_Man_t * pNtk );
extern ABC_DLL int             Nwk_ManGetTotalFanins( Nwk_Man_t * pNtk );
extern ABC_DLL int             Nwk_ManPiNum( Nwk_Man_t * pNtk );
extern ABC_DLL int             Nwk_ManPoNum( Nwk_Man_t * pNtk );
extern ABC_DLL int             Nwk_ManGetAigNodeNum( Nwk_Man_t * pNtk );
extern ABC_DLL int             Nwk_NodeCompareLevelsIncrease( Nwk_Obj_t ** pp1, Nwk_Obj_t ** pp2 );
extern ABC_DLL int             Nwk_NodeCompareLevelsDecrease( Nwk_Obj_t ** pp1, Nwk_Obj_t ** pp2 );
extern ABC_DLL void            Nwk_ObjPrint( Nwk_Obj_t * pObj );
extern ABC_DLL void            Nwk_ManDumpBlif( Nwk_Man_t * pNtk, char * pFileName, Vec_Ptr_t * vCiNames, Vec_Ptr_t * vCoNames );
extern ABC_DLL void            Nwk_ManPrintFanioNew( Nwk_Man_t * pNtk );
extern ABC_DLL void            Nwk_ManCleanMarks( Nwk_Man_t * pNtk );
extern ABC_DLL void            Nwk_ManMinimumBase( Nwk_Man_t * pNtk, int fVerbose );
extern ABC_DLL void            Nwk_ManRemoveDupFanins( Nwk_Man_t * pNtk, int fVerbose );



//ABC_NAMESPACE_HEADER_END





////////////////////////////////////////////////////////////////////////
///                       END OF FILE                                ///
////////////////////////////////////////////////////////////////////////

