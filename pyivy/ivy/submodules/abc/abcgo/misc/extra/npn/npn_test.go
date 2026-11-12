package npn

import (
	"crypto/sha256"
	"encoding/binary"
	"fmt"
	"testing"
)

func TestExhaustiveCanonicalization(t *testing.T) {
	c, p, r, m := Extra_Truth4VarNPN()
	permutations := Permutations()
	representatives := map[uint16]bool{}
	for f := 0; f < 65536; f++ {
		if got := Transform(c[f], p[f], permutations[r[f]]); got != uint16(f) {
			t.Fatalf("%04x reconstructed as %04x", f, got)
		}
		if c[c[f]] != c[f] || m[f] != m[c[f]] {
			t.Fatalf("noncanonical mapping %04x", f)
		}
		if c[f] > uint16(f) {
			t.Fatalf("canonical larger than input %04x", f)
		}
		representatives[c[f]] = true
		if c[f] != c[65535^f] || m[f] != m[65535^f] {
			t.Fatalf("output complement %04x", f)
		}
	}
	if len(representatives) != 222 {
		t.Fatal(len(representatives))
	}
	classes := map[byte]bool{}
	for _, v := range m {
		classes[v] = true
	}
	if len(classes) != 222 {
		t.Fatal("class IDs", len(classes))
	}
	for representative := range representatives {
		for phase := byte(0); phase < 32; phase++ {
			for _, perm := range permutations {
				f := Transform(representative, phase, perm)
				if c[f] != representative {
					t.Fatalf("orbit mismatch %04x", f)
				}
			}
		}
	}
}
func TestPermutationOrderAndFreshTables(t *testing.T) {
	perms := Permutations()
	if perms[0] != [4]byte{0, 1, 2, 3} {
		t.Fatal("ABC identity index", perms[0])
	}
	seen := map[[4]byte]bool{}
	for _, p := range perms {
		seen[p] = true
	}
	if len(seen) != 24 {
		t.Fatal("duplicates")
	}
	c, _, _, _ := Extra_Truth4VarNPN()
	c[0] = 9
	d, _, _, _ := Extra_Truth4VarNPN()
	if d[0] != 0 {
		t.Fatal("tables alias")
	}
}

// Oracle obtained by compiling the original six C functions from
// src/misc/extra/extraUtilMisc.c and emitting canons (little-endian uint16),
// phases and permutation indices in table order. No C compiler is needed here.
func TestOriginalCTableOracle(t *testing.T) {
	c, p, r, _ := Extra_Truth4VarNPN()
	data := make([]byte, 262144)
	for i, v := range c {
		binary.LittleEndian.PutUint16(data[2*i:], v)
	}
	copy(data[131072:], p)
	copy(data[196608:], r)
	got := fmt.Sprintf("%x", sha256.Sum256(data))
	if got != "64d528981c19fcb7753125f7832f2f2135f05dbd5d8437e095f3e07e8baa4ea8" {
		t.Fatal(got)
	}
}
