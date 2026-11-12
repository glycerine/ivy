package io

import (
	"fmt"
	"time"
)

/**CFile****************************************************************

  FileName    [ioReadBblif.c]

  SystemName  [ABC: Logic synthesis and verification system.]

  PackageName [Command processing package.]

  Synopsis    [Procedures to read AIG in the binary format.]

  Author      [Alan Mishchenko]
  
  Affiliation [UC Berkeley]

  Date        [Ver. 1.0. Started - June 20, 2005.]

  Revision    [$Id: ioReadBblif.c,v 1.00 2005/06/20 00:00:00 alanmi Exp $]

***********************************************************************/

//#include "ioAbc.h"
//#include "bool/dec/dec.h"
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

  Synopsis    [Constructs ABC network from the manager.]

  Description [The ABC network is started, as well as the array vCopy,
  which will map the new ID of each object in the BBLIF manager into
  the ponter ot the corresponding object in the ABC. For each internal
  node, determined by Bbl_ObjIsLut(), the SOP representation is created
  by retrieving the SOP representation of the BBLIF object. Finally,
  the objects are connected using fanin/fanout creation, and the dummy
  names are assigned because ABC requires each CI/CO to have a name.]
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Bbl_ManToAbc(p *Bbl_Man_t) *Abc_Ntk_t {
    var pNtk *Abc_Ntk_t;
    var pObjNew *Abc_Obj_t  = nil;
    var pObj, pFanin *Bbl_Obj_t;
    var vCopy *Vec_Ptr_t;
    // start the network
    pNtk = Abc_NtkAlloc( ABC_NTK_LOGIC, ABC_FUNC_SOP, 1 );
    pNtk.pName = Extra_UtilStrsav( Bbl_ManName(p) );
    // create objects
    vCopy = Vec_PtrStart( 1000 );
    // inlining of macro Bbl_ManForEachObj
    for  pObj = Bbl_ManObjFirst(p); (pObj != nil); pObj = Bbl_ManObjNext(p, pObj)  {
        if Bbl_ObjIsInput(pObj) {
            pObjNew = Abc_NtkCreatePi( pNtk );
        } else if Bbl_ObjIsOutput(pObj) {
            pObjNew = Abc_NtkCreatePo( pNtk );
        } else if Bbl_ObjIsLut(pObj) {
            pObjNew = Abc_NtkCreateNode( pNtk );
        } else { assert(false, "0"); }
        if Bbl_ObjIsLut(pObj) {
            pObjNew.pData = Abc_SopRegister( (Mem_Flex_t *)pNtk.pManFunc, Bbl_ObjSop(p, pObj) ); }
        Vec_PtrSetEntry( vCopy, Bbl_ObjId(pObj), pObjNew );
    }
    // connect objects
    // inlining of macro Bbl_ManForEachObj
    for  pObj = Bbl_ManObjFirst(p); (pObj != nil); pObj = Bbl_ManObjNext(p, pObj)  {
        // inlining of macro Bbl_ObjForEachFanin
        for  pFanin = Bbl_ObjFaninFirst(pObj); (pFanin != nil); pFanin = Bbl_ObjFaninNext(pObj, pFanin)  {
            Abc_ObjAddFanin( (Abc_Obj_t *)Vec_PtrEntry(vCopy, Bbl_ObjId(pObj)), (Abc_Obj_t *)Vec_PtrEntry(vCopy, Bbl_ObjId(pFanin)) );
        }
    }
    // finalize
    Vec_PtrFree( vCopy );
    Abc_NtkAddDummyPiNames( pNtk );
    Abc_NtkAddDummyPoNames( pNtk );
    if !Abc_NtkCheck( pNtk ) {
        fmt.Printf("Bbl_ManToAbc(): Network check has failed.\n"); }
    return pNtk;
}

