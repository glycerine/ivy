package abc

import "fmt"

/**CFile****************************************************************

  FileName    [abcLib.c]

  SystemName  [ABC: Logic synthesis and verification system.]

  PackageName [Network and node package.]

  Synopsis    [Functions to manipulate verilog libraries.]

  Author      [Alan Mishchenko]
  
  Affiliation [UC Berkeley]

  Date        [Ver. 1.0. Started - June 20, 2005.]

  Revision    [$Id: abcLib.c,v 1.00 2005/06/20 00:00:00 alanmi Exp $]

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

  Synopsis    [Create the library.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Abc_DesCreate(pName string) *Abc_Des_t {
    var p *Abc_Des_t;
    // inlining of macro ABC_ALLOC
    p = ((Abc_Des_t *) malloc(sizeof(Abc_Des_t) * (size_t)(1)));
    memset( p, 0, sizeof(Abc_Des_t) );
    p.pName    = Abc_UtilStrsav( pName );
    p.tModules = st__init_table( st__strcmp, st__strhash );
    p.vTops    = Vec_PtrAlloc( 100 );
    p.vModules = Vec_PtrAlloc( 100 );
    p.pManFunc = Hop_ManStart();
    p.pLibrary = nil;
    return p;
}

/**Function*************************************************************

  Synopsis    [Removes all pointers to the manager.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Abc_DesCleanManPointer(p *Abc_Des_t, pMan any) {
    var pTemp *Abc_Ntk_t;
    var i int32;
    if p == nil {
        return; }
    if p.pManFunc == pMan {
        p.pManFunc = nil; }
    // inlining of macro Vec_PtrForEachEntry
    for  i = 0; (i < Vec_PtrSize(p.vModules)) && (((pTemp) = (Abc_Ntk_t *)Vec_PtrEntry(p.vModules, i)), 1); i++  {
        if pTemp.pManFunc == pMan {
            pTemp.pManFunc = nil; }
    }
}

/**Function*************************************************************

  Synopsis    [Frees the library.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Abc_DesFree(p *Abc_Des_t, pNtkSave *Abc_Ntk_t) {
    var pNtk *Abc_Ntk_t;
    var i int32;
    if p.pName {
        // inlining of macro ABC_FREE
        ((p.pName) ? (free((char *) (p.pName)), (p.pName) = 0) : 0); }
    if p.pManFunc {
        Hop_ManStop( (Hop_Man_t *)p.pManFunc ); }
    if p.tModules {
        st__free_table( p.tModules ); }
    if p.vModules {
        // inlining of macro Vec_PtrForEachEntry
        for  i = 0; (i < Vec_PtrSize(p.vModules)) && (((pNtk) = (Abc_Ntk_t *)Vec_PtrEntry(p.vModules, i)), 1); i++  {
            if pNtk == pNtkSave {
                continue; }
            pNtk.pDesign = nil;
            if ((pNtkSave != nil) && pNtk.pManFunc == pNtkSave.pManFunc) || (pNtk.pManFunc == p.pManFunc) {
                pNtk.pManFunc = nil; }
            Abc_NtkDelete( pNtk );
        }
        Vec_PtrFree( p.vModules );
    }
    if p.vTops {
        Vec_PtrFree( p.vTops ); }
    // inlining of macro ABC_FREE
    ((p) ? (free((char *) (p)), (p) = 0) : 0);
}

/**Function*************************************************************

  Synopsis    [Duplicated the library.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Abc_DesDup(p *Abc_Des_t) *Abc_Des_t {
    var pNew *Abc_Des_t;
    var pTemp *Abc_Ntk_t;
    var pObj *Abc_Obj_t;
    var i, k int32;
    pNew = Abc_DesCreate( p.pName );
    // inlining of macro Vec_PtrForEachEntry
    for  i = 0; (i < Vec_PtrSize(p.vModules)) && (((pTemp) = (Abc_Ntk_t *)Vec_PtrEntry(p.vModules, i)), 1); i++  {
        Abc_DesAddModel( pNew, Abc_NtkDup(pTemp) );
    }
    // inlining of macro Vec_PtrForEachEntry
    for  i = 0; (i < Vec_PtrSize(p.vTops)) && (((pTemp) = (Abc_Ntk_t *)Vec_PtrEntry(p.vTops, i)), 1); i++  {
        Vec_PtrPush( pNew.vTops, pTemp.pCopy );
    }
    // inlining of macro Vec_PtrForEachEntry
    for  i = 0; (i < Vec_PtrSize(p.vModules)) && (((pTemp) = (Abc_Ntk_t *)Vec_PtrEntry(p.vModules, i)), 1); i++  {
        if pTemp.pAltView {
            pTemp.pCopy.pAltView = pTemp.pAltView.pCopy;
        } else {
            pTemp.pCopy.pAltView = nil;
        }
    }
    // update box models
    // inlining of macro Vec_PtrForEachEntry
    for  i = 0; (i < Vec_PtrSize(p.vModules)) && (((pTemp) = (Abc_Ntk_t *)Vec_PtrEntry(p.vModules, i)), 1); i++  {
        // inlining of macro Abc_NtkForEachBox
        for  k = 0; (k < Vec_PtrSize((pTemp).vBoxes)) && (((pObj) = Abc_NtkBox(pTemp, k)), 1); k++  {
            if Abc_ObjIsWhitebox(pObj) || Abc_ObjIsBlackbox(pObj) {
                pObj.pCopy.pData = Abc_ObjModel(pObj).pCopy; }
        }
    }
    return pNew;
}

/**Function*************************************************************

  Synopsis    [Frees the library.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Abc_DesDupBlackboxes(p *Abc_Des_t, pNtkSave *Abc_Ntk_t) *Abc_Des_t {
    var pNew *Abc_Des_t;
    var pNtkTemp *Abc_Ntk_t;
    var i int32;
    assert(Vec_PtrSize(p.vTops) > 0, "Vec_PtrSize(p.vTops) > 0");
    assert(Vec_PtrSize(p.vModules) > 1, "Vec_PtrSize(p.vModules) > 1");
    pNew = Abc_DesCreate( p.pName );
//    pNew->pManFunc = pNtkSave->pManFunc;
    Vec_PtrPush( pNew.vTops, pNtkSave );
    Vec_PtrPush( pNew.vModules, pNtkSave );
    // inlining of macro Vec_PtrForEachEntry
    for  i = 0; (i < Vec_PtrSize(p.vModules)) && (((pNtkTemp) = (Abc_Ntk_t *)Vec_PtrEntry(p.vModules, i)), 1); i++  {
        if Abc_NtkHasBlackbox( pNtkTemp ) {
            Vec_PtrPush( pNew.vModules, Abc_NtkDup(pNtkTemp) ); }
    }
    return pNew;
}


/**Function*************************************************************

  Synopsis    [Prints the library.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Abc_DesPrint(p *Abc_Des_t) {
    var pNtk *Abc_Ntk_t;
    var pObj *Abc_Obj_t;
    var i, k int32;
    fmt.Printf("Models of design %s:\n", p.pName);
    // inlining of macro Vec_PtrForEachEntry
    for  i = 0; (i < Vec_PtrSize(p.vModules)) && (((pNtk) = (Abc_Ntk_t *)Vec_PtrEntry(p.vModules, i)), 1); i++  {
        fmt.Printf("%2d : %20s   ", i+1, pNtk.pName);
        fmt.Printf("nd = %6d   lat = %6d   whitebox = %3d   blackbox = %3d\n", Abc_NtkNodeNum(pNtk), Abc_NtkLatchNum(pNtk), Abc_NtkWhiteboxNum(pNtk), Abc_NtkBlackboxNum(pNtk));
        if Abc_NtkBlackboxNum(pNtk) == 0 {
            continue; }
        // inlining of macro Abc_NtkForEachWhitebox
        for  k = 0; (k < Vec_PtrSize((pNtk).vBoxes)) && (((pObj) = Abc_NtkBox(pNtk, k)), 1); k++  {           if !Abc_ObjIsWhitebox(pObj) {} else {
            fmt.Printf("     %20s (whitebox)\n", Abc_NtkName((Abc_Ntk_t *)pObj.pData)); }
        }
        // inlining of macro Abc_NtkForEachBlackbox
        for  k = 0; (k < Vec_PtrSize((pNtk).vBoxes)) && (((pObj) = Abc_NtkBox(pNtk, k)), 1); k++  {           if !Abc_ObjIsBlackbox(pObj) {} else {
            fmt.Printf("     %20s (blackbox)\n", Abc_NtkName((Abc_Ntk_t *)pObj.pData)); }
        }
    }
}

/**Function*************************************************************

  Synopsis    [Create the library.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Abc_DesAddModel(p *Abc_Des_t, pNtk *Abc_Ntk_t) int32 {
    // inlining of macro st__is_member
    if  st__lookup(p.tModules,(char *)pNtk.pName,(char **) 0) {
        return 0; }
    st__insert( p.tModules, (char *)pNtk.pName, (char *)pNtk );
    assert(pNtk.Id == 0, "pNtk.Id == 0");
    pNtk.Id = Vec_PtrSize(p.vModules);
    Vec_PtrPush( p.vModules, pNtk );
    pNtk.pDesign = p;
    return 1;
}

/**Function*************************************************************

  Synopsis    [Create the library.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Abc_DesFindModelByName(p *Abc_Des_t, pName string) *Abc_Ntk_t {
    var pNtk *Abc_Ntk_t;
    // inlining of macro st__is_member
    if ! st__lookup(p.tModules,(char *)pName,(char **) 0) {
        return nil; }
    st__lookup( p.tModules, (char *)pName, (char **)&pNtk );
    return pNtk;
}

/**Function*************************************************************

  Synopsis    [Frees the library.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Abc_DesDeriveRoot(p *Abc_Des_t) *Abc_Ntk_t {
    var pNtk *Abc_Ntk_t;
    if Vec_PtrSize(p.vModules) > 1 {
        fmt.Printf("The design includes more than one module and is currently not used.\n");
        return nil;
    }
    pNtk = (Abc_Ntk_t *)Vec_PtrEntry( p.vModules, 0 );  Vec_PtrClear( p.vModules );
    pNtk.pManFunc = p.pManFunc;           p.pManFunc = nil;
    return pNtk;
}

/**Function*************************************************************

  Synopsis    [Detects the top-level models.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Abc_DesFindTopLevelModels(p *Abc_Des_t) int32 {
    var pNtk, pNtkBox *Abc_Ntk_t;
    var pObj *Abc_Obj_t;
    var i, k int32;
    assert(Vec_PtrSize( p.vModules ) > 0, "Vec_PtrSize( p.vModules ) > 0");
    // clear the models
    // inlining of macro Vec_PtrForEachEntry
    for  i = 0; (i < Vec_PtrSize(p.vModules)) && (((pNtk) = (Abc_Ntk_t *)Vec_PtrEntry(p.vModules, i)), 1); i++  {
        pNtk.fHieVisited = 0;
    }
    // mark all the models reachable from other models
    // inlining of macro Vec_PtrForEachEntry
    for  i = 0; (i < Vec_PtrSize(p.vModules)) && (((pNtk) = (Abc_Ntk_t *)Vec_PtrEntry(p.vModules, i)), 1); i++  {
        // inlining of macro Abc_NtkForEachBox
        for  k = 0; (k < Vec_PtrSize((pNtk).vBoxes)) && (((pObj) = Abc_NtkBox(pNtk, k)), 1); k++  {
            if Abc_ObjIsLatch(pObj) {
                continue; }
            if pObj.pData == nil {
                continue; }
            pNtkBox = (Abc_Ntk_t *)pObj.pData;
            pNtkBox.fHieVisited = 1;
        }
    }
    // collect the models that are not marked
    Vec_PtrClear( p.vTops );
    // inlining of macro Vec_PtrForEachEntry
    for  i = 0; (i < Vec_PtrSize(p.vModules)) && (((pNtk) = (Abc_Ntk_t *)Vec_PtrEntry(p.vModules, i)), 1); i++  {
        if pNtk.fHieVisited == 0 {
            Vec_PtrPush( p.vTops, pNtk );
        } else {
            pNtk.fHieVisited = 0; }
    }
    return Vec_PtrSize( p.vTops );
}


/**Function*************************************************************

  Synopsis    [Derive the AIG of the logic in the netlist.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Abc_NodeStrashUsingNetwork_rec(pNtkAig *Abc_Ntk_t, pObj *Abc_Obj_t) {
    var pFanin *Abc_Obj_t;
    var i int32;
    assert(!Abc_ObjIsNet(pObj), "!Abc_ObjIsNet(pObj)");
    if pObj.pCopy {
        return; }
    // call for the fanins
    // inlining of macro Abc_ObjForEachFanin
    for  i = 0; (i < Abc_ObjFaninNum(pObj)) && (((pFanin) = Abc_ObjFanin(pObj, i)), 1); i++  {
        Abc_NodeStrashUsingNetwork_rec( pNtkAig, Abc_ObjFanin0Ntk(Abc_ObjFanin0(pObj)) );
    }
    // compute for the node
    pObj.pCopy = Abc_NodeStrash( pNtkAig, pObj, 0 );
    // set for the fanout net
    Abc_ObjFanout0(pObj).pCopy = pObj.pCopy;
}

/**Function*************************************************************

  Synopsis    [Derive the AIG of the logic in the netlist.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Abc_NodeStrashUsingNetwork(pNtkAig *Abc_Ntk_t, pBox *Abc_Obj_t) { 
    var pNtkGate *Abc_Ntk_t;
    var pObj *Abc_Obj_t;
    var pPolarity *uint32;
    var i, fCompl int32;
    assert(Abc_ObjIsBox(pBox), "Abc_ObjIsBox(pBox)");
    pNtkGate = (Abc_Ntk_t *)pBox.pData;
    pPolarity = (unsigned *)pBox.pNext;
    assert(Abc_NtkIsNetlist(pNtkGate), "Abc_NtkIsNetlist(pNtkGate)");
    assert(Abc_NtkLatchNum(pNtkGate) == 0, "Abc_NtkLatchNum(pNtkGate) == 0");
    Abc_NtkCleanCopy( pNtkGate );
    // set the PI values
    // inlining of macro Abc_NtkForEachPi
    for  i = 0; (i < Abc_NtkPiNum(pNtkGate)) && (((pObj) = Abc_NtkPi(pNtkGate, i)), 1); i++  {
        fCompl = (pPolarity && Abc_InfoHasBit(pPolarity, i));
        pObj.pCopy = Abc_ObjNotCond( Abc_ObjFanin(pBox,i).pCopy, fCompl );
        Abc_ObjFanout0(pObj).pCopy = pObj.pCopy;
    }
    // build recursively and set the PO values
    // inlining of macro Abc_NtkForEachPo
    for  i = 0; (i < Abc_NtkPoNum(pNtkGate)) && (((pObj) = Abc_NtkPo(pNtkGate, i)), 1); i++  {
        Abc_NodeStrashUsingNetwork_rec( pNtkAig, Abc_ObjFanin0Ntk(Abc_ObjFanin0(pObj)) );
        Abc_ObjFanout(pBox,i).pCopy = Abc_ObjFanin0(pObj).pCopy;
    }
//printf( "processing %d\n", pBox->Id );
}

////////////////////////////////////////////////////////////////////////
///                       END OF FILE                                ///
////////////////////////////////////////////////////////////////////////


//ABC_NAMESPACE_IMPL_END
