package abc

// Tests for src/opt/fsim/fsimTsim.c.

import (
	"encoding/json"
	"errors"
	"os"
	"slices"
	"testing"
)

type fsimTernaryOracle struct {
	Cases []struct {
		States []string `json:"states"`
		Hashes []int    `json:"hashes"`
	} `json:"cases"`
}

func fsimTernaryGraphs(t testing.TB) []*AigManager {
	t.Helper()
	finish := func(graph *AigManager, registers int) *AigManager {
		if err := graph.SetRegisterCount(registers); err != nil {
			t.Fatal(err)
		}
		return graph
	}
	first := NewAigManager()
	r0, _ := first.CreateCI()
	first.CreateCO(r0)

	second := NewAigManager()
	pi, _ := second.CreateCI()
	r1, _ := second.CreateCI()
	second.CreateCO(pi)
	second.CreateCO(r1)
	second.CreateCO(pi)

	third := NewAigManager()
	a, _ := third.CreateCI()
	b, _ := third.CreateCI()
	third.CreateCO(a)
	third.CreateCO(b)
	third.CreateCO(b ^ 1)
	third.CreateCO(a)

	fourth := NewAigManager()
	registers := make([]AigLit, 17)
	for index := range registers {
		registers[index], _ = fourth.CreateCI()
	}
	for _, register := range registers {
		fourth.CreateCO(register ^ 1)
	}
	return []*AigManager{finish(first, 1), finish(second, 1), finish(third, 2), finish(fourth, 17)}
}

func fsimTernaryText(state []AigTernary) string {
	text := make([]byte, len(state))
	for index, value := range state {
		text[index] = map[AigTernary]byte{AigTernaryZero: '0', AigTernaryOne: '1', AigTernaryUnknown: 'x'}[value]
	}
	return string(text)
}

func TestFsimTernaryNative(t *testing.T) {
	data, err := os.ReadFile("test_vectors/oracle/fsim_ternary.json")
	if err != nil {
		t.Fatal(err)
	}
	var oracle fsimTernaryOracle
	if err = json.Unmarshal(data, &oracle); err != nil || len(oracle.Cases) != 4 {
		t.Fatal(len(oracle.Cases), err)
	}
	for index, graph := range fsimTernaryGraphs(t) {
		states, err := graph.FsimTernaryStates(FsimTernaryOptions{})
		if err != nil || len(states) != len(oracle.Cases[index].States) {
			t.Fatal(index, len(states), err)
		}
		for stateIndex, state := range states {
			if text := fsimTernaryText(state); text != oracle.Cases[index].States[stateIndex] {
				t.Fatal(index, stateIndex, text, oracle.Cases[index].States[stateIndex])
			}
			hash, err := FsimTernaryStateHash(packFsimTernaryState(state), 503)
			if err != nil || hash != oracle.Cases[index].Hashes[stateIndex] {
				t.Fatal(index, stateIndex, hash, err)
			}
		}
		if len(states[0]) > 0 {
			states[0][0] = AigTernaryUnknown
			again, err := graph.FsimTernaryStates(FsimTernaryOptions{})
			if err != nil || fsimTernaryText(again[0]) != oracle.Cases[index].States[0] {
				t.Fatal("ownership", index, err)
			}
		}
	}
	formatted, err := FormatFsimTernaryState([]AigTernary{AigTernaryZero, AigTernaryOne, AigTernaryUnknown})
	if err != nil || formatted != "01x (0=    1, 1=    1, x=    1)\n" {
		t.Fatal(formatted, err)
	}
}

