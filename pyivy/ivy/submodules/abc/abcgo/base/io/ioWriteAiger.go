// AIGER codecs and network serialization adapters.
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
	"github.com/glycerine/abcgo/misc/util"
)

// Abc_Cex_t is the CEX representation retained by ABC's utility package.
type Abc_Cex_t = util.Abc_Cex_t

// Io_ObjMakeLit constructs an AIGER literal from a variable and polarity.
func Io_ObjMakeLit(variable, complement int32) uint32 {
	if variable < 0 || uint64(variable) > uint64(^uint32(0)>>1) {
		panic("AIGER variable is out of range")
	}
	return uint32(variable)<<1 | uint32(complement)&1
}

// Io_ObjAigerNum reads the temporary AIGER number used by legacy callers.
func Io_ObjAigerNum(object *Abc_Obj_t) uint32 {
	if object == nil || object.Temp < 0 {
		panic("AIGER number is not assigned")
	}
	return uint32(object.Temp)
}

// Io_ObjSetAigerNum stores a temporary AIGER number for legacy callers.
func Io_ObjSetAigerNum(object *Abc_Obj_t, number uint32) {
	if object == nil || uint64(number) > uint64(^uint32(0)>>1) {
		panic("invalid AIGER object number")
	}
	object.Temp = int(number)
}

// Io_WriteAigerEncode appends one variable-length unsigned integer to buffer.
func Io_WriteAigerEncode(buffer []byte, position int32, value uint32) int32 {
	if position < 0 {
		panic("negative AIGER buffer position")
	}
	var encoded [5]byte
	n := ioa.Ioa_WriteAigerEncode(encoded[:], 0, value)
	if int64(position)+int64(n) > int64(len(buffer)) {
		panic("short AIGER output buffer")
	}
	copy(buffer[position:], encoded[:n])
	return position + n
}

// Io_WriteAigerLiterals returns latch next-state and output literals in file order.
func Io_WriteAigerLiterals(network *abc.Network) ([]uint32, error) {
	model, err := aigerModelFromNetwork(network)
	if err != nil {
		return nil, err
	}
	result := make([]uint32, 0, len(model.Latches)+len(model.Outputs)+len(model.Bad)+len(model.Constraints))
	for _, latch := range model.Latches {
		result = append(result, latch.Next)
	}
	result = append(result, model.Outputs...)
	result = append(result, model.Bad...)
	result = append(result, model.Constraints...)
	return result, nil
}

// Io_WriteEncodeLiterals encodes the compact AIGER literal-delta list.
func Io_WriteEncodeLiterals(literals []uint32) ([]byte, error) {
	return ioa.Ioa_WriteEncodeLiterals(literals)
}

// WriteAiger serializes an ABC network to AIGER. The extension selects ASCII
// for .aag; other supported names use binary AIGER. Symbols are included.
func WriteAiger(network *abc.Network, filename string) error {
	ascii := strings.HasSuffix(strings.ToLower(filename), ".aag")
	return writeAiger(network, filename, ioa.WriteOptions{ASCII: ascii, Symbols: true}, "")
}

// WriteAigerWithOptions exposes ASCII/compact/symbol options without mutating
// network scratch fields.
func WriteAigerWithOptions(network *abc.Network, filename string, options ioa.WriteOptions) error {
	compression := ""
	lower := strings.ToLower(filename)
	if strings.HasSuffix(lower, ".gz") {
		compression = "gzip"
	} else if strings.HasSuffix(lower, ".bz2") {
		compression = "bzip2"
	}
	return writeAiger(network, filename, options, compression)
}

func writeAiger(network *abc.Network, filename string, options ioa.WriteOptions, compression string) error {
	model, err := aigerModelFromNetwork(network)
	if err != nil {
		return err
	}
	data, err := ioa.Marshal(model, options)
	if err != nil {
		return err
	}
	switch compression {
	case "gzip":
		var encoded bytes.Buffer
		writer := gzip.NewWriter(&encoded)
		if _, err = writer.Write(data); err == nil {
			err = writer.Close()
		} else {
			_ = writer.Close()
		}
		if err != nil {
			return err
		}
		data = encoded.Bytes()
	case "bzip2":
		data, err = bzlib.CompressBytes(data, 9)
		if err != nil {
			return err
		}
	}
	return os.WriteFile(filename, data, 0644)
}

