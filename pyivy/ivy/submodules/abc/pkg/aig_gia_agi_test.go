package abc

// Go-only representation, support, and truth checks for src/aig/gia/giaAgi.c.

import (
	"errors"
	"math/rand"
	"slices"
	"testing"
)

func giaAgiFixture(t testing.TB) *GiaManager {
	t.Helper()
	m := NewGiaManager()
	m.Name, m.Spec = "agi", "agi.aig"
	a, _ := m.AppendCI()
	b, _ := m.AppendCI()
	c, _ := m.AppendCI()
	r, _ := m.AppendCI()
	x, _ := m.AppendAnd(a, b)
	y, _ := m.AppendAnd(x, c.Not())
	z, _ := m.AppendAnd(a.Not(), r)
	w, _ := m.AppendAnd(y, z)
	for _, driver := range []GiaLit{w, y.Not(), x} {
		if _, err := m.AppendCO(driver); err != nil {
			t.Fatal(err)
		}
	}
	if err := m.SetRegisterCount(1); err != nil {
		t.Fatal(err)
	}
	return m
}

func TestGiaAgiConversionKindsTruthAndOwnership(t *testing.T) {
	m := giaAgiFixture(t)
	before := slices.Clone(m.objects)
	agi, mapping, err := AgiFromGia(m, 0)
	if err != nil {
		t.Fatal(err)
	}
	if agi.Name != m.Name || agi.Spec != m.Spec || agi.ObjectCount() != m.ObjectCount() || agi.InputCount() != m.InputCount() || agi.OutputCount() != m.OutputCount() || agi.NodeCount() != m.AndCount() || agi.RegisterCount() != 1 || len(mapping) != m.ObjectCount() || !slices.Equal(before, m.objects) {
		t.Fatal(agi, mapping)
	}
	for id := 0; id < m.ObjectCount(); id++ {
		if mapping[id] != AgiLit(id)<<1 {
			t.Fatal(id, mapping[id])
		}
	}
	inputs := agi.Inputs()
	outputs := agi.Outputs()
	if !agi.IsPI(inputs[0]) || !agi.IsPI(inputs[2]) || agi.IsPI(inputs[3]) || !agi.IsRO(inputs[3]) || !agi.IsPO(outputs[0]) || agi.IsPO(outputs[2]) || !agi.IsRI(outputs[2]) || !agi.IsConstant(0) {
		t.Fatal(inputs, outputs)
	}
	for assignment := 0; assignment < 1<<m.InputCount(); assignment++ {
		values := make([]bool, m.InputCount())
		for i := range values {
			values[i] = assignment&(1<<i) != 0
		}
		want, sourceErr := m.Evaluate(values)
		got, targetErr := agi.Evaluate(values)
		if sourceErr != nil || targetErr != nil || !slices.Equal(got, want) {
			t.Fatal(assignment, got, want, sourceErr, targetErr)
		}
	}
	copyInputs := agi.Inputs()
	copyInputs[0] = 0
	if agi.Inputs()[0] == 0 {
		t.Fatal("borrowed inputs")
	}
}

func TestGiaAgiSupportComparisonAndExactSizes(t *testing.T) {
	m := giaAgiFixture(t)
	comparison, err := CompareGiaAgiSupports(m, 2, AgiSupportOptions{})
	if err != nil || !comparison.Equivalent || comparison.AgiReport.Small != comparison.GiaSmall || !slices.Equal(comparison.AgiReport.Sizes, comparison.GiaSizes) {
		t.Fatal(comparison, err)
	}
	wantNodes := []AgiID{5, 6, 7, 8}
	wantSizes := []int{2, 3, 2, 4}
	if !slices.Equal(comparison.AgiReport.Nodes, wantNodes) || !slices.Equal(comparison.AgiReport.Sizes, wantSizes) || comparison.AgiReport.Small != 2 {
		t.Fatal(comparison.AgiReport)
	}
	for i, id := range wantNodes {
		size, sizeErr := comparison.Agi.SupportSize(id, AgiSupportOptions{})
		if sizeErr != nil || size != wantSizes[i] {
			t.Fatal(id, size, sizeErr)
		}
	}
	if size, err := comparison.Agi.SupportSize(0, AgiSupportOptions{}); err != nil || size != 0 {
		t.Fatal(size, err)
	}
	for i, id := range comparison.Agi.Outputs() {
		size, sizeErr := comparison.Agi.SupportSize(id, AgiSupportOptions{})
		if sizeErr != nil || size != []int{4, 3, 2}[i] {
			t.Fatal(i, size, sizeErr)
		}
	}
}

