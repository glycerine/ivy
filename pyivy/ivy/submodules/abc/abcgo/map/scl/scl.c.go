package scl

import (
	"fmt"
	"path/filepath"
	"strconv"
	"strings"

	abcmain "github.com/glycerine/abcgo/base/main"
	"github.com/glycerine/abcgo/map/mio"
	"github.com/glycerine/abcgo/misc/util"
)

func (p *SC_Lib) Close() error { Abc_SclLibFree(p); return nil }
func Abc_SclLoad(lib *SC_Lib, current **SC_Lib) {
	if *current != nil && *current != lib {
		Abc_SclLibFree(*current)
	}
	*current = lib
}
func Scl_ConGetMan(f *abcmain.Frame) *Scl_Con_t       { p, _ := f.Resource("pAbcCon").(*Scl_Con_t); return p }
func Scl_ConFreeMan(f *abcmain.Frame)                 { f.SetResource("pAbcCon", nil) }
func Scl_ConUpdateMan(f *abcmain.Frame, p *Scl_Con_t) { f.SetResource("pAbcCon", p) }
func Scl_End(f *abcmain.Frame)                        { f.SetResource("pLibScl", nil); Scl_ConFreeMan(f) }
func Scl_Init(f *abcmain.Frame) {
	registrations := []struct {
		name    string
		handler func(*abcmain.Frame, int32, []string) int32
		changes bool
	}{
		{"read_lib", Scl_CommandReadLib, false},
		{"write_lib", Scl_CommandWriteLib, false},
		{"print_lib", Scl_CommandPrintLib, false},
		{"leak2area", Scl_CommandLeak2Area, false},
		{"read_scl", Scl_CommandReadScl, false},
		{"write_scl", Scl_CommandWriteScl, false},
		{"dump_genlib", Scl_CommandDumpGen, false},
		{"print_gs", Scl_CommandPrintGS, false},
		{"stime", Scl_CommandStime, false},
		{"minsize", Scl_CommandMinsize, true},
		{"maxsize", Scl_CommandMaxsize, true},
		{"dnsize", Scl_CommandDnsize, true},
		{"print_buf", Scl_CommandPrintBuf, false},
		{"read_constr", Scl_CommandReadConstr, false},
		{"write_constr", Scl_CommandWriteConstr, false},
		{"print_constr", Scl_CommandPrintConstr, false},
		{"reset_constr", Scl_CommandResetConstr, false},
	}
	for _, registration := range registrations {
		handler := registration.handler
		_ = f.RegisterCommand(abcmain.Command{Name: registration.name, Group: "SCL mapping", ChangesNetwork: registration.changes, Handler: func(f *abcmain.Frame, args []string) int32 { return handler(f, int32(len(args)), args) }})
	}
	f.SetResource("mio.libertyReader", LibertyMIOReader{})
}

type sclOptions struct {
	flags   map[byte]bool
	numbers map[byte]float32
	args    []string
}

