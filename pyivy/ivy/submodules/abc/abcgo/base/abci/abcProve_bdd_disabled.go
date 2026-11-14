//go:build !abc_use_cudd

package abci

const proveBddEnabled = false

// The upstream C branch is guarded by ABC_USE_CUDD. The Go build exposes the
// matching optional behavior with the explicit abc_use_cudd build tag.
func proveBddFallback(n *Abc_Ntk_t, p *Prove_Params_t) (*Abc_Ntk_t, int32, bool) {
	return n, -1, false
}
