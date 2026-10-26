package abc

// Object-ID CNF from sat/cnf/cnfWrite.c:Cnf_ManWriteCnfOther.

import "slices"

type AigObjectCNF struct {
	*AigMappedCNF
	starts, counts []int
	mapping        []int32
}

func (c *AigObjectCNF) ObjectClauseStarts() []int { return slices.Clone(c.starts) }
func (c *AigObjectCNF) ObjectClauseCounts() []int { return slices.Clone(c.counts) }
func (c *AigObjectCNF) MappingArray() []int32     { return slices.Clone(c.mapping) }

// WriteAigObjectCNF4 uses object IDs as SAT variable IDs, aliases every CO,
// and records each object's contiguous clause range. CI counts are zero;
// unmapped objects have start/count -1. Constant one is variable zero and its
// unit clause is last. Supplied functions are trusted; graph storage is unchanged.
func WriteAigObjectCNF4(g *AigManager, cuts []AigCNFCut4, o MappedCNFOptions) (*AigObjectCNF, error) {
	compact, options, e := objectCNFOptions(g, o)
	if e != nil {
		return nil, e
	}
	c, e := WriteAigMappedCNF4(&compact, cuts, len(g.cos), options)
	if e != nil {
		return nil, e
	}
	return reindexObjectCNF(g, c, o.Stop)
}

// WriteAigObjectCNF additionally supports supplied ISOP cuts through sixteen inputs.
func WriteAigObjectCNF(g *AigManager, cuts []AigCNFCut, o MappedCNFOptions) (*AigObjectCNF, error) {
	compact, options, e := objectCNFOptions(g, o)
	if e != nil {
		return nil, e
	}
	c, e := WriteAigMappedCNF(&compact, cuts, len(g.cos), options)
	if e != nil {
		return nil, e
	}
	return reindexObjectCNF(g, c, o.Stop)
}
func objectCNFOptions(g *AigManager, o MappedCNFOptions) (AigManager, MappedCNFOptions, error) {
	if g == nil || o.MaxVariables < 0 {
		return AigManager{}, o, ErrDimension
	}
	limit := o.MaxVariables
	if limit == 0 {
		limit = 1 << 20
	}
	if len(g.objects) > limit || uint64(len(g.objects)) > uint64(MaxLiteralID)+1 {
		return AigManager{}, o, ErrLimit
	}
	if o.Stop != nil && o.Stop() {
		return AigManager{}, o, ErrCancelled
	}
	compact := *g
	compact.registers = 0
	// The temporary numbering reserves variable zero; the final representation
	// uses it for constant one. Apply caller's variable cap to the final span.
	o.MaxVariables = len(g.objects) + 1
	return compact, o, nil
}
func reindexObjectCNF(g *AigManager, c *AigMappedCNF, stop func() bool) (*AigObjectCNF, error) {
	r := &AigObjectCNF{AigMappedCNF: &AigMappedCNF{variables: len(g.objects), objectVariables: make([]int, len(g.objects))}, starts: make([]int, len(g.objects)), counts: make([]int, len(g.objects))}
	reverse := make([]AigID, c.variables)
	for id, v := range c.objectVariables {
		r.objectVariables[id] = -1
		r.starts[id] = -1
		r.counts[id] = -1
		if v >= 0 {
			reverse[v] = AigID(id)
			r.objectVariables[id] = id
		}
	}
	for _, id := range g.cis {
		r.inputs = append(r.inputs, SatLit(id)<<1)
		r.counts[id] = 0
	}
	for _, id := range g.cos {
		r.outputs = append(r.outputs, SatLit(id)<<1)
	}
	mappedCount := len(c.clauses) - 2*len(g.cos) - 1
	appendClause := func(index int, owner AigID) error {
		if stop != nil && stop() {
			return ErrCancelled
		}
		if r.starts[owner] < 0 {
			r.starts[owner] = len(r.clauses)
			r.counts[owner] = 0
		}
		r.counts[owner]++
		clause := c.clauses[index]
		for i, l := range clause {
			clause[i] = SatLit(reverse[l.Var()])<<1 | l&1
		}
		r.clauses = append(r.clauses, clause)
		return nil
	}
	for i := 0; i < mappedCount; i++ {
		owner := reverse[c.clauses[i][0].Var()]
		if e := appendClause(i, owner); e != nil {
			return nil, e
		}
	}
	for i, id := range g.cos {
		for j := 0; j < 2; j++ {
			if e := appendClause(mappedCount+1+2*i+j, id); e != nil {
				return nil, e
			}
		}
	}
	if e := appendClause(mappedCount, 0); e != nil {
		return nil, e
	}
	if stop != nil && stop() {
		return nil, ErrCancelled
	}
	return r, nil
}

// DeriveAigObjectCNF4 matches Cnf_DeriveOtherWithMan, including its preorder
// packed mapping. All outputs (including latch inputs) are represented by CO IDs.
func DeriveAigObjectCNF4(g *AigManager, o AigCNFDeriveOptions) (*AigObjectCNF, error) {
	a, b := o.Cuts.Stop, o.CNF.Stop
	stop := func() bool { return a != nil && a() || b != nil && b() }
	o.Cuts.Stop = stop
	o.CNF.Stop = stop
	d, e := ComputeDARCuts4(g, o.Cuts)
	if e != nil {
		return nil, e
	}
	m, e := SelectCNFAreaFlow4(d.Graph, d.Candidates, CNFAreaFlowOptions{MaxObjects: o.CNF.MaxObjects, Stop: stop})
	if e != nil {
		return nil, e
	}
	scan, e := ScanCNFMapping4(d.Graph, m.Cuts, CNFMappingScanOptions{Preorder: true, MaxObjects: o.CNF.MaxObjects, Stop: stop})
	if e != nil {
		return nil, e
	}
	r, e := WriteAigObjectCNF4(d.Graph, scan.Cuts, o.CNF)
	if e != nil {
		return nil, e
	}
	r.mapping, e = WriteCNFMappingArray4(d.Graph, scan.Cuts, CNFMappingArrayOptions{MaxObjects: o.CNF.MaxObjects, Stop: stop})
	if e != nil {
		return nil, e
	}
	return r, nil
}
