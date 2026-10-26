package abc

// Go-only exhaustive SAT and model checks for src/aig/gia/giaSat3.c.

import (
	"errors"
	"math/rand"
	"slices"
	"testing"
)

func giaSat3Fixture(t testing.TB) *GiaManager {
	t.Helper()
	m := NewGiaManager()
	m.Name, m.Spec = "sat3", "sat3.aig"
	a, _ := m.AppendCI()
	b, _ := m.AppendCI()
	c, _ := m.AppendCI()
	and, _ := m.AppendAnd(a, b)
	xor, _ := m.AppendXorReal(a, c)
	mux, _ := m.AppendMuxReal(b, xor, and.Not())
	for _, output := range []GiaLit{and, GiaConst0, GiaConst1, mux} {
		if _, err := m.AppendCO(output); err != nil {
			t.Fatal(err)
		}
	}
	return m
}

func giaSat3Expected(manager *GiaManager, output int) bool {
	for assignment := 0; assignment < 1<<manager.InputCount(); assignment++ {
		inputs := make([]bool, manager.InputCount())
		for i := range inputs {
			inputs[i] = assignment&(1<<i) != 0
		}
		values, err := manager.Evaluate(inputs)
		if err != nil {
			return false
		}
		if output < 0 {
			for _, value := range values {
				if value {
					return true
				}
			}
		} else if values[output] {
			return true
		}
	}
	return false
}

func giaCheckSat3Model(t testing.TB, manager *GiaManager, report GiaSat3Report) {
	t.Helper()
	if report.Status != SatSatisfiable {
		if report.Model != nil {
			t.Fatal(report)
		}
		return
	}
	if len(report.Model) != manager.InputCount() {
		t.Fatal(len(report.Model), manager.InputCount())
	}
	values, err := manager.Evaluate(report.Model)
	if err != nil {
		t.Fatal(err)
	}
	if report.Output >= 0 {
		if !values[report.Output] {
			t.Fatal(report.Output, report.Model, values)
		}
		return
	}
	for _, value := range values {
		if value {
			return
		}
	}
	t.Fatal(report.Model, values)
}

func TestGiaSat3WholeSplitModelsAndOwnership(t *testing.T) {
	m := giaSat3Fixture(t)
	before := slices.Clone(m.objects)
	whole, err := m.SolveGiaSat3(false, GiaSat3Options{})
	if err != nil || whole.Split || len(whole.Reports) != 1 || whole.Reports[0].Output != -1 || whole.Reports[0].Status != SatSatisfiable {
		t.Fatal(whole, err)
	}
	giaCheckSat3Model(t, m, whole.Reports[0])
	if !slices.Equal(before, m.objects) || m.Name != "sat3" || m.Spec != "sat3.aig" {
		t.Fatal("source mutated")
	}

	split, err := m.SolveGiaSat3(true, GiaSat3Options{})
	if err != nil || !split.Split || len(split.Reports) != m.OutputCount() {
		t.Fatal(split, err)
	}
	for output, report := range split.Reports {
		want := SatUnsatisfiable
		if giaSat3Expected(m, output) {
			want = SatSatisfiable
		}
		if report.Output != output || report.Status != want || report.Objects > m.ObjectCount()+8 || report.Variables < 1 || report.Clauses < 1 || report.Literals < 1 {
			t.Fatal(output, report, want)
		}
		giaCheckSat3Model(t, m, report)
	}
}

func TestGiaSat3SelectedOutputsSessionAndLifecycle(t *testing.T) {
	m := giaSat3Fixture(t)
	for _, test := range []struct {
		outputs []int
		status  SatStatus
	}{
		{[]int{1}, SatUnsatisfiable},
		{[]int{0}, SatSatisfiable},
		{[]int{0, 1}, SatSatisfiable},
		{[]int{2}, SatSatisfiable},
	} {
		session, err := NewGiaSat3Session(m, test.outputs, GiaSat3Options{})
		if err != nil {
			t.Fatal(test, err)
		}
		report, err := session.Solve()
		if err != nil || report.Status != test.status {
			t.Fatal(test, report, err)
		}
		if report.Status == SatSatisfiable {
			values, evalErr := m.Evaluate(report.Model)
			if evalErr != nil {
				t.Fatal(evalErr)
			}
			found := false
			for _, output := range test.outputs {
				found = found || values[output]
			}
			if !found {
				t.Fatal(test, report.Model, values)
			}
		}
		session.Close()
		session.Close()
		if report, solveErr := session.Solve(); !errors.Is(solveErr, ErrDimension) || report.Status != SatUnknown {
			t.Fatal(report, solveErr)
		}
	}
}

