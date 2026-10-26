package abc

// Four-input mapped CNF emission from sat/cnf/cnfWrite.c:Cnf_ManWriteCnf.

import "slices"

type AigCNFCut4 struct {
	Root   AigID
	Leaves []AigID
	Truth  uint16 // stretched to four variables, with Leaves[0] the least significant input
}
type MappedCNFOptions struct {
	ISOP                                              TruthISOPOptions // per wide-cut polarity; TryBoth is disabled by the writer
	MaxObjects, MaxVariables, MaxClauses, MaxLiterals int              // zero: 1M, 1M, 1M, 4M
	Stop                                              func() bool
}
type AigMappedCNF struct {
	variables       int
	clauses         [][]SatLit
	objectVariables []int // -1 for unmapped objects and asserted COs
	inputs, outputs []SatLit
}

func (c *AigMappedCNF) VariableCount() int { return c.variables }
func (c *AigMappedCNF) Clauses() [][]SatLit {
	r := make([][]SatLit, len(c.clauses))
	for i, v := range c.clauses {
		r[i] = slices.Clone(v)
	}
	return r
}
func (c *AigMappedCNF) ObjectVariables() []int   { return slices.Clone(c.objectVariables) }
func (c *AigMappedCNF) InputLiterals() []SatLit  { return slices.Clone(c.inputs) }
func (c *AigMappedCNF) OutputLiterals() []SatLit { return slices.Clone(c.outputs) }

// WriteAigMappedCNF4 emits native raw clause order for caller-provided cuts.
// It does not derive cuts or prove their truth tables match the AIG. Variables
// start at one (zero is unused): exposed COs, mapped roots in supplied order,
// CIs, then constant one. Exposed outputs must be all combinational COs or the
// sequential latch-input suffix; remaining outputs are asserted true.
// Leaves may repeat but must precede their root and have assigned variables.
// Errors/cancellation expose no partial CNF and never mutate source/cuts.
func WriteAigMappedCNF4(g *AigManager, cuts []AigCNFCut4, exposedOutputs int, o MappedCNFOptions) (*AigMappedCNF, error) {
	if o.MaxObjects < 0 {
		return nil, ErrDimension
	}
	limit := o.MaxObjects
	if limit == 0 {
		limit = 1 << 20
	}
	if len(cuts) > limit {
		return nil, ErrLimit
	}
	if o.Stop != nil && o.Stop() {
		return nil, ErrCancelled
	}
	internal := make([]aigMappedCut, len(cuts))
	for i, c := range cuts {
		internal[i] = aigMappedCut{Root: c.Root, Leaves: c.Leaves, Truth: c.Truth}
	}
	return writeAigMappedCNF(g, internal, exposedOutputs, o)
}

type AigCNFCut struct {
	Root   AigID
	Leaves []AigID
	Truth  *TruthTable
}
type aigMappedCut struct {
	Root   AigID
	Leaves []AigID
	Truth  uint16
	table  *TruthTable
}

// WriteAigMappedCNF emits native mapped CNF through sixteen-input cuts. Small
// cuts use the exact four-input database; wider cuts derive ordered direct and
// complement ISOP covers. It trusts cut functions and does not derive mapping.
func WriteAigMappedCNF(g *AigManager, cuts []AigCNFCut, exposedOutputs int, o MappedCNFOptions) (*AigMappedCNF, error) {
	if o.MaxObjects < 0 {
		return nil, ErrDimension
	}
	limit := o.MaxObjects
	if limit == 0 {
		limit = 1 << 20
	}
	if len(cuts) > limit {
		return nil, ErrLimit
	}
	if o.Stop != nil && o.Stop() {
		return nil, ErrCancelled
	}
	internal := make([]aigMappedCut, len(cuts))
	for i, c := range cuts {
		internal[i] = aigMappedCut{Root: c.Root, Leaves: c.Leaves, table: c.Truth}
		if c.Truth == nil {
			return nil, ErrDimension
		}
	}
	return writeAigMappedCNF(g, internal, exposedOutputs, o)
}

