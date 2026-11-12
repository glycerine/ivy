package cmd

import (
	"fmt"
	"strings"
)

import "os"

/**CFile****************************************************************

  FileName    [cmdUtils.c]

  SystemName  [ABC: Logic synthesis and verification system.]

  PackageName [Command processing package.]

  Synopsis    [Various utilities of the command package.]

  Author      [Alan Mishchenko]
  
  Affiliation [UC Berkeley]

  Date        [Ver. 1.0. Started - June 20, 2005.]

  Revision    [$Id: cmdUtils.c,v 1.00 2005/06/20 00:00:00 alanmi Exp $]

***********************************************************************/

//#include "base/abc/abc.h"
//#include "base/main/mainInt.h"
//#include "cmdInt.h"
//#include <ctype.h>

//ABC_NAMESPACE_IMPL_START
    // proper declaration of isspace

////////////////////////////////////////////////////////////////////////
///                        DECLARATIONS                              ///
////////////////////////////////////////////////////////////////////////

// static int CmdCommandPrintCompare( Abc_Command ** ppC1, Abc_Command ** ppC2 );


////////////////////////////////////////////////////////////////////////
///                     FUNCTION DEFINITIONS                         ///
////////////////////////////////////////////////////////////////////////

/**Function*************************************************************

  Synopsis    []

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func cmdCheckShellEscape(pAbc *Abc_Frame_t, argc int32, argv *string) int32 {
    var RetValue int32;
    if argv[0][0] == '!' {
        int size = 4096;
        var i int32;
        var buffer string  = ABC_ALLOC(char, 10000);
        strncpy (buffer, &argv[0][1], size);
        for i = 1; i < argc; ++i {
                strncat (buffer, " ", size);
                strncat (buffer, argv[i], size);
        }
        if buffer[0] == 0 { 
            strncpy (buffer, "/bin/sh", size); }
        RetValue = system (buffer);
        ABC_FREE( buffer );

        // NOTE: Since we reconstruct the cmdline by concatenating
        // the parts, we lose information. So a command like
        // `!ls "file name"` will be sent to the system as
        // `ls file name` which is a BUG

        return 1;
    
    } else {
        return 0;
    }
}

/**Function*************************************************************

  Synopsis    [Executes one command.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func CmdCommandDispatch(pAbc *Abc_Frame_t, pargc *int32, pargv **string) int32 {
    var argc int32  = *pargc;
    var argv *string  = *pargv;
    var argv2 *string;

    var pNetCopy *Abc_Ntk_t;
    var pFunc Cmd_CommandFuncType;
    var pCommand *Abc_Command;
    var value string;
    var fError int32;
    var clk float64;

    if argc == 0 {
        return 0; }

    if cmdCheckShellEscape( pAbc, argc, argv ) == 1 {
        return 0; }

    // get the command
    if ! st__lookup( pAbc.tCommands, argv[0], (char **)&pCommand ) {   // the command is not in the table
        // if there is only one word with an extension, assume this is file to be read
        if argc == 1 && strstr( argv[0], "." ) {
            // add command 'read' assuming that this is the file name
            argv2 = CmdAddToArgv( argc, argv );
            CmdFreeArgv( argc, argv );
            argc = argc+1;
            argv = argv2;
            *pargc = argc;
            *pargv = argv;
            if ! st__lookup( pAbc.tCommands, argv[0], (char **)&pCommand ) {
                assert(false, "0"); }
        
        } else {
            fmt.Fprintf(pAbc.Err, "** cmd error: unknown command '%s'\n", argv[0]);
            fmt.Fprintf(pAbc.Err, "(this is likely caused by using an alias defined in \"abc.rc\"\n");
            fmt.Fprintf(pAbc.Err, "without having this file in the current or parent directory)\n");
            return 1;
        }
    }

    // get the backup network if the command is going to change the network
    if pCommand.fChange {
        if pAbc.pNtkCur && Abc_FrameIsFlagEnabled( "backup" ) {
            pNetCopy = Abc_NtkDup( pAbc.pNtkCur );
            Abc_FrameSetCurrentNetwork( pAbc, pNetCopy );
            // swap the current network and the backup network 
            // to prevent the effect of resetting the short names
            Abc_FrameSwapCurrentAndBackup( pAbc );
        }
    }

    // execute the command
    clk = Extra_CpuTimeDouble();
#ifdef __EMSCRIPTEN__
    Cmd_CommandId_t CommandId;
    CommandId = Cmd_CommandNameToId( argv[0] );
    if ( CmdCommandDispatchBasicById( CommandId, pAbc, argc, argv, &fError ) ||
         IoCommandDispatchById( CommandId, pAbc, argc, argv, &fError ) ||
         AbcCommandDispatchById( CommandId, pAbc, argc, argv, &fError ) )
    {
        pAbc->TimeCommand += Extra_CpuTimeDouble() - clk;
    }
    else
#endif
    {
    pFunc = pCommand.pFunc;
    fError = (*pFunc)( pAbc, argc, argv );
    pAbc.TimeCommand += Extra_CpuTimeDouble() - clk;
    }

    // automatic execution of arbitrary command after each command 
    // usually this is a passive command ... 
    if fError == 0 && !pAbc.fAutoexac {
        if st__lookup( pAbc.tFlags, "autoexec", &value ) {
            pAbc.fAutoexac = 1;
            fError = Cmd_CommandExecute( pAbc, value );
            pAbc.fAutoexac = 0;
        }
    }
    return fError;
}

/**Function*************************************************************

  Synopsis    [Splits the command line string into individual commands.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func CmdSplitLine(pAbc *Abc_Frame_t, sCommand string, argc *int32, argv **string) string {
    char *p, *start;
    var c byte;
    var i, j int32;
    var new_arg string;
    var vArgs *Vec_Ptr_t;
    var single_quote, double_quote int32;

    vArgs = Vec_PtrAlloc( 10 );

    p = sCommand;
    for {
        // skip leading white space 
        for  isspace( int32(   *p) )  {
            p++;
        }

        // skip until end of this token 
        single_quote = double_quote = 0;
        for  start = p; ( c = *p ) != '\0'; p++  {
            if c == ';' || c == '#' || isspace( int32(   c) ) {
                if !single_quote && !(double_quote != 0) {
                    break;
                }
            }
            if c == '\'' {
                single_quote = !single_quote;
            }
            if c == '"' {
                double_quote = !double_quote;
            }
        }
        if single_quote || (double_quote != 0) {
            ( void ) fmt.Fprintf(pAbc.Err, "** cmd warning: ignoring unbalanced quote ...\n");
        }
        if start == p {
            break; }

        new_arg = ABC_ALLOC( char, p - start + 1 );
        j = 0;
        for  i = 0; i < p - start; i++  {
            c = start[i];
            if ( c != '\'' ) && ( c != '\"' ) {
                new_arg[j++] = isspace( int32(   c) ) ? ' ' : start[i];
            }
        }
        new_arg[j] = '\0';
        Vec_PtrPush( vArgs, new_arg );
    }

    *argc = vArgs.nSize;
    *argv = (char **)Vec_PtrReleaseArray( vArgs );
    Vec_PtrFree( vArgs );
    if *p == ';' {
        p++;
    
    } else if *p == '#' {
        for  ; *p != 0; p++  {} // skip to end of line 
    }
    return p;
}

/**Function*************************************************************

  Synopsis    [Replaces parts of the command line string by aliases if given.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func CmdApplyAlias(pAbc *Abc_Frame_t, argcp *int32, argvp **string, loop *int32) int32 {
    var i, argc, stopit, added, offset, did_subst, subst, fError, newc, j int32;
    char *arg;
    var argv, newv *string;
    var alias *Abc_Alias;

    argc = *argcp;
    argv = *argvp;
    stopit = 0;
    for  ; *loop < 200; ( *loop )++  {
        if argc == 0 {
            return 0; }
        if stopit != 0 || st__lookup( pAbc.tAliases, argv[0],  (char **) &alias ) == 0 {
            return 0;
        }
        if int32(strings.Compare(argv[0], alias.argv[0])) == 0 {
            stopit = 1;
        }
        ABC_FREE( argv[0] );
        added = alias.argc - 1;

        /* shift all the arguments to the right */
        if added != 0 {
            argv = ABC_REALLOC( char *, argv, argc + added );
            for  i = argc - 1; i >= 1; i--  {
                argv[i + added] = argv[i];
            }
            for  i = 1; i <= added; i++  {
                argv[i] = nil;
            }
            argc += added;
        }
        subst = 0;
        for  i, offset = 0, 0; i < alias.argc; i, offset = i+1, offset+1  {
            arg = CmdHistorySubstitution( pAbc, alias.argv[i], &did_subst );
            if arg == nil {
                *argcp = argc;
                *argvp = argv;
                return ( 1 );
            }
            if did_subst != 0 {
                subst = 1;
            }
            fError = 0;
            for
            {
                arg = CmdSplitLine( pAbc, arg, &newc, &newv );
                /*
                 * If there's a complete `;' terminated command in `arg',
                 * when split_line() returns arg[0] != '\0'.
                 */
                if arg[0] == '\0' { /* just a bunch of words */
                    break;
                }
                fError = CmdApplyAlias( pAbc, &newc, &newv, loop );
                if fError == 0 {
                       fError = CmdCommandDispatch( pAbc, &newc, &newv );
                }
                CmdFreeArgv( newc, newv );
            
if !( fError == 0 ) { break }
}
            
            if fError != 0 {
                *argcp = argc;
                *argvp = argv;
                return ( 1 );
            }
            added = newc - 1;
            if added != 0 {
                argv = ABC_REALLOC( char *, argv, argc + added );
                for  j = argc - 1; j > offset; j--  {
                    argv[j + added] = argv[j];
                }
                argc += added;
            }
            for  j = 0; j <= added; j++  {
                argv[j + offset] = newv[j];
            }
            ABC_FREE( newv );
            offset += added;
        }
        if subst == 1 {
            for  i = offset; i < argc; i++  {
                ABC_FREE( argv[i] );
            }
            argc = offset;
        }
        *argcp = argc;
        *argvp = argv;
    }

    fmt.Fprintf(pAbc.Err, "** cmd warning: alias loop\n");
    return 1;
}

