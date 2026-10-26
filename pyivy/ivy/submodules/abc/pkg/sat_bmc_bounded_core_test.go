package abc

// Stand-alone tests for bounded learned-core extraction from bmcBCore.c.

import (
	"errors"
	"os"
	"path/filepath"
	"slices"
	"strings"
	"testing"
)

func boundedCoreSequentialFixture(t testing.TB) (*GiaManager, GiaID) {
	t.Helper()
	manager := NewGiaManager()
	register := giaTestCI(t, manager)
	giaTestCO(t, manager, register)
	giaTestCO(t, manager, register)
	if err := manager.SetRegisterCount(1); err != nil {
		t.Fatal(err)
	}
	giaTestCheck(t, manager)
	return manager, register.ID()
}

func TestGiaBoundedCorePivotParsingAndCollection(t *testing.T) {
	values, err := ParseGiaBoundedCorePivots([]byte(" 1\n2\t-3  2 "), 4)
	if err != nil || !slices.Equal(values, []int{1, 2, -3, 2}) {
		t.Fatal(values, err)
	}
	for _, test := range []struct {
		data string
		max  int
	}{
		{"x", 1}, {"1 2", 1}, {"1", -1},
	} {
		if result, parseErr := ParseGiaBoundedCorePivots([]byte(test.data), test.max); result != nil || parseErr == nil {
			t.Fatal(test, result, parseErr)
		}
	}
	directory := t.TempDir()
	path := filepath.Join(directory, "pivots.txt")
	if err = os.WriteFile(path, []byte("1 0 99"), 0o600); err != nil {
		t.Fatal(err)
	}
	read, err := ReadGiaBoundedCorePivots(path, 32, 3)
	if err != nil || !slices.Equal(read, []int{1, 0, 99}) {
		t.Fatal(read, err)
	}
	if read, err = ReadGiaBoundedCorePivots(path, 2, 3); read != nil || !errors.Is(err, ErrLimit) {
		t.Fatal(read, err)
	}
	manager, register := boundedCoreSequentialFixture(t)
	cnf, err := manager.BuildGiaBoundedCoreCNF(GiaBoundedCoreOptions{Frame: 2, PivotObjectIDs: []int{int(register), 0, 99, int(register)}})
	if err != nil {
		t.Fatal(err)
	}
	if !slices.Equal(cnf.InvalidPivotObjectIDs, []int{0, 99}) || len(cnf.PivotVariables) != 4 {
		t.Fatal(cnf.PivotVariables, cnf.InvalidPivotObjectIDs, cnf.Variables)
	}
	frames := make([]int, len(cnf.PivotVariables))
	for index, variable := range cnf.PivotVariables {
		frames[index] = cnf.Variables[variable].Frame
	}
	if !slices.Equal(frames, []int{2, 1, 0, -1}) {
		t.Fatal(frames, cnf.Variables)
	}
}

func solveBoundedCoreCNF(t testing.TB, cnf GiaBoundedCoreCNF, assumptions []SatLit, clauses []int) SatResult {
	t.Helper()
	solver, err := NewSatSolverWithLimit(len(cnf.Variables))
	if err != nil {
		t.Fatal(err)
	}
	if err = solver.SetVariableCount(len(cnf.Variables)); err != nil {
		t.Fatal(err)
	}
	if clauses == nil {
		clauses = make([]int, len(cnf.Clauses))
		for index := range clauses {
			clauses[index] = index + 1
		}
	}
	for _, handle := range clauses {
		if handle <= 0 || handle > len(cnf.Clauses) {
			t.Fatal("bad root-core handle", handle)
		}
		if ok, addErr := solver.AddClause(cnf.Clauses[handle-1]); addErr != nil || !ok {
			t.Fatal(handle, ok, addErr)
		}
	}
	result, err := solver.Solve(assumptions, SatLimits{})
	if err != nil {
		t.Fatal(err)
	}
	return result
}

func TestGiaBoundedCoreCombinationalCNFSemantics(t *testing.T) {
	manager := NewGiaManager()
	a, b := giaTestCI(t, manager), giaTestCI(t, manager)
	and, err := manager.AppendAnd(a.Not(), b)
	if err != nil {
		t.Fatal(err)
	}
	giaTestCO(t, manager, and)
	cnf, err := manager.BuildGiaBoundedCoreCNF(GiaBoundedCoreOptions{Frame: 0, PivotObjectIDs: []int{int(and.ID())}})
	if err != nil {
		t.Fatal(err)
	}
	if len(cnf.Variables) != 6 || len(cnf.Clauses) != 6 || cnf.Literals != 12 || len(cnf.PivotVariables) != 1 {
		t.Fatal(len(cnf.Variables), len(cnf.Clauses), cnf.Literals, cnf.PivotVariables)
	}
	inputVariables := make(map[GiaID]SatVar)
	for variable, mapped := range cnf.Variables {
		if mapped.Object == a.ID() || mapped.Object == b.ID() {
			inputVariables[mapped.Object] = SatVar(variable)
		}
	}
	for word := 0; word < 4; word++ {
		inputs := []bool{word&1 != 0, word&2 != 0}
		assumptions := []SatLit{
			boundedCoreLit(int(inputVariables[a.ID()]), !inputs[0]),
			boundedCoreLit(int(inputVariables[b.ID()]), !inputs[1]),
		}
		answer := solveBoundedCoreCNF(t, cnf, assumptions, nil)
		want := !inputs[0] && inputs[1]
		if (answer.Status == SatSatisfiable) != want {
			t.Fatal(word, answer.Status, want, cnf)
		}
	}
	before := slices.Clone(manager.objects)
	if _, err = manager.BuildGiaBoundedCoreCNF(GiaBoundedCoreOptions{Frame: 0}); err != nil || !slices.Equal(before, manager.objects) {
		t.Fatal("GIA scratch changed", err)
	}
}

