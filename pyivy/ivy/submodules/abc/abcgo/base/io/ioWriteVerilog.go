package io

import (
	"fmt"
	"strings"
)

import "os"

/**CFile****************************************************************

  FileName    [ioWriteVerilog.c]

  SystemName  [ABC: Logic synthesis and verification system.]

  PackageName [Command processing package.]

  Synopsis    [Procedures to output a special subset of Verilog.]

  Author      [Alan Mishchenko]
  
  Affiliation [UC Berkeley]

  Date        [Ver. 1.0. Started - June 20, 2005.]

  Revision    [$Id: ioWriteVerilog.c,v 1.00 2005/06/20 00:00:00 alanmi Exp $]

***********************************************************************/

//#include "ioAbc.h"
//#include "base/main/main.h"
//#include "map/mio/mio.h"

//ABC_NAMESPACE_IMPL_START


////////////////////////////////////////////////////////////////////////
///                        DECLARATIONS                              ///
////////////////////////////////////////////////////////////////////////

// static void Io_WriteVerilogInt( FILE * pFile, Abc_Ntk_t * pNtk, int fOnlyAnds );

// static void Io_WriteVerilogPis( FILE * pFile, Abc_Ntk_t * pNtk, int Start );

// static void Io_WriteVerilogPos( FILE * pFile, Abc_Ntk_t * pNtk, int Start );

// static void Io_WriteVerilogWires( FILE * pFile, Abc_Ntk_t * pNtk, int Start );

// static void Io_WriteVerilogRegs( FILE * pFile, Abc_Ntk_t * pNtk, int Start );

// static void Io_WriteVerilogLatches( FILE * pFile, Abc_Ntk_t * pNtk );

// static void Io_WriteVerilogObjects( FILE * pFile, Abc_Ntk_t * pNtk, int fOnlyAnds );

// static int  Io_WriteVerilogWiresCount( Abc_Ntk_t * pNtk );

// static char * Io_WriteVerilogGetName( char * pName );


////////////////////////////////////////////////////////////////////////
///                     FUNCTION DEFINITIONS                         ///
////////////////////////////////////////////////////////////////////////

