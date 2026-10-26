package abc

// Go-only tests for src/base/io/ioReadDsd.c.

import (
	"errors"
	"slices"
	"sync"
	"testing"
)

func abcDSDOutputTruth(t testing.TB, network *ABCNetwork) uint64 {
	t.Helper()
	converted, err := ABCNetworkToAIG(network, ABCFunctionalOptions{MaxOperations: 1 << 20, MaxTruthVariables: 20})
	if err != nil {
		t.Fatal(err)
	}
	inputs := converted.Graph.InputCount()
	if inputs > 6 || converted.Graph.OutputCount() != 1 {
		t.Fatalf("dimensions: inputs=%d outputs=%d", inputs, converted.Graph.OutputCount())
	}
	var truth uint64
	for assignment := 0; assignment < 1<<inputs; assignment++ {
		values := make([]bool, inputs)
		for index := range values {
			values[index] = assignment>>index&1 != 0
		}
		outputs, evaluateErr := converted.Graph.Evaluate(values)
		if evaluateErr != nil {
			t.Fatal(evaluateErr)
		}
		if outputs[0] {
			truth |= uint64(1) << assignment
		}
	}
	return truth
}

func abcDSDVariableTruth(variable, variables int) uint64 {
	var result uint64
	for assignment := 0; assignment < 1<<variables; assignment++ {
		if assignment>>variable&1 != 0 {
			result |= uint64(1) << assignment
		}
	}
	return result
}

func TestABCDSDFindSplitAndExactNetwork(t *testing.T) {
	end, err := ABCDSDFindEnd("x((a*b)+c)y", 1, ABCDSDReadOptions{})
	if err != nil || end != 9 {
		t.Fatalf("find end: %d, %v", end, err)
	}
	if _, err = ABCDSDFindEnd("(a", 0, ABCDSDReadOptions{}); !errors.Is(err, ErrFormat) {
		t.Fatalf("missing end: %v", err)
	}
	parts, xor, err := SplitABCDSDParts("!a+8((b),(c))+(d*e)", ABCDSDReadOptions{})
	if err != nil || !xor || !slices.Equal(parts, []string{"!a", "8((b),(c))", "(d*e)"}) {
		t.Fatalf("split: %q xor=%v err=%v", parts, xor, err)
	}
	for _, malformed := range []string{"", "a*", "a*b+c", "a/b", "!!a", "(a"} {
		if got, _, splitErr := SplitABCDSDParts(malformed, ABCDSDReadOptions{}); got != nil || splitErr == nil {
			t.Fatalf("split accepted %q: %q, %v", malformed, got, splitErr)
		}
	}

	network, err := ReadABCDSD("a*b", ABCDSDReadOptions{})
	if err != nil {
		t.Fatal(err)
	}
	if network.Name != "dsd" || network.ObjectCountByType(ABCObjectPI) != 2 || network.ObjectCountByType(ABCObjectNode) != 1 || network.ObjectCountByType(ABCObjectPO) != 1 {
		t.Fatalf("network counts: objects=%d PIs=%d nodes=%d POs=%d", network.ObjectCount(), network.ObjectCountByType(ABCObjectPI), network.ObjectCountByType(ABCObjectNode), network.ObjectCountByType(ABCObjectPO))
	}
	if network.objects[network.pis[0]].Name != "a" || network.objects[network.pis[1]].Name != "b" || network.objects[network.pos[0]].Name != "F" {
		t.Fatal("terminal names")
	}
	node := network.objects[3]
	if node.Function != "11 1\n" || len(node.Fanins) != 2 || node.Fanins[0].ID != network.pis[0] || node.Fanins[1].ID != network.pis[1] {
		t.Fatalf("AND node: %+v", node)
	}

	parenthesized, err := ReadABCDSD("(a*b)", ABCDSDReadOptions{})
	if err != nil {
		t.Fatal(err)
	}
	if parenthesized.ObjectCountByType(ABCObjectNode) != 3 {
		t.Fatalf("doubled parenthesis nodes: %d", parenthesized.ObjectCountByType(ABCObjectNode))
	}
	prime, err := ReadABCDSD("8(a,b)", ABCDSDReadOptions{})
	if err != nil {
		t.Fatal(err)
	}
	if prime.ObjectCountByType(ABCObjectNode) != 4 {
		t.Fatalf("prime nodes: %d", prime.ObjectCountByType(ABCObjectNode))
	}
	primeNode := prime.objects[4]
	if primeNode.Function != "11 1\n" || len(primeNode.Fanins) != 2 {
		t.Fatalf("prime function/order: %+v", primeNode)
	}
}