func TestGiaBoundedCoreSequentialProofAndFormatting(t *testing.T) {
	manager, register := boundedCoreSequentialFixture(t)
	options := GiaBoundedCoreOptions{Frame: 0, PivotObjectIDs: []int{int(register)}}
	cnf, err := manager.BuildGiaBoundedCoreCNF(options)
	if err != nil {
		t.Fatal(err)
	}
	if len(cnf.Variables) != 5 || len(cnf.Clauses) != 5 {
		t.Fatal(cnf.Variables, cnf.Clauses)
	}
	result, err := manager.BoundedCore(options)
	if err != nil || result.Status != SatUnsatisfiable || len(result.RootClauseHandles) == 0 || len(result.LearnedCore) == 0 {
		t.Fatal(result.Status, result.RootClauseHandles, result.LearnedCore, err)
	}
	if core := solveBoundedCoreCNF(t, cnf, nil, result.RootClauseHandles); core.Status != SatUnsatisfiable {
		t.Fatal("reported root core is not independently UNSAT", core.Status)
	}
	finals := 0
	for index, clause := range result.LearnedCore {
		if index > 0 && result.LearnedCore[index-1].Node >= clause.Node {
			t.Fatal("proof order", result.LearnedCore)
		}
		if clause.Final {
			finals++
			if len(clause.Literals) != 0 {
				t.Fatal("nonempty final clause", clause)
			}
		}
	}
	if finals != 1 || !result.LearnedCore[len(result.LearnedCore)-1].Final {
		t.Fatal("final proof node", result.LearnedCore)
	}
	text, err := FormatGiaBoundedCoreProof(result)
	if err != nil || !strings.HasPrefix(text, "UNSAT contains ") || !strings.Contains(text, "Empty\n") || !strings.Contains(text, "(0)") {
		t.Fatal(text, err)
	}
	path := filepath.Join(t.TempDir(), "proof.txt")
	if err = WriteGiaBoundedCoreProof(path, result); err != nil {
		t.Fatal(err)
	}
	written, err := os.ReadFile(path)
	if err != nil || string(written) != text {
		t.Fatal(string(written), text, err)
	}
	sat := NewGiaManager()
	giaTestCO(t, sat, GiaConst1)
	satResult, err := sat.BoundedCore(GiaBoundedCoreOptions{Frame: 0})
	if err != nil || satResult.Status != SatSatisfiable || len(satResult.LearnedCore) != 0 {
		t.Fatal(satResult, err)
	}
	if _, err = FormatGiaBoundedCoreProof(satResult); !errors.Is(err, ErrInvariant) {
		t.Fatal(err)
	}
}

func TestGiaBoundedCoreSequentialReachability(t *testing.T) {
	manager := NewGiaManager()
	register := giaTestCI(t, manager)
	giaTestCO(t, manager, register)
	giaTestCO(t, manager, register.Not())
	if err := manager.SetRegisterCount(1); err != nil {
		t.Fatal(err)
	}
	for frame := 0; frame < 4; frame++ {
		result, err := manager.BoundedCore(GiaBoundedCoreOptions{Frame: frame, MaxVariables: 100, MaxClauses: 500, MaxLiterals: 1500})
		want := SatUnsatisfiable
		if frame&1 != 0 {
			want = SatSatisfiable
		}
		if err != nil || result.Status != want {
			t.Fatal(frame, result.Status, want, err)
		}
		if want == SatUnsatisfiable && (len(result.RootClauseHandles) == 0 || len(result.LearnedCore) == 0) {
			t.Fatal(frame, "missing checked proof", result)
		}
	}
}

