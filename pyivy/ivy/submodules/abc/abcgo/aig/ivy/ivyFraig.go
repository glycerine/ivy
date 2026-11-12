package ivy

import (
	"fmt"
	"math"
	"os"
	"time"
)

/**CFile****************************************************************

  FileName    [ivyFraig.c]

  SystemName  [ABC: Logic synthesis and verification system.]

  PackageName [And-Inverter Graph package.]

  Synopsis    [Functional reduction of AIGs]

  Author      [Alan Mishchenko]
  
  Affiliation [UC Berkeley]

  Date        [Ver. 1.0. Started - May 11, 2006.]

  Revision    [$Id: ivyFraig.c,v 1.00 2006/05/11 00:00:00 alanmi Exp $]

***********************************************************************/

//#include <math.h>

//#include "sat/bsat/satSolver.h"
//#include "misc/extra/extra.h"
//#include "ivy.h"

//ABC_NAMESPACE_IMPL_START

////////////////////////////////////////////////////////////////////////
///                        DECLARATIONS                              ///
////////////////////////////////////////////////////////////////////////

type Ivy_FraigMan_t =        Ivy_FraigMan_t_;
type Ivy_FraigSim_t =        Ivy_FraigSim_t_;
         

type Ivy_FraigList_t struct {
    pHead          *Ivy_Obj_t;
    pTail          *Ivy_Obj_t;
    nItems                 int32;
}

type Ivy_FraigSim_t_ struct {
    Type                 int32;
    pNext     *Ivy_FraigSim_t;
    pFanin0     *Ivy_FraigSim_t;
    pFanin1     *Ivy_FraigSim_t;
    pData [0]uint32;
}

type Ivy_FraigMan_t_ struct {
    // general info 
    pParams  *Ivy_FraigParams_t;        // various parameters
    // temporary backtrack limits because "ABC_INT64_T" cannot be defined in Ivy_FraigParams_t ...
    nBTLimitGlobal              ABC_INT64_T; // global limit on the number of backtracks
    nInsLimitGlobal              ABC_INT64_T;// global limit on the number of clause inspects
    // AIG manager
    pManAig          *Ivy_Man_t;        // the starting AIG manager
    pManFraig          *Ivy_Man_t;      // the final AIG manager
    // simulation information
    nSimWords                 int32;      // the number of words
    pSimWords               string;      // the simulation info
    pSimStart     *Ivy_FraigSim_t;      // the list of simulation info for internal nodes
    // counter example storage
    nPatWords                 int32;      // the number of words in the counter example
    pPatWords           *uint32;      // the counter example
    pPatScores                *int32;     // the scores of each pattern
    // equivalence classes 
    lClasses     Ivy_FraigList_t;       // equivalence classes
    lCand     Ivy_FraigList_t;          // candidatates
    nPairs                 int32;         // the number of pairs of nodes
    // equivalence checking
    pSat         *sat_solver;           // SAT solver
    nSatVars                 int32;       // the number of variables currently used
    vPiVars          *Vec_Ptr_t;        // the PIs of the cone used 
    // other 
    pProgress        *ProgressBar;
    // statistics
    nSimRounds                 int32;
    nNodesMiter                 int32;
    nClassesZero                 int32;
    nClassesBeg                 int32;
    nClassesEnd                 int32;
    nPairsBeg                 int32;
    nPairsEnd                 int32;
    nSatCalls                 int32;
    nSatCallsSat                 int32;
    nSatCallsUnsat                 int32;
    nSatProof                 int32;
    nSatFails                 int32;
    nSatFailsReal                 int32;
    // runtime
    timeSim             abctime;
    timeTrav             abctime;
    timeSat             abctime;
    timeSatUnsat             abctime;
    timeSatSat             abctime;
    timeSatFail             abctime;
    timeRef             abctime;
    timeTotal             abctime;
    time1             abctime;
    time2             abctime;
}

        
type Prove_Params_t struct {
    // general parameters
    fUseFraiging     int32;          // enables fraiging
    fUseRewriting     int32;         // enables rewriting
    fUseBdds     int32;              // enables BDD construction when other methods fail
    fVerbose     int32;              // prints verbose stats
    // iterations
    nItersMax     int32;             // the number of iterations
    // mitering 
    nMiteringLimitStart     int32;   // starting mitering limit
    nMiteringLimitMulti   float32;   // multiplicative coefficient to increase the limit in each iteration
    // rewriting 
    nRewritingLimitStart     int32;  // the number of rewriting iterations
    nRewritingLimitMulti   float32;  // multiplicative coefficient to increase the limit in each iteration
    // fraiging 
    nFraigingLimitStart     int32;   // starting backtrack(conflict) limit
    nFraigingLimitMulti   float32;   // multiplicative coefficient to increase the limit in each iteration
    // last-gasp BDD construction
    nBddSizeLimit     int32;         // the number of BDD nodes when construction is aborted
    fBddReorder     int32;           // enables dynamic BDD variable reordering
    // last-gasp mitering
    nMiteringLimitLast     int32;    // final mitering limit
    // global SAT solver limits
    nTotalBacktrackLimit  ABC_INT64_T;  // global limit on the number of backtracks
    nTotalInspectLimit  ABC_INT64_T;    // global limit on the number of clause inspects
    // global resources applied
    nTotalBacktracksMade  ABC_INT64_T;  // the total number of backtracks made
    nTotalInspectsMade  ABC_INT64_T;    // the total number of inspects made
}

func Ivy_ObjSim(pObj *Ivy_Obj_t) *Ivy_FraigSim_t { return (Ivy_FraigSim_t *)pObj.pFanout;  }
func Ivy_ObjClassNodeLast(pObj *Ivy_Obj_t) *Ivy_Obj_t { return pObj.pNextFan0;  }
func Ivy_ObjClassNodeRepr(pObj *Ivy_Obj_t) *Ivy_Obj_t { return pObj.pNextFan0;  }
func Ivy_ObjClassNodeNext(pObj *Ivy_Obj_t) *Ivy_Obj_t { return pObj.pNextFan1;  }
func Ivy_ObjNodeHashNext(pObj *Ivy_Obj_t) *Ivy_Obj_t { return pObj.pPrevFan0;  }
func Ivy_ObjEquivListNext(pObj *Ivy_Obj_t) *Ivy_Obj_t { return pObj.pPrevFan0;  }
func Ivy_ObjEquivListPrev(pObj *Ivy_Obj_t) *Ivy_Obj_t { return pObj.pPrevFan1;  }
func Ivy_ObjFraig(pObj *Ivy_Obj_t) *Ivy_Obj_t { return pObj.pEquiv;     }
func Ivy_ObjSatNum(pObj *Ivy_Obj_t) int32 { return int32((ABC_PTRUINT_T))pObj.pNextFan0;         }
func Ivy_ObjFaninVec(pObj *Ivy_Obj_t) *Vec_Ptr_t { return (Vec_Ptr_t *)pObj.pNextFan1; }

func Ivy_ObjSetSim(pObj *Ivy_Obj_t, pSim *Ivy_FraigSim_t) { pObj.pFanout = (Ivy_Obj_t *)pSim; }
func Ivy_ObjSetClassNodeLast(pObj *Ivy_Obj_t, pLast *Ivy_Obj_t) { pObj.pNextFan0 = pLast; }
func Ivy_ObjSetClassNodeRepr(pObj *Ivy_Obj_t, pRepr *Ivy_Obj_t) { pObj.pNextFan0 = pRepr; }
func Ivy_ObjSetClassNodeNext(pObj *Ivy_Obj_t, pNext *Ivy_Obj_t) { pObj.pNextFan1 = pNext; }
func Ivy_ObjSetNodeHashNext(pObj *Ivy_Obj_t, pNext *Ivy_Obj_t) { pObj.pPrevFan0 = pNext; }
func Ivy_ObjSetEquivListNext(pObj *Ivy_Obj_t, pNext *Ivy_Obj_t) { pObj.pPrevFan0 = pNext; }
func Ivy_ObjSetEquivListPrev(pObj *Ivy_Obj_t, pPrev *Ivy_Obj_t) { pObj.pPrevFan1 = pPrev; }
func Ivy_ObjSetFraig(pObj *Ivy_Obj_t, pNode *Ivy_Obj_t) { pObj.pEquiv    = pNode; }
func Ivy_ObjSetSatNum(pObj *Ivy_Obj_t, Num int32) { pObj.pNextFan0 = (Ivy_Obj_t *)(ABC_PTRUINT_T)Num;     }
func Ivy_ObjSetFaninVec(pObj *Ivy_Obj_t, vFanins *Vec_Ptr_t) { pObj.pNextFan1 = (Ivy_Obj_t *)vFanins; }

func Ivy_ObjRandomSim() uint32 { return (rand() << 24) ^ (rand() << 12) ^ rand(); }

// iterate through equivalence classes
// Original macro:
// #define Ivy_FraigForEachEquivClass( pList, pEnt )                 \
//     for ( pEnt = pList;                                           \
//           pEnt;                                                   \
//           pEnt = Ivy_ObjEquivListNext(pEnt) )

// Original macro:
// #define Ivy_FraigForEachEquivClassSafe( pList, pEnt, pEnt2 )      \
//     for ( pEnt = pList,                                           \
//           pEnt2 = pEnt? Ivy_ObjEquivListNext(pEnt): NULL;         \
//           pEnt;                                                   \
//           pEnt = pEnt2,                                           \
//           pEnt2 = pEnt? Ivy_ObjEquivListNext(pEnt): NULL )

// iterate through nodes in one class
// Original macro:
// #define Ivy_FraigForEachClassNode( pClass, pEnt )                 \
//     for ( pEnt = pClass;                                          \
//           pEnt;                                                   \
//           pEnt = Ivy_ObjClassNodeNext(pEnt) )

// iterate through nodes in the hash table
// Original macro:
// #define Ivy_FraigForEachBinNode( pBin, pEnt )                     \
//     for ( pEnt = pBin;                                            \
//           pEnt;                                                   \
//           pEnt = Ivy_ObjNodeHashNext(pEnt) )


// static Ivy_FraigMan_t * Ivy_FraigStart( Ivy_Man_t * pManAig, Ivy_FraigParams_t * pParams );

// static Ivy_FraigMan_t * Ivy_FraigStartSimple( Ivy_Man_t * pManAig, Ivy_FraigParams_t * pParams );

// static Ivy_Man_t *   Ivy_FraigPerform_int( Ivy_Man_t * pManAig, Ivy_FraigParams_t * pParams, ABC_INT64_T nBTLimitGlobal, ABC_INT64_T nInsLimitGlobal, ABC_INT64_T * pnSatConfs, ABC_INT64_T * pnSatInspects );

// static void          Ivy_FraigPrint( Ivy_FraigMan_t * p );

// static void          Ivy_FraigStop( Ivy_FraigMan_t * p );

// static void          Ivy_FraigSimulate( Ivy_FraigMan_t * p );

// static void          Ivy_FraigSweep( Ivy_FraigMan_t * p );

// static Ivy_Obj_t *   Ivy_FraigAnd( Ivy_FraigMan_t * p, Ivy_Obj_t * pObjOld );

// static int           Ivy_FraigNodesAreEquiv( Ivy_FraigMan_t * p, Ivy_Obj_t * pObj0, Ivy_Obj_t * pObj1 );

// static int           Ivy_FraigNodeIsConst( Ivy_FraigMan_t * p, Ivy_Obj_t * pObj );

// static void          Ivy_FraigNodeAddToSolver( Ivy_FraigMan_t * p, Ivy_Obj_t * pObj0, Ivy_Obj_t * pObj1 );

// static int           Ivy_FraigSetActivityFactors( Ivy_FraigMan_t * p, Ivy_Obj_t * pOld, Ivy_Obj_t * pNew );

// static void          Ivy_FraigAddToPatScores( Ivy_FraigMan_t * p, Ivy_Obj_t * pClass, Ivy_Obj_t * pClassNew );

// static int           Ivy_FraigMiterStatus( Ivy_Man_t * pMan );

// static void          Ivy_FraigMiterProve( Ivy_FraigMan_t * p );

// static void          Ivy_FraigMiterPrint( Ivy_Man_t * pNtk, char * pString, abctime clk, int fVerbose );

// static int *         Ivy_FraigCreateModel( Ivy_FraigMan_t * p );


// static int Ivy_FraigNodesAreEquivBdd( Ivy_Obj_t * pObj1, Ivy_Obj_t * pObj2 );

// static int Ivy_FraigCheckCone( Ivy_FraigMan_t * pGlo, Ivy_Man_t * p, Ivy_Obj_t * pObj1, Ivy_Obj_t * pObj2, int nConfLimit );


var s_nBTLimitGlobal ABC_INT64_T;
var s_nInsLimitGlobal ABC_INT64_T;

////////////////////////////////////////////////////////////////////////
///                     FUNCTION DEFINITIONS                         ///
////////////////////////////////////////////////////////////////////////

