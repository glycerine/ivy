package abc

// Binary AIG Format reader and writer from src/base/io/ioReadBaf.c and
// src/base/io/ioWriteBaf.c.

import (
	"encoding/binary"
	"slices"
	"strconv"
	"strings"
)

// ABCBAFOptions bounds BAF parsing, formatting and retained state. Numeric
// zeroes select defaults.
type ABCBAFOptions struct {
	MaxInputBytes  int
	MaxOutputBytes int
	MaxObjects     int
	MaxNameBytes   int
	MaxWork        int
	MaxBytes       int
	Stop           func() bool
}

func normalizeABCBAFOptions(options ABCBAFOptions) (ABCBAFOptions, error) {
	values := []*int{&options.MaxInputBytes, &options.MaxOutputBytes, &options.MaxObjects, &options.MaxNameBytes, &options.MaxWork, &options.MaxBytes}
	defaults := [...]int{256 << 20, 256 << 20, 1 << 24, 64 << 20, 1 << 26, 512 << 20}
	for index, value := range values {
		if *value < 0 {
			return options, ErrDimension
		}
		if *value == 0 {
			*value = defaults[index]
		}
	}
	return options, nil
}

type abcBAFBudget struct {
	options ABCBAFOptions
	work    int
}

func (budget *abcBAFBudget) take(amount int) error {
	if amount < 0 || budget.work > budget.options.MaxWork-amount {
		return ErrLimit
	}
	if budget.options.Stop != nil && budget.options.Stop() {
		return ErrCancelled
	}
	budget.work += amount
	return nil
}

// ABCBAFLatch preserves the three names and the two-bit initialization value
// stored for one BAF latch.
type ABCBAFLatch struct {
	Name       string
	InputName  string
	OutputName string
	Init       ABCLatchInit
}

// ABCBAFModel is an immutable owned BAF design. Name vectors are kept apart
// because the source format permits output names to overlap other categories.
type ABCBAFModel struct {
	name    string
	inputs  []string
	outputs []string
	latches []ABCBAFLatch
	graph   *AigManager
}

// Name returns the design name.
func (model *ABCBAFModel) Name() string {
	if model == nil {
		return ""
	}
	return model.name
}

// InputNames returns an owned copy in primary-input order.
func (model *ABCBAFModel) InputNames() []string {
	if model == nil {
		return nil
	}
	return slices.Clone(model.inputs)
}

// OutputNames returns an owned copy in primary-output order.
func (model *ABCBAFModel) OutputNames() []string {
	if model == nil {
		return nil
	}
	return slices.Clone(model.outputs)
}

// Latches returns owned latch metadata in register order.
func (model *ABCBAFModel) Latches() []ABCBAFLatch {
	if model == nil {
		return nil
	}
	return slices.Clone(model.latches)
}

// Graph returns an independent graph clone.
func (model *ABCBAFModel) Graph() *AigManager {
	if model == nil || model.graph == nil {
		return nil
	}
	return model.graph.Clone()
}

func abcBAFNameBytes(model *ABCBAFModel) (int, error) {
	total := len(model.name)
	for _, name := range model.inputs {
		var err error
		total, err = jsonCheckedAdd(total, len(name))
		if err != nil {
			return 0, err
		}
	}
	for _, name := range model.outputs {
		var err error
		total, err = jsonCheckedAdd(total, len(name))
		if err != nil {
			return 0, err
		}
	}
	for _, latch := range model.latches {
		for _, name := range [...]string{latch.Name, latch.InputName, latch.OutputName} {
			var err error
			total, err = jsonCheckedAdd(total, len(name))
			if err != nil {
				return 0, err
			}
		}
	}
	return total, nil
}

