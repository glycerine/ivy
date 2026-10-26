package abc

// Go-only semantic and atomicity tests for dauNpn2 formula GIA lowering.

import (
	"errors"
	"testing"
)

func TestDauLogicFormulaGiaSemantics(t *testing.T) {
	formula, err := ParseDauLogicFormula(dauNPN2Example, DauLogicFormulaOptions{})
	if err != nil {
		t.Fatal(err)
	}
	for _, native := range []bool{false, true} {
		graph, inputs := dauGiaInputs(t, 5, native)
		before := graph.ObjectCount()
		result, err := graph.WithDauLogicFormula(dauNPN2Example, inputs, DauLogicGiaOptions{NativeGates: native})
		if err != nil {
			t.Fatal(native, err)
		}
		if result.Graph == graph || graph.ObjectCount() != before || result.Graph.Hashing() != native {
			t.Fatal("ownership/hash state", native)
		}
		for assignment := 0; assignment < 32; assignment++ {
			values := make([]bool, 5)
			for variable := range values {
				values[variable] = assignment&(1<<uint(variable)) != 0
			}
			all, err := result.Graph.evaluateValues(values)
			if err != nil {
				t.Fatal(err)
			}
			got := all[result.Root.ID()] != result.Root.Complemented()
			want, _ := formula.Evaluate(values)
			if got != want {
				t.Fatal(native, assignment, got, want)
			}
		}
	}
}

func TestDauLogicFormulaGiaAtomicErrors(t *testing.T) {
	graph, inputs := dauGiaInputs(t, 4, false)
	before := graph.ObjectCount()
	for _, test := range []struct {
		text    string
		inputs  []GiaLit
		options DauLogicGiaOptions
		want    error
	}{
		{"(a&", inputs, DauLogicGiaOptions{}, ErrFormat},
		{"(a&d)", inputs[:3], DauLogicGiaOptions{}, ErrDimension},
		{"(a&b)", []GiaLit{InvalidGiaLit, inputs[1]}, DauLogicGiaOptions{}, ErrInvalidLiteral},
		{"(a&b)", inputs, DauLogicGiaOptions{MaxObjects: before}, ErrLimit},
		{"(a&b)", inputs, DauLogicGiaOptions{Stop: func() bool { return true }}, ErrCancelled},
	} {
		result, err := graph.WithDauLogicFormula(test.text, test.inputs, test.options)
		if result.Graph != nil || !errors.Is(err, test.want) || graph.ObjectCount() != before {
			t.Fatal(test.text, result, err, graph.ObjectCount())
		}
	}
}

func TestBuildDauLogicLibrary(t *testing.T) {
	data := []byte("aaaaaaaa = a\n66666666 = (a^b)\n037d037d = " + dauNPN2Example + "\n")
	graph, err := BuildDauLogicLibrary(data, 5, DauLogicGiaOptions{NativeGates: true})
	if err != nil || graph.InputCount() != 5 || graph.OutputCount() != 3 || graph.Hashing() {
		t.Fatal(graph, err)
	}
	for assignment := 0; assignment < 32; assignment++ {
		inputs := make([]bool, 5)
		for variable := range inputs {
			inputs[variable] = assignment&(1<<uint(variable)) != 0
		}
		outputs, err := graph.Evaluate(inputs)
		if err != nil {
			t.Fatal(err)
		}
		formula, _ := ParseDauLogicFormula(dauNPN2Example, DauLogicFormulaOptions{})
		third, _ := formula.Evaluate(inputs)
		want := []bool{inputs[0], inputs[0] != inputs[1], third}
		for index := range want {
			if outputs[index] != want[index] {
				t.Fatal(assignment, outputs, want)
			}
		}
	}
	if graph, err := BuildDauLogicLibrary(data, 5, DauLogicGiaOptions{MaxObjects: 5}); graph != nil || !errors.Is(err, ErrLimit) {
		t.Fatal(graph, err)
	}
}

func FuzzDauLogicFormulaGia(f *testing.F) {
	f.Add(dauNPN2Example, false)
	f.Add("~(a^(b&~c))", true)
	f.Fuzz(func(t *testing.T, text string, native bool) {
		parse := DauLogicFormulaOptions{MaxBytes: 128, MaxNodes: 64, MaxDepth: 16}
		formula, err := ParseDauLogicFormula(text, parse)
		if err != nil {
			return
		}
		graph, inputs := dauGiaInputs(t, 6, false)
		result, err := graph.WithDauLogicFormula(text, inputs, DauLogicGiaOptions{NativeGates: native, MaxObjects: 1000, Parse: parse})
		if err != nil {
			t.Fatal(text, err)
		}
		for assignment := 0; assignment < 64; assignment++ {
			values := make([]bool, 6)
			for variable := range values {
				values[variable] = assignment&(1<<uint(variable)) != 0
			}
			all, err := result.Graph.evaluateValues(values)
			if err != nil {
				t.Fatal(err)
			}
			got := all[result.Root.ID()] != result.Root.Complemented()
			want, err := formula.Evaluate(values)
			if err != nil || got != want {
				t.Fatal(text, native, assignment, got, want, err)
			}
		}
	})
}
