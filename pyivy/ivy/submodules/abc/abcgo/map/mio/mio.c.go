// Frame-owned MIO commands from mio.c. Mapper and Liberty integration use typed
// adapters registered by their owning packages, avoiding reverse import cycles.
package mio

import (
	"fmt"
	"io"
	"math"
	"os"
	"strconv"
	"strings"

	abcmain "github.com/glycerine/abcgo/base/main"
	"github.com/glycerine/abcgo/misc/vec"
)

type Abc_Frame_t = abcmain.Frame

func abcMioFrame() *Abc_Frame_t { return abcmain.Abc_FrameGetGlobalFrame() }

type LibertyReader interface {
	ReadMioLiberty(filename string, verbose bool) (*Mio_Library_t, error)
}
type GenlibPreparer interface {
	PrepareMioGenlib(name, text string) (any, error)
}

func mioFrameLibrary(f *Abc_Frame_t) *Mio_Library_t {
	lib, _ := f.Resource("pLibGen").(*Mio_Library_t)
	return lib
}
func mioInstall(f *Abc_Frame_t, lib *Mio_Library_t) {
	if old := mioFrameLibrary(f); old != nil && old != lib {
		abcmain.Abc_FrameUnmapAllNetworks(f)
	}
	f.SetResource("pLibSuper", nil)
	f.SetResource("pLibGen2", nil)
	f.SetResource("pLibGen", lib)
}
func Mio_UpdateGenlib(lib *Mio_Library_t) { mioInstall(abcmain.Abc_FrameGetGlobalFrame(), lib) }
func Mio_UpdateGenlib2(text, alternate *Vec_Str_t, name string, verbose int32) int32 {
	lib, err := ReadLibrary(name, strings.TrimRight(string(vec.Vec_StrArray(text)), "\x00"), nil)
	if err != nil {
		return 0
	}
	f := abcmain.Abc_FrameGetGlobalFrame()
	preparer, ok := f.Resource("mio.genlibPreparer").(GenlibPreparer)
	if !ok {
		fmt.Fprintln(f.Err, "MIO: secondary mapper genlib preparer is unavailable")
		return 0
	}
	secondary, err := preparer.PrepareMioGenlib(name, strings.TrimRight(string(vec.Vec_StrArray(alternate)), "\x00"))
	if err != nil {
		fmt.Fprintln(f.Err, err)
		return 0
	}
	mioInstall(f, lib)
	f.SetResource("pLibGen2", secondary)
	return 1
}
func Mio_Init(f *Abc_Frame_t) {
	commands := map[string]func(*Abc_Frame_t, int32, []string) int32{"read_genlib": Mio_CommandReadGenlib, "write_genlib": Mio_CommandWriteGenlib, "print_genlib": Mio_CommandPrintGenlib, "read_profile": Mio_CommandReadProfile, "write_profile": Mio_CommandWriteProfile, "print_profile": Mio_CommandPrintProfile, "read_library": Mio_CommandReadGenlib, "write_library": Mio_CommandWriteGenlib, "print_library": Mio_CommandPrintGenlib}
	for name, handler := range commands {
		h := handler
		if err := f.RegisterCommand(abcmain.Command{Name: name, Group: "SC mapping", Handler: func(frame *Abc_Frame_t, args []string) int32 { return h(frame, int32(len(args)), args) }}); err != nil {
			fmt.Fprintln(f.Err, err)
		}
	}
}
func Mio_End(f *Abc_Frame_t) { mioInstall(f, nil) }
func mioCommandArgs(argc int32, args []string, flags string) (map[byte]bool, map[byte]string, []string, error) {
	if argc < 1 || int(argc) > len(args) {
		return nil, nil, nil, fmt.Errorf("invalid argument count")
	}
	toggles := make(map[byte]bool)
	values := make(map[byte]string)
	var files []string
	for i := 1; i < int(argc); i++ {
		arg := args[i]
		if arg == "--" {
			files = append(files, args[i+1:argc]...)
			break
		}
		if len(arg) < 2 || arg[0] != '-' {
			files = append(files, arg)
			continue
		}
		for j := 1; j < len(arg); j++ {
			flag := arg[j]
			if !strings.ContainsRune(flags, rune(flag)) || flag == 'h' {
				return nil, nil, nil, fmt.Errorf("unknown or help option -%c", flag)
			}
			if flag == 'W' || flag == 'E' {
				if j+1 < len(arg) {
					values[flag] = arg[j+1:]
				} else {
					i++
					if i >= int(argc) {
						return nil, nil, nil, fmt.Errorf("option -%c requires a value", flag)
					}
					values[flag] = args[i]
				}
				break
			}
			toggles[flag] = !toggles[flag]
		}
	}
	return toggles, values, files, nil
}
func mioCommandError(f *Abc_Frame_t, err error) int32 { fmt.Fprintln(f.Err, err); return 1 }
func Mio_CommandReadGenlib(f *Abc_Frame_t, argc int32, args []string) int32 {
	flags, values, files, err := mioCommandArgs(argc, args, "WEnvh")
	if err != nil || len(files) != 1 {
		return mioCommandError(f, fmt.Errorf("usage: read_genlib [-nv] [-W delay] [-E excludes] file"))
	}
	excludes := make(map[string]bool)
	if path := values['E']; path != "" {
		if Mio_LibraryReadExclude(path, excludes) < 0 {
			return mioCommandError(f, fmt.Errorf("cannot read exclude file %s", path))
		}
	}
	shift := float64(0)
	if value := values['W']; value != "" {
		shift, err = strconv.ParseFloat(value, 64)
		if err != nil || shift <= 0 || math.IsNaN(shift) || math.IsInf(shift, 0) {
			return mioCommandError(f, fmt.Errorf("wire delay must be positive"))
		}
	}
	lib, err := ReadLibraryFile(files[0], excludes)
	if err != nil {
		return mioCommandError(f, err)
	}
	if flags['n'] {
		Mio_LibraryShortNames(lib)
	}
	if shift != 0 {
		Mio_LibraryShiftDelay(lib, shift)
	}
	var secondary any
	if preparer, ok := f.Resource("mio.genlibPreparer").(GenlibPreparer); ok {
		data, readErr := os.ReadFile(files[0])
		if readErr != nil {
			return mioCommandError(f, readErr)
		}
		secondary, err = preparer.PrepareMioGenlib(files[0], string(data))
		if err != nil {
			return mioCommandError(f, err)
		}
	}
	mioInstall(f, lib)
	if secondary != nil {
		f.SetResource("pLibGen2", secondary)
	}
	return 0
}
func Mio_CommandReadLiberty(f *Abc_Frame_t, argc int32, args []string) int32 {
	flags, _, files, err := mioCommandArgs(argc, args, "vh")
	if err != nil || len(files) != 1 {
		return mioCommandError(f, fmt.Errorf("usage: read_liberty [-v] file"))
	}
	reader, ok := f.Resource("mio.libertyReader").(LibertyReader)
	if !ok {
		return mioCommandError(f, fmt.Errorf("Liberty reader adapter is unavailable"))
	}
	lib, err := reader.ReadMioLiberty(files[0], flags['v'])
	if err != nil {
		return mioCommandError(f, err)
	}
	mioInstall(f, lib)
	return 0
}
func mioWriteCommand(f *Abc_Frame_t, argc int32, args []string, profile, print bool) int32 {
	allowed := "agvh"
	if print {
		allowed = "savh"
	}
	if profile {
		allowed = "h"
	}
	flags, _, files, err := mioCommandArgs(argc, args, allowed)
	if err != nil || print && len(files) != 0 || !print && len(files) != 1 {
		return mioCommandError(f, fmt.Errorf("invalid library/profile command arguments"))
	}
	lib := mioFrameLibrary(f)
	if lib == nil {
		return mioCommandError(f, fmt.Errorf("mapping library is not available"))
	}
	var writer io.Writer = f.Out
	var file *os.File
	if !print {
		file, err = os.Create(files[0])
		if err != nil {
			return mioCommandError(f, err)
		}
		writer = file
	}
	if profile {
		err = Mio_LibraryWriteProfile(writer, lib)
	} else if flags['g'] {
		err = Mio_WriteLibraryVerilog(writer, lib, 0, 0, mioBool(flags['a']))
	} else {
		err = Mio_WriteLibrary(writer, lib, 0, mioBool(flags['s']), mioBool(flags['a']))
	}
	if file != nil {
		if closeErr := file.Close(); err == nil {
			err = closeErr
		}
	}
	if err != nil {
		return mioCommandError(f, err)
	}
	return 0
}
func Mio_CommandWriteGenlib(f *Abc_Frame_t, n int32, a []string) int32 {
	return mioWriteCommand(f, n, a, false, false)
}
func Mio_CommandPrintGenlib(f *Abc_Frame_t, n int32, a []string) int32 {
	return mioWriteCommand(f, n, a, false, true)
}
func Mio_CommandWriteProfile(f *Abc_Frame_t, n int32, a []string) int32 {
	return mioWriteCommand(f, n, a, true, false)
}
func Mio_CommandPrintProfile(f *Abc_Frame_t, n int32, a []string) int32 {
	return mioWriteCommand(f, n, a, true, true)
}
func Mio_CommandReadProfile(f *Abc_Frame_t, n int32, a []string) int32 {
	_, _, files, err := mioCommandArgs(n, a, "h")
	if err != nil || len(files) != 1 {
		return mioCommandError(f, fmt.Errorf("usage: read_profile file"))
	}
	lib := mioFrameLibrary(f)
	if lib == nil {
		return mioCommandError(f, fmt.Errorf("mapping library is not available"))
	}
	file, err := os.Open(files[0])
	if err != nil {
		return mioCommandError(f, err)
	}
	defer file.Close()
	if err = Mio_LibraryReadProfile(file, lib); err != nil {
		return mioCommandError(f, err)
	}
	return 0
}