/**Function*************************************************************

  Synopsis    [Sets the default solving parameters.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Ivy_FraigParamsDefault(pParams *Ivy_FraigParams_t) {
    memset( pParams, 0, sizeof(Ivy_FraigParams_t) );
    pParams.nSimWords        =      32;  // the number of words in the simulation info
    pParams.dSimSatur        =   0.005;  // the ratio of refined classes when saturation is reached
    pParams.fPatScores       =       0;  // enables simulation pattern scoring
    pParams.MaxScore         =      25;  // max score after which resimulation is used
    pParams.fDoSparse        =       1;  // skips sparse functions
//    pParams->dActConeRatio    =    0.05;  // the ratio of cone to be bumped
//    pParams->dActConeBumpMax  =     5.0;  // the largest bump of activity
    pParams.dActConeRatio    =     0.3;  // the ratio of cone to be bumped
    pParams.dActConeBumpMax  =    10.0;  // the largest bump of activity

    pParams.nBTLimitNode     =     100;  // conflict limit at a node
    pParams.nBTLimitMiter    =  500000;  // conflict limit at an output
//    pParams->nBTLimitGlobal   =       0;  // conflict limit global
//    pParams->nInsLimitGlobal  =       0;  // inspection limit global
}

/**Function*************************************************************

  Synopsis    [Performs combinational equivalence checking for the miter.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Ivy_FraigProve(ppManAig **Ivy_Man_t, pPars any) int32 {
    var pParams *Prove_Params_t  = (Prove_Params_t *)pPars;
    var Params Ivy_FraigParams_t;
    var pIvyParams *Ivy_FraigParams_t = &Params; 
    var pManAig, pManTemp *Ivy_Man_t;
    var RetValue, nIter int32;
    var clk abctime;//, Counter;
    var nSatConfs, nSatInspects ABC_INT64_T;

    // start the network and parameters
    pManAig = *ppManAig;
    Ivy_FraigParamsDefault( pIvyParams );
    pIvyParams.fVerbose = pParams.fVerbose;
    pIvyParams.fProve = 1;

    if pParams.fVerbose {
        fmt.Printf("RESOURCE LIMITS: Iterations = %d. Rewriting = %s. Fraiging = %s.\n", pParams.nItersMax, pParams.fUseRewriting? "yes":"no", pParams.fUseFraiging? "yes":"no");
        fmt.Printf("Miter = %d (%3.1f).  Rwr = %d (%3.1f).  Fraig = %d (%3.1f).  Last = %d.\n", pParams.nMiteringLimitStart, pParams.nMiteringLimitMulti, pParams.nRewritingLimitStart, pParams.nRewritingLimitMulti, pParams.nFraigingLimitStart, pParams.nFraigingLimitMulti, pParams.nMiteringLimitLast);
    }

    // if SAT only, solve without iteration
    if !pParams.fUseRewriting && !pParams.fUseFraiging {
        clk = Abc_Clock();
        pIvyParams.nBTLimitMiter = pParams.nMiteringLimitLast / Ivy_ManPoNum(pManAig);
        pManAig = Ivy_FraigMiter( pManTemp = pManAig, pIvyParams );  Ivy_ManStop( pManTemp );
        RetValue = Ivy_FraigMiterStatus( pManAig );
        Ivy_FraigMiterPrint( pManAig, "SAT solving", clk, pParams.fVerbose );
        *ppManAig = pManAig;
        return RetValue;
    }

    if Ivy_ManNodeNum(pManAig) < 500 {
        // run the first mitering
        clk = Abc_Clock();
        pIvyParams.nBTLimitMiter = pParams.nMiteringLimitStart / Ivy_ManPoNum(pManAig);
        pManAig = Ivy_FraigMiter( pManTemp = pManAig, pIvyParams );  Ivy_ManStop( pManTemp );
        RetValue = Ivy_FraigMiterStatus( pManAig );
        Ivy_FraigMiterPrint( pManAig, "SAT solving", clk, pParams.fVerbose );
        if RetValue >= 0 {
            *ppManAig = pManAig;
            return RetValue;
        }
    }

    // check the current resource limits
    RetValue = -1;
    for  nIter = 0; nIter < pParams.nItersMax; nIter++  {
        if pParams.fVerbose {
            fmt.Printf("ITERATION %2d : Confs = %6d. FraigBTL = %3d. \n", nIter+1, int32((pParams.nMiteringLimitStart * math.Pow(float64(pParams.nMiteringLimitMulti), float64(nIter)))), int32((pParams.nFraigingLimitStart * math.Pow(float64(pParams.nFraigingLimitMulti), float64(nIter)))));
            fflush( os.Stdout );
        }

        // try rewriting
        if pParams.fUseRewriting { // bug in Ivy_NodeFindCutsAll() when leaves are identical!
/*
            clk = Abc_Clock();
            Counter = (int)(pParams->nRewritingLimitStart * pow(pParams->nRewritingLimitMulti,nIter));
            pManAig = Ivy_ManRwsat( pManAig, 0 );  
            RetValue = Ivy_FraigMiterStatus( pManAig );
            Ivy_FraigMiterPrint( pManAig, "Rewriting  ", clk, pParams->fVerbose );
*/
        }
        if RetValue >= 0 {
            break; }

        // catch the situation when ref pattern detects the bug
        RetValue = Ivy_FraigMiterStatus( pManAig );
        if RetValue >= 0 {
            break; }

        // try fraiging followed by mitering
        if pParams.fUseFraiging {
            clk = Abc_Clock();
            pIvyParams.nBTLimitNode  = int32((pParams.nFraigingLimitStart * math.Pow(float64(pParams.nFraigingLimitMulti), float64(nIter))));
            pIvyParams.nBTLimitMiter = 1 + int32((pParams.nMiteringLimitStart * math.Pow(float64(pParams.nMiteringLimitMulti), float64(nIter)))) / Ivy_ManPoNum(pManAig);
            pManAig = Ivy_FraigPerform_int( pManTemp = pManAig, pIvyParams, pParams.nTotalBacktrackLimit, pParams.nTotalInspectLimit, &nSatConfs, &nSatInspects );  Ivy_ManStop( pManTemp );
            RetValue = Ivy_FraigMiterStatus( pManAig );
            Ivy_FraigMiterPrint( pManAig, "Fraiging   ", clk, pParams.fVerbose );
        }
        if RetValue >= 0 {
            break; }

        // add to the number of backtracks and inspects
        pParams.nTotalBacktracksMade += nSatConfs;
        pParams.nTotalInspectsMade   += nSatInspects;
        // check if global resource limit is reached
        if (pParams.nTotalBacktrackLimit && pParams.nTotalBacktracksMade >= pParams.nTotalBacktrackLimit) ||
             (pParams.nTotalInspectLimit   && pParams.nTotalInspectsMade   >= pParams.nTotalInspectLimit) {
            fmt.Printf("Reached global limit on conflicts/inspects. Quitting.\n");
            *ppManAig = pManAig;
            return -1;
        }
    }    
/*
    if ( RetValue < 0 )
    {
        if ( pParams->fVerbose )
        {
            printf( "Attempting SAT with conflict limit %d ...\n", pParams->nMiteringLimitLast );
            fflush( stdout );
        }
        clk = Abc_Clock();
        pIvyParams->nBTLimitMiter = pParams->nMiteringLimitLast / Ivy_ManPoNum(pManAig);
        if ( pParams->nTotalBacktrackLimit )
            s_nBTLimitGlobal  = pParams->nTotalBacktrackLimit - pParams->nTotalBacktracksMade;
        if ( pParams->nTotalInspectLimit )
            s_nInsLimitGlobal = pParams->nTotalInspectLimit -   pParams->nTotalInspectsMade;        
        pManAig = Ivy_FraigMiter( pManTemp = pManAig, pIvyParams );  Ivy_ManStop( pManTemp );
        s_nBTLimitGlobal  = 0;
        s_nInsLimitGlobal = 0;        
        RetValue = Ivy_FraigMiterStatus( pManAig );
        Ivy_FraigMiterPrint( pManAig, "SAT solving", clk, pParams->fVerbose );
        // make sure that the sover never returns "undecided" when infinite resource limits are set
        if( RetValue == -1 && pParams->nTotalInspectLimit == 0 &&
            pParams->nTotalBacktrackLimit == 0 )
        {
            extern void Prove_ParamsPrint( Prove_Params_t * pParams );
            Prove_ParamsPrint( pParams );
            printf("ERROR: ABC has returned \"undecided\" in spite of no limits...\n");
            exit(1);
        }
    }
*/
    // assign the model if it was proved by rewriting (const 1 miter)
    if RetValue == 0 && pManAig.pData == nil {
        // inlining of macro ABC_ALLOC
        pManAig.pData = ((int *) malloc(sizeof(int) * (size_t)(Ivy_ManPiNum(pManAig))));
        memset( pManAig.pData, 0, sizeof(int) * Ivy_ManPiNum(pManAig) );
    }
    *ppManAig = pManAig;
    return RetValue;
}

