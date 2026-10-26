package abc

// Tests for the Boolean analyses ported from src/bdd/cudd/cuddEssent.c,
// cuddLiteral.c, cuddBddCorr.c, cuddDecomp.c, cuddGenCof.c, cuddSat.c,
// cuddSolve.c, and cuddSplit.c.

import (
	"errors"
	"math"
	"math/bits"
	"testing"
)

func cuddAnalysisManager(t *testing.T, variables int) *ExtraBDDManager {
	t.Helper()
	manager, err := NewExtraBDDManager(variables, 1<<16)
	if err != nil {
		t.Fatal(err)
	}
	return manager
}

func cuddAnalysisFromTable(t *testing.T, manager *ExtraBDDManager, table uint64) ExtraBDDLit {
	t.Helper()
	count := 1 << uint(manager.variables)
	values := make([]bool, count)
	for assignment := range values {
		values[assignment] = table&(uint64(1)<<uint(assignment)) != 0
	}
	options, err := normalizeExtraBDDOptions(ExtraBDDOptions{})
	if err != nil {
		t.Fatal(err)
	}
	transaction := manager.beginTransaction()
	work := extraBDDWork{options: options, transaction: &transaction}
	root, err := manager.analysisBuild(values, &work)
	if err != nil {
		manager.rollback(&transaction)
		t.Fatal(err)
	}
	return root
}

func cuddAnalysisTable(t *testing.T, manager *ExtraBDDManager, root ExtraBDDLit) uint64 {
	t.Helper()
	var table uint64
	assignment := make([]bool, manager.variables)
	for index := 0; index < 1<<uint(manager.variables); index++ {
		for variable := range assignment {
			assignment[variable] = index&(1<<uint(variable)) != 0
		}
		value, err := manager.Evaluate(root, assignment)
		if err != nil {
			t.Fatal(err)
		}
		if value {
			table |= uint64(1) << uint(index)
		}
	}
	return table
}

func cuddAnalysisEssential(table uint64, variables int) []CuddLiteral {
	result := make([]CuddLiteral, 0, variables)
	for variable := 0; variable < variables; variable++ {
		canZero, canOne := false, false
		for assignment := 0; assignment < 1<<uint(variables); assignment++ {
			if table&(uint64(1)<<uint(assignment)) == 0 {
				continue
			}
			if assignment&(1<<uint(variable)) == 0 {
				canZero = true
			} else {
				canOne = true
			}
		}
		if canZero != canOne {
			result = append(result, CuddLiteral{Variable: variable, Value: canOne})
		}
	}
	return result
}

func TestCuddAnalysisExhaustiveFourVariableFunctions(t *testing.T) {
	manager := cuddAnalysisManager(t, 4)
	for table := uint64(0); table < 1<<16; table++ {
		root := cuddAnalysisFromTable(t, manager, table)
		count, err := manager.SatCount(root, ExtraBDDOptions{})
		if err != nil || count != uint64(bits.OnesCount64(table)) {
			t.Fatalf("table %#x: SatCount=(%d,%v)", table, count, err)
		}
		essential, err := manager.EssentialLiterals(root, ExtraBDDOptions{})
		wantEssential := cuddAnalysisEssential(table, 4)
		if err != nil || !equalCuddLiterals(essential, wantEssential) {
			t.Fatalf("table %#x: essential=%v want %v err=%v", table, essential, wantEssential, err)
		}
		clauses, err := manager.TwoLiteralClauses(root, ExtraBDDOptions{})
		if err != nil {
			t.Fatalf("table %#x: clauses: %v", table, err)
		}
		for _, clause := range clauses {
			for assignment := 0; assignment < 16; assignment++ {
				if table&(uint64(1)<<uint(assignment)) == 0 {
					continue
				}
				first := assignment&(1<<uint(clause.First.Variable)) != 0
				second := assignment&(1<<uint(clause.Second.Variable)) != 0
				if first != clause.First.Value && (!clause.Unit && second != clause.Second.Value) {
					t.Fatalf("table %#x: clause %+v rejected model %d", table, clause, assignment)
				}
			}
		}
	}
}

