// Native format dispatch and netlist-construction helpers.
package io

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/glycerine/abcgo/base/abc"
)

func Io_ReadFileType(filename string) Io_FileType_t {
	lowered := strings.ToLower(filename)
	for _, suffix := range []string{".gz", ".bz2"} {
		if strings.HasSuffix(lowered, suffix) {
			filename = filename[:len(filename)-len(suffix)]
			break
		}
	}
	switch strings.ToLower(filepath.Ext(filename)) {
	case ".aig", ".aiger", ".aag":
		return IO_FILE_AIGER
	case ".baf":
		return IO_FILE_BAF
	case ".bblif":
		return IO_FILE_BBLIF
	case ".blif":
		return IO_FILE_BLIF
	case ".mv", ".blifmv":
		return IO_FILE_BLIFMV
	case ".bench":
		return IO_FILE_BENCH
	case ".book":
		return IO_FILE_BOOK
	case ".cnf":
		return IO_FILE_CNF
	case ".dot":
		return IO_FILE_DOT
	case ".edif", ".edf":
		return IO_FILE_EDIF
	case ".eqn":
		return IO_FILE_EQN
	case ".gml":
		return IO_FILE_GML
	case ".json":
		return IO_FILE_JSON
	case ".list":
		return IO_FILE_LIST
	case ".pla":
		return IO_FILE_PLA
	case ".mopla":
		return IO_FILE_MOPLA
	case ".smv":
		return IO_FILE_SMV
	case ".v", ".sv":
		return IO_FILE_VERILOG
	default:
		return IO_FILE_UNKNOWN
	}
}
func ReadNetwork(filename string, kind Io_FileType_t, check bool) (*abc.Network, error) {
	if kind == IO_FILE_NONE || kind == IO_FILE_UNKNOWN {
		kind = Io_ReadFileType(filename)
	}
	formats.RLock()
	reader := formats.handlers[kind].read
	formats.RUnlock()
	if reader == nil {
		return nil, fmt.Errorf("no reader for file type %d (%s)", kind, filename)
	}
	return reader(filename, check)
}
func WriteNetwork(n *abc.Network, filename string, kind Io_FileType_t) error {
	if n == nil {
		return fmt.Errorf("nil network")
	}
	if kind == IO_FILE_NONE || kind == IO_FILE_UNKNOWN {
		kind = Io_ReadFileType(filename)
	}
	formats.RLock()
	writer := formats.handlers[kind].write
	formats.RUnlock()
	if writer == nil {
		return fmt.Errorf("no writer for file type %d (%s)", kind, filename)
	}
	return writer(n, filename)
}
func Io_ReadNetlist(filename string, kind Io_FileType_t, check int32) *abc.Network {
	n, err := ReadNetwork(filename, kind, check != 0)
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		return nil
	}
	return n
}

