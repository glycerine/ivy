package cgt

/**CFile****************************************************************

  FileName    [cgtInt.h]

  SystemName  [ABC: Logic synthesis and verification system.]

  PackageName [Clock gating package.]

  Synopsis    [Internal declarations.]

  Author      [Alan Mishchenko]
  
  Affiliation [UC Berkeley]

  Date        [Ver. 1.0. Started - June 20, 2005.]

  Revision    [$Id: cgtInt.h,v 1.00 2005/06/20 00:00:00 alanmi Exp $]

***********************************************************************/
 




////////////////////////////////////////////////////////////////////////
///                          INCLUDES                                ///
////////////////////////////////////////////////////////////////////////

//#include "aig/saig/saig.h"
//#include "sat/bsat/satSolver.h"
//#include "sat/cnf/cnf.h"
//#include "cgt.h"

////////////////////////////////////////////////////////////////////////
///                         PARAMETERS                               ///
////////////////////////////////////////////////////////////////////////



//ABC_NAMESPACE_HEADER_START


////////////////////////////////////////////////////////////////////////
///                         BASIC TYPES                              ///
////////////////////////////////////////////////////////////////////////

   
type Cgt_Man_t struct {
    // user's data
    pPars   *Cgt_Par_t;          // user's parameters
    pAig   *Aig_Man_t;           // user's AIG manager
    vUseful   *Vec_Int_t;        // user's candidate nodes
    // user's constraints
    pCare   *Aig_Man_t;          // constraint cones
    vSuppsInv   *Vec_Vec_t;      // inverse support of the constraints
    // result of clock-gating
    vGatesAll   *Vec_Vec_t;      // the computed clock-gates
    vGates   *Vec_Ptr_t;         // the selected clock-gates
    // internal data
    pFrame   *Aig_Man_t;         // clock gate AIG manager
    vFanout   *Vec_Ptr_t;        // temporary storage for fanouts
    vVisited   *Vec_Ptr_t;       // temporary storage for visited nodes
    // SAT solving
    pPart   *Aig_Man_t;          // partition
    pCnf   *Cnf_Dat_t;           // CNF of the partition
    pSat  *sat_solver;           // SAT solver 
    vPatts   *Vec_Ptr_t;         // simulation patterns
    nPatts          int32;         // the number of patterns accumulated
    nPattWords          int32;     // the number of pattern words
    // statistics
    nRecycles          int32;      // recycles 
    nCalls          int32;         // total calls
    nCallsSat          int32;      // satisfiable calls
    nCallsUnsat          int32;    // unsatisfiable calls  
    nCallsUndec          int32;    // undecided calls
    nCallsFiltered          int32; // filtered out calls
    timeAig      abctime;        // constructing AIG
    timePrepare      abctime;    // partitioning and SAT solving
    timeSat      abctime;        // total runtime
    timeSatSat      abctime;     // satisfiable runtime 
    timeSatUnsat      abctime;   // unsatisfiable runtime 
    timeSatUndec      abctime;   // undecided runtime
    timeDecision      abctime;   // making decision about what gates to use
    timeOther      abctime;      // other runtime
    timeTotal      abctime;      // total runtime
}

////////////////////////////////////////////////////////////////////////
///                      MACRO DEFINITIONS                           ///
////////////////////////////////////////////////////////////////////////

////////////////////////////////////////////////////////////////////////
///                    FUNCTION DECLARATIONS                         ///
////////////////////////////////////////////////////////////////////////

/*=== cgtAig.c ==========================================================*/
// extern void             Cgt_ManDetectCandidates( Aig_Man_t * pAig, Vec_Int_t * vUseful, Aig_Obj_t * pObj, int nLevelMax, Vec_Ptr_t * vCands );

// extern Aig_Man_t *      Cgt_ManDeriveAigForGating( Cgt_Man_t * p );

// extern Aig_Man_t *      Cgt_ManDupPartition( Aig_Man_t * pAig, int nVarsMin, int nFlopsMin, int iStart, Aig_Man_t * pCare, Vec_Vec_t * vSuppsInv, int * pnOutputs );

// extern Aig_Man_t *      Cgt_ManDeriveGatedAig( Aig_Man_t * pAig, Vec_Vec_t * vGates, int fReduce, int * pnUsedNodes );

/*=== cgtDecide.c ==========================================================*/
// extern Vec_Vec_t *      Cgt_ManDecideSimple( Aig_Man_t * pAig, Vec_Vec_t * vGatesAll, int nOdcMax, int fVerbose );

// extern Vec_Vec_t *      Cgt_ManDecideArea( Aig_Man_t * pAig, Vec_Vec_t * vGatesAll, int nOdcMax, int fVerbose );

/*=== cgtMan.c ==========================================================*/
// extern Cgt_Man_t *      Cgt_ManCreate( Aig_Man_t * pAig, Aig_Man_t * pCare, Cgt_Par_t * pPars );

// extern void             Cgt_ManClean( Cgt_Man_t * p );

// extern void             Cgt_ManStop( Cgt_Man_t * p );

/*=== cgtSat.c ==========================================================*/
// extern int              Cgt_CheckImplication( Cgt_Man_t * p, Aig_Obj_t * pGate, Aig_Obj_t * pFlop );




//ABC_NAMESPACE_HEADER_END





////////////////////////////////////////////////////////////////////////
///                       END OF FILE                                ///
////////////////////////////////////////////////////////////////////////