/**Function*************************************************************

  Synopsis    [Performs history substitution (now, disabled).]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func CmdHistorySubstitution(pAbc *Abc_Frame_t, line string, changed *int32) string {
    // as of today, no history substitution 
    *changed = 0;
    return line;
}

/**Function*************************************************************

  Synopsis    [Opens the file with path (now, disabled).]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func CmdFileOpen(pAbc *Abc_Frame_t, sFileName string, sMode string, pFileNameReal *string, silent int32) *os.File {
    var sRealName, sPathUsr, sPathLib, sPathAll string;
    var pFile *os.File;
    
    if int32(strings.Compare(sFileName, "-")) == 0 {
        if int32(strings.Compare(sMode, "w")) == 0 {
            sRealName = Extra_UtilStrsav( "stdout" );
            pFile = os.Stdout;
        
        } else {
            sRealName = Extra_UtilStrsav( "stdin" );
            pFile = os.Stdin;
        }
    
    } else {
        sRealName = nil;
        if int32(strings.Compare(sMode, "r")) == 0 {
            
            /* combine both pathes if exist */
            sPathUsr = Cmd_FlagReadByName(pAbc,"open_path");
            sPathLib = Cmd_FlagReadByName(pAbc,"lib_path");
            
            if sPathUsr == nil && sPathLib == nil {
                sPathAll = nil;
            
            } else if sPathUsr == nil {
                sPathAll = Extra_UtilStrsav( sPathLib );
            
            } else if sPathLib == nil {
                sPathAll = Extra_UtilStrsav( sPathUsr );
            
            } else {
                sPathAll = ABC_ALLOC( char, len(sPathLib)+len(sPathUsr)+5 );
                sPathAll = fmt.Sprintf("%s:%s", sPathUsr, sPathLib);
            }
            if sPathAll != nil {
                sRealName = Extra_UtilFileSearch(sFileName, sPathAll, "r");
                ABC_FREE( sPathAll );
            }
        }
        if sRealName == nil {
            sRealName = Extra_UtilTildeExpand(sFileName);
        }

        if (pFile = fopen(sRealName, sMode)) == nil {
            if ! (silent != 0) {
//                perror(sRealName);
                Abc_Print( 1, "Cannot open file \"%s\".\n", sRealName );
            }
        
        } else {
            // print the path/name of the resource file 'abc.rc' that is being loaded
            if !(silent != 0) && len(sRealName) >= 6 && int32(strings.Compare(sRealName + len(sRealName) - 6, "abc.rc")) == 0 {            
                Abc_Print( 1, "Loading resource file \"%s\".\n", sRealName ); }
        }
    }
    if pFileNameReal != nil {
        *pFileNameReal = sRealName;
    } else {
        ABC_FREE(sRealName); }
    
    return pFile;
}

