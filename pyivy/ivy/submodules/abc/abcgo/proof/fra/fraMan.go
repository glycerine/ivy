package fra

import (
	"fmt"
	"os"
	"time"
)

/**CFile****************************************************************

  FileName    [fraMan.c]

  SystemName  [ABC: Logic synthesis and verification system.]

  PackageName [New FRAIG package.]

  Synopsis    [Starts the FRAIG manager.]

  Author      [Alan Mishchenko]
  
  Affiliation [UC Berkeley]

  Date        [Ver. 1.0. Started - June 30, 2007.]

  Revision    [$Id: fraMan.c,v 1.00 2007/06/30 00:00:00 alanmi Exp $]

***********************************************************************/

//#include "fra.h"

//ABC_NAMESPACE_IMPL_START


////////////////////////////////////////////////////////////////////////
///                        DECLARATIONS                              ///
////////////////////////////////////////////////////////////////////////

////////////////////////////////////////////////////////////////////////
///                     FUNCTION DEFINITIONS                         ///
////////////////////////////////////////////////////////////////////////

/**Function*************************************************************

  Synopsis    [Sets the default solving parameters.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Fra_ParamsDefault(pPars *Fra_Par_t) {
    memset( pPars, 0, sizeof(Fra_Par_t) );
    pPars.nSimWords        =       32;  // the number of words in the simulation info
    pPars.dSimSatur        =    0.005;  // the ratio of refined classes when saturation is reached
    pPars.fPatScores       =        0;  // enables simulation pattern scoring
    pPars.MaxScore         =       25;  // max score after which resimulation is used
    pPars.fDoSparse        =        1;  // skips sparse functions
//    pPars->dActConeRatio    =     0.05;  // the ratio of cone to be bumped
//    pPars->dActConeBumpMax  =      5.0;  // the largest bump of activity
    pPars.dActConeRatio    =      0.3;  // the ratio of cone to be bumped
    pPars.dActConeBumpMax  =     10.0;  // the largest bump of activity
    pPars.nBTLimitNode     =      100;  // conflict limit at a node
    pPars.nBTLimitMiter    =   500000;  // conflict limit at an output
    pPars.nFramesK         =        0;  // the number of timeframes to unroll
    pPars.fConeBias        =        1;
    pPars.fRewrite         =        0;
}

/**Function*************************************************************

  Synopsis    [Sets the default solving parameters.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Fra_ParamsDefaultSeq(pPars *Fra_Par_t) {
    memset( pPars, 0, sizeof(Fra_Par_t) );
    pPars.nSimWords        =        1;  // the number of words in the simulation info
    pPars.dSimSatur        =    0.005;  // the ratio of refined classes when saturation is reached
    pPars.fPatScores       =        0;  // enables simulation pattern scoring
    pPars.MaxScore         =       25;  // max score after which resimulation is used
    pPars.fDoSparse        =        1;  // skips sparse functions
    pPars.dActConeRatio    =      0.3;  // the ratio of cone to be bumped
    pPars.dActConeBumpMax  =     10.0;  // the largest bump of activity
    pPars.nBTLimitNode     = 10000000;  // conflict limit at a node
    pPars.nBTLimitMiter    =   500000;  // conflict limit at an output
    pPars.nFramesK         =        1;  // the number of timeframes to unroll
    pPars.fConeBias        =        0;
    pPars.fRewrite         =        0;
    pPars.fLatchCorr       =        0;
} 

/**Function*************************************************************

  Synopsis    [Starts the fraiging manager.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Fra_ManStart(pManAig *Aig_Man_t, pPars *Fra_Par_t) *Fra_Man_t {
    var p *Fra_Man_t;
    var pObj *Aig_Obj_t;
    var i int32;
    // allocate the fraiging manager
    // inlining of macro ABC_ALLOC
    p = ((Fra_Man_t *) malloc(sizeof(Fra_Man_t) * (size_t)(1)));
    memset( p, 0, sizeof(Fra_Man_t) );
    p.pPars      = pPars;
    p.pManAig    = pManAig;
    p.nSizeAlloc = Aig_ManObjNumMax( pManAig );
    p.nFramesAll = pPars.nFramesK + 1;
    // allocate storage for sim pattern
    p.nPatWords  = Abc_BitWordNum( (Aig_ManCiNum(pManAig) - Aig_ManRegNum(pManAig)) * p.nFramesAll + Aig_ManRegNum(pManAig) );
    // inlining of macro ABC_ALLOC
    p.pPatWords  = ((unsigned *) malloc(sizeof(unsigned) * (size_t)(p.nPatWords))); 
    p.vPiVars    = Vec_PtrAlloc( 100 );
    // equivalence classes
    p.pCla       = Fra_ClassesStart( pManAig );
    // allocate other members
    // inlining of macro ABC_ALLOC
    p.pMemFraig  = ((Aig_Obj_t * *) malloc(sizeof(Aig_Obj_t *) * (size_t)(p.nSizeAlloc * p.nFramesAll)));
    memset( p.pMemFraig, 0, sizeof(Aig_Obj_t *) * p.nSizeAlloc * p.nFramesAll );
    // set random number generator
//    srand( 0xABCABC );
    Aig_ManRandom(1);
    // set the pointer to the manager
    // inlining of macro Aig_ManForEachObj
    // inlining of macro Vec_PtrForEachEntry
    for  i = 0; (i < Vec_PtrSize(p.pManAig.vObjs)) && (((pObj) = (Aig_Obj_t *)Vec_PtrEntry(p.pManAig.vObjs, i)), 1); i++  { if (pObj) == nil {} else {
        pObj.pData = p; }
    }
    return p;
}

/**Function*************************************************************

  Synopsis    [Starts the fraiging manager.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Fra_ManClean(p *Fra_Man_t, nNodesMax int32) {
    var i int32;
    // remove old arrays
    for  i = 0; i < p.nMemAlloc; i++  {
        if p.pMemFanins[i] && p.pMemFanins[i] != (void *)1 {
            Vec_PtrFree( p.pMemFanins[i] ); }
    }
    // realloc for the new size
    if p.nMemAlloc < nNodesMax {
        var nMemAllocNew int32  = nNodesMax + 5000;
        // inlining of macro ABC_REALLOC
        
        if p.pMemFanins {
            p.pMemFanins = ((Vec_Ptr_t * *) realloc((char *)(p.pMemFanins), sizeof(Vec_Ptr_t *) * (size_t)(nMemAllocNew)));
        } else {
            p.pMemFanins = ((Vec_Ptr_t * *) malloc(sizeof(Vec_Ptr_t *) * (size_t)(nMemAllocNew)));
        }
        // inlining of macro ABC_REALLOC
        
        if p.pMemSatNums {
            p.pMemSatNums = ((int *) realloc((char *)(p.pMemSatNums), sizeof(int) * (size_t)(nMemAllocNew)));
        } else {
            p.pMemSatNums = ((int *) malloc(sizeof(int) * (size_t)(nMemAllocNew)));
        }
        p.nMemAlloc = nMemAllocNew;
    }
    // prepare for the new run
    memset( p.pMemFanins, 0, sizeof(Vec_Ptr_t *) * p.nMemAlloc );
    memset( p.pMemSatNums, 0, sizeof(int) * p.nMemAlloc );
}

/**Function*************************************************************

  Synopsis    [Prepares the new manager to begin fraiging.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Fra_ManPrepareComb(p *Fra_Man_t) *Aig_Man_t {
    var pManFraig *Aig_Man_t;
    var pObj *Aig_Obj_t;
    var i int32;
    assert(p.pManFraig == nil, "p.pManFraig == nil");
    // start the fraig package
    pManFraig = Aig_ManStart( Aig_ManObjNumMax(p.pManAig) );
    pManFraig.pName = Abc_UtilStrsav( p.pManAig.pName );
    pManFraig.pSpec = Abc_UtilStrsav( p.pManAig.pSpec );
    pManFraig.nRegs    = p.pManAig.nRegs;
    pManFraig.nAsserts = p.pManAig.nAsserts;
    // set the pointers to the available fraig nodes
    Fra_ObjSetFraig( Aig_ManConst1(p.pManAig), 0, Aig_ManConst1(pManFraig) );
    // inlining of macro Aig_ManForEachCi
    // inlining of macro Vec_PtrForEachEntry
    for  i = 0; (i < Vec_PtrSize(p.pManAig.vCis)) && (((pObj) = (Aig_Obj_t *)Vec_PtrEntry(p.pManAig.vCis, i)), 1); i++  {
        Fra_ObjSetFraig( pObj, 0, Aig_ObjCreateCi(pManFraig) );
    }
    // set the pointers to the manager
    // inlining of macro Aig_ManForEachObj
    // inlining of macro Vec_PtrForEachEntry
    for  i = 0; (i < Vec_PtrSize(pManFraig.vObjs)) && (((pObj) = (Aig_Obj_t *)Vec_PtrEntry(pManFraig.vObjs, i)), 1); i++  { if (pObj) == nil {} else {
        pObj.pData = p; }
    }
    // allocate memory for mapping FRAIG nodes into SAT numbers and fanins
    p.nMemAlloc = p.nSizeAlloc;
    // inlining of macro ABC_ALLOC
    p.pMemFanins = ((Vec_Ptr_t * *) malloc(sizeof(Vec_Ptr_t *) * (size_t)(p.nMemAlloc)));
    memset( p.pMemFanins, 0, sizeof(Vec_Ptr_t *) * p.nMemAlloc );
    // inlining of macro ABC_ALLOC
    p.pMemSatNums = ((int *) malloc(sizeof(int) * (size_t)(p.nMemAlloc)));
    memset( p.pMemSatNums, 0, sizeof(int) * p.nMemAlloc );
    // make sure the satisfying assignment is node assigned
    assert(pManFraig.pData == nil, "pManFraig.pData == nil");
    return pManFraig;
}

/**Function*************************************************************

  Synopsis    [Finalizes the combinational miter after fraiging.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Fra_ManFinalizeComb(p *Fra_Man_t) {
    var pObj *Aig_Obj_t;
    var i int32;
    // add the POs
    // inlining of macro Aig_ManForEachCo
    // inlining of macro Vec_PtrForEachEntry
    for  i = 0; (i < Vec_PtrSize(p.pManAig.vCos)) && (((pObj) = (Aig_Obj_t *)Vec_PtrEntry(p.pManAig.vCos, i)), 1); i++  {
        Aig_ObjCreateCo( p.pManFraig, Fra_ObjChild0Fra(pObj,0) );
    }
    // postprocess
    Aig_ManCleanMarkB( p.pManFraig );
}


/**Function*************************************************************

  Synopsis    [Stops the fraiging manager.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Fra_ManStop(p *Fra_Man_t) {
    if p.pPars.fVerbose {
        Fra_ManPrint( p ); }
    // save mapping from original nodes into FRAIG nodes
    if p.pManAig {
        if p.pManAig.pObjCopies {
            // inlining of macro ABC_FREE
            ((p.pManAig.pObjCopies) ? (free((char *) (p.pManAig.pObjCopies)), (p.pManAig.pObjCopies) = 0) : 0); }
        p.pManAig.pObjCopies = p.pMemFraig;
        p.pMemFraig = nil;
    }
    Fra_ManClean( p, 0 );
    if p.vTimeouts { Vec_PtrFree( p.vTimeouts ); }
    if p.vPiVars {   Vec_PtrFree( p.vPiVars ); }
    if p.pSat {      sat_solver_delete( p.pSat ); }
    if p.pCla {     Fra_ClassesStop( p.pCla ); }
    if p.pSml {      Fra_SmlStop( p.pSml ); }
    if p.vCex {      Vec_IntFree( p.vCex ); }
    if p.vOneHots {  Vec_IntFree( p.vOneHots ); }
    // inlining of macro ABC_FREE
    ((p.pMemFraig) ? (free((char *) (p.pMemFraig)), (p.pMemFraig) = 0) : 0);
    // inlining of macro ABC_FREE
    ((p.pMemFanins) ? (free((char *) (p.pMemFanins)), (p.pMemFanins) = 0) : 0);
    // inlining of macro ABC_FREE
    ((p.pMemSatNums) ? (free((char *) (p.pMemSatNums)), (p.pMemSatNums) = 0) : 0);
    // inlining of macro ABC_FREE
    ((p.pPatWords) ? (free((char *) (p.pPatWords)), (p.pPatWords) = 0) : 0);
    // inlining of macro ABC_FREE
    ((p) ? (free((char *) (p)), (p) = 0) : 0);
}

/**Function*************************************************************

  Synopsis    [Prints stats for the fraiging manager.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Fra_ManPrint(p *Fra_Man_t) {
    var nMemory float64  = 1.0*Aig_ManObjNumMax(p.pManAig)*(p.pSml.nWordsTotal*sizeof(unsigned)+6*sizeof(void*))/(1<<20);
    fmt.Printf("SimWord = %d. Round = %d.  Mem = %0.2f MB.  LitBeg = %d.  LitEnd = %d. (%6.2f %%).\n", p.pPars.nSimWords, p.pSml.nSimRounds, nMemory, p.nLitsBeg, p.nLitsEnd, 100.0*p.nLitsEnd/(p.nLitsBeg?p.nLitsBeg:1));
    fmt.Printf("Proof = %d. Cex = %d. Fail = %d. FailReal = %d. C-lim = %d. ImpRatio = %6.2f %%\n", p.nSatProof, p.nSatCallsSat, p.nSatFails, p.nSatFailsReal, p.pPars.nBTLimitNode, Fra_ImpComputeStateSpaceRatio(p));
    fmt.Printf("NBeg = %d. NEnd = %d. (Gain = %6.2f %%).  RBeg = %d. REnd = %d. (Gain = %6.2f %%).\n", p.nNodesBeg, p.nNodesEnd, 100.0*(p.nNodesBeg-p.nNodesEnd)/(p.nNodesBeg?p.nNodesBeg:1), p.nRegsBeg, p.nRegsEnd, 100.0*(p.nRegsBeg-p.nRegsEnd)/(p.nRegsBeg?p.nRegsBeg:1));
    if p.pSat {             Sat_SolverPrintStats( os.Stdout, p.pSat ); }
    if p.pPars.fUse1Hot {  Fra_OneHotEstimateCoverage( p, p.vOneHots ); }
    // inlining of macro ABC_PRT
    Abc_Print(1, "%s =", ("AIG simulation  "));
Abc_Print(1, "%9.2f sec\n", 1.0*(float64((p.pSml.timeSim)))/(float64(int64(time.Second))));;
    // inlining of macro ABC_PRT
    Abc_Print(1, "%s =", ("AIG traversal   "));
Abc_Print(1, "%9.2f sec\n", 1.0*(float64((p.timeTrav)))/(float64(int64(time.Second))));;
    if p.timeRwr {
    // inlining of macro ABC_PRT
    Abc_Print(1, "%s =", ("AIG rewriting   "));
Abc_Print(1, "%9.2f sec\n", 1.0*(float64((p.timeRwr)))/(float64(int64(time.Second))));;
    }
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
    if p.nSpeculs {
    fmt.Printf("Speculations = %d.\n", p.nSpeculs); }
    fflush( os.Stdout );
}

////////////////////////////////////////////////////////////////////////
///                       END OF FILE                                ///
////////////////////////////////////////////////////////////////////////


//ABC_NAMESPACE_IMPL_END

