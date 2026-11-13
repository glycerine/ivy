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


//ABC_NAMESPACE_IMPL_END

//#include "proof/fra/fra.h"

//ABC_NAMESPACE_IMPL_START



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


////////////////////////////////////////////////////////////////////////
///                       END OF FILE                                ///
////////////////////////////////////////////////////////////////////////


//ABC_NAMESPACE_IMPL_END
