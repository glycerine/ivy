package abc

// Standalone tests for the port of src/sat/csat/csat_apis.c.

import (
	"errors"
	"os"
	"path/filepath"
	"slices"
	"strings"
	"testing"
)

func newCSATTestManager(t *testing.T) *CSATManager {
	t.Helper()
	m, err := NewCSATManager(CSATLimits{})
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = m.Close() })
	return m
}

func addCSATTestGate(t *testing.T, m *CSATManager, kind CSATGateType, name string, fanins ...string) {
	t.Helper()
	if err := m.AddGate(kind, name, fanins, 0); err != nil {
		t.Fatalf("AddGate(%v, %q, %v): %v", kind, name, fanins, err)
	}
}

func solveCSATTestTarget(t *testing.T, m *CSATManager, names []string, values []int) CSATStatus {
	t.Helper()
	if id, err := m.AddTarget(names, values); err != nil || id != 1 {
		t.Fatalf("AddTarget: id=%d err=%v", id, err)
	}
	if err := m.SolveInit(); err != nil {
		t.Fatal(err)
	}
	status, err := m.Solve()
	if err != nil {
		t.Fatal(err)
	}
	return status
}

func csatExpected(kind CSATGateType, inputs []bool) bool {
	switch kind {
	case CSATConst:
		return true
	case CSATBAND, CSATBNAND:
		value := true
		for _, input := range inputs {
			value = value && input
		}
		return value != (kind == CSATBNAND)
	case CSATBOR, CSATBNOR:
		value := false
		for _, input := range inputs {
			value = value || input
		}
		return value != (kind == CSATBNOR)
	case CSATBXOR, CSATBXNOR:
		value := inputs[0]
		if len(inputs) == 2 {
			value = value != inputs[1]
		}
		return value != (kind == CSATBXNOR)
	case CSATBINV:
		return !inputs[0]
	case CSATBBUF, CSATBPPO, CSATBPO:
		return inputs[0]
	default:
		panic("unsupported test gate")
	}
}

func TestCSATGateTruthTablesExhaustive(t *testing.T) {
	cases := []struct {
		kind   CSATGateType
		fanins int
	}{
		{CSATConst, 0}, {CSATBAND, 1}, {CSATBAND, 3}, {CSATBNAND, 3},
		{CSATBOR, 1}, {CSATBOR, 3}, {CSATBNOR, 3}, {CSATBXOR, 1},
		{CSATBXOR, 2}, {CSATBXNOR, 1}, {CSATBXNOR, 2}, {CSATBINV, 1},
		{CSATBBUF, 1}, {CSATBPPO, 1}, {CSATBPO, 1},
	}
	for _, test := range cases {
		for assignment := 0; assignment < 1<<test.fanins; assignment++ {
			m := newCSATTestManager(t)
			names := make([]string, test.fanins)
			values := make([]int, test.fanins)
			bools := make([]bool, test.fanins)
			for i := range names {
				names[i] = string(rune('a' + i))
				values[i] = assignment >> i & 1
				bools[i] = values[i] != 0
				addCSATTestGate(t, m, CSATBPI, names[i])
			}
			addCSATTestGate(t, m, test.kind, "z", names...)
			want := csatExpected(test.kind, bools)
			targetNames := append(slices.Clone(names), "z")
			targetValues := append(slices.Clone(values), boolToIntCSAT(want))
			if got := solveCSATTestTarget(t, m, targetNames, targetValues); got != CSATSatisfiable {
				t.Fatalf("gate=%v assignment=%d matching status=%v", test.kind, assignment, got)
			}
			targetValues[len(targetValues)-1] ^= 1
			if got := solveCSATTestTarget(t, m, targetNames, targetValues); got != CSATUnsatisfiable {
				t.Fatalf("gate=%v assignment=%d opposite status=%v", test.kind, assignment, got)
			}
		}
	}
}

func boolToIntCSAT(value bool) int {
	if value {
		return 1
	}
	return 0
}

