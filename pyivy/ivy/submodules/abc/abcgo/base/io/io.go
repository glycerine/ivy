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
