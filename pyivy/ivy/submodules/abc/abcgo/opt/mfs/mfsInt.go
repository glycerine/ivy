package mfs

/**CFile****************************************************************

  FileName    [mfsInt.h]

  SystemName  [ABC: Logic synthesis and verification system.]

  PackageName [The good old minimization with complete don't-cares.]

  Synopsis    [Internal declarations.]

  Author      [Alan Mishchenko]
  
  Affiliation [UC Berkeley]

  Date        [Ver. 1.0. Started - June 20, 2005.]

  Revision    [$Id: mfsInt.h,v 1.00 2005/06/20 00:00:00 alanmi Exp $]

***********************************************************************/
 




////////////////////////////////////////////////////////////////////////
///                          INCLUDES                                ///
////////////////////////////////////////////////////////////////////////

//#include "base/abc/abc.h"
//#include "mfs.h"
//#include "aig/aig/aig.h"
//#include "sat/cnf/cnf.h"
//#include "sat/bsat/satSolver.h"
//#include "sat/bsat/satStore.h"
//#include "bool/bdc/bdc.h"
//#include "aig/gia/gia.h"

////////////////////////////////////////////////////////////////////////
///                         PARAMETERS                               ///
////////////////////////////////////////////////////////////////////////



//ABC_NAMESPACE_HEADER_START


const MFS_FANIN_MAX =   12

   
type Mfs_Man_t struct {
    // input data
    pPars          *Mfs_Par_t;
    pNtk          *Abc_Ntk_t;
    pCare          *Aig_Man_t;
    vSuppsInv          *Vec_Ptr_t;
    nFaninMax                 int32;
    // intermeditate data for the node
    vRoots          *Vec_Ptr_t;    // the roots of the window
    vSupp          *Vec_Ptr_t;     // the support of the window
    vNodes          *Vec_Ptr_t;    // the internal nodes of the window
    vDivs          *Vec_Ptr_t;     // the divisors of the node
    vDivLits          *Vec_Int_t;  // the SAT literals of divisor nodes
    vProjVarsCnf          *Vec_Int_t; // the projection variables
    vProjVarsSat          *Vec_Int_t; // the projection variables
    // intermediate simulation data
    vDivCexes          *Vec_Ptr_t; // the counter-example for dividors
    nDivWords                 int32; // the number of words
    nCexes                 int32;    // the numbe rof current counter-examples
    nSatCalls                 int32; 
    nSatCexes                 int32;
/*
    // intermediate AIG data
    Gia_Man_t *         pGia;      // replica of the AIG in the new package
//    Gia_Obj_t **        pSat2Gia;  // mapping of PO SAT var into internal GIA nodes
    Tas_Man_t *         pTas;      // the SAT solver
    Vec_Int_t *         vCex;      // the counter-example
    Vec_Ptr_t *         vGiaLits;  // literals given as assumptions
*/
    // used for bidecomposition
    vTruth          *Vec_Int_t;
    pManDec          *Bdc_Man_t;
    nNodesDec                 int32;
    nNodesGained                 int32;
    nNodesGainedLevel                 int32;
    // solving data
    pAigWin          *Aig_Man_t;   // window AIG with constraints
    pCnf          *Cnf_Dat_t;      // the CNF for the window
    pSat         *sat_solver;      // the SAT solver used 
    pMan          *Int_Man_t;      // interpolation manager;
    vMem          *Vec_Int_t;      // memory for intermediate SOPs
    vLevels          *Vec_Vec_t;   // levelized structure for updating
    vMfsFanins          *Vec_Ptr_t;   // the new set of fanins
    nTotConfLim                 int32; // total conflict limit
    nTotConfLevel                 int32; // total conflicts on this level
    // switching activity
    vProbs          *Vec_Int_t; 
    // the result of solving
    nFanins                 int32;   // the number of fanins
    nWords                 int32;    // the number of words
    nCares                 int32;    // the number of care minterms
    uCare [(MFS_FANIN_MAX<=5)?1:1<<(MFS_FANIN_MAX-5)]uint32;  // the computed care-set
    // performance statistics
    nTryRemoves                 int32; // number of fanin removals
    nTryResubs                 int32;  // number of resubstitutions
    nRemoves                 int32;    // number of fanin removals
    nResubs                 int32;     // number of resubstitutions
    nNodesTried                 int32;
    nNodesResub                 int32;
    nMintsCare                 int32;
    nMintsTotal                 int32;
    nNodesBad                 int32;
    nTotalDivs                 int32;
    nTimeOuts                 int32;
    nTimeOutsLevel                 int32;
    nDcMints                 int32;
    nMaxDivs                 int32;
    dTotalRatios              float64;
    // node/edge stats
    nTotalNodesBeg                 int32;
    nTotalNodesEnd                 int32;
    nTotalEdgesBeg                 int32;
    nTotalEdgesEnd                 int32;
    TotalSwitchingBeg               float32;
    TotalSwitchingEnd               float32;
    // statistics
    timeWin             abctime;
    timeDiv             abctime;
    timeAig             abctime;
    timeGia             abctime;
    timeCnf             abctime;
    timeSat             abctime;
    timeInt             abctime;
    timeTotal             abctime;
}

