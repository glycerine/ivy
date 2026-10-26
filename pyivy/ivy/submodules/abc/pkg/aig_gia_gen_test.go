package abc

// Stand-alone Go model, dataset, and fuzz tests for src/aig/gia/giaGen.c.

import (
	"encoding/binary"
	"encoding/json"
	"errors"
	"fmt"
	"slices"
	"strings"
	"testing"
)

func giaGenerationGraph(t testing.TB) (*GiaManager, GiaLit, GiaLit, GiaLit) {
	t.Helper()
	graph := NewGiaManager()
	graph.Name = "generation"
	a, err := graph.AppendCI()
	if err != nil {
		t.Fatal(err)
	}
	b, err := graph.AppendCI()
	if err != nil {
		t.Fatal(err)
	}
	root, err := graph.AppendAnd(a.Not(), b)
	if err != nil {
		t.Fatal(err)
	}
	if _, err = graph.AppendCO(root.Not()); err != nil {
		t.Fatal(err)
	}
	return graph, a, b, root
}

func TestGiaInputWordSimulationUsesIndependentBooleanSemantics(t *testing.T) {
	graph, a, b, root := giaGenerationGraph(t)
	inputs := []uint64{0x0123456789abcdef, 0xf0f00f0faa55aa55, 0xcccc3333a5a55a5a, 0x9696696996696996}
	original := slices.Clone(inputs)
	simulation, err := graph.SimulateInputWords(inputs, GiaPackedSimulationOptions{Words: 2})
	if err != nil {
		t.Fatal(err)
	}
	for _, test := range []struct {
		id   GiaID
		want []uint64
	}{
		{a.ID(), inputs[:2]},
		{b.ID(), inputs[2:]},
		{root.ID(), []uint64{^inputs[0] & inputs[2], ^inputs[1] & inputs[3]}},
		{graph.cos[0], []uint64{^(^inputs[0] & inputs[2]), ^(^inputs[1] & inputs[3])}},
	} {
		got, wordErr := simulation.ObjectWords(test.id)
		if wordErr != nil || !slices.Equal(got, test.want) {
			t.Fatalf("object %d: got %x want %x: %v", test.id, got, test.want, wordErr)
		}
	}
	if !slices.Equal(inputs, original) {
		t.Fatal("caller input was modified")
	}
	inputs[0] = 0
	got, _ := simulation.ObjectWords(a.ID())
	if got[0] != original[0] {
		t.Fatal("simulation aliases caller input")
	}
	for sample := 0; sample < 128; sample++ {
		assignment := []bool{
			original[sample/64]>>uint(sample&63)&1 != 0,
			original[2+sample/64]>>uint(sample&63)&1 != 0,
		}
		want, evalErr := graph.Evaluate(assignment)
		actual, bitErr := simulation.Bit(graph.cos[0], sample)
		if evalErr != nil || bitErr != nil || actual != want[0] {
			t.Fatalf("sample %d: got %v want %v: %v %v", sample, actual, want, evalErr, bitErr)
		}
	}
}

func packedDatasetBit(data []byte, sample, feature, features int) bool {
	position := sample*features + feature
	return data[12+position/8]>>uint(position&7)&1 != 0
}

func TestGiaBinaryDatasetsAreOwnedDeterministicAndSelfChecking(t *testing.T) {
	graph, _, _, _ := giaGenerationGraph(t)
	first, err := GenerateGiaBinaryDatasets(graph, 1, 1, 3, "logic", GiaGenerationOptions{})
	if err != nil {
		t.Fatal(err)
	}
	second, err := GenerateGiaBinaryDatasets(graph, 1, 1, 3, "logic", GiaGenerationOptions{})
	if err != nil || len(first.Files) != 5 || len(second.Files) != 5 {
		t.Fatal(len(first.Files), len(second.Files), err)
	}
	wantNames := []string{"logic_x.train.data", "logic_y.train.data", "logic_x.test.data", "logic_y.test.data", "logic.flist"}
	for index, file := range first.Files {
		if file.Name != wantNames[index] || !slices.Equal(file.Data, second.Files[index].Data) {
			t.Fatalf("file %d: %q deterministic=%v", index, file.Name, slices.Equal(file.Data, second.Files[index].Data))
		}
	}
	if string(first.Files[4].Data) != strings.Join(wantNames[:4], "\n")+"\n" {
		t.Fatalf("file list: %q", first.Files[4].Data)
	}
	for pair := 0; pair < 2; pair++ {
		x, y := first.Files[pair*2].Data, first.Files[pair*2+1].Data
		for _, test := range []struct {
			data     []byte
			features uint32
			length   int
		}{{x, 2, 28}, {y, 1, 20}} {
			if len(test.data) != test.length || binary.LittleEndian.Uint32(test.data) != 2 || binary.LittleEndian.Uint32(test.data[4:]) != 64 || binary.LittleEndian.Uint32(test.data[8:]) != test.features {
				t.Fatalf("bad binary header/length: %d %v", len(test.data), test.data[:12])
			}
		}
		for sample := 0; sample < 64; sample++ {
			assignment := []bool{packedDatasetBit(x, sample, 0, 2), packedDatasetBit(x, sample, 1, 2)}
			want, evalErr := graph.Evaluate(assignment)
			if evalErr != nil || packedDatasetBit(y, sample, 0, 1) != want[0] {
				t.Fatalf("dataset %d sample %d: %v %v", pair, sample, want, evalErr)
			}
		}
	}
	first.Files[0].Data[12] ^= 1
	if slices.Equal(first.Files[0].Data, second.Files[0].Data) {
		t.Fatal("separate calls alias their output")
	}
}

