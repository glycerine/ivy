package io

import (
	"fmt"
	"strconv"
	"strings"
)

import "os"

/**CFile****************************************************************

  FileName    [ioUtil.c]

  SystemName  [ABC: Logic synthesis and verification system.]

  PackageName [Command processing package.]

  Synopsis    [Procedures to write the network in BENCH format.]

  Author      [Alan Mishchenko]
  
  Affiliation [UC Berkeley]

  Date        [Ver. 1.0. Started - June 20, 2005.]

  Revision    [$Id: ioUtil.c,v 1.00 2005/06/20 00:00:00 alanmi Exp $]

***********************************************************************/

//#include "ioAbc.h"
//#include "base/main/main.h"

//ABC_NAMESPACE_IMPL_START


////////////////////////////////////////////////////////////////////////
///                        DECLARATIONS                              ///
////////////////////////////////////////////////////////////////////////

////////////////////////////////////////////////////////////////////////
///                     FUNCTION DEFINITIONS                         ///
////////////////////////////////////////////////////////////////////////

/**Function*************************************************************

  Synopsis    [Returns the file type.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Io_ReadFileType(pFileName string) Io_FileType_t {
    var pExt string;
    if pFileName == nil {
        return IO_FILE_NONE; }
    pExt = Extra_FileNameExtension( pFileName );
    if pExt == nil {
        return IO_FILE_NONE; }
    if !(int32(strings.Compare(pExt, "aig")) != 0) {
        return IO_FILE_AIGER; }
    if !(int32(strings.Compare(pExt, "baf")) != 0) {
        return IO_FILE_BAF; }
    if !(int32(strings.Compare(pExt, "bblif")) != 0) {
        return IO_FILE_BBLIF; }
    if !(int32(strings.Compare(pExt, "blif")) != 0) {
        return IO_FILE_BLIF; }
    if !(int32(strings.Compare(pExt, "bench")) != 0) {
        return IO_FILE_BENCH; }
    if !(int32(strings.Compare(pExt, "cnf")) != 0) {
        return IO_FILE_CNF; }
    if !(int32(strings.Compare(pExt, "dot")) != 0) {
        return IO_FILE_DOT; }
    if !(int32(strings.Compare(pExt, "edif")) != 0) {
        return IO_FILE_EDIF; }
    if !(int32(strings.Compare(pExt, "eqn")) != 0) {
        return IO_FILE_EQN; }
    if !(int32(strings.Compare(pExt, "gml")) != 0) {
        return IO_FILE_GML; }
    if !(int32(strings.Compare(pExt, "list")) != 0) {
        return IO_FILE_LIST; }
    if !(int32(strings.Compare(pExt, "mv")) != 0) {
        return IO_FILE_BLIFMV; }
    if !(int32(strings.Compare(pExt, "pla")) != 0) {
        return IO_FILE_PLA; }
    if !(int32(strings.Compare(pExt, "smv")) != 0) {
        return IO_FILE_SMV; }
    if !(int32(strings.Compare(pExt, "v")) != 0) {
        return IO_FILE_VERILOG; }
    return IO_FILE_UNKNOWN;
}

/**Function*************************************************************

  Synopsis    [Read the network from a file.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Io_ReadNetlist(pFileName string, FileType Io_FileType_t, fCheck int32) *Abc_Ntk_t {
    var pFile *os.File;
    var pNtk *Abc_Ntk_t;
    if FileType == IO_FILE_NONE || FileType == IO_FILE_UNKNOWN {
        fmt.Fprintf(os.Stdout, "Generic file reader requires a known file extension to open \"%s\".\n", pFileName);
        return nil;
    }
    // check if the file exists
    pFile = os.Open(pFileName);
    if pFile == nil {
        fmt.Fprintf(os.Stdout, "Cannot open input file \"%s\". ", pFileName);
        if pFileName = Extra_FileGetSimilarName( pFileName, ".blif", ".bench", ".pla", ".baf", ".aig" ) {
            fmt.Fprintf(os.Stdout, "Did you mean \"%s\"?", pFileName); }
        fmt.Fprintf(os.Stdout, "\n");
       return nil;
    }
    _ = (pFile).Close();
    // read the AIG
    if FileType == IO_FILE_AIGER || FileType == IO_FILE_BAF || FileType == IO_FILE_BBLIF {
        if FileType == IO_FILE_AIGER {
            pNtk = Io_ReadAiger( pFileName, fCheck );
        } else if FileType == IO_FILE_BAF {
            pNtk = Io_ReadBaf( pFileName, fCheck );
        } else { // if ( FileType == IO_FILE_BBLIF )
            pNtk = Io_ReadBblif( pFileName, fCheck ); }
        if pNtk == nil {
            fmt.Fprintf(os.Stdout, "Reading AIG from file has failed.\n");
            return nil;
        }
        return pNtk;
    }
    // read the new netlist
    if FileType == IO_FILE_BLIF {
//        pNtk = Io_ReadBlif( pFileName, fCheck );
        pNtk = Io_ReadBlifMv( pFileName, 0, fCheck );
    } else if Io_ReadFileType(pFileName) == IO_FILE_BLIFMV {
        pNtk = Io_ReadBlifMv( pFileName, 1, fCheck );
    } else if FileType == IO_FILE_BENCH {
        pNtk = Io_ReadBench( pFileName, fCheck );
    } else if FileType == IO_FILE_EDIF {
        pNtk = Io_ReadEdif( pFileName, fCheck );
    } else if FileType == IO_FILE_EQN {
        pNtk = Io_ReadEqn( pFileName, fCheck );
    } else if FileType == IO_FILE_PLA {
        pNtk = Io_ReadPla( pFileName, 0, 0, 0, 0, fCheck );
    } else if FileType == IO_FILE_VERILOG {
        pNtk = Io_ReadVerilog( pFileName, fCheck );
    } else {
        fmt.Fprintf(os.Stderr, "Unknown file format.\n");
        return nil;
    }
    if pNtk == nil {
        fmt.Fprintf(os.Stdout, "Reading network from file has failed.\n");
        return nil;
    }
    if (fCheck != 0) && (Abc_NtkBlackboxNum(pNtk) || Abc_NtkWhiteboxNum(pNtk)) {
        var i, fCycle int32;
        var pModel *Abc_Ntk_t;
//        fprintf( stdout, "Warning: The network contains hierarchy.\n" );
        // inlining of macro Vec_PtrForEachEntry
        for  i = 0; (i < Vec_PtrSize(pNtk.pDesign.vModules)) && (((pModel) = (Abc_Ntk_t *)Vec_PtrEntry(pNtk.pDesign.vModules, i)), 1); i++  {
                if !Abc_NtkIsAcyclicWithBoxes( pModel ) {
                    fCycle = 1; }
        }
        if fCycle != 0 {
            Abc_NtkDelete( pNtk );
            return nil;    
        }
    }
    return pNtk;
}

/**Function*************************************************************

  Synopsis    []

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func temporaryLtlStore(pNtk *Abc_Ntk_t) *Vec_Ptr_t {
    var tempStore *Vec_Ptr_t;
    var pFormula string;
    var i int32;

    if (pNtk != nil) && Vec_PtrSize( pNtk.vLtlProperties ) > 0 {
        tempStore = Vec_PtrAlloc( Vec_PtrSize( pNtk.vLtlProperties ) );
        // inlining of macro Vec_PtrForEachEntry
        for  i = 0; (i < Vec_PtrSize(pNtk.vLtlProperties)) && (((pFormula) = (char *)Vec_PtrEntry(pNtk.vLtlProperties, i)), 1); i++  {
            Vec_PtrPush( tempStore, pFormula );
        }
        assert(Vec_PtrSize( tempStore ) == Vec_PtrSize( pNtk.vLtlProperties ), "Vec_PtrSize( tempStore ) == Vec_PtrSize( pNtk.vLtlProperties )");
        return tempStore;
    
    } else {
        return nil; }
}

/**Function*************************************************************

  Synopsis    []

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func updateLtlStoreOfNtk(pNtk *Abc_Ntk_t, tempLtlStore *Vec_Ptr_t) {
    var i int32;
    var pFormula string;

    assert(tempLtlStore != nil, "tempLtlStore != nil");
    // inlining of macro Vec_PtrForEachEntry
    for  i = 0; (i < Vec_PtrSize(tempLtlStore)) && (((pFormula) = (char *)Vec_PtrEntry(tempLtlStore, i)), 1); i++  {
        Vec_PtrPush( pNtk.vLtlProperties, pFormula );
    }
}

/**Function*************************************************************

  Synopsis    [Read the network from a file.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Io_Read(pFileName string, FileType Io_FileType_t, fCheck int32, fBarBufs int32) *Abc_Ntk_t {
    var pNtk, pTemp *Abc_Ntk_t;
    var vLtl *Vec_Ptr_t;
    // get the netlist
    pNtk = Io_ReadNetlist( pFileName, FileType, fCheck );
    if pNtk == nil {
        return nil; }
    vLtl = temporaryLtlStore( pNtk );
    if !Abc_NtkIsNetlist(pNtk) {
        return pNtk; }
    // derive barbufs
    if fBarBufs != 0 {
        pNtk = Abc_NtkToBarBufs( pTemp = pNtk );
        Abc_NtkDelete( pTemp );
        assert(Abc_NtkIsLogic(pNtk), "Abc_NtkIsLogic(pNtk)");
        return pNtk;
    }
    // flatten logic hierarchy
    assert(Abc_NtkIsNetlist(pNtk), "Abc_NtkIsNetlist(pNtk)");
    if Abc_NtkWhiteboxNum(pNtk) > 0 {
        pNtk = Abc_NtkFlattenLogicHierarchy( pTemp = pNtk );
        Abc_NtkDelete( pTemp );
        if pNtk == nil {
            fmt.Fprintf(os.Stdout, "Flattening logic hierarchy has failed.\n");
            return nil;
        }
    }
    // convert blackboxes
    if Abc_NtkBlackboxNum(pNtk) > 0 {
        fmt.Printf("Hierarchy reader converted %d instances of blackboxes.\n", Abc_NtkBlackboxNum(pNtk));
        pNtk = Abc_NtkConvertBlackboxes( pTemp = pNtk );
        Abc_NtkDelete( pTemp );
        if pNtk == nil {
            fmt.Fprintf(os.Stdout, "Converting blackboxes has failed.\n");
            return nil;
        }
    }
    // consider the case of BLIF-MV
    if Io_ReadFileType(pFileName) == IO_FILE_BLIFMV {
        pNtk = Abc_NtkStrashBlifMv( pTemp = pNtk );
        Abc_NtkDelete( pTemp );
        if pNtk == nil {
            fmt.Fprintf(os.Stdout, "Converting BLIF-MV to AIG has failed.\n");
            return nil;
        }
        return pNtk;
    }
    // convert the netlist into the logic network
    pNtk = Abc_NtkToLogic( pTemp = pNtk );
    if vLtl != nil {
        updateLtlStoreOfNtk( pNtk, vLtl ); }
    Abc_NtkDelete( pTemp );
    if pNtk == nil {
        fmt.Fprintf(os.Stdout, "Converting netlist to logic network after reading has failed.\n");
        return nil;
    }
    return pNtk;
}

/**Function*************************************************************

  Synopsis    [Write the network into file.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Io_Write(pNtk *Abc_Ntk_t, pFileName string, FileType Io_FileType_t) {
    var pNtkTemp, pNtkCopy *Abc_Ntk_t;
    // check if the current network is available
    if pNtk == nil {
        fmt.Fprintf(os.Stdout, "Empty network.\n");
        return;
    }
    // check if the file extension if given
    if FileType == IO_FILE_NONE || FileType == IO_FILE_UNKNOWN {
        fmt.Fprintf(os.Stdout, "The generic file writer requires a known file extension.\n");
        return;
    }
    // write the AIG formats
    if FileType == IO_FILE_AIGER || FileType == IO_FILE_BAF {
        if !Abc_NtkIsStrash(pNtk) {
            fmt.Fprintf(os.Stdout, "Writing this format is only possible for structurally hashed AIGs.\n");
            return;
        }
        if FileType == IO_FILE_AIGER {
            Io_WriteAiger( pNtk, pFileName, 1, 0, 0 );
        } else { //if ( FileType == IO_FILE_BAF )
            Io_WriteBaf( pNtk, pFileName ); }
        return;
    }
    // write non-netlist types
    if FileType == IO_FILE_CNF {
        Io_WriteCnf( pNtk, pFileName, 0 );
        return;
    }
    if FileType == IO_FILE_DOT {
        Io_WriteDot( pNtk, pFileName );
        return;
    }
    if FileType == IO_FILE_GML {
        Io_WriteGml( pNtk, pFileName );
        return;
    }
    if FileType == IO_FILE_BBLIF {
        if !Abc_NtkIsLogic(pNtk) {
            fmt.Fprintf(os.Stdout, "Writing Binary BLIF is only possible for logic networks.\n");
            return;
        }
        if !Abc_NtkHasSop(pNtk) {
            Abc_NtkToSop( pNtk, -1, ABC_INFINITY ); }
        Io_WriteBblif( pNtk, pFileName );
        return;
    }
/*
    if ( FileType == IO_FILE_BLIFMV )
    {
        Io_WriteBlifMv( pNtk, pFileName );
        return;
    }
*/
    // convert logic network into netlist
    if FileType == IO_FILE_PLA {
        if Abc_NtkLevel(pNtk) > 1 {
            fmt.Fprintf(os.Stdout, "PLA writing is available for collapsed networks.\n");
            return;
        }
        if Abc_NtkIsComb(pNtk) {
            pNtkTemp = Abc_NtkToNetlist( pNtk );
        } else {
            fmt.Fprintf(os.Stdout, "Latches are writen into the PLA file at PI/PO pairs.\n");
            pNtkCopy = Abc_NtkDup( pNtk );
            Abc_NtkMakeComb( pNtkCopy, 0 );
            pNtkTemp = Abc_NtkToNetlist( pNtk );
            Abc_NtkDelete( pNtkCopy );
        }
        if !Abc_NtkToSop( pNtkTemp, 1, ABC_INFINITY ) {
            return; }
    
    } else if FileType == IO_FILE_MOPLA {
        pNtkTemp = Abc_NtkStrash( pNtk, 0, 0, 0 );
    
    } else if FileType == IO_FILE_BENCH {
        if !Abc_NtkIsStrash(pNtk) {
            fmt.Fprintf(os.Stdout, "Writing traditional BENCH is available for AIGs only (use \"write_bench\").\n");
            return;
        }
        pNtkTemp = Abc_NtkToNetlistBench( pNtk );
    
    } else if FileType == IO_FILE_SMV {
        if !Abc_NtkIsStrash(pNtk) {
            fmt.Fprintf(os.Stdout, "Writing traditional SMV is available for AIGs only.\n");
            return;
        }
        pNtkTemp = Abc_NtkToNetlistBench( pNtk );
    
    } else {
        pNtkTemp = Abc_NtkToNetlist( pNtk ); }

    if pNtkTemp == nil {
        fmt.Fprintf(os.Stdout, "Converting to netlist has failed.\n");
        return;
    }

    if FileType == IO_FILE_BLIF {
        if !Abc_NtkHasSop(pNtkTemp) && !Abc_NtkHasMapping(pNtkTemp) {
            Abc_NtkToSop( pNtkTemp, -1, ABC_INFINITY ); }
        Io_WriteBlif( pNtkTemp, pFileName, 1, 0, 0 );
    
    } else if FileType == IO_FILE_BLIFMV {
        if !Abc_NtkConvertToBlifMv( pNtkTemp ) {
            return; }
        Io_WriteBlifMv( pNtkTemp, pFileName );
    
    } else if FileType == IO_FILE_BENCH {
        Io_WriteBench( pNtkTemp, pFileName );
    } else if FileType == IO_FILE_BOOK {
        Io_WriteBook( pNtkTemp, pFileName );
    } else if FileType == IO_FILE_PLA {
        Io_WritePla( pNtkTemp, pFileName );
    } else if FileType == IO_FILE_MOPLA {
        Io_WriteMoPla( pNtkTemp, pFileName );
    } else if FileType == IO_FILE_EQN {
        if !Abc_NtkHasAig(pNtkTemp) {
            Abc_NtkToAig( pNtkTemp ); }
        Io_WriteEqn( pNtkTemp, pFileName );
    
    } else if FileType == IO_FILE_SMV {
        Io_WriteSmv( pNtkTemp, pFileName );
    } else if FileType == IO_FILE_VERILOG {
        if !Abc_NtkHasAig(pNtkTemp) && !Abc_NtkHasMapping(pNtkTemp) {
            Abc_NtkToAig( pNtkTemp ); }
        Io_WriteVerilog( pNtkTemp, pFileName, 0 );
    
    } else { 
        fmt.Fprintf(os.Stderr, "Unknown file format.\n"); }
    Abc_NtkDelete( pNtkTemp );
}

