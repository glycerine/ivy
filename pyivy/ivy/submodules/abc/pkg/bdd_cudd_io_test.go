package abc

// Stand-alone tests for CUDD bridge and textual I/O adaptations.

import (
	"errors"
	"math/rand"
	"strconv"
	"strings"
	"testing"
)

func cuddIOManagers(t testing.TB, variables int) (*ExtraBDDManager, *CuddADDManager) {
	t.Helper()
	bdd, err := NewExtraBDDManager(variables, 1<<15)
	if err != nil {
		t.Fatal(err)
	}
	add, err := NewCuddADDManager(variables, 1<<15)
	if err != nil {
		t.Fatal(err)
	}
	return bdd, add
}

func cuddIOAssignment(variables int, value int) []bool {
	assignment := make([]bool, variables)
	for variable := range variables {
		assignment[variable] = value&(1<<uint(variable)) != 0
	}
	return assignment
}

func TestCuddIOBridgeExhaustive(t *testing.T) {
	bdd, add := cuddIOManagers(t, 3)
	values := []float64{-2, -1, 0, 0.5, 1, 2, 3, 4}
	edge, err := add.FromDense(values, ExtraBDDOptions{})
	if err != nil {
		t.Fatal(err)
	}
	tests := []struct {
		name string
		make func() (ExtraBDDLit, error)
		want func(float64) bool
	}{
		{"threshold", func() (ExtraBDDLit, error) { return CuddADDThresholdBDD(add, edge, bdd, 1, false, ExtraBDDOptions{}) }, func(value float64) bool { return value >= 1 }},
		{"strict", func() (ExtraBDDLit, error) { return CuddADDThresholdBDD(add, edge, bdd, 1, true, ExtraBDDOptions{}) }, func(value float64) bool { return value > 1 }},
		{"interval", func() (ExtraBDDLit, error) { return CuddADDIntervalBDD(add, edge, bdd, -1, 1, ExtraBDDOptions{}) }, func(value float64) bool { return value >= -1 && value <= 1 }},
		{"pattern", func() (ExtraBDDLit, error) { return CuddADDPatternBDD(add, edge, bdd, ExtraBDDOptions{}) }, func(value float64) bool { return value != 0 }},
		{"bit", func() (ExtraBDDLit, error) { return CuddADDBitBDD(add, edge, bdd, 1, ExtraBDDOptions{}) }, func(value float64) bool { return uint64(int64(value))&2 != 0 }},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			root, makeErr := test.make()
			if makeErr != nil {
				t.Fatal(makeErr)
			}
			for assignment, value := range values {
				got, evaluateErr := bdd.Evaluate(root, cuddIOAssignment(3, assignment))
				if evaluateErr != nil || got != test.want(value) {
					t.Fatalf("assignment %d: got %v, err %v", assignment, got, evaluateErr)
				}
			}
		})
	}
	pattern, err := CuddADDPatternBDD(add, edge, bdd, ExtraBDDOptions{})
	if err != nil {
		t.Fatal(err)
	}
	converted, err := CuddBDDToADD(bdd, pattern, add, ExtraBDDOptions{})
	if err != nil {
		t.Fatal(err)
	}
	dense, err := add.Dense(converted)
	if err != nil {
		t.Fatal(err)
	}
	for index, value := range dense {
		want := 0.0
		if values[index] != 0 {
			want = 1
		}
		if value != want {
			t.Fatalf("converted[%d]=%g, want %g", index, value, want)
		}
	}
	destination, err := NewExtraBDDManager(3, 1<<15)
	if err != nil {
		t.Fatal(err)
	}
	transferred, err := CuddBDDTransfer(bdd, pattern, destination, ExtraBDDOptions{})
	if err != nil {
		t.Fatal(err)
	}
	for assignment := range values {
		got, _ := destination.Evaluate(transferred, cuddIOAssignment(3, assignment))
		if got != (values[assignment] != 0) {
			t.Fatalf("transfer assignment %d", assignment)
		}
	}
}