func TestABCDSDSemanticsHexOrderAndSparseNames(t *testing.T) {
	tests := []struct {
		formula   string
		variables int
		want      func(int) bool
	}{
		{"a*b", 2, func(a int) bool { return a&1 != 0 && a&2 != 0 }},
		{"a+b", 2, func(a int) bool { return (a&1 != 0) != (a&2 != 0) }},
		{"!a", 1, func(a int) bool { return a&1 == 0 }},
		{"8(a,b)", 2, func(a int) bool { return a == 3 }},
		{"E(a,b)", 2, func(a int) bool { return a != 0 }},
		{"96(a,b,c)", 3, func(a int) bool { return ((a&1)+((a>>1)&1)+((a>>2)&1))&1 != 0 }},
		{"0(a,b)", 2, func(int) bool { return false }},
		{"F(a,b)", 2, func(int) bool { return true }},
		{"(a*b)+c", 3, func(a int) bool { return (a&3 == 3) != (a&4 != 0) }},
	}
	for _, test := range tests {
		network, err := ReadABCDSD(test.formula, ABCDSDReadOptions{})
		if err != nil {
			t.Fatalf("%s: %v", test.formula, err)
		}
		got := abcDSDOutputTruth(t, network)
		var want uint64
		for assignment := 0; assignment < 1<<test.variables; assignment++ {
			if test.want(assignment) {
				want |= uint64(1) << assignment
			}
		}
		if got != want {
			t.Fatalf("%s: got %016x want %016x", test.formula, got, want)
		}
	}

	sparse, err := ReadABCDSD("z", ABCDSDReadOptions{})
	if err != nil {
		t.Fatal(err)
	}
	if len(sparse.pis) != 26 || sparse.objects[sparse.pis[0]].Name != "a0" || sparse.objects[sparse.pis[25]].Name != "z0" {
		t.Fatalf("sparse fake names: count=%d first=%q last=%q", len(sparse.pis), sparse.objects[sparse.pis[0]].Name, sparse.objects[sparse.pis[len(sparse.pis)-1]].Name)
	}

	hex, err := ReadABCDSD("96(a,b,c)", ABCDSDReadOptions{})
	if err != nil {
		t.Fatal(err)
	}
	var prime ABCNetworkObject
	for _, object := range hex.objects {
		if object.Alive && object.Type == ABCObjectNode && object.Function == "001 1\n111 1\n100 1\n010 1\n" {
			prime = object
			break
		}
	}
	if !prime.Alive {
		t.Fatal("native hex minterm order was not retained")
	}
}

func TestABCDSDMalformedLimitsCancellationAndOwnership(t *testing.T) {
	for _, formula := range []string{"", "a*", "a*b+c", "(a", "a)", "G(a,b)", "8(a)", "123(a,b,c)", "8(a,b)junk", "a b"} {
		if network, err := ReadABCDSD(formula, ABCDSDReadOptions{}); network != nil || err == nil {
			t.Fatalf("malformed %q: network=%v err=%v", formula, network, err)
		}
	}
	formula := "96(a,b,c)"
	for _, options := range []ABCDSDReadOptions{
		{MaxFormulaBytes: len(formula) - 1}, {MaxObjects: 4}, {MaxEdges: 1}, {MaxFanins: 1}, {MaxDepth: 1},
		{MaxCubes: 1}, {MaxWork: 1}, {MaxBytes: 1}, {Stop: func() bool { return true }},
	} {
		if network, err := ReadABCDSD(formula, options); network != nil || !(errors.Is(err, ErrLimit) || errors.Is(err, ErrCancelled)) {
			t.Fatalf("limit %+v: network=%v err=%v", options, network, err)
		}
	}
	if network, err := ReadABCDSD("a", ABCDSDReadOptions{MaxDepth: -1}); network != nil || !errors.Is(err, ErrDimension) {
		t.Fatalf("negative limit: network=%v err=%v", network, err)
	}
	network, err := ReadABCDSD("a*b", ABCDSDReadOptions{})
	if err != nil {
		t.Fatal(err)
	}
	clone := network.Clone()
	network.objects[3].Function = "changed"
	if clone.objects[3].Function != "11 1\n" {
		t.Fatal("network clone aliases function state")
	}
}

