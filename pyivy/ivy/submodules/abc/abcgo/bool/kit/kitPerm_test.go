package kit

import (
	"math/rand"
	"testing"
)

func TestPermutationReverseVariables(t *testing.T) {
	rng := rand.New(rand.NewSource(42))
	for V := int32(6); V <= 16; V++ {
		n := 1 << uint(V-6)
		input := make([]uint64, n)
		for i := range input {
			input[i] = rng.Uint64()
		}
		output := make([]uint64, n)
		Kit_PermCompute(output, input, V)
		for x := 0; x < n*64; x++ {
			reverse := 0
			for b := int32(0); b < V; b++ {
				reverse |= ((x >> uint(b)) & 1) << uint(V-1-b)
			}
			want := (input[x>>6] >> uint(x&63)) & 1
			got := (output[reverse>>6] >> uint(reverse&63)) & 1
			if got != want {
				t.Fatalf("V=%d x=%d: %d != %d", V, x, got, want)
			}
		}
		Kit_PermComputeNaive(output, V)
		for i := range input {
			if output[i] != input[i] {
				t.Fatal("reversal is not involutive", V, i)
			}
		}
	}
}
func TestRecursivePermutationReference(t *testing.T) {
	rng := rand.New(rand.NewSource(19))
	for i := 0; i < 1000; i++ {
		input := []uint64{rng.Uint64()}
		output := make([]uint64, 1)
		Kit_PermCompute(output, input, 6)
		if got := Tf(input[0], 6); got != output[0] {
			t.Fatalf("%x: recursive %x, permutation %x", input[0], got, output[0])
		}
	}
}
func TestPermutationTablesAndCycle(t *testing.T) {
	for i := 0; i < 256; i++ {
		var expected uint16
		for bit := 0; bit < 8; bit++ {
			expected |= uint16((i>>bit)&1) << uint(2*bit)
		}
		if S[i] != expected {
			t.Fatal(i, S[i], expected)
		}
		p := int(P[i])
		if p >= 256 || P[p] != byte(i) {
			t.Fatal("invalid swap permutation", i, p)
		}
	}
	input := []byte{0x81, 0x42}
	out := make([]uint16, 1)
	Kit_PermCycleOne(out, input, 3)
	if out[0] != S[0x81]|S[0x42]<<1 {
		t.Fatal(out)
	}
}
func TestPermutationRejectsShortBuffers(t *testing.T) {
	defer func() {
		if recover() == nil {
			t.Fatal("expected bounds rejection")
		}
	}()
	Kit_PermCompute(nil, nil, 7)
}
