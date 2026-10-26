package abc

// Go-only tests for src/aig/ioa.

import (
	"bytes"
	"errors"
	"os"
	"path/filepath"
	"slices"
	"testing"
	"time"
)

func aigIOAIndependentEvaluate(graph *AigManager, inputs []bool) []bool {
	values := make([]bool, len(graph.objects))
	values[0] = true
	for index, id := range graph.cis {
		values[id] = inputs[index]
	}
	read := func(literal AigLit) bool {
		value := values[literal.ID()]
		if literal.Complemented() {
			value = !value
		}
		return value
	}
	for id, object := range graph.objects {
		switch object.Type {
		case AigAND:
			values[id] = read(object.Fanin0) && read(object.Fanin1)
		case AigBuffer, AigCO:
			values[id] = read(object.Fanin0)
		}
	}
	result := make([]bool, len(graph.cos))
	for index, id := range graph.cos {
		result[index] = values[id]
	}
	return result
}

func aigIOAFixture() *AigManager {
	graph := NewAigManager()
	input, _ := graph.CreateCI()
	latch, _ := graph.CreateCI()
	conjunction, _ := graph.And(input, latch.Not())
	graph.CreateCO(conjunction)
	graph.CreateCO(input.Not()) // internal ABC constraint polarity
	graph.CreateCO(conjunction.Not())
	graph.SetRegisterCount(1)
	graph.SetConstraintCount(1)
	graph.Name = "fixture"
	return graph
}

func TestAigIOAExactStandardCompactAndPolarity(t *testing.T) {
	graph := aigIOAFixture()
	memory, err := IOAWriteAigerIntoMemory(graph, AigIOAOptions{Check: true, Compact: true})
	wantMemory := append([]byte("aig 3 1 1 2 1\n7\n6\n3\n"), 1, 3)
	wantMemory = append(wantMemory, []byte("cnfixture\x00")...)
	if err != nil || !bytes.Equal(memory, wantMemory) {
		t.Fatalf("memory\n got %q\nwant %q\nerr %v", memory, wantMemory, err)
	}
	memoryStr, err := IOAWriteAigerIntoMemoryStr(graph, AigIOAOptions{})
	if err != nil || !bytes.Equal(memoryStr, wantMemory[:len(wantMemory)-len("nfixture\x00")]) {
		t.Fatal("memory str", memoryStr, err)
	}
	standard, err := IOAMarshalAiger(graph, AigIOAOptions{Check: true})
	wantStandard := append([]byte("aig 3 1 1 0 1 1 1\n7\n6\n2\n"), 1, 3)
	wantStandard = append(wantStandard, []byte("cnfixture\x00")...)
	if err != nil || !bytes.Equal(standard, wantStandard) {
		t.Fatalf("standard\n got %q\nwant %q\nerr %v", standard, wantStandard, err)
	}
	compact, err := IOAMarshalAiger(graph, AigIOAOptions{Compact: true, Check: true})
	wantCompact := append([]byte("aig2 3 1 1 0 1 1 1\n"), 7, 3, 9, 1, 3)
	wantCompact = append(wantCompact, []byte("cnfixture\x00")...)
	if err != nil || !bytes.Equal(compact, wantCompact) {
		t.Fatalf("compact\n got %q\nwant %q\nerr %v", compact, wantCompact, err)
	}
	for _, encoded := range [][]byte{standard, compact} {
		round, readErr := IOAReadAigerFromMemory(encoded, AigIOAOptions{Check: true})
		if readErr != nil || round.Name != "fixture" || round.RegisterCount() != 1 || round.ConstraintCount() != 1 {
			t.Fatal(readErr, round)
		}
		for assignment := 0; assignment < 4; assignment++ {
			inputs := []bool{assignment&1 != 0, assignment&2 != 0}
			first := aigIOAIndependentEvaluate(graph, inputs)
			second := aigIOAIndependentEvaluate(round, inputs)
			if !slices.Equal(first, second) {
				t.Fatal(assignment, first, second)
			}
		}
	}
}

func TestAigIOAVarintsDriverCodecAndUtilities(t *testing.T) {
	for _, value := range []uint32{0, 1, 127, 128, 16384, 1<<31 - 1, ^uint32(0)} {
		encoded := IOAEncodeAigerUint(value)
		decoded, used, err := IOADecodeAigerUint(encoded, 0)
		if err != nil || decoded != value || used != len(encoded) {
			t.Fatal(value, encoded, decoded, used, err)
		}
	}
	literals := []uint32{7, 6, 2, 1000, 999, 0}
	encoded, err := IOAEncodeAigerLiterals(literals, 100)
	decoded, used, decodeErr := IOADecodeAigerLiterals(encoded, len(literals), 100)
	if err != nil || decodeErr != nil || used != len(encoded) || !slices.Equal(decoded, literals) {
		t.Fatal(encoded, decoded, used, err, decodeErr)
	}
	if graphLiterals, graphErr := IOAWriteAigerLiterals(aigIOAFixture(), AigIOAOptions{}); graphErr != nil || !slices.Equal(graphLiterals, []uint32{7, 6, 3}) {
		t.Fatal(graphLiterals, graphErr)
	}
	if IOAFileNameGeneric("dir/a.b.aig") != "dir/a.b" || IOAFileNameGenericAppend("dir/a.b.aig", ".out") != "a.b.out" || IOAFileNameGenericAppend("", "dir/x.aig") != "x.aig" {
		t.Fatal("filename utilities")
	}
	stamp := time.Date(2026, time.October, 8, 9, 7, 6, 0, time.UTC)
	if IOATimeStampAt(stamp) != "Thu Oct  8 09:07:06 2026" || IOATimeStamp() == "" {
		t.Fatal(IOATimeStampAt(stamp))
	}
}

