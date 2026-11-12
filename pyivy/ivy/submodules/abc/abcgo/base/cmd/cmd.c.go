package cmd

import (
	"fmt"
	"strconv"
	"strings"
)

import "os"

/**CFile****************************************************************

  FileName    [cmd.c]

  SystemName  [ABC: Logic synthesis and verification system.]

  PackageName [Command processing package.]

  Synopsis    [Command file.]

  Author      [Alan Mishchenko]
  
  Affiliation [UC Berkeley]

  Date        [Ver. 1.0. Started - June 20, 2005.]

  Revision    [$Id: cmd.c,v 1.00 2005/06/20 00:00:00 alanmi Exp $]

***********************************************************************/

#ifdef WIN32
#include <process.h>
#else
#include <unistd.h>
#endif

//#include "base/abc/abc.h"
//#include "base/main/mainInt.h"
//#include "cmdInt.h"
//#include "misc/util/utilSignal.h"

//ABC_NAMESPACE_IMPL_START

////////////////////////////////////////////////////////////////////////
///                        DECLARATIONS                              ///
////////////////////////////////////////////////////////////////////////

// static int CmdCommandTime          ( Abc_Frame_t * pAbc, int argc, char ** argv );

// static int CmdCommandEcho          ( Abc_Frame_t * pAbc, int argc, char ** argv );

// static int CmdCommandQuit          ( Abc_Frame_t * pAbc, int argc, char ** argv );

// static int CmdCommandAbcrc         ( Abc_Frame_t * pAbc, int argc, char ** argv );

// static int CmdCommandHistory       ( Abc_Frame_t * pAbc, int argc, char ** argv );

// static int CmdCommandAlias         ( Abc_Frame_t * pAbc, int argc, char ** argv );

// static int CmdCommandUnalias       ( Abc_Frame_t * pAbc, int argc, char ** argv );

// static int CmdCommandHelp          ( Abc_Frame_t * pAbc, int argc, char ** argv );

// static int CmdCommandSource        ( Abc_Frame_t * pAbc, int argc, char ** argv );

// static int CmdCommandSetVariable   ( Abc_Frame_t * pAbc, int argc, char ** argv );

// static int CmdCommandUnsetVariable ( Abc_Frame_t * pAbc, int argc, char ** argv );

// static int CmdCommandUndo          ( Abc_Frame_t * pAbc, int argc, char ** argv );

// static int CmdCommandRecall        ( Abc_Frame_t * pAbc, int argc, char ** argv );

// static int CmdCommandEmpty         ( Abc_Frame_t * pAbc, int argc, char ** argv );

#if defined(WIN32) && !defined(__cplusplus)
static int CmdCommandScanDir       ( Abc_Frame_t * pAbc, int argc, char ** argv );
static int CmdCommandRenameFiles   ( Abc_Frame_t * pAbc, int argc, char ** argv );
static int CmdCommandLs            ( Abc_Frame_t * pAbc, int argc, char ** argv );
static int CmdCommandScrGen        ( Abc_Frame_t * pAbc, int argc, char ** argv );
#endif
// static int CmdCommandVersion       ( Abc_Frame_t * pAbc, int argc, char ** argv );

// static int CmdCommandSis           ( Abc_Frame_t * pAbc, int argc, char ** argv );

// static int CmdCommandMvsis         ( Abc_Frame_t * pAbc, int argc, char ** argv );

// static int CmdCommandCapo          ( Abc_Frame_t * pAbc, int argc, char ** argv );

// static int CmdCommandStarter       ( Abc_Frame_t * pAbc, int argc, char ** argv );

// static int CmdCommandAutoTuner     ( Abc_Frame_t * pAbc, int argc, char ** argv );


// extern int Cmd_CommandAbcLoadPlugIn( Abc_Frame_t * pAbc, int argc, char ** argv );


////////////////////////////////////////////////////////////////////////
///                     FUNCTION DEFINITIONS                         ///
////////////////////////////////////////////////////////////////////////

/**Function********************************************************************

  Synopsis    [Initializes the command package.]

  SideEffects [Commands are added to the command table.]

  SeeAlso     [Cmd_End]

******************************************************************************/
func Cmd_Init(pAbc *Abc_Frame_t) {
    pAbc.tCommands = st__init_table(nil, nil);
    pAbc.tAliases  = st__init_table(nil, nil);
    pAbc.tFlags    = st__init_table(nil, nil);
    pAbc.aHistory  = Vec_PtrAlloc( 100 );
    Cmd_HistoryRead( pAbc );

    Cmd_CommandAdd( pAbc, "Basic", "time",          CmdCommandTime,            0 );
    Cmd_CommandAdd( pAbc, "Basic", "echo",          CmdCommandEcho,            0 );
    Cmd_CommandAdd( pAbc, "Basic", "quit",          CmdCommandQuit,            0 );
    Cmd_CommandAdd( pAbc, "Basic", "abcrc",         CmdCommandAbcrc,           0 );
    Cmd_CommandAdd( pAbc, "Basic", "history",       CmdCommandHistory,         0 );
    Cmd_CommandAdd( pAbc, "Basic", "alias",         CmdCommandAlias,           0 );
    Cmd_CommandAdd( pAbc, "Basic", "unalias",       CmdCommandUnalias,         0 );
    Cmd_CommandAdd( pAbc, "Basic", "help",          CmdCommandHelp,            0 );
    Cmd_CommandAdd( pAbc, "Basic", "source",        CmdCommandSource,          0 );
    Cmd_CommandAdd( pAbc, "Basic", "set",           CmdCommandSetVariable,     0 );
    Cmd_CommandAdd( pAbc, "Basic", "unset",         CmdCommandUnsetVariable,   0 );
    Cmd_CommandAdd( pAbc, "Basic", "undo",          CmdCommandUndo,            0 );
    Cmd_CommandAdd( pAbc, "Basic", "recall",        CmdCommandRecall,          0 );
    Cmd_CommandAdd( pAbc, "Basic", "empty",         CmdCommandEmpty,           0 );
#if defined(WIN32) && !defined(__cplusplus)
    Cmd_CommandAdd( pAbc, "Basic", "scandir",       CmdCommandScanDir,         0 );
    Cmd_CommandAdd( pAbc, "Basic", "renamefiles",   CmdCommandRenameFiles,     0 );
    Cmd_CommandAdd( pAbc, "Basic", "ls",            CmdCommandLs,              0 );
    Cmd_CommandAdd( pAbc, "Basic", "scrgen",        CmdCommandScrGen,          0 );
#endif
    Cmd_CommandAdd( pAbc, "Basic", "version",       CmdCommandVersion,         0 );

    Cmd_CommandAdd( pAbc, "Various", "sis",         CmdCommandSis,             1 );
    Cmd_CommandAdd( pAbc, "Various", "mvsis",       CmdCommandMvsis,           1 );
    Cmd_CommandAdd( pAbc, "Various", "capo",        CmdCommandCapo,            0 );
    Cmd_CommandAdd( pAbc, "Various", "starter",     CmdCommandStarter,         0 );
    Cmd_CommandAdd( pAbc, "Various", "autotuner",   CmdCommandAutoTuner,       0 );

    Cmd_CommandAdd( pAbc, "Various", "load_plugin", Cmd_CommandAbcLoadPlugIn,  0 );
}

func CmdCommandDispatchBasicById(Id Cmd_CommandId_t, pAbc *Abc_Frame_t, argc int32, argv *string, pResult *int32) int32 {
    switch  Id  {
    case CMD_COMMAND_ALIAS:
        *pResult = CmdCommandAlias( pAbc, argc, argv );
        return 1;
    case CMD_COMMAND_SOURCE:
        *pResult = CmdCommandSource( pAbc, argc, argv );
        return 1;
    case CMD_COMMAND_SET:
        *pResult = CmdCommandSetVariable( pAbc, argc, argv );
        return 1;
    default:
        return 0;
    }
}

/**Function********************************************************************

  Synopsis    [Ends the command package.]

  Description [Ends the command package. Tables are freed.]

  SideEffects []

  SeeAlso     []

******************************************************************************/
func Cmd_End(pAbc *Abc_Frame_t) {
    var gen *st__generator;
    var pKey, pValue string;
    Cmd_HistoryWrite( pAbc, ABC_INFINITY );

//    st__free_table( pAbc->tCommands, (void (*)()) 0, CmdCommandFree );
//    st__free_table( pAbc->tAliases,  (void (*)()) 0, CmdCommandAliasFree );
//    st__free_table( pAbc->tFlags,    free, free );

    st__foreach_item( pAbc.tCommands, gen, (char **)&pKey, (char **)&pValue )
        CmdCommandFree( (Abc_Command *)pValue );
    st__free_table( pAbc.tCommands );

    st__foreach_item( pAbc.tAliases, gen, (char **)&pKey, (char **)&pValue )
        CmdCommandAliasFree( (Abc_Alias *)pValue );
    st__free_table( pAbc.tAliases );

    st__foreach_item( pAbc.tFlags, gen, (char **)&pKey, (char **)&pValue )
        ABC_FREE( pKey );
ABC_FREE( pValue );
    st__free_table( pAbc.tFlags );

    Vec_PtrFreeFree( pAbc.aHistory );
}



/**Function********************************************************************

  Synopsis    []

  Description []

  SideEffects []

  SeeAlso     []

******************************************************************************/
func CmdCommandTime(pAbc *Abc_Frame_t, argc int32, argv *string) int32 {
    var c int32;
    var fClear int32;

    fClear = 0;
    Extra_UtilGetoptReset();
    for  ( c = Extra_UtilGetopt( argc, argv, "ch" ) ) != -1  {
        switch  c  {
        case 'c':
            fClear ^= 1;
            break;
        case 'h':
            goto usage;
        default:
            goto usage;
        }
    }

    if fClear != 0 {
        pAbc.TimeTotal += pAbc.TimeCommand;
        pAbc.TimeCommand = 0.0;
        return 0;
    }

    if argc != globalUtilOptind {
        goto usage;
    }


    pAbc.TimeTotal += pAbc.TimeCommand;
    fmt.Fprintf(pAbc.Out, "elapse: %3.2f seconds, total: %3.2f seconds\n", pAbc.TimeCommand, pAbc.TimeTotal);
/*
    {
        FILE * pTable;
        pTable = fopen( "runtimes.txt", "a+" );
        fprintf( pTable, "%4.2f\n", pAbc->TimeCommand );
        fclose( pTable );
    }
*/
    pAbc.TimeCommand = 0.0;
    return 0;

  usage:
    fmt.Fprintf(pAbc.Err, "usage: time [-ch]\n");
    fmt.Fprintf(pAbc.Err, "      \t\tprint the runtime since the last call\n");
    fmt.Fprintf(pAbc.Err, "   -c \t\tclears the elapsed time without printing it\n");
    fmt.Fprintf(pAbc.Err, "   -h \t\tprint the command usage\n");
    return 1;
}

