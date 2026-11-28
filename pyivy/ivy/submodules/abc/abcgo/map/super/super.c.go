package super

import (
	"fmt"
	"strconv"
	"strings"

	abcio "github.com/glycerine/abcgo/base/io"
	abcmain "github.com/glycerine/abcgo/base/main"
	"github.com/glycerine/abcgo/map/mio"
	"github.com/glycerine/abcgo/misc/extra"
)

func Super_Init(frame *abcmain.Frame) {
	for _, entry := range []struct {
		name    string
		handler func(*abcmain.Frame, int32, []string) int32
	}{{"super", Super_CommandSupergates}, {"super2", Super_CommandSupergatesAnd}} {
		handler := entry.handler
		_ = frame.RegisterCommand(abcmain.Command{Name: entry.name, Group: "SC mapping", Handler: func(f *abcmain.Frame, args []string) int32 { return handler(f, int32(len(args)), args) }})
	}
}
func Super_End(frame *abcmain.Frame) { frame.RemoveCommand("super"); frame.RemoveCommand("super2") }

type commandOptions struct {
	numbers   map[byte]string
	flags     map[byte]bool
	arguments []string
}

func parseCommand(args []string, numeric, flags string) (commandOptions, error) {
	result := commandOptions{numbers: make(map[byte]string), flags: make(map[byte]bool)}
	for i := 1; i < len(args); i++ {
		arg := args[i]
		if arg == "--" {
			result.arguments = append(result.arguments, args[i+1:]...)
			break
		}
		if !strings.HasPrefix(arg, "-") || arg == "-" {
			result.arguments = append(result.arguments, arg)
			continue
		}
		for j := 1; j < len(arg); j++ {
			flag := arg[j]
			if flag == 'h' {
				return result, fmt.Errorf("help requested")
			}
			if strings.IndexByte(flags, flag) >= 0 {
				result.flags[flag] = !result.flags[flag]
				continue
			}
			if strings.IndexByte(numeric, flag) < 0 {
				return result, fmt.Errorf("unknown option -%c", flag)
			}
			value := arg[j+1:]
			if value == "" {
				i++
				if i == len(args) {
					return result, fmt.Errorf("missing value for -%c", flag)
				}
				value = args[i]
			}
			result.numbers[flag] = value
			break
		}
	}
	return result, nil
}
func commandError(frame *abcmain.Frame, err error, usage string) int32 {
	fmt.Fprintln(frame.Err, err)
	fmt.Fprintln(frame.Err, usage)
	return 1
}
func commandInt(o commandOptions, key byte, value int) (int, error) {
	s, ok := o.numbers[key]
	if !ok {
		return value, nil
	}
	n, err := strconv.Atoi(s)
	if err != nil || n < 0 || int64(n) > 1<<31-1 {
		return 0, fmt.Errorf("invalid nonnegative -%c value", key)
	}
	return n, nil
}
func Super_CommandSupergatesAnd(frame *abcmain.Frame, argc int32, args []string) int32 {
	const usage = "usage: super2 [-I inputs] [-L levels] [-v]"
	o, err := parseCommand(args, "IL", "v")
	if err != nil {
		return commandError(frame, err, usage)
	}
	if len(o.arguments) != 0 {
		return commandError(frame, fmt.Errorf("unexpected filename"), usage)
	}
	inputs, err := commandInt(o, 'I', 4)
	if err != nil {
		return commandError(frame, err, usage)
	}
	levels, err := commandInt(o, 'L', 3)
	if err != nil {
		return commandError(frame, err, usage)
	}
	lib, err := GenerateAND(inputs, levels)
	if err != nil {
		return commandError(frame, err, usage)
	}
	if err = Super2_LibWrite(lib); err != nil {
		return commandError(frame, err, usage)
	}
	if o.flags['v'] {
		fmt.Fprintf(frame.Out, "Computed %d AND/inverter supergates\n", len(lib.Gates))
	}
	return 0
}

