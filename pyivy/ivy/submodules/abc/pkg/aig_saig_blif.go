package abc

// Restricted sequential-AIG BLIF input/output from src/aig/saig/saigIoa.c.

import (
	"bytes"
	"fmt"
	"os"
	"strconv"
	"strings"
)

type SAIGBLIFOptions struct {
	MaxInputBytes  int
	MaxOutputBytes int
	MaxObjects     int
	MaxTokens      int
	MaxTokenBytes  int
	MaxWork        int
	Stop           func() bool
}

type saigBLIFBudget struct {
	options SAIGBLIFOptions
	work    int
}

func normalizeSAIGBLIFOptions(options SAIGBLIFOptions) (SAIGBLIFOptions, error) {
	if options.MaxInputBytes < 0 || options.MaxOutputBytes < 0 || options.MaxObjects < 0 || options.MaxTokens < 0 || options.MaxTokenBytes < 0 || options.MaxWork < 0 {
		return options, ErrDimension
	}
	defaults := []struct {
		value *int
		set   int
	}{
		{&options.MaxInputBytes, 64 << 20},
		{&options.MaxOutputBytes, 64 << 20},
		{&options.MaxObjects, 1 << 20},
		{&options.MaxTokens, 16 << 20},
		{&options.MaxTokenBytes, 1 << 20},
		{&options.MaxWork, 32 << 20},
	}
	for _, item := range defaults {
		if *item.value == 0 {
			*item.value = item.set
		}
	}
	return options, nil
}

func (budget *saigBLIFBudget) step(amount int) error {
	if amount < 0 || budget.work > budget.options.MaxWork-amount {
		return ErrLimit
	}
	budget.work += amount
	if budget.options.Stop != nil && budget.options.Stop() {
		return ErrCancelled
	}
	return nil
}

type saigBLIFNames struct {
	graph       *AigManager
	ciOrdinal   []int
	coOrdinal   []int
	objectWidth int
	piWidth     int
	poWidth     int
	regWidth    int
}

func newSAIGBLIFNames(graph *AigManager) *saigBLIFNames {
	names := &saigBLIFNames{
		graph:       graph,
		ciOrdinal:   make([]int, graph.ObjectCount()),
		coOrdinal:   make([]int, graph.ObjectCount()),
		objectWidth: HopBase10Log(uint(graph.ObjectCount())),
		piWidth:     HopBase10Log(uint(graph.PrimaryInputCount())),
		poWidth:     HopBase10Log(uint(graph.PrimaryOutputCount())),
		regWidth:    HopBase10Log(uint(graph.RegisterCount())),
	}
	for index := range names.ciOrdinal {
		names.ciOrdinal[index], names.coOrdinal[index] = -1, -1
	}
	for index, id := range graph.cis {
		names.ciOrdinal[id] = index
	}
	for index, id := range graph.cos {
		names.coOrdinal[id] = index
	}
	return names
}

func (names *saigBLIFNames) name(id AigID) (string, error) {
	if uint64(id) >= uint64(names.graph.ObjectCount()) {
		return "", ErrInvalidLiteral
	}
	object := names.graph.objects[id]
	switch object.Type {
	case AigConstant1, AigAND:
		return fmt.Sprintf("n%0*d", names.objectWidth, id), nil
	case AigCI:
		index := names.ciOrdinal[id]
		if index < 0 {
			return "", ErrInvariant
		}
		if index < names.graph.PrimaryInputCount() {
			return fmt.Sprintf("pi%0*d", names.piWidth, index), nil
		}
		return fmt.Sprintf("lo%0*d", names.regWidth, index-names.graph.PrimaryInputCount()), nil
	case AigCO:
		index := names.coOrdinal[id]
		if index < 0 {
			return "", ErrInvariant
		}
		if index < names.graph.PrimaryOutputCount() {
			return fmt.Sprintf("po%0*d", names.poWidth, index), nil
		}
		return fmt.Sprintf("li%0*d", names.regWidth, index-names.graph.PrimaryOutputCount()), nil
	case AigBuffer:
		return "", ErrUnsupported
	default:
		return "", ErrInvalidLiteral
	}
}

// SAIGObjectName returns the owned counterpart of Saig_ObjName's static
// buffer. Names follow the source's object and terminal widths exactly.
func SAIGObjectName(graph *AigManager, id AigID) (string, error) {
	if graph == nil {
		return "", ErrDimension
	}
	if err := graph.Check(); err != nil {
		return "", err
	}
	return newSAIGBLIFNames(graph).name(id)
}

type saigBLIFWriter struct {
	buffer bytes.Buffer
	budget *saigBLIFBudget
}

