package abci

import "fmt"

/**CFile****************************************************************

  FileName    [abcDress.c]

  SystemName  [ABC: Logic synthesis and verification system.]

  PackageName [Network and node package.]

  Synopsis    [Transfers names from one netlist to the other.]

  Author      [Alan Mishchenko]
  
  Affiliation [UC Berkeley]

  Date        [Ver. 1.0. Started - June 20, 2005.]

  Revision    [$Id: abcDress.c,v 1.00 2005/06/20 00:00:00 alanmi Exp $]

***********************************************************************/

//#include "base/abc/abc.h"
//#include "base/io/ioAbc.h"

//ABC_NAMESPACE_IMPL_START


////////////////////////////////////////////////////////////////////////
///                        DECLARATIONS                              ///
////////////////////////////////////////////////////////////////////////

// static stmm_table * Abc_NtkDressDeriveMapping( Abc_Ntk_t * pNtk );

// static void         Abc_NtkDressTransferNames( Abc_Ntk_t * pNtk, stmm_table * tMapping, int fVerbose );


// extern Abc_Ntk_t *  Abc_NtkIvyFraig( Abc_Ntk_t * pNtk, int nConfLimit, int fDoSparse, int fProve, int fTransfer, int fVerbose );


////////////////////////////////////////////////////////////////////////
///                     FUNCTION DEFINITIONS                         ///
////////////////////////////////////////////////////////////////////////

