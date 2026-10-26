package abc

// Stand-alone truth-table, lifecycle, and fuzz tests for extraUtilMult.c.

import (
	"errors"
	"strconv"
	"strings"
	"testing"
)

func TestExtraBDDBooleanCanonicalExampleAndFormatting(t *testing.T) {
	manager, err := NewExtraBDDManager(4, 100)
	if err != nil {
		t.Fatal(err)
	}
	example, err := manager.Example(ExtraBDDOptions{})
	if err != nil {
		t.Fatal(err)
	}
	for assignment := 0; assignment < 16; assignment++ {
		values := make([]bool, 4)
		for variable := range values {
			values[variable] = assignment&(1<<uint(variable)) != 0
		}
		got, evalErr := manager.Evaluate(example.Root, values)
		want := values[0] && values[1] || values[2] && values[3]
		if evalErr != nil || got != want {
			t.Fatal(assignment, got, want, evalErr)
		}
	}
	wantText := "BDD " + strconv.FormatUint(uint64(example.Root), 10) + " = -0+2+3 +0-1+2+3 +0+1 \n"
	if example.Text != wantText || example.Nodes <= 0 {
		t.Fatal(example.Text, wantText, example.Nodes)
	}
	a, _ := manager.Variable(0)
	b, _ := manager.Variable(1)
	ab, err := manager.And(a, b, ExtraBDDOptions{})
	if err != nil {
		t.Fatal(err)
	}
	ba, err := manager.And(b, a, ExtraBDDOptions{})
	if err != nil || ba != ab {
		t.Fatal(ab, ba, err)
	}
	contradiction, err := manager.And(a, a.Not(), ExtraBDDOptions{})
	if err != nil || contradiction != ExtraBDDFalse {
		t.Fatal(contradiction, err)
	}
	tautology, err := manager.Or(a, a.Not(), ExtraBDDOptions{})
	if err != nil || tautology != ExtraBDDTrue {
		t.Fatal(tautology, err)
	}
	stats, _ := manager.Statistics()
	if stats.Variables != 4 || stats.Nodes <= 5 || stats.CacheLookups == 0 || stats.CacheMisses == 0 || !strings.HasPrefix(stats.Format(), "BDD stats: Var = 4") {
		t.Fatal(stats, stats.Format())
	}
}

func TestExtraBDDCountsOwnershipAndLifecycle(t *testing.T) {
	manager, _ := NewExtraBDDManager(3, 100)
	a, _ := manager.Variable(0)
	b, _ := manager.Variable(1)
	ab, _ := manager.And(a, b, ExtraBDDOptions{})
	shared, err := manager.CountSharedNodes([]ExtraBDDLit{ab, b}, ExtraBDDOptions{})
	separate, separateErr := manager.CountSeparateNodes([]ExtraBDDLit{ab, b}, ExtraBDDOptions{})
	if err != nil || separateErr != nil || shared != 2 || separate != 3 {
		t.Fatal(shared, separate, err, separateErr)
	}
	count, err := manager.CountNodes(ab, ExtraBDDOptions{})
	if err != nil || count != 2 {
		t.Fatal(count, err)
	}
	stats, err := manager.Close()
	if err != nil || stats.Nodes == 0 {
		t.Fatal(stats, err)
	}
	if _, err = manager.Statistics(); !errors.Is(err, ErrInvariant) {
		t.Fatal(err)
	}
	if _, err = manager.Close(); !errors.Is(err, ErrInvariant) {
		t.Fatal(err)
	}
}

func TestExtraBDDAtomicErrorsLimitsAndCancellation(t *testing.T) {
	for _, row := range [][2]int{{-1, 2}, {256, 300}, {2, 2}} {
		if _, err := NewExtraBDDManager(row[0], row[1]); err == nil {
			t.Fatal(row)
		}
	}
	manager, _ := NewExtraBDDManager(2, 3)
	a, _ := manager.Variable(0)
	b, _ := manager.Variable(1)
	before, _ := manager.Statistics()
	if _, err := manager.And(a, b, ExtraBDDOptions{}); !errors.Is(err, ErrLimit) {
		t.Fatal(err)
	}
	after, _ := manager.Statistics()
	if before != after {
		t.Fatal("failed transaction mutated manager", before, after)
	}
	roomy, _ := NewExtraBDDManager(3, 100)
	x, _ := roomy.Variable(0)
	y, _ := roomy.Variable(1)
	before, _ = roomy.Statistics()
	if _, err := roomy.And(x, y, ExtraBDDOptions{Stop: func() bool { return true }}); !errors.Is(err, ErrCancelled) {
		t.Fatal(err)
	}
	after, _ = roomy.Statistics()
	if before != after {
		t.Fatal("cancelled transaction mutated manager", before, after)
	}
	if _, err := roomy.And(ExtraBDDLit(9999), y, ExtraBDDOptions{}); !errors.Is(err, ErrInvalidLiteral) {
		t.Fatal(err)
	}
	if _, err := roomy.Evaluate(x, []bool{true}); !errors.Is(err, ErrDimension) {
		t.Fatal(err)
	}
	if _, err := roomy.Format(x, ExtraBDDOptions{MaxBytes: 1}); !errors.Is(err, ErrLimit) {
		t.Fatal(err)
	}
	if _, err := roomy.CountNodes(x, ExtraBDDOptions{MaxOperations: 1}); !errors.Is(err, ErrLimit) {
		t.Fatal(err)
	}
}