/**Function********************************************************************

  Synopsis    []

  Description []

  SideEffects []

  SeeAlso     []

******************************************************************************/
func CmdCommandEcho(pAbc *Abc_Frame_t, argc int32, argv *string) int32 {
    var i int32;
    var c int32;
    var n int32  = 1;

    Extra_UtilGetoptReset();
    for  ( c = Extra_UtilGetopt( argc, argv, "hn" ) ) != -1  {
        switch  c  {
        case 'n':
            n = 0;
            break;
        case 'h':
            goto usage;
            break;
        default:
            goto usage;
        }
    }

    if pAbc.Out == os.Stdout{
        for  i = globalUtilOptind; i < argc; i++  {
            Abc_Print( 1, "%s ", argv[i] );
        }
        if n != 0 {
            Abc_Print( 1, "\n" ); }

    }else{
        for  i = globalUtilOptind; i < argc; i++  {
            fmt.Fprintf(pAbc.Out, "%s ", argv[i]);
        }
        if n != 0 {
            fmt.Fprintf(pAbc.Out, "\n"); }

        fflush ( pAbc.Out );
    }
    return 0;

  usage:
    fmt.Fprintf(pAbc.Err, "usage: echo [-h] string \n");
    fmt.Fprintf(pAbc.Err, "   -n \t\tsuppress newline at the end\n");
    fmt.Fprintf(pAbc.Err, "   -h \t\tprint the command usage\n");
    return ( 1 );
}

/**Function********************************************************************

  Synopsis    []

  Description []

  SideEffects []

  SeeAlso     []

******************************************************************************/
func CmdCommandQuit(pAbc *Abc_Frame_t, argc int32, argv *string) int32 {
    var c int32;

    Extra_UtilGetoptReset();
    for  ( c = Extra_UtilGetopt( argc, argv, "hs" ) ) != -1  {
        switch  c  {
        case 'h':
            goto usage;
            break;
        case 's':
            return -2;
            break;
        default:
            goto usage;
        }
    }

    if argc != globalUtilOptind {
        goto usage; }
    return -1;

  usage:
    fmt.Fprintf(pAbc.Err, "usage: quit [-sh]\n");
    fmt.Fprintf(pAbc.Err, "   -h  print the command usage\n");
    fmt.Fprintf(pAbc.Err, "   -s  frees all the memory before quitting\n");
    return 1;
}

/**Function********************************************************************

  Synopsis    []

  Description []

  SideEffects []

  SeeAlso     []

******************************************************************************/
func CmdCommandAbcrc(pAbc *Abc_Frame_t, argc int32, argv *string) int32 {
    var c int32;

    Extra_UtilGetoptReset();
    for  ( c = Extra_UtilGetopt( argc, argv, "h" ) ) != -1  {
        switch  c  {
        case 'h':
            goto usage;
            break;
        default:
            goto usage;
        }
    }
    Abc_UtilsSource( pAbc );
    return 0;

  usage:
    fmt.Fprintf(pAbc.Err, "usage: abcrc [-h]\n");
    fmt.Fprintf(pAbc.Err, "   -h  sources \"abc.rc\" from the current/parent/grandparent directory\n");
    return 1;
}

/**Function********************************************************************

  Synopsis    []

  Description []

  SideEffects []

  SeeAlso     []

******************************************************************************/
func CmdCommandHistory(pAbc *Abc_Frame_t, argc int32, argv *string) int32 {
    var pName string;
    var i, c int32;
    var nPrints int32  = 20;
    Extra_UtilGetoptReset();
    for  ( c = Extra_UtilGetopt( argc, argv, "h" ) ) != -1  {
        switch  c  {
            case 'h':
                goto usage;
            default :
                goto usage;
        }
    }
    if argc > globalUtilOptind + 1 {
        goto usage; }
    // get the number from the command line
    if argc == globalUtilOptind + 1 {
        nPrints = strconv.Atoi(argv[globalUtilOptind]); }
    // print the commands
    Vec_PtrForEachEntryStart( char *, pAbc.aHistory, pName, i, Abc_MaxInt(0, Vec_PtrSize(pAbc.aHistory)-nPrints) )
        fmt.Fprintf(pAbc.Out, "%2d : %s\n", Vec_PtrSize(pAbc.aHistory)-i, pName);
    return 0;

usage:
    fmt.Fprintf(pAbc.Err, "usage: history [-h] <num>\n");
    fmt.Fprintf(pAbc.Err, "\t        lists the last commands entered on the command line\n");
    fmt.Fprintf(pAbc.Err, "\t-h    : print the command usage\n");
    fmt.Fprintf(pAbc.Err, "\t<num> : the maximum number of entries to show [default = %d]\n", nPrints);
    return ( 1 );
}

/**Function********************************************************************

  Synopsis    []

  Description []

  SideEffects []

  SeeAlso     []

******************************************************************************/
func CmdCommandAlias(pAbc *Abc_Frame_t, argc int32, argv *string) int32 {
    char *key;
    var value string;
    var c int32;

    Extra_UtilGetoptReset();
    for  ( c = Extra_UtilGetopt( argc, argv, "h" ) ) != -1  {
        switch  c  {
        case 'h':
            goto usage;
            break;
        default:
            goto usage;
        }
    }


    if argc == 1 {
        CmdPrintTable( pAbc.tAliases, 1 );
        return 0;

    
    } else if argc == 2 {
        if st__lookup( pAbc.tAliases, argv[1], &value ) {
            CmdCommandAliasPrint( pAbc, ( Abc_Alias * ) value ); }
        return 0;
    }

    // delete any existing alias
    key = argv[1];
    if st__delete( pAbc.tAliases, &key, &value ) {
        CmdCommandAliasFree( ( Abc_Alias * ) value ); }
    CmdCommandAliasAdd( pAbc, argv[1], argc - 2, argv + 2 );
    return 0;

usage:
    fmt.Fprintf(pAbc.Err, "usage: alias [-h] [command [string]]\n");
    fmt.Fprintf(pAbc.Err, "   -h \t\tprint the command usage\n");
    return ( 1 );
}

/**Function********************************************************************

  Synopsis    []

  Description []

  SideEffects []

  SeeAlso     []

******************************************************************************/
func CmdCommandUnalias(pAbc *Abc_Frame_t, argc int32, argv *string) int32 {
    var i int32;
    char *key;
    var value string;
    var c int32;

    Extra_UtilGetoptReset();
    for  ( c = Extra_UtilGetopt( argc, argv, "h" ) ) != -1  {
        switch  c  {
        case 'h':
            goto usage;
            break;
        default:
            goto usage;
        }
    }

    if argc < 2 {
        goto usage;
    }

    for  i = 1; i < argc; i++  {
        key = argv[i];
        if st__delete( pAbc.tAliases, &key, &value ) {
            CmdCommandAliasFree( ( Abc_Alias * ) value );
        }
    }
    return 0;

  usage:
    fmt.Fprintf(pAbc.Err, "usage: unalias [-h] alias_names\n");
    fmt.Fprintf(pAbc.Err, "   -h \t\tprint the command usage\n");
    return 1;
}

/**Function********************************************************************

  Synopsis    []

  Description []

  SideEffects []

  SeeAlso     []

******************************************************************************/
func CmdCommandHelp(pAbc *Abc_Frame_t, argc int32, argv *string) int32 {
    var fPrintAll, fDetails int32;
    var c int32;

    fPrintAll = 0;
    fDetails = 0;
    Extra_UtilGetoptReset();
    for  ( c = Extra_UtilGetopt( argc, argv, "adh" ) ) != -1  {
        switch  c  {
        case 'a',
             'v':
                fPrintAll ^= 1;
                break;
            break;
        case 'd':
            fDetails ^= 1;
            break;
        case 'h':
            goto usage;
            break;
        default:
            goto usage;
        }
    }

    if argc != globalUtilOptind {
        goto usage; }

    CmdCommandPrint( pAbc, fPrintAll, fDetails );
    return 0;

usage:
    fmt.Fprintf(pAbc.Err, "usage: help [-a] [-d] [-h]\n");
    fmt.Fprintf(pAbc.Err, "       prints the list of available commands by group\n");
    fmt.Fprintf(pAbc.Err, " -a       toggle printing hidden commands [default = %s]\n", fPrintAll? "yes": "no");
    fmt.Fprintf(pAbc.Err, " -d       print usage details to all commands [default = %s]\n", fDetails? "yes": "no");
    fmt.Fprintf(pAbc.Err, " -h       print the command usage\n");
    return 1;
}

