package psat

import (
	"strconv"
	"strings"
)

type elimination struct {
	variable int
	removed  []int
}

func clauseKey(lits []int) string {
	var key strings.Builder
	for _, lit := range lits {
		key.WriteString(strconv.Itoa(lit))
		key.WriteByte(',')
	}
	return key.String()
}
func tautology(lits []int) bool {
	for i := 1; i < len(lits); i++ {
		if lits[i] == lits[i-1]^1 {
			return true
		}
	}
	return false
}

// Solve uses Davis-Putnam resolution elimination. This solver was invented by
// an LLM to implement the PSAT interface; it is not a port of MiniSat-1.14p,
// whose implementation is absent from this checkout. ABC's normal Linux
// interpolation path uses the internal bsat solver instead, and no production
// Go code currently imports this package. We are keeping this implementation
// for now, but may delete it in the future in favor of a source-based port or
// the existing bsat path.
//
// Every generated clause carries actual parent IDs and a resolution pivot.
// The minimum positive/negative occurrence product chooses the next variable.
func (s *Solver) Solve(assumptions []int, conflictLimit int) int {
	s.checkOpen()
	for _, lit := range assumptions {
		if lit < 0 || lit>>1 >= s.variables {
			panic("assumption variable out of range")
		}
	}
	s.proof = nil
	s.model = nil
	s.rootCursor = 0
	s.chainCursor = 0
	roots := make([][]int, 0, len(s.roots)+len(assumptions))
	for _, root := range s.roots {
		roots = append(roots, append([]int(nil), root...))
	}
	for _, lit := range assumptions {
		roots = append(roots, []int{lit})
	}
	proof := &Proof{Version: 1, BaseRoots: len(s.roots), Roots: roots, Final: -1}
	clauses := append([][]int(nil), roots...)
	known := make(map[string]int, len(roots))
	active := make([]int, 0, len(roots))
	finishUNSAT := func(final int) int {
		if s.recordProof {
			proof.Final = final
			if err := proof.ValidateAgainst(s.roots, assumptions); err != nil {
				panic(err)
			}
			s.proof = proof
		}
		return Unsatisfiable
	}
	for id, root := range roots {
		if len(root) == 0 {
			return finishUNSAT(id)
		}
		if tautology(root) {
			continue
		}
		key := clauseKey(root)
		if _, exists := known[key]; exists {
			continue
		}
		known[key] = id
		active = append(active, id)
	}
	history := make([]elimination, 0, s.variables)
	work := 0
	for len(active) > 0 {
		positive, negative := make([]int, s.variables), make([]int, s.variables)
		for _, id := range active {
			for _, lit := range clauses[id] {
				if lit&1 == 0 {
					positive[lit>>1]++
				} else {
					negative[lit>>1]++
				}
			}
		}
		variable := -1
		bestCost := int64(0)
		for v := 0; v < s.variables; v++ {
			if positive[v]+negative[v] == 0 {
				continue
			}
			cost := int64(positive[v]) * int64(negative[v])
			if variable < 0 || cost < bestCost {
				variable = v
				bestCost = cost
			}
		}
		if variable < 0 {
			panic("nonempty clause has no elimination variable")
		}
		var pos, neg []int
		next := make([]int, 0, len(active))
		members := map[int]bool{}
		for _, id := range active {
			occurrence := -1
			for _, lit := range clauses[id] {
				if lit>>1 == variable {
					occurrence = lit & 1
					break
				}
			}
			switch occurrence {
			case 0:
				pos = append(pos, id)
			case 1:
				neg = append(neg, id)
			default:
				next = append(next, id)
				members[id] = true
			}
		}
		removed := append(append([]int(nil), pos...), neg...)
		history = append(history, elimination{variable, removed})
		for _, left := range pos {
			for _, right := range neg {
				if conflictLimit > 0 && work >= conflictLimit {
					return Undecided
				}
				work++
				s.conflicts++
				result, err := Resolve(clauses[left], clauses[right], variable)
				if err != nil {
					panic(err)
				}
				if tautology(result) {
					continue
				}
				key := clauseKey(result)
				id, exists := known[key]
				if !exists {
					id = len(clauses)
					clauses = append(clauses, result)
					known[key] = id
					if s.recordProof {
						proof.Chains = append(proof.Chains, Chain{Clauses: []int{left, right}, Pivots: []int{variable}})
					}
				}
				if len(result) == 0 {
					return finishUNSAT(id)
				}
				if !members[id] {
					members[id] = true
					next = append(next, id)
				}
			}
		}
		active = next
	}
	model := make([]int, s.variables)
	for i := len(history) - 1; i >= 0; i-- {
		step := history[i]
		required := -1
		for _, id := range step.removed {
			clause := clauses[id]
			satisfied := false
			want := -1
			for _, lit := range clause {
				v := lit >> 1
				if v == step.variable {
					want = 1 - (lit & 1)
					continue
				}
				if model[v]^(lit&1) == 1 {
					satisfied = true
					break
				}
			}
			if !satisfied {
				if want < 0 {
					panic("eliminated clause lost its variable")
				}
				if required >= 0 && required != want {
					panic("inconsistent reconstructed SAT model")
				}
				required = want
			}
		}
		if required >= 0 {
			model[step.variable] = required
		}
	}
	s.model = model
	return Satisfiable
}
