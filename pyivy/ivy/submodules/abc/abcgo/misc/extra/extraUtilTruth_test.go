package extra

import (
	"crypto/sha256"
	"encoding/binary"
	"fmt"
	"math/bits"
	"math/rand"
	"slices"
	"testing"
)

func truthOracle(in []uint32, mapBit func(int) int) []uint32 {
	out := make([]uint32, len(in))
	for bit := 0; bit < len(in)*32; bit++ {
		src := mapBit(bit)
		if in[src/32]&(1<<uint(src%32)) != 0 {
			out[bit/32] |= 1 << uint(bit%32)
		}
	}
	return out
}
func TestTruthWordOperationsAgainstBitOracle(t *testing.T) {
	rng := rand.New(rand.NewSource(7))
	for n := int32(1); n <= 10; n++ {
		in := make([]uint32, truthWords(n))
		for i := range in {
			in[i] = rng.Uint32()
		}
		for v := int32(0); v < n; v++ {
			bit := 1 << uint(v)
			zero := truthOracle(in, func(i int) int { return i &^ bit })
			one := truthOracle(in, func(i int) int { return i | bit })
			got := slices.Clone(in)
			Extra_TruthCofactor0(got, n, v)
			if !slices.Equal(got, zero) {
				t.Fatal("cof0", n, v)
			}
			got = slices.Clone(in)
			Extra_TruthCofactor1(got, n, v)
			if !slices.Equal(got, one) {
				t.Fatal("cof1", n, v)
			}
			want := truthOracle(in, func(i int) int { return i ^ bit })
			got = slices.Clone(in)
			Extra_TruthChangePhase(got, n, v)
			if !slices.Equal(got, want) {
				t.Fatal("phase", n, v)
			}
			Extra_TruthChangePhase(got, n, v)
			if !slices.Equal(got, in) {
				t.Fatal("phase involution")
			}
			for _, forall := range []bool{false, true} {
				got = slices.Clone(in)
				if forall {
					Extra_TruthForall(got, n, v)
				} else {
					Extra_TruthExist(got, n, v)
				}
				for w := range got {
					want := zero[w] | one[w]
					if forall {
						want = zero[w] & one[w]
					}
					if got[w] != want {
						t.Fatal("quantify", n, v)
					}
				}
			}
			out := make([]uint32, len(in))
			Extra_TruthMux(out, zero, one, n, v)
			if !slices.Equal(out, in) {
				t.Fatal("mux", n, v)
			}
			Extra_TruthMux(zero, zero, one, n, v)
			if !slices.Equal(zero, in) {
				t.Fatal("alias mux", n, v)
			}
			supported := int32(0)
			if !slices.Equal(truthOracle(in, func(i int) int { return i &^ bit }), one) {
				supported = 1
			}
			if Extra_TruthVarInSupport(in, n, v) != supported {
				t.Fatal("support", n, v)
			}
			if v+1 < n {
				swap := func(i int) int {
					a, b := (i>>uint(v))&1, (i>>uint(v+1))&1
					if a != b {
						return i ^ bit ^ (bit << 1)
					}
					return i
				}
				want = truthOracle(in, swap)
				out = make([]uint32, len(in))
				Extra_TruthSwapAdjacentVars(out, in, n, v)
				if !slices.Equal(out, want) {
					t.Fatal("swap", n, v)
				}
				alias := slices.Clone(in)
				Extra_TruthSwapAdjacentVars(alias, alias, n, v)
				if !slices.Equal(alias, want) {
					t.Fatal("alias swap", n, v)
				}
			}
		}
		counts := make([]int32, 2*n)
		Extra_TruthCountOnesInCofs(in, n, counts)
		oracle := make([]int32, 2*n)
		for bit := 0; bit < len(in)*32; bit++ {
			if in[bit/32]&(1<<uint(bit%32)) == 0 {
				continue
			}
			for v := int32(0); v < n; v++ {
				oracle[2*int(v)+(bit>>uint(v))&1]++
			}
		}
		if !slices.Equal(counts, oracle) {
			t.Fatal("counts", n, counts, oracle)
		}
	}
}
func TestTruthStretchShrinkAndElementary(t *testing.T) {
	for all := int32(2); all <= 8; all++ {
		for phase := uint32(0); phase < 1<<uint(all); phase++ {
			n := int32(bits.OnesCount32(phase))
			if n == 0 {
				continue
			}
			table := make([]uint32, truthWords(all))
			for i := range table {
				table[i] = 0x96696996
			}
			out := make([]uint32, len(table))
			back := make([]uint32, len(table))
			Extra_TruthStretch(out, table, n, all, phase)
			Extra_TruthShrink(back, out, n, all, phase)
			if !slices.Equal(back, table) {
				t.Fatal("roundtrip", n, all, phase)
			}
		}
	}
	for n := int32(1); n <= 10; n++ {
		elementary := Extra_TruthElementary(n)
		for v, table := range elementary {
			if Extra_TruthSupport(table, n) != 1<<uint(v) {
				t.Fatal(n, v)
			}
		}
	}
}
func TestTruthSymmetryAndOverlap(t *testing.T) {
	for n := int32(2); n <= 8; n++ {
		vars := Extra_TruthElementary(n)
		xor := make([]uint32, truthWords(n))
		for i := range xor {
			xor[i] = vars[0][i] ^ vars[1][i]
		}
		if Extra_TruthVarsSymm(xor, n, 0, 1) != 1 || Extra_TruthVarsAntiSymm(xor, n, 0, 1) != 1 {
			t.Fatal("xor symmetry")
		}
		if Extra_TruthVarsSymm(vars[0], n, 0, 1) != 0 {
			t.Fatal("asymmetric")
		}
		overlap, v := Extra_TruthMinCofSuppOverlap(xor, n)
		if overlap != 1 || v != 0 {
			t.Fatal(overlap, v)
		}
	}
}
func TestTruthHashPrimes(t *testing.T) {
	if truthHashPrimes[0] != 2 || truthHashPrimes[1023] != 8161 {
		t.Fatal("primes")
	}
	if Extra_TruthHash([]uint32{10, 20}, 2) != (2*10)^(3*20) {
		t.Fatal("hash")
	}
}