/**Function********************************************************************

  Synopsis    []

  Description []

  SideEffects []

  SeeAlso     []

******************************************************************************/
func CmdCommandSource(pAbc *Abc_Frame_t, argc int32, argv *string) int32 {
    var c, echo, prompt, silent, interactive, quit_count, lp_count int32;
    var status int32;             /* initialize so that lint doesn't complain */
    var lp_file_index, did_subst int32;
    var prompt_string string;
    var real_filename string;
    var line [ABC_MAX_STR]byte;
    var command string;
    var fp *os.File;

    interactive = silent = prompt = echo = 0;

    Extra_UtilGetoptReset();
    for  ( c = Extra_UtilGetopt( argc, argv, "ipsxh" ) ) != -1  {
        switch  c  {
        case 'i':               /* a hack to distinguish EOF from stdin */
            interactive = 1;
            break;
        case 'p':
            prompt ^= 1;
            break;
        case 's':
            silent ^= 1;
            break;
        case 'x':
            echo ^= 1;
            break;
        case 'h':
            goto usage;
        default:
            goto usage;
        }
    }

    /* added to avoid core-dumping when no script file is specified */
    if argc == globalUtilOptind {
        goto usage;
    }

    lp_file_index = globalUtilOptind;
    lp_count = 0;

    /*
     * FIX (Tom, 5/7/95):  I'm not sure what the purpose of this outer do loop
     * is. In particular, lp_file_index is never modified in the loop, so it
     * looks it would just read the same file over again.  Also, SIS had
     * lp_count initialized to -1, and hence, any file sourced by SIS (if -l or
     * -t options on "source" were used in SIS) would actually be executed twice.
     */
    pAbc.fSource = 1;
    do
    {
        var pFileName, pTemp string;

        // get the input file name
        pFileName = argv[lp_file_index];
        // fix the wrong symbol
        for  pTemp = pFileName; *pTemp; pTemp++  {
            if *pTemp == '>' {
                *pTemp = '\\'; }
        }

        lp_count++;             /* increment the loop counter */

        fp = CmdFileOpen( pAbc, pFileName, "r", &real_filename, silent );
        if fp == nil {
            pAbc.fSource = 0;
            ABC_FREE( real_filename );
            return !silent;     /* error return if not silent */
        }

        quit_count = 0;
        do
        {
            if prompt {
                prompt_string = Cmd_FlagReadByName( pAbc, "prompt" );
                if prompt_string == nil {
                    prompt_string = "abc> "; }

            
            } else {
                prompt_string = nil;
            }

            /* clear errors -- e.g., EOF reached from stdin */
            clearerr( fp );

            /* read another command line */
            if fgets( line, ABC_MAX_STR, fp ) == nil {
                if interactive {
                    if quit_count++ < 5 {
                        fmt.Fprintf(pAbc.Err, "\nUse \"quit\" to leave ABC.\n");
                        continue;
                    }
                    status = -1;    /* fake a 'quit' */
                
                } else {
                    status = 0; /* successful end of 'source' ; loop? */
                }
                break;
            }
            quit_count = 0;

            if echo {
                fprintf( pAbc.Out, "abc - > %s", line );
            }
            command = CmdHistorySubstitution( pAbc, line, &did_subst );
            if command == nil {
                status = 1;
                break;
            }
            if did_subst != 0 {
                if interactive {
                    fmt.Fprintf(pAbc.Out, "%s\n", command);
                }
            }
            if command != line {
                strcpy( line, command );
            }
            if interactive && *line != '\0' {
                Cmd_HistoryAddCommand( pAbc, line );
                if pAbc.Hst != nil {
                    fprintf( pAbc.Hst, "%s\n", line );
                    fflush( pAbc.Hst );
                }
            }

            fflush( pAbc.Out );
            status = Cmd_CommandExecute( pAbc, line );
        }
        while ( status == 0 );

        if fp != os.Stdin {
            if status > 0 {
                fmt.Fprintf(pAbc.Err, "** cmd error: aborting 'source %s'\n", real_filename);
            }
            _ = (fp).Close();
        }
        ABC_FREE( real_filename );

    }
    while ( ( status == 0 ) && ( lp_count <= 0 ) );
    pAbc.fSource = 0;
    return status;

  usage:
    fmt.Fprintf(pAbc.Err, "usage: source [-psxh] <file_name>\n");
    fmt.Fprintf(pAbc.Err, "\t-p     supply prompt before reading each line [default = %s]\n", prompt? "yes": "no");
    fmt.Fprintf(pAbc.Err, "\t-s     silently ignore nonexistant file [default = %s]\n", silent? "yes": "no");
    fmt.Fprintf(pAbc.Err, "\t-x     echo each line as it is executed [default = %s]\n", echo? "yes": "no");
    fmt.Fprintf(pAbc.Err, "\t-h     print the command usage\n");
    return 1;
}

/**Function********************************************************************

  Synopsis    []

  Description []

  SideEffects []

  SeeAlso     []

******************************************************************************/
func CmdCommandSetVariable(pAbc *Abc_Frame_t, argc int32, argv *string) int32 {
    var flag_value, value string;
    char* key;
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
    if argc == 0 || argc > 3 {
        goto usage;
    
    } else if argc == 1 {
        CmdPrintTable( pAbc.tFlags, 0 );
        return 0;
    
    } else {
        key = argv[1];
        if st__delete( pAbc.tFlags, &key, &value ) {
            ABC_FREE( key );
            ABC_FREE( value );
        }

        if argc == 2 {
            flag_value = Extra_UtilStrsav( "" );
        } else {
            flag_value = Extra_UtilStrsav( argv[2] );
        }
//        flag_value = argc == 2 ? NULL : Extra_UtilStrsav(argv[2]);
        st__insert( pAbc.tFlags, Extra_UtilStrsav(argv[1]), flag_value );

        if int32(strings.Compare(argv[1], "abcout")) == 0 {
            if pAbc.Out != os.Stdout {
                _ = (pAbc.Out).Close(); }
            if int32(strings.Compare(flag_value, "")) == 0 {
                flag_value = "-"; }
            pAbc.Out = CmdFileOpen( pAbc, flag_value, "w", nil, 0 );
            if pAbc.Out == nil {
                pAbc.Out = os.Stdout; }
#if HAVE_SETVBUF
            setvbuf( pAbc->Out, ( char * ) NULL, _IOLBF, 0 );
#endif
        }
        if int32(strings.Compare(argv[1], "abcerr")) == 0 {
            if pAbc.Err != os.Stderr {
                _ = (pAbc.Err).Close(); }
            if int32(strings.Compare(flag_value, "")) == 0 {
                flag_value = "-"; }
            pAbc.Err = CmdFileOpen( pAbc, flag_value, "w", nil, 0 );
            if pAbc.Err == nil {
                pAbc.Err = os.Stderr; }
#if HAVE_SETVBUF
            setvbuf( pAbc->Err, ( char * ) NULL, _IOLBF, 0 );
#endif
        }
        if int32(strings.Compare(argv[1], "history")) == 0 {
            if pAbc.Hst != nil {
                _ = (pAbc.Hst).Close(); }
            if int32(strings.Compare(flag_value, "")) == 0 {
                pAbc.Hst = nil;
            } else {
                pAbc.Hst = CmdFileOpen( pAbc, flag_value, "w", nil, 0 );
                if pAbc.Hst == nil {
                    pAbc.Hst = nil; }
            }
        }
        return 0;
    }

  usage:
    fmt.Fprintf(pAbc.Err, "usage: set [-h] <name> <value>\n");
    fmt.Fprintf(pAbc.Err, "\t        sets the value of parameter <name>\n");
    fmt.Fprintf(pAbc.Err, "\t-h    : print the command usage\n");
    return 1;

}

/**Function********************************************************************

  Synopsis    []

  Description []

  SideEffects []

  SeeAlso     []

******************************************************************************/
func CmdCommandUnsetVariable(pAbc *Abc_Frame_t, argc int32, argv *string) int32 {
    var i int32;
    char *key;
    var value string;
    var c int32;

    Extra_UtilGetoptReset();
    for  ( c = Extra_UtilGetopt( argc, argv, "h" ) ) != -1  {
        switch  c  {
        case 'h':
            goto usage;
            break;
        default:
            goto usage;
        }
    }

    if argc < 2 {
        goto usage;
    }

    for  i = 1; i < argc; i++  {
        key = argv[i];
        if st__delete( pAbc.tFlags, &key, &value ) {
            ABC_FREE( key );
            ABC_FREE( value );
        }
    }
    return 0;


  usage:
    fmt.Fprintf(pAbc.Err, "usage: unset [-h] <name> \n");
    fmt.Fprintf(pAbc.Err, "\t        removes the value of parameter <name>\n");
    fmt.Fprintf(pAbc.Err, "\t-h    : print the command usage\n");
    return 1;
}

/**Function********************************************************************

  Synopsis    []

  Description []

  SideEffects []

  SeeAlso     []

******************************************************************************/
func CmdCommandUndo(pAbc *Abc_Frame_t, argc int32, argv *string) int32 {
    if argc == 2 && !(int32(strings.Compare(argv[1], "-h")) != 0) {
        goto usage; }

    if pAbc.pNtkCur == nil {
        fmt.Fprintf(pAbc.Out, "Empty network.\n");
        return 0;
    }

    // if there are no arguments on the command line
    // set the current network to be the network from the previous step
    if argc == 1 {
        return CmdCommandRecall( pAbc, argc, argv ); }

usage:
    fmt.Fprintf(pAbc.Err, "usage: undo\n");
    fmt.Fprintf(pAbc.Err, "         sets the current network to be the previously saved network\n");
    return 1;

}

/**Function********************************************************************

  Synopsis    []

  Description []

  SideEffects []

  SeeAlso     []

******************************************************************************/
func CmdCommandRecall(pAbc *Abc_Frame_t, argc int32, argv *string) int32 {
    var pNtk *Abc_Ntk_t;
    var iStep, iStepFound int32;
    var nNetsToSave, c int32;
    var pValue string;
    var iStepStart, iStepStop int32;

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

    // get the number of networks to save
    pValue = Cmd_FlagReadByName( pAbc, "savesteps" );
    // if the value of steps to save is not set, assume 1-level undo
    if pValue == nil {
        nNetsToSave = 1;
    } else {
        nNetsToSave = strconv.Atoi(pValue); }

    // if there are no arguments on the command line
    // set the current network to be the network from the previous step
    if argc == 1 {
        // get the previously saved network
        pNtk = Abc_NtkBackup(pAbc.pNtkCur);
        if pNtk == nil {
            fmt.Fprintf(pAbc.Out, "There is no previously saved network.\n");
        } else { // set the current network to be the copy of the previous one
            Abc_FrameSetCurrentNetwork( pAbc, Abc_NtkDup(pNtk) ); }
         return 0;
    }
    if argc == 2 { // the second argument is the number of the step to return to
    
        // read the number of the step to return to
        iStep = strconv.Atoi(argv[1]);
        // check whether it is reasonable
        if iStep >= pAbc.nSteps {
            iStepStart = pAbc.nSteps - nNetsToSave;
            if iStepStart <= 0 {
                iStepStart = 1; }
            iStepStop  = pAbc.nSteps;
            if iStepStop <= 0 {
                iStepStop = 1; }
            if iStepStart == iStepStop {
                fmt.Fprintf(pAbc.Out, "Can only recall step %d.\n", iStepStop);
            } else {
                fmt.Fprintf(pAbc.Out, "Can only recall steps %d-%d.\n", iStepStart, iStepStop); }
        
        } else if iStep < 0 {
            fmt.Fprintf(pAbc.Out, "Cannot recall step %d.\n", iStep);
        } else if iStep == 0 {
            Abc_FrameDeleteAllNetworks( pAbc );
        } else {
            // scroll backward through the list of networks
            // to determine if such a network exist
            iStepFound = 0;
            for  pNtk = pAbc.pNtkCur; (pNtk != nil); pNtk = Abc_NtkBackup(pNtk)  {
                if (iStepFound = Abc_NtkStep(pNtk)) == iStep {
                    break; }
            }
            if pNtk == nil {
                iStepStart = iStepFound;
                if iStepStart <= 0 {
                    iStepStart = 1; }
                iStepStop  = pAbc.nSteps;
                if iStepStop <= 0 {
                    iStepStop = 1; }
                if iStepStart == iStepStop {
                    fmt.Fprintf(pAbc.Out, "Can only recall step %d.\n", iStepStop);
                } else {
                    fmt.Fprintf(pAbc.Out, "Can only recall steps %d-%d.\n", iStepStart, iStepStop); }
            
            } else {
                Abc_FrameSetCurrentNetwork( pAbc, Abc_NtkDup(pNtk) ); }
        }
        return 0;
    }

usage:

    fmt.Fprintf(pAbc.Err, "usage: recall -h <num>\n");
    fmt.Fprintf(pAbc.Err, "         set the current network to be one of the previous networks\n");
    fmt.Fprintf(pAbc.Err, "<num> :  level to return to [default = previous]\n");
    fmt.Fprintf(pAbc.Err, "   -h :  print the command usage\n");
    return 1;
}