func validateABCBAFModel(model *ABCBAFModel, budget *abcBAFBudget) error {
	if model == nil || model.graph == nil {
		return ErrDimension
	}
	if err := model.graph.Check(); err != nil {
		return err
	}
	if model.graph.PrimaryInputCount() != len(model.inputs) || model.graph.PrimaryOutputCount() != len(model.outputs) || model.graph.RegisterCount() != len(model.latches) {
		return ErrInvariant
	}
	if model.graph.ConstraintCount() != 0 {
		return ErrUnsupported
	}
	if model.graph.ObjectCount() > budget.options.MaxObjects {
		return ErrLimit
	}
	nameBytes, err := abcBAFNameBytes(model)
	if err != nil || nameBytes > budget.options.MaxNameBytes {
		return ErrLimit
	}
	if nameBytes > budget.options.MaxBytes || model.graph.ObjectCount() > (budget.options.MaxBytes-nameBytes)/32 {
		return ErrLimit
	}
	for _, name := range append(append([]string{model.name}, model.inputs...), model.outputs...) {
		if strings.IndexByte(name, 0) >= 0 {
			return ErrFormat
		}
	}
	for _, latch := range model.latches {
		if latch.Init > ABCInitDC || strings.IndexByte(latch.Name, 0) >= 0 || strings.IndexByte(latch.InputName, 0) >= 0 || strings.IndexByte(latch.OutputName, 0) >= 0 {
			return ErrFormat
		}
	}
	return budget.take(model.graph.ObjectCount() + len(model.inputs) + len(model.outputs) + 3*len(model.latches))
}

// NewABCBAFModel snapshots a graph and complete BAF name metadata.
func NewABCBAFModel(name string, inputs, outputs []string, latches []ABCBAFLatch, graph *AigManager, options ABCBAFOptions) (*ABCBAFModel, error) {
	options, err := normalizeABCBAFOptions(options)
	if err != nil {
		return nil, err
	}
	if graph == nil {
		return nil, ErrDimension
	}
	model := &ABCBAFModel{name: name, inputs: slices.Clone(inputs), outputs: slices.Clone(outputs), latches: slices.Clone(latches), graph: graph.Clone()}
	budget := abcBAFBudget{options: options}
	if err = validateABCBAFModel(model, &budget); err != nil {
		return nil, err
	}
	return model, nil
}

func abcBAFReadCString(data []byte, cursor *int, end int, nameBytes *int, budget *abcBAFBudget) (string, error) {
	start := *cursor
	for *cursor < end && data[*cursor] != 0 {
		*cursor++
	}
	if *cursor >= end {
		return "", ErrFormat
	}
	name := string(data[start:*cursor])
	*cursor++
	var err error
	*nameBytes, err = jsonCheckedAdd(*nameBytes, len(name))
	if err != nil || *nameBytes > budget.options.MaxNameBytes {
		return "", ErrLimit
	}
	if err = budget.take(len(name) + 1); err != nil {
		return "", err
	}
	return name, nil
}

func abcBAFParseCount(text string, maximum int) (int, error) {
	if text == "" {
		return 0, ErrFormat
	}
	for index := range text {
		if text[index] < '0' || text[index] > '9' {
			return 0, ErrFormat
		}
	}
	value, err := strconv.ParseUint(text, 10, 63)
	if err != nil || value > uint64(maximum) {
		return 0, ErrLimit
	}
	return int(value), nil
}

func abcBAFDecodeLiteral(word uint32, mapping []AigLit, maximum int) (AigLit, error) {
	id := int(word >> 1)
	if id < 0 || id >= maximum || id >= len(mapping) || mapping[id] == InvalidAigLit {
		return InvalidAigLit, ErrFormat
	}
	return mapping[id].NotIf(word&1 != 0), nil
}

