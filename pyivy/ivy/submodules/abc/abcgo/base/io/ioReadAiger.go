// AIGER readers and conversion between AIGER models and ABC networks.
package io

import (
	"bytes"
	"compress/gzip"
	"fmt"
	"io"
	"os"
	"strings"

	"github.com/glycerine/abcgo/aig/ioa"
	"github.com/glycerine/abcgo/base/abc"
	"github.com/glycerine/abcgo/misc/bzlib"
	"github.com/glycerine/abcgo/misc/extra"
	"github.com/glycerine/abcgo/misc/zlib"
)

// aigerNetworkMetadata retains AIGER-only properties that ABC's generic
// network shape cannot represent directly.
type aigerNetworkMetadata struct {
	comment                   string
	badCount, constraintCount int
	outputCount               int
	fairness                  []uint32
	justice                   [][]uint32
	symbols                   map[ioa.Symbol]string
}

// ReadAiger reads standard AIGER binary/ASCII and ABC compact AIGER files.
func ReadAiger(filename string, _ bool) (*abc.Network, error) {
	data, err := readAigerFile(filename)
	if err != nil {
		return nil, err
	}
	model, err := ioa.Read(data)
	if err != nil {
		return nil, fmt.Errorf("%s: %w", filename, err)
	}
	network, err := networkFromCheckedAigerModel(model)
	if err != nil {
		return nil, fmt.Errorf("%s: %w", filename, err)
	}
	if network.Name == "" {
		network.Name = extra.Extra_FileNameGeneric(filename)
	}
	network.Spec = filename
	if len(model.Symbols) == 0 {
		abc.Abc_NtkShortNames(network)
	}
	return network, nil
}

// ReadAigerFromMemory decodes an AIGER byte slice into an ABC network.
func ReadAigerFromMemory(data []byte, _ bool) (*abc.Network, error) {
	model, err := ioa.Read(data)
	if err != nil {
		return nil, err
	}
	return networkFromCheckedAigerModel(model)
}

// Io_ReadAiger preserves ABC's nil-on-error entry point. New callers should use
// ReadAiger to receive parse and I/O errors.
func Io_ReadAiger(filename string, check int32) *Abc_Ntk_t {
	network, err := ReadAiger(filename, check != 0)
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		return nil
	}
	return network
}

func init() { RegisterFormat(IO_FILE_AIGER, ReadAiger, WriteAiger) }

func readAigerFile(filename string) ([]byte, error) {
	data, err := os.ReadFile(filename)
	if err != nil {
		return nil, err
	}
	if bytes.HasPrefix(data, []byte("BZh")) || strings.HasSuffix(strings.ToLower(filename), ".bz2") {
		decoded, decodeErr := bzlib.DecompressBytes(data)
		if decodeErr != nil {
			return nil, fmt.Errorf("decompress bzip2 AIGER: %w", decodeErr)
		}
		return decoded, nil
	}
	if len(data) >= 2 && data[0] == 0x1f && data[1] == 0x8b || strings.HasSuffix(strings.ToLower(filename), ".gz") {
		reader, openErr := gzip.NewReader(bytes.NewReader(data))
		if openErr != nil {
			return nil, fmt.Errorf("open gzip AIGER: %w", openErr)
		}
		decoded, readErr := io.ReadAll(reader)
		closeErr := reader.Close()
		if readErr != nil {
			return nil, readErr
		}
		if closeErr != nil {
			return nil, closeErr
		}
		return decoded, nil
	}
	return data, nil
}

// Ioa_ReadLoadFileBz2Aig keeps the historical string-based loader. A plain
// file is accepted as a passthrough, matching bzlib's magic-header behavior.
func Ioa_ReadLoadFileBz2Aig(filename string, size *int32) string {
	data, err := os.ReadFile(filename)
	if err != nil {
		return ""
	}
	if bytes.HasPrefix(data, []byte("BZh")) {
		data, err = bzlib.DecompressBytes(data)
		if err != nil {
			return ""
		}
	}
	if size != nil {
		if len(data) > int(^uint32(0)>>1) {
			return ""
		}
		*size = int32(len(data))
	}
	return string(data)
}

// Ioa_ReadLoadFileGzAig loads gzip data, while preserving the old passthrough
// behavior for uncompressed files.
func Ioa_ReadLoadFileGzAig(filename string, size *int32) string {
	data, err := os.ReadFile(filename)
	if err != nil {
		return ""
	}
	if len(data) >= 2 && data[0] == 0x1f && data[1] == 0x8b {
		reader, openErr := gzip.NewReader(bytes.NewReader(data))
		if openErr != nil {
			return ""
		}
		data, err = io.ReadAll(reader)
		closeErr := reader.Close()
		if err != nil || closeErr != nil {
			return ""
		}
	}
	if size != nil {
		if len(data) > int(^uint32(0)>>1) {
			return ""
		}
		*size = int32(len(data))
	}
	return string(data)
}

