package ivy2cpp

import "github.com/glycerine/ivy/goivy"

func (g *Generator) experimentalUninterpretedRangeFor(s goivy.Sort) (*goivy.RangeSort, bool) {
	return nil, false
}

func (g *Generator) experimentalUninterpretedBounds(s goivy.Sort) (string, string, bool) {
	return "", "", false
}

func (g *Generator) experimentalUninterpretedCardinality(s goivy.Sort) (int, bool) {
	return 0, false
}

func (g *Generator) experimentalUninterpretedSort(s goivy.Sort) (*goivy.UninterpretedSort, bool) {
	return nil, false
}
