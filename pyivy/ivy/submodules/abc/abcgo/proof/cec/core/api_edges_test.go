package core

import (
	"context"
	"testing"
)

type nilMapNetwork map[int]gate

func (nilMapNetwork) ObjectCount() int                      { panic("ObjectCount called on nil map network") }
func (nilMapNetwork) InputLiterals() []int32                { return nil }
func (nilMapNetwork) OutputLiterals() []int32               { return nil }
func (nilMapNetwork) Gate(int32) (int32, int32, int32, int) { panic("Gate called on nil map network") }

type nilSliceNetwork []gate

func (nilSliceNetwork) ObjectCount() int        { panic("ObjectCount called on nil slice network") }
func (nilSliceNetwork) InputLiterals() []int32  { return nil }
func (nilSliceNetwork) OutputLiterals() []int32 { return nil }
func (nilSliceNetwork) Gate(int32) (int32, int32, int32, int) {
	panic("Gate called on nil slice network")
}

func TestPublicOptionsAndInterfaceRejections(t *testing.T) {
	for _, network := range []Network{nilMapNetwork(nil), nilSliceNetwork(nil)} {
		if _, err := NewModel(network); err == nil || err.Error() != "cec: nil network" {
			t.Fatalf("typed nil %T: expected nil network error, got %v", network, err)
		}
	}
	a := newCircuit(1)
	a.outputs = []int32{a.inputs[0]}
	b := newCircuit(2)
	b.outputs = []int32{b.inputs[0]}
	for _, options := range [][]Options{{{ConflictLimit: -1}}, {{MaxSATVariables: -1}}, {{RecycleQueries: -1}}, {{}, {}}} {
		if _, err := NewManager(a, options...); err == nil {
			t.Errorf("NewManager accepted invalid options %+v", options)
		}
		if _, err := VerifyMiter(a, options...); err == nil {
			t.Errorf("VerifyMiter accepted invalid options %+v", options)
		}
	}
	if _, err := VerifyTwo(a, b); err == nil {
		t.Error("accepted mismatched input counts")
	}
	b = newCircuit(1)
	b.outputs = []int32{b.inputs[0], b.inputs[0]}
	if _, err := VerifyTwo(a, b); err == nil {
		t.Error("accepted mismatched output counts")
	}
	if _, err := VerifyPairs(a); err == nil {
		t.Error("accepted odd dual-output miter")
	}
	b = newCircuit(1)
	b.outputs = []int32{b.inputs[0]}
	b.registers = 1
	if _, err := VerifyTwo(a, b, Options{TreatRegistersAsInputs: true}); err == nil {
		t.Error("accepted mismatched register counts")
	}
	m, err := NewManager(a)
	if err != nil {
		t.Fatal(err)
	}
	for _, bad := range []int32{-1, 4, 100} {
		if _, err := m.CheckPair(a.inputs[0], bad); err == nil {
			t.Errorf("accepted invalid pair literal %d", bad)
		}
		if _, err := m.CheckConjunction(bad, a.inputs[0]); err == nil {
			t.Errorf("accepted invalid conjunction literal %d", bad)
		}
		if _, err := m.Model().Value([]bool{false}, bad); err == nil {
			t.Errorf("accepted invalid model literal %d", bad)
		}
	}
}

func TestModelInterfaceCopiesAndCounts(t *testing.T) {
	c := newCircuit(2)
	c.outputs = []int32{c.inputs[1] ^ 1, c.inputs[0]}
	c.registers = 1
	model, err := NewModel(c)
	if err != nil {
		t.Fatal(err)
	}
	if model.InputCount() != 2 || model.OutputCount() != 2 || model.RegisterCount() != 1 {
		t.Fatalf("incorrect interface sizes: %d %d %d", model.InputCount(), model.OutputCount(), model.RegisterCount())
	}
	in, out := model.Inputs(), model.Outputs()
	in[0] = 100
	out[0] = 100
	if model.Inputs()[0] != c.inputs[0] || model.Outputs()[0] != c.outputs[0] {
		t.Fatal("interface slices alias model storage")
	}
}

func TestCanceledVerificationVariants(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	a := newCircuit(1)
	a.outputs = []int32{a.inputs[0], a.inputs[0] ^ 1}
	b := newCircuit(1)
	b.outputs = append([]int32(nil), a.outputs...)
	for name, call := range map[string]func() (Result, error){
		"two":   func() (Result, error) { return VerifyTwo(a, b, Options{Context: ctx}) },
		"pairs": func() (Result, error) { return VerifyPairs(a, Options{Context: ctx}) },
		"miter": func() (Result, error) { return VerifyMiter(a, Options{Context: ctx}) },
	} {
		t.Run(name, func(t *testing.T) {
			r, err := call()
			if err != nil || r.Status != Undecided || r.Counterexample != nil {
				t.Fatalf("%+v %v", r, err)
			}
		})
	}
}