// Oracle is SHA256 of the original C semi-canonicalizer run on all 65536
// four-input functions, emitting canonical uint32, phase uint32, 4 permutation
// bytes, then 8 cofactor int16 counts, all integers little-endian.
func TestTruthSemiCanonicalOriginalCOracle(t *testing.T) {
	hash := sha256.New()
	record := make([]byte, 28)
	for f := uint32(0); f < 65536; f++ {
		table := []uint32{f | f<<16}
		aux := make([]uint32, 1)
		perm := []byte{0, 1, 2, 3}
		store := make([]int32, 8)
		phase := Extra_TruthSemiCanonicize(table, aux, 4, perm, store)
		binary.LittleEndian.PutUint32(record, table[0])
		binary.LittleEndian.PutUint32(record[4:], phase)
		copy(record[8:], perm)
		for i, count := range store {
			binary.LittleEndian.PutUint16(record[12+2*i:], uint16(count))
		}
		hash.Write(record)
	}
	if got := fmt.Sprintf("%x", hash.Sum(nil)); got != "4ed150f1a1b0c3132a7186d6919afc49f67823388eba888a994b49ae1e1b1f5c" {
		t.Fatal(got)
	}
}
func TestTruthCountNoShortOverflow(t *testing.T) {
	table := make([]uint32, truthWords(17))
	for i := range table {
		table[i] = ^uint32(0)
	}
	counts := make([]int32, 34)
	Extra_TruthCountOnesInCofs(table, 17, counts)
	for _, count := range counts {
		if count != 65536 {
			t.Fatal(count)
		}
	}
}

func TestTruthPartialAliasAndValidation(t *testing.T) {
	data := []uint32{0x12345678, 0xabcdef01, 0}
	want := make([]uint32, 2)
	Extra_TruthSwapAdjacentVars(want, data[:2], 6, 4)
	Extra_TruthSwapAdjacentVars(data[1:], data[:2], 6, 4)
	if !slices.Equal(data[1:], want) {
		t.Fatal("partially overlapping swap")
	}
	for _, run := range []func(){func() { Extra_TruthSupport(nil, -1) }, func() { Extra_TruthChangePhase([]uint32{0}, 4, 4) }, func() { Extra_TruthSwapAdjacentVars([]uint32{0}, []uint32{0}, 4, 3) }, func() { Extra_TruthStretch([]uint32{0}, []uint32{0}, 2, 4, 1) }} {
		func() {
			defer func() {
				if recover() == nil {
					t.Error("invalid truth operation accepted")
				}
			}()
			run()
		}()
	}
}