/**Function*************************************************************

  Synopsis    [Transfers names from one netlist to the other.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Abc_NtkDress(pNtkLogic *Abc_Ntk_t, pFileName string, fVerbose int32) {
    var pNtkOrig, pNtkLogicOrig *Abc_Ntk_t;
    var pMiter, pMiterFraig *Abc_Ntk_t;
    var tMapping *stmm_table;

    assert(Abc_NtkIsLogic(pNtkLogic), "Abc_NtkIsLogic(pNtkLogic)");

    // get the original netlist
    pNtkOrig = Io_ReadNetlist( pFileName, Io_ReadFileType(pFileName), 1 );
    if pNtkOrig == nil {
        return; }
    assert(Abc_NtkIsNetlist(pNtkOrig), "Abc_NtkIsNetlist(pNtkOrig)");

    Abc_NtkCleanCopy(pNtkLogic);
    Abc_NtkCleanCopy(pNtkOrig);

    // convert it into the logic network
    pNtkLogicOrig = Abc_NtkToLogic( pNtkOrig );
    // check that the networks have the same PIs/POs/latches
    if !Abc_NtkCompareSignals( pNtkLogic, pNtkLogicOrig, 1, 1 ) {
        Abc_NtkDelete( pNtkOrig );
        Abc_NtkDelete( pNtkLogicOrig );
        return;
    }

    // convert the current logic network into an AIG
    pMiter = Abc_NtkStrash( pNtkLogic, 1, 0, 0 );

    // convert it into the AIG and make the netlist point to the AIG
    Abc_NtkAppend( pMiter, pNtkLogicOrig, 1 );
    Abc_NtkTransferCopy( pNtkOrig );
    Abc_NtkDelete( pNtkLogicOrig );

if fVerbose != 0 {
fmt.Printf("After mitering:\n");
fmt.Printf("Logic:  Nodes = %5d. Copy = %5d. \n", Abc_NtkNodeNum(pNtkLogic), Abc_NtkCountCopy(pNtkLogic));
fmt.Printf("Orig:   Nodes = %5d. Copy = %5d. \n", Abc_NtkNodeNum(pNtkOrig), Abc_NtkCountCopy(pNtkOrig));
}

    // fraig the miter (miter nodes point to the fraiged miter)
    pMiterFraig = Abc_NtkIvyFraig( pMiter, 100, 1, 0, 1, 0 );
    // make netlists point to the fraiged miter
    Abc_NtkTransferCopy( pNtkLogic );
    Abc_NtkTransferCopy( pNtkOrig );
    Abc_NtkDelete( pMiter );

if fVerbose != 0 {
fmt.Printf("After fraiging:\n");
fmt.Printf("Logic:  Nodes = %5d. Copy = %5d. \n", Abc_NtkNodeNum(pNtkLogic), Abc_NtkCountCopy(pNtkLogic));
fmt.Printf("Orig:   Nodes = %5d. Copy = %5d. \n", Abc_NtkNodeNum(pNtkOrig), Abc_NtkCountCopy(pNtkOrig));
}

    // derive mapping from the fraiged nodes into their prototype nodes in the original netlist
    tMapping = Abc_NtkDressDeriveMapping( pNtkOrig );

    // transfer the names to the new netlist
    Abc_NtkDressTransferNames( pNtkLogic, tMapping, fVerbose );

    // clean up
    stmm_free_table( tMapping );
    Abc_NtkDelete( pMiterFraig );
    Abc_NtkDelete( pNtkOrig );
}

/**Function*************************************************************

  Synopsis    [Returns the mapping from the fraig nodes point into the nodes of the netlist.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Abc_NtkDressDeriveMapping(pNtk *Abc_Ntk_t) *stmm_table {
    var tResult *stmm_table;
    var pNode, pNodeMap, pNodeFraig *Abc_Obj_t;
    var i int32;
    assert(Abc_NtkIsNetlist(pNtk), "Abc_NtkIsNetlist(pNtk)");
    tResult = stmm_init_table(stmm_ptrcmp,stmm_ptrhash);
    // inlining of macro Abc_NtkForEachNode
    for  i = 0; (i < Vec_PtrSize((pNtk).vObjs)) && (((pNode) = Abc_NtkObj(pNtk, i)), 1); i++  {           if (pNode) == nil || !Abc_ObjIsNode(pNode) {} else {
        // get the fraiged node
        pNodeFraig = Abc_ObjRegular(pNode.pCopy);
        // if this node is already mapped, skip
        // inlining of macro stmm_is_member
        if  stmm_lookup(tResult,(char *)pNodeFraig,(char **) 0) {
            continue; }
        // get the mapping of this node
        pNodeMap = Abc_ObjNotCond( pNode, Abc_ObjIsComplement(pNode.pCopy) );
        // add the mapping
        stmm_insert( tResult, (char *)pNodeFraig, (char *)pNodeMap );
    }
    }
    return tResult;
}

/**Function*************************************************************

  Synopsis    [Attaches the names of to the new netlist.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Abc_NtkDressTransferNames(pNtk *Abc_Ntk_t, tMapping *stmm_table, fVerbose int32) {
    var pNet, pNode, pNodeMap, pNodeFraig *Abc_Obj_t;
    var pName string;
    // inlining of macro stmm_count
    int i, Counter = 0, CounterInv = 0, CounterInit = ((tMapping).num_entries);
    assert(Abc_NtkIsLogic(pNtk), "Abc_NtkIsLogic(pNtk)");
    // inlining of macro Abc_NtkForEachNode
    for  i = 0; (i < Vec_PtrSize((pNtk).vObjs)) && (((pNode) = Abc_NtkObj(pNtk, i)), 1); i++  {           if (pNode) == nil || !Abc_ObjIsNode(pNode) {} else {
        // if the node already has a name, quit
        pName = Nm_ManFindNameById( pNtk.pManName, pNode.Id );
        if pName != nil {
            continue; }
        // get the fraiged node
        pNodeFraig = Abc_ObjRegular(pNode.pCopy);
        // find the matching node of the original netlist
        if !stmm_lookup( tMapping, (char *)pNodeFraig, (char **)&pNodeMap ) {
            continue; }
        // find the true match
        pNodeMap = Abc_ObjNotCond( pNodeMap, Abc_ObjIsComplement(pNode.pCopy) );
        // get the name
        pNet = Abc_ObjFanout0(Abc_ObjRegular(pNodeMap));
        pName = Nm_ManFindNameById( pNet.pNtk.pManName, pNet.Id );
        assert(pName != nil, "pName != nil");
        // set the name
        if Abc_ObjIsComplement(pNodeMap) {
            Abc_ObjAssignName( pNode, pName, "_inv" );
            CounterInv++;
        
        } else {
            Abc_ObjAssignName( pNode, pName, nil );
            Counter++;
        }
        // remove the name
        stmm_delete( tMapping, (char **)&pNodeFraig, (char **)&pNodeMap );
    }
    }
    if fVerbose != 0 {
        fmt.Printf("Total number of names collected = %5d.\n", CounterInit);
        fmt.Printf("Total number of names assigned  = %5d. (Dir = %5d. Compl = %5d.)\n", Counter + CounterInv, Counter, CounterInv);
    }
}

////////////////////////////////////////////////////////////////////////
///                       END OF FILE                                ///
////////////////////////////////////////////////////////////////////////


//ABC_NAMESPACE_IMPL_END

