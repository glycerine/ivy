package sim

import (
	"fmt"
	"time"
)

/**CFile****************************************************************

  FileName    [simMan.c]

  SystemName  [ABC: Logic synthesis and verification system.]

  PackageName [Network and node package.]

  Synopsis    [Simulation manager.]

  Author      [Alan Mishchenko]
  
  Affiliation [UC Berkeley]

  Date        [Ver. 1.0. Started - June 20, 2005.]

  Revision    [$Id: simMan.c,v 1.00 2005/06/20 00:00:00 alanmi Exp $]

***********************************************************************/

//#include "base/abc/abc.h"
//#include "sim.h"

//ABC_NAMESPACE_IMPL_START


////////////////////////////////////////////////////////////////////////
///                        DECLARATIONS                              ///
////////////////////////////////////////////////////////////////////////

////////////////////////////////////////////////////////////////////////
///                     FUNCTION DEFINITIONS                         ///
////////////////////////////////////////////////////////////////////////

/**Function*************************************************************

  Synopsis    [Starts the simulation manager.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Sym_ManStart(pNtk *Abc_Ntk_t, fVerbose int32) *Sym_Man_t {
    var p *Sym_Man_t;
    var i, v int32; 
    // start the manager
    // inlining of macro ABC_ALLOC
    p = ((Sym_Man_t *) malloc(sizeof(Sym_Man_t) * (size_t)(1)));
    memset( p, 0, sizeof(Sym_Man_t) );
    p.pNtk = pNtk;
    p.vNodes     = Abc_NtkDfs( pNtk, 0 );
    p.nInputs    = Abc_NtkCiNum(p.pNtk);
    p.nOutputs   = Abc_NtkCoNum(p.pNtk);
    // internal simulation information
    // inlining of macro SIM_NUM_WORDS
    p.nSimWords  = (((p.nInputs)>>5) + (((p.nInputs)&31) > 0));
    p.vSim       = Sim_UtilInfoAlloc( Abc_NtkObjNumMax(pNtk), p.nSimWords, 0 );
    // symmetry info for each output
    p.vMatrSymms    = Vec_PtrStart( p.nOutputs );
    p.vMatrNonSymms = Vec_PtrStart( p.nOutputs );
    p.vPairsTotal   = Vec_IntStart( p.nOutputs );
    p.vPairsSym     = Vec_IntStart( p.nOutputs );
    p.vPairsNonSym  = Vec_IntStart( p.nOutputs );
    for  i = 0; i < p.nOutputs; i++  {
        p.vMatrSymms.pArray[i]    = Extra_BitMatrixStart( p.nInputs );
        p.vMatrNonSymms.pArray[i] = Extra_BitMatrixStart( p.nInputs );
    }
    // temporary patterns
    // inlining of macro ABC_ALLOC
    p.uPatRand = ((unsigned *) malloc(sizeof(unsigned) * (size_t)(p.nSimWords)));
    // inlining of macro ABC_ALLOC
    p.uPatCol  = ((unsigned *) malloc(sizeof(unsigned) * (size_t)(p.nSimWords)));
    // inlining of macro ABC_ALLOC
    p.uPatRow  = ((unsigned *) malloc(sizeof(unsigned) * (size_t)(p.nSimWords)));
    p.vVarsU   = Vec_IntStart( 100 );
    p.vVarsV   = Vec_IntStart( 100 );
    // compute supports
    p.vSuppFun  = Sim_ComputeFunSupp( pNtk, fVerbose );
    p.vSupports = Vec_VecStart( p.nOutputs );
    for  i = 0; i < p.nOutputs; i++  {
        for  v = 0; v < p.nInputs; v++  {
            // inlining of macro Sim_SuppFunHasVar
            // inlining of macro Sim_HasBit
            if  (((unsigned*)(p.vSuppFun).pArray[i])[((v))>>5]  & (1<<(((v)) & 31))) > 0 {
                Vec_VecPushInt( p.vSupports, i, v ); }
        }
    }
    return p;
}

/**Function*************************************************************

  Synopsis    [Stops the simulation manager.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Sym_ManStop(p *Sym_Man_t) {
    var i int32;
    Sym_ManPrintStats( p );
    if p.vSuppFun {     Sim_UtilInfoFree( p.vSuppFun ); }   
    if p.vSim {         Sim_UtilInfoFree( p.vSim ); }   
    if p.vNodes {       Vec_PtrFree( p.vNodes ); }
    if p.vSupports {    Vec_VecFree( p.vSupports ); }
    for  i = 0; i < p.nOutputs; i++  {
        Extra_BitMatrixStop( (Extra_BitMat_t *)p.vMatrSymms.pArray[i] );
        Extra_BitMatrixStop( (Extra_BitMat_t *)p.vMatrNonSymms.pArray[i] );
    }
    Vec_IntFree( p.vVarsU );
    Vec_IntFree( p.vVarsV );
    Vec_PtrFree( p.vMatrSymms );
    Vec_PtrFree( p.vMatrNonSymms );
    Vec_IntFree( p.vPairsTotal );
    Vec_IntFree( p.vPairsSym );
    Vec_IntFree( p.vPairsNonSym );
    p.uPatRand = nil;
    p.uPatCol = nil;
    p.uPatRow = nil;
    p = nil;
}

/**Function*************************************************************

  Synopsis    [Prints the manager statisticis.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Sym_ManPrintStats(p *Sym_Man_t) {
//    printf( "Inputs = %5d. Outputs = %5d. Sim words = %5d.\n", 
//        Abc_NtkCiNum(p->pNtk), Abc_NtkCoNum(p->pNtk), p->nSimWords );
    fmt.Printf("Total symm         = %8d.\n", p.nPairsSymm);
    fmt.Printf("Structural symm    = %8d.\n", p.nPairsSymmStr);
    fmt.Printf("Total non-sym      = %8d.\n", p.nPairsNonSymm);
    fmt.Printf("Total var pairs    = %8d.\n", p.nPairsTotal);
    fmt.Printf("Sat runs SAT       = %8d.\n", p.nSatRunsSat);
    fmt.Printf("Sat runs UNSAT     = %8d.\n", p.nSatRunsUnsat);
    // inlining of macro ABC_PRT
    Abc_Print(1, "%s =", ("Structural  "));
Abc_Print(1, "%9.2f sec\n", 1.0*(float64((p.timeStruct)))/(float64(int64(time.Second))));;
    // inlining of macro ABC_PRT
    Abc_Print(1, "%s =", ("Simulation  "));
Abc_Print(1, "%9.2f sec\n", 1.0*(float64((p.timeSim)))/(float64(int64(time.Second))));;
    // inlining of macro ABC_PRT
    Abc_Print(1, "%s =", ("Matrix      "));
Abc_Print(1, "%9.2f sec\n", 1.0*(float64((p.timeMatr)))/(float64(int64(time.Second))));;
    // inlining of macro ABC_PRT
    Abc_Print(1, "%s =", ("Counting    "));
Abc_Print(1, "%9.2f sec\n", 1.0*(float64((p.timeCount)))/(float64(int64(time.Second))));;
    // inlining of macro ABC_PRT
    Abc_Print(1, "%s =", ("Fraiging    "));
Abc_Print(1, "%9.2f sec\n", 1.0*(float64((p.timeFraig)))/(float64(int64(time.Second))));;
    // inlining of macro ABC_PRT
    Abc_Print(1, "%s =", ("SAT         "));
Abc_Print(1, "%9.2f sec\n", 1.0*(float64((p.timeSat)))/(float64(int64(time.Second))));;
    // inlining of macro ABC_PRT
    Abc_Print(1, "%s =", ("TOTAL       "));
Abc_Print(1, "%9.2f sec\n", 1.0*(float64((p.timeTotal)))/(float64(int64(time.Second))));;
}


/**Function*************************************************************

  Synopsis    [Starts the simulation manager.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Sim_ManStart(pNtk *Abc_Ntk_t, fLightweight int32) *Sim_Man_t {
    var p *Sim_Man_t;
    // start the manager
    // inlining of macro ABC_ALLOC
    p = ((Sim_Man_t *) malloc(sizeof(Sim_Man_t) * (size_t)(1)));
    memset( p, 0, sizeof(Sim_Man_t) );
    p.pNtk = pNtk;
    p.nInputs    = Abc_NtkCiNum(p.pNtk);
    p.nOutputs   = Abc_NtkCoNum(p.pNtk);
    // internal simulation information
    p.nSimBits   = 2048;
    // inlining of macro SIM_NUM_WORDS
    p.nSimWords  = (((p.nSimBits)>>5) + (((p.nSimBits)&31) > 0));
    p.vSim0      = Sim_UtilInfoAlloc( Abc_NtkObjNumMax(pNtk), p.nSimWords, 0 );
    p.fLightweight = fLightweight;
    if !p.fLightweight {
        p.vSim1      = Sim_UtilInfoAlloc( Abc_NtkObjNumMax(pNtk), p.nSimWords, 0 );
        // support information
        p.nSuppBits  = Abc_NtkCiNum(pNtk);
        // inlining of macro SIM_NUM_WORDS
        p.nSuppWords = (((p.nSuppBits)>>5) + (((p.nSuppBits)&31) > 0));
        p.vSuppStr   = Sim_ComputeStrSupp( pNtk );
        p.vSuppFun   = Sim_UtilInfoAlloc( Abc_NtkCoNum(p.pNtk),  p.nSuppWords, 1 );
        // other data
        p.pMmPat     = Extra_MmFixedStart( sizeof(Sim_Pat_t) + p.nSuppWords * sizeof(unsigned) ); 
        p.vFifo      = Vec_PtrAlloc( 100 );
        p.vDiffs     = Vec_IntAlloc( 100 );
        // allocate support targets (array of unresolved outputs for each input)
        p.vSuppTargs = Vec_VecStart( p.nInputs );
    }
    return p;
}

/**Function*************************************************************

  Synopsis    [Stops the simulation manager.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Sim_ManStop(p *Sim_Man_t) {
    Sim_ManPrintStats( p );
    if p.vSim0 {        Sim_UtilInfoFree( p.vSim0 ); }       
    if p.vSim1 {        Sim_UtilInfoFree( p.vSim1 ); }       
    if p.vSuppStr {     Sim_UtilInfoFree( p.vSuppStr ); }    
//    if ( p->vSuppFun )     Sim_UtilInfoFree( p->vSuppFun );    
    if p.vSuppTargs {   Vec_VecFree( p.vSuppTargs ); }
    if p.pMmPat {       Extra_MmFixedStop( p.pMmPat ); }
    if p.vFifo {        Vec_PtrFree( p.vFifo ); }
    if p.vDiffs {       Vec_IntFree( p.vDiffs ); }
    p = nil;
}

/**Function*************************************************************

  Synopsis    [Prints the manager statisticis.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Sim_ManPrintStats(p *Sim_Man_t) {
//    printf( "Inputs = %5d. Outputs = %5d. Sim words = %5d.\n", 
//        Abc_NtkCiNum(p->pNtk), Abc_NtkCoNum(p->pNtk), p->nSimWords );
    fmt.Printf("Total func supps   = %8d.\n", Sim_UtilCountSuppSizes(p, 0));
    fmt.Printf("Total struct supps = %8d.\n", Sim_UtilCountSuppSizes(p, 1));
    fmt.Printf("Sat runs SAT       = %8d.\n", p.nSatRunsSat);
    fmt.Printf("Sat runs UNSAT     = %8d.\n", p.nSatRunsUnsat);
    // inlining of macro ABC_PRT
    Abc_Print(1, "%s =", ("Simulation  "));
Abc_Print(1, "%9.2f sec\n", 1.0*(float64((p.timeSim)))/(float64(int64(time.Second))));;
    // inlining of macro ABC_PRT
    Abc_Print(1, "%s =", ("Traversal   "));
Abc_Print(1, "%9.2f sec\n", 1.0*(float64((p.timeTrav)))/(float64(int64(time.Second))));;
    // inlining of macro ABC_PRT
    Abc_Print(1, "%s =", ("Fraiging    "));
Abc_Print(1, "%9.2f sec\n", 1.0*(float64((p.timeFraig)))/(float64(int64(time.Second))));;
    // inlining of macro ABC_PRT
    Abc_Print(1, "%s =", ("SAT         "));
Abc_Print(1, "%9.2f sec\n", 1.0*(float64((p.timeSat)))/(float64(int64(time.Second))));;
    // inlining of macro ABC_PRT
    Abc_Print(1, "%s =", ("TOTAL       "));
Abc_Print(1, "%9.2f sec\n", 1.0*(float64((p.timeTotal)))/(float64(int64(time.Second))));;
}



/**Function*************************************************************

  Synopsis    [Returns one simulation pattern.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Sim_ManPatAlloc(p *Sim_Man_t) *Sim_Pat_t {
    var pPat *Sim_Pat_t;
    pPat = (Sim_Pat_t *)Extra_MmFixedEntryFetch( p.pMmPat );
    pPat.Output = -1;
    pPat.pData  = (unsigned *)((char *)pPat + sizeof(Sim_Pat_t));
    memset( pPat.pData, 0, p.nSuppWords * sizeof(unsigned) );
    return pPat;
}

/**Function*************************************************************

  Synopsis    [Returns one simulation pattern.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Sim_ManPatFree(p *Sim_Man_t, pPat *Sim_Pat_t) {
    Extra_MmFixedEntryRecycle( p.pMmPat, (char *)pPat );
}

////////////////////////////////////////////////////////////////////////
///                       END OF FILE                                ///
////////////////////////////////////////////////////////////////////////


//ABC_NAMESPACE_IMPL_END