func TestExtraBDDFromGia(t *testing.T) {
	graph := NewGiaManager()
	a, _ := graph.AppendCI()
	b, _ := graph.AppendCI()
	root, _ := graph.AppendAnd(a, b.Not())
	_, _ = graph.AppendCO(root)
	result, err := BuildExtraBDDFromGia(graph, 100, true, ExtraBDDOptions{})
	if err != nil || len(result.Outputs) != 1 || len(result.Text) != 1 || result.SharedNodes == 0 {
		t.Fatal(result, err)
	}
	for assignment := 0; assignment < 4; assignment++ {
		inputs := []bool{assignment&1 != 0, assignment&2 != 0}
		want, giaErr := graph.Evaluate(inputs)
		got, bddErr := result.Manager.Evaluate(result.Outputs[0], inputs)
		if giaErr != nil || bddErr != nil || got != want[0] {
			t.Fatal(assignment, got, want, giaErr, bddErr)
		}
	}
	extended := NewGiaManager()
	x, _ := extended.AppendCI()
	y, _ := extended.AppendCI()
	xor, _ := extended.AppendXorReal(x, y)
	_, _ = extended.AppendCO(xor)
	if _, err := BuildExtraBDDFromGia(extended, 100, false, ExtraBDDOptions{}); !errors.Is(err, ErrUnsupported) {
		t.Fatal(err)
	}
	if _, err := BuildExtraBDDFromGia(graph, 100, false, ExtraBDDOptions{Stop: func() bool { return true }}); !errors.Is(err, ErrCancelled) {
		t.Fatal(err)
	}
}

func extraBDDVariableTruth(variable int) uint16 {
	var result uint16
	for assignment := 0; assignment < 16; assignment++ {
		if assignment&(1<<uint(variable)) != 0 {
			result |= 1 << uint(assignment)
		}
	}
	return result
}

func FuzzExtraBDD(f *testing.F) {
	f.Add([]byte{0, 1, 0, 2, 3, 1, 4, 0, 2})
	f.Add([]byte{0, 0, 2})
	f.Fuzz(func(t *testing.T, operations []byte) {
		manager, err := NewExtraBDDManager(4, 4096)
		if err != nil {
			t.Fatal(err)
		}
		roots := make([]ExtraBDDLit, 4)
		truths := make([]uint16, 4)
		for variable := range roots {
			roots[variable], _ = manager.Variable(variable)
			truths[variable] = extraBDDVariableTruth(variable)
		}
		for index := 0; index+2 < len(operations) && len(roots) < 64; index += 3 {
			first := int(operations[index]) % len(roots)
			second := int(operations[index+1]) % len(roots)
			var root ExtraBDDLit
			var truth uint16
			switch operations[index+2] % 3 {
			case 0:
				root, err = manager.And(roots[first], roots[second], ExtraBDDOptions{MaxOperations: 1 << 18})
				truth = truths[first] & truths[second]
			case 1:
				root, err = manager.Or(roots[first], roots[second], ExtraBDDOptions{MaxOperations: 1 << 18})
				truth = truths[first] | truths[second]
			case 2:
				root, truth = roots[first].Not(), ^truths[first]
			}
			if errors.Is(err, ErrLimit) {
				return
			}
			if err != nil {
				t.Fatal(index, err)
			}
			roots, truths = append(roots, root), append(truths, truth)
			for assignment := 0; assignment < 16; assignment++ {
				inputs := []bool{assignment&1 != 0, assignment&2 != 0, assignment&4 != 0, assignment&8 != 0}
				got, evalErr := manager.Evaluate(root, inputs)
				want := truth&(1<<uint(assignment)) != 0
				if evalErr != nil || got != want {
					t.Fatal(index, assignment, root, got, want, evalErr)
				}
			}
		}
		if _, err := manager.CountSharedNodes(roots, ExtraBDDOptions{MaxOperations: 1 << 20}); err != nil {
			t.Fatal(err)
		}
	})
}
