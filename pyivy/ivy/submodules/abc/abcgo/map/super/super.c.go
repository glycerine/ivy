package super

import (
	"fmt"
	"strconv"
)

import "os"

/**CFile****************************************************************

  FileName    [super.c]

  PackageName [MVSIS 1.3: Multi-valued logic synthesis system.]

  Synopsis    [Pre-computation of supergates.]

  Author      [MVSIS Group]
  
  Affiliation [UC Berkeley]

  Date        [Ver. 1.0. Started - August 18, 2003.]

  Revision    [$Id: super.c,v 1.6 2004/10/30 20:51:11 satrajit Exp $]

***********************************************************************/

//#include "superInt.h"
//#include "base/main/mainInt.h"
//#include "map/mio/mio.h"

//ABC_NAMESPACE_IMPL_START


////////////////////////////////////////////////////////////////////////
///                        DECLARATIONS                              ///
////////////////////////////////////////////////////////////////////////

// static int Super_CommandSupergates   ( Abc_Frame_t * pAbc, int argc, char **argv );

// static int Super_CommandSupergatesAnd( Abc_Frame_t * pAbc, int argc, char **argv );


////////////////////////////////////////////////////////////////////////
///                     FUNCTION DEFINITIONS                         ///
////////////////////////////////////////////////////////////////////////

/**Function*************************************************************

  Synopsis    []

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Super_Init(pAbc *Abc_Frame_t) {
    Cmd_CommandAdd( pAbc, "SC mapping",  "super",   Super_CommandSupergates,    0 ); 
    Cmd_CommandAdd( pAbc, "SC mapping",  "super2",  Super_CommandSupergatesAnd, 0 ); 
}

/**Function*************************************************************

  Synopsis    []

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Super_End(pAbc *Abc_Frame_t) {
}



/**Function*************************************************************

  Synopsis    []

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Super_CommandSupergatesAnd(pAbc *Abc_Frame_t, argc int32, argv *string) int32 {
    var pOut, pErr *os.File;
    var nVarsMax, nLevels int32;
    var fVerbose int32;
    var c int32;

    pOut = Abc_FrameReadOut(pAbc);
    pErr = Abc_FrameReadErr(pAbc);

    // set the defaults
    nVarsMax = 4;
    nLevels  = 3;
    fVerbose = 0;
    Extra_UtilGetoptReset();
    for  (c = Extra_UtilGetopt(argc, argv, "ILvh")) != -1  {
        switch c {
            case 'I':
                nVarsMax = strconv.Atoi(argv[globalUtilOptind]);
                globalUtilOptind++;
                if nVarsMax < 0 { 
                    goto usage; }
                break;
            case 'L':
                nLevels = strconv.Atoi(argv[globalUtilOptind]);
                globalUtilOptind++;
                if nLevels < 0 { 
                    goto usage; }
                break;
            case 'v':
                fVerbose ^= 1;
                break;
            case 'h':
                goto usage;
                break;
            default:
                goto usage;
        }
    }

    Super2_Precompute( nVarsMax, nLevels, fVerbose );

    return 0;

usage:
    fmt.Fprintf(pErr, "usage: super2 [-IL num] [-vh]\n");
    fmt.Fprintf(pErr, "\t         precomputes the supergates composed of AND2s and INVs\n");  
    fmt.Fprintf(pErr, "\t-I num : the max number of inputs to the supergate [default = %d]\n", nVarsMax);
    fmt.Fprintf(pErr, "\t-L num : the max number of logic levels of gates [default = %d]\n", nLevels);
    fmt.Fprintf(pErr, "\t-v     : enable verbose output\n");
    fmt.Fprintf(pErr, "\t-h     : print the help message\n");
    return 1;       /* error exit */
}


