package bac

import (
	"fmt"
	"strconv"
	"strings"
)

import "os"

/**CFile****************************************************************

  FileName    [bacCom.c]

  SystemName  [ABC: Logic synthesis and verification system.]

  PackageName [Hierarchical word-level netlist.]

  Synopsis    [Command handlers.]

  Author      [Alan Mishchenko]
  
  Affiliation [UC Berkeley]

  Date        [Ver. 1.0. Started - November 29, 2014.]

  Revision    [$Id: bacCom.c,v 1.00 2014/11/29 00:00:00 alanmi Exp $]

***********************************************************************/

//#include "bac.h"
//#include "bacPrs.h"
//#include "proof/cec/cec.h"
//#include "base/main/mainInt.h"

//ABC_NAMESPACE_IMPL_START

////////////////////////////////////////////////////////////////////////
///                        DECLARATIONS                              ///
////////////////////////////////////////////////////////////////////////

// static int  Bac_CommandRead     ( Abc_Frame_t * pAbc, int argc, char ** argv );

// static int  Bac_CommandWrite    ( Abc_Frame_t * pAbc, int argc, char ** argv );

// static int  Bac_CommandPs       ( Abc_Frame_t * pAbc, int argc, char ** argv );

// static int  Bac_CommandPut      ( Abc_Frame_t * pAbc, int argc, char ** argv );

// static int  Bac_CommandGet      ( Abc_Frame_t * pAbc, int argc, char ** argv );

// static int  Bac_CommandClp      ( Abc_Frame_t * pAbc, int argc, char ** argv );

// static int  Bac_CommandCec      ( Abc_Frame_t * pAbc, int argc, char ** argv );

// static int  Bac_CommandTest     ( Abc_Frame_t * pAbc, int argc, char ** argv );


func Bac_AbcGetMan(pAbc *Abc_Frame_t) *Bac_Man_t { return (Bac_Man_t *)pAbc.pAbcBac;                        }
func Bac_AbcFreeMan(pAbc *Abc_Frame_t) { if pAbc.pAbcBac { Bac_ManFree(Bac_AbcGetMan(pAbc)); }    }
func Bac_AbcUpdateMan(pAbc *Abc_Frame_t, p *Bac_Man_t) { Bac_AbcFreeMan(pAbc); pAbc.pAbcBac = p;                  }

////////////////////////////////////////////////////////////////////////
///                     FUNCTION DEFINITIONS                         ///
////////////////////////////////////////////////////////////////////////

/**Function********************************************************************

  Synopsis    [Accessing current Bac_Ntk_t.]

  Description []

  SideEffects []

  SeeAlso     []

******************************************************************************/
func Abc_FrameImportPtr(vPtr *Vec_Ptr_t) {
    var p *Bac_Man_t;
    if Abc_FrameGetGlobalFrame() == nil {
        fmt.Printf("ABC framework is not started.\n");
        return;
    }
    p = Bac_PtrTransformToCba( vPtr );
    if p == nil {
        fmt.Printf("Converting from Ptr failed.\n"); }
    Bac_AbcUpdateMan( Abc_FrameGetGlobalFrame(), p );
}
func Abc_FrameExportPtr() *Vec_Ptr_t {
    var vPtr *Vec_Ptr_t;
    var p *Bac_Man_t;
    if Abc_FrameGetGlobalFrame() == nil {
        fmt.Printf("ABC framework is not started.\n");
        return nil;
    }
    p = Bac_AbcGetMan( Abc_FrameGetGlobalFrame() );
    if p == nil {
        fmt.Printf("There is no CBA design present.\n"); }
    vPtr = Bac_PtrDeriveFromCba( p );
    if vPtr == nil {
        fmt.Printf("Converting to Ptr has failed.\n"); }
    return vPtr;
}

