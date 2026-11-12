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


////////////////////////////////////////////////////////////////////////
///                       END OF FILE                                ///
////////////////////////////////////////////////////////////////////////


//ABC_NAMESPACE_IMPL_END
