package abc

// AIGER I/O from src/aig/ioa/{ioa.h,ioaReadAig.c,ioaWriteAig.c,ioaUtil.c}.

import (
	"bytes"
	"io"
	"math"
	"os"
	"path/filepath"
	"strconv"
	"time"
)

// AigIOAOptions bounds parsing and serialization. Compact selects ABC's aig2
// driver encoding. Zero limits select conservative defaults.
type AigIOAOptions struct {
	Compact      bool
	Check        bool
	MaxBytes     int
	MaxVariables int
	MaxOutputs   int
	Stop         func() bool
	Timestamp    time.Time
}

func normalizeAigIOAOptions(options AigIOAOptions) (AigIOAOptions, error) {
	if options.MaxBytes < 0 || options.MaxVariables < 0 || options.MaxOutputs < 0 {
		return options, ErrDimension
	}
	if options.MaxBytes == 0 {
		options.MaxBytes = 32 << 20
	}
	if options.MaxVariables == 0 {
		options.MaxVariables = 1 << 20
	}
	if options.MaxOutputs == 0 {
		options.MaxOutputs = 1 << 20
	}
	if uint64(options.MaxVariables) > uint64(MaxLiteralID) {
		return options, ErrLimit
	}
	return options, nil
}

func aigIOAStop(options AigIOAOptions) error {
	if options.Stop != nil && options.Stop() {
		return ErrCancelled
	}
	return nil
}

// IOAEncodeAigerUint implements Ioa_WriteAigerEncode and
// Ioa_WriteAigerEncodeStr.
func IOAEncodeAigerUint(value uint32) []byte {
	result := make([]byte, 0, 5)
	for value >= 128 {
		result = append(result, byte(value)|0x80)
		value >>= 7
	}
	return append(result, byte(value))
}

// IOADecodeAigerUint implements Ioa_ReadAigerDecode with explicit bounds.
func IOADecodeAigerUint(data []byte, offset int) (uint32, int, error) {
	if offset < 0 || offset > len(data) {
		return 0, offset, ErrDimension
	}
	var value uint32
	for index := 0; index < 5; index++ {
		if offset == len(data) {
			return 0, offset, ErrFormat
		}
		entry := data[offset]
		offset++
		if index == 4 && entry > 15 {
			return 0, offset, ErrFormat
		}
		value |= uint32(entry&127) << uint(7*index)
		if entry < 128 {
			return value, offset, nil
		}
	}
	return 0, offset, ErrFormat
}

// IOAEncodeAigerLiterals preserves Ioa_WriteEncodeLiterals' first-literal
// encoding followed by signed adjacent differences.
func IOAEncodeAigerLiterals(literals []uint32, maxBytes int) ([]byte, error) {
	if maxBytes < 0 {
		return nil, ErrDimension
	}
	if maxBytes == 0 {
		maxBytes = 32 << 20
	}
	result := make([]byte, 0, min(maxBytes, len(literals)))
	previous := uint32(0)
	for index, literal := range literals {
		encoded := literal
		if index != 0 {
			var difference uint32
			if literal < previous {
				difference = previous - literal
				if difference > math.MaxUint32>>1 {
					return nil, ErrLimit
				}
				encoded = difference<<1 | 1
			} else {
				difference = literal - previous
				if difference > math.MaxUint32>>1 {
					return nil, ErrLimit
				}
				encoded = difference << 1
			}
		}
		piece := IOAEncodeAigerUint(encoded)
		if len(piece) > maxBytes-len(result) {
			return nil, ErrLimit
		}
		result = append(result, piece...)
		previous = literal
	}
	return result, nil
}

// IOADecodeAigerLiterals implements Ioa_WriteDecodeLiterals.
func IOADecodeAigerLiterals(data []byte, entries, maxEntries int) ([]uint32, int, error) {
	if entries < 0 || maxEntries < 0 {
		return nil, 0, ErrDimension
	}
	if maxEntries == 0 {
		maxEntries = 1 << 20
	}
	if entries > maxEntries {
		return nil, 0, ErrLimit
	}
	result := make([]uint32, entries)
	offset := 0
	var previous int64
	for index := range result {
		encoded, next, err := IOADecodeAigerUint(data, offset)
		if err != nil {
			return nil, offset, err
		}
		offset = next
		if index == 0 {
			previous = int64(encoded)
		} else {
			difference := int64(encoded >> 1)
			if encoded&1 != 0 {
				difference = -difference
			}
			previous += difference
		}
		if previous < 0 || previous > math.MaxUint32 {
			return nil, offset, ErrFormat
		}
		result[index] = uint32(previous)
	}
	return result, offset, nil
}

