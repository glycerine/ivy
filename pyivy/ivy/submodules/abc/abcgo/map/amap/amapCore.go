package amap

import (
	"fmt"
	"time"
)

/**CFile****************************************************************

  FileName    [amapCore.c]

  SystemName  [ABC: Logic synthesis and verification system.]

  PackageName [Technology mapper for standard cells.]

  Synopsis    [Core mapping procedures.]

  Author      [Alan Mishchenko]
  
  Affiliation [UC Berkeley]

  Date        [Ver. 1.0. Started - June 20, 2005.]

  Revision    [$Id: amapCore.c,v 1.00 2005/06/20 00:00:00 alanmi Exp $]

***********************************************************************/

//#include "amapInt.h"
//#include "base/main/main.h"

//ABC_NAMESPACE_IMPL_START


////////////////////////////////////////////////////////////////////////
///                        DECLARATIONS                              ///
////////////////////////////////////////////////////////////////////////

////////////////////////////////////////////////////////////////////////
///                     FUNCTION DEFINITIONS                         ///
////////////////////////////////////////////////////////////////////////

/**Function*************************************************************

  Synopsis    [This procedure sets default parameters.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Amap_ManSetDefaultParams(p *Amap_Par_t) {
    memset( p, 0, sizeof(Amap_Par_t) );
    p.nIterFlow = 1;            // iterations of area flow
    p.nIterArea = 4;            // iteratoins of exact area
    p.nCutsMax  = 500;          // the maximum number of cuts at a node
    p.fUseMuxes = 0;            // enables the use of MUXes
    p.fUseXors  = 1;            // enables the use of XORs
    p.fFreeInvs = 0;            // assume inverters are free (area = 0)
    p.fEpsilon  = float32(0.001); // used to compare floating point numbers
    p.fVerbose  = 0;            // verbosity flag
}

/**Function*************************************************************

  Synopsis    []

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Amap_ManTest(pAig *Aig_Man_t, pPars *Amap_Par_t) *Vec_Ptr_t {
//    extern void * Abc_FrameReadLibGen2();
    var vRes *Vec_Ptr_t;
    var p *Amap_Man_t;
    var pLib *Amap_Lib_t;
    var clkTotal abctime  = Abc_Clock();
    pLib = (Amap_Lib_t *)Abc_FrameReadLibGen2();
    if pLib == nil {
        fmt.Printf("Library is not available.\n");
        return nil;
    }
    p = Amap_ManStart( Aig_ManNodeNum(pAig) );
    p.pPars = pPars;
    p.pLib  = pLib;
    if pPars.fFreeInvs {
        p.fAreaInv = 0.0;
    } else {
        p.fAreaInv = pLib.pGateInv.dArea;
    }
    p.fUseMux = pPars.fUseMuxes && pLib.fHasMux;
    p.fUseXor = pPars.fUseXors && pLib.fHasXor;
    // inlining of macro ABC_CALLOC
    p.ppCutsTemp = ((Amap_Cut_t * *) calloc((size_t)(2 * pLib.nNodes), sizeof(Amap_Cut_t *)));
    // inlining of macro ABC_CALLOC
    p.pMatsTemp = ((int *) calloc((size_t)(2 * pLib.nNodes), sizeof(int)));
    Amap_ManCreate( p, pAig );
    Amap_ManMap( p );
    vRes = nil;
    vRes = Amap_ManProduceMapped( p );
    Amap_ManStop( p );
if pPars.fVerbose {
// inlining of macro ABC_PRT
Abc_Print(1, "%s =", ("Total runtime"));
Abc_Print(1, "%9.2f sec\n", 1.0*(float64((Abc_Clock() - clkTotal)))/(float64(int64(time.Second))));;
}
    return vRes;
}


////////////////////////////////////////////////////////////////////////
///                       END OF FILE                                ///
////////////////////////////////////////////////////////////////////////


//ABC_NAMESPACE_IMPL_END

