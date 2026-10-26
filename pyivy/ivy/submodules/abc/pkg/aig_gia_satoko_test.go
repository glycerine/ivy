package abc

// Go-only exhaustive, model, parser, and mapping checks for src/aig/gia/giaSatoko.c.

import (
	"errors"
	"math/rand"
	"slices"
	"testing"
)

func giaSatokoFixture(t testing.TB) *GiaManager {
	t.Helper()
	manager := NewGiaManager()
	a, _ := manager.AppendCI()
	b, _ := manager.AppendCI()
	c, _ := manager.AppendCI()
	and, _ := manager.AppendAnd(a, b)
	xor, _ := manager.AppendXorReal(a, c)
	mux, _ := manager.AppendMuxReal(b, xor, and.Not())
	for _, signal := range []GiaLit{and, GiaConst0, GiaConst1, mux} {
		if _, err := manager.AppendCO(signal); err != nil {
			t.Fatal(err)
		}
	}
	return manager
}

func giaSatokoExpected(manager *GiaManager, output int) bool {
	for assignment := 0; assignment < 1<<manager.InputCount(); assignment++ {
		inputs := make([]bool, manager.InputCount())
		for index := range inputs {
			inputs[index] = assignment&(1<<index) != 0
		}
		values, err := manager.Evaluate(inputs)
		if err != nil {
			return false
		}
		if output >= 0 && values[output] {
			return true
		}
		if output < 0 && slices.Contains(values, true) {
			return true
		}
	}
	return false
}