// IOAWriteAigerLiterals implements Ioa_WriteAigerLiterals in latch-input then
// primary-output order using the graph's current polarity. The caller does not
// need to pre-number graph objects and the graph is not mutated.
func IOAWriteAigerLiterals(graph *AigManager, options AigIOAOptions) ([]uint32, error) {
	options, err := normalizeAigIOAOptions(options)
	if err != nil {
		return nil, err
	}
	if graph == nil {
		return nil, ErrDimension
	}
	if err = graph.Check(); err != nil {
		return nil, err
	}
	if graph.PrimaryOutputCount() > options.MaxOutputs || uint64(graph.InputCount())+uint64(graph.AndCount()) > uint64(options.MaxVariables) {
		return nil, ErrLimit
	}
	if err = aigIOAStop(options); err != nil {
		return nil, err
	}
	raw := make([]uint32, len(graph.objects))
	raw[0] = 1
	for index, id := range graph.cis {
		raw[id] = 2 * uint32(index+1)
	}
	next := uint32(len(graph.cis) + 1)
	convert := func(literal AigLit) uint32 { return raw[literal.ID()] ^ uint32(literal&1) }
	for id, object := range graph.objects {
		if err = aigIOAStop(options); err != nil {
			return nil, err
		}
		switch object.Type {
		case AigBuffer:
			raw[id] = convert(object.Fanin0)
		case AigAND:
			raw[id] = 2 * next
			next++
		}
	}
	result := make([]uint32, 0, graph.OutputCount())
	for index := 0; index < graph.RegisterCount(); index++ {
		if err = aigIOAStop(options); err != nil {
			return nil, err
		}
		id := graph.cos[graph.PrimaryOutputCount()+index]
		result = append(result, convert(graph.objects[id].Fanin0))
	}
	for _, id := range graph.cos[:graph.PrimaryOutputCount()] {
		if err = aigIOAStop(options); err != nil {
			return nil, err
		}
		result = append(result, convert(graph.objects[id].Fanin0))
	}
	return result, nil
}

type aigIOACursor struct {
	data []byte
	pos  int
}

func (cursor *aigIOACursor) line() ([]byte, error) {
	if cursor.pos > len(cursor.data) {
		return nil, ErrFormat
	}
	end := bytes.IndexByte(cursor.data[cursor.pos:], '\n')
	if end < 0 {
		return nil, ErrFormat
	}
	start := cursor.pos
	cursor.pos += end + 1
	return cursor.data[start : start+end], nil
}

func aigIOADecimal(data []byte, maximum uint32) (uint32, error) {
	if len(data) == 0 {
		return 0, ErrFormat
	}
	var value uint32
	for _, entry := range data {
		if entry < '0' || entry > '9' {
			return 0, ErrFormat
		}
		digit := uint32(entry - '0')
		if digit > maximum || value > (maximum-digit)/10 {
			return 0, ErrFormat
		}
		value = value*10 + digit
	}
	return value, nil
}

func (cursor *aigIOACursor) variable() (uint32, error) {
	value, next, err := IOADecodeAigerUint(cursor.data, cursor.pos)
	if err != nil {
		return 0, err
	}
	cursor.pos = next
	return value, nil
}

