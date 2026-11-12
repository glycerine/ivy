package hop

import "fmt"

/**CFile****************************************************************

  FileName    [cudd2.c]

  SystemName  [ABC: Logic synthesis and verification system.]

  PackageName [Minimalistic And-Inverter Graph package.]

  Synopsis    [Recording AIGs for the BDD operations.]

  Author      [Alan Mishchenko]
  
  Affiliation [UC Berkeley]

  Date        [Ver. 1.0. Started - October 3, 2006.]

  Revision    [$Id: cudd2.c,v 1.00 2006/10/03 00:00:00 alanmi Exp $]

***********************************************************************/

//#include "hop.h"
//#include "misc/st/st.h"

//ABC_NAMESPACE_IMPL_START


////////////////////////////////////////////////////////////////////////
///                        DECLARATIONS                              ///
////////////////////////////////////////////////////////////////////////

          
type Aig_CuddMan_t struct {
    pAig   *Aig_Man_t;   // internal AIG package
    pTable    *st__table; // hash table mapping BDD nodes into AIG nodes
}

// static Cudd AIG manager used in this experiment
var s_pCuddMan *Aig_CuddMan_t = nil;

////////////////////////////////////////////////////////////////////////
///                     FUNCTION DEFINITIONS                         ///
////////////////////////////////////////////////////////////////////////