func TestGiaPLADatasetsParseCompareAndStats(t *testing.T) {
	graph, _, _, _ := giaGenerationGraph(t)
	files, err := GenerateGiaPLADatasets(graph, 1, 1, 7, "", GiaGenerationOptions{})
	if err != nil || len(files.Files) != 3 {
		t.Fatal(len(files.Files), err)
	}
	wantNames := []string{"generation.train.pla", "generation.valid.pla", "generation.test.pla"}
	for index, file := range files.Files {
		if file.Name != wantNames[index] || !strings.HasPrefix(string(file.Data), ".i 2\n.o 1\n.p 64\n.type fr\n") || !strings.HasSuffix(string(file.Data), ".e\n") {
			t.Fatalf("file %d: %q\n%s", index, file.Name, file.Data)
		}
		dataset, parseErr := ParseGiaSimulationPLA(file.Data, GiaGenerationOptions{})
		if parseErr != nil || dataset.Inputs != 2 || dataset.Words != 1 || len(dataset.Outputs) != 64 {
			t.Fatal(index, dataset, parseErr)
		}
		comparison, compareErr := graph.CompareSimulationDataset(dataset, GiaGenerationOptions{})
		if compareErr != nil || comparison.Correct != 64 || comparison.Total != 64 || comparison.Inputs != 2 || comparison.Outputs != 1 || comparison.ANDs != 1 || comparison.Level != 1 {
			t.Fatal(index, comparison, compareErr)
		}
		for sample := 0; sample < 64; sample++ {
			assignment := []bool{dataset.InputWords[0]>>uint(sample)&1 != 0, dataset.InputWords[1]>>uint(sample)&1 != 0}
			want, _ := graph.Evaluate(assignment)
			if dataset.Outputs[sample] != want[0] {
				t.Fatal(index, sample, dataset.Outputs[sample], want[0])
			}
		}
	}
	if slices.Equal(files.Files[1].Data, files.Files[2].Data) {
		t.Fatal("validation and test streams did not advance")
	}
	dataset, _ := ParseGiaSimulationPLA(files.Files[0].Data, GiaGenerationOptions{})
	dataset.Outputs[0] = !dataset.Outputs[0]
	comparison, err := graph.CompareSimulationDataset(dataset, GiaGenerationOptions{})
	if err != nil || comparison.Correct != 63 {
		t.Fatal(comparison, err)
	}
	encoded, err := comparison.MarshalStatsJSON(1024)
	if err != nil {
		t.Fatal(err)
	}
	var fields map[string]any
	if err = json.Unmarshal(encoded, &fields); err != nil {
		t.Fatal(err)
	}
	for _, key := range []string{"name", "input", "output", "and", "level", "total", "correct", "guess"} {
		if _, ok := fields[key]; !ok {
			t.Fatalf("missing JSON field %q in %s", key, encoded)
		}
	}
}

func TestGiaGenerationRejectsMalformedLimitsAndExtensions(t *testing.T) {
	graph, _, _, _ := giaGenerationGraph(t)
	for _, data := range [][]byte{
		nil,
		[]byte("00 1\n"),
		[]byte(strings.Repeat("00 1\n", 63)),
		[]byte(strings.Repeat("00 1\n", 63) + "0 1\n"),
	} {
		if _, err := ParseGiaSimulationPLA(data, GiaGenerationOptions{}); !errors.Is(err, ErrDimension) {
			t.Fatalf("malformed PLA accepted: %q: %v", data, err)
		}
	}
	valid := []byte(strings.Repeat("00 1\n", 64))
	for _, options := range []GiaGenerationOptions{{MaxBytes: len(valid) - 1}, {MaxSamples: 63}, {MaxWords: -1}, {Stop: func() bool { return true }}} {
		if _, err := ParseGiaSimulationPLA(valid, options); !(errors.Is(err, ErrLimit) || errors.Is(err, ErrDimension) || errors.Is(err, ErrCancelled)) {
			t.Fatal(options, err)
		}
	}
	for _, options := range []GiaGenerationOptions{{MaxWords: 63}, {MaxSamples: 63}, {MaxBytes: 11}, {Stop: func() bool { return true }}} {
		if _, err := GenerateGiaBinaryDatasets(graph, 1, 0, 0, "x", options); !(errors.Is(err, ErrLimit) || errors.Is(err, ErrCancelled)) {
			t.Fatal(options, err)
		}
	}
	if _, err := graph.SimulateInputWords([]uint64{0, 0}, GiaPackedSimulationOptions{Words: 2}); !errors.Is(err, ErrDimension) {
		t.Fatal(err)
	}
	buffered := NewGiaManager()
	input, _ := buffered.AppendCI()
	buffer, _ := buffered.AppendBuffer(input)
	buffered.AppendCO(buffer)
	if _, err := GenerateGiaPLADatasets(buffered, 1, 1, 0, "x", GiaGenerationOptions{}); !errors.Is(err, ErrUnsupported) {
		t.Fatal(err)
	}
	dataset := GiaSimulationDataset{Inputs: 2, Words: 1, InputWords: []uint64{0, 0}, Outputs: make([]bool, 63)}
	if _, err := graph.CompareSimulationDataset(dataset, GiaGenerationOptions{}); !errors.Is(err, ErrDimension) {
		t.Fatal(err)
	}
	if _, err := (GiaSimulationComparison{}).MarshalStatsJSON(1); !errors.Is(err, ErrLimit) {
		t.Fatal(err)
	}
}

