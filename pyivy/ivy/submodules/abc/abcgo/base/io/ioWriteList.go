package io

import "fmt"

import "os"

/**CFile****************************************************************

  FileName    [ioWriteList.c]

  SystemName  [ABC: Logic synthesis and verification system.]

  PackageName [Command processing package.]

  Synopsis    [Procedures to write the graph structure of sequential AIG.]

  Author      [Alan Mishchenko]
  
  Affiliation [UC Berkeley]

  Date        [Ver. 1.0. Started - June 20, 2005.]

  Revision    [$Id: ioWriteList.c,v 1.00 2005/06/20 00:00:00 alanmi Exp $]

***********************************************************************/

//#include "ioAbc.h"

//ABC_NAMESPACE_IMPL_START


/*
-------- Original Message --------
Subject: Re: abc release and retiming
Date:    Sun, 13 Nov 2005 20:31:18 -0500 (EST)
From:    Luca Carloni <luca@cs.columbia.edu>
To:      Alan Mishchenko <alanmi@eecs.berkeley.edu>

Alan,

My graph-representation file format is based on an adjacency list
representation and is indeed quite simple, in fact maybe too simple... I
used it in order to reason on relatively small weighed direct graphs. I
simply list all vertices, one per line and for each vertex "V_source" I
list all vertices that are "sinks" with respect to it, i.e. such that
there is a distinct arc between "V_source" and each of them (in
parenthesis I list the name of the edge and its weight (number of latency
on that path). For instance, if you look at the following graph, you have
that vertex "v_5" is connected to vertex "v_6" through a directed arc
called "v_5_to_v_6" whose latency is equal to 3, i.e. there are three
flip-flops on this arc. Still, notice that I sometime interpret the graph
also as the representation of a  LIS where each node corresponds to a
shell encapsulating a sequential core module (i.e. a module which does not
contain any combinational path between its inputs and its outputs). With
this representation an arc of latency 3 is interpreted as a wire where two
relay stations have been inserted in addition to the flip-flop terminating
the output of the core module.

Finally, notice that the name of the arc does not necessarily have to be
"v_5_to_v_6", but it could have been something like "arc_222" or "xyz" as
long as it is a unique name in the graph.

Thanks,
Luca

Example of graph representation
-----------------------------------------------------------------------------
v_5    >    v_6 ([v_5_to_v_6] = 3), v_12 ([v_5_to_v_12] = 2).
v_2    >    v_4 ([v_2_to_v_4] = 1), v_10_s0 ([v_2_to_v_10_s0] = 6), v_12 ([v_2_to_v_12] = 3).
v_9    >    v_10_s0 ([v_9_to_v_10_s0] = 5), v_12 ([v_9_to_v_12] = 2).
v_12    >    v_13 ([v_12_to_v_13] = 5).
v_13    >    v_14 ([v_13_to_v_14] = 1).
v_6    >    v_7 ([v_6_to_v_7] = 2).
v_4    >    v_5 ([v_4_to_v_5] = 2).
v_1    >    v_2 ([v_1_to_v_2] = 1).
v_7    >    v_8 ([v_7_to_v_8] = 2).
t    >    .
v_14    >    t ([v_14_to_t] = 1), v_5 ([v_14_to_v_5] = 1).
v_8    >    v_9 ([v_8_to_v_9] = 2).
s    >    v_1 ([s_to_v_1] = 1).
v_10_s0    >    v_10_s1 ([v_10_s0_to_v_10_s1] = 1).
v_10_s1    >    v_4 ([v_10_s1__v_4] = 1), v_8 ([v_10_s1__v_8] = 1).
-----------------------------------------------------------------------------
*/

////////////////////////////////////////////////////////////////////////
///                        DECLARATIONS                              ///
////////////////////////////////////////////////////////////////////////

// static void Io_WriteListEdge( FILE * pFile, Abc_Obj_t * pObj );

// static void Io_WriteListHost( FILE * pFile, Abc_Ntk_t * pNtk );


////////////////////////////////////////////////////////////////////////
///                     FUNCTION DEFINITIONS                         ///
////////////////////////////////////////////////////////////////////////

