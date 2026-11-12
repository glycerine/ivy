package sfm

/**CFile****************************************************************

  FileName    [rsbInt.h]

  SystemName  [ABC: Logic synthesis and verification system.]

  PackageName [SAT-based optimization using internal don't-cares.]

  Synopsis    [Internal declarations.]

  Author      [Alan Mishchenko]
  
  Affiliation [UC Berkeley]

  Date        [Ver. 1.0. Started - June 20, 2005.]

  Revision    [$Id: rsbInt.h,v 1.00 2005/06/20 00:00:00 alanmi Exp $]

***********************************************************************/
 




////////////////////////////////////////////////////////////////////////
///                          INCLUDES                                ///
////////////////////////////////////////////////////////////////////////

//#include <stdio.h>
//#include <stdlib.h>
//#include <string.h>
//#include <assert.h>

//#include "misc/vec/vec.h"
//#include "sat/bsat/satSolver.h"
//#include "misc/util/utilNam.h"
//#include "map/scl/sclLib.h"
//#include "map/scl/sclCon.h"
//#include "misc/st/st.h"
//#include "map/mio/mio.h"
//#include "base/abc/abc.h"
//#include "sfm.h"

////////////////////////////////////////////////////////////////////////
///                         PARAMETERS                               ///
////////////////////////////////////////////////////////////////////////

//ABC_NAMESPACE_HEADER_START

const SFM_FANIN_MAX = 6
const SFM_SAT_UNDEC = 0x1234567812345678
const SFM_SAT_SAT =   0x8765432187654321

const SFM_SUPP_MAX =  8
#define SFM_WORD_MAX ((SFM_SUPP_MAX>6) ? (1<<(SFM_SUPP_MAX-6)) : 1)
const SFM_WIN_MAX =   1000
const SFM_DEC_MAX =   4
const SFM_SIM_WORDS = 8

////////////////////////////////////////////////////////////////////////
///                         BASIC TYPES                              ///
////////////////////////////////////////////////////////////////////////

type Sfm_Fun_t = Sfm_Fun_t_; 
type Sfm_Lib_t = Sfm_Lib_t_; 
type Sfm_Tim_t = Sfm_Tim_t_;
type Sfm_Mit_t = Sfm_Mit_t_;

type Sfm_Ntk_t_ struct {
    // parameters
    pPars        *Sfm_Par_t;       // parameters
    // objects
    nPis               int32;        // PI count (PIs should be first objects)
    nPos               int32;        // PO count (POs should be last objects)
    nNodes               int32;      // internal nodes
    nObjs               int32;       // total objects
    nLevelMax               int32;   // maximum level
    // user data
    vFixed        *Vec_Str_t;      // persistent objects
    vEmpty        *Vec_Str_t;      // transparent objects
    vTruths        *Vec_Wrd_t;     // truth tables
    vFanins         Vec_Wec_t;     // fanins
    vStarts        *Vec_Int_t;     // offsets
    vTruths2        *Vec_Wrd_t;    // truth tables (large)
    // attributes
    vFanouts         Vec_Wec_t;    // fanouts
    vLevels         Vec_Int_t;     // logic level
    vLevelsR         Vec_Int_t;    // logic level
    vCounts         Vec_Int_t;     // fanin counters
    vId2Var         Vec_Int_t;     // ObjId -> SatVar
    vVar2Id         Vec_Int_t;     // SatVar -> ObjId
    vCnfs        *Vec_Wec_t;       // CNFs
    vCover        *Vec_Int_t;      // temporary
    // traversal IDs
    vTravIds         Vec_Int_t;    // traversal IDs
    vTravIds2         Vec_Int_t;   // traversal IDs
    nTravIds               int32;    // traversal IDs
    nTravIds2               int32;   // traversal IDs
    // window
    iPivotNode               int32;  // window pivot
    vNodes        *Vec_Int_t;      // internal
    vDivs        *Vec_Int_t;       // divisors
    vRoots        *Vec_Int_t;      // roots
    vTfo        *Vec_Int_t;        // TFO (excluding iNode)
    // SAT solving
    pSat       *sat_solver;        // SAT solver
    nSatVars               int32;    // the number of variables
    nTryRemoves               int32; // number of fanin removals
    nTryImproves               int32;// number of node improvements
    nTryResubs               int32;  // number of resubstitutions
    nRemoves               int32;    // number of fanin removals
    nImproves               int32;   // number of node improvements
    nResubs               int32;     // number of resubstitutions
    // counter-examples
    nCexes               int32;      // number of CEXes
    vDivCexes        *Vec_Wrd_t;   // counter-examples
    // intermediate data
    vOrder        *Vec_Int_t;      // object order
    vDivVars        *Vec_Int_t;    // divisor SAT variables
    vDivIds        *Vec_Int_t;     // divisors indexes
    vLits        *Vec_Int_t;       // literals
    vValues        *Vec_Int_t;     // SAT variable values
    vClauses        *Vec_Wec_t;    // CNF clauses for the node
    vFaninMap        *Vec_Int_t;   // mapping fanins into their SAT vars
    // nodes
    nTotalNodesBeg               int32;
    nTotalEdgesBeg               int32;
    nTotalNodesEnd               int32;
    nTotalEdgesEnd               int32;
    nNodesTried               int32;
    nTotalDivs               int32;
    nSatCalls               int32;
    nTimeOuts               int32;
    nMaxDivs               int32;
    // runtime
    timeWin           abctime;
    timeDiv           abctime;
    timeCnf           abctime;
    timeSat           abctime;
    timeOther           abctime;
    timeTotal           abctime;
//    abctime           time1;
}

