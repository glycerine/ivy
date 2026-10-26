package abc

import (
	"errors"
	"slices"
	"sync"
	"testing"
)

func aigTestCI(t testing.TB, m *AigManager) AigLit {
	t.Helper()
	l, err := m.CreateCI()
	if err != nil {
		t.Fatal(err)
	}
	return l
}
func aigTestAnd(t testing.TB, m *AigManager, a, b AigLit) AigLit {
	t.Helper()
	l, err := m.And(a, b)
	if err != nil {
		t.Fatal(err)
	}
	return l
}
func aigTestCO(t testing.TB, m *AigManager, a AigLit) AigID {
	t.Helper()
	id, err := m.CreateCO(a)
	if err != nil {
		t.Fatal(err)
	}
	return id
}
func aigTestCheck(t testing.TB, m *AigManager) {
	t.Helper()
	if err := m.Check(); err != nil {
		t.Fatal(err)
	}
}

func TestAigAndIdentities(t *testing.T) {
	m := NewAigManager()
	a, b := aigTestCI(t, m), aigTestCI(t, m)
	for _, v := range []struct{ a, b, want AigLit }{
		{a, a, a}, {a, a.Not(), AigConst0}, {a, AigConst0, AigConst0}, {a, AigConst1, a},
		{AigConst0, a, AigConst0}, {AigConst1, a, a}, {AigConst1, AigConst1, AigConst1},
		{AigConst0, AigConst0, AigConst0}, {AigConst1, AigConst0, AigConst0},
	} {
		if got := aigTestAnd(t, m, v.a, v.b); got != v.want {
			t.Fatal(v, got)
		}
	}
	ab := aigTestAnd(t, m, b, a)
	if aigTestAnd(t, m, a, b) != ab || m.AndCount() != 1 {
		t.Fatal("structural hash")
	}
	obj, err := m.Object(ab.ID())
	if err != nil || obj.Fanin0 != a || obj.Fanin1 != b || obj.RefCount != 0 || obj.Level != 1 || obj.Phase {
		t.Fatal(obj, err)
	}
	first, _ := m.Object(a.ID())
	if first.RefCount != 1 {
		t.Fatal(first)
	}
	obj.RefCount = 99
	again, _ := m.Object(ab.ID())
	if again.RefCount != 0 {
		t.Fatal("object aliases")
	}
	aigTestCheck(t, m)
}
func TestAigExhaustiveBooleanOperations(t *testing.T) {
	// Each primitive's expected function is evaluated directly, without the
	// graph evaluator or the construction formulas used by the production code.
	for _, neg := range []uint{0, 1, 2, 3, 4, 5, 6, 7} {
		m := NewAigManager()
		a, b, c := aigTestCI(t, m), aigTestCI(t, m), aigTestCI(t, m)
		a = a.NotIf(neg&1 != 0)
		b = b.NotIf(neg&2 != 0)
		c = c.NotIf(neg&4 != 0)
		and := aigTestAnd(t, m, a, b)
		or, err := m.Or(a, b)
		if err != nil {
			t.Fatal(err)
		}
		xor, err := m.Xor(a, b)
		if err != nil {
			t.Fatal(err)
		}
		mux, err := m.Mux(a, b, c)
		if err != nil {
			t.Fatal(err)
		}
		maj, err := m.Majority(a, b, c)
		if err != nil {
			t.Fatal(err)
		}
		all, err := m.Multi([]AigLit{a, b, c}, AigAND)
		if err != nil {
			t.Fatal(err)
		}
		parity, err := m.Multi([]AigLit{a, b, c}, AigXOR)
		if err != nil {
			t.Fatal(err)
		}
		for _, l := range []AigLit{and, or, xor, mux, maj, all, parity, AigConst1, AigConst0} {
			aigTestCO(t, m, l)
		}
		aigTestCheck(t, m)
		for assignment := 0; assignment < 8; assignment++ {
			inputs := []bool{assignment&1 != 0, assignment&2 != 0, assignment&4 != 0}
			av, bv, cv := inputs[0] != (neg&1 != 0), inputs[1] != (neg&2 != 0), inputs[2] != (neg&4 != 0)
			selected := cv
			if av {
				selected = bv
			}
			votes := 0
			for _, v := range []bool{av, bv, cv} {
				if v {
					votes++
				}
			}
			want := []bool{av && bv, av || bv, av != bv, selected, votes >= 2, av && bv && cv, (av != bv) != cv, true, false}
			got, err := m.Evaluate(inputs)
			if err != nil || !slices.Equal(got, want) {
				t.Fatalf("neg %d assignment %d: got %v want %v (%v)", neg, assignment, got, want, err)
			}
		}
	}
}
func TestAigCleanupAndCloning(t *testing.T) {
	m := NewAigManager()
	a, b, c := aigTestCI(t, m), aigTestCI(t, m), aigTestCI(t, m)
	live := aigTestAnd(t, m, a, b)
	dead := aigTestAnd(t, m, a.Not(), b)
	dead2 := aigTestAnd(t, m, dead, c)
	buf, err := m.CreateBuffer(dead2)
	if err != nil {
		t.Fatal(err)
	}
	aigTestCO(t, m, live)
	before := m.Clone()
	if removed := m.Cleanup(); removed != 3 {
		t.Fatal("removed", removed)
	}
	if _, err := m.Object(dead.ID()); !errors.Is(err, ErrInvalidLiteral) {
		t.Fatal("stale ID accepted")
	}
	if _, err := m.And(dead, c); !errors.Is(err, ErrInvalidLiteral) {
		t.Fatal("stale literal accepted")
	}
	if _, err := m.Object(buf.ID()); err == nil {
		t.Fatal("dangling buffer survived")
	}
	if m.Cleanup() != 0 {
		t.Fatal("cleanup not idempotent")
	}
	aigTestCheck(t, m)
	aigTestCheck(t, before)
	revived := aigTestAnd(t, m, a.Not(), b)
	if revived.ID() <= buf.ID() {
		t.Fatal("ID reused")
	}
	if _, err := before.Object(revived.ID()); err == nil {
		t.Fatal("clone shares object storage")
	}
	for assignment := 0; assignment < 8; assignment++ {
		inputs := []bool{assignment&1 != 0, assignment&2 != 0, assignment&4 != 0}
		got, _ := m.Evaluate(inputs)
		want, _ := before.Evaluate(inputs)
		if !slices.Equal(got, want) {
			t.Fatal(got, want)
		}
	}
	copyInputs := m.Inputs()
	copyInputs[0] = InvalidAigID
	copyOutputs := m.Outputs()
	copyOutputs[0] = InvalidAigID
	aigTestCheck(t, m)
}
func TestAigSequentialReplay(t *testing.T) {
	m := NewAigManager()
	pi, reg := aigTestCI(t, m), aigTestCI(t, m)
	aigTestCO(t, m, reg)
	aigTestCO(t, m, pi)
	if err := m.SetRegisterCount(1); err != nil {
		t.Fatal(err)
	}
	if m.PrimaryInputCount() != 1 || m.PrimaryOutputCount() != 1 {
		t.Fatal("partition")
	}
	outputs, next, err := m.Step([]bool{true}, []bool{false})
	if err != nil || !slices.Equal(outputs, []bool{false}) || !slices.Equal(next, []bool{true}) {
		t.Fatal(outputs, next, err)
	}
	c := cexTestNew(t, 1, 1, 2)
	c.SetInput(0, 0, true)
	if err := m.ReplayCounterexample(c); err != nil {
		t.Fatal(err)
	}
	c.SetInput(0, 0, false)
	if err := m.ReplayCounterexample(c); err == nil {
		t.Fatal("false witness accepted")
	}
	c.Property = 1
	if err := m.ReplayCounterexample(c); err == nil {
		t.Fatal("next-state output treated as property")
	}
	if err := m.ReplayCounterexample(new(Counterexample)); err == nil {
		t.Fatal("zero trace accepted")
	}
	if err := m.ReplayCounterexample(nil); err == nil {
		t.Fatal("nil trace accepted")
	}
	for _, r := range []int{-1, 3} {
		if m.SetRegisterCount(r) == nil {
			t.Fatal("invalid register count")
		}
	}
	aigTestCheck(t, m)
}
func TestAigErrorsAndCheck(t *testing.T) {
	m := NewAigManager()
	a := aigTestCI(t, m)
	co := aigTestCO(t, m, a)
	for _, l := range []AigLit{InvalidAigLit, AigLit(co) << 1, AigLit(m.ObjectCount()+100) << 1} {
		before := m.ObjectCount()
		if _, err := m.And(l, a); err == nil {
			t.Fatal(l)
		}
		if _, err := m.Or(l, a); err == nil {
			t.Fatal(l)
		}
		if _, err := m.Xor(l, a); err == nil {
			t.Fatal(l)
		}
		if _, err := m.Mux(a, l, a); err == nil {
			t.Fatal(l)
		}
		if _, err := m.Majority(a, l, a); err == nil {
			t.Fatal(l)
		}
		if _, err := m.CreateCO(l); err == nil {
			t.Fatal(l)
		}
		if _, err := m.CreateBuffer(l); err == nil {
			t.Fatal(l)
		}
		if m.ObjectCount() != before {
			t.Fatal("failed validation mutated graph")
		}
	}
	if _, err := m.Multi(nil, AigAND); err == nil {
		t.Fatal("empty multi")
	}
	if _, err := m.Multi([]AigLit{a}, AigCO); err == nil {
		t.Fatal("invalid multi type")
	}
	if _, err := m.Evaluate(nil); err == nil {
		t.Fatal("wrong CI count")
	}
	if _, _, err := m.Step(nil, nil); err == nil {
		t.Fatal("wrong PI count")
	}
	if _, err := new(AigManager).CreateCI(); err == nil {
		t.Fatal("zero manager")
	}
	if new(AigManager).Check() == nil {
		t.Fatal("zero manager check")
	}
	if _, err := m.IthVar(-1); err == nil {
		t.Fatal("negative CI index")
	}
	var3, err := m.IthVar(3)
	if err != nil || m.InputCount() != 4 {
		t.Fatal(var3, err)
	}
	b := aigTestAnd(t, m, a, var3)
	aigTestCheck(t, m)
	corruptions := []func(*AigManager){
		func(d *AigManager) { d.objects[0].Phase = false },
		func(d *AigManager) { d.objects[0].Type = AigNone },
		func(d *AigManager) { d.objects[a.ID()].RefCount++ },
		func(d *AigManager) { d.objects[b.ID()].Level++ },
		func(d *AigManager) { d.objects[b.ID()].Fanin1 = b },
		func(d *AigManager) { d.objects[b.ID()].Fanin0 = AigConst1 },
		func(d *AigManager) { d.objects[b.ID()].Phase = !d.objects[b.ID()].Phase },
		func(d *AigManager) { delete(d.hash, aigHashKey{a, var3}) },
		func(d *AigManager) { d.cis[0] = InvalidAigID },
		func(d *AigManager) { d.cos[0] = InvalidAigID },
	}
	for i, corrupt := range corruptions {
		d := m.Clone()
		corrupt(d)
		if d.Check() == nil {
			t.Fatal("missed corruption", i)
		}
	}
	aigTestCheck(t, m)
}
func TestAigIndependentManagers(t *testing.T) {
	var group sync.WaitGroup
	for i := 0; i < 8; i++ {
		group.Go(func() {
			m := NewAigManager()
			a := aigTestCI(t, m)
			b := aigTestCI(t, m)
			aigTestCO(t, m, aigTestAnd(t, m, a, b))
			for j := 0; j < 20; j++ {
				m = m.Clone()
				aigTestCheck(t, m)
				m.Evaluate([]bool{true, false})
			}
		})
	}
	group.Wait()
}

