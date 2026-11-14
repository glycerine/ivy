package bac

import (
	"fmt"
	"github.com/glycerine/abcgo/aig/gia"
	abcapi "github.com/glycerine/abcgo/base/abc"
	abcmain "github.com/glycerine/abcgo/base/main"
	"strconv"
	"strings"
)

import (
	"github.com/glycerine/abcgo/misc/vec"
	"os"
)

func boolStr(value bool) string {
	if value {
		return "yes"
	}
	return "no"
}

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

func Bac_AbcGetMan(pAbc *Abc_Frame_t) *Bac_Man_t {
	if m, ok := pAbc.Resource("bac.manager").(*Bac_Man_t); ok {
		return m
	}
	return nil
}
func Bac_AbcFreeMan(pAbc *Abc_Frame_t) {
	if m := Bac_AbcGetMan(pAbc); m != nil {
		Bac_ManFree(m)
	}
	pAbc.SetResource("bac.manager", nil)
}
func Bac_AbcUpdateMan(pAbc *Abc_Frame_t, p *Bac_Man_t) {
	Bac_AbcFreeMan(pAbc)
	pAbc.SetResource("bac.manager", p)
}

////////////////////////////////////////////////////////////////////////
///                     FUNCTION DEFINITIONS                         ///
////////////////////////////////////////////////////////////////////////

/*
*Function********************************************************************

	Synopsis    [Accessing current Bac_Ntk_t.]

	Description []

	SideEffects []

	SeeAlso     []

*****************************************************************************
*/
func Abc_FrameImportPtr(vPtr *vec.Vec_Ptr_t) {
	var p *Bac_Man_t
	if Abc_FrameGetGlobalFrame() == nil {
		fmt.Printf("ABC framework is not started.\n")
		return
	}
	p = Bac_PtrTransformToCba(vPtr)
	if p == nil {
		fmt.Printf("Converting from Ptr failed.\n")
	}
	Bac_AbcUpdateMan(Abc_FrameGetGlobalFrame(), p)
}
func Abc_FrameExportPtr() *vec.Vec_Ptr_t {
	var vPtr *vec.Vec_Ptr_t
	var p *Bac_Man_t
	if Abc_FrameGetGlobalFrame() == nil {
		fmt.Printf("ABC framework is not started.\n")
		return nil
	}
	p = Bac_AbcGetMan(Abc_FrameGetGlobalFrame())
	if p == nil {
		fmt.Printf("There is no CBA design present.\n")
	}
	vPtr = Bac_PtrDeriveFromCba(p)
	if vPtr == nil {
		fmt.Printf("Converting to Ptr has failed.\n")
	}
	return vPtr
}

/*
*Function********************************************************************

	Synopsis    []

	Description []

	SideEffects []

	SeeAlso     []

*****************************************************************************
*/
func Bac_Init(pAbc *Abc_Frame_t) {
	Cmd_CommandAdd(pAbc, "New word level", "@_read", Bac_CommandRead, 0)
	Cmd_CommandAdd(pAbc, "New word level", "@_write", Bac_CommandWrite, 0)
	Cmd_CommandAdd(pAbc, "New word level", "@_ps", Bac_CommandPs, 0)
	Cmd_CommandAdd(pAbc, "New word level", "@_put", Bac_CommandPut, 0)
	Cmd_CommandAdd(pAbc, "New word level", "@_get", Bac_CommandGet, 0)
	Cmd_CommandAdd(pAbc, "New word level", "@_clp", Bac_CommandClp, 0)
	Cmd_CommandAdd(pAbc, "New word level", "@_cec", Bac_CommandCec, 0)
	Cmd_CommandAdd(pAbc, "New word level", "@_test", Bac_CommandTest, 0)
}

/*
*Function********************************************************************

	Synopsis    []

	Description []

	SideEffects []

	SeeAlso     []

*****************************************************************************
*/
func Bac_End(pAbc *Abc_Frame_t) {
	Bac_AbcFreeMan(pAbc)
}

