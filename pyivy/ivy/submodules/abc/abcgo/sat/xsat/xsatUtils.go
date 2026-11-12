package xsat

/**CFile****************************************************************

  FileName    [xsatUtils.h]

  SystemName  [ABC: Logic synthesis and verification system.]

  PackageName [xSAT - A SAT solver written in C.
               Read the license file for more info.]

  Synopsis    [Utility functions used in xSAT]

  Author      [Bruno Schmitt <boschmitt@inf.ufrgs.br>]

  Affiliation [UC Berkeley / UFRGS]

  Date        [Ver. 1.0. Started - November 10, 2016.]

  Revision    []

***********************************************************************/



////////////////////////////////////////////////////////////////////////
///                          INCLUDES                                ///
////////////////////////////////////////////////////////////////////////
//#include "misc/util/abc_global.h"

//ABC_NAMESPACE_HEADER_START

/**Function*************************************************************

  Synopsis    []

  Description []

  SideEffects []

  SeeAlso     []

***********************************************************************/
func xSAT_UtilSelectSort(pArray *any, nSize int32, CompFnct func(any, any) int32) {
    var i, j, iBest int32;
    var pTmp any;

    for  i = 0; i < ( nSize - 1 ); i++  {
        iBest = i;
        for  j = i + 1; j < nSize; j++  {
            if CompFnct( pArray[j], pArray[iBest] ) {
                iBest = j; }
        }
        pTmp = pArray[i];
        pArray[i] = pArray[iBest];
        pArray[iBest] = pTmp;
    }
}

/**Function*************************************************************

  Synopsis    []

  Description []

  SideEffects []

  SeeAlso     []

***********************************************************************/
func xSAT_UtilSort(pArray *any, nSize int32, CompFnct func(any, any) int32) {
    if nSize <= 15 {
        xSAT_UtilSelectSort( pArray, nSize, CompFnct );
    } else {
        var pPivot any  = pArray[nSize / 2];
        var pTmp any;
        var i int32  = -1;
        var j int32  = nSize;

        for {
            do i++; while( CompFnct( pArray[i], pPivot ) );
            do j--; while( CompFnct( pPivot, pArray[j] ) );

            if i >= j {
                break; }

            pTmp = pArray[i];
            pArray[i] = pArray[j];
            pArray[j] = pTmp;
        }

        xSAT_UtilSort( pArray, i, CompFnct );
        xSAT_UtilSort( pArray + i, ( nSize - i ), CompFnct );
    }
}

//ABC_NAMESPACE_HEADER_END


////////////////////////////////////////////////////////////////////////
///                       END OF FILE                                ///
////////////////////////////////////////////////////////////////////////
