package if_

/**CFile****************************************************************

  FileName    [ifMatch2.c]

  SystemName  [ABC: Logic synthesis and verification system.]

  PackageName [FPGA mapping based on priority cuts.]

  Synopsis    [Specialized matching.]

  Author      [Alan Mishchenko]
  
  Affiliation [UC Berkeley]

  Date        [Ver. 1.0. Started - September 1, 2009.]

  Revision    [$Id: ifMatch2.c,v 1.00 2009/09/01 00:00:00 alanmi Exp $]

***********************************************************************/

//#include "if.h"

//ABC_NAMESPACE_IMPL_START


////////////////////////////////////////////////////////////////////////
///                        DECLARATIONS                              ///
////////////////////////////////////////////////////////////////////////

////////////////////////////////////////////////////////////////////////
///                     FUNCTION DEFINITIONS                         ///
////////////////////////////////////////////////////////////////////////

/**Function*************************************************************

  Synopsis    []

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Bat_ManFuncSetupTable() {
}
func Bat_ManFuncSetdownTable() {
}
func Bat_ManCellFuncLookup(pMan any, pTruth *uint32, nVars int32, nLeaves int32, pStr string) int32 {
    return 1;
}

////////////////////////////////////////////////////////////////////////
///                       END OF FILE                                ///
////////////////////////////////////////////////////////////////////////


//ABC_NAMESPACE_IMPL_END