func TestCSATValidationAtomicityAndLimits(t *testing.T) {
	if manager, err := NewCSATManager(CSATLimits{MaxGates: -1}); manager != nil || !errors.Is(err, ErrDimension) {
		t.Fatal(manager, err)
	}
	m, err := NewCSATManager(CSATLimits{MaxGates: 3, MaxFanins: 1, MaxTargets: 2, MaxNameBytes: 3})
	if err != nil {
		t.Fatal(err)
	}
	defer m.Close()
	addCSATTestGate(t, m, CSATBPI, "a")
	before := m.GateCount()
	for _, trial := range []struct {
		kind    CSATGateType
		name    string
		fanins  []string
		wantErr error
	}{
		{CSATBMUX, "x", []string{"a"}, ErrUnsupported},
		{CSATBBUF, "a", []string{"a"}, ErrInvariant},
		{CSATBBUF, "x", []string{"missing"}, ErrInvalidLiteral},
		{CSATBAND, "x", nil, ErrDimension},
		{CSATBBUF, "bad\n", []string{"a"}, ErrDimension},
	} {
		if e := m.AddGate(trial.kind, trial.name, trial.fanins, 0); !errors.Is(e, trial.wantErr) {
			t.Fatalf("%q: got %v want %v", trial.name, e, trial.wantErr)
		}
		if m.GateCount() != before {
			t.Fatal("failed AddGate mutated manager")
		}
	}
	addCSATTestGate(t, m, CSATBBUF, "b", "a")
	if err = m.AddGate(CSATBBUF, "c", []string{"b"}, 0); !errors.Is(err, ErrLimit) {
		t.Fatalf("cumulative fanin limit: %v", err)
	}
	if id, e := m.AddTarget([]string{"a"}, []int{1}); e != nil || id != 1 {
		t.Fatal(id, e)
	}
	old := slices.Clone(m.target.gates)
	if _, e := m.AddTarget([]string{"a", "b", "a"}, []int{1, 1, 1}); !errors.Is(e, ErrLimit) {
		t.Fatal(e)
	}
	if _, e := m.AddTarget([]string{"a", "missing"}, []int{1, 0}); !errors.Is(e, ErrInvalidLiteral) {
		t.Fatal(e)
	}
	if _, e := m.AddTarget([]string{"a", "b"}, []int{1, 2}); !errors.Is(e, ErrDimension) {
		t.Fatal(e)
	}
	if _, e := m.AddTarget([]string{"a"}, nil); !errors.Is(e, ErrDimension) {
		t.Fatal(e)
	}
	if !slices.Equal(old, m.target.gates) {
		t.Fatal("failed AddTarget mutated prior target")
	}
}

