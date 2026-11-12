package pdr

/**CFile****************************************************************

  FileName    [pdrInt.h]

  SystemName  [ABC: Logic synthesis and verification system.]

  PackageName [Property driven reachability.]

  Synopsis    [Internal declarations.]

  Author      [Alan Mishchenko]
  
  Affiliation [UC Berkeley]

  Date        [Ver. 1.0. Started - November 20, 2010.]

  Revision    [$Id: pdrInt.h,v 1.00 2010/11/20 00:00:00 alanmi Exp $]

***********************************************************************/
 



////////////////////////////////////////////////////////////////////////
///                          INCLUDES                                ///
////////////////////////////////////////////////////////////////////////

//#include "aig/saig/saig.h"
//#include "misc/vec/vecWec.h"
//#include "sat/cnf/cnf.h"
//#include "pdr.h" 
//#include "misc/hash/hashInt.h"
//#include "aig/gia/giaAig.h"

//#define PDR_USE_SATOKO 1

#ifndef PDR_USE_SATOKO
    #include "sat/bsat/satSolver.h"
#else
    #include "sat/satoko/satoko.h"
    #define l_Undef  0
    #define l_True   1
    #define l_False -1
    #define sat_solver                       satoko_t
    #define sat_solver_new                   satoko_create
    #define sat_solver_delete                satoko_destroy
    #define zsat_solver_new_seed(s)          satoko_create()
    #define zsat_solver_restart_seed(s,a)    satoko_reset(s)
    #define sat_solver_nvars                 satoko_varnum
    #define sat_solver_setnvars              satoko_setnvars
    #define sat_solver_addclause(s,b,e)      satoko_add_clause(s,b,e-b)
    #define sat_solver_final                 satoko_final_conflict
    #define sat_solver_solve(s,b,e,c,x,y,z)  satoko_solve_assumptions_limit(s,b,e-b,(int)c)
    #define sat_solver_var_value             satoko_read_cex_varvalue
    #define sat_solver_set_runtime_limit     satoko_set_runtime_limit
    #define sat_solver_set_runid             satoko_set_runid           
    #define sat_solver_set_stop_func         satoko_set_stop_func          
    #define sat_solver_compress(s)             
#endif

//ABC_NAMESPACE_HEADER_START

////////////////////////////////////////////////////////////////////////
///                         PARAMETERS                               ///
////////////////////////////////////////////////////////////////////////
             
////////////////////////////////////////////////////////////////////////
///                         BASIC TYPES                              ///
////////////////////////////////////////////////////////////////////////

type Txs_Man_t =  Txs_Man_t_;
type Txs3_Man_t = Txs3_Man_t_;

   
type Pdr_Set_t struct {
    Sign        word;      // signature
    nRefs         int32;     // ref counter
    nTotal         int32;    // total literals
    nLits         int32;     // num flop literals
    Lits [0]int32;
}

   
type Pdr_Obl_t struct {
    iFrame         int32;    // time frame
    prio         int32;      // priority
    nRefs         int32;     // reference counter
    pState  *Pdr_Set_t;    // state cube
    pNext  *Pdr_Obl_t;     // next one
    pLink  *Pdr_Obl_t;     // queue link
}

   
type Pdr_Man_t struct {
    // input problem
    pPars  *Pdr_Par_t;     // parameters
    pAig  *Aig_Man_t;      // user's AIG
    pGia  *Gia_Man_t;      // user's AIG
    // static CNF representation
    pCnfMan  *Cnf_Man_t;   // CNF manager
    pCnf1  *Cnf_Dat_t;     // CNF for this AIG
    vVar2Reg  *Vec_Int_t;  // mapping of SAT var into registers
    // dynamic CNF representation
    pCnf2  *Cnf_Dat_t;     // CNF for this AIG
    pvId2Vars  *Vec_Int_t; // for each used ObjId, maps frame into SAT var
    vVar2Ids   Vec_Ptr_t;  // for each used frame, maps SAT var into ObjId
    vVLits  *Vec_Wec_t;    // CNF literals
    // data representation
    iOutCur         int32;   // current output
    nPrioShift         int32;// priority shift
    vCexes  *Vec_Ptr_t;    // counter-examples for each output
    vSolvers  *Vec_Ptr_t;  // SAT solvers
    vClauses  *Vec_Vec_t;  // clauses by timeframe
    pQueue  *Pdr_Obl_t;    // proof obligations
    pOrder        *int32;    // ordering of the lits
    vActVars  *Vec_Int_t;  // the counter of activation variables
    iUseFrame         int32; // the first used frame
    nAbsFlops         int32; // the number of flops used
    vAbsFlops  *Vec_Int_t; // flops currently used
    vMapFf2Ppi  *Vec_Int_t;
    vMapPpi2Ff  *Vec_Int_t;
    nCexes         int32;
    nCexesTotal         int32;
    // terminary simulation
    pTxs3  *Txs3_Man_t;      
    // internal use
    vPrio  *Vec_Int_t;     // priority flops
    vLits  *Vec_Int_t;     // array of literals
    vCiObjs  *Vec_Int_t;   // cone leaves
    vCoObjs  *Vec_Int_t;   // cone roots
    vCiVals  *Vec_Int_t;   // cone leaf values
    vCoVals  *Vec_Int_t;   // cone root values
    vNodes  *Vec_Int_t;    // cone nodes
    vUndo  *Vec_Int_t;     // cone undos
    vVisits  *Vec_Int_t;   // intermediate
    vCi2Rem  *Vec_Int_t;   // CIs to be removed
    vRes  *Vec_Int_t;      // final result
    pTime4Outs    *abctime;// timeout per output
    vInfCubes  *Vec_Ptr_t; // infinity clauses/cubes
    // statistics
    nBlocks         int32;   // the number of times blockState was called
    nObligs         int32;   // the number of proof obligations derived
    nCubes         int32;    // the number of cubes derived
    nCalls         int32;    // the number of SAT calls
    nCallsS         int32;   // the number of SAT calls (sat)
    nCallsU         int32;   // the number of SAT calls (unsat)
    nStarts         int32;   // the number of SAT solver restarts
    nFrames         int32;   // frames explored
    nCasesSS         int32;
    nCasesSU         int32;
    nCasesUS         int32;
    nCasesUU         int32;
    nQueCur         int32;
    nQueMax         int32;
    nQueLim         int32;
    nXsimRuns         int32;
    nXsimLits         int32;
    // runtime
    timeToStop     abctime;
    timeToStopOne     abctime;
    // time stats
    tSat     abctime;
    tSatSat     abctime;
    tSatUnsat     abctime;
    tGeneral     abctime;
    tPush     abctime;
    tTsim     abctime;
    tContain     abctime;
    tCnf     abctime;
    tAbs     abctime;
    tTotal     abctime;
}

