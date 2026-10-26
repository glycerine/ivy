package abc

// Fast-CNF node-cone size diagnostics from sat/cnf/cnfFast.c.

// FastCNFNodeSize is the native diagnostic size before output and constant
// clauses are added. MarkedObjects is not the final CNF variable count.
type FastCNFNodeSize struct {
	MarkedObjects int
	Clauses       int
}

// CountAigFastCNFNodes marks the graph and counts the clauses contributed by
// marked AND cones. It preserves the source graph and returns no partial count.
func CountAigFastCNFNodes(g *AigManager, o FastCNFOptions) (FastCNFNodeSize, error) {
	c := o.CNF
	if c.MaxObjects < 0 || c.MaxVariables < 0 || c.MaxClauses < 0 || c.MaxLiterals < 0 || c.ISOP.MaxElements < 0 || c.ISOP.MaxVisits < 0 {
		return FastCNFNodeSize{}, ErrDimension
	}
	if c.MaxObjects == 0 {
		c.MaxObjects = 1 << 20
	}
	if c.MaxVariables == 0 {
		c.MaxVariables = 1 << 20
	}
	if c.MaxClauses == 0 {
		c.MaxClauses = 1 << 20
	}
	a, b, d := o.Cone.Stop, c.Stop, c.ISOP.Stop
	stop := func() bool { return a != nil && a() || b != nil && b() || d != nil && d() }
	o.Cone.Stop = stop
	c.ISOP.Stop = stop
	c.ISOP.TryBoth = false
	if g != nil && len(g.objects) > c.MaxObjects {
		return FastCNFNodeSize{}, ErrLimit
	}
	marking, e := MarkFastCNF(g, o.Cone)
	if e != nil {
		return FastCNFNodeSize{}, e
	}
	if marking.OutputDriverViolations != 0 || marking.AndGateViolations != 0 {
		return FastCNFNodeSize{}, ErrInvariant
	}
	var result FastCNFNodeSize
	for _, marked := range marking.Marked {
		if !marked {
			continue
		}
		result.MarkedObjects++
		if result.MarkedObjects > c.MaxVariables {
			return FastCNFNodeSize{}, ErrLimit
		}
	}
	visit := fastCNFConeVisit(o.Cone)
	add := func(n int) error {
		if stop() {
			return ErrCancelled
		}
		var e error
		result.Clauses, e = checkedSize(result.Clauses, n, 1, c.MaxClauses)
		return e
	}
	for id, obj := range g.objects {
		if obj.Type != AigAND || !marking.Marked[id] {
			continue
		}
		leaves, e := collectFastCNFLeaves(g, AigID(id), marking.Marked, false, visit)
		if e != nil {
			return FastCNFNodeSize{}, e
		}
		nodes, e := fastCNFVolume(g, AigID(id), leaves, visit)
		if e != nil {
			return FastCNFNodeSize{}, e
		}
		if len(leaves) > 6 {
			if e := add(len(leaves) + 1); e != nil {
				return FastCNFNodeSize{}, e
			}
			continue
		}
		ids := make([]AigID, len(leaves))
		for i, leaf := range leaves {
			ids[i] = leaf.ID()
		}
		truth, e := deriveFastCNFTruth(g, ids, nodes, visit)
		if e != nil {
			return FastCNFNodeSize{}, e
		}
		clauses := 0
		for polarity := 0; polarity < 2; polarity++ {
			word := truth
			if polarity != 0 {
				word = ^word
			}
			table, e := TruthTableFromWords(len(leaves), []uint64{word}, 0)
			if e != nil {
				return FastCNFNodeSize{}, e
			}
			cover, e := table.ISOP(c.ISOP)
			if e != nil {
				return FastCNFNodeSize{}, e
			}
			clauses, e = checkedSize(clauses, len(cover.Cubes), 1, c.MaxClauses)
			if e != nil {
				return FastCNFNodeSize{}, e
			}
		}
		if e := add(clauses); e != nil {
			return FastCNFNodeSize{}, e
		}
	}
	if stop() {
		return FastCNFNodeSize{}, ErrCancelled
	}
	return result, nil
}