func TestCuddIOSourceInventory(t *testing.T) {
	inventory := CuddIOSourceInventory()
	wantCounts := []int{15, 10, 1, 2}
	if len(inventory) != len(wantCounts) {
		t.Fatalf("inventory length %d", len(inventory))
	}
	seen := make(map[string]bool)
	for index, source := range inventory {
		if source.Source == "" || source.Disposition == "" || len(source.Definitions) != wantCounts[index] {
			t.Fatalf("source %d: %#v", index, source)
		}
		for _, definition := range source.Definitions {
			if definition == "" || seen[definition] {
				t.Fatalf("definition %q", definition)
			}
			seen[definition] = true
		}
	}
	for _, definition := range []string{"Cudd_addBddThreshold", "cuddBddTransferRecur", "Cudd_DumpFactoredForm", "ddDoDumpBlif", "Cudd_addHarwell", "Cudd_addRead", "Cudd_bddRead"} {
		if !seen[definition] {
			t.Fatalf("missing %s", definition)
		}
	}
}

func TestCuddIOSparseReadAndBuild(t *testing.T) {
	matrix, err := ParseCuddSparseMatrix("3 4\n0 1 2.5\n2 3 -1\n0 1 7\n", true, CuddIOOptions{})
	if err != nil || matrix.Rows != 3 || matrix.Columns != 4 || len(matrix.Entries) != 3 {
		t.Fatalf("parse: %#v, %v", matrix, err)
	}
	bdd, add := cuddIOManagers(t, 4)
	layout := CuddMatrixLayout{RowVariables: []int{0, 1}, ColumnVariables: []int{2, 3}}
	addRoot, err := matrix.BuildADD(add, layout, 0, CuddIOOptions{})
	if err != nil {
		t.Fatal(err)
	}
	bddRoot, err := matrix.BuildBDD(bdd, layout, CuddIOOptions{})
	if err != nil {
		t.Fatal(err)
	}
	for assignment := 0; assignment < 16; assignment++ {
		row := cuddMatrixIndex(assignment, layout.RowVariables)
		column := cuddMatrixIndex(assignment, layout.ColumnVariables)
		want := 0.0
		if row == 0 && column == 1 {
			want = 7
		}
		if row == 2 && column == 3 {
			want = -1
		}
		got, evaluateErr := add.Evaluate(addRoot, uint64(assignment))
		if evaluateErr != nil || got != want {
			t.Fatalf("add row %d column %d: %g, %v", row, column, got, evaluateErr)
		}
		set, evaluateErr := bdd.Evaluate(bddRoot, cuddIOAssignment(4, assignment))
		if evaluateErr != nil || set != (want != 0) {
			t.Fatalf("bdd row %d column %d: %v, %v", row, column, set, evaluateErr)
		}
	}
	unweighted, err := ParseCuddSparseMatrix("2 2 0 1 1 0", false, CuddIOOptions{})
	if err != nil || len(unweighted.Entries) != 2 || unweighted.Entries[0].Value != 1 {
		t.Fatalf("unweighted: %#v, %v", unweighted, err)
	}
}

func TestCuddIOHarwell(t *testing.T) {
	text := "small matrix                                                         KEY00001\n" +
		"4 1 1 1 0\n" +
		"RUA 3 2 3 0\n" +
		"(3I3) (3I3) (3E12.4)\n" +
		"1 3 4\n1 3 2\n2.5 -1.0 4D+0\n"
	matrix, err := ParseCuddHarwellMatrix(text, CuddIOOptions{})
	if err != nil {
		t.Fatal(err)
	}
	want := []CuddSparseEntry{{Row: 0, Column: 0, Value: 2.5}, {Row: 2, Column: 0, Value: -1}, {Row: 1, Column: 1, Value: 4}}
	if matrix.Rows != 3 || matrix.Columns != 2 || len(matrix.Entries) != len(want) {
		t.Fatalf("matrix %#v", matrix)
	}
	for index := range want {
		if matrix.Entries[index] != want[index] {
			t.Fatalf("entry %d: %#v != %#v", index, matrix.Entries[index], want[index])
		}
	}
}

