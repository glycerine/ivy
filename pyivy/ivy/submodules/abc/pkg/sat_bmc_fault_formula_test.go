package abc

// Stand-alone formula semantics for src/sat/bmc/bmcFault.c.

import (
	"errors"
	"reflect"
	"testing"
)

func TestFaultFormulaParsingAndEvaluation(t *testing.T) {
	formula, err := ParseFaultFormula("((a&~b)^(p?q:~r))", FaultFormulaOptions{})
	if err != nil {
		t.Fatal(err)
	}
	if formula.String() != "((a&~b)^(p?q:~r))" || formula.ParameterCount() != 3 || formula.NodeCount() != 10 {
		t.Fatal(formula.String(), formula.ParameterCount(), formula.NodeCount())
	}
	for mask := 0; mask < 32; mask++ {
		variables := [2]bool{mask&1 != 0, mask&2 != 0}
		parameters := []bool{mask&4 != 0, mask&8 != 0, mask&16 != 0}
		got, err := formula.Evaluate(variables, parameters)
		want := variables[0] && !variables[1]
		branch := !parameters[2]
		if parameters[0] {
			branch = parameters[1]
		}
		want = want != branch
		if err != nil || got != want {
			t.Fatal(mask, got, want, err)
		}
	}
	for _, text := range []string{"", "a&b", "(a&b", "(a&b)&p", "(a&b)", "(a&p)", "(a&x)", "(a?b)", "(a?b:p:q)", "(a b)", "(a&b& p)"} {
		if _, err := ParseFaultFormula(text, FaultFormulaOptions{}); err == nil {
			t.Fatal("accepted invalid formula", text)
		}
	}
	if _, err = ParseFaultFormula("((a&b)|w)", FaultFormulaOptions{}); err != nil {
		t.Fatal("all eight parameters should be accepted", err)
	}
}

func TestFaultFormulaRealization(t *testing.T) {
	formula, err := ParseFaultFormula("((a^b)?(p|a):(q&~b))", FaultFormulaOptions{})
	if err != nil {
		t.Fatal(err)
	}
	graph := NewGiaManager()
	a, _ := graph.AppendCI()
	b, _ := graph.AppendCI()
	p, _ := graph.AppendCI()
	q, _ := graph.AppendCI()
	root, err := RealizeFaultFormula(graph, formula, [2]GiaLit{a, b}, []GiaLit{p, q})
	if err != nil {
		t.Fatal(err)
	}
	if _, err = graph.AppendCO(root); err != nil || graph.Check() != nil {
		t.Fatal(err)
	}
	for mask := 0; mask < 16; mask++ {
		inputs := []bool{mask&1 != 0, mask&2 != 0, mask&4 != 0, mask&8 != 0}
		outputs, err := graph.Evaluate(inputs)
		want, modelErr := formula.Evaluate([2]bool{inputs[0], inputs[1]}, inputs[2:])
		if err != nil || modelErr != nil || outputs[0] != want {
			t.Fatal(mask, outputs, want, err, modelErr)
		}
	}
}

func TestDeriveGiaFormulaSemantics(t *testing.T) {
	graph := NewGiaManager()
	a, _ := graph.AppendCI()
	b, _ := graph.AppendCI()
	p, _ := graph.AppendCI()
	x, _ := graph.AppendXorReal(b, p)
	y, _ := graph.AppendAnd(a.Not(), p)
	root, _ := graph.AppendMuxReal(a, x.Not(), y)
	_, _ = graph.AppendCO(root.Not())
	text, err := DeriveGiaFormula(graph, 0, []string{"a", "b", "p"}, FaultFormulaOptions{})
	if err != nil {
		t.Fatal(err)
	}
	formula, err := ParseFaultFormula(text, FaultFormulaOptions{})
	if err != nil {
		t.Fatal(text, err)
	}
	for mask := 0; mask < 8; mask++ {
		inputs := []bool{mask&1 != 0, mask&2 != 0, mask&4 != 0}
		outputs, _ := graph.Evaluate(inputs)
		got, err := formula.Evaluate([2]bool{inputs[0], inputs[1]}, inputs[2:])
		if err != nil || got != outputs[0] {
			t.Fatal(text, mask, got, outputs[0], err)
		}
	}
}

func TestUnfoldGiaFaultFormula(t *testing.T) {
	source := NewGiaManager()
	a, _ := source.AppendCI()
	b, _ := source.AppendCI()
	gate, _ := source.AppendAnd(a, b.Not())
	_, _ = source.AppendCO(gate)
	result, err := UnfoldGiaFaultFormula(source, "(p?a:b)", false, FaultFormulaOptions{})
	if err != nil {
		t.Fatal(err)
	}
	if result.Graph.InputCount() != 3 || !reflect.DeepEqual(result.SourceGates, []GiaID{gate.ID()}) || len(result.ParameterInputs) != 1 || len(result.ParameterInputs[0]) != 1 {
		t.Fatal(result.Graph.InputCount(), result.SourceGates, result.ParameterInputs)
	}
	object, _ := source.Object(gate.ID())
	first, _ := object.Fanin0(gate.ID())
	second, _ := object.Fanin1(gate.ID())
	for mask := 0; mask < 8; mask++ {
		inputs := []bool{mask&1 != 0, mask&2 != 0, mask&4 != 0}
		outputs, err := result.Graph.Evaluate(inputs)
		read := func(l GiaLit) bool {
			value := inputs[int(source.objects[l.ID()].Diff1())]
			if l.Complemented() {
				value = !value
			}
			return value
		}
		want := read(second)
		if inputs[2] {
			want = read(first)
		}
		if err != nil || outputs[0] != want {
			t.Fatal(mask, outputs, want, err)
		}
	}
}