/*
*Function********************************************************************

	Synopsis    []

	Description []

	SideEffects []

	SeeAlso     []

*****************************************************************************
*/
func Bac_CommandRead(pAbc *Abc_Frame_t, argc int32, argv []string) int32 {
	var pFile *os.File
	var err error
	var p *Bac_Man_t = nil
	var vDes *vec.Vec_Ptr_t = nil
	var pFileName string
	var c, fUseAbc, fUsePtr, fVerbose int32
	Extra_UtilGetoptReset()
	for c = Extra_UtilGetopt(argc, argv, "apvh"); c != -1; c = Extra_UtilGetopt(argc, argv, "apvh") {
		switch c {
		case 'a':
			fUseAbc ^= 1
			break
		case 'p':
			fUsePtr ^= 1
			break
		case 'v':
			fVerbose ^= 1
			break
		case 'h':
			goto usage
		default:
			goto usage
		}
	}
	if argc != Extra_UtilGetoptIndex()+1 {
		fmt.Printf("Bac_CommandRead(): Input file name should be given on the command line.\n")
		return 0
	}
	// get the file name
	pFileName = argv[Extra_UtilGetoptIndex()]
	pFile, err = os.Open(pFileName)
	if err != nil {
		Abc_Print(1, "Cannot open input file \"%s\". ", pFileName)
		if similar := Extra_FileGetSimilarName(pFileName, ".v", ".blif", ".smt", ".bac", ""); similar != "" {
			pFileName = similar
			Abc_Print(1, "Did you mean \"%s\"?", pFileName)
		}
		Abc_Print(1, "\n")
		return 0
	}
	_ = (pFile).Close()
	// perform reading
	if fUseAbc != 0 || fUsePtr != 0 {
		var pAbcNtk *Abc_Ntk_t = Io_ReadNetlist(pFileName, Io_ReadFileType(pFileName), 0)
		var vDes *vec.Vec_Ptr_t = Ptr_AbcDeriveDes(pAbcNtk)
		p = Bac_PtrTransformToCba(vDes)
		Bac_PtrFree(vDes) // points to names in pAbcNtk
		if p != nil {
			p.pSpec = Abc_UtilStrsav(abcapi.Abc_NtkSpec(pAbcNtk))
		}
		Abc_NtkDelete(pAbcNtk)

	} else if !(int32(strings.Compare(Extra_FileNameExtension(pFileName), "blif")) != 0) {
		vDes = Psr_ManReadBlif(pFileName)
		if (vDes != nil) && vec.Vec_PtrSize(vDes) != 0 {
			p = Psr_ManBuildCba(pFileName, vDes)
		}
		if vDes != nil {
			Psr_ManVecFree(vDes)
		}

	} else if !(int32(strings.Compare(Extra_FileNameExtension(pFileName), "v")) != 0) {
		vDes = Psr_ManReadVerilog(pFileName)
		if (vDes != nil) && vec.Vec_PtrSize(vDes) != 0 {
			p = Psr_ManBuildCba(pFileName, vDes)
		}
		if vDes != nil {
			Psr_ManVecFree(vDes)
		}

	} else if !(int32(strings.Compare(Extra_FileNameExtension(pFileName), "smt")) != 0) {
		vDes = nil //Psr_ManReadSmt( pFileName );
		if (vDes != nil) && vec.Vec_PtrSize(vDes) != 0 {
			p = Psr_ManBuildCba(pFileName, vDes)
		}
		if vDes != nil {
			Psr_ManVecFree(vDes)
		}

	} else if !(int32(strings.Compare(Extra_FileNameExtension(pFileName), "bac")) != 0) {
		p = Bac_ManReadBac(pFileName)

	} else {
		fmt.Printf("Unrecognized input file extension.\n")
		return 0
	}
	Bac_AbcUpdateMan(pAbc, p)
	return 0
usage:
	Abc_Print(-2, "usage: @_read [-apvh] <file_name>\n")
	Abc_Print(-2, "\t         reads hierarchical design in BLIF or Verilog\n")
	Abc_Print(-2, "\t-a     : toggle using old ABC parser [default = %s]\n", boolStr(fUseAbc != 0))
	Abc_Print(-2, "\t-p     : toggle using Ptr construction [default = %s]\n", boolStr(fUsePtr != 0))
	Abc_Print(-2, "\t-v     : toggle printing verbose information [default = %s]\n", boolStr(fVerbose != 0))
	Abc_Print(-2, "\t-h     : print the command usage\n")
	return 1
}

