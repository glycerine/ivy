package ivy2cpp

import (
	"fmt"

	"github.com/glycerine/ivy/goivy"
)

const (
	experimentalUninterpretedSortLower = 0
	experimentalUninterpretedSortUpper = 100

	experimentalUninterpretedSortLowerText = "0"
	experimentalUninterpretedSortUpperText = "100"
)

func (g *Generator) experimentalUninterpretedRangeFor(s goivy.Sort) (*goivy.RangeSort, bool) {
	us, ok := g.experimentalUninterpretedSort(s)
	if !ok {
		return nil, false
	}
	g.warnExperimentalUninterpretedSort(us.Name)
	return &goivy.RangeSort{
		Name: us.Name,
		Lb:   goivy.NumeralBound{Value: experimentalUninterpretedSortLowerText},
		Ub:   goivy.NumeralBound{Value: experimentalUninterpretedSortUpperText},
	}, true
}

func (g *Generator) experimentalUninterpretedBounds(s goivy.Sort) (string, string, bool) {
	if _, ok := g.experimentalUninterpretedRangeFor(s); !ok {
		return "", "", false
	}
	return experimentalUninterpretedSortLowerText, experimentalUninterpretedSortUpperText, true
}

func (g *Generator) experimentalUninterpretedCardinality(s goivy.Sort) (int, bool) {
	if _, ok := g.experimentalUninterpretedRangeFor(s); !ok {
		return 0, false
	}
	return experimentalUninterpretedSortUpper - experimentalUninterpretedSortLower + 1, true
}

func (g *Generator) experimentalUninterpretedSort(s goivy.Sort) (*goivy.UninterpretedSort, bool) {
	if g == nil || !g.usesZ3() {
		return nil, false
	}
	us, ok := s.(*goivy.UninterpretedSort)
	if !ok || us.Name == "" {
		return nil, false
	}
	if _, ok := g.rangeSortFor(s); ok {
		return nil, false
	}
	if _, ok := g.cppInterpType(s); ok {
		return nil, false
	}
	if g.isRecordRange(s) || g.hasStringInterp(s) {
		return nil, false
	}
	if _, ok := g.variantSubtypeName(s); ok {
		return nil, false
	}
	return us, true
}

func (g *Generator) warnExperimentalUninterpretedSort(name string) {
	if g == nil || name == "" {
		return
	}
	g.warnOnce(fmt.Sprintf(
		"ivy2cpp: warning: assuming uninterpreted sort %s has experimental finite range 0..100",
		name,
	))
}

func (g *Generator) warnOnce(msg string) {
	if g == nil || msg == "" {
		return
	}
	if g.warningSet == nil {
		g.warningSet = map[string]bool{}
	}
	if g.warningSet[msg] {
		return
	}
	g.warningSet[msg] = true
	g.warnings = append(g.warnings, msg)
}
