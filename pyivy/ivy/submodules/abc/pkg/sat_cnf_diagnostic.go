package abc

// Diagnostic formatting and graph-variable collection from sat/cnf/cnfMan.c
// and sat/cnf/cnfUtil.c.

import "strconv"

type CNFDiagnosticOptions struct {
	Readable bool // raw zero-based IDs; complemented variable zero prints -0
	MaxBytes int  // zero: 64 MiB
	Limits   CNFTransformOptions
}

// DiagnosticText preserves Cnf_DataPrint output. Unlike DIMACS, clauses have
// no zero terminator and the output has no comment line, so it is diagnostic
// text rather than a round-trippable interchange format.
func (f *CNFFormula) DiagnosticText(o CNFDiagnosticOptions) ([]byte, error) {
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
	appendBytes := func(b []byte) error {
		if len(b) > o.MaxBytes-len(out) {
			return ErrLimit
		}
		out = append(out, b...)
		return nil
	}
	appendString := func(s string) error { return appendBytes([]byte(s)) }
	appendNumber := func(n int) error {
		var scratch [24]byte
		return appendBytes(strconv.AppendInt(scratch[:0], int64(n), 10))
	}
	stop := func() bool { return o.Limits.Stop != nil && o.Limits.Stop() }
	if e := appendString("p cnf "); e != nil {
		return nil, e
	}
	if e := appendNumber(f.variables); e != nil {
		return nil, e
	}
	if e := appendString(" "); e != nil {
		return nil, e
	}
	if e := appendNumber(len(f.clauses)); e != nil {
		return nil, e
	}
	if e := appendString("\n"); e != nil {
		return nil, e
	}
	for _, clause := range f.clauses {
		if stop() {
			return nil, ErrCancelled
		}
		for _, l := range clause {
			if stop() {
				return nil, ErrCancelled
			}
			if l.Negative() {
				if e := appendString("-"); e != nil {
					return nil, e
				}
			}
			number := int(l.Var())
			if !o.Readable {
				number++
			}
			if e := appendNumber(number); e != nil {
				return nil, e
			}
			if e := appendString(" "); e != nil {
				return nil, e
			}
		}
		if e := appendString("\n"); e != nil {
			return nil, e
		}
	}
	if e := appendString("\n"); e != nil {
		return nil, e
	}
	if stop() {
		return nil, ErrCancelled
	}
	return out, nil
}

func (c *AigMappedCNF) DiagnosticText(o CNFDiagnosticOptions) ([]byte, error) {
	if c == nil {
		return nil, ErrDimension
	}
	return (&CNFFormula{variables: c.variables, clauses: c.clauses}).DiagnosticText(o)
}
func (c *AigCNF) DiagnosticText(o CNFDiagnosticOptions) ([]byte, error) {
	if c == nil {
		return nil, ErrDimension
	}
	return (&CNFFormula{variables: c.variables, clauses: c.clauses}).DiagnosticText(o)
}
func (f *AigCNFFrames) DiagnosticText(o CNFDiagnosticOptions) ([]byte, error) {
	if f == nil {
		return nil, ErrDimension
	}
	return (&CNFFormula{variables: f.variables, clauses: f.clauses}).DiagnosticText(o)
}

type CNFVariableMapOptions struct {
	MaxObjects int // zero: one million
	Stop       func() bool
}

func (c *AigMappedCNF) collectObjectVariables(g *AigManager, ids []AigID, o CNFVariableMapOptions) ([]int, error) {
	if c == nil || g == nil || o.MaxObjects < 0 || len(c.objectVariables) != len(g.objects) {
		return nil, ErrDimension
	}
	if o.MaxObjects == 0 {
		o.MaxObjects = 1 << 20
	}
	if len(g.objects) > o.MaxObjects || len(ids) > o.MaxObjects {
		return nil, ErrLimit
	}
	if o.Stop != nil && o.Stop() {
		return nil, ErrCancelled
	}
	if e := g.Check(); e != nil {
		return nil, e
	}
	result := make([]int, len(ids))
	for i, id := range ids {
		if o.Stop != nil && o.Stop() {
			return nil, ErrCancelled
		}
		v := c.objectVariables[id]
		if v < -1 || v >= c.variables {
			return nil, ErrDimension
		}
		result[i] = v
	}
	if o.Stop != nil && o.Stop() {
		return nil, ErrCancelled
	}
	return result, nil
}

// CIVariables returns object-variable entries for CIs in graph order. Values
// may be -1 when a caller-provided CNF left an object unassigned.
func (c *AigMappedCNF) CIVariables(g *AigManager, o CNFVariableMapOptions) ([]int, error) {
	if g == nil {
		return nil, ErrDimension
	}
	return c.collectObjectVariables(g, g.cis, o)
}

// PIVariables is the native compatibility spelling. Like Cnf_DataCollectPiSatNums,
// it includes every CI, including latch outputs in a sequential graph.
func (c *AigMappedCNF) PIVariables(g *AigManager, o CNFVariableMapOptions) ([]int, error) {
	return c.CIVariables(g, o)
}

// COVariables returns object-variable entries for COs in graph order.
func (c *AigMappedCNF) COVariables(g *AigManager, o CNFVariableMapOptions) ([]int, error) {
	if g == nil {
		return nil, ErrDimension
	}
	return c.collectObjectVariables(g, g.cos, o)
}