/*
*Function********************************************************************

	Synopsis    []

	Description []

	SideEffects []

	SeeAlso     []

*****************************************************************************
*/
func Bac_CommandWrite(pAbc *Abc_Frame_t, argc int32, argv []string) int32 {
	var p *Bac_Man_t = Bac_AbcGetMan(pAbc)
	var pFileName string
	var fUseAssign int32 = 1
	var fUsePtr int32
	var c, fVerbose int32
	Extra_UtilGetoptReset()
	for c = Extra_UtilGetopt(argc, argv, "apvh"); c != -1; c = Extra_UtilGetopt(argc, argv, "apvh") {
		switch c {
		case 'a':
			fUseAssign ^= 1
			break
		case 'p':
			fUsePtr ^= 1
			break
		case 'v':
			fVerbose ^= 1
			break
		case 'h':
			goto usage
		default:
			goto usage
		}
	}
	if p == nil {
		Abc_Print(1, "Bac_CommandWrite(): There is no current design.\n")
		return 0
	}
	if argc == Extra_UtilGetoptIndex()+1 {
		pFileName = argv[Extra_UtilGetoptIndex()]
	} else if argc == Extra_UtilGetoptIndex() && (p != nil) {
		pFileName = Extra_FileNameGenericAppend(Bac_ManName(p), "_out.v")
	} else {
		fmt.Printf("Output file name should be given on the command line.\n")
		return 0
	}
	// perform writing
	if !(int32(strings.Compare(Extra_FileNameExtension(pFileName), "blif")) != 0) {
		Bac_ManWriteBlif(pFileName, p)
	} else if !(int32(strings.Compare(Extra_FileNameExtension(pFileName), "v")) != 0) {
		if fUsePtr != 0 {
			var vPtr *vec.Vec_Ptr_t = Bac_PtrDeriveFromCba(p)
			if vPtr == nil {
				fmt.Printf("Converting to Ptr has failed.\n")
			} else {
				Bac_PtrDumpVerilog(pFileName, vPtr)
				Bac_PtrFree(vPtr)
			}

		} else {
			Bac_ManWriteVerilog(pFileName, p, fUseAssign)
		}

	} else if !(int32(strings.Compare(Extra_FileNameExtension(pFileName), "bac")) != 0) {
		Bac_ManWriteBac(pFileName, p)
	} else {
		fmt.Printf("Unrecognized output file extension.\n")
		return 0
	}
	return 0
usage:
	Abc_Print(-2, "usage: @_write [-apvh]\n")
	Abc_Print(-2, "\t         writes the design into a file in BLIF or Verilog\n")
	Abc_Print(-2, "\t-a     : toggle using assign-statement for primitives [default = %s]\n", boolStr(fUseAssign != 0))
	Abc_Print(-2, "\t-p     : toggle using Ptr construction (mapped Verilog only) [default = %s]\n", boolStr(fUsePtr != 0))
	Abc_Print(-2, "\t-v     : toggle printing verbose information [default = %s]\n", boolStr(fVerbose != 0))
	Abc_Print(-2, "\t-h     : print the command usage\n")
	return 1
}

