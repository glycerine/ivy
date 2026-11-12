package abci

import (
	"fmt"
	"time"
)

/**CFile****************************************************************

  FileName    [abcDebug.c]

  SystemName  [ABC: Logic synthesis and verification system.]

  PackageName [Network and node package.]

  Synopsis    [Automated debugging procedures.]

  Author      [Alan Mishchenko]
  
  Affiliation [UC Berkeley]

  Date        [Ver. 1.0. Started - June 20, 2005.]

  Revision    [$Id: abcDebug.c,v 1.00 2005/06/20 00:00:00 alanmi Exp $]

***********************************************************************/

//#include "base/abc/abc.h"
//#include "base/io/ioAbc.h"

//ABC_NAMESPACE_IMPL_START


////////////////////////////////////////////////////////////////////////
///                        DECLARATIONS                              ///
////////////////////////////////////////////////////////////////////////

// static int Abc_NtkCountFaninsTotal( Abc_Ntk_t * pNtk );

// static Abc_Ntk_t * Abc_NtkAutoDebugModify( Abc_Ntk_t * pNtk, int ObjNum, int fConst1 );


////////////////////////////////////////////////////////////////////////
///                     FUNCTION DEFINITIONS                         ///
////////////////////////////////////////////////////////////////////////

/**Function*************************************************************

  Synopsis    [Takes a network and a procedure to test.]

  Description [The network demonstrates the bug in the procedure.
  Procedure should return 1 if the bug is demonstrated.]
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Abc_NtkAutoDebug(pNtk *Abc_Ntk_t, pFuncError func(*Abc_Ntk_t) int32) {
    var pNtkMod *Abc_Ntk_t;
    var pFileName string  = "bug_found.blif";
    var i int32;
    var nSteps int32;
    var nIter int32;
    var ModNum int32;
    var RandNum int32 = 1;
    var clk abctime;
    var clkTotal abctime = Abc_Clock();
    assert(Abc_NtkIsLogic(pNtk), "Abc_NtkIsLogic(pNtk)");
    srand( 0x123123 );
    // create internal copy of the network
    pNtk = Abc_NtkDup(pNtk);
    if !(*pFuncError)( pNtk ) {
        fmt.Printf("The original network does not cause the bug. Quitting.\n");
        Abc_NtkDelete( pNtk );
        return;
    }
    // perform incremental modifications
    for  nIter = 0; ; nIter++  {
        clk = Abc_Clock();
        // count how many ways of modifying the network exists
        nSteps = 2 * Abc_NtkCountFaninsTotal(pNtk);
        // try modifying the network as many times
        RandNum ^= rand();
        for  i = 0; i < nSteps; i++  {
            // get the shifted number of bug
            ModNum = (i + RandNum) % nSteps;
            // get the modified network
            pNtkMod = Abc_NtkAutoDebugModify( pNtk, ModNum/2, ModNum%2 );
            // write the network
            Io_WriteBlifLogic( pNtk, "bug_temp.blif", 1 );   
            // check if the bug is still there
            if (*pFuncError)( pNtkMod ) { // bug is still there
            
                Abc_NtkDelete( pNtk );
                pNtk = pNtkMod;
                break;
            
            } else { // no bug
                Abc_NtkDelete( pNtkMod ); }
        }
        fmt.Printf("Iter %6d : Latches = %6d. Nodes = %6d. Steps = %6d. Error step = %3d.  ", nIter, Abc_NtkLatchNum(pNtk), Abc_NtkNodeNum(pNtk), nSteps, i);
        // inlining of macro ABC_PRT
        Abc_Print(1, "%s =", ("Time"));
Abc_Print(1, "%9.2f sec\n", 1.0*(float64((Abc_Clock() - clk)))/(float64(int64(time.Second))));;
        if i == nSteps { // could not modify it while preserving the bug
            break; }
    }
    // write out the final network
    Io_WriteBlifLogic( pNtk, pFileName, 1 );
    fmt.Printf("Final network written into file \"%s\". ", pFileName);
    // inlining of macro ABC_PRT
    Abc_Print(1, "%s =", ("Total time"));
Abc_Print(1, "%9.2f sec\n", 1.0*(float64((Abc_Clock() - clkTotal)))/(float64(int64(time.Second))));;
    Abc_NtkDelete( pNtk );
}

/**Function*************************************************************

  Synopsis    [Counts the total number of fanins.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Abc_NtkCountFaninsTotal(pNtk *Abc_Ntk_t) int32 {
    var pObj, pFanin *Abc_Obj_t;
    var i, k, Counter int32;
    // inlining of macro Abc_NtkForEachObj
    for  i = 0; (i < Vec_PtrSize((pNtk).vObjs)) && (((pObj) = Abc_NtkObj(pNtk, i)), 1); i++  {            if (pObj) == nil {} else {
        // inlining of macro Abc_ObjForEachFanin
        for  k = 0; (k < Abc_ObjFaninNum(pObj)) && (((pFanin) = Abc_ObjFanin(pObj, k)), 1); k++  {
            if !Abc_ObjIsNode(pObj) && !Abc_ObjIsPo(pObj) {
                continue; }
            if Abc_ObjIsPo(pObj) && Abc_NtkPoNum(pNtk) == 1 {
                continue; }
            if Abc_ObjIsNode(pObj) && Abc_NodeIsConst(pFanin) {
                continue; }
            Counter++;
        } }
    }
    return Counter;
}

/**Function*************************************************************

  Synopsis    [Returns the node and fanin to be modified.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Abc_NtkFindGivenFanin(pNtk *Abc_Ntk_t, Step int32, ppObj **Abc_Obj_t, ppFanin **Abc_Obj_t) int32 {
    var pObj, pFanin *Abc_Obj_t;
    var i, k, Counter int32;
    // inlining of macro Abc_NtkForEachObj
    for  i = 0; (i < Vec_PtrSize((pNtk).vObjs)) && (((pObj) = Abc_NtkObj(pNtk, i)), 1); i++  {            if (pObj) == nil {} else {
        // inlining of macro Abc_ObjForEachFanin
        for  k = 0; (k < Abc_ObjFaninNum(pObj)) && (((pFanin) = Abc_ObjFanin(pObj, k)), 1); k++  {
            if !Abc_ObjIsNode(pObj) && !Abc_ObjIsPo(pObj) {
                continue; }
            if Abc_ObjIsPo(pObj) && Abc_NtkPoNum(pNtk) == 1 {
                continue; }
            if Abc_ObjIsNode(pObj) && Abc_NodeIsConst(pFanin) {
                continue; }
            if Counter++ == Step {
                *ppObj   = pObj;
                *ppFanin = pFanin;
                return 1;
            }
        } }
    }
    return 0;
}

/**Function*************************************************************

  Synopsis    [Perform modification with the given number.]

  Description [Modification consists of replacing the node by a constant.]
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Abc_NtkAutoDebugModify(pNtkInit *Abc_Ntk_t, Step int32, fConst1 int32) *Abc_Ntk_t {
    extern void Abc_NtkCycleInitStateSop( Abc_Ntk_t * pNtk, int nFrames, int fVerbose );
    var pNtk *Abc_Ntk_t;
    var pObj, pFanin, pConst *Abc_Obj_t;
    // copy the network
    pNtk = Abc_NtkDup( pNtkInit );
    assert(Abc_NtkNodeNum(pNtk) == Abc_NtkNodeNum(pNtkInit), "Abc_NtkNodeNum(pNtk) == Abc_NtkNodeNum(pNtkInit)");
    // find the object number
    Abc_NtkFindGivenFanin( pNtk, Step, &pObj, &pFanin );
    // consider special case 
    if Abc_ObjIsPo(pObj) && Abc_NodeIsConst(pFanin) {
        Abc_NtkDeleteAll_rec( pObj );
        return pNtk;
    }
    // plug in a constant node
    if fConst1 != 0 {
        pConst = Abc_NtkCreateNodeConst1(pNtk);
    } else {
        pConst = Abc_NtkCreateNodeConst0(pNtk);
    }
    Abc_ObjTransferFanout( pFanin, pConst );
    Abc_NtkDeleteAll_rec( pFanin );

    Abc_NtkSweep( pNtk, 0 );
    Abc_NtkCleanupSeq( pNtk, 0, 0, 0 );
    Abc_NtkToSop( pNtk, -1, ABC_INFINITY );
    Abc_NtkCycleInitStateSop( pNtk, 50, 0 );
    return pNtk;
}

////////////////////////////////////////////////////////////////////////
///                       END OF FILE                                ///
////////////////////////////////////////////////////////////////////////


//ABC_NAMESPACE_IMPL_END