// fuzzAigOperations builds a bounded DAG and carries an independent 16-row
// truth mask for each expression. The model uses ordinary bitwise operations;
// it never reads graph nodes or invokes graph evaluation to derive expectations.
func fuzzAigOperations(t testing.TB, data []byte) {
	t.Helper()
	m := NewAigManager()
	literals := []AigLit{AigConst0, AigConst1}
	truths := []uint16{0, 0xffff}
	for input := 0; input < 4; input++ {
		literals = append(literals, aigTestCI(t, m))
		var table uint16
		for assignment := 0; assignment < 16; assignment++ {
			if assignment&(1<<input) != 0 {
				table |= 1 << assignment
			}
		}
		truths = append(truths, table)
	}
	data = data[:min(len(data), 4*128)]
	for i := 0; i+3 < len(data); i += 4 {
		ai, bi, ci := int(data[i+1])%len(literals), int(data[i+2])%len(literals), int(data[i+3])%len(literals)
		a, b, c := literals[ai], literals[bi], literals[ci]
		av, bv, cv := truths[ai], truths[bi], truths[ci]
		if data[i]&8 != 0 {
			a = a.Not()
			av = ^av
		}
		if data[i]&16 != 0 {
			b = b.Not()
			bv = ^bv
		}
		if data[i]&32 != 0 {
			c = c.Not()
			cv = ^cv
		}
		var l AigLit
		var want uint16
		var err error
		switch data[i] % 6 {
		case 0:
			l, err = m.And(a, b)
			want = av & bv
		case 1:
			l, err = m.Or(a, b)
			want = av | bv
		case 2:
			l, err = m.Xor(a, b)
			want = av ^ bv
		case 3:
			l, err = m.Mux(a, b, c)
			want = (av & bv) | (^av & cv)
		case 4:
			l, err = m.Majority(a, b, c)
			want = (av & bv) | (av & cv) | (bv & cv)
		case 5:
			l, err = m.CreateBuffer(a)
			want = av
		}
		if err != nil {
			t.Fatal(err)
		}
		literals = append(literals, l)
		truths = append(truths, want)
		aigTestCheck(t, m)
	}
	start := max(0, len(literals)-8)
	for _, l := range literals[start:] {
		aigTestCO(t, m, l)
	}
	checkModel := func(graph *AigManager) {
		t.Helper()
		aigTestCheck(t, graph)
		for assignment := 0; assignment < 16; assignment++ {
			input := []bool{assignment&1 != 0, assignment&2 != 0, assignment&4 != 0, assignment&8 != 0}
			got, err := graph.Evaluate(input)
			if err != nil {
				t.Fatal(err)
			}
			for output, truth := range truths[start:] {
				if got[output] != (truth&(1<<assignment) != 0) {
					t.Fatalf("assignment %d output %d: got %v mask %04x", assignment, output, got[output], truth)
				}
			}
		}
	}
	checkModel(m)
	clone := m.Clone()
	m.Cleanup()
	checkModel(m)
	checkModel(clone)
}
func FuzzAigOperations(f *testing.F) {
	for _, seed := range [][]byte{{}, {0, 2, 3, 0, 2, 3, 4, 0, 3, 6, 7, 5}, {4, 2, 3, 4, 13, 3, 4, 0, 0, 6, 7, 0}, {5, 2, 0, 0, 5, 6, 0, 0}} {
		f.Add(seed)
	}
	f.Fuzz(func(t *testing.T, data []byte) { fuzzAigOperations(t, data) })
}
