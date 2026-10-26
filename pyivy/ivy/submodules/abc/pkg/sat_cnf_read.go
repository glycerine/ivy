package abc

// Checked CNF reading from sat/cnf/cnfUtil.c:Cnf_DataReadFromFile.

import (
	"bytes"
	"fmt"
	"io"
	"os"
)

type CNFDIMACSReadOptions struct {
	NativeLines            bool // one clause per line, ignore tokens after zero, accept count mismatch
	MaxBytes, MaxLineBytes int  // zero: 64 MiB and one million bytes
	Limits                 CNFTransformOptions
}

func normalizeCNFRead(o CNFDIMACSReadOptions) (CNFDIMACSReadOptions, error) {
	if o.MaxBytes < 0 || o.MaxLineBytes < 0 || o.Limits.MaxVariables < 0 || o.Limits.MaxClauses < 0 || o.Limits.MaxLiterals < 0 {
		return o, ErrDimension
	}
	if o.MaxBytes == 0 {
		o.MaxBytes = 64 << 20
	}
	if o.MaxLineBytes == 0 {
		o.MaxLineBytes = 1_000_000
	}
	if o.Limits.MaxVariables == 0 {
		o.Limits.MaxVariables = 1 << 20
	}
	if o.Limits.MaxClauses == 0 {
		o.Limits.MaxClauses = 1 << 20
	}
	if o.Limits.MaxLiterals == 0 {
		o.Limits.MaxLiterals = 4 << 20
	}
	return o, nil
}
func cnfSpace(b byte) bool {
	return b == ' ' || b == '\t' || b == '\r' || b == '\n' || b == '\v' || b == '\f'
}
func cnfDecimal(token []byte) (uint64, bool, bool) {
	negative := false
	if len(token) > 0 && (token[0] == '-' || token[0] == '+') {
		negative = token[0] == '-'
		token = token[1:]
	}
	if len(token) == 0 {
		return 0, false, false
	}
	var value uint64
	for _, b := range token {
		if b < '0' || b > '9' || value > (^uint64(0)-uint64(b-'0'))/10 {
			return 0, false, false
		}
		value = value*10 + uint64(b-'0')
	}
	return value, negative, true
}

// ParseCNFDIMACS reads standard one-based CNF with strict header/count/token
// validation, clauses spanning lines and multiple clauses per line. NativeLines
// selects ABC's line-based/count-warning behavior, returning actual clauses
// without printing warnings. Malformed decimal text is rejected in both modes.
// Quantified prefixes and raw-ID diagnostic text are not plain CNF input.
func ParseCNFDIMACS(data []byte, o CNFDIMACSReadOptions) (*CNFFormula, error) {
	var e error
	o, e = normalizeCNFRead(o)
	if e != nil {
		return nil, e
	}
	if len(data) > o.MaxBytes {
		return nil, ErrLimit
	}
	stop := func() bool { return o.Limits.Stop != nil && o.Limits.Stop() }
	if stop() {
		return nil, ErrCancelled
	}
	header := false
	variables, declared, total, lineNumber := 0, 0, 0, 0
	var clauses [][]SatLit
	var current []SatLit
	bad := func(message string) error { return fmt.Errorf("%w at line %d: %s", ErrFormat, lineNumber, message) }
	for offset := 0; offset < len(data); {
		if stop() {
			return nil, ErrCancelled
		}
		lineNumber++
		length := bytes.IndexByte(data[offset:], '\n')
		if length < 0 {
			length = len(data) - offset
		}
		if length > o.MaxLineBytes {
			return nil, ErrLimit
		}
		line := data[offset : offset+length]
		offset += length
		if offset < len(data) {
			offset++
		}
		cursor := 0
		next := func() []byte {
			for cursor < len(line) && cnfSpace(line[cursor]) {
				cursor++
			}
			start := cursor
			for cursor < len(line) && !cnfSpace(line[cursor]) {
				cursor++
			}
			return line[start:cursor]
		}
		token := next()
		if len(token) == 0 {
			continue
		}
		if token[0] == 'c' {
			continue
		}
		if token[0] == 'p' {
			if header {
				return nil, bad("duplicate header")
			}
			if !bytes.Equal(token, []byte("p")) {
				return nil, bad("invalid header")
			}
			if !bytes.Equal(next(), []byte("cnf")) {
				return nil, bad("expected cnf header")
			}
			nv, neg, ok := cnfDecimal(next())
			if !ok || neg {
				return nil, bad("invalid variable count")
			}
			nc, neg, ok := cnfDecimal(next())
			if !ok || neg || len(next()) != 0 {
				return nil, bad("invalid clause count")
			}
			if nv > uint64(o.Limits.MaxVariables) || nv > uint64(MaxLiteralID)+1 || nc > uint64(o.Limits.MaxClauses) {
				return nil, ErrLimit
			}
			if o.NativeLines && (nv == 0 || nc == 0) {
				return nil, bad("native reader requires positive counts")
			}
			variables, declared = int(nv), int(nc)
			header = true
			continue
		}
		if !header {
			return nil, bad("missing header")
		}
		terminated := false
		for len(token) > 0 {
			if stop() {
				return nil, ErrCancelled
			}
			magnitude, negative, ok := cnfDecimal(token)
			if !ok {
				return nil, bad("invalid integer")
			}
			if magnitude == 0 {
				if len(clauses) == o.Limits.MaxClauses {
					return nil, ErrLimit
				}
				clauses = append(clauses, current)
				current = nil
				terminated = true
				if !o.NativeLines && len(clauses) > declared {
					return nil, bad("too many clauses")
				}
				if o.NativeLines {
					break
				}
			} else {
				if magnitude > uint64(variables) {
					return nil, bad("literal outside declared span")
				}
				if total == o.Limits.MaxLiterals {
					return nil, ErrLimit
				}
				total++
				l := SatLit(magnitude-1) << 1
				if negative {
					l ^= 1
				}
				current = append(current, l)
				terminated = false
			}
			token = next()
		}
		if o.NativeLines && !terminated {
			return nil, bad("missing line zero terminator")
		}
	}
	if !header {
		return nil, bad("missing header")
	}
	if len(current) != 0 {
		return nil, bad("missing final zero terminator")
	}
	if !o.NativeLines && len(clauses) != declared {
		return nil, bad("clause count mismatch")
	}
	if stop() {
		return nil, ErrCancelled
	}
	return &CNFFormula{variables: variables, clauses: clauses}, nil
}

// ReadCNFDIMACSFile bounds plain-file reads before parsing; gzip input is not
// implicitly expanded, matching the native reader's plain-file scope.
func ReadCNFDIMACSFile(path string, o CNFDIMACSReadOptions) (*CNFFormula, error) {
	var e error
	o, e = normalizeCNFRead(o)
	if e != nil {
		return nil, e
	}
	if o.Limits.Stop != nil && o.Limits.Stop() {
		return nil, ErrCancelled
	}
	file, e := os.Open(path)
	if e != nil {
		return nil, e
	}
	defer file.Close()
	var data []byte
	buffer := make([]byte, 32<<10)
	for {
		if o.Limits.Stop != nil && o.Limits.Stop() {
			return nil, ErrCancelled
		}
		n, e := file.Read(buffer)
		if n > o.MaxBytes-len(data) {
			return nil, ErrLimit
		}
		data = append(data, buffer[:n]...)
		if e == io.EOF {
			break
		}
		if e != nil {
			return nil, e
		}
	}
	return ParseCNFDIMACS(data, o)
}
