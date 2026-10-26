package abc

// Owned sparse matrices from src/misc/espresso/matrix.c and sparse headers.

import (
	"bytes"
	"slices"
	"strconv"
	"strings"
)

type EspressoSparseLimits struct {
	MaxRows, MaxColumns, MaxEntries, MaxBytes int
}

func normalizeEspressoSparseLimits(limits EspressoSparseLimits) (EspressoSparseLimits, error) {
	if limits.MaxRows < 0 || limits.MaxColumns < 0 || limits.MaxEntries < 0 || limits.MaxBytes < 0 {
		return limits, ErrDimension
	}
	if limits.MaxRows == 0 {
		limits.MaxRows = 1 << 20
	}
	if limits.MaxColumns == 0 {
		limits.MaxColumns = 1 << 20
	}
	if limits.MaxEntries == 0 {
		limits.MaxEntries = 16 << 20
	}
	if limits.MaxBytes == 0 {
		limits.MaxBytes = 64 << 20
	}
	return limits, nil
}

type espressoSparseCell struct{ row, column int }

type EspressoSparseMatrix struct {
	Tag               int
	rows              []*EspressoSparseRow
	columns           []*EspressoSparseColumn
	rowIDs, columnIDs []int
	cellTags          map[espressoSparseCell]int
	entries           int
	closed            bool
}

func NewEspressoSparseMatrix() *EspressoSparseMatrix {
	return &EspressoSparseMatrix{cellTags: make(map[espressoSparseCell]int)}
}

func NewEspressoSparseMatrixSize(maxRow, maxColumn int) (*EspressoSparseMatrix, error) {
	matrix := NewEspressoSparseMatrix()
	if err := matrix.Reserve(maxRow, maxColumn); err != nil {
		return nil, err
	}
	return matrix, nil
}

func (matrix *EspressoSparseMatrix) check() error {
	if matrix == nil {
		return ErrDimension
	}
	if matrix.closed || matrix.entries < 0 || matrix.cellTags == nil {
		return ErrInvariant
	}
	return nil
}

func sparseGrowRows(values []*EspressoSparseRow, maximum int) []*EspressoSparseRow {
	if maximum < len(values) {
		return values
	}
	size := max(len(values)*2, maximum+1)
	return append(values, make([]*EspressoSparseRow, size-len(values))...)
}

func sparseGrowColumns(values []*EspressoSparseColumn, maximum int) []*EspressoSparseColumn {
	if maximum < len(values) {
		return values
	}
	size := max(len(values)*2, maximum+1)
	return append(values, make([]*EspressoSparseColumn, size-len(values))...)
}

func (matrix *EspressoSparseMatrix) Reserve(maxRow, maxColumn int) error {
	if err := matrix.check(); err != nil {
		return err
	}
	if maxRow < 0 || maxColumn < 0 {
		return ErrDimension
	}
	matrix.rows = sparseGrowRows(matrix.rows, maxRow)
	matrix.columns = sparseGrowColumns(matrix.columns, maxColumn)
	return nil
}

func (matrix *EspressoSparseMatrix) Close() error {
	if err := matrix.check(); err != nil {
		return err
	}
	matrix.rows, matrix.columns = nil, nil
	matrix.rowIDs, matrix.columnIDs = nil, nil
	matrix.cellTags = nil
	matrix.entries = 0
	matrix.closed = true
	return nil
}

func (matrix *EspressoSparseMatrix) Clone() (*EspressoSparseMatrix, error) {
	if err := matrix.check(); err != nil {
		return nil, err
	}
	result := NewEspressoSparseMatrix()
	result.Tag = matrix.Tag
	result.rows = make([]*EspressoSparseRow, len(matrix.rows))
	result.columns = make([]*EspressoSparseColumn, len(matrix.columns))
	result.rowIDs = append(result.rowIDs, matrix.rowIDs...)
	result.columnIDs = append(result.columnIDs, matrix.columnIDs...)
	result.entries = matrix.entries
	for _, id := range matrix.rowIDs {
		result.rows[id], _ = matrix.rows[id].Clone()
	}
	for _, id := range matrix.columnIDs {
		result.columns[id], _ = matrix.columns[id].Clone()
	}
	for cell, tag := range matrix.cellTags {
		result.cellTags[cell] = tag
	}
	return result, nil
}

