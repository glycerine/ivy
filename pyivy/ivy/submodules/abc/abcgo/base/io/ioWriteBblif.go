package io

/**CFile****************************************************************

  FileName    [ioWriteBblif.c]

  SystemName  [ABC: Logic synthesis and verification system.]

  PackageName [Command processing package.]

  Synopsis    [Procedures to write AIG in the binary format.]

  Author      [Alan Mishchenko]
  
  Affiliation [UC Berkeley]

  Date        [Ver. 1.0. Started - June 20, 2005.]

  Revision    [$Id: ioWriteBblif.c,v 1.00 2005/06/20 00:00:00 alanmi Exp $]

***********************************************************************/

//#include "ioAbc.h"
//#include "misc/bbl/bblif.h"

//ABC_NAMESPACE_IMPL_START


// For description of Binary BLIF format, refer to "abc/src/aig/bbl/bblif.h"

////////////////////////////////////////////////////////////////////////
///                        DECLARATIONS                              ///
////////////////////////////////////////////////////////////////////////

////////////////////////////////////////////////////////////////////////
///                     FUNCTION DEFINITIONS                         ///
////////////////////////////////////////////////////////////////////////

/**Fnction*************************************************************

  Synopsis    [Construct manager from the ABC network.]

  Description [In the ABC network each object has a unique integer ID.
  This ID is used when we construct objects of the BBLIF manager 
  corresponding to each object of the ABC network. The objects can be
  added to the manager in any order (although below they are added in the
  topological order), but by the time fanin/fanout connections are created, 
  corresponding objects are already constructed. In the end the checking
  procedure is called.]
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Bbl_ManFromAbc(pNtk *Abc_Ntk_t) *Bbl_Man_t {
    var p *Bbl_Man_t;
    var vNodes *Vec_Ptr_t;
    var pObj, pFanin *Abc_Obj_t;
    var i, k int32;
    assert(Abc_NtkIsSopLogic(pNtk), "Abc_NtkIsSopLogic(pNtk)");
    // start the data manager
    p = Bbl_ManStart( Abc_NtkName(pNtk) );
    // collect internal nodes to be added
    vNodes = Abc_NtkDfs( pNtk, 0 );
    // create combinational inputs
    // inlining of macro Abc_NtkForEachCi
    for  i = 0; (i < Abc_NtkCiNum(pNtk)) && (((pObj) = Abc_NtkCi(pNtk, i)), 1); i++  {
        Bbl_ManCreateObject( p, BBL_OBJ_CI, Abc_ObjId(pObj), 0, nil );
    }
    // create internal nodes 
    // inlining of macro Vec_PtrForEachEntry
    for  i = 0; (i < Vec_PtrSize(vNodes)) && (((pObj) = (Abc_Obj_t *)Vec_PtrEntry(vNodes, i)), 1); i++  {
        Bbl_ManCreateObject( p, BBL_OBJ_NODE, Abc_ObjId(pObj), Abc_ObjFaninNum(pObj), (char *)pObj.pData );
    }
    // create combinational outputs
    // inlining of macro Abc_NtkForEachCo
    for  i = 0; (i < Abc_NtkCoNum(pNtk)) && (((pObj) = Abc_NtkCo(pNtk, i)), 1); i++  {
        Bbl_ManCreateObject( p, BBL_OBJ_CO, Abc_ObjId(pObj), 1, nil );
    }
    // create fanin/fanout connections for internal nodes
    // inlining of macro Vec_PtrForEachEntry
    for  i = 0; (i < Vec_PtrSize(vNodes)) && (((pObj) = (Abc_Obj_t *)Vec_PtrEntry(vNodes, i)), 1); i++  {
        // inlining of macro Abc_ObjForEachFanin
        for  k = 0; (k < Abc_ObjFaninNum(pObj)) && (((pFanin) = Abc_ObjFanin(pObj, k)), 1); k++  {
            Bbl_ManAddFanin( p, Abc_ObjId(pObj), Abc_ObjId(pFanin) );
        }
    }
    // create fanin/fanout connections for combinational outputs
    // inlining of macro Abc_NtkForEachCo
    for  i = 0; (i < Abc_NtkCoNum(pNtk)) && (((pObj) = Abc_NtkCo(pNtk, i)), 1); i++  {
        // inlining of macro Abc_ObjForEachFanin
        for  k = 0; (k < Abc_ObjFaninNum(pObj)) && (((pFanin) = Abc_ObjFanin(pObj, k)), 1); k++  {
            Bbl_ManAddFanin( p, Abc_ObjId(pObj), Abc_ObjId(pFanin) );
        }
    }
    Vec_PtrFree( vNodes );
    // sanity check
    Bbl_ManCheck( p );
    return p;
}

/**Function*************************************************************

  Synopsis    [Writes the AIG in the binary format.]

  Description []
  
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Io_WriteBblif(pNtk *Abc_Ntk_t, pFileName string) {
    var p *Bbl_Man_t;
    p = Bbl_ManFromAbc( pNtk );
//Bbl_ManPrintStats( p );
//Bbl_ManDumpBlif( p, "test_bbl.blif" );
    Bbl_ManDumpBinaryBlif( p, pFileName );
    Bbl_ManStop( p );
}


////////////////////////////////////////////////////////////////////////
///                       END OF FILE                                ///
////////////////////////////////////////////////////////////////////////


//ABC_NAMESPACE_IMPL_END

