package mapper

import (
	"math/rand"
	"testing"
)

func TestCanonPhaseAndSlowMinimum(t *testing.T) {
	var signs [6][2]uint32
	signs[0][0] = 0xaaaaaaaa
	truth := []uint32{0xaaaaaaaa, 0}
	phases, result := make([]uint8, 4), make([]uint32, 2)
	if got := Map_CanonComputePhase(signs, 1, truth[0], 1); got != 0x55555555 {
		t.Fatalf("one-variable phase = %#08x, want %#08x", got, uint32(0x55555555))
	}
	if n := Map_CanonComputeSlow(signs, 1, 1, truth, phases, result); n != 1 {
		t.Fatalf("canonical phase count = %d, want 1", n)
	}
	if result[0] != 0x55555555 || result[1] != result[0] || phases[0] != 1 {
		t.Fatalf("canonical result=%#v phases=%v", result, phases[:1])
	}
}

func TestSlowCanonicalizationAgainstTruthEnumeration(t *testing.T) {
	var signs [6][2]uint32
	Map_MappingSetupTruthTables(signs[:])
	rng := rand.New(rand.NewSource(80213))
	for variables := int32(1); variables <= 6; variables++ {
		for trial := 0; trial < 40; trial++ {
			var source uint64
			for input := 0; input < 1<<variables; input++ {
				if rng.Intn(2) != 0 {
					source |= uint64(1) << input
				}
			}
			var repeated uint64
			for input := 0; input < 64; input++ {
				if source>>(input&((1<<variables)-1))&1 != 0 {
					repeated |= uint64(1) << input
				}
			}
			want := ^uint64(0)
			var wantPhases []uint8
			for phase := 0; phase < 1<<variables; phase++ {
				var transformed uint64
				for input := 0; input < 64; input++ {
					if repeated>>(input^phase)&1 != 0 {
						transformed |= uint64(1) << input
					}
				}
				if transformed < want {
					want, wantPhases = transformed, []uint8{uint8(phase)}
				} else if transformed == want && len(wantPhases) < 4 {
					wantPhases = append(wantPhases, uint8(phase))
				}
			}
			truth := []uint32{uint32(repeated), uint32(repeated >> 32)}
			result, phases := make([]uint32, 2), make([]uint8, 4)
			count := Map_CanonComputeSlow(signs, variables, variables, truth, phases, result)
			got := uint64(result[0]) | uint64(result[1])<<32
			if got != want || int(count) != len(wantPhases) {
				t.Fatalf("vars=%d truth=%x got=%x want=%x phases=%v want=%v", variables, repeated, got, want, phases[:count], wantPhases)
			}
			for i := range wantPhases {
				if phases[i] != wantPhases[i] {
					t.Fatalf("phase order=%v want=%v", phases[:count], wantPhases)
				}
			}
		}
	}
}