func checkGiaSatokoModel(t testing.TB, manager *GiaManager, report GiaSatokoReport) {
	t.Helper()
	if report.Status != SatSatisfiable {
		if report.Model != nil {
			t.Fatal("non-SAT result retained a model")
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
	} else if !slices.Contains(values, true) {
		t.Fatal(report.Model, values)
	}
}

func TestGiaSatokoModesStatusesModelsAndOwnership(t *testing.T) {
	manager := giaSatokoFixture(t)
	before := slices.Clone(manager.objects)
	for _, mode := range []GiaSatokoMode{GiaSatokoAggregate, GiaSatokoSplit, GiaSatokoIncremental} {
		result, err := manager.SolveGiaSatoko(mode, GiaSatokoOptions{})
		if err != nil || result.Mode != mode {
			t.Fatal(mode, result, err)
		}
		if mode == GiaSatokoAggregate {
			if len(result.Reports) != 1 || result.Reports[0].Output != -1 || result.Reports[0].Status != SatSatisfiable {
				t.Fatal(result)
			}
			checkGiaSatokoModel(t, manager, result.Reports[0])
			continue
		}
		if len(result.Reports) != manager.OutputCount() {
			t.Fatal(len(result.Reports))
		}
		for output, report := range result.Reports {
			want := SatUnsatisfiable
			if giaSatokoExpected(manager, output) {
				want = SatSatisfiable
			}
			if report.Output != output || report.Status != want || report.Variables < 1 || report.Clauses < 1 || report.Literals < 1 {
				t.Fatal(mode, output, report, want)
			}
			if mode == GiaSatokoIncremental && len(report.ConeVariables) == 0 {
				t.Fatal("missing incremental cone", output)
			}
			checkGiaSatokoModel(t, manager, report)
		}
	}
	if !slices.Equal(before, manager.objects) {
		t.Fatal("source graph mutated")
	}
}

func TestGiaSatokoSelectedSessionLifecycleAndReports(t *testing.T) {
	manager := giaSatokoFixture(t)
	for _, test := range []struct {
		outputs []int
		status  SatStatus
	}{{[]int{1}, SatUnsatisfiable}, {[]int{0}, SatSatisfiable}, {[]int{0, 1}, SatSatisfiable}, {[]int{2}, SatSatisfiable}} {
		session, err := NewGiaSatokoSession(manager, test.outputs, GiaSatokoOptions{})
		if err != nil {
			t.Fatal(test, err)
		}
		report, err := session.Solve()
		if err != nil || report.Status != test.status || report.Conflicts != report.Statistics.Conflicts {
			t.Fatal(test, report, err)
		}
		session.Close()
		session.Close()
		if _, err = session.Solve(); !errors.Is(err, ErrDimension) {
			t.Fatal(err)
		}
	}
}

func TestCollectCNFMappingVariablesPostorderErrorsAndDepth(t *testing.T) {
	// Variables 0 and 1 are leaves; 2=(0,1), 3=(2,1), 4=(2,3).
	mapping := []int{0, 0, 5, 8, 11, 2, 0, 1, 2, 2, 1, 2, 2, 3}
	got, err := CollectCNFMappingVariables(4, mapping, 5, 0, nil)
	if err != nil || !slices.Equal(got, []SatVar{0, 1, 2, 3, 4}) {
		t.Fatal(got, err)
	}
	leaf, err := CollectCNFMappingVariables(1, mapping, 5, 0, nil)
	if err != nil || !slices.Equal(leaf, []SatVar{1}) {
		t.Fatal(leaf, err)
	}
	for _, malformed := range []struct {
		mapping []int
		vars    int
	}{{[]int{2}, 1}, {[]int{-1}, 1}, {[]int{1, -1}, 1}, {[]int{1, 1, 4}, 1}} {
		if _, err = CollectCNFMappingVariables(0, malformed.mapping, malformed.vars, 0, nil); err == nil {
			t.Fatal(malformed)
		}
	}
	cycle := []int{2, 4, 1, 1, 1, 0}
	if _, err = CollectCNFMappingVariables(0, cycle, 2, 0, nil); !errors.Is(err, ErrInvariant) {
		t.Fatal(err)
	}
	if _, err = CollectCNFMappingVariables(4, mapping, 5, 3, nil); !errors.Is(err, ErrLimit) {
		t.Fatal(err)
	}
	if _, err = CollectCNFMappingVariables(4, mapping, 5, 0, func() bool { return true }); !errors.Is(err, ErrCancelled) {
		t.Fatal(err)
	}

	const count = 10001
	deep := make([]int, count)
	for variable := 1; variable < count; variable++ {
		deep[variable] = len(deep)
		deep = append(deep, 1, variable-1)
	}
	got, err = CollectCNFMappingVariables(count-1, deep, count, count, nil)
	if err != nil || len(got) != count || got[0] != 0 || got[count-1] != count-1 {
		t.Fatal(len(got), err)
	}
}

func TestGiaSatokoDIMACSStandalone(t *testing.T) {
	for _, test := range []struct {
		text   string
		status SatStatus
	}{{"c comment\np cnf 3 3\n1 -2 0\n+2 3 0\n-3 0\n", SatSatisfiable}, {"p cnf 1 2\n1 0\n-1 0\n", SatUnsatisfiable}, {"p cnf 0 1\n0\n", SatUnsatisfiable}, {"p cnf 0 0\n", SatSatisfiable}} {
		session, err := ParseGiaSatokoDIMACS([]byte(test.text), GiaSatokoOptions{})
		if err != nil {
			t.Fatal(test.text, err)
		}
		report, err := session.Solve()
		session.Close()
		if err != nil || report.Status != test.status {
			t.Fatal(test.text, report, err)
		}
		if report.Status == SatSatisfiable && len(report.Model) != report.Variables {
			t.Fatal(report)
		}
	}
	for _, text := range []string{"p cnf 1 2\n1 0\n", "p cnf 1 1\n2 0\n", "p cnf -1 0\n", "p nope 1 1\n1 0\n", "p cnf 1 1\n1\n", "x 0\n", "p cnf 1 0\np cnf 1 0\n"} {
		if session, err := ParseGiaSatokoDIMACS([]byte(text), GiaSatokoOptions{}); err == nil || session != nil {
			t.Fatal(text, session, err)
		}
	}
}

func TestGiaSatokoErrorsLimitsAndCancellation(t *testing.T) {
	manager := giaSatokoFixture(t)
	for _, selected := range [][]int{{}, {-1}, {manager.OutputCount()}, {0, 0}} {
		if session, err := NewGiaSatokoSession(manager, selected, GiaSatokoOptions{}); err == nil || session != nil {
			t.Fatal(selected, session, err)
		}
	}
	for _, options := range []GiaSatokoOptions{{MaxObjects: -1}, {MaxMappingEntries: -1}, {MaxVariables: -1}, {MaxClauses: -1}, {MaxLiterals: -1}, {MaxObjects: 1}, {MaxMappingEntries: 1}, {MaxVariables: 1}, {MaxClauses: 1}, {MaxLiterals: 1}} {
		if result, err := manager.SolveGiaSatoko(GiaSatokoAggregate, options); err == nil || result.Reports != nil {
			t.Fatal(options, result, err)
		}
	}
	if _, err := (*GiaManager)(nil).SolveGiaSatoko(GiaSatokoAggregate, GiaSatokoOptions{}); !errors.Is(err, ErrDimension) {
		t.Fatal(err)
	}
	if _, err := manager.SolveGiaSatoko(GiaSatokoMode(99), GiaSatokoOptions{}); !errors.Is(err, ErrDimension) {
		t.Fatal(err)
	}
	for threshold := 1; threshold <= 12; threshold++ {
		calls := 0
		_, err := manager.SolveGiaSatoko(GiaSatokoIncremental, GiaSatokoOptions{Stop: func() bool { calls++; return calls >= threshold }})
		if !errors.Is(err, ErrCancelled) {
			t.Fatal(threshold, calls, err)
		}
	}
	session, err := NewGiaSatokoSession(manager, nil, GiaSatokoOptions{SAT: SatLimits{Stop: func() bool { return true }}})
	if err != nil {
		t.Fatal(err)
	}
	report, err := session.Solve()
	session.Close()
	if err != nil || report.Status != SatUnknown || report.Reason != SatCancelled {
		t.Fatal(report, err)
	}
}

func FuzzGiaSatokoModes(f *testing.F) {
	f.Add(uint64(1))
	f.Add(uint64(0x9e3779b97f4a7c15))
	f.Fuzz(func(t *testing.T, seed uint64) {
		random := rand.New(rand.NewSource(int64(seed)))
		manager := NewGiaManager()
		signals := []GiaLit{GiaConst0, GiaConst1}
		for index := 0; index < 1+random.Intn(4); index++ {
			literal, _ := manager.AppendCI()
			signals = append(signals, literal)
		}
		for index := 0; index < random.Intn(8); index++ {
			left := signals[random.Intn(len(signals))].NotIf(random.Intn(2) != 0)
			right := signals[random.Intn(len(signals))].NotIf(random.Intn(2) != 0)
			literal, err := manager.AppendAnd(left, right)
			if err == nil {
				signals = append(signals, literal)
			}
		}
		for index := 0; index < 1+random.Intn(4); index++ {
			manager.AppendCO(signals[random.Intn(len(signals))].NotIf(random.Intn(2) != 0))
		}
		for _, mode := range []GiaSatokoMode{GiaSatokoAggregate, GiaSatokoSplit, GiaSatokoIncremental} {
			result, err := manager.SolveGiaSatoko(mode, GiaSatokoOptions{MaxObjects: 4096, MaxMappingEntries: 16384, MaxVariables: 4096, MaxClauses: 16384, MaxLiterals: 65536})
			if err != nil {
				t.Fatal(seed, mode, err)
			}
			for _, report := range result.Reports {
				want := SatUnsatisfiable
				if giaSatokoExpected(manager, report.Output) {
					want = SatSatisfiable
				}
				if report.Status != want {
					t.Fatal(seed, mode, report.Output, report.Status, want)
				}
				checkGiaSatokoModel(t, manager, report)
			}
		}
	})
}
