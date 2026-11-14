package cba

import (
	"fmt"
	"strconv"
	"strings"
)

import "os"

/**CFile****************************************************************

  FileName    [cbaCom.c]

  SystemName  [ABC: Logic synthesis and verification system.]

  PackageName [Hierarchical word-level netlist.]

  Synopsis    [Command handlers.]

  Author      [Alan Mishchenko]

  Affiliation [UC Berkeley]

  Date        [Ver. 1.0. Started - November 29, 2014.]

  Revision    [$Id: cbaCom.c,v 1.00 2014/11/29 00:00:00 alanmi Exp $]

***********************************************************************/

//#include "cba.h"
//#include "proof/cec/cec.h"
//#include "base/main/mainInt.h"

//ABC_NAMESPACE_IMPL_START

////////////////////////////////////////////////////////////////////////
///                        DECLARATIONS                              ///
////////////////////////////////////////////////////////////////////////

// static int  Cba_CommandRead     ( Abc_Frame_t * pAbc, int argc, char ** argv );

// static int  Cba_CommandWrite    ( Abc_Frame_t * pAbc, int argc, char ** argv );

// static int  Cba_CommandPs       ( Abc_Frame_t * pAbc, int argc, char ** argv );

// static int  Cba_CommandPut      ( Abc_Frame_t * pAbc, int argc, char ** argv );

// static int  Cba_CommandGet      ( Abc_Frame_t * pAbc, int argc, char ** argv );

// static int  Cba_CommandClp      ( Abc_Frame_t * pAbc, int argc, char ** argv );

// static int  Cba_CommandBlast    ( Abc_Frame_t * pAbc, int argc, char ** argv );

// static int  Cba_CommandCec      ( Abc_Frame_t * pAbc, int argc, char ** argv );

// static int  Cba_CommandTest     ( Abc_Frame_t * pAbc, int argc, char ** argv );

func Cba_AbcGetMan(pAbc *Abc_Frame_t) *Cba_Man_t {
	p, _ := pAbc.Resource("pAbcCba").(*Cba_Man_t)
	return p
}
func Cba_AbcFreeMan(pAbc *Abc_Frame_t) {
	if p := Cba_AbcGetMan(pAbc); p != nil {
		Cba_ManFree(p)
	}
	pAbc.SetResource("pAbcCba", nil)
}
func Cba_AbcUpdateMan(pAbc *Abc_Frame_t, p *Cba_Man_t) {
	Cba_AbcFreeMan(pAbc)
	pAbc.SetResource("pAbcCba", p)
}

////////////////////////////////////////////////////////////////////////
///                     FUNCTION DEFINITIONS                         ///
////////////////////////////////////////////////////////////////////////

/*
*Function********************************************************************

	Synopsis    []

	Description []

	SideEffects []

	SeeAlso     []

*****************************************************************************
*/
func yesNo(flag int32) string {
	if flag != 0 {
		return "yes"
	}
	return "no"
}

func Cba_Init(pAbc *Abc_Frame_t) {
	Cmd_CommandAdd(pAbc, "New word level", ":read", Cba_CommandRead, 0)
	Cmd_CommandAdd(pAbc, "New word level", ":write", Cba_CommandWrite, 0)
	Cmd_CommandAdd(pAbc, "New word level", ":ps", Cba_CommandPs, 0)
	Cmd_CommandAdd(pAbc, "New word level", ":put", Cba_CommandPut, 0)
	Cmd_CommandAdd(pAbc, "New word level", ":get", Cba_CommandGet, 0)
	Cmd_CommandAdd(pAbc, "New word level", ":clp", Cba_CommandClp, 0)
	Cmd_CommandAdd(pAbc, "New word level", ":blast", Cba_CommandBlast, 0)
	Cmd_CommandAdd(pAbc, "New word level", ":cec", Cba_CommandCec, 0)
	Cmd_CommandAdd(pAbc, "New word level", ":test", Cba_CommandTest, 0)
}

/*
*Function********************************************************************

	Synopsis    []

	Description []

	SideEffects []

	SeeAlso     []

*****************************************************************************
*/
func Cba_End(pAbc *Abc_Frame_t) {
	Cba_AbcFreeMan(pAbc)
}

