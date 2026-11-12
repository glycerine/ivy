package dar

import "github.com/glycerine/abcgo/aig/aig"

// RewriteStats counts actual cut searches and committed replacements.
type RewriteStats struct{ Roots, Cuts, Replacements, NodesSaved int }

// Manager owns mutable state for a single rewriting run; library DAGs are shared.
type Manager struct {
	Graph   *aig.Graph
	Params  RewriteParams
	Stats   RewriteStats
	library *PreparedLibrary
}
type Dar_Man_t = Manager
