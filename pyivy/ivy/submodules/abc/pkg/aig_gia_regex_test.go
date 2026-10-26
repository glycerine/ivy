package abc

// Stand-alone language, scalar, and bit-parallel models for
// src/aig/gia/giaRex.c.

import (
	"errors"
	"math/rand"
	"regexp"
	"slices"
	"strings"
	"testing"
)

func giaRegexWords(alphabet string, maximum int, visit func(string)) {
	var generate func(string, int)
	generate = func(prefix string, left int) {
		visit(prefix)
		if left == 0 {
			return
		}
		for i := 0; i < len(alphabet); i++ {
			generate(prefix+alphabet[i:i+1], left-1)
		}
	}
	generate("", maximum)
}

func checkGiaRegexLanguage(t testing.TB, expression string, maximum int) {
	t.Helper()
	want, err := regexp.Compile("^(?:" + expression + ")$")
	if err != nil {
		t.Fatal(expression, err)
	}
	for _, ordered := range []bool{false, true} {
		automaton, compileErr := CompileGiaRegex(expression, ordered, GiaRegexOptions{})
		if compileErr != nil {
			t.Fatal(expression, ordered, compileErr)
		}
		if automaton.Graph.Name != expression || automaton.Graph.PrimaryOutputCount() != 1 || automaton.Graph.RegisterCount() != len(automaton.InitialState) {
			t.Fatal(expression, automaton)
		}
		if ordered && !slices.IsSorted(automaton.Alphabet) {
			t.Fatal(expression, automaton.Alphabet)
		}
		alphabet := "ABCX"
		giaRegexWords(alphabet, maximum, func(word string) {
			accepted, simulationErr := automaton.Simulate(word)
			if simulationErr != nil {
				t.Fatal(expression, word, simulationErr)
			}
			for prefix := 1; prefix <= len(word); prefix++ {
				if accepted[prefix-1] != want.MatchString(word[:prefix]) {
					t.Fatal(expression, ordered, word[:prefix], accepted[prefix-1])
				}
			}
		})
	}
}

func TestGiaRegexLanguagesAlphabetAndMetadata(t *testing.T) {
	for _, expression := range []string{
		"A",
		"A*",
		"(A|B)",
		"((A|B)*C)",
		"((A*B|AC)D)",
		" ( ( A | B ) * C ) ",
	} {
		compact := strings.NewReplacer(" ", "", "\t", "", "\n", "", "\r", "").Replace(expression)
		if expression != compact {
			automaton, err := CompileGiaRegex(expression, true, GiaRegexOptions{})
			if err != nil {
				t.Fatal(err)
			}
			plain, err := CompileGiaRegex(compact, true, GiaRegexOptions{})
			if err != nil || !slices.Equal(automaton.Alphabet, plain.Alphabet) {
				t.Fatal(expression, err)
			}
			for _, word := range []string{"A", "B", "ABC", "AAC"} {
				first, _ := automaton.Simulate(word)
				second, _ := plain.Simulate(word)
				if !slices.Equal(first, second) {
					t.Fatal(expression, word, first, second)
				}
			}
			continue
		}
		checkGiaRegexLanguage(t, expression, 4)
	}
	first, err := CompileGiaRegex("C(A|B)", false, GiaRegexOptions{})
	if err != nil || !slices.Equal(first.Alphabet, []byte{'C', 'A', 'B'}) {
		t.Fatal(first.Alphabet, err)
	}
	ordered, err := CompileGiaRegex("C(A|B)", true, GiaRegexOptions{})
	if err != nil || !slices.Equal(ordered.Alphabet, []byte{'A', 'B', 'C'}) {
		t.Fatal(ordered.Alphabet, err)
	}
}

func TestGiaRegexExactlyOneAndTranspose(t *testing.T) {
	for variables := 1; variables <= 9; variables++ {
		graph := NewGiaManager()
		inputs := make([]GiaLit, variables)
		for i := range inputs {
			inputs[i], _ = graph.AppendCI()
		}
		graph.HashStart()
		one, err := giaRegexExactlyOne(graph, inputs, nil)
		graph.HashStop()
		if err != nil {
			t.Fatal(variables, err)
		}
		graph.AppendCO(one)
		for assignment := 0; assignment < 1<<variables; assignment++ {
			values := make([]bool, variables)
			count := 0
			for i := range values {
				values[i] = assignment&(1<<i) != 0
				if values[i] {
					count++
				}
			}
			got, _ := graph.Evaluate(values)
			if got[0] != (count == 1) {
				t.Fatal(variables, assignment, got)
			}
		}
	}
	random := rand.New(rand.NewSource(17))
	var matrix, want [64]uint64
	for row := range matrix {
		matrix[row] = random.Uint64()
		for column := 0; column < 64; column++ {
			if matrix[row]&(uint64(1)<<column) != 0 {
				want[63-column] |= uint64(1) << (63 - row)
			}
		}
	}
	original := matrix
	GiaTranspose64(&matrix)
	if matrix != want {
		t.Fatal("transpose")
	}
	GiaTranspose64(&matrix)
	if matrix != original {
		t.Fatal("transpose involution")
	}
}