/**Function*************************************************************

  Synopsis    [Writes the adjacency list for a sequential AIG.]

  Description []
  
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Io_WriteList(pNtk *Abc_Ntk_t, pFileName string, fUseHost int32) {
    var pFile *os.File;
    var pObj *Abc_Obj_t;
    var i int32;

//    assert( Abc_NtkIsSeq(pNtk)  );

    // start the output stream
    pFile = os.OpenFile(pFileName, os.O_WRONLY | os.O_CREATE | os.O_TRUNC, 0666);
    if pFile == nil {
        fmt.Fprintf(os.Stdout, "Io_WriteList(): Cannot open the output file \"%s\".\n", pFileName);
        return;
    }

    fmt.Fprintf(pFile, "# Adjacency list for sequential AIG \"%s\"\n", pNtk.pName);
    fmt.Fprintf(pFile, "# written by ABC on %s\n", Extra_TimeStamp());

    // write the constant node
    if Abc_ObjFanoutNum( Abc_AigConst1(pNtk) ) > 0 {
        Io_WriteListEdge( pFile, Abc_AigConst1(pNtk) ); }

    // write the PI edges
    // inlining of macro Abc_NtkForEachPi
    for  i = 0; (i < Abc_NtkPiNum(pNtk)) && (((pObj) = Abc_NtkPi(pNtk, i)), 1); i++  {
        Io_WriteListEdge( pFile, pObj );
    }

    // write the internal nodes
    // inlining of macro Abc_AigForEachAnd
    for  i = 0; (i < Vec_PtrSize((pNtk).vObjs)) && (((pObj) = Abc_NtkObj(pNtk, i)), 1); i++  {           if (pObj) == nil || !Abc_AigNodeIsAnd(pObj) {} else {
        Io_WriteListEdge( pFile, pObj ); }
    }

    // write the host node
    if fUseHost != 0 {
        Io_WriteListHost( pFile, pNtk );
    } else {
        // inlining of macro Abc_NtkForEachPo
        for  i = 0; (i < Abc_NtkPoNum(pNtk)) && (((pObj) = Abc_NtkPo(pNtk, i)), 1); i++  {
            Io_WriteListEdge( pFile, pObj );
        } }

    fmt.Fprintf(pFile, "\n");
    _ = (pFile).Close();
}

/**Function*************************************************************

  Synopsis    [Writes the adjacency list for one edge in a sequential AIG.]

  Description []
  
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Io_WriteListEdge(pFile *os.File, pObj *Abc_Obj_t) {
    var pFanout *Abc_Obj_t;
    var i int32;
    fmt.Fprintf(pFile, "%-10s >    ", Abc_ObjName(pObj));
    // inlining of macro Abc_ObjForEachFanout
    for  i = 0; (i < Abc_ObjFanoutNum(pObj)) && (((pFanout) = Abc_ObjFanout(pObj, i)), 1); i++  {
        fmt.Fprintf(pFile, " %s", Abc_ObjName(pFanout));
        fmt.Fprintf(pFile, " ([%s_to_", Abc_ObjName(pObj));
//        fprintf( pFile, "%s] = %d)", Abc_ObjName(pFanout), Seq_ObjFanoutL(pObj, pFanout) );
        fmt.Fprintf(pFile, "%s] = %d)", Abc_ObjName(pFanout), 0);
        if i != Abc_ObjFanoutNum(pObj) - 1 {
            fmt.Fprintf(pFile, ","); }
    }
    fmt.Fprintf(pFile, ".");
    fmt.Fprintf(pFile, "\n");
}

/**Function*************************************************************

  Synopsis    [Writes the adjacency list for one edge in a sequential AIG.]

  Description []
  
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Io_WriteListHost(pFile *os.File, pNtk *Abc_Ntk_t) {
    var pObj *Abc_Obj_t;
    var i int32;

    // inlining of macro Abc_NtkForEachPo
    for  i = 0; (i < Abc_NtkPoNum(pNtk)) && (((pObj) = Abc_NtkPo(pNtk, i)), 1); i++  {
        fmt.Fprintf(pFile, "%-10s >    ", Abc_ObjName(pObj));
        fmt.Fprintf(pFile, " %s ([%s_to_%s] = %d)", "HOST", Abc_ObjName(pObj), "HOST", 0);
        fmt.Fprintf(pFile, ".");
        fmt.Fprintf(pFile, "\n");
    }

    fmt.Fprintf(pFile, "%-10s >    ", "HOST");
    // inlining of macro Abc_NtkForEachPi
    for  i = 0; (i < Abc_NtkPiNum(pNtk)) && (((pObj) = Abc_NtkPi(pNtk, i)), 1); i++  {
        fmt.Fprintf(pFile, " %s", Abc_ObjName(pObj));
        fmt.Fprintf(pFile, " ([%s_to_%s] = %d)", "HOST", Abc_ObjName(pObj), 0);
        if i != Abc_NtkPiNum(pNtk) - 1 {
            fmt.Fprintf(pFile, ","); }
    }
    fmt.Fprintf(pFile, ".");
    fmt.Fprintf(pFile, "\n");
}


/**Function*************************************************************

  Synopsis    [Writes the adjacency list for a sequential AIG.]

  Description []
  
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Io_WriteCellNet(pNtk *Abc_Ntk_t, pFileName string) {
    var pFile *os.File;
    var pObj, pFanout *Abc_Obj_t;
    var i, k int32;

    assert(Abc_NtkIsLogic(pNtk), "Abc_NtkIsLogic(pNtk)");

    // start the output stream
    pFile = os.OpenFile(pFileName, os.O_WRONLY | os.O_CREATE | os.O_TRUNC, 0666);
    if pFile == nil {
        fmt.Fprintf(os.Stdout, "Io_WriteCellNet(): Cannot open the output file \"%s\".\n", pFileName);
        return;
    }

    fmt.Fprintf(pFile, "# CellNet file for network \"%s\" written by ABC on %s\n", pNtk.pName, Extra_TimeStamp());

    // the only tricky part with writing is handling latches:
    // each latch comes with (a) single-input latch-input node, (b) latch proper, (c) single-input latch-output node
    // we arbitrarily decide to use the interger ID of the latch-input node to represent the latch in the file
    // (this ID is used for both the cell and the net driven by that cell)

    // write the PIs
    // inlining of macro Abc_NtkForEachPi
    for  i = 0; (i < Abc_NtkPiNum(pNtk)) && (((pObj) = Abc_NtkPi(pNtk, i)), 1); i++  {
        fmt.Fprintf(pFile, "cell %d is 0\n", pObj.Id);
    }
    // write the POs
    // inlining of macro Abc_NtkForEachPo
    for  i = 0; (i < Abc_NtkPoNum(pNtk)) && (((pObj) = Abc_NtkPo(pNtk, i)), 1); i++  {
        fmt.Fprintf(pFile, "cell %d is 1\n", pObj.Id);
    }
    // write the latches (use the ID of latch input)
    // inlining of macro Abc_NtkForEachLatch
    for  i = 0; (i < Vec_PtrSize((pNtk).vBoxes)) && (((pObj) = Abc_NtkBox(pNtk, i)), 1); i++  {           if !Abc_ObjIsLatch(pObj) {} else {
        fmt.Fprintf(pFile, "cell %d is 2\n", Abc_ObjFanin0(pObj).Id); }
    }
    // write the logic nodes
    // inlining of macro Abc_NtkForEachNode
    for  i = 0; (i < Vec_PtrSize((pNtk).vObjs)) && (((pObj) = Abc_NtkObj(pNtk, i)), 1); i++  {           if (pObj) == nil || !Abc_ObjIsNode(pObj) {} else {
        fmt.Fprintf(pFile, "cell %d is %d\n", pObj.Id, 3+Abc_ObjFaninNum(pObj)); }
    }

    // write the nets driven by PIs
    // inlining of macro Abc_NtkForEachPi
    for  i = 0; (i < Abc_NtkPiNum(pNtk)) && (((pObj) = Abc_NtkPi(pNtk, i)), 1); i++  {
        fmt.Fprintf(pFile, "net %d  %d 0", pObj.Id, pObj.Id);
        // inlining of macro Abc_ObjForEachFanout
        for  k = 0; (k < Abc_ObjFanoutNum(pObj)) && (((pFanout) = Abc_ObjFanout(pObj, k)), 1); k++  {
            fmt.Fprintf(pFile, "  %d %d", pFanout.Id, 1 + Abc_ObjFanoutFaninNum(pFanout, pObj));
        }
        fmt.Fprintf(pFile, "\n");
    }
    // write the nets driven by latches
    // inlining of macro Abc_NtkForEachLatch
    for  i = 0; (i < Vec_PtrSize((pNtk).vBoxes)) && (((pObj) = Abc_NtkBox(pNtk, i)), 1); i++  {           if !Abc_ObjIsLatch(pObj) {} else {
        fmt.Fprintf(pFile, "net %d  %d 0", Abc_ObjFanin0(pObj).Id, Abc_ObjFanin0(pObj).Id);
        pObj = Abc_ObjFanout0(pObj);
        // inlining of macro Abc_ObjForEachFanout
        for  k = 0; (k < Abc_ObjFanoutNum(pObj)) && (((pFanout) = Abc_ObjFanout(pObj, k)), 1); k++  {
            fmt.Fprintf(pFile, "  %d %d", pFanout.Id, 1 + Abc_ObjFanoutFaninNum(pFanout, pObj));
        }
        fmt.Fprintf(pFile, "\n");
    }
    }
    // write the nets driven by nodes
    // inlining of macro Abc_NtkForEachNode
    for  i = 0; (i < Vec_PtrSize((pNtk).vObjs)) && (((pObj) = Abc_NtkObj(pNtk, i)), 1); i++  {           if (pObj) == nil || !Abc_ObjIsNode(pObj) {} else {
        fmt.Fprintf(pFile, "net %d  %d 0", pObj.Id, pObj.Id);
        // inlining of macro Abc_ObjForEachFanout
        for  k = 0; (k < Abc_ObjFanoutNum(pObj)) && (((pFanout) = Abc_ObjFanout(pObj, k)), 1); k++  {
            fmt.Fprintf(pFile, "  %d %d", pFanout.Id, 1 + Abc_ObjFanoutFaninNum(pFanout, pObj));
        }
        fmt.Fprintf(pFile, "\n");
    }
    }

    fmt.Fprintf(pFile, "\n");
    _ = (pFile).Close();
}

////////////////////////////////////////////////////////////////////////
///                       END OF FILE                                ///
////////////////////////////////////////////////////////////////////////


//ABC_NAMESPACE_IMPL_END

