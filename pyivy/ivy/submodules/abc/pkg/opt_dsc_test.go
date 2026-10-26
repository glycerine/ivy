package abc

// Go-only tests for src/opt/dsc/dsc.c and dsc.h.

import (
	"errors"
	"slices"
	"testing"
)

func dscTestTruth(t testing.TB, variables int, mask uint64) *TruthTable {
	t.Helper()
	truth, err := NewConstantTruthTable(variables, false, 1<<16)
	if err != nil {
		t.Fatal(err)
	}
	for assignment := 0; assignment < 1<<uint(variables); assignment++ {
		if mask&(uint64(1)<<uint(assignment)) != 0 {
			if err = truth.Set(uint64(assignment), true); err != nil {
				t.Fatal(err)
			}
		}
	}
	return truth
}

func dscCheckResult(t testing.TB, input *TruthTable, result DSCResult) {
	t.Helper()
	if !result.Complete {
		if result.Formula != "" {
			t.Fatalf("incomplete result retained formula %q", result.Formula)
		}
		return
	}
	formula, err := ParseDauDSD(result.Formula, DauDSDOptions{MaxBytes: dscSourceMaxFormula, MaxVariables: dscSourceMaxVariables})
	if err != nil {
		t.Fatalf("parse %q: %v", result.Formula, err)
	}
	got, err := formula.TruthTable(input.Variables(), 1<<16, nil)
	if err != nil || !got.Equal(input) {
		t.Fatalf("formula %q does not reproduce %x: %v", result.Formula, input.Words(), err)
	}
}

func TestDSCConstantsVariablesAndSourceOrder(t *testing.T) {
	for _, test := range []struct {
		variables int
		mask      uint64
		formula   string
	}{
		{0, 0, "0"},
		{0, 1, "1"},
		{1, 2, "a"},
		{1, 1, "!a"},
		{2, 8, "(ba)"},
		{2, 6, "[ba]"},
		{2, 14, "!(!b!a)"},
	} {
		truth := dscTestTruth(t, test.variables, test.mask)
		result, err := DecomposeDSC(truth, nil, DSCOptions{})
		if err != nil || !result.Complete || result.Formula != test.formula {
			t.Fatalf("vars=%d mask=%x: %+v %v want %q", test.variables, test.mask, result, err, test.formula)
		}
		dscCheckResult(t, truth, result)
	}
}

func dscOracleFunctions(variables int) map[uint64]struct{} {
	logicalMask := uint64(1)<<(1<<uint(variables)) - 1
	memo := make(map[int]map[uint64]struct{})
	var build func(int) map[uint64]struct{}
	build = func(subset int) map[uint64]struct{} {
		if cached := memo[subset]; cached != nil {
			return cached
		}
		result := make(map[uint64]struct{})
		if subset == 0 {
			result[0] = struct{}{}
			result[logicalMask] = struct{}{}
			memo[subset] = result
			return result
		}
		if subset&(subset-1) == 0 {
			variable := 0
			for subset != 1<<uint(variable) {
				variable++
			}
			var truth uint64
			for assignment := 0; assignment < 1<<uint(variables); assignment++ {
				if assignment&(1<<uint(variable)) != 0 {
					truth |= uint64(1) << uint(assignment)
				}
			}
			result[truth] = struct{}{}
			result[truth^logicalMask] = struct{}{}
			memo[subset] = result
			return result
		}
		firstBit := subset & -subset
		for left := (subset - 1) & subset; left != 0; left = (left - 1) & subset {
			if left&firstBit == 0 || left == subset {
				continue
			}
			right := subset &^ left
			for first := range build(left) {
				for second := range build(right) {
					and := first & second
					xor := first ^ second
					result[and] = struct{}{}
					result[and^logicalMask] = struct{}{}
					result[xor] = struct{}{}
					result[xor^logicalMask] = struct{}{}
				}
			}
		}
		memo[subset] = result
		return result
	}
	all := make(map[uint64]struct{})
	for subset := 0; subset < 1<<uint(variables); subset++ {
		for truth := range build(subset) {
			all[truth] = struct{}{}
		}
	}
	return all
}

