package core

import (
	"context"
	"github.com/glycerine/abcgo/sat/satoko"
	"math/rand"
	"testing"
	"time"
)

type circuit struct {
	nodes           []gate
	inputs, outputs []int32
	registers       int
}

func newCircuit(inputs int) *circuit {
	c := &circuit{nodes: []gate{{kind: Constant}}}
	for i := 0; i < inputs; i++ {
		c.inputs = append(c.inputs, c.add(Input, 0, 0, 0))
	}
	return c
}
func (c *circuit) add(kind int, a, b, d int32) int32 {
	lit := int32(len(c.nodes)) << 1
	c.nodes = append(c.nodes, gate{kind, [3]int32{a, b, d}})
	return lit
}
func (c *circuit) ObjectCount() int        { return len(c.nodes) }
func (c *circuit) InputLiterals() []int32  { return c.inputs }
func (c *circuit) OutputLiterals() []int32 { return c.outputs }
func (c *circuit) RegisterCount() int      { return c.registers }
func (c *circuit) Gate(id int32) (int32, int32, int32, int) {
	n := c.nodes[id]
	return n.fanins[0], n.fanins[1], n.fanins[2], n.kind
}
func TestCNFGatesExhaustive(t *testing.T) {
	for _, flip := range []bool{false, true} {
		c := newCircuit(3)
		a, b, d := c.inputs[0], c.inputs[1], c.inputs[2]
		and := c.add(And, a^1, b, 0)
		xor := c.add(Xor, a, b^1, 0)
		mux := c.add(Mux, d, a^1, b)
		same := c.add(Mux, d, a, a)
		buffer := c.add(Buffer, mux^1, 0, 0)
		output := c.add(Output, buffer^1, 0, 0)
		c.outputs = []int32{and, xor, mux, same, buffer, output}
		m, err := NewManager(c, Options{PolarFlip: flip})
		if err != nil {
			t.Fatal(err)
		}
		for mask := 0; mask < 8; mask++ {
			inputs := []bool{mask&1 != 0, mask&2 != 0, mask&4 != 0}
			for _, lit := range c.outputs {
				expected, _ := m.models[0].Value(inputs, lit)
				encoded := m.encode(0, lit)
				assumptions := make([]int, 0, 4)
				for i, value := range inputs {
					input := m.inputLiteral(i)
					if !value {
						input ^= 1
					}
					assumptions = append(assumptions, input)
				}
				assumed := encoded
				if !expected {
					assumed ^= 1
				}
				assumptions = append(assumptions, assumed)
				if result := m.solve(assumptions...); result != satoko.SATOKO_SAT {
					t.Fatalf("flip%v mask%d lit%d expectedSAT got%d", flip, mask, lit, result)
				}
				assumptions[len(assumptions)-1] ^= 1
				if result := m.solve(assumptions...); result != satoko.SATOKO_UNSAT {
					t.Fatalf("wrong gate value accepted: flip%v mask%d lit%d result%d", flip, mask, lit, result)
				}
			}
		}
	}
}
func TestIncrementalNodeChecksAndValidatedWitness(t *testing.T) {
	c := newCircuit(2)
	a, b := c.inputs[0], c.inputs[1]
	zero := c.add(And, a, a^1, 0)
	left, right := c.add(And, a, b, 0), c.add(And, b, a, 0)
	m, err := NewManager(c, Options{})
	if err != nil {
		t.Fatal(err)
	}
	if r, err := m.CheckLiteral(zero); err != nil || r.Status != Equivalent {
		t.Fatal(r, err)
	}
	if r, err := m.CheckPair(left, right); err != nil || r.Status != Equivalent {
		t.Fatal(r, err)
	}
	r, err := m.CheckPair(left, a)
	if err != nil || r.Status != Different || r.Counterexample == nil {
		t.Fatal(r, err)
	}
	x, _ := m.models[0].Value(r.Counterexample.Inputs, left)
	y, _ := m.models[0].Value(r.Counterexample.Inputs, a)
	if x == y {
		t.Fatal("invalid counterexample")
	}
	if stats := m.Statistics(); stats.Queries != 3 || stats.SATCalls < 4 {
		t.Fatal(stats)
	}
	if _, err := m.CheckLiteral(-1); err == nil {
		t.Fatal("invalid literal accepted")
	}
}