func writeAigMappedCNF(g *AigManager, cuts []aigMappedCut, exposedOutputs int, o MappedCNFOptions) (*AigMappedCNF, error) {
	if g == nil || exposedOutputs < 0 || o.MaxObjects < 0 || o.MaxVariables < 0 || o.MaxClauses < 0 || o.MaxLiterals < 0 || o.ISOP.MaxElements < 0 || o.ISOP.MaxVisits < 0 {
		return nil, ErrDimension
	}
	if o.MaxObjects == 0 {
		o.MaxObjects = 1 << 20
	}
	if o.MaxVariables == 0 {
		o.MaxVariables = 1 << 20
	}
	if o.MaxClauses == 0 {
		o.MaxClauses = 1 << 20
	}
	if o.MaxLiterals == 0 {
		o.MaxLiterals = 4 << 20
	}
	if len(g.objects) > o.MaxObjects || len(cuts) > o.MaxObjects {
		return nil, ErrLimit
	}
	if e := g.Check(); e != nil {
		return nil, e
	}
	if exposedOutputs > len(g.cos) || exposedOutputs != 0 && (g.registers == 0 && exposedOutputs != len(g.cos) || g.registers != 0 && exposedOutputs != g.registers) {
		return nil, ErrDimension
	}
	stopped := func() bool { return o.Stop != nil && o.Stop() }
	if stopped() {
		return nil, ErrCancelled
	}
	variables := 1
	for _, count := range []int{exposedOutputs, len(cuts), len(g.cis), 1} {
		var e error
		variables, e = checkedSize(variables, 1, count, min(o.MaxVariables, int(MaxLiteralID)+1))
		if e != nil {
			return nil, e
		}
	}
	c := &AigMappedCNF{variables: variables, objectVariables: make([]int, len(g.objects))}
	for i := range c.objectVariables {
		c.objectVariables[i] = -1
	}
	next := 1
	for _, id := range g.cos[len(g.cos)-exposedOutputs:] {
		c.objectVariables[id] = next
		next++
	}
	for _, cut := range cuts {
		if stopped() {
			return nil, ErrCancelled
		}
		if int(cut.Root) >= len(g.objects) || g.objects[cut.Root].Type != AigAND || c.objectVariables[cut.Root] >= 0 || (len(cut.Leaves) > 16 || cut.table == nil && len(cut.Leaves) > 4) {
			return nil, ErrDimension
		}
		c.objectVariables[cut.Root] = next
		next++
	}
	for _, id := range g.cis {
		c.objectVariables[id] = next
		c.inputs = append(c.inputs, SatLit(next)<<1)
		next++
	}
	c.objectVariables[0] = next
	literals := 0
	add := func(clause []SatLit) error {
		if stopped() {
			return ErrCancelled
		}
		if len(c.clauses) >= o.MaxClauses || len(clause) > o.MaxLiterals-literals {
			return ErrLimit
		}
		literals += len(clause)
		c.clauses = append(c.clauses, clause)
		return nil
	}
	for _, cut := range cuts {
		if stopped() {
			return nil, ErrCancelled
		}
		n := len(cut.Leaves)
		truth4 := cut.Truth
		if cut.table != nil {
			if cut.table.variables != n {
				return nil, ErrDimension
			}
			if e := cut.table.Check(); e != nil {
				return nil, e
			}
			truth4 = uint16(cut.table.words[0])
		}
		if n <= 4 {
			mask := (1 << uint(n)) - 1
			for a := 0; a < 16; a++ {
				if truth4>>uint(a)&1 != truth4>>uint(a&mask)&1 {
					return nil, ErrDimension
				}
			}
		}
		vars := make([]SatVar, n)
		for i, id := range cut.Leaves {
			if id >= cut.Root || int(id) >= len(c.objectVariables) || c.objectVariables[id] < 0 {
				return nil, ErrDimension
			}
			vars[i] = SatVar(c.objectVariables[id])
		}
		for polarity := 0; polarity < 2; polarity++ {
			var cover []uint32
			if n <= 4 {
				truth := truth4
				if polarity == 1 {
					truth = ^truth
				}
				small, e := CNFSOP4Cover(truth)
				if e != nil {
					return nil, e
				}
				cover = make([]uint32, len(small))
				for i, c := range small {
					cover[i] = uint32(c)
				}
			} else {
				truth := cut.table
				if polarity == 1 {
					truth = truth.Not()
				}
				limits := o.ISOP
				limits.TryBoth = false
				prior := limits.Stop
				limits.Stop = func() bool { return stopped() || prior != nil && prior() }
				isop, e := truth.ISOP(limits)
				if e != nil {
					return nil, e
				}
				cover = isop.Cubes
			}
			for _, cube := range cover {
				body, e := CNFISOPCubeLiterals(uint32(cube), vars)
				if e != nil {
					return nil, e
				}
				root := SatLit(c.objectVariables[cut.Root])<<1 | SatLit(polarity)
				clause := make([]SatLit, 1, 1+len(body))
				clause[0] = root
				clause = append(clause, body...)
				if e = add(clause); e != nil {
					return nil, e
				}
			}
		}
	}
	if e := add([]SatLit{SatLit(c.objectVariables[0]) << 1}); e != nil {
		return nil, e
	}
	for i, id := range g.cos {
		if stopped() {
			return nil, ErrCancelled
		}
		driver := g.objects[id].Fanin0
		variable := c.objectVariables[driver.ID()]
		if variable < 0 {
			return nil, ErrDimension
		}
		out := SatLit(variable)<<1 | SatLit(driver&1)
		if i < len(g.cos)-exposedOutputs {
			if e := add([]SatLit{out}); e != nil {
				return nil, e
			}
			c.outputs = append(c.outputs, out)
		} else {
			po := SatLit(c.objectVariables[id]) << 1
			if e := add([]SatLit{po, out.Not()}); e != nil {
				return nil, e
			}
			if e := add([]SatLit{po.Not(), out}); e != nil {
				return nil, e
			}
			c.outputs = append(c.outputs, po)
		}
	}
	if stopped() {
		return nil, ErrCancelled
	}
	return c, nil
}