// Io_WriteAiger keeps ABC's flag-based writer API and reports errors to stderr.
func Io_WriteAiger(network *Abc_Ntk_t, filename string, writeSymbols, compact, unique int32) {
	options := ioa.WriteOptions{Symbols: writeSymbols != 0, Compact: compact != 0}
	if unique != 0 {
		options.Symbols, options.Compact = false, false
	}
	if err := writeAiger(network, filename, options, aigerCompression(filename)); err != nil {
		fmt.Fprintln(os.Stderr, err)
	}
}

// Io_WriteAiger_old is the pre-unique source-compatible writer entry point.
func Io_WriteAiger_old(network *Abc_Ntk_t, filename string, writeSymbols, compact int32) {
	Io_WriteAiger(network, filename, writeSymbols, compact, 0)
}

// Io_WriteAigerGz emits binary AIGER through gzip compression.
func Io_WriteAigerGz(network *Abc_Ntk_t, filename string, writeSymbols int32) {
	if err := writeAiger(network, filename, ioa.WriteOptions{Symbols: writeSymbols != 0}, "gzip"); err != nil {
		fmt.Fprintln(os.Stderr, err)
	}
}

func aigerCompression(filename string) string {
	lower := strings.ToLower(filename)
	if strings.HasSuffix(lower, ".gz") {
		return "gzip"
	}
	if strings.HasSuffix(lower, ".bz2") {
		return "bzip2"
	}
	return ""
}