/**Function********************************************************************

  Synopsis    []

  Description []

  SideEffects []

  SeeAlso     []

******************************************************************************/
func Bac_Init(pAbc *Abc_Frame_t) {
    Cmd_CommandAdd( pAbc, "New word level", "@_read",       Bac_CommandRead,      0 );
    Cmd_CommandAdd( pAbc, "New word level", "@_write",      Bac_CommandWrite,     0 );
    Cmd_CommandAdd( pAbc, "New word level", "@_ps",         Bac_CommandPs,        0 );
    Cmd_CommandAdd( pAbc, "New word level", "@_put",        Bac_CommandPut,       0 );
    Cmd_CommandAdd( pAbc, "New word level", "@_get",        Bac_CommandGet,       0 );
    Cmd_CommandAdd( pAbc, "New word level", "@_clp",        Bac_CommandClp,       0 );
    Cmd_CommandAdd( pAbc, "New word level", "@_cec",        Bac_CommandCec,       0 );
    Cmd_CommandAdd( pAbc, "New word level", "@_test",       Bac_CommandTest,      0 );
}

/**Function********************************************************************

  Synopsis    []

  Description []

  SideEffects []

  SeeAlso     []

******************************************************************************/
func Bac_End(pAbc *Abc_Frame_t) {
    Bac_AbcFreeMan( pAbc );
}


