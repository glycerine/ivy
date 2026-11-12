package dch

/**CFile****************************************************************

  FileName    [dchInt.h]

  SystemName  [ABC: Logic synthesis and verification system.]

  PackageName [Choice computation for tech-mapping.]

  Synopsis    [External declarations.]

  Author      [Alan Mishchenko]
  
  Affiliation [UC Berkeley]

  Date        [Ver. 1.0. Started - June 29, 2008.]

  Revision    [$Id: dchInt.h,v 1.00 2008/07/29 00:00:00 alanmi Exp $]

***********************************************************************/
 




////////////////////////////////////////////////////////////////////////
///                          INCLUDES                                ///
////////////////////////////////////////////////////////////////////////

//#include "aig/aig/aig.h"
//#include "sat/bsat/satSolver.h"
//#include "dch.h"

////////////////////////////////////////////////////////////////////////
///                         PARAMETERS                               ///
////////////////////////////////////////////////////////////////////////



//ABC_NAMESPACE_HEADER_START


////////////////////////////////////////////////////////////////////////
///                         BASIC TYPES                              ///
////////////////////////////////////////////////////////////////////////

// equivalence classes
type Dch_Cla_t = Dch_Cla_t_;

// choicing manager
   
type Dch_Man_t struct {
    // parameters
    pPars      *Dch_Pars_t;          // choicing parameters
    // AIGs used in the package
//    Vec_Ptr_t *      vAigs;          // user-given AIGs
    pAigTotal       *Aig_Man_t;      // intermediate AIG
    pAigFraig       *Aig_Man_t;      // final AIG
    // equivalence classes
    ppClasses       *Dch_Cla_t;      // equivalence classes of nodes
    pReprsProved      **Aig_Obj_t;   // equivalences proved
    // SAT solving
    pSat      *sat_solver;           // recyclable SAT solver
    nSatVars              int32;       // the counter of SAT variables
    pSatVars             *int32;       // mapping of each node into its SAT var
    vUsedNodes       *Vec_Ptr_t;     // nodes whose SAT vars are assigned
    nRecycles              int32;      // the number of times SAT solver was recycled
    nCallsSince              int32;    // the number of calls since the last recycle
    vFanins       *Vec_Ptr_t;        // fanins of the CNF node
    vSimRoots       *Vec_Ptr_t;      // the roots of cand const 1 nodes to simulate
    vSimClasses       *Vec_Ptr_t;    // the roots of cand equiv classes to simulate
    // solver cone size
    nConeThis              int32;
    nConeMax              int32;
    // SAT calls statistics
    nSatCalls              int32;      // the number of SAT calls
    nSatProof              int32;      // the number of proofs
    nSatFailsReal              int32;  // the number of timeouts
    nSatCallsUnsat              int32; // the number of unsat SAT calls
    nSatCallsSat              int32;   // the number of sat SAT calls
    // choice node statistics
    nLits              int32;          // the number of lits in the cand equiv classes
    nReprs              int32;         // the number of proved equivalent pairs
    nEquivs              int32;        // the number of final equivalences
    nChoices              int32;       // the number of final choice nodes
    // runtime stats
    timeSimInit          abctime;    // simulation and class computation
    timeSimSat          abctime;     // simulation of the counter-examples
    timeSat          abctime;        // solving SAT
    timeSatSat          abctime;     // sat
    timeSatUnsat          abctime;   // unsat
    timeSatUndec          abctime;   // undecided
    timeChoice          abctime;     // choice computation
    timeOther          abctime;      // other runtime
    timeTotal          abctime;      // total runtime
}

////////////////////////////////////////////////////////////////////////
///                      MACRO DEFINITIONS                           ///
////////////////////////////////////////////////////////////////////////

func Dch_ObjSatNum(p *Dch_Man_t, pObj *Aig_Obj_t) int32 { return p.pSatVars[pObj.Id]; }
func Dch_ObjSetSatNum(p *Dch_Man_t, pObj *Aig_Obj_t, Num int32) { p.pSatVars[pObj.Id] = Num;  }

func Dch_ObjFraig(pObj *Aig_Obj_t) *Aig_Obj_t { return (Aig_Obj_t *)pObj.pData;  }
func Dch_ObjSetFraig(pObj *Aig_Obj_t, pNode *Aig_Obj_t) { pObj.pData = pNode; }