/**Function********************************************************************

  Synopsis    []

  Description []

  SideEffects []

  SeeAlso     []

******************************************************************************/
func CmdCommandEmpty(pAbc *Abc_Frame_t, argc int32, argv *string) int32 {
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

    Abc_FrameDeleteAllNetworks( pAbc );
    Abc_FrameRestart( pAbc );
    return 0;
usage:

    fmt.Fprintf(pAbc.Err, "usage: empty [-h]\n");
    fmt.Fprintf(pAbc.Err, "         removes all the currently stored networks\n");
    fmt.Fprintf(pAbc.Err, "   -h :  print the command usage\n");
    return 1;
}


     

                                                                               

                                 

                

                

                

                                                                               
                                                               
 
                         
              

                                                                
     
                    
         
                 
                       
                  
                
                       
         
     
                                   
                          
                                 
                               
     
                                      
                                                
                  
                                                            
        
                                                            

             
      
                                          
                                                                                      
             
 

      


#if defined(WIN32) && !defined(__cplusplus)
#include <direct.h>
#include <io.h>

/**Function*************************************************************

  Synopsis    [Command to print the contents of the current directory (Windows).]

  Description []

  SideEffects []

  SeeAlso     []

***********************************************************************/
int CmdCommandScanDir( Abc_Frame_t * pAbc, int argc, char **argv )
{
    struct _finddata_t c_file;
    char * pDirStr = NULL;
    char*  pDirCur = NULL;
    long   hFile;
    char   c;

    Extra_UtilGetoptReset();
    while ( (c = Extra_UtilGetopt(argc, argv, "D") ) != EOF )
    {
        switch (c)
        {
            case 'D':
                if ( globalUtilOptind >= argc )
                {
                    fprintf( pAbc->Err, "Command line switch \"-D\" should be followed by a string.\n" );
                    goto usage;
                }
                pDirStr = argv[globalUtilOptind];
                globalUtilOptind++;
                break;
            default:
                goto usage;
        }
    }


    if ( pDirStr )
    {
        if( (pDirCur = _getcwd( NULL, 0 )) == NULL )
        {
            printf( "Cannot read current directory\n" );
            return 0;
        }
        if ( _chdir(pDirStr) )
        {
            printf( "Cannot change to directory: %s\n", pDirStr );
            return 0;
        }
    }

    if( (hFile = _findfirst( "*.txt", &c_file )) == -1L )
    {
        if ( pDirStr )
            printf( "No .txt files in the current directory.\n" );
        else
            printf( "No .txt files in directory: %s\n", pDirStr );
    }
    else
    {
        do
        {
            FILE * pFile = fopen( c_file.name, "rb" );
            char * pStr1 = "Property UNDECIDED.  Time =";
            char * pStr2 = "Property proved.  Time =";
            char * pStr3 = "Time =";
            char * pBuffer, * pPlace, * pThis, * pThat;
            char FileName[100];
            float Time = 0;
            // get the file name
            sprintf( FileName, "%s", c_file.name );
            pThis = strrchr( FileName, '_' );
            pThat = strchr( FileName, '.' );
            if ( pThis == NULL || pThat == NULL || pThis >= pThat )
            {
//                printf( "Something is wrong with the file name %s\n", c_file.name );
                continue;
            }
            *pThat = 0;
            pThis++;
            // get the time
            if ( pFile == NULL )
            {
                printf( "Cannot open file %s\n", c_file.name );
                continue;
            }
            fclose( pFile );
            pBuffer = Extra_FileReadContents( c_file.name );
            pPlace = strstr( pBuffer, pStr1 );
            if ( pPlace == NULL )
            {
                pPlace = strstr( pBuffer, pStr2 );
                if ( pPlace == NULL )
                {
                    pPlace = strstr( pBuffer, pStr3 );
                    if ( pPlace == NULL )
                    {
//                        printf( "Cannot find substrings in file %s\n", c_file.name );
                        ABC_FREE( pBuffer );
                        continue;
                    }
                    else
                        pPlace += strlen( pStr3 );
                }
                else
                    pPlace += strlen( pStr2 );
            }
            else
                pPlace += strlen( pStr1 );
            sscanf( pPlace, "%f", &Time );
            printf( "%s %.2f\n", pThis, Time );
            ABC_FREE( pBuffer );
        }
        while( _findnext( hFile, &c_file ) == 0 );
        _findclose( hFile );
    }
    if ( pDirStr )
    {
        if ( _chdir(pDirCur) )
        {
            ABC_FREE( pDirCur );
            printf( "Cannot change to directory: %s\n", pDirCur );
            return 0;
        }
        ABC_FREE( pDirCur );
    }
    return 0;

usage:
    fprintf( pAbc->Err, "usage: scandir [-D string]\n" );
    fprintf( pAbc->Err, "            performs custom scanning of the files in the given directory\n" );
    fprintf( pAbc->Err, "\t-D str  : the directory to read files from [default = current]\n" );
    return 1;
}



/**Function*************************************************************

  Synopsis    []

  Description []

  SideEffects []

  SeeAlso     []

***********************************************************************/
int CmfFindNumber( char * pName )
{
    char * pTemp;
    for ( pTemp = pName; *pTemp; pTemp++ )
        if ( *pTemp == '.' )
            break;
    if ( *pTemp == 0 )
        return -1;
    for ( --pTemp; pTemp > pName; pTemp-- )
        if ( *pTemp < '0' || *pTemp > '9' )
        {
            pTemp++;
            break;
        }
    if ( *pTemp == '.' )
        return -2;
    return atoi( pTemp );
}

/**Function*************************************************************

  Synopsis    []

  Description []

  SideEffects []

  SeeAlso     []

***********************************************************************/
void CnfDupFileUnzip( char * pOldName )
{
    extern char * Io_MvLoadFileBz2( char * pFileName, int * pnFileSize );
    char pNewName[1000];
    FILE * pFile;
    int nFileSize;
    char * pBuffer = Io_MvLoadFileBz2( pOldName, &nFileSize );
    assert( strlen(pOldName) < 1000 );
    sprintf( pNewName, "%s.v", pOldName );
    pFile = fopen( pNewName, "wb" );
    fwrite( pBuffer, nFileSize, 1, pFile );
    fclose( pFile );
    ABC_FREE( pBuffer );
}

/**Function*************************************************************

  Synopsis    [Command to print the contents of the current directory (Windows).]

  Description []

  SideEffects []

  SeeAlso     []

***********************************************************************/
int CmdCommandRenameFiles( Abc_Frame_t * pAbc, int argc, char **argv )
{
    struct _finddata_t c_file;
    long   hFile;
    char pNewName[1000];
    char * pDirStr = NULL;
    char * pDirCur = NULL;
    char * pNameNew = NULL;
    char * pNameExt = NULL;
    int c, i, nBase = 0;
    Extra_UtilGetoptReset();
    while ( (c = Extra_UtilGetopt(argc, argv, "DENB") ) != EOF )
    {
        switch (c)
        {
            case 'D':
                if ( globalUtilOptind >= argc )
                {
                    fprintf( pAbc->Err, "Command line switch \"-D\" should be followed by a string.\n" );
                    goto usage;
                }
                pDirStr = argv[globalUtilOptind];
                globalUtilOptind++;
                break;
            case 'E':
                if ( globalUtilOptind >= argc )
                {
                    fprintf( pAbc->Err, "Command line switch \"-E\" should be followed by a string.\n" );
                    goto usage;
                }
                pNameExt = argv[globalUtilOptind];
                globalUtilOptind++;
                break;
            case 'N':
                if ( globalUtilOptind >= argc )
                {
                    fprintf( pAbc->Err, "Command line switch \"-N\" should be followed by a string.\n" );
                    goto usage;
                }
                pNameNew = argv[globalUtilOptind];
                globalUtilOptind++;
                break;
            case 'B':
                if ( globalUtilOptind >= argc )
                {
                    fprintf( pAbc->Err, "Command line switch \"-B\" should be followed by a positive integer.\n" );
                    goto usage;
                }
                nBase = atoi(argv[globalUtilOptind]);
                globalUtilOptind++;
                if ( nBase < 0 )
                    goto usage;
                break;
            default:
                goto usage;
        }
    }

    if ( pNameExt == NULL )
    {
        printf( "Extension of the files should be given on the command line.\n" );
        return 0;
    }

    if ( pDirStr )
    {
        if( (pDirCur = _getcwd( NULL, 0 )) == NULL )
        {
            printf( "Cannot read current directory\n" );
            return 0;
        }
        if ( _chdir(pDirStr) )
        {
            printf( "Cannot change to directory: %s\n", pDirStr );
            return 0;
        }
    }

    sprintf( pNewName, "*.%s", pNameExt );
    if( (hFile = _findfirst( pNewName, &c_file )) == -1L )
    {
        if ( pDirStr )
            printf( "No .aig files in the current directory.\n" );
        else
            printf( "No .aig files in directory: %s\n", pDirStr );
    }
    else
    {
        char * pName, * pOldName;
        int nDigits, * pOrder;
        Vec_Ptr_t * vNames = Vec_PtrAlloc( 1000 );
        Vec_Int_t * vNums  = Vec_IntAlloc( 1000 );
        // collect names
        do {
            Vec_PtrPush( vNames, Abc_UtilStrsav(c_file.name) );
        } while( _findnext( hFile, &c_file ) == 0 );
        _findclose( hFile );
        // sort files by number
        Vec_PtrForEachEntry( char *, vNames, pName, i )
        {
            Vec_IntPush( vNums, CmfFindNumber(pName) );
            if ( Vec_IntEntryLast(vNums) < 0 )
            {
                printf( "Directory \"%s\" contains file (%s) with extension %s without number\n", pDirStr, pName, pNameExt );
                Vec_PtrFreeFree( vNames );
                Vec_IntFree( vNums );
                return 0;
            }
        }
        // sort by number
        pOrder = Abc_QuickSortCost( Vec_IntArray(vNums), Vec_IntSize(vNums), 0 );
        // rename files in that order
//        nDigits = Abc_Base10Log( nBase + Vec_IntSize(vNums) );
        nDigits = Abc_Base10Log( nBase + Vec_IntEntry(vNums, pOrder[Vec_IntSize(vNums)-1]) + 1 );
        for ( i = 0; i < Vec_IntSize(vNums); i++ )
        {
            pOldName = (char *)Vec_PtrEntry( vNames, pOrder[i] );
            sprintf( pNewName, "%s%0*d.%s", pNameNew ? pNameNew : "", nDigits, nBase+Vec_IntEntry(vNums, pOrder[i]), pNameExt );
            rename( pOldName, pNewName );
            printf( "%s -> %s\n", pOldName, pNewName );
//            CnfDupFileUnzip( pOldName );
        }
        // cleanup
        Vec_PtrFreeFree( vNames );
        Vec_IntFree( vNums );
        ABC_FREE( pOrder );
    }
    if ( pDirStr )
    {
        if ( _chdir(pDirCur) )
        {
            ABC_FREE( pDirCur );
            printf( "Cannot change to directory: %s\n", pDirCur );
            return 0;
        }
        ABC_FREE( pDirCur );
    }
    return 0;

usage:
    fprintf( pAbc->Err, "usage: renamefiles [-DEN str] [-B num]\n" );
    fprintf( pAbc->Err, "            performs renaming of files in the given directory\n" );
    fprintf( pAbc->Err, "\t-D str  : the directory to read files from [default = current]\n" );
    fprintf( pAbc->Err, "\t-E str  : the extension of files to look for [default = none]\n" );
    fprintf( pAbc->Err, "\t-N str  : the root of the resulting files [default = none]\n" );
    fprintf( pAbc->Err, "\t-B num  : the base number for all files [default = %d]\n", nBase );
    return 1;
}


