package cmd

import (
	"fmt"
	"strconv"
	"strings"
)

import "os"

/**CFile****************************************************************

  FileName    [cmdPlugin.c]

  SystemName  [ABC: Logic synthesis and verification system.]

  PackageName [Command processing package.]

  Synopsis    [Integrating external binary.]

  Author      [Alan Mishchenko, Niklas Een]
  
  Affiliation [UC Berkeley]

  Date        [Ver. 1.0. Started - September 29, 2010.]

  Revision    [$Id: cmdPlugin.c,v 1.00 2010/09/29 00:00:00 alanmi Exp $]

***********************************************************************/
 
#ifdef WIN32
#include <io.h> 
#include <process.h> 
#else
#include <unistd.h>
#endif

//#include "base/abc/abc.h"
//#include "base/main/mainInt.h"
//#include "cmd.h"
//#include "cmdInt.h"
//#include "misc/util/utilSignal.h"

//ABC_NAMESPACE_IMPL_START

////////////////////////////////////////////////////////////////////////
///                        DECLARATIONS                              ///
////////////////////////////////////////////////////////////////////////

/*

-------- Original Message --------
Subject:     ABC/ZZ integration
Date:     Wed, 29 Sep 2010 00:34:32 -0700
From:     Niklas Een <niklas@een.se>
To:     Alan Mishchenko <alanmi@EECS.Berkeley.EDU>

Hi Alan,

Since the interface is file-based, it is important that we generate 
good, unique filenames (we may run multiple instances of ABC in the 
same directory), so I have attached some portable code for doing that 
(tmpFile.c). You can integrate it appropriately.

This is how my interface is meant to work:

(1) As part of your call to Bip, give it first argument "-abc". 
    This will alter Bip's behavior slightly.

(2) To list the commands, call 'bip -list-commands'. 
    My commands begin with a comma (so that's my prefix).

(3) All commands expect an input file and an output file. 
    The input file should be in AIGER format. 
    The output will be a text file. 
    Example:
       bip -input=tmp.aig -output=tmp.out ,pdr -check -prop=5

    So you just auto-generate the two temporary files (output file is 
    closed and left empty) and stitch the ABC command line at the end. 
    All you need to check for is if the ABC line begins with a comma.

(4) The result written to the output file will contain a number 
    of object. Each object is on a separate line of the form:

    <object name>: <object data>

That is: name, colon, space, data, newline. If you see a name you don't 
recognize, just skip that line (so you will ignore future extensions by me). 
I currently define the following objects:

    result:
    counter-example:
    proof-invariant:
    bug-free-depth:
    abstraction:
  
"result:" is one of "proved", "failed", "undetermined" (=reached resource limit), "error" 
(only used by the abstraction command, and only if resource limit was so tight that the 
abstraction was still empty -- no abstraction is returned in this special case).

"counter-example:" -- same format as before

"proof-invariant:" contains an text-encoded single-output AIG. If you want 
you can parse it and validate the invariant.

"bug-free-depth:" the depth up to which the procedure has checked for counter-example. 
Starts at -1 (not even the initial states have been verified).

"abstraction:" -- same format as before

(5) I propose that you add a command "load_plugin <path/binary>". That way Bob can put 
Bip where ever he likes and just modify his abc_rc file.

The intention is that ABC can read this file and act on it without knowing what 
particular command was used. If there is an abstraction, you will always apply it. 
If there is a "bug-free-depth" you will store that data somewhere so that Bob can query 
it through the Python interface, and so on. If we need different actions for different 
command, then we add a new object for the new action.

// N.

*/

////////////////////////////////////////////////////////////////////////
///                     FUNCTION DEFINITIONS                         ///
////////////////////////////////////////////////////////////////////////