func (matrix *EspressoSparseMatrix) Insert(row, column int) (bool, error) {
	if err := matrix.check(); err != nil {
		return false, err
	}
	if row < 0 || column < 0 {
		return false, ErrDimension
	}
	matrix.rows = sparseGrowRows(matrix.rows, row)
	matrix.columns = sparseGrowColumns(matrix.columns, column)
	if matrix.rows[row] != nil {
		_, found := sparseSortedFind(matrix.rows[row].values, column)
		if found {
			return false, nil
		}
	}
	if matrix.rows[row] == nil {
		matrix.rows[row] = &EspressoSparseRow{Number: row}
		matrix.rowIDs, _, _ = sparseSortedInsert(matrix.rowIDs, row)
	}
	if matrix.columns[column] == nil {
		matrix.columns[column] = &EspressoSparseColumn{Number: column}
		matrix.columnIDs, _, _ = sparseSortedInsert(matrix.columnIDs, column)
	}
	matrix.rows[row].values, _, _ = sparseSortedInsert(matrix.rows[row].values, column)
	matrix.columns[column].values, _, _ = sparseSortedInsert(matrix.columns[column].values, row)
	matrix.entries++
	return true, nil
}

func (matrix *EspressoSparseMatrix) Has(row, column int) (bool, error) {
	if err := matrix.check(); err != nil {
		return false, err
	}
	if row < 0 || column < 0 {
		return false, ErrDimension
	}
	if row >= len(matrix.rows) || matrix.rows[row] == nil {
		return false, nil
	}
	_, found := sparseSortedFind(matrix.rows[row].values, column)
	return found, nil
}

func (matrix *EspressoSparseMatrix) Remove(row, column int) (bool, error) {
	if err := matrix.check(); err != nil {
		return false, err
	}
	if row < 0 || column < 0 {
		return false, ErrDimension
	}
	if row >= len(matrix.rows) || matrix.rows[row] == nil {
		return false, nil
	}
	var removed bool
	matrix.rows[row].values, removed, _ = sparseSortedRemove(matrix.rows[row].values, column)
	if !removed {
		return false, nil
	}
	matrix.columns[column].values, _, _ = sparseSortedRemove(matrix.columns[column].values, row)
	delete(matrix.cellTags, espressoSparseCell{row: row, column: column})
	matrix.entries--
	if len(matrix.rows[row].values) == 0 {
		matrix.rows[row] = nil
		matrix.rowIDs, _, _ = sparseSortedRemove(matrix.rowIDs, row)
	}
	if len(matrix.columns[column].values) == 0 {
		matrix.columns[column] = nil
		matrix.columnIDs, _, _ = sparseSortedRemove(matrix.columnIDs, column)
	}
	return true, nil
}

func (matrix *EspressoSparseMatrix) DeleteRow(row int) (bool, error) {
	if err := matrix.check(); err != nil {
		return false, err
	}
	if row < 0 {
		return false, ErrDimension
	}
	if row >= len(matrix.rows) || matrix.rows[row] == nil {
		return false, nil
	}
	columns := append([]int(nil), matrix.rows[row].values...)
	for _, column := range columns {
		if _, err := matrix.Remove(row, column); err != nil {
			return false, err
		}
	}
	return true, nil
}

func (matrix *EspressoSparseMatrix) DeleteColumn(column int) (bool, error) {
	if err := matrix.check(); err != nil {
		return false, err
	}
	if column < 0 {
		return false, ErrDimension
	}
	if column >= len(matrix.columns) || matrix.columns[column] == nil {
		return false, nil
	}
	rows := append([]int(nil), matrix.columns[column].values...)
	for _, row := range rows {
		if _, err := matrix.Remove(row, column); err != nil {
			return false, err
		}
	}
	return true, nil
}

func (matrix *EspressoSparseMatrix) Row(row int) (*EspressoSparseRow, bool, error) {
	if err := matrix.check(); err != nil {
		return nil, false, err
	}
	if row < 0 {
		return nil, false, ErrDimension
	}
	if row >= len(matrix.rows) || matrix.rows[row] == nil {
		return nil, false, nil
	}
	result, _ := matrix.rows[row].Clone()
	return result, true, nil
}

