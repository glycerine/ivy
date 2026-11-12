package io

import (
	"fmt"
	"strconv"
	"strings"
)

import "os"

/**CFile****************************************************************

  FileName    [io.c]

  SystemName  [ABC: Logic synthesis and verification system.]

  PackageName [Command processing package.]

  Synopsis    [Command file.]

  Author      [Alan Mishchenko]
  
  Affiliation [UC Berkeley]

  Date        [Ver. 1.0. Started - June 20, 2005.]

  Revision    [$Id: io.c,v 1.00 2005/06/20 00:00:00 alanmi Exp $]

***********************************************************************/

//#include "ioAbc.h"
//#include "base/main/mainInt.h"
//#include "aig/saig/saig.h"
//#include "proof/abs/abs.h"
//#include "sat/bmc/bmc.h"

#ifdef WIN32
#include <process.h> 
#define unlink _unlink
#else
#include <unistd.h>
#endif

//ABC_NAMESPACE_IMPL_START

////////////////////////////////////////////////////////////////////////
///                        DECLARATIONS                              ///
////////////////////////////////////////////////////////////////////////

// static int IoCommandRead        ( Abc_Frame_t * pAbc, int argc, char **argv );

// static int IoCommandReadAiger   ( Abc_Frame_t * pAbc, int argc, char **argv );

// static int IoCommandReadBaf     ( Abc_Frame_t * pAbc, int argc, char **argv );

// static int IoCommandReadBblif   ( Abc_Frame_t * pAbc, int argc, char **argv );

// static int IoCommandReadBlif    ( Abc_Frame_t * pAbc, int argc, char **argv );

// static int IoCommandReadBlifMv  ( Abc_Frame_t * pAbc, int argc, char **argv );

// static int IoCommandReadBench   ( Abc_Frame_t * pAbc, int argc, char **argv );

// static int IoCommandReadDsd     ( Abc_Frame_t * pAbc, int argc, char **argv );

// static int IoCommandReadEdif    ( Abc_Frame_t * pAbc, int argc, char **argv );

// static int IoCommandReadEqn     ( Abc_Frame_t * pAbc, int argc, char **argv );

// static int IoCommandReadFins    ( Abc_Frame_t * pAbc, int argc, char **argv );

// static int IoCommandReadInit    ( Abc_Frame_t * pAbc, int argc, char **argv );

// static int IoCommandReadPla     ( Abc_Frame_t * pAbc, int argc, char **argv );

// static int IoCommandReadPlaMo   ( Abc_Frame_t * pAbc, int argc, char **argv );

// static int IoCommandReadTruth   ( Abc_Frame_t * pAbc, int argc, char **argv );

// static int IoCommandReadVerilog ( Abc_Frame_t * pAbc, int argc, char **argv );

// static int IoCommandReadStatus  ( Abc_Frame_t * pAbc, int argc, char **argv );

// static int IoCommandReadGig     ( Abc_Frame_t * pAbc, int argc, char **argv );

// static int IoCommandReadJson    ( Abc_Frame_t * pAbc, int argc, char **argv );

// static int IoCommandReadSF      ( Abc_Frame_t * pAbc, int argc, char **argv );


// static int IoCommandWrite       ( Abc_Frame_t * pAbc, int argc, char **argv );

// static int IoCommandWriteHie    ( Abc_Frame_t * pAbc, int argc, char **argv );

// static int IoCommandWriteAiger  ( Abc_Frame_t * pAbc, int argc, char **argv );

// static int IoCommandWriteAigerCex( Abc_Frame_t * pAbc, int argc, char **argv );

// static int IoCommandWriteBaf    ( Abc_Frame_t * pAbc, int argc, char **argv );

// static int IoCommandWriteBblif  ( Abc_Frame_t * pAbc, int argc, char **argv );

// static int IoCommandWriteBlif   ( Abc_Frame_t * pAbc, int argc, char **argv );

// static int IoCommandWriteBlifMv ( Abc_Frame_t * pAbc, int argc, char **argv );

// static int IoCommandWriteBench  ( Abc_Frame_t * pAbc, int argc, char **argv );

// static int IoCommandWriteBook   ( Abc_Frame_t * pAbc, int argc, char **argv );

// static int IoCommandWriteCellNet( Abc_Frame_t * pAbc, int argc, char **argv );

// static int IoCommandWriteCnf    ( Abc_Frame_t * pAbc, int argc, char **argv );

// static int IoCommandWriteCnf2   ( Abc_Frame_t * pAbc, int argc, char **argv );

// static int IoCommandWriteCex    ( Abc_Frame_t * pAbc, int argc, char **argv );

// static int IoCommandWriteDot    ( Abc_Frame_t * pAbc, int argc, char **argv );

// static int IoCommandWriteEqn    ( Abc_Frame_t * pAbc, int argc, char **argv );

// static int IoCommandWriteGml    ( Abc_Frame_t * pAbc, int argc, char **argv );

// static int IoCommandWriteList   ( Abc_Frame_t * pAbc, int argc, char **argv );

// static int IoCommandWritePla    ( Abc_Frame_t * pAbc, int argc, char **argv );

// static int IoCommandWriteVerilog( Abc_Frame_t * pAbc, int argc, char **argv );

// static int IoCommandWriteSortCnf( Abc_Frame_t * pAbc, int argc, char **argv );

// static int IoCommandWriteTruth  ( Abc_Frame_t * pAbc, int argc, char **argv );

// static int IoCommandWriteTruths ( Abc_Frame_t * pAbc, int argc, char **argv );

// static int IoCommandWriteStatus ( Abc_Frame_t * pAbc, int argc, char **argv );

// static int IoCommandWriteSmv    ( Abc_Frame_t * pAbc, int argc, char **argv );

// static int IoCommandWriteJson   ( Abc_Frame_t * pAbc, int argc, char **argv );


// extern void Abc_FrameCopyLTLDataBase( Abc_Frame_t *pAbc, Abc_Ntk_t * pNtk );


var glo_fMapped int32;

////////////////////////////////////////////////////////////////////////
///                     FUNCTION DEFINITIONS                         ///
////////////////////////////////////////////////////////////////////////