/*
*Function********************************************************************

	Synopsis    []

	Description []

	SideEffects []

	SeeAlso     []

*****************************************************************************
*/
func Bac_CommandPs(pAbc *Abc_Frame_t, argc int32, argv []string) int32 {
	optIndex := Extra_UtilGetoptIndex()
	var p *Bac_Man_t = Bac_AbcGetMan(pAbc)
	var c, nModules, fVerbose int32
	Extra_UtilGetoptReset()
	for c = Extra_UtilGetopt(argc, argv, "Mvh"); c != -1; c = Extra_UtilGetopt(argc, argv, "Mvh") {
		switch c {
		case 'M':
			if Extra_UtilGetoptIndex() >= argc {
				Abc_Print(-1, "Command line switch \"-M\" should be followed by an integer.\n")
				goto usage
			}
			parsed, err := strconv.Atoi(argv[Extra_UtilGetoptIndex()])
			if err != nil {
				goto usage
			}
			nModules = int32(parsed)
			optIndex = Extra_UtilGetoptIndex() + 1
			Extra_UtilGetoptSetIndex(optIndex)
			if nModules < 0 {
				goto usage
			}
			break
		case 'v':
			fVerbose ^= 1
			break
		case 'h':
			goto usage
		default:
			goto usage
		}
	}
	if p == nil {
		Abc_Print(1, "Bac_CommandPs(): There is no current design.\n")
		return 0
	}
	Bac_ManPrintStats(p, nModules, fVerbose)
	return 0
usage:
	Abc_Print(-2, "usage: @_ps [-M num] [-vh]\n")
	Abc_Print(-2, "\t         prints statistics\n")
	Abc_Print(-2, "\t-M num : the number of first modules to report [default = %d]\n", nModules)
	Abc_Print(-2, "\t-v     : toggle printing verbose information [default = %s]\n", boolStr(fVerbose != 0))
	Abc_Print(-2, "\t-h     : print the command usage\n")
	return 1
}

/*
*Function********************************************************************

	Synopsis    []

	Description []

	SideEffects []

	SeeAlso     []

*****************************************************************************
*/
func Bac_CommandPut(pAbc *Abc_Frame_t, argc int32, argv []string) int32 {
	var p *Bac_Man_t = Bac_AbcGetMan(pAbc)
	var pGia *Gia_Man_t = nil
	var c int32
	var fBarBufs int32 = 1
	var fVerbose int32
	Extra_UtilGetoptReset()
	for c = Extra_UtilGetopt(argc, argv, "bvh"); c != -1; c = Extra_UtilGetopt(argc, argv, "bvh") {
		switch c {
		case 'b':
			fBarBufs ^= 1
			break
		case 'v':
			fVerbose ^= 1
			break
		case 'h':
			goto usage
		default:
			goto usage
		}
	}
	if p == nil {
		Abc_Print(1, "Bac_CommandPut(): There is no current design.\n")
		return 0
	}
	pGia = Bac_ManExtract(p, fBarBufs, fVerbose)
	if pGia == nil {
		Abc_Print(1, "Bac_CommandPut(): Conversion to AIG has failed.\n")
		return 0
	}
	Abc_FrameUpdateGia(pAbc, pGia)
	return 0
usage:
	Abc_Print(-2, "usage: @_put [-bvh]\n")
	Abc_Print(-2, "\t         extracts AIG from the hierarchical design\n")
	Abc_Print(-2, "\t-b     : toggle using barrier buffers [default = %s]\n", boolStr(fBarBufs != 0))
	Abc_Print(-2, "\t-v     : toggle printing verbose information [default = %s]\n", boolStr(fVerbose != 0))
	Abc_Print(-2, "\t-h     : print the command usage\n")
	return 1
}

