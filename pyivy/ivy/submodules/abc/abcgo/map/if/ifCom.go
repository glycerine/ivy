package if_

import (
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"

	abccmd "github.com/glycerine/abcgo/base/cmd"
	abcmain "github.com/glycerine/abcgo/base/main"
)

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

/*
*Function*************************************************************

	Synopsis    [Package initialization procedure.]

	Description []

	SideEffects []

	SeeAlso     []

**********************************************************************
*/
func If_Init(pAbc *Abc_Frame_t) {
	// set the default library
	var s_LutLib = If_LibLut_t{pName: "lutlib", LutMax: 4, pLutAreas: [IF_MAX_LUTSIZE + 1]float32{0, 1, 1, 1, 1}, pLutDelays: [IF_MAX_LUTSIZE + 1][IF_MAX_LUTSIZE + 1]float32{{0}, {1}, {1}, {1}, {1}}}
	Abc_FrameSetLibLut(If_LibLutDup(&s_LutLib))

	abccmd.Cmd_CommandAdd(pAbc, "FPGA mapping", "read_lut", If_CommandReadLut, 0)
	abccmd.Cmd_CommandAdd(pAbc, "FPGA mapping", "print_lut", If_CommandPrintLut, 0)

	abccmd.Cmd_CommandAdd(pAbc, "FPGA mapping", "read_box", If_CommandReadBox, 0)
	abccmd.Cmd_CommandAdd(pAbc, "FPGA mapping", "print_box", If_CommandPrintBox, 0)
}

/*
*Function*************************************************************

	Synopsis    [Package ending procedure.]

	Description []

	SideEffects []

	SeeAlso     []

**********************************************************************
*/
func If_End(pAbc *Abc_Frame_t) {
	If_LibLutFree(Abc_FrameReadLibLut())
	If_LibBoxFree(Abc_FrameReadLibBox())
}

