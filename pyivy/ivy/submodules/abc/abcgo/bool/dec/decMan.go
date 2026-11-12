package dec

/**CFile****************************************************************

  FileName    [decMan.c]

  PackageName [MVSIS 2.0: Multi-valued logic synthesis system.]

  Synopsis    [Decomposition manager.]

  Author      [MVSIS Group]
  
  Affiliation [UC Berkeley]

  Date        [Ver. 1.0. Started - February 1, 2003.]

  Revision    [$Id: decMan.c,v 1.1 2003/05/22 19:20:05 alanmi Exp $]

***********************************************************************/

//#include "base/abc/abc.h"
//#include "misc/mvc/mvc.h"
//#include "dec.h"

//ABC_NAMESPACE_IMPL_START


////////////////////////////////////////////////////////////////////////
///                        DECLARATIONS                              ///
////////////////////////////////////////////////////////////////////////

////////////////////////////////////////////////////////////////////////
///                     FUNCTION DEFINITIONS                         ///
////////////////////////////////////////////////////////////////////////

/**Function*************************************************************

  Synopsis    [Start the MVC manager used in the factoring package.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Dec_ManStart() *Dec_Man_t {
    var p *Dec_Man_t;
//    abctime clk = Abc_Clock();
    // inlining of macro ABC_ALLOC
    p = ((Dec_Man_t *) malloc(sizeof(Dec_Man_t) * (size_t)(1)));
    p.pMvcMem = Mvc_ManagerStart();
    p.vCubes = Vec_IntAlloc( 8 );
    p.vLits = Vec_IntAlloc( 8 );
    // canonical forms, phases, perms
    Extra_Truth4VarNPN( &p.puCanons, &p.pPhases, &p.pPerms, &p.pMap );
//ABC_PRT( "NPN classes precomputation time", Abc_Clock() - clk ); 
    return p;
}

/**Function*************************************************************

  Synopsis    [Stops the MVC maanager used in the factoring package.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Dec_ManStop(p *Dec_Man_t) {
    Mvc_ManagerFree( (Mvc_Manager_t *)p.pMvcMem );
    Vec_IntFree( p.vCubes );
    Vec_IntFree( p.vLits );
    // inlining of macro ABC_FREE
    ((p.puCanons) ? (free((char *) (p.puCanons)), (p.puCanons) = 0) : 0);
    // inlining of macro ABC_FREE
    ((p.pPhases) ? (free((char *) (p.pPhases)), (p.pPhases) = 0) : 0);
    // inlining of macro ABC_FREE
    ((p.pPerms) ? (free((char *) (p.pPerms)), (p.pPerms) = 0) : 0);
    // inlining of macro ABC_FREE
    ((p.pMap) ? (free((char *) (p.pMap)), (p.pMap) = 0) : 0);
    // inlining of macro ABC_FREE
    ((p) ? (free((char *) (p)), (p) = 0) : 0);
}

////////////////////////////////////////////////////////////////////////
///                       END OF FILE                                ///
////////////////////////////////////////////////////////////////////////


//ABC_NAMESPACE_IMPL_END