/**Function*************************************************************

  Synopsis    []

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Io_Init(pAbc *Abc_Frame_t) {
    Cmd_CommandAdd( pAbc, "I/O", "read",          IoCommandRead,         1 );
    Cmd_CommandAdd( pAbc, "I/O", "read_aiger",    IoCommandReadAiger,    1 );
    Cmd_CommandAdd( pAbc, "I/O", "read_baf",      IoCommandReadBaf,      1 );
    Cmd_CommandAdd( pAbc, "I/O", "read_bblif",    IoCommandReadBblif,    1 );
    Cmd_CommandAdd( pAbc, "I/O", "read_blif",     IoCommandReadBlif,     1 );
    Cmd_CommandAdd( pAbc, "I/O", "read_blif_mv",  IoCommandReadBlifMv,   1 );
    Cmd_CommandAdd( pAbc, "I/O", "read_bench",    IoCommandReadBench,    1 );
    Cmd_CommandAdd( pAbc, "I/O", "read_dsd",      IoCommandReadDsd,      1 );
    Cmd_CommandAdd( pAbc, "I/O", "read_formula",  IoCommandReadDsd,      1 );
//    Cmd_CommandAdd( pAbc, "I/O", "read_edif",     IoCommandReadEdif,     1 );
    Cmd_CommandAdd( pAbc, "I/O", "read_eqn",      IoCommandReadEqn,      1 );
    Cmd_CommandAdd( pAbc, "I/O", "read_fins",     IoCommandReadFins,     0 );
    Cmd_CommandAdd( pAbc, "I/O", "read_init",     IoCommandReadInit,     1 );
    Cmd_CommandAdd( pAbc, "I/O", "read_pla",      IoCommandReadPla,      1 );
    Cmd_CommandAdd( pAbc, "I/O", "read_plamo",    IoCommandReadPlaMo,    1 );
    Cmd_CommandAdd( pAbc, "I/O", "read_truth",    IoCommandReadTruth,    1 );
    Cmd_CommandAdd( pAbc, "I/O", "read_verilog",  IoCommandReadVerilog,  1 );
    Cmd_CommandAdd( pAbc, "I/O", "read_status",   IoCommandReadStatus,   0 );
    Cmd_CommandAdd( pAbc, "I/O", "&read_gig",     IoCommandReadGig,      0 );
    Cmd_CommandAdd( pAbc, "I/O", "read_json",     IoCommandReadJson,     0 );
    Cmd_CommandAdd( pAbc, "I/O", "read_sf",       IoCommandReadSF,       0 );

    Cmd_CommandAdd( pAbc, "I/O", "write",         IoCommandWrite,        0 );
    Cmd_CommandAdd( pAbc, "I/O", "write_hie",     IoCommandWriteHie,     0 );
    Cmd_CommandAdd( pAbc, "I/O", "write_aiger",   IoCommandWriteAiger,   0 );
    Cmd_CommandAdd( pAbc, "I/O", "write_aiger_cex",   IoCommandWriteAigerCex,   0 );
    Cmd_CommandAdd( pAbc, "I/O", "write_baf",     IoCommandWriteBaf,     0 );
    Cmd_CommandAdd( pAbc, "I/O", "write_bblif",   IoCommandWriteBblif,   0 );
    Cmd_CommandAdd( pAbc, "I/O", "write_blif",    IoCommandWriteBlif,    0 );
    Cmd_CommandAdd( pAbc, "I/O", "write_blif_mv", IoCommandWriteBlifMv,  0 );
    Cmd_CommandAdd( pAbc, "I/O", "write_bench",   IoCommandWriteBench,   0 );
    Cmd_CommandAdd( pAbc, "I/O", "write_book",    IoCommandWriteBook,    0 );
    Cmd_CommandAdd( pAbc, "I/O", "write_cellnet", IoCommandWriteCellNet, 0 );
    Cmd_CommandAdd( pAbc, "I/O", "write_cex",     IoCommandWriteCex,     0 );
    Cmd_CommandAdd( pAbc, "I/O", "write_cnf",     IoCommandWriteCnf,     0 );
    Cmd_CommandAdd( pAbc, "I/O", "&write_cnf",    IoCommandWriteCnf2,    0 );
    Cmd_CommandAdd( pAbc, "I/O", "write_dot",     IoCommandWriteDot,     0 );
    Cmd_CommandAdd( pAbc, "I/O", "write_eqn",     IoCommandWriteEqn,     0 );
    Cmd_CommandAdd( pAbc, "I/O", "write_gml",     IoCommandWriteGml,     0 );
//    Cmd_CommandAdd( pAbc, "I/O", "write_list",    IoCommandWriteList,    0 );
    Cmd_CommandAdd( pAbc, "I/O", "write_pla",     IoCommandWritePla,     0 );
    Cmd_CommandAdd( pAbc, "I/O", "write_verilog", IoCommandWriteVerilog, 0 );
//    Cmd_CommandAdd( pAbc, "I/O", "write_verlib",  IoCommandWriteVerLib,  0 );
    Cmd_CommandAdd( pAbc, "I/O", "write_sorter_cnf", IoCommandWriteSortCnf,  0 );
    Cmd_CommandAdd( pAbc, "I/O", "write_truth",   IoCommandWriteTruth,   0 );
    Cmd_CommandAdd( pAbc, "I/O", "&write_truths", IoCommandWriteTruths,  0 );
    Cmd_CommandAdd( pAbc, "I/O", "write_status",  IoCommandWriteStatus,  0 );
    Cmd_CommandAdd( pAbc, "I/O", "write_smv",     IoCommandWriteSmv,     0 );
    Cmd_CommandAdd( pAbc, "I/O", "write_json",    IoCommandWriteJson,    0 );
}

func IoCommandDispatchById(Id Cmd_CommandId_t, pAbc *Abc_Frame_t, argc int32, argv *string, pResult *int32) int32 {
    switch  Id  {
    case CMD_COMMAND_READ_BLIF:
        *pResult = IoCommandReadBlif( pAbc, argc, argv );
        return 1;
    case CMD_COMMAND_READ_AIGER:
        *pResult = IoCommandReadAiger( pAbc, argc, argv );
        return 1;
    case CMD_COMMAND_WRITE_AIGER_CEX:
        *pResult = IoCommandWriteAigerCex( pAbc, argc, argv );
        return 1;
    default:
        return 0;
    }
}

/**Function*************************************************************

  Synopsis    []

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Io_End(pAbc *Abc_Frame_t) {
}

/**Function*************************************************************

  Synopsis    []

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/


/**Function*************************************************************

  Synopsis    []

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func IoCommandReadAiger(pAbc *Abc_Frame_t, argc int32, argv *string) int32 {
    var pNtk *Abc_Ntk_t;
    var pFileName string;
    var fCheck int32;
    var c int32;

    fCheck = 1;
    Extra_UtilGetoptReset();
    for  ( c = Extra_UtilGetopt( argc, argv, "ch" ) ) != -1  {
        switch  c  {
            case 'c':
                fCheck ^= 1;
                break;
            case 'h':
                goto usage;
            default:
                goto usage;
        }
    }
    if argc != globalUtilOptind + 1 {
        goto usage; }
    // get the input file name
    pFileName = argv[globalUtilOptind];
    // read the file using the corresponding file reader
    pNtk = Io_Read( pFileName, IO_FILE_AIGER, fCheck, 0 );
    if pNtk == nil {
        return 1; }
    // replace the current network
    Abc_FrameReplaceCurrentNetwork( pAbc, pNtk );
    Abc_FrameClearVerifStatus( pAbc );
    return 0;

usage:
    fmt.Fprintf(pAbc.Err, "usage: read_aiger [-ch] <file>\n");
    fmt.Fprintf(pAbc.Err, "\t         reads the network in the AIGER format (http://fmv.jku.at/aiger)\n");
    fmt.Fprintf(pAbc.Err, "\t-c     : toggle network check after reading [default = %s]\n", fCheck? "yes":"no");
    fmt.Fprintf(pAbc.Err, "\t-h     : prints the command summary\n");
    fmt.Fprintf(pAbc.Err, "\tfile   : the name of a file to read\n");
    return 1;
}

/**Function*************************************************************

  Synopsis    []

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func IoCommandReadBaf(pAbc *Abc_Frame_t, argc int32, argv *string) int32 {
    var pNtk *Abc_Ntk_t;
    var pFileName string;
    var fCheck int32;
    var c int32;

    fCheck = 1;
    Extra_UtilGetoptReset();
    for  ( c = Extra_UtilGetopt( argc, argv, "ch" ) ) != -1  {
        switch  c  {
            case 'c':
                fCheck ^= 1;
                break;
            case 'h':
                goto usage;
            default:
                goto usage;
        }
    }
    if argc != globalUtilOptind + 1 {
        goto usage; }
    // get the input file name
    pFileName = argv[globalUtilOptind];
    // read the file using the corresponding file reader
    pNtk = Io_Read( pFileName, IO_FILE_BAF, fCheck, 0 );
    if pNtk == nil {
        return 1; }
    // replace the current network
    Abc_FrameReplaceCurrentNetwork( pAbc, pNtk );
    Abc_FrameClearVerifStatus( pAbc );
    return 0;

usage:
    fmt.Fprintf(pAbc.Err, "usage: read_baf [-ch] <file>\n");
    fmt.Fprintf(pAbc.Err, "\t         reads the network in Binary Aig Format (BAF)\n");
    fmt.Fprintf(pAbc.Err, "\t-c     : toggle network check after reading [default = %s]\n", fCheck? "yes":"no");
    fmt.Fprintf(pAbc.Err, "\t-h     : prints the command summary\n");
    fmt.Fprintf(pAbc.Err, "\tfile   : the name of a file to read\n");
    return 1;
}

/**Function*************************************************************

  Synopsis    []

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func IoCommandReadBblif(pAbc *Abc_Frame_t, argc int32, argv *string) int32 {
    var pNtk *Abc_Ntk_t;
    var pFileName string;
    var fCheck int32;
    var c int32;

    fCheck = 1;
    Extra_UtilGetoptReset();
    for  ( c = Extra_UtilGetopt( argc, argv, "ch" ) ) != -1  {
        switch  c  {
            case 'c':
                fCheck ^= 1;
                break;
            case 'h':
                goto usage;
            default:
                goto usage;
        }
    }
    if argc != globalUtilOptind + 1 {
        goto usage; }
    // get the input file name
    pFileName = argv[globalUtilOptind];
    // read the file using the corresponding file reader
    pNtk = Io_Read( pFileName, IO_FILE_BBLIF, fCheck, 0 );
    if pNtk == nil {
        return 1; }
    // replace the current network
    Abc_FrameReplaceCurrentNetwork( pAbc, pNtk );
    Abc_FrameClearVerifStatus( pAbc );
    return 0;

usage:
    fmt.Fprintf(pAbc.Err, "usage: read_bblif [-ch] <file>\n");
    fmt.Fprintf(pAbc.Err, "\t         reads the network in a binary BLIF format\n");
    fmt.Fprintf(pAbc.Err, "\t-c     : toggle network check after reading [default = %s]\n", fCheck? "yes":"no");
    fmt.Fprintf(pAbc.Err, "\t-h     : prints the command summary\n");
    fmt.Fprintf(pAbc.Err, "\tfile   : the name of a file to read\n");
    return 1;
}

/**Function*************************************************************

  Synopsis    []

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func IoCommandReadBlif(pAbc *Abc_Frame_t, argc int32, argv *string) int32 {
    var pNtk *Abc_Ntk_t;
    var pFileName string;
    var fReadAsAig int32;
    var fCheck int32;
    var fUseNewParser int32;
    var fSaveNames int32;
    var c int32;
    extern Abc_Ntk_t * Io_ReadBlifAsAig( char * pFileName, int fCheck );

    fCheck = 1;
    fReadAsAig = 0;
    fUseNewParser = 1;
    fSaveNames = 0;
    Extra_UtilGetoptReset();
    for  ( c = Extra_UtilGetopt( argc, argv, "nmach" ) ) != -1  {
        switch  c  {
            case 'n':
                fUseNewParser ^= 1;
                break;
            case 'm':
                fSaveNames ^= 1;
                break;
            case 'a':
                fReadAsAig ^= 1;
                break;
            case 'c':
                fCheck ^= 1;
                break;
            case 'h':
                goto usage;
            default:
                goto usage;
        }
    }
    if argc != globalUtilOptind + 1 {
        goto usage; }
    // get the input file name
    pFileName = argv[globalUtilOptind];
    // read the file using the corresponding file reader
    if fReadAsAig != 0 {
        pNtk = Io_ReadBlifAsAig( pFileName, fCheck );
    } else if fUseNewParser != 0 {
        pNtk = Io_Read( pFileName, IO_FILE_BLIF, fCheck, 0 );
    } else {
        var pTemp *Abc_Ntk_t;
        pNtk = Io_ReadBlif( pFileName, fCheck );
        if pNtk == nil {
            return 1; }
        if fSaveNames != 0 {
            Abc_NtkStartNameIds( pNtk ); }
        pNtk = Abc_NtkToLogic( pTemp = pNtk );
        if fSaveNames != 0 {
            Abc_NtkTransferNameIds( pTemp, pNtk ); }
        Abc_NtkDelete( pTemp );
    }

    if pNtk == nil {
        return 1; }
    // replace the current network
    Abc_FrameReplaceCurrentNetwork( pAbc, pNtk );
    Abc_FrameClearVerifStatus( pAbc );
    return 0;

usage:
    fmt.Fprintf(pAbc.Err, "usage: read_blif [-nmach] <file>\n");
    fmt.Fprintf(pAbc.Err, "\t         reads the network in binary BLIF format\n");
    fmt.Fprintf(pAbc.Err, "\t         (if this command does not work, try \"read\")\n");
    fmt.Fprintf(pAbc.Err, "\t-n     : toggle using old BLIF parser without hierarchy support [default = %s]\n", !fUseNewParser? "yes":"no");
    fmt.Fprintf(pAbc.Err, "\t-m     : toggle saving original circuit names into a file [default = %s]\n", fSaveNames? "yes":"no");
    fmt.Fprintf(pAbc.Err, "\t-a     : toggle creating AIG while reading the file [default = %s]\n", fReadAsAig? "yes":"no");
    fmt.Fprintf(pAbc.Err, "\t-c     : toggle network check after reading [default = %s]\n", fCheck? "yes":"no");
    fmt.Fprintf(pAbc.Err, "\t-h     : prints the command summary\n");
    fmt.Fprintf(pAbc.Err, "\tfile   : the name of a file to read\n");
    return 1;
}

/**Function*************************************************************

  Synopsis    []

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func IoCommandReadBlifMv(pAbc *Abc_Frame_t, argc int32, argv *string) int32 {
    var pNtk *Abc_Ntk_t;
    var pFileName string;
    var fCheck int32;
    var c int32;

    fCheck = 1;
    Extra_UtilGetoptReset();
    for  ( c = Extra_UtilGetopt( argc, argv, "ch" ) ) != -1  {
        switch  c  {
            case 'c':
                fCheck ^= 1;
                break;
            case 'h':
                goto usage;
            default:
                goto usage;
        }
    }
    if argc != globalUtilOptind + 1 {
        goto usage; }
    // get the input file name
    pFileName = argv[globalUtilOptind];
    // read the file using the corresponding file reader
    pNtk = Io_Read( pFileName, IO_FILE_BLIFMV, fCheck, 0 );
    if pNtk == nil {
        return 1; }
    // replace the current network
    Abc_FrameReplaceCurrentNetwork( pAbc, pNtk );
    Abc_FrameClearVerifStatus( pAbc );
    return 0;

usage:
    fmt.Fprintf(pAbc.Err, "usage: read_blif_mv [-ch] <file>\n");
    fmt.Fprintf(pAbc.Err, "\t         reads the network in BLIF-MV format\n");
    fmt.Fprintf(pAbc.Err, "\t         (if this command does not work, try \"read\")\n");
    fmt.Fprintf(pAbc.Err, "\t-c     : toggle network check after reading [default = %s]\n", fCheck? "yes":"no");
    fmt.Fprintf(pAbc.Err, "\t-h     : prints the command summary\n");
    fmt.Fprintf(pAbc.Err, "\tfile   : the name of a file to read\n");
    return 1;
}

/**Function*************************************************************

  Synopsis    []

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func IoCommandReadBench(pAbc *Abc_Frame_t, argc int32, argv *string) int32 {
    var pNtk *Abc_Ntk_t;
    var pFileName string;
    var fCheck int32;
    var c int32;

    fCheck = 1;
    Extra_UtilGetoptReset();
    for  ( c = Extra_UtilGetopt( argc, argv, "ch" ) ) != -1  {
        switch  c  {
            case 'c':
                fCheck ^= 1;
                break;
            case 'h':
                goto usage;
            default:
                goto usage;
        }
    }
    if argc != globalUtilOptind + 1 {
        goto usage; }
    // get the input file name
    pFileName = argv[globalUtilOptind];
    // read the file using the corresponding file reader
    pNtk = Io_Read( pFileName, IO_FILE_BENCH, fCheck, 0 );
    if pNtk == nil {
        return 1; }
    // replace the current network
    Abc_FrameReplaceCurrentNetwork( pAbc, pNtk );
    Abc_FrameClearVerifStatus( pAbc );
    return 0;

usage:
    fmt.Fprintf(pAbc.Err, "usage: read_bench [-ch] <file>\n");
    fmt.Fprintf(pAbc.Err, "\t         reads the network in BENCH format\n");
    fmt.Fprintf(pAbc.Err, "\t-c     : toggle network check after reading [default = %s]\n", fCheck? "yes":"no");
    fmt.Fprintf(pAbc.Err, "\t-h     : prints the command summary\n");
    fmt.Fprintf(pAbc.Err, "\tfile   : the name of a file to read\n");
    return 1;
}

/**Function*************************************************************

  Synopsis    []

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func IoCommandReadDsd(pAbc *Abc_Frame_t, argc int32, argv *string) int32 {
    var pNtk *Abc_Ntk_t;
    var pString string;
    var fCheck int32;
    var c int32;
    extern Abc_Ntk_t * Io_ReadDsd( char * pFormula );

    fCheck = 1;
    Extra_UtilGetoptReset();
    for  ( c = Extra_UtilGetopt( argc, argv, "ch" ) ) != -1  {
        switch  c  {
            case 'c':
                fCheck ^= 1;
                break;
            case 'h':
                goto usage;
            default:
                goto usage;
        }
    }
    if argc != globalUtilOptind + 1 {
        goto usage; }
    // get the input file name
    pString = argv[globalUtilOptind];
    // read the file using the corresponding file reader
    pNtk = Io_ReadDsd( pString );
    if pNtk == nil {
        return 1; }
    // replace the current network
    Abc_FrameReplaceCurrentNetwork( pAbc, pNtk );
    Abc_FrameClearVerifStatus( pAbc );
    return 0;

usage:
    fmt.Fprintf(pAbc.Err, "usage: read_dsd [-h] <formula>\n");
    fmt.Fprintf(pAbc.Err, "\t          parses a formula representing DSD of a function\n");
    fmt.Fprintf(pAbc.Err, "\t-h      : prints the command summary\n");
    fmt.Fprintf(pAbc.Err, "\tformula : the formula representing disjoint-support decomposition (DSD)\n");
    fmt.Fprintf(pAbc.Err, "\t          Example of a formula: !(a*(b+CA(!d,e*f,c))*79B3(g,h,i,k))\n");
    fmt.Fprintf(pAbc.Err, "\t          where \'!\' is an INV, \'*\' is an AND, \'+\' is an XOR, \n");
    fmt.Fprintf(pAbc.Err, "\t          CA and 79B3 are hexadecimal representations of truth tables\n");
    fmt.Fprintf(pAbc.Err, "\t          (in this case CA=11001010 is truth table of MUX(Data0,Data1,Ctrl))\n");
    fmt.Fprintf(pAbc.Err, "\t          The lower chars (a,b,c,etc) are reserved for elementary variables.\n");
    fmt.Fprintf(pAbc.Err, "\t          The upper chars (A,B,C,etc) are reserved for hexadecimal digits.\n");
    fmt.Fprintf(pAbc.Err, "\t          No spaces are allowed in formulas. In parentheses, LSB goes first.\n");
    return 1;
}

/**Function*************************************************************

  Synopsis    []

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func IoCommandReadEdif(pAbc *Abc_Frame_t, argc int32, argv *string) int32 {
    var pNtk *Abc_Ntk_t;
    var pFileName string;
    var fCheck int32;
    var c int32;

    fCheck = 1;
    Extra_UtilGetoptReset();
    for  ( c = Extra_UtilGetopt( argc, argv, "ch" ) ) != -1  {
        switch  c  {
            case 'c':
                fCheck ^= 1;
                break;
            case 'h':
                goto usage;
            default:
                goto usage;
        }
    }
    if argc != globalUtilOptind + 1 {
        goto usage; }
    // get the input file name
    pFileName = argv[globalUtilOptind];
    // read the file using the corresponding file reader
    pNtk = Io_Read( pFileName, IO_FILE_EDIF, fCheck, 0 );
    if pNtk == nil {
        return 1; }
    // replace the current network
    Abc_FrameReplaceCurrentNetwork( pAbc, pNtk );
    Abc_FrameClearVerifStatus( pAbc );
    return 0;

usage:
    fmt.Fprintf(pAbc.Err, "usage: read_edif [-ch] <file>\n");
    fmt.Fprintf(pAbc.Err, "\t         reads the network in EDIF (works only for ISCAS benchmarks)\n");
    fmt.Fprintf(pAbc.Err, "\t-c     : toggle network check after reading [default = %s]\n", fCheck? "yes":"no");
    fmt.Fprintf(pAbc.Err, "\t-h     : prints the command summary\n");
    fmt.Fprintf(pAbc.Err, "\tfile   : the name of a file to read\n");
    return 1;
}

/**Function*************************************************************

  Synopsis    []

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func IoCommandReadEqn(pAbc *Abc_Frame_t, argc int32, argv *string) int32 {
    var pNtk *Abc_Ntk_t;
    var pFileName string;
    var fCheck int32;
    var c int32;

    fCheck = 1;
    Extra_UtilGetoptReset();
    for  ( c = Extra_UtilGetopt( argc, argv, "ch" ) ) != -1  {
        switch  c  {
            case 'c':
                fCheck ^= 1;
                break;
            case 'h':
                goto usage;
            default:
                goto usage;
        }
    }
    if argc != globalUtilOptind + 1 {
        goto usage; }
    // get the input file name
    pFileName = argv[globalUtilOptind];
    // read the file using the corresponding file reader
    pNtk = Io_Read( pFileName, IO_FILE_EQN, fCheck, 0 );
    if pNtk == nil {
        return 1; }
    // replace the current network
    Abc_FrameReplaceCurrentNetwork( pAbc, pNtk );
    Abc_FrameClearVerifStatus( pAbc );
    return 0;

usage:
    fmt.Fprintf(pAbc.Err, "usage: read_eqn [-ch] <file>\n");
    fmt.Fprintf(pAbc.Err, "\t         reads the network in equation format\n");
    fmt.Fprintf(pAbc.Err, "\t-c     : toggle network check after reading [default = %s]\n", fCheck? "yes":"no");
    fmt.Fprintf(pAbc.Err, "\t-h     : prints the command summary\n");
    fmt.Fprintf(pAbc.Err, "\tfile   : the name of a file to read\n");
    return 1;
}

/**Function*************************************************************

  Synopsis    []

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func IoCommandReadFins(pAbc *Abc_Frame_t, argc int32, argv *string) int32 {
    extern Vec_Int_t * Io_ReadFins( Abc_Ntk_t * pNtk, char * pFileName, int fVerbose );
    var pNtk *Abc_Ntk_t  = Abc_FrameReadNtk(pAbc);
    var pFileName string;
    var c, fVerbose int32;

    Extra_UtilGetoptReset();
    for  ( c = Extra_UtilGetopt( argc, argv, "vh" ) ) != -1  {
        switch  c  {
            case 'v':
                fVerbose ^= 1;
                break;
            case 'h':
                goto usage;
            default:
                goto usage;
        }
    }
    if argc != globalUtilOptind + 1 {
        goto usage; }
    // get the input file name
    pFileName = argv[globalUtilOptind];
    // check current network
    if pNtk == nil {
        Abc_Print( -1, "Empty network.\n" );
        return 1;
    }
    // compute information and save it in the network
    Vec_IntFreeP( &pNtk.vFins );
    pNtk.vFins = Io_ReadFins( pNtk, pFileName, fVerbose );
    return 0;

usage:
    fmt.Fprintf(pAbc.Err, "usage: read_fins [-vh] <file>\n");
    fmt.Fprintf(pAbc.Err, "\t         reads the network in equation format\n");
    fmt.Fprintf(pAbc.Err, "\t-v     : enable verbose output [default = %s].\n", fVerbose? "yes": "no");
    fmt.Fprintf(pAbc.Err, "\t-h     : prints the command summary\n");
    fmt.Fprintf(pAbc.Err, "\tfile   : the name of a file to read\n");
    return 1;
}

/**Function*************************************************************

  Synopsis    []

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func IoCommandReadInit(pAbc *Abc_Frame_t, argc int32, argv *string) int32 {
    var pOut, pErr *os.File;
    var pNtk *Abc_Ntk_t;
    var pFileName string;
    var c int32;

    pNtk = Abc_FrameReadNtk(pAbc);
    pOut = Abc_FrameReadOut(pAbc);
    pErr = Abc_FrameReadErr(pAbc);

    Extra_UtilGetoptReset();
    for  ( c = Extra_UtilGetopt( argc, argv, "h" ) ) != -1  {
        switch  c  {
            case 'h':
                goto usage;
            default:
                goto usage;
        }
    }
    if argc != globalUtilOptind && argc != globalUtilOptind + 1 {
        goto usage; }

    if pNtk == nil {
        fmt.Fprintf(pErr, "Empty network.\n");
        return 1;
    }
    // get the input file name
    if argc == globalUtilOptind + 1 {
        pFileName = argv[globalUtilOptind];
    } else if pNtk.pSpec {
        pFileName = Extra_FileNameGenericAppend( pNtk.pSpec, ".init" );
    } else {
        fmt.Printf("File name should be given on the command line.\n");
        return 1;
    }

    // read the file using the corresponding file reader
    pNtk = Abc_NtkDup( pNtk );
    Io_ReadBenchInit( pNtk, pFileName );
    // replace the current network
    Abc_FrameReplaceCurrentNetwork( pAbc, pNtk );
    Abc_FrameClearVerifStatus( pAbc );
    return 0;

usage:
    fmt.Fprintf(pAbc.Err, "usage: read_init [-h] <file>\n");
    fmt.Fprintf(pAbc.Err, "\t         reads initial state of the network in BENCH format\n");
    fmt.Fprintf(pAbc.Err, "\t-h     : prints the command summary\n");
    fmt.Fprintf(pAbc.Err, "\tfile   : the name of a file to read\n");
    return 1;
}

/**Function*************************************************************

  Synopsis    []

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func IoCommandReadPla(pAbc *Abc_Frame_t, argc int32, argv *string) int32 {
    var pNtk *Abc_Ntk_t;
    var pFileName string;
    var c int32;
    var fZeros int32;
    var fBoth int32;
    var fOnDc int32;
    var fSkipPrepro int32;
    var fCheck int32 = 1;

    Extra_UtilGetoptReset();
    for  ( c = Extra_UtilGetopt( argc, argv, "zbdxch" ) ) != -1  {
        switch  c  {
            case 'z':
                fZeros ^= 1;
                break;
            case 'b':
                fBoth ^= 1;
                break;
            case 'd':
                fOnDc ^= 1;
                break;
            case 'x':
                fSkipPrepro ^= 1;
                break;
            case 'c':
                fCheck ^= 1;
                break;
            case 'h':
                goto usage;
            default:
                goto usage;
        }
    }
    if argc != globalUtilOptind + 1 {
        goto usage; }
    // get the input file name
    pFileName = argv[globalUtilOptind];
    // read the file using the corresponding file reader
    if (fZeros != 0) || (fBoth != 0) || (fOnDc != 0) || (fSkipPrepro != 0) {
        var pTemp *Abc_Ntk_t;
        pNtk = Io_ReadPla( pFileName, fZeros, fBoth, fOnDc, fSkipPrepro, fCheck );
        if pNtk == nil {
            fmt.Printf("Reading PLA file has failed.\n");
            return 1;
        }
        pNtk = Abc_NtkToLogic( pTemp = pNtk );
        Abc_NtkDelete( pTemp );
    
    } else {
        pNtk = Io_Read( pFileName, IO_FILE_PLA, fCheck, 0 ); }
    if pNtk == nil {
        return 1; }
    // replace the current network
    Abc_FrameReplaceCurrentNetwork( pAbc, pNtk );
    Abc_FrameClearVerifStatus( pAbc );
    return 0;

usage:
    fmt.Fprintf(pAbc.Err, "usage: read_pla [-zbdxch] <file>\n");
    fmt.Fprintf(pAbc.Err, "\t         reads the network in PLA\n");
    fmt.Fprintf(pAbc.Err, "\t-z     : toggle reading on-set and off-set [default = %s]\n", fZeros? "off-set":"on-set");
    fmt.Fprintf(pAbc.Err, "\t-b     : toggle reading both on-set and off-set as on-set [default = %s]\n", fBoth? "off-set":"on-set");
    fmt.Fprintf(pAbc.Err, "\t-d     : toggle reading both on-set and dc-set as on-set [default = %s]\n", fOnDc? "off-set":"on-set");
    fmt.Fprintf(pAbc.Err, "\t-x     : toggle reading Exclusive SOP rather than SOP [default = %s]\n", fSkipPrepro? "yes":"no");
    fmt.Fprintf(pAbc.Err, "\t-c     : toggle network check after reading [default = %s]\n", fCheck? "yes":"no");
    fmt.Fprintf(pAbc.Err, "\t-h     : prints the command summary\n");
    fmt.Fprintf(pAbc.Err, "\tfile   : the name of a file to read\n");
    return 1;
}

/**Function*************************************************************

  Synopsis    []

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func IoCommandReadPlaMo(pAbc *Abc_Frame_t, argc int32, argv *string) int32 {
    extern Abc_Ntk_t * Mop_ManTest( char * pFileName, int fMerge, int fVerbose );
    var pNtk *Abc_Ntk_t;
    var c, fMerge, fVerbose int32;
    Extra_UtilGetoptReset();
    for  ( c = Extra_UtilGetopt( argc, argv, "mvh" ) ) != -1  {
        switch  c  {
            case 'm':
                fMerge ^= 1;
                break;
            case 'v':
                fVerbose ^= 1;
                break;
            case 'h':
                goto usage;
            default:
                goto usage;
        }
    }
    if argc != globalUtilOptind + 1 {
        goto usage; }
    // get the input file name
    pNtk = Mop_ManTest( argv[globalUtilOptind], fMerge, fVerbose );
    if pNtk == nil {
        return 1; }
    // replace the current network
    Abc_FrameReplaceCurrentNetwork( pAbc, pNtk );
    Abc_FrameClearVerifStatus( pAbc );
    return 0;

usage:
    fmt.Fprintf(pAbc.Err, "usage: read_plamo [-mvh] <file>\n");
    fmt.Fprintf(pAbc.Err, "\t         reads the network in multi-output PLA\n");
    fmt.Fprintf(pAbc.Err, "\t-m     : toggle dist-1 merge for cubes with identical outputs [default = %s]\n", fMerge? "yes":"no");
    fmt.Fprintf(pAbc.Err, "\t-v     : toggle printing verbose information [default = %s]\n", fVerbose? "yes":"no");
    fmt.Fprintf(pAbc.Err, "\t-h     : prints the command summary\n");
    fmt.Fprintf(pAbc.Err, "\tfile   : the name of a file to read\n");
    return 1;
}

/**Function*************************************************************

  Synopsis    []

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func IoCommandReadTruth(pAbc *Abc_Frame_t, argc int32, argv *string) int32 {
    var pNtk *Abc_Ntk_t;
    var pStr string  = nil;
    var pSopCover string;
    var fHex int32   = 1;
    var fFile int32;
    var c int32;

    Extra_UtilGetoptReset();
    for  ( c = Extra_UtilGetopt( argc, argv, "xfh" ) ) != -1  {
        switch  c  {
            case 'x':
                fHex ^= 1;
                break;
            case 'f':
                fFile ^= 1;
                break;
            case 'h':
                goto usage;
            default:
                goto usage;
        }
    }

    if argc != globalUtilOptind + 1 {
        goto usage; }

    if fFile != 0 {
        pStr = Extra_FileReadContents( argv[globalUtilOptind] );
    } else {
        pStr = argv[globalUtilOptind]; }
    for  pStr[ len(pStr) - 1 ] == '\n' || pStr[ len(pStr) - 1 ] == '\r'  {
        pStr[ len(pStr) - 1 ] = '\0';
    }

    // convert truth table to SOP
    if fHex != 0 {
        pSopCover = Abc_SopFromTruthHex(pStr);
    } else {
        pSopCover = Abc_SopFromTruthBin(pStr); }
    if fFile != 0 {
        ABC_FREE( pStr ); }
    if pSopCover == nil || pSopCover[0] == 0 {
        ABC_FREE( pSopCover );
        fmt.Fprintf(pAbc.Err, "Reading truth table has failed.\n");
        return 1;
    }

    pNtk = Abc_NtkCreateWithNode( pSopCover );
    ABC_FREE( pSopCover );
    if pNtk == nil {
        fmt.Fprintf(pAbc.Err, "Deriving the network has failed.\n");
        return 1;
    }
    // replace the current network
    Abc_FrameReplaceCurrentNetwork( pAbc, pNtk );
    Abc_FrameClearVerifStatus( pAbc );
    return 0;

usage:
    fmt.Fprintf(pAbc.Err, "usage: read_truth [-xfh] <truth> <file>\n");
    fmt.Fprintf(pAbc.Err, "\t         creates network with node having given truth table\n");
    fmt.Fprintf(pAbc.Err, "\t-x     : toggles between bin and hex representation [default = %s]\n", fHex?  "hex":"bin");
    fmt.Fprintf(pAbc.Err, "\t-f     : toggles reading truth table from file [default = %s]\n", fFile? "yes": "no");
    fmt.Fprintf(pAbc.Err, "\t-h     : prints the command summary\n");
    fmt.Fprintf(pAbc.Err, "\ttruth  : truth table with most signficant bit first (e.g. 1000 for AND(a,b))\n");
    fmt.Fprintf(pAbc.Err, "\tfile   : file name with the truth table\n");
    return 1;
}

/**Function*************************************************************

  Synopsis    []

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func IoCommandReadVerilog(pAbc *Abc_Frame_t, argc int32, argv *string) int32 {
    var pNtk *Abc_Ntk_t;
    var pFileName string;
    var fCheck, fBarBufs int32;
    var c int32;

    fCheck = 1;
    fBarBufs = 0;
    glo_fMapped = 0;
    Extra_UtilGetoptReset();
    for  ( c = Extra_UtilGetopt( argc, argv, "mcbh" ) ) != -1  {
        switch  c  {
            case 'm':
                glo_fMapped ^= 1;
                break;
            case 'c':
                fCheck ^= 1;
                break;
            case 'b':
                fBarBufs ^= 1;
                break;
            case 'h':
                goto usage;
            default:
                goto usage;
        }
    }
    if argc != globalUtilOptind + 1 {
        goto usage; }
    // get the input file name
    pFileName = argv[globalUtilOptind];
    // read the file using the corresponding file reader
    pNtk = Io_Read( pFileName, IO_FILE_VERILOG, fCheck, fBarBufs );
    if pNtk == nil {
        return 1; }
    // replace the current network
    Abc_FrameReplaceCurrentNetwork( pAbc, pNtk );
    Abc_FrameClearVerifStatus( pAbc );
    return 0;

usage:
    fmt.Fprintf(pAbc.Err, "usage: read_verilog [-mcbh] <file>\n");
    fmt.Fprintf(pAbc.Err, "\t         reads the network in Verilog (IWLS 2002/2005 subset)\n");
    fmt.Fprintf(pAbc.Err, "\t-m     : toggle reading mapped Verilog [default = %s]\n", glo_fMapped? "yes":"no");
    fmt.Fprintf(pAbc.Err, "\t-c     : toggle network check after reading [default = %s]\n", fCheck? "yes":"no");
    fmt.Fprintf(pAbc.Err, "\t-b     : toggle reading barrier buffers [default = %s]\n", fBarBufs? "yes":"no");
    fmt.Fprintf(pAbc.Err, "\t-h     : prints the command summary\n");
    fmt.Fprintf(pAbc.Err, "\tfile   : the name of a file to read\n");
    return 1;
}

/**Function*************************************************************

  Synopsis    []

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func IoCommandReadStatus(pAbc *Abc_Frame_t, argc int32, argv *string) int32 {
    var pFileName string;
    var pFile *os.File;
    var c int32;
    extern int Abc_NtkReadLogFile( char * pFileName, Abc_Cex_t ** ppCex, int * pnFrames );

    Extra_UtilGetoptReset();
    for  ( c = Extra_UtilGetopt( argc, argv, "h" ) ) != -1  {
        switch  c  {
            case 'h':
                goto usage;
            default:
                goto usage;
        }
    }
    if argc != globalUtilOptind + 1 {
        goto usage;
    }

    // get the input file name
    pFileName = argv[globalUtilOptind];
    if (pFile = os.Open(pFileName)) == nil {
        fmt.Fprintf(pAbc.Err, "Cannot open input file \"%s\". \n", pFileName);
        return 1;
    }
    _ = (pFile).Close();

    // set the new network
    Abc_FrameClearVerifStatus( pAbc );
    pAbc.Status = Abc_NtkReadLogFile( pFileName, &pAbc.pCex, &pAbc.nFrames );
    return 0;

usage:
    fmt.Fprintf(pAbc.Err, "usage: read_status [-ch] <file>\n");
    fmt.Fprintf(pAbc.Err, "\t         reads verification log file\n");
    fmt.Fprintf(pAbc.Err, "\t-h     : prints the command summary\n");
    fmt.Fprintf(pAbc.Err, "\tfile   : the name of a file to read\n");
    return 1;
}

/**Function*************************************************************

  Synopsis    []

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func IoCommandReadGig(pAbc *Abc_Frame_t, argc int32, argv *string) int32 {
    extern Gia_Man_t * Gia_ManReadGig( char * pFileName );
    var pAig *Gia_Man_t;
    var pFileName string;
    var pFile *os.File;
    var c int32;

    Extra_UtilGetoptReset();
    for  ( c = Extra_UtilGetopt( argc, argv, "h" ) ) != -1  {
        switch  c  {
            case 'h':
                goto usage;
            default:
                goto usage;
        }
    }
    if argc != globalUtilOptind + 1 {
        goto usage;
    }

    // get the input file name
    pFileName = argv[globalUtilOptind];
    if (pFile = os.Open(pFileName)) == nil {
        fmt.Fprintf(pAbc.Err, "Cannot open input file \"%s\". \n", pFileName);
        return 1;
    }
    _ = (pFile).Close();

    // set the new network
    pAig = Gia_ManReadGig( pFileName );
    Abc_FrameUpdateGia( pAbc, pAig );
    return 0;

usage:
    fmt.Fprintf(pAbc.Err, "usage: &read_gig [-h] <file>\n");
    fmt.Fprintf(pAbc.Err, "\t         reads design in GIG format\n");
    fmt.Fprintf(pAbc.Err, "\t-h     : prints the command summary\n");
    fmt.Fprintf(pAbc.Err, "\tfile   : the name of a file to read\n");
    return 1;
}

/**Function*************************************************************

  Synopsis    []

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func IoCommandReadJson(pAbc *Abc_Frame_t, argc int32, argv *string) int32 {
    extern Vec_Wec_t * Json_Read( char * pFileName, Abc_Nam_t ** ppStrs );
    var vObjs *Vec_Wec_t;
    var pStrs *Abc_Nam_t;
    var pFileName string;
    var pFile *os.File;
    var c int32;

    Extra_UtilGetoptReset();
    for  ( c = Extra_UtilGetopt( argc, argv, "h" ) ) != -1  {
        switch  c  {
            case 'h':
                goto usage;
            default:
                goto usage;
        }
    }
    if argc != globalUtilOptind + 1 {
        goto usage;
    }

    // get the input file name
    pFileName = argv[globalUtilOptind];
    if (pFile = os.Open(pFileName)) == nil {
        fmt.Fprintf(pAbc.Err, "Cannot open input file \"%s\". \n", pFileName);
        return 1;
    }
    _ = (pFile).Close();

    // set the new network
    vObjs = Json_Read( pFileName, &pStrs );
    if vObjs == nil {
        return 0; }
    Abc_FrameSetJsonStrs( pStrs );
    Abc_FrameSetJsonObjs( vObjs );
    return 0;

usage:
    fmt.Fprintf(pAbc.Err, "usage: read_json [-h] <file>\n");
    fmt.Fprintf(pAbc.Err, "\t         reads file in JSON format\n");
    fmt.Fprintf(pAbc.Err, "\t-h     : prints the command summary\n");
    fmt.Fprintf(pAbc.Err, "\tfile   : the name of a file to read\n");
    return 1;
}

/**Function*************************************************************

  Synopsis    []

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func IoCommandReadSF(pAbc *Abc_Frame_t, argc int32, argv *string) int32 {
    extern void Io_TransformSF2PLA( char * pNameIn, char * pNameOut );

    var pNtk *Abc_Ntk_t;
    var pFile *os.File;
    var pFileName string;
    var pFileTemp string = "_temp_sf_.pla";
    var c int32;

    Extra_UtilGetoptReset();
    for  ( c = Extra_UtilGetopt( argc, argv, "h" ) ) != -1  {
        switch  c  {
            case 'h':
                goto usage;
            default:
                goto usage;
        }
    }
    if argc != globalUtilOptind + 1 {
        goto usage;
    }

    // get the input file name
    pFileName = argv[globalUtilOptind];
    if (pFile = os.Open(pFileName)) == nil {
        fmt.Fprintf(pAbc.Err, "Cannot open input file \"%s\". \n", pFileName);
        return 1;
    }
    _ = (pFile).Close();
    Io_TransformSF2PLA( pFileName, pFileTemp );
    pNtk = Io_Read( pFileTemp, IO_FILE_PLA, 1, 0 );
    unlink( pFileTemp );
    if pNtk == nil {
        return 1; }
    ABC_FREE( pNtk.pName );
    pNtk.pName = Extra_FileNameGeneric( pFileName );
    ABC_FREE( pNtk.pSpec );
    pNtk.pSpec = Abc_UtilStrsav( pFileName );
    // replace the current network
    Abc_FrameReplaceCurrentNetwork( pAbc, pNtk );
    Abc_FrameClearVerifStatus( pAbc );

    return 0;

usage:
    fmt.Fprintf(pAbc.Err, "usage: read_sf [-h] <file>\n");
    fmt.Fprintf(pAbc.Err, "\t         reads file in SF format\n");
    fmt.Fprintf(pAbc.Err, "\t-h     : prints the command summary\n");
    fmt.Fprintf(pAbc.Err, "\tfile   : the name of a file to read\n");
    return 1;
}


/**Function*************************************************************

  Synopsis    []

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/


/**Function*************************************************************

  Synopsis    []

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func IoCommandWriteHie(pAbc *Abc_Frame_t, argc int32, argv *string) int32 {
    var pBaseName, pFileName string;
    var c int32;

    glo_fMapped = 0;
    Extra_UtilGetoptReset();
    for  ( c = Extra_UtilGetopt( argc, argv, "mh" ) ) != -1  {
        switch  c  {
            case 'm':
                glo_fMapped ^= 1;
                break;
            case 'h':
                goto usage;
            default:
                goto usage;
        }
    }
    if pAbc.pNtkCur == nil {
        fmt.Fprintf(pAbc.Out, "Empty network.\n");
        return 0;
    }
    if argc != globalUtilOptind + 2 {
        goto usage; }
    // get the output file name
    pBaseName = argv[globalUtilOptind];
    pFileName = argv[globalUtilOptind+1];
    // call the corresponding file writer
//    Io_Write( pAbc->pNtkCur, pFileName, Io_ReadFileType(pFileName) );
    Io_WriteHie( pAbc.pNtkCur, pBaseName, pFileName );
    return 0;

usage:
    fmt.Fprintf(pAbc.Err, "usage: write_hie [-h] <orig> <file>\n");
    fmt.Fprintf(pAbc.Err, "\t         writes the current network into <file> by calling\n");
    fmt.Fprintf(pAbc.Err, "\t         the hierarchical writer that matches the extension of <file>\n");
    fmt.Fprintf(pAbc.Err, "\t-m     : toggle reading mapped Verilog for <orig> [default = %s]\n", glo_fMapped? "yes":"no");
    fmt.Fprintf(pAbc.Err, "\t-h     : print the help massage\n");
    fmt.Fprintf(pAbc.Err, "\torig   : the name of the original file with the hierarchical design\n");
    fmt.Fprintf(pAbc.Err, "\tfile   : the name of the file to write\n");
    return 1;
}

/**Function*************************************************************

  Synopsis    []

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func IoCommandWriteAiger(pAbc *Abc_Frame_t, argc int32, argv *string) int32 {
    var pFileName string;
    var fWriteSymbols int32;
    var fCompact int32;
    var fUnique int32;
    var fVerbose int32;
    var c int32;

    fWriteSymbols = 0;
    fCompact      = 0;
    fUnique       = 0;
    fVerbose      = 0;
    Extra_UtilGetoptReset();
    for  ( c = Extra_UtilGetopt( argc, argv, "scuvh" ) ) != -1  {
        switch  c  {
            case 's':
                fWriteSymbols ^= 1;
                break;
            case 'c':
                fCompact ^= 1;
                break;
            case 'u':
                fUnique ^= 1;
                break;
            case 'v':
                fVerbose ^= 1;
                break;
            case 'h':
                goto usage;
            default:
                goto usage;
        }
    }
    if pAbc.pNtkCur == nil {
        fmt.Fprintf(pAbc.Out, "Empty network.\n");
        return 0;
    }
    if argc != globalUtilOptind + 1 {
        goto usage; }
    // get the output file name
    pFileName = argv[globalUtilOptind];
    // call the corresponding file writer
    if !Abc_NtkIsStrash(pAbc.pNtkCur) {
        fmt.Fprintf(os.Stdout, "Writing this format is only possible for structurally hashed AIGs.\n");
        return 1;
    }
    if fUnique != 0 {
        extern Aig_Man_t * Abc_NtkToDar( Abc_Ntk_t * pNtk, int fExors, int fRegisters );
        extern Abc_Ntk_t * Abc_NtkFromAigPhase( Aig_Man_t * pMan );
        var pAig *Aig_Man_t  = Abc_NtkToDar( pAbc.pNtkCur, 0, 1 );
        var pCan *Aig_Man_t  = Saig_ManDupIsoCanonical( pAig, fVerbose );
        var pTemp *Abc_Ntk_t  = Abc_NtkFromAigPhase( pCan );
        Aig_ManStop( pCan );
        Aig_ManStop( pAig );
        Io_WriteAiger( pTemp, pFileName, fWriteSymbols, fCompact, fUnique );
        Abc_NtkDelete( pTemp );
    
    } else {
        Io_WriteAiger( pAbc.pNtkCur, pFileName, fWriteSymbols, fCompact, fUnique ); }
    return 0;

usage:
    fmt.Fprintf(pAbc.Err, "usage: write_aiger [-scuvh] <file>\n");
    fmt.Fprintf(pAbc.Err, "\t         writes the network in the AIGER format (http://fmv.jku.at/aiger)\n");
    fmt.Fprintf(pAbc.Err, "\t-s     : toggle saving I/O names [default = %s]\n", fWriteSymbols? "yes" : "no");
    fmt.Fprintf(pAbc.Err, "\t-c     : toggle writing more compactly [default = %s]\n", fCompact? "yes" : "no");
    fmt.Fprintf(pAbc.Err, "\t-u     : toggle writing canonical AIG structure [default = %s]\n", fUnique? "yes" : "no");
    fmt.Fprintf(pAbc.Err, "\t-v     : toggle printing verbose information [default = %s]\n", fVerbose? "yes" : "no");
    fmt.Fprintf(pAbc.Err, "\t-h     : print the help massage\n");
    fmt.Fprintf(pAbc.Err, "\tfile   : the name of the file to write (extension .aig)\n");
    return 1;
}

/**Function*************************************************************

  Synopsis    []

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func IoCommandWriteAigerCex(pAbc *Abc_Frame_t, argc int32, argv *string) int32 {
    var pFileName string;
    var c int32;
    Extra_UtilGetoptReset();
    for  ( c = Extra_UtilGetopt( argc, argv, "h" ) ) != -1  {
        switch  c  {
            case 'h':
                goto usage;
            default:
                goto usage;
        }
    }
    if pAbc.pCex == nil {
        fmt.Fprintf(pAbc.Out, "There is no current CEX.\n");
        return 0;
    }
    if argc != globalUtilOptind + 1 {
        goto usage; }
    // get the output file name
    pFileName = argv[globalUtilOptind];
    Io_WriteAigerCex( pAbc.pCex, pAbc.pNtkCur, pAbc.pGia, pFileName );
    return 0;

usage:
    fmt.Fprintf(pAbc.Err, "usage: write_aiger_cex [-h] <file>\n");
    fmt.Fprintf(pAbc.Err, "\t         writes the current CEX in the AIGER format (http://fmv.jku.at/aiger)\n");
    fmt.Fprintf(pAbc.Err, "\t-h     : print the help massage\n");
    fmt.Fprintf(pAbc.Err, "\tfile   : the name of the file to write\n");
    return 1;
}

/**Function*************************************************************

  Synopsis    []

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func IoCommandWriteBaf(pAbc *Abc_Frame_t, argc int32, argv *string) int32 {
    var pFileName string;
    var c int32;

    Extra_UtilGetoptReset();
    for  ( c = Extra_UtilGetopt( argc, argv, "h" ) ) != -1  {
        switch  c  {
            case 'h':
                goto usage;
            default:
                goto usage;
        }
    }
    if pAbc.pNtkCur == nil {
        fmt.Fprintf(pAbc.Out, "Empty network.\n");
        return 0;
    }
    if argc != globalUtilOptind + 1 {
        goto usage; }
    // get the output file name
    pFileName = argv[globalUtilOptind];
    // call the corresponding file writer
    Io_Write( pAbc.pNtkCur, pFileName, IO_FILE_BAF );
    return 0;

usage:
    fmt.Fprintf(pAbc.Err, "usage: write_baf [-h] <file>\n");
    fmt.Fprintf(pAbc.Err, "\t         writes the network into a BLIF file\n");
    fmt.Fprintf(pAbc.Err, "\t-h     : print the help massage\n");
    fmt.Fprintf(pAbc.Err, "\tfile   : the name of the file to write (extension .baf)\n");
    return 1;
}

/**Function*************************************************************

  Synopsis    []

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func IoCommandWriteBblif(pAbc *Abc_Frame_t, argc int32, argv *string) int32 {
    var pFileName string;
    var c int32;

    Extra_UtilGetoptReset();
    for  ( c = Extra_UtilGetopt( argc, argv, "h" ) ) != -1  {
        switch  c  {
            case 'h':
                goto usage;
            default:
                goto usage;
        }
    }
    if pAbc.pNtkCur == nil {
        fmt.Fprintf(pAbc.Out, "Empty network.\n");
        return 0;
    }
    if argc != globalUtilOptind + 1 {
        goto usage; }
    // get the output file name
    pFileName = argv[globalUtilOptind];
    // call the corresponding file writer
    Io_Write( pAbc.pNtkCur, pFileName, IO_FILE_BBLIF );
    return 0;

usage:
    fmt.Fprintf(pAbc.Err, "usage: write_bblif [-h] <file>\n");
    fmt.Fprintf(pAbc.Err, "\t         writes the network into a binary BLIF file\n");
    fmt.Fprintf(pAbc.Err, "\t-h     : print the help massage\n");
    fmt.Fprintf(pAbc.Err, "\tfile   : the name of the file to write (extension .bblif)\n");
    return 1;
}

/**Function*************************************************************

  Synopsis    []

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func IoCommandWriteBlif(pAbc *Abc_Frame_t, argc int32, argv *string) int32 {
    var pFileName string;
    var pLutStruct string  = nil;
    var c, fSpecial int32;
    var fUseHie int32;

    Extra_UtilGetoptReset();
    for  ( c = Extra_UtilGetopt( argc, argv, "Sjah" ) ) != -1  {
        switch  c  {
            case 'S':
                if globalUtilOptind >= argc {
                    Abc_Print( -1, "Command line switch \"-S\" should be followed by string.\n" );
                    goto usage;
                }
                pLutStruct = argv[globalUtilOptind];
                globalUtilOptind++;
                if len(pLutStruct) != 2 && len(pLutStruct) != 3 {
                    Abc_Print( -1, "Command line switch \"-S\" should be followed by a 2- or 3-char string (e.g. \"44\" or \"555\").\n" );
                    goto usage;
                }
                break;
            case 'j':
                fSpecial ^= 1;
                break;
            case 'a':
                fUseHie ^= 1;
                break;
            case 'h':
                goto usage;
            default:
                goto usage;
        }
    }
    if pAbc.pNtkCur == nil {
        fmt.Fprintf(pAbc.Out, "Empty network.\n");
        return 0;
    }
    if argc != globalUtilOptind + 1 {
        goto usage; }
    // get the output file name
    pFileName = argv[globalUtilOptind];
    // call the corresponding file writer
    if (fSpecial != 0) || pLutStruct {
        Io_WriteBlifSpecial( pAbc.pNtkCur, pFileName, pLutStruct, fUseHie );
    } else {
        Io_Write( pAbc.pNtkCur, pFileName, IO_FILE_BLIF ); }
    return 0;

usage:
    fmt.Fprintf(pAbc.Err, "usage: write_blif [-S str] [-jah] <file>\n");
    fmt.Fprintf(pAbc.Err, "\t         writes the network into a BLIF file\n");
    fmt.Fprintf(pAbc.Err, "\t-S str : string representing the LUT structure [default = %s]\n", pLutStruct ? pLutStruct : "not used");  
    fmt.Fprintf(pAbc.Err, "\t-j     : enables special BLIF writing [default = %s]\n", fSpecial? "yes" : "no");;
    fmt.Fprintf(pAbc.Err, "\t-a     : enables hierarchical BLIF writing for LUT structures [default = %s]\n", fUseHie? "yes" : "no");;
    fmt.Fprintf(pAbc.Err, "\t-h     : print the help massage\n");
    fmt.Fprintf(pAbc.Err, "\tfile   : the name of the file to write (extension .blif)\n");
    return 1;
}

/**Function*************************************************************

  Synopsis    []

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func IoCommandWriteBlifMv(pAbc *Abc_Frame_t, argc int32, argv *string) int32 {
    var pFileName string;
    var c int32;

    Extra_UtilGetoptReset();
    for  ( c = Extra_UtilGetopt( argc, argv, "h" ) ) != -1  {
        switch  c  {
            case 'h':
                goto usage;
            default:
                goto usage;
        }
    }
    if pAbc.pNtkCur == nil {
        fmt.Fprintf(pAbc.Out, "Empty network.\n");
        return 0;
    }
    if argc != globalUtilOptind + 1 {
        goto usage; }
    // get the output file name
    pFileName = argv[globalUtilOptind];
    // call the corresponding file writer
    Io_Write( pAbc.pNtkCur, pFileName, IO_FILE_BLIFMV );
    return 0;

usage:
    fmt.Fprintf(pAbc.Err, "usage: write_blif_mv [-h] <file>\n");
    fmt.Fprintf(pAbc.Err, "\t         writes the network into a BLIF-MV file\n");
    fmt.Fprintf(pAbc.Err, "\t-h     : print the help massage\n");
    fmt.Fprintf(pAbc.Err, "\tfile   : the name of the file to write (extension .mv)\n");
    return 1;
}

/**Function*************************************************************

  Synopsis    []

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func IoCommandWriteBench(pAbc *Abc_Frame_t, argc int32, argv *string) int32 {
    var pFileName string;
    var fUseLuts int32;
    var c int32;

    fUseLuts = 1;
    Extra_UtilGetoptReset();
    for  ( c = Extra_UtilGetopt( argc, argv, "lh" ) ) != -1  {
        switch  c  {
            case 'l':
                fUseLuts ^= 1;
                break;
            case 'h':
                goto usage;
            default:
                goto usage;
        }
    }
    if pAbc.pNtkCur == nil {
        fmt.Fprintf(pAbc.Out, "Empty network.\n");
        return 0;
    }
    if argc != globalUtilOptind + 1 {
        goto usage; }
    // get the output file name
    pFileName = argv[globalUtilOptind];
    // call the corresponding file writer
    if !(fUseLuts != 0) {
        Io_Write( pAbc.pNtkCur, pFileName, IO_FILE_BENCH );
    } else if pAbc.pNtkCur {
        var pNtkTemp *Abc_Ntk_t;
        pNtkTemp = Abc_NtkToNetlist( pAbc.pNtkCur );
        Abc_NtkToAig( pNtkTemp );
        Io_WriteBenchLut( pNtkTemp, pFileName );
        Abc_NtkDelete( pNtkTemp );
    
    } else {
        fmt.Printf("There is no current network.\n"); }
    return 0;

usage:
    fmt.Fprintf(pAbc.Err, "usage: write_bench [-lh] <file>\n");
    fmt.Fprintf(pAbc.Err, "\t         writes the network in BENCH format\n");
    fmt.Fprintf(pAbc.Err, "\t-l     : toggle using LUTs in the output [default = %s]\n", fUseLuts? "yes" : "no");
    fmt.Fprintf(pAbc.Err, "\t-h     : print the help massage\n");
    fmt.Fprintf(pAbc.Err, "\tfile   : the name of the file to write (extension .bench)\n");
    return 1;
}

/**Function*************************************************************

  Synopsis    []

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func IoCommandWriteBook(pAbc *Abc_Frame_t, argc int32, argv *string) int32 {
    var pFileName string;
        var c int32;
    Extra_UtilGetoptReset();
    for  ( c = Extra_UtilGetopt( argc, argv, "h" ) ) != -1  {
        switch  c  {
            case 'h':
                goto usage;
            default:
                goto usage;
        }
    }
    if argc != globalUtilOptind + 1 {
        goto usage; }
    // get the output file name
    pFileName = argv[globalUtilOptind];
        // call the corresponding file writer
    Io_Write( pAbc.pNtkCur, pFileName, IO_FILE_BOOK );
    return 0;

usage:
    fmt.Fprintf(pAbc.Err, "usage: write_book [-h] <file> [-options]\n");
    fmt.Fprintf(pAbc.Err, "\t-h     : prints the help massage\n");
    fmt.Fprintf(pAbc.Err, "\tfile   : the name of the file to write (extension .aux, .nodes, .nets)\n");
    fmt.Fprintf(pAbc.Err, "\t\n");
    fmt.Fprintf(pAbc.Err, "\tThis command is developed by Myungchul Kim (University of Michigan).\n");
    return 1;
}

/**Function*************************************************************

  Synopsis    []

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func IoCommandWriteCellNet(pAbc *Abc_Frame_t, argc int32, argv *string) int32 {
    var pNtk *Abc_Ntk_t;
    var pFileName string;
    var c int32;
    extern void Io_WriteCellNet( Abc_Ntk_t * pNtk, char * pFileName );

    Extra_UtilGetoptReset();
    for  ( c = Extra_UtilGetopt( argc, argv, "h" ) ) != -1  {
        switch  c  {
            case 'h':
                goto usage;
            default:
                goto usage;
        }
    }
    if pAbc.pNtkCur == nil {
        fmt.Fprintf(pAbc.Out, "Empty network.\n");
        return 0;
    }
    if argc != globalUtilOptind + 1 {
        goto usage; }
    pNtk = pAbc.pNtkCur;
    // get the output file name
    pFileName = argv[globalUtilOptind];
    // call the corresponding file writer
    if !Abc_NtkIsLogic(pNtk) {
        fmt.Fprintf(pAbc.Out, "The network should be a logic network (if it an AIG, use command \"logic\")\n");
        return 0;
    }
    Io_WriteCellNet( pNtk, pFileName );
    return 0;

usage:
    fmt.Fprintf(pAbc.Err, "usage: write_cellnet [-h] <file>\n");
    fmt.Fprintf(pAbc.Err, "\t         writes the network is the cellnet format\n");
    fmt.Fprintf(pAbc.Err, "\t-h     : print the help massage\n");
    fmt.Fprintf(pAbc.Err, "\tfile   : the name of the file to write\n");
    return 1;
}

/**Function*************************************************************

  Synopsis    []

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func IoCommandWriteCnf(pAbc *Abc_Frame_t, argc int32, argv *string) int32 {
    var pFileName string;
    var c int32;
    var fNewAlgo int32;
    var fFastAlgo int32;
    var fAllPrimes int32;
    var fChangePol int32;
    var fVerbose int32;
    extern Abc_Ntk_t * Abc_NtkDarToCnf( Abc_Ntk_t * pNtk, char * pFileName, int fFastAlgo, int fChangePol, int fVerbose );

    fNewAlgo = 1;
    fFastAlgo = 0;
    fAllPrimes = 0;
    fChangePol = 1;
    fVerbose = 0;
    Extra_UtilGetoptReset();
    for  ( c = Extra_UtilGetopt( argc, argv, "nfpcvh" ) ) != -1  {
        switch  c  {
            case 'n':
                fNewAlgo ^= 1;
                break;
            case 'f':
                fFastAlgo ^= 1;
                break;
            case 'p':
                fAllPrimes ^= 1;
                break;
            case 'c':
                fChangePol ^= 1;
                break;
            case 'v':
                fVerbose ^= 1;
                break;
            case 'h':
                goto usage;
            default:
                goto usage;
        }
    }
    if pAbc.pNtkCur == nil {
        fmt.Fprintf(pAbc.Out, "Empty network.\n");
        return 0;
    }
    if argc != globalUtilOptind + 1 {
        goto usage; }
    // get the output file name
    pFileName = argv[globalUtilOptind];
    // check if the feature will be used
    if Abc_NtkIsStrash(pAbc.pNtkCur) && (fAllPrimes != 0) {
        fAllPrimes = 0;
        fmt.Printf("Warning: Selected option to write all primes has no effect when deriving CNF from AIG.\n");
    }
    // call the corresponding file writer
    if fFastAlgo != 0 {
        Abc_NtkDarToCnf( pAbc.pNtkCur, pFileName, 1, fChangePol, fVerbose );
    } else if fNewAlgo != 0 {
        Abc_NtkDarToCnf( pAbc.pNtkCur, pFileName, 0, fChangePol, fVerbose );
    } else if fAllPrimes != 0 {
        Io_WriteCnf( pAbc.pNtkCur, pFileName, 1 );
    } else {
        Io_Write( pAbc.pNtkCur, pFileName, IO_FILE_CNF ); }
    return 0;

usage:
    fmt.Fprintf(pAbc.Err, "usage: write_cnf [-nfpcvh] <file>\n");
    fmt.Fprintf(pAbc.Err, "\t         generates CNF for the miter (see also \"&write_cnf\")\n");
    fmt.Fprintf(pAbc.Err, "\t-n     : toggle using new algorithm [default = %s]\n", fNewAlgo? "yes" : "no");
    fmt.Fprintf(pAbc.Err, "\t-f     : toggle using fast algorithm [default = %s]\n", fFastAlgo? "yes" : "no");
    fmt.Fprintf(pAbc.Err, "\t-p     : toggle using all primes to enhance implicativity [default = %s]\n", fAllPrimes? "yes" : "no");
    fmt.Fprintf(pAbc.Err, "\t-c     : toggle adjasting polarity of internal variables [default = %s]\n", fChangePol? "yes" : "no");
    fmt.Fprintf(pAbc.Err, "\t-v     : toggle printing verbose information [default = %s]\n", fVerbose? "yes" : "no");
    fmt.Fprintf(pAbc.Err, "\t-h     : print the help massage\n");
    fmt.Fprintf(pAbc.Err, "\tfile   : the name of the file to write\n");
    return 1;
}

/**Function*************************************************************

  Synopsis    []

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func IoCommandWriteCnf2(pAbc *Abc_Frame_t, argc int32, argv *string) int32 {
    extern void Jf_ManDumpCnf( Gia_Man_t * p, char * pFileName, int fVerbose );
    extern void Mf_ManDumpCnf( Gia_Man_t * p, char * pFileName, int nLutSize, int fCnfObjIds, int fAddOrCla, int fVerbose );
    var pFile *os.File;
    var pFileName string;
    var nLutSize int32     = 8;
    var fNewAlgo int32     = 1;
    var fCnfObjIds int32;
    var fAddOrCla int32    = 1;
    var c, fVerbose int32;
    Extra_UtilGetoptReset();
    for  ( c = Extra_UtilGetopt( argc, argv, "Kaiovh" ) ) != -1  {
        switch  c  {
            case 'K':
                if globalUtilOptind >= argc {
                    Abc_Print( -1, "Command line switch \"-K\" should be followed by an integer.\n" );
                    goto usage;
                }
                nLutSize = strconv.Atoi(argv[globalUtilOptind]);
                globalUtilOptind++;
                break;
            case 'a':
                fNewAlgo ^= 1;
                break;
            case 'i':
                fCnfObjIds ^= 1;
                break;
            case 'o':
                fAddOrCla ^= 1;
                break;
            case 'v':
                fVerbose ^= 1;
                break;
            case 'h':
                goto usage;
            default:
                goto usage;
        }
    }
    if pAbc.pGia == nil {
        Abc_Print( -1, "IoCommandWriteCnf2(): There is no AIG.\n" );
        return 1;
    } 
    if Gia_ManRegNum(pAbc.pGia) > 0 {
        Abc_Print( -1, "IoCommandWriteCnf2(): Works only for combinational miters.\n" );
        return 0;
    }
    if nLutSize < 3 || nLutSize > 8 {
        Abc_Print( -1, "IoCommandWriteCnf2(): Invalid LUT size (%d).\n", nLutSize );
        return 0;
    }
    if !(fNewAlgo != 0) && !Sdm_ManCanRead() {
        Abc_Print( -1, "IoCommandWriteCnf2(): Cannot input precomputed DSD information.\n" );
        return 0;
    }
    if argc != globalUtilOptind + 1 {
        goto usage; }
    // get the input file name
    pFileName = argv[globalUtilOptind];
    pFile = os.OpenFile(pFileName, os.O_WRONLY | os.O_CREATE | os.O_TRUNC, 0666);
    if pFile == nil {
        fmt.Printf("Cannot open file \"%s\" for writing.\n", pFileName);
        return 0;
    }
    _ = (pFile).Close();
    if fNewAlgo != 0 {
        Mf_ManDumpCnf( pAbc.pGia, pFileName, nLutSize, fCnfObjIds, fAddOrCla, fVerbose );
    } else {
        Jf_ManDumpCnf( pAbc.pGia, pFileName, fVerbose ); }
    return 0;

usage:
    fmt.Fprintf(pAbc.Err, "usage: &write_cnf [-Kaiovh] <file>\n");
    fmt.Fprintf(pAbc.Err, "\t           writes CNF produced by a new generator\n");
    fmt.Fprintf(pAbc.Err, "\t-K <num> : the LUT size (3 <= num <= 8) [default = %d]\n", nLutSize);
    fmt.Fprintf(pAbc.Err, "\t-a       : toggle using new algorithm [default = %s]\n", fNewAlgo? "yes" : "no");
    fmt.Fprintf(pAbc.Err, "\t-i       : toggle using AIG object IDs as CNF variables [default = %s]\n", fCnfObjIds? "yes" : "no");
    fmt.Fprintf(pAbc.Err, "\t-o       : toggle adding OR clause for the outputs [default = %s]\n", fAddOrCla? "yes" : "no");
    fmt.Fprintf(pAbc.Err, "\t-v       : toggle printing verbose information [default = %s]\n", fVerbose? "yes" : "no");
    fmt.Fprintf(pAbc.Err, "\t-h       : print the help massage\n");
    fmt.Fprintf(pAbc.Err, "\tfile     : the name of the file to write\n");
    fmt.Fprintf(pAbc.Err, "\n");
    fmt.Fprintf(pAbc.Err, "\t           CNF variable mapping rules:\n");
    fmt.Fprintf(pAbc.Err, "\n");
    fmt.Fprintf(pAbc.Err, "\t           Assume CNF has N variables, with variable IDs running from 0 to N-1.\n");
    fmt.Fprintf(pAbc.Err, "\t           Variable number 0 is not used in the CNF.\n");
    fmt.Fprintf(pAbc.Err, "\t           Variables 1, 2, 3,... <nPOs> represent POs in their natural order.\n");
    fmt.Fprintf(pAbc.Err, "\t           Variables N-<nPIs>, N-<nPIs>+1, N-<nPIs>+2, ... N-1, represent PIs in their natural order.\n");
    fmt.Fprintf(pAbc.Err, "\t           The internal variables are ordered in a reverse topological order from outputs to inputs.\n");
    fmt.Fprintf(pAbc.Err, "\t           That is, smaller variable IDs tend to be closer to the outputs, while larger\n");
    fmt.Fprintf(pAbc.Err, "\t           variable IDs tend to be closer to the inputs. It was found that this ordering\n");
    fmt.Fprintf(pAbc.Err, "\t           leads to faster SAT solving for hard UNSAT CEC problems.\n");
    return 1;
}


/**Function*************************************************************

  Synopsis    []

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func IoCommandWriteDot(pAbc *Abc_Frame_t, argc int32, argv *string) int32 {
    var pFileName string;
    var c int32;

    Extra_UtilGetoptReset();
    for  ( c = Extra_UtilGetopt( argc, argv, "h" ) ) != -1  {
        switch  c  {
            case 'h':
                goto usage;
            default:
                goto usage;
        }
    }
    if pAbc.pNtkCur == nil {
        fmt.Fprintf(pAbc.Out, "Empty network.\n");
        return 0;
    }
    if argc != globalUtilOptind + 1 {
        goto usage; }
    // get the output file name
    pFileName = argv[globalUtilOptind];
    // call the corresponding file writer
    Io_Write( pAbc.pNtkCur, pFileName, IO_FILE_DOT );
    return 0;

usage:
    fmt.Fprintf(pAbc.Err, "usage: write_dot [-h] <file>\n");
    fmt.Fprintf(pAbc.Err, "\t         writes the current network into a DOT file\n");
    fmt.Fprintf(pAbc.Err, "\t-h     : print the help massage\n");
    fmt.Fprintf(pAbc.Err, "\tfile   : the name of the file to write\n");
    return 1;
}

//ABC_NAMESPACE_IMPL_END

//#include "proof/fra/fra.h"

//ABC_NAMESPACE_IMPL_START

func Abc_NtkCheckSpecialPi(pNtk *Abc_Ntk_t) int32 {
    var pObj *Abc_Obj_t;  var i int32;
    Abc_NtkForEachPi( pNtk, pObj, i )
        if ( !int32(strings.Compare(Abc_ObjName(pObj), "_abc_190121_abc_")) )
            return 1;
    return 0;
}

/**Function*************************************************************

  Synopsis    []

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Abc_NtkDumpOneCexSpecial(pFile *os.File, pNtk *Abc_Ntk_t, pCex *Abc_Cex_t) {
    var pCare *Abc_Cex_t  = nil; var i, f int32; var pObj *Abc_Obj_t;
    extern Aig_Man_t * Abc_NtkToDar( Abc_Ntk_t * pNtk, int fExors, int fRegisters );
    var pAig *Aig_Man_t  = Abc_NtkToDar( pNtk, 0, 1 );
    //fprintf( pFile, "# FALSIFYING OUTPUTS:");                                       
    //fprintf( pFile, " %s", Abc_ObjName(Abc_NtkCo(pNtk, pCex->iPo)) ); 
    pCare = Bmc_CexCareMinimize( pAig, Saig_ManPiNum(pAig), pCex, 4, 0, 0 );
    Aig_ManStop( pAig );
    if pCare == nil {
        fmt.Printf("Counter-example minimization has failed.\n"); 
        return;
    }
    // output flop values (unaffected by the minimization)
    Abc_NtkForEachLatch( pNtk, pObj, i )
        fmt.Fprintf(pFile, "CEX: %s@0=%c\n", Abc_ObjName(Abc_ObjFanout0(pObj)), '0'+!Abc_LatchIsInit0(pObj));
    // output PI values (while skipping the minimized ones)
    for  f = 0; f <= pCex.iFrame; f++  {
        Abc_NtkForEachPi( pNtk, pObj, i )
            if ( !pCare || Abc_InfoHasBit(pCare.pData, pCare.nRegs+pCare.nPis*f + i) )
                fmt.Fprintf(pFile, "CEX: %s@%d=%c\n", Abc_ObjName(pObj), f, '0'+Abc_InfoHasBit(pCex.pData, pCex.nRegs+pCex.nPis*f + i));
    }
    Abc_CexFreeP( &pCare );
}


func Abc_NtkDumpOneCex(pFile *os.File, pNtk *Abc_Ntk_t, pCex *Abc_Cex_t, fPrintFull int32, fNames int32, fUseFfNames int32, fMinimize int32, fUseOldMin int32, fCheckCex int32, fUseSatBased int32, fHighEffort int32, fAiger int32, fVerbose int32) {
    var pObj *Abc_Obj_t;
    var i, f int32;
    if fPrintFull != 0 {
        extern Aig_Man_t * Abc_NtkToDar( Abc_Ntk_t * pNtk, int fExors, int fRegisters );
        var pAig *Aig_Man_t  = Abc_NtkToDar( pNtk, 0, 1 );
        var pCexFull *Abc_Cex_t  = Saig_ManExtendCex( pAig, pCex );
        Aig_ManStop( pAig );
        // output PI values (while skipping the minimized ones)
        assert(pCexFull.nBits == Abc_NtkCiNum(pNtk) * (pCex.iFrame + 1), "pCexFull.nBits == Abc_NtkCiNum(pNtk) * (pCex.iFrame + 1)");
        for  f = 0; f <= pCex.iFrame; f++  {
            Abc_NtkForEachCi( pNtk, pObj, i )
                fmt.Fprintf(pFile, "%s@%d=%c ", Abc_ObjName(pObj), f, '0'+Abc_InfoHasBit(pCexFull.pData, Abc_NtkCiNum(pNtk)*f + i));
        }
        Abc_CexFreeP( &pCexFull );
    
    } else if fNames != 0 {
        var pCare *Abc_Cex_t  = nil;
        if fMinimize != 0 {
            extern Aig_Man_t * Abc_NtkToDar( Abc_Ntk_t * pNtk, int fExors, int fRegisters );
            var pAig *Aig_Man_t  = Abc_NtkToDar( pNtk, 0, 1 );
            fmt.Fprintf(pFile, "# FALSIFYING OUTPUTS:");                                       
            fmt.Fprintf(pFile, " %s", Abc_ObjName(Abc_NtkCo(pNtk, pCex.iPo))); 
            if fUseOldMin != 0 {
                pCare = Saig_ManCbaFindCexCareBits( pAig, pCex, 0, fVerbose );
                if fCheckCex != 0 {
                    Bmc_CexCareVerify( pAig, pCex, pCare, fVerbose ); }
            
            } else if fUseSatBased != 0 {
                pCare = Bmc_CexCareSatBasedMinimize( pAig, Saig_ManPiNum(pAig), pCex, fHighEffort, fCheckCex, fVerbose );
            } else {
                pCare = Bmc_CexCareMinimize( pAig, Saig_ManPiNum(pAig), pCex, 4, fCheckCex, fVerbose ); }
            Aig_ManStop( pAig );
            if pCare == nil {                                           
                fmt.Printf("Counter-example minimization has failed.\n"); } 
        
        } else {
            fmt.Fprintf(pFile, "# FALSIFYING OUTPUTS:");                        
            fmt.Fprintf(pFile, " %s", Abc_ObjName(Abc_NtkCo(pNtk, pCex.iPo)));
        }
        fmt.Fprintf(pFile, "\n");                                           
        fmt.Fprintf(pFile, "# COUNTEREXAMPLE LENGTH: %d\n", pCex.iFrame+1);
        if (fUseFfNames != 0) && (Abc_NtkCheckSpecialPi(pNtk) != 0) {
            var pValues *int32;
            var nXValues int32;
            var iFlop int32;
            var iPivotPi int32 = -1;
            Abc_NtkForEachPi( pNtk, pObj, iPivotPi )
                if ( !int32(strings.Compare(Abc_ObjName(pObj), "_abc_190121_abc_")) )
                    break;
            if iPivotPi == Abc_NtkPiNum(pNtk) {
                fmt.Fprintf(os.Stdout, "IoCommandWriteCex(): Cannot find special PI required by switch \"-z\".\n");
                return;
            }
            // count X-valued flops
            for  i = iPivotPi+1; i < Abc_NtkPiNum(pNtk); i++  {
                if Abc_ObjName(Abc_NtkPi(pNtk, i))[0] == 'x' {
                    nXValues++; }
            }
            // map X-valued flops into auxiliary PIs
            pValues = ABC_FALLOC( int, Abc_NtkPiNum(pNtk) );
            for  i = iPivotPi+1; i < Abc_NtkPiNum(pNtk); i++  {
                if Abc_ObjName(Abc_NtkPi(pNtk, i))[0] == 'x' {
                    pValues[i] = iPivotPi - nXValues + iFlop++; }
            }
            assert(iFlop == nXValues, "iFlop == nXValues");
            // write flop values
            for  i = iPivotPi+1; i < Abc_NtkPiNum(pNtk); i++  {
                if pValues[i] == -1 {
                    fmt.Fprintf(pFile, "%s@0=%c\n", Abc_ObjName(Abc_NtkPi(pNtk, i))+1, Abc_ObjName(Abc_NtkPi(pNtk, i))[0]);
                } else if Abc_InfoHasBit(pCare.pData, pCare.nRegs + pValues[i]) {
                    fmt.Fprintf(pFile, "%s@0=%c\n", Abc_ObjName(Abc_NtkPi(pNtk, i))+1, '0'+Abc_InfoHasBit(pCex.pData, pCex.nRegs + pValues[i])); }
            }
            ABC_FREE( pValues );
            // output PI values (while skipping the minimized ones)
            for ( f = 0; f <= pCex.iFrame; f++ )
                Abc_NtkForEachPi( pNtk, pObj, i )
                {
                    if i == iPivotPi - nXValues { break; }
                    if !(pCare != nil) || Abc_InfoHasBit(pCare.pData, pCare.nRegs+pCare.nPis*f + i) {
                        fmt.Fprintf(pFile, "%s@%d=%c\n", Abc_ObjName(pObj), f, '0'+Abc_InfoHasBit(pCex.pData, pCex.nRegs+pCex.nPis*f + i)); }
                }
        
        } else {
            // output flop values (unaffected by the minimization)
            Abc_NtkForEachLatch( pNtk, pObj, i )
                fmt.Fprintf(pFile, "%s@0=%c\n", Abc_ObjName(Abc_ObjFanout0(pObj)), '0'+!Abc_LatchIsInit0(pObj));
            // output PI values (while skipping the minimized ones)
            for  f = 0; f <= pCex.iFrame; f++  {
                Abc_NtkForEachPi( pNtk, pObj, i )
                    if ( !pCare || Abc_InfoHasBit(pCare.pData, pCare.nRegs+pCare.nPis*f + i) )
                        fmt.Fprintf(pFile, "%s@%d=%c\n", Abc_ObjName(pObj), f, '0'+Abc_InfoHasBit(pCex.pData, pCex.nRegs+pCex.nPis*f + i));
            }
        }
        Abc_CexFreeP( &pCare );
    
    } else {
        Abc_NtkForEachLatch( pNtk, pObj, i )
            fmt.Fprintf(pFile, "%c", '0'+!Abc_LatchIsInit0(pObj));

        for  i = pCex.nRegs; i < pCex.nBits; i++  {
            if (fAiger != 0) && (i-pCex.nRegs)%pCex.nPis == 0 {
                fmt.Fprintf(pFile, "\n"); }
            fmt.Fprintf(pFile, "%c", '0'+Abc_InfoHasBit(pCex.pData, i));
        }
    }
}

/**Function*************************************************************

  Synopsis    []

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func IoCommandWriteCex(pAbc *Abc_Frame_t, argc int32, argv *string) int32 {
    var pNtk *Abc_Ntk_t;
    var pFileName string;
    var c, fNames int32;
    var fMinimize int32;
    var fUseSatBased int32;
    var fHighEffort int32;
    var fUseOldMin int32;
    var fCheckCex int32;
    var forceSeq int32;
    var fAiger int32;
    var fPrintFull int32;
    var fUseFfNames int32;
    var fVerbose int32;

    Extra_UtilGetoptReset();
    for  ( c = Extra_UtilGetopt( argc, argv, "snmueocafzvh" ) ) != -1  {
        switch  c  {
            case 's':
                forceSeq ^= 1;
                break;
            case 'n':
                fNames ^= 1;
                break;
            case 'm':
                fMinimize ^= 1;
                break;
            case 'u':
                fUseSatBased ^= 1;
                break;
            case 'e':
                fHighEffort ^= 1;
                break;
            case 'o':
                fUseOldMin ^= 1;
                break;
            case 'c':
                fCheckCex ^= 1;
                break;
            case 'a':
                fAiger ^= 1;
                break;
            case 'f':
                fPrintFull ^= 1;
                break;
            case 'z':
                fUseFfNames ^= 1;
                break;
            case 'v':
                fVerbose ^= 1;
                break;
            case 'h':
                goto usage;
            default:
                goto usage;
        }
    }
    pNtk = pAbc.pNtkCur;
    if pNtk == nil {
        fmt.Fprintf(pAbc.Out, "Empty network.\n");
        return 0;
    }
    if pNtk.pModel == nil && pAbc.pCex == nil && pAbc.vCexVec == nil {
        fmt.Fprintf(pAbc.Out, "Counter-example is not available.\n");
        return 0;
    }
    if argc != globalUtilOptind + 1 {
        fmt.Printf("File name is missing on the command line.\n");
        goto usage;
    }

    // get the input file name
    pFileName = argv[globalUtilOptind];
    // write the counter-example into the file
    if pAbc.pCex || pAbc.vCexVec { 
        var pCex *Abc_Cex_t  = pAbc.pCex;
        var pFile *os.File; var i int32;
        /*
        Abc_NtkForEachLatch( pNtk, pObj, i )
            if ( !Abc_LatchIsInit0(pObj) )
            {
                fprintf( stdout, "IoCommandWriteCex(): The init-state should be all-0 for counter-example to work.\n" );
                fprintf( stdout, "Run commands \"undc\" and \"zero\" and then rerun the equivalence check.\n" );
                return 1;
            }
        */
        pFile = os.OpenFile(pFileName, os.O_WRONLY | os.O_CREATE | os.O_TRUNC, 0666);
        if pFile == nil {
            fmt.Fprintf(os.Stdout, "IoCommandWriteCex(): Cannot open the output file \"%s\".\n", pFileName);
            return 1;
        }
        if pAbc.pCex {
            Abc_NtkDumpOneCex( pFile, pNtk, pCex, 
                fPrintFull, fNames, fUseFfNames, fMinimize, fUseOldMin, 
                fCheckCex, fUseSatBased, fHighEffort, fAiger, fVerbose );
        
        } else if pAbc.vCexVec {
            Vec_PtrForEachEntry( Abc_Cex_t *, pAbc.vCexVec, pCex, i )
            {
                if pCex == nil {
                    continue; }
                fmt.Fprintf(pFile, "#\n#\n# CEX for output %d\n#\n", i); 
                Abc_NtkDumpOneCex( pFile, pNtk, pCex, 
                    fPrintFull, fNames, fUseFfNames, fMinimize, fUseOldMin, 
                    fCheckCex, fUseSatBased, fHighEffort, fAiger, fVerbose );
            }
        }
        fmt.Fprintf(pFile, "# DONE\n"); 
        _ = (pFile).Close();
    
    } else {
        var pObj *Abc_Obj_t;
        var pFile *os.File  = os.OpenFile(pFileName, os.O_WRONLY | os.O_CREATE | os.O_TRUNC, 0666);
        var i int32;
        if pFile == nil {
            fmt.Fprintf(os.Stdout, "IoCommandWriteCex(): Cannot open the output file \"%s\".\n", pFileName);
            return 1;
        }
        if fNames != 0 {
            char *cycle_ctr = forceSeq?"@0":"";
            Abc_NtkForEachPi( pNtk, pObj, i )
//                fprintf( pFile, "%s=%c\n", Abc_ObjName(pObj), '0'+(pNtk->pModel[i]==1) );
                fmt.Fprintf(pFile, "%s%s=%c\n", Abc_ObjName(pObj), cycle_ctr, '0'+(pNtk.pModel[i]==1));
        
        } else {
            Abc_NtkForEachPi( pNtk, pObj, i )
                fmt.Fprintf(pFile, "%c", '0'+(pNtk.pModel[i]==1));
        }
        fmt.Fprintf(pFile, "\n");
        _ = (pFile).Close();
    }

    return 0;