func Sfm_NtkPiNum(p *Sfm_Ntk_t) int32 { return p.nPis;                                   }
func Sfm_NtkPoNum(p *Sfm_Ntk_t) int32 { return p.nPos;                                   }
func Sfm_NtkNodeNum(p *Sfm_Ntk_t) int32 { return p.nObjs - p.nPis - p.nPos;              }

func Sfm_ObjIsPi(p *Sfm_Ntk_t, i int32) int32 { return i < p.nPis;                               }
func Sfm_ObjIsPo(p *Sfm_Ntk_t, i int32) int32 { return i + p.nPos >= p.nObjs;                   }
func Sfm_ObjIsNode(p *Sfm_Ntk_t, i int32) int32 { return i >= p.nPis && i + p.nPos < p.nObjs;    }
func Sfm_ObjIsFixed(p *Sfm_Ntk_t, i int32) int32 { return Vec_StrEntry(p.vFixed, i);                }
func Sfm_ObjAddsLevelArray(p *Vec_Str_t, i int32) int32 { return p == nil || Vec_StrEntry(p, i) == 0;      }
func Sfm_ObjAddsLevel(p *Sfm_Ntk_t, i int32) int32 { return Sfm_ObjAddsLevelArray(p.vEmpty, i);       }

func Sfm_ObjFiArray(p *Sfm_Ntk_t, i int32) *Vec_Int_t { return Vec_WecEntry(&p.vFanins, i);              }
func Sfm_ObjFoArray(p *Sfm_Ntk_t, i int32) *Vec_Int_t { return Vec_WecEntry(&p.vFanouts, i);             }

func Sfm_ObjFaninNum(p *Sfm_Ntk_t, i int32) int32 { return Vec_IntSize(Sfm_ObjFiArray(p, i));         }
func Sfm_ObjFanoutNum(p *Sfm_Ntk_t, i int32) int32 { return Vec_IntSize(Sfm_ObjFoArray(p, i));         }

func Sfm_ObjRefIncrement(p *Sfm_Ntk_t, iObj int32) int32 { return ++Sfm_ObjFoArray(p, iObj).nSize;          } 
func Sfm_ObjRefDecrement(p *Sfm_Ntk_t, iObj int32) int32 { return --Sfm_ObjFoArray(p, iObj).nSize;          } 

func Sfm_ObjFanin(p *Sfm_Ntk_t, i int32, k int32) int32 { return Vec_IntEntry(Sfm_ObjFiArray(p, i), k);     }
func Sfm_ObjFanout(p *Sfm_Ntk_t, i int32, k int32) int32 { return Vec_IntEntry(Sfm_ObjFoArray(p, i), k);     }

func Sfm_ObjSatVar(p *Sfm_Ntk_t, iObj int32) int32 { assert(Vec_IntEntry(&p.vId2Var, iObj) > 0, "Vec_IntEntry(&p.vId2Var, iObj) > 0"); return Vec_IntEntry(&p.vId2Var, iObj);  }
func Sfm_ObjSetSatVar(p *Sfm_Ntk_t, iObj int32, Num int32) { assert(Vec_IntEntry(&p.vId2Var, iObj) == -1, "Vec_IntEntry(&p.vId2Var, iObj) == -1"); Vec_IntWriteEntry(&p.vId2Var, iObj, Num);  Vec_IntWriteEntry(&p.vVar2Id, Num, iObj);  }
func Sfm_ObjCleanSatVar(p *Sfm_Ntk_t, Num int32) { var iObj int32  = Vec_IntEntry(&p.vVar2Id, Num); assert(Vec_IntEntry(&p.vId2Var, iObj) > 0, "Vec_IntEntry(&p.vId2Var, iObj) > 0"); Vec_IntWriteEntry(&p.vId2Var, iObj, -1);  Vec_IntWriteEntry(&p.vVar2Id, Num, -1); }
func Sfm_NtkCleanVars(p *Sfm_Ntk_t) { var i int32; for  i = 1; i < p.nSatVars; i++  {  if Vec_IntEntry(&p.vVar2Id, i) != -1 { Sfm_ObjCleanSatVar( p, i ); } } }