/**Function*************************************************************

  Synopsis    []

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Abc_GetBinaryName(pAbc *Abc_Frame_t, argc int32, argv *string) string {
    var pTemp string;
    var i int32;
    Vec_PtrForEachEntry( char *, pAbc.vPlugInComBinPairs, pTemp, i )
    {
        i++;
        if int32(strings.Compare(pTemp, argv[0])) == 0 {
            return (char *)Vec_PtrEntry( pAbc.vPlugInComBinPairs, i ); }
    }
    assert(false, "0");
    return nil;
}

/**Function*************************************************************

  Synopsis    [Read flop map.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Abc_ManReadFile(pFileName string) *Vec_Str_t {
    var pFile *os.File;
    var vStr *Vec_Str_t;
    var c int32;
    pFile = os.Open(pFileName);
    if pFile == nil {
        fmt.Printf("Cannot open file \"%s\".\n", pFileName);
        return nil;
    }
    vStr = Vec_StrAlloc( 100 );
    for  (c = fgetc(pFile)) != -1  {
        Vec_StrPush( vStr, byte(c) );
    }
    Vec_StrPush( vStr, '\0' );
    _ = (pFile).Close();
    return vStr;
}

/**Function*************************************************************

  Synopsis    [Read flop map.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Abc_ManReadBinary(pFileName string, pToken string) *Vec_Int_t {
    var vMap *Vec_Int_t  = nil;
    var vStr *Vec_Str_t;
    var pStr string;
    var i, Length int32;
    vStr = Abc_ManReadFile( pFileName );
    if vStr == nil {
        return nil; }
    pStr = Vec_StrArray( vStr );
    pStr = strstr( pStr, pToken );
    if pStr != nil {
        pStr  += len(pToken);
        vMap   = Vec_IntAlloc( 100 );
        Length = len(pStr);
        for  i = 0; i < Length; i++  {
            if pStr[i] == '0' || pStr[i] == '?' {
                Vec_IntPush( vMap, 0 );
            } else if pStr[i] == '1' {
                Vec_IntPush( vMap, 1 ); }
            if ('a' <= pStr[i] && pStr[i] <= 'z') || 
                 ('A' <= pStr[i] && pStr[i] <= 'Z') {
                break; }
        }
    }
    Vec_StrFree( vStr );
    return vMap;
}

/**Function*************************************************************

  Synopsis    [Read flop map.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Abc_ManReadInteger(pFileName string, pToken string) int32 {
    var Result int32  = -1;
    var vStr *Vec_Str_t;
    var pStr string;
    vStr = Abc_ManReadFile( pFileName );
    if vStr == nil {
        return -1; }
    pStr = Vec_StrArray( vStr );
    pStr = strstr( pStr, pToken );
    if pStr != nil {
        Result = strconv.Atoi(pStr + len(pToken)); }
    Vec_StrFree( vStr );
    return Result;
}

/**Function*************************************************************

  Synopsis    [Read flop map.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Abc_ManReadStatus(pFileName string, pToken string) int32 {
    var Result int32  = -1;
    var vStr *Vec_Str_t;
    var pStr string;
    vStr = Abc_ManReadFile( pFileName );
    if vStr == nil {
        return -1; }
    pStr = Vec_StrArray( vStr );
    pStr = strstr( pStr, pToken );
    if pStr != nil {
        if strncmp(pStr+8, "proved", 6) == 0 {
            Result = 1; 
        } else if strncmp(pStr+8, "failed", 6) == 0 {
            Result = 0; } 
    }
    Vec_StrFree( vStr );
    return Result;
}

/**Function*************************************************************

  Synopsis    [Work-around to insert 0s for PIs without fanout.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Abc_ManExpandCex(pGia *Gia_Man_t, vCex *Vec_Int_t) *Vec_Int_t {
    var vCexNew *Vec_Int_t;
    var pObj *Gia_Obj_t;
    var i, k int32;

    // start with register outputs
    vCexNew = Vec_IntAlloc( Vec_IntSize(vCex) );
    Gia_ManForEachRo( pGia, pObj, i )
        Vec_IntPush( vCexNew, 0 );

    ABC_FREE( pGia.pRefs );
    Gia_ManCreateRefs( pGia );
    k = Gia_ManRegNum( pGia );
    for {
        Gia_ManForEachPi( pGia, pObj, i )
        {
            if Gia_ObjRefNum(pGia, pObj) == 0 {
                Vec_IntPush( vCexNew, 0 );
            } else {
                if k == Vec_IntSize(vCex) {
                    break; }
                Vec_IntPush( vCexNew, Vec_IntEntry(vCex, k++) );
            }
        }
        if k == Vec_IntSize(vCex) {
            break; }
    }
    return vCexNew;
}

/**Function*************************************************************

  Synopsis    [Procedure to convert the AIG from text into binary form.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
static unsigned textToBin(char* text, unsigned long text_sz)
{
    var dst string  = text;
    char* src = text;
    var sz, i uint32;
    sscanf(src, "%u ", &sz);
    for *src++ != ' ' {; }
    for  i = 0; i < sz; i += 3  {
        dst[0] = byte(( uint32(src[0]) - '0'))       | ((uint32(src[1]) - '0') << 6);
        dst[1] = byte(((uint32(src[1]) - '0') >> 2)) | ((uint32(src[2]) - '0') << 4);
        dst[2] = byte(((uint32(src[2]) - '0') >> 4)) | ((uint32(src[3]) - '0') << 2);
        src += 4;
        dst += 3;
    }
    return sz;
}

/**Function*************************************************************

  Synopsis    [Derives AIG from the text string in the file.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Abc_ManReadAig(pFileName string, pToken string) *Gia_Man_t {
    var pGia *Gia_Man_t  = nil;
    var nBinaryPart uint32;
    var vStr *Vec_Str_t;
    var pStr, pEnd string;
    vStr = Abc_ManReadFile( pFileName );
    if vStr == nil {
        return nil; }
    pStr = Vec_StrArray( vStr );
    pStr = strstr( pStr, pToken );
    if pStr != nil {
        // skip token
        pStr += len(pToken);
        // skip spaces
        for  *pStr == ' '  {
            pStr++;
        }
        // set the end at newline
        for  pEnd = pStr; *pEnd; pEnd++  {
            if *pEnd == '\r' || *pEnd == '\n' {
                *pEnd = 0;
                break;
            }
        }
        // convert into binary AIGER
        nBinaryPart = textToBin( pStr, len(pStr) );
        // dump it into file
        if false {
            var pFile *os.File  = os.OpenFile("test.aig", os.O_WRONLY | os.O_CREATE | os.O_TRUNC, 0666);
            fwrite( pStr, 1, nBinaryPart, pFile );
            _ = (pFile).Close();
        }
        // derive AIG
        pGia = Gia_AigerReadFromMemory( pStr, nBinaryPart, 0, 0, 0 ); 
    }
    Vec_StrFree( vStr );
    return pGia;

}

/**Function*************************************************************

  Synopsis    []

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Cmd_CommandAbcPlugIn(pAbc *Abc_Frame_t, argc int32, argv *string) int32 {
    var pFileIn, pFileOut string;
    var vCommand *Vec_Str_t;
    var vCex *Vec_Int_t;
    var pFile *os.File;
    var pGia *Gia_Man_t;
    var i, fd int32;
    var clk abctime;
    var fLeaveFiles int32;
/*
    Abc_Ntk_t * pNtk = Abc_FrameReadNtk(pAbc);
    if ( pNtk == NULL )
    {
        Abc_Print( -1, "Current network does not exist\n" );
        return 1;
    }
    if ( !Abc_NtkIsStrash( pNtk) )
    {
        Abc_Print( -1, "The current network is not an AIG. Cannot continue.\n" );
        return 1;
    }
*/

    if pAbc.pGia == nil {
        if argc == 2 && int32(strings.Compare(argv[1], "-h")) == 0 {
            // Run command to produce help string:
            vCommand = Vec_StrAlloc( 100 );
            Vec_StrAppend( vCommand, Abc_GetBinaryName( pAbc, argc, argv ) );
            Vec_StrAppend( vCommand, " -abc " );
            Vec_StrAppend( vCommand, argv[0] );
            Vec_StrAppend( vCommand, " -h" );
            Vec_StrPush( vCommand, 0 );
            Util_SignalSystem( Vec_StrArray(vCommand) );
            Vec_StrFree( vCommand );
        
        } else {
            Abc_Print( -1, "Current AIG does not exist (try command &ps).\n" );
        }
        return 1;
    }

    // create temp file
    fd = Util_SignalTmpFile( "__abctmp_", ".aig", &pFileIn );
    if fd == -1 {
        Abc_Print( -1, "Cannot create a temporary file.\n" );
        return 1;
    }
