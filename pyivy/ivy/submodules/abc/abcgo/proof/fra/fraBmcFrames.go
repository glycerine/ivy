package fra

import (
	"github.com/glycerine/abcgo/aig/aig"
	"github.com/glycerine/abcgo/proof/framgr"
)

// Use the same frame unroller as the shared FRA manager.
func fraUnrollInitialized(original *aig.Graph, frames int, keepOutputs bool) (*aig.Graph, []aig.Edge) {
	return framgr.UnrollInitialized(original, frames, keepOutputs)
}