/**Function*************************************************************

  Synopsis    [Performs fraiging of the AIG.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Ivy_FraigPerform_int(pManAig *Ivy_Man_t, pParams *Ivy_FraigParams_t, nBTLimitGlobal ABC_INT64_T, nInsLimitGlobal ABC_INT64_T, pnSatConfs *ABC_INT64_T, pnSatInspects *ABC_INT64_T) *Ivy_Man_t { 
    var p *Ivy_FraigMan_t;
    var pManAigNew *Ivy_Man_t;
    var clk abctime;
    if Ivy_ManNodeNum(pManAig) == 0 {
        return Ivy_ManDup(pManAig); }
clk = Abc_Clock();
    assert(Ivy_ManLatchNum(pManAig) == 0, "Ivy_ManLatchNum(pManAig) == 0");
    p = Ivy_FraigStart( pManAig, pParams );
    // set global limits
    p.nBTLimitGlobal  = nBTLimitGlobal;
    p.nInsLimitGlobal = nInsLimitGlobal; 
    
    Ivy_FraigSimulate( p );
    Ivy_FraigSweep( p );
    pManAigNew = p.pManFraig;
p.timeTotal = Abc_Clock() - clk;
    if pnSatConfs != nil {
        *pnSatConfs = p.pSat? p.pSat.stats.conflicts : 0; }
    if pnSatInspects != nil {
        *pnSatInspects = p.pSat? p.pSat.stats.inspects : 0; }
    Ivy_FraigStop( p );
    return pManAigNew;
}

/**Function*************************************************************

  Synopsis    [Performs fraiging of the AIG.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Ivy_FraigPerform(pManAig *Ivy_Man_t, pParams *Ivy_FraigParams_t) *Ivy_Man_t {
    var p *Ivy_FraigMan_t;
    var pManAigNew *Ivy_Man_t;
    var clk abctime;
    if Ivy_ManNodeNum(pManAig) == 0 {
        return Ivy_ManDup(pManAig); }
clk = Abc_Clock();
    assert(Ivy_ManLatchNum(pManAig) == 0, "Ivy_ManLatchNum(pManAig) == 0");
    p = Ivy_FraigStart( pManAig, pParams );
    Ivy_FraigSimulate( p );
    Ivy_FraigSweep( p );
    pManAigNew = p.pManFraig;
p.timeTotal = Abc_Clock() - clk;
    Ivy_FraigStop( p );
    return pManAigNew;
}

/**Function*************************************************************

  Synopsis    [Applies brute-force SAT to the miter.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Ivy_FraigMiter(pManAig *Ivy_Man_t, pParams *Ivy_FraigParams_t) *Ivy_Man_t {
    var p *Ivy_FraigMan_t;
    var pManAigNew *Ivy_Man_t;
    var pObj *Ivy_Obj_t;
    var i int32;
    var clk abctime;
clk = Abc_Clock();
    assert(Ivy_ManLatchNum(pManAig) == 0, "Ivy_ManLatchNum(pManAig) == 0");
    p = Ivy_FraigStartSimple( pManAig, pParams );
    // set global limits
    p.nBTLimitGlobal  = s_nBTLimitGlobal;
    p.nInsLimitGlobal = s_nInsLimitGlobal; 
    // duplicate internal nodes
    // inlining of macro Ivy_ManForEachNode
    // inlining of macro Ivy_ManForEachObj
    // inlining of macro Vec_PtrForEachEntry
    for  i = 0; (i < Vec_PtrSize(p.pManAig.vObjs)) && (((pObj) = (Ivy_Obj_t *)Vec_PtrEntry(p.pManAig.vObjs, i)), 1); i++  { if (pObj) == nil {} else if !Ivy_ObjIsNode(pObj) {} else {
        pObj.pEquiv = Ivy_And( p.pManFraig, Ivy_ObjChild0Equiv(pObj), Ivy_ObjChild1Equiv(pObj) ); }
    }
    // try to prove each output of the miter
    Ivy_FraigMiterProve( p );
    // add the POs
    // inlining of macro Ivy_ManForEachPo
    // inlining of macro Vec_PtrForEachEntry
    for  i = 0; (i < Vec_PtrSize(p.pManAig.vPos)) && (((pObj) = (Ivy_Obj_t *)Vec_PtrEntry(p.pManAig.vPos, i)), 1); i++  {
        Ivy_ObjCreatePo( p.pManFraig, Ivy_ObjChild0Equiv(pObj) );
    }
    // clean the new manager
    // inlining of macro Ivy_ManForEachObj
    // inlining of macro Vec_PtrForEachEntry
    for  i = 0; (i < Vec_PtrSize(p.pManFraig.vObjs)) && (((pObj) = (Ivy_Obj_t *)Vec_PtrEntry(p.pManFraig.vObjs, i)), 1); i++  { if (pObj) == nil {} else {
        if Ivy_ObjFaninVec(pObj) != nil {
            Vec_PtrFree( Ivy_ObjFaninVec(pObj) ); }
        pObj.pNextFan0 = pObj.pNextFan1 = nil;
    }
    }
    // remove dangling nodes 
    Ivy_ManCleanup( p.pManFraig );
    pManAigNew = p.pManFraig;
p.timeTotal = Abc_Clock() - clk;

//printf( "Final nodes = %6d. ", Ivy_ManNodeNum(pManAigNew) );
//ABC_PRT( "Time", p->timeTotal );
    Ivy_FraigStop( p );
    return pManAigNew;
}

/**Function*************************************************************

  Synopsis    [Starts the fraiging manager without simulation info.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Ivy_FraigStartSimple(pManAig *Ivy_Man_t, pParams *Ivy_FraigParams_t) *Ivy_FraigMan_t {
    var p *Ivy_FraigMan_t;
    // allocat the fraiging manager
    // inlining of macro ABC_ALLOC
    p = ((Ivy_FraigMan_t *) malloc(sizeof(Ivy_FraigMan_t) * (size_t)(1)));
    memset( p, 0, sizeof(Ivy_FraigMan_t) );
    p.pParams   = pParams;
    p.pManAig   = pManAig;
    p.pManFraig = Ivy_ManStartFrom( pManAig );
    p.vPiVars   = Vec_PtrAlloc( 100 );
    return p;
}

/**Function*************************************************************

  Synopsis    [Starts the fraiging manager.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Ivy_FraigStart(pManAig *Ivy_Man_t, pParams *Ivy_FraigParams_t) *Ivy_FraigMan_t {
    var p *Ivy_FraigMan_t;
    var pSims *Ivy_FraigSim_t;
    var pObj *Ivy_Obj_t;
    var i, k, EntrySize int32;
    // clean the fanout representation
    // inlining of macro Ivy_ManForEachObj
    // inlining of macro Vec_PtrForEachEntry
    for  i = 0; (i < Vec_PtrSize(pManAig.vObjs)) && (((pObj) = (Ivy_Obj_t *)Vec_PtrEntry(pManAig.vObjs, i)), 1); i++  { if (pObj) == nil {} else {
//        pObj->pEquiv = pObj->pFanout = pObj->pNextFan0 = pObj->pNextFan1 = pObj->pPrevFan0 = pObj->pPrevFan1 = NULL;
        assert(!pObj.pEquiv && !pObj.pFanout, "!pObj.pEquiv && !pObj.pFanout"); }
    }
    // allocat the fraiging manager
    // inlining of macro ABC_ALLOC
    p = ((Ivy_FraigMan_t *) malloc(sizeof(Ivy_FraigMan_t) * (size_t)(1)));
    memset( p, 0, sizeof(Ivy_FraigMan_t) );
    p.pParams   = pParams;
    p.pManAig   = pManAig;
    p.pManFraig = Ivy_ManStartFrom( pManAig );
    // allocate simulation info
    p.nSimWords    = pParams.nSimWords;
//    p->pSimWords    = ABC_ALLOC( unsigned, Ivy_ManObjNum(pManAig) * p->nSimWords ); 
    EntrySize    = sizeof(Ivy_FraigSim_t) + sizeof(unsigned) * p.nSimWords;
    // inlining of macro ABC_ALLOC
    p.pSimWords = (char *)((char *) malloc(sizeof(char) * (size_t)(Ivy_ManObjNum(pManAig) * EntrySize))); 
    memset( p.pSimWords, 0, (size_t)EntrySize );
    k = 0;
    // inlining of macro Ivy_ManForEachObj
    // inlining of macro Vec_PtrForEachEntry
    for  i = 0; (i < Vec_PtrSize(pManAig.vObjs)) && (((pObj) = (Ivy_Obj_t *)Vec_PtrEntry(pManAig.vObjs, i)), 1); i++  { if (pObj) == nil {} else {
        pSims = (Ivy_FraigSim_t *)(p.pSimWords + EntrySize * k++);
        pSims.pNext = nil;
        if Ivy_ObjIsNode(pObj) {
            if p.pSimStart == nil {
                p.pSimStart = pSims;
            } else {
                ((Ivy_FraigSim_t *)(p.pSimWords + EntrySize * (k-2))).pNext = pSims; }
            pSims.pFanin0 = Ivy_ObjSim( Ivy_ObjFanin0(pObj) );
            pSims.pFanin1 = Ivy_ObjSim( Ivy_ObjFanin1(pObj) );
            pSims.Type = (Ivy_ObjFaninPhase(Ivy_ObjChild0(pObj)) << 2) | (Ivy_ObjFaninPhase(Ivy_ObjChild1(pObj)) << 1) | pObj.fPhase;
        
        } else {
            pSims.pFanin0 = nil;
            pSims.pFanin1 = nil;
            pSims.Type = 0;
        }
        Ivy_ObjSetSim( pObj, pSims );
    }
    }
    assert(k == Ivy_ManObjNum(pManAig), "k == Ivy_ManObjNum(pManAig)");
    // allocate storage for sim pattern
    p.nPatWords  = Ivy_BitWordNum( Ivy_ManPiNum(pManAig) );
    // inlining of macro ABC_ALLOC
    p.pPatWords  = ((unsigned *) malloc(sizeof(unsigned) * (size_t)(p.nPatWords))); 
    // inlining of macro ABC_ALLOC
    p.pPatScores = ((int *) malloc(sizeof(int) * (size_t)(32 * p.nSimWords))); 
    p.vPiVars    = Vec_PtrAlloc( 100 );
    // set random number generator
    srand( 0xABCABC );
    return p;
}

/**Function*************************************************************

  Synopsis    [Stops the fraiging manager.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Ivy_FraigStop(p *Ivy_FraigMan_t) {
    if p.pParams.fVerbose {
        Ivy_FraigPrint( p ); }
    if p.vPiVars { Vec_PtrFree( p.vPiVars ); }
    if p.pSat { sat_solver_delete( p.pSat ); }
    // inlining of macro ABC_FREE
    ((p.pPatScores) ? (free((char *) (p.pPatScores)), (p.pPatScores) = 0) : 0);
    // inlining of macro ABC_FREE
    ((p.pPatWords) ? (free((char *) (p.pPatWords)), (p.pPatWords) = 0) : 0);
    // inlining of macro ABC_FREE
    ((p.pSimWords) ? (free((char *) (p.pSimWords)), (p.pSimWords) = 0) : 0);
    // inlining of macro ABC_FREE
    ((p) ? (free((char *) (p)), (p) = 0) : 0);
}

/**Function*************************************************************

  Synopsis    [Prints stats for the fraiging manager.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Ivy_FraigPrint(p *Ivy_FraigMan_t) {
    var nMemory float64;
    nMemory = float64(Ivy_ManObjNum(p.pManAig))*p.nSimWords*sizeof(unsigned)/(1<<20);
    fmt.Printf("SimWords = %d. Rounds = %d. Mem = %0.2f MB.  ", p.nSimWords, p.nSimRounds, nMemory);
    fmt.Printf("Classes: Beg = %d. End = %d.\n", p.nClassesBeg, p.nClassesEnd);
//    printf( "Limits: BTNode = %d. BTMiter = %d.\n", p->pParams->nBTLimitNode, p->pParams->nBTLimitMiter );
    fmt.Printf("Proof = %d. Counter-example = %d. Fail = %d. FailReal = %d. Zero = %d.\n", p.nSatProof, p.nSatCallsSat, p.nSatFails, p.nSatFailsReal, p.nClassesZero);
    fmt.Printf("Final = %d. Miter = %d. Total = %d. Mux = %d. (Exor = %d.) SatVars = %d.\n", Ivy_ManNodeNum(p.pManFraig), p.nNodesMiter, Ivy_ManNodeNum(p.pManAig), 0, 0, p.nSatVars);
    if p.pSat { Sat_SolverPrintStats( os.Stdout, p.pSat ); }
    // inlining of macro ABC_PRT
    Abc_Print(1, "%s =", ("AIG simulation  "));
Abc_Print(1, "%9.2f sec\n", 1.0*(float64((p.timeSim)))/(float64(int64(time.Second))));;
    // inlining of macro ABC_PRT
    Abc_Print(1, "%s =", ("AIG traversal   "));
Abc_Print(1, "%9.2f sec\n", 1.0*(float64((p.timeTrav)))/(float64(int64(time.Second))));;
    // inlining of macro ABC_PRT
    Abc_Print(1, "%s =", ("SAT solving     "));
Abc_Print(1, "%9.2f sec\n", 1.0*(float64((p.timeSat)))/(float64(int64(time.Second))));;
    // inlining of macro ABC_PRT
    Abc_Print(1, "%s =", ("    Unsat       "));
Abc_Print(1, "%9.2f sec\n", 1.0*(float64((p.timeSatUnsat)))/(float64(int64(time.Second))));;
    // inlining of macro ABC_PRT
    Abc_Print(1, "%s =", ("    Sat         "));
Abc_Print(1, "%9.2f sec\n", 1.0*(float64((p.timeSatSat)))/(float64(int64(time.Second))));;
    // inlining of macro ABC_PRT
    Abc_Print(1, "%s =", ("    Fail        "));
Abc_Print(1, "%9.2f sec\n", 1.0*(float64((p.timeSatFail)))/(float64(int64(time.Second))));;
    // inlining of macro ABC_PRT
    Abc_Print(1, "%s =", ("Class refining  "));
Abc_Print(1, "%9.2f sec\n", 1.0*(float64((p.timeRef)))/(float64(int64(time.Second))));;
    // inlining of macro ABC_PRT
    Abc_Print(1, "%s =", ("TOTAL RUNTIME   "));
Abc_Print(1, "%9.2f sec\n", 1.0*(float64((p.timeTotal)))/(float64(int64(time.Second))));;
    // inlining of macro ABC_PRT
    if p.time1 { Abc_Print(1, "%s =", ("time1           "));
Abc_Print(1, "%9.2f sec\n", 1.0*(float64((p.time1)))/(float64(int64(time.Second))));; }
}



/**Function*************************************************************

  Synopsis    [Assigns random patterns to the PI node.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Ivy_NodeAssignRandom(p *Ivy_FraigMan_t, pObj *Ivy_Obj_t) {
    var pSims *Ivy_FraigSim_t;
    var i int32;
    assert(Ivy_ObjIsPi(pObj), "Ivy_ObjIsPi(pObj)");
    pSims = Ivy_ObjSim(pObj);
    for  i = 0; i < p.nSimWords; i++  {
        pSims.pData[i] = Ivy_ObjRandomSim();
    }
}

/**Function*************************************************************

  Synopsis    [Assigns constant patterns to the PI node.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Ivy_NodeAssignConst(p *Ivy_FraigMan_t, pObj *Ivy_Obj_t, fConst1 int32) {
    var pSims *Ivy_FraigSim_t;
    var i int32;
    assert(Ivy_ObjIsPi(pObj), "Ivy_ObjIsPi(pObj)");
    pSims = Ivy_ObjSim(pObj);
    for  i = 0; i < p.nSimWords; i++  {
        pSims.pData[i] = fConst1? ~uint32(0) : 0;
    }
}

/**Function*************************************************************

  Synopsis    [Assings random simulation info for the PIs.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Ivy_FraigAssignRandom(p *Ivy_FraigMan_t) {
    var pObj *Ivy_Obj_t;
    var i int32;
    // inlining of macro Ivy_ManForEachPi
    // inlining of macro Vec_PtrForEachEntry
    for  i = 0; (i < Vec_PtrSize(p.pManAig.vPis)) && (((pObj) = (Ivy_Obj_t *)Vec_PtrEntry(p.pManAig.vPis, i)), 1); i++  {
        Ivy_NodeAssignRandom( p, pObj );
    }
}

/**Function*************************************************************

  Synopsis    [Assings distance-1 simulation info for the PIs.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Ivy_FraigAssignDist1(p *Ivy_FraigMan_t, pPat *uint32) {
    var pObj *Ivy_Obj_t;
    var i, Limit int32;
    // inlining of macro Ivy_ManForEachPi
    // inlining of macro Vec_PtrForEachEntry
    for  i = 0; (i < Vec_PtrSize(p.pManAig.vPis)) && (((pObj) = (Ivy_Obj_t *)Vec_PtrEntry(p.pManAig.vPis, i)), 1); i++  {
        Ivy_NodeAssignConst( p, pObj, Ivy_InfoHasBit(pPat, i) );
//        printf( "%d", Ivy_InfoHasBit(pPat, i) );
    }
//    printf( "\n" );

    // inlining of macro IVY_MIN
    
    if (Ivy_ManPiNum(p.pManAig)) < (p.nSimWords * 32 - 1) {
        Limit = (Ivy_ManPiNum(p.pManAig));
    } else {
        Limit = (p.nSimWords * 32 - 1);
    }
    for  i = 0; i < Limit; i++  {
        Ivy_InfoXorBit( Ivy_ObjSim( Ivy_ManPi(p.pManAig,i) ).pData, i+1 );
    }
}

/**Function*************************************************************

  Synopsis    [Returns 1 if simulation info is composed of all zeros.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Ivy_NodeHasZeroSim(p *Ivy_FraigMan_t, pObj *Ivy_Obj_t) int32 {
    var pSims *Ivy_FraigSim_t;
    var i int32;
    pSims = Ivy_ObjSim(pObj);
    for  i = 0; i < p.nSimWords; i++  {
        if pSims.pData[i] {
            return 0; }
    }
    return 1;
}

/**Function*************************************************************

  Synopsis    [Returns 1 if simulation info is composed of all zeros.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Ivy_NodeComplementSim(p *Ivy_FraigMan_t, pObj *Ivy_Obj_t) {
    var pSims *Ivy_FraigSim_t;
    var i int32;
    pSims = Ivy_ObjSim(pObj);
    for  i = 0; i < p.nSimWords; i++  {
        pSims.pData[i] = ~pSims.pData[i];
    }
}

/**Function*************************************************************

  Synopsis    [Returns 1 if simulation infos are equal.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Ivy_NodeCompareSims(p *Ivy_FraigMan_t, pObj0 *Ivy_Obj_t, pObj1 *Ivy_Obj_t) int32 {
    var pSims0, pSims1 *Ivy_FraigSim_t;
    var i int32;
    pSims0 = Ivy_ObjSim(pObj0);
    pSims1 = Ivy_ObjSim(pObj1);
    for  i = 0; i < p.nSimWords; i++  {
        if pSims0.pData[i] != pSims1.pData[i] {
            return 0; }
    }
    return 1;
}

/**Function*************************************************************

  Synopsis    [Simulates one node.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Ivy_NodeSimulateSim(p *Ivy_FraigMan_t, pSims *Ivy_FraigSim_t) {
    var pData, pData0, pData1 *uint32;
    var i int32;
    pData  = pSims.pData;
    pData0 = pSims.pFanin0.pData;
    pData1 = pSims.pFanin1.pData;
    switch  pSims.Type  {
    case 0:
        for ( i = 0; i < p.nSimWords; i++ )
            pData[i] = (pData0[i] & pData1[i]);
        break;
    case 1:
        for ( i = 0; i < p.nSimWords; i++ )
            pData[i] = ~(pData0[i] & pData1[i]);
        break;
    case 2:
        for ( i = 0; i < p.nSimWords; i++ )
            pData[i] = (pData0[i] & ~pData1[i]);
        break;
    case 3:
        for ( i = 0; i < p.nSimWords; i++ )
            pData[i] = (~pData0[i] | pData1[i]);
        break;
    case 4:
        for ( i = 0; i < p.nSimWords; i++ )
            pData[i] = (~pData0[i] & pData1[i]);
        break;
    case 5:
        for ( i = 0; i < p.nSimWords; i++ )
            pData[i] = (pData0[i] | ~pData1[i]);
        break;
    case 6:
        for ( i = 0; i < p.nSimWords; i++ )
            pData[i] = ~(pData0[i] | pData1[i]);
        break;
    case 7:
        for ( i = 0; i < p.nSimWords; i++ )
            pData[i] = (pData0[i] | pData1[i]);
        break;
    }
}

/**Function*************************************************************

  Synopsis    [Simulates one node.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Ivy_NodeSimulate(p *Ivy_FraigMan_t, pObj *Ivy_Obj_t) {
    var pSims, pSims0, pSims1 *Ivy_FraigSim_t;
    var fCompl, fCompl0, fCompl1, i int32;
    assert(!Ivy_IsComplement(pObj), "!Ivy_IsComplement(pObj)");
    // get hold of the simulation information
    pSims  = Ivy_ObjSim(pObj);
    pSims0 = Ivy_ObjSim(Ivy_ObjFanin0(pObj));
    pSims1 = Ivy_ObjSim(Ivy_ObjFanin1(pObj));
    // get complemented attributes of the children using their random info
    fCompl  = pObj.fPhase;
    fCompl0 = Ivy_ObjFaninPhase(Ivy_ObjChild0(pObj));
    fCompl1 = Ivy_ObjFaninPhase(Ivy_ObjChild1(pObj));
    // simulate
    if fCompl0 && fCompl1 {
        if fCompl {
            for  i = 0; i < p.nSimWords; i++  {
                pSims.pData[i] = (pSims0.pData[i] | pSims1.pData[i]);
            }
        } else {
            for  i = 0; i < p.nSimWords; i++  {
                pSims.pData[i] = ~(pSims0.pData[i] | pSims1.pData[i]);
            } }
    
    } else if fCompl0 && !fCompl1 {
        if fCompl {
            for  i = 0; i < p.nSimWords; i++  {
                pSims.pData[i] = (pSims0.pData[i] | ~pSims1.pData[i]);
            }
        } else {
            for  i = 0; i < p.nSimWords; i++  {
                pSims.pData[i] = (~pSims0.pData[i] & pSims1.pData[i]);
            } }
    
    } else if !fCompl0 && fCompl1 {
        if fCompl {
            for  i = 0; i < p.nSimWords; i++  {
                pSims.pData[i] = (~pSims0.pData[i] | pSims1.pData[i]);
            }
        } else {
            for  i = 0; i < p.nSimWords; i++  {
                pSims.pData[i] = (pSims0.pData[i] & ~pSims1.pData[i]);
            } }
    
    } else { // if ( !fCompl0 && !fCompl1 )
    
        if fCompl {
            for  i = 0; i < p.nSimWords; i++  {
                pSims.pData[i] = ~(pSims0.pData[i] & pSims1.pData[i]);
            }
        } else {
            for  i = 0; i < p.nSimWords; i++  {
                pSims.pData[i] = (pSims0.pData[i] & pSims1.pData[i]);
            } }
    }
}

/**Function*************************************************************

  Synopsis    [Computes hash value using simulation info.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
var spStatic_Ivy_NodeHash_s_FPrimes [128]int32 = [128]int32{ 
        1009, 1049, 1093, 1151, 1201, 1249, 1297, 1361, 1427, 1459, 
        1499, 1559, 1607, 1657, 1709, 1759, 1823, 1877, 1933, 1997, 
        2039, 2089, 2141, 2213, 2269, 2311, 2371, 2411, 2467, 2543, 
        2609, 2663, 2699, 2741, 2797, 2851, 2909, 2969, 3037, 3089, 
        3169, 3221, 3299, 3331, 3389, 3461, 3517, 3557, 3613, 3671, 
        3719, 3779, 3847, 3907, 3943, 4013, 4073, 4129, 4201, 4243, 
        4289, 4363, 4441, 4493, 4549, 4621, 4663, 4729, 4793, 4871, 
        4933, 4973, 5021, 5087, 5153, 5227, 5281, 5351, 5417, 5471, 
        5519, 5573, 5651, 5693, 5749, 5821, 5861, 5923, 6011, 6073, 
        6131, 6199, 6257, 6301, 6353, 6397, 6481, 6563, 6619, 6689, 
        6737, 6803, 6863, 6917, 6977, 7027, 7109, 7187, 7237, 7309, 
        7393, 7477, 7523, 7561, 7607, 7681, 7727, 7817, 7877, 7933, 
        8011, 8039, 8059, 8081, 8093, 8111, 8123, 8147,
    };
func Ivy_NodeHash(p *Ivy_FraigMan_t, pObj *Ivy_Obj_t) uint32 {
    
    var pSims *Ivy_FraigSim_t;
    var uHash uint32;
    var i int32;
    assert(p.nSimWords <= 128, "p.nSimWords <= 128");
    uHash = 0;
    pSims  = Ivy_ObjSim(pObj);
    for  i = 0; i < p.nSimWords; i++  {
        uHash ^= pSims.pData[i] * spStatic_Ivy_NodeHash_s_FPrimes[i];
    }
    return uHash;
}

/**Function*************************************************************

  Synopsis    [Simulates AIG manager.]

  Description [Assumes that the PI simulation info is attached.]
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Ivy_FraigSimulateOne(p *Ivy_FraigMan_t) {
    var pObj *Ivy_Obj_t;
    var i int32;
    var clk abctime;
clk = Abc_Clock();
    // inlining of macro Ivy_ManForEachNode
    // inlining of macro Ivy_ManForEachObj
    // inlining of macro Vec_PtrForEachEntry
    for  i = 0; (i < Vec_PtrSize(p.pManAig.vObjs)) && (((pObj) = (Ivy_Obj_t *)Vec_PtrEntry(p.pManAig.vObjs, i)), 1); i++  { if (pObj) == nil {} else if !Ivy_ObjIsNode(pObj) {} else {
        Ivy_NodeSimulate( p, pObj );
/*
        if ( Ivy_ObjFraig(pObj) == NULL )
            printf( "%3d --- -- %d  :  ", pObj->Id, pObj->fPhase );
        else
            printf( "%3d %3d %2d %d  :  ", pObj->Id, Ivy_Regular(Ivy_ObjFraig(pObj))->Id, Ivy_ObjSatNum(Ivy_Regular(Ivy_ObjFraig(pObj))), pObj->fPhase );
        Extra_PrintBinary( stdout, Ivy_ObjSim(pObj), 30 );
        printf( "\n" );
*/
    }
    }
