package abc

// Go-only projection tests for bmcFault.c GIA CNF helpers.

import (
	"errors"
	"reflect"
	"slices"
	"testing"
)

func faultCNFGraph(t *testing.T) *GiaManager {
	t.Helper()
	g := NewGiaManager()
	a, _ := g.AppendCI()
	b, _ := g.AppendCI()
	c, _ := g.AppendCI()
	x, err := g.AppendXorReal(a, b.Not())
	if err != nil {
		t.Fatal(err)
	}
	y, err := g.AppendMuxReal(c, x, a.Not())
	if err != nil {
		t.Fatal(err)
	}
	_, _ = g.AppendCO(x)
	_, _ = g.AppendCO(y.Not())
	return g
}

func faultCNFAssumption(literal SatLit, value bool) SatLit {
	if value {
		return literal
	}
	return literal.Not()
}

func checkFaultCNFProjection(t *testing.T, graph *GiaManager, relation *GiaFaultCNF, offset int) {
	t.Helper()
	inputs := relation.InputLiterals()
	outputs := relation.OutputLiterals()
	for mask := 0; mask < 1<<graph.InputCount(); mask++ {
		values := make([]bool, graph.InputCount())
		assumptions := make([]SatLit, graph.InputCount())
		for i := range values {
			values[i] = mask&(1<<i) != 0
			assumptions[i] = faultCNFAssumption(inputs[i], values[i])
		}
		want, err := graph.Evaluate(values)
		if err != nil {
			t.Fatal(err)
		}
		for output := range outputs {
			for _, value := range []bool{want[output], !want[output]} {
				solver := NewSatSolver()
				ok, err := relation.LoadInto(solver)
				if err != nil || !ok {
					t.Fatal(err, ok)
				}
				query := append(slices.Clone(assumptions), faultCNFAssumption(outputs[output], value))
				result, err := solver.Solve(query, SatLimits{})
				if err != nil || (result.Status == SatSatisfiable) != (value == want[output]) {
					t.Fatalf("offset=%d mask=%b output=%d value=%v want=%v status=%d err=%v", offset, mask, output, value, want[output], result.Status, err)
				}
			}
		}
	}
}

func TestGiaFaultCNFDeriveAndLift(t *testing.T) {
	graph := faultCNFGraph(t)
	relation, err := DeriveGiaFaultCNF(graph, GiaFaultCNFOptions{})
	if err != nil {
		t.Fatal(err)
	}
	if relation.VariableCount() < graph.InputCount()+1 || len(relation.ObjectLiterals()) != graph.ObjectCount() || len(relation.InputLiterals()) != graph.InputCount() || len(relation.OutputLiterals()) != graph.OutputCount() {
		t.Fatal(relation.VariableCount(), relation.ObjectLiterals(), relation.InputLiterals(), relation.OutputLiterals())
	}
	checkFaultCNFProjection(t, graph, relation, 0)
	originalClauses := relation.Clauses()
	originalObjects := relation.ObjectLiterals()
	lifted, err := relation.Lift(7, GiaFaultCNFOptions{})
	if err != nil {
		t.Fatal(err)
	}
	if lifted.VariableCount() != relation.VariableCount()+7 {
		t.Fatal(lifted.VariableCount(), relation.VariableCount())
	}
	for i, clause := range originalClauses {
		got := lifted.Clauses()[i]
		for k, literal := range clause {
			if got[k] != SatLit(uint64(literal.Var())+7)<<1|literal&1 {
				t.Fatal(i, k, literal, got[k])
			}
		}
	}
	for i, literal := range originalObjects {
		got, err := lifted.ObjectLiteral(GiaID(i))
		if err != nil || got != SatLit(uint64(literal.Var())+7)<<1|literal&1 {
			t.Fatal(i, literal, got, err)
		}
	}
	if !reflect.DeepEqual(originalClauses, relation.Clauses()) || !slices.Equal(originalObjects, relation.ObjectLiterals()) {
		t.Fatal("lift mutated source")
	}
	checkFaultCNFProjection(t, graph, lifted, 7)
}

