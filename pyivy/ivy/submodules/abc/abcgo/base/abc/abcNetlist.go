package abc

/**CFile****************************************************************

  FileName    [abcNetlist.c]

  SystemName  [ABC: Logic synthesis and verification system.]

  PackageName [Network and node package.]

  Synopsis    [Transforms netlist into a logic network and vice versa.]

  Author      [Alan Mishchenko]

  Affiliation [UC Berkeley]

  Date        [Ver. 1.0. Started - June 20, 2005.]

  Revision    [$Id: abcNetlist.c,v 1.00 2005/06/20 00:00:00 alanmi Exp $]

***********************************************************************/

//#include "abc.h"
//#include "base/main/main.h"
//#include "seq.h"

//ABC_NAMESPACE_IMPL_START

////////////////////////////////////////////////////////////////////////
///                        DECLARATIONS                              ///
////////////////////////////////////////////////////////////////////////

// static void Abc_NtkAddPoBuffers( Abc_Ntk_t * pNtk );

// static Abc_Ntk_t * Abc_NtkLogicToNetlist( Abc_Ntk_t * pNtk );

// static Abc_Ntk_t * Abc_NtkAigToLogicSop( Abc_Ntk_t * pNtk );

// static Abc_Ntk_t * Abc_NtkAigToLogicSopBench( Abc_Ntk_t * pNtk );

////////////////////////////////////////////////////////////////////////
///                     FUNCTION DEFINITIONS                         ///
////////////////////////////////////////////////////////////////////////

/**Function*************************************************************

  Synopsis    [Transform the netlist into a logic network.]

  Description []

  SideEffects []

  SeeAlso     []

***********************************************************************/

/**Function*************************************************************

  Synopsis    [Transform the logic network into a netlist.]

  Description []

  SideEffects []

  SeeAlso     []

***********************************************************************/

/**Function*************************************************************

  Synopsis    [Converts the AIG into the netlist.]

  Description [This procedure does not copy the choices.]

  SideEffects []

  SeeAlso     []

***********************************************************************/

/**Function*************************************************************

  Synopsis    [Transform the logic network into a netlist.]

  Description [The logic network given to this procedure should
  have exactly the same structure as the resulting netlist. The COs
  can only point to CIs if they have identical names. Otherwise,
  they should have a node between them, even if this node is
  inverter or buffer.]

  SideEffects []

  SeeAlso     []

***********************************************************************/

/**Function*************************************************************

  Synopsis    [Converts the AIG into the logic network with SOPs.]

  Description [Correctly handles the case of choice nodes.]

  SideEffects []

  SeeAlso     []

***********************************************************************/

/**Function*************************************************************

  Synopsis    [Converts the AIG into the logic network with SOPs for bench writing.]

  Description [This procedure does not copy the choices.]

  SideEffects []

  SeeAlso     []

***********************************************************************/

/**Function*************************************************************

  Synopsis    [Converts the AIG into the logic network with SOPs for bench writing.]

  Description [This procedure does not copy the choices.]

  SideEffects []

  SeeAlso     []

***********************************************************************/

/**Function*************************************************************

  Synopsis    [Adds buffers for each PO.]

  Description []

  SideEffects []

  SeeAlso     []

***********************************************************************/

////////////////////////////////////////////////////////////////////////
///                       END OF FILE                                ///
////////////////////////////////////////////////////////////////////////

//ABC_NAMESPACE_IMPL_END
