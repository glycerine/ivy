package fra

import (
	"github.com/glycerine/abcgo/aig/aig"
	"github.com/glycerine/abcgo/aig/saig"
)

// Aig_ManFraigPartitioned supplies ABC's FRA implementation to the
// cycle-free AIG partitioning routine.
func Aig_ManFraigPartitioned(g *aig.Graph, size, conflicts, level, verbose int32) *aig.Graph {
	return aig.Aig_ManFraigPartitioned(g, size, conflicts, level, verbose, Fra_FraigChoice)
}

// Aig_ManChoicePartitioned computes partition choices with ABC's FRA sweep.
func Aig_ManChoicePartitioned(graphs []*aig.Graph, size, conflicts, level, verbose int32) *aig.Graph {
	return aig.Aig_ManChoicePartitioned(graphs, size, conflicts, level, verbose, Fra_FraigChoice)
}

// Saig_ManTimeframeSimplify uses the original FRA equivalence sweep.
func Saig_ManTimeframeSimplify(g *aig.Graph, frames, maxFrames, init, verbose int32) *aig.Graph {
	return saig.Saig_ManTimeframeSimplify(g, frames, maxFrames, init, verbose, Fra_FraigEquivence)
}