p.timeSim += Abc_Clock() - clk;
p.nSimRounds++;
}

/**Function*************************************************************

  Synopsis    [Simulates AIG manager.]

  Description [Assumes that the PI simulation info is attached.]
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Ivy_FraigSimulateOneSim(p *Ivy_FraigMan_t) {
    var pSims *Ivy_FraigSim_t;
    var clk abctime;
clk = Abc_Clock();
    for  pSims = p.pSimStart; (pSims != nil); pSims = pSims.pNext  {
        Ivy_NodeSimulateSim( p, pSims );
    }
p.timeSim += Abc_Clock() - clk;
p.nSimRounds++;
}

/**Function*************************************************************

  Synopsis    [Adds one node to the equivalence class.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Ivy_NodeAddToClass(pClass *Ivy_Obj_t, pObj *Ivy_Obj_t) {
    if Ivy_ObjClassNodeNext(pClass) == nil {
        Ivy_ObjSetClassNodeNext( pClass, pObj );
    } else {
        Ivy_ObjSetClassNodeNext( Ivy_ObjClassNodeLast(pClass), pObj ); }
    Ivy_ObjSetClassNodeLast( pClass, pObj );
    Ivy_ObjSetClassNodeRepr( pObj, pClass );
    Ivy_ObjSetClassNodeNext( pObj, nil );
}

/**Function*************************************************************

  Synopsis    [Adds equivalence class to the list of classes.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Ivy_FraigAddClass(pList *Ivy_FraigList_t, pClass *Ivy_Obj_t) {
    if pList.pHead == nil {
        pList.pHead = pClass;
        pList.pTail = pClass;
        Ivy_ObjSetEquivListPrev( pClass, nil );
        Ivy_ObjSetEquivListNext( pClass, nil ); 
    
    } else {
        Ivy_ObjSetEquivListNext( pList.pTail, pClass ); 
        Ivy_ObjSetEquivListPrev( pClass, pList.pTail );
        Ivy_ObjSetEquivListNext( pClass, nil ); 
        pList.pTail = pClass;
    }
    pList.nItems++;
}
 
/**Function*************************************************************

  Synopsis    [Updates the list of classes after base class has split.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Ivy_FraigInsertClass(pList *Ivy_FraigList_t, pBase *Ivy_Obj_t, pClass *Ivy_Obj_t) {
    Ivy_ObjSetEquivListPrev( pClass, pBase );
    Ivy_ObjSetEquivListNext( pClass, Ivy_ObjEquivListNext(pBase) ); 
    if Ivy_ObjEquivListNext(pBase) != nil {
        Ivy_ObjSetEquivListPrev( Ivy_ObjEquivListNext(pBase), pClass ); }
    Ivy_ObjSetEquivListNext( pBase, pClass ); 
    if pList.pTail == pBase {
        pList.pTail = pClass; }
    pList.nItems++;
}

/**Function*************************************************************

  Synopsis    [Removes equivalence class from the list of classes.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Ivy_FraigRemoveClass(pList *Ivy_FraigList_t, pClass *Ivy_Obj_t) {
    if pList.pHead == pClass {
        pList.pHead = Ivy_ObjEquivListNext(pClass); }
    if pList.pTail == pClass {
        pList.pTail = Ivy_ObjEquivListPrev(pClass); }
    if Ivy_ObjEquivListPrev(pClass) != nil {
        Ivy_ObjSetEquivListNext( Ivy_ObjEquivListPrev(pClass), Ivy_ObjEquivListNext(pClass) ); } 
    if Ivy_ObjEquivListNext(pClass) != nil {
        Ivy_ObjSetEquivListPrev( Ivy_ObjEquivListNext(pClass), Ivy_ObjEquivListPrev(pClass) ); }
    Ivy_ObjSetEquivListNext( pClass, nil ); 
    Ivy_ObjSetEquivListPrev( pClass, nil );
    pClass.fMarkA = 0;
    pList.nItems--;
}

/**Function*************************************************************

  Synopsis    [Count the number of pairs.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Ivy_FraigCountPairsClasses(p *Ivy_FraigMan_t) int32 {
    var pClass, pNode *Ivy_Obj_t;
    var nPairs, nNodes int32;
    return nPairs;

    // inlining of macro Ivy_FraigForEachEquivClass
    for  pClass = p.lClasses.pHead;                                                     (pClass != nil);                                                             pClass = Ivy_ObjEquivListNext(pClass)  {
        nNodes = 0;
        // inlining of macro Ivy_FraigForEachClassNode
        for  pNode = pClass;                                                    (pNode != nil);                                                             pNode = Ivy_ObjClassNodeNext(pNode)  {
            nNodes++;
        }
        nPairs += nNodes * (nNodes - 1) / 2;
    }
    return nPairs;
}

/**Function*************************************************************

  Synopsis    [Creates initial simulation classes.]

  Description [Assumes that simulation info is assigned.]
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Ivy_FraigCreateClasses(p *Ivy_FraigMan_t) {
    var pTable **Ivy_Obj_t;
    var pObj, pConst1, pBin, pEntry *Ivy_Obj_t;
    var i, nTableSize int32;
    var Hash uint32;
    pConst1 = Ivy_ManConst1(p.pManAig);
    // allocate the table
    nTableSize = Ivy_ManObjNum(p.pManAig) / 2 + 13;
    // inlining of macro ABC_ALLOC
    pTable = ((Ivy_Obj_t * *) malloc(sizeof(Ivy_Obj_t *) * (size_t)(nTableSize))); 
    memset( pTable, 0, sizeof(Ivy_Obj_t *) * nTableSize );
    // collect nodes into the table
    // inlining of macro Ivy_ManForEachObj
    // inlining of macro Vec_PtrForEachEntry
    for  i = 0; (i < Vec_PtrSize(p.pManAig.vObjs)) && (((pObj) = (Ivy_Obj_t *)Vec_PtrEntry(p.pManAig.vObjs, i)), 1); i++  { if (pObj) == nil {} else {
        if !Ivy_ObjIsPi(pObj) && !Ivy_ObjIsNode(pObj) {
            continue; }
        Hash = Ivy_NodeHash( p, pObj );
        if Hash == 0 && (Ivy_NodeHasZeroSim( p, pObj ) != 0) {
            Ivy_NodeAddToClass( pConst1, pObj );
            continue;
        }
        // add the node to the table
        pBin = pTable[Hash % nTableSize];
        // inlining of macro Ivy_FraigForEachBinNode
        for  pEntry = pBin;                                                      (pEntry != nil);                                                             pEntry = Ivy_ObjNodeHashNext(pEntry)  {
            if Ivy_NodeCompareSims( p, pEntry, pObj ) != 0 {
                Ivy_NodeAddToClass( pEntry, pObj );
                break;
            }
        }
        // check if the entry was added
        if pEntry != nil {
            continue; }
        Ivy_ObjSetNodeHashNext( pObj, pBin );
        pTable[Hash % nTableSize] = pObj;
    }
    }
    // collect non-trivial classes
    assert(p.lClasses.pHead == nil, "p.lClasses.pHead == nil");
    // inlining of macro Ivy_ManForEachObj
    // inlining of macro Vec_PtrForEachEntry
    for  i = 0; (i < Vec_PtrSize(p.pManAig.vObjs)) && (((pObj) = (Ivy_Obj_t *)Vec_PtrEntry(p.pManAig.vObjs, i)), 1); i++  { if (pObj) == nil {} else {
        if !Ivy_ObjIsConst1(pObj) && !Ivy_ObjIsPi(pObj) && !Ivy_ObjIsNode(pObj) {
            continue; }
        Ivy_ObjSetNodeHashNext( pObj, nil );
        if Ivy_ObjClassNodeRepr(pObj) == nil {
            assert(Ivy_ObjClassNodeNext(pObj) == nil, "Ivy_ObjClassNodeNext(pObj) == nil");
            continue;
        }
        // recognize the head of the class
        if Ivy_ObjClassNodeNext( Ivy_ObjClassNodeRepr(pObj) ) != nil {
            continue; }
        // clean the class representative and add it to the list
        Ivy_ObjSetClassNodeRepr( pObj, nil );
        Ivy_FraigAddClass( &p.lClasses, pObj );
    }
    }
    // free the table
    // inlining of macro ABC_FREE
    ((pTable) ? (free((char *) (pTable)), (pTable) = 0) : 0);
}

/**Function*************************************************************

  Synopsis    [Recursively refines the class after simulation.]

  Description [Returns 1 if the class has changed.]
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Ivy_FraigRefineClass_rec(p *Ivy_FraigMan_t, pClass *Ivy_Obj_t) int32 {
    var pClassNew, pListOld, pListNew, pNode *Ivy_Obj_t;
    var RetValue int32;
    // check if there is refinement
    pListOld = pClass;
    // inlining of macro Ivy_FraigForEachClassNode
    for  pClassNew = Ivy_ObjClassNodeNext(pClass);                                                    (pClassNew != nil);                                                             pClassNew = Ivy_ObjClassNodeNext(pClassNew)  {
        if !(Ivy_NodeCompareSims(p, pClass, pClassNew) != 0) {
            if p.pParams.fPatScores {
                Ivy_FraigAddToPatScores( p, pClass, pClassNew ); }
            break;
        }
        pListOld = pClassNew;
    }
    if pClassNew == nil {
        return 0; }
    // set representative of the new class
    Ivy_ObjSetClassNodeRepr( pClassNew, nil );
    // start the new list
    pListNew = pClassNew;
    // go through the remaining nodes and sort them into two groups:
    // (1) matches of the old node; (2) non-matches of the old node
    // inlining of macro Ivy_FraigForEachClassNode
    for  pNode = Ivy_ObjClassNodeNext(pClassNew);                                                    (pNode != nil);                                                             pNode = Ivy_ObjClassNodeNext(pNode)  {
        if Ivy_NodeCompareSims( p, pClass, pNode ) != 0 {
            Ivy_ObjSetClassNodeNext( pListOld, pNode );
            pListOld = pNode;
        
        } else {
            Ivy_ObjSetClassNodeNext( pListNew, pNode );
            Ivy_ObjSetClassNodeRepr( pNode, pClassNew );
            pListNew = pNode;
        }
    }
    // finish both lists
    Ivy_ObjSetClassNodeNext( pListNew, nil );
    Ivy_ObjSetClassNodeNext( pListOld, nil );
    // update the list of classes
    Ivy_FraigInsertClass( &p.lClasses, pClass, pClassNew );
    // if the old class is trivial, remove it
    if Ivy_ObjClassNodeNext(pClass) == nil {
        Ivy_FraigRemoveClass( &p.lClasses, pClass ); }
    // if the new class is trivial, remove it; otherwise, try to refine it
    if Ivy_ObjClassNodeNext(pClassNew) == nil {
        Ivy_FraigRemoveClass( &p.lClasses, pClassNew );
    } else {
        RetValue = Ivy_FraigRefineClass_rec( p, pClassNew ); }
    return RetValue + 1;
}
 
/**Function*************************************************************

  Synopsis    [Creates the counter-example from the successful pattern.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Ivy_FraigCheckOutputSimsSavePattern(p *Ivy_FraigMan_t, pObj *Ivy_Obj_t) { 
    var pSims *Ivy_FraigSim_t;
    var i int32;
    var k int32;
    var BestPat int32;
    var pModel *int32;
    // find the word of the pattern
    pSims = Ivy_ObjSim(pObj);
    for  i = 0; i < p.nSimWords; i++  {
        if pSims.pData[i] {
            break; }
    }
    assert(i < p.nSimWords, "i < p.nSimWords");
    // find the bit of the pattern
    for  k = 0; k < 32; k++  {
        if pSims.pData[i] & (1 << k) {
            break; }
    }
    assert(k < 32, "k < 32");
    // determine the best pattern
    BestPat = i * 32 + k;
    // fill in the counter-example data
    // inlining of macro ABC_ALLOC
    pModel = ((int *) malloc(sizeof(int) * (size_t)(Ivy_ManPiNum(p.pManFraig))));
    // inlining of macro Ivy_ManForEachPi
    // inlining of macro Vec_PtrForEachEntry
    for  i = 0; (i < Vec_PtrSize(p.pManAig.vPis)) && (((pObj) = (Ivy_Obj_t *)Vec_PtrEntry(p.pManAig.vPis, i)), 1); i++  {
        pModel[i] = Ivy_InfoHasBit(Ivy_ObjSim(pObj).pData, BestPat);
//        printf( "%d", pModel[i] );
    }
//    printf( "\n" );
    // set the model
    assert(p.pManFraig.pData == nil, "p.pManFraig.pData == nil");
    p.pManFraig.pData = pModel;
    return;
}

/**Function*************************************************************

  Synopsis    [Returns 1 if the one of the output is already non-constant 0.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Ivy_FraigCheckOutputSims(p *Ivy_FraigMan_t) int32 {
    var pObj *Ivy_Obj_t;
    var i int32;
    // make sure the reference simulation pattern does not detect the bug
//    pObj = Ivy_ManPo( p->pManAig, 0 );
    // inlining of macro Ivy_ManForEachPo
    // inlining of macro Vec_PtrForEachEntry
    for  i = 0; (i < Vec_PtrSize(p.pManAig.vPos)) && (((pObj) = (Ivy_Obj_t *)Vec_PtrEntry(p.pManAig.vPos, i)), 1); i++  {
        //assert( Ivy_ObjFanin0(pObj)->fPhase == (unsigned)Ivy_ObjFaninC0(pObj) ); // Ivy_ObjFaninPhase(Ivy_ObjChild0(pObj)) == 0
        // complement simulation info
//        if ( Ivy_ObjFanin0(pObj)->fPhase ^ Ivy_ObjFaninC0(pObj) ) // Ivy_ObjFaninPhase(Ivy_ObjChild0(pObj))
//            Ivy_NodeComplementSim( p, Ivy_ObjFanin0(pObj) );
        // check 
        if !(Ivy_NodeHasZeroSim( p, Ivy_ObjFanin0(pObj) ) != 0) {
            // create the counter-example from this pattern
            Ivy_FraigCheckOutputSimsSavePattern( p, Ivy_ObjFanin0(pObj) );
            return 1;
        }
        // complement simulation info
//        if ( Ivy_ObjFanin0(pObj)->fPhase ^ Ivy_ObjFaninC0(pObj) )
//            Ivy_NodeComplementSim( p, Ivy_ObjFanin0(pObj) );
    }
    return 0;
}

/**Function*************************************************************

  Synopsis    [Refines the classes after simulation.]

  Description [Assumes that simulation info is assigned. Returns the
  number of classes refined.]
               
  SideEffects [Large equivalence class of constant 0 may cause problems.]

  SeeAlso     []

***********************************************************************/
func Ivy_FraigRefineClasses(p *Ivy_FraigMan_t) int32 {
    var pClass, pClass2 *Ivy_Obj_t;
    var RetValue, Counter int32;
    var clk abctime;
    // check if some outputs already became non-constant
    // this is a special case when computation can be stopped!!!
    if p.pParams.fProve {
        Ivy_FraigCheckOutputSims( p ); }
    if p.pManFraig.pData {
        return 0; }
    // refine the classed
clk = Abc_Clock();
    // inlining of macro Ivy_FraigForEachEquivClassSafe
    for  pClass = p.lClasses.pHead,                                                     pClass2 = pClass? Ivy_ObjEquivListNext(pClass): nil;                   (pClass != nil);                                                             pClass = pClass2,                                                     pClass2 = pClass? Ivy_ObjEquivListNext(pClass): nil  {
        if pClass.fMarkA {
            continue; }
        RetValue = Ivy_FraigRefineClass_rec( p, pClass );
        Counter += ( RetValue > 0 );
//if ( Ivy_ObjIsConst1(pClass) )
//printf( "%d ", RetValue );
//if ( Ivy_ObjIsConst1(pClass) )
//    p->time1 += Abc_Clock() - clk;
    }
p.timeRef += Abc_Clock() - clk;
    return Counter;
}