// IOAReadAigerFromMemory reads standard binary AIGER and ABC's compact aig2
// variant. Input bytes remain owned by the caller and are never retained.
func IOAReadAigerFromMemory(data []byte, options AigIOAOptions) (*AigManager, error) {
	options, err := normalizeAigIOAOptions(options)
	if err != nil {
		return nil, err
	}
	if len(data) > options.MaxBytes {
		return nil, ErrLimit
	}
	if err = aigIOAStop(options); err != nil {
		return nil, err
	}
	cursor := aigIOACursor{data: data}
	header, err := cursor.line()
	if err != nil {
		return nil, err
	}
	fields := bytes.SplitN(header, []byte{' '}, 11)
	if len(fields) < 6 || len(fields) > 10 {
		return nil, ErrFormat
	}
	compact := bytes.Equal(fields[0], []byte("aig2"))
	if !compact && !bytes.Equal(fields[0], []byte("aig")) {
		return nil, ErrUnsupported
	}
	var counts [9]uint32
	for index := 1; index < len(fields); index++ {
		counts[index-1], err = aigIOADecimal(fields[index], math.MaxInt32)
		if err != nil {
			return nil, err
		}
	}
	maximum, inputs, latches, outputs, ands := counts[0], counts[1], counts[2], counts[3], counts[4]
	bad, constraints, justice, fairness := counts[5], counts[6], counts[7], counts[8]
	if justice != 0 || fairness != 0 {
		return nil, ErrUnsupported
	}
	if len(fields) > 6 && outputs != 0 {
		return nil, ErrFormat
	}
	if uint64(inputs)+uint64(latches)+uint64(ands) != uint64(maximum) {
		return nil, ErrFormat
	}
	primaryOutputs := uint64(outputs) + uint64(bad) + uint64(constraints)
	if uint64(maximum) > uint64(options.MaxVariables) || primaryOutputs > uint64(options.MaxOutputs) {
		return nil, ErrLimit
	}
	objectCount := uint64(maximum) + primaryOutputs + uint64(latches) + 1
	if objectCount > uint64(MaxLiteralID)+1 {
		return nil, ErrLimit
	}
	minimumBytes := uint64(latches) + primaryOutputs
	if !compact {
		minimumBytes *= 2
	}
	minimumBytes += 2 * uint64(ands)
	if minimumBytes > uint64(len(data)-cursor.pos) {
		return nil, ErrFormat
	}
	driverCount := int(latches) + int(primaryOutputs)
	drivers := make([]uint32, driverCount)
	literalMaximum := 2*maximum + 1
	if compact {
		var previous int64
		for index := range drivers {
			if err = aigIOAStop(options); err != nil {
				return nil, err
			}
			encoded, decodeErr := cursor.variable()
			if decodeErr != nil {
				return nil, decodeErr
			}
			if index == 0 {
				previous = int64(encoded)
			} else {
				difference := int64(encoded >> 1)
				if encoded&1 != 0 {
					difference = -difference
				}
				previous += difference
			}
			if previous < 0 || previous > int64(literalMaximum) {
				return nil, ErrFormat
			}
			drivers[index] = uint32(previous)
		}
	} else {
		for index := range drivers {
			if err = aigIOAStop(options); err != nil {
				return nil, err
			}
			line, lineErr := cursor.line()
			if lineErr != nil {
				return nil, ErrFormat
			}
			fields := bytes.Fields(line)
			if len(fields) == 0 || len(fields) > 1 && (index >= int(latches) || len(fields) > 2) {
				return nil, ErrFormat
			}
			drivers[index], err = aigIOADecimal(fields[0], literalMaximum)
			if err != nil {
				return nil, err
			}
			if len(fields) == 2 {
				initial, initialErr := aigIOADecimal(fields[1], literalMaximum)
				if initialErr != nil {
					return nil, initialErr
				}
				if initial != 0 {
					return nil, ErrUnsupported
				}
			}
		}
	}
	graph := NewAigManager()
	mapping := make([]AigLit, int(maximum)+1)
	mapping[0] = AigConst0
	for index := uint32(0); index < inputs+latches; index++ {
		if err = aigIOAStop(options); err != nil {
			return nil, err
		}
		mapping[index+1], err = graph.CreateCI()
		if err != nil {
			return nil, err
		}
	}
	convert := func(raw uint32) AigLit { return mapping[raw>>1].NotIf(raw&1 != 0) }
	for index := uint32(0); index < ands; index++ {
		if err = aigIOAStop(options); err != nil {
			return nil, err
		}
		lhs := 2 * (inputs + latches + index + 1)
		first, firstErr := cursor.variable()
		if firstErr != nil || first > lhs {
			return nil, ErrFormat
		}
		rightHigh := lhs - first
		second, secondErr := cursor.variable()
		if secondErr != nil || second > rightHigh {
			return nil, ErrFormat
		}
		rightLow := rightHigh - second
		if rightHigh>>1 >= lhs>>1 || rightLow>>1 >= lhs>>1 {
			return nil, ErrFormat
		}
		mapping[lhs>>1], err = graph.And(convert(rightHigh), convert(rightLow))
		if err != nil {
			return nil, err
		}
	}
	for index, raw := range drivers[int(latches):] {
		if err = aigIOAStop(options); err != nil {
			return nil, err
		}
		driver := convert(raw)
		if index >= int(primaryOutputs)-int(constraints) {
			driver = driver.Not()
		}
		if _, err = graph.CreateCO(driver); err != nil {
			return nil, err
		}
	}
	for _, raw := range drivers[:int(latches)] {
		if err = aigIOAStop(options); err != nil {
			return nil, err
		}
		if _, err = graph.CreateCO(convert(raw)); err != nil {
			return nil, err
		}
	}
	if err = graph.SetRegisterCount(int(latches)); err != nil {
		return nil, err
	}
	if err = graph.SetConstraintCount(int(constraints)); err != nil {
		return nil, err
	}
	graph.Cleanup()
	if cursor.pos+2 <= len(data) && data[cursor.pos] == 'c' && data[cursor.pos+1] == 'n' {
		name := data[cursor.pos+2:]
		end := bytes.IndexByte(name, 0)
		if end < 0 {
			return nil, ErrFormat
		}
		graph.Name = string(name[:end])
	}
	if options.Check {
		if err = graph.Check(); err != nil {
			return nil, err
		}
	}
	if err = aigIOAStop(options); err != nil {
		return nil, err
	}
	return graph, nil
}