func (matrix *EspressoSparseMatrix) Column(column int) (*EspressoSparseColumn, bool, error) {
	if err := matrix.check(); err != nil {
		return nil, false, err
	}
	if column < 0 {
		return nil, false, ErrDimension
	}
	if column >= len(matrix.columns) || matrix.columns[column] == nil {
		return nil, false, nil
	}
	result, _ := matrix.columns[column].Clone()
	return result, true, nil
}

func (matrix *EspressoSparseMatrix) RowIDs() ([]int, error) {
	if err := matrix.check(); err != nil {
		return nil, err
	}
	return append([]int(nil), matrix.rowIDs...), nil
}

func (matrix *EspressoSparseMatrix) ColumnIDs() ([]int, error) {
	if err := matrix.check(); err != nil {
		return nil, err
	}
	return append([]int(nil), matrix.columnIDs...), nil
}

func (matrix *EspressoSparseMatrix) Counts() (rows, columns, entries int, err error) {
	if err = matrix.check(); err != nil {
		return 0, 0, 0, err
	}
	return len(matrix.rowIDs), len(matrix.columnIDs), matrix.entries, nil
}

func (matrix *EspressoSparseMatrix) SetCellTag(row, column, tag int) error {
	present, err := matrix.Has(row, column)
	if err != nil {
		return err
	}
	if !present {
		return ErrDimension
	}
	matrix.cellTags[espressoSparseCell{row: row, column: column}] = tag
	return nil
}

func (matrix *EspressoSparseMatrix) CellTag(row, column int) (int, bool, error) {
	present, err := matrix.Has(row, column)
	if err != nil || !present {
		return 0, false, err
	}
	value, tagged := matrix.cellTags[espressoSparseCell{row: row, column: column}]
	return value, tagged, nil
}

func (matrix *EspressoSparseMatrix) SetRowMetadata(row, flag, tag int) error {
	if err := matrix.check(); err != nil {
		return err
	}
	if row < 0 || row >= len(matrix.rows) || matrix.rows[row] == nil {
		return ErrDimension
	}
	matrix.rows[row].Flag, matrix.rows[row].Tag = flag, tag
	return nil
}

func (matrix *EspressoSparseMatrix) SetColumnMetadata(column, flag, tag int) error {
	if err := matrix.check(); err != nil {
		return err
	}
	if column < 0 || column >= len(matrix.columns) || matrix.columns[column] == nil {
		return ErrDimension
	}
	matrix.columns[column].Flag, matrix.columns[column].Tag = flag, tag
	return nil
}

func (matrix *EspressoSparseMatrix) CopyRow(destinationRow int, row *EspressoSparseRow) error {
	if err := matrix.check(); err != nil {
		return err
	}
	if destinationRow < 0 || row == nil {
		return ErrDimension
	}
	for _, column := range row.values {
		if _, err := matrix.Insert(destinationRow, column); err != nil {
			return err
		}
	}
	return nil
}

// CopyColumnToRow preserves sm_copy_col's transposing argument order.
func (matrix *EspressoSparseMatrix) CopyColumnToRow(destinationRow int, column *EspressoSparseColumn) error {
	if err := matrix.check(); err != nil {
		return err
	}
	if destinationRow < 0 || column == nil {
		return ErrDimension
	}
	for _, sourceRow := range column.values {
		if _, err := matrix.Insert(destinationRow, sourceRow); err != nil {
			return err
		}
	}
	return nil
}

func (matrix *EspressoSparseMatrix) LongestRow() (*EspressoSparseRow, bool, error) {
	if err := matrix.check(); err != nil {
		return nil, false, err
	}
	best := -1
	for _, row := range matrix.rowIDs {
		if best < 0 || len(matrix.rows[row].values) > len(matrix.rows[best].values) {
			best = row
		}
	}
	if best < 0 {
		return nil, false, nil
	}
	result, _ := matrix.rows[best].Clone()
	return result, true, nil
}

