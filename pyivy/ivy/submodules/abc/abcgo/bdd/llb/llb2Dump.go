package llb

import "fmt"

import "os"

/**CFile****************************************************************

  FileName    [llb2Dump.c]

  SystemName  [ABC: Logic synthesis and verification system.]

  PackageName [BDD based reachability.]

  Synopsis    [Dumps the BDD of reached states into a file.]

  Author      [Alan Mishchenko]
  
  Affiliation [UC Berkeley]

  Date        [Ver. 1.0. Started - June 20, 2005.]

  Revision    [$Id: llb2Dump.c,v 1.00 2005/06/20 00:00:00 alanmi Exp $]

***********************************************************************/

//#include "llbInt.h"

//ABC_NAMESPACE_IMPL_START


////////////////////////////////////////////////////////////////////////
///                        DECLARATIONS                              ///
////////////////////////////////////////////////////////////////////////

////////////////////////////////////////////////////////////////////////
///                     FUNCTION DEFINITIONS                         ///
////////////////////////////////////////////////////////////////////////

/**Function*************************************************************

  Synopsis    [Returns a dummy name.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
var spStatic_Llb_ManGetDummyName_Buffer string // was [2000]byte in C

func Llb_ManGetDummyName(pPrefix string, Num int32, nDigits int32) string {
    
    spStatic_Llb_ManGetDummyName_Buffer = fmt.Sprintf("%s%0*d", pPrefix, nDigits, Num);
    return spStatic_Llb_ManGetDummyName_Buffer;
}

/**Function*************************************************************

  Synopsis    [Writes reached state BDD into a BLIF file.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Llb_ManDumpReached(ddG *DdManager, bReached *DdNode, pModel string, pFileName string) {
    var pFile *os.File;
    var vNamesIn, vNamesOut *Vec_Ptr_t;
    var pName string;
    var i, nDigits int32;
    // reorder the BDD
    Cudd_ReduceHeap( ddG, CUDD_REORDER_SYMM_SIFT, 1 );

    // create input names
    nDigits = Abc_Base10Log( Cudd_ReadSize(ddG) );
    vNamesIn = Vec_PtrAlloc( Cudd_ReadSize(ddG) );
    for  i = 0; i < Cudd_ReadSize(ddG); i++  {
        pName = Llb_ManGetDummyName( "ff", i, nDigits );
        Vec_PtrPush( vNamesIn, Extra_UtilStrsav(pName) );
    }
    // create output names
    vNamesOut = Vec_PtrAlloc( 1 );
    Vec_PtrPush( vNamesOut, Extra_UtilStrsav("Reached") );

    // write the file
    pFile = os.OpenFile(pFileName, os.O_WRONLY | os.O_CREATE | os.O_TRUNC, 0666);
    Cudd_DumpBlif( ddG, 1, &bReached, (char **)Vec_PtrArray(vNamesIn), (char **)Vec_PtrArray(vNamesOut), pModel, pFile, 0 );
    _ = (pFile).Close();

    // cleanup
    Vec_PtrForEachEntry( char *, vNamesIn, pName, i )
        ABC_FREE( pName );
    Vec_PtrForEachEntry( char *, vNamesOut, pName, i )
        ABC_FREE( pName );
    Vec_PtrFree( vNamesIn );
    Vec_PtrFree( vNamesOut );
}

////////////////////////////////////////////////////////////////////////
///                       END OF FILE                                ///
////////////////////////////////////////////////////////////////////////


//ABC_NAMESPACE_IMPL_END

