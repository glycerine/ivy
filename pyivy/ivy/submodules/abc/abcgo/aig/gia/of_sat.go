package gia

import (
	"github.com/glycerine/abcgo/misc/vec"
	"github.com/glycerine/abcgo/sat/satoko"
)

// ofAtMost builds a sequential counter: state[i][j] means that at least j+1
// of the first i+1 inputs are true. Only forward implications are needed for
// an upper-bound constraint.
func ofAtMost(s *satoko.Solver, variables []int, limit int) bool {
	if limit < 0 {
		return s.AddClause(nil)
	}
	if limit >= len(variables) {
		return true
	}
	if limit == 0 {
		for _, v := range variables {
			if !s.AddClause([]int{2*v + 1}) {
				return false
			}
		}
		return true
	}
	previous := make([]int, limit)
	for i, x := range variables {
		current := make([]int, limit)
		for j := range current {
			current[j] = s.AddVariable(false)
		}
		if !s.AddClause([]int{2*x + 1, 2 * current[0]}) {
			return false
		}
		if i > 0 {
			for j := range current {
				if !s.AddClause([]int{2*previous[j] + 1, 2 * current[j]}) {
					return false
				}
			}
			for j := 1; j < limit; j++ {
				if !s.AddClause([]int{2*x + 1, 2*previous[j-1] + 1, 2 * current[j]}) {
					return false
				}
			}
			if !s.AddClause([]int{2*x + 1, 2*previous[limit-1] + 1}) {
				return false
			}
		}
		previous = current
	}
	return true
}

// satProblem constructs the source mapper's node/cut implication problem and
// bounds the number of selected internal nodes. Variables use ABC literals.
func (p *ofMapper) satProblem(limit int) *satoko.Solver {
	g := p.cuts.graph
	ids := make([]int32, 0)
	indices := make(map[int32]int)
	for _, obj := range g.pObjs {
		if Gia_ObjIsAnd(obj) != 0 {
			indices[obj.id] = len(ids)
			ids = append(ids, obj.id)
		}
	}
	cutCount := 0
	for _, id := range ids {
		if p.cuts.sets[id] != 0 {
			cutCount += int(p.cuts.storage.at(p.cuts.sets[id])[0])
		}
	}
	s := satoko.New()
	s.SetVariables(len(ids) + cutCount)
	next := len(ids)
	for _, id := range ids {
		node := indices[id]
		var choices []int
		p.eachCut(id, func(_ int32, cut []int32) {
			v := next
			next++
			choices = append(choices, 2*v)
			s.AddClause([]int{2*v + 1, 2 * node})
			for pin := range ofStoredLeaves(cut) {
				if fanin, ok := indices[ofStoredVar(cut, pin)]; ok {
					s.AddClause([]int{2*v + 1, 2 * fanin})
				}
			}
		})
		s.AddClause(append([]int{2*node + 1}, choices...))
	}
	for _, id := range vec.Vec_IntArray(g.vCos) {
		co := Gia_ManObj(g, id)
		if node, ok := indices[Gia_ObjFaninId0(co, id)]; ok {
			s.AddClause([]int{2 * node})
		}
	}
	variables := make([]int, len(ids))
	for i := range variables {
		variables[i] = i
	}
	ofAtMost(s, variables, limit)
	return s
}