////////////////////////////////////////////////////////////////////////
///                      MACRO DEFINITIONS                           ///
////////////////////////////////////////////////////////////////////////

func Pdr_ManSolver(p *Pdr_Man_t, k int32) *sat_solver { return (sat_solver *)Vec_PtrEntry(p.vSolvers, k); }

func Pdr_ManTimeLimit(p *Pdr_Man_t) abctime {
    if p.timeToStop == 0 {
        return p.timeToStopOne; }
    if p.timeToStopOne == 0 {
        return p.timeToStop; }
    if p.timeToStop < p.timeToStopOne {
        return p.timeToStop; }
    return p.timeToStopOne;
}

////////////////////////////////////////////////////////////////////////
///                    FUNCTION DECLARATIONS                         ///
////////////////////////////////////////////////////////////////////////

/*=== pdrCnf.c ==========================================================*/
// extern int             Pdr_ObjSatVar( Pdr_Man_t * p, int k, int Pol, Aig_Obj_t * pObj );

// extern int             Pdr_ObjRegNum( Pdr_Man_t * p, int k, int iSatVar );

// extern int             Pdr_ManFreeVar( Pdr_Man_t * p, int k );

// extern sat_solver *    Pdr_ManNewSolver( sat_solver * pSat, Pdr_Man_t * p, int k, int fInit );

/*=== pdrCore.c ==========================================================*/
// extern int             Pdr_ManCheckContainment( Pdr_Man_t * p, int k, Pdr_Set_t * pSet );

/*=== pdrInv.c ==========================================================*/
// extern Vec_Int_t *     Pdr_ManCountFlopsInv( Pdr_Man_t * p );

// extern void            Pdr_ManPrintProgress( Pdr_Man_t * p, int fClose, abctime Time );

// extern void            Pdr_ManPrintClauses( Pdr_Man_t * p, int kStart );

// extern void            Pdr_ManDumpClauses( Pdr_Man_t * p, char * pFileName, int fProved );

// extern Vec_Str_t *     Pdr_ManDumpString( Pdr_Man_t * p );

// extern void            Pdr_ManReportInvariant( Pdr_Man_t * p );

// extern void            Pdr_ManVerifyInvariant( Pdr_Man_t * p );

// extern Vec_Int_t *     Pdr_ManDeriveInfinityClauses( Pdr_Man_t * p, int fReduce );

/*=== pdrMan.c ==========================================================*/
// extern Pdr_Man_t *     Pdr_ManStart( Aig_Man_t * pAig, Pdr_Par_t * pPars, Vec_Int_t * vPrioInit );

// extern void            Pdr_ManStop( Pdr_Man_t * p );

// extern Abc_Cex_t *     Pdr_ManDeriveCex( Pdr_Man_t * p );

// extern Abc_Cex_t *     Pdr_ManDeriveCexAbs( Pdr_Man_t * p );

/*=== pdrSat.c ==========================================================*/
// extern sat_solver *    Pdr_ManCreateSolver( Pdr_Man_t * p, int k );

// extern sat_solver *    Pdr_ManFetchSolver( Pdr_Man_t * p, int k );

// extern void            Pdr_ManSetPropertyOutput( Pdr_Man_t * p, int k );