/**Function*************************************************************

  Synopsis    [Print the class.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Ivy_FraigPrintClass(pClass *Ivy_Obj_t) {
    var pObj *Ivy_Obj_t;
    fmt.Printf("Class {");
    // inlining of macro Ivy_FraigForEachClassNode
    for  pObj = pClass;                                                    pObj;                                                             pObj = Ivy_ObjClassNodeNext(pObj)  {
        fmt.Printf(" %d(%d)%c", pObj.Id, pObj.Level, pObj.fPhase? '+' : '-');
    }
    fmt.Printf(" }\n");
}

/**Function*************************************************************

  Synopsis    [Count the number of elements in the class.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Ivy_FraigCountClassNodes(pClass *Ivy_Obj_t) int32 {
    var pObj *Ivy_Obj_t;
    var Counter int32;
    // inlining of macro Ivy_FraigForEachClassNode
    for  pObj = pClass;                                                    pObj;                                                             pObj = Ivy_ObjClassNodeNext(pObj)  {
        Counter++;
    }
    return Counter;
}

/**Function*************************************************************

  Synopsis    [Prints simulation classes.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Ivy_FraigPrintSimClasses(p *Ivy_FraigMan_t) {
    var pClass *Ivy_Obj_t;
    // inlining of macro Ivy_FraigForEachEquivClass
    for  pClass = p.lClasses.pHead;                                                     (pClass != nil);                                                             pClass = Ivy_ObjEquivListNext(pClass)  {
//        Ivy_FraigPrintClass( pClass );
        fmt.Printf("%d ", Ivy_FraigCountClassNodes( pClass ));
    }
//    printf( "\n" );
}

/**Function*************************************************************

  Synopsis    [Generated const 0 pattern.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Ivy_FraigSavePattern0(p *Ivy_FraigMan_t) {
    memset( p.pPatWords, 0, sizeof(unsigned) * p.nPatWords );
}

/**Function*************************************************************

  Synopsis    [[Generated const 1 pattern.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Ivy_FraigSavePattern1(p *Ivy_FraigMan_t) {
    memset( p.pPatWords, 0xff, sizeof(unsigned) * p.nPatWords );
}

/**Function*************************************************************

  Synopsis    [Generates the counter-example satisfying the miter.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Ivy_FraigCreateModel(p *Ivy_FraigMan_t) *int32 {
    var pModel *int32;
    var pObj *Ivy_Obj_t;
    var i int32;
    // inlining of macro ABC_ALLOC
    pModel = ((int *) malloc(sizeof(int) * (size_t)(Ivy_ManPiNum(p.pManFraig))));
    // inlining of macro Ivy_ManForEachPi
    // inlining of macro Vec_PtrForEachEntry
    for  i = 0; (i < Vec_PtrSize(p.pManFraig.vPis)) && (((pObj) = (Ivy_Obj_t *)Vec_PtrEntry(p.pManFraig.vPis, i)), 1); i++  {
//        pModel[i] = ( p->pSat->model.ptr[Ivy_ObjSatNum(pObj)] == l_True );
        pModel[i] = ( p.pSat.model[Ivy_ObjSatNum(pObj)] == l_True );
    }
    return pModel;
}

/**Function*************************************************************

  Synopsis    [Copy pattern from the solver into the internal storage.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Ivy_FraigSavePattern(p *Ivy_FraigMan_t) {
    var pObj *Ivy_Obj_t;
    var i int32;
    memset( p.pPatWords, 0, sizeof(unsigned) * p.nPatWords );
    // inlining of macro Ivy_ManForEachPi
    // inlining of macro Vec_PtrForEachEntry
    for  i = 0; (i < Vec_PtrSize(p.pManFraig.vPis)) && (((pObj) = (Ivy_Obj_t *)Vec_PtrEntry(p.pManFraig.vPis, i)), 1); i++  {
//    Vec_PtrForEachEntry( Ivy_Obj_t *, p->vPiVars, pObj, i )
//        if ( p->pSat->model.ptr[Ivy_ObjSatNum(pObj)] == l_True )
        if p.pSat.model[Ivy_ObjSatNum(pObj)] == l_True {
            Ivy_InfoSetBit( p.pPatWords, i ); }
    }
//            Ivy_InfoSetBit( p->pPatWords, pObj->Id - 1 );
}

/**Function*************************************************************

  Synopsis    [Copy pattern from the solver into the internal storage.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Ivy_FraigSavePattern2(p *Ivy_FraigMan_t) {
    var pObj *Ivy_Obj_t;
    var i int32;
    memset( p.pPatWords, 0, sizeof(unsigned) * p.nPatWords );
//    Ivy_ManForEachPi( p->pManFraig, pObj, i )
    // inlining of macro Vec_PtrForEachEntry
    for  i = 0; (i < Vec_PtrSize(p.vPiVars)) && (((pObj) = (Ivy_Obj_t *)Vec_PtrEntry(p.vPiVars, i)), 1); i++  {
//        if ( p->pSat->model.ptr[Ivy_ObjSatNum(pObj)] == l_True )
        if p.pSat.model[Ivy_ObjSatNum(pObj)] == l_True {
//            Ivy_InfoSetBit( p->pPatWords, i );
            Ivy_InfoSetBit( p.pPatWords, pObj.Id - 1 ); }
    }
}

/**Function*************************************************************

  Synopsis    [Copy pattern from the solver into the internal storage.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Ivy_FraigSavePattern3(p *Ivy_FraigMan_t) {
    var pObj *Ivy_Obj_t;
    var i int32;
    for  i = 0; i < p.nPatWords; i++  {
        p.pPatWords[i] = Ivy_ObjRandomSim();
    }
    // inlining of macro Vec_PtrForEachEntry
    for  i = 0; (i < Vec_PtrSize(p.vPiVars)) && (((pObj) = (Ivy_Obj_t *)Vec_PtrEntry(p.vPiVars, i)), 1); i++  {
//        if ( Ivy_InfoHasBit( p->pPatWords, pObj->Id - 1 ) ^ (p->pSat->model.ptr[Ivy_ObjSatNum(pObj)] == l_True) )
        if Ivy_InfoHasBit( p.pPatWords, pObj.Id - 1 ) ^ sat_solver_var_value(p.pSat, Ivy_ObjSatNum(pObj)) {
            Ivy_InfoXorBit( p.pPatWords, pObj.Id - 1 ); }
    }
}


/**Function*************************************************************

  Synopsis    [Performs simulation of the manager.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Ivy_FraigSimulate(p *Ivy_FraigMan_t) {
    var nChanges, nClasses int32;
    // start the classes
    Ivy_FraigAssignRandom( p );
    Ivy_FraigSimulateOne( p );
    Ivy_FraigCreateClasses( p );
//printf( "Starting classes = %5d.   Pairs = %6d.\n", p->lClasses.nItems, Ivy_FraigCountPairsClasses(p) );
    // refine classes by walking 0/1 patterns
    Ivy_FraigSavePattern0( p );
    Ivy_FraigAssignDist1( p, p.pPatWords );
    Ivy_FraigSimulateOne( p );
    nChanges = Ivy_FraigRefineClasses( p );
    if p.pManFraig.pData {
        return; }
//printf( "Refined classes  = %5d.   Changes = %4d.   Pairs = %6d.\n", p->lClasses.nItems, nChanges, Ivy_FraigCountPairsClasses(p) );
    Ivy_FraigSavePattern1( p );
    Ivy_FraigAssignDist1( p, p.pPatWords );
    Ivy_FraigSimulateOne( p );
    nChanges = Ivy_FraigRefineClasses( p );
    if p.pManFraig.pData {
        return; }
//printf( "Refined classes  = %5d.   Changes = %4d.   Pairs = %6d.\n", p->lClasses.nItems, nChanges, Ivy_FraigCountPairsClasses(p) );
    // refine classes by random simulation
    for {
        Ivy_FraigAssignRandom( p );
        Ivy_FraigSimulateOne( p );
        nClasses = p.lClasses.nItems;
        nChanges = Ivy_FraigRefineClasses( p );
        if p.pManFraig.pData {
            return; }
//printf( "Refined classes  = %5d.   Changes = %4d.   Pairs = %6d.\n", p->lClasses.nItems, nChanges, Ivy_FraigCountPairsClasses(p) );
    
if !( float64(nChanges) / nClasses > p.pParams.dSimSatur ) { break }
} 
//    Ivy_FraigPrintSimClasses( p );
}



/**Function*************************************************************

  Synopsis    [Cleans pattern scores.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Ivy_FraigCleanPatScores(p *Ivy_FraigMan_t) {
    var i int32;
    var nLimit int32 = p.nSimWords * 32;
    for  i = 0; i < nLimit; i++  {
        p.pPatScores[i] = 0;
    }
}

/**Function*************************************************************

  Synopsis    [Adds to pattern scores.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Ivy_FraigAddToPatScores(p *Ivy_FraigMan_t, pClass *Ivy_Obj_t, pClassNew *Ivy_Obj_t) {
    var pSims0, pSims1 *Ivy_FraigSim_t;
    var uDiff uint32;
    var i, w int32;
    // get hold of the simulation information
    pSims0 = Ivy_ObjSim(pClass);
    pSims1 = Ivy_ObjSim(pClassNew);
    // iterate through the differences and record the score
    for  w = 0; w < p.nSimWords; w++  {
        uDiff = pSims0.pData[w] ^ pSims1.pData[w];
        if uDiff == 0 {
            continue; }
        for  i = 0; i < 32; i++  {
            if uDiff & ( 1 << i ) {
                p.pPatScores[w*32+i]++; }
        }
    }
}

/**Function*************************************************************

  Synopsis    [Selects the best pattern.]

  Description [Returns 1 if such pattern is found.]
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Ivy_FraigSelectBestPat(p *Ivy_FraigMan_t) int32 {
    var pSims *Ivy_FraigSim_t;
    var pObj *Ivy_Obj_t;
    var i int32;
    var nLimit int32 = p.nSimWords * 32;
    var MaxScore int32;
    var BestPat int32 = -1;
    for  i = 1; i < nLimit; i++  {
        if MaxScore < p.pPatScores[i] {
            MaxScore = p.pPatScores[i];
            BestPat = i;
        }
    }
    if MaxScore == 0 {
        return 0; }
//    if ( MaxScore > p->pParams->MaxScore )
//    printf( "Max score is %3d.  ", MaxScore );
    // copy the best pattern into the selected pattern
    memset( p.pPatWords, 0, sizeof(unsigned) * p.nPatWords );
    // inlining of macro Ivy_ManForEachPi
    // inlining of macro Vec_PtrForEachEntry
    for  i = 0; (i < Vec_PtrSize(p.pManAig.vPis)) && (((pObj) = (Ivy_Obj_t *)Vec_PtrEntry(p.pManAig.vPis, i)), 1); i++  {
        pSims = Ivy_ObjSim(pObj);
        if Ivy_InfoHasBit(pSims.pData, BestPat) {
            Ivy_InfoSetBit(p.pPatWords, i); }
    }
    return MaxScore;
}

/**Function*************************************************************

  Synopsis    [Resimulates fraiging manager after finding a counter-example.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Ivy_FraigResimulate(p *Ivy_FraigMan_t) {
    var nChanges int32;
    Ivy_FraigAssignDist1( p, p.pPatWords );
    Ivy_FraigSimulateOne( p );
    if p.pParams.fPatScores {
        Ivy_FraigCleanPatScores( p ); }
    nChanges = Ivy_FraigRefineClasses( p );
    if p.pManFraig.pData {
        return; }
    if nChanges < 1 {
        fmt.Printf("Error: A counter-example did not refine classes!\n"); }
    assert(nChanges >= 1, "nChanges >= 1");
//printf( "Refined classes! = %5d.   Changes = %4d.\n", p->lClasses.nItems, nChanges );
    if !p.pParams.fPatScores {
        return; }

    // perform additional simulation using dist1 patterns derived from successful patterns
    for  Ivy_FraigSelectBestPat(p) > p.pParams.MaxScore  {
        Ivy_FraigAssignDist1( p, p.pPatWords );
        Ivy_FraigSimulateOne( p );
        Ivy_FraigCleanPatScores( p );
        nChanges = Ivy_FraigRefineClasses( p );
        if p.pManFraig.pData {
            return; }
//printf( "Refined class!!! = %5d.   Changes = %4d.   Pairs = %6d.\n", p->lClasses.nItems, nChanges, Ivy_FraigCountPairsClasses(p) );
        if nChanges == 0 {
            break; }
    }
}


/**Function*************************************************************

  Synopsis    [Prints the status of the miter.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Ivy_FraigMiterPrint(pNtk *Ivy_Man_t, pString string, clk abctime, fVerbose int32) {
    if !(fVerbose != 0) {
        return; }
    fmt.Printf("Nodes = %7d.  Levels = %4d.  ", Ivy_ManNodeNum(pNtk), Ivy_ManLevels(pNtk));
    // inlining of macro ABC_PRT
    Abc_Print(1, "%s =", (pString));
Abc_Print(1, "%9.2f sec\n", 1.0*(float64((Abc_Clock() - clk)))/(float64(int64(time.Second))));;
}

/**Function*************************************************************

  Synopsis    [Reports the status of the miter.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Ivy_FraigMiterStatus(pMan *Ivy_Man_t) int32 {
    var pObj, pObjNew *Ivy_Obj_t;
    var i, CountConst0, CountNonConst0, CountUndecided int32;
    if pMan.pData {
        return 0; }
    // inlining of macro Ivy_ManForEachPo
    // inlining of macro Vec_PtrForEachEntry
    for  i = 0; (i < Vec_PtrSize(pMan.vPos)) && (((pObj) = (Ivy_Obj_t *)Vec_PtrEntry(pMan.vPos, i)), 1); i++  {
        pObjNew = Ivy_ObjChild0(pObj);
        // check if the output is constant 1
        if pObjNew == pMan.pConst1 {
            CountNonConst0++;
            continue;
        }
        // check if the output is constant 0
        if pObjNew == Ivy_Not(pMan.pConst1) {
            CountConst0++;
            continue;
        }
/*
        // check if the output is a primary input
        if ( Ivy_ObjIsPi(Ivy_Regular(pObjNew)) )
        {
            CountNonConst0++;
            continue;
        }
*/
        // check if the output can be constant 0
        if Ivy_Regular(pObjNew).fPhase != uint32(Ivy_IsComplement(pObjNew)) {
            CountNonConst0++;
            continue;
        }
        CountUndecided++;
    }
/*
    if ( p->pParams->fVerbose )
    {
        printf( "Miter has %d outputs. ", Ivy_ManPoNum(p->pManAig) );
        printf( "Const0 = %d.  ", CountConst0 );
        printf( "NonConst0 = %d.  ", CountNonConst0 );
        printf( "Undecided = %d.  ", CountUndecided );
        printf( "\n" );
    }
*/
    if CountNonConst0 {
        return 0; }
    if CountUndecided != 0 {
        return -1; }
    return 1;
}

