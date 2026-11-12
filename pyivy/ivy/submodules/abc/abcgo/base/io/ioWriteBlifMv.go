package io

import (
	"fmt"
	"strconv"
)

import "os"

/**CFile****************************************************************

  FileName    [ioWriteBlifMv.c]

  SystemName  [ABC: Logic synthesis and verification system.]

  PackageName [Command processing package.]

  Synopsis    [Procedures to write BLIF-MV files.]

  Author      [Alan Mishchenko]
  
  Affiliation [UC Berkeley]

  Date        [Ver. 1.0. Started - June 20, 2005.]

  Revision    [$Id: ioWriteBlifMv.c,v 1.00 2005/06/20 00:00:00 alanmi Exp $]

***********************************************************************/

//#include "ioAbc.h"
//#include "base/main/main.h"
//#include "map/mio/mio.h"

//ABC_NAMESPACE_IMPL_START


////////////////////////////////////////////////////////////////////////
///                        DECLARATIONS                              ///
////////////////////////////////////////////////////////////////////////

// static void Io_NtkWriteBlifMv( FILE * pFile, Abc_Ntk_t * pNtk );

// static void Io_NtkWriteBlifMvOne( FILE * pFile, Abc_Ntk_t * pNtk );

// static void Io_NtkWriteBlifMvPis( FILE * pFile, Abc_Ntk_t * pNtk );

// static void Io_NtkWriteBlifMvPos( FILE * pFile, Abc_Ntk_t * pNtk );

// static void Io_NtkWriteBlifMvAsserts( FILE * pFile, Abc_Ntk_t * pNtk );

// static void Io_NtkWriteBlifMvNodeFanins( FILE * pFile, Abc_Obj_t * pNode );

// static void Io_NtkWriteBlifMvNode( FILE * pFile, Abc_Obj_t * pNode );

// static void Io_NtkWriteBlifMvLatch( FILE * pFile, Abc_Obj_t * pLatch );

// static void Io_NtkWriteBlifMvSubckt( FILE * pFile, Abc_Obj_t * pNode );

// static void Io_NtkWriteBlifMvValues( FILE * pFile, Abc_Obj_t * pNode );


////////////////////////////////////////////////////////////////////////
///                     FUNCTION DEFINITIONS                         ///
////////////////////////////////////////////////////////////////////////

