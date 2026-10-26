package abc

// Go-only semantic, model and fuzz tests for src/aig/gia/giaTis.c.

import (
	"errors"
	"math/rand"
	"slices"
	"testing"
)

func appendTisMuxPattern(t testing.TB, m *GiaManager, control, whenTrue, whenFalse GiaLit) (GiaLit, GiaLit, GiaLit) {
	t.Helper()
	left, err := m.AppendAnd(control, whenTrue)
	if err != nil {
		t.Fatal(err)
	}
	right, err := m.AppendAnd(control.Not(), whenFalse)
	if err != nil {
		t.Fatal(err)
	}
	root, err := m.AppendAnd(left.Not(), right.Not())
	if err != nil {
		t.Fatal(err)
	}
	return root, left, right
}

func appendTisXorPattern(t testing.TB, m *GiaManager, a, b GiaLit) GiaLit {
	t.Helper()
	left, err := m.AppendAnd(a, b.Not())
	if err != nil {
		t.Fatal(err)
	}
	right, err := m.AppendAnd(a.Not(), b)
	if err != nil {
		t.Fatal(err)
	}
	root, err := m.AppendAnd(left.Not(), right.Not())
	if err != nil {
		t.Fatal(err)
	}
	return root
}

func TestGiaTisDuplicateRecognizesMuxAndXor(t *testing.T) {
	m := NewGiaManager()
	m.Name, m.Spec = "tis", "tis.aig"
	inputs := make([]GiaLit, 5)
	for i := range inputs {
		inputs[i], _ = m.AppendCI()
	}
	mux, _, _ := appendTisMuxPattern(t, m, inputs[0], inputs[1].Not(), inputs[2])
	xor := appendTisXorPattern(t, m, inputs[1], inputs[3])
	constraint, _ := m.AppendAnd(inputs[0], inputs[3])
	for _, driver := range []GiaLit{mux, xor.Not(), constraint, inputs[4]} {
		if _, err := m.AppendCO(driver); err != nil {
			t.Fatal(err)
		}
	}
	if err := m.SetRegisterCount(1); err != nil {
		t.Fatal(err)
	}
	if err := m.SetConstraintCount(1); err != nil {
		t.Fatal(err)
	}
	before := slices.Clone(m.objects)
	converted, err := m.DuplicateWithRecognizedMuxes(GiaTisOptions{})
	if err != nil {
		t.Fatal(err)
	}
	if converted.MuxCount() != 1 || converted.XorCount() != 1 || converted.BufferCount() != 0 || converted.Name != m.Name || converted.Spec != m.Spec || converted.RegisterCount() != 1 || converted.ConstraintCount() != 1 {
		t.Fatal(converted.MuxCount(), converted.XorCount(), converted.ObjectCount())
	}
	if converted.ObjectCount() >= m.ObjectCount() || converted.Hashing() || !slices.Equal(before, m.objects) {
		t.Fatal("cleanup, hashing, or source ownership", converted.ObjectCount(), m.ObjectCount())
	}
	for assignment := 0; assignment < 1<<len(inputs); assignment++ {
		values := make([]bool, len(inputs))
		for i := range values {
			values[i] = assignment&(1<<i) != 0
		}
		want, wantErr := m.Evaluate(values)
		got, gotErr := converted.Evaluate(values)
		if wantErr != nil || gotErr != nil || !slices.Equal(got, want) {
			t.Fatal(assignment, got, want, gotErr, wantErr)
		}
	}
}

func TestGiaTisDuplicateKeepsDoublySharedPattern(t *testing.T) {
	m := NewGiaManager()
	c, _ := m.AppendCI()
	a, _ := m.AppendCI()
	b, _ := m.AppendCI()
	root, left, right := appendTisMuxPattern(t, m, c, a, b)
	for _, driver := range []GiaLit{root, left, right} {
		m.AppendCO(driver)
	}
	converted, err := m.DuplicateWithRecognizedMuxes(GiaTisOptions{})
	if err != nil || converted.MuxCount() != 0 || converted.XorCount() != 0 {
		t.Fatal(converted, err)
	}
	for assignment := 0; assignment < 8; assignment++ {
		values := []bool{assignment&1 != 0, assignment&2 != 0, assignment&4 != 0}
		want, _ := m.Evaluate(values)
		got, gotErr := converted.Evaluate(values)
		if gotErr != nil || !slices.Equal(got, want) {
			t.Fatal(assignment, got, want, gotErr)
		}
	}
}

