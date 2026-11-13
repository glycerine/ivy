package framgr

import (
	"github.com/glycerine/abcgo/aig/aig"
	"github.com/glycerine/abcgo/misc/vec"
)

// SweepBMCImplications applies the original BMC implication callback while
// the shared manager still owns its live SAT solver and node mappings. The
// implication vector is updated in place, as in Fra_BmcFilterImplications.
func SweepBMCImplications(frames, source *aig.Graph, prefix, depth int32, mappings []aig.Edge, implications *vec.Vec_Int_t, limit int32) (*aig.Graph, []aig.Edge) {
	if frames == nil || implications == nil {
		panic("fra: invalid BMC implication arguments")
	}
	b := Fra_BmcStart(source, prefix, depth)
	if len(mappings) != len(b.pObjToFrames) {
		panic("fra: invalid BMC frame mapping")
	}
	b.pObjToFrames = mappings
	b.pAigFrames = frames
	b.vImps = implications
	defer Fra_BmcStop(b)
	return Fra_FraigEquivenceMapped(frames, limit, 0, func(manager *Fra_Man_t) { Fra_BmcFilterImplications(manager, b) })
}
