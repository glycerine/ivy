package wlc

import (
	"fmt"
	"github.com/glycerine/abcgo/aig/miniaig"
	wln "github.com/glycerine/abcgo/base/wln"
	"strconv"
	"strings"
)

import "os"

/**CFile****************************************************************

  FileName    [wlcCom.c]

  SystemName  [ABC: Logic synthesis and verification system.]

  PackageName [Verilog parser.]

  Synopsis    [Command handlers.]

  Author      [Alan Mishchenko]

  Affiliation [UC Berkeley]

  Date        [Ver. 1.0. Started - August 22, 2014.]

  Revision    [$Id: wlcCom.c,v 1.00 2014/09/12 00:00:00 alanmi Exp $]

***********************************************************************/

//#include "wlc.h"
//#include "base/wln/wln.h"
//#include "base/main/mainInt.h"
//#include "aig/miniaig/ndr.h"

//ABC_NAMESPACE_IMPL_START

////////////////////////////////////////////////////////////////////////
///                        DECLARATIONS                              ///
////////////////////////////////////////////////////////////////////////

// static int  Abc_CommandReadWlc    ( Abc_Frame_t * pAbc, int argc, char ** argv );

// static int  Abc_CommandWriteWlc   ( Abc_Frame_t * pAbc, int argc, char ** argv );

// static int  Abc_CommandPs         ( Abc_Frame_t * pAbc, int argc, char ** argv );

// static int  Abc_CommandCone       ( Abc_Frame_t * pAbc, int argc, char ** argv );

// static int  Abc_CommandAbs        ( Abc_Frame_t * pAbc, int argc, char ** argv );

// static int  Abc_CommandPdrAbs     ( Abc_Frame_t * pAbc, int argc, char ** argv );

// static int  Abc_CommandAbs2       ( Abc_Frame_t * pAbc, int argc, char ** argv );

// static int  Abc_CommandMemAbs     ( Abc_Frame_t * pAbc, int argc, char ** argv );

// static int  Abc_CommandMemAbs2    ( Abc_Frame_t * pAbc, int argc, char ** argv );

// static int  Abc_CommandBlast      ( Abc_Frame_t * pAbc, int argc, char ** argv );

// static int  Abc_CommandBlastMem   ( Abc_Frame_t * pAbc, int argc, char ** argv );

// static int  Abc_CommandGraft      ( Abc_Frame_t * pAbc, int argc, char ** argv );

// static int  Abc_CommandRetime     ( Abc_Frame_t * pAbc, int argc, char ** argv );

// static int  Abc_CommandProfile    ( Abc_Frame_t * pAbc, int argc, char ** argv );

// static int  Abc_CommandShortNames ( Abc_Frame_t * pAbc, int argc, char ** argv );

// static int  Abc_CommandShow       ( Abc_Frame_t * pAbc, int argc, char ** argv );

// static int  Abc_CommandInvPs      ( Abc_Frame_t * pAbc, int argc, char ** argv );

// static int  Abc_CommandInvPrint   ( Abc_Frame_t * pAbc, int argc, char ** argv );

// static int  Abc_CommandInvCheck   ( Abc_Frame_t * pAbc, int argc, char ** argv );

// static int  Abc_CommandInvGet     ( Abc_Frame_t * pAbc, int argc, char ** argv );

// static int  Abc_CommandInvPut     ( Abc_Frame_t * pAbc, int argc, char ** argv );

// static int  Abc_CommandInvMin     ( Abc_Frame_t * pAbc, int argc, char ** argv );

// static int  Abc_CommandTest       ( Abc_Frame_t * pAbc, int argc, char ** argv );

func wlcYesNo(v int32) string {
	if v != 0 {
		return "yes"
	}
	return "no"
}
func wlcYesNoBool(v bool) string {
	if v {
		return "yes"
	}
	return "no"
}
func wlcAtoi(s string) int32 {
	n, _ := strconv.Atoi(s)
	return int32(n)
}
func wlcSelect[T any](ok bool, a, b T) T {
	if ok {
		return a
	}
	return b
}

func Wlc_AbcGetNtk(pAbc *Abc_Frame_t) *Wlc_Ntk_t {
	n, _ := pAbc.Resource("pAbcWlc").(*Wlc_Ntk_t)
	return n
}
func Wlc_AbcFreeNtk(pAbc *Abc_Frame_t) {
	if Wlc_AbcGetNtk(pAbc) != nil {
		Wlc_NtkFree(Wlc_AbcGetNtk(pAbc))
	}
}
func Wlc_AbcUpdateNtk(pAbc *Abc_Frame_t, pNtk *Wlc_Ntk_t) {
	Wlc_AbcFreeNtk(pAbc)
	pAbc.SetResource("pAbcWlc", pNtk)
}

