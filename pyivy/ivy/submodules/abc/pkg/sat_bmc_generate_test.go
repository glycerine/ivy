package abc

// Go-only tests for packed BMC simulation and SAT diagnostics from bmcGen.c.

import (
	"errors"
	"reflect"
	"slices"
	"testing"
)

func bmcGenerationGraph(t testing.TB) (*GiaManager, GiaLit, GiaLit, GiaLit) {
	t.Helper()
	graph := NewGiaManager()
	a, _ := graph.AppendCI()
	b, _ := graph.AppendCI()
	root, err := graph.AppendAnd(a.Not(), b)
	if err != nil {
		t.Fatal(err)
	}
	graph.AppendCO(root.Not())
	return graph, a, b, root
}

func bmcGenerationRandomWords(count int) []uint64 {
	z, w := uint32(3716960521), uint32(2174103536)
	random := func() uint32 {
		z = 36969*(z&65535) + (z >> 16)
		w = 18000*(w&65535) + (w >> 16)
		return (z << 16) + w
	}
	random()
	result := make([]uint64, count)
	for index := range result {
		result[index] = uint64(random())<<32 | uint64(random())
	}
	return result
}

func TestGiaPackedSimulationExactStreamAndSemantics(t *testing.T) {
	graph, a, b, root := bmcGenerationGraph(t)
	before := graph.Clone()
	simulation, err := graph.SimulateRandomWords(GiaPackedSimulationOptions{Words: 2})
	if err != nil {
		t.Fatal(err)
	}
	if simulation.Words() != 2 || simulation.Objects() != graph.ObjectCount() || len(simulation.Values()) != 2*graph.ObjectCount() {
		t.Fatal(simulation.Words(), simulation.Objects(), len(simulation.Values()))
	}
	random := bmcGenerationRandomWords(4)
	wantA, wantB := random[:2], random[2:]
	wantRoot := []uint64{^wantA[0] & wantB[0], ^wantA[1] & wantB[1]}
	wantOutput := []uint64{^wantRoot[0], ^wantRoot[1]}
	for _, test := range []struct {
		id   GiaID
		want []uint64
	}{{a.ID(), wantA}, {b.ID(), wantB}, {root.ID(), wantRoot}, {graph.cos[0], wantOutput}} {
		got, wordErr := simulation.ObjectWords(test.id)
		if wordErr != nil || !slices.Equal(got, test.want) {
			t.Fatal(test.id, got, test.want, wordErr)
		}
		got[0] = 0
		again, _ := simulation.ObjectWords(test.id)
		if again[0] != test.want[0] {
			t.Fatal("object words alias")
		}
	}
	for sample := 0; sample < 128; sample++ {
		inputs := make([]bool, graph.InputCount())
		for index, id := range graph.cis {
			inputs[index], err = simulation.Bit(id, sample)
			if err != nil {
				t.Fatal(err)
			}
		}
		outputs, evalErr := graph.Evaluate(inputs)
		bit, bitErr := simulation.Bit(graph.cos[0], sample)
		if evalErr != nil || bitErr != nil || bit != outputs[0] {
			t.Fatal(sample, bit, outputs, evalErr, bitErr)
		}
	}
	again, err := graph.SimulateRandomWords(GiaPackedSimulationOptions{Words: 2})
	if err != nil || !slices.Equal(simulation.Values(), again.Values()) || !reflect.DeepEqual(graph, before) {
		t.Fatal("stream reset, ownership or repeatability", err)
	}
	if words, wordErr := simulation.ObjectWords(GiaID(graph.ObjectCount())); words != nil || !errors.Is(wordErr, ErrInvalidLiteral) {
		t.Fatal(words, wordErr)
	}
	for _, sample := range []int{-1, 128} {
		if _, bitErr := simulation.Bit(0, sample); !errors.Is(bitErr, ErrDimension) {
			t.Fatal(sample, bitErr)
		}
	}
}

func TestGiaSATEnumerationDiagnosticBehavior(t *testing.T) {
	graph, _, _, _ := bmcGenerationGraph(t)
	result, err := graph.EnumerateSATDiagnostic(0, GiaSATEnumerationOptions{Words: 2})
	if err != nil || len(result.OutputBits) != 128 || len(result.Statuses) != 128 || result.SATCount != 128 {
		t.Fatal(result.SATCount, len(result.OutputBits), len(result.Statuses), err)
	}
	for _, status := range result.Statuses {
		if status != SatSatisfiable {
			t.Fatal(status)
		}
	}
	simulation, _ := graph.SimulateRandomWords(GiaPackedSimulationOptions{Words: 2})
	for sample, bit := range result.OutputBits {
		want, _ := simulation.Bit(graph.cos[0], sample)
		if bit != want {
			t.Fatal(sample, bit, want)
		}
	}

	for _, test := range []struct {
		literal GiaLit
		status  SatStatus
		count   int
	}{{GiaConst0, SatUnsatisfiable, 0}, {GiaConst1, SatSatisfiable, 64}} {
		constant := NewGiaManager()
		constant.AppendCO(test.literal)
		answer, solveErr := constant.EnumerateSATDiagnostic(0, GiaSATEnumerationOptions{})
		if solveErr != nil || answer.SATCount != test.count {
			t.Fatal(test, answer, solveErr)
		}
		for _, status := range answer.Statuses {
			if status != test.status {
				t.Fatal(test, status)
			}
		}
	}

	unknown, err := graph.EnumerateSATDiagnostic(0, GiaSATEnumerationOptions{SAT: SatLimits{Stop: func() bool { return true }}})
	if err != nil || unknown.SATCount != 0 || len(unknown.Statuses) != 64 {
		t.Fatal(unknown, err)
	}
	for _, status := range unknown.Statuses {
		if status != SatUnknown {
			t.Fatal(status)
		}
	}
}

