package io

import "fmt"

import "os"

/**CFile****************************************************************

  FileName    [ioWriteBlif.c]

  SystemName  [ABC: Logic synthesis and verification system.]

  PackageName [Command processing package.]

  Synopsis    [Procedures to write BLIF files.]

  Author      [Alan Mishchenko]
  
  Affiliation [UC Berkeley]

  Date        [Ver. 1.0. Started - June 20, 2005.]

  Revision    [$Id: ioWriteBlif.c,v 1.00 2005/06/20 00:00:00 alanmi Exp $]

***********************************************************************/

//#include "ioAbc.h"
//#include "base/main/main.h"
//#include "map/mio/mio.h"
//#include "bool/kit/kit.h"
//#include "map/if/if.h"

//ABC_NAMESPACE_IMPL_START


////////////////////////////////////////////////////////////////////////
///                        DECLARATIONS                              ///
////////////////////////////////////////////////////////////////////////

// static void Io_NtkWrite( FILE * pFile, Abc_Ntk_t * pNtk, int fWriteLatches, int fBb2Wb, int fSeq );

// static void Io_NtkWriteOne( FILE * pFile, Abc_Ntk_t * pNtk, int fWriteLatches, int fBb2Wb, int fSeq );

// static void Io_NtkWritePis( FILE * pFile, Abc_Ntk_t * pNtk, int fWriteLatches );

// static void Io_NtkWritePos( FILE * pFile, Abc_Ntk_t * pNtk, int fWriteLatches );

// static void Io_NtkWriteSubckt( FILE * pFile, Abc_Obj_t * pNode );

// static void Io_NtkWriteAsserts( FILE * pFile, Abc_Ntk_t * pNtk );

// static void Io_NtkWriteNodeFanins( FILE * pFile, Abc_Obj_t * pNode );

// static int  Io_NtkWriteNode( FILE * pFile, Abc_Obj_t * pNode, int Length );

// static void Io_NtkWriteLatch( FILE * pFile, Abc_Obj_t * pLatch );


////////////////////////////////////////////////////////////////////////
///                     FUNCTION DEFINITIONS                         ///
////////////////////////////////////////////////////////////////////////

/**Function*************************************************************

  Synopsis    [Write the network into a BLIF file with the given name.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Io_WriteBlifLogic(pNtk *Abc_Ntk_t, FileName string, fWriteLatches int32) {
    var pNtkTemp *Abc_Ntk_t;
    // derive the netlist
    pNtkTemp = Abc_NtkToNetlist(pNtk);
    if pNtkTemp == nil {
        fmt.Fprintf(os.Stdout, "Writing BLIF has failed.\n");
        return;
    }
    Io_WriteBlif( pNtkTemp, FileName, fWriteLatches, 0, 0 );
    Abc_NtkDelete( pNtkTemp );
}

/**Function*************************************************************

  Synopsis    [Write the network into a BLIF file with the given name.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Io_WriteBlif(pNtk *Abc_Ntk_t, FileName string, fWriteLatches int32, fBb2Wb int32, fSeq int32) {
    var pFile *os.File;
    var pNtkTemp *Abc_Ntk_t;
    var i int32;
    assert(Abc_NtkIsNetlist(pNtk), "Abc_NtkIsNetlist(pNtk)");
    // start writing the file
    pFile = os.OpenFile(FileName, os.O_WRONLY | os.O_CREATE | os.O_TRUNC, 0666);
    if pFile == nil {
        fmt.Fprintf(os.Stdout, "Io_WriteBlif(): Cannot open the output file.\n");
        return;
    }
    fmt.Fprintf(pFile, "# Benchmark \"%s\" written by ABC on %s\n", pNtk.pName, Extra_TimeStamp());
    // write the master network
    Io_NtkWrite( pFile, pNtk, fWriteLatches, fBb2Wb, fSeq );
    // make sure there is no logic hierarchy
//    assert( Abc_NtkWhiteboxNum(pNtk) == 0 );
    // write the hierarchy if present
    if Abc_NtkBlackboxNum(pNtk) > 0 || Abc_NtkWhiteboxNum(pNtk) > 0 {
        Vec_PtrForEachEntry( Abc_Ntk_t *, pNtk.pDesign.vModules, pNtkTemp, i )
        {
            if pNtkTemp == pNtk {
                continue; }
            fmt.Fprintf(pFile, "\n\n");
            Io_NtkWrite( pFile, pNtkTemp, fWriteLatches, fBb2Wb, fSeq );
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
func Io_NtkWrite(pFile *os.File, pNtk *Abc_Ntk_t, fWriteLatches int32, fBb2Wb int32, fSeq int32) {
    var pExdc *Abc_Ntk_t;
    assert(Abc_NtkIsNetlist(pNtk), "Abc_NtkIsNetlist(pNtk)");
    // write the model name
    fmt.Fprintf(pFile, ".model %s\n", Abc_NtkName(pNtk));
    // write the network
    Io_NtkWriteOne( pFile, pNtk, fWriteLatches, fBb2Wb, fSeq );
    // write EXDC network if it exists
    pExdc = Abc_NtkExdc( pNtk );
    if pExdc != nil {
        fmt.Fprintf(pFile, "\n");
        fmt.Fprintf(pFile, ".exdc\n");
        Io_NtkWriteOne( pFile, pExdc, fWriteLatches, fBb2Wb, fSeq );
    }
    // finalize the file
    fmt.Fprintf(pFile, ".end\n");
}

/**Function*************************************************************

  Synopsis    [Write one network.]

  Description []
               
  SideEffects []

  SeeAlso     [] 

***********************************************************************/
func Io_NtkWriteConvertedBox(pFile *os.File, pNtk *Abc_Ntk_t, fSeq int32) {
    var pObj *Abc_Obj_t;
    var i, v int32;
    if fSeq != 0 {
        fmt.Fprintf(pFile, ".attrib white box seq\n");
    
    } else {
        fmt.Fprintf(pFile, ".attrib white box comb\n");
        fmt.Fprintf(pFile, ".delay 1\n");
    }
    Abc_NtkForEachPo( pNtk, pObj, i )
    { 
        // write the .names line
        fmt.Fprintf(pFile, ".names");
        Io_NtkWritePis( pFile, pNtk, 1 );
        if fSeq != 0 {
            fmt.Fprintf(pFile, " %s_in\n", Abc_ObjName(Abc_ObjFanin0(pObj)));
        } else {
            fmt.Fprintf(pFile, " %s\n", Abc_ObjName(Abc_ObjFanin0(pObj))); }
        for  v = 0; v < Abc_NtkPiNum(pNtk); v++  {
            fmt.Fprintf(pFile, "1");
        }
        fmt.Fprintf(pFile, " 1\n");
        if fSeq != 0 {
            fmt.Fprintf(pFile, ".latch %s_in %s 1\n", Abc_ObjName(Abc_ObjFanin0(pObj)), Abc_ObjName(Abc_ObjFanin0(pObj))); }
    }
}

