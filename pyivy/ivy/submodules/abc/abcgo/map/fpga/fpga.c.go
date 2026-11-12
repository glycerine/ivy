package fpga

import "fmt"

import "os"

/**CFile****************************************************************

  FileName    [fpga.c]

  PackageName [MVSIS 1.3: Multi-valued logic synthesis system.]

  Synopsis    [Command file for the FPGA package.]

  Author      [MVSIS Group]
  
  Affiliation [UC Berkeley]

  Date        [Ver. 2.0. Started - August 18, 2004.]

  Revision    [$Id: fpga.c,v 1.4 2004/10/28 17:36:07 alanmi Exp $]

***********************************************************************/

//#include "fpgaInt.h"
//#include "base/main/main.h"

//ABC_NAMESPACE_IMPL_START


////////////////////////////////////////////////////////////////////////
///                        DECLARATIONS                              ///
////////////////////////////////////////////////////////////////////////

// static int Fpga_CommandReadLibrary( Abc_Frame_t * pAbc, int argc, char **argv );

// static int Fpga_CommandPrintLibrary( Abc_Frame_t * pAbc, int argc, char **argv );


// the library file format should be as follows:
/*
# The area/delay of k-variable LUTs:
# k  area   delay
1      1      1
2      2      2
3      4      3
4      8      4
5     16      5
6     32      6
*/

////////////////////////////////////////////////////////////////////////
///                     FUNCTION DEFINITIONS                         ///
////////////////////////////////////////////////////////////////////////

/**Function*************************************************************

  Synopsis    [Package initialization procedure.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Fpga_Init(pAbc *Abc_Frame_t) {
    // set the default library
    //Fpga_LutLib_t s_LutLib = { "lutlib", 6, 0, {0,1,2,4,8,16,32}, {{0},{1},{2},{3},{4},{5},{6}} };
//    Fpga_LutLib_t s_LutLib = { "lutlib", 5, 0, {0,1,1,1,1,1}, {{0},{1},{1},{1},{1},{1}} };
    var s_LutLib Fpga_LutLib_t  = Fpga_LutLib_t{ "lutlib", 4, 0, {0,1,1,1,1}, {{0},{1},{1},{1},{1}} };
    //Fpga_LutLib_t s_LutLib = { "lutlib", 3, 0, {0,1,1,1}, {{0},{1},{1},{1}} };

    Abc_FrameSetLibLut( Fpga_LutLibDup(&s_LutLib) );

    Cmd_CommandAdd( pAbc, "FPGA mapping", "read_lut",   Fpga_CommandReadLibrary,   0 ); 
    Cmd_CommandAdd( pAbc, "FPGA mapping", "print_lut",  Fpga_CommandPrintLibrary,  0 ); 
}

/**Function*************************************************************

  Synopsis    [Package ending procedure.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Fpga_End(pAbc *Abc_Frame_t) {
    Fpga_LutLibFree( (Fpga_LutLib_t *)Abc_FrameReadLibLut() );
}


/**Function*************************************************************

  Synopsis    [Command procedure to read LUT libraries.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Fpga_CommandReadLibrary(pAbc *Abc_Frame_t, argc int32, argv *string) int32 {
    var pFile *os.File;
    var pOut, pErr *os.File;
    var pLib *Fpga_LutLib_t;
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
        goto usage;
    }

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
    pLib = Fpga_LutLibRead( FileName, fVerbose );
    if pLib == nil {
        fmt.Fprintf(pErr, "Reading LUT library has failed.\n");
        goto usage;
    }
    // replace the current library
    Fpga_LutLibFree( (Fpga_LutLib_t *)Abc_FrameReadLibLut() );
    Abc_FrameSetLibLut( pLib );
    return 0;

usage:
    fmt.Fprintf(pErr, "usage: read_lut [-vh]\n");
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
func Fpga_CommandPrintLibrary(pAbc *Abc_Frame_t, argc int32, argv *string) int32 {
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
        goto usage;
    }

    // set the new network
    Fpga_LutLibPrint( (Fpga_LutLib_t *)Abc_FrameReadLibLut() );
    return 0;

usage:
    fmt.Fprintf(pErr, "usage: print_lut [-vh]\n");
    fmt.Fprintf(pErr, "\t          print the current LUT library\n");  
    fmt.Fprintf(pErr, "\t-v      : toggles enabling of verbose output [default = %s]\n", (fVerbose? "yes" : "no"));
    fmt.Fprintf(pErr, "\t-h      : print the command usage\n");
    return 1;       /* error exit */
}

/**Function*************************************************************

  Synopsis    [Sets simple LUT library.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Fpga_SetSimpleLutLib(nLutSize int32) {
    var s_LutLib10 Fpga_LutLib_t = Fpga_LutLib_t{ "lutlib",10, 0, {0,1,1,1,1,1,1,1,1,1,1}, {{0},{1},{1},{1},{1},{1},{1},{1},{1},{1},{1}} };
    var s_LutLib9 Fpga_LutLib_t  = Fpga_LutLib_t{ "lutlib", 9, 0, {0,1,1,1,1,1,1,1,1,1}, {{0},{1},{1},{1},{1},{1},{1},{1},{1},{1}} };
    var s_LutLib8 Fpga_LutLib_t  = Fpga_LutLib_t{ "lutlib", 8, 0, {0,1,1,1,1,1,1,1,1}, {{0},{1},{1},{1},{1},{1},{1},{1},{1}} };
    var s_LutLib7 Fpga_LutLib_t  = Fpga_LutLib_t{ "lutlib", 7, 0, {0,1,1,1,1,1,1,1}, {{0},{1},{1},{1},{1},{1},{1},{1}} };
    var s_LutLib6 Fpga_LutLib_t  = Fpga_LutLib_t{ "lutlib", 6, 0, {0,1,1,1,1,1,1}, {{0},{1},{1},{1},{1},{1},{1}} };
    var s_LutLib5 Fpga_LutLib_t  = Fpga_LutLib_t{ "lutlib", 5, 0, {0,1,1,1,1,1}, {{0},{1},{1},{1},{1},{1}} };
    var s_LutLib4 Fpga_LutLib_t  = Fpga_LutLib_t{ "lutlib", 4, 0, {0,1,1,1,1}, {{0},{1},{1},{1},{1}} };
    var s_LutLib3 Fpga_LutLib_t  = Fpga_LutLib_t{ "lutlib", 3, 0, {0,1,1,1}, {{0},{1},{1},{1}} };
    var pLutLib *Fpga_LutLib_t;
    assert(nLutSize >= 3 && nLutSize <= 10, "nLutSize >= 3 && nLutSize <= 10");
    switch  nLutSize  {
        case 3:  pLutLib = &s_LutLib3; break;
        case 4:  pLutLib = &s_LutLib4; break;
        case 5:  pLutLib = &s_LutLib5; break;
        case 6:  pLutLib = &s_LutLib6; break;
        case 7:  pLutLib = &s_LutLib7; break;
        case 8:  pLutLib = &s_LutLib8; break;
        case 9:  pLutLib = &s_LutLib9; break;
        case 10: pLutLib = &s_LutLib10; break;
        default: pLutLib = nil; break;
    }
    if pLutLib == nil {
        return; }
    Fpga_LutLibFree( (Fpga_LutLib_t *)Abc_FrameReadLibLut() );
    Abc_FrameSetLibLut( Fpga_LutLibDup(pLutLib) );
}

////////////////////////////////////////////////////////////////////////
///                       END OF FILE                                ///
////////////////////////////////////////////////////////////////////////


//ABC_NAMESPACE_IMPL_END

