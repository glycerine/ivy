package abc

// Tests for the Go-only ACEC polynomial, ordering and adder-tree foundation.

import (
	"errors"
	"math/big"
	"math/rand"
	"slices"
	"testing"
)

func acecAppendInputs(t *testing.T, graph *GiaManager, count int) []GiaLit {
	t.Helper()
	inputs := make([]GiaLit, count)
	for i := range inputs {
		var err error
		inputs[i], err = graph.AppendCI()
		if err != nil {
			t.Fatal(err)
		}
	}
	return inputs
}

func TestAcecSignFieldsAndBoxOwnership(t *testing.T) {
	adders := make([]int, 12)
	if err := AcecSetSignBit(adders, 1, 3, false, true); err != nil {
		t.Fatal(err)
	}
	if err := AcecSetSignBit(adders, 1, 7, true, true); err != nil {
		t.Fatal(err)
	}
	for _, row := range []struct {
		bit, box int
		second   bool
		want     bool
	}{{3, 1, false, true}, {7, 1, true, true}, {3, 0, false, false}} {
		got, err := AcecSignBit(adders, row.box, row.bit, row.second)
		if err != nil || got != row.want {
			t.Fatalf("sign field %+v: %v, %v", row, got, err)
		}
	}
	if err := AcecSetSignBit(adders, 1, 3, false, false); err != nil {
		t.Fatal(err)
	}
	if got, _ := AcecSignBit(adders, 1, 3, false); !got {
		t.Fatal("source-compatible false setter cleared a bit")
	}
	if _, err := AcecSignBit(adders, -1, 0, false); !errors.Is(err, ErrDimension) {
		t.Fatalf("negative box: %v", err)
	}
	if _, err := AcecSignBit(adders, int(^uint(0)>>1), 0, false); !errors.Is(err, ErrDimension) {
		t.Fatalf("overflowing box: %v", err)
	}

	graph := NewGiaManager()
	inputs := acecAppendInputs(t, graph, 2)
	leaves := [][]GiaLit{{inputs[0], inputs[1]}}
	box, err := NewAcecBox(graph, nil, leaves, nil, nil, nil)
	if err != nil {
		t.Fatal(err)
	}
	leaves[0][0] = GiaConst0
	if box.Leaves[0][0] != inputs[0] {
		t.Fatal("box borrowed caller ranks")
	}
}

func TestAcecHalfAndFullAddersExhaustive(t *testing.T) {
	for count := 2; count <= 3; count++ {
		graph := NewGiaManager()
		inputs := acecAppendInputs(t, graph, count)
		var outputs [2]GiaLit
		var err error
		if count == 2 {
			outputs, err = AppendAcecHalfAdder(graph, [2]GiaLit{inputs[0], inputs[1]})
		} else {
			outputs, err = AppendAcecFullAdder(graph, [3]GiaLit{inputs[0], inputs[1], inputs[2]})
		}
		if err != nil {
			t.Fatal(err)
		}
		for _, output := range outputs {
			if _, err = graph.AppendCO(output); err != nil {
				t.Fatal(err)
			}
		}
		if err = graph.Check(); err != nil {
			t.Fatal(err)
		}
		for assignment := 0; assignment < 1<<count; assignment++ {
			bits := make([]bool, count)
			ones := 0
			for i := range bits {
				bits[i] = assignment>>i&1 != 0
				if bits[i] {
					ones++
				}
			}
			got, err := graph.Evaluate(bits)
			if err != nil {
				t.Fatal(err)
			}
			if got[0] != (ones&1 != 0) || got[1] != (ones >= 2) {
				t.Fatalf("%d-input adder assignment %b: %v", count, assignment, got)
			}
		}
	}
}