func Sfm_ObjLevel(p *Sfm_Ntk_t, iObj int32) int32 { return Vec_IntEntry( &p.vLevels, iObj );                         }
func Sfm_ObjSetLevel(p *Sfm_Ntk_t, iObj int32, Lev int32) { Vec_IntWriteEntry( &p.vLevels, iObj, Lev );                      }

func Sfm_ObjLevelR(p *Sfm_Ntk_t, iObj int32) int32 { return Vec_IntEntry( &p.vLevelsR, iObj );                        }
func Sfm_ObjSetLevelR(p *Sfm_Ntk_t, iObj int32, Lev int32) { Vec_IntWriteEntry( &p.vLevelsR, iObj, Lev );                     }

func Sfm_ObjUpdateFaninCount(p *Sfm_Ntk_t, iObj int32) int32 { return Vec_IntAddToEntry(&p.vCounts, iObj, -1);                  }
func Sfm_ObjResetFaninCount(p *Sfm_Ntk_t, iObj int32) { Vec_IntWriteEntry(&p.vCounts, iObj, Sfm_ObjFaninNum(p, iObj)-1); }

// extern void        Kit_DsdPrintFromTruth( unsigned * pTruth, int nVars );


////////////////////////////////////////////////////////////////////////
///                      MACRO DEFINITIONS                           ///
////////////////////////////////////////////////////////////////////////

// Original macro:
// #define Sfm_NtkForEachPi( p, i )                 for ( i = 0; i < p->nPis; i++ )

// Original macro:
// #define Sfm_NtkForEachPo( p, i )                 for ( i = p->nObjs - p->nPos; i < p->nObjs; i++ )

// Original macro:
// #define Sfm_NtkForEachNode( p, i )               for ( i = p->nPis; i + p->nPos < p->nObjs; i++ )

// Original macro:
// #define Sfm_NtkForEachNodeReverse( p, i )        for ( i = p->nObjs - p->nPos - 1; i >= p->nPis; i-- )

// Original macro:
// #define Sfm_ObjForEachFanin( p, Node, Fan, i )   for ( i = 0; i < Sfm_ObjFaninNum(p, Node)  && ((Fan = Sfm_ObjFanin(p, Node, i)), 1);  i++ )

// Original macro:
// #define Sfm_ObjForEachFanout( p, Node, Fan, i )  for ( i = 0; i < Sfm_ObjFanoutNum(p, Node) && ((Fan = Sfm_ObjFanout(p, Node, i)), 1); i++ )


////////////////////////////////////////////////////////////////////////
///                    FUNCTION DECLARATIONS                         ///
////////////////////////////////////////////////////////////////////////

/*=== sfmCnf.c ==========================================================*/
// extern void         Sfm_PrintCnf( Vec_Str_t * vCnf );

// extern int          Sfm_TruthToCnf( word Truth, word * pTruth, int nVars, Vec_Int_t * vCover, Vec_Str_t * vCnf );

// extern Vec_Wec_t *  Sfm_CreateCnf( Sfm_Ntk_t * p );

// extern void         Sfm_TranslateCnf( Vec_Wec_t * vRes, Vec_Str_t * vCnf, Vec_Int_t * vFaninMap, int iPivotVar );

/*=== sfmCore.c ==========================================================*/
/*=== sfmLib.c ==========================================================*/
// extern int          Sfm_LibFindComplInputGate( Vec_Wrd_t * vFuncs, int iGate, int nFanins, int iFanin, int * piFaninNew );

// extern Sfm_Lib_t *  Sfm_LibPrepare( int nVars, int fTwo, int fDelay, int fVerbose, int fLibVerbose );

// extern void         Sfm_LibPrint( Sfm_Lib_t * p );

// extern void         Sfm_LibStop( Sfm_Lib_t * p );

// extern int          Sfm_LibFindAreaMatch( Sfm_Lib_t * p, word * pTruth, int nFanins, int * piObj );

// extern int          Sfm_LibFindDelayMatches( Sfm_Lib_t * p, word * pTruth, int * pFanins, int nFanins, Vec_Ptr_t * vGates, Vec_Ptr_t * vFans );

// extern int          Sfm_LibImplementSimple( Sfm_Lib_t * p, word * pTruth, int * pFanins, int nFanins, Vec_Int_t * vGates, Vec_Wec_t * vFanins );

// extern int          Sfm_LibImplementGatesArea( Sfm_Lib_t * p, int * pFanins, int nFanins, int iObj, Vec_Int_t * vGates, Vec_Wec_t * vFanins );