/**Function*************************************************************

  Synopsis    [Tries to prove each output of the miter until encountering a sat output.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Ivy_FraigMiterProve(p *Ivy_FraigMan_t) {
    var pObj, pObjNew *Ivy_Obj_t;
    var i, RetValue int32;
    var clk abctime  = Abc_Clock();
    var fVerbose int32;
    // inlining of macro Ivy_ManForEachPo
    // inlining of macro Vec_PtrForEachEntry
    for  i = 0; (i < Vec_PtrSize(p.pManAig.vPos)) && (((pObj) = (Ivy_Obj_t *)Vec_PtrEntry(p.pManAig.vPos, i)), 1); i++  {
        if (i != 0) && (fVerbose != 0) {
            // inlining of macro ABC_PRT
            Abc_Print(1, "%s =", ("Time"));
Abc_Print(1, "%9.2f sec\n", 1.0*(float64((Abc_Clock() -clk)))/(float64(int64(time.Second))));;
        }
        pObjNew = Ivy_ObjChild0Equiv(pObj);
        // check if the output is constant 1
        if pObjNew == p.pManFraig.pConst1 {
            if fVerbose != 0 {
                fmt.Printf("Output %2d (out of %2d) is constant 1.  ", i, Ivy_ManPoNum(p.pManAig)); }
            // assing constant 0 model
            // inlining of macro ABC_ALLOC
            p.pManFraig.pData = ((int *) malloc(sizeof(int) * (size_t)(Ivy_ManPiNum(p.pManFraig))));
            memset( p.pManFraig.pData, 0, sizeof(int) * Ivy_ManPiNum(p.pManFraig) );
            break;
        }
        // check if the output is constant 0
        if pObjNew == Ivy_Not(p.pManFraig.pConst1) {
            if fVerbose != 0 {
                fmt.Printf("Output %2d (out of %2d) is already constant 0.  ", i, Ivy_ManPoNum(p.pManAig)); }
            continue;
        }
        // check if the output can be constant 0
        if Ivy_Regular(pObjNew).fPhase != uint32(Ivy_IsComplement(pObjNew)) {
            if fVerbose != 0 {
                fmt.Printf("Output %2d (out of %2d) cannot be constant 0.  ", i, Ivy_ManPoNum(p.pManAig)); }
            // assing constant 0 model
            // inlining of macro ABC_ALLOC
            p.pManFraig.pData = ((int *) malloc(sizeof(int) * (size_t)(Ivy_ManPiNum(p.pManFraig))));
            memset( p.pManFraig.pData, 0, sizeof(int) * Ivy_ManPiNum(p.pManFraig) );
            break;
        }
/*
        // check the representative of this node
        pRepr = Ivy_ObjClassNodeRepr(Ivy_ObjFanin0(pObj));
        if ( Ivy_Regular(pRepr) != p->pManAig->pConst1 )
            printf( "Representative is not constant 1.\n" );
        else
            printf( "Representative is constant 1.\n" );
*/
        // try to prove the output constant 0
        RetValue = Ivy_FraigNodeIsConst( p, Ivy_Regular(pObjNew) );
        if RetValue == 1 {  // proved equivalent
        
            if fVerbose != 0 {
                fmt.Printf("Output %2d (out of %2d) was proved constant 0.  ", i, Ivy_ManPoNum(p.pManAig)); }
            // set the constant miter
            Ivy_ObjFanin0(pObj).pEquiv = Ivy_NotCond( p.pManFraig.pConst1, !Ivy_ObjFaninC0(pObj) );
            continue;
        }
        if RetValue == -1 { // failed
        
            if fVerbose != 0 {
                fmt.Printf("Output %2d (out of %2d) has timed out at %d backtracks.  ", i, Ivy_ManPoNum(p.pManAig), p.pParams.nBTLimitMiter); }
            continue;
        }
        // proved satisfiable
        if fVerbose != 0 {
            fmt.Printf("Output %2d (out of %2d) was proved NOT a constant 0.  ", i, Ivy_ManPoNum(p.pManAig)); }
        // create the model
        p.pManFraig.pData = Ivy_FraigCreateModel(p);
        break;
    }
    if fVerbose != 0 {
        // inlining of macro ABC_PRT
        Abc_Print(1, "%s =", ("Time"));
Abc_Print(1, "%9.2f sec\n", 1.0*(float64((Abc_Clock() -clk)))/(float64(int64(time.Second))));;
    }
}

/**Function*************************************************************

  Synopsis    [Performs fraiging for the internal nodes.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Ivy_FraigSweep(p *Ivy_FraigMan_t) {
    var pObj *Ivy_Obj_t;//, * pTemp;
    var i, k int32;
if p.lClasses.pHead {
    if Ivy_ObjIsConst1(p.lClasses.pHead) {
        p.nClassesZero = Ivy_FraigCountClassNodes(p.lClasses.pHead);
    } else {
        p.nClassesZero = 0;
    }
} else {
    p.nClassesZero = 0;
}
p.nClassesBeg  = p.lClasses.nItems;
    // duplicate internal nodes
    p.pProgress = Extra_ProgressBarStart( os.Stdout, Ivy_ManNodeNum(p.pManAig) );
    // inlining of macro Ivy_ManForEachNode
    // inlining of macro Ivy_ManForEachObj
    // inlining of macro Vec_PtrForEachEntry
    for  i = 0; (i < Vec_PtrSize(p.pManAig.vObjs)) && (((pObj) = (Ivy_Obj_t *)Vec_PtrEntry(p.pManAig.vObjs, i)), 1); i++  { if (pObj) == nil {} else if !Ivy_ObjIsNode(pObj) {} else {
        Extra_ProgressBarUpdate( p.pProgress, k++, nil );
        // default to simple strashing if simulation detected a counter-example for a PO
        if p.pManFraig.pData {
            pObj.pEquiv = Ivy_And( p.pManFraig, Ivy_ObjChild0Equiv(pObj), Ivy_ObjChild1Equiv(pObj) );
        } else {
            pObj.pEquiv = Ivy_FraigAnd( p, pObj ); }
        assert(pObj.pEquiv != nil, "pObj.pEquiv != nil");
//        pTemp = Ivy_Regular(pObj->pEquiv);
//        assert( Ivy_Regular(pObj->pEquiv)->Type );
    }
    }
    Extra_ProgressBarStop( p.pProgress );
p.nClassesEnd = p.lClasses.nItems;
    // try to prove the outputs of the miter
    p.nNodesMiter = Ivy_ManNodeNum(p.pManFraig);
//    Ivy_FraigMiterStatus( p->pManFraig );
    if p.pParams.fProve && p.pManFraig.pData == nil {
        Ivy_FraigMiterProve( p ); }
    // add the POs
    // inlining of macro Ivy_ManForEachPo
    // inlining of macro Vec_PtrForEachEntry
    for  i = 0; (i < Vec_PtrSize(p.pManAig.vPos)) && (((pObj) = (Ivy_Obj_t *)Vec_PtrEntry(p.pManAig.vPos, i)), 1); i++  {
        Ivy_ObjCreatePo( p.pManFraig, Ivy_ObjChild0Equiv(pObj) );
    }
    // clean the old manager
    // inlining of macro Ivy_ManForEachObj
    // inlining of macro Vec_PtrForEachEntry
    for  i = 0; (i < Vec_PtrSize(p.pManAig.vObjs)) && (((pObj) = (Ivy_Obj_t *)Vec_PtrEntry(p.pManAig.vObjs, i)), 1); i++  { if (pObj) == nil {} else {
        pObj.pFanout = pObj.pNextFan0 = pObj.pNextFan1 = pObj.pPrevFan0 = pObj.pPrevFan1 = nil; }
    }
    // clean the new manager
    // inlining of macro Ivy_ManForEachObj
    // inlining of macro Vec_PtrForEachEntry
    for  i = 0; (i < Vec_PtrSize(p.pManFraig.vObjs)) && (((pObj) = (Ivy_Obj_t *)Vec_PtrEntry(p.pManFraig.vObjs, i)), 1); i++  { if (pObj) == nil {} else {
        if Ivy_ObjFaninVec(pObj) != nil {
            Vec_PtrFree( Ivy_ObjFaninVec(pObj) ); }
        pObj.pNextFan0 = pObj.pNextFan1 = nil;
        pObj.pEquiv = nil;
    }
    }
    // remove dangling nodes 
    Ivy_ManCleanup( p.pManFraig );
    // clean up the class marks
    // inlining of macro Ivy_FraigForEachEquivClass
    for  pObj = p.lClasses.pHead;                                                     pObj;                                                             pObj = Ivy_ObjEquivListNext(pObj)  {
        pObj.fMarkA = 0;
    }
}

/**Function*************************************************************

  Synopsis    [Performs fraiging for one node.]

  Description [Returns the fraiged node.]
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Ivy_FraigAnd(p *Ivy_FraigMan_t, pObjOld *Ivy_Obj_t) *Ivy_Obj_t { 
    var pObjNew, pFanin0New, pFanin1New, pObjReprNew *Ivy_Obj_t;
    var RetValue int32;
    // get the fraiged fanins
    pFanin0New = Ivy_ObjChild0Equiv(pObjOld);
    pFanin1New = Ivy_ObjChild1Equiv(pObjOld);
    // get the candidate fraig node
    pObjNew = Ivy_And( p.pManFraig, pFanin0New, pFanin1New );
    // get representative of this class
    if Ivy_ObjClassNodeRepr(pObjOld) == nil || // this is a unique node
         (!p.pParams.fDoSparse && Ivy_ObjClassNodeRepr(pObjOld) == p.pManAig.pConst1) { // this is a sparse node
    
//        assert( Ivy_Regular(pFanin0New) != Ivy_Regular(pFanin1New) );
//        assert( pObjNew != Ivy_Regular(pFanin0New) );
//        assert( pObjNew != Ivy_Regular(pFanin1New) );
        return pObjNew;
    }
    // get the fraiged representative
    pObjReprNew = Ivy_ObjFraig(Ivy_ObjClassNodeRepr(pObjOld));
    // if the fraiged nodes are the same return
    if Ivy_Regular(pObjNew) == Ivy_Regular(pObjReprNew) {
        return pObjNew; }
    assert(Ivy_Regular(pObjNew) != Ivy_ManConst1(p.pManFraig), "Ivy_Regular(pObjNew) != Ivy_ManConst1(p.pManFraig)");
//    printf( "Node = %d. Repr = %d.\n", pObjOld->Id, Ivy_ObjClassNodeRepr(pObjOld)->Id );

    // they are different (the counter-example is in p->pPatWords)
    RetValue = Ivy_FraigNodesAreEquiv( p, Ivy_Regular(pObjReprNew), Ivy_Regular(pObjNew) );
    if RetValue == 1 {  // proved equivalent
    
        // mark the class as proved
        if Ivy_ObjClassNodeNext(pObjOld) == nil {
            Ivy_ObjClassNodeRepr(pObjOld).fMarkA = 1; }
        return Ivy_NotCond( pObjReprNew, pObjOld.fPhase ^ Ivy_ObjClassNodeRepr(pObjOld).fPhase );
    }
    if RetValue == -1 { // failed
        return pObjNew; }
    // simulate the counter-example and return the new node
    Ivy_FraigResimulate( p );
    return pObjNew;
}

/**Function*************************************************************

  Synopsis    [Prints variable activity.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Ivy_FraigPrintActivity(p *Ivy_FraigMan_t) {
    var i int32;
    for  i = 0; i < p.nSatVars; i++  {
        fmt.Printf("%d %d  ", i, int32(p.pSat.activity[i]));
    }
    fmt.Printf("\n");
}

/**Function*************************************************************

  Synopsis    [Runs equivalence test for the two nodes.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Ivy_FraigNodesAreEquiv(p *Ivy_FraigMan_t, pOld *Ivy_Obj_t, pNew *Ivy_Obj_t) int32 {
    var pLits [4]int32;
    var RetValue int32;
    var RetValue1 int32;
    var nBTLimit int32;
    var clk abctime; //, clk2 = Abc_Clock();

    // make sure the nodes are not complemented
    assert(!Ivy_IsComplement(pNew), "!Ivy_IsComplement(pNew)");
    assert(!Ivy_IsComplement(pOld), "!Ivy_IsComplement(pOld)");
    assert(pNew != pOld, "pNew != pOld");

    // if at least one of the nodes is a failed node, perform adjustments:
    // if the backtrack limit is small, simply skip this node
    // if the backtrack limit is > 10, take the quare root of the limit
    nBTLimit = p.pParams.nBTLimitNode;
    if nBTLimit > 0 && (pOld.fFailTfo || pNew.fFailTfo) {
        p.nSatFails++;
        // fail immediately
//        return -1;
        if nBTLimit <= 10 {
            return -1; }
        nBTLimit = int32(math.Pow(float64(nBTLimit), float64(0.7)));
    }
    p.nSatCalls++;

    // make sure the solver is allocated and has enough variables
    if p.pSat == nil {
        p.pSat = sat_solver_new();
        sat_solver_setnvars( p.pSat, 1000 );
        // inlining of macro ABC_CALLOC
        p.pSat.factors = ((double *) calloc((size_t)(p.pSat.cap), sizeof(double)));
        p.nSatVars = 1;
        // var 0 is reserved for const1 node - add the clause
//        pLits[0] = toLit( 0 );
//        sat_solver_addclause( p->pSat, pLits, pLits + 1 );
    }

    // if the nodes do not have SAT variables, allocate them
    Ivy_FraigNodeAddToSolver( p, pOld, pNew );

    // prepare variable activity
    Ivy_FraigSetActivityFactors( p, pOld, pNew ); 

    // solve under assumptions
    // A = 1; B = 0     OR     A = 1; B = 1 
clk = Abc_Clock();
    pLits[0] = toLitCond( Ivy_ObjSatNum(pOld), 0 );
    pLits[1] = toLitCond( Ivy_ObjSatNum(pNew), pOld.fPhase == pNew.fPhase );
//Sat_SolverWriteDimacs( p->pSat, "temp.cnf", pLits, pLits + 2, 1 );
    RetValue1 = sat_solver_solve( p.pSat, pLits, pLits + 2, 
        (ABC_INT64_T)nBTLimit, (ABC_INT64_T)0, 
        p.nBTLimitGlobal, p.nInsLimitGlobal );
p.timeSat += Abc_Clock() - clk;
    if RetValue1 == l_False {
p.timeSatUnsat += Abc_Clock() - clk;
        pLits[0] = lit_neg( pLits[0] );
        pLits[1] = lit_neg( pLits[1] );
        RetValue = sat_solver_addclause( p.pSat, pLits, pLits + 2 );
        assert((RetValue != 0), "RetValue");
        // continue solving the other implication
        p.nSatCallsUnsat++;
    
    } else if RetValue1 == l_True {
p.timeSatSat += Abc_Clock() - clk;
        Ivy_FraigSavePattern( p );
        p.nSatCallsSat++;
        return 0;
    
    } else { // if ( RetValue1 == l_Undef )
    
p.timeSatFail += Abc_Clock() - clk;
/*
        if ( nBTLimit > 1000 )
        {
            RetValue = Ivy_FraigCheckCone( p, p->pManFraig, pOld, pNew, nBTLimit );
            if ( RetValue != -1 )
                return RetValue;
        }
*/
        // mark the node as the failed node
        if pOld != p.pManFraig.pConst1 { 
            pOld.fFailTfo = 1; }
        pNew.fFailTfo = 1;
        p.nSatFailsReal++;
        return -1;
    }

    // if the old node was constant 0, we already know the answer
    if pOld == p.pManFraig.pConst1 {
        p.nSatProof++;
        return 1;
    }

    // solve under assumptions
    // A = 0; B = 1     OR     A = 0; B = 0 
clk = Abc_Clock();
    pLits[0] = toLitCond( Ivy_ObjSatNum(pOld), 1 );
    pLits[1] = toLitCond( Ivy_ObjSatNum(pNew), pOld.fPhase ^ pNew.fPhase );
    RetValue1 = sat_solver_solve( p.pSat, pLits, pLits + 2, 
        (ABC_INT64_T)nBTLimit, (ABC_INT64_T)0, 
        p.nBTLimitGlobal, p.nInsLimitGlobal );