/**Function*************************************************************

  Synopsis    [Command to print the contents of the current directory (Windows).]

  Description []

  SideEffects []

  SeeAlso     []

***********************************************************************/
int CmdCommandLs( Abc_Frame_t * pAbc, int argc, char **argv )
{
    struct _finddata_t c_file;
    long   hFile;
    int    fLong = 0;
    int    fOnlyBLIF = 0;
    char   Buffer[25];
    int    Counter = 0;
    int    fPrintedNewLine;
    char   c;

    Extra_UtilGetoptReset();
    while ( (c = Extra_UtilGetopt(argc, argv, "lb") ) != EOF )
    {
        switch (c)
        {
            case 'l':
              fLong = 1;
              break;
            case 'b':
              fOnlyBLIF = 1;
              break;
            default:
              goto usage;
        }
    }

    // find first .mv file in current directory
    if( (hFile = _findfirst( ((fOnlyBLIF)? "*.mv": "*.*"), &c_file )) == -1L )
    {
        if ( fOnlyBLIF )
            fprintf( pAbc->Out, "No *.mv files in the current directory.\n" );
        else
            fprintf( pAbc->Out, "No files in the current directory.\n" );
    }
    else
    {
        if ( fLong )
        {
            fprintf( pAbc->Out, " File              Date           Size |  File             Date           Size \n" );
            fprintf( pAbc->Out, " ----------------------------------------------------------------------------- \n" );
            do
            {
                strcpy( Buffer, ctime( &(c_file.time_write) ) );
                Buffer[16] = 0;
                fprintf( pAbc->Out, " %-17s %.24s%7ld", c_file.name, Buffer+4, c_file.size );
                if ( ++Counter % 2 == 0 )
                {
                    fprintf( pAbc->Out, "\n" );
                    fPrintedNewLine = 1;
                }
                else
                {
                    fprintf( pAbc->Out, " |" );
                    fPrintedNewLine = 0;
                }
            }
            while( _findnext( hFile, &c_file ) == 0 );
        }
        else
        {
            do
            {
                fprintf( pAbc->Out, " %-18s", c_file.name );
                if ( ++Counter % 4 == 0 )
                {
                    fprintf( pAbc->Out, "\n" );
                    fPrintedNewLine = 1;
                }
                else
                {
                    fprintf( pAbc->Out, " " );
                    fPrintedNewLine = 0;
                }
            }
            while( _findnext( hFile, &c_file ) == 0 );
        }
        if ( !fPrintedNewLine )
            fprintf( pAbc->Out, "\n" );
        _findclose( hFile );
    }
    return 0;

usage:
    fprintf( pAbc->Err, "usage: ls [-l] [-b]\n" );
    fprintf( pAbc->Err, "       print the file names in the current directory\n" );
    fprintf( pAbc->Err, "        -l : print in the long format [default = short]\n" );
    fprintf( pAbc->Err, "        -b : print only .mv files [default = all]\n" );
    return 1;
}


/**Function*************************************************************

  Synopsis    [Generates the script for running ABC.]

  Description []

  SideEffects []

  SeeAlso     []

***********************************************************************/
int CmdCommandScrGen( Abc_Frame_t * pAbc, int argc, char **argv )
{
    struct _finddata_t c_file;
    long   hFile;
    FILE * pFile = NULL;
    char * pFileStr = "test.s";
    char * pDirStr = NULL;
    char * pComStr = "ps";
    char * pWriteStr = NULL;
    char   Buffer[1000], Line[2000];
    int    nFileNameMax, nFileNameCur;
    int    Counter = 0;
    int    fUseCurrent;
    char   c;

    fUseCurrent = 0;
    Extra_UtilGetoptReset();
    while ( (c = Extra_UtilGetopt(argc, argv, "FDCWch") ) != EOF )
    {
        switch (c)
        {
        case 'F':
            if ( globalUtilOptind >= argc )
            {
                fprintf( pAbc->Err, "Command line switch \"-F\" should be followed by a string.\n" );
                goto usage;
            }
            pFileStr = argv[globalUtilOptind];
            globalUtilOptind++;
            break;
        case 'D':
            if ( globalUtilOptind >= argc )
            {
                fprintf( pAbc->Err, "Command line switch \"-D\" should be followed by a string.\n" );
                goto usage;
            }
            pDirStr = argv[globalUtilOptind];
            globalUtilOptind++;
            break;
        case 'C':
            if ( globalUtilOptind >= argc )
            {
                fprintf( pAbc->Err, "Command line switch \"-C\" should be followed by a string.\n" );
                goto usage;
            }
            pComStr = argv[globalUtilOptind];
            globalUtilOptind++;
            break;
        case 'W':
            if ( globalUtilOptind >= argc )
            {
                fprintf( pAbc->Err, "Command line switch \"-W\" should be followed by a string.\n" );
                goto usage;
            }
            pWriteStr = argv[globalUtilOptind];
            globalUtilOptind++;
            break;
        case 'c':
            fUseCurrent ^= 1;
            break;
        default:
            goto usage;
        }
    }

//    printf( "File = %s.\n", pFileStr );
//    printf( "Dir = %s.\n", pDirStr );
//    printf( "Com = %s.\n", pComStr );
    if ( pDirStr == NULL )
        fUseCurrent = 1;

    if ( _getcwd( Buffer, 1000 ) == NULL )
    {
        printf( "Cannot get the current directory.\n" );
        return 0;
    }
    if ( fUseCurrent )
        pFile = fopen( pFileStr, "w" );
    if ( pDirStr )
    {
        if ( _chdir(pDirStr) )
        {
            printf( "Cannot change to directory: %s\n", pDirStr );
            return 0;
        }
    }
    if ( !fUseCurrent )
        pFile = fopen( pFileStr, "w" );
    if ( pFile == NULL )
    {
        printf( "Cannot open file %s.\n", pFileStr );
        if ( pDirStr && _chdir(Buffer) )
        {
            printf( "Cannot change to the current directory: %s\n", Buffer );
            return 0;
        }
        return 0;
    }

    // find the first file in the directory
    if( (hFile = _findfirst( "*.*", &c_file )) == -1L )
    {
        if ( pDirStr )
            printf( "No files in the current directory.\n" );
        else
            printf( "No files in directory: %s\n", pDirStr );
        if ( pDirStr && _chdir(Buffer) )
        {
            printf( "Cannot change to the current directory: %s\n", Buffer );
            return 0;
        }
    }

    // get the longest file name
    {
        nFileNameMax = 0;
        do
        {
            // skip script and txt files
            nFileNameCur = strlen(c_file.name);
            if ( c_file.name[nFileNameCur-1] == '.' )
                continue;
            if ( nFileNameCur > 2 &&
                 c_file.name[nFileNameCur-1] == 's' &&
                 c_file.name[nFileNameCur-2] == '.' )
                 continue;
            if ( nFileNameCur > 4 &&
                 c_file.name[nFileNameCur-1] == 't' &&
                 c_file.name[nFileNameCur-2] == 'x' &&
                 c_file.name[nFileNameCur-3] == 't' &&
                 c_file.name[nFileNameCur-4] == '.' )
                 continue;
            if ( nFileNameMax < nFileNameCur )
                nFileNameMax = nFileNameCur;
        }
        while( _findnext( hFile, &c_file ) == 0 );
        _findclose( hFile );
    }

    // print the script file
    {
        if( (hFile = _findfirst( "*.*", &c_file )) == -1L )
        {
            if ( pDirStr )
                printf( "No files in the current directory.\n" );
            else
                printf( "No files in directory: %s\n", pDirStr );
        }
        fprintf( pFile, "# Script file produced by ABC on %s\n", Extra_TimeStamp() );
        fprintf( pFile, "# Command line was: scrgen -F %s -D %s -C \"%s\"%s%s\n",
            pFileStr, pDirStr, pComStr, pWriteStr?" -W ":"", pWriteStr?pWriteStr:"" );
        do
        {
            // skip script and txt files
            nFileNameCur = strlen(c_file.name);
            if ( c_file.name[nFileNameCur-1] == '.' )
                continue;
            if ( nFileNameCur > 2 &&
                 c_file.name[nFileNameCur-1] == 's' &&
                 c_file.name[nFileNameCur-2] == '.' )
                 continue;
            if ( nFileNameCur > 4 &&
                 c_file.name[nFileNameCur-1] == 't' &&
                 c_file.name[nFileNameCur-2] == 'x' &&
                 c_file.name[nFileNameCur-3] == 't' &&
                 c_file.name[nFileNameCur-4] == '.' )
                 continue;
            sprintf( Line, "r %s%s%-*s ; %s", pDirStr?pDirStr:"", pDirStr?"/":"", nFileNameMax, c_file.name, pComStr );
            for ( c = (int)strlen(Line)-1; c >= 0; c-- )
                if ( Line[c] == '\\' )
                    Line[c] = '/';
            fprintf( pFile, "%s", Line );
            if ( pWriteStr )
            {
                sprintf( Line, " ; w %s/%-*s", pWriteStr, nFileNameMax, c_file.name );
                for ( c = (int)strlen(Line)-1; c >= 0; c-- )
                    if ( Line[c] == '\\' )
                        Line[c] = '/';
                fprintf( pFile, "%s", Line );
            }
            fprintf( pFile, "\n", Line );
        }
        while( _findnext( hFile, &c_file ) == 0 );
        _findclose( hFile );
    }
    fclose( pFile );
    if ( pDirStr && _chdir(Buffer) )
    {
        printf( "Cannot change to the current directory: %s\n", Buffer );
        return 0;
    }

    // report
    if ( fUseCurrent )
        printf( "Script file \"%s\" was saved in the current directory.\n", pFileStr );
    else
        printf( "Script file \"%s\" was saved in directory: %s\n", pFileStr, pDirStr );
    return 0;

usage:
    fprintf( pAbc->Err, "usage: scrgen -F <str> -D <str> -C <str> -W <str> -ch\n" );
    fprintf( pAbc->Err, "\t          generates script for running ABC\n" );
    fprintf( pAbc->Err, "\t-F str  : the name of the script file [default = \"test.s\"]\n" );
    fprintf( pAbc->Err, "\t-D str  : the directory to read files from [default = current]\n" );
    fprintf( pAbc->Err, "\t-C str  : the sequence of commands to run [default = \"ps\"]\n" );
    fprintf( pAbc->Err, "\t-W str  : the directory to write the resulting files [default = no writing]\n" );
    fprintf( pAbc->Err, "\t-c      : toggle placing file in current/target dir [default = %s]\n", fUseCurrent? "current": "target" );
    fprintf( pAbc->Err, "\t-h      : print the command usage\n\n");
    fprintf( pAbc->Err, "\tExample : scrgen -F test1.s -D a/in -C \"ps; st; ps\" -W a/out\n" );
    return 1;
}
#endif