func Abc_MfsObjProb(p *Mfs_Man_t, pObj *Abc_Obj_t) float32 { if p.vProbs && pObj.Id < Vec_IntSize(p.vProbs) {
    return Abc_Int2Float(Vec_IntEntry(p.vProbs,pObj.Id));
} else {
    return 0.0;
} }

////////////////////////////////////////////////////////////////////////
///                         BASIC TYPES                              ///
////////////////////////////////////////////////////////////////////////

////////////////////////////////////////////////////////////////////////
///                      MACRO DEFINITIONS                           ///
////////////////////////////////////////////////////////////////////////

////////////////////////////////////////////////////////////////////////
///                    FUNCTION DECLARATIONS                         ///
////////////////////////////////////////////////////////////////////////

/*=== mfsDiv.c ==========================================================*/
// extern Vec_Ptr_t *      Abc_MfsComputeDivisors( Mfs_Man_t * p, Abc_Obj_t * pNode, int nLevDivMax );

/*=== mfsInter.c ==========================================================*/
// extern sat_solver *     Abc_MfsCreateSolverResub( Mfs_Man_t * p, int * pCands, int nCands, int fInvert );

// extern Hop_Obj_t *      Abc_NtkMfsInterplate( Mfs_Man_t * p, int * pCands, int nCands );

// extern int              Abc_NtkMfsInterplateEval( Mfs_Man_t * p, int * pCands, int nCands );

/*=== mfsMan.c ==========================================================*/
// extern Mfs_Man_t *      Mfs_ManAlloc( Mfs_Par_t * pPars );

// extern void             Mfs_ManStop( Mfs_Man_t * p );

// extern void             Mfs_ManClean( Mfs_Man_t * p );

/*=== mfsResub.c ==========================================================*/
// extern void             Abc_NtkMfsPrintResubStats( Mfs_Man_t * p );

// extern int              Abc_NtkMfsEdgeSwapEval( Mfs_Man_t * p, Abc_Obj_t * pNode );

// extern int              Abc_NtkMfsEdgePower( Mfs_Man_t * p, Abc_Obj_t * pNode );

// extern int              Abc_NtkMfsResubNode( Mfs_Man_t * p, Abc_Obj_t * pNode );

// extern int              Abc_NtkMfsResubNode2( Mfs_Man_t * p, Abc_Obj_t * pNode );

/*=== mfsSat.c ==========================================================*/
// extern int              Abc_NtkMfsSolveSat( Mfs_Man_t * p, Abc_Obj_t * pNode );

// extern int              Abc_NtkAddOneHotness( Mfs_Man_t * p );

/*=== mfsStrash.c ==========================================================*/
// extern Aig_Man_t *      Abc_NtkConstructAig( Mfs_Man_t * p, Abc_Obj_t * pNode );

// extern double           Abc_NtkConstraintRatio( Mfs_Man_t * p, Abc_Obj_t * pNode );

/*=== mfsWin.c ==========================================================*/
// extern Vec_Ptr_t *      Abc_MfsComputeRoots( Abc_Obj_t * pNode, int nWinTfoMax, int nFanoutLimit );


/*=== mfsGia.c ==========================================================*/
// extern void             Abc_NtkMfsConstructGia( Mfs_Man_t * p );

// extern void             Abc_NtkMfsDeconstructGia( Mfs_Man_t * p );

// extern int              Abc_NtkMfsTryResubOnceGia( Mfs_Man_t * p, int * pCands, int nCands );




//ABC_NAMESPACE_HEADER_END





////////////////////////////////////////////////////////////////////////
///                       END OF FILE                                ///
////////////////////////////////////////////////////////////////////////

