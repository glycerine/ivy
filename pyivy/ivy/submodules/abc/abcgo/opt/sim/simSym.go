package sim

import "github.com/glycerine/abcgo/aig/aig"

// Symmetries returns exact per-output positive (ordinary variable-swap)
// symmetries on functional-support variables. Non-support variables are omitted.
func Symmetries(g *aig.Graph, verbose bool) []Matrix {
	p := Sym_ManStart(g, verbose)
	Sim_SymmsStructCompute(g, p.Symmetric, p.Functional)
	pattern := make([]uint32, p.Words)
	// Source uses 1000 single-flip squares before exact SAT completion.
	for i := 0; i < 1000; i++ {
		fillRandom(pattern, p.rng)
		Sim_SymmsSimulate(p, pattern, p.NonSymmetric)
	}
	for Sim_SymmsGetPatternUsingSat(p, pattern) {
		for i := 0; i < 4; i++ {
			Sim_SymmsSimulate(p, pattern, p.NonSymmetric)
			v := p.Var1
			if i&1 != 0 {
				v = p.Var2
			}
			Sim_XorBit(pattern, v)
		}
	}
	if !Sim_UtilMatrsAreDisjoint(p) {
		panic("inconsistent symmetry proof and witness")
	}
	if verbose {
		Sym_ManPrintStats(p)
	}
	return p.Symmetric
}
func Sim_ComputeTwoVarSymms(g *aig.Graph, verbose bool) int {
	m := Symmetries(g, verbose)
	vars := make([]int, g.InputCount())
	for i := range vars {
		vars[i] = i
	}
	total := 0
	for _, matrix := range m {
		total += Sim_UtilCountPairsOne(matrix, vars)
	}
	return total
}
