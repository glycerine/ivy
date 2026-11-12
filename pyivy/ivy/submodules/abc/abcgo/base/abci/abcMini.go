package abci

import (
	"fmt"
	"os"
)

/**CFile****************************************************************

  FileName    [abcMini.c]

  SystemName  [ABC: Logic synthesis and verification system.]

  PackageName [Network and node package.]

  Synopsis    [Interface to the minimalistic AIG package.]

  Author      [Alan Mishchenko]
  
  Affiliation [UC Berkeley]

  Date        [Ver. 1.0. Started - June 20, 2005.]

  Revision    [$Id: abcMini.c,v 1.00 2005/06/20 00:00:00 alanmi Exp $]

***********************************************************************/

//#include "base/abc/abc.h"
//#include "base/main/main.h"
//#include "aig/miniaig/miniaig.h"

//ABC_NAMESPACE_IMPL_START


////////////////////////////////////////////////////////////////////////
///                        DECLARATIONS                              ///
////////////////////////////////////////////////////////////////////////

////////////////////////////////////////////////////////////////////////
///                     FUNCTION DEFINITIONS                         ///
////////////////////////////////////////////////////////////////////////


/**Function*************************************************************

  Synopsis    [Converts the network from the AIG manager into ABC.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Abc_NodeFanin0Copy(pNtk *Abc_Ntk_t, vCopies *Vec_Int_t, p *Mini_Aig_t, Id int32) *Abc_Obj_t {
    var Lit int32  = Mini_AigNodeFanin0( p, Id );
    var AbcLit int32  = Abc_LitNotCond( Vec_IntEntry(vCopies, Abc_Lit2Var(Lit)), Abc_LitIsCompl(Lit) );
    return Abc_ObjFromLit( pNtk, AbcLit );
}
func Abc_NodeFanin1Copy(pNtk *Abc_Ntk_t, vCopies *Vec_Int_t, p *Mini_Aig_t, Id int32) *Abc_Obj_t {
    var Lit int32  = Mini_AigNodeFanin1( p, Id );
    var AbcLit int32  = Abc_LitNotCond( Vec_IntEntry(vCopies, Abc_Lit2Var(Lit)), Abc_LitIsCompl(Lit) );
    return Abc_ObjFromLit( pNtk, AbcLit );
}
func Abc_NtkFromMiniAig(p *Mini_Aig_t) *Abc_Ntk_t {
    var pNtk *Abc_Ntk_t;
    var pObj *Abc_Obj_t  = nil;
    var vCopies *Vec_Int_t;
    var i, nNodes int32;
    // get the number of nodes
    nNodes = Mini_AigNodeNum(p);
    // create ABC network
    pNtk = Abc_NtkAlloc( ABC_NTK_STRASH, ABC_FUNC_AIG, 1 );
    pNtk.pName = Abc_UtilStrsav( "MiniAig" );
    // create mapping from MiniAIG objects into ABC objects
    vCopies = Vec_IntAlloc( nNodes );
    Vec_IntPush( vCopies, Abc_LitNot(Abc_ObjToLit(Abc_AigConst1(pNtk))) );
    // iterate through the objects
    for  i = 1; i < nNodes; i++  {
        if Mini_AigNodeIsPi( p, i ) {
            pObj = Abc_NtkCreatePi(pNtk);
        } else if Mini_AigNodeIsPo( p, i ) {
            Abc_ObjAddFanin( (pObj = Abc_NtkCreatePo(pNtk)), Abc_NodeFanin0Copy(pNtk, vCopies, p, i) );
        } else if Mini_AigNodeIsAnd( p, i ) {
            pObj = Abc_AigAnd((Abc_Aig_t *)pNtk.pManFunc, Abc_NodeFanin0Copy(pNtk, vCopies, p, i), Abc_NodeFanin1Copy(pNtk, vCopies, p, i));
        } else { assert(false, "0"); }
        Vec_IntPush( vCopies, Abc_ObjToLit(pObj) );
    }
    assert(Vec_IntSize(vCopies) == nNodes, "Vec_IntSize(vCopies) == nNodes");
    Abc_AigCleanup( (Abc_Aig_t *)pNtk.pManFunc );
    Vec_IntFree( vCopies );
    Abc_NtkAddDummyPiNames( pNtk );
    Abc_NtkAddDummyPoNames( pNtk );
    if !Abc_NtkCheck( pNtk ) {
        fmt.Fprintf(os.Stdout, "Abc_NtkFromMini(): Network check has failed.\n"); }
    // add latches
    if Mini_AigRegNum(p) > 0 {
        extern Abc_Ntk_t * Abc_NtkRestrashWithLatches( Abc_Ntk_t * pNtk, int nLatches );
        var pTemp *Abc_Ntk_t;
        pNtk = Abc_NtkRestrashWithLatches( pTemp = pNtk, Mini_AigRegNum(p) );
        Abc_NtkDelete( pTemp );
    }
    return pNtk;
}

/**Function*************************************************************

  Synopsis    [Converts the network from ABC into the AIG manager.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Abc_NodeFanin0Copy2(pObj *Abc_Obj_t) int32 {
    return Abc_LitNotCond( Abc_ObjFanin0(pObj).iTemp, Abc_ObjFaninC0(pObj) );
}
func Abc_NodeFanin1Copy2(pObj *Abc_Obj_t) int32 {
    return Abc_LitNotCond( Abc_ObjFanin1(pObj).iTemp, Abc_ObjFaninC1(pObj) );
}
func Abc_NtkToMiniAig(pNtk *Abc_Ntk_t) *Mini_Aig_t {
    var p *Mini_Aig_t;
    var pObj *Abc_Obj_t  = nil;
    var i int32;
    assert(Abc_NtkIsStrash(pNtk), "Abc_NtkIsStrash(pNtk)");
    // create the manager
    p = Mini_AigStart();
    // create mapping from MiniAIG into ABC objects
    Abc_NtkCleanCopy( pNtk );
    Abc_AigConst1(pNtk).iTemp = Mini_AigLitConst1();
    // create primary inputs
    // inlining of macro Abc_NtkForEachCi
    for  i = 0; (i < Abc_NtkCiNum(pNtk)) && (((pObj) = Abc_NtkCi(pNtk, i)), 1); i++  {
        pObj.iTemp = Mini_AigCreatePi(p);
    }
    // create internal nodes
    // inlining of macro Abc_NtkForEachNode
    for  i = 0; (i < Vec_PtrSize((pNtk).vObjs)) && (((pObj) = Abc_NtkObj(pNtk, i)), 1); i++  {           if (pObj) == nil || !Abc_ObjIsNode(pObj) {} else {
        pObj.iTemp = Mini_AigAnd( p, Abc_NodeFanin0Copy2(pObj), Abc_NodeFanin1Copy2(pObj) ); }
    }
    // create primary outputs
    // inlining of macro Abc_NtkForEachCo
    for  i = 0; (i < Abc_NtkCoNum(pNtk)) && (((pObj) = Abc_NtkCo(pNtk, i)), 1); i++  {
        pObj.iTemp = Mini_AigCreatePo( p, Abc_NodeFanin0Copy2(pObj) );
    }
    // set registers
    Mini_AigSetRegNum( p, Abc_NtkLatchNum(pNtk) );
    return p;
}

/**Function*************************************************************

  Synopsis    [Procedures to update internal ABC network using AIG node array.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Abc_NtkInputMiniAig(pAbc *Abc_Frame_t, p any) {
    var pNtk *Abc_Ntk_t;
    if pAbc == nil {
        fmt.Printf("ABC framework is not initialized by calling Abc_Start()\n"); }
    pNtk = Abc_NtkFromMiniAig( (Mini_Aig_t *)p );
    Abc_FrameReplaceCurrentNetwork( pAbc, pNtk );
//    Abc_NtkDelete( pNtk );
}
func Abc_NtkOutputMiniAig(pAbc *Abc_Frame_t) any {
    var pNtk *Abc_Ntk_t;
    if pAbc == nil {
        fmt.Printf("ABC framework is not initialized by calling Abc_Start()\n"); }
    pNtk = Abc_FrameReadNtk( pAbc );
    if pNtk == nil {
        fmt.Printf("Current network in ABC framework is not defined.\n"); }
    return Abc_NtkToMiniAig( pNtk );
}
func Abc_NtkSetFlopNum(pAbc *Abc_Frame_t, nFlops int32) {
    extern void Abc_NtkMakeSeq( Abc_Ntk_t * pNtk, int nFlops );
    var pNtk *Abc_Ntk_t;
    if pAbc == nil {
        fmt.Printf("ABC framework is not initialized by calling Abc_Start()\n"); }
    pNtk = Abc_FrameReadNtk( pAbc );
    if pNtk == nil {
        fmt.Printf("Current network in ABC framework is not defined.\n"); }
    Abc_NtkMakeSeq( pNtk, nFlops );
}

/**Function*************************************************************

  Synopsis    [Testing the above code.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Abc_NtkMiniAigTest(pNtk *Abc_Ntk_t) {
    var pNtkNew *Abc_Ntk_t;
    var p *Mini_Aig_t;
    p = Abc_NtkToMiniAig( pNtk );
    pNtkNew = Abc_NtkFromMiniAig( p );
    Mini_AigStop( p );
    Abc_NtkPrintStats( pNtkNew, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0 );
    Abc_NtkDelete( pNtkNew );

    // test dumping
    p = Abc_NtkToMiniAig( pNtk );
    Mini_AigDump( p, "miniaig.data" );
    Mini_AigPrintStats( p );
    Mini_AigStop( p );

    p = Mini_AigLoad( "miniaig.data" );
    Mini_AigPrintStats( p );
    Mini_AigStop( p );
}

/*
    if ( pNtk )
    {
        extern void Abc_NtkMiniAigTest( Abc_Ntk_t * pNtk );
        Abc_NtkMiniAigTest( pNtk );

    }
*/

////////////////////////////////////////////////////////////////////////
///                       END OF FILE                                ///
////////////////////////////////////////////////////////////////////////


//ABC_NAMESPACE_IMPL_END