/**Function*************************************************************

  Synopsis    [Write the network into file.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Io_WriteHie(pNtk *Abc_Ntk_t, pBaseName string, pFileName string) {
    var pNtkTemp *Abc_Ntk_t;
    var pNtkResult *Abc_Ntk_t;
    var pNtkBase *Abc_Ntk_t = nil;
    var i int32;
    // check if the current network is available
    if pNtk == nil {
        fmt.Fprintf(os.Stdout, "Empty network.\n");
        return;
    }

    // read the base network
    assert(Abc_NtkIsStrash(pNtk) || Abc_NtkIsLogic(pNtk), "Abc_NtkIsStrash(pNtk) || Abc_NtkIsLogic(pNtk)");
    if Io_ReadFileType(pBaseName) == IO_FILE_BLIF {
        pNtkBase = Io_ReadBlifMv( pBaseName, 0, 1 );
    } else if Io_ReadFileType(pBaseName) == IO_FILE_BLIFMV {
        pNtkBase = Io_ReadBlifMv( pBaseName, 1, 1 );
    } else if Io_ReadFileType(pBaseName) == IO_FILE_VERILOG {
        pNtkBase = Io_ReadVerilog( pBaseName, 1 );
    } else { 
        fmt.Fprintf(os.Stderr, "Unknown input file format.\n"); }
    if pNtkBase == nil {
        return; }

    // flatten logic hierarchy if present
    if Abc_NtkWhiteboxNum(pNtkBase) > 0 && pNtk.nBarBufs == 0 {
        pNtkBase = Abc_NtkFlattenLogicHierarchy( pNtkTemp = pNtkBase );
        Abc_NtkDelete( pNtkTemp );
        if pNtkBase == nil {
            return; }
    }

    // reintroduce the boxes into the netlist
    if pNtk.nBarBufs > 0 {
        // derive the netlist
        pNtkResult = Abc_NtkToNetlist( pNtk );
        pNtkResult = Abc_NtkFromBarBufs( pNtkBase, pNtkTemp = pNtkResult );
        Abc_NtkDelete( pNtkTemp );
        if pNtkResult != nil {
            fmt.Printf("Hierarchy writer replaced %d barbufs by hierarchy boundaries.\n", pNtk.nBarBufs); }
    
    } else if Io_ReadFileType(pBaseName) == IO_FILE_BLIFMV {
        if Abc_NtkBlackboxNum(pNtkBase) > 0 {
            fmt.Printf("Hierarchy writer does not support BLIF-MV with blackboxes.\n");
            Abc_NtkDelete( pNtkBase );
            return;
        }
        // convert the current network to BLIF-MV
        assert(!Abc_NtkIsNetlist(pNtk), "!Abc_NtkIsNetlist(pNtk)");
        pNtkResult = Abc_NtkToNetlist( pNtk );
        if !Abc_NtkConvertToBlifMv( pNtkResult ) {
            Abc_NtkDelete( pNtkBase );
            return;
        }
        // reintroduce the network
        pNtkResult = Abc_NtkInsertBlifMv( pNtkBase, pNtkTemp = pNtkResult );
        Abc_NtkDelete( pNtkTemp );
    
    } else if Abc_NtkBlackboxNum(pNtkBase) > 0 {
        // derive the netlist
        pNtkResult = Abc_NtkToNetlist( pNtk );
        pNtkResult = Abc_NtkInsertNewLogic( pNtkBase, pNtkTemp = pNtkResult );
        Abc_NtkDelete( pNtkTemp );
        if pNtkResult != nil {
            fmt.Printf("Hierarchy writer reintroduced %d instances of blackboxes.\n", Abc_NtkBlackboxNum(pNtkBase)); }
    
    } else {
        fmt.Printf("Warning: The output network does not contain blackboxes.\n");
        pNtkResult = Abc_NtkToNetlist( pNtk );
    }
    Abc_NtkDelete( pNtkBase );
    if pNtkResult == nil {
        return; }

    // write the resulting network
    if Io_ReadFileType(pFileName) == IO_FILE_BLIF {
        if pNtkResult.pDesign {
            // inlining of macro Vec_PtrForEachEntry
            for  i = 0; (i < Vec_PtrSize(pNtkResult.pDesign.vModules)) && (((pNtkTemp) = (Abc_Ntk_t *)Vec_PtrEntry(pNtkResult.pDesign.vModules, i)), 1); i++  {
                if !Abc_NtkHasSop(pNtkTemp) && !Abc_NtkHasMapping(pNtkTemp) {
                    Abc_NtkToSop( pNtkTemp, -1, ABC_INFINITY ); }
            }
        
        } else {
            if !Abc_NtkHasSop(pNtkResult) && !Abc_NtkHasMapping(pNtkResult) {
                Abc_NtkToSop( pNtkResult, -1, ABC_INFINITY ); }
        }
        Io_WriteBlif( pNtkResult, pFileName, 1, 0, 0 );
    
    } else if Io_ReadFileType(pFileName) == IO_FILE_VERILOG {
        if pNtkResult.pDesign {
            // inlining of macro Vec_PtrForEachEntry
            for  i = 0; (i < Vec_PtrSize(pNtkResult.pDesign.vModules)) && (((pNtkTemp) = (Abc_Ntk_t *)Vec_PtrEntry(pNtkResult.pDesign.vModules, i)), 1); i++  {
                if !Abc_NtkHasAig(pNtkTemp) && !Abc_NtkHasMapping(pNtkTemp) {
                    Abc_NtkToAig( pNtkTemp ); }
            }
        
        } else {
            if !Abc_NtkHasAig(pNtkResult) && !Abc_NtkHasMapping(pNtkResult) {
                Abc_NtkToAig( pNtkResult ); }
        }
        Io_WriteVerilog( pNtkResult, pFileName, 0 );
    
    } else if Io_ReadFileType(pFileName) == IO_FILE_BLIFMV {
        Io_WriteBlifMv( pNtkResult, pFileName );
    
    } else { 
        fmt.Fprintf(os.Stderr, "Unknown output file format.\n"); }

    Abc_NtkDelete( pNtkResult );
}

/**Function*************************************************************

  Synopsis    [Creates PI terminal and net.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Io_ReadCreatePi(pNtk *Abc_Ntk_t, pName string) *Abc_Obj_t {
    var pNet, pTerm *Abc_Obj_t;
    // get the PI net
    pNet  = Abc_NtkFindNet( pNtk, pName );
    if pNet != nil {
        fmt.Printf("Warning: PI \"%s\" appears twice in the list.\n", pName); }
    pNet  = Abc_NtkFindOrCreateNet( pNtk, pName );
    // add the PI node
    pTerm = Abc_NtkCreatePi( pNtk );
    Abc_ObjAddFanin( pNet, pTerm );
    return pTerm;
}

/**Function*************************************************************

  Synopsis    [Creates PO terminal and net.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Io_ReadCreatePo(pNtk *Abc_Ntk_t, pName string) *Abc_Obj_t {
    var pNet, pTerm *Abc_Obj_t;
    // get the PO net
    pNet  = Abc_NtkFindNet( pNtk, pName );
    if (pNet != nil) && Abc_ObjFaninNum(pNet) == 0 {
        fmt.Printf("Warning: PO \"%s\" appears twice in the list.\n", pName); }
    pNet  = Abc_NtkFindOrCreateNet( pNtk, pName );
    // add the PO node
    pTerm = Abc_NtkCreatePo( pNtk );
    Abc_ObjAddFanin( pTerm, pNet );
    return pTerm;
}

/**Function*************************************************************

  Synopsis    [Create a latch with the given input/output.]

  Description [By default, the latch value is unknown (ABC_INIT_NONE).]
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Io_ReadCreateLatch(pNtk *Abc_Ntk_t, pNetLI string, pNetLO string) *Abc_Obj_t {
    var pLatch, pTerm, pNet *Abc_Obj_t;
    // get the LI net
    pNet = Abc_NtkFindOrCreateNet( pNtk, pNetLI );
    // add the BO terminal
    pTerm = Abc_NtkCreateBi( pNtk );
    Abc_ObjAddFanin( pTerm, pNet );
    // add the latch box
    pLatch = Abc_NtkCreateLatch( pNtk );
    Abc_ObjAddFanin( pLatch, pTerm  );
    // add the BI terminal
    pTerm = Abc_NtkCreateBo( pNtk );
    Abc_ObjAddFanin( pTerm, pLatch );
    // get the LO net
    pNet = Abc_NtkFindOrCreateNet( pNtk, pNetLO );
    Abc_ObjAddFanin( pNet, pTerm );
    // set latch name
    Abc_ObjAssignName( pLatch, pNetLO, "L" );
    return pLatch;
}

/**Function*************************************************************

  Synopsis    [Create the reset latch with data=1 and init=0.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Io_ReadCreateResetLatch(pNtk *Abc_Ntk_t, fBlifMv int32) *Abc_Obj_t {
    var pLatch, pNode *Abc_Obj_t;
    var pNetLI, pNetLO *Abc_Obj_t;
    // create latch with 0 init value
//    pLatch = Io_ReadCreateLatch( pNtk, "_resetLI_", "_resetLO_" );
    pNetLI = Abc_NtkCreateNet( pNtk );
    pNetLO = Abc_NtkCreateNet( pNtk );
    Abc_ObjAssignName( pNetLI, Abc_ObjName(pNetLI), nil );
    Abc_ObjAssignName( pNetLO, Abc_ObjName(pNetLO), nil );
    pLatch = Io_ReadCreateLatch( pNtk, Abc_ObjName(pNetLI), Abc_ObjName(pNetLO) );
    // set the initial value
    Abc_LatchSetInit0( pLatch );
    // feed the latch with constant1- node
//    pNode = Abc_NtkCreateNode( pNtk );   
//    pNode->pData = Abc_SopRegister( (Extra_MmFlex_t *)pNtk->pManFunc, "2\n1\n" );
    pNode = Abc_NtkCreateNodeConst1( pNtk );
    Abc_ObjAddFanin( Abc_ObjFanin0(Abc_ObjFanin0(pLatch)), pNode );
    return pLatch;
}

/**Function*************************************************************

  Synopsis    [Create node and the net driven by it.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Io_ReadCreateNode(pNtk *Abc_Ntk_t, pNameOut string, pNamesIn []string, nInputs int32) *Abc_Obj_t {
    var pNet, pNode *Abc_Obj_t;
    var i int32;
    // create a new node 
    pNode = Abc_NtkCreateNode( pNtk );
    // add the fanin nets
    for  i = 0; i < nInputs; i++  {
        pNet = Abc_NtkFindOrCreateNet( pNtk, pNamesIn[i] );
        Abc_ObjAddFanin( pNode, pNet );
    }
    // add the fanout net
    pNet = Abc_NtkFindOrCreateNet( pNtk, pNameOut );
    Abc_ObjAddFanin( pNet, pNode );
    return pNode;
}

/**Function*************************************************************

  Synopsis    [Create a constant 0 node driving the net with this name.]

  Description [Assumes that the net already exists.]
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Io_ReadCreateConst(pNtk *Abc_Ntk_t, pName string, fConst1 int32) *Abc_Obj_t {
    var pNet, pTerm *Abc_Obj_t;
    if fConst1 != 0 {
        pTerm = Abc_NtkCreateNodeConst1(pNtk);
    } else {
        pTerm = Abc_NtkCreateNodeConst0(pNtk);
    }
    pNet  = Abc_NtkFindNet(pNtk, pName);    assert((pNet != nil), "pNet");
    Abc_ObjAddFanin( pNet, pTerm );
    return pTerm;
}

/**Function*************************************************************

  Synopsis    [Create an inverter or buffer for the given net.]

  Description [Assumes that the nets already exist.]
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Io_ReadCreateInv(pNtk *Abc_Ntk_t, pNameIn string, pNameOut string) *Abc_Obj_t {
    var pNet, pNode *Abc_Obj_t;
    pNet  = Abc_NtkFindNet(pNtk, pNameIn);     assert((pNet != nil), "pNet");
    pNode = Abc_NtkCreateNodeInv(pNtk, pNet);
    pNet  = Abc_NtkFindNet(pNtk, pNameOut);    assert((pNet != nil), "pNet");
    Abc_ObjAddFanin( pNet, pNode );
    return pNode;
}

/**Function*************************************************************

  Synopsis    [Create an inverter or buffer for the given net.]

  Description [Assumes that the nets already exist.]
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Io_ReadCreateBuf(pNtk *Abc_Ntk_t, pNameIn string, pNameOut string) *Abc_Obj_t {
    var pNet, pNode *Abc_Obj_t;
    pNet  = Abc_NtkFindNet(pNtk, pNameIn);     assert((pNet != nil), "pNet");
    pNode = Abc_NtkCreateNodeBuf(pNtk, pNet);
    pNet  = Abc_NtkFindNet(pNtk, pNameOut);    assert((pNet != nil), "pNet");
    Abc_ObjAddFanin( pNet, pNode );
    return pNet;
}


/**Function*************************************************************

  Synopsis    [Provide an fopen replacement with path lookup]

  Description [Provide an fopen replacement where the path stored
               in pathvar MVSIS variable is used to look up the path
               for name. Returns NULL if file cannot be opened.]
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Io_FileOpen(FileName string, PathVar string, Mode string, fVerbose int32) *os.File {
    var t, c, i string;

    if PathVar == 0 {
        return fopen( FileName, Mode );
    
    } else {
        if c = Abc_FrameReadFlag( (char*)PathVar ) {
            var ActualFileName [4096]byte;
            var fp *os.File;
            t = Extra_UtilStrsav( c );
            for i = strtok( t, ":" ); i != 0; i = strtok( 0, ":")  {
#ifdef WIN32
                _snprintf ( ActualFileName, 4096, "%s/%s", i, FileName );
#else
                snprintf ( ActualFileName, 4096, "%s/%s", i, FileName );
#endif
                if fp = fopen ( ActualFileName, Mode ) {
                    if fVerbose != 0 {
                    fprintf ( os.Stdout, "Using file %s\n", ActualFileName ); }
                    t = nil;
                    return fp;
                }
            }
            t = nil;
            return 0;
        
        } else {
            return fopen( FileName, Mode );
        }
    }
}

/**Function*************************************************************

  Synopsis    [Tranform SF into PLA.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Io_TransformSF2PLA(pNameIn string, pNameOut string) {
    var fStart int32;
    var Size int32 = 1000000;
    var pBuffer, pToken string;
    var pFileIn *os.File   = os.Open(pNameIn);
    var pFileOut *os.File  = os.OpenFile(pNameOut, os.O_WRONLY | os.O_CREATE | os.O_TRUNC, 0666);
    if pFileIn == nil {
        if pFileOut != nil { _ = (pFileOut).Close(); }
        fmt.Printf("Cannot open file \"%s\" for reading.\n", pNameIn);
        return;
    }
    if pFileOut == nil {
        if pFileIn != nil {  _ = (pFileIn).Close(); }
        fmt.Printf("Cannot open file \"%s\" for reading.\n", pNameOut);
        return;
    }
    // inlining of macro ABC_ALLOC
    pBuffer = ((char *) malloc(sizeof(char) * (size_t)(Size)));
    fmt.Fprintf(pFileOut, ".type fd\n");
    for  fgets(pBuffer, Size, pFileIn)  {
        if strstr(pBuffer, "END_SDF") {
            break; }
        if strstr(pBuffer, "SDF") {
            var pRes string  = fgets(pBuffer, Size, pFileIn);
            assert(pRes != nil, "pRes != nil");
            if pToken = strtok( pBuffer, " \t\r\n" ) {
                fmt.Fprintf(pFileOut, ".i %d\n", strconv.Atoi(pToken)); }
            if pToken = strtok( nil, " \t\r\n" ) {
                fmt.Fprintf(pFileOut, ".o %d\n", strconv.Atoi(pToken)); }
            if pToken = strtok( nil, " \t\r\n" ) {
                fmt.Fprintf(pFileOut, ".p %d\n", strconv.Atoi(pToken)); }
            fStart = 1;
        
        } else if fStart != 0 {
            fmt.Fprintf(pFileOut, "%s", pBuffer); }
    }
    fmt.Fprintf(pFileOut, ".e\n");
    _ = (pFileIn).Close();
    _ = (pFileOut).Close();
    pBuffer = nil;
}

////////////////////////////////////////////////////////////////////////
///                       END OF FILE                                ///
////////////////////////////////////////////////////////////////////////


//ABC_NAMESPACE_IMPL_END

