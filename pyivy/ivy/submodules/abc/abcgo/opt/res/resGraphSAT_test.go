package res

import (
	"testing"

	"github.com/glycerine/abcgo/aig/aig"
	"github.com/glycerine/abcgo/sat/bsat"
)

func TestGraphSATEncodesPhasedAIGNodes(t *testing.T) {
	g := aig.NewGraph()
	a, b, c := g.NewInput(), g.NewInput(), g.NewInput()
	and := g.AND(a.Not(), b)
	xor := g.NewXor(and, c.Not())
	buf := g.AddOutput(xor.Not())
	g.AddOutput(g.OR(a, c))
	phaseXor := g.AddOutput(g.NewXor(a.Not(), b.Not()))

	s := bsat.NewSolver()
	reserved := s.NewVariable()
	if !s.AddClause(bsat.Positive(reserved)) {
		t.Fatal("could not reserve an unrelated SAT variable")
	}
	p, err := Res_SatEncodeGraph(s, g)
	if err != nil {
		t.Fatal(err)
	}
	for assignment := 0; assignment < 8; assignment++ {
		wantInputs := []bool{assignment&1 != 0, assignment&2 != 0, assignment&4 != 0}
		wantOutputs, err := g.Evaluate(wantInputs)
		if err != nil {
			t.Fatal(err)
		}
		constraints := []aig.Edge{a, b, c, buf, phaseXor}
		values := []bool{wantInputs[0], wantInputs[1], wantInputs[2], wantOutputs[0], wantOutputs[2]}
		if got, status, err := p.Solve(constraints, values); err != nil || status != bsat.Satisfiable {
			t.Fatalf("assignment %d expected satisfiable output constraint: status=%v err=%v", assignment, status, err)
		} else if evaluated, err := g.Evaluate(got); err != nil || evaluated[0] != wantOutputs[0] {
			t.Fatalf("assignment %d returned model %v, outputs=%v err=%v", assignment, got, evaluated, err)
		}
	}
}

func TestGraphSATSolvesConstrainedSimulationAndEquivalence(t *testing.T) {
	g := aig.NewGraph()
	a, b := g.NewInput(), g.NewInput()
	left := g.XOR(a, b)
	right := g.OR(g.AND(a, b.Not()), g.AND(a.Not(), b))
	if equivalent, err := Res_SatProveGraphEquivalent(g, left, right); err != nil || !equivalent {
		t.Fatalf("equivalent XOR forms: got %v, %v", equivalent, err)
	}
	if equivalent, err := Res_SatProveGraphEquivalent(g, left, left.Not()); err != nil || equivalent {
		t.Fatalf("complemented outputs: got equivalent=%v, err=%v", equivalent, err)
	}

	p, err := Res_SatEncodeGraph(bsat.NewSolver(), g)
	if err != nil {
		t.Fatal(err)
	}
	inputs, status, err := p.Solve([]aig.Edge{left, a}, []bool{true, true})
	if err != nil || status != bsat.Satisfiable || len(inputs) != 2 || !inputs[0] || inputs[1] {
		t.Fatalf("expected a=1,b=0 witness, got %v status=%v err=%v", inputs, status, err)
	}
	if _, status, err := p.Solve([]aig.Edge{left, left.Not()}, []bool{true, true}); err != nil || status != bsat.Unsatisfiable {
		t.Fatalf("contradictory edge assumptions: status=%v err=%v", status, err)
	}
}

func TestGraphSATRejectsInvalidInputs(t *testing.T) {
	if _, err := Res_SatEncodeGraph(nil, aig.NewGraph()); err == nil {
		t.Fatal("expected nil solver error")
	}
	p, err := Res_SatEncodeGraph(bsat.NewSolver(), aig.NewGraph())
	if err != nil {
		t.Fatal(err)
	}
	if _, _, err := p.Solve([]aig.Edge{aig.InvalidEdge}, []bool{true}); err == nil {
		t.Fatal("expected invalid edge error")
	}
	if _, _, err := p.Solve([]aig.Edge{aig.Edge(0)}, nil); err == nil {
		t.Fatal("expected mismatched constraint lengths error")
	}
}