func TestCSATGateInventoryAritiesAndUnsupportedKinds(t *testing.T) {
	allKinds := [...]CSATGateType{
		CSATConst, CSATBPI, CSATBPPI, CSATBAND, CSATBNAND, CSATBOR, CSATBNOR,
		CSATBXOR, CSATBXNOR, CSATBINV, CSATBBUF, CSATBMUX, CSATBDFF, CSATBSDFF,
		CSATBTRIH, CSATBTRIL, CSATBBUS, CSATBPPO, CSATBPO, CSATBCNF, CSATBDC,
	}
	for value, kind := range allKinds {
		if int(kind) != value {
			t.Fatalf("gate inventory value %d is %d", value, kind)
		}
	}

	for _, kind := range []CSATGateType{
		CSATBMUX, CSATBDFF, CSATBSDFF, CSATBTRIH, CSATBTRIL, CSATBBUS, CSATBCNF, CSATBDC,
	} {
		m := newCSATTestManager(t)
		if err := m.AddGate(kind, "unsupported", nil, 0); !errors.Is(err, ErrUnsupported) {
			t.Fatalf("gate %v: got %v, want ErrUnsupported", kind, err)
		}
		if m.GateCount() != 0 {
			t.Fatalf("gate %v mutated manager", kind)
		}
	}

	m := newCSATTestManager(t)
	addCSATTestGate(t, m, CSATBPI, "a")
	addCSATTestGate(t, m, CSATBPI, "b")
	invalid := []struct {
		kind   CSATGateType
		fanins []string
	}{
		{CSATConst, []string{"a"}},
		{CSATBPI, []string{"a"}},
		{CSATBPPI, []string{"a"}},
		{CSATBAND, nil},
		{CSATBNAND, nil},
		{CSATBOR, nil},
		{CSATBNOR, nil},
		{CSATBXOR, nil},
		{CSATBXOR, []string{"a", "b", "a"}},
		{CSATBXNOR, nil},
		{CSATBXNOR, []string{"a", "b", "a"}},
		{CSATBINV, nil},
		{CSATBINV, []string{"a", "b"}},
		{CSATBBUF, nil},
		{CSATBBUF, []string{"a", "b"}},
		{CSATBPPO, nil},
		{CSATBPPO, []string{"a", "b"}},
		{CSATBPO, nil},
		{CSATBPO, []string{"a", "b"}},
	}
	for i, test := range invalid {
		before := m.GateCount()
		if err := m.AddGate(test.kind, "bad"+string(rune('A'+i)), test.fanins, 0); !errors.Is(err, ErrDimension) {
			t.Fatalf("gate %v fanins=%v: got %v, want ErrDimension", test.kind, test.fanins, err)
		}
		if m.GateCount() != before {
			t.Fatalf("gate %v fanins=%v mutated manager", test.kind, test.fanins)
		}
	}
}

func TestCSATLifecycleResultOwnershipAndOptions(t *testing.T) {
	m := newCSATTestManager(t)
	if err := m.SolveInit(); !errors.Is(err, ErrDimension) {
		t.Fatal(err)
	}
	if err := m.SetSolveOption(CSATSolveOption(255)); !errors.Is(err, ErrDimension) {
		t.Fatal(err)
	}
	if err := m.SetCaller(CSATCaller(255)); !errors.Is(err, ErrDimension) {
		t.Fatal(err)
	}
	if err := m.SetSolveOption(CSATExplicitLearning); err != nil {
		t.Fatal(err)
	}
	if err := m.SetCaller(CSATCallerSatori); err != nil {
		t.Fatal(err)
	}
	if err := m.UseOnlyCoreSatSolver(); err != nil {
		t.Fatal(err)
	}
	if m.option != CSATExplicitLearning || m.caller != CSATCallerSatori || !m.coreOnly {
		t.Fatal("option state was not retained")
	}
	for _, set := range []func(int) error{m.SetTimeLimit, m.SetLearnLimit, m.SetLearnBacktrackLimit, m.SetSolveImplicationLimit, m.SetSolveBacktrackLimit} {
		if err := set(-1); !errors.Is(err, ErrDimension) {
			t.Fatal(err)
		}
		if err := set(3); err != nil {
			t.Fatal(err)
		}
	}
	addCSATTestGate(t, m, CSATBPI, "a")
	addCSATTestGate(t, m, CSATBINV, "n", "a")
	addCSATTestGate(t, m, CSATBPO, "z", "n")
	if err := m.Finalize(); err != nil || !m.finalized {
		t.Fatal(err, m.finalized)
	}
	if err := m.AnalyzeTargets(); err != nil {
		t.Fatal(err)
	}
	if status := solveCSATTestTarget(t, m, []string{"z"}, []int{1}); status != CSATSatisfiable {
		t.Fatal(status)
	}
	first, err := m.TargetResult(1)
	if err != nil || len(first.Names) != 1 || first.Names[0] != "a" || first.Values[0] != 0 {
		t.Fatal(first, err)
	}
	first.Names[0], first.Values[0] = "changed", 7
	second, err := m.TargetResult(1)
	if err != nil || second.Names[0] != "a" || second.Values[0] != 0 {
		t.Fatal("result storage escaped", second, err)
	}
	formatted, err := m.FormatResult(1)
	if err != nil || formatted != "status=2 a=0" {
		t.Fatal(formatted, err)
	}
	if _, err = m.Solve(); !errors.Is(err, ErrInvariant) {
		t.Fatal("Solve without SolveInit", err)
	}
	if err = m.Close(); err != nil {
		t.Fatal(err)
	}
	if err = m.Close(); err != nil {
		t.Fatal("Close is not idempotent", err)
	}
	if err = m.AddGate(CSATBPI, "after", nil, 0); !errors.Is(err, ErrInvariant) {
		t.Fatal(err)
	}
}

