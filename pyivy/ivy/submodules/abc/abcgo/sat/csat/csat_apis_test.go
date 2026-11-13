package csat

import (
	"fmt"
	"math/rand"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestRandomNetworksAgainstBooleanEvaluation(t *testing.T) {
	rng := rand.New(rand.NewSource(513))
	for trial := 0; trial < 100; trial++ {
		m := ABC_InitManager()
		if trial%2 != 0 {
			ABC_UseOnlyCoreSatSolver(m)
		}
		names := []string{"a", "b", "c"}
		for _, name := range names {
			addGate(t, m, CSAT_BPI, name)
		}
		type operation struct {
			typ  GateType
			a, b int
		}
		var operations []operation
		for i := 0; i < 6; i++ {
			op := operation{typ: []GateType{CSAT_BAND, CSAT_BOR, CSAT_BXOR, CSAT_BNAND}[rng.Intn(4)], a: rng.Intn(len(names)), b: rng.Intn(len(names))}
			name := fmt.Sprintf("g%d", i)
			addGate(t, m, op.typ, name, names[op.a], names[op.b])
			operations = append(operations, op)
			names = append(names, name)
		}
		for mask := 0; mask < 8; mask++ {
			values := []bool{mask&1 != 0, mask&2 != 0, mask&4 != 0}
			for _, op := range operations {
				a, b := values[op.a], values[op.b]
				var value bool
				switch op.typ {
				case CSAT_BAND:
					value = a && b
				case CSAT_BOR:
					value = a || b
				case CSAT_BXOR:
					value = a != b
				case CSAT_BNAND:
					value = !(a && b)
				}
				values = append(values, value)
			}
			targetNames := []string{"a", "b", "c", names[len(names)-1]}
			targetValues := []int32{boolInt(values[0]), boolInt(values[1]), boolInt(values[2]), boolInt(values[len(values)-1])}
			if got := solveTarget(t, m, targetNames, targetValues); got != SATISFIABLE {
				t.Fatalf("trial=%d mask=%d true output status=%v", trial, mask, got)
			}
			targetValues[3] ^= 1
			if got := solveTarget(t, m, targetNames, targetValues); got != UNSATISFIABLE {
				t.Fatalf("trial=%d mask=%d false output status=%v", trial, mask, got)
			}
		}
		ABC_ReleaseManager(m)
	}
}

func addGate(t *testing.T, m ABC_Manager, typ GateType, name string, inputs ...string) {
	t.Helper()
	if got := ABC_AddGate(m, typ, name, int32(len(inputs)), inputs, 0); got != 1 {
		t.Fatalf("ABC_AddGate(%s): %v", name, m.LastError())
	}
}

func solveTarget(t *testing.T, m ABC_Manager, names []string, values []int32) CSAT_StatusT {
	t.Helper()
	if ABC_AddTarget(m, int32(len(names)), names, values) != 1 {
		t.Fatalf("ABC_AddTarget: %v", m.LastError())
	}
	ABC_SolveInit(m)
	return ABC_Solve(m)
}

func TestGateTruthTables(t *testing.T) {
	cases := []struct {
		typ GateType
		fn  func(bool, bool) bool
	}{
		{CSAT_BAND, func(a, b bool) bool { return a && b }},
		{CSAT_BNAND, func(a, b bool) bool { return !(a && b) }},
		{CSAT_BOR, func(a, b bool) bool { return a || b }},
		{CSAT_BNOR, func(a, b bool) bool { return !(a || b) }},
		{CSAT_BXOR, func(a, b bool) bool { return a != b }},
		{CSAT_BXNOR, func(a, b bool) bool { return a == b }},
	}
	for _, tc := range cases {
		for _, a := range []bool{false, true} {
			for _, b := range []bool{false, true} {
				for _, want := range []bool{false, true} {
					t.Run(strings.Join([]string{gateName(tc.typ), boolName(a), boolName(b), boolName(want)}, "/"), func(t *testing.T) {
						m := ABC_InitManager()
						defer ABC_ReleaseManager(m)
						addGate(t, m, CSAT_BPI, "a")
						addGate(t, m, CSAT_BPI, "b")
						addGate(t, m, tc.typ, "y", "a", "b")
						wantVal := int32(0)
						if want {
							wantVal = 1
						}
						names := []string{"a", "b", "y"}
						values := []int32{boolInt(a), boolInt(b), wantVal}
						got := solveTarget(t, m, names, values)
						expected := UNSATISFIABLE
						if tc.fn(a, b) == want {
							expected = SATISFIABLE
						}
						if got != expected {
							t.Fatalf("status=%v, want %v", got, expected)
						}
					})
				}
			}
		}
	}
}

func TestUnaryConstantsAndModeParity(t *testing.T) {
	for mode := 0; mode <= 1; mode++ {
		m := ABC_InitManager()
		if mode == 1 {
			ABC_UseOnlyCoreSatSolver(m)
		}
		addGate(t, m, CSAT_CONST, "one")
		addGate(t, m, CSAT_BPI, "x")
		addGate(t, m, CSAT_BINV, "nx", "x")
		addGate(t, m, CSAT_BBUF, "copy", "nx")
		addGate(t, m, CSAT_BPO, "out", "copy")
		if got := solveTarget(t, m, []string{"one", "x", "nx", "out"}, []int32{1, 0, 1, 1}); got != SATISFIABLE {
			t.Fatalf("mode %d: got %v", mode, got)
		}
		res := ABC_Get_Target_Result(m, 42)
		if res == nil || res.Status != SATISFIABLE || res.NoSig != 1 || res.Names[0] != "x" || res.Values[0] != 0 {
			t.Fatalf("unexpected result: %#v", res)
		}
		res.Names[0] = "caller mutation"
		if ABC_Get_Target_Result(m, 0).Names[0] != "x" {
			t.Fatal("result must be returned as an owned copy")
		}
		ABC_ReleaseManager(m)
	}
}

func TestPPIAndPPOAliases(t *testing.T) {
	m := ABC_InitManager()
	defer ABC_ReleaseManager(m)
	addGate(t, m, CSAT_BPPI, "state")
	addGate(t, m, CSAT_BBUF, "buffer", "state")
	addGate(t, m, CSAT_BPPO, "next", "buffer")
	if got := solveTarget(t, m, []string{"state", "buffer", "next"}, []int32{1, 1, 1}); got != SATISFIABLE {
		t.Fatalf("PPI/PPO alias network returned %v", got)
	}
	if got := solveTarget(t, m, []string{"state", "next"}, []int32{0, 1}); got != UNSATISFIABLE {
		t.Fatalf("PPI/PPO aliases should preserve their value, got %v", got)
	}
}

func TestSupportedMultiInputAndUnaryXorArities(t *testing.T) {
	for _, tc := range []struct {
		typ GateType
		fn  func([]bool) bool
	}{
		{CSAT_BAND, func(v []bool) bool {
			out := true
			for _, x := range v {
				out = out && x
			}
			return out
		}},
		{CSAT_BNAND, func(v []bool) bool {
			out := true
			for _, x := range v {
				out = out && x
			}
			return !out
		}},
		{CSAT_BOR, func(v []bool) bool {
			out := false
			for _, x := range v {
				out = out || x
			}
			return out
		}},
		{CSAT_BNOR, func(v []bool) bool {
			out := false
			for _, x := range v {
				out = out || x
			}
			return !out
		}},
	} {
		for n := 1; n <= 3; n++ {
			for bits := 0; bits < 1<<n; bits++ {
				inputs := make([]bool, n)
				names := make([]string, n+1)
				values := make([]int32, n+1)
				fanins := make([]string, n)
				for i := 0; i < n; i++ {
					inputs[i] = bits&(1<<i) != 0
					fanins[i] = fmt.Sprintf("i%d", i)
					names[i] = fanins[i]
					values[i] = boolInt(inputs[i])
				}
				names[n] = "y"
				want := tc.fn(inputs)
				values[n] = boolInt(want)
				m := ABC_InitManager()
				for _, name := range fanins {
					addGate(t, m, CSAT_BPI, name)
				}
				addGate(t, m, tc.typ, "y", fanins...)
				if got := solveTarget(t, m, names, values); got != SATISFIABLE {
					t.Fatalf("%s/%d-input assignment %v: got %v", gateName(tc.typ), n, inputs, got)
				}
				values[n] = boolInt(!want)
				if got := solveTarget(t, m, names, values); got != UNSATISFIABLE {
					t.Fatalf("%s/%d-input wrong output %v: got %v", gateName(tc.typ), n, inputs, got)
				}
				ABC_ReleaseManager(m)
			}
		}
	}

	for _, tc := range []struct {
		typ     GateType
		inverse bool
	}{{CSAT_BXOR, false}, {CSAT_BXNOR, true}} {
		m := ABC_InitManager()
		addGate(t, m, CSAT_BPI, "x")
		addGate(t, m, tc.typ, "y", "x")
		want := int32(1)
		if tc.inverse {
			want = 0
		}
		if got := solveTarget(t, m, []string{"x", "y"}, []int32{1, want}); got != SATISFIABLE {
			t.Fatalf("unary %s got %v", gateName(tc.typ), got)
		}
		ABC_ReleaseManager(m)
	}
}

func TestUnsupportedCEnumPathsAreRejected(t *testing.T) {
	// The original ABC_AddGate switch has no construction case for these enum
	// values; its default path leaves pObj/pSop nil. Rejecting them is safer than
	// creating a named gate with undefined semantics.
	for _, typ := range []GateType{CSAT_BMUX, CSAT_BDFF, CSAT_BSDFF, CSAT_BTRIH, CSAT_BTRIL, CSAT_BBUS, CSAT_BCNF, CSAT_BDC} {
		m := ABC_InitManager()
		got := ABC_AddGate(m, typ, "unsupported", 0, nil, 0)
		if got != 0 || m.LastError() == nil {
			t.Errorf("gate type %d was not rejected with an error", typ)
		}
		ABC_ReleaseManager(m)
	}
}

func TestInvalidConstructionIsAtomic(t *testing.T) {
	m := ABC_InitManager()
	defer ABC_ReleaseManager(m)
	if ABC_AddGate(m, CSAT_BPI, "missing-fanin", 1, []string{"x"}, 0) != 0 {
		t.Fatal("accepted absent fanin")
	}
	if m.LastError() == nil {
		t.Fatal("expected diagnostic")
	}
	addGate(t, m, CSAT_BPI, "x")
	if ABC_AddGate(m, CSAT_BAND, "bad", 0, nil, 0) != 0 {
		t.Fatal("accepted zero-input AND")
	}
	if ABC_AddGate(m, CSAT_BDC, "dc", 2, []string{"x", "x"}, 0) != 0 {
		t.Fatal("accepted an unsupported don't-care primitive")
	}
	addGate(t, m, CSAT_BINV, "nx", "x")
	if ABC_AddGate(m, CSAT_BPI, "x", 0, nil, 0) != 0 {
		t.Fatal("accepted duplicate name")
	}
	if ABC_AddTarget(m, 1, []string{"absent"}, []int32{1}) != 0 {
		t.Fatal("accepted absent target")
	}
	if ABC_AddTarget(m, 1, []string{"x"}, []int32{2}) != 0 {
		t.Fatal("accepted non-Boolean target")
	}
}

func TestIntegrityAndConflictingTargets(t *testing.T) {
	m := ABC_InitManager()
	defer ABC_ReleaseManager(m)
	addGate(t, m, CSAT_BPI, "x")
	addGate(t, m, CSAT_CONST, "one")
	addGate(t, m, CSAT_BAND, "unused", "x", "x")
	addGate(t, m, CSAT_BPO, "out", "x")
	if ABC_Check_Integrity(m) != 0 {
		t.Fatal("dangling gate was not detected")
	}
	m2 := ABC_InitManager()
	defer ABC_ReleaseManager(m2)
	addGate(t, m2, CSAT_BPI, "x")
	if ABC_Check_Integrity(m2) != 1 {
		t.Fatal("PI-only circuit should be structurally valid")
	}
	if got := solveTarget(t, m2, []string{"x", "x"}, []int32{0, 1}); got != UNSATISFIABLE {
		t.Fatalf("conflicting targets returned %v", got)
	}
}

func TestLimitsAndBenchDump(t *testing.T) {
	m := ABC_InitManager()
	defer ABC_ReleaseManager(m)
	addGate(t, m, CSAT_BPI, "x")
	addGate(t, m, CSAT_CONST, "one")
	addGate(t, m, CSAT_BINV, "nx", "x")
	addGate(t, m, CSAT_BPO, "out", "nx")
	ABC_SetLearnLimit(m, 1)
	dir := t.TempDir()
	file := filepath.Join(dir, "circuit.bench")
	ABC_EnableDump(m, file)
	if got := solveTarget(t, m, []string{"x", "nx"}, []int32{0, 1}); got != SATISFIABLE {
		t.Fatalf("solve status %v", got)
	}
	data, err := os.ReadFile(file)
	if err != nil {
		t.Fatal(err)
	}
	text := string(data)
	for _, want := range []string{"INPUT(x)", "one = 1", "nx = NOT(x)", "OUTPUT(nx)"} {
		if !strings.Contains(text, want) {
			t.Errorf("BENCH output lacks %q:\n%s", want, text)
		}
	}
}

func TestImplicationLimitUsesPropagationBudget(t *testing.T) {
	m := ABC_InitManager()
	defer ABC_ReleaseManager(m)
	ABC_UseOnlyCoreSatSolver(m)
	addGate(t, m, CSAT_BPI, "x")
	addGate(t, m, CSAT_BPI, "y")
	addGate(t, m, CSAT_BAND, "z", "x", "y")
	ABC_SetSolveImplicationLimit(m, 1)
	if got := solveTarget(t, m, []string{"x"}, []int32{1}); got != TIME_OUT {
		t.Fatalf("propagation limit returned %v, want timeout", got)
	}
	ABC_SetSolveImplicationLimit(m, 0)
	ABC_SolveInit(m)
	if got := ABC_Solve(m); got != SATISFIABLE {
		t.Fatalf("cleared propagation limit returned %v", got)
	}
}

func boolInt(b bool) int32 {
	if b {
		return 1
	}
	return 0
}
func boolName(b bool) string {
	if b {
		return "true"
	}
	return "false"
}
func gateName(t GateType) string {
	return []string{"const", "pi", "ppi", "and", "nand", "or", "nor", "xor", "xnor", "inv", "buf"}[t]
}
