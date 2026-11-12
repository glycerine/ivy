package ssw

/**CFile****************************************************************

  FileName    [sswSimSat.c]

  SystemName  [ABC: Logic synthesis and verification system.]

  PackageName [Inductive prover with constraints.]

  Synopsis    [Performs resimulation using counter-examples.]

  Author      [Alan Mishchenko]
  
  Affiliation [UC Berkeley]

  Date        [Ver. 1.0. Started - September 1, 2008.]

  Revision    [$Id: sswSimSat.c,v 1.00 2008/09/01 00:00:00 alanmi Exp $]

***********************************************************************/

//#include "sswInt.h"

//ABC_NAMESPACE_IMPL_START


////////////////////////////////////////////////////////////////////////
///                        DECLARATIONS                              ///
////////////////////////////////////////////////////////////////////////

////////////////////////////////////////////////////////////////////////
///                     FUNCTION DEFINITIONS                         ///
////////////////////////////////////////////////////////////////////////

/**Function*************************************************************

  Synopsis    [Handle the counter-example.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Ssw_ManResimulateBit(p *Ssw_Man_t, pCand *Aig_Obj_t, pRepr *Aig_Obj_t) {
    var pObj *Aig_Obj_t;
    var i, RetValue1, RetValue2 int32;
    var clk abctime  = Abc_Clock();
    // set the PI simulation information
    Aig_ManConst1(p.pAig).fMarkB = 1;
    // inlining of macro Aig_ManForEachCi
    // inlining of macro Vec_PtrForEachEntry
    for  i = 0; (i < Vec_PtrSize(p.pAig.vCis)) && (((pObj) = (Aig_Obj_t *)Vec_PtrEntry(p.pAig.vCis, i)), 1); i++  {
        pObj.fMarkB = Abc_InfoHasBit( p.pPatWords, i );
    }
    // simulate internal nodes
    // inlining of macro Aig_ManForEachNode
    // inlining of macro Vec_PtrForEachEntry
    for  i = 0; (i < Vec_PtrSize(p.pAig.vObjs)) && (((pObj) = (Aig_Obj_t *)Vec_PtrEntry(p.pAig.vObjs, i)), 1); i++  { if (pObj) == nil || !Aig_ObjIsNode(pObj) {} else {
        pObj.fMarkB = ( Aig_ObjFanin0(pObj).fMarkB ^ Aig_ObjFaninC0(pObj) )
                     & ( Aig_ObjFanin1(pObj).fMarkB ^ Aig_ObjFaninC1(pObj) ); }
    }
    // if repr is given, perform refinement
    if pRepr != nil {
        // check equivalence classes
        RetValue1 = Ssw_ClassesRefineConst1( p.ppClasses, 0 );
        RetValue2 = Ssw_ClassesRefine( p.ppClasses, 0 );
        // make sure refinement happened
        if Aig_ObjIsConst1(pRepr) {
            assert(RetValue1, "RetValue1");
            if RetValue1 == 0 {
                Abc_Print( 1, "\nSsw_ManResimulateBit() Error: RetValue1 does not hold.\n" ); }
        
        } else {
            assert((RetValue2 != 0), "RetValue2");
            if RetValue2 == 0 {
                Abc_Print( 1, "\nSsw_ManResimulateBit() Error: RetValue2 does not hold.\n" ); }
        }
    }
p.timeSimSat += Abc_Clock() - clk;
}

/**Function*************************************************************

  Synopsis    [Handle the counter-example.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Ssw_ManResimulateWord(p *Ssw_Man_t, pCand *Aig_Obj_t, pRepr *Aig_Obj_t, f int32) {
    var RetValue1, RetValue2 int32;
    var clk abctime  = Abc_Clock();
    // set the PI simulation information
    Ssw_SmlAssignDist1Plus( p.pSml, p.pPatWords );
    // simulate internal nodes
    Ssw_SmlSimulateOne( p.pSml );
    // check equivalence classes
    RetValue1 = Ssw_ClassesRefineConst1( p.ppClasses, 1 );
    RetValue2 = Ssw_ClassesRefine( p.ppClasses, 1 );
    // make sure refinement happened
    if Aig_ObjIsConst1(pRepr) {
        assert(RetValue1, "RetValue1");
        if RetValue1 == 0 {
            Abc_Print( 1, "\nSsw_ManResimulateWord() Error: RetValue1 does not hold.\n" ); }
    
    } else {
        assert((RetValue2 != 0), "RetValue2");
        if RetValue2 == 0 {
            Abc_Print( 1, "\nSsw_ManResimulateWord() Error: RetValue2 does not hold.\n" ); }
    }
p.timeSimSat += Abc_Clock() - clk;
}

////////////////////////////////////////////////////////////////////////
///                       END OF FILE                                ///
////////////////////////////////////////////////////////////////////////


//ABC_NAMESPACE_IMPL_END