func TestConjunctionQueriesAndEquivalenceRemainDistinct(t *testing.T) {
	for _, flip := range []bool{false, true} {
		c := newCircuit(2)
		a, b := c.inputs[0], c.inputs[1]
		m, err := NewManager(c, Options{PolarFlip: flip})
		if err != nil {
			t.Fatal(err)
		}
		for _, tc := range []struct {
			x, y   int32
			status Status
		}{
			{a, a, Different}, {a, a ^ 1, Equivalent}, {a, b, Different}, {a ^ 1, b, Different}, {0, b, Equivalent}, {1, b, Different},
		} {
			got, err := m.CheckConjunction(tc.x, tc.y)
			if err != nil || got.Status != tc.status {
				t.Fatalf("flip=%v literals=%d,%d got=%+v err=%v", flip, tc.x, tc.y, got, err)
			}
			if got.Status == Different {
				x, _ := m.Model().Value(got.Counterexample.Inputs, tc.x)
				y, _ := m.Model().Value(got.Counterexample.Inputs, tc.y)
				if !x || !y {
					t.Fatalf("invalid witness %+v", got.Counterexample)
				}
			}
		}
		got, err := m.CheckPair(a, a)
		if err != nil || got.Status != Equivalent {
			t.Fatalf("equality changed: %+v %v", got, err)
		}
	}
}
func TestVerifyTwoPairsAndMiter(t *testing.T) {
	left, right := newCircuit(2), newCircuit(2)
	left.outputs = []int32{left.add(And, 2, 4, 0)}
	right.outputs = []int32{right.add(And, 4, 2, 0)}
	if r, err := VerifyTwo(left, right, Options{}); err != nil || r.Status != Equivalent {
		t.Fatal(r, err)
	}
	right.outputs[0] ^= 1
	r, err := VerifyTwo(left, right, Options{})
	if err != nil || r.Status != Different || r.Counterexample.Output != 0 {
		t.Fatal(r, err)
	}
	pair := newCircuit(2)
	a := pair.add(And, 2, 4, 0)
	b := pair.add(And, 4, 2, 0)
	pair.outputs = []int32{a, b, 2, 4}
	r, err = VerifyPairs(pair, Options{})
	if err != nil || r.Status != Different || r.Counterexample.Output != 1 {
		t.Fatal(r, err)
	}
	pair.outputs = pair.outputs[:1]
	if _, err := VerifyPairs(pair); err == nil {
		t.Fatal("odd output miter accepted")
	}
	miter := newCircuit(1)
	miter.outputs = []int32{miter.add(And, 2, 3, 0), 0}
	if r, err := VerifyMiter(miter, Options{}); err != nil || r.Status != Equivalent {
		t.Fatal(r, err)
	}
	miter.outputs = append(miter.outputs, 2)
	if r, err := VerifyMiter(miter, Options{}); err != nil || r.Status != Different || r.Counterexample.Output != 2 {
		t.Fatal(r, err)
	}
}
func TestModelValidationAndSnapshot(t *testing.T) {
	for _, mutate := range []func(*circuit){func(c *circuit) { c.inputs = append(c.inputs, c.inputs[0]) }, func(c *circuit) { c.outputs = []int32{-1} }, func(c *circuit) { c.nodes[0].kind = Input }, func(c *circuit) { c.add(And, 2, 99, 0) }, func(c *circuit) { lit := c.add(Buffer, 0, 0, 0); c.nodes[lit>>1].fanins[0] = lit }, func(c *circuit) { c.add(Input, 0, 0, 0) }, func(c *circuit) { c.add(8, 0, 0, 0) }} {
		c := newCircuit(1)
		mutate(c)
		if _, err := NewModel(c); err == nil {
			t.Fatal("malformed network accepted")
		}
	}
	var nilCircuit *circuit
	if _, err := NewModel(nilCircuit); err == nil {
		t.Fatal("typed nil accepted")
	}
	c := newCircuit(1)
	c.outputs = []int32{2}
	model, err := NewModel(c)
	if err != nil {
		t.Fatal(err)
	}
	c.outputs[0] = 1
	c.nodes[1].kind = Absent
	got, err := model.Evaluate([]bool{false})
	if err != nil || got[0] {
		t.Fatal("snapshot changed", got, err)
	}
	if _, err := model.Evaluate(nil); err == nil {
		t.Fatal("wrong input dimensions accepted")
	}
}
func TestCancellationSequentialPolicyAndRecycling(t *testing.T) {
	c := newCircuit(2)
	a, b := c.add(And, 2, 4, 0), c.add(And, 4, 2, 0)
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	m, _ := NewManager(c, Options{Context: ctx})
	if r, err := m.CheckPair(a, b); err != nil || r.Status != Undecided {
		t.Fatal(r, err)
	}
	m, _ = NewManager(c, Options{Deadline: time.Now().Add(-time.Second)})
	if r, err := m.CheckLiteral(a); err != nil || r.Status != Undecided {
		t.Fatal(r, err)
	}
	m, _ = NewManager(c, Options{MaxSATVariables: 1, RecycleQueries: 0})
	for i := 0; i < 4; i++ {
		if r, err := m.CheckPair(a, b); err != nil || r.Status != Equivalent {
			t.Fatal(r, err)
		}
	}
	if stats := m.Statistics(); stats.Recycles != 3 {
		t.Fatal(stats)
	}
	c.outputs = []int32{2}
	c.registers = 1
	if _, err := NewManager(c); err == nil {
		t.Fatal("sequential model silently proved combinationally")
	}
	if _, err := NewManager(c, Options{TreatRegistersAsInputs: true}); err != nil {
		t.Fatal(err)
	}
}
func TestRandomCircuitProofsAgainstExhaustiveEnumeration(t *testing.T) {
	rng := rand.New(rand.NewSource(27))
	for trial := 0; trial < 120; trial++ {
		c := newCircuit(4)
		for i := 0; i < 16; i++ {
			a, b, d := int32(rng.Intn(2*len(c.nodes))), int32(rng.Intn(2*len(c.nodes))), int32(rng.Intn(2*len(c.nodes)))
			kind := []int{And, Xor, Mux, Buffer}[rng.Intn(4)]
			c.add(kind, a, b, d)
		}
		model, err := NewModel(c)
		if err != nil {
			t.Fatal(err)
		}
		m, err := NewManager(c, Options{PolarFlip: trial%2 == 0})
		if err != nil {
			t.Fatal(err)
		}
		for query := 0; query < 8; query++ {
			a, b := int32(rng.Intn(2*len(c.nodes))), int32(rng.Intn(2*len(c.nodes)))
			equal := true
			for mask := 0; mask < 16; mask++ {
				inputs := []bool{mask&1 != 0, mask&2 != 0, mask&4 != 0, mask&8 != 0}
				x, _ := model.Value(inputs, a)
				y, _ := model.Value(inputs, b)
				if x != y {
					equal = false
					break
				}
			}
			result, err := m.CheckPair(a, b)
			if err != nil || (result.Status == Equivalent) != equal || result.Status == Undecided {
				t.Fatalf("trial%d query%d equal%v result%v err%v", trial, query, equal, result, err)
			}
		}
	}
}

