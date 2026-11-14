package cudd

import (
	"fmt"
	"io"
	"math/bits"
	"strconv"
	"strings"

	"github.com/glycerine/abcgo/bdd/cudd/core"
)

// HarwellMatrix is the ADD and variable set read from a Harwell–Boeing
// assembled sparse matrix. X and Y contain row and column indicator vars;
// XNeg and YNeg contain their arithmetic complements.
type HarwellMatrix struct {
	Function       *DdNode
	X, Y           []*DdNode
	XNeg, YNeg     []*DdNode
	Rows, Cols     int
	RightHandSides int
}

// Cudd_addHarwell reads a real, unsymmetric assembled matrix (RUA) from r.
// Variable indices are assigned bx+i*sx to rows and by+i*sy to columns.
// As in CUDD, optional right hand sides are represented as extra columns.
// Unsupported matrix forms and malformed records return an error without
// publishing a partial result.
func Cudd_addHarwell(r io.Reader, dd *DdManager, bx, sx, by, sy int, verbose io.Writer) (*HarwellMatrix, error) {
	if r == nil || dd == nil || sx == 0 || sy == 0 {
		return nil, fmt.Errorf("invalid Harwell reader arguments")
	}
	result, ok, err := Cudd_addHarwellNext(r, dd, bx, sx, by, sy, verbose)
	if err != nil {
		return nil, err
	}
	if !ok {
		return nil, io.EOF
	}
	return result, nil
}

// Cudd_addHarwellNext reads exactly one Harwell–Boeing matrix. It reads the
// stream one byte at a time at record boundaries so repeated calls preserve
// the next matrix even when r is not buffered. ok is false only for clean EOF
// before the next matrix's first header record.
func Cudd_addHarwellNext(r io.Reader, dd *DdManager, bx, sx, by, sy int, verbose io.Writer) (*HarwellMatrix, bool, error) {
	if r == nil || dd == nil || sx == 0 || sy == 0 {
		return nil, false, fmt.Errorf("invalid Harwell reader arguments")
	}
	records, err := readHarwellMatrixRecords(r)
	if err == io.EOF {
		return nil, false, nil
	}
	if err != nil {
		return nil, false, err
	}
	result, err := parseHarwellMatrixRecords(records, dd, bx, sx, by, sy, verbose)
	return result, true, err
}

func readHarwellLine(r io.Reader) (string, error) {
	var line strings.Builder
	var one [1]byte
	for {
		n, err := io.ReadFull(r, one[:])
		if n == 1 {
			if one[0] == '\n' {
				return strings.TrimSuffix(line.String(), "\r"), nil
			}
			line.WriteByte(one[0])
		}
		if err != nil {
			if err == io.EOF && line.Len() > 0 {
				return line.String(), nil
			}
			return "", err
		}
	}
}

func readHarwellMatrixRecords(r io.Reader) ([]string, error) {
	records := make([]string, 0, 8)
	for i := 0; i < 4; i++ {
		line, err := readHarwellLine(r)
		if err != nil {
			if i == 0 && err == io.EOF {
				return nil, io.EOF
			}
			return nil, fmt.Errorf("Harwell–Boeing input has a truncated header: %w", err)
		}
		records = append(records, line)
	}
	counts, err := parseHBInts(records[1], 5)
	if err != nil {
		return nil, fmt.Errorf("header record 2: %w", err)
	}
	meta := strings.Fields(records[2])
	if len(meta) < 5 {
		return nil, fmt.Errorf("header record 3 is malformed")
	}
	nrow, e1 := strconv.Atoi(meta[1])
	ncol, e2 := strconv.Atoi(meta[2])
	nnz, e3 := strconv.Atoi(meta[3])
	if e1 != nil || e2 != nil || e3 != nil || nrow < 0 || ncol < 0 || nnz < 0 {
		return nil, fmt.Errorf("header record 3 has invalid dimensions")
	}
	rhs := 0
	if counts[4] > 0 {
		line, err := readHarwellLine(r)
		if err != nil {
			return nil, fmt.Errorf("missing right-hand-side header: %w", err)
		}
		records = append(records, line)
		fields := strings.Fields(line)
		if len(fields) < 3 {
			return nil, fmt.Errorf("right-hand-side header is malformed")
		}
		rhs, err = strconv.Atoi(fields[1])
		if err != nil || rhs < 0 {
			return nil, fmt.Errorf("invalid right-hand-side count")
		}
	}
	need := ncol + 1 + 2*nnz + rhs*nrow
	fieldCount := 0
	for fieldCount < need {
		line, err := readHarwellLine(r)
		if err != nil {
			return nil, fmt.Errorf("matrix data is truncated: need %d numeric fields, found %d: %w", need, fieldCount, err)
		}
		records = append(records, line)
		fieldCount += len(strings.Fields(line))
	}
	return records, nil
}