func TestGiaSat3ErrorsLimitsCancellationAndDeep(t *testing.T) {
	m := giaSat3Fixture(t)
	for _, selected := range [][]int{{}, {-1}, {m.OutputCount()}, {0, 0}} {
		if session, err := NewGiaSat3Session(m, selected, GiaSat3Options{}); err == nil || session != nil {
			t.Fatal(selected, session, err)
		}
	}
	for _, options := range []GiaSat3Options{
		{MaxObjects: -1}, {MaxMappingEntries: -1}, {MaxVariables: -1}, {MaxClauses: -1}, {MaxLiterals: -1},
		{MaxObjects: 1}, {MaxMappingEntries: 1}, {MaxVariables: 1}, {MaxClauses: 1}, {MaxLiterals: 1},
	} {
		if result, err := m.SolveGiaSat3(false, options); err == nil || result.Reports != nil {
			t.Fatal(options, result, err)
		}
	}
	if result, err := (*GiaManager)(nil).SolveGiaSat3(false, GiaSat3Options{}); !errors.Is(err, ErrDimension) || result.Reports != nil {
		t.Fatal(result, err)
	}
	empty := NewGiaManager()
	if result, err := empty.SolveGiaSat3(false, GiaSat3Options{}); !errors.Is(err, ErrDimension) || result.Reports != nil {
		t.Fatal(result, err)
	}
	for threshold := 1; threshold <= 12; threshold++ {
		calls := 0
		result, err := m.SolveGiaSat3(true, GiaSat3Options{Stop: func() bool { calls++; return calls >= threshold }})
		if !errors.Is(err, ErrCancelled) || result.Reports != nil {
			t.Fatal(threshold, calls, result, err)
		}
	}
	session, err := NewGiaSat3Session(m, nil, GiaSat3Options{SAT: SatSolver3Limits{Stop: func() bool { return true }}})
	if err != nil {
		t.Fatal(err)
	}
	report, err := session.Solve()
	session.Close()
	if err != nil || report.Status != SatUnknown || report.Reason != SatCancelled {
		t.Fatal(report, err)
	}

	deep := NewGiaManager()
	a, _ := deep.AppendCI()
	b, _ := deep.AppendCI()
	chain, _ := deep.AppendAnd(a, b)
	for i := 0; i < 10000; i++ {
		chain, _ = deep.AppendAnd(chain, a)
	}
	deep.AppendCO(chain)
	result, err := deep.SolveGiaSat3(false, GiaSat3Options{})
	if err != nil || len(result.Reports) != 1 || result.Reports[0].Status != SatSatisfiable {
		t.Fatal(result, err)
	}
	giaCheckSat3Model(t, deep, result.Reports[0])
}

func FuzzGiaSat3StatusAndModels(f *testing.F) {
	f.Add(uint64(1))
	f.Add(uint64(0x9e3779b97f4a7c15))
	f.Fuzz(func(t *testing.T, seed uint64) {
		random := rand.New(rand.NewSource(int64(seed)))
		m := NewGiaManager()
		inputs := 1 + random.Intn(4)
		signals := []GiaLit{GiaConst0, GiaConst1}
		for i := 0; i < inputs; i++ {
			literal, _ := m.AppendCI()
			signals = append(signals, literal)
		}
		for i := 0; i < random.Intn(8); i++ {
			first := signals[2+random.Intn(len(signals)-2)]
			second := signals[2+random.Intn(len(signals)-2)]
			if first.ID() == second.ID() {
				continue
			}
			gate, err := m.AppendAnd(first.NotIf(random.Intn(2) != 0), second.NotIf(random.Intn(2) != 0))
			if err == nil {
				signals = append(signals, gate)
			}
		}
		outputs := 1 + random.Intn(4)
		for i := 0; i < outputs; i++ {
			driver := signals[random.Intn(len(signals))].NotIf(random.Intn(2) != 0)
			m.AppendCO(driver)
		}
		split, err := m.SolveGiaSat3(true, GiaSat3Options{MaxObjects: 4096, MaxMappingEntries: 4096, MaxVariables: 4096, MaxClauses: 16384, MaxLiterals: 65536})
		if err != nil {
			t.Fatal(err)
		}
		for output, report := range split.Reports {
			want := SatUnsatisfiable
			if giaSat3Expected(m, output) {
				want = SatSatisfiable
			}
			if report.Status != want {
				t.Fatal(seed, output, report.Status, want)
			}
			giaCheckSat3Model(t, m, report)
		}
		whole, err := m.SolveGiaSat3(false, GiaSat3Options{MaxObjects: 4096, MaxMappingEntries: 4096, MaxVariables: 4096, MaxClauses: 16384, MaxLiterals: 65536})
		if err != nil {
			t.Fatal(err)
		}
		want := SatUnsatisfiable
		if giaSat3Expected(m, -1) {
			want = SatSatisfiable
		}
		if whole.Reports[0].Status != want {
			t.Fatal(seed, whole.Reports[0].Status, want)
		}
		giaCheckSat3Model(t, m, whole.Reports[0])
	})
}
