package io

import "fmt"

import "os"

/**CFile****************************************************************

  FileName    [ioWriteBaf.c]

  SystemName  [ABC: Logic synthesis and verification system.]

  PackageName [Command processing package.]

  Synopsis    [Procedures to write AIG in the binary format.]

  Author      [Alan Mishchenko]
  
  Affiliation [UC Berkeley]

  Date        [Ver. 1.0. Started - June 20, 2005.]

  Revision    [$Id: ioWriteBaf.c,v 1.00 2005/06/20 00:00:00 alanmi Exp $]

***********************************************************************/

//#include "ioAbc.h"

//ABC_NAMESPACE_IMPL_START


////////////////////////////////////////////////////////////////////////
///                        DECLARATIONS                              ///
////////////////////////////////////////////////////////////////////////

/*
    Binary Aig Format

    The motivation for this format is to have
    - compact binary representation of large AIGs (~10x more compact than BLIF)
    - consequently, fast reading/writing of large AIGs (~10x faster than BLIF)
    - representation for all tech-ind info related to an AIG
    - human-readable file header

    The header:
    (1) May contain several lines of human-readable comments.
        Each comment line begins with symbol '#' and ends with symbol '\n'.
    (2) Always contains the following data. 
        - benchmark name
        - number of primary inputs
        - number of primary outputs
        - number of latches
        - number of AIG nodes (excluding the constant 1 node)
        Each entry is followed by 0-byte (character '\0'):
    (3) Next follow the names of the PIs, POs, and latches in this order. 
        Each name is followed by 0-byte (character '\0').
        Inside each set of names (PIs, POs, latches) there should be no
        identical names but the PO names may coincide with PI/latch names.

    The body:
    (1) First part of the body contains binary information about the internal AIG nodes.
        Each internal AIG node is represented using two edges (each edge is a 4-byte integer). 
        Each integer is the fanin ID followed by 1-bit representation of the complemented attribute.
        (For example, complemented edge to node 10 will be represented as 2*10 + 1 = 21.)
        The IDs of the nodes are created as follows: Constant 1 node has ID=0. 
        CIs (PIs and latch outputs) have 1-based IDs assigned in that order.
        Each node in the array of the internal AIG nodes has the ID assigned in that order.
        The constant 1 node is not written into the file.
    (2) Second part of the body contains binary information about the edges connecting 
        the COs (POs and latch inputs) to the internal AIG nodes.
        Each edge is a 4-byte integer the same way as a node fanin.
        The latch initial value (2 bits) is stored in this integer.
*/

////////////////////////////////////////////////////////////////////////
///                     FUNCTION DEFINITIONS                         ///
////////////////////////////////////////////////////////////////////////