// Super_CommandSupergates is the "super" command of super.c.
func Super_CommandSupergates(pAbc *abcmain.Frame, argc int32, argv []string) int32 {
	var pLib *mio.Mio_Library_t
	var FileName, ExcludeFile string
	var DelayLimit, AreaLimit float32
	var fSkipInvs, fWriteOldFormat int32
	var nVarsMax, nLevels, nGatesMax, TimeLimit int32
	var fVerbose int32
	var c int32

	pErr := pAbc.Err

	// set the defaults
	nVarsMax = 5
	nLevels = 2
	DelayLimit = 0
	AreaLimit = 0
	nGatesMax = 0
	TimeLimit = 0
	fSkipInvs = 1
	fVerbose = 0
	fWriteOldFormat = 0
	ExcludeFile = ""

	extra.Extra_UtilGetoptReset()
	for {
		if c = extra.Extra_UtilGetopt(argc, argv, "ILNTDAEsovh"); c == -1 {
			break
		}
		globalUtilOptind := extra.Extra_UtilGetoptIndex()
		next := func() string {
			if int(globalUtilOptind) >= len(argv) {
				return ""
			}
			return argv[globalUtilOptind]
		}
		switch c {
		case 'I':
			nVarsMax = superAtoi(next())
			extra.Extra_UtilGetoptSetIndex(globalUtilOptind + 1)
			if nVarsMax < 0 {
				goto usage
			}
		case 'L':
			nLevels = superAtoi(next())
			extra.Extra_UtilGetoptSetIndex(globalUtilOptind + 1)
			if nLevels < 0 {
				goto usage
			}
		case 'N':
			nGatesMax = superAtoi(next())
			extra.Extra_UtilGetoptSetIndex(globalUtilOptind + 1)
			if nGatesMax < 0 {
				goto usage
			}
		case 'T':
			TimeLimit = superAtoi(next())
			extra.Extra_UtilGetoptSetIndex(globalUtilOptind + 1)
			if TimeLimit < 0 {
				goto usage
			}
		case 'D':
			DelayLimit = float32(superAtof(next()))
			extra.Extra_UtilGetoptSetIndex(globalUtilOptind + 1)
			if DelayLimit <= 0.0 {
				goto usage
			}
		case 'A':
			AreaLimit = float32(superAtof(next()))
			extra.Extra_UtilGetoptSetIndex(globalUtilOptind + 1)
			if AreaLimit <= 0.0 {
				goto usage
			}
		case 'E':
			if int(globalUtilOptind) >= len(argv) {
				goto usage
			}
			ExcludeFile = argv[globalUtilOptind]
			extra.Extra_UtilGetoptSetIndex(globalUtilOptind + 1)
		case 's':
			fSkipInvs ^= 1
		case 'o':
			fWriteOldFormat ^= 1
		case 'v':
			fVerbose ^= 1
		default:
			goto usage
		}
	}

	if argc != extra.Extra_UtilGetoptIndex()+1 {
		fmt.Fprintf(pErr, "The genlib library file should be given on the command line.\n")
		goto usage
	}

	if nVarsMax < 2 || nVarsMax > 6 {
		fmt.Fprintf(pErr, "The max number of variables (%d) should be more than 1 and less than 7.\n", nVarsMax)
		goto usage
	}

	// get the input file name
	FileName = argv[extra.Extra_UtilGetoptIndex()]
	if pFile := abcio.Io_FileOpen(FileName, "open_path", "r", 0); pFile == nil {
		fmt.Fprintf(pErr, "Cannot open input file \"%s\". ", FileName)
		if FileName = extra.Extra_FileGetSimilarName(FileName, ".genlib", ".lib", ".gen", ".g", ""); FileName != "" {
			fmt.Fprintf(pErr, "Did you mean \"%s\"?", FileName)
		}
		fmt.Fprintf(pErr, "\n")
		return 1
	} else {
		pFile.Close()
	}

	// set the new network
	pLib = mio.Mio_LibraryRead(FileName, "", ExcludeFile, fVerbose)
	if pLib == nil {
		fmt.Fprintf(pErr, "Reading library has failed.\n")
		goto usage
	}

	// compute the gates
	FileName = extra.Extra_FileNameGenericAppend(mio.Mio_LibraryReadName(pLib), ".super")
	Super_Precompute(pLib, nVarsMax, nLevels, nGatesMax, DelayLimit, AreaLimit, TimeLimit, fSkipInvs, fVerbose, FileName)

	// delete the library
	mio.Mio_LibraryDelete(pLib)
	return 0

usage:
	_ = fWriteOldFormat // parsed, but (as in C) not passed to Super_Precompute
	fmt.Fprintf(pErr, "usage: super [-ILNT num] [-DA float] [-E file] [-sovh] <genlib_file>\n")
	fmt.Fprintf(pErr, "\t         precomputes the supergates for the given genlib library\n")
	fmt.Fprintf(pErr, "\t-I num   : the max number of supergate inputs [default = %d]\n", nVarsMax)
	fmt.Fprintf(pErr, "\t-L num   : the max number of levels of gates [default = %d]\n", nLevels)
	fmt.Fprintf(pErr, "\t-N num   : the limit on the number of considered supergates [default = %d]\n", nGatesMax)
	fmt.Fprintf(pErr, "\t-T num   : the approximate runtime limit in seconds [default = %d]\n", TimeLimit)
	fmt.Fprintf(pErr, "\t-D float : the max delay of the supergates [default = %.2f]\n", DelayLimit)
	fmt.Fprintf(pErr, "\t-A float : the max area of the supergates [default = %.2f]\n", AreaLimit)
	fmt.Fprintf(pErr, "\t-E file  : file contains list of genlib gates to exclude\n")
	fmt.Fprintf(pErr, "\t-s       : toggle the use of inverters at the inputs [default = %s]\n", superYesNo(fSkipInvs == 0))
	fmt.Fprintf(pErr, "\t-o       : toggle dumping the supergate library in old format [default = %s]\n", superYesNo(fWriteOldFormat != 0))
	fmt.Fprintf(pErr, "\t-v       : enable verbose output [default = %s]\n", superYesNo(fVerbose != 0))
	fmt.Fprintf(pErr, "\t-h       : print the help message\n")
	fmt.Fprintf(pErr, "\n")
	fmt.Fprintf(pErr, "\tHere is a piece of advice on precomputing supergate libraries:\n")
	fmt.Fprintf(pErr, "\t\n")
	fmt.Fprintf(pErr, "\tStart with the number of inputs equal to 5 (-I 5), the number of \n")
	fmt.Fprintf(pErr, "\tlevels equal to 2 (-L 2), the delay equal to 2-3 delays of inverter, \n")
	fmt.Fprintf(pErr, "\tthe area equal to 2-3 areas of two input NAND, and runtime limit equal \n")
	fmt.Fprintf(pErr, "\tto 10 seconds (-T 10). Run precomputation and learn from the result.\n")
	fmt.Fprintf(pErr, "\tDetermine what parameter is most constraining and try to increase \n")
	fmt.Fprintf(pErr, "\tthe value of that parameter. The goal is to have a well-balanced\n")
	fmt.Fprintf(pErr, "\tset of constraints and the resulting supergate library containing\n")
	fmt.Fprintf(pErr, "\tapproximately 5K-20K supergates. Typically, it is better to increase\n")
	fmt.Fprintf(pErr, "\tdelay limit rather than area limit, because having large-area supergates\n")
	fmt.Fprintf(pErr, "\tmay result in a considerable increase in area.\n")
	fmt.Fprintf(pErr, "\t\n")
	fmt.Fprintf(pErr, "\tNote that a good supergate library for experiments typically can be \n")
	fmt.Fprintf(pErr, "\tprecomputed in 30 sec or less. Increasing runtime limit makes sense when\n")
	fmt.Fprintf(pErr, "\tother parameters are well-balanced and it is needed to enumerate more\n")
	fmt.Fprintf(pErr, "\tchoices to have a good result. In the end, to compute the final library\n")
	fmt.Fprintf(pErr, "\tthe runtime can be set to 300 sec to ensure the ultimate quality.\n")
	fmt.Fprintf(pErr, "\tIn some cases, the runtime has to be reduced if the supergate library\n")
	fmt.Fprintf(pErr, "\tcontains too many supergates (> 500K).\n")
	fmt.Fprintf(pErr, "\t\n")
	fmt.Fprintf(pErr, "\tWhen precomputing libraries of 6 inputs (-i 6), start with even more \n")
	fmt.Fprintf(pErr, "\trestricted parameters and gradually increase them until the goal is met.\n")
	fmt.Fprintf(pErr, "\t\n")
	return 1 /* error exit */
}

func superYesNo(b bool) string {
	if b {
		return "yes"
	}
	return "no"
}

// superAtoi is C atoi: optional blanks and sign, then the decimal prefix.
func superAtoi(s string) int32 {
	s = strings.TrimLeft(s, " \t\n\r\f\v")
	end := 0
	if end < len(s) && (s[end] == '+' || s[end] == '-') {
		end++
	}
	for end < len(s) && s[end] >= '0' && s[end] <= '9' {
		end++
	}
	n, _ := strconv.ParseInt(s[:end], 10, 64)
	return int32(n)
}

// superAtof is C atof on the longest decimal prefix.
func superAtof(s string) float64 {
	s = strings.TrimLeft(s, " \t\n\r\f\v")
	for end := len(s); end > 0; end-- {
		if v, err := strconv.ParseFloat(s[:end], 64); err == nil {
			return v
		}
	}
	return 0
}
