package abc

// Stand-alone bridge and textual I/O behavior from src/bdd/cudd/cuddBridge.c,
// src/bdd/cudd/cuddExport.c, src/bdd/cudd/cuddHarwell.c, and
// src/bdd/cudd/cuddRead.c.

import (
	"math"
	"strconv"
	"strings"
)

type CuddDumpFormat uint8

const (
	CuddDumpBLIF CuddDumpFormat = iota
	CuddDumpDOT
	CuddDumpDaVinci
	CuddDumpDDcal
	CuddDumpFactored
)

type CuddIOOptions struct {
	MaxEntries   int
	MaxDimension int
	MaxTextBytes int
	DD           ExtraBDDOptions
}

type CuddMatrixLayout struct {
	RowVariables    []int
	ColumnVariables []int
}

type CuddSparseEntry struct {
	Row, Column int
	Value       float64
}

type CuddSparseMatrix struct {
	Rows, Columns int
	Entries       []CuddSparseEntry
}

type CuddDumpOptions struct {
	Format        CuddDumpFormat
	ModelName     string
	InputNames    []string
	OutputNames   []string
	MaxBytes      int
	MaxNodes      int
	MaxOperations int
	Stop          func() bool
}

type CuddIOSourceCoverage struct {
	Source      string
	Definitions []string
	Disposition string
}

func CuddIOSourceInventory() []CuddIOSourceCoverage {
	return []CuddIOSourceCoverage{
		{
			Source: "cuddBridge.c",
			Definitions: []string{
				"Cudd_addBddThreshold", "Cudd_addBddStrictThreshold", "Cudd_addBddInterval", "Cudd_addBddIthBit", "Cudd_BddToAdd", "Cudd_addBddPattern", "Cudd_bddTransfer", "cuddBddTransfer", "cuddAddBddDoPattern", "addBddDoThreshold", "addBddDoStrictThreshold", "addBddDoInterval", "addBddDoIthBit", "ddBddToAddRecur", "cuddBddTransferRecur",
			},
			Disposition: "All public semantic conversions and transfer are represented by bounded fixed-order operations; manager-local recursive helpers are absorbed. Dynamic reordering restarts and shared-manager ADD/BDD node representation are not reproduced.",
		},
		{
			Source: "cuddExport.c",
			Definitions: []string{
				"Cudd_DumpBlif", "Cudd_DumpBlifBody", "Cudd_DumpDot", "Cudd_DumpDaVinci", "Cudd_DumpDDcal", "Cudd_DumpFactoredForm", "ddDoDumpBlif", "ddDoDumpDaVinci", "ddDoDumpDDcal", "ddDoDumpFactoredForm",
			},
			Disposition: "Deterministic owned BLIF, DOT, DaVinci, DDcal, and factored text is represented with explicit bounds. Byte-identical pointer-derived names, stream failures, BLIF-MV, and native layout conventions are not reproduced.",
		},
		{
			Source:      "cuddHarwell.c",
			Definitions: []string{"Cudd_addHarwell"},
			Disposition: "Real unsymmetric assembled matrices without right-hand sides are parsed and built. Fixed-width FORTRAN descriptors, right-hand sides, diagnostics, and other matrix classes remain unsupported.",
		},
		{
			Source:      "cuddRead.c",
			Definitions: []string{"Cudd_addRead", "Cudd_bddRead"},
			Disposition: "Weighted and unweighted sparse row-column lists, MSB-first explicit layouts, duplicate updates, limits, and rollback are represented with owned strings and matrices. Caller-owned variable-array extension and first-index/stride creation remain outside the owned API.",
		},
	}
}

func normalizeCuddIOOptions(options CuddIOOptions) (CuddIOOptions, error) {
	if options.MaxEntries < 0 || options.MaxDimension < 0 || options.MaxTextBytes < 0 {
		return options, ErrDimension
	}
	if options.MaxEntries == 0 {
		options.MaxEntries = 1 << 20
	}
	if options.MaxDimension == 0 {
		options.MaxDimension = 1 << 20
	}
	if options.MaxTextBytes == 0 {
		options.MaxTextBytes = 16 << 20
	}
	var err error
	options.DD, err = normalizeExtraBDDOptions(options.DD)
	return options, err
}

