package if_

import "fmt"

import "os"

/**CFile****************************************************************

  FileName    [ifCom.c]

  SystemName  [ABC: Logic synthesis and verification system.]

  PackageName [FPGA mapping based on priority cuts.]

  Synopsis    [Command handlers.]

  Author      [Alan Mishchenko]
  
  Affiliation [UC Berkeley]

  Date        [Ver. 1.0. Started - November 21, 2006.]

  Revision    [$Id: ifCom.c,v 1.00 2006/11/21 00:00:00 alanmi Exp $]

***********************************************************************/

//#include "if.h"
//#include "base/main/main.h"

//ABC_NAMESPACE_IMPL_START


////////////////////////////////////////////////////////////////////////
///                        DECLARATIONS                              ///
////////////////////////////////////////////////////////////////////////

// static int If_CommandReadLut ( Abc_Frame_t * pAbc, int argc, char **argv );

// static int If_CommandPrintLut( Abc_Frame_t * pAbc, int argc, char **argv );

// static int If_CommandReadBox ( Abc_Frame_t * pAbc, int argc, char **argv );

// static int If_CommandPrintBox( Abc_Frame_t * pAbc, int argc, char **argv );


////////////////////////////////////////////////////////////////////////
///                     FUNCTION DEFINITIONS                         ///
////////////////////////////////////////////////////////////////////////

/**Function*************************************************************

  Synopsis    [Package initialization procedure.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func If_Init(pAbc *Abc_Frame_t) {
    // set the default library
    var s_LutLib If_LibLut_t  = If_LibLut_t{ "lutlib", 4, 0, {0,1,1,1,1}, {{0},{1},{1},{1},{1}} };
    Abc_FrameSetLibLut( If_LibLutDup(&s_LutLib) );

    Cmd_CommandAdd( pAbc, "FPGA mapping", "read_lut",   If_CommandReadLut,   0 ); 
    Cmd_CommandAdd( pAbc, "FPGA mapping", "print_lut",  If_CommandPrintLut,  0 ); 

    Cmd_CommandAdd( pAbc, "FPGA mapping", "read_box",   If_CommandReadBox,   0 ); 
    Cmd_CommandAdd( pAbc, "FPGA mapping", "print_box",  If_CommandPrintBox,  0 ); 
}

/**Function*************************************************************

  Synopsis    [Package ending procedure.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func If_End(pAbc *Abc_Frame_t) {
    If_LibLutFree( (If_LibLut_t *)   Abc_FrameReadLibLut() );
    If_LibBoxFree( (If_LibBox_t *)Abc_FrameReadLibBox() );
}

/**Function*************************************************************

  Synopsis    [Command procedure to read LUT libraries.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func If_CommandReadLut(pAbc *Abc_Frame_t, argc int32, argv *string) int32 {
    var pFile *os.File;
    var pOut, pErr *os.File;
    var pLib *If_LibLut_t;
    var pNet *Abc_Ntk_t;
    var FileName string;
    var fVerbose int32;
    var c int32;

    pNet = Abc_FrameReadNtk(pAbc);
    pOut = Abc_FrameReadOut(pAbc);
    pErr = Abc_FrameReadErr(pAbc);

    // set the defaults
    fVerbose = 1;
    Extra_UtilGetoptReset();
    for  (c = Extra_UtilGetopt(argc, argv, "vh")) != -1  {
        switch c {
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
        goto usage; }

    // get the input file name
    FileName = argv[globalUtilOptind];
    if (pFile = os.Open(FileName)) == nil {
        fmt.Fprintf(pErr, "Cannot open input file \"%s\". ", FileName);
        if FileName = Extra_FileGetSimilarName( FileName, ".genlib", ".lib", ".gen", ".g", nil ) {
            fmt.Fprintf(pErr, "Did you mean \"%s\"?", FileName); }
        fmt.Fprintf(pErr, "\n");
        return 1;
    }
    _ = (pFile).Close();

    // set the new network
    pLib = If_LibLutRead( FileName );
    if pLib == nil {
        fmt.Fprintf(pErr, "Reading LUT library has failed.\n");
        goto usage;
    }
    // replace the current library
    If_LibLutFree( (If_LibLut_t *)Abc_FrameReadLibLut() );
    Abc_FrameSetLibLut( pLib );
    return 0;

usage:
    fmt.Fprintf(pErr, "\nusage: read_lut [-vh]\n");
    fmt.Fprintf(pErr, "\t          read the LUT library from the file\n");  
    fmt.Fprintf(pErr, "\t-v      : toggles enabling of verbose output [default = %s]\n", (fVerbose? "yes" : "no"));
    fmt.Fprintf(pErr, "\t-h      : print the command usage\n");
    fmt.Fprintf(pErr, "\t                                        \n");
    fmt.Fprintf(pErr, "\t          File format for a LUT library:\n");
    fmt.Fprintf(pErr, "\t          (the default library is shown)\n");
    fmt.Fprintf(pErr, "\t                                        \n");
    fmt.Fprintf(pErr, "\t          # The area/delay of k-variable LUTs:\n");
    fmt.Fprintf(pErr, "\t          # k  area   delay\n");
    fmt.Fprintf(pErr, "\t          1      1      1\n");
    fmt.Fprintf(pErr, "\t          2      2      2\n");
    fmt.Fprintf(pErr, "\t          3      4      3\n");
    fmt.Fprintf(pErr, "\t          4      8      4\n");
    fmt.Fprintf(pErr, "\t          5     16      5\n");
    fmt.Fprintf(pErr, "\t          6     32      6\n");
    return 1;       /* error exit */
}