#ifdef WIN32
    _close( fd );
#else
    close( fd );
#endif

    // create temp file
    fd = Util_SignalTmpFile( "__abctmp_", ".out", &pFileOut );
    if fd == -1 {
        ABC_FREE( pFileIn );
        Abc_Print( -1, "Cannot create a temporary file.\n" );
        return 1;
    }
#ifdef WIN32
    _close( fd );
#else
    close( fd );
#endif


    // write current network into a file
/*
    {
        extern Aig_Man_t * Abc_NtkToDar( Abc_Ntk_t * pNtk, int fExors, int fRegisters );
        Aig_Man_t * pAig;
        pAig = Abc_NtkToDar( pNtk, 0, 1 );
        Ioa_WriteAiger( pAig, pFileIn, 0, 0 );
        Aig_ManStop( pAig );
    }
*/ 
    // check what to do with the files
    fLeaveFiles = 0;
    if int32(strings.Compare(argv[argc-1], "!")) == 0 {
        Abc_Print( 0, "Input file \"%s\" and output file \"%s\" are not deleted.\n", pFileIn, pFileOut );
        fLeaveFiles = 1;
        argc--;
    }

    // create input file
    Gia_AigerWrite( pAbc.pGia, pFileIn, 0, 0, 0 );

    // create command line
    vCommand = Vec_StrAlloc( 100 );
    Vec_StrAppend( vCommand, Abc_GetBinaryName( pAbc, argc, argv ) );
    // add input/output file
    Vec_StrAppend( vCommand, " -abc" );