// extern int          Sfm_LibImplementGatesDelay( Sfm_Lib_t * p, int * pFanins, Mio_Gate_t * pGateB, Mio_Gate_t * pGateT, char * pFansB, char * pFansT, Vec_Int_t * vGates, Vec_Wec_t * vFanins );

/*=== sfmNtk.c ==========================================================*/
// extern Sfm_Ntk_t *  Sfm_ConstructNetwork( Vec_Wec_t * vFanins, int nPis, int nPos );

// extern void         Sfm_NtkPrepare( Sfm_Ntk_t * p );

// extern void         Sfm_NtkUpdate( Sfm_Ntk_t * p, int iNode, int f, int iFaninNew, word uTruth );

/*=== sfmSat.c ==========================================================*/
// extern int          Sfm_NtkWindowToSolver( Sfm_Ntk_t * p );

// extern word         Sfm_ComputeInterpolant( Sfm_Ntk_t * p );

// extern word         Sfm_ComputeInterpolant2( Sfm_Ntk_t * p );

/*=== sfmTim.c ==========================================================*/
// extern Sfm_Tim_t *  Sfm_TimStart( Mio_Library_t * pLib, Scl_Con_t * pExt, Abc_Ntk_t * pNtk, int DeltaCrit );

// extern void         Sfm_TimStop( Sfm_Tim_t * p );

// extern int          Sfm_TimReadNtkDelay( Sfm_Tim_t * p );

// extern int          Sfm_TimReadObjDelay( Sfm_Tim_t * p, int iObj );

// extern void         Sfm_TimUpdateTiming( Sfm_Tim_t * p, Vec_Int_t * vTimeNodes );

// extern int          Sfm_TimSortArrayByArrival( Sfm_Tim_t * p, Vec_Int_t * vNodes, int iPivot );

// extern int          Sfm_TimPriorityNodes( Sfm_Tim_t * p, Vec_Int_t * vCands, int Window );

// extern int          Sfm_TimNodeIsNonCritical( Sfm_Tim_t * p, Abc_Obj_t * pPivot, Abc_Obj_t * pNode );

// extern int          Sfm_TimEvalRemapping( Sfm_Tim_t * p, Vec_Int_t * vFanins, Vec_Int_t * vMap, Mio_Gate_t * pGate1, char * pFans1, Mio_Gate_t * pGate2, char * pFans2 );

/*=== sfmMit.c ==========================================================*/
// extern Sfm_Mit_t *  Sfm_MitStart( Mio_Library_t * pLib, SC_Lib * pScl, Scl_Con_t * pExt, Abc_Ntk_t * pNtk, int DeltaCrit );

// extern void         Sfm_MitStop( Sfm_Mit_t * p );

// extern int          Sfm_MitReadNtkDelay( Sfm_Mit_t * p );

// extern int          Sfm_MitReadNtkMinSlack( Sfm_Mit_t * p );

// extern int          Sfm_MitReadObjDelay( Sfm_Mit_t * p, int iObj );

// extern void         Sfm_MitTransferLoad( Sfm_Mit_t * p, Abc_Obj_t * pNew, Abc_Obj_t * pOld );

// extern void         Sfm_MitTimingGrow( Sfm_Mit_t * p );

// extern void         Sfm_MitUpdateLoad( Sfm_Mit_t * p, Vec_Int_t * vTimeNodes, int fAdd );

// extern void         Sfm_MitUpdateTiming( Sfm_Mit_t * p, Vec_Int_t * vTimeNodes );

// extern int          Sfm_MitSortArrayByArrival( Sfm_Mit_t * p, Vec_Int_t * vNodes, int iPivot );

// extern int          Sfm_MitPriorityNodes( Sfm_Mit_t * p, Vec_Int_t * vCands, int Window );

// extern int          Sfm_MitNodeIsNonCritical( Sfm_Mit_t * p, Abc_Obj_t * pPivot, Abc_Obj_t * pNode );

// extern int          Sfm_MitEvalRemapping( Sfm_Mit_t * p, Vec_Int_t * vMffc, Abc_Obj_t * pObj, Vec_Int_t * vFanins, Vec_Int_t * vMap, Mio_Gate_t * pGate1, char * pFans1, Mio_Gate_t * pGate2, char * pFans2 );

/*=== sfmWin.c ==========================================================*/
// extern int          Sfm_ObjMffcSize( Sfm_Ntk_t * p, int iObj );

// extern int          Sfm_NtkCreateWindow( Sfm_Ntk_t * p, int iNode, int fVerbose );


//ABC_NAMESPACE_HEADER_END



////////////////////////////////////////////////////////////////////////
///                       END OF FILE                                ///
////////////////////////////////////////////////////////////////////////

