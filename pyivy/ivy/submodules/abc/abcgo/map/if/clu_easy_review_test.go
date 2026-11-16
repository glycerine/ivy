package if_

import (
	"fmt"
	"math/bits"
	"slices"
	"testing"
)

func TestCluEasyCofactorSourceWordPriority(t *testing.T) {
	// C compares masked low-variable cofactors with an unmasked complement,
	// so low-variable XOR does not match state4. High-variable word pairs do.
	for _, tc := range []struct{ n, v, want int32 }{{6, 0, -1}, {6, 5, -1}, {7, 0, -1}, {7, 5, -1}, {7, 6, 4}} {
		truth := make([]uint64, If_CluWordNum(tc.n))
		for x := 0; x < 64*len(truth); x++ {
			if bits.OnesCount(uint(x))%2 != 0 {
				truth[x>>6] |= 1 << uint(x&63)
			}
		}
		if got := If_CluDetectSpecialCaseCofs(truth, tc.n, tc.v); got != tc.want {
			t.Errorf("parity inputs%d variable%d state%d want%d", tc.n, tc.v, got, tc.want)
		}
	}
	// Cof1 is globally zero, but the source's per-word else-if priority
	// classifies the all-zero first word as state0 and the second as state2.
	truth := make([]uint64, 2)
	for x := 64; x < 128; x++ {
		if x&1 == 0 && bits.OnesCount(uint(x>>1&31))%2 != 0 {
			truth[x>>6] |= 1 << uint(x&63)
		}
	}
	if got := If_CluDetectSpecialCaseCofs(truth, 7, 0); got != -1 {
		t.Errorf("mixed word priority state%d want-1", got)
	}
	for i := range truth {
		truth[i] = ^truth[i]
	}
	if got := If_CluDetectSpecialCaseCofs(truth, 7, 0); got != -1 {
		t.Errorf("complemented mixed word priority state%d want-1", got)
	}
	// A high-variable XOR also loses its easy classification when one
	// word is consumed by an earlier constant-cofactor state.
	parity := uint64(0x6996966996696996)
	high := []uint64{0, parity, ^uint64(0), ^parity}
	if got := If_CluDetectSpecialCaseCofs(high, 8, 7); got != -1 {
		t.Errorf("high XOR mixed priority state%d want-1", got)
	}
}

func TestCluParityUsesSourceInitialBoundSet(t *testing.T) {
	var truth uint64
	for x := 0; x < 64; x++ {
		if bits.OnesCount(uint(x))%2 != 0 {
			truth |= 1 << x
		}
	}
	if g := If_CluDecUsingCofs([]uint64{truth}, 6, 4); g.nVars != 0 {
		t.Errorf("low-variable parity incorrectly took easy shortcut: %+v", g)
	}
	var root If_Grp_t
	var f0, f1 uint64
	g := If_CluCheck(nil, []uint64{truth}, 6, 0, 0, 4, 4, &root, &f0, &f1, nil, 0)
	if g.nVars != 4 || g.nMyu != 2 || !slices.Equal(g.pVars[:4], []int32{2, 3, 4, 5}) || root.nVars != 3 || !slices.Equal(root.pVars[:3], []int32{6, 0, 1}) {
		t.Errorf("source groups bound%+v root%+v", g, root)
	}
	if f0 != 0x9696969696969696 || f1 != 0x6996699669966996 {
		t.Errorf("source parity functions%016x/%016x", f0, f1)
	}
}

func TestCluEasyCofactorStateDimensions(t *testing.T) {
	for _, n := range []int{3, 6, 7, 10, 16} {
		for variable := 0; variable < n; variable++ {
			for state := 0; state <= 5; state++ {
				t.Run(fmt.Sprintf("inputs%d/variable%d/state%d", n, variable, state), func(t *testing.T) {
					truth := make([]uint64, If_CluWordNum(int32(n)))
					for x := 0; x < 64*len(truth); x++ {
						others := x & ((1 << n) - 1) &^ (1 << variable)
						parity := bits.OnesCount(uint(others))%2 != 0
						bit := x>>variable&1 != 0
						value := false
						switch state {
						case 0:
							value = bit && parity
						case 1:
							value = !bit || parity
						case 2:
							value = !bit && parity
						case 3:
							value = bit || parity
						case 4:
							value = bit != parity
						case 5:
							value = others == ((1<<n)-1)&^(1<<variable)
							if bit {
								value = parity
							}
						}
						if value {
							truth[x>>6] |= 1 << uint(x&63)
						}
					}
					before := slices.Clone(truth)
					want := int32(state)
					if state == 5 || state == 4 && variable < 6 {
						want = -1
					}
					if got := If_CluDetectSpecialCaseCofs(truth, int32(n), int32(variable)); got != want {
						t.Fatalf("state%d want%d", got, want)
					}
					if !slices.Equal(truth, before) {
						t.Fatal("detection modified source truth")
					}
				})
			}
		}
	}
}