usage:
    fmt.Fprintf(pAbc.Err, "usage: write_cex [-snmueocfzvh] <file>\n");
    fmt.Fprintf(pAbc.Err, "\t         saves counter-example (CEX) derived by \"sat\", \"iprove\", \"dprove\", etc\n");
    fmt.Fprintf(pAbc.Err, "\t         the output file <file> contains values for each PI in natural order\n");
    fmt.Fprintf(pAbc.Err, "\t-s     : always report a sequential CEX (cycle 0 for comb) [default = %s]\n", forceSeq? "yes": "no");
    fmt.Fprintf(pAbc.Err, "\t-n     : write input names into the file [default = %s]\n", fNames? "yes": "no");
    fmt.Fprintf(pAbc.Err, "\t-m     : minimize CEX by dropping don't-care values [default = %s]\n", fMinimize? "yes": "no");
    fmt.Fprintf(pAbc.Err, "\t-u     : use fast SAT-based CEX minimization [default = %s]\n", fUseSatBased? "yes": "no");
    fmt.Fprintf(pAbc.Err, "\t-e     : use high-effort SAT-based CEX minimization [default = %s]\n", fHighEffort? "yes": "no");
    fmt.Fprintf(pAbc.Err, "\t-o     : use old CEX minimization algorithm [default = %s]\n", fUseOldMin? "yes": "no");
    fmt.Fprintf(pAbc.Err, "\t-c     : check generated CEX using ternary simulation [default = %s]\n", fCheckCex? "yes": "no");
    fmt.Fprintf(pAbc.Err, "\t-a     : print cex in AIGER 1.9 format [default = %s]\n", fAiger? "yes": "no");
    fmt.Fprintf(pAbc.Err, "\t-f     : enable printing flop values in each timeframe [default = %s]\n", fPrintFull? "yes": "no");  
    fmt.Fprintf(pAbc.Err, "\t-z     : toggle using saved flop names [default = %s]\n", fUseFfNames? "yes": "no");  
    fmt.Fprintf(pAbc.Err, "\t-v     : enable verbose output [default = %s]\n", fVerbose? "yes": "no");  
    fmt.Fprintf(pAbc.Err, "\t-h     : print the help massage\n");
    fmt.Fprintf(pAbc.Err, "\t<file> : the name of the file to write\n");
    return 1;
}