//    Vec_StrAppend( vCommand, " -input=C:/_projects/abc/_TEST/hwmcc/139442p0.aig" );
    Vec_StrAppend( vCommand, " -input=" );
    Vec_StrAppend( vCommand, pFileIn );
    Vec_StrAppend( vCommand, " -output=" );
    Vec_StrAppend( vCommand, pFileOut );
    // add other arguments
    for  i = 0; i < argc; i++  {
        Vec_StrAppend( vCommand, " " );
        Vec_StrAppend( vCommand, argv[i] );
    }
    Vec_StrPush( vCommand, 0 );

    // run the command line
//printf( "Running command line: %s\n", Vec_StrArray(vCommand) ); 

    clk = Abc_Clock();
    if Util_SignalSystem( Vec_StrArray(vCommand) ) {
        Abc_Print( -1, "The following command has returned non-zero exit status:\n" );
        Abc_Print( -1, "\"%s\"\n", Vec_StrArray(vCommand) );
        return 1;
    }
    clk = Abc_Clock() - clk;
    Vec_StrFree( vCommand );

    // check if the output file exists
    if (pFile = os.Open(pFileOut)) == nil {
        Abc_Print( -1, "There is no output file \"%s\".\n", pFileOut );
        return 1;
    }
    _ = (pFile).Close();

    // process the output
    if Extra_FileSize(pFileOut) > 0 {
        // get status
        pAbc.Status  = Abc_ManReadStatus( pFileOut, "result:" );
        // get bug-free depth
        pAbc.nFrames = Abc_ManReadInteger( pFileOut, "bug-free-depth:" );
        // get abstraction
        pAbc.pGia.vFlopClasses = Abc_ManReadBinary( pFileOut, "abstraction:" );
        // get counter-example
        vCex = Abc_ManReadBinary( pFileOut, "counter-example:" );
        if vCex != nil {
            var nFrames, nRemain int32;

            nFrames = (Vec_IntSize(vCex) - Gia_ManRegNum(pAbc.pGia)) / Gia_ManPiNum(pAbc.pGia);
            nRemain = (Vec_IntSize(vCex) - Gia_ManRegNum(pAbc.pGia)) % Gia_ManPiNum(pAbc.pGia);
            if nRemain != 0 {   
                var vTemp *Vec_Int_t;
                Abc_Print( 1, "Adjusting counter-example by adding zeros for PIs without fanout.\n" );
                // expand the counter-example to include PIs without fanout
                vCex = Abc_ManExpandCex( pAbc.pGia, vTemp = vCex );
                Vec_IntFree( vTemp );
            }

            nFrames = (Vec_IntSize(vCex) - Gia_ManRegNum(pAbc.pGia)) / Gia_ManPiNum(pAbc.pGia);
            nRemain = (Vec_IntSize(vCex) - Gia_ManRegNum(pAbc.pGia)) % Gia_ManPiNum(pAbc.pGia);
            if nRemain != 0 {
                Abc_Print( 1, "Counter example has a wrong length.\n" );
            } else {
                Abc_Print( 1, "Problem is satisfiable. Found counter-example in frame %d.  ", nFrames-1 );
                Abc_PrintTime( 1, "Time", clk );
                ABC_FREE( pAbc.pCex );
                pAbc.pCex = Abc_CexCreate( Gia_ManRegNum(pAbc.pGia), Gia_ManPiNum(pAbc.pGia), Vec_IntArray(vCex), nFrames-1, 0, 0 );

//                Abc_CexPrint( pAbc->pCex );

//                if ( !Gia_ManVerifyCex( pAbc->pGia, pAbc->pCex, 0 ) )
//                    Abc_Print( 1, "Generated counter-example is INVALID.\n" );

                // remporary work-around to detect the output number - October 18, 2010
                pAbc.pCex.iPo = Gia_ManFindFailedPoCex( pAbc.pGia, pAbc.pCex, 0 );
                if pAbc.pCex.iPo == -1 {
                    Abc_Print( 1, "Generated counter-example is INVALID.\n" );
                    ABC_FREE( pAbc.pCex );
                
                } else {
                    Abc_Print( 1, "Returned counter-example successfully verified in ABC.\n" );
                }
            }
            Vec_IntFreeP( &vCex );
        }
        // get AIG if present
        pGia = Abc_ManReadAig( pFileOut, "aig:" );
        if pGia != nil {
            Gia_ManStopP( &pAbc.pGia );
            pAbc.pGia = pGia;                
        }
    }    

    // clean up
    Util_SignalTmpFileRemove( pFileIn, fLeaveFiles );
    Util_SignalTmpFileRemove( pFileOut, fLeaveFiles );
    
    ABC_FREE( pFileIn );
    ABC_FREE( pFileOut );
    return 0;
}