// IOAReadAiger reads a bounded file and applies the source's filename-derived
// Name and full-path Spec fields.
func IOAReadAiger(path string, options AigIOAOptions) (*AigManager, error) {
	options, err := normalizeAigIOAOptions(options)
	if err != nil {
		return nil, err
	}
	info, err := os.Stat(path)
	if err != nil {
		return nil, err
	}
	if info.Size() < 0 || info.Size() > int64(options.MaxBytes) {
		return nil, ErrLimit
	}
	file, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	readLimit := int64(options.MaxBytes)
	if readLimit < math.MaxInt64 {
		readLimit++
	}
	data, readErr := io.ReadAll(io.LimitReader(file, readLimit))
	closeErr := file.Close()
	if readErr != nil {
		return nil, readErr
	}
	if closeErr != nil {
		return nil, closeErr
	}
	if len(data) > options.MaxBytes {
		return nil, ErrLimit
	}
	graph, err := IOAReadAigerFromMemory(data, options)
	if err != nil {
		return nil, err
	}
	graph.Name = IOAFileNameGeneric(path)
	graph.Spec = path
	return graph, nil
}

type aigIOAOutput struct {
	data    []byte
	options AigIOAOptions
	steps   int
}

func (output *aigIOAOutput) append(data ...byte) error {
	if len(data) > output.options.MaxBytes-len(output.data) {
		return ErrLimit
	}
	output.data = append(output.data, data...)
	output.steps++
	if output.steps&255 == 0 {
		return aigIOAStop(output.options)
	}
	return nil
}

func (output *aigIOAOutput) number(value uint32) error {
	return output.append(strconv.AppendUint(nil, uint64(value), 10)...)
}

func (output *aigIOAOutput) variable(value uint32) error {
	return output.append(IOAEncodeAigerUint(value)...)
}