#ifdef WIN32
#define unlink _unlink
#endif

/**Function********************************************************************

  Synopsis    [Calls SIS internally.]

  Description []

  SideEffects []

  SeeAlso     []

******************************************************************************/
func CmdCommandSis(pAbc *Abc_Frame_t, argc int32, argv *string) int32 {
    var pFile *os.File;
    var pOut, pErr *os.File;
    var pNtk, pNtkNew, pNetlist *Abc_Ntk_t;
    var pNameWin string  = "sis.exe";
    var pNameUnix string  = "sis";
    var Command string // was [1000]byte in C

    var Buffer string // was [100]byte in C

    var pSisName string;
    var i int32;

    pNtk = Abc_FrameReadNtk(pAbc);
    pOut = Abc_FrameReadOut(pAbc);
    pErr = Abc_FrameReadErr(pAbc);

    if argc == 1 {
        goto usage; }
    if int32(strings.Compare(argv[1], "-h")) == 0 {
        goto usage; }
    if int32(strings.Compare(argv[1], "-?")) == 0 {
        goto usage; }

    if pNtk == nil {
        fmt.Fprintf(pErr, "Empty network.\n");
        goto usage;
    }

    if int32(strings.Compare(argv[0], "sis")) != 0 {
        fmt.Fprintf(pErr, "Wrong command: \"%s\".\n", argv[0]);
        goto usage;
    }

    // get the names from the resource file
    if Cmd_FlagReadByName(pAbc, "siswin") {
        pNameWin = Cmd_FlagReadByName(pAbc, "siswin"); }
    if Cmd_FlagReadByName(pAbc, "sisunix") {
        pNameUnix = Cmd_FlagReadByName(pAbc, "sisunix"); }

    // check if SIS is available
    if pFile = fopen( pNameWin, "r" ) {
        pSisName = pNameWin;
    } else if pFile = fopen( pNameUnix, "r" ) {
        pSisName = pNameUnix;
    } else if pFile == nil {
        fmt.Fprintf(pErr, "Cannot find \"%s\" or \"%s\" in the current directory.\n", pNameWin, pNameUnix);
        goto usage;
    }
    _ = (pFile).Close();

    if Abc_NtkIsMappedLogic(pNtk) {
        Abc_NtkMapToSop(pNtk);
        fmt.Printf("The current network is unmapped before calling SIS.\n");
    }

    // write out the current network
    if Abc_NtkIsLogic(pNtk) {
        Abc_NtkToSop(pNtk, -1, ABC_INFINITY); }
    pNetlist = Abc_NtkToNetlist(pNtk);
    if pNetlist == nil {
        fmt.Fprintf(pErr, "Cannot produce the intermediate network.\n");
        goto usage;
    }
    Io_WriteBlif( pNetlist, "_sis_in.blif", 1, 0, 0 );
    Abc_NtkDelete( pNetlist );

    // create the file for sis
    Command = fmt.Sprintf("%s -x -c ", pSisName);
    strcat ( Command, "\"" );
    strcat ( Command, "read_blif _sis_in.blif" );
    strcat ( Command, "; " );
    for  i = 1; i < argc; i++  {
        Buffer = fmt.Sprintf(" %s", argv[i]);
        strcat( Command, Buffer );
    }
    strcat( Command, "; " );
    strcat( Command, "write_blif _sis_out.blif" );
    strcat( Command, "\"" );

    // call SIS
    if Util_SignalSystem( Command ) {
        fmt.Fprintf(pErr, "The following command has returned non-zero exit status:\n");
        fprintf( pErr, "\"%s\"\n", Command );
        unlink( "_sis_in.blif" );
        goto usage;
    }

    // read in the SIS output
    if (pFile = os.Open("_sis_out.blif")) == nil {
        fmt.Fprintf(pErr, "Cannot open SIS output file \"%s\".\n", "_sis_out.blif");
        unlink( "_sis_in.blif" );
        goto usage;
    }
    _ = (pFile).Close();

    // set the new network
    pNtkNew = Io_Read( "_sis_out.blif", IO_FILE_BLIF, 1, 0 );
    // set the original spec of the new network
    if pNtk.pSpec {
        ABC_FREE( pNtkNew.pSpec );
        pNtkNew.pSpec = Extra_UtilStrsav( pNtk.pSpec );
    }
    // replace the current network
    Abc_FrameReplaceCurrentNetwork( pAbc, pNtkNew );

    // remove temporary networks
    unlink( "_sis_in.blif" );
    unlink( "_sis_out.blif" );
    return 0;

usage:
    fmt.Fprintf(pErr, "Usage: sis [-h] <com>\n");
    fmt.Fprintf(pErr, "         invokes SIS command for the current ABC network\n");
    fmt.Fprintf(pErr, "         (the executable of SIS should be in the same directory)\n");
    fmt.Fprintf(pErr, "   -h  : print the command usage\n");
    fmt.Fprintf(pErr, " <com> : a SIS command (or a semicolon-separated list of commands in quotes)\n");
    fmt.Fprintf(pErr, "         Example 1: sis eliminate 0\n");
    fmt.Fprintf(pErr, "         Example 2: sis \"ps; rd; fx; ps\"\n");
    fmt.Fprintf(pErr, "         Example 3: sis source script.rugged\n");
    return 1;                   // error exit
}