/*
*Function*************************************************************

	Synopsis    [Command procedure to read LUT libraries.]

	Description []

	SideEffects []

	SeeAlso     []

**********************************************************************
*/
func If_CommandReadLut(pAbc *Abc_Frame_t, argc int32, argv []string) int32 {
	var pFile *os.File
	var pErr io.Writer
	var pLib *If_LibLut_t
	var FileName string
	var fVerbose int32
	var c int32

	pErr = Abc_FrameReadErr(pAbc)

	// set the defaults
	fVerbose = 1
	ifGetoptReset()
	for c = ifGetopt(argc, argv, "vh"); c != -1; c = ifGetopt(argc, argv, "vh") {
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

	if argc != ifGetoptIndex()+1 {
		goto usage
	}

	// get the input file name
	FileName = argv[int(ifGetoptIndex())]
	pFile, _ = os.Open(FileName)
	if pFile == nil {
		fmt.Fprintf(pErr, "Cannot open input file \"%s\". ", FileName)
		if similar := ifSimilarLibraryFile(FileName, ".genlib", ".lib", ".gen", ".g"); similar != "" {
			fmt.Fprintf(pErr, "Did you mean \"%s\"?", FileName)
		}
		fmt.Fprintf(pErr, "\n")
		return 1
	}
	_ = (pFile).Close()

	// set the new network
	pLib = If_LibLutRead(FileName)
	if pLib == nil {
		fmt.Fprintf(pErr, "Reading LUT library has failed.\n")
		goto usage
	}
	// replace the current library
	If_LibLutFree(Abc_FrameReadLibLut())
	Abc_FrameSetLibLut(pLib)
	return 0

usage:
	fmt.Fprintf(pErr, "\nusage: read_lut [-vh]\n")
	fmt.Fprintf(pErr, "\t          read the LUT library from the file\n")
	fmt.Fprintf(pErr, "\t-v      : toggles enabling of verbose output [default = %s]\n", ifVerbosity(fVerbose))
	fmt.Fprintf(pErr, "\t-h      : print the command usage\n")
	fmt.Fprintf(pErr, "\t                                        \n")
	fmt.Fprintf(pErr, "\t          File format for a LUT library:\n")
	fmt.Fprintf(pErr, "\t          (the default library is shown)\n")
	fmt.Fprintf(pErr, "\t                                        \n")
	fmt.Fprintf(pErr, "\t          # The area/delay of k-variable LUTs:\n")
	fmt.Fprintf(pErr, "\t          # k  area   delay\n")
	fmt.Fprintf(pErr, "\t          1      1      1\n")
	fmt.Fprintf(pErr, "\t          2      2      2\n")
	fmt.Fprintf(pErr, "\t          3      4      3\n")
	fmt.Fprintf(pErr, "\t          4      8      4\n")
	fmt.Fprintf(pErr, "\t          5     16      5\n")
	fmt.Fprintf(pErr, "\t          6     32      6\n")
	return 1 /* error exit */
}

/*
*Function*************************************************************

	Synopsis    [Command procedure to read LUT libraries.]

	Description []

	SideEffects []

	SeeAlso     []

**********************************************************************
*/
func If_CommandPrintLut(pAbc *Abc_Frame_t, argc int32, argv []string) int32 {
	var pErr io.Writer
	var fVerbose int32
	var c int32

	pErr = Abc_FrameReadErr(pAbc)

	// set the defaults
	fVerbose = 1
	ifGetoptReset()
	for c = ifGetopt(argc, argv, "vh"); c != -1; c = ifGetopt(argc, argv, "vh") {
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

	if argc != ifGetoptIndex() {
		goto usage
	}

	// set the new network
	If_LibLutPrint(Abc_FrameReadLibLut())
	return 0

usage:
	fmt.Fprintf(pErr, "\nusage: print_lut [-vh]\n")
	fmt.Fprintf(pErr, "\t          print the current LUT library\n")
	fmt.Fprintf(pErr, "\t-v      : toggles enabling of verbose output [default = %s]\n", ifVerbosity(fVerbose))
	fmt.Fprintf(pErr, "\t-h      : print the command usage\n")
	return 1 /* error exit */
}

/*
*Function*************************************************************

	Synopsis    []

	Description []

	SideEffects []

	SeeAlso     []

**********************************************************************
*/
func If_CommandReadBox(pAbc *Abc_Frame_t, argc int32, argv []string) int32 {
	var pFile *os.File
	var pErr io.Writer
	var pLib *If_LibBox_t
	var FileName string
	var fExtended int32
	var fVerbose int32
	var c int32

	pErr = Abc_FrameReadErr(pAbc)

	// set the defaults
	fExtended = 0
	fVerbose = 1
	ifGetoptReset()
	for c = ifGetopt(argc, argv, "evh"); c != -1; c = ifGetopt(argc, argv, "evh") {
		switch c {
		case 'e':
			fExtended ^= 1
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

	if argc != ifGetoptIndex()+1 {
		goto usage
	}

	// get the input file name
	FileName = argv[int(ifGetoptIndex())]
	pFile, _ = os.Open(FileName)
	if pFile == nil {
		fmt.Fprintf(pErr, "Cannot open input file \"%s\". ", FileName)
		if similar := ifSimilarLibraryFile(FileName, ".genlib", ".lib", ".gen", ".g"); similar != "" {
			fmt.Fprintf(pErr, "Did you mean \"%s\"?", FileName)
		}
		fmt.Fprintf(pErr, "\n")
		return 1
	}
	_ = (pFile).Close()

	// set the new network
	if fExtended != 0 {
		pLib = If_LibBoxRead2(FileName)
	} else {
		pLib = If_LibBoxRead(FileName)
	}
	if pLib == nil {
		fmt.Fprintf(pErr, "Reading box library has failed.\n")
		goto usage
	}
	// replace the current library
	If_LibBoxFree(Abc_FrameReadLibBox())
	Abc_FrameSetLibBox(pLib)
	return 0

usage:
	fmt.Fprintf(pErr, "\nusage: read_box [-evh]\n")
	fmt.Fprintf(pErr, "\t          read the box library from the file\n")
	fmt.Fprintf(pErr, "\t-e      : toggles reading extended format [default = %s]\n", ifVerbosity(fExtended))
	fmt.Fprintf(pErr, "\t-v      : toggles enabling of verbose output [default = %s]\n", ifVerbosity(fVerbose))
	fmt.Fprintf(pErr, "\t-h      : print the command usage\n")
	return 1 /* error exit */
}

/*
*Function*************************************************************

	Synopsis    [Command procedure to read LUT libraries.]

	Description []

	SideEffects []

	SeeAlso     []

**********************************************************************
*/
func If_CommandPrintBox(pAbc *Abc_Frame_t, argc int32, argv []string) int32 {
	var pErr io.Writer
	var fVerbose int32
	var c int32

	pErr = Abc_FrameReadErr(pAbc)

	// set the defaults
	fVerbose = 1
	ifGetoptReset()
	for c = ifGetopt(argc, argv, "vh"); c != -1; c = ifGetopt(argc, argv, "vh") {
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

	if argc != ifGetoptIndex() {
		goto usage
	}

	// set the new network
	If_LibBoxPrint(os.Stdout, Abc_FrameReadLibBox())
	return 0

usage:
	fmt.Fprintf(pErr, "\nusage: print_box [-vh]\n")
	fmt.Fprintf(pErr, "\t          print the current box library\n")
	fmt.Fprintf(pErr, "\t-v      : toggles enabling of verbose output [default = %s]\n", ifVerbosity(fVerbose))
	fmt.Fprintf(pErr, "\t-h      : print the command usage\n")
	return 1 /* error exit */
}

////////////////////////////////////////////////////////////////////////
///                       END OF FILE                                ///
////////////////////////////////////////////////////////////////////////

//ABC_NAMESPACE_IMPL_END

// The four library commands use only flag-only getopt options; this local state
// avoids coupling the mapper to ABC's broad command-line utility package.
var ifOptionIndex, ifOptionOffset int

func ifGetoptReset()       { ifOptionIndex, ifOptionOffset = 1, 1 }
func ifGetoptIndex() int32 { return int32(ifOptionIndex) }
func ifGetopt(argc int32, argv []string, spec string) int32 {
	for ifOptionIndex < len(argv) && int32(ifOptionIndex) < argc {
		arg := argv[ifOptionIndex]
		if ifOptionOffset == 1 {
			if len(arg) < 2 || arg[0] != '-' || arg == "-" {
				return -1
			}
			if arg == "--" {
				ifOptionIndex++
				return -1
			}
			ifOptionOffset = 1
		}
		if ifOptionOffset >= len(arg) {
			ifOptionIndex++
			ifOptionOffset = 1
			continue
		}
		option := arg[ifOptionOffset]
		if !strings.ContainsRune(spec, rune(option)) {
			ifOptionOffset++
			if ifOptionOffset >= len(arg) {
				ifOptionIndex++
				ifOptionOffset = 1
			}
			return int32('?')
		}
		ifOptionOffset++
		if ifOptionOffset >= len(arg) {
			ifOptionIndex++
			ifOptionOffset = 1
		}
		return int32(option)
	}
	return -1
}

func ifVerbosity(enabled int32) string {
	if enabled != 0 {
		return "yes"
	}
	return "no"
}

func ifSimilarLibraryFile(wrong string, extensions ...string) string {
	base := strings.TrimSuffix(wrong, filepath.Ext(wrong))
	for _, ext := range extensions {
		candidate := base + ext
		if _, err := os.Stat(candidate); err == nil {
			return candidate
		}
	}
	return ""
}

type Abc_Frame_t = abcmain.Frame

func Abc_FrameReadOut(f *Abc_Frame_t) io.Writer { return abcmain.Abc_FrameReadOut(f) }
func Abc_FrameReadErr(f *Abc_Frame_t) io.Writer { return abcmain.Abc_FrameReadErr(f) }
func Abc_FrameReadLibLut() *If_LibLut_t {
	value, _ := abcmain.Abc_FrameReadLibLut().(*If_LibLut_t)
	return value
}
func Abc_FrameSetLibLut(value *If_LibLut_t) { abcmain.Abc_FrameSetLibLut(value) }
func Abc_FrameReadLibBox() *If_LibBox_t {
	value, _ := abcmain.Abc_FrameReadLibBox().(*If_LibBox_t)
	return value
}
func Abc_FrameSetLibBox(value *If_LibBox_t) { abcmain.Abc_FrameSetLibBox(value) }
