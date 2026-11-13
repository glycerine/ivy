package saig

import (
	"fmt"
	"time"

	"github.com/glycerine/abcgo/aig/aig"
)

// Saig_ManTimeframeSimplify finds equivalences in an arbitrary-state unrolling
// and uses them to simplify a longer sequence of frames.
func Saig_ManTimeframeSimplify(g *aig.Graph, nFrames, nFramesMax, init, verbose int32, fraig func(*aig.Graph, int32, int32) *aig.Graph) *aig.Graph {
	return Saig_ManTimeframeSimplifyWithFraig(g, nFrames, nFramesMax, init, verbose, fraig)
}

// Saig_ManTimeframeSimplifyWithFraig uses FRA's equivalence sweep on the
// uninitialized frames. The backend must record representatives on its input.
func Saig_ManTimeframeSimplifyWithFraig(g *aig.Graph, nFrames, nFramesMax, init, verbose int32, fraig func(*aig.Graph, int32, int32) *aig.Graph) *aig.Graph {
	if g == nil || nFrames <= 0 {
		return nil
	}
	if Saig_ManRegNum(g) <= 0 {
		panic("saig: timeframe simplification requires registers")
	}
	if fraig == nil {
		panic("saig: timeframe simplification requires a FRA backend")
	}
	frames := Saig_ManFramesNonInitial(g, nFrames)
	if frames == nil {
		return nil
	}
	start := time.Now()
	reduced := fraig(frames, 1000, 0)
	if reduced == nil {
		panic("saig: FRA backend returned nil")
	}
	if !aig.Aig_ManHasReprs(frames) {
		panic("saig: FRA backend did not provide representatives")
	}
	if verbose != 0 {
		aig.Aig_ManPrintStats(frames)
		aig.Aig_ManPrintStats(reduced)
		fmt.Printf("Fraiging =%9.2f sec\n", time.Since(start).Seconds())
	}
	Saig_ManCreateMapping(g, frames, nFrames)
	Saig_ManStopMap1(g)
	defer func() {
		if Saig_ManHasMap2(g) != 0 {
			Saig_ManStopMap2(g)
		}
	}()
	if nFrames > nFramesMax {
		panic("saig: initial frame count exceeds maximum")
	}
	start = time.Now()
	result := Saig_ManFramesInitialMapped(g, nFrames, nFramesMax, init)
	fmt.Printf("Mapped =%9.2f sec\n", time.Since(start).Seconds())
	Saig_ManStopMap2(g)
	start = time.Now()
	normal := Saig_ManFramesInitialMapped(g, nFrames, nFramesMax, init)
	fmt.Printf("Normal =%9.2f sec\n", time.Since(start).Seconds())
	if verbose != 0 {
		aig.Aig_ManPrintStats(normal)
		aig.Aig_ManPrintStats(result)
	}
	return result
}