func TestGiaRegexBitParallelStepAndWalk(t *testing.T) {
	automaton, err := CompileGiaRegex("((A|B)*C)", true, GiaRegexOptions{})
	if err != nil {
		t.Fatal(err)
	}
	graph := automaton.Graph
	for stateBits := 0; stateBits < 1<<graph.RegisterCount(); stateBits++ {
		state := make([]bool, graph.RegisterCount())
		packed := uint64(0)
		for i := range state {
			state[i] = stateBits&(1<<i) != 0
			if state[i] {
				packed |= uint64(1) << (63 - i)
			}
		}
		output, next, stepErr := graph.AutomatonStep64(packed)
		if stepErr != nil || len(next) != graph.PrimaryInputCount() {
			t.Fatal(stepErr, len(next))
		}
		for input := 0; input < graph.PrimaryInputCount(); input++ {
			oneHot := make([]bool, graph.PrimaryInputCount())
			oneHot[input] = true
			wantOutput, wantNext, scalarErr := graph.Step(oneHot, state)
			gotNext := uint64(0)
			for i, value := range wantNext {
				if value {
					gotNext |= uint64(1) << (63 - i)
				}
			}
			if scalarErr != nil || (output&(uint64(1)<<(63-input)) != 0) != wantOutput[0] || next[input] != gotNext {
				t.Fatal(stateBits, input, output, next[input], wantOutput, gotNext, scalarErr)
			}
		}
	}
	first, err := graph.WalkAutomaton(GiaAutomatonWalkOptions{Steps: 30, Walks: 4, Seed: 91})
	second, secondErr := graph.WalkAutomaton(GiaAutomatonWalkOptions{Steps: 30, Walks: 4, Seed: 91})
	if err != nil || secondErr != nil || !slices.Equal(first.States, second.States) || !slices.Equal(first.Counts, second.Counts) || len(first.Walks) != 4 {
		t.Fatal(first, second, err, secondErr)
	}
	for _, walk := range first.Walks {
		if len(walk.States) != len(walk.Inputs)+1 || len(walk.Accepted) != len(walk.Inputs) {
			t.Fatal(walk)
		}
		for step, input := range walk.Inputs {
			output, next, stepErr := graph.AutomatonStep64(walk.States[step])
			if stepErr != nil || next[input] != walk.States[step+1] || (output&(uint64(1)<<(63-input)) != 0) != walk.Accepted[step] {
				t.Fatal(step, input, walk, stepErr)
			}
		}
	}
}

func TestGiaRegexErrorsLimitsCancellationAndDeep(t *testing.T) {
	for _, expression := range []string{"", "()", "(A", "A)", "(A|B|C)"} {
		if result, err := CompileGiaRegex(expression, true, GiaRegexOptions{}); err == nil || result.Graph != nil {
			t.Fatal(expression, result, err)
		}
	}
	for _, options := range []GiaRegexOptions{{MaxExpression: -1}, {MaxObjects: -1}, {MaxEntries: -1}, {MaxExpression: 2}, {MaxObjects: 2}, {MaxEntries: 2}} {
		if result, err := CompileGiaRegex("(A|B)", true, options); err == nil || result.Graph != nil {
			t.Fatal(options, result, err)
		}
	}
	for threshold := 1; threshold <= 12; threshold++ {
		calls := 0
		result, err := CompileGiaRegex("((A|B)*C)", true, GiaRegexOptions{Stop: func() bool { calls++; return calls >= threshold }})
		if !errors.Is(err, ErrCancelled) || result.Graph != nil {
			t.Fatal(threshold, calls, result, err)
		}
	}
	automaton, err := CompileGiaRegex("A", true, GiaRegexOptions{})
	if err != nil {
		t.Fatal(err)
	}
	for _, options := range []GiaAutomatonWalkOptions{{Steps: -1}, {Walks: -1}, {MaxStates: -1}, {MaxVisits: -1}, {Steps: 10, Walks: 2, MaxVisits: 1}} {
		if result, walkErr := automaton.Graph.WalkAutomaton(options); walkErr == nil || result.States != nil {
			t.Fatal(options, result, walkErr)
		}
	}
	growing := NewGiaManager()
	input, _ := growing.AppendCI()
	growing.AppendCI()
	growing.AppendCO(GiaConst0)
	growing.AppendCO(input)
	growing.SetRegisterCount(1)
	if result, walkErr := growing.WalkAutomaton(GiaAutomatonWalkOptions{Steps: 2, Walks: 1, MaxStates: 1}); !errors.Is(walkErr, ErrLimit) || result.States != nil {
		t.Fatal(result, walkErr)
	}
	if result, err := automaton.Graph.WalkAutomaton(GiaAutomatonWalkOptions{Steps: 1, Walks: 1, Stop: func() bool { return true }}); !errors.Is(err, ErrCancelled) || result.States != nil {
		t.Fatal(result, err)
	}
	deep := strings.Repeat("A", 1000)
	compiled, err := CompileGiaRegex(deep, true, GiaRegexOptions{})
	if err != nil {
		t.Fatal(err)
	}
	accepted, err := compiled.Simulate(deep)
	if err != nil || len(accepted) != len(deep) || !accepted[len(accepted)-1] {
		t.Fatal(len(accepted), err)
	}
}