func (writer *saigBLIFWriter) add(text string) error {
	if len(text) > writer.budget.options.MaxOutputBytes-writer.buffer.Len() {
		return ErrLimit
	}
	if err := writer.budget.step(1); err != nil {
		return err
	}
	writer.buffer.WriteString(text)
	return nil
}

// MarshalSAIGBLIF emits the exact structural subset consumed by
// ParseSAIGBLIF. It preserves CI/CO partitions and complemented edge cubes.
func MarshalSAIGBLIF(graph *AigManager, options SAIGBLIFOptions) ([]byte, error) {
	options, err := normalizeSAIGBLIFOptions(options)
	if err != nil {
		return nil, err
	}
	if graph == nil {
		return nil, ErrDimension
	}
	if err = graph.Check(); err != nil {
		return nil, err
	}
	if graph.OutputCount() == 0 || graph.ObjectCount() > options.MaxObjects || graph.Name == "" || len(strings.Fields(graph.Name)) != 1 {
		if graph.ObjectCount() > options.MaxObjects {
			return nil, ErrLimit
		}
		return nil, ErrDimension
	}
	for _, object := range graph.objects {
		if object.Type == AigBuffer {
			return nil, ErrUnsupported
		}
	}
	budget := &saigBLIFBudget{options: options}
	writer := &saigBLIFWriter{budget: budget}
	names := newSAIGBLIFNames(graph)
	live, nodes := 0, 0
	for _, object := range graph.objects {
		if object.Type != AigNone {
			live++
		}
		if object.Type == AigAND {
			nodes++
		}
	}
	if err = writer.add("# BLIF file written by procedure Saig_ManDumpBlif()\n"); err != nil {
		return nil, err
	}
	if err = writer.add("# If unedited, this file can be read by Saig_ManReadBlif()\n"); err != nil {
		return nil, err
	}
	if err = writer.add(fmt.Sprintf("# AIG stats: pi=%d po=%d reg=%d and=%d obj=%d maxid=%d\n", graph.PrimaryInputCount(), graph.PrimaryOutputCount(), graph.RegisterCount(), nodes, live, graph.ObjectCount())); err != nil {
		return nil, err
	}
	if err = writer.add(".model " + graph.Name + "\n.inputs"); err != nil {
		return nil, err
	}
	for _, id := range graph.cis[:graph.PrimaryInputCount()] {
		name, nameErr := names.name(id)
		if nameErr != nil {
			return nil, nameErr
		}
		if err = writer.add(" " + name); err != nil {
			return nil, err
		}
	}
	if err = writer.add("\n.outputs"); err != nil {
		return nil, err
	}
	for _, id := range graph.cos[:graph.PrimaryOutputCount()] {
		name, nameErr := names.name(id)
		if nameErr != nil {
			return nil, nameErr
		}
		if err = writer.add(" " + name); err != nil {
			return nil, err
		}
	}
	if err = writer.add("\n"); err != nil {
		return nil, err
	}
	for index := 0; index < graph.RegisterCount(); index++ {
		li, nameErr := names.name(graph.cos[graph.PrimaryOutputCount()+index])
		if nameErr != nil {
			return nil, nameErr
		}
		lo, nameErr := names.name(graph.cis[graph.PrimaryInputCount()+index])
		if nameErr != nil {
			return nil, nameErr
		}
		if err = writer.add(".latch " + li + " " + lo + " 0\n"); err != nil {
			return nil, err
		}
	}
	if graph.objects[0].RefCount != 0 {
		constant, _ := names.name(0)
		if err = writer.add(".names " + constant + "\n 1\n"); err != nil {
			return nil, err
		}
	}
	for id, object := range graph.objects {
		if object.Type != AigAND {
			continue
		}
		first, nameErr := names.name(object.Fanin0.ID())
		if nameErr != nil {
			return nil, nameErr
		}
		second, nameErr := names.name(object.Fanin1.ID())
		if nameErr != nil {
			return nil, nameErr
		}
		output, _ := names.name(AigID(id))
		pattern := []byte{'1', '1'}
		if object.Fanin0.Complemented() {
			pattern[0] = '0'
		}
		if object.Fanin1.Complemented() {
			pattern[1] = '0'
		}
		if err = writer.add(fmt.Sprintf(".names %s %s %s\n%s 1\n", first, second, output, pattern)); err != nil {
			return nil, err
		}
	}
	for _, id := range graph.cos {
		object := graph.objects[id]
		fanin, nameErr := names.name(object.Fanin0.ID())
		if nameErr != nil {
			return nil, nameErr
		}
		output, _ := names.name(id)
		pattern := byte('1')
		if object.Fanin0.Complemented() {
			pattern = '0'
		}
		if err = writer.add(fmt.Sprintf(".names %s %s\n%c 1\n", fanin, output, pattern)); err != nil {
			return nil, err
		}
	}
	if err = writer.add(".end\n"); err != nil {
		return nil, err
	}
	return bytes.Clone(writer.buffer.Bytes()), nil
}