func TestGiaFaultCNFObjectMappingAndOwnership(t *testing.T) {
	graph := faultCNFGraph(t)
	relation, err := DeriveGiaFaultCNF(graph, GiaFaultCNFOptions{})
	if err != nil {
		t.Fatal(err)
	}
	for id := GiaID(0); int(id) < graph.ObjectCount(); id++ {
		literal, err := relation.ObjectLiteral(id)
		if err != nil || !literal.Valid() || int(literal.Var()) >= relation.VariableCount() {
			t.Fatal(id, literal, err)
		}
	}
	clauses := relation.Clauses()
	inputs := relation.InputLiterals()
	objects := relation.ObjectLiterals()
	clauses[0][0] = InvalidSatLit
	inputs[0] = InvalidSatLit
	objects[0] = InvalidSatLit
	if relation.Clauses()[0][0] == InvalidSatLit || relation.InputLiterals()[0] == InvalidSatLit || relation.ObjectLiterals()[0] == InvalidSatLit {
		t.Fatal("result getters leaked storage")
	}
	if _, err := relation.ObjectLiteral(GiaID(graph.ObjectCount())); !errors.Is(err, ErrInvalidLiteral) {
		t.Fatal(err)
	}
}

func TestGiaFaultCNFLimitsAndCancellation(t *testing.T) {
	graph := faultCNFGraph(t)
	for _, options := range []GiaFaultCNFOptions{
		{MaxObjects: 1}, {MaxVariables: 1}, {MaxClauses: 1}, {MaxLiterals: 1},
	} {
		if result, err := DeriveGiaFaultCNF(graph, options); result != nil || !errors.Is(err, ErrLimit) {
			t.Fatal(options, result, err)
		}
	}
	if result, err := DeriveGiaFaultCNF(graph, GiaFaultCNFOptions{Stop: func() bool { return true }}); result != nil || !errors.Is(err, ErrCancelled) {
		t.Fatal(result, err)
	}
	relation, _ := DeriveGiaFaultCNF(graph, GiaFaultCNFOptions{})
	for _, tc := range []struct {
		offset  int
		options GiaFaultCNFOptions
		want    error
	}{
		{-1, GiaFaultCNFOptions{}, ErrDimension},
		{10, GiaFaultCNFOptions{MaxVariables: relation.VariableCount() + 9}, ErrLimit},
		{1, GiaFaultCNFOptions{Stop: func() bool { return true }}, ErrCancelled},
	} {
		if result, err := relation.Lift(tc.offset, tc.options); result != nil || !errors.Is(err, tc.want) {
			t.Fatal(tc, result, err)
		}
	}
}

func FuzzGiaFaultCNFProjection(f *testing.F) {
	for _, seed := range []byte{0, 1, 2, 3, 7, 15, 255} {
		f.Add(seed, seed^0x69)
	}
	f.Fuzz(func(t *testing.T, selector, mask byte) {
		graph := NewGiaManager()
		a, _ := graph.AppendCI()
		b, _ := graph.AppendCI()
		c, _ := graph.AppendCI()
		var root GiaLit
		var err error
		switch selector % 3 {
		case 0:
			root, err = graph.AppendAnd(a.NotIf(selector&4 != 0), b.NotIf(selector&8 != 0))
		case 1:
			root, err = graph.AppendXorReal(a.NotIf(selector&4 != 0), b.NotIf(selector&8 != 0))
		case 2:
			root, err = graph.AppendMuxReal(c, a.NotIf(selector&4 != 0), b.NotIf(selector&8 != 0))
		}
		if err != nil {
			t.Fatal(err)
		}
		_, _ = graph.AppendCO(root.NotIf(selector&16 != 0))
		relation, err := DeriveGiaFaultCNF(graph, GiaFaultCNFOptions{})
		if err != nil {
			t.Fatal(err)
		}
		inputs := relation.InputLiterals()
		values := []bool{mask&1 != 0, mask&2 != 0, mask&4 != 0}
		want, _ := graph.Evaluate(values)
		assumptions := make([]SatLit, 0, 4)
		for i := range values {
			assumptions = append(assumptions, faultCNFAssumption(inputs[i], values[i]))
		}
		assumptions = append(assumptions, faultCNFAssumption(relation.OutputLiterals()[0], !want[0]))
		solver := NewSatSolver()
		ok, err := relation.LoadInto(solver)
		if err != nil || !ok {
			t.Fatal(err, ok)
		}
		result, err := solver.Solve(assumptions, SatLimits{})
		if err != nil || result.Status != SatUnsatisfiable {
			t.Fatal(selector, mask, result.Status, err)
		}
	})
}