func ioaSerializeAiger(graph *AigManager, options AigIOAOptions, fileStyle, includeName bool) ([]byte, error) {
	options, err := normalizeAigIOAOptions(options)
	if err != nil {
		return nil, err
	}
	if graph == nil {
		return nil, ErrDimension
	}
	if err = graph.Check(); err != nil {
		return nil, err
	}
	if err = aigIOAStop(options); err != nil {
		return nil, err
	}
	inputs, registers := graph.PrimaryInputCount(), graph.RegisterCount()
	primaryOutputs, constraints := graph.PrimaryOutputCount(), graph.ConstraintCount()
	if uint64(graph.InputCount())+uint64(graph.AndCount()) > uint64(options.MaxVariables) || primaryOutputs > options.MaxOutputs {
		return nil, ErrLimit
	}
	raw := make([]uint32, len(graph.objects))
	raw[0] = 1
	for index, id := range graph.cis {
		raw[id] = 2 * uint32(index+1)
	}
	convert := func(literal AigLit) uint32 { return raw[literal.ID()] ^ uint32(literal&1) }
	next := uint32(len(graph.cis) + 1)
	type deltaPair struct{ high, low uint32 }
	deltas := make([]deltaPair, 0, graph.AndCount())
	for id, object := range graph.objects {
		if err = aigIOAStop(options); err != nil {
			return nil, err
		}
		switch object.Type {
		case AigBuffer:
			raw[id] = convert(object.Fanin0)
		case AigAND:
			lhs := 2 * next
			next++
			high, low := convert(object.Fanin0), convert(object.Fanin1)
			if high < low {
				high, low = low, high
			}
			if high >= lhs {
				return nil, ErrInvariant
			}
			deltas = append(deltas, deltaPair{lhs - high, high - low})
			raw[id] = lhs
		}
	}
	drivers := make([]uint32, 0, graph.OutputCount())
	for index := 0; index < registers; index++ {
		if err = aigIOAStop(options); err != nil {
			return nil, err
		}
		id := graph.cos[primaryOutputs+index]
		drivers = append(drivers, convert(graph.objects[id].Fanin0))
	}
	for index, id := range graph.cos[:primaryOutputs] {
		if err = aigIOAStop(options); err != nil {
			return nil, err
		}
		driver := graph.objects[id].Fanin0
		if fileStyle && index >= primaryOutputs-constraints {
			driver = driver.Not()
		}
		drivers = append(drivers, convert(driver))
	}
	output := aigIOAOutput{options: options}
	magic := "aig "
	if fileStyle && options.Compact {
		magic = "aig2 "
	}
	if err = output.append([]byte(magic)...); err != nil {
		return nil, err
	}
	counts := []uint32{next - 1, uint32(inputs), uint32(registers)}
	if !fileStyle || constraints == 0 {
		counts = append(counts, uint32(primaryOutputs), uint32(len(deltas)))
	} else {
		counts = append(counts, 0, uint32(len(deltas)), uint32(primaryOutputs-constraints), uint32(constraints))
	}
	for index, count := range counts {
		if index != 0 {
			if err = output.append(' '); err != nil {
				return nil, err
			}
		}
		if err = output.number(count); err != nil {
			return nil, err
		}
	}
	if err = output.append('\n'); err != nil {
		return nil, err
	}
	if fileStyle && options.Compact {
		encoded, encodeErr := IOAEncodeAigerLiterals(drivers, options.MaxBytes-len(output.data))
		if encodeErr != nil {
			return nil, encodeErr
		}
		if err = output.append(encoded...); err != nil {
			return nil, err
		}
	} else {
		for _, driver := range drivers {
			if err = output.number(driver); err != nil {
				return nil, err
			}
			if err = output.append('\n'); err != nil {
				return nil, err
			}
		}
	}
	for _, pair := range deltas {
		if err = output.variable(pair.high); err != nil {
			return nil, err
		}
		if err = output.variable(pair.low); err != nil {
			return nil, err
		}
	}
	if err = output.append('c'); err != nil {
		return nil, err
	}
	if includeName && graph.Name != "" {
		if bytes.ContainsRune([]byte(graph.Name), 0) {
			return nil, ErrDimension
		}
		if err = output.append('n'); err != nil {
			return nil, err
		}
		if err = output.append([]byte(graph.Name)...); err != nil {
			return nil, err
		}
		if err = output.append(0); err != nil {
			return nil, err
		}
	}
	if err = aigIOAStop(options); err != nil {
		return nil, err
	}
	return output.data, nil
}

// IOAWriteAigerIntoMemoryStr implements Ioa_WriteAigerIntoMemoryStr. Like the
// source helper, it always emits standard drivers, treats every primary output
// alike, ends at the comment marker, and omits the model name.
func IOAWriteAigerIntoMemoryStr(graph *AigManager, options AigIOAOptions) ([]byte, error) {
	return ioaSerializeAiger(graph, options, false, false)
}