func aigerModelFromNetwork(network *abc.Network) (*ioa.Model, error) {
	if network == nil {
		return nil, fmt.Errorf("nil ABC network")
	}
	if err := network.Check(); err != nil {
		return nil, fmt.Errorf("invalid ABC network: %w", err)
	}
	model := &ioa.Model{Name: network.Name, Comment: []byte(network.Spec), Symbols: make(map[ioa.Symbol]string)}
	metadata, _ := network.Data.(*aigerNetworkMetadata)
	if metadata != nil {
		model.Justice = cloneAigerLists(metadata.justice)
		model.Fairness = append([]uint32(nil), metadata.fairness...)
	}
	literalByObject := map[*abc.Object]uint32{network.ObjectsOfKind(abc.ObjectConst1)[0]: 1}
	nextVariable := uint32(1)
	for index, input := range network.ObjectsOfKind(abc.ObjectPI) {
		literal := nextVariable << 1
		nextVariable++
		literalByObject[input] = literal
		model.Inputs = append(model.Inputs, literal)
		if name := input.Name(); name != "" {
			model.Symbols[ioa.Symbol{Kind: 'i', Index: uint32(index)}] = name
		}
	}
	latches := network.ObjectsOfKind(abc.ObjectLatch)
	for index, latch := range latches {
		if len(latch.Fanins()) != 1 || latch.Fanins()[0].Object.Kind != abc.ObjectBI || len(latch.Fanouts()) != 1 || latch.Fanouts()[0].Kind != abc.ObjectBO {
			return nil, fmt.Errorf("latch %d lacks BI/BO terminals", index)
		}
		bo := latch.Fanouts()[0]
		literal := nextVariable << 1
		nextVariable++
		literalByObject[bo] = literal
		init := uint32(0)
		switch latch.Data {
		case abc.ABC_INIT_ZERO:
			init = 0
		case abc.ABC_INIT_ONE:
			init = 1
		case abc.ABC_INIT_DC, abc.ABC_INIT_NONE, nil:
			init = literal
		default:
			return nil, fmt.Errorf("latch %d has unsupported reset %v", index, latch.Data)
		}
		model.Latches = append(model.Latches, ioa.Latch{Literal: literal, Init: init})
		if name := bo.Name(); name != "" {
			model.Symbols[ioa.Symbol{Kind: 'l', Index: uint32(index)}] = name
		} else if name := latch.Name(); name != "" {
			model.Symbols[ioa.Symbol{Kind: 'l', Index: uint32(index)}] = name
		}
	}
	active := make(map[*abc.Object]bool)
	var literalForEdge func(abc.Edge) (uint32, error)
	var literalForObject func(*abc.Object) (uint32, error)
	literalForEdge = func(edge abc.Edge) (uint32, error) {
		literal, err := literalForObject(edge.Object)
		if edge.Complement {
			literal ^= 1
		}
		return literal, err
	}
	literalForObject = func(object *abc.Object) (uint32, error) {
		if literal, ok := literalByObject[object]; ok {
			return literal, nil
		}
		if object == nil || object.Network() != network {
			return 0, fmt.Errorf("foreign or nil AIGER object")
		}
		if active[object] {
			return 0, fmt.Errorf("combinational cycle at ABC object %d", object.ID)
		}
		active[object] = true
		defer delete(active, object)
		switch object.Kind {
		case abc.ObjectNet, abc.ObjectBI:
			fanins := object.Fanins()
			if len(fanins) != 1 {
				return 0, fmt.Errorf("wire object %d does not have one driver", object.ID)
			}
			return literalForEdge(fanins[0])
		case abc.ObjectLatch:
			fanouts := object.Fanouts()
			if len(fanouts) != 1 || fanouts[0].Kind != abc.ObjectBO {
				return 0, fmt.Errorf("latch %d lacks output terminal", object.ID)
			}
			return literalForObject(fanouts[0])
		case abc.ObjectNode:
			if object.Data != "abcgo:AIG AND" {
				return 0, fmt.Errorf("object %d is not a structurally hashed AIG AND", object.ID)
			}
			fanins := object.Fanins()
			if len(fanins) != 2 {
				return 0, fmt.Errorf("AIG AND %d has %d inputs", object.ID, len(fanins))
			}
			left, err := literalForEdge(fanins[0])
			if err != nil {
				return 0, err
			}
			right, err := literalForEdge(fanins[1])
			if err != nil {
				return 0, err
			}
			literal := nextVariable << 1
			nextVariable++
			literalByObject[object] = literal
			if left < right {
				left, right = right, left
			}
			model.ANDs = append(model.ANDs, ioa.And{Literal: literal, Left: left, Right: right})
			return literal, nil
		case abc.ObjectConst1:
			return 1, nil
		default:
			return 0, fmt.Errorf("object %d of kind %d is not an AIG signal", object.ID, object.Kind)
		}
	}
	// Include dangling ANDs as well as nodes reachable from outputs.
	for _, node := range network.ObjectsOfKind(abc.ObjectNode) {
		if _, err := literalForObject(node); err != nil {
			return nil, err
		}
	}
	for i, latch := range latches {
		input := latch.Fanins()[0].Object
		next, err := literalForObject(input)
		if err != nil {
			return nil, err
		}
		model.Latches[i].Next = next
	}
	allOutputs := make([]uint32, 0, len(network.ObjectsOfKind(abc.ObjectPO)))
	outputs := network.ObjectsOfKind(abc.ObjectPO)
	for _, po := range outputs {
		fanins := po.Fanins()
		if len(fanins) != 1 {
			return nil, fmt.Errorf("PO %d does not have one driver", po.ID)
		}
		literal, err := literalForEdge(fanins[0])
		if err != nil {
			return nil, err
		}
		allOutputs = append(allOutputs, literal)
	}
	outputCount, badCount, constraintCount := len(allOutputs), 0, 0
	if metadata != nil && (metadata.badCount != 0 || metadata.constraintCount != 0) {
		if metadata.outputCount != len(allOutputs) {
			return nil, fmt.Errorf("AIGER extended-output metadata no longer matches network POs")
		}
		badCount, constraintCount = metadata.badCount, metadata.constraintCount
		outputCount -= badCount + constraintCount
	}
	model.Outputs = append(model.Outputs, allOutputs[:outputCount]...)
	if badCount != 0 {
		model.Bad = append(model.Bad, allOutputs[outputCount:outputCount+badCount]...)
	}
	if constraintCount != 0 {
		model.Constraints = append(model.Constraints, allOutputs[outputCount+badCount:]...)
	}
	for index, po := range outputs {
		if name := po.Name(); name != "" {
			kind, symbolIndex := byte('o'), index
			if index >= outputCount && index < outputCount+badCount {
				kind, symbolIndex = 'b', index-outputCount
			} else if index >= outputCount+badCount {
				kind, symbolIndex = 'c', index-outputCount-badCount
			}
			model.Symbols[ioa.Symbol{Kind: kind, Index: uint32(symbolIndex)}] = name
		}
	}
	if metadata != nil {
		for symbol, name := range metadata.symbols {
			if symbol.Kind == 'j' || symbol.Kind == 'f' {
				model.Symbols[symbol] = name
			}
		}
	}
	model.MaxVar = nextVariable - 1
	if err := model.Check(); err != nil {
		return nil, fmt.Errorf("invalid generated AIGER model: %w", err)
	}
	return model, nil
}

// Io_WriteAigerCex writes the ABC trace text format associated with a CEX.
func Io_WriteAigerCex(cex *Abc_Cex_t, network *Abc_Ntk_t, provider any, filename string) {
	if err := WriteAigerCounterexample(cex, network, provider, filename); err != nil {
		fmt.Fprintln(os.Stderr, err)
	}
}