func parseSclOptions(argv []string, toggles, numeric string) (sclOptions, error) {
	o := sclOptions{flags: map[byte]bool{}, numbers: map[byte]float32{}}
	for i := 1; i < len(argv); i++ {
		arg := argv[i]
		if arg == "--" {
			o.args = append(o.args, argv[i+1:]...)
			break
		}
		if len(arg) < 2 || arg[0] != '-' {
			o.args = append(o.args, arg)
			continue
		}
		for j := 1; j < len(arg); j++ {
			key := arg[j]
			if key == 'h' {
				return o, fmt.Errorf("help requested")
			}
			if strings.IndexByte(toggles, key) >= 0 {
				o.flags[key] = !o.flags[key]
				continue
			}
			if strings.IndexByte(numeric, key) < 0 {
				return o, fmt.Errorf("unknown option -%c", key)
			}
			value := arg[j+1:]
			if value == "" {
				i++
				if i == len(argv) {
					return o, fmt.Errorf("option -%c requires a number", key)
				}
				value = argv[i]
			}
			v, e := strconv.ParseFloat(value, 32)
			if e != nil || v < 0 || v != v || v > 3.402823466e38 || strings.IndexByte("SGAB", key) >= 0 && v == 0 {
				return o, fmt.Errorf("invalid value for -%c: %s", key, value)
			}
			o.numbers[key] = float32(v)
			break
		}
	}
	return o, nil
}
func sclCommandError(f *abcmain.Frame, e error) int32 { fmt.Fprintln(f.Err, e); return 1 }
func sclLibrary(f *abcmain.Frame) (*SC_Lib, error) {
	p, _ := f.Resource("pLibScl").(*SC_Lib)
	if p == nil {
		return nil, fmt.Errorf("there is no Liberty library available")
	}
	return p, nil
}
func installFrameGenlib(f *abcmain.Frame, p *SC_Lib, slew, gain float32, minimum int) error {
	lib, e := Abc_SclDeriveGenlib(p, nil, slew, gain, minimum, 1)
	if e != nil {
		return e
	}
	for g := mio.Mio_LibraryReadGates(lib); g != nil; g = mio.Mio_GateReadNext(g) {
		if mio.Mio_GateReadPinNum(g) == 0 {
			continue
		}
		id, ok := p.MioCellID(mio.Mio_GateReadName(g))
		if !ok {
			return fmt.Errorf("cannot find cell %s", mio.Mio_GateReadName(g))
		}
		mio.Mio_GateSetCell(g, id)
	}
	f.SetResource("pLibGen", lib)
	return nil
}
func Scl_CommandReadLib(f *abcmain.Frame, argc int32, argv []string) int32 {
	o, e := parseSclOptions(argv, "dnvw", "SGM")
	if e != nil || len(o.args) != 1 {
		return sclCommandError(f, fmt.Errorf("usage: read_lib [-SG float] [-M num] [-dnvwh] <file>"))
	}
	p, e := ReadLiberty(o.args[0])
	if e != nil {
		return sclCommandError(f, e)
	}
	if len(p.CellClasses) < 3 {
		fmt.Fprintf(f.Err, "Library with only %d cell classes cannot be used.\n", len(p.CellClasses))
		return 0
	}
	if o.flags['n'] {
		Abc_SclShortNames(p)
	}
	if o.flags['d'] {
		path := strings.TrimSuffix(o.args[0], filepath.Ext(o.args[0])) + "_temp.lib"
		if e = Abc_SclWriteLiberty(path, p); e != nil {
			return sclCommandError(f, e)
		}
	}
	if e = installFrameGenlib(f, p, o.numbers['S'], o.numbers['G'], int(o.numbers['M'])); e != nil {
		return sclCommandError(f, e)
	}
	f.SetResource("pLibScl", p)
	return 0
}
func Scl_CommandReadScl(f *abcmain.Frame, argc int32, argv []string) int32 {
	o, e := parseSclOptions(argv, "d", "")
	if e != nil || len(o.args) != 1 {
		return sclCommandError(f, fmt.Errorf("usage: read_scl [-dh] <file>"))
	}
	p, e := Abc_SclReadFromFile(o.args[0])
	if e != nil {
		return sclCommandError(f, e)
	}
	if o.flags['d'] {
		if e = Abc_SclWriteLiberty(strings.TrimSuffix(o.args[0], filepath.Ext(o.args[0]))+"_temp.lib", p); e != nil {
			return sclCommandError(f, e)
		}
	}
	if e = installFrameGenlib(f, p, 0, 0, 0); e != nil {
		return sclCommandError(f, e)
	}
	f.SetResource("pLibScl", p)
	return 0
}
func writeLibraryCommand(f *abcmain.Frame, argv []string, binary bool) int32 {
	o, e := parseSclOptions(argv, "", "")
	if e != nil || len(o.args) != 1 {
		return sclCommandError(f, fmt.Errorf("usage: %s <file>", argv[0]))
	}
	p, e := sclLibrary(f)
	if e != nil {
		return sclCommandError(f, e)
	}
	if binary {
		e = Abc_SclWriteScl(o.args[0], p)
	} else {
		e = Abc_SclWriteLiberty(o.args[0], p)
	}
	if e != nil {
		return sclCommandError(f, e)
	}
	return 0
}
func Scl_CommandWriteLib(f *abcmain.Frame, argc int32, argv []string) int32 {
	return writeLibraryCommand(f, argv, false)
}
func Scl_CommandWriteScl(f *abcmain.Frame, argc int32, argv []string) int32 {
	return writeLibraryCommand(f, argv, true)
}
func Scl_CommandPrintLib(f *abcmain.Frame, argc int32, argv []string) int32 {
	o, e := parseSclOptions(argv, "is", "SG")
	if e != nil {
		return sclCommandError(f, fmt.Errorf("usage: print_lib [-SG float] [-ish]"))
	}
	p, e := sclLibrary(f)
	if e != nil {
		return sclCommandError(f, e)
	}
	if e = PrintCells(f.Out, p, o.numbers['S'], o.numbers['G'], o.flags['i'], o.flags['s']); e != nil {
		return sclCommandError(f, e)
	}
	return 0
}
func Scl_CommandLeak2Area(f *abcmain.Frame, argc int32, argv []string) int32 {
	o, e := parseSclOptions(argv, "v", "AB")
	if e != nil {
		return sclCommandError(f, fmt.Errorf("usage: leak2area [-AB float] [-vh]"))
	}
	p, e := sclLibrary(f)
	if e != nil {
		return sclCommandError(f, e)
	}
	a, b := float32(1), float32(1)
	if v, ok := o.numbers['A']; ok {
		a = v
	}
	if v, ok := o.numbers['B']; ok {
		b = v
	}
	Abc_SclConvertLeakageIntoArea(p, a, b)
	return 0
}
func Scl_CommandDumpGen(f *abcmain.Frame, argc int32, argv []string) int32 {
	o, e := parseSclOptions(argv, "v", "SGM")
	if e != nil || len(o.args) > 1 {
		return sclCommandError(f, fmt.Errorf("usage: dump_genlib [-SG float] [-M num] [-vh] [file]"))
	}
	p, e := sclLibrary(f)
	if e != nil {
		return sclCommandError(f, e)
	}
	path := p.Name + ".genlib"
	if len(o.args) == 1 {
		path = o.args[0]
	}
	if e = Abc_SclDumpGenlib(path, p, o.numbers['S'], o.numbers['G'], int(o.numbers['M'])); e != nil {
		return sclCommandError(f, e)
	}
	return 0
}
func Scl_CommandWriteConstr(f *abcmain.Frame, argc int32, argv []string) int32 {
	o, e := parseSclOptions(argv, "v", "")
	if e != nil || len(o.args) > 1 {
		return sclCommandError(f, fmt.Errorf("usage: write_constr [-vh] [file]"))
	}
	p := Scl_ConGetMan(f)
	if p == nil {
		fmt.Fprintln(f.Out, "Scl_CommandWriteConstr(): There is no constraint manager.")
		return 0
	}
	path := strings.TrimSuffix(p.FileName, filepath.Ext(p.FileName)) + "_out.constr"
	if len(o.args) == 1 {
		path = o.args[0]
	}
	if filepath.Ext(path) != ".constr" {
		return sclCommandError(f, fmt.Errorf("unrecognized output file extension"))
	}
	if e = Scl_ConWrite(p, path); e != nil {
		return sclCommandError(f, e)
	}
	return 0
}
func Scl_CommandPrintConstr(f *abcmain.Frame, argc int32, argv []string) int32 {
	_, e := parseSclOptions(argv, "v", "")
	if e != nil {
		return sclCommandError(f, fmt.Errorf("usage: print_constr [-vh]"))
	}
	p := Scl_ConGetMan(f)
	if p == nil {
		fmt.Fprintln(f.Out, "There is no constraint manager.")
		return 0
	}
	if e = WriteConstraints(f.Out, p); e != nil {
		return sclCommandError(f, e)
	}
	return 0
}
func Scl_CommandResetConstr(f *abcmain.Frame, argc int32, argv []string) int32 {
	_, e := parseSclOptions(argv, "v", "")
	if e != nil {
		return sclCommandError(f, fmt.Errorf("usage: reset_constr [-vh]"))
	}
	f.SetStringValue("pDrivingCell", "")
	f.MaxLoad = 0
	Scl_ConFreeMan(f)
	return 0
}