func TestGiaAgiConstructionErrorsLimitsCancellationAndDeep(t *testing.T) {
	if manager, err := NewAgiManager(0); !errors.Is(err, ErrDimension) || manager != nil {
		t.Fatal(manager, err)
	}
	manager, err := NewAgiManager(5)
	if err != nil {
		t.Fatal(err)
	}
	a, _ := manager.AppendCI()
	b, _ := manager.AppendCI()
	if _, err = manager.AppendAND(a, a.Not()); !errors.Is(err, ErrInvalidLiteral) {
		t.Fatal(err)
	}
	gate, err := manager.AppendAND(a, b)
	if err != nil {
		t.Fatal(err)
	}
	manager.AppendCO(gate)
	if _, err = manager.AppendCI(); !errors.Is(err, ErrLimit) {
		t.Fatal(err)
	}
	if err = manager.SetRegisterCount(2); !errors.Is(err, ErrDimension) {
		t.Fatal(err)
	}
	if _, err = manager.SupportSize(AgiID(manager.ObjectCount()), AgiSupportOptions{}); !errors.Is(err, ErrInvalidLiteral) {
		t.Fatal(err)
	}
	for _, options := range []AgiSupportOptions{{MaxObjects: -1}, {MaxVisits: -1}, {MaxObjects: 1}, {MaxVisits: 1}} {
		if report, reportErr := manager.AnalyzeSupports(16, options); reportErr == nil || report.Nodes != nil {
			t.Fatal(options, report, reportErr)
		}
	}
	for threshold := 1; threshold <= 3; threshold++ {
		calls := 0
		report, reportErr := manager.AnalyzeSupports(16, AgiSupportOptions{Stop: func() bool { calls++; return calls >= threshold }})
		if !errors.Is(reportErr, ErrCancelled) || report.Nodes != nil {
			t.Fatal(threshold, calls, report, reportErr)
		}
	}
	for _, extended := range []func(*GiaManager, GiaLit, GiaLit) error{
		func(g *GiaManager, x, y GiaLit) error { _, e := g.AppendBuffer(x); return e },
		func(g *GiaManager, x, y GiaLit) error { _, e := g.AppendXorReal(x, y); return e },
		func(g *GiaManager, x, y GiaLit) error { _, e := g.AppendMuxReal(x, y, GiaConst0); return e },
	} {
		g := NewGiaManager()
		x, _ := g.AppendCI()
		y, _ := g.AppendCI()
		if err := extended(g, x, y); err != nil {
			t.Fatal(err)
		}
		if converted, mapping, convertErr := AgiFromGia(g, 0); !errors.Is(convertErr, ErrUnsupported) || converted != nil || mapping != nil {
			t.Fatal(converted, mapping, convertErr)
		}
	}

	deep := NewGiaManager()
	x, _ := deep.AppendCI()
	y, _ := deep.AppendCI()
	root, _ := deep.AppendAnd(x, y)
	for i := 0; i < 10000; i++ {
		root, _ = deep.AppendAnd(root, x.NotIf(i&1 != 0))
	}
	deep.AppendCO(root)
	converted, _, err := AgiFromGia(deep, 0)
	if err != nil {
		t.Fatal(err)
	}
	size, err := converted.SupportSize(AgiID(root.ID()), AgiSupportOptions{})
	if err != nil || size != 2 {
		t.Fatal(size, err)
	}
	manager.Close()
	manager.Close()
	if err := manager.Check(); !errors.Is(err, ErrInvariant) {
		t.Fatal(err)
	}
}

func FuzzGiaAgiConversionAndSupports(f *testing.F) {
	f.Add(uint64(1), uint8(5))
	f.Add(uint64(0x9e3779b97f4a7c15), uint8(11))
	f.Fuzz(func(t *testing.T, seed uint64, assignment uint8) {
		random := rand.New(rand.NewSource(int64(seed)))
		m := NewGiaManager()
		inputs := 1 + random.Intn(4)
		signals := make([]GiaLit, inputs)
		for i := range signals {
			signals[i], _ = m.AppendCI()
		}
		for i := 0; i < 1+random.Intn(10); i++ {
			first := signals[random.Intn(len(signals))]
			second := signals[random.Intn(len(signals))]
			if first.ID() == second.ID() {
				continue
			}
			gate, err := m.AppendAnd(first.NotIf(random.Intn(2) != 0), second.NotIf(random.Intn(2) != 0))
			if err == nil {
				signals = append(signals, gate)
			}
		}
		for i := 0; i < 1+random.Intn(3); i++ {
			m.AppendCO(signals[random.Intn(len(signals))].NotIf(random.Intn(2) != 0))
		}
		comparison, err := CompareGiaAgiSupports(m, int(seed%5), AgiSupportOptions{MaxObjects: 4096, MaxVisits: 65536})
		if err != nil || !comparison.Equivalent || comparison.AgiReport.Small != comparison.GiaSmall {
			t.Fatal(seed, comparison, err)
		}
		values := make([]bool, inputs)
		for i := range values {
			values[i] = assignment&(1<<i) != 0
		}
		want, sourceErr := m.Evaluate(values)
		got, targetErr := comparison.Agi.Evaluate(values)
		if sourceErr != nil || targetErr != nil || !slices.Equal(got, want) {
			t.Fatal(seed, assignment, got, want, sourceErr, targetErr)
		}
	})
}
