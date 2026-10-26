package abc

// Text serialization from sat/cnf/cnfMan.c:Cnf_DataWriteIntoFile.

import "strconv"

type CNFDIMACSOptions struct {
	Readable       bool     // native raw-ID format; variable zero prints zero, so this is not standard DIMACS
	ForAll, Exists []SatVar // nil omits the line; nonnil empty emits a terminated empty prefix
	MaxBytes       int      // zero: 64 MiB
	Limits         CNFTransformOptions
}

// DIMACS preserves native header, quantifier order, spaces and final blank
// line. Default numbering is standard one-based DIMACS; Readable selects the
// native raw-ID diagnostic mode. Prefix entries are checked, not reordered or
// deduplicated. Output is owned; failure/cancellation returns no partial bytes.
func (f *CNFFormula) DIMACS(o CNFDIMACSOptions) ([]byte, error) {
	if o.MaxBytes < 0 {
		return nil, ErrDimension
	}
	if o.MaxBytes == 0 {
		o.MaxBytes = 64 << 20
	}
	if e := checkCNFFormula(f, o.Limits); e != nil {
		return nil, e
	}
	var out []byte
	appendText := func(s string) error {
		if len(s) > o.MaxBytes-len(out) {
			return ErrLimit
		}
		out = append(out, s...)
		return nil
	}
	integer := func(v int64) error {
		var scratch [24]byte
		b := strconv.AppendInt(scratch[:0], v, 10)
		if len(b) > o.MaxBytes-len(out) {
			return ErrLimit
		}
		out = append(out, b...)
		return nil
	}
	stop := func() bool { return o.Limits.Stop != nil && o.Limits.Stop() }
	if e := appendText("c Result of efficient AIG-to-CNF conversion using package CNF\np cnf "); e != nil {
		return nil, e
	}
	if e := integer(int64(f.variables)); e != nil {
		return nil, e
	}
	if e := appendText(" "); e != nil {
		return nil, e
	}
	if e := integer(int64(len(f.clauses))); e != nil {
		return nil, e
	}
	if e := appendText("\n"); e != nil {
		return nil, e
	}
	for i, list := range [][]SatVar{o.ForAll, o.Exists} {
		if list == nil {
			continue
		}
		if e := appendText([]string{"a ", "e "}[i]); e != nil {
			return nil, e
		}
		for _, v := range list {
			if stop() {
				return nil, ErrCancelled
			}
			if uint64(v) >= uint64(f.variables) {
				return nil, ErrInvalidLiteral
			}
			number := int64(v)
			if !o.Readable {
				number++
			}
			if e := integer(number); e != nil {
				return nil, e
			}
			if e := appendText(" "); e != nil {
				return nil, e
			}
		}
		if e := appendText("0\n"); e != nil {
			return nil, e
		}
	}
	for _, clause := range f.clauses {
		if stop() {
			return nil, ErrCancelled
		}
		for _, l := range clause {
			if stop() {
				return nil, ErrCancelled
			}
			number := int64(l.Var())
			if !o.Readable {
				number++
			}
			if l&1 != 0 {
				number = -number
			}
			if e := integer(number); e != nil {
				return nil, e
			}
			if e := appendText(" "); e != nil {
				return nil, e
			}
		}
		if e := appendText("0\n"); e != nil {
			return nil, e
		}
	}
	if e := appendText("\n"); e != nil {
		return nil, e
	}
	if stop() {
		return nil, ErrCancelled
	}
	return out, nil
}
func (c *AigMappedCNF) DIMACS(o CNFDIMACSOptions) ([]byte, error) {
	if c == nil {
		return nil, ErrDimension
	}
	return (&CNFFormula{variables: c.variables, clauses: c.clauses}).DIMACS(o)
}
func (f *AigCNFFrames) DIMACS(o CNFDIMACSOptions) ([]byte, error) {
	if f == nil {
		return nil, ErrDimension
	}
	return (&CNFFormula{variables: f.variables, clauses: f.clauses}).DIMACS(o)
}