/**Function********************************************************************

  Synopsis    [Calls SIS internally.]

  Description []

  SideEffects []

  SeeAlso     []

******************************************************************************/
func CmdCommandMvsis(pAbc *Abc_Frame_t, argc int32, argv *string) int32 {
    var pFile *os.File;
    var pOut, pErr *os.File;
    var pNtk, pNtkNew, pNetlist *Abc_Ntk_t;
    var Command string // was [1000]byte in C

    var Buffer string // was [100]byte in C

    var pNameWin string  = "mvsis.exe";
    var pNameUnix string  = "mvsis";
    var pMvsisName string;
    var i int32;

    pNtk = Abc_FrameReadNtk(pAbc);
    pOut = Abc_FrameReadOut(pAbc);
    pErr = Abc_FrameReadErr(pAbc);

    if argc == 1 {
        goto usage; }
    if int32(strings.Compare(argv[1], "-h")) == 0 {
        goto usage; }
    if int32(strings.Compare(argv[1], "-?")) == 0 {
        goto usage; }

    if pNtk == nil {
        fmt.Fprintf(pErr, "Empty network.\n");
        goto usage;
    }

    if int32(strings.Compare(argv[0], "mvsis")) != 0 {
        fmt.Fprintf(pErr, "Wrong command: \"%s\".\n", argv[0]);
        goto usage;
    }

    // get the names from the resource file
    if Cmd_FlagReadByName(pAbc, "mvsiswin") {
        pNameWin = Cmd_FlagReadByName(pAbc, "mvsiswin"); }
    if Cmd_FlagReadByName(pAbc, "mvsisunix") {
        pNameUnix = Cmd_FlagReadByName(pAbc, "mvsisunix"); }

    // check if MVSIS is available
    if pFile = fopen( pNameWin, "r" ) {
        pMvsisName = pNameWin;
    } else if pFile = fopen( pNameUnix, "r" ) {
        pMvsisName = pNameUnix;
    } else if pFile == nil {
        fmt.Fprintf(pErr, "Cannot find \"%s\" or \"%s\" in the current directory.\n", pNameWin, pNameUnix);
        goto usage;
    }
    _ = (pFile).Close();

    if Abc_NtkIsMappedLogic(pNtk) {
        Abc_NtkMapToSop(pNtk);
        fmt.Printf("The current network is unmapped before calling MVSIS.\n");
    }

    // write out the current network
    if Abc_NtkIsLogic(pNtk) {
        Abc_NtkToSop(pNtk, -1, ABC_INFINITY); }
    pNetlist = Abc_NtkToNetlist(pNtk);
    if pNetlist == nil {
        fmt.Fprintf(pErr, "Cannot produce the intermediate network.\n");
        goto usage;
    }
    Io_WriteBlif( pNetlist, "_mvsis_in.blif", 1, 0, 0 );
    Abc_NtkDelete( pNetlist );

    // create the file for MVSIS
    Command = fmt.Sprintf("%s -x -c ", pMvsisName);
    strcat ( Command, "\"" );
    strcat ( Command, "read_blif _mvsis_in.blif" );
    strcat ( Command, "; " );
    for  i = 1; i < argc; i++  {
        Buffer = fmt.Sprintf(" %s", argv[i]);
        strcat( Command, Buffer );
    }
    strcat( Command, "; " );
    strcat( Command, "write_blif _mvsis_out.blif" );
    strcat( Command, "\"" );

    // call MVSIS
    if Util_SignalSystem( Command ) {
        fmt.Fprintf(pErr, "The following command has returned non-zero exit status:\n");
        fprintf( pErr, "\"%s\"\n", Command );
        unlink( "_mvsis_in.blif" );
        goto usage;
    }

    // read in the MVSIS output
    if (pFile = os.Open("_mvsis_out.blif")) == nil {
        fmt.Fprintf(pErr, "Cannot open MVSIS output file \"%s\".\n", "_mvsis_out.blif");
        unlink( "_mvsis_in.blif" );
        goto usage;
    }
    _ = (pFile).Close();

    // set the new network
    pNtkNew = Io_Read( "_mvsis_out.blif", IO_FILE_BLIF, 1, 0 );
    // set the original spec of the new network
    if pNtk.pSpec {
        ABC_FREE( pNtkNew.pSpec );
        pNtkNew.pSpec = Extra_UtilStrsav( pNtk.pSpec );
    }
    // replace the current network
    Abc_FrameReplaceCurrentNetwork( pAbc, pNtkNew );

    // remove temporary networks
    unlink( "_mvsis_in.blif" );
    unlink( "_mvsis_out.blif" );
    return 0;

usage:
    fmt.Fprintf(pErr, "Usage: mvsis [-h] <com>\n");
    fmt.Fprintf(pErr, "         invokes MVSIS command for the current ABC network\n");
    fmt.Fprintf(pErr, "         (the executable of MVSIS should be in the same directory)\n");
    fmt.Fprintf(pErr, "   -h  : print the command usage\n");
    fmt.Fprintf(pErr, " <com> : a MVSIS command (or a semicolon-separated list of commands in quotes)\n");
    fmt.Fprintf(pErr, "         Example 1: mvsis fraig_sweep\n");
    fmt.Fprintf(pErr, "         Example 2: mvsis \"ps; fxu; ps\"\n");
    fmt.Fprintf(pErr, "         Example 3: mvsis source mvsis.rugged\n");
    return 1;                   // error exit
}


/**Function*************************************************************

  Synopsis    [Computes dimentions of the graph.]

  Description []

  SideEffects []

  SeeAlso     []

***********************************************************************/
func Gia_ManGnuplotShow(pPlotFileName string) {
    var pFile *os.File;
    var pAbc any;
    var pProgNameGnuplotWin string   = "wgnuplot.exe";
    var pProgNameGnuplotUnix string  = "gnuplot";
    var pProgNameGnuplot string  = nil;

    // read in the Capo plotting output
    if (pFile = os.Open(pPlotFileName)) == nil {
        fmt.Fprintf(os.Stdout, "Cannot open the plot file \"%s\".\n\n", pPlotFileName);
        return;
    }
    _ = (pFile).Close();

    pAbc = Abc_FrameGetGlobalFrame();

    // get the names from the plotting software
    if Cmd_FlagReadByName((Abc_Frame_t *)pAbc, "gnuplotwin") {
        pProgNameGnuplotWin = Cmd_FlagReadByName((Abc_Frame_t *)pAbc, "gnuplotwin"); }
    if Cmd_FlagReadByName((Abc_Frame_t *)pAbc, "gnuplotunix") {
        pProgNameGnuplotUnix = Cmd_FlagReadByName((Abc_Frame_t *)pAbc, "gnuplotunix"); }

    // check if Gnuplot is available
    if pFile = fopen( pProgNameGnuplotWin, "r" ) {
        pProgNameGnuplot = pProgNameGnuplotWin;
    } else if pFile = fopen( pProgNameGnuplotUnix, "r" ) {
        pProgNameGnuplot = pProgNameGnuplotUnix;
    } else if pFile == nil {
        fmt.Fprintf(os.Stdout, "Cannot find \"%s\" or \"%s\" in the current directory.\n", pProgNameGnuplotWin, pProgNameGnuplotUnix);
        return;
    }
    _ = (pFile).Close();

    // spawn the viewer
#ifdef WIN32
    if ( _spawnl( _P_NOWAIT, pProgNameGnuplot, pProgNameGnuplot, pPlotFileName, NULL ) == -1 )
    {
        fprintf( stdout, "Cannot find \"%s\".\n", pProgNameGnuplot );
        return;
    }
#else
    {
        char Command[1000];
        sprintf( Command, "%s %s ", pProgNameGnuplot, pPlotFileName );
        if ( system( Command ) == -1 )
        {
            fprintf( stdout, "Cannot execute \"%s\".\n", Command );
            return;
        }
    }
#endif
}

/**Function********************************************************************

  Synopsis    [Calls Capo internally.]

  Description []

  SideEffects []

  SeeAlso     []

******************************************************************************/
func CmdCommandCapo(pAbc *Abc_Frame_t, argc int32, argv *string) int32 {
    var pFile *os.File;
    var pOut, pErr *os.File;
    var pNtk, pNetlist *Abc_Ntk_t;
    var Command string // was [1000]byte in C

    var Buffer string // was [100]byte in C

    var pProgNameCapoWin string      = "capo.exe";
    var pProgNameCapoUnix string     = "capo";
    var pProgNameGnuplotWin string   = "wgnuplot.exe";
    var pProgNameGnuplotUnix string  = "gnuplot";
    var pProgNameCapo string;
    var pProgNameGnuplot string;
    var pPlotFileName string;
    var i int32;

    pNtk = Abc_FrameReadNtk(pAbc);
    pOut = Abc_FrameReadOut(pAbc);
    pErr = Abc_FrameReadErr(pAbc);

    if argc > 1 {
        if int32(strings.Compare(argv[1], "-h")) == 0 {
            goto usage; }
        if int32(strings.Compare(argv[1], "-?")) == 0 {
            goto usage; }
    }

    if pNtk == nil {
        fmt.Fprintf(pErr, "Empty network.\n");
        goto usage;
    }

    if int32(strings.Compare(argv[0], "capo")) != 0 {
        fmt.Fprintf(pErr, "Wrong command: \"%s\".\n", argv[0]);
        goto usage;
    }

    // get the names from the resource file
    if Cmd_FlagReadByName(pAbc, "capowin") {
        pProgNameCapoWin = Cmd_FlagReadByName(pAbc, "capowin"); }
    if Cmd_FlagReadByName(pAbc, "capounix") {
        pProgNameCapoUnix = Cmd_FlagReadByName(pAbc, "capounix"); }

    // check if capo is available
    if pFile = fopen( pProgNameCapoWin, "r" ) {
        pProgNameCapo = pProgNameCapoWin;
    } else if pFile = fopen( pProgNameCapoUnix, "r" ) {
        pProgNameCapo = pProgNameCapoUnix;
    } else if pFile == nil {
        fmt.Fprintf(pErr, "Cannot find \"%s\" or \"%s\" in the current directory.\n", pProgNameCapoWin, pProgNameCapoUnix);
        goto usage;
    }
    _ = (pFile).Close();

    if Abc_NtkIsMappedLogic(pNtk) {
        Abc_NtkMapToSop(pNtk);
        fmt.Printf("The current network is unmapped before calling Capo.\n");
    }

    // write out the current network
    if Abc_NtkIsLogic(pNtk) {
        Abc_NtkToSop(pNtk, -1, ABC_INFINITY); }
    pNetlist = Abc_NtkToNetlist(pNtk);
    if pNetlist == nil {
        fmt.Fprintf(pErr, "Cannot produce the intermediate network.\n");
        goto usage;
    }
    Io_WriteBlif( pNetlist, "_capo_in.blif", 1, 0, 0 );
    Abc_NtkDelete( pNetlist );

    // create the file for Capo
    Command = fmt.Sprintf("%s -f _capo_in.blif -log out.txt ", pProgNameCapo);
    pPlotFileName = nil;
    for  i = 1; i < argc; i++  {
        Buffer = fmt.Sprintf(" %s", argv[i]);
        strcat( Command, Buffer );
        if !(int32(strings.Compare(argv[i], "-plot")) != 0) {
            pPlotFileName = argv[i+1]; }
    }

    // call Capo
    if Util_SignalSystem( Command ) {
        fmt.Fprintf(pErr, "The following command has returned non-zero exit status:\n");
        fprintf( pErr, "\"%s\"\n", Command );
        unlink( "_capo_in.blif" );
        goto usage;
    }
    // remove temporary networks
    unlink( "_capo_in.blif" );
    if pPlotFileName == nil {
        return 0; }

    // get the file name
    Buffer = fmt.Sprintf("%s.plt", pPlotFileName);
    pPlotFileName = Buffer;

    // read in the Capo plotting output
    if (pFile = os.Open(pPlotFileName)) == nil {
        fmt.Fprintf(pErr, "Cannot open the plot file \"%s\".\n\n", pPlotFileName);
        goto usage;
    }
    _ = (pFile).Close();

    // get the names from the plotting software
    if Cmd_FlagReadByName(pAbc, "gnuplotwin") {
        pProgNameGnuplotWin = Cmd_FlagReadByName(pAbc, "gnuplotwin"); }
    if Cmd_FlagReadByName(pAbc, "gnuplotunix") {
        pProgNameGnuplotUnix = Cmd_FlagReadByName(pAbc, "gnuplotunix"); }

    // check if Gnuplot is available
    pProgNameGnuplot = nil;
    if pFile = fopen( pProgNameGnuplotWin, "r" ) {
        pProgNameGnuplot = pProgNameGnuplotWin;
    } else if pFile = fopen( pProgNameGnuplotUnix, "r" ) {
        pProgNameGnuplot = pProgNameGnuplotUnix;
    } else if pFile == nil {
        fmt.Fprintf(pErr, "Cannot find \"%s\" or \"%s\" in the current directory.\n", pProgNameGnuplotWin, pProgNameGnuplotUnix);
        goto usage;
    }
    _ = (pFile).Close();

    // spawn the viewer
#ifdef WIN32
    if ( _spawnl( _P_NOWAIT, pProgNameGnuplot, pProgNameGnuplot, pPlotFileName, NULL ) == -1 )
    {
        fprintf( stdout, "Cannot find \"%s\".\n", pProgNameGnuplot );
        goto usage;
    }
#else
    {
        sprintf( Command, "%s %s ", pProgNameGnuplot, pPlotFileName );
        if ( Util_SignalSystem( Command ) == -1 )
        {
            fprintf( stdout, "Cannot execute \"%s\".\n", Command );
            goto usage;
        }
    }
#endif

    // remove temporary networks
//    unlink( pPlotFileName );
    return 0;

usage:
    fmt.Fprintf(pErr, "Usage: capo [-h] <com>\n");
    fmt.Fprintf(pErr, "         peforms placement of the current network using Capo\n");
    fmt.Fprintf(pErr, "         a Capo binary should be present in the same directory\n");
    fmt.Fprintf(pErr, "         (if plotting, the Gnuplot binary should also be present)\n");
    fmt.Fprintf(pErr, "   -h  : print the command usage\n");
    fmt.Fprintf(pErr, " <com> : a Capo command\n");
    fmt.Fprintf(pErr, "         Example 1: capo\n");
    fmt.Fprintf(pErr, "                    (performs placement with default options)\n");
    fmt.Fprintf(pErr, "         Example 2: capo -AR <aspec_ratio> -WS <whitespace_percentage> -save\n");
    fmt.Fprintf(pErr, "                    (specifies the aspect ratio [default = 1.0] and\n");
    fmt.Fprintf(pErr, "                    the whitespace percentage [0%%; 100%%) [default = 15%%])\n");
    fmt.Fprintf(pErr, "         Example 3: capo -plot <base_fileName>\n");
    fmt.Fprintf(pErr, "                    (produces <base_fileName.plt> and visualize it using Gnuplot)\n");
    fmt.Fprintf(pErr, "         Example 4: capo -help\n");
    fmt.Fprintf(pErr, "                    (prints the default usage message of the Capo binary)\n");
    fmt.Fprintf(pErr, "         Please refer to the Capo webpage for additional information:\n");
    fmt.Fprintf(pErr, "         http://vlsicad.eecs.umich.edu/BK/PDtools/\n");
    return 1;                   // error exit
}

