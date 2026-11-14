package bmc

import "github.com/glycerine/abcgo/aig/aig"

// Saig_ManFramesBmc constructs a combinational frame unrolling whose outputs
// are the property outputs from each frame in time order.
func Saig_ManFramesBmc(graph *aig.Graph, frames int32) *aig.Graph {
	if graph == nil || frames <= 0 {
		return nil
	}
	return aig.Aig_ManFrames(graph, frames, 1, 1, 0, 0, nil)
}

// Saig_ManBmcSimple keeps the legacy API and status convention while running
// through the shared native unroller, exact CNF encoder, and BSAT solver.
// Rewrite, cofactoring, node-limit, and alternate-backend knobs are accepted
// for source compatibility; the common implementation uses the full graph.
func Saig_ManBmcSimple(graph *aig.Graph, frames, sizeMax, conflictLimit, rewrite, verbose int32, failedFrame *int32, cofactorFanins, useSatoko int32) int32 {
	if graph == nil || frames < 0 {
		return -1
	}
	params := Saig_ParBmc_t{NFramesMax: frames, NConfLimit: conflictLimit, FVerbose: verbose}
	status := Saig_ManBmcScalable(graph, &params)
	if failedFrame != nil {
		*failedFrame = frames
		if status == 0 {
			*failedFrame = params.IFrame
		}
	}
	_ = sizeMax
	_ = rewrite
	_ = cofactorFanins
	_ = useSatoko
	if status == 0 {
		return 0
	}
	return -1
}
