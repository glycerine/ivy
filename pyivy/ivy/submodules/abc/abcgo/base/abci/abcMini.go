package abci

import (
	"fmt"
	"os"
)

/**CFile****************************************************************

  FileName    [abcMini.c]

  SystemName  [ABC: Logic synthesis and verification system.]

  PackageName [Network and node package.]

  Synopsis    [Interface to the minimalistic AIG package.]

  Author      [Alan Mishchenko]
  
  Affiliation [UC Berkeley]

  Date        [Ver. 1.0. Started - June 20, 2005.]

  Revision    [$Id: abcMini.c,v 1.00 2005/06/20 00:00:00 alanmi Exp $]

***********************************************************************/

//#include "base/abc/abc.h"
//#include "base/main/main.h"
//#include "aig/miniaig/miniaig.h"

//ABC_NAMESPACE_IMPL_START


////////////////////////////////////////////////////////////////////////
///                        DECLARATIONS                              ///
////////////////////////////////////////////////////////////////////////

////////////////////////////////////////////////////////////////////////
///                     FUNCTION DEFINITIONS                         ///
////////////////////////////////////////////////////////////////////////


/**Function*************************************************************

  Synopsis    [Converts the network from the AIG manager into ABC.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
// MiniAIG conversion functions are translated in mini_convert.go.

/**Function*************************************************************

  Synopsis    [Procedures to update internal ABC network using AIG node array.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Abc_NtkInputMiniAig(pAbc *Abc_Frame_t, p any) {
    var pNtk *Abc_Ntk_t;
    if pAbc == nil {
        fmt.Printf("ABC framework is not initialized by calling Abc_Start()\n"); }
    pNtk = Abc_NtkFromMiniAig( (Mini_Aig_t *)p );
    Abc_FrameReplaceCurrentNetwork( pAbc, pNtk );
//    Abc_NtkDelete( pNtk );
}
func Abc_NtkOutputMiniAig(pAbc *Abc_Frame_t) any {
    var pNtk *Abc_Ntk_t;
    if pAbc == nil {
        fmt.Printf("ABC framework is not initialized by calling Abc_Start()\n"); }
    pNtk = Abc_FrameReadNtk( pAbc );
    if pNtk == nil {
        fmt.Printf("Current network in ABC framework is not defined.\n"); }
    return Abc_NtkToMiniAig( pNtk );
}
func Abc_NtkSetFlopNum(pAbc *Abc_Frame_t, nFlops int32) {
    extern void Abc_NtkMakeSeq( Abc_Ntk_t * pNtk, int nFlops );
    var pNtk *Abc_Ntk_t;
    if pAbc == nil {
        fmt.Printf("ABC framework is not initialized by calling Abc_Start()\n"); }
    pNtk = Abc_FrameReadNtk( pAbc );
    if pNtk == nil {
        fmt.Printf("Current network in ABC framework is not defined.\n"); }
    Abc_NtkMakeSeq( pNtk, nFlops );
}

/**Function*************************************************************

  Synopsis    [Testing the above code.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Abc_NtkMiniAigTest(pNtk *Abc_Ntk_t) {
    var pNtkNew *Abc_Ntk_t;
    var p *Mini_Aig_t;
    p = Abc_NtkToMiniAig( pNtk );
    pNtkNew = Abc_NtkFromMiniAig( p );
    Mini_AigStop( p );
    Abc_NtkPrintStats( pNtkNew, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0 );
    Abc_NtkDelete( pNtkNew );

    // test dumping
    p = Abc_NtkToMiniAig( pNtk );
    Mini_AigDump( p, "miniaig.data" );
    Mini_AigPrintStats( p );
    Mini_AigStop( p );

    p = Mini_AigLoad( "miniaig.data" );
    Mini_AigPrintStats( p );
    Mini_AigStop( p );
}

/*
    if ( pNtk )
    {
        extern void Abc_NtkMiniAigTest( Abc_Ntk_t * pNtk );
        Abc_NtkMiniAigTest( pNtk );

    }
*/

////////////////////////////////////////////////////////////////////////
///                       END OF FILE                                ///
////////////////////////////////////////////////////////////////////////


//ABC_NAMESPACE_IMPL_END