func equalCuddLiterals(left, right []CuddLiteral) bool {
	if len(left) != len(right) {
		return false
	}
	for index := range left {
		if left[index] != right[index] {
			return false
		}
	}
	return true
}

func TestCuddAnalysisAlgebraAndReconstruction(t *testing.T) {
	manager := cuddAnalysisManager(t, 4)
	x, _ := manager.Variable(0)
	y, _ := manager.Variable(1)
	z, _ := manager.Variable(2)
	w, _ := manager.Variable(3)
	xy, _ := manager.And(x, y, ExtraBDDOptions{})
	zw, _ := manager.And(z, w, ExtraBDDOptions{})
	conjunction, _ := manager.And(xy, zw, ExtraBDDOptions{})
	disjunction, _ := manager.Or(xy, zw, ExtraBDDOptions{})

	conj, err := manager.ConjunctiveDecomposition(conjunction, ExtraBDDOptions{})
	if err != nil {
		t.Fatal(err)
	}
	rebuilt, err := manager.And(conj.First, conj.Second, ExtraBDDOptions{})
	if err != nil || cuddAnalysisTable(t, manager, rebuilt) != cuddAnalysisTable(t, manager, conjunction) {
		t.Fatalf("conjunctive reconstruction failed: %v", err)
	}
	disj, err := manager.DisjunctiveDecomposition(disjunction, ExtraBDDOptions{})
	if err != nil {
		t.Fatal(err)
	}
	rebuilt, err = manager.Or(disj.First, disj.Second, ExtraBDDOptions{})
	if err != nil || cuddAnalysisTable(t, manager, rebuilt) != cuddAnalysisTable(t, manager, disjunction) {
		t.Fatalf("disjunctive reconstruction failed: %v", err)
	}

	if increasing, err := manager.Increasing(disjunction, 0, ExtraBDDOptions{}); err != nil || !increasing {
		t.Fatalf("Increasing=(%v,%v)", increasing, err)
	}
	if decreasing, err := manager.Decreasing(disjunction.Not(), 0, ExtraBDDOptions{}); err != nil || !decreasing {
		t.Fatalf("Decreasing=(%v,%v)", decreasing, err)
	}
	if equivalent, err := manager.EquivalentDontCare(x, y, ExtraBDDTrue, ExtraBDDOptions{}); err != nil || !equivalent {
		t.Fatalf("EquivalentDontCare=(%v,%v)", equivalent, err)
	}
	if leq, err := manager.LeqUnless(x, ExtraBDDFalse, x, ExtraBDDOptions{}); err != nil || !leq {
		t.Fatalf("LeqUnless=(%v,%v)", leq, err)
	}

	correlation, err := manager.Correlation(x, y, nil, ExtraBDDOptions{})
	if err != nil || math.Abs(correlation-0.5) > 1e-12 {
		t.Fatalf("Correlation=%g,%v", correlation, err)
	}
	correlation, err = manager.Correlation(x, y, []float64{0.8, 0.25, 0.5, 0.5}, ExtraBDDOptions{})
	if err != nil || math.Abs(correlation-0.35) > 1e-12 {
		t.Fatalf("weighted Correlation=%g,%v", correlation, err)
	}

	cube, err := manager.LargestCube(disjunction, ExtraBDDOptions{})
	if err != nil || !cube.Exists || len(cube.Assignments) != 2 || cube.Minterms != 4 {
		t.Fatalf("LargestCube=%+v,%v", cube, err)
	}
	if subset, err := manager.Leq(cube.Root, disjunction, ExtraBDDOptions{}); err != nil || !subset {
		t.Fatalf("cube implication=(%v,%v)", subset, err)
	}
	length, err := manager.ShortestLength(disjunction, ExtraBDDOptions{})
	if err != nil || length != 2 {
		t.Fatalf("ShortestLength=(%d,%v)", length, err)
	}
	empty, err := manager.ShortestPath(ExtraBDDFalse, ExtraBDDOptions{})
	if err != nil || empty.Exists || empty.Root != ExtraBDDFalse {
		t.Fatalf("empty ShortestPath=%+v,%v", empty, err)
	}
	if length, err = manager.ShortestLength(ExtraBDDFalse, ExtraBDDOptions{}); err != nil || length != -1 {
		t.Fatalf("empty ShortestLength=(%d,%v)", length, err)
	}
}

