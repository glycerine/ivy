package saig

import (
	"github.com/glycerine/abcgo/aig/aig"
	"testing"
)

func TestDemiterSecondHalfAgainstCombinationalCone(t *testing.T) {
	p := aig.NewGraph()
	a, b := p.NewInput(), p.NewInput()
	q0, q1 := p.NewInput(), p.NewInput()
	comb := p.AND(a, b)
	p.AddOutput(p.XOR(comb, q1))
	p.AddOutput(q0.Not())
	p.AddOutput(q1.Not())
	aig.Aig_ManSetRegNum(p, 2)
	var left, right *aig.Graph
	if Saig_ManDemiterSimpleDiff(p, &left, &right) != 1 {
		t.Fatal("valid independent halves rejected")
	}
	for mask := 0; mask < 8; mask++ {
		input := []bool{mask&1 != 0, mask&2 != 0, mask&4 != 0}
		l, err := left.Evaluate(input)
		if err != nil {
			t.Fatal(err)
		}
		r, err := right.Evaluate(input)
		if err != nil {
			t.Fatal(err)
		}
		if l[0] != (input[0] && input[1]) || r[0] != input[2] {
			t.Fatal("halves incorrectly oriented")
		}
	}
}

func TestGeneralDemiterLabelsLaterOutputCones(t *testing.T) {
	p := aig.NewGraph()
	a, b := p.NewInput(), p.NewInput()
	q0, q1 := p.NewInput(), p.NewInput()
	l, r := p.AND(q0, a), p.AND(q1, b)
	p.AddOutput(p.XOR(q0, q1))
	p.AddOutput(p.XOR(l, r))
	p.AddOutput(q0)
	p.AddOutput(q1)
	aig.Aig_ManSetRegNum(p, 2)
	if Saig_ManDemiter(p, nil, nil) != 1 {
		t.Fatal("diagnostic demiter rejected")
	}
	if !p.Node(l).MarkA || p.Node(l).MarkB || !p.Node(r).MarkB || p.Node(r).MarkA {
		t.Fatal("later output cones were not labeled in their register halves")
	}
}