// extern Vec_Int_t *     Pdr_ManCubeToLits( Pdr_Man_t * p, int k, Pdr_Set_t * pCube, int fCompl, int fNext );

// extern Vec_Int_t *     Pdr_ManLitsToCube( Pdr_Man_t * p, int k, int * pArray, int nArray );

// extern void            Pdr_ManSolverAddClause( Pdr_Man_t * p, int k, Pdr_Set_t * pCube );

// extern void            Pdr_ManCollectValues( Pdr_Man_t * p, int k, Vec_Int_t * vObjIds, Vec_Int_t * vValues );

// extern int             Pdr_ManCheckCubeCs( Pdr_Man_t * p, int k, Pdr_Set_t * pCube );

// extern int             Pdr_ManCheckCube( Pdr_Man_t * p, int k, Pdr_Set_t * pCube, Pdr_Set_t ** ppPred, int nConfLimit, int fTryConf, int fUseLit );

/*=== pdrTsim.c ==========================================================*/
// extern Pdr_Set_t *     Pdr_ManTernarySim( Pdr_Man_t * p, int k, Pdr_Set_t * pCube );

/*=== pdrTsim2.c ==========================================================*/
// extern Txs_Man_t *     Txs_ManStart( Pdr_Man_t * pMan, Aig_Man_t * pAig, Vec_Int_t * vPrio );

// extern void            Txs_ManStop( Txs_Man_t * );

// extern Pdr_Set_t *     Txs_ManTernarySim( Txs_Man_t * p, int k, Pdr_Set_t * pCube );

/*=== pdrTsim3.c ==========================================================*/
// extern Txs3_Man_t *    Txs3_ManStart( Pdr_Man_t * pMan, Aig_Man_t * pAig, Vec_Int_t * vPrio );

// extern void            Txs3_ManStop( Txs3_Man_t * );

// extern Pdr_Set_t *     Txs3_ManTernarySim( Txs3_Man_t * p, int k, Pdr_Set_t * pCube );

/*=== pdrUtil.c ==========================================================*/
// extern Pdr_Set_t *     Pdr_SetAlloc( int nSize );

// extern Pdr_Set_t *     Pdr_SetCreate( Vec_Int_t * vLits, Vec_Int_t * vPiLits );

// extern Pdr_Set_t *     Pdr_SetCreateFrom( Pdr_Set_t * pSet, int iRemove );

// extern Pdr_Set_t *     Pdr_SetCreateSubset( Pdr_Set_t * pSet, int * pLits, int nLits );

// extern Pdr_Set_t *     Pdr_SetDup( Pdr_Set_t * pSet );

// extern Pdr_Set_t *     Pdr_SetRef( Pdr_Set_t * p );

// extern void            Pdr_SetDeref( Pdr_Set_t * p );

// extern Pdr_Set_t *     ZPdr_SetIntersection( Pdr_Set_t * p1, Pdr_Set_t * p2, Hash_Int_t * keep );

// extern int             Pdr_SetContains( Pdr_Set_t * pOld, Pdr_Set_t * pNew );

// extern int             Pdr_SetContainsSimple( Pdr_Set_t * pOld, Pdr_Set_t * pNew );

// extern int             Pdr_SetIsInit( Pdr_Set_t * p, int iRemove );

// extern int             ZPdr_SetIsInit( Pdr_Set_t * p );

// extern void            Pdr_SetPrint( FILE * pFile, Pdr_Set_t * p, int nRegs, Vec_Int_t * vFlopCounts );

// extern void            ZPdr_SetPrint( Pdr_Set_t * p );

// extern void            Pdr_SetPrintStr( Vec_Str_t * vStr, Pdr_Set_t * p, int nRegs, Vec_Int_t * vFlopCounts );

// extern int             Pdr_SetCompare( Pdr_Set_t ** pp1, Pdr_Set_t ** pp2 );

// extern Pdr_Obl_t *     Pdr_OblStart( int k, int prio, Pdr_Set_t * pState, Pdr_Obl_t * pNext );

// extern Pdr_Obl_t *     Pdr_OblRef( Pdr_Obl_t * p );

// extern void            Pdr_OblDeref( Pdr_Obl_t * p );

// extern int             Pdr_QueueIsEmpty( Pdr_Man_t * p );

// extern Pdr_Obl_t *     Pdr_QueueHead( Pdr_Man_t * p );

// extern Pdr_Obl_t *     Pdr_QueuePop( Pdr_Man_t * p );

// extern void            Pdr_QueueClean( Pdr_Man_t * p );

// extern void            Pdr_QueuePush( Pdr_Man_t * p, Pdr_Obl_t * pObl );

// extern void            Pdr_QueuePrint( Pdr_Man_t * p );

// extern void            Pdr_QueueStop( Pdr_Man_t * p );


//ABC_NAMESPACE_HEADER_END




////////////////////////////////////////////////////////////////////////
///                       END OF FILE                                ///
////////////////////////////////////////////////////////////////////////