func TestGiaBoundedCoreLimitsCancellationExtendedAndDeep(t *testing.T) {
	manager := NewGiaManager()
	a, b := giaTestCI(t, manager), giaTestCI(t, manager)
	node, err := manager.AppendAnd(a, b)
	if err != nil {
		t.Fatal(err)
	}
	giaTestCO(t, manager, node)
	bad := []GiaBoundedCoreOptions{
		{Frame: -1}, {Output: -1}, {Output: 1}, {MaxVariables: -1},
		{MaxClauses: -1}, {MaxLiterals: -1}, {MaxVariables: 2},
		{MaxClauses: 1}, {MaxLiterals: 1},
	}
	for _, options := range bad {
		if result, err := manager.BuildGiaBoundedCoreCNF(options); len(result.Variables) != 0 || err == nil {
			t.Fatal(options, result, err)
		}
	}
	if result, err := manager.BuildGiaBoundedCoreCNF(GiaBoundedCoreOptions{Stop: func() bool { return true }}); len(result.Variables) != 0 || !errors.Is(err, ErrCancelled) {
		t.Fatal(result, err)
	}
	unknown, err := manager.BoundedCore(GiaBoundedCoreOptions{SAT: SatLimits{Stop: func() bool { return true }}})
	if err != nil || unknown.Status != SatUnknown || unknown.Reason != SatCancelled || len(unknown.LearnedCore) != 0 {
		t.Fatal(unknown, err)
	}
	for _, extended := range []func(*GiaManager, GiaLit, GiaLit) (GiaLit, error){
		func(m *GiaManager, x, _ GiaLit) (GiaLit, error) { return m.AppendBuffer(x) },
		func(m *GiaManager, x, y GiaLit) (GiaLit, error) { return m.AppendXorReal(x, y) },
	} {
		graph := NewGiaManager()
		x, y := giaTestCI(t, graph), giaTestCI(t, graph)
		gate, err := extended(graph, x, y)
		if err != nil {
			t.Fatal(err)
		}
		giaTestCO(t, graph, gate)
		if _, err = graph.BuildGiaBoundedCoreCNF(GiaBoundedCoreOptions{}); !errors.Is(err, ErrUnsupported) {
			t.Fatal(err)
		}
	}
	deep := NewGiaManager()
	left, right := giaTestCI(t, deep), giaTestCI(t, deep)
	current := left
	for index := 0; index < 10000; index++ {
		var err error
		current, err = deep.AppendAnd(current, right.NotIf(index&1 != 0))
		if err != nil {
			t.Fatal(index, err)
		}
	}
	giaTestCO(t, deep, current)
	deepCNF, err := deep.BuildGiaBoundedCoreCNF(GiaBoundedCoreOptions{MaxVariables: 20000, MaxClauses: 40000, MaxLiterals: 100000})
	if err != nil || len(deepCNF.Clauses) != 30003 {
		t.Fatal(len(deepCNF.Variables), len(deepCNF.Clauses), err)
	}
}

func FuzzGiaBoundedCoreCombinational(f *testing.F) {
	for _, seed := range []uint64{0, 1, 2, 3, 0x123456789abcdef0, ^uint64(0)} {
		f.Add(seed)
	}
	f.Fuzz(func(t *testing.T, seed uint64) {
		manager := NewGiaManager()
		signals := []GiaLit{giaTestCI(t, manager), giaTestCI(t, manager), giaTestCI(t, manager)}
		state := seed
		for count := 0; count < 7; count++ {
			state = state*6364136223846793005 + 1442695040888963407
			left := signals[int(state%uint64(len(signals)))]
			state = state*6364136223846793005 + 1442695040888963407
			right := signals[int(state%uint64(len(signals)))]
			if left.ID() == right.ID() {
				right = signals[(int(state>>8)+1)%len(signals)]
				if left.ID() == right.ID() {
					continue
				}
			}
			left = left.NotIf(state&(1<<16) != 0)
			right = right.NotIf(state&(1<<17) != 0)
			node, err := manager.AppendAnd(left, right)
			if err != nil {
				t.Fatal(err)
			}
			signals = append(signals, node)
		}
		output := signals[len(signals)-1].NotIf(seed&(1<<63) != 0)
		giaTestCO(t, manager, output)
		wantSAT := false
		for word := 0; word < 8; word++ {
			values, err := manager.Evaluate([]bool{word&1 != 0, word&2 != 0, word&4 != 0})
			if err != nil {
				t.Fatal(err)
			}
			wantSAT = wantSAT || values[0]
		}
		result, err := manager.BoundedCore(GiaBoundedCoreOptions{Frame: 0, MaxVariables: 100, MaxClauses: 1000, MaxLiterals: 3000})
		if err != nil || (result.Status == SatSatisfiable) != wantSAT {
			t.Fatal(seed, result.Status, wantSAT, err)
		}
		if result.Status == SatUnsatisfiable {
			cnf, err := manager.BuildGiaBoundedCoreCNF(GiaBoundedCoreOptions{Frame: 0, MaxVariables: 100, MaxClauses: 1000, MaxLiterals: 3000})
			if err != nil || solveBoundedCoreCNF(t, cnf, nil, result.RootClauseHandles).Status != SatUnsatisfiable {
				t.Fatal(seed, err, result.RootClauseHandles)
			}
		}
	})
}
