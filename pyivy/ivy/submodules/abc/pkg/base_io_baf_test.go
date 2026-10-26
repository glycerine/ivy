package abc

// Go-only tests for src/base/io/ioReadBaf.c and ioWriteBaf.c.

import (
	"bytes"
	"encoding/binary"
	"errors"
	"slices"
	"strconv"
	"sync"
	"testing"
)

func abcBAFTestModel(t testing.TB, overlappingOutput bool) *ABCBAFModel {
	t.Helper()
	graph := NewAigManager()
	a, err := graph.CreateCI()
	if err != nil {
		t.Fatal(err)
	}
	b, _ := graph.CreateCI()
	state, _ := graph.CreateCI()
	first, err := graph.And(a, b.Not())
	if err != nil {
		t.Fatal(err)
	}
	next, err := graph.And(first, state)
	if err != nil {
		t.Fatal(err)
	}
	if _, err = graph.CreateCO(first); err != nil {
		t.Fatal(err)
	}
	if _, err = graph.CreateCO(state.Not()); err != nil {
		t.Fatal(err)
	}
	if _, err = graph.CreateCO(next); err != nil {
		t.Fatal(err)
	}
	if err = graph.SetRegisterCount(1); err != nil {
		t.Fatal(err)
	}
	output0 := "out0"
	if overlappingOutput {
		output0 = "a"
	}
	model, err := NewABCBAFModel("sequential", []string{"a", "b"}, []string{output0, "out1"}, []ABCBAFLatch{{Name: "reg", InputName: "reg_in", OutputName: "reg_out", Init: ABCInitDC}}, graph, ABCBAFOptions{})
	if err != nil {
		t.Fatal(err)
	}
	return model
}

func abcBAFCompareGraphs(t testing.TB, first, second *AigManager) {
	t.Helper()
	if first.InputCount() != second.InputCount() || first.OutputCount() != second.OutputCount() || first.RegisterCount() != second.RegisterCount() {
		t.Fatalf("graph dimensions: (%d,%d,%d) != (%d,%d,%d)", first.InputCount(), first.OutputCount(), first.RegisterCount(), second.InputCount(), second.OutputCount(), second.RegisterCount())
	}
	if first.InputCount() > 12 {
		t.Fatal("test graph too large for exhaustive comparison")
	}
	for assignment := 0; assignment < 1<<first.InputCount(); assignment++ {
		inputs := make([]bool, first.InputCount())
		for index := range inputs {
			inputs[index] = assignment>>index&1 != 0
		}
		left, leftErr := first.Evaluate(inputs)
		right, rightErr := second.Evaluate(inputs)
		if leftErr != nil || rightErr != nil || !slices.Equal(left, right) {
			t.Fatalf("assignment %d: %v/%v, %v/%v", assignment, left, leftErr, right, rightErr)
		}
	}
}

type abcBAFIndependent struct {
	name                                     string
	inputs, outputs, registers, ands, bodyAt int
	fields                                   []string
}

func decodeABCBAFHeaderIndependent(t testing.TB, data []byte) abcBAFIndependent {
	t.Helper()
	cursor := 0
	for cursor < len(data) && data[cursor] == '#' {
		newline := bytes.IndexByte(data[cursor:], '\n')
		if newline < 0 {
			t.Fatal("unterminated comment")
		}
		cursor += newline + 1
	}
	read := func() string {
		end := bytes.IndexByte(data[cursor:], 0)
		if end < 0 {
			t.Fatal("missing terminator")
		}
		value := string(data[cursor : cursor+end])
		cursor += end + 1
		return value
	}
	result := abcBAFIndependent{name: read()}
	counts := []*int{&result.inputs, &result.outputs, &result.registers, &result.ands}
	for _, target := range counts {
		value, err := strconv.Atoi(read())
		if err != nil {
			t.Fatal(err)
		}
		*target = value
	}
	for index := 0; index < result.inputs+result.outputs+3*result.registers; index++ {
		result.fields = append(result.fields, read())
	}
	result.bodyAt = cursor
	return result
}