/**Function*************************************************************

  Synopsis    [Write the network into a BLIF file with the given name.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Io_WriteBlifMv(pNtk *Abc_Ntk_t, FileName string) {
    var pFile *os.File;
    var pNtkTemp *Abc_Ntk_t;
    var i int32;
    assert(Abc_NtkIsNetlist(pNtk), "Abc_NtkIsNetlist(pNtk)");
    assert(Abc_NtkHasBlifMv(pNtk), "Abc_NtkHasBlifMv(pNtk)");
    // start writing the file
    pFile = os.OpenFile(FileName, os.O_WRONLY | os.O_CREATE | os.O_TRUNC, 0666);
    if pFile == nil {
        fmt.Fprintf(os.Stdout, "Io_WriteBlifMv(): Cannot open the output file.\n");
        return;
    }
    fmt.Fprintf(pFile, "# Benchmark \"%s\" written by ABC on %s\n", pNtk.pName, Extra_TimeStamp());
    // write the master network
    Io_NtkWriteBlifMv( pFile, pNtk );
    // write the remaining networks
    if pNtk.pDesign {
        // inlining of macro Vec_PtrForEachEntry
        for  i = 0; (i < Vec_PtrSize(pNtk.pDesign.vModules)) && (((pNtkTemp) = (Abc_Ntk_t *)Vec_PtrEntry(pNtk.pDesign.vModules, i)), 1); i++  {
            if pNtkTemp == pNtk {
                continue; }
            fmt.Fprintf(pFile, "\n\n");
            Io_NtkWriteBlifMv( pFile, pNtkTemp );
        }
    }
    _ = (pFile).Close();
}

/**Function*************************************************************

  Synopsis    [Write the network into a BLIF file with the given name.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Io_NtkWriteBlifMv(pFile *os.File, pNtk *Abc_Ntk_t) {
    assert(Abc_NtkIsNetlist(pNtk), "Abc_NtkIsNetlist(pNtk)");
    // write the model name
    fmt.Fprintf(pFile, ".model %s\n", Abc_NtkName(pNtk));
    // write the network
    Io_NtkWriteBlifMvOne( pFile, pNtk );
    // write EXDC network if it exists
    if Abc_NtkExdc(pNtk) {
        fmt.Printf("Io_NtkWriteBlifMv(): EXDC is not written.\n"); }
    // finalize the file
    fmt.Fprintf(pFile, ".end\n\n\n");
}

/**Function*************************************************************

  Synopsis    [Write one network.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Io_NtkWriteBlifMvOne(pFile *os.File, pNtk *Abc_Ntk_t) {
    var pProgress *ProgressBar;
    var pNode, pTerm, pLatch *Abc_Obj_t;
    var i int32;

    // write the PIs
    fmt.Fprintf(pFile, ".inputs");
    Io_NtkWriteBlifMvPis( pFile, pNtk );
    fmt.Fprintf(pFile, "\n");

    // write the POs
    fmt.Fprintf(pFile, ".outputs");
    Io_NtkWriteBlifMvPos( pFile, pNtk );
    fmt.Fprintf(pFile, "\n");

    // write the MV directives
    fmt.Fprintf(pFile, "\n");
    // inlining of macro Abc_NtkForEachCi
    for  i = 0; (i < Abc_NtkCiNum(pNtk)) && (((pTerm) = Abc_NtkCi(pNtk, i)), 1); i++  {
        if Abc_ObjMvVarNum(Abc_ObjFanout0(pTerm)) > 2 {
            fmt.Fprintf(pFile, ".mv %s %d\n", Abc_ObjName(Abc_ObjFanout0(pTerm)), Abc_ObjMvVarNum(Abc_ObjFanout0(pTerm))); }
    }
    // inlining of macro Abc_NtkForEachCo
    for  i = 0; (i < Abc_NtkCoNum(pNtk)) && (((pTerm) = Abc_NtkCo(pNtk, i)), 1); i++  {
        if Abc_ObjMvVarNum(Abc_ObjFanin0(pTerm)) > 2 {
            fmt.Fprintf(pFile, ".mv %s %d\n", Abc_ObjName(Abc_ObjFanin0(pTerm)), Abc_ObjMvVarNum(Abc_ObjFanin0(pTerm))); }
    }

    // write the blackbox
    if Abc_NtkHasBlackbox( pNtk ) {
        fmt.Fprintf(pFile, ".blackbox\n");
        return;
    }

    // write the timing info
//    Io_WriteTimingInfo( pFile, pNtk );

    // write the latches
    if !Abc_NtkIsComb(pNtk) {
        fmt.Fprintf(pFile, "\n");
        // inlining of macro Abc_NtkForEachLatch
        for  i = 0; (i < Vec_PtrSize((pNtk).vBoxes)) && (((pLatch) = Abc_NtkBox(pNtk, i)), 1); i++  {           if !Abc_ObjIsLatch(pLatch) {} else {
            Io_NtkWriteBlifMvLatch( pFile, pLatch ); }
        }
        fmt.Fprintf(pFile, "\n");
    }
/*
    // write the subcircuits
    assert( Abc_NtkWhiteboxNum(pNtk) == 0 );
    if ( Abc_NtkBlackboxNum(pNtk) > 0 )
    {
        fprintf( pFile, "\n" );
        Abc_NtkForEachBlackbox( pNtk, pNode, i )
            Io_NtkWriteBlifMvSubckt( pFile, pNode );
        fprintf( pFile, "\n" );
    }
*/
    if Abc_NtkBlackboxNum(pNtk) > 0 || Abc_NtkWhiteboxNum(pNtk) > 0 {
        fmt.Fprintf(pFile, "\n");
        // inlining of macro Abc_NtkForEachBox
        for  i = 0; (i < Vec_PtrSize((pNtk).vBoxes)) && (((pNode) = Abc_NtkBox(pNtk, i)), 1); i++  {
            if Abc_ObjIsLatch(pNode) {
                continue; }
            Io_NtkWriteBlifMvSubckt( pFile, pNode );
        }
        fmt.Fprintf(pFile, "\n");
    }

    // write each internal node
    pProgress = Extra_ProgressBarStart( os.Stdout, Abc_NtkObjNumMax(pNtk) );
    // inlining of macro Abc_NtkForEachNode
    for  i = 0; (i < Vec_PtrSize((pNtk).vObjs)) && (((pNode) = Abc_NtkObj(pNtk, i)), 1); i++  {           if (pNode) == nil || !Abc_ObjIsNode(pNode) {} else {
        Extra_ProgressBarUpdate( pProgress, i, nil );
        Io_NtkWriteBlifMvNode( pFile, pNode );
    }
    }
    Extra_ProgressBarStop( pProgress );
}


