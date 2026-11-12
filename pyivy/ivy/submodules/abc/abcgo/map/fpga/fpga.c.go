package fpga

import (
	"fmt"
	abcmain "github.com/glycerine/abcgo/base/main"
)

func Fpga_Init(f *abcmain.Frame) {
	if f.Resource("pLibLut") == nil {
		f.SetResource("pLibLut", DefaultLibrary(4))
	}
	for _, entry := range []struct {
		name    string
		handler func(*abcmain.Frame, int32, []string) int32
	}{{"read_lut", Fpga_CommandReadLibrary}, {"print_lut", Fpga_CommandPrintLibrary}} {
		h := entry.handler
		_ = f.RegisterCommand(abcmain.Command{Name: entry.name, Group: "FPGA mapping", Handler: func(f *abcmain.Frame, args []string) int32 { return h(f, int32(len(args)), args) }})
	}
}
func Fpga_End(f *abcmain.Frame) {
	f.RemoveCommand("read_lut")
	f.RemoveCommand("print_lut")
	if lib, ok := f.TakeResource("pLibLut").(*Fpga_LutLib_t); ok {
		_ = lib.Close()
	}
}
func Fpga_CommandReadLibrary(f *abcmain.Frame, argc int32, args []string) int32 {
	verbose := int32(1)
	var path string
	for _, arg := range args[1:] {
		if arg == "-v" {
			verbose = 1 - verbose
		} else if len(arg) > 0 && arg[0] == '-' || path != "" {
			fmt.Fprintln(f.Err, "usage: read_lut [-v] file")
			return 1
		} else {
			path = arg
		}
	}
	if path == "" {
		fmt.Fprintln(f.Err, "usage: read_lut [-v] file")
		return 1
	}
	p := Fpga_LutLibRead(path, verbose)
	if p == nil {
		fmt.Fprintf(f.Err, "Cannot read LUT library %q.\n", path)
		return 1
	}
	f.SetResource("pLibLut", p)
	return 0
}
func Fpga_CommandPrintLibrary(f *abcmain.Frame, argc int32, args []string) int32 {
	if len(args) != 1 {
		fmt.Fprintln(f.Err, "usage: print_lut")
		return 1
	}
	p, ok := f.Resource("pLibLut").(*Fpga_LutLib_t)
	if !ok || p == nil {
		fmt.Fprintln(f.Err, "No LUT library.")
		return 1
	}
	if e := p.Write(f.Out); e != nil {
		fmt.Fprintln(f.Err, e)
		return 1
	}
	return 0
}
func Fpga_SetSimpleLutLib(size int32) {
	if size < 1 || size > FPGA_MAX_LUTSIZE {
		panic("fpga: invalid LUT size")
	}
	abcmain.Abc_FrameGetGlobalFrame().SetResource("pLibLut", DefaultLibrary(int(size)))
}