// WriteAigerCounterexample writes the input/output/register trace for a CEX.
func WriteAigerCounterexample(cex *Abc_Cex_t, network *abc.Network, provider any, filename string) error {
	if cex == nil || cex.IFrame < 0 || cex.NPis < 0 || cex.NRegs < 0 {
		return fmt.Errorf("invalid counterexample")
	}
	var model *ioa.Model
	var err error
	if network != nil {
		model, err = aigerModelFromNetwork(network)
	} else if source, ok := provider.(ioa.AigerProvider); ok {
		model, err = source.AigerModel()
	} else {
		return fmt.Errorf("AIGER counterexample needs a network or AigerProvider")
	}
	if err != nil {
		return err
	}
	if int(cex.NPis) != len(model.Inputs) || int(cex.NRegs) != len(model.Latches) {
		return fmt.Errorf("AIG parameters do not match those of the CEX")
	}
	needed := int64(cex.NRegs) + int64(cex.IFrame+1)*int64(cex.NPis)
	if needed > int64(cex.NBits) || int64(len(cex.PData))*32 < needed {
		return fmt.Errorf("counterexample data is truncated")
	}
	file, err := os.Create(filename)
	if err != nil {
		return err
	}
	writeErr := writeCexTrace(file, model, cex)
	closeErr := file.Close()
	if writeErr != nil {
		return writeErr
	}
	return closeErr
}

func writeCexTrace(w io.Writer, model *ioa.Model, cex *Abc_Cex_t) error {
	if _, err := fmt.Fprintln(w, "1"); err != nil {
		return err
	}
	state := make([]bool, len(model.Latches))
	for range state {
		if _, err := io.WriteString(w, "0"); err != nil {
			return err
		}
	}
	if _, err := io.WriteString(w, " "); err != nil {
		return err
	}
	for frame := int32(0); frame <= cex.IFrame; frame++ {
		// ABC's trace format repeats the current register state at the start
		// of every frame after the first. The initial state was emitted above.
		if frame != 0 {
			for _, stateBit := range state {
				if _, err := io.WriteString(w, bitChar(stateBit)); err != nil {
					return err
				}
			}
			if _, err := io.WriteString(w, " "); err != nil {
				return err
			}
		}
		inputs := make([]bool, len(model.Inputs))
		for i := range inputs {
			bit := cex.NRegs + frame*cex.NPis + int32(i)
			inputs[i] = util.Abc_InfoHasBit(cex.PData, bit)
			if _, err := io.WriteString(w, bitChar(inputs[i])); err != nil {
				return err
			}
		}
		values, err := evaluateAigerModel(model, inputs, state)
		if err != nil {
			return err
		}
		if _, err = io.WriteString(w, " "); err != nil {
			return err
		}
		for _, output := range append(append(append([]uint32(nil), model.Outputs...), model.Bad...), model.Constraints...) {
			if _, err = io.WriteString(w, bitChar(aigerLiteralValue(values, output))); err != nil {
				return err
			}
		}
		next := make([]bool, len(model.Latches))
		if _, err = io.WriteString(w, " "); err != nil {
			return err
		}
		for i, latch := range model.Latches {
			next[i] = aigerLiteralValue(values, latch.Next)
			if _, err = io.WriteString(w, bitChar(next[i])); err != nil {
				return err
			}
		}
		if _, err = io.WriteString(w, "\n"); err != nil {
			return err
		}
		state = next
	}
	return nil
}

func evaluateAigerModel(model *ioa.Model, inputs, state []bool) (map[uint32]bool, error) {
	if err := model.Check(); err != nil {
		return nil, err
	}
	values := map[uint32]bool{0: true}
	for i, literal := range model.Inputs {
		values[literal>>1] = inputs[i]
	}
	for i, latch := range model.Latches {
		values[latch.Literal>>1] = state[i]
	}
	for _, gate := range model.ANDs {
		left, leftOK := values[gate.Left>>1]
		right, rightOK := values[gate.Right>>1]
		if !leftOK || !rightOK {
			return nil, fmt.Errorf("AIGER gate %d is not topologically ordered", gate.Literal)
		}
		if gate.Left&1 != 0 {
			left = !left
		}
		if gate.Right&1 != 0 {
			right = !right
		}
		values[gate.Literal>>1] = left && right
	}
	return values, nil
}

func aigerLiteralValue(values map[uint32]bool, literal uint32) bool {
	value := values[literal>>1]
	if literal&1 != 0 {
		return !value
	}
	return value
}

func bitChar(value bool) string {
	if value {
		return "1"
	}
	return "0"
}
