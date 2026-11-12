package abc

import (
	"fmt"
	"os"
	"strings"
)

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
func Abc_NtkToLogic(pNtk *Abc_Ntk_t) *Abc_Ntk_t {
    var pNtkNew *Abc_Ntk_t; 
    var pObj, pFanin *Abc_Obj_t;
    var i, k int32;
    // consider the case of the AIG
    if Abc_NtkIsStrash(pNtk) {
        return Abc_NtkAigToLogicSop( pNtk ); }
    assert(Abc_NtkIsNetlist(pNtk), "Abc_NtkIsNetlist(pNtk)");
    // consider simple case when there is hierarchy
//    assert( pNtk->pDesign == NULL );
    assert(Abc_NtkWhiteboxNum(pNtk) == 0, "Abc_NtkWhiteboxNum(pNtk) == 0");
    assert(Abc_NtkBlackboxNum(pNtk) == 0, "Abc_NtkBlackboxNum(pNtk) == 0");
    // start the network
    pNtkNew = Abc_NtkStartFrom( pNtk, ABC_NTK_LOGIC, pNtk.ntkFunc );
    // duplicate the nodes 
    // inlining of macro Abc_NtkForEachNode
    for  i = 0; (i < Vec_PtrSize((pNtk).vObjs)) && (((pObj) = Abc_NtkObj(pNtk, i)), 1); i++  {           if (pObj) == nil || !Abc_ObjIsNode(pObj) {} else {
        Abc_NtkDupObj(pNtkNew, pObj, 0);
        Abc_ObjAssignName( pObj.pCopy, Abc_ObjName(Abc_ObjFanout0(pObj)), nil );
    }
    }
    // reconnect the internal nodes in the new network
    // inlining of macro Abc_NtkForEachNode
    for  i = 0; (i < Vec_PtrSize((pNtk).vObjs)) && (((pObj) = Abc_NtkObj(pNtk, i)), 1); i++  {           if (pObj) == nil || !Abc_ObjIsNode(pObj) {} else {
        // inlining of macro Abc_ObjForEachFanin
        for  k = 0; (k < Abc_ObjFaninNum(pObj)) && (((pFanin) = Abc_ObjFanin(pObj, k)), 1); k++  {
            Abc_ObjAddFanin( pObj.pCopy, Abc_ObjFanin0(pFanin).pCopy );
        } }
    }
    // collect the CO nodes
    Abc_NtkFinalize( pNtk, pNtkNew );
    // fix the problem with CO pointing directly to CIs
    Abc_NtkLogicMakeSimpleCos( pNtkNew, 0 );
    // duplicate EXDC 
    if pNtk.pExdc {
        pNtkNew.pExdc = Abc_NtkToLogic( pNtk.pExdc ); }
    if !Abc_NtkCheck( pNtkNew ) {
        fmt.Fprintf(os.Stdout, "Abc_NtkToLogic(): Network check has failed.\n"); }
    return pNtkNew;
}