/**Function*************************************************************

  Synopsis    [Start AIG recording.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
void Cudd2_Init( unsigned int numVars, unsigned int numVarsZ, unsigned int numSlots, unsigned int cacheSize, unsigned long maxMemory, void * pCudd )
{
    var v int32;
    // start the BDD-to-AIG manager when the first BDD manager is allocated
    if s_pCuddMan != nil {
        return; }
    s_pCuddMan = ALLOC( Aig_CuddMan_t, 1 );
    s_pCuddMan.pAig = Aig_ManStart();
    s_pCuddMan.pTable = st__init_table( st__ptrcmp, st__ptrhash );
    for  v = 0; v < int32(numVars); v++  {
        Aig_ObjCreatePi( s_pCuddMan.pAig );
    }
}

/**Function*************************************************************

  Synopsis    [Stops AIG recording.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Cudd2_Quit(pCudd any) {
    assert(s_pCuddMan != nil, "s_pCuddMan != nil");
    Aig_ManDumpBlif( s_pCuddMan.pAig, "aig_temp.blif", nil, nil );
    Aig_ManStop( s_pCuddMan.pAig );
    st__free_table( s_pCuddMan.pTable );
    free( s_pCuddMan );
    s_pCuddMan = nil;
}

/**Function*************************************************************

  Synopsis    [Fetches AIG node corresponding to the BDD node from the hash table.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Cudd2_GetArg(pArg any) *Aig_Obj_t {
    var pNode *Aig_Obj_t;
    assert(s_pCuddMan != nil, "s_pCuddMan != nil");
    if ! st__lookup( s_pCuddMan.pTable, (char *)Aig_Regular(pArg), (char **)&pNode ) {
        fmt.Printf("Cudd2_GetArg(): An argument BDD is not in the hash table.\n");
        return nil;
    }
    return Aig_NotCond( pNode, Aig_IsComplement(pArg) );
}

/**Function*************************************************************

  Synopsis    [Inserts the AIG node corresponding to the BDD node into the hash table.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Cudd2_SetArg(pNode *Aig_Obj_t, pResult any) {
    assert(s_pCuddMan != nil, "s_pCuddMan != nil");
    // inlining of macro st__is_member
    if  st__lookup(s_pCuddMan.pTable,(char *)Aig_Regular(pResult),(char **) 0) {
        return; }
    pNode = Aig_NotCond( pNode,  Aig_IsComplement(pResult) );
    st__insert( s_pCuddMan.pTable, (char *)Aig_Regular(pResult), (char *)pNode );
}

/**Function*************************************************************

  Synopsis    [Registers constant 1 node.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Cudd2_bddOne(pCudd any, pResult any) {
    Cudd2_SetArg( Aig_ManConst1(s_pCuddMan.pAig), pResult );
}

/**Function*************************************************************

  Synopsis    [Adds elementary variable.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Cudd2_bddIthVar(pCudd any, iVar int32, pResult any) {
    var v int32;
    assert(s_pCuddMan != nil, "s_pCuddMan != nil");
    for  v = Aig_ManPiNum(s_pCuddMan.pAig); v <= iVar; v++  {
        Aig_ObjCreatePi( s_pCuddMan.pAig );
    }
    Cudd2_SetArg( Aig_ManPi(s_pCuddMan.pAig, iVar), pResult );
}

/**Function*************************************************************

  Synopsis    [Performs BDD operation.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Cudd2_bddAnd(pCudd any, pArg0 any, pArg1 any, pResult any) {
    var pNode0, pNode1, pNode *Aig_Obj_t;
    pNode0 = Cudd2_GetArg( pArg0 );
    pNode1 = Cudd2_GetArg( pArg1 );
    pNode  = Aig_And( s_pCuddMan.pAig, pNode0, pNode1 );
    Cudd2_SetArg( pNode, pResult );
}

/**Function*************************************************************

  Synopsis    [Performs BDD operation.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Cudd2_bddOr(pCudd any, pArg0 any, pArg1 any, pResult any) {
    Cudd2_bddAnd( pCudd, Aig_Not(pArg0), Aig_Not(pArg1), Aig_Not(pResult) );
}

/**Function*************************************************************

  Synopsis    [Performs BDD operation.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Cudd2_bddNand(pCudd any, pArg0 any, pArg1 any, pResult any) {
    Cudd2_bddAnd( pCudd, pArg0, pArg1, Aig_Not(pResult) );
}

/**Function*************************************************************

  Synopsis    [Performs BDD operation.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Cudd2_bddNor(pCudd any, pArg0 any, pArg1 any, pResult any) {
    Cudd2_bddAnd( pCudd, Aig_Not(pArg0), Aig_Not(pArg1), pResult );
}

/**Function*************************************************************

  Synopsis    [Performs BDD operation.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Cudd2_bddXor(pCudd any, pArg0 any, pArg1 any, pResult any) {
    var pNode0, pNode1, pNode *Aig_Obj_t;
    pNode0 = Cudd2_GetArg( pArg0 );
    pNode1 = Cudd2_GetArg( pArg1 );
    pNode  = Aig_Exor( s_pCuddMan.pAig, pNode0, pNode1 );
    Cudd2_SetArg( pNode, pResult );
}

/**Function*************************************************************

  Synopsis    [Performs BDD operation.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Cudd2_bddXnor(pCudd any, pArg0 any, pArg1 any, pResult any) {
    Cudd2_bddXor( pCudd, pArg0, pArg1, Aig_Not(pResult) );
}

/**Function*************************************************************

  Synopsis    [Performs BDD operation.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Cudd2_bddIte(pCudd any, pArg0 any, pArg1 any, pArg2 any, pResult any) {
    var pNode0, pNode1, pNode2, pNode *Aig_Obj_t;
    pNode0 = Cudd2_GetArg( pArg0 );
    pNode1 = Cudd2_GetArg( pArg1 );
    pNode2 = Cudd2_GetArg( pArg2 );
    pNode  = Aig_Mux( s_pCuddMan.pAig, pNode0, pNode1, pNode2 );
    Cudd2_SetArg( pNode, pResult );
}

/**Function*************************************************************

  Synopsis    [Performs BDD operation.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Cudd2_bddCompose(pCudd any, pArg0 any, pArg1 any, v int32, pResult any) {
    var pNode0, pNode1, pNode *Aig_Obj_t;
    pNode0 = Cudd2_GetArg( pArg0 );
    pNode1 = Cudd2_GetArg( pArg1 );
    pNode  = Aig_Compose( s_pCuddMan.pAig, pNode0, pNode1, v );
    Cudd2_SetArg( pNode, pResult );
}

/**Function*************************************************************

  Synopsis    [Should be called after each containment check.]

  Description [Result should be 1 if Cudd2_bddLeq returned 1.]
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Cudd2_bddLeq(pCudd any, pArg0 any, pArg1 any, Result int32) {
    var pNode0, pNode1, pNode *Aig_Obj_t;
    pNode0 = Cudd2_GetArg( pArg0 );
    pNode1 = Cudd2_GetArg( pArg1 );
    pNode  = Aig_And( s_pCuddMan.pAig, pNode0, Aig_Not(pNode1) );
    Aig_ObjCreatePo( s_pCuddMan.pAig, pNode );
}

/**Function*************************************************************

  Synopsis    [Should be called after each equality check.]

  Description [Result should be 1 if they are equal.]
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Cudd2_bddEqual(pCudd any, pArg0 any, pArg1 any, Result int32) {
    var pNode0, pNode1, pNode *Aig_Obj_t;
    pNode0 = Cudd2_GetArg( pArg0 );
    pNode1 = Cudd2_GetArg( pArg1 );
    pNode  = Aig_Exor( s_pCuddMan.pAig, pNode0, pNode1 );
    Aig_ObjCreatePo( s_pCuddMan.pAig, pNode );
}

////////////////////////////////////////////////////////////////////////
///                       END OF FILE                                ///
////////////////////////////////////////////////////////////////////////


//ABC_NAMESPACE_IMPL_END

