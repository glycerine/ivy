package dar

import (
	"fmt"
	"os"
	"time"
)

/**CFile****************************************************************

  FileName    [darMan.c]

  SystemName  [ABC: Logic synthesis and verification system.]

  PackageName [DAG-aware AIG rewriting.]

  Synopsis    [AIG manager.]

  Author      [Alan Mishchenko]
  
  Affiliation [UC Berkeley]

  Date        [Ver. 1.0. Started - April 28, 2007.]

  Revision    [$Id: darMan.c,v 1.00 2007/04/28 00:00:00 alanmi Exp $]

***********************************************************************/

//#include "darInt.h"

//ABC_NAMESPACE_IMPL_START

////////////////////////////////////////////////////////////////////////
///                        DECLARATIONS                              ///
////////////////////////////////////////////////////////////////////////

////////////////////////////////////////////////////////////////////////
///                     FUNCTION DEFINITIONS                         ///
////////////////////////////////////////////////////////////////////////

/**Function*************************************************************

  Synopsis    [Starts the rewriting manager.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Dar_ManStart(pAig *Aig_Man_t, pPars *Dar_RwrPar_t) *Dar_Man_t {
    var p *Dar_Man_t;
    Aig_ManCleanData( pAig );
    // inlining of macro ABC_ALLOC
    p = ((Dar_Man_t *) malloc(sizeof(Dar_Man_t) * (size_t)(1)));
    memset( p, 0, sizeof(Dar_Man_t) );
    p.pPars = pPars;
    p.pAig  = pAig;
    p.vCutNodes = Vec_PtrAlloc( 1000 );
    p.pMemCuts = Aig_MmFixedStart( p.pPars.nCutsMax * sizeof(Dar_Cut_t), 1024 );
    p.vLeavesBest = Vec_PtrAlloc( 4 );
    return p;
}

/**Function*************************************************************

  Synopsis    [Stops the rewriting manager.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Dar_ManStop(p *Dar_Man_t) {
    if p.pPars.fVerbose {
        Dar_ManPrintStats( p ); }
    if p.vCutNodes {
        Vec_PtrFree( p.vCutNodes ); }
    if p.pMemCuts {
        Aig_MmFixedStop( p.pMemCuts, 0 ); }
    if p.vLeavesBest { 
        Vec_PtrFree( p.vLeavesBest ); }
    // inlining of macro ABC_FREE
    ((p) ? (free((char *) (p)), (p) = 0) : 0);
}

/**Function*************************************************************

  Synopsis    [Stops the AIG manager.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Dar_ManPrintStats(p *Dar_Man_t) {
    var pCanons [222]uint32;
    var Gain, i int32;
    extern void Kit_DsdPrintFromTruth( unsigned * pTruth, int nVars );

    Gain = p.nNodesInit - Aig_ManNodeNum(p.pAig);
    fmt.Printf("Tried = %8d. Beg = %8d. End = %8d. Gain = %6d. (%6.2f %%).  Cut mem = %d MB\n", p.nNodesTried, p.nNodesInit, Aig_ManNodeNum(p.pAig), Gain, 100.0*Gain/p.nNodesInit, p.nCutMemUsed);
    fmt.Printf("Cuts = %8d. Tried = %8d. Used = %8d. Bad = %5d. Skipped = %5d. Ave = %.2f.\n", p.nCutsAll, p.nCutsTried, p.nCutsUsed, p.nCutsBad, p.nCutsSkipped, float32(p.nCutsUsed)/Aig_ManNodeNum(p.pAig));

    fmt.Printf("Bufs = %5d. BufMax = %5d. BufReplace = %6d. BufFix = %6d.  Levels = %4d.\n", Aig_ManBufNum(p.pAig), p.pAig.nBufMax, p.pAig.nBufReplaces, p.pAig.nBufFixes, Aig_ManLevels(p.pAig));
    // inlining of macro ABC_PRT
    Abc_Print(1, "%s =", ("Cuts  "));
Abc_Print(1, "%9.2f sec\n", 1.0*(float64((p.timeCuts)))/(float64(int64(time.Second))));;
    // inlining of macro ABC_PRT
    Abc_Print(1, "%s =", ("Eval  "));
Abc_Print(1, "%9.2f sec\n", 1.0*(float64((p.timeEval)))/(float64(int64(time.Second))));;
    // inlining of macro ABC_PRT
    Abc_Print(1, "%s =", ("Other "));
Abc_Print(1, "%9.2f sec\n", 1.0*(float64((p.timeOther)))/(float64(int64(time.Second))));;
    // inlining of macro ABC_PRT
    Abc_Print(1, "%s =", ("TOTAL "));
Abc_Print(1, "%9.2f sec\n", 1.0*(float64((p.timeTotal)))/(float64(int64(time.Second))));;

    if !p.pPars.fVeryVerbose {
        return; }
    Dar_LibReturnCanonicals( pCanons );
    for  i = 0; i < 222; i++  {
        if p.ClassGains[i] == 0 && p.ClassTimes[i] == 0 {
            continue; }
        fmt.Printf("%3d : ", i);
        fmt.Printf("G = %6d (%5.2f %%)  ", p.ClassGains[i], Gain? 100.0*p.ClassGains[i]/Gain : 0.0);
        fmt.Printf("S = %8d (%5.2f %%)  ", p.ClassSubgs[i], p.nTotalSubgs? 100.0*p.ClassSubgs[i]/p.nTotalSubgs : 0.0);
        fmt.Printf("R = %7d   ", p.ClassGains[i]? p.ClassSubgs[i]/p.ClassGains[i] : 9999999);
//        Kit_DsdPrintFromTruth( pCanons + i, 4 );
//        ABC_PRTP( "T", p->ClassTimes[i], p->timeEval );
        fmt.Printf("\n");
    }
    fflush( os.Stdout );
}


/**Function*************************************************************

  Synopsis    []

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
     

                      

                         

                        

                         
 
                          
          
                                       
                               
     
                            
                                                  
                           
                                     
                                                     
                               
     
 
      

////////////////////////////////////////////////////////////////////////
///                       END OF FILE                                ///
////////////////////////////////////////////////////////////////////////


//ABC_NAMESPACE_IMPL_END