func TestABCDSDConcurrentIndependentReaders(t *testing.T) {
	const formula = "!(96(a,b,c)+8(a,b))"
	want, err := ReadABCDSD(formula, ABCDSDReadOptions{})
	if err != nil {
		t.Fatal(err)
	}
	wantTruth := abcDSDOutputTruth(t, want)
	var wait sync.WaitGroup
	failures := make(chan error, 16)
	for worker := 0; worker < 16; worker++ {
		wait.Add(1)
		go func() {
			defer wait.Done()
			network, readErr := ReadABCDSD(formula, ABCDSDReadOptions{})
			if readErr != nil {
				failures <- readErr
				return
			}
			converted, convertErr := ABCNetworkToAIG(network, ABCFunctionalOptions{})
			if convertErr != nil {
				failures <- convertErr
				return
			}
			var truth uint64
			for assignment := 0; assignment < 8; assignment++ {
				inputs := []bool{assignment&1 != 0, assignment&2 != 0, assignment&4 != 0}
				outputs, evaluateErr := converted.Graph.Evaluate(inputs)
				if evaluateErr != nil {
					failures <- evaluateErr
					return
				}
				if outputs[0] {
					truth |= uint64(1) << assignment
				}
			}
			if truth != wantTruth {
				failures <- ErrInvariant
			}
		}()
	}
	wait.Wait()
	close(failures)
	for failure := range failures {
		t.Error(failure)
	}
}

func fuzzABCDSDFormula(data []byte) (string, int, uint64) {
	variables := 1
	if len(data) != 0 {
		variables += int(data[0] % 4)
	}
	mask := uint64(1)<<(1<<variables) - 1
	formula := string([]byte{byte('a' + variables - 1)})
	truth := abcDSDVariableTruth(variables-1, variables)
	for index := 1; index < len(data) && index < 9; index++ {
		value := data[index]
		variable := int(value>>1) % variables
		leaf := string([]byte{byte('a' + variable)})
		leafTruth := abcDSDVariableTruth(variable, variables)
		if value&4 != 0 {
			leaf = "!" + leaf
			leafTruth = (^leafTruth) & mask
		}
		if value&32 != 0 && variables >= 2 {
			leaf = "8(a,b)"
			leafTruth = abcDSDVariableTruth(0, variables) & abcDSDVariableTruth(1, variables)
		}
		operator := byte('*')
		if value&1 != 0 {
			operator = '+'
		}
		formula = "(" + formula + string(operator) + leaf + ")"
		if operator == '*' {
			truth &= leafTruth
		} else {
			truth ^= leafTruth
		}
		if value&8 != 0 {
			formula = "!" + formula
			truth = (^truth) & mask
		}
	}
	return formula, variables, truth
}

func FuzzABCDSDReadSemantics(f *testing.F) {
	for _, seed := range [][]byte{{}, {1, 2, 3}, {3, 0xff, 8, 32}, {2, 5, 9, 17, 33}} {
		f.Add(seed)
	}
	f.Fuzz(func(t *testing.T, data []byte) {
		formula, variables, want := fuzzABCDSDFormula(data)
		if len(formula) > 4096 {
			t.Skip()
		}
		network, err := ReadABCDSD(formula, ABCDSDReadOptions{MaxFormulaBytes: 4096, MaxObjects: 4096, MaxEdges: 8192, MaxFanins: 32, MaxDepth: 128, MaxCubes: 256, MaxWork: 1 << 20, MaxBytes: 1 << 20})
		if err != nil {
			t.Fatalf("%q: %v", formula, err)
		}
		converted, err := ABCNetworkToAIG(network, ABCFunctionalOptions{MaxOperations: 1 << 20, MaxTruthVariables: 20})
		if err != nil {
			t.Fatal(err)
		}
		if converted.Graph.InputCount() != variables {
			t.Fatalf("input count: got %d want %d", converted.Graph.InputCount(), variables)
		}
		var got uint64
		for assignment := 0; assignment < 1<<variables; assignment++ {
			inputs := make([]bool, variables)
			for index := range inputs {
				inputs[index] = assignment>>index&1 != 0
			}
			outputs, evaluateErr := converted.Graph.Evaluate(inputs)
			if evaluateErr != nil {
				t.Fatal(evaluateErr)
			}
			if outputs[0] {
				got |= uint64(1) << assignment
			}
		}
		if got != want {
			t.Fatalf("%q: got %016x want %016x", formula, got, want)
		}
	})
}