func TestUnfoldGiaFaultFormulaRegisterDrivers(t *testing.T) {
	source := NewGiaManager()
	a, _ := source.AppendCI()
	r, _ := source.AppendCI()
	first, _ := source.AppendAnd(a, r)
	next, _ := source.AppendAnd(first, a.Not())
	_, _ = source.AppendCO(first)
	_, _ = source.AppendCO(next)
	_ = source.SetRegisterCount(1)
	result, err := UnfoldGiaFaultFormula(source, "((a&b)|(p&0))", true, FaultFormulaOptions{})
	if err != nil {
		t.Fatal(err)
	}
	if result.Graph.RegisterCount() != 1 || result.Graph.PrimaryInputCount() != 2 || !reflect.DeepEqual(result.SourceGates, []GiaID{next.ID()}) {
		t.Fatal(result.Graph.RegisterCount(), result.Graph.PrimaryInputCount(), result.SourceGates)
	}
	for mask := 0; mask < 8; mask++ {
		primary := []bool{mask&1 != 0, mask&4 != 0}
		state := []bool{mask&2 != 0}
		outputs, nextState, err := result.Graph.Step(primary, state)
		wantOutputs, wantNext, sourceErr := source.Step(primary[:1], state)
		if err != nil || sourceErr != nil || !reflect.DeepEqual(outputs, wantOutputs) || !reflect.DeepEqual(nextState, wantNext) {
			t.Fatal(mask, outputs, nextState, wantOutputs, wantNext, err, sourceErr)
		}
	}
}

func TestFaultFormulaFailuresAtomicAndLimits(t *testing.T) {
	formula, _ := ParseFaultFormula("((a&b)|p)", FaultFormulaOptions{})
	graph := NewGiaManager()
	a, _ := graph.AppendCI()
	b, _ := graph.AppendCI()
	before := graph.Clone()
	if _, err := RealizeFaultFormula(graph, formula, [2]GiaLit{a, b}, nil); !errors.Is(err, ErrDimension) || !reflect.DeepEqual(before, graph) {
		t.Fatal(err)
	}
	for _, options := range []FaultFormulaOptions{
		{MaxBytes: 3}, {MaxNodes: 2}, {Stop: func() bool { return true }},
	} {
		if _, err := ParseFaultFormula("((a&b)|p)", options); !(errors.Is(err, ErrLimit) || errors.Is(err, ErrCancelled)) {
			t.Fatal(options, err)
		}
	}
	source := NewGiaManager()
	x, _ := source.AppendCI()
	y, _ := source.AppendCI()
	z, _ := source.AppendAnd(x, y)
	_, _ = source.AppendCO(z)
	if result, err := UnfoldGiaFaultFormula(source, "((a&b)|p)", false, FaultFormulaOptions{MaxObjects: 4}); result != nil || !errors.Is(err, ErrLimit) {
		t.Fatal(result, err)
	}
	calls := 0
	if result, err := UnfoldGiaFaultFormula(source, "((a&b)|p)", false, FaultFormulaOptions{Stop: func() bool { calls++; return calls == 6 }}); result != nil || !errors.Is(err, ErrCancelled) {
		t.Fatal(result, calls, err)
	}
}

func FuzzFaultFormulaSemantics(f *testing.F) {
	for _, seed := range []byte{0, 1, 7, 31, 127, 255} {
		f.Add(seed, seed^0x5a)
	}
	formulas := []string{"((a&b)|p)", "((a^b)?p:q)", "(p?~a:b)", "((a|p)&(b|q))"}
	f.Fuzz(func(t *testing.T, selector, mask byte) {
		formula, err := ParseFaultFormula(formulas[int(selector)%len(formulas)], FaultFormulaOptions{})
		if err != nil {
			t.Fatal(err)
		}
		variables := [2]bool{mask&1 != 0, mask&2 != 0}
		parameters := make([]bool, formula.ParameterCount())
		for i := range parameters {
			parameters[i] = mask&(4<<i) != 0
		}
		got, err := formula.Evaluate(variables, parameters)
		if err != nil {
			t.Fatal(err)
		}
		var want bool
		switch int(selector) % len(formulas) {
		case 0:
			want = variables[0] && variables[1] || parameters[0]
		case 1:
			want = parameters[1]
			if variables[0] != variables[1] {
				want = parameters[0]
			}
		case 2:
			want = variables[1]
			if parameters[0] {
				want = !variables[0]
			}
		case 3:
			want = (variables[0] || parameters[0]) && (variables[1] || parameters[1])
		}
		if got != want {
			t.Fatal(formula.String(), variables, parameters, got, want)
		}
	})
}
