package io

import "fmt"

import "os"

/**CFile****************************************************************

  FileName    [ioWriteCnf.c]

  SystemName  [ABC: Logic synthesis and verification system.]

  PackageName [Command processing package.]

  Synopsis    [Procedures to output CNF of the miter cone.]

  Author      [Alan Mishchenko]
  
  Affiliation [UC Berkeley]

  Date        [Ver. 1.0. Started - June 20, 2005.]

  Revision    [$Id: ioWriteCnf.c,v 1.00 2005/06/20 00:00:00 alanmi Exp $]

***********************************************************************/

//#include "ioAbc.h"
//#include "sat/bsat/satSolver.h"

//ABC_NAMESPACE_IMPL_START


////////////////////////////////////////////////////////////////////////
///                        DECLARATIONS                              ///
////////////////////////////////////////////////////////////////////////
 
var s_pNtk *Abc_Ntk_t = nil;

////////////////////////////////////////////////////////////////////////
///                     FUNCTION DEFINITIONS                         ///
////////////////////////////////////////////////////////////////////////

/**Function*************************************************************

  Synopsis    [Write the miter cone into a CNF file for the SAT solver.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Io_WriteCnf(pNtk *Abc_Ntk_t, pFileName string, fAllPrimes int32) int32 {
    var pSat *sat_solver;
    if Abc_NtkIsStrash(pNtk) {
        fmt.Printf("Io_WriteCnf() warning: Generating CNF by applying heuristic AIG to CNF conversion.\n");
    } else {
        fmt.Printf("Io_WriteCnf() warning: Generating CNF by convering logic nodes into CNF clauses.\n"); }
    if Abc_NtkPoNum(pNtk) != 1 {
        fmt.Fprintf(os.Stdout, "Io_WriteCnf(): Currently can only process the miter (the network with one PO).\n");
        return 0;
    }
    if Abc_NtkLatchNum(pNtk) != 0 {
        fmt.Fprintf(os.Stdout, "Io_WriteCnf(): Currently can only process the miter for combinational circuits.\n");
        return 0;
    }
    if Abc_NtkNodeNum(pNtk) == 0 {
        fmt.Fprintf(os.Stdout, "The network has no logic nodes. No CNF file is generaled.\n");
        return 0;
    }
    // convert to logic BDD network
    if Abc_NtkIsLogic(pNtk) {
        Abc_NtkToBdd( pNtk ); }
    // create solver with clauses
    pSat = (sat_solver *)Abc_NtkMiterSatCreate( pNtk, fAllPrimes );
    if pSat == nil {
        fmt.Fprintf(os.Stdout, "The problem is trivially UNSAT. No CNF file is generated.\n");
        return 1;
    }        
    // write the clauses
    s_pNtk = pNtk;
    Sat_SolverWriteDimacs( pSat, pFileName, 0, 0, 1 );
    s_pNtk = nil;
    // free the solver
    sat_solver_delete( pSat );
    return 1;
}

/**Function*************************************************************

  Synopsis    [Output the mapping of PIs into variable numbers.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Io_WriteCnfOutputPiMapping(pFile *os.File, incrementVars int32) {
    extern Vec_Int_t * Abc_NtkGetCiSatVarNums( Abc_Ntk_t * pNtk );
    var pNtk *Abc_Ntk_t  = s_pNtk;
    var vCiIds *Vec_Int_t;
    var pObj *Abc_Obj_t;
    var i int32;
    vCiIds = Abc_NtkGetCiSatVarNums( pNtk );
    fmt.Fprintf(pFile, "c PI variable numbers: <PI_name> <SAT_var_number>\n");
    // inlining of macro Abc_NtkForEachCi
    for  i = 0; (i < Abc_NtkCiNum(pNtk)) && (((pObj) = Abc_NtkCi(pNtk, i)), 1); i++  {
        fmt.Fprintf(pFile, "c %s %d\n", Abc_ObjName(pObj), Vec_IntEntry(vCiIds, i) + int32((incrementVars > 0)));
    }
    Vec_IntFree( vCiIds );
}

////////////////////////////////////////////////////////////////////////
///                       END OF FILE                                ///
////////////////////////////////////////////////////////////////////////


//ABC_NAMESPACE_IMPL_END

