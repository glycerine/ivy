package abc

import (
	"fmt"
	"os"
	"strings"
)

/**CFile****************************************************************

  FileName    [abcHie.c]

  SystemName  [ABC: Logic synthesis and verification system.]

  PackageName [Network and node package.]

  Synopsis    [Procedures to handle hierarchy.]

  Author      [Alan Mishchenko]
  
  Affiliation [UC Berkeley]

  Date        [Ver. 1.0. Started - June 20, 2005.]

  Revision    [$Id: abcHie.c,v 1.00 2005/06/20 00:00:00 alanmi Exp $]

***********************************************************************/

//#include "abc.h"

//ABC_NAMESPACE_IMPL_START


////////////////////////////////////////////////////////////////////////
///                        DECLARATIONS                              ///
////////////////////////////////////////////////////////////////////////

////////////////////////////////////////////////////////////////////////
///                     FUNCTION DEFINITIONS                         ///
////////////////////////////////////////////////////////////////////////

/**Function*************************************************************

  Synopsis    [Recursively flattens logic hierarchy of the netlist.]

  Description [When this procedure is called, the PI/PO nets of the old 
  netlist point to the corresponding nets of the flattened netlist.]
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Abc_NtkFlattenLogicHierarchy2_rec(pNtkNew *Abc_Ntk_t, pNtk *Abc_Ntk_t, pCounter *int32) {
    var Suffix string // was [2000]byte in C

    var pNtkModel *Abc_Ntk_t;
    var pObj, pTerm, pNet, pFanin *Abc_Obj_t;
    var i, k int32;

    // process the blackbox
    if Abc_NtkHasBlackbox(pNtk) {
        // duplicate the blackbox
        assert(Abc_NtkBoxNum(pNtk) == 1, "Abc_NtkBoxNum(pNtk) == 1");
        pObj = Abc_NtkBox( pNtk, 0 );
        Abc_NtkDupBox( pNtkNew, pObj, 1 );
        pObj.pCopy.pData = pNtk;

        // connect blackbox fanins to the PI nets
        assert(Abc_ObjFaninNum(pObj.pCopy) == Abc_NtkPiNum(pNtk), "Abc_ObjFaninNum(pObj.pCopy) == Abc_NtkPiNum(pNtk)");
        // inlining of macro Abc_NtkForEachPi
        for  i = 0; (i < Abc_NtkPiNum(pNtk)) && (((pTerm) = Abc_NtkPi(pNtk, i)), 1); i++  {
            Abc_ObjAddFanin( Abc_ObjFanin(pObj.pCopy,i), Abc_ObjFanout0(pTerm).pCopy );
        }

        // connect blackbox fanouts to the PO nets
        assert(Abc_ObjFanoutNum(pObj.pCopy) == Abc_NtkPoNum(pNtk), "Abc_ObjFanoutNum(pObj.pCopy) == Abc_NtkPoNum(pNtk)");
        // inlining of macro Abc_NtkForEachPo
        for  i = 0; (i < Abc_NtkPoNum(pNtk)) && (((pTerm) = Abc_NtkPo(pNtk, i)), 1); i++  {
            Abc_ObjAddFanin( Abc_ObjFanin0(pTerm).pCopy, Abc_ObjFanout(pObj.pCopy,i) );
        }
        return;
    }

    (*pCounter)++;

    // create the suffix, which will be appended to the internal names
    if *pCounter {
        Suffix = fmt.Sprintf("_%s_%d", Abc_NtkName(pNtk), *pCounter); }

    // duplicate nets of all boxes, including latches
    // inlining of macro Abc_NtkForEachBox
    for  i = 0; (i < Vec_PtrSize((pNtk).vBoxes)) && (((pObj) = Abc_NtkBox(pNtk, i)), 1); i++  {
        // inlining of macro Abc_ObjForEachFanin
        for  k = 0; (k < Abc_ObjFaninNum(pObj)) && (((pTerm) = Abc_ObjFanin(pObj, k)), 1); k++  {
            pNet = Abc_ObjFanin0(pTerm);
            if pNet.pCopy {
                continue; }
            pNet.pCopy = Abc_NtkFindOrCreateNet( pNtkNew, Abc_ObjNameSuffix(pNet, Suffix) );
        }
        // inlining of macro Abc_ObjForEachFanout
        for  k = 0; (k < Abc_ObjFanoutNum(pObj)) && (((pTerm) = Abc_ObjFanout(pObj, k)), 1); k++  {
            pNet = Abc_ObjFanout0(pTerm);
            if pNet.pCopy {
                continue; }
            pNet.pCopy = Abc_NtkFindOrCreateNet( pNtkNew, Abc_ObjNameSuffix(pNet, Suffix) );
        }
    }

    // mark objects that will not be used
    Abc_NtkIncrementTravId( pNtk );
    // inlining of macro Abc_NtkForEachPi
    for  i = 0; (i < Abc_NtkPiNum(pNtk)) && (((pTerm) = Abc_NtkPi(pNtk, i)), 1); i++  {
        Abc_NodeSetTravIdCurrent( pTerm );
    }
    // inlining of macro Abc_NtkForEachPo
    for  i = 0; (i < Abc_NtkPoNum(pNtk)) && (((pTerm) = Abc_NtkPo(pNtk, i)), 1); i++  {
        Abc_NodeSetTravIdCurrent( pTerm );
        // if the netlist has net names beginning with "abc_property_"
        // these names will be addes as primary outputs of the network
        pNet = Abc_ObjFanin0(pTerm);
        if strncmp( Abc_ObjName(pNet), "abc_property", 12 ) {
            continue; }
        Abc_ObjAddFanin( Abc_NtkCreatePo(pNet.pCopy.pNtk), pNet.pCopy );
        if Nm_ManFindNameById(pNet.pCopy.pNtk.pManName, pNet.pCopy.Id) {
            Nm_ManDeleteIdName(pNet.pCopy.pNtk.pManName, pNet.pCopy.Id); }
        Abc_ObjAssignName( pNet.pCopy, Abc_ObjName(pNet), Suffix );
    }
    // inlining of macro Abc_NtkForEachBox
    for  i = 0; (i < Vec_PtrSize((pNtk).vBoxes)) && (((pObj) = Abc_NtkBox(pNtk, i)), 1); i++  {
        if Abc_ObjIsLatch(pObj) {
            continue; }
        Abc_NodeSetTravIdCurrent( pObj );
        // inlining of macro Abc_ObjForEachFanin
        for  k = 0; (k < Abc_ObjFaninNum(pObj)) && (((pTerm) = Abc_ObjFanin(pObj, k)), 1); k++  {
            Abc_NodeSetTravIdCurrent( pTerm );
        }
        // inlining of macro Abc_ObjForEachFanout
        for  k = 0; (k < Abc_ObjFanoutNum(pObj)) && (((pTerm) = Abc_ObjFanout(pObj, k)), 1); k++  {
            Abc_NodeSetTravIdCurrent( pTerm );
        }
    }

    // duplicate objects that do not have prototypes yet
    // inlining of macro Abc_NtkForEachObj
    for  i = 0; (i < Vec_PtrSize((pNtk).vObjs)) && (((pObj) = Abc_NtkObj(pNtk, i)), 1); i++  {            if (pObj) == nil {} else {
        if Abc_NodeIsTravIdCurrent(pObj) {
            continue; }
        if pObj.pCopy {
            continue; }
        Abc_NtkDupObj( pNtkNew, pObj, 0 );
    }
    }

    // connect objects
    // inlining of macro Abc_NtkForEachObj
    for  i = 0; (i < Vec_PtrSize((pNtk).vObjs)) && (((pObj) = Abc_NtkObj(pNtk, i)), 1); i++  {            if (pObj) == nil {} else
        if !Abc_NodeIsTravIdCurrent(pObj) {
            // inlining of macro Abc_ObjForEachFanin
            for  k = 0; (k < Abc_ObjFaninNum(pObj)) && (((pFanin) = Abc_ObjFanin(pObj, k)), 1); k++  {
                if !Abc_NodeIsTravIdCurrent(pFanin) {
                    Abc_ObjAddFanin( pObj.pCopy, pFanin.pCopy ); }
            } }
    }

    // call recursively
    // inlining of macro Abc_NtkForEachBox
    for  i = 0; (i < Vec_PtrSize((pNtk).vBoxes)) && (((pObj) = Abc_NtkBox(pNtk, i)), 1); i++  {
        if Abc_ObjIsLatch(pObj) {
            continue; }
        pNtkModel = (Abc_Ntk_t *)pObj.pData;
        // check the match between the number of actual and formal parameters
        assert(Abc_ObjFaninNum(pObj) == Abc_NtkPiNum(pNtkModel), "Abc_ObjFaninNum(pObj) == Abc_NtkPiNum(pNtkModel)");
        assert(Abc_ObjFanoutNum(pObj) == Abc_NtkPoNum(pNtkModel), "Abc_ObjFanoutNum(pObj) == Abc_NtkPoNum(pNtkModel)");
        // clean the node copy fields
        Abc_NtkCleanCopy( pNtkModel );
        // map PIs/POs
        // inlining of macro Abc_ObjForEachFanin
        for  k = 0; (k < Abc_ObjFaninNum(pObj)) && (((pTerm) = Abc_ObjFanin(pObj, k)), 1); k++  {
            Abc_ObjFanout0( Abc_NtkPi(pNtkModel, k) ).pCopy = Abc_ObjFanin0(pTerm).pCopy;
        }
        // inlining of macro Abc_ObjForEachFanout
        for  k = 0; (k < Abc_ObjFanoutNum(pObj)) && (((pTerm) = Abc_ObjFanout(pObj, k)), 1); k++  {
            Abc_ObjFanin0( Abc_NtkPo(pNtkModel, k) ).pCopy = Abc_ObjFanout0(pTerm).pCopy;
        }
        // call recursively
        Abc_NtkFlattenLogicHierarchy2_rec( pNtkNew, pNtkModel, pCounter );
    }

    // if it is a BLIF-MV netlist transfer the values of all nets
    if Abc_NtkHasBlifMv(pNtk) && Abc_NtkMvVar(pNtk) {
        if Abc_NtkMvVar( pNtkNew ) == nil {
            Abc_NtkStartMvVars( pNtkNew ); }
        // inlining of macro Abc_NtkForEachNet
        for  i = 0; (i < Vec_PtrSize((pNtk).vObjs)) && (((pObj) = Abc_NtkObj(pNtk, i)), 1); i++  {            if (pObj) == nil || !Abc_ObjIsNet(pObj) {} else {
            Abc_NtkSetMvVarValues( pObj.pCopy, Abc_ObjMvVarNum(pObj) ); }
        }
    }
}

/**Function*************************************************************

  Synopsis    [Flattens the logic hierarchy of the netlist.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Abc_NtkFlattenLogicHierarchy2(pNtk *Abc_Ntk_t) *Abc_Ntk_t {
    var pNtkNew *Abc_Ntk_t; 
    var pTerm, pNet *Abc_Obj_t;
    var i, Counter int32;
    extern Abc_Des_t * Abc_DesDupBlackboxes( Abc_Des_t * p, Abc_Ntk_t * pNtkSave );

    assert(Abc_NtkIsNetlist(pNtk), "Abc_NtkIsNetlist(pNtk)");
    // start the network
    pNtkNew = Abc_NtkAlloc( pNtk.ntkType, pNtk.ntkFunc, 1 );
    // duplicate the name and the spec
    pNtkNew.pName = Abc_UtilStrsav(pNtk.pName);
    pNtkNew.pSpec = Abc_UtilStrsav(pNtk.pSpec);

    // clean the node copy fields
    Abc_NtkCleanCopy( pNtk );

    // duplicate PIs/POs and their nets
    // inlining of macro Abc_NtkForEachPi
    for  i = 0; (i < Abc_NtkPiNum(pNtk)) && (((pTerm) = Abc_NtkPi(pNtk, i)), 1); i++  {
        Abc_NtkDupObj( pNtkNew, pTerm, 0 );
        pNet = Abc_ObjFanout0( pTerm );
        pNet.pCopy = Abc_NtkFindOrCreateNet( pNtkNew, Abc_ObjName(pNet) );
        Abc_ObjAddFanin( pNet.pCopy, pTerm.pCopy );
    }
    // inlining of macro Abc_NtkForEachPo
    for  i = 0; (i < Abc_NtkPoNum(pNtk)) && (((pTerm) = Abc_NtkPo(pNtk, i)), 1); i++  {
        Abc_NtkDupObj( pNtkNew, pTerm, 0 );
        pNet = Abc_ObjFanin0( pTerm );
        pNet.pCopy = Abc_NtkFindOrCreateNet( pNtkNew, Abc_ObjName(pNet) );
        Abc_ObjAddFanin( pTerm.pCopy, pNet.pCopy );
    }

    // recursively flatten hierarchy, create internal logic, add new PI/PO names if there are black boxes
    Counter = -1;
    Abc_NtkFlattenLogicHierarchy2_rec( pNtkNew, pNtk, &Counter );
    fmt.Printf("Hierarchy reader flattened %d instances of logic boxes and left %d black boxes.\n", Counter, Abc_NtkBlackboxNum(pNtkNew));

    if pNtk.pDesign {
        // pass on the design
        assert(Vec_PtrEntry(pNtk.pDesign.vTops, 0) == pNtk, "Vec_PtrEntry(pNtk.pDesign.vTops, 0) == pNtk");
        pNtkNew.pDesign = Abc_DesDupBlackboxes( pNtk.pDesign, pNtkNew );
        // update the pointers
        // inlining of macro Abc_NtkForEachBlackbox
        for  i = 0; (i < Vec_PtrSize((pNtkNew).vBoxes)) && (((pTerm) = Abc_NtkBox(pNtkNew, i)), 1); i++  {           if !Abc_ObjIsBlackbox(pTerm) {} else {
            pTerm.pData = ((Abc_Ntk_t *)pTerm.pData).pCopy; }
        }
    }

    // we may have added property outputs
    Abc_NtkOrderCisCos( pNtkNew );

    // copy the timing information
//    Abc_ManTimeDup( pNtk, pNtkNew );
    // duplicate EXDC 
    if pNtk.pExdc {
        fmt.Printf("EXDC is not transformed.\n"); }
    if !Abc_NtkCheck( pNtkNew ) {
        fmt.Fprintf(os.Stdout, "Abc_NtkFlattenLogicHierarchy2(): Network check has failed.\n");
        Abc_NtkDelete( pNtkNew );
        return nil;
    }
    return pNtkNew;
}



/**Function*************************************************************

  Synopsis    [Recursively flattens logic hierarchy of the netlist.]

  Description [When this procedure is called, the PI/PO nets of the old 
  netlist point to the corresponding nets of the flattened netlist.]
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Abc_NtkFlattenLogicHierarchy_rec(pNtkNew *Abc_Ntk_t, pNtk *Abc_Ntk_t, pCounter *int32, vPref *Vec_Str_t) {
    var pNtkModel *Abc_Ntk_t;
    var pObj, pTerm, pNet, pFanin *Abc_Obj_t;
    var i, k, Length int32;

    // process the blackbox
    if Abc_NtkHasBlackbox(pNtk) {
        //printf( "Flatting black box \"%s\".\n", pNtk->pName );
        // duplicate the blackbox
        assert(Abc_NtkBoxNum(pNtk) == 1, "Abc_NtkBoxNum(pNtk) == 1");
        pObj = Abc_NtkBox( pNtk, 0 );
        Abc_NtkDupBox( pNtkNew, pObj, 1 );
        pObj.pCopy.pData = pNtk;

        // connect blackbox fanins to the PI nets
        assert(Abc_ObjFaninNum(pObj.pCopy) == Abc_NtkPiNum(pNtk), "Abc_ObjFaninNum(pObj.pCopy) == Abc_NtkPiNum(pNtk)");
        // inlining of macro Abc_NtkForEachPi
        for  i = 0; (i < Abc_NtkPiNum(pNtk)) && (((pTerm) = Abc_NtkPi(pNtk, i)), 1); i++  {
            Abc_ObjAddFanin( Abc_ObjFanin(pObj.pCopy,i), Abc_ObjFanout0(pTerm).pCopy );
        }

        // connect blackbox fanouts to the PO nets
        assert(Abc_ObjFanoutNum(pObj.pCopy) == Abc_NtkPoNum(pNtk), "Abc_ObjFanoutNum(pObj.pCopy) == Abc_NtkPoNum(pNtk)");
        // inlining of macro Abc_NtkForEachPo
        for  i = 0; (i < Abc_NtkPoNum(pNtk)) && (((pTerm) = Abc_NtkPo(pNtk, i)), 1); i++  {
            Abc_ObjAddFanin( Abc_ObjFanin0(pTerm).pCopy, Abc_ObjFanout(pObj.pCopy,i) );
        }
        return;
    }

    (*pCounter)++;

    // create the suffix, which will be appended to the internal names
    if *pCounter {
        var Buffer string // was [20]byte in C

        Buffer = fmt.Sprintf("(%d)", *pCounter);
        Vec_StrPrintStr( vPref, Buffer );
    }
    Vec_StrPush( vPref, '|' );
    Vec_StrPush( vPref, 0 );

    // duplicate nets of all boxes, including latches
    // inlining of macro Abc_NtkForEachBox
    for  i = 0; (i < Vec_PtrSize((pNtk).vBoxes)) && (((pObj) = Abc_NtkBox(pNtk, i)), 1); i++  {
        // inlining of macro Abc_ObjForEachFanin
        for  k = 0; (k < Abc_ObjFaninNum(pObj)) && (((pTerm) = Abc_ObjFanin(pObj, k)), 1); k++  {
            pNet = Abc_ObjFanin0(pTerm);
            if pNet.pCopy {
                continue; }
            pNet.pCopy = Abc_NtkFindOrCreateNet( pNtkNew, Abc_ObjNamePrefix(pNet, Vec_StrArray(vPref)) );
        }
        // inlining of macro Abc_ObjForEachFanout
        for  k = 0; (k < Abc_ObjFanoutNum(pObj)) && (((pTerm) = Abc_ObjFanout(pObj, k)), 1); k++  {
            pNet = Abc_ObjFanout0(pTerm);
            if pNet.pCopy {
                continue; }
            pNet.pCopy = Abc_NtkFindOrCreateNet( pNtkNew, Abc_ObjNamePrefix(pNet, Vec_StrArray(vPref)) );
        }
    }

    // mark objects that will not be used
    Abc_NtkIncrementTravId( pNtk );
    // inlining of macro Abc_NtkForEachPi
    for  i = 0; (i < Abc_NtkPiNum(pNtk)) && (((pTerm) = Abc_NtkPi(pNtk, i)), 1); i++  {
        Abc_NodeSetTravIdCurrent( pTerm );
    }
    // inlining of macro Abc_NtkForEachPo
    for  i = 0; (i < Abc_NtkPoNum(pNtk)) && (((pTerm) = Abc_NtkPo(pNtk, i)), 1); i++  {
        Abc_NodeSetTravIdCurrent( pTerm );
        // if the netlist has net names beginning with "abc_property_"
        // these names will be addes as primary outputs of the network
        pNet = Abc_ObjFanin0(pTerm);
        if strncmp( Abc_ObjName(pNet), "abc_property", 12 ) {
            continue; }
        Abc_ObjAddFanin( Abc_NtkCreatePo(pNet.pCopy.pNtk), pNet.pCopy );
        if Nm_ManFindNameById(pNet.pCopy.pNtk.pManName, pNet.pCopy.Id) {
            Nm_ManDeleteIdName(pNet.pCopy.pNtk.pManName, pNet.pCopy.Id); }
        Abc_ObjAssignName( pNet.pCopy, Vec_StrArray(vPref), Abc_ObjName(pNet) );
    }
    // inlining of macro Abc_NtkForEachBox
    for  i = 0; (i < Vec_PtrSize((pNtk).vBoxes)) && (((pObj) = Abc_NtkBox(pNtk, i)), 1); i++  {
        if Abc_ObjIsLatch(pObj) {
            continue; }
        Abc_NodeSetTravIdCurrent( pObj );
        // inlining of macro Abc_ObjForEachFanin
        for  k = 0; (k < Abc_ObjFaninNum(pObj)) && (((pTerm) = Abc_ObjFanin(pObj, k)), 1); k++  {
            Abc_NodeSetTravIdCurrent( pTerm );
        }
        // inlining of macro Abc_ObjForEachFanout
        for  k = 0; (k < Abc_ObjFanoutNum(pObj)) && (((pTerm) = Abc_ObjFanout(pObj, k)), 1); k++  {
            Abc_NodeSetTravIdCurrent( pTerm );
        }
    }

    // duplicate objects that do not have prototypes yet
    // inlining of macro Abc_NtkForEachObj
    for  i = 0; (i < Vec_PtrSize((pNtk).vObjs)) && (((pObj) = Abc_NtkObj(pNtk, i)), 1); i++  {            if (pObj) == nil {} else {
        if Abc_NodeIsTravIdCurrent(pObj) {
            continue; }
        if pObj.pCopy {
            continue; }
        Abc_NtkDupObj( pNtkNew, pObj, 0 );
    }
    }

    // connect objects
    // inlining of macro Abc_NtkForEachObj
    for  i = 0; (i < Vec_PtrSize((pNtk).vObjs)) && (((pObj) = Abc_NtkObj(pNtk, i)), 1); i++  {            if (pObj) == nil {} else
        if !Abc_NodeIsTravIdCurrent(pObj) {
            // inlining of macro Abc_ObjForEachFanin
            for  k = 0; (k < Abc_ObjFaninNum(pObj)) && (((pFanin) = Abc_ObjFanin(pObj, k)), 1); k++  {
                if !Abc_NodeIsTravIdCurrent(pFanin) {
                    Abc_ObjAddFanin( pObj.pCopy, pFanin.pCopy ); }
            } }
    }

    // call recursively
    Vec_StrPop( vPref );
    Length = Vec_StrSize( vPref );
    // inlining of macro Abc_NtkForEachBox
    for  i = 0; (i < Vec_PtrSize((pNtk).vBoxes)) && (((pObj) = Abc_NtkBox(pNtk, i)), 1); i++  {
        if Abc_ObjIsLatch(pObj) {
            continue; }
        pNtkModel = (Abc_Ntk_t *)pObj.pData;
        // check the match between the number of actual and formal parameters
        assert(Abc_ObjFaninNum(pObj) == Abc_NtkPiNum(pNtkModel), "Abc_ObjFaninNum(pObj) == Abc_NtkPiNum(pNtkModel)");
        assert(Abc_ObjFanoutNum(pObj) == Abc_NtkPoNum(pNtkModel), "Abc_ObjFanoutNum(pObj) == Abc_NtkPoNum(pNtkModel)");
        // clean the node copy fields
        Abc_NtkCleanCopy( pNtkModel );
        // map PIs/POs
        // inlining of macro Abc_ObjForEachFanin
        for  k = 0; (k < Abc_ObjFaninNum(pObj)) && (((pTerm) = Abc_ObjFanin(pObj, k)), 1); k++  {
            Abc_ObjFanout0( Abc_NtkPi(pNtkModel, k) ).pCopy = Abc_ObjFanin0(pTerm).pCopy;
        }
        // inlining of macro Abc_ObjForEachFanout
        for  k = 0; (k < Abc_ObjFanoutNum(pObj)) && (((pTerm) = Abc_ObjFanout(pObj, k)), 1); k++  {
            Abc_ObjFanin0( Abc_NtkPo(pNtkModel, k) ).pCopy = Abc_ObjFanout0(pTerm).pCopy;
        }
        // create name
        Vec_StrShrink( vPref, Length );
        Vec_StrPrintStr( vPref, Abc_NtkName(pNtkModel) );
        // call recursively
        Abc_NtkFlattenLogicHierarchy_rec( pNtkNew, pNtkModel, pCounter, vPref );
    }

    // if it is a BLIF-MV netlist transfer the values of all nets
    if Abc_NtkHasBlifMv(pNtk) && Abc_NtkMvVar(pNtk) {
        if Abc_NtkMvVar( pNtkNew ) == nil {
            Abc_NtkStartMvVars( pNtkNew ); }
        // inlining of macro Abc_NtkForEachNet
        for  i = 0; (i < Vec_PtrSize((pNtk).vObjs)) && (((pObj) = Abc_NtkObj(pNtk, i)), 1); i++  {            if (pObj) == nil || !Abc_ObjIsNet(pObj) {} else {
            Abc_NtkSetMvVarValues( pObj.pCopy, Abc_ObjMvVarNum(pObj) ); }
        }
    }
}

/**Function*************************************************************

  Synopsis    [Returns 0 if CI names are repeated.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Abc_NtkCompareNames(p1 **Abc_Ntk_t, p2 **Abc_Ntk_t) int32 {
    return int32(strings.Compare(Abc_NtkName(*p1), Abc_NtkName(*p2)));
}

/**Function*************************************************************

  Synopsis    [Prints information about boxes.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Abc_NtkPrintBoxInfo(pNtk *Abc_Ntk_t) {
    var vMods *Vec_Ptr_t;
    var pModel, pBoxModel *Abc_Ntk_t;
    var pObj *Abc_Obj_t;
    var vCounts *Vec_Int_t;
    var i, k, Num int32;
    if pNtk.pDesign == nil || pNtk.pDesign.vModules == nil {
//        printf( "There is no hierarchy information.\n" );
        return;
    }
    // sort models by name
    vMods = pNtk.pDesign.vModules;
    Vec_PtrSort( vMods, (int(*)())Abc_NtkCompareNames );
//    Vec_PtrForEachEntry( Abc_Ntk_t *, vMods, pModel, i )
//        printf( "%s\n", Abc_NtkName(pModel) );

    // swap the first model
    Num = Vec_PtrFind( vMods, pNtk );
    assert(Num >= 0 && Num < Vec_PtrSize(vMods), "Num >= 0 && Num < Vec_PtrSize(vMods)");
    pBoxModel = (Abc_Ntk_t *)Vec_PtrEntry(vMods, 0);
    Vec_PtrWriteEntry(vMods, 0, (Abc_Ntk_t *)Vec_PtrEntry(vMods, Num) );
    Vec_PtrWriteEntry(vMods, Num, pBoxModel );

    // print models
    vCounts = Vec_IntStart( Vec_PtrSize(vMods) );
    // inlining of macro Vec_PtrForEachEntry
    for  i = 0; (i < Vec_PtrSize(vMods)) && (((pModel) = (Abc_Ntk_t *)Vec_PtrEntry(vMods, i)), 1); i++  {
        if Abc_NtkBoxNum(pModel) == 0 {
            continue; }
        Vec_IntFill( vCounts, Vec_IntSize(vCounts), 0 );
        // inlining of macro Abc_NtkForEachBox
        for  k = 0; (k < Vec_PtrSize((pModel).vBoxes)) && (((pObj) = Abc_NtkBox(pModel, k)), 1); k++  {
            pBoxModel = (Abc_Ntk_t *)pObj.pData;
            if pBoxModel == nil {
                continue; }
            Num = Vec_PtrFind( vMods, pBoxModel );
            assert(Num >= 0 && Num < Vec_PtrSize(vMods), "Num >= 0 && Num < Vec_PtrSize(vMods)");
            Vec_IntAddToEntry( vCounts, Num, 1 );
        }

//        Abc_NtkPrintStats( pModel, 0, 0, 0, 0, 0, 0, 0, 0, 0 );
        fmt.Printf("MODULE  ");
        fmt.Printf("%-30s : ", Abc_NtkName(pModel));
        fmt.Printf("PI=%6d ", Abc_NtkPiNum(pModel));
        fmt.Printf("PO=%6d ", Abc_NtkPoNum(pModel));
        fmt.Printf("BB=%6d ", Abc_NtkBoxNum(pModel));
        fmt.Printf("ND=%6d ", Abc_NtkNodeNum(pModel)); // sans constants
        fmt.Printf("Lev=%5d ", Abc_NtkLevel(pModel));
        fmt.Printf("\n");

        // inlining of macro Vec_IntForEachEntry
        for  k = 0; (k < Vec_IntSize(vCounts)) && (((Num) = Vec_IntEntry(vCounts, k)), 1); k++  {
            if Num != 0 {
                fmt.Printf("%15d : %s\n", Num, Abc_NtkName((Abc_Ntk_t *)Vec_PtrEntry(vMods, k))); }
        }
    }
    Vec_IntFree( vCounts );
    // inlining of macro Vec_PtrForEachEntry
    for  i = 0; (i < Vec_PtrSize(vMods)) && (((pModel) = (Abc_Ntk_t *)Vec_PtrEntry(vMods, i)), 1); i++  {
        if Abc_NtkBoxNum(pModel) != 0 {
            continue; }
        fmt.Printf("MODULE   ");
        fmt.Printf("%-30s : ", Abc_NtkName(pModel));
        fmt.Printf("PI=%6d ", Abc_NtkPiNum(pModel));
        fmt.Printf("PO=%6d ", Abc_NtkPoNum(pModel));
        fmt.Printf("BB=%6d ", Abc_NtkBoxNum(pModel));
        fmt.Printf("ND=%6d ", Abc_NtkNodeNum(pModel));
        fmt.Printf("Lev=%5d ", Abc_NtkLevel(pModel));
        fmt.Printf("\n");
    }
}

/**Function*************************************************************

  Synopsis    [Flattens the logic hierarchy of the netlist.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Abc_NtkFlattenLogicHierarchy(pNtk *Abc_Ntk_t) *Abc_Ntk_t {
    extern Abc_Des_t * Abc_DesDupBlackboxes( Abc_Des_t * p, Abc_Ntk_t * pNtkSave );
    var vPref *Vec_Str_t;
    var pNtkNew *Abc_Ntk_t; 
    var pTerm, pNet *Abc_Obj_t;
    var i int32;
    var Counter int32 = -1;

    assert(Abc_NtkIsNetlist(pNtk), "Abc_NtkIsNetlist(pNtk)");
//    Abc_NtkPrintBoxInfo( pNtk );

    // start the network
    pNtkNew = Abc_NtkAlloc( pNtk.ntkType, pNtk.ntkFunc, 1 );
    // duplicate the name and the spec
    pNtkNew.pName = Extra_UtilStrsav(pNtk.pName);
    pNtkNew.pSpec = Extra_UtilStrsav(pNtk.pSpec);

    // clean the node copy fields
    Abc_NtkCleanCopy( pNtk );

    // duplicate PIs/POs and their nets
    // inlining of macro Abc_NtkForEachPi
    for  i = 0; (i < Abc_NtkPiNum(pNtk)) && (((pTerm) = Abc_NtkPi(pNtk, i)), 1); i++  {
        Abc_NtkDupObj( pNtkNew, pTerm, 0 );
        pNet = Abc_ObjFanout0( pTerm );
        pNet.pCopy = Abc_NtkFindOrCreateNet( pNtkNew, Abc_ObjName(pNet) );
        Abc_ObjAddFanin( pNet.pCopy, pTerm.pCopy );
    }
    // inlining of macro Abc_NtkForEachPo
    for  i = 0; (i < Abc_NtkPoNum(pNtk)) && (((pTerm) = Abc_NtkPo(pNtk, i)), 1); i++  {
        Abc_NtkDupObj( pNtkNew, pTerm, 0 );
        pNet = Abc_ObjFanin0( pTerm );
        pNet.pCopy = Abc_NtkFindOrCreateNet( pNtkNew, Abc_ObjName(pNet) );
        Abc_ObjAddFanin( pTerm.pCopy, pNet.pCopy );
    }

    // recursively flatten hierarchy, create internal logic, add new PI/PO names if there are black boxes
    vPref = Vec_StrAlloc( 1000 );
    Vec_StrPrintStr( vPref, Abc_NtkName(pNtk) );
    Abc_NtkFlattenLogicHierarchy_rec( pNtkNew, pNtk, &Counter, vPref );
    fmt.Printf("Hierarchy reader flattened %d instances of logic boxes and left %d black boxes.\n", Counter, Abc_NtkBlackboxNum(pNtkNew));
    Vec_StrFree( vPref );

    if pNtk.pDesign {
        // pass on the design
        assert(Vec_PtrEntry(pNtk.pDesign.vTops, 0) == pNtk, "Vec_PtrEntry(pNtk.pDesign.vTops, 0) == pNtk");
        pNtkNew.pDesign = Abc_DesDupBlackboxes( pNtk.pDesign, pNtkNew );
        // update the pointers
        // inlining of macro Abc_NtkForEachBlackbox
        for  i = 0; (i < Vec_PtrSize((pNtkNew).vBoxes)) && (((pTerm) = Abc_NtkBox(pNtkNew, i)), 1); i++  {           if !Abc_ObjIsBlackbox(pTerm) {} else {
            pTerm.pData = ((Abc_Ntk_t *)pTerm.pData).pCopy; }
        }
    }

    // we may have added property outputs
    Abc_NtkOrderCisCos( pNtkNew );

    // copy the timing information
//    Abc_ManTimeDup( pNtk, pNtkNew );
    // duplicate EXDC 
    if pNtk.pExdc {
        fmt.Printf("EXDC is not transformed.\n"); }
    if !Abc_NtkCheck( pNtkNew ) {
        fmt.Fprintf(os.Stdout, "Abc_NtkFlattenLogicHierarchy(): Network check has failed.\n");
        Abc_NtkDelete( pNtkNew );
        return nil;
    }
    return pNtkNew;
}


/**Function*************************************************************

  Synopsis    [Extracts blackboxes by making them into additional PIs/POs.]

  Description [The input netlist has not logic hierarchy. The resulting
  netlist has additional PIs/POs for each blackbox input/output.]
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Abc_NtkConvertBlackboxes(pNtk *Abc_Ntk_t) *Abc_Ntk_t {
    var pNtkNew *Abc_Ntk_t;
    var pObj, pNet, pFanin, pTerm *Abc_Obj_t;
    var i, k int32;

    assert(Abc_NtkIsNetlist(pNtk), "Abc_NtkIsNetlist(pNtk)");
    assert(Abc_NtkWhiteboxNum(pNtk) == 0, "Abc_NtkWhiteboxNum(pNtk) == 0");

    // start the network
    pNtkNew = Abc_NtkAlloc( pNtk.ntkType, pNtk.ntkFunc, 1 );
    // duplicate the name and the spec
    pNtkNew.pName = Extra_UtilStrsav( pNtk.pName );
    pNtkNew.pSpec = Extra_UtilStrsav( pNtk.pSpec );

    // clean the node copy fields
    Abc_NtkCleanCopy( pNtk );

    // mark the nodes that should not be connected
    Abc_NtkIncrementTravId( pNtk );
    // inlining of macro Abc_NtkForEachBlackbox
    for  i = 0; (i < Vec_PtrSize((pNtk).vBoxes)) && (((pObj) = Abc_NtkBox(pNtk, i)), 1); i++  {           if !Abc_ObjIsBlackbox(pObj) {} else {
        Abc_NodeSetTravIdCurrent( pObj ); }
    }
    // inlining of macro Abc_NtkForEachCi
    for  i = 0; (i < Abc_NtkCiNum(pNtk)) && (((pTerm) = Abc_NtkCi(pNtk, i)), 1); i++  {
        Abc_NodeSetTravIdCurrent( pTerm );
    }
    // inlining of macro Abc_NtkForEachCo
    for  i = 0; (i < Abc_NtkCoNum(pNtk)) && (((pTerm) = Abc_NtkCo(pNtk, i)), 1); i++  {
        Abc_NodeSetTravIdCurrent( pTerm );
    }
    // unmark PIs and LIs/LOs
    // inlining of macro Abc_NtkForEachPi
    for  i = 0; (i < Abc_NtkPiNum(pNtk)) && (((pTerm) = Abc_NtkPi(pNtk, i)), 1); i++  {
        Abc_NodeSetTravIdPrevious( pTerm );
    }
    // inlining of macro Abc_NtkForEachLatchInput
    for  i = 0; (i < Vec_PtrSize((pNtk).vBoxes)); i++  {                                                  if !(Abc_ObjIsLatch(Abc_NtkBox(pNtk, i)) && (((pTerm) = Abc_ObjFanin0(Abc_NtkBox(pNtk, i))), 1)) {} else {
        Abc_NodeSetTravIdPrevious( pTerm ); }
    }
    // inlining of macro Abc_NtkForEachLatchOutput
    for  i = 0; (i < Vec_PtrSize((pNtk).vBoxes)); i++  {                                                  if !(Abc_ObjIsLatch(Abc_NtkBox(pNtk, i)) && (((pTerm) = Abc_ObjFanout0(Abc_NtkBox(pNtk, i))), 1)) {} else {
        Abc_NodeSetTravIdPrevious( pTerm ); }
    }
    // copy the box outputs
    // inlining of macro Abc_NtkForEachBlackbox
    for  i = 0; (i < Vec_PtrSize((pNtk).vBoxes)) && (((pObj) = Abc_NtkBox(pNtk, i)), 1); i++  {           if !Abc_ObjIsBlackbox(pObj) {} else {
        // inlining of macro Abc_ObjForEachFanout
        for  k = 0; (k < Abc_ObjFanoutNum(pObj)) && (((pTerm) = Abc_ObjFanout(pObj, k)), 1); k++  {
            pTerm.pCopy = Abc_NtkCreatePi( pNtkNew );
        } }
    }

    // duplicate other objects
    // inlining of macro Abc_NtkForEachObj
    for  i = 0; (i < Vec_PtrSize((pNtk).vObjs)) && (((pObj) = Abc_NtkObj(pNtk, i)), 1); i++  {            if (pObj) == nil {} else
        if !Abc_NodeIsTravIdCurrent(pObj) {
            Abc_NtkDupObj( pNtkNew, pObj, Abc_ObjIsNet(pObj) ); }
    }

    // connect all objects
    // inlining of macro Abc_NtkForEachObj
    for  i = 0; (i < Vec_PtrSize((pNtk).vObjs)) && (((pObj) = Abc_NtkObj(pNtk, i)), 1); i++  {            if (pObj) == nil {} else
        if !Abc_NodeIsTravIdCurrent(pObj) {
            // inlining of macro Abc_ObjForEachFanin
            for  k = 0; (k < Abc_ObjFaninNum(pObj)) && (((pFanin) = Abc_ObjFanin(pObj, k)), 1); k++  {
                Abc_ObjAddFanin( pObj.pCopy, pFanin.pCopy );
            } }
    }

    // create unique PO for each net feeding into blackboxes or POs
    Abc_NtkIncrementTravId( pNtk );
    // inlining of macro Abc_NtkForEachCo
    for  i = 0; (i < Abc_NtkCoNum(pNtk)) && (((pTerm) = Abc_NtkCo(pNtk, i)), 1); i++  {
        // skip latch inputs
        assert(Abc_ObjFanoutNum(pTerm) <= 1, "Abc_ObjFanoutNum(pTerm) <= 1");
        if Abc_ObjFanoutNum(pTerm) > 0 && Abc_ObjIsLatch(Abc_ObjFanout0(pTerm)) {
            continue; }
        // check if the net is visited
        pNet = Abc_ObjFanin0(pTerm);
        if Abc_NodeIsTravIdCurrent(pNet) {
            continue; }
        // create PO
        Abc_NodeSetTravIdCurrent( pNet );
        Abc_ObjAddFanin( Abc_NtkCreatePo(pNtkNew), pNet.pCopy );
    }

    // check integrity
    if !Abc_NtkCheck( pNtkNew ) {
        fmt.Fprintf(os.Stdout, "Abc_NtkConvertBlackboxes(): Network check has failed.\n");
        Abc_NtkDelete( pNtkNew );
        return nil;
    }
    return pNtkNew;
}

/**Function*************************************************************

  Synopsis    [Inserts blackboxes into the netlist.]

  Description [The first arg is the netlist with blackboxes without logic hierarchy.
  The second arg is a non-hierarchical netlist derived from the above netlist after processing.
  This procedure create a new netlist, which is comparable to the original netlist with
  blackboxes, except that it contains logic nodes from the netlist after processing.]
               
  SideEffects [This procedure silently assumes that blackboxes appear
  only in the top-level model. If they appear in other models as well, 
  the name of the model and its number were appended to the names of 
  blackbox inputs/outputs.]

  SeeAlso     []

***********************************************************************/
func Abc_NtkInsertNewLogic(pNtkH *Abc_Ntk_t, pNtkL *Abc_Ntk_t) *Abc_Ntk_t {
    var pDesign *Abc_Des_t;
    var pNtkNew *Abc_Ntk_t;
    var pObjH, pObjL, pNetH, pNetL, pTermH *Abc_Obj_t;
    var i, k int32;

    assert(Abc_NtkIsNetlist(pNtkH), "Abc_NtkIsNetlist(pNtkH)");
    assert(Abc_NtkWhiteboxNum(pNtkH) == 0, "Abc_NtkWhiteboxNum(pNtkH) == 0");
    assert(Abc_NtkBlackboxNum(pNtkH) > 0, "Abc_NtkBlackboxNum(pNtkH) > 0");

    assert(Abc_NtkIsNetlist(pNtkL), "Abc_NtkIsNetlist(pNtkL)");
    assert(Abc_NtkWhiteboxNum(pNtkL) == 0, "Abc_NtkWhiteboxNum(pNtkL) == 0");
    assert(Abc_NtkBlackboxNum(pNtkL) == 0, "Abc_NtkBlackboxNum(pNtkL) == 0");

    // prepare the logic network for copying
    Abc_NtkCleanCopy( pNtkL );

    // start the network
    pNtkNew = Abc_NtkAlloc( pNtkL.ntkType, pNtkL.ntkFunc, 1 );
    // duplicate the name and the spec
    pNtkNew.pName = Extra_UtilStrsav( pNtkH.pName );
    pNtkNew.pSpec = Extra_UtilStrsav( pNtkH.pSpec );

    // make sure every PI/PO has a PI/PO in the processed network
    // inlining of macro Abc_NtkForEachPi
    for  i = 0; (i < Abc_NtkPiNum(pNtkH)) && (((pObjH) = Abc_NtkPi(pNtkH, i)), 1); i++  {
        pNetH = Abc_ObjFanout0(pObjH);
        pNetL = Abc_NtkFindNet( pNtkL, Abc_ObjName(pNetH) );
        if pNetL == nil || !Abc_ObjIsPi( Abc_ObjFanin0(pNetL) ) {
            fmt.Printf("Error in Abc_NtkInsertNewLogic(): There is no PI corresponding to the PI %s.\n", Abc_ObjName(pNetH));
            Abc_NtkDelete( pNtkNew );
            return nil;
        }
        if pNetL.pCopy {
            fmt.Printf("Error in Abc_NtkInsertNewLogic(): Primary input %s is repeated twice.\n", Abc_ObjName(pNetH));
            Abc_NtkDelete( pNtkNew );
            return nil;
        }
        // create the new net
        pNetL.pCopy = Abc_NtkFindOrCreateNet( pNtkNew, Abc_ObjName(pNetH) );
        Abc_NtkDupObj( pNtkNew, Abc_ObjFanin0(pNetL), 0 );
    }

    // make sure every BB has a PI/PO in the processed network
    // inlining of macro Abc_NtkForEachBlackbox
    for  i = 0; (i < Vec_PtrSize((pNtkH).vBoxes)) && (((pObjH) = Abc_NtkBox(pNtkH, i)), 1); i++  {           if !Abc_ObjIsBlackbox(pObjH) {} else {
        // duplicate the box 
        Abc_NtkDupBox( pNtkNew, pObjH, 0 );
        pObjH.pCopy.pData = pObjH.pData;
        // create PIs
        // inlining of macro Abc_ObjForEachFanout
        for  k = 0; (k < Abc_ObjFanoutNum(pObjH)) && (((pTermH) = Abc_ObjFanout(pObjH, k)), 1); k++  {
            pNetH = Abc_ObjFanout0( pTermH );
            pNetL = Abc_NtkFindNet( pNtkL, Abc_ObjName(pNetH) );
            if pNetL == nil || !Abc_ObjIsPi( Abc_ObjFanin0(pNetL) ) {
                fmt.Printf("Error in Abc_NtkInsertNewLogic(): There is no PI corresponding to the inpout %s of blackbox %s.\n", Abc_ObjName(pNetH), Abc_ObjName(pObjH));
                Abc_NtkDelete( pNtkNew );
                return nil;
            }
            if pNetL.pCopy {
                fmt.Printf("Error in Abc_NtkInsertNewLogic(): Box output %s is repeated twice.\n", Abc_ObjName(pNetH));
                Abc_NtkDelete( pNtkNew );
                return nil;
            }
            // create net and map the PI
            pNetL.pCopy = Abc_NtkFindOrCreateNet( pNtkNew, Abc_ObjName(pNetH) );
            Abc_ObjFanin0(pNetL).pCopy = pTermH.pCopy;
        }
    }
    }

    // inlining of macro Abc_NtkForEachPo
    for  i = 0; (i < Abc_NtkPoNum(pNtkH)) && (((pObjH) = Abc_NtkPo(pNtkH, i)), 1); i++  {
        pNetH = Abc_ObjFanin0(pObjH);
        pNetL = Abc_NtkFindNet( pNtkL, Abc_ObjName(pNetH) );
        if pNetL == nil || !Abc_ObjIsPo( Abc_ObjFanout0(pNetL) ) {
            fmt.Printf("Error in Abc_NtkInsertNewLogic(): There is no PO corresponding to the PO %s.\n", Abc_ObjName(pNetH));
            Abc_NtkDelete( pNtkNew );
            return nil;
        }
        if pNetL.pCopy {
            continue; }
        // create the new net
        pNetL.pCopy = Abc_NtkFindOrCreateNet( pNtkNew, Abc_ObjName(pNetH) );
        Abc_NtkDupObj( pNtkNew, Abc_ObjFanout0(pNetL), 0 );
    }
    // inlining of macro Abc_NtkForEachBlackbox
    for  i = 0; (i < Vec_PtrSize((pNtkH).vBoxes)) && (((pObjH) = Abc_NtkBox(pNtkH, i)), 1); i++  {           if !Abc_ObjIsBlackbox(pObjH) {} else {
        // inlining of macro Abc_ObjForEachFanin
        for  k = 0; (k < Abc_ObjFaninNum(pObjH)) && (((pTermH) = Abc_ObjFanin(pObjH, k)), 1); k++  {
            var pName string;
            pNetH = Abc_ObjFanin0( pTermH );
            pName = Abc_ObjName(pNetH);
            pNetL = Abc_NtkFindNet( pNtkL, Abc_ObjName(pNetH) );
            if pNetL == nil || !Abc_ObjIsPo( Abc_ObjFanout0(pNetL) ) {
                fmt.Printf("There is no PO corresponding to the input %s of blackbox %s.\n", Abc_ObjName(pNetH), Abc_ObjName(pObjH));
                Abc_NtkDelete( pNtkNew );
                return nil;
            }
            // create net and map the PO
            if pNetL.pCopy {
                if Abc_ObjFanout0(pNetL).pCopy == nil {
                    Abc_ObjFanout0(pNetL).pCopy = pTermH.pCopy;
                } else {
                    Abc_ObjAddFanin( pTermH.pCopy, pNetL.pCopy ); }
                continue;
            }
            pNetL.pCopy = Abc_NtkFindOrCreateNet( pNtkNew, Abc_ObjName(pNetH) );
            Abc_ObjFanout0(pNetL).pCopy = pTermH.pCopy;
        }
    }
    }

    // duplicate other objects of the logic network
    // inlining of macro Abc_NtkForEachObj
    for  i = 0; (i < Vec_PtrSize((pNtkL).vObjs)) && (((pObjL) = Abc_NtkObj(pNtkL, i)), 1); i++  {            if (pObjL) == nil {} else
        if pObjL.pCopy == nil && !Abc_ObjIsPo(pObjL) { // skip POs feeding into PIs
            Abc_NtkDupObj( pNtkNew, pObjL, Abc_ObjIsNet(pObjL) ); }
    }

    // connect objects
    // inlining of macro Abc_NtkForEachObj
    for  i = 0; (i < Vec_PtrSize((pNtkL).vObjs)) && (((pObjL) = Abc_NtkObj(pNtkL, i)), 1); i++  {            if (pObjL) == nil {} else {
        // inlining of macro Abc_ObjForEachFanin
        for  k = 0; (k < Abc_ObjFaninNum(pObjL)) && (((pNetL) = Abc_ObjFanin(pObjL, k)), 1); k++  {
            if pObjL.pCopy {
                Abc_ObjAddFanin( pObjL.pCopy, pNetL.pCopy ); }
        } }
    }

    // transfer the design
    pDesign = pNtkH.pDesign;  pNtkH.pDesign = nil;
    assert(Vec_PtrEntry( pDesign.vModules, 0 ) == pNtkH, "Vec_PtrEntry( pDesign.vModules, 0 ) == pNtkH");
    Vec_PtrWriteEntry( pDesign.vModules, 0, pNtkNew );
    pNtkNew.pDesign = pDesign;

    // check integrity
    if !Abc_NtkCheck( pNtkNew ) {
        fmt.Fprintf(os.Stdout, "Abc_NtkInsertNewLogic(): Network check has failed.\n");
        Abc_NtkDelete( pNtkNew );
        return nil;
    }
    return pNtkNew;
}

////////////////////////////////////////////////////////////////////////
///                       END OF FILE                                ///
////////////////////////////////////////////////////////////////////////


//ABC_NAMESPACE_IMPL_END