func TestABCBAFModelMarshalParseAndIndependentBody(t *testing.T) {
	model := abcBAFTestModel(t, true)
	data, err := MarshalABCBAFModel(model, ABCBAFOptions{})
	if err != nil {
		t.Fatal(err)
	}
	if !bytes.HasPrefix(data, []byte("# BAF (Binary Aig Format)\n")) {
		t.Fatalf("comment: %q", data[:min(len(data), 40)])
	}
	header := decodeABCBAFHeaderIndependent(t, data)
	if header.name != "sequential" || header.inputs != 2 || header.outputs != 2 || header.registers != 1 || header.ands != 2 {
		t.Fatalf("header: %+v", header)
	}
	wantFields := []string{"a", "b", "a", "out1", "reg", "reg_in", "reg_out"}
	if !slices.Equal(header.fields, wantFields) {
		t.Fatalf("names: %q", header.fields)
	}
	if len(data)-header.bodyAt != (2*header.ands+header.outputs+header.registers)*4 {
		t.Fatal("body size")
	}
	words := make([]uint32, (len(data)-header.bodyAt)/4)
	for index := range words {
		words[index] = binary.LittleEndian.Uint32(data[header.bodyAt+4*index:])
	}
	if !slices.Equal(words, []uint32{2, 5, 6, 8, 8, 7, 43}) {
		t.Fatalf("body words: %v", words)
	}
	parsed, err := ParseABCBAF(data, ABCBAFOptions{})
	if err != nil {
		t.Fatal(err)
	}
	if parsed.Name() != model.Name() || !slices.Equal(parsed.InputNames(), model.InputNames()) || !slices.Equal(parsed.OutputNames(), model.OutputNames()) || !slices.Equal(parsed.Latches(), model.Latches()) {
		t.Fatalf("metadata: %q %q %q %+v", parsed.Name(), parsed.InputNames(), parsed.OutputNames(), parsed.Latches())
	}
	abcBAFCompareGraphs(t, model.graph, parsed.graph)
	again, err := MarshalABCBAFModel(parsed, ABCBAFOptions{})
	if err != nil || !bytes.Equal(data, again) {
		t.Fatalf("unstable bytes: %v\n%x\n%x", err, data, again)
	}
}

func TestABCBAFOwnershipNetworkBridgeAndChecks(t *testing.T) {
	graph := NewAigManager()
	input, _ := graph.CreateCI()
	graph.CreateCO(input)
	inputNames := []string{"i"}
	outputNames := []string{"o"}
	model, err := NewABCBAFModel("owned", inputNames, outputNames, nil, graph, ABCBAFOptions{})
	if err != nil {
		t.Fatal(err)
	}
	inputNames[0], outputNames[0] = "changed", "changed"
	graph.CreateCO(input.Not())
	gotInputs, gotOutputs := model.InputNames(), model.OutputNames()
	gotInputs[0], gotOutputs[0] = "mutated", "mutated"
	if model.InputNames()[0] != "i" || model.OutputNames()[0] != "o" || model.graph.OutputCount() != 1 {
		t.Fatal("model aliases constructor or accessor storage")
	}

	sequential := abcBAFTestModel(t, false)
	network, err := ABCNetworkFromBAFModel(sequential, true, ABCBAFOptions{})
	if err != nil {
		t.Fatal(err)
	}
	if network.Name != "sequential" || network.ObjectCountByType(ABCObjectLatch) != 1 {
		t.Fatalf("network metadata: %q boxes=%d", network.Name, network.ObjectCountByType(ABCObjectLatch))
	}
	latch := network.objects[network.boxes[0]]
	if latch.Init != ABCInitDC || latch.Name != "reg" {
		t.Fatalf("latch: %+v", latch)
	}
	data, err := MarshalABCBAF(network, ABCBAFOptions{})
	if err != nil {
		t.Fatal(err)
	}
	read, err := ReadABCBAF(data, true, ABCBAFOptions{})
	if err != nil {
		t.Fatal(err)
	}
	left, _ := ABCNetworkToAIG(network, ABCFunctionalOptions{})
	right, _ := ABCNetworkToAIG(read, ABCFunctionalOptions{})
	abcBAFCompareGraphs(t, left.Graph, right.Graph)

	overlap := abcBAFTestModel(t, true)
	if converted, err := ABCNetworkFromBAFModel(overlap, true, ABCBAFOptions{}); converted != nil || !errors.Is(err, ErrUnsupported) {
		t.Fatalf("overlapping names: network=%v err=%v", converted, err)
	}
}

