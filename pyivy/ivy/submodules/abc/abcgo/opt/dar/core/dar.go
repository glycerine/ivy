// Package dar performs DAG-aware AIG rewriting, balancing and refactoring.
// Graph edges and cut boundaries are typed values; immutable rewrite-library
// data is shared while each optimization run owns its mutable search state.
package core
import "context"

// RewriteParams controls four-input cut rewriting. Defaults match ABC's rw
// command; Context provides the standard-library equivalent of Time2Quit.
type RewriteParams struct {
	CutsMax, SubgraphsMax                         int
	Fanout, UpdateLevel, UseZeros, Power, Recycle bool
	Verbose, VeryVerbose                          bool
	Context                                       context.Context
}
type RefactorParams struct {
	MffcMin, LeafMax, CutsMax                           int
	Extend, UpdateLevel, UseZeros, Verbose, VeryVerbose bool
	Context                                             context.Context
}
type Dar_RwrPar_t = RewriteParams
type Dar_RefPar_t = RefactorParams

func DefaultRewriteParams() RewriteParams {
	return RewriteParams{CutsMax: 8, SubgraphsMax: 5, Fanout: true, Recycle: true}
}
func Dar_ManDefaultRwrParams(p *RewriteParams) { *p = DefaultRewriteParams() }
func DefaultRefactorParams() RefactorParams {
	return RefactorParams{MffcMin: 2, LeafMax: 12, CutsMax: 5}
}
func Dar_ManDefaultRefParams(p *RefactorParams) { *p = DefaultRefactorParams() }