func TestAcecAdderTreeWeightedValueAndAtomicFailure(t *testing.T) {
	graph := NewGiaManager()
	inputs := acecAppendInputs(t, graph, 5)
	ranks := [][]GiaLit{{inputs[0], inputs[1], inputs[2]}, {inputs[3], inputs[4]}}
	roots, err := BuildAcecAdderTree(graph, ranks, AcecPolynomialOptions{})
	if err != nil {
		t.Fatal(err)
	}
	for _, root := range roots {
		if _, err = graph.AppendCO(root); err != nil {
			t.Fatal(err)
		}
	}
	for assignment := 0; assignment < 32; assignment++ {
		bits := make([]bool, 5)
		want := 0
		for i := range bits {
			bits[i] = assignment>>i&1 != 0
			if bits[i] {
				if i < 3 {
					want++
				} else {
					want += 2
				}
			}
		}
		gotBits, err := graph.Evaluate(bits)
		if err != nil {
			t.Fatal(err)
		}
		got := 0
		for i, bit := range gotBits {
			if bit {
				got += 1 << i
			}
		}
		if got != want {
			t.Fatalf("assignment %05b: got %d want %d", assignment, got, want)
		}
	}

	limited, err := NewGiaManagerWithLimit(4)
	if err != nil {
		t.Fatal(err)
	}
	two := acecAppendInputs(t, limited, 2)
	before := limited.ObjectCount()
	if _, err = AppendAcecHalfAdder(limited, [2]GiaLit{two[0], two[1]}); !errors.Is(err, ErrLimit) {
		t.Fatalf("limited half adder: %v", err)
	}
	if limited.ObjectCount() != before {
		t.Fatal("failed half adder changed graph")
	}
}

func buildAcecPolynomialFixture(t *testing.T) *GiaManager {
	t.Helper()
	graph := NewGiaManager()
	inputs := acecAppendInputs(t, graph, 3)
	and, err := graph.AppendAnd(inputs[0].Not(), inputs[1])
	if err != nil {
		t.Fatal(err)
	}
	xor, err := graph.AppendXorReal(inputs[0], inputs[2].Not())
	if err != nil {
		t.Fatal(err)
	}
	mux, err := graph.AppendMuxReal(inputs[1], xor, and.Not())
	if err != nil {
		t.Fatal(err)
	}
	for _, output := range []GiaLit{and, xor.Not(), mux} {
		if _, err = graph.AppendCO(output); err != nil {
			t.Fatal(err)
		}
	}
	return graph
}

func TestAcecPolynomialMatchesIndependentGraphEvaluation(t *testing.T) {
	graph := buildAcecPolynomialFixture(t)
	for _, signed := range []bool{false, true} {
		polynomial, err := ExtractAcecPolynomial(graph, signed, AcecPolynomialOptions{})
		if err != nil {
			t.Fatal(err)
		}
		if polynomial.IsZero() || polynomial.String() == "" {
			t.Fatal("missing polynomial")
		}
		for assignment := 0; assignment < 8; assignment++ {
			inputs := []bool{assignment&1 != 0, assignment&2 != 0, assignment&4 != 0}
			objects, err := graph.evaluateValues(inputs)
			if err != nil {
				t.Fatal(err)
			}
			got, err := polynomial.Evaluate(objects)
			if err != nil {
				t.Fatal(err)
			}
			outputs, err := graph.Evaluate(inputs)
			if err != nil {
				t.Fatal(err)
			}
			want := int64(0)
			for i, bit := range outputs {
				if bit {
					weight := int64(1 << i)
					if signed && i == len(outputs)-1 {
						weight = -weight
					}
					want += weight
				}
			}
			if got.Cmp(big.NewInt(want)) != 0 {
				t.Fatalf("signed=%v assignment=%03b polynomial=%s got %s want %d", signed, assignment, polynomial.String(), got, want)
			}
		}
		terms := polynomial.Terms()
		if len(terms) != 0 {
			terms[0].Coefficient.SetInt64(123456)
			terms[0].Variables = nil
			if polynomial.String() == "123456" {
				t.Fatal("Terms exposed polynomial storage")
			}
		}
	}
}