func TestGiaPackedSimulationLimitsCancellationAndDepth(t *testing.T) {
	graph, _, _, _ := bmcGenerationGraph(t)
	for _, options := range []GiaPackedSimulationOptions{
		{}, {Words: -1}, {Words: 1, MaxWords: -1}, {Words: 1, MaxEntries: -1},
		{Words: 2, MaxWords: 1}, {Words: 2, MaxEntries: 1},
		{Words: 1, Stop: func() bool { return true }},
	} {
		if result, err := graph.SimulateRandomWords(options); result.values != nil || !(errors.Is(err, ErrDimension) || errors.Is(err, ErrLimit) || errors.Is(err, ErrCancelled)) {
			t.Fatal(options, result, err)
		}
	}
	if result, err := (*GiaManager)(nil).SimulateRandomWords(GiaPackedSimulationOptions{Words: 1}); result.values != nil || !errors.Is(err, ErrDimension) {
		t.Fatal(result, err)
	}
	for _, options := range []GiaSATEnumerationOptions{
		{Words: -1}, {MaxSamples: -1}, {MaxVariables: -1}, {MaxClauses: -1}, {MaxLiterals: -1},
		{Words: 1, MaxSamples: 63}, {Stop: func() bool { return true }},
	} {
		if result, err := graph.EnumerateSATDiagnostic(0, options); result.Statuses != nil || !(errors.Is(err, ErrDimension) || errors.Is(err, ErrLimit) || errors.Is(err, ErrCancelled)) {
			t.Fatal(options, result, err)
		}
	}
	for _, output := range []int{-1, 1} {
		if _, err := graph.EnumerateSATDiagnostic(output, GiaSATEnumerationOptions{}); !errors.Is(err, ErrDimension) {
			t.Fatal(output, err)
		}
	}

	buffered := NewGiaManager()
	input, _ := buffered.AppendCI()
	buffer, _ := buffered.AppendBuffer(input)
	buffered.AppendCO(buffer)
	if result, err := buffered.SimulateRandomWords(GiaPackedSimulationOptions{Words: 1}); result.values != nil || !errors.Is(err, ErrUnsupported) {
		t.Fatal(result, err)
	}

	deep := NewGiaManager()
	root, _ := deep.AppendCI()
	for index := 0; index < 10000; index++ {
		input, _ := deep.AppendCI()
		root, _ = deep.AppendAnd(root, input.NotIf(index&1 != 0))
	}
	deep.AppendCO(root)
	result, err := deep.SimulateRandomWords(GiaPackedSimulationOptions{Words: 1, MaxEntries: 30000})
	if err != nil || len(result.values) != deep.ObjectCount() {
		t.Fatal(len(result.values), deep.ObjectCount(), err)
	}
}

func FuzzGiaPackedSimulation(f *testing.F) {
	f.Add([]byte{1, 3, 5, 7, 9, 11, 13})
	f.Fuzz(func(t *testing.T, data []byte) {
		if len(data) == 0 {
			return
		}
		graph := NewGiaManager()
		width := 1 + int(data[0]%6)
		signals := make([]GiaLit, width)
		for index := range signals {
			signals[index], _ = graph.AppendCI()
		}
		for index := 1; index < len(data) && len(signals) < 48; index++ {
			leftIndex := int(data[index]) % len(signals)
			rightIndex := (leftIndex + 1 + int(data[index]>>4)) % len(signals)
			if signals[leftIndex].ID() == signals[rightIndex].ID() {
				continue
			}
			left := signals[leftIndex].NotIf(data[index]&1 != 0)
			right := signals[rightIndex].NotIf(data[index]&2 != 0)
			if literal, err := graph.AppendAnd(left, right); err == nil {
				signals = append(signals, literal)
			}
		}
		root := signals[int(data[len(data)-1])%len(signals)].NotIf(data[len(data)-1]&0x80 != 0)
		graph.AppendCO(root)
		simulation, err := graph.SimulateRandomWords(GiaPackedSimulationOptions{Words: 1})
		if err != nil {
			t.Fatal(err)
		}
		for sample := 0; sample < 64; sample++ {
			inputs := make([]bool, graph.InputCount())
			for index, id := range graph.cis {
				inputs[index], err = simulation.Bit(id, sample)
				if err != nil {
					t.Fatal(err)
				}
			}
			outputs, evalErr := graph.Evaluate(inputs)
			bit, bitErr := simulation.Bit(graph.cos[0], sample)
			if evalErr != nil || bitErr != nil || bit != outputs[0] {
				t.Fatal(sample, bit, outputs, evalErr, bitErr)
			}
		}
	})
}
