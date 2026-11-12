package io

import "fmt"

import "os"

/**CFile****************************************************************

  FileName    [ioWriteDot.c]

  SystemName  [ABC: Logic synthesis and verification system.]

  PackageName [Command processing package.]

  Synopsis    [Procedures to write the graph structure of AIG in DOT.]

  Author      [Alan Mishchenko]
  
  Affiliation [UC Berkeley]

  Date        [Ver. 1.0. Started - June 20, 2005.]

  Revision    [$Id: ioWriteDot.c,v 1.00 2005/06/20 00:00:00 alanmi Exp $]

***********************************************************************/

//#include "ioAbc.h"
//#include "base/main/main.h"
//#include "map/mio/mio.h"

//ABC_NAMESPACE_IMPL_START


////////////////////////////////////////////////////////////////////////
///                        DECLARATIONS                              ///
////////////////////////////////////////////////////////////////////////

// static char * Abc_NtkPrintSop( char * pSop );

// static int    Abc_NtkCountLogicNodes( Vec_Ptr_t * vNodes );


////////////////////////////////////////////////////////////////////////
///                     FUNCTION DEFINITIONS                         ///
////////////////////////////////////////////////////////////////////////

/**Function*************************************************************

  Synopsis    [Writes the graph structure of network for DOT.]

  Description [Useful for graph visualization using tools such as GraphViz: 
  http://www.graphviz.org/]
  
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Io_WriteDot(pNtk *Abc_Ntk_t, FileName string) {
    var vNodes *Vec_Ptr_t;
    vNodes = Abc_NtkCollectObjects( pNtk );
    Io_WriteDotNtk( pNtk, vNodes, nil, FileName, 0, 0 );
    Vec_PtrFree( vNodes );
}

/**Function*************************************************************

  Synopsis    [Writes the graph structure of network for DOT.]

  Description [Useful for graph visualization using tools such as GraphViz: 
  http://www.graphviz.org/]
  
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Io_WriteDotNtk(pNtk *Abc_Ntk_t, vNodes *Vec_Ptr_t, vNodesShow *Vec_Ptr_t, pFileName string, fGateNames int32, fUseReverse int32) {
    var pFile *os.File;
    var pNode, pFanin *Abc_Obj_t;
    var pSopString string;
    var LevelMin, LevelMax, fHasCos, Level, i, k, fHasBdds, fCompl, Prev int32;
    var Limit int32  = 500;

    assert(Abc_NtkIsStrash(pNtk) || Abc_NtkIsLogic(pNtk), "Abc_NtkIsStrash(pNtk) || Abc_NtkIsLogic(pNtk)");

    if vNodes.nSize < 1 {
        fmt.Printf("The set has no nodes. DOT file is not written.\n");
        return;
    }

    if vNodes.nSize > Limit {
        fmt.Printf("The set has more than %d nodes. DOT file is not written.\n", Limit);
        return;
    }

    // start the stream
    if (pFile = os.OpenFile(pFileName, os.O_WRONLY | os.O_CREATE | os.O_TRUNC, 0666)) == nil {
        fmt.Fprintf(os.Stdout, "Cannot open the intermediate file \"%s\".\n", pFileName);
        return;
    }

    // transform logic functions from BDD to SOP
    if fHasBdds = Abc_NtkIsBddLogic(pNtk) {
        if !Abc_NtkBddToSop(pNtk, -1, ABC_INFINITY) {
            fmt.Printf("Io_WriteDotNtk(): Converting to SOPs has failed.\n");
            return;
        }
    }

    // mark the nodes from the set
    // inlining of macro Vec_PtrForEachEntry
    for  i = 0; (i < Vec_PtrSize(vNodes)) && (((pNode) = (Abc_Obj_t *)Vec_PtrEntry(vNodes, i)), 1); i++  {
        pNode.fMarkC = 1;
    }
    if vNodesShow != nil {
        // inlining of macro Vec_PtrForEachEntry
        for  i = 0; (i < Vec_PtrSize(vNodesShow)) && (((pNode) = (Abc_Obj_t *)Vec_PtrEntry(vNodesShow, i)), 1); i++  {
            pNode.fMarkB = 1;
        } }

    // get the levels of nodes
    LevelMax = Abc_NtkLevel( pNtk );
    if fUseReverse != 0 {
        LevelMin = Abc_NtkLevelReverse( pNtk );
        assert(LevelMax == LevelMin, "LevelMax == LevelMin");
        // inlining of macro Vec_PtrForEachEntry
        for  i = 0; (i < Vec_PtrSize(vNodes)) && (((pNode) = (Abc_Obj_t *)Vec_PtrEntry(vNodes, i)), 1); i++  {
            if Abc_ObjIsNode(pNode) {
                pNode.Level = LevelMax - pNode.Level + 1; }
        }
    }

    // find the largest and the smallest levels
    LevelMin = 10000;
    LevelMax = -1;
    fHasCos  = 0;
    // inlining of macro Vec_PtrForEachEntry
    for  i = 0; (i < Vec_PtrSize(vNodes)) && (((pNode) = (Abc_Obj_t *)Vec_PtrEntry(vNodes, i)), 1); i++  {
        if Abc_ObjIsCo(pNode) {
            fHasCos = 1;
            continue;
        }
        if LevelMin > int32(pNode.Level) {
            LevelMin = pNode.Level; }
        if LevelMax < int32(pNode.Level) {
            LevelMax = pNode.Level; }
    }

    // set the level of the CO nodes
    if fHasCos {
        LevelMax++;
        // inlining of macro Vec_PtrForEachEntry
        for  i = 0; (i < Vec_PtrSize(vNodes)) && (((pNode) = (Abc_Obj_t *)Vec_PtrEntry(vNodes, i)), 1); i++  {
            if Abc_ObjIsCo(pNode) {
                pNode.Level = LevelMax; }
        }
    }

    // write the DOT header
    fmt.Fprintf(pFile, "# %s\n", "Network structure generated by ABC");
    fmt.Fprintf(pFile, "\n");
    fmt.Fprintf(pFile, "digraph network {\n");
    fmt.Fprintf(pFile, "size = \"7.5,10\";\n");
//    fprintf( pFile, "size = \"10,8.5\";\n" );
//    fprintf( pFile, "size = \"14,11\";\n" );
//    fprintf( pFile, "page = \"8,11\";\n" );
//  fprintf( pFile, "ranksep = 0.5;\n" );
//  fprintf( pFile, "nodesep = 0.5;\n" );
    fmt.Fprintf(pFile, "center = true;\n");
//    fprintf( pFile, "orientation = landscape;\n" );
//  fprintf( pFile, "edge [fontsize = 10];\n" );
//  fprintf( pFile, "edge [dir = none];\n" );
    fmt.Fprintf(pFile, "edge [dir = back];\n");
    fmt.Fprintf(pFile, "\n");

    // labels on the left of the picture
    fmt.Fprintf(pFile, "{\n");
    fmt.Fprintf(pFile, "  node [shape = plaintext];\n");
    fmt.Fprintf(pFile, "  edge [style = invis];\n");
    fmt.Fprintf(pFile, "  LevelTitle1 [label=\"\"];\n");
    fmt.Fprintf(pFile, "  LevelTitle2 [label=\"\"];\n");
    // generate node names with labels
    for  Level = LevelMax; Level >= LevelMin; Level--  {
        // the visible node name
        fmt.Fprintf(pFile, "  Level%d", Level);
        fmt.Fprintf(pFile, " [label = ");
        // label name
        fmt.Fprintf(pFile, "\"");
        fmt.Fprintf(pFile, "\"");
        fmt.Fprintf(pFile, "];\n");
    }

    // genetate the sequence of visible/invisible nodes to mark levels
    fmt.Fprintf(pFile, "  LevelTitle1 ->  LevelTitle2 ->");
    for  Level = LevelMax; Level >= LevelMin; Level--  {
        // the visible node name
        fmt.Fprintf(pFile, "  Level%d", Level);
        // the connector
        if Level != LevelMin {
            fmt.Fprintf(pFile, " ->");
        } else {
            fmt.Fprintf(pFile, ";"); }
    }
    fmt.Fprintf(pFile, "\n");
    fmt.Fprintf(pFile, "}");
    fmt.Fprintf(pFile, "\n");
    fmt.Fprintf(pFile, "\n");

    // generate title box on top
    fmt.Fprintf(pFile, "{\n");
    fmt.Fprintf(pFile, "  rank = same;\n");
    fmt.Fprintf(pFile, "  LevelTitle1;\n");
    fmt.Fprintf(pFile, "  title1 [shape=plaintext,\n");
    fmt.Fprintf(pFile, "          fontsize=20,\n");
    fmt.Fprintf(pFile, "          fontname = \"Times-Roman\",\n");
    fmt.Fprintf(pFile, "          label=\"");
    fmt.Fprintf(pFile, "%s", "Network structure visualized by ABC");
    fmt.Fprintf(pFile, "\\n");
    fmt.Fprintf(pFile, "Benchmark \\\"%s\\\". ", pNtk.pName);
    fmt.Fprintf(pFile, "Time was %s. ", Extra_TimeStamp());
    fmt.Fprintf(pFile, "\"\n");
    fmt.Fprintf(pFile, "         ];\n");
    fmt.Fprintf(pFile, "}");
    fmt.Fprintf(pFile, "\n");
    fmt.Fprintf(pFile, "\n");

    // generate statistics box
    fmt.Fprintf(pFile, "{\n");
    fmt.Fprintf(pFile, "  rank = same;\n");
    fmt.Fprintf(pFile, "  LevelTitle2;\n");
    fmt.Fprintf(pFile, "  title2 [shape=plaintext,\n");
    fmt.Fprintf(pFile, "          fontsize=18,\n");
    fmt.Fprintf(pFile, "          fontname = \"Times-Roman\",\n");
    fmt.Fprintf(pFile, "          label=\"");
    if Abc_NtkObjNum(pNtk) == Vec_PtrSize(vNodes) {
        fmt.Fprintf(pFile, "The network contains %d logic nodes and %d latches.", Abc_NtkNodeNum(pNtk), Abc_NtkLatchNum(pNtk));
    } else {
        fmt.Fprintf(pFile, "The set contains %d logic nodes and spans %d levels.", Abc_NtkCountLogicNodes(vNodes), LevelMax - LevelMin + 1); }
    fmt.Fprintf(pFile, "\\n");
    fmt.Fprintf(pFile, "\"\n");
    fmt.Fprintf(pFile, "         ];\n");
    fmt.Fprintf(pFile, "}");
    fmt.Fprintf(pFile, "\n");
    fmt.Fprintf(pFile, "\n");

    // generate the POs
    if fHasCos {
        fmt.Fprintf(pFile, "{\n");
        fmt.Fprintf(pFile, "  rank = same;\n");
        // the labeling node of this level
        fmt.Fprintf(pFile, "  Level%d;\n", LevelMax);
        // generate the PO nodes
        // inlining of macro Vec_PtrForEachEntry
        for  i = 0; (i < Vec_PtrSize(vNodes)) && (((pNode) = (Abc_Obj_t *)Vec_PtrEntry(vNodes, i)), 1); i++  {
            if !Abc_ObjIsCo(pNode) {
                continue; }
            fmt.Fprintf(pFile, "  Node%d [label = \"%s%s\"", pNode.Id, (Abc_ObjIsBi(pNode)? Abc_ObjName(Abc_ObjFanout0(pNode)):Abc_ObjName(pNode)), (Abc_ObjIsBi(pNode)? "_in":""));
            fmt.Fprintf(pFile, ", shape = %s", (Abc_ObjIsBi(pNode)? "box":"invtriangle"));
            if pNode.fMarkB {
                fmt.Fprintf(pFile, ", style = filled"); }
            fmt.Fprintf(pFile, ", color = coral, fillcolor = coral");
            fmt.Fprintf(pFile, "];\n");
        }
        fmt.Fprintf(pFile, "}");
        fmt.Fprintf(pFile, "\n");
        fmt.Fprintf(pFile, "\n");
    }

    // generate nodes of each rank
    for  Level = LevelMax - fHasCos; Level >= LevelMin && Level > 0; Level--  {
        fmt.Fprintf(pFile, "{\n");
        fmt.Fprintf(pFile, "  rank = same;\n");
        // the labeling node of this level
        fmt.Fprintf(pFile, "  Level%d;\n", Level);
        // inlining of macro Vec_PtrForEachEntry
        for  i = 0; (i < Vec_PtrSize(vNodes)) && (((pNode) = (Abc_Obj_t *)Vec_PtrEntry(vNodes, i)), 1); i++  {
            if int32(pNode.Level) != Level {
                continue; }
            if Abc_ObjFaninNum(pNode) == 0 {
                continue; }

/*
            int SuppSize;
            Vec_Ptr_t * vSupp;
            if ( (int)pNode->Level != Level )
                continue;
            if ( Abc_ObjFaninNum(pNode) == 0 )
                continue;
            vSupp = Abc_NtkNodeSupport( pNtk, &pNode, 1 );
            SuppSize = Vec_PtrSize( vSupp );
            Vec_PtrFree( vSupp ); 
*/