// ParseABCBAF reads owned bytes using the source BAF layout. The body is
// explicitly little-endian, replacing the C implementation's host-endian
// unsigned writes with a deterministic standalone representation.
func ParseABCBAF(data []byte, options ABCBAFOptions) (*ABCBAFModel, error) {
	options, err := normalizeABCBAFOptions(options)
	if err != nil {
		return nil, err
	}
	if len(data) > options.MaxInputBytes || len(data) > options.MaxBytes {
		return nil, ErrLimit
	}
	budget := abcBAFBudget{options: options}
	cursor := 0
	for cursor < len(data) && data[cursor] == '#' {
		newline := cursor
		for newline < len(data) && data[newline] != '\n' {
			newline++
		}
		if newline == len(data) {
			return nil, ErrFormat
		}
		if err = budget.take(newline + 1 - cursor); err != nil {
			return nil, err
		}
		cursor = newline + 1
	}
	nameBytes := 0
	name, err := abcBAFReadCString(data, &cursor, len(data), &nameBytes, &budget)
	if err != nil {
		return nil, err
	}
	counts := [4]int{}
	for index := range counts {
		text, readErr := abcBAFReadCString(data, &cursor, len(data), &nameBytes, &budget)
		if readErr != nil {
			return nil, readErr
		}
		counts[index], err = abcBAFParseCount(text, options.MaxObjects)
		if err != nil {
			return nil, err
		}
	}
	inputs, outputs, registers, ands := counts[0], counts[1], counts[2], counts[3]
	ciCount, err := checkedSize(inputs, registers, 1, options.MaxObjects)
	if err != nil {
		return nil, ErrLimit
	}
	coCount, err := checkedSize(outputs, registers, 1, options.MaxObjects)
	if err != nil {
		return nil, ErrLimit
	}
	objects, err := checkedSize(1, ciCount, 1, options.MaxObjects)
	if err == nil {
		objects, err = checkedSize(objects, ands, 1, options.MaxObjects)
	}
	if err == nil {
		objects, err = checkedSize(objects, coCount, 1, options.MaxObjects)
	}
	if err != nil || objects > options.MaxObjects || objects > (options.MaxBytes-nameBytes)/32 {
		return nil, ErrLimit
	}
	words, err := checkedSize(0, ands, 2, int(^uint(0)>>1))
	if err == nil {
		words, err = checkedSize(words, coCount, 1, int(^uint(0)>>1))
	}
	bodyBytes, sizeErr := checkedSize(0, words, 4, len(data))
	if err != nil || sizeErr != nil || bodyBytes > len(data)-cursor {
		return nil, ErrFormat
	}
	bodyStart := len(data) - bodyBytes
	inputNames := make([]string, inputs)
	for index := range inputNames {
		inputNames[index], err = abcBAFReadCString(data, &cursor, bodyStart, &nameBytes, &budget)
		if err != nil {
			return nil, err
		}
	}
	outputNames := make([]string, outputs)
	for index := range outputNames {
		outputNames[index], err = abcBAFReadCString(data, &cursor, bodyStart, &nameBytes, &budget)
		if err != nil {
			return nil, err
		}
	}
	latches := make([]ABCBAFLatch, registers)
	for index := range latches {
		latches[index].Name, err = abcBAFReadCString(data, &cursor, bodyStart, &nameBytes, &budget)
		if err == nil {
			latches[index].InputName, err = abcBAFReadCString(data, &cursor, bodyStart, &nameBytes, &budget)
		}
		if err == nil {
			latches[index].OutputName, err = abcBAFReadCString(data, &cursor, bodyStart, &nameBytes, &budget)
		}
		if err != nil {
			return nil, err
		}
	}
	if cursor != bodyStart || nameBytes > options.MaxNameBytes {
		return nil, ErrFormat
	}
	manager := NewAigManager()
	mapping := make([]AigLit, 1+ciCount+ands)
	for index := range mapping {
		mapping[index] = InvalidAigLit
	}
	mapping[0] = AigConst1
	for index := 0; index < ciCount; index++ {
		mapping[1+index], err = manager.CreateCI()
		if err != nil {
			return nil, err
		}
	}
	wordAt := func(index int) uint32 { return binary.LittleEndian.Uint32(data[bodyStart+4*index:]) }
	for index := 0; index < ands; index++ {
		if err = budget.take(1); err != nil {
			return nil, err
		}
		maximum := 1 + ciCount + index
		first, firstErr := abcBAFDecodeLiteral(wordAt(2*index), mapping, maximum)
		second, secondErr := abcBAFDecodeLiteral(wordAt(2*index+1), mapping, maximum)
		if firstErr != nil || secondErr != nil {
			return nil, ErrFormat
		}
		mapping[maximum], err = manager.And(first, second)
		if err != nil {
			return nil, err
		}
	}
	for index := 0; index < coCount; index++ {
		if err = budget.take(1); err != nil {
			return nil, err
		}
		word := wordAt(2*ands + index)
		if index >= outputs {
			latches[index-outputs].Init = ABCLatchInit(word & 3)
			word >>= 2
		}
		driver, decodeErr := abcBAFDecodeLiteral(word, mapping, len(mapping))
		if decodeErr != nil {
			return nil, decodeErr
		}
		if _, err = manager.CreateCO(driver); err != nil {
			return nil, err
		}
	}
	if err = manager.SetRegisterCount(registers); err != nil {
		return nil, err
	}
	manager.Name = name
	model := &ABCBAFModel{name: name, inputs: inputNames, outputs: outputNames, latches: latches, graph: manager}
	if err = validateABCBAFModel(model, &budget); err != nil {
		return nil, err
	}
	return model, nil
}

