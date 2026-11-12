package extra

import (
	"crypto/sha256"
	"encoding/binary"
	"fmt"
	"math/rand"
	"slices"
	"testing"
)

func canonOracle(truth uint64, n int32) (uint64, []byte) {
	best := ^uint64(0)
	var phases []byte
	for p := 0; p < 1<<uint(n); p++ {
		var value uint64
		for bit := 0; bit < 1<<uint(n); bit++ {
			if truth&(uint64(1)<<uint(bit^p)) != 0 {
				value |= uint64(1) << uint(bit)
			}
		}
		if value < best {
			best = value
			phases = phases[:0]
		}
		if value == best {
			phases = append(phases, byte(p))
		}
	}
	return best, phases
}
func TestCanonNegationExhaustiveFourVariables(t *testing.T) {
	for f := uint32(0); f < 65536; f++ {
		words, phases := Extra_TruthCanonFastN(5, 4, []uint32{f})
		want, wantedPhases := canonOracle(uint64(f), 4)
		got := uint32(want) | uint32(want)<<16
		if words[0] != got {
			t.Fatalf("%04x got %08x want %08x", f, words[0], got)
		}
		slices.Sort(phases)
		if !slices.Equal(phases, wantedPhases) {
			t.Fatal("missing phase", f, phases, wantedPhases)
		}
	}
}
func TestCanonNegationFiveAndSix(t *testing.T) {
	rng := rand.New(rand.NewSource(91))
	for n := int32(5); n <= 6; n++ {
		for trial := 0; trial < 400; trial++ {
			value := rng.Uint64()
			words := []uint32{uint32(value), uint32(value >> 32)}
			if n == 5 {
				value = uint64(words[0])
			}
			got, phases := Extra_TruthCanonFastN(n, n, words)
			want, wantedPhases := canonOracle(value, n)
			canonical := uint64(got[0])
			if n == 6 {
				canonical |= uint64(got[1]) << 32
			}
			if canonical != want {
				t.Fatal(n, value, canonical, want)
			}
			slices.Sort(phases)
			if !slices.Equal(phases, wantedPhases) {
				t.Fatal(n, "phases")
			}
		}
	}
	for n := int32(3); n <= 6; n++ {
		for _, value := range []uint32{0, ^uint32(0)} {
			words, phases := Extra_TruthCanonFastN(6, n, []uint32{value, value})
			if words[0] != value || words[1] != value || len(phases) != 1<<uint(n) {
				t.Fatal("constant phases", n, words, phases)
			}
		}
	}
}
func TestCanonResultsOwnedAndDimensions(t *testing.T) {
	a, pa := Extra_TruthCanonFastN(6, 3, []uint32{1})
	b, pb := Extra_TruthCanonFastN(6, 3, []uint32{1})
	a[0] = 9
	pa[0] = 9
	if b[0] != 0x01010101 || b[1] != b[0] || pb[0] != 0 {
		t.Fatal("static storage alias")
	}
	for _, run := range []func(){func() { Extra_TruthCanonFastN(7, 4, []uint32{0}) }, func() { Extra_TruthCanonFastN(5, 6, []uint32{0, 0}) }} {
		func() {
			defer func() {
				if recover() == nil {
					t.Error("invalid dimensions")
				}
			}()
			run()
		}()
	}
	// This API is source-empty, not an unimplemented simulation substitute.
	Extra_CommandCfs(nil, 0, 0, 0, 0, 0, 0, 0, 0, 0)
}

// Canonical truth words emitted by the compiled original C fast-N routine for
// every 4-input function. Phase completeness is tested independently above:
// the original C trivial cases omit valid phases of constant cofactors.
func TestCanonOriginalCTruthOracle(t *testing.T) {
	hash := sha256.New()
	record := make([]byte, 4)
	for f := uint32(0); f < 65536; f++ {
		words, _ := Extra_TruthCanonFastN(5, 4, []uint32{f})
		binary.LittleEndian.PutUint32(record, words[0])
		hash.Write(record)
	}
	if got := fmt.Sprintf("%x", hash.Sum(nil)); got != "cc0feac8b1ff8256c194e977d2b8dc6fb94f5c55d87ed77044cabce1a08c59ac" {
		t.Fatal(got)
	}
}
func TestCanonConcurrentOwnership(t *testing.T) {
	for i := 0; i < 8; i++ {
		t.Run(fmt.Sprint(i), func(t *testing.T) {
			t.Parallel()
			for j := uint32(0); j < 100; j++ {
				words, phases := Extra_TruthCanonFastN(6, 4, []uint32{j})
				if len(words) != 2 || len(phases) == 0 {
					t.Fatal("invalid result")
				}
				words[0] = 99
				phases[0] = 99
			}
		})
	}
}