p.timeSat += Abc_Clock() - clk;
    if RetValue1 == l_False {
p.timeSatUnsat += Abc_Clock() - clk;
        pLits[0] = lit_neg( pLits[0] );
        pLits[1] = lit_neg( pLits[1] );
        RetValue = sat_solver_addclause( p.pSat, pLits, pLits + 2 );
        assert((RetValue != 0), "RetValue");
        p.nSatCallsUnsat++;
    
    } else if RetValue1 == l_True {
p.timeSatSat += Abc_Clock() - clk;
        Ivy_FraigSavePattern( p );
        p.nSatCallsSat++;
        return 0;
    
    } else { // if ( RetValue1 == l_Undef )
    
p.timeSatFail += Abc_Clock() - clk;
/*
        if ( nBTLimit > 1000 )
        {
            RetValue = Ivy_FraigCheckCone( p, p->pManFraig, pOld, pNew, nBTLimit );
            if ( RetValue != -1 )
                return RetValue;
        }
*/
        // mark the node as the failed node
        pOld.fFailTfo = 1;
        pNew.fFailTfo = 1;
        p.nSatFailsReal++;
        return -1;
    }
/*
    // check BDD proof
    {
        int RetVal;
        ABC_PRT( "Sat", Abc_Clock() - clk2 );
        clk2 = Abc_Clock();
        RetVal = Ivy_FraigNodesAreEquivBdd( pOld, pNew );
//        printf( "%d ", RetVal );
        assert( RetVal );
        ABC_PRT( "Bdd", Abc_Clock() - clk2 );
        printf( "\n" );
    }
*/
    // return SAT proof
    p.nSatProof++;
    return 1;
}

/**Function*************************************************************

  Synopsis    [Runs equivalence test for one node.]

  Description [Returns the fraiged node.]
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Ivy_FraigNodeIsConst(p *Ivy_FraigMan_t, pNew *Ivy_Obj_t) int32 {
    var pLits [2]int32;
    var RetValue1 int32;
    var clk abctime;
    var RetValue int32;

    // make sure the nodes are not complemented
    assert(!Ivy_IsComplement(pNew), "!Ivy_IsComplement(pNew)");
    assert(pNew != p.pManFraig.pConst1, "pNew != p.pManFraig.pConst1");
    p.nSatCalls++;

    // make sure the solver is allocated and has enough variables
    if p.pSat == nil {
        p.pSat = sat_solver_new();
        sat_solver_setnvars( p.pSat, 1000 );
        // inlining of macro ABC_CALLOC
        p.pSat.factors = ((double *) calloc((size_t)(p.pSat.cap), sizeof(double)));
        p.nSatVars = 1;
        // var 0 is reserved for const1 node - add the clause
//        pLits[0] = toLit( 0 );
//        sat_solver_addclause( p->pSat, pLits, pLits + 1 );
    }

    // if the nodes do not have SAT variables, allocate them
    Ivy_FraigNodeAddToSolver( p, nil, pNew );

    // prepare variable activity
    Ivy_FraigSetActivityFactors( p, nil, pNew ); 

    // solve under assumptions
clk = Abc_Clock();
    pLits[0] = toLitCond( Ivy_ObjSatNum(pNew), pNew.fPhase );
    RetValue1 = sat_solver_solve( p.pSat, pLits, pLits + 1, 
        (ABC_INT64_T)p.pParams.nBTLimitMiter, (ABC_INT64_T)0, 
        p.nBTLimitGlobal, p.nInsLimitGlobal );
p.timeSat += Abc_Clock() - clk;
    if RetValue1 == l_False {
p.timeSatUnsat += Abc_Clock() - clk;
        pLits[0] = lit_neg( pLits[0] );
        RetValue = sat_solver_addclause( p.pSat, pLits, pLits + 1 );
        assert((RetValue != 0), "RetValue");
        // continue solving the other implication
        p.nSatCallsUnsat++;
    
    } else if RetValue1 == l_True {
p.timeSatSat += Abc_Clock() - clk;
        if p.pPatWords {
            Ivy_FraigSavePattern( p ); }
        p.nSatCallsSat++;
        return 0;
    
    } else { // if ( RetValue1 == l_Undef )
    
p.timeSatFail += Abc_Clock() - clk;
/*
        if ( p->pParams->nBTLimitMiter > 1000 )
        {
            RetValue = Ivy_FraigCheckCone( p, p->pManFraig, p->pManFraig->pConst1, pNew, p->pParams->nBTLimitMiter );
            if ( RetValue != -1 )
                return RetValue;
        }
*/
        // mark the node as the failed node
        pNew.fFailTfo = 1;
        p.nSatFailsReal++;
        return -1;
    }

    // return SAT proof
    p.nSatProof++;
    return 1;
}

