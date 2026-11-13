package framgr

import "github.com/glycerine/abcgo/aig/aig"

// CombOptions carries every FRA manager parameter across the package boundary.
type CombOptions struct {
	NSimWords       int32
	FPatScores      int32
	MaxScore        int32
	DSimSatur       float64
	DActConeRatio   float64
	DActConeBumpMax float64
	FChoicing       int32
	FSpeculate      int32
	FProve          int32
	FVerbose        int32
	FDoSparse       int32
	FConeBias       int32
	NBTLimitNode    int32
	NBTLimitMiter   int32
	NLevelMax       int32
	NFramesP        int32
	NFramesK        int32
	NMaxImps        int32
	NMaxLevs        int32
	FRewrite        int32
	FLatchCorr      int32
	FUseImps        int32
	FUse1Hot        int32
	FWriteImps      int32
	FDontShowBar    int32
}

// Sweep runs FRA's combinational manager using the supplied configuration.
func Sweep(graph *aig.Graph, options CombOptions) *aig.Graph {
	result, _ := SweepMapped(graph, options)
	return result
}

// SweepMapped additionally returns the source-to-result node mapping.
func SweepMapped(graph *aig.Graph, options CombOptions) (*aig.Graph, []aig.Edge) {
	return SweepMappedCallback(graph, options, nil)
}

// SweepMappedCallback invokes callback on the live manager before finalizing
// the sweep. Mutations to the manager are visible to the remaining work.
func SweepMappedCallback(graph *aig.Graph, options CombOptions, callback func(*Manager)) (*aig.Graph, []aig.Edge) {
	return fraFraigPerformMapped(graph, &options, callback)
}