func cuddMatrixCell(row, column, columns int) (int, error) {
	if row < 0 || column < 0 || columns <= 0 || column >= columns || row > (int(^uint(0)>>1)-column)/columns {
		return 0, ErrDimension
	}
	return row*columns + column, nil
}

func cuddBitsForSize(size int) (int, error) {
	if size <= 0 {
		return 0, ErrDimension
	}
	bits := 0
	for value := size - 1; value != 0; value >>= 1 {
		bits++
	}
	return bits, nil
}

func cuddValidateLayout(variables, rows, columns int, layout CuddMatrixLayout) error {
	rowBits, err := cuddBitsForSize(rows)
	if err != nil {
		return err
	}
	columnBits, err := cuddBitsForSize(columns)
	if err != nil {
		return err
	}
	if len(layout.RowVariables) != rowBits || len(layout.ColumnVariables) != columnBits {
		return ErrDimension
	}
	used := make([]bool, variables)
	for _, list := range [][]int{layout.RowVariables, layout.ColumnVariables} {
		for _, variable := range list {
			if variable < 0 || variable >= variables || used[variable] {
				return ErrDimension
			}
			used[variable] = true
		}
	}
	return nil
}

func cuddMatrixIndex(assignment int, variables []int) int {
	result := 0
	for _, variable := range variables {
		result <<= 1
		if assignment&(1<<uint(variable)) != 0 {
			result++
		}
	}
	return result
}

func cuddBDDEvaluateBounded(manager *ExtraBDDManager, root ExtraBDDLit, assignment []bool, step func() error) (bool, error) {
	literal := root
	for depth := 0; depth <= len(manager.nodes); depth++ {
		if literal < 2 {
			return literal == ExtraBDDTrue, nil
		}
		if !manager.valid(literal) {
			return false, ErrInvalidLiteral
		}
		if err := step(); err != nil {
			return false, err
		}
		node := manager.nodes[literal.ID()]
		if int(node.variable) >= len(assignment) || !manager.valid(node.then) || !manager.valid(node.otherwise) {
			return false, ErrInvariant
		}
		if assignment[node.variable] {
			literal = manager.thenOf(literal)
		} else {
			literal = manager.elseOf(literal)
		}
	}
	return false, ErrInvariant
}

func ParseCuddSparseMatrix(text string, weighted bool, options CuddIOOptions) (CuddSparseMatrix, error) {
	options, err := normalizeCuddIOOptions(options)
	if err != nil {
		return CuddSparseMatrix{}, err
	}
	if len(text) > options.MaxTextBytes {
		return CuddSparseMatrix{}, ErrLimit
	}
	fields := strings.Fields(text)
	if len(fields) < 2 {
		return CuddSparseMatrix{}, ErrFormat
	}
	rows, err := strconv.Atoi(fields[0])
	if err != nil || rows <= 0 {
		return CuddSparseMatrix{}, ErrFormat
	}
	columns, err := strconv.Atoi(fields[1])
	if err != nil || columns <= 0 {
		return CuddSparseMatrix{}, ErrFormat
	}
	if rows > options.MaxDimension || columns > options.MaxDimension {
		return CuddSparseMatrix{}, ErrLimit
	}
	width := 2
	if weighted {
		width = 3
	}
	if (len(fields)-2)%width != 0 || (len(fields)-2)/width > options.MaxEntries {
		if (len(fields)-2)/width > options.MaxEntries {
			return CuddSparseMatrix{}, ErrLimit
		}
		return CuddSparseMatrix{}, ErrFormat
	}
	entryCount := (len(fields) - 2) / width
	if int64(entryCount)*32 > int64(options.DD.MaxBytes) {
		return CuddSparseMatrix{}, ErrLimit
	}
	entries := make([]CuddSparseEntry, 0, entryCount)
	for offset := 2; offset < len(fields); offset += width {
		if len(entries) >= options.DD.MaxOperations {
			return CuddSparseMatrix{}, ErrLimit
		}
		if options.DD.Stop != nil && options.DD.Stop() {
			return CuddSparseMatrix{}, ErrCancelled
		}
		row, rowErr := strconv.Atoi(fields[offset])
		column, columnErr := strconv.Atoi(fields[offset+1])
		value := 1.0
		if weighted {
			value, err = strconv.ParseFloat(fields[offset+2], 64)
		}
		if rowErr != nil || columnErr != nil || err != nil || row < 0 || row >= rows || column < 0 || column >= columns || math.IsNaN(value) {
			return CuddSparseMatrix{}, ErrFormat
		}
		entries = append(entries, CuddSparseEntry{Row: row, Column: column, Value: value})
	}
	return CuddSparseMatrix{Rows: rows, Columns: columns, Entries: entries}, nil
}