/**Function*************************************************************

  Synopsis    [Command procedure to read LUT libraries.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func If_CommandPrintLut(pAbc *Abc_Frame_t, argc int32, argv *string) int32 {
    var pOut, pErr *os.File;
    var pNet *Abc_Ntk_t;
    var fVerbose int32;
    var c int32;

    pNet = Abc_FrameReadNtk(pAbc);
    pOut = Abc_FrameReadOut(pAbc);
    pErr = Abc_FrameReadErr(pAbc);

    // set the defaults
    fVerbose = 1;
    Extra_UtilGetoptReset();
    for  (c = Extra_UtilGetopt(argc, argv, "vh")) != -1  {
        switch c {
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

    if argc != globalUtilOptind {
        goto usage; }

    // set the new network
    If_LibLutPrint( (If_LibLut_t *)Abc_FrameReadLibLut() );
    return 0;

usage:
    fmt.Fprintf(pErr, "\nusage: print_lut [-vh]\n");
    fmt.Fprintf(pErr, "\t          print the current LUT library\n");  
    fmt.Fprintf(pErr, "\t-v      : toggles enabling of verbose output [default = %s]\n", (fVerbose? "yes" : "no"));
    fmt.Fprintf(pErr, "\t-h      : print the command usage\n");
    return 1;       /* error exit */
}

/**Function*************************************************************

  Synopsis    []

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func If_CommandReadBox(pAbc *Abc_Frame_t, argc int32, argv *string) int32 {
    var pFile *os.File;
    var pOut, pErr *os.File;
    var pLib *If_LibBox_t;
    var pNet *Abc_Ntk_t;
    var FileName string;
    var fExtended int32;
    var fVerbose int32;
    var c int32;

    pNet = Abc_FrameReadNtk(pAbc);
    pOut = Abc_FrameReadOut(pAbc);
    pErr = Abc_FrameReadErr(pAbc);

    // set the defaults
    fExtended = 0;
    fVerbose = 1;
    Extra_UtilGetoptReset();
    for  (c = Extra_UtilGetopt(argc, argv, "evh")) != -1  {
        switch c {
            case 'e':
                fExtended ^= 1;
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
        goto usage; }

    // get the input file name
    FileName = argv[globalUtilOptind];
    if (pFile = os.Open(FileName)) == nil {
        fmt.Fprintf(pErr, "Cannot open input file \"%s\". ", FileName);
        if FileName = Extra_FileGetSimilarName( FileName, ".genlib", ".lib", ".gen", ".g", nil ) {
            fmt.Fprintf(pErr, "Did you mean \"%s\"?", FileName); }
        fmt.Fprintf(pErr, "\n");
        return 1;
    }
    _ = (pFile).Close();

    // set the new network
    if fExtended != 0 {
        pLib = If_LibBoxRead2( FileName );
    } else {
        pLib = If_LibBoxRead( FileName );
    }
    if pLib == nil {
        fmt.Fprintf(pErr, "Reading box library has failed.\n");
        goto usage;
    }
    // replace the current library
    If_LibBoxFree( (If_LibBox_t *)Abc_FrameReadLibBox() );
    Abc_FrameSetLibBox( pLib );
    return 0;

usage:
    fmt.Fprintf(pErr, "\nusage: read_box [-evh]\n");
    fmt.Fprintf(pErr, "\t          read the box library from the file\n");  
    fmt.Fprintf(pErr, "\t-e      : toggles reading extended format [default = %s]\n", (fExtended? "yes" : "no"));
    fmt.Fprintf(pErr, "\t-v      : toggles enabling of verbose output [default = %s]\n", (fVerbose? "yes" : "no"));
    fmt.Fprintf(pErr, "\t-h      : print the command usage\n");
    return 1;       /* error exit */
}

/**Function*************************************************************

  Synopsis    [Command procedure to read LUT libraries.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func If_CommandPrintBox(pAbc *Abc_Frame_t, argc int32, argv *string) int32 {
    var pOut, pErr *os.File;
    var pNet *Abc_Ntk_t;
    var fVerbose int32;
    var c int32;

    pNet = Abc_FrameReadNtk(pAbc);
    pOut = Abc_FrameReadOut(pAbc);
    pErr = Abc_FrameReadErr(pAbc);

    // set the defaults
    fVerbose = 1;
    Extra_UtilGetoptReset();
    for  (c = Extra_UtilGetopt(argc, argv, "vh")) != -1  {
        switch c {
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

    if argc != globalUtilOptind {
        goto usage; }

    // set the new network
    If_LibBoxPrint( os.Stdout, (If_LibBox_t *)Abc_FrameReadLibBox() );
    return 0;

usage:
    fmt.Fprintf(pErr, "\nusage: print_box [-vh]\n");
    fmt.Fprintf(pErr, "\t          print the current box library\n");  
    fmt.Fprintf(pErr, "\t-v      : toggles enabling of verbose output [default = %s]\n", (fVerbose? "yes" : "no"));
    fmt.Fprintf(pErr, "\t-h      : print the command usage\n");
    return 1;       /* error exit */
}


////////////////////////////////////////////////////////////////////////
///                       END OF FILE                                ///
////////////////////////////////////////////////////////////////////////


//ABC_NAMESPACE_IMPL_END

