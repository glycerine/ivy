package core

import "time"
import "github.com/glycerine/abcgo/aig/aig"

// RewriteStats counts actual cut searches and committed replacements.
type RewriteStats struct {
	Roots, Cuts, CutsBad, Replacements, NodesSaved, PowerRejected int
	// Dar_Man_t's counters for Dar_ManPrintStats
	NodesInit, NodesTried, CutsAll, CutsTried, CutsUsed, CutsSkipped, CutMemUsed int
	TimeCuts, TimeEval, TimeTotal                                                time.Duration
}

// Manager owns mutable state for a single rewriting run; library DAGs are shared.
type Manager struct {
	Graph         *aig.Graph
	Params        RewriteParams
	Stats         RewriteStats
	library       *PreparedLibrary
	probabilities []float32
}
type Dar_Man_t = Manager
