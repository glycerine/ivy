package gia

import "time"

/**CFile****************************************************************

  FileName    [giaSim.c]

  SystemName  [ABC: Logic synthesis and verification system.]

  PackageName [Scalable AIG package.]

  Synopsis    [Fast sequential simulator.]

  Author      [Alan Mishchenko]
  
  Affiliation [UC Berkeley]

  Date        [Ver. 1.0. Started - June 20, 2005.]

  Revision    [$Id: giaSim.c,v 1.00 2005/06/20 00:00:00 alanmi Exp $]

***********************************************************************/

//#include "gia.h"

//ABC_NAMESPACE_IMPL_START


////////////////////////////////////////////////////////////////////////
///                        DECLARATIONS                              ///
////////////////////////////////////////////////////////////////////////

   
// extern void Gia_ManResetRandom( Gia_ParSim_t * pPars );


////////////////////////////////////////////////////////////////////////
///                     FUNCTION DEFINITIONS                         ///
////////////////////////////////////////////////////////////////////////

/**Function*************************************************************

  Synopsis    []

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/

/**Function*************************************************************

  Synopsis    [Creates fast simulation manager.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/

/**Function*************************************************************

  Synopsis    []

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/

/**Function*************************************************************

  Synopsis    []

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/

/**Function*************************************************************

  Synopsis    []

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/

/**Function*************************************************************

  Synopsis    []

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/

/**Function*************************************************************

  Synopsis    []

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/

/**Function*************************************************************

  Synopsis    []

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/

/**Function*************************************************************

  Synopsis    []

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/

/**Function*************************************************************

  Synopsis    []

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/

/**Function*************************************************************

  Synopsis    [Compares simulation info of two nodes.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/

/**Function*************************************************************

  Synopsis    [Compares simulation info of one node.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/

/**Function*************************************************************

  Synopsis    [Creates equivalence class.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/

/**Function*************************************************************

  Synopsis    [Refines one equivalence class.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/

/**Function*************************************************************

  Synopsis    [Computes hash key of the simuation info.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
var spStatic_Gia_Sim2HashKey_s_Primes [16]int32 = [16]int32{ 
        1291, 1699, 1999, 2357, 2953, 3313, 3907, 4177, 
        4831, 5147, 5647, 6343, 6899, 7103, 7873, 8147 };

/**Function*************************************************************

  Synopsis    [Refines nodes belonging to candidate constant class.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/

/**Function*************************************************************

  Synopsis    [Refines equivalences after one simulation round.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/

/**Function*************************************************************

  Synopsis    [Returns index of the first pattern that failed.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/

/**Function*************************************************************

  Synopsis    [Returns index of the PO and pattern that failed it.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/

/**Function*************************************************************

  Synopsis    [Returns the counter-example.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Gia_Sim2GenerateCounter(pAig *Gia_Man_t, iFrame int32, iOut int32, nWords int32, iPat int32) *Abc_Cex_t {
    var p *Abc_Cex_t;
    var pData *uint32;
    var f, i, w, Counter int32;
    p = Abc_CexAlloc( Gia_ManRegNum(pAig), Gia_ManPiNum(pAig), iFrame+1 );
    p.iFrame = iFrame;
    p.iPo = iOut;
    // fill in the binary data
    Counter = p.nRegs;
    // inlining of macro ABC_ALLOC
    pData = ((unsigned *) malloc(sizeof(unsigned) * (size_t)(nWords)));
    for  f = 0; f <= iFrame; f++, Counter += p.nPis  {
    for  i = 0; i < Gia_ManPiNum(pAig); i++  {
        for  w = nWords-1; w >= 0; w--  {
            pData[w] = Gia_ManRandom( 0 );
        }
        if Abc_InfoHasBit( pData, iPat ) {
            Abc_InfoSetBit( p.pData, Counter + i ); }
    }
    }
    // inlining of macro ABC_FREE
    pData = nil;
    return p;
}
 
/**Function*************************************************************

  Synopsis    [Performs simulation to refine equivalence classes.]

  Description [Returns 1 if counter-example is detected.]
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Gia_ManSimSimulateEquiv(pAig *Gia_Man_t, pPars *Gia_ParSim_t) int32 {
    var p *Gia_Sim2_t;
    var pObj *Gia_Obj_t;
    var clkTotal abctime  = Abc_Clock();
    var i, RetValue, iOut, iPat int32;
    var nTimeToStop abctime;
    if pPars.TimeLimit {
        nTimeToStop = pPars.TimeLimit * int64(time.Second) + Abc_Clock();
    } else {
        nTimeToStop = 0;
    }
    assert(pAig.pReprs && pAig.pNexts, "pAig.pReprs && pAig.pNexts");
    // inlining of macro ABC_FREE
    pAig.pCexSeq = nil;
    p = Gia_Sim2Create( pAig, pPars );
    Gia_ManResetRandom( pPars );
    // inlining of macro Gia_ManForEachRo
    for  i = 0; (i < Gia_ManRegNum(p.pAig)) && ((pObj) = Gia_ManCi(p.pAig, Gia_ManPiNum(p.pAig)+i)); i++  {
        Gia_Sim2InfoZero( p, Gia_Sim2Data(p, Gia_ObjValue(pObj)) );
    }
    for  i = 0; i < pPars.nIters; i++  {
        Gia_Sim2SimulateRound( p );
        if pPars.fVerbose {
            Abc_Print( 1, "Frame %4d out of %4d and timeout %3d sec. ", i+1, pPars.nIters, pPars.TimeLimit );
            if pAig.pReprs && pAig.pNexts {
                Abc_Print( 1, "Lits = %4d. ", Gia_ManEquivCountLitsAll(pAig) ); }
            Abc_Print( 1, "Time = %7.2f sec\r", (1.0*Abc_Clock()-clkTotal)/int64(time.Second) );
        }
        if pPars.fCheckMiter && (Gia_Sim2CheckPos( p, &iOut, &iPat ) != 0) {
            Gia_ManResetRandom( pPars );
            pPars.iOutFail = iOut;
            pAig.pCexSeq = Gia_Sim2GenerateCounter( pAig, i, iOut, p.nWords, iPat );
            Abc_Print( 1, "Output %d of miter \"%s\" was asserted in frame %d.  ", iOut, pAig.pName, i );
            if !Gia_ManVerifyCex( pAig, pAig.pCexSeq, 0 ) {
//                Abc_Print( 1, "\n" );
                Abc_Print( 1, "\nGenerated counter-example is INVALID.                    " );
//                Abc_Print( 1, "\n" );
            
            } else {
//                Abc_Print( 1, "\n" );
//                if ( pPars->fVerbose )
//                Abc_Print( 1, "\nGenerated counter-example is verified correctly.         " );
//                Abc_Print( 1, "\n" );
            }
            RetValue = 1;
            break;
        }
        if pAig.pReprs && pAig.pNexts {
            Gia_Sim2InfoRefineEquivs( p ); }
        if Abc_Clock() > nTimeToStop {
            i++;
            break;
        }
        if i < pPars.nIters - 1 {
            Gia_Sim2InfoTransfer( p ); }
    }
    Gia_Sim2Delete( p );
    if pAig.pCexSeq == nil {
        Abc_Print( 1, "No bug detected after simulating %d frames with %d words.  ", i, pPars.nWords ); }
    Abc_PrintTime( 1, "Time", Abc_Clock() - clkTotal );
    return RetValue;
}

////////////////////////////////////////////////////////////////////////
///                       END OF FILE                                ///
////////////////////////////////////////////////////////////////////////


//ABC_NAMESPACE_IMPL_END