func Wlc_AbcGetInv(pAbc *Abc_Frame_t) *Vec_Int_t {
	values := pAbc.IntValues("pAbcWlcInv")
	if len(values) == 0 {
		return nil
	}
	v := Vec_IntAlloc(int32(len(values)))
	Vec_IntPushArray(v, values, int32(len(values)))
	return v
}
func Wlc_AbcSetInv(pAbc *Abc_Frame_t, inv *Vec_Int_t) {
	if inv == nil {
		pAbc.SetIntValues("pAbcWlcInv", nil)
		return
	}
	pAbc.SetIntValues("pAbcWlcInv", Vec_IntArray(inv))
}
func Wlc_AbcGetGia(pAbc *Abc_Frame_t) *Gia_Man_t {
	g, _ := pAbc.Resource("pGia").(*Gia_Man_t)
	return g
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
func Wlc_Init(pAbc *Abc_Frame_t) {
	Cmd_CommandAdd(pAbc, "Word level", "%read", Abc_CommandReadWlc, 0)
	Cmd_CommandAdd(pAbc, "Word level", "%write", Abc_CommandWriteWlc, 0)
	Cmd_CommandAdd(pAbc, "Word level", "%ps", Abc_CommandPs, 0)
	Cmd_CommandAdd(pAbc, "Word level", "%cone", Abc_CommandCone, 0)
	Cmd_CommandAdd(pAbc, "Word level", "%abs", Abc_CommandAbs, 0)
	Cmd_CommandAdd(pAbc, "Word level", "%pdra", Abc_CommandPdrAbs, 0)
	Cmd_CommandAdd(pAbc, "Word level", "%abs2", Abc_CommandAbs2, 0)
	Cmd_CommandAdd(pAbc, "Word level", "%memabs", Abc_CommandMemAbs, 0)
	Cmd_CommandAdd(pAbc, "Word level", "%memabs2", Abc_CommandMemAbs2, 0)
	Cmd_CommandAdd(pAbc, "Word level", "%blast", Abc_CommandBlast, 0)
	Cmd_CommandAdd(pAbc, "Word level", "%blastmem", Abc_CommandBlastMem, 0)
	Cmd_CommandAdd(pAbc, "Word level", "%graft", Abc_CommandGraft, 0)
	Cmd_CommandAdd(pAbc, "Word level", "%retime", Abc_CommandRetime, 0)
	Cmd_CommandAdd(pAbc, "Word level", "%profile", Abc_CommandProfile, 0)
	Cmd_CommandAdd(pAbc, "Word level", "%short_names", Abc_CommandShortNames, 0)
	Cmd_CommandAdd(pAbc, "Word level", "%show", Abc_CommandShow, 0)
	Cmd_CommandAdd(pAbc, "Word level", "%test", Abc_CommandTest, 0)

	Cmd_CommandAdd(pAbc, "Word level", "inv_ps", Abc_CommandInvPs, 0)
	Cmd_CommandAdd(pAbc, "Word level", "inv_print", Abc_CommandInvPrint, 0)
	Cmd_CommandAdd(pAbc, "Word level", "inv_check", Abc_CommandInvCheck, 0)
	Cmd_CommandAdd(pAbc, "Word level", "inv_get", Abc_CommandInvGet, 0)
	Cmd_CommandAdd(pAbc, "Word level", "inv_put", Abc_CommandInvPut, 0)
	Cmd_CommandAdd(pAbc, "Word level", "inv_min", Abc_CommandInvMin, 0)
}

/*
*Function********************************************************************

	Synopsis    []

	Description []

	SideEffects []

	SeeAlso     []

*****************************************************************************
*/
func Wlc_End(pAbc *Abc_Frame_t) {
	Wlc_AbcFreeNtk(pAbc)
}

/*
*Function********************************************************************

	Synopsis    []

	Description []

	SideEffects []

	SeeAlso     []

*****************************************************************************
*/
func Wlc_SetNtk(pAbc *Abc_Frame_t, pNtk *Wlc_Ntk_t) {
	Wlc_AbcUpdateNtk(pAbc, pNtk)
}

/*
*Function********************************************************************

	Synopsis    []

	Description []

	SideEffects []

	SeeAlso     []

*****************************************************************************
*/
func Abc_CommandReadWlc(pAbc *Abc_Frame_t, argc int32, argv []string) int32 {
	var pFile *os.File
	var pNtk *Wlc_Ntk_t = nil
	var pFileName string
	var fOldParser int32
	var fPrintTree int32
	var c, fVerbose int32
	Extra_UtilGetoptReset()
	for c = Extra_UtilGetopt(argc, argv, "opvh"); c != -1; c = Extra_UtilGetopt(argc, argv, "opvh") {
		switch c {
		case 'o':
			fOldParser ^= 1
			break
		case 'p':
			fPrintTree ^= 1
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
	if argc != globalUtilOptind+1 {
		fmt.Printf("Abc_CommandReadWlc(): Input file name should be given on the command line.\n")
		return 0
	}
	// get the file name
	pFileName = argv[globalUtilOptind]
	pFile, _ = os.Open(pFileName)
	if pFile == nil {
		Abc_Print(1, "Cannot open input file \"%s\". ", pFileName)
		if pFileName = Extra_FileGetSimilarName(pFileName, ".v", ".smt", ".smt2", ".ndr", ""); pFileName != "" {
			Abc_Print(1, "Did you mean \"%s\"?", pFileName)
		}
		Abc_Print(1, "\n")
		return 0
	}
	_ = (pFile).Close()

	// perform reading
	if !(int32(strings.Compare(Extra_FileNameExtension(pFileName), "v")) != 0) {
		pNtk = Wlc_ReadVer(pFileName, "")
	} else if !(int32(strings.Compare(Extra_FileNameExtension(pFileName), "smt")) != 0) || !(int32(strings.Compare(Extra_FileNameExtension(pFileName), "smt2")) != 0) {
		pNtk = Wlc_ReadSmt(pFileName, fOldParser, fPrintTree)
	} else if !(int32(strings.Compare(Extra_FileNameExtension(pFileName), "ndr")) != 0) {
		pNtk = Wlc_ReadNdr(pFileName)
	} else {
		fmt.Printf("Abc_CommandReadWlc(): Unknown file extension.\n")
		return 0
	}
	Wlc_AbcUpdateNtk(pAbc, pNtk)
	return 0
usage:
	Abc_Print(-2, "usage: %%read [-opvh] <file_name>\n")
	Abc_Print(-2, "\t         reads word-level design from Verilog file\n")
	Abc_Print(-2, "\t-o     : toggle using old SMT-LIB parser [default = %s]\n", wlcYesNo(fOldParser))
	Abc_Print(-2, "\t-p     : toggle printing parse SMT-LIB tree [default = %s]\n", wlcYesNo(fPrintTree))
	Abc_Print(-2, "\t-v     : toggle printing verbose information [default = %s]\n", wlcYesNo(fVerbose))
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
func Abc_CommandWriteWlc(pAbc *Abc_Frame_t, argc int32, argv []string) int32 {
	var pNtk *Wlc_Ntk_t = Wlc_AbcGetNtk(pAbc)
	var pFileName string
	var fAddCos int32
	var fSplitNodes int32
	var fNoFlops int32
	var c, fVerbose int32
	Extra_UtilGetoptReset()
	for c = Extra_UtilGetopt(argc, argv, "anfvh"); c != -1; c = Extra_UtilGetopt(argc, argv, "anfvh") {
		switch c {
		case 'a':
			fAddCos ^= 1
			break
		case 'n':
			fSplitNodes ^= 1
			break
		case 'f':
			fNoFlops ^= 1
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
	if pNtk == nil {
		Abc_Print(1, "Abc_CommandWriteWlc(): There is no current design.\n")
		return 0
	}
	if argc == globalUtilOptind {
		pFileName = Extra_FileNameGenericAppend(pNtk.pName, "_out.v")
	} else if argc == globalUtilOptind+1 {
		pFileName = argv[globalUtilOptind]
	} else {
		fmt.Printf("Output file name should be given on the command line.\n")
		return 0
	}
	if !(int32(strings.Compare(Extra_FileNameExtension(pFileName), "ndr")) != 0) {
		Wlc_WriteNdr(pNtk, pFileName)
	} else if fSplitNodes != 0 {
		pNtk = Wlc_NtkDupSingleNodes(pNtk)
		Wlc_WriteVer(pNtk, pFileName, fAddCos, fNoFlops)
		Wlc_NtkFree(pNtk)

	} else {
		Wlc_WriteVer(pNtk, pFileName, fAddCos, fNoFlops)
	}

	return 0
usage:
	Abc_Print(-2, "usage: %%write [-anfvh]\n")
	Abc_Print(-2, "\t         writes the design into a file\n")
	Abc_Print(-2, "\t-a     : toggle adding a CO for each node [default = %s]\n", wlcYesNo(fAddCos))
	Abc_Print(-2, "\t-n     : toggle splitting into individual nodes [default = %s]\n", wlcYesNo(fSplitNodes))
	Abc_Print(-2, "\t-f     : toggle skipping flops when writing file [default = %s]\n", wlcYesNo(fNoFlops))
	Abc_Print(-2, "\t-v     : toggle printing verbose information [default = %s]\n", wlcYesNo(fVerbose))
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
func Abc_CommandPs(pAbc *Abc_Frame_t, argc int32, argv []string) int32 {
	var pNtk *Wlc_Ntk_t = Wlc_AbcGetNtk(pAbc)
	var fShowCones int32
	var fShowMulti int32
	var fShowAdder int32
	var fShowMem int32
	var fDistrib int32
	var fTwoSides int32
	var fAllObjects int32
	var c, fVerbose int32
	Extra_UtilGetoptReset()
	for c = Extra_UtilGetopt(argc, argv, "cbamdtovh"); c != -1; c = Extra_UtilGetopt(argc, argv, "cbamdtovh") {
		switch c {
		case 'c':
			fShowCones ^= 1
			break
		case 'b':
			fShowMulti ^= 1
			break
		case 'a':
			fShowAdder ^= 1
			break
		case 'm':
			fShowMem ^= 1
			break
		case 'd':
			fDistrib ^= 1
			break
		case 't':
			fTwoSides ^= 1
			break
		case 'o':
			fAllObjects ^= 1
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
	if pNtk == nil {
		Abc_Print(1, "Abc_CommandPs(): There is no current design.\n")
		return 0
	}
	Wlc_NtkPrintStats(pNtk, fDistrib, fTwoSides, fVerbose)
	if fShowCones != 0 {
		Wlc_NtkProfileCones(pNtk)
	}
	if fShowMulti != 0 {
		Wlc_NtkPrintNodes(pNtk, int32(WLC_OBJ_ARI_MULTI))
	}
	if fShowAdder != 0 {
		Wlc_NtkPrintNodes(pNtk, int32(WLC_OBJ_ARI_ADD))
	}
	if fShowMem != 0 {
		Wlc_NtkPrintMemory(pNtk)
	}
	if fAllObjects != 0 {
		Wlc_NtkPrintObjects(pNtk)
	}
	return 0
usage:
	Abc_Print(-2, "usage: %%ps [-cbamdtovh]\n")
	Abc_Print(-2, "\t         prints statistics\n")
	Abc_Print(-2, "\t-c     : toggle printing cones [default = %s]\n", wlcYesNo(fShowCones))
	Abc_Print(-2, "\t-b     : toggle printing multipliers [default = %s]\n", wlcYesNo(fShowMulti))
	Abc_Print(-2, "\t-a     : toggle printing adders [default = %s]\n", wlcYesNo(fShowAdder))
	Abc_Print(-2, "\t-m     : toggle printing memories [default = %s]\n", wlcYesNo(fShowMem))
	Abc_Print(-2, "\t-d     : toggle printing distrubition [default = %s]\n", wlcYesNo(fDistrib))
	Abc_Print(-2, "\t-t     : toggle printing stats for LHS and RHS [default = %s]\n", wlcYesNo(fTwoSides))
	Abc_Print(-2, "\t-o     : toggle printing all objects [default = %s]\n", wlcYesNo(fAllObjects))
	Abc_Print(-2, "\t-v     : toggle printing verbose information [default = %s]\n", wlcYesNo(fVerbose))
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
func Abc_CommandCone(pAbc *Abc_Frame_t, argc int32, argv []string) int32 {
	var pNtk *Wlc_Ntk_t = Wlc_AbcGetNtk(pAbc)
	var c int32
	var iOutput int32 = -1
	var Range int32 = 1
	var fAllPis int32
	var fSeq int32
	var fVerbose int32
	var pName string
	Extra_UtilGetoptReset()
	for c = Extra_UtilGetopt(argc, argv, "ORisvh"); c != -1; c = Extra_UtilGetopt(argc, argv, "ORisvh") {
		switch c {
		case 'O':
			if globalUtilOptind >= argc {
				Abc_Print(-1, "Command line switch \"-O\" should be followed by an integer.\n")
				goto usage
			}
			iOutput = wlcAtoi(argv[globalUtilOptind])
			globalUtilOptind++
			if iOutput < 0 {
				goto usage
			}
			break
		case 'R':
			if globalUtilOptind >= argc {
				Abc_Print(-1, "Command line switch \"-R\" should be followed by an integer.\n")
				goto usage
			}
			Range = wlcAtoi(argv[globalUtilOptind])
			globalUtilOptind++
			if Range < 0 {
				goto usage
			}
			break
		case 'i':
			fAllPis ^= 1
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
	if pNtk == nil {
		Abc_Print(1, "Abc_CommandCone(): There is no current design.\n")
		return 0
	}
	if iOutput < 0 || iOutput >= Wlc_NtkCoNum(pNtk) {
		Abc_Print(1, "Abc_CommandCone(): Illegal output index (%d) (should be 0 <= num < %d).\n", iOutput, Wlc_NtkCoNum(pNtk))
		return 0
	}
	fmt.Printf("Extracting output %d as a %s word-level network.\n", iOutput, wlcSelect(fSeq != 0, "sequential", "combinational"))
	pName = Wlc_NtkNewName(pNtk, iOutput, fSeq)
	Wlc_NtkMarkCone(pNtk, iOutput, Range, fSeq, fAllPis)
	pNtk = Wlc_NtkDupDfs(pNtk, 1, fSeq)
	pNtk.pName = ""
	pNtk.pName = Abc_UtilStrsav(pName)
	Wlc_AbcUpdateNtk(pAbc, pNtk)
	return 0
usage:
	Abc_Print(-2, "usage: %%cone [-OR num] [-isvh]\n")
	Abc_Print(-2, "\t         extracts logic cone of one or more word-level outputs\n")
	Abc_Print(-2, "\t-O num : zero-based index of the first word-level output to extract [default = %d]\n", iOutput)
	Abc_Print(-2, "\t-R num : total number of word-level outputs to extract [default = %d]\n", Range)
	Abc_Print(-2, "\t-i     : toggle using support composed of all primary inputs [default = %s]\n", wlcYesNo(fAllPis))
	Abc_Print(-2, "\t-s     : toggle performing extracting sequential cones [default = %s]\n", wlcYesNo(fSeq))
	Abc_Print(-2, "\t-v     : toggle printing verbose information [default = %s]\n", wlcYesNo(fVerbose))
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
func Abc_CommandPdrAbs(pAbc *Abc_Frame_t, argc int32, argv []string) int32 {
	var pNtk *Wlc_Ntk_t = Wlc_AbcGetNtk(pAbc)
	var Pars Wlc_Par_t
	var pPars *Wlc_Par_t = &Pars
	var c int32
	Wlc_ManSetDefaultParams(pPars)
	Extra_UtilGetoptReset()
	for c = Extra_UtilGetopt(argc, argv, "AMXFILabrcdilpqmstuxvwh"); c != -1; c = Extra_UtilGetopt(argc, argv, "AMXFILabrcdilpqmstuxvwh") {
		switch c {
		case 'A':
			if globalUtilOptind >= argc {
				Abc_Print(-1, "Command line switch \"-A\" should be followed by an integer.\n")
				goto usage
			}
			pPars.nBitsAdd = wlcAtoi(argv[globalUtilOptind])
			globalUtilOptind++
			if pPars.nBitsAdd < 0 {
				goto usage
			}
			break
		case 'M':
			if globalUtilOptind >= argc {
				Abc_Print(-1, "Command line switch \"-M\" should be followed by an integer.\n")
				goto usage
			}
			pPars.nBitsMul = wlcAtoi(argv[globalUtilOptind])
			globalUtilOptind++
			if pPars.nBitsMul < 0 {
				goto usage
			}
			break
		case 'X':
			if globalUtilOptind >= argc {
				Abc_Print(-1, "Command line switch \"-X\" should be followed by an integer.\n")
				goto usage
			}
			pPars.nBitsMux = wlcAtoi(argv[globalUtilOptind])
			globalUtilOptind++
			if pPars.nBitsMux < 0 {
				goto usage
			}
			break
		case 'F':
			if globalUtilOptind >= argc {
				Abc_Print(-1, "Command line switch \"-F\" should be followed by an integer.\n")
				goto usage
			}
			pPars.nBitsFlop = wlcAtoi(argv[globalUtilOptind])
			globalUtilOptind++
			if pPars.nBitsFlop < 0 {
				goto usage
			}
			break
		case 'I':
			if globalUtilOptind >= argc {
				Abc_Print(-1, "Command line switch \"-I\" should be followed by an integer.\n")
				goto usage
			}
			pPars.nIterMax = wlcAtoi(argv[globalUtilOptind])
			globalUtilOptind++
			if pPars.nIterMax < 0 {
				goto usage
			}
			break
		case 'L':
			if globalUtilOptind >= argc {
				Abc_Print(-1, "Command line switch \"-L\" should be followed by an integer.\n")
				goto usage
			}
			pPars.nLimit = wlcAtoi(argv[globalUtilOptind])
			globalUtilOptind++
			if pPars.nLimit < 0 {
				goto usage
			}
			break
		case 'a':
			pPars.fPdra ^= 1
			break
		case 'b':
			pPars.fProofRefine ^= 1
			break
		case 'r':
			pPars.fHybrid ^= 1
			break
		case 'x':
			pPars.fXorOutput ^= 1
			break
		case 'c':
			pPars.fCheckClauses ^= 1
			break
		case 'd':
			pPars.fAbs2 ^= 1
			break
		case 'i':
			pPars.fProofUsePPI ^= 1
			break
		case 'l':
			pPars.fLoadTrace ^= 1
			break
		case 'p':
			pPars.fPushClauses ^= 1
			break
		case 'q':
			pPars.fUseBmc3 ^= 1
			break
		case 'm':
			pPars.fMFFC ^= 1
			break
		case 's':
			pPars.fShrinkAbs ^= 1
			break
		case 't':
			pPars.fShrinkScratch ^= 1
			break
		case 'u':
			pPars.fCheckCombUnsat ^= 1
			break
		case 'v':
			pPars.fVerbose ^= 1
			break
		case 'w':
			pPars.fPdrVerbose ^= 1
			break
		case 'h':
			goto usage
		default:
			goto usage
		}
	}
	if pNtk == nil {
		Abc_Print(1, "Abc_CommandCone(): There is no current design.\n")
		return 0
	}
	Wlc_NtkPdrAbs(pNtk, pPars)
	return 0
usage:
	Abc_Print(-2, "usage: %%pdra [-AMXFIL num] [-abrcdilpqmxstuvwh]\n")
	Abc_Print(-2, "\t         abstraction for word-level networks\n")
	Abc_Print(-2, "\t-A num : minimum bit-width of an adder/subtractor to abstract [default = %d]\n", pPars.nBitsAdd)
	Abc_Print(-2, "\t-M num : minimum bit-width of a multiplier to abstract [default = %d]\n", pPars.nBitsMul)
	Abc_Print(-2, "\t-X num : minimum bit-width of a MUX operator to abstract [default = %d]\n", pPars.nBitsMux)
	Abc_Print(-2, "\t-F num : minimum bit-width of a flip-flop to abstract [default = %d]\n", pPars.nBitsFlop)
	Abc_Print(-2, "\t-I num : maximum number of CEGAR iterations [default = %d]\n", pPars.nIterMax)
	Abc_Print(-2, "\t-L num : maximum number of each type of signals [default = %d]\n", pPars.nLimit)
	Abc_Print(-2, "\t-x     : toggle XORing outputs of word-level miter [default = %s]\n", wlcYesNo(pPars.fXorOutput))
	Abc_Print(-2, "\t-a     : toggle running pdr with -nct [default = %s]\n", wlcYesNo(pPars.fPdra))
	Abc_Print(-2, "\t-b     : toggle using proof-based refinement [default = %s]\n", wlcYesNo(pPars.fProofRefine))
	Abc_Print(-2, "\t-r     : toggle using both cex-based and proof-based refinement [default = %s]\n", wlcYesNo(pPars.fHybrid))
	Abc_Print(-2, "\t-c     : toggle checking clauses in the reloaded trace [default = %s]\n", wlcYesNo(pPars.fCheckClauses))
	Abc_Print(-2, "\t-d     : toggle using another way of creating abstractions [default = %s]\n", wlcYesNo(pPars.fAbs2))
	Abc_Print(-2, "\t-i     : toggle using PPI values in proof-based refinement [default = %s]\n", wlcYesNo(pPars.fProofUsePPI))
	Abc_Print(-2, "\t-l     : toggle loading previous PDR traces [default = %s]\n", wlcYesNo(pPars.fLoadTrace))
	Abc_Print(-2, "\t-s     : toggle shrinking abstractions with BMC [default = %s]\n", wlcYesNo(pPars.fShrinkAbs))
	Abc_Print(-2, "\t-t     : toggle restarting pdr from scratch after shrinking abstractions with BMC [default = %s]\n", wlcYesNo(pPars.fShrinkScratch))
	Abc_Print(-2, "\t-u     : toggle checking combinationally unsat [default = %s]\n", wlcYesNo(pPars.fCheckCombUnsat))
	Abc_Print(-2, "\t-p     : toggle pushing clauses in the reloaded trace [default = %s]\n", wlcYesNo(pPars.fPushClauses))
	Abc_Print(-2, "\t-q     : toggle running bmc3 in parallel for CEX [default = %s]\n", wlcYesNo(pPars.fUseBmc3))
	Abc_Print(-2, "\t-m     : toggle refining the whole MFFC of a PPI [default = %s]\n", wlcYesNo(pPars.fMFFC))
	Abc_Print(-2, "\t-v     : toggle printing verbose information [default = %s]\n", wlcYesNo(pPars.fVerbose))
	Abc_Print(-2, "\t-w     : toggle printing verbose PDR output [default = %s]\n", wlcYesNo(pPars.fPdrVerbose))
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
func Abc_CommandAbs(pAbc *Abc_Frame_t, argc int32, argv []string) int32 {
	var pNtk *Wlc_Ntk_t = Wlc_AbcGetNtk(pAbc)
	var Pars Wlc_Par_t
	var pPars *Wlc_Par_t = &Pars
	var c int32
	Wlc_ManSetDefaultParams(pPars)
	Extra_UtilGetoptReset()
	for c = Extra_UtilGetopt(argc, argv, "AMXFILdxvwh"); c != -1; c = Extra_UtilGetopt(argc, argv, "AMXFILdxvwh") {
		switch c {
		case 'A':
			if globalUtilOptind >= argc {
				Abc_Print(-1, "Command line switch \"-A\" should be followed by an integer.\n")
				goto usage
			}
			pPars.nBitsAdd = wlcAtoi(argv[globalUtilOptind])
			globalUtilOptind++
			if pPars.nBitsAdd < 0 {
				goto usage
			}
			break
		case 'M':
			if globalUtilOptind >= argc {
				Abc_Print(-1, "Command line switch \"-M\" should be followed by an integer.\n")
				goto usage
			}
			pPars.nBitsMul = wlcAtoi(argv[globalUtilOptind])
			globalUtilOptind++
			if pPars.nBitsMul < 0 {
				goto usage
			}
			break
		case 'X':
			if globalUtilOptind >= argc {
				Abc_Print(-1, "Command line switch \"-X\" should be followed by an integer.\n")
				goto usage
			}
			pPars.nBitsMux = wlcAtoi(argv[globalUtilOptind])
			globalUtilOptind++
			if pPars.nBitsMux < 0 {
				goto usage
			}
			break
		case 'F':
			if globalUtilOptind >= argc {
				Abc_Print(-1, "Command line switch \"-F\" should be followed by an integer.\n")
				goto usage
			}
			pPars.nBitsFlop = wlcAtoi(argv[globalUtilOptind])
			globalUtilOptind++
			if pPars.nBitsFlop < 0 {
				goto usage
			}
			break
		case 'I':
			if globalUtilOptind >= argc {
				Abc_Print(-1, "Command line switch \"-I\" should be followed by an integer.\n")
				goto usage
			}
			pPars.nIterMax = wlcAtoi(argv[globalUtilOptind])
			globalUtilOptind++
			if pPars.nIterMax < 0 {
				goto usage
			}
			break
		case 'L':
			if globalUtilOptind >= argc {
				Abc_Print(-1, "Command line switch \"-L\" should be followed by an integer.\n")
				goto usage
			}
			pPars.nLimit = wlcAtoi(argv[globalUtilOptind])
			globalUtilOptind++
			if pPars.nLimit < 0 {
				goto usage
			}
			break
		case 'd':
			pPars.fAbs2 ^= 1
			break
		case 'x':
			pPars.fXorOutput ^= 1
			break
		case 'v':
			pPars.fVerbose ^= 1
			break
		case 'w':
			pPars.fPdrVerbose ^= 1
			break
		case 'h':
			goto usage
		default:
			goto usage
		}
	}
	if pNtk == nil {
		Abc_Print(1, "Abc_CommandCone(): There is no current design.\n")
		return 0
	}
	Wlc_NtkAbsCore(pNtk, pPars)
	return 0
usage:
	Abc_Print(-2, "usage: %%abs [-AMXFIL num] [-dxvwh]\n")
	Abc_Print(-2, "\t         abstraction for word-level networks\n")
	Abc_Print(-2, "\t-A num : minimum bit-width of an adder/subtractor to abstract [default = %d]\n", pPars.nBitsAdd)
	Abc_Print(-2, "\t-M num : minimum bit-width of a multiplier to abstract [default = %d]\n", pPars.nBitsMul)
	Abc_Print(-2, "\t-X num : minimum bit-width of a MUX operator to abstract [default = %d]\n", pPars.nBitsMux)
	Abc_Print(-2, "\t-F num : minimum bit-width of a flip-flop to abstract [default = %d]\n", pPars.nBitsFlop)
	Abc_Print(-2, "\t-I num : maximum number of CEGAR iterations [default = %d]\n", pPars.nIterMax)
	Abc_Print(-2, "\t-L num : maximum number of each type of signals [default = %d]\n", pPars.nLimit)
	Abc_Print(-2, "\t-d     : toggle using another way of creating abstractions [default = %s]\n", wlcYesNo(pPars.fAbs2))
	Abc_Print(-2, "\t-x     : toggle XORing outputs of word-level miter [default = %s]\n", wlcYesNo(pPars.fXorOutput))
	Abc_Print(-2, "\t-v     : toggle printing verbose information [default = %s]\n", wlcYesNo(pPars.fVerbose))
	Abc_Print(-2, "\t-w     : toggle printing verbose PDR output [default = %s]\n", wlcYesNo(pPars.fPdrVerbose))
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
func Abc_CommandAbs2(pAbc *Abc_Frame_t, argc int32, argv []string) int32 {
	var pNtk *Wlc_Ntk_t = Wlc_AbcGetNtk(pAbc)
	var Pars Wlc_Par_t
	var pPars *Wlc_Par_t = &Pars
	var c int32
	Wlc_ManSetDefaultParams(pPars)
	Extra_UtilGetoptReset()
	for c = Extra_UtilGetopt(argc, argv, "AMXFIxvwh"); c != -1; c = Extra_UtilGetopt(argc, argv, "AMXFIxvwh") {
		switch c {
		case 'A':
			if globalUtilOptind >= argc {
				Abc_Print(-1, "Command line switch \"-A\" should be followed by an integer.\n")
				goto usage
			}
			pPars.nBitsAdd = wlcAtoi(argv[globalUtilOptind])
			globalUtilOptind++
			if pPars.nBitsAdd < 0 {
				goto usage
			}
			break
		case 'M':
			if globalUtilOptind >= argc {
				Abc_Print(-1, "Command line switch \"-M\" should be followed by an integer.\n")
				goto usage
			}
			pPars.nBitsMul = wlcAtoi(argv[globalUtilOptind])
			globalUtilOptind++
			if pPars.nBitsMul < 0 {
				goto usage
			}
			break
		case 'X':
			if globalUtilOptind >= argc {
				Abc_Print(-1, "Command line switch \"-X\" should be followed by an integer.\n")
				goto usage
			}
			pPars.nBitsMux = wlcAtoi(argv[globalUtilOptind])
			globalUtilOptind++
			if pPars.nBitsMux < 0 {
				goto usage
			}
			break
		case 'F':
			if globalUtilOptind >= argc {
				Abc_Print(-1, "Command line switch \"-F\" should be followed by an integer.\n")
				goto usage
			}
			pPars.nBitsFlop = wlcAtoi(argv[globalUtilOptind])
			globalUtilOptind++
			if pPars.nBitsFlop < 0 {
				goto usage
			}
			break
		case 'I':
			if globalUtilOptind >= argc {
				Abc_Print(-1, "Command line switch \"-I\" should be followed by an integer.\n")
				goto usage
			}
			pPars.nIterMax = wlcAtoi(argv[globalUtilOptind])
			globalUtilOptind++
			if pPars.nIterMax < 0 {
				goto usage
			}
			break
		case 'x':
			pPars.fXorOutput ^= 1
			break
		case 'v':
			pPars.fVerbose ^= 1
			break
		case 'w':
			pPars.fPdrVerbose ^= 1
			break
		case 'h':
			goto usage
		default:
			goto usage
		}
	}
	if pNtk == nil {
		Abc_Print(1, "Abc_CommandCone(): There is no current design.\n")
		return 0
	}
	Wlc_NtkAbsCore2(pNtk, pPars)
	return 0
usage:
	Abc_Print(-2, "usage: %%abs2 [-AMXFI num] [-xvwh]\n")
	Abc_Print(-2, "\t         abstraction for word-level networks\n")
	Abc_Print(-2, "\t-A num : minimum bit-width of an adder/subtractor to abstract [default = %d]\n", pPars.nBitsAdd)
	Abc_Print(-2, "\t-M num : minimum bit-width of a multiplier to abstract [default = %d]\n", pPars.nBitsMul)
	Abc_Print(-2, "\t-X num : minimum bit-width of a MUX operator to abstract [default = %d]\n", pPars.nBitsMux)
	Abc_Print(-2, "\t-F num : minimum bit-width of a flip-flop to abstract [default = %d]\n", pPars.nBitsFlop)
	Abc_Print(-2, "\t-I num : maximum number of CEGAR iterations [default = %d]\n", pPars.nIterMax)
	Abc_Print(-2, "\t-x     : toggle XORing outputs of word-level miter [default = %s]\n", wlcYesNo(pPars.fXorOutput))
	Abc_Print(-2, "\t-v     : toggle printing verbose information [default = %s]\n", wlcYesNo(pPars.fVerbose))
	Abc_Print(-2, "\t-w     : toggle printing verbose PDR output [default = %s]\n", wlcYesNo(pPars.fPdrVerbose))
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
func Abc_CommandMemAbs(pAbc *Abc_Frame_t, argc int32, argv []string) int32 {
	var pNtk *Wlc_Ntk_t = Wlc_AbcGetNtk(pAbc)
	var c int32
	var nIterMax int32 = 1000
	var fDumpAbs int32
	var fPdrVerbose int32
	var fVerbose int32
	Extra_UtilGetoptReset()
	for c = Extra_UtilGetopt(argc, argv, "Idwvh"); c != -1; c = Extra_UtilGetopt(argc, argv, "Idwvh") {
		switch c {
		case 'I':
			if globalUtilOptind >= argc {
				Abc_Print(-1, "Command line switch \"-I\" should be followed by an integer.\n")
				goto usage
			}
			nIterMax = wlcAtoi(argv[globalUtilOptind])
			globalUtilOptind++
			if nIterMax <= 0 {
				goto usage
			}
			break
		case 'd':
			fDumpAbs ^= 1
			break
		case 'w':
			fPdrVerbose ^= 1
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
	if pNtk == nil {
		Abc_Print(1, "Abc_CommandMemAbs(): There is no current design.\n")
		return 0
	}
	Wlc_NtkMemAbstract(pNtk, nIterMax, fDumpAbs, fPdrVerbose, fVerbose)
	return 0
usage:
	Abc_Print(-2, "usage: %%memabs [-I num] [-dwvh]\n")
	Abc_Print(-2, "\t         memory abstraction for word-level networks\n")
	Abc_Print(-2, "\t-I num : maximum number of CEGAR iterations [default = %d]\n", nIterMax)
	Abc_Print(-2, "\t-d     : toggle dumping abstraction as an AIG [default = %s]\n", wlcYesNo(fDumpAbs))
	Abc_Print(-2, "\t-w     : toggle printing verbose PDR output [default = %s]\n", wlcYesNo(fPdrVerbose))
	Abc_Print(-2, "\t-v     : toggle printing verbose information [default = %s]\n", wlcYesNo(fVerbose))
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
func Abc_CommandMemAbs2(pAbc *Abc_Frame_t, argc int32, argv []string) int32 {
	var pNtk *Wlc_Ntk_t = Wlc_AbcGetNtk(pAbc)
	var c, nFrames, fVerbose int32
	Extra_UtilGetoptReset()
	for c = Extra_UtilGetopt(argc, argv, "Fvh"); c != -1; c = Extra_UtilGetopt(argc, argv, "Fvh") {
		switch c {
		case 'F':
			if globalUtilOptind >= argc {
				Abc_Print(-1, "Command line switch \"-I\" should be followed by an integer.\n")
				goto usage
			}
			nFrames = wlcAtoi(argv[globalUtilOptind])
			globalUtilOptind++
			if nFrames <= 0 {
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
	if pNtk == nil {
		Abc_Print(1, "Abc_CommandMemAbs2(): There is no current design.\n")
		return 0
	}
	pNtk = Wlc_NtkAbstractMem(pNtk, nFrames, fVerbose)
	Wlc_AbcUpdateNtk(pAbc, pNtk)
	return 0
usage:
	Abc_Print(-2, "usage: %%memabs2 [-F num] [-vh]\n")
	Abc_Print(-2, "\t         memory abstraction for word-level networks\n")
	Abc_Print(-2, "\t-F num : the number of timeframes [default = %d]\n", nFrames)
	Abc_Print(-2, "\t-v     : toggle printing verbose information [default = %s]\n", wlcYesNo(fVerbose))
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
func Abc_CommandBlast(pAbc *Abc_Frame_t, argc int32, argv []string) int32 {
	var pNtk *Wlc_Ntk_t = Wlc_AbcGetNtk(pAbc)
	var pNew *Gia_Man_t = nil
	var c, fMiter, fDumpNames, fPrintInputInfo int32
	var Par Wlc_BstPar_t
	var pPar *Wlc_BstPar_t = &Par
	Wlc_BstParDefault(pPar)
	pPar.nOutputRange = 2
	Extra_UtilGetoptReset()
	for c = Extra_UtilGetopt(argc, argv, "ORAMcombqaydestnizvh"); c != -1; c = Extra_UtilGetopt(argc, argv, "ORAMcombqaydestnizvh") {
		switch c {
		case 'O':
			if globalUtilOptind >= argc {
				Abc_Print(-1, "Command line switch \"-O\" should be followed by an integer.\n")
				goto usage
			}
			pPar.iOutput = wlcAtoi(argv[globalUtilOptind])
			globalUtilOptind++
			if pPar.iOutput < 0 {
				goto usage
			}
			break
		case 'R':
			if globalUtilOptind >= argc {
				Abc_Print(-1, "Command line switch \"-R\" should be followed by an integer.\n")
				goto usage
			}
			pPar.nOutputRange = wlcAtoi(argv[globalUtilOptind])
			globalUtilOptind++
			if pPar.nOutputRange < 0 {
				goto usage
			}
			break
		case 'A':
			if globalUtilOptind >= argc {
				Abc_Print(-1, "Command line switch \"-A\" should be followed by an integer.\n")
				goto usage
			}
			pPar.nAdderLimit = wlcAtoi(argv[globalUtilOptind])
			globalUtilOptind++
			if pPar.nAdderLimit < 0 {
				goto usage
			}
			break
		case 'M':
			if globalUtilOptind >= argc {
				Abc_Print(-1, "Command line switch \"-M\" should be followed by an integer.\n")
				goto usage
			}
			pPar.nMultLimit = wlcAtoi(argv[globalUtilOptind])
			globalUtilOptind++
			if pPar.nMultLimit < 0 {
				goto usage
			}
			break
		case 'c':
			pPar.fGiaSimple ^= 1
			break
		case 'o':
			pPar.fAddOutputs ^= 1
			break
		case 'm':
			pPar.fMulti ^= 1
			break
		case 'b':
			pPar.fBooth ^= 1
			break
		case 'q':
			pPar.fNonRest ^= 1
			break
		case 'a':
			pPar.fCla ^= 1
			break
		case 'y':
			pPar.fDivBy0 ^= 1
			break
		case 'd':
			pPar.fCreateMiter ^= 1
			break
		case 'e':
			pPar.fCreateWordMiter ^= 1
			break
		case 's':
			pPar.fDecMuxes ^= 1
			break
		case 't':
			pPar.fCreateMiter ^= 1
			fMiter ^= 1
			break
		case 'n':
			fDumpNames ^= 1
			break
		case 'i':
			fPrintInputInfo ^= 1
			break
		case 'z':
			pPar.fSaveFfNames ^= 1
			break
		case 'v':
			pPar.fVerbose ^= 1
			break
		case 'h':
			goto usage
		default:
			goto usage
		}
	}
	if pNtk == nil {
		Abc_Print(1, "Abc_CommandBlast(): There is no current design.\n")
		return 0
	}
	if pNtk.fAsyncRst != 0 {
		Abc_Print(1, "Abc_CommandBlast(): Trying to bit-blast network with asynchronous reset.\n")
		return 0
	}
	if fPrintInputInfo != 0 {
		Wlc_NtkPrintInputInfo(pNtk)
	}
	if pPar.fMulti != 0 {
		pPar.vBoxIds = Wlc_NtkCollectMultipliers(pNtk)
		if pPar.vBoxIds == nil {
			Abc_Print(1, "Warning:  There is no multipliers in the design.\n")
		}

	} else if pPar.nAdderLimit != 0 || pPar.nMultLimit != 0 {
		var CountA, CountM int32
		pPar.vBoxIds = Wlc_NtkCollectAddMult(pNtk, pPar, &CountA, &CountM)
		if pPar.vBoxIds == nil {
			Abc_Print(1, "Warning:  There is no adders and multipliers that will not be blasted.\n")
		} else {
			Abc_Print(1, "Warning:  %d adders and %d multipliers will not be blasted.\n", CountA, CountM)
		}
	}
	if pPar.iOutput >= 0 && pPar.iOutput+pPar.nOutputRange > Wlc_NtkPoNum(pNtk) {
		Abc_Print(1, "Abc_CommandBlast(): The output range [%d:%d] is incorrect.\n", pPar.iOutput, pPar.iOutput+pPar.nOutputRange-1)
		return 0
	}
	// transform
	pNew = Wlc_NtkBitBlast(pNtk, pPar)
	Vec_IntFreeP(&pPar.vBoxIds)
	if pNew == nil {
		Abc_Print(1, "Abc_CommandBlast(): Bit-blasting has failed.\n")
		return 0
	}
	// generate miter
	if fMiter != 0 {
		var pTemp *Gia_Man_t = pNew
		pNew = Gia_ManTransformMiter(pNew)
		Gia_ManStop(pTemp)
		Abc_Print(1, "Bit-blasting created a traditional multi-output miter by XORing POs pair-wise.\n")
		if fDumpNames != 0 {
			var i int32
			var pName string
			var pFile *os.File
			pFile, _ = os.OpenFile("pio_name_map.txt", os.O_WRONLY|os.O_CREATE|os.O_TRUNC, 0666)
			if inputNames := pNew.InputNames(); inputNames != nil {
				for i = 0; i < int32(len(inputNames)); i++ {
					pName = inputNames[i]
					fmt.Fprintf(pFile, "i%d %s\n", i, pName)
				}
			}
			if outputNames := pNew.OutputNames(); outputNames != nil {
				for i = 0; i < int32(len(outputNames)); i++ {
					pName = outputNames[i]
					fmt.Fprintf(pFile, "o%d %s\n", i, pName)
				}
			}
			_ = (pFile).Close()
			Abc_Print(1, "Finished dumping file \"pio_name_map.txt\" containing PI/PO name mapping.\n")
		}
	}
	Abc_FrameUpdateGia(pAbc, pNew)
	return 0
usage:
	Abc_Print(-2, "usage: %%blast [-ORAM num] [-combqaydestnizvh]\n")
	Abc_Print(-2, "\t         performs bit-blasting of the word-level design\n")
	Abc_Print(-2, "\t-O num : zero-based index of the first word-level PO to bit-blast [default = %d]\n", pPar.iOutput)
	Abc_Print(-2, "\t-R num : the total number of word-level POs to bit-blast [default = %d]\n", pPar.nOutputRange)
	Abc_Print(-2, "\t-A num : blast adders smaller than this (0 = unused) [default = %d]\n", pPar.nAdderLimit)
	Abc_Print(-2, "\t-M num : blast multipliers smaller than this (0 = unused) [default = %d]\n", pPar.nMultLimit)
	Abc_Print(-2, "\t-c     : toggle using AIG w/o const propagation and strashing [default = %s]\n", wlcYesNo(pPar.fGiaSimple))
	Abc_Print(-2, "\t-o     : toggle using additional POs on the word-level boundaries [default = %s]\n", wlcYesNo(pPar.fAddOutputs))
	Abc_Print(-2, "\t-m     : toggle creating boxes for all multipliers in the design [default = %s]\n", wlcYesNo(pPar.fMulti))
	Abc_Print(-2, "\t-b     : toggle generating radix-4 Booth multipliers [default = %s]\n", wlcYesNo(pPar.fBooth))
	Abc_Print(-2, "\t-q     : toggle generating non-restoring square root [default = %s]\n", wlcYesNo(pPar.fNonRest))
	Abc_Print(-2, "\t-a     : toggle generating carry-look-ahead adder [default = %s]\n", wlcYesNo(pPar.fCla))
	Abc_Print(-2, "\t-y     : toggle creating different divide-by-0 condition [default = %s]\n", wlcYesNo(pPar.fDivBy0))
	Abc_Print(-2, "\t-d     : toggle creating dual-output multi-output miter [default = %s]\n", wlcYesNo(pPar.fCreateMiter))
	Abc_Print(-2, "\t-e     : toggle creating miter with output word bits combined [default = %s]\n", wlcYesNo(pPar.fCreateWordMiter))
	Abc_Print(-2, "\t-s     : toggle creating decoded MUXes [default = %s]\n", wlcYesNo(pPar.fDecMuxes))
	Abc_Print(-2, "\t-t     : toggle creating regular multi-output miter [default = %s]\n", wlcYesNo(fMiter))
	Abc_Print(-2, "\t-n     : toggle dumping signal names into a text file [default = %s]\n", wlcYesNo(fDumpNames))
	Abc_Print(-2, "\t-i     : toggle to print input names after blasting [default = %s]\n", wlcYesNo(fPrintInputInfo))
	Abc_Print(-2, "\t-z     : toggle saving flop names after blasting [default = %s]\n", wlcYesNo(pPar.fSaveFfNames))
	Abc_Print(-2, "\t-v     : toggle printing verbose information [default = %s]\n", wlcYesNo(pPar.fVerbose))
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
func Abc_CommandBlastMem(pAbc *Abc_Frame_t, argc int32, argv []string) int32 {
	var pNtk *Wlc_Ntk_t = Wlc_AbcGetNtk(pAbc)
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
	if pNtk == nil {
		Abc_Print(1, "Abc_CommandBlastMem(): There is no current design.\n")
		return 0
	}
	pNtk = Wlc_NtkMemBlast(pNtk)
	Wlc_AbcUpdateNtk(pAbc, pNtk)
	return 0
usage:
	Abc_Print(-2, "usage: %%blastmem [-vh]\n")
	Abc_Print(-2, "\t         performs blasting of memory read/write ports\n")
	Abc_Print(-2, "\t-v     : toggle printing verbose information [default = %s]\n", wlcYesNo(fVerbose))
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
func Abc_CommandGraft(pAbc *Abc_Frame_t, argc int32, argv []string) int32 {
	var pNtk *Wlc_Ntk_t = Wlc_AbcGetNtk(pAbc)
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
	if pNtk == nil {
		Abc_Print(1, "Abc_CommandGraft(): There is no current design.\n")
		return 0
	}
	pNtk = Wlc_NtkGraftMulti(pNtk, fVerbose)
	Wlc_AbcUpdateNtk(pAbc, pNtk)
	return 0
usage:
	Abc_Print(-2, "usage: %%graft [-vh]\n")
	Abc_Print(-2, "\t         detects multipliers in LHS of the miter and moves them to RHS\n")
	Abc_Print(-2, "\t-v     : toggle printing verbose information [default = %s]\n", wlcYesNo(fVerbose))
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
func Abc_CommandRetime(pAbc *Abc_Frame_t, argc int32, argv []string) int32 {
	var pFile *os.File
	var pFileName string
	var fSkipSimple int32
	var c, fDump, fVerbose int32
	Extra_UtilGetoptReset()
	for c = Extra_UtilGetopt(argc, argv, "sdvh"); c != -1; c = Extra_UtilGetopt(argc, argv, "sdvh") {
		switch c {
		case 's':
			fSkipSimple ^= 1
			break
		case 'd':
			fDump ^= 1
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
	if pAbc.Resource("pNdr") != nil {
		var vMoves *Vec_Int_t
		var pNtk *wln.Wln_Ntk_t = wln.Wln_NtkFromNdr(pAbc.Resource("pNdr").(*miniaig.Ndr_Data_t), fDump)
		wln.Wln_NtkRetimeCreateDelayInfo(pNtk)
		if pNtk == nil {
			fmt.Printf("Transforming NDR into internal represnetation has failed.\n")
			return 0
		}
		vMoves = wln.Wln_NtkRetime(pNtk, fSkipSimple, fVerbose)
		wln.Wln_NtkFree(pNtk)
		pAbc.SetIntValues("pNdrArray", nil)
		if vMoves != nil {
			pAbc.SetIntValues("pNdrArray", Vec_IntReleaseNewArray(vMoves))
		}
		Vec_IntFreeP(&vMoves)
		return 0
	}
	if argc != globalUtilOptind+1 {
		fmt.Printf("Abc_CommandRetime(): Input file name should be given on the command line.\n")
		return 0
	}
	// get the file name
	pFileName = argv[globalUtilOptind]
	pFile, _ = os.Open(pFileName)
	if pFile == nil {
		Abc_Print(1, "Cannot open input file \"%s\". ", pFileName)
		if pFileName = Extra_FileGetSimilarName(pFileName, ".ndr", "", "", "", ""); pFileName != "" {
			Abc_Print(1, "Did you mean \"%s\"?", pFileName)
		}
		Abc_Print(1, "\n")
		return 0
	}
	_ = (pFile).Close()
	wln.Wln_NtkRetimeTest(pFileName, fSkipSimple, fDump, fVerbose)
	return 0
usage:
	Abc_Print(-2, "usage: %%retime [-sdvh]\n")
	Abc_Print(-2, "\t         performs retiming for the NDR design\n")
	Abc_Print(-2, "\t-s     : toggle printing simple nodes [default = %s]\n", wlcYesNoBool(fSkipSimple == 0))
	Abc_Print(-2, "\t-d     : toggle dumping the network in Verilog [default = %s]\n", wlcYesNo(fDump))
	Abc_Print(-2, "\t-v     : toggle printing verbose information [default = %s]\n", wlcYesNo(fVerbose))
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
func Abc_CommandProfile(pAbc *Abc_Frame_t, argc int32, argv []string) int32 {
	var pNtk *Wlc_Ntk_t = Wlc_AbcGetNtk(pAbc)
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
	if pNtk == nil {
		Abc_Print(1, "Abc_CommandProfile(): There is no current design.\n")
		return 0
	}
	Wlc_WinProfileArith(pNtk)
	return 0
usage:
	Abc_Print(-2, "usage: %%profile [-vh]\n")
	Abc_Print(-2, "\t         profiles arithmetic components in the word-level networks\n")
	Abc_Print(-2, "\t-v     : toggle printing verbose information [default = %s]\n", wlcYesNo(fVerbose))
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
func Abc_CommandShortNames(pAbc *Abc_Frame_t, argc int32, argv []string) int32 {
	var pNtk *Wlc_Ntk_t = Wlc_AbcGetNtk(pAbc)
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
	if pNtk == nil {
		Abc_Print(1, "Abc_CommandProfile(): There is no current design.\n")
		return 0
	}
	Wlc_NtkShortNames(pNtk)
	return 0
usage:
	Abc_Print(-2, "usage: %%short_names [-vh]\n")
	Abc_Print(-2, "\t         derives short names for all objects of the network\n")
	Abc_Print(-2, "\t-v     : toggle printing verbose information [default = %s]\n", wlcYesNo(fVerbose))
	Abc_Print(-2, "\t-h     : print the command usage\n")
	return 1
}

/*
*Function*************************************************************

	Synopsis    []

	Description []

	SideEffects []

	SeeAlso     []

**********************************************************************
*/
func Abc_CommandShow(pAbc *Abc_Frame_t, argc int32, argv []string) int32 {
	var pNtk *Wlc_Ntk_t = Wlc_AbcGetNtk(pAbc)
	var c, fMemory, fVerbose int32
	// set defaults
	Extra_UtilGetoptReset()
	for c = Extra_UtilGetopt(argc, argv, "mvh"); c != -1; c = Extra_UtilGetopt(argc, argv, "mvh") {
		switch c {
		case 'm':
			fMemory ^= 1
			break
		case 'v':
			fVerbose ^= 1
			break
		default:
			goto usage
		}
	}
	if pNtk == nil {
		Abc_Print(-1, "Empty network.\n")
		return 1
	}
	if fMemory != 0 {
		var vTemp *Vec_Int_t = Wlc_NtkCollectMemory(pNtk, 1)
		Wlc_NtkShow(pNtk, vTemp)
		Vec_IntFree(vTemp)

	} else {
		Wlc_NtkShow(pNtk, nil)
	}
	return 0

usage:
	Abc_Print(-2, "usage: %%show [-mh]\n")
	Abc_Print(-2, "          visualizes the network structure using DOT and GSVIEW\n")
	Abc_Print(-2, "          \"dot.exe\" and \"gsview32.exe\" should be set in the paths\n")
	Abc_Print(-2, "          (\"gsview32.exe\" may be in \"C:\\Program Files\\Ghostgum\\gsview\\\")\n")
	Abc_Print(-2, "\t-m   :  toggle showing memory subsystem [default = %s]\n", wlcYesNo(fMemory))
	Abc_Print(-2, "\t-h   :  print the command usage\n")
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
func Abc_CommandInvPs(pAbc *Abc_Frame_t, argc int32, argv []string) int32 {
	var pNtk *Wlc_Ntk_t = Wlc_AbcGetNtk(pAbc)
	var vCounts *Vec_Int_t
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
	if pNtk == nil {
		Abc_Print(1, "Abc_CommandInvPs(): There is no current design.\n")
		return 0
	}
	if Wlc_AbcGetInv(pAbc) == nil {
		Abc_Print(1, "Abc_CommandInvPs(): Invariant is not available.\n")
		return 0
	}
	vCounts = Pdr_InvCounts(Wlc_AbcGetInv(pAbc))
	Wlc_NtkPrintInvStats(pNtk, vCounts, fVerbose)
	Vec_IntFree(vCounts)
	return 0
usage:
	Abc_Print(-2, "usage: inv_ps [-vh]\n")
	Abc_Print(-2, "\t         prints statistics for inductive invariant\n")
	Abc_Print(-2, "\t         (in the case of 'sat' or 'undecided', inifity clauses are used)\n")
	Abc_Print(-2, "\t-v     : toggle printing verbose information [default = %s]\n", wlcYesNo(fVerbose))
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
func Abc_CommandInvPrint(pAbc *Abc_Frame_t, argc int32, argv []string) int32 {
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
	if Wlc_AbcGetInv(pAbc) == nil {
		Abc_Print(1, "Abc_CommandInvPs(): Invariant is not available.\n")
		return 0
	}
	Pdr_InvPrint(Wlc_AbcGetInv(pAbc), fVerbose)
	return 0
usage:
	Abc_Print(-2, "usage: inv_print [-vh]\n")
	Abc_Print(-2, "\t         prints the current inductive invariant\n")
	Abc_Print(-2, "\t         (in the case of 'sat' or 'undecided', inifity clauses are used)\n")
	Abc_Print(-2, "\t-v     : toggle printing verbose information [default = %s]\n", wlcYesNo(fVerbose))
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
func Abc_CommandInvCheck(pAbc *Abc_Frame_t, argc int32, argv []string) int32 {
	var clk abctime = Abc_Clock()
	var c, nFailed, fVerbose int32
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
	if Wlc_AbcGetGia(pAbc) == nil {
		Abc_Print(1, "Abc_CommandInvMin(): There is no current design.\n")
		return 0
	}
	if Wlc_AbcGetInv(pAbc) == nil {
		Abc_Print(1, "Abc_CommandInvMin(): There is no saved invariant.\n")
		return 0
	}
	if Gia_ManRegNum(Wlc_AbcGetGia(pAbc)) != Vec_IntEntryLast(Wlc_AbcGetInv(pAbc)) {
		Abc_Print(1, "Abc_CommandInvMin(): The number of flops in the invariant and in GIA should be the same.\n")
		return 0
	}
	nFailed = Pdr_InvCheck(Wlc_AbcGetGia(pAbc), Wlc_AbcGetInv(pAbc), fVerbose)
	if nFailed != 0 {
		fmt.Printf("Invariant verification failed for %d clauses (out of %d). ", nFailed, Vec_IntEntry(Wlc_AbcGetInv(pAbc), 0))
	} else {
		fmt.Printf("Invariant verification succeeded.    ")
	}
	Abc_PrintTime(1, "Time", Abc_Clock()-clk)
	return 0
usage:
	Abc_Print(-2, "usage: inv_check [-vh]\n")
	Abc_Print(-2, "\t         checks that the invariant is indeed an inductive invariant\n")
	Abc_Print(-2, "\t         (AIG representing the design should be in the &-space)\n")
	Abc_Print(-2, "\t-v     : toggle printing verbose information [default = %s]\n", wlcYesNo(fVerbose))
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
func Abc_CommandInvGet(pAbc *Abc_Frame_t, argc int32, argv []string) int32 {
	var pMainNtk *Abc_Ntk_t
	var pNtk *Wlc_Ntk_t = Wlc_AbcGetNtk(pAbc)
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
	if Wlc_AbcGetInv(pAbc) == nil {
		Abc_Print(1, "Abc_CommandInvGet(): Invariant is not available.\n")
		return 0
	}
	// derive the network
	pMainNtk = Wlc_NtkGetInv(pNtk, Wlc_AbcGetInv(pAbc))
	// replace the current network
	if pMainNtk != nil {
		Abc_FrameReplaceCurrentNetwork(pAbc, pMainNtk)
	}
	return 0
usage:
	Abc_Print(-2, "usage: inv_get [-vh]\n")
	Abc_Print(-2, "\t         places invariant found by PDR as the current network in the main-space\n")
	Abc_Print(-2, "\t         (in the case of 'sat' or 'undecided', inifity clauses are used)\n")
	Abc_Print(-2, "\t-v     : toggle printing verbose information [default = %s]\n", wlcYesNo(fVerbose))
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
func Abc_CommandInvPut(pAbc *Abc_Frame_t, argc int32, argv []string) int32 {
	var vInv *Vec_Int_t = nil
	var pNtk *Abc_Ntk_t = Abc_FrameReadNtk(pAbc)
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
	if pNtk == nil {
		Abc_Print(1, "Abc_CommandInvPut(): There is no current design.\n")
		return 0
	}
	if Wlc_AbcGetGia(pAbc) == nil {
		Abc_Print(1, "Abc_CommandInvPut(): There is no current AIG.\n")
		return 0
	}
	// derive the network
	vInv = Wlc_NtkGetPut(pNtk, Wlc_AbcGetGia(pAbc))
	if vInv != nil {
		Wlc_AbcSetInv(pAbc, vInv)
	}
	return 0
usage:
	Abc_Print(-2, "usage: inv_put [-vh]\n")
	Abc_Print(-2, "\t         inputs the current network in the main-space as an invariant\n")
	Abc_Print(-2, "\t         (AIG representing the design should be in the &-space)\n")
	Abc_Print(-2, "\t-v     : toggle printing verbose information [default = %s]\n", wlcYesNo(fVerbose))
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
func Abc_CommandInvMin(pAbc *Abc_Frame_t, argc int32, argv []string) int32 {
	var vInv, vInv2 *Vec_Int_t
	var c, fLits, fVerbose int32
	Extra_UtilGetoptReset()
	for c = Extra_UtilGetopt(argc, argv, "lvh"); c != -1; c = Extra_UtilGetopt(argc, argv, "lvh") {
		switch c {
		case 'l':
			fLits ^= 1
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
	if Wlc_AbcGetGia(pAbc) == nil {
		Abc_Print(1, "Abc_CommandInvMin(): There is no current design.\n")
		return 0
	}
	if Wlc_AbcGetInv(pAbc) == nil {
		Abc_Print(1, "Abc_CommandInvMin(): Invariant is not available.\n")
		return 0
	}
	vInv = Wlc_AbcGetInv(pAbc)
	if Gia_ManRegNum(Wlc_AbcGetGia(pAbc)) != Vec_IntEntryLast(vInv) {
		Abc_Print(1, "Abc_CommandInvMin(): The number of flops in the invariant and in GIA should be the same.\n")
		return 0
	}
	if fLits != 0 {
		vInv2 = Pdr_InvMinimizeLits(Wlc_AbcGetGia(pAbc), vInv, fVerbose)
	} else {
		vInv2 = Pdr_InvMinimize(Wlc_AbcGetGia(pAbc), vInv, fVerbose)
	}
	if vInv2 != nil {
		Wlc_AbcSetInv(pAbc, vInv2)
	}
	return 0
usage:
	Abc_Print(-2, "usage: inv_min [-lvh]\n")
	Abc_Print(-2, "\t         performs minimization of the current invariant\n")
	Abc_Print(-2, "\t         (AIG representing the design should be in the &-space)\n")
	Abc_Print(-2, "\t-l     : toggle minimizing literals rather than clauses [default = %s]\n", wlcYesNo(fLits))
	Abc_Print(-2, "\t-v     : toggle printing verbose information [default = %s]\n", wlcYesNo(fVerbose))
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
func Abc_CommandTest(pAbc *Abc_Frame_t, argc int32, argv []string) int32 {
	var pNtk *Wlc_Ntk_t = Wlc_AbcGetNtk(pAbc)
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
	//    if ( pNtk == NULL )
	//    {
	//        Abc_Print( 1, "Abc_CommandTest(): There is no current design.\n" );
	//        return 0;
	//    }

	// transform
	//pNtk = Wlc_NtkUifNodePairs( pNtk, NULL );
	//pNtk = Wlc_NtkAbstractNodes( pNtk, NULL );
	//Wlc_AbcUpdateNtk( pAbc, pNtk );
	//Wlc_GenerateSmtStdout( pAbc );
	//Wlc_NtkSimulateTest( (Wlc_Ntk_t *)pAbc->pAbcWlc );
	//pNtk = Wlc_NtkDupSingleNodes( pNtk );
	//Wlc_AbcUpdateNtk( pAbc, pNtk );
	//Wln_ReadNdrTest();
	//pNtk = Wlc_NtkMemAbstractTest( pNtk );
	//Wlc_AbcUpdateNtk( pAbc, pNtk );
	//Wln_NtkFromWlcTest( pNtk );
	Wlc_NtkExploreMem(pNtk, 0)
	return 0
usage:
	Abc_Print(-2, "usage: %%test [-vh]\n")
	Abc_Print(-2, "\t         experiments with word-level networks\n")
	Abc_Print(-2, "\t-v     : toggle printing verbose information [default = %s]\n", wlcYesNo(fVerbose))
	Abc_Print(-2, "\t-h     : print the command usage\n")
	return 1
}

////////////////////////////////////////////////////////////////////////
///                       END OF FILE                                ///
////////////////////////////////////////////////////////////////////////

//ABC_NAMESPACE_IMPL_END