func TestAigIOAFilesOwnershipLimitsCancellationAndMalformed(t *testing.T) {
	graph := aigIOAFixture()
	directory := t.TempDir()
	path := filepath.Join(directory, "model.aig")
	stamp := time.Date(2026, 1, 2, 3, 4, 5, 0, time.UTC)
	if err := IOAWriteAiger(graph, path, AigIOAOptions{Timestamp: stamp}); err != nil {
		t.Fatal(err)
	}
	bufferPath := filepath.Join(directory, "buffer.aig")
	if err := IOAWriteAigerBufferTest(graph, bufferPath, AigIOAOptions{Compact: true, Timestamp: stamp}); err != nil {
		t.Fatal(err)
	}
	if bufferData, readErr := os.ReadFile(bufferPath); readErr != nil || !bytes.HasPrefix(bufferData, []byte("aig 3 1 1 2 1\n")) {
		t.Fatal("buffer writer did not preserve standard memory format", readErr)
	}
	size, err := IOAFileSize(path)
	if err != nil || size == 0 {
		t.Fatal(size, err)
	}
	read, err := IOAReadAiger(path, AigIOAOptions{Check: true})
	if err != nil || read.Name != IOAFileNameGeneric(path) || read.Spec != path {
		t.Fatal(read, err)
	}
	data, _ := os.ReadFile(path)
	data[0] = 'X'
	if read.Name == "" {
		t.Fatal("reader retained input")
	}
	if err = os.WriteFile(path, []byte("sentinel"), 0o600); err != nil {
		t.Fatal(err)
	}
	if err = IOAWriteAiger(graph, path, AigIOAOptions{Stop: func() bool { return true }}); !errors.Is(err, ErrCancelled) {
		t.Fatal(err)
	}
	if unchanged, readErr := os.ReadFile(path); readErr != nil || string(unchanged) != "sentinel" {
		t.Fatal("cancelled file write was not atomic", string(unchanged), readErr)
	}
	for _, options := range []AigIOAOptions{{MaxBytes: 1}, {MaxVariables: 1}, {MaxOutputs: 1}, {MaxBytes: -1}, {Stop: func() bool { return true }}} {
		if _, err = IOAWriteAigerIntoMemory(graph, options); err == nil {
			t.Fatal("write limit", options)
		}
	}
	valid, _ := IOAMarshalAiger(graph, AigIOAOptions{})
	for _, options := range []AigIOAOptions{{MaxBytes: 1}, {MaxVariables: 1}, {MaxOutputs: 1}, {Stop: func() bool { return true }}} {
		if _, err = IOAReadAigerFromMemory(valid, options); err == nil {
			t.Fatal("read limit", options)
		}
	}
	malformed := [][]byte{
		nil, []byte("aag 0 0 0 0 0\n"), []byte("aig 0 0 0 0 0"), []byte("aig 1 0 0 0 0\n"),
		[]byte("aig 0 0 0 1 0\n2\n"), []byte("aig 1 1 0 1 0 1\n2\n2\n"),
		append([]byte("aig 1 0 0 1 1\n2\n"), 0, 0),
		append([]byte("aig2 1 1 0 1 0\n"), 4),
		append([]byte("aig 0 0 0 0 0\ncnmissing"), byte('x')),
	}
	for _, input := range malformed {
		if _, readErr := IOAReadAigerFromMemory(input, AigIOAOptions{}); readErr == nil {
			t.Fatalf("accepted malformed %q", input)
		}
	}
	for _, complete := range [][]byte{
		append([]byte("aig 3 1 1 0 1 1 1\n7\n6\n2\n"), 1, 3),
		append([]byte("aig2 3 1 1 0 1 1 1\n"), 7, 3, 9, 1, 3),
	} {
		for end := 0; end < len(complete); end++ {
			if _, readErr := IOAReadAigerFromMemory(complete[:end], AigIOAOptions{}); readErr == nil {
				t.Fatal("accepted truncation", end, complete)
			}
		}
	}
	if _, _, err = IOADecodeAigerUint([]byte{128, 128, 128, 128, 16}, 0); err == nil {
		t.Fatal("varint overflow")
	}
	if _, err = IOAEncodeAigerLiterals([]uint32{0, ^uint32(0)}, 100); !errors.Is(err, ErrLimit) {
		t.Fatal(err)
	}
	if _, _, err = IOADecodeAigerLiterals([]byte{0}, 2, 1); !errors.Is(err, ErrLimit) {
		t.Fatal(err)
	}
	zeroInit := append([]byte("aig 2 1 1 1 0\n0 0\n2\n"), 'c')
	if parsed, readErr := IOAReadAigerFromMemory(zeroInit, AigIOAOptions{Check: true}); readErr != nil || parsed.RegisterCount() != 1 {
		t.Fatal("zero latch initialization", parsed, readErr)
	}
	for _, unsupported := range [][]byte{
		append([]byte("aig 2 1 1 1 0\n0 1\n2\n"), 'c'),
		append([]byte("aig 2 1 1 1 0\n0 4\n2\n"), 'c'),
	} {
		if _, readErr := IOAReadAigerFromMemory(unsupported, AigIOAOptions{}); !errors.Is(readErr, ErrUnsupported) {
			t.Fatal("nonzero latch initialization", readErr)
		}
	}
}

