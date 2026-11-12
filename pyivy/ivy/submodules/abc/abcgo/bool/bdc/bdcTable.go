package bdc

/**CFile****************************************************************

  FileName    [bdcTable.c]

  SystemName  [ABC: Logic synthesis and verification system.]

  PackageName [Truth-table-based bi-decomposition engine.]

  Synopsis    [Hash table for intermediate nodes.]

  Author      [Alan Mishchenko]
  
  Affiliation [UC Berkeley]

  Date        [Ver. 1.0. Started - January 30, 2007.]

  Revision    [$Id: bdcTable.c,v 1.00 2007/01/30 00:00:00 alanmi Exp $]

***********************************************************************/

//#include "bdcInt.h"

//ABC_NAMESPACE_IMPL_START


////////////////////////////////////////////////////////////////////////
///                        DECLARATIONS                              ///
////////////////////////////////////////////////////////////////////////

////////////////////////////////////////////////////////////////////////
///                     FUNCTION DEFINITIONS                         ///
////////////////////////////////////////////////////////////////////////

/**Function*************************************************************

  Synopsis    [Checks containment of the function in the ISF.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Bdc_TableCheckContainment(p *Bdc_Man_t, pIsf *Bdc_Isf_t, puTruth *uint32) int32 {
    return Kit_TruthIsImply( pIsf.puOn, puTruth, p.nVars ) &&
         Kit_TruthIsDisjoint( puTruth, pIsf.puOff, p.nVars );
}

/**Function*************************************************************

  Synopsis    [Adds the new entry to the hash table.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Bdc_TableLookup(p *Bdc_Man_t, pIsf *Bdc_Isf_t) *Bdc_Fun_t {
    var fDisableCache int32;
    var pFunc *Bdc_Fun_t;
    if (fDisableCache != 0) && Kit_WordCountOnes(pIsf.uSupp) > 1 {
        return nil; }
    if pIsf.uSupp == 0 {
        assert(p.pTable[pIsf.uSupp] == p.pNodes, "p.pTable[pIsf.uSupp] == p.pNodes");
        if Kit_TruthIsConst1( pIsf.puOn, p.nVars ) {
            return p.pNodes; }
        assert(Kit_TruthIsConst1( pIsf.puOff, p.nVars ), "Kit_TruthIsConst1( pIsf.puOff, p.nVars )");
        return Bdc_Not(p.pNodes);
    }
    for  pFunc = p.pTable[pIsf.uSupp]; (pFunc != nil); pFunc = pFunc.pNext  {
        if Bdc_TableCheckContainment( p, pIsf, pFunc.puFunc ) != 0 {
             return pFunc; }
    }
    Bdc_IsfNot( pIsf );
    for  pFunc = p.pTable[pIsf.uSupp]; (pFunc != nil); pFunc = pFunc.pNext  {
        if Bdc_TableCheckContainment( p, pIsf, pFunc.puFunc ) != 0 {
            Bdc_IsfNot( pIsf );
            return Bdc_Not(pFunc);
        }
    }
    Bdc_IsfNot( pIsf );
    return nil;
}

/**Function*************************************************************

  Synopsis    [Adds the new entry to the hash table.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Bdc_TableAdd(p *Bdc_Man_t, pFunc *Bdc_Fun_t) {
    if p.pTable[pFunc.uSupp] == nil {
        Vec_IntPush( p.vSpots, pFunc.uSupp ); }
    pFunc.pNext = p.pTable[pFunc.uSupp];
    p.pTable[pFunc.uSupp] = pFunc;
}

/**Function*************************************************************

  Synopsis    [Adds the new entry to the hash table.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Bdc_TableClear(p *Bdc_Man_t) {
    var Spot, i int32;
    Vec_IntForEachEntry( p.vSpots, Spot, i )
        p.pTable[Spot] = nil;
    Vec_IntClear( p.vSpots );
}

////////////////////////////////////////////////////////////////////////
///                       END OF FILE                                ///
////////////////////////////////////////////////////////////////////////


//ABC_NAMESPACE_IMPL_END