func TestDSCExhaustiveThreeVariableCompleteness(t *testing.T) {
	oracle := dscOracleFunctions(3)
	for mask := uint64(0); mask < 256; mask++ {
		truth := dscTestTruth(t, 3, mask)
		result, err := DecomposeDSC(truth, nil, DSCOptions{})
		if err != nil {
			t.Fatalf("%02x: %v", mask, err)
		}
		_, want := oracle[mask]
		if result.Complete != want {
			t.Fatalf("%02x complete=%v want=%v formula=%q", mask, result.Complete, want, result.Formula)
		}
		dscCheckResult(t, truth, result)
	}
}

func dscModelFormula(state *uint64, variables int) string {
	parts := make([]string, variables)
	for variable := range parts {
		parts[variable] = string(byte('a' + variable))
		if *state&1 != 0 {
			parts[variable] = "!" + parts[variable]
		}
		*state = *state*6364136223846793005 + 1442695040888963407
	}
	for len(parts) > 1 {
		first := int(*state % uint64(len(parts)))
		*state = *state*6364136223846793005 + 1442695040888963407
		second := int(*state % uint64(len(parts)-1))
		if second >= first {
			second++
		}
		left, right := parts[first], parts[second]
		open, close := "(", ")"
		if *state&1 != 0 {
			open, close = "[", "]"
		}
		joined := open + left + right + close
		if *state&2 != 0 {
			joined = "!" + joined
		}
		if first > second {
			first, second = second, first
		}
		parts[first] = joined
		parts = append(parts[:second], parts[second+1:]...)
		*state = *state*6364136223846793005 + 1442695040888963407
	}
	return parts[0]
}

func TestDSCIndependentModelCampaign(t *testing.T) {
	state := uint64(0x9e3779b97f4a7c15)
	for execution := 0; execution < 5000; execution++ {
		variables := 1 + int(state%6)
		formulaText := dscModelFormula(&state, variables)
		formula, err := ParseDauDSD(formulaText, DauDSDOptions{})
		if err != nil {
			t.Fatal(err)
		}
		truth, err := formula.TruthTable(variables, 1<<16, nil)
		if err != nil {
			t.Fatal(err)
		}
		result, err := DecomposeDSC(truth, nil, DSCOptions{})
		if err != nil || !result.Complete {
			t.Fatalf("execution %d input=%q result=%+v err=%v", execution, formulaText, result, err)
		}
		dscCheckResult(t, truth, result)
	}
}