/**Function*************************************************************

  Synopsis    [Writes the primary input list.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Io_NtkWriteBlifMvPis(pFile *os.File, pNtk *Abc_Ntk_t) {
    var pTerm, pNet *Abc_Obj_t;
    var LineLength int32;
    var AddedLength int32;
    var NameCounter int32;
    var i int32;

    LineLength  = 7;
    NameCounter = 0;

    // inlining of macro Abc_NtkForEachPi
    for  i = 0; (i < Abc_NtkPiNum(pNtk)) && (((pTerm) = Abc_NtkPi(pNtk, i)), 1); i++  {
        pNet = Abc_ObjFanout0(pTerm);
        // get the line length after this name is written
        AddedLength = len(Abc_ObjName(pNet)) + 1;
        if (NameCounter != 0) && LineLength + AddedLength + 3 > IO_WRITE_LINE_LENGTH { // write the line extender
            fmt.Fprintf(pFile, " \\\n");
            // reset the line length
            LineLength  = 0;
            NameCounter = 0;
        }
        fmt.Fprintf(pFile, " %s", Abc_ObjName(pNet));
        LineLength += AddedLength;
        NameCounter++;
    }
}

/**Function*************************************************************

  Synopsis    [Writes the primary input list.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Io_NtkWriteBlifMvPos(pFile *os.File, pNtk *Abc_Ntk_t) {
    var pTerm, pNet *Abc_Obj_t;
    var LineLength int32;
    var AddedLength int32;
    var NameCounter int32;
    var i int32;

    LineLength  = 8;
    NameCounter = 0;

    // inlining of macro Abc_NtkForEachPo
    for  i = 0; (i < Abc_NtkPoNum(pNtk)) && (((pTerm) = Abc_NtkPo(pNtk, i)), 1); i++  {
        pNet = Abc_ObjFanin0(pTerm);
        // get the line length after this name is written
        AddedLength = len(Abc_ObjName(pNet)) + 1;
        if (NameCounter != 0) && LineLength + AddedLength + 3 > IO_WRITE_LINE_LENGTH { // write the line extender
            fmt.Fprintf(pFile, " \\\n");
            // reset the line length
            LineLength  = 0;
            NameCounter = 0;
        }
        fmt.Fprintf(pFile, " %s", Abc_ObjName(pNet));
        LineLength += AddedLength;
        NameCounter++;
    }
}

/**Function*************************************************************

  Synopsis    [Write the latch into a file.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Io_NtkWriteBlifMvLatch(pFile *os.File, pLatch *Abc_Obj_t) {
    var pNetLi, pNetLo *Abc_Obj_t;
    var Reset int32;
    pNetLi = Abc_ObjFanin0( Abc_ObjFanin0(pLatch) );
    pNetLo = Abc_ObjFanout0( Abc_ObjFanout0(pLatch) );
    Reset  = int32((ABC_PTRUINT_T))Abc_ObjData( pLatch );
    // write the latch line
    fmt.Fprintf(pFile, ".latch");
    fmt.Fprintf(pFile, " %10s", Abc_ObjName(pNetLi));
    fmt.Fprintf(pFile, " %10s", Abc_ObjName(pNetLo));
    fmt.Fprintf(pFile, "\n");
    // write the reset node
    fmt.Fprintf(pFile, ".reset %s\n", Abc_ObjName(pNetLo));
    fmt.Fprintf(pFile, "%d\n", Reset-1);
}

/**Function*************************************************************

  Synopsis    [Write the latch into a file.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Io_NtkWriteBlifMvSubckt(pFile *os.File, pNode *Abc_Obj_t) {
    var pModel *Abc_Ntk_t  = (Abc_Ntk_t *)pNode.pData;
    var pTerm *Abc_Obj_t;
    var i int32;
    // write the MV directives
    fmt.Fprintf(pFile, "\n");
    // inlining of macro Abc_ObjForEachFanin
    for  i = 0; (i < Abc_ObjFaninNum(pNode)) && (((pTerm) = Abc_ObjFanin(pNode, i)), 1); i++  {
        if Abc_ObjMvVarNum(pTerm) > 2 {
            fmt.Fprintf(pFile, ".mv %s %d\n", Abc_ObjName(pTerm), Abc_ObjMvVarNum(pTerm)); }
    }
    // inlining of macro Abc_ObjForEachFanout
    for  i = 0; (i < Abc_ObjFanoutNum(pNode)) && (((pTerm) = Abc_ObjFanout(pNode, i)), 1); i++  {
        if Abc_ObjMvVarNum(pTerm) > 2 {
            fmt.Fprintf(pFile, ".mv %s %d\n", Abc_ObjName(pTerm), Abc_ObjMvVarNum(pTerm)); }
    }
    // write the subcircuit
    fmt.Fprintf(pFile, ".subckt %s %s", Abc_NtkName(pModel), Abc_ObjName(pNode));
    // write pairs of the formal=actual names
    // inlining of macro Abc_NtkForEachPi
    for  i = 0; (i < Abc_NtkPiNum(pModel)) && (((pTerm) = Abc_NtkPi(pModel, i)), 1); i++  {
        fmt.Fprintf(pFile, " %s", Abc_ObjName(Abc_ObjFanout0(pTerm)));
        pTerm = Abc_ObjFanin( pNode, i );
        fmt.Fprintf(pFile, "=%s", Abc_ObjName(Abc_ObjFanin0(pTerm)));
    }
    // inlining of macro Abc_NtkForEachPo
    for  i = 0; (i < Abc_NtkPoNum(pModel)) && (((pTerm) = Abc_NtkPo(pModel, i)), 1); i++  {
        fmt.Fprintf(pFile, " %s", Abc_ObjName(Abc_ObjFanin0(pTerm)));
        pTerm = Abc_ObjFanout( pNode, i );
        fmt.Fprintf(pFile, "=%s", Abc_ObjName(Abc_ObjFanout0(pTerm)));
    }
    fmt.Fprintf(pFile, "\n");
}


/**Function*************************************************************

  Synopsis    [Write the node into a file.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Io_NtkWriteBlifMvNode(pFile *os.File, pNode *Abc_Obj_t) {
    var pFanin *Abc_Obj_t;
    var pCur string;
    var nValues, iFanin, i int32;

    // write .mv directives for the fanins
    fmt.Fprintf(pFile, "\n");
    // inlining of macro Abc_ObjForEachFanin
    for  i = 0; (i < Abc_ObjFaninNum(pNode)) && (((pFanin) = Abc_ObjFanin(pNode, i)), 1); i++  {
//        nValues = atoi(pCur);
        nValues = Abc_ObjMvVarNum( pFanin );
        if nValues > 2 {
            fmt.Fprintf(pFile, ".mv %s %d\n", Abc_ObjName(pFanin), nValues); }
//        while ( *pCur++ != ' ' );
    }

    // write .mv directives for the node
//    nValues = atoi(pCur);
    nValues = Abc_ObjMvVarNum( Abc_ObjFanout0(pNode) );
    if nValues > 2 {
        fmt.Fprintf(pFile, ".mv %s %d\n", Abc_ObjName(Abc_ObjFanout0(pNode)), nValues); }
//    while ( *pCur++ != '\n' );

    // write the .names line
    fmt.Fprintf(pFile, ".table");
    Io_NtkWriteBlifMvNodeFanins( pFile, pNode );
    fmt.Fprintf(pFile, "\n");

    // write the cubes
    pCur = (char *)Abc_ObjData(pNode);
    if *pCur == 'd' {
        fmt.Fprintf(pFile, ".default ");
        pCur++;
    }
    // write the literals
    for  ; *pCur; pCur++  {
        fmt.Fprintf(pFile, "%c", *pCur);
        if *pCur != '=' {
            continue; }
        // get the number
        iFanin = strconv.Atoi(pCur+1);
        fmt.Fprintf(pFile, "%s", Abc_ObjName(Abc_ObjFanin(pNode,iFanin)));
        // scroll on to the next symbol
        for  *pCur != ' ' && *pCur != '\n'  {
            pCur++;
        }
        pCur--;
    }
}

/**Function*************************************************************

  Synopsis    [Writes the primary input list.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Io_NtkWriteBlifMvNodeFanins(pFile *os.File, pNode *Abc_Obj_t) {
    var pNet *Abc_Obj_t;
    var LineLength int32;
    var AddedLength int32;
    var NameCounter int32;
    var pName string;
    var i int32;

    LineLength  = 6;
    NameCounter = 0;
    // inlining of macro Abc_ObjForEachFanin
    for  i = 0; (i < Abc_ObjFaninNum(pNode)) && (((pNet) = Abc_ObjFanin(pNode, i)), 1); i++  {
        // get the fanin name
        pName = Abc_ObjName(pNet);
        // get the line length after the fanin name is written
        AddedLength = len(pName) + 1;
        if (NameCounter != 0) && LineLength + AddedLength + 3 > IO_WRITE_LINE_LENGTH { // write the line extender
            fmt.Fprintf(pFile, " \\\n");
            // reset the line length
            LineLength  = 0;
            NameCounter = 0;
        }
        fmt.Fprintf(pFile, " %s", pName);
        LineLength += AddedLength;
        NameCounter++;
    }

    // get the output name
    pName = Abc_ObjName(Abc_ObjFanout0(pNode));
    // get the line length after the output name is written
    AddedLength = len(pName) + 1;
    if (NameCounter != 0) && LineLength + AddedLength > 75 { // write the line extender
        fmt.Fprintf(pFile, " \\\n");
        // reset the line length
        LineLength  = 0;
        NameCounter = 0;
    }
    fmt.Fprintf(pFile, " %s", pName);
}

////////////////////////////////////////////////////////////////////////
///                       END OF FILE                                ///
////////////////////////////////////////////////////////////////////////


//ABC_NAMESPACE_IMPL_END

