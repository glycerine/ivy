// Ported to Go from concept_alpha.py.
//
// Alpha implements the alpha abstraction for concept domains:
// given a concept domain and a state formula, it determines for each
// concept fact whether it is necessarily true in the state.
package webui

import (
	"fmt"

	"github.com/glycerine/ivy/goivy/logic"
	solver "github.com/glycerine/ivy/goivy/z3bridge"
)

// Alpha computes the alpha abstraction of a concept domain against a state formula.
//
// For each fact (tag, formula) produced by domain.GetFacts(projection),
// it checks whether the state formula implies the fact formula using Z3.
// Results that are in the cache are reused without invoking the solver.
//
// The cache maps TagString(tag) -> bool and is updated in-place.
//
// Returns a list of (Tag, bool) pairs.
func Alpha(domain *CDConceptDomain, state logic.Expr, cache map[string]bool, projection func(string, string) bool) []TagValue {
	facts := domain.GetFacts(projection)

	if cache == nil {
		cache = make(map[string]bool)
	}

	slv := solver.NewSolver(nil, nil)
	var result []TagValue

	for _, fact := range facts {
		key := TagString(fact.Tag)
		if value, ok := cache[key]; ok {
			result = append(result, TagValue{Tag: fact.Tag, Value: value})
			continue
		}

		// Check: state => formula?
		// Equivalently: is (state & ~formula) unsatisfiable?
		// Recover from Z3 panics (sort mismatches, etc.) so one bad
		// formula doesn't crash the web server.
		value := false
		func() {
			defer func() {
				if r := recover(); r != nil {
					fmt.Printf("Alpha: Z3 panic for tag %v: %v\n", fact.Tag, r)
				}
			}()
			if fact.Formula == nil || state == nil {
				return
			}
			// Skip formulas that still contain TopSort — Z3 will panic.
			if logic.ContainsTopSort(fact.Formula) {
				return
			}
			notF, err := logic.NewNot(fact.Formula)
			if err != nil {
				return
			}
			conj, err2 := logic.NewAnd(state, notF)
			if err2 != nil {
				return
			}
			sat, err3 := slv.IsSat(conj)
			if err3 != nil {
				return
			}
			value = !sat // if unsat, state implies formula
		}()

		cache[key] = value
		result = append(result, TagValue{Tag: fact.Tag, Value: value})
	}

	return result
}

// AlphaNoSolver computes alpha abstraction without Z3, using only the cache.
// Facts not in the cache default to false.
func AlphaNoSolver(domain *CDConceptDomain, cache map[string]bool, projection func(string, string) bool) []TagValue {
	facts := domain.GetFacts(projection)
	var result []TagValue
	for _, fact := range facts {
		key := TagString(fact.Tag)
		value := false
		if cache != nil {
			if v, ok := cache[key]; ok {
				value = v
			}
		}
		result = append(result, TagValue{Tag: fact.Tag, Value: value})
	}
	return result
}
