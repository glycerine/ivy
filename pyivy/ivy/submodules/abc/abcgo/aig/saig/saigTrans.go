package saig

import (
	"fmt"
	"time"

	"github.com/glycerine/abcgo/aig/aig"
)

// Saig_ManTimeframeSimplify finds equivalences in an arbitrary-state unrolling
// and uses them to simplify a longer sequence of frames.
func Saig_ManTimeframeSimplify(g *aig.Graph, nFrames, nFramesMax, init, verbose int32) *aig.Graph {
	if g == nil || nFrames <= 0 || nFrames > nFramesMax || Saig_ManRegNum(g) <= 0 {
		return nil
	}
	frames := Saig_ManFramesNonInitial(g, nFrames)
	if frames == nil {
		return nil
	}
	start := time.Now()
	// A single partition gives the existing exact equivalence engine the whole
	// unrolling and leaves representatives on frames for the mapping pass.
	aig.Aig_ManFraigPartitioned(frames, int32(max(1, frames.OutputCount())), 1000, 0, 0)
	if verbose != 0 {
		fmt.Printf("Frame equivalence: %s\n", time.Since(start))
	}
	Saig_ManCreateMapping(g, frames, nFrames)
	Saig_ManStopMap1(g)
	defer Saig_ManStopMap2(g)
	return Saig_ManFramesInitialMapped(g, nFrames, nFramesMax, init)
}