/**Fnction*************************************************************

  Synopsis    [Collects internal nodes in the DFS order.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Bbl_ManDfs_rec(pObj *Bbl_Obj_t, vNodes *Vec_Ptr_t) {
    extern void Bbl_ObjMark( Bbl_Obj_t * p );
    extern int  Bbl_ObjIsMarked( Bbl_Obj_t * p );
    var pFanin *Bbl_Obj_t;
    if Bbl_ObjIsMarked(pObj) || Bbl_ObjIsInput(pObj) {
        return; }
    // inlining of macro Bbl_ObjForEachFanin
    for  pFanin = Bbl_ObjFaninFirst(pObj); (pFanin != nil); pFanin = Bbl_ObjFaninNext(pObj, pFanin)  {
        Bbl_ManDfs_rec( pFanin, vNodes );
    }
    assert(!Bbl_ObjIsMarked(pObj), "!Bbl_ObjIsMarked(pObj)"); // checks if acyclic
    Bbl_ObjMark( pObj );
    Vec_PtrPush( vNodes, pObj );
}

/**Fnction*************************************************************

  Synopsis    [Collects internal nodes in the DFS order.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Bbl_ManDfs(p *Bbl_Man_t) *Vec_Ptr_t {
    var vNodes *Vec_Ptr_t;
    var pObj *Bbl_Obj_t;
    vNodes = Vec_PtrAlloc( 1000 );
    // inlining of macro Bbl_ManForEachObj
    for  pObj = Bbl_ManObjFirst(p); (pObj != nil); pObj = Bbl_ManObjNext(p, pObj)  {
        if Bbl_ObjIsLut(pObj) {
            Bbl_ManDfs_rec( pObj, vNodes ); }
    }
    return vNodes;
}

/**Fnction*************************************************************

  Synopsis    [Constructs AIG in ABC from the manager.]

  Description [The ABC network is started, as well as the array vCopy,
  which will map the new ID of each object in the BBLIF manager into
  the ponter ot the corresponding AIG object in the ABC. For each internal
  node in a topological oder the AIG representation is created
  by factoring the SOP representation of the BBLIF object. Finally,
  the CO objects are created, and the dummy names are assigned because 
  ABC requires each CI/CO to have a name.]
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Bbl_ManToAig(p *Bbl_Man_t) *Abc_Ntk_t {
    extern int Bbl_ManFncSize( Bbl_Man_t * p );
    extern int Bbl_ObjFncHandle( Bbl_Obj_t * p );
    extern Abc_Obj_t * Dec_GraphToAig( Abc_Ntk_t * pNtk, Dec_Graph_t * pFForm, Vec_Ptr_t * vFaninAigs );
    var fVerbose int32;
    var pNtk *Abc_Ntk_t;
    var pObjNew *Abc_Obj_t  = nil;
    var pObj, pFanin *Bbl_Obj_t;
    var vCopy, vNodes, vFaninAigs *Vec_Ptr_t;
    var pFForms **Dec_Graph_t;
    var i int32;
    var clk abctime;
clk = Abc_Clock();
    // map SOP handles into factored forms
    // inlining of macro ABC_CALLOC
    pFForms = ((Dec_Graph_t * *) calloc((size_t)(Bbl_ManFncSize(p)), sizeof(Dec_Graph_t *)));
    // inlining of macro Bbl_ManForEachObj
    for  pObj = Bbl_ManObjFirst(p); (pObj != nil); pObj = Bbl_ManObjNext(p, pObj)  {
        if pFForms[Bbl_ObjFncHandle(pObj)] == nil {
            pFForms[Bbl_ObjFncHandle(pObj)] = Dec_Factor( Bbl_ObjSop(p, pObj) ); }
    }
if fVerbose != 0 {
// inlining of macro ABC_PRT
Abc_Print(1, "%s =", ("Fct"));
Abc_Print(1, "%9.2f sec\n", 1.0*(float64((Abc_Clock() - clk)))/(float64(int64(time.Second))));; }
    // start the network
    pNtk = Abc_NtkAlloc( ABC_NTK_STRASH, ABC_FUNC_AIG, 1 );
    pNtk.pName = Extra_UtilStrsav( Bbl_ManName(p) );
    vCopy = Vec_PtrStart( 1000 );
    // create CIs
    // inlining of macro Bbl_ManForEachObj
    for  pObj = Bbl_ManObjFirst(p); (pObj != nil); pObj = Bbl_ManObjNext(p, pObj)  {
        if !Bbl_ObjIsInput(pObj) {
            continue; }
        Vec_PtrSetEntry( vCopy, Bbl_ObjId(pObj), Abc_NtkCreatePi(pNtk) );
    }
clk = Abc_Clock();
    // create internal nodes
    vNodes = Bbl_ManDfs( p );
    vFaninAigs = Vec_PtrAlloc( 100 );
    // inlining of macro Vec_PtrForEachEntry
    for  i = 0; (i < Vec_PtrSize(vNodes)) && (((pObj) = (Bbl_Obj_t *)Vec_PtrEntry(vNodes, i)), 1); i++  {
        // collect fanin AIGs
        Vec_PtrClear( vFaninAigs );
        // inlining of macro Bbl_ObjForEachFanin
        for  pFanin = Bbl_ObjFaninFirst(pObj); (pFanin != nil); pFanin = Bbl_ObjFaninNext(pObj, pFanin)  {
            Vec_PtrPush( vFaninAigs, Vec_PtrEntry( vCopy, Bbl_ObjId(pFanin) ) );
        }
        // create the new node
        pObjNew = Dec_GraphToAig( pNtk, pFForms[Bbl_ObjFncHandle(pObj)], vFaninAigs );
        Vec_PtrSetEntry( vCopy, Bbl_ObjId(pObj), pObjNew );
    }
    Vec_PtrFree( vFaninAigs );
    Vec_PtrFree( vNodes );
if fVerbose != 0 {
// inlining of macro ABC_PRT
Abc_Print(1, "%s =", ("AIG"));
Abc_Print(1, "%9.2f sec\n", 1.0*(float64((Abc_Clock() - clk)))/(float64(int64(time.Second))));; }
    // create COs
    // inlining of macro Bbl_ManForEachObj
    for  pObj = Bbl_ManObjFirst(p); (pObj != nil); pObj = Bbl_ManObjNext(p, pObj)  {
        if !Bbl_ObjIsOutput(pObj) {
            continue; }
        pObjNew = (Abc_Obj_t *)Vec_PtrEntry( vCopy, Bbl_ObjId(Bbl_ObjFaninFirst(pObj)) );
        Abc_ObjAddFanin( Abc_NtkCreatePo(pNtk), pObjNew );
    }
    Abc_AigCleanup( (Abc_Aig_t *)pNtk.pManFunc );
    // clear factored forms
    for  i = Bbl_ManFncSize(p) - 1; i >= 0; i--  {
        if pFForms[i] {
            Dec_GraphFree( pFForms[i] ); }
    }
    // inlining of macro ABC_FREE
    ((pFForms) ? (free((char *) (pFForms)), (pFForms) = 0) : 0);
    // finalize
clk = Abc_Clock();
    Vec_PtrFree( vCopy );
    Abc_NtkAddDummyPiNames( pNtk );
    Abc_NtkAddDummyPoNames( pNtk );
if fVerbose != 0 {
// inlining of macro ABC_PRT
Abc_Print(1, "%s =", ("Nam"));
Abc_Print(1, "%9.2f sec\n", 1.0*(float64((Abc_Clock() - clk)))/(float64(int64(time.Second))));; }
//    if ( !Abc_NtkCheck( pNtk ) )
//        printf( "Bbl_ManToAig(): Network check has failed.\n" );
    return pNtk;
}

/**Fnction*************************************************************

  Synopsis    [Verifies equivalence for two combinational networks.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Bbl_ManVerify(pNtk1 *Abc_Ntk_t, pNtk2 *Abc_Ntk_t) {
    extern void Abc_NtkCecFraig( Abc_Ntk_t * pNtk1, Abc_Ntk_t * pNtk2, int nSeconds, int fVerbose );
    var pAig1, pAig2 *Abc_Ntk_t;
    pAig1 = Abc_NtkStrash( pNtk1, 0, 1, 0 );
    pAig2 = Abc_NtkStrash( pNtk2, 0, 1, 0 );
    Abc_NtkShortNames( pAig1 );
    Abc_NtkShortNames( pAig2 );
    Abc_NtkCecFraig( pAig1, pAig2, 0, 0 );
    Abc_NtkDelete( pAig1 );
    Abc_NtkDelete( pAig2 );
}

/**Fnction*************************************************************

  Synopsis    [Performs testing of the new manager.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Bbl_ManTest(pNtk *Abc_Ntk_t) {
    extern Bbl_Man_t * Bbl_ManFromAbc( Abc_Ntk_t * pNtk );

    var pNtkNew *Abc_Ntk_t;
    var p, pNew *Bbl_Man_t;
    var pFileName string  = "test.bblif";
    var clk, clk1, clk2, clk3, clk4, clk5 abctime;
clk = Abc_Clock();
    p = Bbl_ManFromAbc( pNtk );
    Bbl_ManPrintStats( p );
clk1 = Abc_Clock() - clk;
//Bbl_ManDumpBlif( p, "test_bbl.blif" );

    // write into file and back
clk = Abc_Clock();
    Bbl_ManDumpBinaryBlif( p, pFileName );
clk2 = Abc_Clock() - clk;

    // read from file
clk = Abc_Clock();
    pNew = Bbl_ManReadBinaryBlif( pFileName );
    Bbl_ManStop( p ); p = pNew;
clk3 = Abc_Clock() - clk;

    // generate ABC network
clk = Abc_Clock();
    pNtkNew = Bbl_ManToAig( p );
//    pNtkNew = Bbl_ManToAbc( p );
    Bbl_ManStop( p );
clk4 = Abc_Clock() - clk;

    // equivalence check
clk = Abc_Clock();
//    Bbl_ManVerify( pNtk, pNtkNew );
    Abc_NtkDelete( pNtkNew );
clk5 = Abc_Clock() - clk;

fmt.Printf("Runtime stats:\n");
// inlining of macro ABC_PRT
Abc_Print(1, "%s =", ("ABC to Man"));
Abc_Print(1, "%9.2f sec\n", 1.0*(float64((clk1)))/(float64(int64(time.Second))));;
// inlining of macro ABC_PRT
Abc_Print(1, "%s =", ("Writing   "));
Abc_Print(1, "%9.2f sec\n", 1.0*(float64((clk2)))/(float64(int64(time.Second))));;
// inlining of macro ABC_PRT
Abc_Print(1, "%s =", ("Reading   "));
Abc_Print(1, "%9.2f sec\n", 1.0*(float64((clk3)))/(float64(int64(time.Second))));;
// inlining of macro ABC_PRT
Abc_Print(1, "%s =", ("Man to ABC"));
Abc_Print(1, "%9.2f sec\n", 1.0*(float64((clk4)))/(float64(int64(time.Second))));;
// inlining of macro ABC_PRT
Abc_Print(1, "%s =", ("Verify    "));
Abc_Print(1, "%9.2f sec\n", 1.0*(float64((clk5)))/(float64(int64(time.Second))));;
}

/**Function*************************************************************

  Synopsis    [Reads the AIG in the binary format.]

  Description []
  
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Io_ReadBblif(pFileName string, fCheck int32) *Abc_Ntk_t {
    var p *Bbl_Man_t;
    var pNtkNew *Abc_Ntk_t;
    // read the file
    p = Bbl_ManReadBinaryBlif( pFileName );
    pNtkNew = Bbl_ManToAig( p );
    Bbl_ManStop( p );
    // check the result
    if (fCheck != 0) && !Abc_NtkCheckRead( pNtkNew ) {
        fmt.Printf("Io_ReadBaf: The network check has failed.\n");
        Abc_NtkDelete( pNtkNew );
        return nil;
    }
    return pNtkNew;
}

////////////////////////////////////////////////////////////////////////
///                       END OF FILE                                ///
////////////////////////////////////////////////////////////////////////


//ABC_NAMESPACE_IMPL_END

