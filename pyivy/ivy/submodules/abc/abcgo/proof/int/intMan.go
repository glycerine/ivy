package int

import (
	"fmt"
	"time"
)

/**CFile****************************************************************

  FileName    [intMan.c]

  SystemName  [ABC: Logic synthesis and verification system.]

  PackageName [Interpolation engine.]

  Synopsis    [Interpolation manager procedures.]

  Author      [Alan Mishchenko]
  
  Affiliation [UC Berkeley]

  Date        [Ver. 1.0. Started - June 24, 2008.]

  Revision    [$Id: intMan.c,v 1.00 2005/06/20 00:00:00 alanmi Exp $]

***********************************************************************/

//#include "intInt.h"
//#include "aig/ioa/ioa.h"

//ABC_NAMESPACE_IMPL_START


////////////////////////////////////////////////////////////////////////
///                        DECLARATIONS                              ///
////////////////////////////////////////////////////////////////////////

////////////////////////////////////////////////////////////////////////
///                     FUNCTION DEFINITIONS                         ///
////////////////////////////////////////////////////////////////////////

/**Function*************************************************************

  Synopsis    [Creates the interpolation manager.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Inter_ManCreate(pAig *Aig_Man_t, pPars *Inter_ManParams_t) *Inter_Man_t {
    var p *Inter_Man_t;
    // create interpolation manager
    // inlining of macro ABC_ALLOC
    p = ((Inter_Man_t *) malloc(sizeof(Inter_Man_t) * (size_t)(1)));
    memset( p, 0, sizeof(Inter_Man_t) );
    p.vVarsAB = Vec_IntAlloc( Aig_ManRegNum(pAig) );
    p.nConfLimit = pPars.nBTLimit;
    p.fVerbose = pPars.fVerbose;
    p.pFileName = pPars.pFileName;
    p.pAig = pAig;
    if pPars.fDropInvar {
        p.vInters = Vec_PtrAlloc( 100 ); }
    return p;
}

/**Function*************************************************************

  Synopsis    [Cleans the interpolation manager.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Inter_ManClean(p *Inter_Man_t) {
    if p.vInters {
        var pMan *Aig_Man_t;
        var i int32;
        // inlining of macro Vec_PtrForEachEntry
        for  i = 0; (i < Vec_PtrSize(p.vInters)) && (((pMan) = (Aig_Man_t *)Vec_PtrEntry(p.vInters, i)), 1); i++  {
            Aig_ManStop( pMan );
        }
        Vec_PtrClear( p.vInters );
    }
    if p.pCnfInter {
        Cnf_DataFree( p.pCnfInter ); }
    if p.pCnfFrames {
        Cnf_DataFree( p.pCnfFrames ); }
    if p.pInter {
        Aig_ManStop( p.pInter ); }
    if p.pFrames {
        Aig_ManStop( p.pFrames ); }
}

/**Function*************************************************************

  Synopsis    [Writes interpolant into a file.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Inter_ManInterDump(p *Inter_Man_t, fProved int32) {
    var pFileName string;
    if p.pFileName {
        pFileName = p.pFileName;
    } else {
        pFileName = (char *)"invar.aig";
    }
    var pMan *Aig_Man_t;
    pMan = Aig_ManDupArray( p.vInters );
    Ioa_WriteAiger( pMan, pFileName, 0, 0 );
    Aig_ManStop( pMan );
    if fProved != 0 {
        fmt.Printf("Inductive invariant is dumped into file \"%s\".\n", pFileName);
    } else {
        fmt.Printf("Interpolants are dumped into file \"%s\".\n", pFileName); }
}

/**Function*************************************************************

  Synopsis    [Frees the interpolation manager.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Inter_ManStop(p *Inter_Man_t, fProved int32) {
    if p.fVerbose {
        p.timeOther = p.timeTotal-p.timeRwr-p.timeCnf-p.timeSat-p.timeInt-p.timeEqu;
        fmt.Printf("Runtime statistics:\n");
        // inlining of macro ABC_PRTP
        Abc_Print(1, "%s =", ("Rewriting  "));
Abc_Print(1, "%9.2f sec (%6.2f %%)\n", 1.0*(float64((p.timeRwr)))/(float64(int64(time.Second))), (float64((p.timeTotal)))? 100.0*(float64((p.timeRwr)))/(float64((p.timeTotal))) : 0.0);;
        // inlining of macro ABC_PRTP
        Abc_Print(1, "%s =", ("CNF mapping"));
Abc_Print(1, "%9.2f sec (%6.2f %%)\n", 1.0*(float64((p.timeCnf)))/(float64(int64(time.Second))), (float64((p.timeTotal)))? 100.0*(float64((p.timeCnf)))/(float64((p.timeTotal))) : 0.0);;
        // inlining of macro ABC_PRTP
        Abc_Print(1, "%s =", ("SAT solving"));
Abc_Print(1, "%9.2f sec (%6.2f %%)\n", 1.0*(float64((p.timeSat)))/(float64(int64(time.Second))), (float64((p.timeTotal)))? 100.0*(float64((p.timeSat)))/(float64((p.timeTotal))) : 0.0);;
        // inlining of macro ABC_PRTP
        Abc_Print(1, "%s =", ("Interpol   "));
Abc_Print(1, "%9.2f sec (%6.2f %%)\n", 1.0*(float64((p.timeInt)))/(float64(int64(time.Second))), (float64((p.timeTotal)))? 100.0*(float64((p.timeInt)))/(float64((p.timeTotal))) : 0.0);;
        // inlining of macro ABC_PRTP
        Abc_Print(1, "%s =", ("Containment"));
Abc_Print(1, "%9.2f sec (%6.2f %%)\n", 1.0*(float64((p.timeEqu)))/(float64(int64(time.Second))), (float64((p.timeTotal)))? 100.0*(float64((p.timeEqu)))/(float64((p.timeTotal))) : 0.0);;
        // inlining of macro ABC_PRTP
        Abc_Print(1, "%s =", ("Other      "));
Abc_Print(1, "%9.2f sec (%6.2f %%)\n", 1.0*(float64((p.timeOther)))/(float64(int64(time.Second))), (float64((p.timeTotal)))? 100.0*(float64((p.timeOther)))/(float64((p.timeTotal))) : 0.0);;
        // inlining of macro ABC_PRTP
        Abc_Print(1, "%s =", ("TOTAL      "));
Abc_Print(1, "%9.2f sec (%6.2f %%)\n", 1.0*(float64((p.timeTotal)))/(float64(int64(time.Second))), (float64((p.timeTotal)))? 100.0*(float64((p.timeTotal)))/(float64((p.timeTotal))) : 0.0);;
    }

    if p.vInters {
        Inter_ManInterDump( p, fProved ); }

    if p.pCnfAig {
        Cnf_DataFree( p.pCnfAig ); }
    if p.pAigTrans {
        Aig_ManStop( p.pAigTrans ); }
    if p.pInterNew {
        Aig_ManStop( p.pInterNew ); }
    Inter_ManClean( p );
    Vec_PtrFreeP( &p.vInters );
    Vec_IntFreeP( &p.vVarsAB );
    p = nil;
}


////////////////////////////////////////////////////////////////////////
///                       END OF FILE                                ///
////////////////////////////////////////////////////////////////////////


//ABC_NAMESPACE_IMPL_END