func TestCuddAnalysisGeneralizedCofactorsSplitAndEquation(t *testing.T) {
	manager := cuddAnalysisManager(t, 3)
	x, _ := manager.Variable(0)
	y, _ := manager.Variable(1)
	z, _ := manager.Variable(2)
	f, _ := manager.Xor(x, y, ExtraBDDOptions{})
	care, _ := manager.Or(x, z, ExtraBDDOptions{})
	cofactor, err := manager.Constrain(f, care, ExtraBDDOptions{})
	if err != nil {
		t.Fatal(err)
	}
	difference, _ := manager.Xor(f, cofactor, ExtraBDDOptions{})
	if agrees, err := manager.Leq(difference, care.Not(), ExtraBDDOptions{}); err != nil || !agrees {
		t.Fatalf("generalized cofactor agreement=(%v,%v)", agrees, err)
	}
	if constrained, err := manager.Constrain(f, ExtraBDDFalse, ExtraBDDOptions{}); err != nil || constrained != ExtraBDDFalse {
		t.Fatalf("Constrain(f,0)=(%d,%v)", constrained, err)
	}
	nonPolluting, err := manager.NPAnd(x, z, ExtraBDDOptions{})
	if err != nil || nonPolluting != x {
		t.Fatalf("NPAnd(x,z)=(%d,%v), want x", nonPolluting, err)
	}

	all, _ := manager.Or(f, z, ExtraBDDOptions{})
	available, _ := manager.SatCount(all, ExtraBDDOptions{})
	for requested := uint64(0); requested <= available; requested++ {
		split, splitErr := manager.SplitSet(all, requested, ExtraBDDOptions{})
		if splitErr != nil {
			t.Fatalf("SplitSet(%d): %v", requested, splitErr)
		}
		got, _ := manager.SatCount(split, ExtraBDDOptions{})
		subset, _ := manager.Leq(split, all, ExtraBDDOptions{})
		if got != requested || !subset {
			t.Fatalf("SplitSet(%d): count=%d subset=%v", requested, got, subset)
		}
	}
	if _, err := manager.SplitSet(all, available+1, ExtraBDDOptions{}); !errors.Is(err, ErrDimension) {
		t.Fatalf("oversized split error=%v", err)
	}

	solution, err := manager.SolveZeroEquation(f, []int{1}, ExtraBDDOptions{})
	if err != nil {
		t.Fatal(err)
	}
	verified, err := manager.VerifyZeroEquation(f, solution, ExtraBDDOptions{})
	if err != nil || !verified {
		t.Fatalf("VerifyZeroEquation=(%v,%v)", verified, err)
	}
	if solution.Consistency != ExtraBDDTrue || cuddAnalysisTable(t, manager, solution.Solutions[0]) != cuddAnalysisTable(t, manager, x) {
		t.Fatalf("unexpected equation solution: %+v", solution)
	}
	cube, _ := manager.And(x, y, ExtraBDDOptions{})
	prime, err := manager.MakePrime(cube, f.Not(), ExtraBDDOptions{})
	if err != nil || cuddAnalysisTable(t, manager, prime) != cuddAnalysisTable(t, manager, cube) {
		t.Fatalf("MakePrime=%#x,%v", cuddAnalysisTable(t, manager, prime), err)
	}

	intersection, err := CuddLiteralSetIntersection(
		[]CuddLiteral{{Variable: 2, Value: true}, {Variable: 0, Value: false}},
		[]CuddLiteral{{Variable: 0, Value: false}, {Variable: 1, Value: true}},
	)
	if err != nil || !equalCuddLiterals(intersection, []CuddLiteral{{Variable: 0, Value: false}}) {
		t.Fatalf("intersection=%v,%v", intersection, err)
	}
}