func (matrix *EspressoSparseMatrix) LongestColumn() (*EspressoSparseColumn, bool, error) {
	if err := matrix.check(); err != nil {
		return nil, false, err
	}
	best := -1
	for _, column := range matrix.columnIDs {
		if best < 0 || len(matrix.columns[column].values) > len(matrix.columns[best].values) {
			best = column
		}
	}
	if best < 0 {
		return nil, false, nil
	}
	result, _ := matrix.columns[best].Clone()
	return result, true, nil
}

func (matrix *EspressoSparseMatrix) MarshalText() ([]byte, error) {
	if err := matrix.check(); err != nil {
		return nil, err
	}
	var output strings.Builder
	for _, row := range matrix.rowIDs {
		for _, column := range matrix.rows[row].values {
			output.WriteString(strconv.Itoa(row))
			output.WriteByte(' ')
			output.WriteString(strconv.Itoa(column))
			output.WriteByte('\n')
		}
	}
	return []byte(output.String()), nil
}

func ParseEspressoSparseMatrix(data []byte, limits EspressoSparseLimits) (*EspressoSparseMatrix, error) {
	limits, err := normalizeEspressoSparseLimits(limits)
	if err != nil {
		return nil, err
	}
	if len(data) > limits.MaxBytes {
		return nil, ErrLimit
	}
	fields := bytes.Fields(data)
	if len(fields)%2 != 0 {
		return nil, ErrFormat
	}
	matrix := NewEspressoSparseMatrix()
	for index := 0; index < len(fields); index += 2 {
		row, rowErr := strconv.Atoi(string(fields[index]))
		column, columnErr := strconv.Atoi(string(fields[index+1]))
		if rowErr != nil || columnErr != nil || row < 0 || column < 0 {
			_ = matrix.Close()
			return nil, ErrFormat
		}
		if row >= limits.MaxRows || column >= limits.MaxColumns {
			_ = matrix.Close()
			return nil, ErrLimit
		}
		inserted, insertErr := matrix.Insert(row, column)
		if insertErr != nil {
			_ = matrix.Close()
			return nil, insertErr
		}
		if inserted && matrix.entries > limits.MaxEntries {
			_ = matrix.Close()
			return nil, ErrLimit
		}
	}
	return matrix, nil
}

func ParseEspressoCompressedSparseMatrix(data []byte, limits EspressoSparseLimits) (*EspressoSparseMatrix, error) {
	limits, err := normalizeEspressoSparseLimits(limits)
	if err != nil {
		return nil, err
	}
	if len(data) > limits.MaxBytes {
		return nil, ErrLimit
	}
	fields := bytes.Fields(data)
	if len(fields) < 2 {
		return nil, ErrFormat
	}
	rowCount, rowErr := strconv.Atoi(string(fields[0]))
	columnCount, columnErr := strconv.Atoi(string(fields[1]))
	if rowErr != nil || columnErr != nil || rowCount < 0 || columnCount < 0 {
		return nil, ErrFormat
	}
	if rowCount > limits.MaxRows || columnCount > limits.MaxColumns {
		return nil, ErrLimit
	}
	words := (columnCount + 31) / 32
	want := 2 + rowCount*(1+words)
	if len(fields) != want {
		return nil, ErrFormat
	}
	matrix := NewEspressoSparseMatrix()
	if rowCount != 0 && columnCount != 0 {
		_ = matrix.Reserve(rowCount, columnCount)
	}
	position := 2
	for row := 0; row < rowCount; row++ {
		if _, err = strconv.ParseUint(string(fields[position]), 16, 64); err != nil {
			_ = matrix.Close()
			return nil, ErrFormat
		}
		position++
		for word := 0; word < words; word++ {
			value, parseErr := strconv.ParseUint(string(fields[position]), 16, 32)
			position++
			if parseErr != nil {
				_ = matrix.Close()
				return nil, ErrFormat
			}
			for bit := 0; bit < 32; bit++ {
				column := word*32 + bit
				if value>>uint(bit)&1 == 0 {
					continue
				}
				if column >= columnCount {
					_ = matrix.Close()
					return nil, ErrFormat
				}
				_, _ = matrix.Insert(row, column)
				if matrix.entries > limits.MaxEntries {
					_ = matrix.Close()
					return nil, ErrLimit
				}
			}
		}
	}
	return matrix, nil
}