func ParseCuddHarwellMatrix(text string, options CuddIOOptions) (CuddSparseMatrix, error) {
	options, err := normalizeCuddIOOptions(options)
	if err != nil {
		return CuddSparseMatrix{}, err
	}
	if len(text) > options.MaxTextBytes {
		return CuddSparseMatrix{}, ErrLimit
	}
	lines := strings.Split(strings.ReplaceAll(text, "\r\n", "\n"), "\n")
	if len(lines) < 5 {
		return CuddSparseMatrix{}, ErrFormat
	}
	counts := strings.Fields(lines[1])
	dimensions := strings.Fields(lines[2])
	if len(counts) != 5 || len(dimensions) != 5 || dimensions[0] != "RUA" {
		return CuddSparseMatrix{}, ErrFormat
	}
	rhsCards, err := strconv.Atoi(counts[4])
	if err != nil || rhsCards != 0 {
		return CuddSparseMatrix{}, ErrFormat
	}
	rows, rowErr := strconv.Atoi(dimensions[1])
	columns, columnErr := strconv.Atoi(dimensions[2])
	nonzero, nonzeroErr := strconv.Atoi(dimensions[3])
	elemental, elementalErr := strconv.Atoi(dimensions[4])
	if rowErr != nil || columnErr != nil || nonzeroErr != nil || elementalErr != nil || rows <= 0 || columns <= 0 || nonzero < 0 || nonzero > options.MaxEntries || elemental != 0 {
		if nonzero > options.MaxEntries {
			return CuddSparseMatrix{}, ErrLimit
		}
		return CuddSparseMatrix{}, ErrFormat
	}
	if rows > options.MaxDimension || columns > options.MaxDimension {
		return CuddSparseMatrix{}, ErrLimit
	}
	if int64(columns+1)*8+int64(nonzero)*48 > int64(options.DD.MaxBytes) {
		return CuddSparseMatrix{}, ErrLimit
	}
	tokens := strings.Fields(strings.Join(lines[4:], " "))
	needed := columns + 1 + nonzero + nonzero
	if len(tokens) != needed {
		return CuddSparseMatrix{}, ErrFormat
	}
	pointers := make([]int, columns+1)
	for index := range pointers {
		pointers[index], err = strconv.Atoi(tokens[index])
		if err != nil || pointers[index] < 1 || pointers[index] > nonzero+1 {
			return CuddSparseMatrix{}, ErrFormat
		}
	}
	if pointers[0] != 1 || pointers[columns] != nonzero+1 {
		return CuddSparseMatrix{}, ErrFormat
	}
	entries := make([]CuddSparseEntry, 0, nonzero)
	positions := make(map[int]int, nonzero)
	operations := 0
	rowBase := columns + 1
	valueBase := rowBase + nonzero
	for column := 0; column < columns; column++ {
		if pointers[column] > pointers[column+1] {
			return CuddSparseMatrix{}, ErrFormat
		}
		for position := pointers[column] - 1; position < pointers[column+1]-1; position++ {
			operations++
			if operations > options.DD.MaxOperations {
				return CuddSparseMatrix{}, ErrLimit
			}
			if options.DD.Stop != nil && options.DD.Stop() {
				return CuddSparseMatrix{}, ErrCancelled
			}
			row, parseErr := strconv.Atoi(tokens[rowBase+position])
			value, valueErr := strconv.ParseFloat(strings.ReplaceAll(strings.ReplaceAll(tokens[valueBase+position], "D", "E"), "d", "e"), 64)
			if parseErr != nil || valueErr != nil || row < 1 || row > rows || math.IsNaN(value) {
				return CuddSparseMatrix{}, ErrFormat
			}
			cell, cellErr := cuddMatrixCell(row-1, column, columns)
			if cellErr != nil {
				return CuddSparseMatrix{}, cellErr
			}
			if entry, ok := positions[cell]; ok {
				entries[entry].Value += value
			} else {
				positions[cell] = len(entries)
				entries = append(entries, CuddSparseEntry{Row: row - 1, Column: column, Value: value})
			}
		}
	}
	return CuddSparseMatrix{Rows: rows, Columns: columns, Entries: entries}, nil
}

