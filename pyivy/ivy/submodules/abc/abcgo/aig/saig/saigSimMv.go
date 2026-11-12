// Multivalued simulation APIs, ported from ABC saigSimMv.c.
package saig

import "github.com/glycerine/abcgo/aig/aig"

// The Graph-level implementation is shared with AIG sequential sweeping, so
// callers get the same symbolic-state semantics without a package import cycle.
type Saig_MvMan_t = aig.MvManager

const (
	SAIG_DIFF_VALUES = aig.SAIG_DIFF_VALUES
	SAIG_UNDEF_VALUE = aig.SAIG_UNDEF_VALUE
)

func Saig_MvManStart(p *aig.Graph, satur int32) *Saig_MvMan_t { return aig.Aig_MvManStart(p, satur) }
func Saig_MvManStop(p *Saig_MvMan_t)                          { aig.Aig_MvManStop(p) }
func Saig_MvPrintState(iter int32, p *Saig_MvMan_t)           { aig.Aig_MvPrintState(iter, p) }
func Saig_MvSimulateFrame(p *Saig_MvMan_t, first, verbose int32) {
	aig.Aig_MvSimulateFrame(p, first, verbose)
}
func Saig_MvSimHash(state []uint32, nFlops, nBins int32) int32 {
	return aig.Aig_MvSimHash(state, nFlops, nBins)
}
func Saig_MvSaveState(p *Saig_MvMan_t) int32             { return aig.Aig_MvSaveState(p) }
func Saig_MvManPostProcess(p *Saig_MvMan_t, state int32) { aig.Aig_MvManPostProcess(p, state) }
func Saig_MvManFindXFlops(p *Saig_MvMan_t) []int32       { return aig.Aig_MvManFindXFlops(p) }
func Saig_MvManCheckOscilator(p *Saig_MvMan_t, index int32) int32 {
	return aig.Aig_MvManCheckOscilator(p, index)
}
func Saig_MvManFindConstBinaryFlops(p *Saig_MvMan_t, binary *[]int32) []int32 {
	return aig.Aig_MvManFindConstBinaryFlops(p, binary)
}
func Saig_MvManFindOscilators(p *Saig_MvMan_t, constant *[]int32) []int32 {
	return aig.Aig_MvManFindOscilators(p, constant)
}
func Saig_MvManCreateNextSkip(p *Saig_MvMan_t) []int32 { return aig.Aig_MvManCreateNextSkip(p) }
func Saig_MvManDeriveMap(p *Saig_MvMan_t, verbose int32) []aig.Edge {
	return aig.Aig_MvManDeriveMap(p, verbose)
}
func Saig_MvManSimulate(p *aig.Graph, symbolic, satur, verbose, veryVerbose int32) []aig.Edge {
	return aig.Aig_MvManSimulate(p, symbolic, satur, verbose, veryVerbose)
}