func TestCuddAnalysisConstantAndSetSentinels(t *testing.T) {
	manager := cuddAnalysisManager(t, 3)
	for _, root := range []ExtraBDDLit{ExtraBDDFalse, ExtraBDDTrue} {
		clauses, err := manager.TwoLiteralClauses(root, ExtraBDDOptions{})
		if err != nil || len(clauses) != 0 {
			t.Fatalf("TwoLiteralClauses(%d)=%v,%v, want empty set", root, clauses, err)
		}
	}
	for variable := 0; variable < 3; variable++ {
		for _, value := range []bool{false, true} {
			essential, err := manager.VariableEssential(ExtraBDDFalse, variable, value, ExtraBDDOptions{})
			if err != nil || !essential {
				t.Fatalf("VariableEssential(0,%d,%v)=(%v,%v), want vacuous implication", variable, value, essential, err)
			}
			essential, err = manager.VariableEssential(ExtraBDDTrue, variable, value, ExtraBDDOptions{})
			if err != nil || essential {
				t.Fatalf("VariableEssential(1,%d,%v)=(%v,%v)", variable, value, essential, err)
			}
		}
	}

	intersection, err := CuddLiteralSetIntersection(
		[]CuddLiteral{{Variable: 1, Value: true}},
		[]CuddLiteral{{Variable: 1, Value: true}, {Variable: 1, Value: true}},
	)
	if err != nil || !equalCuddLiterals(intersection, []CuddLiteral{{Variable: 1, Value: true}}) {
		t.Fatalf("duplicate intersection=%v,%v", intersection, err)
	}
}

func TestCuddAnalysisRestrictCareIsNotConstrain(t *testing.T) {
	manager := cuddAnalysisManager(t, 3)
	x, _ := manager.Variable(0)
	y, _ := manager.Variable(1)
	z, _ := manager.Variable(2)
	root, _ := manager.Ite(x, y, z, ExtraBDDOptions{})
	care, _ := manager.Xor(y, z, ExtraBDDOptions{})
	before, _ := manager.Statistics()
	restricted, err := manager.RestrictCare(root, care, ExtraBDDOptions{})
	if err != nil || restricted != root {
		t.Fatalf("RestrictCare=%d,%v, want source algorithm's size-safe root %d", restricted, err, root)
	}
	after, _ := manager.Statistics()
	if after.Nodes != before.Nodes {
		t.Fatalf("size-safe restrict retained discarded nodes: before=%d after=%d", before.Nodes, after.Nodes)
	}
	constrained, err := manager.Constrain(root, care, ExtraBDDOptions{})
	if err != nil {
		t.Fatal(err)
	}
	if cuddAnalysisTable(t, manager, constrained) == cuddAnalysisTable(t, manager, restricted) {
		t.Fatal("restrict and constrain unexpectedly collapsed to the same operation")
	}
	difference, _ := manager.Xor(root, restricted, ExtraBDDOptions{})
	if agrees, agreeErr := manager.Leq(difference, care.Not(), ExtraBDDOptions{}); agreeErr != nil || !agrees {
		t.Fatalf("restrict agreement=(%v,%v)", agrees, agreeErr)
	}
}

func TestCuddAnalysisLimitsCancellationAndRollback(t *testing.T) {
	manager := cuddAnalysisManager(t, 4)
	root := cuddAnalysisFromTable(t, manager, 0x6996)
	before, _ := manager.Statistics()
	if _, err := manager.SplitSet(root, 3, ExtraBDDOptions{Stop: func() bool { return true }}); !errors.Is(err, ErrCancelled) {
		t.Fatalf("cancellation error=%v", err)
	}
	after, _ := manager.Statistics()
	if before.Nodes != after.Nodes || before.CacheLookups != after.CacheLookups || before.CacheMisses != after.CacheMisses {
		t.Fatalf("cancelled operation changed manager: before=%+v after=%+v", before, after)
	}
	if _, err := manager.LargestCube(root, ExtraBDDOptions{MaxOperations: 1}); !errors.Is(err, ErrLimit) {
		t.Fatalf("operation limit error=%v", err)
	}
	after, _ = manager.Statistics()
	if before.Nodes != after.Nodes {
		t.Fatalf("limited operation leaked nodes: before=%d after=%d", before.Nodes, after.Nodes)
	}
	if _, err := manager.SatCount(root, ExtraBDDOptions{MaxBytes: 1}); !errors.Is(err, ErrLimit) {
		t.Fatalf("byte limit error=%v", err)
	}
}