// Ioa_ReadLoadFileZlibAig retains the old generic zlib-file loading helper.
func Ioa_ReadLoadFileZlibAig(filename string, size *int32) string {
	file, err := zlib.Open(filename, "rb")
	if err != nil {
		return ""
	}
	data, err := io.ReadAll(file)
	closeErr := file.Close()
	if err != nil || closeErr != nil {
		return ""
	}
	if size != nil {
		if len(data) > int(^uint32(0)>>1) {
			return ""
		}
		*size = int32(len(data))
	}
	return string(data)
}

// Io_ReadAigerDecode decodes one AIGER unsigned integer and advances ppPos.
// Malformed input panics for compatibility; parser code should use ioa.Read.
func Io_ReadAigerDecode(ppPos *string) uint32 {
	if ppPos == nil {
		panic("nil AIGER cursor")
	}
	data := []byte(*ppPos)
	value, err := ioa.Ioa_ReadAigerDecode(&data)
	if err != nil {
		panic(err)
	}
	*ppPos = (*ppPos)[len(*ppPos)-len(data):]
	return value
}

// Io_WriteDecodeLiterals decodes compact literal deltas and advances ppPos.
func Io_WriteDecodeLiterals(ppPos *string, entries int32) []uint32 {
	if ppPos == nil {
		panic("nil AIGER cursor")
	}
	data := []byte(*ppPos)
	values, err := ioa.Ioa_WriteDecodeLiterals(&data, entries)
	if err != nil {
		panic(err)
	}
	*ppPos = (*ppPos)[len(*ppPos)-len(data):]
	return values
}

// NetworkFromAigerModel converts a validated AIGER model into an ABC AIG
// network. This is also the shared adapter for other native readers.
func NetworkFromAigerModel(model *ioa.Model) (*abc.Network, error) {
	if err := model.Check(); err != nil {
		return nil, err
	}
	return networkFromCheckedAigerModel(model)
}

// ioa.Read and the public model converter validate before reaching this point.
func networkFromCheckedAigerModel(model *ioa.Model) (*abc.Network, error) {
	outputCount := len(model.Outputs) + len(model.Bad) + len(model.Constraints)
	network := abc.NewAIGNetworkWithCapacity(model.Name, len(model.Inputs), len(model.Latches), outputCount, len(model.ANDs))
	network.Type, network.Function = uint8(abc.ABC_NTK_STRASH), uint8(abc.ABC_FUNC_AIG)
	// Like Abc_NtkAlloc(ABC_NTK_STRASH, ...), a strashed network owns an AIG manager.
	network.FunctionManager = abc.Abc_AigAlloc(network)
	network.Spec = string(model.Comment)
	// As Io_ReadAiger: the last outputs are the constraints.
	if n := len(model.Constraints); n == 1 {
		fmt.Printf("Warning: The last output is interpreted as a constraint.\n")
	} else if n > 1 {
		fmt.Printf("Warning: The last %d outputs are interpreted as constraints.\n", n)
	}
	network.Constraints = int32(len(model.Constraints))
	network.Data = &aigerNetworkMetadata{
		comment:  string(model.Comment),
		badCount: len(model.Bad), constraintCount: len(model.Constraints),
		outputCount: outputCount,
		fairness:    append([]uint32(nil), model.Fairness...), justice: cloneAigerLists(model.Justice),
		symbols: cloneAigerSymbols(model.Symbols),
	}
	inputObjects := make([]*abc.Object, len(model.Inputs))
	for i := range model.Inputs {
		pi := network.NewObject(abc.ObjectPI)
		inputObjects[i] = pi
	}
	// Like Io_ReadAiger: PIs, then all outputs, then latches, then ANDs.
	outputObjects := make([]*abc.Object, outputCount)
	for i := range outputObjects {
		outputObjects[i] = network.NewObject(abc.ObjectPO)
	}
	latchBoxes := make([]*abc.Object, len(model.Latches))
	latchInputs := make([]*abc.Object, len(model.Latches))
	latchOutputs := make([]*abc.Object, len(model.Latches))
	for i, latchModel := range model.Latches {
		latch, bi, bo := network.NewObject(abc.ObjectLatch), network.NewObject(abc.ObjectBI), network.NewObject(abc.ObjectBO)
		latchBoxes[i], latchInputs[i], latchOutputs[i] = latch, bi, bo
		if err := network.AddFanin(latch, bi.Signal()); err != nil {
			return nil, err
		}
		if err := network.AddFanin(bo, latch.Signal()); err != nil {
			return nil, err
		}
		switch latchModel.Init {
		case 0:
			latch.Data = abc.ABC_INIT_ZERO
		case 1:
			latch.Data = abc.ABC_INIT_ONE
		default:
			latch.Data = abc.ABC_INIT_DC
		}
	}
	edgeForLiteral, err := aigerGateEdges(network, model, inputObjects, latchOutputs)
	if err != nil {
		return nil, err
	}
	for i, latchModel := range model.Latches {
		next, err := edgeForLiteral(latchModel.Next)
		if err != nil {
			return nil, err
		}
		if err = network.AddFanin(latchInputs[i], next); err != nil {
			return nil, err
		}
	}
	outputLits := make([]uint32, 0, len(model.Outputs)+len(model.Bad)+len(model.Constraints))
	outputLits = append(outputLits, model.Outputs...)
	outputLits = append(outputLits, model.Bad...)
	outputLits = append(outputLits, model.Constraints...)
	for i, literal := range outputLits {
		edge, err := edgeForLiteral(literal)
		if err != nil {
			return nil, err
		}
		if err = network.AddFanin(outputObjects[i], edge); err != nil {
			return nil, err
		}
	}
	for symbol, name := range model.Symbols {
		if name == "" {
			continue
		}
		switch symbol.Kind {
		case 'i':
			inputObjects[symbol.Index].SetName(name)
		case 'l':
			latchBoxes[symbol.Index].SetName(name)
			latchOutputs[symbol.Index].SetName(name)
		case 'o':
			outputObjects[symbol.Index].SetName(name)
		case 'b':
			outputObjects[len(model.Outputs)+int(symbol.Index)].SetName(name)
		case 'c':
			outputObjects[len(model.Outputs)+len(model.Bad)+int(symbol.Index)].SetName(name)
		}
	}
	// as Io_ReadAiger: ABC keeps constraints inverted relative to the file
	if len(model.Bad) != 0 || len(model.Constraints) != 0 || len(model.Justice) != 0 || len(model.Fairness) != 0 {
		abc.Abc_NtkInvertConstraints(network)
	}
	if err := network.Check(); err != nil {
		return nil, fmt.Errorf("invalid ABC network after AIGER read: %w", err)
	}
	return network, nil
}