func TestConflictStatisticsAccumulateAcrossQueriesAndRecycling(t *testing.T) {
	c := newCircuit(7)
	left, right := c.inputs[0], c.inputs[6]
	for i := 1; i < 7; i++ {
		left = c.add(Xor, left, c.inputs[i], 0)
	}
	for i := 5; i >= 0; i-- {
		right = c.add(Xor, right, c.inputs[i], 0)
	}
	m, err := NewManager(c, Options{})
	if err != nil {
		t.Fatal(err)
	}
	if result, err := m.CheckPair(left, right); err != nil || result.Status != Equivalent {
		t.Fatal(result, err)
	}
	first := m.Statistics().Conflicts
	if first == 0 {
		t.Fatal("parity proof did not exercise conflict accounting")
	}
	if result, err := m.CheckLiteral(c.inputs[0]); err != nil || result.Status != Different {
		t.Fatal(result, err)
	}
	if m.Statistics().Conflicts < first {
		t.Fatal("a later solve erased previous conflicts")
	}
	m.options.MaxSATVariables = 1
	m.options.RecycleQueries = 0
	if _, err := m.CheckLiteral(c.inputs[1]); err != nil {
		t.Fatal(err)
	}
	if stats := m.Statistics(); stats.Recycles != 1 || stats.Conflicts < first {
		t.Fatal("recycling lost proof statistics", stats)
	}
}
func TestUnrelatedInputsAreNotAllocatedAsSATVariables(t *testing.T) {
	c := newCircuit(100)
	zero := c.add(And, c.inputs[0], c.inputs[0]^1, 0)
	m, err := NewManager(c, Options{})
	if err != nil {
		t.Fatal(err)
	}
	if result, err := m.CheckLiteral(zero); err != nil || result.Status != Equivalent {
		t.Fatal(result, err)
	}
	if variables := m.Statistics().Variables; variables != 3 {
		t.Fatalf("cone proof allocated %d variables, expected constant+one input+one gate", variables)
	}
}

func TestMutableConflictBudget(t *testing.T) {
	c := newCircuit(1)
	m, err := NewManager(c)
	if err != nil {
		t.Fatal(err)
	}
	if err = m.SetConflictLimit(-1); err == nil {
		t.Fatal("negative budget accepted")
	}
	oldSolver := m.solver
	if err = m.SetConflictLimit(0); err != nil {
		t.Fatal(err)
	}
	if m.solver != oldSolver || m.options.ConflictLimit != 0 {
		t.Fatal("budget update discarded solver state")
	}
	result, err := m.CheckPair(c.inputs[0], c.inputs[0])
	if err != nil || result.Status != Equivalent {
		t.Fatalf("query after budget update: %+v %v", result, err)
	}
}