// ReadLogic follows ABC's generic reader pipeline: preserve raw netlists in
// ReadNetwork, but flatten/convert them when a logic network is requested.
func ReadLogic(filename string, kind Io_FileType_t, check, barBuffers bool) (*abc.Network, error) {
	n, err := ReadNetwork(filename, kind, check)
	if err != nil {
		return nil, err
	}
	if n.Type != uint8(abc.ABC_NTK_NETLIST) {
		return n, nil
	}
	if barBuffers {
		n = abc.Abc_NtkToBarBufs(n)
		if n == nil {
			return nil, fmt.Errorf("%s: barrier-buffer conversion failed", filename)
		}
		return n, nil
	}
	if len(n.ObjectsOfKind(abc.ObjectWhitebox)) > 0 {
		n = abc.Abc_NtkFlattenLogicHierarchy(n)
		if n == nil {
			return nil, fmt.Errorf("%s: hierarchy flattening failed", filename)
		}
	}
	if len(n.ObjectsOfKind(abc.ObjectBlackbox)) > 0 {
		n = abc.Abc_NtkConvertBlackboxes(n)
		if n == nil {
			return nil, fmt.Errorf("%s: blackbox conversion failed", filename)
		}
	}
	if n.Function == uint8(abc.ABC_FUNC_BLIFMV) {
		n = abc.Abc_NtkStrashBlifMv(n)
	} else {
		n = abc.Abc_NtkToLogic(n)
	}
	if n == nil {
		return nil, fmt.Errorf("%s: netlist conversion failed", filename)
	}
	return n, nil
}
func Io_Read(filename string, kind Io_FileType_t, check, barBuffers int32) *abc.Network {
	n, err := ReadLogic(filename, kind, check != 0, barBuffers != 0)
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		return nil
	}
	return n
}
func Io_Write(n *abc.Network, filename string, kind Io_FileType_t) error {
	return WriteNetwork(n, filename, kind)
}
func Io_WriteHie(n *abc.Network, baseName, filename string) error {
	if n == nil || abc.Abc_NtkIsStrash(n) == 0 && abc.Abc_NtkIsLogic(n) == 0 {
		return fmt.Errorf("io: hierarchy output requires a logic or strashed network")
	}
	kind := Io_ReadFileType(baseName)
	if kind != IO_FILE_BLIF && kind != IO_FILE_BLIFMV && kind != IO_FILE_VERILOG {
		return fmt.Errorf("io: unknown hierarchy input format")
	}
	base, err := ReadNetwork(baseName, kind, true)
	if err != nil {
		return err
	}
	if len(base.ObjectsOfKind(abc.ObjectWhitebox)) != 0 && n.Barriers == 0 {
		old := base
		base = abc.Abc_NtkFlattenLogicHierarchy(old)
		abc.Abc_NtkDelete(old)
		if base == nil {
			return fmt.Errorf("io: cannot flatten base hierarchy")
		}
	}
	defer abc.Abc_NtkDelete(base)
	flat := abc.Abc_NtkToNetlist(n)
	if flat == nil {
		return fmt.Errorf("io: cannot convert network to netlist")
	}
	var result *abc.Network
	switch {
	case n.Barriers > 0:
		result = abc.Abc_NtkFromBarBufs(base, flat)
	case kind == IO_FILE_BLIFMV:
		if len(base.ObjectsOfKind(abc.ObjectBlackbox)) != 0 {
			abc.Abc_NtkDelete(flat)
			return fmt.Errorf("io: BLIF-MV hierarchy with blackboxes is unsupported")
		}
		if abc.Abc_NtkConvertToBlifMv(flat) == 0 {
			abc.Abc_NtkDelete(flat)
			return fmt.Errorf("io: cannot convert network to BLIF-MV")
		}
		result = abc.Abc_NtkInsertBlifMv(base, flat)
	case len(base.ObjectsOfKind(abc.ObjectBlackbox)) > 0:
		result = abc.Abc_NtkInsertNewLogic(base, flat)
	default:
		result = flat
	}
	if result != flat {
		abc.Abc_NtkDelete(flat)
	}
	if result == nil {
		return fmt.Errorf("io: cannot reinsert hierarchy")
	}
	defer abc.Abc_NtkDelete(result)
	modules := []*abc.Network{result}
	if result.Design != nil {
		modules = result.Design.Modules
	}
	switch Io_ReadFileType(filename) {
	case IO_FILE_BLIF:
		for _, module := range modules {
			if abc.Abc_NtkHasSop(module) == 0 && abc.Abc_NtkHasMapping(module) == 0 && abc.Abc_NtkToSop(module, -1, 1_000_000_000) == 0 {
				return fmt.Errorf("io: cannot convert module %q to SOP", module.Name)
			}
		}
		return Io_WriteBlif(result, filename, 1, 0, 0)
	case IO_FILE_VERILOG:
		for _, module := range modules {
			if abc.Abc_NtkHasAig(module) == 0 && abc.Abc_NtkHasMapping(module) == 0 && abc.Abc_NtkToAig(module) == 0 {
				return fmt.Errorf("io: cannot convert module %q to AIG", module.Name)
			}
		}
		f, err := os.Create(filename)
		if err != nil {
			return err
		}
		err = WriteVerilog(f, result, 0)
		closeErr := f.Close()
		if err != nil {
			return err
		}
		return closeErr
	case IO_FILE_BLIFMV:
		return Io_WriteBlifMv(result, filename)
	default:
		return fmt.Errorf("io: unknown hierarchy output format")
	}
}
func ioBuilder(n *abc.Network) *blifModel {
	m := &blifModel{network: n, nets: map[string]*abc.Object{}, domains: map[string]*abc.MVVariable{}}
	for _, net := range n.ObjectsOfKind(abc.ObjectNet) {
		m.nets[net.Name()] = net
	}
	return m
}
func Io_ReadCreatePi(n *abc.Network, name string) *abc.Object {
	m := ioBuilder(n)
	if err := m.input(name); err != nil {
		panic(err)
	}
	return m.net(name).Fanins()[0].Object
}
func Io_ReadCreatePo(n *abc.Network, name string) *abc.Object {
	m := ioBuilder(n)
	if err := m.output(name); err != nil {
		panic(err)
	}
	ports := n.ObjectsOfKind(abc.ObjectPO)
	return ports[len(ports)-1]
}
func Io_ReadCreateLatch(n *abc.Network, input, output string) *abc.Object {
	if err := blifLatch(ioBuilder(n), []string{input, output}); err != nil {
		panic(err)
	}
	latches := n.ObjectsOfKind(abc.ObjectLatch)
	return latches[len(latches)-1]
}
func Io_ReadCreateNode(n *abc.Network, output string, inputs []string, count int32) *abc.Object {
	if count < 0 || int(count) > len(inputs) {
		panic("invalid node arity")
	}
	names := append(append([]string(nil), inputs[:count]...), output)
	node, err := ioBuilder(n).node(names, nil)
	if err != nil {
		panic(err)
	}
	return node
}
func Io_ReadCreateConst(n *abc.Network, name string, value int32) *abc.Object {
	cover := " 0\n"
	if value != 0 {
		cover = " 1\n"
	}
	node, err := ioBuilder(n).node([]string{name}, cover)
	if err != nil {
		panic(err)
	}
	return node
}
func Io_ReadCreateInv(n *abc.Network, input, output string) *abc.Object {
	node, err := ioBuilder(n).node([]string{input, output}, "0 1\n")
	if err != nil {
		panic(err)
	}
	return node
}
func Io_ReadCreateBuf(n *abc.Network, input, output string) *abc.Object {
	node, err := ioBuilder(n).node([]string{input, output}, "1 1\n")
	if err != nil {
		panic(err)
	}
	return node
}
func Io_FileOpen(filename, pathVariable, mode string, verbose int32) *os.File {
	candidates := []string{filename}
	if !filepath.IsAbs(filename) && pathVariable != "" {
		for _, directory := range filepath.SplitList(os.Getenv(pathVariable)) {
			candidates = append(candidates, filepath.Join(directory, filename))
		}
	}
	flags := os.O_RDONLY
	if strings.Contains(mode, "w") {
		flags = os.O_WRONLY | os.O_CREATE | os.O_TRUNC
	} else if strings.Contains(mode, "a") {
		flags = os.O_WRONLY | os.O_CREATE | os.O_APPEND
	}
	if strings.Contains(mode, "+") {
		flags = (flags &^ os.O_WRONLY) | os.O_RDWR
	}
	for _, candidate := range candidates {
		if file, err := os.OpenFile(candidate, flags, 0666); err == nil {
			return file
		}
	}
	if verbose != 0 {
		fmt.Fprintf(os.Stderr, "cannot open %s\n", filename)
	}
	return nil
}
func init() {
	RegisterFormat(IO_FILE_BLIF, func(f string, c bool) (*abc.Network, error) { return ReadBLIF(f, false, c) }, func(n *abc.Network, f string) error { return WriteBLIFFile(n, f, false, true) })
	RegisterFormat(IO_FILE_BLIFMV, func(f string, c bool) (*abc.Network, error) { return ReadBLIF(f, true, c) }, func(n *abc.Network, f string) error { return WriteBLIFFile(n, f, true, true) })
	RegisterFormat(IO_FILE_BENCH, ReadBench, nil)
}