func (matrix CuddSparseMatrix) BuildBDD(manager *ExtraBDDManager, layout CuddMatrixLayout, options CuddIOOptions) (ExtraBDDLit, error) {
	options, err := normalizeCuddIOOptions(options)
	if err != nil {
		return 0, err
	}
	if err = manager.check(); err != nil {
		return 0, err
	}
	if err = cuddValidateLayout(manager.variables, matrix.Rows, matrix.Columns, layout); err != nil || len(matrix.Entries) > options.MaxEntries {
		if len(matrix.Entries) > options.MaxEntries {
			return 0, ErrLimit
		}
		return 0, err
	}
	return manager.analysisRun(nil, 1, options.DD, func(count int, work *extraBDDWork) (ExtraBDDLit, error) {
		if int64(len(matrix.Entries))*48+int64(count) > int64(options.DD.MaxBytes) {
			return 0, ErrLimit
		}
		cells := make(map[int]bool, len(matrix.Entries))
		for _, entry := range matrix.Entries {
			if err := work.step(); err != nil {
				return 0, err
			}
			if entry.Row < 0 || entry.Row >= matrix.Rows || entry.Column < 0 || entry.Column >= matrix.Columns {
				return 0, ErrDimension
			}
			cell, cellErr := cuddMatrixCell(entry.Row, entry.Column, matrix.Columns)
			if cellErr != nil {
				return 0, cellErr
			}
			cells[cell] = entry.Value != 0
		}
		values := make([]bool, count)
		for assignment := range count {
			if err := work.step(); err != nil {
				return 0, err
			}
			cell, cellErr := cuddMatrixCell(cuddMatrixIndex(assignment, layout.RowVariables), cuddMatrixIndex(assignment, layout.ColumnVariables), matrix.Columns)
			if cellErr != nil {
				return 0, cellErr
			}
			values[assignment] = cells[cell]
		}
		return manager.analysisBuild(values, work)
	})
}

