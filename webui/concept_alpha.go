// Copyright (c) Microsoft Corporation. All Rights Reserved.
// Ported to Go from concept_alpha.py.
//
// Alpha implements the alpha abstraction for concept domains:
// given a concept domain and a state formula, it determines for each
// concept fact whether it is necessarily true in the state.
package webui

import (
	"github.com/glycerine/goivy/logic"
	"github.com/glycerine/goivy/solver"
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
func Alpha(domain *CDConceptDomain, state logic.Node, cache map[string]bool, projection func(string, string) bool) []TagValue {
	facts := domain.GetFacts(projection)

	if cache == nil {
		cache = make(map[string]bool)
	}

	slv := solver.New()
	var result []TagValue

	for _, fact := range facts {
		key := TagString(fact.Tag)
		if value, ok := cache[key]; ok {
			result = append(result, TagValue{Tag: fact.Tag, Value: value})
			continue
		}

		// Check: state => formula?
		// Equivalently: is (state & ~formula) unsatisfiable?
		value := false
		notF, err := logic.NewNot(fact.Formula)
		if err == nil && state != nil {
			conj, err2 := logic.NewAnd(state, notF)
			if err2 == nil {
				sat, err3 := slv.IsSat(conj)
				if err3 == nil {
					value = !sat // if unsat, state implies formula
				}
			}
		}

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