/*
*Function********************************************************************

	Synopsis    []

	Description []

	SideEffects []

	SeeAlso     []

*****************************************************************************
*/
func Bac_CommandGet(pAbc *Abc_Frame_t, argc int32, argv []string) int32 {
	var pNew *Bac_Man_t = nil
	var p *Bac_Man_t = Bac_AbcGetMan(pAbc)
	var c, fMapped, fVerbose int32
	Extra_UtilGetoptReset()
	for c = Extra_UtilGetopt(argc, argv, "mvh"); c != -1; c = Extra_UtilGetopt(argc, argv, "mvh") {
		switch c {
		case 'm':
			fMapped ^= 1
			break
		case 'v':
			fVerbose ^= 1
			break
		case 'h':
			goto usage
		default:
			goto usage
		}
	}
	if p == nil {
		Abc_Print(1, "Bac_CommandGet(): There is no current design.\n")
		return 0
	}
	if fMapped != 0 {
		pCurrent, _ := abcmain.Abc_FrameReadNtk(pAbc).(*abcapi.Network)
		if pCurrent == nil {
			Abc_Print(1, "Bac_CommandGet(): There is no current mapped design.\n")
			return 0
		}
		pNew = Bac_ManInsertAbc(p, pCurrent).(*Bac_Man_t)

	} else {
		pCurrent, _ := Abc_FrameReadGia(pAbc).(*Gia_Man_t)
		if pCurrent == nil {
			Abc_Print(1, "Bac_CommandGet(): There is no current AIG.\n")
			return 0
		}
		pNew = Bac_ManInsertGia(p, pCurrent)
	}
	Bac_AbcUpdateMan(pAbc, pNew)
	return 0
usage:
	Abc_Print(-2, "usage: @_get [-mvh]\n")
	Abc_Print(-2, "\t         inserts AIG or mapped network into the hierarchical design\n")
	Abc_Print(-2, "\t-m     : toggle using mapped network from main-space [default = %s]\n", boolStr(fMapped != 0))
	Abc_Print(-2, "\t-v     : toggle printing verbose information [default = %s]\n", boolStr(fVerbose != 0))
	Abc_Print(-2, "\t-h     : print the command usage\n")
	return 1
}

/*
*Function********************************************************************

	Synopsis    []

	Description []

	SideEffects []

	SeeAlso     []

*****************************************************************************
*/
func Bac_CommandClp(pAbc *Abc_Frame_t, argc int32, argv []string) int32 {
	var pNew *Bac_Man_t = nil
	var p *Bac_Man_t = Bac_AbcGetMan(pAbc)
	var c, fVerbose int32
	Extra_UtilGetoptReset()
	for c = Extra_UtilGetopt(argc, argv, "vh"); c != -1; c = Extra_UtilGetopt(argc, argv, "vh") {
		switch c {
		case 'v':
			fVerbose ^= 1
			break
		case 'h':
			goto usage
		default:
			goto usage
		}
	}
	if p == nil {
		Abc_Print(1, "Bac_CommandGet(): There is no current design.\n")
		return 0
	}
	pNew = Bac_ManCollapse(p)
	Bac_AbcUpdateMan(pAbc, pNew)
	return 0
usage:
	Abc_Print(-2, "usage: @_clp [-vh]\n")
	Abc_Print(-2, "\t         collapses the current hierarchical design\n")
	Abc_Print(-2, "\t-v     : toggle printing verbose information [default = %s]\n", boolStr(fVerbose != 0))
	Abc_Print(-2, "\t-h     : print the command usage\n")
	return 1
}