func TestGiaTisMFFCOrderAndExperiment(t *testing.T) {
	m := NewGiaManager()
	a, _ := m.AppendCI()
	b, _ := m.AppendCI()
	c, _ := m.AppendCI()
	shared, _ := m.AppendAnd(a, b)
	branch, _ := m.AppendAnd(shared, c)
	root, _ := m.AppendAnd(shared, branch)
	m.AppendCO(root)
	before := slices.Clone(m.objects)
	report, err := m.CollectTisMFFC(root.ID(), GiaTisOptions{})
	if err != nil || report.Root != root.ID() || !slices.Equal(report.Nodes, []GiaID{branch.ID(), root.ID()}) || !slices.Equal(report.Leaves, []GiaID{shared.ID(), c.ID()}) {
		t.Fatal(report, err)
	}
	if !slices.Equal(before, m.objects) {
		t.Fatal("collection mutated source")
	}
	// A multiply referenced root remains included by the public collector.
	m.AppendCO(root.Not())
	report, err = m.CollectTisMFFC(root.ID(), GiaTisOptions{})
	if err != nil || report.Nodes[len(report.Nodes)-1] != root.ID() {
		t.Fatal(report, err)
	}

	experimentSource := NewGiaManager()
	c0, _ := experimentSource.AppendCI()
	t0, _ := experimentSource.AppendCI()
	e0, _ := experimentSource.AppendCI()
	mux, _, _ := appendTisMuxPattern(t, experimentSource, c0, t0, e0)
	experimentSource.AppendCO(mux)
	experiment, err := experimentSource.AnalyzeTis(GiaTisOptions{})
	if err != nil || experiment.Graph == nil || experiment.Graph.MuxCount() != 1 || len(experiment.Reports) != 1 {
		t.Fatal(experiment.Graph, experiment.Reports, err)
	}
	got := experiment.Reports[0]
	if got.Root != GiaID(experiment.Graph.ObjectCount()-2) || !slices.Equal(got.Nodes, []GiaID{got.Root}) || len(got.Leaves) != 3 {
		t.Fatal(got)
	}
}

func TestGiaTisLimitsCancellationUnsupportedAndDeep(t *testing.T) {
	m := NewGiaManager()
	a, _ := m.AppendCI()
	b, _ := m.AppendCI()
	root, _ := m.AppendAnd(a, b)
	m.AppendCO(root)
	for _, options := range []GiaTisOptions{{MaxObjects: -1}, {MaxVisits: -1}, {MaxReports: -1}, {MaxObjects: 1}, {MaxVisits: 1}} {
		if graph, err := m.DuplicateWithRecognizedMuxes(options); err == nil || graph != nil {
			t.Fatal(options, graph, err)
		}
	}
	for _, invalid := range []GiaID{0, a.ID(), GiaID(m.ObjectCount())} {
		if report, err := m.CollectTisMFFC(invalid, GiaTisOptions{}); err == nil || report.Root != 0 || report.Nodes != nil || report.Leaves != nil {
			t.Fatal(invalid, report, err)
		}
	}
	if graph, err := (*GiaManager)(nil).DuplicateWithRecognizedMuxes(GiaTisOptions{}); !errors.Is(err, ErrDimension) || graph != nil {
		t.Fatal(graph, err)
	}
	if report, err := (*GiaManager)(nil).CollectTisMFFC(root.ID(), GiaTisOptions{}); !errors.Is(err, ErrDimension) || report.Nodes != nil || report.Leaves != nil {
		t.Fatal(report, err)
	}
	if report, err := m.CollectTisMFFC(root.ID(), GiaTisOptions{MaxVisits: 1}); !errors.Is(err, ErrLimit) || report.Nodes != nil || report.Leaves != nil {
		t.Fatal(report, err)
	}
	calls := 0
	if report, err := m.CollectTisMFFC(root.ID(), GiaTisOptions{Stop: func() bool { calls++; return calls == 2 }}); !errors.Is(err, ErrCancelled) || report.Nodes != nil || report.Leaves != nil {
		t.Fatal(report, err, calls)
	}
	for _, makeUnsupported := range []func(*GiaManager, GiaLit, GiaLit) error{
		func(g *GiaManager, x, y GiaLit) error { _, err := g.AppendXorReal(x, y); return err },
		func(g *GiaManager, x, _ GiaLit) error { _, err := g.AppendBuffer(x); return err },
		func(g *GiaManager, x, y GiaLit) error { _, err := g.AppendMuxReal(x, y, GiaConst0); return err },
	} {
		g := NewGiaManager()
		x, _ := g.AppendCI()
		y, _ := g.AppendCI()
		if err := makeUnsupported(g, x, y); err != nil {
			t.Fatal(err)
		}
		if result, err := g.DuplicateWithRecognizedMuxes(GiaTisOptions{}); !errors.Is(err, ErrUnsupported) || result != nil {
			t.Fatal(result, err)
		}
	}
	for threshold := 1; threshold <= 7; threshold++ {
		calls := 0
		graph, err := m.DuplicateWithRecognizedMuxes(GiaTisOptions{Stop: func() bool { calls++; return calls == threshold }})
		if !errors.Is(err, ErrCancelled) || graph != nil {
			t.Fatal(threshold, calls, graph, err)
		}
	}
	two := NewGiaManager()
	u, _ := two.AppendCI()
	v, _ := two.AppendCI()
	first, _ := two.AppendAnd(u, v)
	second, _ := two.AppendAnd(u.Not(), v)
	two.AppendCO(first)
	two.AppendCO(second)
	if experiment, err := two.AnalyzeTis(GiaTisOptions{MaxReports: 1}); !errors.Is(err, ErrLimit) || experiment.Graph != nil || experiment.Reports != nil {
		t.Fatal(experiment, err)
	}
	calls = 0
	if experiment, err := two.AnalyzeTis(GiaTisOptions{Stop: func() bool { calls++; return calls == 10 }}); !errors.Is(err, ErrCancelled) || experiment.Graph != nil || experiment.Reports != nil {
		t.Fatal(experiment, err, calls)
	}

	deep := NewGiaManager()
	x, _ := deep.AppendCI()
	y, _ := deep.AppendCI()
	chain, _ := deep.AppendAnd(x, y)
	for i := 0; i < 10000; i++ {
		chain, _ = deep.AppendAnd(chain, x.NotIf(i&1 != 0))
	}
	deep.AppendCO(chain)
	report, err := deep.CollectTisMFFC(chain.ID(), GiaTisOptions{})
	if err != nil || len(report.Nodes) != 10001 || !slices.Equal(report.Leaves, []GiaID{x.ID(), y.ID()}) {
		t.Fatal(len(report.Nodes), report.Leaves, err)
	}
}