/**Function*************************************************************

  Synopsis    []

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func IoCommandWriteEqn(pAbc *Abc_Frame_t, argc int32, argv *string) int32 {
    var pFileName string;
    var c int32;

    Extra_UtilGetoptReset();
    for  ( c = Extra_UtilGetopt( argc, argv, "h" ) ) != -1  {
        switch  c  {
            case 'h':
                goto usage;
            default:
                goto usage;
        }
    }
    if pAbc.pNtkCur == nil {
        fmt.Fprintf(pAbc.Out, "Empty network.\n");
        return 0;
    }
    if argc != globalUtilOptind + 1 {
        goto usage; }
    // get the output file name
    pFileName = argv[globalUtilOptind];
    // call the corresponding file writer
    Io_Write( pAbc.pNtkCur, pFileName, IO_FILE_EQN );
    return 0;

usage:
    fmt.Fprintf(pAbc.Err, "usage: write_eqn [-h] <file>\n");
    fmt.Fprintf(pAbc.Err, "\t         writes the current network in the equation format\n");
    fmt.Fprintf(pAbc.Err, "\t-h     : print the help massage\n");
    fmt.Fprintf(pAbc.Err, "\tfile   : the name of the file to write\n");
    return 1;
}

/**Function*************************************************************

  Synopsis    []

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func IoCommandWriteGml(pAbc *Abc_Frame_t, argc int32, argv *string) int32 {
    var pFileName string;
    var c int32;

    Extra_UtilGetoptReset();
    for  ( c = Extra_UtilGetopt( argc, argv, "h" ) ) != -1  {
        switch  c  {
            case 'h':
                goto usage;
            default:
                goto usage;
        }
    }
    if pAbc.pNtkCur == nil {
        fmt.Fprintf(pAbc.Out, "Empty network.\n");
        return 0;
    }
    if argc != globalUtilOptind + 1 {
        goto usage; }
    // get the output file name
    pFileName = argv[globalUtilOptind];
    // call the corresponding file writer
    Io_Write( pAbc.pNtkCur, pFileName, IO_FILE_GML );
    return 0;

usage:
    fmt.Fprintf(pAbc.Err, "usage: write_gml [-h] <file>\n");
    fmt.Fprintf(pAbc.Err, "\t         writes network using graph representation formal GML\n");
    fmt.Fprintf(pAbc.Err, "\t-h     : print the help massage\n");
    fmt.Fprintf(pAbc.Err, "\tfile   : the name of the file to write\n");
    return 1;
}

/**Function*************************************************************

  Synopsis    []

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func IoCommandWriteList(pAbc *Abc_Frame_t, argc int32, argv *string) int32 {
    var pFileName string;
    var fUseHost int32;
    var c int32;

    fmt.Printf("This command currently does not work.\n");
    return 0;

    fUseHost = 1;
    Extra_UtilGetoptReset();
    for  ( c = Extra_UtilGetopt( argc, argv, "nh" ) ) != -1  {
        switch  c  {
            case 'n':
                fUseHost ^= 1;
                break;
            case 'h':
                goto usage;
            default:
                goto usage;
        }
    }
    if pAbc.pNtkCur == nil {
        fmt.Fprintf(pAbc.Out, "Empty network.\n");
        return 0;
    }
    if argc != globalUtilOptind + 1 {
        goto usage; }
/*
    if ( !Abc_NtkIsSeq(pAbc->pNtkCur) )
    {
        fprintf( stdout, "IoCommandWriteList(): Can write adjacency list for sequential AIGs only.\n" );
        return 0;
    }
*/
    // get the input file name
    pFileName = argv[globalUtilOptind];
    // write the file
    Io_WriteList( pAbc.pNtkCur, pFileName, fUseHost );
    return 0;