func TestFsimTernaryErrorsAndLimits(t *testing.T) {
	graph := fsimTernaryGraphs(t)[2]
	for _, options := range []FsimTernaryOptions{{MaxStates: -1}, {MaxStateWords: -1}} {
		if states, err := graph.FsimTernaryStates(options); states != nil || !errors.Is(err, ErrDimension) {
			t.Fatal(states, err)
		}
	}
	if states, err := (*AigManager)(nil).FsimTernaryStates(FsimTernaryOptions{}); states != nil || !errors.Is(err, ErrDimension) {
		t.Fatal(states, err)
	}
	if states, err := NewAigManager().FsimTernaryStates(FsimTernaryOptions{}); states != nil || !errors.Is(err, ErrDimension) {
		t.Fatal(states, err)
	}
	if states, err := graph.FsimTernaryStates(FsimTernaryOptions{MaxStates: 1}); states != nil || !errors.Is(err, ErrLimit) {
		t.Fatal(states, err)
	}
	if states, err := fsimTernaryGraphs(t)[3].FsimTernaryStates(FsimTernaryOptions{MaxStateWords: 1}); states != nil || !errors.Is(err, ErrLimit) {
		t.Fatal(states, err)
	}
	if states, err := graph.FsimTernaryStates(FsimTernaryOptions{Stop: func() bool { return true }}); states != nil || !errors.Is(err, ErrCancelled) {
		t.Fatal(states, err)
	}
	buffered := NewAigManager()
	ci, _ := buffered.CreateCI()
	buffer, _ := buffered.CreateBuffer(ci)
	buffered.CreateCO(buffer)
	buffered.SetRegisterCount(1)
	if states, err := buffered.FsimTernaryStates(FsimTernaryOptions{}); states != nil || !errors.Is(err, ErrUnsupported) {
		t.Fatal(states, err)
	}
	if _, err := FsimTernaryStateHash(nil, 0); !errors.Is(err, ErrDimension) {
		t.Fatal(err)
	}
	if text, err := FormatFsimTernaryState([]AigTernary{0}); text != "" || !errors.Is(err, ErrInvariant) {
		t.Fatal(text, err)
	}
}

func FuzzFsimTernaryStates(f *testing.F) {
	f.Add([]byte{0, 1, 2, 3, 4}, uint8(2), uint8(1))
	f.Add([]byte{255, 0, 17}, uint8(6), uint8(0))
	f.Fuzz(func(t *testing.T, choices []byte, rawRegisters, rawInputs uint8) {
		registerCount := int(rawRegisters)%6 + 1
		inputCount := int(rawInputs) % 4
		graph := NewAigManager()
		signals := []AigLit{AigConst0, AigConst1}
		for index := 0; index < inputCount+registerCount; index++ {
			literal, err := graph.CreateCI()
			if err != nil {
				t.Fatal(err)
			}
			signals = append(signals, literal)
		}
		if len(choices) == 0 {
			choices = []byte{0}
		}
		for index := 0; index < registerCount; index++ {
			choice := choices[index%len(choices)]
			left := signals[int(choice)%len(signals)]
			if choice&1 != 0 {
				left ^= 1
			}
			driver := left
			if choice&2 != 0 {
				right := signals[(int(choice)+index+1)%len(signals)]
				if choice&4 != 0 {
					right ^= 1
				}
				var err error
				driver, err = graph.And(left, right)
				if err != nil {
					t.Fatal(err)
				}
				signals = append(signals, driver)
			}
			if _, err := graph.CreateCO(driver); err != nil {
				t.Fatal(err)
			}
		}
		if err := graph.SetRegisterCount(registerCount); err != nil {
			t.Fatal(err)
		}
		states, err := graph.FsimTernaryStates(FsimTernaryOptions{MaxStates: 1024, MaxStateWords: 4096})
		if err != nil {
			t.Fatal(err)
		}
		if len(states) < 2 || !slices.Equal(states[len(states)-1], states[len(states)-2]) && !func() bool {
			for _, state := range states[:len(states)-2] {
				if slices.Equal(state, states[len(states)-1]) {
					return true
				}
			}
			return false
		}() {
			t.Fatal("missing repeated terminal state", len(states))
		}
		for first := 0; first < len(states)-1; first++ {
			for second := first + 1; second < len(states)-1; second++ {
				if slices.Equal(states[first], states[second]) {
					t.Fatal("premature duplicate", first, second)
				}
			}
		}
	})
}