/**Function*************************************************************

  Synopsis    [Addes clauses to the solver.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Ivy_FraigAddClausesMux(p *Ivy_FraigMan_t, pNode *Ivy_Obj_t) {
    var pNodeI, pNodeT, pNodeE *Ivy_Obj_t;
    var pLits [4]int32;
    var RetValue int32;
    var VarF int32;
    var VarI int32;
    var VarT int32;
    var VarE int32;
    var fCompT int32;
    var fCompE int32;

    assert(!Ivy_IsComplement( pNode ), "!Ivy_IsComplement( pNode )");
    assert(Ivy_ObjIsMuxType( pNode ), "Ivy_ObjIsMuxType( pNode )");
    // get nodes (I = if, T = then, E = else)
    pNodeI = Ivy_ObjRecognizeMux( pNode, &pNodeT, &pNodeE );
    // get the variable numbers
    VarF = Ivy_ObjSatNum(pNode);
    VarI = Ivy_ObjSatNum(pNodeI);
    VarT = Ivy_ObjSatNum(Ivy_Regular(pNodeT));
    VarE = Ivy_ObjSatNum(Ivy_Regular(pNodeE));
    // get the complementation flags
    fCompT = Ivy_IsComplement(pNodeT);
    fCompE = Ivy_IsComplement(pNodeE);

    // f = ITE(i, t, e)

    // i' + t' + f
    // i' + t  + f'
    // i  + e' + f
    // i  + e  + f'

    // create four clauses
    pLits[0] = toLitCond(VarI, 1);
    pLits[1] = toLitCond(VarT, 1^fCompT);
    pLits[2] = toLitCond(VarF, 0);
    RetValue = sat_solver_addclause( p.pSat, pLits, pLits + 3 );
    assert((RetValue != 0), "RetValue");
    pLits[0] = toLitCond(VarI, 1);
    pLits[1] = toLitCond(VarT, 0^fCompT);
    pLits[2] = toLitCond(VarF, 1);
    RetValue = sat_solver_addclause( p.pSat, pLits, pLits + 3 );
    assert((RetValue != 0), "RetValue");
    pLits[0] = toLitCond(VarI, 0);
    pLits[1] = toLitCond(VarE, 1^fCompE);
    pLits[2] = toLitCond(VarF, 0);
    RetValue = sat_solver_addclause( p.pSat, pLits, pLits + 3 );
    assert((RetValue != 0), "RetValue");
    pLits[0] = toLitCond(VarI, 0);
    pLits[1] = toLitCond(VarE, 0^fCompE);
    pLits[2] = toLitCond(VarF, 1);
    RetValue = sat_solver_addclause( p.pSat, pLits, pLits + 3 );
    assert((RetValue != 0), "RetValue");

    // two additional clauses
    // t' & e' -> f'
    // t  & e  -> f 

    // t  + e   + f'
    // t' + e'  + f 

    if VarT == VarE {
//        assert( fCompT == !fCompE );
        return;
    }

    pLits[0] = toLitCond(VarT, 0^fCompT);
    pLits[1] = toLitCond(VarE, 0^fCompE);
    pLits[2] = toLitCond(VarF, 1);
    RetValue = sat_solver_addclause( p.pSat, pLits, pLits + 3 );
    assert((RetValue != 0), "RetValue");
    pLits[0] = toLitCond(VarT, 1^fCompT);
    pLits[1] = toLitCond(VarE, 1^fCompE);
    pLits[2] = toLitCond(VarF, 0);
    RetValue = sat_solver_addclause( p.pSat, pLits, pLits + 3 );
    assert((RetValue != 0), "RetValue");
}

/**Function*************************************************************

  Synopsis    [Addes clauses to the solver.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Ivy_FraigAddClausesSuper(p *Ivy_FraigMan_t, pNode *Ivy_Obj_t, vSuper *Vec_Ptr_t) {
    var pFanin *Ivy_Obj_t;
    var pLits *int32;
    var nLits int32;
    var RetValue int32;
    var i int32;
    assert(!Ivy_IsComplement(pNode), "!Ivy_IsComplement(pNode)");
    assert(Ivy_ObjIsNode( pNode ), "Ivy_ObjIsNode( pNode )");
    // create storage for literals
    nLits = Vec_PtrSize(vSuper) + 1;
    // inlining of macro ABC_ALLOC
    pLits = ((int *) malloc(sizeof(int) * (size_t)(nLits)));
    // suppose AND-gate is A & B = C
    // add !A => !C   or   A + !C
    // inlining of macro Vec_PtrForEachEntry
    for  i = 0; (i < Vec_PtrSize(vSuper)) && (((pFanin) = (Ivy_Obj_t *)Vec_PtrEntry(vSuper, i)), 1); i++  {
        pLits[0] = toLitCond(Ivy_ObjSatNum(Ivy_Regular(pFanin)), Ivy_IsComplement(pFanin));
        pLits[1] = toLitCond(Ivy_ObjSatNum(pNode), 1);
        RetValue = sat_solver_addclause( p.pSat, pLits, pLits + 2 );
        assert((RetValue != 0), "RetValue");
    }
    // add A & B => C   or   !A + !B + C
    // inlining of macro Vec_PtrForEachEntry
    for  i = 0; (i < Vec_PtrSize(vSuper)) && (((pFanin) = (Ivy_Obj_t *)Vec_PtrEntry(vSuper, i)), 1); i++  {
        pLits[i] = toLitCond(Ivy_ObjSatNum(Ivy_Regular(pFanin)), !Ivy_IsComplement(pFanin));
    }
    pLits[nLits-1] = toLitCond(Ivy_ObjSatNum(pNode), 0);
    RetValue = sat_solver_addclause( p.pSat, pLits, pLits + nLits );
    assert((RetValue != 0), "RetValue");
    // inlining of macro ABC_FREE
    ((pLits) ? (free((char *) (pLits)), (pLits) = 0) : 0);
}

/**Function*************************************************************

  Synopsis    [Collects the supergate.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Ivy_FraigCollectSuper_rec(pObj *Ivy_Obj_t, vSuper *Vec_Ptr_t, fFirst int32, fUseMuxes int32) {
    // if the new node is complemented or a PI, another gate begins
    if Ivy_IsComplement(pObj) || Ivy_ObjIsPi(pObj) || (!(fFirst != 0) && Ivy_ObjRefs(pObj) > 1) || 
         ((fUseMuxes != 0) && Ivy_ObjIsMuxType(pObj)) {
        Vec_PtrPushUnique( vSuper, pObj );
        return;
    }
    // go through the branches
    Ivy_FraigCollectSuper_rec( Ivy_ObjChild0(pObj), vSuper, 0, fUseMuxes );
    Ivy_FraigCollectSuper_rec( Ivy_ObjChild1(pObj), vSuper, 0, fUseMuxes );
}

/**Function*************************************************************

  Synopsis    [Collects the supergate.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Ivy_FraigCollectSuper(pObj *Ivy_Obj_t, fUseMuxes int32) *Vec_Ptr_t {
    var vSuper *Vec_Ptr_t;
    assert(!Ivy_IsComplement(pObj), "!Ivy_IsComplement(pObj)");
    assert(!Ivy_ObjIsPi(pObj), "!Ivy_ObjIsPi(pObj)");
    vSuper = Vec_PtrAlloc( 4 );
    Ivy_FraigCollectSuper_rec( pObj, vSuper, 1, fUseMuxes );
    return vSuper;
}

/**Function*************************************************************

  Synopsis    [Updates the solver clause database.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Ivy_FraigObjAddToFrontier(p *Ivy_FraigMan_t, pObj *Ivy_Obj_t, vFrontier *Vec_Ptr_t) {
    assert(!Ivy_IsComplement(pObj), "!Ivy_IsComplement(pObj)");
    if Ivy_ObjSatNum(pObj) != 0 {
        return; }
    assert(Ivy_ObjSatNum(pObj) == 0, "Ivy_ObjSatNum(pObj) == 0");
    assert(Ivy_ObjFaninVec(pObj) == nil, "Ivy_ObjFaninVec(pObj) == nil");
    if Ivy_ObjIsConst1(pObj) {
        return; }
//printf( "Assigning node %d number %d\n", pObj->Id, p->nSatVars );
    Ivy_ObjSetSatNum( pObj, p.nSatVars++ );
    if Ivy_ObjIsNode(pObj) {
        Vec_PtrPush( vFrontier, pObj ); }
}

/**Function*************************************************************

  Synopsis    [Updates the solver clause database.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Ivy_FraigNodeAddToSolver(p *Ivy_FraigMan_t, pOld *Ivy_Obj_t, pNew *Ivy_Obj_t) { 
    var vFrontier, vFanins *Vec_Ptr_t;
    var pNode, pFanin *Ivy_Obj_t;
    var i int32;
    var k int32;
    var fUseMuxes int32 = 1;
    assert((pOld != nil) || (pNew != nil), "pOld || pNew");
    // quit if CNF is ready
    if (!(pOld != nil) || (Ivy_ObjFaninVec(pOld) != nil)) && (!(pNew != nil) || (Ivy_ObjFaninVec(pNew) != nil)) {
        return; }
    // start the frontier
    vFrontier = Vec_PtrAlloc( 100 );
    if pOld != nil { Ivy_FraigObjAddToFrontier( p, pOld, vFrontier ); }
    if pNew != nil { Ivy_FraigObjAddToFrontier( p, pNew, vFrontier ); }
    // explore nodes in the frontier
    // inlining of macro Vec_PtrForEachEntry
    for  i = 0; (i < Vec_PtrSize(vFrontier)) && (((pNode) = (Ivy_Obj_t *)Vec_PtrEntry(vFrontier, i)), 1); i++  {
        // create the supergate
        assert((Ivy_ObjSatNum(pNode) != 0), "Ivy_ObjSatNum(pNode)");
        assert(Ivy_ObjFaninVec(pNode) == nil, "Ivy_ObjFaninVec(pNode) == nil");
        if (fUseMuxes != 0) && Ivy_ObjIsMuxType(pNode) {
            vFanins = Vec_PtrAlloc( 4 );
            Vec_PtrPushUnique( vFanins, Ivy_ObjFanin0( Ivy_ObjFanin0(pNode) ) );
            Vec_PtrPushUnique( vFanins, Ivy_ObjFanin0( Ivy_ObjFanin1(pNode) ) );
            Vec_PtrPushUnique( vFanins, Ivy_ObjFanin1( Ivy_ObjFanin0(pNode) ) );
            Vec_PtrPushUnique( vFanins, Ivy_ObjFanin1( Ivy_ObjFanin1(pNode) ) );
            // inlining of macro Vec_PtrForEachEntry
            for  k = 0; (k < Vec_PtrSize(vFanins)) && (((pFanin) = (Ivy_Obj_t *)Vec_PtrEntry(vFanins, k)), 1); k++  {
                Ivy_FraigObjAddToFrontier( p, Ivy_Regular(pFanin), vFrontier );
            }
            Ivy_FraigAddClausesMux( p, pNode );
        
        } else {
            vFanins = Ivy_FraigCollectSuper( pNode, fUseMuxes );
            // inlining of macro Vec_PtrForEachEntry
            for  k = 0; (k < Vec_PtrSize(vFanins)) && (((pFanin) = (Ivy_Obj_t *)Vec_PtrEntry(vFanins, k)), 1); k++  {
                Ivy_FraigObjAddToFrontier( p, Ivy_Regular(pFanin), vFrontier );
            }
            Ivy_FraigAddClausesSuper( p, pNode, vFanins );
        }
        assert(Vec_PtrSize(vFanins) > 1, "Vec_PtrSize(vFanins) > 1");
        Ivy_ObjSetFaninVec( pNode, vFanins );
    }
    Vec_PtrFree( vFrontier );
    sat_solver_simplify( p.pSat );
}

/**Function*************************************************************

  Synopsis    [Sets variable activities in the cone.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Ivy_FraigSetActivityFactors_rec(p *Ivy_FraigMan_t, pObj *Ivy_Obj_t, LevelMin int32, LevelMax int32) int32 {
    var vFanins *Vec_Ptr_t;
    var pFanin *Ivy_Obj_t;
    var i, Counter int32;
    assert(!Ivy_IsComplement(pObj), "!Ivy_IsComplement(pObj)");
    assert((Ivy_ObjSatNum(pObj) != 0), "Ivy_ObjSatNum(pObj)");
    // skip visited variables
    if Ivy_ObjIsTravIdCurrent(p.pManFraig, pObj) {
        return 0; }
    Ivy_ObjSetTravIdCurrent(p.pManFraig, pObj);
    // add the PI to the list
    if pObj.Level <= uint32(LevelMin) || Ivy_ObjIsPi(pObj) {
        return 0; }
    // set the factor of this variable
    // (LevelMax-LevelMin) / (pObj->Level-LevelMin) = p->pParams->dActConeBumpMax / ThisBump
    p.pSat.factors[Ivy_ObjSatNum(pObj)] = p.pParams.dActConeBumpMax * (pObj.Level - LevelMin)/(LevelMax - LevelMin);
    veci_push(&p.pSat.act_vars, Ivy_ObjSatNum(pObj));
    // explore the fanins
    vFanins = Ivy_ObjFaninVec( pObj );
    // inlining of macro Vec_PtrForEachEntry
    for  i = 0; (i < Vec_PtrSize(vFanins)) && (((pFanin) = (Ivy_Obj_t *)Vec_PtrEntry(vFanins, i)), 1); i++  {
        Counter += Ivy_FraigSetActivityFactors_rec( p, Ivy_Regular(pFanin), LevelMin, LevelMax );
    }
    return 1 + Counter;
}

/**Function*************************************************************

  Synopsis    [Sets variable activities in the cone.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Ivy_FraigSetActivityFactors(p *Ivy_FraigMan_t, pOld *Ivy_Obj_t, pNew *Ivy_Obj_t) int32 {
    var LevelMin, LevelMax int32;
    var clk abctime;
    assert((pOld != nil) || (pNew != nil), "pOld || pNew");
clk = Abc_Clock(); 
    // reset the active variables
    veci_resize(&p.pSat.act_vars, 0);
    // prepare for traversal
    Ivy_ManIncrementTravId( p.pManFraig );
    // determine the min and max level to visit
    assert(p.pParams.dActConeRatio > 0 && p.pParams.dActConeRatio < 1, "p.pParams.dActConeRatio > 0 && p.pParams.dActConeRatio < 1");
    // inlining of macro IVY_MAX
    
    if ((pNew ? pNew.Level : 0)) > ((pOld ? pOld.Level : 0)) {
        if pNew != nil {
            LevelMax = pNew.Level;
        } else {
            LevelMax = 0;
        }
    } else {
        if pOld != nil {
            LevelMax = pOld.Level;
        } else {
            LevelMax = 0;
        }
    }
    LevelMin = int32((LevelMax * (1.0 - p.pParams.dActConeRatio)));
    // traverse
    if (pOld != nil) && !Ivy_ObjIsConst1(pOld) {
        Ivy_FraigSetActivityFactors_rec( p, pOld, LevelMin, LevelMax ); }
    if (pNew != nil) && !Ivy_ObjIsConst1(pNew) {
        Ivy_FraigSetActivityFactors_rec( p, pNew, LevelMin, LevelMax ); }
//Ivy_FraigPrintActivity( p );
p.timeTrav += Abc_Clock() - clk;
    return 1;
}

//ABC_NAMESPACE_IMPL_END

#ifdef ABC_USE_CUDD
#include "bdd/cudd/cuddInt.h"
#endif

//ABC_NAMESPACE_IMPL_START

#ifdef ABC_USE_CUDD

/**Function*************************************************************

  Synopsis    [Checks equivalence using BDDs.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
DdNode * Ivy_FraigNodesAreEquivBdd_int( DdManager * dd, DdNode * bFunc, Vec_Ptr_t * vFront, int Level )
{
    DdNode ** pFuncs;
    DdNode * bFuncNew;
    Vec_Ptr_t * vTemp;
    Ivy_Obj_t * pObj, * pFanin;
    int i, NewSize;
    // create new frontier
    vTemp = Vec_PtrAlloc( 100 );
    Vec_PtrForEachEntry( Ivy_Obj_t *, vFront, pObj, i )
    {
        if ( (int)pObj->Level != Level )
        {
            pObj->fMarkB = 1;
            pObj->TravId = Vec_PtrSize(vTemp);
            Vec_PtrPush( vTemp, pObj );
            continue;
        }

        pFanin = Ivy_ObjFanin0(pObj);
        if ( pFanin->fMarkB == 0 )
        {
            pFanin->fMarkB = 1;
            pFanin->TravId = Vec_PtrSize(vTemp);
            Vec_PtrPush( vTemp, pFanin );
        }

        pFanin = Ivy_ObjFanin1(pObj);
        if ( pFanin->fMarkB == 0 )
        {
            pFanin->fMarkB = 1;
            pFanin->TravId = Vec_PtrSize(vTemp);
            Vec_PtrPush( vTemp, pFanin );
        }
    }
    // collect the permutation
    NewSize = IVY_MAX(dd->size, Vec_PtrSize(vTemp));
    pFuncs = ABC_ALLOC( DdNode *, NewSize );
    Vec_PtrForEachEntry( Ivy_Obj_t *, vFront, pObj, i )
    {
        if ( (int)pObj->Level != Level )
            pFuncs[i] = Cudd_bddIthVar( dd, pObj->TravId );
        else
            pFuncs[i] = Cudd_bddAnd( dd, 
                Cudd_NotCond( Cudd_bddIthVar(dd, Ivy_ObjFanin0(pObj)->TravId), Ivy_ObjFaninC0(pObj) ),
                Cudd_NotCond( Cudd_bddIthVar(dd, Ivy_ObjFanin1(pObj)->TravId), Ivy_ObjFaninC1(pObj) ) );
        Cudd_Ref( pFuncs[i] );
    }
    // add the remaining vars
    assert( NewSize == dd->size );
    for ( i = Vec_PtrSize(vFront); i < dd->size; i++ )
    {
        pFuncs[i] = Cudd_bddIthVar( dd, i );
        Cudd_Ref( pFuncs[i] );
    }

    // create new
    bFuncNew = Cudd_bddVectorCompose( dd, bFunc, pFuncs ); Cudd_Ref( bFuncNew );
    // clean trav Id
    Vec_PtrForEachEntry( Ivy_Obj_t *, vTemp, pObj, i )
    {
        pObj->fMarkB = 0;
        pObj->TravId = 0;
    }
    // deref
    for ( i = 0; i < dd->size; i++ )
        Cudd_RecursiveDeref( dd, pFuncs[i] );
    ABC_FREE( pFuncs );

    ABC_FREE( vFront->pArray );
    *vFront = *vTemp;

    vTemp->nCap = vTemp->nSize = 0;
    vTemp->pArray = NULL;
    Vec_PtrFree( vTemp );

    Cudd_Deref( bFuncNew );
    return bFuncNew;
}

/**Function*************************************************************

  Synopsis    [Checks equivalence using BDDs.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
int Ivy_FraigNodesAreEquivBdd( Ivy_Obj_t * pObj1, Ivy_Obj_t * pObj2 )
{
    static DdManager * dd = NULL;
    DdNode * bFunc, * bTemp;
    Vec_Ptr_t * vFront;
    Ivy_Obj_t * pObj;
    int i, RetValue, Iter, Level;
    // start the manager
    if ( dd == NULL )
        dd = Cudd_Init( 50, 0, CUDD_UNIQUE_SLOTS, CUDD_CACHE_SLOTS, 0 );
    // create front
    vFront = Vec_PtrAlloc( 100 );
    Vec_PtrPush( vFront, pObj1 );
    Vec_PtrPush( vFront, pObj2 );
    // get the function
    bFunc = Cudd_bddXor( dd, Cudd_bddIthVar(dd,0), Cudd_bddIthVar(dd,1) );  Cudd_Ref( bFunc );
    bFunc = Cudd_NotCond( bFunc, pObj1->fPhase != pObj2->fPhase );
    // try running BDDs
    for ( Iter = 0; ; Iter++ )
    {
        // find max level
        Level = 0;
        Vec_PtrForEachEntry( Ivy_Obj_t *, vFront, pObj, i )
            if ( Level < (int)pObj->Level )
                Level = (int)pObj->Level;
        if ( Level == 0 )
            break;            
        bFunc = Ivy_FraigNodesAreEquivBdd_int( dd, bTemp = bFunc, vFront, Level ); Cudd_Ref( bFunc );
        Cudd_RecursiveDeref( dd, bTemp );
        if ( bFunc == Cudd_ReadLogicZero(dd) ) // proved
            {printf( "%d", Iter ); break;}
        if ( Cudd_DagSize(bFunc) > 1000 )
            {printf( "b" ); break;}
        if ( dd->size > 120 )
            {printf( "s" ); break;}
        if ( Iter > 50 )
            {printf( "i" ); break;}
    }
    if ( bFunc == Cudd_ReadLogicZero(dd) ) // unsat
        RetValue = 1;
    else if ( Level == 0 ) // sat
        RetValue = 0;
    else 
        RetValue = -1; // spaceout/timeout
    Cudd_RecursiveDeref( dd, bFunc );
    Vec_PtrFree( vFront );
    return RetValue;
}
#endif

//ABC_NAMESPACE_IMPL_END

//#include "aig/aig/aig.h"

//ABC_NAMESPACE_IMPL_START


/**Function*************************************************************

  Synopsis    [Computes truth table of the cut.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Ivy_FraigExtractCone_rec(p *Ivy_Man_t, pNode *Ivy_Obj_t, vLeaves *Vec_Int_t, vNodes *Vec_Int_t) {
    if pNode.fMarkB {
        return; }
    pNode.fMarkB = 1;
    if Ivy_ObjIsPi(pNode) {
        Vec_IntPush( vLeaves, pNode.Id );
        return;
    }
    assert(Ivy_ObjIsAnd(pNode), "Ivy_ObjIsAnd(pNode)");
    Ivy_FraigExtractCone_rec( p, Ivy_ObjFanin0(pNode), vLeaves, vNodes );
    Ivy_FraigExtractCone_rec( p, Ivy_ObjFanin1(pNode), vLeaves, vNodes );
    Vec_IntPush( vNodes, pNode.Id );
}

/**Function*************************************************************

  Synopsis    [Checks equivalence using BDDs.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Ivy_FraigExtractCone(p *Ivy_Man_t, pObj1 *Ivy_Obj_t, pObj2 *Ivy_Obj_t, vLeaves *Vec_Int_t) *Aig_Man_t {
    var pMan *Aig_Man_t;
    var pMiter *Aig_Obj_t;
    var vNodes *Vec_Int_t;
    var pObjIvy *Ivy_Obj_t;
    var i int32;
    // collect nodes
    vNodes  = Vec_IntAlloc( 100 );
    Ivy_ManConst1(p).fMarkB = 1;
    Ivy_FraigExtractCone_rec( p, pObj1, vLeaves, vNodes );
    Ivy_FraigExtractCone_rec( p, pObj2, vLeaves, vNodes );
    Ivy_ManConst1(p).fMarkB = 0;
    // create new manager
    pMan = Aig_ManStart( 1000 );
    Ivy_ManConst1(p).pEquiv = (Ivy_Obj_t *)Aig_ManConst1(pMan);
    Ivy_ManForEachNodeVec( p, vLeaves, pObjIvy, i )
    {
        pObjIvy.pEquiv = (Ivy_Obj_t *)Aig_ObjCreateCi( pMan );
        pObjIvy.fMarkB = 0;
    }
    // duplicate internal nodes
    Ivy_ManForEachNodeVec( p, vNodes, pObjIvy, i )
    {

        pObjIvy.pEquiv = (Ivy_Obj_t *)Aig_And( pMan, (Aig_Obj_t *)Ivy_ObjChild0Equiv(pObjIvy), (Aig_Obj_t *)Ivy_ObjChild1Equiv(pObjIvy) );
        pObjIvy.fMarkB = 0;

        pMiter = (Aig_Obj_t *)pObjIvy.pEquiv;
        assert(pMiter.fPhase == pObjIvy.fPhase, "pMiter.fPhase == pObjIvy.fPhase");
    }
    // create the PO
    pMiter = Aig_Exor( pMan, (Aig_Obj_t *)pObj1.pEquiv, (Aig_Obj_t *)pObj2.pEquiv );
    pMiter = Aig_NotCond( pMiter, Aig_Regular(pMiter).fPhase ^ Aig_IsComplement(pMiter) );

/*
printf( "Polarity = %d\n", pMiter->fPhase );
    if ( Ivy_ObjIsConst1(pObj1) || Ivy_ObjIsConst1(pObj2) )
    {
        pMiter = Aig_NotCond( pMiter, 1 );
printf( "***************\n" );
    }
*/
    pMiter = Aig_ObjCreateCo( pMan, pMiter );
//printf( "Polarity = %d\n", pMiter->fPhase );
    Aig_ManCleanup( pMan );
    Vec_IntFree( vNodes );
    return pMan;
}

/**Function*************************************************************

  Synopsis    [Checks equivalence using BDDs.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Ivy_FraigCheckCone(pGlo *Ivy_FraigMan_t, p *Ivy_Man_t, pObj1 *Ivy_Obj_t, pObj2 *Ivy_Obj_t, nConfLimit int32) int32 {
    extern int Fra_FraigSat( Aig_Man_t * pMan, ABC_INT64_T nConfLimit, ABC_INT64_T nInsLimit, int nLearnedStart, int nLearnedDelta, int nLearnedPerce, int fFlipBits, int fAndOuts, int fNewSolver, int fVerbose );
    var vLeaves *Vec_Int_t;
    var pMan *Aig_Man_t;
    var pObj *Aig_Obj_t;
    var i, RetValue int32;
    vLeaves  = Vec_IntAlloc( 100 );
    pMan     = Ivy_FraigExtractCone( p, pObj1, pObj2, vLeaves );
    RetValue = Fra_FraigSat( pMan, nConfLimit, 0, 0, 0, 0, 0, 0, 0, 1 ); 
    if RetValue == 0 {
        var Counter int32;
        memset( pGlo.pPatWords, 0, sizeof(unsigned) * pGlo.nPatWords );
        // inlining of macro Aig_ManForEachCi
        Vec_PtrForEachEntry( Aig_Obj_t *, pMan.vCis, pObj, i )
            if ((int *)pMan.pData)[i] {
                var iObjIvy int32  = Vec_IntEntry( vLeaves, i );
                assert(iObjIvy > 0 && iObjIvy <= Ivy_ManPiNum(p), "iObjIvy > 0 && iObjIvy <= Ivy_ManPiNum(p)");
                Ivy_InfoSetBit( pGlo.pPatWords, iObjIvy-1 );
//printf( "%d ", iObjIvy );
Counter++;
            }
        assert(Counter > 0, "Counter > 0");
    }
    Vec_IntFree( vLeaves );
    if RetValue == 1 {
        fmt.Printf("UNSAT\n");
    } else if RetValue == 0 {
        fmt.Printf("SAT\n");
    } else if RetValue == -1 {
        fmt.Printf("UNDEC\n"); }

//    p->pModel = (int *)pMan->pData, pMan2->pData = NULL;
    Aig_ManStop( pMan );
    return RetValue;
}

////////////////////////////////////////////////////////////////////////
///                       END OF FILE                                ///
////////////////////////////////////////////////////////////////////////


//ABC_NAMESPACE_IMPL_END