/**Function*************************************************************

  Synopsis    []

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Cmd_CommandAbcLoadPlugIn(pAbc *Abc_Frame_t, argc int32, argv *string) int32 {
    var fPath, fVerbose int32;
    var fd int32 = -1;
    var RetValue int32 = -1;
    var c int32;
    var pFile *os.File  = nil;
    var pStrDirBin string = nil;
    var pStrSection string = nil;
    var sCommandLine *Vec_Str_t  = nil;
    var pTempFile string  = nil;
    var pBuffer [1000]byte;

    // set defaults
    fPath = 0;
    fVerbose = 0;

    Extra_UtilGetoptReset();
    for  ( c = Extra_UtilGetopt( argc, argv, "vph" ) ) != -1  {
        switch  c  {
        case 'p':
            fPath ^= 1;
            break;
        case 'v':
            fVerbose ^= 1;
            break;
        default:
        goto usage;
    }
    }

    if argc != globalUtilOptind + 2 {
        goto usage; }

    pStrDirBin  = argv[argc-2];
    pStrSection = argv[argc-1];

    // check if the file exists
    if !fPath {
        var pFile *os.File  = os.Open(pStrDirBin);

        if !(pFile != nil) {
            Abc_Print( ABC_ERROR, "Cannot run the binary \"%s\". File does not exist.\n", pStrDirBin );
            goto cleanup;
    }

    _ = (pFile).Close();
    }

    // create temp file
    fd = Util_SignalTmpFile( "__abctmp_", ".txt", &pTempFile );

    if fd == -1 {
        Abc_Print( ABC_ERROR, "Cannot create a temporary file.\n" );
        goto cleanup;
    }

#ifdef WIN32
    _close( fd );
#else
    close( fd );
#endif

    // get command list
    sCommandLine = Vec_StrAlloc(1000);

    Vec_StrPrintF(sCommandLine, "%s -abc -list-commands > %s", pStrDirBin, pTempFile );
    Vec_StrPush(sCommandLine, '\0');

    if fVerbose != 0 {
        Abc_Print(ABC_VERBOSE, "Running command %s\n", Vec_StrArray(sCommandLine));
    }

    RetValue = Util_SignalSystem( Vec_StrArray(sCommandLine) );

    if RetValue != 0 {
        Abc_Print( ABC_ERROR, "Command \"%s\" failed.\n", Vec_StrArray(sCommandLine) );
        goto cleanup;
    }

    // create commands
    pFile = os.Open(pTempFile);

    if pFile == nil {
        Abc_Print( -1, "Cannot open file with the list of commands.\n" );

        RetValue = -1;
        goto cleanup;
    }

    for  fgets( pBuffer, 1000, pFile ) != nil  {
        if pBuffer[strlen(pBuffer)-1] == '\n' {
            pBuffer[strlen(pBuffer)-1] = 0; }

        Cmd_CommandAdd( pAbc, pStrSection, pBuffer, Cmd_CommandAbcPlugIn, 1 );

        Vec_PtrPush( pAbc.vPlugInComBinPairs, Extra_UtilStrsav(pBuffer) );
        Vec_PtrPush( pAbc.vPlugInComBinPairs, Extra_UtilStrsav(pStrDirBin) );

        if fVerbose != 0 {
            Abc_Print(ABC_VERBOSE, "Creating command %s with binary %s\n", pBuffer, pStrDirBin);
    }
    }

cleanup:

    if pFile != nil {
    _ = (pFile).Close(); }

    if pTempFile {
    Util_SignalTmpFileRemove( pTempFile, 0 ); }

    Vec_StrFreeP(&sCommandLine);

    ABC_FREE( pTempFile );

    return RetValue;

usage:

    Abc_Print( -2, "usage: load_plugin [-pvh] <plugin_dir\\binary_name> <section_name>\n" );
    Abc_Print( -2, "\t        loads external binary as a plugin\n" );
    Abc_Print( -2, "\t-p    : toggle searching the command in PATH [default = %s].\n", fPath? "yes": "no" );
    Abc_Print( -2, "\t-v    : enable verbose output [default = %s].\n", fVerbose? "yes": "no" );
    Abc_Print( -2, "\t-h    : print the command usage\n");

    return 1;
}


////////////////////////////////////////////////////////////////////////
///                       END OF FILE                                ///
////////////////////////////////////////////////////////////////////////


//ABC_NAMESPACE_IMPL_END
