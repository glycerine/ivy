package ssc

/**CFile****************************************************************

  FileName    [sscInt.h]

  SystemName  [ABC: Logic synthesis and verification system.]

  PackageName [Choice computation for tech-mapping.]

  Synopsis    [Internal declarations.]

  Author      [Alan Mishchenko]
  
  Affiliation [UC Berkeley]

  Date        [Ver. 1.0. Started - June 29, 2008.]

  Revision    [$Id: sscInt.h,v 1.00 2008/07/29 00:00:00 alanmi Exp $]

***********************************************************************/
 




////////////////////////////////////////////////////////////////////////
///                          INCLUDES                                ///
////////////////////////////////////////////////////////////////////////

//#include "aig/gia/gia.h"
//#include "sat/bsat/satSolver.h"
//#include "ssc.h"

////////////////////////////////////////////////////////////////////////
///                         PARAMETERS                               ///
////////////////////////////////////////////////////////////////////////


//ABC_NAMESPACE_HEADER_START


////////////////////////////////////////////////////////////////////////
///                         BASIC TYPES                              ///
////////////////////////////////////////////////////////////////////////

// choicing manager
   
type Ssc_Man_t struct {
    // user data
    pPars      *Ssc_Pars_t;          // choicing parameters
    pAig       *Gia_Man_t;           // subject AIG
    pCare       *Gia_Man_t;          // care set AIG
    // internal data
    pFraig       *Gia_Man_t;         // resulting AIG
    pSat      *sat_solver;           // recyclable SAT solver
    vId2Var       *Vec_Int_t;        // mapping of each node into its SAT var
    vVar2Id       *Vec_Int_t;        // mapping of each SAT var into its node
    vPivot       *Vec_Int_t;         // one SAT pattern
    nSatVarsPivot              int32;  // the number of variables for constraints
    nSatVars              int32;       // the current number of variables  
    // temporary storage
    vFront       *Vec_Int_t;         // supergate fanins
    vFanins       *Vec_Int_t;        // supergate fanins
    vPattern       *Vec_Int_t;       // counter-example
    vDisPairs       *Vec_Int_t;      // disproved pairs
    // SAT calls statistics
    nSimRounds              int32;     // the number of simulation rounds
    nRecycles              int32;      // the number of times SAT solver was recycled
    nCallsSince              int32;    // the number of calls since the last recycle
    nSatCalls              int32;      // the number of SAT calls
    nSatCallsUnsat              int32; // the number of unsat SAT calls
    nSatCallsSat              int32;   // the number of sat SAT calls
    nSatCallsUndec              int32; // the number of undec SAT calls
    // runtime stats
    timeSimInit          abctime;    // simulation and class computation
    timeSimSat          abctime;     // simulation of the counter-examples
    timeCnfGen          abctime;     // generation of CNF
    timeSat          abctime;        // total SAT time
    timeSatSat          abctime;     // sat
    timeSatUnsat          abctime;   // unsat
    timeSatUndec          abctime;   // undecided
    timeOther          abctime;      // other runtime
    timeTotal          abctime;      // total runtime
}

////////////////////////////////////////////////////////////////////////
///                      MACRO DEFINITIONS                           ///
////////////////////////////////////////////////////////////////////////

func Ssc_ObjSatVar(p *Ssc_Man_t, iObj int32) int32 { return Vec_IntEntry(p.vId2Var, iObj);     }
func Ssc_ObjSetSatVar(p *Ssc_Man_t, iObj int32, Num int32) { Vec_IntWriteEntry(p.vId2Var, iObj, Num);  Vec_IntWriteEntry(p.vVar2Id, Num, iObj);  }
func Ssc_ObjCleanSatVar(p *Ssc_Man_t, Num int32) { Vec_IntWriteEntry(p.vId2Var, Vec_IntEntry(p.vVar2Id, Num), Num);  Vec_IntWriteEntry(p.vVar2Id, Num, 0);                        }

func Ssc_ObjFraig(p *Ssc_Man_t, pObj *Gia_Obj_t) int32 { return pObj.Value;           }
func Ssc_ObjSetFraig(pObj *Gia_Obj_t, iNode int32) { pObj.Value = iNode;          }

////////////////////////////////////////////////////////////////////////
///                    FUNCTION DECLARATIONS                         ///
////////////////////////////////////////////////////////////////////////

/*=== sscClass.c =================================================*/
// extern void          Ssc_GiaClassesInit( Gia_Man_t * p );

// extern int           Ssc_GiaClassesRefine( Gia_Man_t * p );

// extern void          Ssc_GiaClassesCheckPairs( Gia_Man_t * p, Vec_Int_t * vDisPairs );

// extern int           Ssc_GiaSimClassRefineOneBit( Gia_Man_t * p, int i );

/*=== sscCnf.c ===================================================*/
// extern void          Ssc_CnfNodeAddToSolver( Ssc_Man_t * p, Gia_Obj_t * pObj );

/*=== sscCore.c ==================================================*/
/*=== sscSat.c ===================================================*/
// extern void          Ssc_ManSatSolverRecycle( Ssc_Man_t * p );

// extern void          Ssc_ManStartSolver( Ssc_Man_t * p );

// extern Vec_Int_t *   Ssc_ManFindPivotSat( Ssc_Man_t * p );

// extern int           Ssc_ManCheckEquivalence( Ssc_Man_t * p, int iRepr, int iObj, int fCompl );

/*=== sscSim.c ===================================================*/
// extern void          Ssc_GiaResetPiPattern( Gia_Man_t * p, int nWords );

// extern void          Ssc_GiaRandomPiPattern( Gia_Man_t * p, int nWords, Vec_Int_t * vPivot );

// extern int           Ssc_GiaTransferPiPattern( Gia_Man_t * pAig, Gia_Man_t * pCare, Vec_Int_t * vPivot );

// extern void          Ssc_GiaSavePiPattern( Gia_Man_t * p, Vec_Int_t * vPat );

// extern void          Ssc_GiaSimRound( Gia_Man_t * p );

// extern Vec_Int_t *   Ssc_GiaFindPivotSim( Gia_Man_t * p );

// extern int           Ssc_GiaEstimateCare( Gia_Man_t * p, int nWords );

/*=== sscUtil.c ===================================================*/
// extern Gia_Man_t *   Ssc_GenerateOneHot( int nVars );



//ABC_NAMESPACE_HEADER_END





////////////////////////////////////////////////////////////////////////
///                       END OF FILE                                ///
////////////////////////////////////////////////////////////////////////