func (matrix *EspressoSparseMatrix) Format() (string, error) {
	if err := matrix.check(); err != nil {
		return "", err
	}
	if len(matrix.columnIDs) == 0 {
		return "", nil
	}
	var output strings.Builder
	last := matrix.columnIDs[len(matrix.columnIDs)-1]
	if last >= 100 {
		output.WriteString("    ")
		for _, column := range matrix.columnIDs {
			output.WriteByte(byte('0' + column/100%10))
		}
		output.WriteByte('\n')
	}
	if last >= 10 {
		output.WriteString("    ")
		for _, column := range matrix.columnIDs {
			output.WriteByte(byte('0' + column/10%10))
		}
		output.WriteByte('\n')
	}
	output.WriteString("    ")
	for _, column := range matrix.columnIDs {
		output.WriteByte(byte('0' + column%10))
	}
	output.WriteString("\n    ")
	output.WriteString(strings.Repeat("-", len(matrix.columnIDs)))
	output.WriteByte('\n')
	for _, row := range matrix.rowIDs {
		rowText := strconv.Itoa(row)
		if len(rowText) < 3 {
			output.WriteString(strings.Repeat(" ", 3-len(rowText)))
		}
		output.WriteString(rowText)
		output.WriteByte(':')
		for _, column := range matrix.columnIDs {
			_, found := sparseSortedFind(matrix.rows[row].values, column)
			if found {
				output.WriteByte('1')
			} else {
				output.WriteByte('.')
			}
		}
		output.WriteByte('\n')
	}
	return output.String(), nil
}

func (matrix *EspressoSparseMatrix) Dump(label string, maximum int) (string, error) {
	if err := matrix.check(); err != nil {
		return "", err
	}
	if maximum < 0 {
		return "", ErrDimension
	}
	var output strings.Builder
	output.WriteString(label)
	output.WriteByte(' ')
	output.WriteString(strconv.Itoa(len(matrix.rowIDs)))
	output.WriteString(" rows by ")
	output.WriteString(strconv.Itoa(len(matrix.columnIDs)))
	output.WriteString(" cols\n")
	if len(matrix.rowIDs) < maximum {
		formatted, err := matrix.Format()
		if err != nil {
			return "", err
		}
		output.WriteString(formatted)
	}
	return output.String(), nil
}

func (matrix *EspressoSparseMatrix) Validate() error {
	if err := matrix.check(); err != nil {
		return err
	}
	if !slices.IsSorted(matrix.rowIDs) || !slices.IsSorted(matrix.columnIDs) {
		return ErrInvariant
	}
	entries := 0
	for _, row := range matrix.rowIDs {
		if row < 0 || row >= len(matrix.rows) || matrix.rows[row] == nil || matrix.rows[row].Number != row || len(matrix.rows[row].values) == 0 {
			return ErrInvariant
		}
		for _, column := range matrix.rows[row].values {
			if column < 0 || column >= len(matrix.columns) || matrix.columns[column] == nil {
				return ErrInvariant
			}
			_, found := sparseSortedFind(matrix.columns[column].values, row)
			if !found {
				return ErrInvariant
			}
			entries++
		}
	}
	for _, column := range matrix.columnIDs {
		if column < 0 || column >= len(matrix.columns) || matrix.columns[column] == nil || matrix.columns[column].Number != column || len(matrix.columns[column].values) == 0 {
			return ErrInvariant
		}
		for _, row := range matrix.columns[column].values {
			if row < 0 || row >= len(matrix.rows) || matrix.rows[row] == nil {
				return ErrInvariant
			}
			_, found := sparseSortedFind(matrix.rows[row].values, column)
			if !found {
				return ErrInvariant
			}
		}
	}
	for cell := range matrix.cellTags {
		if cell.row < 0 || cell.row >= len(matrix.rows) || matrix.rows[cell.row] == nil {
			return ErrInvariant
		}
		_, found := sparseSortedFind(matrix.rows[cell.row].values, cell.column)
		if !found {
			return ErrInvariant
		}
	}
	if entries != matrix.entries {
		return ErrInvariant
	}
	return nil
}
