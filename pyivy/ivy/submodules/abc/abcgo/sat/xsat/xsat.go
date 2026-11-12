package xsat

/**CFile****************************************************************

  FileName    [xsat.h]

  SystemName  [ABC: Logic synthesis and verification system.]

  PackageName [xSAT - A SAT solver written in C. 
               Read the license file for more info.]

  Synopsis    [External definitions of the solver.]

  Author      [Bruno Schmitt <boschmitt@inf.ufrgs.br>]

  Affiliation [UC Berkeley / UFRGS]

  Date        [Ver. 1.0. Started - November 10, 2016.]

  Revision    []

***********************************************************************/



////////////////////////////////////////////////////////////////////////
///                          INCLUDES                                ///
////////////////////////////////////////////////////////////////////////
//#include "misc/util/abc_global.h"
//#include "misc/vec/vecInt.h"

//ABC_NAMESPACE_HEADER_START

////////////////////////////////////////////////////////////////////////
///                    STRUCTURE DEFINITIONS                         ///
////////////////////////////////////////////////////////////////////////
struct xSAT_Solver_t_;
type xSAT_Solver_t = xSAT_Solver_t_;

////////////////////////////////////////////////////////////////////////
///                     FUNCTION DECLARATIONS                        ///
////////////////////////////////////////////////////////////////////////
/*=== xsatCnfReader.c ================================================*/
// extern int xSAT_SolverParseDimacs( FILE *, xSAT_Solver_t ** );


/*=== xsatSolverAPI.c ================================================*/
// extern xSAT_Solver_t * xSAT_SolverCreate();

// extern void xSAT_SolverDestroy( xSAT_Solver_t * );


// extern int xSAT_SolverAddClause( xSAT_Solver_t *, Vec_Int_t * );

// extern int xSAT_SolverSimplify( xSAT_Solver_t * );

// extern int xSAT_SolverSolve( xSAT_Solver_t * );


// extern void xSAT_SolverPrintStats( xSAT_Solver_t * );


//ABC_NAMESPACE_HEADER_END


////////////////////////////////////////////////////////////////////////
///                       END OF FILE                                ///
////////////////////////////////////////////////////////////////////////