//            fprintf( pFile, "  Node%d [label = \"%d\"", pNode->Id, pNode->Id );
            if Abc_NtkIsStrash(pNtk) {
                pSopString = "";
            } else if Abc_NtkHasMapping(pNtk) && (fGateNames != 0) {
                pSopString = Mio_GateReadName((Mio_Gate_t *)pNode.pData);
            } else if Abc_NtkHasMapping(pNtk) {
                pSopString = Abc_NtkPrintSop(Mio_GateReadSop((Mio_Gate_t *)pNode.pData));
            } else {
                pSopString = Abc_NtkPrintSop((char *)pNode.pData); }
            fmt.Fprintf(pFile, "  Node%d [label = \"%d\\n%s\"", pNode.Id, pNode.Id, pSopString);
//            fprintf( pFile, "  Node%d [label = \"%d\\n%s\"", pNode->Id, 
//                SuppSize, 
//                pSopString );

            fmt.Fprintf(pFile, ", shape = ellipse");
            if pNode.fMarkB {
                fmt.Fprintf(pFile, ", style = filled"); }
            fmt.Fprintf(pFile, "];\n");
        }
        fmt.Fprintf(pFile, "}");
        fmt.Fprintf(pFile, "\n");
        fmt.Fprintf(pFile, "\n");
    }

    // generate the PI nodes if any
    if LevelMin == 0 {
        fmt.Fprintf(pFile, "{\n");
        fmt.Fprintf(pFile, "  rank = same;\n");
        // the labeling node of this level
        fmt.Fprintf(pFile, "  Level%d;\n", LevelMin);
        // generate the PO nodes
        // inlining of macro Vec_PtrForEachEntry
        for  i = 0; (i < Vec_PtrSize(vNodes)) && (((pNode) = (Abc_Obj_t *)Vec_PtrEntry(vNodes, i)), 1); i++  {
            if !Abc_ObjIsCi(pNode) {
                // check if the costant node is present
                if Abc_ObjFaninNum(pNode) == 0 && Abc_ObjFanoutNum(pNode) > 0 {
                    fmt.Fprintf(pFile, "  Node%d [label = \"Const%d\"", pNode.Id, Abc_NtkIsStrash(pNode.pNtk) || Abc_NodeIsConst1(pNode));
                    fmt.Fprintf(pFile, ", shape = ellipse");
                    if pNode.fMarkB {
                        fmt.Fprintf(pFile, ", style = filled"); }
                    fmt.Fprintf(pFile, ", color = coral, fillcolor = coral");
                    fmt.Fprintf(pFile, "];\n");
                }
                continue;
            }
            fmt.Fprintf(pFile, "  Node%d [label = \"%s\"", pNode.Id, (Abc_ObjIsBo(pNode)? Abc_ObjName(Abc_ObjFanin0(pNode)):Abc_ObjName(pNode)));
            fmt.Fprintf(pFile, ", shape = %s", (Abc_ObjIsBo(pNode)? "box":"triangle"));
            if pNode.fMarkB {
                fmt.Fprintf(pFile, ", style = filled"); }
            fmt.Fprintf(pFile, ", color = coral, fillcolor = coral");
            fmt.Fprintf(pFile, "];\n");
        }
        fmt.Fprintf(pFile, "}");
        fmt.Fprintf(pFile, "\n");
        fmt.Fprintf(pFile, "\n");
    }

    // generate invisible edges from the square down
    fmt.Fprintf(pFile, "title1 -> title2 [style = invis];\n");
    // inlining of macro Vec_PtrForEachEntry
    for  i = 0; (i < Vec_PtrSize(vNodes)) && (((pNode) = (Abc_Obj_t *)Vec_PtrEntry(vNodes, i)), 1); i++  {
        if int32(pNode.Level) != LevelMax {
            continue; }
        fmt.Fprintf(pFile, "title2 -> Node%d [style = invis];\n", pNode.Id);
    }
    // generate invisible edges among the COs
    Prev = -1;
    // inlining of macro Vec_PtrForEachEntry
    for  i = 0; (i < Vec_PtrSize(vNodes)) && (((pNode) = (Abc_Obj_t *)Vec_PtrEntry(vNodes, i)), 1); i++  {
        if int32(pNode.Level) != LevelMax {
            continue; }
        if !Abc_ObjIsPo(pNode) {
            continue; }
        if Prev >= 0 {
            fmt.Fprintf(pFile, "Node%d -> Node%d [style = invis];\n", Prev, pNode.Id); }
        Prev = pNode.Id;
    }

    // generate edges
    // inlining of macro Vec_PtrForEachEntry
    for  i = 0; (i < Vec_PtrSize(vNodes)) && (((pNode) = (Abc_Obj_t *)Vec_PtrEntry(vNodes, i)), 1); i++  {
        if Abc_ObjIsLatch(pNode) {
            continue; }
        // inlining of macro Abc_ObjForEachFanin
        for  k = 0; (k < Abc_ObjFaninNum(pNode)) && (((pFanin) = Abc_ObjFanin(pNode, k)), 1); k++  {
            if Abc_ObjIsLatch(pFanin) {
                continue; }
            fCompl = 0;
            if Abc_NtkIsStrash(pNtk) {
                fCompl = Abc_ObjFaninC(pNode, k); }
            // generate the edge from this node to the next
            fmt.Fprintf(pFile, "Node%d", pNode.Id);
            fmt.Fprintf(pFile, " -> ");
            fmt.Fprintf(pFile, "Node%d", pFanin.Id);
            fmt.Fprintf(pFile, " [style = %s", fCompl? "dotted" : "solid");
//            fprintf( pFile, ", label = \"%c\"", 'a' + k );
            fmt.Fprintf(pFile, "]");
            fmt.Fprintf(pFile, ";\n");
        }
    }

    fmt.Fprintf(pFile, "}");
    fmt.Fprintf(pFile, "\n");
    fmt.Fprintf(pFile, "\n");
    _ = (pFile).Close();

    // unmark the nodes from the set
    // inlining of macro Vec_PtrForEachEntry
    for  i = 0; (i < Vec_PtrSize(vNodes)) && (((pNode) = (Abc_Obj_t *)Vec_PtrEntry(vNodes, i)), 1); i++  {
        pNode.fMarkC = 0;
    }
    if vNodesShow != nil {
        // inlining of macro Vec_PtrForEachEntry
        for  i = 0; (i < Vec_PtrSize(vNodesShow)) && (((pNode) = (Abc_Obj_t *)Vec_PtrEntry(vNodesShow, i)), 1); i++  {
            pNode.fMarkB = 0;
        } }

    // convert the network back into BDDs if this is how it was
    if fHasBdds {
        Abc_NtkSopToBdd(pNtk); }
}