/**Function*************************************************************

  Synopsis    [Frees the previously allocated argv array.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func CmdFreeArgv(argc int32, argv *string) {
    var i int32;
    for  i = 0; i < argc; i++  {
        ABC_FREE( argv[i] );
    }
    ABC_FREE( argv );
}
func CmdDupArgv(argc int32, argv *string) *string {
    var argvNew *string  = ABC_ALLOC( char *, argc );
    var i int32;
    for  i = 0; i < argc; i++  {
        argvNew[i] = Abc_UtilStrsav( argv[i] );
    }
    return argvNew;
}

/**Function*************************************************************

  Synopsis    [Frees the previously allocated argv array.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func CmdAddToArgv(argc int32, argv *string) *string {
    var argv2 *string;
    var i int32;
    argv2 = ABC_ALLOC( char *, argc + 1 ); 
    argv2[0] = Extra_UtilStrsav( "read" ); 
//    argv2[0] = Extra_UtilStrsav( "&r" ); 
    for  i = 0; i < argc; i++  {
        argv2[i+1] = Extra_UtilStrsav( argv[i] );
    } 
    return argv2;
}

/**Function*************************************************************

  Synopsis    [Frees the previously allocated command.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func CmdCommandFree(pCommand *Abc_Command) {
    ABC_FREE( pCommand.sGroup );
    ABC_FREE( pCommand.sName );
    ABC_FREE( pCommand );
}


/**Function*************************************************************

  Synopsis    [Prints commands alphabetically by group.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func CmdCommandPrint(pAbc *Abc_Frame_t, fPrintAll int32, fDetails int32) {
    char *key;
    var value string;
    var gen *st__generator;
    var ppCommands **Abc_Command;
    var pCommands *Abc_Command;
    var nCommands, iGroupStart, i, j int32;
    var sGroupCur string;
    var LenghtMax, nColumns, iCom int32;
    var backupErr *os.File  = pAbc.Err;

    // put all commands into one array
    nCommands = st__count( pAbc.tCommands );
    ppCommands = ABC_ALLOC( Abc_Command *, nCommands );
    i = 0;
    st__foreach_item( pAbc.tCommands, gen, &key, &value )
    {
        pCommands = (Abc_Command *)value;
        if (fPrintAll != 0) || pCommands.sName[0] != '_' {
            ppCommands[i++] = pCommands; }
    }
    nCommands = i;

    // sort command by group and then by name, alphabetically
    qsort( (void *)ppCommands, (size_t)nCommands, sizeof(Abc_Command *), 
                (int (*)(void *, void *)) CmdCommandPrintCompare );
    assert(CmdCommandPrintCompare( ppCommands, ppCommands + nCommands - 1 ) <= 0, "CmdCommandPrintCompare( ppCommands, ppCommands + nCommands - 1 ) <= 0");

    // get the longest command name
    LenghtMax = 0;
    for  i = 0; i < nCommands; i++  {
        if LenghtMax < int32(len(ppCommands[i].sName)) {
             LenghtMax = int32(len(ppCommands[i].sName)); }
    } 
    // get the number of columns
    nColumns = 79 / (LenghtMax + 2);

    // print the starting message 
    fmt.Fprintf(pAbc.Out, "      Welcome to ABC compiled on %s %s!", __DATE__, __TIME__);

    // print the command by group
    sGroupCur = nil;
    iGroupStart = 0;
    pAbc.Err = pAbc.Out;
    for  i = 0; i < nCommands; i++  {
        if sGroupCur && int32(strings.Compare(sGroupCur, ppCommands[i].sGroup)) == 0 { // this command belongs to the same group as the previous one
            if iCom++ % nColumns == 0 {
                fmt.Fprintf(pAbc.Out, "\n"); } 
            // print this command
            fmt.Fprintf(pAbc.Out, " %-*s", LenghtMax, ppCommands[i].sName);
        
        } else { // this command starts the new group of commands
            // start the new group
            if (fDetails != 0) && i != iGroupStart { // print help messages for all commands in the previous groups
                fmt.Fprintf(pAbc.Out, "\n");
                for  j = iGroupStart; j < i; j++  {
                    var tmp_cmd string;
                    fmt.Fprintf(pAbc.Out, "\n");
                    // fprintf( pAbc->Out, "--- %s ---\n", ppCommands[j]->sName );
                    tmp_cmd = ABC_ALLOC(char, len(ppCommands[j].sName)+4);
                    (void) sprintf(tmp_cmd, "%s -h", ppCommands[j].sName);
                    (void) Cmd_CommandExecute( pAbc, tmp_cmd );
                    ABC_FREE(tmp_cmd);
                }
                fmt.Fprintf(pAbc.Out, "\n");
                fmt.Fprintf(pAbc.Out, "   ----------------------------------------------------------------------");
                iGroupStart = i;
            }
            fmt.Fprintf(pAbc.Out, "\n");
            fmt.Fprintf(pAbc.Out, "\n");
            fmt.Fprintf(pAbc.Out, "%s commands:\n", ppCommands[i].sGroup);
            // print this command
            fmt.Fprintf(pAbc.Out, " %-*s", LenghtMax, ppCommands[i].sName);
            // remember current command group
            sGroupCur = ppCommands[i].sGroup;
            // reset the command counter
            iCom = 1;
        }
    }
    if (fDetails != 0) && i != iGroupStart { // print help messages for all commands in the previous groups
        fmt.Fprintf(pAbc.Out, "\n");
        for  j = iGroupStart; j < i; j++  {
            var tmp_cmd string;
            fmt.Fprintf(pAbc.Out, "\n");
            // fprintf( pAbc->Out, "--- %s ---\n", ppCommands[j]->sName );
            tmp_cmd = ABC_ALLOC(char, len(ppCommands[j].sName)+4);
            (void) sprintf(tmp_cmd, "%s -h", ppCommands[j].sName);
            (void) Cmd_CommandExecute( pAbc, tmp_cmd );
            ABC_FREE(tmp_cmd);
        }
    }
    pAbc.Err = backupErr;
    fmt.Fprintf(pAbc.Out, "\n");
    ABC_FREE( ppCommands );
}
 
/**Function*************************************************************

  Synopsis    [Comparision function used for sorting commands.]

  Description []
                
  SideEffects []

  SeeAlso     []

***********************************************************************/
func CmdCommandPrintCompare(ppC1 **Abc_Command, ppC2 **Abc_Command) int32 {
    var pC1 *Abc_Command  = *ppC1;
    var pC2 *Abc_Command  = *ppC2;
    var RetValue int32;

    RetValue = int32(strings.Compare(pC1.sGroup, pC2.sGroup));
    if RetValue < 0 {
        return -1; }
    if RetValue > 0 {
        return 1; }
    // the command belong to the same group

    // put commands with "_" at the end of the list
    if pC1.sName[0] != '_' && pC2.sName[0] == '_' {
        return -1; }
    if pC1.sName[0] == '_' && pC2.sName[0] != '_' {
        return 1; }

    RetValue = int32(strings.Compare(pC1.sName, pC2.sName));
    if RetValue < 0 {
        return -1; }
    if RetValue > 0 {
        return 1; }
     // should not be two indentical commands
    assert(false, "0");
    return 0;
}
 