usage:
    fmt.Fprintf(pAbc.Err, "usage: write_list [-nh] <file>\n");
    fmt.Fprintf(pAbc.Err, "\t         writes network using graph representation formal GML\n");
    fmt.Fprintf(pAbc.Err, "\t-n     : toggle writing host node [default = %s]\n", fUseHost? "yes":"no");
    fmt.Fprintf(pAbc.Err, "\t-h     : print the help massage\n");
    fmt.Fprintf(pAbc.Err, "\tfile   : the name of the file to write\n");
    return 1;
}

/**Function*************************************************************

  Synopsis    []

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func IoCommandWritePla(pAbc *Abc_Frame_t, argc int32, argv *string) int32 {
    extern int Io_WriteMoPlaM( Abc_Ntk_t * pNtk, char * pFileName, int nMints );
    var pFileName string;
    var c, fUseMoPla, nMints int32;

    Extra_UtilGetoptReset();
    for  ( c = Extra_UtilGetopt( argc, argv, "Mmh" ) ) != -1  {
        switch  c  {
            case 'M':
                if globalUtilOptind >= argc {
                    Abc_Print( -1, "Command line switch \"-M\" should be followed by an integer.\n" );
                    goto usage;
                }
                nMints = strconv.Atoi(argv[globalUtilOptind]);
                globalUtilOptind++;
                break;
            case 'm':
                fUseMoPla ^= 1;
                break;
            case 'h':
                goto usage;
            default:
                goto usage;
        }
    }
    if pAbc.pNtkCur == nil {
        fmt.Fprintf(pAbc.Out, "Empty network.\n");
        return 0;
    }
    if argc != globalUtilOptind + 1 {
        goto usage; }
    // get the output file name
    pFileName = argv[globalUtilOptind];
    // call the corresponding file writer
    if nMints != 0 {
        if Abc_NtkIsBddLogic(pAbc.pNtkCur) {
            Io_WriteMoPlaM( pAbc.pNtkCur, pFileName, nMints );
        } else {
            var pStrash *Abc_Ntk_t  = Abc_NtkStrash( pAbc.pNtkCur, 0, 0, 0 );
            Io_WriteMoPlaM( pStrash, pFileName, nMints );
            Abc_NtkDelete( pStrash );
        }
    
    } else {
        Io_Write( pAbc.pNtkCur, pFileName, fUseMoPla ? IO_FILE_MOPLA : IO_FILE_PLA ); }
    return 0;

usage:
    fmt.Fprintf(pAbc.Err, "usage: write_pla [-M <num>] [-mh] <file>\n");
    fmt.Fprintf(pAbc.Err, "\t           writes the collapsed network into a PLA file\n");
    fmt.Fprintf(pAbc.Err, "\t-M <num> : the number of on-set minterms to write [default = %d]\n", nMints);
    fmt.Fprintf(pAbc.Err, "\t-m       : toggle writing multi-output PLA [default = %s]\n", fUseMoPla? "yes":"no");
    fmt.Fprintf(pAbc.Err, "\t-h       : print the help massage\n");
    fmt.Fprintf(pAbc.Err, "\tfile     : the name of the file to write\n");
    return 1;
}

/**Function*************************************************************

  Synopsis    []

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func IoCommandWriteVerilog(pAbc *Abc_Frame_t, argc int32, argv *string) int32 {
    extern void Io_WriteVerilogLut( Abc_Ntk_t * pNtk, char * pFileName, int nLutSize );
    var pFileName string;
    var c, fOnlyAnds int32;
    var nLutSize int32  = -1;

    Extra_UtilGetoptReset();
    for  ( c = Extra_UtilGetopt( argc, argv, "Kah" ) ) != -1  {
        switch  c  {
            case 'K':
                if globalUtilOptind >= argc {
                    Abc_Print( -1, "Command line switch \"-K\" should be followed by an integer.\n" );
                    goto usage;
                }
                nLutSize = strconv.Atoi(argv[globalUtilOptind]);
                globalUtilOptind++;
                if nLutSize < 2 || nLutSize > 6 {
                    goto usage; }
                break;
            case 'a':
                fOnlyAnds ^= 1;
                break;
            case 'h':
                goto usage;
            default:
                goto usage;
        }
    }
    if pAbc.pNtkCur == nil {
        fmt.Fprintf(pAbc.Out, "Empty network.\n");
        return 0;
    }
    if argc != globalUtilOptind + 1 {
        goto usage; }
    // get the output file name
    pFileName = argv[globalUtilOptind];
    // call the corresponding file writer
    if fOnlyAnds != 0 {
        var pNtkTemp *Abc_Ntk_t  = Abc_NtkToNetlist( pAbc.pNtkCur );
        if !Abc_NtkHasAig(pNtkTemp) && !Abc_NtkHasMapping(pNtkTemp) {
            Abc_NtkToAig( pNtkTemp ); }
        Io_WriteVerilog( pNtkTemp, pFileName, 1 );
        Abc_NtkDelete( pNtkTemp );
    
    } else if nLutSize >= 2 && nLutSize <= 6 {
        Io_WriteVerilogLut( pAbc.pNtkCur, pFileName, nLutSize );
    } else {
        Io_Write( pAbc.pNtkCur, pFileName, IO_FILE_VERILOG ); }
    return 0;

usage:
    fmt.Fprintf(pAbc.Err, "usage: write_verilog [-K num] [-ah] <file>\n");
    fmt.Fprintf(pAbc.Err, "\t         writes the current network in Verilog format\n");
    fmt.Fprintf(pAbc.Err, "\t-K num : write the network using instances of K-LUTs (2 <= K <= 6) [default = not used]\n");
    fmt.Fprintf(pAbc.Err, "\t-a     : toggle writing expressions with only ANDs (without XORs and MUXes) [default = %s]\n", fOnlyAnds? "yes":"no");
    fmt.Fprintf(pAbc.Err, "\t-h     : print the help massage\n");
    fmt.Fprintf(pAbc.Err, "\tfile   : the name of the file to write\n");
    return 1;
}

/**Function*************************************************************

  Synopsis    []

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func IoCommandWriteSortCnf(pAbc *Abc_Frame_t, argc int32, argv *string) int32 {
    var pFileName string;
    var c int32;
    var nVars int32  = 16;
    var nQueens int32  = 4;
    extern void Abc_NtkWriteSorterCnf( char * pFileName, int nVars, int nQueens );

    Extra_UtilGetoptReset();
    for  ( c = Extra_UtilGetopt( argc, argv, "NQh" ) ) != -1  {
        switch  c  {
            case 'N':
                if globalUtilOptind >= argc {
                    fmt.Fprintf(os.Stdout, "Command line switch \"-N\" should be followed by an integer.\n");
                    goto usage;
                }
                nVars = strconv.Atoi(argv[globalUtilOptind]);
                globalUtilOptind++;
                if nVars <= 0 { 
                    goto usage; }
                break;
            case 'Q':
                if globalUtilOptind >= argc {
                    fmt.Fprintf(os.Stdout, "Command line switch \"-Q\" should be followed by an integer.\n");
                    goto usage;
                }
                nQueens = strconv.Atoi(argv[globalUtilOptind]);
                globalUtilOptind++;
                if nQueens <= 0 { 
                    goto usage; }
                break;
            case 'h':
                goto usage;
            default:
                goto usage;
        }
    }
    if argc != globalUtilOptind + 1 {
        goto usage; }
    // get the output file name
    pFileName = argv[globalUtilOptind];
    Abc_NtkWriteSorterCnf( pFileName, nVars, nQueens );
    // call the corresponding file writer
    return 0;

usage:
    fmt.Fprintf(pAbc.Err, "usage: write_sorter_cnf [-N <num>] [-Q <num>] <file>\n");
    fmt.Fprintf(pAbc.Err, "\t         writes CNF for the sorter\n");
    fmt.Fprintf(pAbc.Err, "\t-N num : the number of sorter bits [default = %d]\n", nVars);
    fmt.Fprintf(pAbc.Err, "\t-Q num : the number of bits to be asserted to 1 [default = %d]\n", nQueens);
    fmt.Fprintf(pAbc.Err, "\t-h     : print the help massage\n");
    fmt.Fprintf(pAbc.Err, "\tfile   : the name of the file to write\n");
    return 1;
}

/**Function*************************************************************

  Synopsis    []

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func IoCommandWriteTruth(pAbc *Abc_Frame_t, argc int32, argv *string) int32 {
    var vTruth *Vec_Int_t;
    var pNtk *Abc_Ntk_t  = pAbc.pNtkCur;
    var pNode *Abc_Obj_t;
    var pFileName string;
    var pFile *os.File;
    var pTruth *uint32;
    var fHex int32  = 1;
    var fReverse int32;
    var c int32;

    Extra_UtilGetoptReset();
    for  ( c = Extra_UtilGetopt( argc, argv, "xrh" ) ) != -1  {
        switch  c  {
            case 'x':
                fHex ^= 1;
                break;
            case 'r':
                fReverse ^= 1;
                break;
            case 'h':
                goto usage;
            default:
                goto usage;
        }
    }
    if pAbc.pNtkCur == nil {
        fmt.Printf("Current network is not available.\n");
        return 0;
    }
    if !Abc_NtkIsLogic(pNtk) {
        fmt.Printf("Current network should not an AIG. Run \"logic\".\n");
        return 0;
    }
    if Abc_NtkPoNum(pNtk) != 1 {
        fmt.Printf("Current network should have exactly one primary output.\n");
        return 0;
    }
    if Abc_NtkNodeNum(pNtk) != 1 {
        fmt.Printf("Current network should have exactly one node.\n");
        return 0;
    }
    pNode = Abc_ObjFanin0( Abc_NtkPo(pNtk, 0) );
    if Abc_ObjFaninNum(pNode) == 0 { 
        fmt.Printf("Can only write logic function with 0 inputs.\n");
        return 0;
    }
    if Abc_ObjFaninNum(pNode) > 16 { 
        fmt.Printf("Can only write logic function with no more than 16 inputs.\n");
        return 0;
    }
    if argc != globalUtilOptind + 1 {
        goto usage; }
    // get the input file name
    pFileName = argv[globalUtilOptind];
    // convert to logic
    Abc_NtkToAig( pNtk );
    vTruth = Vec_IntAlloc( 0 );
    pTruth = Hop_ManConvertAigToTruth( (Hop_Man_t *)pNtk.pManFunc, (Hop_Obj_t *)pNode.pData, Abc_ObjFaninNum(pNode), vTruth, fReverse );
    pFile = os.OpenFile(pFileName, os.O_WRONLY | os.O_CREATE | os.O_TRUNC, 0666);
    if pFile == nil {
        Vec_IntFree( vTruth );
        fmt.Printf("Cannot open file \"%s\" for writing.\n", pFileName);
        return 0;
    }
    if fHex != 0 {
        Extra_PrintHex2( pFile, pTruth, Abc_ObjFaninNum(pNode) );
    } else {
        Extra_PrintBinary( pFile, pTruth, 1<<Abc_ObjFaninNum(pNode) ); }
    _ = (pFile).Close();
    Vec_IntFree( vTruth );
    return 0;

usage:
    fmt.Fprintf(pAbc.Err, "usage: write_truth [-xrh] <file>\n");
    fmt.Fprintf(pAbc.Err, "\t         writes truth table into a file\n");
    fmt.Fprintf(pAbc.Err, "\t-x     : toggles between bin and hex representation [default = %s]\n", fHex?  "hex":"bin");
    fmt.Fprintf(pAbc.Err, "\t-r     : toggle reversing bits in the truth table [default = %s]\n", fReverse? "yes":"no");
    fmt.Fprintf(pAbc.Err, "\t-h     : print the help massage\n");
    fmt.Fprintf(pAbc.Err, "\tfile   : the name of the file to write\n");
    return 1;
}


/**Function*************************************************************

  Synopsis    []

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func IoCommandWriteTruths(pAbc *Abc_Frame_t, argc int32, argv *string) int32 {
    var pObj *Gia_Obj_t;
    var pFileName string;
    var pFile *os.File;
    var pTruth *word;
    var nBytes int32;
    var fReverse int32;
    var fBinary int32;
    var c, i int32;
 
    Extra_UtilGetoptReset();
    for  ( c = Extra_UtilGetopt( argc, argv, "rbh" ) ) != -1  {
        switch  c  {
            case 'r':
                fReverse ^= 1;
                break;
            case 'b':
                fBinary ^= 1;
                break;
            case 'h':
                goto usage;
            default:
                goto usage;
        }
    }
    if pAbc.pGia == nil {
        Abc_Print( -1, "IoCommandWriteTruths(): There is no AIG.\n" );
        return 1;
    } 
    if Gia_ManPiNum(pAbc.pGia) > 16 {
        Abc_Print( -1, "IoCommandWriteTruths(): Can write truth tables up to 16 inputs.\n" );
        return 0;
    }
    if Gia_ManPiNum(pAbc.pGia) < 3 {
        Abc_Print( -1, "IoCommandWriteTruths(): Can write truth tables for 3 inputs or more.\n" );
        return 0;
    }
    if argc != globalUtilOptind + 1 {
        goto usage; }
    // get the input file name
    pFileName = argv[globalUtilOptind];
    // convert to logic
    pFile = os.OpenFile(pFileName, os.O_WRONLY | os.O_CREATE | os.O_TRUNC, 0666);
    if pFile == nil {
        fmt.Printf("Cannot open file \"%s\" for writing.\n", pFileName);
        return 0;
    }
    nBytes = 8 * Abc_Truth6WordNum( Gia_ManPiNum(pAbc.pGia) );
    Gia_ManForEachCo( pAbc.pGia, pObj, i )
    {
        pTruth = Gia_ObjComputeTruthTable( pAbc.pGia, pObj );
        if fBinary != 0 {
            fwrite( pTruth, nBytes, 1, pFile );
        } else {
            Extra_PrintHex( pFile, (unsigned *)pTruth, Gia_ManPiNum(pAbc.pGia) );
fmt.Fprintf(pFile, "\n"); }
    }
    _ = (pFile).Close();
    return 0;

usage:
    fmt.Fprintf(pAbc.Err, "usage: &write_truths [-rbh] <file>\n");
    fmt.Fprintf(pAbc.Err, "\t         writes truth tables of each PO of GIA manager into a file\n");
    fmt.Fprintf(pAbc.Err, "\t-r     : toggle reversing bits in the truth table [default = %s]\n", fReverse? "yes":"no");
    fmt.Fprintf(pAbc.Err, "\t-b     : toggle using binary format [default = %s]\n", fBinary? "yes":"no");
    fmt.Fprintf(pAbc.Err, "\t-h     : print the help massage\n");
    fmt.Fprintf(pAbc.Err, "\tfile   : the name of the file to write\n");
    return 1;
}


/**Function*************************************************************

  Synopsis    []

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func IoCommandWriteStatus(pAbc *Abc_Frame_t, argc int32, argv *string) int32 {
    var pFileName string;
    var c int32;
    extern void Abc_NtkWriteLogFile( char * pFileName, Abc_Cex_t * pCex, int Status, int nFrames, char * pCommand );

    Extra_UtilGetoptReset();
    for  ( c = Extra_UtilGetopt( argc, argv, "h" ) ) != -1  {
        switch  c  {
            case 'h':
                goto usage;
            default:
                goto usage;
        }
    }
    if argc != globalUtilOptind + 1 {
        goto usage; }
    // get the input file name
    pFileName = argv[globalUtilOptind];
    Abc_NtkWriteLogFile( pFileName, pAbc.pCex, pAbc.Status, pAbc.nFrames, nil );
    return 0;

usage:
    fmt.Fprintf(pAbc.Err, "usage: write_status [-h] <file>\n");
    fmt.Fprintf(pAbc.Err, "\t         writes verification log file\n");
    fmt.Fprintf(pAbc.Err, "\t-h     : print the help massage\n");
    fmt.Fprintf(pAbc.Err, "\tfile   : the name of the file to write\n");
    return 1;
}

/**Function*************************************************************

  Synopsis    []

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func IoCommandWriteSmv(pAbc *Abc_Frame_t, argc int32, argv *string) int32 {
    var pFileName string;
    var c int32;

    Extra_UtilGetoptReset();
    for  ( c = Extra_UtilGetopt( argc, argv, "h" ) ) != -1  {
        switch  c  {
            case 'h':
                goto usage;
            default:
                goto usage;
        }
    }
    if pAbc.pNtkCur == nil {
        fmt.Fprintf(pAbc.Out, "Empty network.\n");
        return 0;
    }
    if argc != globalUtilOptind + 1 {
        goto usage; }
    // get the output file name
    pFileName = argv[globalUtilOptind];
    // call the corresponding file writer
    Io_Write( pAbc.pNtkCur, pFileName, IO_FILE_SMV );
    return 0;

usage:
    fmt.Fprintf(pAbc.Err, "usage: write_smv [-h] <file>\n");
    fmt.Fprintf(pAbc.Err, "\t         write the network in SMV format\n");
    fmt.Fprintf(pAbc.Err, "\t-h     : print the help message\n");
    fmt.Fprintf(pAbc.Err, "\tfile   : the name of the file to write (extension .smv)\n");
    return 1;
}

/**Function*************************************************************

  Synopsis    []

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func IoCommandWriteJson(pAbc *Abc_Frame_t, argc int32, argv *string) int32 {
    extern void Json_Write( char * pFileName, Abc_Nam_t * pStr, Vec_Wec_t * vObjs );
    extern void Json_Extract( char * pFileName, Abc_Nam_t * pStr, Vec_Wec_t * vObjs );
    var c, fExtract int32;
    var pFileName string;

    Extra_UtilGetoptReset();
    for  ( c = Extra_UtilGetopt( argc, argv, "ch" ) ) != -1  {
        switch  c  {
            case 'c':
                fExtract ^= 1;
                break;
            case 'h':
                goto usage;
            default:
                goto usage;
        }
    }
    if Abc_FrameReadJsonStrs(Abc_FrameReadGlobalFrame()) == nil {
        fmt.Fprintf(pAbc.Out, "No JSON info is available.\n");
        return 0;
    }
    if argc != globalUtilOptind + 1 {
        goto usage; }
    pFileName = argv[globalUtilOptind];
    if fExtract != 0 {
        Json_Extract( pFileName, Abc_FrameReadJsonStrs(Abc_FrameReadGlobalFrame()), Abc_FrameReadJsonObjs(Abc_FrameReadGlobalFrame()) );
    } else {
        Json_Write( pFileName, Abc_FrameReadJsonStrs(Abc_FrameReadGlobalFrame()), Abc_FrameReadJsonObjs(Abc_FrameReadGlobalFrame()) ); }
    return 0;

usage:
    fmt.Fprintf(pAbc.Err, "usage: write_json [-ch] <file>\n");
    fmt.Fprintf(pAbc.Err, "\t         write the network in JSON format\n");
    fmt.Fprintf(pAbc.Err, "\t-c     : output extracted version\n");
    fmt.Fprintf(pAbc.Err, "\t-h     : print the help message\n");
    fmt.Fprintf(pAbc.Err, "\tfile   : the name of the file to write (extension .json)\n");
    return 1;
}

////////////////////////////////////////////////////////////////////////
///                       END OF FILE                                ///
////////////////////////////////////////////////////////////////////////


//ABC_NAMESPACE_IMPL_END