/**Function********************************************************************

  Synopsis    []

  Description []

  SideEffects []

  SeeAlso     []

******************************************************************************/
func Bac_CommandRead(pAbc *Abc_Frame_t, argc int32, argv *string) int32 {
    var pFile *os.File;
    var p *Bac_Man_t  = nil;
    var vDes *Vec_Ptr_t  = nil;
    var pFileName string  = nil;
    var c, fUseAbc, fUsePtr, fVerbose int32;
    Extra_UtilGetoptReset();
    for  ( c = Extra_UtilGetopt( argc, argv, "apvh" ) ) != -1  {
        switch  c  {
        case 'a':
            fUseAbc ^= 1;
            break;
        case 'p':
            fUsePtr ^= 1;
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
        fmt.Printf("Bac_CommandRead(): Input file name should be given on the command line.\n");
        return 0;
    }
    // get the file name
    pFileName = argv[globalUtilOptind];
    if (pFile = os.Open(pFileName)) == nil {
        Abc_Print( 1, "Cannot open input file \"%s\". ", pFileName );
        if pFileName = Extra_FileGetSimilarName( pFileName, ".v", ".blif", ".smt", ".bac", nil ) {
            Abc_Print( 1, "Did you mean \"%s\"?", pFileName ); }
        Abc_Print( 1, "\n" );
        return 0;
    }
    _ = (pFile).Close();
    // perform reading
    if fUseAbc || (fUsePtr != 0) {
        extern Vec_Ptr_t * Ptr_AbcDeriveDes( Abc_Ntk_t * pNtk );
        var pAbcNtk *Abc_Ntk_t  = Io_ReadNetlist( pFileName, Io_ReadFileType(pFileName), 0 );
        var vDes *Vec_Ptr_t  = Ptr_AbcDeriveDes( pAbcNtk );
        p = Bac_PtrTransformToCba( vDes );
        Bac_PtrFree( vDes ); // points to names in pAbcNtk
        if p != nil {
            ABC_FREE( p.pSpec );
            p.pSpec = Abc_UtilStrsav( pAbcNtk.pSpec );
        }
        Abc_NtkDelete( pAbcNtk );
    
    } else if !(int32(strings.Compare(Extra_FileNameExtension(pFileName), "blif")) != 0) {
        vDes = Psr_ManReadBlif( pFileName );
        if (vDes != nil) && Vec_PtrSize(vDes) {
            p = Psr_ManBuildCba( pFileName, vDes ); }
        if vDes != nil {
            Psr_ManVecFree( vDes ); }
    
    } else if !(int32(strings.Compare(Extra_FileNameExtension(pFileName), "v")) != 0) {
        vDes = Psr_ManReadVerilog( pFileName );
        if (vDes != nil) && Vec_PtrSize(vDes) {
            p = Psr_ManBuildCba( pFileName, vDes ); }
        if vDes != nil {
            Psr_ManVecFree( vDes ); }
    
    } else if !(int32(strings.Compare(Extra_FileNameExtension(pFileName), "smt")) != 0) {
        vDes = nil;//Psr_ManReadSmt( pFileName );
        if (vDes != nil) && Vec_PtrSize(vDes) {
            p = Psr_ManBuildCba( pFileName, vDes ); }
        if vDes != nil {
            Psr_ManVecFree( vDes ); }
    
    } else if !(int32(strings.Compare(Extra_FileNameExtension(pFileName), "bac")) != 0) {
        p = Bac_ManReadBac( pFileName );
    
    } else {
        fmt.Printf("Unrecognized input file extension.\n");
        return 0;
    }
    Bac_AbcUpdateMan( pAbc, p );
    return 0;
usage:
    Abc_Print( -2, "usage: @_read [-apvh] <file_name>\n" );
    Abc_Print( -2, "\t         reads hierarchical design in BLIF or Verilog\n" );
    Abc_Print( -2, "\t-a     : toggle using old ABC parser [default = %s]\n", fUseAbc? "yes": "no" );
    Abc_Print( -2, "\t-p     : toggle using Ptr construction [default = %s]\n", fUsePtr? "yes": "no" );
    Abc_Print( -2, "\t-v     : toggle printing verbose information [default = %s]\n", fVerbose? "yes": "no" );
    Abc_Print( -2, "\t-h     : print the command usage\n");
    return 1;
}

/**Function********************************************************************

  Synopsis    []

  Description []

  SideEffects []

  SeeAlso     []

******************************************************************************/
func Bac_CommandWrite(pAbc *Abc_Frame_t, argc int32, argv *string) int32 {
    var p *Bac_Man_t  = Bac_AbcGetMan(pAbc);
    var pFileName string  = nil;
    var fUseAssign int32    =    1;
    var fUsePtr int32; 
    var c, fVerbose int32;
    Extra_UtilGetoptReset();
    for  ( c = Extra_UtilGetopt( argc, argv, "apvh" ) ) != -1  {
        switch  c  {
        case 'a':
            fUseAssign ^= 1;
            break;
        case 'p':
            fUsePtr ^= 1;
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
    if p == nil {
        Abc_Print( 1, "Bac_CommandWrite(): There is no current design.\n" );
        return 0;
    }
    if argc == globalUtilOptind + 1 {
        pFileName = argv[globalUtilOptind];
    } else if argc == globalUtilOptind && (p != nil) {
        pFileName = Extra_FileNameGenericAppend( Bac_ManName(p), "_out.v" );
    } else {
        fmt.Printf("Output file name should be given on the command line.\n");
        return 0;
    }
    // perform writing
    if !(int32(strings.Compare(Extra_FileNameExtension(pFileName), "blif")) != 0) {
        Bac_ManWriteBlif( pFileName, p );
    } else if !(int32(strings.Compare(Extra_FileNameExtension(pFileName), "v")) != 0) {
        if fUsePtr != 0 {
            var vPtr *Vec_Ptr_t  = Bac_PtrDeriveFromCba( p );
            if vPtr == nil {
                fmt.Printf("Converting to Ptr has failed.\n");
            } else {
                Bac_PtrDumpVerilog( pFileName, vPtr );
                Bac_PtrFree( vPtr ); 
            }
        
        } else {
            Bac_ManWriteVerilog( pFileName, p, fUseAssign ); }        
    
    } else if !(int32(strings.Compare(Extra_FileNameExtension(pFileName), "bac")) != 0) {
        Bac_ManWriteBac( pFileName, p );
    } else {
        fmt.Printf("Unrecognized output file extension.\n");
        return 0;
    }
    return 0;
usage:
    Abc_Print( -2, "usage: @_write [-apvh]\n" );
    Abc_Print( -2, "\t         writes the design into a file in BLIF or Verilog\n" );
    Abc_Print( -2, "\t-a     : toggle using assign-statement for primitives [default = %s]\n",  fUseAssign? "yes": "no" );
    Abc_Print( -2, "\t-p     : toggle using Ptr construction (mapped Verilog only) [default = %s]\n", fUsePtr? "yes": "no" );
    Abc_Print( -2, "\t-v     : toggle printing verbose information [default = %s]\n",  fVerbose? "yes": "no" );
    Abc_Print( -2, "\t-h     : print the command usage\n");
    return 1;
}


/**Function********************************************************************

  Synopsis    []

  Description []

  SideEffects []

  SeeAlso     []

******************************************************************************/
func Bac_CommandPs(pAbc *Abc_Frame_t, argc int32, argv *string) int32 {
    var p *Bac_Man_t  = Bac_AbcGetMan(pAbc);
    var c, nModules, fVerbose int32;
    Extra_UtilGetoptReset();
    for  ( c = Extra_UtilGetopt( argc, argv, "Mvh" ) ) != -1  {
        switch  c  {
        case 'M':
            if globalUtilOptind >= argc {
                Abc_Print( -1, "Command line switch \"-M\" should be followed by an integer.\n" );
                goto usage;
            }
            nModules = strconv.Atoi(argv[globalUtilOptind]);
            globalUtilOptind++;
            if nModules < 0 {
                goto usage; }
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
    if p == nil {
        Abc_Print( 1, "Bac_CommandPs(): There is no current design.\n" );
        return 0;
    }
    Bac_ManPrintStats( p, nModules, fVerbose );
    return 0;
usage:
    Abc_Print( -2, "usage: @_ps [-M num] [-vh]\n" );
    Abc_Print( -2, "\t         prints statistics\n" );
    Abc_Print( -2, "\t-M num : the number of first modules to report [default = %d]\n", nModules );
    Abc_Print( -2, "\t-v     : toggle printing verbose information [default = %s]\n", fVerbose? "yes": "no" );
    Abc_Print( -2, "\t-h     : print the command usage\n");
    return 1;
}

/**Function********************************************************************

  Synopsis    []

  Description []

  SideEffects []

  SeeAlso     []

******************************************************************************/
func Bac_CommandPut(pAbc *Abc_Frame_t, argc int32, argv *string) int32 {
    var p *Bac_Man_t  = Bac_AbcGetMan(pAbc);
    var pGia *Gia_Man_t  = nil;
    var c int32;
    var fBarBufs int32 = 1;
    var fVerbose int32;
    Extra_UtilGetoptReset();
    for  ( c = Extra_UtilGetopt( argc, argv, "bvh" ) ) != -1  {
        switch  c  {
        case 'b':
            fBarBufs ^= 1;
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
    if p == nil {
        Abc_Print( 1, "Bac_CommandPut(): There is no current design.\n" );
        return 0;
    }
    pGia = Bac_ManExtract( p, fBarBufs, fVerbose );
    if pGia == nil {
        Abc_Print( 1, "Bac_CommandPut(): Conversion to AIG has failed.\n" );
        return 0;
    }
    Abc_FrameUpdateGia( pAbc, pGia );
    return 0;
usage:
    Abc_Print( -2, "usage: @_put [-bvh]\n" );
    Abc_Print( -2, "\t         extracts AIG from the hierarchical design\n" );
    Abc_Print( -2, "\t-b     : toggle using barrier buffers [default = %s]\n", fBarBufs? "yes": "no" );
    Abc_Print( -2, "\t-v     : toggle printing verbose information [default = %s]\n", fVerbose? "yes": "no" );
    Abc_Print( -2, "\t-h     : print the command usage\n");
    return 1;
}

/**Function********************************************************************

  Synopsis    []

  Description []

  SideEffects []

  SeeAlso     []

******************************************************************************/
func Bac_CommandGet(pAbc *Abc_Frame_t, argc int32, argv *string) int32 {
    var pNew *Bac_Man_t = nil;
    var p *Bac_Man_t = Bac_AbcGetMan(pAbc);
    var c, fMapped, fVerbose int32;
    Extra_UtilGetoptReset();
    for  ( c = Extra_UtilGetopt( argc, argv, "mvh" ) ) != -1  {
        switch  c  {
        case 'm':
            fMapped ^= 1;
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
    if p == nil {
        Abc_Print( 1, "Bac_CommandGet(): There is no current design.\n" );
        return 0;
    }
    if fMapped {
        if pAbc.pNtkCur == nil {
            Abc_Print( 1, "Bac_CommandGet(): There is no current mapped design.\n" );
            return 0;
        }
        pNew = (Bac_Man_t *)Bac_ManInsertAbc( p, pAbc.pNtkCur );
    
    } else {
        if pAbc.pGia == nil {
            Abc_Print( 1, "Bac_CommandGet(): There is no current AIG.\n" );
            return 0;
        }
        pNew = Bac_ManInsertGia( p, pAbc.pGia );
    }
    Bac_AbcUpdateMan( pAbc, pNew );
    return 0;
usage:
    Abc_Print( -2, "usage: @_get [-mvh]\n" );
    Abc_Print( -2, "\t         inserts AIG or mapped network into the hierarchical design\n" );
    Abc_Print( -2, "\t-m     : toggle using mapped network from main-space [default = %s]\n", fMapped? "yes": "no" );
    Abc_Print( -2, "\t-v     : toggle printing verbose information [default = %s]\n", fVerbose? "yes": "no" );
    Abc_Print( -2, "\t-h     : print the command usage\n");
    return 1;
}

/**Function********************************************************************

  Synopsis    []

  Description []

  SideEffects []

  SeeAlso     []

******************************************************************************/
func Bac_CommandClp(pAbc *Abc_Frame_t, argc int32, argv *string) int32 {
    var pNew *Bac_Man_t = nil;
    var p *Bac_Man_t = Bac_AbcGetMan(pAbc);
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
    if p == nil {
        Abc_Print( 1, "Bac_CommandGet(): There is no current design.\n" );
        return 0;
    }
    pNew = Bac_ManCollapse( p );
    Bac_AbcUpdateMan( pAbc, pNew );
    return 0;
usage:
    Abc_Print( -2, "usage: @_clp [-vh]\n" );
    Abc_Print( -2, "\t         collapses the current hierarchical design\n" );
    Abc_Print( -2, "\t-v     : toggle printing verbose information [default = %s]\n", fVerbose? "yes": "no" );
    Abc_Print( -2, "\t-h     : print the command usage\n");
    return 1;
}

/**Function********************************************************************

  Synopsis    []

  Description []

  SideEffects []

  SeeAlso     []

******************************************************************************/
func Bac_CommandCec(pAbc *Abc_Frame_t, argc int32, argv *string) int32 {
    var p *Bac_Man_t  = Bac_AbcGetMan(pAbc);
    var pFirst, pSecond, pMiter *Gia_Man_t;
    var ParsCec Cec_ParCec_t;
    var pPars *Cec_ParCec_t = &ParsCec;
    var vDes *Vec_Ptr_t;
    var FileName string;
    var pStr string;
    var pArgvNew *string;
    var c, nArgcNew, fDumpMiter int32;
    var pFile *os.File;
    Cec_ManCecSetDefaultParams( pPars );
    Extra_UtilGetoptReset();
    for  ( c = Extra_UtilGetopt( argc, argv, "vh" ) ) != -1  {
        switch  c  {
        case 'v':
            pPars.fVerbose ^= 1;
            break;
        case 'h':
            goto usage;
        default:
            goto usage;
        }
    }
    if p == nil {
        Abc_Print( 1, "Bac_CommandCec(): There is no current design.\n" );
        return 0;
    }
    pArgvNew = argv + globalUtilOptind;
    nArgcNew = argc - globalUtilOptind;
    if nArgcNew != 1 {
        if p.pSpec == nil {
            Abc_Print( -1, "File name is not given on the command line.\n" );
            return 1;
        }
        FileName = p.pSpec;
    
    } else {
        FileName = pArgvNew[0]; }
    // fix the wrong symbol
    for  pStr = FileName; *pStr; pStr++  {
        if *pStr == '>' {
            *pStr = '\\'; }
    }
    if (pFile = os.Open(FileName)) == nil {
        Abc_Print( -1, "Cannot open input file \"%s\". ", FileName );
        if FileName = Extra_FileGetSimilarName( FileName, ".v", ".blif", nil, nil, nil ) {
            Abc_Print( 1, "Did you mean \"%s\"?", FileName ); }
        Abc_Print( 1, "\n" );
        return 1;
    }
    _ = (pFile).Close();

    // extract AIG from the current design
    pFirst = Bac_ManExtract( p, 0, 0 );
    if pFirst == nil {
        Abc_Print( -1, "Extracting AIG from the current design has failed.\n" );
        return 0;
    }
    // extract AIG from the second design
    if !(int32(strings.Compare(Extra_FileNameExtension(FileName), "blif")) != 0) {
        vDes = Psr_ManReadBlif( FileName );
    } else if !(int32(strings.Compare(Extra_FileNameExtension(FileName), "v")) != 0) {
        vDes = Psr_ManReadVerilog( FileName );
    } else { assert(false, "0"); }
    p = Psr_ManBuildCba( FileName, vDes );
    Psr_ManVecFree( vDes );
    pSecond = Bac_ManExtract( p, 0, 0 );
    Bac_ManFree( p );
    if pSecond == nil {
        Gia_ManStop( pFirst );
        Abc_Print( -1, "Extracting AIG from the original design has failed.\n" );
        return 0;
    }
    // compute the miter
    pMiter = Gia_ManMiter( pFirst, pSecond, 0, 1, 0, 0, pPars.fVerbose );
    if pMiter != nil {
        if fDumpMiter != 0 {
            Abc_Print( 0, "The verification miter is written into file \"%s\".\n", "cec_miter.aig" );
            Gia_AigerWrite( pMiter, "cec_miter.aig", 0, 0, 0 );
        }
        pAbc.Status = Cec_ManVerify( pMiter, pPars );
        //Abc_FrameReplaceCex( pAbc, &pAbc->pGia->pCexComb );
        Gia_ManStop( pMiter );
    }
    Gia_ManStop( pFirst );
    Gia_ManStop( pSecond );
    return 0;
usage:
    Abc_Print( -2, "usage: @_cec [-vh]\n" );
    Abc_Print( -2, "\t         combinational equivalence checking\n" );
    Abc_Print( -2, "\t-v     : toggle printing verbose information [default = %s]\n", pPars.fVerbose? "yes": "no" );
    Abc_Print( -2, "\t-h     : print the command usage\n");
    return 1;
}

/**Function********************************************************************

  Synopsis    []

  Description []

  SideEffects []

  SeeAlso     []

******************************************************************************/
func Bac_CommandTest(pAbc *Abc_Frame_t, argc int32, argv *string) int32 {
    extern void Psr_ManReadBlifTest();
    extern void Psr_ManReadVerilogTest();
    extern void Psr_SmtReadSmtTest();
    //Bac_Man_t * p = Bac_AbcGetMan(pAbc);
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
/*
    if ( p == NULL )
    {
        Abc_Print( 1, "Bac_CommandTest(): There is no current design.\n" );
        return 0;
    }
*/
    //Bac_PtrTransformTestTest();
    //Psr_ManReadVerilogTest();
    //Psr_SmtReadSmtTest();
    return 0;
usage:
    Abc_Print( -2, "usage: @_test [-vh]\n" );
    Abc_Print( -2, "\t         experiments with word-level networks\n" );
    Abc_Print( -2, "\t-v     : toggle printing verbose information [default = %s]\n", fVerbose? "yes": "no" );
    Abc_Print( -2, "\t-h     : print the command usage\n");
    return 1;
}

////////////////////////////////////////////////////////////////////////
///                       END OF FILE                                ///
////////////////////////////////////////////////////////////////////////


//ABC_NAMESPACE_IMPL_END