func (matrix CuddSparseMatrix) BuildADD(manager *CuddADDManager, layout CuddMatrixLayout, background float64, options CuddIOOptions) (CuddADDEdge, error) {
	options, err := normalizeCuddIOOptions(options)
	if err != nil {
		return 0, err
	}
	if err = manager.check(); err != nil {
		return 0, err
	}
	if len(matrix.Entries) > options.MaxEntries {
		return 0, ErrLimit
	}
	if math.IsNaN(background) {
		return 0, ErrDimension
	}
	if err = cuddValidateLayout(manager.variables, matrix.Rows, matrix.Columns, layout); err != nil {
		return 0, err
	}
	count := 1 << uint(manager.variables)
	if int64(count)*8+int64(len(matrix.Entries))*48 > int64(options.DD.MaxBytes) {
		return 0, ErrLimit
	}
	values := make([]float64, count)
	for index := range values {
		values[index] = background
	}
	operations := 0
	cells := make(map[int]float64, len(matrix.Entries))
	for _, entry := range matrix.Entries {
		if entry.Row < 0 || entry.Row >= matrix.Rows || entry.Column < 0 || entry.Column >= matrix.Columns || math.IsNaN(entry.Value) {
			return 0, ErrDimension
		}
		operations++
		if operations > options.DD.MaxOperations {
			return 0, ErrLimit
		}
		if options.DD.Stop != nil && options.DD.Stop() {
			return 0, ErrCancelled
		}
		cell, cellErr := cuddMatrixCell(entry.Row, entry.Column, matrix.Columns)
		if cellErr != nil {
			return 0, cellErr
		}
		cells[cell] = entry.Value
	}
	for assignment := range count {
		operations++
		if operations > options.DD.MaxOperations {
			return 0, ErrLimit
		}
		if options.DD.Stop != nil && options.DD.Stop() {
			return 0, ErrCancelled
		}
		cell, cellErr := cuddMatrixCell(cuddMatrixIndex(assignment, layout.RowVariables), cuddMatrixIndex(assignment, layout.ColumnVariables), matrix.Columns)
		if cellErr != nil {
			return 0, cellErr
		}
		if value, ok := cells[cell]; ok {
			values[assignment] = value
		}
	}
	remaining := options.DD
	remaining.MaxOperations -= operations
	if remaining.MaxOperations <= 0 {
		return 0, ErrLimit
	}
	return manager.FromDense(values, remaining)
}

func CuddADDToBDD(add *CuddADDManager, edge CuddADDEdge, bdd *ExtraBDDManager, predicate func(float64) bool, options ExtraBDDOptions) (ExtraBDDLit, error) {
	if predicate == nil || add == nil || bdd == nil || add.variables != bdd.variables {
		return 0, ErrDimension
	}
	if err := add.checkEdges(edge); err != nil {
		return 0, err
	}
	return bdd.analysisRun(nil, 1, options, func(count int, work *extraBDDWork) (ExtraBDDLit, error) {
		values := make([]bool, count)
		for assignment := range count {
			if err := work.step(); err != nil {
				return 0, err
			}
			values[assignment] = predicate(add.evaluateUnchecked(edge, uint64(assignment)))
		}
		return bdd.analysisBuild(values, work)
	})
}

func CuddADDThresholdBDD(add *CuddADDManager, edge CuddADDEdge, bdd *ExtraBDDManager, threshold float64, strict bool, options ExtraBDDOptions) (ExtraBDDLit, error) {
	if math.IsNaN(threshold) {
		return 0, ErrDimension
	}
	return CuddADDToBDD(add, edge, bdd, func(value float64) bool {
		if strict {
			return value > threshold
		}
		return value >= threshold
	}, options)
}

func CuddADDIntervalBDD(add *CuddADDManager, edge CuddADDEdge, bdd *ExtraBDDManager, lower, upper float64, options ExtraBDDOptions) (ExtraBDDLit, error) {
	if math.IsNaN(lower) || math.IsNaN(upper) || lower > upper {
		return 0, ErrDimension
	}
	return CuddADDToBDD(add, edge, bdd, func(value float64) bool { return value >= lower && value <= upper }, options)
}

func CuddADDPatternBDD(add *CuddADDManager, edge CuddADDEdge, bdd *ExtraBDDManager, options ExtraBDDOptions) (ExtraBDDLit, error) {
	return CuddADDToBDD(add, edge, bdd, func(value float64) bool { return value != 0 }, options)
}

func CuddADDBitBDD(add *CuddADDManager, edge CuddADDEdge, bdd *ExtraBDDManager, bit int, options ExtraBDDOptions) (ExtraBDDLit, error) {
	if bit < 0 || bit >= 63 {
		return 0, ErrDimension
	}
	return CuddADDToBDD(add, edge, bdd, func(value float64) bool {
		if math.IsInf(value, 0) || value < math.MinInt64 || value > math.MaxInt64 {
			return false
		}
		return uint64(int64(value))&(uint64(1)<<uint(bit)) != 0
	}, options)
}