/*
*Function********************************************************************

	Synopsis    []

	Description []

	SideEffects []

	SeeAlso     []

*****************************************************************************
*/
func Bac_CommandCec(pAbc *Abc_Frame_t, argc int32, argv []string) int32 {
	var p *Bac_Man_t = Bac_AbcGetMan(pAbc)
	var pFirst, pSecond, pMiter *Gia_Man_t
	var ParsCec Cec_ParCec_t
	var pPars *Cec_ParCec_t = &ParsCec
	var vDes *vec.Vec_Ptr_t
	var FileName string
	var pArgvNew []string
	var c, nArgcNew, fDumpMiter int32
	var pFile *os.File
	var err error
	Cec_ManCecSetDefaultParams(pPars)
	Extra_UtilGetoptReset()
	for c = Extra_UtilGetopt(argc, argv, "vh"); c != -1; c = Extra_UtilGetopt(argc, argv, "vh") {
		switch c {
		case 'v':
			Cec_ParCecSetVerbose(pPars, Cec_ParCecVerbose(pPars)^1)
			break
		case 'h':
			goto usage
		default:
			goto usage
		}
	}
	if p == nil {
		Abc_Print(1, "Bac_CommandCec(): There is no current design.\n")
		return 0
	}
	pArgvNew = argv[Extra_UtilGetoptIndex():]
	nArgcNew = argc - Extra_UtilGetoptIndex()
	if nArgcNew != 1 {
		if p.pSpec == "" {
			Abc_Print(-1, "File name is not given on the command line.\n")
			return 1
		}
		FileName = p.pSpec

	} else {
		FileName = pArgvNew[0]
	}
	// fix the wrong symbol
	FileName = strings.ReplaceAll(FileName, ">", "\\")
	pFile, err = os.Open(FileName)
	if err != nil {
		Abc_Print(-1, "Cannot open input file \"%s\". ", FileName)
		if similar := Extra_FileGetSimilarName(FileName, ".v", ".blif", "", "", ""); similar != "" {
			FileName = similar
			Abc_Print(1, "Did you mean \"%s\"?", FileName)
		}
		Abc_Print(1, "\n")
		return 1
	}
	_ = (pFile).Close()

	// extract AIG from the current design
	pFirst = Bac_ManExtract(p, 0, 0)
	if pFirst == nil {
		Abc_Print(-1, "Extracting AIG from the current design has failed.\n")
		return 0
	}
	// extract AIG from the second design
	if !(int32(strings.Compare(Extra_FileNameExtension(FileName), "blif")) != 0) {
		vDes = Psr_ManReadBlif(FileName)
	} else if !(int32(strings.Compare(Extra_FileNameExtension(FileName), "v")) != 0) {
		vDes = Psr_ManReadVerilog(FileName)
	} else {
		assert(false, "0")
	}
	p = Psr_ManBuildCba(FileName, vDes)
	Psr_ManVecFree(vDes)
	pSecond = Bac_ManExtract(p, 0, 0)
	Bac_ManFree(p)
	if pSecond == nil {
		gia.Gia_ManStop(pFirst)
		Abc_Print(-1, "Extracting AIG from the original design has failed.\n")
		return 0
	}
	// compute the miter
	pMiter = gia.Gia_ManMiter(pFirst, pSecond, 0, 1, 0, 0, Cec_ParCecVerbose(pPars))
	if pMiter != nil {
		if fDumpMiter != 0 {
			Abc_Print(0, "The verification miter is written into file \"%s\".\n", "cec_miter.aig")
			gia.Gia_AigerWrite(pMiter, "cec_miter.aig", 0, 0, 0)
		}
		pAbc.Status = Cec_ManVerify(pMiter, pPars)
		//Abc_FrameReplaceCex( pAbc, &pAbc->pGia->pCexComb );
		gia.Gia_ManStop(pMiter)
	}
	gia.Gia_ManStop(pFirst)
	gia.Gia_ManStop(pSecond)
	return 0
usage:
	Abc_Print(-2, "usage: @_cec [-vh]\n")
	Abc_Print(-2, "\t         combinational equivalence checking\n")
	Abc_Print(-2, "\t-v     : toggle printing verbose information [default = %s]\n", boolStr(Cec_ParCecVerbose(pPars) != 0))
	Abc_Print(-2, "\t-h     : print the command usage\n")
	return 1
}

/*
*Function********************************************************************

	Synopsis    []

	Description []

	SideEffects []

	SeeAlso     []

*****************************************************************************
*/
func Bac_CommandTest(pAbc *Abc_Frame_t, argc int32, argv []string) int32 {
	//Bac_Man_t * p = Bac_AbcGetMan(pAbc);
	var c, fVerbose int32
	Extra_UtilGetoptReset()
	for c = Extra_UtilGetopt(argc, argv, "vh"); c != -1; c = Extra_UtilGetopt(argc, argv, "vh") {
		switch c {
		case 'v':
			fVerbose ^= 1
			break
		case 'h':
			goto usage
		default:
			goto usage
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
	return 0
usage:
	Abc_Print(-2, "usage: @_test [-vh]\n")
	Abc_Print(-2, "\t         experiments with word-level networks\n")
	Abc_Print(-2, "\t-v     : toggle printing verbose information [default = %s]\n", boolStr(fVerbose != 0))
	Abc_Print(-2, "\t-h     : print the command usage\n")
	return 1
}

////////////////////////////////////////////////////////////////////////
///                       END OF FILE                                ///
////////////////////////////////////////////////////////////////////////

//ABC_NAMESPACE_IMPL_END
