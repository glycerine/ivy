package gia

import (
	"fmt"
	"slices"
	"testing"
)

func TestFramesRetainedRegistersKeepInitialStates(t *testing.T) {
	for frames := int32(1); frames <= 3; frames++ {
		for disable := int32(0); disable < 2; disable++ {
			for or := int32(0); or < 2; or++ {
				t.Run(fmt.Sprintf("f%d/disable%d/or%d", frames, disable, or), func(t *testing.T) {
					g := framesReviewGenerated(19)
					defer Gia_ManStop(g)
					if err := g.SetAigerLatchInitialStates([]int32{1, -1}); err != nil {
						t.Fatal(err)
					}
					out := Gia_ManFrames(g, &Gia_ParFra_t{NFrames: frames, FDisableSt: disable, FOrPos: or})
					defer Gia_ManStop(out)
					if got := out.AigerLatchInitialStates(); !slices.Equal(got, []int32{1, -1}) {
						t.Fatalf("retained register order lost initial cube: %v", got)
					}
					out.aigerLatchInits[0] = 0
					if !slices.Equal(g.AigerLatchInitialStates(), []int32{1, -1}) {
						t.Fatal("returned reset storage aliases source")
					}
				})
			}
		}
	}
}
