package mvc

/**CFile****************************************************************

  FileName    [mvcMan.c]

  PackageName [MVSIS 2.0: Multi-valued logic synthesis system.]

  Synopsis    [Procedures working with the MVC memory manager.]

  Author      [MVSIS Group]
  
  Affiliation [UC Berkeley]

  Date        [Ver. 1.0. Started - February 1, 2003.]

  Revision    [$Id: mvcMan.c,v 1.3 2003/03/19 19:50:26 alanmi Exp $]

***********************************************************************/

//#include <string.h>
//#include "mvc.h"

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
func Mvc_ManagerStart() *Mvc_Manager_t {
    var p *Mvc_Manager_t;
    // inlining of macro ABC_ALLOC
    p = ((Mvc_Manager_t *) malloc(sizeof(Mvc_Manager_t) * (size_t)(1)));
    memset( p, 0, sizeof(Mvc_Manager_t) );
    p.pMan1 = Extra_MmFixedStart( sizeof(Mvc_Cube_t)                              );
    p.pMan2 = Extra_MmFixedStart( sizeof(Mvc_Cube_t) +     sizeof(Mvc_CubeWord_t) );
    p.pMan4 = Extra_MmFixedStart( sizeof(Mvc_Cube_t) + 3 * sizeof(Mvc_CubeWord_t) );
    p.pManC = Extra_MmFixedStart( sizeof(Mvc_Cover_t) );
    return p;
}

/**Function*************************************************************

  Synopsis    []

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Mvc_ManagerFree(p *Mvc_Manager_t) {
    Extra_MmFixedStop( p.pMan1 );
    Extra_MmFixedStop( p.pMan2 );
    Extra_MmFixedStop( p.pMan4 );
    Extra_MmFixedStop( p.pManC );
    // inlining of macro ABC_FREE
    ((p) ? (free((char *) (p)), (p) = 0) : 0);
}

////////////////////////////////////////////////////////////////////////
///                       END OF FILE                                ///
////////////////////////////////////////////////////////////////////////


//ABC_NAMESPACE_IMPL_END