func tisMFFCModel(m *GiaManager, root GiaID) GiaTisMFFC {
	refs := make([]uint64, len(m.objects))
	for i, object := range m.objects {
		id := GiaID(i)
		if !object.IsAnd() && !object.IsCO() {
			continue
		}
		first, _ := object.Fanin0(id)
		refs[first.ID()]++
		if object.IsAnd() && !object.IsBuffer() {
			second, _ := object.Fanin1(id)
			refs[second.ID()]++
		}
		if control, ok := m.MuxControl(id); ok {
			refs[control.ID()]++
		}
	}
	result := GiaTisMFFC{Root: root}
	seen := make([]bool, len(m.objects))
	var visit func(GiaID)
	visit = func(id GiaID) {
		if seen[id] {
			return
		}
		seen[id] = true
		object := m.objects[id]
		if refs[id] > 1 || object.IsCI() || id == 0 {
			result.Leaves = append(result.Leaves, id)
			return
		}
		first, _ := object.Fanin0(id)
		visit(first.ID())
		if !object.IsBuffer() {
			second, _ := object.Fanin1(id)
			visit(second.ID())
		}
		if control, ok := m.MuxControl(id); ok {
			visit(control.ID())
		}
		result.Nodes = append(result.Nodes, id)
	}
	object := m.objects[root]
	first, _ := object.Fanin0(root)
	visit(first.ID())
	if !object.IsBuffer() {
		second, _ := object.Fanin1(root)
		visit(second.ID())
	}
	if control, ok := m.MuxControl(root); ok {
		visit(control.ID())
	}
	result.Nodes = append(result.Nodes, root)
	return result
}

func FuzzGiaTis(f *testing.F) {
	for _, seed := range []uint64{1, 2, 3, 17, 1001} {
		f.Add(seed)
	}
	f.Fuzz(func(t *testing.T, seed uint64) {
		random := rand.New(rand.NewSource(int64(seed)))
		m := NewGiaManager()
		inputCount := 2 + random.Intn(4)
		literals := make([]GiaLit, inputCount)
		for i := range literals {
			literals[i], _ = m.AppendCI()
		}
		for i := 0; i < 20; i++ {
			left := literals[random.Intn(len(literals))].NotIf(random.Intn(2) != 0)
			right := literals[random.Intn(len(literals))].NotIf(random.Intn(2) != 0)
			for right.ID() == left.ID() {
				right = literals[random.Intn(len(literals))].NotIf(random.Intn(2) != 0)
			}
			gate, err := m.AppendAnd(left, right)
			if err != nil {
				t.Fatal(err)
			}
			literals = append(literals, gate)
		}
		mux, _, _ := appendTisMuxPattern(t, m, literals[0], literals[1], literals[len(literals)-1])
		xor := appendTisXorPattern(t, m, literals[0], literals[1])
		for _, driver := range []GiaLit{mux.NotIf(random.Intn(2) != 0), xor.NotIf(random.Intn(2) != 0), literals[len(literals)-1]} {
			m.AppendCO(driver)
		}
		converted, err := m.DuplicateWithRecognizedMuxes(GiaTisOptions{})
		if err != nil {
			t.Fatal(seed, err)
		}
		for assignment := 0; assignment < 1<<inputCount; assignment++ {
			inputs := make([]bool, inputCount)
			for i := range inputs {
				inputs[i] = assignment&(1<<i) != 0
			}
			want, _ := m.Evaluate(inputs)
			got, gotErr := converted.Evaluate(inputs)
			if gotErr != nil || !slices.Equal(got, want) {
				t.Fatal(seed, assignment, got, want, gotErr)
			}
		}
		root := literals[len(literals)-1].ID()
		got, err := m.CollectTisMFFC(root, GiaTisOptions{})
		want := tisMFFCModel(m, root)
		if err != nil || got.Root != want.Root || !slices.Equal(got.Nodes, want.Nodes) || !slices.Equal(got.Leaves, want.Leaves) {
			t.Fatal(seed, got, want, err)
		}
	})
}