func TestCSATIntegrityCancellationAndBudgets(t *testing.T) {
	m := newCSATTestManager(t)
	addCSATTestGate(t, m, CSATConst, "first")
	addCSATTestGate(t, m, CSATConst, "dangling")
	ok, err := m.CheckIntegrity()
	if err != nil || ok {
		t.Fatal(ok, err)
	}
	addCSATTestGate(t, m, CSATBPO, "z", "dangling")
	ok, err = m.CheckIntegrity()
	if err != nil || !ok {
		t.Fatal(ok, err)
	}
	if err = m.SetTotalBacktrackLimit(2); err != nil {
		t.Fatal(err)
	}
	if err = m.SetTotalInspectLimit(3); err != nil {
		t.Fatal(err)
	}
	m.totalBacktracks, m.totalInspects = 2, 0
	if got := solveCSATTestTarget(t, m, []string{"z"}, []int{1}); got != CSATUndetermined {
		t.Fatal(got)
	}
	if m.TotalBacktracksMade() != 2 || m.TotalInspectsMade() != 0 {
		t.Fatal("budget counters changed without solving")
	}

	calls := 0
	cancelled, err := NewCSATManager(CSATLimits{Stop: func() bool {
		calls++
		return calls >= 2
	}})
	if err != nil {
		t.Fatal(err)
	}
	defer cancelled.Close()
	if err = cancelled.AddGate(CSATBPI, "a", nil, 0); !errors.Is(err, ErrCancelled) || cancelled.GateCount() != 0 {
		t.Fatal(err, cancelled.GateCount())
	}

	stop := false
	atomic, err := NewCSATManager(CSATLimits{Stop: func() bool { return stop }})
	if err != nil {
		t.Fatal(err)
	}
	defer atomic.Close()
	addCSATTestGate(t, atomic, CSATBPI, "a")
	if _, err = atomic.AddTarget([]string{"a"}, []int{1}); err != nil {
		t.Fatal(err)
	}
	if err = atomic.SolveInit(); err != nil {
		t.Fatal(err)
	}
	stop = true
	if _, err = atomic.Solve(); !errors.Is(err, ErrCancelled) || !atomic.ready {
		t.Fatal("cancelled solve was not atomic", err, atomic.ready)
	}
}

func TestCSATBenchStandalone(t *testing.T) {
	m := newCSATTestManager(t)
	addCSATTestGate(t, m, CSATBPI, "a")
	addCSATTestGate(t, m, CSATConst, "one")
	addCSATTestGate(t, m, CSATBAND, "n", "a", "one")
	addCSATTestGate(t, m, CSATBPO, "z", "n")
	want := "INPUT(a)\nOUTPUT(z)\none = 1\nn = AND(a,one)\nz = BUFF(n)\n"
	data, err := m.Bench()
	if err != nil || string(data) != want {
		t.Fatalf("got %q err=%v", data, err)
	}
	path := filepath.Join(t.TempDir(), "network.bench")
	if err = m.EnableDump(path); err != nil {
		t.Fatal(err)
	}
	if err = m.EnableDump("bad\x00path"); !errors.Is(err, ErrDimension) {
		t.Fatal(err)
	}
	if err = m.EnableDump(path); err != nil {
		t.Fatal(err)
	}
	if err = m.DumpBench(); err != nil {
		t.Fatal(err)
	}
	written, err := os.ReadFile(path)
	if err != nil || string(written) != want {
		t.Fatal(string(written), err)
	}
	if !strings.Contains(string(written), "z = BUFF(n)") {
		t.Fatal("output is declared but undefined")
	}
}