/**Function*************************************************************

  Synopsis    []

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Super_CommandSupergates(pAbc *Abc_Frame_t, argc int32, argv *string) int32 {
    var pFile *os.File;
    var pOut, pErr *os.File;
    var pLib *Mio_Library_t;
    var FileName, ExcludeFile string;
    var DelayLimit float32;
    var AreaLimit float32;
    var fSkipInvs int32;
    var fWriteOldFormat int32; 
    var nVarsMax, nLevels, nGatesMax, TimeLimit int32;
    var fVerbose int32;
    var c int32;

    pOut = Abc_FrameReadOut(pAbc);
    pErr = Abc_FrameReadErr(pAbc);

    // set the defaults
    nVarsMax   = 5;
    nLevels    = 2;
    DelayLimit = 0;
    AreaLimit  = 0;
    nGatesMax  = 0;
    TimeLimit  = 0;
    fSkipInvs  = 1;
    fVerbose   = 0;
    fWriteOldFormat = 0;
    ExcludeFile = 0;

    Extra_UtilGetoptReset();
    for  (c = Extra_UtilGetopt(argc, argv, "ILNTDAEsovh")) != -1  {
        switch c {
            case 'I':
                nVarsMax = strconv.Atoi(argv[globalUtilOptind]);
                globalUtilOptind++;
                if nVarsMax < 0 { 
                    goto usage; }
                break;
            case 'L':
                nLevels = strconv.Atoi(argv[globalUtilOptind]);
                globalUtilOptind++;
                if nLevels < 0 { 
                    goto usage; }
                break;
            case 'N':
                nGatesMax = strconv.Atoi(argv[globalUtilOptind]);
                globalUtilOptind++;
                if nGatesMax < 0 { 
                    goto usage; }
                break;
            case 'T':
                TimeLimit = strconv.Atoi(argv[globalUtilOptind]);
                globalUtilOptind++;
                if TimeLimit < 0 { 
                    goto usage; }
                break;
            case 'D':
                DelayLimit = float32(atof(argv[globalUtilOptind]));
                globalUtilOptind++;
                if DelayLimit <= 0.0 { 
                    goto usage; }
                break;
            case 'A':
                AreaLimit = float32(atof(argv[globalUtilOptind]));
                globalUtilOptind++;
                if AreaLimit <= 0.0 { 
                    goto usage; }
                break;
            case 'E':
                ExcludeFile = argv[globalUtilOptind];
                if ExcludeFile == 0 {
                    goto usage; }
                globalUtilOptind++;
                break;
            case 's':
                fSkipInvs ^= 1;
                break;
            case 'o':
                fWriteOldFormat ^= 1;
                break;
            case 'v':
                fVerbose ^= 1;
                break;
            case 'h':
                goto usage;
                break;
            default:
                goto usage;
        }
    }


    if argc != globalUtilOptind + 1 {
        fmt.Fprintf(pErr, "The genlib library file should be given on the command line.\n");
        goto usage;
    }

    if nVarsMax < 2 || nVarsMax > 6 {
        fmt.Fprintf(pErr, "The max number of variables (%d) should be more than 1 and less than 7.\n", nVarsMax);
        goto usage;
    }

    // get the input file name
    FileName = argv[globalUtilOptind];
    if (pFile = Io_FileOpen( FileName, "open_path", "r", 0 )) == nil {
//    if ( (pFile = fopen( FileName, "r" )) == NULL )
    
        fmt.Fprintf(pErr, "Cannot open input file \"%s\". ", FileName);
        if FileName = Extra_FileGetSimilarName( FileName, ".genlib", ".lib", ".gen", ".g", nil ) {
            fmt.Fprintf(pErr, "Did you mean \"%s\"?", FileName); }
        fmt.Fprintf(pErr, "\n");
        return 1;
    }
    _ = (pFile).Close();

    // set the new network
    pLib = Mio_LibraryRead( FileName, nil, ExcludeFile, fVerbose );
    if pLib == nil {
        fmt.Fprintf(pErr, "Reading library has failed.\n");
        goto usage;
    }

    // compute the gates
    FileName = Extra_FileNameGenericAppend(Mio_LibraryReadName(pLib), ".super");
    Super_Precompute( pLib, nVarsMax, nLevels, nGatesMax, DelayLimit, AreaLimit, TimeLimit, fSkipInvs, fVerbose, FileName );

    // delete the library
    Mio_LibraryDelete( pLib );
    return 0;

usage:
    fmt.Fprintf(pErr, "usage: super [-ILNT num] [-DA float] [-E file] [-sovh] <genlib_file>\n");
    fmt.Fprintf(pErr, "\t         precomputes the supergates for the given genlib library\n");  
    fmt.Fprintf(pErr, "\t-I num   : the max number of supergate inputs [default = %d]\n", nVarsMax);
    fmt.Fprintf(pErr, "\t-L num   : the max number of levels of gates [default = %d]\n", nLevels);
    fmt.Fprintf(pErr, "\t-N num   : the limit on the number of considered supergates [default = %d]\n", nGatesMax);
    fmt.Fprintf(pErr, "\t-T num   : the approximate runtime limit in seconds [default = %d]\n", TimeLimit);
    fmt.Fprintf(pErr, "\t-D float : the max delay of the supergates [default = %.2f]\n", DelayLimit);
    fmt.Fprintf(pErr, "\t-A float : the max area of the supergates [default = %.2f]\n", AreaLimit);
    fmt.Fprintf(pErr, "\t-E file  : file contains list of genlib gates to exclude\n");
    fmt.Fprintf(pErr, "\t-s       : toggle the use of inverters at the inputs [default = %s]\n", (fSkipInvs? "no": "yes"));
    fmt.Fprintf(pErr, "\t-o       : toggle dumping the supergate library in old format [default = %s]\n", (fWriteOldFormat? "yes": "no"));
    fmt.Fprintf(pErr, "\t-v       : enable verbose output [default = %s]\n", (fVerbose? "yes" : "no"));
    fmt.Fprintf(pErr, "\t-h       : print the help message\n");
    fmt.Fprintf(pErr, "\n");
    fmt.Fprintf(pErr, "\tHere is a piece of advice on precomputing supergate libraries:\n");
    fmt.Fprintf(pErr, "\t\n");
    fmt.Fprintf(pErr, "\tStart with the number of inputs equal to 5 (-I 5), the number of \n");
    fmt.Fprintf(pErr, "\tlevels equal to 2 (-L 2), the delay equal to 2-3 delays of inverter, \n");
    fmt.Fprintf(pErr, "\tthe area equal to 2-3 areas of two input NAND, and runtime limit equal \n");
    fmt.Fprintf(pErr, "\tto 10 seconds (-T 10). Run precomputation and learn from the result.\n");
    fmt.Fprintf(pErr, "\tDetermine what parameter is most constraining and try to increase \n");
    fmt.Fprintf(pErr, "\tthe value of that parameter. The goal is to have a well-balanced\n");
    fmt.Fprintf(pErr, "\tset of constraints and the resulting supergate library containing\n");
    fmt.Fprintf(pErr, "\tapproximately 5K-20K supergates. Typically, it is better to increase\n");
    fmt.Fprintf(pErr, "\tdelay limit rather than area limit, because having large-area supergates\n");
    fmt.Fprintf(pErr, "\tmay result in a considerable increase in area.\n");
    fmt.Fprintf(pErr, "\t\n");
    fmt.Fprintf(pErr, "\tNote that a good supergate library for experiments typically can be \n");
    fmt.Fprintf(pErr, "\tprecomputed in 30 sec or less. Increasing runtime limit makes sense when\n");
    fmt.Fprintf(pErr, "\tother parameters are well-balanced and it is needed to enumerate more\n");
    fmt.Fprintf(pErr, "\tchoices to have a good result. In the end, to compute the final library\n");
    fmt.Fprintf(pErr, "\tthe runtime can be set to 300 sec to ensure the ultimate quality.\n");
    fmt.Fprintf(pErr, "\tIn some cases, the runtime has to be reduced if the supergate library\n");
    fmt.Fprintf(pErr, "\tcontains too many supergates (> 500K).\n");
    fmt.Fprintf(pErr, "\t\n");
    fmt.Fprintf(pErr, "\tWhen precomputing libraries of 6 inputs (-i 6), start with even more \n");
    fmt.Fprintf(pErr, "\trestricted parameters and gradually increase them until the goal is met.\n");
    fmt.Fprintf(pErr, "\t\n");
    return 1;       /* error exit */
}

////////////////////////////////////////////////////////////////////////
///                       END OF FILE                                ///
////////////////////////////////////////////////////////////////////////


//ABC_NAMESPACE_IMPL_END