/**Function*************************************************************

  Synopsis    [Writes the AIG in the binary format.]

  Description []
  
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Io_WriteBaf(pNtk *Abc_Ntk_t, pFileName string) {
    var pProgress *ProgressBar;
    var pFile *os.File;
    var pObj *Abc_Obj_t;
    var i, nNodes, nAnds, nBufferSize int32;
    var pBufferNode *uint32;
    assert(Abc_NtkIsStrash(pNtk), "Abc_NtkIsStrash(pNtk)");
    // start the output stream
    pFile = os.OpenFile(pFileName, os.O_WRONLY | os.O_CREATE | os.O_TRUNC, 0666);
    if pFile == nil {
        fmt.Fprintf(os.Stdout, "Io_WriteBaf(): Cannot open the output file \"%s\".\n", pFileName);
        return;
    }

    // write the comment
    fmt.Fprintf(pFile, "# BAF (Binary Aig Format) for \"%s\" written by ABC on %s\n", pNtk.pName, Extra_TimeStamp());

    // write the network name
    fmt.Fprintf(pFile, "%s%c", pNtk.pName, 0);
    // write the number of PIs
    fmt.Fprintf(pFile, "%d%c", Abc_NtkPiNum(pNtk), 0);
    // write the number of POs
    fmt.Fprintf(pFile, "%d%c", Abc_NtkPoNum(pNtk), 0);
    // write the number of latches
    fmt.Fprintf(pFile, "%d%c", Abc_NtkLatchNum(pNtk), 0);
    // write the number of internal nodes
    fmt.Fprintf(pFile, "%d%c", Abc_NtkNodeNum(pNtk), 0);

    // write PIs
    // inlining of macro Abc_NtkForEachPi
    for  i = 0; (i < Abc_NtkPiNum(pNtk)) && (((pObj) = Abc_NtkPi(pNtk, i)), 1); i++  {
        fmt.Fprintf(pFile, "%s%c", Abc_ObjName(pObj), 0);
    }
    // write POs
    // inlining of macro Abc_NtkForEachPo
    for  i = 0; (i < Abc_NtkPoNum(pNtk)) && (((pObj) = Abc_NtkPo(pNtk, i)), 1); i++  {
        fmt.Fprintf(pFile, "%s%c", Abc_ObjName(pObj), 0);
    }
    // write latches
    // inlining of macro Abc_NtkForEachLatch
    for  i = 0; (i < Vec_PtrSize((pNtk).vBoxes)) && (((pObj) = Abc_NtkBox(pNtk, i)), 1); i++  {           if !Abc_ObjIsLatch(pObj) {} else {
        fmt.Fprintf(pFile, "%s%c", Abc_ObjName(pObj), 0);
        fmt.Fprintf(pFile, "%s%c", Abc_ObjName(Abc_ObjFanin0(pObj)), 0);
        fmt.Fprintf(pFile, "%s%c", Abc_ObjName(Abc_ObjFanout0(pObj)), 0);
    }
    }

    // set the node numbers to be used in the output file
    Abc_NtkCleanCopy( pNtk );
    nNodes = 1;
    // inlining of macro Abc_NtkForEachCi
    for  i = 0; (i < Abc_NtkCiNum(pNtk)) && (((pObj) = Abc_NtkCi(pNtk, i)), 1); i++  {
        pObj.pCopy = (Abc_Obj_t *)(ABC_PTRINT_T)nNodes++;
    }
    // inlining of macro Abc_AigForEachAnd
    for  i = 0; (i < Vec_PtrSize((pNtk).vObjs)) && (((pObj) = Abc_NtkObj(pNtk, i)), 1); i++  {           if (pObj) == nil || !Abc_AigNodeIsAnd(pObj) {} else {
        pObj.pCopy = (Abc_Obj_t *)(ABC_PTRINT_T)nNodes++; }
    }

    // write the nodes into the buffer
    nAnds = 0;
    nBufferSize = Abc_NtkNodeNum(pNtk) * 2 + Abc_NtkCoNum(pNtk);
    // inlining of macro ABC_ALLOC
    pBufferNode = ((unsigned *) malloc(sizeof(unsigned) * (size_t)(nBufferSize)));
    pProgress = Extra_ProgressBarStart( os.Stdout, nBufferSize );
    // inlining of macro Abc_AigForEachAnd
    for  i = 0; (i < Vec_PtrSize((pNtk).vObjs)) && (((pObj) = Abc_NtkObj(pNtk, i)), 1); i++  {           if (pObj) == nil || !Abc_AigNodeIsAnd(pObj) {} else {
        Extra_ProgressBarUpdate( pProgress, nAnds, nil );
        pBufferNode[nAnds++] = ((int32((ABC_PTRINT_T))Abc_ObjFanin0(pObj).pCopy) << 1) | int32(Abc_ObjFaninC0(pObj));
        pBufferNode[nAnds++] = ((int32((ABC_PTRINT_T))Abc_ObjFanin1(pObj).pCopy) << 1) | int32(Abc_ObjFaninC1(pObj));
    }
    }

    // write the COs into the buffer
    // inlining of macro Abc_NtkForEachCo
    for  i = 0; (i < Abc_NtkCoNum(pNtk)) && (((pObj) = Abc_NtkCo(pNtk, i)), 1); i++  {
        Extra_ProgressBarUpdate( pProgress, nAnds, nil );
        pBufferNode[nAnds] = ((int32((ABC_PTRINT_T))Abc_ObjFanin0(pObj).pCopy) << 1) | int32(Abc_ObjFaninC0(pObj));
        if Abc_ObjFanoutNum(pObj) > 0 && Abc_ObjIsLatch(Abc_ObjFanout0(pObj)) {
            pBufferNode[nAnds] = (pBufferNode[nAnds] << 2) | (int32((ABC_PTRINT_T))Abc_ObjData(Abc_ObjFanout0(pObj)) & 3); }
        nAnds++;
    }
    Extra_ProgressBarStop( pProgress );
    assert(nBufferSize == nAnds, "nBufferSize == nAnds");

    // write the buffer
    fwrite( pBufferNode, 1, sizeof(int) * nBufferSize, pFile );
    _ = (pFile).Close();
    // inlining of macro ABC_FREE
    ((pBufferNode) ? (free((char *) (pBufferNode)), (pBufferNode) = 0) : 0);
}


////////////////////////////////////////////////////////////////////////
///                       END OF FILE                                ///
////////////////////////////////////////////////////////////////////////


//ABC_NAMESPACE_IMPL_END