type csatCampaignRNG uint64

func (r *csatCampaignRNG) next() uint64 {
	x := uint64(*r)
	x ^= x << 13
	x ^= x >> 7
	x ^= x << 17
	*r = csatCampaignRNG(x)
	return x
}

func runCSATModelCase(t *testing.T, seed uint64) {
	t.Helper()
	if seed == 0 {
		seed = 1
	}
	rng := csatCampaignRNG(seed)
	m, err := NewCSATManager(CSATLimits{})
	if err != nil {
		t.Fatal(err)
	}
	defer m.Close()
	inputCount := 1 + int(rng.next()%4)
	names := make([]string, inputCount)
	values := make([]bool, inputCount)
	for i := range names {
		names[i] = "i" + string(rune('0'+i))
		values[i] = rng.next()&1 != 0
		if err = m.AddGate(CSATBPI, names[i], nil, 0); err != nil {
			t.Fatal(err)
		}
	}
	gateCount := 1 + int(rng.next()%9)
	allNames := slices.Clone(names)
	allValues := slices.Clone(values)
	kinds := [...]CSATGateType{CSATBAND, CSATBNAND, CSATBOR, CSATBNOR, CSATBXOR, CSATBXNOR, CSATBINV, CSATBBUF}
	for i := 0; i < gateCount; i++ {
		kind := kinds[rng.next()%uint64(len(kinds))]
		arity := 1
		if kind >= CSATBAND && kind <= CSATBNOR {
			arity = 1 + int(rng.next()%3)
		} else if kind == CSATBXOR || kind == CSATBXNOR {
			arity = 1 + int(rng.next()%2)
		}
		fanins := make([]string, arity)
		faninValues := make([]bool, arity)
		for j := range fanins {
			id := int(rng.next() % uint64(len(allNames)))
			fanins[j], faninValues[j] = allNames[id], allValues[id]
		}
		name := "n" + string(rune('A'+i))
		if err = m.AddGate(kind, name, fanins, int(rng.next())); err != nil {
			t.Fatal(err)
		}
		allNames = append(allNames, name)
		allValues = append(allValues, csatExpected(kind, faninValues))
	}
	targetNames := append(slices.Clone(names), allNames[len(allNames)-1])
	targetValues := make([]int, len(targetNames))
	for i := range names {
		targetValues[i] = boolToIntCSAT(values[i])
	}
	targetValues[len(targetValues)-1] = boolToIntCSAT(allValues[len(allValues)-1])
	if got := solveCSATTestTargetTB(t, m, targetNames, targetValues); got != CSATSatisfiable {
		t.Fatalf("seed=%d matching model status=%v", seed, got)
	}
	targetValues[len(targetValues)-1] ^= 1
	if got := solveCSATTestTargetTB(t, m, targetNames, targetValues); got != CSATUnsatisfiable {
		t.Fatalf("seed=%d contradictory model status=%v", seed, got)
	}
}

func solveCSATTestTargetTB(t *testing.T, m *CSATManager, names []string, values []int) CSATStatus {
	t.Helper()
	if _, err := m.AddTarget(names, values); err != nil {
		t.Fatal(err)
	}
	if err := m.SolveInit(); err != nil {
		t.Fatal(err)
	}
	status, err := m.Solve()
	if err != nil {
		t.Fatal(err)
	}
	return status
}

func TestCSATIndependentModelCampaign5000(t *testing.T) {
	for seed := uint64(1); seed <= 5000; seed++ {
		runCSATModelCase(t, seed*0x9e3779b97f4a7c15)
	}
}

func FuzzCSATIndependentModel(f *testing.F) {
	for _, seed := range []uint64{1, 2, 3, 0xdeadbeef, ^uint64(0)} {
		f.Add(seed)
	}
	f.Fuzz(func(t *testing.T, seed uint64) {
		runCSATModelCase(t, seed)
	})
}
