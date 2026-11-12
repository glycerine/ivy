package mapper

import "testing"

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