func abcBAFGraphRemap(graph *AigManager, budget *abcBAFBudget) ([]uint32, []AigID, error) {
	remap := make([]uint32, len(graph.objects))
	for index := range remap {
		remap[index] = ^uint32(0)
	}
	remap[0] = 0
	next := uint32(1)
	for _, id := range graph.cis {
		if next > MaxLiteralID {
			return nil, nil, ErrLimit
		}
		remap[id] = next
		next++
	}
	ands := make([]AigID, 0, graph.AndCount())
	for id, object := range graph.objects {
		if err := budget.take(1); err != nil {
			return nil, nil, err
		}
		switch object.Type {
		case AigNone, AigConstant1, AigCI, AigCO:
			continue
		case AigAND:
			if next > MaxLiteralID {
				return nil, nil, ErrLimit
			}
			remap[id] = next
			next++
			ands = append(ands, AigID(id))
		default:
			return nil, nil, ErrUnsupported
		}
	}
	return remap, ands, nil
}

func abcBAFEncodeLiteral(literal AigLit, remap []uint32) (uint32, error) {
	if !literal.Valid() || int(literal.ID()) >= len(remap) || remap[literal.ID()] == ^uint32(0) {
		return 0, ErrInvariant
	}
	return remap[literal.ID()]<<1 | uint32(literal&1), nil
}

// MarshalABCBAFModel emits deterministic little-endian BAF bytes. A stable
// comment replaces Io_WriteBaf's process-clock timestamp.
func MarshalABCBAFModel(model *ABCBAFModel, options ABCBAFOptions) ([]byte, error) {
	options, err := normalizeABCBAFOptions(options)
	if err != nil {
		return nil, err
	}
	budget := abcBAFBudget{options: options}
	if err = validateABCBAFModel(model, &budget); err != nil {
		return nil, err
	}
	remap, ands, err := abcBAFGraphRemap(model.graph, &budget)
	if err != nil {
		return nil, err
	}
	words, err := checkedSize(0, len(ands), 2, int(^uint(0)>>1))
	if err == nil {
		words, err = checkedSize(words, model.graph.OutputCount(), 1, int(^uint(0)>>1))
	}
	if err != nil {
		return nil, ErrLimit
	}
	result := make([]byte, 0)
	appendText := func(text string) error {
		if len(text) > options.MaxOutputBytes-len(result) || len(text) > options.MaxBytes-len(result) {
			return ErrLimit
		}
		if takeErr := budget.take(len(text)); takeErr != nil {
			return takeErr
		}
		result = append(result, text...)
		return nil
	}
	appendCString := func(text string) error {
		if strings.IndexByte(text, 0) >= 0 {
			return ErrFormat
		}
		if err := appendText(text); err != nil {
			return err
		}
		return appendText("\x00")
	}
	if err = appendText("# BAF (Binary Aig Format)\n"); err != nil {
		return nil, err
	}
	for _, field := range [...]string{model.name, strconv.Itoa(len(model.inputs)), strconv.Itoa(len(model.outputs)), strconv.Itoa(len(model.latches)), strconv.Itoa(len(ands))} {
		if err = appendCString(field); err != nil {
			return nil, err
		}
	}
	for _, name := range model.inputs {
		if err = appendCString(name); err != nil {
			return nil, err
		}
	}
	for _, name := range model.outputs {
		if err = appendCString(name); err != nil {
			return nil, err
		}
	}
	for _, latch := range model.latches {
		for _, name := range [...]string{latch.Name, latch.InputName, latch.OutputName} {
			if err = appendCString(name); err != nil {
				return nil, err
			}
		}
	}
	bodyBytes, err := checkedSize(0, words, 4, options.MaxOutputBytes-len(result))
	if err != nil || bodyBytes > options.MaxBytes-len(result) {
		return nil, ErrLimit
	}
	if err = budget.take(words); err != nil {
		return nil, err
	}
	bodyStart := len(result)
	result = append(result, make([]byte, bodyBytes)...)
	wordIndex := 0
	put := func(word uint32) {
		binary.LittleEndian.PutUint32(result[bodyStart+4*wordIndex:], word)
		wordIndex++
	}
	for _, id := range ands {
		object := model.graph.objects[id]
		first, firstErr := abcBAFEncodeLiteral(object.Fanin0, remap)
		second, secondErr := abcBAFEncodeLiteral(object.Fanin1, remap)
		if firstErr != nil || secondErr != nil {
			return nil, ErrInvariant
		}
		put(first)
		put(second)
	}
	for index, id := range model.graph.cos {
		word, encodeErr := abcBAFEncodeLiteral(model.graph.objects[id].Fanin0, remap)
		if encodeErr != nil {
			return nil, encodeErr
		}
		if index >= len(model.outputs) {
			if word > ^uint32(0)>>2 {
				return nil, ErrLimit
			}
			word = word<<2 | uint32(model.latches[index-len(model.outputs)].Init)
		}
		put(word)
	}
	if wordIndex != words {
		return nil, ErrInvariant
	}
	return result, nil
}