// IOAWriteAigerIntoMemory implements Ioa_WriteAigerIntoMemory, including its
// optional cn model-name extension.
func IOAWriteAigerIntoMemory(graph *AigManager, options AigIOAOptions) ([]byte, error) {
	return ioaSerializeAiger(graph, options, false, true)
}

// IOAMarshalAiger exposes the complete file writer's byte stream before its
// timestamp trailer. It supports aig2 and externalizes constraint polarity.
func IOAMarshalAiger(graph *AigManager, options AigIOAOptions) ([]byte, error) {
	return ioaSerializeAiger(graph, options, true, true)
}

func ioaWriteAigerFile(data []byte, path string, options AigIOAOptions) error {
	stamp := options.Timestamp
	if stamp.IsZero() {
		stamp = time.Now()
	}
	trailer := []byte("\nThis file was produced by the IOA package in ABC on " + IOATimeStampAt(stamp) + "\nFor information about AIGER format, refer to http://fmv.jku.at/aiger\n")
	if len(trailer) > options.MaxBytes-len(data) {
		return ErrLimit
	}
	data = append(data, trailer...)
	if err := aigIOAStop(options); err != nil {
		return err
	}
	directory, base := filepath.Dir(path), filepath.Base(path)
	temporary, err := os.CreateTemp(directory, "."+base+".tmp-")
	if err != nil {
		return err
	}
	temporaryName := temporary.Name()
	keep := false
	defer func() {
		temporary.Close()
		if !keep {
			os.Remove(temporaryName)
		}
	}()
	if _, err = temporary.Write(data); err != nil {
		return err
	}
	if err = temporary.Sync(); err != nil {
		return err
	}
	if err = temporary.Close(); err != nil {
		return err
	}
	if err = os.Rename(temporaryName, path); err != nil {
		return err
	}
	keep = true
	return nil
}

// IOAWriteAigerBufferTest implements Ioa_WriteAigerBufferTest: it uses the
// standard memory serializer even when Compact is set, then adds the source
// trailer. It atomically replaces path only after complete serialization.
func IOAWriteAigerBufferTest(graph *AigManager, path string, options AigIOAOptions) error {
	var err error
	options, err = normalizeAigIOAOptions(options)
	if err != nil {
		return err
	}
	if graph == nil || graph.OutputCount() == 0 {
		return ErrDimension
	}
	data, err := IOAWriteAigerIntoMemory(graph, options)
	if err != nil {
		return err
	}
	return ioaWriteAigerFile(data, path, options)
}

// IOAWriteAiger atomically replaces path after complete bounded
// serialization. Symbol writing remains absent because the source's symbol
// block is disabled; graph names retain the source cn extension.
func IOAWriteAiger(graph *AigManager, path string, options AigIOAOptions) error {
	var err error
	options, err = normalizeAigIOAOptions(options)
	if err != nil {
		return err
	}
	if graph == nil || graph.OutputCount() == 0 {
		return ErrDimension
	}
	data, err := IOAMarshalAiger(graph, options)
	if err != nil {
		return err
	}
	return ioaWriteAigerFile(data, path, options)
}

func IOAFileSize(path string) (int64, error) {
	info, err := os.Stat(path)
	if err != nil {
		return 0, err
	}
	return info.Size(), nil
}

func IOAFileNameGeneric(name string) string {
	if index := bytes.LastIndexByte([]byte(name), '.'); index >= 0 {
		return name[:index]
	}
	return name
}

func IOAFileNameGenericAppend(base, suffix string) string {
	combined := suffix
	if base != "" {
		combined = IOAFileNameGeneric(base) + suffix
	}
	start := len(combined)
	for start > 0 {
		entry := combined[start-1]
		if entry >= '0' && entry <= '9' || entry >= 'a' && entry <= 'z' || entry >= 'A' && entry <= 'Z' || entry == '_' || entry == '.' {
			start--
			continue
		}
		break
	}
	return combined[start:]
}

func IOATimeStamp() string { return IOATimeStampAt(time.Now()) }

func IOATimeStampAt(at time.Time) string {
	return at.Format("Mon Jan _2 15:04:05 2006")
}