func TestCuddIOHarwellDuplicateCoordinatesAdd(t *testing.T) {
	text := "duplicate matrix                                                     KEY00002\n" +
		"4 1 1 1 0\n" +
		"RUA 2 1 2 0\n" +
		"(2I3) (2I3) (2E12.4)\n" +
		"1 3\n1 1\n2.5 3.5\n"
	matrix, err := ParseCuddHarwellMatrix(text, CuddIOOptions{})
	if err != nil {
		t.Fatal(err)
	}
	if len(matrix.Entries) != 1 || matrix.Entries[0] != (CuddSparseEntry{Row: 0, Column: 0, Value: 6}) {
		t.Fatalf("duplicates were not added: %#v", matrix.Entries)
	}
	_, add := cuddIOManagers(t, 1)
	root, err := matrix.BuildADD(add, CuddMatrixLayout{RowVariables: []int{0}}, 0, CuddIOOptions{})
	if err != nil {
		t.Fatal(err)
	}
	value, err := add.Evaluate(root, 0)
	if err != nil || value != 6 {
		t.Fatalf("built duplicate value %g, %v", value, err)
	}
}

func TestCuddIODumpsDeterministicAndBounded(t *testing.T) {
	bdd, _ := cuddIOManagers(t, 2)
	x, _ := bdd.Variable(0)
	y, _ := bdd.Variable(1)
	root, err := bdd.Xor(x, y, ExtraBDDOptions{})
	if err != nil {
		t.Fatal(err)
	}
	for format := CuddDumpBLIF; format <= CuddDumpFactored; format++ {
		options := CuddDumpOptions{Format: format, ModelName: "xor", InputNames: []string{"a", "b"}, OutputNames: []string{"z"}}
		first, dumpErr := CuddDumpBDDs(bdd, []ExtraBDDLit{root}, options)
		second, secondErr := CuddDumpBDDs(bdd, []ExtraBDDLit{root}, options)
		if dumpErr != nil || secondErr != nil || first == "" || first != second || !strings.Contains(first, "z") {
			t.Fatalf("format %d: %q, %v, %v", format, first, dumpErr, secondErr)
		}
	}
	blif, err := CuddDumpBDDs(bdd, []ExtraBDDLit{root}, CuddDumpOptions{Format: CuddDumpBLIF, ModelName: "xor", InputNames: []string{"a", "b"}, OutputNames: []string{"z"}})
	if err != nil || !strings.Contains(blif, "\n.outputs z\n.names") || !strings.Contains(blif, "01 1\n") || !strings.Contains(blif, "10 1\n") || strings.Contains(blif, "00 1\n") {
		t.Fatalf("blif %q, %v", blif, err)
	}
	if text, err := CuddDumpBDDs(bdd, []ExtraBDDLit{root}, CuddDumpOptions{Format: CuddDumpBLIF, MaxNodes: 1}); text != "" || !errors.Is(err, ErrLimit) {
		t.Fatalf("node limit: %q, %v", text, err)
	}
	if text, err := CuddDumpBDDs(bdd, []ExtraBDDLit{root}, CuddDumpOptions{Format: CuddDumpDOT, MaxBytes: 10}); text != "" || !errors.Is(err, ErrLimit) {
		t.Fatalf("byte limit: %q, %v", text, err)
	}
	if text, err := CuddDumpBDDs(bdd, []ExtraBDDLit{root}, CuddDumpOptions{Format: CuddDumpDOT, Stop: func() bool { return true }}); text != "" || !errors.Is(err, ErrCancelled) {
		t.Fatalf("cancel: %q, %v", text, err)
	}
}