// ABCBAFModelFromNetwork is the owned-model counterpart of Io_WriteBaf's
// network traversal.
func ABCBAFModelFromNetwork(network *ABCNetwork, options ABCBAFOptions) (*ABCBAFModel, error) {
	if network == nil || network.Type != ABCNetworkStrash || network.Function != ABCFunctionAIG {
		return nil, ErrDimension
	}
	options, err := normalizeABCBAFOptions(options)
	if err != nil {
		return nil, err
	}
	converted, err := ABCNetworkToAIG(network, ABCFunctionalOptions{MaxOperations: options.MaxWork, Stop: options.Stop})
	if err != nil {
		return nil, err
	}
	inputs := make([]string, len(network.pis))
	for index, id := range network.pis {
		inputs[index] = network.objects[id].Name
	}
	outputs := make([]string, len(network.pos))
	for index, id := range network.pos {
		outputs[index] = network.objects[id].Name
	}
	latchInputs, latchOutputs, err := abcLatchTerminals(network)
	if err != nil {
		return nil, err
	}
	latches := make([]ABCBAFLatch, len(network.boxes))
	for index, id := range network.boxes {
		latches[index] = ABCBAFLatch{Name: network.objects[id].Name, InputName: network.objects[latchInputs[index]].Name, OutputName: network.objects[latchOutputs[index]].Name, Init: network.objects[id].Init}
	}
	return NewABCBAFModel(network.Name, inputs, outputs, latches, converted.Graph, options)
}

// MarshalABCBAF is the owned-byte counterpart of Io_WriteBaf.
func MarshalABCBAF(network *ABCNetwork, options ABCBAFOptions) ([]byte, error) {
	model, err := ABCBAFModelFromNetwork(network, options)
	if err != nil {
		return nil, err
	}
	return MarshalABCBAFModel(model, options)
}

// ABCNetworkFromBAFModel converts a complete BAF model to the concrete ABC
// network. Name collisions unsupported by ABCNetwork's unique-name index are
// reported rather than changing the model's exact metadata.
func ABCNetworkFromBAFModel(model *ABCBAFModel, check bool, options ABCBAFOptions) (*ABCNetwork, error) {
	options, err := normalizeABCBAFOptions(options)
	if err != nil {
		return nil, err
	}
	budget := abcBAFBudget{options: options}
	if err = validateABCBAFModel(model, &budget); err != nil {
		return nil, err
	}
	network, err := AIGToABCNetwork(model.graph, ABCFunctionalOptions{MaxOperations: options.MaxWork, Stop: options.Stop})
	if err != nil {
		return nil, err
	}
	network.Name = model.name
	setName := func(id int, name string) error {
		if name == "" {
			return nil
		}
		return network.SetObjectName(id, name)
	}
	for index, id := range network.pis {
		if err = setName(id, model.inputs[index]); err != nil {
			return nil, ErrUnsupported
		}
	}
	for index, id := range network.pos {
		if err = setName(id, model.outputs[index]); err != nil {
			return nil, ErrUnsupported
		}
	}
	latchInputs, latchOutputs, err := abcLatchTerminals(network)
	if err != nil {
		return nil, err
	}
	for index, id := range network.boxes {
		latch := model.latches[index]
		if err = setName(id, latch.Name); err == nil {
			err = setName(latchInputs[index], latch.InputName)
		}
		if err == nil {
			err = setName(latchOutputs[index], latch.OutputName)
		}
		if err != nil {
			return nil, ErrUnsupported
		}
		network.objects[id].Init = latch.Init
	}
	if check {
		if err = network.Validate(); err != nil {
			return nil, err
		}
	}
	return network, nil
}

// ReadABCBAF is the owned-byte counterpart of Io_ReadBaf.
func ReadABCBAF(data []byte, check bool, options ABCBAFOptions) (*ABCNetwork, error) {
	model, err := ParseABCBAF(data, options)
	if err != nil {
		return nil, err
	}
	return ABCNetworkFromBAFModel(model, check, options)
}