/*
*Function********************************************************************

	Synopsis    []

	Description []

	SideEffects []

	SeeAlso     []

*****************************************************************************
*/
func Cba_CommandRead(pAbc *Abc_Frame_t, argc int32, argv []string) int32 {
	var pFile *os.File
	var err error
	var p *Cba_Man_t = nil
	var pFileName string
	var c, fTest, fDfs, fVerbose int32
	Extra_UtilGetoptReset()
	for c = Extra_UtilGetopt(argc, argv, "tdvh"); c != -1; c = Extra_UtilGetopt(argc, argv, "tdvh") {
		switch c {
		case 't':
			fTest ^= 1
			break
		case 'd':
			fDfs ^= 1
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
		fmt.Printf("Cba_CommandRead(): Input file name should be given on the command line.\n")
		return 0
	}
	// get the file name
	pFileName = argv[int(Extra_UtilGetoptIndex())]
	if pFile, err = os.Open(pFileName); err != nil {
		Abc_Print(1, "Cannot open input file \"%s\". ", pFileName)
		if similar := Extra_FileGetSimilarName(pFileName, ".v", ".blif", ".smt", ".cba", ""); similar != "" {
			Abc_Print(1, "Did you mean \"%s\"?", similar)
		}
		Abc_Print(1, "\n")
		return 0
	}
	_ = (pFile).Close()
	if fTest != 0 {
		if !(int32(strings.Compare(Extra_FileNameExtension(pFileName), "blif")) != 0) {
			Prs_ManReadBlifTest(pFileName)
		} else if !(int32(strings.Compare(Extra_FileNameExtension(pFileName), "v")) != 0) {
			Prs_ManReadVerilogTest(pFileName)
		} else {
			fmt.Printf("Unrecognized input file extension.\n")
			return 0
		}
		return 0
	}
	if !(int32(strings.Compare(Extra_FileNameExtension(pFileName), "blif")) != 0) {
		p = Cba_ManReadBlif(pFileName)
	} else if !(int32(strings.Compare(Extra_FileNameExtension(pFileName), "v")) != 0) {
		p = Cba_ManReadVerilog(pFileName)
	} else if !(int32(strings.Compare(Extra_FileNameExtension(pFileName), "cba")) != 0) {
		p = Cba_ManReadCba(pFileName)
	} else {
		fmt.Printf("Unrecognized input file extension.\n")
		return 0
	}
	if fDfs != 0 {
		var pTemp *Cba_Man_t
		pTemp = p
		p = Cba_ManDup(pTemp, Cba_NtkCollectDfs)
		Cba_ManFree(pTemp)
	}
	Cba_AbcUpdateMan(pAbc, p)
	return 0
usage:
	Abc_Print(-2, "usage: :read [-tdvh] <file_name>\n")
	Abc_Print(-2, "\t         reads hierarchical design\n")
	Abc_Print(-2, "\t-t     : toggle testing the parser [default = %s]\n", yesNo(fTest))
	Abc_Print(-2, "\t-d     : toggle computing DFS ordering [default = %s]\n", yesNo(fDfs))
	Abc_Print(-2, "\t-v     : toggle printing verbose information [default = %s]\n", yesNo(fVerbose))
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
func Cba_CommandWrite(pAbc *Abc_Frame_t, argc int32, argv []string) int32 {
	var p *Cba_Man_t = Cba_AbcGetMan(pAbc)
	var pFileName string
	var fInclineCats int32
	var c, fVerbose int32
	Extra_UtilGetoptReset()
	for c = Extra_UtilGetopt(argc, argv, "cvh"); c != -1; c = Extra_UtilGetopt(argc, argv, "cvh") {
		switch c {
		case 'c':
			fInclineCats ^= 1
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
		Abc_Print(1, "Cba_CommandWrite(): There is no current design.\n")
		return 0
	}

	if argc == Extra_UtilGetoptIndex()+1 {
		pFileName = argv[int(Extra_UtilGetoptIndex())]
	} else if argc == Extra_UtilGetoptIndex() && (p != nil) {
		pFileName = Extra_FileNameGenericAppend(func() string {
			if Cba_ManSpec(p) != "" {
				return Cba_ManSpec(p)
			}
			return Cba_ManName(p)
		}(), "_out.v")
		fmt.Printf("Generated output file name \"%s\".\n", pFileName)

	} else {
		fmt.Printf("Output file name should be given on the command line.\n")
		return 0
	}
	// perform writing
	if !(int32(strings.Compare(Extra_FileNameExtension(pFileName), "blif")) != 0) {
		Cba_ManWriteBlif(pFileName, p)
	} else if !(int32(strings.Compare(Extra_FileNameExtension(pFileName), "v")) != 0) {
		Cba_ManWriteVerilog(pFileName, p, fInclineCats)
	} else if !(int32(strings.Compare(Extra_FileNameExtension(pFileName), "cba")) != 0) {
		Cba_ManWriteCba(pFileName, p)
	} else {
		fmt.Printf("Unrecognized output file extension.\n")
		return 0
	}
	return 0
usage:
	Abc_Print(-2, "usage: :write [-cvh]\n")
	Abc_Print(-2, "\t         writes the design into a file in BLIF or Verilog\n")
	Abc_Print(-2, "\t-c     : toggle inlining input concatenations [default = %s]\n", yesNo(fInclineCats))
	Abc_Print(-2, "\t-v     : toggle printing verbose information [default = %s]\n", yesNo(fVerbose))
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
func Cba_CommandPs(pAbc *Abc_Frame_t, argc int32, argv []string) int32 {
	var p *Cba_Man_t = Cba_AbcGetMan(pAbc)
	var nModules int32
	var fShowMulti int32
	var fShowAdder int32
	var fDistrib int32
	var c, fVerbose int32
	Extra_UtilGetoptReset()
	for c = Extra_UtilGetopt(argc, argv, "Mmadvh"); c != -1; c = Extra_UtilGetopt(argc, argv, "Mmadvh") {
		switch c {
		case 'M':
			if Extra_UtilGetoptIndex() >= argc {
				Abc_Print(-1, "Command line switch \"-M\" should be followed by an integer.\n")
				goto usage
			}
			parsed, _ := strconv.Atoi(argv[int(Extra_UtilGetoptIndex())])
			nModules = int32(parsed)
			Extra_UtilGetoptSetIndex(Extra_UtilGetoptIndex() + 1)
			if nModules < 0 {
				goto usage
			}
			break
		case 'm':
			fShowMulti ^= 1
			break
		case 'a':
			fShowAdder ^= 1
			break
		case 'd':
			fDistrib ^= 1
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
		Abc_Print(1, "Cba_CommandPs(): There is no current design.\n")
		return 0
	}
	if nModules != 0 {
		Cba_ManPrintStats(p, nModules, fVerbose)
		return 0
	}
	Cba_NtkPrintStatsFull(Cba_ManRoot(p), fDistrib, fVerbose)
	if fShowMulti != 0 {
		Cba_NtkPrintNodes(Cba_ManRoot(p), int32(CBA_BOX_MUL))
	}
	if fShowAdder != 0 {
		Cba_NtkPrintNodes(Cba_ManRoot(p), int32(CBA_BOX_ADD))
	}
	return 0
usage:
	Abc_Print(-2, "usage: :ps [-M num] [-madvh]\n")
	Abc_Print(-2, "\t         prints statistics\n")
	Abc_Print(-2, "\t-M num : the number of first modules to report [default = %d]\n", nModules)
	Abc_Print(-2, "\t-m     : toggle printing multipliers [default = %s]\n", yesNo(fShowMulti))
	Abc_Print(-2, "\t-a     : toggle printing adders [default = %s]\n", yesNo(fShowAdder))
	Abc_Print(-2, "\t-d     : toggle printing distrubition [default = %s]\n", yesNo(fDistrib))
	Abc_Print(-2, "\t-v     : toggle printing verbose information [default = %s]\n", yesNo(fVerbose))
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
func Cba_CommandPut(pAbc *Abc_Frame_t, argc int32, argv []string) int32 {
	var p *Cba_Man_t = Cba_AbcGetMan(pAbc)
	var pGia *Gia_Man_t = nil
	var c int32
	var fBarBufs int32 = 1
	var fSeq int32
	var fVerbose int32
	Extra_UtilGetoptReset()
	for c = Extra_UtilGetopt(argc, argv, "bsvh"); c != -1; c = Extra_UtilGetopt(argc, argv, "bsvh") {
		switch c {
		case 'b':
			fBarBufs ^= 1
			break
		case 's':
			fSeq ^= 1
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
		Abc_Print(1, "Cba_CommandPut(): There is no current design.\n")
		return 0
	}
	pGia = Cba_ManBlast(p, fBarBufs, fSeq, fVerbose)
	if pGia == nil {
		Abc_Print(1, "Cba_CommandPut(): Conversion to AIG has failed.\n")
		return 0
	}
	Abc_FrameUpdateGia(pAbc, pGia)
	return 0
usage:
	Abc_Print(-2, "usage: :put [-bsvh]\n")
	Abc_Print(-2, "\t         extracts AIG from the hierarchical design\n")
	Abc_Print(-2, "\t-b     : toggle using barrier buffers [default = %s]\n", yesNo(fBarBufs))
	Abc_Print(-2, "\t-s     : toggle blasting sequential elements [default = %s]\n", yesNo(fSeq))
	Abc_Print(-2, "\t-v     : toggle printing verbose information [default = %s]\n", yesNo(fVerbose))
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
func Cba_CommandGet(pAbc *Abc_Frame_t, argc int32, argv []string) int32 {
	var pNew *Cba_Man_t = nil
	var p *Cba_Man_t = Cba_AbcGetMan(pAbc)
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
		Abc_Print(1, "Cba_CommandGet(): There is no current design.\n")
		return 0
	}

	if fMapped != 0 {
		pNtkCur := Abc_FrameReadNtk(pAbc)
		if pNtkCur == nil {
			Abc_Print(1, "Cba_CommandGet(): There is no current mapped design.\n")
			return 0
		}
		pNew = Cba_ManInsertAbc(p, pNtkCur)

	} else {
		pGiaCurrent, _ := Abc_FrameReadGia(pAbc).(*Gia_Man_t)
		if pGiaCurrent == nil {
			Abc_Print(1, "Cba_CommandGet(): There is no current AIG.\n")
			return 0
		}
		pNew = Cba_ManInsertGia(p, pGiaCurrent)
	}
	Cba_AbcUpdateMan(pAbc, pNew)
	return 0
usage:
	Abc_Print(-2, "usage: :get [-mvh]\n")
	Abc_Print(-2, "\t         extracts AIG or mapped network into the hierarchical design\n")
	Abc_Print(-2, "\t-m     : toggle using mapped network from main-space [default = %s]\n", yesNo(fMapped))
	Abc_Print(-2, "\t-v     : toggle printing verbose information [default = %s]\n", yesNo(fVerbose))
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
func Cba_CommandClp(pAbc *Abc_Frame_t, argc int32, argv []string) int32 {
	var pNew *Cba_Man_t = nil
	var p *Cba_Man_t = Cba_AbcGetMan(pAbc)
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
		Abc_Print(1, "Cba_CommandGet(): There is no current design.\n")
		return 0
	}
	pNew = Cba_ManCollapse(p)
	Cba_AbcUpdateMan(pAbc, pNew)
	return 0
usage:
	Abc_Print(-2, "usage: :clp [-vh]\n")
	Abc_Print(-2, "\t         collapses the current hierarchical design\n")
	Abc_Print(-2, "\t-v     : toggle printing verbose information [default = %s]\n", yesNo(fVerbose))
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
func Cba_CommandBlast(pAbc *Abc_Frame_t, argc int32, argv []string) int32 {
	var pNew *Gia_Man_t = nil
	var p *Cba_Man_t = Cba_AbcGetMan(pAbc)
	var c, fSeq, fVerbose int32
	Extra_UtilGetoptReset()
	for c = Extra_UtilGetopt(argc, argv, "svh"); c != -1; c = Extra_UtilGetopt(argc, argv, "svh") {
		switch c {
		case 's':
			fSeq ^= 1
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
		Abc_Print(1, "Cba_CommandBlast(): There is no current design.\n")
		return 0
	}
	pNew = Cba_ManBlast(p, 0, fSeq, fVerbose)
	if pNew == nil {
		Abc_Print(1, "Cba_CommandBlast(): Bit-blasting has failed.\n")
		return 0
	}
	Abc_FrameUpdateGia(pAbc, pNew)
	return 0
usage:
	Abc_Print(-2, "usage: :blast [-svh]\n")
	Abc_Print(-2, "\t         performs bit-blasting of the word-level design\n")
	Abc_Print(-2, "\t-s     : toggle blasting sequential elements [default = %s]\n", yesNo(fSeq))
	Abc_Print(-2, "\t-v     : toggle printing verbose information [default = %s]\n", yesNo(fVerbose))
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
func Cba_CommandCec(pAbc *Abc_Frame_t, argc int32, argv []string) int32 {
	var p *Cba_Man_t = Cba_AbcGetMan(pAbc)
	var pTemp *Cba_Man_t
	var pFirst, pSecond, pMiter *Gia_Man_t
	var ParsCec Cec_ParCec_t
	var pPars *Cec_ParCec_t = &ParsCec
	var pFileName string
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
		Abc_Print(1, "Cba_CommandCec(): There is no current design.\n")
		return 0
	}

	pArgvNew = argv[int(Extra_UtilGetoptIndex()):]
	nArgcNew = argc - Extra_UtilGetoptIndex()
	if nArgcNew != 1 {
		if p.pSpec == "" {
			Abc_Print(-1, "File name is not given on the command line.\n")
			return 1
		}
		pFileName = p.pSpec

	} else {
		pFileName = pArgvNew[0]
	}
	// fix the wrong symbol
	pFileName = strings.ReplaceAll(pFileName, ">", "\\")
	if pFile, err = os.Open(pFileName); err != nil {
		Abc_Print(-1, "Cannot open input file \"%s\". ", pFileName)
		if similar := Extra_FileGetSimilarName(pFileName, ".v", ".blif", "", "", ""); similar != "" {
			Abc_Print(1, "Did you mean \"%s\"?", similar)
		}
		Abc_Print(1, "\n")
		return 1
	}
	_ = (pFile).Close()

	// extract AIG from the current design
	pFirst = Cba_ManBlast(p, 0, 0, 0)
	if pFirst == nil {
		Abc_Print(-1, "Extracting AIG from the current design has failed.\n")
		return 0
	}
	// extract AIG from the second design

	if !(int32(strings.Compare(Extra_FileNameExtension(pFileName), "blif")) != 0) {
		pTemp = Cba_ManReadBlif(pFileName)
	} else if !(int32(strings.Compare(Extra_FileNameExtension(pFileName), "v")) != 0) {
		pTemp = Cba_ManReadVerilog(pFileName)
	} else if !(int32(strings.Compare(Extra_FileNameExtension(pFileName), "cba")) != 0) {
		pTemp = Cba_ManReadCba(pFileName)
	} else {
		assert(false, "0")
	}
	pSecond = Cba_ManBlast(pTemp, 0, 0, 0)
	Cba_ManFree(pTemp)
	if pSecond == nil {
		Gia_ManStop(pFirst)
		Abc_Print(-1, "Extracting AIG from the original design has failed.\n")
		return 0
	}
	// compute the miter
	pMiter = Gia_ManMiter(pFirst, pSecond, 0, 1, 0, 0, Cec_ParCecVerbose(pPars))
	if pMiter != nil {
		if fDumpMiter != 0 {
			Abc_Print(0, "The verification miter is written into file \"%s\".\n", "cec_miter.aig")
			_ = Gia_AigerWrite(pMiter, "cec_miter.aig", 0, 0, 0)
		}
		Abc_FrameSetStatus(Cec_ManVerify(pMiter, pPars))
		//Abc_FrameReplaceCex( pAbc, &pAbc->pGia->pCexComb );
		Gia_ManStop(pMiter)
	}
	Gia_ManStop(pFirst)
	Gia_ManStop(pSecond)
	return 0
usage:
	Abc_Print(-2, "usage: :cec [-vh]\n")
	Abc_Print(-2, "\t         combinational equivalence checking\n")
	Abc_Print(-2, "\t-v     : toggle printing verbose information [default = %s]\n", yesNo(Cec_ParCecVerbose(pPars)))
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
func Cba_CommandTest(pAbc *Abc_Frame_t, argc int32, argv []string) int32 {
	var p *Cba_Man_t = Cba_AbcGetMan(pAbc)
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
		Abc_Print(1, "Cba_CommandTest(): There is no current design.\n")
		return 0
	}
	return 0
usage:
	Abc_Print(-2, "usage: :test [-vh]\n")
	Abc_Print(-2, "\t         experiments with word-level networks\n")
	Abc_Print(-2, "\t-v     : toggle printing verbose information [default = %s]\n", yesNo(fVerbose))
	Abc_Print(-2, "\t-h     : print the command usage\n")
	return 1
}

////////////////////////////////////////////////////////////////////////
///                       END OF FILE                                ///
////////////////////////////////////////////////////////////////////////

//ABC_NAMESPACE_IMPL_END
