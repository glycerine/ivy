package abc

// Fast supergate/ISOP derivation from sat/cnf/cnfFast.c.

type FastCNFOptions struct {
	Cone FastCNFConeOptions
	CNF  MappedCNFOptions
}

func fastCNFVolume(g *AigManager, root AigID, leaves []AigLit, visit func() error) ([]AigID, error) {
	seen := make(map[AigID]bool, len(leaves))
	for _, l := range leaves {
		seen[l.ID()] = true
	}
	type event struct {
		id   AigID
		exit bool
	}
	stack := []event{{id: root}}
	var nodes []AigID
	for len(stack) > 0 {
		if e := visit(); e != nil {
			return nil, e
		}
		v := stack[len(stack)-1]
		stack = stack[:len(stack)-1]
		if v.exit {
			nodes = append(nodes, v.id)
			continue
		}
		if seen[v.id] {
			continue
		}
		seen[v.id] = true
		obj := g.objects[v.id]
		if obj.Type != AigAND {
			return nil, ErrInvariant
		}
		stack = append(stack, event{id: v.id, exit: true}, event{id: obj.Fanin1.ID()}, event{id: obj.Fanin0.ID()})
	}
	return nodes, nil
}

// DeriveAigFastCNF preserves native reverse-object variable/clause order,
// flattened AND gates and direct six-input ISOP covers. It neither cleans nor
// mutates the graph. Invalid marking diagnostics or unsupported cones return
// errors rather than emitting unassigned literals. All output is owned.
func DeriveAigFastCNF(g *AigManager, exposedOutputs int, o FastCNFOptions) (*AigMappedCNF, error) {
	c := o.CNF
	if exposedOutputs < 0 || c.MaxObjects < 0 || c.MaxVariables < 0 || c.MaxClauses < 0 || c.MaxLiterals < 0 || c.ISOP.MaxElements < 0 || c.ISOP.MaxVisits < 0 {
		return nil, ErrDimension
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
	if c.MaxLiterals == 0 {
		c.MaxLiterals = 4 << 20
	}
	a, b, d := o.Cone.Stop, c.Stop, c.ISOP.Stop
	stop := func() bool { return a != nil && a() || b != nil && b() || d != nil && d() }
	o.Cone.Stop = stop
	c.ISOP.Stop = stop
	c.ISOP.TryBoth = false
	if g != nil && len(g.objects) > c.MaxObjects {
		return nil, ErrLimit
	}
	marking, e := MarkFastCNF(g, o.Cone)
	if e != nil {
		return nil, e
	}
	if marking.OutputDriverViolations != 0 || marking.AndGateViolations != 0 {
		return nil, ErrInvariant
	}
	if exposedOutputs > len(g.cos) || exposedOutputs != 0 && (g.registers == 0 && exposedOutputs != len(g.cos) || g.registers != 0 && exposedOutputs != g.registers) {
		return nil, ErrDimension
	}
	r := &AigMappedCNF{variables: 1, objectVariables: make([]int, len(g.objects))}
	for i := range r.objectVariables {
		r.objectVariables[i] = -1
	}
	assign := func(id AigID) error {
		next, e := checkedSize(r.variables, 1, 1, min(c.MaxVariables, int(MaxLiteralID)+1))
		if e != nil {
			return e
		}
		r.objectVariables[id] = r.variables
		r.variables = next
		return nil
	}
	for _, id := range g.cos[len(g.cos)-exposedOutputs:] {
		if e := assign(id); e != nil {
			return nil, e
		}
	}
	for id := len(g.objects) - 1; id >= 0; id-- {
		if stop() {
			return nil, ErrCancelled
		}
		if g.objects[id].Type == AigAND && marking.Marked[id] {
			if e := assign(AigID(id)); e != nil {
				return nil, e
			}
		}
	}
	for _, id := range g.cis {
		if e := assign(id); e != nil {
			return nil, e
		}
		r.inputs = append(r.inputs, SatLit(r.objectVariables[id])<<1)
	}
	if e := assign(0); e != nil {
		return nil, e
	}
	literal := func(l AigLit) (SatLit, error) {
		v := r.objectVariables[l.ID()]
		if v < 1 {
			return 0, ErrInvariant
		}
		return SatLit(v)<<1 | SatLit(l&1), nil
	}
	literals := 0
	add := func(clause []SatLit) error {
		if stop() {
			return ErrCancelled
		}
		if len(r.clauses) == c.MaxClauses || len(clause) > c.MaxLiterals-literals {
			return ErrLimit
		}
		literals += len(clause)
		r.clauses = append(r.clauses, clause)
		return nil
	}
	visit := fastCNFConeVisit(o.Cone)
	for id := len(g.objects) - 1; id >= 0; id-- {
		if g.objects[id].Type != AigAND || !marking.Marked[id] {
			continue
		}
		if e := visit(); e != nil {
			return nil, e
		}
		leaves, e := collectFastCNFLeaves(g, AigID(id), marking.Marked, false, visit)
		if e != nil {
			return nil, e
		}
		nodes, e := fastCNFVolume(g, AigID(id), leaves, visit)
		if e != nil {
			return nil, e
		}
		isAND := true
		for _, node := range nodes {
			n := g.objects[node]
			if n.Fanin0.Complemented() && !marking.Marked[n.Fanin0.ID()] || n.Fanin1.Complemented() && !marking.Marked[n.Fanin1.ID()] {
				isAND = false
				break
			}
		}
		out := SatLit(r.objectVariables[id]) << 1
		if isAND {
			signed, e := collectFastCNFLeaves(g, AigID(id), marking.Marked, true, visit)
			if e != nil {
				return nil, e
			}
			big := []SatLit{out}
			for _, l := range signed {
				v, e := literal(l)
				if e != nil {
					return nil, e
				}
				big = append(big, v.Not())
			}
			if e := add(big); e != nil {
				return nil, e
			}
			for _, l := range signed {
				v, e := literal(l)
				if e != nil {
					return nil, e
				}
				if e := add([]SatLit{out.Not(), v}); e != nil {
					return nil, e
				}
			}
			continue
		}
		if len(leaves) > 6 || len(nodes) == 0 || len(leaves)+len(nodes) > 256 {
			return nil, ErrLimit
		}
		ids := make([]AigID, len(leaves))
		vars := make([]SatVar, len(leaves))
		for i, l := range leaves {
			ids[i] = l.ID()
			v, e := literal(l)
			if e != nil {
				return nil, e
			}
			vars[i] = v.Var()
		}
		truth, e := deriveFastCNFTruth(g, ids, nodes, visit)
		if e != nil {
			return nil, e
		}
		if truth == 0 || truth == ^uint64(0) {
			l := out
			if truth == 0 {
				l = l.Not()
			}
			if e := add([]SatLit{l}); e != nil {
				return nil, e
			}
			continue
		}
		for polarity := 0; polarity < 2; polarity++ {
			word := truth
			head := out
			if polarity == 1 {
				word = ^word
				head = head.Not()
			}
			table, e := TruthTableFromWords(len(leaves), []uint64{word}, 0)
			if e != nil {
				return nil, e
			}
			cover, e := table.ISOP(c.ISOP)
			if e != nil {
				return nil, e
			}
			for _, cube := range cover.Cubes {
				tail, e := CNFISOPCubeLiterals(cube, vars)
				if e != nil {
					return nil, e
				}
				clause := append([]SatLit{head}, tail...)
				if e := add(clause); e != nil {
					return nil, e
				}
			}
		}
	}
	for i, id := range g.cos {
		driver, e := literal(g.objects[id].Fanin0)
		if e != nil {
			return nil, e
		}
		if i < len(g.cos)-exposedOutputs {
			if e := add([]SatLit{driver}); e != nil {
				return nil, e
			}
			r.outputs = append(r.outputs, driver)
		} else {
			out := SatLit(r.objectVariables[id]) << 1
			if e := add([]SatLit{out, driver.Not()}); e != nil {
				return nil, e
			}
			if e := add([]SatLit{out.Not(), driver}); e != nil {
				return nil, e
			}
			r.outputs = append(r.outputs, out)
		}
	}
	constant := SatLit(r.objectVariables[0]) << 1
	if e := add([]SatLit{constant}); e != nil {
		return nil, e
	}
	if stop() {
		return nil, ErrCancelled
	}
	return r, nil
}