func TestAcecPolynomialLimitsCancellationAndBooleanReduction(t *testing.T) {
	graph := NewGiaManager()
	inputs := acecAppendInputs(t, graph, 2)
	xor, err := graph.AppendXorReal(inputs[0], inputs[1])
	if err != nil {
		t.Fatal(err)
	}
	if _, err = graph.AppendCO(xor); err != nil {
		t.Fatal(err)
	}
	polynomial, err := ExtractAcecPolynomial(graph, false, AcecPolynomialOptions{})
	if err != nil {
		t.Fatal(err)
	}
	if got := polynomial.String(); got != "1*x1-2*x1*x2+1*x2" {
		t.Fatalf("XOR polynomial %q", got)
	}
	if _, err = ExtractAcecPolynomial(graph, false, AcecPolynomialOptions{MaxTerms: 2}); !errors.Is(err, ErrLimit) {
		t.Fatalf("term limit: %v", err)
	}
	if _, err = ExtractAcecPolynomial(graph, false, AcecPolynomialOptions{MaxDegree: 1}); !errors.Is(err, ErrLimit) {
		t.Fatalf("degree limit: %v", err)
	}
	if _, err = ExtractAcecPolynomial(graph, false, AcecPolynomialOptions{Stop: func() bool { return true }}); !errors.Is(err, ErrCancelled) {
		t.Fatalf("cancellation: %v", err)
	}
	variable := acecVariable(inputs[0].ID())
	square, err := acecMultiply(variable, variable, AcecPolynomialOptions{MaxTerms: 10, MaxDegree: 10})
	if err != nil || square.String() != variable.String() {
		t.Fatalf("Boolean x*x reduction: %s, %v", square.String(), err)
	}
}

func TestAcecOrderAndXORCollection(t *testing.T) {
	graph := NewGiaManager()
	inputs := acecAppendInputs(t, graph, 3)
	inner, err := graph.AppendXorReal(inputs[0], inputs[1])
	if err != nil {
		t.Fatal(err)
	}
	root, err := graph.AppendXorReal(inner, inputs[2])
	if err != nil {
		t.Fatal(err)
	}
	if _, err = graph.AppendCO(root); err != nil {
		t.Fatal(err)
	}
	chain, err := CollectAcecLastXOR(graph, AcecPolynomialOptions{})
	if err != nil {
		t.Fatal(err)
	}
	if want := []GiaID{root.ID(), inner.ID()}; !slices.Equal(chain, want) {
		t.Fatalf("chain %v want %v", chain, want)
	}
	counts, err := AnalyzeAcecXORChains(graph, AcecPolynomialOptions{})
	if err != nil || !slices.Equal(counts, []int{2}) {
		t.Fatalf("counts %v, %v", counts, err)
	}
	order, err := FindAcecPolynomialOrder(graph, nil, nil, AcecPolynomialOptions{})
	if err != nil {
		t.Fatal(err)
	}
	last := 0
	for _, id := range append(graph.Inputs(), inner.ID(), root.ID()) {
		if order[id] <= last {
			t.Fatalf("non-topological order %v", order)
		}
		last = order[id]
	}

	if _, err = graph.AppendCO(inner); err != nil {
		t.Fatal(err)
	}
	chain, err = CollectAcecLastXOR(graph, AcecPolynomialOptions{})
	if err != nil || len(chain) != 0 { // The last output now directly drives a shared XOR.
		t.Fatalf("shared last XOR chain %v, %v", chain, err)
	}
}

func TestDuplicateAcecTopMostRange(t *testing.T) {
	short := NewGiaManager()
	if _, err := short.AppendCO(GiaConst0); err != nil {
		t.Fatal(err)
	}
	if _, err := DuplicateAcecTopMostRange(short, GiaCofactorOptions{}); !errors.Is(err, ErrDimension) {
		t.Fatalf("short range: %v", err)
	}
	graph := NewGiaManager()
	for i := 0; i < 52; i++ {
		if _, err := graph.AppendCO(GiaConst0); err != nil {
			t.Fatal(err)
		}
	}
	result, err := DuplicateAcecTopMostRange(graph, GiaCofactorOptions{})
	if err != nil {
		t.Fatal(err)
	}
	if result.Graph.OutputCount() != 7 || len(result.Inputs) != 0 {
		t.Fatalf("range dimensions: outputs=%d inputs=%v", result.Graph.OutputCount(), result.Inputs)
	}
}