// WriteSAIGBLIF is the path-based counterpart of Saig_ManDumpBlif.
func WriteSAIGBLIF(graph *AigManager, path string, options SAIGBLIFOptions) error {
	if path == "" {
		return ErrDimension
	}
	data, err := MarshalSAIGBLIF(graph, options)
	if err != nil {
		return err
	}
	return os.WriteFile(path, data, 0o666)
}

func saigBLIFLines(data []byte, budget *saigBLIFBudget) ([][]string, error) {
	physical := bytes.Split(data, []byte{'\n'})
	lines := make([][]string, 0, len(physical))
	tokens := 0
	for _, raw := range physical {
		trimmed := strings.TrimSpace(string(raw))
		if trimmed == "" || strings.HasPrefix(trimmed, "#") {
			continue
		}
		fields := strings.Fields(trimmed)
		if len(fields) > budget.options.MaxTokens-tokens {
			return nil, ErrLimit
		}
		tokens += len(fields)
		for _, field := range fields {
			if len(field) > budget.options.MaxTokenBytes {
				return nil, ErrLimit
			}
		}
		if err := budget.step(len(fields) + 1); err != nil {
			return nil, err
		}
		lines = append(lines, fields)
	}
	return lines, nil
}

func saigBLIFNumber(token string) (kind string, number int, err error) {
	for _, prefix := range []string{"pi", "po", "lo", "li", "n"} {
		if !strings.HasPrefix(token, prefix) {
			continue
		}
		digits := token[len(prefix):]
		if digits == "" {
			return "", 0, ErrFormat
		}
		value, parseErr := strconv.ParseUint(digits, 10, 31)
		if parseErr != nil {
			return "", 0, ErrFormat
		}
		return prefix, int(value), nil
	}
	return "", 0, ErrFormat
}

// SAIGBLIFReadNumber is the checked counterpart of Saig_ManReadNumber.
func SAIGBLIFReadNumber(token string) (int, error) {
	_, number, err := saigBLIFNumber(token)
	return number, err
}

func saigBLIFResolve(graph *AigManager, primaryInputs, registers int, nodes map[int]AigLit, token string) (AigLit, error) {
	kind, number, err := saigBLIFNumber(token)
	if err != nil {
		return InvalidAigLit, err
	}
	switch kind {
	case "n":
		literal, exists := nodes[number]
		if !exists {
			return InvalidAigLit, ErrFormat
		}
		return literal, nil
	case "pi":
		if number < 0 || number >= primaryInputs {
			return InvalidAigLit, ErrFormat
		}
		return AigLit(graph.cis[number]) << 1, nil
	case "lo":
		if number < 0 || number >= registers {
			return InvalidAigLit, ErrFormat
		}
		return AigLit(graph.cis[primaryInputs+number]) << 1, nil
	default:
		return InvalidAigLit, ErrFormat
	}
}

func saigBLIFHeaderNames(fields []string, kind string) error {
	for index, token := range fields[1:] {
		got, number, err := saigBLIFNumber(token)
		if err != nil || got != kind || number != index {
			return ErrFormat
		}
	}
	return nil
}

