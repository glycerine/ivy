package mapper

import (
	"fmt"
	"os"

	abcmain "github.com/glycerine/abcgo/base/main"
	"github.com/glycerine/abcgo/map/mio"
	"github.com/glycerine/abcgo/misc/extra"
)

func Map_Init(frame *abcmain.Frame) {
	_ = frame.RegisterCommand(abcmain.Command{
		Name: "read_super", Group: "SC mapping",
		Handler: func(frame *abcmain.Frame, args []string) int32 {
			return Map_CommandReadLibrary(frame, int32(len(args)), args)
		},
	})
}

func Map_End(frame *abcmain.Frame) {
	frame.RemoveCommand("read_super")
	if lib, ok := frame.TakeResource("pLibSuper").(*Map_SuperLib_t); ok {
		Map_SuperLibFree(lib)
	}
}

func Map_CommandReadLibrary(frame *abcmain.Frame, _ int32, args []string) int32 {
	var exclude string
	verbose, algorithm := int32(1), int32(1)
	path := ""
	for i := 1; i < len(args); i++ {
		arg := args[i]
		switch arg {
		case "-e":
			i++
			if i >= len(args) {
				return mapReadSuperUsage(frame)
			}
			exclude = args[i]
		case "-o":
			algorithm ^= 1
		case "-v":
			verbose ^= 1
		case "-h":
			return mapReadSuperUsage(frame)
		default:
			if len(arg) > 0 && arg[0] == '-' || path != "" {
				return mapReadSuperUsage(frame)
			}
			path = arg
		}
	}
	if path == "" {
		return mapReadSuperUsage(frame)
	}

	f, err := os.Open(path)
	if err != nil {
		fmt.Fprintf(frame.Err, "Cannot open input file %q. ", path)
		if guess := extra.Extra_FileGetSimilarName(path, ".genlib", ".lib", ".gen", ".g", ""); guess != "" {
			fmt.Fprintf(frame.Err, "Did you mean %q?", guess)
		}
		fmt.Fprintln(frame.Err)
		return 1
	}
	_ = f.Close()

	genlib, ok := frame.Resource("pLibGen").(*mio.Mio_Library_t)
	if !ok || genlib == nil {
		fmt.Fprintln(frame.Err, "Genlib library should be read in first.")
		return 1
	}
	lib := Map_SuperLibCreate(genlib, nil, path, exclude, algorithm, verbose)
	if lib == nil {
		fmt.Fprintln(frame.Err, "Reading supergate library has failed.")
		return 1
	}
	if old, ok := frame.TakeResource("pLibSuper").(*Map_SuperLib_t); ok {
		Map_SuperLibFree(old)
	}
	frame.SetResource("pLibSuper", lib)
	return 0
}

func mapReadSuperUsage(frame *abcmain.Frame) int32 {
	fmt.Fprintln(frame.Err, "\nusage: read_super [-e file] [-ovh] file")
	fmt.Fprintln(frame.Err, "\tread the supergate library from the file")
	fmt.Fprintln(frame.Err, "\t-e file : file contains list of genlib gates to exclude")
	fmt.Fprintln(frame.Err, "\t-o      : toggles the old file format")
	fmt.Fprintln(frame.Err, "\t-v      : toggles verbose output")
	fmt.Fprintln(frame.Err, "\t-h      : print this usage")
	return 1
}