/**Function*************************************************************

  Synopsis    []

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func CmdCommandStarter(pAbc *Abc_Frame_t, argc int32, argv *string) int32 {
    extern void Cmd_RunStarter( char * pFileName, char * pBinary, char * pCommand, int nCores );
    var pFile *os.File;
    var pFileName string;
    var pCommand string  = nil;
    var c int32;
    var nCores int32 = 3;
    var fVerbose int32;
    Extra_UtilGetoptReset();
    for  ( c = Extra_UtilGetopt( argc, argv, "NCvh" ) ) != -1  {
        switch  c  {
        case 'N':
            if globalUtilOptind >= argc {
                Abc_Print( -1, "Command line switch \"-N\" should be followed by an integer.\n" );
                goto usage;
            }
            nCores = strconv.Atoi(argv[globalUtilOptind]);
            globalUtilOptind++;
            if nCores < 0 { 
                goto usage; }
            break;
        case 'C':
            if globalUtilOptind >= argc {
                Abc_Print( -1, "Command line switch \"-C\" should be followed by a string (possibly in quotes).\n" );
                goto usage;
            }
            pCommand = argv[globalUtilOptind];
            globalUtilOptind++;
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
        Abc_Print( -2, "The file name should be given on the command line.\n" );
        return 1;
    }
    // get the input file name
    pFileName = argv[globalUtilOptind];
    if (pFile = Io_FileOpen( pFileName, "open_path", "rb", 0 )) == nil {
//    if ( (pFile = fopen( pFileName, "rb" )) == NULL )
    
        Abc_Print( -2, "Cannot open input file \"%s\". ", pFileName );
        if pFileName = Extra_FileGetSimilarName( pFileName, ".c", ".s", ".scr", ".script", nil ) {
            Abc_Print( -2, "Did you mean \"%s\"?", pFileName ); }
        Abc_Print( -2, "\n" );
        return 1;
    }
    _ = (pFile).Close();
    // run commands
    Cmd_RunStarter( pFileName, pAbc.sBinary, pCommand, nCores );
    return 0;

usage:
    Abc_Print( -2, "usage: starter [-N num] [-C cmd] [-vh] <file>\n" );
    Abc_Print( -2, "\t         runs command lines listed in <file> concurrently on <num> CPUs\n" );
    Abc_Print( -2, "\t-N num : the number of concurrent jobs including the controler [default = %d]\n", nCores );
    Abc_Print( -2, "\t-C cmd : (optional) ABC command line to execute on benchmarks in <file>\n" );
    Abc_Print( -2, "\t-v     : toggle printing verbose information [default = %s]\n", fVerbose? "yes": "no" );
    Abc_Print( -2, "\t-h     : print the command usage\n");
    Abc_Print( -2, "\t<file> : file name with ABC command lines (or benchmark names, if <cmd> is given)\n");
    return 1;
}

/**Function*************************************************************

  Synopsis    []

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func CmdCommandAutoTuner(pAbc *Abc_Frame_t, argc int32, argv *string) int32 {
    extern void Cmd_RunAutoTuner( char * pConfig, char * pFileList, int nCores );
    var pFile *os.File;
    var pFileConf string  = nil;
    var pFileList string  = nil;
    var pFileName string;
    var c int32;
    var nCores int32 = 3;
    var fVerbose int32;
    Extra_UtilGetoptReset();
    for  ( c = Extra_UtilGetopt( argc, argv, "NCFvh" ) ) != -1  {
        switch  c  {
        case 'N':
            if globalUtilOptind >= argc {
                Abc_Print( -1, "Command line switch \"-N\" should be followed by an integer.\n" );
                goto usage;
            }
            nCores = strconv.Atoi(argv[globalUtilOptind]);
            globalUtilOptind++;
            if nCores < 0 { 
                goto usage; }
            break;
        case 'C':
            if globalUtilOptind >= argc {
                Abc_Print( -1, "Command line switch \"-C\" should be followed by a string (possibly in quotes).\n" );
                goto usage;
            }
            pFileConf = argv[globalUtilOptind];
            globalUtilOptind++;
            break;
        case 'F':
            if globalUtilOptind >= argc {
                Abc_Print( -1, "Command line switch \"-F\" should be followed by a string (possibly in quotes).\n" );
                goto usage;
            }
            pFileList = argv[globalUtilOptind];
            globalUtilOptind++;
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
    if pFileConf == nil {
        Abc_Print( -2, "File containing configuration for autotuning is not given.\n" );
        return 1;
    }
    if pFileList == nil {
        Abc_Print( -2, "File contining list of files for autotuning is not given.\n" );
        return 1;
    }
    // get the input file name
    pFileName = pFileConf;
    if (pFile = Io_FileOpen( pFileName, "open_path", "rb", 0 )) == nil {
//    if ( (pFile = fopen( pFileName, "rb" )) == NULL )
    
        Abc_Print( -2, "Cannot open configuration file \"%s\". ", pFileName );
        if pFileName = Extra_FileGetSimilarName( pFileName, ".c", ".s", ".scr", ".script", nil ) {
            Abc_Print( -2, "Did you mean \"%s\"?", pFileName ); }
        Abc_Print( -2, "\n" );
        return 1;
    }
    _ = (pFile).Close();
    // get the input file name
    pFileName = pFileList;
    if (pFile = Io_FileOpen( pFileName, "open_path", "rb", 0 )) == nil {
//    if ( (pFile = fopen( pFileName, "rb" )) == NULL )
    
        Abc_Print( -2, "Cannot open the file list \"%s\". ", pFileName );
        if pFileName = Extra_FileGetSimilarName( pFileName, ".c", ".s", ".scr", ".script", nil ) {
            Abc_Print( -2, "Did you mean \"%s\"?", pFileName ); }
        Abc_Print( -2, "\n" );
        return 1;
    }
    _ = (pFile).Close();
    // run commands
    Cmd_RunAutoTuner( pFileConf, pFileList, nCores );
    return 0;

usage:
    Abc_Print( -2, "usage: autotuner [-N num] [-C file] [-F file] [-vh]\n" );
    Abc_Print( -2, "\t         performs autotuning\n" );
    Abc_Print( -2, "\t-N num : the number of concurrent jobs including the controler [default = %d]\n", nCores );
    Abc_Print( -2, "\t-C cmd : configuration file with settings for autotuning\n" );
    Abc_Print( -2, "\t-F cmd : list of AIGER files to be used for autotuning\n" );
    Abc_Print( -2, "\t-v     : toggle printing verbose information [default = %s]\n", fVerbose? "yes": "no" );
    Abc_Print( -2, "\t-h     : print the command usage\n");
    return 1;
}

/**Function********************************************************************

  Synopsis    [Print the version string.]

  Description []

  SideEffects []

  SeeAlso     []

******************************************************************************/
func CmdCommandVersion(pAbc *Abc_Frame_t, argc int32, argv *string) int32 {
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

    fmt.Printf("%s\n", Abc_UtilsGetVersion(pAbc));
    return 0;

usage:
    fmt.Fprintf(pAbc.Err, "usage: version [-h]\n");
    fmt.Fprintf(pAbc.Err, "         print the version string\n");
    fmt.Fprintf(pAbc.Err, "   -h :  print the command usage\n");
    return 1;
}


////////////////////////////////////////////////////////////////////////
///                       END OF FILE                                ///
////////////////////////////////////////////////////////////////////////


//ABC_NAMESPACE_IMPL_END