/**Function*************************************************************

  Synopsis    [Write one network.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Io_NtkWriteOne(pFile *os.File, pNtk *Abc_Ntk_t, fWriteLatches int32, fBb2Wb int32, fSeq int32) {
    var pProgress *ProgressBar;
    var pNode, pLatch *Abc_Obj_t;
    var i, Length int32;

    // write the PIs
    fmt.Fprintf(pFile, ".inputs");
    Io_NtkWritePis( pFile, pNtk, fWriteLatches );
    fmt.Fprintf(pFile, "\n");

    // write the POs
    fmt.Fprintf(pFile, ".outputs");
    Io_NtkWritePos( pFile, pNtk, fWriteLatches );
    fmt.Fprintf(pFile, "\n");

    // write the blackbox
    if Abc_NtkHasBlackbox( pNtk ) {
        if fBb2Wb != 0 {
            Io_NtkWriteConvertedBox( pFile, pNtk, fSeq );
        } else {
            fmt.Fprintf(pFile, ".blackbox\n"); }
        return;
    }

    // write the timing info
    Io_WriteTimingInfo( pFile, pNtk );

    // write the latches
    if (fWriteLatches != 0) && !Abc_NtkIsComb(pNtk) {
        fmt.Fprintf(pFile, "\n");
        Abc_NtkForEachLatch( pNtk, pLatch, i )
            Io_NtkWriteLatch( pFile, pLatch );
        fmt.Fprintf(pFile, "\n");
    }

    // write the subcircuits
//    assert( Abc_NtkWhiteboxNum(pNtk) == 0 );
    if Abc_NtkBlackboxNum(pNtk) > 0 || Abc_NtkWhiteboxNum(pNtk) > 0 {
        fmt.Fprintf(pFile, "\n");
        Abc_NtkForEachBlackbox( pNtk, pNode, i )
            Io_NtkWriteSubckt( pFile, pNode );
        fmt.Fprintf(pFile, "\n");
        Abc_NtkForEachWhitebox( pNtk, pNode, i )
            Io_NtkWriteSubckt( pFile, pNode );
        fmt.Fprintf(pFile, "\n");
    }

    // write each internal node
    if Abc_NtkHasMapping(pNtk) {
        Length = Mio_LibraryReadGateNameMax((Mio_Library_t *)pNtk.pManFunc);
    } else {
        Length = 0;
    }
    pProgress = Extra_ProgressBarStart( os.Stdout, Abc_NtkObjNumMax(pNtk) );
    Abc_NtkForEachNode( pNtk, pNode, i )
    {
        Extra_ProgressBarUpdate( pProgress, i, nil );
        if Io_NtkWriteNode( pFile, pNode, Length ) != 0 { // skip the next node
            i++; }
    }
    Extra_ProgressBarStop( pProgress );
}


/**Function*************************************************************

  Synopsis    [Writes the primary input list.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Io_NtkWritePis(pFile *os.File, pNtk *Abc_Ntk_t, fWriteLatches int32) {
    var pTerm, pNet *Abc_Obj_t;
    var LineLength int32;
    var AddedLength int32;
    var NameCounter int32;
    var i int32;

    LineLength  = 7;
    NameCounter = 0;

    if fWriteLatches != 0 {
        Abc_NtkForEachPi( pNtk, pTerm, i )
        {
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
    
    } else {
        Abc_NtkForEachCi( pNtk, pTerm, i )
        {
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
}

/**Function*************************************************************

  Synopsis    [Writes the primary input list.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Io_NtkWritePos(pFile *os.File, pNtk *Abc_Ntk_t, fWriteLatches int32) {
    var pTerm, pNet *Abc_Obj_t;
    var LineLength int32;
    var AddedLength int32;
    var NameCounter int32;
    var i int32;

    LineLength  = 8;
    NameCounter = 0;

    if fWriteLatches != 0 {
        Abc_NtkForEachPo( pNtk, pTerm, i )
        {
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
    
    } else {
        Abc_NtkForEachCo( pNtk, pTerm, i )
        {
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
}

/**Function*************************************************************

  Synopsis    [Write the latch into a file.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Io_NtkWriteSubckt(pFile *os.File, pNode *Abc_Obj_t) {
    var pModel *Abc_Ntk_t  = (Abc_Ntk_t *)pNode.pData;
    var pTerm *Abc_Obj_t;
    var i int32;
    // write the subcircuit
//    fprintf( pFile, ".subckt %s %s", Abc_NtkName(pModel), Abc_ObjName(pNode) );
    fmt.Fprintf(pFile, ".subckt %s", Abc_NtkName(pModel));
    // write pairs of the formal=actual names
    Abc_NtkForEachPi( pModel, pTerm, i )
    {
        fmt.Fprintf(pFile, " %s", Abc_ObjName(Abc_ObjFanout0(pTerm)));
        pTerm = Abc_ObjFanin( pNode, i );
        fmt.Fprintf(pFile, "=%s", Abc_ObjName(Abc_ObjFanin0(pTerm)));
    }
    Abc_NtkForEachPo( pModel, pTerm, i )
    {
        fmt.Fprintf(pFile, " %s", Abc_ObjName(Abc_ObjFanin0(pTerm)));
        pTerm = Abc_ObjFanout( pNode, i );
        fmt.Fprintf(pFile, "=%s", Abc_ObjName(Abc_ObjFanout0(pTerm)));
    }
    fmt.Fprintf(pFile, "\n");
}

/**Function*************************************************************

  Synopsis    [Write the latch into a file.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Io_NtkWriteLatch(pFile *os.File, pLatch *Abc_Obj_t) {
    var pNetLi, pNetLo *Abc_Obj_t;
    var Reset int32;
    pNetLi = Abc_ObjFanin0( Abc_ObjFanin0(pLatch) );
    pNetLo = Abc_ObjFanout0( Abc_ObjFanout0(pLatch) );
    Reset  = int32((ABC_PTRUINT_T))Abc_ObjData( pLatch );
    // write the latch line
    fmt.Fprintf(pFile, ".latch");
    fmt.Fprintf(pFile, " %10s", Abc_ObjName(pNetLi));
    fmt.Fprintf(pFile, " %10s", Abc_ObjName(pNetLo));
    fmt.Fprintf(pFile, "  %d\n", Reset-1);
}


/**Function*************************************************************

  Synopsis    [Writes the primary input list.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Io_NtkWriteNodeFanins(pFile *os.File, pNode *Abc_Obj_t) {
    var pNet *Abc_Obj_t;
    var LineLength int32;
    var AddedLength int32;
    var NameCounter int32;
    var pName string;
    var i int32;

    LineLength  = 6;
    NameCounter = 0;
    Abc_ObjForEachFanin( pNode, pNet, i )
    {
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

/**Function*************************************************************

  Synopsis    [Writes the primary input list.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Io_NtkWriteSubcktFanins(pFile *os.File, pNode *Abc_Obj_t) {
    var pNet *Abc_Obj_t;
    var LineLength int32;
    var AddedLength int32;
    var NameCounter int32;
    var pName string;
    var i int32;

    LineLength  = 6;
    NameCounter = 0;

    // get the output name
    pName = Abc_ObjName(Abc_ObjFanout0(pNode));
    // get the line length after the output name is written
    AddedLength = len(pName) + 1;
    fmt.Fprintf(pFile, " m%d", Abc_ObjId(pNode));

    // get the input names
    Abc_ObjForEachFanin( pNode, pNet, i )
    {
        // get the fanin name
        pName = Abc_ObjName(pNet);
        // get the line length after the fanin name is written
        AddedLength = len(pName) + 3;
        if (NameCounter != 0) && LineLength + AddedLength + 3 > IO_WRITE_LINE_LENGTH { // write the line extender
            fmt.Fprintf(pFile, " \\\n");
            // reset the line length
            LineLength  = 0;
            NameCounter = 0;
        }
        fmt.Fprintf(pFile, " %c=%s", 'a'+i, pName);
        LineLength += AddedLength;
        NameCounter++;
    }

    // get the output name
    pName = Abc_ObjName(Abc_ObjFanout0(pNode));
    // get the line length after the output name is written
    AddedLength = len(pName) + 3;
    if (NameCounter != 0) && LineLength + AddedLength > 75 { // write the line extender
        fmt.Fprintf(pFile, " \\\n");
        // reset the line length
        LineLength  = 0;
        NameCounter = 0;
    }
    fmt.Fprintf(pFile, " %c=%s", 'o', pName);
}


/**Function*************************************************************

  Synopsis    [Writes the primary input list.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
var spStatic_Io_NtkWriteNodeGate_fReport int32;
func Io_NtkWriteNodeGate(pFile *os.File, pNode *Abc_Obj_t, Length int32) int32 {
    
    var pGate *Mio_Gate_t  = (Mio_Gate_t *)pNode.pData;
    var pGatePin *Mio_Pin_t;
    var pNode2 *Abc_Obj_t;
    var i int32;
    fmt.Fprintf(pFile, " %-*s ", Length, Mio_GateReadName(pGate));
    for  pGatePin = Mio_GateReadPins(pGate), i = 0; (pGatePin != nil); pGatePin = Mio_PinReadNext(pGatePin), i++  {
        fmt.Fprintf(pFile, "%s=%s ", Mio_PinReadName(pGatePin), Abc_ObjName( Abc_ObjFanin(pNode,i) ));
    }
    assert(i == Abc_ObjFaninNum(pNode), "i == Abc_ObjFaninNum(pNode)");
    fmt.Fprintf(pFile, "%s=%s", Mio_GateReadOutName(pGate), Abc_ObjName( Abc_ObjFanout0(pNode) ));
    if Mio_GateReadTwin(pGate) == nil {
        return 0; }
    pNode2 = Abc_NtkFetchTwinNode( pNode );
    if pNode2 == nil {
        if !(spStatic_Io_NtkWriteNodeGate_fReport != 0) {
            spStatic_Io_NtkWriteNodeGate_fReport = 1;
fmt.Printf("Warning: Missing second output of gate(s) \"%s\".\n", Mio_GateReadName(pGate)); }
        return 0;
    }
    fmt.Fprintf(pFile, " %s=%s", Mio_GateReadOutName((Mio_Gate_t *)pNode2.pData), Abc_ObjName( Abc_ObjFanout0(pNode2) ));
    return 1;
}

/**Function*************************************************************

  Synopsis    [Write the node into a file.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Io_NtkWriteNode(pFile *os.File, pNode *Abc_Obj_t, Length int32) int32 {
    var RetValue int32;
    if Abc_NtkHasMapping(pNode.pNtk) {
        // write the .gate line
        if Abc_ObjIsBarBuf(pNode) {
            fmt.Fprintf(pFile, ".barbuf ");
            fmt.Fprintf(pFile, "%s %s", Abc_ObjName(Abc_ObjFanin0(pNode)), Abc_ObjName(Abc_ObjFanout0(pNode)));
            fmt.Fprintf(pFile, "\n");
        
        } else {
            fmt.Fprintf(pFile, ".gate");
            RetValue = Io_NtkWriteNodeGate( pFile, pNode, Length );
            fmt.Fprintf(pFile, "\n");
        }
    
    } else {
        // write the .names line
        fmt.Fprintf(pFile, ".names");
        Io_NtkWriteNodeFanins( pFile, pNode );
        fmt.Fprintf(pFile, "\n");
        // write the cubes
        fmt.Fprintf(pFile, "%s", (char*)Abc_ObjData(pNode));
    }
    return RetValue;
}

/**Function*************************************************************

  Synopsis    [Write the node into a file.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Io_NtkWriteNodeSubckt(pFile *os.File, pNode *Abc_Obj_t, Length int32) int32 {
    var RetValue int32;
    fmt.Fprintf(pFile, ".subckt");
    Io_NtkWriteSubcktFanins( pFile, pNode );
    fmt.Fprintf(pFile, "\n");
    return RetValue;
}

/**Function*************************************************************

  Synopsis    [Writes the timing info.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Io_WriteTimingInfo(pFile *os.File, pNtk *Abc_Ntk_t) {
    var pNode *Abc_Obj_t;
    var pTime, pTimeDefIn, pTimeDefOut *Abc_Time_t;
    var i int32;

    if pNtk.pManTime == nil {
        return; }

    fmt.Fprintf(pFile, "\n");
    if pNtk.AndGateDelay != 0.0 {
        fmt.Fprintf(pFile, ".and_gate_delay %g\n", pNtk.AndGateDelay); }
    pTimeDefIn = Abc_NtkReadDefaultArrival( pNtk );
    //if ( pTimeDefIn->Rise != 0.0 || pTimeDefIn->Fall != 0.0 )
        fmt.Fprintf(pFile, ".default_input_arrival %g %g\n", pTimeDefIn.Rise, pTimeDefIn.Fall);
    pTimeDefOut = Abc_NtkReadDefaultRequired( pNtk );
    //if ( pTimeDefOut->Rise != ABC_INFINITY || pTimeDefOut->Fall != ABC_INFINITY )
        fmt.Fprintf(pFile, ".default_output_required %g %g\n", pTimeDefOut.Rise, pTimeDefOut.Fall);

    fmt.Fprintf(pFile, "\n");
    Abc_NtkForEachPi( pNtk, pNode, i )
    {
        pTime = Abc_NodeReadArrival(pNode);
        if pTime.Rise == pTimeDefIn.Rise && pTime.Fall == pTimeDefIn.Fall {
            continue; }
        fmt.Fprintf(pFile, ".input_arrival %s %g %g\n", Abc_ObjName(Abc_ObjFanout0(pNode)), pTime.Rise, pTime.Fall);
    }
    Abc_NtkForEachPo( pNtk, pNode, i )
    {
        pTime = Abc_NodeReadRequired(pNode);
        if pTime.Rise == pTimeDefOut.Rise && pTime.Fall == pTimeDefOut.Fall {
            continue; }
        fmt.Fprintf(pFile, ".output_required %s %g %g\n", Abc_ObjName(Abc_ObjFanin0(pNode)), pTime.Rise, pTime.Fall);
    }

    fmt.Fprintf(pFile, "\n");
    pTimeDefIn = Abc_NtkReadDefaultInputDrive( pNtk );
    if pTimeDefIn.Rise != 0.0 || pTimeDefIn.Fall != 0.0 {
        fmt.Fprintf(pFile, ".default_input_drive %g %g\n", pTimeDefIn.Rise, pTimeDefIn.Fall); }
    if ( Abc_NodeReadInputDrive( pNtk, 0 ) )
        Abc_NtkForEachPi( pNtk, pNode, i )
        {
            pTime = Abc_NodeReadInputDrive( pNtk, i );
            if pTime.Rise == pTimeDefIn.Rise && pTime.Fall == pTimeDefIn.Fall {
                continue; }
            fmt.Fprintf(pFile, ".input_drive %s %g %g\n", Abc_ObjName(Abc_ObjFanout0(pNode)), pTime.Rise, pTime.Fall);
        }

    pTimeDefOut = Abc_NtkReadDefaultOutputLoad( pNtk );
    if pTimeDefOut.Rise != 0.0 || pTimeDefOut.Fall != 0.0 {
        fmt.Fprintf(pFile, ".default_output_load %g %g\n", pTimeDefOut.Rise, pTimeDefOut.Fall); }
    if ( Abc_NodeReadOutputLoad( pNtk, 0 ) )
        Abc_NtkForEachPo( pNtk, pNode, i )
        {
            pTime = Abc_NodeReadOutputLoad( pNtk, i );
            if pTime.Rise == pTimeDefOut.Rise && pTime.Fall == pTimeDefOut.Fall {
                continue; }
            fmt.Fprintf(pFile, ".output_load %s %g %g\n", Abc_ObjName(Abc_ObjFanin0(pNode)), pTime.Rise, pTime.Fall);
        }

    fmt.Fprintf(pFile, "\n");
}


/**Function*************************************************************

  Synopsis    []

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Abc_NtkConvertBb2Wb(pFileNameIn string, pFileNameOut string, fSeq int32, fVerbose int32) {
    var pFile *os.File;
    var pNetlist *Abc_Ntk_t;
    // check the files
    pFile = os.Open(pFileNameIn);
    if pFile == nil {
        fmt.Printf("Input file \"%s\" cannot be opened.\n", pFileNameIn);
        return;
    }
    _ = (pFile).Close();
    // check the files
    pFile = os.OpenFile(pFileNameOut, os.O_WRONLY | os.O_CREATE | os.O_TRUNC, 0666);
    if pFile == nil {
        fmt.Printf("Output file \"%s\" cannot be opened.\n", pFileNameOut);
        return;
    }
    _ = (pFile).Close();
    // derive AIG for signal correspondence
    pNetlist = Io_ReadNetlist( pFileNameIn, Io_ReadFileType(pFileNameIn), 1 );
    if pNetlist == nil {
        fmt.Printf("Reading input file \"%s\" has failed.\n", pFileNameIn);
        return;
    }
    Io_WriteBlif( pNetlist, pFileNameOut, 1, 1, fSeq );
    Abc_NtkDelete( pNetlist );
}






/**Function*************************************************************

  Synopsis    [Transforms truth table into an SOP.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Io_NtkDeriveSop(pMem *Mem_Flex_t, uTruth word, nVars int32, vCover *Vec_Int_t) string {
    var pSop string;
    var RetValue int32  = Kit_TruthIsop( (unsigned *)&uTruth, nVars, vCover, 1 );
    assert(RetValue == 0 || RetValue == 1, "RetValue == 0 || RetValue == 1");
    // check the case of constant cover
    if Vec_IntSize(vCover) == 0 || (Vec_IntSize(vCover) == 1 && Vec_IntEntry(vCover,0) == 0) {
        var pStr0 string = " 0\n";
        var pStr1 string = " 1\n";
        assert(RetValue == 0, "RetValue == 0");
        if Vec_IntSize(vCover) == 0 {
            return pStr0;
        } else {
            return pStr1;
        }
    }
    // derive the AIG for that tree
    pSop = Abc_SopCreateFromIsop( pMem, nVars, vCover );
    if RetValue != 0 {
        Abc_SopComplement( pSop ); }
    return pSop;
}

/**Function*************************************************************

  Synopsis    [Write the node into a file.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Io_NtkWriteNodeInt(pFile *os.File, pNode *Abc_Obj_t, vCover *Vec_Int_t) {
    var pNet *Abc_Obj_t;
    var i int32;
    var nVars int32 = Abc_ObjFaninNum(pNode);
    if nVars > 7 {
        fmt.Printf("Node \"%s\" has more than 7 inputs. Writing BLIF has failed.\n", Abc_ObjName(Abc_ObjFanout0(pNode)));
        return;
    }

    fmt.Fprintf(pFile, "\n");
    if nVars <= 4 {
        // write the .names line
        fmt.Fprintf(pFile, ".names");
        Abc_ObjForEachFanin( pNode, pNet, i )
            fmt.Fprintf(pFile, " %s", Abc_ObjName(pNet));
        // get the output name
        fmt.Fprintf(pFile, " %s\n", Abc_ObjName(Abc_ObjFanout0(pNode)));
        // write the cubes
        fmt.Fprintf(pFile, "%s", (char*)Abc_ObjData(pNode));
    
    } else {
        extern int  If_Dec6PickBestMux( word t, word Cofs[2] );
        extern int  If_Dec7PickBestMux( word t[2], word c0r[2], word c1r[2] );
        extern word If_Dec6MinimumBase( word uTruth, int * pSupp, int nVarsAll, int * pnVars );
        extern void If_Dec7MinimumBase( word uTruth[2], int * pSupp, int nVarsAll, int * pnVars );
        extern word If_Dec6Perform( word t, int fDerive );
        extern word If_Dec7Perform( word t[2], int fDerive );

        var pSop string;
        var z word;
        var uTruth6 word;
        var uTruth7 [2]word;
        var Cofs6 [2]word;
        var Cofs7 [2][2]word;
        var c int32;
        var iVar int32;
        var nVarsMin [2]int32;
        var pVars [2][10]int32;

        // collect variables
        Abc_ObjForEachFanin( pNode, pNet, i )
            pVars[0][i] = pVars[1][i] = i;

        // derive truth table
        if nVars == 7 {
            Abc_SopToTruth7( (char*)Abc_ObjData(pNode), nVars, uTruth7 );
            iVar = If_Dec7PickBestMux( uTruth7, Cofs7[0], Cofs7[1] );
        
        } else {
            uTruth6 = Abc_SopToTruth( (char*)Abc_ObjData(pNode), nVars );
            iVar = If_Dec6PickBestMux( uTruth6, Cofs6 );
        }

        // perform MUX decomposition
        if iVar >= 0 {
            if nVars == 7 {
                If_Dec7MinimumBase( Cofs7[0], pVars[0], nVars, &nVarsMin[0] );
                If_Dec7MinimumBase( Cofs7[1], pVars[1], nVars, &nVarsMin[1] );
            
            } else {
                Cofs6[0] = If_Dec6MinimumBase( Cofs6[0], pVars[0], nVars, &nVarsMin[0] );
                Cofs6[1] = If_Dec6MinimumBase( Cofs6[1], pVars[1], nVars, &nVarsMin[1] );
            }
            assert(nVarsMin[0] < 5, "nVarsMin[0] < 5");
            assert(nVarsMin[1] < 5, "nVarsMin[1] < 5");
            // write MUX
            fmt.Fprintf(pFile, ".names");
            fmt.Fprintf(pFile, " %s", Abc_ObjName(Abc_ObjFanin(pNode,iVar)));
            fmt.Fprintf(pFile, " %s_cascade0", Abc_ObjName(Abc_ObjFanout0(pNode)));
            fmt.Fprintf(pFile, " %s_cascade1", Abc_ObjName(Abc_ObjFanout0(pNode)));
            fmt.Fprintf(pFile, " %s\n", Abc_ObjName(Abc_ObjFanout0(pNode)));
            fmt.Fprintf(pFile, "1-1 1\n01- 1\n");
            // write cofactors
            for  c = 0; c < 2; c++  {
                pSop = Io_NtkDeriveSop( (Mem_Flex_t *)Abc_ObjNtk(pNode).pManFunc, 
                    (word)(nVars == 7 ? Cofs7[c][0] : Cofs6[c]), nVarsMin[c], vCover );
                fmt.Fprintf(pFile, ".names");
                for  i = 0; i < nVarsMin[c]; i++  {
                    fmt.Fprintf(pFile, " %s", Abc_ObjName(Abc_ObjFanin(pNode,pVars[c][i])));
                }
                fmt.Fprintf(pFile, " %s_cascade%d\n", Abc_ObjName(Abc_ObjFanout0(pNode)), c);
                fmt.Fprintf(pFile, "%s", pSop);
            }
            return;
        }
        assert(nVars == 6 || nVars == 7, "nVars == 6 || nVars == 7");

        // try cascade decomposition
        if nVars == 7 {
            z = If_Dec7Perform( uTruth7, 1 );
            //If_Dec7Verify( uTruth7, z );
        
        } else {
            z = If_Dec6Perform( uTruth6, 1 );
            //If_Dec6Verify( uTruth6, z );
        }
        if z == 0 {
            fmt.Printf("Node \"%s\" is not decomposable. Writing BLIF has failed.\n", Abc_ObjName(Abc_ObjFanout0(pNode)));
            return;
        }

        // derive nodes
        for  c = 1; c >= 0; c--  {
            // collect fanins
            uTruth7[c]  = ((c ? z >> 32 : z) & 0xffff);
            uTruth7[c] |= (uTruth7[c] << 16);
            uTruth7[c] |= (uTruth7[c] << 32);
            for  i = 0; i < 4; i++  {
                pVars[c][i] = (z >> (c*32+16+4*i)) & 7;
            }

            // minimize truth table
            Cofs6[c] = If_Dec6MinimumBase( uTruth7[c], pVars[c], 4, &nVarsMin[c] );

            // write the nodes
            fmt.Fprintf(pFile, ".names");
            for  i = 0; i < nVarsMin[c]; i++  {
                if pVars[c][i] == 7 {
                    fmt.Fprintf(pFile, " %s_cascade", Abc_ObjName(Abc_ObjFanout0(pNode)));
                } else {
                    fmt.Fprintf(pFile, " %s", Abc_ObjName(Abc_ObjFanin(pNode,pVars[c][i]))); }
            }
            fmt.Fprintf(pFile, " %s%s\n", Abc_ObjName(Abc_ObjFanout0(pNode)), c? "" : "_cascade");

            // write SOP
            pSop = Io_NtkDeriveSop( (Mem_Flex_t *)Abc_ObjNtk(pNode).pManFunc, 
                (word)Cofs6[c], nVarsMin[c], vCover );
            fmt.Fprintf(pFile, "%s", pSop);
        }
    }
}

/**Function*************************************************************

  Synopsis    [Write the node into a file.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Io_NtkWriteNodeIntStruct(pFile *os.File, pNode *Abc_Obj_t, vCover *Vec_Int_t, pStr string) {
    var pNet *Abc_Obj_t;
    var nLeaves int32  = Abc_ObjFaninNum(pNode);
    var i, nLutLeaf, nLutLeaf2, nLutRoot, Length int32;

    // quit if parameters are wrong
    Length = len(pStr);
    if Length != 2 && Length != 3 {
        fmt.Printf("Wrong LUT struct (%s)\n", pStr);
        return;
    }
    for  i = 0; i < Length; i++  {
        if pStr[i] - '0' < 3 || pStr[i] - '0' > 6 {
            fmt.Printf("The LUT size (%d) should belong to {3,4,5,6}.\n", pStr[i] - '0');
            return;
        }
    }

    nLutLeaf  =                   pStr[0] - '0';
    if Length == 3 {
        nLutLeaf2 = pStr[1] - '0';
    } else {
        nLutLeaf2 = 0;
    }
    nLutRoot  =                   pStr[Length-1] - '0';
    if nLeaves > nLutLeaf - 1 + (nLutLeaf2 ? nLutLeaf2 - 1 : 0) + nLutRoot {
        fmt.Printf("The node size (%d) is too large for the LUT structure %s.\n", nLeaves, pStr);
        return;
    }

    // consider easy case
    fmt.Fprintf(pFile, "\n");
    if nLeaves <= Abc_MaxInt( nLutLeaf2, Abc_MaxInt(nLutLeaf, nLutRoot) ) {
        // write the .names line
        fmt.Fprintf(pFile, ".names");
        Abc_ObjForEachFanin( pNode, pNet, i )
            fmt.Fprintf(pFile, " %s", Abc_ObjName(pNet));
        // get the output name
        fmt.Fprintf(pFile, " %s\n", Abc_ObjName(Abc_ObjFanout0(pNode)));
        // write the cubes
        fmt.Fprintf(pFile, "%s", (char*)Abc_ObjData(pNode));
        return;
    
    } else {
        extern int If_CluMinimumBase( word * t, int * pSupp, int nVarsAll, int * pnVars );

        static word TruthStore[16][1<<10] = {{0}}, * pTruths[16];
        var pCube [1<<10]word;
        var pRes [1<<10]word;
        var Func0 word;
        var Func1 word;
        var Func2 word;
        var pLut0 [32]byte;
        var pLut1 [32]byte;
        var pLut2 [32]byte ;
        var pSop string;
//        int nVarsMin[3], pVars[3][20];

        if TruthStore[0][0] == 0 {
            static word Truth6[6] = {
                ABC_CONST(0xAAAAAAAAAAAAAAAA),
                ABC_CONST(0xCCCCCCCCCCCCCCCC),
                ABC_CONST(0xF0F0F0F0F0F0F0F0),
                ABC_CONST(0xFF00FF00FF00FF00),
                ABC_CONST(0xFFFF0000FFFF0000),
                ABC_CONST(0xFFFFFFFF00000000)
            };
            var nVarsMax int32  = 16;
            var nWordsMax int32  = (1 << 10);
            var i, k int32;
            assert(nVarsMax <= 16, "nVarsMax <= 16");
            for  i = 0; i < nVarsMax; i++  {
                pTruths[i] = TruthStore[i];
            }
            for  i = 0; i < 6; i++  {
                for  k = 0; k < nWordsMax; k++  {
                    pTruths[i][k] = Truth6[i];
                }
            }
            for  i = 6; i < nVarsMax; i++  {
                for  k = 0; k < nWordsMax; k++  {
                    pTruths[i][k] = ((k >> (i-6)) & 1) ? ~(word)0 : 0;
                }
            }
        }

        // collect variables
//        Abc_ObjForEachFanin( pNode, pNet, i )
//            pVars[0][i] = pVars[1][i] = pVars[2][i] = i;

        // derive truth table
        Abc_SopToTruthBig( (char*)Abc_ObjData(pNode), nLeaves, pTruths, pCube, pRes );
        if Kit_TruthIsConst0((unsigned *)pRes, nLeaves) || Kit_TruthIsConst1((unsigned *)pRes, nLeaves) {
            fmt.Fprintf(pFile, ".names %s\n %d\n", Abc_ObjName(Abc_ObjFanout0(pNode)), Kit_TruthIsConst1((unsigned *)pRes, nLeaves));
            return;
        }

//        Extra_PrintHex( stdout, (unsigned *)pRes, nLeaves );  printf( "    " );
//        Kit_DsdPrintFromTruth( (unsigned*)pRes, nLeaves );  printf( "\n" );

        // perform decomposition
        if Length == 2 {
            if !If_CluCheckExt( nil, pRes, nLeaves, nLutLeaf, nLutRoot, pLut0, pLut1, &Func0, &Func1 ) {
                Extra_PrintHex( os.Stdout, (unsigned *)pRes, nLeaves );  fmt.Printf("    ");
                Kit_DsdPrintFromTruth( (unsigned*)pRes, nLeaves );  fmt.Printf("\n");
                fmt.Printf("Node \"%s\" is not decomposable. Writing BLIF has failed.\n", Abc_ObjName(Abc_ObjFanout0(pNode)));
                return;
            }
        
        } else {
            if !If_CluCheckExt3( nil, pRes, nLeaves, nLutLeaf, nLutLeaf2, nLutRoot, pLut0, pLut1, pLut2, &Func0, &Func1, &Func2 ) {
                Extra_PrintHex( os.Stdout, (unsigned *)pRes, nLeaves );  fmt.Printf("    ");
                Kit_DsdPrintFromTruth( (unsigned*)pRes, nLeaves );  fmt.Printf("\n");
                fmt.Printf("Node \"%s\" is not decomposable. Writing BLIF has failed.\n", Abc_ObjName(Abc_ObjFanout0(pNode)));
                return;
            }
        }

        // write leaf node
        fmt.Fprintf(pFile, ".names");
        for  i = 0; i < pLut1[0]; i++  {
            fmt.Fprintf(pFile, " %s", Abc_ObjName(Abc_ObjFanin(pNode,pLut1[2+i])));
        }
        fmt.Fprintf(pFile, " %s_lut1\n", Abc_ObjName(Abc_ObjFanout0(pNode)));
        // write SOP
        pSop = Io_NtkDeriveSop( (Mem_Flex_t *)Abc_ObjNtk(pNode).pManFunc, Func1, pLut1[0], vCover );
        fmt.Fprintf(pFile, "%s", pSop);

        if Length == 3 && pLut2[0] > 0 {
            // write leaf node
            fmt.Fprintf(pFile, ".names");
            for  i = 0; i < pLut2[0]; i++  {
                if pLut2[2+i] == nLeaves {
                    fmt.Fprintf(pFile, " %s_lut1", Abc_ObjName(Abc_ObjFanout0(pNode)));
                } else {
                    fmt.Fprintf(pFile, " %s", Abc_ObjName(Abc_ObjFanin(pNode,pLut2[2+i]))); }
            }
            fmt.Fprintf(pFile, " %s_lut2\n", Abc_ObjName(Abc_ObjFanout0(pNode)));
            // write SOP
            pSop = Io_NtkDeriveSop( (Mem_Flex_t *)Abc_ObjNtk(pNode).pManFunc, Func2, pLut2[0], vCover );
            fmt.Fprintf(pFile, "%s", pSop);
        }

        // write root node
        fmt.Fprintf(pFile, ".names");
        for  i = 0; i < pLut0[0]; i++  {
            if pLut0[2+i] == nLeaves {
                fmt.Fprintf(pFile, " %s_lut1", Abc_ObjName(Abc_ObjFanout0(pNode)));
            } else if pLut0[2+i] == nLeaves+1 {
                fmt.Fprintf(pFile, " %s_lut2", Abc_ObjName(Abc_ObjFanout0(pNode)));
            } else {
                fmt.Fprintf(pFile, " %s", Abc_ObjName(Abc_ObjFanin(pNode,pLut0[2+i]))); }
        }
        fmt.Fprintf(pFile, " %s\n", Abc_ObjName(Abc_ObjFanout0(pNode)));
        // write SOP
        pSop = Io_NtkDeriveSop( (Mem_Flex_t *)Abc_ObjNtk(pNode).pManFunc, Func0, pLut0[0], vCover );
        fmt.Fprintf(pFile, "%s", pSop);
    }
}

/**Function*************************************************************

  Synopsis    [Write the node into a file.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Io_NtkWriteModelIntStruct(pFile *os.File, pNode *Abc_Obj_t, vCover *Vec_Int_t, pStr string) {
    var pNet *Abc_Obj_t;
    var nLeaves int32  = Abc_ObjFaninNum(pNode);
    var i, nLutLeaf, nLutLeaf2, nLutRoot, Length int32;

    // write the header
    fmt.Fprintf(pFile, "\n");
    fmt.Fprintf(pFile, ".model m%d\n", Abc_ObjId(pNode));
    fmt.Fprintf(pFile, ".inputs");
    for  i = 0; i < Abc_ObjFaninNum(pNode); i++  {
        fmt.Fprintf(pFile, " %c", 'a' + i);
    }
    fmt.Fprintf(pFile, "\n");
    fmt.Fprintf(pFile, ".outputs o\n");

    // quit if parameters are wrong
    Length = len(pStr);
    if Length != 2 && Length != 3 {
        fmt.Printf("Wrong LUT struct (%s)\n", pStr);
        return;
    }
    for  i = 0; i < Length; i++  {
        if pStr[i] - '0' < 3 || pStr[i] - '0' > 6 {
            fmt.Printf("The LUT size (%d) should belong to {3,4,5,6}.\n", pStr[i] - '0');
            return;
        }
    }

    nLutLeaf  =                   pStr[0] - '0';
    if Length == 3 {
        nLutLeaf2 = pStr[1] - '0';
    } else {
        nLutLeaf2 = 0;
    }
    nLutRoot  =                   pStr[Length-1] - '0';
    if nLeaves > nLutLeaf - 1 + (nLutLeaf2 ? nLutLeaf2 - 1 : 0) + nLutRoot {
        fmt.Printf("The node size (%d) is too large for the LUT structure %s.\n", nLeaves, pStr);
        return;
    }

    // consider easy case
    if nLeaves <= Abc_MaxInt( nLutLeaf2, Abc_MaxInt(nLutLeaf, nLutRoot) ) {
        // write the .names line
        fmt.Fprintf(pFile, ".names");
        Abc_ObjForEachFanin( pNode, pNet, i )
            fmt.Fprintf(pFile, " %c", 'a' + i);
        // get the output name
        fmt.Fprintf(pFile, " %s\n", "o");
        // write the cubes
        fmt.Fprintf(pFile, "%s", (char*)Abc_ObjData(pNode));
        fmt.Fprintf(pFile, ".end\n");
        return;
    
    } else {
        extern int If_CluMinimumBase( word * t, int * pSupp, int nVarsAll, int * pnVars );

        static word TruthStore[16][1<<10] = {{0}}, * pTruths[16];
        var pCube [1<<10]word;
        var pRes [1<<10]word;
        var Func0 word;
        var Func1 word;
        var Func2 word;
        var pLut0 [32]byte;
        var pLut1 [32]byte;
        var pLut2 [32]byte ;
        var pSop string;
//        int nVarsMin[3], pVars[3][20];

        if TruthStore[0][0] == 0 {
            static word Truth6[6] = {
                ABC_CONST(0xAAAAAAAAAAAAAAAA),
                ABC_CONST(0xCCCCCCCCCCCCCCCC),
                ABC_CONST(0xF0F0F0F0F0F0F0F0),
                ABC_CONST(0xFF00FF00FF00FF00),
                ABC_CONST(0xFFFF0000FFFF0000),
                ABC_CONST(0xFFFFFFFF00000000)
            };
            var nVarsMax int32  = 16;
            var nWordsMax int32  = (1 << 10);
            var i, k int32;
            assert(nVarsMax <= 16, "nVarsMax <= 16");
            for  i = 0; i < nVarsMax; i++  {
                pTruths[i] = TruthStore[i];
            }
            for  i = 0; i < 6; i++  {
                for  k = 0; k < nWordsMax; k++  {
                    pTruths[i][k] = Truth6[i];
                }
            }
            for  i = 6; i < nVarsMax; i++  {
                for  k = 0; k < nWordsMax; k++  {
                    pTruths[i][k] = ((k >> (i-6)) & 1) ? ~(word)0 : 0;
                }
            }
        }

        // collect variables
//        Abc_ObjForEachFanin( pNode, pNet, i )
//            pVars[0][i] = pVars[1][i] = pVars[2][i] = i;

        // derive truth table
        Abc_SopToTruthBig( (char*)Abc_ObjData(pNode), nLeaves, pTruths, pCube, pRes );
        if Kit_TruthIsConst0((unsigned *)pRes, nLeaves) || Kit_TruthIsConst1((unsigned *)pRes, nLeaves) {
            fmt.Fprintf(pFile, ".names %s\n %d\n", "o", Kit_TruthIsConst1((unsigned *)pRes, nLeaves));
            fmt.Fprintf(pFile, ".end\n");
            return;
        }

//        Extra_PrintHex( stdout, (unsigned *)pRes, nLeaves );  printf( "    " );
//        Kit_DsdPrintFromTruth( (unsigned*)pRes, nLeaves );  printf( "\n" );

        // perform decomposition
        if Length == 2 {
            if !If_CluCheckExt( nil, pRes, nLeaves, nLutLeaf, nLutRoot, pLut0, pLut1, &Func0, &Func1 ) {
                Extra_PrintHex( os.Stdout, (unsigned *)pRes, nLeaves );  fmt.Printf("    ");
                Kit_DsdPrintFromTruth( (unsigned*)pRes, nLeaves );  fmt.Printf("\n");
                fmt.Printf("Node \"%s\" is not decomposable. Writing BLIF has failed.\n", Abc_ObjName(Abc_ObjFanout0(pNode)));
                return;
            }
        
        } else {
            if !If_CluCheckExt3( nil, pRes, nLeaves, nLutLeaf, nLutLeaf2, nLutRoot, pLut0, pLut1, pLut2, &Func0, &Func1, &Func2 ) {
                Extra_PrintHex( os.Stdout, (unsigned *)pRes, nLeaves );  fmt.Printf("    ");
                Kit_DsdPrintFromTruth( (unsigned*)pRes, nLeaves );  fmt.Printf("\n");
                fmt.Printf("Node \"%s\" is not decomposable. Writing BLIF has failed.\n", Abc_ObjName(Abc_ObjFanout0(pNode)));
                return;
            }
        }

        // write leaf node
        fmt.Fprintf(pFile, ".names");
        for  i = 0; i < pLut1[0]; i++  {
            fmt.Fprintf(pFile, " %c", 'a' + pLut1[2+i]);
        }
        fmt.Fprintf(pFile, " lut1\n");
        // write SOP
        pSop = Io_NtkDeriveSop( (Mem_Flex_t *)Abc_ObjNtk(pNode).pManFunc, Func1, pLut1[0], vCover );
        fmt.Fprintf(pFile, "%s", pSop);

        if Length == 3 && pLut2[0] > 0 {
            // write leaf node
            fmt.Fprintf(pFile, ".names");
            for  i = 0; i < pLut2[0]; i++  {
                if pLut2[2+i] == nLeaves {
                    fmt.Fprintf(pFile, " lut1");
                } else {
                    fmt.Fprintf(pFile, " %c", 'a' + pLut2[2+i]); }
            }
            fmt.Fprintf(pFile, " lut2\n");
            // write SOP
            pSop = Io_NtkDeriveSop( (Mem_Flex_t *)Abc_ObjNtk(pNode).pManFunc, Func2, pLut2[0], vCover );
            fmt.Fprintf(pFile, "%s", pSop);
        }

        // write root node
        fmt.Fprintf(pFile, ".names");
        for  i = 0; i < pLut0[0]; i++  {
            if pLut0[2+i] == nLeaves {
                fmt.Fprintf(pFile, " lut1");
            } else if pLut0[2+i] == nLeaves+1 {
                fmt.Fprintf(pFile, " lut2");
            } else {
                fmt.Fprintf(pFile, " %c", 'a' + pLut0[2+i]); }
        }
        fmt.Fprintf(pFile, " %s\n", "o");
        // write SOP
        pSop = Io_NtkDeriveSop( (Mem_Flex_t *)Abc_ObjNtk(pNode).pManFunc, Func0, pLut0[0], vCover );
        fmt.Fprintf(pFile, "%s", pSop);
        fmt.Fprintf(pFile, ".end\n");
    }
}


/**Function*************************************************************

  Synopsis    [Write the network into a BLIF file with the given name.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Io_WriteBlifInt(pNtk *Abc_Ntk_t, FileName string, pLutStruct string, fUseHie int32) {
    var pFile *os.File;
    var vCover *Vec_Int_t;
    var pNode, pLatch *Abc_Obj_t;
    var i int32;
    assert(Abc_NtkIsNetlist(pNtk), "Abc_NtkIsNetlist(pNtk)");
    // start writing the file
    pFile = os.OpenFile(FileName, os.O_WRONLY | os.O_CREATE | os.O_TRUNC, 0666);
    if pFile == nil {
        fmt.Fprintf(os.Stdout, "Io_WriteBlifInt(): Cannot open the output file.\n");
        return;
    }
    fmt.Fprintf(pFile, "# Benchmark \"%s\" written by ABC on %s\n", pNtk.pName, Extra_TimeStamp());
    // write the model name
    fmt.Fprintf(pFile, ".model %s\n", Abc_NtkName(pNtk));
    // write the PIs
    fmt.Fprintf(pFile, ".inputs");
    Io_NtkWritePis( pFile, pNtk, 1 );
    fmt.Fprintf(pFile, "\n");
    // write the POs
    fmt.Fprintf(pFile, ".outputs");
    Io_NtkWritePos( pFile, pNtk, 1 );
    fmt.Fprintf(pFile, "\n");
    // write the latches
    if Abc_NtkLatchNum(pNtk) {
        fmt.Fprintf(pFile, "\n"); }
    Abc_NtkForEachLatch( pNtk, pLatch, i )
        Io_NtkWriteLatch( pFile, pLatch );
    if Abc_NtkLatchNum(pNtk) {
        fmt.Fprintf(pFile, "\n"); }
    // write the hierarchy
    vCover = Vec_IntAlloc( (1<<16) );
    if fUseHie != 0 {
        // write each internal node
        fmt.Fprintf(pFile, "\n");
        Abc_NtkForEachNode( pNtk, pNode, i )
            Io_NtkWriteNodeSubckt( pFile, pNode, 0 );
        fmt.Fprintf(pFile, ".end\n\n");
        // write models
        Abc_NtkForEachNode( pNtk, pNode, i )
            Io_NtkWriteModelIntStruct( pFile, pNode, vCover, pLutStruct );
        fmt.Fprintf(pFile, "\n");
    
    } else {
        // write each internal node
        Abc_NtkForEachNode( pNtk, pNode, i )
        {
            if pLutStruct {
                Io_NtkWriteNodeIntStruct( pFile, pNode, vCover, pLutStruct );
            } else {
                Io_NtkWriteNodeInt( pFile, pNode, vCover ); }
        }
        fmt.Fprintf(pFile, ".end\n\n");
    }
    Vec_IntFree( vCover );
    _ = (pFile).Close();
}

/**Function*************************************************************

  Synopsis    [Write the network into a BLIF file with the given name.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Io_WriteBlifSpecial(pNtk *Abc_Ntk_t, FileName string, pLutStruct string, fUseHie int32) {
    var pNtkTemp *Abc_Ntk_t;
    assert(Abc_NtkIsLogic(pNtk), "Abc_NtkIsLogic(pNtk)");
    Abc_NtkToSop( pNtk, -1, ABC_INFINITY );
    // derive the netlist
    pNtkTemp = Abc_NtkToNetlist(pNtk);
    if pNtkTemp == nil {
        fmt.Fprintf(os.Stdout, "Writing BLIF has failed.\n");
        return;
    }
    if pLutStruct && (fUseHie != 0) {
        Io_WriteBlifInt( pNtkTemp, FileName, pLutStruct, 1 );
    } else {
        Io_WriteBlifInt( pNtkTemp, FileName, pLutStruct, 0 ); }
    Abc_NtkDelete( pNtkTemp );
}

////////////////////////////////////////////////////////////////////////
///                       END OF FILE                                ///
////////////////////////////////////////////////////////////////////////


//ABC_NAMESPACE_IMPL_END