func Dch_ObjIsConst1Cand(pAig *Aig_Man_t, pObj *Aig_Obj_t) int32 {
    return Aig_ObjRepr(pAig, pObj) == Aig_ManConst1(pAig);
}
func Dch_ObjSetConst1Cand(pAig *Aig_Man_t, pObj *Aig_Obj_t) {
    assert(!(Dch_ObjIsConst1Cand( pAig, pObj ) != 0), "!Dch_ObjIsConst1Cand( pAig, pObj )");
    Aig_ObjSetRepr( pAig, pObj, Aig_ManConst1(pAig) );
}

////////////////////////////////////////////////////////////////////////
///                    FUNCTION DECLARATIONS                         ///
////////////////////////////////////////////////////////////////////////

/*=== dchAig.c ===================================================*/
/*=== dchChoice.c ===================================================*/
// extern int           Dch_DeriveChoiceCountReprs( Aig_Man_t * pAig );

// extern int           Dch_DeriveChoiceCountEquivs( Aig_Man_t * pAig );

// extern Aig_Man_t *   Dch_DeriveChoiceAig( Aig_Man_t * pAig, int fSkipRedSupps );

/*=== dchClass.c =================================================*/
// extern Dch_Cla_t *   Dch_ClassesStart( Aig_Man_t * pAig );

// extern void          Dch_ClassesSetData( Dch_Cla_t * p, void * pManData,
//                          unsigned (*pFuncNodeHash)(void *,Aig_Obj_t *),
//                          int (*pFuncNodeIsConst)(void *,Aig_Obj_t *),
//                          int (*pFuncNodesAreEqual)(void *,Aig_Obj_t *, Aig_Obj_t *) );

// extern void          Dch_ClassesStop( Dch_Cla_t * p );

// extern int           Dch_ClassesLitNum( Dch_Cla_t * p );

// extern Aig_Obj_t **  Dch_ClassesReadClass( Dch_Cla_t * p, Aig_Obj_t * pRepr, int * pnSize );

// extern void          Dch_ClassesPrint( Dch_Cla_t * p, int fVeryVerbose );

// extern void          Dch_ClassesPrepare( Dch_Cla_t * p, int fLatchCorr, int nMaxLevs );

// extern int           Dch_ClassesRefine( Dch_Cla_t * p );

// extern int           Dch_ClassesRefineOneClass( Dch_Cla_t * p, Aig_Obj_t * pRepr, int fRecursive );

// extern void          Dch_ClassesCollectOneClass( Dch_Cla_t * p, Aig_Obj_t * pRepr, Vec_Ptr_t * vRoots );

// extern void          Dch_ClassesCollectConst1Group( Dch_Cla_t * p, Aig_Obj_t * pObj, int nNodes, Vec_Ptr_t * vRoots );

// extern int           Dch_ClassesRefineConst1Group( Dch_Cla_t * p, Vec_Ptr_t * vRoots, int fRecursive );

/*=== dchCnf.c ===================================================*/
// extern void          Dch_CnfNodeAddToSolver( Dch_Man_t * p, Aig_Obj_t * pObj );

/*=== dchMan.c ===================================================*/
// extern Dch_Man_t *   Dch_ManCreate( Aig_Man_t * pAig, Dch_Pars_t * pPars );

// extern void          Dch_ManStop( Dch_Man_t * p );

// extern void          Dch_ManSatSolverRecycle( Dch_Man_t * p );

/*=== dchSat.c ===================================================*/
// extern int           Dch_NodesAreEquiv( Dch_Man_t * p, Aig_Obj_t * pObj1, Aig_Obj_t * pObj2 );

/*=== dchSim.c ===================================================*/
// extern Dch_Cla_t *   Dch_CreateCandEquivClasses( Aig_Man_t * pAig, int nWords, int fVerbose );

/*=== dchSimSat.c ===================================================*/
// extern void          Dch_ManResimulateCex( Dch_Man_t * p, Aig_Obj_t * pObj, Aig_Obj_t * pRepr );

// extern void          Dch_ManResimulateCex2( Dch_Man_t * p, Aig_Obj_t * pObj, Aig_Obj_t * pRepr );

/*=== dchSweep.c ===================================================*/
// extern void          Dch_ManSweep( Dch_Man_t * p );




//ABC_NAMESPACE_HEADER_END





////////////////////////////////////////////////////////////////////////
///                       END OF FILE                                ///
////////////////////////////////////////////////////////////////////////