func FuzzGiaInputWordSimulation(f *testing.F) {
	f.Add([]byte{2, 1, 7, 4, 9, 3}, uint64(0x0123456789abcdef), uint64(0xfedcba9876543210))
	f.Fuzz(func(t *testing.T, data []byte, word0, word1 uint64) {
		if len(data) == 0 {
			return
		}
		graph := NewGiaManager()
		width := 1 + int(data[0]%6)
		signals := make([]GiaLit, width)
		words := make([]uint64, width)
		for index := range signals {
			signals[index], _ = graph.AppendCI()
			words[index] = word0 ^ uint64(index)*word1
		}
		for index := 1; index < len(data) && len(signals) < 40; index++ {
			left := int(data[index]) % len(signals)
			right := (left + 1 + int(data[index]>>3)) % len(signals)
			if signals[left].ID() == signals[right].ID() {
				continue
			}
			gate, err := graph.AppendAnd(signals[left].NotIf(data[index]&1 != 0), signals[right].NotIf(data[index]&2 != 0))
			if err == nil {
				signals = append(signals, gate)
			}
		}
		root := signals[int(data[len(data)-1])%len(signals)].NotIf(data[len(data)-1]&0x80 != 0)
		graph.AppendCO(root)
		simulation, err := graph.SimulateInputWords(words, GiaPackedSimulationOptions{Words: 1})
		if err != nil {
			t.Fatal(err)
		}
		for sample := 0; sample < 64; sample++ {
			assignment := make([]bool, width)
			for input := range assignment {
				assignment[input] = words[input]>>uint(sample)&1 != 0
			}
			want, evalErr := graph.Evaluate(assignment)
			actual, bitErr := simulation.Bit(graph.cos[0], sample)
			if evalErr != nil || bitErr != nil || actual != want[0] {
				t.Fatal(fmt.Sprintf("sample %d: got %v want %v: %v %v", sample, actual, want, evalErr, bitErr))
			}
		}
	})
}

func FuzzGiaSimulationPLA(f *testing.F) {
	f.Add([]byte(strings.Repeat("01 1\n", 64)))
	f.Add([]byte(".i 2\n.o 1\n.p 64\n.type fr\n" + strings.Repeat("10 0\n", 64) + ".e\n"))
	f.Fuzz(func(t *testing.T, data []byte) {
		if len(data) > 4096 {
			return
		}
		dataset, err := ParseGiaSimulationPLA(data, GiaGenerationOptions{MaxBytes: 4096, MaxSamples: 512, MaxWords: 8, MaxEntries: 4096})
		if err != nil {
			return
		}
		var canonical strings.Builder
		fmt.Fprintf(&canonical, ".i %d\n.o 1\n.p %d\n.type fr\n", dataset.Inputs, len(dataset.Outputs))
		for sample, output := range dataset.Outputs {
			for input := 0; input < dataset.Inputs; input++ {
				bit := dataset.InputWords[input*dataset.Words+sample/64]>>uint(sample&63)&1 != 0
				canonical.WriteByte("01"[boolIndex(bit)])
			}
			canonical.WriteByte(' ')
			canonical.WriteByte("01"[boolIndex(output)])
			canonical.WriteByte('\n')
		}
		canonical.WriteString(".e\n")
		again, err := ParseGiaSimulationPLA([]byte(canonical.String()), GiaGenerationOptions{MaxBytes: canonical.Len(), MaxSamples: 512, MaxWords: 8, MaxEntries: 4096})
		if err != nil || again.Inputs != dataset.Inputs || again.Words != dataset.Words || !slices.Equal(again.InputWords, dataset.InputWords) || !slices.Equal(again.Outputs, dataset.Outputs) {
			t.Fatalf("canonical round trip failed: %v", err)
		}
	})
}