func TestABCBAFMalformedUnsupportedLimitsAndCancellation(t *testing.T) {
	model := abcBAFTestModel(t, false)
	data, err := MarshalABCBAFModel(model, ABCBAFOptions{})
	if err != nil {
		t.Fatal(err)
	}
	malformed := [][]byte{
		nil,
		[]byte("# no newline"),
		[]byte("name\x001\x000\x000\x000\x00"),
		[]byte("name\x00x\x000\x000\x000\x00"),
		append(slices.Clone(data), 0),
	}
	forward := slices.Clone(data)
	header := decodeABCBAFHeaderIndependent(t, forward)
	binary.LittleEndian.PutUint32(forward[header.bodyAt:], uint32(1000<<1))
	malformed = append(malformed, forward)
	for _, input := range malformed {
		if parsed, parseErr := ParseABCBAF(input, ABCBAFOptions{}); parsed != nil || parseErr == nil {
			t.Fatalf("malformed accepted: %x model=%v", input, parsed)
		}
	}
	for _, options := range []ABCBAFOptions{
		{MaxInputBytes: len(data) - 1}, {MaxObjects: 1}, {MaxNameBytes: 1}, {MaxWork: 1}, {MaxBytes: 1}, {Stop: func() bool { return true }},
	} {
		if parsed, parseErr := ParseABCBAF(data, options); parsed != nil || !(errors.Is(parseErr, ErrLimit) || errors.Is(parseErr, ErrCancelled)) {
			t.Fatalf("parse limit %+v: model=%v err=%v", options, parsed, parseErr)
		}
	}
	for _, options := range []ABCBAFOptions{
		{MaxOutputBytes: len(data) - 1}, {MaxObjects: 1}, {MaxNameBytes: 1}, {MaxWork: 1}, {MaxBytes: 1}, {Stop: func() bool { return true }},
	} {
		if output, writeErr := MarshalABCBAFModel(model, options); output != nil || !(errors.Is(writeErr, ErrLimit) || errors.Is(writeErr, ErrCancelled)) {
			t.Fatalf("write limit %+v: output=%x err=%v", options, output, writeErr)
		}
	}
	if _, err := NewABCBAFModel("bad\x00name", []string{"a", "b"}, []string{"o0", "o1"}, model.Latches(), model.graph, ABCBAFOptions{}); !errors.Is(err, ErrFormat) {
		t.Fatalf("NUL name: %v", err)
	}
	bufferGraph := NewAigManager()
	ci, _ := bufferGraph.CreateCI()
	buffer, _ := bufferGraph.CreateBuffer(ci)
	bufferGraph.CreateCO(buffer)
	bufferModel, err := NewABCBAFModel("buffer", []string{"i"}, []string{"o"}, nil, bufferGraph, ABCBAFOptions{})
	if err != nil {
		t.Fatal(err)
	}
	if output, err := MarshalABCBAFModel(bufferModel, ABCBAFOptions{}); output != nil || !errors.Is(err, ErrUnsupported) {
		t.Fatalf("buffer graph: %x, %v", output, err)
	}
}

func TestABCBAFConcurrentFrozenReaders(t *testing.T) {
	model := abcBAFTestModel(t, true)
	want, err := MarshalABCBAFModel(model, ABCBAFOptions{})
	if err != nil {
		t.Fatal(err)
	}
	var wait sync.WaitGroup
	failures := make(chan error, 16)
	for worker := 0; worker < 16; worker++ {
		wait.Add(1)
		go func() {
			defer wait.Done()
			got, writeErr := MarshalABCBAFModel(model, ABCBAFOptions{})
			if writeErr != nil {
				failures <- writeErr
				return
			}
			if !bytes.Equal(got, want) {
				failures <- ErrInvariant
			}
		}()
	}
	wait.Wait()
	close(failures)
	for err := range failures {
		t.Error(err)
	}
}

func FuzzABCBAFRoundTrip(f *testing.F) {
	model := abcBAFTestModel(f, true)
	seed, err := MarshalABCBAFModel(model, ABCBAFOptions{})
	if err != nil {
		f.Fatal(err)
	}
	f.Add(seed)
	f.Add([]byte("tiny\x000\x000\x000\x000\x00"))
	f.Fuzz(func(t *testing.T, data []byte) {
		options := ABCBAFOptions{MaxInputBytes: 4096, MaxOutputBytes: 16384, MaxObjects: 128, MaxNameBytes: 4096, MaxWork: 32768, MaxBytes: 65536}
		if len(data) > options.MaxInputBytes {
			t.Skip()
		}
		first, err := ParseABCBAF(data, options)
		if err != nil {
			return
		}
		encoded, err := MarshalABCBAFModel(first, options)
		if err != nil {
			if errors.Is(err, ErrLimit) || errors.Is(err, ErrUnsupported) {
				return
			}
			t.Fatal(err)
		}
		second, err := ParseABCBAF(encoded, options)
		if err != nil {
			t.Fatalf("encoded model rejected: %v", err)
		}
		if first.Name() != second.Name() || !slices.Equal(first.InputNames(), second.InputNames()) || !slices.Equal(first.OutputNames(), second.OutputNames()) || !slices.Equal(first.Latches(), second.Latches()) {
			t.Fatal("metadata changed")
		}
		if first.graph.InputCount() != second.graph.InputCount() || first.graph.OutputCount() != second.graph.OutputCount() {
			t.Fatal("graph dimensions changed")
		}
		for pattern := 0; pattern < 16; pattern++ {
			inputs := make([]bool, first.graph.InputCount())
			for index := range inputs {
				inputs[index] = pattern>>(index&3)&1 != 0
			}
			left, leftErr := first.graph.Evaluate(inputs)
			right, rightErr := second.graph.Evaluate(inputs)
			if leftErr != nil || rightErr != nil || !slices.Equal(left, right) {
				t.Fatalf("semantic mismatch: %v/%v %v/%v", left, leftErr, right, rightErr)
			}
		}
	})
}