/**Function*************************************************************

  Synopsis    [Comparision function used for sorting commands.]

  Description []
                
  SideEffects []

  SeeAlso     []

***********************************************************************/
func CmdNamePrintCompare(ppC1 *string, ppC2 *string) int32 {
    return int32(strings.Compare(*ppC1, *ppC2));
}

/**Function*************************************************************

  Synopsis    [Comparision function used for sorting commands.]

  Description []
                
  SideEffects []

  SeeAlso     []

***********************************************************************/
func CmdPrintTable(tTable *st__table, fAliases int32) {
    var gen *st__generator;
    char ** ppNames;
    char * key;
    var value string;
    var nNames, i int32;

    // collect keys in the array
    ppNames = ABC_ALLOC( char *, st__count(tTable) );
    nNames = 0;
    st__foreach_item( tTable, gen, &key, &value )
        ppNames[nNames++] = key;

    // sort array by name
    qsort( (void *)ppNames, (size_t)nNames, sizeof(char *), 
            (int (*)(void *, void *))CmdNamePrintCompare );

    // print in this order
    for  i = 0; i < nNames; i++  {
        st__lookup( tTable, ppNames[i], &value );
        if fAliases != 0 {
            CmdCommandAliasPrint( Abc_FrameGetGlobalFrame(), (Abc_Alias *)value );
        } else {
            fmt.Fprintf(os.Stdout, "%-15s %-15s\n", ppNames[i], value); }
    }
    ABC_FREE( ppNames );
}

////////////////////////////////////////////////////////////////////////
///                       END OF FILE                                ///
////////////////////////////////////////////////////////////////////////
//ABC_NAMESPACE_IMPL_END