func TestCuddAnalysisIndependentRandomExpressions5000(t *testing.T) {
	manager := cuddAnalysisManager(t, 4)
	variables := make([]ExtraBDDLit, 4)
	tables := []uint64{0xaaaa, 0xcccc, 0xf0f0, 0xff00}
	for index := range variables {
		variables[index], _ = manager.Variable(index)
	}
	state := uint64(0x8f3f73b5cf1c9ade)
	for iteration := 0; iteration < 5000; iteration++ {
		state ^= state << 7
		state ^= state >> 9
		state ^= state << 8
		left := int(state & 3)
		right := int((state >> 2) & 3)
		switch (state >> 4) % 3 {
		case 0:
			variables[left], _ = manager.And(variables[left], variables[right], ExtraBDDOptions{})
			tables[left] &= tables[right]
		case 1:
			variables[left], _ = manager.Or(variables[left], variables[right], ExtraBDDOptions{})
			tables[left] |= tables[right]
		default:
			variables[left], _ = manager.Xor(variables[left], variables[right], ExtraBDDOptions{})
			tables[left] ^= tables[right]
		}
		if state&(1<<8) != 0 {
			variables[left] = variables[left].Not()
			tables[left] = ^tables[left] & 0xffff
		}
		if got := cuddAnalysisTable(t, manager, variables[left]); got != tables[left] {
			t.Fatalf("iteration %d: expression table=%#x want %#x", iteration, got, tables[left])
		}
		count, err := manager.SatCount(variables[left], ExtraBDDOptions{})
		if err != nil || count != uint64(bits.OnesCount64(tables[left])) {
			t.Fatalf("iteration %d: count=(%d,%v)", iteration, count, err)
		}
	}
}

func FuzzCuddAnalysis(f *testing.F) {
	f.Add(uint8(4), uint64(0x6996), uint64(0xaffe), uint8(3))
	f.Add(uint8(1), uint64(2), uint64(1), uint8(1))
	f.Fuzz(func(t *testing.T, rawVariables uint8, leftTable, rightTable uint64, requested uint8) {
		variables := int(rawVariables%4) + 1
		count := 1 << uint(variables)
		mask := uint64(1)<<uint(count) - 1
		leftTable &= mask
		rightTable &= mask
		manager := cuddAnalysisManager(t, variables)
		left := cuddAnalysisFromTable(t, manager, leftTable)
		right := cuddAnalysisFromTable(t, manager, rightTable)
		correlation, err := manager.Correlation(left, right, nil, ExtraBDDOptions{})
		if err != nil {
			t.Fatal(err)
		}
		wantEqual := bits.OnesCount64(^(leftTable ^ rightTable) & mask)
		if math.Abs(correlation-float64(wantEqual)/float64(count)) > 1e-12 {
			t.Fatalf("correlation=%g want %g", correlation, float64(wantEqual)/float64(count))
		}
		available := bits.OnesCount64(leftTable)
		want := int(requested) % (available + 1)
		split, err := manager.SplitSet(left, uint64(want), ExtraBDDOptions{})
		if err != nil {
			t.Fatal(err)
		}
		if got, _ := manager.SatCount(split, ExtraBDDOptions{}); got != uint64(want) {
			t.Fatalf("split count=%d want %d", got, want)
		}
		if subset, _ := manager.Leq(split, left, ExtraBDDOptions{}); !subset {
			t.Fatal("split is not a subset")
		}
	})
}