func TestAigIOAIndependentSimulationRoundTrips(t *testing.T) {
	state := uint64(0x9e3779b97f4a7c15)
	for execution := 0; execution < 5000; execution++ {
		state ^= state << 7
		state ^= state >> 9
		state ^= state << 8
		graph := NewAigManager()
		inputs := 1 + int(state%4)
		literals := []AigLit{AigConst0, AigConst1}
		for index := 0; index < inputs; index++ {
			literal, _ := graph.CreateCI()
			literals = append(literals, literal)
		}
		for index := 0; index < 1+int(state%6); index++ {
			state = state*6364136223846793005 + 1442695040888963407
			left := literals[int(state%uint64(len(literals)))].NotIf(state>>8&1 != 0)
			state = state*6364136223846793005 + 1442695040888963407
			right := literals[int(state%uint64(len(literals)))].NotIf(state>>9&1 != 0)
			literal, _ := graph.And(left, right)
			literals = append(literals, literal)
		}
		for output := 0; output < 1+int(state%3); output++ {
			state = state*6364136223846793005 + 1442695040888963407
			driver := literals[int(state%uint64(len(literals)))].NotIf(state>>10&1 != 0)
			graph.CreateCO(driver)
		}
		encoded, err := IOAMarshalAiger(graph, AigIOAOptions{Compact: execution&1 != 0, MaxBytes: 1 << 16, MaxVariables: 128, MaxOutputs: 16})
		if err != nil {
			t.Fatal(execution, err)
		}
		round, err := IOAReadAigerFromMemory(encoded, AigIOAOptions{Check: true, MaxBytes: 1 << 16, MaxVariables: 128, MaxOutputs: 16})
		if err != nil {
			t.Fatal(execution, err)
		}
		assignment := make([]bool, inputs)
		for pattern := 0; pattern < 8; pattern++ {
			for index := range assignment {
				assignment[index] = pattern&(1<<uint(index)) != 0
			}
			want := aigIOAIndependentEvaluate(graph, assignment)
			got := aigIOAIndependentEvaluate(round, assignment)
			public, publicErr := round.Evaluate(assignment)
			if publicErr != nil || !slices.Equal(got, want) || !slices.Equal(public, want) {
				t.Fatal(execution, pattern, got, want)
			}
		}
	}
}

func FuzzAigIOAReadWrite(f *testing.F) {
	graph := aigIOAFixture()
	standard, _ := IOAMarshalAiger(graph, AigIOAOptions{})
	compact, _ := IOAMarshalAiger(graph, AigIOAOptions{Compact: true})
	f.Add(standard, false)
	f.Add(compact, true)
	f.Add([]byte("aig 0 0 0 0 0\nc"), false)
	f.Fuzz(func(t *testing.T, data []byte, compactOutput bool) {
		options := AigIOAOptions{MaxBytes: 1 << 16, MaxVariables: 128, MaxOutputs: 128, Check: true}
		graph, err := IOAReadAigerFromMemory(data, options)
		if err != nil {
			return
		}
		encoded, err := IOAMarshalAiger(graph, AigIOAOptions{Compact: compactOutput, MaxBytes: 1 << 16, MaxVariables: 128, MaxOutputs: 128})
		if err != nil {
			t.Fatal(err)
		}
		round, err := IOAReadAigerFromMemory(encoded, options)
		if err != nil {
			t.Fatal(err)
		}
		if graph.InputCount() != round.InputCount() || graph.OutputCount() != round.OutputCount() || graph.RegisterCount() != round.RegisterCount() || graph.ConstraintCount() != round.ConstraintCount() {
			t.Fatal("dimension round trip")
		}
		assignment := make([]bool, graph.InputCount())
		for pattern := 0; pattern < 16; pattern++ {
			for index := range assignment {
				assignment[index] = pattern&(1<<uint(index%4)) != 0
			}
			want := aigIOAIndependentEvaluate(graph, assignment)
			got := aigIOAIndependentEvaluate(round, assignment)
			if !slices.Equal(got, want) {
				t.Fatal(pattern, got, want)
			}
		}
	})
}