func CuddBDDToADD(bdd *ExtraBDDManager, root ExtraBDDLit, add *CuddADDManager, options ExtraBDDOptions) (CuddADDEdge, error) {
	if bdd == nil || add == nil || bdd.variables != add.variables {
		return 0, ErrDimension
	}
	if err := bdd.checkLiteral(root); err != nil {
		return 0, err
	}
	if err := add.check(); err != nil {
		return 0, err
	}
	normalized, err := normalizeExtraBDDOptions(options)
	if err != nil {
		return 0, err
	}
	count := 1 << uint(bdd.variables)
	if count > normalized.MaxBytes/8 {
		return 0, ErrLimit
	}
	values := make([]float64, count)
	assignment := make([]bool, bdd.variables)
	operations := 0
	step := func() error {
		operations++
		if operations > normalized.MaxOperations {
			return ErrLimit
		}
		if normalized.Stop != nil && normalized.Stop() {
			return ErrCancelled
		}
		return nil
	}
	for index := range count {
		if err := step(); err != nil {
			return 0, err
		}
		for variable := range assignment {
			assignment[variable] = index&(1<<uint(variable)) != 0
		}
		value, evaluateErr := cuddBDDEvaluateBounded(bdd, root, assignment, step)
		if evaluateErr != nil {
			return 0, evaluateErr
		}
		if value {
			values[index] = 1
		}
	}
	normalized.MaxOperations -= operations
	if normalized.MaxOperations <= 0 {
		return 0, ErrLimit
	}
	return add.FromDense(values, normalized)
}

func CuddBDDTransfer(source *ExtraBDDManager, root ExtraBDDLit, destination *ExtraBDDManager, options ExtraBDDOptions) (ExtraBDDLit, error) {
	if source == nil || destination == nil || source.variables != destination.variables {
		return 0, ErrDimension
	}
	if err := source.checkLiteral(root); err != nil {
		return 0, err
	}
	return destination.analysisRun(nil, 1, options, func(count int, work *extraBDDWork) (ExtraBDDLit, error) {
		values := make([]bool, count)
		assignment := make([]bool, source.variables)
		for index := range count {
			if err := work.step(); err != nil {
				return 0, err
			}
			for variable := range assignment {
				assignment[variable] = index&(1<<uint(variable)) != 0
			}
			value, err := cuddBDDEvaluateBounded(source, root, assignment, work.step)
			if err != nil {
				return 0, err
			}
			values[index] = value
		}
		return destination.analysisBuild(values, work)
	})
}

func cuddSafeName(name string) bool {
	if name == "" || !((name[0] >= 'A' && name[0] <= 'Z') || (name[0] >= 'a' && name[0] <= 'z') || name[0] == '_') {
		return false
	}
	for index := 1; index < len(name); index++ {
		character := name[index]
		if !((character >= 'A' && character <= 'Z') || (character >= 'a' && character <= 'z') || (character >= '0' && character <= '9') || character == '_') {
			return false
		}
	}
	return true
}