func TestCuddIOErrorsLimitsAndRollback(t *testing.T) {
	badSparse := []string{"", "0 2", "2 2 0", "2 2 -1 0", "2 2 0 2", "2 2 0 x", "2 2 0 0 NaN"}
	for _, text := range badSparse {
		if _, err := ParseCuddSparseMatrix(text, strings.Count(text, " ") > 4, CuddIOOptions{}); err == nil {
			t.Fatalf("accepted %q", text)
		}
	}
	if _, err := ParseCuddSparseMatrix("2 2 0 0 1 1", false, CuddIOOptions{MaxEntries: 1}); !errors.Is(err, ErrLimit) {
		t.Fatalf("entry limit: %v", err)
	}
	if _, err := ParseCuddSparseMatrix("1000001 1", false, CuddIOOptions{MaxDimension: 1000000}); !errors.Is(err, ErrLimit) {
		t.Fatalf("dimension limit: %v", err)
	}
	if _, err := cuddMatrixCell(int(^uint(0)>>1), 1, 2); !errors.Is(err, ErrDimension) {
		t.Fatalf("cell overflow: %v", err)
	}
	bdd, add := cuddIOManagers(t, 4)
	values := make([]float64, 16)
	for index := range values {
		values[index] = float64(index)
	}
	edge, err := add.FromDense(values, ExtraBDDOptions{})
	if err != nil {
		t.Fatal(err)
	}
	before := len(bdd.nodes)
	if root, bridgeErr := CuddADDPatternBDD(add, edge, bdd, ExtraBDDOptions{MaxOperations: 1}); root != 0 || !errors.Is(bridgeErr, ErrLimit) || len(bdd.nodes) != before {
		t.Fatalf("bridge rollback: %v, %v, %d/%d", root, bridgeErr, len(bdd.nodes), before)
	}
	matrix := CuddSparseMatrix{Rows: 4, Columns: 4, Entries: []CuddSparseEntry{{Row: 3, Column: 3, Value: 1}}}
	layout := CuddMatrixLayout{RowVariables: []int{0, 1}, ColumnVariables: []int{2, 3}}
	if root, buildErr := matrix.BuildBDD(bdd, layout, CuddIOOptions{DD: ExtraBDDOptions{MaxOperations: 1}}); root != 0 || !errors.Is(buildErr, ErrLimit) || len(bdd.nodes) != before {
		t.Fatalf("matrix rollback: %v, %v, %d/%d", root, buildErr, len(bdd.nodes), before)
	}
	if _, err := CuddADDBitBDD(add, edge, bdd, 63, ExtraBDDOptions{}); !errors.Is(err, ErrDimension) {
		t.Fatalf("bit: %v", err)
	}
	if _, err := CuddADDIntervalBDD(add, edge, bdd, 2, 1, ExtraBDDOptions{}); !errors.Is(err, ErrDimension) {
		t.Fatalf("interval: %v", err)
	}
}

func TestCuddIODumpWorkNamesAndMalformedGraphs(t *testing.T) {
	large, err := NewExtraBDDManager(30, 64)
	if err != nil {
		t.Fatal(err)
	}
	if text, dumpErr := CuddDumpBDDs(large, []ExtraBDDLit{ExtraBDDFalse}, CuddDumpOptions{Format: CuddDumpBLIF, MaxOperations: 4}); text != "" || !errors.Is(dumpErr, ErrLimit) {
		t.Fatalf("bounded constant dump: %q, %v", text, dumpErr)
	}
	bdd, _ := cuddIOManagers(t, 1)
	variable, err := bdd.Variable(0)
	if err != nil {
		t.Fatal(err)
	}
	if text, dumpErr := CuddDumpBDDs(bdd, []ExtraBDDLit{variable}, CuddDumpOptions{Format: CuddDumpDOT, OutputNames: []string{"bad-name"}}); text != "" || !errors.Is(dumpErr, ErrDimension) {
		t.Fatalf("unsafe name: %q, %v", text, dumpErr)
	}
	bdd.nodes[variable.ID()].then = ExtraBDDLit(len(bdd.nodes)+1) << 1
	if text, dumpErr := CuddDumpBDDs(bdd, []ExtraBDDLit{variable}, CuddDumpOptions{Format: CuddDumpDOT}); text != "" || !errors.Is(dumpErr, ErrInvariant) {
		t.Fatalf("malformed child: %q, %v", text, dumpErr)
	}
}

