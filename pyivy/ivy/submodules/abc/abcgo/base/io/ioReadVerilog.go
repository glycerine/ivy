package io

import "fmt"

/**CFile****************************************************************

  FileName    [ioReadVerilog.c]

  SystemName  [ABC: Logic synthesis and verification system.]

  PackageName [Command processing package.]

  Synopsis    [Procedure to read network from file.]

  Author      [Alan Mishchenko]
  
  Affiliation [UC Berkeley]

  Date        [Ver. 1.0. Started - June 20, 2005.]

  Revision    [$Id: ioReadVerilog.c,v 1.00 2005/06/20 00:00:00 alanmi Exp $]

***********************************************************************/

//#include "ioAbc.h"
//#include "base/ver/ver.h"

//ABC_NAMESPACE_IMPL_START


////////////////////////////////////////////////////////////////////////
///                        DECLARATIONS                              ///
////////////////////////////////////////////////////////////////////////

//extern Abc_Des_t * Ver_ParseFile( char * pFileName, Abc_Des_t * pGateLib, int fCheck, int fUseMemMan );

////////////////////////////////////////////////////////////////////////
///                     FUNCTION DEFINITIONS                         ///
////////////////////////////////////////////////////////////////////////

/**Function*************************************************************

  Synopsis    [Reads hierarchical design from the Verilog file.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Io_ReadVerilog(pFileName string, fCheck int32) *Abc_Ntk_t {
    var pNtk, pTemp *Abc_Ntk_t;
    var pDesign *Abc_Des_t;
    var i, RetValue int32;

    // parse the verilog file
    pDesign = Ver_ParseFile( pFileName, nil, fCheck, 1 );
    if pDesign == nil {
        return nil; }

    // detect top-level model
    RetValue = Abc_DesFindTopLevelModels( pDesign );
    pNtk = (Abc_Ntk_t *)Vec_PtrEntry( pDesign.vTops, 0 );
    if RetValue > 1 {
        fmt.Printf("Warning: The design has %d root-level modules: ", Vec_PtrSize(pDesign.vTops));
        // inlining of macro Vec_PtrForEachEntry
        for  i = 0; (i < Vec_PtrSize(pDesign.vTops)) && (((pTemp) = (Abc_Ntk_t *)Vec_PtrEntry(pDesign.vTops, i)), 1); i++  {
            fmt.Printf(" %s", Abc_NtkName(pTemp));
        }
        fmt.Printf("\n");
        fmt.Printf("The first one (%s) will be used.\n", pNtk.pName);
    }

    // extract the master network
    pNtk.pDesign = pDesign;
    pDesign.pManFunc = nil;

    // verify the design for cyclic dependence
    assert(Vec_PtrSize(pDesign.vModules) > 0, "Vec_PtrSize(pDesign.vModules) > 0");
    if Vec_PtrSize(pDesign.vModules) == 1 {
//        printf( "Warning: The design is not hierarchical.\n" );
        Abc_DesFree( pDesign, pNtk );
        pNtk.pDesign = nil;
        pNtk.pSpec = Extra_UtilStrsav( pFileName );
    
    } else {
        // check that there is no cyclic dependency
        Abc_NtkIsAcyclicHierarchy( pNtk );
    }

//Io_WriteVerilog( pNtk, "_temp.v" );
//    Abc_NtkPrintBoxInfo( pNtk );
    return pNtk;
}

////////////////////////////////////////////////////////////////////////
///                       END OF FILE                                ///
////////////////////////////////////////////////////////////////////////



//ABC_NAMESPACE_IMPL_END