func CuddDumpBDDs(manager *ExtraBDDManager, roots []ExtraBDDLit, options CuddDumpOptions) (string, error) {
	if err := manager.check(); err != nil {
		return "", err
	}
	if len(roots) == 0 || options.Format > CuddDumpFactored || options.MaxBytes < 0 || options.MaxNodes < 0 || options.MaxOperations < 0 {
		return "", ErrDimension
	}
	if options.MaxBytes == 0 {
		options.MaxBytes = 16 << 20
	}
	if options.MaxNodes == 0 {
		options.MaxNodes = 1 << 20
	}
	if options.MaxOperations == 0 {
		options.MaxOperations = 1 << 20
	}
	for _, root := range roots {
		if !manager.valid(root) {
			return "", ErrInvalidLiteral
		}
	}
	if options.ModelName == "" {
		options.ModelName = "DD"
	}
	if !cuddSafeName(options.ModelName) || (len(options.InputNames) != 0 && len(options.InputNames) != manager.variables) || (len(options.OutputNames) != 0 && len(options.OutputNames) != len(roots)) {
		return "", ErrDimension
	}
	inputName := func(variable int) string {
		if len(options.InputNames) != 0 {
			return options.InputNames[variable]
		}
		return "x" + strconv.Itoa(variable)
	}
	outputName := func(index int) string {
		if len(options.OutputNames) != 0 {
			return options.OutputNames[index]
		}
		return "f" + strconv.Itoa(index)
	}
	for _, names := range [][]string{options.InputNames, options.OutputNames} {
		for _, name := range names {
			if !cuddSafeName(name) {
				return "", ErrDimension
			}
		}
	}
	states := make([]uint8, len(manager.nodes))
	order := make([]uint32, 0)
	visitedNodes := 0
	operations := 0
	step := func() error {
		operations++
		if operations > options.MaxOperations {
			return ErrLimit
		}
		if options.Stop != nil && options.Stop() {
			return ErrCancelled
		}
		return nil
	}
	var visit func(ExtraBDDLit) error
	visit = func(literal ExtraBDDLit) error {
		if literal < 2 {
			return nil
		}
		id := literal.ID()
		if uint64(id) >= uint64(len(manager.nodes)) {
			return ErrInvalidLiteral
		}
		if states[id] == 2 {
			return nil
		}
		if states[id] == 1 {
			return ErrInvariant
		}
		if err := step(); err != nil {
			return err
		}
		if visitedNodes >= options.MaxNodes {
			return ErrLimit
		}
		visitedNodes++
		states[id] = 1
		node := manager.nodes[id]
		if int(node.variable) >= manager.variables || !manager.valid(node.then) || !manager.valid(node.otherwise) {
			return ErrInvariant
		}
		for _, child := range []ExtraBDDLit{node.then, node.otherwise} {
			if child >= 2 && manager.nodes[child.ID()].variable <= node.variable {
				return ErrInvariant
			}
		}
		if err := visit(node.then); err != nil {
			return err
		}
		if err := visit(node.otherwise); err != nil {
			return err
		}
		states[id] = 2
		order = append(order, id)
		return nil
	}
	for _, root := range roots {
		if err := visit(root); err != nil {
			return "", err
		}
	}
	var builder strings.Builder
	appendText := func(text string) error {
		if builder.Len()+len(text) > options.MaxBytes {
			return ErrLimit
		}
		builder.WriteString(text)
		return nil
	}
	edgeName := func(literal ExtraBDDLit) string {
		if literal == ExtraBDDFalse {
			return "0"
		}
		if literal == ExtraBDDTrue {
			return "1"
		}
		name := "n" + strconv.FormatUint(uint64(literal.ID()), 10)
		if literal.Complemented() {
			name = "!" + name
		}
		return name
	}
	switch options.Format {
	case CuddDumpBLIF:
		if err := appendText(".model " + options.ModelName + "\n.inputs"); err != nil {
			return "", err
		}
		for variable := 0; variable < manager.variables; variable++ {
			if err := appendText(" " + inputName(variable)); err != nil {
				return "", err
			}
		}
		if err := appendText("\n.outputs"); err != nil {
			return "", err
		}
		for index := range roots {
			if err := appendText(" " + outputName(index)); err != nil {
				return "", err
			}
		}
		if err := appendText("\n"); err != nil {
			return "", err
		}
		if manager.variables >= 63 {
			return "", ErrLimit
		}
		assignment := make([]bool, manager.variables)
		for index, root := range roots {
			line := ".names"
			for variable := 0; variable < manager.variables; variable++ {
				line += " " + inputName(variable)
			}
			if err := appendText(line + " " + outputName(index) + "\n"); err != nil {
				return "", err
			}
			for value := uint64(0); value < uint64(1)<<uint(manager.variables); value++ {
				if err := step(); err != nil {
					return "", err
				}
				for variable := range assignment {
					assignment[variable] = value&(uint64(1)<<uint(variable)) != 0
				}
				set, evaluateErr := manager.Evaluate(root, assignment)
				if evaluateErr != nil {
					return "", evaluateErr
				}
				if !set {
					continue
				}
				cube := make([]byte, manager.variables)
				for variable := range assignment {
					if assignment[variable] {
						cube[variable] = '1'
					} else {
						cube[variable] = '0'
					}
				}
				row := string(cube) + " 1\n"
				if len(cube) == 0 {
					row = "1\n"
				}
				if err := appendText(row); err != nil {
					return "", err
				}
			}
		}
		if err := appendText(".end\n"); err != nil {
			return "", err
		}
	case CuddDumpDOT:
		if err := appendText("digraph \"" + options.ModelName + "\" {\n  0 [shape=box]; 1 [shape=box];\n"); err != nil {
			return "", err
		}
		for _, id := range order {
			node := manager.nodes[id]
			line := "  n" + strconv.FormatUint(uint64(id), 10) + " [label=\"" + inputName(int(node.variable)) + "\"];\n"
			highMark := ""
			if node.then.Complemented() {
				highMark = ",color=red,arrowhead=odot"
			}
			lowMark := ""
			if node.otherwise.Complemented() {
				lowMark = ",color=red,arrowhead=odot"
			}
			line += "  n" + strconv.FormatUint(uint64(id), 10) + " -> " + strings.TrimPrefix(edgeName(node.then), "!") + " [label=\"1\"" + highMark + "] ;\n"
			line += "  n" + strconv.FormatUint(uint64(id), 10) + " -> " + strings.TrimPrefix(edgeName(node.otherwise), "!") + " [label=\"0\",style=dashed" + lowMark + "];\n"
			if err := appendText(line); err != nil {
				return "", err
			}
		}
		for index, root := range roots {
			mark := ""
			if root.Complemented() {
				mark = " [color=red,arrowhead=odot]"
			}
			if err := appendText("  " + outputName(index) + " [shape=plaintext]; " + outputName(index) + " -> " + strings.TrimPrefix(edgeName(root), "!") + mark + ";\n"); err != nil {
				return "", err
			}
		}
		if err := appendText("}\n"); err != nil {
			return "", err
		}
	case CuddDumpDaVinci, CuddDumpDDcal:
		prefix := "davinci"
		if options.Format == CuddDumpDDcal {
			prefix = "ddcal"
		}
		if err := appendText(prefix + " " + options.ModelName + "\n"); err != nil {
			return "", err
		}
		for _, id := range order {
			node := manager.nodes[id]
			if err := appendText("n" + strconv.FormatUint(uint64(id), 10) + " " + inputName(int(node.variable)) + " " + edgeName(node.then) + " " + edgeName(node.otherwise) + "\n"); err != nil {
				return "", err
			}
		}
		for index, root := range roots {
			if err := appendText(outputName(index) + " " + edgeName(root) + "\n"); err != nil {
				return "", err
			}
		}
	case CuddDumpFactored:
		memo := make(map[ExtraBDDLit]string)
		var expression func(ExtraBDDLit) (string, error)
		expression = func(literal ExtraBDDLit) (string, error) {
			if text, ok := memo[literal]; ok {
				return text, nil
			}
			if literal < 2 {
				return edgeName(literal), nil
			}
			if err := step(); err != nil {
				return "", err
			}
			if literal.Complemented() {
				base, err := expression(literal.Not())
				if err != nil {
					return "", err
				}
				if len(base) > options.MaxBytes-3 {
					return "", ErrLimit
				}
				text := "!(" + base + ")"
				memo[literal] = text
				return text, nil
			}
			node := manager.nodes[literal.ID()]
			high, err := expression(node.then)
			if err != nil {
				return "", err
			}
			low, err := expression(node.otherwise)
			if err != nil {
				return "", err
			}
			name := inputName(int(node.variable))
			needed := int64(10) + 2*int64(len(name)) + int64(len(high)) + int64(len(low))
			if needed > int64(options.MaxBytes) {
				return "", ErrLimit
			}
			text := "((" + name + "&" + high + ")|(!" + name + "&" + low + "))"
			memo[literal] = text
			return text, nil
		}
		for index, root := range roots {
			text, err := expression(root)
			if err != nil {
				return "", err
			}
			if err = appendText(outputName(index) + " = " + text + "\n"); err != nil {
				return "", err
			}
		}
	}
	return builder.String(), nil
}