// ConstraintNetwork supplies real port names; numeric IDs are never invented as
// net names when interpreting a user constraint file.
type ConstraintNetwork interface {
	SclInputs() []int
	SclOutputs() []int
	SclObjectName(int) string
}

func Scl_CommandReadConstr(f *abcmain.Frame, argc int32, argv []string) int32 {
	o, e := parseSclOptions(argv, "nv", "")
	if e != nil || len(o.args) != 1 {
		return sclCommandError(f, fmt.Errorf("usage: read_constr [-nvh] <file>"))
	}
	verbose := int32(0)
	if o.flags['v'] {
		verbose = 1
	}
	if !o.flags['n'] {
		if e = Abc_SclReadTimingConstr(f, o.args[0], verbose); e != nil {
			return sclCommandError(f, e)
		}
		return 0
	}
	net, ok := f.Network().(ConstraintNetwork)
	if !ok {
		return sclCommandError(f, fmt.Errorf("current network does not provide named ports for constraints"))
	}
	names := func(ids []int) *util.Abc_Nam_t {
		p := util.Abc_NamStart(int32(len(ids)), 16)
		for _, id := range ids {
			util.Abc_NamStrFindOrAdd(p, net.SclObjectName(id), nil)
		}
		return p
	}
	p, e := Scl_ConRead(o.args[0], names(net.SclInputs()), names(net.SclOutputs()))
	if e != nil {
		return sclCommandError(f, e)
	}
	Scl_ConUpdateMan(f, p)
	return 0
}