// ParseSAIGBLIF accepts the deterministic structural dialect emitted by
// MarshalSAIGBLIF and returns a checked, owned sequential AIG.
func ParseSAIGBLIF(data []byte, spec string, options SAIGBLIFOptions) (*AigManager, error) {
	options, err := normalizeSAIGBLIFOptions(options)
	if err != nil {
		return nil, err
	}
	if len(data) > options.MaxInputBytes {
		return nil, ErrLimit
	}
	budget := &saigBLIFBudget{options: options}
	lines, err := saigBLIFLines(data, budget)
	if err != nil {
		return nil, err
	}
	if len(lines) < 4 || len(lines[0]) != 2 || lines[0][0] != ".model" || len(strings.Fields(lines[0][1])) != 1 || lines[1][0] != ".inputs" || lines[2][0] != ".outputs" {
		return nil, ErrFormat
	}
	if err = saigBLIFHeaderNames(lines[1], "pi"); err != nil {
		return nil, err
	}
	if err = saigBLIFHeaderNames(lines[2], "po"); err != nil {
		return nil, err
	}
	primaryInputs, primaryOutputs := len(lines[1])-1, len(lines[2])-1
	index, registers := 3, 0
	for index < len(lines) && lines[index][0] == ".latch" {
		fields := lines[index]
		if len(fields) != 4 || fields[3] != "0" {
			return nil, ErrFormat
		}
		liKind, liNumber, liErr := saigBLIFNumber(fields[1])
		loKind, loNumber, loErr := saigBLIFNumber(fields[2])
		if liErr != nil || loErr != nil || liKind != "li" || loKind != "lo" || liNumber != registers || loNumber != registers {
			return nil, ErrFormat
		}
		registers++
		index++
	}
	if primaryInputs+registers+1 > options.MaxObjects {
		return nil, ErrLimit
	}
	graph := NewAigManager()
	graph.Name, graph.Spec = lines[0][1], spec
	for ci := 0; ci < primaryInputs+registers; ci++ {
		if err = budget.step(1); err != nil {
			return nil, err
		}
		if _, err = graph.CreateCI(); err != nil {
			return nil, err
		}
	}
	nodes := map[int]AigLit{0: AigConst1}
	outputs := 0
	for index < len(lines) && lines[index][0] == ".names" {
		header := lines[index]
		if index+1 >= len(lines) {
			return nil, ErrFormat
		}
		cube := lines[index+1]
		index += 2
		if err = budget.step(len(header) + len(cube)); err != nil {
			return nil, err
		}
		switch len(header) {
		case 2:
			kind, number, numberErr := saigBLIFNumber(header[1])
			if numberErr != nil || kind != "n" || number != 0 || len(cube) != 1 || cube[0] != "1" {
				return nil, ErrFormat
			}
		case 3:
			if len(cube) != 2 || len(cube[0]) != 1 || cube[1] != "1" || cube[0][0] != '0' && cube[0][0] != '1' {
				return nil, ErrFormat
			}
			fanin, resolveErr := saigBLIFResolve(graph, primaryInputs, registers, nodes, header[1])
			if resolveErr != nil {
				return nil, resolveErr
			}
			fanin = fanin.NotIf(cube[0][0] == '0')
			kind, number, numberErr := saigBLIFNumber(header[2])
			wantKind, wantNumber := "po", outputs
			if outputs >= primaryOutputs {
				wantKind, wantNumber = "li", outputs-primaryOutputs
			}
			if numberErr != nil || kind != wantKind || number != wantNumber || outputs >= primaryOutputs+registers {
				return nil, ErrFormat
			}
			if _, err = graph.CreateCO(fanin); err != nil {
				return nil, err
			}
			outputs++
		case 4:
			if len(cube) != 2 || len(cube[0]) != 2 || cube[1] != "1" || strings.IndexFunc(cube[0], func(r rune) bool { return r != '0' && r != '1' }) >= 0 {
				return nil, ErrFormat
			}
			left, resolveErr := saigBLIFResolve(graph, primaryInputs, registers, nodes, header[1])
			if resolveErr != nil {
				return nil, resolveErr
			}
			right, resolveErr := saigBLIFResolve(graph, primaryInputs, registers, nodes, header[2])
			if resolveErr != nil {
				return nil, resolveErr
			}
			left = left.NotIf(cube[0][0] == '0')
			right = right.NotIf(cube[0][1] == '0')
			kind, number, numberErr := saigBLIFNumber(header[3])
			if numberErr != nil || kind != "n" {
				return nil, ErrFormat
			}
			literal, andErr := graph.And(left, right)
			if andErr != nil {
				return nil, andErr
			}
			if literal.Complemented() {
				return nil, ErrFormat
			}
			if _, exists := nodes[number]; exists {
				return nil, ErrFormat
			}
			nodes[number] = literal
		default:
			return nil, ErrFormat
		}
		if graph.ObjectCount() > options.MaxObjects {
			return nil, ErrLimit
		}
	}
	if index != len(lines)-1 || len(lines[index]) != 1 || lines[index][0] != ".end" || outputs != primaryOutputs+registers {
		return nil, ErrFormat
	}
	if err = graph.SetRegisterCount(registers); err != nil {
		return nil, err
	}
	if err = graph.Check(); err != nil {
		return nil, err
	}
	return graph, nil
}

// ReadSAIGBLIF is the path-based counterpart of Saig_ManReadBlif.
func ReadSAIGBLIF(path string, options SAIGBLIFOptions) (*AigManager, error) {
	options, err := normalizeSAIGBLIFOptions(options)
	if err != nil {
		return nil, err
	}
	if path == "" {
		return nil, ErrDimension
	}
	info, err := os.Stat(path)
	if err != nil {
		return nil, err
	}
	if info.Size() < 0 || info.Size() > int64(options.MaxInputBytes) {
		return nil, ErrLimit
	}
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}
	return ParseSAIGBLIF(data, path, options)
}
