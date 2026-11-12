package io

import "fmt"

import "os"

/**CFile****************************************************************

  FileName    [ioWriteGml.c]

  SystemName  [ABC: Logic synthesis and verification system.]

  PackageName [Command processing package.]

  Synopsis    [Procedures to write the graph structure of AIG in GML.]

  Author      [Alan Mishchenko]
  
  Affiliation [UC Berkeley]

  Date        [Ver. 1.0. Started - June 20, 2005.]

  Revision    [$Id: ioWriteGml.c,v 1.00 2005/06/20 00:00:00 alanmi Exp $]

***********************************************************************/

//#include "ioAbc.h"

//ABC_NAMESPACE_IMPL_START


////////////////////////////////////////////////////////////////////////
///                        DECLARATIONS                              ///
////////////////////////////////////////////////////////////////////////

////////////////////////////////////////////////////////////////////////
///                     FUNCTION DEFINITIONS                         ///
////////////////////////////////////////////////////////////////////////

/**Function*************************************************************

  Synopsis    [Writes the graph structure of AIG in GML.]

  Description [Useful for graph visualization using tools such as yEd: 
  http://www.yworks.com/]
  
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Io_WriteGml(pNtk *Abc_Ntk_t, pFileName string) {
    var pFile *os.File;
    var pObj, pFanin *Abc_Obj_t;
    var i, k int32;

    assert(Abc_NtkIsStrash(pNtk) || Abc_NtkIsLogic(pNtk), "Abc_NtkIsStrash(pNtk) || Abc_NtkIsLogic(pNtk)");

    // start the output stream
    pFile = os.OpenFile(pFileName, os.O_WRONLY | os.O_CREATE | os.O_TRUNC, 0666);
    if pFile == nil {
        fmt.Fprintf(os.Stdout, "Io_WriteGml(): Cannot open the output file \"%s\".\n", pFileName);
        return;
    }
    fmt.Fprintf(pFile, "# GML for \"%s\" written by ABC on %s\n", pNtk.pName, Extra_TimeStamp());
    fmt.Fprintf(pFile, "graph [\n");

    // output the POs
    fmt.Fprintf(pFile, "\n");
    // inlining of macro Abc_NtkForEachPo
    for  i = 0; (i < Abc_NtkPoNum(pNtk)) && (((pObj) = Abc_NtkPo(pNtk, i)), 1); i++  {
        fmt.Fprintf(pFile, "    node [ id %5d label \"%s\"\n", pObj.Id, Abc_ObjName(pObj));
        fmt.Fprintf(pFile, "        graphics [ type \"triangle\" fill \"#00FFFF\" ]\n");   // blue
        fmt.Fprintf(pFile, "    ]\n");
    }
    // output the PIs
    fmt.Fprintf(pFile, "\n");
    // inlining of macro Abc_NtkForEachPi
    for  i = 0; (i < Abc_NtkPiNum(pNtk)) && (((pObj) = Abc_NtkPi(pNtk, i)), 1); i++  {
        fmt.Fprintf(pFile, "    node [ id %5d label \"%s\"\n", pObj.Id, Abc_ObjName(pObj));
        fmt.Fprintf(pFile, "        graphics [ type \"triangle\" fill \"#00FF00\" ]\n");   // green
        fmt.Fprintf(pFile, "    ]\n");
    }
    // output the latches
    fmt.Fprintf(pFile, "\n");
    // inlining of macro Abc_NtkForEachLatch
    for  i = 0; (i < Vec_PtrSize((pNtk).vBoxes)) && (((pObj) = Abc_NtkBox(pNtk, i)), 1); i++  {           if !Abc_ObjIsLatch(pObj) {} else {
        fmt.Fprintf(pFile, "    node [ id %5d label \"%s\"\n", pObj.Id, Abc_ObjName(pObj));
        fmt.Fprintf(pFile, "        graphics [ type \"rectangle\" fill \"#FF0000\" ]\n");   // red
        fmt.Fprintf(pFile, "    ]\n");
    }
    }
    // output the nodes
    fmt.Fprintf(pFile, "\n");
    // inlining of macro Abc_NtkForEachNode
    for  i = 0; (i < Vec_PtrSize((pNtk).vObjs)) && (((pObj) = Abc_NtkObj(pNtk, i)), 1); i++  {           if (pObj) == nil || !Abc_ObjIsNode(pObj) {} else {
        fmt.Fprintf(pFile, "    node [ id %5d label \"%s\"\n", pObj.Id, Abc_ObjName(pObj));
        fmt.Fprintf(pFile, "        graphics [ type \"ellipse\" fill \"#CCCCFF\" ]\n");     // grey
        fmt.Fprintf(pFile, "    ]\n");
    }
    }

    // output the edges
    fmt.Fprintf(pFile, "\n");
    // inlining of macro Abc_NtkForEachObj
    for  i = 0; (i < Vec_PtrSize((pNtk).vObjs)) && (((pObj) = Abc_NtkObj(pNtk, i)), 1); i++  {            if (pObj) == nil {} else {
        // inlining of macro Abc_ObjForEachFanin
        for  k = 0; (k < Abc_ObjFaninNum(pObj)) && (((pFanin) = Abc_ObjFanin(pObj, k)), 1); k++  {
            fmt.Fprintf(pFile, "    edge [ source %5d   target %5d\n", pObj.Id, pFanin.Id);
            fmt.Fprintf(pFile, "        graphics [ type \"line\" arrow \"first\" ]\n");
            fmt.Fprintf(pFile, "    ]\n");
        }
    }
    }

    fmt.Fprintf(pFile, "]\n");
    fmt.Fprintf(pFile, "\n");
    _ = (pFile).Close();
}


////////////////////////////////////////////////////////////////////////
///                       END OF FILE                                ///
////////////////////////////////////////////////////////////////////////


//ABC_NAMESPACE_IMPL_END