func TestCuddIOTransferRejectsCycles(t *testing.T) {
	source, _ := cuddIOManagers(t, 1)
	destination, err := NewExtraBDDManager(1, 32)
	if err != nil {
		t.Fatal(err)
	}
	root, err := source.Variable(0)
	if err != nil {
		t.Fatal(err)
	}
	source.nodes[root.ID()].then = root
	if result, transferErr := CuddBDDTransfer(source, root, destination, ExtraBDDOptions{}); result != 0 || !errors.Is(transferErr, ErrInvariant) {
		t.Fatalf("cycle transfer: %v, %v", result, transferErr)
	}
}

func TestCuddIOIndependentRandomModel(t *testing.T) {
	random := rand.New(rand.NewSource(0x51a7))
	for trial := 0; trial < 5000; trial++ {
		variables := 1 + random.Intn(4)
		bdd, add := cuddIOManagers(t, variables)
		values := make([]float64, 1<<uint(variables))
		for index := range values {
			values[index] = float64(random.Intn(9) - 4)
		}
		edge, err := add.FromDense(values, ExtraBDDOptions{})
		if err != nil {
			t.Fatal(err)
		}
		threshold := float64(random.Intn(7) - 3)
		root, err := CuddADDThresholdBDD(add, edge, bdd, threshold, trial&1 != 0, ExtraBDDOptions{})
		if err != nil {
			t.Fatal(err)
		}
		for assignment, value := range values {
			got, evaluateErr := bdd.Evaluate(root, cuddIOAssignment(variables, assignment))
			want := value >= threshold
			if trial&1 != 0 {
				want = value > threshold
			}
			if evaluateErr != nil || got != want {
				t.Fatalf("trial %d assignment %d", trial, assignment)
			}
		}
	}
}

func FuzzCuddIOSparseAndBridge(f *testing.F) {
	for _, seed := range []string{"2 2 0 0", "3 4 0 1 2 3", "1 1", "bad", "2 2 -1 0"} {
		f.Add(seed, uint8(0))
	}
	f.Fuzz(func(t *testing.T, text string, threshold uint8) {
		matrix, err := ParseCuddSparseMatrix(text, false, CuddIOOptions{MaxEntries: 64, MaxTextBytes: 4096, DD: ExtraBDDOptions{MaxOperations: 4096, MaxBytes: 1 << 16}})
		if err != nil {
			if !errors.Is(err, ErrFormat) && !errors.Is(err, ErrLimit) && !errors.Is(err, ErrCancelled) {
				t.Fatal(err)
			}
			return
		}
		if matrix.Rows > 16 || matrix.Columns > 16 {
			return
		}
		rowBits, _ := cuddBitsForSize(matrix.Rows)
		columnBits, _ := cuddBitsForSize(matrix.Columns)
		if rowBits+columnBits > 8 {
			return
		}
		bdd, add := cuddIOManagers(t, rowBits+columnBits)
		layout := CuddMatrixLayout{}
		for index := 0; index < rowBits; index++ {
			layout.RowVariables = append(layout.RowVariables, index)
		}
		for index := 0; index < columnBits; index++ {
			layout.ColumnVariables = append(layout.ColumnVariables, rowBits+index)
		}
		edge, buildErr := matrix.BuildADD(add, layout, 0, CuddIOOptions{MaxEntries: 64, DD: ExtraBDDOptions{MaxOperations: 1 << 16, MaxBytes: 1 << 20}})
		if buildErr != nil {
			return
		}
		root, bridgeErr := CuddADDThresholdBDD(add, edge, bdd, float64(threshold%3), false, ExtraBDDOptions{MaxOperations: 1 << 16, MaxBytes: 1 << 20})
		if bridgeErr != nil {
			return
		}
		textOut, dumpErr := CuddDumpBDDs(bdd, []ExtraBDDLit{root}, CuddDumpOptions{Format: CuddDumpFactored, MaxBytes: 1 << 20})
		if dumpErr != nil || !strings.Contains(textOut, "f0 = ") {
			t.Fatalf("dump %q, %v; input length %s", textOut, dumpErr, strconv.Itoa(len(text)))
		}
	})
}
