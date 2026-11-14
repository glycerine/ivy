package bsat

import "github.com/glycerine/abcgo/aig/aig"

// ProofInterpolant follows Sat_ProofInterpolant in satProof.c. Local maps replace
// its temporary clause/proof Id annotations, leaving the stored proof intact.
func (p *Solver2) ProofInterpolant(globals []int32) *aig.Graph {
	s := p.solver
	if s.hProofLast == -1 {
		return nil
	}
	if s.pPrf1 == nil || s.pInt2 != nil {
		panic("stored resolution proof unavailable")
	}
	positions := make(map[int32]int32, len(globals))
	result := aig.NewGraph()
	result.Name = "interpol"
	for i, v := range globals {
		if v < 0 || v >= s.size {
			panic("interpolant variable outside solver")
		}
		positions[v] = int32(i)
		result.NewInput()
	}
	inputs := result.Inputs()
	used := Proof_CollectUsedIter(s.pPrf1, []int32{s.hProofLast}, 1)
	defer Proof_CleanCollected(s.pPrf1, used)
	clauses := make(map[int32]*clause)
	for _, handle := range s.Mem.handles(0) {
		c := clause2_read(s, handle)
		clauses[clause_id(c)] = c
	}
	roots := make(map[int32]aig.Edge)
	for _, id := range Sat_ProofCollectCore(s.pPrf1, used) {
		c := clauses[id]
		if c == nil {
			panic("resolution proof references missing root clause")
		}
		value := result.Const1()
		if c.partA != 0 {
			value = result.Const0()
			for _, literal := range c.lits[:c.size] {
				if position, ok := positions[lit_var(literal)]; ok {
					variable := inputs[position]
					if lit_sign(literal) != 0 {
						variable = variable.Not()
					}
					value = result.OR(value, variable)
				}
			}
		}
		roots[id] = value
	}
	values := make(map[int32]aig.Edge, len(used))
	for _, handle := range used {
		entries := Proof_NodeRead(s.pPrf1, handle).pEnts
		if len(entries) == 0 {
			panic("empty resolution proof node")
		}
		var value aig.Edge
		for i, entry := range entries {
			table := values
			if entry&1 != 0 {
				table = roots
			}
			child, ok := table[entry>>2]
			if !ok {
				panic("resolution antecedent is not available")
			}
			if i == 0 {
				value = child
			} else if entry&2 != 0 {
				value = result.OR(value, child)
			} else {
				value = result.AND(value, child)
			}
		}
		values[handle] = value
	}
	output, ok := values[s.hProofLast]
	if !ok {
		panic("missing final resolution node")
	}
	result.AddOutput(output)
	aig.Aig_ManCleanup(result)
	return result
}