/**Function*************************************************************

  Synopsis    [Transform the logic network into a netlist.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Abc_NtkToNetlist(pNtk *Abc_Ntk_t) *Abc_Ntk_t {
    var pNtkNew, pNtkTemp *Abc_Ntk_t; 
    assert(Abc_NtkIsLogic(pNtk) || Abc_NtkIsStrash(pNtk), "Abc_NtkIsLogic(pNtk) || Abc_NtkIsStrash(pNtk)");
    if Abc_NtkIsStrash(pNtk) {
        pNtkTemp = Abc_NtkAigToLogicSop(pNtk);
        pNtkNew = Abc_NtkLogicToNetlist( pNtkTemp );
        Abc_NtkDelete( pNtkTemp );
        return pNtkNew;
    }
    return Abc_NtkLogicToNetlist( pNtk );
}

/**Function*************************************************************

  Synopsis    [Converts the AIG into the netlist.]

  Description [This procedure does not copy the choices.]
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Abc_NtkToNetlistBench(pNtk *Abc_Ntk_t) *Abc_Ntk_t {
    var pNtkNew, pNtkTemp *Abc_Ntk_t; 
    assert(Abc_NtkIsStrash(pNtk), "Abc_NtkIsStrash(pNtk)");
    pNtkTemp = Abc_NtkAigToLogicSopBench( pNtk );
    pNtkNew = Abc_NtkLogicToNetlist( pNtkTemp );
    Abc_NtkDelete( pNtkTemp );
    return pNtkNew;
}

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
func Abc_NtkLogicToNetlist(pNtk *Abc_Ntk_t) *Abc_Ntk_t {
    var pNtkNew *Abc_Ntk_t; 
    var pObj, pNet, pDriver, pFanin *Abc_Obj_t;
    var i, k int32;

    assert(Abc_NtkIsLogic(pNtk), "Abc_NtkIsLogic(pNtk)");

    // remove dangling nodes
    Abc_NtkCleanup( pNtk, 0 );

    // make sure the CO names are unique
    Abc_NtkCheckUniqueCiNames( pNtk );
    Abc_NtkCheckUniqueCoNames( pNtk );
    Abc_NtkCheckUniqueCioNames( pNtk );

//    assert( Abc_NtkLogicHasSimpleCos(pNtk) );
    if !Abc_NtkLogicHasSimpleCos(pNtk) {
        if !Abc_FrameReadFlag("silentmode") {
            fmt.Printf("Abc_NtkLogicToNetlist() warning: The network is converted to have simple COs.\n"); }
        Abc_NtkLogicMakeSimpleCos( pNtk, 0 );
    }

    // start the netlist by creating PI/PO/Latch objects
    pNtkNew = Abc_NtkStartFrom( pNtk, ABC_NTK_NETLIST, pNtk.ntkFunc );
    // create the CI nets and remember them in the new CI nodes
    // inlining of macro Abc_NtkForEachCi
    for  i = 0; (i < Abc_NtkCiNum(pNtk)) && (((pObj) = Abc_NtkCi(pNtk, i)), 1); i++  {
        pNet = Abc_NtkFindOrCreateNet( pNtkNew, Abc_ObjName(pObj) );
        Abc_ObjAddFanin( pNet, pObj.pCopy );
        pObj.pCopy.pCopy = pNet;
    }
    // duplicate all nodes
    // inlining of macro Abc_NtkForEachNode
    for  i = 0; (i < Vec_PtrSize((pNtk).vObjs)) && (((pObj) = Abc_NtkObj(pNtk, i)), 1); i++  {           if (pObj) == nil || !Abc_ObjIsNode(pObj) {} else {
        Abc_NtkDupObj(pNtkNew, pObj, 0); }
    }
    // first add the nets to the CO drivers
    // inlining of macro Abc_NtkForEachCo
    for  i = 0; (i < Abc_NtkCoNum(pNtk)) && (((pObj) = Abc_NtkCo(pNtk, i)), 1); i++  {
        pDriver = Abc_ObjFanin0(pObj);
        if Abc_ObjIsCi(pDriver) {
            assert(!(int32(strings.Compare(Abc_ObjName(pDriver), Abc_ObjName(pObj))) != 0), "!strcmp( Abc_ObjName(pDriver), Abc_ObjName(pObj) )");
            Abc_ObjAddFanin( pObj.pCopy, pDriver.pCopy.pCopy );
            continue;
        }
        assert(Abc_ObjIsNode(pDriver), "Abc_ObjIsNode(pDriver)");
        // if the CO driver has no net, create it
        if pDriver.pCopy.pCopy == nil {
            // create the CO net and connect it to CO
            //if ( Abc_NtkFindNet(pNtkNew, Abc_ObjName(pDriver)) == NULL )
            //    pNet = Abc_NtkFindOrCreateNet( pNtkNew, Abc_ObjName(pDriver) );
            //else
            pNet = Abc_NtkFindOrCreateNet( pNtkNew, Abc_ObjName(pObj) );
            Abc_ObjAddFanin( pObj.pCopy, pNet );
            // connect the CO net to the new driver and remember it in the new driver
            Abc_ObjAddFanin( pNet, pDriver.pCopy );
            pDriver.pCopy.pCopy = pNet;
        
        } else {
            assert(!(int32(strings.Compare(Abc_ObjName(pDriver.pCopy.pCopy), Abc_ObjName(pObj))) != 0), "!strcmp( Abc_ObjName(pDriver.pCopy.pCopy), Abc_ObjName(pObj) )");
            Abc_ObjAddFanin( pObj.pCopy, pDriver.pCopy.pCopy );
        }
    }
    // create the missing nets
    // inlining of macro Abc_NtkForEachNode
    for  i = 0; (i < Vec_PtrSize((pNtk).vObjs)) && (((pObj) = Abc_NtkObj(pNtk, i)), 1); i++  {           if (pObj) == nil || !Abc_ObjIsNode(pObj) {} else {
        var Buffer string // was [1000]byte in C

        if pObj.pCopy.pCopy { // the net of the new object is already created
            continue; }
        // create the new net
        Buffer = fmt.Sprintf("new_%s_", Abc_ObjName(pObj));
        //pNet = Abc_NtkFindOrCreateNet( pNtkNew, Abc_ObjName(pObj) ); // here we create net names such as "n48", where 48 is the ID of the node
        pNet = Abc_NtkFindOrCreateNet( pNtkNew, Buffer ); 
        Abc_ObjAddFanin( pNet, pObj.pCopy );
        pObj.pCopy.pCopy = pNet;
    }
    }
    // connect nodes to the fanins nets
    // inlining of macro Abc_NtkForEachNode
    for  i = 0; (i < Vec_PtrSize((pNtk).vObjs)) && (((pObj) = Abc_NtkObj(pNtk, i)), 1); i++  {           if (pObj) == nil || !Abc_ObjIsNode(pObj) {} else {
        // inlining of macro Abc_ObjForEachFanin
        for  k = 0; (k < Abc_ObjFaninNum(pObj)) && (((pFanin) = Abc_ObjFanin(pObj, k)), 1); k++  {
            Abc_ObjAddFanin( pObj.pCopy, pFanin.pCopy.pCopy );
        } }
    }
    // duplicate EXDC 
    if pNtk.pExdc {
        pNtkNew.pExdc = Abc_NtkToNetlist( pNtk.pExdc ); }
    if !Abc_NtkCheck( pNtkNew ) {
        fmt.Fprintf(os.Stdout, "Abc_NtkLogicToNetlist(): Network check has failed.\n"); }
    return pNtkNew;
}

/**Function*************************************************************

  Synopsis    [Converts the AIG into the logic network with SOPs.]

  Description [Correctly handles the case of choice nodes.]
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Abc_NtkAigToLogicSop(pNtk *Abc_Ntk_t) *Abc_Ntk_t {
    extern int Abc_NtkLogicMakeSimpleCos2( Abc_Ntk_t * pNtk, int fDuplicate );

    var pNtkNew *Abc_Ntk_t; 
    var pObj, pFanin, pNodeNew *Abc_Obj_t;
    var vInts *Vec_Int_t;
    var i, k, fChoices int32;
    assert(Abc_NtkIsStrash(pNtk), "Abc_NtkIsStrash(pNtk)");
    // start the network
    pNtkNew = Abc_NtkStartFrom( pNtk, ABC_NTK_LOGIC, ABC_FUNC_SOP );
    // if the constant node is used, duplicate it
    pObj = Abc_AigConst1(pNtk);
    if Abc_ObjFanoutNum(pObj) > 0 {
        pObj.pCopy = Abc_NtkCreateNodeConst1(pNtkNew); }
    // duplicate the nodes and create node functions
    // inlining of macro Abc_NtkForEachNode
    for  i = 0; (i < Vec_PtrSize((pNtk).vObjs)) && (((pObj) = Abc_NtkObj(pNtk, i)), 1); i++  {           if (pObj) == nil || !Abc_ObjIsNode(pObj) {} else {
        Abc_NtkDupObj(pNtkNew, pObj, 0);
        pObj.pCopy.pData = Abc_SopCreateAnd2( (Mem_Flex_t *)pNtkNew.pManFunc, Abc_ObjFaninC0(pObj), Abc_ObjFaninC1(pObj) );
    }
    }
    // create the choice nodes
    // inlining of macro Abc_NtkForEachNode
    for  i = 0; (i < Vec_PtrSize((pNtk).vObjs)) && (((pObj) = Abc_NtkObj(pNtk, i)), 1); i++  {           if (pObj) == nil || !Abc_ObjIsNode(pObj) {} else {
        if !Abc_AigNodeIsChoice(pObj) {
            continue; }
        // create an OR gate
        pNodeNew = Abc_NtkCreateNode(pNtkNew);
        // add fanins
        vInts = Vec_IntAlloc( 10 );
        for  pFanin = pObj; (pFanin != nil); pFanin = (Abc_Obj_t *)pFanin.pData  {
            Vec_IntPush( vInts, int32((pObj.fPhase != pFanin.fPhase)) );
            Abc_ObjAddFanin( pNodeNew, pFanin.pCopy );
        }
        // create the logic function
        pNodeNew.pData = Abc_SopCreateOrMultiCube( (Mem_Flex_t *)pNtkNew.pManFunc, Vec_IntSize(vInts), Vec_IntArray(vInts) );
        // set the new node
        pObj.pCopy.pCopy = pNodeNew;
        Vec_IntFree( vInts );
        fChoices = 1;
    }
    }
    // connect the internal nodes
    // inlining of macro Abc_NtkForEachNode
    for  i = 0; (i < Vec_PtrSize((pNtk).vObjs)) && (((pObj) = Abc_NtkObj(pNtk, i)), 1); i++  {           if (pObj) == nil || !Abc_ObjIsNode(pObj) {} else {
        // inlining of macro Abc_ObjForEachFanin
        for  k = 0; (k < Abc_ObjFaninNum(pObj)) && (((pFanin) = Abc_ObjFanin(pObj, k)), 1); k++  {
            if pFanin.pCopy.pCopy {
                Abc_ObjAddFanin( pObj.pCopy, pFanin.pCopy.pCopy );
            } else {
                Abc_ObjAddFanin( pObj.pCopy, pFanin.pCopy ); }
        } }
    }
    // connect the COs
//    Abc_NtkFinalize( pNtk, pNtkNew );
    // inlining of macro Abc_NtkForEachCo
    for  i = 0; (i < Abc_NtkCoNum(pNtk)) && (((pObj) = Abc_NtkCo(pNtk, i)), 1); i++  {
        pFanin = Abc_ObjFanin0(pObj);
        if pFanin.pCopy.pCopy {
            pNodeNew = Abc_ObjNotCond(pFanin.pCopy.pCopy, Abc_ObjFaninC0(pObj));
        } else {
            pNodeNew = Abc_ObjNotCond(pFanin.pCopy, Abc_ObjFaninC0(pObj)); }
        Abc_ObjAddFanin( pObj.pCopy, pNodeNew );
    }

    // fix the problem with complemented and duplicated CO edges
    if fChoices != 0 {
        Abc_NtkLogicMakeSimpleCos2( pNtkNew, 0 );
    } else {
        Abc_NtkLogicMakeSimpleCos( pNtkNew, 0 ); }
    // duplicate the EXDC Ntk
    if pNtk.pExdc {
        if Abc_NtkIsStrash(pNtk.pExdc) {
            pNtkNew.pExdc = Abc_NtkAigToLogicSop( pNtk.pExdc );
        } else {
            pNtkNew.pExdc = Abc_NtkDup( pNtk.pExdc ); }
    }
    if !Abc_NtkCheck( pNtkNew ) {
        fmt.Fprintf(os.Stdout, "Abc_NtkAigToLogicSop(): Network check has failed.\n"); }
    return pNtkNew;
}

/**Function*************************************************************

  Synopsis    [Converts the AIG into the logic network with SOPs for bench writing.]

  Description [This procedure does not copy the choices.]
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Abc_NtkAigToLogicSopBench(pNtk *Abc_Ntk_t) *Abc_Ntk_t {
    var pNtkNew *Abc_Ntk_t; 
    var pObj, pFanin *Abc_Obj_t;
    var vNodes *Vec_Ptr_t;
    var i, k int32;
    assert(Abc_NtkIsStrash(pNtk), "Abc_NtkIsStrash(pNtk)");
    if Abc_NtkGetChoiceNum(pNtk) {
        fmt.Printf("Warning: Choice nodes are skipped.\n"); }
    // start the network
    pNtkNew = Abc_NtkStartFrom( pNtk, ABC_NTK_LOGIC, ABC_FUNC_SOP );
    // collect the nodes to be used (marks all nodes with current TravId)
    vNodes = Abc_NtkDfs( pNtk, 0 );
    // create inverters for the constant node
    pObj = Abc_AigConst1(pNtk);
    if Abc_ObjFanoutNum(pObj) > 0 {
        pObj.pCopy = Abc_NtkCreateNodeConst1(pNtkNew); }
    if Abc_AigNodeHasComplFanoutEdgeTrav(pObj) {
        pObj.pCopy.pCopy = Abc_NtkCreateNodeInv( pNtkNew, pObj.pCopy ); }
    // create inverters for the CIs
    // inlining of macro Abc_NtkForEachCi
    for  i = 0; (i < Abc_NtkCiNum(pNtk)) && (((pObj) = Abc_NtkCi(pNtk, i)), 1); i++  {
        if Abc_AigNodeHasComplFanoutEdgeTrav(pObj) {
            pObj.pCopy.pCopy = Abc_NtkCreateNodeInv( pNtkNew, pObj.pCopy ); }
    }
    // duplicate the nodes, create node functions, and inverters
    // inlining of macro Vec_PtrForEachEntry
    for  i = 0; (i < Vec_PtrSize(vNodes)) && (((pObj) = (Abc_Obj_t *)Vec_PtrEntry(vNodes, i)), 1); i++  {
        Abc_NtkDupObj( pNtkNew, pObj, 0 );
        pObj.pCopy.pData = Abc_SopCreateAnd( (Mem_Flex_t *)pNtkNew.pManFunc, 2, nil );
        if Abc_AigNodeHasComplFanoutEdgeTrav(pObj) {
            pObj.pCopy.pCopy = Abc_NtkCreateNodeInv( pNtkNew, pObj.pCopy ); }
    }
    // connect the objects
    // inlining of macro Vec_PtrForEachEntry
    for  i = 0; (i < Vec_PtrSize(vNodes)) && (((pObj) = (Abc_Obj_t *)Vec_PtrEntry(vNodes, i)), 1); i++  {
        // inlining of macro Abc_ObjForEachFanin
        for  k = 0; (k < Abc_ObjFaninNum(pObj)) && (((pFanin) = Abc_ObjFanin(pObj, k)), 1); k++  {
            if Abc_ObjFaninC( pObj, k ) {
                Abc_ObjAddFanin( pObj.pCopy, pFanin.pCopy.pCopy );
            } else {
                Abc_ObjAddFanin( pObj.pCopy, pFanin.pCopy ); }
        }
    }
    Vec_PtrFree( vNodes );
    // connect the COs
    // inlining of macro Abc_NtkForEachCo
    for  i = 0; (i < Abc_NtkCoNum(pNtk)) && (((pObj) = Abc_NtkCo(pNtk, i)), 1); i++  {
        pFanin = Abc_ObjFanin0(pObj);
        if Abc_ObjFaninC0( pObj ) {
            Abc_ObjAddFanin( pObj.pCopy, pFanin.pCopy.pCopy );
        } else {
            Abc_ObjAddFanin( pObj.pCopy, pFanin.pCopy ); }
    }
    // fix the problem with complemented and duplicated CO edges
    Abc_NtkLogicMakeSimpleCos( pNtkNew, 0 );
    // duplicate the EXDC Ntk
    if pNtk.pExdc {
        fmt.Printf("Warning: The EXDc network is skipped.\n"); }
    if !Abc_NtkCheck( pNtkNew ) {
        fmt.Fprintf(os.Stdout, "Abc_NtkAigToLogicSopBench(): Network check has failed.\n"); }
    return pNtkNew;
}

/**Function*************************************************************

  Synopsis    [Converts the AIG into the logic network with SOPs for bench writing.]

  Description [This procedure does not copy the choices.]
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Abc_NtkAigToLogicSopNand(pNtk *Abc_Ntk_t) *Abc_Ntk_t {
    var pNtkNew *Abc_Ntk_t; 
    var pObj, pFanin *Abc_Obj_t;
    var vNodes *Vec_Ptr_t;
    var i, k int32;
    assert(Abc_NtkIsStrash(pNtk), "Abc_NtkIsStrash(pNtk)");
    if Abc_NtkGetChoiceNum(pNtk) {
        fmt.Printf("Warning: Choice nodes are skipped.\n"); }
    // convert complemented edges
    // inlining of macro Abc_NtkForEachObj
    for  i = 0; (i < Vec_PtrSize((pNtk).vObjs)) && (((pObj) = Abc_NtkObj(pNtk, i)), 1); i++  {            if (pObj) == nil {} else {
        // inlining of macro Abc_ObjForEachFanin
        for  k = 0; (k < Abc_ObjFaninNum(pObj)) && (((pFanin) = Abc_ObjFanin(pObj, k)), 1); k++  {
            if Abc_ObjIsNode(pFanin) {
                Abc_ObjXorFaninC( pObj, k ); }
        } }
    }
    // start the network
    pNtkNew = Abc_NtkStartFrom( pNtk, ABC_NTK_LOGIC, ABC_FUNC_SOP );
    // collect the nodes to be used (marks all nodes with current TravId)
    vNodes = Abc_NtkDfs( pNtk, 0 );
    // create inverters for the constant node
    pObj = Abc_AigConst1(pNtk);
    if Abc_ObjFanoutNum(pObj) > 0 {
        pObj.pCopy = Abc_NtkCreateNodeConst1(pNtkNew); }
    if Abc_AigNodeHasComplFanoutEdgeTrav(pObj) {
        pObj.pCopy.pCopy = Abc_NtkCreateNodeInv( pNtkNew, pObj.pCopy ); }
    // create inverters for the CIs
    // inlining of macro Abc_NtkForEachCi
    for  i = 0; (i < Abc_NtkCiNum(pNtk)) && (((pObj) = Abc_NtkCi(pNtk, i)), 1); i++  {
        if Abc_AigNodeHasComplFanoutEdgeTrav(pObj) {
            pObj.pCopy.pCopy = Abc_NtkCreateNodeInv( pNtkNew, pObj.pCopy ); }
    }
    // duplicate the nodes, create node functions, and inverters
    // inlining of macro Vec_PtrForEachEntry
    for  i = 0; (i < Vec_PtrSize(vNodes)) && (((pObj) = (Abc_Obj_t *)Vec_PtrEntry(vNodes, i)), 1); i++  {
        Abc_NtkDupObj( pNtkNew, pObj, 0 );
        pObj.pCopy.pData = Abc_SopCreateNand( (Mem_Flex_t *)pNtkNew.pManFunc, 2 );
        if Abc_AigNodeHasComplFanoutEdgeTrav(pObj) {
            pObj.pCopy.pCopy = Abc_NtkCreateNodeInv( pNtkNew, pObj.pCopy ); }
    }
    // connect the objects
    // inlining of macro Vec_PtrForEachEntry
    for  i = 0; (i < Vec_PtrSize(vNodes)) && (((pObj) = (Abc_Obj_t *)Vec_PtrEntry(vNodes, i)), 1); i++  {
        // inlining of macro Abc_ObjForEachFanin
        for  k = 0; (k < Abc_ObjFaninNum(pObj)) && (((pFanin) = Abc_ObjFanin(pObj, k)), 1); k++  {
            if Abc_ObjFaninC( pObj, k ) {
                Abc_ObjAddFanin( pObj.pCopy, pFanin.pCopy.pCopy );
            } else {
                Abc_ObjAddFanin( pObj.pCopy, pFanin.pCopy ); }
        }
    }
    Vec_PtrFree( vNodes );
    // connect the COs
    // inlining of macro Abc_NtkForEachCo
    for  i = 0; (i < Abc_NtkCoNum(pNtk)) && (((pObj) = Abc_NtkCo(pNtk, i)), 1); i++  {
        pFanin = Abc_ObjFanin0(pObj);
        if Abc_ObjFaninC0( pObj ) {
            Abc_ObjAddFanin( pObj.pCopy, pFanin.pCopy.pCopy );
        } else {
            Abc_ObjAddFanin( pObj.pCopy, pFanin.pCopy ); }
    }
    // fix the problem with complemented and duplicated CO edges
    Abc_NtkLogicMakeSimpleCos( pNtkNew, 0 );
    // convert complemented edges
    // inlining of macro Abc_NtkForEachObj
    for  i = 0; (i < Vec_PtrSize((pNtk).vObjs)) && (((pObj) = Abc_NtkObj(pNtk, i)), 1); i++  {            if (pObj) == nil {} else {
        // inlining of macro Abc_ObjForEachFanin
        for  k = 0; (k < Abc_ObjFaninNum(pObj)) && (((pFanin) = Abc_ObjFanin(pObj, k)), 1); k++  {
            if Abc_ObjIsNode(pFanin) {
                Abc_ObjXorFaninC( pObj, k ); }
        } }
    }
    // duplicate the EXDC Ntk
    if pNtk.pExdc {
        fmt.Printf("Warning: The EXDc network is skipped.\n"); }
    if !Abc_NtkCheck( pNtkNew ) {
        fmt.Fprintf(os.Stdout, "Abc_NtkAigToLogicSopBench(): Network check has failed.\n"); }
    return pNtkNew;
}

/**Function*************************************************************

  Synopsis    [Adds buffers for each PO.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Abc_NtkAddPoBuffers(pNtk *Abc_Ntk_t) {
    var pObj, pFanin, pFaninNew *Abc_Obj_t;
    var i int32;
    assert(Abc_NtkIsStrash(pNtk), "Abc_NtkIsStrash(pNtk)");
    // inlining of macro Abc_NtkForEachPo
    for  i = 0; (i < Abc_NtkPoNum(pNtk)) && (((pObj) = Abc_NtkPo(pNtk, i)), 1); i++  {
        pFanin = Abc_ObjChild0(pObj);
        pFaninNew = Abc_NtkCreateNode(pNtk);
        Abc_ObjAddFanin( pFaninNew, pFanin );
        Abc_ObjPatchFanin( pObj, pFanin, pFaninNew );
    }
}

////////////////////////////////////////////////////////////////////////
///                       END OF FILE                                ///
////////////////////////////////////////////////////////////////////////


//ABC_NAMESPACE_IMPL_END