func sclMappedNetwork(f *abcmain.Frame) (*SC_Lib, TimingNetwork, error) {
	lib, e := sclLibrary(f)
	if e != nil {
		return nil, nil, e
	}
	net, ok := f.Network().(TimingNetwork)
	if !ok {
		return nil, nil, fmt.Errorf("current network does not provide standard-cell topology/mapping")
	}
	if e = ensureSclMapping(lib, net); e != nil {
		return nil, nil, e
	}
	return lib, net, nil
}
func Scl_CommandStime(f *abcmain.Frame, argc int32, argv []string) int32 {
	o, e := parseSclOptions(argv, "capd", "X")
	if e != nil || len(o.args) != 0 {
		return sclCommandError(f, fmt.Errorf("usage: stime [-X num] [-capdh]"))
	}
	lib, net, e := sclMappedNetwork(f)
	if e != nil {
		return sclCommandError(f, e)
	}
	if e = Abc_SclTimePerform(lib, net, int32(o.numbers['X']), boolInt(o.flags['c']), boolInt(o.flags['a']), boolInt(o.flags['p']), boolInt(o.flags['d'])); e != nil {
		return sclCommandError(f, e)
	}
	return 0
}
func Scl_CommandPrintGS(f *abcmain.Frame, argc int32, argv []string) int32 {
	_, e := parseSclOptions(argv, "", "")
	if e != nil {
		return sclCommandError(f, fmt.Errorf("usage: print_gs [-h]"))
	}
	lib, net, e := sclMappedNetwork(f)
	if e != nil {
		return sclCommandError(f, e)
	}
	if e = Abc_SclPrintGateSizes(lib, net); e != nil {
		return sclCommandError(f, e)
	}
	return 0
}
func Scl_CommandPrintBuf(f *abcmain.Frame, argc int32, argv []string) int32 {
	o, e := parseSclOptions(argv, "v", "")
	if e != nil {
		return sclCommandError(f, fmt.Errorf("usage: print_buf [-vh]"))
	}
	lib, net, e := sclMappedNetwork(f)
	if e != nil {
		return sclCommandError(f, e)
	}
	if !Abc_SclHasDelayInfo(lib) {
		return sclCommandError(f, fmt.Errorf("library delay info is not available"))
	}
	if e = Abc_SclPrintBuffers(lib, net, boolInt(o.flags['v'])); e != nil {
		return sclCommandError(f, e)
	}
	return 0
}
func sizeExtremeCommand(f *abcmain.Frame, argv []string, maximum bool) int32 {
	o, e := parseSclOptions(argv, "v", "")
	if e != nil || len(o.args) != 0 {
		return sclCommandError(f, fmt.Errorf("usage: %s [-vh]", argv[0]))
	}
	lib, net, e := sclMappedNetwork(f)
	if e != nil {
		return sclCommandError(f, e)
	}
	if e = Abc_SclMinsizePerform(lib, net, boolInt(maximum), boolInt(o.flags['v'])); e != nil {
		return sclCommandError(f, e)
	}
	return 0
}
func Scl_CommandMinsize(f *abcmain.Frame, argc int32, argv []string) int32 {
	return sizeExtremeCommand(f, argv, false)
}
func Scl_CommandMaxsize(f *abcmain.Frame, argc int32, argv []string) int32 {
	return sizeExtremeCommand(f, argv, true)
}
func Scl_CommandDnsize(f *abcmain.Frame, argc int32, argv []string) int32 {
	o, e := parseSclOptions(argv, "csdvw", "IJNDGTX")
	if e != nil || len(o.args) != 0 {
		return sclCommandError(f, fmt.Errorf("usage: dnsize [-IJNDGTX num] [-csdvwh]"))
	}
	pars := SC_SizePars{Iterations: 5, IterationsNoChange: 50, Notches: 1000, DelayGap: 1000, UseDepartures: !o.flags['s'], UseWireLoads: o.flags['c'], DumpStats: o.flags['d'], Verbose: o.flags['v'], VeryVerbose: o.flags['w']}
	for key, dst := range map[byte]*int{'I': &pars.Iterations, 'J': &pars.IterationsNoChange, 'N': &pars.Notches, 'D': &pars.DelayUser, 'G': &pars.DelayGap, 'T': &pars.Timeout, 'X': &pars.BufferTreeEstimate} {
		if v, ok := o.numbers[key]; ok {
			if v != float32(int(v)) {
				return sclCommandError(f, fmt.Errorf("-%c requires an integer", key))
			}
			*dst = int(v)
		}
	}
	lib, net, e := sclMappedNetwork(f)
	if e != nil {
		return sclCommandError(f, e)
	}
	p, e := Abc_SclManStart(lib, net, boolInt(pars.UseWireLoads), boolInt(pars.UseDepartures), float32(pars.DelayUser), int32(pars.BufferTreeEstimate))
	if e != nil {
		return sclCommandError(f, e)
	}
	defer Abc_SclManFree(p)
	p.Output = f.Out
	if e = DownsizeManager(p, &pars); e != nil {
		return sclCommandError(f, e)
	}
	return 0
}
