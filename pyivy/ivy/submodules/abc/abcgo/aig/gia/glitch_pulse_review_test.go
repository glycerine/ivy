package gia

import "testing"

func TestGlitchReconvergentPulseCountsIndependent(t *testing.T) {
	for depth := 2; depth <= 8; depth++ {
		for _, gate := range []struct {
			name               string
			truth              uint64
			pulseUp, pulseDown int32
		}{{"AND", 8, 2, 0}, {"OR", 14, 0, 2}, {"XOR", 6, 2, 2}} {
			g := glitchReviewGraph{cis: 1}
			for i := 0; i < depth; i++ {
				truth := uint64(0xaaaaaaaaaaaaaaaa)
				if i == depth-1 {
					truth = 0x5555555555555555
				}
				g.nodes = append(g.nodes, glitchReviewNode{[]int32{int32(i)}, truth})
			}
			g.nodes = append(g.nodes, glitchReviewNode{[]int32{0, int32(depth)}, gate.truth})
			g.outputs = []int32{int32(depth + 1)}
			p := g.build()
			Gli_ManFinalize(p)
			var glitches int32
			for transition := 0; transition < 6; transition++ {
				Vec_IntClear(p.vCisChanged)
				gliFlip(p, p.objects[0])
				Gli_ManSwitching(p)
				Gli_ManGlitching(p)
				Gli_ManVerify(p)
				if transition%2 == 0 {
					glitches += gate.pulseUp
				} else {
					glitches += gate.pulseDown
				}
				out := p.objects[depth+1]
				if out.nSwitches != 0 || out.nGlitches != glitches {
					t.Fatalf("%s depth%d transition%d switches%d glitches%d want%d", gate.name, depth, transition, out.nSwitches, out.nGlitches, glitches)
				}
				for i := 0; i <= depth; i++ {
					o := p.objects[i]
					if o.nSwitches != int32(transition+1) || o.nGlitches != int32(transition+1) || o.fMark != 0 {
						t.Fatalf("path activity or mark mismatch depth%d object%d", depth, i)
					}
				}
			}
			Gli_ManStop(p)
		}
	}
}

func TestGlitchDeepPropagationAndDuplicateFanouts(t *testing.T) {
	g := glitchReviewChainGraph(256)
	// Repeated edges share a marked frontier node; they must not double count it.
	g.nodes = append(g.nodes, glitchReviewNode{[]int32{0, 0}, 8})
	for i := 0; i < 128; i++ {
		g.nodes = append(g.nodes, glitchReviewNode{[]int32{0}, 0xaaaaaaaaaaaaaaaa})
	}
	p := g.build()
	defer Gli_ManStop(p)
	Gli_ManFinalize(p)
	for transition := int32(1); transition <= 4; transition++ {
		Vec_IntClear(p.vCisChanged)
		gliFlip(p, p.objects[0])
		Gli_ManSwitching(p)
		Gli_ManGlitching(p)
		Gli_ManVerify(p)
		for _, o := range p.objects[:1+len(g.nodes)] {
			if o.nSwitches != transition || o.nGlitches != transition || o.fPhase != uint32(transition&1) || o.fMark != 0 {
				t.Fatalf("transition%d object%d state/count mismatch", transition, o.Handle)
			}
		}
	}
}
