package abci

import "fmt"

/**CFile****************************************************************

  FileName    [abcBmc.c]

  SystemName  [ABC: Logic synthesis and verification system.]

  PackageName [Network and node package.]

  Synopsis    [Performs bounded model check.]

  Author      [Alan Mishchenko]
  
  Affiliation [UC Berkeley]

  Date        [Ver. 1.0. Started - June 20, 2005.]

  Revision    [$Id: abcBmc.c,v 1.00 2005/06/20 00:00:00 alanmi Exp $]

***********************************************************************/

//#include "base/abc/abc.h"
//#include "aig/ivy/ivy.h"

//ABC_NAMESPACE_IMPL_START


////////////////////////////////////////////////////////////////////////
///                        DECLARATIONS                              ///
////////////////////////////////////////////////////////////////////////

// extern Ivy_Man_t * Abc_NtkIvyBefore( Abc_Ntk_t * pNtk, int fSeq, int fUseDc );


// static void Abc_NtkBmcReport( Ivy_Man_t * pMan, Ivy_Man_t * pFrames, Ivy_Man_t * pFraig, Vec_Ptr_t * vMapping, int nFrames );


////////////////////////////////////////////////////////////////////////
///                     FUNCTION DEFINITIONS                         ///
////////////////////////////////////////////////////////////////////////

/**Function*************************************************************

  Synopsis    []

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Abc_NtkBmc(pNtk *Abc_Ntk_t, nFrames int32, fInit int32, fVerbose int32) {
    var Params Ivy_FraigParams_t;
    var pParams *Ivy_FraigParams_t = &Params; 
    var pMan, pFrames, pFraig *Ivy_Man_t;
    var vMapping *Vec_Ptr_t;
    // convert to IVY manager
    pMan = Abc_NtkIvyBefore( pNtk, 0, 0 );
    // generate timeframes
    pFrames = Ivy_ManFrames( pMan, Abc_NtkLatchNum(pNtk), nFrames, fInit, &vMapping );
    // fraig the timeframes
    Ivy_FraigParamsDefault( pParams );
    pParams.nBTLimitNode = ABC_INFINITY;
    pParams.fVerbose = 0;
    pParams.fProve = 0;
    pFraig = Ivy_FraigPerform( pFrames, pParams );
fmt.Printf("Frames have %6d nodes.  ", Ivy_ManNodeNum(pFrames));
fmt.Printf("Fraig has %6d nodes.\n", Ivy_ManNodeNum(pFraig));
    // report the classes
//    if ( fVerbose )
//        Abc_NtkBmcReport( pMan, pFrames, pFraig, vMapping, nFrames );
    // free stuff
    Vec_PtrFree( vMapping );
    Ivy_ManStop( pFraig );
    Ivy_ManStop( pFrames );
    Ivy_ManStop( pMan );
}
 
/**Function*************************************************************

  Synopsis    []

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Abc_NtkBmcReport(pMan *Ivy_Man_t, pFrames *Ivy_Man_t, pFraig *Ivy_Man_t, vMapping *Vec_Ptr_t, nFrames int32) {
    var pFirst1 *Ivy_Obj_t;
    var pFirst2 *Ivy_Obj_t = nil;
    var pFirst3 *Ivy_Obj_t = nil;
    var i, f, nIdMax, Prev2, Prev3 int32;
    nIdMax = Ivy_ManObjIdMax(pMan);
    // check what is the number of nodes in each frame
    Prev2 = Prev3 = 0;
    for  f = 0; f < nFrames; f++  {
        // inlining of macro Ivy_ManForEachNode
        // inlining of macro Ivy_ManForEachObj
        // inlining of macro Vec_PtrForEachEntry
        for  i = 0; (i < Vec_PtrSize(pMan.vObjs)) && (((pFirst1) = (Ivy_Obj_t *)Vec_PtrEntry(pMan.vObjs, i)), 1); i++  { if (pFirst1) == nil {} else if !Ivy_ObjIsNode(pFirst1) {} else {
            pFirst2 = Ivy_Regular( (Ivy_Obj_t *)Vec_PtrEntry(vMapping, f * nIdMax + pFirst1.Id) );
            if Ivy_ObjIsConst1(pFirst2) || pFirst2.Type == 0 {
                continue; }
            pFirst3 = Ivy_Regular( pFirst2.pEquiv );
            if Ivy_ObjIsConst1(pFirst3) || pFirst3.Type == 0 {
                continue; }
            break;
        }
        }
        assert((pFirst2 != nil), "pFirst2");
        assert((pFirst3 != nil), "pFirst3");
        if f != nil {
            fmt.Printf("Frame %3d :  Strash = %5d  Fraig = %5d\n", f, pFirst2.Id - Prev2, pFirst3.Id - Prev3); }
        Prev2 = pFirst2.Id;
        Prev3 = pFirst3.Id;   
    }
}

////////////////////////////////////////////////////////////////////////
///                       END OF FILE                                ///
////////////////////////////////////////////////////////////////////////


//ABC_NAMESPACE_IMPL_END