func cloneAigerLists(in [][]uint32) [][]uint32 {
	if in == nil {
		return nil
	}
	out := make([][]uint32, len(in))
	for i := range in {
		out[i] = append([]uint32(nil), in[i]...)
	}
	return out
}

func cloneAigerSymbols(in map[ioa.Symbol]string) map[ioa.Symbol]string {
	out := make(map[ioa.Symbol]string, len(in))
	for symbol, name := range in {
		out[symbol] = name
	}
	return out
}

func aigerGateEdges(network *abc.Network, model *ioa.Model, inputs, latches []*abc.Object) (func(uint32) (abc.Edge, error), error) {
	if model.HasDenseTopologicalOrder() {
		edges := make([]abc.Edge, int(model.MaxVar)+1)
		edges[0] = network.One().Not()
		for i, pi := range inputs {
			edges[1+i] = pi.Signal()
		}
		for i, bo := range latches {
			edges[1+len(inputs)+i] = bo.Signal()
		}
		edgeAt := func(lit uint32) abc.Edge {
			edge := edges[lit>>1]
			if lit&1 != 0 {
				return edge.Not()
			}
			return edge
		}
		for _, gate := range model.ANDs {
			edges[gate.Literal>>1] = network.AndEdges(edgeAt(gate.Left), edgeAt(gate.Right))
		}
		return func(lit uint32) (abc.Edge, error) { return edgeAt(lit), nil }, nil
	}
	literalEdges := map[uint32]abc.Edge{0: network.One().Not()}
	for i, lit := range model.Inputs {
		literalEdges[lit>>1] = inputs[i].Signal()
	}
	for i, latch := range model.Latches {
		literalEdges[latch.Literal>>1] = latches[i].Signal()
	}
	gates := make(map[uint32]ioa.And, len(model.ANDs))
	for _, gate := range model.ANDs {
		gates[gate.Literal>>1] = gate
	}
	active := make(map[uint32]bool)
	var edgeForLiteral func(uint32) (abc.Edge, error)
	edgeForLiteral = func(literal uint32) (abc.Edge, error) {
		variable := literal >> 1
		edge, ok := literalEdges[variable]
		if !ok {
			gate, found := gates[variable]
			if !found {
				return abc.Edge{}, fmt.Errorf("undefined AIGER variable %d", variable)
			}
			if active[variable] {
				return abc.Edge{}, fmt.Errorf("AIGER combinational cycle at variable %d", variable)
			}
			active[variable] = true
			left, err := edgeForLiteral(gate.Left)
			if err != nil {
				return abc.Edge{}, err
			}
			right, err := edgeForLiteral(gate.Right)
			if err != nil {
				return abc.Edge{}, err
			}
			edge = network.AndEdges(left, right)
			delete(active, variable)
			literalEdges[variable] = edge
		}
		if literal&1 != 0 {
			edge = edge.Not()
		}
		return edge, nil
	}
	for _, gate := range model.ANDs {
		if _, err := edgeForLiteral(gate.Literal); err != nil {
			return nil, err
		}
	}
	return edgeForLiteral, nil
}
