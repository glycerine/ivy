package cudd

import (
	"math"

	"github.com/glycerine/abcgo/bdd/cudd/core"
)

// Cudd_bddCorrelation returns the fraction of assignments on which f and g
// agree, assuming every manager variable is independently equiprobable.
func Cudd_bddCorrelation(dd *DdManager, f, g *DdNode) float64 {
	if dd == nil || f == nil || g == nil {
		return math.NaN()
	}
	difference := core.Cudd_bddXor(dd, f, g)
	if difference == nil {
		return math.NaN()
	}
	nvars := Cudd_ReadSize(dd)
	if nvars == 0 {
		if difference == Cudd_ReadLogicZero(dd) {
			return 1
		}
		return 0
	}
	return 1 - core.Cudd_CountMinterm(dd, difference, int(nvars))/math.Ldexp(1, int(nvars))
}

// Cudd_bddCorrelationWeights computes agreement probability under independent
// input probabilities. prob[i] is the probability that variable i is true.
func Cudd_bddCorrelationWeights(dd *DdManager, f, g *DdNode, prob []float64) float64 {
	if dd == nil || f == nil || g == nil || len(prob) < int(Cudd_ReadSize(dd)) {
		return math.NaN()
	}
	difference := core.Cudd_bddXor(dd, f, g)
	if difference == nil {
		return math.NaN()
	}
	values := prob[:Cudd_ReadSize(dd)]
	for _, p := range values {
		if math.IsNaN(p) || p < 0 || p > 1 {
			return math.NaN()
		}
	}
	return 1 - weightedTrueProbability(difference, values, map[*DdNode]float64{})
}

func weightedTrueProbability(f *DdNode, probabilities []float64, memo map[*DdNode]float64) float64 {
	if Cudd_IsConstant(f) {
		truth := Cudd_Regular(f).Value != 0
		if Cudd_IsComplement(f) {
			truth = !truth
		}
		if truth {
			return 1
		}
		return 0
	}
	if value, ok := memo[f]; ok {
		return value
	}
	regular := Cudd_Regular(f)
	thenBranch, elseBranch := Cudd_T(regular), Cudd_E(regular)
	if Cudd_IsComplement(f) {
		thenBranch, elseBranch = Cudd_Not(thenBranch), Cudd_Not(elseBranch)
	}
	index := int(Cudd_NodeReadIndex(f))
	if index < 0 || index >= len(probabilities) {
		return math.NaN()
	}
	p := probabilities[index]
	value := p*weightedTrueProbability(thenBranch, probabilities, memo) + (1-p)*weightedTrueProbability(elseBranch, probabilities, memo)
	memo[f] = value
	return value
}