func TestDSCWorkspaceMatchesCountsAndErrors(t *testing.T) {
	workspace, err := NewDSCWorkspace(4, 1000)
	if err != nil {
		t.Fatal(err)
	}
	if words, err := workspace.CapacityWords(); err != nil || words != 12 {
		t.Fatal(words, err)
	}
	truth := dscTestTruth(t, 4, 0x6996)
	result, err := DecomposeDSC(truth, workspace, DSCOptions{})
	if err != nil || !result.Complete {
		t.Fatal(result, err)
	}
	workspace.Close()
	workspace.Close()
	if _, err = workspace.CapacityWords(); !errors.Is(err, ErrInvariant) {
		t.Fatal(err)
	}
	if _, err = DecomposeDSC(truth, workspace, DSCOptions{}); !errors.Is(err, ErrDimension) {
		t.Fatal(err)
	}
	matches, err := DSCMatches("!([ab](cd))", 100)
	if err != nil || !slices.Equal(matches, []int{0, 10, 5, 0, 0, 0, 9, 0, 0, 0, 0}) {
		t.Fatal(matches, err)
	}
	matches, err = DSCMatches("!([ab]<abc>{def})", 100)
	if err != nil || !slices.Equal(matches, []int{0, 16, 5, 0, 0, 0, 10, 0, 0, 0, 0, 15, 0, 0, 0, 0, 0}) {
		t.Fatal(matches, err)
	}
	for _, test := range []struct {
		formula string
		want    int
	}{{"a", 0}, {"p", 0}, {"(ap)", 1}, {"(abc)", 2}, {"[abc]", 6}, {"<abc>", 3}, {"8{ab}", 3}, {"(a[bc])", 4}, {"<abc>{d!e[fg]}", 6}} {
		got, countErr := DSCAndCount(test.formula, 100)
		if countErr != nil || got != test.want {
			t.Fatalf("%q: %d %v want %d", test.formula, got, countErr, test.want)
		}
	}
	if _, err = DSCMatches("([a)", 100); err == nil {
		t.Fatal("malformed matches accepted")
	}
	if _, err = DSCAndCount("(a)", 100); err == nil {
		t.Fatal("malformed formula accepted")
	}
	if _, err = DecomposeDSC(nil, nil, DSCOptions{}); !errors.Is(err, ErrDimension) {
		t.Fatal(err)
	}
	if _, err = DecomposeDSC(truth, nil, DSCOptions{MaxVariables: 3}); !errors.Is(err, ErrLimit) {
		t.Fatal(err)
	}
	if _, err = DecomposeDSC(truth, nil, DSCOptions{MaxWords: -1}); !errors.Is(err, ErrDimension) {
		t.Fatal(err)
	}
	largeTruth := dscTestTruth(t, 7, 1)
	if result, limitErr := DecomposeDSC(largeTruth, nil, DSCOptions{MaxWords: 1}); result != (DSCResult{}) || !errors.Is(limitErr, ErrLimit) {
		t.Fatalf("word limit: %+v %v", result, limitErr)
	}
	if workspace, limitErr := NewDSCWorkspace(4, 11); workspace != nil || !errors.Is(limitErr, ErrLimit) {
		t.Fatalf("workspace word limit: %+v %v", workspace, limitErr)
	}
	if _, err = DecomposeDSC(truth, nil, DSCOptions{Stop: func() bool { return true }}); !errors.Is(err, ErrCancelled) {
		t.Fatal(err)
	}
	if result, err := DecomposeDSC(dscTestTruth(t, 3, 0xe8), nil, DSCOptions{}); err != nil || result.Complete || result.Formula != "" {
		t.Fatalf("prime function disposition: %+v %v", result, err)
	}
	if result, err := DecomposeDSC(dscTestTruth(t, 2, 8), nil, DSCOptions{MaxFormulaBytes: 3}); result != (DSCResult{}) || !errors.Is(err, ErrLimit) {
		t.Fatalf("formula limit: %+v %v", result, err)
	}
	calls := 0
	lateStop := func() bool {
		calls++
		return calls == 5
	}
	if result, err := DecomposeDSC(dscTestTruth(t, 6, 0x6996966996696996), nil, DSCOptions{Stop: lateStop}); result != (DSCResult{}) || !errors.Is(err, ErrCancelled) || calls != 5 {
		t.Fatalf("late cancellation: %+v %v calls=%d", result, err, calls)
	}
	short, _ := NewDSCWorkspace(3, 100)
	if _, err = DecomposeDSC(truth, short, DSCOptions{}); !errors.Is(err, ErrDimension) {
		t.Fatal(err)
	}
	if got := truth.Words(); !slices.Equal(got, []uint64{0x6996699669966996}) {
		t.Fatalf("input mutated: %x", got)
	}
}

func FuzzDSCDecomposition(f *testing.F) {
	f.Add(uint8(2), uint64(8))
	f.Add(uint8(3), uint64(0x96))
	f.Add(uint8(4), uint64(0x6996))
	f.Fuzz(func(t *testing.T, rawVariables uint8, mask uint64) {
		variables := int(rawVariables % 7)
		if variables < 6 {
			mask &= uint64(1)<<(1<<uint(variables)) - 1
		}
		truth := dscTestTruth(t, variables, mask)
		before := truth.Words()
		result, err := DecomposeDSC(truth, nil, DSCOptions{})
		if err != nil {
			t.Fatal(err)
		}
		dscCheckResult(t, truth, result)
		if !slices.Equal(before, truth.Words()) {
			t.Fatal("decomposition mutated input")
		}
	})
}
