package live

import "github.com/glycerine/abcgo/aig/aig"

// The simulation variant permits SAVE to restart shadow capture and event
// accumulation, while the proof-oriented variant captures only once.
func LivenessToSafetyTransformationSim(n *Network, p *aig.Graph, live, fair []aig.Edge) *aig.Graph {
	result, err := livenessToSafety(sourceNetwork(n, p), TransformOptions{Mode: IGNORE_SAFETY_KEEP_LIVENESS_MODE, Liveness: live, Fairness: fair, Restartable: true}, true, false)
	if err != nil {
		panic(err)
	}
	return result.Graph
}
func LivenessToSafetyTransformationOneStepLoopSim(n *Network, p *aig.Graph, live, fair []aig.Edge) *aig.Graph {
	result, err := livenessToSafety(sourceNetwork(n, p), TransformOptions{Mode: IGNORE_SAFETY_KEEP_LIVENESS_ONE_LOOP_MODE, Liveness: live, Fairness: fair, Restartable: true}, true, false)
	if err != nil {
		panic(err)
	}
	return result.Graph
}
func Abc_CommandAbcLivenessToSafetySim(frame *Frame, argc int32, argv []string) int32 {
	return commandTransform(frame, argv, false, true, false)
}