func parseHarwellMatrixRecords(records []string, dd *DdManager, bx, sx, by, sy int, verbose io.Writer) (*HarwellMatrix, error) {
	if dd == nil || len(records) < 4 {
		return nil, fmt.Errorf("invalid Harwell matrix records")
	}
	title, key := records[0], ""
	if len(title) > 72 {
		key = strings.TrimSpace(title[72:])
		title = title[:72]
	}
	if len(records[0]) >= 80 {
		title, key = records[0][:72], strings.TrimSpace(records[0][72:80])
	}
	counts, err := parseHBInts(records[1], 5)
	if err != nil {
		return nil, fmt.Errorf("header record 2: %w", err)
	}
	rhsRecords := counts[4]
	meta := strings.Fields(records[2])
	if len(meta) < 5 || len(meta[0]) < 3 {
		return nil, fmt.Errorf("header record 3 is malformed")
	}
	matrixType := meta[0][:3]
	nrow, err := strconv.Atoi(meta[1])
	if err != nil {
		return nil, fmt.Errorf("invalid row count: %w", err)
	}
	ncol, err := strconv.Atoi(meta[2])
	if err != nil {
		return nil, fmt.Errorf("invalid column count: %w", err)
	}
	nnz, err := strconv.Atoi(meta[3])
	if err != nil {
		return nil, fmt.Errorf("invalid nonzero count: %w", err)
	}
	nelt, err := strconv.Atoi(meta[4])
	if err != nil {
		return nil, fmt.Errorf("invalid element count: %w", err)
	}
	if matrixType != "RUA" {
		return nil, fmt.Errorf("%s: unsupported matrix type %q (only RUA is supported)", key, matrixType)
	}
	if nelt != 0 || nrow < 0 || ncol < 0 || nnz < 0 {
		return nil, fmt.Errorf("invalid assembled matrix dimensions")
	}
	// Record 4 contains FORTRAN field formats. It is intentionally parsed as
	// a whole record because descriptors such as (10I8) are not whitespace data.
	dataAt := 4
	rhs := 0
	if rhsRecords > 0 {
		if len(records) <= dataAt {
			return nil, fmt.Errorf("missing right-hand-side header")
		}
		rh := strings.Fields(records[dataAt])
		dataAt++
		if len(rh) < 3 || !strings.HasPrefix(rh[0], "F") {
			return nil, fmt.Errorf("sparse right-hand sides are unsupported")
		}
		rhs, err = strconv.Atoi(rh[1])
		if err != nil || rhs < 0 {
			return nil, fmt.Errorf("invalid right-hand-side count")
		}
		if _, err = strconv.Atoi(rh[2]); err != nil {
			return nil, fmt.Errorf("invalid right-hand-side index count")
		}
	}
	if verbose != nil {
		fmt.Fprintf(verbose, "%s: type %s, %d rows, %d columns, %d entries\n", key, matrixType, nrow, ncol, nnz)
		if rhs > 0 {
			fmt.Fprintf(verbose, "%d right-hand side(s)\n", rhs)
		}
		if title != "" {
			fmt.Fprintln(verbose, strings.TrimSpace(title))
		}
	}
	data := strings.Join(records[dataAt:], " ")
	fields := strings.Fields(data)
	need := ncol + 1 + nnz + nnz + rhs*nrow
	if len(fields) < need {
		return nil, fmt.Errorf("matrix data is truncated: need %d numeric fields, found %d", need, len(fields))
	}
	pos := 0
	readInts := func(n int) ([]int, error) {
		out := make([]int, n)
		for i := range out {
			v, e := strconv.Atoi(fields[pos])
			if e != nil {
				return nil, e
			}
			out[i] = v - 1
			pos++
		}
		return out, nil
	}
	colPtr, err := readInts(ncol + 1)
	if err != nil {
		return nil, fmt.Errorf("invalid column pointer: %w", err)
	}
	rowIndex, err := readInts(nnz)
	if err != nil {
		return nil, fmt.Errorf("invalid row index: %w", err)
	}
	if len(colPtr) == 0 || colPtr[0] != 0 || colPtr[ncol] != nnz {
		return nil, fmt.Errorf("column pointers do not span the nonzero entries")
	}
	for i := 0; i < ncol; i++ {
		if colPtr[i] < 0 || colPtr[i] > colPtr[i+1] || colPtr[i+1] > nnz {
			return nil, fmt.Errorf("column pointers are not monotone")
		}
	}
	for _, row := range rowIndex {
		if row < 0 || row >= nrow {
			return nil, fmt.Errorf("row index %d is outside matrix", row+1)
		}
	}
	values := make([]float64, nnz+rhs*nrow)
	for i := range values {
		values[i], err = strconv.ParseFloat(fields[pos], 64)
		if err != nil {
			return nil, fmt.Errorf("invalid matrix value %q: %w", fields[pos], err)
		}
		pos++
	}

	rowBits := bits.Len(uint(maxInt(nrow-1, 0)))
	colDomain := ncol
	if rhs > 0 {
		colDomain = 2 * maxInt(rhs, ncol)
	}
	colBits := bits.Len(uint(maxInt(colDomain-1, 0)))
	result := &HarwellMatrix{Rows: nrow, Cols: ncol, RightHandSides: rhs}
	result.X, result.XNeg, err = makeAddLiterals(dd, rowBits, bx, sx)
	if err != nil {
		return nil, err
	}
	result.Y, result.YNeg, err = makeAddLiterals(dd, colBits, by, sy)
	if err != nil {
		return nil, err
	}
	zero := core.Cudd_ReadZero(dd)
	function := zero
	addProduct := func(a, b *DdNode) *DdNode { return core.Cudd_addApply(dd, core.Cudd_addTimes, a, b) }
	addTerms := func(row, col int, value float64) bool {
		term := core.Cudd_addConst(dd, value)
		for bit := 0; bit < rowBits; bit++ {
			lit := result.XNeg[rowBits-1-bit]
			if row&(1<<bit) != 0 {
				lit = result.X[rowBits-1-bit]
			}
			term = addProduct(term, lit)
			if term == nil {
				return false
			}
		}
		for bit := 0; bit < colBits; bit++ {
			lit := result.YNeg[colBits-1-bit]
			if col&(1<<bit) != 0 {
				lit = result.Y[colBits-1-bit]
			}
			term = addProduct(term, lit)
			if term == nil {
				return false
			}
		}
		function = core.Cudd_addApply(dd, core.Cudd_addPlus, function, term)
		return function != nil
	}
	for col := 0; col < ncol; col++ {
		for entry := colPtr[col]; entry < colPtr[col+1]; entry++ {
			if !addTerms(rowIndex[entry], col, values[entry]) {
				return nil, fmt.Errorf("failed to construct ADD for matrix entry")
			}
		}
	}
	for rhsIndex := 0; rhsIndex < rhs; rhsIndex++ {
		col := (1 << maxInt(colBits-1, 0)) + rhsIndex
		for row := 0; row < nrow; row++ {
			v := values[nnz+rhsIndex*nrow+row]
			if v != 0 && !addTerms(row, col, v) {
				return nil, fmt.Errorf("failed to construct ADD for right-hand side")
			}
		}
	}
	Cudd_Ref(function)
	result.Function = function
	return result, nil
}

func makeAddLiterals(dd *DdManager, count, first, step int) (positive, negative []*DdNode, err error) {
	positive, negative = make([]*DdNode, count), make([]*DdNode, count)
	one, zero := core.Cudd_ReadOne(dd), core.Cudd_ReadZero(dd)
	for i := 0; i < count; i++ {
		v := core.Cudd_addIthVar(dd, first+i*step)
		if v == nil {
			return nil, nil, fmt.Errorf("failed to create ADD variable %d", first+i*step)
		}
		positive[i] = v
		negative[i] = core.Cudd_addIte(dd, v, zero, one)
		if negative[i] == nil {
			return nil, nil, fmt.Errorf("failed to create complemented ADD variable %d", first+i*step)
		}
	}
	return positive, negative, nil
}

func parseHBInts(record string, count int) ([]int, error) {
	fields := strings.Fields(record)
	if len(fields) < count {
		return nil, fmt.Errorf("expected %d integers, found %d", count, len(fields))
	}
	values := make([]int, count)
	for i := range values {
		n, err := strconv.Atoi(fields[i])
		if err != nil {
			return nil, err
		}
		values[i] = n
	}
	return values, nil
}

func maxInt(a, b int) int {
	if a > b {
		return a
	}
	return b
}