func TestAcecPolynomialDeterministicModel5000(t *testing.T) {
	rng := rand.New(rand.NewSource(0xacec))
	for trial := 0; trial < 5000; trial++ {
		graph := NewGiaManager()
		inputs := acecAppendInputs(t, graph, 3)
		signals := slices.Clone(inputs)
		for gate := 0; gate < 3; gate++ {
			a := signals[rng.Intn(len(signals))].NotIf(rng.Intn(2) != 0)
			b := signals[rng.Intn(len(signals))].NotIf(rng.Intn(2) != 0)
			var result GiaLit
			var err error
			if rng.Intn(2) == 0 {
				result, err = graph.AppendAndSimplified(a, b)
			} else {
				result, err = graph.AppendXorSimplified(a, b)
			}
			if err != nil {
				t.Fatal(err)
			}
			signals = append(signals, result)
		}
		if _, err := graph.AppendCO(signals[len(signals)-1]); err != nil {
			t.Fatal(err)
		}
		polynomial, err := ExtractAcecPolynomial(graph, false, AcecPolynomialOptions{MaxTerms: 256, MaxDegree: 16})
		if err != nil {
			t.Fatalf("trial %d: %v", trial, err)
		}
		assignment := rng.Intn(8)
		bits := []bool{assignment&1 != 0, assignment&2 != 0, assignment&4 != 0}
		objects, _ := graph.evaluateValues(bits)
		got, err := polynomial.Evaluate(objects)
		if err != nil {
			t.Fatal(err)
		}
		outputs, _ := graph.Evaluate(bits)
		want := int64(0)
		if outputs[0] {
			want = 1
		}
		if got.Int64() != want {
			t.Fatalf("trial %d: got %s want %d", trial, got, want)
		}
	}
}

func FuzzAcecPolynomial(f *testing.F) {
	f.Add([]byte{3, 0, 1, 0, 1, 2, 1, 0, 2})
	f.Add([]byte{2, 1, 0, 0, 0, 1})
	f.Fuzz(func(t *testing.T, data []byte) {
		if len(data) == 0 {
			return
		}
		inputCount := int(data[0]%4) + 1
		graph := NewGiaManager()
		inputs := acecAppendInputs(t, graph, inputCount)
		signals := slices.Clone(inputs)
		cursor := 1
		for gate := 0; gate < 8 && cursor+2 < len(data); gate++ {
			a := signals[int(data[cursor])%len(signals)].NotIf(data[cursor]&0x80 != 0)
			b := signals[int(data[cursor+1])%len(signals)].NotIf(data[cursor+1]&0x80 != 0)
			op := data[cursor+2] % 2
			cursor += 3
			var result GiaLit
			var err error
			if op == 0 {
				result, err = graph.AppendAndSimplified(a, b)
			} else {
				result, err = graph.AppendXorSimplified(a, b)
			}
			if err != nil {
				t.Fatal(err)
			}
			signals = append(signals, result)
		}
		output := signals[len(signals)-1]
		if _, err := graph.AppendCO(output); err != nil {
			t.Fatal(err)
		}
		polynomial, err := ExtractAcecPolynomial(graph, false, AcecPolynomialOptions{MaxTerms: 1 << 14, MaxDegree: 64})
		if err != nil {
			t.Fatal(err)
		}
		for assignment := 0; assignment < 1<<inputCount; assignment++ {
			bits := make([]bool, inputCount)
			for i := range bits {
				bits[i] = assignment>>i&1 != 0
			}
			objects, err := graph.evaluateValues(bits)
			if err != nil {
				t.Fatal(err)
			}
			got, err := polynomial.Evaluate(objects)
			if err != nil {
				t.Fatal(err)
			}
			outputs, err := graph.Evaluate(bits)
			if err != nil {
				t.Fatal(err)
			}
			want := int64(0)
			if outputs[0] {
				want = 1
			}
			if got.Int64() != want {
				t.Fatalf("assignment %b polynomial=%s got=%s want=%d", assignment, polynomial.String(), got, want)
			}
		}
	})
}