func giaRandomRegex(random *rand.Rand, depth int) string {
	if depth == 0 || random.Intn(4) == 0 {
		return string(byte('A' + random.Intn(3)))
	}
	switch random.Intn(3) {
	case 0:
		return "(" + giaRandomRegex(random, depth-1) + giaRandomRegex(random, depth-1) + ")"
	case 1:
		return "(" + giaRandomRegex(random, depth-1) + "|" + giaRandomRegex(random, depth-1) + ")"
	default:
		return "(" + giaRandomRegex(random, depth-1) + ")*"
	}
}

func FuzzGiaRegexLanguage(f *testing.F) {
	f.Add(uint64(1), uint8(7))
	f.Add(uint64(0x9e3779b97f4a7c15), uint8(19))
	f.Fuzz(func(t *testing.T, seed uint64, wordBits uint8) {
		random := rand.New(rand.NewSource(int64(seed)))
		expression := giaRandomRegex(random, 3)
		want, err := regexp.Compile("^(?:" + expression + ")$")
		if err != nil {
			t.Fatal(expression, err)
		}
		automaton, err := CompileGiaRegex(expression, seed&1 != 0, GiaRegexOptions{MaxObjects: 10000, MaxEntries: 10000})
		if err != nil {
			t.Fatal(expression, err)
		}
		length := int(wordBits%7) + 1
		word := make([]byte, length)
		for i := range word {
			word[i] = byte('A' + (wordBits>>uint(i%6))%4)
		}
		accepted, err := automaton.Simulate(string(word))
		if err != nil {
			t.Fatal(err)
		}
		for prefix := 1; prefix <= len(word); prefix++ {
			if accepted[prefix-1] != want.Match(word[:prefix]) {
				t.Fatal(expression, string(word[:prefix]), accepted[prefix-1])
			}
		}
	})
}

func FuzzGiaRegexPackedStep(f *testing.F) {
	f.Add(uint64(3), uint64(0))
	f.Add(uint64(0x243f6a8885a308d3), uint64(0x13198a2e03707344))
	f.Fuzz(func(t *testing.T, seed, stateBits uint64) {
		random := rand.New(rand.NewSource(int64(seed)))
		expression := giaRandomRegex(random, 3)
		automaton, err := CompileGiaRegex(expression, seed&1 != 0, GiaRegexOptions{MaxObjects: 10000, MaxEntries: 10000})
		if err != nil {
			t.Fatal(expression, err)
		}
		graph := automaton.Graph
		state := make([]bool, graph.RegisterCount())
		packedState := uint64(0)
		for i := range state {
			state[i] = stateBits&(uint64(1)<<uint(i)) != 0
			if state[i] {
				packedState |= uint64(1) << (63 - i)
			}
		}
		output, successors, err := graph.AutomatonStep64(packedState)
		if err != nil {
			t.Fatal(expression, err)
		}
		for input := 0; input < graph.PrimaryInputCount(); input++ {
			oneHot := make([]bool, graph.PrimaryInputCount())
			oneHot[input] = true
			wantOutput, wantNext, stepErr := graph.Step(oneHot, state)
			packedNext := uint64(0)
			for i, value := range wantNext {
				if value {
					packedNext |= uint64(1) << (63 - i)
				}
			}
			if stepErr != nil || successors[input] != packedNext || (output&(uint64(1)<<(63-input)) != 0) != wantOutput[0] {
				t.Fatal(expression, input, successors[input], packedNext, output, wantOutput, stepErr)
			}
		}
	})
}