/**Function*************************************************************

  Synopsis    [Write verilog.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/


/**Function*************************************************************

  Synopsis    [Writes verilog.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
 

/**Function*************************************************************

  Synopsis    [Writes the primary inputs.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
 

/**Function*************************************************************

  Synopsis    [Writes the primary outputs.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/


/**Function*************************************************************

  Synopsis    [Writes the wires.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/


/**Function*************************************************************

  Synopsis    [Writes the regs.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/


/**Function*************************************************************

  Synopsis    [Writes the latches.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/


/**Function*************************************************************

  Synopsis    [Writes the nodes and boxes.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/


/**Function*************************************************************

  Synopsis    [Counts the number of wires.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/


/**Function*************************************************************

  Synopsis    [Prepares the name for writing the Verilog file.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/



/**Function*************************************************************

  Synopsis    [Write the network of K-input LUTs.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/

func Io_WriteVerilogObjectsLut(pFile *os.File, pNtk *Abc_Ntk_t, nLutSize int32) {
    var pNtkBox *Abc_Ntk_t;
    var pObj, pTerm *Abc_Obj_t;
    var i, k, Counter, nDigits, Length int32;

    // write boxes
    nDigits = Abc_Base10Log( Abc_NtkBoxNum(pNtk)-Abc_NtkLatchNum(pNtk) );
    Counter = 0;
    // inlining of macro Abc_NtkForEachBox
    for  i = 0; (i < Vec_PtrSize((pNtk).vBoxes)) && (((pObj) = Abc_NtkBox(pNtk, i)), 1); i++  {
        if Abc_ObjIsLatch(pObj) {
            continue; }
        pNtkBox = (Abc_Ntk_t *)pObj.pData;
        fmt.Fprintf(pFile, "  %s box%0*d", pNtkBox.pName, nDigits, Counter++);
        fmt.Fprintf(pFile, "(");
        // inlining of macro Abc_NtkForEachPi
        for  k = 0; (k < Abc_NtkPiNum(pNtkBox)) && (((pTerm) = Abc_NtkPi(pNtkBox, k)), 1); k++  {
            fmt.Fprintf(pFile, ".%s", Io_WriteVerilogGetName(Abc_ObjName(Abc_ObjFanout0(pTerm))));
            fmt.Fprintf(pFile, "(%s), ", Io_WriteVerilogGetName(Abc_ObjName(Abc_ObjFanin0(Abc_ObjFanin(pObj,k)))));
        }
        // inlining of macro Abc_NtkForEachPo
        for  k = 0; (k < Abc_NtkPoNum(pNtkBox)) && (((pTerm) = Abc_NtkPo(pNtkBox, k)), 1); k++  {
            fmt.Fprintf(pFile, ".%s", Io_WriteVerilogGetName(Abc_ObjName(Abc_ObjFanin0(pTerm))));
            fmt.Fprintf(pFile, "(%s)%s", Io_WriteVerilogGetName(Abc_ObjName(Abc_ObjFanout0(Abc_ObjFanout(pObj,k)))), k==Abc_NtkPoNum(pNtkBox)-1? "":", ");
        }
        fmt.Fprintf(pFile, ");\n");
    }

    // find the longest signal name
    // inlining of macro Abc_NtkForEachNode
    for  i = 0; (i < Vec_PtrSize((pNtk).vObjs)) && (((pObj) = Abc_NtkObj(pNtk, i)), 1); i++  {           if (pObj) == nil || !Abc_ObjIsNode(pObj) {} else {
        Length = Abc_MaxInt( Length, len(Io_WriteVerilogGetName(Abc_ObjName(Abc_ObjFanout0(pObj)))) );
        // inlining of macro Abc_ObjForEachFanin
        for  k = 0; (k < Abc_ObjFaninNum(pObj)) && (((pTerm) = Abc_ObjFanin(pObj, k)), 1); k++  {
            Length = Abc_MaxInt( Length, len(Io_WriteVerilogGetName(Abc_ObjName(pTerm))) );
        }
    }
    }

    // write LUT instances
    nDigits = Abc_Base10Log( Abc_NtkNodeNum(pNtk) );
    Counter = 0;
    // inlining of macro Abc_NtkForEachNode
    for  i = 0; (i < Vec_PtrSize((pNtk).vObjs)) && (((pObj) = Abc_NtkObj(pNtk, i)), 1); i++  {           if (pObj) == nil || !Abc_ObjIsNode(pObj) {} else {
        var Truth word  = Abc_SopToTruth( (char *)pObj.pData, Abc_ObjFaninNum(pObj) );
        fmt.Fprintf(pFile, "  lut%d #(%d\'h", nLutSize, 1<<nLutSize);
        if nLutSize == 6 {
            fmt.Fprintf(pFile, "%08x%08x", uint32((Truth >> 32)), uint32(Truth));
        } else {
            fmt.Fprintf(pFile, "%0*x", 1<<(nLutSize-2), Abc_InfoMask(1 << nLutSize) & uint32(Truth)); }
        fmt.Fprintf(pFile, ") lut_%0*d ( {", nDigits, Counter++);
        for  k = nLutSize - 1; k >= Abc_ObjFaninNum(pObj); k--  {
            fmt.Fprintf(pFile, "%*s, ", Length, "1\'b0");
        }
        for  k = Abc_ObjFaninNum(pObj) - 1; k >= 0; k--  {
            fmt.Fprintf(pFile, "%*s%s", Length, Io_WriteVerilogGetName(Abc_ObjName(Abc_ObjFanin(pObj, k))), k==0 ? "":", ");
        }
        fmt.Fprintf(pFile, "}, %*s );\n", Length, Io_WriteVerilogGetName(Abc_ObjName(Abc_ObjFanout0(pObj))));
    }
    }
}
func Io_WriteVerilogLutInt(pFile *os.File, pNtk *Abc_Ntk_t, nLutSize int32) {
    // write inputs and outputs
//    fprintf( pFile, "module %s ( gclk,\n   ", Abc_NtkName(pNtk) );
    fmt.Fprintf(pFile, "module %s ( ", Io_WriteVerilogGetName(Abc_NtkName(pNtk)));
    // add the clock signal if it does not exist
    if Abc_NtkLatchNum(pNtk) > 0 && Nm_ManFindIdByName(pNtk.pManName, "clock", ABC_OBJ_PI) == -1 {
        fmt.Fprintf(pFile, "clock, "); }
    // write other primary inputs
    fmt.Fprintf(pFile, "\n   ");
    if Abc_NtkPiNum(pNtk) > 0 {
        Io_WriteVerilogPis( pFile, pNtk, 3 );
        fmt.Fprintf(pFile, ",\n   ");
    }
    if Abc_NtkPoNum(pNtk) > 0 {
        Io_WriteVerilogPos( pFile, pNtk, 3 ); }
    fmt.Fprintf(pFile, "  );\n\n");
    // add the clock signal if it does not exist
    if Abc_NtkLatchNum(pNtk) > 0 && Nm_ManFindIdByName(pNtk.pManName, "clock", ABC_OBJ_PI) == -1 {
        fmt.Fprintf(pFile, "  input  clock;\n"); }
    // write inputs, outputs, registers, and wires
    if Abc_NtkPiNum(pNtk) > 0 {
//        fprintf( pFile, "  input gclk," );
        fmt.Fprintf(pFile, "  input ");
        Io_WriteVerilogPis( pFile, pNtk, 10 );
        fmt.Fprintf(pFile, ";\n");
    }
    if Abc_NtkPoNum(pNtk) > 0 {
        fmt.Fprintf(pFile, "  output");
        Io_WriteVerilogPos( pFile, pNtk, 5 );
        fmt.Fprintf(pFile, ";\n\n");
    }
    // if this is not a blackbox, write internal signals
    if !Abc_NtkHasBlackbox(pNtk) {
        if Abc_NtkLatchNum(pNtk) > 0 {
            fmt.Fprintf(pFile, "  reg");
            Io_WriteVerilogRegs( pFile, pNtk, 4 );
            fmt.Fprintf(pFile, ";\n\n");
        }
        if Io_WriteVerilogWiresCount(pNtk) > 0 {
            fmt.Fprintf(pFile, "  wire");
            Io_WriteVerilogWires( pFile, pNtk, 4 );
            fmt.Fprintf(pFile, ";\n\n");
        }
        // write nodes
        Io_WriteVerilogObjectsLut( pFile, pNtk, nLutSize );        
        // write registers
        if Abc_NtkLatchNum(pNtk) > 0 {
            fmt.Fprintf(pFile, "\n");
            Io_WriteVerilogLatches( pFile, pNtk );
        }
    }
    // finalize the file
    fmt.Fprintf(pFile, "\nendmodule\n\n");
} 
func Io_WriteVerilogLut(pNtk *Abc_Ntk_t, pFileName string, nLutSize int32) {
    var pFile *os.File;
    var pNtkTemp *Abc_Ntk_t;
    var pObj *Abc_Obj_t; 
    var i, Counter int32;
    // inlining of macro Abc_NtkForEachNode
    for  i = 0; (i < Vec_PtrSize((pNtk).vObjs)) && (((pObj) = Abc_NtkObj(pNtk, i)), 1); i++  {           if (pObj) == nil || !Abc_ObjIsNode(pObj) {} else
        if Abc_ObjFaninNum(pObj) > nLutSize {
            if Counter < 3 {
                fmt.Printf("Node \"%s\" has the fanin count (%d) larger than the LUT size (%d).\n", Abc_ObjName(pObj), Abc_ObjFaninNum(pObj), nLutSize); }
            Counter++;
        }
    }
    if Counter != 0 {
        fmt.Printf("In total, %d internal logic nodes exceed the fanin count limit. Verilog is not written.\n", Counter);
        return;
    }

    // start the output stream
    pFile = os.OpenFile(pFileName, os.O_WRONLY | os.O_CREATE | os.O_TRUNC, 0666);
    if pFile == nil {
        fmt.Fprintf(os.Stdout, "Io_WriteVerilog(): Cannot open the output file \"%s\".\n", pFileName);
        return;
    }

    // write the equations for the network
    fmt.Fprintf(pFile, "// Benchmark \"%s\" written by ABC on %s\n", pNtk.pName, Extra_TimeStamp());
    fmt.Fprintf(pFile, "\n");
    Io_WriteLutModule( pFile, nLutSize );

    pNtkTemp = Abc_NtkToNetlist( pNtk );
    Abc_NtkToSop( pNtkTemp, -1, ABC_INFINITY );
    Io_WriteVerilogLutInt( pFile, pNtkTemp, nLutSize );
    Abc_NtkDelete( pNtkTemp );

    fmt.Fprintf(pFile, "\n");
    _ = (pFile).Close();
}

////////////////////////////////////////////////////////////////////////
///                       END OF FILE                                ///
////////////////////////////////////////////////////////////////////////


//ABC_NAMESPACE_IMPL_END