/**Function*************************************************************

  Synopsis    [Writes the graph structure of network for DOT.]

  Description [Useful for graph visualization using tools such as GraphViz: 
  http://www.graphviz.org/]
  
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Io_WriteDotSeq(pNtk *Abc_Ntk_t, vNodes *Vec_Ptr_t, vNodesShow *Vec_Ptr_t, pFileName string, fGateNames int32, fUseReverse int32) {
    var pFile *os.File;
    var pNode, pFanin *Abc_Obj_t;
    var pSopString string;
    var LevelMin, LevelMax, fHasCos, Level, i, k, fHasBdds, fCompl, Prev int32;
    var Limit int32  = 300;

    assert(Abc_NtkIsStrash(pNtk) || Abc_NtkIsLogic(pNtk), "Abc_NtkIsStrash(pNtk) || Abc_NtkIsLogic(pNtk)");

    if vNodes.nSize < 1 {
        fmt.Printf("The set has no nodes. DOT file is not written.\n");
        return;
    }

    if vNodes.nSize > Limit {
        fmt.Printf("The set has more than %d nodes. DOT file is not written.\n", Limit);
        return;
    }

    // start the stream
    if (pFile = os.OpenFile(pFileName, os.O_WRONLY | os.O_CREATE | os.O_TRUNC, 0666)) == nil {
        fmt.Fprintf(os.Stdout, "Cannot open the intermediate file \"%s\".\n", pFileName);
        return;
    }

    // transform logic functions from BDD to SOP
    if fHasBdds = Abc_NtkIsBddLogic(pNtk) {
        if !Abc_NtkBddToSop(pNtk, -1, ABC_INFINITY) {
            fmt.Printf("Io_WriteDotNtk(): Converting to SOPs has failed.\n");
            return;
        }
    }

    // mark the nodes from the set
    // inlining of macro Vec_PtrForEachEntry
    for  i = 0; (i < Vec_PtrSize(vNodes)) && (((pNode) = (Abc_Obj_t *)Vec_PtrEntry(vNodes, i)), 1); i++  {
        pNode.fMarkC = 1;
    }
    if vNodesShow != nil {
        // inlining of macro Vec_PtrForEachEntry
        for  i = 0; (i < Vec_PtrSize(vNodesShow)) && (((pNode) = (Abc_Obj_t *)Vec_PtrEntry(vNodesShow, i)), 1); i++  {
            pNode.fMarkB = 1;
        } }

    // get the levels of nodes
    LevelMax = Abc_NtkLevel( pNtk );
    if fUseReverse != 0 {
        LevelMin = Abc_NtkLevelReverse( pNtk );
        assert(LevelMax == LevelMin, "LevelMax == LevelMin");
        // inlining of macro Vec_PtrForEachEntry
        for  i = 0; (i < Vec_PtrSize(vNodes)) && (((pNode) = (Abc_Obj_t *)Vec_PtrEntry(vNodes, i)), 1); i++  {
            if Abc_ObjIsNode(pNode) {
                pNode.Level = LevelMax - pNode.Level + 1; }
        }
    }

    // find the largest and the smallest levels
    LevelMin = 10000;
    LevelMax = -1;
    fHasCos  = 0;
    // inlining of macro Vec_PtrForEachEntry
    for  i = 0; (i < Vec_PtrSize(vNodes)) && (((pNode) = (Abc_Obj_t *)Vec_PtrEntry(vNodes, i)), 1); i++  {
        if Abc_ObjIsCo(pNode) {
            fHasCos = 1;
            continue;
        }
        if LevelMin > int32(pNode.Level) {
            LevelMin = pNode.Level; }
        if LevelMax < int32(pNode.Level) {
            LevelMax = pNode.Level; }
    }

    // set the level of the CO nodes
    if fHasCos {
        LevelMax++;
        // inlining of macro Vec_PtrForEachEntry
        for  i = 0; (i < Vec_PtrSize(vNodes)) && (((pNode) = (Abc_Obj_t *)Vec_PtrEntry(vNodes, i)), 1); i++  {
            if Abc_ObjIsCo(pNode) {
                pNode.Level = LevelMax; }
        }
    }

    // write the DOT header
    fmt.Fprintf(pFile, "# %s\n", "Network structure generated by ABC");
    fmt.Fprintf(pFile, "\n");
    fmt.Fprintf(pFile, "digraph network {\n");
    fmt.Fprintf(pFile, "size = \"7.5,10\";\n");
//    fprintf( pFile, "size = \"10,8.5\";\n" );
//    fprintf( pFile, "size = \"14,11\";\n" );
//    fprintf( pFile, "page = \"8,11\";\n" );
//  fprintf( pFile, "ranksep = 0.5;\n" );
//  fprintf( pFile, "nodesep = 0.5;\n" );
    fmt.Fprintf(pFile, "center = true;\n");
//    fprintf( pFile, "orientation = landscape;\n" );
//  fprintf( pFile, "edge [fontsize = 10];\n" );
//  fprintf( pFile, "edge [dir = none];\n" );
    fmt.Fprintf(pFile, "edge [dir = back];\n");
    fmt.Fprintf(pFile, "\n");

    // labels on the left of the picture
    fmt.Fprintf(pFile, "{\n");
    fmt.Fprintf(pFile, "  node [shape = plaintext];\n");
    fmt.Fprintf(pFile, "  edge [style = invis];\n");
    fmt.Fprintf(pFile, "  LevelTitle1 [label=\"\"];\n");
    fmt.Fprintf(pFile, "  LevelTitle2 [label=\"\"];\n");
    // generate node names with labels
    for  Level = LevelMax; Level >= LevelMin; Level--  {
        // the visible node name
        fmt.Fprintf(pFile, "  Level%d", Level);
        fmt.Fprintf(pFile, " [label = ");
        // label name
        fmt.Fprintf(pFile, "\"");
        fmt.Fprintf(pFile, "\"");
        fmt.Fprintf(pFile, "];\n");
    }

    // genetate the sequence of visible/invisible nodes to mark levels
    fmt.Fprintf(pFile, "  LevelTitle1 ->  LevelTitle2 ->");
    for  Level = LevelMax; Level >= LevelMin; Level--  {
        // the visible node name
        fmt.Fprintf(pFile, "  Level%d", Level);
        // the connector
        if Level != LevelMin {
            fmt.Fprintf(pFile, " ->");
        } else {
            fmt.Fprintf(pFile, ";"); }
    }
    fmt.Fprintf(pFile, "\n");
    fmt.Fprintf(pFile, "}");
    fmt.Fprintf(pFile, "\n");
    fmt.Fprintf(pFile, "\n");

    // generate title box on top
    fmt.Fprintf(pFile, "{\n");
    fmt.Fprintf(pFile, "  rank = same;\n");
    fmt.Fprintf(pFile, "  LevelTitle1;\n");
    fmt.Fprintf(pFile, "  title1 [shape=plaintext,\n");
    fmt.Fprintf(pFile, "          fontsize=20,\n");
    fmt.Fprintf(pFile, "          fontname = \"Times-Roman\",\n");
    fmt.Fprintf(pFile, "          label=\"");
    fmt.Fprintf(pFile, "%s", "Network structure visualized by ABC");
    fmt.Fprintf(pFile, "\\n");
    fmt.Fprintf(pFile, "Benchmark \\\"%s\\\". ", pNtk.pName);
    fmt.Fprintf(pFile, "Time was %s. ", Extra_TimeStamp());
    fmt.Fprintf(pFile, "\"\n");
    fmt.Fprintf(pFile, "         ];\n");
    fmt.Fprintf(pFile, "}");
    fmt.Fprintf(pFile, "\n");
    fmt.Fprintf(pFile, "\n");

    // generate statistics box
    fmt.Fprintf(pFile, "{\n");
    fmt.Fprintf(pFile, "  rank = same;\n");
    fmt.Fprintf(pFile, "  LevelTitle2;\n");
    fmt.Fprintf(pFile, "  title2 [shape=plaintext,\n");
    fmt.Fprintf(pFile, "          fontsize=18,\n");
    fmt.Fprintf(pFile, "          fontname = \"Times-Roman\",\n");
    fmt.Fprintf(pFile, "          label=\"");
    if Abc_NtkObjNum(pNtk) == Vec_PtrSize(vNodes) {
        fmt.Fprintf(pFile, "The network contains %d logic nodes and %d latches.", Abc_NtkNodeNum(pNtk), Abc_NtkLatchNum(pNtk));
    } else {
        fmt.Fprintf(pFile, "The set contains %d logic nodes and spans %d levels.", Abc_NtkCountLogicNodes(vNodes), LevelMax - LevelMin + 1); }
    fmt.Fprintf(pFile, "\\n");
    fmt.Fprintf(pFile, "\"\n");
    fmt.Fprintf(pFile, "         ];\n");
    fmt.Fprintf(pFile, "}");
    fmt.Fprintf(pFile, "\n");
    fmt.Fprintf(pFile, "\n");

    // generate the POs
    if fHasCos {
        fmt.Fprintf(pFile, "{\n");
        fmt.Fprintf(pFile, "  rank = same;\n");
        // the labeling node of this level
        fmt.Fprintf(pFile, "  Level%d;\n", LevelMax);
        // generate the PO nodes
        // inlining of macro Vec_PtrForEachEntry
        for  i = 0; (i < Vec_PtrSize(vNodes)) && (((pNode) = (Abc_Obj_t *)Vec_PtrEntry(vNodes, i)), 1); i++  {
            if !Abc_ObjIsPo(pNode) {
                continue; }
            fmt.Fprintf(pFile, "  Node%d [label = \"%s\"", pNode.Id, Abc_ObjName(pNode));
            fmt.Fprintf(pFile, ", shape = %s", "invtriangle");
            if pNode.fMarkB {
                fmt.Fprintf(pFile, ", style = filled"); }
            fmt.Fprintf(pFile, ", color = coral, fillcolor = coral");
            fmt.Fprintf(pFile, "];\n");
        }
        fmt.Fprintf(pFile, "}");
        fmt.Fprintf(pFile, "\n");
        fmt.Fprintf(pFile, "\n");
    }

    // generate nodes of each rank
    for  Level = LevelMax - fHasCos; Level >= LevelMin && Level > 0; Level--  {
        fmt.Fprintf(pFile, "{\n");
        fmt.Fprintf(pFile, "  rank = same;\n");
        // the labeling node of this level
        fmt.Fprintf(pFile, "  Level%d;\n", Level);
        // inlining of macro Abc_NtkForEachNode
        for  i = 0; (i < Vec_PtrSize((pNtk).vObjs)) && (((pNode) = Abc_NtkObj(pNtk, i)), 1); i++  {           if (pNode) == nil || !Abc_ObjIsNode(pNode) {} else {
            if int32(pNode.Level) != Level {
                continue; }
//            fprintf( pFile, "  Node%d [label = \"%d\"", pNode->Id, pNode->Id );
            if Abc_NtkIsStrash(pNtk) {
                pSopString = "";
            } else if Abc_NtkHasMapping(pNtk) && (fGateNames != 0) {
                pSopString = Mio_GateReadName((Mio_Gate_t *)pNode.pData);
            } else if Abc_NtkHasMapping(pNtk) {
                pSopString = Abc_NtkPrintSop(Mio_GateReadSop((Mio_Gate_t *)pNode.pData));
            } else {
                pSopString = Abc_NtkPrintSop((char *)pNode.pData); }
            fmt.Fprintf(pFile, "  Node%d [label = \"%d\\n%s\"", pNode.Id, pNode.Id, pSopString);

            fmt.Fprintf(pFile, ", shape = ellipse");
            if pNode.fMarkB {
                fmt.Fprintf(pFile, ", style = filled"); }
            fmt.Fprintf(pFile, "];\n");
        }
        }
        fmt.Fprintf(pFile, "}");
        fmt.Fprintf(pFile, "\n");
        fmt.Fprintf(pFile, "\n");
    }

    // generate the PI nodes if any
    if LevelMin == 0 {
        fmt.Fprintf(pFile, "{\n");
        fmt.Fprintf(pFile, "  rank = same;\n");
        // the labeling node of this level
        fmt.Fprintf(pFile, "  Level%d;\n", LevelMin);
        // generate the PO nodes
        // inlining of macro Vec_PtrForEachEntry
        for  i = 0; (i < Vec_PtrSize(vNodes)) && (((pNode) = (Abc_Obj_t *)Vec_PtrEntry(vNodes, i)), 1); i++  {
            if pNode.Level > 0 {
                continue; }
            if !Abc_ObjIsPi(pNode) {
                // check if the costant node is present
                if Abc_ObjFaninNum(pNode) == 0 && Abc_ObjFanoutNum(pNode) > 0 {
                    fmt.Fprintf(pFile, "  Node%d [label = \"Const1\"", pNode.Id);
                    fmt.Fprintf(pFile, ", shape = ellipse");
                    if pNode.fMarkB {
                        fmt.Fprintf(pFile, ", style = filled"); }
                    fmt.Fprintf(pFile, ", color = coral, fillcolor = coral");
                    fmt.Fprintf(pFile, "];\n");
                }
                continue;
            }
            fmt.Fprintf(pFile, "  Node%d [label = \"%s\"", pNode.Id, Abc_ObjName(pNode));
            fmt.Fprintf(pFile, ", shape = %s", "triangle");
            if pNode.fMarkB {
                fmt.Fprintf(pFile, ", style = filled"); }
            fmt.Fprintf(pFile, ", color = coral, fillcolor = coral");
            fmt.Fprintf(pFile, "];\n");
        }
        fmt.Fprintf(pFile, "}");
        fmt.Fprintf(pFile, "\n");
        fmt.Fprintf(pFile, "\n");
    }

//    fprintf( pFile, "{\n" );
    // inlining of macro Vec_PtrForEachEntry
    for  i = 0; (i < Vec_PtrSize(vNodes)) && (((pNode) = (Abc_Obj_t *)Vec_PtrEntry(vNodes, i)), 1); i++  {
        if !Abc_ObjIsLatch(pNode) {
            continue; }
        fmt.Fprintf(pFile, "Node%d [label = \"%s\"", pNode.Id, Abc_ObjName(pNode));
        fmt.Fprintf(pFile, ", shape = box");
        if pNode.fMarkB {
            fmt.Fprintf(pFile, ", style = filled"); }
        fmt.Fprintf(pFile, ", color = coral, fillcolor = coral");
        fmt.Fprintf(pFile, "];\n");
    }
//    fprintf( pFile, "}" );
//    fprintf( pFile, "\n" );
    fmt.Fprintf(pFile, "\n");

    // generate invisible edges from the square down
    fmt.Fprintf(pFile, "title1 -> title2 [style = invis];\n");
    // inlining of macro Vec_PtrForEachEntry
    for  i = 0; (i < Vec_PtrSize(vNodes)) && (((pNode) = (Abc_Obj_t *)Vec_PtrEntry(vNodes, i)), 1); i++  {
        if int32(pNode.Level) != LevelMax {
            continue; }
        if !Abc_ObjIsPo(pNode) {
            continue; }
        fmt.Fprintf(pFile, "title2 -> Node%d [style = invis];\n", pNode.Id);
    }
    // generate invisible edges among the COs
    Prev = -1;
    // inlining of macro Vec_PtrForEachEntry
    for  i = 0; (i < Vec_PtrSize(vNodes)) && (((pNode) = (Abc_Obj_t *)Vec_PtrEntry(vNodes, i)), 1); i++  {
        if int32(pNode.Level) != LevelMax {
            continue; }
        if !Abc_ObjIsPo(pNode) {
            continue; }
        if Prev >= 0 {
            fmt.Fprintf(pFile, "Node%d -> Node%d [style = invis];\n", Prev, pNode.Id); }
        Prev = pNode.Id;
    }

    // generate edges
    // inlining of macro Vec_PtrForEachEntry
    for  i = 0; (i < Vec_PtrSize(vNodes)) && (((pNode) = (Abc_Obj_t *)Vec_PtrEntry(vNodes, i)), 1); i++  {
        if Abc_ObjIsBi(pNode) || Abc_ObjIsBo(pNode) {
            continue; }
        // inlining of macro Abc_ObjForEachFanin
        for  k = 0; (k < Abc_ObjFaninNum(pNode)) && (((pFanin) = Abc_ObjFanin(pNode, k)), 1); k++  {
            fCompl = 0;
            if Abc_NtkIsStrash(pNtk) {
                if Abc_ObjIsBi(pFanin) {
                    fCompl = Abc_ObjFaninC(pFanin, k);
                } else {
                    fCompl = Abc_ObjFaninC(pNode, k); }
            }
            if Abc_ObjIsBi(pFanin) || Abc_ObjIsBo(pFanin) {
                pFanin = Abc_ObjFanin0(pFanin); }
            if Abc_ObjIsBi(pFanin) || Abc_ObjIsBo(pFanin) {
                pFanin = Abc_ObjFanin0(pFanin); }
            if !pFanin.fMarkC {
                continue; }

            // generate the edge from this node to the next
            fmt.Fprintf(pFile, "Node%d", pNode.Id);
            fmt.Fprintf(pFile, " -> ");
            fmt.Fprintf(pFile, "Node%d", pFanin.Id);
            fmt.Fprintf(pFile, " [style = %s", fCompl? "dotted" : "solid");
//            fprintf( pFile, ", label = \"%c\"", 'a' + k );
            fmt.Fprintf(pFile, "]");
            fmt.Fprintf(pFile, ";\n");
        }
    }

    fmt.Fprintf(pFile, "}");
    fmt.Fprintf(pFile, "\n");
    fmt.Fprintf(pFile, "\n");
    _ = (pFile).Close();

    // unmark the nodes from the set
    // inlining of macro Vec_PtrForEachEntry
    for  i = 0; (i < Vec_PtrSize(vNodes)) && (((pNode) = (Abc_Obj_t *)Vec_PtrEntry(vNodes, i)), 1); i++  {
        pNode.fMarkC = 0;
    }
    if vNodesShow != nil {
        // inlining of macro Vec_PtrForEachEntry
        for  i = 0; (i < Vec_PtrSize(vNodesShow)) && (((pNode) = (Abc_Obj_t *)Vec_PtrEntry(vNodesShow, i)), 1); i++  {
            pNode.fMarkB = 0;
        } }

    // convert the network back into BDDs if this is how it was
    if fHasBdds {
        Abc_NtkSopToBdd(pNtk); }
}


/**Function*************************************************************

  Synopsis    [Computes the printable SOP form.]

  Description []
  
  SideEffects []

  SeeAlso     []

***********************************************************************/
var spStatic_Abc_NtkPrintSop_Buffer [1000]byte;
func Abc_NtkPrintSop(pSop string) string {
    
    var pGet, pSet string;
    pSet = spStatic_Abc_NtkPrintSop_Buffer;
    for  pGet = pSop; *pGet; pGet++  {        
        if *pGet == '\n' {
            *pSet++ = '\\';
            *pSet++ = 'n';
        
        } else {
            *pSet++ = *pGet; }
    }
    *(pSet-2) = 0;
    return spStatic_Abc_NtkPrintSop_Buffer;
}

/**Function*************************************************************

  Synopsis    [Computes the printable SOP form.]

  Description []
  
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Abc_NtkCountLogicNodes(vNodes *Vec_Ptr_t) int32 {
    var pObj *Abc_Obj_t;
    var i, Counter int32;
    // inlining of macro Vec_PtrForEachEntry
    for  i = 0; (i < Vec_PtrSize(vNodes)) && (((pObj) = (Abc_Obj_t *)Vec_PtrEntry(vNodes, i)), 1); i++  {
        if !Abc_ObjIsNode(pObj) {
            continue; }
        if Abc_ObjFaninNum(pObj) == 0 && Abc_ObjFanoutNum(pObj) == 0 {
            continue; }
        Counter ++;
    }
    return Counter;
}

////////////////////////////////////////////////////////////////////////
///                       END OF FILE                                ///
////////////////////////////////////////////////////////////////////////


//ABC_NAMESPACE_IMPL_END

